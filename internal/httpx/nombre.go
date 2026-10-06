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

// sinCuerpo es el cuerpo de la petición que no envía nada: toda la que no es el
// envío de un formulario. Con él, el nombre, el resumen y el fichero de una
// grabación son los de antes de que hubiera cuerpos (research D6 de H23).
const sinCuerpo = ""

// nombreDeGrabacion es el nombre del fichero donde se graba una petición y
// donde la reproducción la busca después. Se deriva del método, de la dirección
// completa —esquema, sitio, ruta y consulta— y, si la petición envía un
// formulario, de su cuerpo codificado, y de nada más, porque nadie fuera del
// paquete lo declara: buena parte de las peticiones que se graban las origina
// el propio cliente —robots.txt, redirecciones, reintentos— y no el adaptador
// que pidió el recurso (FR-039, research.md D12). El cuerpo entra porque dos
// envíos a la misma dirección con campos distintos son dos grabaciones (FR-034
// de H23).
//
// El resultado es determinista —la misma petición da el mismo nombre en
// cualquier máquina—, legible en un diff y portable: solo lleva caracteres de
// [A-Za-z0-9._-], así que no hay ninguno de los que Windows prohíbe ni espacios.
// No es único: dos direcciones distintas, o dos cuerpos, pueden sanearse al
// mismo nombre, y por eso tanto la grabación como la reproducción comparan
// además la petición guardada dentro del fichero (contrato de grabación §2, §3
// y §4; contrato httpx-formulario §5 de H23).
func nombreDeGrabacion(metodo string, direccion *url.URL, cuerpo string) string {
	nombre := saneado(partesDelNombre(metodo, direccion, cuerpo))

	// El nombre saneado es ASCII puro, así que contarlo y cortarlo por bytes es
	// contarlo y cortarlo por caracteres.
	if len(nombre) > largoMaximoDelNombre {
		nombre = nombre[:largoDelPrefijoRecortado] + "-" + resumenDeLaPeticion(metodo, direccion, cuerpo)
	}

	return nombre + extensionDeGrabacion
}

// partesDelNombre encadena las partes de la petición que entran en el nombre,
// con la forma del contrato §2 y, detrás, el cuerpo de un envío:
// <MÉTODO>_<esquema>_<host>[-<puerto>]<ruta>[_q_<consulta>][_c_<cuerpo>].
//
// La ruta, la consulta y el cuerpo se escriben **como vienen**; que la barra de
// la ruta y el `=`, el `&` y el `%` de la consulta y del cuerpo acaben en guion
// bajo no se hace aquí, sino en el saneado, porque todos están fuera del
// conjunto admitido y esa es exactamente su regla. Escribirlo dos veces daría el
// mismo nombre y una regla más que mantener.
//
// La ruta va sin descodificar de más ni de menos: se toma url.URL.Path, la
// forma descodificada, para que un espacio de la ruta acabe en el mismo guion
// bajo que cualquier otro carácter fuera del conjunto y no en un `%20` que no
// hay quien lea.
func partesDelNombre(metodo string, direccion *url.URL, cuerpo string) string {
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

	if cuerpo != sinCuerpo {
		partes.WriteString("_c_")
		partes.WriteString(cuerpo)
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
// petición entera, con la forma «<MÉTODO> <dirección>» y, si envía un
// formulario, su cuerpo en la línea siguiente. Va detrás del nombre recortado
// para que dos peticiones largas que comparten los cien primeros caracteres
// —dos anexos del mismo expediente, o dos envíos cuyos campos empiezan igual—
// no acaben en el mismo fichero solo por ser largas.
//
// El resumen no protege nada: solo desempata nombres. Se toma de la dirección y
// del cuerpo tal como se emitieron, que son las mismas cadenas que el fichero
// guarda dentro. El de una petición sin cuerpo es el de siempre: sin él no hay
// línea siguiente.
func resumenDeLaPeticion(metodo string, direccion *url.URL, cuerpo string) string {
	resumida := metodo + " " + direccion.String()

	if cuerpo != sinCuerpo {
		resumida += "\n" + cuerpo
	}

	resumen := sha256.Sum256([]byte(resumida))

	return hex.EncodeToString(resumen[:])[:digitosDelResumen]
}
