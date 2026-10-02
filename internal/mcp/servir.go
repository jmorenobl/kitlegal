package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// nombreDelServidor es el nombre con el que el servidor se presenta en
// `serverInfo` (contracts/servidor-mcp.md §5 de H21).
const nombreDelServidor = "kitlegal"

// Herramienta es una operación del servidor, ya hecha: el adaptador la anuncia
// y, cuando la llaman, le pasa los argumentos y devuelve lo que ella entrega
// (data-model §1 de H21).
type Herramienta struct {
	// Nombre es el nombre con el que se anuncia y se llama. Descripcion es lo
	// que el agente lee de ella.
	Nombre, Descripcion string

	// Entrada y Salida son sus dos esquemas, en JSON: se anuncian como
	// `inputSchema` y `outputSchema` con los bytes recibidos. El SDK exige que
	// el de entrada sea un objeto JSON con `type: object`.
	Entrada, Salida json.RawMessage

	// Llamar atiende una llamada con los argumentos tal como llegan del
	// cliente, sin comprobar: nulos si el cliente no envía ninguno. El servidor
	// atiende las llamadas a la vez, cada una en su gorrutina.
	Llamar func(argumentos json.RawMessage) Resultado
}

// Resultado es lo que una Herramienta entrega de una llamada (data-model §2 de
// H21).
type Resultado struct {
	// Sobre es el documento JSON de la llamada. Va dos veces en el resultado:
	// como texto, en su único bloque de contenido, y como `structuredContent`.
	Sobre []byte

	// Fallo marca el resultado como error de herramienta (`isError`).
	Fallo bool
}

// Servicio es lo que Servir necesita para atender una conexión (data-model §3
// de H21).
type Servicio struct {
	// Entrada y Salida son el transporte, y el único: por la primera llegan
	// los mensajes del cliente y por la segunda salen los del servidor, uno por
	// línea. Servir no cierra ninguna de las dos: son de quien se las da.
	Entrada io.Reader
	Salida  io.Writer

	// Version es la del binario, para `serverInfo`.
	Version string

	// Registrador recibe las líneas del SDK: sus errores y, si su nivel lo
	// admite, las que solo cuentan lo que hace. Con uno nulo se descartan.
	Registrador *slog.Logger

	// Herramientas son las que el servidor anuncia y atiende.
	Herramientas []Herramienta
}

// Servir atiende el protocolo MCP por la entrada y la salida del servicio hasta
// que la entrada se cierra o el contexto termina. Anuncia solo herramientas
// —`capabilities` es exactamente `{"tools":{}}`—, todas de solo lectura, con
// Instrucciones como `instructions` y `kitlegal` y la versión como
// `serverInfo`; las versiones del protocolo son las del SDK (FR-001, FR-005 a
// FR-008 de H21; research.md D7).
//
// Devuelve nil cuando la entrada llega a su fin, que es el final normal de un
// servidor de entrada y salida estándar, haya o no llamadas en curso: las que
// haya terminan y sus resultados no se escriben (FR-024). En cualquier otro
// caso devuelve el error con el que termina la sesión: una línea que no es un
// mensaje del protocolo, una salida rota o el contexto terminado.
//
// Que la entrada se cerró lo dice el lector, no el error: con una llamada en
// curso el SDK termina con un error propio que no envuelve io.EOF, y
// reconocerlo por su texto dependería de una cadena de un paquete interno suyo
// (research.md V10, V11).
func Servir(ctx context.Context, servicio Servicio) error {
	servidor := sdk.NewServer(
		&sdk.Implementation{Name: nombreDelServidor, Version: servicio.Version},
		&sdk.ServerOptions{
			Instructions: Instrucciones,
			Logger:       servicio.Registrador,
			// Con el campo Tools puesto y sin ListChanged, el servidor anuncia
			// `{"tools":{}}` y nada más: sin él añadiría `listChanged` al
			// registrar la primera herramienta, y sin Capabilities, `logging`
			// (research.md V8).
			Capabilities: &sdk.ServerCapabilities{Tools: &sdk.ToolCapabilities{}},
		},
	)

	for _, herramienta := range servicio.Herramientas {
		// La AddTool del servidor es la de bajo nivel: no valida ni la entrada
		// ni la salida, y deja componer el resultado. La genérica respondería
		// a unos argumentos inválidos con un texto suyo, y no con el sobre de
		// fallo que entrega la herramienta (research.md V7, D5).
		servidor.AddTool(&sdk.Tool{
			Name:         herramienta.Nombre,
			Description:  herramienta.Descripcion,
			InputSchema:  herramienta.Entrada,
			OutputSchema: herramienta.Salida,
			Annotations:  &sdk.ToolAnnotations{ReadOnlyHint: true},
		}, manejador(herramienta))
	}

	entrada := &lectorConFin{lector: servicio.Entrada}

	err := servidor.Run(ctx, &sdk.IOTransport{Reader: entrada, Writer: escritorAjeno{servicio.Salida}})
	if entrada.terminada() {
		return nil
	}

	return err
}

// manejador adapta una Herramienta a la firma del SDK: el resultado lleva el
// sobre en un solo bloque de texto y, como documento, en `structuredContent`, y
// va marcado como error solo si la herramienta lo dice. Nunca devuelve un
// error: para el SDK sería un error del protocolo, y el fallo de una
// herramienta es un resultado (FR-010, FR-011 de H21).
func manejador(herramienta Herramienta) sdk.ToolHandler {
	return func(_ context.Context, peticion *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
		resultado := herramienta.Llamar(peticion.Params.Arguments)

		return &sdk.CallToolResult{
			Content:           []sdk.Content{&sdk.TextContent{Text: string(resultado.Sobre)}},
			StructuredContent: json.RawMessage(resultado.Sobre),
			IsError:           resultado.Fallo,
		}, nil
	}
}

// lectorConFin es la entrada del servicio con una nota: si llegó a su fin. Es
// lo que deja a Servir distinguir el cierre de la entrada de cualquier otro
// final de la sesión sin mirar el error del SDK.
type lectorConFin struct {
	lector io.Reader

	// fin lo escribe la gorrutina del SDK que lee y lo lee Servir cuando Run
	// vuelve, que con una llamada en curso no es después de que aquella termine.
	fin atomic.Bool
}

// Read lee de la entrada y anota su fin.
func (l *lectorConFin) Read(p []byte) (int, error) {
	n, err := l.lector.Read(p)
	if errors.Is(err, io.EOF) {
		l.fin.Store(true)
	}

	return n, err
}

// Close no cierra nada: la entrada es de quien se la da al servicio. El SDK lo
// llama al terminar la sesión.
func (l *lectorConFin) Close() error { return nil }

// terminada dice si la entrada llegó a su fin.
func (l *lectorConFin) terminada() bool { return l.fin.Load() }

// escritorAjeno es la salida del servicio con el Close que el transporte del
// SDK exige, que no cierra nada: la salida es de quien se la da al servicio.
type escritorAjeno struct{ io.Writer }

// Close no cierra nada.
func (escritorAjeno) Close() error { return nil }
