package skills_test

import (
	"slices"
	"strings"
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

// comienzoDeLasLeyesVertebrales es lo que abre toda referencia de leyes
// vertebrales, hasta la fila de separación de su tabla (contrato de la skill
// legal-core §2): la cabecera nombra data/normas.yaml, del que sale aunque no se
// llame como ella, y los encabezados son los de la referencia de normas.
const comienzoDeLasLeyesVertebrales = "<!-- generado desde data/normas.yaml, no editar -->\n" +
	"\n" +
	"# Leyes vertebrales\n" +
	"\n" +
	"| Norma | Abreviatura | Identificador | Rango | Materias |\n" +
	"|---|---|---|---|---|\n"

// normasDePruebaConDosVertebrales son normasDePrueba con la marca vertebral en la
// de la barra en el título, BOE-A-2015-10566, y en la LRBRL, BOE-A-1985-5392.
func normasDePruebaConDosVertebrales() []skills.Norma {
	normas := normasDePrueba()
	normas[0].Vertebral = true
	normas[2].Vertebral = true

	return normas
}

// TestRenderizarLeyesVertebrales fija RenderizarLeyesVertebrales (contrato de la
// skill legal-core §2 y §4; research.md D20; FR-065, FR-067): solo las normas
// marcadas vertebral, con la cabecera que nombra data/normas.yaml, su propio
// título y los encabezados, el orden y los escapes de la referencia de normas; sin
// ninguna marcada, la tabla vacía; y las normas que recibe no cambian.
func TestRenderizarLeyesVertebrales(t *testing.T) {
	t.Parallel()

	t.Run("solo-las-marcadas", func(t *testing.T) {
		t.Parallel()

		esperada := comienzoDeLasLeyesVertebrales +
			"| Ley 7/1985, de 2 de abril, reguladora de las Bases del Régimen Local. | LRBRL | `BOE-A-1985-5392` | Ley | régimen local |\n" +
			"| Norma de prueba con una barra \\| en el título. |  | `BOE-A-2015-10566` | Ley | régimen jurídico del sector público |\n"

		assert.Equal(t, esperada, string(skills.RenderizarLeyesVertebrales(normasDePruebaConDosVertebrales())))
	})

	t.Run("todas-marcadas-dan-las-filas-de-las-normas", func(t *testing.T) {
		t.Parallel()

		normas := normasDePrueba()
		for indice := range normas {
			normas[indice].Vertebral = true
		}

		filas := strings.TrimPrefix(referenciaDeLasNormasDePrueba, comienzoDeLaReferencia)
		assert.Equal(t, comienzoDeLasLeyesVertebrales+filas, string(skills.RenderizarLeyesVertebrales(normas)))
	})

	t.Run("ninguna-marcada", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, comienzoDeLasLeyesVertebrales, string(skills.RenderizarLeyesVertebrales(normasDePrueba())))
	})

	t.Run("no-cambia-las-normas-que-recibe", func(t *testing.T) {
		t.Parallel()

		normas := normasDePruebaConDosVertebrales()
		skills.RenderizarLeyesVertebrales(normas)

		assert.Equal(t, normasDePruebaConDosVertebrales(), normas)
	})
}

// referenciaDeLaJerarquiaConEscapes son los bytes exactos de la referencia de
// jerarquiaConEscapes: una fila por nivel en el orden en que llegan, con los tipos
// de norma separados por «, », la barra escrita \| y cada salto de línea como un
// espacio; y detrás la lista de las reglas, cada una con su código y su
// enunciado en una sola línea.
const referenciaDeLaJerarquiaConEscapes = "<!-- generado desde data/jerarquia.yaml, no editar -->\n" +
	"\n" +
	"# Jerarquía normativa\n" +
	"\n" +
	"| Nivel | Boletín | Tipos de norma, de mayor a menor rango |\n" +
	"|---|---|---|\n" +
	"| Unión Europea | Diario de prueba \\| con una barra | Reglamento, Directiva |\n" +
	"| Estado | Boletín de prueba en dos líneas | Ley \\| con una barra |\n" +
	"\n" +
	"## Reglas de interpretación\n" +
	"\n" +
	"- `ley-posterior`: La posterior deroga a la anterior, con saltos de Windows y de retorno de carro.\n" +
	"- `ley-especial`: La especial | prevalece sobre la general.\n"

// jerarquiaConEscapes es la jerarquía de TestRenderizarJerarquia: dos niveles y dos
// reglas, con una barra y saltos de línea de las tres formas en sus textos.
func jerarquiaConEscapes() skills.Jerarquia {
	return skills.Jerarquia{
		Niveles: []skills.NivelNormativo{
			{
				Codigo:  "ue",
				Nombre:  "Unión Europea",
				Boletin: "Diario de prueba | con una barra",
				Normas:  []string{"Reglamento", "Directiva"},
			},
			{Codigo: "estado", Nombre: "Estado", Boletin: "Boletín de prueba\nen dos líneas", Normas: []string{"Ley | con una barra"}},
		},
		Reglas: []skills.ReglaDeInterpretacion{
			{
				Codigo:    "ley-posterior",
				Enunciado: "La posterior deroga a la anterior,\r\ncon saltos de Windows\ry de retorno de carro.",
			},
			{Codigo: "ley-especial", Enunciado: "La especial | prevalece\nsobre la general."},
		},
	}
}

// TestRenderizarJerarquia fija RenderizarJerarquia (contrato de la skill
// legal-core §2 y §4; research.md D20; FR-066, FR-067): los bytes exactos de la
// referencia, con la cabecera que nombra data/jerarquia.yaml, la tabla de niveles
// con su boletín y sus tipos de norma, y la lista de reglas; los niveles y las
// reglas van en el orden en que llegan, que es el que el esquema impone al
// documento; y la jerarquía que recibe no cambia.
func TestRenderizarJerarquia(t *testing.T) {
	t.Parallel()

	t.Run("bytes-exactos", func(t *testing.T) {
		t.Parallel()

		jerarquia := jerarquiaConEscapes()

		assert.Equal(t, referenciaDeLaJerarquiaConEscapes, string(skills.RenderizarJerarquia(jerarquia)))
		assert.Equal(t, jerarquiaConEscapes(), jerarquia, "la jerarquía que recibe no cambia")
	})

	t.Run("en-el-orden-en-que-llegan", func(t *testing.T) {
		t.Parallel()

		jerarquia := jerarquiaConEscapes()
		slices.Reverse(jerarquia.Niveles)
		slices.Reverse(jerarquia.Reglas)

		referencia := string(skills.RenderizarJerarquia(jerarquia))

		assert.Less(t, strings.Index(referencia, "| Estado |"), strings.Index(referencia, "| Unión Europea |"))
		assert.Less(t, strings.Index(referencia, "`ley-especial`"), strings.Index(referencia, "`ley-posterior`"))
	})
}
