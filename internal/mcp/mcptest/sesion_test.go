package mcptest

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// instruccionesDePrueba son las `instructions` que dan los servidores de estos
// tests.
const instruccionesDePrueba = "Instrucciones del servidor de prueba."

// Los esquemas y los sobres de las herramientas de ejemplo. El sobre del eco
// lleva cifras, una lista y los tres caracteres que el codificador del SDK
// escapa en el cable (research.md V16 de H21): es lo que vería un cliente que
// comparase el texto con el documento estructurado de otro modo que como
// documentos.
const (
	entradaDelEco   = `{"type":"object","properties":{"texto":{"type":"string"}},"additionalProperties":false}`
	salidaDelEco    = `{"type":"object","required":["ok","data"]}`
	entradaDelFallo = `{"type":"object","additionalProperties":false}`

	sobreDeFallo = `{"ok":false,"fuente":"kitlegal.cli","data":{"clase":"no-encontrado","mensaje":"no está"}}`
)

// sobreDelEco es el sobre que la herramienta `eco` devuelve de unos argumentos:
// los propios argumentos, o `null` si no le llega ninguno.
func sobreDelEco(argumentos string) string {
	if argumentos == "" {
		argumentos = "null"
	}

	return `{"ok":true,"fuente":"kitlegal.ejemplo","data":{"argumentos":` + argumentos +
		`,"cifras":[1,2.5,-300],"texto":"a < b & «c» é"}}`
}

// clienteDePrueba es el cliente del SDK o el anterior, con lo que cada uno da
// de suyo.
type clienteDePrueba struct {
	nombre   string
	anterior bool

	// protocolo es la versión que negocia con un servidor del SDK.
	protocolo string

	// sinArgumentos es lo que le llega al servidor de una llamada sin
	// argumentos: el anterior no envía `arguments`, y el del SDK, un objeto
	// vacío.
	sinArgumentos string
}

func clientesDePrueba() []clienteDePrueba {
	return []clienteDePrueba{
		{nombre: "el cliente vigente", anterior: false, protocolo: "2026-07-28", sinArgumentos: "{}"},
		{nombre: "el cliente anterior", anterior: true, protocolo: "2025-11-25", sinArgumentos: ""},
	}
}

// TestSesion comprueba el cliente del SDK y el anterior contra servidores
// montados en proceso con el SDK sobre tuberías: lo que dan del saludo, de la lista de
// herramientas y de una llamada, y lo que comprueban por su cuenta —que el
// documento estructurado es el del único bloque de texto, que `isError`
// acompaña a un sobre con `ok` falso y, el anterior, que cada línea de la salida
// del servidor es un mensaje JSON-RPC 2.0 y que cada petición recibe su
// respuesta—, además del error propio de una conexión que se cierra antes del
// saludo (FR-007, FR-010, FR-011, FR-023, FR-071 a FR-073 de H21;
// contracts/arnes-e2e.md §3 y §5).
func TestSesion(t *testing.T) {
	t.Parallel()

	for _, cliente := range clientesDePrueba() {
		t.Run(cliente.nombre, func(t *testing.T) {
			t.Parallel()

			cliente.subpruebas(t)
		})
	}

	t.Run("el cliente anterior, línea a línea", testAnteriorLineaALinea)
}

// subpruebas registra lo que se comprueba con cada cliente, que es lo mismo.
func (c clienteDePrueba) subpruebas(t *testing.T) {
	t.Helper()

	t.Run("da el protocolo negociado, las instrucciones y las capacidades", c.testSaludo)
	t.Run("da las capacidades por su nombre y ordenadas", c.testCapacidadesOrdenadas)
	t.Run("da cada herramienta con su nombre, su descripción, sus esquemas y si es de solo lectura",
		c.testHerramientas)
	t.Run("da todas las herramientas de una lista que llega en varias páginas", c.testListaPaginada)
	t.Run("da el sobre de una llamada, con los argumentos enviados y sin ninguno", c.testLlamadas)
	t.Run("da el sobre de fallo de una llamada que falla, marcado", c.testLlamadaQueFalla)
	t.Run("da el error del protocolo de una herramienta que no está, y sigue", c.testErrorDeProtocolo)
	t.Run("rechaza unos argumentos que no son JSON sin enviar nada", c.testArgumentosIlegibles)
	t.Run("rechaza el resultado que no cumple lo que comprueba por su cuenta", c.testComprobaciones)
	t.Run("entrega a cada llamada simultánea su resultado", c.testLlamadasSimultaneas)
	t.Run("no llama con el contexto terminado", c.testContextoTerminado)
	t.Run("deja de esperar una respuesta cuando el contexto termina, y sigue", c.testContextoQueTermina)
	t.Run("al cerrar, cierra la entrada del servidor, que termina", c.testCierre)
	t.Run("da el fallo de cerrar la entrada del servidor", c.testCierreQueFalla)
	t.Run("da un error si el servidor termina con una llamada en curso", c.testSalidaCerradaEnUnaLlamada)
	t.Run("da su error propio si el servidor termina sin leer el saludo", c.testSinSaludoSinLeer)
	t.Run("da su error propio si el servidor lee el saludo y termina", c.testSinSaludoTrasLeer)
}

func (c clienteDePrueba) testSaludo(t *testing.T) {
	t.Parallel()

	sesion := c.abrir(t, montar(t, opcionesDelServidor{}, herramientasDeEjemplo(&argumentosRecibidos{})))

	assert.Equal(t, c.protocolo, sesion.Protocolo())
	assert.Equal(t, instruccionesDePrueba, sesion.Instrucciones())
	assert.Equal(t, []string{"tools"}, sesion.Capacidades())
}

func (c clienteDePrueba) testCapacidadesOrdenadas(t *testing.T) {
	t.Parallel()

	sesion := c.abrir(t, montar(t, opcionesDelServidor{capacidades: &sdk.ServerCapabilities{
		Tools:     &sdk.ToolCapabilities{},
		Resources: &sdk.ResourceCapabilities{},
		Prompts:   &sdk.PromptCapabilities{},
	}}, herramientasDeEjemplo(&argumentosRecibidos{})))

	assert.Equal(t, []string{"prompts", "resources", "tools"}, sesion.Capacidades())
}

func (c clienteDePrueba) testHerramientas(t *testing.T) {
	t.Parallel()

	sesion := c.abrir(t, montar(t, opcionesDelServidor{}, herramientasDeEjemplo(&argumentosRecibidos{})))

	herramientas, err := sesion.Herramientas(t.Context())
	require.NoError(t, err)
	require.Len(t, herramientas, 2)

	// El orden es el del servidor, que con el SDK es por nombre.
	eco, fallo := herramientas[0], herramientas[1]

	assert.Equal(t, "eco", eco.Nombre)
	assert.Equal(t, "Devuelve los argumentos que recibe.", eco.Descripcion)
	assert.JSONEq(t, entradaDelEco, string(eco.Entrada))
	assert.JSONEq(t, salidaDelEco, string(eco.Salida))
	assert.True(t, eco.SoloLectura, "eco se anuncia de solo lectura")

	assert.Equal(t, "fallo", fallo.Nombre)
	assert.Equal(t, "Devuelve el sobre de fallo.", fallo.Descripcion)
	assert.JSONEq(t, entradaDelFallo, string(fallo.Entrada))
	assert.Empty(t, fallo.Salida, "fallo se anuncia sin esquema de salida")
	assert.False(t, fallo.SoloLectura, "fallo se anuncia sin anotaciones")
}

func (c clienteDePrueba) testListaPaginada(t *testing.T) {
	t.Parallel()

	sesion := c.abrir(t, montar(t, opcionesDelServidor{pagina: 1}, herramientasDeEjemplo(&argumentosRecibidos{})))

	herramientas, err := sesion.Herramientas(t.Context())
	require.NoError(t, err)

	var nombres []string
	for _, herramienta := range herramientas {
		nombres = append(nombres, herramienta.Nombre)
	}

	assert.Equal(t, []string{"eco", "fallo"}, nombres, "las herramientas de las dos páginas, en su orden")
}

func (c clienteDePrueba) testLlamadas(t *testing.T) {
	t.Parallel()

	recibidos := &argumentosRecibidos{}
	sesion := c.abrir(t, montar(t, opcionesDelServidor{}, herramientasDeEjemplo(recibidos)))

	resultado, err := sesion.Llamar(t.Context(), "eco", json.RawMessage(`{"texto":"hola"}`))
	require.NoError(t, err)
	assert.Equal(t, Resultado{Sobre: []byte(sobreDelEco(`{"texto":"hola"}`))}, resultado,
		"el sobre es el texto del bloque, carácter a carácter")

	resultado, err = sesion.Llamar(t.Context(), "eco", nil)
	require.NoError(t, err)
	assert.Equal(t, Resultado{Sobre: []byte(sobreDelEco(c.sinArgumentos))}, resultado)

	assert.Equal(t, []string{`{"texto":"hola"}`, c.sinArgumentos}, recibidos.todos(),
		"al servidor le llegan los argumentos enviados y, sin ninguno, lo que cada cliente envía")
}

func (c clienteDePrueba) testLlamadaQueFalla(t *testing.T) {
	t.Parallel()

	sesion := c.abrir(t, montar(t, opcionesDelServidor{}, herramientasDeEjemplo(&argumentosRecibidos{})))

	resultado, err := sesion.Llamar(t.Context(), "fallo", json.RawMessage(`{}`))
	require.NoError(t, err, "el fallo de una herramienta es un resultado, no un error")
	assert.Equal(t, Resultado{Sobre: []byte(sobreDeFallo), Fallo: true}, resultado)
}

func (c clienteDePrueba) testErrorDeProtocolo(t *testing.T) {
	t.Parallel()

	recibidos := &argumentosRecibidos{}
	sesion := c.abrir(t, montar(t, opcionesDelServidor{}, herramientasDeEjemplo(recibidos)))

	_, err := sesion.Llamar(t.Context(), "ausente", json.RawMessage(`{}`))

	var deProtocolo *ErrorDeProtocolo
	require.ErrorAs(t, err, &deProtocolo)
	assert.Equal(t, int64(-32602), deProtocolo.Codigo)
	assert.Equal(t, `unknown tool "ausente"`, deProtocolo.Mensaje)
	assert.Empty(t, recibidos.todos(), "ninguna herramienta recibe la llamada")

	_, err = sesion.Llamar(t.Context(), "eco", json.RawMessage(`{"texto":"después"}`))
	require.NoError(t, err, "la sesión sigue después de un error del protocolo")
}

func (c clienteDePrueba) testArgumentosIlegibles(t *testing.T) {
	t.Parallel()

	recibidos := &argumentosRecibidos{}
	sesion := c.abrir(t, montar(t, opcionesDelServidor{}, herramientasDeEjemplo(recibidos)))

	_, err := sesion.Llamar(t.Context(), "eco", json.RawMessage(`{"texto":`))
	require.Error(t, err)

	var deProtocolo *ErrorDeProtocolo
	require.NotErrorAs(t, err, &deProtocolo, "no es una respuesta del servidor: la llamada no ha salido")

	_, err = sesion.Llamar(t.Context(), "eco", json.RawMessage(`{"texto":"después"}`))
	require.NoError(t, err, "la sesión sigue")
	assert.Equal(t, []string{`{"texto":"después"}`}, recibidos.todos())
}

func (c clienteDePrueba) testComprobaciones(t *testing.T) {
	t.Parallel()

	sesion := c.abrir(t, montar(t, opcionesDelServidor{}, herramientasQueIncumplen()))

	for _, caso := range []struct{ herramienta, mensaje string }{
		{"otro_documento", "structuredContent no es el mismo documento que el bloque de texto"},
		{"sin_estructurado", "structuredContent no es el mismo documento que el bloque de texto"},
		{"error_con_ok", "isError es true y el sobre no lleva ok false"},
		{"fallo_sin_error", "isError es false y el sobre lleva ok false"},
		{"dos_bloques", "lleva 2 bloques de contenido y no uno"},
		{"sin_bloques", "lleva 0 bloques de contenido y no uno"},
		{"imagen", "el bloque de contenido no es de texto"},
		{"texto_suelto", "el bloque de texto no es un documento JSON"},
	} {
		_, err := sesion.Llamar(t.Context(), caso.herramienta, json.RawMessage(`{}`))
		require.ErrorContains(t, err, caso.mensaje, caso.herramienta)
		require.ErrorContains(t, err, caso.herramienta, "el error nombra la herramienta")

		var deProtocolo *ErrorDeProtocolo
		require.NotErrorAs(t, err, &deProtocolo, "%s: lo que falla es el resultado, no el protocolo", caso.herramienta)
	}

	// La sesión sigue: lo comprobado es cada resultado, no la conexión.
	resultado, err := sesion.Llamar(t.Context(), "sobre_sin_ok", json.RawMessage(`{}`))
	require.NoError(t, err, "un sobre sin ok no lleva ok falso: sin isError, cumple")
	assert.Equal(t, Resultado{Sobre: []byte(`{"data":1}`)}, resultado)
}

func (c clienteDePrueba) testLlamadasSimultaneas(t *testing.T) {
	t.Parallel()

	const cuantas = 8

	sesion := c.abrir(t, montar(t, opcionesDelServidor{}, herramientasDeEjemplo(&argumentosRecibidos{})))

	resultados := make([]Resultado, cuantas)
	errores := make([]error, cuantas)

	var grupo sync.WaitGroup
	for n := range cuantas {
		grupo.Go(func() {
			resultados[n], errores[n] = sesion.Llamar(t.Context(), "eco", argumentosDe(n))
		})
	}

	grupo.Wait()

	for n := range cuantas {
		require.NoError(t, errores[n])
		assert.Equal(t, sobreDelEco(string(argumentosDe(n))), string(resultados[n].Sobre), "la llamada %d", n)
	}
}

func (c clienteDePrueba) testContextoTerminado(t *testing.T) {
	t.Parallel()

	recibidos := &argumentosRecibidos{}
	sesion := c.abrir(t, montar(t, opcionesDelServidor{}, herramientasDeEjemplo(recibidos)))

	terminado, cancelar := context.WithCancel(t.Context())
	cancelar()

	_, err := sesion.Llamar(terminado, "eco", json.RawMessage(`{}`))
	require.ErrorIs(t, err, context.Canceled)

	_, err = sesion.Herramientas(terminado)
	require.ErrorIs(t, err, context.Canceled)
}

func (c clienteDePrueba) testContextoQueTermina(t *testing.T) {
	t.Parallel()

	enCurso := make(chan struct{})
	soltar := make(chan struct{})

	servidor := montar(t, opcionesDelServidor{}, herramientasDeEjemplo(&argumentosRecibidos{}))
	servidor.registrar(&sdk.Tool{Name: "lenta", InputSchema: json.RawMessage(entradaDelFallo)},
		func(ctx context.Context, _ *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
			close(enCurso)

			// El cliente del SDK avisa al servidor de que ya no espera; el
			// anterior no, y la herramienta sigue hasta que el test la suelta.
			select {
			case <-soltar:
			case <-ctx.Done():
			}

			return resultadoCon(sobreDeFallo, sobreDeFallo, true), nil
		})

	sesion := c.abrir(t, servidor)

	hastaAqui, cancelar := context.WithCancel(t.Context())
	defer cancelar()

	final := make(chan error, 1)

	go func() {
		_, err := sesion.Llamar(hastaAqui, "lenta", json.RawMessage(`{}`))
		final <- err
	}()

	<-enCurso
	cancelar()
	require.ErrorIs(t, <-final, context.Canceled)

	// La respuesta que ya nadie espera llega ahora, y no es la de la llamada
	// siguiente.
	close(soltar)

	resultado, err := sesion.Llamar(t.Context(), "eco", json.RawMessage(`{"texto":"después"}`))
	require.NoError(t, err)
	assert.Equal(t, Resultado{Sobre: []byte(sobreDelEco(`{"texto":"después"}`))}, resultado)
}

func (c clienteDePrueba) testCierre(t *testing.T) {
	t.Parallel()

	servidor := montar(t, opcionesDelServidor{}, herramientasDeEjemplo(&argumentosRecibidos{}))
	sesion := c.abrir(t, servidor)

	_, err := sesion.Herramientas(t.Context())
	require.NoError(t, err)

	_, err = sesion.Llamar(t.Context(), "eco", json.RawMessage(`{"texto":"hola"}`))
	require.NoError(t, err)

	assert.Zero(t, servidor.escribe.cierres(), "la entrada del servidor sigue abierta mientras dura la sesión")

	require.NoError(t, sesion.Cerrar(t.Context()))
	assert.Equal(t, 1, servidor.escribe.cierres(), "Cerrar cierra la entrada del servidor")
	require.NoError(t, <-servidor.final, "el servidor termina porque su entrada se ha cerrado")

	if !c.anterior {
		assert.Empty(t, sesion.Lineas(), "el cliente del SDK no guarda las líneas")

		return
	}

	// Cerrar ha leído la salida del servidor hasta el final: están todas.
	escritas := servidor.escrito.lineas()
	require.Len(t, escritas, 3, "el saludo, la lista y la llamada")
	assert.Equal(t, escritas, sesion.Lineas(), "cada línea que el servidor ha escrito")
}

func (c clienteDePrueba) testCierreQueFalla(t *testing.T) {
	t.Parallel()

	servidor := montar(t, opcionesDelServidor{}, herramientasDeEjemplo(&argumentosRecibidos{}))
	servidor.escribe.falloAlCerrar = errors.New("no se deja cerrar")

	sesion := c.abrir(t, servidor)

	require.ErrorIs(t, sesion.Cerrar(t.Context()), servidor.escribe.falloAlCerrar)
}

func (c clienteDePrueba) testSalidaCerradaEnUnaLlamada(t *testing.T) {
	t.Parallel()

	servidor := montar(t, opcionesDelServidor{}, nil)
	servidor.registrar(&sdk.Tool{Name: "muda", InputSchema: json.RawMessage(entradaDelFallo)},
		func(context.Context, *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
			// El proceso que muere con la llamada en curso: su salida se cierra
			// sin la respuesta.
			servidor.terminar()

			return resultadoCon(sobreDeFallo, sobreDeFallo, true), nil
		})

	sesion := c.abrir(t, servidor)

	_, err := sesion.Llamar(t.Context(), "muda", json.RawMessage(`{}`))
	require.Error(t, err)

	var deProtocolo *ErrorDeProtocolo
	require.NotErrorAs(t, err, &deProtocolo)

	if c.anterior {
		require.ErrorContains(t, err, "se ha quedado sin respuesta", "cada petición con id recibe su respuesta")
	}
}

func (c clienteDePrueba) testSinSaludoSinLeer(t *testing.T) {
	t.Parallel()

	lee, escribe := conversar(t, func(_ *bufio.Reader, _ io.Writer) {})

	sesion, err := Abrir(t.Context(), lee, escribe, c.anterior)
	require.ErrorIs(t, err, ErrSinSaludo)
	assert.Nil(t, sesion)
	assert.Zero(t, escribe.cierres(), "quien abre no cierra la entrada del servidor: sigue abierta hasta que termina")
}

func (c clienteDePrueba) testSinSaludoTrasLeer(t *testing.T) {
	t.Parallel()

	lee, escribe := conversar(t, func(entrada *bufio.Reader, _ io.Writer) {
		_, err := entrada.ReadBytes('\n')
		assert.NoError(t, err, "el servidor lee la primera línea del cliente")
	})

	sesion, err := Abrir(t.Context(), lee, escribe, c.anterior)
	require.ErrorIs(t, err, ErrSinSaludo)
	assert.Nil(t, sesion)
	assert.Zero(t, escribe.cierres(), "quien abre no cierra la entrada del servidor: sigue abierta hasta que termina")
}

// testAnteriorLineaALinea reúne lo que solo hace el cliente anterior, que lee la
// salida del servidor línea a línea: comprobar que cada una es un mensaje
// JSON-RPC 2.0 (FR-023, FR-072 de H21) y que cada petición con `id` recibe su
// respuesta.
func testAnteriorLineaALinea(t *testing.T) {
	t.Parallel()

	t.Run("rechaza una línea que no es del protocolo antes del saludo", testLineaAjenaAntesDelSaludo)
	t.Run("rechaza una línea que no es del protocolo durante una llamada, y la guarda", testLineaAjenaEnUnaLlamada)
	t.Run("rechaza una línea que no es del protocolo tras cerrar la entrada", testLineaAjenaAlTerminar)
	t.Run("guarda las intercaladaes y las peticiones del servidor, que no son la respuesta", testMensajesIntercalados)
	t.Run("no toma por suya la respuesta a otra petición", testRespuestaAOtraPeticion)
	t.Run("da el error del protocolo con el que el servidor rechaza el saludo", testSaludoRechazado)
	t.Run("da un error si el resultado no tiene la forma de su petición", testResultadoConOtraForma)
	t.Run("da su error propio si el servidor responde al saludo y termina", testSinSaludoTrasResponder)
	t.Run("al cerrar, deja de esperar el final de la salida con el contexto", testCierreConLaSalidaAbierta)
}

func testLineaAjenaAntesDelSaludo(t *testing.T) {
	t.Parallel()

	for nombre, linea := range map[string]string{
		"un aviso en texto":              "aviso: las skills instaladas son de otra versión",
		"una línea vacía":                "",
		"un objeto sin jsonrpc":          `{"aviso":"las skills instaladas son de otra versión"}`,
		"un mensaje de otra versión":     `{"jsonrpc":"1.0","method":"notifications/message"}`,
		"un jsonrpc que no es la cadena": `{"jsonrpc":2.0,"method":"notifications/message"}`,
		"un lote de mensajes":            `[{"jsonrpc":"2.0","method":"notifications/message"}]`,
		"un valor suelto":                `null`,
	} {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			servidor := montar(t, opcionesDelServidor{antes: linea + "\n"}, herramientasDeEjemplo(&argumentosRecibidos{}))

			sesion, err := Abrir(t.Context(), servidor.lee, servidor.escribe, true)
			require.ErrorContains(t, err, "no es un mensaje JSON-RPC 2.0")
			require.ErrorContains(t, err, strconv.Quote(linea), "el error cita la línea")
			require.NotErrorIs(t, err, ErrSinSaludo, "el servidor no ha cerrado nada: ha escrito lo que no debía")
			assert.Nil(t, sesion)
		})
	}
}

func testLineaAjenaEnUnaLlamada(t *testing.T) {
	t.Parallel()

	const aviso = "aviso: una línea que no es del protocolo"

	servidor := montar(t, opcionesDelServidor{}, nil)
	servidor.registrar(&sdk.Tool{Name: "charlatana", InputSchema: json.RawMessage(entradaDelFallo)},
		func(context.Context, *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
			if _, err := io.WriteString(servidor.salida, aviso+"\n"); err != nil {
				t.Errorf("el servidor no ha podido escribir su aviso: %v", err)
			}

			return resultadoCon(sobreDeFallo, sobreDeFallo, true), nil
		})

	sesion, err := Abrir(t.Context(), servidor.lee, servidor.escribe, true)
	require.NoError(t, err)

	_, err = sesion.Llamar(t.Context(), "charlatana", json.RawMessage(`{}`))
	require.ErrorContains(t, err, "no es un mensaje JSON-RPC 2.0")

	lineas := sesion.Lineas()
	require.Len(t, lineas, 2, "el saludo y la línea que no es del protocolo")
	assert.Equal(t, aviso, lineas[1], "la línea queda guardada, para quien la quiera ver")
}

func testLineaAjenaAlTerminar(t *testing.T) {
	t.Parallel()

	servidor := montar(t, opcionesDelServidor{despues: "hasta luego\n"}, herramientasDeEjemplo(&argumentosRecibidos{}))

	sesion, err := Abrir(t.Context(), servidor.lee, servidor.escribe, true)
	require.NoError(t, err)

	require.ErrorContains(t, sesion.Cerrar(t.Context()), "no es un mensaje JSON-RPC 2.0")
	assert.Equal(t, "hasta luego", sesion.Lineas()[1])
}

func testMensajesIntercalados(t *testing.T) {
	t.Parallel()

	const (
		intercalada = `{"jsonrpc":"2.0","method":"notifications/message","params":{"level":"info","data":"x"}}`
		peticion    = `{"jsonrpc":"2.0","id":2,"method":"ping"}`
		lista       = `{"jsonrpc":"2.0","id":2,"result":{"tools":[{"name":"una","inputSchema":{"type":"object"}}]}}`
	)

	lee, escribe := conversar(t, guion(saludoDelGuion, intercalada+"\n"+peticion+"\n"+lista+"\n"))

	sesion, err := Abrir(t.Context(), lee, escribe, true)
	require.NoError(t, err)

	herramientas, err := sesion.Herramientas(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []Herramienta{{Nombre: "una", Entrada: json.RawMessage(`{"type":"object"}`)}}, herramientas)

	require.NoError(t, sesion.Cerrar(t.Context()))
	assert.Equal(t, []string{strings.TrimSuffix(saludoDelGuion, "\n"), intercalada, peticion, lista}, sesion.Lineas())
}

func testRespuestaAOtraPeticion(t *testing.T) {
	t.Parallel()

	lee, escribe := conversar(t, guion(saludoDelGuion, `{"jsonrpc":"2.0","id":7,"result":{"tools":[]}}`+"\n"))

	sesion, err := Abrir(t.Context(), lee, escribe, true)
	require.NoError(t, err)

	_, err = sesion.Herramientas(t.Context())
	require.ErrorContains(t, err, "se ha quedado sin respuesta")
}

func testSaludoRechazado(t *testing.T) {
	t.Parallel()

	lee, escribe := conversar(t, guion(`{"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"no"}}`+"\n"))

	sesion, err := Abrir(t.Context(), lee, escribe, true)

	var deProtocolo *ErrorDeProtocolo
	require.ErrorAs(t, err, &deProtocolo)
	assert.Equal(t, ErrorDeProtocolo{Codigo: -32602, Mensaje: "no"}, *deProtocolo)
	require.NotErrorIs(t, err, ErrSinSaludo, "el servidor ha respondido")
	assert.Nil(t, sesion)
}

func testResultadoConOtraForma(t *testing.T) {
	t.Parallel()

	lee, escribe := conversar(t, guion(saludoDelGuion, `{"jsonrpc":"2.0","id":2,"result":{"tools":"ninguna"}}`+"\n"))

	sesion, err := Abrir(t.Context(), lee, escribe, true)
	require.NoError(t, err)

	_, err = sesion.Herramientas(t.Context())
	require.ErrorContains(t, err, "no tiene la forma")
}

func testSinSaludoTrasResponder(t *testing.T) {
	t.Parallel()

	// El saludo no está completo hasta que el cliente envía su notificación, y
	// el servidor ya no está para leerla.
	lee, escribe := conversar(t, guion(saludoDelGuion))

	sesion, err := Abrir(t.Context(), lee, escribe, true)
	require.ErrorIs(t, err, ErrSinSaludo)
	assert.Nil(t, sesion)
}

func testCierreConLaSalidaAbierta(t *testing.T) {
	t.Parallel()

	servidor := montar(t, opcionesDelServidor{salidaAbierta: true}, herramientasDeEjemplo(&argumentosRecibidos{}))

	sesion, err := Abrir(t.Context(), servidor.lee, servidor.escribe, true)
	require.NoError(t, err)

	terminado, cancelar := context.WithCancel(t.Context())
	cancelar()

	require.ErrorIs(t, sesion.Cerrar(terminado), context.Canceled,
		"un servidor que no termina tras cerrarle la entrada no deja a Cerrar esperando")
	assert.Equal(t, 1, servidor.escribe.cierres(), "la entrada se cierra igualmente")
}

// abrir abre una sesión de este cliente con el servidor.
func (c clienteDePrueba) abrir(t *testing.T, servidor *servidorDePrueba) *Sesion {
	t.Helper()

	sesion, err := Abrir(t.Context(), servidor.lee, servidor.escribe, c.anterior)
	require.NoError(t, err)
	require.NotNil(t, sesion)

	return sesion
}

// opcionesDelServidor es lo que un servidor de prueba hace distinto del que
// sirve sus herramientas y termina cuando su entrada se cierra.
type opcionesDelServidor struct {
	// capacidades son las que anuncia; sin ellas, solo herramientas.
	capacidades *sdk.ServerCapabilities

	// pagina es el tamaño de página de sus listas; con cero, el del SDK.
	pagina int

	// antes y despues son lo que escribe en su salida antes de servir y cuando
	// termina de servir, fuera del protocolo.
	antes, despues string

	// salidaAbierta lo deja con la salida abierta cuando termina de servir,
	// como un proceso que no termina tras cerrarle la entrada.
	salidaAbierta bool
}

// servidorDePrueba es un servidor del SDK en marcha sobre dos tuberías en
// proceso, sin descriptores del sistema ni red, con los dos extremos que un
// cliente recibe.
type servidorDePrueba struct {
	// lee y escribe son los extremos del cliente: la salida del servidor y su
	// entrada.
	lee     *io.PipeReader
	escribe *entradaEspiada

	// salida es la del servidor: lo que se escribe en ella le llega al cliente
	// y queda en escrito.
	salida  io.Writer
	escrito *lineasEscritas

	// final recibe el error con el que el servidor termina de servir.
	final <-chan error

	registrar func(*sdk.Tool, sdk.ToolHandler)

	// terminar cierra los dos extremos del servidor, como el proceso que muere.
	terminar func()
}

// herramientaDePrueba es una herramienta con el manejador que la atiende.
type herramientaDePrueba struct {
	anuncio   *sdk.Tool
	manejador sdk.ToolHandler
}

// montar pone a servir las herramientas con el SDK. Al terminar el test cierra
// los cuatro extremos, de modo que ninguna gorrutina quede esperando.
func montar(t *testing.T, opciones opcionesDelServidor, herramientas []herramientaDePrueba) *servidorDePrueba {
	t.Helper()

	entradaDelServidor, haciaElServidor := io.Pipe()
	desdeElServidor, salidaDelServidor := io.Pipe()

	terminar := func() {
		assert.NoError(t, entradaDelServidor.Close())
		assert.NoError(t, salidaDelServidor.Close())
	}

	t.Cleanup(func() {
		terminar()
		assert.NoError(t, haciaElServidor.Close())
		assert.NoError(t, desdeElServidor.Close())
	})

	capacidades := opciones.capacidades
	if capacidades == nil {
		capacidades = &sdk.ServerCapabilities{Tools: &sdk.ToolCapabilities{}}
	}

	servidor := sdk.NewServer(&sdk.Implementation{Name: "servidor-de-prueba", Version: "0"}, &sdk.ServerOptions{
		Instructions: instruccionesDePrueba,
		Capabilities: capacidades,
		PageSize:     opciones.pagina,
	})

	for _, herramienta := range herramientas {
		servidor.AddTool(herramienta.anuncio, herramienta.manejador)
	}

	escrito := &lineasEscritas{}
	salida := io.MultiWriter(escrito, salidaDelServidor)
	final := make(chan error, 1)

	go func() {
		_, errAntes := io.WriteString(salida, opciones.antes)
		err := servidor.Run(t.Context(), &sdk.IOTransport{Reader: entradaDelServidor, Writer: sinCierre{salida}})
		_, errDespues := io.WriteString(salida, opciones.despues)

		if !opciones.salidaAbierta {
			terminar()
		}

		final <- errors.Join(errAntes, err, errDespues)
	}()

	return &servidorDePrueba{
		lee:       desdeElServidor,
		escribe:   &entradaEspiada{escritor: haciaElServidor},
		salida:    salida,
		escrito:   escrito,
		final:     final,
		registrar: servidor.AddTool,
		terminar:  terminar,
	}
}

// herramientasDeEjemplo son dos herramientas, dadas en un orden que no es el de
// sus nombres: una que falla, sin anotaciones ni esquema de salida, y una de
// solo lectura que devuelve los argumentos que recibe y los anota.
func herramientasDeEjemplo(recibidos *argumentosRecibidos) []herramientaDePrueba {
	return []herramientaDePrueba{
		{
			anuncio: &sdk.Tool{
				Name:        "fallo",
				Description: "Devuelve el sobre de fallo.",
				InputSchema: json.RawMessage(entradaDelFallo),
			},
			manejador: func(context.Context, *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
				return resultadoCon(sobreDeFallo, sobreDeFallo, true), nil
			},
		},
		{
			anuncio: &sdk.Tool{
				Name:         "eco",
				Description:  "Devuelve los argumentos que recibe.",
				InputSchema:  json.RawMessage(entradaDelEco),
				OutputSchema: json.RawMessage(salidaDelEco),
				Annotations:  &sdk.ToolAnnotations{ReadOnlyHint: true},
			},
			manejador: func(_ context.Context, peticion *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
				argumentos := string(peticion.Params.Arguments)
				recibidos.anotar(argumentos)

				sobre := sobreDelEco(argumentos)

				return resultadoCon(sobre, sobre, false), nil
			},
		},
	}
}

// herramientasQueIncumplen son las que devuelven un resultado que la sesión
// tiene que rechazar, cada una por una cosa, y una que cumple por poco.
func herramientasQueIncumplen() []herramientaDePrueba {
	const (
		bien = `{"ok":true,"data":1}`
		mal  = `{"ok":false,"data":{"clase":"inesperado","mensaje":"x"}}`
	)

	resultados := map[string]*sdk.CallToolResult{
		"otro_documento":   resultadoCon(bien, `{"ok":true,"data":2}`, false),
		"sin_estructurado": resultadoCon(bien, "", false),
		"error_con_ok":     resultadoCon(bien, bien, true),
		"fallo_sin_error":  resultadoCon(mal, mal, false),
		"texto_suelto":     resultadoCon("no encontrado", `"no encontrado"`, false),
		"sobre_sin_ok":     resultadoCon(`{"data":1}`, `{"data":1}`, false),
		"sin_bloques":      {Content: []sdk.Content{}, StructuredContent: json.RawMessage(bien)},
		"dos_bloques": {
			Content:           []sdk.Content{&sdk.TextContent{Text: bien}, &sdk.TextContent{Text: bien}},
			StructuredContent: json.RawMessage(bien),
		},
		"imagen": {
			Content:           []sdk.Content{&sdk.ImageContent{Data: []byte{1}, MIMEType: "image/png"}},
			StructuredContent: json.RawMessage(bien),
		},
	}

	herramientas := make([]herramientaDePrueba, 0, len(resultados))
	for nombre, resultado := range resultados {
		herramientas = append(herramientas, herramientaDePrueba{
			anuncio: &sdk.Tool{Name: nombre, InputSchema: json.RawMessage(entradaDelFallo)},
			manejador: func(context.Context, *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
				return resultado, nil
			},
		})
	}

	return herramientas
}

// resultadoCon compone el resultado de una llamada con un bloque de texto y,
// si se da, un documento estructurado.
func resultadoCon(texto, estructurado string, esError bool) *sdk.CallToolResult {
	resultado := &sdk.CallToolResult{Content: []sdk.Content{&sdk.TextContent{Text: texto}}, IsError: esError}
	if estructurado != "" {
		resultado.StructuredContent = json.RawMessage(estructurado)
	}

	return resultado
}

// argumentosDe son los argumentos de la llamada n de una tanda.
func argumentosDe(n int) json.RawMessage {
	return json.RawMessage(`{"texto":"llamada ` + strconv.Itoa(n) + `"}`)
}

// saludoDelGuion es la respuesta de un servidor de guion al `initialize` del
// cliente anterior, que es su primera petición.
const saludoDelGuion = `{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-11-25",` +
	`"capabilities":{"tools":{}},"serverInfo":{"name":"guion","version":"0"}}}` + "\n"

// guion es un servidor que no habla el protocolo: a cada petición con `id` que
// lee escribe la entrada siguiente de sus respuestas, tal cual, y cuando se le
// acaban termina. Es lo que deja ver qué hace el cliente anterior con lo que un
// servidor del SDK no escribe nunca.
func guion(respuestas ...string) func(*bufio.Reader, io.Writer) {
	return func(entrada *bufio.Reader, salida io.Writer) {
		for _, respuesta := range respuestas {
			for {
				linea, err := entrada.ReadBytes('\n')
				if err != nil {
					return
				}

				if bytes.Contains(linea, []byte(`"id":`)) {
					break
				}
			}

			if _, err := io.WriteString(salida, respuesta); err != nil {
				return
			}
		}
	}
}

// conversar pone en marcha un servidor escrito a mano sobre dos tuberías y
// devuelve los extremos del cliente. Cuando el servidor vuelve se cierran sus
// dos extremos, como los de un proceso que termina.
func conversar(t *testing.T, servidor func(entrada *bufio.Reader, salida io.Writer)) (*io.PipeReader, *entradaEspiada) {
	t.Helper()

	entradaDelServidor, haciaElServidor := io.Pipe()
	desdeElServidor, salidaDelServidor := io.Pipe()

	t.Cleanup(func() {
		for _, extremo := range []io.Closer{haciaElServidor, salidaDelServidor, entradaDelServidor, desdeElServidor} {
			assert.NoError(t, extremo.Close())
		}
	})

	go func() {
		servidor(bufio.NewReader(entradaDelServidor), salidaDelServidor)

		assert.NoError(t, entradaDelServidor.Close())
		assert.NoError(t, salidaDelServidor.Close())
	}()

	return desdeElServidor, &entradaEspiada{escritor: haciaElServidor}
}

// entradaEspiada es la entrada del servidor tal como la recibe el cliente, con
// la cuenta de las veces que se cierra.
type entradaEspiada struct {
	escritor *io.PipeWriter
	cerrada  atomic.Int32

	// falloAlCerrar, si se da, es lo que devuelve Close.
	falloAlCerrar error
}

func (e *entradaEspiada) Write(p []byte) (int, error) { return e.escritor.Write(p) }

func (e *entradaEspiada) Close() error {
	e.cerrada.Add(1)

	return errors.Join(e.falloAlCerrar, e.escritor.Close())
}

func (e *entradaEspiada) cierres() int { return int(e.cerrada.Load()) }

// sinCierre es la salida del servidor con el Close que el transporte del SDK
// exige, que no cierra nada: la cierra el test cuando el servidor termina.
type sinCierre struct{ io.Writer }

func (sinCierre) Close() error { return nil }

// argumentosRecibidos anota los argumentos de cada llamada que llega a una
// herramienta de ejemplo. El servidor atiende las llamadas cada una en su
// gorrutina, así que se guardan bajo un cerrojo.
type argumentosRecibidos struct {
	cerrojo    sync.Mutex
	argumentos []string
}

func (r *argumentosRecibidos) anotar(argumentos string) {
	r.cerrojo.Lock()
	defer r.cerrojo.Unlock()

	r.argumentos = append(r.argumentos, argumentos)
}

func (r *argumentosRecibidos) todos() []string {
	r.cerrojo.Lock()
	defer r.cerrojo.Unlock()

	return slices.Clone(r.argumentos)
}

// lineasEscritas guarda lo que el servidor escribe en su salida, que el cliente
// lee a la vez por la tubería. Quien escribe son gorrutinas del SDK, así que va
// bajo un cerrojo.
type lineasEscritas struct {
	cerrojo sync.Mutex
	bytes   bytes.Buffer
}

func (l *lineasEscritas) Write(p []byte) (int, error) {
	l.cerrojo.Lock()
	defer l.cerrojo.Unlock()

	return l.bytes.Write(p)
}

// lineas devuelve las líneas escritas hasta ahora, sin su salto.
func (l *lineasEscritas) lineas() []string {
	l.cerrojo.Lock()
	defer l.cerrojo.Unlock()

	var lineas []string
	for linea := range strings.Lines(l.bytes.String()) {
		lineas = append(lineas, strings.TrimSuffix(linea, "\n"))
	}

	return lineas
}
