package app

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"maps"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/graph"
	"github.com/jmorenobl/kitlegal/internal/httpx"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// El registro local de estos tests: el applet graph compuesto con un reloj fijo
// y world.db en un directorio temporal, poblado por la API pública de
// internal/graph —la misma entrega que hace el kernel después de presentar—,
// nunca en el world.db de la cuenta de quien los ejecuta (plan.md, obligación
// 10). Todo es inventado a propósito —ids, fuentes, urls, fechas y el cuerpo del
// texto—, de modo que nada pueda tomarse por un dato real escrito de memoria; lo
// que observan de verdad boe y territorio lo ejercen los guiones de la suite de
// aceptación.
//
// La muestra son dos lotes, como los de dos invocaciones: uno con la forma de lo
// que emite boe articulo —una Norma, su Bloque y la versión del bloque, con sus
// dos aristas, el texto por su huella y una vigencia de una semana— y otro con
// la de territorio resolver —un Municipio y el Organo que pertenece a él, sin
// vigencia y con una fecha con desplazamiento, que show tiene que devolver
// carácter a carácter—.
const (
	fuenteDeLaNorma = "prueba.legislacion"
	urlDeLaNorma    = "https://legislacion.example/prueba/2026/1/bloque/a1"
	fechaDeLaNorma  = "2026-09-28T12:00:00Z"
	// caducidadDeLaNorma es fechaDeLaNorma más la vigencia de una semana.
	caducidadDeLaNorma = "2026-10-05T12:00:00Z"

	fuenteDelMunicipio = "kitlegal.prueba"
	urlDelMunicipio    = "kitlegal:applet/prueba"
	fechaDelMunicipio  = "2026-09-29T08:15:00+02:00"

	idDeLaNorma             = "eli/prueba/l/2026/1"
	idDelBloque             = idDeLaNorma + "#a1"
	idDelMunicipio          = "ine:99001"
	idDelOrgano             = "L01990011"
	idQueNoEsta             = "ine:00000"
	identificadorDeLaNorma  = "PRUEBA-2026-1"
	fechaDeVigenciaDePrueba = "20260101"

	// instanteDelGrafo es el del reloj fijo del registro local: la fecha de
	// consulta de cada sobre y el instante de check, posterior a la caducidad
	// de la norma.
	instanteDelGrafo = "2026-10-06T09:30:00Z"
	// instanteSinCaducar es anterior a la caducidad de la norma.
	instanteSinCaducar = "2026-10-01T00:00:00Z"

	// vigenciaDeLaNorma es la del primer lote: una semana.
	vigenciaDeLaNorma = 7 * 24 * time.Hour
)

// cuerpoDelBloque es el texto del bloque de la muestra, que guarda la entrega y
// que no sale por ningún verbo, con --json ni sin él (FR-070).
const cuerpoDelBloque = "Art\xc3\xadculo 1. Objeto de esta norma inventada.\n" +
	"Esta norma de prueba regula lo que ninguna otra regula.\n" +
	"Sus efectos empiezan el d\xc3\xada siguiente al de su publicaci\xc3\xb3n."

// firmaDelGrafo es la procedencia con la que firma graph, escrita aquí entera y
// no con las constantes del applet, para que un cambio en ellas no pase por
// aquí en silencio (FR-051; contracts/applet-graph.md §2).
var firmaDelGrafo = schema.Procedencia{Fuente: "kitlegal.graph", URL: "kitlegal:applet/graph"}

// Las descripciones literales de contracts/applet-graph.md §1, las de check con
// el 50 de H7.1 FR-010 escrito aquí y no con grafo.MaximoDeHallazgos, para que
// un cambio de la cota no pase por aquí en silencio.
const (
	descripcionDelGrafo = "Lee el grafo del mundo: lo que el binario ha observado de las fuentes, con su procedencia."
	ayudaDelID          = "Id del nodo: un ELI, \xc2\xabine:<c\xc3\xb3digo>\xc2\xbb, un DIR3\xe2\x80\xa6"
	descripcionDeCheck  = "Comprueba lo consultado de una norma, de algunos de sus bloques o, sin argumentos, todo" +
		" lo consultado, y lista como mucho 50 hallazgos: redacciones que han cambiado desde la lectura anterior" +
		" y consultas caducadas."
	ayudaDeLaNorma    = "Identificador BOE de la norma, BOE-A-<a\xc3\xb1o>-<n\xc3\xbamero>; sin \xc3\xa9l, todo lo consultado."
	ayudaDeLosBloques = "Ids de bloque de esa norma, como a21; sin ellos, todos los suyos."
)

// verboDelGrafo es un verbo del contrato: su nombre, su descripción literal y
// el valor cero del tipo de su data.
type verboDelGrafo struct {
	nombre      string
	descripcion string
	salida      any
}

// verbosDelGrafo son los tres del contrato, en su orden y ninguno por omisión.
func verbosDelGrafo() []verboDelGrafo {
	return []verboDelGrafo{
		{
			nombre:      "show",
			descripcion: "Devuelve un nodo del grafo del mundo con sus aristas y su procedencia, sin texto legal.",
			salida:      grafo.Ficha{},
		},
		{
			nombre: "stats",
			descripcion: "Cuenta los nodos, las aristas y los textos del grafo del mundo por tipo, relaci\xc3\xb3n" +
				" y fuente.",
			salida: grafo.Recuento{},
		},
		{nombre: "check", descripcion: descripcionDeCheck, salida: grafo.Comprobacion{}},
	}
}

// relojDePrueba es un reloj fijo que cuenta cuántas veces se lee: el applet lo
// lee una sola vez por invocación (contracts/applet-graph.md §1).
type relojDePrueba struct {
	instante time.Time
	lecturas int
}

func (r *relojDePrueba) ahora() time.Time {
	r.lecturas++

	return r.instante
}

// nuevoReloj es un reloj fijo en el instante, escrito en RFC 3339.
func nuevoReloj(t *testing.T, instante string) *relojDePrueba {
	t.Helper()

	fijo, err := time.Parse(time.RFC3339, instante)
	require.NoError(t, err)

	return &relojDePrueba{instante: fijo}
}

// registroDelGrafo es el registro local: el applet graph con el reloj dado y
// world.db en el directorio, y nada más.
func registroDelGrafo(t *testing.T, reloj func() time.Time, directorio string) *Registro {
	t.Helper()

	return registroConElGrafo(t, DependenciasDeGrafo{
		Reloj:   reloj,
		Almacen: []graph.Opcion{graph.ConDirectorio(directorio)},
	})
}

// registroConElGrafo es un registro con el applet graph compuesto con esas
// dependencias.
func registroConElGrafo(t *testing.T, dependencias DependenciasDeGrafo) *Registro {
	t.Helper()

	var registro Registro

	require.NoError(t, registro.Registrar(AppletGrafo(dependencias)))

	return &registro
}

// argvDelGrafo es la invocación de «kitlegal graph» con los argumentos.
func argvDelGrafo(argumentos ...string) []string {
	return slices.Concat([]string{"kitlegal", "graph"}, argumentos)
}

// huellaDelCuerpo es la del texto de la muestra.
func huellaDelCuerpo() string {
	return huellaDelTexto(cuerpoDelBloque)
}

// huellaDelTexto es la de un texto: sha256 de sus bytes.
func huellaDelTexto(cuerpo string) string {
	suma := sha256.Sum256([]byte(cuerpo))

	return "sha256:" + hex.EncodeToString(suma[:])
}

// idDeLaVersion es el de la BloqueVersion de la muestra.
func idDeLaVersion() string {
	return idDeUnaVersion(fechaDeVigenciaDePrueba, cuerpoDelBloque)
}

// idDeUnaVersion es el de la BloqueVersion del bloque de la muestra con esa
// fecha de vigencia y ese texto.
func idDeUnaVersion(fechaDeVigencia, cuerpo string) string {
	return idDelBloque + "@" + fechaDeVigencia + ":" + huellaDelTexto(cuerpo)
}

// lotesDeLaMuestra son los dos lotes de la muestra.
func lotesDeLaMuestra() []core.Lote {
	return []core.Lote{
		loteDelBloque(fechaDeLaNorma, fechaDeVigenciaDePrueba, cuerpoDelBloque),
		{
			Fuente: fuenteDelMunicipio, URL: urlDelMunicipio, FechaConsulta: fechaDelMunicipio,
			Operaciones: []schema.Operacion{
				schema.Nodo{ID: idDelMunicipio, Tipo: grafo.TipoMunicipio, Datos: map[string]any{
					grafo.DatoCodigoINE: "99001", grafo.DatoNombre: "Villaprueba",
				}},
				schema.Nodo{ID: idDelOrgano, Tipo: grafo.TipoOrgano, Datos: map[string]any{grafo.DatoDIR3: idDelOrgano}},
				schema.Arista{Origen: idDelOrgano, Relacion: grafo.RelacionPerteneceA, Destino: idDelMunicipio},
			},
		},
	}
}

// loteDelBloque es un lote con la forma de lo que emite boe articulo del bloque
// de la muestra en una redacción, consultada en esa fecha por la fuente de la
// norma y con una semana de vigencia: la Norma, el Bloque y la BloqueVersion de
// esa fecha de vigencia y ese texto, con sus dos aristas, y el texto por su
// huella (contracts/emision.md §1).
func loteDelBloque(fechaDeConsulta, fechaDeVigencia, cuerpo string) core.Lote {
	huella := huellaDelTexto(cuerpo)
	version := idDeUnaVersion(fechaDeVigencia, cuerpo)

	return core.Lote{
		Fuente: fuenteDeLaNorma, URL: urlDeLaNorma, FechaConsulta: fechaDeConsulta, Vigencia: vigenciaDeLaNorma,
		Operaciones: []schema.Operacion{
			schema.Nodo{
				ID: idDeLaNorma, Tipo: grafo.TipoNorma,
				Datos: map[string]any{grafo.DatoIdentificador: identificadorDeLaNorma},
			},
			schema.Nodo{ID: idDelBloque, Tipo: grafo.TipoBloque, Datos: map[string]any{grafo.DatoBloque: "a1"}},
			schema.Nodo{ID: version, Tipo: grafo.TipoBloqueVersion, Datos: map[string]any{
				grafo.DatoFechaVigencia: fechaDeVigencia, grafo.DatoHashTexto: huella,
			}},
			schema.Arista{Origen: idDeLaNorma, Relacion: grafo.RelacionTieneParte, Destino: idDelBloque},
			schema.Arista{Origen: idDelBloque, Relacion: grafo.RelacionTieneVersion, Destino: version},
			schema.Texto{Huella: huella, Cuerpo: cuerpo},
		},
	}
}

// poblarElGrafo entrega los dos lotes de la muestra al world.db del directorio
// y comprueba la premisa de las lecturas sin rastro: al cerrar la entrega no
// queda ningún auxiliar junto a world.db, que es el estado normal (FR-004).
func poblarElGrafo(t *testing.T, directorio string) {
	t.Helper()

	almacen := graph.Nuevo(graph.ConDirectorio(directorio))

	for _, lote := range lotesDeLaMuestra() {
		require.NoError(t, almacen.Apply(t.Context(), lote))
	}

	for _, auxiliar := range []string{"world.db-wal", "world.db-shm", "world.db-journal"} {
		require.NoFileExists(t, filepath.Join(directorio, auxiliar), "premisa: sin auxiliares")
	}
}

// procedenciaJSON es el objeto de una procedencia como lo escriben show y check.
func procedenciaJSON(fuente, url, fecha string) string {
	return fmt.Sprintf(`{"fuente": %q, "url": %q, "fecha_consulta": %q}`, fuente, url, fecha)
}

// fichaDelBloque es el data de show del Bloque: sus datos, su primera y su
// última observación, la arista saliente hacia su versión y la entrante desde
// su Norma, cada una con su procedencia (contracts/applet-graph.md §3.1).
func fichaDelBloque() string {
	norma := procedenciaJSON(fuenteDeLaNorma, urlDeLaNorma, fechaDeLaNorma)

	return fmt.Sprintf(`{
		"nodo": {"id": %q, "tipo": "Bloque", "datos": {"bloque": "a1"},
		         "primera_observacion": %q, "ultima_observacion": %s},
		"salientes": [{"relacion": "eli:has_version", "id": %q, "primera_observacion": %q, "ultima_observacion": %s}],
		"entrantes": [{"relacion": "eli:has_part", "id": %q, "primera_observacion": %q, "ultima_observacion": %s}]
	}`, idDelBloque, fechaDeLaNorma, norma, idDeLaVersion(), fechaDeLaNorma, norma, idDeLaNorma, fechaDeLaNorma, norma)
}

// fichaDelMunicipio es el data de show del Municipio: sin aristas salientes,
// que salen como una lista vacía, y con la entrante desde su Organo; las fechas,
// con su desplazamiento, carácter a carácter (FR-053).
func fichaDelMunicipio() string {
	municipio := procedenciaJSON(fuenteDelMunicipio, urlDelMunicipio, fechaDelMunicipio)

	return fmt.Sprintf(`{
		"nodo": {"id": %q, "tipo": "Municipio", "datos": {"codigo_ine": "99001", "nombre": "Villaprueba"},
		         "primera_observacion": %q, "ultima_observacion": %s},
		"salientes": [],
		"entrantes": [{"relacion": "lb:pertenece_a", "id": %q, "primera_observacion": %q, "ultima_observacion": %s}]
	}`, idDelMunicipio, fechaDelMunicipio, municipio, idDelOrgano, fechaDelMunicipio, municipio)
}

// recuentoDeLaMuestra es el data de stats sobre la muestra, con los pares por
// la fuente de la última observación y en orden de bytes (FR-054).
const recuentoDeLaMuestra = `{
	"nodos": 5, "aristas": 3, "textos": 1,
	"nodos_por_tipo": [
		{"tipo": "Bloque", "fuente": "prueba.legislacion", "nodos": 1},
		{"tipo": "BloqueVersion", "fuente": "prueba.legislacion", "nodos": 1},
		{"tipo": "Municipio", "fuente": "kitlegal.prueba", "nodos": 1},
		{"tipo": "Norma", "fuente": "prueba.legislacion", "nodos": 1},
		{"tipo": "Organo", "fuente": "kitlegal.prueba", "nodos": 1}
	],
	"aristas_por_relacion": [
		{"relacion": "eli:has_part", "fuente": "prueba.legislacion", "aristas": 1},
		{"relacion": "eli:has_version", "fuente": "prueba.legislacion", "aristas": 1},
		{"relacion": "lb:pertenece_a", "fuente": "kitlegal.prueba", "aristas": 1}
	]
}`

// recuentoVacio es el de stats con el grafo ausente o vacío: tres ceros y dos
// listas vacías, nunca nulas (FR-054).
const recuentoVacio = `{"nodos": 0, "aristas": 0, "textos": 0, "nodos_por_tipo": [], "aristas_por_relacion": []}`

// hallazgosDeLaMuestra es el data de check sin argumentos sobre la muestra en
// instanteDelGrafo: todo lo consultado, con una fuente caducada por cada nodo
// del primer lote, que declaró una semana de vigencia, y ninguna por los del
// segundo, que no declaró ninguna; en orden de id y con la cita de cada tipo
// (contracts/applet-graph.md §3, §4 y §5).
func hallazgosDeLaMuestra() string {
	procedencia := procedenciaJSON(fuenteDeLaNorma, urlDeLaNorma, fechaDeLaNorma)
	hallazgo := func(id, cita string) string {
		explicacion := "La consulta de " + cita + " a " + fuenteDeLaNorma + " en " + urlDeLaNorma + " del " +
			fechaDeLaNorma + " ten\xc3\xada una vigencia de 604800 s y caduc\xc3\xb3 el " + caducidadDeLaNorma + "."

		return fmt.Sprintf(`{"clase": "fuente-caducada", "id": %q, "explicacion": %q, "procedencia": %s,
			"vigencia_segundos": 604800}`, id, explicacion, procedencia)
	}
	citaDelBloque := "[" + identificadorDeLaNorma + ", bloque a1]"

	return comprobacionJSON("", nil, 0, 3,
		hallazgo(idDeLaNorma, identificadorDeLaNorma),
		hallazgo(idDelBloque, citaDelBloque),
		hallazgo(idDeLaVersion(), citaDelBloque),
	)
}

// comprobacionJSON es el data de check con ese ámbito —la norma y los bloques
// pedidos—, esos totales de version-obsoleta y de fuente-caducada y esos
// hallazgos, ya escritos en JSON; ninguno omitido, porque las muestras de estos
// tests dan muchos menos de 50 (contracts/applet-graph.md §3).
func comprobacionJSON(norma string, bloques []string, obsoletas, caducadas int, hallazgos ...string) string {
	pedidos := make([]string, 0, len(bloques))
	for _, bloque := range bloques {
		pedidos = append(pedidos, strconv.Quote(bloque))
	}

	return fmt.Sprintf(`{"norma": %q, "bloques": [%s], "version-obsoleta": %d, "fuente-caducada": %d,
		"omitidos": 0, "hallazgos": [%s]}`, norma, strings.Join(pedidos, ","), obsoletas, caducadas,
		strings.Join(hallazgos, ","))
}

// sinHallazgosEnTodo es el data de check sin argumentos cuando no hay nada que
// volver a comprobar.
func sinHallazgosEnTodo() string {
	return comprobacionJSON("", nil, 0, 0)
}

// datosFirmados comprueba un sobre correcto de graph —código 0, nada en la
// salida de error y la firma del applet con la fecha de su reloj— y devuelve su
// data como JSON (FR-051).
func datosFirmados(t *testing.T, res invocacionDePrueba, fecha string) string {
	t.Helper()

	require.Equal(t, 0, res.codigo, res.errores)
	assert.Empty(t, res.errores)

	sobre := sobreDelJSON(t, res.salida)

	assert.Equal(t, true, sobre["ok"])
	assert.Equal(t, firmaDelGrafo.Fuente, sobre["fuente"])
	assert.Equal(t, firmaDelGrafo.URL, sobre["url"])
	assert.Equal(t, fecha, sobre["fecha_consulta"])
	assert.Regexp(t, `\Asha256:[0-9a-f]{64}\z`, sobre["hash"])

	datos, err := json.Marshal(sobre["data"])
	require.NoError(t, err)

	return string(datos)
}

// exigirFalloDelGrafo comprueba un fallo que decide el applet: su código, su
// clase, la firma del applet con la fecha de su reloj y el mismo mensaje en el
// sobre y en la salida de error, que devuelve.
func exigirFalloDelGrafo(
	t *testing.T, res invocacionDePrueba, clase schema.Clase, codigo int, fecha string,
) string {
	t.Helper()

	exigirSobreDeFallo(t, res, clase, codigo, firmaDelGrafo)

	sobre := sobreDelJSON(t, res.salida)
	assert.Equal(t, fecha, sobre["fecha_consulta"], "el fallo lo fecha el reloj de la invocación")

	mensaje, esTexto := datosDelSobre(t, sobre)["mensaje"].(string)
	require.True(t, esTexto, "el mensaje es un texto")
	assert.Contains(t, res.errores, mensaje, "el mismo mensaje sale para la persona")

	return mensaje
}

// huellasDelDirectorio es cada entrada bajo el directorio por su nombre
// relativo, con su modo y, si es un fichero, la huella de su contenido: dos
// iguales dicen que nada ha cambiado, aparecido ni desaparecido (FR-004,
// FR-005).
func huellasDelDirectorio(t *testing.T, directorio string) map[string]string {
	t.Helper()

	sistema := os.DirFS(directorio)
	huellas := map[string]string{}

	err := fs.WalkDir(sistema, ".", func(nombre string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := entrada.Info()
		if err != nil {
			return err
		}

		descripcion := info.Mode().String()

		if info.Mode().IsRegular() {
			contenido, err := fs.ReadFile(sistema, nombre)
			if err != nil {
				return err
			}

			suma := sha256.Sum256(contenido)
			descripcion += " " + hex.EncodeToString(suma[:])
		}

		huellas[nombre] = descripcion

		return nil
	})
	require.NoError(t, err)

	return huellas
}

// comprobacionDelGrafo es un subtest de TestAppletGrafo.
type comprobacionDelGrafo struct {
	nombre    string
	comprobar func(t *testing.T)
}

// TestAppletGrafo fija el applet graph sobre un registro local, con un reloj
// fijo y world.db bajo t.TempDir() (contracts/applet-graph.md §1-§4 de H7 y de
// H7.1; research.md D19): su declaración y sus textos literales, la ayuda y
// --describe de cada verbo; el grafo ausente, que se lee como vacío sin crear
// nada; show, stats y check sobre la muestra, con la firma del applet y la fecha
// de su reloj, que se lee una vez por invocación y es el instante de check; los
// ids que no están (3) y los argumentos que no valen (2); check con su norma y
// sus bloques; --no-graph, --offline y el nombre del programa, que no cambian
// nada; la salida legible sin --json, sin nada del sobre ni texto legal, y el
// sobre de siempre con --json; --dry-run, que lee igual;
// el plazo agotado (4); y el reloj que falta (1). Ningún verbo cambia un byte
// del directorio.
func TestAppletGrafo(t *testing.T) {
	t.Parallel()

	for _, caso := range []comprobacionDelGrafo{
		{nombre: "lo-que-declara", comprobar: compruebaDeclaracionDelGrafo},
		{nombre: "dependencias-del-sistema", comprobar: compruebaDependenciasDelGrafoDelSistema},
		{nombre: "ayuda-y-describe", comprobar: compruebaAyudaDelGrafo},
		{nombre: "sin-world-db", comprobar: compruebaGrafoAusente},
		{nombre: "con-la-muestra", comprobar: compruebaGrafoPoblado},
		{nombre: "check-con-el-reloj-de-la-invocacion", comprobar: compruebaCheckConElReloj},
		{nombre: "no-encontrado", comprobar: compruebaIDsQueNoEstan},
		{nombre: "argumentos", comprobar: compruebaArgumentosDelGrafo},
		{nombre: "check-acotado", comprobar: compruebaCheckAcotado},
		{nombre: "misma-salida-con-no-graph-offline-y-multicall", comprobar: compruebaMismaSalidaDelGrafo},
		{nombre: "legible-sin-json-y-sin-texto-legal", comprobar: compruebaLegibleDelGrafo},
		{nombre: "dry-run", comprobar: compruebaEnsayoDelGrafo},
		{nombre: "plazo-agotado", comprobar: compruebaPlazoDelGrafo},
		{nombre: "sin-reloj", comprobar: compruebaGrafoSinReloj},
	} {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			caso.comprobar(t)
		})
	}
}

// compruebaDeclaracionDelGrafo fija lo que declara el applet: su nombre, su
// descripción y sus tres verbos con sus textos literales, en su orden, ninguno
// por omisión, cada uno con una fábrica de argumentos y el tipo de su data
// (contracts/applet-graph.md §1 y §3).
func compruebaDeclaracionDelGrafo(t *testing.T) {
	t.Helper()

	applet := AppletGrafo(DependenciasDelGrafoDelSistema())

	assert.Equal(t, "graph", applet.Nombre())
	assert.Equal(t, descripcionDelGrafo, applet.Descripcion())

	verbos := applet.Verbos()
	contrato := verbosDelGrafo()
	require.Len(t, verbos, len(contrato))

	for i, esperado := range contrato {
		verbo := verbos[i]

		assert.Equal(t, esperado.nombre, verbo.Nombre)
		assert.Equal(t, esperado.descripcion, verbo.Descripcion, verbo.Nombre)
		assert.False(t, verbo.PorOmision, "%s no es un verbo por omisión", verbo.Nombre)
		assert.IsType(t, esperado.salida, verbo.Salida, verbo.Nombre)
		require.NotNil(t, verbo.Argumentos, verbo.Nombre)

		primera, segunda := verbo.Argumentos(), verbo.Argumentos()
		assert.NotSame(t, primera, segunda, "%s: una instancia por invocación", verbo.Nombre)
	}
}

// compruebaDependenciasDelGrafoDelSistema fija las de la raíz de producción: el reloj
// del sistema y la regla de ubicación de la caché, sin ninguna opción.
func compruebaDependenciasDelGrafoDelSistema(t *testing.T) {
	t.Helper()

	dependencias := DependenciasDelGrafoDelSistema()

	require.NotNil(t, dependencias.Reloj)
	assert.Empty(t, dependencias.Almacen, "sin opciones: la regla de ubicación de la caché")

	antes := time.Now()
	leido := dependencias.Reloj()
	assert.WithinRange(t, leido, antes, time.Now(), "el reloj del sistema")
}

// compruebaAyudaDelGrafo fija la ayuda del applet y la de show, el applet sin
// verbo, que sale con 2 nombrando los tres, y --describe de cada verbo, que no
// ejecuta nada ni crea world.db (FR-050).
func compruebaAyudaDelGrafo(t *testing.T) {
	t.Helper()

	directorio := t.TempDir()
	reloj := nuevoReloj(t, instanteDelGrafo)
	registro := registroDelGrafo(t, reloj.ahora, directorio)

	ayuda := invocar(t, registro, argvDelGrafo("--help")...)
	require.Equal(t, 0, ayuda.codigo, ayuda.errores)
	assert.Regexp(t, `\Auso: graph <verbo> \[banderas\]\n`, ayuda.salida)

	for _, verbo := range verbosDelGrafo() {
		assert.Regexp(t, `(?m)^  `+verbo.nombre+` +`+regexp.QuoteMeta(verbo.descripcion)+`$`, ayuda.salida)
	}

	assert.NotContains(t, ayuda.salida, "por omisi\xc3\xb3n")

	sinVerbo := invocar(t, registro, argvDelGrafo()...)
	assert.Equal(t, 2, sinVerbo.codigo)
	assert.Contains(t, sinVerbo.errores, "verbos de graph: show, stats, check")

	show := invocar(t, registro, argvDelGrafo("show", "--help")...)
	require.Equal(t, 0, show.codigo, show.errores)
	assert.Contains(t, show.salida, "Usage: graph show <id> [flags]")
	assert.Regexp(t, `(?m)<id> +`+regexp.QuoteMeta(ayudaDelID)+`$`, show.salida)

	// La ayuda de check parte en líneas lo que no cabe en una: se busca con el
	// espacio en blanco reducido.
	check := invocar(t, registro, argvDelGrafo("check", "--help")...)
	require.Equal(t, 0, check.codigo, check.errores)
	assert.Regexp(t, `(?m)^Usage: graph check \[<norma> \[<bloques> \.\.\.\]\] \[flags\]$`, check.salida)

	for _, texto := range []string{descripcionDeCheck, ayudaDeLaNorma, ayudaDeLosBloques} {
		assert.Contains(t, blancosReducidos(check.salida), texto)
	}

	for _, argumentos := range [][]string{{"show", idDelBloque}, {"stats"}, {"check"}} {
		verbo := argumentos[0]
		descripcion := invocar(t, registro, argvDelGrafo(slices.Concat(argumentos, []string{"--describe"})...)...)
		require.Equal(t, 0, descripcion.codigo, descripcion.errores)

		var documento map[string]any
		require.NoError(t, json.Unmarshal([]byte(descripcion.salida), &documento), verbo)
		assert.Equal(t, "graph "+verbo, documento["title"])

		if verbo == "check" {
			compruebaEntradaDeCheck(t, documento)
		}
	}

	assert.Zero(t, reloj.lecturas, "ni la ayuda ni --describe ejecutan ningún verbo")

	entradas, err := os.ReadDir(directorio)
	require.NoError(t, err)
	assert.Empty(t, entradas, "ni la ayuda ni --describe crean nada")
}

// compruebaEntradaDeCheck fija la entrada que describe --describe de check: la
// norma, un texto, y los bloques, una lista de textos, ninguno de los dos en
// required, porque los dos son opcionales (contracts/applet-graph.md §1).
func compruebaEntradaDeCheck(t *testing.T, documento map[string]any) {
	t.Helper()

	assert.Equal(t, map[string]any{"type": "string"},
		valorDelEsquema(t, documento, "properties", "entrada", "properties", "norma"))
	assert.Equal(t, map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		valorDelEsquema(t, documento, "properties", "entrada", "properties", "bloques"))

	entrada, esObjeto := valorDelEsquema(t, documento, "properties", "entrada").(map[string]any)
	require.True(t, esObjeto, "la entrada es un objeto")

	// Sin required, ninguna propiedad es obligatoria: la lista vacía.
	requeridas, _ := entrada["required"].([]any)
	assert.NotContains(t, requeridas, "norma", "la norma es opcional")
	assert.NotContains(t, requeridas, "bloques", "los bloques son opcionales")
}

// compruebaGrafoAusente fija el grafo ausente —el directorio vacío y uno que ni
// siquiera existe—: stats da tres ceros y dos listas vacías, check una
// comprobación sin hallazgos y show un 3 que nombra el id, y ninguno crea
// world.db, sus auxiliares ni su directorio (FR-004, FR-053, FR-054, FR-060).
func compruebaGrafoAusente(t *testing.T) {
	t.Helper()

	vacio := t.TempDir()
	inexistente := filepath.Join(t.TempDir(), "no", "existe")

	for _, directorio := range []string{vacio, inexistente} {
		registro := registroDelGrafo(t, nuevoReloj(t, instanteDelGrafo).ahora, directorio)

		stats := invocar(t, registro, argvDelGrafo("stats", "--json")...)
		assert.JSONEq(t, recuentoVacio, datosFirmados(t, stats, instanteDelGrafo), directorio)

		check := invocar(t, registro, argvDelGrafo("check", "--json")...)
		assert.JSONEq(t, sinHallazgosEnTodo(), datosFirmados(t, check, instanteDelGrafo), directorio)

		show := invocar(t, registro, argvDelGrafo("show", idDelMunicipio, "--json")...)
		mensaje := exigirFalloDelGrafo(t, show, schema.ClaseNoEncontrado, 3, instanteDelGrafo)
		assert.Contains(t, mensaje, strconv.Quote(idDelMunicipio), "el mensaje nombra el id")
	}

	entradas, err := os.ReadDir(vacio)
	require.NoError(t, err)
	assert.Empty(t, entradas, "no se crea world.db ni ningún auxiliar")
	assert.NoDirExists(t, filepath.Dir(inexistente), "no se crea el directorio")
}

// compruebaGrafoPoblado fija show, stats y check sobre la muestra: el data
// entero de cada uno, la firma con la fecha del reloj, una sola lectura del
// reloj por invocación y el directorio igual byte a byte (FR-005, FR-051,
// FR-053, FR-054, FR-060).
func compruebaGrafoPoblado(t *testing.T) {
	t.Helper()

	directorio := t.TempDir()
	poblarElGrafo(t, directorio)

	antes := huellasDelDirectorio(t, directorio)
	reloj := nuevoReloj(t, instanteDelGrafo)
	registro := registroDelGrafo(t, reloj.ahora, directorio)

	casos := []struct {
		argv  []string
		datos string
	}{
		{argv: argvDelGrafo("show", idDelBloque, "--json"), datos: fichaDelBloque()},
		{argv: argvDelGrafo("show", idDelMunicipio, "--json"), datos: fichaDelMunicipio()},
		{argv: argvDelGrafo("stats", "--json"), datos: recuentoDeLaMuestra},
		{argv: argvDelGrafo("check", "--json"), datos: hallazgosDeLaMuestra()},
	}

	for i, caso := range casos {
		res := invocar(t, registro, caso.argv...)

		assert.JSONEq(t, caso.datos, datosFirmados(t, res, instanteDelGrafo), "%q", caso.argv)
		assert.Equal(t, i+1, reloj.lecturas, "%q lee el reloj una sola vez", caso.argv)
	}

	assert.Equal(t, antes, huellasDelDirectorio(t, directorio), "ningún verbo cambia nada")
}

// compruebaCheckConElReloj fija que check compara con el instante del reloj de
// la invocación, que es también la fecha del sobre (FR-051, FR-066, FR-067):
// antes de la caducidad y en su instante exacto no hay hallazgos; después, sí.
func compruebaCheckConElReloj(t *testing.T) {
	t.Helper()

	directorio := t.TempDir()
	poblarElGrafo(t, directorio)

	for _, caso := range []struct {
		instante string
		datos    string
	}{
		{instante: instanteSinCaducar, datos: sinHallazgosEnTodo()},
		{instante: caducidadDeLaNorma, datos: sinHallazgosEnTodo()},
		{instante: instanteDelGrafo, datos: hallazgosDeLaMuestra()},
	} {
		registro := registroDelGrafo(t, nuevoReloj(t, caso.instante).ahora, directorio)
		res := invocar(t, registro, argvDelGrafo("check", "--json")...)

		assert.JSONEq(t, caso.datos, datosFirmados(t, res, caso.instante), caso.instante)
	}
}

// compruebaIDsQueNoEstan fija el 3 de show con un id que no está en el grafo:
// uno que no es de ningún nodo, uno con espacios en blanco y algo más, que no
// se recorta, uno con U+200B y uno con un byte que no es UTF-8, que se buscan
// tal cual (FR-052, FR-053). El mensaje nombra el id con sus bytes, escritos
// con %q, y nada cambia.
//
// El grafo tiene, además de la muestra, el nodo cuyo id es el del byte que no
// es UTF-8 con U+FFFD en su lugar, que es el que se buscaría si el id llegara
// al applet como un string (el analizador cambia ese byte por U+FFFD), y show
// lo encuentra: que el id con el byte salga con 3 es que se busca con sus
// bytes y no con los de otro id.
func compruebaIDsQueNoEstan(t *testing.T) {
	t.Helper()

	directorio := t.TempDir()
	poblarElGrafo(t, directorio)

	conSustituto := idDeLaNorma + "\xef\xbf\xbd"
	require.NoError(t, graph.Nuevo(graph.ConDirectorio(directorio)).Apply(t.Context(), core.Lote{
		Fuente: fuenteDeLaNorma, URL: urlDeLaNorma, FechaConsulta: fechaDeLaNorma, Vigencia: vigenciaDeLaNorma,
		Operaciones: []schema.Operacion{schema.Nodo{
			ID: conSustituto, Tipo: grafo.TipoNorma,
			Datos: map[string]any{grafo.DatoIdentificador: identificadorDeLaNorma},
		}},
	}))

	antes := huellasDelDirectorio(t, directorio)
	registro := registroDelGrafo(t, nuevoReloj(t, instanteDelGrafo).ahora, directorio)

	var ficha struct {
		Nodo struct {
			ID string `json:"id"`
		} `json:"nodo"`
	}

	datos := datosFirmados(t, invocar(t, registro, argvDelGrafo("show", conSustituto, "--json")...), instanteDelGrafo)
	require.NoError(t, json.Unmarshal([]byte(datos), &ficha))
	require.Equal(t, conSustituto, ficha.Nodo.ID, "premisa: el id con U+FFFD está en el grafo")

	for _, id := range []string{
		idQueNoEsta,
		"a b",
		" " + idDeLaNorma,
		idDeLaNorma + " ",
		idDeLaNorma + "\xe2\x80\x8b",
		idDeLaNorma + "\xff",
	} {
		res := invocar(t, registro, argvDelGrafo("show", id, "--json")...)

		mensaje := exigirFalloDelGrafo(t, res, schema.ClaseNoEncontrado, 3, instanteDelGrafo)
		assert.Contains(t, mensaje, strconv.Quote(id), "el mensaje nombra el id %q con sus bytes", id)
	}

	assert.Equal(t, antes, huellasDelDirectorio(t, directorio))
}

// compruebaArgumentosDelGrafo fija el 2 de los argumentos que no valen
// (FR-052, FR-054, FR-060; H7.1 FR-004, FR-006, SC-009; contracts/applet-graph.md
// §2 y §4 de H7 y de H7.1): los que rechaza el analizador antes del applet —show
// sin id y con dos, stats con uno, una bandera desconocida—, firmados por el
// kernel; los ids de show que rechaza el applet —vacío, solo espacio en blanco o
// con un carácter de control—, firmados por él; y la norma y los bloques de
// check que rechaza el applet antes de abrir nada —una norma sin la forma
// BOE-A-<año>-<número>, también vacía y también con bloques detrás, y un bloque
// vacío o de solo espacio en blanco—, firmados por él con el mensaje de
// contracts/applet-graph.md §2, que nombra el valor. Nada cambia.
func compruebaArgumentosDelGrafo(t *testing.T) {
	t.Helper()

	directorio := t.TempDir()
	poblarElGrafo(t, directorio)

	antes := huellasDelDirectorio(t, directorio)
	registro := registroDelGrafo(t, nuevoReloj(t, instanteDelGrafo).ahora, directorio)

	for _, argv := range [][]string{
		{"show"},
		{"show", idDeLaNorma, idDelBloque},
		{"stats", idDeLaNorma},
		{"stats", "--no-existe"},
		{"check", "--no-existe"},
	} {
		res := invocar(t, registro, argvDelGrafo(slices.Concat(argv, []string{"--json"})...)...)

		exigirSobreDeFallo(t, res, schema.ClaseArgumentos, 2, cli.ProcedenciaKernel())
	}

	for _, id := range []string{"", " ", "\xc2\xa0", "\xe2\x80\x83", "\t", "\xc2\x85", "\x7f", "a\x00b"} {
		res := invocar(t, registro, argvDelGrafo("show", id, "--json")...)

		mensaje := exigirFalloDelGrafo(t, res, schema.ClaseArgumentos, 2, instanteDelGrafo)
		assert.Contains(t, mensaje, strconv.Quote(id), "el mensaje nombra el id %q", id)
	}

	const (
		normaSinSuForma = "argumentos inv\xc3\xa1lidos: la norma %q no tiene la forma BOE-A-<a\xc3\xb1o>-<n\xc3\xbamero>," +
			" con cuatro d\xc3\xadgitos en el a\xc3\xb1o y de uno a nueve en el n\xc3\xbamero"
		bloqueEnBlanco = "argumentos inv\xc3\xa1lidos: el bloque %q est\xc3\xa1 vac\xc3\xado o solo tiene espacio en blanco"
	)

	for _, caso := range []struct {
		argumentos []string
		mensaje    string
	}{
		{[]string{"a21"}, fmt.Sprintf(normaSinSuForma, "a21")},
		{[]string{idDeLaNorma}, fmt.Sprintf(normaSinSuForma, idDeLaNorma)},
		{[]string{"BOE-B-2015-10565"}, fmt.Sprintf(normaSinSuForma, "BOE-B-2015-10565")},
		{[]string{""}, fmt.Sprintf(normaSinSuForma, "")},
		{[]string{"", "a21"}, fmt.Sprintf(normaSinSuForma, "")},
		{[]string{"BOE-A-2015-10565", " "}, fmt.Sprintf(bloqueEnBlanco, " ")},
		{[]string{"BOE-A-2015-10565", ""}, fmt.Sprintf(bloqueEnBlanco, "")},
		{[]string{"BOE-A-2015-10565", "a21", "\t\xc2\xa0"}, fmt.Sprintf(bloqueEnBlanco, "\t\xc2\xa0")},
	} {
		res := invocar(t, registro, argvDelGrafo(slices.Concat([]string{"check"}, caso.argumentos, []string{"--json"})...)...)

		mensaje := exigirFalloDelGrafo(t, res, schema.ClaseArgumentos, 2, instanteDelGrafo)
		assert.Equal(t, caso.mensaje, mensaje, "%q", caso.argumentos)
	}

	assert.Equal(t, antes, huellasDelDirectorio(t, directorio))
}

// compruebaCheckAcotado fija lo que el applet hace con los argumentos de check
// que valen (H7.1 FR-001 a FR-005, FR-012, FR-014): sin norma, todo lo
// consultado; con una norma bien formada, la lectura acotada a ella, que aquí no
// trae nada porque la muestra no tiene ninguna norma del BOE —la acotada de
// verdad la fijan TestInstantanea y la suite de aceptación—, y la norma y los
// bloques pedidos, copiados en data tal cual y en su orden, con 0. Ninguna
// invocación cambia nada del directorio.
func compruebaCheckAcotado(t *testing.T) {
	t.Helper()

	directorio := t.TempDir()
	poblarElGrafo(t, directorio)

	antes := huellasDelDirectorio(t, directorio)
	registro := registroDelGrafo(t, nuevoReloj(t, instanteDelGrafo).ahora, directorio)

	for _, caso := range []struct {
		argumentos []string
		datos      string
	}{
		{nil, hallazgosDeLaMuestra()},
		{[]string{"BOE-A-2099-99999"}, comprobacionJSON("BOE-A-2099-99999", nil, 0, 0)},
		{[]string{"BOE-A-2099-99999", "a1"}, comprobacionJSON("BOE-A-2099-99999", []string{"a1"}, 0, 0)},
		{[]string{"BOE-A-2099-99999", "a99", "a1"}, comprobacionJSON("BOE-A-2099-99999", []string{"a99", "a1"}, 0, 0)},
	} {
		res := invocar(t, registro, argvDelGrafo(slices.Concat([]string{"check"}, caso.argumentos, []string{"--json"})...)...)

		assert.JSONEq(t, caso.datos, datosFirmados(t, res, instanteDelGrafo), "%q", caso.argumentos)
	}

	assert.Equal(t, antes, huellasDelDirectorio(t, directorio), "check no escribe nada")
}

// compruebaMismaSalidaDelGrafo fija que --no-graph y --offline no cambian nada
// —la misma lectura, la misma salida y el mismo código (FR-031;
// contracts/applet-graph.md §2)—, y que invocar el applet por el nombre del
// programa tampoco (FR-050): con --json y sin él, en los tres verbos y en sus
// códigos 0, 2 y 3.
func compruebaMismaSalidaDelGrafo(t *testing.T) {
	t.Helper()

	directorio := t.TempDir()
	poblarElGrafo(t, directorio)

	antes := huellasDelDirectorio(t, directorio)
	registro := registroDelGrafo(t, nuevoReloj(t, instanteDelGrafo).ahora, directorio)

	for _, argumentos := range [][]string{
		{"show", idDelBloque, "--json"},
		{"show", idDelBloque},
		{"show", idQueNoEsta, "--json"},
		{"show", " ", "--json"},
		{"stats", "--json"},
		{"stats"},
		{"check", "--json"},
		{"check"},
	} {
		referencia := sinRegistroDeEventos(invocar(t, registro, argvDelGrafo(argumentos...)...))

		for _, argv := range [][]string{
			argvDelGrafo(slices.Concat(argumentos, []string{"--no-graph"})...),
			argvDelGrafo(slices.Concat(argumentos, []string{"--offline"})...),
			argvDelGrafo(slices.Concat(argumentos, []string{"--offline", "--no-graph"})...),
			slices.Concat([]string{"/usr/local/bin/graph"}, argumentos),
		} {
			res := sinRegistroDeEventos(invocar(t, registro, argv...))

			assert.Equal(t, referencia, res, "%q da lo mismo que %q", argv, argumentos)
		}
	}

	assert.Equal(t, antes, huellasDelDirectorio(t, directorio))
}

// sinRegistroDeEventos es la invocación sin las líneas del registro de eventos
// en la salida de error, que llevan la hora y la duración de cada una: lo demás
// —el código, la salida estándar y el mensaje para la persona— es lo que tiene
// que coincidir.
func sinRegistroDeEventos(res invocacionDePrueba) invocacionDePrueba {
	lineas := slices.DeleteFunc(strings.SplitAfter(res.errores, "\n"), func(linea string) bool {
		return strings.HasPrefix(linea, "time=")
	})
	res.errores = strings.Join(lineas, "")

	return res
}

// delSobreEnLoLegible casa con lo que la salida legible no lleva (H7.1 FR-060,
// SC-010): las cuatro líneas de procedencia del sobre, la firma de graph y los
// pares ruta/valor aplanados de la tabla mínima.
var delSobreEnLoLegible = regexp.MustCompile(`(?m)^(fuente|url|fecha_consulta|hash)\s|kitlegal\.graph|` +
	`kitlegal:applet/graph|^(nodo|salientes|entrantes|nodos_por_tipo|aristas_por_relacion|hallazgos)\.`)

// statsLegibleDeLaMuestra es lo que cuenta stats de la muestra, alineado en
// columna (contracts/applet-graph.md §5.1).
const statsLegibleDeLaMuestra = "El grafo del mundo tiene 5 nodos, 3 aristas y 1 texto.\n" +
	"\n" +
	"Nodos por tipo y fuente:\n" +
	"  Bloque         prueba.legislacion  1\n" +
	"  BloqueVersion  prueba.legislacion  1\n" +
	"  Municipio      kitlegal.prueba     1\n" +
	"  Norma          prueba.legislacion  1\n" +
	"  Organo         kitlegal.prueba     1\n" +
	"\n" +
	"Aristas por relación y fuente:\n" +
	"  eli:has_part     prueba.legislacion  1\n" +
	"  eli:has_version  prueba.legislacion  1\n" +
	"  lb:pertenece_a   kitlegal.prueba     1\n"

// showLegibleDelMunicipio es lo que cuenta show del Municipio de la muestra:
// sus dos datos, sus fechas con su desplazamiento, carácter a carácter, ninguna
// arista saliente y la entrante desde su Organo (contracts/applet-graph.md
// §5.2).
func showLegibleDelMunicipio() string {
	observacion := fechaDelMunicipio + " · " + fuenteDelMunicipio + " · " + urlDelMunicipio

	return lineas(
		"Municipio "+idDelMunicipio,
		"  codigo_ine: 99001",
		"  nombre: Villaprueba",
		"Primera observación: "+fechaDelMunicipio,
		"Última observación: "+observacion,
		"",
		"Aristas salientes: ninguna.",
		"",
		"Aristas entrantes:",
		"  lb:pertenece_a ← "+idDelOrgano,
		"    última observación: "+observacion,
	)
}

// checkLegibleDeLaMuestra es lo que cuenta check sin argumentos de la muestra
// en instanteDelGrafo: las tres consultas caducadas del primer lote, en orden
// de id, y cómo acotar (contracts/applet-graph.md §5.3).
func checkLegibleDeLaMuestra() string {
	citaDelBloque := "[" + identificadorDeLaNorma + ", bloque a1]"

	return lineas(
		"Hallazgos en todo lo consultado: 0 version-obsoleta y 3 fuente-caducada; se listan 3 y se omiten 0.",
		"",
		"fuente-caducada (3):",
		"  - "+caducadaDePrueba(idDeLaNorma, identificadorDeLaNorma).Explicacion,
		"    "+idDeLaNorma,
		"  - "+caducadaDePrueba(idDelBloque, citaDelBloque).Explicacion,
		"    "+idDelBloque,
		"  - "+caducadaDePrueba(idDeLaVersion(), citaDelBloque).Explicacion,
		"    "+idDeLaVersion(),
		"",
		paraAcotar,
	)
}

// sobreDelGrafoEnJSON es el sobre que escribe --json para un data correcto de
// graph, byte a byte: la firma del applet con la fecha de su reloj, la huella de
// data y data compacta, en una línea (H7 FR-051, FR-053, FR-054).
func sobreDelGrafoEnJSON(t *testing.T, fecha, datos string) string {
	t.Helper()

	var compacto bytes.Buffer
	require.NoError(t, json.Compact(&compacto, []byte(datos)))

	huella, err := schema.Huella(json.RawMessage(compacto.Bytes()))
	require.NoError(t, err)

	return `{"ok":true,"fuente":"` + firmaDelGrafo.Fuente + `","url":"` + firmaDelGrafo.URL + `","fecha_consulta":"` +
		fecha + `","hash":"` + huella + `","data":` + compacto.String() + "}\n"
}

// compruebaLegibleDelGrafo fija la salida sin --json (H7.1 FR-060 a FR-063,
// SC-010; contracts/applet-graph.md §5): stats, show y check sobre la muestra
// dan el texto legible byte a byte —check con hallazgos, sin ellos y acotado,
// y con la versión obsoleta delante de las consultas caducadas—, sin nada del
// sobre, sin pares aplanados y sin ninguna línea del texto guardado, y nada en
// la salida de error; un fallo sigue sin salida estándar y con su mensaje en la
// de error. Con --json, stats y show dan el mismo sobre que antes del hito, byte
// a byte (FR-060).
func compruebaLegibleDelGrafo(t *testing.T) {
	t.Helper()

	directorio := t.TempDir()
	poblarElGrafo(t, directorio)

	registro := registroDelGrafo(t, nuevoReloj(t, instanteDelGrafo).ahora, directorio)
	sinCaducar := registroDelGrafo(t, nuevoReloj(t, instanteSinCaducar).ahora, directorio)

	for _, caso := range []struct {
		registro   *Registro
		argumentos []string
		texto      string
	}{
		{registro, []string{"stats"}, statsLegibleDeLaMuestra},
		{registro, []string{"show", idDelBloque}, showLegibleDePruebaDelBloque()},
		{registro, []string{"show", idDelMunicipio}, showLegibleDelMunicipio()},
		{registro, []string{"check"}, checkLegibleDeLaMuestra()},
		{sinCaducar, []string{"check"}, lineas("No hay nada que volver a comprobar en todo lo consultado.", "", paraAcotar)},
		{registro, []string{"check", normaDePrueba, "a1"}, lineas("No hay nada que volver a comprobar de " +
			normaDePrueba + ", bloque a1.")},
	} {
		res := invocar(t, caso.registro, argvDelGrafo(caso.argumentos...)...)

		require.Equal(t, 0, res.codigo, res.errores)
		assert.Empty(t, res.errores, "%q", caso.argumentos)
		assert.Equal(t, caso.texto, res.salida, "%q", caso.argumentos)
		assert.NotRegexp(t, delSobreEnLoLegible, res.salida, "%q", caso.argumentos)

		for _, linea := range strings.Split(cuerpoDelBloque, "\n") {
			assert.NotContains(t, res.salida, linea, "%q no devuelve texto legal", caso.argumentos)
		}
	}

	for _, caso := range []struct {
		argumentos []string
		datos      string
	}{
		{[]string{"stats"}, recuentoDeLaMuestra},
		{[]string{"show", idDelBloque}, fichaDelBloque()},
		{[]string{"show", idDelMunicipio}, fichaDelMunicipio()},
	} {
		res := invocar(t, registro, argvDelGrafo(slices.Concat(caso.argumentos, []string{"--json"})...)...)

		require.Equal(t, 0, res.codigo, res.errores)
		assert.Equal(t, sobreDelGrafoEnJSON(t, instanteDelGrafo, caso.datos), res.salida, "%q", caso.argumentos)
	}

	for _, caso := range []struct {
		argumentos []string
		codigo     int
		mensaje    string
	}{
		{[]string{"show", idQueNoEsta}, 3, strconv.Quote(idQueNoEsta)},
		{[]string{"check", "a21"}, 2, `la norma "a21"`},
	} {
		res := invocar(t, registro, argvDelGrafo(caso.argumentos...)...)

		assert.Equal(t, caso.codigo, res.codigo, "%q", caso.argumentos)
		assert.Empty(t, res.salida, "%q: un fallo no tiene forma legible", caso.argumentos)
		assert.Contains(t, res.errores, caso.mensaje, "%q: su mensaje va a la salida de error", caso.argumentos)
	}

	compruebaLegibleConLasDosClases(t)
}

// compruebaLegibleConLasDosClases fija el orden de los grupos de check sin
// --json con hallazgos de las dos clases: con la versión posterior leída, el de
// version-obsoleta, sobre la anterior, va delante del de fuente-caducada
// (FR-063).
func compruebaLegibleConLasDosClases(t *testing.T) {
	t.Helper()

	directorio := t.TempDir()
	poblarConUnaVersionPosterior(t, directorio)

	res := invocar(t, registroDelGrafo(t, nuevoReloj(t, instanteDelGrafo).ahora, directorio), argvDelGrafo("check")...)

	require.Equal(t, 0, res.codigo, res.errores)
	assert.Regexp(t, `\AHallazgos en todo lo consultado: 1 version-obsoleta y 3 fuente-caducada; se listan 4 y se`+
		` omiten 0\.\n\nversion-obsoleta \(1\):\n  - La versión de .+\n    `+regexp.QuoteMeta(idDeLaVersion())+
		`\nfuente-caducada \(3\):\n(  - .+\n    .+\n){3}\n`+regexp.QuoteMeta(paraAcotar)+`\n\z`, res.salida)
	assert.NotRegexp(t, delSobreEnLoLegible, res.salida)
}

// compruebaEnsayoDelGrafo fija --dry-run (contracts/applet-graph.md §2): los
// verbos leen igual, sin describir ninguna operación, y el kernel escribe su
// línea en la salida de error sin sobre; un fallo sigue con su código. Nada
// cambia.
func compruebaEnsayoDelGrafo(t *testing.T) {
	t.Helper()

	directorio := t.TempDir()
	poblarElGrafo(t, directorio)

	antes := huellasDelDirectorio(t, directorio)
	reloj := nuevoReloj(t, instanteDelGrafo)
	registro := registroDelGrafo(t, reloj.ahora, directorio)

	for _, argumentos := range [][]string{{"show", idDelBloque}, {"stats"}, {"check"}} {
		res := invocar(t, registro, argvDelGrafo(slices.Concat(argumentos, []string{"--json", "--dry-run"})...)...)

		assert.Equal(t, 0, res.codigo, res.errores)
		assert.Empty(t, res.salida, "con --dry-run no hay sobre")
		assert.True(t, strings.HasPrefix(res.errores, prefijoDeEnsayo), "la línea del kernel: %q", res.errores)
		assert.NotContains(t, res.errores, "se habr\xc3\xada pedido", "graph no pide nada")
	}

	assert.Equal(t, 3, reloj.lecturas, "cada verbo se ejecuta y lee el reloj")

	noEsta := invocar(t, registro, argvDelGrafo("show", idQueNoEsta, "--dry-run")...)
	assert.Equal(t, 3, noEsta.codigo)
	assert.True(t, strings.HasPrefix(noEsta.errores, prefijoDeEnsayo), "la línea del kernel: %q", noEsta.errores)

	assert.Equal(t, antes, huellasDelDirectorio(t, directorio))
}

// compruebaPlazoDelGrafo fija el 4 del plazo de --timeout agotado mientras se
// lee world.db (contracts/applet-graph.md §4): el fallo de internal/graph, de
// clase fuente-no-disponible, llega con su mensaje, y nada cambia.
func compruebaPlazoDelGrafo(t *testing.T) {
	t.Helper()

	directorio := t.TempDir()
	poblarElGrafo(t, directorio)

	antes := huellasDelDirectorio(t, directorio)
	registro := registroDelGrafo(t, nuevoReloj(t, instanteDelGrafo).ahora, directorio)

	for _, argumentos := range [][]string{{"show", idDelBloque}, {"stats"}, {"check"}} {
		res := invocar(t, registro, argvDelGrafo(slices.Concat(argumentos, []string{"--json", "--timeout", "1ns"})...)...)

		mensaje := exigirFalloDelGrafo(t, res, schema.ClaseFuenteNoDisponible, 4, instanteDelGrafo)
		assert.Contains(t, mensaje, "grafo: el plazo termin\xc3\xb3 antes de leer "+
			strconv.Quote(filepath.Join(directorio, "world.db")), "%q", argumentos)
	}

	assert.Equal(t, antes, huellasDelDirectorio(t, directorio))
}

// compruebaGrafoSinReloj fija el reloj nulo, que es un defecto de composición
// (contracts/applet-graph.md §1): los tres verbos salen con 1, sin procedencia
// del applet —firma el kernel— y sin abrir ni crear nada.
func compruebaGrafoSinReloj(t *testing.T) {
	t.Helper()

	directorio := t.TempDir()
	registro := registroConElGrafo(t, DependenciasDeGrafo{Almacen: []graph.Opcion{graph.ConDirectorio(directorio)}})

	for _, argumentos := range [][]string{{"show", idDelBloque}, {"stats"}, {"check"}} {
		res := invocar(t, registro, argvDelGrafo(slices.Concat(argumentos, []string{"--json"})...)...)

		exigirSobreDeFallo(t, res, schema.ClaseInesperado, 1, cli.ProcedenciaKernel())
		assert.Contains(t, res.errores, "sin reloj", "%q", argumentos)
	}

	entradas, err := os.ReadDir(directorio)
	require.NoError(t, err)
	assert.Empty(t, entradas)
}

// TestCodigosDelGrafo fija los códigos de lo que no deja leer world.db, con las
// dependencias de la raíz de producción y la regla de ubicación de la caché
// (contracts/applet-graph.md §4; H7 FR-011, FR-012; H7.1 FR-070, SC-011):
//
//   - world.db que no es una base de datos —el vehículo de la regla genérica— o
//     cuyo esquema es de una versión posterior: 1 en los tres verbos, con y sin
//     --no-graph, con el mensaje que nombra su ruta y el directorio igual byte a
//     byte;
//   - KITLEGAL_CACHE_DIR presente y vacía, o que nombra un fichero, y sin ella y
//     sin HOME: 2 en los tres verbos, con y sin --no-graph, con el mensaje que
//     dice que no se puede ubicar world.db.
//
// La tabla va entera en secuencia, sin t.Parallel: cada caso fija el entorno
// con t.Setenv, que es justo lo que mide, y cada subprueba lo declara en su
// cuerpo, también HOME, de modo que lo que haya en la máquina de quien ejecuta
// el test no cambie ningún resultado ni reciba ninguna escritura.
func TestCodigosDelGrafo(t *testing.T) {
	t.Setenv(cache.VariableDirectorio, t.TempDir())

	for _, caso := range casosInutilizables() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())

			directorio := t.TempDir()
			t.Setenv(cache.VariableDirectorio, directorio)

			ruta := filepath.Join(directorio, "world.db")
			caso.preparar(t, ruta)

			antes := huellasDelDirectorio(t, directorio)

			compruebaCodigosDelSistema(t, schema.ClaseInesperado, 1, func(t *testing.T, mensaje string) {
				t.Helper()

				esperado := "grafo: " + strconv.Quote(ruta) + " " + caso.motivo
				if caso.conCausa {
					assert.Regexp(t, `\A`+regexp.QuoteMeta(esperado)+`: .+\z`, mensaje)

					return
				}

				assert.Equal(t, esperado, mensaje)
			})

			assert.Equal(t, antes, huellasDelDirectorio(t, directorio), "world.db con su huella")
		})
	}

	for _, caso := range casosDeUbicacion() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())

			caso.preparar(t)

			compruebaCodigosDelSistema(t, schema.ClaseArgumentos, 2, func(t *testing.T, mensaje string) {
				t.Helper()

				assert.True(t, strings.HasPrefix(mensaje, "grafo: no se puede ubicar world.db: "), mensaje)
				assert.Contains(t, mensaje, caso.nombra)
			})
		})
	}
}

// casoInutilizable es un world.db que no deja leerse: cómo se prepara y el
// motivo con el que lo nombra el mensaje, detrás de su ruta
// (contracts/almacen-world-db.md §6).
type casoInutilizable struct {
	nombre   string
	preparar func(t *testing.T, ruta string)
	motivo   string
	// conCausa dice que el mensaje sigue, detrás del motivo, con «: » y la
	// causa que da SQLite, que no se fija aquí: la regla genérica (H7.1 FR-070).
	conCausa bool
}

// casosInutilizables son el de H7.1 SC-011, lo que no es una base de datos, y
// el esquema posterior de H7 FR-012.
func casosInutilizables() []casoInutilizable {
	return []casoInutilizable{
		{
			nombre: "no-es-una-base",
			preparar: func(t *testing.T, ruta string) {
				t.Helper()

				escribirFicheroDePrueba(t, ruta, noEsUnaBase)
			},
			motivo:   "no es una base de datos utilizable",
			conCausa: true,
		},
		{
			nombre: "esquema-de-una-version-posterior",
			preparar: func(t *testing.T, ruta string) {
				t.Helper()

				escribirFicheroDePrueba(t, ruta, baseDeUnaVersionPosterior(t))
			},
			motivo: "tiene el esquema en la versi\xc3\xb3n 3 y este binario conoce la 2: no se modifica",
		},
	}
}

// casoDeUbicacion es un entorno en el que la regla de la caché no da un
// directorio: cómo se prepara y lo que el mensaje nombra para corregirlo
// (contracts/almacen-world-db.md §2).
type casoDeUbicacion struct {
	nombre   string
	preparar func(t *testing.T)
	nombra   string
}

// casosDeUbicacion son los tres de FR-011.
func casosDeUbicacion() []casoDeUbicacion {
	return []casoDeUbicacion{
		{
			nombre: "variable-vacia",
			preparar: func(t *testing.T) {
				t.Helper()

				t.Setenv(cache.VariableDirectorio, "")
			},
			nombra: cache.VariableDirectorio,
		},
		{
			nombre: "variable-que-nombra-un-fichero",
			preparar: func(t *testing.T) {
				t.Helper()

				fichero := filepath.Join(t.TempDir(), "fichero")
				contenido := []byte("No es un directorio.\n")
				escribirFicheroDePrueba(t, fichero, contenido)
				t.Setenv(cache.VariableDirectorio, fichero)

				t.Cleanup(func() {
					leido, err := os.ReadFile(filepath.Clean(fichero))
					assert.NoError(t, err)
					assert.Equal(t, contenido, leido, "el fichero no cambia")
				})
			},
			nombra: cache.VariableDirectorio,
		},
		{
			nombre: "sin-variable-ni-home",
			preparar: func(t *testing.T) {
				t.Helper()

				t.Setenv(cache.VariableDirectorio, "")
				require.NoError(t, os.Unsetenv(cache.VariableDirectorio))
				t.Setenv("HOME", "")
				t.Setenv("USERPROFILE", "")
			},
			nombra: "declara HOME o " + cache.VariableDirectorio,
		},
	}
}

// compruebaCodigosDelSistema invoca los tres verbos, con y sin --no-graph, con
// las dependencias de la raíz de producción, y exige a cada uno el código, la
// clase, la firma del applet y el mensaje que comprueba mensajeEsperado.
func compruebaCodigosDelSistema(
	t *testing.T, clase schema.Clase, codigo int, mensajeEsperado func(t *testing.T, mensaje string),
) {
	t.Helper()

	registro := registroConElGrafo(t, DependenciasDelGrafoDelSistema())

	for _, argumentos := range [][]string{{"show", idDelMunicipio}, {"stats"}, {"check"}} {
		for _, banderas := range [][]string{{"--json"}, {"--json", "--no-graph"}} {
			res := invocar(t, registro, argvDelGrafo(slices.Concat(argumentos, banderas)...)...)

			exigirSobreDeFallo(t, res, clase, codigo, firmaDelGrafo)

			mensaje, esTexto := datosDelSobre(t, sobreDelJSON(t, res.salida))["mensaje"].(string)
			require.True(t, esTexto, "el mensaje es un texto")
			assert.Contains(t, res.errores, mensaje)
			mensajeEsperado(t, mensaje)
		}
	}
}

// La versión posterior del bloque de la muestra: otra redacción, con una fecha
// de vigencia posterior, que la misma fuente observa después, como la
// observaría boe articulo tras el cambio. Leída después de la de la muestra,
// la supera (H7.1 FR-023) y pasa a ser la redacción vista del bloque; en
// instanteDelGrafo ya ha caducado su consulta, que renovó la de la Norma y la
// del Bloque, así que check da una version-obsoleta sobre la versión de la
// muestra y una fuente-caducada sobre la Norma, el Bloque y esta, y ninguna
// sobre la de la muestra, que ya no es la vista (H7.1 FR-030).
const (
	fechaDeLaVersionPosterior  = "2026-09-29T09:00:00Z"
	fechaDeVigenciaPosterior   = "20270101"
	cuerpoDeLaVersionPosterior = cuerpoDelBloque + "\nSu redacci\xc3\xb3n posterior de prueba a\xc3\xb1ade esta frase."
)

// poblarConUnaVersionPosterior entrega la muestra y, después, el lote de la
// versión posterior al world.db del directorio: dos lecturas sucesivas del
// bloque, cada una en su propia entrega, la última de la versión posterior.
func poblarConUnaVersionPosterior(t *testing.T, directorio string) {
	t.Helper()

	poblarElGrafo(t, directorio)

	lote := loteDelBloque(fechaDeLaVersionPosterior, fechaDeVigenciaPosterior, cuerpoDeLaVersionPosterior)
	require.NoError(t, graph.Nuevo(graph.ConDirectorio(directorio)).Apply(t.Context(), lote))
}

// sinWorldDB deja el directorio como está, vacío: el grafo ausente.
func sinWorldDB(t *testing.T, _ string) {
	t.Helper()
}

// noEsUnaBase es el contenido de un world.db que no es una base de datos
// SQLite: el vehículo de la regla genérica (H7.1 FR-070).
var noEsUnaBase = []byte("Este fichero no es una base de datos SQLite.\n")

// conWorldDBQueNoEsUnaBase pone en el lugar de world.db un fichero que no es
// una base de datos, que no se puede leer (contracts/almacen-world-db.md §6).
func conWorldDBQueNoEsUnaBase(t *testing.T, directorio string) {
	t.Helper()

	escribirFicheroDePrueba(t, filepath.Join(directorio, "world.db"), noEsUnaBase)
}

// TestSalidaDelGrafoContraSchemas es el punto 4 de la Definition of Done sobre
// el applet graph (FR-051, FR-094; contracts/applet-graph.md §3, §6 y §7): el
// sobre real que emite el kernel con --json, sobre el registro local, valida
// contra la parte de su verbo leída de schemas/grafo.json, y no contra lo que
// emite --describe mientras se ejecuta el test. Lo hace toda salida correcta de
// show —la de cada tipo de nodo, con aristas en los dos sentidos y con alguna
// lista vacía—, de stats —con el grafo ausente y con nodos— y de check —con
// hallazgos de las dos clases, de una sola, sin hallazgos, con el grafo ausente
// y acotado a una norma y a sus bloques—, y también la de cada fallo que decide
// el applet, con 2 —también los de la norma y los bloques de check—, 3, 4 y 1
// (gates/supuestos.md, T017); los que decide el kernel antes de llegar al
// applet no son salida suya. La validación restringe: el mismo sobre con una
// clave de más o de menos en su data no valida, y en check tampoco con un
// elemento que no es un hallazgo ni con una lista nula (H7.1 FR-012, FR-082).
func TestSalidaDelGrafoContraSchemas(t *testing.T) {
	t.Parallel()

	esquemas := map[string]*jsonschema.Schema{}

	for _, verbo := range verbosDelGrafo() {
		publicado, id := ficheroPublicadoDelVerbo(t, verbo.nombre)
		require.Equal(t, raizDeLosEsquemas+"grafo.json", id, "la parte de %q la publica grafo.json", verbo.nombre)

		esquemas[verbo.nombre] = salidaPublicada(t, publicado, id, verbo.nombre)
	}

	for _, caso := range salidasDelGrafo() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			caso.prepara(t, directorio)

			registro := registroDelGrafo(t, nuevoReloj(t, caso.instante).ahora, directorio)
			res := invocar(t, registro, argvDelGrafo(slices.Concat(caso.argumentos, []string{"--json"})...)...)

			require.Equal(t, caso.codigo, res.codigo, res.errores)

			sobre := sobreDelJSON(t, res.salida)
			assert.Equal(t, caso.codigo == 0, sobre["ok"], "ok decide la rama del esquema contra la que se valida data")
			assert.Equal(t, firmaDelGrafo.Fuente, sobre["fuente"], "la salida es del applet y no del kernel")

			for _, rasgo := range caso.rasgos {
				assert.Contains(t, res.salida, rasgo, "la salida es la que nombra el caso")
			}

			exigirSalidaDelGrafoPublicada(t, esquemas[caso.argumentos[0]], res.salida)
		})
	}
}

// salidaDelGrafo es una invocación de TestSalidaDelGrafoContraSchemas: cómo se
// prepara el directorio de world.db, el instante del reloj, el código con el
// que termina, los argumentos sin --json, el primero de ellos el verbo, y los
// fragmentos que su salida tiene que llevar para que el caso sea el que dice, y
// no otra salida que también valide.
type salidaDelGrafo struct {
	nombre     string
	prepara    func(t *testing.T, directorio string)
	instante   string
	codigo     int
	argumentos []string
	rasgos     []string
}

// salidasDelGrafo son las invocaciones cuyo sobre se valida.
func salidasDelGrafo() []salidaDelGrafo {
	sinAristas := func(sentido string) string { return `"` + sentido + `":[]` }
	caducada := `"clase":"` + string(grafo.ClaseFuenteCaducada) + `"`
	obsoleta := `"clase":"` + string(grafo.ClaseVersionObsoleta) + `"`
	sinHallazgos := `"version-obsoleta":0,"fuente-caducada":0,"omitidos":0,"hallazgos":[]`
	enTodo := `"data":{"norma":"","bloques":[],`

	return []salidaDelGrafo{
		{
			"show-norma", poblarElGrafo, instanteDelGrafo, 0,
			[]string{"show", idDeLaNorma},
			[]string{`"tipo":"Norma"`, sinAristas("entrantes"), `"relacion":"eli:has_part"`},
		},
		{
			"show-bloque", poblarElGrafo, instanteDelGrafo, 0,
			[]string{"show", idDelBloque},
			[]string{`"tipo":"Bloque"`, `"relacion":"eli:has_part"`, `"relacion":"eli:has_version"`},
		},
		{
			"show-bloque-version", poblarElGrafo, instanteDelGrafo, 0,
			[]string{"show", idDeLaVersion()},
			[]string{`"tipo":"BloqueVersion"`, sinAristas("salientes"), `"hash_texto":"` + huellaDelCuerpo() + `"`},
		},
		{
			"show-municipio", poblarElGrafo, instanteDelGrafo, 0,
			[]string{"show", idDelMunicipio},
			[]string{`"tipo":"Municipio"`, sinAristas("salientes"), fechaDelMunicipio},
		},
		{
			"show-organo", poblarElGrafo, instanteDelGrafo, 0,
			[]string{"show", idDelOrgano},
			[]string{`"tipo":"Organo"`, sinAristas("entrantes"), `"relacion":"lb:pertenece_a"`},
		},
		{
			"show-bloque-con-dos-versiones", poblarConUnaVersionPosterior, instanteDelGrafo, 0,
			[]string{"show", idDelBloque},
			[]string{idDeLaVersion(), idDeUnaVersion(fechaDeVigenciaPosterior, cuerpoDeLaVersionPosterior)},
		},

		{
			"stats-sin-world-db", sinWorldDB, instanteDelGrafo, 0,
			[]string{"stats"},
			[]string{`"nodos":0`, `"nodos_por_tipo":[]`, `"aristas_por_relacion":[]`},
		},
		{"stats-con-la-muestra", poblarElGrafo, instanteDelGrafo, 0, []string{"stats"}, []string{`"nodos":5`}},
		{
			"stats-con-una-version-posterior", poblarConUnaVersionPosterior, instanteDelGrafo, 0,
			[]string{"stats"},
			[]string{`"nodos":6`, `"textos":2`},
		},

		{"check-sin-world-db", sinWorldDB, instanteDelGrafo, 0, []string{"check"}, []string{enTodo + sinHallazgos}},
		{
			"check-sin-hallazgos", poblarElGrafo, instanteSinCaducar, 0,
			[]string{"check"},
			[]string{enTodo + sinHallazgos},
		},
		{
			"check-con-consultas-caducadas", poblarElGrafo, instanteDelGrafo, 0,
			[]string{"check"},
			[]string{enTodo + `"version-obsoleta":0,"fuente-caducada":3,`, caducada},
		},
		{
			"check-con-hallazgos-de-las-dos-clases", poblarConUnaVersionPosterior, instanteDelGrafo, 0,
			[]string{"check"},
			[]string{
				enTodo + `"version-obsoleta":1,"fuente-caducada":3,`, caducada, obsoleta,
				`"fecha_vigencia_reciente":"` + fechaDeVigenciaPosterior + `"`,
				`"id":"` + idDeUnaVersion(fechaDeVigenciaPosterior, cuerpoDeLaVersionPosterior) + `"`,
			},
		},
		{
			"check-acotado-a-una-norma-y-sus-bloques", poblarElGrafo, instanteDelGrafo, 0,
			[]string{"check", "BOE-A-2099-99999", "a1", "a99"},
			[]string{`"data":{"norma":"BOE-A-2099-99999","bloques":["a1","a99"],` + sinHallazgos},
		},

		{
			"fallo-2-id-en-blanco", poblarElGrafo, instanteDelGrafo, 2,
			[]string{"show", " "},
			[]string{`"clase":"argumentos"`},
		},
		{
			"fallo-2-norma-sin-su-forma", poblarElGrafo, instanteDelGrafo, 2,
			[]string{"check", "a21"},
			[]string{`"clase":"argumentos"`, `la norma \"a21\"`},
		},
		{
			"fallo-2-bloque-en-blanco", poblarElGrafo, instanteDelGrafo, 2,
			[]string{"check", "BOE-A-2099-99999", " "},
			[]string{`"clase":"argumentos"`, `el bloque \" \"`},
		},
		{
			"fallo-3-id-que-no-esta", poblarElGrafo, instanteDelGrafo, 3,
			[]string{"show", idQueNoEsta},
			[]string{`"clase":"no-encontrado"`},
		},
		{
			"fallo-3-sin-world-db", sinWorldDB, instanteDelGrafo, 3,
			[]string{"show", idDelBloque},
			[]string{`"clase":"no-encontrado"`},
		},
		{
			"fallo-4-plazo-agotado", poblarElGrafo, instanteDelGrafo, 4,
			[]string{"check", "--timeout", "1ns"},
			[]string{`"clase":"fuente-no-disponible"`},
		},
		{
			"fallo-1-world-db-no-es-una-base", conWorldDBQueNoEsUnaBase, instanteDelGrafo, 1,
			[]string{"stats"},
			[]string{`"clase":"inesperado"`},
		},
	}
}

// exigirSalidaDelGrafoPublicada exige que el sobre real valide contra la parte
// publicada de su verbo y que la validación restrinja data, que es siempre un
// objeto —la ficha de show, el recuento de stats, la comprobación de check o
// los datos de un fallo— y lo exige exigirSalidaPublicada; si es la
// comprobación de check, exigirHallazgosPublicados exige además lo de sus
// listas.
func exigirSalidaDelGrafoPublicada(t *testing.T, esquema *jsonschema.Schema, salida string) {
	t.Helper()

	exigirSalidaPublicada(t, esquema, salida)

	if _, esComprobacion := objetoDeData(t, sobreValidable(t, salida))["hallazgos"]; esComprobacion {
		exigirHallazgosPublicados(t, esquema, salida)
	}
}

// Las claves de un hallazgo (contracts/applet-graph.md §3.3; data-model §5): las
// cuatro de todo hallazgo, que el esquema publicado exige, y las propias de su
// clase, que llevan omitempty y que el esquema por eso admite sin exigirlas
// (research.md V7). Cada hallazgo lleva las suyas y ninguna de la otra clase.
var (
	clavesDeTodoHallazgo = []string{"clase", "explicacion", "id", "procedencia"}
	clavesDeSuClase      = map[grafo.ClaseDeHallazgo][]string{
		grafo.ClaseFuenteCaducada:  {"vigencia_segundos"},
		grafo.ClaseVersionObsoleta: {"fecha_vigencia", "fecha_vigencia_reciente"},
	}
)

// exigirHallazgosPublicados exige que el sobre real de check valide contra su
// parte publicada y que la validación restrinja sus listas: no valida con los
// bloques ni con los hallazgos nulos (FR-060; H7.1 FR-012), ni con un elemento
// de más en los hallazgos que no es un hallazgo —lo que también se comprueba
// con la lista vacía—, ni con una clave de más en cualquiera de sus hallazgos,
// ni sin cualquiera de las cuatro de todo hallazgo. Cada hallazgo lleva
// exactamente esas cuatro y las de su clase.
func exigirHallazgosPublicados(t *testing.T, esquema *jsonschema.Schema, salida string) {
	t.Helper()

	require.NoError(t, esquema.Validate(sobreValidable(t, salida)), "el sobre real valida contra su parte de schemas/")

	for _, lista := range []string{"bloques", "hallazgos"} {
		nula := sobreValidable(t, salida)
		objetoDeData(t, nula)[lista] = nil
		require.Errorf(t, esquema.Validate(nula), "la lista %q nula no valida", lista)
	}

	conUnoDeMas := sobreValidable(t, salida)
	objetoDeData(t, conUnoDeMas)["hallazgos"] = append(hallazgosDelSobre(t, conUnoDeMas),
		map[string]any{"ajena": "no declarada"})
	require.Error(t, esquema.Validate(conUnoDeMas), "un elemento que no es un hallazgo no valida")

	for posicion := range hallazgosDelSobre(t, sobreValidable(t, salida)) {
		hallazgo := hallazgoDelSobre(t, sobreValidable(t, salida), posicion)

		clase, esTexto := hallazgo["clase"].(string)
		require.True(t, esTexto, "la clase del hallazgo %d es un texto", posicion)
		assert.ElementsMatch(t, slices.Concat(clavesDeTodoHallazgo, clavesDeSuClase[grafo.ClaseDeHallazgo(clase)]),
			slices.Collect(maps.Keys(hallazgo)), "el hallazgo %d lleva las claves de su clase", posicion)

		conClaveDeMas := sobreValidable(t, salida)
		hallazgoDelSobre(t, conClaveDeMas, posicion)["ajena"] = "no declarada"
		require.Errorf(t, esquema.Validate(conClaveDeMas), "el hallazgo %d con una clave de más no valida", posicion)

		for _, clave := range clavesDeTodoHallazgo {
			sinLaClave := sobreValidable(t, salida)
			delete(hallazgoDelSobre(t, sinLaClave, posicion), clave)
			assert.Errorf(t, esquema.Validate(sinLaClave), "el hallazgo %d sin la clave %q no valida", posicion, clave)
		}
	}
}

// hallazgosDelSobre es la lista de hallazgos del data del sobre de check.
func hallazgosDelSobre(t *testing.T, sobre map[string]any) []any {
	t.Helper()

	hallazgos, esLista := objetoDeData(t, sobre)["hallazgos"].([]any)
	require.True(t, esLista, "los hallazgos de check son una lista")

	return hallazgos
}

// hallazgoDelSobre es el hallazgo de esa posición en el data del sobre de check.
// Pertenece al sobre, así que cambiarlo cambia el sobre.
func hallazgoDelSobre(t *testing.T, sobre map[string]any, posicion int) map[string]any {
	t.Helper()

	hallazgo, esObjeto := hallazgosDelSobre(t, sobre)[posicion].(map[string]any)
	require.True(t, esObjeto, "el hallazgo %d es un objeto", posicion)

	return hallazgo
}

// grabacionesDeEvals es la carpeta de la fuente boe con las respuestas que
// grabaron H5 y H5.1, relativa a este paquete: la de evals.GrabacionesDeH5, que
// resuelve igual desde internal/app, escrita aquí porque internal/evals importa
// este paquete y no se puede importar desde sus tests.
const grabacionesDeEvals = "../../testdata/evals/" + boe.NombreDeLaFuente

// minimoDeCaracteres es la longitud a partir de la cual una línea del texto de
// un bloque no puede aparecer en la salida de graph (FR-070; SC-006).
const minimoDeCaracteres = 20

// avisoDeEntregaFallida es el comienzo de la línea que el kernel escribe si lo
// observado no llega al grafo (contracts/resultado-y-entrega.md §4).
const avisoDeEntregaFallida = "kitlegal: lo observado no ha llegado al grafo del mundo"

// TestNingunVerboDelGrafoDevuelveTexto fija FR-070 y SC-006 sobre lo que boe
// observa de verdad: entrega al world.db de un directorio temporal, por el
// kernel y con el applet boe servido desde la reproducción, cada respuesta
// grabada de un bloque —las de H4 y las de H5 y H5.1—, con articulo y con
// articulos, y exige que ninguna línea no vacía de 20 caracteres o más del texto
// de ningún bloque aparezca en la salida de show de cada nodo del grafo, de
// stats ni de check, con --json ni sin él. Con --json se busca en la salida tal
// cual y en cada texto del documento, clave o valor, ya sin los escapes de JSON;
// sin ella, en la salida legible, que escribe los textos tal cual y no lleva
// nada del sobre ni pares aplanados (H7.1 FR-060, SC-010).
//
// Las premisas dicen que no pasa en vacío: cada línea buscada aparece, con la
// misma búsqueda, en la tabla del boe articulo que la devolvió; el grafo guarda
// un texto por cada huella que boe publicó y una BloqueVersion con cada una; y
// check, en un instante en que todas las consultas han caducado, lista
// hallazgos —como mucho 50—, cuyas explicaciones citan bloques.
func TestNingunVerboDelGrafoDevuelveTexto(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	textos := entregarLasRespuestasGrabadas(t, directorio)

	var lineas []string

	for _, texto := range textos {
		lineas = append(lineas, lineasDelTexto(texto)...)
	}

	require.NotEmpty(t, lineas, "premisa: hay líneas que buscar")

	instantanea := instantaneaDelGrafo(t, directorio)
	compruebaLoEntregado(t, instantanea, textos)

	registro := registroDelGrafo(t, nuevoReloj(t, instanteDelGrafo).ahora, directorio)

	invocaciones := [][]string{{"stats"}, {"check"}}
	for _, nodo := range instantanea.Nodos {
		invocaciones = append(invocaciones, []string{"show", nodo.ID})
	}

	for _, argumentos := range invocaciones {
		for _, conJSON := range []bool{true, false} {
			argv := argumentos
			if conJSON {
				argv = slices.Concat(argumentos, []string{"--json"})
			}

			res := invocar(t, registro, argvDelGrafo(argv...)...)
			require.Equal(t, 0, res.codigo, res.errores)

			switch {
			case slices.Equal(argv, []string{"stats", "--json"}):
				var recuento grafo.Recuento
				require.NoError(t, json.Unmarshal([]byte(datosFirmados(t, res, instanteDelGrafo)), &recuento))
				assert.Equal(t, len(textos), recuento.Textos, "premisa: el grafo guarda un texto por cada huella")
			case slices.Equal(argv, []string{"check", "--json"}):
				var comprobacion grafo.Comprobacion
				require.NoError(t, json.Unmarshal([]byte(datosFirmados(t, res, instanteDelGrafo)), &comprobacion))
				assert.NotEmpty(t, comprobacion.Hallazgos, "premisa: check lista hallazgos")
			case !conJSON:
				assert.NotRegexp(t, delSobreEnLoLegible, res.salida, "%q es la salida legible", argv)
			}

			assert.Empty(t, lineasQueAparecen(t, lineas, res.salida, conJSON), "%q no devuelve texto legal", argv)
		}
	}
}

// entregarLasRespuestasGrabadas invoca boe articulo por cada respuesta grabada
// de un bloque y boe articulos por cada norma con los bloques grabados de ella,
// sobre un registro que entrega al world.db del directorio, y devuelve el texto
// de cada bloque por su huella. Cada invocación sale con 0 y sin la línea de una
// entrega fallida; articulos devuelve de cada bloque el mismo texto que
// articulo, y cada línea que se buscará aparece en la tabla de boe articulo.
func entregarLasRespuestasGrabadas(t *testing.T, directorio string) map[string]string {
	t.Helper()

	textos := map[string]string{}

	for _, carpeta := range []string{grabacionesDeBoe, grabacionesDeEvals} {
		bloques := bloquesGrabados(t, carpeta)
		require.NotEmpty(t, bloques, "premisa: %s tiene respuestas de bloques", carpeta)

		porNorma := map[string][]string{}
		registro := registroDeBoeQueEntrega(t, carpeta, directorio)

		for _, grabado := range bloques {
			res := invocarBoeEntregando(t, registro, "articulo", grabado.norma, grabado.bloque, "--json")
			texto, huella := textoDelArticulo(t, sobreDelJSON(t, res.salida)["data"])
			textos[huella] = texto

			lineas := lineasDelTexto(texto)
			tabla := invocarBoeEntregando(t, registro, "articulo", grabado.norma, grabado.bloque)
			assert.Equal(t, lineas, lineasQueAparecen(t, lineas, tabla.salida, false),
				"premisa: la misma búsqueda encuentra cada línea en la tabla de boe articulo %v", grabado)

			porNorma[grabado.norma] = append(porNorma[grabado.norma], grabado.bloque)
		}

		registro = registroDeBoeQueEntrega(t, carpeta, directorio)

		for _, norma := range slices.Sorted(maps.Keys(porNorma)) {
			argumentos := slices.Concat([]string{"articulos", norma}, porNorma[norma], []string{"--json"})
			res := invocarBoeEntregando(t, registro, argumentos...)

			articulos, esLista := sobreDelJSON(t, res.salida)["data"].([]any)
			require.True(t, esLista, "el data de articulos es una lista")
			require.Len(t, articulos, len(porNorma[norma]))

			for _, articulo := range articulos {
				texto, huella := textoDelArticulo(t, articulo)
				assert.Equal(t, textos[huella], texto, "articulos devuelve el texto que devolvió articulo")
			}
		}
	}

	return textos
}

// bloqueGrabado es el bloque de una norma cuya respuesta está grabada, con el
// estado HTTP con el que la sirvió la fuente.
type bloqueGrabado struct {
	norma  string
	bloque string
	estado int
}

// direccionDeUnBloque es la de la API de la que sale el texto de un bloque, con
// la norma y el bloque.
var direccionDeUnBloque = regexp.MustCompile(
	`\Ahttps://www\.boe\.es/datosabiertos/api/legislacion-consolidada/id/([^/]+)/texto/bloque/([^/?#]+)\z`)

// bloquesGrabados son los bloques cuya respuesta grabada en la carpeta es un
// texto, en el orden de sus ficheros: cada grabación de un bloque que la fuente
// sirvió con 200. La de un bloque que la fuente no tiene, servida con 404, no
// devuelve texto y queda fuera.
func bloquesGrabados(t *testing.T, carpeta string) []bloqueGrabado {
	t.Helper()

	return slices.DeleteFunc(respuestasDeBloques(t, carpeta), func(grabado bloqueGrabado) bool {
		return grabado.estado != 200
	})
}

// respuestasDeBloques son los bloques con una respuesta grabada en la carpeta,
// la sirviera la fuente con el estado que la sirviera, en el orden de sus
// ficheros.
func respuestasDeBloques(t *testing.T, carpeta string) []bloqueGrabado {
	t.Helper()

	grabaciones := os.DirFS(carpeta)

	entradas, err := fs.ReadDir(grabaciones, ".")
	require.NoError(t, err)

	var bloques []bloqueGrabado

	for _, entrada := range entradas {
		contenido, err := fs.ReadFile(grabaciones, entrada.Name())
		require.NoError(t, err)

		var grabacion struct {
			Peticion struct {
				URL string `json:"url"`
			} `json:"peticion"`
			Respuesta struct {
				Estado int `json:"estado"`
			} `json:"respuesta"`
		}
		require.NoError(t, json.Unmarshal(contenido, &grabacion), entrada.Name())

		partes := direccionDeUnBloque.FindStringSubmatch(grabacion.Peticion.URL)
		if partes == nil {
			continue
		}

		bloques = append(bloques, bloqueGrabado{norma: partes[1], bloque: partes[2], estado: grabacion.Respuesta.Estado})
	}

	return bloques
}

// registroDeBoeQueEntrega es el registro de producción en lo que aquí importa:
// el applet boe servido desde la reproducción de la carpeta, con una caché
// nueva, y la entrega al world.db del directorio.
func registroDeBoeQueEntrega(t *testing.T, carpeta, directorio string) *Registro {
	t.Helper()

	return registroConLaEntrega(t, directorio, AppletBoe(nuevoBancoDeBoe(t, carpeta).dependencias))
}

// registroConLaEntrega es un registro con los applets y la entrega al world.db
// del directorio, la misma que compone la raíz de producción con la ubicación
// de la caché.
func registroConLaEntrega(t *testing.T, directorio string, applets ...Applet) *Registro {
	t.Helper()

	var registro Registro

	for _, applet := range applets {
		require.NoError(t, registro.Registrar(applet))
	}

	registro.EntregarAlGrafo(graph.Nuevo(graph.ConDirectorio(directorio)))

	return &registro
}

// invocarBoeEntregando invoca boe con los argumentos sobre el registro y exige
// que salga con 0 y que lo observado llegue al grafo.
func invocarBoeEntregando(t *testing.T, registro *Registro, argumentos ...string) invocacionDePrueba {
	t.Helper()

	res := invocar(t, registro, argvDeBoe(argumentos...)...)
	require.Equal(t, 0, res.codigo, res.errores)
	require.NotContains(t, res.errores, avisoDeEntregaFallida, "%q", argumentos)

	return res
}

// textoDelArticulo es el texto de un artículo del data de boe y su huella.
func textoDelArticulo(t *testing.T, data any) (string, string) {
	t.Helper()

	articulo, esObjeto := data.(map[string]any)
	require.True(t, esObjeto, "el artículo es un objeto")

	texto, esTexto := articulo["texto"].(string)
	require.True(t, esTexto, "el texto del artículo es un texto")

	huella, esTexto := articulo["hash_texto"].(string)
	require.True(t, esTexto, "la huella del artículo es un texto")

	return texto, huella
}

// lineasDelTexto son las líneas no vacías de 20 caracteres o más del texto de
// un bloque, tal cual; los caracteres se cuentan como runas y no como bytes.
func lineasDelTexto(texto string) []string {
	var lineas []string

	for linea := range strings.SplitSeq(texto, "\n") {
		if strings.TrimSpace(linea) != "" && utf8.RuneCountInString(linea) >= minimoDeCaracteres {
			lineas = append(lineas, linea)
		}
	}

	return lineas
}

// instantaneaDelGrafo lee todo el grafo del world.db del directorio por la API
// pública de internal/graph.
func instantaneaDelGrafo(t *testing.T, directorio string) grafo.Instantanea {
	t.Helper()

	lectura, err := graph.Leer(t.Context(), graph.ConDirectorio(directorio))
	require.NoError(t, err)

	instantanea, err := lectura.Instantanea(t.Context(), grafo.Ambito{})
	require.NoError(t, err)
	require.NoError(t, lectura.Close())

	return instantanea
}

// compruebaLoEntregado es la premisa de que el grafo guarda lo que boe publicó:
// un texto por cada huella y, por cada una, la BloqueVersion que la lleva.
func compruebaLoEntregado(t *testing.T, instantanea grafo.Instantanea, textos map[string]string) {
	t.Helper()

	huellas := map[string]bool{}

	for _, nodo := range instantanea.Nodos {
		if nodo.Tipo == grafo.TipoBloqueVersion {
			huella, esTexto := nodo.Datos[grafo.DatoHashTexto].(string)
			require.True(t, esTexto, "la huella de %q es un texto", nodo.ID)

			huellas[huella] = true
		}
	}

	assert.ElementsMatch(t, slices.Collect(maps.Keys(textos)), slices.Collect(maps.Keys(huellas)),
		"premisa: una BloqueVersion por cada texto que publicó boe")
}

// lineasQueAparecen son las líneas que aparecen en la salida: en ella tal cual
// y, con --json, en cada clave y cada valor de texto del documento. Se comparan
// con cada tramo de espacio en blanco, en la línea y en la salida, reducido a
// un espacio: la tabla alinea con espacios lo que escribe, y un tabulador del
// texto sale en ella como relleno, así que una línea que lo lleve no aparecería
// literalmente aunque su contenido estuviera ahí.
func lineasQueAparecen(t *testing.T, lineas []string, salida string, conJSON bool) []string {
	t.Helper()

	textos := []string{blancosReducidos(salida)}

	if conJSON {
		var documento any
		require.NoError(t, json.Unmarshal([]byte(salida), &documento))

		for _, texto := range textosDelDocumento(documento) {
			textos = append(textos, blancosReducidos(texto))
		}
	}

	var aparecen []string

	for _, linea := range lineas {
		buscada := blancosReducidos(linea)

		if slices.ContainsFunc(textos, func(texto string) bool { return strings.Contains(texto, buscada) }) {
			aparecen = append(aparecen, linea)
		}
	}

	return aparecen
}

// blancosReducidos es el texto con cada tramo de espacio en blanco —en el
// sentido de unicode.IsSpace— reducido a un espacio y sin los de los extremos.
func blancosReducidos(texto string) string {
	return strings.Join(strings.Fields(texto), " ")
}

// textosDelDocumento son las claves y los valores de texto de un documento JSON
// a cualquier profundidad.
func textosDelDocumento(valor any) []string {
	var textos []string

	switch v := valor.(type) {
	case string:
		textos = append(textos, v)
	case []any:
		for _, elemento := range v {
			textos = append(textos, textosDelDocumento(elemento)...)
		}
	case map[string]any:
		for clave, elemento := range v {
			textos = append(textos, clave)
			textos = append(textos, textosDelDocumento(elemento)...)
		}
	}

	return textos
}

// Las horas de la reproducción de los tests de la entrega: la que declara cada
// petición que sirve la reproducción, fijada con httpx.ConHora. Llevan fracción
// de segundo y un desplazamiento que no es UTC, de modo que una fecha del lote
// escrita de otra forma que la del sobre —sin la fracción, en UTC— no pasaría
// por igual; y van en ese orden: la temprana, la tardía y la última.
var (
	desplazamientoDePrueba = time.FixedZone("", 2*60*60)
	horaTemprana           = time.Date(2026, time.September, 28, 10, 0, 0, 250_000_000, desplazamientoDePrueba)
	horaTardia             = time.Date(2026, time.September, 28, 11, 0, 0, 500_000_000, desplazamientoDePrueba)
	horaUltima             = time.Date(2026, time.September, 28, 12, 0, 0, 750_000_000, desplazamientoDePrueba)
)

// bancoConLaHora es el banco de la carpeta con la hora de emisión fijada en el
// instante: toda petición que sirve la reproducción lo declara, así que dos
// bancos nuevos —cada uno con su caché vacía— piden lo mismo a la fuente y
// fechan igual su sobre.
func bancoConLaHora(t *testing.T, carpeta string, instante time.Time) *bancoDeBoe {
	t.Helper()

	banco := nuevoBancoDeBoe(t, carpeta)
	banco.dependencias.Cliente = func(*slog.Logger) (*httpx.Cliente, error) {
		banco.construcciones.Add(1)

		return httpx.Replay(carpeta,
			httpx.ConFuente(boe.NombreDeLaFuente),
			httpx.ConRegistrador(banco.eventos.registrador()),
			httpx.ConHora(func() time.Time { return instante }))
	}

	return banco
}

// TestLaSalidaDeBoeNoCambiaConElGrafo fija SC-003, SC-004, FR-031, FR-034 y
// FR-042 por el kernel en proceso: para cada respuesta grabada de boe articulo
// —las de H4, también la del bloque que la fuente no tiene, y las de H5 y
// H5.1—, con --json y sin ella, la salida estándar y el código son los mismos
// byte a byte con --no-graph que sin ella; y con --no-graph y con --dry-run,
// world.db y sus auxiliares quedan con la misma huella o siguen sin existir.
// Cada invocación sale de un banco nuevo, con su caché vacía y la hora de la
// reproducción fijada, de modo que todas piden lo mismo a la fuente y fechan
// igual su sobre; y el almacén vive en un directorio temporal. La salida de
// error no se compara entera, porque lleva la hora y la duración del registro
// de eventos de un fallo; de ella se exige que la entrega no deje ningún aviso.
//
// Las premisas dicen que no pasa en vacío: sin --no-graph, un artículo que sale
// con 0 llega a world.db; y el ensayo sobre la caché que ya lo guarda lo sirve
// de su entrada sin pedir nada, así que el applet devuelve lo que observa y es
// el kernel el que no lo entrega.
func TestLaSalidaDeBoeNoCambiaConElGrafo(t *testing.T) {
	t.Parallel()

	for _, grabaciones := range []struct{ nombre, carpeta string }{
		{nombre: "h4", carpeta: grabacionesDeBoe},
		{nombre: "evals", carpeta: grabacionesDeEvals},
	} {
		respuestas := respuestasDeBloques(t, grabaciones.carpeta)
		require.NotEmpty(t, respuestas, "premisa: %s tiene respuestas de bloques", grabaciones.carpeta)

		for _, grabado := range respuestas {
			for _, banderas := range [][]string{{"--json"}, nil} {
				argumentos := slices.Concat([]string{"articulo", grabado.norma, grabado.bloque}, banderas)

				t.Run(strings.Join(slices.Concat([]string{grabaciones.nombre}, argumentos), "/"), func(t *testing.T) {
					t.Parallel()

					banco, directorio := compruebaLaMismaSalida(t, grabaciones.carpeta, grabado, argumentos)
					compruebaQueElEnsayoNoEntrega(t, banco, directorio, grabado, argumentos)
				})
			}
		}
	}
}

// compruebaLaMismaSalida invoca boe con los argumentos tres veces, cada una
// sobre un banco nuevo de la carpeta y con la entrega al world.db de un mismo
// directorio temporal: con --no-graph y el directorio vacío, que sigue vacío;
// sin ella, con la misma salida estándar y el mismo código, sin ningún aviso de
// la entrega y, si el bloque se sirvió con 200, con lo observado en world.db; y
// otra vez con --no-graph, con lo mismo y sin cambiar un byte del directorio.
// Devuelve el banco de la invocación que entregó, cuya caché guarda ya lo que
// leyó, y el directorio.
func compruebaLaMismaSalida(
	t *testing.T, carpeta string, grabado bloqueGrabado, argumentos []string,
) (*bancoDeBoe, string) {
	t.Helper()

	directorio := t.TempDir()
	vacio := huellasDelDirectorio(t, directorio)
	sinGrafo := argvDeBoe(slices.Concat(argumentos, []string{"--no-graph"})...)
	invocarConUnBancoNuevo := func(argv []string) (*bancoDeBoe, invocacionDePrueba) {
		banco := bancoConLaHora(t, carpeta, horaTardia)

		return banco, invocar(t, registroConLaEntrega(t, directorio, AppletBoe(banco.dependencias)), argv...)
	}

	_, descartado := invocarConUnBancoNuevo(sinGrafo)
	assert.Equal(t, vacio, huellasDelDirectorio(t, directorio), "con --no-graph, world.db sigue sin existir")

	banco, entregado := invocarConUnBancoNuevo(argvDeBoe(argumentos...))
	compruebaLaMismaSalidaYElMismoCodigo(t, descartado, entregado)
	assert.NotContains(t, entregado.errores, avisoDeEntregaFallida, "la entrega no deja ningún aviso")

	if grabado.estado == 200 {
		require.Equal(t, 0, entregado.codigo, entregado.errores)
		require.FileExists(t, filepath.Join(directorio, "world.db"), "premisa: lo observado llega al grafo")
	} else {
		require.NotEqual(t, 0, entregado.codigo, "premisa: el bloque que la fuente no tiene no se lee")
	}

	poblado := huellasDelDirectorio(t, directorio)
	_, otraVez := invocarConUnBancoNuevo(sinGrafo)
	compruebaLaMismaSalidaYElMismoCodigo(t, entregado, otraVez)
	assert.Equal(t, poblado, huellasDelDirectorio(t, directorio), "con --no-graph, world.db y sus auxiliares no cambian")

	return banco, directorio
}

// compruebaLaMismaSalidaYElMismoCodigo exige que dos invocaciones den la misma
// salida estándar, byte a byte, y el mismo código.
func compruebaLaMismaSalidaYElMismoCodigo(t *testing.T, una, otra invocacionDePrueba) {
	t.Helper()

	assert.Equal(t, una.salida, otra.salida, "la misma salida estándar byte a byte")
	assert.Equal(t, una.codigo, otra.codigo, "el mismo código")
}

// compruebaQueElEnsayoNoEntrega invoca boe con los argumentos y --dry-run sobre
// el banco que ya leyó el bloque, con la entrega al world.db del directorio y
// con la de un directorio vacío: ninguno cambia. Si el bloque se sirvió con 200,
// su artículo está en la caché del banco y el ensayo lo sirve de su entrada, sin
// pedir nada ni describir ninguna petición: el applet devuelve lo que observa y
// es el kernel el que no lo entrega (FR-034).
func compruebaQueElEnsayoNoEntrega(
	t *testing.T, banco *bancoDeBoe, directorio string, grabado bloqueGrabado, argumentos []string,
) {
	t.Helper()

	ensayo := argvDeBoe(slices.Concat(argumentos, []string{"--dry-run"})...)

	for _, destino := range []string{directorio, t.TempDir()} {
		antes := huellasDelDirectorio(t, destino)
		pedidas := len(banco.eventos.pedidas(t))

		res := invocar(t, registroConLaEntrega(t, destino, AppletBoe(banco.dependencias)), ensayo...)
		assert.Equal(t, antes, huellasDelDirectorio(t, destino),
			"con --dry-run, world.db y sus auxiliares no cambian o siguen sin existir")

		if grabado.estado == 200 {
			require.Equal(t, 0, res.codigo, res.errores)
			assert.Len(t, banco.eventos.pedidas(t), pedidas, "premisa: el ensayo sirve el artículo de su entrada")
			assert.NotContains(t, res.errores, prefijoDeEnsayo+"se habr\xc3\xada pedido",
				"premisa: el ensayo no describe ninguna petición")
		}
	}
}

// claseDeTexto es la de un texto en el estado del grafo; la de un nodo lleva su
// tipo y la de una arista, su relación.
const claseDeTexto = "texto"

// claseDeNodo es la de un nodo del tipo en el estado del grafo.
func claseDeNodo(tipo string) string {
	return "nodo " + tipo
}

// claseDeArista es la de una arista de la relación en el estado del grafo.
func claseDeArista(relacion string) string {
	return "arista " + relacion
}

// guardado es lo que el grafo guarda de un nodo, de una arista o de un texto:
// su clase; la procedencia de su última observación, que en un texto es la de
// su observación más antigua; la fecha de su primera observación, que un texto
// no tiene; y los datos de un nodo, en JSON.
type guardado struct {
	clase       string
	procedencia grafo.Procedencia
	primera     string
	datos       string
}

// cambiosDelGrafo son las clases de lo que una invocación crea en el grafo, de
// lo que ya estaba y cambia de última observación —o de procedencia, en un
// texto— y de lo que ya estaba y cambia de primera observación.
type cambiosDelGrafo struct {
	creados      []string
	procedencias []string
	primeras     []string
}

// pasoDeLaEntrega es una invocación de TestLaEntregaLlevaLaProcedenciaDelSobre:
// su nombre, el applet que la atiende, sus argumentos, la hora de la
// reproducción con la que se fecha su sobre —ninguna fuera de boe— y lo que
// cambia en el grafo.
type pasoDeLaEntrega struct {
	nombre  string
	applet  func(t *testing.T) Applet
	argv    []string
	hora    time.Time
	cambios cambiosDelGrafo
}

// TestLaEntregaLlevaLaProcedenciaDelSobre fija FR-021, FR-089 y SC-002 por el
// kernel en proceso, sobre el world.db de un directorio temporal con la muestra
// ya entregada: boe articulo, boe articulos y territorio resolver, invocados uno
// tras otro con la entrega registrada, crean o actualizan cada nodo, arista y
// texto con la fuente, la url y la fecha de consulta del sobre que devolvió la
// misma invocación, byte a byte; lo que no tocan conserva lo suyo; y nada queda
// en el grafo sin fuente, url o fecha. Los nodos y las aristas se leen por la
// API pública de internal/graph; los textos, del fichero (textosGuardados).
//
// El orden de las invocaciones ejerce cada caso de FR-023 y fija lo que cambia
// cada una: la primera crea; la segunda, con una hora anterior, crea lo de su
// otro bloque, adelanta la primera observación de lo que ya había observado la
// primera sin tocar su última y da al texto que ya estaba la procedencia de la
// observación más antigua; la tercera, posterior a todas, cambia la última
// observación de lo que ya estaba de su bloque y de la Norma y deja el texto
// con la suya; territorio resolver crea lo suyo sin tocar nada de boe; y su
// repetición no cambia nada.
func TestLaEntregaLlevaLaProcedenciaDelSobre(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	poblarElGrafo(t, directorio)

	antes := estadoDelGrafo(t, directorio)

	for _, paso := range pasosDeLaEntrega(t) {
		res := invocar(t, registroConLaEntrega(t, directorio, paso.applet(t)), paso.argv...)
		require.Equal(t, 0, res.codigo, "%s: %s", paso.nombre, res.errores)
		require.Empty(t, res.errores, "%s: lo observado llega al grafo", paso.nombre)

		sobre := procedenciaDelSobre(t, res.salida)
		if !paso.hora.IsZero() {
			require.Equal(t, paso.hora.Format(time.RFC3339Nano), sobre.FechaConsulta,
				"%s: premisa: el sobre lleva la hora de la reproducción", paso.nombre)
		}

		despues := estadoDelGrafo(t, directorio)
		cambios := cambiosDeLaInvocacion(t, paso.nombre, antes, despues, sobre)

		assert.ElementsMatch(t, paso.cambios.creados, cambios.creados, "%s: lo que crea", paso.nombre)
		assert.ElementsMatch(t, paso.cambios.procedencias, cambios.procedencias,
			"%s: lo que cambia de última observación", paso.nombre)
		assert.ElementsMatch(t, paso.cambios.primeras, cambios.primeras,
			"%s: lo que cambia de primera observación", paso.nombre)

		antes = despues
	}
}

// pasosDeLaEntrega son las invocaciones de
// TestLaEntregaLlevaLaProcedenciaDelSobre en su orden, con lo que cambia cada
// una por clase: boe articulo observa de un bloque la Norma, el Bloque, su
// BloqueVersion, sus dos aristas y el texto (contracts/emision.md §1), y
// territorio resolver, el Municipio, el Organo de su DIR3 y la arista entre
// ellos (§2), sobre los datos de territorio del binario.
func pasosDeLaEntrega(t *testing.T) []pasoDeLaEntrega {
	t.Helper()

	fuentes, err := FuentesEmbebidas()
	require.NoError(t, err)

	territorio := AppletTerritorio(fuentes)
	deTerritorio := func(*testing.T) Applet { return territorio }
	deBoe := func(hora time.Time) func(*testing.T) Applet {
		return func(t *testing.T) Applet {
			t.Helper()

			return AppletBoe(bancoConLaHora(t, grabacionesDeBoe, hora).dependencias)
		}
	}

	delBloque := []string{
		claseDeNodo(grafo.TipoBloque), claseDeNodo(grafo.TipoBloqueVersion),
		claseDeArista(grafo.RelacionTieneParte), claseDeArista(grafo.RelacionTieneVersion),
	}
	norma, resolver := claseDeNodo(grafo.TipoNorma), []string{"kitlegal", "territorio", "resolver", "28074", "--json"}

	return []pasoDeLaEntrega{
		{
			nombre: "articulo", applet: deBoe(horaTardia), hora: horaTardia,
			argv:    argvDeBoe("articulo", normaDeBoe, "a21", "--json"),
			cambios: cambiosDelGrafo{creados: slices.Concat(delBloque, []string{norma, claseDeTexto})},
		},
		{
			nombre: "articulos anteriores", applet: deBoe(horaTemprana), hora: horaTemprana,
			argv: argvDeBoe("articulos", normaDeBoe, "a22", "a21", "a22", "--json"),
			cambios: cambiosDelGrafo{
				creados:      slices.Concat(delBloque, []string{claseDeTexto}),
				procedencias: []string{claseDeTexto},
				primeras:     slices.Concat(delBloque, []string{norma}),
			},
		},
		{
			nombre: "articulo posterior", applet: deBoe(horaUltima), hora: horaUltima,
			argv:    argvDeBoe("articulo", normaDeBoe, "a22", "--json"),
			cambios: cambiosDelGrafo{procedencias: slices.Concat(delBloque, []string{norma})},
		},
		{
			nombre: "resolver", applet: deTerritorio, argv: resolver,
			cambios: cambiosDelGrafo{creados: []string{
				claseDeNodo(grafo.TipoMunicipio), claseDeNodo(grafo.TipoOrgano), claseDeArista(grafo.RelacionPerteneceA),
			}},
		},
		{nombre: "resolver de nuevo", applet: deTerritorio, argv: resolver},
	}
}

// procedenciaDelSobre es la fuente, la url y la fecha de consulta del sobre
// correcto de una invocación con --json.
func procedenciaDelSobre(t *testing.T, salida string) grafo.Procedencia {
	t.Helper()

	sobre := sobreDelJSON(t, salida)
	require.Equal(t, true, sobre["ok"])

	texto := func(clave string) string {
		valor, esTexto := sobre[clave].(string)
		require.True(t, esTexto, "%s es un texto", clave)

		return valor
	}

	return grafo.Procedencia{Fuente: texto("fuente"), URL: texto("url"), FechaConsulta: texto("fecha_consulta")}
}

// cambiosDeLaInvocacion compara lo que el grafo guardaba antes de una
// invocación con lo que guarda después y exige FR-089: nada desaparece ni
// cambia de clase; lo que la invocación crea lleva como procedencia la del
// sobre que devolvió y, si no es un texto, su fecha de consulta como primera
// observación; lo que ya estaba y cambia de procedencia o de datos pasa a la
// procedencia de ese sobre, y lo que cambia de primera observación, a su fecha;
// lo demás conserva lo suyo; y nada queda sin fuente, url o fecha. Devuelve las
// clases de lo que creó y de lo que cambió.
func cambiosDeLaInvocacion(
	t *testing.T, paso string, antes, despues map[string]guardado, sobre grafo.Procedencia,
) cambiosDelGrafo {
	t.Helper()

	for clave := range antes {
		assert.Contains(t, despues, clave, "%s: nada desaparece del grafo", paso)
	}

	var cambios cambiosDelGrafo

	for clave, ahora := range despues {
		compruebaConFuente(t, paso, clave, ahora)

		previo, estaba := antes[clave]

		switch {
		case !estaba:
			cambios.creados = append(cambios.creados, ahora.clase)
			assert.Equal(t, sobre, ahora.procedencia, "%s: %q se crea con la procedencia del sobre", paso, clave)

			if ahora.clase != claseDeTexto {
				assert.Equal(t, sobre.FechaConsulta, ahora.primera, "%s: %q se crea con la fecha del sobre", paso, clave)
			}
		default:
			assert.Equal(t, previo.clase, ahora.clase, "%s: %q no cambia de clase", paso, clave)

			if ahora.procedencia != previo.procedencia || ahora.datos != previo.datos {
				cambios.procedencias = append(cambios.procedencias, ahora.clase)
				assert.Equal(t, sobre, ahora.procedencia, "%s: %q cambia a la procedencia del sobre", paso, clave)
			}

			if ahora.primera != previo.primera {
				cambios.primeras = append(cambios.primeras, ahora.clase)
				assert.Equal(t, sobre.FechaConsulta, ahora.primera, "%s: %q cambia a la fecha del sobre", paso, clave)
			}
		}
	}

	return cambios
}

// compruebaConFuente exige que lo guardado tenga fuente, url y fecha de
// consulta y, si no es un texto, fecha de primera observación: ninguna
// operación entra en el grafo sin fuente (FR-024).
func compruebaConFuente(t *testing.T, paso, clave string, ahora guardado) {
	t.Helper()

	assert.NotEmpty(t, ahora.procedencia.Fuente, "%s: %q tiene fuente", paso, clave)
	assert.NotEmpty(t, ahora.procedencia.URL, "%s: %q tiene url", paso, clave)
	assert.NotEmpty(t, ahora.procedencia.FechaConsulta, "%s: %q tiene fecha de consulta", paso, clave)

	if ahora.clase != claseDeTexto {
		assert.NotEmpty(t, ahora.primera, "%s: %q tiene primera observación", paso, clave)
	}
}

// estadoDelGrafo es todo lo que guarda el world.db del directorio, por la clave
// de cada elemento: cada nodo y cada arista, con su primera y su última
// observación, leídos por la API pública de internal/graph —la ficha de cada
// nodo de la instantánea, con sus aristas salientes—, y cada texto con su
// procedencia, que ninguna lectura de internal/graph devuelve.
func estadoDelGrafo(t *testing.T, directorio string) map[string]guardado {
	t.Helper()

	lectura, err := graph.Leer(t.Context(), graph.ConDirectorio(directorio))
	require.NoError(t, err)

	instantanea, err := lectura.Instantanea(t.Context(), grafo.Ambito{})
	require.NoError(t, err)

	estado := map[string]guardado{}
	aristas := 0

	for _, nodo := range instantanea.Nodos {
		ficha, esta, err := lectura.Ficha(t.Context(), nodo.ID)
		require.NoError(t, err)
		require.True(t, esta, "la ficha de %q", nodo.ID)

		datos, err := json.Marshal(ficha.Nodo.Datos)
		require.NoError(t, err)

		estado["nodo\x00"+nodo.ID] = guardado{
			clase:       claseDeNodo(ficha.Nodo.Tipo),
			procedencia: ficha.Nodo.UltimaObservacion,
			primera:     ficha.Nodo.PrimeraObservacion,
			datos:       string(datos),
		}

		for _, arista := range ficha.Salientes {
			estado["arista\x00"+nodo.ID+"\x00"+arista.Relacion+"\x00"+arista.ID] = guardado{
				clase:       claseDeArista(arista.Relacion),
				procedencia: arista.UltimaObservacion,
				primera:     arista.PrimeraObservacion,
			}
			aristas++
		}
	}

	require.NoError(t, lectura.Close())
	require.Len(t, instantanea.Aristas, aristas, "premisa: cada arista sale de la ficha de su origen")

	for huella, procedencia := range textosGuardados(t, directorio) {
		estado["texto\x00"+huella] = guardado{clase: claseDeTexto, procedencia: procedencia}
	}

	return estado
}

// textosGuardados son los textos del world.db del directorio, por su huella,
// con su procedencia, leídos del fichero con filasDeLaTablaSQLite: la tabla
// texts de la migración 0001, con la huella, el cuerpo, la fecha de consulta,
// la fuente y la url. Las premisas son que world.db está entero en su fichero
// —sin auxiliares, que ninguna entrega ni lectura deja al cerrar— y que cada
// cuerpo leído tiene la huella con la que se guardó, lo que dice que el árbol y
// sus páginas de desbordamiento se han leído bien.
func textosGuardados(t *testing.T, directorio string) map[string]grafo.Procedencia {
	t.Helper()

	for _, auxiliar := range []string{"world.db-wal", "world.db-shm", "world.db-journal"} {
		require.NoFileExists(t, filepath.Join(directorio, auxiliar), "premisa: world.db está entero en su fichero")
	}

	base, err := fs.ReadFile(os.DirFS(directorio), "world.db")
	require.NoError(t, err)

	textos := map[string]grafo.Procedencia{}

	for _, fila := range filasDeLaTablaSQLite(t, base, "texts") {
		require.Len(t, fila, 5, "la huella, el cuerpo, la fecha de consulta, la fuente y la url")

		valores := make([]string, 0, len(fila))

		for _, valor := range fila {
			texto, esTexto := valor.(string)
			require.True(t, esTexto, "cada valor de una fila de texts es un texto: %#v", valor)

			valores = append(valores, texto)
		}

		huella := valores[0]
		require.Equal(t, huellaDelTexto(valores[1]), huella, "premisa: el cuerpo leído es el que se guardó")

		textos[huella] = grafo.Procedencia{Fuente: valores[3], URL: valores[4], FechaConsulta: valores[2]}
	}

	return textos
}

// escribirFicheroDePrueba escribe un fichero de la prueba con acceso reservado a la
// cuenta.
func escribirFicheroDePrueba(t *testing.T, ruta string, contenido []byte) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Clean(ruta), contenido, 0o600))
}

// derivadasDelE2E es la carpeta de las grabaciones derivadas de las de H4 que
// el arnés copia en el $WORK/derivadas/ de cada guion, relativa a este paquete:
// una subcarpeta por caso, con una grabación que lleva el nombre de la de H4 que
// sustituye (contracts/arnes-e2e.md §3; research.md D22).
const derivadasDelE2E = "testdata/derivadas"

// grafosPreviosDeLasEvals es la carpeta de las grabaciones derivadas de las de
// H4 con las que el job prepara el grafo previo de una eval, relativa a este
// paquete: una subcarpeta por grafo previo (contracts/evals-y-skill.md §3 y §4;
// research.md D22). Es la de evals.GrafosPrevios, que resuelve igual desde
// internal/app, escrita aquí porque internal/evals importa este paquete y no se
// puede importar desde sus tests.
const grafosPreviosDeLasEvals = "../../testdata/evals/grafo-previo"

// evalsRetiradas es la carpeta de las evals que salieron del conjunto y de las
// que siguen dependiendo casos etiquetados del juez, relativa a este paquete: el
// fichero de cada eval y, en grafo-previo/, las grabaciones derivadas de su
// grafo previo, restaurados de la historia del repositorio
// (contracts/medida-del-juez.md §4 de H24; research.md D15 de H24). No es la
// carpeta de un conjunto: ningún plan la lee.
const evalsRetiradas = "../../testdata/evals/retiradas"

// evalRetiradaDeLaConsultaRepetida es el fichero de la eval de la consulta
// repetida sobre el art. 21 de la Ley 39/2015, la que H7.2 retiró, en
// evalsRetiradas.
const evalRetiradaDeLaConsultaRepetida = "19-lpac-articulo-21-redaccion-cambiada.yaml"

// Los nombres de las grabaciones de H4 que sustituyen las derivadas, los que les
// da la dirección de la que salen: la del bloque a21 y la de los metadatos de la
// Ley 39/2015.
const (
	grabacionDelBloqueA21   = "GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json"
	grabacionDeLosMetadatos = "GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_metadatos.json"
)

// grabacionDelArticulo118 es la grabación de H4 del bloque a1-30 de la Ley
// 9/2017 (LCSP), su art. 118, con dos redacciones: la original, de vigencia
// 20180309, y la vigente, de 20200206 (research.md V15 de H7.2).
const grabacionDelArticulo118 = "GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_" +
	"BOE-A-2017-12902_texto_bloque_a1-30.json"

// grabacionDeLaDisposicionAdicionalTercera es la grabación de H4 del bloque da-3
// de la LCSP, su disposición adicional tercera, con dos redacciones: la
// original, de vigencia 20180309, y la vigente, de 20230101 (research.md V15 de
// H7.4).
const grabacionDeLaDisposicionAdicionalTercera = "GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_" +
	"BOE-A-2017-12902_texto_bloque_da-3.json"

// fechaDeLaRedaccionOriginalDeLaLCSP es la fecha de vigencia de la redacción
// original de los bloques de la LCSP con la que se prepara el grafo previo de
// las evals de la consulta repetida.
const fechaDeLaRedaccionOriginalDeLaLCSP = "20180309"

// actualizarDerivadas es la bandera con la que TestGrabacionesDerivadas
// escribe, antes de comprobarlas, las derivadas del grafo previo desde su
// grabación con derivacionDelGrafoPrevio, como -actualizar-esquemas los
// esquemas publicados. Solo la usa la tarea [datos] que entrega una derivada,
// con la orden de contracts/eval-y-derivada.md §2 de H7.2; make ci nunca la
// pasa, y sin ella el test solo comprueba (FR-010).
var actualizarDerivadas = flag.Bool("actualizar-derivadas", false,
	"escribe en testdata/evals/grafo-previo/ cada derivada del grafo previo desde su grabación antes de comprobarlas")

// parrafoDeLaVersionPosterior es el párrafo que marca como sintética la
// redacción de la derivada version-posterior: el último de su versión.
const parrafoDeLaVersionPosterior = "[Redacci\xc3\xb3n sint\xc3\xa9tica de prueba: versi\xc3\xb3n posterior " +
	"derivada de la grabaci\xc3\xb3n de H4.]"

// parrafoDeLaVersionUlterior es el que marca como sintética la redacción de la
// derivada version-ulterior, la redacción C del e2e (research.md D23 de H7.1): el
// último de su versión.
const parrafoDeLaVersionUlterior = "[Redacci\xc3\xb3n sint\xc3\xa9tica de prueba: versi\xc3\xb3n ulterior " +
	"derivada de la grabaci\xc3\xb3n de H4.]"

// parrafoDeLaVersionAnterior es el que marca como sintética la redacción de la
// derivada lpac-a21-version-anterior, el grafo previo de la eval de la consulta
// repetida que H7.2 retiró: el último de su versión.
const parrafoDeLaVersionAnterior = "[Redacci\xc3\xb3n sint\xc3\xa9tica de prueba: versi\xc3\xb3n anterior " +
	"derivada de la grabaci\xc3\xb3n de H4.]"

// grabacionDerivada es una derivada con lo que dice su nombre: la carpeta en la
// que está, el nombre de la grabación de H4 que sustituye y la comprobación de
// que, leída con boe, solo cambia eso.
type grabacionDerivada struct {
	carpeta, fichero string
	// comprueba recibe la carpeta de reproducción con las grabaciones de H4 y
	// otra igual con la derivada en lugar de la suya.
	comprueba func(t *testing.T, original, derivada string)
}

// grabacionesDerivadas son las derivadas del e2e (research.md D22), cada una con
// lo que dice su nombre: version-posterior, la fecha de vigencia 20250101 y el
// párrafo sintético al final del texto, con la huella de ese texto;
// version-ulterior, lo mismo con la fecha 20260101 y su párrafo; sin-eli, la
// url_eli vacía; y eli-sin-segmento, una url_eli sin el segmento eli. Las del
// grafo previo de las evals no caben aquí: son las de derivadasDelGrafoPrevio.
func grabacionesDerivadas() []grabacionDerivada {
	return []grabacionDerivada{
		versionDelArticulo21(filepath.Join(derivadasDelE2E, "version-posterior"), "20250101",
			parrafoDeLaVersionPosterior),
		versionDelArticulo21(filepath.Join(derivadasDelE2E, "version-ulterior"), "20260101",
			parrafoDeLaVersionUlterior),
		{
			carpeta: filepath.Join(derivadasDelE2E, "sin-eli"),
			fichero: grabacionDeLosMetadatos,
			comprueba: func(t *testing.T, original, derivada string) {
				t.Helper()

				compruebaLaDerivacion(t, original, derivada, []string{"metadatos", normaDeBoe},
					func(metadatos *boe.Metadatos) { metadatos.URLELI = "" })
			},
		},
		{
			carpeta: filepath.Join(derivadasDelE2E, "eli-sin-segmento"),
			fichero: grabacionDeLosMetadatos,
			comprueba: func(t *testing.T, original, derivada string) {
				t.Helper()

				compruebaLaDerivacion(t, original, derivada, []string{"metadatos", normaDeBoe},
					func(metadatos *boe.Metadatos) {
						metadatos.URLELI = "https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565"
					})
			},
		},
	}
}

// versionDelArticulo21 es la derivada de la carpeta que sustituye la grabación
// del bloque a21 de la Ley 39/2015 con otra versión: la fecha de vigencia y el
// párrafo sintético al final del texto, con la huella de ese texto.
func versionDelArticulo21(carpeta, fechaVigencia, parrafo string) grabacionDerivada {
	return grabacionDerivada{
		carpeta: carpeta,
		fichero: grabacionDelBloqueA21,
		comprueba: func(t *testing.T, original, derivada string) {
			t.Helper()

			compruebaLaDerivacion(t, original, derivada, []string{"articulo", normaDeBoe, "a21"},
				func(articulo *boe.Articulo) {
					articulo.FechaVigencia = fechaVigencia
					articulo.Texto += "\n" + parrafo
					articulo.HashTexto = huellaDelTexto(articulo.Texto)
				})
		},
	}
}

// derivadasDeLasEvalsRetiradas son las derivadas del grafo previo de las evals
// de evalsRetiradas, con lo que dice su nombre: lpac-a21-version-anterior, la
// fecha de vigencia 20151002 y su párrafo sintético al final del texto, con la
// huella de ese texto. Es la entrada que tenía en grabacionesDerivadas antes de
// que H7.2 retirara su eval, sobre la carpeta a la que vuelve (research.md D15
// de H24; FR-024): una redacción sintética, y no una que la grabada traiga, que
// es lo que compruebaLaDerivadaDelGrafoPrevio exige a las de las evals de hoy.
func derivadasDeLasEvalsRetiradas() []grabacionDerivada {
	return []grabacionDerivada{
		versionDelArticulo21(filepath.Join(evalsRetiradas, "grafo-previo", "lpac-a21-version-anterior"), "20151002",
			parrafoDeLaVersionAnterior),
	}
}

// rutasDeLasDerivadas son las rutas de los ficheros de las derivadas, cada una
// con su carpeta delante, relativas a este paquete.
func rutasDeLasDerivadas(derivadas []grabacionDerivada) []string {
	rutas := make([]string, 0, len(derivadas))
	for _, derivada := range derivadas {
		rutas = append(rutas, filepath.Join(derivada.carpeta, derivada.fichero))
	}

	return rutas
}

// derivadaDelGrafoPrevio es una derivada de grafosPreviosDeLasEvals: su
// subcarpeta, el nombre de la grabación de H4 de la que sale y que sustituye,
// los argumentos de boe con los que se lee y la fecha de vigencia de la
// redacción que da. No lleva comprobación propia: TestGrabacionesDerivadas
// aplica a todas la reproducción byte a byte y la de la redacción de la grabada
// (contracts/eval-y-derivada.md §3 de H7.2).
type derivadaDelGrafoPrevio struct {
	subcarpeta, fichero string
	argumentos          []string
	fechaVigencia       string
}

// carpeta es la subcarpeta de la derivada, relativa a este paquete.
func (d derivadaDelGrafoPrevio) carpeta() string {
	return filepath.Join(grafosPreviosDeLasEvals, d.subcarpeta)
}

// subprueba es el nombre de la subprueba de la derivada en
// TestGrabacionesDerivadas: su subcarpeta y el bloque que lee, que la distingue
// de las demás de la misma subcarpeta.
func (d derivadaDelGrafoPrevio) subprueba() string {
	return d.subcarpeta + "/" + d.argumentos[len(d.argumentos)-1]
}

// derivadasDelGrafoPrevio son las derivadas del grafo previo de las evals
// (research.md D13 de H7.2), aparte de las del e2e de grabacionesDerivadas: la
// de la eval de la consulta repetida y, desde H7.4, las dos de la eval de los
// dos bloques de la LCSP, cada una con la grabación de H4 de su bloque
// (research.md D11 de H7.4; FR-051).
func derivadasDelGrafoPrevio() []derivadaDelGrafoPrevio {
	const dosBloques = "lcsp-a1-30-y-da-3-redaccion-original"

	return []derivadaDelGrafoPrevio{
		redaccionOriginalDelArticulo118(),
		redaccionOriginalDeLaLCSP(dosBloques, grabacionDelArticulo118, "a1-30"),
		redaccionOriginalDeLaLCSP(dosBloques, grabacionDeLaDisposicionAdicionalTercera, "da-3"),
	}
}

// redaccionOriginalDelArticulo118 es el grafo previo de la eval de la consulta
// repetida: la grabación del art. 118 de la LCSP sin su redacción vigente, que
// da la original, la de vigencia 20180309 (FR-010).
func redaccionOriginalDelArticulo118() derivadaDelGrafoPrevio {
	return redaccionOriginalDeLaLCSP("lcsp-a1-30-redaccion-original", grabacionDelArticulo118, "a1-30")
}

// redaccionOriginalDeLaLCSP es la derivada de la subcarpeta que sustituye la
// grabación del bloque de la LCSP con la misma sin sus redacciones posteriores
// a la original, la de vigencia 20180309, que es la que da.
func redaccionOriginalDeLaLCSP(subcarpeta, fichero, bloque string) derivadaDelGrafoPrevio {
	return derivadaDelGrafoPrevio{
		subcarpeta:    subcarpeta,
		fichero:       fichero,
		argumentos:    []string{"articulo", "BOE-A-2017-12902", bloque},
		fechaVigencia: fechaDeLaRedaccionOriginalDeLaLCSP,
	}
}

// TestGrabacionesDerivadas es el control de derivación de research.md D22 (FR-090,
// FR-095) y de research.md D13 de H7.2. Cada derivada del e2e lleva el nombre de
// una grabación de H4 y, servida en su lugar, boe la lee y da el mismo Articulo
// o los mismos metadatos que la grabación salvo exactamente lo que dice su
// nombre (grabacionesDerivadas); la premisa de cada una dice que no pasa en
// vacío: lo que dice su nombre cambia algo de lo que da la grabación. Las del
// grafo previo de las evals (derivadasDelGrafoPrevio) pasan todas por
// compruebaLaDerivadaDelGrafoPrevio: son, byte a byte, la derivación de su
// grabación, y dan una redacción que la grabada trae, la de su fecha (FR-010,
// FR-011). Con -actualizar-derivadas, el test las escribe antes desde su
// grabación. La del grafo previo de la eval retirada
// (derivadasDeLasEvalsRetiradas; research.md D15 de H24, FR-024) pasa por la
// comprobación de las del e2e: leída con boe, es la grabación de H4 salvo
// exactamente lo que dice su nombre.
//
// La carpeta decide la comprobación (FR-013): los ficheros de derivadasDelE2E
// son los de las entradas del e2e; los de grafosPreviosDeLasEvals, los de las
// del grafo previo; y los de evalsRetiradas, la derivada restaurada y el fichero
// de su eval, que no lleva comprobación —nada la valida contra el esquema de
// eval de hoy—, en los dos sentidos. Todo fichero tiene la comprobación de su
// carpeta y toda comprobación, su fichero: una derivada nueva que no dijera qué
// cambia no pasa, y una entrada del e2e con fichero en la carpeta del grafo
// previo falla nombrada, como un fichero de más o de menos en evalsRetiradas.
func TestGrabacionesDerivadas(t *testing.T) {
	t.Parallel()

	delGrafoPrevio := derivadasDelGrafoPrevio()

	if *actualizarDerivadas {
		for _, derivada := range delGrafoPrevio {
			require.NoError(t, os.MkdirAll(derivada.carpeta(), 0o750))
			escribirFicheroDePrueba(t, filepath.Join(derivada.carpeta(), derivada.fichero),
				derivacionDelGrafoPrevio(t, grabadaDeBoe(t, derivada.fichero), derivada.fechaVigencia))
		}
	}

	derivadas := grabacionesDerivadas()
	retiradas := derivadasDeLasEvalsRetiradas()

	previas := make([]string, 0, len(delGrafoPrevio))
	for _, derivada := range delGrafoPrevio {
		previas = append(previas, filepath.Join(derivada.carpeta(), derivada.fichero))
	}

	compruebaLaCarpetaConSusEntradas(t, derivadasDelE2E, "del e2e", rutasDeLasDerivadas(derivadas))
	compruebaLaCarpetaConSusEntradas(t, grafosPreviosDeLasEvals, "del grafo previo", previas)
	compruebaLaCarpetaConSusEntradas(t, evalsRetiradas, "de las evals retiradas",
		append(rutasDeLasDerivadas(retiradas), filepath.Join(evalsRetiradas, evalRetiradaDeLaConsultaRepetida)))

	for _, derivada := range slices.Concat(derivadas, retiradas) {
		t.Run(filepath.Base(derivada.carpeta), func(t *testing.T) {
			t.Parallel()

			require.FileExists(t, filepath.Join(grabacionesDeBoe, derivada.fichero),
				"la derivada lleva el nombre de una grabación de H4")

			derivada.comprueba(t, grabacionesDeBoe, reproduccionConLaDerivada(t, derivada))
		})
	}

	for _, derivada := range delGrafoPrevio {
		t.Run(derivada.subprueba(), func(t *testing.T) {
			t.Parallel()

			compruebaLaDerivadaDelGrafoPrevio(t, derivada)
		})
	}
}

// compruebaLaCarpetaConSusEntradas exige que los ficheros de la carpeta sean los
// de las entradas de su lista, en los dos sentidos: cada entrada, con su
// fichero en la carpeta, y cada fichero, con su entrada. Así la carpeta, y no la
// clase de la entrada, decide la comprobación: una entrada de otra lista con
// fichero en esta falla nombrada, como entrada sin fichero en la carpeta de su
// lista y con su fichero sin entrada en esta (contracts/eval-y-derivada.md §3
// de H7.2).
func compruebaLaCarpetaConSusEntradas(t *testing.T, carpeta, lista string, entradas []string) {
	t.Helper()

	ficheros := ficherosDeLaCarpeta(t, carpeta)

	assert.Empty(t, sinPareja(entradas, ficheros), "entradas %s sin fichero en %s", lista, carpeta)
	assert.Empty(t, sinPareja(ficheros, entradas), "ficheros de %s sin entrada %s", carpeta, lista)
}

// sinPareja son las rutas que quedan de las primeras al emparejar cada una con
// una igual de las segundas, una a una: una ruta repetida en las primeras
// necesita otras tantas en las segundas.
func sinPareja(primeras, segundas []string) []string {
	pendientes := slices.Clone(segundas)

	var sobran []string

	for _, ruta := range primeras {
		posicion := slices.Index(pendientes, ruta)
		if posicion == -1 {
			sobran = append(sobran, ruta)

			continue
		}

		pendientes = slices.Delete(pendientes, posicion, posicion+1)
	}

	return sobran
}

// compruebaLaDerivadaDelGrafoPrevio es lo que TestGrabacionesDerivadas exige a
// toda derivada del grafo previo: que lleve el nombre de una grabación de H4;
// que sea, byte a byte, la derivación de esa grabación con su fecha de
// vigencia, sin ningún otro cambio (comprobación 1, FR-010); y que, servida en
// su lugar, boe dé la redacción de esa fecha que trae la grabada y ninguna
// otra (comprobación 2, FR-011). La premisa, que la derivación con la fecha de
// la última redacción devuelve la grabación byte a byte: sin ella, la
// comparación de la 1 podría estar midiendo el formato con que se escribe y no
// lo que la derivación quita.
func compruebaLaDerivadaDelGrafoPrevio(t *testing.T, derivada derivadaDelGrafoPrevio) {
	t.Helper()

	grabada := grabadaDeBoe(t, derivada.fichero)

	redacciones := redaccionesDelCuerpo(t, leerLaGrabacion(t, grabada).Respuesta.Cuerpo)
	require.NotEmpty(t, redacciones, "premisa: la grabación trae alguna redacción del bloque")
	require.Equal(t, string(grabada),
		string(derivacionDelGrafoPrevio(t, grabada, redacciones[len(redacciones)-1].fechaVigencia)),
		"premisa: la derivación con la fecha de la última redacción devuelve la grabación byte a byte")

	escrita, err := fs.ReadFile(os.DirFS(derivada.carpeta()), derivada.fichero)
	require.NoError(t, err)
	assert.Equal(t, string(derivacionDelGrafoPrevio(t, grabada, derivada.fechaVigencia)), string(escrita),
		"%s es, byte a byte, la derivación de su grabación con la fecha de vigencia %s",
		filepath.Join(derivada.carpeta(), derivada.fichero), derivada.fechaVigencia)

	assert.NoError(t, comprobarLaRedaccionDeLaGrabada(t, derivada, derivada.carpeta()))
}

// TestGrabacionesDerivadasInventadas es el test de que la comprobación 2 rechaza
// lo inventado (FR-012, SC-005; contracts/eval-y-derivada.md §3 de H7.2): tres
// derivadas armadas en t.TempDir() desde la grabación del art. 118 de la LCSP
// —la redacción original con la fecha de vigencia 20151002, con un párrafo de
// más al final, y con las dos cosas, como la de la eval de la consulta repetida
// retirada— no pasan, cada una con un error que nombra su fichero y dice que la
// redacción que da no es ninguna de las de la grabada. La huella no hace falta
// inventarla aparte: boe la calcula desde el texto. La premisa, que la armada
// sin inventar nada pasa: lo que rechaza la comprobación es lo inventado, no
// cómo se arma.
func TestGrabacionesDerivadasInventadas(t *testing.T) {
	t.Parallel()

	derivada := redaccionOriginalDelArticulo118()

	fechaInventada := inventoEnElCuerpo{
		antes:   `fecha_vigencia="20180309"`,
		despues: `fecha_vigencia="20151002"`,
	}
	parrafoInventado := inventoEnElCuerpo{
		antes: "</p>\n      </version>",
		despues: "</p>\n        <p class=\"parrafo\">[P\xc3\xa1rrafo inventado: no lo trae ninguna redacci\xc3\xb3n " +
			"de la grabada.]</p>\n      </version>",
	}

	require.NoError(t, comprobarLaRedaccionDeLaGrabada(t, derivada, derivadaInventada(t, derivada)),
		"premisa: la derivada armada sin inventar nada pasa")

	casos := []struct {
		nombre   string
		inventos []inventoEnElCuerpo
	}{
		{nombre: "fecha-inventada", inventos: []inventoEnElCuerpo{fechaInventada}},
		{nombre: "parrafo-inventado", inventos: []inventoEnElCuerpo{parrafoInventado}},
		{nombre: "fecha-y-parrafo-inventados", inventos: []inventoEnElCuerpo{fechaInventada, parrafoInventado}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			carpeta := derivadaInventada(t, derivada, caso.inventos...)

			err := comprobarLaRedaccionDeLaGrabada(t, derivada, carpeta)
			require.ErrorContains(t, err, filepath.Join(carpeta, derivada.fichero),
				"la derivada inventada no pasa, y el error nombra su fichero")
			assert.ErrorContains(t, err, "no es ninguna de las de la grabada")
		})
	}
}

// inventoEnElCuerpo es un cambio que la grabada no trae en ninguna de sus
// redacciones: el texto del cuerpo que sustituye y el que pone en su lugar.
type inventoEnElCuerpo struct {
	antes, despues string
}

// derivadaInventada arma en una carpeta nueva de t.TempDir(), con el nombre de
// la grabación, la derivación de la grabación con la fecha de la derivada y
// los inventos aplicados a su cuerpo, y devuelve la carpeta. La premisa de cada
// invento, que el texto que sustituye está una sola vez en el cuerpo: si no
// estuviera, la derivada no inventaría nada.
func derivadaInventada(t *testing.T, derivada derivadaDelGrafoPrevio, inventos ...inventoEnElCuerpo) string {
	t.Helper()

	grabacion := leerLaGrabacion(t,
		derivacionDelGrafoPrevio(t, grabadaDeBoe(t, derivada.fichero), derivada.fechaVigencia))

	for _, invento := range inventos {
		require.Equal(t, 1, strings.Count(grabacion.Respuesta.Cuerpo, invento.antes),
			"premisa: %q está una sola vez en el cuerpo", invento.antes)

		grabacion.Respuesta.Cuerpo = strings.Replace(grabacion.Respuesta.Cuerpo, invento.antes, invento.despues, 1)
	}

	carpeta := t.TempDir()
	escribirFicheroDePrueba(t, filepath.Join(carpeta, derivada.fichero), escribirLaGrabacion(t, grabacion))

	return carpeta
}

// comprobarLaRedaccionDeLaGrabada es la comprobación 2 (FR-011): el fichero de
// la derivada en la carpeta, servido en lugar de su grabación, da con boe y sus
// argumentos un Articulo igual, campo a campo, a exactamente uno de los que da
// boe sobre la grabación reducida a cada una de sus redacciones con
// derivacionDelGrafoPrevio, y es el de la fecha de vigencia de la derivada. Lo
// que no lo cumple es un error que nombra el fichero: una fecha, un texto o una
// huella que la grabada no trae en ninguna de sus redacciones no pasan por
// ninguna de ellas (FR-012).
func comprobarLaRedaccionDeLaGrabada(t *testing.T, derivada derivadaDelGrafoPrevio, carpeta string) error {
	t.Helper()

	servida, err := fs.ReadFile(os.DirFS(carpeta), derivada.fichero)
	require.NoError(t, err)

	leida := leidoConBoe[boe.Articulo](t, reproduccionConLaGrabacion(t, derivada.fichero, servida),
		derivada.argumentos)

	grabada := grabadaDeBoe(t, derivada.fichero)

	var coinciden []string

	for _, redaccion := range redaccionesDelCuerpo(t, leerLaGrabacion(t, grabada).Respuesta.Cuerpo) {
		reducida := reproduccionConLaGrabacion(t, derivada.fichero,
			derivacionDelGrafoPrevio(t, grabada, redaccion.fechaVigencia))

		if reflect.DeepEqual(leida, leidoConBoe[boe.Articulo](t, reducida, derivada.argumentos)) {
			coinciden = append(coinciden, redaccion.fechaVigencia)
		}
	}

	ruta, orden := filepath.Join(carpeta, derivada.fichero), strings.Join(derivada.argumentos, " ")

	switch {
	case len(coinciden) == 0:
		return fmt.Errorf("%s: la redacción que da boe %s no es ninguna de las de la grabada", ruta, orden)
	case !slices.Equal(coinciden, []string{derivada.fechaVigencia}):
		return fmt.Errorf("%s: la redacción que da boe %s es la de vigencia %s de la grabada, y la derivada declara "+
			"la de %s", ruta, orden, strings.Join(coinciden, " y "), derivada.fechaVigencia)
	}

	return nil
}

// grabadaDeBoe es el contenido de la grabación de H4 del fichero.
func grabadaDeBoe(t *testing.T, fichero string) []byte {
	t.Helper()

	grabada, err := fs.ReadFile(os.DirFS(grabacionesDeBoe), fichero)
	require.NoError(t, err, "la derivada lleva el nombre de una grabación de H4")

	return grabada
}

// grabacionDeHTTPX es una grabación leída con los campos del formato de
// grabación de httpx en su orden (internal/httpx/grabar.go), que es el orden en
// que el codificador los escribe: lo que la derivación no toca sale como
// entró. Solo lleva el cuerpo como texto, el de las respuestas XML del BOE;
// leerLaGrabacion no admite ninguna otra clave.
type grabacionDeHTTPX struct {
	Formato   int              `json:"formato"`
	GrabadoEn string           `json:"grabado_en"`
	Peticion  peticionDeHTTPX  `json:"peticion"`
	Respuesta respuestaDeHTTPX `json:"respuesta"`
}

// peticionDeHTTPX es la petición de una grabación de httpx.
type peticionDeHTTPX struct {
	Metodo    string              `json:"metodo"`
	URL       string              `json:"url"`
	Cabeceras map[string][]string `json:"cabeceras"`
}

// respuestaDeHTTPX es la respuesta de una grabación de httpx, con el cuerpo
// como texto.
type respuestaDeHTTPX struct {
	Estado    int                 `json:"estado"`
	Cabeceras map[string][]string `json:"cabeceras"`
	Cuerpo    string              `json:"cuerpo"`
}

// leerLaGrabacion lee una grabación de httpx sin admitir ninguna clave que
// grabacionDeHTTPX no tenga: nada de lo grabado se pierde al escribirla.
func leerLaGrabacion(t *testing.T, contenido []byte) grabacionDeHTTPX {
	t.Helper()

	decodificador := json.NewDecoder(bytes.NewReader(contenido))
	decodificador.DisallowUnknownFields()

	var grabacion grabacionDeHTTPX

	require.NoError(t, decodificador.Decode(&grabacion), "la grabación tiene el formato de httpx")

	return grabacion
}

// escribirLaGrabacion es la grabación con el codificador de las grabaciones de
// httpx: sangrado de dos espacios, sin escapar HTML y con el salto de línea
// final que pone Encode.
func escribirLaGrabacion(t *testing.T, grabacion grabacionDeHTTPX) []byte {
	t.Helper()

	var contenido bytes.Buffer

	codificador := json.NewEncoder(&contenido)
	codificador.SetIndent("", "  ")
	codificador.SetEscapeHTML(false)

	require.NoError(t, codificador.Encode(grabacion))

	return contenido.Bytes()
}

// redaccionDelCuerpo es una <version> hija del <bloque> del cuerpo de una
// respuesta del BOE: su fecha de vigencia y el desplazamiento, en bytes del
// cuerpo, justo detrás de su </version>.
type redaccionDelCuerpo struct {
	fechaVigencia string
	fin           int64
}

// redaccionesDelCuerpo son las <version> hijas del <bloque> del cuerpo, en su
// orden, localizadas con encoding/xml.
func redaccionesDelCuerpo(t *testing.T, cuerpo string) []redaccionDelCuerpo {
	t.Helper()

	decodificador := xml.NewDecoder(strings.NewReader(cuerpo))

	var (
		abiertos    []string
		fecha       string
		redacciones []redaccionDelCuerpo
	)

	for {
		ficha, err := decodificador.Token()
		if errors.Is(err, io.EOF) {
			return redacciones
		}

		require.NoError(t, err, "el cuerpo de la grabación es XML")

		switch elemento := ficha.(type) {
		case xml.StartElement:
			abiertos = append(abiertos, elemento.Name.Local)

			if esVersionDelBloque(abiertos) {
				fecha = fechaDeVigenciaDeLaVersion(elemento)
			}
		case xml.EndElement:
			if esVersionDelBloque(abiertos) {
				redacciones = append(redacciones, redaccionDelCuerpo{fechaVigencia: fecha, fin: decodificador.InputOffset()})
			}

			abiertos = abiertos[:len(abiertos)-1]
		}
	}
}

// esVersionDelBloque dice si el último de los elementos abiertos es una
// <version> hija de un <bloque>.
func esVersionDelBloque(abiertos []string) bool {
	return len(abiertos) >= 2 && slices.Equal(abiertos[len(abiertos)-2:], []string{"bloque", "version"})
}

// fechaDeVigenciaDeLaVersion es el atributo fecha_vigencia de la <version>, o
// nada si no lo lleva.
func fechaDeVigenciaDeLaVersion(version xml.StartElement) string {
	for _, atributo := range version.Attr {
		if atributo.Name.Space == "" && atributo.Name.Local == "fecha_vigencia" {
			return atributo.Value
		}
	}

	return ""
}

// derivacionDelGrafoPrevio es la derivación de una derivada del grafo previo
// (contracts/eval-y-derivada.md §2 de H7.2): la grabación sin las redacciones
// posteriores a la de la fecha de vigencia, quitando del cuerpo los bytes desde
// el final del </version> de esa hasta el final del de la última —las
// redacciones posteriores y el blanco que las precede— y nada más, escrita con
// el codificador de las grabaciones.
func derivacionDelGrafoPrevio(t *testing.T, grabada []byte, fechaVigencia string) []byte {
	t.Helper()

	grabacion := leerLaGrabacion(t, grabada)
	cuerpo := grabacion.Respuesta.Cuerpo
	redacciones := redaccionesDelCuerpo(t, cuerpo)

	posicion := slices.IndexFunc(redacciones, func(redaccion redaccionDelCuerpo) bool {
		return redaccion.fechaVigencia == fechaVigencia
	})
	require.NotEqual(t, -1, posicion, "la grabación trae una redacción con la fecha de vigencia %s", fechaVigencia)

	grabacion.Respuesta.Cuerpo = cuerpo[:redacciones[posicion].fin] + cuerpo[redacciones[len(redacciones)-1].fin:]

	return escribirLaGrabacion(t, grabacion)
}

// ficherosDeLaCarpeta son las rutas de los ficheros regulares que hay por
// debajo de la carpeta, con ella delante; cualquier otra cosa que no sea un
// directorio hace fallar la prueba.
func ficherosDeLaCarpeta(t *testing.T, carpeta string) []string {
	t.Helper()

	var ficheros []string

	err := fs.WalkDir(os.DirFS(carpeta), ".", func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil || entrada.IsDir() {
			return err
		}

		require.True(t, entrada.Type().IsRegular(), "%s es un fichero regular", ruta)

		ficheros = append(ficheros, filepath.Join(carpeta, filepath.FromSlash(ruta)))

		return nil
	})
	require.NoError(t, err)

	return ficheros
}

// reproduccionConLaDerivada es una carpeta de reproducción nueva con las
// grabaciones de H4 y la derivada en lugar de la suya, como la deja un guion
// que la pone en juego con cp.
func reproduccionConLaDerivada(t *testing.T, derivada grabacionDerivada) string {
	t.Helper()

	contenido, err := fs.ReadFile(os.DirFS(derivada.carpeta), derivada.fichero)
	require.NoError(t, err)

	return reproduccionConLaGrabacion(t, derivada.fichero, contenido)
}

// reproduccionConLaGrabacion es una carpeta de reproducción nueva con las
// grabaciones de H4 y el contenido en lugar de la del fichero. La escribe a
// través de un os.Root: nada de lo que escribe puede salir de la carpeta.
func reproduccionConLaGrabacion(t *testing.T, fichero string, contenido []byte) string {
	t.Helper()

	carpeta := filepath.Join(t.TempDir(), boe.NombreDeLaFuente)
	require.NoError(t, os.CopyFS(carpeta, os.DirFS(grabacionesDeBoe)))

	raiz, err := os.OpenRoot(carpeta)
	require.NoError(t, err)

	escrita := raiz.WriteFile(fichero, contenido, 0o600)
	require.NoError(t, raiz.Close())
	require.NoError(t, escrita)

	return carpeta
}

// compruebaLaDerivacion lee con boe, con los argumentos, la reproducción
// original y la derivada, y exige que la derivada dé lo que da la original con
// el cambio aplicado; la premisa, que el cambio cambie algo.
func compruebaLaDerivacion[T any](t *testing.T, original, derivada string, argumentos []string, cambio func(*T)) {
	t.Helper()

	leido := leidoConBoe[T](t, original, argumentos)

	esperado := leidoConBoe[T](t, original, argumentos)
	cambio(&esperado)

	require.NotEqual(t, leido, esperado, "premisa: lo que dice el nombre cambia lo que da la grabación")
	assert.Equal(t, esperado, leidoConBoe[T](t, derivada, argumentos), "la derivada solo cambia lo que dice su nombre")
}

// leidoConBoe es el data de boe con los argumentos y --json sobre la
// reproducción de la carpeta, con una caché nueva, que tiene que salir con 0,
// leído en T sin admitir ninguna clave que T no tenga: lo que se compara es todo
// lo que boe da.
func leidoConBoe[T any](t *testing.T, carpeta string, argumentos []string) T {
	t.Helper()

	res := nuevoBancoDeBoe(t, carpeta).invocar(t, argvDeBoe(slices.Concat(argumentos, []string{"--json"})...)...)
	require.Equal(t, 0, res.codigo, res.errores)

	var sobre struct {
		Data json.RawMessage `json:"data"`
	}

	require.NoError(t, json.Unmarshal([]byte(res.salida), &sobre))

	decodificador := json.NewDecoder(bytes.NewReader(sobre.Data))
	decodificador.DisallowUnknownFields()

	var leido T

	require.NoError(t, decodificador.Decode(&leido), "%s", sobre.Data)

	return leido
}

// Las bases que ninguna entrega escribe —un esquema de una versión posterior,
// una fecha de consulta que no es RFC 3339— se escriben aquí byte a byte, con el
// formato de fichero de SQLite (https://www.sqlite.org/fileformat2.html): R3
// reserva database/sql y el controlador a los paquetes de almacenamiento, y este
// paquete no los importa ni en sus tests. Cada base lleva la página 1, con la
// cabecera del fichero y sqlite_schema, y una página hoja por tabla, con el
// diario clásico. Las sentencias no declaran más claves que el rowid: una clave
// de texto exigiría su índice en sqlite_schema, y las consultas del grafo no
// lo necesitan.

// paginaSQLite es el tamaño de página de esas bases.
const paginaSQLite = 4096

// tablaSQLite es una tabla: su nombre, la sentencia que la crea, como la guarda
// sqlite_schema, y sus filas en orden de rowid.
type tablaSQLite struct {
	nombre    string
	sentencia string
	filas     []filaSQLite
}

// filaSQLite es una fila: su rowid y sus valores, cada uno nil, un entero de 0
// a 127 o un texto.
type filaSQLite struct {
	rowid   int
	valores []any
}

// sentenciaDeSchemaVersion es la de schema_version, con la versión como alias
// del rowid, igual que en la migración 0001: su valor es el rowid de la fila y
// el registro lo lleva nulo.
const sentenciaDeSchemaVersion = "CREATE TABLE schema_version (version INTEGER PRIMARY KEY, aplicada_en TEXT NOT NULL)"

// baseDeUnaVersionPosterior es un world.db con el esquema en la versión 3: el de
// un binario posterior que migró desde la 1 y la 2 (FR-012).
func baseDeUnaVersionPosterior(t *testing.T) []byte {
	t.Helper()

	return baseSQLite(t, tablaSQLite{
		nombre:    "schema_version",
		sentencia: sentenciaDeSchemaVersion,
		filas: []filaSQLite{
			{rowid: 1, valores: []any{nil, "2026-09-01T00:00:00Z"}},
			{rowid: 2, valores: []any{nil, "2026-09-02T00:00:00Z"}},
			{rowid: 3, valores: []any{nil, "2026-09-03T00:00:00Z"}},
		},
	})
}

// baseSQLite escribe la base con las tablas dadas: la página 1 con la cabecera
// y sqlite_schema, que nombra la página raíz de cada tabla, y las hojas de las
// tablas en las páginas siguientes, en su orden.
func baseSQLite(t *testing.T, tablas ...tablaSQLite) []byte {
	t.Helper()

	paginas := 1 + len(tablas)
	base := make([]byte, paginas*paginaSQLite)
	esquema := make([]filaSQLite, 0, len(tablas))

	for i, tabla := range tablas {
		raiz := i + 2
		esquema = append(esquema, filaSQLite{
			rowid:   i + 1,
			valores: []any{"table", tabla.nombre, tabla.nombre, raiz, tabla.sentencia},
		})

		hojaSQLite(t, base[(raiz-1)*paginaSQLite:raiz*paginaSQLite], 0, tabla.filas)
	}

	cabeceraSQLite(t, base, paginas)
	hojaSQLite(t, base[:paginaSQLite], 100, esquema)

	return base
}

// cabeceraSQLite escribe los 100 bytes de la cabecera del fichero. Los campos
// que no se nombran van a cero: sin páginas libres, sin tamaño de caché
// sugerido, sin autovacuum y sin versión de usuario.
func cabeceraSQLite(t *testing.T, base []byte, paginas int) {
	t.Helper()

	copy(base, "SQLite format 3\x00")
	dosBytesSQLite(t, base[16:18], paginaSQLite)
	base[18], base[19] = 1, 1                 // escritura y lectura con el diario clásico, no WAL
	base[21], base[22], base[23] = 64, 32, 32 // las fracciones de carga útil, fijas en el formato
	base[27] = 1                              // el contador de cambios
	dosBytesSQLite(t, base[30:32], paginas)   // el tamaño en páginas; sus dos bytes altos, a cero
	base[43] = 1                              // la cookie del esquema
	base[47] = 4                              // el formato del esquema
	base[59] = 1                              // la codificación, UTF-8
	base[95] = 1                              // el tamaño en páginas vale para este contador de cambios
}

// hojaSQLite escribe en la página una hoja de tabla con las filas: su cabecera
// desde el byte dado —100 en la página 1, detrás de la del fichero; 0 en las
// demás—, los punteros a las celdas en orden de rowid detrás de ella y las
// celdas desde el final de la página hacia atrás.
func hojaSQLite(t *testing.T, pagina []byte, desde int, filas []filaSQLite) {
	t.Helper()

	pagina[desde] = 0x0d // hoja de una tabla
	dosBytesSQLite(t, pagina[desde+3:desde+5], len(filas))

	punteros := desde + 8
	fin := len(pagina)

	for i, fila := range filas {
		celda := celdaSQLite(t, fila)
		fin -= len(celda)
		require.GreaterOrEqual(t, fin, punteros+2*len(filas), "las filas caben en una página")

		copy(pagina[fin:], celda)
		dosBytesSQLite(t, pagina[punteros+2*i:punteros+2*i+2], fin)
	}

	dosBytesSQLite(t, pagina[desde+5:desde+7], fin) // el comienzo del área de celdas
}

// celdaSQLite es la celda de una fila: la longitud del registro, el rowid y el
// registro, que es su cabecera —su longitud y el tipo de cada valor— y los
// valores. Ninguna fila de estas bases desborda la página.
func celdaSQLite(t *testing.T, fila filaSQLite) []byte {
	t.Helper()

	var tipos, valores []byte

	for _, valor := range fila.valores {
		switch v := valor.(type) {
		case nil:
			tipos = append(tipos, 0)
		case int:
			tipos = append(tipos, 1) // entero de un byte
			valores = append(valores, unByteSQLite(t, v))
		case string:
			tipos = append(tipos, varintSQLite(t, 13+2*len(v))...)
			valores = append(valores, v...)
		default:
			require.FailNow(t, "un valor sin tipo en estas bases", "%#v", valor)
		}
	}

	registro := slices.Concat(varintSQLite(t, 1+len(tipos)), tipos, valores)
	require.Less(t, len(tipos), 0x7f, "la cabecera del registro cabe en un byte de longitud")
	require.LessOrEqual(t, len(registro), paginaSQLite-35, "el registro no desborda la página")

	return slices.Concat(varintSQLite(t, len(registro)), varintSQLite(t, fila.rowid), registro)
}

// varintSQLite es el entero de longitud variable de SQLite, de uno o dos bytes:
// siete bits por byte, los más altos primero, con el bit alto puesto en todos
// menos en el último. Un entero de 0 a 127 es también, tal cual, un entero de un
// byte de un registro.
func varintSQLite(t *testing.T, valor int) []byte {
	t.Helper()

	if valor < 0x80 {
		return []byte{unByteSQLite(t, valor)}
	}

	return []byte{0x80 | unByteSQLite(t, valor>>7), unByteSQLite(t, valor&0x7f)}
}

// unByteSQLite es el entero en un byte, que tiene que caber en siete bits.
func unByteSQLite(t *testing.T, valor int) byte {
	t.Helper()

	if valor < 0 || valor > 0x7f {
		require.FailNow(t, "el entero no cabe en siete bits", "%d", valor)

		return 0
	}

	return byte(valor)
}

// dosBytesSQLite escribe el entero en dos bytes, los altos primero.
func dosBytesSQLite(t *testing.T, destino []byte, valor int) {
	t.Helper()

	if valor < 0 || valor > math.MaxUint16 {
		require.FailNow(t, "el entero no cabe en dos bytes", "%d", valor)

		return
	}

	binary.BigEndian.PutUint16(destino, uint16(valor))
}

// Lo que ninguna lectura de internal/graph devuelve —la procedencia de cada
// texto— se lee aquí del fichero, con el mismo formato
// (https://www.sqlite.org/fileformat2.html) y por la misma razón que las bases
// de arriba se escriben a mano: R3. El lector baja por el árbol de una tabla
// desde la página raíz que nombra sqlite_schema, por sus páginas interiores
// hasta sus hojas, sigue las páginas de desbordamiento de una fila que no cabe
// en su hoja y lee los valores de un registro de los tipos que escribe la
// migración 0001: nulo, entero y texto.

// lectorSQLite es una base SQLite entera en memoria, con el tamaño de sus
// páginas y el útil de cada una, sin su espacio reservado.
type lectorSQLite struct {
	base   []byte
	tamano int
	util   int
}

// filasDeLaTablaSQLite son las filas de la tabla de la base, en orden de rowid,
// cada una con sus valores: nil, un int64 o un texto.
func filasDeLaTablaSQLite(t *testing.T, base []byte, tabla string) [][]any {
	t.Helper()

	lector := nuevoLectorSQLite(t, base)

	for _, carga := range lector.cargas(t, 1) {
		// Cada fila de sqlite_schema: su tipo, su nombre, su tabla, su página
		// raíz y su sentencia.
		esquema := valoresSQLite(t, carga)
		require.Len(t, esquema, 5, "una fila de sqlite_schema")

		if esquema[0] != "table" || esquema[1] != tabla {
			continue
		}

		raiz, esEntero := esquema[3].(int64)
		require.True(t, esEntero, "la página raíz de %q es un entero", tabla)

		var filas [][]any

		for _, carga := range lector.cargas(t, int(raiz)) {
			filas = append(filas, valoresSQLite(t, carga))
		}

		return filas
	}

	require.FailNow(t, "la tabla no está en la base", "%q", tabla)

	return nil
}

// nuevoLectorSQLite es el lector de la base, con el tamaño de página de su
// cabecera —1 es 65 536— y el útil, sin los bytes que reserva.
func nuevoLectorSQLite(t *testing.T, base []byte) lectorSQLite {
	t.Helper()

	require.GreaterOrEqual(t, len(base), 100, "la base tiene su cabecera")
	require.Equal(t, "SQLite format 3\x00", string(base[:16]), "premisa: es una base SQLite")

	tamano := int(binary.BigEndian.Uint16(base[16:18]))
	if tamano == 1 {
		tamano = 1 << 16
	}

	require.Zero(t, len(base)%tamano, "la base son páginas enteras")

	return lectorSQLite{base: base, tamano: tamano, util: tamano - int(base[20])}
}

// pagina es la página del número, contando desde 1.
func (l lectorSQLite) pagina(t *testing.T, numero int) []byte {
	t.Helper()

	require.True(t, numero >= 1 && numero*l.tamano <= len(l.base), "la página %d está en la base", numero)

	return l.base[(numero-1)*l.tamano : numero*l.tamano]
}

// cargas son las de las filas del árbol de tabla con raíz en la página, en
// orden de rowid: las de las celdas de una hoja y, en una página interior, las
// del hijo de cada celda y después las del hijo de más a la derecha. La
// cabecera de la página 1 va detrás de la del fichero.
func (l lectorSQLite) cargas(t *testing.T, raiz int) [][]byte {
	t.Helper()

	pagina := l.pagina(t, raiz)

	cabecera := 0
	if raiz == 1 {
		cabecera = 100
	}

	celdas := int(binary.BigEndian.Uint16(pagina[cabecera+3:]))

	var cargas [][]byte

	switch pagina[cabecera] {
	case 0x0d: // hoja de una tabla
		for i := range celdas {
			cargas = append(cargas, l.cargaDeLaCelda(t, pagina, punteroSQLite(pagina, cabecera+8, i)))
		}
	case 0x05: // página interior de una tabla
		for i := range celdas {
			hijo := binary.BigEndian.Uint32(pagina[punteroSQLite(pagina, cabecera+12, i):])
			cargas = append(cargas, l.cargas(t, int(hijo))...)
		}

		cargas = append(cargas, l.cargas(t, int(binary.BigEndian.Uint32(pagina[cabecera+8:])))...)
	default:
		require.FailNow(t, "la página no es de un árbol de tabla", "página %d, tipo %#x", raiz, pagina[cabecera])
	}

	return cargas
}

// punteroSQLite es el desplazamiento de la celda i de la página, el que guarda
// en dos bytes su puntero en el vector que empieza en punteros.
func punteroSQLite(pagina []byte, punteros, i int) int {
	return int(binary.BigEndian.Uint16(pagina[punteros+2*i:]))
}

// cargaDeLaCelda es la carga entera de la celda de una hoja de tabla que empieza
// en el desplazamiento: detrás de su longitud y de su rowid, dos enteros de
// longitud variable, la parte que cabe en la página y, si no cabe entera, el
// número de la primera página de desbordamiento, cada una de las cuales guarda
// en sus cuatro primeros bytes el de la siguiente y después lo que sigue.
func (l lectorSQLite) cargaDeLaCelda(t *testing.T, pagina []byte, desde int) []byte {
	t.Helper()

	longitud, bytesDeLaLongitud := varintDeSQLite(t, pagina[desde:])
	_, bytesDelRowid := varintDeSQLite(t, pagina[desde+bytesDeLaLongitud:])
	desde += bytesDeLaLongitud + bytesDelRowid

	total := int(longitud)
	local := l.local(total)
	require.LessOrEqual(t, desde+local, len(pagina), "la parte local de la carga cabe en la página")

	carga := slices.Clone(pagina[desde : desde+local])
	if local == total {
		return carga
	}

	siguiente := int(binary.BigEndian.Uint32(pagina[desde+local:]))

	for len(carga) < total {
		require.NotZero(t, siguiente, "las páginas de desbordamiento llegan al final de la carga")

		desbordada := l.pagina(t, siguiente)
		siguiente = int(binary.BigEndian.Uint32(desbordada))
		carga = append(carga, desbordada[4:4+min(total-len(carga), l.util-4)]...)
	}

	return carga
}

// local es la parte de una carga de esa longitud que se guarda en su celda de
// una hoja de tabla, con las cotas del formato: entera hasta el máximo; si no,
// el mínimo más lo que sobre de llenar páginas de desbordamiento enteras, si
// cabe, o el mínimo.
func (l lectorSQLite) local(longitud int) int {
	maximo := l.util - 35
	if longitud <= maximo {
		return longitud
	}

	minimo := (l.util-12)*32/255 - 23

	if local := minimo + (longitud-minimo)%(l.util-4); local <= maximo {
		return local
	}

	return minimo
}

// valoresSQLite son los valores del registro de una carga: detrás de la
// longitud de su cabecera, el tipo de cada valor, enteros de longitud variable,
// y detrás de la cabecera los valores, en ese orden.
func valoresSQLite(t *testing.T, carga []byte) []any {
	t.Helper()

	longitud, desde := varintDeSQLite(t, carga)
	cabecera := int(longitud)
	require.LessOrEqual(t, cabecera, len(carga), "la cabecera del registro cabe en su carga")

	var tipos []int

	for desde < cabecera {
		tipo, bytes := varintDeSQLite(t, carga[desde:])
		tipos = append(tipos, int(tipo))
		desde += bytes
	}

	valores := make([]any, 0, len(tipos))

	for _, tipo := range tipos {
		valor, bytes := valorSQLite(t, tipo, carga[desde:])
		valores = append(valores, valor)
		desde += bytes
	}

	require.Equal(t, len(carga), desde, "el registro ocupa la carga entera")

	return valores
}

// valorSQLite es el valor de ese tipo del principio de los bytes, y cuántos
// ocupa: nulo; un entero de 1, 2, 3, 4, 6 u 8 bytes, o las constantes 0 y 1; o
// un texto de (tipo - 13) / 2 bytes. Los flotantes y los blobs no los escribe
// la migración 0001.
func valorSQLite(t *testing.T, tipo int, datos []byte) (any, int) {
	t.Helper()

	switch {
	case tipo == 0:
		return nil, 0
	case tipo >= 1 && tipo <= 6:
		bytes := bytesDeUnEnteroSQLite(tipo)
		require.LessOrEqual(t, bytes, len(datos), "el entero está entero")

		return enteroSQLite(datos[:bytes]), bytes
	case tipo == 8 || tipo == 9:
		return int64(tipo - 8), 0
	case tipo >= 13 && tipo%2 == 1:
		bytes := (tipo - 13) / 2
		require.LessOrEqual(t, bytes, len(datos), "el texto está entero")

		return string(datos[:bytes]), bytes
	default:
		require.FailNow(t, "un valor de un tipo que la migración 0001 no escribe", "tipo %d", tipo)

		return nil, 0
	}
}

// bytesDeUnEnteroSQLite son los que ocupa en un registro un entero de ese
// tipo, del 1 al 6: los mismos que el tipo hasta el 4, 6 el 5 y 8 el 6.
func bytesDeUnEnteroSQLite(tipo int) int {
	switch tipo {
	case 5:
		return 6
	case 6:
		return 8
	default:
		return tipo
	}
}

// enteroSQLite es el entero con signo de los bytes, los altos primero, en
// complemento a dos.
func enteroSQLite(datos []byte) int64 {
	var valor int64

	for _, b := range datos {
		valor = valor<<8 | int64(b)
	}

	extension := 64 - 8*len(datos)

	return valor << extension >> extension
}

// varintDeSQLite es el entero de longitud variable del principio de los datos,
// y cuántos bytes ocupa: siete bits por byte, los más altos primero, mientras
// el bit alto está puesto, y hasta un noveno byte, que aporta los ocho suyos.
func varintDeSQLite(t *testing.T, datos []byte) (int64, int) {
	t.Helper()

	var valor int64

	for i := range 8 {
		require.Less(t, i, len(datos), "el entero de longitud variable está entero")

		valor = valor<<7 | int64(datos[i]&0x7f)
		if datos[i]&0x80 == 0 {
			return valor, i + 1
		}
	}

	require.Less(t, 8, len(datos), "el entero de longitud variable está entero")

	return valor<<8 | int64(datos[8]), 9
}
