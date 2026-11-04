package evals

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// ficheroDelSondeo es el fichero con el que el arnés del laboratorio marca un
// sondeo, con el modelo y la versión de Claude Code de sus sesiones
// (contracts/arnes.md §3 de H39; data-model §6 de H39).
const ficheroDelSondeo = "sondeo.json"

// UsoDeLaTablaDelLaboratorio es la forma de la orden laboratorio tabla
// (contracts/tabla.md §1 de H39), la que dice su línea de uso.
const UsoDeLaTablaDelLaboratorio = "laboratorio tabla <ruta>"

// errUsoDeLaTabla es el fallo de unos argumentos que no son una sola ruta no
// vacía.
var errUsoDeLaTabla = errors.New("uso: " + UsoDeLaTablaDelLaboratorio)

// herramientaRead es la herramienta con la que Claude Code lee un fichero, entero
// o por tramos (research D10 de H39).
const herramientaRead = "Read"

// zDeWilson es el cuantil de la normal del intervalo de Wilson al 95 %
// (FR-041 de H39; research D11 de H39).
const zDeWilson = 1.96

// Lo que escribe el guion (contracts/tabla.md §3 y §4 de H39): el título, la
// tabla, con una fila por variante, modo, modelo y versión, y el detalle, con una
// fila por sesión.
const (
	tituloDeLaTabla   = "# Sondeo de la skill de laboratorio"
	cabeceraDeLaTabla = "| Variante | Modo | Modelo | Claude Code | Sin terminar | n | Activan antes | Entrada antes | " +
		"Las dos antes | Abren la referencia | Turno de la primera apertura | Cumplen la regla |"
	tituloDelDetalle   = "## Detalle"
	cabeceraDelDetalle = "| Sesión | Terminada | Activa antes | Entrada antes | Abre la referencia | Cumple |"
)

// Los textos de las celdas del detalle que no son un sí o un no de siONo
// (contracts/tabla.md §4 de H39), y el motivo de una sesión que LeerSesion da
// por terminada y que no trae respuesta, que tampoco entra en n (FR-042 de H39).
const (
	prefijoSinTerminar = "no: "
	prefijoDelTurno    = "turno "
	motivoSinRespuesta = "sin respuesta"
)

// sondeoLeido es el sondeo.json de un sondeo: el modelo y la versión de Claude
// Code de todas sus sesiones (data-model §6 de H39).
type sondeoLeido struct {
	Modelo     string `json:"modelo"`
	ClaudeCode string `json:"claude_code"`
}

// sesionContada es lo que el guion saca de una sesión (data-model §6 de H39;
// contracts/tabla.md §2 de H39).
type sesionContada struct {
	// ruta es la de su directorio, relativa a la ruta dada y con barras.
	ruta string

	// motivo es por qué no entra en n: el de LeerSesion, o el de una sesión
	// terminada sin respuesta; vacío si entra.
	motivo string

	// activaAntes y entradaAntes dicen si activa la skill de la familia, y la de
	// entrada, antes de la primera herramienta de kitlegal.
	activaAntes, entradaAntes bool

	// aperturas son, por cada referencia de la familia, en su orden, el turno de
	// su primera apertura; 0 si no la abre (data-model §5 de H40).
	aperturas []int

	// cumple dice si la respuesta cumple la regla testigo.
	cumple bool
}

// claveDeFila es lo que define una fila de la tabla: la variante, por su
// posición en la declaración, el modo, por la suya en modosDeLaTabla, el modelo
// y la versión de Claude Code (data-model §7 de H39).
type claveDeFila struct {
	variante, modo  int
	modelo, version string
}

// filaDelSondeo es una fila de la tabla con sus sesiones.
type filaDelSondeo struct {
	claveDeFila

	sesiones []sesionContada
}

// modosDeLaTabla son los modos en el orden de las filas: orden antes que
// herramienta (contracts/tabla.md §3 de H39).
var modosDeLaTabla = []Modo{ModoOrden, ModoHerramienta}

// contadorDelLaboratorio es lo que el guion necesita para contar las sesiones
// de un sondeo: la familia de su tabla, que da las variantes y su orden, las
// skills, las referencias y lo que escribe, las herramientas del registro y el
// intérprete de las invocaciones del binario.
type contadorDelLaboratorio struct {
	familia      familiaDeLaTabla
	herramientas []string
	interprete   interprete

	// filas son las filas por su clave.
	filas map[claveDeFila]*filaDelSondeo
}

// usoDelTranscript es un bloque tool_use del transcript de una sesión con el
// turno de la respuesta que lo emite (research D10 de H39).
type usoDelTranscript struct {
	turno   int
	nombre  string
	entrada json.RawMessage
}

// familiaDeLaTabla es lo que el guion necesita de la familia de un sondeo para
// contar sus sesiones y escribir su tabla (contracts/tabla.md §1 a §4 de H40;
// data-model §1 y §6 de H40): las dos familias cuentan igual, cada una con sus
// skills y sus referencias, y cada una escribe su tabla.
type familiaDeLaTabla struct {
	// variantes son las de la familia, en el orden de su declaración, que es el
	// de las filas.
	variantes []varianteDelLaboratorio

	// ajena es lo que dice el defecto, detrás de su ruta, de un directorio que no
	// es ninguna variante de la familia.
	ajena string

	// skill y entrada son las skills cuya activación se cuenta, por el nombre de
	// su frontmatter; entrada, vacía en una familia sin skill de entrada.
	skill, entrada string

	// referencias son los ficheros de references/ de skill cuya apertura se
	// cuenta, en el orden de las aperturas de cada sesión contada.
	referencias []string

	// titulo, cabeceraDeLaTabla y cabeceraDelDetalle son las líneas con que
	// empiezan la tabla y su detalle.
	titulo, cabeceraDeLaTabla, cabeceraDelDetalle string

	// celdas son las celdas de la familia en una fila con alguna sesión
	// terminada y en el detalle de una sesión terminada.
	celdas celdasDeLaFamilia
}

// celdasDeLaFamilia son las celdas que cada familia escribe detrás de las que
// comparten las dos: las seis primeras de una fila (variante, modo, modelo,
// versión, sin terminar y n) y las dos primeras del detalle (sesión y
// terminada). Una fila sin ninguna sesión terminada y una sesión sin terminar
// llevan vacías las demás, hasta tener tantas celdas como su cabecera.
type celdasDeLaFamilia interface {
	// deLaFila son las celdas de los recuentos de una fila de esa variante con
	// sus sesiones terminadas, que son alguna.
	deLaFila(variante varianteDelLaboratorio, terminadas []sesionContada) []string

	// delDetalle son las celdas de una sesión terminada de esa variante.
	delDetalle(variante varianteDelLaboratorio, sesion sesionContada) []string
}

// laReferenciaDeH39 es la posición, entre las aperturas de una sesión contada
// con la familia de H39, de la de la referencia de su declaración, que es la
// única que cuenta su tabla.
const laReferenciaDeH39 = 0

// tablaDeH39 es la familia de la tabla de la declaración del laboratorio de H39
// (contracts/tabla.md de H39), sin cambios (FR-023 de H40): sus variantes, su
// skill de laboratorio y la de entrada, la referencia de las variantes que la
// cambian, y lo que escribe.
func tablaDeH39(declarada declaracionDelLaboratorio) familiaDeLaTabla {
	return familiaDeLaTabla{
		variantes:          declarada.Variantes,
		ajena:              "no es ninguna variante de la declaración del laboratorio",
		skill:              declarada.Skill,
		entrada:            declarada.Entrada,
		referencias:        []string{declarada.Referencia},
		titulo:             tituloDeLaTabla,
		cabeceraDeLaTabla:  cabeceraDeLaTabla,
		cabeceraDelDetalle: cabeceraDelDetalle,
		celdas:             celdasDeH39{},
	}
}

// tablaDeNormativa es la familia de la tabla de las variantes de normativa
// (contracts/tabla.md §3 y §4 de H40; research D10 de H40): sus variantes, su
// skill, sin skill de entrada, sus ocho referencias, una columna por cada una en
// su orden, y la regla testigo solo en la variante que la lleva.
func tablaDeNormativa() familiaDeLaTabla {
	familia := familiaDeNormativa
	referencias := strings.Join(familia.referencias, " | ")

	return familiaDeLaTabla{
		variantes:   familia.variantes,
		ajena:       "no es ninguna variante de " + familia.skill,
		skill:       familia.skill,
		referencias: familia.referencias,
		titulo:      "# Sondeo de las variantes de " + familia.skill,
		cabeceraDeLaTabla: "| Variante | Modo | Modelo | Claude Code | Sin terminar | n | Activan antes | " +
			referencias + " | Cumplen la regla |",
		cabeceraDelDetalle: "| Sesión | Terminada | Activa antes | " + referencias + " | Cumple |",
		celdas:             celdasDeNormativa{conLaRegla: familia.conLaRegla, referencias: len(familia.referencias)},
	}
}

// familiaDeLosSondeos es la familia de la tabla de esos sondeos (contracts/
// tabla.md §1 de H40; research D9 de H40): la de la primera variante que lee, en
// orden de nombre de los sondeos y, dentro de cada uno, de sus variantes. Si es
// de normativa, la de normativa; si no, o sin ninguna variante, la de la
// declaración del laboratorio del repositorio de esa raíz, que solo se lee
// entonces, con su comportamiento de hoy (FR-023 de H40).
func familiaDeLosSondeos(raiz string, sondeos []string) (familiaDeLaTabla, error) {
	primera, err := primeraVariante(sondeos)
	if err != nil {
		return familiaDeLaTabla{}, err
	}

	if familiaDeNormativa.tiene(primera) {
		return tablaDeNormativa(), nil
	}

	declarada, err := leerLaDeclaracionDelRepositorio(raiz)
	if err != nil {
		return familiaDeLaTabla{}, err
	}

	return tablaDeH39(declarada), nil
}

// primeraVariante es el nombre del primer directorio de variante de los
// sondeos, en orden de nombre de los sondeos y de sus subdirectorios; vacío si
// ninguno tiene ninguno.
func primeraVariante(sondeos []string) (string, error) {
	for _, sondeo := range sondeos {
		variantes, err := subdirectorios(sondeo)
		if err != nil {
			return "", err
		}

		if len(variantes) > 0 {
			return variantes[0], nil
		}
	}

	return "", nil
}

// TablaDelLaboratorio atiende laboratorio tabla con sus argumentos, los que
// siguen a la orden (FR-040 a FR-043 de H39; FR-020 a FR-023 de H40;
// contracts/tabla.md de H39 y de H40; research D9 a D11 de H39 y D9 y D10 de
// H40): lee la ruta dada como un sondeo si tiene sondeo.json o, si no, cada
// subdirectorio que lo tenga, en orden de nombre; cuenta sus sesiones con la
// familia de la primera variante que lee, la de normativa o la de la declaración
// del laboratorio del repositorio de esa raíz; y escribe en salida la tabla y el
// detalle de esa familia. Sin red, modelo ni credencial. Sin ningún sondeo, con
// un directorio de variante que no es de esa familia, o con un sondeo.json, una
// sesión o un transcript que no sabe leer, lo dice con la ruta en una línea de
// la salida de error y devuelve 1, sin escribir nada en salida (defecto
// inesperado, ADR 0023).
//
// El error de escribir esa línea no se propaga, y es una decisión y no un
// silencio: el código ya dice que hubo un defecto, que es lo que lee quien lo
// lanza, y contarlo iría a la misma salida que acaba de fallar.
func TablaDelLaboratorio(raiz string, args []string, salida, errores io.Writer) int {
	tabla, err := tablaDelLaboratorio(raiz, args)
	if err == nil {
		_, err = io.WriteString(salida, tabla)
		if err != nil {
			err = fmt.Errorf("la tabla no se puede escribir: %w", err)
		}
	}

	if err != nil {
		_, _ = fmt.Fprintf(errores, "%s%v\n", PrefijoDelLaboratorio, err)

		return codigoDelDefecto
	}

	return 0
}

// tablaDelLaboratorio es el documento de la tabla de los sondeos de la ruta de
// los argumentos.
func tablaDelLaboratorio(raiz string, args []string) (string, error) {
	if len(args) != 1 || args[0] == "" {
		return "", errUsoDeLaTabla
	}

	ruta := args[0]

	sondeos, err := sondeosDeLaRuta(ruta)
	if err != nil {
		return "", err
	}

	familia, err := familiaDeLosSondeos(raiz, sondeos)
	if err != nil {
		return "", err
	}

	herramientas, err := herramientasDelRegistro()
	if err != nil {
		return "", err
	}

	interprete, err := nuevoInterprete()
	if err != nil {
		return "", err
	}

	contador := contadorDelLaboratorio{
		familia: familia, herramientas: herramientas, interprete: interprete,
		filas: map[claveDeFila]*filaDelSondeo{},
	}

	for _, sondeo := range sondeos {
		if err := contador.contarElSondeo(ruta, sondeo); err != nil {
			return "", err
		}
	}

	return contador.documento(), nil
}

// sondeosDeLaRuta son los sondeos de la ruta: ella misma si tiene sondeo.json
// y, si no, sus subdirectorios que lo tienen, en orden de nombre.
// Sin ninguno, es un error que la nombra.
func sondeosDeLaRuta(ruta string) ([]string, error) {
	esSondeo, err := tieneSondeo(ruta)
	if err != nil {
		return nil, err
	}

	if esSondeo {
		return []string{ruta}, nil
	}

	entradas, err := os.ReadDir(ruta)
	if err != nil {
		return nil, fmt.Errorf("%s no se puede leer: %w", ruta, err)
	}

	var sondeos []string

	for _, entrada := range entradas {
		if !entrada.IsDir() {
			continue
		}

		subdirectorio := filepath.Join(ruta, entrada.Name())

		esSondeo, err := tieneSondeo(subdirectorio)
		if err != nil {
			return nil, err
		}

		if esSondeo {
			sondeos = append(sondeos, subdirectorio)
		}
	}

	if len(sondeos) == 0 {
		return nil, fmt.Errorf("%s no tiene ningún sondeo: ni él ni ninguno de sus subdirectorios tiene %s",
			ruta, ficheroDelSondeo)
	}

	return sondeos, nil
}

// tieneSondeo dice si el directorio tiene sondeo.json. Un sondeo.json que no se
// puede mirar es un error que lo nombra.
func tieneSondeo(directorio string) (bool, error) {
	fichero := filepath.Join(directorio, ficheroDelSondeo)

	_, err := os.Stat(fichero)

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("%s no se puede leer: %w", fichero, err)
	default:
		return true, nil
	}
}

// leerElSondeo lee el sondeo.json del sondeo, el objeto con el modelo y la
// versión que escribe el arnés.
func leerElSondeo(sondeo string) (sondeoLeido, error) {
	fichero := filepath.Join(sondeo, ficheroDelSondeo)

	contenido, err := leerFichero(fichero)
	if err != nil {
		return sondeoLeido{}, fmt.Errorf("%s no se puede leer: %w", fichero, err)
	}

	var leido sondeoLeido

	if err := json.Unmarshal(contenido, &leido); err != nil {
		return sondeoLeido{}, fmt.Errorf("%s no es un objeto con modelo y claude_code: %w", fichero, err)
	}

	return leido, nil
}

// contarElSondeo cuenta las sesiones del sondeo, cada una en
// <variante>/<modo>/<sesión>/: la variante es una de la familia, y el modo, uno
// de los suyos. Lo que no es un directorio no es de ninguna sesión y no se lee.
func (c *contadorDelLaboratorio) contarElSondeo(ruta, sondeo string) error {
	leido, err := leerElSondeo(sondeo)
	if err != nil {
		return err
	}

	variantes, err := subdirectorios(sondeo)
	if err != nil {
		return err
	}

	for _, nombre := range variantes {
		posicion := slices.IndexFunc(c.familia.variantes,
			func(variante varianteDelLaboratorio) bool { return variante.Nombre == nombre })
		if posicion < 0 {
			return fmt.Errorf("%s %s", filepath.Join(sondeo, nombre), c.familia.ajena)
		}

		if err := c.contarLaVariante(ruta, sondeo, leido, posicion); err != nil {
			return err
		}
	}

	return nil
}

// contarLaVariante cuenta las sesiones de cada modo de la variante de esa
// posición en el sondeo.
func (c *contadorDelLaboratorio) contarLaVariante(ruta, sondeo string, leido sondeoLeido, posicion int) error {
	variante := c.familia.variantes[posicion]
	directorio := filepath.Join(sondeo, variante.Nombre)

	modos, err := subdirectorios(directorio)
	if err != nil {
		return err
	}

	for _, nombre := range modos {
		modo := Modo(nombre)
		if !slices.Contains(variante.Modos, modo) {
			return fmt.Errorf("%s no es ningún modo de la variante %s", filepath.Join(directorio, nombre),
				variante.Nombre)
		}

		clave := claveDeFila{
			variante: posicion, modo: slices.Index(modosDeLaTabla, modo),
			modelo: leido.Modelo, version: leido.ClaudeCode,
		}

		sesiones, err := subdirectorios(filepath.Join(directorio, nombre))
		if err != nil {
			return err
		}

		for _, sesion := range sesiones {
			contada, err := c.contarLaSesion(ruta, filepath.Join(directorio, nombre, sesion), modo)
			if err != nil {
				return err
			}

			c.anotar(clave, contada)
		}
	}

	return nil
}

// subdirectorios son los nombres de los subdirectorios del directorio, en orden
// de nombre.
func subdirectorios(directorio string) ([]string, error) {
	entradas, err := os.ReadDir(directorio)
	if err != nil {
		return nil, fmt.Errorf("%s no se puede leer: %w", directorio, err)
	}

	var nombres []string

	for _, entrada := range entradas {
		if entrada.IsDir() {
			nombres = append(nombres, entrada.Name())
		}
	}

	return nombres, nil
}

// anotar pone la sesión en la fila de esa clave.
func (c *contadorDelLaboratorio) anotar(clave claveDeFila, sesion sesionContada) {
	fila, existe := c.filas[clave]
	if !existe {
		fila = &filaDelSondeo{claveDeFila: clave}
		c.filas[clave] = fila
	}

	fila.sesiones = append(fila.sesiones, sesion)
}

// contarLaSesion cuenta la sesión de ese directorio, de ese modo (contracts/
// tabla.md §2 de H39 y de H40): si LeerSesion no la da por terminada o no trae
// respuesta, solo su motivo; si no, de su transcript, si activa cada skill antes
// de la primera herramienta de kitlegal, el turno de la primera apertura de cada
// referencia y, de su respuesta, si cumple la regla testigo.
func (c *contadorDelLaboratorio) contarLaSesion(ruta, directorio string, modo Modo) (sesionContada, error) {
	relativa, err := filepath.Rel(ruta, directorio)
	if err != nil {
		return sesionContada{}, fmt.Errorf("%s no está dentro de %s: %w", directorio, ruta, err)
	}

	contada := sesionContada{ruta: filepath.ToSlash(relativa)}

	sesion, err := LeerSesion(directorio)
	if err != nil {
		return sesionContada{}, fmt.Errorf("la sesión %s no se puede leer: %w", directorio, err)
	}

	switch {
	case !sesion.Terminada:
		contada.motivo = sesion.MotivoSinTerminar

		return contada, nil
	case sesion.Respuesta == "":
		contada.motivo = motivoSinRespuesta

		return contada, nil
	}

	usos, err := leerLosUsos(directorio)
	if err != nil {
		return sesionContada{}, fmt.Errorf("la sesión %s no se puede leer: %w", directorio, err)
	}

	if err := c.anotarLosUsos(&contada, usos, modo); err != nil {
		return sesionContada{}, fmt.Errorf("la sesión %s no se puede leer: %w", directorio, err)
	}

	contada.cumple = reglaTestigo(sesion.Respuesta)

	return contada, nil
}

// anotarLosUsos anota en la sesión lo que dicen sus usos de herramienta: el
// turno en que activa cada skill, el de la primera herramienta de kitlegal del
// modo y el de la primera apertura de cada referencia, cada uno el del primer
// uso que lo es, de modo que una referencia abierta varias veces cuenta una
// (research D10 de H39; contracts/tabla.md §2 de H40).
func (c *contadorDelLaboratorio) anotarLosUsos(contada *sesionContada, usos []usoDelTranscript, modo Modo) error {
	var activa, entrada, primera int

	contada.aperturas = make([]int, len(c.familia.referencias))

	for _, uso := range usos {
		if uso.nombre == herramientaSkill {
			skill, err := skillDelUso(uso)
			if err != nil {
				return err
			}

			activa = primerTurno(activa, uso.turno, skill == c.familia.skill)
			entrada = primerTurno(entrada, uso.turno, c.familia.entrada != "" && skill == c.familia.entrada)
		}

		primera = primerTurno(primera, uso.turno, c.esDeKitlegal(uso, modo))

		for posicion, referencia := range c.familia.referencias {
			contada.aperturas[posicion] = primerTurno(contada.aperturas[posicion], uso.turno,
				c.abreLaReferencia(uso, referencia))
		}
	}

	contada.activaAntes = activaAntes(activa, primera)
	contada.entradaAntes = activaAntes(entrada, primera)

	return nil
}

// primerTurno es el turno anotado o, si todavía no hay ninguno y el uso es de
// los que se buscan, el del uso.
func primerTurno(anotado, turno int, loEs bool) int {
	if anotado == 0 && loEs {
		return turno
	}

	return anotado
}

// activaAntes dice si una skill activada en ese turno lo está antes de la
// primera herramienta de kitlegal, la de ese turno (FR-041 de H39): en un turno
// anterior, o activada y sin ninguna herramienta. En el mismo turno no, porque
// las llamadas de una respuesta se emiten sin haber leído el resultado de las
// otras. El turno 0 es ninguno.
func activaAntes(activa, primera int) bool {
	return activa > 0 && (primera == 0 || activa < primera)
}

// skillDelUso es la input.skill de un uso de Skill, cuya entrada es {skill,
// args?} (research.md V6).
func skillDelUso(uso usoDelTranscript) (string, error) {
	var entrada struct {
		Skill string `json:"skill"`
	}

	if err := json.Unmarshal(uso.entrada, &entrada); err != nil {
		return "", fmt.Errorf("el bloque tool_use de %s del turno %d no tiene la entrada {skill, args?}: %w",
			herramientaSkill, uso.turno, err)
	}

	return entrada.Skill, nil
}

// esDeKitlegal dice si el uso es la llamada a una herramienta de kitlegal del
// modo de la sesión (FR-041 de H39; research D10 de H39): en el modo herramienta,
// un uso cuyo nombre, o lo que sigue a su último «__», es una herramienta del
// registro, de modo que ToolSearch, que carga un esquema, no lo es; en el modo
// orden, una orden de Bash con una orden simple que invoca kitlegal con un verbo
// que da herramienta y que consulta.
func (c *contadorDelLaboratorio) esDeKitlegal(uso usoDelTranscript, modo Modo) bool {
	if modo == ModoHerramienta {
		nombre := uso.nombre
		if corte := strings.LastIndex(nombre, separadorDelPrefijo); corte >= 0 {
			nombre = nombre[corte+len(separadorDelPrefijo):]
		}

		return slices.Contains(c.herramientas, nombre)
	}

	orden, deBash := ordenDelUso(uso)
	if !deBash {
		return false
	}

	for _, palabras := range ordenesSimples(orden) {
		if c.consultaKitlegal(palabras) {
			return true
		}
	}

	return false
}

// consultaKitlegal dice si las palabras de una orden simple invocan kitlegal y
// consultan con un verbo que da herramienta: su primera palabra que no es una
// asignación NOMBRE=valor es kitlegal o una ruta que termina en /kitlegal, y su
// argv, desde ella, es para InterpretarInvocacion de un applet, consulta —sin
// ayuda, --describe ni --dry-run— y su <applet>_<verbo> es una herramienta del
// registro, así que no lo son ni which kitlegal, ni --help, ni skills list.
func (c *contadorDelLaboratorio) consultaKitlegal(palabras []palabraDeSh) bool {
	inicio := slices.IndexFunc(palabras, func(palabra palabraDeSh) bool { return !palabra.asignacion })
	if inicio < 0 {
		return false
	}

	programa := palabras[inicio].texto
	if programa != binarioMulticall && !strings.HasSuffix(programa, "/"+binarioMulticall) {
		return false
	}

	argv := make([]string, 0, len(palabras)-inicio)
	for _, palabra := range palabras[inicio:] {
		argv = append(argv, palabra.texto)
	}

	invocacion, deApplet := c.interprete.interpretar(argv)

	return deApplet && invocacion.Consulta &&
		slices.Contains(c.herramientas, invocacion.Applet+separadorDeHerramienta+invocacion.Verbo)
}

// abreLaReferencia dice si el uso abre esa referencia de la skill de la familia
// (FR-041 de H39; research D10 de H39; contracts/tabla.md §2 de H40): un Read
// cuyo file_path es <skill>/references/<referencia> o termina en
// /<skill>/references/<referencia>, entero o por tramos, o un Bash cuya orden
// contiene references/<referencia>. Grep y Glob no la abren.
func (c *contadorDelLaboratorio) abreLaReferencia(uso usoDelTranscript, referencia string) bool {
	enLaSkill := path.Join(carpetaDeLasReferenciasDeLaSkill, referencia)

	if orden, deBash := ordenDelUso(uso); deBash {
		return strings.Contains(orden, enLaSkill)
	}

	if uso.nombre != herramientaRead {
		return false
	}

	var entrada struct {
		FilePath string `json:"file_path"`
	}

	// Un input sin el file_path como texto —el de una llamada que el modelo dio
	// sin su forma, y que Read rechaza— no abre nada.
	if json.Unmarshal(uso.entrada, &entrada) != nil {
		return false
	}

	fichero := path.Join(c.familia.skill, enLaSkill)

	return entrada.FilePath == fichero || strings.HasSuffix(entrada.FilePath, "/"+fichero)
}

// ordenDelUso es la input.command de un uso de Bash, y si lo es. Un input sin la
// orden como texto —el de una llamada que el modelo dio sin su forma, y que Bash
// rechaza— no es una orden.
func ordenDelUso(uso usoDelTranscript) (string, bool) {
	if uso.nombre != herramientaBash {
		return "", false
	}

	var entrada struct {
		Command string `json:"command"`
	}

	if json.Unmarshal(uso.entrada, &entrada) != nil {
		return "", false
	}

	return entrada.Command, true
}

// leerLosUsos lee del transcript de la sesión sus usos de herramienta, cada uno
// con su turno (research D10 de H39): empieza un turno el primer mensaje
// assistant, y cada assistant que llega después de un user con algún
// tool_result y lleva un message.id distinto del del assistant anterior; así, los
// bloques de una misma respuesta que el transcript da en varios mensajes son del
// mismo turno aunque un resultado llegue entre ellos (research S2 de H39).
//
// LeerSesion ya ha leído el transcript; de él solo se exige aquí el message.id
// de cada assistant, sin el que no hay turno.
func leerLosUsos(directorio string) ([]usoDelTranscript, error) {
	contenido, err := leerFicheroDeSesion(directorio, ficheroDelTranscript)
	if err != nil {
		return nil, err
	}

	var (
		usos     []usoDelTranscript
		turno    int
		anterior string
		resuelto bool
	)

	numero := 0

	for linea := range strings.Lines(string(contenido)) {
		numero++

		var mensaje struct {
			Type    string `json:"type"`
			Message struct {
				ID      string          `json:"id"`
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		}

		if err := json.Unmarshal([]byte(linea), &mensaje); err != nil {
			return nil, fmt.Errorf("%s, línea %d: no es un mensaje JSON: %w", ficheroDelTranscript, numero, err)
		}

		switch mensaje.Type {
		case mensajeAssistant:
			if mensaje.Message.ID == "" {
				return nil, fmt.Errorf("%s, línea %d: el mensaje assistant no tiene message.id",
					ficheroDelTranscript, numero)
			}

			if turno == 0 || (resuelto && mensaje.Message.ID != anterior) {
				turno++
			}

			anterior, resuelto = mensaje.Message.ID, false

			delMensaje, err := usosDelMensaje(mensaje.Message.Content, turno)
			if err != nil {
				return nil, fmt.Errorf("%s, línea %d: %w", ficheroDelTranscript, numero, err)
			}

			usos = append(usos, delMensaje...)
		case mensajeUser:
			resuelto = resuelto || conResultado(mensaje.Message.Content)
		}
	}

	return usos, nil
}

// usosDelMensaje son los bloques tool_use del message.content de un mensaje
// assistant, del turno dado.
func usosDelMensaje(contenido json.RawMessage, turno int) ([]usoDelTranscript, error) {
	var bloques []bloqueDeContenido

	if err := json.Unmarshal(contenido, &bloques); err != nil {
		return nil, fmt.Errorf("el mensaje assistant no tiene la lista message.content: %w", err)
	}

	var usos []usoDelTranscript

	for _, bloque := range bloques {
		if bloque.Type == bloqueToolUse {
			usos = append(usos, usoDelTranscript{turno: turno, nombre: bloque.Name, entrada: bloque.Input})
		}
	}

	return usos, nil
}

// bloqueConTipo es lo que se lee de un bloque del contenido de un mensaje user:
// su type.
type bloqueConTipo struct {
	Type string `json:"type"`
}

// conResultado dice si el message.content de un mensaje user trae algún bloque
// tool_result. Como en LeerSesion, de un mensaje user no se exige más forma que
// la de la API, en la que su contenido es un texto o una lista de bloques: un
// texto no trae ninguno, y no es un defecto.
func conResultado(contenido json.RawMessage) bool {
	var bloques []bloqueConTipo

	if json.Unmarshal(contenido, &bloques) != nil {
		return false
	}

	return slices.ContainsFunc(bloques, func(bloque bloqueConTipo) bool { return bloque.Type == bloqueToolResult })
}

// documento es la tabla y el detalle de las filas contadas, con las cabeceras y
// las celdas de la familia (contracts/tabla.md §3 y §4 de H39 y de H40), en el
// orden de las filas: el de la declaración de la familia, orden antes que
// herramienta, y modelo y versión por orden de texto; y en el detalle, el de las
// filas y, dentro, por ruta.
func (c *contadorDelLaboratorio) documento() string {
	filas := make([]*filaDelSondeo, 0, len(c.filas))
	for _, fila := range c.filas {
		slices.SortFunc(fila.sesiones, func(a, b sesionContada) int { return strings.Compare(a.ruta, b.ruta) })
		filas = append(filas, fila)
	}

	slices.SortFunc(filas, func(a, b *filaDelSondeo) int { return compararClaves(a.claveDeFila, b.claveDeFila) })

	var documento strings.Builder

	escribirLineas(&documento, c.familia.titulo, "", c.familia.cabeceraDeLaTabla,
		separadorDeColumnas(c.familia.cabeceraDeLaTabla))

	for _, fila := range filas {
		escribirLineas(&documento, c.lineaDeLaFila(fila))
	}

	escribirLineas(&documento, "", tituloDelDetalle, "", c.familia.cabeceraDelDetalle,
		separadorDeColumnas(c.familia.cabeceraDelDetalle))

	for _, fila := range filas {
		variante := c.familia.variantes[fila.variante]

		for _, sesion := range fila.sesiones {
			escribirLineas(&documento, c.lineaDelDetalle(variante, sesion))
		}
	}

	return documento.String()
}

// compararClaves ordena dos claves de fila por variante, modo, modelo y versión.
func compararClaves(a, b claveDeFila) int {
	if a.variante != b.variante {
		return a.variante - b.variante
	}

	if a.modo != b.modo {
		return a.modo - b.modo
	}

	if a.modelo != b.modelo {
		return strings.Compare(a.modelo, b.modelo)
	}

	return strings.Compare(a.version, b.version)
}

// lineaDeLaFila es la línea de la tabla de una fila (contracts/tabla.md §3 de
// H39 y de H40): la variante, el modo, el modelo, la versión, las sesiones sin
// terminar y n, y detrás, las celdas de la familia o, con n = 0, todas vacías
// (FR-042 de H39; FR-021 de H40).
func (c *contadorDelLaboratorio) lineaDeLaFila(fila *filaDelSondeo) string {
	variante := c.familia.variantes[fila.variante]

	var terminadas []sesionContada

	for _, sesion := range fila.sesiones {
		if sesion.motivo == "" {
			terminadas = append(terminadas, sesion)
		}
	}

	n := len(terminadas)
	celdas := []string{
		variante.Nombre, string(modosDeLaTabla[fila.modo]), fila.modelo, fila.version,
		strconv.Itoa(len(fila.sesiones) - n), strconv.Itoa(n),
	}

	if n == 0 {
		return lineaDeCeldas(conLasVacias(celdas, c.familia.cabeceraDeLaTabla)...)
	}

	return lineaDeCeldas(append(celdas, c.familia.celdas.deLaFila(variante, terminadas)...)...)
}

// conLasVacias son las celdas seguidas de las vacías que les faltan para tener
// tantas como esa cabecera.
func conLasVacias(celdas []string, cabecera string) []string {
	return append(celdas, make([]string, max(celdasDe(cabecera)-len(celdas), 0))...)
}

// celdasDeH39 son las celdas de la tabla de la declaración del laboratorio de
// H39 (contracts/tabla.md §3 y §4 de H39), sin cambios (FR-023 de H40).
type celdasDeH39 struct{}

// deLaFila son las de los recuentos de una fila de H39 con alguna sesión
// terminada. Están vacías las celdas de la entrada en una variante que no la
// cambia, las de la referencia en una que no la cambia y la del turno si ninguna
// sesión la abre.
func (celdasDeH39) deLaFila(variante varianteDelLaboratorio, terminadas []sesionContada) []string {
	conEntrada := slices.Contains(variante.Cambia, parteEntrada)
	conReferencia := slices.Contains(variante.Cambia, parteReferencia)

	cuantas := func(loEs func(sesionContada) bool) string {
		return recuento(cuantasLoSon(terminadas, loEs), len(terminadas))
	}

	return []string{
		cuantas(func(s sesionContada) bool { return s.activaAntes }),
		siLaCambia(conEntrada, func() string { return cuantas(func(s sesionContada) bool { return s.entradaAntes }) }),
		siLaCambia(conEntrada, func() string {
			return cuantas(func(s sesionContada) bool { return s.activaAntes && s.entradaAntes })
		}),
		siLaCambia(conReferencia, func() string {
			return cuantas(func(s sesionContada) bool { return s.aperturas[laReferenciaDeH39] > 0 })
		}),
		siLaCambia(conReferencia, func() string { return turnosDeLaApertura(terminadas, laReferenciaDeH39) }),
		cuantas(func(s sesionContada) bool { return s.cumple }),
	}
}

// delDetalle son las de una sesión terminada de H39: si activa antes, si activa
// antes la de entrada en una variante que la cambia, si abre la referencia y en
// qué turno en una que la cambia, y si cumple la regla.
func (celdasDeH39) delDetalle(variante varianteDelLaboratorio, sesion sesionContada) []string {
	return []string{
		siONo(sesion.activaAntes),
		siLaCambia(slices.Contains(variante.Cambia, parteEntrada), func() string { return siONo(sesion.entradaAntes) }),
		siLaCambia(slices.Contains(variante.Cambia, parteReferencia), func() string {
			return celdaDeLaApertura(sesion.aperturas[laReferenciaDeH39])
		}),
		siONo(sesion.cumple),
	}
}

// celdasDeNormativa son las celdas de la tabla de las variantes de normativa
// (contracts/tabla.md §3 y §4 de H40; research D10 de H40): «Activan antes», una
// por referencia y la regla testigo, solo en la variante que la lleva.
type celdasDeNormativa struct {
	// conLaRegla es la variante en cuya fila y cuyo detalle se escribe la regla
	// testigo; en las demás, la celda va vacía, no a cero.
	conLaRegla string

	// referencias es cuántas referencias tiene la familia, las de las aperturas
	// de cada sesión contada.
	referencias int
}

// deLaFila son las de los recuentos de una fila de normativa con alguna sesión
// terminada: cuántas activan antes; por cada referencia, cuántas la abren y, si
// alguna, « · turno <mediana> (<mínimo>–<máximo>)» de la primera apertura de
// las que la abren; y cuántas cumplen la regla, en la variante que la lleva.
func (c celdasDeNormativa) deLaFila(variante varianteDelLaboratorio, terminadas []sesionContada) []string {
	n := len(terminadas)
	celdas := make([]string, 0, c.referencias+2)
	celdas = append(celdas, recuento(cuantasLoSon(terminadas, func(s sesionContada) bool { return s.activaAntes }), n))

	for referencia := range c.referencias {
		celda := recuento(cuantasLoSon(terminadas, func(s sesionContada) bool { return s.aperturas[referencia] > 0 }), n)
		if turnos := turnosDeLaApertura(terminadas, referencia); turnos != "" {
			celda += " · " + prefijoDelTurno + turnos
		}

		celdas = append(celdas, celda)
	}

	regla := ""
	if variante.Nombre == c.conLaRegla {
		regla = recuento(cuantasLoSon(terminadas, func(s sesionContada) bool { return s.cumple }), n)
	}

	return append(celdas, regla)
}

// delDetalle son las de una sesión terminada de normativa: si activa antes; por
// cada referencia, «no» o el turno de su primera apertura; y si cumple la regla,
// en la variante que la lleva.
func (c celdasDeNormativa) delDetalle(variante varianteDelLaboratorio, sesion sesionContada) []string {
	celdas := make([]string, 0, len(sesion.aperturas)+2)
	celdas = append(celdas, siONo(sesion.activaAntes))

	for _, turno := range sesion.aperturas {
		celdas = append(celdas, celdaDeLaApertura(turno))
	}

	cumple := ""
	if variante.Nombre == c.conLaRegla {
		cumple = siONo(sesion.cumple)
	}

	return append(celdas, cumple)
}

// celdaDeLaApertura es la celda del detalle de una referencia: «no» si la
// sesión no la abre y, si la abre, «turno <t>» con el de su primera apertura.
func celdaDeLaApertura(turno int) string {
	if turno == 0 {
		return siONo(false)
	}

	return prefijoDelTurno + strconv.Itoa(turno)
}

// cuantasLoSon es cuántas de las sesiones lo son.
func cuantasLoSon(sesiones []sesionContada, loEs func(sesionContada) bool) int {
	cuantas := 0

	for _, sesion := range sesiones {
		if loEs(sesion) {
			cuantas++
		}
	}

	return cuantas
}

// siLaCambia es la celda que da celda si la variante cambia la parte y, si no,
// la vacía.
func siLaCambia(cambia bool, celda func() string) string {
	if !cambia {
		return ""
	}

	return celda()
}

// recuento es la celda de k de n: «<k> de <n> · <p> % [<lo>–<hi> %]», con un
// decimal y coma decimal, y el intervalo de Wilson al 95 % (FR-041 de H39;
// research D11 de H39).
func recuento(k, n int) string {
	inferior, superior := intervaloDeWilson(k, n)

	return fmt.Sprintf("%d de %d · %s %% [%s–%s %%]", k, n,
		conUnDecimal(100*float64(k)/float64(n)), conUnDecimal(inferior), conUnDecimal(superior))
}

// intervaloDeWilson son los extremos, en porcentaje, del intervalo de Wilson al
// 95 % (z = 1,96) de k de n, cada uno recortado a [0, 100] antes de redondear:
// sin recortar, el inferior de 0 de n sale como un negativo de redondeo.
func intervaloDeWilson(k, n int) (float64, float64) {
	total := float64(n)
	observada := float64(k) / total
	cuadrado := zDeWilson * zDeWilson
	divisor := 1 + cuadrado/total
	centro := (observada + cuadrado/(2*total)) / divisor
	radio := zDeWilson * math.Sqrt(observada*(1-observada)/total+cuadrado/(4*total*total)) / divisor

	return 100 * min(max(centro-radio, 0), 1), 100 * min(max(centro+radio, 0), 1)
}

// conUnDecimal escribe un porcentaje con un decimal y coma decimal.
func conUnDecimal(valor float64) string {
	return strings.Replace(strconv.FormatFloat(valor, 'f', 1, 64), ".", ",", 1)
}

// turnosDeLaApertura es la celda del turno de la primera apertura de la
// referencia de esa posición en las sesiones que la abren: «<mediana>
// (<mínimo>–<máximo>)», con la mediana entera o con «,5» si cae entre dos
// turnos; vacía si ninguna la abre.
func turnosDeLaApertura(sesiones []sesionContada, referencia int) string {
	var turnos []int

	for _, sesion := range sesiones {
		if turno := sesion.aperturas[referencia]; turno > 0 {
			turnos = append(turnos, turno)
		}
	}

	if len(turnos) == 0 {
		return ""
	}

	slices.Sort(turnos)

	mitad := len(turnos) / 2
	doble := 2 * turnos[mitad]

	if len(turnos)%2 == 0 {
		doble = turnos[mitad-1] + turnos[mitad]
	}

	mediana := strconv.Itoa(doble / 2)
	if doble%2 != 0 {
		mediana += ",5"
	}

	return fmt.Sprintf("%s (%d–%d)", mediana, turnos[0], turnos[len(turnos)-1])
}

// lineaDelDetalle es la línea del detalle de una sesión de esa variante
// (contracts/tabla.md §4 de H39 y de H40): sin terminar, su ruta, su motivo y
// las demás celdas vacías; terminada, su ruta, un sí y las celdas de la familia.
func (c *contadorDelLaboratorio) lineaDelDetalle(variante varianteDelLaboratorio, sesion sesionContada) string {
	if sesion.motivo != "" {
		return lineaDeCeldas(conLasVacias([]string{sesion.ruta, prefijoSinTerminar + sesion.motivo},
			c.familia.cabeceraDelDetalle)...)
	}

	return lineaDeCeldas(append([]string{sesion.ruta, siONo(true)}, c.familia.celdas.delDetalle(variante, sesion)...)...)
}

// lineaDeCeldas es una línea de una tabla de Markdown con esas celdas. Una
// celda no puede llevar una barra vertical, que la partiría, ni un salto de
// línea, que partiría la fila: la barra va escapada y el salto, como un espacio.
// Solo el motivo de una sesión sin terminar, que puede traer el texto de un
// error, llega a llevarlos.
func lineaDeCeldas(celdas ...string) string {
	escapadas := make([]string, 0, len(celdas))
	for _, celda := range celdas {
		escapadas = append(escapadas, escaparLaCelda.Replace(celda))
	}

	return "| " + strings.Join(escapadas, " | ") + " |"
}

// escaparLaCelda escapa la barra vertical de una celda y pone un espacio en
// lugar de cada salto de línea.
var escaparLaCelda = strings.NewReplacer("|", `\|`, "\r\n", " ", "\n", " ", "\r", " ")

// separadorDeColumnas es la línea que separa la cabecera de una tabla de
// Markdown de sus filas, con una columna por cada una de la cabecera.
func separadorDeColumnas(cabecera string) string {
	return "|" + strings.Repeat("---|", celdasDe(cabecera))
}

// celdasDe es cuántas celdas tiene la cabecera de una tabla de Markdown, en la
// que ninguna lleva una barra vertical.
func celdasDe(cabecera string) int {
	return strings.Count(cabecera, "|") - 1
}

// escribirLineas escribe cada línea seguida de su salto.
func escribirLineas(documento *strings.Builder, lineas ...string) {
	for _, linea := range lineas {
		documento.WriteString(linea)
		documento.WriteByte('\n')
	}
}

// palabraDeSh es una palabra de una orden simple de sh, tras quitarle las
// comillas y las barras invertidas, y si es una asignación NOMBRE=valor: su
// principio, hasta el primer «=», es un nombre de sh sin comillas ni barras.
type palabraDeSh struct {
	texto      string
	asignacion bool
}

// ordenesSimples parte una orden de Bash en sus órdenes simples, cada una en sus
// palabras, con las reglas de sh que el guion necesita (research D10 de H39): la
// separan |, ||, &&, ;, & y los saltos de línea, salvo el & de una redirección
// como 2>&1; las palabras, los espacios y los tabuladores; y las comillas
// simples y dobles y la barra invertida se quitan como en sh, con la barra
// invertida seguida de un salto de línea como continuación. Una subshell, una
// sustitución de orden —$( ) o comillas invertidas— y lo que va dentro son parte
// de la palabra en que están, sin partir: el guion no cuenta kitlegal dentro de
// ellas (research D10 de H39, alternativa c). Lo demás de sh no se distingue: un
// # es un carácter más, y unas comillas, un paréntesis o unas comillas
// invertidas sin cerrar llegan hasta el final de la orden.
func ordenesSimples(orden string) [][]palabraDeSh {
	lector := lectorDeSh{}

	for posicion := 0; posicion < len(orden); {
		posicion = lector.leerDesde(orden, posicion)
	}

	lector.cerrarLaOrden()

	return lector.ordenes
}

// estadoDeAsignacion es lo que se sabe, mientras se lee una palabra, de si es
// una asignación.
type estadoDeAsignacion int

const (
	// sinLetras: la palabra no tiene todavía nada.
	sinLetras estadoDeAsignacion = iota
	// enElNombre: todo lo leído son letras, cifras o _ sin comillas, y no empieza
	// por una cifra.
	enElNombre
	// esAsignacion: tras el nombre vino un = sin comillas.
	esAsignacion
	// noEsAsignacion: cualquier otra cosa.
	noEsAsignacion
)

// lectorDeSh es el estado de ordenesSimples mientras lee una orden.
type lectorDeSh struct {
	// ordenes son las órdenes simples ya cerradas que tienen alguna palabra.
	ordenes [][]palabraDeSh

	// actual son las palabras de la orden simple que se está leyendo.
	actual []palabraDeSh

	// palabra es la palabra que se está leyendo, si enLaPalabra.
	palabra     strings.Builder
	enLaPalabra bool
	asignacion  estadoDeAsignacion

	// redireccion dice si lo último de la palabra es un > o un < sin comillas,
	// tras el que & es parte de la redirección.
	redireccion bool
}

// leerDesde lee lo que empieza en esa posición de la orden y devuelve la
// posición siguiente.
func (l *lectorDeSh) leerDesde(orden string, posicion int) int {
	switch caracter := orden[posicion]; caracter {
	case '\\':
		return l.leerLaBarra(orden, posicion)
	case '\'', '"', '`', '(':
		return l.leerElTramo(orden, posicion)
	case ' ', '\t':
		l.cerrarLaPalabra()
	case '\n', ';', '|':
		l.cerrarLaOrden()
	case '&':
		if !l.redireccion {
			l.cerrarLaOrden()

			break
		}

		l.anadirSinComillas(caracter)
	default:
		l.anadirSinComillas(caracter)
	}

	return posicion + 1
}

// leerLaBarra lee la barra invertida de esa posición, fuera de comillas: con un
// salto de línea detrás, una continuación, que no es nada; con otro carácter, ese
// carácter tal cual; y al final de la orden, ella misma.
func (l *lectorDeSh) leerLaBarra(orden string, posicion int) int {
	if posicion+1 >= len(orden) {
		l.anadirSinComillas('\\')

		return posicion + 1
	}

	if orden[posicion+1] != '\n' {
		l.anadirEntrecomillado(orden[posicion+1 : posicion+2])
	}

	return posicion + 2
}

// leerElTramo lee el tramo entre comillas, entre paréntesis o entre comillas
// invertidas que empieza en esa posición y devuelve la posición que sigue a su
// cierre: de unas comillas simples, su contenido tal cual; de unas dobles, su
// contenido sin las barras invertidas que escapan; de lo demás, el tramo entero,
// sin partir. Sin cierre, el tramo llega hasta el final de la orden.
func (l *lectorDeSh) leerElTramo(orden string, posicion int) int {
	cierre := cierreDelTramo(orden, posicion)
	siguiente := min(cierre+1, len(orden))

	switch orden[posicion] {
	case '\'':
		l.anadirEntrecomillado(orden[posicion+1 : cierre])
	case '"':
		l.anadirEntrecomillado(sinEscapesDeLasDobles(orden[posicion+1 : cierre]))
	default:
		l.anadirEntrecomillado(orden[posicion:siguiente])
	}

	return siguiente
}

// cierreDelTramo es la posición del carácter que cierra el tramo que empieza en
// esa posición, o el final de la orden si no lo cierra ninguno: la comilla
// simple siguiente; la doble o la invertida siguiente que no escapa una barra; o
// el paréntesis que cierra el de apertura.
func cierreDelTramo(orden string, posicion int) int {
	switch orden[posicion] {
	case '\'':
		if cierre := strings.IndexByte(orden[posicion+1:], '\''); cierre >= 0 {
			return posicion + 1 + cierre
		}

		return len(orden)
	case '(':
		return cierreDelParentesis(orden, posicion)
	default:
		return cierreSinEscapar(orden, posicion)
	}
}

// cierreSinEscapar es la posición del primer carácter igual al de esa posición
// que no escapa una barra invertida, el cierre de unas comillas dobles o
// invertidas, o el final de la orden.
func cierreSinEscapar(orden string, posicion int) int {
	for actual := posicion + 1; actual < len(orden); actual++ {
		switch orden[actual] {
		case '\\':
			actual++
		case orden[posicion]:
			return actual
		}
	}

	return len(orden)
}

// cierreDelParentesis es la posición del paréntesis que cierra el de esa
// posición, o el final de la orden, saltando los tramos entre comillas que lleva
// dentro, que pueden llevar paréntesis sin pareja.
func cierreDelParentesis(orden string, posicion int) int {
	profundidad := 0

	for actual := posicion; actual < len(orden); actual++ {
		switch orden[actual] {
		case '\\':
			actual++
		case '\'', '"', '`':
			actual = cierreDelTramo(orden, actual)
		case '(':
			profundidad++
		case ')':
			profundidad--
			if profundidad == 0 {
				return actual
			}
		}
	}

	return len(orden)
}

// sinEscapesDeLasDobles quita del contenido de unas comillas dobles las barras
// invertidas que escapan en ellas: las de $, `, ", \ y el salto de línea, que
// con la suya desaparece. Las demás barras se quedan.
func sinEscapesDeLasDobles(contenido string) string {
	var texto strings.Builder

	for posicion := 0; posicion < len(contenido); posicion++ {
		if contenido[posicion] == '\\' && posicion+1 < len(contenido) &&
			strings.IndexByte("$`\"\\\n", contenido[posicion+1]) >= 0 {
			posicion++

			if contenido[posicion] == '\n' {
				continue
			}
		}

		texto.WriteByte(contenido[posicion])
	}

	return texto.String()
}

// anadirSinComillas añade a la palabra un carácter sin comillas.
func (l *lectorDeSh) anadirSinComillas(caracter byte) {
	l.empezarLaPalabra()
	l.palabra.WriteByte(caracter)
	l.redireccion = caracter == '>' || caracter == '<'
	l.asignacion = asignacionTras(l.asignacion, caracter)
}

// asignacionTras es lo que se sabe de si una palabra es una asignación tras
// añadirle, sin comillas, ese carácter.
func asignacionTras(estado estadoDeAsignacion, caracter byte) estadoDeAsignacion {
	switch {
	case estado == esAsignacion:
		return esAsignacion
	case estado == sinLetras && esDeUnNombreDeSh(caracter, true), estado == enElNombre && esDeUnNombreDeSh(caracter, false):
		return enElNombre
	case estado == enElNombre && caracter == '=':
		return esAsignacion
	default:
		return noEsAsignacion
	}
}

// esDeUnNombreDeSh dice si el carácter puede ir en un nombre de sh: una letra
// sin tilde, el _ y, salvo en el primero, una cifra.
func esDeUnNombreDeSh(caracter byte, primero bool) bool {
	switch {
	case caracter == '_', 'a' <= caracter && caracter <= 'z', 'A' <= caracter && caracter <= 'Z':
		return true
	default:
		return !primero && '0' <= caracter && caracter <= '9'
	}
}

// anadirEntrecomillado añade a la palabra un texto entre comillas o escapado,
// que ya no puede ser parte del nombre de una asignación.
func (l *lectorDeSh) anadirEntrecomillado(texto string) {
	l.empezarLaPalabra()
	l.palabra.WriteString(texto)
	l.redireccion = false

	if l.asignacion != esAsignacion {
		l.asignacion = noEsAsignacion
	}
}

// empezarLaPalabra marca que hay una palabra, aunque quede vacía, como la de
// unas comillas sin nada.
func (l *lectorDeSh) empezarLaPalabra() {
	l.enLaPalabra = true
}

// cerrarLaPalabra añade a la orden simple la palabra que se estaba leyendo, si
// la hay.
func (l *lectorDeSh) cerrarLaPalabra() {
	if l.enLaPalabra {
		l.actual = append(l.actual, palabraDeSh{texto: l.palabra.String(), asignacion: l.asignacion == esAsignacion})
	}

	l.palabra.Reset()
	l.enLaPalabra = false
	l.asignacion = sinLetras
	l.redireccion = false
}

// cerrarLaOrden cierra la palabra y la orden simple que se estaban leyendo; una
// orden sin palabras, como la que queda entre los dos | de ||, no se anota.
func (l *lectorDeSh) cerrarLaOrden() {
	l.cerrarLaPalabra()

	if len(l.actual) > 0 {
		l.ordenes = append(l.ordenes, l.actual)
	}

	l.actual = nil
}
