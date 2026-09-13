package httpx

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"
)

// marcaDeEmision es donde la cadena deja el instante en que emitió la petición
// que Pedir está siguiendo. Ese instante solo se conoce dentro de la cadena —los
// reintentos y el ritmo viven por debajo de Pedir— y un http.RoundTripper no
// puede devolver nada más que la respuesta, así que viaja como un valor de
// ámbito de petición en el contexto, que es lo que atraviesa los decoradores sin
// cambiar ninguna firma (research D4 de H4; go doc context.WithValue).
//
// Vacía significa que en el salto en curso no ha salido ninguna petición. Es
// segura para uso concurrente, como todo lo que viaja en un contexto: «Contexts
// are safe for simultaneous use by multiple goroutines» (go doc context).
type marcaDeEmision struct {
	instante atomic.Pointer[time.Time]
}

// anotar deja en la marca el instante de una emisión. Cada intento sobrescribe
// el del anterior, de modo que lo que queda es el del último.
func (m *marcaDeEmision) anotar(instante time.Time) {
	m.instante.Store(&instante)
}

// reiniciar vacía la marca. Pedir lo hace antes de cada salto, para que el que no
// llega a emitirse no se quede con el instante del anterior.
func (m *marcaDeEmision) reiniciar() {
	m.instante.Store(nil)
}

// instanteO devuelve el instante anotado o, con la marca vacía, el que da la hora
// al pedírselo, que es el del abandono (contrato httpx-acepta-e-instante §4 de
// H4).
func (m *marcaDeEmision) instanteO(ahora func() time.Time) time.Time {
	if anotado := m.instante.Load(); anotado != nil {
		return *anotado
	}

	return ahora()
}

// claveDeLaMarca es la clave de la marca en el contexto. Es de un tipo propio y
// sin exportar, que ningún otro paquete puede nombrar ni, por tanto, pisar: «The
// provided key must be comparable and should not be of type string or any other
// built-in type to avoid collisions between packages using context» (go doc
// context.WithValue).
type claveDeLaMarca struct{}

// contextoConMarca devuelve el contexto que lleva la marca en la que la cadena
// anotará el instante de emisión.
func contextoConMarca(ctx context.Context, marca *marcaDeEmision) context.Context {
	return context.WithValue(ctx, claveDeLaMarca{}, marca)
}

// contextoSinMarca oculta la marca que el contexto pudiera llevar, sin tocar su
// plazo ni su cancelación: lo que baja por la cadena con él se emite igual, pero
// no anota nada. Es el que usa el decorador del robots.txt para obtener su
// fichero, que no es la petición pedida y cuya emisión no puede pasar por la de
// ella (contrato httpx-acepta-e-instante §4 de H4).
func contextoSinMarca(ctx context.Context) context.Context {
	return contextoConMarca(ctx, nil)
}

// marcaDelContexto devuelve la marca que lleva el contexto, o nil si no lleva
// ninguna o la lleva oculta.
func marcaDelContexto(ctx context.Context) *marcaDeEmision {
	marca, _ := ctx.Value(claveDeLaMarca{}).(*marcaDeEmision)

	return marca
}

// decoradorDeMarcaDeEmision es el escalón que anota en la marca la hora del
// cliente justo antes de entregar la petición al transporte, de red o de
// reproducción: es el último por el que pasa cada intento, ya identificado, con
// su turno esperado y grabado, así que la hora que anota es la de la petición que
// de verdad sale (contrato httpx-acepta-e-instante §4 de H4).
//
// No decide nada ni espera nada: una petición sin marca en su contexto —la del
// robots.txt, que la lleva oculta— baja igual y no anota nada.
type decoradorDeMarcaDeEmision struct {
	// hora es la del cliente, la que declara ConHora.
	hora func() time.Time
	// siguiente es el transporte.
	siguiente http.RoundTripper
}

// conMarcaDeEmision envuelve el transporte que recibe con el decorador, con la
// hora del cliente. New y Replay lo ponen justo encima de sus transportes.
func conMarcaDeEmision(siguiente http.RoundTripper, hora func() time.Time) http.RoundTripper {
	return &decoradorDeMarcaDeEmision{hora: hora, siguiente: siguiente}
}

// RoundTrip anota la hora en la marca de la petición, si la lleva, y la entrega
// al transporte.
func (d *decoradorDeMarcaDeEmision) RoundTrip(peticion *http.Request) (*http.Response, error) {
	if marca := marcaDelContexto(peticion.Context()); marca != nil {
		marca.anotar(d.hora())
	}

	return d.siguiente.RoundTrip(peticion)
}
