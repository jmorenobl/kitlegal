package grafo

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Las formas de salida de los verbos de `graph` —el data de show, stats y
// check— con las claves JSON de data-model §5 en español y en el orden de
// contracts/applet-graph.md §3, y la instantánea que lee check, que no sale.
//
// Una lista de show o de stats sale como [] y los datos de un nodo como {},
// nunca null, aunque lleguen nulos: el kernel escribe el sobre con
// encoding/json, que escribe null para un slice o un mapa nil, y Ficha y
// Recuento lo corrigen al escribirse (FR-053, FR-054). La lista de hallazgos de
// check es un []Hallazgo, y no nula la devuelve Comprobar (FR-060). Las claves
// propias de cada clase de hallazgo van con omitempty, así que no aparecen en
// la otra.

// Procedencia es de dónde y cuándo llegó una observación: la fuente, la url y
// la fecha de consulta del sobre que la sostuvo. La fecha es el texto que
// escribió ese sobre, carácter a carácter, sin normalizarla a otro
// desplazamiento ni quitarle o añadirle fracción de segundo (FR-053, FR-061).
type Procedencia struct {
	// Fuente es la `fuente` del sobre.
	Fuente string `json:"fuente"`
	// URL es la `url` del sobre.
	URL string `json:"url"`
	// FechaConsulta es la `fecha_consulta` del sobre, RFC 3339.
	FechaConsulta string `json:"fecha_consulta"`
}

// Ficha es lo que `graph show` devuelve de un nodo: el nodo y sus aristas
// salientes y entrantes, ordenadas por relación y después por el id del otro
// extremo, comparando bytes (FR-053; contracts/applet-graph.md §3.1). Nunca
// lleva el cuerpo de un texto (FR-070).
type Ficha struct {
	// Nodo es el nodo pedido.
	Nodo NodoDeFicha `json:"nodo"`
	// Salientes son las aristas que salen de él.
	Salientes []AristaDeFicha `json:"salientes"`
	// Entrantes son las aristas que llegan a él.
	Entrantes []AristaDeFicha `json:"entrantes"`
}

// NodoDeFicha es un nodo tal como lo da `graph show`: su id, su tipo, sus
// datos identificativos —los de su última observación— y su primera y su
// última observación (FR-053).
type NodoDeFicha struct {
	// ID es su id natural.
	ID string `json:"id"`
	// Tipo es su tipo.
	Tipo string `json:"tipo"`
	// Datos son sus datos identificativos, un objeto JSON.
	Datos map[string]any `json:"datos"`
	// PrimeraObservacion es la fecha de consulta de su primera observación,
	// como la escribió su sobre.
	PrimeraObservacion string `json:"primera_observacion"`
	// UltimaObservacion es la procedencia de su última observación.
	UltimaObservacion Procedencia `json:"ultima_observacion"`
}

// AristaDeFicha es una arista de un nodo tal como la da `graph show`: su
// relación, el id del otro extremo y su primera y su última observación
// (FR-053).
type AristaDeFicha struct {
	// Relacion es la de la arista.
	Relacion string `json:"relacion"`
	// ID es el del otro extremo: el destino de una saliente, el origen de una
	// entrante.
	ID string `json:"id"`
	// PrimeraObservacion es la fecha de consulta de su primera observación,
	// como la escribió su sobre.
	PrimeraObservacion string `json:"primera_observacion"`
	// UltimaObservacion es la procedencia de su última observación.
	UltimaObservacion Procedencia `json:"ultima_observacion"`
}

// MarshalJSON escribe la ficha como la escribiría encoding/json sin este
// método, salvo que unos datos nulos salen como {} y unas aristas nulas como
// []: nunca null (FR-053).
func (f Ficha) MarshalJSON() ([]byte, error) {
	type sinMetodos Ficha

	escrita := sinMetodos(f)
	escrita.Nodo.Datos = objetoNoNulo(escrita.Nodo.Datos)
	escrita.Salientes = listaNoNula(escrita.Salientes)
	escrita.Entrantes = listaNoNula(escrita.Entrantes)

	return codificar(escrita)
}

// Recuento es lo que `graph stats` devuelve: los totales de nodos, aristas y
// textos, y los nodos de cada par (tipo, fuente) y las aristas de cada par
// (relación, fuente) que hay en el grafo, donde la fuente es la de la última
// observación (FR-054; contracts/applet-graph.md §3.2). Los textos solo se
// cuentan en total, uno por huella.
type Recuento struct {
	// Nodos es el total de nodos.
	Nodos int `json:"nodos"`
	// Aristas es el total de aristas, una por terna.
	Aristas int `json:"aristas"`
	// Textos es el total de textos, uno por huella.
	Textos int `json:"textos"`
	// NodosPorTipo son los pares con algún nodo, ordenados por tipo y después
	// por fuente, comparando bytes.
	NodosPorTipo []RecuentoDeNodos `json:"nodos_por_tipo"`
	// AristasPorRelacion son los pares con alguna arista, ordenados por
	// relación y después por fuente, comparando bytes.
	AristasPorRelacion []RecuentoDeAristas `json:"aristas_por_relacion"`
}

// RecuentoDeNodos es el número de nodos de un tipo cuya última observación es
// de una fuente; nunca cero (FR-054).
type RecuentoDeNodos struct {
	// Tipo es el de los nodos.
	Tipo string `json:"tipo"`
	// Fuente es la de su última observación.
	Fuente string `json:"fuente"`
	// Nodos es cuántos son.
	Nodos int `json:"nodos"`
}

// RecuentoDeAristas es el número de aristas de una relación cuya última
// observación es de una fuente; nunca cero (FR-054).
type RecuentoDeAristas struct {
	// Relacion es la de las aristas.
	Relacion string `json:"relacion"`
	// Fuente es la de su última observación.
	Fuente string `json:"fuente"`
	// Aristas es cuántas son.
	Aristas int `json:"aristas"`
}

// MarshalJSON escribe el recuento como lo escribiría encoding/json sin este
// método, salvo que unas listas nulas salen como []: nunca null (FR-054).
func (r Recuento) MarshalJSON() ([]byte, error) {
	type sinMetodos Recuento

	escrito := sinMetodos(r)
	escrito.NodosPorTipo = listaNoNula(escrito.NodosPorTipo)
	escrito.AristasPorRelacion = listaNoNula(escrito.AristasPorRelacion)

	return codificar(escrito)
}

// Hallazgo es lo que `graph check` encuentra en el grafo: una versión superada
// de un bloque o una consulta cuya vigencia ha pasado, con su explicación
// citable y la procedencia en la que se apoya (FR-061, FR-063, FR-066;
// contracts/applet-graph.md §3.3). Un hallazgo es un dato de un resultado
// correcto, no un fallo (ADR 0023).
type Hallazgo struct {
	// Clase es la del hallazgo.
	Clase ClaseDeHallazgo `json:"clase"`
	// ID es el del nodo del que se dice.
	ID string `json:"id"`
	// Explicacion es la frase de contracts/applet-graph.md §5.
	Explicacion string `json:"explicacion"`
	// Procedencia es la de la última observación en la que se apoya: la de la
	// versión más reciente en una versión obsoleta, la del propio nodo en una
	// fuente caducada.
	Procedencia Procedencia `json:"procedencia"`
	// FechaVigencia es la de la versión superada; solo en una versión
	// obsoleta.
	FechaVigencia string `json:"fecha_vigencia,omitempty"`
	// FechaVigenciaReciente es la de la versión más reciente; solo en una
	// versión obsoleta.
	FechaVigenciaReciente string `json:"fecha_vigencia_reciente,omitempty"`
	// VigenciaSegundos es la que declaró la consulta caducada; solo en una
	// fuente caducada.
	VigenciaSegundos int64 `json:"vigencia_segundos,omitempty"`
}

// Instantanea es todo el grafo de una lectura consistente, lo que lee `graph
// check`: cada nodo con su última observación y su vigencia, y cada arista
// (data-model §5; research.md D15). No sale en ningún sobre y no lleva
// etiquetas JSON.
type Instantanea struct {
	// Nodos son todos los nodos.
	Nodos []NodoDeInstantanea
	// Aristas son todas las aristas, una por terna.
	Aristas []schema.Arista
}

// NodoDeInstantanea es un nodo de una instantánea: su id, su tipo, sus datos
// identificativos y su última observación, con la vigencia que declaró.
type NodoDeInstantanea struct {
	// ID es su id natural.
	ID string
	// Tipo es su tipo.
	Tipo string
	// Datos son sus datos identificativos.
	Datos map[string]any
	// UltimaObservacion es la procedencia de su última observación.
	UltimaObservacion Procedencia
	// Vigencia es la que declaró su última observación; cero si no declaró
	// ninguna (FR-065).
	Vigencia time.Duration
}

// listaNoNula devuelve la lista, o una vacía si es nil, que encoding/json
// escribe como [] y no como null.
func listaNoNula[E any](lista []E) []E {
	if lista == nil {
		return []E{}
	}

	return lista
}

// objetoNoNulo devuelve el objeto, o uno vacío si es nil, que encoding/json
// escribe como {} y no como null.
func objetoNoNulo(objeto map[string]any) map[string]any {
	if objeto == nil {
		return map[string]any{}
	}

	return objeto
}

// codificar escribe valor con encoding/json sin escapar los caracteres de HTML
// y sin el salto de línea final. Es lo que escribiría el kernel para un valor
// sin MarshalJSON: el kernel compacta lo que devuelve MarshalJSON dentro del
// sobre y es quien decide si escapa «&», «<» y «>», que un escape anticipado
// aquí dejaría escapados siempre (internal/render).
func codificar(valor any) ([]byte, error) {
	var escrito bytes.Buffer

	codificador := json.NewEncoder(&escrito)
	codificador.SetEscapeHTML(false)

	if err := codificador.Encode(valor); err != nil {
		return nil, err
	}

	return bytes.TrimSuffix(escrito.Bytes(), []byte("\n")), nil
}
