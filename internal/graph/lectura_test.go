package graph

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

const (
	// diarioWAL y diarioClasico son los dos modos de diario con los que las
	// pruebas escriben world.db: el de las entregas y el de una base de fuera.
	diarioWAL     = "wal"
	diarioClasico = "delete"

	fuenteDelBOE        = "boe.legislacion-consolidada"
	fuenteDelTerritorio = "kitlegal.territorio"
	// fuenteAjena empieza por mayúscula para que el orden por bytes la ponga
	// delante de las que empiezan por minúscula, al revés que un orden por
	// idioma.
	fuenteAjena = "Zeta.fuente"

	urlDeLaNorma      = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565"
	urlDelBloque      = urlDeLaNorma + "/texto/bloque/a21"
	urlDelTerritorio  = "https://www.ine.es/daco/daco42/codmun/diccionario26.xlsx"
	urlAjena          = "https://ajena.example/organo"
	relacionMayuscula = "Z:enlace"

	idNorma            = "eli/es/l/2015/10/01/39"
	idBloque           = idNorma + "#a21"
	idMunicipio        = "ine:28074"
	idOrgano           = "L01280745"
	idOrganoAjeno      = "L01280740"
	idOrganoMinusculas = "l00000000"

	// Las fechas se escriben de formas distintas a propósito: la lectura las
	// devuelve carácter a carácter, sin llevarlas a otro desplazamiento ni
	// quitarles o añadirles fracción de segundo (FR-053).
	fechaConDesplazamiento = "2026-09-28T12:00:00+02:00"
	fechaConFraccion       = "2026-09-29T10:00:00.5Z"
	fechaDelBloque         = "2026-09-28T10:00:00Z"
	fechaDelTerritorio     = "2026-02-04T00:00:00Z"
	fechaAjena             = "2026-03-01T00:00:00-05:00"

	cuerpo2016 = "CUERPO-PRIVADO-2016 del bloque a21, que ninguna lectura devuelve"
	cuerpo2025 = "CUERPO-PRIVADO-2025 del bloque a21, que ninguna lectura devuelve"

	semana = 7 * 24 * time.Hour
)

// grafoDePrueba son las filas que una prueba escribe en world.db, en el orden
// en que las escribe.
type grafoDePrueba struct {
	nodos   []grafo.RegistroDeNodo
	aristas []grafo.RegistroDeArista
	textos  []grafo.RegistroDeTexto
}

// muestraDelGrafo es el grafo de las pruebas de lectura, cada fila con su
// nombre para poder escribir lo esperado a partir de ella.
type muestraDelGrafo struct {
	norma, bloque, v2016, v2025, municipio, organo, organoAjeno, organoMinusculas grafo.RegistroDeNodo

	parte, version2025, version2016, enlace, perteneceMinusculas, pertenece, perteneceAjeno grafo.RegistroDeArista

	texto2016, texto2025 grafo.RegistroDeTexto
}

// laMuestra construye la muestra: una norma, su bloque y dos versiones del
// bloque, un municipio y tres órganos que pertenecen a él —uno cuya última
// observación es de otra fuente—, y una arista con una relación en mayúscula,
// que el orden por bytes pone delante de las demás.
func laMuestra(t *testing.T) muestraDelGrafo {
	t.Helper()

	boe := func(url, fecha string) grafo.Procedencia {
		return grafo.Procedencia{Fuente: fuenteDelBOE, URL: url, FechaConsulta: fecha}
	}
	territorio := grafo.Procedencia{Fuente: fuenteDelTerritorio, URL: urlDelTerritorio, FechaConsulta: fechaDelTerritorio}
	ajena := grafo.Procedencia{Fuente: fuenteAjena, URL: urlAjena, FechaConsulta: fechaAjena}

	nodo := func(id, tipo string, datos map[string]any, primera, ultima grafo.Procedencia, vigencia time.Duration,
	) grafo.RegistroDeNodo {
		canonicos, err := grafo.DatosCanonicos(datos)
		require.NoError(t, err)

		return grafo.RegistroDeNodo{
			ID: id, Tipo: tipo, Datos: canonicos,
			PrimeraObservacion: primera, UltimaObservacion: ultima, Vigencia: vigencia,
		}
	}

	arista := func(origen, relacion, destino string, primera, ultima grafo.Procedencia, vigencia time.Duration,
	) grafo.RegistroDeArista {
		return grafo.RegistroDeArista{
			Origen: origen, Relacion: relacion, Destino: destino,
			PrimeraObservacion: primera, UltimaObservacion: ultima, Vigencia: vigencia,
		}
	}

	texto2016 := grafo.RegistroDeTexto{Huella: huella(cuerpo2016), Cuerpo: cuerpo2016, Procedencia: boe(urlDelBloque, fechaDelBloque)}
	texto2025 := grafo.RegistroDeTexto{Huella: huella(cuerpo2025), Cuerpo: cuerpo2025, Procedencia: boe(urlDelBloque, fechaConFraccion)}
	id2016 := idBloque + "@20161002:" + texto2016.Huella
	id2025 := idBloque + "@20250101:" + texto2025.Huella

	return muestraDelGrafo{
		norma: nodo(idNorma, grafo.TipoNorma, map[string]any{grafo.DatoIdentificador: "BOE-A-2015-10565"},
			boe(urlDeLaNorma, fechaConDesplazamiento), boe(urlDeLaNorma, fechaConFraccion), semana),
		bloque: nodo(idBloque, grafo.TipoBloque, map[string]any{grafo.DatoBloque: "a21"},
			boe(urlDelBloque, fechaDelBloque), boe(urlDelBloque, fechaDelBloque), semana),
		v2016: nodo(id2016, grafo.TipoBloqueVersion,
			map[string]any{grafo.DatoFechaVigencia: "20161002", grafo.DatoHashTexto: texto2016.Huella},
			boe(urlDelBloque, fechaDelBloque), boe(urlDelBloque, fechaDelBloque), semana),
		v2025: nodo(id2025, grafo.TipoBloqueVersion,
			map[string]any{grafo.DatoFechaVigencia: "20250101", grafo.DatoHashTexto: texto2025.Huella},
			boe(urlDelBloque, fechaConFraccion), boe(urlDelBloque, fechaConFraccion), semana),
		municipio: nodo(idMunicipio, grafo.TipoMunicipio,
			map[string]any{grafo.DatoCodigoINE: "28074", grafo.DatoNombre: "Legan\xc3\xa9s"}, territorio, territorio, 0),
		organo: nodo(idOrgano, grafo.TipoOrgano, map[string]any{grafo.DatoDIR3: idOrgano}, territorio, territorio, 0),
		organoAjeno: nodo(idOrganoAjeno, grafo.TipoOrgano, map[string]any{grafo.DatoDIR3: idOrganoAjeno},
			territorio, ajena, 0),
		organoMinusculas: nodo(idOrganoMinusculas, grafo.TipoOrgano,
			map[string]any{grafo.DatoDIR3: idOrganoMinusculas}, territorio, territorio, 0),

		parte: arista(idNorma, grafo.RelacionTieneParte, idBloque,
			boe(urlDeLaNorma, fechaConDesplazamiento), boe(urlDeLaNorma, fechaConFraccion), semana),
		version2025: arista(idBloque, grafo.RelacionTieneVersion, id2025,
			boe(urlDelBloque, fechaConFraccion), boe(urlDelBloque, fechaConFraccion), semana),
		version2016: arista(idBloque, grafo.RelacionTieneVersion, id2016,
			boe(urlDelBloque, fechaDelBloque), boe(urlDelBloque, fechaDelBloque), semana),
		enlace: arista(idBloque, relacionMayuscula, idMunicipio,
			boe(urlDelBloque, fechaDelBloque), boe(urlDelBloque, fechaDelBloque), 0),
		perteneceMinusculas: arista(idOrganoMinusculas, grafo.RelacionPerteneceA, idMunicipio, territorio, territorio, 0),
		pertenece:           arista(idOrgano, grafo.RelacionPerteneceA, idMunicipio, territorio, territorio, 0),
		perteneceAjeno:      arista(idOrganoAjeno, grafo.RelacionPerteneceA, idMunicipio, territorio, ajena, 0),

		texto2016: texto2016,
		texto2025: texto2025,
	}
}

// grafo son las filas de la muestra. Las aristas que llegan al municipio se
// escriben en el orden inverso al de los bytes de su origen, que es el orden en
// que las devuelve el índice por destino: si la lectura no las ordenara, se
// vería.
func (m muestraDelGrafo) grafo() grafoDePrueba {
	return grafoDePrueba{
		nodos: []grafo.RegistroDeNodo{
			m.organoMinusculas, m.municipio, m.v2025, m.v2016, m.bloque, m.norma, m.organo, m.organoAjeno,
		},
		aristas: []grafo.RegistroDeArista{
			m.parte, m.version2025, m.version2016, m.enlace, m.perteneceMinusculas, m.pertenece, m.perteneceAjeno,
		},
		textos: []grafo.RegistroDeTexto{m.texto2025, m.texto2016},
	}
}

// muestra son las filas de la muestra.
func muestra(t *testing.T) grafoDePrueba {
	t.Helper()

	return laMuestra(t).grafo()
}

// recuentoDeLaMuestra es lo que graph stats cuenta en la muestra: por la fuente
// de la última observación, en el orden de los bytes y sin ningún par a cero.
func recuentoDeLaMuestra() grafo.Recuento {
	return grafo.Recuento{
		Nodos:   8,
		Aristas: 7,
		Textos:  2,
		NodosPorTipo: []grafo.RecuentoDeNodos{
			{Tipo: grafo.TipoBloque, Fuente: fuenteDelBOE, Nodos: 1},
			{Tipo: grafo.TipoBloqueVersion, Fuente: fuenteDelBOE, Nodos: 2},
			{Tipo: grafo.TipoMunicipio, Fuente: fuenteDelTerritorio, Nodos: 1},
			{Tipo: grafo.TipoNorma, Fuente: fuenteDelBOE, Nodos: 1},
			{Tipo: grafo.TipoOrgano, Fuente: fuenteAjena, Nodos: 1},
			{Tipo: grafo.TipoOrgano, Fuente: fuenteDelTerritorio, Nodos: 2},
		},
		AristasPorRelacion: []grafo.RecuentoDeAristas{
			{Relacion: relacionMayuscula, Fuente: fuenteDelBOE, Aristas: 1},
			{Relacion: grafo.RelacionTieneParte, Fuente: fuenteDelBOE, Aristas: 1},
			{Relacion: grafo.RelacionTieneVersion, Fuente: fuenteDelBOE, Aristas: 2},
			{Relacion: grafo.RelacionPerteneceA, Fuente: fuenteAjena, Aristas: 1},
			{Relacion: grafo.RelacionPerteneceA, Fuente: fuenteDelTerritorio, Aristas: 2},
		},
	}
}

// recuentoVacio es lo que graph stats cuenta en un grafo ausente o vacío: tres
// ceros y dos listas vacías, nunca nulas (FR-054).
func recuentoVacio() grafo.Recuento {
	return grafo.Recuento{NodosPorTipo: []grafo.RecuentoDeNodos{}, AristasPorRelacion: []grafo.RecuentoDeAristas{}}
}

// huella es la del cuerpo de un texto: «sha256:» y los 64 hexadecimales.
func huella(cuerpo string) string {
	suma := sha256.Sum256([]byte(cuerpo))

	return "sha256:" + hex.EncodeToString(suma[:])
}

// casoDeLectura es una fila de TestLeerEstados: cómo se prepara el directorio
// de la caché y lo que se lee, o el mensaje del fallo, que en esta tabla es
// siempre «inesperado».
type casoDeLectura struct {
	nombre string
	// preparar deja el estado bajo la raíz y devuelve el directorio de la caché.
	preparar func(t *testing.T, raiz string) string
	// llena dice que se lee la muestra; sin ella y sin fallo, el grafo vacío.
	llena bool
	// fallo es el mensaje esperado para la ruta de world.db; nil si no falla.
	fallo func(ruta string) string
}

// TestLeerEstados fija la lectura en cada estado de world.db de data-model §3.1
// que no necesita otra conexión (contracts/almacen-world-db.md §3 y §6; FR-004,
// FR-005, FR-010, FR-012): ausente, de 0 bytes o sin esquema, un grafo vacío;
// en la versión 1, la muestra, con permiso de escritura o sin él, en WAL o en
// modo rollback y a través de un enlace; un directorio, lo que no es una base,
// una base dañada y un esquema posterior, «inesperado» con su mensaje. En todos,
// ningún fichero bajo la raíz aparece, cambia o desaparece.
func TestLeerEstados(t *testing.T) {
	t.Parallel()

	noUtilizable := func(ruta string) string {
		return "grafo: " + strconv.Quote(ruta) + " no es una base de datos utilizable; no se modifica"
	}
	posterior := func(ruta string) string {
		return "grafo: " + strconv.Quote(ruta) +
			" tiene el esquema en la versi\xc3\xb3n 2 y este binario conoce la 1: no se modifica"
	}

	casos := []casoDeLectura{
		{nombre: "ausente", preparar: cacheVacia},
		{
			nombre: "sin su directorio",
			preparar: func(_ *testing.T, raiz string) string {
				return filepath.Join(raiz, "no-existe", "cache")
			},
		},
		{nombre: "bajo un componente que no es directorio", preparar: bajoUnFichero},
		{nombre: "un enlace sin destino", preparar: conEnlaceSinDestino},
		{nombre: "de 0 bytes", preparar: conBase(nil, 0o600)},
		{nombre: "de 0 bytes y sin permiso de escritura", preparar: conBase(nil, 0o400)},
		{nombre: "sin esquema, con una tabla ajena que se llama nodes", preparar: conSentencias(diarioWAL, 0o600, tablaAjena)},
		{nombre: "sin esquema y sin permiso de escritura", preparar: conSentencias(diarioWAL, 0o400, tablaAjena)},
		{nombre: "versión 1 en WAL", preparar: conMuestra(diarioWAL, 0o600), llena: true},
		{nombre: "versión 1 en modo rollback", preparar: conMuestra(diarioClasico, 0o600), llena: true},
		{nombre: "versión 1 en WAL y sin permiso de escritura", preparar: conMuestra(diarioWAL, 0o400), llena: true},
		{
			nombre:   "versión 1 en modo rollback y sin permiso de escritura",
			preparar: conMuestra(diarioClasico, 0o400),
			llena:    true,
		},
		{nombre: "versión 1 a través de un enlace", preparar: conMuestraEnlazada, llena: true},
		{
			nombre:   "un directorio",
			preparar: conDirectorioEnSuLugar,
			fallo: func(ruta string) string {
				return "grafo: " + strconv.Quote(ruta) + " es un directorio y no una base de datos utilizable; no se modifica"
			},
		},
		{nombre: "lo que no es una base de datos", preparar: conBase(contenido, 0o600), fallo: noUtilizable},
		{nombre: "lo que no es una base, sin permiso de escritura", preparar: conBase(contenido, 0o400), fallo: noUtilizable},
		{nombre: "una base dañada", preparar: conBaseDanada, fallo: noUtilizable},
		{nombre: "un esquema posterior", preparar: conSentencias(diarioWAL, 0o600, versionPosterior), fallo: posterior},
		{
			nombre:   "un esquema posterior, sin permiso de escritura",
			preparar: conSentencias(diarioClasico, 0o400, versionPosterior),
			fallo:    posterior,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			compruebaEstado(t, caso)
		})
	}

	t.Run("la ruta no resoluble", func(t *testing.T) {
		t.Parallel()

		lectura, err := Leer(t.Context(), ConDirectorio(""))
		assert.Nil(t, lectura)

		var fallo *Error

		require.ErrorAs(t, err, &fallo)
		assert.Equal(t, schema.ClaseArgumentos, fallo.Clase())
		assert.Equal(t, operacionLeer, fallo.Operacion)
		assert.True(t, strings.HasPrefix(err.Error(), "grafo: no se puede ubicar world.db: "), err.Error())
	})
}

// TestLeerSinPermisoDeLectura fija world.db que el proceso no puede leer
// (FR-010; contracts/almacen-world-db.md §6, fila 2): «inesperado» con la
// variante del acceso denegado, que es lo que quien lo lee puede arreglar, y
// sin crear, cambiar ni retirar nada. Como el fichero no se deja leer, lo que se
// compara es la lista del directorio y el tamaño, los permisos y la fecha de
// cada entrada.
func TestLeerSinPermisoDeLectura(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	ruta := crearGrafo(t, directorio, diarioWAL, muestra(t))
	restringir(t, ruta, permisosDenegados)

	_, err := os.Open(filepath.Clean(ruta))
	require.ErrorIs(t, err, fs.ErrPermission, "premisa: el fichero no se deja leer (¿el proceso corre como root?)")

	antes := entradasDe(t, directorio)

	lectura, err := Leer(t.Context(), ConDirectorio(directorio))
	assert.Nil(t, lectura)
	compruebaFallo(t, err, schema.ClaseInesperado, ruta,
		"grafo: "+strconv.Quote(ruta)+" no es una base de datos utilizable: acceso denegado; no se modifica")
	assert.Equal(t, antes, entradasDe(t, directorio))
}

// permisosDenegados son los de un fichero que nadie salvo root puede leer ni
// escribir.
const permisosDenegados fs.FileMode = 0o000

// entradasDe describe cada entrada del directorio por su nombre, su tamaño, sus
// permisos y su fecha de modificación, sin leer su contenido.
func entradasDe(t *testing.T, directorio string) map[string]string {
	t.Helper()

	entradas, err := os.ReadDir(directorio)
	require.NoError(t, err)

	descritas := map[string]string{}

	for _, entrada := range entradas {
		info, err := entrada.Info()
		require.NoError(t, err)

		descritas[entrada.Name()] = fmt.Sprintf("%d %s %s", info.Size(), info.Mode(), info.ModTime().Format(time.RFC3339Nano))
	}

	return descritas
}

// compruebaEstado es el cuerpo de cada fila de TestLeerEstados.
func compruebaEstado(t *testing.T, caso casoDeLectura) {
	t.Helper()

	raiz := t.TempDir()
	directorio := caso.preparar(t, raiz)
	ruta := filepath.Join(directorio, "world.db")
	antes := huellasDelArbol(t, raiz)

	lectura, err := Leer(t.Context(), ConDirectorio(directorio))

	if caso.fallo != nil {
		assert.Nil(t, lectura)
		compruebaFallo(t, err, schema.ClaseInesperado, ruta, caso.fallo(ruta))
	} else {
		require.NoError(t, err)
		compruebaContenido(t, lectura, caso.llena)
		require.NoError(t, lectura.Close())
	}

	assert.Equal(t, antes, huellasDelArbol(t, raiz), "leer no crea, no cambia ni retira nada")
}

// compruebaContenido lee con los tres verbos y compara con la muestra o con el
// grafo vacío.
func compruebaContenido(t *testing.T, lectura *Lectura, llena bool) {
	t.Helper()

	recuento, err := lectura.Recuento(t.Context())
	require.NoError(t, err)

	_, encontrado, err := lectura.Ficha(t.Context(), idNorma)
	require.NoError(t, err)

	instantanea, err := lectura.Instantanea(t.Context())
	require.NoError(t, err)

	if llena {
		assert.Equal(t, recuentoDeLaMuestra(), recuento)
		assert.True(t, encontrado, "la norma de la muestra está")
		assert.Len(t, instantanea.Nodos, 8)
		assert.Len(t, instantanea.Aristas, 7)

		return
	}

	assert.Equal(t, recuentoVacio(), recuento)
	assert.False(t, encontrado, "en el grafo vacío no hay ningún nodo")
	assert.Equal(t, grafo.Instantanea{Nodos: []grafo.NodoDeInstantanea{}, Aristas: []schema.Arista{}}, instantanea)
}

// compruebaFallo compara el fallo de una lectura con su clase, su ruta y su
// mensaje exacto.
func compruebaFallo(t *testing.T, err error, clase schema.Clase, ruta, mensaje string) {
	t.Helper()

	var fallo *Error

	require.ErrorAs(t, err, &fallo)
	assert.Equal(t, clase, fallo.Clase())
	assert.Equal(t, operacionLeer, fallo.Operacion)
	assert.Equal(t, ruta, fallo.Ruta)
	assert.Equal(t, mensaje, err.Error())
}

// TestLeerSinRastro fija que una lectura completa —Leer, los tres verbos y
// Close— sin auxiliares no cambia ni un byte de nada en el directorio, pueda el
// proceso escribir world.db o no, en WAL o en modo rollback, y también si
// world.db es un enlace (contracts/almacen-world-db.md §3, paso 6; FR-004,
// FR-005; SC-004).
func TestLeerSinRastro(t *testing.T) {
	t.Parallel()

	casos := map[string]func(*testing.T, string) string{
		"en WAL, con permiso":                   conMuestra(diarioWAL, 0o600),
		"en WAL, sin permiso":                   conMuestra(diarioWAL, 0o400),
		"en modo rollback, con permiso":         conMuestra(diarioClasico, 0o600),
		"en modo rollback, sin permiso":         conMuestra(diarioClasico, 0o400),
		"a través de un enlace":                 conMuestraEnlazada,
		"a través de un enlace, sin permiso":    conMuestraEnlazadaSinPermiso,
		"sin esquema, en WAL y con permiso":     conSentencias(diarioWAL, 0o600, tablaAjena),
		"sin esquema, en rollback, sin permiso": conSentencias(diarioClasico, 0o400, tablaAjena),
	}

	for nombre, preparar := range casos {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			raiz := t.TempDir()
			directorio := preparar(t, raiz)
			antes := huellasDelArbol(t, raiz)

			lectura, err := Leer(t.Context(), ConDirectorio(directorio))
			require.NoError(t, err)

			_, err = lectura.Recuento(t.Context())
			require.NoError(t, err)

			_, _, err = lectura.Ficha(t.Context(), idBloque)
			require.NoError(t, err)

			_, err = lectura.Instantanea(t.Context())
			require.NoError(t, err)

			require.NoError(t, lectura.Close())

			assert.Equal(t, antes, huellasDelArbol(t, raiz), "ningún fichero aparece, cambia o desaparece")
		})
	}
}

// TestLeerConAuxiliares fija la lectura cuando SQLite dejó auxiliares
// (contracts/almacen-world-db.md §3, pasos 1, 4 y 5; data-model §3.1; research.md
// D10, V36 D, V43, V49). Cada caso lleva su control, que comprueba sobre otra
// copia de los mismos ficheros la premisa que lo hace necesario: sin él, la
// prueba pasaría también con una lectura que no eligiera bien.
func TestLeerConAuxiliares(t *testing.T) {
	t.Parallel()

	t.Run("un -wal huérfano: se lee lo confirmado en él y world.db y el -wal no cambian", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := filepath.Join(directorio, "world.db")
		copiaDeUnEscritorAbierto(t, directorio, muestra(t))
		antes := huellasDeFicheros(t, ruta, ruta+sufijoWAL)

		lectura, err := Leer(t.Context(), ConDirectorio(directorio))
		require.NoError(t, err)
		compruebaContenido(t, lectura, true)
		require.NoError(t, lectura.Close())

		assert.Equal(t, antes, huellasDeFicheros(t, ruta, ruta+sufijoWAL))
		assert.FileExists(t, ruta+sufijoMemoriaCompartida, "el -shm es la desviación declarada: SQLite lo reescribe")

		control := t.TempDir()
		copiaDeUnEscritorAbierto(t, control, muestra(t))
		leerLaVersionEn(t, filepath.Join(control, "world.db"), modoSinAuxiliares)
		assert.NoFileExists(t, filepath.Join(control, "world.db"+sufijoWAL),
			"control: el modo sin auxiliares hace el checkpoint al cerrar y retira el -wal")
	})

	t.Run("un diario caliente: inesperado y nada cambia", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := filepath.Join(directorio, "world.db")
		copiaDeUnDiarioCaliente(t, directorio)
		antes := huellasDelArbol(t, directorio)

		lectura, err := Leer(t.Context(), ConDirectorio(directorio))
		assert.Nil(t, lectura)
		compruebaFallo(t, err, schema.ClaseInesperado, ruta, "grafo: "+strconv.Quote(ruta)+
			" tiene una transacci\xc3\xb3n interrumpida sin deshacer; no se modifica")
		assert.Equal(t, antes, huellasDelArbol(t, directorio))

		control := t.TempDir()
		copiaDeUnDiarioCaliente(t, control)
		leerLaVersionEn(t, filepath.Join(control, "world.db"), modoSinAuxiliares)
		assert.NoFileExists(t, filepath.Join(control, "world.db"+sufijoDiario),
			"control: el modo sin auxiliares deshace la transacción interrumpida al leer")
	})

	for nombre, ajuste := range map[string]struct {
		conMemoriaCompartida bool
		permisos             os.FileMode
	}{
		"de 0 bytes con un -wal no vacío":                {permisos: 0o600},
		"de 0 bytes con un -wal no vacío y su -shm":      {conMemoriaCompartida: true, permisos: 0o600},
		"de 0 bytes, sin permiso y con un -wal no vacío": {permisos: 0o400},
	} {
		t.Run(nombre+": el grafo vacío sin abrir SQLite", func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ceroBytesConUnWAL(t, directorio, ajuste.conMemoriaCompartida, ajuste.permisos)
			antes := huellasDelArbol(t, directorio)

			lectura, err := Leer(t.Context(), ConDirectorio(directorio))
			require.NoError(t, err)
			compruebaContenido(t, lectura, false)
			require.NoError(t, lectura.Close())

			assert.Equal(t, antes, huellasDelArbol(t, directorio), "tampoco el -wal, que abrir SQLite borraría")

			control := t.TempDir()
			ceroBytesConUnWAL(t, control, ajuste.conMemoriaCompartida, 0o600)
			leerLaVersionEn(t, filepath.Join(control, "world.db"), modoConAuxiliares)
			assert.NoFileExists(t, filepath.Join(control, "world.db"+sufijoWAL),
				"control: SQLite borra el -wal de una base de 0 páginas al abrirla")
		})
	}
}

// TestLeerReabreInmutable fija el recurso de §3, paso 5, en un directorio que no
// admite los auxiliares: sin -wal ni -journal, la lectura se reabre inmutable y
// lee sin crear nada (research.md V9); con un -wal, lo confirmado en él no se
// puede leer sin su memoria compartida y el fichero es inutilizable.
func TestLeerReabreInmutable(t *testing.T) {
	t.Parallel()

	t.Run("sin -wal ni -journal: se lee", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		crearGrafo(t, directorio, diarioWAL, muestra(t))
		directorioSinEscritura(t, directorio)
		antes := huellasDelArbol(t, directorio)

		decidida, err := decidirApertura(filepath.Join(directorio, "world.db"))
		require.NoError(t, err)
		require.Equal(t, modoSinAuxiliares, decidida.modo, "premisa: el fichero se puede escribir y no hay auxiliares")

		lectura, err := Leer(t.Context(), ConDirectorio(directorio))
		require.NoError(t, err)
		compruebaContenido(t, lectura, true)
		require.NoError(t, lectura.Close())

		assert.Equal(t, antes, huellasDelArbol(t, directorio))
	})

	t.Run("con un -wal y sin su -shm: inutilizable", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := filepath.Join(directorio, "world.db")
		copiaDeUnEscritorAbierto(t, directorio, muestra(t))
		require.NoError(t, os.Remove(ruta+sufijoMemoriaCompartida))
		directorioSinEscritura(t, directorio)
		antes := huellasDelArbol(t, directorio)

		lectura, err := Leer(t.Context(), ConDirectorio(directorio))
		assert.Nil(t, lectura)
		compruebaFallo(t, err, schema.ClaseInesperado, ruta,
			"grafo: "+strconv.Quote(ruta)+" no es una base de datos utilizable; no se modifica")
		assert.Equal(t, antes, huellasDelArbol(t, directorio))
	})
}

// TestLeerConElPlazoAgotado fija que la lectura atiende el contexto (FR-014;
// contracts/almacen-world-db.md §4, «Esperas», y §6): terminado antes de leer, o
// mientras otra invocación retiene world.db —al leer la versión y dentro de
// cada verbo—, es «fuente-no-disponible» con su mensaje, y llega sin agotar la
// espera propia.
func TestLeerConElPlazoAgotado(t *testing.T) {
	t.Parallel()

	plazo := func(ruta string) string {
		return "grafo: el plazo termin\xc3\xb3 antes de leer " + strconv.Quote(ruta)
	}

	t.Run("el contexto terminado antes de leer", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := crearGrafo(t, directorio, diarioWAL, muestra(t))
		antes := huellasDelArbol(t, directorio)

		terminado, cancela := context.WithCancel(t.Context())
		cancela()

		lectura, err := Leer(terminado, ConDirectorio(directorio))
		assert.Nil(t, lectura)
		compruebaFallo(t, err, schema.ClaseFuenteNoDisponible, ruta, plazo(ruta))
		require.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, antes, huellasDelArbol(t, directorio))
	})

	t.Run("otra invocación retiene world.db al abrirla", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := crearGrafo(t, directorio, diarioClasico, muestra(t))
		retenerElBloqueoExclusivo(t, ruta)

		conPlazo, cancela := context.WithTimeout(t.Context(), 300*time.Millisecond)
		defer cancela()

		inicio := time.Now()
		lectura, err := Leer(conPlazo, ConDirectorio(directorio))
		tardo := time.Since(inicio)

		assert.Nil(t, lectura)
		compruebaFallo(t, err, schema.ClaseFuenteNoDisponible, ruta, plazo(ruta))
		assert.Less(t, tardo, esperaPropia, "el plazo llega antes que la espera propia")
	})

	t.Run("otra invocación retiene world.db durante cada verbo", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := crearGrafo(t, directorio, diarioClasico, muestra(t))

		lectura, err := Leer(t.Context(), ConDirectorio(directorio))
		require.NoError(t, err)

		t.Cleanup(func() { assert.NoError(t, lectura.Close()) })

		retenerElBloqueoExclusivo(t, ruta)

		conPlazo := func() context.Context {
			ctx, cancela := context.WithTimeout(t.Context(), 200*time.Millisecond)
			t.Cleanup(cancela)

			return ctx
		}

		_, _, err = lectura.Ficha(conPlazo(), idNorma)
		compruebaFallo(t, err, schema.ClaseFuenteNoDisponible, ruta, plazo(ruta))

		_, err = lectura.Recuento(conPlazo())
		compruebaFallo(t, err, schema.ClaseFuenteNoDisponible, ruta, plazo(ruta))

		_, err = lectura.Instantanea(conPlazo())
		compruebaFallo(t, err, schema.ClaseFuenteNoDisponible, ruta, plazo(ruta))
	})
}

// TestLeerBloqueada fija la espera propia agotada al leer (FR-014;
// contracts/almacen-world-db.md §6): con otra invocación reteniendo world.db
// más de 5 s y el contexto vivo, «inesperado» con su mensaje.
func TestLeerBloqueada(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	ruta := crearGrafo(t, directorio, diarioClasico, muestra(t))
	retenerElBloqueoExclusivo(t, ruta)

	inicio := time.Now()
	lectura, err := Leer(t.Context(), ConDirectorio(directorio))
	tardo := time.Since(inicio)

	assert.Nil(t, lectura)
	compruebaFallo(t, err, schema.ClaseInesperado, ruta, "grafo: "+strconv.Quote(ruta)+
		" est\xc3\xa1 bloqueada por otra invocaci\xc3\xb3n y la espera de 5s se agot\xc3\xb3")
	assert.GreaterOrEqual(t, tardo, esperaPropia, "se espera la espera propia entera")
}

// TestFicha fija lo que graph show lee de un nodo (FR-053, FR-070;
// contracts/applet-graph.md §3.1): sus datos guardados, su primera y su última
// observación con las fechas carácter a carácter, y sus aristas salientes y
// entrantes ordenadas por relación y después por el id del otro extremo,
// comparando bytes; nunca nulas y nunca con el cuerpo de un texto. Un id que no
// está, o el grafo vacío, no se encuentra.
func TestFicha(t *testing.T) {
	t.Parallel()

	m := laMuestra(t)
	directorio := t.TempDir()
	crearGrafo(t, directorio, diarioWAL, m.grafo())

	lectura, err := Leer(t.Context(), ConDirectorio(directorio))
	require.NoError(t, err)

	t.Cleanup(func() { assert.NoError(t, lectura.Close()) })

	esperadas := map[string]grafo.Ficha{
		idNorma: {
			Nodo:      nodoDeFicha(m.norma, map[string]any{grafo.DatoIdentificador: "BOE-A-2015-10565"}),
			Salientes: []grafo.AristaDeFicha{aristaDeFicha(m.parte, idBloque)},
			Entrantes: []grafo.AristaDeFicha{},
		},
		idBloque: {
			Nodo: nodoDeFicha(m.bloque, map[string]any{grafo.DatoBloque: "a21"}),
			Salientes: []grafo.AristaDeFicha{
				aristaDeFicha(m.enlace, idMunicipio),
				aristaDeFicha(m.version2016, m.v2016.ID),
				aristaDeFicha(m.version2025, m.v2025.ID),
			},
			Entrantes: []grafo.AristaDeFicha{aristaDeFicha(m.parte, idNorma)},
		},
		idMunicipio: {
			Nodo: nodoDeFicha(m.municipio,
				map[string]any{grafo.DatoCodigoINE: "28074", grafo.DatoNombre: "Legan\xc3\xa9s"}),
			Salientes: []grafo.AristaDeFicha{},
			Entrantes: []grafo.AristaDeFicha{
				aristaDeFicha(m.enlace, idBloque),
				aristaDeFicha(m.perteneceAjeno, idOrganoAjeno),
				aristaDeFicha(m.pertenece, idOrgano),
				aristaDeFicha(m.perteneceMinusculas, idOrganoMinusculas),
			},
		},
	}

	for id, esperada := range esperadas {
		ficha, encontrado, err := lectura.Ficha(t.Context(), id)
		require.NoError(t, err)
		require.True(t, encontrado, id)
		assert.Equal(t, esperada, ficha, id)
		assertSinCuerpos(t, ficha)
	}

	ficha, encontrado, err := lectura.Ficha(t.Context(), idMunicipio+" ")
	require.NoError(t, err)
	assert.False(t, encontrado, "un id se compara byte a byte: con un espacio más no está")
	assert.Equal(t, grafo.Ficha{}, ficha)

	vacia, err := Leer(t.Context(), ConDirectorio(t.TempDir()))
	require.NoError(t, err)

	ficha, encontrado, err = vacia.Ficha(t.Context(), idNorma)
	require.NoError(t, err)
	assert.False(t, encontrado, "en el grafo vacío no hay ningún nodo")
	assert.Equal(t, grafo.Ficha{}, ficha)
	require.NoError(t, vacia.Close())
}

// nodoDeFicha es el nodo de una ficha tal como se espera a partir de su fila:
// los datos, con el tipo que les da JSON, y las fechas tal como se escribieron.
func nodoDeFicha(registro grafo.RegistroDeNodo, datos map[string]any) grafo.NodoDeFicha {
	return grafo.NodoDeFicha{
		ID:                 registro.ID,
		Tipo:               registro.Tipo,
		Datos:              datos,
		PrimeraObservacion: registro.PrimeraObservacion.FechaConsulta,
		UltimaObservacion:  registro.UltimaObservacion,
	}
}

// aristaDeFicha es una arista de una ficha tal como se espera a partir de su
// fila, con el id del otro extremo.
func aristaDeFicha(registro grafo.RegistroDeArista, otro string) grafo.AristaDeFicha {
	return grafo.AristaDeFicha{
		Relacion:           registro.Relacion,
		ID:                 otro,
		PrimeraObservacion: registro.PrimeraObservacion.FechaConsulta,
		UltimaObservacion:  registro.UltimaObservacion,
	}
}

// assertSinCuerpos comprueba que lo que un verbo devuelve, escrito como lo
// escribiría el sobre o como lo vería quien lo depure, no lleva el cuerpo de
// ningún texto (FR-070).
func assertSinCuerpos(t *testing.T, valor any) {
	t.Helper()

	escrito, err := json.Marshal(valor)
	require.NoError(t, err)

	for _, cuerpo := range []string{cuerpo2016, cuerpo2025} {
		assert.NotContains(t, string(escrito), cuerpo)
		assert.NotContains(t, fmt.Sprintf("%#v", valor), cuerpo)
	}
}

// TestRecuento fija lo que graph stats cuenta (FR-054; contracts/applet-graph.md
// §3.2; research.md D16): los totales y los pares (tipo, fuente) y (relación,
// fuente) por la fuente de la última observación, ordenados por bytes y sin
// ningún par a cero; con el grafo vacío, tres ceros y dos listas vacías.
func TestRecuento(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	crearGrafo(t, directorio, diarioWAL, muestra(t))

	lectura, err := Leer(t.Context(), ConDirectorio(directorio))
	require.NoError(t, err)

	recuento, err := lectura.Recuento(t.Context())
	require.NoError(t, err)
	assert.Equal(t, recuentoDeLaMuestra(), recuento)
	assertSinCuerpos(t, recuento)
	require.NoError(t, lectura.Close())

	vacia, err := Leer(t.Context(), ConDirectorio(t.TempDir()))
	require.NoError(t, err)

	recuento, err = vacia.Recuento(t.Context())
	require.NoError(t, err)
	assert.Equal(t, recuentoVacio(), recuento)

	escrito, err := json.Marshal(recuento)
	require.NoError(t, err)
	assert.JSONEq(t, `{"nodos":0,"aristas":0,"textos":0,"nodos_por_tipo":[],"aristas_por_relacion":[]}`,
		string(escrito))
	require.NoError(t, vacia.Close())
}

// TestInstantanea fija lo que graph check lee (data-model §5; research.md D15):
// cada nodo con sus datos, su última observación y la vigencia que declaró —cero
// si no declaró ninguna—, y cada arista por su terna, en un orden que no
// depende del motor: los nodos por id y las aristas por origen, relación y
// destino, comparando bytes. Nunca el cuerpo de un texto.
func TestInstantanea(t *testing.T) {
	t.Parallel()

	m := laMuestra(t)
	directorio := t.TempDir()
	crearGrafo(t, directorio, diarioWAL, m.grafo())

	lectura, err := Leer(t.Context(), ConDirectorio(directorio))
	require.NoError(t, err)

	t.Cleanup(func() { assert.NoError(t, lectura.Close()) })

	nodo := func(registro grafo.RegistroDeNodo, datos map[string]any) grafo.NodoDeInstantanea {
		return grafo.NodoDeInstantanea{
			ID: registro.ID, Tipo: registro.Tipo, Datos: datos,
			UltimaObservacion: registro.UltimaObservacion, Vigencia: registro.Vigencia,
		}
	}
	terna := func(registro grafo.RegistroDeArista) schema.Arista {
		return schema.Arista{Origen: registro.Origen, Relacion: registro.Relacion, Destino: registro.Destino}
	}

	esperada := grafo.Instantanea{
		Nodos: []grafo.NodoDeInstantanea{
			nodo(m.organoAjeno, map[string]any{grafo.DatoDIR3: idOrganoAjeno}),
			nodo(m.organo, map[string]any{grafo.DatoDIR3: idOrgano}),
			nodo(m.norma, map[string]any{grafo.DatoIdentificador: "BOE-A-2015-10565"}),
			nodo(m.bloque, map[string]any{grafo.DatoBloque: "a21"}),
			nodo(m.v2016, map[string]any{grafo.DatoFechaVigencia: "20161002", grafo.DatoHashTexto: m.texto2016.Huella}),
			nodo(m.v2025, map[string]any{grafo.DatoFechaVigencia: "20250101", grafo.DatoHashTexto: m.texto2025.Huella}),
			nodo(m.municipio, map[string]any{grafo.DatoCodigoINE: "28074", grafo.DatoNombre: "Legan\xc3\xa9s"}),
			nodo(m.organoMinusculas, map[string]any{grafo.DatoDIR3: idOrganoMinusculas}),
		},
		Aristas: []schema.Arista{
			terna(m.perteneceAjeno),
			terna(m.pertenece),
			terna(m.parte),
			terna(m.enlace),
			terna(m.version2016),
			terna(m.version2025),
			terna(m.perteneceMinusculas),
		},
	}

	instantanea, err := lectura.Instantanea(t.Context())
	require.NoError(t, err)
	assert.Equal(t, esperada, instantanea)
	assert.Equal(t, semana, instantanea.Nodos[2].Vigencia, "la vigencia declarada, en segundos")
	assert.Zero(t, instantanea.Nodos[0].Vigencia, "sin vigencia declarada, cero")
	assertSinCuerpos(t, instantanea)
}

// TestLecturaDeFilasDanadas fija que una fila que la entrega nunca escribiría
// —unos datos que no son un objeto JSON, una vigencia negativa o que no cabe en
// una duración— es una base dañada: «inesperado», sin inventar un valor.
func TestLecturaDeFilasDanadas(t *testing.T) {
	t.Parallel()

	casos := map[string]struct {
		sentencia string
		verbo     func(context.Context, *Lectura) error
	}{
		"datos que no son JSON, en la ficha": {
			sentencia: `UPDATE nodes SET props = 'no soy JSON' WHERE id = '` + idNorma + `'`,
			verbo: func(ctx context.Context, lectura *Lectura) error {
				_, _, err := lectura.Ficha(ctx, idNorma)

				return err
			},
		},
		"datos que no son JSON, en la instantánea": {
			sentencia: `UPDATE nodes SET props = '[1, 2]' WHERE id = '` + idNorma + `'`,
			verbo:     instantaneaDe,
		},
		"una vigencia negativa": {
			sentencia: `UPDATE nodes SET ttl = -5 WHERE id = '` + idNorma + `'`,
			verbo:     instantaneaDe,
		},
		"una vigencia que no cabe en una duración": {
			sentencia: `UPDATE nodes SET ttl = ` + strconv.FormatInt(math.MaxInt64, 10) + ` WHERE id = '` + idNorma + `'`,
			verbo:     instantaneaDe,
		},
	}

	for nombre, caso := range casos {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ruta := crearGrafo(t, directorio, diarioWAL, muestra(t))
			alterar(t, ruta, caso.sentencia)

			lectura, err := Leer(t.Context(), ConDirectorio(directorio))
			require.NoError(t, err)

			t.Cleanup(func() { assert.NoError(t, lectura.Close()) })

			compruebaFallo(t, caso.verbo(t.Context(), lectura), schema.ClaseInesperado, ruta,
				"grafo: "+strconv.Quote(ruta)+" no es una base de datos utilizable; no se modifica")
		})
	}
}

// instantaneaDe lee la instantánea y devuelve solo el fallo.
func instantaneaDe(ctx context.Context, lectura *Lectura) error {
	_, err := lectura.Instantanea(ctx)

	return err
}

// TestLecturaClose fija que Close es idempotente —también sobre el grafo vacío
// y sobre una lectura nula— y que, cerrada, ningún verbo devuelve un grafo:
// devuelven un fallo que lo dice, nunca un grafo vacío que no lo es.
func TestLecturaClose(t *testing.T) {
	t.Parallel()

	llena := t.TempDir()
	crearGrafo(t, llena, diarioWAL, muestra(t))

	for nombre, directorio := range map[string]string{"con la muestra": llena, "el grafo vacío": t.TempDir()} {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			lectura, err := Leer(t.Context(), ConDirectorio(directorio))
			require.NoError(t, err)
			require.NoError(t, lectura.Close())
			require.NoError(t, lectura.Close(), "cerrar dos veces no falla")

			_, _, err = lectura.Ficha(t.Context(), idNorma)
			compruebaCerrada(t, err)

			_, err = lectura.Recuento(t.Context())
			compruebaCerrada(t, err)

			_, err = lectura.Instantanea(t.Context())
			compruebaCerrada(t, err)
		})
	}

	t.Run("una lectura nula", func(t *testing.T) {
		t.Parallel()

		var nula *Lectura

		require.NoError(t, nula.Close())

		_, err := nula.Recuento(t.Context())
		compruebaCerrada(t, err)
	})
}

// compruebaCerrada comprueba el fallo de un verbo sobre una lectura cerrada.
func compruebaCerrada(t *testing.T, err error) {
	t.Helper()

	var fallo *Error

	require.ErrorAs(t, err, &fallo)
	assert.Equal(t, schema.ClaseInesperado, fallo.Clase())
	require.ErrorIs(t, err, errLecturaCerrada)
	assert.Contains(t, err.Error(), "la lectura ya est\xc3\xa1 cerrada")
}

// TestLeerSinOpcion fija que, sin ConDirectorio, la lectura resuelve world.db
// por la regla de la caché en el instante de leer, y que un directorio que no
// existe es el grafo vacío sin crearlo (FR-001, FR-004, FR-011).
//
// No declara t.Parallel(): usa t.Setenv, y el entorno es lo que mide.
func TestLeerSinOpcion(t *testing.T) {
	raiz := t.TempDir()
	llena := filepath.Join(raiz, "llena")
	require.NoError(t, os.Mkdir(llena, 0o700))
	crearGrafo(t, llena, diarioWAL, muestra(t))

	t.Setenv(cache.VariableDirectorio, llena)

	lectura, err := Leer(t.Context())
	require.NoError(t, err)
	compruebaContenido(t, lectura, true)
	require.NoError(t, lectura.Close())

	antes := huellasDelArbol(t, raiz)

	t.Setenv(cache.VariableDirectorio, filepath.Join(raiz, "no-existe"))

	lectura, err = Leer(t.Context())
	require.NoError(t, err)
	compruebaContenido(t, lectura, false)
	require.NoError(t, lectura.Close())

	assert.Equal(t, antes, huellasDelArbol(t, raiz), "el directorio que no existe sigue sin existir")

	t.Setenv(cache.VariableDirectorio, "")

	lectura, err = Leer(t.Context())
	assert.Nil(t, lectura)

	var fallo *Error

	require.ErrorAs(t, err, &fallo)
	assert.Equal(t, schema.ClaseArgumentos, fallo.Clase())
}

// tablaAjena es una base sin el esquema del grafo con una tabla que se llama
// como una de las suyas y tiene otra forma: si la lectura consultara las tablas
// de una base en la versión 0, fallaría.
const tablaAjena = `CREATE TABLE nodes (x)`

// versionPosterior es un esquema de una versión que este binario no conoce.
const versionPosterior = `CREATE TABLE schema_version (version INTEGER PRIMARY KEY, aplicada_en TEXT NOT NULL);
INSERT INTO schema_version VALUES (1, '2026-09-28T12:00:00Z'), (2, '2030-01-01T00:00:00Z')`

// conMuestra prepara la muestra en world.db con el diario y los permisos dados.
func conMuestra(diario string, permisos os.FileMode) func(*testing.T, string) string {
	return func(t *testing.T, raiz string) string {
		t.Helper()

		directorio := cacheVacia(t, raiz)
		ruta := crearGrafo(t, directorio, diario, muestra(t))
		require.NoError(t, os.Chmod(ruta, permisos))

		return directorio
	}
}

// conMuestraEnlazada prepara world.db como enlace a la muestra, en WAL, en otro
// directorio.
func conMuestraEnlazada(t *testing.T, raiz string) string {
	t.Helper()

	return muestraEnlazada(t, raiz, 0o600)
}

// conMuestraEnlazadaSinPermiso es conMuestraEnlazada con el destino en 0400.
func conMuestraEnlazadaSinPermiso(t *testing.T, raiz string) string {
	t.Helper()

	return muestraEnlazada(t, raiz, 0o400)
}

// muestraEnlazada prepara la muestra en otro directorio, con los permisos
// dados, y world.db como enlace a ella.
func muestraEnlazada(t *testing.T, raiz string, permisos os.FileMode) string {
	t.Helper()

	otro := filepath.Join(raiz, "otro")
	require.NoError(t, os.Mkdir(otro, 0o700))

	destino := filepath.Join(otro, "grafo.db")
	require.NoError(t, os.Rename(crearGrafo(t, otro, diarioWAL, muestra(t)), destino))
	require.NoError(t, os.Chmod(destino, permisos))

	directorio := cacheVacia(t, raiz)
	require.NoError(t, os.Symlink(destino, filepath.Join(directorio, "world.db")))

	return directorio
}

// conSentencias prepara world.db con el diario dado y las sentencias, sin el
// esquema del grafo, y los permisos dados.
func conSentencias(diario string, permisos os.FileMode, sentencias string) func(*testing.T, string) string {
	return func(t *testing.T, raiz string) string {
		t.Helper()

		directorio := cacheVacia(t, raiz)
		ruta := filepath.Join(directorio, "world.db")
		base := abrirBaseDePrueba(t, ruta, pragmaDelTramo)
		fijarDiario(t, base, diario)

		_, err := base.ExecContext(t.Context(), sentencias)
		require.NoError(t, err)
		require.NoError(t, base.Close())
		require.NoError(t, os.Chmod(ruta, permisos))

		return directorio
	}
}

// conBaseDanada prepara una base en modo rollback cuya cabecera declara más
// páginas de las que tiene el fichero, con el contador de cambios igual a
// version-valid-for para que SQLite se crea ese tamaño: al leer da
// SQLITE_CORRUPT y no cambia nada (research.md V41).
func conBaseDanada(t *testing.T, raiz string) string {
	t.Helper()

	directorio := cacheVacia(t, raiz)
	ruta := crearGrafo(t, directorio, diarioClasico, muestra(t))

	fichero := leerFichero(t, ruta)
	require.Equal(t, fichero[24:28], fichero[92:96], "premisa: el tamaño de la cabecera es válido")

	fichero[28], fichero[29], fichero[30], fichero[31] = 0, 0, 0x10, 0
	escribirFichero(t, ruta, fichero)

	return directorio
}

// crearGrafo escribe el grafo en world.db, dentro del directorio, con el
// esquema de la versión 1 y el diario dado, y lo cierra: no queda ningún
// auxiliar. Devuelve la ruta de world.db.
func crearGrafo(t *testing.T, directorio, diario string, g grafoDePrueba) string {
	t.Helper()

	ruta := filepath.Join(directorio, "world.db")
	base := abrirBaseDePrueba(t, ruta, pragmaDelTramo+"&_txlock=immediate")
	fijarDiario(t, base, diario)
	escribirGrafo(t, base, ruta, g)
	require.NoError(t, base.Close())

	for _, sufijo := range []string{sufijoWAL, sufijoMemoriaCompartida, sufijoDiario} {
		require.NoFileExists(t, ruta+sufijo, "al cerrar la última conexión no queda ningún auxiliar")
	}

	return ruta
}

// fijarDiario pone la base en el modo de diario dado y comprueba que lo está.
func fijarDiario(t *testing.T, base *sql.DB, diario string) {
	t.Helper()

	var modo string

	require.NoError(t, base.QueryRowContext(t.Context(), "PRAGMA journal_mode="+diario).Scan(&modo))
	require.Equal(t, diario, modo)
}

// escribirGrafo crea el esquema de la versión 1 y escribe las filas en una
// transacción, como lo haría una entrega.
func escribirGrafo(t *testing.T, base *sql.DB, ruta string, g grafoDePrueba) {
	t.Helper()

	tx := empezar(t, base)
	require.NoError(t, migrar(t.Context(), tx, ruta))

	for _, nodo := range g.nodos {
		_, err := tx.ExecContext(t.Context(), `INSERT INTO nodes (id, type, props, first_seen, first_source,
			first_url, last_seen, source, url, ttl) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			nodo.ID, nodo.Tipo, nodo.Datos, nodo.PrimeraObservacion.FechaConsulta, nodo.PrimeraObservacion.Fuente,
			nodo.PrimeraObservacion.URL, nodo.UltimaObservacion.FechaConsulta, nodo.UltimaObservacion.Fuente,
			nodo.UltimaObservacion.URL, segundos(nodo.Vigencia))
		require.NoError(t, err)
	}

	for _, arista := range g.aristas {
		_, err := tx.ExecContext(t.Context(), `INSERT INTO edges (src, rel, dst, first_seen, first_source,
			first_url, last_seen, source, url, ttl) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			arista.Origen, arista.Relacion, arista.Destino, arista.PrimeraObservacion.FechaConsulta,
			arista.PrimeraObservacion.Fuente, arista.PrimeraObservacion.URL, arista.UltimaObservacion.FechaConsulta,
			arista.UltimaObservacion.Fuente, arista.UltimaObservacion.URL, segundos(arista.Vigencia))
		require.NoError(t, err)
	}

	for _, texto := range g.textos {
		_, err := tx.ExecContext(t.Context(),
			`INSERT INTO texts (hash, body, fetched_at, source, url) VALUES (?, ?, ?, ?, ?)`,
			texto.Huella, texto.Cuerpo, texto.Procedencia.FechaConsulta, texto.Procedencia.Fuente,
			texto.Procedencia.URL)
		require.NoError(t, err)
	}

	require.NoError(t, tx.Commit())
}

// segundos es la columna ttl de una vigencia: NULL si no se declaró.
func segundos(vigencia time.Duration) any {
	if vigencia == 0 {
		return nil
	}

	return int64(vigencia / time.Second)
}

// alterar ejecuta una sentencia sobre world.db con otra conexión y la cierra.
func alterar(t *testing.T, ruta, sentencia string) {
	t.Helper()

	base := abrirBaseDePrueba(t, ruta, pragmaDelTramo)

	_, err := base.ExecContext(t.Context(), sentencia)
	require.NoError(t, err)
	require.NoError(t, base.Close())
}

// copiaDeUnEscritorAbierto deja en el destino una copia de world.db, su -wal y
// su -shm tal como los tiene un escritor que sigue abierto, con el grafo
// confirmado en el WAL y todavía no en world.db: al no haber ninguna otra
// conexión sobre la copia, su -wal es un -wal huérfano (research.md V36 D).
func copiaDeUnEscritorAbierto(t *testing.T, destino string, g grafoDePrueba) {
	t.Helper()

	ruta := filepath.Join(t.TempDir(), "world.db")
	escritor := abrirBaseDePrueba(t, ruta, pragmaDelTramo+"&_pragma=wal_autocheckpoint(0)&_txlock=immediate")
	fijarDiario(t, escritor, diarioWAL)
	escribirGrafo(t, escritor, ruta, g)

	for _, sufijo := range []string{"", sufijoWAL, sufijoMemoriaCompartida} {
		copiar(t, ruta+sufijo, filepath.Join(destino, "world.db"+sufijo))
	}

	require.NoError(t, escritor.Close())
}

// copiaDeUnDiarioCaliente deja en el destino una copia de world.db y de su
// diario tal como los tiene un escritor en modo rollback con una transacción a
// medias que ya volcó páginas en world.db: sin ningún bloqueo sobre la copia, su
// diario está caliente (research.md V43).
func copiaDeUnDiarioCaliente(t *testing.T, destino string) {
	t.Helper()

	ruta := crearGrafo(t, t.TempDir(), diarioClasico, muestra(t))
	escritor := abrirBaseDePrueba(t, ruta, pragmaDelTramo+"&_pragma=cache_size(2)&_txlock=immediate")
	tx := empezar(t, escritor)

	relleno := strings.Repeat("x", 4000)

	for i := range 100 {
		_, err := tx.ExecContext(t.Context(),
			`INSERT INTO texts (hash, body, fetched_at, source, url) VALUES (?, ?, ?, ?, ?)`,
			"relleno:"+strconv.Itoa(i), relleno, fechaDelBloque, fuenteDelBOE, urlDelBloque)
		require.NoError(t, err)
	}

	require.FileExists(t, ruta+sufijoDiario, "premisa: la transacción a medias tiene su diario")

	diario := leerFichero(t, ruta+sufijoDiario)
	require.NotEmpty(t, diario)
	require.NotZero(t, diario[0], "premisa: el diario ya tiene su cabecera, así que está caliente")

	for _, sufijo := range []string{"", sufijoDiario} {
		copiar(t, ruta+sufijo, filepath.Join(destino, "world.db"+sufijo))
	}

	require.NoError(t, tx.Rollback())
	require.NoError(t, escritor.Close())
}

// ceroBytesConUnWAL deja en el directorio un world.db de 0 bytes, con los
// permisos dados, junto al -wal no vacío de un escritor con marcos y, si se
// pide, su -shm (research.md V49).
func ceroBytesConUnWAL(t *testing.T, directorio string, conMemoriaCompartida bool, permisos os.FileMode) {
	t.Helper()

	escritor := t.TempDir()
	copiaDeUnEscritorAbierto(t, escritor, muestra(t))

	ruta := filepath.Join(directorio, "world.db")
	copiar(t, filepath.Join(escritor, "world.db"+sufijoWAL), ruta+sufijoWAL)

	if conMemoriaCompartida {
		copiar(t, filepath.Join(escritor, "world.db"+sufijoMemoriaCompartida), ruta+sufijoMemoriaCompartida)
	}

	require.NoError(t, os.WriteFile(ruta, nil, 0o600))
	require.NoError(t, os.Chmod(ruta, permisos))

	info, err := os.Stat(ruta + sufijoWAL)
	require.NoError(t, err)
	require.Positive(t, info.Size(), "premisa: el -wal no está vacío")
}

// leerLaVersionEn abre world.db en el modo dado, lee la versión y cierra: es lo
// que hacen los controles para comprobar qué haría otro modo con los mismos
// ficheros.
func leerLaVersionEn(t *testing.T, ruta string, modo modoDeLectura) {
	t.Helper()

	base, err := sql.Open(controladorSQLite, cadenaDeLectura(ruta, modo))
	require.NoError(t, err)

	_, err = versionRegistrada(t.Context(), base)
	require.NoError(t, err)
	require.NoError(t, base.Close())
}

// copiar copia un fichero byte a byte, con acceso reservado a la cuenta.
func copiar(t *testing.T, origen, destino string) {
	t.Helper()

	escribirFichero(t, destino, leerFichero(t, origen))
}

// leerFichero devuelve los bytes de un fichero de la prueba.
func leerFichero(t *testing.T, ruta string) []byte {
	t.Helper()

	contenido, err := os.ReadFile(filepath.Clean(ruta))
	require.NoError(t, err)

	return contenido
}

// escribirFichero escribe un fichero de la prueba con acceso reservado a la
// cuenta.
func escribirFichero(t *testing.T, ruta string, contenido []byte) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Clean(ruta), contenido, 0o600))
}

// huellasDeFicheros es la huella SHA-256 de cada fichero.
func huellasDeFicheros(t *testing.T, rutas ...string) map[string]string {
	t.Helper()

	huellas := map[string]string{}

	for _, ruta := range rutas {
		huellas[ruta] = huella(string(leerFichero(t, ruta)))
	}

	return huellas
}

// permisosSinEscritura son los de un directorio que no admite ficheros nuevos.
const permisosSinEscritura fs.FileMode = 0o500

// directorioSinEscritura quita el permiso de escritura al directorio y lo
// devuelve al terminar, antes de que t.TempDir lo retire. Comprueba la premisa:
// si el proceso puede crear un fichero dentro, corre con privilegios que se
// saltan los permisos y la prueba no mediría nada.
func directorioSinEscritura(t *testing.T, directorio string) {
	t.Helper()

	restringir(t, directorio, permisosSinEscritura)

	sonda := filepath.Join(directorio, "sonda")

	err := os.WriteFile(sonda, nil, 0o600)
	if err == nil {
		require.NoError(t, os.Remove(sonda))
	}

	require.ErrorIs(t, err, fs.ErrPermission, "premisa: el directorio no admite ficheros nuevos "+
		"(¿el proceso corre como root?)")
}

// restringir cambia los permisos de la ruta y devuelve los de antes al
// terminar la prueba.
func restringir(t *testing.T, ruta string, permisos fs.FileMode) {
	t.Helper()

	info, err := os.Stat(ruta)
	require.NoError(t, err)

	permisosDeAntes := info.Mode().Perm()

	require.NoError(t, os.Chmod(ruta, permisos))
	t.Cleanup(func() {
		assert.NoError(t, os.Chmod(ruta, permisosDeAntes), "se restauran los permisos de %s", ruta)
	})
}

// retenerElBloqueoExclusivo abre otra conexión sobre una base en modo rollback
// —la de otra invocación— y toma en ella el bloqueo exclusivo, que no deja leer
// a nadie hasta que termina la prueba.
func retenerElBloqueoExclusivo(t *testing.T, ruta string) {
	t.Helper()

	otra := abrirBaseDePrueba(t, ruta, pragmaDelTramo+"&_txlock=exclusive")

	tx, err := otra.BeginTx(context.WithoutCancel(t.Context()), nil)
	require.NoError(t, err, "la otra conexión toma el bloqueo exclusivo")

	t.Cleanup(func() {
		if err := tx.Rollback(); !errors.Is(err, sql.ErrTxDone) {
			assert.NoError(t, err)
		}
	})
}
