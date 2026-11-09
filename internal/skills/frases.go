package skills

import "strings"

// Lo que Frases toma por un segmento del cuerpo de un SKILL.md
// (contracts/remisiones.md §6 de H41; research D5 y D6; FR-001, FR-023).
const (
	// marcaDeEncabezado abre, tras su sangría, la línea de un encabezado.
	marcaDeEncabezado = "#"

	// aperturaDeComentario y cierreDeComentario son las marcas de una línea que
	// es entera un comentario HTML.
	aperturaDeComentario = "<!--"
	cierreDeComentario   = "-->"

	// finDeFrase es el punto que, seguido de un espacio o de un salto de línea,
	// cierra una frase.
	finDeFrase = '.'
)

// Frases son los segmentos del cuerpo de un SKILL.md, en su orden: cada línea de
// un bloque de código, de la valla que lo abre a la que lo cierra con la misma
// regla que lineasFueraDeBloques, tal cual y sin su salto; cada línea de
// encabezado —tras su sangría, empieza por «#»— y cada línea que es entera un
// comentario HTML, sin sus blancos de los extremos; y, en lo demás, cada frase,
// con sus blancos y saltos juntados en un espacio y sin los de sus extremos. Una
// frase empieza tras la anterior, una línea en blanco, un segmento de los otros
// o al principio del cuerpo, y también al principio de una línea que abre una
// viñeta o un paso (elementoDeLista), con su marca dentro; acaba en el primer
// punto seguido de un espacio o de un salto de línea, que va en ella, al final
// del cuerpo o donde empieza otra frase u otro segmento. El punto de «.md» en un
// enlace no la cierra: lo sigue una letra. Las líneas en blanco fuera de los
// bloques no dan segmento.
func Frases(cuerpo string) []string {
	var partidas particion

	for linea := range strings.Lines(cuerpo) {
		partidas.anadir(strings.TrimSuffix(linea, "\n"))
	}

	partidas.cerrarLaFrase()

	return partidas.segmentos
}

// particion son los segmentos de un cuerpo leído hasta una línea, la frase que
// sigue abierta y la valla del bloque de código abierto, o nada fuera de uno.
type particion struct {
	segmentos []string
	frase     strings.Builder
	abierta   string
}

// anadir parte la línea siguiente del cuerpo, sin su salto.
func (p *particion) anadir(linea string) {
	valla, resto := vallaDeLaLinea(linea)

	switch {
	case p.abierta == "" && valla != "":
		p.cerrarLaFrase()

		p.abierta = valla
		p.segmentos = append(p.segmentos, linea)
	case p.abierta != "":
		if valla != "" && valla[0] == p.abierta[0] && len(valla) >= len(p.abierta) && strings.TrimSpace(resto) == "" {
			p.abierta = ""
		}

		p.segmentos = append(p.segmentos, linea)
	case esUnSegmentoSolo(linea):
		p.cerrarLaFrase()

		p.segmentos = append(p.segmentos, strings.TrimSpace(linea))
	case strings.TrimSpace(linea) == "":
		p.cerrarLaFrase()
	default:
		p.anadirAFrases(linea)
	}
}

// anadirAFrases parte en frases una línea que no es de un bloque de código, ni un
// segmento solo, ni está en blanco: la que abre una viñeta o un paso cierra la
// frase abierta, y cada punto que cierra una frase la da.
func (p *particion) anadirAFrases(linea string) {
	desde := 0

	if elemento := elementoDeLista.FindStringSubmatchIndex(linea); elemento != nil {
		p.cerrarLaFrase()

		desde = elemento[2]
	}

	inicio := 0

	for _, final := range finalesDeFrase(linea, desde) {
		p.frase.WriteString(linea[inicio:final])
		p.cerrarLaFrase()

		inicio = final
	}

	p.frase.WriteString(linea[inicio:])
	p.frase.WriteString("\n")
}

// cerrarLaFrase da la frase abierta, con sus blancos juntados en un espacio y sin
// los de sus extremos, si le queda algo, y abre otra.
func (p *particion) cerrarLaFrase() {
	if juntada := strings.Join(strings.Fields(p.frase.String()), " "); juntada != "" {
		p.segmentos = append(p.segmentos, juntada)
	}

	p.frase.Reset()
}

// esUnSegmentoSolo dice si la línea, fuera de un bloque de código, es un
// segmento por sí misma: un encabezado o, sin sus blancos de los extremos, un
// comentario HTML entero.
func esUnSegmentoSolo(linea string) bool {
	recortada := strings.TrimSpace(linea)

	return strings.HasPrefix(strings.TrimLeft(linea, " \t"), marcaDeEncabezado) ||
		(strings.HasPrefix(recortada, aperturaDeComentario) && strings.HasSuffix(recortada, cierreDeComentario))
}

// finalesDeFrase son las posiciones de la línea, a partir de la dada, en que
// acaba una frase, cada una tras el punto que la cierra: un punto seguido de un
// espacio o del final de la línea, que es su salto o el final del cuerpo. Antes
// de desde está la marca de una viñeta o de un paso, cuyo punto no la cierra.
func finalesDeFrase(linea string, desde int) []int {
	var finales []int

	for posicion := desde; posicion < len(linea); posicion++ {
		if linea[posicion] == finDeFrase && (posicion+1 == len(linea) || linea[posicion+1] == ' ') {
			finales = append(finales, posicion+1)
		}
	}

	return finales
}
