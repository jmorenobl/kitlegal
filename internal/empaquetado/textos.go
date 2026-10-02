package empaquetado

import (
	"fmt"
	"unicode/utf8"
)

// Los textos con los que las dos piezas y el catálogo se presentan a quien los
// instala. Este es el único sitio en el que están escritos: de aquí los leen el
// manifiesto de la extensión, plugin.json y el catálogo, y quien compara con
// ellos (H22 FR-015; data-model §5; research.md D4).
const (
	// NombreVisible es `display_name` del manifiesto: el nombre con el que la
	// app enseña la extensión.
	NombreVisible = "kitlegal"

	// Descripcion es la descripción corta: `description` del manifiesto, de
	// plugin.json y de la entrada del catálogo. Es la que la release ya da al
	// cask, al bucket y a los paquetes, y no pasa de maximoDeLaDescripcion.
	Descripcion = "Tu asistente de IA responde con la ley vigente del BOE y la cita exacta"

	// DescripcionLarga es `long_description` del manifiesto: qué hace la
	// extensión, que todo corre en el equipo de quien la usa y que hace falta
	// también el plugin.
	DescripcionLarga = "kitlegal da a Claude herramientas para leer la legislación consolidada del Boletín " +
		"Oficial del Estado y situar una pregunta en su municipio: el texto vigente de cada artículo, con su " +
		"norma y su bloque para citarlo. Todo corre en tu equipo: lee fuentes públicas, guarda una caché en " +
		"~/.cache/kitlegal/ y no envía tus preguntas a ningún servidor de kitlegal. Para que las respuestas " +
		"lleven la cita con su forma, instala también el plugin de kitlegal, que trae las skills " +
		"(kitlegal-plugin.zip)."

	// Autoria es `author.name` del manifiesto, de plugin.json y de la entrada
	// del catálogo, y `owner.name` del catálogo: el proyecto, sin nombre de
	// persona ni correo.
	Autoria = "kitlegal"
)

// maximoDeLaDescripcion son los caracteres que la descripción corta puede
// tener como mucho en la ficha de la extensión (H22 FR-015).
const maximoDeLaDescripcion = 120

// comprobarDescripcion rechaza una descripción corta de más de
// maximoDeLaDescripcion caracteres, contados como caracteres y no como bytes:
// una letra con tilde es uno (H22 FR-015, FR-066; research.md D20).
func comprobarDescripcion(descripcion string) error {
	if caracteres := utf8.RuneCountInString(descripcion); caracteres > maximoDeLaDescripcion {
		return fmt.Errorf("la descripción corta tiene %d caracteres y el máximo es %d",
			caracteres, maximoDeLaDescripcion)
	}

	return nil
}
