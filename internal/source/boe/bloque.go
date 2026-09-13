package boe

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"unicode/utf8"
)

// Los nombres que la lectura del bloque busca en el XML, siempre sin espacio de
// nombres (refs/boe.py 102-118).
const (
	// elementoBloque es el del bloque pedido (102).
	elementoBloque = "bloque"
	// elementoVersion es el de cada versión del bloque (109).
	elementoVersion = "version"
	// atributoTitulo y atributoTipo son los del bloque (106 y 107).
	atributoTitulo = "titulo"
	atributoTipo   = "tipo"
	// atributoFechaDePublicacion, atributoFechaDeVigencia y
	// atributoNormaModificadora son los de la versión vigente (116-118).
	atributoFechaDePublicacion = "fecha_publicacion"
	atributoFechaDeVigencia    = "fecha_vigencia"
	atributoNormaModificadora  = "id_norma"
)

// fechaOriginal es la fecha de la versión vigente cuando la última versión no
// trae fecha_publicacion o el bloque no tiene versiones: un valor que refs/boe.py
// da (líneas 112 y 116), no el marcador de campo ausente (FR-011, FR-016).
const fechaOriginal = "original"

// espacioEnBlancoDeXML es el espacio en blanco de XML 1.0 §2.3 (producción S),
// el único texto que puede ir fuera del elemento raíz.
const espacioEnBlancoDeXML = " \t\r\n"

// Los motivos de un bloque que no se puede leer, además de motivoDelCuerpoNoUTF8:
// dicen qué no se pudo interpretar —el cuerpo, la raíz, la declaración o el
// bloque— y nunca llevan el cuerpo ni un recorte suyo (contrato
// errores-y-codigos §2, fila 16).
const (
	motivoDelXMLIlegible        = "el cuerpo no es XML legible"
	motivoSinRaiz               = "el XML no tiene elemento raíz"
	motivoDeVariasRaices        = "el XML tiene más de un elemento raíz"
	motivoDelTextoFueraDeLaRaiz = "el XML tiene texto fuera del elemento raíz"
	motivoDeLaDeclaracion       = "el XML trae una declaración <!…> que la lectura no interpreta"
	motivoSinBloque             = "el XML no tiene ningún elemento bloque por debajo de la raíz"
)

// bloqueLeido es lo que la lectura saca del XML de un bloque, lo mismo que
// xml_bloque_to_text (refs/boe.py 95-129): quien pide el bloque lo compone con la
// norma, el id pedido, los avisos y las direcciones en el artículo de data
// (data-model.md §2.1).
type bloqueLeido struct {
	// titulo y tipo son los atributos del bloque, vacíos si faltan.
	titulo string
	tipo   string
	// fechaVersion es fecha_publicacion de la última versión si el atributo
	// está, aunque venga vacío, y fechaOriginal si falta o no hay versiones.
	fechaVersion string
	// fechaVigencia y normaModificadora son fecha_vigencia e id_norma de la
	// última versión, vacíos si faltan o si no hay versiones.
	fechaVigencia     string
	normaModificadora string
	// texto es el de la última versión, o el de todo el bloque si no tiene
	// versiones, con sus líneas normalizadas.
	texto string
}

// leerBloque lee el cuerpo XML de la respuesta de un bloque con el mismo
// recorrido que hace ElementTree en xml_bloque_to_text (refs/boe.py 95-136;
// data-model.md §3.2, research.md D6):
//
//   - X1: el cuerpo es UTF-8 válido, que es como refs/boe.py lo decodifica antes
//     de analizarlo (línea 82).
//   - X2: la codificación que declara no cuenta (entradaSinTraducir).
//   - X3: tiene un único elemento raíz y fuera de él solo espacio en blanco,
//     comentarios o instrucciones de proceso, como exige expat (S5 (f)).
//     encoding/xml no lo comprueba al leer por fichas —solo que las de apertura
//     y cierre casen y que no quede ninguna abierta (xml.go 275-290)—, así que lo
//     comprueba la lectura. Una declaración <!…> tampoco se admite, ni fuera ni
//     dentro de la raíz: no es nada de lo anterior, y la de tipo de documento
//     podría declarar entidades o atributos que expat aplicaría y encoding/xml no.
//   - X4: el bloque es el primer elemento bloque sin espacio de nombres, en el
//     orden del documento, que no es la raíz: find(".//bloque") no mira la raíz
//     misma (línea 102; ElementPath.py 182-209).
//   - X5 y X6: título y tipo son atributos del bloque (106-107); las versiones,
//     sus hijos version de primer nivel (109); y la vigente, la última (115), de
//     la que salen la fecha —fecha_publicacion si el atributo está, aunque venga
//     vacío, y fechaOriginal si falta—, fecha_vigencia e id_norma (116-118). Sin
//     versiones, la fecha es fechaOriginal y lo demás va vacío (110-112). El valor
//     de cada atributo es el de atributosDeLaEtiqueta.
//   - X7: el texto es el de ET.tostring con method="text" (línea 134): todo el
//     CharData y el CDATA del elemento y de sus descendientes, en orden, más su
//     cola, el texto que sigue a su cierre hasta la etiqueta siguiente
//     (ElementTree.py 406-423 y 979-983). Los comentarios y las instrucciones de
//     proceso no aportan texto ni cortan el que los rodea, porque el TreeBuilder
//     ni los inserta ni vacía el texto acumulado (1413-1415 y 1486-1511; S5 (e)).
//     Es el de la última versión, o el de todo el bloque sin versiones (111 y
//     120).
//   - X8: sus líneas son las de lineasNormalizadas. El CRLF y el retorno suelto
//     ya llegan como salto (xml.go 1129-1136), igual que en expat.
//
// El documento se lee entero antes de devolver nada, como ET.fromstring: un
// bloque seguido de XML mal formado o de una segunda raíz tampoco se presenta.
// Lo que no se puede leer es el error de errorDeLectura con un motivo que dice
// qué no se pudo interpretar, nunca el cuerpo ni un recorte suyo (FR-014), y con
// el detalle técnico del analizador como causa, que va al registro y no al
// mensaje.
func leerBloque(cuerpo []byte) (bloqueLeido, error) {
	if !utf8.Valid(cuerpo) {
		return bloqueLeido{}, errorDeLectura(nil, motivoDelCuerpoNoUTF8)
	}

	decodificador := xml.NewDecoder(bytes.NewReader(cuerpo))
	decodificador.CharsetReader = entradaSinTraducir

	var recorrido recorridoDelBloque
	for {
		// InputOffset marca el final de la última ficha y el principio de la
		// siguiente, de modo que cada ficha es este tramo del cuerpo.
		inicio := decodificador.InputOffset()

		token, err := decodificador.Token()
		if errors.Is(err, io.EOF) {
			return recorrido.bloque()
		}

		if err != nil {
			return bloqueLeido{}, errorDeLectura(err, motivoDelXMLIlegible)
		}

		if err = recorrido.visitar(token, cuerpo[inicio:decodificador.InputOffset()]); err != nil {
			return bloqueLeido{}, err
		}
	}
}

// entradaSinTraducir es el CharsetReader del decodificador, que encoding/xml
// solo llama cuando la codificación declarada no es UTF-8 y sin el que el
// análisis fallaría (xml.go 643-658): devuelve la misma entrada, sin traducir,
// sea cual sea la declarada (X2). refs/boe.py analiza la cadena que ya
// decodificó como UTF-8 (líneas 82 y 98), y XMLParser.feed con una cadena no
// atiende a la declaración (S5 (c)). Al ser la misma entrada, InputOffset sigue
// contando las posiciones del cuerpo.
func entradaSinTraducir(_ string, entrada io.Reader) (io.Reader, error) {
	return entrada, nil
}

// recorridoDelBloque es el estado de leerBloque mientras visita las fichas del
// decodificador en el orden del documento. No construye el árbol: guarda lo
// único que xml_bloque_to_text usa de él —los atributos del bloque y de su
// última versión, y el texto de los dos con su cola—, y la cola de un elemento
// termina en la primera etiqueta, de apertura o de cierre, que llega tras su
// cierre, como el _flush del TreeBuilder (ElementTree.py 1439-1449).
type recorridoDelBloque struct {
	// profundidad es el número de elementos abiertos: la de un elemento, antes
	// de abrirlo, es el número de sus antecesores.
	profundidad int
	// hayRaiz dice si ya se abrió el elemento raíz.
	hayRaiz bool

	// hayBloque dice si ya se encontró el bloque, y bloqueAbierto, si aún no se
	// ha cerrado. nivelDelBloque es su profundidad; atributosDelBloque, sus
	// atributos; colaDelBloque dice si el texto que llega es su cola; y
	// textoDelBloque es su texto con la cola.
	hayBloque          bool
	bloqueAbierto      bool
	nivelDelBloque     int
	atributosDelBloque []xml.Attr
	colaDelBloque      bool
	textoDelBloque     strings.Builder

	// Lo mismo de la última versión de primer nivel abierta hasta ahora, que al
	// abrirse otra se sustituye.
	hayVersiones         bool
	versionAbierta       bool
	atributosDeLaVersion []xml.Attr
	colaDeLaVersion      bool
	textoDeLaVersion     strings.Builder
}

// visitar pasa una ficha por el recorrido. crudo es la ficha tal como viene en
// el cuerpo.
func (r *recorridoDelBloque) visitar(token xml.Token, crudo []byte) error {
	switch token := token.(type) {
	case xml.StartElement:
		return r.abrir(token.Name, crudo)
	case xml.EndElement:
		r.cerrar()
	case xml.CharData:
		return r.juntarTexto(token, crudo)
	case xml.Comment, xml.ProcInst:
		// Ni aportan texto ni cortan el que los rodea (X7), así que tampoco
		// terminan ninguna cola.
	case xml.Directive:
		return errorDeLectura(nil, motivoDeLaDeclaracion)
	}

	return nil
}

// abrir es la etiqueta de apertura de un elemento: cualquiera termina la cola
// que estuviera en curso (X7); a profundidad cero, es la raíz, y no puede haber
// dos (X3); si aún no hay bloque y se llama bloque, es el bloque (X4); y si es un
// hijo version de primer nivel del bloque abierto, es la última versión hasta
// ahora (X6). Los atributos de los dos se leen con atributosDeLaEtiqueta.
func (r *recorridoDelBloque) abrir(nombre xml.Name, etiqueta []byte) error {
	r.colaDelBloque, r.colaDeLaVersion = false, false

	var err error

	switch {
	case r.profundidad == 0 && r.hayRaiz:
		return errorDeLectura(nil, motivoDeVariasRaices)
	case r.profundidad == 0:
		r.hayRaiz = true
	case !r.hayBloque && sinEspacioDeNombres(nombre, elementoBloque):
		r.hayBloque, r.bloqueAbierto = true, true
		r.nivelDelBloque = r.profundidad
		r.atributosDelBloque, err = atributosDeLaEtiqueta(etiqueta)
	case r.bloqueAbierto && r.profundidad == r.nivelDelBloque+1 && sinEspacioDeNombres(nombre, elementoVersion):
		r.hayVersiones, r.versionAbierta = true, true
		r.atributosDeLaVersion, err = atributosDeLaEtiqueta(etiqueta)
		r.textoDeLaVersion.Reset()
	}

	r.profundidad++

	return err
}

// cerrar es la etiqueta de cierre de un elemento: termina la cola que estuviera
// en curso y, si cierra la versión o el bloque, empieza la suya (X7). El
// decodificador ya ha comprobado que casa con la de apertura.
func (r *recorridoDelBloque) cerrar() {
	r.colaDelBloque, r.colaDeLaVersion = false, false
	r.profundidad--

	switch {
	case r.versionAbierta && r.profundidad == r.nivelDelBloque+1:
		r.versionAbierta, r.colaDeLaVersion = false, true
	case r.bloqueAbierto && r.profundidad == r.nivelDelBloque:
		r.bloqueAbierto, r.colaDelBloque = false, true
	}
}

// juntarTexto es un tramo de CharData, o una sección CDATA, que encoding/xml
// entrega igual (xml.go 699-715): fuera de la raíz solo puede ser espacio en
// blanco escrito tal cual —ni CDATA ni referencias, que tampoco admite XML 1.0
// §2.8— (X3); dentro, entra en el texto del bloque y en el de la versión si está
// dentro de ellos o en su cola (X7).
func (r *recorridoDelBloque) juntarTexto(texto xml.CharData, crudo []byte) error {
	if r.profundidad == 0 {
		if len(bytes.Trim(crudo, espacioEnBlancoDeXML)) > 0 {
			return errorDeLectura(nil, motivoDelTextoFueraDeLaRaiz)
		}

		return nil
	}

	if r.bloqueAbierto || r.colaDelBloque {
		r.textoDelBloque.Write(texto)
	}

	if r.versionAbierta || r.colaDeLaVersion {
		r.textoDeLaVersion.Write(texto)
	}

	return nil
}

// bloque es lo leído cuando el documento termina bien formado: el error si no
// tuvo raíz (X3) o bloque (X4), y si no, sus atributos y su texto, de la última
// versión si la hay (X5, X6).
func (r *recorridoDelBloque) bloque() (bloqueLeido, error) {
	switch {
	case !r.hayRaiz:
		return bloqueLeido{}, errorDeLectura(nil, motivoSinRaiz)
	case !r.hayBloque:
		return bloqueLeido{}, errorDeLectura(nil, motivoSinBloque)
	}

	titulo, _ := valorDelAtributo(r.atributosDelBloque, atributoTitulo)
	tipo, _ := valorDelAtributo(r.atributosDelBloque, atributoTipo)
	leido := bloqueLeido{
		titulo:       titulo,
		tipo:         tipo,
		fechaVersion: fechaOriginal,
		texto:        lineasNormalizadas(r.textoDelBloque.String()),
	}

	if !r.hayVersiones {
		return leido, nil
	}

	if fecha, esta := valorDelAtributo(r.atributosDeLaVersion, atributoFechaDePublicacion); esta {
		leido.fechaVersion = fecha
	}

	leido.fechaVigencia, _ = valorDelAtributo(r.atributosDeLaVersion, atributoFechaDeVigencia)
	leido.normaModificadora, _ = valorDelAtributo(r.atributosDeLaVersion, atributoNormaModificadora)
	leido.texto = lineasNormalizadas(r.textoDeLaVersion.String())

	return leido, nil
}

// atributosDeLaEtiqueta son los atributos de una etiqueta de apertura con el
// valor normalizado como manda XML 1.0 §3.3.3 para un atributo CDATA, que es lo
// que expat entrega a ElementTree (X5, S5 (d)) y encoding/xml no hace (xml.go
// 860-868): cada tabulador, salto o retorno escrito tal cual pasa a un espacio
// —el CRLF, a uno solo, porque el fin de línea de §2.11 lo reduce antes a un
// salto—, sin recortar el valor, y el carácter de una referencia (&#9;, &#10;,
// &#13;) se conserva.
//
// El decodificador resuelve las referencias al leer, así que en xml.Attr un
// salto escrito y &#10; ya no se distinguen: por eso se normaliza la etiqueta tal
// como viene en el cuerpo y se vuelve a leer con el mismo decodificador, que
// resuelve entidades y referencias como la primera vez. Dentro de una etiqueta no
// hay comentarios, CDATA ni texto, así que lo que va entre comillas es el valor
// de un atributo.
//
// leerBloque solo le pasa etiquetas que el decodificador ya aceptó; lo que no se
// pueda volver a leer como etiqueta de apertura es igualmente ilegible.
func atributosDeLaEtiqueta(etiqueta []byte) ([]xml.Attr, error) {
	normalizada := make([]byte, 0, len(etiqueta))

	var comilla byte

	for posicion, caracter := range etiqueta {
		switch {
		case comilla == 0:
			if caracter == '"' || caracter == '\'' {
				comilla = caracter
			}
		case caracter == comilla:
			comilla = 0
		case caracter == '\r' && posicion+1 < len(etiqueta) && etiqueta[posicion+1] == '\n':
			continue
		case caracter == '\t' || caracter == '\n' || caracter == '\r':
			caracter = ' '
		}

		normalizada = append(normalizada, caracter)
	}

	token, err := xml.NewDecoder(bytes.NewReader(normalizada)).RawToken()

	apertura, esApertura := token.(xml.StartElement)
	if !esApertura {
		return nil, errorDeLectura(err, motivoDelXMLIlegible)
	}

	return apertura.Attr, nil
}

// valorDelAtributo es el valor del atributo sin espacio de nombres que tiene ese
// nombre, y si está: lo que distinguen bloque.get("titulo", "") y
// ultima.get("fecha_publicacion", "original") en refs/boe.py (106-107 y 116-118),
// donde un atributo con espacio de nombres se llama {uri}nombre y no casa.
func valorDelAtributo(atributos []xml.Attr, nombre string) (string, bool) {
	for _, atributo := range atributos {
		if sinEspacioDeNombres(atributo.Name, nombre) {
			return atributo.Value, true
		}
	}

	return "", false
}

// sinEspacioDeNombres dice si un nombre del XML es ese nombre local sin espacio
// de nombres, el único que casa con el bloque, la versión y los atributos que
// ElementTree busca por su nombre a secas (X4).
func sinEspacioDeNombres(nombre xml.Name, local string) bool {
	return nombre.Space == "" && nombre.Local == local
}

// lineasNormalizadas es la normalización de _elem_all_text (refs/boe.py 132-136;
// X8): el texto partido por saltos de línea, cada línea recortada con el espacio
// en blanco de str.strip —esEspacioDePython—, sin las que quedan vacías y unidas
// de nuevo con un salto. El espacio del interior de cada línea no se toca.
func lineasNormalizadas(texto string) string {
	lineas := make([]string, 0, strings.Count(texto, "\n")+1)

	for linea := range strings.SplitSeq(texto, "\n") {
		if recortada := strings.TrimFunc(linea, esEspacioDePython); recortada != "" {
			lineas = append(lineas, recortada)
		}
	}

	return strings.Join(lineas, "\n")
}
