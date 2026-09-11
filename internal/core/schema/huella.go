package schema

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// PrefijoHuella nombra el algoritmo que produjo la huella. Va delante del
// hexadecimal para que se pueda cambiar de algoritmo sin romper a quien la lee,
// que debe comprobar el prefijo antes de interpretar el resto (FR-012).
const PrefijoHuella = "sha256:"

// PatronHuella es la expresión regular que cumple toda huella que produce
// Huella: el prefijo del algoritmo y los 64 dígitos hexadecimales en minúscula
// del SHA-256 (contracts/sobre-de-salida.md §2). Es el mismo patrón que la
// etiqueta de Sobre.Hash declara para la descripción formal del sobre; se
// exporta para que quien valide una huella fuera del dominio no lo reescriba.
const PatronHuella = "^sha256:[0-9a-f]{64}$"

// Huella devuelve la huella del contenido de un sobre: el prefijo del algoritmo
// seguido de los 64 dígitos hexadecimales en minúscula del SHA-256 de la forma
// canónica de data. El mismo contenido produce siempre la misma huella, con
// independencia del orden de las claves y del instante de la consulta, y un
// contenido que difiere en un solo byte produce otra distinta (FR-011, SC-005).
//
// Si data no es serializable devuelve un error —nunca un pánico—, de modo que
// la construcción del sobre falle antes de que nada llegue a la salida estándar
// (contracts/sobre-de-salida.md §4).
func Huella(data any) (string, error) {
	forma, err := canonico(data)
	if err != nil {
		return "", err
	}

	suma := sha256.Sum256(forma)

	return PrefijoHuella + hex.EncodeToString(suma[:]), nil
}

// canonico lleva data a su forma canónica en los tres pasos del contrato
// (contracts/sobre-de-salida.md §4): serializarlo a JSON, releerlo conservando
// los números como literales y volver a serializarlo con las claves de todo
// objeto ordenadas —encoding/json las ordena al serializar un mapa—, sin
// espacios ni saltos de línea y sin escapar caracteres HTML.
func canonico(data any) ([]byte, error) {
	crudo, err := json.Marshal(data)
	if err != nil {
		return nil, errorDeSerializacion{causa: err}
	}

	var raiz nodo
	if err := json.Unmarshal(crudo, &raiz); err != nil {
		return nil, errorDeSerializacion{causa: err}
	}

	var emitido acumulador
	codificador := json.NewEncoder(&emitido)
	codificador.SetEscapeHTML(false)
	if err := codificador.Encode(raiz.valor); err != nil {
		return nil, errorDeSerializacion{causa: err}
	}

	return emitido.sinSaltoFinal(), nil
}

// nodo reconstruye el valor genérico que produce un decodificador con
// UseNumber: mapas, listas, cadenas, booleanos, nulos y los números conservados
// como literales, sin pasar por la coma flotante que destruiría la precisión de
// un entero grande. Se implementa como json.Unmarshaler porque json.Decoder —la
// vía directa a UseNumber— exige un io.Reader, e io queda fuera de las cuatro
// importaciones que R1 permite al dominio.
type nodo struct {
	valor any
}

// UnmarshalJSON reparte por el primer octeto del valor, que el analizador de
// encoding/json ya ha validado antes de entregarlo aquí.
func (n *nodo) UnmarshalJSON(crudo []byte) error {
	switch primerOcteto(crudo) {
	case '{':
		campos := map[string]nodo{}
		if err := json.Unmarshal(crudo, &campos); err != nil {
			return err
		}
		objeto := make(map[string]any, len(campos))
		for clave, campo := range campos {
			objeto[clave] = campo.valor
		}
		n.valor = objeto
	case '[':
		var elementos []nodo
		if err := json.Unmarshal(crudo, &elementos); err != nil {
			return err
		}
		lista := make([]any, len(elementos))
		for i, elemento := range elementos {
			lista[i] = elemento.valor
		}
		n.valor = lista
	case '"':
		var cadena string
		if err := json.Unmarshal(crudo, &cadena); err != nil {
			return err
		}
		n.valor = cadena
	case 't', 'f':
		var booleano bool
		if err := json.Unmarshal(crudo, &booleano); err != nil {
			return err
		}
		n.valor = booleano
	case 'n':
		n.valor = nil
	default:
		n.valor = json.Number(crudo)
	}

	return nil
}

// primerOcteto devuelve el primer octeto del valor, o cero si viniera vacío
// —cosa que el analizador de encoding/json no hace—, de modo que el reparto no
// pueda terminar en un pánico (FR-033). Un valor vacío acabaría en la rama del
// número y la volvería a serializar como literal inválido, que es un error.
func primerOcteto(crudo []byte) byte {
	if len(crudo) == 0 {
		return 0
	}

	return crudo[0]
}

// acumulador recoge en memoria lo que escribe el codificador: cumple el
// contrato de escritura que json.NewEncoder espera sin que el dominio importe
// io ni bytes (R1).
type acumulador []byte

func (a *acumulador) Write(bloque []byte) (int, error) {
	*a = append(*a, bloque...)

	return len(bloque), nil
}

// sinSaltoFinal devuelve lo acumulado sin el salto de línea con el que
// json.Encoder.Encode termina siempre: la forma canónica no lo lleva.
func (a acumulador) sinSaltoFinal() []byte {
	if ultimo := len(a) - 1; ultimo >= 0 && a[ultimo] == '\n' {
		return a[:ultimo]
	}

	return a
}

// errorDeSerializacion envuelve el fallo de encoding/json cuando el contenido
// de data no se puede llevar a su forma canónica. Es un tipo y no un
// fmt.Errorf porque fmt queda fuera de las cuatro importaciones del dominio.
type errorDeSerializacion struct {
	causa error
}

func (e errorDeSerializacion) Error() string {
	return "schema: el contenido de data no es serializable: " + e.causa.Error()
}

func (e errorDeSerializacion) Unwrap() error { return e.causa }
