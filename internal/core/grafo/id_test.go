package grafo_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestValidarID fija qué id de `graph show` no puede ser el de ningún nodo y
// sale con 2, clase «argumentos» (FR-052; contracts/applet-graph.md §4 y §7;
// research.md D35): el vacío, el formado solo por caracteres de espacio en
// blanco —la propiedad White_Space de Unicode, unicode.IsSpace— y el que
// contiene algún carácter de control —la categoría Cc, unicode.IsControl—.
// Todo lo demás es un id válido que se busca tal cual, sin recortar: uno con
// espacios y algo más, uno con un carácter de formato como U+200B, que no es
// de ninguna de las dos clases, y uno con bytes que no son UTF-8, que se leen
// como U+FFFD y tampoco lo son; si no está, sale con 3 (FR-053). Los
// caracteres que no son ASCII van con su escape de Go, para que se vean en el
// diff.
func TestValidarID(t *testing.T) {
	t.Parallel()

	const espacios = "solo tiene caracteres de espacio en blanco"

	invalidos := []struct {
		nombre string
		id     string
		motivo string
	}{
		// Los de contracts/applet-graph.md §7, uno por clase.
		{"vacio", "", "está vacío"},
		{"solo U+0020", " ", espacios},
		{"solo U+00A0", "\u00a0", espacios},
		{"solo U+2003", "\u2003", espacios},
		{"solo un tabulador", "\t", espacios},
		{"solo U+0085, de las dos clases", "\u0085", espacios},
		{"solo U+007F", "\u007f", "contiene el carácter de control U+007F"},
		{"U+0000 dentro", "a\u0000b", "contiene el carácter de control U+0000"},
		// Y los bordes de cada clase: muchos espacios distintos juntos, un
		// control que también es espacio junto a otro carácter y un control de
		// la parte alta de Latin-1.
		{"varios espacios distintos", " \t\n\v\f\r\u00a0\u2003\u3000", espacios},
		{"un salto de linea detras", "ine:28074\n", "contiene el carácter de control U+000A"},
		{"U+009F dentro", "ine:\u009f28074", "contiene el carácter de control U+009F"},
	}

	for _, caso := range invalidos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			err := grafo.ValidarID(caso.id)
			require.EqualError(t, err, fmt.Sprintf("el id %q no puede ser el de ningún nodo: %s", caso.id, caso.motivo))

			// La clase llega al kernel también envuelta, que es como la recibe
			// del applet.
			var conClase schema.ConClase
			require.ErrorAs(t, fmt.Errorf("graph show: %w", err), &conClase)
			assert.Equal(t, schema.ClaseArgumentos, conClase.Clase())
		})
	}

	validos := []struct {
		nombre string
		id     string
	}{
		{"un ELI", "eli/es/l/2015/10/01/39#a21"},
		{"un codigo INE", "ine:28074"},
		{"un DIR3", "L01280745"},
		{"espacios y algo mas", "a b"},
		{"un espacio delante", " a"},
		{"U+00A0 entre letras", "a\u00a0b"},
		{"solo U+200B, que no es de ninguna clase", "\u200b"},
		{"un byte que no es UTF-8", "a\xffb"},
		{"solo un byte que no es UTF-8", "\xff"},
	}

	for _, caso := range validos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, grafo.ValidarID(caso.id))
		})
	}
}
