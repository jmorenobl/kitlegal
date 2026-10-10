package evals

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// La pregunta, la respuesta y los textos con los que juzgan los tests del juez.
// La respuesta lleva, en su segundo párrafo, énfasis de Markdown, un salto de
// línea dentro de la frase y un tramo de código: lo que la comprobación de la
// frase tolera (contracts/juez-y-voto.md §6 de H24).
const (
	preguntaJuzgada = "¿Qué plazo tiene la Administración para resolver un procedimiento?"

	respuestaJuzgada = "El artículo 21 de la Ley 39/2015 obliga a la Administración a dictar resolución expresa.\n\n" +
		"El plazo máximo **no puede exceder de seis meses**,\nsalvo que una norma con rango de `ley` establezca uno mayor.\n\n" +
		"El artículo 22 regula la suspensión del plazo, y el artículo 24, el silencio administrativo.\n\n" +
		"He comprobado la redacción y no ha cambiado.\n\n" +
		"[BOE-A-2015-10565, bloque a21]"
)

// Frases que los votos grabados citan de respuestaJuzgada: tres que están tal
// cual, una que solo difiere de la respuesta en blancos y énfasis, la del
// proceso y una que no está.
const (
	fraseDelArticulo22       = "El artículo 22 regula la suspensión del plazo"
	fraseDelArticulo24       = "el artículo 24, el silencio administrativo"
	fraseDeLosDosArticulos   = "regula la suspensión del plazo, y el artículo 24"
	fraseSinEnfasisNiSaltos  = "no puede exceder de seis meses, salvo que una norma con rango de ley establezca uno mayor"
	fraseDelProceso          = "He comprobado la redacción y no ha cambiado."
	fraseQueNoEsta           = "El artículo 23 regula la ampliación del plazo"
	preceptoDeLosVotos       = "Art. 22 de la Ley 39/2015"
	motivoDeLosVotosQueSi    = "Dice de qué trata un artículo que ninguna herramienta devolvió."
	motivoDeLosVotosQueNo    = "Todo lo que expone está en el texto devuelto."
	motivoDelProcesoContado  = "Cuenta que hizo una comprobación y lo que encontró."
	motivoDelProcesoSinNada  = "No cuenta comprobaciones ni nombra herramientas."
	claseAfirmaLoNoLeido     = "afirma_lo_no_leido"
	claseCuentaSuProceso     = "cuenta_su_proceso"
	motivoDelTopeDelVotoUno  = "voto 1: tope de 35 s agotado"
	motivoDelTopeDelVotoDos  = "voto 2: tope de 35 s agotado"
	prefijoDeSesionConError  = "voto 1: la sesión del juez terminó con error: "
	prefijoDeSalidaSinForma  = "voto 1: la respuesta no tiene la forma del esquema: "
	limiteDeUsoDeLaSesion    = "You've hit your limit · resets 3am (Europe/Madrid)"
	respuestaQueNoEsUnJuicio = "No puedo responder a esas dos preguntas."
)

// La pregunta, la respuesta y el texto con los que juzgan los tests del juez
// cuyo esquema lleva sentencia, los del ejemplo de
// contracts/juez-de-jurisprudencia.md §2 de H25: la respuesta lleva la línea de
// la sentencia no comprobada y, en su segundo párrafo, lo que dice que declaró
// una sentencia cuyo texto no tiene delante.
const (
	preguntaDeLaSentencia = "resúmeme la STS 241/2013, de 9 de mayo"

	parrafoDeLaSentencia = "La STS 241/2013, de 9 de mayo, declaró la nulidad de las cláusulas suelo por falta de transparencia."

	respuestaConLaSentencia = "⚠ SENTENCIA NO COMPROBADA: STS 241/2013, de 9 de mayo\n\n" + parrafoDeLaSentencia + "\n"

	ordenDeLaSentencia = "kitlegal cita preparar --resolucion 241/2013 --fecha 2013-05-09 --json"
	sobreDeLaSentencia = `{"ok":true,"fuente":"kitlegal.cita","url":"https://www.poderjudicial.es/search/indexAN.jsp",` +
		`"data":{"referencia":{"forma":"resolucion","valor":"241/2013","fecha":"2013-05-09"}}}`
)

// Lo que los votos grabados de ese juez dicen de respuestaConLaSentencia, que
// es lo que dice el voto del ejemplo de contracts/juez-de-jurisprudencia.md §3
// de H25: la frase que citan de parrafoDeLaSentencia, la sentencia de la que
// habla y el motivo de cada respuesta en cada clase.
const (
	fraseDeLaNulidad    = "declaró la nulidad de las cláusulas suelo por falta de transparencia"
	sentenciaDeLosVotos = "STS 241/2013, de 9 de mayo"

	motivoDeLaSentenciaNoLeida = "La respuesta dice qué declaró la sentencia, y su texto no está ni en la pregunta ni en " +
		"lo que devolvió la herramienta."
	motivoDeLaSentenciaSinResumir = "La respuesta no dice nada de lo que declara la sentencia."
	motivoDeLaQueNoDiceQueExiste  = "La respuesta lleva la línea de la sentencia no comprobada y no dice que exista."
	motivoDeLaQueDiceQueExiste    = "La respuesta dice que la sentencia existe, y su documento no está en la pregunta."

	claseAfirmaQueExiste = "afirma_que_existe"
)

// juicioDelEjemploConSentencia es el voto del ejemplo de
// contracts/juez-de-jurisprudencia.md §3 de H25, escrito a mano: lo que el juez
// de jurisprudencia da por structured_output cuando dice sí en la clase que
// decide, con su frase y su sentencia, y no en la que solo se publica.
const juicioDelEjemploConSentencia = `{"afirma_lo_no_leido":{"motivo":"La respuesta dice qué declaró la sentencia, ` +
	`y su texto no está ni en la pregunta ni en lo que devolvió la herramienta.","respuesta":"si",` +
	`"frase":"declaró la nulidad de las cláusulas suelo por falta de transparencia",` +
	`"sentencia":"STS 241/2013, de 9 de mayo"},` +
	`"afirma_que_existe":{"motivo":"La respuesta lleva la línea de la sentencia no comprobada y no dice que exista.",` +
	`"respuesta":"no","frase":""}}`

// Dos blancos que no son ASCII, escritos por sus bytes en UTF-8 para que se
// vean en el fuente: el espacio de no separación, U+00A0, y el separador de
// línea, U+2028. Los dos son de unicode.IsSpace (research V6 de H24).
const (
	espacioDeNoSeparacion = "\xc2\xa0"
	separadorDeLinea      = "\xe2\x80\xa8"
)

// esquemaDeLasDosClases es el esquema de la respuesta de un juez con las dos
// clases de boe-legislacion, con la forma del que validó el ADR 0037: las dos
// obligatorias, sin más propiedades, y precepto solo en la que decide.
const esquemaDeLasDosClases = `{"type":"object","additionalProperties":false,` +
	`"required":["afirma_lo_no_leido","cuenta_su_proceso"],"properties":{` +
	`"afirma_lo_no_leido":{"type":"object","additionalProperties":false,` +
	`"required":["motivo","respuesta","frase","precepto"],"properties":{"motivo":{"type":"string"},` +
	`"respuesta":{"type":"string","enum":["si","no"]},"frase":{"type":"string"},"precepto":{"type":"string"}}},` +
	`"cuenta_su_proceso":{"type":"object","additionalProperties":false,` +
	`"required":["motivo","respuesta","frase"],"properties":{"motivo":{"type":"string"},` +
	`"respuesta":{"type":"string","enum":["si","no"]},"frase":{"type":"string"}}}}}`

// esquemaConSentencia es el esquema de la respuesta de un juez con las dos
// clases de jurisprudencia, con la forma del que validó el ADR 0037 para esa
// skill: las dos obligatorias, sin más propiedades, y sentencia, donde el de
// boe-legislacion lleva precepto, solo en la que decide.
const esquemaConSentencia = `{"type":"object","additionalProperties":false,` +
	`"required":["afirma_lo_no_leido","afirma_que_existe"],"properties":{` +
	`"afirma_lo_no_leido":{"type":"object","additionalProperties":false,` +
	`"required":["motivo","respuesta","frase","sentencia"],"properties":{"motivo":{"type":"string"},` +
	`"respuesta":{"type":"string","enum":["si","no"]},"frase":{"type":"string"},"sentencia":{"type":"string"}}},` +
	`"afirma_que_existe":{"type":"object","additionalProperties":false,` +
	`"required":["motivo","respuesta","frase"],"properties":{"motivo":{"type":"string"},` +
	`"respuesta":{"type":"string","enum":["si","no"]},"frase":{"type":"string"}}}}}`

// La salida estándar de una sesión del juez, con la forma de la de claude -p
// --output-format json que lee voto_real de evidencias/adr-0037/guiones/juez.py
// (is_error, structured_output y result, entre otras claves que no se leen), en
// dos trozos entre los que va el structured_output. En una sesión de un paso no
// se ha podido ejecutar claude para contrastarla (research S5 de H24).
const (
	cabeceraDeLaSalidaDelJuez = `{"type":"result","subtype":"success","is_error":false,"duration_ms":8123,` +
		`"duration_api_ms":7990,"num_turns":1,"result":"","stop_reason":"end_turn",` +
		`"session_id":"00000000-0000-4000-8000-000000000101","total_cost_usd":0.0312,` +
		`"usage":{"input_tokens":2804,"output_tokens":187},` +
		`"modelUsage":{"claude-opus-5-5":{"inputTokens":2804,"outputTokens":187,"costUSD":0.0312}},` +
		`"permission_denials":[],"structured_output":`
	pieDeLaSalidaDelJuez = `,"uuid":"00000000-0000-4000-8000-000000000102"}` + "\n"
)

// textosJuzgados son los textos de las herramientas de la respuesta que juzgan
// los tests: uno del modo orden y otro del modo herramienta.
func textosJuzgados() []Texto {
	return []Texto{
		{Orden: ordenDelArticulo, Salida: sobreDelArticulo},
		{Orden: "boe_indice BOE-A-2015-10565", Salida: sobreDelIndice},
	}
}

// juezDeLasDosClases es un juez con las dos clases de boe-legislacion: la que
// decide y la que solo se publica, en ese orden.
func juezDeLasDosClases() *Juez {
	return &Juez{
		Clases: []ClaseDelJuez{
			{Nombre: claseAfirmaLoNoLeido, Decide: true},
			{Nombre: claseCuentaSuProceso},
		},
		Esquema: esquemaDeLasDosClases,
	}
}

// juezConSentencia es un juez con las dos clases de jurisprudencia: la que
// decide, cuyo esquema lleva sentencia, y la que solo se publica, en ese orden.
func juezConSentencia() *Juez {
	return &Juez{
		Clases: []ClaseDelJuez{
			{Nombre: claseAfirmaLoNoLeido, Decide: true},
			{Nombre: claseAfirmaQueExiste},
		},
		Esquema: esquemaConSentencia,
	}
}

// respuestaSobreLaSentencia es la respuesta que juzgan los tests del juez cuyo
// esquema lleva sentencia: respuestaConLaSentencia, con su pregunta y el texto
// de su herramienta.
func respuestaSobreLaSentencia() respuestaAJuzgar {
	return respuestaAJuzgar{
		pregunta:  preguntaDeLaSentencia,
		respuesta: respuestaConLaSentencia,
		textos:    []Texto{{Orden: ordenDeLaSentencia, Salida: sobreDeLaSentencia}},
	}
}

// dicho es lo que un voto grabado dice de una clase, con los campos de
// esquema.json: precepto y sentencia son nil en la clase que no los tiene, y
// ninguna tiene los dos.
type dicho struct {
	motivo    string
	respuesta string
	frase     string
	precepto  *string
	sentencia *string
}

// afirmaQueSi es lo que dice de afirma_lo_no_leido el voto que la marca con
// esa frase.
func afirmaQueSi(frase string) dicho {
	return dicho{motivo: motivoDeLosVotosQueSi, respuesta: "si", frase: frase, precepto: new(preceptoDeLosVotos)}
}

// afirmaQueNo es lo que dice de afirma_lo_no_leido el voto que no la marca.
func afirmaQueNo() dicho {
	return dicho{motivo: motivoDeLosVotosQueNo, respuesta: "no", precepto: new("")}
}

// cuentaQueSi es lo que dice de cuenta_su_proceso el voto que dice sí con esa
// frase.
func cuentaQueSi(frase string) dicho {
	return dicho{motivo: motivoDelProcesoContado, respuesta: "si", frase: frase}
}

// cuentaQueNo es lo que dice de cuenta_su_proceso el voto que dice no.
func cuentaQueNo() dicho {
	return dicho{motivo: motivoDelProcesoSinNada, respuesta: "no"}
}

// afirmaConSentenciaQueSi es lo que dice de afirma_lo_no_leido, con el juez
// cuyo esquema lleva sentencia, el voto que la marca con esa frase.
func afirmaConSentenciaQueSi(frase string) dicho {
	return dicho{
		motivo: motivoDeLaSentenciaNoLeida, respuesta: "si", frase: frase, sentencia: new(sentenciaDeLosVotos),
	}
}

// afirmaConSentenciaQueNo es lo que dice de afirma_lo_no_leido, con ese juez,
// el voto que no la marca: su sentencia va vacía.
func afirmaConSentenciaQueNo() dicho {
	return dicho{motivo: motivoDeLaSentenciaSinResumir, respuesta: "no", sentencia: new("")}
}

// existeQueNo es lo que dice de afirma_que_existe el voto que dice no.
func existeQueNo() dicho {
	return dicho{motivo: motivoDeLaQueNoDiceQueExiste, respuesta: "no"}
}

// existeQueSi es lo que dice de afirma_que_existe el voto que dice sí con esa
// frase.
func existeQueSi(frase string) dicho {
	return dicho{motivo: motivoDeLaQueDiceQueExiste, respuesta: "si", frase: frase}
}

// voto es el voto de la clase que la lectura tiene que dejar de lo dicho: sus
// campos, tal como los dio el juez, con el número del voto, si el voto entero es
// nulo y si su frase está en la respuesta.
func (d dicho) voto(numero int, nulo, enLaRespuesta bool) VotoDeClase {
	return VotoDeClase{
		Voto:               numero,
		Nulo:               nulo,
		Motivo:             d.motivo,
		Respuesta:          d.respuesta,
		Frase:              d.frase,
		Precepto:           d.precepto,
		Sentencia:          d.sentencia,
		FraseEnLaRespuesta: enLaRespuesta,
	}
}

// grabacion es lo que el votante de los tests devuelve de un voto: la salida
// estándar de la sesión del juez y el error del votante.
type grabacion struct {
	salida string
	err    error
}

// juicioGrabado es el juicio de un voto, el objeto que el juez da por
// structured_output: por cada clase, lo dicho de ella.
func juicioGrabado(t *testing.T, dichos map[string]dicho) string {
	t.Helper()

	juicio := make(map[string]map[string]string, len(dichos))

	for clase, dicho := range dichos {
		campos := map[string]string{"motivo": dicho.motivo, "respuesta": dicho.respuesta, "frase": dicho.frase}
		if dicho.precepto != nil {
			campos["precepto"] = *dicho.precepto
		}

		if dicho.sentencia != nil {
			campos["sentencia"] = *dicho.sentencia
		}

		juicio[clase] = campos
	}

	codificado, err := json.Marshal(juicio)
	require.NoError(t, err)

	return string(codificado)
}

// votoGrabado es la grabación de un voto que llega a darse: la salida de la
// sesión del juez con ese juicio en structured_output.
func votoGrabado(t *testing.T, dichos map[string]dicho) grabacion {
	t.Helper()

	return grabacion{salida: cabeceraDeLaSalidaDelJuez + juicioGrabado(t, dichos) + pieDeLaSalidaDelJuez}
}

// votoDeLasDosClases es la grabación de un voto del juez de las dos clases, con
// lo que dice de cada una.
func votoDeLasDosClases(t *testing.T, afirma, cuenta dicho) grabacion {
	t.Helper()

	return votoGrabado(t, map[string]dicho{claseAfirmaLoNoLeido: afirma, claseCuentaSuProceso: cuenta})
}

// votoConSentencia es la grabación de un voto del juez cuyo esquema lleva
// sentencia, con lo que dice de cada una de sus dos clases.
func votoConSentencia(t *testing.T, afirma, existe dicho) grabacion {
	t.Helper()

	return votoGrabado(t, map[string]dicho{claseAfirmaLoNoLeido: afirma, claseAfirmaQueExiste: existe})
}

// votanteGrabado es un votante que devuelve, por orden, sus grabaciones y
// anota el mensaje de cada voto que se le pide: tantos como llamadas recibe.
// Si se le pide un voto de más, el test falla y el voto no llega a darse.
type votanteGrabado struct {
	t           *testing.T
	grabaciones []grabacion
	mensajes    []string
}

func (v *votanteGrabado) votar(mensaje string) ([]byte, error) {
	v.mensajes = append(v.mensajes, mensaje)

	if len(v.mensajes) > len(v.grabaciones) {
		v.t.Errorf("se pide el voto %d y solo hay %d grabados", len(v.mensajes), len(v.grabaciones))

		return nil, errors.New("voto sin grabar")
	}

	grabacion := v.grabaciones[len(v.mensajes)-1]

	return []byte(grabacion.salida), grabacion.err
}

// juzgarConGrabaciones juzga respuestaJuzgada con ese juez y un votante que
// devuelve esas grabaciones, y da el juicio y los votos pedidos. Cada voto se
// pide con el mensaje de la respuesta, el mismo en todos.
func juzgarConGrabaciones(t *testing.T, juez *Juez, grabaciones ...grabacion) (JuicioDeRespuesta, int) {
	t.Helper()

	return juzgarLaRespuesta(t, juez, respuestaAJuzgar{
		pregunta: preguntaJuzgada, respuesta: respuestaJuzgada, textos: textosJuzgados(),
	}, grabaciones...)
}

// juzgarLaRespuesta juzga esa respuesta con ese juez y un votante que devuelve
// esas grabaciones, y da el juicio y los votos pedidos. Cada voto se pide con
// el mensaje de la respuesta, el mismo en todos.
func juzgarLaRespuesta(
	t *testing.T, juez *Juez, juzgada respuestaAJuzgar, grabaciones ...grabacion,
) (JuicioDeRespuesta, int) {
	t.Helper()

	votante := &votanteGrabado{t: t, grabaciones: grabaciones}

	votacion, err := nuevaVotacion(juez, votante.votar)
	require.NoError(t, err)

	juicio := votacion.juzgar(juzgada.pregunta, juzgada.respuesta, juzgada.textos)

	for _, mensaje := range votante.mensajes {
		assert.Equal(t, mensajeDelVoto(juzgada.pregunta, juzgada.respuesta, juzgada.textos), mensaje,
			"cada voto se pide con el mensaje de la respuesta")
	}

	return juicio, len(votante.mensajes)
}

// errorDeProceso es el error de un votante cuyo proceso terminó con ese código,
// con la forma de *exec.ExitError: lo lleva en ExitCode.
type errorDeProceso struct {
	codigo int
	texto  string
}

func (e errorDeProceso) Error() string {
	return e.texto
}

// ExitCode es el código con el que terminó el proceso del voto.
func (e errorDeProceso) ExitCode() int {
	return e.codigo
}

// TestMensajeDelVoto fija el mensaje con el que se pide cada voto, el de
// prompt_de de evidencias/adr-0037/guiones/juez.py, byte a byte
// (contracts/juez-y-voto.md §3 y §9 de H24; research D3 de H24; FR-001, FR-002,
// FR-004, FR-107; SC-007): un <texto orden="…"> por texto, en su orden y sean
// del modo que sean, con su salida sin los blancos de los extremos —los de
// unicode.IsSpace más U+001C a U+001F, que son los que quita strip en Python— o
// «(sin salida)» si queda vacía; la línea de que ninguna herramienta devolvió
// ningún texto si no hay ninguno; y la pregunta y la respuesta, tal cual. No
// lleva nada más: compuesto desde una eval y el juicio sin modelo de su sesión,
// no tiene ni un byte de lo que la eval espera, de los motivos del juicio ni de
// SKILL.md, que está en el transcript.
//
// Desde H25 (contracts/juez-de-jurisprudencia.md §2 y §9 de H25; FR-010, FR-011,
// FR-109; SC-009), lo fija también con una sesión de jurisprudencia de cada
// modo, con una orden kitlegal cita cotejar o con una llamada a cita_cotejar,
// sobre la pregunta con el fragmento pegado: el mensaje lleva la pregunta
// entera, la respuesta y cada texto, en su orden, también el de la consulta que
// falla, y nada de SKILL.md, de lo que la eval espera ni del juicio sin modelo.
// Lo que los casos de H24 esperan del mensaje no cambia.
func TestMensajeDelVoto(t *testing.T) {
	t.Parallel()

	const (
		ordenDeLaComprobacion = "kitlegal graph check BOE-A-2015-10565 --json"
		sobreDeLaComprobacion = `{"ok":true,"fuente":"kitlegal.graph","data":{"hallazgos":[]}}`
		preguntaDeLaEval      = "¿Qué dice el artículo 21 de la Ley 39/2015?\n"
	)

	casos := []struct {
		nombre    string
		pregunta  string
		respuesta string
		textos    []Texto
		mensaje   string
	}{
		{
			nombre:    "con textos de los dos modos",
			pregunta:  preguntaJuzgada,
			respuesta: respuestaJuzgada,
			textos: []Texto{
				{Orden: ordenDelArticulo, Salida: "\n  " + sobreDelArticulo + "\n\n"},
				{Orden: "boe_indice BOE-A-2015-10565", Salida: sobreDelIndice},
				{
					Orden:  ordenDeLaComprobacion,
					Salida: "\x1f" + espacioDeNoSeparacion + "\t" + sobreDeLaComprobacion + " " + separadorDeLinea + "\x1c\r\n",
				},
			},
			mensaje: "<textos_de_las_herramientas>\n" +
				"\n" +
				`<texto orden="cd /tmp/trabajo && /usr/local/bin/kitlegal boe articulo BOE-A-2015-10565 a21 --json | head -c 4000">` + "\n" +
				sobreDelArticulo + "\n" +
				"</texto>\n" +
				"\n" +
				`<texto orden="boe_indice BOE-A-2015-10565">` + "\n" +
				sobreDelIndice + "\n" +
				"</texto>\n" +
				"\n" +
				`<texto orden="kitlegal graph check BOE-A-2015-10565 --json">` + "\n" +
				sobreDeLaComprobacion + "\n" +
				"</texto>\n" +
				"\n" +
				"</textos_de_las_herramientas>\n" +
				"\n" +
				"<pregunta>\n" +
				"¿Qué plazo tiene la Administración para resolver un procedimiento?\n" +
				"</pregunta>\n" +
				"\n" +
				"<respuesta>\n" +
				respuestaJuzgada + "\n" +
				"</respuesta>\n" +
				"\n" +
				"Responde a las dos preguntas de la rúbrica sobre esta respuesta.",
		},
		{
			nombre:    "sin textos, y con la pregunta y la respuesta tal cual",
			pregunta:  preguntaDeLaEval,
			respuesta: "  ⚠ SIN CONSULTA AL BOE: no he podido leer el artículo.\n",
			mensaje: "<textos_de_las_herramientas>\n" +
				"\n" +
				"(ninguna herramienta devolvió ningún texto)\n" +
				"\n" +
				"</textos_de_las_herramientas>\n" +
				"\n" +
				"<pregunta>\n" +
				"¿Qué dice el artículo 21 de la Ley 39/2015?\n" +
				"\n" +
				"</pregunta>\n" +
				"\n" +
				"<respuesta>\n" +
				"  ⚠ SIN CONSULTA AL BOE: no he podido leer el artículo.\n" +
				"\n" +
				"</respuesta>\n" +
				"\n" +
				"Responde a las dos preguntas de la rúbrica sobre esta respuesta.",
		},
		{
			nombre:    "con la salida vacía",
			pregunta:  preguntaJuzgada,
			respuesta: "No hay nada que citar.",
			textos: []Texto{
				{Orden: ordenDelIndice},
				{Orden: "boe_articulo BOE-A-2015-10565 a21", Salida: " \n\t" + espacioDeNoSeparacion + "\x1d"},
			},
			mensaje: "<textos_de_las_herramientas>\n" +
				"\n" +
				`<texto orden="kitlegal boe indice BOE-A-2015-10565 --json">` + "\n" +
				"(sin salida)\n" +
				"</texto>\n" +
				"\n" +
				`<texto orden="boe_articulo BOE-A-2015-10565 a21">` + "\n" +
				"(sin salida)\n" +
				"</texto>\n" +
				"\n" +
				"</textos_de_las_herramientas>\n" +
				"\n" +
				"<pregunta>\n" +
				"¿Qué plazo tiene la Administración para resolver un procedimiento?\n" +
				"</pregunta>\n" +
				"\n" +
				"<respuesta>\n" +
				"No hay nada que citar.\n" +
				"</respuesta>\n" +
				"\n" +
				"Responde a las dos preguntas de la rúbrica sobre esta respuesta.",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.mensaje, mensajeDelVoto(caso.pregunta, caso.respuesta, caso.textos))
		})
	}

	t.Run("sin nada de la skill, de la eval ni del juicio sin modelo", func(t *testing.T) {
		t.Parallel()

		const centinela = "CENTINELA"

		eval := Eval{
			Fichero:  "01-" + centinela + "-eval.yaml",
			Pregunta: preguntaJuzgada,
			Activa:   true,
			Comandos: []ComandoEsperado{
				{Applet: "boe", Verbo: "articulo", Norma: centinela + "-NORMA", Bloque: centinela + "-BLOQUE"},
			},
			Citas:     []CitaEsperada{{Norma: centinela + "-CITA", Bloque: centinela + "-a21"}},
			Avisos:    []string{centinela + "-AVISO"},
			Hallazgos: []string{centinela + "-HALLAZGO"},
		}

		transcript := mensajeInit +
			mensajeDeLlamadas(t,
				usoDeHerramienta{id: "toolu_01", nombre: herramientaSkill, entrada: activacionDeLaSkill},
				usoDeHerramienta{id: "toolu_02", nombre: "Read", entrada: `{"file_path":"/tmp/trabajo/.claude/skills/boe-legislacion/SKILL.md"}`},
				usoDeHerramienta{id: "toolu_03", nombre: "Bash", entrada: entradaDeBash(t, ordenDelArticulo)},
				usoDeHerramienta{id: "toolu_04", nombre: "mcp__kitlegal__boe_indice", entrada: argumentosDelIndice},
			) +
			mensajeDeResultados(t,
				resultadoDeHerramienta{id: "toolu_01", contenido: cadenaJSON(t, "Launching skill: "+centinela+"-SKILL")},
				resultadoDeHerramienta{id: "toolu_02", contenido: cadenaJSON(t, "# "+centinela+"-SKILL-MD\n\nProtocolo de la skill.")},
				resultadoDeHerramienta{id: "toolu_03", contenido: cadenaJSON(t, sobreDelArticulo)},
				resultadoDeHerramienta{id: "toolu_04", contenido: cadenaJSON(t, sobreDelIndice)},
			) +
			`{"type":"user","message":{"role":"user","content":[{"type":"text","text":"Base directory for this skill: ` +
			centinela + `-SKILL-MD"}]}}` + "\n" +
			`{"type":"result","subtype":"success","is_error":false,"result":` + cadenaJSON(t, respuestaJuzgada) + `}` + "\n"

		sesion, err := LeerSesion(escribirSesion(t, transcript))
		require.NoError(t, err)

		resultado := Juzgar(eval, sesion, skillDeLasSesiones)

		require.Contains(t, strings.Join(resultado.Motivos, "\n"), centinela,
			"el juicio sin modelo de la sesión nombra lo que la eval espera y no está")
		require.Contains(t, transcript, centinela+"-SKILL-MD")

		mensaje := mensajeDelVoto(eval.Pregunta, resultado.Respuesta, sesion.Textos)

		assert.Equal(t, mensajeDelVoto(preguntaJuzgada, respuestaJuzgada, textosJuzgados()), mensaje,
			"el mensaje es el de la pregunta, la respuesta y los textos de las herramientas")
		assert.NotContains(t, mensaje, centinela)

		for _, motivo := range resultado.Motivos {
			assert.NotContains(t, mensaje, motivo)
		}
	})

	for _, modo := range []Modo{ModoOrden, ModoHerramienta} {
		t.Run("de una sesión de jurisprudencia del modo "+string(modo), func(t *testing.T) {
			t.Parallel()

			probarElMensajeDeUnaSesionDeCita(t, modo)
		})
	}
}

// probarElMensajeDeUnaSesionDeCita compone el mensaje del voto de una sesión de
// jurisprudencia de ese modo sobre la pregunta con el fragmento pegado
// (sesionDeCitaEn), desde su eval y el juicio sin modelo de la sesión, y exige
// que sea, byte a byte, el escrito a mano (contracts/juez-de-jurisprudencia.md
// §2 de H25): el texto del cotejo y, detrás, el de la consulta que falla, cada
// uno con su orden tal cual; la pregunta entera, con el fragmento; y la
// respuesta. Y que no lleve nada más: ni un byte de SKILL.md, que está en el
// transcript, ni de lo que la eval espera, ni de los motivos del juicio.
func probarElMensajeDeUnaSesionDeCita(t *testing.T, modo Modo) {
	t.Helper()

	const (
		centinela = "CENTINELA"
		respuesta = "Es la STS 1088/2023, de 4 de julio " + citaDelFragmento + "."
	)

	fragmento := string(contenidoDelFichero(t, fragmentoDelRepositorio))
	pregunta := entradaConElTexto + "\n\n" + fragmento

	eval := Eval{
		Fichero:  "04-" + centinela + "-eval.yaml",
		Pregunta: pregunta,
		Activa:   true,
		Comandos: []ComandoEsperado{{Applet: "cita", Verbo: "cotejar", ROJ: centinela + "-ROJ-DEL-COMANDO"}},
		Sentencias: SentenciasEsperadas{
			Citas:       []CitaDeSentenciaEsperada{{ECLI: centinela + "-ECLI", ROJ: centinela + "-ROJ"}},
			Direcciones: []string{"https://" + centinela + ".example/buscador"},
			Casillas:    []CasillaEsperada{{Nombre: centinela + "-CASILLA", Valor: centinela + "-VALOR"}},
		},
	}

	deCita := sesionDeCitaEn(t, modo, fragmento, "# "+centinela+"-SKILL-MD\n\nProtocolo de la skill.")
	transcript := mensajeInit + deCita.transcript +
		`{"type":"result","subtype":"success","is_error":false,"result":` + cadenaJSON(t, respuesta) + `}` + "\n"

	sesion, err := LeerSesion(escribirSesion(t, transcript))
	require.NoError(t, err)

	resultado := Juzgar(eval, sesion, skillDeJurisprudencia)

	require.Contains(t, strings.Join(resultado.Motivos, "\n"), centinela,
		"el juicio sin modelo de la sesión nombra lo que la eval espera y no está")
	require.Contains(t, transcript, centinela+"-SKILL-MD")
	require.Len(t, deCita.textos, 2, "premisa: la sesión deja el texto del cotejo y el de la consulta que falla")
	require.Contains(t, deCita.textos[0].Orden, fragmento, "premisa: la orden del cotejo lleva el documento")

	mensaje := mensajeDelVoto(eval.Pregunta, resultado.Respuesta, sesion.Textos)

	assert.Equal(t, "<textos_de_las_herramientas>\n"+
		"\n"+
		`<texto orden="`+deCita.textos[0].Orden+`">`+"\n"+
		strings.TrimSuffix(deCita.textos[0].Salida, "\n")+"\n"+
		"</texto>\n"+
		"\n"+
		`<texto orden="`+deCita.textos[1].Orden+`">`+"\n"+
		deCita.textos[1].Salida+"\n"+
		"</texto>\n"+
		"\n"+
		"</textos_de_las_herramientas>\n"+
		"\n"+
		"<pregunta>\n"+
		entradaConElTexto+"\n"+
		"\n"+
		fragmento+"\n"+
		"</pregunta>\n"+
		"\n"+
		"<respuesta>\n"+
		respuesta+"\n"+
		"</respuesta>\n"+
		"\n"+
		"Responde a las dos preguntas de la rúbrica sobre esta respuesta.", mensaje)
	assert.NotContains(t, mensaje, centinela)

	for _, motivo := range resultado.Motivos {
		assert.NotContains(t, mensaje, motivo)
	}
}

// TestFraseEnLaRespuesta fija la comprobación sin modelo de que la frase que
// cita un voto está en la respuesta, la de normal y frase_esta de
// evidencias/adr-0037/guiones/juez.py (contracts/juez-y-voto.md §6 y §9 de H24;
// research D8, V5 y V6 de H24; FR-005, FR-102): de la frase y de la respuesta se
// quitan *, _ y el acento grave; cada serie de blancos, que son los de
// unicode.IsSpace más U+001C a U+001F, se cambia por un espacio; se recortan
// los extremos; y la frase está si, sin quedar vacía, es subcadena de la
// respuesta. Nada más se tolera: una mayúscula, un acento, una coma o un blanco
// de menos hacen que no esté.
func TestFraseEnLaRespuesta(t *testing.T) {
	t.Parallel()

	const respuesta = "La Administración está obligada a dictar resolución expresa\n" +
		"y a notificarla en todos los procedimientos.\n\n" +
		"El plazo máximo es de `seis meses`, salvo que una norma disponga otro.\n\n" +
		"Los **efectos del silencio** son _estimatorios_ en este caso.\n\n" +
		"Véase el art." + espacioDeNoSeparacion + "24 de la Ley\x1f39/2015."

	casos := []struct {
		nombre string
		frase  string
		esta   bool
	}{
		{nombre: "literal", frase: "y a notificarla en todos los procedimientos.", esta: true},
		{nombre: "cruza un salto de línea", frase: "dictar resolución expresa y a notificarla", esta: true},
		{nombre: "lleva otros blancos que la respuesta", frase: "  dictar \t resolución\r\nexpresa\n\ny a notificarla ", esta: true},
		{nombre: "pierde un acento grave", frase: "El plazo máximo es de seis meses, salvo", esta: true},
		{nombre: "pierde el asterisco y el guion bajo", frase: "Los efectos del silencio son estimatorios en este caso.", esta: true},
		{nombre: "lleva el énfasis que la respuesta no tiene", frase: "**y a notificarla** en `todos` los _procedimientos_.", esta: true},
		{nombre: "lleva un espacio de no separación", frase: "Véase el art. 24 de la Ley", esta: true},
		{nombre: "lleva un separador de unidad", frase: "de la Ley 39/2015.", esta: true},
		{nombre: "vacía", frase: "", esta: false},
		{nombre: "vacía sin sus blancos y su énfasis", frase: " ** _\n` " + espacioDeNoSeparacion, esta: false},
		{nombre: "cambia una mayúscula", frase: "el plazo máximo es de seis meses", esta: false},
		{nombre: "cambia un acento", frase: "El plazo maximo es de seis meses", esta: false},
		{nombre: "cambia una coma", frase: "seis meses salvo que una norma", esta: false},
		{nombre: "pierde el blanco entre dos palabras", frase: "resolución expresay a notificarla", esta: false},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.esta, fraseEsta(caso.frase, respuesta))
		})
	}
}

// guionDelVoto es scripts/evals-voto.sh, relativo al directorio de este
// paquete, que es donde go test ejecuta los tests.
const guionDelVoto = "../../scripts/evals-voto.sh"

// Lo que TestOrdenDelVoto da al votante del guion: el id del modelo del juez,
// que la orden recibe tal cual y ninguna sesión usa; y una rúbrica con varias
// líneas, comillas, un dólar, un acento grave y una barra invertida, que la
// orden recibe carácter a carácter, y con una línea en blanco delante de su
// salto de línea final, que $(…) de la shell quitaría con él.
const (
	modeloDelJuezDelGuion = "claude-opus-5-5"

	rubricaDelGuion = "# Rúbrica del juez\n\n" +
		"Responde \"si\" o 'no' a cada pregunta: ni $HOME, ni `orden`, ni \\n se tocan.\n\n"
)

// El tope y el margen de los votos de los tests que agotan el tope —los de
// verdad son 35 s y 5 s (contracts/juez-y-voto.md §4 de H24)— y lo que el
// sustituto de claude espera en ellos sin terminar: el votante que lo esperase
// entero tardaría eso, y el que lo corta, el tope o el tope y el margen. Un
// voto llega a la espera del sustituto en unos 30 ms, también con el paquete
// entero en marcha, así que el tope no lo corta antes; y el margen es la
// holgura con la que se exige que el proceso se termine al agotarse el tope, y
// no después.
const (
	topeDelVotoDePrueba   = 2 * time.Second
	margenDelVotoDePrueba = 3 * time.Second
	esperaSinTerminar     = 30 * time.Second
)

// variablesDeLaShell son las que una shell pone por su cuenta en el entorno de
// lo que ejecuta: el guion del voto es de bash y el sustituto de claude, de sh,
// así que el entorno que el sustituto anota las lleva además de las que el
// votante da al voto.
var variablesDeLaShell = []string{"PWD", "OLDPWD", "SHLVL", "_"}

// ordenDelVotoConElSustituto es la orden de los votos de un test: con ese juez,
// la ruta absoluta de scripts/evals-voto.sh, el sustituto de claude delante en
// el PATH y una credencial sin forma de secreto.
func ordenDelVotoConElSustituto(t *testing.T, claude claudeDelJuez, juez *Juez) ordenDelVoto {
	t.Helper()

	guion, err := filepath.Abs(guionDelVoto)
	require.NoError(t, err)

	return ordenDelVoto{
		juez:        juez,
		modelo:      modeloDelJuezDelGuion,
		guion:       guion,
		path:        claude.path(),
		suscripcion: valorDeLaSuscripcion,
	}
}

// votoPedido es lo que un votante devolvió de un voto que se le pidió.
type votoPedido struct {
	salida []byte
	err    error
}

// juzgarConElVotante juzga respuestaJuzgada con el juez de las dos clases y ese
// votante, y da el juicio y lo que el votante devolvió de cada voto, en su
// orden.
func juzgarConElVotante(t *testing.T, votar Votante) (JuicioDeRespuesta, []votoPedido) {
	t.Helper()

	var pedidos []votoPedido

	votacion, err := nuevaVotacion(juezDeLasDosClases(), func(mensaje string) ([]byte, error) {
		salida, err := votar(mensaje)
		pedidos = append(pedidos, votoPedido{salida: salida, err: err})

		return salida, err
	})
	require.NoError(t, err)

	return votacion.juzgar(preguntaJuzgada, respuestaJuzgada, textosJuzgados()), pedidos
}

// TestOrdenDelVoto fija la orden con la que el votante del guion abre cada
// voto y su tope (contracts/juez-y-voto.md §4 y §9 de H24; research D4, D6 y V22
// de H24; FR-003, FR-004, FR-007, FR-107; SC-007), con scripts/evals-voto.sh y
// un claude sustituto delante en el PATH del voto, sin ninguna sesión con
// modelo:
//
//   - cada voto es un proceso nuevo de claude con los argumentos de voto_real
//     de evidencias/adr-0037/guiones/juez.py, uno a uno y en su orden, con la
//     rúbrica entera y el esquema sin su salto de línea final; con el mensaje
//     por su entrada estándar; en el directorio cwd, vacío, del directorio del
//     juez, un temporal que se retira al acabar; y con cuatro variables de
//     entorno y ninguna más;
//   - el votante devuelve la salida estándar del proceso, también si termina
//     con error, y entonces con un error que lleva su código y lo que escribió
//     en su salida de error;
//   - pasado el tope, termina el proceso, que no queda vivo, y devuelve el
//     error del tope; si el proceso deja sus salidas abiertas en otro, las
//     cierra pasado el margen y no lo espera;
//   - y si se corta el contexto que recibió, corta el voto abierto, con un
//     error que no es el del tope.
func TestOrdenDelVoto(t *testing.T) {
	t.Parallel()

	t.Run("orden", probarLaOrdenDelVoto)
	t.Run("proceso-que-termina-con-error", probarElVotoQueTerminaConError)
	t.Run("guion-que-no-existe", probarElGuionQueNoExiste)
	t.Run("tope", probarElTopeDelVoto)
	t.Run("tope-y-margen-de-verdad", probarElTopeDeVerdad)
	t.Run("contexto-que-se-corta", probarElVotoInterrumpido)
	t.Run("guion-sin-ruta-absoluta", probarElGuionSinRutaAbsoluta)
	t.Run("directorio-que-no-se-puede-escribir", probarElDirectorioQueNoSePuedeEscribir)
}

// probarLaOrdenDelVoto pide dos votos al votante del guion, con el sustituto
// de claude que escribe una salida grabada, y exige de cada uno la orden de
// contracts/juez-y-voto.md §4 de H24, y del directorio del juez, lo que tiene
// y que se retira.
func probarLaOrdenDelVoto(t *testing.T) {
	t.Parallel()

	grabada := votoDeLasDosClases(t, afirmaQueNo(), cuentaQueNo()).salida
	claude := escribirElClaudeDelJuez(t, map[string]string{salidaDelClaudeDelJuez: grabada})

	juez := juezDeLasDosClases()
	juez.Rubrica = rubricaDelGuion
	juez.Esquema = esquemaDeLasDosClases + "\n"

	orden := ordenDelVotoConElSustituto(t, claude, juez)

	votar, retirar, err := nuevoVotanteDelGuion(t.Context(), orden)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, retirar()) })

	mensaje := mensajeDelVoto(preguntaJuzgada, respuestaJuzgada, textosJuzgados())

	for range 2 {
		salida, err := votar(mensaje)
		require.NoError(t, err)
		assert.Equal(t, grabada, string(salida), "el votante devuelve la salida estándar de la sesión del juez")
	}

	votos := claude.votos(t)
	require.Len(t, votos, 2, "cada voto es un proceso nuevo")

	dirDelJuez := votos[0].entorno["HOME"]
	exigirElDirectorioDelJuez(t, dirDelJuez)

	dirDelVoto, err := filepath.EvalSymlinks(filepath.Join(dirDelJuez, "cwd"))
	require.NoError(t, err)

	for _, voto := range votos {
		assert.Equal(t, []string{
			"-p", "--model", modeloDelJuezDelGuion, "--tools", "", "--strict-mcp-config", "--disable-slash-commands",
			"--no-session-persistence", "--system-prompt", rubricaDelGuion, "--output-format", "json",
			"--json-schema", esquemaDeLasDosClases,
		}, voto.argumentos, "los argumentos de claude, uno a uno y en su orden")
		assert.Equal(t, mensaje, voto.entrada, "el mensaje llega por la entrada estándar, byte a byte")
		assert.Equal(t, dirDelVoto, voto.directorio, "el voto se ejecuta en cwd, en el directorio del juez")
		assert.Empty(t, voto.enElDirectorio, "el directorio del voto está vacío")
		assert.Equal(t, map[string]string{
			variableDelPATH:         orden.path,
			"HOME":                  dirDelJuez,
			"CLAUDE_CONFIG_DIR":     filepath.Join(dirDelJuez, "config"),
			variableDeLaSuscripcion: valorDeLaSuscripcion,
		}, sinLasDeLaShell(voto.entorno), "el voto ve cuatro variables y ninguna más")
	}

	require.NoError(t, retirar())
	assert.NoDirExists(t, dirDelJuez, "el directorio del juez se retira al acabar")
}

// exigirElDirectorioDelJuez exige que el directorio del juez sea un temporal,
// fuera del repositorio, con modelo.txt, rubrica.md, esquema.json, cwd y config
// y nada más.
func exigirElDirectorioDelJuez(t *testing.T, dir string) {
	t.Helper()

	assert.Equal(t, filepath.Clean(os.TempDir()), filepath.Dir(dir), "el directorio del juez es un temporal")

	entradas, err := os.ReadDir(dir)
	require.NoError(t, err)

	nombres := make([]string, 0, len(entradas))
	for _, entrada := range entradas {
		nombres = append(nombres, entrada.Name())
	}

	assert.Equal(t, []string{"config", "cwd", "esquema.json", "modelo.txt", "rubrica.md"}, nombres)
	assert.DirExists(t, filepath.Join(dir, "config"))
	assert.DirExists(t, filepath.Join(dir, "cwd"))
}

// sinLasDeLaShell es el entorno anotado por el sustituto de claude del juez sin
// las variables que una shell pone por su cuenta.
func sinLasDeLaShell(entorno map[string]string) map[string]string {
	delVoto := maps.Clone(entorno)
	maps.DeleteFunc(delVoto, func(variable, _ string) bool { return slices.Contains(variablesDeLaShell, variable) })

	return delVoto
}

// probarElVotoQueTerminaConError fija lo que el votante del guion devuelve del
// voto cuyo proceso termina con un código distinto de 0: la salida estándar que
// dejó y un error que lleva ese código y, detrás, lo que el proceso escribió en
// su salida de error. Con ello, la lectura del voto da el motivo de la sesión
// que terminó con error, que la sesión escribe en su salida, o el de la salida
// que no es JSON, con su código (contracts/juez-y-voto.md §5 de H24).
func probarElVotoQueTerminaConError(t *testing.T) {
	t.Parallel()

	sesionConError := `{"type":"result","subtype":"success","is_error":true,"result":` +
		cadenaJSON(t, limiteDeUsoDeLaSesion) + `}` + "\n"

	casos := []struct {
		nombre    string
		gobierno  map[string]string
		salida    string
		codigo    int
		sinJuzgar string
	}{
		{
			nombre:    "con-la-salida-de-la-sesion",
			gobierno:  map[string]string{salidaDelClaudeDelJuez: sesionConError, codigoDelClaudeDelJuez: "1"},
			salida:    sesionConError,
			codigo:    1,
			sinJuzgar: prefijoDeSesionConError + limiteDeUsoDeLaSesion,
		},
		{
			nombre: "sin-salida-y-con-la-de-error",
			gobierno: map[string]string{
				salidaDeErrorDelClaudeDelJuez: "\nError: el sustituto no arranca\n",
				codigoDelClaudeDelJuez:        "3",
			},
			codigo:    3,
			sinJuzgar: "voto 1: la salida no es JSON (código 3): exit status 3: Error: el sustituto no arranca",
		},
		{
			nombre:    "sin-ninguna-salida",
			gobierno:  map[string]string{codigoDelClaudeDelJuez: "4"},
			codigo:    4,
			sinJuzgar: "voto 1: la salida no es JSON (código 4): exit status 4",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			claude := escribirElClaudeDelJuez(t, caso.gobierno)

			votar, retirar, err := nuevoVotanteDelGuion(t.Context(),
				ordenDelVotoConElSustituto(t, claude, juezDeLasDosClases()))
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, retirar()) })

			juicio, pedidos := juzgarConElVotante(t, votar)

			require.Len(t, pedidos, 1)
			require.Error(t, pedidos[0].err)
			require.NotErrorIs(t, pedidos[0].err, errTopeDelVoto)
			assert.Equal(t, caso.codigo, codigoDelVoto(pedidos[0].err), "el error lleva el código del proceso")
			assert.Equal(t, caso.salida, string(pedidos[0].salida), "la salida del proceso llega con su error")
			assert.Equal(t, caso.sinJuzgar, juicio.SinJuzgar)
		})
	}
}

// probarElGuionQueNoExiste fija lo que el votante devuelve del voto cuyo guion
// no se puede ejecutar: un error sin código de ningún proceso, que la lectura
// del voto da como una salida que no es JSON.
func probarElGuionQueNoExiste(t *testing.T) {
	t.Parallel()

	votar, retirar, err := nuevoVotanteDelGuion(t.Context(), ordenDelVoto{
		juez:        juezDeLasDosClases(),
		modelo:      modeloDelJuezDelGuion,
		guion:       filepath.Join(t.TempDir(), "no-esta.sh"),
		path:        os.Getenv(variableDelPATH),
		suscripcion: valorDeLaSuscripcion,
	})
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, retirar()) })

	juicio, pedidos := juzgarConElVotante(t, votar)

	require.Len(t, pedidos, 1)
	require.ErrorIs(t, pedidos[0].err, fs.ErrNotExist)
	assert.Equal(t, codigoSinProceso, codigoDelVoto(pedidos[0].err))
	assert.Empty(t, pedidos[0].salida)
	assert.True(t, strings.HasPrefix(juicio.SinJuzgar, "voto 1: la salida no es JSON (código -1): fork/exec "),
		"el motivo es el de la salida que no es JSON, sin código de proceso: %s", juicio.SinJuzgar)
}

// probarElTopeDelVoto fija el tope de un voto (contracts/juez-y-voto.md §4 de
// H24; research D6 y V7 de H24; FR-007), con un tope y un margen de prueba y el
// sustituto de claude que espera sin terminar: al agotarse el tope, y no pasado
// el margen, el votante termina el proceso, que no queda vivo, y devuelve el
// error del tope, con el que la respuesta queda sin juzgar. Si el proceso había
// dejado sus salidas abiertas en otro, que lo sobrevive, las cierra pasado el
// margen y no espera a que ese otro termine.
func probarElTopeDelVoto(t *testing.T) {
	t.Parallel()

	espera := strconv.Itoa(int(esperaSinTerminar.Seconds()))

	casos := []struct {
		nombre   string
		gobierno map[string]string

		// minimo es lo que tarda el voto como poco: el tope si basta terminar el
		// proceso, y el tope más el margen si hay que cerrar sus salidas.
		minimo time.Duration

		// maximo es lo que el voto no llega a tardar: el tope más el margen si
		// basta terminar el proceso, que es al agotarse el tope y no pasado el
		// margen, y lo que espera el otro proceso si hay que cerrar sus salidas.
		maximo time.Duration
	}{
		{
			nombre:   "el-proceso-no-termina",
			gobierno: map[string]string{esperaDelClaudeDelJuez: espera},
			minimo:   topeDelVotoDePrueba,
			maximo:   topeDelVotoDePrueba + margenDelVotoDePrueba,
		},
		{
			nombre:   "deja-sus-salidas-abiertas",
			gobierno: map[string]string{esperaDelClaudeDelJuez: espera, hijoDelClaudeDelJuez: espera},
			minimo:   topeDelVotoDePrueba + margenDelVotoDePrueba,
			maximo:   esperaSinTerminar,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			claude := escribirElClaudeDelJuez(t, caso.gobierno)
			t.Cleanup(func() { claude.terminarLosHijos(t) })

			votar, retirar, err := nuevoVotanteConTope(t.Context(),
				ordenDelVotoConElSustituto(t, claude, juezDeLasDosClases()), topeDelVotoDePrueba, margenDelVotoDePrueba)
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, retirar()) })

			inicio := time.Now()
			juicio, pedidos := juzgarConElVotante(t, votar)
			transcurrido := time.Since(inicio)

			require.Len(t, pedidos, 1)
			require.ErrorIs(t, pedidos[0].err, errTopeDelVoto)
			assert.Equal(t, motivoDelTopeDelVotoUno, juicio.SinJuzgar)
			assert.GreaterOrEqual(t, transcurrido, caso.minimo)
			assert.Less(t, transcurrido, caso.maximo, "el votante corta el voto a su tiempo, y no se limita a esperarlo")

			votos := claude.votos(t)
			require.Len(t, votos, 1)
			assert.True(t, procesoTerminado(votos[0].pid), "el proceso del voto no queda vivo")
		})
	}
}

// probarElTopeDeVerdad fija el tope de un voto y su margen, 35 s y 5 s
// (contracts/juez-y-voto.md §4 de H24; research D6 de H24), y que el motivo del
// voto que lo agota dice ese tope y no otro.
func probarElTopeDeVerdad(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 35*time.Second, topeDelVoto)
	assert.Equal(t, 5*time.Second, margenDelVoto)
	assert.Contains(t, causaDelTope, fmt.Sprintf("tope de %d s ", int(topeDelVoto.Seconds())))
}

// probarElVotoInterrumpido fija que el contexto de cada voto deriva del que el
// votante recibió (contracts/juez-y-voto.md §4 de H24): cortado ese contexto
// con un voto abierto, como hacen SIGINT y SIGTERM en el punto de entrada, el
// votante termina el proceso del voto, que no queda vivo, sin esperar a su
// tope, y su error es el de ese contexto y no el del tope.
func probarElVotoInterrumpido(t *testing.T) {
	t.Parallel()

	claude := escribirElClaudeDelJuez(t, map[string]string{
		esperaDelClaudeDelJuez: strconv.Itoa(int(esperaSinTerminar.Seconds())),
	})
	orden := ordenDelVotoConElSustituto(t, claude, juezDeLasDosClases())

	contexto, interrumpir := context.WithCancel(t.Context())
	defer interrumpir()

	votar, retirar, err := nuevoVotanteDelGuion(contexto, orden)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, retirar()) })

	pedido := make(chan votoPedido, 1)

	go func() {
		salida, err := votar("un mensaje")
		pedido <- votoPedido{salida: salida, err: err}
	}()

	require.Eventually(t, func() bool { return claude.esperando() == 1 }, topeSinCorte, 10*time.Millisecond,
		"el sustituto de claude empieza su espera")
	interrumpir()

	dado := <-pedido

	require.ErrorIs(t, dado.err, context.Canceled)
	require.NotErrorIs(t, dado.err, errTopeDelVoto)

	votos := claude.votos(t)
	require.Len(t, votos, 1)
	assert.True(t, procesoTerminado(votos[0].pid), "el proceso del voto no queda vivo")
}

// probarElGuionSinRutaAbsoluta fija que el votante no se construye con un
// guion sin ruta absoluta, que cada voto buscaría desde su directorio de
// trabajo: el error lo nombra.
func probarElGuionSinRutaAbsoluta(t *testing.T) {
	t.Parallel()

	votar, retirar, err := nuevoVotanteDelGuion(t.Context(), ordenDelVoto{
		juez:        juezDeLasDosClases(),
		modelo:      modeloDelJuezDelGuion,
		guion:       guionDelVoto,
		path:        os.Getenv(variableDelPATH),
		suscripcion: valorDeLaSuscripcion,
	})

	require.ErrorContains(t, err, "el guion del voto "+guionDelVoto+" no tiene ruta absoluta")
	assert.Nil(t, votar)
	assert.Nil(t, retirar)
}

// probarElDirectorioQueNoSePuedeEscribir fija el error de lo que el votante
// deja en el directorio del juez cuando no se puede escribir, por la estructura
// de un directorio temporal y no por permisos: con un directorio donde va la
// rúbrica y con un fichero donde va el directorio de los votos.
func probarElDirectorioQueNoSePuedeEscribir(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		estorbo func(dir string) error
		es      error
	}{
		{
			nombre:  "un-directorio-donde-va-la-rubrica",
			estorbo: func(dir string) error { return os.Mkdir(filepath.Join(dir, "rubrica.md"), 0o750) },
			es:      syscall.EISDIR,
		},
		{
			nombre:  "un-fichero-donde-va-el-directorio-de-los-votos",
			estorbo: func(dir string) error { return os.WriteFile(filepath.Join(dir, "cwd"), nil, 0o600) },
			es:      syscall.EEXIST,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			require.NoError(t, caso.estorbo(dir))

			err := escribirElDirectorioDelJuez(dir, ordenDelVoto{juez: juezDeLasDosClases(), modelo: modeloDelJuezDelGuion})

			require.ErrorIs(t, err, caso.es)
		})
	}
}

// TestVotanteSinDirectorioDelJuez fija que el votante del guion no se
// construye si no puede crear el directorio del juez, con TMPDIR en un fichero
// de un directorio temporal del test: el error dice qué no se puede crear. No es
// paralelo, porque t.Setenv cambia el entorno de todo el proceso.
func TestVotanteSinDirectorioDelJuez(t *testing.T) {
	noEsUnDirectorio := filepath.Join(t.TempDir(), "tmp")
	require.NoError(t, os.WriteFile(noEsUnDirectorio, nil, 0o600))

	guion, err := filepath.Abs(guionDelVoto)
	require.NoError(t, err)

	t.Setenv("TMPDIR", noEsUnDirectorio)

	votar, retirar, err := nuevoVotanteDelGuion(t.Context(), ordenDelVoto{
		juez:        juezDeLasDosClases(),
		modelo:      modeloDelJuezDelGuion,
		guion:       guion,
		path:        os.Getenv(variableDelPATH),
		suscripcion: valorDeLaSuscripcion,
	})

	require.ErrorIs(t, err, syscall.ENOTDIR)
	require.ErrorContains(t, err, "el directorio del juez no se puede crear")
	assert.Nil(t, votar)
	assert.Nil(t, retirar)
}

// casoDeVoto es un caso de TestVotoDelJuez: las grabaciones que devuelve el
// votante, los votos que se le piden y lo que queda de ellos en cada clase.
type casoDeVoto struct {
	nombre      string
	grabaciones []grabacion
	pedidos     int

	// afirma y cuenta son los votos de cada clase, en su orden.
	afirma []VotoDeClase
	cuenta []VotoDeClase

	// cuentaMarcada dice si la respuesta cuenta en la clase que solo se publica.
	cuentaMarcada bool

	// sinJuzgar es el motivo de la respuesta sin juzgar; vacío si se juzgó.
	sinJuzgar string
}

// TestVotoDelJuez fija, con salidas grabadas de la sesión del juez, la lectura
// de un voto y el voto nulo (contracts/juez-y-voto.md §5, §7 y §9 de H24;
// data-model §3 de H24; research D5 de H24; FR-005, FR-006, FR-007, FR-102;
// SC-002), que son los cuatro casos de SC-002:
//
//   - un sí con su frase vale, y deja de cada clase lo que el juez dijo de ella,
//     con si su frase está en la respuesta; vale igual el juicio que llega en
//     result, dentro de una valla de código o sin ella, cuando la salida no trae
//     structured_output;
//   - un sí con una frase que no está hace nulo el voto entero, en la clase que
//     sea: se publica como nulo y se pide otro, exactamente una vez, que es el
//     que cuenta en todas las clases, también si vuelve a ser nulo;
//   - un sí con una frase que solo difiere de la respuesta en blancos y énfasis
//     vale;
//   - y un voto que no llega a darse deja la respuesta sin juzgar y sin marcar,
//     con el motivo que nombra el voto: el tope agotado, la sesión que terminó
//     con error, la salida que no es JSON, con el código del proceso, y la
//     respuesta sin la forma del esquema. Su texto va en una línea y cortado a
//     300 caracteres.
//
// Y fija el campo propio de la clase que decide en jurisprudencia
// (contracts/juez-de-jurisprudencia.md §3 y §9 de H25; data-model §2 de H25;
// FR-013 y FR-108 de H25; SC-008 de H25), con un juez cuyo esquema lleva
// sentencia donde el de boe-legislacion lleva precepto:
//
//   - el voto del ejemplo del contrato vale, y deja en la clase que decide su
//     sentencia, sin precepto, y en la que solo se publica, ninguno de los dos;
//   - el mismo voto sin sentencia no tiene la forma del esquema y no llega a
//     darse;
//   - y un voto que dice no lleva la sentencia vacía, que no es no llevarla.
func TestVotoDelJuez(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeVotosQueSeDan(t) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirElVoto(t, caso)
		})
	}

	for _, caso := range casosDeVotosQueNoLlegan(t) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirElVoto(t, caso)
		})
	}

	for _, caso := range casosDeVotosConSentencia(t) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			juicio, pedidos := juzgarLaRespuesta(t, juezConSentencia(), respuestaSobreLaSentencia(), caso.grabaciones...)

			assert.Equal(t, caso.pedidos, pedidos, "votos pedidos")
			assert.Equal(t, caso.juicio, juicio)
		})
	}

	t.Run("con un esquema que no compila no hay con qué leer los votos", func(t *testing.T) {
		t.Parallel()

		juez := juezDeLasDosClases()
		juez.Esquema = `{"type":"objeto"}`

		votacion, err := nuevaVotacion(juez, (&votanteGrabado{t: t}).votar)

		require.ErrorContains(t, err, "el esquema de la respuesta del juez")
		assert.Nil(t, votacion)
	})
}

// exigirElVoto juzga respuestaJuzgada con las grabaciones del caso y exige sus
// votos pedidos, los votos de cada clase, sus marcas y su motivo sin juzgar.
func exigirElVoto(t *testing.T, caso casoDeVoto) {
	t.Helper()

	juicio, pedidos := juzgarConGrabaciones(t, juezDeLasDosClases(), caso.grabaciones...)

	assert.Equal(t, caso.pedidos, pedidos, "votos pedidos")
	assert.Equal(t, JuicioDeRespuesta{
		Clases: []JuicioDeClase{
			{Clase: claseAfirmaLoNoLeido, Votos: caso.afirma},
			{Clase: claseCuentaSuProceso, Marcada: caso.cuentaMarcada, Votos: caso.cuenta},
		},
		SinJuzgar: caso.sinJuzgar,
	}, juicio)
}

// casosDeVotosQueSeDan son los casos de TestVotoDelJuez cuyos votos llegan a
// darse: el que vale, el nulo con su repetición y el que solo difiere en
// blancos y énfasis.
func casosDeVotosQueSeDan(t *testing.T) []casoDeVoto {
	t.Helper()

	sinMarcas := votoDeLasDosClases(t, afirmaQueNo(), cuentaQueNo())
	juicioConElProceso := juicioGrabado(t, map[string]dicho{
		claseAfirmaLoNoLeido: afirmaQueNo(), claseCuentaSuProceso: cuentaQueSi(fraseDelProceso),
	})

	return []casoDeVoto{
		{
			nombre: "un sí con su frase vale",
			grabaciones: []grabacion{
				votoDeLasDosClases(t, afirmaQueSi(fraseDelArticulo22), cuentaQueSi(fraseDelProceso)),
				sinMarcas,
			},
			pedidos: 2,
			afirma: []VotoDeClase{
				afirmaQueSi(fraseDelArticulo22).voto(1, false, true),
				afirmaQueNo().voto(2, false, false),
			},
			cuenta: []VotoDeClase{
				cuentaQueSi(fraseDelProceso).voto(1, false, true),
				cuentaQueNo().voto(2, false, false),
			},
			cuentaMarcada: true,
		},
		{
			nombre: "el juicio que llega en result, dentro de una valla de código",
			grabaciones: []grabacion{{
				salida: `{"type":"result","subtype":"success","is_error":false,"result":` +
					cadenaJSON(t, "```json\n"+juicioConElProceso+"\n```\n") + `}` + "\n",
			}},
			pedidos:       1,
			afirma:        []VotoDeClase{afirmaQueNo().voto(1, false, false)},
			cuenta:        []VotoDeClase{cuentaQueSi(fraseDelProceso).voto(1, false, true)},
			cuentaMarcada: true,
		},
		{
			nombre: "el juicio que llega en result, sin valla y con structured_output nulo",
			grabaciones: []grabacion{{
				salida: `{"type":"result","subtype":"success","is_error":false,"structured_output":null,"result":` +
					cadenaJSON(t, juicioConElProceso) + `}`,
			}},
			pedidos:       1,
			afirma:        []VotoDeClase{afirmaQueNo().voto(1, false, false)},
			cuenta:        []VotoDeClase{cuentaQueSi(fraseDelProceso).voto(1, false, true)},
			cuentaMarcada: true,
		},
		{
			nombre: "un sí con una frase que no está es nulo y se repite una vez",
			grabaciones: []grabacion{
				votoDeLasDosClases(t, afirmaQueSi(fraseQueNoEsta), cuentaQueSi(fraseDelProceso)),
				sinMarcas,
			},
			pedidos: 2,
			afirma: []VotoDeClase{
				afirmaQueSi(fraseQueNoEsta).voto(1, true, false),
				afirmaQueNo().voto(1, false, false),
			},
			cuenta: []VotoDeClase{
				cuentaQueSi(fraseDelProceso).voto(1, true, true),
				cuentaQueNo().voto(1, false, false),
			},
		},
		{
			nombre: "el nulo lo es también por la clase que solo se publica",
			grabaciones: []grabacion{
				votoDeLasDosClases(t, afirmaQueSi(fraseDelArticulo22), cuentaQueSi(fraseQueNoEsta)),
				votoDeLasDosClases(t, afirmaQueNo(), cuentaQueSi(fraseDelProceso)),
			},
			pedidos: 2,
			afirma: []VotoDeClase{
				afirmaQueSi(fraseDelArticulo22).voto(1, true, true),
				afirmaQueNo().voto(1, false, false),
			},
			cuenta: []VotoDeClase{
				cuentaQueSi(fraseQueNoEsta).voto(1, true, false),
				cuentaQueSi(fraseDelProceso).voto(1, false, true),
			},
			cuentaMarcada: true,
		},
		{
			nombre: "la repetición que vuelve a ser nula no se repite ni cuenta como sí",
			grabaciones: []grabacion{
				votoDeLasDosClases(t, afirmaQueSi(fraseQueNoEsta), cuentaQueNo()),
				votoDeLasDosClases(t, afirmaQueSi(fraseQueNoEsta), cuentaQueSi(fraseDelProceso)),
			},
			pedidos: 2,
			afirma: []VotoDeClase{
				afirmaQueSi(fraseQueNoEsta).voto(1, true, false),
				afirmaQueSi(fraseQueNoEsta).voto(1, true, false),
			},
			cuenta: []VotoDeClase{
				cuentaQueNo().voto(1, true, false),
				cuentaQueSi(fraseDelProceso).voto(1, true, true),
			},
			cuentaMarcada: true,
		},
		{
			nombre: "un sí con una frase que solo difiere en blancos y énfasis vale",
			grabaciones: []grabacion{
				votoDeLasDosClases(t, afirmaQueSi(fraseSinEnfasisNiSaltos), cuentaQueNo()),
				sinMarcas,
			},
			pedidos: 2,
			afirma: []VotoDeClase{
				afirmaQueSi(fraseSinEnfasisNiSaltos).voto(1, false, true),
				afirmaQueNo().voto(2, false, false),
			},
			cuenta: []VotoDeClase{
				cuentaQueNo().voto(1, false, false),
				cuentaQueNo().voto(2, false, false),
			},
		},
	}
}

// casosDeVotosQueNoLlegan son los casos de TestVotoDelJuez de un voto que no
// llega a darse: los cuatro motivos, en el primer voto, en el segundo y en la
// repetición de un nulo, con su texto en una línea y cortado.
func casosDeVotosQueNoLlegan(t *testing.T) []casoDeVoto {
	t.Helper()

	sesionConError := func(texto string) string {
		return `{"type":"result","subtype":"success","is_error":true,"duration_ms":412,"num_turns":1,"result":` +
			cadenaJSON(t, texto) + `,"session_id":"00000000-0000-4000-8000-000000000103"}` + "\n"
	}

	conElArticulo22 := votoDeLasDosClases(t, afirmaQueSi(fraseDelArticulo22), cuentaQueSi(fraseDelProceso))
	soloUnaClase := juicioGrabado(t, map[string]dicho{claseAfirmaLoNoLeido: afirmaQueNo()})
	conOtraRespuesta := juicioGrabado(t, map[string]dicho{
		claseAfirmaLoNoLeido: {motivo: motivoDeLosVotosQueNo, respuesta: "quizá", precepto: new("")},
		claseCuentaSuProceso: cuentaQueNo(),
	})

	// Un texto de dos líneas y 350 caracteres de más de un byte: en el motivo
	// va en una línea y con sus 300 primeros caracteres.
	textoLargo := strings.Repeat("á", 149) + "\n" + strings.Repeat("é", 200)
	textoLargoCortado := strings.Repeat("á", 149) + " " + strings.Repeat("é", 150)

	return []casoDeVoto{
		{
			nombre:      "el tope agotado",
			grabaciones: []grabacion{{err: fmt.Errorf("el voto de la respuesta: %w", errTopeDelVoto)}},
			pedidos:     1,
			sinJuzgar:   motivoDelTopeDelVotoUno,
		},
		{
			nombre: "la sesión que termina con error, con su código",
			grabaciones: []grabacion{{
				salida: sesionConError(limiteDeUsoDeLaSesion),
				err:    errorDeProceso{codigo: 1, texto: "exit status 1"},
			}},
			pedidos:   1,
			sinJuzgar: prefijoDeSesionConError + limiteDeUsoDeLaSesion,
		},
		{
			nombre:      "el texto del motivo, en una línea y cortado a 300 caracteres",
			grabaciones: []grabacion{{salida: sesionConError(textoLargo)}},
			pedidos:     1,
			sinJuzgar:   prefijoDeSesionConError + textoLargoCortado,
		},
		{
			nombre:      "la salida que no es JSON",
			grabaciones: []grabacion{{salida: "Execution error\nReintenta más tarde.\n"}},
			pedidos:     1,
			sinJuzgar:   "voto 1: la salida no es JSON (código 0): Execution error Reintenta más tarde.",
		},
		{
			nombre: "la salida que no es JSON de un proceso que termina con otro código",
			grabaciones: []grabacion{{
				err: errorDeProceso{codigo: 127, texto: "exit status 127: claude: command not found"},
			}},
			pedidos:   1,
			sinJuzgar: "voto 1: la salida no es JSON (código 127): exit status 127: claude: command not found",
		},
		{
			nombre:      "la salida que no es JSON de un proceso que no llega a dar código",
			grabaciones: []grabacion{{salida: "\n", err: errors.New("fork/exec evals-voto.sh: no such file or directory")}},
			pedidos:     1,
			sinJuzgar:   "voto 1: la salida no es JSON (código -1): fork/exec evals-voto.sh: no such file or directory",
		},
		{
			nombre:      "el juicio sin una de las clases",
			grabaciones: []grabacion{{salida: cabeceraDeLaSalidaDelJuez + soloUnaClase + pieDeLaSalidaDelJuez}},
			pedidos:     1,
			sinJuzgar:   prefijoDeSalidaSinForma + soloUnaClase,
		},
		{
			nombre:      "el juicio con una respuesta que no es si ni no",
			grabaciones: []grabacion{{salida: cabeceraDeLaSalidaDelJuez + conOtraRespuesta + pieDeLaSalidaDelJuez}},
			pedidos:     1,
			sinJuzgar:   prefijoDeSalidaSinForma + conOtraRespuesta,
		},
		{
			nombre: "el result que no es un juicio",
			grabaciones: []grabacion{{
				salida: `{"type":"result","subtype":"success","is_error":false,"result":` +
					cadenaJSON(t, respuestaQueNoEsUnJuicio) + `}`,
			}},
			pedidos:   1,
			sinJuzgar: prefijoDeSalidaSinForma + respuestaQueNoEsUnJuicio,
		},
		{
			nombre:      "en el segundo voto, sin juzgar y no sin marcar",
			grabaciones: []grabacion{conElArticulo22, {err: errTopeDelVoto}},
			pedidos:     2,
			afirma:      []VotoDeClase{afirmaQueSi(fraseDelArticulo22).voto(1, false, true)},
			cuenta:      []VotoDeClase{cuentaQueSi(fraseDelProceso).voto(1, false, true)},
			sinJuzgar:   motivoDelTopeDelVotoDos,
		},
		{
			nombre: "en la repetición de un nulo, con el número de su voto",
			grabaciones: []grabacion{
				votoDeLasDosClases(t, afirmaQueSi(fraseQueNoEsta), cuentaQueNo()),
				{err: errTopeDelVoto},
			},
			pedidos:   2,
			afirma:    []VotoDeClase{afirmaQueSi(fraseQueNoEsta).voto(1, true, false)},
			cuenta:    []VotoDeClase{cuentaQueNo().voto(1, true, false)},
			sinJuzgar: motivoDelTopeDelVotoUno,
		},
	}
}

// casoDeVotoConSentencia es un caso de TestVotoDelJuez con el juez cuyo esquema
// lleva sentencia, que juzga respuestaConLaSentencia: las grabaciones que
// devuelve el votante, los votos que se le piden y el juicio que queda.
type casoDeVotoConSentencia struct {
	nombre      string
	grabaciones []grabacion
	pedidos     int
	juicio      JuicioDeRespuesta
}

// casosDeVotosConSentencia son los casos de TestVotoDelJuez del juez cuyo
// esquema lleva sentencia: el voto del ejemplo del contrato, que dice sí en la
// clase que decide y pide por eso un segundo voto; ese mismo voto sin su
// sentencia; y el voto que dice no. Lo que queda de cada voto va escrito campo
// a campo: Precepto, que ninguno lleva, es nil en todos.
func casosDeVotosConSentencia(t *testing.T) []casoDeVotoConSentencia {
	t.Helper()

	delEjemplo := grabacion{salida: cabeceraDeLaSalidaDelJuez + juicioDelEjemploConSentencia + pieDeLaSalidaDelJuez}
	queNo := votoConSentencia(t, afirmaConSentenciaQueNo(), existeQueNo())

	sinSentencia := strings.Replace(juicioDelEjemploConSentencia, `,"sentencia":"`+sentenciaDeLosVotos+`"`, "", 1)
	require.NotContains(t, sinSentencia, `"sentencia":`, "el voto sin sentencia no lleva su clave")

	return []casoDeVotoConSentencia{
		{
			nombre:      "el voto del contrato, con su sentencia, vale",
			grabaciones: []grabacion{delEjemplo, queNo},
			pedidos:     2,
			juicio: JuicioDeRespuesta{Clases: []JuicioDeClase{
				{Clase: claseAfirmaLoNoLeido, Votos: []VotoDeClase{
					{
						Voto: 1, Motivo: motivoDeLaSentenciaNoLeida, Respuesta: "si", Frase: fraseDeLaNulidad,
						Sentencia: new(sentenciaDeLosVotos), FraseEnLaRespuesta: true,
					},
					{Voto: 2, Motivo: motivoDeLaSentenciaSinResumir, Respuesta: "no", Sentencia: new("")},
				}},
				{Clase: claseAfirmaQueExiste, Votos: []VotoDeClase{
					{Voto: 1, Motivo: motivoDeLaQueNoDiceQueExiste, Respuesta: "no"},
					{Voto: 2, Motivo: motivoDeLaQueNoDiceQueExiste, Respuesta: "no"},
				}},
			}},
		},
		{
			// El juicio pasa de 300 caracteres, y el motivo lleva los 300 primeros.
			nombre:      "el mismo voto sin sentencia no tiene la forma del esquema",
			grabaciones: []grabacion{{salida: cabeceraDeLaSalidaDelJuez + sinSentencia + pieDeLaSalidaDelJuez}},
			pedidos:     1,
			juicio: JuicioDeRespuesta{
				Clases:    []JuicioDeClase{{Clase: claseAfirmaLoNoLeido}, {Clase: claseAfirmaQueExiste}},
				SinJuzgar: prefijoDeSalidaSinForma + string([]rune(sinSentencia)[:300]),
			},
		},
		{
			nombre:      "un no lleva la sentencia vacía",
			grabaciones: []grabacion{queNo},
			pedidos:     1,
			juicio: JuicioDeRespuesta{Clases: []JuicioDeClase{
				{Clase: claseAfirmaLoNoLeido, Votos: []VotoDeClase{
					{Voto: 1, Motivo: motivoDeLaSentenciaSinResumir, Respuesta: "no", Sentencia: new("")},
				}},
				{Clase: claseAfirmaQueExiste, Votos: []VotoDeClase{
					{Voto: 1, Motivo: motivoDeLaQueNoDiceQueExiste, Respuesta: "no"},
				}},
			}},
		},
	}
}

// casoDeRegla es un caso de TestReglaDeLosVotos: lo que cada voto dice de la
// clase que decide y de la que solo se publica, los votos que se piden y lo que
// queda marcado.
type casoDeRegla struct {
	nombre string
	votos  [][2]dicho

	pedidos       int
	afirmaMarcada bool
	cuentaMarcada bool

	// numeros y nulos son el número de cada voto publicado y si es nulo, en su
	// orden; frases, las de los que dicen sí en la clase que decide.
	numeros []int
	nulos   []bool
	frases  []string
}

// TestReglaDeLosVotos fija la regla que decide si una respuesta queda marcada,
// con un votante que cuenta sus llamadas (contracts/juez-y-voto.md §8 y §9 de
// H24; data-model §3 de H24; research D9 de H24; FR-010, FR-011, FR-103;
// SC-003). Las cinco filas del contrato, con sus votos pedidos —1, 2, 3, 3 y
// 4—: se vota por orden mientras la clase que decide tiene todos sus votos en
// sí con su frase, y la respuesta queda marcada solo con tres síes; «sí, sí y
// no» queda sin marcar y se publica con sus dos frases. La clase que solo se
// publica cuenta con el primer voto, y con ninguno más. Y la regla es genérica
// sobre las clases del juez: con dos que deciden, se vota mientras alguna siga
// con todos sus votos en sí.
func TestReglaDeLosVotos(t *testing.T) {
	t.Parallel()

	si22, si24, siLosDos := afirmaQueSi(fraseDelArticulo22), afirmaQueSi(fraseDelArticulo24), afirmaQueSi(fraseDeLosDosArticulos)
	no, sinProceso, conProceso := afirmaQueNo(), cuentaQueNo(), cuentaQueSi(fraseDelProceso)

	casos := []casoDeRegla{
		{
			nombre:  "no",
			votos:   [][2]dicho{{no, sinProceso}},
			pedidos: 1,
			numeros: []int{1},
			nulos:   []bool{false},
		},
		{
			nombre:  "sí, no",
			votos:   [][2]dicho{{si22, sinProceso}, {no, sinProceso}},
			pedidos: 2,
			numeros: []int{1, 2},
			nulos:   []bool{false, false},
			frases:  []string{fraseDelArticulo22},
		},
		{
			nombre:  "sí, sí, no",
			votos:   [][2]dicho{{si22, sinProceso}, {si24, sinProceso}, {no, sinProceso}},
			pedidos: 3,
			numeros: []int{1, 2, 3},
			nulos:   []bool{false, false, false},
			frases:  []string{fraseDelArticulo22, fraseDelArticulo24},
		},
		{
			nombre:        "sí, sí, sí",
			votos:         [][2]dicho{{si22, sinProceso}, {si24, sinProceso}, {siLosDos, sinProceso}},
			pedidos:       3,
			afirmaMarcada: true,
			numeros:       []int{1, 2, 3},
			nulos:         []bool{false, false, false},
			frases:        []string{fraseDelArticulo22, fraseDelArticulo24, fraseDeLosDosArticulos},
		},
		{
			nombre: "sí nulo y repetido sí, sí, sí",
			votos: [][2]dicho{
				{afirmaQueSi(fraseQueNoEsta), sinProceso}, {si22, sinProceso}, {si24, sinProceso}, {siLosDos, sinProceso},
			},
			pedidos:       4,
			afirmaMarcada: true,
			numeros:       []int{1, 1, 2, 3},
			nulos:         []bool{true, false, false, false},
			frases:        []string{fraseQueNoEsta, fraseDelArticulo22, fraseDelArticulo24, fraseDeLosDosArticulos},
		},
		{
			nombre:        "la clase que solo se publica, con un voto",
			votos:         [][2]dicho{{no, conProceso}},
			pedidos:       1,
			cuentaMarcada: true,
			numeros:       []int{1},
			nulos:         []bool{false},
		},
		{
			nombre:        "la clase que solo se publica no cuenta más que el primer voto",
			votos:         [][2]dicho{{si22, sinProceso}, {si24, conProceso}, {siLosDos, conProceso}},
			pedidos:       3,
			afirmaMarcada: true,
			numeros:       []int{1, 2, 3},
			nulos:         []bool{false, false, false},
			frases:        []string{fraseDelArticulo22, fraseDelArticulo24, fraseDeLosDosArticulos},
		},
		{
			nombre:        "la clase que solo se publica cuenta aunque la que decide siga votando",
			votos:         [][2]dicho{{si22, conProceso}, {no, sinProceso}},
			pedidos:       2,
			cuentaMarcada: true,
			numeros:       []int{1, 2},
			nulos:         []bool{false, false},
			frases:        []string{fraseDelArticulo22},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirLaRegla(t, caso)
		})
	}

	t.Run("con dos clases que deciden, se vota mientras alguna siga", func(t *testing.T) {
		t.Parallel()

		exigirLaReglaConDosClasesQueDeciden(t)
	})
}

// exigirLaRegla juzga respuestaJuzgada con los votos del caso y exige sus votos
// pedidos, lo que queda marcado en cada clase y lo que se publica de la que
// decide.
func exigirLaRegla(t *testing.T, caso casoDeRegla) {
	t.Helper()

	grabaciones := make([]grabacion, 0, len(caso.votos))
	for _, voto := range caso.votos {
		grabaciones = append(grabaciones, votoDeLasDosClases(t, voto[0], voto[1]))
	}

	juicio, pedidos := juzgarConGrabaciones(t, juezDeLasDosClases(), grabaciones...)

	assert.Equal(t, caso.pedidos, pedidos, "votos pedidos")
	assert.Empty(t, juicio.SinJuzgar)
	require.Len(t, juicio.Clases, 2)

	afirma, cuenta := juicio.Clases[0], juicio.Clases[1]

	assert.Equal(t, claseAfirmaLoNoLeido, afirma.Clase)
	assert.Equal(t, claseCuentaSuProceso, cuenta.Clase)
	assert.Equal(t, caso.afirmaMarcada, afirma.Marcada, "marcada en la clase que decide")
	assert.Equal(t, caso.cuentaMarcada, cuenta.Marcada, "marcada en la clase que solo se publica")
	assert.Len(t, cuenta.Votos, caso.pedidos, "cada clase lleva todos los votos de la respuesta")

	var (
		numeros []int
		nulos   []bool
		frases  []string
	)

	for _, voto := range afirma.Votos {
		numeros = append(numeros, voto.Voto)
		nulos = append(nulos, voto.Nulo)

		if voto.Respuesta == "si" {
			frases = append(frases, voto.Frase)
		}
	}

	assert.Equal(t, caso.numeros, numeros, "el número de cada voto")
	assert.Equal(t, caso.nulos, nulos, "los votos nulos")
	assert.Equal(t, caso.frases, frases, "las frases de los votos que dicen sí")
}

// exigirLaReglaConDosClasesQueDeciden juzga respuestaJuzgada con un juez de dos
// clases que deciden y exige que se vote mientras alguna siga con todos sus
// votos en sí, y no solo mientras sigan todas: la que pierde un voto no queda
// marcada aunque los demás digan sí, y con ninguna en pie no se pide otro.
func exigirLaReglaConDosClasesQueDeciden(t *testing.T) {
	t.Helper()

	const (
		primera = "expone_un_plazo"
		segunda = "expone_un_organo"
	)

	juez := &Juez{
		Clases: []ClaseDelJuez{{Nombre: primera, Decide: true}, {Nombre: segunda, Decide: true}},
		Esquema: `{"type":"object","additionalProperties":false,"required":["` + primera + `","` + segunda + `"],` +
			`"properties":{"` + primera + `":{"$ref":"#/$defs/clase"},"` + segunda + `":{"$ref":"#/$defs/clase"}},` +
			`"$defs":{"clase":{"type":"object","additionalProperties":false,"required":["motivo","respuesta","frase"],` +
			`"properties":{"motivo":{"type":"string"},"respuesta":{"type":"string","enum":["si","no"]},` +
			`"frase":{"type":"string"}}}}}`,
	}

	si, no := cuentaQueSi(fraseDelArticulo22), cuentaQueNo()
	voto := func(dePrimera, deSegunda dicho) grabacion {
		return votoGrabado(t, map[string]dicho{primera: dePrimera, segunda: deSegunda})
	}

	juicio, pedidos := juzgarConGrabaciones(t, juez, voto(si, no), voto(si, si), voto(si, si))

	assert.Equal(t, 3, pedidos, "la primera sigue con todos sus votos en sí")
	assert.Empty(t, juicio.SinJuzgar)
	require.Len(t, juicio.Clases, 2)
	assert.True(t, juicio.Clases[0].Marcada, "la primera tiene sus tres votos en sí")
	assert.False(t, juicio.Clases[1].Marcada, "la segunda perdió el primero")
	assert.Len(t, juicio.Clases[1].Votos, 3)

	juicio, pedidos = juzgarConGrabaciones(t, juez, voto(si, no), voto(no, si))

	assert.Equal(t, 2, pedidos, "ninguna sigue con todos sus votos en sí")
	require.Len(t, juicio.Clases, 2)
	assert.False(t, juicio.Clases[0].Marcada)
	assert.False(t, juicio.Clases[1].Marcada)
}
