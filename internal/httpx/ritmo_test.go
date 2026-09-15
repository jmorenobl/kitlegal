package httpx

import (
	"context"
	"net/http"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestRitmoSeparaPeticionesDelMismoSitio es la primera mitad del control literal
// de SC-004: con el ritmo fijado a una petición por intervalo, dos peticiones
// seguidas al mismo sitio llegan al servidor separadas al menos ese intervalo.
// Se mide donde el criterio lo dice —en las llegadas al servidor, no en lo que
// tarda quien llama— y sin que el cliente emita nada por su cuenta (FR-019,
// FR-021).
//
// El limitador separa los turnos, anclados al primero, que ocupa la primera
// petición que el cliente hace al sitio, la de su robots.txt: la petición i
// no sale antes de i veces el intervalo tras ese turno cero. Por eso la medida
// empieza antes del robots.txt y cada llegada se compara con su turno contado
// desde ahí. Si el cubo admitiera más de un token sería justo la primera
// petición tras el robots.txt la que saldría sin esperar, y medir solo a partir
// de la segunda lo dejaría pasar: por eso se anota también la llegada del
// robots.txt y se exige que **ninguna** se adelante a su turno, y por eso la
// operación entera tiene que ocupar el intervalo tantas veces como peticiones
// se hacen (D8).
func TestRitmoSeparaPeticionesDelMismoSitio(t *testing.T) {
	t.Parallel()

	const intervalo = 150 * time.Millisecond

	const peticiones = 3

	anotador := &llegadas{}
	contador := &contadorDeIdentificacion{}

	// El anotador envuelve también el robots.txt del sitio, no solo el recurso:
	// la llegada del robots.txt es la primera que el ritmo espacia.
	servidor := servidorLocal(t, contador.vigila(anotador.anota(robotsSinReglas(redireccionesDePrueba()))))
	cliente := clienteDePrueba(t, ConIntervalo(intervalo))

	comienzo := time.Now()

	for range peticiones {
		respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, respuesta.Estado)
	}

	assert.Equal(t, int64(peticiones), contador.total.Load(), "el ritmo separa las peticiones, no las descarta")
	assert.Equal(t, int64(1), contador.robots.Load(), "y el robots.txt del sitio es la primera de todas (FR-013)")

	// Con un solo token en el cubo, el robots.txt lo consume al instante y cada
	// una de las tres peticiones espera el suyo: la operación entera ocupa al
	// menos tres veces el intervalo. Con ráfaga 2, la primera petición saldría
	// con el token sobrante y la operación ocuparía una vez menos, así que esta
	// cota no tiene holgura: es exacta y la fija el limitador, no el reloj de la
	// máquina.
	assert.GreaterOrEqual(t, time.Since(comienzo), peticiones*intervalo,
		"tres peticiones tras el robots.txt ocupan al menos tres veces el intervalo: el cubo es de un solo token (FR-019, D8)")

	anotadas := anotador.instantes()
	require.Len(t, anotadas, peticiones+1, "el robots.txt y las tres peticiones")

	// Lo que el limitador espacia con exactitud es el turno; la llegada añade a
	// cada turno lo que tarde en despacharse, y ese añadido no es igual en
	// todas: la primera —la del robots.txt— paga además la apertura de la
	// conexión, que las siguientes reutilizan, y cualquiera puede pagar un
	// despertar tardío de la máquina. Dos llegadas consecutivas pueden por eso
	// acercarse por debajo del intervalo sin que el ritmo haya fallado. Lo que
	// sí es exacto es que la llegada i no se adelanta a su turno, que va i
	// veces el intervalo por detrás del comienzo: la cota la fija el limitador,
	// no el reloj de la máquina, y no lleva holgura. Sin limitador o con ráfaga
	// 2, alguna llegada se adelantaría a su turno.
	for i, llegada := range anotadas {
		assert.GreaterOrEqual(t, llegada.Sub(comienzo), time.Duration(i)*intervalo,
			"la llegada %d al sitio no se adelanta a su turno, a %d × intervalo del comienzo (SC-004)", i, i)
	}
}

// TestRitmoNoRetrasaOtroSitio es la otra mitad de SC-004: el cupo es de cada
// sitio, así que una petición a un sitio distinto no espera por la anterior. Los
// dos servidores locales comparten host (127.0.0.1) y solo se distinguen por el
// puerto, que es precisamente lo que la definición de sitio convierte en dos
// sitios distintos (FR-019, data-model.md §4).
func TestRitmoNoRetrasaOtroSitio(t *testing.T) {
	t.Parallel()

	// Un intervalo muy por encima del margen: si los dos servidores
	// compartieran cupo, la operación contra el segundo costaría un intervalo
	// más, y ninguna lentitud de la máquina puede confundirse con eso.
	const intervalo = 3 * time.Second

	const margen = time.Second

	uno, _ := servidorIdentificado(t, redireccionesDePrueba())
	otro, contadorDelOtro := servidorIdentificado(t, redireccionesDePrueba())

	direccionDeUno := direccionDePrueba(t, uno.URL)
	direccionDelOtro := direccionDePrueba(t, otro.URL)

	require.Equal(t, direccionDeUno.Hostname(), direccionDelOtro.Hostname(),
		"los dos servidores del test están en el mismo host")
	require.NotEqual(t, claveDeSitio(direccionDeUno), claveDeSitio(direccionDelOtro),
		"y aun así son dos sitios, porque el puerto entra en la clave (SC-004)")

	cliente := clienteDePrueba(t, ConIntervalo(intervalo))

	_, err := cliente.Pedir(t.Context(), schema.Contexto{},
		Peticion{Metodo: http.MethodGet, URL: uno.URL + "/norma"})
	require.NoError(t, err, "la operación contra el primer sitio deja su cupo agotado")

	comienzo := time.Now()

	_, err = cliente.Pedir(t.Context(), schema.Contexto{},
		Peticion{Metodo: http.MethodGet, URL: otro.URL + "/norma"})
	require.NoError(t, err)

	// Contra el otro sitio el cliente hace dos peticiones: la de su robots.txt,
	// que ocupa el turno recién nacido de su cubo, y la del recurso, que espera
	// un intervalo. Si los dos sitios compartieran cupo, ya la del robots.txt
	// tendría que esperar el turno que el primero acaba de consumir, y la
	// operación entera costaría el doble de lo que separa a dos peticiones.
	assert.Less(t, time.Since(comienzo), intervalo+margen,
		"el otro sitio tiene su propio cupo y no espera por el turno del primero (SC-004)")
	assert.Equal(t, int64(1), contadorDelOtro.total.Load())
}

// TestRitmoRespetaElContexto comprueba FR-022 en sus dos formas: mientras se
// espera turno no hay ninguna petición emitida, así que el contexto que termina
// durante esa espera —porque se cancela, o porque el plazo que queda ni siquiera
// da para el turno— deja la operación sin petición y con la clase «fuente no
// disponible» (tabla de clases §3, fila 2).
func TestRitmoRespetaElContexto(t *testing.T) {
	t.Parallel()

	// El turno del segundo intento está diez segundos por delante: si la espera
	// no terminara con el contexto, el fallo sería una prueba que no termina y
	// no un margen apurado.
	const intervalo = 10 * time.Second

	casos := []struct {
		nombre   string
		contexto func(context.Context) (context.Context, context.CancelFunc)
		mencion  string
		causa    error
	}{
		{
			nombre: "la cancelación durante la espera impide emitir la petición",
			contexto: func(padre context.Context) (context.Context, context.CancelFunc) {
				ctx, cancelar := context.WithCancel(padre)
				time.AfterFunc(50*time.Millisecond, cancelar)

				return ctx, cancelar
			},
			mencion: "esperando turno",
			causa:   context.Canceled,
		},
		{
			nombre: "un plazo más corto que el turno ni lo espera ni emite la petición",
			contexto: func(padre context.Context) (context.Context, context.CancelFunc) {
				return context.WithTimeout(padre, time.Second)
			},
			mencion: "no llega dentro del plazo",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			servidor, contador := servidorIdentificado(t, redireccionesDePrueba())
			cliente := clienteDePrueba(t, ConIntervalo(intervalo))
			norma := Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"}

			_, err := cliente.Pedir(t.Context(), schema.Contexto{}, norma)
			require.NoError(t, err, "la primera petición consume el turno del sitio")

			ctx, cancelar := caso.contexto(t.Context())
			defer cancelar()

			comienzo := time.Now()
			respuesta, err := cliente.Pedir(ctx, schema.Contexto{}, norma)

			fallo := falloDe(t, err)
			assert.Equal(t, schema.ClaseFuenteNoDisponible, fallo.Clase(),
				"quedarse sin turno es «la fuente no sabe entregar el recurso» (FR-022)")
			assert.Contains(t, fallo.Error(), caso.mencion, "el mensaje dice qué terminó la espera (FR-034)")
			assert.Contains(t, fallo.Error(), claveDeSitio(direccionDePrueba(t, servidor.URL)),
				"y en qué sitio se esperaba (FR-034)")
			assert.Equal(t, Respuesta{}, respuesta, "sin turno no hay respuesta que entregar")
			assert.Equal(t, int64(1), contador.total.Load(),
				"la petición no se emite: el servidor solo ha visto la que consumió el turno (FR-022)")
			assert.Less(t, time.Since(comienzo), intervalo,
				"la espera termina con el contexto, no con el turno")

			if caso.causa != nil {
				require.ErrorIs(t, err, caso.causa, "la causa del contexto llega intacta hasta quien llama")
			}
		})
	}
}

// llegadas anota el instante en que cada petición llega al servidor, que es
// donde SC-004 mide la separación: lo que el criterio dice es cómo las ve la
// fuente, no lo que tarde quien llama.
type llegadas struct {
	mu       sync.Mutex
	anotadas []time.Time
}

// anota envuelve el manejador del servidor de prueba para quedarse con el
// instante de cada llegada antes de atenderla.
func (l *llegadas) anota(manejador http.HandlerFunc) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		l.mu.Lock()
		l.anotadas = append(l.anotadas, time.Now())
		l.mu.Unlock()

		manejador(escritor, peticion)
	}
}

// instantes devuelve una copia de lo anotado hasta ahora, de modo que el test
// lea los instantes sin compartir la lista con el servidor.
func (l *llegadas) instantes() []time.Time {
	l.mu.Lock()
	defer l.mu.Unlock()

	return slices.Clone(l.anotadas)
}
