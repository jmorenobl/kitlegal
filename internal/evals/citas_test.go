package evals

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExtraerCitas fija la extracción de la parte mecánica de una cita, los
// corchetes que terminan en <identificador>, bloque <id de bloque>], con la
// expresión del contrato skill-boe-legislacion §3: dentro de la redacción que la
// rodea, con el id del bloque delimitado por el corchete de cierre aunque termine
// en punto, con la forma legible delante del identificador dentro de los
// corchetes —las citas reales de las pruebas de red de T030, una por intento—,
// dentro de otro corchete, varias en una línea en el orden en que aparecen, y
// ninguna sin los dos corchetes en la misma línea, sin «bloque», con texto detrás
// del id o con el identificador pegado a una letra o una cifra (FR-008, FR-072;
// US4, escenario 3; research D23).
func TestExtraerCitas(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		respuesta string
		citas     []Cita
	}{
		{
			nombre: "forma-fija-con-redaccion",
			respuesta: "Según el art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21], la Administración " +
				"está obligada a dictar resolución expresa.",
			citas: []Cita{{Norma: normaDeLasTrazas, Bloque: "a21"}},
		},
		{
			// La sesión 10 del intento 4 de T030: el nombre de la norma.
			nombre:    "forma-legible-dentro-con-el-nombre-de-la-norma",
			respuesta: "Son treinta días naturales [Real Decreto Legislativo 2/2015, BOE-A-2015-11430, bloque a38].",
			citas:     []Cita{{Norma: "BOE-A-2015-11430", Bloque: "a38"}},
		},
		{
			// La sesión 09 del intento 5 de T030: el nombre de la norma, sola en su
			// línea tras la transcripción del artículo.
			nombre: "forma-legible-dentro-sola-tras-una-cita-textual",
			respuesta: "> **Artículo 140**\n>\n> La Constitución garantiza la autonomía de los municipios.\n\n" +
				"[Constitución Española, BOE-A-1978-31229, bloque a140]\n",
			citas: []Cita{{Norma: "BOE-A-1978-31229", Bloque: "a140"}},
		},
		{
			// La sesión 08 del intento 6 de T030: el artículo, el apartado y la sigla
			// de la norma, en sus tres citas.
			nombre: "forma-legible-dentro-con-articulo-apartado-y-sigla",
			respuesta: "Se resuelve en un mes [art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20], ampliable " +
				"por otro [art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20]; sin resolución, se entiende " +
				"desestimada [art. 20.4 de la LTAIBG, BOE-A-2013-12887, bloque a20].",
			citas: []Cita{
				{Norma: "BOE-A-2013-12887", Bloque: "a20"},
				{Norma: "BOE-A-2013-12887", Bloque: "a20"},
				{Norma: "BOE-A-2013-12887", Bloque: "a20"},
			},
		},
		{
			// La sesión 02 del intento 6 de T030: la forma legible en un corchete que
			// envuelve la forma fija es una sola cita.
			nombre:    "dentro-de-otro-corchete",
			respuesta: "El expediente lleva el informe del órgano [art. 118.2, LCSP [BOE-A-2017-12902, bloque a1-30]].",
			citas:     []Cita{{Norma: "BOE-A-2017-12902", Bloque: "a1-30"}},
		},
		{
			nombre: "forma-legible-dentro-sin-bloque",
			respuesta: "Sin la palabra ni el id: [art. 20.1 de la LTAIBG, BOE-A-2013-12887], " +
				"[Constitución Española, BOE-A-1978-31229, a140] y [Real Decreto Legislativo 2/2015, BOE-A-2015-11430 a38].",
		},
		{
			nombre: "forma-legible-sin-corchetes",
			respuesta: "Sin corchetes: art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20.\n" +
				"Sin el de cierre: [art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20\n" +
				"Sin el de abrir: art. 20.4 de la LTAIBG, BOE-A-2013-12887, bloque a20]\n" +
				"Sin el de abrir, tras otros corchetes: [nota] art. 20.4 de la LTAIBG, BOE-A-2013-12887, bloque a20]",
		},
		{
			nombre:    "texto-detras-del-id",
			respuesta: "Detrás del id: [BOE-A-2015-10565, bloque a21, art. 21] y [Ley 39/2015, BOE-A-2015-10565, bloque a21 (LPAC)].",
		},
		{
			nombre:    "identificador-pegado",
			respuesta: "Pegado a una letra o a una cifra: [LPACBOE-A-2015-10565, bloque a21] y [Ley 39/2015BOE-A-2015-10565, bloque a21].",
		},
		{
			nombre:    "bloque-que-termina-en-punto",
			respuesta: "El art. 85 bis de la Ley 7/1985 se cita [BOE-A-1985-5392, bloque a85bis.] con el punto de su id.",
			citas:     []Cita{{Norma: "BOE-A-1985-5392", Bloque: "a85bis."}},
		},
		{
			nombre: "sin-corchete",
			respuesta: "Sin corchetes: BOE-A-2015-10565, bloque a21.\n" +
				"Sin el de cierre: [BOE-A-2015-10565, bloque a22\n" +
				"Sin el de abrir: BOE-A-2015-10565, bloque a23]",
		},
		{
			nombre:    "sin-bloque",
			respuesta: "Sin la palabra: [BOE-A-2015-10565, a21], [BOE-A-2015-10565 a22] y [BOE-A-2015-10565, art. a23].",
		},
		{
			nombre: "varias-en-una-linea",
			respuesta: "El art. 21 [BOE-A-2015-10565, bloque a21] y el art. 22 [BOE-A-2015-10565, bloque a22], " +
				"y el expediente del contrato menor [BOE-A-2017-12902, bloque a1-30].",
			citas: []Cita{
				{Norma: normaDeLasTrazas, Bloque: "a21"},
				{Norma: normaDeLasTrazas, Bloque: "a22"},
				{Norma: "BOE-A-2017-12902", Bloque: "a1-30"},
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.citas, ExtraerCitas(caso.respuesta))
		})
	}
}

// TestExtraerSinConsulta fija el reconocimiento de la línea ⚠ SIN CONSULTA AL
// BOE: de una respuesta (contracts/skills.md §3 y contracts/evals-en-dos-modos.md
// §4 de H21; data-model §9; FR-035, FR-047): la línea que empieza por la marca,
// la etiqueta y los dos puntos, con la tolerancia de las etiquetas de los avisos
// —blancos y énfasis de Markdown alrededor de sus partes, cualquier selector de
// la marca y sin distinguir mayúsculas—, y si esa misma línea lleva
// https://kitlegal.es/instalar/. La dirección en otra línea no cuenta, tampoco
// con otra forma, y una etiqueta que no empieza su línea, incompleta o sin sus
// dos puntos no es la línea.
func TestExtraerSinConsulta(t *testing.T) {
	t.Parallel()

	const causa = "kitlegal no está instalado en este equipo"

	casos := []struct {
		nombre       string
		respuesta    string
		conLinea     bool
		conDireccion bool
	}{
		{
			nombre: "la-linea-con-su-direccion",
			respuesta: "⚠ SIN CONSULTA AL BOE: " + causa + ". Para consultarlo hace falta instalar kitlegal: " +
				"https://kitlegal.es/instalar/",
			conLinea:     true,
			conDireccion: true,
		},
		{
			nombre: "entre-otras-lineas",
			respuesta: "No he podido consultar la norma.\n\n" +
				"⚠ SIN CONSULTA AL BOE: " + causa + ". Para consultarlo hace falta instalar kitlegal: " +
				"https://kitlegal.es/instalar/\n\nCuando esté instalado, vuelve a preguntar.\n",
			conLinea:     true,
			conDireccion: true,
		},
		{
			nombre: "con-enfasis",
			respuesta: "**⚠ SIN CONSULTA AL BOE:** " + causa + ". Para consultarlo hace falta instalar kitlegal: " +
				"<https://kitlegal.es/instalar/>",
			conLinea:     true,
			conDireccion: true,
		},
		{
			nombre: "con-enfasis-blancos-selector-y-minusculas",
			// La marca con su selector de presentación U+FE0F, escrita con sus bytes.
			respuesta: "  _\xe2\x9a\xa0\xef\xb8\x8f  **Sin  consulta** al\tBOE_ : " + causa +
				" ([instalar kitlegal](https://kitlegal.es/instalar/)).",
			conLinea:     true,
			conDireccion: true,
		},
		{
			nombre:    "sin-la-direccion-en-su-linea",
			respuesta: "⚠ SIN CONSULTA AL BOE: " + causa + ". Para consultarlo hace falta instalar kitlegal.",
			conLinea:  true,
		},
		{
			nombre: "con-la-direccion-en-otra-linea",
			respuesta: "⚠ SIN CONSULTA AL BOE: " + causa + ". Para consultarlo hace falta instalar kitlegal:\n" +
				"https://kitlegal.es/instalar/\n",
			conLinea: true,
		},
		{
			nombre:    "con-otra-direccion",
			respuesta: "⚠ SIN CONSULTA AL BOE: " + causa + ". Más en https://kitlegal.es/instalar y en https://kitlegal.es/.",
			conLinea:  true,
		},
		{
			nombre: "una-de-dos-lineas-con-la-direccion",
			respuesta: "⚠ SIN CONSULTA AL BOE: " + causa + ".\n" +
				"⚠ SIN CONSULTA AL BOE: " + causa + ": https://kitlegal.es/instalar/\n",
			conLinea:     true,
			conDireccion: true,
		},
		{
			nombre:    "sin-la-linea",
			respuesta: "No he podido consultar el BOE. Instala kitlegal: https://kitlegal.es/instalar/\n",
		},
		{
			nombre: "la-etiqueta-no-empieza-su-linea",
			respuesta: "Aviso: ⚠ SIN CONSULTA AL BOE: " + causa + ". https://kitlegal.es/instalar/\n" +
				"> ⚠ SIN CONSULTA AL BOE: " + causa + ". https://kitlegal.es/instalar/\n",
		},
		{
			nombre: "otra-etiqueta-o-sin-los-dos-puntos",
			respuesta: "⚠ SIN CONSULTA: " + causa + ". https://kitlegal.es/instalar/\n" +
				"⚠ SIN CONSULTA AL BOE. " + causa + ". https://kitlegal.es/instalar/\n" +
				"SIN CONSULTA AL BOE: " + causa + ". https://kitlegal.es/instalar/\n" +
				"⚠ NORMA DEROGADA: esta norma ha sido derogada. https://kitlegal.es/instalar/\n",
		},
		{
			nombre: "respuesta-vacia",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			conLinea, conDireccion := ExtraerSinConsulta(caso.respuesta)

			assert.Equal(t, caso.conLinea, conLinea, "la respuesta tiene la línea")
			assert.Equal(t, caso.conDireccion, conDireccion, "la línea lleva la dirección")
		})
	}
}
