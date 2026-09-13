package httpx

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// topeDeRedirecciones es el número de saltos que el cliente sigue antes de
// declarar que la fuente no sabe entregar el recurso. Es el mismo de la política
// por omisión de la biblioteca —«stop after 10 consecutive requests» (go doc
// net/http.Client)— y queda por encima de los cinco que RFC 9309 §2.3.1.2 pide
// como mínimo para el robots.txt, que reutiliza este tope (D10).
const topeDeRedirecciones = 10

// cabeceraDelFormato es la cabecera por la que una fuente sabe en qué formato se
// le pide el recurso (RFC 9110 §12.5.1), y la única que quien pide elige: la
// rellena Peticion.Acepta (research D4 de H4).
const cabeceraDelFormato = "Accept"

// esquemasDeRed son los dos únicos esquemas que una dirección puede traer. Lo
// que no es uno de ellos no es una fuente caída sino una invocación que quien
// llama puede corregir (research D19).
const (
	esquemaHTTP  = "http"
	esquemaHTTPS = "https"
)

// intervaloPorOmision es la separación mínima entre dos peticiones a un mismo
// sitio cuando no se declara otra: un segundo, más lento que lo que cualquier
// fuente pública del proyecto tolera. Ser conservador por omisión es lo que hace
// que un adaptador nuevo no pueda maltratar un sitio por descuido; el ritmo que
// cada fuente admite llegará con la fuente (FR-020, D8).
const intervaloPorOmision = time.Second

// intentosPorOmision son los intentos que se hacen con cada petición cuando no
// se declara otra cosa: una petición y dos reintentos. Tres es lo que absorbe
// el tropiezo pasajero de una fuente pública sin insistirle a una que está
// caída, y el número no se deduce del ritmo ni de nada más (FR-024, D9).
const intentosPorOmision = 3

// Cliente es el único objeto del módulo capaz de emitir una petición HTTP, y
// Pedir su única operación: no hay ninguna otra forma de salir a la red desde
// este paquete, ni forma alguna de construir uno al que le falte una de sus
// garantías. Sus campos son privados porque ninguna de ellas se negocia desde
// fuera (FR-001, FR-002, D1).
//
// Es seguro para uso concurrente: no guarda nada de una petición a la siguiente.
type Cliente struct {
	// cliente ejecuta la cadena de decoradores, que se compone una vez en New y
	// no cambia después (D3).
	cliente *http.Client
	// sitios es el registro de sitios de este cliente: uno solo, del que sale el
	// ritmo de cada sitio y del que saldrán sus reglas de robots.txt, porque la
	// clave de sitio es una y su exclusión tiene que ser compartida
	// (data-model.md §4, D14). Es nulo en un cliente de reproducción, cuya
	// cadena no lleva ni ritmo ni robots.txt y que por tanto no tiene nada que
	// guardar por sitio (FR-049).
	sitios *sitios
	// fuente es el nombre lógico de la fuente que usa este cliente, el que
	// nombrará el directorio de las grabaciones (FR-039).
	fuente string
	// registrador es el destino de los eventos. Nunca es nulo: sin la opción,
	// descarta lo que se le entregue (FR-035, D15).
	registrador *slog.Logger
	// hora es de donde sale el instante de emisión que Pedir entrega. Nunca es
	// nula: sin la opción, es time.Now (contrato httpx-acepta-e-instante §3 de
	// H4).
	hora func() time.Time
}

// configuracionDelCliente es lo que las opciones rellenan antes de que New
// construya el cliente. No se exporta: lo que se declara son las opciones, no la
// estructura que escriben.
type configuracionDelCliente struct {
	fuente          string
	raizDeGrabacion string
	intervalo       time.Duration
	intentos        int
	reloj           relojDeEspera
	registrador     *slog.Logger
	hora            func() time.Time
}

// Opcion declara una variación del cliente. Devuelve el error de la opción
// inválida, pero quien lo entrega es el constructor: validar en la construcción
// y no en cada uso es lo que hace que un cliente exista solo si es válido
// (patrón de golang-design-patterns; contrato §2).
type Opcion func(*configuracionDelCliente) error

// ConFuente declara el nombre lógico de la fuente —«boe», «placsp»…—, que es el
// que nombra el directorio de las grabaciones. Un nombre vacío o que pudiera
// escaparse de ese directorio es un error de argumentos (FR-039).
func ConFuente(nombre string) Opcion {
	return func(config *configuracionDelCliente) error {
		if nombre == "" {
			return errorDeArgumentos(Peticion{}, nil,
				"el nombre de la fuente no puede ir vacío (ConFuente)")
		}

		if strings.ContainsAny(nombre, `/\`) || strings.Contains(nombre, "..") {
			return errorDeArgumentos(Peticion{}, nil,
				`el nombre de la fuente no puede llevar «/», «\» ni «..» (ConFuente): `+nombre)
		}

		config.fuente = nombre

		return nil
	}
}

// ConRaizDeGrabacion declara el directorio bajo el que la grabación escribe
// <fuente>/<nombre>.json. No tiene valor por omisión y el cliente no lo deduce
// de ninguna parte —ni del directorio de trabajo del proceso, ni de la raíz del
// módulo, ni de ninguna ruta relativa escrita en el código—: dónde viven los
// fixtures lo sabe quien llama, y este paquete es deliberadamente agnóstico a
// esa decisión. Es la simetría exacta de Replay(dir), que también recibe el suyo
// (FR-064).
//
// Solo se usa cuando la grabación está activa; declararla sin la variable de
// entorno no la enciende (FR-041). Un directorio vacío es un error de
// argumentos: no declarar la raíz y declararla vacía son lo mismo, y lo que
// FR-064 prohíbe es justamente que el paquete rellene ese hueco por su cuenta.
func ConRaizDeGrabacion(dir string) Opcion {
	return func(config *configuracionDelCliente) error {
		if dir == "" {
			return errorDeArgumentos(Peticion{}, nil,
				"la raíz de grabación no puede ir vacía (ConRaizDeGrabacion)")
		}

		config.raizDeGrabacion = dir

		return nil
	}
}

// ConIntervalo declara la separación mínima entre dos peticiones a un mismo
// sitio. Vale para todos los sitios de ese cliente: el ámbito del limitador es
// el sitio —dos sitios distintos no compiten entre sí (FR-019)—, pero su valor
// es uno solo por cliente, y el que cada fuente tolera se fijará donde se declare
// la fuente (FR-020).
//
// Un intervalo nulo o negativo es un error de argumentos: un ritmo sin espera es
// lo contrario de lo que este cliente garantiza, y no hay ninguna opción para
// desactivarlo (D8).
func ConIntervalo(d time.Duration) Opcion {
	return func(config *configuracionDelCliente) error {
		if d <= 0 {
			return errorDeArgumentos(Peticion{}, nil,
				"el intervalo entre peticiones a un mismo sitio tiene que ser mayor que cero (ConIntervalo): "+d.String())
		}

		config.intervalo = d

		return nil
	}
}

// ConIntentos declara cuántas veces se pide como mucho cada petición, la
// primera incluida: tres por omisión, es decir una petición y dos reintentos
// (FR-024). Lo que se repite y lo que se espera entre un intento y el siguiente
// no se negocia —es la lista cerrada de FR-025 y la ley de D9—; lo único que se
// declara aquí es cuántas veces.
//
// Menos de un intento es un error de argumentos: cero peticiones no es una
// política de reintentos sino una petición que no se emite. No hay ninguna
// opción para desactivarlos; declarar un solo intento es la forma de no
// reintentar (FR-002).
func ConIntentos(n int) Opcion {
	return func(config *configuracionDelCliente) error {
		if n < 1 {
			return errorDeArgumentos(Peticion{}, nil,
				"el número de intentos por petición tiene que ser al menos uno (ConIntentos): "+strconv.Itoa(n))
		}

		config.intentos = n

		return nil
	}
}

// ConRegistrador declara el destino de los eventos del cliente, que es el mismo
// registrador que el kernel entrega al applet. Sin esta opción los eventos se
// descartan; nunca se emiten por el registrador global ni por la salida estándar
// (FR-035, D15).
func ConRegistrador(registrador *slog.Logger) Opcion {
	return func(config *configuracionDelCliente) error {
		if registrador == nil {
			return errorDeArgumentos(Peticion{}, nil,
				"el registrador de eventos no puede ser nulo (ConRegistrador)")
		}

		config.registrador = registrador

		return nil
	}
}

// ConHora declara de dónde sale el instante de emisión que Pedir entrega en
// Respuesta.Instante y en Error.Instante: time.Now por omisión. Vale para New y
// para Replay, porque el cliente de reproducción también declara cuándo sirvió
// cada grabación, y sustituirla es lo que permite fijar ese instante sin depender
// del reloj del sistema (contrato httpx-acepta-e-instante §3 de H4).
//
// Una hora nula es un error de argumentos: no declararla ya da la del sistema, y
// declararla nula solo puede ser un descuido que dejaría a Pedir sin instante.
func ConHora(ahora func() time.Time) Opcion {
	return func(config *configuracionDelCliente) error {
		if ahora == nil {
			return errorDeArgumentos(Peticion{}, nil,
				"la hora de la que sale el instante de emisión no puede ser nula (ConHora)")
		}

		config.hora = ahora

		return nil
	}
}

// New construye el cliente contra la red, con sus garantías ya puestas sin
// declarar ninguna opción: identificación en toda petición, robots.txt del sitio
// consultado antes de la primera petición, reintentos de lo que puede ser
// pasajero, ritmo por sitio, plazo del contexto y redirecciones seguidas por él
// mismo. Las opciones se aplican en orden —la última repetida gana— y la primera
// inválida termina la construcción con su clase (FR-001, contrato §2 y §3).
//
// El orden de la cadena no es arbitrario: los reintentos van por encima del
// ritmo para que cada uno espere su turno en el sitio, y el robots.txt por
// encima de los dos para que su propia obtención se reintente y espere turno
// igual que cualquier otra petición (FR-017, FR-021, D3).
func New(opciones ...Opcion) (*Cliente, error) {
	config, err := configuracionDe(opciones)
	if err != nil {
		return nil, err
	}

	config.completar()

	// La variable de entorno se lee una sola vez y aquí, y no en la raíz de
	// composición como la del registro de eventos: el cliente lo construye el
	// adaptador y no el kernel, y una opción para inyectar el entorno sería una
	// forma de encender la grabación desde el código, que es lo que FR-041
	// prohíbe. Todo lo que la grabación necesita se valida en este punto, antes
	// de que exista ningún cliente que pudiera escribir nada (FR-042, D12).
	valor, declarada := os.LookupEnv(VariableGrabacion)

	directorio, err := directorioDeGrabacion(valor, declarada, config)
	if err != nil {
		return nil, err
	}

	registro := nuevosSitios(config.intervalo)

	// La cadena se compone de dentro afuera, que es el orden en que cada
	// escalón depende del anterior: la marca de emisión es lo que envuelve al
	// transporte —anota la hora justo antes de entregarle cada intento, ya con
	// su turno esperado (contrato httpx-acepta-e-instante §4 de H4)—, la
	// grabación va por encima —graba la petición tal como salió y la respuesta
	// tal como llegó—, el ritmo por encima porque es la última espera que una
	// petición atraviesa antes de emitirse, los reintentos por encima para que
	// cada intento espere su turno, el robots.txt por encima de ellos para que
	// su propia obtención pase por los dos, y la identificación arriba del todo
	// (FR-017, FR-021, D3).
	//
	// Sin grabación no hay escalón que grabe: apagarla no es una bandera que el
	// decorador consulte, sino una cadena en la que no está (FR-041).
	cadena := conMarcaDeEmision(nuevoTransporte(), config.hora)

	if directorio != "" {
		cadena = conGrabacion(cadena, directorio)
	}

	cadena = conRitmo(cadena, registro)
	cadena = conReintentos(cadena, config.intentos, config.reloj)
	cadena = conRobots(cadena, registro)
	cadena = conIdentificacion(cadena)

	return &Cliente{
		cliente:     nuevoClienteHTTP(cadena),
		sitios:      registro,
		fuente:      config.fuente,
		registrador: config.registrador,
		hora:        config.hora,
	}, nil
}

// Replay construye el cliente que responde exclusivamente desde las grabaciones
// de un directorio y que no abre ninguna conexión bajo ninguna circunstancia, ni
// siquiera para el robots.txt: en su cadena no hay transporte de red que pudiera
// abrirla (FR-045, FR-046, D13).
//
// El directorio es el de grabaciones de una fuente —el <raíz>/<fuente> que deja
// la grabación—, y lo declara quien llama: este paquete no lo deduce de ninguna
// parte, exactamente igual que no deduce la raíz de grabación (FR-064).
//
// La cadena se reduce a la lista cerrada de garantías vigentes de FR-049:
// contexto, identificación, método, redirecciones grabadas con su tope
// —resueltas dentro del propio directorio, donde cada salto es una búsqueda
// más— y clasificación por el estado grabado. No lleva robots.txt —una grabación
// suya en el directorio queda sin usar y no es un error—, ni ritmo, ni
// reintentos: los tres carecen de sentido sin fuente real y romperían el
// determinismo de FR-048.
//
// Entre la identificación y la reproducción va la marca de emisión, que no es
// una garantía sino una medida: no abre nada ni espera nada, y anota la hora del
// cliente al servir cada grabación, que es el instante que declara la respuesta
// reproducida (contrato httpx-acepta-e-instante §4 de H4).
func Replay(dir string, opciones ...Opcion) (*Cliente, error) {
	config, err := configuracionDe(opciones)
	if err != nil {
		return nil, err
	}

	if err := comprobarReproduccion(dir, config); err != nil {
		return nil, err
	}

	// De los valores por omisión, esta cadena solo usa el registrador y la hora:
	// ni el ritmo ni los intentos tienen escalón que los lea, y por eso
	// declararlos es un error y no una preferencia que no se cumple.
	config.completar()

	cadena := conIdentificacion(conMarcaDeEmision(nuevoTransporteDeReproduccion(dir), config.hora))

	return &Cliente{
		cliente:     nuevoClienteHTTP(cadena),
		fuente:      config.fuente,
		registrador: config.registrador,
		hora:        config.hora,
	}, nil
}

// configuracionDe aplica las opciones en orden sobre una configuración a cero
// —la última repetida gana— y devuelve el primer fallo, que es el de la opción
// inválida con su clase (contrato §2).
//
// Lo que ninguna opción escribe se queda a cero, y eso es lo que distingue «no
// declarada» de «declarada con el valor por omisión»: sin esa distinción, Replay
// no podría rechazar las opciones que no tienen sentido en reproducción
// (contrato §3).
func configuracionDe(opciones []Opcion) (configuracionDelCliente, error) {
	var config configuracionDelCliente

	for _, opcion := range opciones {
		if err := opcion(&config); err != nil {
			return configuracionDelCliente{}, err
		}
	}

	return config, nil
}

// completar rellena con los valores por omisión del contrato §2 lo que ninguna
// opción declaró. Se aplica después de las opciones y nunca antes, que es lo que
// deja el cero de cada campo disponible para decidir quién la declaró.
func (config *configuracionDelCliente) completar() {
	if config.intervalo == 0 {
		config.intervalo = intervaloPorOmision
	}

	if config.intentos == 0 {
		config.intentos = intentosPorOmision
	}

	if config.reloj == nil {
		config.reloj = dormirInterrumpible
	}

	if config.registrador == nil {
		config.registrador = slog.New(slog.DiscardHandler)
	}

	if config.hora == nil {
		config.hora = time.Now
	}
}

// comprobarReproduccion aplica la tabla de construcción de Replay del contrato
// §3, y la aplica entera antes de que exista ningún cliente: el directorio tiene
// que existir y ser un directorio —se declara para leer de algo que ya está—, la
// grabación y la reproducción no pueden estar activas a la vez (FR-043) y las
// opciones que solo tienen sentido contra una fuente real se rechazan en vez de
// aceptarse y no hacer nada. Todo es de la clase «argumentos», que quien llama
// corrige (FR-063).
func comprobarReproduccion(dir string, config configuracionDelCliente) error {
	if dir == "" {
		return errorDeArgumentos(Peticion{}, nil,
			"el directorio de reproducción no puede ir vacío (Replay)")
	}

	delDirectorio, err := os.Stat(dir)
	if err != nil {
		return errorDeArgumentos(Peticion{}, err,
			"el directorio de reproducción no existe o no se puede consultar: "+dir)
	}

	if !delDirectorio.IsDir() {
		return errorDeArgumentos(Peticion{}, nil,
			"el directorio de reproducción no es un directorio: "+dir)
	}

	// Cualquier valor no vacío de la variable se rechaza, y no solo el que
	// enciende la grabación: reproduciendo no hay nada que grabar, así que una
	// variable puesta —aunque traiga un valor que New también rechazaría— solo
	// puede ser un descuido, y resolverlo por azar es lo que FR-043 prohíbe.
	if valor, declarada := os.LookupEnv(VariableGrabacion); declarada && valor != "" {
		return errorDeArgumentos(Peticion{}, nil,
			"la grabación y la reproducción no pueden estar activas a la vez: "+
				VariableGrabacion+"="+valor+" con la reproducción de "+dir)
	}

	return comprobarOpcionesDeReproduccion(config)
}

// comprobarOpcionesDeReproduccion rechaza las tres opciones que solo tienen
// sentido contra una fuente real: la raíz de grabación —que además sería grabar
// y reproducir a la vez (FR-043)— y el ritmo y los intentos, cuyos escalones no
// están en esta cadena. Rechazarlas es lo que impide que quien las declare crea
// que hacen algo (contrato §3, D1).
func comprobarOpcionesDeReproduccion(config configuracionDelCliente) error {
	switch {
	case config.raizDeGrabacion != "":
		return errorDeArgumentos(Peticion{}, nil,
			"ConRaizDeGrabacion no tiene sentido en reproducción: un cliente de reproducción no graba nada")

	case config.intervalo != 0:
		return errorDeArgumentos(Peticion{}, nil,
			"ConIntervalo no tiene sentido en reproducción: la reproducción no espera nunca (FR-048)")

	case config.intentos != 0:
		return errorDeArgumentos(Peticion{}, nil,
			"ConIntentos no tiene sentido en reproducción: un estado grabado no se vuelve a buscar")

	default:
		return nil
	}
}

// Pedir es la única operación de red del módulo. Exige el contexto de
// cancelación como primer parámetro —el plazo de la operación es el suyo y solo
// el suyo (FR-003, FR-004)— y el contexto de ejecución del kernel como segundo,
// que es lo que hace que --dry-run llegue hasta aquí sin que quien llama pueda
// olvidarlo (FR-050, D1).
//
// Fuera de la cadena de decoradores viven las tres cosas que necesitan decidir
// antes o después de ella: la comprobación del método, de la dirección y del
// formato, que ocurre antes de abrir nada; el ensayo, que devuelve sin bajar por
// la cadena; y el bucle de redirecciones con la clasificación del resultado (D3).
//
// Lo que entrega, respuesta o fallo, lleva el instante de emisión de la última
// petición que salió hacia la fuente —el último intento del último salto—, que
// es la fecha de consulta que la fuente declarará (FR-096, SC-008). Solo se
// conoce dentro de la cadena, así que lo recoge una marca que Pedir pone en el
// contexto y lee al volver. Vacía porque en el último salto no salió ninguna
// petición —el robots.txt la deniega, no hay turno, el plazo terminó antes—, el
// instante es el del abandono, y el de un fallo de argumentos también; en ensayo
// no hay ninguno (contrato httpx-acepta-e-instante §2 y §4 de H4).
func (c *Cliente) Pedir(ctx context.Context, ejecucion schema.Contexto, p Peticion) (Respuesta, error) {
	direccion, fallo := comprobarPeticion(p)
	if fallo != nil {
		return Respuesta{}, fallo.fechado(c.hora())
	}

	if ejecucion.DryRun {
		c.registrador.DebugContext(ctx, "ensayo: la petición no se emite",
			"metodo", p.Metodo, "url", p.URL)

		return Respuesta{Peticion: p, URL: p.URL, Ensayo: true}, nil
	}

	marca := &marcaDeEmision{}

	respuesta, fallo := c.seguirLaCadena(contextoConMarca(ctx, marca), p, direccion, marca)

	instante := marca.instanteO(c.hora)

	if fallo != nil {
		return Respuesta{}, fallo.fechado(instante)
	}

	respuesta.Instante = instante

	return respuesta, nil
}

// comprobarPeticion aplica las tres comprobaciones que no necesitan red y que por
// eso rigen también en ensayo: el método es GET o HEAD (FR-010), la dirección es
// absoluta, de esquema de red y con sitio (research D19), y el formato, si se
// pide, se puede poner en una cabecera (research D4 de H4). Las tres son de la
// clase «argumentos» y ninguna llega a abrir una conexión.
func comprobarPeticion(p Peticion) (*url.URL, *Error) {
	if p.Metodo != http.MethodGet && p.Metodo != http.MethodHead {
		return nil, errorDeArgumentos(p, nil, "el método de una petición solo puede ser GET o HEAD")
	}

	direccion, err := url.Parse(p.URL)
	if err != nil {
		return nil, errorDeArgumentos(p, err, "la dirección no se puede interpretar")
	}

	if direccion.Scheme != esquemaHTTP && direccion.Scheme != esquemaHTTPS {
		return nil, errorDeArgumentos(p, nil, "la dirección tiene que ser absoluta y de esquema http o https")
	}

	if direccion.Host == "" {
		return nil, errorDeArgumentos(p, nil, "la dirección no nombra ningún sitio")
	}

	if fallo := comprobarFormato(p); fallo != nil {
		return nil, fallo
	}

	return direccion, nil
}

// comprobarFormato valida Peticion.Acepta antes de que llegue a la cabecera
// Accept. Vacío no pide ningún formato. Lo demás tiene que ser un tipo de
// contenido que mime.ParseMediaType acepte y no llevar ningún carácter de
// control, y las dos condiciones hacen falta: el analizador recorta los espacios
// que rodean el tipo —un «\r\n» final incluido— y admite tabuladores entre
// parámetros y casi cualquier byte dentro de un valor entrecomillado, de modo que
// por sí solo dejaría pasar un valor que partiría la cabecera o que la biblioteca
// rechazaría ya en la red (contrato httpx-acepta-e-instante §1 de H4).
//
// El valor se cita entrecomillado con strconv.Quote, para que un carácter de
// control no llegue crudo al mensaje ni a quien lo registre.
func comprobarFormato(p Peticion) *Error {
	if p.Acepta == "" {
		return nil
	}

	if strings.ContainsFunc(p.Acepta, unicode.IsControl) {
		return errorDeArgumentos(p, nil,
			"el formato pedido en la cabecera "+cabeceraDelFormato+" lleva un carácter de control: "+
				strconv.Quote(p.Acepta))
	}

	if _, _, err := mime.ParseMediaType(p.Acepta); err != nil {
		return errorDeArgumentos(p, err,
			"el formato pedido en la cabecera "+cabeceraDelFormato+" no es un tipo de contenido: "+
				strconv.Quote(p.Acepta))
	}

	return nil
}

// seguirLaCadena es el bucle de redirecciones: cada salto es una petición
// completa que baja por la cadena entera, de modo que el destino no hereda nada
// del origen (FR-011, D10). Termina cuando la respuesta ya no es una redirección
// seguible, cuando una dirección se repite o cuando se agota el tope; en los dos
// últimos casos, con la clase de quien no sabe entregar el recurso.
//
// Recibe, además del contexto que la lleva, la marca de emisión, para vaciarla
// antes de cada salto. Devuelve el *Error del paquete y no un error cualquiera,
// como las funciones de las que depende, porque Pedir fecha todo fallo antes de
// entregarlo y así ninguno puede llegarle sin fechar.
func (c *Cliente) seguirLaCadena(
	ctx context.Context, p Peticion, destino *url.URL, marca *marcaDeEmision,
) (Respuesta, *Error) {
	visitadas := make(map[string]struct{}, topeDeRedirecciones+1)

	for salto := 0; salto <= topeDeRedirecciones; salto++ {
		// Cada salto es una petición completa, y el que no llega a emitirse no
		// puede quedarse con el instante del anterior (contrato
		// httpx-acepta-e-instante §4 de H4).
		marca.reiniciar()

		// El contexto se comprueba antes de cada salto y no solo dentro del
		// transporte: lo que corta la operación es el plazo de quien la pidió, y
		// tiene que cortarla igual cuando abajo no hay ninguna conexión que
		// esperar sino una grabación que leer (FR-005, FR-049 a, D13).
		if err := ctx.Err(); err != nil {
			return Respuesta{}, errorDeFuenteNoDisponible(p, 0, err,
				"la operación ha terminado antes de pedir "+destino.String())
		}

		if _, repetida := visitadas[destino.String()]; repetida {
			return Respuesta{}, errorDeFuenteNoDisponible(p, 0, nil,
				"la cadena de redirecciones vuelve sobre una dirección ya visitada: "+destino.String())
		}

		visitadas[destino.String()] = struct{}{}

		recibida, fallo := c.emitir(ctx, p, destino)
		if fallo != nil {
			return Respuesta{}, fallo
		}

		if !esRedireccionSeguible(recibida.estado) {
			return entregar(p, destino, recibida)
		}

		siguiente, fallo := siguienteDestino(p, destino, recibida)
		if fallo != nil {
			return Respuesta{}, fallo
		}

		c.registrador.DebugContext(ctx, "redirección seguida",
			"de", destino.String(), "a", siguiente.String(), "estado", recibida.estado)

		destino = siguiente
	}

	return Respuesta{}, errorDeFuenteNoDisponible(p, 0, nil,
		"la cadena de redirecciones supera los "+strconv.Itoa(topeDeRedirecciones)+" saltos")
}

// recibida es lo que emitir devuelve: la respuesta con el cuerpo ya leído y el
// *http.Response ya cerrado. El único de todo el módulo se abre y se cierra
// aquí dentro, que es lo que impide que cerrarlo quede nunca en manos de quien
// llama (FR-002, D2).
type recibida struct {
	estado    int
	cabeceras Cabeceras
	cuerpo    []byte
}

// emitir baja una petición por la cadena y devuelve lo que la fuente respondió.
// La petición la construye nuevaPeticionIdentificada, que es el único
// constructor del paquete: así nace identificada aunque la origine el bucle y no
// quien llama (FR-009, D3).
//
// El formato lo pone aquí, en cada salto, porque es de la petición pedida y no de
// la cadena: cada salto es una petición completa que no hereda nada del anterior
// (D10), y la del robots.txt, que fabrica su propio escalón, no lo lleva
// (research D4 de H4).
func (c *Cliente) emitir(ctx context.Context, p Peticion, destino *url.URL) (recibida, *Error) {
	peticion, err := nuevaPeticionIdentificada(ctx, p.Metodo, destino.String())
	if err != nil {
		return recibida{}, errorDeArgumentos(p, err, "la dirección no se puede pedir")
	}

	if p.Acepta != "" {
		peticion.Header.Set(cabeceraDelFormato, p.Acepta)
	}

	comienzo := time.Now()

	respuesta, err := c.cliente.Do(peticion)
	if err != nil {
		return recibida{}, falloAlEmitir(ctx, p, err)
	}

	cuerpo, errDeLectura := io.ReadAll(respuesta.Body)

	if errAlCerrar := respuesta.Body.Close(); errAlCerrar != nil {
		c.registrador.DebugContext(ctx, "el cuerpo de la respuesta no se ha cerrado limpiamente",
			"url", destino.String(), "error", errAlCerrar.Error())
	}

	if errDeLectura != nil {
		return recibida{}, errorDeFuenteNoDisponible(p, respuesta.StatusCode, errDeLectura,
			"el cuerpo de la respuesta no se ha podido leer entero")
	}

	c.registrador.DebugContext(ctx, "petición emitida",
		"metodo", p.Metodo, "url", destino.String(),
		"estado", respuesta.StatusCode, "duracion", time.Since(comienzo))

	return recibida{
		estado:    respuesta.StatusCode,
		cabeceras: Cabeceras(respuesta.Header),
		cuerpo:    cuerpo,
	}, nil
}

// falloAlEmitir distingue las formas en que una petición puede no llegar a
// respuesta. Todas son de la misma clase —la fuente no sabe entregar el
// recurso—, pero no son lo mismo para quien lee el mensaje.
//
// Un escalón de la cadena que ya declaró su clase y su motivo se entrega tal
// cual: nadie desde fuera sabe mejor que él qué ocurrió, y volver a envolverlo
// cambiaría un motivo cierto —hoy el del ritmo, que se queda sin turno sin haber
// emitido nada (FR-022)— por la conjetura de un transporte que no llegó a
// intentarse. El contexto que vence o se cancela corta la operación en ese
// instante (FR-005, FR-029), y lo que queda es el sitio que no responde.
func falloAlEmitir(ctx context.Context, p Peticion, causa error) *Error {
	var declarado *Error
	if errors.As(causa, &declarado) {
		return declarado
	}

	if ctx.Err() != nil {
		return errorDeFuenteNoDisponible(p, 0, causa,
			"la operación ha terminado antes de recibir la respuesta")
	}

	return errorDeFuenteNoDisponible(p, 0, causa, "no se ha podido conectar con el sitio")
}

// esRedireccionSeguible dice si el estado es uno de los cinco que el cliente
// sigue, que son los mismos que sigue la biblioteca (go doc net/http.Client.Get).
// Cualquier otro 3xx —300, 304, 305— no se puede seguir y tampoco se entrega
// (FR-032, D10).
func esRedireccionSeguible(estado int) bool {
	switch estado {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	default:
		return false
	}
}

// siguienteDestino resuelve la dirección del salto sobre la del salto actual,
// que es lo que hace que una Location relativa funcione (RFC 3986 §5.2). Una
// redirección que no dice a dónde ir, o que lo dice con una dirección
// ininterpretable, no se puede seguir: es la fuente la que no sabe entregar el
// recurso (FR-011, D10).
func siguienteDestino(p Peticion, actual *url.URL, recibida recibida) (*url.URL, *Error) {
	destino := recibida.cabeceras.Get("Location")
	if destino == "" {
		return nil, errorDeFuenteNoDisponible(p, recibida.estado, nil,
			"la redirección no trae la cabecera Location y no dice a dónde ir")
	}

	referencia, err := url.Parse(destino)
	if err != nil {
		return nil, errorDeFuenteNoDisponible(p, recibida.estado, err,
			"la redirección apunta a una dirección que no se puede interpretar: "+destino)
	}

	return actual.ResolveReference(referencia), nil
}

// entregar clasifica la respuesta final de la cadena: el 429 y el 5xx tienen su
// clase, un 3xx que no se ha podido seguir no es entregable, y todo lo demás
// —el «no encontrado» incluido— se entrega a quien llama, que es quien sabe qué
// significa en su fuente (FR-029, FR-030, FR-032; tabla del contrato §3).
func entregar(p Peticion, direccion *url.URL, recibida recibida) (Respuesta, *Error) {
	switch {
	case recibida.estado == http.StatusTooManyRequests:
		return Respuesta{}, errorDe429(p, recibida.cabeceras, time.Now(),
			"la fuente ha alcanzado su límite de peticiones")

	case recibida.estado >= http.StatusInternalServerError:
		return Respuesta{}, errorDeFuenteNoDisponible(p, recibida.estado, nil,
			"el sitio no ha sabido entregar el recurso")

	case recibida.estado >= http.StatusMultipleChoices && recibida.estado < http.StatusBadRequest:
		return Respuesta{}, errorDeFuenteNoDisponible(p, recibida.estado, nil,
			"el sitio ha respondido con una redirección que no se puede seguir")

	default:
		return Respuesta{
			Peticion:  p,
			URL:       direccion.String(),
			Estado:    recibida.estado,
			Cabeceras: recibida.cabeceras,
			Cuerpo:    recibida.cuerpo,
		}, nil
	}
}
