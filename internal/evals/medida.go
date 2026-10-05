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
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

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
// juez.
type DefectosDeLaMedida struct {
	// Casos es cuántos son.
	Casos int

	// SinMarcar es cuántos de ellos no quedaron marcados: con alguno, la medida
	// no se cumple.
	SinMarcar int
}

// CorrectosDeLaMedida son los casos etiquetados como correctos de una medida del
// juez.
type CorrectosDeLaMedida struct {
	// Casos es cuántos son.
	Casos int

	// Marcados es cuántos de ellos quedaron marcados: con alguno, la medida no
	// se cumple.
	Marcados int
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

// leerMedidaDelJuez lee la medida del juez de esa ruta
// (contracts/medida-del-juez.md §1 de H24). No tiene esquema publicado: se lee
// como un objeto JSON del que solo se miran las diez claves de MedidaDelJuez,
// cada una con su tipo, un texto o un entero; una clave repetida es un error,
// nunca la última que gana. Es un error que no se pueda leer, que no sea ese
// documento o que le falte alguna de las diez, y una clave a null es una clave
// que falta: el error las nombra todas, en el orden del contrato, con las de un
// objeto detrás de la suya y de un punto, como rubrica.sha256. Ningún error
// lleva delante de qué medida habla: lo pone quien la lee.
func leerMedidaDelJuez(ruta string) (MedidaDelJuez, error) {
	contenido, err := leerFichero(ruta)
	if err != nil {
		return MedidaDelJuez{}, fmt.Errorf("no se puede leer: %w", err)
	}

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
// H24, carácter a carácter, y las dos de lo que no se puede leer. Cada fichero
// lleva el nombre que tiene dentro de la carpeta de evals de la skill,
// juez/<fichero>, como en los errores del conjunto.
const (
	// medidaQueNoCorresponde es la de la medida que no se puede leer o a la que
	// le falta alguna de sus claves; sigue con el error de leerMedidaDelJuez.
	medidaQueNoCorresponde = "la medida versionada (" + carpetaDelJuez + "/" + ficheroDeMedidaDelJuez + ") no corresponde: %v"

	rubricaDeOtraMedida = "la rúbrica (" + carpetaDelJuez + "/" + ficheroDeRubricaDelJuez + ") no es la de la medida versionada"
	casosDeOtraMedida   = "los casos (" + carpetaDelJuez + "/" + ficheroDeCasosDelJuez + ") no son los de la medida versionada"

	// casosQueNoSeLeen es la de los casos que han dejado de poder leerse desde
	// que se leyó el juez; sigue con el error.
	casosQueNoSeLeen = "los casos (" + carpetaDelJuez + "/" + ficheroDeCasosDelJuez + ") no se pueden leer: %v"

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
// últimas, las de una que no se cumple. La medida que no se puede leer, o sin
// alguna de las claves que se leen, tampoco corresponde, y da una sola línea,
// con el error de leerMedidaDelJuez: sin la medida entera no hay con qué
// comparar lo demás.
//
// El modelo y la versión fijados los da quien la llama, y el juez es el de la
// skill, nunca nil. Lee los casos y la medida de sus rutas, y nada más: no usa
// ningún modelo ni abre ningún proceso.
func comprobarLaMedida(juez *Juez, modeloFijado, versionFijada string) []string {
	medida, err := leerMedidaDelJuez(juez.Medida)
	if err != nil {
		return []string{fmt.Sprintf(medidaQueNoCorresponde, err)}
	}

	lineas := lineasDeLasHuellas(juez, medida)
	lineas = append(lineas, lineasDeLoFijado(medida, modeloFijado, versionFijada)...)
	lineas = append(lineas, lineasDeLasClases(juez.Clases, medida.Clase)...)

	return append(lineas, lineasDeLosRecuentos(medida)...)
}

// lineasDeLasHuellas compara las huellas SHA-256 de la rúbrica del juez, que es
// la que va tal cual como sus instrucciones, y de sus casos, leídos de su ruta,
// con las de la medida. Unos casos que no se pueden leer no tienen huella que
// comparar: dan su línea, con el error.
func lineasDeLasHuellas(juez *Juez, medida MedidaDelJuez) []string {
	var lineas []string

	if huellaSHA256([]byte(juez.Rubrica)) != medida.HuellaDeLaRubrica {
		lineas = append(lineas, rubricaDeOtraMedida)
	}

	casos, err := leerFichero(juez.Casos)

	switch {
	case err != nil:
		lineas = append(lineas, fmt.Sprintf(casosQueNoSeLeen, err))
	case huellaSHA256(casos) != medida.HuellaDeLosCasos:
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
// respuesta de una sesión de un informe versionado del job de evals, con la
// etiqueta de lo que el juez tiene que decir de ella (data-model §4 y
// contracts/medida-del-juez.md §4 de H24; FR-024). Nada suyo está escrito a
// mano: el fichero dice de dónde sale, y resolverlo (reconstructor.resolver)
// le pone la pregunta, la respuesta y los textos.
type CasoEtiquetado struct {
	// Informe es la ruta del informe versionado del que sale, desde la raíz del
	// repositorio y con la barra de separador.
	Informe string `yaml:"informe"`

	// Sesion es el nombre de la sesión de ese informe cuya respuesta es la del
	// caso.
	Sesion string `yaml:"sesion"`

	// Quitado es, en un derivado, el bloque cuyo texto se quita de los de la
	// sesión, que es lo que hace de su respuesta un defecto por construcción;
	// nil en los demás.
	Quitado *BloqueQuitado `yaml:"quitado"`

	// Grupo es el de la validación en la que se usó el caso, ajuste o medida.
	Grupo string `yaml:"grupo"`

	// Etiqueta es etiquetaDefecto o etiquetaCorrecto.
	Etiqueta string `yaml:"etiqueta"`

	// Procedencia dice de dónde sale la etiqueta: bitacora, derivado o lectura.
	Procedencia string `yaml:"procedencia"`

	// Frase es, en un caso de la lectura, la frase de la respuesta que decide su
	// etiqueta; vacía en los demás.
	Frase string `yaml:"frase"`

	// Pregunta es la del fichero de eval que nombra la sesión. No es una clave
	// del fichero: la pone la resolución.
	Pregunta string `yaml:"-"`

	// Respuesta es la de la sesión en su informe, tal cual. La pone la
	// resolución.
	Respuesta string `yaml:"-"`

	// Textos son los que devolvieron las herramientas de la sesión,
	// reconstruidos, y, en un derivado, sin el del bloque quitado. Los pone la
	// resolución.
	Textos []Texto `yaml:"-"`
}

// BloqueQuitado es el bloque de una norma cuyo texto se quita de los de la
// sesión de un caso derivado.
type BloqueQuitado struct {
	// Norma es el identificador de la norma, como BOE-A-2015-10565.
	Norma string `yaml:"norma"`

	// Bloque es el id del bloque, como a21.
	Bloque string `yaml:"bloque"`
}

// nombre nombra el caso en un error: su informe y su sesión y, en un derivado,
// el bloque que se quita, como «<informe> <sesión> sin <norma> <bloque>».
func (c CasoEtiquetado) nombre() string {
	if c.Quitado == nil {
		return c.Informe + " " + c.Sesion
	}

	return fmt.Sprintf("%s %s sin %s %s", c.Informe, c.Sesion, c.Quitado.Norma, c.Quitado.Bloque)
}

// leerCasosEtiquetados lee los casos etiquetados del juez de esa ruta
// (contracts/medida-del-juez.md §4 de H24), con el lector común de documentos
// YAML de internal/skills —una clave repetida es un defecto con sus dos líneas,
// nunca la última que gana— y sin esquema publicado: de un caso solo exige que
// su etiqueta sea una de las dos, porque con otra no se podría contar ni como
// defecto ni como correcto. Lo demás lo exige resolverlo: un caso que no se
// puede resolver es un error de la reconstrucción. Ningún error lleva delante de
// qué casos habla: lo pone quien los lee.
func leerCasosEtiquetados(ruta string) (CasosEtiquetados, error) {
	contenido, err := leerFichero(ruta)
	if err != nil {
		return CasosEtiquetados{}, fmt.Errorf("no se pueden leer: %w", err)
	}

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
// informe: su orden y si fue una llamada a una herramienta, que los informes
// anteriores a los dos modos no dicen.
type invocacionDelInforme struct {
	Orden   string `json:"orden"`
	Llamada bool   `json:"llamada"`
}

// textoReconstruido es el texto de una invocación de una sesión, reconstruido:
// su orden, la del informe, y lo que escribió en la salida estándar al repetirla
// en proceso, con lo que hace falta para saber si es el de un bloque que un
// derivado quita.
type textoReconstruido struct {
	Texto

	// argumentos son los de la invocación repetida, sin el nombre del programa:
	// el applet, el verbo y lo que les sigue.
	argumentos []string

	// codigo es el código con el que terminó.
	codigo int
}

// sesionReconstruida es lo que la resolución de un caso toma de su sesión: la
// pregunta de su eval, su respuesta y los textos de sus invocaciones.
type sesionReconstruida struct {
	pregunta  string
	respuesta string
	textos    []textoReconstruido
}

// sesionDeUnInforme identifica una sesión de un informe versionado.
type sesionDeUnInforme struct {
	informe string
	sesion  string
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
// Recuerda lo que ya ha reconstruido de cada sesión, de modo que resolver otra
// vez un caso suyo, u otro de la misma sesión, no la repite: con un mismo
// reconstructor, cada sesión se reconstruye una sola vez. Se puede usar desde
// varias gorrutinas, que resuelven una detrás de otra.
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

	// informes son las sesiones de cada informe ya leído, por su nombre.
	informes map[string]map[string]sesionDelInforme

	// sesiones son las ya reconstruidas.
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
		informes:  map[string]map[string]sesionDelInforme{},
		sesiones:  map[sesionDeUnInforme]sesionReconstruida{},
	}
}

// resolver devuelve los casos, en su orden, resueltos: cada uno con la
// pregunta, la respuesta y los textos de su sesión y, si es un derivado, sin el
// texto del bloque que quita (sinElBloque). El primer caso que no se puede
// resolver es el error, que lo nombra, y entonces no devuelve ninguno: su
// informe no se puede leer o no tiene su sesión, la eval de esa sesión no está
// ni entre las de hoy ni entre las retiradas, lo grabado no sirve la base o el
// grafo previo, o el derivado no tiene ningún texto que quitar.
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
// sesión; en un derivado, los textos sin el del bloque que quita.
func (r *reconstructor) resolverElCaso(caso CasoEtiquetado) (CasoEtiquetado, error) {
	sesion, err := r.sesionDelCaso(caso)
	if err != nil {
		return CasoEtiquetado{}, err
	}

	caso.Pregunta, caso.Respuesta = sesion.pregunta, sesion.respuesta

	if caso.Quitado != nil {
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

// sesionDelCaso es la sesión del caso, reconstruida: la que ya se recordaba o,
// si es la primera vez, la que se lee de su informe y se reconstruye con su
// eval, que entonces se recuerda.
func (r *reconstructor) sesionDelCaso(caso CasoEtiquetado) (sesionReconstruida, error) {
	clave := sesionDeUnInforme{informe: caso.Informe, sesion: caso.Sesion}
	if reconstruida, recordada := r.sesiones[clave]; recordada {
		return reconstruida, nil
	}

	sesion, err := r.sesionDelInforme(caso.Informe, caso.Sesion)
	if err != nil {
		return sesionReconstruida{}, err
	}

	eval, grafosPrevios, err := r.evalDeLaSesion(sesion.Eval)
	if err != nil {
		return sesionReconstruida{}, err
	}

	textos, err := r.textosDeLaSesion(sesion, eval, grafosPrevios)
	if err != nil {
		return sesionReconstruida{}, err
	}

	reconstruida := sesionReconstruida{pregunta: eval.Pregunta, respuesta: sesion.Respuesta, textos: textos}
	r.sesiones[clave] = reconstruida

	return reconstruida, nil
}

// sesionDelInforme es la sesión de ese nombre del informe de esa ruta, que se
// lee la primera vez que un caso lo nombra. Es un error que el informe no se
// pueda leer o que no tenga esa sesión.
func (r *reconstructor) sesionDelInforme(informe, nombre string) (sesionDelInforme, error) {
	sesiones, leido := r.informes[informe]
	if !leido {
		var err error

		sesiones, err = leerSesionesDelInforme(r.raiz, informe)
		if err != nil {
			return sesionDelInforme{}, err
		}

		r.informes[informe] = sesiones
	}

	sesion, esta := sesiones[nombre]
	if !esta {
		return sesionDelInforme{}, fmt.Errorf("el informe %s no tiene la sesión %s", informe, nombre)
	}

	return sesion, nil
}

// leerSesionesDelInforme lee las sesiones del informe versionado de esa ruta,
// que va desde la raíz y con la barra de separador, y las devuelve por su
// nombre. El informe solo se lee, y nunca de fuera de la raíz: una ruta que
// sale de ella no se puede leer. Es un error que no sea un documento JSON con la
// forma de un informe o que repita el nombre de una sesión: no habría una sola
// respuesta que dar por la del caso.
func leerSesionesDelInforme(raiz, informe string) (map[string]sesionDelInforme, error) {
	contenido, err := fs.ReadFile(os.DirFS(raiz), informe)
	if err != nil {
		return nil, fmt.Errorf("el informe %s no se puede leer: %w", informe, err)
	}

	var leido struct {
		Evals []sesionDelInforme `json:"evals"`
	}

	if err := json.Unmarshal(contenido, &leido); err != nil {
		return nil, fmt.Errorf("el informe %s no es un informe del job de evals: %w", informe, err)
	}

	sesiones := make(map[string]sesionDelInforme, len(leido.Evals))

	for _, sesion := range leido.Evals {
		if _, repetida := sesiones[sesion.Sesion]; repetida {
			return nil, fmt.Errorf("el informe %s tiene repetida la sesión %s", informe, sesion.Sesion)
		}

		sesiones[sesion.Sesion] = sesion
	}

	return sesiones, nil
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
// (repetirLasInvocaciones). El grafo de la sesión, que vive en esa caché, se
// descarta con ella.
func (r *reconstructor) textosDeLaSesion(sesion sesionDelInforme, eval Eval, grafosPrevios string) (
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

	return repetirLasInvocaciones(registro, sesion.Invocaciones), nil
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
// con el reloj del sistema, sobre el grafo del mundo de esa caché; y la entrega
// a ese mismo grafo de lo que cada invocación observa. Son los únicos applets
// que invocaron las sesiones de los informes (research M6 de H24).
func registroDeLaSesion(sinGrabaciones, dirCache string) (*app.Registro, error) {
	registro, err := registroDeBoe(sinGrabaciones, cache.ConDirectorio(dirCache))
	if err != nil {
		return nil, err
	}

	delGrafo := app.DependenciasDelGrafoDelSistema()
	delGrafo.Almacen = []graph.Opcion{graph.ConDirectorio(dirCache)}

	grafo := app.AppletGrafo(delGrafo)
	if err := registro.Registrar(grafo); err != nil {
		return nil, fmt.Errorf("el applet %s no se puede registrar: %w", grafo.Nombre(), err)
	}

	registro.EntregarAlGrafo(graph.Nuevo(graph.ConDirectorio(dirCache)))

	return registro, nil
}

// repetirLasInvocaciones repite en proceso, con app.Main y en su orden, cada
// invocación de una sesión salvo la del proceso del servidor, y devuelve un
// texto por cada una: su orden, la del informe tal cual, y lo que escribe en la
// salida estándar, termine con el código que termine. Lo que escribe en la
// salida de error no es de ningún texto.
func repetirLasInvocaciones(registro *app.Registro, invocaciones []invocacionDelInforme) []textoReconstruido {
	textos := make([]textoReconstruido, 0, len(invocaciones))

	for _, invocacion := range invocaciones {
		if invocacion.Orden == ordenDelProcesoDelServidor {
			continue
		}

		argumentos := argumentosDeLaInvocacion(invocacion)

		var salida bytes.Buffer

		codigo := app.Main(slices.Concat([]string{programaDeLasConsultas}, argumentos), registro, &salida, io.Discard,
			sinDatosDeConstruccion, sinDatosDeConstruccion, sinDatosDeConstruccion)

		textos = append(textos, textoReconstruido{
			Texto:      Texto{Orden: invocacion.Orden, Salida: salida.String()},
			argumentos: argumentos,
			codigo:     codigo,
		})
	}

	return textos
}

// argumentosDeLaInvocacion son los argumentos con los que se repite una
// invocación de una sesión (contracts/medida-del-juez.md §5, paso 3, de H24):
// las palabras de su orden; en una llamada a una herramienta, cuya orden empieza
// por <applet>_<verbo>, el applet y el verbo por separado y, detrás de sus
// argumentos, --json, para que su texto sea el sobre, como el de una orden; y,
// detrás de todo, --offline, para que nada llegue a la red.
func argumentosDeLaInvocacion(invocacion invocacionDelInforme) []string {
	palabras := strings.Fields(invocacion.Orden)

	if invocacion.Llamada && len(palabras) > 0 {
		applet, verbo, _ := strings.Cut(palabras[0], "_")
		palabras = slices.Concat([]string{applet, verbo}, palabras[1:], []string{"--json"})
	}

	return append(palabras, "--offline")
}

// sinElBloque devuelve los textos de una sesión sin el del bloque que quita un
// derivado (contracts/medida-del-juez.md §5, paso 4, de H24): fuera el texto de
// cada boe articulo de esa norma y ese bloque que terminó con 0; y del de cada
// boe articulos que lo pidió y terminó con 0, los elementos de data con ese
// bloque, con el sobre vuelto a escribir, o fuera el texto si no le queda
// ninguno (sobreSinElBloque). Los demás van tal cual y en su orden. Si no hay
// nada que quitar, no hay derivado: es un error.
func sinElBloque(textos []textoReconstruido, quitado BloqueQuitado) ([]Texto, error) {
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
func (t textoReconstruido) leyo(bloque BloqueQuitado) bool {
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
