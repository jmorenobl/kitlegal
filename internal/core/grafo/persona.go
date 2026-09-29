package grafo

import (
	"encoding/json/v2"
	"regexp"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// documentoDeIdentidad es la expresión de contracts/almacen-world-db.md §5,
// tal cual: una secuencia con la forma de un DNI (ocho cifras y una letra), de
// un NIE (X, Y o Z, siete cifras y una letra) o de un NIF de persona jurídica
// (una letra, siete cifras y una letra o una cifra), sin comprobar la letra ni
// la cifra de control, con un separador opcional —exactamente «.», «-» o el
// espacio U+0020— entre sus grupos, y entre dos límites: el principio o el
// final de la cadena o un carácter que no es letra ni cifra ASCII (FR-025).
// Las mayúsculas y las minúsculas van en clases ASCII explícitas, sin la
// bandera (?i): una letra es A-Z o a-z y una cifra es 0-9 (H7.1 FR-074).
var documentoDeIdentidad = regexp.MustCompile(`(?:^|[^A-Za-z0-9])(?:[0-9]{2}[.\- ]?[0-9]{3}[.\- ]?[0-9]{3}[.\- ]?[A-Za-z]|[XYZxyz][.\- ]?[0-9][.\- ]?[0-9]{3}[.\- ]?[0-9]{3}[.\- ]?[A-Za-z]|[A-Za-z][.\- ]?[0-9]{2}[.\- ]?[0-9]{3}[.\- ]?[0-9]{2}[.\- ]?[0-9A-Za-z])(?:[^A-Za-z0-9]|$)`)

// validarPersona rechaza una Persona cuyo id, o cualquier clave o valor de
// cadena de sus datos a cualquier profundidad de objetos y listas, tiene la
// forma de un documento de identidad (FR-025, constitución VII). Los valores
// que no son cadenas —números, booleanos, nulos— no pueden llevar la letra que
// exigen las tres formas y no se examinan.
//
// Los datos se examinan en su forma JSON canónica, la que se guarda: así se
// examina toda cadena que llegaría a world.db, venga de un []any o de un
// []string, de un map[string]any o de un mapa con otro tipo de valor, o de un
// tipo propio de cadena. Unos datos sin forma JSON no se pueden examinar, y se
// rechazan: ante la duda sobre un dato de una persona, prevalece el rechazo.
//
// El Rechazo nombra la Persona por su tipo y dice dónde está la forma, pero no
// la repite: ni el id, ni la clave, ni el valor, ni el error del codificador,
// que puede citar la ruta de claves hasta el valor que no tiene forma JSON.
func validarPersona(persona schema.Nodo) error {
	if documentoDeIdentidad.MatchString(persona.ID) {
		return &Rechazo{Operacion: persona, Motivo: "su id tiene la forma de un DNI, un NIE o un NIF"}
	}

	canonicos, err := DatosCanonicos(persona.Datos)

	var arbol any
	if err == nil {
		err = json.Unmarshal([]byte(canonicos), &arbol)
	}

	if err != nil {
		return &Rechazo{
			Operacion: persona,
			Motivo:    "sus datos no tienen forma JSON, y sin ella no se puede comprobar que no llevan un DNI, un NIE o un NIF",
		}
	}

	if llevaDocumento(arbol) {
		return &Rechazo{
			Operacion: persona,
			Motivo:    "sus datos llevan una clave o un valor con la forma de un DNI, un NIE o un NIF",
		}
	}

	return nil
}

// llevaDocumento dice si alguna de las claves o de las cadenas de un valor
// JSON decodificado —un objeto es un map[string]any; una lista, un []any; una
// cadena, un string—, a cualquier profundidad, tiene la forma de un documento
// de identidad.
func llevaDocumento(valor any) bool {
	switch v := valor.(type) {
	case string:
		return documentoDeIdentidad.MatchString(v)
	case map[string]any:
		for clave, dentro := range v {
			if documentoDeIdentidad.MatchString(clave) || llevaDocumento(dentro) {
				return true
			}
		}
	case []any:
		for _, dentro := range v {
			if llevaDocumento(dentro) {
				return true
			}
		}
	}

	return false
}
