package app_test

import (
	"bytes"
	"encoding/json"
	"io"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/app/testdata/ejemplo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Los datos de construcción con los que se invoca a la raíz de composición. Este
// test no mira el verbo reservado «version»: lo que importa es que los reciba y
// no los interprete.
const (
	versionDelEjemplo = "1.2.3"
	commitDelEjemplo  = "0123456789abcdef"
	fechaDelEjemplo   = "2026-01-01T00:00:00Z"
)

// borrador es el del esquema que emite --describe (FR-046).
const borrador = "https://json-schema.org/draft/2020-12/schema"

// clavesDelSobre son las seis claves exactas del nivel superior del sobre, ni
// una más ni una menos (FR-010).
var clavesDelSobre = []string{"ok", "fuente", "url", "fecha_consulta", "hash", "data"}

// clavesDeProcedencia son las cuatro filas con las que empieza la tabla mínima y
// sin las cuales no hay cita, en el orden en que se presentan (FR-041, FR-043).
var clavesDeProcedencia = []string{"fuente", "url", "fecha_consulta", "hash"}

// clavesCompartidas son las tres de la procedencia que dos invocaciones iguales
// tienen que escribir idénticas, sea cual sea la forma de presentación. La cuarta
// —la fecha de consulta— queda fuera porque es el instante de cada invocación, y
// la huella entra porque se calcula sobre el contenido y no sobre el instante
// (FR-011).
var clavesCompartidas = []string{"fuente", "url", "hash"}

// bandera es una de las ocho globales y la forma en que se escribe en una
// invocación.
type bandera struct {
	nombre  string
	escrita []string
}

// banderasGlobales son las ocho del kernel, escritas aquí **otra vez** y no
// tomadas de internal/cli a propósito: una lista derivada de la implementación
// seguiría cuadrando si el kernel dejara de ofrecer una de ellas, que es
// justamente lo que este test tiene que detectar (FR-018, SC-010).
var banderasGlobales = []bandera{
	{nombre: "--json", escrita: []string{"--json"}},
	{nombre: "--timeout", escrita: []string{"--timeout", "5s"}},
	{nombre: "--offline", escrita: []string{"--offline"}},
	{nombre: "--dry-run", escrita: []string{"--dry-run"}},
	{nombre: "--describe", escrita: []string{"--describe"}},
	{nombre: "--no-graph", escrita: []string{"--no-graph"}},
	{nombre: "--asunto", escrita: []string{"--asunto", "demo"}},
	{nombre: "--verbose", escrita: []string{"--verbose"}},
}

// ejemplar es lo único que este test necesita saber de un applet de ejemplo para
// someterlo a la **misma** tabla que al otro: con qué verbo se le ejerce, qué
// argumento acepta, con qué procedencia firma y qué claves lleva su `data`.
//
// No hay ningún campo para las banderas, ni para el sobre, ni para la huella, ni
// para los códigos de salida, ni para las formas de presentación: nada de eso
// distingue a un applet de otro porque nada de eso lo declara ninguno (SC-010).
type ejemplar struct {
	applet     app.Applet
	verbo      string
	argumento  string
	fuente     string
	url        string
	clavesData []string
}

// invocacion es todo lo observable de una invocación completa del kernel.
type invocacion struct {
	codigo  int
	salida  string
	errores string
}

// fila es un par ruta/valor de la tabla mínima, tal y como se lee de la salida.
type fila struct {
	ruta  string
	valor string
}

// TestAppletHereda es la comprobación de SC-010: los **dos** applets de ejemplo
// pasan la misma tabla, y el segundo no declara nada más que su nombre, sus
// verbos y el contenido de su `data`.
//
// Es el test que invoca por su nombre el escenario 7 de quickstart.md. Que exista
// no es un detalle: `go test -run` sobre un test que no existe termina en 0, así
// que sin este fichero el escenario daría por demostrado lo que nadie comprobó.
func TestAppletHereda(t *testing.T) {
	t.Parallel()

	for _, caso := range ejemplares(t) {
		t.Run(caso.applet.Nombre(), func(t *testing.T) {
			t.Parallel()

			t.Run("no declara nada más que nombre, verbos y el contenido de data",
				func(t *testing.T) {
					t.Parallel()
					exigirQueNoDeclaraNadaMas(t, caso)
				})

			t.Run("acepta las ocho banderas globales sin declarar ninguna",
				func(t *testing.T) {
					t.Parallel()
					exigirLasOchoBanderas(t, caso)
				})

			t.Run("emite el sobre con las seis claves", func(t *testing.T) {
				t.Parallel()
				exigirElSobre(t, caso)
			})

			t.Run("la misma tabla de códigos de salida", func(t *testing.T) {
				t.Parallel()
				exigirLaTablaDeCodigos(t, caso)
			})

			t.Run("responde a --describe y a --help", func(t *testing.T) {
				t.Parallel()
				exigirDescribeYAyuda(t, caso)
			})

			t.Run("las dos formas de presentación, idénticas en estructura",
				func(t *testing.T) {
					t.Parallel()
					exigirLasDosPresentaciones(t, caso)
				})
		})
	}
}

// ejemplares empareja cada applet del registro de ejemplo con lo que se espera
// de él. El emparejamiento va por nombre y no por posición, de modo que reordenar
// el registro no cambie en silencio qué se comprueba de cuál.
func ejemplares(t *testing.T) []ejemplar {
	t.Helper()

	esperado := map[string]ejemplar{
		"echo": {
			verbo:      "repetir",
			argumento:  "hola",
			fuente:     "kitlegal.echo",
			url:        "kitlegal:applet/echo",
			clavesData: []string{"mensaje"},
		},
		"contar": {
			verbo:      "letras",
			argumento:  "hola",
			fuente:     "kitlegal.contar",
			url:        "kitlegal:applet/contar",
			clavesData: []string{"texto", "total"},
		},
	}

	applets := ejemplo.Applets()
	require.Len(t, applets, 2,
		"hacen falta dos applets: con uno solo no hay nada que comparar y SC-010 no se demuestra")

	lista := make([]ejemplar, 0, len(applets))

	for _, applet := range applets {
		caso, hay := esperado[applet.Nombre()]
		require.Truef(t, hay, "el applet %q no tiene caso en esta tabla", applet.Nombre())

		caso.applet = applet
		lista = append(lista, caso)
	}

	return lista
}

// exigirQueNoDeclaraNadaMas comprueba la premisa sin la cual el resto de la tabla
// no probaría nada: que el applet declara exactamente los tres métodos del
// contrato y ningún estado propio, y que ninguno de sus verbos declara un campo
// que se llame como una de las ocho banderas globales. Si declarara una,
// aceptarla no sería heredarla (SC-010).
func exigirQueNoDeclaraNadaMas(t *testing.T, caso ejemplar) {
	t.Helper()

	tipo := reflect.TypeOf(caso.applet)

	assert.Equal(t, reflect.Struct, tipo.Kind())
	assert.Zero(t, tipo.NumField(),
		"un applet no guarda estado propio: lo que declara son sus tres métodos")
	assert.Equal(t, reflect.TypeFor[app.Applet]().NumMethod(), tipo.NumMethod(),
		"el applet no declara ningún método más que los tres del contrato")

	globales := make([]string, 0, len(banderasGlobales))
	for _, b := range banderasGlobales {
		globales = append(globales, normalizar(b.nombre))
	}

	for _, verbo := range caso.applet.Verbos() {
		require.NotNilf(t, verbo.Argumentos, "el verbo %q no declara fábrica", verbo.Nombre)

		argumentos := reflect.TypeOf(verbo.Argumentos()).Elem()
		for i := range argumentos.NumField() {
			campo := argumentos.Field(i)
			assert.NotContainsf(t, globales, normalizar(nombreDeBandera(campo)),
				"el verbo %q declara una bandera global en el campo %q",
				verbo.Nombre, campo.Name)
		}
	}
}

// nombreDeBandera es el nombre con el que el analizador expondría un campo: el de
// la etiqueta `name` si la lleva, y si no el del propio campo.
func nombreDeBandera(campo reflect.StructField) string {
	if nombre, hay := campo.Tag.Lookup("name"); hay {
		return nombre
	}

	return campo.Name
}

// normalizar lleva un nombre de bandera a la forma en que se comparan aquí: sin
// guiones y en minúsculas, de modo que «--no-graph», «no-graph» y «SinGrafo» no
// se escapen de la comparación por cómo se escriben.
func normalizar(nombre string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimPrefix(nombre, "--"), "-", ""))
}

// exigirLasOchoBanderas comprueba que el applet acepta las ocho, una por una y
// las seis combinables a la vez, sin haber declarado ninguna (FR-018, SC-010).
func exigirLasOchoBanderas(t *testing.T, caso ejemplar) {
	t.Helper()

	for _, b := range banderasGlobales {
		res := invocarEjemplo(t, caso.invocacionCon(b.escrita...)...)

		assert.Equalf(t, 0, res.codigo, "el applet acepta %s sin declararla: %s",
			b.nombre, res.errores)
	}

	// Las seis que no excluyen la ejecución, juntas y en la misma invocación:
	// --describe y --dry-run se quedan fuera porque no ejecutan nada, y lo que
	// aquí se comprueba es que el applet sigue trabajando con las demás puestas.
	res := invocarEjemplo(t, caso.invocacionCon(
		"--json", "--timeout", "5s", "--offline", "--no-graph", "--asunto", "demo", "--verbose",
	)...)

	assert.Equal(t, 0, res.codigo, res.errores)
	assert.Equal(t, caso.fuente, sobreDelDocumento(t, res.salida)["fuente"])
}

// exigirElSobre comprueba que el applet emite el sobre de seis claves con la
// huella y la fecha de consulta que el kernel monta, sin haber escrito ninguna de
// las seis (FR-010, FR-011, FR-044).
func exigirElSobre(t *testing.T, caso ejemplar) {
	t.Helper()

	res := invocarEjemplo(t, caso.invocacionCon("--json")...)

	require.Equal(t, 0, res.codigo, res.errores)
	assert.Empty(t, res.errores,
		"una invocación correcta no tiene nada que decir en la salida de error")

	sobre := sobreDelDocumento(t, res.salida)

	assert.ElementsMatch(t, clavesDelSobre, slices.Collect(maps.Keys(sobre)))
	assert.Equal(t, true, sobre["ok"])
	assert.Equal(t, caso.fuente, sobre["fuente"])
	assert.Equal(t, caso.url, sobre["url"])
	assert.NotEmpty(t, sobre["fecha_consulta"])
	assert.Regexp(t, `^sha256:[0-9a-f]{64}$`, sobre["hash"])

	datos, esObjeto := sobre["data"].(map[string]any)
	require.True(t, esObjeto, "el contenido del applet es un objeto")
	assert.ElementsMatch(t, caso.clavesData, slices.Collect(maps.Keys(datos)))
}

// exigirLaTablaDeCodigos somete a los dos applets a la misma tabla: las mismas
// invocaciones terminan con el mismo código, y ninguno de los dos declara uno
// solo (FR-030, SC-010, contracts/banderas-y-exit-codes.md §4).
func exigirLaTablaDeCodigos(t *testing.T, caso ejemplar) {
	t.Helper()

	casos := []struct {
		nombre string
		argv   []string
		codigo int
	}{
		{"la invocación correcta", caso.invocacionCon(), 0},
		{"--describe, que excluye la ejecución", caso.invocacionCon("--describe"), 0},
		{"--dry-run, que no ejecuta nada", caso.invocacionCon("--dry-run"), 0},
		{"la ayuda del applet", []string{caso.applet.Nombre(), "--help"}, 0},
		{"la ayuda del verbo", []string{caso.applet.Nombre(), caso.verbo, "--help"}, 0},
		{"una bandera desconocida", caso.invocacionCon("--jsno"), 2},
		{"un valor con formato inválido", caso.invocacionCon("--timeout", "abc"), 2},
		{"un plazo que no es positivo", caso.invocacionCon("--timeout", "0s"), 2},
		{"una bandera global sin su valor", caso.invocacionCon("--asunto"), 2},
	}

	for _, fila := range casos {
		res := invocarEjemplo(t, fila.argv...)

		assert.Equalf(t, fila.codigo, res.codigo, "%s: %s", fila.nombre, res.errores)
	}
}

// exigirDescribeYAyuda comprueba las dos formas en que un applet se explica sin
// haber escrito ninguna de las dos: el esquema de --describe, que nombra el verbo
// descrito y no ejecuta nada, y la ayuda, derivada del registro y del catálogo de
// verbos (FR-026, FR-046, FR-047, FR-049).
func exigirDescribeYAyuda(t *testing.T, caso ejemplar) {
	t.Helper()

	nombre := caso.applet.Nombre()

	descrito := invocarEjemplo(t, caso.invocacionCon("--describe")...)
	require.Equal(t, 0, descrito.codigo, descrito.errores)
	assert.Empty(t, descrito.errores)
	assert.NotContains(t, descrito.salida, schema.PrefijoHuella,
		"describirse y actuar son excluyentes: no hay sobre que firmar")

	var esquema map[string]any
	require.NoError(t, json.Unmarshal([]byte(descrito.salida), &esquema))
	assert.Equal(t, borrador, esquema["$schema"])
	assert.Equal(t, nombre+" "+caso.verbo, esquema["title"])

	propiedades, esObjeto := esquema["properties"].(map[string]any)
	require.True(t, esObjeto, "el esquema declara sus propiedades")
	assert.ElementsMatch(t, []string{"entrada", "salida"},
		slices.Collect(maps.Keys(propiedades)))

	ayuda := invocarEjemplo(t, nombre, "--help")
	require.Equal(t, 0, ayuda.codigo, ayuda.errores)
	assert.Empty(t, ayuda.errores)
	assert.Contains(t, ayuda.salida, "uso: "+nombre+" <verbo>")

	for _, verbo := range caso.applet.Verbos() {
		assert.Contains(t, ayuda.salida, verbo.Nombre, "la ayuda sale del catálogo de verbos")
	}

	// La ayuda del verbo enumera las ocho banderas que el applet hereda: es donde
	// quien invoca las ve por primera vez sin haber leído ningún manual.
	delVerbo := invocarEjemplo(t, nombre, caso.verbo, "--help")
	require.Equal(t, 0, delVerbo.codigo, delVerbo.errores)

	for _, b := range banderasGlobales {
		assert.Contains(t, delVerbo.salida, b.nombre)
	}
}

// exigirLasDosPresentaciones comprueba que las dos formas de salida presentan lo
// mismo: la tabla mínima lleva las cuatro filas de procedencia y después el
// contenido, exactamente las claves que el sobre en JSON trae dentro de `data`, y
// los dos citan la misma fuente y la misma huella —que se calcula sobre el
// contenido y no sobre cómo se presenta— (FR-011, FR-019, FR-041, FR-043).
func exigirLasDosPresentaciones(t *testing.T, caso ejemplar) {
	t.Helper()

	enJSON := invocarEjemplo(t, caso.invocacionCon("--json")...)
	enTabla := invocarEjemplo(t, caso.invocacionCon()...)

	require.Equal(t, 0, enJSON.codigo, enJSON.errores)
	require.Equal(t, 0, enTabla.codigo, enTabla.errores)

	sobre := sobreDelDocumento(t, enJSON.salida)
	filas := filasDeLaTabla(t, enTabla.salida)

	estructura := make([]string, 0, len(filas))
	valores := make(map[string]string, len(filas))

	for _, f := range filas {
		estructura = append(estructura, f.ruta)
		valores[f.ruta] = f.valor
	}

	// Las claves del contenido se presentan en orden alfabético, el mismo con el
	// que el dominio ordena la forma canónica de la que sale la huella.
	esperada := append(slices.Clone(clavesDeProcedencia),
		slices.Sorted(slices.Values(caso.clavesData))...)

	assert.Equal(t, esperada, estructura,
		"las dos formas presentan lo mismo y en el mismo orden")

	for _, clave := range clavesCompartidas {
		assert.Equalf(t, sobre[clave], valores[clave],
			"la forma legible y la legible por máquina citan lo mismo en %q", clave)
	}

	// La fecha de consulta es lo único que dos invocaciones no comparten: es el
	// instante de cada una. Lo que sí comparten es el formato, y que las dos lo
	// escriban igual es lo que permite comparar dos citas hechas por vías
	// distintas (contracts/sobre-de-salida.md §2).
	for _, fecha := range []any{sobre["fecha_consulta"], valores["fecha_consulta"]} {
		texto, esTexto := fecha.(string)
		require.True(t, esTexto, "la fecha de consulta se presenta como texto")

		_, err := time.Parse(time.RFC3339Nano, texto)
		assert.NoErrorf(t, err, "la fecha %q no está en el formato del sobre", texto)
	}
}

// invocacionCon es la invocación mínima del applet —su nombre, su verbo y su
// argumento— seguida de lo que el caso añada. Escribirla una sola vez es lo que
// permite que los dos applets pasen por la misma tabla sin que ninguna fila se
// escriba dos veces.
func (e ejemplar) invocacionCon(extra ...string) []string {
	argv := []string{e.applet.Nombre(), e.verbo, e.argumento}

	return append(argv, extra...)
}

// invocarEjemplo ejecuta la raíz de composición entera con el registro de los dos
// applets de ejemplo —el mismo mecanismo de registro que el binario que se
// publica (FR-001)— contra dos buffers.
//
// Que esta función retorne y el test siga vivo después es además la comprobación
// de que ninguna ruta llama a os.Exit (FR-035).
func invocarEjemplo(t *testing.T, argv ...string) invocacion {
	t.Helper()

	registro, err := ejemplo.Registro()
	require.NoError(t, err, "el registro de ejemplo es válido por construcción")

	var salida, errores bytes.Buffer

	codigo := app.Main(
		append([]string{nombreDelBinario}, argv...), registro, &salida, &errores,
		versionDelEjemplo, commitDelEjemplo, fechaDelEjemplo)

	return invocacion{
		codigo:  codigo,
		salida:  salida.String(),
		errores: errores.String(),
	}
}

// sobreDelDocumento analiza la salida estándar de una invocación con --json.
// Comprueba de paso que lleva **un único** documento y nada más (FR-042, SC-002).
func sobreDelDocumento(t *testing.T, salida string) map[string]any {
	t.Helper()

	decodificador := json.NewDecoder(strings.NewReader(salida))

	var sobre map[string]any

	require.NoError(t, decodificador.Decode(&sobre))
	require.ErrorIs(t, decodificador.Decode(new(json.RawMessage)), io.EOF,
		"con --json la salida estándar lleva un único documento JSON y nada más")

	return sobre
}

// filasDeLaTabla lee la tabla mínima como lo que es: un par ruta/valor por línea,
// separados por el relleno de la columna.
func filasDeLaTabla(t *testing.T, salida string) []fila {
	t.Helper()

	require.NotEmpty(t, salida, "la tabla mínima no puede estar vacía")

	lineas := strings.Split(strings.TrimRight(salida, "\n"), "\n")
	filas := make([]fila, 0, len(lineas))

	for _, linea := range lineas {
		ruta, valor, hay := strings.Cut(linea, "  ")
		require.Truef(t, hay, "la fila %q no separa la ruta del valor", linea)

		filas = append(filas, fila{ruta: ruta, valor: strings.TrimSpace(valor)})
	}

	return filas
}
