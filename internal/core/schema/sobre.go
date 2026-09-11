// Package schema declara el dominio del sobre de salida: la forma que envuelve
// el resultado de cualquier applet —con éxito o con fallo—, su procedencia, la
// huella reproducible de su contenido, las opciones globales que el kernel le
// entrega ya interpretadas y el vocabulario de clases de error.
//
// Es dominio puro: no hace entrada ni salida y no importa más que
// crypto/sha256, encoding/hex, encoding/json y time (R1 de
// contracts/reglas-de-arquitectura.md). Quien monta el sobre es el kernel, en
// un único punto; un applet solo devuelve un Resultado.
package schema

import "time"

// Sobre es el documento que el binario emite: seis claves exactas en el nivel
// superior, ni una más ni una menos, tanto en éxito como en fallo. Ninguna
// lleva omitempty, de modo que una clave con el valor cero se sigue emitiendo
// y ningún consumidor tiene que contar con que falte
// (contracts/sobre-de-salida.md §1, FR-010).
//
// Las etiquetas `jsonschema` son la descripción formal del contrato escrita
// junto a cada clave (contracts/sobre-de-salida.md §6): el kernel deriva de
// ellas el esquema que emite --describe, así que lo que ese esquema exige de
// `fuente`, `url` y `hash` se declara aquí, una sola vez, y no en una copia
// mantenida a mano (FR-017, FR-048). Son texto: el dominio no importa la
// biblioteca que las lee. La huella de la fecha —RFC 3339— no necesita
// etiqueta: el tipo time.Time ya se describe como `format: date-time`.
type Sobre struct {
	// Ok es verdadero si y solo si el código de salida del proceso es 0.
	Ok bool `json:"ok"`
	// Fuente identifica la procedencia comprobable del contenido y nunca va
	// vacía.
	Fuente string `json:"fuente" jsonschema:"minLength=1"`
	// URL es el URI absoluto de esa procedencia y nunca va vacía.
	URL string `json:"url" jsonschema:"minLength=1,format=uri"`
	// FechaConsulta es el instante de la consulta, serializado en RFC 3339 con
	// desplazamiento horario explícito.
	FechaConsulta time.Time `json:"fecha_consulta"`
	// Hash es la huella del contenido de Data en su forma canónica, precedida
	// del algoritmo que la produjo. El patrón de la etiqueta es PatronHuella,
	// escrito aquí como literal porque una etiqueta no puede nombrar una
	// constante; que los dos digan lo mismo lo comprueba TestSobre.
	Hash string `json:"hash" jsonschema:"pattern=^sha256:[0-9a-f]{64}$"`
	// Data es el contenido del applet cuando Ok, y DatosError cuando no.
	Data any `json:"data"`
}

// Procedencia es el par que hace citable a un sobre: de dónde sale el contenido
// y dónde puede comprobarse. Un applet que no consulta ninguna fuente externa
// usa el espacio de nombres reservado kitlegal. / kitlegal:, que señala un
// resultado calculado y no una cita de fuente pública (FR-016).
type Procedencia struct {
	Fuente string
	URL    string
}

// Validar rechaza las procedencias que no podrían sostener una cita: fuente
// vacía, url vacía y url que no sea un URI absoluto. Es el mismo criterio que
// aplica la descripción formal del sobre, que declara `format: uri` sobre esa
// clave (contracts/sobre-de-salida.md §6, FR-017).
func (p Procedencia) Validar() error {
	switch {
	case p.Fuente == "":
		return ErrFuenteVacia
	case p.URL == "":
		return ErrURLVacia
	case !esURIAbsoluto(p.URL):
		return ErrURLNoAbsoluta
	}

	return nil
}

// Resultado es lo que devuelve un applet: de dónde viene el contenido y el
// contenido. Ni Ok, ni Hash, ni FechaConsulta, ni forma de presentación, ni
// código de salida; de todo eso se ocupa el kernel, que es quien monta el sobre
// (FR-015, FR-044).
type Resultado struct {
	Procedencia Procedencia
	Datos       any
}

// errorDeValidacion expresa un error del dominio como constante: el paquete no
// importa errors ni fmt, que quedan fuera de las cuatro importaciones que R1 le
// permite.
type errorDeValidacion string

func (e errorDeValidacion) Error() string { return string(e) }

// Errores con los que Validar rechaza una procedencia que no sostiene una cita.
const (
	// ErrFuenteVacia lo devuelve una procedencia sin fuente.
	ErrFuenteVacia errorDeValidacion = "schema: la fuente del sobre no puede ir vacía"
	// ErrURLVacia lo devuelve una procedencia sin url.
	ErrURLVacia errorDeValidacion = "schema: la url del sobre no puede ir vacía"
	// ErrURLNoAbsoluta lo devuelve una procedencia cuya url no es un URI
	// absoluto.
	ErrURLNoAbsoluta errorDeValidacion = "schema: la url del sobre debe ser un uri absoluto"
)

// esURIAbsoluto comprueba lo que hace absoluto a un URI en RFC 3986 §4.3: un
// esquema bien formado seguido de dos puntos, que es también lo que exige
// net/url.URL.IsAbs —el criterio del validador formal del contrato—. Se
// rechazan además los caracteres de control, que hacen fallar a net/url.Parse
// antes de llegar a IsAbs. net/url queda fuera de las cuatro importaciones del
// dominio, así que la comprobación se escribe aquí.
func esURIAbsoluto(uri string) bool {
	if tieneCaracterDeControl(uri) {
		return false
	}

	return tieneEsquema(uri)
}

// tieneEsquema comprueba la producción `scheme` de RFC 3986 §3.1: una letra
// seguida de letras, dígitos, «+», «-» o «.», y cerrada por los dos puntos.
func tieneEsquema(uri string) bool {
	for i := range len(uri) {
		octeto := uri[i]
		switch {
		case octeto == ':':
			return i > 0
		case esLetra(octeto):
			continue
		case i > 0 && (esDigito(octeto) || octeto == '+' || octeto == '-' || octeto == '.'):
			continue
		default:
			return false
		}
	}

	return false
}

// tieneCaracterDeControl busca los octetos que net/url.Parse rechaza antes de
// analizar nada.
func tieneCaracterDeControl(uri string) bool {
	for i := range len(uri) {
		if uri[i] < 0x20 || uri[i] == 0x7f {
			return true
		}
	}

	return false
}

func esLetra(octeto byte) bool {
	return (octeto >= 'a' && octeto <= 'z') || (octeto >= 'A' && octeto <= 'Z')
}

func esDigito(octeto byte) bool { return octeto >= '0' && octeto <= '9' }
