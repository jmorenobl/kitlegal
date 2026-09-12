package httpx

import (
	"context"
	"net/http"
)

// cabeceraDeIdentificacion es la cabecera por la que quien administra un sitio
// sabe qué le está pidiendo datos (FR-006). Sin ella la biblioteca enviaría la
// suya —«Go-http-client/1.1»—, que es justo lo que el principio I prohíbe.
const cabeceraDeIdentificacion = "User-Agent"

// decoradorDeIdentificacion es el escalón de más arriba de la cadena —el «UA»
// de docs/ROADMAP.md §2—: pone la identificación del proyecto en toda petición
// que baja por ella, sea del recurso pedido, de un salto de una redirección o de
// un reintento (FR-006, SC-001, D3).
//
// Es redundante a propósito, y conviene saberlo: toda petición que llega hasta
// aquí nace ya identificada de nuevaPeticionIdentificada, el único constructor
// del paquete y el único dueño de la cabecera. El decorador no es un segundo
// dueño sino la reafirmación, en la propia cadena, de que nada de lo que la
// atraviesa sale sin ella. Por eso ninguna medida lo echa en falta si se quita
// —lo que los tests detectan es vaciar la cabecera en ponerIdentificacion, que
// los dos comparten— y su sitio en la cadena lo fija un test de forma,
// TestCadenaEmpiezaPorLaIdentificacion (D3, enmienda tras la revisión final).
//
// Va arriba del todo para que la grabación, que está abajo, guarde la petición
// ya identificada (FR-037).
type decoradorDeIdentificacion struct {
	// siguiente es el escalón al que se entrega la petición ya identificada.
	siguiente http.RoundTripper
}

// conIdentificacion envuelve el escalón que recibe con el decorador. New y Replay
// lo ponen arriba del todo de sus cadenas.
func conIdentificacion(siguiente http.RoundTripper) http.RoundTripper {
	return &decoradorDeIdentificacion{siguiente: siguiente}
}

// RoundTrip identifica la petición y la entrega al escalón siguiente. Lo que
// identifica es una copia, porque «RoundTrip should not modify the request»
// (go doc net/http.RoundTripper): la que baja desde arriba no se toca.
func (i *decoradorDeIdentificacion) RoundTrip(peticion *http.Request) (*http.Response, error) {
	identificada := peticion.Clone(peticion.Context())
	ponerIdentificacion(identificada)

	return i.siguiente.RoundTrip(identificada)
}

// nuevaPeticionIdentificada es el único constructor de peticiones del paquete, y
// por eso vive aquí: un decorador solo ve lo que baja desde arriba, así que las
// peticiones que el propio paquete origina por debajo de él —la de robots.txt,
// que fabrica robots.go— no volverían a pasar por el decorador y nacerían sin
// cabecera. Naciendo de aquí, la identificación tiene un solo dueño y no hay
// descuido que la omita (FR-009, D3); TestSoloIdentificarConstruyePeticiones lo
// comprueba sobre el árbol sintáctico del paquete, y
// TestIdentificacionEnTodaPeticion, en ejecución, sobre las cuatro procedencias.
func nuevaPeticionIdentificada(ctx context.Context, metodo, direccion string) (*http.Request, error) {
	peticion, err := http.NewRequestWithContext(ctx, metodo, direccion, nil)
	if err != nil {
		return nil, err
	}

	ponerIdentificacion(peticion)

	return peticion, nil
}

// ponerIdentificacion fija la cabecera con la identificación del proyecto, que
// no se configura ni se puede vaciar desde fuera (FR-008).
func ponerIdentificacion(peticion *http.Request) {
	peticion.Header.Set(cabeceraDeIdentificacion, AgenteDeUsuario())
}
