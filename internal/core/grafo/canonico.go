package grafo

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
)

// DatosCanonicos devuelve los datos identificativos de un nodo en su forma JSON
// canónica, la del esquema de canonicalización JSON de RFC 8785 (JCS): sin
// espacios, con las claves de cada objeto ordenadas por sus unidades UTF-16,
// cada cadena con los escapes mínimos y cada número como un doble de IEEE 754
// escrito con el algoritmo de ECMAScript. Es la forma con la que se guardan y
// con la que se comparan, byte a byte, con los guardados para escribir solo lo
// que cambia (research.md D17, V5). Unos datos nulos o vacíos son «{}».
//
// La forma no depende de cómo se construyen los datos: ni del orden en el que
// Go visita las claves de un mapa ni del tipo de Go de cada número o de cada
// objeto. Unos datos que no la tienen —un número que no es finito, una cadena
// o una clave que no es UTF-8, un valor que no es de JSON— dan un error y no
// se corrigen: sustituir un byte por U+FFFD cambiaría el dato.
func DatosCanonicos(datos map[string]any) (string, error) {
	codificados, err := json.Marshal(datos)
	canonicos := jsontext.Value(codificados)

	// Lo que da Marshal ya es JSON; Canonicalize solo lo reescribe en su forma
	// canónica, que Marshal por sí solo no garantiza.
	if err == nil {
		err = canonicos.Canonicalize()
	}

	if err != nil {
		return "", fmt.Errorf("los datos no tienen forma JSON canónica (RFC 8785): %w", err)
	}

	return string(canonicos), nil
}
