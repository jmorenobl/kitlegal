package boe

// Los códigos de aviso, el enumerado de Aviso.Codigo: uno por condición de
// _check_vigencia, en kebab-case y en el orden de las condiciones (FR-012, Q5).
const (
	codigoDeConsolidacionNoFinalizada = "consolidacion-no-finalizada"
	codigoDeNormaDerogada             = "derogada"
	codigoDeVigenciaAgotada           = "vigencia-agotada"
)

// Las frases de los avisos, carácter a carácter las de _check_vigencia
// (refs/boe.py 202-211), con su prefijo. No son las de cmd_metadatos
// (485-498), que redacta a su manera: cada código tiene una sola frase en
// articulo, articulos y metadatos (FR-050; entrada 17 del porte anotado en
// doc.go).
const (
	fraseDeConsolidacionNoFinalizada = "⚠ TEXTO POSIBLEMENTE DESACTUALIZADO: la consolidación de esta norma " +
		"no está finalizada. Puede haber modificaciones recientes aún no integradas."
	fraseDeNormaDerogada   = "⚠ NORMA DEROGADA: esta norma ha sido derogada."
	fraseDeVigenciaAgotada = "⚠ VIGENCIA AGOTADA: esta norma ya no está en vigor."
)

// Lo que miran las condiciones en el objeto de metadatos (refs/boe.py 199-211).
const (
	// claveDeEstadoDeConsolidacion y claveDeCodigoDeConsolidacion son el objeto
	// del estado de la consolidación y su código (refs/boe.py 199-200).
	claveDeEstadoDeConsolidacion = "estado_consolidacion"
	claveDeCodigoDeConsolidacion = "codigo"
	// consolidacionNoFinalizada es el código de ese estado con el que avisa
	// (refs/boe.py 201).
	consolidacionNoFinalizada = "4"
	// claveDeEstatusDeDerogacion y claveDeVigenciaAgotada son los dos campos que
	// avisan con afirmativo (refs/boe.py 207 y 210).
	claveDeEstatusDeDerogacion = "estatus_derogacion"
	claveDeVigenciaAgotada     = "vigencia_agotada"
	afirmativo                 = "S"
)

// CodigosDeAviso son los valores del enumerado codigo de Aviso, en el orden de
// sus condiciones: consolidacion-no-finalizada, derogada y vigencia-agotada
// (FR-012). Cada llamada devuelve una lista nueva, que quien la recibe puede
// cambiar sin cambiar la de nadie más.
func CodigosDeAviso() []string {
	return []string{codigoDeConsolidacionNoFinalizada, codigoDeNormaDerogada, codigoDeVigenciaAgotada}
}

// avisosDe son los avisos de vigencia que se derivan del objeto de metadatos de
// una norma, con las tres condiciones de _check_vigencia y en su orden
// (refs/boe.py 197-211; FR-012, FR-050, data-model.md §2.2):
//
//  1. estado_consolidacion es un objeto y su codigo es la cadena "4" →
//     consolidacion-no-finalizada
//  2. estatus_derogacion es la cadena "S" → derogada
//  3. vigencia_agotada es la cadena "S" → vigencia-agotada
//
// Las comparaciones son de cadenas exactas, como el == de Python: un código
// numérico, otra caja o un espacio no avisan. Si no se cumple ninguna, la lista
// vacía y nunca nula, que se emite como [].
//
// Recibe el objeto que da primerElemento; lo que refs/boe.py hace cuando los
// metadatos no se obtienen o no se interpretan —devolver None y presentar el
// artículo sin avisos— no llega aquí: es un fallo con su clase (entrada 10 del
// porte anotado en doc.go).
func avisosDe(metadatos map[string]any) []Aviso {
	avisos := []Aviso{}

	if consolidacionSinFinalizar(metadatos) {
		avisos = append(avisos, Aviso{Codigo: codigoDeConsolidacionNoFinalizada, Texto: fraseDeConsolidacionNoFinalizada})
	}

	if esLaCadena(metadatos[claveDeEstatusDeDerogacion], afirmativo) {
		avisos = append(avisos, Aviso{Codigo: codigoDeNormaDerogada, Texto: fraseDeNormaDerogada})
	}

	if esLaCadena(metadatos[claveDeVigenciaAgotada], afirmativo) {
		avisos = append(avisos, Aviso{Codigo: codigoDeVigenciaAgotada, Texto: fraseDeVigenciaAgotada})
	}

	return avisos
}

// consolidacionSinFinalizar es la primera condición: estado_consolidacion es un
// objeto y su codigo es la cadena "4". Lo que no es un objeto da el código vacío
// de refs/boe.py 200 y no avisa.
func consolidacionSinFinalizar(metadatos map[string]any) bool {
	estado, esObjeto := metadatos[claveDeEstadoDeConsolidacion].(map[string]any)

	return esObjeto && esLaCadena(estado[claveDeCodigoDeConsolidacion], consolidacionNoFinalizada)
}

// esLaCadena dice si el valor de la respuesta es exactamente esa cadena: un
// json.Number, un booleano o cualquier otro tipo no lo es aunque su literal
// coincida.
func esLaCadena(valor any, esperada string) bool {
	cadena, esCadena := valor.(string)

	return esCadena && cadena == esperada
}
