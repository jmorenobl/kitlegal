package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/render"
)

// formatoDeVersion es el texto del verbo reservado «version»: las mismas tres
// líneas que fijó H0 en su contrato, sin sobre y sin banderas. Lo único que
// cambia respecto de H0 es por dónde sale —el presentador— y que el error de
// escritura se propaga en lugar de descartarse (D16, D27).
const formatoDeVersion = "kitlegal %s\ncommit: %s\nfecha:  %s\n"

// verboVersion es el único verbo reservado que el binario atiende hoy, y el
// mismo que verbosReservados impide tomar como nombre de applet (registro.go,
// D17). Cuando aparezca un segundo, esta comparación se convierte en un switch
// y el registro seguirá siendo la única lista.
const verboVersion = "version"

// prefijoDeEnsayo encabeza todas las líneas que --dry-run deja en la salida de
// error —la del kernel y una por cada operación que una capa con efectos
// describió—, de modo que quien las lea sepa desde el primer carácter de cada
// una que no se ejecutó nada.
const prefijoDeEnsayo = "--dry-run: "

// prefijoDeCampo nombra los campos de la gramática que se construye al vuelo. El
// nombre no se deriva del verbo a propósito: el campo de un struct tiene que ser
// un identificador Go exportado y el nombre de un verbo es texto libre, así que
// este viaja en la etiqueta `name`, que es de donde Kong lo toma.
const prefijoDeCampo = "Verbo"

// Los tres fallos que esta raíz puede encontrar y que **no** son de quien
// invoca sino de quien escribió el applet o de una ampliación del kernel que se
// dejó a medias. Ninguno lleva los sentinelas de internal/cli, así que salen con
// el código de lo que nadie previó y nunca con uno reservado (FR-031).
var (
	// errGramaticaImposible es el verbo cuya fábrica de argumentos falta o no
	// devuelve lo que el contrato promete.
	errGramaticaImposible = errors.New("app: la gramática del applet no se pudo construir")
	// errVerboDesconocido es el verbo que el análisis seleccionó y que el applet
	// no declara. No lo puede provocar ninguna invocación: la gramática sale del
	// mismo catálogo que se consulta después.
	errVerboDesconocido = errors.New("app: el verbo seleccionado no es del applet")
	// errSinAtender es el destino o la decisión que nadie ha enseñado a atender:
	// lo que ocurriría si se añadiera un valor al vocabulario y se olvidara esta
	// rama.
	errSinAtender = errors.New("app: la invocación no se pudo atender")
)

// Main es la raíz de composición del kernel: monta el presentador con los dos
// descriptores que recibe, lo inyecta en todo lo que escribe, resuelve la
// invocación entera y devuelve su código de salida.
//
// **Nunca llama a os.Exit** —quien termina el proceso es el punto de entrada, y
// es lo que hace comprobable el binario sin lanzar un subproceso— y ningún
// camino de usuario termina en pánico: los fallos se propagan como error hasta
// el único punto que los traduce (FR-030, FR-033, FR-035, research.md D2).
//
// argv es la lista entera, con el nombre del programa a la cabeza, de la que
// sale el despacho multicall. registro es el registro ya construido —el del
// binario que se publica o el del binario de e2e—. version, commit y fecha
// son los datos de construcción que inyecta -ldflags y que atiende el verbo
// reservado «version» (D16).
func Main(
	argv []string,
	registro *Registro,
	stdout, stderr io.Writer,
	version, commit, fecha string,
) int {
	desarmarTuberiaCerrada()

	presentador := render.Nuevo(stdout, stderr)
	previo := cli.PreEscanear(argumentosDe(argv))

	fin := resolver(presentador, registro, argv, previo,
		fmt.Sprintf(formatoDeVersion, version, commit, fecha))

	// Un desenlace sin sobre y sin fallo —«version», la ayuda, --describe y
	// --dry-run— no cita nada: su código sale de la misma traducción, que sin
	// error es 0 (FR-022, FR-026, FR-049).
	if fin.err == nil && !fin.conSobre {
		return cli.CodigoSalida(nil)
	}

	// El único punto del proyecto que traduce un error a código de salida, y el
	// único que emite el sobre de fallo. Las dos cosas, juntas y aquí
	// (FR-030, FR-045).
	var montador cli.Montador

	return montador.Emitir(presentador, fin.enJSON, fin.resultado, fin.err)
}

// desarmarTuberiaCerrada hace que escribir en una tubería cuyo lector ha
// terminado —«kitlegal … | head -1»— sea un error corriente y no la muerte del
// proceso.
//
// Sin esto, el runtime de Go termina el proceso con SIGPIPE en cuanto una
// escritura en el descriptor 1 o 2 falla con EPIPE (os/signal, «SIGPIPE»): sin
// código de salida, sin mensaje y sin pasar por la traducción única, de modo que
// la propagación del presentador y el código 1 que promete el contrato nunca
// llegarían a ejecutarse. Con la señal ignorada, la escritura devuelve EPIPE y
// sube como cualquier otro fallo de escritura (FR-031, FR-033,
// contracts/banderas-y-exit-codes.md §4, research.md D15).
//
// Vive aquí y no en cada `package main` porque la garantía es del kernel y no
// de un binario concreto: todo ejecutable que compone el kernel —el distribuido
// y el del e2e— la hereda por llamar a Main, sin poder olvidarla. Ignorar la
// señal es idempotente y no toca nada más del proceso, que es lo que permite
// que los tests llamen a Main tantas veces como quieran en el mismo proceso.
func desarmarTuberiaCerrada() {
	signal.Ignore(syscall.SIGPIPE)
}

// desenlace es en qué queda una invocación una vez resuelta y antes de
// traducirla a código de salida.
type desenlace struct {
	// enJSON es la forma de presentación ya definitiva: la que dedujo el
	// pre-escaneo mientras no había gramática, y la de --json en cuanto el
	// análisis termina bien (research.md D25).
	enJSON bool
	// conSobre dice si la invocación produjo algo que presentar en la salida
	// estándar. Es falso en «version», en la ayuda, en --describe y en
	// --dry-run, que escriben por su cuenta o no escriben nada.
	conSobre bool
	// resultado es lo que devolvió el applet. En un fallo aporta la procedencia
	// de la fuente que se estaba consultando, que es la que cita el sobre de
	// fallo cuando se conoce (FR-045).
	resultado schema.Resultado
	// err es el fallo de la invocación, o nulo si no lo hubo.
	err error
}

// resolver decide quién atiende la invocación y la atiende, dejando la
// traducción del fallo a código de salida para quien llama.
func resolver(
	p cli.Presentador,
	registro *Registro,
	argv []string,
	previo cli.Preliminar,
	textoDeVersion string,
) desenlace {
	fin := desenlace{enJSON: previo.JSON}

	// El nivel se resuelve antes de que exista gramática que analizar, de modo
	// que un fallo del análisis también quede registrado. La variable de entorno
	// la lee la raíz de composición y no el kernel, que así queda libre de
	// entrada y salida (research.md D14).
	registrador, err := cli.NuevoRegistrador(p, os.Getenv(cli.VariableNivel), previo.Verbose)
	if err != nil {
		fin.err = err

		return fin
	}

	despacho, err := Despachar(registro, argv, previo)
	if err != nil {
		fin.err = err

		return fin
	}

	switch despacho.Destino {
	case DestinoReservado:
		fin.err = atenderReservado(p, despacho.Reservado, textoDeVersion)

		return fin
	case DestinoAyuda:
		fin.err = p.Texto(despacho.Ayuda)

		return fin
	case DestinoApplet:
		return atender(p, registrador, despacho, previo)
	}

	// Destino es un tipo con base string, así que existen valores fuera de las
	// tres constantes. Ninguna ruta de este paquete los produce, pero lo que no
	// se ha previsto sale como tal y no como un silencio (FR-031).
	fin.err = fmt.Errorf("%w: destino %q", errSinAtender, despacho.Destino)

	return fin
}

// atenderReservado responde a los verbos que el binario se guarda para sí. Hoy
// es uno solo: «version», que sale por el presentador sin sobre y sin banderas
// (D16).
func atenderReservado(p cli.Presentador, reservado, textoDeVersion string) error {
	if reservado != verboVersion {
		return fmt.Errorf("%w: verbo reservado %q", errSinAtender, reservado)
	}

	return p.Texto(textoDeVersion)
}

// atender ejecuta la parte de la invocación que corresponde a un applet y emite
// el **único** registro de eventos que el kernel produce por invocación, con la
// clase del fallo si lo hubo (FR-038, research.md D14).
func atender(
	p cli.Presentador,
	registrador *slog.Logger,
	despacho Despacho,
	previo cli.Preliminar,
) desenlace {
	inicio := time.Now()
	fin, verbo := resolverApplet(p, registrador, despacho, previo)

	cli.RegistrarEvento(context.Background(), registrador, cli.Evento{
		Applet:     despacho.Applet.Nombre(),
		Verbo:      verbo,
		Duracion:   time.Since(inicio),
		Clase:      claseDe(fin.err),
		Argumentos: despacho.Args,
	})

	return fin
}

// resolverApplet construye la gramática del applet, la analiza y atiende la
// decisión que gane la prelación. Devuelve además el verbo que quedó
// seleccionado, que es lo que el registro de eventos necesita saber.
func resolverApplet(
	p cli.Presentador,
	registrador *slog.Logger,
	despacho Despacho,
	previo cli.Preliminar,
) (desenlace, string) {
	fin := desenlace{enJSON: previo.JSON}

	gramatica, err := construirGramatica(despacho.Applet)
	if err != nil {
		fin.err = err

		return fin, ""
	}

	// El nombre que se le da al analizador es el del applet y no el de la
	// invocación: es lo que hace que la ayuda por enlace simbólico y la ayuda
	// por primer argumento salgan idénticas byte a byte (FR-007, SC-003).
	analisis, err := cli.Analizar(
		p, despacho.Applet.Nombre(), gramatica.valor.Interface(), despacho.Args)
	if err != nil {
		fin.err = err

		return fin, ""
	}

	// En cuanto el análisis termina bien manda lo analizado, y el pre-escaneo se
	// descarta (research.md D25).
	fin.enJSON = analisis.Globales.JSON

	switch analisis.Decision {
	case cli.DecisionAyuda:
		// La ayuda del verbo ya salió por el escritor del presentador: no queda
		// sobre que emitir y el código es 0 (research.md D11).
		return fin, analisis.Verbo
	case cli.DecisionDescribir:
		fin.err = describir(p, despacho.Applet, analisis.Verbo)

		return fin, analisis.Verbo
	case cli.DecisionEjecutar:
		return ejecutarVerbo(p, registrador, despacho, analisis, gramatica, fin)
	}

	fin.err = fmt.Errorf("%w: decisión %q", errSinAtender, analisis.Decision)

	return fin, analisis.Verbo
}

// ejecutarVerbo entrega el control al applet con el contexto de ejecución —las
// seis opciones globales que le llegan ya interpretadas— y el registrador ya
// montado, de modo que no tenga que leer ninguna bandera (FR-018).
func ejecutarVerbo(
	p cli.Presentador,
	registrador *slog.Logger,
	despacho Despacho,
	analisis cli.Analisis,
	gramatica gramaticaDelApplet,
	fin desenlace,
) (desenlace, string) {
	argumentos, hay := gramatica.argumentos(analisis.Verbo)
	if !hay {
		fin.err = fmt.Errorf("%w: %q de %q",
			errVerboDesconocido, analisis.Verbo, despacho.Applet.Nombre())

		return fin, analisis.Verbo
	}

	ejecucion := analisis.Globales.Contexto()

	// --timeout es el plazo de **toda** la operación y no el de una petición
	// suelta (FR-020, research.md D9).
	ctx, cancelar := context.WithTimeout(context.Background(), ejecucion.Timeout)
	defer cancelar()

	resultado, err := argumentos.Ejecutar(ctx, ejecucion, registrador)
	err = conPlazoAgotado(ctx, err)

	if ejecucion.DryRun {
		// --dry-run no ha cortado antes del applet: la bandera viajó en el
		// contexto de ejecución, y es cada capa con efectos la que la honra
		// describiendo en lugar de ejecutar. Lo que describan llega aquí en
		// Resultado.Ensayo y lo presenta el kernel, que es quien tiene el
		// presentador (FR-022, FR-051, research.md D10, D6).
		fin.err = conDescripcion(p, despacho, analisis.Verbo, resultado.Ensayo, err)

		return fin, analisis.Verbo
	}

	fin.conSobre = true
	fin.resultado = resultado
	fin.err = err

	return fin, analisis.Verbo
}

// conDescripcion escribe la descripción de --dry-run y decide qué fallo sigue
// mandando sobre el código de salida.
//
// La descripción se escribe siempre, también cuando el applet ha fallado: es un
// mensaje dirigido a la persona y no la consecuencia de un éxito. Por eso se
// presenta igualmente lo que el applet alcanzara a dejar en Ensayo antes de
// fallar. El fallo del applet, si lo hubo, es el que se cuenta —el de la
// escritura del aviso no lo sustituye—, con el mismo criterio por el que el
// montador no reintenta un descriptor que acaba de fallar.
func conDescripcion(
	p cli.Presentador,
	despacho Despacho,
	verbo string,
	ensayo []string,
	err error,
) error {
	errAviso := p.Aviso(descripcionDeLaOperacion(despacho, verbo, ensayo))
	if err != nil {
		return err
	}

	return errAviso
}

// descripcionDeLaOperacion es lo que --dry-run deja en la salida de error: qué
// applet, qué verbo y con qué argumentos se habría ejecutado, y a continuación
// una línea por cada operación que una capa con efectos describió en lugar de
// ejecutar. Va por el presentador y no por el registro de eventos, que es lo que
// la hace visible con cualquier nivel (FR-022, FR-051, research.md D10, D6).
//
// Sale como un único texto y no como un aviso por línea porque es un solo
// mensaje: así ninguna escritura puede fallar a medias y dejar en la salida de
// error una descripción incompleta sin que nadie lo note.
func descripcionDeLaOperacion(despacho Despacho, verbo string, ensayo []string) string {
	lineas := make([]string, 0, 1+len(ensayo))

	lineas = append(lineas, fmt.Sprintf(
		prefijoDeEnsayo+"no se ha ejecutado nada; se habría ejecutado el applet %q,"+
			" el verbo %q, con los argumentos %q",
		despacho.Applet.Nombre(), verbo, despacho.Args))

	for _, operacion := range ensayo {
		lineas = append(lineas, prefijoDeEnsayo+"se habría pedido "+operacion)
	}

	return strings.Join(lineas, "\n")
}

// conPlazoAgotado convierte el plazo vencido en el error tipado de fuente no
// disponible. Lo decide el contexto y no el applet: uno que devuelve el error
// del plazo, uno que devuelve otro y uno que devuelve un resultado terminan los
// tres en el código 4 si el plazo se agotó (FR-020, research.md D9).
//
// El error del applet se incorpora como texto y **no** con %w a propósito:
// envolverlo dejaría en la cadena un sentinela de otra clase —«no encontrado»,
// por ejemplo— y la clasificación, que mira los cinco en orden, devolvería
// esa otra clase. Lo que clasifica aquí es el plazo, y su mensaje no se pierde.
func conPlazoAgotado(ctx context.Context, err error) error {
	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return err
	}

	if err == nil {
		return fmt.Errorf("%w: se agotó el plazo de la operación", cli.ErrFuenteNoDisponible)
	}

	return fmt.Errorf("%w: se agotó el plazo de la operación: %s",
		cli.ErrFuenteNoDisponible, err.Error())
}

// claseDe es la clase que se registra del desenlace de una invocación: la del
// fallo, o vacía si no lo hubo. Sale de la misma clasificación que el código de
// salida, así que el registro y el código nunca se contradicen.
func claseDe(err error) schema.Clase {
	if err == nil {
		return ""
	}

	return cli.Clasificar(err)
}

// describir emite el esquema de entrada y salida del verbo seleccionado. La
// traducción del Verbo del registro al que el kernel describe la hace este
// paquete, de modo que internal/cli siga sin conocer el tipo Applet
// (FR-046, research.md D2).
func describir(p cli.Presentador, applet Applet, nombre string) error {
	verbo, hay := verboDe(applet, nombre)
	if !hay || verbo.Argumentos == nil {
		return fmt.Errorf("%w: %q de %q", errVerboDesconocido, nombre, applet.Nombre())
	}

	ayuda := verbo.Descripcion

	return cli.Describir(p, cli.Verbo{
		Applet:     applet.Nombre(),
		Verbo:      verbo.Nombre,
		Ayuda:      ayuda,
		Argumentos: verbo.Argumentos(),
		Salida:     verbo.Salida,
	})
}

// verboDe busca en el catálogo del applet el verbo que se llama así. El registro
// ya comprobó al construirse que no hay dos con el mismo nombre (FR-008).
func verboDe(applet Applet, nombre string) (Verbo, bool) {
	for _, verbo := range applet.Verbos() {
		if verbo.Nombre == nombre {
			return verbo, true
		}
	}

	return Verbo{}, false
}

// gramaticaDelApplet es la parte de la gramática que el kernel no conoce: un
// struct con un mandato por verbo del applet, construido al vuelo porque los
// verbos no se saben al compilar el kernel sino al leer el registro. Viaja a
// internal/cli como `any` —dentro de kong.Plugins, que exige punteros a struct—
// y vuelve con los argumentos ya analizados (research.md D1, D2,
// gates/supuestos-kong.md).
type gramaticaDelApplet struct {
	// valor es el puntero al struct de la gramática.
	valor reflect.Value
	// campos dice en qué campo quedó cada verbo, para recuperar sus argumentos
	// cuando el análisis diga cuál se seleccionó.
	campos map[string]int
}

// construirGramatica arma el struct de la invocación a partir del catálogo del
// applet. Cada campo se rellena con lo que devolvió la fábrica del verbo, de
// modo que los valores que la fábrica dejara puestos llegan al análisis y dos
// invocaciones no comparten nunca la misma instancia de argumentos.
func construirGramatica(applet Applet) (gramaticaDelApplet, error) {
	verbos := applet.Verbos()
	campos := make([]reflect.StructField, 0, len(verbos))
	valores := make([]reflect.Value, 0, len(verbos))
	indice := make(map[string]int, len(verbos))

	for i, verbo := range verbos {
		argumentos, err := argumentosDelVerbo(applet, verbo)
		if err != nil {
			return gramaticaDelApplet{}, err
		}

		campos = append(campos, reflect.StructField{
			Name: prefijoDeCampo + strconv.Itoa(i),
			Type: argumentos.Type(),
			Tag:  etiquetaDelVerbo(verbo),
		})
		valores = append(valores, argumentos)
		indice[verbo.Nombre] = i
	}

	gramatica := reflect.New(reflect.StructOf(campos))
	for i, valor := range valores {
		gramatica.Elem().Field(i).Set(valor)
	}

	return gramaticaDelApplet{valor: gramatica, campos: indice}, nil
}

// argumentosDelVerbo llama a la fábrica del verbo y comprueba que devuelve lo
// que el contrato promete: un puntero a struct con las etiquetas de la
// gramática. Devuelve el struct apuntado, que es a la vez el tipo del campo y el
// valor con el que se rellena.
//
// Una fábrica ausente o que devuelve otra cosa es un defecto de quien escribió
// el applet y no de quien invoca: sale como fallo inesperado, y nunca como un
// pánico dentro de la reflexión (FR-031, FR-033).
func argumentosDelVerbo(applet Applet, verbo Verbo) (reflect.Value, error) {
	if verbo.Argumentos == nil {
		return reflect.Value{}, fmt.Errorf(
			"%w: el verbo %q de %q no declara fábrica de argumentos",
			errGramaticaImposible, verbo.Nombre, applet.Nombre())
	}

	valor := reflect.ValueOf(verbo.Argumentos())
	if valor.Kind() != reflect.Pointer || valor.IsNil() || valor.Elem().Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf(
			"%w: la fábrica del verbo %q de %q no devuelve un puntero a struct",
			errGramaticaImposible, verbo.Nombre, applet.Nombre())
	}

	return valor.Elem(), nil
}

// etiquetaDelVerbo escribe la etiqueta del campo tal y como se escribiría a
// mano: `cmd:"" name:"<verbo>" help:"<descripción>"`. Los dos valores se citan
// con strconv.Quote, que es exactamente la forma que el analizador de etiquetas
// de Kong deshace, de modo que una descripción con comillas no rompa la
// gramática.
func etiquetaDelVerbo(verbo Verbo) reflect.StructTag {
	ayuda := verbo.Descripcion

	return reflect.StructTag(fmt.Sprintf(`cmd:"" name:%s help:%s`,
		strconv.Quote(verbo.Nombre), strconv.Quote(ayuda)))
}

// argumentos recupera los argumentos ya analizados del verbo seleccionado. El
// puntero al campo implementa Argumentos porque el tipo del campo salió de lo
// que devolvió la fábrica, que es ese mismo puntero.
func (g gramaticaDelApplet) argumentos(verbo string) (Argumentos, bool) {
	indice, hay := g.campos[verbo]
	if !hay {
		return nil, false
	}

	argumentos, correcto := g.valor.Elem().Field(indice).Addr().Interface().(Argumentos)

	return argumentos, correcto
}
