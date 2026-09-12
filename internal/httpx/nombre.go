package httpx

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
)

// extensionDeGrabacion es la de todo fichero de grabación: son objetos JSON y
// se leen en una revisión humana (contrato de grabación §1).
const extensionDeGrabacion = ".json"

const (
	// largoMaximoDelNombre es hasta dónde puede crecer el nombre —sin la
	// extensión— antes de que haya que recortarlo. No es un capricho: hay
	// sistemas de ficheros que no admiten un tramo de ruta de más de 255 bytes, y
	// la de una grabación lleva por delante la raíz y el nombre de la fuente. El
	// tope está bastante por debajo para que quepa esa cabecera y siga siendo un
	// nombre que alguien pueda leer de un vistazo.
	largoMaximoDelNombre = 120
	// largoDelPrefijoRecortado es lo que se conserva del nombre cuando hay que
	// recortarlo: bastante para reconocer la petición de un vistazo, y con
	// veinte caracteres de margen hasta el tope para el resumen y su guion.
	largoDelPrefijoRecortado = 100
	// digitosDelResumen son los del resumen que desempata dos peticiones cuyos
	// cien primeros caracteres coinciden. Ocho dígitos hexadecimales bastan: el
	// nombre no es único por construcción de todos modos —la colisión se resuelve
	// por contenido (contrato de grabación §3 y §4)— y esto solo evita que dos
	// rutas largas y parecidas se pisen a diario.
	digitosDelResumen = 8
)

// nombreDeGrabacion es el nombre del fichero donde se graba una petición y
// donde la reproducción la busca después. Se deriva del método y de la
// dirección completa —esquema, sitio, ruta y consulta— y de nada más, porque
// nadie fuera del paquete lo declara: buena parte de las peticiones que se
// graban las origina el propio cliente —robots.txt, redirecciones, reintentos—
// y no el adaptador que pidió el recurso (FR-039, research.md D12).
//
// El resultado es determinista —la misma petición da el mismo nombre en
// cualquier máquina—, legible en un diff y portable: solo lleva caracteres de
// [A-Za-z0-9._-], así que no hay ninguno de los que Windows prohíbe ni espacios.
// No es único: dos direcciones distintas pueden sanearse al mismo nombre, y por
// eso tanto la grabación como la reproducción comparan además la petición
// guardada dentro del fichero (contrato de grabación §2, §3 y §4).
func nombreDeGrabacion(metodo string, direccion *url.URL) string {
	nombre := saneado(partesDelNombre(metodo, direccion))

	// El nombre saneado es ASCII puro, así que contarlo y cortarlo por bytes es
	// contarlo y cortarlo por caracteres.
	if len(nombre) > largoMaximoDelNombre {
		nombre = nombre[:largoDelPrefijoRecortado] + "-" + resumenDeLaPeticion(metodo, direccion)
	}

	return nombre + extensionDeGrabacion
}

// partesDelNombre encadena las partes de la petición que entran en el nombre,
// con la forma del contrato §2: <MÉTODO>_<esquema>_<host>[-<puerto>]<ruta>[_q_<consulta>].
//
// La ruta y la consulta se escriben **como vienen**; que la barra de la ruta y
// el `=` y el `&` de la consulta acaben en guion bajo no se hace aquí, sino en
// el saneado, porque los tres están fuera del conjunto admitido y esa es
// exactamente su regla. Escribirlo dos veces daría el mismo nombre y una regla
// más que mantener.
//
// La ruta va sin descodificar de más ni de menos: se toma url.URL.Path, la
// forma descodificada, para que un espacio de la ruta acabe en el mismo guion
// bajo que cualquier otro carácter fuera del conjunto y no en un `%20` que no
// hay quien lea.
func partesDelNombre(metodo string, direccion *url.URL) string {
	var partes strings.Builder

	partes.WriteString(metodo)
	partes.WriteString("_")
	partes.WriteString(direccion.Scheme)
	partes.WriteString("_")
	// url.Parse ya devuelve el esquema en minúsculas; el host no, así que se
	// normaliza aquí, como en la clave de sitio.
	partes.WriteString(strings.ToLower(direccion.Hostname()))

	// Solo el puerto que la dirección declara: el que el esquema implica no
	// aparece, para que http://fuente/x y http://fuente:80/x se graben en el
	// mismo fichero, que es lo que la clave de sitio ya dice que son.
	if puerto := direccion.Port(); puerto != "" {
		partes.WriteString("-")
		partes.WriteString(puerto)
	}

	partes.WriteString(direccion.Path)

	if consulta := direccion.RawQuery; consulta != "" {
		partes.WriteString("_q_")
		partes.WriteString(consulta)
	}

	return partes.String()
}

// saneado deja el nombre en el conjunto portable del contrato §2: todo carácter
// fuera de [A-Za-z0-9._-] se vuelve un guion bajo, las secuencias de guiones
// bajos —los que ya venían y los que acaban de aparecer— se colapsan en uno, y
// los de los extremos se recortan. De ahí sale que una ruta vacía o `/` no
// aporte nada al nombre.
func saneado(partes string) string {
	var nombre strings.Builder

	nombre.Grow(len(partes))

	separadorPendiente := false

	for _, caracter := range partes {
		if esSeparadorDelNombre(caracter) {
			separadorPendiente = nombre.Len() > 0

			continue
		}

		if separadorPendiente {
			nombre.WriteByte('_')

			separadorPendiente = false
		}

		nombre.WriteRune(caracter)
	}

	return nombre.String()
}

// esSeparadorDelNombre dice si un carácter se escribe como guion bajo: o lo es
// ya, o está fuera del conjunto admitido. Los dos casos van juntos porque
// colapsan juntos: `/a,b`, `/a_b` y `/a__b` dan el mismo nombre.
func esSeparadorDelNombre(caracter rune) bool {
	admitido := (caracter >= 'a' && caracter <= 'z') ||
		(caracter >= 'A' && caracter <= 'Z') ||
		(caracter >= '0' && caracter <= '9') ||
		caracter == '.' || caracter == '-'

	return !admitido
}

// resumenDeLaPeticion son los primeros dígitos hexadecimales del resumen de la
// petición entera, con la forma «<MÉTODO> <dirección>». Va detrás del nombre
// recortado para que dos peticiones largas que comparten los cien primeros
// caracteres —dos anexos del mismo expediente, por ejemplo— no acaben en el
// mismo fichero solo por ser largas.
//
// El resumen no protege nada: solo desempata nombres. Se toma de la dirección
// tal como se emitió, que es la misma cadena que el fichero guarda dentro.
func resumenDeLaPeticion(metodo string, direccion *url.URL) string {
	resumen := sha256.Sum256([]byte(metodo + " " + direccion.String()))

	return hex.EncodeToString(resumen[:])[:digitosDelResumen]
}
