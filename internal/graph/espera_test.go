package graph

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

const (
	// esperaCorta es la espera propia con la que las pruebas provocan en
	// milisegundos un bloqueo que dura más que ella.
	esperaCorta = 300 * time.Millisecond

	// tiempoDeSobra es lo que una espera cortada por el contexto, o una espera
	// corta, no puede tardar: la mitad de la espera propia entera. Es lo que
	// distingue «respeta el contexto» de «agotó los cinco segundos», con margen
	// para una máquina cargada.
	tiempoDeSobra = esperaPropia / 2
)

// TestConstantesDeLaEspera fija los números de contracts/almacen-world-db.md
// §4, «Esperas»: tramos de 100 ms dentro del motor, que es el busy_timeout de
// cada conexión, y una espera propia total de 5 s.
func TestConstantesDeLaEspera(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 100*time.Millisecond, tramoDeEspera)
	assert.Equal(t, 5*time.Second, esperaPropia)
	assert.Equal(t, "_pragma=busy_timeout(100)", pragmaDelTramo, "el tramo del motor es el de la espera")
	assert.Equal(t, espera{total: esperaPropia}, esperaDeLaInvocacion())
}

// TestReintentar fija la espera por tramos (FR-014): solo un bloqueo se repite;
// entre intento e intento se mira el contexto y se completa el tramo aunque
// SQLite conteste sin esperar, de modo que no se reintenta en caliente; y la
// espera termina cuando se suelta el bloqueo, cuando termina el contexto o
// cuando se agota la espera propia, lo que ocurra antes.
//
// El intento es sintético y devuelve el SQLITE_BUSY que SQLite dio de verdad a
// una conexión que pedía el bloqueo de escritura que otra retenía: así el
// número de intentos y el tiempo no dependen del motor.
func TestReintentar(t *testing.T) {
	t.Parallel()

	ocupada := bloqueoDeVerdad(t)
	otro := errors.New("otro fallo")

	casos := []struct {
		nombre string
		espera espera
		plazo  time.Duration
		// respuestas son las de cada intento, en orden; el último se repite.
		respuestas []error
		// intentos es cuántos se esperan, o el máximo si minimo no es cero.
		intentos, minimo int
		// esperado es lo que devuelve reintentar.
		esperado error
		// tarda es lo que tiene que tardar al menos, y sobra lo que no puede
		// alcanzar.
		tarda, sobra time.Duration
	}{
		{
			nombre:     "sin bloqueo, un solo intento",
			espera:     esperaDeLaInvocacion(),
			respuestas: []error{nil},
			intentos:   1,
			sobra:      tramoDeEspera,
		},
		{
			nombre:     "otro fallo no se repite",
			espera:     esperaDeLaInvocacion(),
			respuestas: []error{otro},
			intentos:   1,
			esperado:   otro,
			sobra:      tramoDeEspera,
		},
		{
			nombre:     "el bloqueo se repite tramo a tramo hasta que se suelta",
			espera:     esperaDeLaInvocacion(),
			respuestas: []error{ocupada, ocupada, nil},
			intentos:   3,
			tarda:      2 * tramoDeEspera,
			sobra:      tiempoDeSobra,
		},
		{
			nombre:     "el bloqueo que no se suelta agota la espera propia",
			espera:     espera{total: esperaCorta},
			respuestas: []error{ocupada},
			minimo:     2,
			intentos:   int(esperaCorta/tramoDeEspera) + 1,
			esperado:   ocupada,
			tarda:      esperaCorta,
			sobra:      tiempoDeSobra,
		},
		{
			nombre:     "el contexto que termina corta la espera",
			espera:     esperaDeLaInvocacion(),
			plazo:      esperaCorta,
			respuestas: []error{ocupada},
			minimo:     2,
			intentos:   int(esperaCorta/tramoDeEspera) + 1,
			esperado:   ocupada,
			tarda:      esperaCorta,
			sobra:      tiempoDeSobra,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			if caso.plazo > 0 {
				conPlazo, cancela := context.WithTimeout(ctx, caso.plazo)
				t.Cleanup(cancela)

				ctx = conPlazo
			}

			var llamadas int

			inicio := time.Now()
			err := caso.espera.reintentar(ctx, func() error {
				llamadas++

				return caso.respuestas[min(llamadas, len(caso.respuestas))-1]
			})
			tardo := time.Since(inicio)

			if caso.esperado == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, caso.esperado, "devuelve el fallo del último intento tal cual")
			}

			if caso.minimo > 0 {
				assert.GreaterOrEqual(t, llamadas, caso.minimo, "se repitió mientras duraba la espera")
				assert.LessOrEqual(t, llamadas, caso.intentos, "sin reintentar en caliente: un intento por tramo")
			} else {
				assert.Equal(t, caso.intentos, llamadas)
			}

			assert.GreaterOrEqual(t, tardo, caso.tarda, "cada tramo se espera entero")
			assert.Less(t, tardo, caso.sobra)
		})
	}
}

// TestReintentarConElContextoYaTerminado fija que un contexto ya terminado no
// espera ningún tramo: el intento se hace una vez —y es la operación la que
// contesta al contexto— y su fallo vuelve en el acto.
func TestReintentarConElContextoYaTerminado(t *testing.T) {
	t.Parallel()

	ocupada := bloqueoDeVerdad(t)

	ctx, cancela := context.WithCancel(t.Context())
	cancela()

	var llamadas int

	inicio := time.Now()
	err := esperaDeLaInvocacion().reintentar(ctx, func() error {
		llamadas++

		return ocupada
	})

	require.ErrorIs(t, err, ocupada)
	assert.Equal(t, 1, llamadas)
	assert.Less(t, time.Since(inicio), tramoDeEspera)
}

// TestFalloDeLaEspera fija cómo se clasifica lo que la espera explica
// (contracts/almacen-world-db.md §4 «Esperas» y §6): el contexto terminado es
// el plazo agotado, «fuente-no-disponible», con el error del contexto y el del
// motor alcanzables; el bloqueo con el contexto vivo es la espera propia
// agotada, «inesperado», con la ruta y la espera en el mensaje; y lo demás no
// es cosa de la espera.
func TestFalloDeLaEspera(t *testing.T) {
	t.Parallel()

	ocupada := bloqueoDeVerdad(t)
	ruta := filepath.Join("cache", "world.db")
	otro := errors.New("otro fallo")

	terminado, cancela := context.WithCancel(t.Context())
	cancela()

	casos := []struct {
		nombre string
		ctx    context.Context
		causa  error
		// clase vacía: la espera no lo explica y devuelve nil.
		clase   schema.Clase
		mensaje string
		alcanza []error
	}{
		{
			nombre:  "el contexto terminado durante un bloqueo es el plazo agotado",
			ctx:     terminado,
			causa:   ocupada,
			clase:   schema.ClaseFuenteNoDisponible,
			mensaje: "grafo: el plazo terminó antes de escribir " + nombrar(ruta),
			alcanza: []error{context.Canceled, ocupada},
		},
		{
			nombre:  "el error del contexto con el contexto vivo también es el plazo",
			ctx:     t.Context(),
			causa:   context.DeadlineExceeded,
			clase:   schema.ClaseFuenteNoDisponible,
			mensaje: "grafo: el plazo terminó antes de escribir " + nombrar(ruta),
			alcanza: []error{context.DeadlineExceeded},
		},
		{
			nombre:  "el bloqueo con el contexto vivo es la espera propia agotada",
			ctx:     t.Context(),
			causa:   ocupada,
			clase:   schema.ClaseInesperado,
			mensaje: "grafo: " + nombrar(ruta) + " está bloqueada por otra invocación y la espera de 300ms se agotó",
			alcanza: []error{ocupada},
		},
		{
			nombre: "otro fallo no lo explica la espera",
			ctx:    t.Context(),
			causa:  otro,
		},
		{
			nombre: "sin fallo no hay nada que explicar",
			ctx:    terminado,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			err := espera{total: esperaCorta}.falloDeLaEspera(caso.ctx, operacionEscribir, ruta, caso.causa)

			if caso.clase == "" {
				require.NoError(t, err)

				return
			}

			var fallo *Error

			require.ErrorAs(t, err, &fallo)
			assert.Equal(t, caso.clase, fallo.Clase())
			assert.Equal(t, caso.mensaje, fallo.Error())
			assert.Equal(t, operacionEscribir, fallo.Operacion)
			assert.Equal(t, ruta, fallo.Ruta)

			for _, alcanzable := range caso.alcanza {
				require.ErrorIs(t, err, alcanzable)
			}
		})
	}
}

// TestEsperaAnteUnBloqueoDeVerdad ejerce la espera contra SQLite: otra conexión
// retiene el bloqueo de escritura con una transacción inmediata y el intento es
// empezar otra, que es como empieza toda entrega. Si la otra lo suelta a tiempo,
// la espera termina bien; si no, la agota la espera propia o el plazo, lo que
// llegue antes, y cada uno se clasifica con su clase (FR-014).
func TestEsperaAnteUnBloqueoDeVerdad(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		espera espera
		plazo  time.Duration
		// suelta es cuándo suelta la otra el bloqueo; cero, nunca.
		suelta time.Duration
		// clase vacía: la espera termina bien.
		clase schema.Clase
		tarda time.Duration
	}{
		{
			nombre: "la otra conexión suelta el bloqueo a tiempo",
			espera: esperaDeLaInvocacion(),
			suelta: 2 * tramoDeEspera,
			tarda:  2 * tramoDeEspera,
		},
		{
			nombre: "el bloqueo dura más que la espera propia",
			espera: espera{total: esperaCorta},
			clase:  schema.ClaseInesperado,
			tarda:  esperaCorta,
		},
		{
			nombre: "el plazo termina antes que la espera propia",
			espera: esperaDeLaInvocacion(),
			plazo:  esperaCorta,
			clase:  schema.ClaseFuenteNoDisponible,
			tarda:  esperaCorta,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			ruta := filepath.Join(t.TempDir(), "world.db")
			suelta := retenerElBloqueo(t, ruta)
			base := abrirBaseDePrueba(t, ruta, pragmaDelTramo+"&_txlock=immediate")

			ctx := t.Context()

			if caso.plazo > 0 {
				conPlazo, cancela := context.WithTimeout(ctx, caso.plazo)
				t.Cleanup(cancela)

				ctx = conPlazo
			}

			if caso.suelta > 0 {
				temporizador := time.AfterFunc(caso.suelta, suelta)
				t.Cleanup(func() { temporizador.Stop() })
			}

			inicio := time.Now()

			var tx *sql.Tx

			err := caso.espera.reintentar(ctx, func() (err error) {
				tx, err = base.BeginTx(context.WithoutCancel(ctx), nil)

				return err
			})
			tardo := time.Since(inicio)

			if caso.clase == "" {
				require.NoError(t, err, "la espera termina bien en cuanto la otra suelta el bloqueo")
				require.NoError(t, tx.Rollback())
				assert.GreaterOrEqual(t, tardo, caso.tarda)

				return
			}

			require.Error(t, err)

			var fallo *Error

			require.ErrorAs(t, caso.espera.falloDeLaEspera(ctx, operacionEscribir, ruta, err), &fallo)
			assert.Equal(t, caso.clase, fallo.Clase(), "%v", fallo)
			assert.Contains(t, fallo.Error(), ruta)
			assert.GreaterOrEqual(t, tardo, caso.tarda, "se esperó hasta el final")
			assert.Less(t, tardo, tiempoDeSobra, "y no la espera propia entera")
		})
	}
}

// bloqueoDeVerdad devuelve el error con que SQLite contesta a una conexión que
// pide el bloqueo de escritura mientras otra lo retiene: un SQLITE_BUSY del
// motor, no uno fabricado.
func bloqueoDeVerdad(t *testing.T) error {
	t.Helper()

	ruta := filepath.Join(t.TempDir(), "world.db")
	retenerElBloqueo(t, ruta)

	base := abrirBaseDePrueba(t, ruta, "_pragma=busy_timeout(0)&_txlock=immediate")

	tx, err := base.BeginTx(context.WithoutCancel(t.Context()), nil)
	if err == nil {
		require.NoError(t, tx.Rollback())
	}

	require.Error(t, err, "con el bloqueo retenido por otra conexión, empezar una transacción inmediata falla")
	require.True(t, esBloqueo(err), "y falla por SQLITE_BUSY: %v", err)

	var delMotor *sqlite.Error

	require.ErrorAs(t, err, &delMotor)
	require.Equal(t, sqlite3.SQLITE_BUSY, delMotor.Code()&0xff)

	return err
}

// retenerElBloqueo abre otra conexión sobre la base —la de otra invocación— y
// empieza en ella una transacción inmediata, que retiene el bloqueo de escritura
// hasta que se llame a lo que devuelve o hasta que termine la prueba. La base
// queda en WAL, como la deja una entrega.
func retenerElBloqueo(t *testing.T, ruta string) (suelta func()) {
	t.Helper()

	otra := abrirBaseDePrueba(t, ruta, pragmaDelTramo+"&_txlock=immediate")

	var modo string

	require.NoError(t, otra.QueryRowContext(t.Context(), "PRAGMA journal_mode=WAL").Scan(&modo))
	require.Equal(t, "wal", modo)

	tx, err := otra.BeginTx(context.WithoutCancel(t.Context()), nil)
	require.NoError(t, err, "la otra conexión toma el bloqueo de escritura")

	var una sync.Once

	suelta = func() {
		una.Do(func() {
			if err := tx.Rollback(); !errors.Is(err, sql.ErrTxDone) {
				assert.NoError(t, err)
			}
		})
	}

	t.Cleanup(suelta)

	return suelta
}

// abrirBaseDePrueba abre la base de la ruta con los parámetros de la cadena de
// conexión que la prueba necesita, con una sola conexión, y la cierra al
// terminar. Protege en la ruta lo que la sintaxis de URI de SQLite interpreta.
func abrirBaseDePrueba(t *testing.T, ruta, ajustesDeConexion string) *sql.DB {
	t.Helper()

	protegida := strings.NewReplacer("%", "%25", "?", "%3F", "#", "%23").Replace(filepath.ToSlash(ruta))

	base, err := sql.Open("sqlite", "file:"+protegida+"?"+ajustesDeConexion)
	require.NoError(t, err)

	base.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, base.Close()) })

	return base
}
