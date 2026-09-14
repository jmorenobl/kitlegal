package boe

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// codigoDeIdentidadHumana es el código de salida que ningún fallo de la fuente
// puede dar (FR-100).
const codigoDeIdentidadHumana = 6

// TestPedirClasificaEstados fija la puerta de la fuente a la red: cómo pide cada
// recurso —GET, en XML el bloque y en JSON lo demás (FR-003)— y qué hace con lo
// que le entrega el Pedidor (contrato errores-y-codigos, filas 6, 7 y 9 a 15;
// research.md D10):
//
//   - un 2xx es lo obtenido, con el cuerpo y el instante de emisión que declara
//     la respuesta (FR-096);
//   - el 404 de un bloque, un índice, unos metadatos o un análisis es «no
//     encontrado»; el de una búsqueda y cualquier otro estado que no sea 2xx,
//     «fuente no disponible», con el estado en el mensaje;
//   - en ensayo, la descripción de la petición que no se emitió (ADR 0011);
//   - y el error del Pedidor —el de httpx, con su clase— lleva la dirección
//     pedida y el instante que declara (FR-095, FR-101).
//
// Las direcciones van escritas enteras, y no con las funciones de
// direcciones.go, para que un cambio en ellas no pase por aquí en silencio. Los
// estados los entrega un Pedidor de prueba; el ensayo, los errores de httpx y
// los instantes de emisión, el cliente real de httpx en reproducción sobre las
// grabaciones y los sintéticos del paquete, sin abrir ninguna conexión.
func TestPedirClasificaEstados(t *testing.T) {
	t.Parallel()

	const (
		api                           = "https://www.boe.es/datosabiertos/api/legislacion-consolidada"
		urlDelBloque                  = api + "/id/BOE-A-2015-10565/texto/bloque/a21"
		urlDelBloqueInexistente       = api + "/id/BOE-A-2015-10565/texto/bloque/a9999"
		urlDelIndice                  = api + "/id/BOE-A-2015-10565/texto/indice"
		urlDelIndiceInexistente       = api + "/id/BOE-A-2099-99999/texto/indice"
		urlDeLosMetadatos             = api + "/id/BOE-A-2015-10565/metadatos"
		urlDeLosMetadatosInexistentes = api + "/id/BOE-A-2099-99999/metadatos"
		urlDelAnalisis                = api + "/id/BOE-A-2015-10565/analisis"
		urlDelAnalisisInexistente     = api + "/id/BOE-A-2099-99999/analisis"
		urlDeLaBusqueda               = api + "?limit=10&query=%7B%22query%22%3A%20%7B%22query_string%22%3A%20" +
			"%7B%22query%22%3A%20%22titulo%3Aprocedimiento%20AND%20titulo%3Aadministrativo%20AND%20" +
			"titulo%3Acom%5Cu00fan%22%7D%7D%7D"
		norma                  = "BOE-A-2015-10565"
		normaInexistente       = "BOE-A-2099-99999"
		carpetaDelLimite       = "testdata/sintetico/limite/" + NombreDeLaFuente
		carpetaDeLaFuenteCaida = "testdata/sintetico/fuente-caida/" + NombreDeLaFuente
	)

	instanteDeLaRespuesta := time.Date(2026, time.September, 13, 10, 30, 0, 123456789, time.UTC)

	t.Run("2xx: el cuerpo y el instante de la respuesta, pedida con GET y el formato de su recurso", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre   string
			pedido   pedido
			estado   int
			peticion httpx.Peticion
		}{
			{
				nombre:   "200 del bloque, en XML",
				pedido:   pedidoDelBloque(norma, "a21"),
				estado:   200,
				peticion: httpx.Peticion{Metodo: "GET", URL: urlDelBloque, Acepta: "application/xml"},
			},
			{
				nombre:   "200 del índice, en JSON",
				pedido:   pedidoDelIndice(norma),
				estado:   200,
				peticion: httpx.Peticion{Metodo: "GET", URL: urlDelIndice, Acepta: "application/json"},
			},
			{
				nombre:   "200 de los metadatos, en JSON",
				pedido:   pedidoDeLosMetadatos(norma),
				estado:   200,
				peticion: httpx.Peticion{Metodo: "GET", URL: urlDeLosMetadatos, Acepta: "application/json"},
			},
			{
				nombre:   "200 del análisis, en JSON",
				pedido:   pedidoDelAnalisis(norma),
				estado:   200,
				peticion: httpx.Peticion{Metodo: "GET", URL: urlDelAnalisis, Acepta: "application/json"},
			},
			{
				nombre:   "200 de la búsqueda, en JSON y con la dirección tal como llega",
				pedido:   pedidoDeLaBusqueda(urlDeLaBusqueda),
				estado:   200,
				peticion: httpx.Peticion{Metodo: "GET", URL: urlDeLaBusqueda, Acepta: "application/json"},
			},
			{
				nombre:   "203, otro 2xx",
				pedido:   pedidoDeLosMetadatos(norma),
				estado:   203,
				peticion: httpx.Peticion{Metodo: "GET", URL: urlDeLosMetadatos, Acepta: "application/json"},
			},
			{
				nombre:   "299, el último 2xx",
				pedido:   pedidoDelBloque(norma, "a21"),
				estado:   299,
				peticion: httpx.Peticion{Metodo: "GET", URL: urlDelBloque, Acepta: "application/xml"},
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				cuerpo := []byte("el cuerpo de " + caso.nombre)
				pedidor := &pedidorDePrueba{responde: respondeCon(caso.estado, cuerpo, instanteDeLaRespuesta)}
				// El contexto de ejecución llega entero al Pedidor, que es quien
				// honra --dry-run.
				ejecucion := schema.Contexto{Timeout: time.Minute, Asunto: "asunto-de-prueba"}

				recibido, err := pedir(t.Context(), pedidor, ejecucion, caso.pedido)

				require.NoError(t, err)
				assert.Equal(t, obtenido{cuerpo: cuerpo, instante: instanteDeLaRespuesta}, recibido)
				assert.Equal(t, []httpx.Peticion{caso.peticion}, pedidor.peticiones)
				assert.Equal(t, []schema.Contexto{ejecucion}, pedidor.ejecuciones)
			})
		}
	})

	t.Run("estados que no son 2xx: no encontrado solo el 404 de lo que puede no existir", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre   string
			pedido   pedido
			estado   int
			esperado falloEsperado
		}{
			{
				nombre: "404 del bloque: no encontrado, con la norma y el bloque",
				pedido: pedidoDelBloque(norma, "a9999"),
				estado: 404,
				esperado: falloEsperado{
					direccion: urlDelBloqueInexistente,
					clase:     schema.ClaseNoEncontrado,
					mensaje:   "la norma BOE-A-2015-10565 no tiene el bloque a9999 (" + urlDelBloqueInexistente + ")",
				},
			},
			{
				nombre: "404 del índice: no encontrado, con la norma y el recurso",
				pedido: pedidoDelIndice(normaInexistente),
				estado: 404,
				esperado: falloEsperado{
					direccion: urlDelIndiceInexistente,
					clase:     schema.ClaseNoEncontrado,
					mensaje:   "la norma BOE-A-2099-99999 no tiene índice (" + urlDelIndiceInexistente + ")",
				},
			},
			{
				nombre: "404 de los metadatos: no encontrado, con la norma y el recurso",
				pedido: pedidoDeLosMetadatos(normaInexistente),
				estado: 404,
				esperado: falloEsperado{
					direccion: urlDeLosMetadatosInexistentes,
					clase:     schema.ClaseNoEncontrado,
					mensaje:   "la norma BOE-A-2099-99999 no tiene metadatos (" + urlDeLosMetadatosInexistentes + ")",
				},
			},
			{
				nombre: "404 del análisis: no encontrado, con la norma y el recurso",
				pedido: pedidoDelAnalisis(normaInexistente),
				estado: 404,
				esperado: falloEsperado{
					direccion: urlDelAnalisisInexistente,
					clase:     schema.ClaseNoEncontrado,
					mensaje:   "la norma BOE-A-2099-99999 no tiene análisis (" + urlDelAnalisisInexistente + ")",
				},
			},
			{
				nombre: "404 en búsqueda: fuente no disponible, porque una búsqueda nunca es no encontrado",
				pedido: pedidoDeLaBusqueda(urlDeLaBusqueda),
				estado: 404,
				esperado: falloEsperado{
					direccion: urlDeLaBusqueda,
					clase:     schema.ClaseFuenteNoDisponible,
					mensaje:   "la fuente ha respondido con el estado 404 a la petición de la búsqueda (" + urlDeLaBusqueda + ")",
				},
			},
			{
				nombre: "400: fuente no disponible, con el estado y la dirección",
				pedido: pedidoDeLosMetadatos(norma),
				estado: 400,
				esperado: falloEsperado{
					direccion: urlDeLosMetadatos,
					clase:     schema.ClaseFuenteNoDisponible,
					mensaje: "la fuente ha respondido con el estado 400 a la petición de los metadatos " +
						"de la norma BOE-A-2015-10565 (" + urlDeLosMetadatos + ")",
				},
			},
			{
				nombre: "403: fuente no disponible, y no límite o TOS",
				pedido: pedidoDelBloque(norma, "a21"),
				estado: 403,
				esperado: falloEsperado{
					direccion: urlDelBloque,
					clase:     schema.ClaseFuenteNoDisponible,
					mensaje: "la fuente ha respondido con el estado 403 a la petición del bloque a21 " +
						"de la norma BOE-A-2015-10565 (" + urlDelBloque + ")",
				},
			},
			{
				nombre: "410 de lo que puede no existir: fuente no disponible, porque solo el 404 es no encontrado",
				pedido: pedidoDelIndice(norma),
				estado: 410,
				esperado: falloEsperado{
					direccion: urlDelIndice,
					clase:     schema.ClaseFuenteNoDisponible,
					mensaje: "la fuente ha respondido con el estado 410 a la petición del índice " +
						"de la norma BOE-A-2015-10565 (" + urlDelIndice + ")",
				},
			},
			{
				nombre: "300, el primero detrás de los 2xx: fuente no disponible",
				pedido: pedidoDelAnalisis(norma),
				estado: 300,
				esperado: falloEsperado{
					direccion: urlDelAnalisis,
					clase:     schema.ClaseFuenteNoDisponible,
					mensaje: "la fuente ha respondido con el estado 300 a la petición del análisis " +
						"de la norma BOE-A-2015-10565 (" + urlDelAnalisis + ")",
				},
			},
			{
				nombre: "199, el último delante de los 2xx: fuente no disponible",
				pedido: pedidoDeLosMetadatos(norma),
				estado: 199,
				esperado: falloEsperado{
					direccion: urlDeLosMetadatos,
					clase:     schema.ClaseFuenteNoDisponible,
					mensaje: "la fuente ha respondido con el estado 199 a la petición de los metadatos " +
						"de la norma BOE-A-2015-10565 (" + urlDeLosMetadatos + ")",
				},
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				pedidor := &pedidorDePrueba{responde: respondeCon(caso.estado, []byte(`{"data": []}`), instanteDeLaRespuesta)}

				recibido, err := pedir(t.Context(), pedidor, schema.Contexto{}, caso.pedido)

				// El sobre de fallo se fecha como el de éxito: con el instante de
				// emisión de la respuesta que trae el estado (FR-096).
				esperado := caso.esperado
				esperado.instante = instanteDeLaRespuesta
				compruebaFalloDePedir(t, recibido, err, esperado)
				assert.Len(t, pedidor.peticiones, 1)
			})
		}
	})

	t.Run("ensayo: la descripción de la petición, sin emitirla ni mirar su estado", func(t *testing.T) {
		t.Parallel()

		// El cliente real sobre un directorio sin ninguna grabación: si la
		// petición llegara a emitirse, la reproducción fallaría por la grabación
		// ausente.
		cliente := clienteDeReproduccion(t, t.TempDir(), relojQueAvanza(instanteDeLaRespuesta))
		ejecucion := schema.Contexto{DryRun: true}

		casos := []struct {
			nombre string
			pedido pedido
			ensayo string
		}{
			{nombre: "el bloque", pedido: pedidoDelBloque(norma, "a21"), ensayo: "GET " + urlDelBloque},
			{nombre: "los metadatos", pedido: pedidoDeLosMetadatos(norma), ensayo: "GET " + urlDeLosMetadatos},
			{nombre: "la búsqueda", pedido: pedidoDeLaBusqueda(urlDeLaBusqueda), ensayo: "GET " + urlDeLaBusqueda},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				pedidor := &pedidorDePrueba{responde: cliente.Pedir}

				recibido, err := pedir(t.Context(), pedidor, ejecucion, caso.pedido)

				require.NoError(t, err)
				assert.Equal(t, obtenido{ensayo: caso.ensayo}, recibido)
				assert.Equal(t, []schema.Contexto{ejecucion}, pedidor.ejecuciones)
				require.Len(t, pedidor.respuestas, 1)
				assert.True(t, pedidor.respuestas[0].Ensayo)
				assert.Zero(t, pedidor.respuestas[0].Estado)
			})
		}
	})

	t.Run("error de httpx: su clase, la dirección pedida y su instante", func(t *testing.T) {
		t.Parallel()

		instanteDeEmision := time.Date(2026, time.September, 13, 11, 54, 35, 987654321, time.UTC)
		horaDeEmision := func() time.Time { return instanteDeEmision }
		limite := clienteDeReproduccion(t, carpetaDelLimite, horaDeEmision)
		caida := clienteDeReproduccion(t, carpetaDeLaFuenteCaida, horaDeEmision)

		casos := []struct {
			nombre    string
			pedido    pedido
			responde  respuestaDePrueba
			direccion string
			clase     schema.Clase
			motivo    string
		}{
			{
				nombre:    "429: límite o TOS",
				pedido:    pedidoDeLosMetadatos(norma),
				responde:  limite.Pedir,
				direccion: urlDeLosMetadatos,
				clase:     schema.ClaseLimiteOTos,
				motivo:    "ha fallado la petición de los metadatos de la norma BOE-A-2015-10565",
			},
			{
				nombre:    "503: fuente no disponible",
				pedido:    pedidoDelIndice(norma),
				responde:  caida.Pedir,
				direccion: urlDelIndice,
				clase:     schema.ClaseFuenteNoDisponible,
				motivo:    "ha fallado la petición del índice de la norma BOE-A-2015-10565",
			},
			{
				nombre: "429 envuelto con %w: la clase sigue siendo la de httpx",
				pedido: pedidoDeLosMetadatos(norma),
				responde: func(ctx context.Context, ec schema.Contexto, p httpx.Peticion) (httpx.Respuesta, error) {
					respuesta, err := limite.Pedir(ctx, ec, p)
					if err != nil {
						return respuesta, fmt.Errorf("el Pedidor de prueba lo envuelve: %w", err)
					}

					return respuesta, nil
				},
				direccion: urlDeLosMetadatos,
				clase:     schema.ClaseLimiteOTos,
				motivo:    "ha fallado la petición de los metadatos de la norma BOE-A-2015-10565",
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				pedidor := &pedidorDePrueba{responde: caso.responde}

				recibido, err := pedir(t.Context(), pedidor, schema.Contexto{}, caso.pedido)

				// El error que entregó el Pedidor es el de httpx, con la petición
				// que la fuente construyó y el instante de su reloj.
				require.Len(t, pedidor.errores, 1)
				devuelto := pedidor.errores[0]
				var deHTTPX *httpx.Error
				require.ErrorAs(t, devuelto, &deHTTPX)
				assert.Equal(t, caso.direccion, deHTTPX.Peticion.URL)
				assert.Equal(t, instanteDeEmision, deHTTPX.Instante)

				// Y la fuente lo entrega con su clase, la dirección pedida y ese
				// instante, conservando su texto detrás de lo pedido y alcanzable
				// con errors.Is (contrato errores-y-codigos §2).
				compruebaFalloDePedir(t, recibido, err, falloEsperado{
					direccion: caso.direccion,
					instante:  instanteDeEmision,
					clase:     caso.clase,
					mensaje:   caso.motivo + " (" + caso.direccion + "): " + devuelto.Error(),
				})
				require.ErrorIs(t, err, devuelto)
			})
		}
	})

	t.Run("error del Pedidor sin una clase que la fuente pueda declarar: inesperado, sin instante", func(t *testing.T) {
		t.Parallel()

		const motivo = "ha fallado la petición del bloque a21 de la norma BOE-A-2015-10565 (" + urlDelBloque + ")"

		identidadHumana := errorConClaseDePrueba{clase: schema.ClaseIdentidadHumana}
		fueraDelVocabulario := errorConClaseDePrueba{clase: "clase-que-no-existe"}
		inesperado := errorConClaseDePrueba{clase: schema.ClaseInesperado}

		casos := []struct {
			nombre  string
			causa   error
			mensaje string
		}{
			{
				nombre:  "sin clase: el detalle técnico no entra en el mensaje",
				causa:   errors.New("dial tcp 192.0.2.1:443: i/o timeout"),
				mensaje: motivo,
			},
			{
				nombre:  "con la clase de identidad humana, que ninguna fuente pública produce",
				causa:   identidadHumana,
				mensaje: motivo + ": " + identidadHumana.Error(),
			},
			{
				nombre:  "con una clase fuera del vocabulario del dominio",
				causa:   fueraDelVocabulario,
				mensaje: motivo + ": " + fueraDelVocabulario.Error(),
			},
			{
				nombre:  "con la clase inesperado",
				causa:   inesperado,
				mensaje: motivo + ": " + inesperado.Error(),
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				pedidor := &pedidorDePrueba{responde: fallaCon(caso.causa)}

				recibido, err := pedir(t.Context(), pedidor, schema.Contexto{}, pedidoDelBloque(norma, "a21"))

				compruebaFalloDePedir(t, recibido, err, falloEsperado{
					direccion: urlDelBloque,
					clase:     schema.ClaseInesperado,
					mensaje:   caso.mensaje,
				})
				require.ErrorIs(t, err, caso.causa)
			})
		}
	})

	t.Run("instante: el de emisión que declara cada respuesta, no el de la vuelta de pedir", func(t *testing.T) {
		t.Parallel()

		// Un reloj que da un instante distinto en cada llamada, y todos muy
		// anteriores a hoy: si pedir fechara con otro reloj, o con la hora de la
		// vuelta, no coincidiría con el que declara cada respuesta.
		cliente := clienteDeReproduccion(t, carpetaDeLasGrabaciones,
			relojQueAvanza(time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC)))
		pedidor := &pedidorDePrueba{responde: cliente.Pedir}

		delBloque, err := pedir(t.Context(), pedidor, schema.Contexto{}, pedidoDelBloque(norma, "a21"))
		require.NoError(t, err)

		delInexistente, errDelInexistente := pedir(t.Context(), pedidor, schema.Contexto{}, pedidoDelBloque(norma, "a9999"))

		require.Len(t, pedidor.respuestas, 2)
		instanteDelBloque := pedidor.respuestas[0].Instante
		instanteDelInexistente := pedidor.respuestas[1].Instante
		require.False(t, instanteDelBloque.IsZero())
		require.False(t, instanteDelInexistente.IsZero())
		require.NotEqual(t, instanteDelBloque, instanteDelInexistente)

		assert.Equal(t, instanteDelBloque, delBloque.instante)
		assert.NotEmpty(t, delBloque.cuerpo)
		compruebaFalloDePedir(t, delInexistente, errDelInexistente, falloEsperado{
			direccion: urlDelBloqueInexistente,
			instante:  instanteDelInexistente,
			clase:     schema.ClaseNoEncontrado,
			mensaje:   "la norma BOE-A-2015-10565 no tiene el bloque a9999 (" + urlDelBloqueInexistente + ")",
		})
	})
}

// respuestaDePrueba es lo que un Pedidor de prueba hace con cada petición: la
// misma firma que Pedidor.Pedir, de modo que el método de un *httpx.Cliente real
// también lo es.
type respuestaDePrueba func(ctx context.Context, ec schema.Contexto, p httpx.Peticion) (httpx.Respuesta, error)

// pedidorDePrueba es el Pedidor de los tests de la fuente: responde a cada
// petición con su función y anota, en orden, la petición y el contexto de
// ejecución que recibió y la respuesta y el error que entregó. No es seguro para
// usarlo desde varias goroutines a la vez: cada test construye el suyo.
type pedidorDePrueba struct {
	responde    respuestaDePrueba
	peticiones  []httpx.Peticion
	ejecuciones []schema.Contexto
	respuestas  []httpx.Respuesta
	errores     []error
}

// Pedir anota la petición, responde y anota lo que entrega.
func (p *pedidorDePrueba) Pedir(ctx context.Context, ec schema.Contexto, peticion httpx.Peticion) (httpx.Respuesta, error) {
	p.peticiones = append(p.peticiones, peticion)
	p.ejecuciones = append(p.ejecuciones, ec)

	respuesta, err := p.responde(ctx, ec, peticion)

	p.respuestas = append(p.respuestas, respuesta)
	p.errores = append(p.errores, err)

	return respuesta, err
}

// El Pedidor de prueba lo es de verdad, y que lo siga siendo no depende de que
// alguien lo recuerde.
var _ Pedidor = (*pedidorDePrueba)(nil)

// respondeCon es la respuesta que entrega httpx cuando la fuente responde con
// ese estado a la primera y sin redirecciones: la de la petición pedida, con el
// cuerpo y el instante de emisión.
func respondeCon(estado int, cuerpo []byte, instante time.Time) respuestaDePrueba {
	return func(_ context.Context, _ schema.Contexto, peticion httpx.Peticion) (httpx.Respuesta, error) {
		return httpx.Respuesta{
			Peticion: peticion,
			URL:      peticion.URL,
			Estado:   estado,
			Cuerpo:   cuerpo,
			Instante: instante,
		}, nil
	}
}

// fallaCon es el Pedidor que falla siempre con el error.
func fallaCon(err error) respuestaDePrueba {
	return func(context.Context, schema.Contexto, httpx.Peticion) (httpx.Respuesta, error) {
		return httpx.Respuesta{}, err
	}
}

// clienteDeReproduccion es el cliente real de httpx que responde desde las
// grabaciones de la carpeta, con la hora de emisión que da el reloj y sin abrir
// ninguna conexión.
func clienteDeReproduccion(t *testing.T, carpeta string, hora func() time.Time) *httpx.Cliente {
	t.Helper()

	cliente, err := httpx.Replay(carpeta, httpx.ConFuente(NombreDeLaFuente), httpx.ConHora(hora))
	require.NoError(t, err)

	return cliente
}

// relojQueAvanza da un segundo más en cada llamada, empezando en el siguiente
// al inicio, de modo que dos instantes que da nunca coinciden.
func relojQueAvanza(inicio time.Time) func() time.Time {
	var llamadas atomic.Int64

	return func() time.Time {
		return inicio.Add(time.Duration(llamadas.Add(1)) * time.Second)
	}
}

// errorConClaseDePrueba es el error de un Pedidor que no es httpx y que declara
// la clase que se le da, esté o no en el vocabulario del dominio.
type errorConClaseDePrueba struct {
	clase schema.Clase
}

// Error nombra la clase que declara.
func (e errorConClaseDePrueba) Error() string {
	return "el Pedidor de prueba ha fallado con la clase " + string(e.clase)
}

// Clase es la que se le dio.
func (e errorConClaseDePrueba) Clase() schema.Clase {
	return e.clase
}

var _ schema.ConClase = errorConClaseDePrueba{}

// falloEsperado es lo que el applet necesita de un fallo de pedir para montar su
// sobre y su código de salida.
type falloEsperado struct {
	direccion string
	instante  time.Time
	clase     schema.Clase
	mensaje   string
}

// compruebaFalloDePedir exige el fallo de pedir tal como lo recibe el applet:
// nada obtenido, un *Error alcanzable con errors.As con la dirección pedida y el
// instante que declara, la clase que le da el kernel —nunca con el código de
// identidad humana— y el mensaje entero.
func compruebaFalloDePedir(t *testing.T, recibido obtenido, err error, esperado falloEsperado) {
	t.Helper()

	assert.Zero(t, recibido)

	var fallo *Error
	require.ErrorAs(t, err, &fallo)
	assert.Equal(t, esperado.direccion, fallo.URL)
	assert.Equal(t, esperado.instante, fallo.Instante)
	assert.Equal(t, esperado.clase, cli.Clasificar(err))
	assert.NotEqual(t, codigoDeIdentidadHumana, cli.CodigoSalida(err))
	assert.Equal(t, esperado.mensaje, err.Error())
}
