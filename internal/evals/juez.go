package evals

import (
	"bytes"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// carpetaDelJuez es el nombre exacto de la carpeta del juez en la carpeta de
// evals de una skill: LeerConjunto la reconoce por él y no la lee como eval
// (contracts/juez-y-voto.md §1 de H24; research D20 de H24).
const carpetaDelJuez = "juez"

// Los cinco ficheros de la carpeta del juez (contracts/juez-y-voto.md §1 de
// H24): la declaración de clases, que se escribe en ella, y las cuatro copias de
// lo que validó al juez.
const (
	ficheroDeClasesDelJuez  = "clases.yaml"
	ficheroDeRubricaDelJuez = "rubrica.md"
	ficheroDeEsquemaDelJuez = "esquema.json"
	ficheroDeCasosDelJuez   = "casos.yaml"
	ficheroDeMedidaDelJuez  = "medida.json"
)

// rutaDelEsquemaDeClasesDelJuez es el esquema publicado de la declaración de
// clases, relativo al directorio de este paquete, como el de eval.
const rutaDelEsquemaDeClasesDelJuez = "../../schemas/juez-clases.yaml.json"

// motivoNoDirectorio es el motivo de la entrada juez que no es un directorio.
const motivoNoDirectorio = "no es un directorio"

// Juez es el juez con modelo de una skill: lo que LeerConjunto lee de la carpeta
// juez de su carpeta de evals (data-model §1 y contracts/juez-y-voto.md §1 de
// H24; FR-020). Una skill sin esa carpeta no tiene juez.
type Juez struct {
	// Clases son las de clases.yaml, en su orden: al menos una y sin nombres
	// repetidos.
	Clases []ClaseDelJuez

	// Rubrica es el contenido de rubrica.md, entero: va tal cual como
	// instrucciones del juez.
	Rubrica string

	// Esquema es el contenido de esquema.json, entero: la forma de la respuesta
	// del juez, con la que se valida cada voto. Sus propiedades son exactamente
	// los nombres de Clases.
	Esquema string

	// Casos es la ruta de casos.yaml, los casos etiquetados con los que se mide
	// al juez.
	Casos string

	// Medida es la ruta de medida.json, la medida versionada del juez.
	Medida string
}

// ClaseDelJuez es una clase de la declaración de clases del juez de una skill,
// clases.yaml: lo que el juez mira en cada respuesta (FR-020 y FR-021 de H24).
type ClaseDelJuez struct {
	// Nombre es el de la clase, con la forma ^[a-z0-9_]+$: el de su propiedad en
	// el esquema de la respuesta del juez.
	Nombre string `yaml:"nombre"`

	// Decide dice si la clase decide —su umbral del informe hace fallar el job—
	// o solo se publica.
	Decide bool `yaml:"decide"`

	// Umbral es la proporción de respuestas marcadas en la clase que se admite,
	// de 0 a 1.
	Umbral float64 `yaml:"umbral"`
}

// declaracionDeClases es el documento clases.yaml de la carpeta del juez.
type declaracionDeClases struct {
	Clases []ClaseDelJuez `yaml:"clases"`
}

// esquemaDeClasesDelJuez compila una sola vez el esquema publicado de la
// declaración de clases, que no cambia mientras se ejecutan los tests.
var esquemaDeClasesDelJuez = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return compilarEsquemaPublicado(rutaDelEsquemaDeClasesDelJuez, "de la declaración de clases del juez")
})

// leerJuez lee el juez de una skill de la entrada juez de su directorio de evals
// (contracts/juez-y-voto.md §1 de H24; FR-020). Si la entrada no es un
// directorio, es ella el fichero mal formado. Si lo es, lee sus cinco ficheros
// en orden de nombre y devuelve un fichero mal formado por cada uno que falta,
// que no se puede leer o que no tiene su forma, con su nombre dentro del
// directorio de evals, juez/<fichero>, que es por el que empieza su error:
//
//   - clases.yaml se lee con leerClasesDelJuez;
//   - esquema.json, con propiedadesDelEsquema, y, si la declaración se ha leído,
//     sus propiedades tienen que ser exactamente las clases declaradas;
//   - de rubrica.md, de casos.yaml y de medida.json solo se exige que se puedan
//     leer.
//
// El juez se devuelve solo si ninguno está mal formado: con la rúbrica y el
// esquema enteros y con la ruta de los casos y la de la medida.
func leerJuez(dir string, entrada fs.DirEntry) (*Juez, []FicheroMalFormado) {
	if !entrada.IsDir() {
		return nil, []FicheroMalFormado{{
			Fichero: carpetaDelJuez,
			Error:   fmt.Errorf("%s: %s", carpetaDelJuez, motivoNoDirectorio),
		}}
	}

	lectura := lecturaDelJuez{carpeta: filepath.Join(dir, carpetaDelJuez)}

	lectura.contenido(ficheroDeCasosDelJuez)
	clases := lectura.clases()
	esquema := lectura.esquema(clases)
	lectura.contenido(ficheroDeMedidaDelJuez)
	rubrica, _ := lectura.contenido(ficheroDeRubricaDelJuez)

	if len(lectura.malFormados) > 0 {
		return nil, lectura.malFormados
	}

	return &Juez{
		Clases:  clases,
		Rubrica: string(rubrica),
		Esquema: string(esquema),
		Casos:   filepath.Join(lectura.carpeta, ficheroDeCasosDelJuez),
		Medida:  filepath.Join(lectura.carpeta, ficheroDeMedidaDelJuez),
	}, nil
}

// lecturaDelJuez es la lectura de la carpeta del juez de una skill: su ruta y
// los ficheros mal formados que va encontrando, en el orden en que se leen.
type lecturaDelJuez struct {
	carpeta     string
	malFormados []FicheroMalFormado
}

// malFormado anota el fichero de la carpeta como mal formado por ese motivo. Su
// nombre es el que tiene dentro del directorio de evals, juez/<fichero>, y su
// error empieza por él, como el de una eval.
func (l *lecturaDelJuez) malFormado(fichero string, motivo error) {
	nombre := carpetaDelJuez + "/" + fichero

	l.malFormados = append(l.malFormados, FicheroMalFormado{Fichero: nombre, Error: fmt.Errorf("%s: %w", nombre, motivo)})
}

// contenido lee entero el fichero de la carpeta con ese nombre. Si falta o no se
// puede leer, lo anota como mal formado y leido es falso.
func (l *lecturaDelJuez) contenido(fichero string) (contenido []byte, leido bool) {
	contenido, err := leerFichero(filepath.Join(l.carpeta, fichero))
	if err != nil {
		l.malFormado(fichero, fmt.Errorf("no se puede leer: %w", err))

		return nil, false
	}

	return contenido, true
}

// clases lee la declaración de clases de la carpeta, clases.yaml, con
// leerClasesDelJuez. Si no se puede leer o está mal formada, la anota y no
// devuelve ninguna clase.
func (l *lecturaDelJuez) clases() []ClaseDelJuez {
	contenido, leido := l.contenido(ficheroDeClasesDelJuez)
	if !leido {
		return nil
	}

	clases, err := leerClasesDelJuez(contenido)
	if err != nil {
		l.malFormado(ficheroDeClasesDelJuez, err)

		return nil
	}

	return clases
}

// esquema lee el esquema de la respuesta del juez de la carpeta, esquema.json, y
// devuelve su contenido. Lo anota como mal formado si no se puede leer, si no es
// un documento JSON del que leer las propiedades o si sus propiedades no son
// exactamente las clases declaradas. Sin clases, que es que la declaración no se
// ha podido leer, no hay con qué compararlas: el fichero mal formado es la
// declaración, que ya está anotada.
func (l *lecturaDelJuez) esquema(clases []ClaseDelJuez) []byte {
	contenido, leido := l.contenido(ficheroDeEsquemaDelJuez)
	if !leido {
		return nil
	}

	propiedades, err := propiedadesDelEsquema(contenido)
	if err != nil {
		l.malFormado(ficheroDeEsquemaDelJuez, err)

		return nil
	}

	if len(clases) == 0 {
		return contenido
	}

	declaradas := make([]string, 0, len(clases))
	for _, clase := range clases {
		declaradas = append(declaradas, clase.Nombre)
	}

	slices.Sort(declaradas)

	if !slices.Equal(propiedades, declaradas) {
		l.malFormado(ficheroDeEsquemaDelJuez, fmt.Errorf("sus propiedades (%s) no son exactamente las clases declaradas (%s)",
			strings.Join(propiedades, ", "), strings.Join(declaradas, ", ")))

		return nil
	}

	return contenido
}

// leerClasesDelJuez lee el contenido de clases.yaml con el lector común de
// documentos YAML de internal/skills —una clave repetida es un defecto con sus
// dos líneas, nunca la última que gana— y lo valida contra
// juez-clases.yaml.json, como la lista de expresiones prohibidas. Devuelve las
// clases en el orden del fichero; un nombre que se repite es un error que lo
// dice, porque el esquema no puede exigir que sean distintos.
func leerClasesDelJuez(contenido []byte) ([]ClaseDelJuez, error) {
	esquema, err := esquemaDeClasesDelJuez()
	if err != nil {
		return nil, err
	}

	leida, err := skills.ValidarDocumentoYAML[declaracionDeClases](contenido, esquema)
	if err != nil {
		return nil, err
	}

	for posicion, clase := range leida.Clases {
		repetida := slices.ContainsFunc(leida.Clases[:posicion], func(anterior ClaseDelJuez) bool {
			return anterior.Nombre == clase.Nombre
		})
		if repetida {
			return nil, fmt.Errorf("la clase %s está repetida", clase.Nombre)
		}
	}

	return leida.Clases, nil
}

// propiedadesDelEsquema devuelve, en orden, los nombres de las propiedades del
// esquema de la respuesta del juez: las claves de properties de su raíz, ninguna
// si no la tiene. El error es el de un contenido que no es un objeto JSON con
// esa forma.
func propiedadesDelEsquema(contenido []byte) ([]string, error) {
	var esquema struct {
		Propiedades map[string]jsontext.Value `json:"properties"`
	}

	if err := json.Unmarshal(contenido, &esquema); err != nil {
		return nil, fmt.Errorf("no se puede leer como JSON: %w", err)
	}

	return slices.Sorted(maps.Keys(esquema.Propiedades)), nil
}

// Las piezas fijas del mensaje del voto, las de prompt_de de
// evidencias/adr-0037/guiones/juez.py (contracts/juez-y-voto.md §3 de H24): lo
// que va en lugar de los textos cuando ninguna herramienta devolvió ninguno, lo
// que va en lugar de la salida de un texto que queda vacía y la frase final. La
// frase dice «las dos preguntas» porque la medida del juez se hizo con ella
// (research D3 de H24).
const (
	mensajeSinTextos = "(ninguna herramienta devolvió ningún texto)"
	mensajeSinSalida = "(sin salida)"
	peticionDelVoto  = "Responde a las dos preguntas de la rúbrica sobre esta respuesta."
)

// mensajeDelVoto es el mensaje con el que se pide cada voto de una respuesta,
// el de prompt_de carácter a carácter (contracts/juez-y-voto.md §3 de H24;
// FR-001, FR-004): sus partes, unidas por una línea en blanco, son un <texto
// orden="…"> por texto de las herramientas, en su orden, con su salida sin
// blancos en los extremos; la pregunta de la eval; la respuesta de la sesión; y
// la frase final. La pregunta, la respuesta y la orden van tal cual.
//
// No lleva nada más (FR-002): ni SKILL.md, ni lo que la eval espera, ni el
// juicio sin modelo de la sesión. Por eso no recibe ni la eval ni su resultado.
func mensajeDelVoto(pregunta, respuesta string, textos []Texto) string {
	partes := make([]string, 0, len(textos)+6)
	partes = append(partes, "<textos_de_las_herramientas>")

	if len(textos) == 0 {
		partes = append(partes, mensajeSinTextos)
	}

	for _, texto := range textos {
		salida := strings.TrimFunc(texto.Salida, esBlanco)
		if salida == "" {
			salida = mensajeSinSalida
		}

		partes = append(partes, `<texto orden="`+texto.Orden+`">`+"\n"+salida+"\n</texto>")
	}

	partes = append(partes,
		"</textos_de_las_herramientas>",
		"<pregunta>\n"+pregunta+"\n</pregunta>",
		"<respuesta>\n"+respuesta+"\n</respuesta>",
		peticionDelVoto,
	)

	return strings.Join(partes, "\n\n")
}

// esBlanco dice si el carácter es un blanco de los que Python quita con strip y
// casa con \s sobre texto: los de unicode.IsSpace más U+001C a U+001F (research
// D8, V6 y S2 de H24). El \s de regexp, que solo tiene cinco de ASCII, no sirve
// (research V5 de H24).
func esBlanco(caracter rune) bool {
	return unicode.IsSpace(caracter) || ('\x1c' <= caracter && caracter <= '\x1f')
}

// sinEnfasisNiBlancosDeMas es el texto como lo compara la comprobación de la
// frase, el de normal de evidencias/adr-0037/guiones/juez.py
// (contracts/juez-y-voto.md §6 de H24): sin *, _ ni acento grave, estén donde
// estén; con cada serie de blancos cambiada por un espacio; y sin los de los
// extremos. Primero se quita el énfasis: los blancos que deja juntos son una
// sola serie.
func sinEnfasisNiBlancosDeMas(texto string) string {
	var normal strings.Builder

	normal.Grow(len(texto))

	trasUnBlanco := false

	for _, caracter := range texto {
		switch {
		case caracter == '*' || caracter == '_' || caracter == '`':
		case esBlanco(caracter):
			trasUnBlanco = true
		default:
			if trasUnBlanco && normal.Len() > 0 {
				normal.WriteByte(' ')
			}

			trasUnBlanco = false

			normal.WriteRune(caracter)
		}
	}

	return normal.String()
}

// fraseEsta dice si la frase que cita un voto está en la respuesta, sin ningún
// modelo: frase_esta de evidencias/adr-0037/guiones/juez.py, con la que se
// calculó la medida del juez (contracts/juez-y-voto.md §6 de H24; FR-005). La
// frase está si, sin énfasis ni blancos de más y sin quedar vacía, es subcadena
// de la respuesta sin los suyos. Nada más se tolera: mayúsculas, acentos y
// puntuación se comparan tal cual. Las formas fijas de avisos.go son otra
// comprobación, y no se usan aquí.
func fraseEsta(frase, respuesta string) bool {
	buscada := sinEnfasisNiBlancosDeMas(frase)

	return buscada != "" && strings.Contains(sinEnfasisNiBlancosDeMas(respuesta), buscada)
}

// Votante pide un voto al juez: recibe el mensaje del voto y devuelve la salida
// estándar de la sesión del juez, la de claude -p --output-format json
// (data-model §3 y contracts/juez-y-voto.md §5 de H24; research D5 de H24). Los
// tests dan uno que devuelve salidas grabadas; los puntos de entrada, el que
// ejecuta el guion del voto.
//
// Su error dice que el proceso del voto no terminó bien: es errTopeDelVoto, o
// lo envuelve, si agotó su tope, y el voto no llega a darse. Cualquier otro no
// impide leer la salida que el proceso dejó, que se devuelve con él: una sesión
// del juez que termina con error la escribe igual. Si el proceso terminó con un
// código, el error lo lleva en un método ExitCode() int, como *exec.ExitError,
// y va en el motivo de la salida que no es JSON.
type Votante func(mensaje string) ([]byte, error)

// errTopeDelVoto es el error del Votante cuyo voto agotó su tope (FR-007 de
// H24).
var errTopeDelVoto = errors.New("el voto agotó su tope")

// errorConCodigo es el error de un Votante que lleva el código con el que
// terminó el proceso del voto, como *exec.ExitError.
type errorConCodigo interface {
	error
	ExitCode() int
}

// codigoSinProceso es el código de un voto cuyo error no lleva ninguno, el de
// un proceso que no llegó a terminar: el mismo que da ExitCode de un proceso
// sin código.
const codigoSinProceso = -1

// El tope de un voto y su margen (contracts/juez-y-voto.md §4 de H24; research
// D6 de H24; FR-007): pasado el tope, el proceso del voto se termina, y pasado
// además el margen, se cierran las salidas que haya dejado abiertas. Con seis
// votos por respuesta como mucho, es lo que cabe en el tope del trabajo del job
// junto al peor caso de sus sesiones. causaDelTope dice el tope.
const (
	topeDelVoto   = 35 * time.Second
	margenDelVoto = 5 * time.Second
)

// Lo que hay en el directorio del juez además de modelo.txt, rubrica.md y
// esquema.json, que el guion del voto lee desde el directorio de cada voto
// (contracts/juez-y-voto.md §4 de H24): el directorio de trabajo de los votos y
// el de Claude Code, los de la validación del ADR 0037. Y el prefijo de su
// nombre en el temporal del sistema.
const (
	// directorioDelVoto es el directorio de trabajo de cada voto, vacío.
	directorioDelVoto = "cwd"

	// directorioDeClaudeDelJuez es el directorio de configuración de Claude Code
	// de los votos, CLAUDE_CONFIG_DIR.
	directorioDeClaudeDelJuez = "config"

	prefijoDelDirectorioDelJuez = "kitlegal-evals-juez-"
)

// ordenDelVoto es con lo que el votante del guion abre cada voto de una
// ejecución, lo mismo en todos (contracts/juez-y-voto.md §4 de H24).
type ordenDelVoto struct {
	// juez es el de la skill: su rúbrica son las instrucciones de cada voto, y
	// su esquema, la forma que se le pide a su respuesta.
	juez *Juez

	// modelo es el id del modelo del juez.
	modelo string

	// guion es la ruta absoluta de scripts/evals-voto.sh, que abre cada voto.
	guion string

	// path es el PATH de cada voto: su primer claude es el Claude Code del juez.
	path string

	// suscripcion es el valor de CLAUDE_CODE_OAUTH_TOKEN de quien pide los
	// votos.
	suscripcion string
}

// nuevoVotanteDelGuion da el Votante de los puntos de entrada, el que abre cada
// voto con el guion del voto, con el tope de un voto y su margen
// (contracts/juez-y-voto.md §4 de H24; research D4 y D5 de H24; FR-003, FR-004,
// FR-007). Crea el directorio del juez, un temporal de esta ejecución, y
// devuelve con el votante la función que lo retira, que llama quien lo pidió
// cuando ya no va a pedir más votos.
//
// El contexto es el de las señales del punto de entrada: Votante no recibe
// ninguno en cada voto (research D5 de H24), así que el de cada voto deriva de
// este, y SIGINT y SIGTERM cortan los votos abiertos. El votante se puede
// llamar desde varias gorrutinas a la vez, una por respuesta que se juzga.
func nuevoVotanteDelGuion(ctx context.Context, orden ordenDelVoto) (Votante, func() error, error) {
	return nuevoVotanteConTope(ctx, orden, topeDelVoto, margenDelVoto)
}

// nuevoVotanteConTope es nuevoVotanteDelGuion con el tope de cada voto y su
// margen como parámetros, para que un test los fije en lo que puede esperar.
// Sin un guion con ruta absoluta no hay votante: cada voto se ejecuta en su
// propio directorio, y desde él se buscaría un guion con ruta relativa.
func nuevoVotanteConTope(ctx context.Context, orden ordenDelVoto, tope, margen time.Duration) (
	Votante, func() error, error,
) {
	if !filepath.IsAbs(orden.guion) {
		return nil, nil, fmt.Errorf("el guion del voto %s no tiene ruta absoluta", orden.guion)
	}

	dir, err := crearElDirectorioDelJuez(orden)
	if err != nil {
		return nil, nil, err
	}

	voto := votoEnMarcha{
		dir: filepath.Join(dir, directorioDelVoto),
		entorno: []string{
			variableDelPATH + "=" + orden.path,
			"HOME=" + dir,
			"CLAUDE_CONFIG_DIR=" + filepath.Join(dir, directorioDeClaudeDelJuez),
			variableDeLaSuscripcion + "=" + orden.suscripcion,
		},
		tope:   tope,
		margen: margen,
	}

	votar := func(mensaje string) ([]byte, error) { return ejecutarElVoto(ctx, orden.guion, voto, mensaje) }
	retirar := func() error { return retirarTemporal(dir) }

	return votar, retirar, nil
}

// crearElDirectorioDelJuez crea el directorio del juez de una ejecución, un
// temporal fuera del repositorio, con lo que escribirElDirectorioDelJuez deja
// en él, y devuelve su ruta. Si no se puede crear entero, no queda nada de él.
func crearElDirectorioDelJuez(orden ordenDelVoto) (string, error) {
	dir, err := os.MkdirTemp("", prefijoDelDirectorioDelJuez)
	if err == nil {
		err = escribirElDirectorioDelJuez(dir, orden)
	}

	if err != nil {
		// Si el directorio llegó a crearse, se retira; si no, dir está vacío y
		// no hay nada que retirar.
		return "", errors.Join(fmt.Errorf("el directorio del juez no se puede crear: %w", err), retirarTemporal(dir))
	}

	return dir, nil
}

// escribirElDirectorioDelJuez deja en el directorio del juez lo que el guion
// del voto lee desde el directorio de cada voto —modelo.txt, con el id del
// modelo en su línea; rubrica.md, con la rúbrica entera; y esquema.json, con el
// esquema entero— y, vacíos y solo para quien vota, el directorio de trabajo de
// los votos y el de Claude Code.
func escribirElDirectorioDelJuez(dir string, orden ordenDelVoto) error {
	ficheros := []struct{ nombre, contenido string }{
		{ficheroDelModelo, orden.modelo + "\n"},
		{ficheroDeRubricaDelJuez, orden.juez.Rubrica},
		{ficheroDeEsquemaDelJuez, orden.juez.Esquema},
	}
	for _, fichero := range ficheros {
		if err := escribirFichero(filepath.Join(dir, fichero.nombre), []byte(fichero.contenido)); err != nil {
			return err
		}
	}

	for _, directorio := range []string{directorioDelVoto, directorioDeClaudeDelJuez} {
		if err := os.Mkdir(filepath.Join(dir, directorio), permisosDeLaSesion); err != nil {
			return err
		}
	}

	return nil
}

// votoEnMarcha es lo que ejecutarElVoto necesita de un voto, además del guion y
// del mensaje: su directorio de trabajo, su entorno, su tope y el margen tras
// él.
type votoEnMarcha struct {
	dir     string
	entorno []string
	tope    time.Duration
	margen  time.Duration
}

// ejecutarElVoto abre un voto: un proceso nuevo del guion del voto, sin
// argumentos y con el mensaje por su entrada estándar, y devuelve lo que
// escribió en la estándar (contracts/juez-y-voto.md §4 de H24; FR-003, FR-007).
// El guion llega como parámetro para que la orden no lleve nada que no sea
// constante (research D4 y V8 de H24).
//
// El contexto del voto deriva del recibido, con el tope: cuando termina, exec
// termina el proceso, y si pasado el margen alguien sigue teniendo abiertas sus
// salidas, las cierra y deja de esperar (research V7 de H24). El error es
// entonces la causa de ese contexto: errTopeDelVoto si se agotó el tope, y la
// del recibido si fue él el que se cortó, que es una interrupción y no un voto
// lento. Del proceso que termina por sí mismo con un código distinto de 0, el
// error es el de conSuSalidaDeError.
func ejecutarElVoto(ctx context.Context, guion string, voto votoEnMarcha, mensaje string) ([]byte, error) {
	contextoDelVoto, cancelar := context.WithTimeoutCause(ctx, voto.tope, errTopeDelVoto)
	defer cancelar()

	// exec solo llama a Cancel si el contexto termina antes que el proceso, así
	// que cortado dice si al voto lo terminó su contexto y no él mismo.
	var cortado atomic.Bool

	orden := exec.CommandContext(contextoDelVoto, guion)
	orden.Dir = voto.dir
	orden.Env = voto.entorno
	orden.Stdin = strings.NewReader(mensaje)
	orden.WaitDelay = voto.margen
	orden.Cancel = func() error {
		err := orden.Process.Kill()
		cortado.Store(err == nil)

		return err
	}

	salida, err := orden.Output()

	switch {
	case err == nil:
		return salida, nil
	case cortado.Load():
		return salida, fmt.Errorf("el voto se ha cortado antes de terminar: %w", context.Cause(contextoDelVoto))
	default:
		return salida, conSuSalidaDeError(err)
	}
}

// conSuSalidaDeError es el error del proceso de un voto con lo que el proceso
// escribió en su salida de error detrás, si terminó con un código y escribió
// algo: es lo que dice por qué una sesión del juez no dio su salida, y va en el
// motivo del voto que no llega a darse, donde
// evidencias/adr-0037/guiones/juez.py ponía esa salida. El error sigue llevando
// el código del proceso.
func conSuSalidaDeError(err error) error {
	delProceso, conCodigo := errors.AsType[*exec.ExitError](err)
	if !conCodigo {
		return err
	}

	escrito := strings.TrimFunc(string(delProceso.Stderr), esBlanco)
	if escrito == "" {
		return err
	}

	return fmt.Errorf("%w: %s", err, escrito)
}

// respuestaSi es la respuesta con la que el juez marca una clase; la otra del
// esquema de su respuesta es «no».
const respuestaSi = "si"

// votosParaMarcar son los votos que una clase que decide necesita en sí, cada
// uno con su frase, para marcar una respuesta, y los que como mucho se piden de
// una respuesta sin contar la repetición de los nulos
// (contracts/juez-y-voto.md §8 de H24; FR-010).
const votosParaMarcar = 3

// Las causas de un voto que no llega a darse (contracts/juez-y-voto.md §5 de
// H24; FR-007): con el número del voto delante dan el motivo de la respuesta
// sin juzgar. Las tres últimas llevan detrás un texto, en una línea y cortado a
// caracteresDeLaCausa caracteres; la de la salida, además, el código del
// proceso.
const (
	causaDelTope    = "tope de 35 s agotado"
	causaDeLaSesion = "la sesión del juez terminó con error: "
	causaDeLaSalida = "la salida no es JSON (código %d): %s"
	causaDeLaForma  = "la respuesta no tiene la forma del esquema: "

	caracteresDeLaCausa = 300
)

// vallaDeCodigo es lo que abre y cierra el bloque de código en el que el juez
// puede envolver su juicio cuando lo da en result.
const vallaDeCodigo = "```"

// VotoDeClase es lo que un voto del juez dice de una clase de una respuesta
// (data-model §3 de H24): los campos de esquema.json para esa clase, tal como
// los dio el juez, y lo que el job comprueba de ellos sin modelo.
type VotoDeClase struct {
	// Voto es el número del voto, de 1 a 3. Un voto nulo y su repetición llevan
	// el mismo.
	Voto int

	// Nulo dice si el voto es nulo: en alguna clase dice sí y su frase no está
	// en la respuesta. Es del voto entero, así que lo llevan igual sus votos de
	// todas las clases (FR-006 de H24).
	Nulo bool

	// Motivo es el motivo que el juez da de su respuesta.
	Motivo string

	// Respuesta es «si» o «no».
	Respuesta string

	// Frase es la frase de la respuesta juzgada que el juez cita como prueba de
	// su sí; vacía con un no.
	Frase string

	// Precepto es el precepto del que habla la frase, en la clase cuyo esquema
	// lo tiene; nil en la que no.
	Precepto *string

	// FraseEnLaRespuesta dice si Frase está en la respuesta juzgada, con la
	// tolerancia de fraseEsta. Una frase vacía no está.
	FraseEnLaRespuesta bool
}

// diceSi dice si el voto dice sí de su clase con su frase en la respuesta, que
// es lo único que cuenta como un sí en la regla de los votos.
func (v VotoDeClase) diceSi() bool {
	return v.Respuesta == respuestaSi && v.FraseEnLaRespuesta
}

// JuicioDeClase es lo que el juez deja de una respuesta en una de sus clases
// (data-model §3 de H24).
type JuicioDeClase struct {
	// Clase es el nombre de la clase.
	Clase string

	// Marcada dice, en una clase que decide, si tres votos dicen sí con su
	// frase; en una que solo se publica, si lo dice el primero. Con la respuesta
	// sin juzgar es falsa.
	Marcada bool

	// Votos son todos los votos de la respuesta, en su orden, con lo que cada
	// uno dice de la clase: también los nulos, delante de su repetición.
	Votos []VotoDeClase
}

// JuicioDeRespuesta es el juicio del juez con modelo sobre una respuesta, el
// que da la regla de los votos (data-model §3 de H24).
type JuicioDeRespuesta struct {
	// Clases tiene una entrada por clase del juez, en el orden de clases.yaml.
	Clases []JuicioDeClase

	// SinJuzgar es el motivo de la respuesta de la que un voto no llegó a darse,
	// que nombra ese voto; vacío si la respuesta se juzgó. Una respuesta sin
	// juzgar no queda marcada en ninguna clase, y conserva los votos que sí
	// llegaron (FR-007 de H24).
	SinJuzgar string
}

// votacion es con lo que se juzgan las respuestas de una skill: las clases de
// su juez, el esquema compilado de la respuesta del juez y quien vota.
type votacion struct {
	clases  []ClaseDelJuez
	esquema *jsonschema.Schema
	votante Votante
}

// nuevaVotacion prepara la votación de las respuestas de una skill con su juez
// y con quien vota: compila una sola vez el esquema de la respuesta del juez,
// con el que se valida cada voto (FR-007 de H24). El error es el de un esquema
// que no compila, que la lectura de la carpeta del juez no comprueba.
func nuevaVotacion(juez *Juez, votante Votante) (*votacion, error) {
	esquema, err := skills.CompilarEsquema([]byte(juez.Esquema))
	if err != nil {
		return nil, fmt.Errorf("el esquema de la respuesta del juez no sirve para validar sus votos: %w", err)
	}

	return &votacion{clases: juez.Clases, esquema: esquema, votante: votante}, nil
}

// juzgar aplica a una respuesta la regla de los votos, genérica sobre las
// clases del juez (contracts/juez-y-voto.md §8 de H24; research D9 de H24;
// FR-010, FR-011):
//
//   - se vota por orden, hasta tres veces, cada vez con el mismo mensaje, y se
//     deja de votar en cuanto ninguna clase que decide sigue con todos sus votos
//     en sí con su frase;
//   - una clase que decide marca la respuesta solo si tiene sus tres votos en
//     sí;
//   - una clase que solo se publica cuenta con el primer voto, y con ninguno
//     más;
//   - y si un voto no llega a darse, la respuesta queda sin juzgar, con el
//     motivo de ese voto y sin marcar en ninguna clase (FR-007).
//
// De un voto nulo cuenta su repetición (votosDelNumero). El juicio lleva, por
// clase, todos los votos que se dieron.
func (v *votacion) juzgar(pregunta, respuesta string, textos []Texto) JuicioDeRespuesta {
	mensaje := mensajeDelVoto(pregunta, respuesta, textos)

	juicio := JuicioDeRespuesta{Clases: make([]JuicioDeClase, len(v.clases))}

	// siguen dice, de cada clase, si decide y todos sus votos dicen sí con su
	// frase.
	siguen := make([]bool, len(v.clases))

	for posicion, clase := range v.clases {
		juicio.Clases[posicion].Clase = clase.Nombre
		siguen[posicion] = clase.Decide
	}

	for numero := 1; numero <= votosParaMarcar; numero++ {
		dados, motivo := v.votosDelNumero(numero, mensaje, respuesta)

		for _, voto := range dados {
			for posicion := range juicio.Clases {
				juicio.Clases[posicion].Votos = append(juicio.Clases[posicion].Votos, voto[posicion])
			}
		}

		if motivo != "" {
			return juicio.sinJuzgar(motivo)
		}

		if !v.contar(&juicio, siguen, numero, dados[len(dados)-1]) {
			return juicio
		}
	}

	for posicion, clase := range v.clases {
		if clase.Decide {
			juicio.Clases[posicion].Marcada = siguen[posicion]
		}
	}

	return juicio
}

// respuestaAJuzgar es una respuesta que se da a juzgar al juez: la pregunta de
// su eval, la respuesta de la sesión y los textos de sus herramientas, que es
// lo que lleva el mensaje de cada uno de sus votos (mensajeDelVoto).
type respuestaAJuzgar struct {
	pregunta  string
	respuesta string
	textos    []Texto
}

// juzgarTodas juzga las respuestas dadas, como mucho concurrencia a la vez, y
// devuelve sus juicios en su mismo orden (contracts/informe-del-job.md §1 de
// H24; research D6 de H24). Los votos de una misma respuesta van uno detrás de
// otro, que es como los pide juzgar, y cada respuesta se juzga en su gorrutina:
// un canal con concurrencia huecos es el semáforo y un sync.WaitGroup espera a
// las que están en curso, como en ejecutarSesiones. La concurrencia es al menos
// 1, y el votante se puede llamar desde varias gorrutinas a la vez.
func (v *votacion) juzgarTodas(respuestas []respuestaAJuzgar, concurrencia int) []JuicioDeRespuesta {
	juicios := make([]JuicioDeRespuesta, len(respuestas))
	huecos := make(chan struct{}, concurrencia)

	var enCurso sync.WaitGroup

	for posicion, respuesta := range respuestas {
		huecos <- struct{}{}

		enCurso.Go(func() {
			defer func() { <-huecos }()

			// Cada gorrutina escribe solo en la posición de su respuesta.
			juicios[posicion] = v.juzgar(respuesta.pregunta, respuesta.respuesta, respuesta.textos)
		})
	}

	enCurso.Wait()

	return juicios
}

// frasesQueCuentan son las frases de los votos de la clase que cuentan, una por
// número de voto y en su orden: de un voto nulo y su repetición cuenta la
// repetición, que va detrás de él con su mismo número (votosDelNumero). En una
// clase que decide con la respuesta marcada, son las tres frases que la marcan
// (contracts/informe-del-job.md §4 de H24; FR-035).
func (j JuicioDeClase) frasesQueCuentan() []string {
	frases := make([]string, 0, votosParaMarcar)

	for posicion, voto := range j.Votos {
		if posicion+1 < len(j.Votos) && j.Votos[posicion+1].Voto == voto.Voto {
			continue
		}

		frases = append(frases, voto.Frase)
	}

	return frases
}

// contar cuenta en cada clase el voto con ese número, que es el que cuenta de
// él, y dice si alguna clase que decide sigue con todos sus votos en sí con su
// frase: en una que decide, un voto que no lo dice la deja sin marcar para
// siempre; en una que solo se publica, el primero dice si la respuesta cuenta
// en ella, y los demás no cambian nada.
func (v *votacion) contar(juicio *JuicioDeRespuesta, siguen []bool, numero int, voto []VotoDeClase) bool {
	algunaSigue := false

	for posicion, clase := range v.clases {
		si := voto[posicion].diceSi()

		switch {
		case clase.Decide:
			siguen[posicion] = siguen[posicion] && si
			algunaSigue = algunaSigue || siguen[posicion]
		case numero == 1:
			juicio.Clases[posicion].Marcada = si
		}
	}

	return algunaSigue
}

// sinJuzgar es el juicio de la respuesta de la que un voto no llegó a darse,
// con su motivo: conserva los votos que sí llegaron y no queda marcada en
// ninguna clase, tampoco en la que solo se publica y contaba ya con su primer
// voto.
func (j JuicioDeRespuesta) sinJuzgar(motivo string) JuicioDeRespuesta {
	for posicion := range j.Clases {
		j.Clases[posicion].Marcada = false
	}

	j.SinJuzgar = motivo

	return j
}

// votosDelNumero pide el voto con ese número y, si es nulo, otro, una sola vez
// (contracts/juez-y-voto.md §7 de H24; FR-006). Devuelve los que se dieron, en
// su orden: el último es el que cuenta, en todas las clases, también si vuelve
// a ser nulo, y entonces su sí sin frase no cuenta como sí. Si alguno no llega
// a darse, devuelve además su motivo, y ninguno cuenta.
func (v *votacion) votosDelNumero(numero int, mensaje, respuesta string) ([][]VotoDeClase, string) {
	voto, motivo := v.pedirElVoto(numero, mensaje, respuesta)
	if motivo != "" {
		return nil, motivo
	}

	if !esNulo(voto) {
		return [][]VotoDeClase{voto}, ""
	}

	repetido, motivo := v.pedirElVoto(numero, mensaje, respuesta)
	if motivo != "" {
		return [][]VotoDeClase{voto}, motivo
	}

	return [][]VotoDeClase{voto, repetido}, ""
}

// esNulo dice si el voto es nulo. Lo dicen igual sus votos de todas las clases.
func esNulo(voto []VotoDeClase) bool {
	return slices.ContainsFunc(voto, func(deClase VotoDeClase) bool { return deClase.Nulo })
}

// pedirElVoto pide un voto al votante y lo lee: lo que dice de cada clase del
// juez, en su orden, con si su frase está en la respuesta y, en todas, si el
// voto es nulo, que es que en alguna dice sí sin que su frase esté. Si el voto
// no llega a darse, devuelve su motivo, que lo nombra por su número.
func (v *votacion) pedirElVoto(numero int, mensaje, respuesta string) ([]VotoDeClase, string) {
	dichos, causa := v.leerElVoto(v.votante(mensaje))
	if causa != "" {
		return nil, fmt.Sprintf("voto %d: %s", numero, causa)
	}

	voto := make([]VotoDeClase, len(v.clases))
	nulo := false

	for posicion, clase := range v.clases {
		dicho := dichos[clase.Nombre]

		voto[posicion] = VotoDeClase{
			Voto:               numero,
			Motivo:             dicho.Motivo,
			Respuesta:          dicho.Respuesta,
			Frase:              dicho.Frase,
			Precepto:           dicho.Precepto,
			FraseEnLaRespuesta: fraseEsta(dicho.Frase, respuesta),
		}

		nulo = nulo || (dicho.Respuesta == respuestaSi && !voto[posicion].FraseEnLaRespuesta)
	}

	for posicion := range voto {
		voto[posicion].Nulo = nulo
	}

	return voto, ""
}

// salidaDelVoto es lo que se lee de la salida estándar de la sesión del juez,
// el objeto de claude -p --output-format json (contracts/juez-y-voto.md §5 de
// H24): si la sesión terminó con error, el juicio con la forma pedida y, si no
// lo trae, el texto de la respuesta.
type salidaDelVoto struct {
	IsError          bool           `json:"is_error"`
	Result           string         `json:"result"`
	StructuredOutput jsontext.Value `json:"structured_output"`
}

// dichoDeClase es lo que el juicio de un voto dice de una clase: los campos de
// esquema.json que se leen de ella. Precepto es nil si la clase no lo tiene.
type dichoDeClase struct {
	Motivo    string  `json:"motivo"`
	Respuesta string  `json:"respuesta"`
	Frase     string  `json:"frase"`
	Precepto  *string `json:"precepto"`
}

// leerElVoto lee el juicio de un voto de lo que devolvió el Votante, como
// voto_real de evidencias/adr-0037/guiones/juez.py y, además, contra el esquema
// de la respuesta del juez (contracts/juez-y-voto.md §5 de H24; research D5 de
// H24; FR-007). Devuelve lo que dice de cada clase o, si el voto no llega a
// darse, su causa, la primera que aplica:
//
//  1. el voto agotó su tope;
//  2. la salida no es un objeto JSON: la causa lleva el código del proceso —0
//     si el votante no dio error— y la salida, con el error detrás si lo hay;
//  3. su is_error es verdadero: la causa lleva su result, que es el texto del
//     error, el de un límite de uso entre ellos;
//  4. el juicio —su structured_output o, si no lo trae, su result sin la valla
//     de código que lo envuelva— no es JSON o no cumple el esquema: la causa
//     lleva ese juicio.
//
// Un error del votante que no es el del tope no impide leer la salida: el
// proceso de una sesión que termina con error escribe su objeto y sale con un
// código distinto de 0.
func (v *votacion) leerElVoto(salida []byte, errDelVotante error) (map[string]dichoDeClase, string) {
	if errors.Is(errDelVotante, errTopeDelVoto) {
		return nil, causaDelTope
	}

	var leida salidaDelVoto

	if err := json.Unmarshal(salida, &leida); err != nil {
		return nil, fmt.Sprintf(causaDeLaSalida, codigoDelVoto(errDelVotante),
			textoDeLaCausa(textoDeLaSalida(salida, errDelVotante)))
	}

	if leida.IsError {
		return nil, causaDeLaSesion + textoDeLaCausa(leida.Result)
	}

	juicio := []byte(leida.StructuredOutput)
	if len(juicio) == 0 || string(juicio) == "null" {
		juicio = []byte(sinVallaDeCodigo(leida.Result))
	}

	dichos, err := v.dichosDelJuicio(juicio)
	if err != nil {
		return nil, causaDeLaForma + textoDeLaCausa(string(juicio))
	}

	return dichos, ""
}

// dichosDelJuicio valida el juicio de un voto contra el esquema de la respuesta
// del juez y devuelve lo que dice de cada clase, por su nombre. El error es el
// de un juicio que no es JSON, que no cumple el esquema o del que no se lee lo
// que dice de sus clases.
func (v *votacion) dichosDelJuicio(juicio []byte) (map[string]dichoDeClase, error) {
	valor, err := jsonschema.UnmarshalJSON(bytes.NewReader(juicio))
	if err != nil {
		return nil, fmt.Errorf("el juicio no es JSON: %w", err)
	}

	if err := v.esquema.Validate(valor); err != nil {
		return nil, fmt.Errorf("el juicio no cumple el esquema: %w", err)
	}

	var dichos map[string]dichoDeClase

	if err := json.Unmarshal(juicio, &dichos); err != nil {
		return nil, fmt.Errorf("del juicio no se lee lo que dice de cada clase: %w", err)
	}

	return dichos, nil
}

// sinVallaDeCodigo es el texto de la respuesta del juez sin la valla de código
// que lo envuelva, como lo deja voto_real: sin blancos en los extremos, sin la
// valla que lo abre, con su «json» y los blancos que la sigan, y sin la que lo
// cierra, con los blancos que la precedan.
func sinVallaDeCodigo(texto string) string {
	texto = strings.TrimFunc(texto, esBlanco)

	if abierto, conValla := strings.CutPrefix(texto, vallaDeCodigo); conValla {
		texto = strings.TrimLeftFunc(strings.TrimPrefix(abierto, "json"), esBlanco)
	}

	if cerrado, conValla := strings.CutSuffix(texto, vallaDeCodigo); conValla {
		texto = strings.TrimRightFunc(cerrado, esBlanco)
	}

	return texto
}

// codigoDelVoto es el código del proceso de un voto, el que va en la causa de
// la salida que no es JSON: 0 si el votante no dio error, el de su error si lo
// lleva y codigoSinProceso si no lleva ninguno.
func codigoDelVoto(errDelVotante error) int {
	if errDelVotante == nil {
		return 0
	}

	if conCodigo, loLleva := errors.AsType[errorConCodigo](errDelVotante); loLleva {
		return conCodigo.ExitCode()
	}

	return codigoSinProceso
}

// textoDeLaSalida es lo que la causa de la salida que no es JSON dice de ella:
// la salida y, detrás, el error del votante si lo dio, que es lo que explica
// una salida vacía.
func textoDeLaSalida(salida []byte, errDelVotante error) string {
	texto := strings.TrimFunc(string(salida), esBlanco)
	if errDelVotante == nil {
		return texto
	}

	return texto + " " + errDelVotante.Error()
}

// textoDeLaCausa es el texto que acompaña a la causa de un voto que no llega a
// darse: sin blancos en los extremos, en una línea y cortado a
// caracteresDeLaCausa caracteres, que no son bytes.
func textoDeLaCausa(texto string) string {
	caracteres := []rune(enUnaLinea.Replace(strings.TrimFunc(texto, esBlanco)))
	if len(caracteres) > caracteresDeLaCausa {
		caracteres = caracteres[:caracteresDeLaCausa]
	}

	return string(caracteres)
}
