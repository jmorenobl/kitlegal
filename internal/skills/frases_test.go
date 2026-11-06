package skills_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// TestFrases fija, con un cuerpo escrito en el test por regla, los segmentos en
// que Frases parte el cuerpo de un SKILL.md (contracts/remisiones.md §6 de H41;
// research D6; FR-001, FR-023): una frase acaba en un punto seguido de un
// espacio o de un salto de línea, y sus blancos se juntan en uno; la marca de una
// viñeta o de un paso abre una; un encabezado y un comentario HTML son segmentos
// solos; cada línea de un bloque de código, vallas incluidas, es un segmento tal
// cual y corta la frase; y el punto de «.md» dentro de un enlace no la cierra.
func TestFrases(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		cuerpo    string
		segmentos []string
	}{
		{
			nombre:    "dos frases en una línea",
			cuerpo:    "Toma el id de la tabla. Si no está, búscalo.\n",
			segmentos: []string{"Toma el id de la tabla.", "Si no está, búscalo."},
		},
		{
			nombre: "una frase partida en tres líneas con sangría",
			cuerpo: "- Sin fecha, lee cada\n  bloque una vez, con su\n  comprobación detrás. Y la cita.\n",
			segmentos: []string{
				"- Sin fecha, lee cada bloque una vez, con su comprobación detrás.",
				"Y la cita.",
			},
		},
		{
			nombre: "una viñeta y un elemento numerado abren frase",
			cuerpo: "Lo que se pide depende de\n- la forma de la referencia\n2. **El cotejo.** es lo primero\n" +
				"* y su fecha\n",
			segmentos: []string{
				"Lo que se pide depende de",
				"- la forma de la referencia",
				"2. **El cotejo.** es lo primero",
				"* y su fecha",
			},
		},
		{
			nombre: "un encabezado y un comentario HTML son segmentos solos",
			cuerpo: "Antes de seguir\n### 3. Cotejar. La ficha\nlee la ficha\n  <!-- paquetes: lo genera make skills-sync -->  \n" +
				"y cita.\n",
			segmentos: []string{
				"Antes de seguir",
				"### 3. Cotejar. La ficha",
				"lee la ficha",
				"<!-- paquetes: lo genera make skills-sync -->",
				"y cita.",
			},
		},
		{
			nombre: "un bloque de código corta la frase y sus líneas son segmentos",
			cuerpo: "Antes de la primera orden, lee\n\n```text\nkitlegal alfa  --json. Y\n\n```\n" +
				"[uno](references/uno.md): lo que trae.\n",
			segmentos: []string{
				"Antes de la primera orden, lee",
				"```text",
				"kitlegal alfa  --json. Y",
				"",
				"```",
				"[uno](references/uno.md): lo que trae.",
			},
		},
		{
			nombre: "el .md de un enlace seguido de dos puntos no cierra la frase",
			cuerpo: "Antes de pedirla, lee [x.md](references/x.md): la orden de cada forma. Lo demás",
			segmentos: []string{
				"Antes de pedirla, lee [x.md](references/x.md): la orden de cada forma.",
				"Lo demás",
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.segmentos, skills.Frases(caso.cuerpo))
		})
	}
}
