package evals

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/graph"
	"github.com/jmorenobl/kitlegal/internal/skills"
)

// MedidaDelJuez es lo que se lee de la medida versionada del juez de una skill,
// medida.json de su carpeta del juez (contracts/medida-del-juez.md §1 y
// data-model §4 de H24; FR-040): contra qué casos etiquetados se midió al juez,
// con qué modelo y qué versión de Claude Code, y qué dio. Corresponde si sus dos
// huellas, su modelo y su versión son los de lo que hay, y se cumple si sus dos
// recuentos son 0 (FR-041): lo dice comprobarLaMedida.
type MedidaDelJuez struct {
	// Clase es la clase que decide a la que se refiere.
	Clase string

	// Fecha es cuándo se midió.
	Fecha string

	// ModeloDelJuez y VersionDeClaudeCode son el id del modelo y la versión de
	// Claude Code de los votos de la medida.
	ModeloDelJuez       string
	VersionDeClaudeCode string

	// HuellaDeLaRubrica y HuellaDeLosCasos son las huellas SHA-256, en
	// hexadecimal, de la rúbrica y de los casos con los que se midió.
	HuellaDeLaRubrica string
	HuellaDeLosCasos  string

	// Defectos y Correctos son los casos etiquetados como defecto y como
	// correctos, con los que no dieron lo que dice su etiqueta.
	Defectos  DefectosDeLaMedida
	Correctos CorrectosDeLaMedida
}

// DefectosDeLaMedida son los casos etiquetados como defecto de una medida del
// juez. Sus claves JSON son las de defectos en la medida que da medirAlJuez.
type DefectosDeLaMedida struct {
	// Casos es cuántos son.
	Casos int `json:"casos"`

	// SinMarcar es cuántos de ellos no quedaron marcados: con alguno, la medida
	// no se cumple.
	SinMarcar int `json:"sin_marcar"`
}

// CorrectosDeLaMedida son los casos etiquetados como correctos de una medida del
// juez. Sus claves JSON son las de correctos en la medida que da medirAlJuez.
type CorrectosDeLaMedida struct {
	// Casos es cuántos son.
	Casos int `json:"casos"`

	// Marcados es cuántos de ellos quedaron marcados: con alguno, la medida no
	// se cumple.
	Marcados int `json:"marcados"`
}

// clavesDeLaMedida son las claves de medida.json que lee leerMedidaDelJuez; las
// demás —skill, votos, origen y el fichero de cada huella— no se leen. Cada una
// es un puntero, que distingue la clave que falta, o que está a null, de la que
// lleva el valor cero: un recuento a 0 es justo el que se cumple.
type clavesDeLaMedida struct {
	Clase               *string `json:"clase"`
	Fecha               *string `json:"fecha"`
	ModeloDelJuez       *string `json:"modelo_del_juez"`
	VersionDeClaudeCode *string `json:"version_de_claude_code"`

	Rubrica struct {
		SHA256 *string `json:"sha256"`
	} `json:"rubrica"`

	Casos struct {
		SHA256 *string `json:"sha256"`
	} `json:"casos"`

	Defectos struct {
		Casos     *int `json:"casos"`
		SinMarcar *int `json:"sin_marcar"`
	} `json:"defectos"`

	Correctos struct {
		Casos    *int `json:"casos"`
		Marcados *int `json:"marcados"`
	} `json:"correctos"`
}

// leerMedidaDelJuez lee la medida del juez de su contenido, el de medida.json
// (contracts/medida-del-juez.md §1 de H24). No tiene esquema publicado: se lee
// como un objeto JSON del que solo se miran las diez claves de MedidaDelJuez,
// cada una con su tipo, un texto o un entero; una clave repetida es un error,
// nunca la última que gana. Es un error que no sea ese documento o que le falte
// alguna de las diez, y una clave a null es una clave que falta: el error las
// nombra todas, en el orden del contrato, con las de un objeto detrás de la
// suya y de un punto, como rubrica.sha256. Ningún error lleva delante de qué
// medida habla: lo pone quien la lee.
func leerMedidaDelJuez(contenido []byte) (MedidaDelJuez, error) {
	var leidas clavesDeLaMedida

	if err := json.Unmarshal(contenido, &leidas); err != nil {
		return MedidaDelJuez{}, fmt.Errorf("no se puede leer como JSON: %w", err)
	}

	var faltan []string

	medida := MedidaDelJuez{
		Clase:               valorDeLaClave(leidas.Clase, "clase", &faltan),
		Fecha:               valorDeLaClave(leidas.Fecha, "fecha", &faltan),
		ModeloDelJuez:       valorDeLaClave(leidas.ModeloDelJuez, "modelo_del_juez", &faltan),
		VersionDeClaudeCode: valorDeLaClave(leidas.VersionDeClaudeCode, "version_de_claude_code", &faltan),
		HuellaDeLaRubrica:   valorDeLaClave(leidas.Rubrica.SHA256, "rubrica.sha256", &faltan),
		HuellaDeLosCasos:    valorDeLaClave(leidas.Casos.SHA256, "casos.sha256", &faltan),
		Defectos: DefectosDeLaMedida{
			Casos:     valorDeLaClave(leidas.Defectos.Casos, "defectos.casos", &faltan),
			SinMarcar: valorDeLaClave(leidas.Defectos.SinMarcar, "defectos.sin_marcar", &faltan),
		},
		Correctos: CorrectosDeLaMedida{
			Casos:    valorDeLaClave(leidas.Correctos.Casos, "correctos.casos", &faltan),
			Marcados: valorDeLaClave(leidas.Correctos.Marcados, "correctos.marcados", &faltan),
		},
	}

	if len(faltan) > 0 {
		return MedidaDelJuez{}, fmt.Errorf("claves que faltan: %s", strings.Join(faltan, ", "))
	}

	return medida, nil
}

// valorDeLaClave es el valor leído de una clave de la medida. Si la clave no
// estaba, o estaba a null, la anota en faltan y devuelve el valor cero.
func valorDeLaClave[T any](leido *T, clave string, faltan *[]string) T {
	if leido == nil {
		*faltan = append(*faltan, clave)

		var cero T

		return cero
	}

	return *leido
}

// Las líneas de comprobarLaMedida: las de contracts/medida-del-juez.md §2 de
// H24, carácter a carácter. Cada fichero lleva el nombre que tiene dentro de la
// carpeta de evals de la skill, juez/<fichero>, como en los errores del
// conjunto.
const (
	// medidaQueNoCorresponde es la de la medida que no es su documento o a la
	// que le falta alguna de sus claves; sigue con el error de
	// leerMedidaDelJuez.
	medidaQueNoCorresponde = "la medida versionada (" + carpetaDelJuez + "/" + ficheroDeMedidaDelJuez + ") no corresponde: %v"

	rubricaDeOtraMedida = "la rúbrica (" + carpetaDelJuez + "/" + ficheroDeRubricaDelJuez + ") no es la de la medida versionada"
	casosDeOtraMedida   = "los casos (" + carpetaDelJuez + "/" + ficheroDeCasosDelJuez + ") no son los de la medida versionada"

	medidaDeOtroModelo  = "la medida versionada es del modelo %s y el fijado para el juez es %s"
	medidaDeOtraVersion = "la medida versionada es de la versión %s de Claude Code y la fijada para los votos del juez es %s"
	medidaDeOtraClase   = "la clase %s decide y la medida versionada es de %s"

	medidaConDefectosSinMarcar = "la medida versionada no se cumple: %d defectos sin marcar"
	medidaConCorrectosMarcados = "la medida versionada no se cumple: %d correctos marcados"
)

// comprobarLaMedida dice si la medida versionada del juez de una skill
// corresponde y se cumple (contracts/medida-del-juez.md §2 de H24; FR-041):
// devuelve una línea por lo que falla, en este orden, y ninguna si no falla
// nada.
//
//   - La huella SHA-256 de la rúbrica del juez no es la de la medida.
//   - La de sus casos no lo es.
//   - El modelo de la medida no es el fijado para el juez.
//   - Su versión de Claude Code no es la fijada para los votos del juez.
//   - Una clase que decide no es la de la medida: una línea por cada una, que
//     una medida es de una sola clase (research D19 de H24).
//   - Su recuento de defectos sin marcar no es 0.
//   - Su recuento de correctos marcados no es 0.
//
// Las cuatro primeras son las de una medida que no corresponde, y las dos
// últimas, las de una que no se cumple. La medida que no es su documento, o sin
// alguna de las claves que se leen, tampoco corresponde, y da una sola línea,
// con el error de leerMedidaDelJuez: sin la medida entera no hay con qué
// comparar lo demás.
//
// El modelo y la versión fijados los da quien la llama, y el juez es el de la
// skill, nunca nil. Compara lo que el juez lleva leído de su carpeta, y nada
// más: no lee ningún fichero, no usa ningún modelo ni abre ningún proceso.
func comprobarLaMedida(juez *Juez, modeloFijado, versionFijada string) []string {
	medida, err := leerMedidaDelJuez([]byte(juez.Medida))
	if err != nil {
		return []string{fmt.Sprintf(medidaQueNoCorresponde, err)}
	}

	lineas := lineasDeLasHuellas(juez, medida)
	lineas = append(lineas, lineasDeLoFijado(medida, modeloFijado, versionFijada)...)
	lineas = append(lineas, lineasDeLasClases(juez.Clases, medida.Clase)...)

	return append(lineas, lineasDeLosRecuentos(medida)...)
}

// lineasDeLasHuellas compara las huellas SHA-256 de la rúbrica del juez, que es
// la que va tal cual como sus instrucciones, y de sus casos con las de la
// medida.
func lineasDeLasHuellas(juez *Juez, medida MedidaDelJuez) []string {
	var lineas []string

	if huellaSHA256([]byte(juez.Rubrica)) != medida.HuellaDeLaRubrica {
		lineas = append(lineas, rubricaDeOtraMedida)
	}

	if huellaSHA256([]byte(juez.Casos)) != medida.HuellaDeLosCasos {
		lineas = append(lineas, casosDeOtraMedida)
	}

	return lineas
}

// lineasDeLoFijado compara el modelo y la versión de Claude Code de la medida
// con los fijados para el juez y para sus votos.
func lineasDeLoFijado(medida MedidaDelJuez, modeloFijado, versionFijada string) []string {
	var lineas []string

	if medida.ModeloDelJuez != modeloFijado {
		lineas = append(lineas, fmt.Sprintf(medidaDeOtroModelo, medida.ModeloDelJuez, modeloFijado))
	}

	if medida.VersionDeClaudeCode != versionFijada {
		lineas = append(lineas, fmt.Sprintf(medidaDeOtraVersion, medida.VersionDeClaudeCode, versionFijada))
	}

	return lineas
}

// lineasDeLasClases compara la clase de la medida con cada clase del juez que
// decide, en su orden: las que solo se publican no necesitan medida.
func lineasDeLasClases(clases []ClaseDelJuez, deLaMedida string) []string {
	var lineas []string

	for _, clase := range clases {
		if clase.Decide && clase.Nombre != deLaMedida {
			lineas = append(lineas, fmt.Sprintf(medidaDeOtraClase, clase.Nombre, deLaMedida))
		}
	}

	return lineas
}

// lineasDeLosRecuentos compara con 0 los dos recuentos de la medida: los
// defectos que no quedaron marcados y los correctos que sí.
func lineasDeLosRecuentos(medida MedidaDelJuez) []string {
	var lineas []string

	if medida.Defectos.SinMarcar != 0 {
		lineas = append(lineas, fmt.Sprintf(medidaConDefectosSinMarcar, medida.Defectos.SinMarcar))
	}

	if medida.Correctos.Marcados != 0 {
		lineas = append(lineas, fmt.Sprintf(medidaConCorrectosMarcados, medida.Correctos.Marcados))
	}

	return lineas
}

// huellaSHA256 es la huella SHA-256 de un contenido, en hexadecimal y en
// minúsculas, como la lleva una medida.
func huellaSHA256(contenido []byte) string {
	suma := sha256.Sum256(contenido)

	return hex.EncodeToString(suma[:])
}

// Las dos etiquetas de un caso etiquetado (contracts/medida-del-juez.md §4 de
// H24; data-model §4): la del que tiene el defecto de la clase, que el juez
// tiene que marcar, y la del que no lo tiene, que no puede marcar.
const (
	etiquetaDefecto  = "defecto"
	etiquetaCorrecto = "correcto"
)

// CasosEtiquetados son los casos etiquetados con los que se mide al juez de una
// skill, el documento casos.yaml de su carpeta del juez
// (contracts/medida-del-juez.md §4 de H24; FR-024).
type CasosEtiquetados struct {
	// Clase es la clase del juez de la que son los casos.
	Clase string `yaml:"clase"`

	// Casos son los casos, en el orden del fichero.
	Casos []CasoEtiquetado `yaml:"casos"`
}

// CasoEtiquetado es un caso de los casos etiquetados del juez de una skill: la
// respuesta de una sesión de un informe versionado —el de un job de evals o el
// de un sondeo—, con la etiqueta de lo que el juez tiene que decir de ella
// (data-model §4 y contracts/medida-del-juez.md §4 de H24; data-model §3 de
// H25; FR-024). Nada suyo está escrito a mano: el fichero dice de dónde sale, y
// resolverlo (reconstructor.resolver) le pone la pregunta, la respuesta y los
// textos.
type CasoEtiquetado struct {
	// Informe es la ruta del informe versionado del que sale, desde la raíz del
	// repositorio y con la barra de separador. Su tipo no lo dice el caso: es
	// el de un sondeo si lleva la clave sondeo (research D8 de H25).
	Informe string `yaml:"informe"`

	// Sesion es el nombre de la sesión de ese informe cuya respuesta es la del
	// caso.
	Sesion string `yaml:"sesion"`

	// Quitado es, en un derivado, lo que se quita de su sesión, que es lo que
	// hace de su respuesta un defecto por construcción o lo que la deja como
	// estaba: el texto de un bloque, de los textos de sus herramientas, o una
	// parte del texto pegado, de su pregunta. nil en los demás.
	Quitado *Quitado `yaml:"quitado"`

	// Grupo es el de la validación en la que se usó el caso, ajuste o medida.
	Grupo string `yaml:"grupo"`

	// Etiqueta es etiquetaDefecto o etiquetaCorrecto.
	Etiqueta string `yaml:"etiqueta"`

	// Procedencia dice de dónde sale la etiqueta: bitacora, derivado o lectura.
	Procedencia string `yaml:"procedencia"`

	// Frase es, en un caso de la lectura, la frase de la respuesta que decide su
	// etiqueta; vacía en los demás.
	Frase string `yaml:"frase"`

	// Pregunta es la de la sesión —la del fichero de eval que nombra o, en un
	// sondeo, la de sus preguntas— y, en un derivado por texto, sin la parte
	// quitada del que lleva pegado. No es una clave del fichero: la pone la
	// resolución.
	Pregunta string `yaml:"-"`

	// Respuesta es la de la sesión en su informe, tal cual. La pone la
	// resolución.
	Respuesta string `yaml:"-"`

	// Textos son los que devolvieron las herramientas de la sesión,
	// reconstruidos o, en un sondeo, leídos de su informe; en un derivado por
	// bloque, sin el del bloque quitado, y en uno por texto, los que la sesión
	// habría tenido con lo que queda del texto pegado. Los pone la resolución.
	Textos []Texto `yaml:"-"`
}

// Quitado es lo que un caso derivado quita de su sesión, en una de dos formas
// (data-model §3 y research D3 de H25): el texto de un bloque de una norma, de
// los textos de sus herramientas, que es la de los derivados de
// boe-legislacion; o una parte del texto pegado, de su pregunta, que es la de
// los de jurisprudencia. El fichero de los casos lleva además, en estos, la
// clave regla, que dice con qué cuenta se etiquetó el derivado y no se lee.
type Quitado struct {
	// Norma es el identificador de la norma del bloque, como BOE-A-2015-10565.
	Norma string `yaml:"norma"`

	// Bloque es el id del bloque, como a21.
	Bloque string `yaml:"bloque"`

	// Texto es la parte del texto pegado que se quita: quitarElDocumento,
	// quitarElFallo o quitarElApartado2.
	Texto string `yaml:"texto"`
}

// nombre nombra el caso en un error: su informe y su sesión y, en un derivado,
// lo que se quita, como «<informe> <sesión> sin <norma> <bloque>» o, si es una
// parte del texto pegado, «<informe> <sesión> sin <texto>».
func (c CasoEtiquetado) nombre() string {
	switch {
	case c.Quitado == nil:
		return c.Informe + " " + c.Sesion
	case c.Quitado.Texto != "":
		return fmt.Sprintf("%s %s sin %s", c.Informe, c.Sesion, c.Quitado.Texto)
	default:
		return fmt.Sprintf("%s %s sin %s %s", c.Informe, c.Sesion, c.Quitado.Norma, c.Quitado.Bloque)
	}
}

// textoQuitado es la parte del texto pegado que quita el caso, o vacío si no
// quita ninguna: no es un derivado, o lo es por un bloque.
func (c CasoEtiquetado) textoQuitado() string {
	if c.Quitado == nil {
		return ""
	}

	return c.Quitado.Texto
}

// leerCasosEtiquetados lee los casos etiquetados del juez de su contenido, el de
// casos.yaml (contracts/medida-del-juez.md §4 de H24), con el lector común de
// documentos YAML de internal/skills —una clave repetida es un defecto con sus
// dos líneas, nunca la última que gana— y sin esquema publicado: de un caso
// solo exige que su etiqueta sea una de las dos, porque con otra no se podría
// contar ni como defecto ni como correcto. Lo demás lo exige resolverlo: un caso
// que no se puede resolver es un error de la reconstrucción. Ningún error lleva
// delante de qué casos habla: lo pone quien los lee.
func leerCasosEtiquetados(contenido []byte) (CasosEtiquetados, error) {
	leidos, err := skills.ValidarDocumentoYAML[CasosEtiquetados](contenido, nil)
	if err != nil {
		return CasosEtiquetados{}, err
	}

	for posicion, caso := range leidos.Casos {
		if caso.Etiqueta != etiquetaDefecto && caso.Etiqueta != etiquetaCorrecto {
			return CasosEtiquetados{}, fmt.Errorf("el caso %d (%s) tiene la etiqueta %q, que no es %s ni %s",
				posicion+1, caso.nombre(), caso.Etiqueta, etiquetaDefecto, etiquetaCorrecto)
		}
	}

	return leidos, nil
}

// vigenciaDeLaBase es lo que sirve la base de una reconstrucción desde que se
// empieza a preparar: con más, se rehace, porque lo que la caché guarda de
// buscar y de metadatos caduca a los 300 s y una orden con --offline no sirve lo
// caducado (contracts/medida-del-juez.md §5, paso 1, de H24).
const vigenciaDeLaBase = 100 * time.Second

// Lo que la reconstrucción crea en el directorio temporal, que retira antes de
// terminar: el directorio de la base y el de cada sesión, con su caché —y, en
// ella, su grafo— y un directorio vacío del que reproducir.
const (
	prefijoDeLaBaseReconstruida   = "kitlegal-evals-medida-base-"
	prefijoDeLaSesionReconstruida = "kitlegal-evals-medida-sesion-"
	directorioSinGrabaciones      = "sin-grabaciones"
)

// ordenDelProcesoDelServidor es la orden con la que un informe publica el
// proceso del servidor MCP de una sesión del modo herramienta, que no se repite:
// no es una consulta y no dio ningún texto.
const ordenDelProcesoDelServidor = appletDelServidor + " " + verboDelServidor

// Lo que la reconstrucción sabe del informe de un sondeo y de sus preguntas
// (contracts/medida-y-casos.md §2 y §5 y data-model §4 de H25): el fichero de
// las preguntas, que está junto al informe; las dos marcas de una plantilla; la
// herramienta cuyas invocaciones dan texto y cómo empieza la orden que lo da; y
// lo que lleva la orden que coteja un documento.
const (
	ficheroDeLasPreguntas = "preguntas.json"

	marcaDelFragmento = "{fragmento}"
	marcaDeLaFicha    = "{ficha}"

	herramientaDeLasOrdenes = "Bash"
	principioDeUnaOrden     = programaDeLasConsultas + " "
	cotejoEnUnaOrden        = appletDeLaCita + " " + verboCotejar
)

// sesionDelInforme es lo que la reconstrucción lee de cada sesión de un informe
// versionado del job de evals (research V2 de H24): su nombre, el fichero de su
// eval, su respuesta y sus invocaciones. Un informe no lleva la pregunta.
type sesionDelInforme struct {
	Sesion       string                 `json:"sesion"`
	Eval         string                 `json:"eval"`
	Respuesta    string                 `json:"respuesta"`
	Invocaciones []invocacionDelInforme `json:"invocaciones"`
}

// invocacionDelInforme es lo que se lee de cada invocación de una sesión de un
// informe del job: su orden; si fue una llamada a una herramienta, que los
// informes anteriores a los dos modos no dicen; y su código.
type invocacionDelInforme struct {
	Orden   string `json:"orden"`
	Llamada bool   `json:"llamada"`

	// Codigo es el código con el que terminó en su sesión, o nil si quedó sin
	// ninguno. Con él se compara el de una orden del applet cita repetida, y
	// solo el suyo (research D7 de H25).
	Codigo *int `json:"codigo"`
}

// sesionDeUnSondeo es lo que la reconstrucción lee de cada sesión del informe
// de un sondeo (data-model §3 de H25): su nombre, el id de su pregunta entre
// las del sondeo, su respuesta y sus invocaciones.
type sesionDeUnSondeo struct {
	Sesion       string                 `json:"sesion"`
	Pregunta     string                 `json:"pregunta"`
	Respuesta    string                 `json:"respuesta"`
	Invocaciones []invocacionDeUnSondeo `json:"invocaciones"`
}

// invocacionDeUnSondeo es lo que se lee de cada invocación de una sesión de un
// sondeo: su herramienta, la orden de su entrada y lo que devolvió. La orden de
// Bash es un texto; la entrada de otra herramienta puede llevar en esa clave
// otra cosa, que no es una orden y se lee sin que el informe deje de serlo.
type invocacionDeUnSondeo struct {
	Herramienta string `json:"herramienta"`
	Entrada     struct {
		Command any `json:"command"`
	} `json:"entrada"`
	Salida string `json:"salida"`
}

// preguntasLeidas son las preguntas de un sondeo: las del fichero
// preguntas.json que está junto a su informe (data-model §4 de H25).
type preguntasLeidas struct {
	// ruta es la del fichero, desde la raíz. No es una clave suya.
	ruta string

	// Fragmento es la ruta, desde la raíz, del fichero cuyo contenido ponen las
	// plantillas en el lugar de sus marcas.
	Fragmento string `json:"fragmento"`

	// Preguntas son las del sondeo, cada una con el id que nombran sus
	// sesiones.
	Preguntas []preguntaLeida `json:"preguntas"`
}

// preguntaLeida es una pregunta de un sondeo: la de una eval de la skill o una
// plantilla.
type preguntaLeida struct {
	ID string `json:"id"`

	// Eval es el fichero de la eval de la skill cuya pregunta es la de ese id.
	Eval string `json:"eval"`

	// Pregunta es, si no lleva eval, la plantilla: marcaDelFragmento es el
	// contenido del fragmento, byte a byte, y marcaDeLaFicha, ese contenido
	// hasta su primera línea en blanco, con un salto de línea al final.
	Pregunta string `json:"pregunta"`
}

// informeVersionado son las sesiones de un informe versionado, por su nombre: las de
// un informe del job de evals o las de un informe de un sondeo, que además
// lleva sus preguntas (data-model §3 de H25).
type informeVersionado struct {
	// deUnSondeo dice si es el informe de un sondeo.
	deUnSondeo bool

	// delJob son las sesiones de un informe del job; nil en el de un sondeo.
	delJob map[string]sesionDelInforme

	// delSondeo son las sesiones de un informe de un sondeo, y preguntas, las
	// de ese sondeo; nil y sin ninguna en el del job.
	delSondeo map[string]sesionDeUnSondeo
	preguntas preguntasLeidas
}

// textoReconstruido es el texto de una invocación de una sesión, reconstruido:
// su orden, la del informe, y lo que escribió en la salida estándar al repetirla
// en proceso, con lo que hace falta para saber si es el de un bloque que un
// derivado quita. El de una invocación de un sondeo, que no se repite, lleva lo
// que devolvió en su sesión y ni argumentos ni código.
type textoReconstruido struct {
	Texto

	// argumentos son los de la invocación repetida, sin el nombre del programa:
	// el applet, el verbo y lo que les sigue.
	argumentos []string

	// codigo es el código con el que terminó.
	codigo int
}

// sesionReconstruida es lo que la resolución de un caso toma de su sesión: su
// pregunta, sin la parte del texto pegado que el caso quite; su respuesta; y
// los textos de sus invocaciones.
type sesionReconstruida struct {
	pregunta  string
	respuesta string
	textos    []textoReconstruido
}

// sesionDeUnInforme identifica lo reconstruido de una sesión de un informe
// versionado: la sesión y la parte del texto pegado en su pregunta que se le
// quita, o ninguna (research D10 de H25).
type sesionDeUnInforme struct {
	informe    string
	sesion     string
	sinElTexto string
}

// reconstructor resuelve los casos etiquetados del juez de una skill sin
// escribir nada a mano (contracts/medida-del-juez.md §4 y §5 de H24; research
// D17 de H24; FR-024, FR-051): de cada caso, la respuesta de su sesión en su
// informe versionado, la pregunta del fichero de eval que nombra esa sesión y
// los textos de sus herramientas, que reconstruye repitiendo en proceso sus
// invocaciones contra las grabaciones. Sin red, sin modelo y sin el binario
// instalado: es el arnés de la validación del ADR 0037 con app.Main en lugar
// del binario.
//
// Resuelve también los casos de una skill como jurisprudencia
// (contracts/medida-y-casos.md §2 a §7 de H25; FR-041 a FR-044): las órdenes
// del applet cita de un informe del job, que repite con su entrada estándar y
// compara con su código; los del informe de un sondeo, cuya pregunta sale de
// las preguntas del sondeo y cuyos textos son los de su informe, sin repetir
// nada; y los derivados que quitan una parte del texto pegado en la pregunta.
//
// Recuerda lo que ya ha reconstruido de cada sesión, de modo que resolver otra
// vez un caso suyo, u otro de la misma sesión que quite lo mismo de su texto
// pegado, no la repite: con un mismo reconstructor, cada sesión se reconstruye
// una sola vez por lo que se le quita. Se puede usar desde varias gorrutinas,
// que resuelven una detrás de otra.
type reconstructor struct {
	// raiz es el directorio del que cuelga el informe que nombra cada caso: la
	// raíz del repositorio.
	raiz string

	// evals es el directorio de las evals de hoy de la skill.
	evals string

	// retiradas es el directorio de las evals retiradas, cada una con su grafo
	// previo.
	retiradas string

	// ahora da el instante con el que se mide la edad de la base.
	ahora func() time.Time

	// candado deja pasar una sola resolución cada vez: guarda todo lo que sigue.
	candado sync.Mutex

	// deHoy son las evals de hoy por su fichero, y consultas, las que la base
	// tiene que servir; nil hasta que se leen.
	deHoy     map[string]Eval
	consultas []Consulta

	// informes son los informes ya leídos, por su ruta.
	informes map[string]informeVersionado

	// sesiones son las ya reconstruidas, cada una por lo que se le quita.
	sesiones map[sesionDeUnInforme]sesionReconstruida

	// base es el directorio de la caché que sirve a todas las sesiones, vacío si
	// no hay ninguna preparada, y preparada, cuándo se empezó a preparar.
	base      string
	preparada time.Time
}

// nuevoReconstructor es el reconstructor de los casos de la skill cuyas evals
// de hoy están en ese directorio, con lo demás donde lo tiene el repositorio,
// relativo al directorio de este paquete: los informes versionados, bajo su
// raíz; las evals retiradas, en EvalsRetiradas; las grabaciones, en
// UnionDeGrabaciones; y el grafo previo de cada eval de hoy, en GrafosPrevios.
// No lee ni prepara nada hasta que resuelve.
func nuevoReconstructor(evals string) *reconstructor {
	return &reconstructor{
		raiz:      raizDelRepositorio,
		evals:     evals,
		retiradas: EvalsRetiradas,
		ahora:     time.Now,
		informes:  map[string]informeVersionado{},
		sesiones:  map[sesionDeUnInforme]sesionReconstruida{},
	}
}

// resolver devuelve los casos, en su orden, resueltos: cada uno con la
// pregunta, la respuesta y los textos de su sesión y, si es un derivado, sin el
// texto del bloque que quita (sinElBloque) o sin la parte del texto pegado en
// su pregunta que quita (sinElTexto). El primer caso que no se puede resolver
// es el error, que lo nombra, y entonces no devuelve ninguno: su informe no se
// puede leer o no tiene su sesión, la eval de esa sesión no está ni entre las de
// hoy ni entre las retiradas, la pregunta de su sondeo no está, lo grabado no
// sirve la base o el grafo previo, una orden del applet cita repetida termina
// con otro código que el de su informe, el derivado no tiene ningún texto que
// quitar o el recorte de su pregunta no se puede hacer
// (contracts/medida-y-casos.md §7 de H25).
//
// La base y el directorio de cada sesión se retiran antes de volver: de lo
// reconstruido quedan los textos, que se recuerdan, y nada en el disco.
func (r *reconstructor) resolver(casos []CasoEtiquetado) (_ []CasoEtiquetado, err error) {
	r.candado.Lock()
	defer r.candado.Unlock()

	defer func() { err = errors.Join(err, r.retirarLaBase()) }()

	resueltos := make([]CasoEtiquetado, 0, len(casos))

	for _, caso := range casos {
		resuelto, err := r.resolverElCaso(caso)
		if err != nil {
			return nil, fmt.Errorf("el caso %s: %w", caso.nombre(), err)
		}

		resueltos = append(resueltos, resuelto)
	}

	return resueltos, nil
}

// resolverElCaso pone al caso la pregunta, la respuesta y los textos de su
// sesión; en un derivado por bloque, los textos sin el del bloque que quita, y
// en uno por texto, la pregunta y los textos de su sesión sin esa parte de su
// texto pegado, que ya vienen así de su reconstrucción. Un derivado quita una
// de las dos cosas: el que lleva las dos formas no se resuelve.
func (r *reconstructor) resolverElCaso(caso CasoEtiquetado) (CasoEtiquetado, error) {
	porBloque := caso.Quitado != nil && caso.Quitado.Texto == ""

	if quitado := caso.Quitado; quitado != nil && !porBloque && (quitado.Norma != "" || quitado.Bloque != "") {
		return CasoEtiquetado{}, fmt.Errorf("quitado lleva un bloque (%s %s) y un texto (%s): un derivado quita una de"+
			" las dos cosas", quitado.Norma, quitado.Bloque, quitado.Texto)
	}

	sesion, err := r.sesionDelCaso(caso)
	if err != nil {
		return CasoEtiquetado{}, err
	}

	caso.Pregunta, caso.Respuesta = sesion.pregunta, sesion.respuesta

	if porBloque {
		caso.Textos, err = sinElBloque(sesion.textos, *caso.Quitado)
		if err != nil {
			return CasoEtiquetado{}, err
		}

		return caso, nil
	}

	caso.Textos = make([]Texto, 0, len(sesion.textos))
	for _, texto := range sesion.textos {
		caso.Textos = append(caso.Textos, texto.Texto)
	}

	return caso, nil
}

// sesionDelCaso es la sesión del caso, reconstruida: la que ya se recordaba de
// su informe, su sesión y la parte del texto pegado que el caso quita o, si es
// la primera vez, la que se reconstruye según el tipo de su informe, que
// entonces se recuerda (research D10 de H25). Con otra parte quitada no vale la
// que se recuerda: cambian la pregunta y lo que recibe la orden que coteja.
func (r *reconstructor) sesionDelCaso(caso CasoEtiquetado) (sesionReconstruida, error) {
	clave := sesionDeUnInforme{informe: caso.Informe, sesion: caso.Sesion, sinElTexto: caso.textoQuitado()}
	if reconstruida, recordada := r.sesiones[clave]; recordada {
		return reconstruida, nil
	}

	informe, err := r.informeDelCaso(caso.Informe)
	if err != nil {
		return sesionReconstruida{}, err
	}

	reconstruir := r.reconstruirLaDelJob
	if informe.deUnSondeo {
		reconstruir = r.reconstruirLaDelSondeo
	}

	reconstruida, err := reconstruir(informe, clave)
	if err != nil {
		return sesionReconstruida{}, err
	}

	r.sesiones[clave] = reconstruida

	return reconstruida, nil
}

// informeSinLaSesion es el error del informe, del job o de un sondeo, que no
// tiene la sesión que nombra un caso.
const informeSinLaSesion = "el informe %s no tiene la sesión %s"

// reconstruirLaDelJob reconstruye esa sesión de un informe del job de evals: su
// pregunta es la de su eval, sin la parte del texto pegado que se le quita, y
// sus textos, los de repetir sus invocaciones con lo que queda de ese texto. Es
// un error que el informe no tenga la sesión.
func (r *reconstructor) reconstruirLaDelJob(informe informeVersionado, clave sesionDeUnInforme) (
	sesionReconstruida, error,
) {
	sesion, esta := informe.delJob[clave.sesion]
	if !esta {
		return sesionReconstruida{}, fmt.Errorf(informeSinLaSesion, clave.informe, clave.sesion)
	}

	eval, grafosPrevios, err := r.evalDeLaSesion(sesion.Eval)
	if err != nil {
		return sesionReconstruida{}, err
	}

	pregunta, queda, err := sinElTexto(eval.Pregunta, clave.sinElTexto)
	if err != nil {
		return sesionReconstruida{}, err
	}

	textos, err := r.textosDeLaSesion(sesion, eval, grafosPrevios, queda)
	if err != nil {
		return sesionReconstruida{}, err
	}

	return sesionReconstruida{pregunta: pregunta, respuesta: sesion.Respuesta, textos: textos}, nil
}

// reconstruirLaDelSondeo reconstruye esa sesión del informe de un sondeo
// (contracts/medida-y-casos.md §2 y §5 de H25; FR-042): su pregunta es la de su
// id entre las del sondeo, sin la parte del texto pegado que se le quita, y sus
// textos, los de su informe (textosDelSondeo). No repite ninguna orden ni
// prepara nada: no hay base ni directorio de la sesión. Es un error que el
// informe no tenga la sesión.
func (r *reconstructor) reconstruirLaDelSondeo(informe informeVersionado, clave sesionDeUnInforme) (
	sesionReconstruida, error,
) {
	sesion, esta := informe.delSondeo[clave.sesion]
	if !esta {
		return sesionReconstruida{}, fmt.Errorf(informeSinLaSesion, clave.informe, clave.sesion)
	}

	compuesta, err := r.preguntaDelSondeo(informe.preguntas, sesion.Pregunta)
	if err != nil {
		return sesionReconstruida{}, err
	}

	pregunta, queda, err := sinElTexto(compuesta, clave.sinElTexto)
	if err != nil {
		return sesionReconstruida{}, err
	}

	return sesionReconstruida{
		pregunta:  pregunta,
		respuesta: sesion.Respuesta,
		textos:    textosDelSondeo(sesion.Invocaciones, queda),
	}, nil
}

// informeDelCaso es el informe de esa ruta, que se lee la primera vez que un
// caso lo nombra.
func (r *reconstructor) informeDelCaso(ruta string) (informeVersionado, error) {
	if informe, leido := r.informes[ruta]; leido {
		return informe, nil
	}

	informe, err := leerInforme(r.raiz, ruta)
	if err != nil {
		return informeVersionado{}, err
	}

	r.informes[ruta] = informe

	return informe, nil
}

// informeQueNoEsDelJob es el error del informe que no es un documento JSON con
// la forma de un informe del job de evals; sigue con el de quien lo lee.
const informeQueNoEsDelJob = "el informe %s no es un informe del job de evals: %w"

// leerInforme lee el informe versionado de esa ruta, que va desde la raíz y con
// la barra de separador, y devuelve sus sesiones por su nombre: las de un
// informe del job de evals o, si lleva la clave sondeo, las de un informe de un
// sondeo, con las preguntas del fichero que está junto a él (research D8 de
// H25). El informe y sus preguntas solo se leen, y nunca de fuera de la raíz:
// una ruta que sale de ella no se puede leer. Es un error que no sea un
// documento JSON con la forma de su tipo de informe, o que las preguntas de un
// sondeo no se puedan leer.
func leerInforme(raiz, ruta string) (informeVersionado, error) {
	contenido, err := fs.ReadFile(os.DirFS(raiz), ruta)
	if err != nil {
		return informeVersionado{}, fmt.Errorf("el informe %s no se puede leer: %w", ruta, err)
	}

	var tipo struct {
		Sondeo jsontext.Value `json:"sondeo"`
	}

	if err := json.Unmarshal(contenido, &tipo); err != nil {
		return informeVersionado{}, fmt.Errorf(informeQueNoEsDelJob, ruta, err)
	}

	if tipo.Sondeo != nil {
		return leerInformeDeUnSondeo(raiz, ruta, contenido)
	}

	var leido struct {
		Evals []sesionDelInforme `json:"evals"`
	}

	if err := json.Unmarshal(contenido, &leido); err != nil {
		return informeVersionado{}, fmt.Errorf(informeQueNoEsDelJob, ruta, err)
	}

	sesiones := make(map[string]sesionDelInforme, len(leido.Evals))
	for _, sesion := range leido.Evals {
		sesiones[sesion.Sesion] = sesion
	}

	return informeVersionado{delJob: sesiones}, nil
}

// leerInformeDeUnSondeo lee del contenido del informe de un sondeo sus
// sesiones, y del fichero preguntas.json de su misma carpeta, sus preguntas
// (contracts/medida-y-casos.md §2 de H25). Es un error que el contenido no
// tenga la forma del informe de un sondeo o que sus preguntas no se puedan
// leer.
func leerInformeDeUnSondeo(raiz, ruta string, contenido []byte) (informeVersionado, error) {
	var leido struct {
		Sesiones []sesionDeUnSondeo `json:"sesiones"`
	}

	if err := json.Unmarshal(contenido, &leido); err != nil {
		return informeVersionado{}, fmt.Errorf("el informe %s no es un informe de un sondeo: %w", ruta, err)
	}

	preguntas, err := leerPreguntasDelSondeo(raiz, path.Join(path.Dir(ruta), ficheroDeLasPreguntas))
	if err != nil {
		return informeVersionado{}, err
	}

	sesiones := make(map[string]sesionDeUnSondeo, len(leido.Sesiones))
	for _, sesion := range leido.Sesiones {
		sesiones[sesion.Sesion] = sesion
	}

	return informeVersionado{deUnSondeo: true, delSondeo: sesiones, preguntas: preguntas}, nil
}

// leerPreguntasDelSondeo lee las preguntas de un sondeo del fichero de esa
// ruta, que va desde la raíz y con la barra de separador. Solo se leen, y nunca
// de fuera de la raíz. Es un error que el fichero no se pueda leer o que no sea
// un documento JSON con su forma.
func leerPreguntasDelSondeo(raiz, ruta string) (preguntasLeidas, error) {
	contenido, err := fs.ReadFile(os.DirFS(raiz), ruta)
	if err != nil {
		return preguntasLeidas{}, fmt.Errorf("las preguntas %s no se pueden leer: %w", ruta, err)
	}

	var preguntas preguntasLeidas
	if err := json.Unmarshal(contenido, &preguntas); err != nil {
		return preguntasLeidas{}, fmt.Errorf("las preguntas %s no son las de un sondeo: %w", ruta, err)
	}

	preguntas.ruta = ruta

	return preguntas, nil
}

// preguntaDelSondeo es la pregunta de ese id entre las de un sondeo
// (contracts/medida-y-casos.md §2 de H25; research D9): si lleva eval, la de
// esa eval de hoy de la skill; si lleva pregunta, esa plantilla con sus marcas
// sustituidas (preguntaCompuesta). Es un error que las preguntas no tengan ese
// id, que su eval no sea de las de hoy o que no lleve ni eval ni pregunta.
func (r *reconstructor) preguntaDelSondeo(preguntas preguntasLeidas, id string) (string, error) {
	posicion := slices.IndexFunc(preguntas.Preguntas, func(pregunta preguntaLeida) bool { return pregunta.ID == id })
	if posicion < 0 {
		return "", fmt.Errorf("las preguntas %s no tienen la pregunta %s", preguntas.ruta, id)
	}

	pregunta := preguntas.Preguntas[posicion]

	switch {
	case pregunta.Eval != "":
		if err := r.leerLasEvalsDeHoy(); err != nil {
			return "", err
		}

		eval, esDeHoy := r.deHoy[pregunta.Eval]
		if !esDeHoy {
			return "", fmt.Errorf("la eval %s de la pregunta %s no es de las de hoy de %s", pregunta.Eval, id, r.evals)
		}

		return eval.Pregunta, nil
	case pregunta.Pregunta != "":
		return r.preguntaCompuesta(preguntas, pregunta)
	default:
		return "", fmt.Errorf("la pregunta %s de %s no lleva eval ni pregunta", id, preguntas.ruta)
	}
}

// preguntaCompuesta es la plantilla de una pregunta de un sondeo con sus marcas
// sustituidas (contracts/medida-y-casos.md §2 de H25): marcaDelFragmento, por el
// contenido del fichero que nombran las preguntas, byte a byte, y
// marcaDeLaFicha, por ese contenido hasta su primera línea en blanco, con un
// salto de línea al final. El fragmento se lee desde la raíz, y solo si la
// plantilla lleva alguna de las dos. Es un error que entonces no se pueda leer
// o que, para la ficha, no tenga ninguna línea en blanco.
func (r *reconstructor) preguntaCompuesta(preguntas preguntasLeidas, pregunta preguntaLeida) (string, error) {
	llevaLaFicha := strings.Contains(pregunta.Pregunta, marcaDeLaFicha)
	if !llevaLaFicha && !strings.Contains(pregunta.Pregunta, marcaDelFragmento) {
		return pregunta.Pregunta, nil
	}

	contenido, err := fs.ReadFile(os.DirFS(r.raiz), preguntas.Fragmento)
	if err != nil {
		return "", fmt.Errorf("el fragmento %s de %s no se puede leer: %w", preguntas.Fragmento, preguntas.ruta, err)
	}

	fragmento, ficha := string(contenido), ""

	if llevaLaFicha {
		hastaLaLineaEnBlanco, _, laTiene := strings.Cut(fragmento, lineaEnBlanco)
		if !laTiene {
			return "", fmt.Errorf("el fragmento %s de %s no tiene ninguna línea en blanco: no hay ficha que poner en la"+
				" pregunta %s", preguntas.Fragmento, preguntas.ruta, pregunta.ID)
		}

		ficha = hastaLaLineaEnBlanco + "\n"
	}

	return strings.NewReplacer(marcaDelFragmento, fragmento, marcaDeLaFicha, ficha).Replace(pregunta.Pregunta), nil
}

// textosDelSondeo son los textos de una sesión de un sondeo, que no repiten
// ninguna orden (contracts/medida-y-casos.md §5 de H25; FR-042): de cada
// invocación de Bash cuya orden empieza por «kitlegal », en su orden, esa orden
// tal cual y lo que devolvió en su sesión, que es lo que dice su informe. Si
// del texto pegado en la pregunta no queda nada, las órdenes que contienen
// «cita cotejar» no dan texto: sin documento no hay nada que cotejar.
func textosDelSondeo(invocaciones []invocacionDeUnSondeo, queda string) []textoReconstruido {
	textos := make([]textoReconstruido, 0, len(invocaciones))

	for _, invocacion := range invocaciones {
		orden, esUnTexto := invocacion.Entrada.Command.(string)
		if invocacion.Herramienta != herramientaDeLasOrdenes || !esUnTexto || !strings.HasPrefix(orden, principioDeUnaOrden) {
			continue
		}

		if queda == "" && strings.Contains(orden, cotejoEnUnaOrden) {
			continue
		}

		textos = append(textos, textoReconstruido{Texto: Texto{Orden: orden, Salida: invocacion.Salida}})
	}

	return textos
}

// evalDeLaSesion es la eval del fichero que nombra una sesión, con el directorio
// en el que está el conjunto de su grafo previo: la de hoy de la skill, con
// GrafosPrevios, o, si la eval se retiró, la del directorio de las retiradas,
// con su subdirectorio grafosPreviosDeLasRetiradas (contracts/medida-del-juez.md
// §4 de H24).
func (r *reconstructor) evalDeLaSesion(fichero string) (Eval, string, error) {
	if err := r.leerLasEvalsDeHoy(); err != nil {
		return Eval{}, "", err
	}

	if eval, esDeHoy := r.deHoy[fichero]; esDeHoy {
		return eval, GrafosPrevios, nil
	}

	eval, err := leerEvalRetirada(r.retiradas, fichero)
	if err != nil {
		return Eval{}, "", fmt.Errorf("la eval %s no es de las de hoy de %s: %w", fichero, r.evals, err)
	}

	return eval, filepath.Join(r.retiradas, grafosPreviosDeLasRetiradas), nil
}

// leerLasEvalsDeHoy lee, la primera vez, las evals de hoy de la skill, que
// tienen que estar todas bien formadas, y con ellas las consultas que la base
// tiene que servir.
func (r *reconstructor) leerLasEvalsDeHoy() error {
	if r.deHoy != nil {
		return nil
	}

	evals, err := leerEvalsBienFormadas(r.evals)
	if err != nil {
		return err
	}

	deHoy := make(map[string]Eval, len(evals))
	for _, eval := range evals {
		deHoy[eval.Fichero] = eval
	}

	r.deHoy, r.consultas = deHoy, ConsultasNecesarias(evals)

	return nil
}

// evalRetirada es lo único que se lee del fichero de una eval retirada: su
// pregunta y su grafo previo.
type evalRetirada struct {
	Pregunta    string      `yaml:"pregunta"`
	GrafoPrevio GrafoPrevio `yaml:"grafo_previo"`
}

// leerEvalRetirada lee del directorio de las evals retiradas la del fichero de
// ese nombre, y de ella solo la pregunta y el grafo previo, con el lector común
// de YAML y sin validarla contra el formato de hoy, que no es el suyo
// (contracts/medida-del-juez.md §4 de H24). Es un error que el fichero no esté,
// o no se pueda leer, y que no tenga pregunta: sin ella no hay qué preguntar al
// juez.
func leerEvalRetirada(retiradas, fichero string) (Eval, error) {
	contenido, err := fs.ReadFile(os.DirFS(retiradas), fichero)
	if err != nil {
		return Eval{}, fmt.Errorf("la eval retirada %s de %s no se puede leer: %w", fichero, retiradas, err)
	}

	leida, err := skills.ValidarDocumentoYAML[evalRetirada](contenido, nil)
	if err != nil {
		return Eval{}, fmt.Errorf("la eval retirada %s de %s: %w", fichero, retiradas, err)
	}

	if leida.Pregunta == "" {
		return Eval{}, fmt.Errorf("la eval retirada %s de %s no tiene pregunta", fichero, retiradas)
	}

	return Eval{Fichero: fichero, Pregunta: leida.Pregunta, GrafoPrevio: leida.GrafoPrevio}, nil
}

// baseVigente devuelve el directorio de la base: la caché que Preparar llena
// con UnionDeGrabaciones y las consultas necesarias de las evals de hoy, y que
// sirve a todas las sesiones (contracts/medida-del-juez.md §5, paso 1, de H24).
// Si no hay ninguna, o la que hay tiene más de vigenciaDeLaBase desde que se
// empezó a preparar, la retira y prepara otra. Es un error que no se pueda
// preparar o que lo grabado no sirva alguna de esas consultas, y entonces no
// queda ninguna.
func (r *reconstructor) baseVigente() (string, error) {
	if r.base != "" && r.ahora().Sub(r.preparada) <= vigenciaDeLaBase {
		return r.base, nil
	}

	if err := r.retirarLaBase(); err != nil {
		return "", err
	}

	if err := r.leerLasEvalsDeHoy(); err != nil {
		return "", err
	}

	preparada := r.ahora()

	base, err := os.MkdirTemp("", prefijoDeLaBaseReconstruida)
	if err != nil {
		return "", fmt.Errorf("el directorio temporal de la base no se puede crear: %w", err)
	}

	faltas, err := Preparar(base, UnionDeGrabaciones(), r.consultas)
	if err == nil && len(faltas) > 0 {
		err = fmt.Errorf("la base no sirve las consultas de las evals de %s:\n%s", r.evals, lineasDeLasFaltas(faltas))
	}

	if err != nil {
		return "", errors.Join(err, retirarTemporal(base))
	}

	r.base, r.preparada = base, preparada

	return base, nil
}

// retirarLaBase retira el directorio de la base, si hay alguna.
func (r *reconstructor) retirarLaBase() error {
	if r.base == "" {
		return nil
	}

	base := r.base
	r.base = ""

	return retirarTemporal(base)
}

// lineasDeLasFaltas son las faltas, cada una en sus líneas.
func lineasDeLasFaltas(faltas []Falta) string {
	lineas := make([]string, 0, len(faltas))
	for _, falta := range faltas {
		lineas = append(lineas, falta.String())
	}

	return strings.Join(lineas, "\n")
}

// textosDeLaSesion reconstruye los textos de las herramientas de una sesión
// (contracts/medida-del-juez.md §5, pasos 2 y 3, de H24): en un directorio
// temporal nuevo, que retira antes de volver, prepara su caché
// (prepararLaCacheDeLaSesion) y repite sobre ella cada una de sus invocaciones
// (repetirLasInvocaciones), con lo que queda del texto pegado en su pregunta.
// El grafo de la sesión, que vive en esa caché, se descarta con ella.
func (r *reconstructor) textosDeLaSesion(sesion sesionDelInforme, eval Eval, grafosPrevios, queda string) (
	_ []textoReconstruido, err error,
) {
	base, err := r.baseVigente()
	if err != nil {
		return nil, err
	}

	dir, err := os.MkdirTemp("", prefijoDeLaSesionReconstruida)
	if err != nil {
		return nil, fmt.Errorf("el directorio temporal de la sesión no se puede crear: %w", err)
	}

	defer func() { err = errors.Join(err, retirarTemporal(dir)) }()

	dirCache := filepath.Join(dir, directorioDeLaCache)
	sinGrabaciones := filepath.Join(dir, directorioSinGrabaciones)

	for _, nuevo := range []string{dirCache, sinGrabaciones} {
		if err := os.Mkdir(nuevo, 0o700); err != nil {
			return nil, fmt.Errorf("el directorio de la sesión no se puede crear: %w", err)
		}
	}

	if err := prepararLaCacheDeLaSesion(dirCache, base, eval, grafosPrevios); err != nil {
		return nil, err
	}

	registro, err := registroDeLaSesion(sinGrabaciones, dirCache)
	if err != nil {
		return nil, err
	}

	return repetirLasInvocaciones(registro, sesion.Invocaciones, queda)
}

// prepararLaCacheDeLaSesion deja en dirCache, que existe y está vacío, lo que la
// sesión de esa eval tenía al empezar (contracts/medida-del-juez.md §5, paso 2,
// de H24): si la eval tiene grafo previo, el grafo del mundo que prepara
// prepararGrafoPrevio con el conjunto de grafosPrevios que nombra; y después,
// los ficheros de la base. Un comando del grafo previo que no termina bien es
// un error, con sus faltas.
func prepararLaCacheDeLaSesion(dirCache, base string, eval Eval, grafosPrevios string) error {
	if eval.GrafoPrevio.Grabaciones != "" {
		faltas, err := prepararGrafoPrevio(dirCache, UnionDeGrabaciones(), grafosPrevios, eval)
		if err != nil {
			return err
		}

		if len(faltas) > 0 {
			return fmt.Errorf("el grafo previo %s de la eval %s no se puede preparar:\n%s",
				eval.GrafoPrevio.Grabaciones, eval.Fichero, lineasDeLasFaltas(faltas))
		}
	}

	if err := copiarGrabaciones(base, dirCache); err != nil {
		return fmt.Errorf("la base no llega a la caché de la sesión: %w", err)
	}

	return nil
}

// registroDeLaSesion monta el registro con el que se repiten las invocaciones
// de una sesión (contracts/medida-del-juez.md §5, paso 3, de H24): el applet boe
// sobre la reproducción de un directorio vacío —una petición que se escapara
// fallaría en lugar de salir a la red— y la caché de la sesión; el applet graph,
// con el reloj del sistema, sobre el grafo del mundo de esa caché; el applet
// cita, que no tiene dependencias porque no pide nada a la red ni abre la caché
// ni el grafo; y la entrega a ese mismo grafo de lo que cada invocación
// observa. Son los únicos applets que invocaron las sesiones de los informes
// (research M6 de H24; research D4 de H25).
func registroDeLaSesion(sinGrabaciones, dirCache string) (*app.Registro, error) {
	registro, err := registroDeBoe(sinGrabaciones, cache.ConDirectorio(dirCache))
	if err != nil {
		return nil, err
	}

	delGrafo := app.DependenciasDelGrafoDelSistema()
	delGrafo.Almacen = []graph.Opcion{graph.ConDirectorio(dirCache)}

	for _, applet := range []app.Applet{app.AppletGrafo(delGrafo), app.AppletCita()} {
		if err := registro.Registrar(applet); err != nil {
			return nil, fmt.Errorf("el applet %s no se puede registrar: %w", applet.Nombre(), err)
		}
	}

	registro.EntregarAlGrafo(graph.Nuevo(graph.ConDirectorio(dirCache)))

	return registro, nil
}

// repetirLasInvocaciones repite en proceso, con app.Main y en su orden, cada
// invocación de una sesión de un informe del job salvo la del proceso del
// servidor, y devuelve un texto por cada una: su orden, la del informe tal
// cual, y lo que escribe en la salida estándar. Lo que escribe en la salida de
// error no es de ningún texto.
//
// La orden de boe o de graph da su texto termine con el código que termine
// (contracts/medida-del-juez.md §5, paso 3, de H24). La del applet cita tiene
// sus reglas (contracts/medida-y-casos.md §4 de H25; FR-041): se repite con sus
// argumentos (partirLaOrdenDeCita) y, si es cotejar sin --documento, con lo que
// queda del texto pegado en la pregunta por la entrada estándar; si no queda
// nada, la orden cotejar no se repite ni da texto, lleve o no --documento,
// porque sin documento no hay nada que cotejar; y es un error que termine con
// un código distinto del que el informe da a esa invocación.
func repetirLasInvocaciones(registro *app.Registro, invocaciones []invocacionDelInforme, queda string) (
	[]textoReconstruido, error,
) {
	textos := make([]textoReconstruido, 0, len(invocaciones))

	for _, invocacion := range invocaciones {
		if invocacion.Orden == ordenDelProcesoDelServidor {
			continue
		}

		if !esUnaOrdenDeCita(invocacion.Orden) {
			textos = append(textos, repetir(registro, invocacion.Orden, argumentosDeLaInvocacion(invocacion), nil))

			continue
		}

		orden := partirLaOrdenDeCita(invocacion.Orden)
		if orden.verbo() == verboCotejar && queda == "" {
			continue
		}

		texto := repetir(registro, invocacion.Orden, orden.argumentos(), orden.entrada(queda))
		if err := texto.conElCodigoDelInforme(invocacion.Codigo); err != nil {
			return nil, err
		}

		textos = append(textos, texto)
	}

	return textos, nil
}

// repetir repite en proceso, con app.Main, la invocación de esa orden con esos
// argumentos y esa entrada estándar, que es ninguna si es nil, y devuelve su
// texto: la orden tal cual, lo que escribe en la salida estándar y el código
// con el que termina. La entrada es la del registro mientras dura: la da el
// kernel al verbo que la lee, y a nadie más.
func repetir(registro *app.Registro, orden string, argumentos []string, entrada io.Reader) textoReconstruido {
	registro.LeerDe(entrada)

	var salida bytes.Buffer

	codigo := app.Main(slices.Concat([]string{programaDeLasConsultas}, argumentos), registro, &salida, io.Discard,
		sinDatosDeConstruccion, sinDatosDeConstruccion, sinDatosDeConstruccion)

	return textoReconstruido{Texto: Texto{Orden: orden, Salida: salida.String()}, argumentos: argumentos, codigo: codigo}
}

// conElCodigoDelInforme dice si el texto de una orden repetida termina con el
// código que su informe da a esa invocación (contracts/medida-y-casos.md §4 y
// §7 de H25; research D7): si termina con otro, o el informe no le da ninguno,
// es un error con la orden y los dos códigos. El error es una línea: de una
// orden de varias, que es la que lleva un documento, va la primera.
func (t textoReconstruido) conElCodigoDelInforme(delInforme *int) error {
	orden, _, hayMas := strings.Cut(t.Orden, "\n")
	if hayMas {
		orden += "…"
	}

	switch {
	case delInforme == nil:
		return fmt.Errorf("la orden «%s» termina con %d y el informe no le da ningún código", orden, t.codigo)
	case *delInforme != t.codigo:
		return fmt.Errorf("la orden «%s» termina con %d y el informe dice %d", orden, t.codigo, *delInforme)
	}

	return nil
}

// argumentosDeLaInvocacion son los argumentos con los que se repite una
// invocación de una sesión (contracts/medida-del-juez.md §5, paso 3, de H24):
// las palabras de su orden; en una llamada a una herramienta, cuya orden empieza
// por <applet>_<verbo>, el applet y el verbo por separado y, detrás de sus
// argumentos, --json, para que su texto sea el sobre, como el de una orden; y,
// detrás de todo, --offline, para que nada llegue a la red.
//
// Los de una orden del applet cita no salen de sus palabras, sino de su propia
// regla (partirLaOrdenDeCita), y no llevan --offline: el applet no pide nada a
// la red (research D5 de H25).
func argumentosDeLaInvocacion(invocacion invocacionDelInforme) []string {
	if esUnaOrdenDeCita(invocacion.Orden) {
		return partirLaOrdenDeCita(invocacion.Orden).argumentos()
	}

	palabras := strings.Fields(invocacion.Orden)

	if invocacion.Llamada && len(palabras) > 0 {
		applet, verbo, _ := strings.Cut(palabras[0], "_")
		palabras = slices.Concat([]string{applet, verbo}, palabras[1:], []string{"--json"})
	}

	return append(palabras, "--offline")
}

// Lo que la reconstrucción sabe de una orden del applet cita de un informe del
// job (contracts/medida-y-casos.md §4 de H25): cómo empieza, por orden y por
// llamada a su herramienta; los guiones de una bandera; la bandera que lleva
// el documento, cuyo valor va tal cual; y la que pide el sobre, que va una vez.
const (
	citaPorOrden   = appletDeLaCita + " "
	citaPorLlamada = appletDeLaCita + "_"

	guionesDeUnaBandera = "--"
	banderaDelDocumento = "documento"
	banderaDelSobre     = "json"
)

// principioDeUnaBandera es por donde se parte una orden del applet cita: un
// blanco seguido de -- y una letra minúscula. Sus blancos son los de
// unicode.IsSpace, como los que separan el nombre de una bandera de su valor.
var principioDeUnaBandera = regexp.MustCompile(`[\s\v\x{85}\p{Z}]--[a-z]`)

// esUnaOrdenDeCita dice si la orden de una invocación de un informe del job es
// del applet cita: empieza por «cita », la de una orden, o por «cita_», la de
// una llamada a una de sus herramientas.
func esUnaOrdenDeCita(orden string) bool {
	return strings.HasPrefix(orden, citaPorOrden) || strings.HasPrefix(orden, citaPorLlamada)
}

// ordenPartidaDeCita es una orden del applet cita de un informe del job, partida para
// repetirla (contracts/medida-y-casos.md §4 de H25).
type ordenPartidaDeCita struct {
	// cabeza son las palabras de delante de la primera bandera: el applet, el
	// verbo y, detrás, la referencia que va sin bandera.
	cabeza []string

	// banderas son las de la orden, en su orden, sin la que pide el sobre.
	banderas []banderaDeCita
}

// banderaDeCita es una bandera de una orden del applet cita: su nombre, sin los
// guiones, y su valor, si la orden le da alguno, aunque sea vacío.
type banderaDeCita struct {
	nombre   string
	valor    string
	conValor bool
}

// partirLaOrdenDeCita parte una orden del applet cita con la regla de la
// validación de su juez (contracts/medida-y-casos.md §4 de H25; research D5),
// que no es la de las demás órdenes: el valor de una bandera puede llevar
// espacios y, en --documento, saltos de línea. La orden, con «cita_» como
// «cita », se parte en cada blanco seguido de -- y una letra minúscula. Las
// palabras del primer trozo son el applet, el verbo y la referencia. Cada uno
// de los demás es una bandera: su nombre, hasta el primer = o blanco, y su
// valor, lo que sigue, sin blancos en los extremos salvo el de --documento, que
// va tal cual; la bandera sin = ni blanco no tiene valor. La que pide el sobre
// se quita: argumentos la pone una vez, al final.
func partirLaOrdenDeCita(orden string) ordenPartidaDeCita {
	if resto, esUnaLlamada := strings.CutPrefix(orden, citaPorLlamada); esUnaLlamada {
		orden = citaPorOrden + resto
	}

	cortes := principioDeUnaBandera.FindAllStringIndex(orden, -1)
	trozos := make([]string, 0, len(cortes)+1)
	desde := 0

	for _, corte := range cortes {
		trozos = append(trozos, orden[desde:corte[0]])

		// La letra con la que termina el corte es la primera del nombre.
		desde = corte[1] - 1
	}

	trozos = append(trozos, orden[desde:])

	partida := ordenPartidaDeCita{cabeza: strings.Fields(trozos[0]), banderas: make([]banderaDeCita, 0, len(cortes))}

	for _, trozo := range trozos[1:] {
		if bandera := banderaDelTrozo(trozo); bandera.nombre != banderaDelSobre {
			partida.banderas = append(partida.banderas, bandera)
		}
	}

	return partida
}

// banderaDelTrozo es la bandera de un trozo de una orden del applet cita, que
// empieza por su nombre: va hasta el primer = o blanco, y lo que sigue es su
// valor.
func banderaDelTrozo(trozo string) banderaDeCita {
	corte := strings.IndexFunc(trozo, func(letra rune) bool { return letra == '=' || unicode.IsSpace(letra) })
	if corte < 0 {
		return banderaDeCita{nombre: trozo}
	}

	_, ancho := utf8.DecodeRuneInString(trozo[corte:])
	bandera := banderaDeCita{nombre: trozo[:corte], valor: trozo[corte+ancho:], conValor: true}

	if bandera.nombre != banderaDelDocumento {
		bandera.valor = strings.TrimSpace(bandera.valor)
	}

	return bandera
}

// verbo es el de la orden: la palabra que sigue al applet, o ninguna.
func (o ordenPartidaDeCita) verbo() string {
	if len(o.cabeza) < 2 {
		return ""
	}

	return o.cabeza[1]
}

// argumentos son los argumentos con los que se repite la orden: el applet, el
// verbo y la referencia; cada bandera, con sus guiones, y detrás su valor, si
// lo tiene, entero; y --json, una sola vez y al final, para que su texto sea el
// sobre.
func (o ordenPartidaDeCita) argumentos() []string {
	argumentos := slices.Clone(o.cabeza)

	for _, bandera := range o.banderas {
		argumentos = append(argumentos, guionesDeUnaBandera+bandera.nombre)

		if bandera.conValor {
			argumentos = append(argumentos, bandera.valor)
		}
	}

	return append(argumentos, guionesDeUnaBandera+banderaDelSobre)
}

// entrada es la entrada estándar con la que se repite la orden: lo que queda
// del texto pegado en la pregunta, para cotejar sin --documento, que es quien
// la lee —lo que leyó en su sesión no está en el informe—, y ninguna para las
// demás.
func (o ordenPartidaDeCita) entrada(queda string) io.Reader {
	conDocumento := slices.ContainsFunc(o.banderas, func(bandera banderaDeCita) bool {
		return bandera.nombre == banderaDelDocumento
	})

	if o.verbo() != verboCotejar || conDocumento {
		return nil
	}

	return strings.NewReader(queda)
}

// sinElBloque devuelve los textos de una sesión sin el del bloque que quita un
// derivado (contracts/medida-del-juez.md §5, paso 4, de H24): fuera el texto de
// cada boe articulo de esa norma y ese bloque que terminó con 0; y del de cada
// boe articulos que lo pidió y terminó con 0, los elementos de data con ese
// bloque, con el sobre vuelto a escribir, o fuera el texto si no le queda
// ninguno (sobreSinElBloque). Los demás van tal cual y en su orden. Si no hay
// nada que quitar, no hay derivado: es un error.
func sinElBloque(textos []textoReconstruido, quitado Quitado) ([]Texto, error) {
	quedan := make([]Texto, 0, len(textos))
	quitados := 0

	for _, texto := range textos {
		if !texto.leyo(quitado) {
			quedan = append(quedan, texto.Texto)

			continue
		}

		if texto.argumentos[1] == verboArticulo {
			quitados++

			continue
		}

		salida, fuera, err := sobreSinElBloque(texto.Salida, quitado.Bloque)
		if err != nil {
			return nil, fmt.Errorf("la salida de «%s» no es el sobre de una lectura de bloques: %w", texto.Orden, err)
		}

		quitados += fuera

		if salida != "" {
			quedan = append(quedan, Texto{Orden: texto.Orden, Salida: salida})
		}
	}

	if quitados == 0 {
		return nil, fmt.Errorf("ninguna lectura de %s %s terminó con 0 en su sesión: no hay ningún texto que quitar",
			quitado.Norma, quitado.Bloque)
	}

	return quedan, nil
}

// leyo dice si el texto es el de una lectura de ese bloque que terminó con 0:
// el de boe articulo o boe articulos con esa norma y con ese bloque entre los
// que pide, que son los argumentos que no son banderas.
func (t textoReconstruido) leyo(bloque Quitado) bool {
	if t.codigo != 0 || len(t.argumentos) < 4 || t.argumentos[0] != appletDeLasNormas {
		return false
	}

	if t.argumentos[1] != verboArticulo && t.argumentos[1] != verboArticulos {
		return false
	}

	return t.argumentos[2] == bloque.Norma && slices.ContainsFunc(t.argumentos[3:], func(argumento string) bool {
		return argumento == bloque.Bloque && !strings.HasPrefix(argumento, "--")
	})
}

// sobreDeBloques es el sobre de una lectura de varios bloques: sus seis claves,
// en su orden, con cada valor como está escrito y data como la lista de sus
// elementos.
type sobreDeBloques struct {
	Ok            jsontext.Value   `json:"ok"`
	Fuente        jsontext.Value   `json:"fuente"`
	URL           jsontext.Value   `json:"url"`
	FechaConsulta jsontext.Value   `json:"fecha_consulta"`
	Hash          jsontext.Value   `json:"hash"`
	Data          []jsontext.Value `json:"data"`
}

// sobreSinElBloque vuelve a escribir el sobre de una lectura de varios bloques
// sin los elementos de data de ese bloque, y dice cuántos ha quitado. El sobre
// sale como lo escribe el binario —sus seis claves en su orden, en una línea y
// con su salto final— y lo que queda de él, byte a byte: ni los demás valores ni
// los elementos que quedan se vuelven a codificar. Si no queda ningún elemento,
// la salida es vacía: el texto entero es del bloque. Es un error que la salida
// no sea ese sobre.
func sobreSinElBloque(salida, bloque string) (string, int, error) {
	var sobre sobreDeBloques
	if err := json.Unmarshal([]byte(salida), &sobre, json.RejectUnknownMembers(true)); err != nil {
		return "", 0, err
	}

	quedan := make([]jsontext.Value, 0, len(sobre.Data))

	for _, elemento := range sobre.Data {
		var leido struct {
			Bloque string `json:"bloque"`
		}

		if err := json.Unmarshal(elemento, &leido); err != nil {
			return "", 0, err
		}

		if leido.Bloque != bloque {
			quedan = append(quedan, elemento)
		}
	}

	fuera := len(sobre.Data) - len(quedan)
	if len(quedan) == 0 {
		return "", fuera, nil
	}

	sobre.Data = quedan

	escrito, err := json.Marshal(sobre, jsontext.PreserveRawStrings(true))
	if err != nil {
		return "", 0, err
	}

	return string(escrito) + "\n", fuera, nil
}

// Las tres partes del texto pegado en una pregunta que un derivado puede
// quitar, que son los valores de quitado.texto (contracts/medida-y-casos.md §3
// de H25): todo el texto, lo que va desde su fallo y el apartado 2.º de su
// fallo.
const (
	quitarElDocumento = "documento"
	quitarElFallo     = "fallo"
	quitarElApartado2 = "apartado-2"
)

// Lo que la reconstrucción sabe del texto que una persona pega en su pregunta,
// el de un documento del CENDOJ (contracts/medida-y-casos.md §3 de H25): que va
// detrás de una línea en blanco y empieza por su ficha, cuál es la línea que
// abre su fallo y cómo empieza el párrafo de su apartado 2.º.
const (
	lineaEnBlanco           = "\n\n"
	principioDelTextoPegado = lineaEnBlanco + "Roj:"
	lineaDelFallo           = "F A L L O"
	principioDelApartado2   = "2.º-"
)

// textoPegado parte una pregunta en su entrada y el texto que lleva pegado
// (contracts/medida-y-casos.md §3 y data-model §5 de H25): si contiene una
// línea en blanco seguida de «Roj:», se parte por su primera línea en blanco,
// con la entrada delante y el texto pegado detrás; si no, no lleva ninguno, y
// la entrada es la pregunta entera.
func textoPegado(pregunta string) (entrada, pegado string) {
	if !strings.Contains(pregunta, principioDelTextoPegado) {
		return pregunta, ""
	}

	entrada, pegado, _ = strings.Cut(pregunta, lineaEnBlanco)

	return entrada, pegado
}

// sinElTexto devuelve la pregunta de un caso que quita esa parte del texto
// pegado en la pregunta de su sesión, y lo que queda de ese texto
// (contracts/medida-y-casos.md §3 de H25; FR-043): la pregunta es la entrada y,
// si queda algo, una línea en blanco y lo que queda. Sin parte que quitar, la
// pregunta va tal cual y queda el texto pegado entero, o nada si no lleva
// ninguno.
//
// Es un error, y entonces no hay derivado, que la pregunta no lleve texto
// pegado o que de él no se pueda quitar esa parte (loQueQueda).
func sinElTexto(pregunta, quitado string) (delCaso, queda string, err error) {
	entrada, pegado := textoPegado(pregunta)

	if quitado == "" {
		return pregunta, pegado, nil
	}

	if pegado == "" {
		return "", "", fmt.Errorf("la pregunta no lleva ningún texto pegado: no se le puede quitar %s", quitado)
	}

	queda, err = loQueQueda(pegado, quitado)
	if err != nil {
		return "", "", err
	}

	if queda == "" {
		return entrada, "", nil
	}

	return entrada + lineaEnBlanco + queda, queda, nil
}

// loQueQueda es lo que queda del texto pegado en una pregunta sin esa parte
// (contracts/medida-y-casos.md §3 de H25): sin el documento, nada; sin el
// fallo, lo anterior a la línea «F A L L O», sin sus saltos de línea finales y
// con uno; y sin su apartado 2.º, el texto sin el párrafo que empieza por
// «2.º-» ni la línea en blanco que lo sigue. Es un error que la parte no sea
// ninguna de las tres, que el texto no tenga la línea del fallo —tampoco para
// quitarle su apartado 2.º, que es de su fallo— o que no tenga ese párrafo.
func loQueQueda(pegado, quitado string) (string, error) {
	switch quitado {
	case quitarElDocumento:
		return "", nil
	case quitarElFallo, quitarElApartado2:
	default:
		return "", fmt.Errorf("quitado.texto es %q y tiene que ser %s, %s o %s",
			quitado, quitarElDocumento, quitarElFallo, quitarElApartado2)
	}

	delFallo, _, laTiene := primeraLinea(pegado, func(linea, _ string) bool {
		return strings.TrimSuffix(linea, "\n") == lineaDelFallo
	})
	if !laTiene {
		return "", fmt.Errorf("el texto pegado no tiene la línea «%s»: no se le puede quitar %s", lineaDelFallo, quitado)
	}

	if quitado == quitarElFallo {
		return strings.TrimRight(pegado[:delFallo], "\n") + "\n", nil
	}

	// El párrafo es una línea, la que empieza por su principio, con una línea en
	// blanco detrás, que se va con él.
	desde, hasta, loTiene := primeraLinea(pegado, func(linea, detras string) bool {
		return strings.HasPrefix(linea, principioDelApartado2) && strings.HasSuffix(linea, "\n") &&
			strings.HasPrefix(detras, "\n")
	})
	if !loTiene {
		return "", fmt.Errorf("el texto pegado no tiene ningún párrafo que empiece por «%s» con una línea en blanco"+
			" detrás", principioDelApartado2)
	}

	return pegado[:desde] + pegado[hasta+len("\n"):], nil
}

// primeraLinea busca en el texto su primera línea que cumple la condición, y
// devuelve dónde empieza, dónde termina —detrás de su salto de línea, si lo
// tiene— y si hay alguna. A la condición le llegan la línea, con su salto de
// línea, y lo que la sigue en el texto.
func primeraLinea(texto string, cumple func(linea, detras string) bool) (desde, hasta int, hay bool) {
	for linea := range strings.Lines(texto) {
		hasta = desde + len(linea)
		if cumple(linea, texto[hasta:]) {
			return desde, hasta, true
		}

		desde = hasta
	}

	return 0, 0, false
}

// MedicionDelJuez es lo que medirAlJuez recibe de quien lanza la medida del
// juez de una skill (contracts/medida-del-juez.md §7 de H24; FR-050 a FR-054).
// No lleva a quien abre las sesiones de evals ni un directorio de sesiones: la
// ejecución de la medida no abre ninguna (FR-051).
type MedicionDelJuez struct {
	// Skill es la skill cuyo juez se mide: el campo skill de la medida.
	Skill string

	// Juez es el juez de la skill, el que LeerConjunto lee de su carpeta de
	// evals: sus clases, su rúbrica, el esquema con el que se valida cada voto y
	// sus casos. Su medida versionada no se usa.
	Juez *Juez

	// Votar es quien da cada voto. Se le llama desde varias gorrutinas a la vez,
	// una por caso que se vota.
	Votar Votante

	// ModeloDelJuez y VersionDelJuez son el id del modelo del juez y la versión
	// de Claude Code de sus votos: van a la medida tal cual.
	ModeloDelJuez  string
	VersionDelJuez string

	// Concurrencia es cuántos casos se votan a la vez como mucho, al menos 1; los
	// votos de un mismo caso van uno detrás de otro.
	Concurrencia int

	// Commit es el commit sobre el que se ejecuta la medida: va en su origen.
	Commit string

	// Fecha es cuándo se ejecuta: la medida lleva su día, el de su huso.
	Fecha time.Time
}

// medidaDada es la medida que da medirAlJuez, con sus claves JSON en el orden
// de contracts/medida-del-juez.md §7 de H24 (FR-052): las diez que
// leerMedidaDelJuez lee de una medida versionada y, con ellas, la skill, el
// fichero de cada huella y el origen. No lleva votos: los de una ejecución no
// se guardan.
type medidaDada struct {
	Skill               string              `json:"skill"`
	Clase               string              `json:"clase"`
	Fecha               string              `json:"fecha"`
	ModeloDelJuez       string              `json:"modelo_del_juez"`
	VersionDeClaudeCode string              `json:"version_de_claude_code"`
	Rubrica             ficheroMedido       `json:"rubrica"`
	Casos               ficheroMedido       `json:"casos"`
	Defectos            DefectosDeLaMedida  `json:"defectos"`
	Correctos           CorrectosDeLaMedida `json:"correctos"`
	Origen              string              `json:"origen"`
}

// ficheroMedido es un fichero de la carpeta del juez con el que se midió: su
// nombre en ella y la huella SHA-256, en hexadecimal, de lo que se leyó de él.
type ficheroMedido struct {
	Fichero string `json:"fichero"`
	SHA256  string `json:"sha256"`
}

// Lo que medirAlJuez dice en su medida y en su error
// (contracts/medida-del-juez.md §7 de H24): el origen de la medida, que sigue
// con el commit; la línea del caso que no da lo que dice su etiqueta, con su
// nombre, su etiqueta y cómo queda, y que sigue con sus frases si las tiene; y
// la del caso sin juzgar, con su nombre y el motivo del voto que no llegó.
const (
	origenDeLaMedidaDada = "ejecución de la medida del juez del job de evals sobre %s"

	casoSinLoDeSuEtiqueta = "%s: etiquetado %s y %s"
	casoMarcado           = "marcado"
	casoSinMarcar         = "sin marcar"

	casoSinJuzgar = "%s: sin juzgar: %s"
)

// medirAlJuez ejecuta la medida del juez de una skill
// (contracts/medida-del-juez.md §7 de H24; FR-050 a FR-054): vota sus casos
// etiquetados y devuelve el texto de la medida de lo que hay, lista para que
// una persona la versione.
//
//  1. No comprueba la medida versionada ni la lee para decidir nada: vota
//     corresponda o no a lo que hay (FR-043, FR-051).
//  2. Lee los casos del juez, que tienen que ser de una clase suya que decide,
//     y los resuelve con el reconstructor dado —el de la skill
//     (nuevoReconstructor) en su punto de entrada, y el que comparten los
//     tests—, sin preparar ni abrir ninguna sesión de evals.
//  3. Vota cada caso con la regla de los votos (votacion.juzgar), como mucho
//     Concurrencia casos a la vez: tres votos por cada defecto que se marca y
//     uno por cada correcto que no.
//  4. Devuelve la medida (medidaDada) con dos espacios de sangría y su salto
//     final. Sus cuatro claves son las de lo que hay —las huellas de la rúbrica
//     del juez y de sus casos, y el modelo y la versión recibidos—, no las de
//     la medida versionada (FR-052).
//  5. Si un defecto no queda marcado o un correcto queda marcado, devuelve
//     además, con la medida y sus recuentos, un error con una línea por caso,
//     en el orden de los casos: «<informe> <sesión> [sin <norma> <bloque>]:
//     etiquetado <etiqueta> y <marcado | sin marcar>» y, si alguno de sus votos
//     dijo sí con su frase, «: «<frase>» · …» (frasesDeLosSies).
//  6. Si algún caso queda sin juzgar, porque un voto suyo no llegó a darse,
//     devuelve solo un error, con una línea por cada uno de esos casos y su
//     motivo: una medida con casos sin juzgar no es una medida (FR-053).
//
// Es un error, sin ningún voto ni ninguna medida, que la medición no tenga
// juez, votante o al menos un caso a la vez, que el esquema del juez no sirva
// para validar sus votos, que los casos no tengan su forma o no sean de una
// clase del juez que decide, o que alguno no se pueda resolver.
//
// No escribe en la salida estándar ni en el repositorio (FR-054): el texto lo
// escribe quien la llama, y lo que la reconstrucción deja en el directorio
// temporal lo retira antes de volver.
func medirAlJuez(deLosCasos *reconstructor, medicion MedicionDelJuez) (string, error) {
	votacion, err := medicion.prepararLaVotacion()
	if err != nil {
		return "", err
	}

	aMedir, err := medicion.leerLosCasos(deLosCasos)
	if err != nil {
		return "", err
	}

	recuento := aMedir.contar(votacion.juzgarTodas(aMedir.aJuzgar(), medicion.Concurrencia))
	if len(recuento.sinJuzgar) > 0 {
		return "", errors.New(strings.Join(recuento.sinJuzgar, "\n"))
	}

	texto, err := medicion.textoDeLaMedida(aMedir, recuento)
	if err != nil {
		return "", err
	}

	if len(recuento.sinLoDeSuEtiqueta) > 0 {
		return texto, errors.New(strings.Join(recuento.sinLoDeSuEtiqueta, "\n"))
	}

	return texto, nil
}

// prepararLaVotacion prepara la votación de los casos con el juez y el votante
// de la medición. Es un error que no tenga juez o votante, que los casos que se
// votan a la vez sean menos de uno o que el esquema de la respuesta del juez no
// sirva para validar sus votos.
func (m MedicionDelJuez) prepararLaVotacion() (*votacion, error) {
	switch {
	case m.Juez == nil:
		return nil, fmt.Errorf("la skill %s no tiene juez", m.Skill)
	case m.Votar == nil:
		return nil, fmt.Errorf("el juez de la skill %s no tiene votante", m.Skill)
	case m.Concurrencia < 1:
		return nil, fmt.Errorf("los casos que se votan a la vez son %d y tienen que ser al menos 1", m.Concurrencia)
	}

	return nuevaVotacion(m.Juez, m.Votar)
}

// casosAMedir son los casos etiquetados con los que se mide al juez en una
// ejecución de la medida: leídos, resueltos y con lo que la medida dice de
// ellos.
type casosAMedir struct {
	// clase es la clase de los casos, y deLaClase, su posición entre las del
	// juez, que es la de su juicio en el de cada caso.
	clase     string
	deLaClase int

	// resueltos son los casos, en el orden del fichero, cada uno con su
	// pregunta, su respuesta y sus textos.
	resueltos []CasoEtiquetado
}

// casosMalFormados es el error de los casos etiquetados del juez que no tienen
// su forma; sigue con el de leerCasosEtiquetados.
const casosMalFormados = "los casos etiquetados del juez (" + carpetaDelJuez + "/" + ficheroDeCasosDelJuez + "): %w"

// leerLosCasos lee los casos etiquetados del juez de la medición y los resuelve
// con ese reconstructor. Es un error que no tengan su forma, que no sean de una
// clase del juez que decide —la medida es de una clase que decide (research D19
// de H24), y con otra no habría qué contar— o que alguno no se pueda resolver.
func (m MedicionDelJuez) leerLosCasos(deLosCasos *reconstructor) (casosAMedir, error) {
	leidos, err := leerCasosEtiquetados([]byte(m.Juez.Casos))
	if err != nil {
		return casosAMedir{}, fmt.Errorf(casosMalFormados, err)
	}

	deLaClase := slices.IndexFunc(m.Juez.Clases, func(clase ClaseDelJuez) bool {
		return clase.Decide && clase.Nombre == leidos.Clase
	})
	if deLaClase < 0 {
		return casosAMedir{}, fmt.Errorf("los casos son de la clase %s, que no es una clase del juez que decide", leidos.Clase)
	}

	resueltos, err := deLosCasos.resolver(leidos.Casos)
	if err != nil {
		return casosAMedir{}, err
	}

	return casosAMedir{clase: leidos.Clase, deLaClase: deLaClase, resueltos: resueltos}, nil
}

// aJuzgar es lo que se da al juez de cada caso, en su orden: su pregunta, su
// respuesta y sus textos, que es lo que lleva el mensaje de cada uno de sus
// votos.
func (c casosAMedir) aJuzgar() []respuestaAJuzgar {
	aJuzgar := make([]respuestaAJuzgar, 0, len(c.resueltos))

	for _, caso := range c.resueltos {
		aJuzgar = append(aJuzgar, respuestaAJuzgar{pregunta: caso.Pregunta, respuesta: caso.Respuesta, textos: caso.Textos})
	}

	return aJuzgar
}

// recuentoDeLaMedida es lo que una ejecución de la medida cuenta de sus casos
// votados (data-model §4 de H24).
type recuentoDeLaMedida struct {
	// defectos y correctos son los casos de cada etiqueta, con los que no dan lo
	// que dice.
	defectos  DefectosDeLaMedida
	correctos CorrectosDeLaMedida

	// sinLoDeSuEtiqueta son las líneas de los defectos que no quedan marcados y
	// de los correctos que quedan marcados, en el orden de los casos.
	sinLoDeSuEtiqueta []string

	// sinJuzgar son las líneas de los casos de los que un voto no llegó a darse,
	// en el orden de los casos: con alguno, no hay medida.
	sinJuzgar []string
}

// contar cuenta los casos con sus juicios, que llegan en su mismo orden. Un
// caso sin juzgar solo da su línea, con el motivo del voto que no llegó; de los
// demás, cada uno cuenta en su etiqueta con lo que el juez deja de su clase.
func (c casosAMedir) contar(juicios []JuicioDeRespuesta) recuentoDeLaMedida {
	var recuento recuentoDeLaMedida

	for posicion, caso := range c.resueltos {
		juicio := juicios[posicion]

		switch {
		case juicio.SinJuzgar != "":
			recuento.sinJuzgar = append(recuento.sinJuzgar, fmt.Sprintf(casoSinJuzgar, caso.nombre(), juicio.SinJuzgar))
		case caso.Etiqueta == etiquetaDefecto:
			recuento.contarElDefecto(caso, juicio.Clases[c.deLaClase])
		default:
			recuento.contarElCorrecto(caso, juicio.Clases[c.deLaClase])
		}
	}

	return recuento
}

// contarElDefecto cuenta un caso etiquetado como defecto: si su clase no queda
// marcada, cuenta además en SinMarcar y da su línea.
func (r *recuentoDeLaMedida) contarElDefecto(caso CasoEtiquetado, deLaClase JuicioDeClase) {
	r.defectos.Casos++

	if !deLaClase.Marcada {
		r.defectos.SinMarcar++
		r.sinLoDeSuEtiqueta = append(r.sinLoDeSuEtiqueta, lineaDelCaso(caso, casoSinMarcar, deLaClase))
	}
}

// contarElCorrecto cuenta un caso etiquetado como correcto: si su clase queda
// marcada, cuenta además en Marcados y da su línea.
func (r *recuentoDeLaMedida) contarElCorrecto(caso CasoEtiquetado, deLaClase JuicioDeClase) {
	r.correctos.Casos++

	if deLaClase.Marcada {
		r.correctos.Marcados++
		r.sinLoDeSuEtiqueta = append(r.sinLoDeSuEtiqueta, lineaDelCaso(caso, casoMarcado, deLaClase))
	}
}

// lineaDelCaso es la línea del error de medirAlJuez de un caso que no da lo que
// dice su etiqueta: su nombre, su etiqueta, cómo queda y, si las tiene, las
// frases de los votos de su clase que dicen sí.
func lineaDelCaso(caso CasoEtiquetado, queda string, deLaClase JuicioDeClase) string {
	linea := fmt.Sprintf(casoSinLoDeSuEtiqueta, caso.nombre(), caso.Etiqueta, queda)

	frases := frasesDeLosSies(deLaClase)
	if len(frases) == 0 {
		return linea
	}

	return linea + ": " + strings.Join(frases, separadorDeLasFrases)
}

// frasesDeLosSies son, entre comillas y en su orden, las frases de los votos
// de la clase que cuentan y dicen sí con su frase en la respuesta: las tres que
// marcan un caso marcado, y las dos, la una o ninguna del que no llega a
// estarlo. De un voto nulo y su repetición cuenta la repetición, que va detrás
// de él con su mismo número (votosDelNumero); y el sí de un voto que cita una
// frase que no está en la respuesta no cuenta como sí, ni su frase como una de
// las del caso.
func frasesDeLosSies(deLaClase JuicioDeClase) []string {
	frases := make([]string, 0, votosParaMarcar)

	for posicion, voto := range deLaClase.Votos {
		seRepite := posicion+1 < len(deLaClase.Votos) && deLaClase.Votos[posicion+1].Voto == voto.Voto
		if !seRepite && voto.diceSi() {
			frases = append(frases, comillaQueAbre+voto.Frase+comillaQueCierra)
		}
	}

	return frases
}

// textoDeLaMedida es el texto de la medida de esa ejecución
// (contracts/medida-del-juez.md §7 de H24; FR-052): la skill, la clase de los
// casos, el día de la fecha, el modelo y la versión recibidos, las huellas de
// la rúbrica del juez, que es la que va como sus instrucciones, y de sus casos,
// que son los votados, los dos recuentos y el origen, con el commit. Va con dos
// espacios de sangría y su salto final, como una medida versionada. Es un error
// que alguno de sus textos no sea UTF-8: una medida que no se puede escribir
// tal cual no se da.
func (m MedicionDelJuez) textoDeLaMedida(casos casosAMedir, recuento recuentoDeLaMedida) (string, error) {
	medida := medidaDada{
		Skill:               m.Skill,
		Clase:               casos.clase,
		Fecha:               m.Fecha.Format(time.DateOnly),
		ModeloDelJuez:       m.ModeloDelJuez,
		VersionDeClaudeCode: m.VersionDelJuez,
		Rubrica:             ficheroMedido{Fichero: ficheroDeRubricaDelJuez, SHA256: huellaSHA256([]byte(m.Juez.Rubrica))},
		Casos:               ficheroMedido{Fichero: ficheroDeCasosDelJuez, SHA256: huellaSHA256([]byte(m.Juez.Casos))},
		Defectos:            recuento.defectos,
		Correctos:           recuento.correctos,
		Origen:              fmt.Sprintf(origenDeLaMedidaDada, m.Commit),
	}

	escrita, err := json.Marshal(medida, jsontext.WithIndent("  "))
	if err != nil {
		return "", fmt.Errorf("la medida no se puede escribir: %w", err)
	}

	return string(escrita) + "\n", nil
}
