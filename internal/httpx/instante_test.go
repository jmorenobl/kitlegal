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

// robotsQueLoDeniegaTodo es el robots.txt de un sitio que no deja pedir ninguna
// ruta a ningún agente.
const robotsQueLoDeniegaTodo = "User-agent: *\nDisallow: /\n"

// TestInstanteDeEmision fija de dónde sale el instante que Pedir entrega en
// Respuesta.Instante y en Error.Instante, que es el que la fuente declarará como
// fecha de la consulta (FR-096, SC-008): el de emisión de la última petición que
// salió hacia la fuente —el último intento del último salto—, el del abandono si
// no llegó a salir ninguna y ninguno en ensayo (contrato httpx-acepta-e-instante
// §2 a §4, research D4 de H4).
//
// La hora de cada subprueba da un instante distinto en cada llamada y el sitio
// anota cuántos llevaba dados al recibir cada petición pedida, de modo que todo
// instante entregado se atribuye a una llamada concreta: la de la marca, que el
// sitio ya ve dada cuando la petición le llega, o la de Pedir al volver, tras la
// que el sitio no recibe nada. Contar las llamadas es además lo que demuestra que
// la obtención del robots.txt no marca nada: si lo hiciera, la hora daría una de
// más.
//
// Ninguna subprueba anida otra, para que la salida detallada tenga exactamente
// una línea por cada situación del inventario de tests; las que comprueban
// varias formas de la misma situación las recorren en secuencia.
func TestInstanteDeEmision(t *testing.T) {
	t.Parallel()

	t.Run("acierto", func(t *testing.T) {
		t.Parallel()

		hora := &horaQueAvanza{}
		servidor, _ := servidorIdentificado(t, hora.anota(redireccionesDePrueba()))

		respuesta, err := clienteConHora(t, hora).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, respuesta.Estado)

		exigeInstante(t, "al primer intento", hora, respuesta.Instante, 1, []int{1})
	})

	t.Run("acierto-tras-reintentos", func(t *testing.T) {
		t.Parallel()

		hora := &horaQueAvanza{}
		servidor, _ := servidorIdentificado(t, hora.anota(servidorQueFalla(2, http.StatusServiceUnavailable)))

		respuesta, err := clienteConHora(t, hora, conReloj(sinEsperar)).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, respuesta.Estado, "el tercer intento acierta")

		exigeInstante(t, "tras dos reintentos, el del último intento", hora, respuesta.Instante, 3, []int{1, 2, 3})
	})

	t.Run("fallo-tras-reintentos", func(t *testing.T) {
		t.Parallel()

		situaciones := []struct {
			nombre    string
			manejador http.HandlerFunc
			espera    relojDeEspera
			clase     schema.Clase
			llegadas  []int
		}{
			{
				nombre:    "agotados los intentos, el del último",
				manejador: servidorQueFalla(intentosPorOmision, http.StatusServiceUnavailable),
				espera:    sinEsperar,
				clase:     schema.ClaseFuenteNoDisponible,
				llegadas:  []int{1, 2, 3},
			},
			{
				nombre:    "un 429, que no se reintenta, el del intento que lo recibió",
				manejador: caidaYLuegoLimitada(),
				espera:    sinEsperar,
				clase:     schema.ClaseLimiteOTos,
				llegadas:  []int{1, 2},
			},
			{
				nombre:    "el plazo que vence esperando para reintentar, el del último intento emitido",
				manejador: servidorQueFalla(intentosPorOmision, http.StatusServiceUnavailable),
				espera:    plazoQueVenceAlEsperar,
				clase:     schema.ClaseFuenteNoDisponible,
				llegadas:  []int{1},
			},
		}

		for _, situacion := range situaciones {
			hora := &horaQueAvanza{}
			servidor, _ := servidorIdentificado(t, hora.anota(situacion.manejador))

			_, err := clienteConHora(t, hora, conReloj(situacion.espera)).Pedir(t.Context(), schema.Contexto{},
				Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})

			fallo := falloDe(t, err)
			assert.Equal(t, situacion.clase, fallo.Clase(), situacion.nombre)
			exigeInstante(t, situacion.nombre, hora, fallo.Instante, len(situacion.llegadas), situacion.llegadas)
		}
	})

	t.Run("denegada-por-robots", func(t *testing.T) {
		t.Parallel()

		// En las dos formas el robots.txt se pide —una vez, o las tres de los
		// reintentos si el sitio no lo entrega— sin que la hora se haya consultado
		// todavía, y la petición pedida no sale: la hora se consulta una sola vez,
		// al abandonar, y no por la obtención ni antes de decidir el permiso.
		situaciones := []struct {
			nombre   string
			robots   http.HandlerFunc
			llegadas []int
		}{
			{nombre: "el robots.txt deniega la ruta", robots: robotsQueDice(robotsQueLoDeniegaTodo), llegadas: []int{0}},
			{
				nombre:   "sin permiso, porque el robots.txt no se ha podido obtener",
				robots:   robotsConEstado(http.StatusServiceUnavailable),
				llegadas: []int{0, 0, 0},
			},
		}

		for _, situacion := range situaciones {
			hora := &horaQueAvanza{}
			servidor, contador := servidorConRobots(t, hora.anota(situacion.robots), hora.anota(redireccionesDePrueba()))

			_, err := clienteConHora(t, hora, conReloj(sinEsperar)).Pedir(t.Context(), schema.Contexto{},
				Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})

			fallo := falloDe(t, err)
			assert.Equal(t, schema.ClaseLimiteOTos, fallo.Clase(), situacion.nombre)
			assert.Zero(t, contador.total.Load(), "%s: la petición pedida no sale", situacion.nombre)
			exigeInstante(t, situacion.nombre, hora, fallo.Instante, 1, situacion.llegadas)
		}
	})

	t.Run("sin-turno", func(t *testing.T) {
		t.Parallel()

		// El robots.txt del sitio ocupa su único turno y el siguiente queda diez
		// segundos por delante; la operación se cancela poco después de servirlo,
		// mientras la petición pedida espera, y esta no llega a salir (FR-022).
		//
		// La cuenta de consultas no basta aquí: una marca puesta antes de esperar
		// turno daría la misma. Lo que las separa es cuándo se consulta la hora, y
		// por eso la de esta subprueba anota además si la operación ya había
		// terminado. La cancelación se arma al servir el robots.txt, y no antes de
		// pedir, para que por lenta que vaya la máquina no pueda llegar antes de
		// que el sitio lo haya recibido.
		const (
			intervalo            = 10 * time.Second
			antesDeLaCancelacion = 100 * time.Millisecond
		)

		ctx, cancelar := context.WithCancel(t.Context())
		defer cancelar()

		hora := &horaQueAvanza{}
		servidor, contador := servidorConRobots(t,
			hora.anota(func(escritor http.ResponseWriter, peticion *http.Request) {
				robotsQueDice("")(escritor, peticion)
				time.AfterFunc(antesDeLaCancelacion, cancelar)
			}),
			hora.anota(redireccionesDePrueba()))

		var terminadaAlConsultar atomic.Bool

		ahora := func() time.Time {
			terminadaAlConsultar.Store(ctx.Err() != nil)

			return hora.ahora()
		}

		_, err := clienteDePrueba(t, ConIntervalo(intervalo), ConHora(ahora)).Pedir(ctx, schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseFuenteNoDisponible, fallo.Clase())
		assert.Contains(t, fallo.Error(), "esperando turno", "lo que la ha detenido es el ritmo del sitio")
		assert.Zero(t, contador.total.Load(), "la petición pedida no sale")
		exigeInstante(t, "sin turno, el del abandono", hora, fallo.Instante, 1, []int{0})
		assert.True(t, terminadaAlConsultar.Load(),
			"la hora se consulta al abandonar, con la operación ya terminada, y no antes de esperar turno")
	})

	t.Run("redireccion", func(t *testing.T) {
		t.Parallel()

		hora := &horaQueAvanza{}
		servidor, _ := servidorIdentificado(t, hora.anota(redireccionesDePrueba()))

		respuesta, err := clienteConHora(t, hora).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/antigua"})
		require.NoError(t, err)
		require.Equal(t, servidor.URL+"/norma", respuesta.URL, "la cadena de la tabla tiene un salto")

		exigeInstante(t, "tras un salto, el del último", hora, respuesta.Instante, 2, []int{1, 2})

		// La marca se reinicia antes de cada salto: el que no llega a salir —el
		// robots.txt de su sitio lo deniega— no hereda el instante del anterior,
		// sino que se fecha al abandonar.
		otraHora := &horaQueAvanza{}
		cerrado, _ := servidorConRobots(t, robotsQueDice(robotsQueLoDeniegaTodo), otraHora.anota(redireccionesDePrueba()))
		origen, _ := servidorIdentificado(t, otraHora.anota(func(escritor http.ResponseWriter, peticion *http.Request) {
			http.Redirect(escritor, peticion, cerrado.URL+"/norma", http.StatusFound)
		}))

		_, err = clienteConHora(t, otraHora).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: origen.URL + "/antigua"})

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseLimiteOTos, fallo.Clase(), "el sitio de destino deniega la ruta")
		exigeInstante(t, "con el segundo salto denegado, el del abandono", otraHora, fallo.Instante, 2, []int{1})
	})

	t.Run("reproduccion", func(t *testing.T) {
		t.Parallel()

		// Con un salto grabado, porque solo así se distingue la marca de la hora
		// al volver: sin sitio que anote las llegadas, lo que las separa es cuántas
		// veces se consulta la hora.
		hora := &horaQueAvanza{}

		respuesta, err := clienteDeReproduccion(t, grabacionesDePrueba(t), ConHora(hora.ahora)).Pedir(
			t.Context(), schema.Contexto{}, Peticion{Metodo: http.MethodGet, URL: antiguaGrabada})
		require.NoError(t, err)
		require.Equal(t, normaGrabada, respuesta.URL, "la grabación de la tabla redirige un salto")

		exigeInstante(t, "al servir la grabación del último salto", hora, respuesta.Instante, 2, nil)
	})

	t.Run("ensayo", func(t *testing.T) {
		t.Parallel()

		hora := &horaQueAvanza{}

		respuesta, err := clienteConHora(t, hora).Pedir(t.Context(), schema.Contexto{DryRun: true}, peticionDeEnsayo)
		require.NoError(t, err)
		require.True(t, respuesta.Ensayo)

		dados, _ := hora.anotado()
		assert.Zero(t, respuesta.Instante, "en ensayo no hubo consulta, así que no hay instante que declarar")
		assert.Empty(t, dados, "y la hora ni siquiera se consulta")
	})

	t.Run("argumentos", func(t *testing.T) {
		t.Parallel()

		hora := &horaQueAvanza{}
		servidor, contador := servidorIdentificado(t, hora.anota(redireccionesDePrueba()))
		cliente := clienteConHora(t, hora)

		malFormadas := []Peticion{
			{Metodo: http.MethodPost, URL: servidor.URL + "/norma"},
			{Metodo: http.MethodGet, URL: "/norma"},
			{Metodo: http.MethodGet, URL: servidor.URL + "/norma", Acepta: "application/"},
		}

		for rechazadas, malFormada := range malFormadas {
			_, err := cliente.Pedir(t.Context(), schema.Contexto{}, malFormada)

			fallo := falloDe(t, err)
			assert.Equal(t, schema.ClaseArgumentos, fallo.Clase())
			exigeInstante(t, "rechazada "+malFormada.Metodo+" "+malFormada.URL+" "+malFormada.Acepta,
				hora, fallo.Instante, rechazadas+1, nil)
		}

		assert.Zero(t, contador.total.Load(), "ninguna petición rechazada abre una conexión")
		assert.Zero(t, contador.robots.Load(), "ni siquiera la del robots.txt del sitio")
	})
}

// horaQueAvanza es la hora de estas tablas: da un instante distinto en cada
// llamada —un minuto después del anterior— y recuerda los que ha dado y cuántos
// llevaba dados al llegar al sitio cada petición que la tabla anota.
type horaQueAvanza struct {
	mu       sync.Mutex
	dados    []time.Time
	llegadas []int
}

// ahora es la función que se declara con ConHora.
func (h *horaQueAvanza) ahora() time.Time {
	h.mu.Lock()
	defer h.mu.Unlock()

	instante := instanteDePrueba.Add(time.Duration(len(h.dados)+1) * time.Minute)
	h.dados = append(h.dados, instante)

	return instante
}

// anota envuelve un manejador para guardar, al llegar cada petición, cuántos
// instantes había dado ya la hora. Las del robots.txt solo pasan por aquí si la
// tabla envuelve también el manejador que las atiende.
func (h *horaQueAvanza) anota(manejador http.HandlerFunc) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		h.mu.Lock()
		h.llegadas = append(h.llegadas, len(h.dados))
		h.mu.Unlock()

		manejador(escritor, peticion)
	}
}

// anotado devuelve una copia de los instantes dados y de las llegadas anotadas.
func (h *horaQueAvanza) anotado() ([]time.Time, []int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	return slices.Clone(h.dados), slices.Clone(h.llegadas)
}

// clienteConHora construye el cliente contra la red de estas tablas, con la hora
// que se le declara y sin la espera del ritmo entre la petición del robots.txt y
// la pedida. Las opciones que se le añaden van detrás y ganan.
func clienteConHora(t *testing.T, hora *horaQueAvanza, opciones ...Opcion) *Cliente {
	t.Helper()

	return clienteDePrueba(t, append([]Opcion{ConIntervalo(time.Millisecond), ConHora(hora.ahora)}, opciones...)...)
}

// exigeInstante atribuye un instante a la llamada de la hora que lo produjo: la
// hora tiene que haberse consultado exactamente las veces declaradas —una por
// petición emitida y, si en el último salto no salió ninguna, una más al
// volver—, cada petición anotada tiene que haber llegado al sitio con los
// instantes declarados ya dados —la pedida, con el suyo; la del robots.txt, sin
// ninguno de más—, y lo entregado tiene que ser el último instante que dio.
func exigeInstante(
	t *testing.T, situacion string, hora *horaQueAvanza, instante time.Time, consultas int, llegadas []int,
) {
	t.Helper()

	dados, anotadas := hora.anotado()

	require.Len(t, dados, consultas,
		"%s: la hora se consulta al emitir cada petición y, si no sale ninguna, al volver; nunca por el robots.txt",
		situacion)
	assert.Equal(t, llegadas, anotadas,
		"%s: la hora se toma justo antes de entregar la pedida al transporte, y nunca por el robots.txt", situacion)
	assert.Equal(t, dados[consultas-1], instante, "%s: y el instante entregado es el último que dio", situacion)
}

// caidaYLuegoLimitada responde 503 a la primera petición, que se reintenta, y
// 429 a las siguientes, que ya no (FR-030).
func caidaYLuegoLimitada() http.HandlerFunc {
	var atendidas atomic.Int64

	return func(escritor http.ResponseWriter, _ *http.Request) {
		if atendidas.Add(1) == 1 {
			escritor.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		escritor.WriteHeader(http.StatusTooManyRequests)
	}
}

// plazoQueVenceAlEsperar es la espera entre dos intentos durante la que vence el
// plazo de la operación: devuelve lo que la de verdad devuelve en ese caso, el
// error del contexto, sin esperar nada.
func plazoQueVenceAlEsperar(context.Context, time.Duration) error {
	return context.DeadlineExceeded
}
