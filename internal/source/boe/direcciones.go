package boe

// Las dos raíces de todo lo que la fuente pide y cita (refs/boe.py 71-72, 298,
// 384 y 430; data-model.md §8). Las dos son https y de www.boe.es, y con ellas
// cada dirección de este fichero y la de la búsqueda: el sobre de la fuente solo
// lleva direcciones que se pueden comprobar en el BOE, nunca del espacio
// reservado kitlegal: (FR-002, FR-003).
//
// Las funciones de este fichero reciben la norma y el bloque ya validados con
// ValidarNorma y ValidarBloque, cuyas gramáticas solo admiten letras y dígitos
// ASCII y, en la norma, guiones: van tal cual como segmentos de la ruta, sin nada
// que escapar, igual que los concatena refs/boe.py.
const (
	// baseDeLaAPI es la base de la API de Legislación Consolidada, la constante
	// BOE de refs/boe.py 72: toda petición de la fuente empieza por ella.
	baseDeLaAPI = "https://www.boe.es/datosabiertos/api/legislacion-consolidada"
	// paginaPublica es act.php, la página en la que una persona lee en el BOE
	// una norma consolidada (refs/boe.py 298, 384 y 430).
	paginaPublica = "https://www.boe.es/buscar/act.php"
)

// direccionDeLaNorma es el recurso de la norma, <base>/id/<norma>: la url del
// sobre de articulos, que se apoya en varios bloques de la misma norma y no en
// uno solo (contrato verbos-y-salidas §4).
func direccionDeLaNorma(norma string) string {
	return baseDeLaAPI + "/id/" + norma
}

// direccionDelBloque es el texto de un bloque de la norma, el recurso que se pide
// en XML (refs/boe.py 406-407; contrato verbos-y-salidas §3).
func direccionDelBloque(norma, bloque string) string {
	return direccionDeLaNorma(norma) + "/texto/bloque/" + bloque
}

// direccionDelIndice es el índice de bloques de la norma (refs/boe.py 355;
// contrato verbos-y-salidas §2).
func direccionDelIndice(norma string) string {
	return direccionDeLaNorma(norma) + "/texto/indice"
}

// direccionDeLosMetadatos son los metadatos de la norma (refs/boe.py 186 y 452;
// contrato verbos-y-salidas §3 y §5): la misma dirección para metadatos y para
// los avisos de vigencia de articulo y articulos.
func direccionDeLosMetadatos(norma string) string {
	return direccionDeLaNorma(norma) + "/metadatos"
}

// direccionDelAnalisis es el análisis de la norma, con sus materias, notas y
// referencias (refs/boe.py 511; contrato verbos-y-salidas §6).
func direccionDelAnalisis(norma string) string {
	return direccionDeLaNorma(norma) + "/analisis"
}

// direccionPublicaDeLaNorma es la página act.php de una norma,
// act.php?id=<identificador> (refs/boe.py 298 y 384): la dirección pública del
// índice y la de cada resultado de búsqueda. El identificador va tal como llega,
// sin validar ni escapar y también vacío, porque el de un resultado de búsqueda
// es el que entrega la fuente (data-model.md §2.3).
func direccionPublicaDeLaNorma(identificador string) string {
	return paginaPublica + "?id=" + identificador
}

// direccionPublicaDelBloque es la página act.php de la norma abierta en el
// bloque, act.php?id=<norma>#<bloque> (refs/boe.py 430): la dirección pública de
// cada artículo (data-model.md §2.1).
func direccionPublicaDelBloque(norma, bloque string) string {
	return direccionPublicaDeLaNorma(norma) + "#" + bloque
}
