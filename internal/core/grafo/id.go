package grafo

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ValidarID decide si el id que recibe `graph show` puede ser el de algún
// nodo (FR-052; contracts/applet-graph.md §4; research.md D35). No puede
// serlo, y da un error de clase «argumentos», si está vacío, si todos sus
// caracteres son de espacio en blanco —la propiedad White_Space de Unicode,
// unicode.IsSpace: U+0020, el tabulador, los saltos, U+0085, U+00A0,
// U+2003…— o si contiene algún carácter de control —la categoría Cc de
// Unicode, unicode.IsControl: U+0000-U+001F y U+007F-U+009F—. Ningún id
// natural —un ELI, «ine:<código>», un DIR3— los lleva.
//
// Cualquier otro id es válido y se busca tal cual, sin recortar: uno con
// espacios y algo más (`a b`, ` a`), uno con un carácter de formato como
// U+200B, que no es de ninguna de las dos clases, y uno con bytes que no son
// UTF-8, que al recorrerlo se leen como U+FFFD y tampoco lo son; el id no se
// cambia, y se busca con esos bytes. Si no está en el grafo, es quien lo busca
// quien dice que no se encuentra (FR-053).
func ValidarID(id string) error {
	if id == "" {
		return idNoValido(id, "está vacío")
	}

	if strings.TrimFunc(id, unicode.IsSpace) == "" {
		return idNoValido(id, "solo tiene caracteres de espacio en blanco")
	}

	if posicion := strings.IndexFunc(id, unicode.IsControl); posicion >= 0 {
		control, _ := utf8.DecodeRuneInString(id[posicion:])

		return idNoValido(id, fmt.Sprintf("contiene el carácter de control %U", control))
	}

	return nil
}
