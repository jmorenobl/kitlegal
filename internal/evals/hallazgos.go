package evals

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
)

// propiedadDeHallazgos es la propiedad del esquema de eval con las clases de
// hallazgo cuya forma fija tiene que llevar la respuesta.
const propiedadDeHallazgos = "hallazgos"

// formasDeHallazgo son las expresiones de la forma fija de cada clase de
// grafo.EtiquetasDeHallazgo, compiladas una sola vez con formasFijas, con la
// misma gramática y las mismas tolerancias que las de los avisos (H7.1 FR-051;
// research D12).
var formasDeHallazgo = sync.OnceValue(func() map[string]*regexp.Regexp {
	return formasFijas(grafo.EtiquetasDeHallazgo())
})

// clasesEtiquetadas son las clases de hallazgo que tienen etiqueta en
// grafo.EtiquetasDeHallazgo, las que una skill traslada con forma fija,
// ordenadas comparando bytes: el orden de ExtraerHallazgos y de los errores de
// las comprobaciones.
func clasesEtiquetadas() []string {
	clases := slices.Sorted(maps.Keys(grafo.EtiquetasDeHallazgo()))

	nombres := make([]string, 0, len(clases))
	for _, clase := range clases {
		nombres = append(nombres, string(clase))
	}

	return nombres
}

// ExtraerHallazgos devuelve las clases de hallazgo etiquetadas cuya forma fija
// —la marca, la etiqueta de grafo.EtiquetasDeHallazgo y los dos puntos, con las
// tolerancias de la de los avisos— lleva el texto, en el orden de las clases y
// sin repetir, o nil si no lleva ninguna. No lee lo que sigue a la forma ni
// ninguna otra redacción: decirlo con otras palabras no es la forma (H7.1 FR-051,
// FR-052).
func ExtraerHallazgos(texto string) []string {
	return extraerFormas(texto, clasesEtiquetadas(), formasDeHallazgo())
}

// ComprobarFormasDeHallazgo comprueba que el texto lleva la forma fija de cada
// clase de hallazgo etiquetada, con ExtraerHallazgos, la función con la que se
// juzga una respuesta: devuelve un error por clase que falta, en el orden de las
// clases y unidos con errors.Join, o nil si están todas. Cada error nombra la
// clase y escribe la forma que falta con la etiqueta de grafo.EtiquetasDeHallazgo
// (H7.1 FR-045; SC-008).
func ComprobarFormasDeHallazgo(texto string) error {
	encontradas := ExtraerHallazgos(texto)
	etiquetas := grafo.EtiquetasDeHallazgo()

	var defectos []error

	for _, clase := range clasesEtiquetadas() {
		if !slices.Contains(encontradas, clase) {
			defectos = append(defectos, fmt.Errorf("falta la forma fija del hallazgo %s: %s",
				clase, formaEscrita(etiquetas[grafo.ClaseDeHallazgo(clase)])))
		}
	}

	return errors.Join(defectos...)
}

// ComprobarClasesDeHallazgo comprueba que las clases que el esquema de eval
// compilado admite en hallazgos —las del enumerado al que llegan su items y cada
// $ref— son exactamente las clases de hallazgo etiquetadas: devuelve un error por
// clase etiquetada que el esquema no enumera, en el orden de las clases, y otro
// por valor que enumera sin ser una de ellas, en el orden del esquema, unidos con
// errors.Join; o nil si coinciden. Un esquema sin la propiedad hallazgos, sin
// items o sin enumerado no enumera ninguna (H7.1 FR-054; contrato evals-y-skill
// §3).
func ComprobarClasesDeHallazgo(esquema *jsonschema.Schema) error {
	clases := clasesEtiquetadas()
	enumeradas := enumeradoDeLaLista(esquema, propiedadDeHallazgos)

	var defectos []error

	for _, clase := range clases {
		// Comparar con una cadena no entra en pánico aunque el enumerado tenga
		// listas u objetos: con otro tipo dinámico, la igualdad es falsa.
		if !slices.Contains(enumeradas, any(clase)) {
			defectos = append(defectos, fmt.Errorf("el esquema de eval no enumera la clase de hallazgo %s en %s",
				clase, propiedadDeHallazgos))
		}
	}

	for _, enumerada := range enumeradas {
		if clase, esCadena := enumerada.(string); !esCadena || !slices.Contains(clases, clase) {
			defectos = append(defectos, fmt.Errorf("el esquema de eval enumera en %s la clase %v, que no es una "+
				"clase de hallazgo etiquetada por el binario", propiedadDeHallazgos, enumerada))
		}
	}

	return errors.Join(defectos...)
}
