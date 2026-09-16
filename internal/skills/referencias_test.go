package skills_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// comienzoDeLaReferencia es lo que abre toda referencia de normas, hasta la fila
// de separación de su tabla (contrato normas-y-referencias §5).
const comienzoDeLaReferencia = "<!-- generado desde data/normas.yaml, no editar -->\n" +
	"\n" +
	"# Normas de referencia\n" +
	"\n" +
	"| Norma | Abreviatura | Identificador | Rango | Materias |\n" +
	"|---|---|---|---|---|\n"

// referenciaDeLasNormasDePrueba son los bytes exactos de la referencia de
// normasDePrueba: por año y número, los dos por su valor; la barra escrita \|;
// cada salto de línea, \n, \r\n o \r, como un espacio; y la abreviatura ausente
// como una celda vacía.
const referenciaDeLasNormasDePrueba = comienzoDeLaReferencia +
	"| Ley 7/1985, de 2 de abril, reguladora de las Bases del Régimen Local. | LRBRL | `BOE-A-1985-5392` | Ley | régimen local |\n" +
	"| Norma de prueba con un salto de línea. | NP\\|9999 | `BOE-A-2015-9999` | Real Decreto | primera materia, segunda \\| materia |\n" +
	"| Norma de prueba con saltos de línea de Windows y de retorno de carro. | NP | `BOE-A-2015-10565` | Real Decreto-ley | procedimiento administrativo, tributos |\n" +
	"| Norma de prueba con una barra \\| en el título. |  | `BOE-A-2015-10566` | Ley | régimen jurídico del sector público |\n"

// normasDePrueba son las normas de TestRenderizarNormas en un orden que no es el
// de la referencia: 9999 va antes que 10565 por su valor, aunque como texto irían
// al revés, y 10565 antes que 10566.
func normasDePrueba() []skills.Norma {
	return []skills.Norma{
		{
			Identificador: "BOE-A-2015-10566",
			Titulo:        "Norma de prueba con una barra | en el título.",
			Rango:         "Ley",
			Materias:      []string{"régimen jurídico del sector público"},
		},
		{
			Identificador: "BOE-A-2015-9999",
			Titulo:        "Norma de prueba con un salto\nde línea.",
			Rango:         "Real Decreto",
			Abreviatura:   "NP|9999",
			Materias:      []string{"primera materia", "segunda | materia"},
		},
		{
			Identificador: "BOE-A-1985-5392",
			Titulo:        "Ley 7/1985, de 2 de abril, reguladora de las Bases del Régimen Local.",
			Rango:         "Ley",
			Abreviatura:   "LRBRL",
			Materias:      []string{"régimen local"},
		},
		{
			Identificador: "BOE-A-2015-10565",
			Titulo:        "Norma de prueba con saltos\r\nde línea de Windows\ry de retorno de carro.",
			Rango:         "Real Decreto-ley",
			Abreviatura:   "NP",
			Materias:      []string{"procedimiento\nadministrativo", "tributos"},
		},
	}
}

// TestRenderizarNormas fija RenderizarNormas (contrato normas-y-referencias §5;
// data-model §4.2; FR-030, FR-031, FR-034): los bytes exactos de la referencia,
// con la cabecera que nombra data/normas.yaml, las filas por año y número del
// identificador por su valor y los tres casos de escape de una celda —la barra,
// el salto de línea y la abreviatura ausente—; y la misma salida en dos llamadas
// y con las normas en cualquier orden.
func TestRenderizarNormas(t *testing.T) {
	t.Parallel()

	t.Run("bytes-exactos", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, referenciaDeLasNormasDePrueba, string(skills.RenderizarNormas(normasDePrueba())))
	})

	t.Run("dos-llamadas-dan-lo-mismo", func(t *testing.T) {
		t.Parallel()

		normas := normasDePrueba()

		primera := skills.RenderizarNormas(normas)
		segunda := skills.RenderizarNormas(normas)

		assert.Equal(t, primera, segunda)
		assert.Equal(t, normasDePrueba(), normas, "las normas que recibe no se reordenan")
	})

	t.Run("orden-sin-depender-de-la-entrada", func(t *testing.T) {
		t.Parallel()

		normas := normasDePrueba()
		slices.Reverse(normas)

		assert.Equal(t, referenciaDeLasNormasDePrueba, string(skills.RenderizarNormas(normas)))
	})

	t.Run("ceros-a-la-izquierda", func(t *testing.T) {
		t.Parallel()

		// 010565 vale lo mismo que 10565 y va antes que 10566, aunque como número
		// de cifras sería mayor; a igual valor decide el identificador tal cual.
		normas := []skills.Norma{
			{Identificador: "BOE-A-2015-10566", Titulo: "Mayor.", Rango: "Ley", Materias: []string{"a"}},
			{Identificador: "BOE-A-2015-010565", Titulo: "Igual, con un cero.", Rango: "Ley", Materias: []string{"a"}},
			{Identificador: "BOE-A-2015-10565", Titulo: "Igual, sin ceros.", Rango: "Ley", Materias: []string{"a"}},
		}
		esperada := comienzoDeLaReferencia +
			"| Igual, con un cero. |  | `BOE-A-2015-010565` | Ley | a |\n" +
			"| Igual, sin ceros. |  | `BOE-A-2015-10565` | Ley | a |\n" +
			"| Mayor. |  | `BOE-A-2015-10566` | Ley | a |\n"

		assert.Equal(t, esperada, string(skills.RenderizarNormas(normas)))

		slices.Reverse(normas)
		assert.Equal(t, esperada, string(skills.RenderizarNormas(normas)))
	})
}
