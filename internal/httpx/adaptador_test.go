// adaptador_test.go es el adaptador de prueba del hito y la demostración de que
// el cliente se usa bien desde donde se usará de verdad: un applet, ejecutado
// por el kernel entero.
//
// Es **material de test** y nada más (FR-060): vive en un fichero `_test.go`,
// no se crea bajo `internal/source/` —directorio que este hito no abre— y no se
// registra en ningún binario, ni en el que se publica ni en el de e2e, de
// modo que el conjunto de verbos de los dos queda exactamente como lo dejó H1
// (FR-062, SC-014). El único registro en el que aparece lo construye este test.
//
// Va en el paquete externo `httpx_test` a propósito (D16): compila únicamente
// contra la superficie exportada, que es la demostración literal de FR-061. No
// hay forma de escribir `Ejecutar` sin el contexto de quien llama ni sin el
// contexto de ejecución, porque `Pedir` es la única operación del paquete y
// exige los dos; un adaptador que quisiera saltarse el plazo, la identificación
// o el robots.txt no tendría a qué llamar.
package httpx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// fuenteDePrueba es el nombre lógico de la fuente del adaptador: el que declara
// `ConFuente` y el que firma la procedencia del sobre. Es también el nombre del
// applet, porque el adaptador de una fuente se invoca por ella.
const fuenteDePrueba = "prueba"

// adaptadorDePrueba es el applet que copia el patrón del contrato del cliente
// §6: declara su identidad y su catálogo de verbos, y nada más. No conoce
// ninguna bandera, ningún código de salida y ninguna forma de presentación; todo
// eso lo hereda del kernel.
type adaptadorDePrueba struct{}

func (adaptadorDePrueba) Nombre() string { return fuenteDePrueba }

func (adaptadorDePrueba) Descripcion() string {
	return "Consulta una dirección con el cliente de internal/httpx."
}

func (adaptadorDePrueba) Verbos() []app.Verbo {
	return []app.Verbo{{
		Nombre:      "consultar",
		Descripcion: "Pide la dirección y devuelve el cuerpo que entregó la fuente.",
		Argumentos:  func() app.Argumentos { return &argumentosConsultar{} },
		Salida:      cuerpoDeLaFuente{},
	}}
}

// cuerpoDeLaFuente es el contenido de `data` del verbo: lo que el adaptador
// extrae de la respuesta. Se declara como tipo y no como un mapa suelto porque
// de él sale la mitad de salida del esquema que emite --describe.
type cuerpoDeLaFuente struct {
	Cuerpo string `json:"cuerpo"`
}

// argumentosConsultar son los argumentos del verbo `consultar <dirección>`, con
// las etiquetas de las que el kernel construye la gramática.
type argumentosConsultar struct {
	URL string `arg:"" help:"Dirección que se consulta."`
}

// Ejecutar es el patrón que copiará cada adaptador de fuente (contrato del
// cliente §6). Recibe el contexto de quien le llama y el contexto de ejecución
// del kernel, y **propaga los dos** a `Pedir`: no puede compilar de otra forma,
// porque `Pedir` es la única operación del paquete y exige ambos (FR-059,
// FR-061).
//
// Del plazo, de la identificación, del robots.txt, del ritmo del sitio y de los
// reintentos no dice nada, y no es que confíe en ellos: no tiene ninguna opción
// con la que desactivarlos. Del ensayo sí se ocupa, porque es lo único que el
// cliente no puede decidir por él: dónde va la descripción de lo que no se hizo
// (FR-051, FR-065, D6).
func (a *argumentosConsultar) Ejecutar(
	ctx context.Context, ejecucion schema.Contexto, registrador *slog.Logger,
) (schema.Resultado, error) {
	cliente, err := httpx.New(httpx.ConFuente(fuenteDePrueba), httpx.ConRegistrador(registrador))
	if err != nil {
		return schema.Resultado{}, err // ya lleva su clase
	}

	respuesta, err := cliente.Pedir(ctx, ejecucion, httpx.Peticion{Metodo: http.MethodGet, URL: a.URL})
	if err != nil {
		return schema.Resultado{}, fmt.Errorf("consultando %s: %w", a.URL, err) // la clase no cambia
	}

	resultado := schema.Resultado{
		Procedencia: schema.Procedencia{Fuente: fuenteDePrueba, URL: respuesta.URL},
	}

	if respuesta.Ensayo {
		resultado.Ensayo = []string{respuesta.Descripcion()}

		return resultado, nil
	}

	resultado.Datos = cuerpoDeLaFuente{Cuerpo: string(respuesta.Cuerpo)}

	return resultado, nil
}

// Las dos comprobaciones en tiempo de compilación del contrato del applet.
var (
	_ app.Applet     = adaptadorDePrueba{}
	_ app.Argumentos = (*argumentosConsultar)(nil)
)

// TestAdaptadorDePruebaConElKernel ejercita el adaptador como se ejercitará
// cualquier adaptador de fuente: invocando al kernel en proceso —`app.Main`
// sobre un registro construido aquí— contra un servidor local. Es la
// comprobación de extremo a extremo que ningún guion `testscript` puede hacer,
// porque `testscript` solo observa binarios y no sabe levantar el servidor
// (spec, *Fuera de alcance*; D16).
func TestAdaptadorDePruebaConElKernel(t *testing.T) {
	t.Parallel()

	t.Run("ensayo", func(t *testing.T) {
		t.Parallel()

		servidor, accesos := servidorDeLaFuente(t)

		invocacion := invocarAlKernel(t, fuenteDePrueba, "consultar", servidor.URL+rutaDeLaNorma, "--dry-run")

		assert.Equal(t, 0, invocacion.codigo,
			"el ensayo no es un fallo: la invocación entera —kernel, adaptador y cliente— sale con 0 (SC-013, FR-065)")
		assert.Zero(t, accesos.recurso.Load(), "en ensayo no se pide el recurso (FR-050)")
		assert.Zero(t, accesos.robots.Load(), "ni el robots.txt del sitio, que es la otra conexión que habría (FR-050)")
		assert.Contains(t, invocacion.errores, "se habría pedido GET "+servidor.URL+rutaDeLaNorma,
			"la descripción de la operación que no se emitió sale por la salida de error (FR-051, FR-052)")
		assert.Empty(t, invocacion.salida,
			"y la salida estándar queda vacía: sin sobre y sin ningún cuerpo inventado (FR-052)")
	})

	t.Run("consulta", func(t *testing.T) {
		t.Parallel()

		servidor, accesos := servidorDeLaFuente(t)

		invocacion := invocarAlKernel(t, fuenteDePrueba, "consultar", servidor.URL+rutaDeLaNorma, "--json")

		require.Equal(t, 0, invocacion.codigo, "la consulta termina bien; salida de error: %s", invocacion.errores)

		sobre := sobreDeLaInvocacion(t, invocacion.salida)
		assert.Equal(t, true, sobre["ok"])
		assert.Equal(t, fuenteDePrueba, sobre["fuente"])
		assert.Equal(t, servidor.URL+rutaDeLaNorma, sobre["url"], "el sobre cita la dirección que entregó el contenido")
		assert.Equal(t, map[string]any{"cuerpo": contenidoDeLaFuente}, sobre["data"],
			"la respuesta llega íntegra hasta quien llamó: el cuerpo entero, ya leído (US5 escenario 1)")

		assert.Equal(t, int64(1), accesos.recurso.Load(), "se pidió el recurso una vez")
		assert.Equal(t, int64(1), accesos.robots.Load(), "y antes el robots.txt del sitio, una sola vez (FR-013)")
		assert.Zero(t, accesos.sinIdentificar.Load(),
			"y ninguna de las dos llegó sin la identificación exacta: el adaptador no tiene forma de quitarla (FR-009, FR-059)")
	})

	t.Run("plazo", func(t *testing.T) {
		t.Parallel()

		servidor, _ := servidorDeLaFuente(t)

		comienzo := time.Now()
		invocacion := invocarAlKernel(t, fuenteDePrueba, "consultar", servidor.URL+rutaQueNoResponde,
			"--json", "--timeout", plazoDeLaPrueba.String())

		assert.Equal(t, 4, invocacion.codigo,
			"el vencimiento del contexto es «fuente no disponible», y el kernel lo traduce al código 4 (FR-029, FR-063)")
		assert.Less(t, time.Since(comienzo), timeoutPorOmision,
			"y corta con el contexto, no con el plazo por omisión del kernel (FR-004, FR-005)")

		sobre := sobreDeLaInvocacion(t, invocacion.salida)
		assert.Equal(t, false, sobre["ok"])

		datos := datosDelFallo(t, sobre)
		assert.Equal(t, string(schema.ClaseFuenteNoDisponible), datos["clase"],
			"la clase que el cliente declara llega intacta al sobre aunque el adaptador la envolviera (FR-063)")
		assert.Contains(t, datos["mensaje"], rutaQueNoResponde,
			"y el mensaje nombra lo que se estaba consultando")
	})
}

// Las dos rutas que publica el servidor de la fuente y el contenido que entrega.
// La segunda no responde nunca por su cuenta: quien termina la operación es el
// contexto de quien la pidió.
const (
	rutaDelRobots       = "/robots.txt"
	rutaDeLaNorma       = "/norma"
	rutaQueNoResponde   = "/lenta"
	contenidoDeLaFuente = "<norma>contenido</norma>"
)

// plazoDeLaPrueba es el --timeout con el que se invoca la subprueba del plazo.
// Es mayor que el intervalo por omisión del ritmo (1 s) a propósito: así la
// petición del recurso llega a emitirse —el turno del sitio le toca— y lo que la
// corta es el vencimiento del contexto mientras se espera a la fuente, que es
// justo lo que la subprueba comprueba. Con un plazo más corto cortaría el
// limitador, que es otra fila de la tabla de clases y ya la cubre T014.
const plazoDeLaPrueba = 1500 * time.Millisecond

// timeoutPorOmision es el plazo que el kernel aplica cuando la invocación no
// trae --timeout (contrato de banderas de H1). Aquí solo sirve de techo: si la
// operación hubiera durado esto, el plazo de la invocación no habría cortado
// nada.
const timeoutPorOmision = 30 * time.Second

// Los datos de construcción con los que se invoca a la raíz de composición.
// Ninguna de las tres subpruebas mira el verbo reservado «version»: lo que
// importa es que el kernel los reciba y no los interprete.
const (
	nombreDelBinario = "kitlegal"
	versionDePrueba  = "dev"
	commitDePrueba   = "none"
	fechaDePrueba    = "unknown"
)

// accesosAlServidor cuenta lo que de verdad llegó al servidor, que es lo único
// que distingue «no se emitió» de «se emitió y se descartó»: las peticiones del
// recurso, las del robots.txt del sitio —que no son ninguna de las que quien
// llama pidió, pero también cuentan— y las que llegaron sin la identificación
// exacta, que tiene que ser siempre cero.
type accesosAlServidor struct {
	recurso        atomic.Int64
	robots         atomic.Int64
	sinIdentificar atomic.Int64
}

// servidorDeLaFuente levanta la fuente local del adaptador: un robots.txt sin
// reglas —que permite toda ruta—, la norma y la ruta que no responde. Ningún
// test del hito conoce una dirección que no sea esta.
func servidorDeLaFuente(t *testing.T) (*httptest.Server, *accesosAlServidor) {
	t.Helper()

	accesos := &accesosAlServidor{}

	servidor := httptest.NewServer(http.HandlerFunc(func(escritor http.ResponseWriter, peticion *http.Request) {
		if peticion.Header.Get("User-Agent") != httpx.AgenteDeUsuario() {
			accesos.sinIdentificar.Add(1)
		}

		switch peticion.URL.Path {
		case rutaDelRobots:
			accesos.robots.Add(1)
			escritor.Header().Set("Content-Type", "text/plain; charset=utf-8")
		case rutaDeLaNorma:
			accesos.recurso.Add(1)
			escritor.Header().Set("Content-Type", "application/xml; charset=utf-8")
			_, _ = io.WriteString(escritor, contenidoDeLaFuente)
		case rutaQueNoResponde:
			accesos.recurso.Add(1)
			<-peticion.Context().Done()
		default:
			accesos.recurso.Add(1)
			http.NotFound(escritor, peticion)
		}
	}))
	t.Cleanup(servidor.Close)

	return servidor, accesos
}

// invocacionDelKernel es en qué queda una invocación: lo único que un consumidor
// del binario puede mirar.
type invocacionDelKernel struct {
	codigo  int
	salida  string
	errores string
}

// invocarAlKernel ejecuta la raíz de composición entera sobre un registro que
// solo existe aquí dentro. Es el único lugar del proyecto donde el adaptador de
// prueba se registra: ningún binario lo enlaza (FR-062, SC-014).
//
// Que esta función retorne y el test siga vivo después es además la comprobación
// de que ninguna ruta llama a os.Exit.
func invocarAlKernel(t *testing.T, argv ...string) invocacionDelKernel {
	t.Helper()

	var registro app.Registro

	require.NoError(t, registro.Registrar(adaptadorDePrueba{}),
		"el adaptador de prueba declara un applet válido")

	var salida, errores bytes.Buffer

	codigo := app.Main(append([]string{nombreDelBinario}, argv...), &registro, &salida, &errores,
		versionDePrueba, commitDePrueba, fechaDePrueba)

	return invocacionDelKernel{codigo: codigo, salida: salida.String(), errores: errores.String()}
}

// sobreDeLaInvocacion analiza la salida estándar de una invocación con --json.
// Comprueba de paso que lleva un único documento y nada más.
func sobreDeLaInvocacion(t *testing.T, salida string) map[string]any {
	t.Helper()

	decodificador := json.NewDecoder(strings.NewReader(salida))

	var sobre map[string]any

	require.NoError(t, decodificador.Decode(&sobre), "la salida estándar es el sobre: %q", salida)
	require.ErrorIs(t, decodificador.Decode(new(json.RawMessage)), io.EOF,
		"y no hay nada más después del sobre")

	return sobre
}

// datosDelFallo devuelve el contenido de `data` de un sobre de fallo: las dos
// claves del contrato y ninguna más, ya comprobado que son texto. El mensaje no
// se fija aquí carácter a carácter —lo escribe el cliente y lo comprueba
// errores_test.go—; lo que esta subprueba mira es qué clase llega y qué nombra.
func datosDelFallo(t *testing.T, sobre map[string]any) map[string]string {
	t.Helper()

	objeto, esMapa := sobre["data"].(map[string]any)
	require.True(t, esMapa, "el sobre de fallo lleva un objeto en data")

	datos := make(map[string]string, len(objeto))

	for clave, valor := range objeto {
		texto, esTexto := valor.(string)
		require.Truef(t, esTexto, "la clave %q del sobre de fallo es texto", clave)

		datos[clave] = texto
	}

	require.Len(t, datos, 2, "data lleva la clase y el mensaje, ni una clave más")

	return datos
}
