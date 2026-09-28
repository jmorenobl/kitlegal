package graph

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestCadenasDeLectura fija las tres cadenas de conexión de la lectura, carácter
// a carácter (contracts/almacen-world-db.md §3, paso 4; research.md D10), y que
// la ruta llega al motor protegida: «%», «?» y «#» son lo único que la sintaxis
// de URI de SQLite interpreta dentro del camino.
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
			nombre:   "sin auxiliares y con permiso: lee sin crear ningún auxiliar",
			ruta:     ruta,
			modo:     modoSinAuxiliares,
			esperada: camino + "?mode=rw&_pragma=busy_timeout(100)&_pragma=query_only(1)",
		},
		{
			nombre:   "sin permiso y sin -wal ni -journal: inmutable, sin bloqueos ni ficheros",
			ruta:     ruta,
			modo:     modoInmutable,
			esperada: camino + "?mode=ro&immutable=1&_pragma=query_only(1)",
		},
		{
			nombre:   "en otro caso: solo lectura, que lee el WAL y no escribe world.db",
			ruta:     ruta,
			modo:     modoConAuxiliares,
			esperada: camino + "?mode=ro&_pragma=busy_timeout(100)&_pragma=query_only(1)",
		},
		{
			nombre:   "un modo que no es de los tres es el que nunca escribe world.db",
			ruta:     ruta,
			modo:     modoDeLectura(0),
			esperada: camino + "?mode=ro&_pragma=busy_timeout(100)&_pragma=query_only(1)",
		},
		{
			nombre: "la ruta con %, ? y # llega protegida",
			ruta:   filepath.Join(string(filepath.Separator)+"c%41?d#e", "world.db"),
			modo:   modoSinAuxiliares,
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
	crearGrafo(t, directorio, diarioWAL, muestra(t))

	antes := huellasDelArbol(t, raiz)

	lectura, err := Leer(t.Context(), ConDirectorio(directorio))
	require.NoError(t, err)

	recuento, err := lectura.Recuento(t.Context())
	require.NoError(t, err)
	assert.Equal(t, recuentoDeLaMuestra(), recuento, "lee la base que está en la ruta")
	require.NoError(t, lectura.Close())

	assert.Equal(t, antes, huellasDelArbol(t, raiz), "no aparece ningún fichero en otro sitio")
}

// TestRutaDeLosAuxiliares fija con qué ruta nombra SQLite los auxiliares de
// world.db (contracts/almacen-world-db.md §3, paso 2; research.md D10, V47): fuera
// de Windows, la resuelta por filepath.EvalSymlinks, porque SQLite sigue los
// enlaces y los crea junto al destino; en Windows, la ruta tal cual, porque allí
// no los sigue. El sistema entra como argumento para poder fijar las dos ramas
// desde cualquier sistema.
func TestRutaDeLosAuxiliares(t *testing.T) {
	t.Parallel()

	raiz := t.TempDir()
	otro := filepath.Join(raiz, "otro")
	deLaCache := filepath.Join(raiz, "cache")
	require.NoError(t, os.Mkdir(otro, 0o700))
	require.NoError(t, os.Mkdir(deLaCache, 0o700))

	destino := filepath.Join(otro, "grafo.db")
	require.NoError(t, os.WriteFile(destino, []byte("base"), 0o600))

	plano := filepath.Join(deLaCache, "plano.db")
	require.NoError(t, os.WriteFile(plano, []byte("base"), 0o600))

	enlace := filepath.Join(deLaCache, "world.db")
	require.NoError(t, os.Symlink(destino, enlace))

	cadena := filepath.Join(deLaCache, "cadena.db")
	require.NoError(t, os.Symlink(enlace, cadena))

	roto := filepath.Join(deLaCache, "roto.db")
	require.NoError(t, os.Symlink(filepath.Join(raiz, "no-existe.db"), roto))

	directorioEnlazado := filepath.Join(raiz, "enlazado")
	require.NoError(t, os.Symlink(otro, directorioEnlazado))

	destinoResuelto := resuelta(t, destino)

	casos := []struct {
		nombre, sistema, ruta, esperada string
		ausente                         bool
	}{
		{nombre: "un fichero sin enlaces", sistema: "linux", ruta: plano, esperada: resuelta(t, plano)},
		{nombre: "un enlace, fuera de Windows", sistema: "darwin", ruta: enlace, esperada: destinoResuelto},
		{nombre: "una cadena de enlaces", sistema: "linux", ruta: cadena, esperada: destinoResuelto},
		{
			nombre:   "un directorio enlazado en la ruta",
			sistema:  "linux",
			ruta:     filepath.Join(directorioEnlazado, "grafo.db"),
			esperada: destinoResuelto,
		},
		{nombre: "un enlace sin destino, fuera de Windows", sistema: "linux", ruta: roto, ausente: true},
		{nombre: "un enlace, en Windows", sistema: "windows", ruta: enlace, esperada: enlace},
		{nombre: "un enlace sin destino, en Windows", sistema: "windows", ruta: roto, esperada: roto},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			base, err := rutaDeLosAuxiliares(caso.ruta, caso.sistema)

			if caso.ausente {
				require.Error(t, err)
				assert.True(t, ausente(err), "un enlace sin destino es un world.db ausente: %v", err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, caso.esperada, base)
		})
	}

	t.Run("el sistema de la máquina", func(t *testing.T) {
		t.Parallel()

		base, err := rutaDeLosAuxiliares(enlace, runtime.GOOS)
		require.NoError(t, err)

		if runtime.GOOS == "windows" {
			assert.Equal(t, enlace, base)
		} else {
			assert.Equal(t, destinoResuelto, base)
		}
	})
}

// TestBuscarAuxiliares fija la búsqueda de los tres auxiliares posibles junto a
// la ruta que les da nombre: cada combinación de -wal, -shm y -journal, también
// de 0 bytes, que siguen siendo ficheros que existen; un componente que no es
// directorio, que es no encontrarlos; y un fallo que no dice si están, que no se
// confunde con su ausencia.
func TestBuscarAuxiliares(t *testing.T) {
	t.Parallel()

	for bits := range 8 {
		esperados := auxiliares{
			wal:               bits&1 != 0,
			memoriaCompartida: bits&2 != 0,
			diario:            bits&4 != 0,
		}

		t.Run(strconv.Itoa(bits), func(t *testing.T) {
			t.Parallel()

			base := filepath.Join(t.TempDir(), "grafo.db")

			for sufijo, presente := range map[string]bool{
				sufijoWAL:               esperados.wal,
				sufijoMemoriaCompartida: esperados.memoriaCompartida,
				sufijoDiario:            esperados.diario,
			} {
				if presente {
					require.NoError(t, os.WriteFile(base+sufijo, nil, 0o600))
				}
			}

			encontrados, err := buscarAuxiliares(base)
			require.NoError(t, err)
			assert.Equal(t, esperados, encontrados)
			assert.Equal(t, bits != 0, encontrados.alguno())
		})
	}

	t.Run("un componente que no es directorio: ninguno", func(t *testing.T) {
		t.Parallel()

		fichero := filepath.Join(t.TempDir(), "fichero")
		require.NoError(t, os.WriteFile(fichero, []byte("no soy un directorio"), 0o600))

		encontrados, err := buscarAuxiliares(filepath.Join(fichero, "world.db"))
		require.NoError(t, err)
		assert.Equal(t, auxiliares{}, encontrados)
	})

	t.Run("un fallo que no dice si están", func(t *testing.T) {
		t.Parallel()

		// Un nombre de 252 bytes cabe en el sistema, pero con el sufijo pasa de
		// los 255 que admite un componente: os.Stat no puede decir si existe.
		base := filepath.Join(t.TempDir(), strings.Repeat("a", 252))

		_, err := os.Stat(base + sufijoWAL)
		require.ErrorIs(t, err, syscall.ENAMETOOLONG, "premisa: el sufijo hace el nombre demasiado largo")

		encontrados, err := buscarAuxiliares(base)
		require.ErrorIs(t, err, syscall.ENAMETOOLONG)
		assert.Equal(t, auxiliares{}, encontrados)
	})
}

// TestComprobarEscritura fija la comprobación de si el proceso puede escribir
// world.db (contracts/almacen-world-db.md §3, paso 3, y §4, paso 4; research.md
// V46): abrirlo para leer y escribir y cerrarlo, sin leer, escribir ni truncar,
// que no cambia sus bytes ni su fecha de modificación. Distingue las tres salidas
// que el contrato separa: nil, puede; fs.ErrNotExist, ya no está; cualquier otro
// error, no puede.
func TestComprobarEscritura(t *testing.T) {
	t.Parallel()

	const (
		puede = iota
		yaNoEsta
		noPuede
	)

	casos := []struct {
		nombre   string
		preparar func(t *testing.T, directorio string) string
		salida   int
	}{
		{
			nombre: "un fichero que se puede escribir",
			preparar: func(t *testing.T, directorio string) string {
				t.Helper()

				return ficheroFechado(t, directorio, 0o600)
			},
			salida: puede,
		},
		{
			nombre: "un fichero sin permiso de escritura",
			preparar: func(t *testing.T, directorio string) string {
				t.Helper()

				return ficheroFechado(t, directorio, 0o400)
			},
			salida: noPuede,
		},
		{
			nombre: "un directorio",
			preparar: func(t *testing.T, directorio string) string {
				t.Helper()

				ruta := filepath.Join(directorio, "world.db")
				require.NoError(t, os.Mkdir(ruta, 0o700))

				return ruta
			},
			salida: noPuede,
		},
		{
			nombre: "un fichero que no existe",
			preparar: func(_ *testing.T, directorio string) string {
				return filepath.Join(directorio, "world.db")
			},
			salida: yaNoEsta,
		},
		{
			nombre: "un enlace sin destino",
			preparar: func(t *testing.T, directorio string) string {
				t.Helper()

				ruta := filepath.Join(directorio, "world.db")
				require.NoError(t, os.Symlink(filepath.Join(directorio, "no-existe.db"), ruta))

				return ruta
			},
			salida: yaNoEsta,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ruta := caso.preparar(t, directorio)
			antes := huellasDelArbol(t, directorio)
			fechas := fechasDelArbol(t, directorio)

			err := comprobarEscritura(ruta)

			switch caso.salida {
			case puede:
				require.NoError(t, err)
			case yaNoEsta:
				require.ErrorIs(t, err, fs.ErrNotExist)
			default:
				require.Error(t, err, "sin permiso de escritura, abrir para escribir falla "+
					"(¿el proceso corre como root?)")
				require.NotErrorIs(t, err, fs.ErrNotExist, "no poder escribir no es no existir")
			}

			assert.Equal(t, antes, huellasDelArbol(t, directorio), "comprobar no cambia ni crea nada")
			assert.Equal(t, fechas, fechasDelArbol(t, directorio), "ni la fecha de modificación")
		})
	}

	t.Run("sin permiso, la causa es el acceso denegado", func(t *testing.T) {
		t.Parallel()

		err := comprobarEscritura(ficheroFechado(t, t.TempDir(), 0o400))
		require.ErrorIs(t, err, fs.ErrPermission, "(¿el proceso corre como root?)")
	})
}

// TestModoPara fija el modo de apertura según los auxiliares y el permiso de
// escritura, en las dieciséis combinaciones (contracts/almacen-world-db.md §3,
// paso 4; research.md D10): con permiso y sin ningún auxiliar, el modo que no
// deja ninguno; sin permiso y sin -wal ni -journal —un -shm suelto no guarda
// nada confirmado—, inmutable; en cualquier otro caso, solo lectura.
func TestModoPara(t *testing.T) {
	t.Parallel()

	var (
		ninguno    = auxiliares{}
		wal        = auxiliares{wal: true}
		shm        = auxiliares{memoriaCompartida: true}
		walYShm    = auxiliares{wal: true, memoriaCompartida: true}
		diario     = auxiliares{diario: true}
		diarioYWal = auxiliares{diario: true, wal: true}
		diarioYShm = auxiliares{diario: true, memoriaCompartida: true}
		losTres    = auxiliares{wal: true, memoriaCompartida: true, diario: true}
		conPermiso = true
		sinPermiso = false
		filas      = []struct {
			puede      bool
			auxiliares auxiliares
			modo       modoDeLectura
		}{
			{conPermiso, ninguno, modoSinAuxiliares},
			{conPermiso, wal, modoConAuxiliares},
			{conPermiso, shm, modoConAuxiliares},
			{conPermiso, walYShm, modoConAuxiliares},
			{conPermiso, diario, modoConAuxiliares},
			{conPermiso, diarioYWal, modoConAuxiliares},
			{conPermiso, diarioYShm, modoConAuxiliares},
			{conPermiso, losTres, modoConAuxiliares},
			{sinPermiso, ninguno, modoInmutable},
			{sinPermiso, wal, modoConAuxiliares},
			{sinPermiso, shm, modoInmutable},
			{sinPermiso, walYShm, modoConAuxiliares},
			{sinPermiso, diario, modoConAuxiliares},
			{sinPermiso, diarioYWal, modoConAuxiliares},
			{sinPermiso, diarioYShm, modoConAuxiliares},
			{sinPermiso, losTres, modoConAuxiliares},
		}
	)

	for _, fila := range filas {
		assert.Equal(t, fila.modo, modoPara(fila.puede, fila.auxiliares),
			"puede escribir: %t; auxiliares: %+v", fila.puede, fila.auxiliares)
	}
}

// casoDeApertura es una fila de TestDecidirApertura: cómo se prepara world.db
// dentro del directorio de la caché y lo que la lectura decide antes de abrir
// SQLite, o el fallo.
type casoDeApertura struct {
	nombre string
	// preparar deja el estado bajo la raíz y devuelve el directorio de la caché.
	preparar func(t *testing.T, raiz string) string
	esperada apertura
	// directorio es el mensaje del fallo de un world.db que es un directorio.
	directorio bool
}

// TestDecidirApertura fija lo que la lectura decide antes de abrir SQLite para
// cada estado de world.db (contracts/almacen-world-db.md §3, pasos 1 a 4;
// data-model §3.1): ausente —también sin su directorio, bajo un componente que no
// es directorio o como enlace sin destino— y de 0 bytes —haya los auxiliares que
// haya y aunque no se pueda escribir—, grafo vacío; un directorio, inutilizable;
// y, en los demás, el modo según los auxiliares, buscados donde SQLite los
// nombra, y el permiso de escritura. Decidir no crea, no cambia ni retira nada.
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
		{nombre: "un enlace sin destino", preparar: conEnlaceSinDestino, esperada: vacia},
		{nombre: "un directorio", preparar: conDirectorioEnSuLugar, directorio: true},
		{nombre: "de 0 bytes", preparar: conBase(nil, 0o600), esperada: vacia},
		{
			nombre:   "de 0 bytes con -wal, -shm y -journal",
			preparar: conBase(nil, 0o600, sufijoWAL, sufijoMemoriaCompartida, sufijoDiario),
			esperada: vacia,
		},
		{nombre: "de 0 bytes, sin permiso y con -wal", preparar: conBase(nil, 0o400, sufijoWAL), esperada: vacia},
		{nombre: "con permiso y sin auxiliares", preparar: conBase(contenido, 0o600), esperada: abrirEn(modoSinAuxiliares)},
		{nombre: "sin permiso y sin auxiliares", preparar: conBase(contenido, 0o400), esperada: abrirEn(modoInmutable)},
		{
			nombre:   "sin permiso y con un -shm suelto",
			preparar: conBase(contenido, 0o400, sufijoMemoriaCompartida),
			esperada: abrirEn(modoInmutable, sufijoMemoriaCompartida),
		},
		{
			nombre:   "sin permiso y con -wal",
			preparar: conBase(contenido, 0o400, sufijoWAL),
			esperada: abrirEn(modoConAuxiliares, sufijoWAL),
		},
		{
			nombre:   "sin permiso y con -journal",
			preparar: conBase(contenido, 0o400, sufijoDiario),
			esperada: abrirEn(modoConAuxiliares, sufijoDiario),
		},
		{
			nombre:   "con permiso y con -wal",
			preparar: conBase(contenido, 0o600, sufijoWAL),
			esperada: abrirEn(modoConAuxiliares, sufijoWAL),
		},
		{
			nombre:   "con permiso y con un -shm suelto",
			preparar: conBase(contenido, 0o600, sufijoMemoriaCompartida),
			esperada: abrirEn(modoConAuxiliares, sufijoMemoriaCompartida),
		},
		{
			nombre:   "con permiso y con -journal",
			preparar: conBase(contenido, 0o600, sufijoDiario),
			esperada: abrirEn(modoConAuxiliares, sufijoDiario),
		},
		{nombre: "un enlace a una base sin auxiliares", preparar: conEnlace(), esperada: abrirEn(modoSinAuxiliares)},
		{
			nombre:   "un enlace a una base con -wal junto al destino",
			preparar: conEnlace(sufijoWAL),
			esperada: abrirEn(modoConAuxiliares, sufijoWAL),
		},
		{
			nombre:   "un enlace con un -wal junto a su nombre y no junto al destino",
			preparar: conEnlaceYAuxiliarJuntoAlNombre,
			esperada: abrirEn(modoSinAuxiliares),
		},
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
	ruta := filepath.Join(directorio, "world.db")
	antes := huellasDelArbol(t, raiz)

	decidida, err := decidirApertura(ruta)

	assert.Equal(t, antes, huellasDelArbol(t, raiz), "decidir no crea, no cambia ni retira nada")

	if !caso.directorio {
		require.NoError(t, err)
		assert.Equal(t, caso.esperada, decidida)

		return
	}

	var fallo *Error

	require.ErrorAs(t, err, &fallo)
	assert.Equal(t, schema.ClaseInesperado, fallo.Clase())
	assert.Equal(t, operacionLeer, fallo.Operacion)
	assert.Equal(t, ruta, fallo.Ruta)
	assert.Equal(t, "grafo: "+strconv.Quote(ruta)+
		" es un directorio y no una base de datos utilizable; no se modifica", err.Error())
}

// contenido es lo que llevan las bases de TestDecidirApertura, que no abre
// SQLite: basta con que no sean de 0 bytes.
var contenido = []byte("no hace falta que sea una base: decidir no la abre")

// abrirEn es la decisión de abrir world.db en el modo dado, con los auxiliares
// encontrados.
func abrirEn(modo modoDeLectura, sufijos ...string) apertura {
	decidida := apertura{modo: modo}

	for _, sufijo := range sufijos {
		switch sufijo {
		case sufijoWAL:
			decidida.auxiliares.wal = true
		case sufijoMemoriaCompartida:
			decidida.auxiliares.memoriaCompartida = true
		case sufijoDiario:
			decidida.auxiliares.diario = true
		}
	}

	return decidida
}

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

// conEnlaceSinDestino es world.db como enlace a algo que no existe.
func conEnlaceSinDestino(t *testing.T, raiz string) string {
	t.Helper()

	directorio := cacheVacia(t, raiz)
	require.NoError(t, os.Symlink(filepath.Join(raiz, "no-existe.db"), filepath.Join(directorio, "world.db")))

	return directorio
}

// conDirectorioEnSuLugar es world.db que es un directorio.
func conDirectorioEnSuLugar(t *testing.T, raiz string) string {
	t.Helper()

	directorio := cacheVacia(t, raiz)
	require.NoError(t, os.Mkdir(filepath.Join(directorio, "world.db"), 0o700))

	return directorio
}

// conBase prepara world.db con los bytes y los permisos dados, y un auxiliar
// vacío por cada sufijo.
func conBase(datos []byte, permisos os.FileMode, sufijos ...string) func(*testing.T, string) string {
	return func(t *testing.T, raiz string) string {
		t.Helper()

		directorio := cacheVacia(t, raiz)
		ruta := filepath.Join(directorio, "world.db")

		for _, sufijo := range sufijos {
			require.NoError(t, os.WriteFile(ruta+sufijo, []byte("auxiliar"), 0o600))
		}

		require.NoError(t, os.WriteFile(ruta, datos, 0o600))
		require.NoError(t, os.Chmod(ruta, permisos))

		return directorio
	}
}

// conEnlace prepara world.db como enlace a una base de otro directorio, con un
// auxiliar junto al destino por cada sufijo.
func conEnlace(sufijos ...string) func(*testing.T, string) string {
	return func(t *testing.T, raiz string) string {
		t.Helper()

		otro := filepath.Join(raiz, "otro")
		require.NoError(t, os.Mkdir(otro, 0o700))

		destino := filepath.Join(otro, "grafo.db")
		require.NoError(t, os.WriteFile(destino, contenido, 0o600))

		for _, sufijo := range sufijos {
			require.NoError(t, os.WriteFile(destino+sufijo, []byte("auxiliar"), 0o600))
		}

		directorio := cacheVacia(t, raiz)
		require.NoError(t, os.Symlink(destino, filepath.Join(directorio, "world.db")))

		return directorio
	}
}

// conEnlaceYAuxiliarJuntoAlNombre es world.db como enlace a una base sin
// auxiliares, con un world.db-wal junto al enlace: SQLite no lo ve, porque nombra
// los auxiliares por el destino (research.md V47).
func conEnlaceYAuxiliarJuntoAlNombre(t *testing.T, raiz string) string {
	t.Helper()

	directorio := conEnlace()(t, raiz)
	require.NoError(t, os.WriteFile(filepath.Join(directorio, "world.db"+sufijoWAL), []byte("ajeno"), 0o600))

	return directorio
}

// ficheroFechado crea world.db con los permisos dados y una fecha de
// modificación en el pasado, para que un cambio de fecha se vea.
func ficheroFechado(t *testing.T, directorio string, permisos os.FileMode) string {
	t.Helper()

	ruta := filepath.Join(directorio, "world.db")
	require.NoError(t, os.WriteFile(ruta, contenido, 0o600))

	pasado := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	require.NoError(t, os.Chtimes(ruta, pasado, pasado))
	require.NoError(t, os.Chmod(ruta, permisos))

	return ruta
}

// fechasDelArbol es la fecha de modificación de cada fichero bajo la raíz.
func fechasDelArbol(t *testing.T, raiz string) map[string]time.Time {
	t.Helper()

	fechas := map[string]time.Time{}

	err := filepath.WalkDir(raiz, func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := entrada.Info()
		if err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			fechas[ruta] = info.ModTime()
		}

		return nil
	})
	require.NoError(t, err)

	return fechas
}

// resuelta es la ruta sin ningún enlace simbólico: la que SQLite usa, fuera de
// Windows, para nombrar los auxiliares.
func resuelta(t *testing.T, ruta string) string {
	t.Helper()

	sinEnlaces, err := filepath.EvalSymlinks(ruta)
	require.NoError(t, err)

	return sinEnlaces
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
