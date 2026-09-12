package httpx

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"net/http"
	"time"
)

// Ley del retardo entre dos intentos (D9, contrato §7). La base y el factor no
// se configuran: lo único que se declara es cuántos intentos se hacen, porque
// un retardo más corto que este es ir más deprisa de lo que una fuente pública
// que ya está fallando puede soportar.
const (
	// esperaBase es lo que se espera antes del primer reintento, antes de
	// aleatorizar la mitad.
	esperaBase = 500 * time.Millisecond
	// factorDeEspera es lo que crece la base de un intento al siguiente.
	factorDeEspera = 2
	// techoDeEspera es donde deja de crecer: media hora de reintentos no ayuda
	// a nadie, y por encima de esto lo que hay es una fuente caída, no un
	// tropiezo pasajero.
	techoDeEspera = 30 * time.Second
)

// relojDeEspera es cómo el decorador espera entre dos intentos: devuelve nil
// cuando la espera terminó y el error del contexto cuando lo que terminó fue la
// operación. Es un campo y no una llamada directa a time para que los tests del
// propio paquete comprueben la política sobre las duraciones pedidas en vez de
// esperarlas de verdad (D9).
type relojDeEspera func(ctx context.Context, espera time.Duration) error

// decoradorDeReintentos es el escalón que repite lo que puede ser pasajero: un
// error del servidor o un fallo de transporte. Va **por encima** del ritmo, de
// modo que cada reintento espere su turno en el sitio igual que la primera
// petición —un reintento es una petición más, y ponerlo por debajo del
// limitador dejaría los reintentos fuera del ritmo (FR-021, D3)—, y por debajo
// de la identificación, de modo que lo que se repite ya viene identificado
// (FR-009).
type decoradorDeReintentos struct {
	// intentos es cuántas veces se pide como mucho, la primera incluida. Nunca
	// es menor que uno: la opción lo valida en la construcción.
	intentos int
	// dormir es la espera entre dos intentos, interrumpible por el contexto.
	dormir relojDeEspera
	// siguiente es el escalón al que se entrega cada intento.
	siguiente http.RoundTripper
}

// conReintentos envuelve el escalón que recibe con el decorador, con el número
// de intentos y el reloj del cliente.
func conReintentos(siguiente http.RoundTripper, intentos int, dormir relojDeEspera) http.RoundTripper {
	return &decoradorDeReintentos{intentos: intentos, dormir: dormir, siguiente: siguiente}
}

// conReloj sustituye el reloj con el que se espera entre dos intentos. No se
// exporta —ni podría—: fuera del paquete no hay ninguna razón legítima para no
// esperar el retardo, y la única para sustituirlo son los tests del propio
// paquete, que comprueban la política sobre lo que el decorador pide esperar en
// vez de tardar los minutos que esperarlo de verdad costaría (FR-002, D9).
func conReloj(reloj relojDeEspera) Opcion {
	return func(config *configuracionDelCliente) error {
		config.reloj = reloj

		return nil
	}
}

// RoundTrip pide hasta agotar los intentos, esperando entre uno y el siguiente.
// Lo que entrega es siempre el resultado del último intento hecho: el acierto
// en cuanto llega, y si no, lo que la fuente respondió la última vez, que es
// quien clasifica la cadena convierte en la clase que le toca (FR-029).
//
// Cada intento baja con una copia de la petición —«Clone returns a deep copy of
// r with its context changed to ctx» (go doc net/http.Request.Clone)—, que en
// GET y HEAD, sin cuerpo, es una copia completa. El bucle no tiene condición de
// salida propia porque las dos que hay son retornos: el intento que no se
// repite y el que ya era el último.
func (r *decoradorDeReintentos) RoundTrip(peticion *http.Request) (*http.Response, error) {
	ctx := peticion.Context()

	for intento := 1; ; intento++ {
		respuesta, err := r.siguiente.RoundTrip(peticion.Clone(ctx))

		if intento == r.intentos || !seReintenta(ctx, respuesta, err) {
			return respuesta, err
		}

		descartar(respuesta)

		if errDeEspera := r.dormir(ctx, esperaDelIntento(intento)); errDeEspera != nil {
			return nil, esperaInterrumpida(peticion, errDeEspera)
		}
	}
}

// seReintenta es la lista cerrada de FR-025, y la cancelación va delante de
// todo: un contexto terminado no deja hacer ningún intento más, venga el fallo
// de donde venga (FR-026). Lo que se repite son los errores del servidor y los
// fallos de transporte, que pueden ser pasajeros; nunca un 4xx —el 429
// incluido, que tiene clase propia y trae su propia espera (FR-030)— ni nada
// que la fuente haya sabido entregar.
//
// Un escalón de más abajo que ya declaró su clase y su motivo no es un fallo
// de transporte, y por eso tampoco se repite: lo que declara es una decisión ya
// tomada —hoy el turno que no cabe en el plazo que queda (FR-022), mañana la
// ruta que el robots.txt deniega (FR-014)—, que no va a cambiar por volver a
// pedirla y cuyo motivo cierto se perdería al insistir. Es la misma regla con
// la que cliente.go clasifica lo que sube.
func seReintenta(ctx context.Context, respuesta *http.Response, err error) bool {
	if ctx.Err() != nil {
		return false
	}

	if err != nil {
		var declarado *Error

		return !errors.As(err, &declarado)
	}

	return respuesta.StatusCode >= http.StatusInternalServerError
}

// descartar deja sin cuerpo abierto la respuesta de un intento que se va a
// repetir: lo lee entero —para que la conexión vuelva al almacén y el intento
// siguiente la reutilice en vez de abrir otra— y lo cierra (FR-027). El fallo
// de transporte no trae ninguna respuesta que descartar.
//
// Ninguno de los dos resultados se comprueba, y no es un error silenciado: esta
// respuesta ya está descartada, así que un cuerpo que no se deja leer o no se
// deja cerrar no cambia nada de lo que quien llama va a recibir —el intento
// siguiente, o el resultado del último—, y convertirlo en un fallo sustituiría
// el motivo cierto, que es lo que la fuente respondió, por uno accesorio.
func descartar(respuesta *http.Response) {
	if respuesta == nil {
		return
	}

	_, _ = io.Copy(io.Discard, respuesta.Body)
	_ = respuesta.Body.Close()
}

// esperaInterrumpida declara la clase del corte aquí, que es donde se sabe qué
// pasó: el contexto terminó mientras se esperaba para volver a intentarlo, de
// modo que no hay ningún intento más (FR-026) y la operación no ha llegado a la
// fuente (clase «fuente no disponible», FR-029). Quien clasifica el resultado
// de la cadena lo entrega tal cual, igual que hace con el del ritmo.
func esperaInterrumpida(peticion *http.Request, causa error) *Error {
	implicada := Peticion{Metodo: peticion.Method, URL: peticion.URL.String()}

	return errorDeFuenteNoDisponible(implicada, 0, causa,
		"la operación ha terminado esperando para reintentar")
}

// esperaDelIntento es lo que se espera después del intento que se nombra, con
// *equal jitter*: la mitad de la base es fija y la otra mitad, aleatoria. Esa
// forma es la que hace ciertas a la vez las dos cosas que SC-005 pide —que las
// esperas crezcan, porque el suelo de un intento es el techo del anterior, y
// que no sean iguales en dos ejecuciones—, cosa que el *full jitter*, que
// sortea sobre el intervalo entero, no garantiza (D9).
//
// La parte aleatoria sale de crypto/rand y no de math/rand: un jitter no
// necesita criptografía, pero gosec marca con G404 los dos generadores de
// math/rand y este proyecto no admite ninguna supresión. El resultado de Read
// no se comprueba porque la función «never returns an error, and always fills b
// entirely» (crypto/rand/rand.go); ante un fallo del lector del sistema el
// proceso termina, y eso no es una ruta de fallo que este paquete pueda
// interceptar ni cubrir en test. El desplazamiento descarta el bit alto para
// que la conversión a int64 no pueda desbordar.
func esperaDelIntento(intento int) time.Duration {
	// La mitad nunca es cero: la base es fija y no hay opción para cambiarla,
	// así que vale 250 ms como mínimo y el módulo siempre tiene divisor.
	mitad := baseDelIntento(intento) / 2

	var octetos [8]byte

	rand.Read(octetos[:])

	aleatoria := time.Duration(int64(binary.BigEndian.Uint64(octetos[:]) >> 1))

	return mitad + aleatoria%mitad
}

// baseDelIntento es min(esperaBase · factor^(intento−1), techo), calculado
// multiplicando y no con potencias para que el techo corte antes de que la
// multiplicación pueda desbordar.
func baseDelIntento(intento int) time.Duration {
	base := esperaBase

	for range intento - 1 {
		if base > techoDeEspera/factorDeEspera {
			return techoDeEspera
		}

		base *= factorDeEspera
	}

	return base
}

// dormirInterrumpible es el reloj de verdad: espera lo que se le pide, pero no
// más de lo que dure la operación. El vencimiento o la cancelación del contexto
// durante la espera la termina en ese instante, que es lo que impide que un
// retardo de treinta segundos sobreviva al plazo de quien llama (FR-005,
// FR-026).
//
// El temporizador de time.After no necesita cancelarse: desde Go 1.23 el que no
// llega a dispararse lo recoge el recolector en cuanto deja de ser alcanzable
// (go doc time.After).
func dormirInterrumpible(ctx context.Context, espera time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(espera):
		return nil
	}
}
