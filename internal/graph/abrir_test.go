package graph

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCadenasDeLectura fija las dos cadenas de conexión de la lectura, carácter
// a carácter (contracts/almacen-world-db.md §3, paso 2; H7.1 research.md D17), y
// que la ruta llega al motor protegida: «%», «?» y «#» son lo único que la
// sintaxis de URI de SQLite interpreta dentro del camino.
func TestCadenasDeLectura(t *testing.T) {
	t.Parallel()

	ruta := filepath.Join(string(filepath.Separator)+"cache", "world.db")
	camino := "file:" + filepath.ToSlash(ruta)

	casos := []struct {
		nombre   string
		ruta     string
		modo     modoDeLectura
		esperada string
	}{
		{
			nombre:   "sin -wal: lee sin poder escribir y sin dejar ningún auxiliar",
			ruta:     ruta,
			modo:     modoSinWAL,
			esperada: camino + "?mode=rw&_pragma=busy_timeout(100)&_pragma=query_only(1)",
		},
		{
			nombre:   "con -wal: solo lectura, que lee lo confirmado en el WAL",
			ruta:     ruta,
			modo:     modoConWAL,
			esperada: camino + "?mode=ro&_pragma=busy_timeout(100)&_pragma=query_only(1)",
		},
		{
			nombre:   "un modo que no es de los dos es el de solo lectura",
			ruta:     ruta,
			modo:     modoDeLectura(0),
			esperada: camino + "?mode=ro&_pragma=busy_timeout(100)&_pragma=query_only(1)",
		},
		{
			nombre: "la ruta con %, ? y # llega protegida",
			ruta:   filepath.Join(string(filepath.Separator)+"c%41?d#e", "world.db"),
			modo:   modoSinWAL,
			esperada: "file:" + filepath.ToSlash(string(filepath.Separator)) +
				"c%2541%3Fd%23e/world.db?mode=rw&_pragma=busy_timeout(100)&_pragma=query_only(1)",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.esperada, cadenaDeLectura(caso.ruta, caso.modo))
		})
	}
}

// TestLeerRutaConCaracteresDeURI comprueba con el motor que la cadena abre el
// fichero de la ruta aunque su directorio lleve «%41», «?» y «#»: sin protegerlos,
// SQLite abriría otro fichero —el camino cortado en el «?» o con «%41» leído como
// «A»— y lo crearía junto a él.
func TestLeerRutaConCaracteresDeURI(t *testing.T) {
	t.Parallel()

	raiz := t.TempDir()
	directorio := filepath.Join(raiz, "c%41?d#e")
	require.NoError(t, os.Mkdir(directorio, 0o700))
	crearGrafo(t, directorio, muestra(t))

	antes := huellasDelArbol(t, raiz)

	lectura, err := Leer(t.Context(), ConDirectorio(directorio))
	require.NoError(t, err)

	recuento, err := lectura.Recuento(t.Context())
	require.NoError(t, err)
	assert.Equal(t, recuentoDeLaMuestra(), recuento, "lee la base que está en la ruta")
	require.NoError(t, lectura.Close())

	assert.Equal(t, antes, huellasDelArbol(t, raiz), "no aparece ningún fichero en otro sitio")
}

// casoDeApertura es una fila de TestDecidirApertura: cómo se prepara world.db
// dentro del directorio de la caché y lo que la lectura decide antes de abrir
// SQLite.
type casoDeApertura struct {
	nombre string
	// preparar deja el estado bajo la raíz y devuelve el directorio de la caché.
	preparar func(t *testing.T, raiz string) string
	esperada apertura
}

// TestDecidirApertura fija lo que la lectura decide antes de abrir SQLite
// (contracts/almacen-world-db.md §3, pasos 1 y 2; H7.1 research.md D17):
// ausente —también sin su directorio o bajo un componente que no es
// directorio— y de 0 bytes —también con el -wal de una escritura
// interrumpida—, el grafo vacío; en los demás, el modo por la existencia de
// world.db-wal. Decidir no crea, no cambia ni retira nada.
func TestDecidirApertura(t *testing.T) {
	t.Parallel()

	vacia := apertura{vacia: true}

	casos := []casoDeApertura{
		{nombre: "ausente", preparar: cacheVacia, esperada: vacia},
		{
			nombre: "sin su directorio",
			preparar: func(_ *testing.T, raiz string) string {
				return filepath.Join(raiz, "no-existe", "cache")
			},
			esperada: vacia,
		},
		{nombre: "bajo un componente que no es directorio", preparar: bajoUnFichero, esperada: vacia},
		{nombre: "de 0 bytes", preparar: conBase(nil), esperada: vacia},
		{nombre: "de 0 bytes con un -wal", preparar: conBase(nil, sufijoWAL), esperada: vacia},
		{nombre: "sin -wal", preparar: conBase(contenido), esperada: apertura{modo: modoSinWAL}},
		{nombre: "con un -wal", preparar: conBase(contenido, sufijoWAL), esperada: apertura{modo: modoConWAL}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			compruebaApertura(t, caso)
		})
	}
}

// compruebaApertura es el cuerpo de cada fila de TestDecidirApertura.
func compruebaApertura(t *testing.T, caso casoDeApertura) {
	t.Helper()

	raiz := t.TempDir()
	directorio := caso.preparar(t, raiz)
	antes := huellasDelArbol(t, raiz)

	decidida, err := decidirApertura(filepath.Join(directorio, "world.db"))
	require.NoError(t, err)
	assert.Equal(t, caso.esperada, decidida)

	assert.Equal(t, antes, huellasDelArbol(t, raiz), "decidir no crea, no cambia ni retira nada")
}

// contenido es lo que llevan las bases de TestDecidirApertura, que no abre
// SQLite: basta con que no sean de 0 bytes.
var contenido = []byte("no hace falta que sea una base: decidir no la abre")

// cacheVacia es el directorio de la caché sin world.db.
func cacheVacia(t *testing.T, raiz string) string {
	t.Helper()

	directorio := filepath.Join(raiz, "cache")
	require.NoError(t, os.Mkdir(directorio, 0o700))

	return directorio
}

// bajoUnFichero es un directorio de la caché que es un fichero, de modo que
// world.db queda bajo un componente que no es directorio.
func bajoUnFichero(t *testing.T, raiz string) string {
	t.Helper()

	fichero := filepath.Join(raiz, "cache")
	require.NoError(t, os.WriteFile(fichero, []byte("no soy un directorio"), 0o600))

	return fichero
}

// conBase prepara world.db con los bytes dados y, por cada sufijo, un auxiliar
// que no está vacío.
func conBase(datos []byte, sufijos ...string) func(*testing.T, string) string {
	return func(t *testing.T, raiz string) string {
		t.Helper()

		directorio := cacheVacia(t, raiz)
		ruta := filepath.Join(directorio, "world.db")

		for _, sufijo := range sufijos {
			require.NoError(t, os.WriteFile(ruta+sufijo, []byte("auxiliar"), 0o600))
		}

		require.NoError(t, os.WriteFile(ruta, datos, 0o600))

		return directorio
	}
}

// TestAbrirParaLeer fija el paso 3 de la lectura (contracts/almacen-world-db.md
// §3; H7.1 research.md D4): sin esquema —la versión 0—, ninguna conexión, que es
// el grafo vacío; con la versión 1, la de H7, sin la tabla lecturas, y con la
// 2, la conexión abierta y la versión leída. Abrir no migra ni escribe nada.
func TestAbrirParaLeer(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		preparar func(t *testing.T, raiz string) string
		version  int64
	}{
		{nombre: "en WAL y sin esquema", preparar: enWALSinTablas, version: 0},
		{nombre: "la versión 1", preparar: conMuestraDeH7, version: 1},
		{nombre: "la versión 2", preparar: conMuestra, version: 2},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := t.TempDir()
			ruta := filepath.Join(caso.preparar(t, raiz), "world.db")
			antes := huellasDelArbol(t, raiz)

			base, version, err := abrirParaLeer(t.Context(), ruta, modoSinWAL)
			require.NoError(t, err)
			assert.Equal(t, caso.version, version)

			if caso.version == 0 {
				assert.Nil(t, base, "el grafo vacío no deja ninguna conexión abierta")
			} else {
				require.NotNil(t, base)
				assert.Equal(t, caso.version, versionDe(t, base), "la conexión abierta es la de esa base")
				require.NoError(t, base.Close())
			}

			assert.Equal(t, antes, huellasDelArbol(t, raiz), "abrir no migra ni escribe nada")
		})
	}
}

// TestAusente fija qué fallos de os.Stat son un fichero ausente: no existe, o
// un componente de la ruta no es un directorio, que en Unix no es
// fs.ErrNotExist.
func TestAusente(t *testing.T) {
	t.Parallel()

	fichero := filepath.Join(t.TempDir(), "fichero")
	require.NoError(t, os.WriteFile(fichero, contenido, 0o600))

	_, noExiste := os.Stat(filepath.Join(filepath.Dir(fichero), "no-existe"))
	_, noEsDirectorio := os.Stat(filepath.Join(fichero, "world.db"))

	assert.True(t, ausente(noExiste))
	assert.True(t, ausente(noEsDirectorio))
	assert.True(t, ausente(errors.Join(errors.New("envuelto"), noExiste)))
	assert.False(t, ausente(nil))
	assert.False(t, ausente(fs.ErrPermission))
}
