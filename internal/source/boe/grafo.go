package boe

import (
	"net/url"
	"slices"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// segmentoELI es el segmento de la ruta de url_eli desde el que empieza el id de
// una Norma (FR-040).
const segmentoELI = "eli"

// observadoDeLosArticulos es lo que observan del mundo los artículos de éxito de
// articulo y articulos (contracts/emision.md §1; research.md D20): las
// operaciones de operacionesDelArticulo de cada bloque distinto, una sola vez y
// en el orden de su primera aparición, con la vigencia con la que se guarda la
// consulta de un artículo, vigenciaLarga (FR-065). Si ningún artículo tiene ELI
// no observa nada, y lo que da es el valor cero, que no declara vigencia.
func observadoDeLosArticulos(articulos ...Articulo) schema.Observado {
	var operaciones []schema.Operacion

	vistos := make(map[string]bool, len(articulos))

	for _, articulo := range articulos {
		if vistos[articulo.Bloque] {
			continue
		}

		vistos[articulo.Bloque] = true
		operaciones = append(operaciones, operacionesDelArticulo(articulo)...)
	}

	if len(operaciones) == 0 {
		return schema.Observado{}
	}

	return schema.Observado{Vigencia: vigenciaLarga, Operaciones: operaciones}
}

// operacionesDelArticulo son las operaciones que emite un artículo con los
// identificadores naturales que ya lleva su data (FR-040): la Norma por el ELI de
// su url_eli, con su identificador BOE; el Bloque, <eli>#<bloque>, con su id de
// bloque; la BloqueVersion, <eli>#<bloque>@<fecha_vigencia>:<hash_texto>, con
// esas dos fechas, la norma modificadora y la huella tal como las lleva el
// artículo; la arista eli:has_part de la Norma al Bloque y la eli:has_version del
// Bloque a la BloqueVersion; y el texto por su hash_texto, que es la SHA-256 de
// sus bytes (FR-071). Ningún nodo lleva el cuerpo, que solo viaja en el texto
// (FR-041).
//
// Sin ELI no emite nada: los ids del Bloque y de la BloqueVersion se componen a
// partir del de la Norma, y ninguno se sustituye por otro (FR-040).
func operacionesDelArticulo(articulo Articulo) []schema.Operacion {
	norma, conELI := eliDeLaNorma(articulo.URLELI)
	if !conELI {
		return nil
	}

	bloque := norma + "#" + articulo.Bloque
	version := bloque + "@" + articulo.FechaVigencia + ":" + articulo.HashTexto

	return []schema.Operacion{
		schema.Nodo{ID: norma, Tipo: grafo.TipoNorma, Datos: map[string]any{grafo.DatoIdentificador: articulo.Norma}},
		schema.Nodo{ID: bloque, Tipo: grafo.TipoBloque, Datos: map[string]any{grafo.DatoBloque: articulo.Bloque}},
		schema.Nodo{ID: version, Tipo: grafo.TipoBloqueVersion, Datos: map[string]any{
			grafo.DatoFechaVigencia:     articulo.FechaVigencia,
			grafo.DatoFechaVersion:      articulo.FechaVersion,
			grafo.DatoNormaModificadora: articulo.NormaModificadora,
			grafo.DatoHashTexto:         articulo.HashTexto,
		}},
		schema.Arista{Origen: norma, Relacion: grafo.RelacionTieneParte, Destino: bloque},
		schema.Arista{Origen: bloque, Relacion: grafo.RelacionTieneVersion, Destino: version},
		schema.Texto{Huella: articulo.HashTexto, Cuerpo: articulo.Texto},
	}
}

// eliDeLaNorma es el id de la Norma que sale de su url_eli (FR-040): la ruta de
// la dirección, analizada con net/url, desde su primer segmento que es
// exactamente eli hasta el final, sin la consulta ni el fragmento; de
// https://www.boe.es/eli/es/l/2015/10/01/39, eli/es/l/2015/10/01/39. Los
// segmentos son los de la ruta tal como va escrita en la dirección, de modo que
// una barra escapada, %2F, no separa dos. Dice si hay id: no lo hay si url_eli
// está vacío, no se puede analizar o su ruta no tiene ningún segmento eli.
func eliDeLaNorma(urlELI string) (string, bool) {
	direccion, err := url.Parse(urlELI)
	if err != nil {
		return "", false
	}

	segmentos := strings.Split(direccion.EscapedPath(), "/")

	desde := slices.Index(segmentos, segmentoELI)
	if desde < 0 {
		return "", false
	}

	return strings.Join(segmentos[desde:], "/"), true
}
