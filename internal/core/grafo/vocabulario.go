package grafo

// Los tipos de nodo del grafo del mundo (data-model §7). Un id tiene siempre
// el mismo tipo: un lote que le da otro se rechaza entero (FR-024).
const (
	// TipoNorma es una norma consolidada, por su ELI.
	TipoNorma = "Norma"
	// TipoBloque es un bloque de una norma —un artículo, una disposición—,
	// por el ELI de la norma y el id del bloque.
	TipoBloque = "Bloque"
	// TipoBloqueVersion es una versión de un bloque, por la fecha desde la que
	// está vigente y la huella de su texto.
	TipoBloqueVersion = "BloqueVersion"
	// TipoMunicipio es un municipio, por «ine:» y su código INE.
	TipoMunicipio = "Municipio"
	// TipoOrgano es un órgano de una administración, por su código DIR3.
	TipoOrgano = "Organo"
	// TipoPersona es una persona. Ningún applet lo emite en este hito; lo
	// vigila la validación del lote, que rechaza una Persona con un documento
	// de identidad en su id o en sus datos (FR-025, constitución VII).
	TipoPersona = "Persona"
	// TipoResolucion es una resolución judicial, por su ECLI. `graph check` no
	// da hallazgos de ella: solo los da de una norma, un bloque o una versión
	// de bloque (H23 FR-042, FR-043).
	TipoResolucion = "Resolucion"
)

// Las relaciones del grafo del mundo, cada una entre dos tipos de nodo
// (data-model §7).
const (
	// RelacionTieneParte va de una Norma a cada uno de sus Bloque.
	RelacionTieneParte = "eli:has_part"
	// RelacionTieneVersion va de un Bloque a cada una de sus BloqueVersion.
	RelacionTieneVersion = "eli:has_version"
	// RelacionPerteneceA va de un Organo al Municipio al que pertenece.
	RelacionPerteneceA = "lb:pertenece_a"
)

// Las claves de los datos identificativos de cada tipo de nodo (data-model
// §7). Ninguna lleva el cuerpo de un bloque, que viaja en un Texto (FR-041).
const (
	// DatoIdentificador es el identificador BOE de una Norma.
	DatoIdentificador = "identificador"
	// DatoBloque es el id de un Bloque dentro de su norma, como «a21».
	DatoBloque = "bloque"
	// DatoFechaVigencia es la fecha, AAAAMMDD, desde la que está vigente una
	// BloqueVersion.
	DatoFechaVigencia = "fecha_vigencia"
	// DatoFechaVersion es la fecha, AAAAMMDD, de la BloqueVersion.
	DatoFechaVersion = "fecha_version"
	// DatoNormaModificadora es el identificador de la norma que dio lugar a
	// la BloqueVersion.
	DatoNormaModificadora = "norma_modificadora"
	// DatoHashTexto es la huella del texto de la BloqueVersion, la del Texto
	// que guarda su cuerpo.
	DatoHashTexto = "hash_texto"
	// DatoCodigoINE es el código INE de cinco cifras de un Municipio.
	DatoCodigoINE = "codigo_ine"
	// DatoNombre es el nombre oficial de un Municipio.
	DatoNombre = "nombre"
	// DatoDIR3 es el código DIR3 de un Organo.
	DatoDIR3 = "dir3"
)

// Las claves de los datos de una Resolucion: sus metadatos, con los nombres
// que llevan en `data` (H23 FR-042; data-model §5). Ninguna lleva el resumen
// ni el texto de la resolución, y el ponente es un dato del nodo: no da lugar
// a ningún nodo Persona.
const (
	// DatoECLI es el ECLI de una Resolucion, que es también su id.
	DatoECLI = "ecli"
	// DatoROJ es el ROJ de una Resolucion, el identificador que le da el
	// CENDOJ.
	DatoROJ = "roj"
	// DatoOrgano es el órgano judicial, con su sala, que dictó una Resolucion.
	DatoOrgano = "organo"
	// DatoFecha es la fecha, AAAA-MM-DD, de una Resolucion.
	DatoFecha = "fecha"
	// DatoNumeroResolucion es el número de una Resolucion.
	DatoNumeroResolucion = "numero_resolucion"
	// DatoNumeroRecurso es el número del recurso que resuelve una Resolucion.
	DatoNumeroRecurso = "numero_recurso"
	// DatoPonente es el nombre de quien fue ponente de una Resolucion.
	DatoPonente = "ponente"
	// DatoURL es la URL del documento de una Resolucion, tal como la da el
	// CENDOJ.
	DatoURL = "url"
)

// ClaseDeHallazgo es la clase de un hallazgo de `graph check`. Es un tipo
// propio para que no se confunda con schema.Clase: un hallazgo no es un fallo,
// sino un dato de un resultado correcto (ADR 0023).
type ClaseDeHallazgo string

// Las clases de hallazgo de `graph check` (data-model §6 y §7).
const (
	// ClaseFuenteCaducada es un nodo cuya última consulta declaró una vigencia
	// que ya ha pasado (FR-066).
	ClaseFuenteCaducada ClaseDeHallazgo = "fuente-caducada"
	// ClaseVersionObsoleta es una versión de un bloque superada por otra de
	// fecha de vigencia posterior (FR-063).
	ClaseVersionObsoleta ClaseDeHallazgo = "version-obsoleta"
)

// etiquetaDeVersionObsoleta es lo que lleva entre la marca y los dos puntos la
// forma fija con la que una skill traslada un version-obsoleta.
const etiquetaDeVersionObsoleta = "REDACCIÓN MODIFICADA"

// EtiquetasDeHallazgo es la etiqueta de cada clase de hallazgo que una skill
// traslada con forma fija: lo que su frase lleva entre la marca y los dos
// puntos, y la única fuente de verdad de esa forma (H7.1 FR-045; data-model §7).
// Solo la tiene ClaseVersionObsoleta: un fuente-caducada no se traslada
// (FR-043). Ninguna es la de un aviso de vigencia. El binario no escribe la
// forma en ninguna salida. Cada llamada devuelve un mapa nuevo, que quien lo
// recibe puede cambiar sin cambiar el de nadie más.
func EtiquetasDeHallazgo() map[ClaseDeHallazgo]string {
	return map[ClaseDeHallazgo]string{ClaseVersionObsoleta: etiquetaDeVersionObsoleta}
}

// MaximoDeHallazgos es cuántos hallazgos lista, como mucho, `graph check`: los
// primeros de su orden; los totales de cada clase cuentan también los que no
// lista (H7.1 FR-010, FR-012; research.md D10). No hay bandera que lo cambie
// (FR-015).
const MaximoDeHallazgos = 50
