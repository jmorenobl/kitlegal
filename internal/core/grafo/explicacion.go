package grafo

import (
	"fmt"
	"time"
)

// Las explicaciones de los hallazgos de `graph check`: una frase en español
// con la plantilla literal de contracts/applet-graph.md §5, en la que cada
// valor —la cita, la fuente, la url, la fecha de consulta y cada fecha de
// vigencia— se copia carácter a carácter de lo guardado, sin reformatearlo
// (FR-061). Solo el instante en que caducó una consulta se calcula, y se
// escribe como el sobre escribe su fecha_consulta (FR-066; research.md D18).
//
// La cita de un nodo es la de data-model §6: el identificador de una Norma;
// «[<identificador de su Norma>, bloque <bloque>]» de un Bloque, con la Norma
// de la arista eli:has_part que llega a él; y la de su Bloque de una
// BloqueVersion, por la arista eli:has_version que llega a ella. Sin esos
// datos —falta alguno, no es un texto o está vacío, la arista no llega desde un
// nodo del tipo que la cita pide, o llegan desde dos y la cita no es una
// sola—, y en cualquier otro tipo de nodo, la cita es el id del nodo.

const (
	// plantillaVersionObsoleta es la de version-obsoleta: la cita de la versión
	// superada, su fecha de vigencia, la de la más reciente, y la url y la
	// fecha de consulta de la última observación de la más reciente.
	plantillaVersionObsoleta = "La versión de %s con fecha de vigencia %s está superada por la de fecha de " +
		"vigencia %s, observada en %s el %s."
	// plantillaFuenteCaducada es la de fuente-caducada: la cita del nodo, la
	// fuente, la url y la fecha de consulta de su última observación, la
	// vigencia en segundos y el instante en que caducó.
	plantillaFuenteCaducada = "La consulta de %s a %s en %s del %s tenía una vigencia de %d s y caducó el %s."
)

// explicarVersionObsoleta es la explicación de la versión citada, de fecha de
// vigencia vigente, superada por la de fecha de vigencia reciente, cuya última
// observación es observada.
func explicarVersionObsoleta(cita, vigente, reciente string, observada Procedencia) string {
	return fmt.Sprintf(plantillaVersionObsoleta, cita, vigente, reciente, observada.URL, observada.FechaConsulta)
}

// explicarFuenteCaducada es la explicación de la consulta del nodo citado,
// cuya última observación es observada y declaró esa vigencia, que caducó en
// caducidad.
func explicarFuenteCaducada(cita string, observada Procedencia, vigencia time.Duration, caducidad time.Time) string {
	return fmt.Sprintf(plantillaFuenteCaducada, cita, observada.Fuente, observada.URL, observada.FechaConsulta,
		int64(vigencia/time.Second), caducidad.Format(time.RFC3339Nano))
}

// instanteDeCaducidad es el de la fecha de consulta observada más la vigencia,
// en el desplazamiento de la fecha de consulta: un cambio de hora de por medio
// en la zona local no lo mueve, y con RFC3339Nano se escribe con Z si la fecha
// lleva Z y con la fracción de segundo sin ceros a la derecha, solo si no es
// cero (research.md D18).
func instanteDeCaducidad(observada time.Time, vigencia time.Duration) time.Time {
	_, desplazamiento := observada.Zone()

	return observada.In(time.FixedZone("", desplazamiento)).Add(vigencia)
}

// citar es la cita del nodo de ese id en una explicación, o su id si no tiene
// otra.
func (i indice) citar(id string) string {
	if cita, citable := i.cita(id); citable {
		return cita
	}

	return id
}

// cita es la del nodo de ese id según su tipo, y si la tiene.
func (i indice) cita(id string) (string, bool) {
	nodo := i.nodos[id]

	switch nodo.Tipo {
	case TipoNorma:
		return datoDeTexto(nodo, DatoIdentificador)
	case TipoBloque:
		return i.citaDeBloque(nodo)
	case TipoBloqueVersion:
		bloque, unico := i.origenUnico(RelacionTieneVersion, id, TipoBloque)
		if !unico {
			return "", false
		}

		return i.citaDeBloque(bloque)
	default:
		return "", false
	}
}

// citaDeBloque es la del Bloque: el identificador de la Norma de la que cuelga
// y su bloque, y si la tiene.
func (i indice) citaDeBloque(bloque nodoLeido) (string, bool) {
	norma, unica := i.origenUnico(RelacionTieneParte, bloque.ID, TipoNorma)
	if !unica {
		return "", false
	}

	identificador, conIdentificador := datoDeTexto(norma, DatoIdentificador)
	parte, conParte := datoDeTexto(bloque, DatoBloque)

	if !conIdentificador || !conParte {
		return "", false
	}

	return "[" + identificador + ", bloque " + parte + "]", true
}

// origenUnico es el nodo del tipo pedido desde el que llega al de ese id una
// arista de la relación, si es uno solo.
func (i indice) origenUnico(relacion, id, tipo string) (nodoLeido, bool) {
	var (
		origen   nodoLeido
		contados int
	)

	for _, candidato := range i.entrantes[extremo{relacion: relacion, id: id}] {
		if nodo, existe := i.nodos[candidato]; existe && nodo.Tipo == tipo {
			origen = nodo
			contados++
		}
	}

	return origen, contados == 1
}

// datoDeTexto es el dato de esa clave del nodo, si es un texto que no está
// vacío.
func datoDeTexto(nodo nodoLeido, clave string) (string, bool) {
	valor, esTexto := nodo.Datos[clave].(string)

	return valor, esTexto && valor != ""
}
