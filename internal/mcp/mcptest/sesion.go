// Package mcptest lleva aquello con lo que los tests hablan con el servidor MCP
// de kitlegal por su entrada y su salida: el cliente del SDK del protocolo,
// que negocia la versión vigente, y uno de la especificación 2025-11-25
// escrito con la biblioteca estándar, que es la que negocia hoy la app de
// escritorio de Claude (H21 FR-007; contracts/arnes-e2e.md §3 y §5,
// research.md D12, V14 y V15).
//
// No es un paquete de test y solo lo usan tests: vive aquí porque el SDK solo
// se importa desde internal/mcp (regla R7), y el arnés e2e y la conformidad
// del servidor, que están en internal/app, necesitan un cliente (plan.md de
// H21, Complexity Tracking). El binario no lo enlaza.
//
// Abrir da una Sesion con cualquiera de los dos. Los dos comprueban por su
// cuenta el resultado de cada llamada, y devuelven un error si no se cumple:
// que lleva un solo bloque de contenido, de texto, que `structuredContent` es
// el mismo documento que ese texto y que `isError` es verdadero si y solo si
// el sobre lleva `ok` falso (FR-010, FR-011). El anterior lee además la salida
// del servidor línea a línea, las guarda todas y comprueba que cada una es un
// mensaje JSON-RPC 2.0 y que cada petición con `id` recibe su respuesta
// (FR-023).
package mcptest

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	// versionAnterior es la versión del protocolo que pide el cliente escrito
	// a mano.
	versionAnterior = "2025-11-25"

	// nombreDelCliente y versionDelCliente son lo que cada cliente dice de sí
	// mismo al servidor.
	nombreDelCliente  = "kitlegal-mcptest"
	versionDelCliente = "0"
)

// ErrSinSaludo es el error de Abrir cuando el servidor cierra la conexión —su
// salida llega al final o su entrada deja de admitir escrituras— antes de
// completar el saludo: es lo que hace un proceso que termina sin servir. Quien
// lo recibe lo distingue con errors.Is de cualquier otro fallo del saludo.
var ErrSinSaludo = errors.New("el servidor ha cerrado la conexión antes de completar el saludo")

var (
	// errSalidaTerminada es el error del cliente anterior cuando la salida del
	// servidor termina y él esperaba una línea más.
	errSalidaTerminada = errors.New("la salida del servidor ha terminado")

	// errEntradaCerrada es el del cliente anterior cuando la entrada del
	// servidor no admite una escritura.
	errEntradaCerrada = errors.New("la entrada del servidor no admite la escritura")

	// errSesionAbandonada es con lo que termina la lectura de la salida del
	// servidor cuando ya nadie va a pedir sus líneas.
	errSesionAbandonada = errors.New("la sesión ya no lee la salida del servidor")
)

// ErrorDeProtocolo es la respuesta del servidor a una petición que rechaza como
// error del protocolo, y no con un resultado: la llamada a una herramienta que
// no anuncia, por ejemplo. Es un error de Llamar distinto del fallo de una
// herramienta, que es un Resultado con Fallo.
type ErrorDeProtocolo struct {
	// Codigo y Mensaje son el `code` y el `message` del error JSON-RPC.
	Codigo  int64
	Mensaje string
}

// Error implementa error.
func (e *ErrorDeProtocolo) Error() string {
	return fmt.Sprintf("error del protocolo %d: %s", e.Codigo, e.Mensaje)
}

// Herramienta es una de las que el servidor anuncia en su lista.
type Herramienta struct {
	// Nombre y Descripcion son su `name` y su `description`.
	Nombre, Descripcion string

	// Entrada y Salida son su `inputSchema` y su `outputSchema`, en JSON:
	// con el cliente anterior, los bytes que el servidor escribió; con el del
	// SDK, que los entrega decodificados, el mismo documento vuelto a
	// serializar. Salida es nula si la herramienta no anuncia el suyo.
	Entrada, Salida json.RawMessage

	// SoloLectura dice si se anuncia con `readOnlyHint` verdadero.
	SoloLectura bool
}

// Resultado es el de una llamada que el servidor ha atendido, ya comprobado.
type Resultado struct {
	// Sobre es el texto de su único bloque de contenido, carácter a carácter:
	// el mismo documento que su `structuredContent`.
	Sobre []byte

	// Fallo dice si el resultado viene marcado como error de herramienta
	// (`isError`), que es tanto como que el sobre lleva `ok` falso.
	Fallo bool
}

// Sesion es una conexión abierta con un servidor MCP, por el cliente del SDK o
// por el anterior. Sus métodos se pueden llamar desde varias gorrutinas: con el
// cliente del SDK las llamadas van a la vez; con el anterior, que espera cada
// respuesta antes de enviar la petición siguiente, una detrás de otra.
type Sesion struct {
	protocolo, instrucciones string
	capacidades              []string

	// escribe es la entrada del servidor, que Cerrar cierra.
	escribe io.Closer

	cliente cliente
}

// cliente es lo que cada uno de los dos hace a su manera una vez hecho el
// saludo.
type cliente interface {
	// herramientas pide la lista entera, con todas sus páginas.
	herramientas(ctx context.Context) ([]Herramienta, error)

	// llamar hace una llamada y devuelve su resultado sin comprobar.
	llamar(ctx context.Context, herramienta string, argumentos json.RawMessage) (respuesta, error)

	// lineas son las que el servidor ha escrito, si el cliente las guarda.
	lineas() []string

	// terminar es lo que le queda por hacer con la entrada del servidor ya
	// cerrada.
	terminar(ctx context.Context) error
}

// Abrir conecta un cliente a un servidor MCP y hace el saludo: lee es la salida
// del servidor y escribe, su entrada. Sin anterior, el cliente es el del SDK,
// que negocia la versión vigente del protocolo; con anterior, uno de la
// especificación 2025-11-25 escrito con la biblioteca estándar, que hace
// `initialize` con su `protocolVersion` y envía la notificación `initialized`.
//
// Si el servidor cierra la conexión antes de completar el saludo, el error es
// ErrSinSaludo. Abrir no cierra nunca la entrada del servidor: sigue abierta
// hasta Cerrar, también si el saludo falla, de modo que un proceso que termina
// sin servir lo hace sin que nadie le haya cerrado la entrada.
//
// Con el cliente del SDK, ctx es el de la sesión entera: cuando termina, el
// SDK deja de leer. Las dos formas leen la salida del servidor en una gorrutina
// que acaba cuando la salida llega a su final; darle ese final —el proceso que
// termina, la tubería que se cierra— es cosa de quien la da.
func Abrir(ctx context.Context, lee io.Reader, escribe io.WriteCloser, anterior bool) (*Sesion, error) {
	if anterior {
		return abrirAnterior(ctx, lee, escribe)
	}

	return abrirVigente(ctx, lee, escribe)
}

// Protocolo es la versión del protocolo negociada en el saludo.
func (s *Sesion) Protocolo() string { return s.protocolo }

// Instrucciones son las `instructions` que el servidor dio en el saludo.
func (s *Sesion) Instrucciones() string { return s.instrucciones }

// Capacidades son los nombres de las capacidades que el servidor anunció en el
// saludo, ordenados. Con el cliente del SDK, las que el SDK conoce.
func (s *Sesion) Capacidades() []string { return slices.Clone(s.capacidades) }

// Herramientas pide al servidor su lista de herramientas y la devuelve entera,
// en el orden del servidor.
func (s *Sesion) Herramientas(ctx context.Context) ([]Herramienta, error) {
	return s.cliente.herramientas(ctx)
}

// Llamar llama a una herramienta con unos argumentos, que son un documento
// JSON, y devuelve su resultado una vez comprobado. Sin argumentos, el cliente
// anterior no envía `arguments` y el del SDK envía un objeto vacío.
//
// El error es un *ErrorDeProtocolo si el servidor rechaza la llamada como error
// del protocolo. Un resultado que no cumple lo que la sesión comprueba por su
// cuenta es también un error, que nombra la herramienta, y la sesión sigue.
func (s *Sesion) Llamar(ctx context.Context, herramienta string, argumentos json.RawMessage) (Resultado, error) {
	if len(argumentos) > 0 && !json.Valid(argumentos) {
		return Resultado{}, fmt.Errorf("los argumentos de %s no son un documento JSON: %s", herramienta, argumentos)
	}

	respuesta, err := s.cliente.llamar(ctx, herramienta, argumentos)
	if err != nil {
		return Resultado{}, err
	}

	resultado, err := respuesta.resultado()
	if err != nil {
		return Resultado{}, fmt.Errorf("el resultado de %s %w", herramienta, err)
	}

	return resultado, nil
}

// Lineas son las líneas que el servidor ha escrito en su salida hasta ahora,
// sin su salto, también la que haya hecho fallar una comprobación. Solo las
// guarda el cliente anterior; después de Cerrar están todas.
func (s *Sesion) Lineas() []string { return s.cliente.lineas() }

// Cerrar cierra la entrada del servidor, que es como un cliente le dice que ha
// terminado, y la sesión. Con el cliente anterior lee además la salida del
// servidor hasta su final, comprobando cada línea, y deja de esperarlo si ctx
// termina antes: es lo que ve un servidor que no termina tras cerrarle la
// entrada. Se llama una vez.
func (s *Sesion) Cerrar(ctx context.Context) error {
	err := s.escribe.Close()
	if err != nil {
		err = fmt.Errorf("cerrar la entrada del servidor: %w", err)
	}

	return errors.Join(err, s.cliente.terminar(ctx))
}

// respuesta es el resultado de una llamada tal como lo da cada cliente, antes
// de comprobarlo.
type respuesta struct {
	// bloques es cuántos bloques de contenido lleva; texto, el del primero, y
	// esTexto, si ese bloque es de texto.
	bloques int
	texto   string
	esTexto bool

	// estructurado es su `structuredContent` decodificado, nulo si no lo
	// lleva.
	estructurado any

	// fallo es su `isError`.
	fallo bool
}

// resultado comprueba la respuesta y la da como Resultado: un solo bloque de
// contenido, de texto; el texto, un documento JSON que es el mismo que
// `structuredContent`; e `isError` verdadero si y solo si ese documento lleva
// `ok` falso (contracts/arnes-e2e.md §3 de H21).
func (r respuesta) resultado() (Resultado, error) {
	if r.bloques != 1 {
		return Resultado{}, fmt.Errorf("lleva %d bloques de contenido y no uno", r.bloques)
	}

	if !r.esTexto {
		return Resultado{}, errors.New("no cumple: el bloque de contenido no es de texto")
	}

	var sobre any
	if err := json.Unmarshal([]byte(r.texto), &sobre); err != nil {
		return Resultado{}, fmt.Errorf("no cumple: el bloque de texto no es un documento JSON: %w", err)
	}

	if !reflect.DeepEqual(sobre, r.estructurado) {
		return Resultado{}, errors.New("no cumple: structuredContent no es el mismo documento que el bloque de texto")
	}

	// Un sobre que no es un objeto, o que no lleva `ok`, no lleva `ok` falso.
	campos, _ := sobre.(map[string]any)
	ok, lleva := campos["ok"].(bool)

	switch okFalso := lleva && !ok; {
	case r.fallo && !okFalso:
		return Resultado{}, errors.New("no cumple: isError es true y el sobre no lleva ok false")
	case !r.fallo && okFalso:
		return Resultado{}, errors.New("no cumple: isError es false y el sobre lleva ok false")
	}

	return Resultado{Sobre: []byte(r.texto), Fallo: r.fallo}, nil
}

// errorDelSaludo es el error de Abrir a partir del que da el cliente: el propio
// si el servidor cerró la conexión.
func errorDelSaludo(err error, cerrada bool) error {
	if cerrada {
		return fmt.Errorf("%w: %w", ErrSinSaludo, err)
	}

	return fmt.Errorf("saludo con el servidor: %w", err)
}

// vigente es el cliente del SDK del protocolo, que negocia la versión vigente
// y no deja pedirle otra (research.md V14 de H21).
type vigente struct {
	sesion *sdk.ClientSession
}

func abrirVigente(ctx context.Context, lee io.Reader, escribe io.WriteCloser) (*Sesion, error) {
	salida := &salidaDelServidor{lector: lee}
	entrada := &entradaDelServidor{escritor: escribe}

	cliente := sdk.NewClient(&sdk.Implementation{Name: nombreDelCliente, Version: versionDelCliente}, nil)

	sesion, err := cliente.Connect(ctx, &sdk.IOTransport{Reader: salida, Writer: entrada}, nil)
	if err != nil {
		return nil, errorDelSaludo(err, salida.terminada.Load() || entrada.rota.Load())
	}

	saludo := sesion.InitializeResult()

	capacidades, err := nombresDeLasCapacidades(saludo.Capabilities)
	if err != nil {
		return nil, errors.Join(errorDelSaludo(err, false), sesion.Close())
	}

	return &Sesion{
		protocolo:     saludo.ProtocolVersion,
		instrucciones: saludo.Instructions,
		capacidades:   capacidades,
		escribe:       escribe,
		cliente:       &vigente{sesion: sesion},
	}, nil
}

// nombresDeLasCapacidades da, ordenados, los nombres de las capacidades que el
// cliente del SDK entrega como una estructura: los de los campos que lleva.
func nombresDeLasCapacidades(capacidades *sdk.ServerCapabilities) ([]string, error) {
	documento, err := json.Marshal(capacidades)
	if err != nil {
		return nil, fmt.Errorf("serializar las capacidades: %w", err)
	}

	var porNombre map[string]json.RawMessage
	if err := json.Unmarshal(documento, &porNombre); err != nil {
		return nil, fmt.Errorf("leer las capacidades: %w", err)
	}

	return slices.Sorted(maps.Keys(porNombre)), nil
}

func (v *vigente) herramientas(ctx context.Context) ([]Herramienta, error) {
	var herramientas []Herramienta

	for anunciada, err := range v.sesion.Tools(ctx, nil) {
		if err != nil {
			return nil, fmt.Errorf("lista de herramientas: %w", err)
		}

		entrada, err := documentoDe(anunciada.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("esquema de entrada de %s: %w", anunciada.Name, err)
		}

		salida, err := documentoDe(anunciada.OutputSchema)
		if err != nil {
			return nil, fmt.Errorf("esquema de salida de %s: %w", anunciada.Name, err)
		}

		herramientas = append(herramientas, Herramienta{
			Nombre:      anunciada.Name,
			Descripcion: anunciada.Description,
			Entrada:     entrada,
			Salida:      salida,
			SoloLectura: anunciada.Annotations != nil && anunciada.Annotations.ReadOnlyHint,
		})
	}

	return herramientas, nil
}

// documentoDe da como JSON lo que el cliente del SDK entrega decodificado, y
// nada si no hay nada.
func documentoDe(valor any) (json.RawMessage, error) {
	if valor == nil {
		return nil, nil
	}

	return json.Marshal(valor)
}

func (v *vigente) llamar(ctx context.Context, herramienta string, argumentos json.RawMessage) (respuesta, error) {
	llamada := &sdk.CallToolParams{Name: herramienta}
	if len(argumentos) > 0 {
		// Solo si los hay: unos argumentos nulos dentro de la interfaz saldrían
		// por el cable como `null`, y sin ellos el SDK envía un objeto vacío.
		llamada.Arguments = argumentos
	}

	resultado, err := v.sesion.CallTool(ctx, llamada)
	if err != nil {
		var deProtocolo *jsonrpc.Error
		if errors.As(err, &deProtocolo) {
			return respuesta{}, &ErrorDeProtocolo{Codigo: deProtocolo.Code, Mensaje: deProtocolo.Message}
		}

		return respuesta{}, fmt.Errorf("llamada a %s: %w", herramienta, err)
	}

	recibida := respuesta{
		bloques:      len(resultado.Content),
		estructurado: resultado.StructuredContent,
		fallo:        resultado.IsError,
	}

	if recibida.bloques > 0 {
		if bloque, esTexto := resultado.Content[0].(*sdk.TextContent); esTexto {
			recibida.texto, recibida.esTexto = bloque.Text, true
		}
	}

	return recibida, nil
}

func (v *vigente) lineas() []string { return nil }

// terminar cierra la sesión del SDK, que no espera a nada: la entrada del
// servidor ya está cerrada y su salida no es suya.
func (v *vigente) terminar(context.Context) error { return v.sesion.Close() }

// salidaDelServidor es la salida del servidor tal como la lee el cliente del
// SDK, con una nota: si ha terminado. Es lo que deja a Abrir reconocer una
// conexión cerrada antes del saludo sin mirar el error del SDK.
type salidaDelServidor struct {
	lector io.Reader

	// terminada la escribe la gorrutina del SDK que lee y la lee Abrir.
	terminada atomic.Bool
}

// Read lee de la salida del servidor y anota su final.
func (s *salidaDelServidor) Read(p []byte) (int, error) {
	n, err := s.lector.Read(p)
	if err != nil {
		s.terminada.Store(true)
	}

	return n, err
}

// Close no cierra nada: la salida del servidor es de quien la da. El SDK lo
// llama al terminar la sesión.
func (s *salidaDelServidor) Close() error { return nil }

// entradaDelServidor es la entrada del servidor tal como la escribe el cliente
// del SDK, con una nota: si una escritura ha fallado.
type entradaDelServidor struct {
	escritor io.Writer

	// rota la escribe la gorrutina del SDK que escribe y la lee Abrir.
	rota atomic.Bool
}

// Write escribe en la entrada del servidor y anota si no se deja.
func (e *entradaDelServidor) Write(p []byte) (int, error) {
	n, err := e.escritor.Write(p)
	if err != nil {
		e.rota.Store(true)
	}

	return n, err
}

// Close no cierra nada: la entrada del servidor la cierra Cerrar, y no el SDK
// cuando un saludo falla.
func (e *entradaDelServidor) Close() error { return nil }

// anterior es el cliente de la especificación 2025-11-25, escrito con la
// biblioteca estándar: envía una petición, espera su respuesta leyendo la
// salida del servidor línea a línea, y guarda y comprueba cada línea que lee
// (research.md V15 de H21).
type anterior struct {
	escribe io.Writer

	// salida entrega las líneas de la salida del servidor, sin su salto, y se
	// cierra cuando no hay más; fin dice entonces por qué. Las escribe la
	// gorrutina que lee, que termina antes si alguien llama a abandonar.
	salida    <-chan []byte
	fin       error
	abandonar func()

	// cerrojo deja pasar una petición cada vez, y guarda lo que sigue.
	cerrojo sync.Mutex
	ultima  int64
	leidas  []string
}

// peticion es un mensaje JSON-RPC 2.0 del cliente: una petición si lleva `id`
// y una notificación si no.
type peticion struct {
	Version string `json:"jsonrpc"`
	ID      *int64 `json:"id,omitempty"`
	Metodo  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// errorDelCable es el `error` de una respuesta JSON-RPC.
type errorDelCable struct {
	Codigo  int64  `json:"code"`
	Mensaje string `json:"message"`
}

// saludoAnterior es lo que el cliente anterior lee del resultado de
// `initialize`.
type saludoAnterior struct {
	Protocolo     string                     `json:"protocolVersion"`
	Capacidades   map[string]json.RawMessage `json:"capabilities"`
	Instrucciones string                     `json:"instructions"`
}

// paginaDeHerramientas es el resultado de `tools/list`.
type paginaDeHerramientas struct {
	Herramientas []struct {
		Nombre      string          `json:"name"`
		Descripcion string          `json:"description"`
		Entrada     json.RawMessage `json:"inputSchema"`
		Salida      json.RawMessage `json:"outputSchema"`
		Anotaciones struct {
			SoloLectura bool `json:"readOnlyHint"`
		} `json:"annotations"`
	} `json:"tools"`
	Siguiente string `json:"nextCursor"`
}

// resultadoDeLlamada es el resultado de `tools/call`.
type resultadoDeLlamada struct {
	Contenido []struct {
		Tipo  string `json:"type"`
		Texto string `json:"text"`
	} `json:"content"`
	Estructurado any  `json:"structuredContent"`
	EsError      bool `json:"isError"`
}

func abrirAnterior(ctx context.Context, lee io.Reader, escribe io.WriteCloser) (*Sesion, error) {
	salida := make(chan []byte)
	abandonada := make(chan struct{})

	cliente := &anterior{
		escribe:   escribe,
		salida:    salida,
		abandonar: sync.OnceFunc(func() { close(abandonada) }),
	}

	go cliente.leer(lee, salida, abandonada)

	saludo, err := cliente.saludar(ctx)
	if err != nil {
		cliente.abandonar()

		return nil, errorDelSaludo(err, errors.Is(err, errSalidaTerminada) || errors.Is(err, errEntradaCerrada))
	}

	return &Sesion{
		protocolo:     saludo.Protocolo,
		instrucciones: saludo.Instrucciones,
		capacidades:   slices.Sorted(maps.Keys(saludo.Capacidades)),
		escribe:       escribe,
		cliente:       cliente,
	}, nil
}

// leer entrega por salida cada línea de la salida del servidor, la última
// aunque no acabe en salto, y cierra el canal cuando la salida termina o la
// sesión se abandona, con la causa en fin. Lee en su gorrutina para que quien
// espera una línea pueda dejar de esperarla con su contexto, y para que el
// servidor nunca se quede escribiendo sin nadie que lo lea.
func (a *anterior) leer(lee io.Reader, salida chan<- []byte, abandonada <-chan struct{}) {
	defer close(salida)

	lector := bufio.NewReader(lee)

	for {
		linea, err := lector.ReadBytes('\n')
		if len(linea) > 0 {
			select {
			case salida <- bytes.TrimSuffix(linea, []byte("\n")):
			case <-abandonada:
				a.fin = errSesionAbandonada

				return
			}
		}

		if err != nil {
			a.fin = err

			return
		}
	}
}

// saludar hace el saludo de la especificación 2025-11-25: `initialize` con su
// versión y, recibida la respuesta, la notificación `initialized`.
func (a *anterior) saludar(ctx context.Context) (saludoAnterior, error) {
	saludo, err := pedir[saludoAnterior](ctx, a, "initialize", map[string]any{
		"protocolVersion": versionAnterior,
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]string{"name": nombreDelCliente, "version": versionDelCliente},
	})
	if err != nil {
		return saludoAnterior{}, err
	}

	a.cerrojo.Lock()
	defer a.cerrojo.Unlock()

	return saludo, a.enviar(peticion{Metodo: "notifications/initialized"})
}

func (a *anterior) herramientas(ctx context.Context) ([]Herramienta, error) {
	var (
		herramientas []Herramienta
		desde        struct {
			Cursor string `json:"cursor,omitempty"`
		}
	)

	for {
		pagina, err := pedir[paginaDeHerramientas](ctx, a, "tools/list", desde)
		if err != nil {
			return nil, err
		}

		for _, anunciada := range pagina.Herramientas {
			herramientas = append(herramientas, Herramienta{
				Nombre:      anunciada.Nombre,
				Descripcion: anunciada.Descripcion,
				Entrada:     anunciada.Entrada,
				Salida:      anunciada.Salida,
				SoloLectura: anunciada.Anotaciones.SoloLectura,
			})
		}

		if pagina.Siguiente == "" {
			return herramientas, nil
		}

		desde.Cursor = pagina.Siguiente
	}
}

func (a *anterior) llamar(ctx context.Context, herramienta string, argumentos json.RawMessage) (respuesta, error) {
	resultado, err := pedir[resultadoDeLlamada](ctx, a, "tools/call", struct {
		Nombre     string          `json:"name"`
		Argumentos json.RawMessage `json:"arguments,omitempty"`
	}{Nombre: herramienta, Argumentos: argumentos})
	if err != nil {
		return respuesta{}, err
	}

	recibida := respuesta{
		bloques:      len(resultado.Contenido),
		estructurado: resultado.Estructurado,
		fallo:        resultado.EsError,
	}

	if recibida.bloques > 0 {
		recibida.texto, recibida.esTexto = resultado.Contenido[0].Texto, resultado.Contenido[0].Tipo == "text"
	}

	return recibida, nil
}

func (a *anterior) lineas() []string {
	a.cerrojo.Lock()
	defer a.cerrojo.Unlock()

	return slices.Clone(a.leidas)
}

// terminar lee la salida del servidor hasta su final, que llega cuando el
// servidor termina tras cerrarle la entrada, comprobando cada línea.
func (a *anterior) terminar(ctx context.Context) error {
	a.cerrojo.Lock()
	defer a.cerrojo.Unlock()

	defer a.abandonar()

	for {
		linea, err := a.siguiente(ctx)
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("leer la salida del servidor hasta su final: %w", err)
		}

		if _, err := a.anotar(linea); err != nil {
			return err
		}
	}
}

// pedir envía una petición del cliente anterior, espera su respuesta y da su
// resultado con la forma que esa petición espera; si la respuesta es un error
// del protocolo, un *ErrorDeProtocolo.
func pedir[T any](ctx context.Context, a *anterior, metodo string, params any) (T, error) {
	var respuesta struct {
		Resultado T              `json:"result"`
		Error     *errorDelCable `json:"error"`
	}

	linea, err := a.esperarRespuesta(ctx, metodo, params)
	if err != nil {
		return respuesta.Resultado, err
	}

	if err := json.Unmarshal(linea, &respuesta); err != nil {
		return respuesta.Resultado, fmt.Errorf("la respuesta a %s no tiene la forma de su petición: %w", metodo, err)
	}

	if respuesta.Error != nil {
		return respuesta.Resultado, &ErrorDeProtocolo{Codigo: respuesta.Error.Codigo, Mensaje: respuesta.Error.Mensaje}
	}

	return respuesta.Resultado, nil
}

// esperarRespuesta envía una petición con un `id` nuevo y lee la salida del
// servidor hasta la línea que le responde, que devuelve. Las que lee antes
// —notificaciones, peticiones del servidor, respuestas a otra petición— quedan
// guardadas y comprobadas, y no son la respuesta. Si la salida termina antes,
// la petición se ha quedado sin la suya.
func (a *anterior) esperarRespuesta(ctx context.Context, metodo string, params any) ([]byte, error) {
	a.cerrojo.Lock()
	defer a.cerrojo.Unlock()

	// Con el contexto ya terminado no sale nada: el select de más abajo
	// elegiría entre él y una respuesta que hubiera llegado.
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("petición %s: %w", metodo, err)
	}

	a.ultima++
	id := a.ultima

	if err := a.enviar(peticion{ID: &id, Metodo: metodo, Params: params}); err != nil {
		return nil, err
	}

	for {
		linea, err := a.siguiente(ctx)
		if errors.Is(err, errSalidaTerminada) {
			return nil, fmt.Errorf("la petición %d (%s) se ha quedado sin respuesta: %w", id, metodo, err)
		}

		if err != nil {
			return nil, fmt.Errorf("esperar la respuesta a la petición %d (%s): %w", id, metodo, err)
		}

		mensaje, err := a.anotar(linea)
		if err != nil {
			return nil, err
		}

		if _, esPeticion := mensaje["method"]; !esPeticion && string(mensaje["id"]) == strconv.FormatInt(id, 10) {
			return linea, nil
		}
	}
}

// enviar escribe un mensaje del cliente en la entrada del servidor, en una
// línea.
func (a *anterior) enviar(mensaje peticion) error {
	mensaje.Version = "2.0"

	linea, err := json.Marshal(mensaje)
	if err != nil {
		return fmt.Errorf("componer %s: %w", mensaje.Metodo, err)
	}

	if _, err := a.escribe.Write(append(linea, '\n')); err != nil {
		return fmt.Errorf("%w: %s: %w", errEntradaCerrada, mensaje.Metodo, err)
	}

	return nil
}

// siguiente da la línea siguiente de la salida del servidor, o deja de
// esperarla cuando el contexto termina. Con la salida terminada da
// errSalidaTerminada, con su causa: io.EOF si llegó a su final.
func (a *anterior) siguiente(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case linea, hay := <-a.salida:
		if !hay {
			return nil, fmt.Errorf("%w: %w", errSalidaTerminada, a.fin)
		}

		return linea, nil
	}
}

// anotar guarda una línea de la salida del servidor y comprueba que es un
// mensaje JSON-RPC 2.0: un objeto con `"jsonrpc":"2.0"`. Devuelve sus miembros.
// Un aviso, un registro de eventos o cualquier otra cosa que el servidor
// escriba ahí es un error (FR-023 de H21).
func (a *anterior) anotar(linea []byte) (map[string]json.RawMessage, error) {
	a.leidas = append(a.leidas, string(linea))

	var mensaje map[string]json.RawMessage
	if err := json.Unmarshal(linea, &mensaje); err != nil || string(mensaje["jsonrpc"]) != `"2.0"` {
		return nil, fmt.Errorf("el servidor ha escrito en su salida una línea que no es un mensaje JSON-RPC 2.0: %q", linea)
	}

	return mensaje, nil
}
