package schema

import "time"

// Observado es lo que una invocación observó del mundo: las identidades, las
// relaciones y los textos que vio, y la vigencia que la fuente declara para esa
// consulta. Viaja en el Resultado junto a la procedencia y los datos, y el
// kernel lo entrega al grafo del mundo después de presentar el sobre, con la
// procedencia de ese sobre (docs/ADR/0014, FR-020, FR-021).
//
// Su valor cero es válido y no emite nada: es lo que devuelve todo applet que
// no observa el mundo, sin implementar nada nuevo (FR-020, FR-047).
type Observado struct {
	// Vigencia es la que la fuente declara para la consulta. Cero significa
	// que no la declara (FR-065).
	Vigencia time.Duration
	// Operaciones son lo observado, en el orden en que el applet las declara.
	Operaciones []Operacion
}

// Operacion es una operación de grafo: un Nodo, una Arista o un Texto, y nada
// más. La interfaz está sellada por un método sin exportar, de modo que ningún
// tipo de otro paquete puede serlo y quien la recibe distingue la clase con un
// type switch sobre esos tres valores (research.md D2).
//
// Ninguna operación lleva fuente, url ni fecha de consulta: son las de la
// Procedencia del Resultado que la trae, y las pone el kernel al entregarla
// (FR-021).
type Operacion interface {
	operacionDeGrafo()
}

// Nodo declara una identidad del mundo por su id natural.
type Nodo struct {
	// ID es el id natural: un ELI, «ine:<código>», un DIR3.
	ID string
	// Tipo es la clase de la identidad: Norma, Bloque, BloqueVersion,
	// Municipio, Organo. Un tipo que no cambia nunca para un mismo id (FR-024).
	Tipo string
	// Datos son los datos identificativos, como un objeto JSON; nunca el
	// cuerpo de un bloque, que viaja en un Texto (FR-041).
	Datos map[string]any
}

// Arista declara una relación dirigida entre dos nodos por sus ids.
type Arista struct {
	Origen   string
	Relacion string
	Destino  string
}

// Texto declara un cuerpo de texto por su huella.
type Texto struct {
	// Huella es «sha256:» seguido de los 64 hexadecimales en minúscula de los
	// bytes de Cuerpo (FR-024, FR-071).
	Huella string
	Cuerpo string
}

func (Nodo) operacionDeGrafo()   {}
func (Arista) operacionDeGrafo() {}
func (Texto) operacionDeGrafo()  {}
