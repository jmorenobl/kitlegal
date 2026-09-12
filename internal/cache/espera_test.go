package cache

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

const (
	// plazoCorto es el plazo con el que estas tablas llaman a New y a Put
	// mientras otra invocación retiene el bloqueo: mucho menor que la espera
	// ante bloqueo, para que lo que termine la operación sea el contexto.
	plazoCorto = 300 * time.Millisecond

	// esperaCorta es la espera ante bloqueo con la que se provoca en
	// milisegundos un bloqueo que dura más que ella.
	esperaCorta = 300 * time.Millisecond

	// tiempoDeSobra es lo que una operación cortada por el contexto, o por una
	// espera corta, no puede tardar: la mitad de la espera entera. Es lo que
	// distingue «respeta el contexto» de «agotó los cinco segundos», con margen
	// para una máquina cargada.
	tiempoDeSobra = esperaAnteBloqueo / 2
)

// conEsperaAnteBloqueo acorta la espera ante bloqueo de un cliente. Solo existe
// para las pruebas: Opcion opera sobre los ajustes privados y el paquete no
// exporta ninguna forma de cambiar la espera (FR-031).
func conEsperaAnteBloqueo(espera time.Duration) Opcion {
	return func(a *ajustes) error {
		a.esperaAnteBloqueo = espera

		return nil
	}
}

// TestNewConLaBaseBloqueadaRespetaElContexto fija FR-003 y la fila 15 del
// contrato de errores sobre la apertura: otra invocación retiene el bloqueo de
// escritura de una base sin migrar y New llega con un plazo corto. La migración
// no puede empezar, y lo que termina la espera es el contexto: «fuente no
// disponible» (4), mucho antes de agotar los cinco segundos, sin devolver
// ningún cliente y sin dejar ninguna conexión abierta.
//
// Que no queda ninguna conexión se mide con los auxiliares del registro de
// escritura: al cerrar quien retenía el bloqueo, SQLite los retira si es la
// última conexión, y una que New hubiera dejado abierta los mantendría.
func TestNewConLaBaseBloqueadaRespetaElContexto(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	ruta := filepath.Join(directorio, ficheroDeLaBase)
	suelta := retieneElBloqueo(t, ruta)

	ctx, cancela := context.WithTimeout(t.Context(), plazoCorto)
	t.Cleanup(cancela)

	inicio := time.Now()
	cliente, err := New(ctx, ConDirectorio(directorio))
	tardo := time.Since(inicio)

	require.Error(t, err)
	assert.Nil(t, cliente, "un fallo al construir no devuelve ningún cliente")
	assert.Equal(t, schema.ClaseFuenteNoDisponible, cli.Clasificar(err))
	assert.Equal(t, 4, cli.CodigoSalida(err), "el plazo de quien llama es la fila 15, no un fichero inutilizable: %v", err)
	require.ErrorIs(t, err, context.DeadlineExceeded, "la causa del contexto sigue alcanzable")
	assert.NotContains(t, err.Error(), "no es una base de datos utilizable",
		"un bloqueo no convierte el fichero en inutilizable (FR-028, FR-035)")
	assert.Less(t, tardo, tiempoDeSobra, "la espera termina cuando termina el contexto, no cuando se agota")

	suelta()

	for _, sufijo := range []string{sufijoRegistroDeEscritura, sufijoMemoriaCompartida} {
		_, err := os.Stat(ruta + sufijo)
		require.ErrorIs(t, err, fs.ErrNotExist,
			"al soltar el bloqueo no queda ninguna conexión: New no dejó ninguna abierta")
	}
}

// TestPutConLaBaseBloqueadaRespetaElContexto fija FR-003 y la fila 15 sobre la
// escritura: con otra invocación reteniendo el bloqueo de escritura, un Put con
// plazo corto termina con «fuente no disponible» (4) en cuanto vence el plazo, y
// no cinco segundos después. Mientras tanto leer no espera a nadie: el diario
// en WAL es lo que hace que un lector no dependa del escritor (FR-030).
func TestPutConLaBaseBloqueadaRespetaElContexto(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	contenido := []byte("<norma>contenido</norma>")
	siembra(t, directorio, claveDePrueba, contenido)

	cliente := clienteAbierto(t, directorio)
	retieneElBloqueo(t, filepath.Join(directorio, ficheroDeLaBase))

	leido, presente, err := cliente.Get(t.Context(), claveDePrueba)
	require.NoError(t, err, "leer no espera al bloqueo de escritura")
	assert.True(t, presente)
	assert.Equal(t, contenido, leido)

	ctx, cancela := context.WithTimeout(t.Context(), plazoCorto)
	t.Cleanup(cancela)

	inicio := time.Now()
	err = cliente.Put(ctx, claveDePrueba, []byte("<norma>otra</norma>"), vigenciaDePrueba)
	tardo := time.Since(inicio)

	require.Error(t, err)
	assert.Equal(t, schema.ClaseFuenteNoDisponible, cli.Clasificar(err))
	assert.Equal(t, 4, cli.CodigoSalida(err), "%v", err)
	require.ErrorIs(t, err, context.DeadlineExceeded, "la causa del contexto sigue alcanzable")
	assert.Contains(t, err.Error(), claveDePrueba, "el mensaje nombra la clave")
	assert.Less(t, tardo, tiempoDeSobra, "la espera termina cuando termina el contexto, no cuando se agota")

	leido, presente, err = cliente.Get(t.Context(), claveDePrueba)
	require.NoError(t, err)
	assert.True(t, presente)
	assert.Equal(t, contenido, leido, "lo que no se pudo escribir no cambió nada")
}

// TestEsperaAQueSueltenElBloqueo fija FR-031: una invocación simultánea no
// falla de inmediato. Quien retiene el bloqueo lo suelta al cabo de un rato y
// New —sobre una base sin migrar— y Put terminan bien, porque la espera por
// tramos sigue mientras el contexto viva y el presupuesto no se agote.
func TestEsperaAQueSueltenElBloqueo(t *testing.T) {
	t.Parallel()

	const retenido = 2 * tramoDeEspera

	operaciones := []struct {
		nombre string
		// prepara deja el directorio como la operación lo necesita y devuelve
		// la llamada que va a encontrar la base bloqueada.
		prepara func(t *testing.T, directorio string) func(ctx context.Context) error
	}{
		{
			nombre: "New",
			prepara: func(t *testing.T, directorio string) func(ctx context.Context) error {
				t.Helper()

				return func(ctx context.Context) error {
					cliente, err := New(ctx, ConDirectorio(directorio))
					cierraAlTerminar(t, cliente)

					return err
				}
			},
		},
		{
			nombre: "Put",
			prepara: func(t *testing.T, directorio string) func(ctx context.Context) error {
				t.Helper()

				cliente := clienteAbierto(t, directorio)

				return func(ctx context.Context) error {
					return cliente.Put(ctx, claveDePrueba, []byte("<norma>contenido</norma>"), vigenciaDePrueba)
				}
			},
		},
	}

	for _, operacion := range operaciones {
		t.Run(operacion.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			opera := operacion.prepara(t, directorio)
			suelta := retieneElBloqueo(t, filepath.Join(directorio, ficheroDeLaBase))

			resultado := make(chan error, 1)

			var grupo sync.WaitGroup

			grupo.Go(func() { resultado <- opera(t.Context()) })
			t.Cleanup(grupo.Wait)

			time.Sleep(retenido)
			suelta()

			require.NoError(t, <-resultado,
				"la operación espera a que suelten el bloqueo y termina bien (FR-031)")
		})
	}
}

// TestBloqueoQueNoSeSueltaAgotaLaEspera fija la fila 14 del contrato de errores
// en su forma de bloqueo: si otra invocación retiene el bloqueo más tiempo que
// la espera, la operación falla como «inesperado» (1) diciendo que la base está
// bloqueada, nombrando la ruta y la espera que se agotó, y nunca que el fichero
// no sirve. La espera se acorta con la opción de las pruebas para no esperar cinco
// segundos; el tiempo que tarda demuestra que se esperó de verdad.
func TestBloqueoQueNoSeSueltaAgotaLaEspera(t *testing.T) {
	t.Parallel()

	operaciones := []struct {
		nombre  string
		prepara func(t *testing.T, directorio string) func(ctx context.Context) error
		nombra  []string
	}{
		{
			nombre: "New",
			prepara: func(t *testing.T, directorio string) func(ctx context.Context) error {
				t.Helper()

				return func(ctx context.Context) error {
					cliente, err := New(ctx, ConDirectorio(directorio), conEsperaAnteBloqueo(esperaCorta))
					assert.Nil(t, cliente, "un fallo al construir no devuelve ningún cliente")

					return err
				}
			},
			nombra: []string{"bloqueada", esperaCorta.String()},
		},
		{
			nombre: "Put",
			prepara: func(t *testing.T, directorio string) func(ctx context.Context) error {
				t.Helper()

				cliente := clienteAbierto(t, directorio, conEsperaAnteBloqueo(esperaCorta))

				return func(ctx context.Context) error {
					return cliente.Put(ctx, claveDePrueba, []byte("<norma>contenido</norma>"), vigenciaDePrueba)
				}
			},
			nombra: []string{"bloqueada", esperaCorta.String(), claveDePrueba},
		},
	}

	for _, operacion := range operaciones {
		t.Run(operacion.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ruta := filepath.Join(directorio, ficheroDeLaBase)
			opera := operacion.prepara(t, directorio)
			retieneElBloqueo(t, ruta)

			inicio := time.Now()
			err := opera(t.Context())
			tardo := time.Since(inicio)

			require.Error(t, err)
			assert.Equal(t, schema.ClaseInesperado, cli.Clasificar(err))
			assert.Equal(t, 1, cli.CodigoSalida(err), "%v", err)
			assert.NotContains(t, err.Error(), "no es una base de datos utilizable",
				"un bloqueo no convierte el fichero en inutilizable (FR-028, FR-035)")

			for _, dato := range append([]string{ruta}, operacion.nombra...) {
				assert.Contains(t, err.Error(), dato, "el mensaje nombra lo que explica el fallo")
			}

			var delControlador *sqlite.Error

			require.ErrorAs(t, err, &delControlador, "la causa es la del controlador")
			assert.Equal(t, sqlite3.SQLITE_BUSY, delControlador.Code()&0xff, "y es un bloqueo")

			assert.GreaterOrEqual(t, tardo, esperaCorta, "se esperó la espera entera antes de rendirse")
			assert.Less(t, tardo, tiempoDeSobra, "y no la de cinco segundos")
		})
	}
}

// retieneElBloqueo abre otra conexión sobre la base —la de otra invocación— y
// empieza en ella una transacción inmediata, que retiene el bloqueo de escritura
// hasta que se llame a lo que devuelve o hasta que termine la prueba. Sobre un
// fichero que no existe lo crea el motor, en diario WAL, sin migrar: es la base
// «sin migrar» sobre la que New tiene que esperar para aplicar el esquema.
func retieneElBloqueo(t *testing.T, ruta string) (suelta func()) {
	t.Helper()

	otra := abreDirectamente(t, dsnNormal(ruta))

	tx, err := otra.BeginTx(t.Context(), nil)
	require.NoError(t, err, "la otra invocación toma el bloqueo de escritura")

	var una sync.Once

	suelta = func() {
		una.Do(func() {
			if err := tx.Rollback(); !errors.Is(err, sql.ErrTxDone) {
				assert.NoError(t, err)
			}

			require.NoError(t, otra.Close())
		})
	}

	t.Cleanup(suelta)

	return suelta
}
