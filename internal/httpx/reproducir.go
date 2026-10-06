package httpx

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
)

// protocoloReproducido es el que declara la respuesta que se sirve desde una
// grabación. El fichero no guarda la versión del protocolo —no forma parte del
// contrato §1— y quien llama no la ve nunca: es solo lo que hace que la
// respuesta esté completa para la biblioteca que la transporta hacia arriba.
const (
	protocoloReproducido     = "HTTP/1.1"
	versionMayorDelProtocolo = 1
	versionMenorDelProtocolo = 1
)

// transporteDeReproduccion es el escalón de más abajo de la cadena de un cliente
// de reproducción y el único que ocupa el sitio del transporte de red: sirve lo
// que la grabación guardó y no abre ninguna conexión bajo ninguna circunstancia,
// porque no tiene con qué abrirla (FR-045, FR-046, D13).
//
// Lee el fichero de cada petición al pedirla, y no el directorio entero al
// construirse: así un fichero roto que ninguna petición usa no rompe un test
// ajeno, y el que sí falla se nombra uno a uno.
type transporteDeReproduccion struct {
	// directorio es el <dir> de Replay, ya validado en la construcción: el
	// directorio de grabaciones de una fuente —el <raíz>/<fuente> que deja la
	// grabación—, donde se busca cada petición.
	directorio string
}

// nuevoTransporteDeReproduccion construye el escalón sobre el directorio que
// Replay validó.
func nuevoTransporteDeReproduccion(directorio string) http.RoundTripper {
	return &transporteDeReproduccion{directorio: directorio}
}

// RoundTrip sirve la grabación de la petición que baja por la cadena, o falla
// con la clase «inesperado» nombrando lo que faltaba: nunca devuelve una
// respuesta vacía, nunca sirve la grabación de otra petición y nunca cae a la
// red (FR-046, FR-047).
func (r *transporteDeReproduccion) RoundTrip(peticion *http.Request) (*http.Response, error) {
	buscada := Peticion{Metodo: peticion.Method, URL: peticion.URL.String()}

	emitida, err := peticionEmitida(peticion, buscada)
	if err != nil {
		return nil, err
	}

	ruta := filepath.Join(r.directorio, nombreDeGrabacion(peticion.Method, peticion.URL, emitida.Cuerpo))

	grabada, err := grabacionDe(ruta, buscada, emitida)
	if err != nil {
		return nil, err
	}

	cuerpo, err := cuerpoGrabado(grabada.Respuesta, ruta, buscada)
	if err != nil {
		return nil, err
	}

	return respuestaReproducida(peticion, grabada.Respuesta, cuerpo), nil
}

// grabacionDe lee la grabación que le toca a la petición y comprueba que es la
// suya: el nombre se deriva del método, de la dirección y del cuerpo (contrato
// §2; contrato httpx-formulario §5 de H23), pero no es único por construcción,
// así que lo que decide es lo que el fichero guarda dentro. El método, la
// dirección y el cuerpo tienen que coincidir **exactamente**; las cabeceras no
// participan, porque la identificación lleva la versión del binario y la haría
// cambiar de una versión a otra (FR-039, FR-047, contrato §4).
//
// Las cuatro formas de no tener grabación que servir son la misma clase y las
// cuatro nombran el fichero: un fixture ausente, ilegible, de otro formato o de
// otra petición no es un fallo de ninguna fuente (FR-063). Un envío cuyos
// campos no están grabados es la primera: su fichero, que lleva el cuerpo en el
// nombre, no está (FR-034 de H23).
func grabacionDe(ruta string, buscada Peticion, emitida peticionGrabada) (*grabacion, error) {
	// filepath.Clean es lo que el control de rutas reconoce como saneado antes
	// de abrir un fichero (gosec G304); la ruta la compone este paquete con el
	// directorio validado en la construcción y un nombre que solo lleva
	// caracteres de [A-Za-z0-9._-], y aquí no se suprime ninguna comprobación.
	limpia := filepath.Clean(ruta)

	contenido, err := os.ReadFile(limpia)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errorInesperado(buscada, err,
			"no hay grabación de "+emitida.descripcion()+": falta el fichero "+ruta)
	}

	if err != nil {
		return nil, errorInesperado(buscada, err, "la grabación "+ruta+" no se ha podido leer")
	}

	var leida grabacion
	if err := json.Unmarshal(contenido, &leida); err != nil {
		return nil, errorInesperado(buscada, err, "la grabación "+ruta+" no se ha podido interpretar")
	}

	if leida.Formato != formatoDeGrabacion {
		return nil, errorInesperado(buscada, nil,
			"la grabación "+ruta+" declara el formato "+strconv.Itoa(leida.Formato)+
				" y solo se puede reproducir el "+strconv.Itoa(formatoDeGrabacion))
	}

	if !leida.Peticion.esLaMisma(emitida) {
		return nil, errorInesperado(buscada, nil,
			"la grabación "+ruta+" guarda "+leida.Peticion.descripcion()+
				" y se buscaba "+emitida.descripcion())
	}

	return &leida, nil
}

// cuerpoGrabado devuelve los bytes del cuerpo en la clave que le toca: texto si
// era UTF-8 válido y base64 si no. Las dos son excluyentes y una de las dos está
// siempre, el cuerpo vacío incluido —que se graba como «"cuerpo": ""»
// (contrato §1)—: con las dos puestas no hay forma de saber cuál vale y sin
// ninguna el fichero no declara ningún cuerpo, y servir en cualquiera de los dos
// casos una respuesta a medias es lo que FR-047 prohíbe.
func cuerpoGrabado(respuesta respuestaGrabada, ruta string, buscada Peticion) ([]byte, error) {
	switch {
	case respuesta.Cuerpo != nil && respuesta.CuerpoBase64 != nil:
		return nil, errorInesperado(buscada, nil,
			"la grabación "+ruta+" declara «cuerpo» y «cuerpo_base64», que son excluyentes")

	case respuesta.Cuerpo != nil:
		return []byte(*respuesta.Cuerpo), nil

	case respuesta.CuerpoBase64 != nil:
		descodificado, err := base64.StdEncoding.DecodeString(*respuesta.CuerpoBase64)
		if err != nil {
			return nil, errorInesperado(buscada, err,
				"el «cuerpo_base64» de la grabación "+ruta+" no es base64 estándar")
		}

		return descodificado, nil

	default:
		return nil, errorInesperado(buscada, nil,
			"la grabación "+ruta+" no declara ningún cuerpo, ni siquiera vacío")
	}
}

// respuestaReproducida construye la respuesta que sube por la cadena con el
// estado, las cabeceras y el cuerpo grabados. Lo que se clasifica arriba es el
// estado que la fuente dio cuando se grabó, de modo que un 5xx o un 429
// guardados producen la misma clase que producirían en red y una redirección
// guardada lleva al salto siguiente (FR-049 d y e).
//
// Las cabeceras se copian a un mapa propio: el de la grabación es de quien la
// leyó, y quien reciba la respuesta no tiene por qué compartir con nadie los
// valores que la biblioteca le deje modificar.
func respuestaReproducida(peticion *http.Request, grabada respuestaGrabada, cuerpo []byte) *http.Response {
	cabeceras := make(http.Header, len(grabada.Cabeceras))
	for nombre, valores := range grabada.Cabeceras {
		cabeceras[http.CanonicalHeaderKey(nombre)] = slices.Clone(valores)
	}

	return &http.Response{
		Status:        strconv.Itoa(grabada.Estado) + " " + http.StatusText(grabada.Estado),
		StatusCode:    grabada.Estado,
		Proto:         protocoloReproducido,
		ProtoMajor:    versionMayorDelProtocolo,
		ProtoMinor:    versionMenorDelProtocolo,
		Header:        cabeceras,
		Body:          io.NopCloser(bytes.NewReader(cuerpo)),
		ContentLength: int64(len(cuerpo)),
		Request:       peticion,
	}
}
