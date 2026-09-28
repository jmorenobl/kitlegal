package app

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"math"
	"os"
	"path/filepath"
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

// Las descripciones literales de contracts/applet-graph.md §1.
const (
	descripcionDelGrafo = "Lee el grafo del mundo: lo que el binario ha observado de las fuentes, con su procedencia."
	ayudaDelID          = "Id del nodo: un ELI, \xc2\xabine:<c\xc3\xb3digo>\xc2\xbb, un DIR3\xe2\x80\xa6"
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
		{
			nombre: "check",
			descripcion: "Comprueba el grafo del mundo y devuelve como hallazgos las versiones superadas y las" +
				" consultas caducadas.",
			salida: []grafo.Hallazgo(nil),
		},
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

// hallazgosDeLaMuestra es el data de check sobre la muestra en instanteDelGrafo:
// una fuente caducada por cada nodo del primer lote, que declaró una semana de
// vigencia, y ninguna por los del segundo, que no declaró ninguna; en orden de
// id y con la cita de cada tipo (contracts/applet-graph.md §3.3 y §5).
func hallazgosDeLaMuestra() string {
	procedencia := procedenciaJSON(fuenteDeLaNorma, urlDeLaNorma, fechaDeLaNorma)
	hallazgo := func(id, cita string) string {
		explicacion := "La consulta de " + cita + " a " + fuenteDeLaNorma + " en " + urlDeLaNorma + " del " +
			fechaDeLaNorma + " ten\xc3\xada una vigencia de 604800 s y caduc\xc3\xb3 el " + caducidadDeLaNorma + "."

		return fmt.Sprintf(`{"clase": "fuente-caducada", "id": %q, "explicacion": %q, "procedencia": %s,
			"vigencia_segundos": 604800}`, id, explicacion, procedencia)
	}
	citaDelBloque := "[" + identificadorDeLaNorma + ", bloque a1]"

	return "[" + strings.Join([]string{
		hallazgo(idDeLaNorma, identificadorDeLaNorma),
		hallazgo(idDelBloque, citaDelBloque),
		hallazgo(idDeLaVersion(), citaDelBloque),
	}, ",") + "]"
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
// fijo y world.db bajo t.TempDir() (contracts/applet-graph.md §1-§4; research.md
// D19): su declaración y sus textos literales, la ayuda y --describe de cada
// verbo; el grafo ausente, que se lee como vacío sin crear nada; show, stats y
// check sobre la muestra, con la firma del applet y la fecha de su reloj, que
// se lee una vez por invocación y es el instante de check; los ids que no están
// (3) y los argumentos que no valen (2); --no-graph, --offline y el nombre del
// programa, que no cambian nada; la tabla mínima sin --json y sin texto legal;
// --dry-run, que lee igual; el plazo agotado (4); y el reloj que falta y lo que
// ninguna entrega guarda (1). Ningún verbo cambia un byte del directorio.
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
		{nombre: "misma-salida-con-no-graph-offline-y-multicall", comprobar: compruebaMismaSalidaDelGrafo},
		{nombre: "tabla-sin-json-y-sin-texto-legal", comprobar: compruebaTablaDelGrafo},
		{nombre: "dry-run", comprobar: compruebaEnsayoDelGrafo},
		{nombre: "plazo-agotado", comprobar: compruebaPlazoDelGrafo},
		{nombre: "sin-reloj", comprobar: compruebaGrafoSinReloj},
		{nombre: "lo-que-ninguna-entrega-guarda", comprobar: compruebaGrafoIncomprobable},
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

	for _, argumentos := range [][]string{{"show", idDelBloque}, {"stats"}, {"check"}} {
		verbo := argumentos[0]
		descripcion := invocar(t, registro, argvDelGrafo(slices.Concat(argumentos, []string{"--describe"})...)...)
		require.Equal(t, 0, descripcion.codigo, descripcion.errores)

		var documento map[string]any
		require.NoError(t, json.Unmarshal([]byte(descripcion.salida), &documento), verbo)
		assert.Equal(t, "graph "+verbo, documento["title"])
	}

	assert.Zero(t, reloj.lecturas, "ni la ayuda ni --describe ejecutan ningún verbo")

	entradas, err := os.ReadDir(directorio)
	require.NoError(t, err)
	assert.Empty(t, entradas, "ni la ayuda ni --describe crean nada")
}

// compruebaGrafoAusente fija el grafo ausente —el directorio vacío y uno que ni
// siquiera existe—: stats da tres ceros y dos listas vacías, check una lista
// vacía y show un 3 que nombra el id, y ninguno crea world.db, sus auxiliares
// ni su directorio (FR-004, FR-053, FR-054, FR-060).
func compruebaGrafoAusente(t *testing.T) {
	t.Helper()

	vacio := t.TempDir()
	inexistente := filepath.Join(t.TempDir(), "no", "existe")

	for _, directorio := range []string{vacio, inexistente} {
		registro := registroDelGrafo(t, nuevoReloj(t, instanteDelGrafo).ahora, directorio)

		stats := invocar(t, registro, argvDelGrafo("stats", "--json")...)
		assert.JSONEq(t, recuentoVacio, datosFirmados(t, stats, instanteDelGrafo), directorio)

		check := invocar(t, registro, argvDelGrafo("check", "--json")...)
		assert.JSONEq(t, `[]`, datosFirmados(t, check, instanteDelGrafo), directorio)

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
		{instante: instanteSinCaducar, datos: `[]`},
		{instante: caducidadDeLaNorma, datos: `[]`},
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
// tal cual (FR-052, FR-053). El mensaje nombra el id y nada cambia.
//
// El byte que no es UTF-8 no llega al applet: el analizador de la línea de
// órdenes, en internal/cli, lleva cada argumento de texto por JSON, que lo
// cambia por U+FFFD. El applet busca tal cual el id que recibe, que tampoco
// puede ser de ningún nodo, y lo nombra así (gates/supuestos.md, T015).
func compruebaIDsQueNoEstan(t *testing.T) {
	t.Helper()

	directorio := t.TempDir()
	poblarElGrafo(t, directorio)

	antes := huellasDelDirectorio(t, directorio)
	registro := registroDelGrafo(t, nuevoReloj(t, instanteDelGrafo).ahora, directorio)

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
		recibido := strings.ToValidUTF8(id, "\xef\xbf\xbd")
		assert.Contains(t, mensaje, strconv.Quote(recibido), "el mensaje nombra el id %q", id)
	}

	assert.Equal(t, antes, huellasDelDirectorio(t, directorio))
}

// compruebaArgumentosDelGrafo fija el 2 de los argumentos que no valen
// (FR-052, FR-054, FR-060; contracts/applet-graph.md §4): los que rechaza el
// analizador antes del applet —show sin id y con dos, stats y check con uno, una
// bandera desconocida—, firmados por el kernel, y los ids que rechaza el
// applet —vacío, solo espacio en blanco o con un carácter de control—, firmados
// por él. Nada cambia.
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
		{"check", idDeLaNorma},
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

	assert.Equal(t, antes, huellasDelDirectorio(t, directorio))
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

// compruebaTablaDelGrafo fija la forma legible sin --json (FR-055): la tabla
// mínima del kernel, con las cuatro líneas de la firma —la misma huella que el
// sobre— y el contenido aplanado, sin texto propio; y que ninguna línea del
// texto guardado sale por ningún verbo, con --json ni sin él (FR-070).
func compruebaTablaDelGrafo(t *testing.T) {
	t.Helper()

	directorio := t.TempDir()
	poblarElGrafo(t, directorio)

	registro := registroDelGrafo(t, nuevoReloj(t, instanteDelGrafo).ahora, directorio)

	casos := []struct {
		argumentos []string
		filas      []string
	}{
		{
			argumentos: []string{"show", idDelBloque},
			filas: []string{
				`nodo\.id +` + regexp.QuoteMeta(idDelBloque), `nodo\.tipo +Bloque`,
				`entrantes\.0\.relacion +eli:has_part`, `salientes\.0\.id +` + regexp.QuoteMeta(idDeLaVersion()),
			},
		},
		{argumentos: []string{"stats"}, filas: []string{`nodos +5`, `aristas +3`, `textos +1`}},
		{argumentos: []string{"check"}, filas: []string{`0\.clase +fuente-caducada`, `2\.vigencia_segundos +604800`}},
	}

	for _, caso := range casos {
		tabla := invocar(t, registro, argvDelGrafo(caso.argumentos...)...)
		require.Equal(t, 0, tabla.codigo, tabla.errores)
		assert.Empty(t, tabla.errores)

		enJSON := invocar(t, registro, argvDelGrafo(slices.Concat(caso.argumentos, []string{"--json"})...)...)
		huella := sobreDelJSON(t, enJSON.salida)["hash"]

		assert.Regexp(t, `\Afuente +kitlegal\.graph\nurl +kitlegal:applet/graph\n`+
			`fecha_consulta +`+regexp.QuoteMeta(instanteDelGrafo)+`\nhash +`+fmt.Sprint(huella)+`\n`, tabla.salida)

		for _, fila := range caso.filas {
			assert.Regexp(t, `(?m)^`+fila+`$`, tabla.salida, "%q", caso.argumentos)
		}

		for _, linea := range strings.Split(cuerpoDelBloque, "\n") {
			for _, salida := range []string{tabla.salida, enJSON.salida} {
				assert.NotContains(t, salida, linea, "%q no devuelve texto legal", caso.argumentos)
			}
		}
	}
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

// compruebaGrafoIncomprobable fija el 1 de check sobre un world.db que se lee
// pero que guarda lo que ninguna entrega escribe: una fecha de consulta que no
// es RFC 3339 (research.md D15; gates/supuestos.md, T014). El mensaje nombra
// world.db y no el id del nodo, y nada cambia. La premisa, que show lee esa base
// y devuelve el nodo con esa fecha, dice que el fallo es de la comprobación y
// no de una base que no se puede abrir.
func compruebaGrafoIncomprobable(t *testing.T) {
	t.Helper()

	directorio := t.TempDir()
	escribirFicheroDePrueba(t, filepath.Join(directorio, "world.db"), baseConUnaFechaIlegible(t))

	antes := huellasDelDirectorio(t, directorio)
	registro := registroDelGrafo(t, nuevoReloj(t, instanteDelGrafo).ahora, directorio)

	show := invocar(t, registro, argvDelGrafo("show", idDeLaNorma, "--json")...)
	assert.Contains(t, datosFirmados(t, show, instanteDelGrafo), `"primera_observacion":"ayer"`,
		"premisa: la base se lee")

	check := invocar(t, registro, argvDelGrafo("check", "--json")...)
	mensaje := exigirFalloDelGrafo(t, check, schema.ClaseInesperado, 1, instanteDelGrafo)
	assert.True(t, strings.HasPrefix(mensaje, "grafo: world.db "), "el mensaje nombra world.db: %q", mensaje)
	assert.NotContains(t, mensaje, idDeLaNorma, "ni el id del nodo, que puede ser de una Persona")

	assert.Equal(t, antes, huellasDelDirectorio(t, directorio))
}

// TestCodigosDelGrafo fija los códigos de lo que no deja leer world.db, con las
// dependencias de la raíz de producción y la regla de ubicación de la caché
// (contracts/applet-graph.md §4; FR-010, FR-011; SC-011):
//
//   - world.db que no es una base de datos, que es un directorio o cuyo esquema
//     es de una versión posterior: 1 en los tres verbos, con y sin --no-graph,
//     con el mensaje que nombra su ruta y el directorio igual byte a byte;
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

				assert.Equal(t, "grafo: "+strconv.Quote(ruta)+" "+caso.motivo, mensaje)
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
}

// casosInutilizables son los tres de SC-011.
func casosInutilizables() []casoInutilizable {
	return []casoInutilizable{
		{
			nombre: "no-es-una-base",
			preparar: func(t *testing.T, ruta string) {
				t.Helper()

				escribirFicheroDePrueba(t, ruta, []byte("Este fichero no es una base de datos SQLite.\n"))
			},
			motivo: "no es una base de datos utilizable; no se modifica",
		},
		{
			nombre: "es-un-directorio",
			preparar: func(t *testing.T, ruta string) {
				t.Helper()

				require.NoError(t, os.Mkdir(ruta, 0o700))
			},
			motivo: "es un directorio y no una base de datos utilizable; no se modifica",
		},
		{
			nombre: "esquema-de-una-version-posterior",
			preparar: func(t *testing.T, ruta string) {
				t.Helper()

				escribirFicheroDePrueba(t, ruta, baseDeUnaVersionPosterior(t))
			},
			motivo: "tiene el esquema en la versi\xc3\xb3n 2 y este binario conoce la 1: no se modifica",
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
// observaría boe articulo tras el cambio. En instanteDelGrafo su consulta sigue
// vigente y la de la versión de la muestra ha caducado, así que check da un
// hallazgo de cada clase sobre esta última.
const (
	fechaDeLaVersionPosterior  = "2026-10-02T10:00:00Z"
	fechaDeVigenciaPosterior   = "20270101"
	cuerpoDeLaVersionPosterior = cuerpoDelBloque + "\nSu redacci\xc3\xb3n posterior de prueba a\xc3\xb1ade esta frase."
)

// poblarConUnaVersionPosterior entrega la muestra y, después, el lote de la
// versión posterior al world.db del directorio.
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

// conWorldDBDirectorio pone en el lugar de world.db un directorio, que no se
// puede leer (contracts/almacen-world-db.md §6).
func conWorldDBDirectorio(t *testing.T, directorio string) {
	t.Helper()

	require.NoError(t, os.Mkdir(filepath.Join(directorio, "world.db"), 0o700))
}

// TestSalidaDelGrafoContraSchemas es el punto 4 de la Definition of Done sobre
// el applet graph (FR-051, FR-094; contracts/applet-graph.md §3, §6 y §7): el
// sobre real que emite el kernel con --json, sobre el registro local, valida
// contra la parte de su verbo leída de schemas/grafo.json, y no contra lo que
// emite --describe mientras se ejecuta el test. Lo hace toda salida correcta de
// show —la de cada tipo de nodo, con aristas en los dos sentidos y con alguna
// lista vacía—, de stats —con el grafo ausente y con nodos— y de check —con
// hallazgos de las dos clases, de una sola, sin hallazgos y con el grafo
// ausente—, y también la de cada fallo que decide el applet, con 2, 3, 4 y 1
// (gates/supuestos.md, T017); los que decide el kernel antes de llegar al
// applet no son salida suya. La validación restringe: el mismo sobre con una
// clave de más o de menos en su data no valida, y en check tampoco con un
// elemento que no es un hallazgo ni con la lista nula.
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
	sinHallazgos := `"data":[]`

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

		{"check-sin-world-db", sinWorldDB, instanteDelGrafo, 0, []string{"check"}, []string{sinHallazgos}},
		{"check-sin-hallazgos", poblarElGrafo, instanteSinCaducar, 0, []string{"check"}, []string{sinHallazgos}},
		{"check-con-consultas-caducadas", poblarElGrafo, instanteDelGrafo, 0, []string{"check"}, []string{caducada}},
		{
			"check-con-hallazgos-de-las-dos-clases", poblarConUnaVersionPosterior, instanteDelGrafo, 0,
			[]string{"check"},
			[]string{caducada, obsoleta, `"fecha_vigencia_reciente":"` + fechaDeVigenciaPosterior + `"`},
		},

		{
			"fallo-2-id-en-blanco", poblarElGrafo, instanteDelGrafo, 2,
			[]string{"show", " "},
			[]string{`"clase":"argumentos"`},
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
			"fallo-1-world-db-es-un-directorio", conWorldDBDirectorio, instanteDelGrafo, 1,
			[]string{"stats"},
			[]string{`"clase":"inesperado"`},
		},
	}
}

// exigirSalidaDelGrafoPublicada exige que el sobre real valide contra la parte
// publicada de su verbo y que la validación restrinja data. Si data es un
// objeto —la ficha de show, el recuento de stats o los datos de un fallo—, lo
// exige exigirSalidaPublicada; si es la lista de check, exigirHallazgosPublicados.
func exigirSalidaDelGrafoPublicada(t *testing.T, esquema *jsonschema.Schema, salida string) {
	t.Helper()

	if _, esLista := sobreValidable(t, salida)["data"].([]any); esLista {
		exigirHallazgosPublicados(t, esquema, salida)

		return
	}

	exigirSalidaPublicada(t, esquema, salida)
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
// parte publicada y que la validación restrinja la lista: no valida nula
// (FR-060), ni con un elemento de más que no es un hallazgo —lo que también se
// comprueba con la lista vacía—, ni con una clave de más en cualquiera de sus
// hallazgos, ni sin cualquiera de las cuatro de todo hallazgo. Cada hallazgo
// lleva exactamente esas cuatro y las de su clase.
func exigirHallazgosPublicados(t *testing.T, esquema *jsonschema.Schema, salida string) {
	t.Helper()

	require.NoError(t, esquema.Validate(sobreValidable(t, salida)), "el sobre real valida contra su parte de schemas/")

	nula := sobreValidable(t, salida)
	nula["data"] = nil
	require.Error(t, esquema.Validate(nula), "la lista nula no valida")

	conUnoDeMas := sobreValidable(t, salida)
	conUnoDeMas["data"] = append(hallazgosDelSobre(t, conUnoDeMas), map[string]any{"ajena": "no declarada"})
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

// hallazgosDelSobre es la lista de data del sobre de check.
func hallazgosDelSobre(t *testing.T, sobre map[string]any) []any {
	t.Helper()

	hallazgos, esLista := sobre["data"].([]any)
	require.True(t, esLista, "el data de check es una lista")

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
// sin ella, en la tabla, que escribe los textos tal cual.
//
// Las premisas dicen que no pasa en vacío: cada línea buscada aparece, con la
// misma búsqueda, en la tabla del boe articulo que la devolvió; el grafo guarda
// un texto por cada huella que boe publicó y una BloqueVersion con cada una; y
// check, en un instante en que todas las consultas han caducado, devuelve
// hallazgos, cuyas explicaciones citan cada bloque.
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
				assert.NotEqual(t, "[]", datosFirmados(t, res, instanteDelGrafo), "premisa: check da hallazgos")
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

// bloqueGrabado es el bloque de una norma cuya respuesta está grabada.
type bloqueGrabado struct {
	norma  string
	bloque string
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
		if partes == nil || grabacion.Respuesta.Estado != 200 {
			continue
		}

		bloques = append(bloques, bloqueGrabado{norma: partes[1], bloque: partes[2]})
	}

	return bloques
}

// registroDeBoeQueEntrega es el registro de producción en lo que aquí importa:
// el applet boe servido desde la reproducción de la carpeta, con una caché
// nueva, y la entrega al world.db del directorio.
func registroDeBoeQueEntrega(t *testing.T, carpeta, directorio string) *Registro {
	t.Helper()

	var registro Registro

	require.NoError(t, registro.Registrar(AppletBoe(nuevoBancoDeBoe(t, carpeta).dependencias)))
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

	instantanea, err := lectura.Instantanea(t.Context())
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

// escribirFicheroDePrueba escribe un fichero de la prueba con acceso reservado a la
// cuenta.
func escribirFicheroDePrueba(t *testing.T, ruta string, contenido []byte) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Clean(ruta), contenido, 0o600))
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

// baseDeUnaVersionPosterior es un world.db con el esquema en la versión 2: el de
// un binario posterior que migró desde la 1 (FR-012).
func baseDeUnaVersionPosterior(t *testing.T) []byte {
	t.Helper()

	return baseSQLite(t, tablaSQLite{
		nombre:    "schema_version",
		sentencia: sentenciaDeSchemaVersion,
		filas: []filaSQLite{
			{rowid: 1, valores: []any{nil, "2026-09-01T00:00:00Z"}},
			{rowid: 2, valores: []any{nil, "2026-09-02T00:00:00Z"}},
		},
	})
}

// baseConUnaFechaIlegible es un world.db con el esquema en la versión 1 y un
// nodo cuya fecha de consulta, «ayer», no es RFC 3339: lo que ninguna entrega
// escribe, porque ValidarLote la rechaza.
func baseConUnaFechaIlegible(t *testing.T) []byte {
	t.Helper()

	// Los campos de observación de nodes y de edges, los de la migración 0001.
	observaciones := "first_seen TEXT, first_source TEXT, first_url TEXT, last_seen TEXT, source TEXT, url TEXT," +
		" ttl INTEGER)"

	return baseSQLite(t,
		tablaSQLite{
			nombre:    "schema_version",
			sentencia: sentenciaDeSchemaVersion,
			filas:     []filaSQLite{{rowid: 1, valores: []any{nil, "2026-09-01T00:00:00Z"}}},
		},
		tablaSQLite{
			nombre:    "nodes",
			sentencia: "CREATE TABLE nodes (id TEXT, type TEXT, props TEXT, " + observaciones,
			filas: []filaSQLite{{rowid: 1, valores: []any{
				idDeLaNorma, grafo.TipoNorma, `{"identificador":"` + identificadorDeLaNorma + `"}`,
				"ayer", fuenteDeLaNorma, urlDeLaNorma, "ayer", fuenteDeLaNorma, urlDeLaNorma, nil,
			}}},
		},
		tablaSQLite{nombre: "edges", sentencia: "CREATE TABLE edges (src TEXT, rel TEXT, dst TEXT, " + observaciones},
	)
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
