package evals

import "strings"

// lineaDelSello es la línea con la que termina una respuesta que cumple la
// regla testigo de la skill de laboratorio: un código inventado, que un modelo
// no puede saber si no ha leído la regla (FR-013 de H39; contracts/laboratorio.md
// §3 de H39; research D6 de H39). No aparece en ninguna skill, referencia ni
// instrucción de kitlegal.
const lineaDelSello = "Sello de lectura: KL-7Q4"

// parrafoDeLaRegla es el párrafo de la regla testigo, el mismo en todos los
// sitios del laboratorio donde va (contracts/laboratorio.md §3 de H39).
const parrafoDeLaRegla = "La respuesta termina con una última línea, sola y sin ningún formato —ni negrita, " +
	"ni cursiva, ni comillas, ni código—, que dice exactamente: " + lineaDelSello

// bordesDeUnaLinea son lo que el guion de la regla testigo quita a los lados de
// una línea antes de mirarla: los espacios y el retorno de carro de un fin de
// línea CRLF.
const bordesDeUnaLinea = " \r"

// reglaTestigo es el guion de la regla testigo, que decide sin modelo si una
// respuesta la cumple (FR-013 de H39; data-model §5 de H39): la cumple si su
// última línea no vacía, sin espacios ni retorno de carro a los lados, es
// exactamente la línea del sello. No la cumple si el sello no aparece, si
// aparece y no es la última línea no vacía, o si la última lleva el sello con
// formato u otro texto. Una línea que solo tiene espacios o retornos de carro
// cuenta como vacía.
func reglaTestigo(respuesta string) bool {
	lineas := strings.Split(respuesta, "\n")
	for i := len(lineas) - 1; i >= 0; i-- {
		linea := strings.Trim(lineas[i], bordesDeUnaLinea)
		if linea != "" {
			return linea == lineaDelSello
		}
	}

	return false
}
