package grafo

import (
	"cmp"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Una observación es la llegada de un nodo, una arista o un texto en un lote:
// su procedencia —fuente, url y fecha de consulta, como texto—, la vigencia que
// declara y, en un nodo, sus datos canónicos. De un nodo o una arista el grafo
// guarda la primera y la última; de un texto, la más antigua (FR-002, FR-023;
// data-model §4.1; research.md D13).
//
// Dos fechas se comparan como instantes: `…T12:00:00+02:00` y `…T10:00:00Z`
// son el mismo. Entre dos observaciones del mismo instante decide la url
// menor, comparando bytes; con la misma url, se queda la guardada (H7.1
// FR-076). La primera es la de menor instante y la última, la de mayor; a
// igualdad de instante, las dos son la de url menor; la de un texto, como la
// primera.
//
// Dos observaciones del mismo instante y la misma url son, en lo que producen
// los emisores, la misma: la de un lote, cuya procedencia comparten todas sus
// operaciones (research.md V29). Entre las que se producen, cada fusión elige
// siempre la misma, lleguen en el orden que lleguen, y el resultado no depende
// del orden de llegada ni de repetir un lote (FR-022). La primera nunca avanza,
// la última nunca retrocede y una observación idéntica a la guardada no cambia
// nada, así que quien guarda compara la fusión con lo guardado y escribe solo
// si cambia.

// RegistroDeNodo es lo que el grafo del mundo guarda de un nodo: su id, su
// tipo, su primera y su última observación, y los datos identificativos y la
// vigencia de la última, que la acompañan enteros y sin mezclarse con los de
// otra (FR-002, FR-023). Es una fila de la tabla nodes (data-model §3).
type RegistroDeNodo struct {
	// ID es su id natural.
	ID string
	// Tipo es su tipo, que no cambia nunca (FR-024).
	Tipo string
	// Datos son los datos identificativos de la última observación, en su
	// forma canónica (DatosCanonicos).
	Datos string
	// PrimeraObservacion es la procedencia de la primera observación.
	PrimeraObservacion Procedencia
	// UltimaObservacion es la procedencia de la última.
	UltimaObservacion Procedencia
	// Vigencia es la que declaró la última; cero si no declaró ninguna
	// (FR-065).
	Vigencia time.Duration
}

// RegistroDeArista es lo que el grafo del mundo guarda de una arista: su
// terna, su primera y su última observación, y la vigencia de la última
// (FR-002, FR-023). Es una fila de la tabla edges (data-model §3).
type RegistroDeArista struct {
	// Origen es el id del nodo del que sale.
	Origen string
	// Relacion es la relación que declara.
	Relacion string
	// Destino es el id del nodo al que llega.
	Destino string
	// PrimeraObservacion es la procedencia de la primera observación.
	PrimeraObservacion Procedencia
	// UltimaObservacion es la procedencia de la última.
	UltimaObservacion Procedencia
	// Vigencia es la que declaró la última; cero si no declaró ninguna
	// (FR-065).
	Vigencia time.Duration
}

// RegistroDeTexto es lo que el grafo del mundo guarda de un texto: su huella,
// su cuerpo y la procedencia de su observación más antigua (FR-002, FR-023).
// Es una fila de la tabla texts (data-model §3).
type RegistroDeTexto struct {
	// Huella es «sha256:» y los 64 hexadecimales de la SHA-256 del cuerpo.
	Huella string
	// Cuerpo es el texto, que no cambia nunca (FR-024).
	Cuerpo string
	// Procedencia es la de la observación más antigua.
	Procedencia Procedencia
}

// FusionarNodo junta lo que el grafo guarda de un nodo con lo que llega de él
// en un lote: la primera observación de las dos que es primera, y la última de
// las dos que es última, con sus datos y su vigencia (FR-023). Devuelve el
// registro fusionado, que es el guardado si lo que llega no cambia nada.
//
// Un tipo distinto del guardado rechaza el lote con un *Rechazo que nombra el
// nodo que llega por su id y su tipo o, si uno de los dos tipos es Persona, la
// Persona por su tipo (FR-024). Dos ids distintos y una fecha que no es RFC
// 3339 son un defecto de quien llama o de lo guardado, y dan un error que no
// es un Rechazo y que no nombra los ids, que pueden ser de una Persona.
func FusionarNodo(guardado, llegado RegistroDeNodo) (RegistroDeNodo, error) {
	switch {
	case guardado.ID != llegado.ID:
		return RegistroDeNodo{}, errors.New("no se pueden fusionar las observaciones de dos nodos distintos")
	case guardado.Tipo != llegado.Tipo:
		return RegistroDeNodo{}, rechazoDeTipo(guardado.Tipo, llegado)
	}

	fusionada, err := fusionarHistorias(guardado.historia(), llegado.historia())
	if err != nil {
		return RegistroDeNodo{}, err
	}

	guardado.Datos = fusionada.ultima.datos
	guardado.PrimeraObservacion = fusionada.primera
	guardado.UltimaObservacion = fusionada.ultima.procedencia
	guardado.Vigencia = fusionada.ultima.vigencia

	return guardado, nil
}

// FusionarArista junta lo que el grafo guarda de una arista con lo que llega
// de ella en un lote: la primera observación de las dos que es primera, y la
// última de las dos que es última, con su vigencia (FR-023). Devuelve el
// registro fusionado, que es el guardado si lo que llega no cambia nada.
//
// Dos ternas distintas y una fecha que no es RFC 3339 son un defecto de quien
// llama o de lo guardado, y dan un error que no es un Rechazo.
func FusionarArista(guardada, llegada RegistroDeArista) (RegistroDeArista, error) {
	if guardada.terna() != llegada.terna() {
		return RegistroDeArista{}, errors.New("no se pueden fusionar las observaciones de dos aristas distintas")
	}

	fusionada, err := fusionarHistorias(guardada.historia(), llegada.historia())
	if err != nil {
		return RegistroDeArista{}, err
	}

	guardada.PrimeraObservacion = fusionada.primera
	guardada.UltimaObservacion = fusionada.ultima.procedencia
	guardada.Vigencia = fusionada.ultima.vigencia

	return guardada, nil
}

// FusionarTexto junta lo que el grafo guarda de un texto con lo que llega de
// él en un lote: un solo texto por huella, con el mismo cuerpo y la procedencia
// de la observación más antigua de las dos (FR-023). La procedencia solo cambia
// si la que llega es de un instante estrictamente anterior, o del mismo y de
// url menor. Devuelve el registro fusionado, que es el guardado si lo que llega
// no cambia nada.
//
// Otro cuerpo con la misma huella rechaza el lote con un *Rechazo que nombra
// el texto que llega por su huella y no repite ningún cuerpo (FR-024). Dos
// huellas distintas y una fecha que no es RFC 3339 son un defecto de quien
// llama o de lo guardado, y dan un error que no es un Rechazo.
func FusionarTexto(guardado, llegado RegistroDeTexto) (RegistroDeTexto, error) {
	switch {
	case guardado.Huella != llegado.Huella:
		return RegistroDeTexto{}, errors.New("no se pueden fusionar las observaciones de dos textos distintos")
	case guardado.Cuerpo != llegado.Cuerpo:
		return RegistroDeTexto{}, &Rechazo{
			Operacion: schema.Texto{Huella: llegado.Huella, Cuerpo: llegado.Cuerpo},
			Motivo:    "el grafo ya guarda otro cuerpo con esa huella",
		}
	}

	orden, err := compararPrimeras(llegado.Procedencia, guardado.Procedencia)
	if err != nil {
		return RegistroDeTexto{}, err
	}

	if orden < 0 {
		guardado.Procedencia = llegado.Procedencia
	}

	return guardado, nil
}

// rechazoDeTipo es el Rechazo de un nodo que llega con otro tipo que el
// guardado. Nombra el nodo que llega, por su id y su tipo, salvo si el
// guardado es una Persona, que se nombra por su tipo, como en validarNodo: el
// mensaje nunca repite el id de una Persona.
func rechazoDeTipo(guardado string, llegado RegistroDeNodo) error {
	rechazado := schema.Nodo{ID: llegado.ID, Tipo: llegado.Tipo}
	if guardado == TipoPersona {
		rechazado.Tipo = TipoPersona
	}

	return &Rechazo{
		Operacion: rechazado,
		Motivo:    fmt.Sprintf("el grafo ya lo tiene con el tipo %q y el lote le da el tipo %q", guardado, llegado.Tipo),
	}
}

// historia es lo que se guarda de las observaciones de un nodo o una arista:
// la primera y la última.
type historia struct {
	primera Procedencia
	ultima  ultima
}

// ultima es la última observación de un nodo o una arista con lo que la
// acompaña y cambia con ella: su vigencia y, en un nodo, sus datos canónicos.
type ultima struct {
	procedencia Procedencia
	vigencia    time.Duration
	datos       string
}

// historia es la del nodo.
func (r RegistroDeNodo) historia() historia {
	return historia{
		primera: r.PrimeraObservacion,
		ultima:  ultima{procedencia: r.UltimaObservacion, vigencia: r.Vigencia, datos: r.Datos},
	}
}

// historia es la de la arista, cuya última no lleva datos.
func (r RegistroDeArista) historia() historia {
	return historia{
		primera: r.PrimeraObservacion,
		ultima:  ultima{procedencia: r.UltimaObservacion, vigencia: r.Vigencia},
	}
}

// terna es la clave de la arista: un registro por terna (FR-022).
func (r RegistroDeArista) terna() schema.Arista {
	return schema.Arista{Origen: r.Origen, Relacion: r.Relacion, Destino: r.Destino}
}

// fusionarHistorias elige, de las dos historias, la primera que es primera y
// la última que es última. A igualdad de instante y de url, se queda la
// guardada.
func fusionarHistorias(guardada, llegada historia) (historia, error) {
	ordenDeLaPrimera, err := compararPrimeras(llegada.primera, guardada.primera)
	if err != nil {
		return historia{}, err
	}

	ordenDeLaUltima, err := compararUltimas(llegada.ultima, guardada.ultima)
	if err != nil {
		return historia{}, err
	}

	if ordenDeLaPrimera < 0 {
		guardada.primera = llegada.primera
	}

	if ordenDeLaUltima < 0 {
		guardada.ultima = llegada.ultima
	}

	return guardada, nil
}

// compararPrimeras es negativo si a es primera antes que b: de menor instante
// o, del mismo, de url menor. Cero si son del mismo instante y la misma url.
func compararPrimeras(a, b Procedencia) (int, error) {
	instanteA, instanteB, err := instantes(a, b)
	if err != nil {
		return 0, err
	}

	return cmp.Or(instanteA.Compare(instanteB), desempatar(a, b)), nil
}

// compararUltimas es negativo si a es última antes que b: de mayor instante o,
// del mismo, de url menor. Cero si son del mismo instante y la misma url.
func compararUltimas(a, b ultima) (int, error) {
	instanteA, instanteB, err := instantes(a.procedencia, b.procedencia)
	if err != nil {
		return 0, err
	}

	return cmp.Or(instanteB.Compare(instanteA), desempatar(a.procedencia, b.procedencia)), nil
}

// desempatar ordena dos procedencias del mismo instante por su url, comparando
// bytes (H7.1 FR-076).
func desempatar(a, b Procedencia) int {
	return strings.Compare(a.URL, b.URL)
}

// instantes son los de las fechas de consulta de dos procedencias.
func instantes(a, b Procedencia) (time.Time, time.Time, error) {
	instanteA, err := instante(a.FechaConsulta)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	instanteB, err := instante(b.FechaConsulta)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return instanteA, instanteB, nil
}

// instante es el de una fecha de consulta RFC 3339, con o sin fracción de
// segundo, que es como la escribe el sobre y como la acepta ValidarLote
// (research.md D5).
func instante(fecha string) (time.Time, error) {
	leido, err := time.Parse(time.RFC3339, fecha)
	if err != nil {
		return time.Time{}, fmt.Errorf("la fecha de consulta %q no es RFC 3339: %w", fecha, err)
	}

	return leido, nil
}
