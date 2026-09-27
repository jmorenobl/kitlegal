package app

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/cli"
)

// Destino es quién atiende una invocación una vez resuelto el despacho. Son tres
// y se excluyen entre sí: un applet, un verbo reservado del propio binario o la
// ayuda, que ya viene escrita.
type Destino string

const (
	// DestinoApplet entrega la invocación a un applet del registro, con el verbo
	// ya normalizado en los argumentos.
	DestinoApplet Destino = "applet"
	// DestinoReservado es el verbo que atiende el binario por su cuenta
	// —«version»—, reconocido antes de mirar el registro.
	DestinoReservado Destino = "reservado"
	// DestinoAyuda es la ayuda derivada del registro, la del binario o la de un
	// applet, ya escrita en Ayuda y lista para el presentador.
	DestinoAyuda Destino = "ayuda"
)

// sufijoEjecutable es lo que windows añade al nombre del ejecutable y que el
// despacho retira, de modo que el mismo código valga en los tres sistemas
// (plan.md, «Target Platform»).
const sufijoEjecutable = ".exe"

// terminadorDeBanderas separa las banderas de los argumentos. El despacho solo
// lo mira para una cosa: lo que va después no nombra ningún verbo, que es la
// salida que el contrato ofrece a quien necesite pasar como argumento una
// palabra que se llama igual que un verbo
// (contracts/registro-y-describe.md §2 bis).
//
// Lo declara también internal/cli para su pre-escaneo. Son dos lecturas
// independientes de la misma convención de Unix, no una compartida: acoplar los
// dos paquetes por una constante de dos caracteres costaría más de lo que ahorra.
const terminadorDeBanderas = "--"

// Despacho es lo que queda decidido antes de que exista gramática que analizar:
// qué atiende la invocación y con qué argumentos.
type Despacho struct {
	// Destino es quién atiende la invocación.
	Destino Destino
	// Applet es el applet resuelto: el que va a ejecutarse y también el que
	// responde su propia ayuda. Es nulo cuando la invocación no resolvió
	// ninguno —un verbo reservado o la ayuda del binario—.
	Applet Applet
	// Reservado es el verbo del binario que atiende la invocación, y va vacío
	// en cualquier otro destino.
	Reservado string
	// Ayuda es el texto ya derivado del registro que hay que escribir, y va
	// vacío en cualquier otro destino.
	Ayuda string
	// Args son los argumentos que quedan tras resolver el applet, con el verbo
	// por omisión ya insertado si tocaba: es lo que recibe la gramática. Cuando
	// lo que se pidió es la ayuda llegan tal cual, sin nada insertado, porque
	// pedir la ayuda de un applet no resuelve ningún verbo. En un verbo
	// reservado van vacíos: no admite ninguno.
	Args []string
}

// Despachar decide quién atiende la invocación y con qué argumentos, con la
// precedencia fija de contracts/registro-y-describe.md §2 y la normalización del
// verbo de §2 bis, **antes** de que se construya ninguna gramática.
//
// argv es la lista entera, con el nombre del programa a la cabeza. previo es lo
// que el pre-escaneo sabe de la invocación, del que aquí solo importa si se pidió
// ayuda: eso suprime la normalización del verbo por omisión (D25, D26).
//
// La lista que recibe no se altera: la normalización construye otra. Lo que
// devuelve como error es siempre un error de argumentos —código 2— con el nombre
// desconocido y la lista que el registro ofrece; nunca un pánico, tampoco con una
// lista vacía (FR-006, FR-033).
func Despachar(registro *Registro, argv []string, previo cli.Preliminar) (Despacho, error) {
	invocacion := nombreDeInvocacion(argv)
	resto := argumentosDe(argv)

	// 1. Si el nombre de invocación está registrado, manda él y los argumentos
	// se entregan íntegros al applet: «./echo boe» ejecuta «echo» con el
	// argumento «boe», y no el applet «boe» (FR-002, FR-004).
	if applet, registrado := registro.Buscar(invocacion); registrado {
		return despachoDeApplet(applet, resto, previo)
	}

	// 2. Si no lo está —incluido «kitlegal»—, el applet sale del primer
	// argumento. Los verbos reservados se reconocen **antes** que el registro,
	// que es justo lo que hace inalcanzable a un applet llamado como uno de
	// ellos y por lo que el registro los rechaza al construirse (FR-003, D17);
	// y --version es «version» con la forma de una bandera.
	if len(resto) > 0 {
		if slices.Contains(verbosReservados, resto[0]) || resto[0] == banderaDeVersion {
			return despachoReservado(resto)
		}

		if applet, registrado := registro.Buscar(resto[0]); registrado {
			return despachoDeApplet(applet, resto[1:], previo)
		}
	}

	// 3. No hay applet que resolver. Quien pidió ayuda la recibe, y quien no,
	// un fallo que nombra lo que no se ha reconocido y enumera lo que existe.
	if previo.Ayuda {
		return Despacho{
			Destino: DestinoAyuda,
			Ayuda:   AyudaDelBinario(invocacion, registro),
			Args:    resto,
		}, nil
	}

	return Despacho{}, appletNoResuelto(invocacion, registro, resto)
}

// despachoReservado atiende un verbo del propio binario, que no admite
// argumentos ni banderas: «version» no tiene sobre ni banderas (D16) y lo que
// sobra tras él no es un argumento que descartar sino una invocación que hay
// que corregir, código 2, con el mensaje nombrando lo que sobra (FR-027; es el
// mismo código con el que H0 respondía a «kitlegal version extra»,
// specs/001-h0-esqueleto-del-repo/contracts/cli-version.md). --version se
// atiende como «version», y el mensaje nombra lo que se escribió.
func despachoReservado(resto []string) (Despacho, error) {
	reservado, forma := resto[0], "el verbo"
	if reservado == banderaDeVersion {
		reservado, forma = verboVersion, "la bandera"
	}

	if len(resto) > 1 {
		return Despacho{}, fmt.Errorf("%w: %s %q no admite argumentos ni banderas, y recibió %q",
			cli.ErrArgumentos, forma, resto[0], resto[1:])
	}

	return Despacho{Destino: DestinoReservado, Reservado: reservado}, nil
}

// despachoDeApplet normaliza el verbo con el applet ya resuelto y **antes** de
// construir la gramática, de modo que el analizador vea siempre un verbo
// explícito (D26).
//
// Las cuatro ramas son las cuatro líneas de contracts/registro-y-describe.md
// §2 bis, en este orden: el primer argumento nombra un verbo y no se toca nada;
// se pidió la ayuda del applet y no se inserta nada, porque pedir la ayuda no
// resuelve ningún verbo; hay verbo por omisión y se inserta a la cabeza; no lo
// hay y nombrar el verbo era obligatorio.
func despachoDeApplet(applet Applet, resto []string, previo cli.Preliminar) (Despacho, error) {
	if nombraVerbo(applet, resto) {
		return Despacho{Destino: DestinoApplet, Applet: applet, Args: resto}, nil
	}

	if previo.Ayuda {
		return Despacho{
			Destino: DestinoAyuda,
			Applet:  applet,
			Ayuda:   AyudaDelApplet(applet),
			Args:    resto,
		}, nil
	}

	if porOmision, hay := verboPorOmision(applet); hay {
		// La lista es nueva: la que recibió quien invoca queda intacta.
		args := make([]string, 0, len(resto)+1)
		args = append(args, porOmision)
		args = append(args, resto...)

		return Despacho{Destino: DestinoApplet, Applet: applet, Args: args}, nil
	}

	return Despacho{}, fmt.Errorf("%w: el applet %q no declara ningún verbo por omisión, así"+
		" que hay que nombrar uno; %s", cli.ErrArgumentos, applet.Nombre(), listaDeVerbos(applet))
}

// nombraVerbo dice si el primer argumento restante es el nombre de un verbo del
// applet. Mira **solo** el primero, y no el primero que no sea una bandera: sin
// gramática no hay forma de saber si un token es el valor de la bandera anterior,
// y la regla tiene que ser una que se pueda escribir en una línea y comprobar
// (D26).
//
// Lo que va tras el terminador tampoco nombra verbo alguno, que es la salida
// documentada para quien necesite pasar como argumento una palabra que se llama
// igual que un verbo.
func nombraVerbo(applet Applet, resto []string) bool {
	if len(resto) == 0 || resto[0] == terminadorDeBanderas {
		return false
	}

	return slices.ContainsFunc(applet.Verbos(), func(verbo Verbo) bool {
		return verbo.Nombre == resto[0]
	})
}

// verboPorOmision devuelve el verbo que se toma cuando la invocación no nombra
// ninguno, y si el applet lo declara. Que no haya más de uno lo comprobó el registro
// al construirse, así que aquí basta con el primero marcado (FR-008).
func verboPorOmision(applet Applet) (string, bool) {
	for _, verbo := range applet.Verbos() {
		if verbo.PorOmision {
			return verbo.Nombre, true
		}
	}

	return "", false
}

// appletNoResuelto construye el fallo de la invocación cuyo applet no puede
// determinarse: sin primer argumento o con uno que no está registrado. El
// mensaje nombra lo desconocido, enumera lo disponible y dice cómo pedir la
// versión, porque quien lo lee —persona o agente— necesita saber qué escribir
// a continuación (FR-006).
func appletNoResuelto(invocacion string, registro *Registro, resto []string) error {
	if len(resto) == 0 {
		return fmt.Errorf("%w: no se ha indicado ningún applet; %s; %s",
			cli.ErrArgumentos, listaDeApplets(registro), comoPedirLaVersion(invocacion))
	}

	return fmt.Errorf("%w: %q no es ningún applet de kitlegal; %s; %s",
		cli.ErrArgumentos, resto[0], listaDeApplets(registro), comoPedirLaVersion(invocacion))
}

// nombreDeInvocacion es el último componente del nombre con el que se llamó al
// binario, sin el sufijo de los ejecutables de windows: lo que hace que un enlace
// simbólico «echo -> kitlegal» invoque el applet «echo» (FR-002).
func nombreDeInvocacion(argv []string) string {
	if len(argv) == 0 {
		return ""
	}

	return strings.TrimSuffix(filepath.Base(argv[0]), sufijoEjecutable)
}

// argumentosDe separa los argumentos del nombre del programa. Una lista vacía no
// puede llegar del sistema operativo, pero sí de quien llame a la raíz de
// composición, y termina en un error de argumentos y no en un pánico (FR-033).
func argumentosDe(argv []string) []string {
	if len(argv) == 0 {
		return nil
	}

	return argv[1:]
}
