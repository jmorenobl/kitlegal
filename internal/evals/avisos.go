package evals

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// Las piezas de la forma fija de un aviso, las de la gramática del contrato de la
// forma fija §1 (research D2). Ninguna admite un salto de línea, así que la forma
// entera va en una línea.
const (
	// marcaDeAviso es la marca U+26A0 seguida, a lo sumo, de uno de sus dos
	// selectores de presentación, U+FE0E o U+FE0F.
	marcaDeAviso = `\x{26A0}[\x{FE0E}\x{FE0F}]?`

	// separadorDeAviso son cero o más tabuladores, blancos de la categoría Zs de
	// Unicode, asteriscos o guiones bajos: los blancos y el énfasis de Markdown
	// alrededor de las partes.
	separadorDeAviso = `[\t\p{Zs}*_]*`

	// entrePalabrasDeAviso es un separador con al menos un blanco, el que va
	// entre dos palabras de la etiqueta.
	entrePalabrasDeAviso = `[\t\p{Zs}*_]*[\t\p{Zs}][\t\p{Zs}*_]*`

	// finalDeAviso son los dos puntos U+003A que cierran la forma.
	finalDeAviso = `:`
)

// propiedadDeAvisos es la propiedad del esquema de eval con los códigos de aviso
// esperados.
const propiedadDeAvisos = "avisos"

// formasDeAviso son las expresiones de la forma fija de cada código de
// boe.CodigosDeAviso, compiladas una sola vez desde boe.EtiquetasDeAviso con
// formasFijas.
var formasDeAviso = sync.OnceValue(func() map[string]*regexp.Regexp {
	return formasFijas(boe.EtiquetasDeAviso())
})

// formasFijas son las expresiones de formaFija de cada etiqueta que tiene
// palabras, por su clave. Una clave cuya etiqueta no tiene palabras no tiene
// expresión y nunca se encuentra.
func formasFijas[Clave ~string](etiquetas map[Clave]string) map[string]*regexp.Regexp {
	formas := map[string]*regexp.Regexp{}

	for clave, etiqueta := range etiquetas {
		if forma := formaFija(etiqueta); forma != nil {
			formas[string(clave)] = forma
		}
	}

	return formas
}

// formaFija es la expresión de la forma fija de una etiqueta, la de los avisos
// y la de los hallazgos (H7.1 FR-051; research D12): la marca, un separador, las
// palabras de la etiqueta en su orden, cada una sin distinguir mayúsculas y con
// al menos un blanco entre dos, otro separador y los dos puntos; nil si la
// etiqueta no tiene palabras. Las piezas son fijas y cada palabra pasa por
// regexp.QuoteMeta, así que la compilación no puede fallar.
func formaFija(etiqueta string) *regexp.Regexp {
	palabras := strings.Fields(etiqueta)
	if len(palabras) == 0 {
		return nil
	}

	for i, palabra := range palabras {
		palabras[i] = `(?i:` + regexp.QuoteMeta(palabra) + `)`
	}

	return regexp.MustCompile(marcaDeAviso + separadorDeAviso +
		strings.Join(palabras, entrePalabrasDeAviso) + separadorDeAviso + finalDeAviso)
}

// formaEscrita es la forma fija de una etiqueta tal como se enseña y como la
// nombra un error: la marca, un espacio, la etiqueta y los dos puntos.
func formaEscrita(etiqueta string) string {
	return "⚠ " + etiqueta + ":"
}

// ExtraerAvisos devuelve los códigos de boe.CodigosDeAviso cuya forma fija —la
// marca, la etiqueta de boe.EtiquetasDeAviso y los dos puntos, con las
// tolerancias de la gramática— lleva el texto, en el orden de CodigosDeAviso y sin
// repetir, o nil si no lleva ninguna. No lee lo que sigue a la forma ni ninguna
// otra redacción (FR-032, FR-033).
func ExtraerAvisos(texto string) []string {
	return extraerFormas(texto, boe.CodigosDeAviso(), formasDeAviso())
}

// extraerFormas devuelve, en el orden de claves, las que tienen en formas una
// expresión que casa con el texto, o nil si no hay ninguna. Una clave sin
// expresión no se encuentra nunca.
func extraerFormas(texto string, claves []string, formas map[string]*regexp.Regexp) []string {
	var encontradas []string

	for _, clave := range claves {
		if forma, conForma := formas[clave]; conForma && forma.MatchString(texto) {
			encontradas = append(encontradas, clave)
		}
	}

	return encontradas
}

// ComprobarFormasDeAviso comprueba que el texto lleva la forma fija de cada código
// de boe.CodigosDeAviso, con ExtraerAvisos: devuelve un error por código que
// falta, en el orden de CodigosDeAviso y unidos con errors.Join, o nil si están
// todos. Cada error escribe la forma que falta con la etiqueta de
// boe.EtiquetasDeAviso (FR-014; data-model §6).
func ComprobarFormasDeAviso(texto string) error {
	encontrados := ExtraerAvisos(texto)
	etiquetas := boe.EtiquetasDeAviso()

	var defectos []error

	for _, codigo := range boe.CodigosDeAviso() {
		if !slices.Contains(encontrados, codigo) {
			defectos = append(defectos, fmt.Errorf("falta la forma fija del aviso %s: %s",
				codigo, formaEscrita(etiquetas[codigo])))
		}
	}

	return errors.Join(defectos...)
}

// ComprobarCodigosDeAviso comprueba que los códigos que el esquema de eval
// compilado admite en avisos —los del enumerado al que llegan su items y cada
// $ref— son exactamente los de boe.CodigosDeAviso: devuelve un error por código
// del binario que el esquema no enumera, en el orden de CodigosDeAviso, y otro por
// código que enumera sin ser del binario, en el orden del esquema, unidos con
// errors.Join; o nil si coinciden. Un esquema sin la propiedad avisos, sin items o
// sin enumerado no enumera ninguno (FR-013; data-model §6).
func ComprobarCodigosDeAviso(esquema *jsonschema.Schema) error {
	codigos := boe.CodigosDeAviso()
	enumerados := enumeradoDeLaLista(esquema, propiedadDeAvisos)

	var defectos []error

	for _, codigo := range codigos {
		// Comparar con una cadena no entra en pánico aunque el enumerado tenga
		// listas u objetos: con otro tipo dinámico, la igualdad es falsa.
		if !slices.Contains(enumerados, any(codigo)) {
			defectos = append(defectos, fmt.Errorf("el esquema de eval no enumera el código de aviso %s en %s",
				codigo, propiedadDeAvisos))
		}
	}

	for _, enumerado := range enumerados {
		if codigo, esCadena := enumerado.(string); !esCadena || !slices.Contains(codigos, codigo) {
			defectos = append(defectos, fmt.Errorf("el esquema de eval enumera en %s el código %v, que no es un "+
				"código de aviso del binario", propiedadDeAvisos, enumerado))
		}
	}

	return errors.Join(defectos...)
}

// enumeradoDeLaLista son los valores del enumerado al que llegan, en el esquema
// compilado, la propiedad nombrada, sus items de 2020-12 y cada $ref, en el
// orden del esquema; nil si no llegan a ninguno. Una cadena de $ref que vuelve a
// un esquema ya visto no llega a ningún enumerado.
func enumeradoDeLaLista(esquema *jsonschema.Schema, propiedad string) []any {
	lista, conLista := esquema.Properties[propiedad]
	if !conLista {
		return nil
	}

	vistos := map[*jsonschema.Schema]bool{}

	for items := lista.Items2020; items != nil && !vistos[items]; items = items.Ref {
		if items.Enum != nil {
			return items.Enum.Values
		}

		vistos[items] = true
	}

	return nil
}
