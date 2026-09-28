package grafo

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// ValidarLote comprueba, sin tocar el disco, lo que un lote tiene que cumplir
// para entrar en el grafo del mundo con independencia de lo ya guardado
// (FR-024, FR-025; contracts/almacen-world-db.md §5; data-model §4.2;
// research.md D14):
//
//   - fuente y url no vacías, url que es un URI absoluto —el criterio de
//     schema.Procedencia.Validar—, fecha de consulta RFC 3339 y vigencia no
//     negativa;
//   - cada nodo con id y tipo, un solo tipo por id en todo el lote y ninguna
//     Persona con la forma de un DNI, un NIE o un NIF en su id o en cualquier
//     cadena de sus datos;
//   - cada arista con origen, relación y destino;
//   - cada texto con la huella de su cuerpo: «sha256:» y los 64 hexadecimales
//     en minúscula de la SHA-256 de sus bytes, tal cual;
//   - y cada operación, un valor schema.Nodo, schema.Arista o schema.Texto.
//
// Un solo incumplimiento rechaza el lote entero: devuelve el primero que
// encuentra, como un *Rechazo que nombra la operación —o el lote, si es su
// procedencia— y el motivo, y nil si no hay ninguno. Los nodos se validan
// antes que las aristas, de modo que una Persona con un documento en su id se
// rechaza por serlo antes de que ninguna arista la nombre por su id.
//
// Lo que necesita lo guardado —un extremo que no está ni en el lote ni en el
// grafo, un id que el grafo ya tiene con otro tipo, una huella guardada con
// otro cuerpo— se comprueba dentro de la transacción; contra un grafo vacío,
// con ValidarContraGrafoVacio.
func ValidarLote(lote core.Lote) error {
	if err := validarProcedencia(lote); err != nil {
		return err
	}

	tipos := make(map[string]schema.Nodo)

	for _, operacion := range lote.Operaciones {
		if nodo, esNodo := operacion.(schema.Nodo); esNodo {
			if err := validarNodo(nodo, tipos); err != nil {
				return err
			}
		}
	}

	for _, operacion := range lote.Operaciones {
		if err := validarOperacion(operacion); err != nil {
			return err
		}
	}

	return nil
}

// ValidarContraGrafoVacio comprueba lo que un lote tiene que cumplir para
// entrar en un grafo vacío —world.db ausente o sin esquema—, que no tiene
// ningún nodo: cada extremo de cada arista tiene que ser un nodo del propio
// lote, en cualquier posición (FR-024; contracts/almacen-world-db.md §4, pasos
// 3.1 y 4, y §5). El primer extremo que falta rechaza el lote entero con un
// *Rechazo que nombra la arista.
//
// Se llama después de ValidarLote y antes de crear o abrir nada, así que un
// lote que el grafo vacío rechaza no toca ningún fichero.
func ValidarContraGrafoVacio(lote core.Lote) error {
	nodos := make(map[string]bool)

	for _, operacion := range lote.Operaciones {
		if nodo, esNodo := operacion.(schema.Nodo); esNodo {
			nodos[nodo.ID] = true
		}
	}

	for _, operacion := range lote.Operaciones {
		arista, esArista := operacion.(schema.Arista)

		switch {
		case !esArista:
			continue
		case !nodos[arista.Origen]:
			return &Rechazo{Operacion: arista, Motivo: "su origen no es un nodo del lote y el grafo está vacío"}
		case !nodos[arista.Destino]:
			return &Rechazo{Operacion: arista, Motivo: "su destino no es un nodo del lote y el grafo está vacío"}
		}
	}

	return nil
}

// Consolidado es un lote válido con cada clave una sola vez: un registro por
// id de nodo, uno por terna de arista y uno por huella de texto, cada uno con
// la observación del lote como primera y como última, y ordenados por su clave
// comparando bytes (data-model §4.2). Es lo que el adaptador fusiona, registro
// a registro, con lo guardado (contracts/almacen-world-db.md §4, paso 6).
type Consolidado struct {
	// Nodos son los registros de los nodos, por id.
	Nodos []RegistroDeNodo
	// Aristas son los registros de las aristas, por origen, relación y
	// destino.
	Aristas []RegistroDeArista
	// Textos son los registros de los textos, por huella.
	Textos []RegistroDeTexto
}

// Consolidar valida el lote con ValidarLote y lo reduce a un registro por
// clave (data-model §4.2; FR-023). Dentro de un lote, dos operaciones con la
// misma clave —id, terna o huella— comparten la observación del lote, con su
// procedencia y su vigencia, y quedan en un solo registro. Dos nodos con el
// mismo id y datos distintos se quedan con los datos canónicos menores,
// comparando bytes: el último criterio del desempate de FusionarNodo, que es
// el único en el que difieren dos observaciones del mismo lote. Dos textos con
// la misma huella tienen el mismo cuerpo, porque ValidarLote exige que la
// huella sea la de su cuerpo.
//
// El resultado no depende del orden de las operaciones ni de que alguna se
// repita. Un lote que ValidarLote rechaza da su Rechazo, y uno con un nodo
// cuyos datos no tienen forma JSON canónica, que no se podrían guardar ni
// comparar, un Rechazo que nombra el nodo y dice por qué.
func Consolidar(lote core.Lote) (Consolidado, error) {
	if err := ValidarLote(lote); err != nil {
		return Consolidado{}, err
	}

	procedencia := Procedencia{Fuente: lote.Fuente, URL: lote.URL, FechaConsulta: lote.FechaConsulta}
	nodos := make(map[string]RegistroDeNodo)
	aristas := make(map[schema.Arista]RegistroDeArista)
	textos := make(map[string]RegistroDeTexto)

	for _, operacion := range lote.Operaciones {
		switch op := operacion.(type) {
		case schema.Nodo:
			datos, err := DatosCanonicos(op.Datos)
			if err != nil {
				return Consolidado{}, &Rechazo{Operacion: op, Motivo: err.Error()}
			}

			if previo, visto := nodos[op.ID]; visto && previo.Datos <= datos {
				continue
			}

			nodos[op.ID] = RegistroDeNodo{
				ID: op.ID, Tipo: op.Tipo, Datos: datos,
				PrimeraObservacion: procedencia, UltimaObservacion: procedencia, Vigencia: lote.Vigencia,
			}
		case schema.Arista:
			aristas[op] = RegistroDeArista{
				Origen: op.Origen, Relacion: op.Relacion, Destino: op.Destino,
				PrimeraObservacion: procedencia, UltimaObservacion: procedencia, Vigencia: lote.Vigencia,
			}
		case schema.Texto:
			textos[op.Huella] = RegistroDeTexto{Huella: op.Huella, Cuerpo: op.Cuerpo, Procedencia: procedencia}
		}
	}

	return Consolidado{
		Nodos: slices.SortedFunc(maps.Values(nodos), func(a, b RegistroDeNodo) int {
			return strings.Compare(a.ID, b.ID)
		}),
		Aristas: slices.SortedFunc(maps.Values(aristas), func(a, b RegistroDeArista) int {
			return cmp.Or(
				strings.Compare(a.Origen, b.Origen),
				strings.Compare(a.Relacion, b.Relacion),
				strings.Compare(a.Destino, b.Destino),
			)
		}),
		Textos: slices.SortedFunc(maps.Values(textos), func(a, b RegistroDeTexto) int {
			return strings.Compare(a.Huella, b.Huella)
		}),
	}, nil
}

// validarProcedencia rechaza el lote mismo si su procedencia no sostiene una
// cita o su vigencia es negativa. Si la fuente y la url no están vacías, lo
// único que schema.Procedencia.Validar puede rechazar es que la url no sea un
// URI absoluto. La fecha tiene que ser RFC 3339, con o sin fracción de
// segundo, que es como la escribe el sobre (research.md D5).
func validarProcedencia(lote core.Lote) error {
	_, errFecha := time.Parse(time.RFC3339, lote.FechaConsulta)

	switch {
	case lote.Fuente == "":
		return &Rechazo{Motivo: "no lleva fuente"}
	case lote.URL == "":
		return &Rechazo{Motivo: "no lleva url"}
	case (schema.Procedencia{Fuente: lote.Fuente, URL: lote.URL}).Validar() != nil:
		return &Rechazo{Motivo: fmt.Sprintf("la url %q no es un URI absoluto", lote.URL)}
	case lote.FechaConsulta == "":
		return &Rechazo{Motivo: "no lleva fecha de consulta"}
	case errFecha != nil:
		return &Rechazo{Motivo: fmt.Sprintf("la fecha de consulta %q no es RFC 3339", lote.FechaConsulta)}
	case lote.Vigencia < 0:
		return &Rechazo{Motivo: fmt.Sprintf("la vigencia %s es negativa", lote.Vigencia)}
	}

	return nil
}

// validarNodo rechaza un nodo sin id o sin tipo, una Persona con un documento
// de identidad y un id al que el lote ya dio otro tipo. tipos guarda, por id,
// el primer nodo del lote con ese id.
//
// Si uno de los dos nodos con el mismo id es una Persona, el Rechazo la
// nombra a ella, que se nombra por su tipo: el mensaje nunca repite el id de
// una Persona.
func validarNodo(nodo schema.Nodo, tipos map[string]schema.Nodo) error {
	switch {
	case nodo.ID == "":
		return &Rechazo{Operacion: nodo, Motivo: "no lleva id"}
	case nodo.Tipo == "":
		return &Rechazo{Operacion: nodo, Motivo: "no lleva tipo"}
	}

	if nodo.Tipo == TipoPersona {
		if err := validarPersona(nodo); err != nil {
			return err
		}
	}

	previo, visto := tipos[nodo.ID]
	if !visto {
		tipos[nodo.ID] = nodo

		return nil
	}

	if previo.Tipo == nodo.Tipo {
		return nil
	}

	rechazado := nodo
	if previo.Tipo == TipoPersona {
		rechazado = previo
	}

	return &Rechazo{
		Operacion: rechazado,
		Motivo:    fmt.Sprintf("el lote le da dos tipos, %q y %q", previo.Tipo, nodo.Tipo),
	}
}

// validarOperacion rechaza una arista sin origen, relación o destino, un
// texto cuya huella no es la de su cuerpo y lo que no es un valor
// schema.Nodo, schema.Arista o schema.Texto: una operación nula o un puntero a
// uno de ellos, que la interfaz sellada también admite. Los nodos ya los
// validó ValidarLote.
func validarOperacion(operacion schema.Operacion) error {
	switch op := operacion.(type) {
	case schema.Nodo:
		return nil
	case schema.Arista:
		return validarArista(op)
	case schema.Texto:
		return validarTexto(op)
	case nil:
		return &Rechazo{Motivo: "trae una operación nula"}
	default:
		return &Rechazo{Operacion: op, Motivo: "solo entran los valores schema.Nodo, schema.Arista y schema.Texto"}
	}
}

// validarArista rechaza una arista a la que le falta un extremo o la relación.
func validarArista(arista schema.Arista) error {
	switch {
	case arista.Origen == "":
		return &Rechazo{Operacion: arista, Motivo: "no lleva origen"}
	case arista.Relacion == "":
		return &Rechazo{Operacion: arista, Motivo: "no lleva relación"}
	case arista.Destino == "":
		return &Rechazo{Operacion: arista, Motivo: "no lleva destino"}
	}

	return nil
}

// validarTexto rechaza un texto cuya huella no es exactamente la de los bytes
// de su cuerpo, tal cual y sin normalizarlo: la forma de hash_texto (FR-071).
// El Rechazo nombra el texto por su huella y nunca repite el cuerpo.
func validarTexto(texto schema.Texto) error {
	suma := sha256.Sum256([]byte(texto.Cuerpo))
	if texto.Huella != schema.PrefijoHuella+hex.EncodeToString(suma[:]) {
		return &Rechazo{
			Operacion: texto,
			Motivo:    "su huella no es la de su cuerpo: «sha256:» y los 64 hexadecimales en minúscula de su SHA-256",
		}
	}

	return nil
}
