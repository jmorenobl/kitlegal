package evals

import (
	"cmp"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cli"
)

// ClaseDeConexion es la clase de una conexión atribuida a una invocación
// (data-model §9, tabla de clases).
type ClaseDeConexion string

const (
	// ConexionLocal es la conexión AF_UNIX o a una dirección de bucle local
	// (127.0.0.0/8, ::1): no sale de la máquina.
	ConexionLocal ClaseDeConexion = "local"

	// ConexionBloqueada es la conexión a otra dirección que terminó con un error
	// distinto de EINPROGRESS: no llegó a la red.
	ConexionBloqueada ClaseDeConexion = "bloqueada"

	// ConexionRed es la conexión a otra dirección con resultado 0 o EINPROGRESS,
	// o sin resultado porque el corte interrumpió la llamada o el fin del proceso
	// la dejó sin terminar: nada muestra que no llegara, y se toma como llegada a
	// la red (FR-076).
	ConexionRed ClaseDeConexion = "red"
)

// Las tres familias de un connect que una invocación puede hacer (data-model §9,
// regla 5): cualquier otra, atribuida a una invocación, hace la traza ilegible.
const (
	familiaInet  = "AF_INET"
	familiaInet6 = "AF_INET6"
	familiaUnix  = "AF_UNIX"
)

// errnoEnCurso es el error del connect no bloqueante que sigue conectando: la
// petición ha salido hacia su destino.
const errnoEnCurso = "EINPROGRESS"

// codigoDeMuertePorSenal es el código de la invocación cuyo proceso murió por una
// señal (+++ killed by SIG… +++): un proceso así no tiene código de salida, y es
// el mismo valor con el que lo expresa os.ProcessState.ExitCode. Es distinto de
// 0, así que la invocación no satisface ningún comando esperado.
const codigoDeMuertePorSenal = -1

// codigoDeSalidaMayor es el mayor N de +++ exited with N +++: strace escribe el
// estado de salida del proceso, que va de 0 a 255.
const codigoDeSalidaMayor = 255

// binarioMulticall es el nombre de invocación del binario, que toma el applet de
// su primer argumento (data-model §9, campo applet).
const binarioMulticall = "kitlegal"

// Las dos banderas globales con las que una invocación no consulta: --describe
// emite el esquema sin ejecutar nada y --dry-run describe la operación sin
// realizarla (research.md D12; data-model §6.1).
const (
	banderaDescribe = "describe"
	banderaDryRun   = "dry-run"
)

// Llamadas de la traza, por su nombre en strace.
const (
	llamadaExecve  = "execve"
	llamadaConnect = "connect"

	// llamadaDesconocida es el nombre que strace escribe cuando el hilo murió en
	// la parada de entrada de una llamada que no llegó a identificar y que no se
	// ejecutó (data-model §9, regla 5, forma D): no es ninguna llamada del filtro
	// y no lleva argumentos.
	llamadaDesconocida = "???"
)

// Lo que strace escribe de la llamada que el fin del proceso deja sin terminar
// (data-model §9, regla 5; research.md V65): la marca, delante del paréntesis de
// cierre solo si al decodificador le quedaba algo por escribir a la salida, o al
// final de la línea de entrada que la línea final del hilo deja sin cerrar; y los
// dos resultados, el de la llamada que no llegó a su parada de salida y el de la
// que llegó sin que strace pudiera leer su resultado.
const (
	marcaSinTerminar      = " <unfinished ...>"
	resultadoSinTerminar  = "?"
	resultadoNoDisponible = "? <unavailable>"
)

// Formas de las líneas y de los argumentos que escribe strace -ff con las
// opciones del contrato job-de-evals §3.2, comprobadas en research.md V53, las
// de señal y las que deja un corte en V54, el relleno de alineación en V63 y en
// el runner, y las de la llamada que el fin del proceso deja sin terminar en V65
// y en el runner. Lo que no casa con ellas no se ignora: hace la traza ilegible
// (data-model §9, regla 5).
var (
	// formaDelNombreDeHilo es la del fichero que strace deja por hilo: t.<n>.
	formaDelNombreDeHilo = regexp.MustCompile(`^t\.([0-9]+)$`)

	// formaDeLlamada es la de una llamada con su resultado: 0 o un número, -1
	// ERRNO (descripción) o, sin resultado, ? ERRNO (descripción), la que
	// interrumpió el corte, y ? o ? <unavailable>, la que el fin del proceso dejó
	// sin terminar (formas A y B de la regla 5). Entre el paréntesis de cierre y
	// el igual, strace escribe un espacio y, si la llamada no llega a la columna
	// de alineación (-a 40, su valor por defecto), el relleno de espacios hasta
	// ella: vfork(), la llamada con la que Claude Code de x86_64 crea sus
	// procesos, sale con 33.
	formaDeLlamada = regexp.MustCompile(`^(execve|clone3|clone|vfork|fork|connect)\((.*)\) += ` +
		`(([0-9]+)|-1 ([A-Z][A-Z0-9_]*) \([^()]*\)|\? [A-Z][A-Z0-9_]* \([^()]*\)|\?|\? <unavailable>)$`)

	// formaDeLlamadaSinCerrar es la de la línea de entrada de una llamada que la
	// línea final del hilo deja sin cerrar, sin paréntesis de cierre ni resultado
	// (formas C y D de la regla 5): la llamada, sus argumentos si los escribió y
	// la marca.
	formaDeLlamadaSinCerrar = regexp.MustCompile(`^(execve|clone3|clone|vfork|fork|connect|\?\?\?)\((.*) <unfinished \.\.\.>$`)

	// formaDeSenal es la de una señal entregada, que se admite y no cuenta.
	formaDeSenal = regexp.MustCompile(`^--- SIG[A-Z0-9_]+ \{.*\} ---$`)

	// formaDeSalida y formaDeMuerte son las dos líneas finales de un fichero. El
	// código de la salida se toma con todas sus cifras, para que uno que no es
	// de 0 a 255 sea un defecto que lo nombra y no otra forma de línea.
	formaDeSalida = regexp.MustCompile(`^\+\+\+ exited with ([0-9]+) \+\+\+$`)
	formaDeMuerte = regexp.MustCompile(`^\+\+\+ killed by SIG[A-Z0-9_]+ \+\+\+$`)

	// formaDeExecve es la de los argumentos de execve: la ruta, argv y el
	// entorno, del que strace solo escribe cuántas variables tiene.
	formaDeExecve = regexp.MustCompile(`^"((?:[^"\\]|\\.)*)", \[(.*)\], 0x[0-9a-f]+ /\* [0-9]+ vars? \*/$`)

	// formaDeCadena es la de una cadena entre comillas al principio del texto.
	formaDeCadena = regexp.MustCompile(`^"((?:[^"\\]|\\.)*)"`)

	// formaDeBanderas es la del argumento flags de clone y clone3, en cualquier
	// posición.
	formaDeBanderas = regexp.MustCompile(`(?:^|[{ ])flags=([A-Za-z0-9_|]+)`)

	// formaDeConnect es la de los argumentos de connect: descriptor, dirección
	// con su familia y longitud.
	formaDeConnect = regexp.MustCompile(`^[0-9]+, \{sa_family=([A-Z][A-Z0-9_]*)(.*)\}, [0-9]+$`)

	// formaDeInet, formaDeInet6 y formaDeUnix son las del resto de la dirección
	// en cada una de las tres familias.
	formaDeInet  = regexp.MustCompile(`^, sin_port=htons\(([0-9]+)\), sin_addr=inet_addr\("([0-9.]+)"\)$`)
	formaDeInet6 = regexp.MustCompile(`^, sin6_port=htons\(([0-9]+)\), sin6_flowinfo=htonl\([0-9]+\), ` +
		`inet_pton\(AF_INET6, "([0-9a-fA-F:.]+)", &sin6_addr\), sin6_scope_id=[0-9]+$`)
	formaDeUnix = regexp.MustCompile(`^, sun_path="((?:[^"\\]|\\.)*)"$`)
)

// escapesSimples son las secuencias de una letra con las que strace escribe
// dentro de una cadena las comillas, la barra invertida y los caracteres de
// control con nombre; el resto de octetos no imprimibles van en octal o, con -x,
// en hexadecimal.
var escapesSimples = map[byte]byte{
	'"':  '"',
	'\\': '\\',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'v':  '\v',
}

// Conexion es una llamada connect atribuida a una invocación (data-model §9).
type Conexion struct {
	// Familia es AF_INET, AF_INET6 o AF_UNIX.
	Familia string

	// Direccion es la dirección y el puerto en AF_INET y AF_INET6; el valor cero
	// en AF_UNIX.
	Direccion netip.AddrPort

	// Ruta es la del socket en AF_UNIX; vacía en las otras dos familias.
	Ruta string

	// Resultado es el de la llamada tal como lo escribe strace: 0,
	// -1 EINPROGRESS (Operation now in progress); en la que interrumpió el
	// corte, ? ERESTARTSYS (To be restarted if SA_RESTART is set); y en la que el
	// fin del proceso dejó sin terminar, ?, ? <unavailable> o, en la línea que
	// quedó sin cerrar, vacío.
	Resultado string

	// SinResultado dice si la llamada quedó sin resultado: el corte la
	// interrumpió (data-model §9, regla 6) o el fin del proceso la dejó sin
	// terminar (regla 5).
	SinResultado bool

	// Clase es local, bloqueada o red.
	Clase ClaseDeConexion
}

// Destino es la conexión como la presenta el informe (data-model §9):
// <dirección>:<puerto> en AF_INET, [<dirección>]:<puerto> en AF_INET6 y
// unix:<ruta> en AF_UNIX.
func (c Conexion) Destino() string {
	if c.Familia == familiaUnix {
		return "unix:" + c.Ruta
	}

	return c.Direccion.String()
}

// Invocacion es una invocación de applet registrada en la traza de una sesión
// (data-model §9): lo que se ejecutó de verdad, no lo que el agente escribió.
type Invocacion struct {
	// Proceso es el número del hilo principal del proceso que la ejecutó.
	Proceso int

	// Argv es el de la última execve con resultado 0 del proceso, decodificado.
	Argv []string

	// Applet es el applet registrado que invoca, por el nombre de invocación
	// multicall: scripts/boe y kitlegal boe son el applet boe.
	Applet string

	// Verbo es el primer token que queda tras el applet sin las banderas
	// globales.
	Verbo string

	// Argumentos son los tokens que siguen al verbo sin las banderas globales, con
	// el valor de --timeout y de --asunto.
	Argumentos []string

	// Consulta dice si la invocación consulta el verbo con sus argumentos: falso
	// si lleva --describe o --dry-run, que no consultan, con o sin valor.
	Consulta bool

	// Codigo es el de la línea final del hilo principal del proceso: N en
	// +++ exited with N +++ y -1 en +++ killed by SIG… +++. Es nil en la
	// invocación sin código, cuyo fichero dejó sin línea final el corte de una
	// sesión: no es 0 ni ningún otro número.
	Codigo *int

	// Conexiones son las llamadas connect atribuidas a la invocación, en orden
	// de fichero, por su número, y de línea, también la que quedó sin resultado.
	Conexiones []Conexion
}

// LeerTrazas lee los ficheros t.<n> que strace -ff deja en el traza/ de una
// sesión, uno por hilo, y devuelve las invocaciones de applet ordenadas por el
// número de su proceso, con su código y sus conexiones (data-model §9; contrato
// job-de-evals §4).
//
// cortada dice si el tope cortó la sesión (Sesion.Cortada, que LeerSesion lee de
// codigo-de-la-sesion): solo entonces admite lo que deja el corte, ficheros sin
// línea final o vacíos y una llamada interrumpida, con el resultado
// ? ERRNO (descripción), a la que solo siguen líneas de señal y la línea final
// (regla 6). La llamada que el fin del proceso deja sin terminar, que no la
// produce el corte, se admite en cualquier sesión con la misma condición, y con
// ella el fichero del hilo que una clone, clone3, fork o vfork así pudo crear,
// sin la línea que lo crea y sin ninguna llamada (reglas 1 y 5).
//
// Lo que no entiende no lo ignora, porque un hilo o una conexión sin atribuir
// dejarían en falso la red sin llegadas (FR-076). El error nombra el directorio
// que no se puede leer o en el que ningún fichero sin la línea que lo crea tiene
// llamadas; los ficheros sin esa línea que no pueden quedar sin ella; o el
// fichero, el número de línea y su texto del primer defecto de una línea, que se
// buscan fichero a fichero en orden de número y línea a línea, antes de comprobar
// el origen de cada uno.
func LeerTrazas(dir string, cortada bool) ([]Invocacion, error) {
	interprete, err := nuevoInterprete()
	if err != nil {
		return nil, fmt.Errorf("la traza %s no se puede leer: %w", dir, err)
	}

	hilos, err := leerHilos(dir, cortada)
	if err != nil {
		return nil, err
	}

	atribuida, err := atribuir(dir, hilos)
	if err != nil {
		return nil, err
	}

	return atribuida.invocaciones(interprete)
}

// InterpretarInvocacion lee argv como invocación de un applet del binario que se
// publica (data-model §9; contrato evals-y-grabaciones §6): el applet es
// base(argv[0]) si es un applet registrado o, si es kitlegal, argv[1] si lo es;
// el verbo y los argumentos, los tokens que quedan sin las banderas globales, con
// su valor en las que lo llevan (--x v o --x=v); y no hay consulta si lleva
// --describe o --dry-run. Devuelve falso si argv no invoca ningún applet
// registrado, como kitlegal version.
//
// Las banderas globales y si llevan valor salen de cli.Globales tal como las
// entiende el analizador del binario, y los applets, de app.RegistroDeProduccion:
// ninguna lista paralela. El error queda para lo que impide leerlos.
func InterpretarInvocacion(argv []string) (Invocacion, bool, error) {
	interprete, err := nuevoInterprete()
	if err != nil {
		return Invocacion{}, false, err
	}

	invocacion, deApplet := interprete.interpretar(argv)

	return invocacion, deApplet, nil
}

// interprete lee argv como invocación con el registro y las banderas globales del
// binario que se publica.
type interprete struct {
	registro *app.Registro

	// conValor da, por el nombre largo de cada bandera global, si lleva valor.
	conValor map[string]bool
}

// nuevoInterprete monta el registro de producción y lee las banderas globales del
// modelo de Kong de una gramática que solo embebe cli.Globales, sin la ayuda
// integrada, que no es una bandera global. No analiza ninguna invocación: los
// escritores y la terminación se sustituyen para que ningún camino de la
// biblioteca pueda escribir en los descriptores ni terminar el proceso, como en
// el analizador del kernel.
func nuevoInterprete() (interprete, error) {
	registro, err := app.RegistroDeProduccion()
	if err != nil {
		return interprete{}, fmt.Errorf("el registro de applets del binario no se puede construir: %w", err)
	}

	var gramatica struct{ cli.Globales }

	analizador, err := kong.New(&gramatica,
		kong.NoDefaultHelp(),
		kong.Writers(io.Discard, io.Discard),
		kong.Exit(func(int) {}),
	)
	if err != nil {
		return interprete{}, fmt.Errorf("las banderas globales de cli.Globales no se pueden leer: %w", err)
	}

	conValor := make(map[string]bool, len(analizador.Model.Flags))
	for _, bandera := range analizador.Model.Flags {
		conValor[bandera.Name] = !bandera.IsBool()
	}

	return interprete{registro: registro, conValor: conValor}, nil
}

// interpretar aplica las reglas de InterpretarInvocacion.
func (i interprete) interpretar(argv []string) (Invocacion, bool) {
	applet, resto, deApplet := i.appletDe(argv)
	if !deApplet {
		return Invocacion{}, false
	}

	invocacion := Invocacion{Argv: slices.Clone(argv), Applet: applet, Consulta: true}
	conVerbo := false

	for posicion := 0; posicion < len(resto); posicion++ {
		token := resto[posicion]

		if nombre, conIgual, global := i.banderaGlobal(token); global {
			if i.conValor[nombre] && !conIgual {
				posicion++
			}

			if nombre == banderaDescribe || nombre == banderaDryRun {
				invocacion.Consulta = false
			}

			continue
		}

		if !conVerbo {
			invocacion.Verbo, conVerbo = token, true

			continue
		}

		invocacion.Argumentos = append(invocacion.Argumentos, token)
	}

	return invocacion, true
}

// appletDe es el applet registrado que invoca argv y los tokens que le siguen:
// el del nombre de invocación o, si es kitlegal, el de su primer argumento.
func (i interprete) appletDe(argv []string) (string, []string, bool) {
	if len(argv) == 0 {
		return "", nil, false
	}

	if nombre := path.Base(argv[0]); i.registrado(nombre) {
		return nombre, argv[1:], true
	} else if nombre == binarioMulticall && len(argv) > 1 && i.registrado(argv[1]) {
		return argv[1], argv[2:], true
	}

	return "", nil, false
}

// registrado dice si el applet está en el registro.
func (i interprete) registrado(nombre string) bool {
	_, registrado := i.registro.Buscar(nombre)

	return registrado
}

// banderaGlobal dice si el token es una bandera global, --x o --x=v, con su
// nombre y si lleva el valor pegado.
func (i interprete) banderaGlobal(token string) (nombre string, conIgual, global bool) {
	larga, esLarga := strings.CutPrefix(token, "--")
	if !esLarga {
		return "", false, false
	}

	nombre, _, conIgual = strings.Cut(larga, "=")
	_, global = i.conValor[nombre]

	return nombre, conIgual, global
}

// llamada es una línea execve, clone, clone3, fork, vfork o connect de un
// fichero de traza.
type llamada struct {
	// linea es su número de línea en el fichero, desde 1, y texto, la línea.
	linea int
	texto string

	// nombre es el de la llamada en strace; ??? en la que no llegó a identificar.
	nombre string

	// resultado es el texto tras «= », vacío en la línea que quedó sin cerrar;
	// sinResultado, si la llamada quedó sin él, porque el corte la interrumpió
	// (? ERRNO (descripción)) o porque el fin del proceso la dejó sin terminar;
	// y sinTerminar, si fue por lo segundo (?, ? <unavailable> o sin cerrar).
	resultado    string
	sinResultado bool
	sinTerminar  bool

	// valor es el resultado numérico, si conValor: 0 en la execve o el connect
	// que terminan bien y el número del hilo creado en clone, clone3, fork y vfork.
	valor    int
	conValor bool

	// errno es el error de un resultado -1 ERRNO (descripción).
	errno string

	// argv es el de una execve, decodificado.
	argv []string

	// enHilo dice si una clone o una clone3 lleva CLONE_THREAD: crea un hilo del
	// mismo proceso.
	enHilo bool

	// conexion es la de un connect, sin su clase, que depende de la atribución.
	conexion Conexion
}

// creacion dice si la llamada es de las que crean un hilo o un proceso: clone,
// clone3, fork o vfork.
func (l llamada) creacion() bool {
	switch l.nombre {
	case "clone", "clone3", "fork", "vfork":
		return true
	default:
		return false
	}
}

// creaHilo dice si la llamada crea un hilo o un proceso con número: una creación
// con resultado. La que el fin del proceso dejó sin terminar no crea ninguno que
// la traza pueda atribuir.
func (l llamada) creaHilo() bool {
	return l.creacion() && l.conValor
}

// hilo es lo leído del fichero t.<n> de un hilo.
type hilo struct {
	numero int
	ruta   string

	// llamadas son sus llamadas, en orden de línea.
	llamadas []llamada

	// terminado dice si el fichero termina en su línea final, y codigo es el que
	// da esa línea.
	terminado bool
	codigo    int
}

// leerHilos lee los ficheros del directorio de la traza en orden de número. Una
// entrada que no es un fichero t.<n> no se salta: la traza es ilegible.
func leerHilos(dir string, cortada bool) ([]hilo, error) {
	entradas, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("la traza %s no se puede leer: %w", dir, err)
	}

	hilos := make([]hilo, 0, len(entradas))

	for _, entrada := range entradas {
		ruta := filepath.Join(dir, entrada.Name())

		partes := formaDelNombreDeHilo.FindStringSubmatch(entrada.Name())
		if partes == nil || !entrada.Type().IsRegular() {
			return nil, fmt.Errorf("traza ilegible: %s no es un fichero t.<n> de strace", ruta)
		}

		numero, err := strconv.Atoi(partes[1])
		if err != nil {
			return nil, fmt.Errorf("traza ilegible: %s: el número del hilo no es un entero: %w", ruta, err)
		}

		hilos = append(hilos, hilo{numero: numero, ruta: ruta})
	}

	slices.SortFunc(hilos, func(a, b hilo) int { return cmp.Compare(a.numero, b.numero) })

	for posicion := range hilos {
		if err := leerHilo(&hilos[posicion], cortada); err != nil {
			return nil, err
		}
	}

	return hilos, nil
}

// leerHilo lee el fichero de un hilo línea a línea hasta su final, y devuelve el
// primer defecto de una línea o del final del fichero (data-model §9, reglas 5 y
// 6). Un fichero sin ninguna llamada —solo su línea final, o vacío en una sesión
// cortada— no es un defecto de línea: si además no tiene la línea que lo crea,
// lo juzga la regla 1.
func leerHilo(h *hilo, cortada bool) error {
	contenido, err := leerFichero(h.ruta)
	if err != nil {
		return fmt.Errorf("traza ilegible: %s no se puede leer: %w", h.ruta, err)
	}

	lector := lectorDeHilo{hilo: h, cortada: cortada}
	numero := 0

	for linea := range strings.Lines(string(contenido)) {
		numero++

		if err := lector.leerLinea(numero, linea); err != nil {
			return err
		}
	}

	if !h.terminado && !cortada {
		return fmt.Errorf("traza ilegible: %s: el fichero no termina en su línea final "+
			"(+++ exited with N +++ o +++ killed by SIG… +++), que solo puede faltar si el tope cortó la sesión", h.ruta)
	}

	return nil
}

// lectorDeHilo lleva el estado de la lectura de un fichero de hilo.
type lectorDeHilo struct {
	hilo    *hilo
	cortada bool

	// sinResultado es la llamada sin resultado ya leída —la que interrumpió el
	// corte o la que el fin del proceso dejó sin terminar—, a la que solo pueden
	// seguir líneas de señal y la línea final.
	sinResultado *llamada
}

// leerLinea lee una línea con su salto de línea: una señal, la línea final o una
// llamada. Una línea sin salto está cortada, aunque su texto casara con una forma.
func (l *lectorDeHilo) leerLinea(numero int, linea string) error {
	texto, completa := strings.CutSuffix(linea, "\n")

	switch {
	case !completa:
		return defectoDeLinea(l.hilo.ruta, numero, "la línea no termina en un salto de línea: está cortada", texto)
	case l.hilo.terminado:
		return defectoDeLinea(l.hilo.ruta, numero, "la línea sigue a la línea final del fichero", texto)
	case formaDeSenal.MatchString(texto):
		return nil
	}

	codigo, final, err := leerLineaFinal(texto)
	if err != nil {
		return defectoDeLinea(l.hilo.ruta, numero, err.Error(), texto)
	}

	if final {
		l.hilo.terminado, l.hilo.codigo = true, codigo

		return nil
	}

	if l.sinResultado != nil {
		return defectoDeLinea(l.hilo.ruta, l.sinResultado.linea,
			"a la llamada sin resultado solo pueden seguirla líneas de señal y la línea final", l.sinResultado.texto)
	}

	leida, err := leerLlamada(texto)
	if err != nil {
		return defectoDeLinea(l.hilo.ruta, numero, err.Error(), texto)
	}

	leida.linea = numero

	if leida.sinResultado {
		if !leida.sinTerminar && !l.cortada {
			return defectoDeLinea(l.hilo.ruta, numero, "llamada interrumpida sin resultado en una sesión que el tope "+
				"no cortó: ? ERRNO (descripción) solo lo deja el corte", texto)
		}

		l.sinResultado = &leida
	}

	l.hilo.llamadas = append(l.hilo.llamadas, leida)

	return nil
}

// defectoDeLinea es el error de una línea ilegible: el fichero, el número de
// línea, el motivo y el texto de la línea.
func defectoDeLinea(ruta string, numero int, motivo, texto string) error {
	return fmt.Errorf("traza ilegible: %s, línea %d: %s: «%s»", ruta, numero, motivo, texto)
}

// leerLineaFinal dice si el texto es una línea final y el código que da: el de
// +++ exited with N +++, de 0 a 255, o el de la muerte por señal.
func leerLineaFinal(texto string) (codigo int, final bool, err error) {
	if formaDeMuerte.MatchString(texto) {
		return codigoDeMuertePorSenal, true, nil
	}

	partes := formaDeSalida.FindStringSubmatch(texto)
	if partes == nil {
		return 0, false, nil
	}

	codigo, err = strconv.Atoi(partes[1])
	if err != nil {
		return 0, false, fmt.Errorf("el código de la línea final no es un entero: %w", err)
	}

	if codigo > codigoDeSalidaMayor {
		return 0, false, fmt.Errorf("el código de la línea final, %d, no es un código de salida de 0 a %d",
			codigo, codigoDeSalidaMayor)
	}

	return codigo, true, nil
}

// leerLlamada lee una llamada con su resultado —o sin él, si el corte la
// interrumpió o el fin del proceso la dejó sin terminar— y, según la llamada, su
// argv, si crea un hilo del mismo proceso o su dirección.
func leerLlamada(texto string) (llamada, error) {
	if partes := formaDeLlamadaSinCerrar.FindStringSubmatch(texto); partes != nil {
		return leerLlamadaSinCerrar(texto, partes[1], partes[2])
	}

	partes := formaDeLlamada.FindStringSubmatch(texto)
	if partes == nil {
		return llamada{}, errors.New("no es ninguna de las formas de línea de la traza: execve, clone, clone3, " +
			"fork, vfork o connect con su resultado o sin terminar, ???( <unfinished ...>, una señal o la línea final")
	}

	argumentos, conMarca := strings.CutSuffix(partes[2], marcaSinTerminar)
	if conMarca && partes[3] != resultadoSinTerminar {
		return llamada{}, fmt.Errorf("la marca%s delante del paréntesis de cierre solo cabe con el resultado %s, "+
			"el de la llamada que el fin del proceso deja sin terminar", marcaSinTerminar, resultadoSinTerminar)
	}

	leida := llamada{
		texto:        texto,
		nombre:       partes[1],
		resultado:    partes[3],
		sinResultado: strings.HasPrefix(partes[3], resultadoSinTerminar),
		sinTerminar:  partes[3] == resultadoSinTerminar || partes[3] == resultadoNoDisponible,
		errno:        partes[5],
	}

	if partes[4] != "" {
		valor, err := strconv.Atoi(partes[4])
		if err != nil {
			return llamada{}, fmt.Errorf("el resultado no es un entero: %w", err)
		}

		leida.valor, leida.conValor = valor, true
	}

	if err := leida.leerArgumentos(argumentos); err != nil {
		return llamada{}, err
	}

	return leida, nil
}

// leerLlamadaSinCerrar lee la línea de entrada que la línea final del hilo dejó
// sin cerrar: una llamada del filtro sin resultado (forma C) o ???, la que
// strace no llegó a identificar y que no se ejecutó, sin argumentos (forma D).
func leerLlamadaSinCerrar(texto, nombre, argumentos string) (llamada, error) {
	leida := llamada{texto: texto, nombre: nombre, sinResultado: true, sinTerminar: true}

	if nombre == llamadaDesconocida {
		if argumentos != "" {
			return llamada{}, fmt.Errorf("%s es la llamada que strace no llegó a identificar, y no lleva argumentos",
				llamadaDesconocida)
		}

		return leida, nil
	}

	if err := leida.leerArgumentos(argumentos); err != nil {
		return llamada{}, err
	}

	return leida, nil
}

// leerArgumentos lee de los argumentos lo que la llamada necesita: el argv de
// execve, si clone y clone3 crean un hilo del mismo proceso y la dirección de
// connect, con el resultado de la llamada. La marca de la llamada sin terminar no
// cabe dentro de ellos.
func (l *llamada) leerArgumentos(argumentos string) error {
	if strings.Contains(argumentos, marcaSinTerminar) {
		return fmt.Errorf("la marca%s solo va delante del paréntesis de cierre o al final de la línea sin cerrar",
			marcaSinTerminar)
	}

	var err error

	switch l.nombre {
	case llamadaExecve:
		l.argv, err = leerArgv(argumentos)
	case "clone", "clone3":
		l.enHilo = conBanderaDeHilo(argumentos)
	case llamadaConnect:
		l.conexion, err = leerConexion(argumentos)
		l.conexion.Resultado, l.conexion.SinResultado = l.resultado, l.sinResultado
	}

	return err
}

// leerArgv lee el argv de los argumentos de una execve, decodificado.
func leerArgv(argumentos string) ([]string, error) {
	partes := formaDeExecve.FindStringSubmatch(argumentos)
	if partes == nil {
		return nil, errors.New("los argumentos de execve no tienen la forma de la traza")
	}

	if _, err := decodificarCadena(partes[1]); err != nil {
		return nil, fmt.Errorf("la ruta de execve: %w", err)
	}

	resto := partes[2]
	if resto == "" {
		return []string{}, nil
	}

	var argv []string

	for {
		cadena := formaDeCadena.FindStringSubmatch(resto)
		if cadena == nil {
			return nil, errors.New("el argv de execve no es una lista de cadenas entre comillas")
		}

		decodificada, err := decodificarCadena(cadena[1])
		if err != nil {
			return nil, fmt.Errorf("el argv de execve: %w", err)
		}

		argv = append(argv, decodificada)

		resto = resto[len(cadena[0]):]
		if resto == "" {
			return argv, nil
		}

		var separada bool
		if resto, separada = strings.CutPrefix(resto, ", "); !separada {
			return nil, errors.New("el argv de execve no es una lista de cadenas entre comillas")
		}
	}
}

// decodificarCadena devuelve los octetos de una cadena tal como strace la
// escribe entre comillas: las secuencias de una letra, octal de uno a tres
// dígitos y hexadecimal de dos.
func decodificarCadena(escrita string) (string, error) {
	var decodificada strings.Builder

	for posicion := 0; posicion < len(escrita); posicion++ {
		if escrita[posicion] != '\\' {
			decodificada.WriteByte(escrita[posicion])

			continue
		}

		octeto, consumidos, err := decodificarEscape(escrita[posicion+1:])
		if err != nil {
			return "", err
		}

		decodificada.WriteByte(octeto)

		posicion += consumidos
	}

	return decodificada.String(), nil
}

// decodificarEscape decodifica la secuencia que sigue a una barra invertida y
// dice cuántos octetos de la secuencia ha consumido.
func decodificarEscape(secuencia string) (octeto byte, consumidos int, err error) {
	switch {
	case secuencia == "":
		return 0, 0, errors.New("la cadena termina en una barra invertida")
	case secuencia[0] == 'x':
		return decodificarHexadecimal(secuencia)
	}

	if simple, es := escapesSimples[secuencia[0]]; es {
		return simple, 1, nil
	}

	return decodificarOctal(secuencia)
}

// decodificarHexadecimal decodifica \xHH, la forma de strace con -x: dos dígitos
// hexadecimales tras la x.
func decodificarHexadecimal(secuencia string) (octeto byte, consumidos int, err error) {
	if len(secuencia) < 3 {
		return 0, 0, fmt.Errorf("la secuencia \\%s no tiene dos dígitos hexadecimales", secuencia)
	}

	octetos, err := hex.DecodeString(secuencia[1:3])
	if err != nil {
		return 0, 0, fmt.Errorf("la secuencia \\%s no tiene dos dígitos hexadecimales: %w", secuencia[:3], err)
	}

	return octetos[0], 3, nil
}

// decodificarOctal decodifica \o, \oo y \ooo: strace escribe los dígitos justos,
// y los tres cuando detrás va otro dígito octal, así que se toman hasta tres.
func decodificarOctal(secuencia string) (octeto byte, consumidos int, err error) {
	for consumidos < len(secuencia) && consumidos < 3 && secuencia[consumidos] >= '0' && secuencia[consumidos] <= '7' {
		consumidos++
	}

	switch {
	case consumidos == 0:
		return 0, 0, fmt.Errorf("la secuencia de escape \\%c es desconocida", secuencia[0])
	case consumidos == 3 && secuencia[0] > '3':
		return 0, 0, fmt.Errorf("la secuencia \\%s no cabe en un octeto", secuencia[:3])
	}

	for _, digito := range []byte(secuencia[:consumidos]) {
		octeto = octeto<<3 | (digito - '0')
	}

	return octeto, consumidos, nil
}

// conBanderaDeHilo dice si los argumentos de una clone o una clone3 llevan
// CLONE_THREAD entre sus banderas, en cualquier orden.
func conBanderaDeHilo(argumentos string) bool {
	banderas := formaDeBanderas.FindStringSubmatch(argumentos)

	return banderas != nil && slices.Contains(strings.Split(banderas[1], "|"), "CLONE_THREAD")
}

// leerConexion lee la familia y la dirección de los argumentos de un connect.
// La dirección solo se lee en AF_INET, AF_INET6 y AF_UNIX; una conexión de otra
// familia es un defecto solo si se atribuye a una invocación (data-model §9,
// regla 5).
func leerConexion(argumentos string) (Conexion, error) {
	partes := formaDeConnect.FindStringSubmatch(argumentos)
	if partes == nil {
		return Conexion{}, errors.New("los argumentos de connect no tienen la forma de la traza")
	}

	conexion := Conexion{Familia: partes[1]}
	resto := partes[2]

	var err error

	switch conexion.Familia {
	case familiaInet:
		conexion.Direccion, err = leerDireccion(formaDeInet, resto, "%s:%s")
	case familiaInet6:
		conexion.Direccion, err = leerDireccion(formaDeInet6, resto, "[%s]:%s")
	case familiaUnix:
		conexion.Ruta, err = leerRuta(resto)
	}

	if err != nil {
		return Conexion{}, fmt.Errorf("la dirección %s de connect: %w", conexion.Familia, err)
	}

	return conexion, nil
}

// leerDireccion lee el puerto y la dirección de una familia de internet con su
// forma, y los compone con el formato de la dirección y el puerto.
func leerDireccion(forma *regexp.Regexp, resto, formato string) (netip.AddrPort, error) {
	partes := forma.FindStringSubmatch(resto)
	if partes == nil {
		return netip.AddrPort{}, errors.New("no tiene la forma de la traza")
	}

	direccion, err := netip.ParseAddrPort(fmt.Sprintf(formato, partes[2], partes[1]))
	if err != nil {
		return netip.AddrPort{}, fmt.Errorf("no es una dirección y un puerto: %w", err)
	}

	return direccion, nil
}

// leerRuta lee la ruta de una dirección AF_UNIX, decodificada.
func leerRuta(resto string) (string, error) {
	partes := formaDeUnix.FindStringSubmatch(resto)
	if partes == nil {
		return "", errors.New("no tiene la forma de la traza")
	}

	return decodificarCadena(partes[1])
}

// creacion es la llamada que crea un hilo: el hilo en cuyo fichero está, su
// línea y si crea un hilo del mismo proceso.
type creacion struct {
	creador int
	ruta    string
	linea   int
	enHilo  bool
}

// traza es la de una sesión con cada hilo atribuido a su proceso.
type traza struct {
	dir   string
	hilos []hilo

	// creaciones da, por el número de cada hilo que no es el raíz, la llamada que
	// lo crea.
	creaciones map[int]creacion

	// procesos da, por el número de cada hilo, el del hilo principal de su
	// proceso.
	procesos map[int]int

	// huerfanos son los ficheros sin la línea que los crea y sin ninguna llamada
	// que admite la regla 1: los hilos que una creación sin terminar pudo crear y
	// que el núcleo mató con el proceso antes de ninguna llamada trazada. No
	// pertenecen a ningún proceso.
	huerfanos map[int]bool
}

// atribuir aplica las reglas 1 y 2 de data-model §9: un fichero raíz, el del
// proceso que arrancó strace, sin la línea clone, clone3, fork o vfork que lo
// crea; los ficheros huérfanos que esa regla admite; y cada hilo creado con
// CLONE_THREAD, del proceso del hilo que lo creó, de forma transitiva. Un número
// que crean dos líneas, o un fichero que no desciende del raíz, harían la
// atribución ambigua o incompleta, y la traza es ilegible.
func atribuir(dir string, hilos []hilo) (traza, error) {
	atribuida := traza{
		dir:        dir,
		hilos:      hilos,
		creaciones: map[int]creacion{},
		procesos:   map[int]int{},
		huerfanos:  map[int]bool{},
	}

	if err := atribuida.anotarCreaciones(); err != nil {
		return traza{}, err
	}

	raiz, err := atribuida.raiz()
	if err != nil {
		return traza{}, err
	}

	hijos := map[int][]int{}
	for numero, creada := range atribuida.creaciones {
		hijos[creada.creador] = append(hijos[creada.creador], numero)
	}

	atribuida.procesos[raiz] = raiz

	for pendientes := []int{raiz}; len(pendientes) > 0; pendientes = pendientes[1:] {
		for _, hijo := range hijos[pendientes[0]] {
			atribuida.procesos[hijo] = hijo
			if atribuida.creaciones[hijo].enHilo {
				atribuida.procesos[hijo] = atribuida.procesos[pendientes[0]]
			}

			pendientes = append(pendientes, hijo)
		}
	}

	var sueltos []string

	for _, h := range hilos {
		if _, atribuido := atribuida.procesos[h.numero]; !atribuido && !atribuida.huerfanos[h.numero] {
			sueltos = append(sueltos, h.ruta)
		}
	}

	if len(sueltos) > 0 {
		return traza{}, fmt.Errorf("traza ilegible: %s: %s tienen la línea que los crea pero no descienden de %s, "+
			"el fichero del proceso que arrancó strace", dir, strings.Join(sueltos, ", "), rutaDeHilo(dir, raiz))
	}

	return atribuida, nil
}

// anotarCreaciones anota la llamada que crea cada hilo con fichero en la traza.
func (t *traza) anotarCreaciones() error {
	existentes := make(map[int]bool, len(t.hilos))
	for _, h := range t.hilos {
		existentes[h.numero] = true
	}

	for _, h := range t.hilos {
		for _, creadora := range h.llamadas {
			if !creadora.creaHilo() || !existentes[creadora.valor] {
				continue
			}

			if anterior, repetida := t.creaciones[creadora.valor]; repetida {
				return fmt.Errorf("traza ilegible: %s: dos líneas crean %s: %s, línea %d, y %s, línea %d", t.dir,
					rutaDeHilo(t.dir, creadora.valor), anterior.ruta, anterior.linea, h.ruta, creadora.linea)
			}

			t.creaciones[creadora.valor] = creacion{
				creador: h.numero,
				ruta:    h.ruta,
				linea:   creadora.linea,
				enHilo:  creadora.enHilo,
			}
		}
	}

	return nil
}

// raiz es el número del fichero raíz, el único sin la línea que lo crea que tiene
// alguna llamada, y anota los huérfanos (regla 1): los demás ficheros sin esa
// línea, que la regla admite si no tienen ninguna llamada y alguna clone, clone3,
// fork o vfork de la traza quedó sin terminar, porque son los hilos que esa
// llamada pudo crear, cuya creación strace no vio. Sin una creación así, o con
// alguna llamada, siguen siendo ilegibles con un error que los nombra.
func (t *traza) raiz() (int, error) {
	var raices, huerfanos []hilo

	for _, h := range t.hilos {
		if _, creado := t.creaciones[h.numero]; creado {
			continue
		}

		if len(h.llamadas) > 0 {
			raices = append(raices, h)
		} else {
			huerfanos = append(huerfanos, h)
		}
	}

	switch {
	case len(huerfanos) > 0 && !t.conCreacionSinTerminar():
		return 0, fmt.Errorf("traza ilegible: %s: %s no tienen la línea clone, clone3, fork o vfork que los crea ni "+
			"ninguna llamada, y ninguna clone, clone3, fork o vfork de la traza quedó sin terminar: solo el hilo que "+
			"una creación sin terminar pudo crear puede quedar sin esa línea", t.dir, rutasDe(huerfanos))
	case len(raices) == 0:
		return 0, fmt.Errorf("traza ilegible: %s: no hay ningún fichero sin la línea clone, clone3, fork o vfork "+
			"que lo crea y con alguna llamada, y tiene que haber uno, el del proceso que arrancó strace", t.dir)
	case len(raices) > 1:
		return 0, fmt.Errorf("traza ilegible: %s: %s no tienen la línea clone, clone3, fork o vfork que los crea, "+
			"y solo puede faltarle a uno con llamadas, el del proceso que arrancó strace", t.dir, rutasDe(raices))
	}

	for _, h := range huerfanos {
		t.huerfanos[h.numero] = true
	}

	return raices[0].numero, nil
}

// conCreacionSinTerminar dice si alguna clone, clone3, fork o vfork de la traza
// es una llamada que el fin del proceso dejó sin terminar: pudo crear un hilo del
// que strace no vio la creación.
func (t *traza) conCreacionSinTerminar() bool {
	for _, h := range t.hilos {
		for _, creadora := range h.llamadas {
			if creadora.creacion() && creadora.sinTerminar {
				return true
			}
		}
	}

	return false
}

// rutasDe son las rutas de los ficheros de los hilos, separadas por comas.
func rutasDe(hilos []hilo) string {
	rutas := make([]string, 0, len(hilos))
	for _, h := range hilos {
		rutas = append(rutas, h.ruta)
	}

	return strings.Join(rutas, ", ")
}

// invocaciones aplica las reglas 3 y 4 de data-model §9: cada proceso cuya última
// execve con resultado 0 es de un applet es una invocación, con las conexiones de
// sus hilos posteriores a esa execve; lo demás no se atribuye.
func (t *traza) invocaciones(interprete interprete) ([]Invocacion, error) {
	var invocaciones []Invocacion

	for _, principal := range t.hilos {
		if t.procesos[principal.numero] != principal.numero {
			continue
		}

		ejecucion, ejecutado := ultimaEjecucion(principal)
		if !ejecutado {
			continue
		}

		invocacion, deApplet := interprete.interpretar(ejecucion.argv)
		if !deApplet {
			continue
		}

		invocacion.Proceso = principal.numero

		if principal.terminado {
			codigo := principal.codigo
			invocacion.Codigo = &codigo
		}

		conexiones, err := t.conexionesDe(principal.numero, ejecucion.linea)
		if err != nil {
			return nil, err
		}

		invocacion.Conexiones = conexiones
		invocaciones = append(invocaciones, invocacion)
	}

	return invocaciones, nil
}

// ultimaEjecucion es la última execve con resultado 0 del hilo principal de un
// proceso: si bash se reemplaza por el applet, la del applet.
func ultimaEjecucion(principal hilo) (llamada, bool) {
	for posicion := len(principal.llamadas) - 1; posicion >= 0; posicion-- {
		ejecucion := principal.llamadas[posicion]
		if ejecucion.nombre == llamadaExecve && ejecucion.conValor && ejecucion.valor == 0 {
			return ejecucion, true
		}
	}

	return llamada{}, false
}

// conexionesDe son las llamadas connect del proceso posteriores a su execve, en
// orden de fichero y de línea: las de su hilo principal tras esa línea y las de
// los hilos creados, directa o transitivamente, desde una línea posterior.
func (t *traza) conexionesDe(proceso, lineaDeEjecucion int) ([]Conexion, error) {
	var conexiones []Conexion

	for _, h := range t.hilos {
		if t.procesos[h.numero] != proceso || !t.creadoTrasLaEjecucion(h.numero, proceso, lineaDeEjecucion) {
			continue
		}

		for _, conectada := range h.llamadas {
			if conectada.nombre != llamadaConnect || (h.numero == proceso && conectada.linea < lineaDeEjecucion) {
				continue
			}

			conexion, err := conexionAtribuida(h, conectada)
			if err != nil {
				return nil, err
			}

			conexiones = append(conexiones, conexion)
		}
	}

	return conexiones, nil
}

// creadoTrasLaEjecucion dice si un hilo del proceso lo creó, directa o
// transitivamente, una línea de su hilo principal posterior a la execve del
// applet. El hilo principal cuenta siempre, y de él solo valen las líneas
// posteriores.
func (t *traza) creadoTrasLaEjecucion(numero, proceso, lineaDeEjecucion int) bool {
	for numero != proceso {
		creada := t.creaciones[numero]
		if creada.creador == proceso {
			return creada.linea > lineaDeEjecucion
		}

		numero = creada.creador
	}

	return true
}

// conexionAtribuida es el connect de una invocación con su clase (data-model §9,
// tabla de clases). Una familia que no es AF_INET, AF_INET6 ni AF_UNIX hace la
// traza ilegible.
func conexionAtribuida(h hilo, conectada llamada) (Conexion, error) {
	conexion := conectada.conexion

	switch conexion.Familia {
	case familiaUnix:
		conexion.Clase = ConexionLocal
	case familiaInet, familiaInet6:
		conexion.Clase = claseDeDireccion(conexion.Direccion.Addr(), conectada)
	default:
		return Conexion{}, defectoDeLinea(h.ruta, conectada.linea, fmt.Sprintf(
			"connect de una invocación con la familia %s, que no es AF_INET, AF_INET6 ni AF_UNIX", conexion.Familia),
			conectada.texto)
	}

	return conexion, nil
}

// claseDeDireccion es la clase de un connect a una dirección de internet: local
// en el bucle local; red con resultado 0, EINPROGRESS o sin resultado; bloqueada
// con cualquier otro error.
func claseDeDireccion(direccion netip.Addr, conectada llamada) ClaseDeConexion {
	switch {
	case direccion.IsLoopback():
		return ConexionLocal
	case conectada.sinResultado || conectada.errno == "" || conectada.errno == errnoEnCurso:
		return ConexionRed
	default:
		return ConexionBloqueada
	}
}

// rutaDeHilo es la ruta del fichero de un hilo en el directorio de la traza.
func rutaDeHilo(dir string, numero int) string {
	return filepath.Join(dir, "t."+strconv.Itoa(numero))
}
