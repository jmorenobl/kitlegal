package boe

import "github.com/jmorenobl/kitlegal/internal/core"

// Los verbos del applet boe, uno por consulta y con los nombres de las órdenes
// de refs/boe.py (contrato puerto-y-applet §3.2).
const (
	verboBuscar    = "buscar"
	verboIndice    = "indice"
	verboArticulo  = "articulo"
	verboArticulos = "articulos"
	verboMetadatos = "metadatos"
	verboAnalisis  = "analisis"
)

// ConsultaBuscar es buscar <texto>…: las normas consolidadas cuyo título casa
// con el texto (FR-030).
type ConsultaBuscar struct {
	// Texto son los argumentos del verbo tal como llegan y en su orden; la fuente
	// los une con un espacio antes de construir la consulta de búsqueda.
	Texto []string
}

// Verbo es buscar.
func (ConsultaBuscar) Verbo() string {
	return verboBuscar
}

// ConsultaIndice es indice <norma>: los bloques de la norma en el orden de la
// fuente (FR-040).
type ConsultaIndice struct {
	// Norma es el identificador de la norma, BOE-A-<año>-<número>, que la fuente
	// valida antes de nada.
	Norma string
}

// Verbo es indice.
func (ConsultaIndice) Verbo() string {
	return verboIndice
}

// ConsultaArticulo es articulo <norma> <bloque>: el texto vigente de un bloque
// con sus avisos de vigencia (FR-010).
type ConsultaArticulo struct {
	// Norma es el identificador de la norma, BOE-A-<año>-<número>, que la fuente
	// valida antes de nada.
	Norma string
	// Bloque es el id del bloque dentro de la norma, como a21 o da3, que la
	// fuente valida antes de nada.
	Bloque string
}

// Verbo es articulo.
func (ConsultaArticulo) Verbo() string {
	return verboArticulo
}

// ConsultaArticulos es articulos <norma> <bloque> [<bloque>…]: varios bloques de
// la misma norma en una invocación (FR-020).
type ConsultaArticulos struct {
	// Norma es el identificador de la norma, BOE-A-<año>-<número>, que la fuente
	// valida antes de nada.
	Norma string
	// Bloques son los ids de bloque en el orden pedido, con sus repeticiones; la
	// fuente los valida todos antes de nada y pide cada id distinto una sola vez.
	Bloques []string
}

// Verbo es articulos.
func (ConsultaArticulos) Verbo() string {
	return verboArticulos
}

// ConsultaMetadatos es metadatos <norma>: los datos de la norma y sus avisos de
// vigencia (FR-050).
type ConsultaMetadatos struct {
	// Norma es el identificador de la norma, BOE-A-<año>-<número>, que la fuente
	// valida antes de nada.
	Norma string
}

// Verbo es metadatos.
func (ConsultaMetadatos) Verbo() string {
	return verboMetadatos
}

// ConsultaAnalisis es analisis <norma>: las materias, las notas y las
// referencias de la norma (FR-060).
type ConsultaAnalisis struct {
	// Norma es el identificador de la norma, BOE-A-<año>-<número>, que la fuente
	// valida antes de nada.
	Norma string
}

// Verbo es analisis.
func (ConsultaAnalisis) Verbo() string {
	return verboAnalisis
}

// Las seis consultas implementan el puerto del dominio por valor, y con ello
// también por puntero, y que lo sigan haciendo no depende de que alguien lo
// recuerde.
var (
	_ core.Consulta = ConsultaBuscar{}
	_ core.Consulta = ConsultaIndice{}
	_ core.Consulta = ConsultaArticulo{}
	_ core.Consulta = ConsultaArticulos{}
	_ core.Consulta = ConsultaMetadatos{}
	_ core.Consulta = ConsultaAnalisis{}
)
