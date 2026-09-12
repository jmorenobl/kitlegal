package cache

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Los dos modos de apertura, escritos una sola vez: cada tabla que mide algo
// que tiene que valer igual se escriba o solo se lea pasa por los dos, y así
// ninguna comprobación queda hecha para un modo y olvidada para el otro
// (FR-015).
var modosDeApertura = []struct {
	nombre   string
	opciones []Opcion
}{
	{nombre: "normal"},
	{nombre: "solo-lectura", opciones: []Opcion{SoloLectura()}},
}

// TestAbrirCreaDirectorioYFicheroConPermisosReservados fija FR-021 sobre el
// disco: el directorio de la caché nace solo para su dueño y el fichero solo
// legible y escribible por él.
//
// Que el fichero lo cree el propio paquete antes de abrirlo con SQLite no es un
// adorno: dejándoselo al controlador nacería a 0644 —legible por cualquiera de
// la máquina— y los auxiliares del registro de escritura heredarían esos
// permisos, porque heredan los del fichero de la base (D4, sonda 1 I).
//
// El directorio no existe cuando empieza el test: crearlo es parte de lo que se
// mide, y sobre un directorio que ya estuviera con otros permisos no se sabría
// si los puso New.
func TestAbrirCreaDirectorioYFicheroConPermisosReservados(t *testing.T) {
	t.Parallel()

	directorio := filepath.Join(t.TempDir(), "kitlegal")

	cliente, err := New(t.Context(), ConDirectorio(directorio))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, cliente.Close()) })

	delDirectorio, err := os.Stat(directorio)
	require.NoError(t, err, "el modo normal crea el directorio que no existe")
	assert.True(t, delDirectorio.IsDir())
	assert.Equal(t, fs.FileMode(0o700), delDirectorio.Mode().Perm(),
		"el directorio de la caché es solo de su dueño (FR-021)")

	delFichero, err := os.Stat(filepath.Join(directorio, ficheroDeLaBase))
	require.NoError(t, err, "el modo normal crea la base de datos")
	assert.Equal(t, fs.FileMode(0o600), delFichero.Mode().Perm(),
		"la base nace con acceso reservado a la cuenta y no con los permisos que "+
			"le pondría SQLite (FR-021, D4)")
}

// TestAbrirAplicaLosPragma mide en la propia base los cuatro PRAGMA del
// contrato de apertura §4, que es lo que SC-006 pide: no que el DSN los
// nombre, sino que la conexión los tenga puestos.
//
// Cada uno protege algo distinto: el diario en WAL es lo que permite leer
// mientras otra invocación escribe; la confirmación sincronizada con el disco
// es la integridad que FR-031 prohíbe cambiar por velocidad; la espera ante
// bloqueo dentro del motor es el tramo de cien milisegundos que el cliente
// reintenta mirando el contexto hasta agotar los cinco segundos —la espera
// entera la miden las pruebas de espera_test.go—; y en solo lectura query_only
// refuerza dentro de la conexión lo que el modo ya promete (FR-003, FR-030,
// FR-031).
//
// Los PRAGMA se consultan por la conexión del cliente y no abriendo otra: el
// contrato prohíbe exponerla (FR-005), y una conexión nueva no sería la que New
// configuró, que es justo lo que aquí se mide.
func TestAbrirAplicaLosPragma(t *testing.T) {
	t.Parallel()

	const tramoEnMilisegundos = 100

	require.Equal(t, int64(tramoEnMilisegundos), tramoDeEspera.Milliseconds(),
		"el tramo que el motor espera en cada intento es el del contrato de apertura §4")

	directorio := t.TempDir()

	normal, err := New(t.Context(), ConDirectorio(directorio))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, normal.Close()) })

	assert.Equal(t, "wal", pragmaTexto(t, normal, "PRAGMA journal_mode"),
		"el diario va en WAL: un lector no bloquea al escritor ni al revés")
	assert.Equal(t, int64(2), pragmaEntero(t, normal, "PRAGMA synchronous"),
		"cada confirmación se sincroniza con el disco (FULL)")
	assert.Equal(t, int64(tramoEnMilisegundos), pragmaEntero(t, normal, "PRAGMA busy_timeout"),
		"ante un bloqueo el motor espera un tramo, no falla de inmediato; el resto de la espera la pone el cliente")
	assert.Equal(t, int64(0), pragmaEntero(t, normal, "PRAGMA query_only"),
		"el modo normal escribe: query_only es de la otra apertura")

	require.NoError(t, normal.Close())

	soloLectura, err := New(t.Context(), ConDirectorio(directorio), SoloLectura())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, soloLectura.Close()) })

	assert.Equal(t, int64(1), pragmaEntero(t, soloLectura, "PRAGMA query_only"),
		"en solo lectura toda escritura falla en la conexión aunque el modo no bastara")
	assert.Equal(t, int64(tramoEnMilisegundos), pragmaEntero(t, soloLectura, "PRAGMA busy_timeout"),
		"el lector también espera por tramos: no falla porque otra invocación esté escribiendo")
	assert.Equal(t, "wal", pragmaTexto(t, soloLectura, "PRAGMA journal_mode"),
		"el diario es del fichero, y el lector lo ve tal como está")
}

// TestRutaParaURI fija cómo la ruta del fichero entra en el URI «file:» de los
// dos DSN: SQLite decodifica en el camino toda secuencia %HH y lo corta en el
// primer «?» o «#», así que esos tres caracteres se escapan y ningún otro se
// toca, tampoco las barras. Es lo que impide que la base acabe en un fichero
// distinto del que declara el directorio (FR-020).
func TestRutaParaURI(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		ruta     string
		esperado string
	}{
		{nombre: "sin nada que escapar", ruta: "/ruta/kitlegal/cache.db", esperado: "/ruta/kitlegal/cache.db"},
		{nombre: "porcentaje", ruta: "/ruta/con%41pct/cache.db", esperado: "/ruta/con%2541pct/cache.db"},
		{nombre: "interrogante", ruta: "/ruta/con?duda/cache.db", esperado: "/ruta/con%3Fduda/cache.db"},
		{nombre: "almohadilla", ruta: "/ruta/con#fragmento/cache.db", esperado: "/ruta/con%23fragmento/cache.db"},
		{nombre: "espacio y ampersand van tal cual", ruta: "/ruta/con espacio&mas/cache.db", esperado: "/ruta/con espacio&mas/cache.db"},
		{nombre: "el porcentaje ya escapado se vuelve a escapar", ruta: "/ruta/%3F/cache.db", esperado: "/ruta/%253F/cache.db"},
		{nombre: "la ruta se sanea antes", ruta: "/ruta//kitlegal/./cache.db", esperado: "/ruta/kitlegal/cache.db"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.esperado, rutaParaURI(caso.ruta))
		})
	}

	assert.True(t, strings.HasPrefix(dsnNormal("/ruta/con?duda/cache.db"), "file:/ruta/con%3Fduda/cache.db?"),
		"el DSN normal lleva la ruta escapada y la primera interrogación es la de los parámetros")
	assert.True(t, strings.HasPrefix(dsnSoloLectura("/ruta/con#frag/cache.db", true), "file:/ruta/con%23frag/cache.db?mode=ro&immutable=1&"),
		"el DSN de solo lectura, en sus dos formas, lleva la ruta escapada")
}

// TestDirectorioConCaracteresDeURI fija FR-020 y FR-021 sobre el disco con los
// nombres de directorio que la sintaxis de URI de SQLite interpretaría: la base aparece
// dentro del directorio declarado y en ningún otro sitio, con sus permisos, y un
// lector de solo lectura sobre ese mismo directorio la encuentra.
//
// Sin escapar la ruta, «?» y «#» cortan el camino y la base se escribe en un
// hermano del directorio declarado mientras cache.db queda a cero bytes, y
// «%41» se decodifica como «A» y abre otro fichero. Se mide sobre el padre
// entero: nada más que el directorio declarado y lo que SQLite deja dentro.
func TestDirectorioConCaracteresDeURI(t *testing.T) {
	t.Parallel()

	nombres := []string{
		"con%41pct",
		"con?interrogante",
		"con#almohadilla",
		"con espacio",
		"con&ampersand",
	}

	for _, nombre := range nombres {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			padre := t.TempDir()
			directorio := filepath.Join(padre, nombre)
			ruta := filepath.Join(directorio, ficheroDeLaBase)
			contenido := []byte("<norma>contenido</norma>")

			escritor := clienteAbierto(t, directorio)
			require.NoError(t, escritor.Put(t.Context(), claveDePrueba, contenido, vigenciaDePrueba))
			require.NoError(t, escritor.Close())

			estado, err := os.Stat(ruta)
			require.NoError(t, err, "la base está dentro del directorio declarado")
			assert.Positive(t, estado.Size(), "y es ahí donde se escribió, no en otro fichero")
			assert.Equal(t, fs.FileMode(0o600), estado.Mode().Perm(), "con los permisos de FR-021")

			for _, entrada := range contenidoDe(t, padre) {
				assert.Equal(t, nombre, entrada, "bajo el padre solo está el directorio declarado")
			}

			lector := clienteAbierto(t, directorio, SoloLectura())

			leido, presente, err := lector.Get(t.Context(), claveDePrueba)
			require.NoError(t, err)
			assert.True(t, presente, "el lector de solo lectura abre la misma base")
			assert.Equal(t, contenido, leido)
		})
	}
}

// TestFicheroInutilizable fija FR-028 y la fila 10 del contrato de errores: en
// la ruta de la base hay un fichero que no es una base de datos que se pueda
// leer, y la caché lo dice y lo deja intacto.
//
// No borrarlo ni rehacerlo es la parte que importa y la que se mide con el
// SHA-256: lo que hay ahí puede ser de otro programa o de otra persona, y
// perderlo sería peor que no poder usar la caché. Vale igual en los dos modos,
// porque el fichero no se toca ni cuando la invocación podría escribir.
func TestFicheroInutilizable(t *testing.T) {
	t.Parallel()

	for _, modo := range modosDeApertura {
		t.Run(modo.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ruta := filepath.Join(directorio, ficheroDeLaBase)
			require.NoError(t, os.WriteFile(ruta,
				[]byte("esto no es una base de datos de SQLite y nadie va a borrarlo"), 0o600))

			antes := huella(t, ruta)

			cliente, err := New(t.Context(),
				append(slices.Clone(modo.opciones), ConDirectorio(directorio))...)

			require.Error(t, err)
			assert.Nil(t, cliente)
			assert.Equal(t, schema.ClaseInesperado, cli.Clasificar(err))
			assert.Equal(t, 1, cli.CodigoSalida(err))
			assert.Contains(t, err.Error(), ruta, "el mensaje nombra el fichero implicado")
			assert.Contains(t, err.Error(), "no se borra",
				"el mensaje dice que el fichero queda donde está (FR-028)")

			require.FileExists(t, ruta)
			assert.Equal(t, antes, huella(t, ruta),
				"un fichero que no sirve como base no se borra ni se rehace")
		})
	}
}

// TestSoloLecturaSinBaseNoCreaNada fija la mitad de FR-015 que solo se puede
// comprobar leyendo: en solo lectura, un directorio de caché sin base —vacío,
// inexistente, o cuyo padre es un fichero y que por tanto tampoco existe como
// directorio— no es un fallo de ruta sino un cliente sin base; toda lectura es
// «fuente no disponible» (4), nunca el 2 de una ruta inservible; y no se crea
// nada: ni el directorio, ni cache.db, ni sus auxiliares. Escribir tampoco crea
// nada, y es «inesperado» (1) como en cualquier cliente de solo lectura
// (FR-017, SC-011).
//
// «No se crea nada» se mide sobre el padre del directorio de la caché, entero y
// con la huella de cada fichero, de modo que el fichero que hace de padre en la
// tercera fila siga siendo el mismo byte a byte (SC-003).
func TestSoloLecturaSinBaseNoCreaNada(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		// directorioBajo prepara lo que haga falta bajo el padre y devuelve el
		// directorio de la caché.
		directorioBajo func(t *testing.T, padre string) string
	}{
		{
			nombre: "directorio vacío",
			directorioBajo: func(t *testing.T, padre string) string {
				t.Helper()

				directorio := filepath.Join(padre, "kitlegal")
				require.NoError(t, os.Mkdir(directorio, 0o700))

				return directorio
			},
		},
		{
			nombre: "directorio inexistente",
			directorioBajo: func(t *testing.T, padre string) string {
				t.Helper()

				return filepath.Join(padre, "kitlegal")
			},
		},
		{
			nombre: "directorio cuyo padre es un fichero",
			directorioBajo: func(t *testing.T, padre string) string {
				t.Helper()

				fichero := filepath.Join(padre, "fichero")
				require.NoError(t, os.WriteFile(fichero, []byte("no soy un directorio"), 0o600))

				return filepath.Join(fichero, "kitlegal")
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			padre := t.TempDir()
			directorio := caso.directorioBajo(t, padre)
			antes := arbolDe(t, padre)

			cliente := clienteAbierto(t, directorio, SoloLectura())
			assert.Nil(t, cliente.db, "sin cache.db el cliente de solo lectura no abre ninguna base")

			contenido, presente, err := cliente.Get(t.Context(), claveDePrueba)
			compruebaAusenciaEnSoloLectura(t, contenido, presente, err, claveDePrueba)

			err = cliente.Put(t.Context(), claveDePrueba, []byte("<norma>contenido</norma>"), vigenciaDePrueba)
			require.Error(t, err)
			assert.Equal(t, schema.ClaseInesperado, cli.Clasificar(err))
			assert.Equal(t, 1, cli.CodigoSalida(err))

			require.NoError(t, cliente.Close())
			assert.Equal(t, antes, arbolDe(t, padre),
				"en solo lectura no se crea nada: ni el directorio, ni la base, ni sus auxiliares")
		})
	}
}

// TestSoloLecturaBaseVacia completa lo que TestSoloLecturaNoMigra deja sin leer:
// un cache.db de cero bytes —el que deja una invocación normal interrumpida antes
// de migrar— se abre en solo lectura sin migrarlo, y leer en él es la ausencia de
// este modo, «fuente no disponible» (4), y no un fallo del fichero (1): una base
// sin esquema no tiene ninguna entrada que servir, y no por eso está estropeada
// (FR-015, FR-016). El fichero sigue con cero bytes y con la misma huella.
func TestSoloLecturaBaseVacia(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	ruta := filepath.Join(directorio, ficheroDeLaBase)
	require.NoError(t, os.WriteFile(ruta, nil, 0o600))
	antes := huella(t, ruta)

	cliente := clienteAbierto(t, directorio, SoloLectura())
	require.Equal(t, int64(0), cliente.versionEsquema, "la base no tiene esquema y no se le aplica ninguno")

	contenido, presente, err := cliente.Get(t.Context(), claveDePrueba)
	compruebaAusenciaEnSoloLectura(t, contenido, presente, err, claveDePrueba)

	require.NoError(t, cliente.Close())

	estado, err := os.Stat(ruta)
	require.NoError(t, err)
	assert.Equal(t, int64(0), estado.Size(), "leer no migra: la base sigue vacía")
	assert.Equal(t, antes, huella(t, ruta))
	assert.Equal(t, int64(0), consultaEntero(t, ruta, tablasDeLaVersion))
}

// TestSoloLecturaNoModificaElFichero fija SC-003 sobre el adaptador: una
// invocación de solo lectura deja cache.db idéntico byte a byte, tanto si lo que
// pide está como si no. La huella se toma antes, tras una lectura presente, tras
// una ausente y tras cerrar el lector.
//
// Solo cuenta cache.db: los auxiliares del registro de escritura puede crearlos
// o dejarlos SQLite al leer en un directorio que lo admite, y SC-003 los excluye.
// Lo que sí se exige es que junto a la base no aparezca ningún otro fichero.
func TestSoloLecturaNoModificaElFichero(t *testing.T) {
	t.Parallel()

	const ausente = claveDePrueba + "/ausente"

	directorio := t.TempDir()
	ruta := filepath.Join(directorio, ficheroDeLaBase)
	contenido := []byte("<norma>contenido</norma>")
	siembra(t, directorio, claveDePrueba, contenido)

	antes := huella(t, ruta)
	lector := clienteAbierto(t, directorio, SoloLectura())

	leido, presente, err := lector.Get(t.Context(), claveDePrueba)
	require.NoError(t, err)
	require.True(t, presente)
	assert.Equal(t, contenido, leido)
	assert.Equal(t, antes, huella(t, ruta), "una lectura presente no cambia ni un byte")

	leido, presente, err = lector.Get(t.Context(), ausente)
	compruebaAusenciaEnSoloLectura(t, leido, presente, err, ausente)
	assert.Equal(t, antes, huella(t, ruta), "una lectura ausente tampoco")

	require.NoError(t, lector.Close())
	assert.Equal(t, antes, huella(t, ruta), "ni cerrar el lector")

	admitidos := []string{
		ficheroDeLaBase,
		ficheroDeLaBase + sufijoRegistroDeEscritura,
		ficheroDeLaBase + sufijoMemoriaCompartida,
	}
	for _, nombre := range contenidoDe(t, directorio) {
		assert.Contains(t, admitidos, nombre, "junto a la base solo pueden estar sus auxiliares")
	}
}

// TestSoloLecturaVeLoConfirmadoEnElWAL fija el escenario 6 de US3 y la
// clarificación Q2: un lector de solo lectura ve toda entrada que otra
// invocación ya confirmó aunque siga en el registro de escritura sin
// consolidarse en cache.db, y no ve la que esa invocación tiene todavía a medias
// (FR-015, FR-032).
//
// El escritor no se cierra hasta el final, que es lo que impide el punto de
// control, y que las entradas vivan solo en el registro se comprueba en vez de
// suponerse: cache.db tiene la misma huella antes y después de confirmarlas, y
// el registro no está vacío. Un lector que abriera la base como inmutable leería
// solo cache.db y declararía ausentes —código 4— todas las confirmadas.
//
// Lo que está a medias es una transacción abierta en la conexión del escritor,
// con su fila insertada y sin confirmar mientras el lector lee: el lector no la
// ve y tampoco se queda esperándola.
func TestSoloLecturaVeLoConfirmadoEnElWAL(t *testing.T) {
	t.Parallel()

	const (
		confirmadas = 8
		pendiente   = claveDePrueba + "/pendiente"
	)

	claveDe := func(indice int) string { return claveDePrueba + "/" + strconv.Itoa(indice) }
	cuerpoDe := func(indice int) []byte { return []byte("<norma>entrada " + strconv.Itoa(indice) + "</norma>") }

	directorio := t.TempDir()
	ruta := filepath.Join(directorio, ficheroDeLaBase)
	escritor := clienteAbierto(t, directorio)

	antes := huella(t, ruta)

	for indice := range confirmadas {
		require.NoError(t, escritor.Put(t.Context(), claveDe(indice), cuerpoDe(indice), vigenciaDePrueba))
	}

	require.Equal(t, antes, huella(t, ruta),
		"lo confirmado sigue en el registro de escritura: no ha habido punto de control")

	registro, err := os.Stat(ruta + sufijoRegistroDeEscritura)
	require.NoError(t, err)
	require.Positive(t, registro.Size(), "lo confirmado está en el registro de escritura")

	tx, err := escritor.db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := tx.Rollback(); !errors.Is(err, sql.ErrTxDone) {
			assert.NoError(t, err)
		}
	})

	_, err = tx.ExecContext(t.Context(),
		`INSERT INTO entradas(clave, contenido, expira_en) VALUES (?, ?, ?)`,
		pendiente, []byte("<norma>a medias</norma>"), time.Now().Add(vigenciaDePrueba).UnixNano())
	require.NoError(t, err)

	lector := clienteAbierto(t, directorio, SoloLectura())

	for indice := range confirmadas {
		leido, presente, err := lector.Get(t.Context(), claveDe(indice))
		require.NoError(t, err, "una entrada confirmada no es una ausencia por seguir en el registro")
		assert.True(t, presente)
		assert.Equal(t, cuerpoDe(indice), leido, "entera, nunca a medias")
	}

	leido, presente, err := lector.Get(t.Context(), pendiente)
	compruebaAusenciaEnSoloLectura(t, leido, presente, err, pendiente)

	require.NoError(t, tx.Rollback())
}

// pragmaTexto y pragmaEntero consultan un PRAGMA por la conexión del cliente.
// La consulta llega entera desde quien llama —y es siempre una constante del
// test— porque componerla aquí sería construir SQL con una cadena de fuera.
func pragmaTexto(t *testing.T, cliente *Cliente, consulta string) string {
	t.Helper()

	var valor string
	require.NoError(t, cliente.db.QueryRowContext(t.Context(), consulta).Scan(&valor))

	return valor
}

func pragmaEntero(t *testing.T, cliente *Cliente, consulta string) int64 {
	t.Helper()

	var valor int64
	require.NoError(t, cliente.db.QueryRowContext(t.Context(), consulta).Scan(&valor))

	return valor
}

// huella es el SHA-256 del fichero en hexadecimal, y es la forma de comprobar
// que algo no cambió ni un byte: que un fichero que no se puede usar no se
// borra ni se rehace (FR-028) y que una invocación de solo lectura deja la base
// idéntica (SC-003).
func huella(t *testing.T, ruta string) string {
	t.Helper()

	suma := sha256.Sum256(leeLaBase(t, ruta))

	return hex.EncodeToString(suma[:])
}

// leeLaBase devuelve cache.db entero, tal como está en el disco: es lo que
// comparan las huellas y lo que las pruebas que estropean páginas toman como
// punto de partida.
func leeLaBase(t *testing.T, ruta string) []byte {
	t.Helper()

	contenido, err := os.ReadFile(filepath.Clean(ruta))
	require.NoError(t, err)

	return contenido
}

// arbolDe describe todo lo que hay bajo una raíz —cada directorio por su ruta
// relativa y cada fichero por su ruta y su huella—, de modo que dos
// descripciones iguales signifiquen que no se creó, no se borró y no cambió
// nada (FR-015, SC-003).
func arbolDe(t *testing.T, raiz string) []string {
	t.Helper()

	var descripcion []string

	err := filepath.WalkDir(raiz, func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relativa, err := filepath.Rel(raiz, ruta)
		if err != nil {
			return err
		}

		if entrada.IsDir() {
			descripcion = append(descripcion, relativa+string(filepath.Separator))

			return nil
		}

		descripcion = append(descripcion, relativa+" "+huella(t, ruta))

		return nil
	})
	require.NoError(t, err)

	return descripcion
}
