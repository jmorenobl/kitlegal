package cache

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestNewRechazaOpcionesInvalidas fija la fila 3 de la tabla cerrada del
// contrato de errores §3: una opción con un valor que no sirve es «argumentos»
// (2) y el mensaje nombra **la opción**, porque lo que hay que corregir está en
// el código que la pasa y no en lo que la persona escribió (D2).
//
// Fija además que las opciones se validan **en orden** y que la primera
// inválida es la que decide: la última fila pasa dos opciones inválidas y exige
// que el mensaje nombre la primera y no la segunda. Sin eso, validarlas en
// cualquier orden pasaría igual y el fallo señalaría un sitio que quien lo lee
// no tocó.
//
// El control positivo de la cabecera es lo que impide que la tabla pase con un
// New que fallara siempre: con las cuatro opciones bien construidas no falla, y
// el cliente que devuelve se cierra —dos veces, porque Close es idempotente
// (FR-004)— sin error.
func TestNewRechazaOpcionesInvalidas(t *testing.T) {
	t.Parallel()

	cliente, err := New(t.Context(),
		ConDirectorio(t.TempDir()),
		SoloLectura(),
		ConReloj(time.Now),
		ConRegistrador(slog.New(slog.DiscardHandler)))
	require.NoError(t, err)
	require.NotNil(t, cliente)
	require.NoError(t, cliente.Close())
	require.NoError(t, cliente.Close(),
		"cerrar dos veces no es un fallo: el defer de quien lo construyó y el cierre "+
			"explícito conviven (FR-004)")

	casos := []struct {
		nombre   string
		opciones []Opcion
		nombra   string
		noNombra string
	}{
		{
			nombre:   "un directorio sin ninguna ruta dentro",
			opciones: []Opcion{ConDirectorio("")},
			nombra:   "ConDirectorio",
		},
		{
			nombre:   "un reloj nulo",
			opciones: []Opcion{ConReloj(nil)},
			nombra:   "ConReloj",
		},
		{
			nombre:   "un registrador nulo",
			opciones: []Opcion{ConRegistrador(nil)},
			nombra:   "ConRegistrador",
		},
		{
			nombre:   "con dos inválidas, la que se nombra es la primera",
			opciones: []Opcion{ConReloj(nil), ConRegistrador(nil)},
			nombra:   "ConReloj",
			noNombra: "ConRegistrador",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			cliente, err := New(t.Context(), caso.opciones...)

			require.Error(t, err)
			assert.Nil(t, cliente)
			assert.Equal(t, schema.ClaseArgumentos, cli.Clasificar(err))
			assert.Equal(t, 2, cli.CodigoSalida(err))
			assert.Contains(t, err.Error(), "opción "+caso.nombra,
				"el mensaje nombra la opción que hay que corregir")

			if caso.noNombra != "" {
				assert.NotContains(t, err.Error(), caso.noNombra,
					"las opciones se validan en orden: la segunda inválida no se llega a mirar")
			}
		})
	}
}

// TestCloseEsIdempotente fija FR-004: la primera llamada cierra y las
// siguientes no hacen nada y no fallan. No es una comodidad, es lo que hace que
// el defer de quien construyó el cliente y un cierre explícito antes de tiempo
// convivan sin que el segundo parezca un error.
//
// Las dos filas son las dos formas de cliente que existen: la que abrió una base
// de datos y la que no llegó a abrir ninguna, porque una invocación de solo
// lectura no encontró cache.db. En la segunda no hay nada que cerrar, y cerrarla
// tampoco puede fallar.
func TestCloseEsIdempotente(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		opciones []Opcion
	}{
		{nombre: "con base de datos"},
		{nombre: "sin base de datos", opciones: []Opcion{SoloLectura()}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			// El directorio está vacío: en modo normal la base se crea, y en
			// solo lectura el cliente queda sin ninguna.
			opciones := append([]Opcion{ConDirectorio(t.TempDir())}, caso.opciones...)

			cliente, err := New(t.Context(), opciones...)
			require.NoError(t, err)

			assert.NoError(t, cliente.Close())
			assert.NoError(t, cliente.Close(), "cerrar dos veces no es un fallo (FR-004)")
			assert.NoError(t, cliente.Close(), "ni tres")
		})
	}
}

// TestOperacionTrasCierre fija la otra mitad de FR-004 y de la fila 14 del
// contrato de errores: después de Close ninguna operación lee ni escribe, las
// dos son «inesperado» (1) y el disco queda como lo dejó el cierre, sin que
// reaparezca ningún auxiliar ni cambie un byte.
//
// Las dos filas son las dos formas de cliente. En la que tiene base, la entrada
// que se pide está guardada y vigente, de modo que servirla sería posible si el
// cierre no se comprobara. En la que no tiene base —solo lectura sobre un
// directorio vacío—, leer tras cerrar es 1 y no la ausencia (4) que el mismo
// cliente devolvería abierto.
func TestOperacionTrasCierre(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre     string
		conEntrada bool
		opciones   []Opcion
	}{
		{nombre: "con base de datos", conEntrada: true},
		{nombre: "sin base de datos", opciones: []Opcion{SoloLectura()}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			if caso.conEntrada {
				siembra(t, directorio, claveDePrueba, []byte("<norma>contenido</norma>"))
			}

			cliente := clienteAbierto(t, directorio, caso.opciones...)
			require.NoError(t, cliente.Close())

			antes := arbolDe(t, directorio)

			contenido, presente, err := cliente.Get(t.Context(), claveDePrueba)
			require.Error(t, err)
			assert.False(t, presente)
			assert.Nil(t, contenido)
			assert.Equal(t, schema.ClaseInesperado, cli.Clasificar(err))
			assert.Equal(t, 1, cli.CodigoSalida(err))
			assert.Contains(t, err.Error(), "cliente cerrado", "el mensaje dice por qué no se lee")
			assert.Contains(t, err.Error(), claveDePrueba, "…y qué clave se pedía")

			err = cliente.Put(t.Context(), claveDePrueba, []byte("<norma>otra</norma>"), vigenciaDePrueba)
			require.Error(t, err)
			assert.Equal(t, schema.ClaseInesperado, cli.Clasificar(err))
			assert.Equal(t, 1, cli.CodigoSalida(err))
			assert.Contains(t, err.Error(), claveDePrueba)

			assert.Equal(t, antes, arbolDe(t, directorio), "después de cerrar no se toca el disco")
		})
	}
}

// TestNewPorOmisionUsaElDirectorioDeLaCuenta fija FR-019 y la última
// precedencia de FR-023: sin opción y sin variable, la caché vive en
// <cuenta>/.cache/kitlegal y en ningún otro sitio.
//
// «En ningún otro sitio» es la mitad que de verdad protege algo: un descuido en
// la resolución de la ruta escribiría en el directorio de trabajo, o en la raíz
// de la cuenta, y la base aparecería igual donde se la busca. Por eso el test
// enumera lo que hay bajo la cuenta y exige que no haya nada más que .cache,
// dentro nada más que kitlegal, y dentro cache.db con —como mucho— los dos
// auxiliares que SQLite pone a su lado.
//
// El directorio de la cuenta se redirige con t.Setenv: ningún test del hito toca
// la caché real de quien lo ejecuta (SC-004, SC-009, FR-040).
func TestNewPorOmisionUsaElDirectorioDeLaCuenta(t *testing.T) {
	cuenta := t.TempDir()
	t.Setenv("HOME", cuenta)
	t.Setenv("USERPROFILE", cuenta)
	sinVariable(t)

	cliente, err := New(t.Context())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, cliente.Close()) })

	directorio := filepath.Join(cuenta, ".cache", "kitlegal")
	require.FileExists(t, filepath.Join(directorio, ficheroDeLaBase),
		"la base por omisión vive en <cuenta>/.cache/kitlegal/cache.db (FR-019, FR-020)")

	assert.Equal(t, []string{".cache"}, contenidoDe(t, cuenta),
		"bajo la cuenta no aparece nada más que .cache")
	assert.Equal(t, []string{"kitlegal"}, contenidoDe(t, filepath.Join(cuenta, ".cache")),
		"dentro de .cache, nada más que kitlegal")

	auxiliares := []string{ficheroDeLaBase, ficheroDeLaBase + "-wal", ficheroDeLaBase + "-shm"}
	for _, nombre := range contenidoDe(t, directorio) {
		assert.Contains(t, auxiliares, nombre,
			"en el directorio de la caché solo están la base y sus auxiliares")
	}
}

// TestClienteDesdeVariasGoroutines fija lo que el comentario de Cliente promete
// y el contrato del puerto y el cliente §7 declara: un mismo cliente se usa
// desde varias goroutines a la vez sin carreras y sin que una llamada responda
// por otra (D10). La conexión única serializa las sentencias y el cerrojo
// protege el estado de cierre; que eso baste lo dice el detector de carreras con
// el que make test ejecuta la suite, y no la lectura del código.
//
// Cada goroutine alterna Get y Put sobre sus propias claves: lee la que todavía
// no está —ausencia—, la guarda y la vuelve a leer exigiendo exactamente lo que
// acaba de guardar. Las claves son distintas a propósito: con una compartida no
// se sabría qué respuesta es la correcta, y lo que aquí se mide es que ninguna
// llamada se cruza con otra. Al terminar, todas siguen ahí y cada una con lo
// suyo (SC-010).
//
// El registrador es de nivel debug y descarta lo que escribe: así cada lectura y
// cada escritura pasan también por el camino que anota el resultado, que con el
// registrador por omisión no se llega a ejecutar (D14).
func TestClienteDesdeVariasGoroutines(t *testing.T) {
	t.Parallel()

	const (
		goroutines  = 8
		operaciones = 16
	)

	registrador := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	cliente := clienteAbierto(t, t.TempDir(), ConRegistrador(registrador))

	resultados := make(chan error, goroutines)

	var grupo sync.WaitGroup

	for goroutine := range goroutines {
		grupo.Go(func() { resultados <- alternaGetYPut(t.Context(), cliente, goroutine, operaciones) })
	}

	grupo.Wait()
	close(resultados)

	for err := range resultados {
		require.NoError(t, err, "un cliente usado desde varias goroutines responde a cada una lo suyo (D10)")
	}

	for goroutine := range goroutines {
		for operacion := range operaciones {
			contenido, presente, err := cliente.Get(t.Context(), claveDeGoroutine(goroutine, operacion))
			require.NoError(t, err)
			require.True(t, presente, "lo que guardó cada goroutine sigue guardado")
			assert.Equal(t, cuerpoDeGoroutine(goroutine, operacion), contenido,
				"cada clave con lo suyo: ninguna responde por otra (SC-010)")
		}
	}
}

// alternaGetYPut es el trabajo de una goroutine de TestClienteDesdeVariasGoroutines:
// para cada operación lee su clave, que todavía no está; la guarda; y la vuelve a
// leer exigiendo exactamente lo guardado. Devuelve el primer desvío, porque desde
// una goroutine que no es la de la prueba no se puede terminar la prueba.
func alternaGetYPut(ctx context.Context, cliente *Cliente, goroutine, operaciones int) error {
	for operacion := range operaciones {
		clave := claveDeGoroutine(goroutine, operacion)
		cuerpo := cuerpoDeGoroutine(goroutine, operacion)

		_, presente, err := cliente.Get(ctx, clave)

		switch {
		case err != nil:
			return fmt.Errorf("leer %q antes de guardarla: %w", clave, err)
		case presente:
			return fmt.Errorf("%q está presente antes de que nadie la guarde: otra llamada respondió por ella", clave)
		}

		if err := cliente.Put(ctx, clave, cuerpo, vigenciaDePrueba); err != nil {
			return fmt.Errorf("guardar %q: %w", clave, err)
		}

		leido, presente, err := cliente.Get(ctx, clave)

		switch {
		case err != nil:
			return fmt.Errorf("leer %q después de guardarla: %w", clave, err)
		case !presente:
			return fmt.Errorf("%q no está después de guardarla", clave)
		case !bytes.Equal(cuerpo, leido):
			return fmt.Errorf("%q no devuelve lo que se guardó: %q en vez de %q", clave, leido, cuerpo)
		}
	}

	return nil
}

// claveDeGoroutine y cuerpoDeGoroutine son la clave y el contenido de una
// operación de una goroutine: distintos para cada par, de modo que una respuesta
// cruzada no pueda pasar por buena.
func claveDeGoroutine(goroutine, operacion int) string {
	return fmt.Sprintf("%s/goroutine-%d/%d", claveDePrueba, goroutine, operacion)
}

func cuerpoDeGoroutine(goroutine, operacion int) []byte {
	return fmt.Appendf(nil, "<norma goroutine=%d operación=%d>", goroutine, operacion)
}

// contenidoDe enumera, ordenados, los nombres que hay en un directorio.
func contenidoDe(t *testing.T, directorio string) []string {
	t.Helper()

	entradas, err := os.ReadDir(directorio)
	require.NoError(t, err)

	nombres := make([]string, 0, len(entradas))
	for _, entrada := range entradas {
		nombres = append(nombres, entrada.Name())
	}

	return nombres
}
