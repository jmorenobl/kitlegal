package evals

import (
	"fmt"
	"path"
	"slices"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/core/territorio"
)

// Los dos códigos de salida estables con los que una invocación con consulta dice
// que no tuvo lo que pedía: 4, fuente no disponible, el de --offline sin la
// consulta en la caché, y 5, límite o términos de uso, el de la petición que el
// proxy del job rechaza (CLAUDE.md, «Contrato de resultados»; contrato job-de-evals
// §6). Una invocación que termina con uno de ellos está fuera de lo grabado
// (data-model §10.2).
const (
	codigoFuenteNoDisponible = 4
	codigoLimiteOTos         = 5
)

// verboArticulos es el otro verbo, junto a articulo, con el que se lee un bloque
// esperado: pide varios bloques de una norma (data-model §6.1).
const verboArticulos = "articulos"

// Principio de los motivos por los que una eval no pasa (data-model §10.2). El de
// la sesión sin terminar lo fija el contrato job-de-evals §5; los de un comando,
// una cita, un aviso, un hallazgo o un elemento del territorio ausentes y el de un
// comando prohibido ejecutado van seguidos de su texto, el mismo con el que los
// presenta el informe, que en un aviso es su código y, en un hallazgo, su clase
// (contrato de formato, juicio e informe §4 de H5.1; contrato de evals §2 de H6;
// contrato evals-y-skill §2 de H7 y de H7.1). Desde H7.4, el de una redacción
// modificada ausente va seguido de su texto (RedaccionEsperada.texto), y el de
// una skill que la eval dice que no se activa, que la nombra en medio, lo
// compone motivoDeLaQueNoSeActiva
// (contracts/evals-y-juicio.md §2 de H7.4). Desde H21, el de una orden de kitlegal
// en una sesión que no lo tiene en el PATH va seguido de la orden, y el de una
// cita en la respuesta de una eval sin binario ni servidor, de su texto
// (contracts/evals-en-dos-modos.md §4 de H21).
const (
	motivoDeSesionSinTerminar = "la sesión no terminó: "
	motivoDeComandoAusente    = "comando ausente: "
	motivoDeComandoProhibido  = "comando prohibido ejecutado: "
	motivoDeOrdenSinKitlegal  = "orden de kitlegal en una sesión sin kitlegal en el PATH: "
	motivoDeCitaAusente       = "cita ausente: "
	motivoDeAvisoAusente      = "aviso ausente: "
	motivoDeHallazgoAusente   = "forma de hallazgo ausente: "
	motivoDeRedaccionAusente  = "redacción modificada ausente: "
	motivoDeTerritorioAusente = "territorio ausente: "
	motivoDeCitaSinConsulta   = "cita en una respuesta sin consulta: "
	motivoDeOtroModelo        = "la sesión no declara el modelo que se le pidió: "
)

// Los dos motivos de texto fijo de la línea con la que la respuesta de una eval
// sin binario ni servidor dice que no ha consultado nada
// (contracts/evals-en-dos-modos.md §4 de H21): el de la línea que no está y el
// de la que está sin su dirección.
const (
	motivoDeLineaSinConsultaAusente = "línea ⚠ " + etiquetaSinConsulta + ": ausente"
	motivoDeLineaSinDireccion       = "la línea ⚠ " + etiquetaSinConsulta + ": no lleva " + direccionParaInstalar
)

// motivoDeLlamadasSinJuzgar es el principio del motivo de una sesión cuyas
// llamadas no se pueden juzgar porque el registro de applets del binario no se
// puede construir, al que sigue el error: sin él no hay verbo del que sacar la
// orden equivalente a ninguna, y la sesión no pasa en vacío.
const motivoDeLlamadasSinJuzgar = "las llamadas a las herramientas no se pueden juzgar: "

// Lo que el juicio sabe de una llamada a una herramienta del servidor MCP
// (contracts/evals-en-dos-modos.md §4 de H21).
const (
	// separadorDeHerramienta une el applet y el verbo en el nombre de una
	// herramienta, <applet>_<verbo> (FR-003 de H21).
	separadorDeHerramienta = "_"

	// terminadorDeLaLlamada es lo que cli.LineaDeLlamada pone entre las banderas
	// de una llamada y sus argumentos de posición, que no es un argumento.
	terminadorDeLaLlamada = "--"
)

// ResultadoDeEval es el juicio de una sesión con su eval (data-model §10.2): lo
// esperado y lo observado, el reparto de los comandos, las citas, los avisos y el
// territorio esperados, lo que hicieron las invocaciones de la sesión, por qué no
// pasa y si pasa. Sus claves JSON son las de cada eval de informe.json (contrato
// job-de-evals §5).
type ResultadoDeEval struct {
	// Sesion es el nombre del directorio de la sesión. Juzgar no lo conoce: lo
	// pone EscribirInforme.
	Sesion string `json:"sesion"`

	// Eval es el nombre del fichero de la eval con la que se juzga la sesión.
	Eval string `json:"eval"`

	// Modelo es el id del modelo con el que se pidió abrir la sesión, el de su
	// modelo.txt, y ModeloDeLaSesion, el que la propia sesión declara en su
	// transcript. Los dos los pone EscribirInforme: Juzgar no los conoce.
	Modelo           string `json:"modelo"`
	ModeloDeLaSesion string `json:"modelo_de_la_sesion"`

	// Modo es el modo con el que se juzga la sesión, lo que tenía para
	// consultar: orden, herramienta o, en la de una eval sin binario ni servidor,
	// ninguno (data-model §6 de H21).
	Modo Modo `json:"modo"`

	// Decide dice si la sesión pertenece a una serie que decide el veredicto; lo
	// pone EscribirInforme al repartir las sesiones en series (data-model §10.4).
	Decide bool `json:"decide"`

	// Activa dice si la eval espera que la skill se active, y Activada, si la
	// sesión la activó: la activación coincide si son iguales.
	Activa   bool `json:"activa"`
	Activada bool `json:"activada"`

	// ComandosEjecutados y ComandosAusentes reparten los comandos esperados, en
	// el orden de la eval, entre los que satisface alguna invocación de la sesión
	// y los que no (data-model §6.1), cada uno con su texto: bloque <applet>
	// <norma> <bloque>, <applet> <verbo> <norma>, <applet> buscar <términos…>,
	// <applet> resolver <municipio>, o <applet> check y, si la lleva, su norma.
	ComandosEjecutados []string `json:"comandos_ejecutados"`
	ComandosAusentes   []string `json:"comandos_ausentes"`

	// ComandosProhibidosEjecutados son los comandos prohibidos de la eval, en su
	// orden y con sus repeticiones, que ejecutó alguna invocación de la sesión,
	// cada uno una vez aunque lo ejecuten varias y con su texto: <applet> <verbo>
	// (contrato evals-y-skill §2 de H7). Vacío si la eval no prohíbe nada o si
	// ninguna invocación ejecutó lo que prohíbe.
	ComandosProhibidosEjecutados []string `json:"comandos_prohibidos_ejecutados"`

	// CitasEncontradas y CitasAusentes reparten las citas esperadas, en el orden
	// de la eval, entre las que están en la respuesta y las que no (data-model
	// §6.2), cada una con su texto: <norma> <bloque>.
	CitasEncontradas []string `json:"citas_encontradas"`
	CitasAusentes    []string `json:"citas_ausentes"`

	// AvisosEncontrados y AvisosAusentes reparten los avisos esperados, en el orden de la eval y con sus
	// repeticiones, entre los que la respuesta lleva con su forma fija y los que no (ExtraerAvisos), cada uno con su
	// código.
	AvisosEncontrados []string `json:"avisos_encontrados"`
	AvisosAusentes    []string `json:"avisos_ausentes"`

	// HallazgosEncontrados y HallazgosAusentes reparten los hallazgos esperados, en el orden de la eval y con sus
	// repeticiones, entre los que la respuesta lleva con su forma fija y los que no (ExtraerHallazgos), cada uno con su
	// clase (contrato evals-y-skill §2 de H7.1).
	HallazgosEncontrados []string `json:"hallazgos_encontrados"`
	HallazgosAusentes    []string `json:"hallazgos_ausentes"`

	// RedaccionesEncontradas y RedaccionesAusentes reparten las redacciones
	// modificadas esperadas, en el orden de la eval, entre las que la respuesta
	// traslada en una línea con la forma fija de version-obsoleta, la cita de su
	// bloque y sus dos fechas en su orden (ExtraerRedaccionesModificadas), y las
	// que no, cada una con su texto: <norma> <bloque> <fecha_vigencia>
	// <fecha_vigencia_reciente> (data-model §4 de H7.4; FR-053).
	RedaccionesEncontradas []string `json:"redacciones_modificadas_encontradas"`
	RedaccionesAusentes    []string `json:"redacciones_modificadas_ausentes"`

	// TerritorioEncontrado y TerritorioAusente reparten los elementos del
	// territorio esperado, en el orden de la eval —comunidad, provincia, cada
	// boletín y cada aspecto de cobertura— y con sus repeticiones, entre los que la
	// respuesta declara con su forma fija y los que no (ExtraerTerritorio), cada
	// uno con su texto: comunidad: <nombre>, provincia: <nombre>, boletín:
	// <código> o <aspecto>: <valor>.
	TerritorioEncontrado []string `json:"territorio_encontrado"`
	TerritorioAusente    []string `json:"territorio_ausente"`

	// LineaSinConsulta dice, en una eval sin binario ni servidor, si la respuesta
	// lleva una línea que empieza por ⚠ SIN CONSULTA AL BOE: con su dirección en
	// esa misma línea (ExtraerSinConsulta). Falso en las demás evals, que no la
	// juzgan (contracts/evals-en-dos-modos.md §4 de H21; FR-047).
	LineaSinConsulta bool `json:"linea_sin_consulta"`

	// CitasSinConsulta son, en una eval sin binario ni servidor, las citas de la
	// respuesta (ExtraerCitas), en su orden y con sus repeticiones, cada una con
	// su texto: <norma> <bloque>. Vacía si no lleva ninguna y en las demás evals.
	CitasSinConsulta []string `json:"citas_sin_consulta"`

	// Invocaciones son todas las invocaciones de applet de la sesión, en su
	// orden: las de su traza y, detrás, las que cuentan por sus llamadas a las
	// herramientas del servidor MCP.
	Invocaciones []InvocacionInformada `json:"invocaciones"`

	// FueraDeLoGrabado son las invocaciones con consulta y código 4 o 5, en su
	// orden: también las llamadas con un error de la clase fuente-no-disponible o
	// limite-o-tos.
	FueraDeLoGrabado []InvocacionFallida `json:"fuera_de_lo_grabado"`

	// OtrasFallidas son las invocaciones con consulta y otro código distinto de
	// 0, en su orden, y con ellas las llamadas con un error de otra clase; no las
	// que quedaron sin código porque el tope cortó la sesión antes de que
	// acabaran, ni las llamadas sin resultado.
	OtrasFallidas []InvocacionFallida `json:"otras_fallidas"`

	// LlegadasALaRed tienen una entrada por cada invocación y destino de sus
	// conexiones de clase red, en su orden.
	LlegadasALaRed []LlegadaALaRed `json:"llegadas_a_la_red"`

	// Respuesta es la de la sesión: vacía si no la hay o si la sesión no se pudo
	// leer.
	Respuesta string `json:"respuesta"`

	// CodigoDeLaSesion es el código de la sesión. Es nil solo si la sesión no se
	// pudo leer, porque el código de una sesión nunca es 0 por omisión.
	CodigoDeLaSesion *int `json:"codigo_de_la_sesion"`

	// FinDeLaSesion es el fin de la sesión con su texto fijo (data-model §10.1):
	// vacío si la sesión no se pudo leer.
	FinDeLaSesion string `json:"fin_de_la_sesion"`

	// SesionTerminada dice si la sesión terminó.
	SesionTerminada bool `json:"sesion_terminada"`

	// ReintentosPorLimiteDeRitmo son los reintentos por rate_limit que registra
	// el transcript de la sesión: 0 si no se pudo leer (FR-033 de H7.3).
	ReintentosPorLimiteDeRitmo int `json:"reintentos_por_limite_de_ritmo"`

	// SinMedir es vacío si la sesión se midió y, si un límite de uso de la cuenta
	// no la dejó terminar, la descripción de su clase (LimiteDeUso). Juzgar no
	// clasifica: lo hace EscribirInforme (data-model §3 y §4 de H7.3; FR-040).
	SinMedir string `json:"sin_medir"`

	// Motivos son las causas por las que la eval no pasa, una por causa y en este
	// orden: la sesión ilegible, que pone EscribirInforme, o sin terminar; la
	// activación que no coincide; cada skill que la eval dice que no se activa y
	// se activó; cada comando ausente; cada comando prohibido ejecutado; cada cita
	// ausente; cada aviso ausente; cada hallazgo ausente; cada redacción
	// modificada ausente; cada elemento del territorio ausente; y el modelo que la
	// sesión declara sin ser el pedido, que pone EscribirInforme. Vacío si pasa.
	// La sesión sin medir lleva solo el del límite (FR-040 de H7.3;
	// contracts/evals-y-juicio.md §2 de H7.4).
	//
	// Desde H21, detrás de los de los comandos prohibidos va el de cada orden de
	// kitlegal de una sesión que no lo tiene en el PATH; y, en una eval sin binario
	// ni servidor, detrás de los del territorio, el de la línea
	// ⚠ SIN CONSULTA AL BOE: ausente o sin su dirección y el de cada cita de la
	// respuesta
	// (contracts/evals-en-dos-modos.md §4 de H21). El de las llamadas que no se
	// pueden juzgar, que solo se da si el registro de applets del binario no se
	// puede construir, va delante de los de los comandos.
	//
	// Desde H23, detrás de los del territorio van los de lo que la eval espera
	// de las sentencias: cada cita de sentencia que falta, que sobra o que va
	// con un ROJ que no se cita; la línea ⚠ SENTENCIA NO COMPROBADA: que falta;
	// cada dirección, cada casilla y cada valor que faltan; y la dirección de
	// búsqueda que ninguna orden devolvió o que falta. El resultado no tiene
	// otra clave para ellos (contracts/evals-jurisprudencia.md §2 de H23).
	Motivos []string `json:"motivos"`

	// Pasa dice si la sesión terminó, la activación coincide, no se activó
	// ninguna skill que la eval dice que no se activa, no falta ningún comando,
	// ninguna cita, ningún aviso, ningún hallazgo, ninguna redacción modificada ni
	// ningún elemento del territorio esperados y no se ejecutó ningún comando
	// prohibido. No lo cambian FueraDeLoGrabado, OtrasFallidas ni LlegadasALaRed
	// (FR-076), ni la forma fija de un aviso o de un hallazgo, ni una línea de
	// redacción modificada, que la
	// eval no espera. Una sesión sin medir no pasa. Desde H21, tampoco pasa la
	// sesión sin kitlegal en el PATH que ejecuta una orden suya ni, en una eval
	// sin binario ni servidor, la respuesta sin la línea con su dirección o con
	// alguna cita. Desde H23, tampoco la que no cumple algo de lo que su eval
	// espera de las sentencias.
	Pasa bool `json:"pasa"`
}

// InvocacionInformada es una invocación de applet de la sesión como la presenta
// el informe (data-model §10.2).
type InvocacionInformada struct {
	// Orden es el applet seguido de los argumentos que le siguen en argv,
	// separados por un espacio; en una llamada a una herramienta, la herramienta
	// seguida de sus argumentos, los de la orden equivalente.
	Orden string `json:"orden"`

	// Codigo es el de la invocación, o nil en la que quedó sin código porque el
	// tope cortó la sesión (data-model §9). En una llamada, 0 si su resultado no
	// es un error, el de la clase de su error si lo es y nil si no tiene
	// resultado.
	Codigo *int `json:"codigo"`

	// Conexiones tiene una entrada por pareja distinta de destino y clase de las
	// conexiones de la invocación, en el orden en que aparece por primera vez;
	// vacía si la invocación no conectó. Las de una llamada no se conocen: son del
	// proceso del servidor, y se informan con él.
	Conexiones []ConexionInformada `json:"conexiones"`

	// Llamada dice si es una llamada a una herramienta del servidor MCP, leída
	// del transcript, y no una invocación de la traza.
	Llamada bool `json:"llamada"`
}

// ConexionInformada es una pareja de destino y clase de las conexiones de una
// invocación (data-model §9).
type ConexionInformada struct {
	// Destino es el de Conexion.Destino.
	Destino string `json:"destino"`

	// Clase es local, bloqueada o red.
	Clase ClaseDeConexion `json:"clase"`
}

// InvocacionFallida es una invocación con consulta que terminó con un código
// distinto de 0, con su orden y su código.
type InvocacionFallida struct {
	Orden  string `json:"orden"`
	Codigo int    `json:"codigo"`
}

// LlegadaALaRed es una invocación con una conexión de clase red, con su orden y
// el destino de esa conexión.
type LlegadaALaRed struct {
	Orden   string `json:"orden"`
	Destino string `json:"destino"`
}

// Juzgar compara una sesión con su eval sin ningún modelo (data-model §10.2;
// contrato evals-y-grabaciones §6; FR-072): si la sesión terminó, si la
// activación de la skill coincide con la esperada, qué comandos esperados
// satisfacen sus invocaciones (data-model §6.1), qué citas esperadas están en su
// respuesta (data-model §6.2) y qué avisos esperados lleva su respuesta con su
// forma fija, la marca, la etiqueta y los dos puntos que reconoce ExtraerAvisos
// (FR-030 a FR-033 de H5.1). Informa además de todas sus invocaciones, de las
// que quedaron fuera de lo grabado, de las otras fallidas y de las que llegaron a
// la red, sin que nada de eso cambie si la eval pasa (FR-076).
//
// Una sesión sin terminar no pasa aunque todo lo demás coincida: sin ella, una
// eval de no activación cuya sesión murió sin activar nada pasaría en vacío. Una
// eval sin avisos deja vacíos los encontrados y los ausentes, y su juicio es el de
// antes de H5.1 (FR-034).
//
// Desde H6, reparte además los elementos del territorio esperado entre los que la
// respuesta declara con su forma fija (ExtraerTerritorio) y los ausentes, y un
// ausente impide pasar; una eval sin territorio esperado deja vacíos los dos y su
// juicio es el de antes (contrato de evals §2 de H6; FR-084).
//
// Desde H7, anota además cada comando prohibido de la eval que ejecuta alguna
// invocación de la sesión que consulta, termine como termine, y uno ejecutado
// impide pasar; el comando de comprobación lo satisface, como los demás, solo
// una invocación que consulta y termina con 0. Una eval sin prohibidos deja vacía
// la lista y su juicio es el de antes (contrato evals-y-skill §2 de H7; FR-085,
// FR-086).
//
// Desde H7.1, reparte además los hallazgos esperados entre los que la respuesta
// lleva con su forma fija (ExtraerHallazgos) y los ausentes, y un ausente impide
// pasar: decirlo con otras palabras no lo traslada. El comando de comprobación
// que lleva norma exige además que la invocación la consulte. Una eval sin
// hallazgos deja vacíos los dos y su juicio es el de antes (contrato
// evals-y-skill §2 de H7.1; FR-050 a FR-052).
//
// Desde H7.3, publica además los reintentos por rate_limit de la sesión, que no
// cambian su juicio (FR-033 y FR-041 de H7.3).
//
// Desde H7.4, anota además, con su motivo detrás del de la activación, cada skill
// de NoSeActivan que la sesión activó, y una activada impide pasar; y reparte las
// redacciones modificadas esperadas entre las que la respuesta traslada
// (ExtraerRedaccionesModificadas: la norma, el bloque y las dos fechas en su
// orden) y las ausentes, con su motivo detrás de los de los hallazgos, y una
// ausente impide pasar. Una eval sin esas claves deja vacías las dos listas y su
// juicio es el de antes (contracts/evals-y-juicio.md §2 de H7.4; FR-003, FR-052,
// FR-053).
//
// Desde H21, juzga la sesión con el modo que da su eval sola, el de una sesión
// sin servidor.json: el modo orden o, si la eval es sin binario ni servidor,
// ninguno. La sesión del modo herramienta la juzga juzgarEnModo, que dice lo que
// el juicio gana en ese hito (contracts/evals-en-dos-modos.md §4 de H21).
//
// Desde H24, no mira la lista de expresiones de la skill, con la que de H7.2 a
// H22 anotaba cada expresión que llevaba la respuesta: ninguna sesión deja de
// pasar por una, y el resultado no las lleva ni tiene un motivo por ellas
// (contracts/informe-del-job.md §6 de H24; FR-070).
func Juzgar(eval Eval, sesion Sesion, skill string) ResultadoDeEval {
	return juzgarEnModo(eval, sesion, skill, modoSinServidor(eval))
}

// modoSinServidor es el modo de una sesión que no tiene servidor.json, que da su
// eval sola (data-model §6 de H21): ninguno, el valor vacío, si la eval es sin
// binario ni servidor, y el modo orden en otro caso.
func modoSinServidor(eval Eval) Modo {
	if eval.SinBinarioNiServidor {
		return ""
	}

	return ModoOrden
}

// juzgarEnModo es Juzgar con el modo de la sesión, el que da su directorio
// (contracts/evals-en-dos-modos.md §4 de H21; research.md D17, D18 y D21 de H21;
// FR-041, FR-042, FR-046, FR-047):
//
//   - cada llamada de la sesión a una herramienta del servidor MCP cuenta como
//     una invocación (invocacionDeLaLlamada), detrás de las de la traza, de modo
//     que los comandos esperados, los prohibidos, lo que queda fuera de lo
//     grabado y las otras fallidas valen para ella con las reglas de la orden;
//   - el proceso del servidor, mcp serve en la traza, no es una consulta: no
//     satisface ni falla nada, y sus conexiones se informan con él;
//   - en una sesión sin kitlegal en el PATH, la del modo herramienta y la de una
//     eval sin binario ni servidor, cada invocación de la traza de otro applet que
//     el del servidor lleva su motivo, y la eval no pasa;
//   - y en una eval sin binario ni servidor, la respuesta tiene que llevar la
//     línea ⚠ SIN CONSULTA AL BOE: con su dirección y ninguna cita, y lo que
//     falte de eso lleva su motivo.
//
// Desde H23, juzga además lo que la eval espera de las sentencias de las que
// habla la respuesta (juzgarLasSentencias), con sus motivos detrás de los del
// territorio, y los dos comandos de cita los satisface la invocación de su
// verbo con el ROJ y el texto que el comando pida, sea una orden o una llamada
// (pideLaCita). Una eval sin sentencias y sin comandos de cita se juzga como
// antes (contracts/evals-jurisprudencia.md §2 de H23; FR-051, FR-052).
func juzgarEnModo(eval Eval, sesion Sesion, skill string, modo Modo) ResultadoDeEval {
	codigo := sesion.Codigo

	resultado := ResultadoDeEval{
		Eval:                       eval.Fichero,
		Modo:                       modo,
		Activa:                     eval.Activa,
		Activada:                   sesion.Activada(skill),
		Respuesta:                  sesion.Respuesta,
		CodigoDeLaSesion:           &codigo,
		FinDeLaSesion:              sesion.Fin,
		SesionTerminada:            sesion.Terminada,
		ReintentosPorLimiteDeRitmo: sesion.ReintentosPorLimiteDeRitmo(),
	}

	if !sesion.Terminada {
		resultado.Motivos = append(resultado.Motivos, motivoDeSesionSinTerminar+sesion.MotivoSinTerminar)
	}

	if resultado.Activa != resultado.Activada {
		resultado.Motivos = append(resultado.Motivos, motivoDeActivacion(skill, eval.Activa))
	}

	citas := ExtraerCitas(sesion.Respuesta)

	activadasSinDeber := resultado.anotarLasQueNoSeActivan(eval.NoSeActivan, sesion)
	invocaciones, conLasLlamadas := resultado.invocacionesDe(sesion)
	resultado.repartirComandos(eval.Comandos, invocaciones)
	resultado.anotarProhibidos(eval.Prohibidos, invocaciones)
	ordenesSinKitlegal := resultado.anotarLasOrdenesSinKitlegal(sesion.Invocaciones)
	resultado.repartirCitas(eval.Citas, citas)
	resultado.AvisosEncontrados, resultado.AvisosAusentes = resultado.repartirFormas(eval.Avisos,
		ExtraerAvisos(sesion.Respuesta), motivoDeAvisoAusente)
	resultado.HallazgosEncontrados, resultado.HallazgosAusentes = resultado.repartirFormas(eval.Hallazgos,
		ExtraerHallazgos(sesion.Respuesta), motivoDeHallazgoAusente)
	resultado.repartirRedacciones(eval.RedaccionesModificadas, ExtraerRedaccionesModificadas(sesion.Respuesta))
	resultado.repartirTerritorio(eval.Territorio, ExtraerTerritorio(sesion.Respuesta, eval.Territorio))
	sentenciasComoSeEsperan := resultado.juzgarLasSentencias(eval.Sentencias, sesion)
	sinConsultaComoSeEspera := resultado.juzgarLaRespuestaSinConsulta(eval, sesion.Respuesta, citas)

	for _, invocacion := range invocaciones {
		resultado.informarLaJuzgada(invocacion)
	}

	resultado.Pasa = sesion.Terminada && activadasSinDeber == 0 && conLasLlamadas && ordenesSinKitlegal == 0 &&
		sentenciasComoSeEsperan && sinConsultaComoSeEspera && resultado.cumpleLoEsperado()

	return resultado
}

// invocacionJuzgada es una invocación de la sesión como la ve el juicio
// (contracts/evals-en-dos-modos.md §4 de H21): la de la traza o la que cuenta
// por una llamada a una herramienta del servidor MCP.
type invocacionJuzgada struct {
	Invocacion

	// orden es la invocación como la presenta el informe: la de
	// ordenDeLaInvocacion en la de la traza y, en una llamada, la herramienta
	// seguida de sus argumentos.
	orden string

	// llamada dice si cuenta por una llamada a una herramienta.
	llamada bool
}

// invocacionesDe son las invocaciones con las que se juzga la sesión: las de su
// traza, en su orden, y detrás, las que cuentan por sus llamadas a las
// herramientas del servidor MCP, en el del transcript. Las dos listas no tienen
// un reloj común: el proceso del servidor arranca antes de la primera llamada, y
// por eso va delante.
//
// Dice además si las llamadas se han podido juzgar. Solo deja de poderse si el
// registro de applets del binario no se puede construir: entonces anota su
// motivo y devuelve solo las de la traza, y la sesión no pasa. Una sesión sin
// llamadas no lo necesita.
func (r *ResultadoDeEval) invocacionesDe(sesion Sesion) ([]invocacionJuzgada, bool) {
	juzgadas := make([]invocacionJuzgada, 0, len(sesion.Invocaciones)+len(sesion.Llamadas))

	for _, invocacion := range sesion.Invocaciones {
		juzgadas = append(juzgadas, invocacionDeLaTraza(invocacion))
	}

	if len(sesion.Llamadas) == 0 {
		return juzgadas, true
	}

	verbos, err := verbosDeLasHerramientas()
	if err != nil {
		r.Motivos = append(r.Motivos, motivoDeLlamadasSinJuzgar+err.Error())

		return juzgadas, false
	}

	for _, llamada := range sesion.Llamadas {
		juzgadas = append(juzgadas, invocacionDeLaLlamada(llamada, verbos[llamada.Herramienta]))
	}

	return juzgadas, true
}

// invocacionDeLaTraza es una invocación de la traza como la ve el juicio. La del
// proceso del servidor MCP, mcp serve, no es una consulta: sirve las llamadas,
// que son las que consultan y las que se juzgan, de modo que no satisface ni
// ejecuta nada ni va a las fallidas, termine como termine —lo normal es que el
// agente lo termine con una señal—, y sus conexiones se informan con ella
// (research.md D17 de H21; FR-041).
func invocacionDeLaTraza(invocacion Invocacion) invocacionJuzgada {
	if invocacion.Applet == appletDelServidor && invocacion.Verbo == verboDelServidor {
		invocacion.Consulta = false
	}

	return invocacionJuzgada{Invocacion: invocacion, orden: ordenDeLaInvocacion(invocacion)}
}

// invocacionDeLaLlamada es la invocación por la que cuenta una llamada a una
// herramienta, con el verbo de esa herramienta (contracts/evals-en-dos-modos.md
// §4 de H21; FR-042): el applet y el verbo son los de la herramienta; los
// argumentos, los de la orden equivalente; consulta siempre, porque una llamada
// no tiene ayuda, --describe ni --dry-run; y su código es el de esa orden
// (codigoDeLaLlamada). Una herramienta sin verbo en el registro no tiene applet
// ni orden equivalente: su llamada se informa, y no satisface ni ejecuta nada.
func invocacionDeLaLlamada(llamada Llamada, verbo cli.Verbo) invocacionJuzgada {
	argumentos := argumentosDeLaLlamada(verbo, llamada.Argumentos)

	return invocacionJuzgada{
		Invocacion: Invocacion{
			Applet:     verbo.Applet,
			Verbo:      verbo.Verbo,
			Argumentos: argumentos,
			Consulta:   true,
			Codigo:     codigoDeLaLlamada(llamada),
		},
		orden:   strings.Join(slices.Concat([]string{llamada.Herramienta}, argumentos), " "),
		llamada: true,
	}
}

// argumentosDeLaLlamada son los argumentos de una llamada como los de la orden
// equivalente: los de cli.LineaDeLlamada, que los da en el orden de la orden y
// no en el del objeto, sin el terminador que separa las banderas de los de
// posición. Ninguno si la llamada no convierte: una propiedad que el verbo no
// declara, un valor de otro tipo o unos argumentos que no son un objeto no
// tienen orden equivalente, y el servidor los rechaza sin ejecutar nada.
func argumentosDeLaLlamada(verbo cli.Verbo, argumentos []byte) []string {
	linea, err := cli.LineaDeLlamada(verbo, argumentos)
	if err != nil {
		return nil
	}

	// El terminador es el primer «--»: delante solo hay banderas con su valor,
	// --<nombre>=<valor>, y detrás, un argumento de posición puede serlo.
	if terminador := slices.Index(linea, terminadorDeLaLlamada); terminador >= 0 {
		linea = slices.Delete(linea, terminador, terminador+1)
	}

	return linea
}

// codigoDeLaLlamada es el código con el que habría terminado la orden
// equivalente a una llamada (data-model §2 de H21; research.md D17 de H21): 0 si
// su resultado no es un error; el de su clase si lo es, que es 1, el de lo
// inesperado, si la clase no se lee o no es del vocabulario; y ninguno si no
// tiene resultado, como la invocación que el tope dejó sin código. Sale de la
// tabla del kernel, cli.CodigoSalida, y no de otra escrita aquí.
func codigoDeLaLlamada(llamada Llamada) *int {
	if !llamada.ConResultado {
		return nil
	}

	codigo := cli.CodigoSalida(nil)
	if llamada.Error {
		codigo = cli.CodigoSalida(falloDeLaLlamada{clase: llamada.Clase})
	}

	return &codigo
}

// falloDeLaLlamada es el error de una llamada cuyo resultado es un sobre de
// fallo, con la clase que el sobre declara: con él se pregunta al kernel por el
// código de esa clase (schema.ConClase).
type falloDeLaLlamada struct {
	clase schema.Clase
}

func (f falloDeLaLlamada) Error() string {
	return fmt.Sprintf("la llamada devolvió un error de la clase %q", f.clase)
}

// Clase es la del sobre de fallo de la llamada.
func (f falloDeLaLlamada) Clase() schema.Clase {
	return f.clase
}

// verbosDeLasHerramientas da, por el nombre de cada herramienta que el servidor
// MCP anuncia con el registro de producción, su verbo como lo describe el
// kernel: el applet, el verbo y los argumentos de su fábrica, que es lo que
// cli.LineaDeLlamada necesita para convertir una llamada en la orden
// equivalente. Qué verbos dan herramienta lo dice app.NombresDeHerramientas, sin
// repetir aquí qué applets no las dan; los argumentos solo se reflejan, de modo
// que un valor por verbo vale para todos los juicios. El registro se construye
// una sola vez, con la versión vacía, la de quien no tiene ninguna.
//
// Todo verbo que da herramienta tiene fábrica de argumentos: sin ella el
// servidor no arranca (FR-002 a FR-004 de H21).
var verbosDeLasHerramientas = sync.OnceValues(func() (map[string]cli.Verbo, error) {
	registro, err := app.RegistroDeProduccion("")
	if err != nil {
		return nil, fmt.Errorf("el registro de applets del binario no se puede construir: %w", err)
	}

	herramientas := app.NombresDeHerramientas(registro)
	verbos := make(map[string]cli.Verbo, len(herramientas))

	for _, nombre := range registro.Nombres() {
		applet, _ := registro.Buscar(nombre)

		for _, verbo := range applet.Verbos() {
			herramienta := nombre + separadorDeHerramienta + verbo.Nombre
			if !slices.Contains(herramientas, herramienta) {
				continue
			}

			verbos[herramienta] = cli.Verbo{Applet: nombre, Verbo: verbo.Nombre, Argumentos: verbo.Argumentos()}
		}
	}

	return verbos, nil
})

// cumpleLoEsperado dice si la activación coincide con la esperada, no falta
// ningún comando, ninguna cita, ningún aviso, ningún hallazgo, ninguna redacción
// modificada ni ningún elemento del territorio esperados y no se ejecutó ningún
// comando prohibido.
func (r *ResultadoDeEval) cumpleLoEsperado() bool {
	return r.Activa == r.Activada &&
		len(r.ComandosAusentes) == 0 && len(r.ComandosProhibidosEjecutados) == 0 &&
		len(r.CitasAusentes) == 0 && len(r.AvisosAusentes) == 0 &&
		len(r.HallazgosAusentes) == 0 && len(r.RedaccionesAusentes) == 0 &&
		len(r.TerritorioAusente) == 0
}

// exigirElModeloPedido deja de pasar, con su motivo, la sesión que declara un
// modelo que no es el que se le pidió: si no, el informe publicaría como medida
// de un modelo lo que hizo otro (contrato job-de-evals §4; ADR 0016). El motivo va
// detrás de los de Juzgar, porque no es un defecto de la skill sino de la
// ejecución. El id declarado puede llevar detrás la fecha de la versión, porque el
// proveedor resuelve el alias que se pidió, pero nada más: claude-sonnet-5 es
// prefijo de claude-sonnet-5-5, que es otro modelo (ADR 0031). Una sesión que no
// llegó a declarar ninguno no tiene modelo que comparar y ya no terminó.
func (r *ResultadoDeEval) exigirElModeloPedido() {
	if r.ModeloDeLaSesion == "" || esElModeloPedido(r.ModeloDeLaSesion, r.Modelo) {
		return
	}

	r.Motivos = append(r.Motivos, motivoDeOtroModelo+r.ModeloDeLaSesion+", y se pidió "+r.Modelo)
	r.Pasa = false
}

// digitosDeLaFechaDelModelo son los de la fecha de la versión que el proveedor
// añade al id de un modelo, AAAAMMDD: claude-haiku-4-5-20251001.
const digitosDeLaFechaDelModelo = 8

// esElModeloPedido dice si el id que declara una sesión es el pedido, tal cual o
// seguido de un guion y la fecha de su versión.
func esElModeloPedido(declarado, pedido string) bool {
	fecha, conPrefijo := strings.CutPrefix(declarado, pedido)
	if !conPrefijo || fecha == "" {
		return conPrefijo
	}

	fecha, conGuion := strings.CutPrefix(fecha, "-")

	return conGuion && len(fecha) == digitosDeLaFechaDelModelo &&
		strings.IndexFunc(fecha, func(r rune) bool { return r < '0' || r > '9' }) == -1
}

// motivoDeActivacion es el motivo de una activación que no coincide con la que la
// eval espera.
func motivoDeActivacion(skill string, activa bool) string {
	if activa {
		return fmt.Sprintf("la activación no coincide: se esperaba que la skill %s se activara y no se activó", skill)
	}

	return fmt.Sprintf("la activación no coincide: se esperaba que la skill %s no se activara y se activó", skill)
}

// anotarLasQueNoSeActivan anota, con su motivo y en el orden de la eval, cada
// skill que la eval dice que no se activa y que la sesión activó, y devuelve
// cuántas son (contracts/evals-y-juicio.md §2 de H7.4; FR-003).
func (r *ResultadoDeEval) anotarLasQueNoSeActivan(noSeActivan []string, sesion Sesion) int {
	activadas := 0

	for _, nombre := range noSeActivan {
		if sesion.Activada(nombre) {
			activadas++

			r.Motivos = append(r.Motivos, motivoDeLaQueNoSeActiva(nombre))
		}
	}

	return activadas
}

// motivoDeLaQueNoSeActiva es el motivo de una skill que la eval dice que no se
// activa y que la sesión activó, que la nombra.
func motivoDeLaQueNoSeActiva(nombre string) string {
	return fmt.Sprintf("se activó la skill %s, que la eval dice que no se activa", nombre)
}

// repartirComandos reparte los comandos esperados entre ejecutados y ausentes,
// con un motivo por cada ausente.
func (r *ResultadoDeEval) repartirComandos(comandos []ComandoEsperado, invocaciones []invocacionJuzgada) {
	for _, comando := range comandos {
		texto := textoDelComando(comando)

		if slices.ContainsFunc(invocaciones, func(juzgada invocacionJuzgada) bool {
			return satisface(juzgada.Invocacion, comando)
		}) {
			r.ComandosEjecutados = append(r.ComandosEjecutados, texto)

			continue
		}

		r.ComandosAusentes = append(r.ComandosAusentes, texto)
		r.Motivos = append(r.Motivos, motivoDeComandoAusente+texto)
	}
}

// anotarProhibidos anota, con su motivo, cada comando prohibido que ejecuta
// alguna invocación de la sesión.
func (r *ResultadoDeEval) anotarProhibidos(prohibidos []ComandoProhibido, invocaciones []invocacionJuzgada) {
	for _, prohibido := range prohibidos {
		if !slices.ContainsFunc(invocaciones, func(juzgada invocacionJuzgada) bool {
			return ejecutaElProhibido(juzgada.Invocacion, prohibido)
		}) {
			continue
		}

		texto := prohibido.Applet + " " + prohibido.Verbo
		r.ComandosProhibidosEjecutados = append(r.ComandosProhibidosEjecutados, texto)
		r.Motivos = append(r.Motivos, motivoDeComandoProhibido+texto)
	}
}

// anotarLasOrdenesSinKitlegal anota, con su motivo y en su orden, cada
// invocación de la traza de una sesión sin kitlegal en el PATH —la de cualquier
// modo que no sea el modo orden: el modo herramienta y la de una eval sin
// binario ni servidor— que no es del applet del servidor, consulte o no y
// termine como termine, y devuelve cuántas son (research.md D18 de H21).
//
// Es lo que hace cierta la definición de esas sesiones: servidor.json lleva la
// ruta absoluta del binario, y una sesión que la lea puede ejecutarlo sin el
// PATH; sin esta regla, sus órdenes satisfarían los comandos y el modo
// herramienta mediría órdenes.
func (r *ResultadoDeEval) anotarLasOrdenesSinKitlegal(invocaciones []Invocacion) int {
	if r.Modo == ModoOrden {
		return 0
	}

	ordenes := 0

	for _, invocacion := range invocaciones {
		if invocacion.Applet == appletDelServidor {
			continue
		}

		ordenes++

		r.Motivos = append(r.Motivos, motivoDeOrdenSinKitlegal+ordenDeLaInvocacion(invocacion))
	}

	return ordenes
}

// juzgarLaRespuestaSinConsulta juzga, en una eval sin binario ni servidor, lo
// que su respuesta tiene que llevar y lo que no (contracts/evals-en-dos-modos.md
// §4 de H21; FR-047): una línea que empieza por ⚠ SIN CONSULTA AL BOE: con su
// dirección en esa misma línea (ExtraerSinConsulta) y ninguna cita, sea de la
// norma que sea. Anota si la línea está con su dirección y las citas de la
// respuesta, con el motivo de la línea ausente o sin su dirección y uno por
// cita, y dice si la respuesta cumple. En las demás evals no juzga nada y
// cumple.
func (r *ResultadoDeEval) juzgarLaRespuestaSinConsulta(eval Eval, respuesta string, citas []Cita) bool {
	if !eval.SinBinarioNiServidor {
		return true
	}

	// Sin la línea no hay dirección: la que va en otra línea no cuenta.
	conLinea, conDireccion := ExtraerSinConsulta(respuesta)
	r.LineaSinConsulta = conDireccion

	switch {
	case !conLinea:
		r.Motivos = append(r.Motivos, motivoDeLineaSinConsultaAusente)
	case !conDireccion:
		r.Motivos = append(r.Motivos, motivoDeLineaSinDireccion)
	}

	for _, cita := range citas {
		texto := cita.Norma + " " + cita.Bloque
		r.CitasSinConsulta = append(r.CitasSinConsulta, texto)
		r.Motivos = append(r.Motivos, motivoDeCitaSinConsulta+texto)
	}

	return r.LineaSinConsulta && len(r.CitasSinConsulta) == 0
}

// ejecutaElProhibido dice si la invocación ejecuta el comando prohibido
// (contrato evals-y-skill §2 de H7): consulta y es del mismo applet y del mismo
// verbo, termine con el código que termine o sin código. La que no consulta —la
// ayuda, --describe o --dry-run— no ejecuta nada.
func ejecutaElProhibido(invocacion Invocacion, prohibido ComandoProhibido) bool {
	return invocacion.Consulta && invocacion.Applet == prohibido.Applet && invocacion.Verbo == prohibido.Verbo
}

// repartirCitas reparte las citas esperadas entre encontradas y ausentes según
// las citas de la respuesta, con un motivo por cada ausente.
func (r *ResultadoDeEval) repartirCitas(esperadas []CitaEsperada, citas []Cita) {
	for _, esperada := range esperadas {
		texto := esperada.Norma + " " + esperada.Bloque

		if slices.Contains(citas, Cita(esperada)) {
			r.CitasEncontradas = append(r.CitasEncontradas, texto)

			continue
		}

		r.CitasAusentes = append(r.CitasAusentes, texto)
		r.Motivos = append(r.Motivos, motivoDeCitaAusente+texto)
	}
}

// repartirFormas reparte los avisos o los hallazgos esperados, en su orden y con
// sus repeticiones, entre los encontrados y los ausentes según los que la
// respuesta lleva con su forma fija, con un motivo por cada ausente: el principio
// dado seguido de su código o su clase. Sin ninguno esperado, las dos listas
// quedan vacías.
func (r *ResultadoDeEval) repartirFormas(esperados, conSuForma []string, motivo string) (encontrados, ausentes []string) {
	for _, esperado := range esperados {
		if slices.Contains(conSuForma, esperado) {
			encontrados = append(encontrados, esperado)

			continue
		}

		ausentes = append(ausentes, esperado)
		r.Motivos = append(r.Motivos, motivo+esperado)
	}

	return encontrados, ausentes
}

// repartirRedacciones reparte las redacciones modificadas esperadas, en su orden,
// entre las encontradas y las ausentes según las que traslada la respuesta: una
// está si alguna trasladada es igual, con la misma norma, el mismo bloque y las
// mismas dos fechas en el mismo orden. Cada ausente lleva su motivo.
func (r *ResultadoDeEval) repartirRedacciones(esperadas, trasladadas []RedaccionEsperada) {
	for _, esperada := range esperadas {
		texto := esperada.texto()

		if slices.Contains(trasladadas, esperada) {
			r.RedaccionesEncontradas = append(r.RedaccionesEncontradas, texto)

			continue
		}

		r.RedaccionesAusentes = append(r.RedaccionesAusentes, texto)
		r.Motivos = append(r.Motivos, motivoDeRedaccionAusente+texto)
	}
}

// repartirTerritorio reparte los elementos del territorio esperado entre
// encontrados y ausentes según lo que de él declara la respuesta, con un motivo
// por cada ausente.
func (r *ResultadoDeEval) repartirTerritorio(esperado, declarado TerritorioEsperado) {
	declarados := declarado.elementos()

	for _, elemento := range esperado.elementos() {
		if slices.Contains(declarados, elemento) {
			r.TerritorioEncontrado = append(r.TerritorioEncontrado, elemento)

			continue
		}

		r.TerritorioAusente = append(r.TerritorioAusente, elemento)
		r.Motivos = append(r.Motivos, motivoDeTerritorioAusente+elemento)
	}
}

// informar añade la invocación a las del resultado con sus parejas de destino y
// clase, una llegada a la red por cada destino de clase red y, si consultó y
// terminó con un código distinto de 0, la lleva a fuera de lo grabado con 4 o 5
// y a las otras fallidas con cualquier otro. La que no consultó (la ayuda,
// --describe o --dry-run) o quedó sin código no va a ninguna de las dos, y
// tampoco el proceso del servidor MCP, que no es una consulta
// (invocacionDeLaTraza).
func (r *ResultadoDeEval) informar(invocacion Invocacion) {
	r.informarLaJuzgada(invocacionDeLaTraza(invocacion))
}

// informarLaJuzgada informa de una invocación como la ve el juicio, con las
// reglas de informar: la de la traza o la que cuenta por una llamada a una
// herramienta, que lleva su marca.
func (r *ResultadoDeEval) informarLaJuzgada(invocacion invocacionJuzgada) {
	informada := InvocacionInformada{
		Orden:   invocacion.orden,
		Codigo:  copiaDelCodigo(invocacion.Codigo),
		Llamada: invocacion.llamada,
	}

	for _, conexion := range invocacion.Conexiones {
		pareja := ConexionInformada{Destino: conexion.Destino(), Clase: conexion.Clase}
		if slices.Contains(informada.Conexiones, pareja) {
			continue
		}

		informada.Conexiones = append(informada.Conexiones, pareja)

		if pareja.Clase == ConexionRed {
			r.LlegadasALaRed = append(r.LlegadasALaRed, LlegadaALaRed{Orden: informada.Orden, Destino: pareja.Destino})
		}
	}

	r.Invocaciones = append(r.Invocaciones, informada)

	if !invocacion.Consulta || invocacion.Codigo == nil || *invocacion.Codigo == 0 {
		return
	}

	fallida := InvocacionFallida{Orden: informada.Orden, Codigo: *invocacion.Codigo}

	if fallida.Codigo == codigoFuenteNoDisponible || fallida.Codigo == codigoLimiteOTos {
		r.FueraDeLoGrabado = append(r.FueraDeLoGrabado, fallida)

		return
	}

	r.OtrasFallidas = append(r.OtrasFallidas, fallida)
}

// satisface dice si la invocación satisface el comando esperado (data-model
// §6.1 de H5 y de H6): tiene que consultar, terminar con código 0 y ser del mismo
// applet; y, según la forma del comando, en la forma bloque, leer ese bloque de
// esa norma; en la consulta de norma, ser el mismo verbo con esa norma; en la
// búsqueda, ser buscar con cada término como palabra de sus argumentos; en el
// comando de territorio, ser resolver con el municipio como argumento; y en la
// comprobación, ser check (contrato evals-y-skill §2 de H7) y, si el comando
// lleva norma, con esa norma (contrato evals-y-skill §2 de H7.1); y en los dos
// de cita, ser de su verbo con el ROJ y el texto que el comando pida
// (pideLaCita; contracts/evals-jurisprudencia.md §2 de H23).
func satisface(invocacion Invocacion, comando ComandoEsperado) bool {
	if !consultoConExito(invocacion) || invocacion.Applet != comando.Applet {
		return false
	}

	var satisfecho bool

	switch formaDelComando(comando) {
	case formaBloque:
		satisfecho = leeElBloque(invocacion, comando.Norma, comando.Bloque)
	case formaConsultaDeNorma:
		satisfecho = invocacion.Verbo == comando.Verbo && esDeLaNorma(invocacion, comando.Norma)
	case formaBusqueda:
		satisfecho = invocacion.Verbo == verboBuscar && contieneLosTerminos(invocacion.Argumentos, comando.Terminos)
	case formaTerritorio:
		satisfecho = invocacion.Verbo == verboResolver && resuelveElMunicipio(invocacion.Argumentos, comando.Municipio)
	case formaComprobacion:
		satisfecho = compruebaLaMemoria(invocacion, comando.Norma)
	case formaPreparar, formaCotejar:
		satisfecho = pideLaCita(invocacion, comando)
	}

	return satisfecho
}

// compruebaLaMemoria dice si la invocación es check y, si el comando lleva
// norma, con esa norma.
func compruebaLaMemoria(invocacion Invocacion, norma string) bool {
	return invocacion.Verbo == verboCheck && (norma == "" || esDeLaNorma(invocacion, norma))
}

// consultoConExito dice si la invocación consultó y terminó con código 0: ni
// --describe ni --dry-run, que no leen nada, ni la que quedó sin código.
func consultoConExito(invocacion Invocacion) bool {
	return invocacion.Consulta && invocacion.Codigo != nil && *invocacion.Codigo == 0
}

// leeElBloque dice si la invocación es articulo o articulos de la norma con el
// bloque entre los que pide.
func leeElBloque(invocacion Invocacion, norma, bloque string) bool {
	return (invocacion.Verbo == verboArticulo || invocacion.Verbo == verboArticulos) &&
		esDeLaNorma(invocacion, norma) && slices.Contains(invocacion.Argumentos[1:], bloque)
}

// esDeLaNorma dice si la norma es el primer argumento de la invocación. Los
// verbos de una norma y los de sus bloques la reciben siempre en primer lugar y no
// tienen banderas propias (internal/app/boe.go), como graph check, que la recibe
// delante de sus bloques (internal/app/grafo.go), e InterpretarInvocacion ya quitó
// las globales.
func esDeLaNorma(invocacion Invocacion, norma string) bool {
	return len(invocacion.Argumentos) > 0 && invocacion.Argumentos[0] == norma
}

// resuelveElMunicipio dice si el único argumento de la invocación es el
// municipio, plegados los dos con territorio.Plegar, el mismo pliegue con el que
// el applet compara los nombres: resolver recibe un solo argumento y no tiene
// banderas propias (contrato del applet territorio §1), e InterpretarInvocacion ya
// quitó las globales. Resolverlo por su código INE no es resolver el municipio que
// la eval escribe.
func resuelveElMunicipio(argumentos []string, municipio string) bool {
	return len(argumentos) == 1 && territorio.Plegar(argumentos[0]) == territorio.Plegar(municipio)
}

// contieneLosTerminos dice si los argumentos, unidos por un espacio, contienen
// cada término como palabra, comparados los dos en minúsculas: así cuenta un
// término dentro de un argumento con espacios, como en buscar "bases del régimen
// local", y el que se escribe con otras mayúsculas.
func contieneLosTerminos(argumentos, terminos []string) bool {
	texto := strings.ToLower(strings.Join(argumentos, " "))

	for _, termino := range terminos {
		if !contieneComoPalabra(texto, strings.ToLower(termino)) {
			return false
		}
	}

	return true
}

// contieneComoPalabra dice si el texto contiene la palabra delimitada a los dos
// lados por el principio o el final del texto o por un carácter que no es letra
// ni cifra: «común» no está en «comúnmente» ni en «intercomún», y «7/1985» sí
// está en «ley 7/1985».
func contieneComoPalabra(texto, palabra string) bool {
	for desde := 0; desde <= len(texto)-len(palabra); {
		posicion := strings.Index(texto[desde:], palabra)
		if posicion < 0 {
			return false
		}

		inicio := desde + posicion
		anterior, _ := utf8.DecodeLastRuneInString(texto[:inicio])
		siguiente, _ := utf8.DecodeRuneInString(texto[inicio+len(palabra):])

		if !esDePalabra(anterior) && !esDePalabra(siguiente) {
			return true
		}

		desde = inicio + 1
	}

	return false
}

// esDePalabra dice si el carácter es letra o cifra. El que da el principio o el
// final del texto, utf8.RuneError, no lo es.
func esDePalabra(caracter rune) bool {
	return unicode.IsLetter(caracter) || unicode.IsNumber(caracter)
}

// ordenDeLaInvocacion es la orden de una invocación como la presenta el informe (data-model
// §10.2): el applet seguido de los argumentos que le siguen en argv, separados
// por un espacio. El applet es el nombre de invocación, como scripts/boe, o, si
// no lo es, el primer argumento de kitlegal (data-model §9).
func ordenDeLaInvocacion(invocacion Invocacion) string {
	argv := invocacion.Argv

	tokensDelApplet := 1
	if len(argv) > 0 && path.Base(argv[0]) != invocacion.Applet {
		tokensDelApplet = 2
	}

	return strings.Join(slices.Concat([]string{invocacion.Applet}, argv[min(tokensDelApplet, len(argv)):]), " ")
}

// textoDelComando es el comando esperado con el texto con el que lo presentan el
// informe y los motivos (contrato job-de-evals §5), según su forma: bloque
// <applet> <norma> <bloque> en la forma bloque, que satisfacen dos verbos;
// <applet> <verbo> <norma> en la consulta de norma; <applet> buscar <términos…> en
// la búsqueda; <applet> resolver <municipio> en el comando de territorio; y
// <applet> check en la comprobación (contrato evals-y-skill §1 de H7), seguido de
// su norma si la lleva (contrato evals-y-skill §2 de H7.1); y, en los dos de
// cita, el de textoDelComandoDeCita (contracts/evals-jurisprudencia.md §2 de
// H23).
func textoDelComando(comando ComandoEsperado) string {
	var partes []string

	switch formaDelComando(comando) {
	case formaBloque:
		partes = []string{"bloque", comando.Applet, comando.Norma, comando.Bloque}
	case formaConsultaDeNorma:
		partes = []string{comando.Applet, comando.Verbo, comando.Norma}
	case formaBusqueda:
		partes = slices.Concat([]string{comando.Applet, comando.Verbo}, comando.Terminos)
	case formaTerritorio:
		partes = []string{comando.Applet, comando.Verbo, comando.Municipio}
	case formaComprobacion:
		partes = []string{comando.Applet, comando.Verbo}
		if comando.Norma != "" {
			partes = append(partes, comando.Norma)
		}
	case formaPreparar, formaCotejar:
		return textoDelComandoDeCita(comando)
	}

	return strings.Join(partes, " ")
}

// copiaDelCodigo es una copia del código de una invocación, para que el resultado
// no comparta memoria con la sesión: nil si la invocación quedó sin código.
func copiaDelCodigo(codigo *int) *int {
	if codigo == nil {
		return nil
	}

	copia := *codigo

	return &copia
}
