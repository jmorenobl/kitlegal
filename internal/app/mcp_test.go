package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"maps"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/graph"
	"github.com/jmorenobl/kitlegal/internal/httpx"
	"github.com/jmorenobl/kitlegal/internal/mcp/mcptest"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// esperaDelFinal es lo más que un test espera a que un servidor arrancado en
// proceso termine o a que una invocación que no sirve vuelva. Es generosa: solo
// se agota cuando el servidor no termina, que es el defecto que se busca.
const esperaDelFinal = 30 * time.Second

// Las clases de un sobre de fallo que afirman estos tests, escritas aquí y no
// tomadas del vocabulario del kernel: son las del contrato
// (contracts/servidor-mcp.md §3 de H21; ADR 0023).
const (
	claseArgumentos         = "argumentos"
	claseNoEncontrado       = "no-encontrado"
	claseFuenteNoDisponible = "fuente-no-disponible"
)

// La norma y los bloques de las grabaciones de H4 con los que llaman estos
// tests, y lo que cada herramienta recibe de ellos.
const (
	normaDeLasLlamadas       = "BOE-A-2015-10565"
	argumentosDelBloqueA21   = `{"norma":"` + normaDeLasLlamadas + `","bloque":"a21"}`
	argumentosDelBloqueA9999 = `{"norma":"` + normaDeLasLlamadas + `","bloque":"a9999"}`
	argumentosDeLeganes      = `{"consulta":"Leganés"}`
)

// lineaDeEntregaFallida es como empieza la línea que el kernel deja en la
// salida de error cuando lo observado no llega al grafo (H7).
const lineaDeEntregaFallida = "kitlegal: lo observado no ha llegado al grafo del mundo: "

// horaDeLaReproduccion es la hora que declara cada respuesta que boe pide a la
// reproducción, y la de graph: con ella, la fecha de consulta de un sobre no
// depende de cuándo corre el test ni de si lo sirve la caché, y el sobre de una
// llamada se compara byte a byte con el de su orden.
func horaDeLaReproduccion() time.Time {
	return time.Date(2026, time.September, 28, 12, 0, 0, 0, time.UTC)
}

// salidaDeError es la salida de error de un servidor arrancado en proceso. Con
// `mcp serve` escriben en ella varias gorrutinas —el registro de eventos y cada
// llamada que falla—, así que el test le da una que lo admite, como lo admite
// el descriptor de un proceso.
type salidaDeError struct {
	cerrojo sync.Mutex
	escrito bytes.Buffer
}

func (s *salidaDeError) Write(p []byte) (int, error) {
	s.cerrojo.Lock()
	defer s.cerrojo.Unlock()

	return s.escrito.Write(p)
}

// lineas son las líneas escritas hasta ahora, sin su salto.
func (s *salidaDeError) lineas() []string {
	s.cerrojo.Lock()
	defer s.cerrojo.Unlock()

	texto := strings.TrimSuffix(s.escrito.String(), "\n")
	if texto == "" {
		return nil
	}

	return strings.Split(texto, "\n")
}

// lineasCon son las líneas que contienen el texto.
func (s *salidaDeError) lineasCon(texto string) []string {
	var con []string

	for _, linea := range s.lineas() {
		if strings.Contains(linea, texto) {
			con = append(con, linea)
		}
	}

	return con
}

// servidorEnProceso es `kitlegal mcp serve` atendido por app.Main en una
// gorrutina del test, con dos tuberías como entrada y salida estándar y un
// cliente de prueba conectado a ellas. En proceso, y no en un subproceso, para
// que el detector de carreras vea el código del servidor (research.md D13 de
// H21).
type servidorEnProceso struct {
	sesion *mcptest.Sesion
	// entrada es la entrada estándar del servidor, para el test que la cierra
	// por su cuenta, con una llamada en curso.
	entrada io.Closer
	errores *salidaDeError
	// terminado se cierra cuando app.Main vuelve, y codigo es lo que devuelve:
	// solo se lee después.
	terminado <-chan struct{}
	codigo    int
}

// registroCon es un registro local del test con esos applets.
func registroCon(t *testing.T, applets ...app.Applet) *app.Registro {
	t.Helper()

	registro := &app.Registro{}

	for _, applet := range applets {
		require.NoError(t, registro.Registrar(applet))
	}

	return registro
}

// arrancarElServidor registra en el registro el applet mcp con una tubería como
// entrada, arranca `mcp serve` con las banderas y conecta el cliente: el del
// SDK, que negocia la versión vigente, o el anterior, que además comprueba que
// cada línea de la salida estándar del servidor es un mensaje del protocolo.
// El applet lo registra aquí, y no quien compone el registro, porque su entrada
// es la tubería de este servidor.
func arrancarElServidor(t *testing.T, registro *app.Registro, anterior bool, banderas ...string) *servidorEnProceso {
	t.Helper()

	entradaLee, entradaEscribe := io.Pipe()
	salidaLee, salidaEscribe := io.Pipe()

	require.NoError(t, registro.Registrar(app.AppletMCP(app.DependenciasDeMCP{
		Entrada: entradaLee,
		Version: versionDelEjemplo,
	})))

	terminado := make(chan struct{})
	servidor := &servidorEnProceso{entrada: entradaEscribe, errores: &salidaDeError{}, terminado: terminado}

	argv := slices.Concat([]string{nombreDelBinario, "mcp", "serve"}, banderas)

	go func() {
		defer close(terminado)

		servidor.codigo = app.Main(argv, registro, salidaEscribe, servidor.errores,
			versionDelEjemplo, commitDelEjemplo, fechaDelEjemplo)

		// Lo que un proceso da al terminar: el final de su salida estándar, que
		// es lo que el cliente espera para dar por cerrada la sesión, y una
		// entrada que ya no admite escrituras, de modo que quien le envíe algo
		// después reciba un error en lugar de quedarse esperando a que lo lea.
		assert.NoError(t, salidaEscribe.Close())
		assert.NoError(t, entradaLee.Close())
	}()

	// Pase lo que pase en el test, el servidor termina antes que él: se le
	// cierra la entrada y se deja de leer su salida, de modo que ni espera un
	// mensaje ni se queda escribiendo uno que nadie lee.
	t.Cleanup(func() {
		assert.NoError(t, entradaEscribe.Close())
		assert.NoError(t, salidaLee.Close())
		<-terminado
	})

	sesion, err := mcptest.Abrir(t.Context(), salidaLee, entradaEscribe, anterior)
	require.NoError(t, err, "el saludo con el servidor: %s", servidor.errores.lineas())

	servidor.sesion = sesion

	return servidor
}

// llamar llama a una herramienta y exige que el servidor la atienda con un
// resultado: lo que la sesión comprueba por su cuenta —un bloque de texto con
// el sobre, el mismo documento estructurado y `isError` si y solo si `ok` es
// falso— vale para toda llamada de estos tests.
func (s *servidorEnProceso) llamar(t *testing.T, herramienta, argumentos string) mcptest.Resultado {
	t.Helper()

	resultado, err := s.pedir(t, herramienta, json.RawMessage(argumentos))
	require.NoError(t, err, "la llamada a %s con %s: %s", herramienta, argumentos, s.errores.lineas())

	return resultado
}

// pedir llama a una herramienta y devuelve lo que la sesión dé, sin esperar su
// respuesta más de esperaDelFinal: una llamada que el servidor deja sin
// responder es un fallo del test, y no un test que no termina.
func (s *servidorEnProceso) pedir(t *testing.T, herramienta string, argumentos json.RawMessage) (mcptest.Resultado, error) {
	t.Helper()

	ctx, cancelar := context.WithTimeout(t.Context(), esperaDelFinal)
	defer cancelar()

	return s.sesion.Llamar(ctx, herramienta, argumentos)
}

// cerrar cierra la entrada del servidor, como hace un cliente que ha terminado,
// y devuelve el código con el que el servidor termina.
func (s *servidorEnProceso) cerrar(t *testing.T) int {
	t.Helper()

	ctx, cancelar := context.WithTimeout(t.Context(), esperaDelFinal)
	defer cancelar()

	require.NoError(t, s.sesion.Cerrar(ctx),
		"el cierre de la sesión, que con el cliente anterior comprueba cada línea de la salida estándar")

	return s.codigoFinal(t)
}

// codigoFinal espera a que el servidor termine y devuelve su código.
func (s *servidorEnProceso) codigoFinal(t *testing.T) int {
	t.Helper()

	select {
	case <-s.terminado:
		return s.codigo
	case <-time.After(esperaDelFinal):
		require.FailNow(t, "el servidor no termina", "salida de error: %s", s.errores.lineas())

		return -1
	}
}

// invocarCon ejecuta una orden con el registro contra dos buffers.
func invocarCon(t *testing.T, registro *app.Registro, argv ...string) invocacion {
	t.Helper()

	var salida, errores bytes.Buffer

	codigo := app.Main(slices.Concat([]string{nombreDelBinario}, argv), registro, &salida, &errores,
		versionDelEjemplo, commitDelEjemplo, fechaDelEjemplo)

	return invocacion{codigo: codigo, salida: salida.String(), errores: errores.String()}
}

// sobreDe lee un sobre: las seis claves, con `data` sin interpretar.
func sobreDe(t *testing.T, documento []byte) map[string]json.RawMessage {
	t.Helper()

	var sobre map[string]json.RawMessage

	require.NoError(t, json.Unmarshal(documento, &sobre), "el sobre es un objeto JSON: %s", documento)
	assert.ElementsMatch(t, clavesDelSobre, slices.Collect(maps.Keys(sobre)), "las seis claves del sobre")

	return sobre
}

// sinFechaDeConsulta es el sobre sin `fecha_consulta`, que es lo único en lo
// que el de una llamada puede no coincidir con el de su orden: cuando la pone
// el reloj, es la de cada invocación (FR-010 de H21).
func sinFechaDeConsulta(t *testing.T, documento []byte) map[string]json.RawMessage {
	t.Helper()

	sobre := sobreDe(t, documento)
	delete(sobre, "fecha_consulta")

	return sobre
}

// falloDe son la clase y el mensaje de un sobre de fallo.
func falloDe(t *testing.T, documento []byte) (clase, mensaje string) {
	t.Helper()

	sobre := sobreDe(t, documento)
	require.JSONEq(t, "false", string(sobre["ok"]), "el sobre es de fallo: %s", documento)

	var datos struct {
		Clase   string `json:"clase"`
		Mensaje string `json:"mensaje"`
	}

	require.NoError(t, json.Unmarshal(sobre["data"], &datos))

	return datos.Clase, datos.Mensaje
}

// boeDeLaReproduccion es lo que compone un registro de consulta y lo que un
// test ve de él desde fuera: la carpeta de las grabaciones que sirve la
// reproducción; la hora que declara cada respuesta, que es también la de la
// caché y la de graph; la carpeta de la caché; la del world.db que lee graph;
// y cuántas veces construye su cliente el applet boe, que es como se sabe si
// una llamada llegó a pedir algo.
type boeDeLaReproduccion struct {
	grabaciones string
	hora        func() time.Time
	cache       string
	grafo       string

	construidos atomic.Int64
}

// dependencias son las de boe sobre las grabaciones, que httpx.Replay sirve sin
// abrir ninguna conexión, con la hora fija y la caché en una carpeta del test
// con esa misma hora.
func (b *boeDeLaReproduccion) dependencias() app.DependenciasDeBoe {
	return app.DependenciasDeBoe{
		Cliente: func(registrador *slog.Logger) (*httpx.Cliente, error) {
			b.construidos.Add(1)

			return httpx.Replay(b.grabaciones,
				httpx.ConFuente(boe.NombreDeLaFuente),
				httpx.ConRegistrador(registrador),
				httpx.ConHora(b.hora))
		},
		Cache: []cache.Opcion{cache.ConDirectorio(b.cache), cache.ConReloj(b.hora)},
	}
}

// registro es un registro local con los applets de consulta de producción
// —boe, sobre la reproducción; graph y territorio—, sin almacén al que
// entregar: quien quiera uno lo registra.
func (b *boeDeLaReproduccion) registro(t *testing.T) *app.Registro {
	t.Helper()

	fuentes, err := app.FuentesEmbebidas()
	require.NoError(t, err)

	return registroCon(t,
		app.AppletBoe(b.dependencias()),
		app.AppletGrafo(app.DependenciasDeGrafo{
			Reloj:   b.hora,
			Almacen: []graph.Opcion{graph.ConDirectorio(b.grafo)},
		}),
		app.AppletTerritorio(fuentes),
	)
}

// registroDeConsulta es el registro de consulta de casi todos estos tests:
// sobre las grabaciones de H4, con la hora de la reproducción y en carpetas
// del test.
func registroDeConsulta(t *testing.T) (*app.Registro, *boeDeLaReproduccion) {
	t.Helper()

	deBoe := &boeDeLaReproduccion{
		grabaciones: grabacionesDeLaFuente,
		hora:        horaDeLaReproduccion,
		cache:       t.TempDir(),
		grafo:       t.TempDir(),
	}

	return deBoe.registro(t), deBoe
}

// almacenAnotado es el doble del grafo del mundo de estos tests: anota cada
// lote que recibe y devuelve el fallo que se le fije. Lo llaman las gorrutinas
// del servidor, una por llamada.
type almacenAnotado struct {
	cerrojo sync.Mutex
	lotes   []core.Lote
	fallo   error
}

var _ core.GraphStore = (*almacenAnotado)(nil)

func (a *almacenAnotado) Apply(_ context.Context, lote core.Lote) error {
	a.cerrojo.Lock()
	defer a.cerrojo.Unlock()

	a.lotes = append(a.lotes, lote)

	return a.fallo
}

func (a *almacenAnotado) recibidos() []core.Lote {
	a.cerrojo.Lock()
	defer a.cerrojo.Unlock()

	return slices.Clone(a.lotes)
}

// appletDeEspera es el applet de estos tests que ningún binario tiene: un
// verbo que espera a que termine su contexto —o a que el test lo suelte— y uno
// que vuelve en el acto, los dos sin red ni disco.
type appletDeEspera struct {
	// empezada recibe una señal cuando el verbo que espera empieza a esperar.
	empezada chan struct{}
	// suelta, cerrado, hace volver al verbo que espera sin que su contexto haya
	// terminado.
	suelta chan struct{}
	soltar func()
}

func nuevoAppletDeEspera() *appletDeEspera {
	applet := &appletDeEspera{empezada: make(chan struct{}, 1), suelta: make(chan struct{})}
	applet.soltar = sync.OnceFunc(func() { close(applet.suelta) })

	return applet
}

func (*appletDeEspera) Nombre() string { return "espera" }

func (*appletDeEspera) Descripcion() string { return "Espera o vuelve, para los tests del servidor." }

func (a *appletDeEspera) Verbos() []app.Verbo {
	return []app.Verbo{
		{
			Nombre:      "esperar",
			Descripcion: "Espera a que termine su contexto.",
			Argumentos:  func() app.Argumentos { return &argumentosDeEsperar{applet: a} },
			Salida:      map[string]any{},
		},
		{
			Nombre:      "volver",
			Descripcion: "Vuelve en el acto.",
			Argumentos:  func() app.Argumentos { return &argumentosDeVolver{} },
			Salida:      map[string]any{},
		},
	}
}

// argumentosDeEsperar no declara ningún argumento: el applet va en un campo sin
// exportar, que la gramática no ve.
type argumentosDeEsperar struct {
	applet *appletDeEspera
}

func (a *argumentosDeEsperar) Ejecutar(
	ctx context.Context, _ schema.Contexto, _ *slog.Logger,
) (schema.Resultado, error) {
	select {
	case a.applet.empezada <- struct{}{}:
	default:
	}

	select {
	case <-ctx.Done():
	case <-a.applet.suelta:
	}

	return resultadoDeEspera("esperar"), nil
}

type argumentosDeVolver struct{}

func (*argumentosDeVolver) Ejecutar(context.Context, schema.Contexto, *slog.Logger) (schema.Resultado, error) {
	return resultadoDeEspera("volver"), nil
}

func resultadoDeEspera(verbo string) schema.Resultado {
	return schema.Resultado{
		Procedencia: schema.Procedencia{Fuente: "kitlegal.espera", URL: "kitlegal:applet/espera"},
		Datos:       map[string]any{"verbo": verbo},
	}
}

// llamadaSimultanea es una de las entradas de TestLlamadasSimultaneas: la
// herramienta, sus argumentos y la orden que da el sobre con el que se compara.
type llamadaSimultanea struct {
	herramienta string
	argumentos  string
	orden       []string
}

func llamadasSimultaneas() []llamadaSimultanea {
	const norma = normaDeLasLlamadas

	deUnBloque := func(bloque string) llamadaSimultanea {
		return llamadaSimultanea{
			herramienta: "boe_articulo",
			argumentos:  `{"norma":"` + norma + `","bloque":"` + bloque + `"}`,
			orden:       []string{"boe", "articulo", norma, bloque},
		}
	}

	deLaNorma := func(verbo string) llamadaSimultanea {
		return llamadaSimultanea{
			herramienta: "boe_" + verbo,
			argumentos:  `{"norma":"` + norma + `"}`,
			orden:       []string{"boe", verbo, norma},
		}
	}

	deUnMunicipio := func(municipio string) llamadaSimultanea {
		return llamadaSimultanea{
			herramienta: "territorio_resolver",
			argumentos:  `{"consulta":"` + municipio + `"}`,
			orden:       []string{"territorio", "resolver", municipio},
		}
	}

	return []llamadaSimultanea{
		// La misma herramienta, con entradas distintas: tres bloques que
		// existen y uno que no, que falla con su propio sobre.
		deUnBloque("a1"), deUnBloque("a21"), deUnBloque("a22"), deUnBloque("a23"), deUnBloque("a9999"),
		// Y otras herramientas, de boe y de otro applet.
		deLaNorma("indice"), deLaNorma("metadatos"), deLaNorma("analisis"),
		deUnMunicipio("Leganés"), deUnMunicipio("Tordesillas"),
	}
}

// TestLlamadasSimultaneas comprueba FR-074 de H21: N llamadas a la vez, a la
// misma herramienta y a otras, sobre la reproducción, reciben cada una el sobre
// de su entrada, sin mezclar el de una con el de otra, y el detector de
// carreras no encuentra ninguna (FR-014, SC-007). El sobre de cada entrada es
// el de su orden sobre otro registro igual, con su propia caché: las llamadas
// empiezan con la suya vacía, de modo que piden a la reproducción, escriben en
// la caché y entregan al grafo del mundo a la vez.
func TestLlamadasSimultaneas(t *testing.T) {
	t.Parallel()

	const repeticiones = 4

	entradas := llamadasSimultaneas()

	deReferencia, _ := registroDeConsulta(t)
	esperados := make([]map[string]json.RawMessage, len(entradas))

	for i, entrada := range entradas {
		orden := invocarCon(t, deReferencia, slices.Concat(entrada.orden, []string{"--json"})...)
		esperados[i] = sinFechaDeConsulta(t, []byte(orden.salida))
	}

	registro, _ := registroDeConsulta(t)
	registro.EntregarAlGrafo(graph.Nuevo(graph.ConDirectorio(t.TempDir())))

	servidor := arrancarElServidor(t, registro, false)

	type recibido struct {
		resultado mcptest.Resultado
		err       error
	}

	recibidos := make([]recibido, len(entradas)*repeticiones)
	salida := make(chan struct{})

	var grupo sync.WaitGroup

	for i := range recibidos {
		entrada := entradas[i%len(entradas)]

		grupo.Go(func() {
			// Todas esperan la misma señal, para salir a la vez.
			<-salida

			resultado, err := servidor.pedir(t, entrada.herramienta, json.RawMessage(entrada.argumentos))
			recibidos[i] = recibido{resultado: resultado, err: err}
		})
	}

	close(salida)
	grupo.Wait()

	for i, llamada := range recibidos {
		entrada := entradas[i%len(entradas)]

		if assert.NoError(t, llamada.err, "la llamada %d, a %s con %s", i, entrada.herramienta, entrada.argumentos) {
			assert.Equal(t, esperados[i%len(entradas)], sinFechaDeConsulta(t, llamada.resultado.Sobre),
				"la llamada %d, a %s con %s, recibe el sobre de su entrada", i, entrada.herramienta, entrada.argumentos)
		}
	}

	assert.Equal(t, 0, servidor.cerrar(t), servidor.errores.lineas())
	assert.Empty(t, servidor.errores.lineasCon(lineaDeEntregaFallida),
		"cada llamada que termina bien entrega al grafo lo que observó, también a la vez que otras")
}

// TestServirSinEntrada comprueba que la entrada que se cierra con una llamada
// en curso es el final normal del servidor: la llamada termina y el servidor
// sale con 0, sin sobre propio en la salida estándar (FR-024 de H21;
// contracts/servidor-mcp.md §3, «Otros casos»; research.md V11).
func TestServirSinEntrada(t *testing.T) {
	t.Parallel()

	espera := nuevoAppletDeEspera()
	servidor := arrancarElServidor(t, registroCon(t, espera), true)
	t.Cleanup(espera.soltar)

	llamada := make(chan error, 1)

	go func() {
		_, err := servidor.pedir(t, "espera_esperar", nil)
		llamada <- err
	}()

	select {
	case <-espera.empezada:
	case <-time.After(esperaDelFinal):
		require.FailNow(t, "la llamada no llega al verbo", "salida de error: %s", servidor.errores.lineas())
	}

	require.NoError(t, servidor.entrada.Close(), "el cliente cierra la entrada con la llamada en curso")

	// El servidor no termina mientras la llamada sigue en curso. La espera solo
	// puede dar un falso verde, nunca un falso rojo: un servidor que espera a su
	// llamada no termina por mucho que tarde el test.
	select {
	case <-servidor.terminado:
		require.FailNow(t, "el servidor termina con una llamada en curso", "código %d", servidor.codigo)
	case <-time.After(100 * time.Millisecond):
	}

	espera.soltar()

	assert.Equal(t, 0, servidor.codigoFinal(t),
		"la entrada cerrada es el final normal, también con una llamada en curso: %s", servidor.errores.lineas())

	// La llamada deja de esperar: con su respuesta, si el servidor llegó a
	// escribirla antes de ver el fin de su entrada, o sin ella.
	select {
	case <-llamada:
	case <-time.After(esperaDelFinal):
		require.FailNow(t, "el cliente sigue esperando la respuesta de un servidor que ha terminado")
	}

	assert.Equal(t, 0, servidor.cerrar(t), "y lo que escribió en su salida estándar son mensajes del protocolo")
}

// invocacionSinServir es una invocación de `mcp serve` que no debe atender
// ningún mensaje, hecha con la entrada abierta: lo que escribe, su código y
// cuántas veces leyó de la entrada.
type invocacionSinServir struct {
	invocacion

	lecturas int64
}

// lectorContado cuenta las lecturas que se le piden.
type lectorContado struct {
	lector   io.Reader
	lecturas atomic.Int64
}

func (l *lectorContado) Read(p []byte) (int, error) {
	l.lecturas.Add(1)

	return l.lector.Read(p)
}

// invocarSinServir ejecuta `mcp serve` con los argumentos y con una entrada que
// sigue abierta —una tubería en la que nadie escribe y que nadie cierra hasta
// que el test termina—: una invocación que sirviera se quedaría leyéndola, y
// el test lo ve porque no vuelve.
func invocarSinServir(t *testing.T, dependencias app.DependenciasDeMCP, otros []app.Applet, argumentos ...string,
) invocacionSinServir {
	t.Helper()

	return invocarSinCliente(t, dependencias, otros, slices.Concat([]string{"mcp", "serve"}, argumentos)...)
}

// invocarSinCliente ejecuta una orden sobre un registro con el applet mcp de
// esas dependencias y los otros applets, sin ningún cliente conectado, y
// espera a que vuelva: lo que escribe, su código y cuántas veces leyó de la
// entrada del applet.
func invocarSinCliente(t *testing.T, dependencias app.DependenciasDeMCP, otros []app.Applet, orden ...string,
) invocacionSinServir {
	t.Helper()

	var entrada *lectorContado

	if dependencias.Entrada != nil {
		entrada = &lectorContado{lector: dependencias.Entrada}
		dependencias.Entrada = entrada
	}

	registro := registroCon(t, slices.Concat([]app.Applet{app.AppletMCP(dependencias)}, otros)...)

	vuelta := make(chan invocacion, 1)

	go func() {
		vuelta <- invocarCon(t, registro, orden...)
	}()

	select {
	case res := <-vuelta:
		sin := invocacionSinServir{invocacion: res}
		if entrada != nil {
			sin.lecturas = entrada.lecturas.Load()
		}

		return sin
	case <-time.After(esperaDelFinal):
		require.FailNow(t, "la orden no vuelve con la entrada abierta", "orden: %q", orden)

		return invocacionSinServir{}
	}
}

// entradaAbierta es una entrada estándar en la que nadie escribe y que no se
// cierra hasta que el test termina.
func entradaAbierta(t *testing.T) io.Reader {
	t.Helper()

	lee, escribe := io.Pipe()

	t.Cleanup(func() { assert.NoError(t, escribe.Close()) })

	return lee
}

// TestServirEnEnsayo comprueba que `mcp serve --dry-run`, con la entrada
// abierta, vuelve sin leerla, con la salida estándar vacía y con la descripción
// del kernel en la salida de error, carácter a carácter (FR-022 de H21;
// contracts/servidor-mcp.md §1). No es paralelo porque fija KITLEGAL_LOG: lo
// que afirma de la salida de error depende del nivel del registro de eventos.
func TestServirEnEnsayo(t *testing.T) {
	t.Setenv(cli.VariableNivel, "")

	res := invocarSinServir(t, app.DependenciasDeMCP{Entrada: entradaAbierta(t)},
		[]app.Applet{nuevoAppletDeEspera()}, "--dry-run")

	assert.Equal(t, 0, res.codigo, res.errores)
	assert.Empty(t, res.salida, "no atiende ningún mensaje ni escribe ningún sobre")
	assert.Equal(t, `--dry-run: no se ha ejecutado nada; se habría ejecutado el applet "mcp", el verbo "serve",`+
		` con los argumentos ["serve" "--dry-run"]`+"\n", res.errores)
	assert.Zero(t, res.lecturas, "no lee de la entrada")
}

// TestServirConAsunto comprueba que `mcp serve --asunto <valor>` es un error de
// argumentos que no atiende ningún mensaje: el código 2, el mensaje del
// contrato carácter a carácter y, con --json, el sobre de fallo del kernel
// (FR-021 de H21; contracts/servidor-mcp.md §1; research.md D9). No es paralelo
// porque fija KITLEGAL_LOG, como TestServirEnEnsayo.
func TestServirConAsunto(t *testing.T) {
	t.Setenv(cli.VariableNivel, "")

	const mensaje = "argumentos inválidos: mcp serve no admite --asunto: el servidor no expone nada del asunto"

	t.Run("sin --json: el evento de la invocación y el mensaje, con la salida estándar vacía", func(t *testing.T) {
		t.Setenv(cli.VariableNivel, "")

		res := invocarSinServir(t, app.DependenciasDeMCP{Entrada: entradaAbierta(t)}, nil, "--asunto", "x")

		assert.Equal(t, 2, res.codigo, res.errores)
		assert.Empty(t, res.salida)
		assert.Zero(t, res.lecturas, "no lee de la entrada")

		lineas := strings.Split(strings.TrimSuffix(res.errores, "\n"), "\n")
		require.Len(t, lineas, 2, res.errores)
		assert.Contains(t, lineas[0], "level=WARN msg=invocación applet=mcp verbo=serve ")
		assert.Contains(t, lineas[0], " clase="+claseArgumentos)
		assert.Equal(t, mensaje, lineas[1])
	})

	t.Run("con --json: además, el sobre de fallo del kernel", func(t *testing.T) {
		t.Setenv(cli.VariableNivel, "")

		res := invocarSinServir(t, app.DependenciasDeMCP{Entrada: entradaAbierta(t)}, nil, "--asunto", "x", "--json")

		assert.Equal(t, 2, res.codigo, res.errores)
		assert.Zero(t, res.lecturas, "no lee de la entrada")
		assert.True(t, strings.HasSuffix(res.errores, mensaje+"\n"), res.errores)

		sobre := sobreDe(t, []byte(res.salida))
		assert.JSONEq(t, `"kitlegal.cli"`, string(sobre["fuente"]))
		assert.JSONEq(t, `"kitlegal:cli"`, string(sobre["url"]))

		clase, mensajeDelSobre := falloDe(t, []byte(res.salida))
		assert.Equal(t, claseArgumentos, clase)
		assert.Equal(t, mensaje, mensajeDelSobre)
	})

	t.Run("con --dry-run: el mismo error, con la descripción del kernel delante", func(t *testing.T) {
		t.Setenv(cli.VariableNivel, "")

		res := invocarSinServir(t, app.DependenciasDeMCP{Entrada: entradaAbierta(t)}, nil, "--asunto", "x", "--dry-run")

		assert.Equal(t, 2, res.codigo, res.errores)
		assert.Empty(t, res.salida)
		assert.Zero(t, res.lecturas, "no lee de la entrada")
		assert.Contains(t, res.errores, "--dry-run: no se ha ejecutado nada; ")
		assert.True(t, strings.HasSuffix(res.errores, mensaje+"\n"), res.errores)
	})
}

// appletSinFabrica es un applet mal escrito: su verbo no declara fábrica de
// argumentos, así que no se puede describir ni llamar.
type appletSinFabrica struct{}

func (appletSinFabrica) Nombre() string { return "roto" }

func (appletSinFabrica) Descripcion() string { return "Un applet con un verbo sin fábrica." }

func (appletSinFabrica) Verbos() []app.Verbo {
	return []app.Verbo{{Nombre: "nada", Descripcion: "No declara argumentos."}}
}

// appletIndescriptible es otro applet mal escrito: el `data` de su verbo lleva
// dos tipos distintos que --describe nombraría igual, así que ni su documento
// ni los esquemas de su herramienta se pueden construir.
type appletIndescriptible struct{}

func (appletIndescriptible) Nombre() string { return "doble" }

func (appletIndescriptible) Descripcion() string {
	return "Un applet con un verbo que no se puede describir."
}

func (appletIndescriptible) Verbos() []app.Verbo {
	return []app.Verbo{{
		Nombre:      "nada",
		Descripcion: "Devuelve dos tipos que se llaman igual.",
		Argumentos:  func() app.Argumentos { return &argumentosDeVolver{} },
		Salida:      salidaConHomonimos(),
	}}
}

// Los dos homónimos son dos tipos distintos con el mismo nombre y el mismo
// paquete, que es lo que el lenguaje permite a dos tipos declarados dentro de
// dos funciones.
func homonimoUno() reflect.Type {
	type Homonimo struct {
		Uno string `json:"uno"`
	}

	return reflect.TypeFor[Homonimo]()
}

func homonimoDos() reflect.Type {
	type Homonimo struct {
		Dos int `json:"dos"`
	}

	return reflect.TypeFor[Homonimo]()
}

// salidaConHomonimos es un `data` con un campo de cada homónimo.
func salidaConHomonimos() any {
	tipo := reflect.StructOf([]reflect.StructField{
		{Name: "Uno", Type: homonimoUno(), Tag: `json:"uno"`},
		{Name: "Dos", Type: homonimoDos(), Tag: `json:"dos"`},
	})

	return reflect.New(tipo).Elem().Interface()
}

// TestServirConUnDefectoDeComposicion comprueba que un servidor que no se puede
// componer no atiende ningún mensaje y termina como cualquier defecto de quien
// escribió el applet o la raíz de composición: clase `inesperado`, código 1 y
// nunca un pánico (contracts/servidor-mcp.md §1 de H21; FR-031 y FR-033 de H1).
func TestServirConUnDefectoDeComposicion(t *testing.T) {
	t.Parallel()

	t.Run("sin entrada de la que leer", func(t *testing.T) {
		t.Parallel()

		res := invocarSinServir(t, app.DependenciasDeMCP{}, nil)

		assert.Equal(t, 1, res.codigo, res.errores)
		assert.Empty(t, res.salida)
		assert.Contains(t, res.errores, "mcp: el applet se compuso sin entrada")
	})

	t.Run("con un verbo sin fábrica de argumentos en el registro", func(t *testing.T) {
		t.Parallel()

		res := invocarSinServir(t, app.DependenciasDeMCP{Entrada: entradaAbierta(t)}, []app.Applet{appletSinFabrica{}})

		assert.Equal(t, 1, res.codigo, res.errores)
		assert.Empty(t, res.salida)
		assert.Zero(t, res.lecturas, "no lee de la entrada")
		assert.Contains(t, res.errores, `el verbo "nada" de "roto" no declara fábrica de argumentos`)
	})

	t.Run("con un verbo cuyos esquemas no se pueden construir", func(t *testing.T) {
		t.Parallel()

		res := invocarSinServir(t, app.DependenciasDeMCP{Entrada: entradaAbierta(t)}, []app.Applet{appletIndescriptible{}})

		assert.Equal(t, 1, res.codigo, res.errores)
		assert.Empty(t, res.salida)
		assert.Zero(t, res.lecturas, "no lee de la entrada")
		assert.Contains(t, res.errores, "la herramienta doble_nada")
		assert.Contains(t, res.errores, "Homonimo", "el fallo nombra el nombre en conflicto")
	})
}

// TestOrdenesDelAppletMCP comprueba lo que el applet mcp responde sin servir,
// como cualquier applet (contracts/servidor-mcp.md §1 de H21; FR-001, FR-022):
// la ayuda del binario lo lista con su descripción; sin verbo no hay ninguno
// por omisión y el fallo enumera `serve`; la ayuda del applet y la del verbo
// salen con 0; y --describe da el documento del verbo, que no declara ningún
// argumento propio ni la forma de su data. Ninguna lee de la entrada.
//
// Lo que responde es lo del registro de producción tal cual, que ya trae el
// applet. Que no lee de la entrada se comprueba antes, con la misma orden
// sobre un registro local cuyo mcp lee de una tubería abierta y contada: la
// de producción es la entrada estándar del proceso, que el test no ve.
func TestOrdenesDelAppletMCP(t *testing.T) {
	t.Parallel()

	const (
		delApplet = "Sirve las herramientas de kitlegal a un agente por el protocolo MCP."
		delVerbo  = "Atiende el protocolo MCP por la entrada y la salida estándar hasta que la entrada se cierra."
	)

	distribuido, err := app.RegistroDeProduccion("")
	require.NoError(t, err)

	invocar := func(t *testing.T, orden ...string) invocacion {
		t.Helper()

		// Una orden que sirviera no volvería de aquí, y la de producción, que
		// leería de la entrada estándar del proceso, no llega a ejecutarse.
		local := invocarSinCliente(t, app.DependenciasDeMCP{Entrada: entradaAbierta(t)}, nil, orden...)
		require.Zero(t, local.lecturas, "%q no lee de la entrada", orden)

		return invocarCon(t, distribuido, orden...)
	}

	t.Run("la ayuda del binario lista el applet", func(t *testing.T) {
		t.Parallel()

		res := invocar(t, "--help")

		assert.Equal(t, 0, res.codigo, res.errores)
		assert.Regexp(t, `(?m)^\s+mcp\s+`+regexp.QuoteMeta(delApplet)+`$`, res.salida)
	})

	t.Run("sin verbo: no hay ninguno por omisión", func(t *testing.T) {
		t.Parallel()

		res := invocar(t, "mcp")

		assert.Equal(t, 2, res.codigo, res.errores)
		assert.Empty(t, res.salida)
		assert.Contains(t, res.errores, "verbos de mcp: serve")
	})

	t.Run("la ayuda del applet enumera su único verbo", func(t *testing.T) {
		t.Parallel()

		res := invocar(t, "mcp", "--help")

		assert.Equal(t, 0, res.codigo, res.errores)
		assert.Contains(t, res.salida, "uso: mcp <verbo> [banderas]\n\n"+delApplet+"\n")
		assert.Regexp(t, `(?m)^\s+serve\s+`+regexp.QuoteMeta(delVerbo)+`$`, res.salida)
	})

	t.Run("la ayuda del verbo", func(t *testing.T) {
		t.Parallel()

		res := invocar(t, "mcp", "serve", "--help")

		assert.Equal(t, 0, res.codigo, res.errores)
		// La descripción sale partida en líneas por el analizador: se afirma
		// su comienzo.
		assert.Contains(t, res.salida, "Usage: mcp serve [flags]\n\nAtiende el protocolo MCP por la entrada y la salida")
	})

	t.Run("--describe da el documento del verbo sin servir", func(t *testing.T) {
		t.Parallel()

		res := invocar(t, "mcp", "serve", "--describe")
		require.Equal(t, 0, res.codigo, res.errores)

		var descrito struct {
			Titulo      string `json:"title"`
			Descripcion string `json:"description"`
			Partes      struct {
				Entrada struct {
					Propiedades  map[string]json.RawMessage `json:"properties"`
					Obligatorios []string                   `json:"required"`
				} `json:"entrada"`
				Salida struct {
					ConExito struct {
						Propiedades map[string]json.RawMessage `json:"properties"`
					} `json:"then"`
				} `json:"salida"`
			} `json:"properties"`
		}

		require.NoError(t, json.Unmarshal([]byte(res.salida), &descrito))

		assert.Equal(t, "mcp serve", descrito.Titulo)
		assert.Equal(t, delVerbo, descrito.Descripcion)

		globales := make([]string, 0, len(banderasGlobales))
		for _, global := range banderasGlobales {
			globales = append(globales, strings.TrimPrefix(global.nombre, "--"))
		}

		assert.ElementsMatch(t, globales, slices.Collect(maps.Keys(descrito.Partes.Entrada.Propiedades)),
			"serve no tiene argumentos propios: su entrada son las ocho banderas globales")
		assert.Empty(t, descrito.Partes.Entrada.Obligatorios)
		assert.JSONEq(t, "true", string(descrito.Partes.Salida.ConExito.Propiedades["data"]),
			"serve no declara su data: al servir no emite ningún sobre")
	})
}

// TestDependenciasDeMCPDelSistema comprueba las dependencias del binario
// distribuido: la entrada estándar del proceso y la versión que recibe
// (data-model §5 de H21).
func TestDependenciasDeMCPDelSistema(t *testing.T) {
	t.Parallel()

	dependencias := app.DependenciasDeMCPDelSistema(versionDelEjemplo)

	assert.Same(t, os.Stdin, dependencias.Entrada)
	assert.Equal(t, versionDelEjemplo, dependencias.Version)
}

// TestServirConUnaEntradaQueNoEsDelProtocolo comprueba que un final que no es
// el cierre de la entrada —aquí, una línea que no es un mensaje del protocolo—
// termina el servidor como cualquier fallo del kernel: clase `inesperado`,
// código 1 y su mensaje en la salida de error (contracts/servidor-mcp.md §1 de
// H21; research.md V13).
func TestServirConUnaEntradaQueNoEsDelProtocolo(t *testing.T) {
	t.Parallel()

	lee, escribe := io.Pipe()

	t.Cleanup(func() { assert.NoError(t, escribe.Close()) })

	escrita := make(chan error, 1)

	go func() {
		_, err := escribe.Write([]byte("esto no es un mensaje del protocolo\n"))
		escrita <- err
	}()

	res := invocarSinServir(t, app.DependenciasDeMCP{Entrada: lee}, nil, "--json")

	require.NoError(t, <-escrita, "el servidor lee la línea")

	assert.Equal(t, 1, res.codigo, res.errores)
	assert.Contains(t, res.errores, "mcp: el servidor ha terminado sin que se cierre su entrada: ")

	clase, mensaje := falloDe(t, []byte(res.salida))
	assert.Equal(t, "inesperado", clase)
	assert.Contains(t, mensaje, "mcp: el servidor ha terminado sin que se cierre su entrada: ")
}

// TestLlamadaComoLaOrden comprueba, llamada a llamada y en proceso, lo que
// contracts/servidor-mcp.md §3 de H21 dice que hace cada una: su sobre son los
// bytes que la orden del verbo escribe con --json, sin el salto final; unos
// argumentos que no valen dan el sobre de la clase `argumentos` sin ejecutar
// nada; lo observado llega al grafo antes de que la llamada devuelva, y solo si
// termina bien; y las banderas de `mcp serve` valen para todas (FR-010,
// FR-011, FR-013, FR-015, FR-020).
func TestLlamadaComoLaOrden(t *testing.T) {
	t.Parallel()

	t.Run("el sobre es el de la orden con --json, sin el salto final", testSobreDeLaOrden)
	t.Run("unos argumentos que no valen: el sobre de la clase argumentos, sin ejecutar nada", testArgumentosQueNoValen)
	t.Run("lo observado llega al grafo antes de devolver, y solo si la llamada termina bien", testEntregaDeLaLlamada)
	t.Run("una entrega que falla no cambia el sobre y deja su línea en la salida de error", testEntregaQueFalla)
	t.Run("los hallazgos de graph_check son un resultado y no un error", testHallazgosQueNoSonUnError)
	t.Run("--no-graph vale para cada llamada: no se entrega nada", testSinGrafoEnCadaLlamada)
	t.Run("--offline vale para cada llamada: boe no pide nada", testSinRedEnCadaLlamada)
}

func testSobreDeLaOrden(t *testing.T) {
	t.Parallel()

	const norma = normaDeLasLlamadas

	casos := []struct {
		herramienta string
		argumentos  string
		orden       []string
		codigo      int
		// fechaDelReloj dice que la fecha de consulta del sobre la pone el
		// reloj del kernel, de modo que es la de cada invocación.
		fechaDelReloj bool
	}{
		{
			herramienta: "territorio_resolver", argumentos: argumentosDeLeganes,
			orden: []string{"territorio", "resolver", "Leganés"},
		},
		{
			herramienta: "boe_articulo", argumentos: argumentosDelBloqueA21,
			orden: []string{"boe", "articulo", norma, "a21"},
		},
		{
			herramienta: "boe_articulos", argumentos: `{"norma":"` + norma + `","bloques":["a21","a22","a23"]}`,
			orden: []string{"boe", "articulos", norma, "a21", "a22", "a23"},
		},
		{
			herramienta: "boe_buscar", argumentos: `{"texto":["procedimiento","administrativo","común"]}`,
			orden: []string{"boe", "buscar", "procedimiento", "administrativo", "común"},
		},
		{
			herramienta: "graph_stats", argumentos: `{}`,
			orden: []string{"graph", "stats"},
		},
		{
			herramienta: "boe_articulo", argumentos: argumentosDelBloqueA9999,
			orden:  []string{"boe", "articulo", norma, "a9999"},
			codigo: 3, fechaDelReloj: true,
		},
		{
			// Falta un argumento obligatorio: lo dice el analizador de la
			// orden, con el mensaje de la orden.
			herramienta: "boe_articulo", argumentos: `{"norma":"` + norma + `"}`,
			orden:  []string{"boe", "articulo", norma},
			codigo: 2, fechaDelReloj: true,
		},
	}

	deReferencia, _ := registroDeConsulta(t)
	registro, _ := registroDeConsulta(t)
	servidor := arrancarElServidor(t, registro, true)

	for _, caso := range casos {
		orden := invocarCon(t, deReferencia, slices.Concat(caso.orden, []string{"--json"})...)
		require.Equal(t, caso.codigo, orden.codigo, "la orden %q: %s", caso.orden, orden.errores)

		resultado := servidor.llamar(t, caso.herramienta, caso.argumentos)

		assert.Equal(t, caso.codigo != 0, resultado.Fallo,
			"%s con %s: error de herramienta si y solo si el código de la orden no es 0", caso.herramienta, caso.argumentos)

		if caso.fechaDelReloj {
			assert.Equal(t, sinFechaDeConsulta(t, []byte(orden.salida)), sinFechaDeConsulta(t, resultado.Sobre),
				"%s con %s: el sobre de la orden, salvo la fecha de consulta", caso.herramienta, caso.argumentos)

			continue
		}

		assert.Equal(t, orden.salida, string(resultado.Sobre)+"\n",
			"%s con %s: los bytes de la orden, sin el salto final", caso.herramienta, caso.argumentos)
	}

	assert.Equal(t, 0, servidor.cerrar(t), "las llamadas que fallan no terminan el servidor (FR-015)")
}

func testArgumentosQueNoValen(t *testing.T) {
	t.Parallel()

	casos := []struct {
		herramienta string
		argumentos  string
		mensaje     string
	}{
		{
			herramienta: "boe_articulo",
			argumentos:  `{"norma":"BOE-A-2015-10565","bloque":"a21","apartado":"1"}`,
			mensaje:     `argumentos inválidos: boe_articulo no tiene el argumento "apartado"`,
		},
		{
			// El nombre de una bandera global es una propiedad de más: las
			// banderas son del servidor (FR-020).
			herramienta: "boe_articulo",
			argumentos:  `{"norma":"BOE-A-2015-10565","bloque":"a21","offline":true}`,
			mensaje:     `argumentos inválidos: boe_articulo no tiene el argumento "offline"`,
		},
		{
			herramienta: "boe_articulo",
			argumentos:  `{"norma":"BOE-A-2015-10565","bloque":21}`,
			mensaje:     `argumentos inválidos: el argumento "bloque" de boe_articulo tiene que ser una cadena`,
		},
		{
			herramienta: "graph_check",
			argumentos:  `{"bloques":["a21"]}`,
			mensaje:     `argumentos inválidos: el argumento "bloques" de graph_check no se puede dar sin "norma"`,
		},
	}

	registro, deBoe := registroDeConsulta(t)
	almacen := &almacenAnotado{}
	registro.EntregarAlGrafo(almacen)

	servidor := arrancarElServidor(t, registro, false)

	for _, caso := range casos {
		resultado := servidor.llamar(t, caso.herramienta, caso.argumentos)
		assert.True(t, resultado.Fallo, "%s con %s es un error de herramienta", caso.herramienta, caso.argumentos)

		sobre := sobreDe(t, resultado.Sobre)
		assert.JSONEq(t, `"kitlegal.cli"`, string(sobre["fuente"]), "firma el kernel: no se llegó a consultar nada")
		assert.JSONEq(t, `"kitlegal:cli"`, string(sobre["url"]))

		clase, mensaje := falloDe(t, resultado.Sobre)
		assert.Equal(t, claseArgumentos, clase)
		assert.Equal(t, caso.mensaje, mensaje)
	}

	assert.Equal(t, 0, servidor.cerrar(t))

	assert.Zero(t, deBoe.construidos.Load(), "ninguna llamada construye un cliente de la fuente")
	assert.Empty(t, almacen.recibidos(), "ni entrega nada al grafo")

	entradas, err := os.ReadDir(deBoe.cache)
	require.NoError(t, err)
	assert.Empty(t, entradas, "ni abre la caché")
}

func testEntregaDeLaLlamada(t *testing.T) {
	t.Parallel()

	registro, _ := registroDeConsulta(t)
	almacen := &almacenAnotado{}
	registro.EntregarAlGrafo(almacen)

	servidor := arrancarElServidor(t, registro, false)

	leido := servidor.llamar(t, "boe_articulo", argumentosDelBloqueA21)
	require.False(t, leido.Fallo, "%s", leido.Sobre)

	// Sin esperar a nada: la llamada ya ha devuelto, así que lo que observó
	// está entregado, y la llamada que el cliente envíe después lo ve (FR-013).
	lotes := almacen.recibidos()
	require.Len(t, lotes, 1, "la lectura entrega un lote antes de devolver su resultado")

	sobre := sobreDe(t, leido.Sobre)
	assert.JSONEq(t, string(sobre["fuente"]), `"`+lotes[0].Fuente+`"`, "con la procedencia de su sobre")
	assert.JSONEq(t, string(sobre["url"]), `"`+lotes[0].URL+`"`)
	assert.JSONEq(t, string(sobre["fecha_consulta"]), `"`+lotes[0].FechaConsulta+`"`)
	assert.NotEmpty(t, lotes[0].Operaciones)

	inexistente := servidor.llamar(t, "boe_articulo", argumentosDelBloqueA9999)
	require.True(t, inexistente.Fallo, "%s", inexistente.Sobre)

	clase, _ := falloDe(t, inexistente.Sobre)
	assert.Equal(t, claseNoEncontrado, clase)
	assert.Len(t, almacen.recibidos(), 1, "una llamada que falla no entrega nada")

	assert.Equal(t, 0, servidor.cerrar(t))
}

func testEntregaQueFalla(t *testing.T) {
	t.Parallel()

	deReferencia, _ := registroDeConsulta(t)
	orden := invocarCon(t, deReferencia, "boe", "articulo", "BOE-A-2015-10565", "a21", "--json")
	require.Equal(t, 0, orden.codigo, orden.errores)

	registro, _ := registroDeConsulta(t)
	almacen := &almacenAnotado{fallo: errors.New("grafo: world.db no se puede escribir:\npermiso denegado")}
	registro.EntregarAlGrafo(almacen)

	servidor := arrancarElServidor(t, registro, false)

	leido := servidor.llamar(t, "boe_articulo", argumentosDelBloqueA21)

	assert.False(t, leido.Fallo, "la llamada terminó bien: su resultado no es un error")
	assert.Equal(t, orden.salida, string(leido.Sobre)+"\n", "el sobre es el de la orden: la entrega fallida no lo cambia")
	assert.Len(t, almacen.recibidos(), 1, "la entrega se intentó")

	assert.Equal(t, 0, servidor.cerrar(t))
	assert.Equal(t,
		[]string{lineaDeEntregaFallida + "grafo: world.db no se puede escribir: permiso denegado"},
		servidor.errores.lineasCon(lineaDeEntregaFallida),
		"la línea de H7, una vez, en la salida de error")
}

// grabacionDelBloqueA21 es el fichero de la reproducción con el bloque a21 de
// la Ley 39/2015, y redaccionPosterior, la carpeta de las derivadas con otra
// redacción suya, de fecha de vigencia posterior (contracts/arnes-e2e.md §3 de
// H7).
const (
	grabacionDelBloqueA21 = "GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_" +
		"BOE-A-2015-10565_texto_bloque_a21.json"
	redaccionPosterior = derivadasDelRepositorio + "/version-posterior"
)

// grabacionesConOtraRedaccion es una copia de las grabaciones de H4 en una
// carpeta del test, con la derivada de la redacción posterior en el sitio del
// bloque a21.
func grabacionesConOtraRedaccion(t *testing.T) string {
	t.Helper()

	carpeta := t.TempDir()
	require.NoError(t, os.CopyFS(carpeta, os.DirFS(grabacionesDeLaFuente)))

	derivada, err := fs.ReadFile(os.DirFS(redaccionPosterior), grabacionDelBloqueA21)
	require.NoError(t, err)

	raiz, err := os.OpenRoot(carpeta)
	require.NoError(t, err)

	defer func() { assert.NoError(t, raiz.Close()) }()

	require.NoError(t, raiz.Remove(grabacionDelBloqueA21))
	require.NoError(t, raiz.WriteFile(grabacionDelBloqueA21, derivada, 0o600))

	return carpeta
}

// testHallazgosQueNoSonUnError hace lo que hace la skill con un bloque cuya
// redacción ha cambiado desde la lectura anterior: una orden lo leyó con la
// grabación de H4 y, un día después y con el mismo grafo, el servidor lo lee
// con otra redacción y comprueba la norma. La comprobación, pedida detrás de
// la lectura, ve lo que esa acaba de entregar, y devuelve el sobre de su
// orden con `ok` verdadero y el hallazgo en `data`, sin marcarlo como error de
// herramienta (FR-012 y FR-013 de H21; ADR 0023).
func testHallazgosQueNoSonUnError(t *testing.T) {
	t.Parallel()

	grafoDelMundo := t.TempDir()

	primera := &boeDeLaReproduccion{
		grabaciones: grabacionesDeLaFuente,
		hora:        horaDeLaReproduccion,
		cache:       t.TempDir(),
		grafo:       grafoDelMundo,
	}

	deLaPrimera := primera.registro(t)
	deLaPrimera.EntregarAlGrafo(graph.Nuevo(graph.ConDirectorio(grafoDelMundo)))

	leida := invocarCon(t, deLaPrimera, "boe", "articulo", normaDeLasLlamadas, "a21", "--json")
	require.Equal(t, 0, leida.codigo, leida.errores)

	segunda := &boeDeLaReproduccion{
		grabaciones: grabacionesConOtraRedaccion(t),
		hora:        func() time.Time { return horaDeLaReproduccion().Add(24 * time.Hour) },
		cache:       t.TempDir(),
		grafo:       grafoDelMundo,
	}

	registro := segunda.registro(t)
	registro.EntregarAlGrafo(graph.Nuevo(graph.ConDirectorio(grafoDelMundo)))

	servidor := arrancarElServidor(t, registro, false)

	releida := servidor.llamar(t, "boe_articulo", argumentosDelBloqueA21)
	require.False(t, releida.Fallo, "%s", releida.Sobre)
	require.NotEqual(t, leida.salida, string(releida.Sobre)+"\n", "la fuente sirve ahora otra redacción")

	comprobado := servidor.llamar(t, "graph_check", `{"norma":"`+normaDeLasLlamadas+`","bloques":["a21"]}`)
	assert.False(t, comprobado.Fallo, "una comprobación que encuentra algo ha funcionado: %s", comprobado.Sobre)

	var datos struct {
		VersionObsoleta int               `json:"version-obsoleta"`
		Hallazgos       []json.RawMessage `json:"hallazgos"`
	}

	require.NoError(t, json.Unmarshal(sobreDe(t, comprobado.Sobre)["data"], &datos))
	assert.Equal(t, 1, datos.VersionObsoleta, "la redacción de la primera lectura está superada")
	assert.Len(t, datos.Hallazgos, 1, "y el hallazgo va en data")

	orden := invocarCon(t, registro, "graph", "check", normaDeLasLlamadas, "a21", "--json")
	require.Equal(t, 0, orden.codigo, orden.errores)
	assert.Equal(t, orden.salida, string(comprobado.Sobre)+"\n", "el sobre es el de la orden")

	assert.Equal(t, 0, servidor.cerrar(t), servidor.errores.lineas())
	assert.Empty(t, servidor.errores.lineasCon(lineaDeEntregaFallida))
}

func testSinGrafoEnCadaLlamada(t *testing.T) {
	t.Parallel()

	registro, _ := registroDeConsulta(t)
	almacen := &almacenAnotado{}
	registro.EntregarAlGrafo(almacen)

	servidor := arrancarElServidor(t, registro, false, "--no-graph")

	leido := servidor.llamar(t, "boe_articulo", argumentosDelBloqueA21)
	assert.False(t, leido.Fallo, "%s", leido.Sobre)

	resuelto := servidor.llamar(t, "territorio_resolver", argumentosDeLeganes)
	assert.False(t, resuelto.Fallo, "%s", resuelto.Sobre)

	assert.Equal(t, 0, servidor.cerrar(t))
	assert.Empty(t, almacen.recibidos(), "con mcp serve --no-graph ninguna llamada entrega nada")
}

func testSinRedEnCadaLlamada(t *testing.T) {
	t.Parallel()

	registro, deBoe := registroDeConsulta(t)
	servidor := arrancarElServidor(t, registro, false, "--offline")

	sinCache := servidor.llamar(t, "boe_articulo", argumentosDelBloqueA21)
	require.True(t, sinCache.Fallo, "%s", sinCache.Sobre)

	clase, _ := falloDe(t, sinCache.Sobre)
	assert.Equal(t, claseFuenteNoDisponible, clase, "con --offline y la caché vacía")
	assert.Zero(t, deBoe.construidos.Load(), "la llamada no construye ningún cliente de la fuente")

	resuelto := servidor.llamar(t, "territorio_resolver", argumentosDeLeganes)
	assert.False(t, resuelto.Fallo, "territorio no consulta ninguna fuente: %s", resuelto.Sobre)

	assert.Equal(t, 0, servidor.cerrar(t))
}

// TestSalidaDeErrorDelServidor comprueba lo que `mcp serve` deja en la salida
// de error mientras sirve (contracts/servidor-mcp.md §4 de H21): el aviso de
// versión una vez por arranque y no una por llamada; un evento por llamada, a
// nivel `warn` si falla y a `info` si termina bien, con el applet y el verbo de
// su herramienta; el mensaje de cada llamada que falla, que es el de su sobre;
// y, al terminar, el evento de `mcp serve` (FR-023). No es paralelo porque fija
// KITLEGAL_LOG, del que depende el nivel del registro de eventos.
func TestSalidaDeErrorDelServidor(t *testing.T) {
	t.Setenv(cli.VariableNivel, "")

	const aviso = "aviso: las skills instaladas son de otra versión"

	// Dos llamadas que terminan bien y dos que fallan: una por lo que responde
	// la fuente y otra por sus argumentos, que no llega a analizarse.
	llamar := func(t *testing.T, servidor *servidorEnProceso) (mensajes []string) {
		t.Helper()

		require.False(t, servidor.llamar(t, "territorio_resolver", argumentosDeLeganes).Fallo)
		require.False(t, servidor.llamar(t, "boe_articulo", argumentosDelBloqueA21).Fallo)

		for _, argumentos := range []string{argumentosDelBloqueA9999, `{"norma":"BOE-A-2015-10565","bloque":21}`} {
			fallida := servidor.llamar(t, "boe_articulo", argumentos)
			require.True(t, fallida.Fallo, "%s", fallida.Sobre)

			_, mensaje := falloDe(t, fallida.Sobre)
			mensajes = append(mensajes, mensaje)
		}

		return mensajes
	}

	registroConAviso := func(t *testing.T) *app.Registro {
		t.Helper()

		registro, _ := registroDeConsulta(t)
		registro.Avisar(func() (string, bool) { return aviso, true })

		return registro
	}

	t.Run("sin --verbose: el aviso una vez, y de cada llamada que falla su evento y su mensaje", func(t *testing.T) {
		t.Setenv(cli.VariableNivel, "")

		servidor := arrancarElServidor(t, registroConAviso(t), true)
		mensajes := llamar(t, servidor)

		require.Equal(t, 0, servidor.cerrar(t))

		lineas := servidor.errores.lineas()
		require.Len(t, lineas, 5, "el aviso y, por cada llamada que falla, su evento y su mensaje: %q", lineas)

		assert.Equal(t, aviso, lineas[0], "el aviso, antes de servir")

		assert.Contains(t, lineas[1], "level=WARN msg=invocación applet=boe verbo=articulo ")
		assert.Contains(t, lineas[1], " clase="+claseNoEncontrado)
		assert.Equal(t, mensajes[0], lineas[2], "el mensaje de la llamada que falla es el de su sobre")

		assert.Contains(t, lineas[3], "level=WARN msg=invocación applet=boe verbo=articulo ")
		assert.Contains(t, lineas[3], " clase="+claseArgumentos)
		assert.Equal(t, mensajes[1], lineas[4])
	})

	t.Run("con --verbose: además, el evento de cada llamada que termina bien y el de mcp serve", func(t *testing.T) {
		t.Setenv(cli.VariableNivel, "")

		servidor := arrancarElServidor(t, registroConAviso(t), true, "--verbose")
		llamar(t, servidor)

		require.Equal(t, 0, servidor.cerrar(t))

		assert.Equal(t, []string{aviso}, servidor.errores.lineasCon("aviso: "),
			"con cuatro llamadas hechas, el aviso está una vez: es del arranque, no de cada llamada")

		assert.Len(t, servidor.errores.lineasCon("level=INFO msg=invocación applet=territorio verbo=resolver "), 1)
		assert.Len(t, servidor.errores.lineasCon("level=INFO msg=invocación applet=boe verbo=articulo "), 1)
		assert.Len(t, servidor.errores.lineasCon("level=WARN msg=invocación applet=boe verbo=articulo "), 2)

		deMCP := servidor.errores.lineasCon("msg=invocación applet=mcp verbo=serve ")
		require.Len(t, deMCP, 1, "el evento de mcp serve, uno, al terminar")
		assert.Contains(t, deMCP[0], "level=INFO ", "el servidor termina bien")

		eventos := servidor.errores.lineasCon("msg=invocación ")
		assert.Equal(t, deMCP[0], eventos[len(eventos)-1], "y es el último")
	})
}

// TestPlazoDeCadaLlamada comprueba que --timeout, dado a `mcp serve`, es el
// plazo de cada llamada y nunca el de la vida del servidor (FR-020 de H21 y el
// caso límite «Una llamada que pasa del plazo de --timeout» del spec;
// contracts/servidor-mcp.md §3, «Otros casos»). En un guion no cabe: ningún
// verbo del binario de e2e se queda esperando (research.md D13).
//
// (1) La llamada al verbo que espera a su contexto devuelve, como error de
// herramienta, el sobre de la clase que da la orden cuando agota su plazo, y no
// antes de que pase. El verbo vuelve también, sin error, si el test lo suelta
// al pasar un tope generoso: una llamada a la que no llegue el plazo falla en
// esa aserción y no cuelga el test. (2) Una llamada enviada después, con el
// plazo ya vencido desde el arranque del servidor, recibe su sobre, y al
// cerrarse la entrada el servidor termina con 0 y sin sobre propio.
func TestPlazoDeCadaLlamada(t *testing.T) {
	t.Parallel()

	const (
		// plazo es holgado para que la llamada que no espera no lo roce con el
		// detector de carreras.
		plazo = time.Second
		tope  = 15 * plazo
	)

	espera := nuevoAppletDeEspera()

	arranque := time.Now()
	servidor := arrancarElServidor(t, registroCon(t, espera), true, "--timeout="+plazo.String())

	// El tope: pasado ese tiempo, el verbo que espera vuelve sin que su contexto
	// haya terminado, y la llamada, con un sobre que no es el del plazo agotado.
	alTope := time.AfterFunc(tope, espera.soltar)

	t.Cleanup(func() {
		alTope.Stop()
		espera.soltar()
	})

	// (1). Ninguna de sus aserciones detiene el test: (2) se comprueba pase lo
	// que pase aquí.
	comienzo := time.Now()
	agotada, err := servidor.pedir(t, "espera_esperar", json.RawMessage("{}"))
	transcurrido := time.Since(comienzo)

	if assert.NoError(t, err, "(1) la llamada que agota el plazo recibe su resultado: %s", servidor.errores.lineas()) {
		if assert.True(t, agotada.Fallo, "(1) que es un error de herramienta: %s", agotada.Sobre) {
			clase, _ := falloDe(t, agotada.Sobre)
			assert.Equal(t, claseFuenteNoDisponible, clase, "(1) con la clase que da la orden cuando agota su plazo")
		}

		assert.GreaterOrEqual(t, transcurrido, plazo, "(1) y no antes de que pase el plazo")
	}

	// (2)
	require.Greater(t, time.Since(arranque), plazo, "(2) el plazo ya ha vencido desde el arranque del servidor")

	resultado, err := servidor.pedir(t, "espera_volver", json.RawMessage("{}"))
	if assert.NoError(t, err,
		"(2) el servidor atiende una llamada pasado el plazo desde su arranque: %s", servidor.errores.lineas()) {
		assert.False(t, resultado.Fallo,
			"(2) y la llamada, que no espera, recibe su sobre con ok verdadero: %s", resultado.Sobre)
		assert.JSONEq(t, "true", string(sobreDe(t, resultado.Sobre)["ok"]))
	}

	assert.Equal(t, 0, servidor.cerrar(t),
		"(2) y al cerrarse la entrada termina con 0 y sin sobre propio: el plazo no es el de su vida: %s",
		servidor.errores.lineas())
}
