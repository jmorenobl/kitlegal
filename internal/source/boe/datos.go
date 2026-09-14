package boe

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Los tipos de data de los seis verbos, con las claves de data-model.md §2 en
// español (research.md D11). Ninguna lleva omitempty: todas son obligatorias en
// el esquema y se emiten siempre, un campo que la fuente no da va como cadena
// vacía (FR-016), nunca con el marcador ? de refs/boe.py, y una lista sin
// elementos va como [] y nunca como null, así que quien compone los datos crea
// vacías las listas que no llena.
//
// Las etiquetas jsonschema son la descripción formal de los valores, de la que
// --describe genera el esquema de cada verbo y de la que salen schemas/norma.json
// y schemas/bloque.json (FR-110). Una etiqueta no puede nombrar una constante,
// así que los enumerados repiten CodigosDeAviso y TiposDeBloque y el patrón de
// hash_texto repite schema.PatronHuella; TestEnumeradosDeLosDatos y
// TestHashTexto exigen que digan lo mismo.

// Articulo es el data de articulo y cada elemento del de articulos: un bloque de
// una norma consolidada en su versión vigente, con los avisos de vigencia de la
// norma y los identificadores naturales que el applet observa (FR-010, FR-011,
// FR-012, FR-015, data-model.md §2.1).
type Articulo struct {
	// Norma es el identificador pedido, ya validado (refs/boe.py 399).
	Norma string `json:"norma"`
	// Bloque es el id de bloque pedido, ya validado (refs/boe.py 399).
	Bloque string `json:"bloque"`
	// Titulo es el atributo titulo del bloque (refs/boe.py 106).
	Titulo string `json:"titulo"`
	// Tipo es el atributo tipo del bloque, no el que TipoDesdeID infiere del id
	// (refs/boe.py 107).
	Tipo string `json:"tipo"`
	// FechaVersion es la fecha_publicacion de la última versión, u «original»
	// si falta o si el bloque no tiene versiones (refs/boe.py 112 y 116).
	FechaVersion string `json:"fecha_version"`
	// FechaVigencia es la fecha_vigencia de la última versión (refs/boe.py 117).
	FechaVigencia string `json:"fecha_vigencia"`
	// NormaModificadora es el id_norma de la última versión (refs/boe.py 118).
	NormaModificadora string `json:"norma_modificadora"`
	// Texto es el texto normalizado de la última versión, o del bloque entero
	// si no tiene versiones (refs/boe.py 111, 120 y 132-136).
	Texto string `json:"texto"`
	// HashTexto es la huella de Texto que da huellaDelTexto (FR-015, SC-013).
	HashTexto string `json:"hash_texto" jsonschema:"pattern=^sha256:[0-9a-f]{64}$"`
	// Avisos son los avisos de vigencia de la norma, derivados de sus metadatos
	// (refs/boe.py 197-211).
	Avisos []Aviso `json:"avisos"`
	// URL es la dirección pública del bloque, act.php con la norma y el bloque
	// como fragmento (refs/boe.py 430).
	URL string `json:"url"`
	// URLELI es la url_eli de los metadatos de la norma (FR-015).
	URLELI string `json:"url_eli"`
}

// Aviso es un aviso de vigencia: el código de la condición que lo produce y su
// frase literal, la misma en articulo, en cada elemento de articulos y en
// metadatos (FR-012, FR-050, data-model.md §2.2). Los compone avisosDe.
type Aviso struct {
	// Codigo es uno de CodigosDeAviso, en kebab-case (Q5).
	Codigo string `json:"codigo" jsonschema:"enum=consolidacion-no-finalizada,enum=derogada,enum=vigencia-agotada"`
	// Texto es la frase de _check_vigencia, carácter a carácter y con su prefijo
	// (refs/boe.py 202-211).
	Texto string `json:"texto"`
}

// ResultadoDeBusqueda es cada elemento del data de buscar, en el orden de la
// fuente (FR-030, FR-032, data-model.md §2.3).
type ResultadoDeBusqueda struct {
	// Identificador es el identificador BOE-A-… del resultado (refs/boe.py 293).
	Identificador string `json:"identificador"`
	// Titulo es el titulo del resultado (refs/boe.py 294).
	Titulo string `json:"titulo"`
	// Rango es rango.texto si rango es un objeto (refs/boe.py 290-291 y 295).
	Rango string `json:"rango"`
	// VigenciaAgotada es la vigencia_agotada del resultado (refs/boe.py 296).
	VigenciaAgotada string `json:"vigencia_agotada"`
	// EstadoConsolidacion es estado_consolidacion.texto si es un objeto
	// (refs/boe.py 288-289 y 297).
	EstadoConsolidacion string `json:"estado_consolidacion"`
	// URL es la dirección pública de la norma, act.php con el identificador tal
	// como llega (refs/boe.py 298).
	URL string `json:"url"`
}

// Indice es el data de indice: los bloques de una norma en el orden de la
// fuente (FR-040, data-model.md §2.4).
type Indice struct {
	// Norma es el identificador pedido (refs/boe.py 349).
	Norma string `json:"norma"`
	// URL es la dirección pública de la norma (refs/boe.py 384).
	URL string `json:"url"`
	// Bloques son las entradas del índice (refs/boe.py 365-392).
	Bloques []EntradaDeIndice `json:"bloques"`
}

// EntradaDeIndice es un bloque del índice de una norma (data-model.md §2.4).
type EntradaDeIndice struct {
	// ID es el id del bloque (refs/boe.py 387).
	ID string `json:"id"`
	// Titulo es el titulo del bloque, tal como llega (refs/boe.py 388).
	Titulo string `json:"titulo"`
	// Tipo es el que TipoDesdeID infiere del id, uno de TiposDeBloque
	// (refs/boe.py 389).
	Tipo string `json:"tipo" jsonschema:"enum=articulo,enum=titulo,enum=capitulo,enum=seccion,enum=preambulo,enum=disposicion_adicional,enum=disposicion_transitoria,enum=disposicion_derogatoria,enum=disposicion_final,enum="`
}

// Metadatos es el data de metadatos: el estado de una norma y los avisos de
// vigencia que se derivan de él (FR-050, data-model.md §2.5).
type Metadatos struct {
	// Norma es el identificador pedido (refs/boe.py 446).
	Norma string `json:"norma"`
	// Titulo es el titulo de la norma (refs/boe.py 473).
	Titulo string `json:"titulo"`
	// Rango es rango.texto si rango es un objeto, o la cadena si es una cadena
	// (refs/boe.py 468-469).
	Rango string `json:"rango"`
	// NumeroOficial es el numero_oficial (refs/boe.py 475).
	NumeroOficial string `json:"numero_oficial"`
	// FechaDisposicion es la fecha_disposicion (refs/boe.py 476).
	FechaDisposicion string `json:"fecha_disposicion"`
	// FechaPublicacion es la fecha_publicacion (refs/boe.py 477).
	FechaPublicacion string `json:"fecha_publicacion"`
	// FechaVigencia es la fecha_vigencia (refs/boe.py 478).
	FechaVigencia string `json:"fecha_vigencia"`
	// EstatusDerogacion es el estatus_derogacion (refs/boe.py 479).
	EstatusDerogacion string `json:"estatus_derogacion"`
	// VigenciaAgotada es la vigencia_agotada (refs/boe.py 480).
	VigenciaAgotada string `json:"vigencia_agotada"`
	// EstadoConsolidacion es el estado de la consolidación (refs/boe.py 465-467).
	EstadoConsolidacion EstadoDeConsolidacion `json:"estado_consolidacion"`
	// URLELI es la url_eli de la norma (refs/boe.py 482, FR-015).
	URLELI string `json:"url_eli"`
	// Avisos son los de _check_vigencia, no los de cmd_metadatos
	// (refs/boe.py 199-211; entrada 17 del porte anotado en doc.go).
	Avisos []Aviso `json:"avisos"`
}

// EstadoDeConsolidacion es el estado_consolidacion de los metadatos
// (data-model.md §2.5).
type EstadoDeConsolidacion struct {
	// Codigo es estado.codigo si el estado es un objeto (refs/boe.py 467).
	Codigo string `json:"codigo"`
	// Texto es estado.texto si el estado es un objeto, o la cadena si es una
	// cadena (refs/boe.py 466).
	Texto string `json:"texto"`
}

// Analisis es el data de analisis: las materias, las notas y las referencias de
// una norma (FR-060, data-model.md §2.6).
type Analisis struct {
	// Norma es el identificador pedido (refs/boe.py 505).
	Norma string `json:"norma"`
	// Materias son las materias de la norma (refs/boe.py 527-535).
	Materias []Materia `json:"materias"`
	// Notas son las notas de la norma, cada una su texto (refs/boe.py 538-541).
	Notas []string `json:"notas"`
	// Referencias son las normas con las que se relaciona (refs/boe.py 544-581).
	Referencias Referencias `json:"referencias"`
}

// Materia es una materia del análisis: una que no es objeto aporta solo su
// texto (refs/boe.py 527-535, data-model.md §2.6).
type Materia struct {
	// Codigo es el codigo de la materia.
	Codigo string `json:"codigo"`
	// Texto es el texto de la materia.
	Texto string `json:"texto"`
}

// Referencias son las dos listas de referencias del análisis, vacías las dos si
// la fuente no da un objeto (refs/boe.py 545, data-model.md §2.6).
type Referencias struct {
	// Anteriores son las normas que esta modifica o deroga
	// (refs/boe.py 544-564).
	Anteriores []ReferenciaAnterior `json:"anteriores"`
	// Posteriores son las normas que la modifican (refs/boe.py 566-581).
	Posteriores []ReferenciaPosterior `json:"posteriores"`
}

// ReferenciaAnterior es una norma que esta modifica o deroga
// (data-model.md §2.6).
type ReferenciaAnterior struct {
	// Relacion es relacion.texto si relacion es un objeto, o la cadena si es una
	// cadena (refs/boe.py 557-560).
	Relacion string `json:"relacion"`
	// Norma es el id_norma de la referencia (refs/boe.py 561).
	Norma string `json:"norma"`
	// Texto es el texto de la referencia, completo y sin el recorte a 200
	// caracteres de refs/boe.py 564 (FR-060; entrada 29 del porte anotado en
	// doc.go).
	Texto string `json:"texto"`
}

// ReferenciaPosterior es una norma que modifica esta (data-model.md §2.6).
type ReferenciaPosterior struct {
	// Relacion es relacion.texto si relacion es un objeto, o la cadena si es una
	// cadena (refs/boe.py 577-580).
	Relacion string `json:"relacion"`
	// Norma es el id_norma de la referencia (refs/boe.py 581).
	Norma string `json:"norma"`
}

// TiposDeBloque son los valores del enumerado tipo de EntradaDeIndice, en su
// orden: los nueve tipos que TipoDesdeID infiere, en el orden de sus reglas, y
// la cadena vacía del id que no casa con ninguna (data-model.md §2.4 y §4). Cada
// llamada devuelve una lista nueva, que quien la recibe puede cambiar sin
// cambiar la de nadie más.
func TiposDeBloque() []string {
	return []string{
		tipoArticulo, tipoTitulo, tipoCapitulo, tipoSeccion, tipoPreambulo, tipoDisposicionAdicional,
		tipoDisposicionTransitoria, tipoDisposicionDerogatoria, tipoDisposicionFinal, sinTipo,
	}
}

// huellaDelTexto es el hash_texto de un artículo: schema.PrefijoHuella, que
// nombra el algoritmo, seguido del hexadecimal en minúsculas del SHA-256 de los
// bytes UTF-8 del texto, tal cual y sin normalizarlo, de modo que cambia si y
// solo si cambia el texto (FR-015, SC-013, data-model.md §2.1). No es
// schema.Huella, que resume la forma canónica de un data entero y no los bytes de
// una cadena.
func huellaDelTexto(texto string) string {
	suma := sha256.Sum256([]byte(texto))

	return schema.PrefijoHuella + hex.EncodeToString(suma[:])
}
