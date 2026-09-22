package evals

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/jmorenobl/kitlegal/internal/core/territorio"
)

// Las propiedades del esquema de eval que llevan la cobertura del territorio
// esperado, y su ruta tal como la nombran los defectos del esquema.
const (
	propiedadDeTerritorio = "territorio"
	propiedadDeCobertura  = "cobertura"
	rutaDeLaCobertura     = propiedadDeTerritorio + "/" + propiedadDeCobertura
)

// separadorDeAspecto es el que va entre la clave de un aspecto de cobertura y su
// valor en la forma <aspecto>: <valor>, la del enumerado aspecto-de-cobertura del
// esquema de eval (contrato de evals §1.2 de H6).
const separadorDeAspecto = ": "

// Lo que precede al nombre o al código de un elemento del territorio esperado en
// el texto con el que lo presentan el informe y los motivos. Un aspecto de
// cobertura no lleva nada delante: su texto ya es <aspecto>: <valor>.
const (
	elementoDeComunidad = "comunidad: "
	elementoDeProvincia = "provincia: "
	elementoDeBoletin   = "boletín: "
)

// Las piezas de la forma fija de un aspecto de cobertura. Ninguna admite un salto
// de línea, así que la forma entera va en una línea.
const (
	// blancoDeAspecto son cero o más tabuladores o blancos de la categoría Zs de
	// Unicode, los que se toleran a cada lado de los dos puntos.
	blancoDeAspecto = `[\t\p{Zs}]*`

	// antesDelAspecto y despuesDelAspecto delimitan la clave y el valor como
	// palabras exactas: el principio o el final del texto, o un carácter que no
	// es letra, cifra, guion bajo ni guion, los que forman las claves y los
	// valores del vocabulario.
	antesDelAspecto   = `(?:^|[^\p{L}\p{N}_-])`
	despuesDelAspecto = `(?:$|[^\p{L}\p{N}_-])`
)

// aspectoDeCobertura es la clave y el valor de un aspecto de cobertura en la forma
// <aspecto>: <valor>.
func aspectoDeCobertura(clave, valor string) string {
	return clave + separadorDeAspecto + valor
}

// aspectosDeCobertura son las combinaciones de clave y valor del vocabulario de
// territorio.AspectosDeCobertura en la forma <aspecto>: <valor>, en el orden de
// sus claves y, en cada una, de sus valores.
func aspectosDeCobertura() []string {
	var combinaciones []string

	for _, aspecto := range territorio.AspectosDeCobertura() {
		for _, valor := range aspecto.Valores {
			combinaciones = append(combinaciones, aspectoDeCobertura(aspecto.Clave, valor))
		}
	}

	return combinaciones
}

// formasDeCobertura son las expresiones de la forma fija de cada combinación del
// vocabulario de territorio.AspectosDeCobertura, compiladas una sola vez: la
// clave, los dos puntos con cualquier blanco o ninguno a cada lado y el valor, sin
// distinguir mayúsculas y delimitados la clave y el valor como palabras. Las
// piezas son fijas y la clave y el valor pasan por regexp.QuoteMeta, así que la
// compilación no puede fallar.
var formasDeCobertura = sync.OnceValue(func() map[string]*regexp.Regexp {
	formas := map[string]*regexp.Regexp{}

	for _, aspecto := range territorio.AspectosDeCobertura() {
		for _, valor := range aspecto.Valores {
			formas[aspectoDeCobertura(aspecto.Clave, valor)] = regexp.MustCompile(`(?i)` + antesDelAspecto +
				regexp.QuoteMeta(aspecto.Clave) + blancoDeAspecto + `:` + blancoDeAspecto + regexp.QuoteMeta(valor) +
				despuesDelAspecto)
		}
	}

	return formas
})

// ExtraerTerritorio devuelve lo que la respuesta declara del territorio esperado,
// cada elemento por su forma fija (contrato de evals §2 de H6):
//
//   - la comunidad y la provincia, si la respuesta plegada con territorio.Plegar
//     contiene su nombre plegado como palabras enteras: sin distinguir mayúsculas
//     ni diacríticos y con cualquier puntuación o blanco entre sus palabras, el
//     único pliegue del árbol, el del dominio;
//   - cada boletín, si la respuesta contiene su código exacto como palabra;
//   - y cada aspecto de cobertura, si la respuesta lleva <aspecto>: <valor> en
//     una línea, con cualquier blanco o ninguno a cada lado de los dos puntos y
//     sin distinguir mayúsculas, exactas las dos palabras. Un aspecto fuera del
//     vocabulario de territorio.AspectosDeCobertura nunca se declara.
//
// Lo que no declara queda vacío; cada boletín y cada aspecto declarados van en el
// orden del esperado y con sus repeticiones. No lee ninguna otra redacción: el
// nombre de una provincia que la comunidad lleva en el suyo, como Madrid en
// Comunidad de Madrid, se da por declarado con ella.
func ExtraerTerritorio(respuesta string, esperado TerritorioEsperado) TerritorioEsperado {
	plegada := " " + territorio.Plegar(respuesta) + " "
	formas := formasDeCobertura()

	var declarado TerritorioEsperado

	if declaraElNombre(plegada, esperado.Comunidad) {
		declarado.Comunidad = esperado.Comunidad
	}

	if declaraElNombre(plegada, esperado.Provincia) {
		declarado.Provincia = esperado.Provincia
	}

	for _, boletin := range esperado.Boletines {
		if contieneComoPalabra(respuesta, boletin) {
			declarado.Boletines = append(declarado.Boletines, boletin)
		}
	}

	for _, aspecto := range esperado.Cobertura {
		if forma, conForma := formas[aspecto]; conForma && forma.MatchString(respuesta) {
			declarado.Cobertura = append(declarado.Cobertura, aspecto)
		}
	}

	return declarado
}

// declaraElNombre dice si la respuesta plegada, con un espacio delante y otro
// detrás, contiene el nombre plegado entre dos espacios. Un nombre que se pliega a
// nada no se declara nunca: el pliegue no deja dos espacios seguidos, así que
// cualquier respuesta lo contendría.
func declaraElNombre(plegada, nombre string) bool {
	plegado := territorio.Plegar(nombre)

	return plegado != "" && strings.Contains(plegada, " "+plegado+" ")
}

// elementos son los del territorio esperado con el texto con el que los presentan
// el informe y los motivos, en este orden: comunidad: <nombre>, provincia:
// <nombre>, boletín: <código> por cada boletín y cada aspecto de cobertura tal
// como lo escribe la eval. Lo que el esperado no declara no tiene elemento, así
// que el de una eval que no espera territorio no tiene ninguno.
func (t TerritorioEsperado) elementos() []string {
	var elementos []string

	if t.Comunidad != "" {
		elementos = append(elementos, elementoDeComunidad+t.Comunidad)
	}

	if t.Provincia != "" {
		elementos = append(elementos, elementoDeProvincia+t.Provincia)
	}

	for _, boletin := range t.Boletines {
		elementos = append(elementos, elementoDeBoletin+boletin)
	}

	return append(elementos, t.Cobertura...)
}

// ComprobarAspectosDeCobertura comprueba que lo que el esquema de eval compilado
// admite en la cobertura del territorio esperado —el enumerado al que llegan su
// items y cada $ref— son exactamente las combinaciones del vocabulario de
// territorio.AspectosDeCobertura en la forma <aspecto>: <valor>: devuelve un error
// por combinación del binario que el esquema no enumera, en el orden del
// vocabulario, y otro por valor que enumera sin ser una de ellas, en el orden del
// esquema, unidos con errors.Join; o nil si coinciden. Un esquema sin la propiedad
// territorio, sin su cobertura, sin items o sin enumerado no enumera ninguna
// (contrato de evals §1.2 de H6).
func ComprobarAspectosDeCobertura(esquema *jsonschema.Schema) error {
	vocabulario := aspectosDeCobertura()
	enumerados := enumeradoDeCobertura(esquema)

	var defectos []error

	for _, aspecto := range vocabulario {
		// Comparar con una cadena no entra en pánico aunque el enumerado tenga
		// números, listas u objetos: con otro tipo dinámico, la igualdad es falsa.
		if !slices.Contains(enumerados, any(aspecto)) {
			defectos = append(defectos, fmt.Errorf("el esquema de eval no enumera el aspecto de cobertura %s en %s",
				aspecto, rutaDeLaCobertura))
		}
	}

	for _, enumerado := range enumerados {
		if aspecto, esCadena := enumerado.(string); !esCadena || !slices.Contains(vocabulario, aspecto) {
			defectos = append(defectos, fmt.Errorf("el esquema de eval enumera en %s el aspecto %v, "+
				"que no es un aspecto de cobertura del binario", rutaDeLaCobertura, enumerado))
		}
	}

	return errors.Join(defectos...)
}

// enumeradoDeCobertura son los valores del enumerado al que llegan, en el esquema
// compilado, la propiedad cobertura de la propiedad territorio, sus items de
// 2020-12 y cada $ref, en el orden del esquema; nil si no llegan a ninguno. Una
// cadena de $ref que vuelve a un esquema ya visto no llega a ningún enumerado.
func enumeradoDeCobertura(esquema *jsonschema.Schema) []any {
	territorioEsperado, conTerritorio := esquema.Properties[propiedadDeTerritorio]
	if !conTerritorio {
		return nil
	}

	cobertura, conCobertura := territorioEsperado.Properties[propiedadDeCobertura]
	if !conCobertura {
		return nil
	}

	vistos := map[*jsonschema.Schema]bool{}

	for items := cobertura.Items2020; items != nil && !vistos[items]; items = items.Ref {
		if items.Enum != nil {
			return items.Enum.Values
		}

		vistos[items] = true
	}

	return nil
}
