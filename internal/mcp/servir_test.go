package mcp

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"slices"
	"sync"
	"testing"
	"testing/synctest"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// versionDePrueba es la versión del binario que los tests dan al servidor, para
// ver que es la que llega a `serverInfo`.
const versionDePrueba = "1.2.3-prueba"

// Los esquemas y los documentos de las herramientas de ejemplo. El adaptador no
// sabe de applets ni de sobres: recibe bytes y los sirve. Las claves de los
// esquemas van en un orden que no es el alfabético, para que una copia que
// pasara por un mapa —y no los bytes recibidos— se viera; y el documento de
// éxito lleva los tres caracteres que el codificador del SDK escapa en el cable
// (research.md V16).
const (
	entradaDeEjemplo = `{"properties":{"texto":{"type":"string"}},"additionalProperties":false,` +
		`"type":"object","required":["texto"]}`
	salidaDeEjemplo = `{"properties":{"ok":{"type":"boolean"},"data":true},"additionalProperties":false,` +
		`"type":"object","required":["ok","data"]}`
	entradaSinArgumentos = `{"additionalProperties":false,"type":"object"}`
	salidaSinRestringir  = `{"type":"object","required":["ok"]}`

	sobreDeExito = `{"ok":true,"fuente":"kitlegal.ejemplo","data":{"texto":"a < b & «c» é"}}`
	sobreDeFallo = `{"ok":false,"fuente":"kitlegal.cli","data":{"clase":"no-encontrado","mensaje":"no está"}}`
)

// soloLectura son las anotaciones con las que el SDK anuncia una herramienta
// registrada con ReadOnlyHint verdadero (contracts/servidor-mcp.md §2).
const soloLectura = `{"idempotentHint":false,"readOnlyHint":true}`

// TestServir comprueba el adaptador con un cliente del SDK conectado por dos
// tuberías en proceso: lo que el servidor anuncia, el resultado de una llamada
// que termina bien y el de una que falla, la herramienta que no está y el final
// con la entrada cerrada (FR-001, FR-005, FR-006, FR-007, FR-008, FR-010,
// FR-011, FR-024).
func TestServir(t *testing.T) {
	t.Parallel()

	t.Run("anuncia solo herramientas, con las instrucciones, el nombre y la versión", testAnuncio)
	t.Run("anuncia cada herramienta con su nombre, su descripción, sus dos esquemas byte a byte y solo lectura",
		testHerramientasAnunciadas)
	t.Run("una llamada que termina bien: un bloque de texto con el sobre y el mismo documento estructurado",
		testLlamadaQueTerminaBien)
	t.Run("una llamada que falla: el sobre de fallo, marcado como error de herramienta", testLlamadaQueFalla)
	t.Run("una herramienta que no está la responde el protocolo, sin llamar a ninguna", testHerramientaQueNoEsta)
	t.Run("nil al cerrarse la entrada sin llamadas", testEntradaCerradaSinLlamadas)
	t.Run("nil al cerrarse la entrada con una llamada en curso, que termina", testEntradaCerradaConUnaLlamadaEnCurso)
	t.Run("el error de Run si la entrada no se ha cerrado", testEntradaQueNoEsDelProtocolo)
}

func testAnuncio(t *testing.T) {
	t.Parallel()

	servidor := arrancar(t, herramientasDeEjemplo(&argumentosRecibidos{}))

	inicio := servidor.sesion.InitializeResult()
	require.NotNil(t, inicio)
	assert.Equal(t, Instrucciones, inicio.Instructions)
	require.NotNil(t, inicio.ServerInfo)
	assert.Equal(t, "kitlegal", inicio.ServerInfo.Name)
	assert.Equal(t, versionDePrueba, inicio.ServerInfo.Version)

	anunciado := servidor.escrito.resultadoCon(t, "capabilities")
	exigeLosBytes(t, `{"tools":{}}`, anunciado["capabilities"],
		"las capacidades: ni listChanged, ni logging, ni prompts, ni resources, ni completions (FR-008)")

	var versiones []string
	require.NoError(t, json.Unmarshal(anunciado["supportedVersions"], &versiones))
	assert.Contains(t, versiones, "2026-07-28", "la especificación vigente (FR-007)")
	assert.Contains(t, versiones, "2025-11-25", "la que negocia hoy la app de escritorio de Claude (FR-007)")
}

func testHerramientasAnunciadas(t *testing.T) {
	t.Parallel()

	servidor := arrancar(t, herramientasDeEjemplo(&argumentosRecibidos{}))

	listado, err := servidor.sesion.ListTools(t.Context(), nil)
	require.NoError(t, err)
	require.Len(t, listado.Tools, 2)

	for _, herramienta := range listado.Tools {
		require.NotNil(t, herramienta.Annotations, "%s se anuncia con sus anotaciones", herramienta.Name)
		assert.True(t, herramienta.Annotations.ReadOnlyHint, "%s se anuncia de solo lectura (FR-005)", herramienta.Name)
	}

	type anunciada struct {
		Name         string          `json:"name"`
		Description  string          `json:"description"`
		InputSchema  json.RawMessage `json:"inputSchema"`
		OutputSchema json.RawMessage `json:"outputSchema"`
		Annotations  json.RawMessage `json:"annotations"`
	}

	var enElCable []anunciada
	require.NoError(t, json.Unmarshal(servidor.escrito.resultadoCon(t, "tools")["tools"], &enElCable))

	// El orden es el del SDK, por nombre, y no el del registro.
	assert.Equal(t, []anunciada{
		{
			Name:         "ejemplo_eco",
			Description:  "Devuelve el documento de éxito.",
			InputSchema:  json.RawMessage(entradaDeEjemplo),
			OutputSchema: json.RawMessage(salidaDeEjemplo),
			Annotations:  json.RawMessage(soloLectura),
		},
		{
			Name:         "ejemplo_fallo",
			Description:  "Devuelve el documento de fallo.",
			InputSchema:  json.RawMessage(entradaSinArgumentos),
			OutputSchema: json.RawMessage(salidaSinRestringir),
			Annotations:  json.RawMessage(soloLectura),
		},
	}, enElCable, "cada esquema llega con los bytes recibidos")
}

func testLlamadaQueTerminaBien(t *testing.T) {
	t.Parallel()

	recibidos := &argumentosRecibidos{}
	servidor := arrancar(t, herramientasDeEjemplo(recibidos))

	resultado, err := servidor.sesion.CallTool(t.Context(), &sdk.CallToolParams{
		Name:      "ejemplo_eco",
		Arguments: map[string]any{"texto": "hola"},
	})
	require.NoError(t, err)

	assert.False(t, resultado.IsError)
	exigeLosBytes(t, sobreDeExito, textoDelUnicoBloque(t, resultado), "el texto del bloque, que es el sobre recibido")
	assert.JSONEq(t, sobreDeExito, documentoEstructurado(t, resultado))

	enElCable := servidor.escrito.resultadoCon(t, "content")
	assert.NotContains(t, enElCable, "isError", "una llamada que termina bien no lleva isError")
	assert.JSONEq(t, sobreDeExito, string(enElCable["structuredContent"]))

	llamadas := recibidos.todos()
	require.Len(t, llamadas, 1, "una llamada, una vez")
	assert.JSONEq(t, `{"texto":"hola"}`, llamadas[0], "los argumentos llegan como los envió el cliente")
}

func testLlamadaQueFalla(t *testing.T) {
	t.Parallel()

	servidor := arrancar(t, herramientasDeEjemplo(&argumentosRecibidos{}))

	resultado, err := servidor.sesion.CallTool(t.Context(), &sdk.CallToolParams{Name: "ejemplo_fallo"})
	require.NoError(t, err, "un fallo de la herramienta no es un error del protocolo")

	assert.True(t, resultado.IsError)
	exigeLosBytes(t, sobreDeFallo, textoDelUnicoBloque(t, resultado), "el texto del bloque, que es el sobre recibido")
	assert.JSONEq(t, sobreDeFallo, documentoEstructurado(t, resultado))

	enElCable := servidor.escrito.resultadoCon(t, "content")
	exigeLosBytes(t, "true", enElCable["isError"], "isError")
	exigeLosBytes(t, sobreDeFallo, enElCable["structuredContent"], "el documento estructurado, que es el sobre recibido")
}

func testHerramientaQueNoEsta(t *testing.T) {
	t.Parallel()

	recibidos := &argumentosRecibidos{}
	servidor := arrancar(t, herramientasDeEjemplo(recibidos))

	_, err := servidor.sesion.CallTool(t.Context(), &sdk.CallToolParams{Name: "ejemplo_ausente"})

	var deProtocolo *jsonrpc.Error
	require.ErrorAs(t, err, &deProtocolo)
	assert.Equal(t, int64(jsonrpc.CodeInvalidParams), deProtocolo.Code)
	assert.Equal(t, `unknown tool "ejemplo_ausente"`, deProtocolo.Message)
	assert.Empty(t, recibidos.todos(), "ninguna herramienta recibe la llamada")
}

func testEntradaCerradaSinLlamadas(t *testing.T) {
	t.Parallel()

	servidor := arrancar(t, herramientasDeEjemplo(&argumentosRecibidos{}))

	require.NoError(t, servidor.entrada.Close())

	assert.NoError(t, <-servidor.final, "la entrada que se cierra es el final normal del servidor (FR-024)")
}

// testEntradaCerradaConUnaLlamadaEnCurso va dentro de una burbuja de synctest
// para que el orden sea el del caso y no dependa del planificador: cuando
// synctest.Wait vuelve, el servidor ya ha visto el fin de su entrada y solo le
// queda esperar a la llamada, que sigue dentro de la herramienta. Ahí el SDK
// termina con un error que no envuelve io.EOF (research.md V11), y Servir tiene
// que dar nil igualmente.
func testEntradaCerradaConUnaLlamadaEnCurso(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		enCurso := make(chan struct{})
		soltar := make(chan struct{})

		servidor := arrancar(t, []Herramienta{{
			Nombre:      "ejemplo_lenta",
			Descripcion: "Espera a que el test la suelte.",
			Entrada:     json.RawMessage(entradaSinArgumentos),
			Salida:      json.RawMessage(salidaSinRestringir),
			Llamar: func(json.RawMessage) Resultado {
				close(enCurso)
				<-soltar

				return Resultado{Sobre: []byte(sobreDeExito)}
			},
		}})

		respuesta := make(chan error, 1)

		go func() {
			_, err := servidor.sesion.CallTool(t.Context(), &sdk.CallToolParams{Name: "ejemplo_lenta"})
			respuesta <- err
		}()

		<-enCurso
		require.NoError(t, servidor.entrada.Close())
		synctest.Wait()

		select {
		case err := <-servidor.final:
			require.Failf(t, "Servir ha vuelto con la llamada en curso", "devolvió %v", err)
		default:
		}

		close(soltar)
		require.NoError(t, <-servidor.final,
			"la entrada cerrada es el final normal también con una llamada en curso (FR-024)")

		// El proceso que termina cierra su salida: el cliente, que ya no
		// escuchaba, se queda sin respuesta.
		require.NoError(t, servidor.salida.Close())
		require.Error(t, <-respuesta, "el resultado de la llamada en curso no se escribe")
		assert.Empty(t, servidor.escrito.resultadosCon(t, "content"))
	})
}

func testEntradaQueNoEsDelProtocolo(t *testing.T) {
	t.Parallel()

	entrada, haciaElServidor := io.Pipe()

	t.Cleanup(func() { assert.NoError(t, haciaElServidor.Close()) })

	// La línea se escribe y la tubería sigue abierta: si Servir vuelve, no es
	// porque la entrada haya llegado a su fin.
	escrita := make(chan error, 1)

	go func() {
		_, err := haciaElServidor.Write([]byte("esto no es un mensaje del protocolo\n"))
		escrita <- err
	}()

	registro := &lineasEscritas{}

	err := Servir(t.Context(), Servicio{
		Entrada:      entrada,
		Salida:       io.Discard,
		Version:      versionDePrueba,
		Registrador:  slog.New(slog.NewTextHandler(registro, &slog.HandlerOptions{Level: slog.LevelError})),
		Herramientas: herramientasDeEjemplo(&argumentosRecibidos{}),
	})

	require.Error(t, err, "con la entrada abierta, lo que termina el servidor es un fallo (research.md V13)")
	require.NoError(t, <-escrita)
	assert.Contains(t, string(registro.copia()), "level=ERROR",
		"los errores del SDK llegan al registrador del servicio (research.md V19)")
}

// servidorDePrueba es un Servir en marcha con un cliente del SDK conectado a
// él, y los dos extremos que el test cierra como los cerraría el proceso de
// cada lado.
type servidorDePrueba struct {
	// sesion es la del cliente del SDK, ya conectado.
	sesion *sdk.ClientSession

	// escrito es todo lo que el servidor ha escrito en su salida.
	escrito *lineasEscritas

	// entrada es el extremo de escritura de la entrada del servidor: cerrarlo
	// es lo que hace el agente que deja de usarlo.
	entrada *io.PipeWriter

	// salida es el extremo de escritura de la salida del servidor: cerrarlo es
	// lo que pasa cuando su proceso termina.
	salida *io.PipeWriter

	// final recibe lo que devuelve Servir.
	final <-chan error
}

// arrancar pone a servir las herramientas y conecta un cliente del SDK por dos
// tuberías en proceso: ni descriptores del sistema ni red. Al terminar el test
// cierra los cuatro extremos, de modo que ninguna gorrutina quede esperando.
func arrancar(t *testing.T, herramientas []Herramienta) *servidorDePrueba {
	t.Helper()

	entradaDelServidor, haciaElServidor := io.Pipe()
	desdeElServidor, salidaDelServidor := io.Pipe()

	t.Cleanup(func() {
		for _, extremo := range []io.Closer{haciaElServidor, salidaDelServidor, entradaDelServidor, desdeElServidor} {
			assert.NoError(t, extremo.Close())
		}
	})

	escrito := &lineasEscritas{}
	final := make(chan error, 1)

	go func() {
		final <- Servir(t.Context(), Servicio{
			Entrada:      entradaDelServidor,
			Salida:       io.MultiWriter(escrito, salidaDelServidor),
			Version:      versionDePrueba,
			Herramientas: herramientas,
		})
	}()

	cliente := sdk.NewClient(&sdk.Implementation{Name: "cliente-de-prueba", Version: "0"}, nil)

	sesion, err := cliente.Connect(t.Context(), &sdk.IOTransport{Reader: desdeElServidor, Writer: haciaElServidor}, nil)
	require.NoError(t, err)

	return &servidorDePrueba{
		sesion:  sesion,
		escrito: escrito,
		entrada: haciaElServidor,
		salida:  salidaDelServidor,
		final:   final,
	}
}

// herramientasDeEjemplo son dos herramientas, dadas en un orden que no es el de
// sus nombres: una que termina bien y anota los argumentos que recibe, y una
// que falla.
func herramientasDeEjemplo(recibidos *argumentosRecibidos) []Herramienta {
	return []Herramienta{
		{
			Nombre:      "ejemplo_fallo",
			Descripcion: "Devuelve el documento de fallo.",
			Entrada:     json.RawMessage(entradaSinArgumentos),
			Salida:      json.RawMessage(salidaSinRestringir),
			Llamar: func(argumentos json.RawMessage) Resultado {
				recibidos.anotar(argumentos)

				return Resultado{Sobre: []byte(sobreDeFallo), Fallo: true}
			},
		},
		{
			Nombre:      "ejemplo_eco",
			Descripcion: "Devuelve el documento de éxito.",
			Entrada:     json.RawMessage(entradaDeEjemplo),
			Salida:      json.RawMessage(salidaDeEjemplo),
			Llamar: func(argumentos json.RawMessage) Resultado {
				recibidos.anotar(argumentos)

				return Resultado{Sobre: []byte(sobreDeExito)}
			},
		},
	}
}

// argumentosRecibidos anota los argumentos de cada llamada que llega a una
// herramienta de ejemplo. El servidor atiende las llamadas cada una en su
// gorrutina, así que se guardan bajo un cerrojo.
type argumentosRecibidos struct {
	cerrojo    sync.Mutex
	argumentos []string
}

func (r *argumentosRecibidos) anotar(argumentos json.RawMessage) {
	r.cerrojo.Lock()
	defer r.cerrojo.Unlock()

	r.argumentos = append(r.argumentos, string(argumentos))
}

func (r *argumentosRecibidos) todos() []string {
	r.cerrojo.Lock()
	defer r.cerrojo.Unlock()

	return slices.Clone(r.argumentos)
}

// lineasEscritas guarda lo que el servidor escribe en su salida, que el cliente
// lee a la vez por la tubería: es donde se ven los bytes del cable, que el
// cliente del SDK entrega ya decodificados. Sirve también de destino de un
// registrador. Quien escribe son gorrutinas del SDK, así que va bajo un cerrojo.
type lineasEscritas struct {
	cerrojo sync.Mutex
	bytes   bytes.Buffer
}

func (l *lineasEscritas) Write(p []byte) (int, error) {
	l.cerrojo.Lock()
	defer l.cerrojo.Unlock()

	return l.bytes.Write(p)
}

// copia devuelve lo escrito hasta ahora.
func (l *lineasEscritas) copia() []byte {
	l.cerrojo.Lock()
	defer l.cerrojo.Unlock()

	return bytes.Clone(l.bytes.Bytes())
}

// resultadosCon devuelve, en el orden en que se escribieron, los resultados de
// las respuestas del servidor que llevan la clave dada: `capabilities` es de
// la respuesta al descubrimiento, `tools` de la lista y `content` de una
// llamada.
func (l *lineasEscritas) resultadosCon(t *testing.T, clave string) []map[string]json.RawMessage {
	t.Helper()

	var resultados []map[string]json.RawMessage

	for linea := range bytes.Lines(l.copia()) {
		var mensaje struct {
			Result map[string]json.RawMessage `json:"result"`
		}

		require.NoError(t, json.Unmarshal(linea, &mensaje), "cada línea de la salida es un mensaje JSON: %s", linea)

		if _, lleva := mensaje.Result[clave]; lleva {
			resultados = append(resultados, mensaje.Result)
		}
	}

	return resultados
}

// resultadoCon devuelve el único resultado escrito que lleva la clave.
func (l *lineasEscritas) resultadoCon(t *testing.T, clave string) map[string]json.RawMessage {
	t.Helper()

	resultados := l.resultadosCon(t, clave)
	require.Len(t, resultados, 1, "el servidor ha escrito un resultado con %q, y solo uno", clave)

	return resultados[0]
}

// exigeLosBytes compara lo que llega del servidor con lo esperado byte a byte, y
// no como documentos JSON: lo que se comprueba es que el servidor no rehace lo
// que recibe.
func exigeLosBytes(t *testing.T, esperados string, obtenidos []byte, queSon string) {
	t.Helper()

	if esperados != string(obtenidos) {
		t.Errorf("%s: llega %s y no %s", queSon, obtenidos, esperados)
	}
}

// textoDelUnicoBloque devuelve el texto del resultado, que tiene que llevar un
// solo bloque y de texto.
func textoDelUnicoBloque(t *testing.T, resultado *sdk.CallToolResult) []byte {
	t.Helper()

	require.Len(t, resultado.Content, 1, "un solo bloque de contenido")

	bloque, esTexto := resultado.Content[0].(*sdk.TextContent)
	require.True(t, esTexto, "el bloque es de texto y no %T", resultado.Content[0])

	return []byte(bloque.Text)
}

// documentoEstructurado devuelve como JSON el `structuredContent` que el
// cliente del SDK ha decodificado.
func documentoEstructurado(t *testing.T, resultado *sdk.CallToolResult) string {
	t.Helper()

	documento, err := json.Marshal(resultado.StructuredContent)
	require.NoError(t, err)

	return string(documento)
}
