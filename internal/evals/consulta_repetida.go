package evals

import (
	"fmt"
	"slices"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
)

// conversacionDeLaConsulta es una de las dos conversaciones de la consulta
// repetida del quickstart, con el ordinal con el que la nombran las líneas de
// comprobarConsultaRepetida.
type conversacionDeLaConsulta struct {
	ordinal string
	sesion  Sesion
}

// comprobarConsultaRepetida comprueba las dos respuestas de la consulta repetida
// del quickstart (contracts/comprobacion-del-quickstart.md §2; FR-061) con el
// mismo código con el que el job juzga una respuesta: la forma de version-obsoleta
// de formasDeHallazgo y ExtraerHallazgos, y la lista de expresiones prohibidas con
// ExtraerExpresionesProhibidas. Devuelve una línea por lo que falla, en este orden,
// o nil si se cumple todo:
//
//  1. cada conversación que no terminó, con su MotivoSinTerminar;
//  2. la primera respuesta no lleva, en ninguna de sus líneas, la forma con las
//     fechas superada y leida como palabras;
//  3. la segunda respuesta lleva la forma;
//  4. cada respuesta que lleva expresiones de la lista, con las que lleva.
//
// Una conversación que no terminó no tiene respuesta que mirar: sus condiciones
// no se dan por cumplidas, y la línea que dice que no terminó basta. No añade
// criterios a Juzgar (FR-056).
func comprobarConsultaRepetida(primera, segunda Sesion, prohibidas ExpresionesProhibidas, superada, leida string) []string {
	conversaciones := []conversacionDeLaConsulta{{ordinal: "primera", sesion: primera}, {ordinal: "segunda", sesion: segunda}}
	forma := formaEscrita(grafo.EtiquetasDeHallazgo()[grafo.ClaseVersionObsoleta])

	var lineas []string

	for _, conversacion := range conversaciones {
		if !conversacion.sesion.Terminada {
			lineas = append(lineas, fmt.Sprintf("la %s conversación no terminó: %s", conversacion.ordinal,
				conversacion.sesion.MotivoSinTerminar))
		}
	}

	if primera.Terminada && !trasladaElCambio(primera.Respuesta, superada, leida) {
		lineas = append(lineas, fmt.Sprintf("la primera respuesta no lleva %s en una línea con %s y %s",
			forma, superada, leida))
	}

	if segunda.Terminada && slices.Contains(ExtraerHallazgos(segunda.Respuesta), string(grafo.ClaseVersionObsoleta)) {
		lineas = append(lineas, "la segunda respuesta lleva "+forma)
	}

	for _, conversacion := range conversaciones {
		if !conversacion.sesion.Terminada {
			continue
		}

		if encontradas := ExtraerExpresionesProhibidas(conversacion.sesion.Respuesta, prohibidas); len(encontradas) > 0 {
			lineas = append(lineas, fmt.Sprintf("la %s respuesta lleva expresiones prohibidas: %s",
				conversacion.ordinal, strings.Join(encontradas, ", ")))
		}
	}

	return lineas
}

// trasladaElCambio dice si alguna línea de la respuesta lleva la forma de
// version-obsoleta, la de formasDeHallazgo con la que se juzga, y las dos fechas
// como palabras (contieneComoPalabra). La forma no admite un salto de línea, así
// que mirar línea a línea no deja fuera ninguna.
func trasladaElCambio(respuesta, superada, leida string) bool {
	forma := formasDeHallazgo()[string(grafo.ClaseVersionObsoleta)]

	for linea := range strings.Lines(respuesta) {
		if forma.MatchString(linea) && contieneComoPalabra(linea, superada) && contieneComoPalabra(linea, leida) {
			return true
		}
	}

	return false
}
