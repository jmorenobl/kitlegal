package skills_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// TestEnlacesDeMarkdown fija lo que los gates de references/ toman por un enlace
// (contracts/gates-de-skills-check.md §2.3; research D14; FR-007, FR-008): un
// [texto](destino) en una línea fuera de los bloques de código, con el destino
// hasta el primer ) o espacio y sin lo que sigue a #; el que solo lleva a un
// apartado no da ninguno. Un bloque va de la línea que abre una valla —hasta
// tres espacios y tres o más acentos graves o virgulillas— a la que la cierra
// con la misma marca, o al final. Y, resuelto desde references/ con path.Join,
// el destino de un enlace de una referencia es la ruta del fichero al que
// lleva.
func TestEnlacesDeMarkdown(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		contenido string
		enlaces   []string
	}{
		{
			nombre: "enlace", contenido: "- [como_se_cita](references/como_se_cita.md): para otra forma.\n",
			enlaces: []string{"references/como_se_cita.md"},
		},
		{nombre: "sin-texto", contenido: "[](references/como_se_cita.md)\n", enlaces: []string{"references/como_se_cita.md"}},
		{
			nombre: "con-apartado", contenido: "[ver](references/como_se_cita.md#los-avisos)\n",
			enlaces: []string{"references/como_se_cita.md"},
		},
		{nombre: "solo-un-apartado", contenido: "[Los avisos](#los-avisos)\n"},
		{
			nombre: "con-titulo", contenido: "[ver](references/como_se_cita.md \"Cómo se cita\")\n",
			enlaces: []string{"references/como_se_cita.md"},
		},
		{
			nombre: "varios-en-una-linea", contenido: "Lee [a](a.md) y [b](b.md#x), y [c](c.md).\n",
			enlaces: []string{"a.md", "b.md", "c.md"},
		},
		{nombre: "en-codigo-en-linea-cuenta", contenido: "`[a](a.md)`\n", enlaces: []string{"a.md"}},
		{nombre: "ruta-sin-enlace", contenido: "`references/como_se_cita.md`\n"},
		{nombre: "espacio-antes-del-destino", contenido: "[a] (a.md)\n"},
		{nombre: "destino-sin-cerrar", contenido: "[a](a.md\n"},
		{nombre: "sin-salto-final", contenido: "[a](a.md)", enlaces: []string{"a.md"}},
		{
			nombre: "fin-de-linea-crlf", contenido: "[a](a.md)\r\n```\r\n[b](b.md)\r\n```\r\n[c](c.md)\r\n",
			enlaces: []string{"a.md", "c.md"},
		},
		{nombre: "dentro-de-un-bloque", contenido: "```text\n[a](references/como_se_cita.md)\n```\n"},
		{
			nombre: "dentro-de-un-bloque-de-virgulillas", contenido: "~~~\n[a](a.md)\n~~~\n[b](b.md)\n",
			enlaces: []string{"b.md"},
		},
		{
			nombre: "valla-con-tres-espacios", contenido: "   ```\n[a](a.md)\n   ```\n[b](b.md)\n",
			enlaces: []string{"b.md"},
		},
		{nombre: "cuatro-espacios-no-son-valla", contenido: "    ```\n[a](a.md)\n", enlaces: []string{"a.md"}},
		{nombre: "dos-acentos-no-son-valla", contenido: "``\n[a](a.md)\n", enlaces: []string{"a.md"}},
		{nombre: "bloque-sin-cerrar", contenido: "[a](a.md)\n```\n[b](b.md)\n", enlaces: []string{"a.md"}},
		{nombre: "otra-marca-no-cierra", contenido: "```\n~~~\n[a](a.md)\n```\n[b](b.md)\n", enlaces: []string{"b.md"}},
		{
			nombre: "valla-mas-corta-no-cierra", contenido: "````\n```\n[a](a.md)\n````\n[b](b.md)\n",
			enlaces: []string{"b.md"},
		},
		{nombre: "valla-mas-larga-cierra", contenido: "```\n[a](a.md)\n`````\n[b](b.md)\n", enlaces: []string{"b.md"}},
		{
			nombre: "valla-con-texto-no-cierra", contenido: "```\n```text\n[a](a.md)\n```\n[b](b.md)\n",
			enlaces: []string{"b.md"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.enlaces, skills.EnlacesDeMarkdown([]byte(caso.contenido)))
		})
	}

	t.Run("resolucion-desde-references", func(t *testing.T) {
		t.Parallel()

		contenido := "[a](varios_bloques.md), [b](./varios_bloques.md), [c](../references/varios_bloques.md),\n" +
			"[d](como_se_cita.md#x), [e](#apartado) y [f](../SKILL.md).\n" +
			"```\n[g](dentro.md)\n```\n"

		assert.Equal(t, []string{
			"references/varios_bloques.md",
			"references/varios_bloques.md",
			"references/varios_bloques.md",
			"references/como_se_cita.md",
			"SKILL.md",
		}, skills.FicherosEnlazados([]byte(contenido), "references"))
	})
}

// TestIndiceDeLaReferencia fija la regla del índice de FR-009
// (contracts/gates-de-skills-check.md §2.3; research D14; Clarificación Q5):
// con 100 líneas o menos no se exige nada; con más, los apartados son las líneas
// fuera de los bloques de código que empiezan por «## », tiene que haber alguno,
// y antes del primero una lista —líneas que empiezan, tras su sangría, por «- »,
// «* » o un número y «. »— en la que, por orden, cada apartado tenga un elemento
// que contenga su texto, los dos sin «[», «](…)», «*», «_» ni «`». Los elementos
// de más no estorban. Lo que falta lo dice con las líneas y el límite.
func TestIndiceDeLaReferencia(t *testing.T) {
	t.Parallel()

	const (
		sinIndice    = "%d líneas sin un índice que nombre sus apartados ## antes del primero (más de 100)"
		sinApartados = "%d líneas sin ningún apartado ## (más de 100)"
	)

	apartados := []string{"## Lo que hace cita", "## Los avisos"}
	indice := []string{"- [Lo que hace cita](#lo-que-hace-cita)", "- [Los avisos](#los-avisos)"}

	casos := []struct {
		nombre string

		// cabeza son las líneas que van antes del relleno, y cola, las de
		// después; referenciaDe las completa hasta las líneas dadas.
		cabeza, cola []string
		lineas       int

		falta string
	}{
		{nombre: "cien-lineas-sin-indice", cabeza: apartados, lineas: 100},
		{nombre: "cien-lineas-sin-apartados", lineas: 100},
		{nombre: "ciento-una-lineas-con-indice", cabeza: slices.Concat(indice, apartados), lineas: 101},
		{nombre: "ciento-una-lineas-sin-indice", cabeza: apartados, lineas: 101, falta: fmt.Sprintf(sinIndice, 101)},
		{nombre: "ciento-una-lineas-sin-apartados", lineas: 101, falta: fmt.Sprintf(sinApartados, 101)},
		{nombre: "sin-apartados-con-una-lista", cabeza: indice, lineas: 101, falta: fmt.Sprintf(sinApartados, 101)},
		{
			nombre: "apartados-solo-en-un-bloque",
			cabeza: slices.Concat([]string{"```markdown"}, apartados, []string{"```"}),
			lineas: 101,
			falta:  fmt.Sprintf(sinApartados, 101),
		},
		{
			nombre: "apartado-de-nivel-tres-no-cuenta",
			cabeza: []string{"### Lo que hace cita"},
			lineas: 101,
			falta:  fmt.Sprintf(sinApartados, 101),
		},
		{
			nombre: "indice-en-otro-orden",
			cabeza: slices.Concat([]string{indice[1], indice[0]}, apartados),
			lineas: 101,
			falta:  fmt.Sprintf(sinIndice, 101),
		},
		{
			nombre: "indice-sin-un-apartado",
			cabeza: slices.Concat(indice[:1], apartados),
			lineas: 101,
			falta:  fmt.Sprintf(sinIndice, 101),
		},
		{
			nombre: "indice-detras-del-primer-apartado",
			cabeza: slices.Concat(apartados[:1], indice, apartados[1:]),
			lineas: 101,
			falta:  fmt.Sprintf(sinIndice, 101),
		},
		{
			nombre: "indice-dentro-de-un-bloque",
			cabeza: slices.Concat([]string{"```text"}, indice, []string{"```"}, apartados),
			lineas: 101,
			falta:  fmt.Sprintf(sinIndice, 101),
		},
		{
			nombre: "indice-que-no-es-una-lista",
			cabeza: slices.Concat([]string{"Lo que hace cita. Los avisos."}, apartados),
			lineas: 101,
			falta:  fmt.Sprintf(sinIndice, 101),
		},
		{
			// Un apartado de más en el fichero, sin elemento.
			nombre: "apartado-sin-elemento-al-final",
			cabeza: slices.Concat(indice, apartados),
			cola:   []string{"## Las fuentes"},
			lineas: 101,
			falta:  fmt.Sprintf(sinIndice, 101),
		},
		{
			nombre: "elementos-de-mas-y-sangrados",
			cabeza: slices.Concat([]string{
				"* Lo que hace cita",
				"  - Una subsección",
				"  1. Los avisos",
				"10. Un elemento de más",
			}, apartados),
			lineas: 101,
		},
		{
			nombre: "sin-la-sintaxis-de-enlace-ni-el-enfasis",
			cabeza: []string{
				"- [**Lo que hace** `cita`](#lo-que-hace-cita)",
				"- _Los avisos_, con su código",
				"## Lo que hace `cita`",
				"## Los *avisos*",
			},
			lineas: 101,
		},
		{
			nombre: "doscientas-lineas-sin-indice",
			cabeza: apartados,
			lineas: 200,
			falta:  fmt.Sprintf(sinIndice, 200),
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			referencia := referenciaDe(t, caso.cabeza, caso.cola, caso.lineas)
			assert.Equal(t, caso.falta, skills.FaltaDelIndice([]byte(referencia)))
		})
	}
}

// referenciaDe es una referencia de prueba con las líneas dadas: el título, una
// línea en blanco, las de cabeza, las de relleno que falten y las de cola, cada
// una con su salto de línea.
func referenciaDe(t *testing.T, cabeza, cola []string, lineas int) string {
	t.Helper()

	fijas := slices.Concat([]string{"# Cómo se cita", ""}, cabeza)
	relleno := lineas - len(fijas) - len(cola)
	require.GreaterOrEqual(t, relleno, 0, "las líneas fijas caben en %d", lineas)

	todas := slices.Concat(fijas, make([]string, relleno), cola)
	for indice := len(fijas); indice < len(fijas)+relleno; indice++ {
		todas[indice] = "Una línea de relleno del protocolo de detalle."
	}

	referencia := strings.Join(todas, "\n") + "\n"
	require.Equal(t, lineas, skills.ContarLineas([]byte(referencia)))

	return referencia
}
