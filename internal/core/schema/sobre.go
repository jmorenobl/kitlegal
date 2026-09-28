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

// Procedencia es lo que hace citable a un sobre: de dónde sale el contenido,
// dónde puede comprobarse y, cuando quien consultó lo sabe, cuándo se
// consultó. Un applet que no consulta ninguna fuente externa usa el espacio de
// nombres reservado kitlegal. / kitlegal:, que señala un resultado calculado y
// no una cita de fuente pública (FR-016).
type Procedencia struct {
	Fuente string
	URL    string
	// FechaConsulta es el instante de la consulta que sostiene el contenido.
	// El valor cero significa «no la declara quien consulta», y entonces el
	// kernel fecha el sobre con su reloj al montarlo, como antes de que este
	// campo existiera. Lo servido desde una caché la declara, para que el sobre
	// no aparente más frescura que la consulta que lo obtuvo (FR-096,
	// docs/ADR/0015).
	FechaConsulta time.Time
}

// Validar rechaza las procedencias que no podrían sostener una cita: fuente
// vacía, url vacía y url que no sea un URI absoluto. Es el mismo criterio que
// aplica la descripción formal del sobre, que declara `format: uri` sobre esa
// clave (contracts/sobre-de-salida.md §6, FR-017). La fecha de consulta no
// interviene: es opcional, y su ausencia no impide la cita porque la pone el
// kernel.
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

// Resultado es lo que devuelve un applet: de dónde —y, si lo sabe, cuándo—
// viene el contenido, el contenido, —si el applet la da— su forma para una
// persona, —solo en ensayo— lo que no llegó a hacerse y —si lo observa— lo que
// vio del mundo. Ni Ok, ni Hash, ni forma de presentación, ni código de salida;
// de todo eso se ocupa el kernel, que es quien monta el sobre y quien lo fecha
// cuando la procedencia no declara la fecha de consulta (FR-015, FR-044,
// FR-096).
type Resultado struct {
	Procedencia Procedencia
	Datos       any
	// Legible es el contenido contado para una persona: el texto que el kernel
	// escribe en la salida estándar, en lugar de la tabla mínima, cuando no se
	// pide --json. Vacío significa que el applet no lo cuenta y se presenta la
	// tabla, que es lo que hacen los applets que consultan fuentes. Lo compone
	// el applet a partir de los mismos Datos, en el mismo instante y sin
	// entrada ni salida, así que no puede decir otra cosa que el sobre; no
	// entra en el sobre ni en la huella, porque no es contenido citable: con
	// --json no cambia ni un byte (docs/ADR/0026).
	Legible string
	// Ensayo describe, una línea por operación, lo que cada capa con efectos
	// habría hecho en lugar de hacerlo. Solo se rellena bajo --dry-run, y solo
	// lo rellena quien tiene el efecto: el dominio no sabe presentarlo ni
	// escribirlo.
	//
	// Es el único camino por el que esa descripción llega a la salida de error
	// siempre visible, sin depender del nivel del registro de eventos: el
	// kernel la presenta por el presentador, y el applet no tiene ningún otro
	// modo de alcanzarlo. No entra en el sobre ni en la huella, porque no es
	// contenido citable (FR-051, FR-065, docs/ADR/0011).
	Ensayo []string
	// Grafo es lo que la invocación observó del mundo, sin fuente propia: el
	// kernel lo entrega al grafo del mundo después de presentar el sobre, con
	// la procedencia de ese sobre, y solo si la invocación termina bien. No
	// entra en el sobre ni en la huella. Su valor cero no emite nada, y es el
	// de todo applet que no observa el mundo (docs/ADR/0014).
	Grafo Observado
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
