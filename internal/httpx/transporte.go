package httpx

import "net/http"

// nuevoTransporte es el escalón inferior de la cadena: el único que de verdad
// abre una conexión. Es un clon del transporte por omisión de la biblioteca
// —«Clone returns a deep copy of t's exported fields» (go doc
// net/http.Transport.Clone)—, de modo que el cliente hereda lo que la biblioteca
// ya trae bien resuelto —proxy del entorno, HTTP/2, reutilización de conexiones
// ociosas y los plazos de conexión y de saludo TLS— sin compartir con nadie el
// estado del global (D11).
//
// Los dos plazos que hereda no son un plazo de la operación: acotan un intento
// de conexión, mientras que el de la operación lo pone el contexto y solo él
// (FR-004).
func nuevoTransporte() *http.Transport {
	return clonDelTransportePorOmision(http.DefaultTransport)
}

// clonDelTransportePorOmision clona el transporte que recibe cuando es el de la
// biblioteca, que es lo que ocurre siempre en un proceso normal.
//
// Cuando no lo es, no se adopta: http.DefaultTransport es una variable pública y
// cualquier paquete del proceso puede sustituirla, y un transporte ajeno podría
// traer la verificación de certificados desactivada. FR-012 no admite que eso
// dependa de nadie de fuera, así que en ese caso el paquete construye el suyo,
// que sale con la verificación activa porque no toca TLSClientConfig.
func clonDelTransportePorOmision(porOmision http.RoundTripper) *http.Transport {
	if deLaBiblioteca, esDeLaBiblioteca := porOmision.(*http.Transport); esDeLaBiblioteca {
		return deLaBiblioteca.Clone()
	}

	return &http.Transport{}
}

// nuevoClienteHTTP envuelve la cadena ya compuesta en el cliente que la ejecuta.
// Es donde se declaran las tres cosas que la biblioteca haría por su cuenta y
// que este paquete no le deja hacer: poner un plazo, seguir redirecciones y
// guardar cookies.
func nuevoClienteHTTP(cadena http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: cadena,
		// «A Timeout of zero means no timeout» (go doc net/http.Client): el
		// plazo de la operación es el del contexto, ni más corto ni más largo
		// (FR-004).
		Timeout: 0,
		// Las redirecciones las sigue Pedir, que es quien puede someter cada
		// salto a la cadena entera y clasificar el exceso o el bucle con su
		// clase (FR-011, D3, D10).
		CheckRedirect: noSeguirRedirecciones,
		// Sin almacén de cookies: ninguna petición hereda estado de otra.
		Jar: nil,
	}
}

// noSeguirRedirecciones es la política que devuelve la última respuesta en vez
// de seguirla: «if CheckRedirect returns ErrUseLastResponse, then the most
// recent response is returned with its body unclosed, along with a nil error»
// (go doc net/http.Client). Cerrar ese cuerpo es cosa del paquete (D2).
func noSeguirRedirecciones(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}
