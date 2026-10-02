package httpx

import (
	"context"
	"net/http"
	"slices"
	"sync"
	"sync/atomic"
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

// TestRitmoCompartido fija qué es de todos y qué no cuando más de un cliente
// recibe el mismo Ritmo con ConRitmo, que es como un proceso que construye un
// cliente por llamada sigue pidiendo a cada sitio una sola petición por
// intervalo: los turnos del sitio son de todos, y lo que cada cliente recuerda
// del robots.txt sigue siendo suyo (FR-010 y FR-014 de H21, research D8).
func TestRitmoCompartido(t *testing.T) {
	t.Parallel()

	t.Run("las llegadas de un cliente y de otro no se adelantan a su turno", probarLlegadasConRitmoCompartido)
	t.Run("el robots.txt sigue siendo de cada cliente", probarRobotsDeCadaCliente)
}

// probarLlegadasConRitmoCompartido mide el ritmo como
// TestRitmoSeparaPeticionesDelMismoSitio —en las llegadas al servidor y contra
// el turno de cada una, contado desde un comienzo tomado antes de la primera
// petición—, pero con un cliente más: cada uno pide su robots.txt y su recurso,
// y las cuatro peticiones ocupan cuatro turnos del mismo sitio.
//
// Los dos piden a la vez, que es como llegan las llamadas simultáneas a un
// servidor: con un ritmo por cliente, el robots.txt del segundo llegaría con el
// del primero, sin haber esperado nada, y con cualquier intervalo. Las llegadas
// se anotan en el orden en que el servidor las ve, y la cota no lleva holgura
// porque la fija el limitador: la llegada i no puede adelantarse a i veces el
// intervalo tras el comienzo.
func probarLlegadasConRitmoCompartido(t *testing.T) {
	t.Parallel()

	const intervalo = 150 * time.Millisecond

	anotador := &llegadas{}
	contador := &contadorDeIdentificacion{}
	servidor := servidorLocal(t, contador.vigila(anotador.anota(robotsSinReglas(redireccionesDePrueba()))))
	norma := Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"}

	ritmo := NuevoRitmo(intervalo)
	delProceso := []*Cliente{
		clienteDePrueba(t, ConRitmo(ritmo)),
		clienteDePrueba(t, ConRitmo(ritmo)),
	}

	// Lo que cada goroutine obtiene se comprueba después, en la del test: dentro
	// de las otras no se puede terminar el test.
	estados := make([]int, len(delProceso))
	fallos := make([]error, len(delProceso))
	ctx := t.Context()

	var grupo sync.WaitGroup

	comienzo := time.Now()

	for i, cliente := range delProceso {
		grupo.Go(func() {
			respuesta, err := cliente.Pedir(ctx, schema.Contexto{}, norma)
			estados[i], fallos[i] = respuesta.Estado, err
		})
	}

	grupo.Wait()

	for i := range delProceso {
		require.NoError(t, fallos[i], "el cliente %d lee el recurso cuando le llega el turno", i)
		require.Equal(t, http.StatusOK, estados[i])
	}

	assert.Equal(t, int64(len(delProceso)), contador.robots.Load(),
		"cada cliente pide el robots.txt del sitio: lo que recuerda de él es solo suyo (FR-010)")
	assert.Equal(t, int64(len(delProceso)), contador.total.Load(), "y cada uno pide su recurso")

	anotadas := anotador.instantes()
	require.Len(t, anotadas, 2*len(delProceso), "el robots.txt y el recurso de cada cliente")

	for i, llegada := range anotadas {
		assert.GreaterOrEqual(t, llegada.Sub(comienzo), time.Duration(i)*intervalo,
			"la llegada %d al sitio no se adelanta a su turno, a %d × intervalo del comienzo: "+
				"los dos esperan turno en el mismo Ritmo (FR-014)", i, i)
	}
}

// probarRobotsDeCadaCliente comprueba la otra mitad: compartir los turnos no es
// compartir lo que se sabe del robots.txt. El sitio responde 429 la primera vez
// que se le pide y sus reglas después; el primer cliente se queda con la
// denegación —y no vuelve a pedirlo, ni antes ni después de que el segundo
// lea—, y el segundo, que nace sin saber nada del sitio, lo pide y lee el
// recurso. Es lo que hace que un fallo pasajero al obtenerlo no pase del cliente
// que lo sufre (research D8, V43).
func probarRobotsDeCadaCliente(t *testing.T) {
	t.Parallel()

	var obtenciones atomic.Int64

	servidor, contador := servidorConRobots(t, func(escritor http.ResponseWriter, peticion *http.Request) {
		if obtenciones.Add(1) == 1 {
			robotsConEstado(http.StatusTooManyRequests)(escritor, peticion)

			return
		}

		robotsQueDice("User-agent: *\nAllow: /\n")(escritor, peticion)
	}, redireccionesDePrueba())

	norma := Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"}

	// El intervalo se aparta del camino: aquí no se mide ningún turno.
	ritmo := NuevoRitmo(time.Millisecond)
	primero := clienteDePrueba(t, ConRitmo(ritmo))

	exigeDenegado := func(obtenidas int64) {
		t.Helper()

		respuesta, err := primero.Pedir(t.Context(), schema.Contexto{}, norma)

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseLimiteOTos, fallo.Clase(),
			"el primer cliente no pudo obtener el robots.txt y el sitio le queda denegado (FR-010)")
		assert.Equal(t, http.StatusTooManyRequests, fallo.Estado, "con el estado con que la fuente respondió")
		assert.Equal(t, Respuesta{}, respuesta, "sin permiso no hay respuesta que entregar")
		assert.Equal(t, obtenidas, contador.robots.Load(), "y no vuelve a pedir el robots.txt")
	}

	exigeDenegado(1)
	exigeDenegado(1)

	assert.Zero(t, contador.total.Load(), "sin permiso no se emite ninguna petición del recurso")

	segundo := clienteDePrueba(t, ConRitmo(ritmo))

	respuesta, err := segundo.Pedir(t.Context(), schema.Contexto{}, norma)
	require.NoError(t, err, "el segundo cliente no hereda la denegación del primero (FR-010)")

	assert.Equal(t, http.StatusOK, respuesta.Estado)
	assert.Equal(t, contenidoDePrueba, string(respuesta.Cuerpo), "y lee el recurso")
	assert.Equal(t, int64(2), contador.robots.Load(), "después de pedir él el robots.txt del sitio")
	assert.Equal(t, int64(1), contador.total.Load())

	// Y lo que el segundo leyó tampoco pasa al primero: cada cliente recuerda lo
	// que él obtuvo, mientras vive.
	exigeDenegado(2)
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
