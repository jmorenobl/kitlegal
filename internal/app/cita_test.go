package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/cita"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// actualizarGoldenDeCita es la bandera con la que TestCitaPreparar y
// TestCitaCotejar escriben, antes de compararlos, los golden de testdata/cita/
// desde lo que da cada verbo: nunca los escribe una mano
// (contracts/applet-cita.md §9 de H23). Solo la usa la tarea [datos] que los
// crea o los cambia, y lo escrito lo revisa una persona; make ci nunca la pasa.
var actualizarGoldenDeCita = flag.Bool("actualizar-golden-de-cita", false,
	"escribe en testdata/cita/ los golden de los dos verbos de cita desde su salida antes de compararlos")

// Lo que estos tests esperan del sobre de cita, escrito aquí otra vez y no
// tomado del código que se prueba (contracts/applet-cita.md §2 de H23).
const (
	// fuenteDeCitaEsperada es la del espacio reservado a lo que el binario
	// calcula: dice que no se ha consultado ninguna fuente.
	fuenteDeCitaEsperada = "kitlegal.cita"
	// urlDeCitaEsperada es la del sobre que no tiene nada que abrir, y la de
	// todo fallo.
	urlDeCitaEsperada = "kitlegal:applet/cita"
	// prefijoDelDocumentoEsperado va delante de la huella del texto recibido en
	// la url del sobre de cotejar.
	prefijoDelDocumentoEsperado = "kitlegal:documento/sha256:"
	// urlDelBuscadorEsperada es la dirección que abre quien busca una
	// referencia.
	urlDelBuscadorEsperada = "https://www.poderjudicial.es/search/indexAN.jsp"
	// urlDeLaBusquedaEsperada es la que abre el buscador con la búsqueda de
	// «cláusula suelo» ya hecha: la tilde, en sus dos octetos, y el espacio.
	urlDeLaBusquedaEsperada = "https://www.poderjudicial.es/search/sentencias/cl%C3%A1usula%20suelo/1/AN"
)

// La evidencia del ADR 0036 que leen estos tests, sin copiarla, y lo que su
// manifiesto dice de ella.
const (
	// rutaDelFragmento es la del fragmento del documento del CENDOJ, relativa al
	// directorio de este paquete.
	rutaDelFragmento = "../../evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt"
	// huellaDelFragmentoEsperada es su SHA-256 en evidencias/adr-0036/manifiesto.json.
	huellaDelFragmentoEsperada = "4886e0c8ca9836527ec08d8732b50315640a76803033374af5287b2a8321ee27"
	// lineasDeLaFichaDelFragmento son las de su ficha, de «Roj:» a «Tipo de
	// Resolución:»: lo que la skill pasa.
	lineasDeLaFichaDelFragmento = 11
)

// Los golden (contracts/applet-cita.md §9 de H23).
const (
	// carpetaDeLosGoldenDeCita es la suya, relativa al directorio de este
	// paquete.
	carpetaDeLosGoldenDeCita = "testdata/cita"
	// fechaDeLosGolden es la fecha_consulta con la que se guardan, que es lo
	// único del sobre que cambia de una invocación a otra.
	fechaDeLosGolden = `"fecha_consulta":"0001-01-01T00:00:00Z"`
)

// Los dos verbos y las referencias de estas tablas, que son las del spec: la
// sentencia del fragmento por sus tres formas y el ECLI del Tribunal
// Constitucional.
const (
	preparar = "preparar"
	cotejar  = "cotejar"

	ecliDelFragmento   = "ECLI:ES:TS:2023:3144"
	rojDelFragmento    = "STS 3144/2023"
	numeroDelFragmento = "1088/2023"
	fechaDelFragmento  = "2023-07-04"
)

// procedenciaDeUnFalloDeCita es la firma de todo fallo del applet.
var procedenciaDeUnFalloDeCita = schema.Procedencia{Fuente: fuenteDeCitaEsperada, URL: urlDeCitaEsperada}

// fragmentoDeLaEvidencia es el fragmento del documento del CENDOJ de la
// evidencia del ADR 0036, leído de su fichero. Su huella es la del manifiesto:
// si el fichero dejara de ser el fragmento, byte a byte, todo test que lo usa
// fallaría aquí.
func fragmentoDeLaEvidencia(t *testing.T) string {
	t.Helper()

	contenido, err := os.ReadFile(rutaDelFragmento)
	require.NoError(t, err, "el fragmento se lee de la evidencia del ADR 0036")
	require.Equal(t, huellaDelFragmentoEsperada, huellaDe(string(contenido)),
		"el fichero no es el fragmento del manifiesto de la evidencia")

	return string(contenido)
}

// fichaDeLaEvidencia es la ficha sola: las once primeras líneas del fragmento.
func fichaDeLaEvidencia(t *testing.T) string {
	t.Helper()

	lineas := strings.SplitAfter(fragmentoDeLaEvidencia(t), "\n")
	require.Greater(t, len(lineas), lineasDeLaFichaDelFragmento)

	return strings.Join(lineas[:lineasDeLaFichaDelFragmento], "")
}

// sinLaLinea devuelve el texto sin la línea que empieza así, que tiene que
// estar: quitar una que no está no probaría nada.
func sinLaLinea(t *testing.T, texto, comienzo string) string {
	t.Helper()

	lineas := strings.SplitAfter(texto, "\n")
	i := slices.IndexFunc(lineas, func(linea string) bool { return strings.HasPrefix(linea, comienzo) })
	require.NotEqual(t, -1, i, "el texto lleva una línea que empieza por %q", comienzo)

	return strings.Join(slices.Delete(lineas, i, i+1), "")
}

// huellaDe es el SHA-256 del texto, byte a byte, en hexadecimal.
func huellaDe(texto string) string {
	huella := sha256.Sum256([]byte(texto))

	return hex.EncodeToString(huella[:])
}

// registroDeCita es el registro de estos tests: el applet cita, compuesto como
// lo componen las dos raíces, y la entrada estándar que se le dé, o ninguna con
// nil. Es un registro local, sin almacén al que entregar.
func registroDeCita(t *testing.T, entrada io.Reader) *Registro {
	t.Helper()

	var registro Registro

	require.NoError(t, registro.Registrar(AppletCita()))
	registro.LeerDe(entrada)

	return &registro
}

// ordenDeCita ejecuta un verbo de cita como orden, con --json, con ese texto en
// la entrada estándar, y devuelve además la entrada, que dice si alguien leyó
// de ella.
func ordenDeCita(t *testing.T, verbo, enLaEntrada string, argumentos ...string) (invocacionDePrueba, *entradaEspia) {
	t.Helper()

	entrada := &entradaEspia{texto: strings.NewReader(enLaEntrada)}
	argv := slices.Concat([]string{"kitlegal", "cita", verbo}, argumentos, []string{"--json"})

	return invocar(t, registroDeCita(t, entrada), argv...), entrada
}

// llamarACita llama a la herramienta de un verbo de cita con esos argumentos,
// por el camino del servidor: la línea de órdenes que compone cada llamada. El
// registro tiene registrada una entrada estándar con un documento entero, y la
// llamada no lee de ella: en el servidor, esa entrada es el protocolo (H23
// FR-020, FR-026). Devuelve el sobre como lo escribe la orden, con su salto
// final, y si es un error de herramienta.
func llamarACita(t *testing.T, verbo string, argumentos map[string]string) (sobre string, fallo bool) {
	t.Helper()

	entrada := &entradaEspia{texto: strings.NewReader(fragmentoDeLaEvidencia(t))}

	herramientas, err := herramientasDe(registroDeCita(t, entrada), schema.Contexto{Timeout: timeoutDelContrato},
		slog.New(slog.DiscardHandler), io.Discard)
	require.NoError(t, err)

	// Un objeto también sin argumentos: un mapa nulo se escribiría `null`.
	propiedades, err := json.Marshal(maps.Collect(maps.All(argumentos)))
	require.NoError(t, err)

	for _, herramienta := range herramientas {
		if herramienta.Nombre != "cita_"+verbo {
			continue
		}

		resultado := herramienta.Llamar(propiedades)

		assert.Zero(t, entrada.lecturas, "una llamada de herramienta no lee nunca la entrada estándar")

		return string(resultado.Sobre) + "\n", resultado.Fallo
	}

	require.Failf(t, "herramienta sin anunciar", "el registro no da la herramienta cita_%s", verbo)

	return "", false
}

// conLaFechaDeLosGolden es el sobre con su fecha_consulta cambiada por la de
// los golden: lo que queda es lo que no cambia de una invocación a otra.
func conLaFechaDeLosGolden(t *testing.T, sobre string) string {
	t.Helper()

	require.Len(t, fechaDeConsultaDelSobre.FindAllString(sobre, -1), 1, "el sobre lleva una fecha_consulta")

	return fechaDeConsultaDelSobre.ReplaceAllString(sobre, fechaDeLosGolden)
}

// esquemaPublicadoDeCita es la salida del verbo en schemas/cita.json.
func esquemaPublicadoDeCita(t *testing.T, verbo string) *jsonschema.Schema {
	t.Helper()

	publicado, id := ficheroPublicadoDelVerbo(t, verbo)
	require.Equal(t, raizDeLosEsquemas+"cita.json", id, "la parte de %s está en schemas/cita.json", verbo)

	return salidaPublicada(t, publicado, id, verbo)
}

// exigirSalidaDeCita exige que el sobre valide contra la salida publicada de su
// verbo y que la validación restrinja data: el mismo sobre con una clave de más
// en data no valida. No se quita ninguna de las suyas, como a las de boe: en
// las de cita, lo que no aplica no está, y varias son opcionales.
func exigirSalidaDeCita(t *testing.T, esquema *jsonschema.Schema, salida string) {
	t.Helper()

	require.NoError(t, esquema.Validate(sobreValidable(t, salida)), "el sobre valida contra su parte de schemas/cita.json")

	conClaveDeMas := sobreValidable(t, salida)
	objetoDeData(t, conClaveDeMas)["ajena"] = "no declarada"
	require.Error(t, esquema.Validate(conClaveDeMas), "data con una clave de más no valida")
}

// compruebaSobreDeCita exige del sobre de una salida correcta lo que fija
// FR-004 de H23: sus seis claves, ok verdadero, la fuente del espacio reservado
// a lo calculado, esa url y la fecha del reloj del kernel, que es la del
// instante de la invocación y no una que el applet declare. Devuelve el sobre.
func compruebaSobreDeCita(t *testing.T, salida, url string, antes, despues time.Time) map[string]any {
	t.Helper()

	sobre := sobreDelJSON(t, salida)

	assert.Equal(t, true, sobre["ok"])
	assert.Equal(t, fuenteDeCitaEsperada, sobre["fuente"], "la fuente dice que no se ha consultado ninguna")
	assert.Equal(t, url, sobre["url"])
	assert.WithinRange(t, fechaDelSobre(t, sobre), antes, despues, "el sobre lo fecha el reloj del kernel")

	return sobre
}

// goldenDeCita es una fila de contracts/applet-cita.md §9 de H23: el fichero, la
// invocación que lo da —los de cotejar, con el fragmento en la entrada
// estándar— y la url de su sobre.
type goldenDeCita struct {
	fichero string
	verbo   string
	orden   []string
	url     string
}

// goldenDeCitaDelContrato son los once, en el orden del contrato.
func goldenDeCitaDelContrato() []goldenDeCita {
	delDocumento := prefijoDelDocumentoEsperado + huellaDelFragmentoEsperada

	return []goldenDeCita{
		{fichero: "preparar-ecli.json", verbo: preparar, url: urlDelBuscadorEsperada, orden: []string{ecliDelFragmento}},
		{fichero: "preparar-roj.json", verbo: preparar, url: urlDelBuscadorEsperada, orden: []string{"--roj", rojDelFragmento}},
		{
			fichero: "preparar-resolucion.json", verbo: preparar, url: urlDelBuscadorEsperada,
			orden: []string{"--resolucion", numeroDelFragmento, "--fecha", fechaDelFragmento},
		},
		{
			fichero: "preparar-texto.json", verbo: preparar, url: urlDeLaBusquedaEsperada,
			orden: []string{"--texto", "cláusula suelo"},
		},
		{fichero: "cotejar-sin-referencia.json", verbo: cotejar, url: delDocumento},
		{fichero: "cotejar-ecli.json", verbo: cotejar, url: delDocumento, orden: []string{ecliDelFragmento}},
		{fichero: "cotejar-roj.json", verbo: cotejar, url: delDocumento, orden: []string{"--roj", rojDelFragmento}},
		{
			fichero: "cotejar-resolucion.json", verbo: cotejar, url: delDocumento,
			orden: []string{"--resolucion", numeroDelFragmento, "--fecha", fechaDelFragmento},
		},
		{fichero: "cotejar-roj-cruzado.json", verbo: cotejar, url: delDocumento, orden: []string{"--roj", "STS 1088/2023"}},
		{
			fichero: "cotejar-otra-fecha.json", verbo: cotejar, url: delDocumento,
			orden: []string{"--resolucion", numeroDelFragmento, "--fecha", "2023-01-01"},
		},
		{
			fichero: "cotejar-resolucion-cruzada.json", verbo: cotejar, url: delDocumento,
			orden: []string{"--resolucion", "3144/2023", "--fecha", fechaDelFragmento},
		},
	}
}

// goldenDelVerbo son los del verbo, que tienen que ser tantos como dice el
// contrato y los únicos de ese verbo en la carpeta: uno de más sería un golden
// que nadie compara.
func goldenDelVerbo(t *testing.T, verbo string, cuantos int) []goldenDeCita {
	t.Helper()

	var (
		delVerbo []goldenDeCita
		ficheros []string
	)

	for _, golden := range goldenDeCitaDelContrato() {
		if golden.verbo == verbo {
			delVerbo = append(delVerbo, golden)
			ficheros = append(ficheros, filepath.Join(carpetaDeLosGoldenDeCita, golden.fichero))
		}
	}

	require.Len(t, delVerbo, cuantos, "los golden de %s del contrato", verbo)

	if !*actualizarGoldenDeCita {
		enLaCarpeta, err := filepath.Glob(filepath.Join(carpetaDeLosGoldenDeCita, verbo+"-*.json"))
		require.NoError(t, err)
		assert.ElementsMatch(t, ficheros, enLaCarpeta, "los golden de %s en %s", verbo, carpetaDeLosGoldenDeCita)
	}

	return delVerbo
}

// comprueba ejecuta la invocación del golden, con ese texto en la entrada
// estándar, y exige: código 0 y el sobre de FR-004; que valide contra su
// esquema publicado; y que, con la fecha de los golden, sea el fichero byte a
// byte, que valida también (H23 FR-080, FR-081; SC-004). Con
// -actualizar-golden-de-cita escribe antes el fichero desde la salida.
func (golden goldenDeCita) comprueba(t *testing.T, esquema *jsonschema.Schema, enLaEntrada string) {
	t.Helper()

	antes := time.Now()
	res, _ := ordenDeCita(t, golden.verbo, enLaEntrada, golden.orden...)
	despues := time.Now()

	require.Equal(t, 0, res.codigo, res.errores)
	assert.Empty(t, res.errores, "una salida correcta no escribe nada en la salida de error")

	compruebaSobreDeCita(t, res.salida, golden.url, antes, despues)
	exigirSalidaDeCita(t, esquema, res.salida)

	emitido := conLaFechaDeLosGolden(t, res.salida)
	ruta := filepath.Join(carpetaDeLosGoldenDeCita, golden.fichero)

	if *actualizarGoldenDeCita {
		require.NoError(t, os.MkdirAll(carpetaDeLosGoldenDeCita, 0o750))
		require.NoError(t, os.WriteFile(ruta, []byte(emitido), 0o600))
	}

	guardado, err := os.ReadFile(filepath.Clean(ruta))
	require.NoError(t, err, "el golden lo escribe el verbo, con -actualizar-golden-de-cita")

	assert.Equal(t, string(guardado), emitido, "%s: el sobre, byte a byte salvo fecha_consulta", golden.fichero)
	exigirSalidaDeCita(t, esquema, string(guardado))
}

// rechazoDeCita es una fila de contracts/applet-cita.md §6 de H23, por sus dos
// caminos: la orden, con sus argumentos y lo que trae su entrada estándar, y la
// llamada de herramienta, con los suyos. Las dos terminan con el mismo error de
// argumentos, cuyo mensaje nombra lo que dice la fila.
type rechazoDeCita struct {
	nombre     string
	orden      []string
	enLaOrden  string
	llamada    map[string]string
	nombra     []string
	sinLeerla  bool
	noSeNombra []string
}

// comprueba exige de la orden el código 2, el sobre de fallo que firma cita con
// la clase argumentos y un mensaje que nombra lo que dice su fila, válido
// contra el esquema publicado del verbo; y de la llamada de herramienta, un
// error de herramienta con el sobre de la orden, byte a byte salvo
// fecha_consulta (H23 FR-006, FR-015, FR-026, FR-030).
func (rechazo rechazoDeCita) comprueba(t *testing.T, verbo string, esquema *jsonschema.Schema) {
	t.Helper()

	res, entrada := ordenDeCita(t, verbo, rechazo.enLaOrden, rechazo.orden...)

	exigirSobreDeFallo(t, res, schema.ClaseArgumentos, 2, procedenciaDeUnFalloDeCita)
	exigirSalidaDeCita(t, esquema, res.salida)

	mensaje, esTexto := datosDelSobre(t, sobreDelJSON(t, res.salida))["mensaje"].(string)
	require.True(t, esTexto, "el mensaje es un texto")

	for _, nombrado := range rechazo.nombra {
		assert.Contains(t, mensaje, nombrado)
	}

	for _, ajeno := range rechazo.noSeNombra {
		assert.NotContains(t, mensaje, ajeno)
	}

	assert.Contains(t, res.errores, mensaje, "el mensaje va también a la salida de error")

	if rechazo.sinLeerla {
		assert.Zero(t, entrada.lecturas, "la entrada estándar no se lee")
	}

	sobre, fallo := llamarACita(t, verbo, rechazo.llamada)

	assert.True(t, fallo, "un error de argumentos es un error de herramienta")
	assert.Equal(t, conLaFechaDeLosGolden(t, res.salida), conLaFechaDeLosGolden(t, sobre),
		"la llamada devuelve el sobre de su orden")
}

// rechazosDeLaReferencia son las filas de la referencia mal dada, que valen en
// los dos verbos (H23 FR-006): cada forma sin la suya, también vacía —un
// argumento escrito con valor vacío está dado—, el número y la fecha uno sin el
// otro, y más de una forma a la vez.
func rechazosDeLaReferencia() []rechazoDeCita {
	const (
		formaDelECLI   = "ECLI:ES:<órgano>:<año>:<número>"
		formaDelROJ    = "<siglas> <número>/<año>"
		formaDelNumero = "<número>/<año>"
		formaDeLaFecha = "AAAA-MM-DD"
		vacio          = `""`
	)

	return []rechazoDeCita{
		{
			nombre: "un ECLI mal formado", orden: []string{"ECLI:ES:TS:2023"},
			llamada: map[string]string{"ecli": "ECLI:ES:TS:2023"},
			nombra:  []string{`"ECLI:ES:TS:2023"`, formaDelECLI},
		},
		{
			nombre: "un ECLI en minúsculas", orden: []string{"ecli:es:ts:2023:3144"},
			llamada: map[string]string{"ecli": "ecli:es:ts:2023:3144"},
			nombra:  []string{`"ecli:es:ts:2023:3144"`, formaDelECLI},
		},
		{
			nombre: "un ECLI con un blanco", orden: []string{"ECLI:ES:TS:2023: 3144"},
			llamada: map[string]string{"ecli": "ECLI:ES:TS:2023: 3144"},
			nombra:  []string{`"ECLI:ES:TS:2023: 3144"`, "no es válido"},
		},
		{
			nombre: "un ECLI vacío", orden: []string{""},
			llamada: map[string]string{"ecli": ""},
			nombra:  []string{"ECLI " + vacio, formaDelECLI},
		},
		{
			nombre: "un ECLI de otro país", orden: []string{"ECLI:FR:CC:2023:1"},
			llamada: map[string]string{"ecli": "ECLI:FR:CC:2023:1"},
			nombra:  []string{`"ECLI:FR:CC:2023:1"`, "no es español"},
		},
		{
			nombre: "un ROJ sin su forma", orden: []string{"--roj", "STS3144/2023"},
			llamada: map[string]string{"roj": "STS3144/2023"},
			nombra:  []string{`"STS3144/2023"`, formaDelROJ},
		},
		{
			nombre: "un ROJ vacío", orden: []string{"--roj", ""},
			llamada: map[string]string{"roj": ""},
			nombra:  []string{"ROJ " + vacio, formaDelROJ},
		},
		{
			nombre: "un ROJ vacío, con el signo igual", orden: []string{"--roj="},
			llamada: map[string]string{"roj": ""},
			nombra:  []string{"ROJ " + vacio, formaDelROJ},
		},
		{
			nombre: "un número de resolución sin su forma", orden: []string{"--resolucion", "1088-2023", "--fecha", fechaDelFragmento},
			llamada: map[string]string{"resolucion": "1088-2023", "fecha": fechaDelFragmento},
			nombra:  []string{`"1088-2023"`, formaDelNumero},
		},
		{
			nombre: "un número de resolución vacío", orden: []string{"--resolucion", "", "--fecha", fechaDelFragmento},
			llamada: map[string]string{"resolucion": "", "fecha": fechaDelFragmento},
			nombra:  []string{"resolución " + vacio, formaDelNumero},
		},
		{
			nombre: "un número de resolución sin su fecha", orden: []string{"--resolucion", numeroDelFragmento},
			llamada: map[string]string{"resolucion": numeroDelFragmento},
			nombra:  []string{numeroDelFragmento, "necesita su fecha"},
		},
		{
			nombre: "una fecha sin su número", orden: []string{"--fecha", fechaDelFragmento},
			llamada: map[string]string{"fecha": fechaDelFragmento},
			nombra:  []string{`"` + fechaDelFragmento + `"`, "solo vale con un número de resolución"},
		},
		{
			nombre: "una fecha vacía sin su número", orden: []string{"--fecha", ""},
			llamada: map[string]string{"fecha": ""},
			nombra:  []string{"fecha " + vacio, "solo vale con un número de resolución"},
		},
		{
			nombre: "una fecha sin su forma", orden: []string{"--resolucion", numeroDelFragmento, "--fecha", "04/07/2023"},
			llamada: map[string]string{"resolucion": numeroDelFragmento, "fecha": "04/07/2023"},
			nombra:  []string{`"04/07/2023"`, formaDeLaFecha},
		},
		{
			nombre: "una fecha vacía", orden: []string{"--resolucion", numeroDelFragmento, "--fecha", ""},
			llamada: map[string]string{"resolucion": numeroDelFragmento, "fecha": ""},
			nombra:  []string{"fecha " + vacio, formaDeLaFecha},
		},
		{
			nombre: "una fecha que no es un día", orden: []string{"--resolucion", numeroDelFragmento, "--fecha", "2023-02-31"},
			llamada: map[string]string{"resolucion": numeroDelFragmento, "fecha": "2023-02-31"},
			nombra:  []string{`"2023-02-31"`, formaDeLaFecha, "un día que existe"},
		},
		{
			nombre: "más de una forma de referencia", orden: []string{ecliDelFragmento, "--roj", rojDelFragmento},
			llamada: map[string]string{"ecli": ecliDelFragmento, "roj": rojDelFragmento},
			nombra:  []string{"un ECLI y un ROJ"},
		},
		{
			nombre: "más de una forma de referencia, una de ellas vacía", orden: []string{ecliDelFragmento, "--roj", ""},
			llamada: map[string]string{"ecli": ecliDelFragmento, "roj": ""},
			nombra:  []string{"un ECLI y un ROJ"},
		},
		{
			nombre:  "las tres formas de referencia",
			orden:   []string{ecliDelFragmento, "--roj", rojDelFragmento, "--resolucion", numeroDelFragmento, "--fecha", fechaDelFragmento},
			llamada: map[string]string{"ecli": ecliDelFragmento, "roj": rojDelFragmento, "resolucion": numeroDelFragmento, "fecha": fechaDelFragmento},
			nombra:  []string{"un ECLI, un ROJ y un número de resolución"},
		},
	}
}

// rechazosDePreparar son las filas propias de cita preparar (H23 FR-015): ni
// una referencia ni un texto, los dos a la vez —también con el texto vacío— y
// un texto vacío o solo de blancos; y la referencia, que se comprueba antes que
// el texto.
func rechazosDePreparar() []rechazoDeCita {
	return []rechazoDeCita{
		{
			nombre: "ni una referencia ni un texto", llamada: map[string]string{},
			nombra: []string{"un ECLI", "--roj", "--resolucion", "--fecha", "--texto"},
		},
		{
			nombre: "una referencia junto a un texto", orden: []string{ecliDelFragmento, "--texto", "cláusula suelo"},
			llamada: map[string]string{"ecli": ecliDelFragmento, "texto": "cláusula suelo"},
			nombra:  []string{"una referencia", "--texto", "a la vez"},
		},
		{
			nombre: "una referencia junto a un texto vacío", orden: []string{ecliDelFragmento, "--texto", ""},
			llamada: map[string]string{"ecli": ecliDelFragmento, "texto": ""},
			nombra:  []string{"una referencia", "--texto", "a la vez"},
		},
		{
			nombre: "un número con su fecha junto a un texto", orden: []string{"--resolucion", numeroDelFragmento, "--fecha", fechaDelFragmento, "--texto", "cláusula suelo"},
			llamada: map[string]string{"resolucion": numeroDelFragmento, "fecha": fechaDelFragmento, "texto": "cláusula suelo"},
			nombra:  []string{"una referencia", "--texto", "a la vez"},
		},
		{
			nombre: "un texto vacío", orden: []string{"--texto", ""},
			llamada: map[string]string{"texto": ""},
			nombra:  []string{`búsqueda ""`, "vacío"},
		},
		{
			nombre: "un texto solo de blancos", orden: []string{"--texto", "  "},
			llamada: map[string]string{"texto": "  "},
			nombra:  []string{`búsqueda "  "`, "solo de blancos"},
		},
		{
			nombre: "una referencia mal dada junto a un texto: la referencia va antes", orden: []string{"ECLI:ES:TS:2023", "--texto", "cláusula suelo"},
			llamada:    map[string]string{"ecli": "ECLI:ES:TS:2023", "texto": "cláusula suelo"},
			nombra:     []string{`"ECLI:ES:TS:2023"`},
			noSeNombra: []string{"a la vez"},
		},
	}
}

// TestCitaPreparar fija el verbo preparar del applet cita, por el kernel entero
// (contracts/applet-cita.md §1 a §3, §6 y §9 de H23):
//
//   - sus cuatro golden son, byte a byte salvo fecha_consulta, lo que da el
//     verbo, y cada salida valida contra su parte de schemas/cita.json (FR-080;
//     SC-004), con el sobre de FR-004: la fuente de lo calculado, la dirección
//     que abre quien pregunta y la fecha del reloj del kernel;
//   - sin nada que abrir —el ECLI del Tribunal Constitucional—, la url es la del
//     applet, y data declara lo que queda fuera (FR-014);
//   - cada fila de errores de §6, también con el argumento escrito vacío, termina
//     con 2 y la clase argumentos, por la línea de órdenes y por la línea que
//     compone una llamada de herramienta, que devuelve el mismo sobre (FR-006,
//     FR-015, FR-030);
//   - no lee la entrada estándar, no devuelve operaciones de grafo y no cuenta
//     su salida de otra forma que el kernel (FR-003).
func TestCitaPreparar(t *testing.T) {
	t.Parallel()

	esquema := esquemaPublicadoDeCita(t, preparar)

	for _, golden := range goldenDelVerbo(t, preparar, 4) {
		t.Run("golden/"+golden.fichero, func(t *testing.T) {
			t.Parallel()

			golden.comprueba(t, esquema, fragmentoDeLaEvidencia(t))
		})
	}

	for _, rechazo := range slices.Concat(rechazosDeLaReferencia(), rechazosDePreparar()) {
		t.Run("rechazo/"+rechazo.nombre, func(t *testing.T) {
			t.Parallel()

			// Con un documento en la entrada: preparar no la lee nunca.
			rechazo.enLaOrden = fragmentoDeLaEvidencia(t)
			rechazo.sinLeerla = true

			rechazo.comprueba(t, preparar, esquema)
		})
	}

	t.Run("fuera de cobertura, la url es la del applet", func(t *testing.T) {
		t.Parallel()

		compruebaFueraDeCobertura(t, esquema)
	})

	t.Run("la llamada con resultado devuelve el sobre de su orden", func(t *testing.T) {
		t.Parallel()

		// Cada uno de sus cinco argumentos va en alguna: el de posición y los
		// cuatro que en la orden son banderas.
		compruebaLlamadasConResultado(t, preparar, "", []llamadaConResultado{
			{orden: []string{ecliDelFragmento}, llamada: map[string]string{"ecli": ecliDelFragmento}},
			{orden: []string{"--roj", rojDelFragmento}, llamada: map[string]string{"roj": rojDelFragmento}},
			{
				orden:   []string{"--resolucion", numeroDelFragmento, "--fecha", fechaDelFragmento},
				llamada: map[string]string{"resolucion": numeroDelFragmento, "fecha": fechaDelFragmento},
			},
			{orden: []string{"--texto", "cláusula suelo"}, llamada: map[string]string{"texto": "cláusula suelo"}},
		})
	})

	t.Run("no lee la entrada, no observa nada y no cuenta su salida", func(t *testing.T) {
		t.Parallel()

		res, entrada := ordenDeCita(t, preparar, fragmentoDeLaEvidencia(t), ecliDelFragmento)
		require.Equal(t, 0, res.codigo, res.errores)
		assert.Zero(t, entrada.lecturas, "preparar no lee la entrada estándar")

		ecli := ecliDelFragmento
		argumentos := &argumentosDePreparar{referenciaDeCita: referenciaDeCita{ECLI: &ecli}}

		compruebaResultadoCalculado(t, argumentos, cita.Consulta{})
	})
}

// compruebaFueraDeCobertura exige del ECLI del Tribunal Constitucional, que no
// es ninguno de los golden, el código 0 y un sobre sin nada que abrir: su url
// es la del applet, y data no lleva dirección y declara la cobertura (H23
// FR-004, FR-014).
func compruebaFueraDeCobertura(t *testing.T, esquema *jsonschema.Schema) {
	t.Helper()

	antes := time.Now()
	res, _ := ordenDeCita(t, preparar, "", "ECLI:ES:TC:2024:79")
	despues := time.Now()

	require.Equal(t, 0, res.codigo, res.errores)

	sobre := compruebaSobreDeCita(t, res.salida, urlDeCitaEsperada, antes, despues)
	exigirSalidaDeCita(t, esquema, res.salida)

	data := objetoDe(t, sobre["data"], "data")
	assert.NotContains(t, data, "direccion", "fuera de cobertura no hay nada que abrir")
	assert.Equal(t, "no-cubierto", objetoDe(t, data["cobertura"], "cobertura")["cendoj"])
}

// llamadaConResultado es una llamada de herramienta que termina bien, con la
// orden que da su mismo sobre.
type llamadaConResultado struct {
	orden   []string
	llamada map[string]string
}

// compruebaLlamadasConResultado exige que cada llamada a la herramienta del
// verbo devuelva un resultado, y no un error de herramienta, con el sobre de su
// orden, byte a byte salvo fecha_consulta, también con los argumentos que en la
// orden son banderas (H23 FR-030; contracts/applet-cita.md §7). enLaOrden es lo
// que trae la entrada estándar de la orden.
func compruebaLlamadasConResultado(t *testing.T, verbo, enLaOrden string, llamadas []llamadaConResultado) {
	t.Helper()

	for _, caso := range llamadas {
		res, _ := ordenDeCita(t, verbo, enLaOrden, caso.orden...)
		require.Equal(t, 0, res.codigo, res.errores)

		sobre, fallo := llamarACita(t, verbo, caso.llamada)

		assert.False(t, fallo, "%v", caso.orden)
		assert.Equal(t, conLaFechaDeLosGolden(t, res.salida), conLaFechaDeLosGolden(t, sobre), "%v", caso.orden)
	}
}

// compruebaResultadoCalculado ejecuta el verbo sin el kernel y exige de su
// resultado lo que el kernel no deja ver en el sobre: los datos del tipo de su
// salida, ninguna operación de grafo, ninguna forma propia para una persona,
// nada que ensayar y ninguna fecha declarada, que pone el reloj del kernel (H23
// FR-003, FR-004; research.md D6 y D10 de H23).
func compruebaResultadoCalculado(t *testing.T, argumentos Argumentos, salida any) {
	t.Helper()

	resultado, err := argumentos.Ejecutar(t.Context(), schema.Contexto{}, slog.New(slog.DiscardHandler))
	require.NoError(t, err)

	assert.IsType(t, salida, resultado.Datos)
	assert.Equal(t, schema.Observado{}, resultado.Grafo, "cita no observa el mundo: lo que lee no viene de una fuente")
	assert.Empty(t, resultado.Legible, "sin --json vale la tabla del kernel")
	assert.Empty(t, resultado.Ensayo, "no tiene ninguna capa con efectos que describir")
	assert.Equal(t, fuenteDeCitaEsperada, resultado.Procedencia.Fuente)
	assert.True(t, resultado.Procedencia.FechaConsulta.IsZero(), "la fecha la pone el reloj del kernel")
}

// rechazosDeCotejar son las filas propias de cita cotejar (H23 FR-026): ningún
// texto —con --documento vacío, que es el texto y deja la entrada sin leer, o
// sin él y con la entrada vacía—, un texto sin ficha y una ficha a la que le
// falta un dato o que lleva uno sin su forma; y la referencia, que se comprueba
// antes que el texto.
func rechazosDeCotejar(t *testing.T) []rechazoDeCita {
	t.Helper()

	const (
		frase  = "Una frase cualquiera, que no es el encabezamiento de ningún documento.\n"
		seDice = "--documento"
	)

	fragmento := fragmentoDeLaEvidencia(t)
	sinPonente := sinLaLinea(t, fragmento, "Ponente:")
	conOtraFecha := strings.Replace(fragmento, "Fecha: 04/07/2023", "Fecha: 31/02/2023", 1)

	require.NotEqual(t, fragmento, conOtraFecha, "el fragmento lleva la línea de la fecha")

	return []rechazoDeCita{
		{
			nombre: "ningún texto: sin documento y con la entrada vacía", llamada: map[string]string{},
			nombra: []string{seDice, "entrada estándar"},
		},
		{
			nombre: "ningún texto: el documento vacío, con un documento en la entrada, que no se lee", orden: []string{"--documento", ""},
			enLaOrden: fragmento, sinLeerla: true,
			llamada: map[string]string{"documento": ""},
			nombra:  []string{seDice, "entrada estándar"},
		},
		{
			nombre: "ningún texto: el documento vacío, con el signo igual", orden: []string{"--documento="},
			enLaOrden: fragmento, sinLeerla: true,
			llamada: map[string]string{"documento": ""},
			nombra:  []string{seDice, "entrada estándar"},
		},
		{
			nombre: "ningún texto, con su referencia", orden: []string{"--roj", "STS 1088/2023"},
			llamada: map[string]string{"roj": "STS 1088/2023"},
			nombra:  []string{seDice, "entrada estándar"},
		},
		{
			nombre: "un texto sin ficha", enLaOrden: frase,
			llamada: map[string]string{"documento": frase},
			nombra:  []string{"«Roj:»"},
		},
		{
			nombre: "una ficha sin un dato", enLaOrden: sinPonente,
			llamada: map[string]string{"documento": sinPonente},
			nombra:  []string{"«Ponente»"},
		},
		{
			nombre: "una ficha con un dato sin su forma", enLaOrden: conOtraFecha,
			llamada: map[string]string{"documento": conOtraFecha},
			nombra:  []string{"«Fecha»", `"31/02/2023"`},
		},
		{
			nombre: "una referencia mal dada y ningún texto: la referencia va antes", orden: []string{"--roj", ""},
			llamada:    map[string]string{"roj": ""},
			nombra:     []string{`ROJ ""`},
			noSeNombra: []string{seDice},
		},
	}
}

// TestCitaCotejar fija el verbo cotejar del applet cita, por el kernel entero
// (contracts/applet-cita.md §1, §2, §4, §6, §7 y §9 de H23):
//
//   - sus siete golden, con el fragmento de la evidencia en la entrada estándar,
//     son byte a byte salvo fecha_consulta lo que da el verbo, y cada salida
//     valida contra su parte de schemas/cita.json (FR-081; SC-004), con el
//     sobre de FR-004: la url lleva la huella del texto recibido, que es la del
//     manifiesto de la evidencia. Que el documento no sea el pedido termina con
//     0 y ok verdadero (FR-025);
//   - el texto es el de --documento si se escribió, con sus bytes, y solo sin él
//     el de la entrada estándar; la ficha sola da el mismo cotejo que el
//     fragmento entero, con otra url (FR-020, FR-022);
//   - cada fila de errores de §6 termina con 2 y la clase argumentos, por la
//     línea de órdenes y por la línea que compone una llamada de herramienta,
//     que devuelve el mismo sobre y nunca lee la entrada; --documento "" con un
//     documento en la entrada no la lee (FR-006, FR-020, FR-026, FR-030);
//   - no devuelve operaciones de grafo ni cuenta su salida de otra forma que el
//     kernel (FR-003).
func TestCitaCotejar(t *testing.T) {
	t.Parallel()

	esquema := esquemaPublicadoDeCita(t, cotejar)

	for _, golden := range goldenDelVerbo(t, cotejar, 7) {
		t.Run("golden/"+golden.fichero, func(t *testing.T) {
			t.Parallel()

			golden.comprueba(t, esquema, fragmentoDeLaEvidencia(t))
		})
	}

	for _, rechazo := range rechazosDeLaReferencia() {
		t.Run("rechazo/"+rechazo.nombre, func(t *testing.T) {
			t.Parallel()

			// Con el documento, de modo que lo único que falla es la referencia,
			// que se comprueba antes de leerlo.
			rechazo.enLaOrden = fragmentoDeLaEvidencia(t)
			rechazo.sinLeerla = true
			rechazo.llamada = maps.Collect(maps.All(rechazo.llamada))
			rechazo.llamada["documento"] = fragmentoDeLaEvidencia(t)

			rechazo.comprueba(t, cotejar, esquema)
		})
	}

	for _, rechazo := range rechazosDeCotejar(t) {
		t.Run("rechazo/"+rechazo.nombre, func(t *testing.T) {
			t.Parallel()

			rechazo.comprueba(t, cotejar, esquema)
		})
	}

	t.Run("el texto es el del documento, con sus bytes, o el de la entrada", func(t *testing.T) {
		t.Parallel()

		compruebaElTextoDeCotejar(t)
	})

	t.Run("la ficha sola da el mismo cotejo, con otra url", func(t *testing.T) {
		t.Parallel()

		compruebaLaFichaSola(t)
	})

	t.Run("la llamada con resultado devuelve el sobre de su orden", func(t *testing.T) {
		t.Parallel()

		// Con la ficha sola, que es lo que la skill pasa: en documento en la
		// llamada y por la entrada estándar en la orden. Una referencia que no
		// es la del documento es un resultado con su hallazgo, y no un error de
		// herramienta (FR-025).
		ficha := fichaDeLaEvidencia(t)

		compruebaLlamadasConResultado(t, cotejar, ficha, []llamadaConResultado{
			{llamada: map[string]string{"documento": ficha}},
			{orden: []string{"--roj", "STS 1088/2023"}, llamada: map[string]string{"roj": "STS 1088/2023", "documento": ficha}},
			{orden: []string{ecliDelFragmento}, llamada: map[string]string{"ecli": ecliDelFragmento, "documento": ficha}},
		})
	})

	t.Run("no observa nada y no cuenta su salida", func(t *testing.T) {
		t.Parallel()

		documento := cli.Literal(fragmentoDeLaEvidencia(t))

		compruebaResultadoCalculado(t, &argumentosDeCotejar{Documento: &documento}, cita.Cotejo{})
	})
}

// compruebaElTextoDeCotejar exige que el texto que coteja el verbo sea el de
// --documento cuando se escribe, sin leer la entrada, y el de la entrada solo
// sin él; y que los dos caminos den el mismo sobre con el mismo texto, también
// con bytes que no son UTF-8, que un argumento de texto corriente cambiaría: la
// url es la huella de lo recibido, byte a byte (H23 FR-004, FR-020; research.md
// D12 de H23).
func compruebaElTextoDeCotejar(t *testing.T) {
	t.Helper()

	textos := map[string]string{
		"el fragmento": fragmentoDeLaEvidencia(t),
		// Dos bytes que no forman ningún carácter, detrás del fragmento.
		"el fragmento con bytes que no son UTF-8": fragmentoDeLaEvidencia(t) + "\xff\xfe",
	}

	for nombre, texto := range textos {
		url := prefijoDelDocumentoEsperado + huellaDe(texto)

		porLaEntrada, entrada := ordenDeCita(t, cotejar, texto)
		require.Equal(t, 0, porLaEntrada.codigo, "%s: %s", nombre, porLaEntrada.errores)
		assert.Positive(t, entrada.lecturas, "%s: sin --documento, el texto se lee de la entrada", nombre)
		assert.Equal(t, url, sobreDelJSON(t, porLaEntrada.salida)["url"], nombre)

		// Con otro documento en la entrada, que daría otra url y un fallo.
		porLaBandera, entrada := ordenDeCita(t, cotejar, "otro texto, sin ficha", "--documento", texto)
		require.Equal(t, 0, porLaBandera.codigo, "%s: %s", nombre, porLaBandera.errores)
		assert.Zero(t, entrada.lecturas, "%s: con --documento, la entrada no se lee", nombre)

		assert.Equal(t, conLaFechaDeLosGolden(t, porLaEntrada.salida), conLaFechaDeLosGolden(t, porLaBandera.salida),
			"%s: el mismo texto da el mismo sobre por los dos caminos", nombre)
	}
}

// compruebaLaFichaSola exige que la ficha sola —lo que la skill pasa— dé los
// mismos data y hash que el fragmento entero, porque del documento solo sale su
// ficha, y otra url, la de su texto (H23 FR-022; contracts/applet-cita.md §4).
func compruebaLaFichaSola(t *testing.T) {
	t.Helper()

	ficha := fichaDeLaEvidencia(t)
	require.NotEqual(t, fragmentoDeLaEvidencia(t), ficha)

	delFragmento, _ := ordenDeCita(t, cotejar, fragmentoDeLaEvidencia(t), "--roj", "STS 1088/2023")
	require.Equal(t, 0, delFragmento.codigo, delFragmento.errores)

	deLaFicha, _ := ordenDeCita(t, cotejar, ficha, "--roj", "STS 1088/2023")
	require.Equal(t, 0, deLaFicha.codigo, deLaFicha.errores)

	entero, sola := sobreDelJSON(t, delFragmento.salida), sobreDelJSON(t, deLaFicha.salida)

	assert.Equal(t, entero["data"], sola["data"], "del documento solo sale su ficha")
	assert.Equal(t, entero["hash"], sola["hash"])
	assert.Equal(t, prefijoDelDocumentoEsperado+huellaDe(ficha), sola["url"], "la url es la del texto recibido")
	assert.NotEqual(t, entero["url"], sola["url"])
}

// TestAppletCita fija lo que declara el applet (contracts/applet-cita.md §1 de
// H23): su nombre, sus dos verbos en su orden, ninguno por omisión —nombrar el
// verbo es obligatorio, y sin él la invocación termina con 2 y los nombra— y el
// valor cero del tipo de su data, del que sale su esquema (H23 FR-001).
func TestAppletCita(t *testing.T) {
	t.Parallel()

	applet := AppletCita()

	assert.Equal(t, "cita", applet.Nombre())
	assert.NotEmpty(t, applet.Descripcion())

	verbos := applet.Verbos()
	require.Equal(t, []string{preparar, cotejar}, nombresDeLosVerbos(applet))

	assert.IsType(t, cita.Consulta{}, verbos[0].Salida)
	assert.IsType(t, cita.Cotejo{}, verbos[1].Salida)

	for _, verbo := range verbos {
		assert.False(t, verbo.PorOmision, "%s no es el verbo por omisión", verbo.Nombre)
		assert.NotEmpty(t, verbo.Descripcion, verbo.Nombre)

		unos, otros := verbo.Argumentos(), verbo.Argumentos()
		assert.NotSame(t, unos, otros, "%s: la fábrica da un valor nuevo cada vez", verbo.Nombre)
	}

	res := invocar(t, registroDeCita(t, nil), "kitlegal", "cita", "--json")

	exigirSobreDeFallo(t, res, schema.ClaseArgumentos, 2, cli.ProcedenciaKernel())
	assert.Contains(t, res.errores, "verbos de cita: preparar, cotejar")

	// La entrada solo la recibe el verbo que la lee.
	_, lee := verbos[0].Argumentos().(lector)
	assert.False(t, lee, "preparar no lee la entrada estándar")

	_, lee = verbos[1].Argumentos().(lector)
	assert.True(t, lee, "cotejar recibe del kernel la entrada de su orden")

	// Ejecutado sin el kernel y sin entrada, cotejar no tiene texto: es el error
	// de argumentos, y no un pánico.
	_, err := (&argumentosDeCotejar{}).Ejecutar(t.Context(), schema.Contexto{}, slog.New(slog.DiscardHandler))
	require.Error(t, err)
	assert.Equal(t, schema.ClaseArgumentos, cli.Clasificar(err))
}
