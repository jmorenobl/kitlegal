package evals

import (
	"cmp"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
)

// Lo que EscribirInforme lee del directorio de una sesión además de lo que lee
// LeerSesion, y lo que deja en su destino (contrato job-de-evals §3.2, §3.3 y
// §5).
const (
	// directorioDeLaTraza es el traza/ de la sesión, con un fichero por hilo.
	directorioDeLaTraza = "traza"

	ficheroDelInformeMD   = "informe.md"
	ficheroDelInformeJSON = "informe.json"
)

// Veredicto es el veredicto global del informe de una ejecución del job de evals
// (data-model §10.3).
type Veredicto string

const (
	// VeredictoAprobado es el de una ejecución sin ningún motivo: sin ficheros de
	// eval mal formados, con todas las sesiones que el plan pide y legibles, con
	// cada serie que decide llegando al umbral, sin ninguna petición llegada a la
	// red (ADR 0016), con cada umbral que decide cumplido (FR-003 y FR-051 de
	// H7.3) y sin ninguna sesión sin medir (FR-043 de H7.3).
	VeredictoAprobado Veredicto = "aprobado"

	// VeredictoFallo es el de cualquier otra ejecución.
	VeredictoFallo Veredicto = "fallo"
)

// Motivos de texto fijo del informe (data-model §10.2 y §10.3; contrato
// job-de-evals §3.3). Los de una sesión ilegible van seguidos del fichero y su
// error; los de la raíz, detrás del fichero o de la sesión de los que hablan.
const (
	motivoDeSesionIlegible    = "sesión ilegible: "
	motivoSinEvalsQueJuzgar   = "ninguna eval bien formada que juzgar"
	motivoDeFicheroMalFormado = ": mal formado: "
	motivoDeLlegadaALaRed     = ": petición llegada a la red: "
)

// Motivos de la raíz que no son de la skill sino de la ejecución: empiezan por
// el prefijo fijo, que los distingue sin modelo de los demás (FR-043 de H7.3;
// contrato informe-del-job §2). El del límite de uso lleva cuántas sesiones
// quedaron sin medir y, separadas por «, », cada una con su motivo entre
// paréntesis.
const (
	motivoDeLaEjecucion  = "de la ejecución, no de la skill: "
	motivoDelLimiteDeUso = "límite de uso de la cuenta: %d sesiones sin medir: %s"
)

// Textos fijos de informe.md (contrato job-de-evals §5).
const (
	ningunoEnElInforme  = "ninguno"
	ningunaEnElInforme  = "ninguna"
	sinLlegadasALaRed   = "ninguna petición llegó a la red de una fuente"
	sinCodigoPorElCorte = "sin código (sesión cortada)"
	sinConexiones       = "sin conexiones"
	sinLeer             = "sin leer"
	vaciaEnElInforme    = "vacía"

	// medidaEnElInforme es la celda «Sin medir» de una sesión que se midió, y
	// serieSinMedir, el resultado de una serie con sesiones sin medir, con
	// cuántas (contrato informe-del-job §4 de H7.3).
	medidaEnElInforme = "no"
	serieSinMedir     = "sin medir (%d)"

	// sinModoEnElInforme es la celda «Modo» de una serie o una sesión que no es
	// de ningún modo: las de una eval sin binario ni servidor
	// (contracts/evals-en-dos-modos.md §5.3 de H21).
	sinModoEnElInforme = "—"
)

// modeloConSuModo es el modelo de lo que el informe publica de un modo cuando
// tiene que distinguirse por su modelo del mismo en otro modo: el id seguido del
// modo entre paréntesis (contracts/evals-en-dos-modos.md §5 de H21; research.md
// D19 y V33 de H21).
const modeloConSuModo = "%s (%s)"

// Encabezados de las tablas de informe.md (contrato job-de-evals §5; contrato
// informe-del-job §4 de H7.3 y de H7.4; contracts/evals-en-dos-modos.md §5.3 de
// H21; contracts/informe-del-job.md §7 de H24, que quita la columna y la tabla
// de las expresiones prohibidas).
var (
	encabezadosDeFueraDeLoGrabado = []string{"Sesión", "Eval", "Orden", "Código"}
	encabezadosDeRed              = []string{"Sesión", "Eval", "Orden", "Destino"}
	encabezadosDeSesiones         = []string{
		"Sesión", "Eval", "Modelo", "Modo", "Activa", "Activada", "Sesión terminada", "Comandos ausentes",
		"Comandos prohibidos ejecutados", "Citas ausentes", "Avisos encontrados", "Avisos ausentes",
		"Hallazgos encontrados", "Hallazgos ausentes", "Redacciones modificadas encontradas",
		"Redacciones modificadas ausentes", "Territorio encontrado", "Territorio ausente",
		"Reintentos por límite de ritmo", "Sin medir", "Resultado",
	}
	encabezadosDeSinMedir     = []string{"Sesión", "Eval", "Modelo", "Motivo"}
	encabezadosDeInvocaciones = []string{"Orden", "Código", "Conexiones", "Llamada"}
	encabezadosDeTasas        = []string{
		"Eval", "Modelo", "Modo", "Decide", "Planificada", "Formas exigidas", "Tasa", "Resultado",
	}
	encabezadosDeUmbrales = []string{"Umbral", "Medida", "Condición", "Cumple", "Hace fallar el veredicto"}
)

// enUnaLinea deja un texto en su línea de informe.md: cada salto de línea —\r\n,
// \r o \n, los tres finales de línea de CommonMark— se sustituye por un espacio.
// \r\n va antes que \r para que cuente como un solo salto.
var enUnaLinea = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ")

// escapeDeCelda deja el texto de una celda en su celda y en su línea: además de
// los saltos de línea, la barra, que la cerraría, se escribe \|.
var escapeDeCelda = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ", "|", `\|`)

// InformeAEscribir es lo que EscribirInforme necesita para escribir el informe de
// una ejecución del job de evals (contrato job-de-evals §3.3).
type InformeAEscribir struct {
	// Skill es la skill evaluada: la activación que se busca y el campo skill del
	// informe.
	Skill string

	// Evals es el directorio de las evals con las que se juzga: el job lo deriva
	// de la skill; TestInforme pasa el sintético.
	Evals string

	// Sesiones es el directorio con un subdirectorio por sesión.
	Sesiones string

	// Destino es el directorio en el que escribe informe.md e informe.json.
	Destino string

	// ModeloQueDecide es el modelo del job cuyas sesiones deciden el veredicto, y
	// ModelosInformativos, los que se ejecutan y se publican sin decidirlo
	// (ADR 0016).
	ModeloQueDecide     string
	ModelosInformativos []string

	// Repeticiones son las sesiones que el plan abre de cada eval con cada
	// modelo, y Umbral, cuántas de ellas tienen que pasar para que la serie pase.
	Repeticiones int
	Umbral       int

	// Modos son los modos del plan, los de PlanDeEvals.Modos: el job da el modo
	// orden y el modo herramienta; vacío es solo el modo orden, que es el plan
	// del sondeo (contracts/evals-en-dos-modos.md §2.1 de H21).
	Modos []Modo

	// Commit es el commit evaluado.
	Commit string

	// SinPython es la ruta de sin-python.txt: su contenido entero, byte a byte,
	// es el sin_python del informe.
	SinPython string

	// SinAbrir son las sesiones del plan que el repartidor no abrió tras una con
	// el mensaje del límite de uso: cuentan en su serie como sin medir (FR-044 de
	// H7.3; data-model §4).
	SinAbrir []SesionPlanificada

	// DuracionDeLasSesiones son los segundos que midió el repartidor: la suma de
	// los de sus tandas, cada una desde que empezó a preparar su primera sesión
	// hasta que terminó la última (contrato informe-del-job §5 de H7.3;
	// contracts/evals-en-dos-modos.md §5 de H21; FR-050).
	DuracionDeLasSesiones int

	// DuracionDeLosModos son, por modo, los segundos de su tanda, que es lo que
	// mide el umbral de la duración de ese modo; la tanda de las evals sin
	// binario ni servidor no es de ninguno, y solo cuenta en
	// DuracionDeLasSesiones (contracts/evals-en-dos-modos.md §2.2 y §5.1 de H21;
	// FR-043, FR-047).
	DuracionDeLosModos map[Modo]int

	// ObjetivoDeDuracion son los segundos que el job admite para la tanda de
	// cada modo: con 0, la skill no tiene objetivo ni umbrales de la duración;
	// uno negativo impide escribir el informe (contrato informe-del-job §5 de
	// H7.3; FR-051).
	ObjetivoDeDuracion int
}

// Informe es el informe de una ejecución del job de evals (data-model §10.3;
// contrato job-de-evals §5). Sus claves JSON, en el orden de sus campos, son las
// de informe.json.
type Informe struct {
	// Skill, los modelos, las repeticiones, el umbral y Commit son los recibidos,
	// tal cual: los modelos son los fijados en el job, no los de las sesiones.
	Skill string `json:"skill"`

	// ModeloQueDecide es el modelo cuyas series deciden el veredicto, y
	// ModelosInformativos, los que se publican como límite inferior (ADR 0016).
	ModeloQueDecide     string   `json:"modelo_que_decide"`
	ModelosInformativos []string `json:"modelos_informativos"`

	// Repeticiones y Umbral son las sesiones que el plan abre de cada eval con
	// cada modelo y cuántas de ellas tienen que pasar.
	Repeticiones int `json:"repeticiones"`
	Umbral       int `json:"umbral"`

	// ModelosDeSesion y VersionesDeClaudeCode son el modelo y la versión de
	// Claude Code que declara cada sesión que LeerSesion leyó, sin repetir y en
	// el orden en que aparece cada uno por primera vez, recorriendo las sesiones
	// por nombre.
	ModelosDeSesion       []string `json:"modelos_de_sesion"`
	VersionesDeClaudeCode []string `json:"versiones_de_claude_code"`

	Commit string `json:"commit"`

	// SinPython es el contenido de sin-python.txt, igual byte a byte: la
	// constancia de cómo se comprobó que no había Python (FR-081).
	SinPython string `json:"sin_python"`

	// FicherosMalFormados son los que LeerConjunto no leyó como eval, en orden de
	// fichero.
	FicherosMalFormados []FicheroMalFormadoDelInforme `json:"ficheros_mal_formados"`

	Veredicto Veredicto `json:"veredicto"`

	// Motivos son las causas del veredicto fallo, en el orden de data-model
	// §10.3; vacío con el veredicto aprobado, que es justo cuando no hay ninguna.
	Motivos []string `json:"motivos"`

	// Tasas son las series —las sesiones de una eval con un modelo en un modo—
	// con cuántas de sus sesiones pasan: primero las que el plan pide, en su
	// orden —las del modo orden, las del modo herramienta y las de las evals sin
	// binario ni servidor—, y después las observadas que el plan no pide
	// (data-model §10.4; contracts/evals-en-dos-modos.md §5 de H21).
	Tasas []TasaDelInforme `json:"tasas"`

	// Umbrales son los del contrato del ADR 0029, en el orden de
	// umbralesDelInforme: nil, [] en informe.json, si la skill no tiene juez ni
	// objetivo de duración (contrato informe-del-job §1 de H7.3; FR-001, FR-006;
	// contracts/informe-del-job.md §2 de H24; research D13 de H24).
	Umbrales []Umbral `json:"umbrales"`

	// DuracionDeLasSesiones son los segundos recibidos, tal cual: la suma de las
	// tandas del repartidor (FR-050 de H7.3; contracts/evals-en-dos-modos.md §5
	// de H21).
	DuracionDeLasSesiones int `json:"duracion_de_las_sesiones"`

	// ReintentosPorLimiteDeRitmo es la suma de los de cada sesión (FR-033 de
	// H7.3).
	ReintentosPorLimiteDeRitmo int `json:"reintentos_por_limite_de_ritmo"`

	// SesionesSinMedir son las que un límite de uso de la cuenta no dejó
	// terminar y las que el repartidor no abrió tras él, en orden de sesión; [] en
	// informe.json si no hay ninguna (data-model §3 de H7.3; FR-043).
	SesionesSinMedir []SesionSinMedir `json:"sesiones_sin_medir"`

	// FueraDeLoGrabado son las invocaciones fuera de lo grabado de todas las
	// sesiones, y Red, sus llegadas a la red, cada una con su sesión y su eval.
	FueraDeLoGrabado []FueraDeLoGrabadoDelInforme `json:"fuera_de_lo_grabado"`
	Red              []RedDelInforme              `json:"red"`

	// Evals son los resultados, uno por sesión y en orden de sesión.
	Evals []ResultadoDeEval `json:"evals"`
}

// FicheroMalFormadoDelInforme es un fichero del directorio de evals que no es una
// eval, con su error.
type FicheroMalFormadoDelInforme struct {
	Fichero string `json:"fichero"`
	Error   string `json:"error"`
}

// TasaDelInforme es una serie —las sesiones de una eval con un modelo en un
// modo— con cuántas de ellas pasan (data-model §10.4). Con el umbral, es lo que
// sustituye a «10 de 10 en una sola tirada» (ADR 0016).
type TasaDelInforme struct {
	// Eval es el fichero de la eval con la que se juzgan sus sesiones.
	Eval string `json:"eval"`

	// Modelo es el id del modelo con el que se abrieron en el modo orden y en una
	// eval sin binario ni servidor, y «<id> (herramienta)» en el modo
	// herramienta: así ninguna pareja de eval y modelo se repite entre modos, y
	// scripts/workflow/informe.sh, que enseña la primera serie de cada una, da su
	// celda a cada serie (modeloDeLaSerie; contracts/evals-en-dos-modos.md §5 de
	// H21; research.md D19 y V33 de H21; FR-048).
	Modelo string `json:"modelo"`

	// Modo es el modo de sus sesiones; ninguno, en las de una eval sin binario ni
	// servidor.
	Modo Modo `json:"modo"`

	// PreguntaAmpliada dice que la pregunta de sus sesiones no es la de la eval
	// sino la de la eval con algo más: la sesión de la prueba de red (contrato
	// job-de-evals §6). Nunca la tiene una serie planificada.
	PreguntaAmpliada bool `json:"pregunta_ampliada"`

	// Planificada dice si la serie es una de las que pide el plan, y Decide, si
	// su resultado decide el veredicto: solo la del modelo que decide sobre una
	// eval que no es informativa.
	Planificada bool `json:"planificada"`
	Decide      bool `json:"decide"`

	// Formas son las formas fijas que exige la eval de la serie, junto a su tasa:
	// la de cada clase de sus hallazgos, en su orden, escrita como se enseña —la
	// marca, un espacio, la etiqueta de grafo.EtiquetasDeHallazgo y los dos
	// puntos—, y detrás, ⚠ REDACCIÓN MODIFICADA: <texto> por cada redacción
	// modificada que espera, en su orden. Nunca nil: vacía si la eval no espera
	// hallazgos ni redacciones o si la serie no es de ninguna eval bien formada
	// (contrato evals-y-skill §6 de H7.1; FR-055; contracts/informe-del-job.md §4
	// de H7.4).
	Formas []string `json:"formas"`

	// Sesiones son las que se leyeron de la serie, y Pasan, cuántas de ellas
	// pasan. SinMedir son las de la serie que quedaron sin medir, abiertas o no
	// (data-model §4 de H7.3). Pasa dice si Pasan llega al umbral sin ninguna
	// sin medir: una serie con alguna ni pasa ni falla (FR-042 de H7.3).
	Sesiones int  `json:"sesiones"`
	Pasan    int  `json:"pasan"`
	SinMedir int  `json:"sin_medir"`
	Pasa     bool `json:"pasa"`
}

// recuentoDeRespuestas es, para el modelo que decide en un modo del plan, el
// recuento de sus respuestas medidas (data-model §5 de H7.4; data-model §10 de
// H21), con el que se mide el umbral de las respuestas sin la skill activada
// (FR-041 de H7.4; FR-034 de H24).
type recuentoDeRespuestas struct {
	modelo string
	modo   Modo

	// respuestas son las medidas del modelo en ese modo: sus sesiones juzgadas
	// —no las ilegibles—, no sin medir y terminadas de las series de ese modo
	// que pide el plan con ese modelo cuya eval espera que la skill se active
	// (contracts/informe-del-job.md §1 de H7.4; FR-045). Y sinActivar, las de
	// ellas que no activaron la skill.
	respuestas, sinActivar int
}

// contar cuenta el resultado de una respuesta medida en su recuento: como
// respuesta y, si no activó la skill, sin la skill activada.
func (r *recuentoDeRespuestas) contar(resultado ResultadoDeEval) {
	r.respuestas++

	if !resultado.Activada {
		r.sinActivar++
	}
}

// modeloDeLaSerie es el modelo que publica la serie de un modelo en un modo: el
// id en el modo orden y en una eval sin binario ni servidor, como hasta H21, y
// el id con su modo en el modo herramienta (contracts/evals-en-dos-modos.md §5
// de H21; research.md D19 de H21). Con él nombran también el modo los motivos
// de la serie.
func modeloDeLaSerie(modelo string, modo Modo) string {
	if modo != ModoHerramienta {
		return modelo
	}

	return fmt.Sprintf(modeloConSuModo, modelo, modo)
}

// SesionSinMedir es una sesión que quedó sin medir por un límite de uso de la
// cuenta, o que el repartidor no abrió tras él, con su eval, su modelo y la
// descripción de su clase (data-model §3 de H7.3).
type SesionSinMedir struct {
	Sesion string `json:"sesion"`
	Eval   string `json:"eval"`
	Modelo string `json:"modelo"`
	Motivo string `json:"motivo"`
}

// FueraDeLoGrabadoDelInforme es una invocación fuera de lo grabado con la sesión
// y la eval de su resultado.
type FueraDeLoGrabadoDelInforme struct {
	Sesion string `json:"sesion"`
	Eval   string `json:"eval"`
	Orden  string `json:"orden"`
	Codigo int    `json:"codigo"`
}

// RedDelInforme es una llegada a la red con la sesión y la eval de su resultado.
type RedDelInforme struct {
	Sesion  string `json:"sesion"`
	Eval    string `json:"eval"`
	Orden   string `json:"orden"`
	Destino string `json:"destino"`
}

// EscribirInforme juzga una ejecución del job de evals y escribe su informe.md y
// su informe.json en e.Destino, con permisos 0o600 y un salto de línea final,
// antes de devolver el Informe (contrato job-de-evals §3.3 y §5; data-model
// §10.3; FR-071, FR-073, FR-076):
//
//  1. lee sin-python.txt y las evals con LeerConjunto: los ficheros mal formados
//     van al informe y las sesiones se juzgan con las bien formadas, sin las
//     reglas del conjunto, que aplica antes el guion;
//  2. por cada entrada de e.Sesiones, en orden de nombre, lee siempre eval.txt,
//     modelo.txt, pregunta.txt y la sesión con LeerSesion y, si LeerSesion la
//     leyó, su traza con LeerTrazas y el corte de la sesión. Solo si todo se leyó
//     y eval.txt nombra una eval bien formada, la juzga con Juzgar y la clasifica
//     con ClasificarElLimite: la que un límite de uso de la cuenta no dejó
//     terminar queda sin medir, con el del límite como único motivo (data-model
//     §3 de H7.3); si no, la sesión no pasa, con un motivo «sesión ilegible:
//     <fichero>: <error>» por cada fichero que falta o no se puede leer, o por
//     el eval.txt que no nombra
//     ninguna eval, en el orden eval.txt, modelo.txt, pregunta.txt,
//     servidor.json —si no se puede saber si lo tiene—, sesión y traza. Una
//     entrada que no es un directorio no se salta: sus ficheros no se pueden
//     leer. Lo que sí se leyó de una sesión sin juzgar —la eval que
//     nombra, lo observado de la sesión y lo que hicieron sus invocaciones— se
//     informa igual, para que ninguna llegada a la red quede sin detectar
//     (FR-076);
//  3. reparte las sesiones en series —eval, modelo, modo y si la pregunta es la
//     de la eval— y las compara con las que pide el plan (PlanDeEvals) en los
//     modos de e.Modos: cada serie lleva su tasa, y una serie que decide pasa si
//     llegan al umbral (data-model §10.4; ADR 0016) y ninguna quedó sin medir,
//     contando como sin medir las de e.SinAbrir (FR-042 y FR-044 de H7.3); con
//     una skill que tiene juez, cuenta además, del modelo que decide y por modo,
//     las respuestas medidas —terminadas— de las series de ese modo que pide el
//     plan cuya eval activa la skill, y las que no la activaron
//     (recontarRespuestas; contracts/informe-del-job.md §1 de H7.4; research D13
//     de H24), sin que eso cambie el juicio de ninguna sesión: una sin terminar
//     sigue sin pasar en su serie;
//  4. con el recuento, compone los umbrales (umbralesDelInforme): si la skill
//     tiene juez, por cada modo, el de las respuestas sin la skill activada del
//     modelo que decide, y los de la duración de cada modo si hay objetivo
//     (contrato informe-del-job §1 de H7.3; contracts/informe-del-job.md §2 de
//     H7.4 y de H24; contracts/evals-en-dos-modos.md §5.1 de H21);
//  5. los motivos de la raíz van en el orden de data-model §10.3: por serie
//     planificada, las sesiones que faltan y, si decide y no llega al umbral, su
//     tasa seguida de los motivos de sus sesiones que no pasan; los de cada
//     sesión ilegible que no se hayan escrito ya; ninguna eval bien formada que
//     juzgar; cada fichero mal formado; cada petición llegada a la red; el de
//     cada umbral que decide y no se cumple, salvo los de la duración (FR-003 de
//     H7.3); con alguna sesión sin medir, uno solo de la ejecución que las nombra
//     (FR-043 de H7.3); y, si deciden y no se cumplen, los de la duración,
//     también de la ejecución (FR-051 de H7.3). El veredicto es fallo si hay
//     algún motivo y aprobado si no hay ninguno, de modo que los motivos son
//     exactamente las causas del fallo, y un umbral que decide y no se cumple lo
//     pone en fallo.
//
// Desde H21 (contracts/evals-en-dos-modos.md §3 a §5; FR-043, FR-044, FR-047,
// FR-048), cada sesión se juzga con el modo que da su directorio
// (modoDeLaSesion): la del modo herramienta, con sus llamadas, y la de una eval
// sin binario ni servidor, sin ninguno. Una serie es de un modo, y la que decide
// y no llega al umbral en uno pone el veredicto en fallo aunque la de la misma
// eval pase en el otro; las medidas de un modo no se suman a las del otro; y las
// sesiones de una eval sin binario ni servidor forman su serie, que decide, y no
// entran en ningún recuento ni en ningún umbral. El modo lo nombran el modelo de
// la serie (modeloDeLaSerie) y el nombre del umbral, y con ellos los motivos.
//
// Desde H24 (contracts/informe-del-job.md §2, §6 y §7 de H24; research D13 y
// D14; FR-034, FR-062, FR-070), el informe no lleva nada de la lista de
// expresiones prohibidas de la skill: ni el recuento por modelo, ni las de cada
// sesión, ni sus umbrales, ni su sección ni su columna de informe.md. Los
// umbrales de las respuestas dependen de que la skill tenga juez.
//
// El error es solo para lo que impide escribir el informe —un plan sin sentido,
// un objetivo de duración negativo, o sin-python.txt, las evals o las sesiones
// que no se pueden leer, o un destino en el que no se puede escribir— y nombra el
// fichero, el directorio o el valor. Todo se lee
// antes de escribir nada y, con cualquiera de esos errores, en el destino no queda
// ni informe.md ni informe.json, de modo que un sin_python vacío o a medias no
// llega nunca a un informe (FR-081).
func EscribirInforme(e InformeAEscribir) (Informe, error) {
	if err := e.plan(nil).Comprobar(); err != nil {
		return Informe{}, fmt.Errorf("el informe no se puede escribir: %w", err)
	}

	if e.Umbral < 1 || e.Umbral > e.Repeticiones {
		return Informe{}, fmt.Errorf(
			"el informe no se puede escribir: el umbral es %d y tiene que estar entre 1 y las %d repeticiones",
			e.Umbral, e.Repeticiones)
	}

	if e.ObjetivoDeDuracion < 0 {
		return Informe{}, fmt.Errorf(
			"el informe no se puede escribir: el objetivo de duración es %d s y no puede ser negativo (0 es sin objetivo)",
			e.ObjetivoDeDuracion)
	}

	sinPython, err := leerFichero(e.SinPython)
	if err != nil {
		return Informe{}, fmt.Errorf("el informe no se puede escribir: la comprobación sin Python %s no se puede leer: %w",
			e.SinPython, err)
	}

	conjunto, err := LeerConjunto(e.Evals)
	if err != nil {
		return Informe{}, fmt.Errorf("el informe no se puede escribir: %w", err)
	}

	sesiones, err := juzgarSesiones(e, conjunto.Evals)
	if err != nil {
		return Informe{}, fmt.Errorf("el informe no se puede escribir: %w", err)
	}

	informe := componerInforme(e, string(sinPython), conjunto, sesiones)

	// Un argv de la traza puede traer cualquier octeto: el que no es UTF-8 válido
	// se escribe como el carácter de sustitución, en lugar de dejar la ejecución
	// sin informe.
	codificado, err := json.Marshal(informe, jsontext.WithIndent("  "), jsontext.AllowInvalidUTF8(true))
	if err != nil {
		return Informe{}, fmt.Errorf("el informe no se puede escribir: %s no se puede codificar: %w", ficheroDelInformeJSON, err)
	}

	if err := escribirElInforme(e.Destino, renderizarInforme(informe, sesiones), append(codificado, '\n')); err != nil {
		return Informe{}, err
	}

	return informe, nil
}

// ejecucionPorTandas es lo que el informe recibe del repartidor cuando ejecuta el
// plan tanda a tanda (contracts/evals-en-dos-modos.md §2.2 y §5 de H21).
type ejecucionPorTandas struct {
	// sinAbrir son las sesiones del plan que no se abrieron tras una sesión con
	// el mensaje del límite de uso, en su orden: las de su tanda y todas las de
	// las tandas siguientes. Es el SinAbrir del informe.
	sinAbrir []SesionPlanificada

	// duracion es la suma de los segundos de las tandas ejecutadas, cada una
	// redondeada hacia arriba: el DuracionDeLasSesiones del informe.
	duracion int

	// porModo son los segundos de la tanda de cada modo que se ejecutó: el
	// DuracionDeLosModos del informe. La tanda de las evals sin binario ni
	// servidor, que no es de ninguno, no está.
	porModo map[Modo]int
}

// ejecutarPorTandas ejecuta el plan con una llamada a ejecutar por tanda, en su
// orden, y reúne lo que el informe necesita de ellas
// (contracts/evals-en-dos-modos.md §2.2 de H21; research.md D16 de H21): la
// duración de un modo es la de su tanda, y la de las sesiones, la suma de las de
// todas. Si una tanda acaba con sesiones sin abrir por el mensaje del límite de
// uso, las siguientes no se ejecutan y todas sus sesiones cuentan como sin
// abrir. El error de una tanda es el de la ejecución, y las siguientes tampoco
// se ejecutan.
//
// ejecutar es el repartidor, ejecutarSesiones, con la tanda como plan: lo recibe
// para que TestEjecucionDelJob, que es quien abre sesiones con modelo, no sea el
// único sitio en el que está el reparto por tandas.
func ejecutarPorTandas(
	plan []SesionPlanificada, ejecutar func(tanda []SesionPlanificada) (EjecucionDeSesiones, error),
) (ejecucionPorTandas, error) {
	ejecucion := ejecucionPorTandas{porModo: map[Modo]int{}}

	for _, tanda := range tandasDelPlan(plan) {
		if len(ejecucion.sinAbrir) > 0 {
			ejecucion.sinAbrir = append(ejecucion.sinAbrir, tanda...)

			continue
		}

		deLaTanda, err := ejecutar(tanda)
		if err != nil {
			return ejecucionPorTandas{}, err
		}

		segundos := segundosHaciaArriba(deLaTanda.Duracion)
		ejecucion.duracion += segundos

		if modo := tanda[0].Modo; modo != "" {
			ejecucion.porModo[modo] = segundos
		}

		// Una copia: lo que se le añada después no puede escribir en el plan.
		ejecucion.sinAbrir = slices.Clone(deLaTanda.SinAbrir)
	}

	return ejecucion, nil
}

// tandasDelPlan parte las sesiones del plan, sin cambiar su orden, en sus
// tandas: cada racha de sesiones seguidas del mismo modo. Con el plan de
// PlanDeEvals.Sesiones son, como mucho, tres: la del modo orden, con la prueba
// de red, la del modo herramienta y la de las evals sin binario ni servidor
// (contracts/evals-en-dos-modos.md §2.1 de H21).
func tandasDelPlan(plan []SesionPlanificada) [][]SesionPlanificada {
	var tandas [][]SesionPlanificada

	inicio := 0

	for posicion := 1; posicion <= len(plan); posicion++ {
		if posicion < len(plan) && plan[posicion].Modo == plan[inicio].Modo {
			continue
		}

		tandas = append(tandas, plan[inicio:posicion])
		inicio = posicion
	}

	return tandas
}

// segundosHaciaArriba son los segundos enteros de una duración, redondeados
// hacia arriba: 900,4 s son 901 y no cumplen un objetivo de 900 (research.md D14
// de H7.3).
func segundosHaciaArriba(duracion time.Duration) int {
	return int((duracion + time.Second - 1) / time.Second)
}

// sesionJuzgada es lo que el informe lleva de una sesión: su resultado y, para su
// sección de informe.md, lo leído de ella.
type sesionJuzgada struct {
	resultado ResultadoDeEval

	// ilegible dice si la sesión quedó sin juzgar porque algo no se pudo leer.
	ilegible bool

	// sesion es lo que leyó LeerSesion, con las invocaciones de su traza si se
	// leyó; leida dice si LeerSesion la leyó, y trazaLeida, si LeerTrazas leyó su
	// traza.
	sesion     Sesion
	leida      bool
	trazaLeida bool

	// pregunta es el contenido de pregunta.txt, y preguntaLeida dice si se leyó.
	pregunta      string
	preguntaLeida bool

	// clave es la serie a la que pertenece la sesión (data-model §10.4). Una
	// sesión de la que no se pudo leer la eval, el modelo o la pregunta queda en
	// una serie que el plan no pide, de modo que a la serie que sí pide le falta
	// una sesión y el informe lo dice.
	clave claveDeSerie
}

// claveDeSerie identifica la serie de una sesión: la eval con la que se juzga, el
// modelo con el que se abrió, el modo que da su directorio y si su pregunta es
// la de la eval o la de la eval con algo más (data-model §10.4; data-model §6 de
// H21).
type claveDeSerie struct {
	eval             string
	modelo           string
	modo             Modo
	preguntaAmpliada bool
}

// serieJuzgada es una serie con su clave, su tasa, las posiciones de sus sesiones
// en el orden en que se leyeron y cuántas de las suyas no abrió el repartidor
// tras el mensaje del límite de uso.
type serieJuzgada struct {
	clave    claveDeSerie
	tasa     TasaDelInforme
	sesiones []int
	sinAbrir int
}

// plan es el plan de sesiones que el informe exige que se haya ejecutado, con las
// evals bien formadas y en los modos recibidos. Sin prueba de red: esa sesión no
// forma serie planificada, porque no pregunta lo que pregunta la eval (contrato
// job-de-evals §6).
func (e InformeAEscribir) plan(evals []Eval) PlanDeEvals {
	return PlanDeEvals{
		Evals:               evals,
		ModeloQueDecide:     e.ModeloQueDecide,
		ModelosInformativos: e.ModelosInformativos,
		Repeticiones:        e.Repeticiones,
		Modos:               e.Modos,
	}
}

// juzgarSesiones juzga cada entrada del directorio de sesiones, en orden de
// nombre. El error queda para el directorio que no se puede listar.
func juzgarSesiones(e InformeAEscribir, evals []Eval) ([]sesionJuzgada, error) {
	entradas, err := os.ReadDir(e.Sesiones)
	if err != nil {
		return nil, fmt.Errorf("el directorio de sesiones %s no se puede listar: %w", e.Sesiones, err)
	}

	sesiones := make([]sesionJuzgada, 0, len(entradas))
	for _, entrada := range entradas {
		sesiones = append(sesiones, juzgarSesion(e, evals, entrada.Name()))
	}

	return sesiones, nil
}

// juzgarSesion lee la sesión del subdirectorio nombre y la juzga con la eval que
// nombra su eval.txt y el modo que da su directorio, o la deja sin pasar con un
// motivo por cada fichero que no se pudo leer (paso 2 de EscribirInforme). Si no
// se puede saber si tiene servidor.json, su motivo va detrás del de pregunta.txt
// y la sesión queda sin modo.
func juzgarSesion(e InformeAEscribir, evals []Eval, nombre string) sesionJuzgada {
	dir := filepath.Join(e.Sesiones, nombre)

	var (
		juzgada sesionJuzgada
		motivos []string
	)

	nombreDeEval, eval, errDeEval := leerEvalDeLaSesion(dir, e.Evals, evals)
	if errDeEval != nil {
		motivos = append(motivos, motivoDeSesionIlegible+errDeEval.Error())
	} else {
		juzgada.clave.eval = nombreDeEval
	}

	modelo, err := leerFicheroDeSesion(dir, ficheroDelModelo)
	if err != nil {
		motivos = append(motivos, motivoDeSesionIlegible+err.Error())
	} else {
		juzgada.clave.modelo = strings.TrimSuffix(string(modelo), "\n")
	}

	pregunta, err := leerFicheroDeSesion(dir, ficheroDeLaPregunta)
	if err != nil {
		motivos = append(motivos, motivoDeSesionIlegible+err.Error())
	} else {
		juzgada.pregunta, juzgada.preguntaLeida = string(pregunta), true
		juzgada.clave.preguntaAmpliada = errDeEval == nil && juzgada.pregunta != eval.Pregunta+"\n"
	}

	// El modo lo da el directorio, y no el transcript ni el nombre de la sesión:
	// con una eval que no se pudo leer, el de una sesión sin servidor.json es el
	// modo orden.
	modo, err := modoDeLaSesion(dir, eval)
	if err != nil {
		motivos = append(motivos, motivoDeSesionIlegible+err.Error())
	} else {
		juzgada.clave.modo = modo
	}

	juzgada.sesion, err = LeerSesion(dir)
	if err != nil {
		motivos = append(motivos, motivoDeSesionIlegible+err.Error())
	} else {
		juzgada.leida = true

		// Sin el código de la sesión no se sabría si el tope la cortó: la traza
		// solo se lee de una sesión que LeerSesion leyó.
		juzgada.sesion.Invocaciones, err = LeerTrazas(filepath.Join(dir, directorioDeLaTraza), juzgada.sesion.Cortada)
		if err != nil {
			motivos = append(motivos, motivoDeSesionIlegible+directorioDeLaTraza+": "+err.Error())
		} else {
			juzgada.trazaLeida = true
		}
	}

	if len(motivos) == 0 {
		juzgada.resultado = juzgarEnModo(eval, juzgada.sesion, e.Skill, modo)
		juzgada.resultado.Sesion = nombre
		juzgada.resultado.Modelo = juzgada.clave.modelo
		juzgada.resultado.ModeloDeLaSesion = juzgada.sesion.Modelo
		juzgada.resultado.exigirElModeloPedido()
		juzgada.resultado.dejarSinMedir(ClasificarElLimite(juzgada.sesion))

		return juzgada
	}

	juzgada.ilegible = true
	juzgada.resultado = ResultadoDeEval{
		Sesion: nombre, Eval: nombreDeEval, Modelo: juzgada.clave.modelo, Modo: juzgada.clave.modo,
		ModeloDeLaSesion: juzgada.sesion.Modelo, Motivos: motivos,
	}

	if errDeEval == nil {
		juzgada.resultado.Activa = eval.Activa
	}

	if juzgada.leida {
		codigo := juzgada.sesion.Codigo
		juzgada.resultado.Activada = juzgada.sesion.Activada(e.Skill)
		juzgada.resultado.Respuesta = juzgada.sesion.Respuesta
		juzgada.resultado.CodigoDeLaSesion = &codigo
		juzgada.resultado.FinDeLaSesion = juzgada.sesion.Fin
		juzgada.resultado.SesionTerminada = juzgada.sesion.Terminada
		juzgada.resultado.ReintentosPorLimiteDeRitmo = juzgada.sesion.ReintentosPorLimiteDeRitmo()
	}

	for _, invocacion := range juzgada.sesion.Invocaciones {
		juzgada.resultado.informar(invocacion)
	}

	return juzgada
}

// leerEvalDeLaSesion lee de eval.txt el nombre de la eval con la que se juzga la
// sesión, sin su salto de línea final, y la busca entre las evals bien formadas
// del directorio de evals. El nombre es vacío si eval.txt no se leyó, y el error,
// que empieza por eval.txt como el de cualquier fichero de la sesión, nombra
// también el que no es ninguna eval bien formada.
func leerEvalDeLaSesion(dir, directorioDeEvals string, evals []Eval) (string, Eval, error) {
	contenido, err := leerFicheroDeSesion(dir, ficheroDeLaEval)
	if err != nil {
		return "", Eval{}, err
	}

	nombre := strings.TrimSuffix(string(contenido), "\n")

	posicion := slices.IndexFunc(evals, func(eval Eval) bool { return eval.Fichero == nombre })
	if posicion < 0 {
		return nombre, Eval{}, fmt.Errorf("%s: %s nombra %q, que no es ninguna eval bien formada de %s",
			ficheroDeLaEval, filepath.Join(dir, ficheroDeLaEval), nombre, directorioDeEvals)
	}

	return nombre, evals[posicion], nil
}

// componerInforme reúne en el informe lo recibido, lo leído y los resultados de
// las sesiones, reparte las sesiones en series y decide sus motivos y su
// veredicto.
func componerInforme(e InformeAEscribir, sinPython string, conjunto Conjunto, sesiones []sesionJuzgada) Informe {
	informe := Informe{
		Skill:                 e.Skill,
		ModeloQueDecide:       e.ModeloQueDecide,
		ModelosInformativos:   e.ModelosInformativos,
		Repeticiones:          e.Repeticiones,
		Umbral:                e.Umbral,
		Commit:                e.Commit,
		SinPython:             sinPython,
		DuracionDeLasSesiones: e.DuracionDeLasSesiones,
	}

	for _, malFormado := range conjunto.MalFormados {
		informe.FicherosMalFormados = append(informe.FicherosMalFormados,
			FicheroMalFormadoDelInforme{Fichero: malFormado.Fichero, Error: malFormado.Error.Error()})
	}

	series := repartirEnSeries(e, conjunto.Evals, sesiones)
	for _, serie := range series {
		informe.Tasas = append(informe.Tasas, serie.tasa)

		for _, posicion := range serie.sesiones {
			sesiones[posicion].resultado.Decide = serie.tasa.Decide
		}
	}

	informe.Umbrales = umbralesDelInforme(e, recontarRespuestas(e, conjunto.Juez, sesiones, series))

	for _, juzgada := range sesiones {
		resultado := juzgada.resultado
		informe.Evals = append(informe.Evals, resultado)
		informe.ReintentosPorLimiteDeRitmo += resultado.ReintentosPorLimiteDeRitmo

		if juzgada.leida {
			informe.ModelosDeSesion = agregarSinRepetir(informe.ModelosDeSesion, juzgada.sesion.Modelo)
			informe.VersionesDeClaudeCode = agregarSinRepetir(informe.VersionesDeClaudeCode,
				juzgada.sesion.VersionDeClaudeCode)
		}

		for _, fuera := range resultado.FueraDeLoGrabado {
			informe.FueraDeLoGrabado = append(informe.FueraDeLoGrabado, FueraDeLoGrabadoDelInforme{
				Sesion: resultado.Sesion, Eval: resultado.Eval, Orden: fuera.Orden, Codigo: fuera.Codigo,
			})
		}

		for _, llegada := range resultado.LlegadasALaRed {
			informe.Red = append(informe.Red, RedDelInforme{
				Sesion: resultado.Sesion, Eval: resultado.Eval, Orden: llegada.Orden, Destino: llegada.Destino,
			})
		}
	}

	informe.SesionesSinMedir = sesionesSinMedir(informe.Evals, e.SinAbrir)
	informe.Motivos = motivosDelInforme(e, informe, conjunto.Evals, sesiones, series)
	informe.Veredicto = veredictoDelInforme(informe)

	return informe
}

// repartirEnSeries reparte las sesiones en series (data-model §10.4): primero las
// que pide el plan, en su orden y aunque no tengan ninguna sesión, y después las
// observadas que el plan no pide —la de la prueba de red y la de cualquier sesión
// de la que no se pudieran leer la eval, el modelo o la pregunta—, en el orden en
// que aparecen. Cada sesión que el repartidor no abrió cuenta en su serie como
// sin medir, igual que las que quedaron sin medir (data-model §4 de H7.3). Una
// serie pasa si no tiene ninguna sin medir y sus sesiones que pasan llegan al
// umbral (FR-042 de H7.3), y declara las formas que exige su eval
// (formasExigidas).
//
// Desde H21, la serie es además de un modo: las del plan van en el orden de
// PlanDeEvals.Series —las del modo orden, las del modo herramienta y las de las
// evals sin binario ni servidor, sin modo—, y cada sesión cae en la de su eval,
// su modelo y el modo que da su directorio, de modo que la que se abrió en un
// modo que el plan no pide queda en una serie que el plan no pide. La tasa
// publica el modo y, como modelo, el de modeloDeLaSerie
// (contracts/evals-en-dos-modos.md §5 de H21).
func repartirEnSeries(e InformeAEscribir, evals []Eval, sesiones []sesionJuzgada) []serieJuzgada {
	var series []serieJuzgada

	posiciones := map[claveDeSerie]int{}

	// anadir añade la serie de la clave, con lo que el plan dice de ella, y da
	// su posición.
	anadir := func(clave claveDeSerie, planificada, decide bool) int {
		posicion := len(series)
		posiciones[clave] = posicion
		series = append(series, serieJuzgada{clave: clave, tasa: TasaDelInforme{
			Eval: clave.eval, Modelo: modeloDeLaSerie(clave.modelo, clave.modo), Modo: clave.modo,
			PreguntaAmpliada: clave.preguntaAmpliada, Planificada: planificada, Decide: decide,
		}})

		return posicion
	}

	// enSeries es la posición de la serie de la clave, que se añade como una de
	// las observadas que el plan no pide si aún no está.
	enSeries := func(clave claveDeSerie) int {
		if posicion, esta := posiciones[clave]; esta {
			return posicion
		}

		return anadir(clave, false, false)
	}

	for _, planificada := range e.plan(evals).Series() {
		anadir(claveDeSerie{eval: planificada.Eval, modelo: planificada.Modelo, modo: planificada.Modo}, true,
			planificada.Decide)
	}

	// enSeries puede añadir a series: la posición se toma antes de indexar, porque
	// en series[enSeries(…)] el orden entre las dos cosas no está especificado.
	for posicion, juzgada := range sesiones {
		enLaSerie := enSeries(juzgada.clave)
		serie := &series[enLaSerie]
		serie.sesiones = append(serie.sesiones, posicion)
		serie.tasa.Sesiones++

		switch {
		case juzgada.resultado.SinMedir != "":
			serie.tasa.SinMedir++
		case juzgada.resultado.Pasa:
			serie.tasa.Pasan++
		}
	}

	for _, sinAbrir := range e.SinAbrir {
		enLaSerie := enSeries(claveDeSerie{
			eval: sinAbrir.Fichero, modelo: sinAbrir.Modelo, modo: sinAbrir.Modo, preguntaAmpliada: sinAbrir.PruebaDeRed,
		})
		serie := &series[enLaSerie]
		serie.sinAbrir++
		serie.tasa.SinMedir++
	}

	for posicion := range series {
		tasa := &series[posicion].tasa
		tasa.Pasa = tasa.SinMedir == 0 && tasa.Pasan >= e.Umbral
		tasa.Formas = formasExigidas(evals, tasa.Eval)
	}

	return series
}

// recontarRespuestas da el recuento de las respuestas medidas del modelo que
// decide (contracts/informe-del-job.md §1 de H7.4; data-model §5 de H7.4;
// research D13 de H7.4; FR-045 y FR-061 de H7.4): sus respuestas medidas —las
// sesiones juzgadas, no las ilegibles, que no tienen respuesta juzgada, ni las
// sin medir, que no se midieron (FR-002 de H7.3), ni las sin terminar, cuya
// respuesta no es la de una sesión que acabó— de las series que pide el plan
// cuya eval espera que la skill se active, y cuántas de ellas no activaron la
// skill. Las de una serie que el plan no pide —la de la prueba de red o la de
// una sesión cuya eval, modelo o pregunta no se pudieron leer— y las sin
// terminar no cuentan.
//
// Desde H21 (contracts/evals-en-dos-modos.md §5 de H21; research.md D19; FR-043,
// FR-047), hay un elemento por modo del plan, el del modo orden delante, y cada
// uno cuenta solo las respuestas de las series de su modo: las de un modo no se
// suman a las del otro. Las series de una eval sin binario ni servidor, que el
// plan pide sin modo, no cuentan en ninguno, aunque su eval espere que la skill
// se active.
//
// Desde H24 (research D13 de H24; FR-034, FR-037), el recuento existe si la
// skill tiene juez, y no si tiene lista de expresiones: nil sin él, y la skill
// no tiene ningún umbral de sus respuestas. Las de los modelos informativos no
// se cuentan: ningún umbral se mide sobre ellas.
func recontarRespuestas(
	e InformeAEscribir, juez *Juez, sesiones []sesionJuzgada, series []serieJuzgada,
) []recuentoDeRespuestas {
	if juez == nil {
		return nil
	}

	modos := e.plan(nil).modos()

	recuento := make([]recuentoDeRespuestas, 0, len(modos))
	for _, modo := range modos {
		recuento = append(recuento, recuentoDeRespuestas{modelo: e.ModeloQueDecide, modo: modo})
	}

	for _, serie := range series {
		// Las series que pide el plan con el modelo que decide son, las que
		// tienen modo, de uno de los suyos.
		enElRecuento := slices.IndexFunc(recuento, func(r recuentoDeRespuestas) bool {
			return r.modelo == serie.clave.modelo && r.modo == serie.clave.modo
		})
		if !serie.tasa.Planificada || enElRecuento < 0 {
			continue
		}

		delModo := &recuento[enElRecuento]

		for _, posicion := range serie.sesiones {
			juzgada := sesiones[posicion]
			if juzgada.ilegible || juzgada.resultado.SinMedir != "" || !juzgada.resultado.SesionTerminada ||
				!juzgada.resultado.Activa {
				continue
			}

			delModo.contar(juzgada.resultado)
		}
	}

	return recuento
}

// formasExigidas son las formas fijas que exige la eval del fichero dado entre
// las bien formadas: la de cada clase de sus hallazgos, en su orden, escrita con
// formaEscrita y la etiqueta de grafo.EtiquetasDeHallazgo (contrato
// evals-y-skill §6 de H7.1), y, detrás, la de version-obsoleta seguida de un
// espacio y el texto de cada redacción modificada que espera, en su orden
// (contracts/evals-y-juicio.md §2 y contracts/informe-del-job.md §4 de H7.4).
// Vacía, nunca nil, si la eval no espera hallazgos ni redacciones modificadas o
// si el fichero no es el de ninguna eval bien formada.
func formasExigidas(evals []Eval, fichero string) []string {
	formas := []string{}

	posicion := slices.IndexFunc(evals, func(eval Eval) bool { return eval.Fichero == fichero })
	if posicion < 0 {
		return formas
	}

	etiquetas := grafo.EtiquetasDeHallazgo()
	for _, clase := range evals[posicion].Hallazgos {
		formas = append(formas, formaEscrita(etiquetas[grafo.ClaseDeHallazgo(clase)]))
	}

	deLaRedaccion := formaEscrita(etiquetas[grafo.ClaseVersionObsoleta])
	for _, redaccion := range evals[posicion].RedaccionesModificadas {
		formas = append(formas, deLaRedaccion+" "+redaccion.texto())
	}

	return formas
}

// agregarSinRepetir añade el valor a la lista si no está ya. Un valor vacío no se
// añade: es el de una sesión cuyo transcript no llegó a declararlo.
func agregarSinRepetir(lista []string, valor string) []string {
	if valor == "" || slices.Contains(lista, valor) {
		return lista
	}

	return append(lista, valor)
}

// motivosDelInforme son los motivos de la raíz del informe en el orden de
// data-model §10.3, y son exactamente las causas del veredicto fallo: por serie
// planificada, las sesiones que faltan o sobran —las que el repartidor no abrió
// cuentan como de la serie— y, si decide, no llega al umbral y no tiene ninguna
// sin medir, su tasa con los motivos de sus sesiones que no pasan; los de cada
// sesión ilegible que no se hayan escrito ya; que no haya ninguna eval que
// juzgar; cada fichero mal formado; cada petición llegada a la red; el de cada
// umbral que decide y no se cumple, salvo los de la duración (FR-003 de H7.3;
// contrato informe-del-job §2.1); con alguna sesión sin medir, uno solo de la
// ejecución que las nombra (FR-043 de H7.3; contrato informe-del-job §2.2); y
// los de la duración de cada modo, también de la ejecución, si deciden y no se
// cumplen (FR-051 de H7.3; contrato informe-del-job §2.3). Una sesión que no
// pasa de una serie que sí llega al umbral no da ningún motivo: eso es lo que el
// umbral absorbe (ADR 0016); y una serie sin medir ni pasa ni falla (FR-042 de
// H7.3). Los de una serie nombran su modelo, y los de un umbral, su nombre: con
// ellos nombran su modo (contracts/evals-en-dos-modos.md §5.2 de H21; FR-044).
func motivosDelInforme(
	e InformeAEscribir, informe Informe, evals []Eval, sesiones []sesionJuzgada, series []serieJuzgada,
) []string {
	var motivos []string

	escritas := map[string]bool{}

	for _, serie := range series {
		motivos = append(motivos, motivosDeLaSerie(e, serie, sesiones, escritas)...)
	}

	for _, juzgada := range sesiones {
		if juzgada.ilegible {
			motivos = append(motivos, motivosDeLaSesion(juzgada, escritas)...)
		}
	}

	if len(evals) == 0 {
		motivos = append(motivos, motivoSinEvalsQueJuzgar)
	}

	for _, malFormado := range informe.FicherosMalFormados {
		motivos = append(motivos, malFormado.Fichero+motivoDeFicheroMalFormado+malFormado.Error)
	}

	for _, llegada := range informe.Red {
		motivos = append(motivos, llegada.Sesion+motivoDeLlegadaALaRed+llegada.Orden+" → "+llegada.Destino)
	}

	motivos = append(motivos, motivosDeLosUmbrales(informe.Umbrales)...)

	if len(informe.SesionesSinMedir) > 0 {
		motivos = append(motivos, motivoDeLasSesionesSinMedir(informe.SesionesSinMedir))
	}

	return append(motivos, motivosDeLaDuracionDeLasSesiones(informe.Umbrales)...)
}

// motivosDeLaSerie son los motivos de la raíz que da una serie: si el plan la
// pide, el de las sesiones que faltan o sobran, contando como suyas las que el
// repartidor no abrió; y, si decide, no llega al umbral y no tiene ninguna sin
// medir, su tasa seguida de los motivos de sus sesiones que no pasan que no se
// hayan escrito ya.
func motivosDeLaSerie(e InformeAEscribir, serie serieJuzgada, sesiones []sesionJuzgada, escritas map[string]bool) []string {
	var motivos []string

	if hay := serie.tasa.Sesiones + serie.sinAbrir; serie.tasa.Planificada && hay != e.Repeticiones {
		motivos = append(motivos, fmt.Sprintf("%s con %s: hay %d sesiones y el plan pide %d",
			serie.tasa.Eval, serie.tasa.Modelo, hay, e.Repeticiones))
	}

	if !serie.tasa.Decide || serie.tasa.Pasa || serie.tasa.SinMedir > 0 {
		return motivos
	}

	motivos = append(motivos, fmt.Sprintf("%s con %s: pasan %d de %d, y el umbral es %d",
		serie.tasa.Eval, serie.tasa.Modelo, serie.tasa.Pasan, serie.tasa.Sesiones, e.Umbral))

	for _, posicion := range serie.sesiones {
		motivos = append(motivos, motivosDeLaSesion(sesiones[posicion], escritas)...)
	}

	return motivos
}

// sesionesSinMedir son las sesiones sin medir del informe, en orden de sesión
// (data-model §3 de H7.3): las juzgadas que un límite de uso no dejó terminar,
// con la descripción de su clase, y las que el repartidor no abrió tras él, con
// «sin abrir tras el límite de uso». Nil, [] en informe.json, si no hay ninguna.
func sesionesSinMedir(resultados []ResultadoDeEval, sinAbrir []SesionPlanificada) []SesionSinMedir {
	var sinMedir []SesionSinMedir

	for _, resultado := range resultados {
		if resultado.SinMedir != "" {
			sinMedir = append(sinMedir, SesionSinMedir{
				Sesion: resultado.Sesion, Eval: resultado.Eval, Modelo: resultado.Modelo, Motivo: resultado.SinMedir,
			})
		}
	}

	for _, planificada := range sinAbrir {
		sinMedir = append(sinMedir, SesionSinMedir{
			Sesion: planificada.Nombre, Eval: planificada.Fichero, Modelo: planificada.Modelo, Motivo: descripcionSinAbrir,
		})
	}

	slices.SortStableFunc(sinMedir, func(a, b SesionSinMedir) int { return cmp.Compare(a.Sesion, b.Sesion) })

	return sinMedir
}

// motivoDeLasSesionesSinMedir es el motivo de la raíz que dan las sesiones sin
// medir: el prefijo de la ejecución, el límite de uso, cuántas son y cada una
// con su motivo, en su orden (contrato informe-del-job §2.2 de H7.3).
func motivoDeLasSesionesSinMedir(sinMedir []SesionSinMedir) string {
	nombradas := make([]string, 0, len(sinMedir))
	for _, sesion := range sinMedir {
		nombradas = append(nombradas, sesion.Sesion+" ("+sesion.Motivo+")")
	}

	return motivoDeLaEjecucion + fmt.Sprintf(motivoDelLimiteDeUso, len(sinMedir), strings.Join(nombradas, ", "))
}

// motivosDeLaSesion son los motivos de una sesión que no pasa, precedidos de su
// nombre, y ninguno si pasa o si ya se escribieron: escritas anota las sesiones
// cuyos motivos ya están en la raíz, para que ninguno salga dos veces.
func motivosDeLaSesion(juzgada sesionJuzgada, escritas map[string]bool) []string {
	nombre := juzgada.resultado.Sesion
	if juzgada.resultado.Pasa || escritas[nombre] {
		return nil
	}

	escritas[nombre] = true

	motivos := make([]string, 0, len(juzgada.resultado.Motivos))
	for _, motivo := range juzgada.resultado.Motivos {
		motivos = append(motivos, nombre+": "+motivo)
	}

	return motivos
}

// veredictoDelInforme es fallo si hay algún motivo y aprobado si no hay ninguno:
// los motivos son las causas del fallo, así que el veredicto es su presencia
// (data-model §10.3).
func veredictoDelInforme(informe Informe) Veredicto {
	if len(informe.Motivos) > 0 {
		return VeredictoFallo
	}

	return VeredictoAprobado
}

// escribirElInforme escribe informe.md y después informe.json en el destino. Si
// el segundo no se puede escribir, retira el primero: ningún informe queda a
// medias. El error nombra el destino y lleva el del fichero.
func escribirElInforme(destino string, md, codificado []byte) error {
	rutaDelMD := filepath.Join(destino, ficheroDelInformeMD)

	if err := escribirFichero(rutaDelMD, md); err != nil {
		return fmt.Errorf("el informe no se puede escribir en %s: %w", destino, err)
	}

	if err := escribirFichero(filepath.Join(destino, ficheroDelInformeJSON), codificado); err != nil {
		return errors.Join(fmt.Errorf("el informe no se puede escribir en %s: %w", destino, err), retirarFichero(rutaDelMD))
	}

	return nil
}

// retirarFichero borra un fichero ya escrito.
func retirarFichero(ruta string) error {
	if err := os.Remove(ruta); err != nil {
		return fmt.Errorf("%s, ya escrito, no se puede retirar: %w", ruta, err)
	}

	return nil
}

// renderizarInforme da informe.md (contrato job-de-evals §5): el título; el
// veredicto y sus motivos; la cabecera, con la duración de las sesiones y los
// reintentos por límite de ritmo al final (contrato informe-del-job §4 de H7.3);
// la comprobación sin Python en un bloque; los ficheros mal formados, las
// invocaciones fuera de lo grabado y las peticiones llegadas a la red; las tasas
// por eval; los umbrales, o «ninguno»; las sesiones sin medir, o «ninguna»; la
// tabla de las sesiones; y una sección por sesión. Las tablas de las tasas y de
// las sesiones llevan el modo detrás del modelo, y la de las invocaciones de
// cada sesión, si cada una es una llamada a una herramienta
// (contracts/evals-en-dos-modos.md §5.3 de H21). Desde H24 no lleva la sección
// «Expresiones prohibidas por modelo» ni la columna «Expresiones prohibidas» de
// la tabla de las sesiones (contracts/informe-del-job.md §7 de H24; FR-062).
func renderizarInforme(informe Informe, sesiones []sesionJuzgada) []byte {
	var md documento

	md.parrafo("# Informe de evals de " + informe.Skill)

	md.parrafo("## Veredicto")
	md.parrafo("Veredicto: " + string(informe.Veredicto))
	md.listaConEtiqueta("Motivos", informe.Motivos)

	md.parrafo("## Cabecera")
	md.parrafo("Modelo que decide: " + informe.ModeloQueDecide)
	md.parrafo("Modelos informativos: " + unidosOVacio(informe.ModelosInformativos, ningunoEnElInforme))
	md.parrafo("Repeticiones por eval: " + strconv.Itoa(informe.Repeticiones))
	md.parrafo("Umbral: " + strconv.Itoa(informe.Umbral))
	md.parrafo("Modelos de las sesiones: " + unidosOVacio(informe.ModelosDeSesion, ningunoEnElInforme))
	md.parrafo("Versiones de Claude Code: " + unidosOVacio(informe.VersionesDeClaudeCode, ningunaEnElInforme))
	md.parrafo("Commit: " + informe.Commit)
	md.parrafo("Duración de las sesiones: " + strconv.Itoa(informe.DuracionDeLasSesiones) + " s")
	md.parrafo("Reintentos por límite de ritmo: " + strconv.Itoa(informe.ReintentosPorLimiteDeRitmo))

	md.parrafo("## Comprobación sin Python")
	md.bloqueDeTexto(informe.SinPython)

	md.parrafo("## Ficheros mal formados")

	if len(informe.FicherosMalFormados) == 0 {
		md.parrafo(ningunoEnElInforme)
	} else {
		lineas := make([]string, 0, len(informe.FicherosMalFormados))
		for _, malFormado := range informe.FicherosMalFormados {
			lineas = append(lineas, malFormado.Fichero+": "+malFormado.Error)
		}

		md.lista(lineas)
	}

	md.parrafo("## Invocaciones fuera de lo grabado")
	md.tablaOVacia(encabezadosDeFueraDeLoGrabado, filasDeFueraDeLoGrabado(informe.FueraDeLoGrabado), ningunaEnElInforme)

	md.parrafo("## Peticiones llegadas a la red")
	md.tablaOVacia(encabezadosDeRed, filasDeRed(informe.Red), sinLlegadasALaRed)

	md.parrafo("## Tasas por eval")
	md.tablaOVacia(encabezadosDeTasas, filasDeTasas(informe.Tasas), ningunaEnElInforme)

	md.parrafo("## Umbrales")
	md.tablaOVacia(encabezadosDeUmbrales, filasDeUmbrales(informe.Umbrales), ningunoEnElInforme)

	md.parrafo("## Sesiones sin medir")
	md.tablaOVacia(encabezadosDeSinMedir, filasDeSinMedir(informe.SesionesSinMedir), ningunaEnElInforme)

	md.parrafo("## Sesiones")
	md.tablaOVacia(encabezadosDeSesiones, filasDeSesiones(informe.Evals), ningunaEnElInforme)

	for _, juzgada := range sesiones {
		md.sesion(juzgada)
	}

	return md.unido()
}

// sesion añade la sección de una sesión: su eval, la pregunta, las invocaciones
// con su código y sus conexiones, la respuesta y, si la sesión no terminó, no se
// pudo leer o quedó sin medir, sus motivos de sesión ilegible, sin terminar o sin
// medir y su salida de error.
func (d *documento) sesion(juzgada sesionJuzgada) {
	resultado := juzgada.resultado

	d.parrafo("## Sesión " + resultado.Sesion)
	d.parrafo("Eval: " + cmp.Or(resultado.Eval, ningunaEnElInforme))
	d.parrafo("Modelo pedido: " + cmp.Or(resultado.Modelo, sinLeer))
	d.parrafo("Modelo de la sesión: " + cmp.Or(resultado.ModeloDeLaSesion, sinLeer))
	d.textoLeido("Pregunta", juzgada.pregunta, juzgada.preguntaLeida)

	switch {
	case !juzgada.trazaLeida:
		d.parrafo("Invocaciones: " + sinLeer)
	case len(resultado.Invocaciones) == 0:
		d.parrafo("Invocaciones: " + ningunaEnElInforme)
	default:
		d.parrafo("Invocaciones:")
		d.tablaOVacia(encabezadosDeInvocaciones, filasDeInvocaciones(resultado.Invocaciones), ningunaEnElInforme)
	}

	d.textoLeido("Respuesta", resultado.Respuesta, juzgada.leida)

	if resultado.SesionTerminada && !juzgada.ilegible && resultado.SinMedir == "" {
		return
	}

	var motivos []string

	for _, motivo := range resultado.Motivos {
		if strings.HasPrefix(motivo, motivoDeSesionIlegible) || strings.HasPrefix(motivo, motivoDeSesionSinTerminar) ||
			strings.HasPrefix(motivo, motivoSinMedir) {
			motivos = append(motivos, motivo)
		}
	}

	d.listaConEtiqueta("Motivos de la sesión", motivos)
	d.textoLeido("Salida de error", juzgada.sesion.SalidaDeError, juzgada.leida)
}

// filasDeFueraDeLoGrabado son las filas de la tabla de invocaciones fuera de lo
// grabado.
func filasDeFueraDeLoGrabado(fuera []FueraDeLoGrabadoDelInforme) [][]string {
	filas := make([][]string, 0, len(fuera))
	for _, invocacion := range fuera {
		filas = append(filas, []string{invocacion.Sesion, invocacion.Eval, invocacion.Orden, strconv.Itoa(invocacion.Codigo)})
	}

	return filas
}

// filasDeRed son las filas de la tabla de peticiones llegadas a la red.
func filasDeRed(red []RedDelInforme) [][]string {
	filas := make([][]string, 0, len(red))
	for _, llegada := range red {
		filas = append(filas, []string{llegada.Sesion, llegada.Eval, llegada.Orden, llegada.Destino})
	}

	return filas
}

// filasDeTasas son las filas de la tabla de las series: eval, modelo, modo, si
// decide, si el plan la pide, las formas que exige su eval junto a la tasa
// (contrato evals-y-skill §6 de H7.1), la tasa «<pasan> de <sesiones>» y si llega
// al umbral o, con alguna sesión sin medir, cuántas (contrato informe-del-job §4
// de H7.3). La eval de una serie con la pregunta ampliada lleva detrás con qué
// se amplió, que es la prueba de red.
func filasDeTasas(tasas []TasaDelInforme) [][]string {
	filas := make([][]string, 0, len(tasas))

	for _, tasa := range tasas {
		eval := tasa.Eval
		if tasa.PreguntaAmpliada {
			eval += " (pregunta ampliada)"
		}

		filas = append(filas, []string{
			eval,
			tasa.Modelo,
			modoEscrito(tasa.Modo),
			siONo(tasa.Decide),
			siONo(tasa.Planificada),
			unidosOVacio(tasa.Formas, ningunaEnElInforme),
			strconv.Itoa(tasa.Pasan) + " de " + strconv.Itoa(tasa.Sesiones),
			resultadoDeLaSerie(tasa),
		})
	}

	return filas
}

// filasDeSinMedir son las filas de la tabla de las sesiones sin medir: sesión,
// eval, modelo y motivo.
func filasDeSinMedir(sinMedir []SesionSinMedir) [][]string {
	filas := make([][]string, 0, len(sinMedir))
	for _, sesion := range sinMedir {
		filas = append(filas, []string{sesion.Sesion, sesion.Eval, sesion.Modelo, sesion.Motivo})
	}

	return filas
}

// modoEscrito es el modo como lo escribe la columna «Modo» de informe.md:
// «orden», «herramienta» o, si no es ninguno, «—»
// (contracts/evals-en-dos-modos.md §5.3 de H21).
func modoEscrito(modo Modo) string {
	return cmp.Or(string(modo), sinModoEnElInforme)
}

// resultadoDeLaSerie dice cuántas sesiones de la serie quedaron sin medir, si
// alguna lo hizo, o si no, si la serie llega al umbral.
func resultadoDeLaSerie(tasa TasaDelInforme) string {
	switch {
	case tasa.SinMedir > 0:
		return fmt.Sprintf(serieSinMedir, tasa.SinMedir)
	case tasa.Pasa:
		return "llega al umbral"
	default:
		return "no llega al umbral"
	}
}

// filasDeSesiones son las filas de la tabla de las sesiones: sesión, eval, modelo,
// modo, activa, activada, sesión terminada con su código, comandos ausentes, comandos
// prohibidos ejecutados, citas ausentes, avisos encontrados, avisos ausentes,
// hallazgos encontrados, hallazgos ausentes, redacciones modificadas encontradas,
// redacciones modificadas ausentes, territorio encontrado, territorio ausente,
// reintentos por límite de ritmo, sin medir —«no» o la clase— y resultado. Los
// comandos prohibidos ejecutados
// van junto a los ausentes, cada uno con su texto (contrato evals-y-skill §2 de
// H7); los avisos, junto a las citas, cada uno con su código (contrato de formato,
// juicio e informe §5 de H5.1); los hallazgos, junto a los avisos, cada uno con su
// clase (contrato evals-y-skill §6 de H7.1); las redacciones modificadas, junto a
// los hallazgos, cada una con su texto (contracts/informe-del-job.md §4 de H7.4);
// el territorio, detrás, cada elemento con su texto (contrato de evals §2 de H6);
// y los reintentos y si quedó sin medir, detrás del territorio (contrato
// informe-del-job §4 de H7.3).
func filasDeSesiones(resultados []ResultadoDeEval) [][]string {
	filas := make([][]string, 0, len(resultados))

	for _, resultado := range resultados {
		codigo := "sin código"
		if resultado.CodigoDeLaSesion != nil {
			codigo = "código " + strconv.Itoa(*resultado.CodigoDeLaSesion)
		}

		pasa := "no pasa"
		if resultado.Pasa {
			pasa = "pasa"
		}

		filas = append(filas, []string{
			resultado.Sesion,
			resultado.Eval,
			resultado.Modelo,
			modoEscrito(resultado.Modo),
			siONo(resultado.Activa),
			siONo(resultado.Activada),
			siONo(resultado.SesionTerminada) + " (" + codigo + ")",
			unidosOVacio(resultado.ComandosAusentes, ningunoEnElInforme),
			unidosOVacio(resultado.ComandosProhibidosEjecutados, ningunoEnElInforme),
			unidosOVacio(resultado.CitasAusentes, ningunaEnElInforme),
			unidosOVacio(resultado.AvisosEncontrados, ningunoEnElInforme),
			unidosOVacio(resultado.AvisosAusentes, ningunoEnElInforme),
			unidosOVacio(resultado.HallazgosEncontrados, ningunoEnElInforme),
			unidosOVacio(resultado.HallazgosAusentes, ningunoEnElInforme),
			unidosOVacio(resultado.RedaccionesEncontradas, ningunaEnElInforme),
			unidosOVacio(resultado.RedaccionesAusentes, ningunaEnElInforme),
			unidosOVacio(resultado.TerritorioEncontrado, ningunoEnElInforme),
			unidosOVacio(resultado.TerritorioAusente, ningunoEnElInforme),
			strconv.Itoa(resultado.ReintentosPorLimiteDeRitmo),
			cmp.Or(resultado.SinMedir, medidaEnElInforme),
			pasa,
		})
	}

	return filas
}

// filasDeInvocaciones son las filas de la tabla de invocaciones de una sesión:
// la orden, su código o «sin código (sesión cortada)», sus conexiones, cada una
// con su destino y su clase, o «sin conexiones», y la marca de las que son una
// llamada a una herramienta del servidor y no una orden de la traza
// (contracts/evals-en-dos-modos.md §5.3 de H21).
func filasDeInvocaciones(invocaciones []InvocacionInformada) [][]string {
	filas := make([][]string, 0, len(invocaciones))

	for _, invocacion := range invocaciones {
		codigo := sinCodigoPorElCorte
		if invocacion.Codigo != nil {
			codigo = strconv.Itoa(*invocacion.Codigo)
		}

		conexiones := make([]string, 0, len(invocacion.Conexiones))
		for _, conexion := range invocacion.Conexiones {
			conexiones = append(conexiones, conexion.Destino+" ("+string(conexion.Clase)+")")
		}

		filas = append(filas, []string{
			invocacion.Orden, codigo, unidosOVacio(conexiones, sinConexiones), siONo(invocacion.Llamada),
		})
	}

	return filas
}

// siONo es «sí» o «no».
func siONo(valor bool) string {
	if valor {
		return "sí"
	}

	return "no"
}

// unidosOVacio son los elementos separados por «, », o el texto de la lista
// vacía.
func unidosOVacio(elementos []string, vacia string) string {
	return cmp.Or(strings.Join(elementos, ", "), vacia)
}

// documento son los bloques de informe.md —párrafo o título, lista, tabla o
// bloque de texto—, cada uno terminado en un salto de línea. Unidos, cada dos
// quedan separados por una línea en blanco.
type documento []string

// parrafo añade una línea de texto, que puede ser un título, con sus saltos de
// línea como espacios.
func (d *documento) parrafo(texto string) {
	*d = append(*d, enUnaLinea.Replace(texto)+"\n")
}

// lista añade una lista con un elemento por línea.
func (d *documento) lista(elementos []string) {
	var lista strings.Builder
	for _, elemento := range elementos {
		lista.WriteString("- " + enUnaLinea.Replace(elemento) + "\n")
	}

	*d = append(*d, lista.String())
}

// listaConEtiqueta añade «<etiqueta>: ninguno» o, con elementos, la etiqueta y
// una lista con uno por línea.
func (d *documento) listaConEtiqueta(etiqueta string, elementos []string) {
	if len(elementos) == 0 {
		d.parrafo(etiqueta + ": " + ningunoEnElInforme)

		return
	}

	d.parrafo(etiqueta + ":")
	d.lista(elementos)
}

// tablaOVacia añade una tabla con sus encabezados y una fila por elemento, con cada
// celda escapada, o el texto de la tabla vacía.
func (d *documento) tablaOVacia(encabezados []string, filas [][]string, vacia string) {
	if len(filas) == 0 {
		d.parrafo(vacia)

		return
	}

	var tabla strings.Builder

	for _, fila := range slices.Concat([][]string{encabezados, slices.Repeat([]string{"---"}, len(encabezados))}, filas) {
		celdas := make([]string, 0, len(fila))
		for _, celda := range fila {
			celdas = append(celdas, escapeDeCelda.Replace(celda))
		}

		tabla.WriteString("| " + strings.Join(celdas, " | ") + " |\n")
	}

	*d = append(*d, tabla.String())
}

// bloqueDeTexto añade el contenido tal cual entre una línea ```text y una línea
// ```. Si el contenido tiene una racha de tres o más comillas invertidas, la
// valla tiene una más que la más larga, para que ninguna línea del contenido
// cierre el bloque; y si no termina en un salto de línea, se le añade uno antes
// de la valla de cierre.
func (d *documento) bloqueDeTexto(contenido string) {
	valla := strings.Repeat("`", max(len("```"), rachaDeComillasMasLarga(contenido)+1))

	if !strings.HasSuffix(contenido, "\n") {
		contenido += "\n"
	}

	*d = append(*d, valla+"text\n"+contenido+valla+"\n")
}

// textoLeido añade «<etiqueta>: sin leer» si el texto no se leyó, «<etiqueta>:
// vacía» si está vacío y, si no, la etiqueta y el texto en un bloque.
func (d *documento) textoLeido(etiqueta, texto string, leido bool) {
	switch {
	case !leido:
		d.parrafo(etiqueta + ": " + sinLeer)
	case texto == "":
		d.parrafo(etiqueta + ": " + vaciaEnElInforme)
	default:
		d.parrafo(etiqueta + ":")
		d.bloqueDeTexto(texto)
	}
}

// unido es informe.md entero: los bloques separados por una línea en blanco, con
// el salto de línea final del último.
func (d documento) unido() []byte {
	return []byte(strings.Join(d, "\n"))
}

// rachaDeComillasMasLarga es la longitud de la racha más larga de comillas
// invertidas seguidas del texto.
func rachaDeComillasMasLarga(texto string) int {
	mayor, racha := 0, 0

	for _, caracter := range []byte(texto) {
		if caracter != '`' {
			racha = 0

			continue
		}

		racha++
		mayor = max(mayor, racha)
	}

	return mayor
}
