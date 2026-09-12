package httpx

import (
	"context"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// direccionDelEscalon es la dirección con la que se prueba el decorador suelto,
// sin servidor: el escalón que lleva debajo no abre ninguna conexión, así que la
// dirección solo tiene que ser interpretable y de un host de prueba.
const direccionDelEscalon = "http://fuente.prueba/norma"

// TestReintentosDosErroresYUnAcierto es el control literal de SC-005: ante dos
// errores de servidor seguidos y un acierto, quien llama recibe la respuesta
// correcta tras exactamente tres peticiones —las tres con la identificación del
// proyecto—, y las esperas entre intentos crecen y no son iguales en dos
// ejecuciones distintas.
//
// Las tres llegadas las separa el ritmo del sitio, no el retardo de los
// reintentos, que aquí no se espera: es la prueba de que el decorador va **por
// encima** del limitador y de que cada reintento ocupa su turno igual que la
// primera petición (FR-021, D3).
func TestReintentosDosErroresYUnAcierto(t *testing.T) {
	t.Parallel()

	primera := dosErroresYUnAcierto(t)
	segunda := dosErroresYUnAcierto(t)

	assert.NotEqual(t, primera, segunda,
		"la parte aleatoria del retardo impide que dos ejecuciones esperen lo mismo (SC-005, D9)")
}

// dosErroresYUnAcierto ejecuta una vez el escenario de SC-005 y devuelve las
// esperas que el decorador pidió, que son las que las dos ejecuciones comparan.
func dosErroresYUnAcierto(t *testing.T) []time.Duration {
	t.Helper()

	// El intervalo es el que separa las llegadas. La holgura es la del test del
	// ritmo, dos órdenes de magnitud por debajo: lo que el limitador espacia con
	// exactitud es el turno, y la llegada le añade lo que tarde en despacharse
	// el manejador.
	const intervalo = 100 * time.Millisecond

	const holgura = 10 * time.Millisecond

	anotador := &llegadas{}
	esperas := &esperasAnotadas{}
	servidor, contador := servidorIdentificado(t,
		anotador.anota(servidorQueFalla(2, http.StatusServiceUnavailable)))
	cliente := clienteDePrueba(t, ConIntervalo(intervalo), conReloj(esperas.reloj))

	respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
		Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})
	require.NoError(t, err, "el tercer intento acierta, así que quien llama recibe la respuesta (SC-005)")

	assert.Equal(t, http.StatusOK, respuesta.Estado)
	assert.Equal(t, contenidoDePrueba, string(respuesta.Cuerpo))
	assert.Equal(t, int64(3), contador.total.Load(),
		"dos errores de servidor y un acierto son exactamente tres peticiones (SC-005)")
	assert.Zero(t, contador.sinIdentificar.Load(),
		"cada reintento vuelve a bajar por la cadena y llega identificado (FR-009)")

	anotadas := anotador.instantes()
	require.Len(t, anotadas, 3)

	for intento := 1; intento < len(anotadas); intento++ {
		assert.GreaterOrEqual(t, anotadas[intento].Sub(anotadas[intento-1]), intervalo-holgura,
			"cada reintento espera su turno en el sitio: el decorador va por encima del ritmo (FR-021)")
	}

	pedidas := esperas.pedidas()
	require.Len(t, pedidas, 2, "entre tres intentos hay dos esperas")

	for intento, pedida := range pedidas {
		exigeEsperaDelIntento(t, intento+1, pedida)
	}

	assert.Greater(t, pedidas[1], pedidas[0], "las esperas entre intentos crecen (SC-005)")

	return pedidas
}

// TestReintentosNoRepite4xx fija la mitad cerrada de FR-025: lo que se vuelve a
// pedir son los errores del servidor y los fallos de transporte, y nada más. Un
// 4xx no mejora por repetirlo —el 429 incluido, que además tiene clase propia y
// trae su propia espera (FR-030)—, así que el servidor ve una sola petición.
//
// El servidor de cada caso falla **una vez** y acierta después: si hubiera
// reintento, el contador marcaría dos y el resultado sería el 200 del segundo.
func TestReintentosNoRepite4xx(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		estado    int
		entregada bool
		clase     schema.Clase
	}{
		{
			nombre:    "el «no encontrado» se entrega tal cual y no se repite",
			estado:    http.StatusNotFound,
			entregada: true,
		},
		{
			nombre:    "una petición mal formada no mejora por repetirla",
			estado:    http.StatusBadRequest,
			entregada: true,
		},
		{
			nombre: "el límite de peticiones tiene su clase y tampoco se repite",
			estado: http.StatusTooManyRequests,
			clase:  schema.ClaseLimiteOTos,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			esperas := &esperasAnotadas{}
			servidor, contador := servidorIdentificado(t, servidorQueFalla(1, caso.estado))
			cliente := clienteDePrueba(t, ConIntervalo(time.Millisecond), conReloj(esperas.reloj))

			respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
				Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})

			if caso.entregada {
				require.NoError(t, err, "un 4xx que no es 429 se entrega a quien llama (FR-032)")
				assert.Equal(t, caso.estado, respuesta.Estado, "y se entrega con el estado que la fuente dio")
			} else {
				fallo := falloDe(t, err)
				assert.Equal(t, caso.clase, fallo.Clase(), "el 429 es «límite de peticiones o términos de uso» (FR-030)")
				assert.Equal(t, Respuesta{}, respuesta)
			}

			assert.Equal(t, int64(1), contador.total.Load(),
				"ningún 4xx se reintenta: el servidor ha visto una sola petición (FR-025)")
			assert.Empty(t, esperas.pedidas(), "y sin reintento no hay ninguna espera entre intentos")
		})
	}
}

// TestReintentosCancelacionGana comprueba FR-026 en las dos formas en que el
// contexto puede terminar durante la política: mientras se espera para volver a
// intentarlo y mientras la petición está en vuelo. En las dos, la cancelación
// gana —no hay ningún intento posterior— y lo que quien llama recibe es la
// clase de la operación que no llegó a la fuente (FR-029).
func TestReintentosCancelacionGana(t *testing.T) {
	t.Parallel()

	t.Run("la cancelación durante la espera no deja emitir el intento siguiente", func(t *testing.T) {
		t.Parallel()

		servidor, contador := servidorIdentificado(t,
			servidorQueFalla(intentosPorOmision, http.StatusServiceUnavailable))

		ctx, cancelar := context.WithCancel(t.Context())
		defer cancelar()

		// El reloj cancela y espera después por el mismo camino interrumpible
		// que el de verdad: lo que se comprueba es que la espera termina con el
		// contexto y que detrás de ella no queda ningún intento.
		cancelaAlEsperar := func(ctx context.Context, espera time.Duration) error {
			cancelar()

			return dormirInterrumpible(ctx, espera)
		}

		cliente := clienteDePrueba(t, ConIntervalo(time.Millisecond), conReloj(cancelaAlEsperar))

		comienzo := time.Now()
		respuesta, err := cliente.Pedir(ctx, schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/caida"})

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseFuenteNoDisponible, fallo.Clase(),
			"la operación que termina con su contexto es «la fuente no sabe entregar el recurso» (FR-029)")
		require.ErrorIs(t, err, context.Canceled, "la causa del contexto llega intacta hasta quien llama")
		assert.Contains(t, fallo.Error(), "reintentar", "el mensaje dice qué terminó la espera (FR-034)")
		assert.Equal(t, Respuesta{}, respuesta, "una operación cancelada no entrega ninguna respuesta")
		assert.Equal(t, int64(1), contador.total.Load(),
			"no hay ningún intento posterior a la cancelación (FR-026)")
		assert.Less(t, time.Since(comienzo), esperaBase,
			"la espera termina con el contexto y no con el retardo, que nunca baja de media base")
	})

	t.Run("la cancelación con la petición en vuelo no se reintenta", func(t *testing.T) {
		t.Parallel()

		ctx, cancelar := context.WithCancel(t.Context())
		defer cancelar()

		esperas := &esperasAnotadas{}
		servidor, contador := servidorIdentificado(t,
			func(_ http.ResponseWriter, peticion *http.Request) {
				// Cancelar y no responder: así el intento termina siempre en el
				// contexto y nunca en una respuesta que la carrera pudiera
				// colar. El manejador vuelve en cuanto el cliente cuelga.
				cancelar()
				<-peticion.Context().Done()
			})
		cliente := clienteDePrueba(t, ConIntervalo(time.Millisecond), conReloj(esperas.reloj))

		respuesta, err := cliente.Pedir(ctx, schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseFuenteNoDisponible, fallo.Clase(),
			"el contexto que termina en cualquier punto de la operación es clase 4 (FR-029)")
		require.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, Respuesta{}, respuesta)
		assert.Equal(t, int64(1), contador.total.Load(),
			"un fallo con el contexto ya terminado no se reintenta (FR-026)")
		assert.Empty(t, esperas.pedidas(), "ni se espera para volver a intentarlo")
	})
}

// TestReintentosCierraCuerposDescartados comprueba FR-027 donde se puede ver.
// El cuerpo de una respuesta que se descarta entre intentos se cierra —y se lee
// entero, para que la conexión vuelva al almacén y el intento siguiente la
// reutilice en vez de abrir otra—, y el de la respuesta que sí sube queda
// intacto, porque cerrarlo es de quien la lee.
//
// El decorador se monta aquí sobre un escalón de mentira, y no sobre un
// servidor, porque es ahí —en el cuerpo que el escalón entregó— donde se puede
// mirar si se cerró.
func TestReintentosCierraCuerposDescartados(t *testing.T) {
	t.Parallel()

	escalon := &escalonQueFalla{estado: http.StatusBadGateway, aciertaEn: intentosPorOmision}
	esperas := &esperasAnotadas{}
	decorador := conReintentos(escalon, intentosPorOmision, esperas.reloj)

	peticion, err := nuevaPeticionIdentificada(t.Context(), http.MethodGet, direccionDelEscalon)
	require.NoError(t, err)

	respuesta, err := decorador.RoundTrip(peticion)
	require.NoError(t, err)
	require.NotNil(t, respuesta)

	entregados := escalon.cuerpos()
	require.Len(t, entregados, intentosPorOmision, "el decorador ha bajado tres veces")

	for intento, descartado := range entregados[:len(entregados)-1] {
		assert.Equal(t, 1, descartado.cerrados(),
			"el cuerpo descartado del intento %d se cierra, y una sola vez (FR-027)", intento+1)
		assert.True(t, descartado.agotado(),
			"y se lee entero, para que la conexión se pueda reutilizar (FR-027)")
	}

	entregado := entregados[len(entregados)-1]
	assert.Zero(t, entregado.cerrados(), "el cuerpo de la respuesta que sube no lo cierra el decorador")

	cuerpo, err := io.ReadAll(respuesta.Body)
	require.NoError(t, err)
	require.NoError(t, respuesta.Body.Close())
	assert.Equal(t, contenidoDePrueba, string(cuerpo), "y llega entero a quien lo lee")
}

// TestReintentosAgotados cierra la política por el otro extremo: cuando el
// sitio no deja de fallar, el número de peticiones emitidas es exactamente el
// de intentos declarados —ni uno más, FR-024 y FR-028— y lo que sobrevive a la
// política se entrega con la clase de la fuente que no sabe dar el recurso
// (FR-029). La fila de ocho intentos cubre además la ley del retardo hasta el
// techo, que los tres por omisión no llegan a tocar.
func TestReintentosAgotados(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		opciones []Opcion
		intentos int
	}{
		{
			nombre:   "sin declarar nada son los tres intentos por omisión",
			intentos: intentosPorOmision,
		},
		{
			nombre:   "un solo intento es la forma de no reintentar",
			opciones: []Opcion{ConIntentos(1)},
			intentos: 1,
		},
		{
			nombre:   "ocho intentos recorren el retardo hasta su techo",
			opciones: []Opcion{ConIntentos(8)},
			intentos: 8,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			esperas := &esperasAnotadas{}
			servidor, contador := servidorIdentificado(t,
				servidorQueFalla(caso.intentos, http.StatusServiceUnavailable))
			opciones := append([]Opcion{ConIntervalo(time.Millisecond), conReloj(esperas.reloj)}, caso.opciones...)
			cliente := clienteDePrueba(t, opciones...)

			respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
				Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/caida"})

			fallo := falloDe(t, err)
			assert.Equal(t, schema.ClaseFuenteNoDisponible, fallo.Clase(),
				"el 5xx que sobrevive a la política es «la fuente no sabe entregar el recurso» (FR-029)")
			assert.Equal(t, http.StatusServiceUnavailable, fallo.Estado, "y conserva el estado que la fuente dio")
			assert.Equal(t, Respuesta{}, respuesta)
			assert.Equal(t, int64(caso.intentos), contador.total.Load(),
				"el número de intentos está acotado y no se pasa de él (FR-024, FR-028)")
			assert.Zero(t, contador.sinIdentificar.Load(), "cada intento llega identificado (FR-009)")

			pedidas := esperas.pedidas()
			require.Len(t, pedidas, caso.intentos-1, "entre n intentos hay n-1 esperas")

			for intento, pedida := range pedidas {
				exigeEsperaDelIntento(t, intento+1, pedida)
			}

			assert.IsIncreasing(t, pedidas, "cada espera arranca donde termina la anterior (SC-005, D9)")
		})
	}

	t.Run("un fallo de transporte se repite igual que un error del servidor", func(t *testing.T) {
		t.Parallel()

		esperas := &esperasAnotadas{}
		servidor, contador := servidorIdentificado(t, redireccionesDePrueba())
		direccion := servidor.URL + "/norma"
		cliente := clienteDePrueba(t, ConIntervalo(time.Millisecond), conReloj(esperas.reloj))

		// Una primera petición deja el robots.txt del sitio ya obtenido y
		// cacheado, de modo que lo que el cierre tumbe después sea la petición
		// del recurso y no la del permiso, que tiene otra clase (FR-015).
		_, err := cliente.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: direccion})
		require.NoError(t, err, "el sitio todavía escucha")
		require.Empty(t, esperas.pedidas(), "y no ha hecho falta reintentar nada")

		// Un servidor que se cierra antes de la petición es la forma
		// determinista de que no haya respuesta ninguna: el intento termina en
		// el transporte, que es la otra mitad de FR-025.
		servidor.Close()

		respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: direccion})

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseFuenteNoDisponible, fallo.Clase(),
			"el transporte que no llega a la fuente es clase 4 (FR-029)")
		assert.Equal(t, Respuesta{}, respuesta)
		assert.Equal(t, int64(1), contador.total.Load(),
			"el sitio ya no escucha: ninguna petición posterior al cierre llegó a atenderse")
		assert.Len(t, esperas.pedidas(), intentosPorOmision-1,
			"y aun así se intentó tantas veces como un 5xx (FR-025)")
	})
}

// exigeEsperaDelIntento comprueba que la espera pedida antes del intento
// siguiente al que se nombra cae donde D9 la deja: el *equal jitter* sobre
// b = min(500 ms · 2^(n−1), 30 s) la mete en [b/2, b). La ley se reescribe aquí
// con sus valores literales —y no con las constantes del paquete— para que sea
// un contraste independiente y no un eco de lo implementado.
func exigeEsperaDelIntento(t *testing.T, intento int, pedida time.Duration) {
	t.Helper()

	base := 500 * time.Millisecond
	for range intento - 1 {
		base = min(base*2, 30*time.Second)
	}

	assert.GreaterOrEqual(t, pedida, base/2,
		"la espera del intento %d no baja de la mitad de su base: el equal jitter solo aleatoriza la otra mitad (D9)",
		intento)
	assert.Less(t, pedida, base,
		"y no alcanza la base del intento %d, que es a la vez su techo y el suelo de la siguiente (SC-005, D9)",
		intento)
}

// esperasAnotadas es el reloj de estos tests: anota lo que el decorador pide
// esperar y espera en su lugar una duración despreciable, por el mismo camino
// interrumpible que el reloj de verdad. Así la política se comprueba sobre las
// duraciones pedidas —que es donde vive— sin que la prueba tarde los minutos
// que costaría esperarlas (D9).
type esperasAnotadas struct {
	mu       sync.Mutex
	anotadas []time.Duration
}

// reloj es lo que conReloj instala en el decorador en lugar de la espera real.
func (e *esperasAnotadas) reloj(ctx context.Context, espera time.Duration) error {
	e.mu.Lock()
	e.anotadas = append(e.anotadas, espera)
	e.mu.Unlock()

	return dormirInterrumpible(ctx, 0)
}

// pedidas devuelve una copia de lo anotado hasta ahora, de modo que el test lea
// las esperas sin compartir la lista con el decorador.
func (e *esperasAnotadas) pedidas() []time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()

	return slices.Clone(e.anotadas)
}

// servidorQueFalla responde con el estado declarado a las primeras peticiones y
// entrega el contenido a partir de la siguiente. Que acierte **después** de los
// fallos es lo que convierte cada tabla en una cuenta exacta: un reintento de
// más se ve en el contador y en el estado de la respuesta.
func servidorQueFalla(fallos, estado int) http.HandlerFunc {
	var atendidas atomic.Int64

	return func(escritor http.ResponseWriter, _ *http.Request) {
		if atendidas.Add(1) <= int64(fallos) {
			escritor.WriteHeader(estado)

			return
		}

		escritor.Header().Set("Content-Type", "application/xml; charset=utf-8")
		_, _ = io.WriteString(escritor, contenidoDePrueba)
	}
}

// escalonQueFalla es el escalón de mentira que el decorador lleva debajo en el
// test de los cuerpos: responde con el estado declarado hasta el intento que
// acierta y se queda con cada cuerpo que entregó, para que el test pueda
// mirarlos después.
type escalonQueFalla struct {
	estado    int
	aciertaEn int

	mu         sync.Mutex
	entregados []*cuerpoVigilado
}

// RoundTrip entrega la respuesta del intento que toca, con un cuerpo vigilado.
func (e *escalonQueFalla) RoundTrip(peticion *http.Request) (*http.Response, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	cuerpo := nuevoCuerpoVigilado(contenidoDePrueba)
	e.entregados = append(e.entregados, cuerpo)

	estado := e.estado
	if len(e.entregados) >= e.aciertaEn {
		estado = http.StatusOK
	}

	return &http.Response{
		StatusCode: estado,
		Header:     make(http.Header),
		Body:       cuerpo,
		Request:    peticion,
	}, nil
}

// cuerpos devuelve una copia de los cuerpos entregados, en orden de intento.
func (e *escalonQueFalla) cuerpos() []*cuerpoVigilado {
	e.mu.Lock()
	defer e.mu.Unlock()

	return slices.Clone(e.entregados)
}

// cuerpoVigilado es el cuerpo de una respuesta de mentira que recuerda cuántas
// veces lo cerraron y si queda algo por leer, que es lo único que hace
// observable la garantía de FR-027.
type cuerpoVigilado struct {
	mu      sync.Mutex
	lector  *strings.Reader
	cierres int
}

// nuevoCuerpoVigilado construye el cuerpo con el contenido que va a entregar.
func nuevoCuerpoVigilado(contenido string) *cuerpoVigilado {
	return &cuerpoVigilado{lector: strings.NewReader(contenido)}
}

// Read entrega el contenido como lo haría el cuerpo de una respuesta real.
func (c *cuerpoVigilado) Read(destino []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.lector.Read(destino)
}

// Close anota el cierre. Nunca falla: lo que el test mide es que se llame.
func (c *cuerpoVigilado) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cierres++

	return nil
}

// cerrados es cuántas veces se cerró este cuerpo.
func (c *cuerpoVigilado) cerrados() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.cierres
}

// agotado dice si el cuerpo se leyó entero, que es la otra mitad de FR-027: sin
// vaciarlo, la conexión no vuelve al almacén y el intento siguiente abre otra.
func (c *cuerpoVigilado) agotado() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.lector.Len() == 0
}
