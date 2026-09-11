package app

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// La ayuda del binario y la de cada applet salen **solo** del registro: no hay
// ninguna lista paralela mantenida a mano, de modo que registrar un applet basta
// para que aparezca y declarar un verbo basta para que se enumere (FR-001,
// FR-026, contracts/registro-y-describe.md §4).
//
// Las dos son texto para una persona, y quien las escribe es el presentador por
// la salida estándar, sin que --json las altere: la contraparte legible por
// máquina de la ayuda es --describe (contracts/banderas-y-exit-codes.md §2).

// entrada es una línea de una lista de la ayuda: lo que se escribe en la
// invocación y lo que explica qué hace.
type entrada struct {
	nombre  string
	detalle string
}

// AyudaDelBinario es lo que responde --help cuando la invocación no resuelve
// ningún applet: cómo se invoca el binario y qué applets registra.
//
// invocacion es el nombre con el que se llamó al binario, y se usa tal cual: un
// enlace con un nombre que no es de ningún applet no inutiliza el binario
// (FR-005), así que su ayuda tampoco puede hablar de otro nombre que el suyo.
func AyudaDelBinario(invocacion string, r *Registro) string {
	var ayuda strings.Builder

	fmt.Fprintf(&ayuda, "uso: %s <applet> [verbo] [banderas]\n", invocacion)

	entradas := entradasDeApplets(r)
	if len(entradas) == 0 {
		// El binario que se publica no registra ninguno en H1, y la ayuda lo
		// dice en lugar de emitir una lista vacía que nadie sabría interpretar
		// (contracts/registro-y-describe.md §3).
		ayuda.WriteString("\nEste binario no registra ningún applet.\n")

		return ayuda.String()
	}

	ayuda.WriteString(bloque("applets:", entradas))
	fmt.Fprintf(&ayuda, "\nLos verbos de cada applet, en «%s <applet> --help».\n", invocacion)

	return ayuda.String()
}

// AyudaDelApplet es lo que responde --help sobre un applet: cómo se invoca y qué
// verbos declara, con el verbo por omisión marcado como tal.
//
// Pedir la ayuda de un applet **no** resuelve ningún verbo, así que la lista es
// la del applet entero (contracts/registro-y-describe.md §2 bis). La última
// línea señala dónde están las banderas de un verbo concreto, que es lo que
// responde la ayuda del analizador cuando el verbo sí se nombra.
//
// El texto depende solo del applet y nunca del nombre con el que se invocó, que
// es lo que permite que la ayuda por enlace simbólico y por primer argumento
// sean idénticas byte a byte (FR-007, SC-003).
func AyudaDelApplet(a Applet) string {
	var ayuda strings.Builder

	nombre := a.Nombre()

	fmt.Fprintf(&ayuda, "uso: %s <verbo> [banderas]\n", nombre)

	//nolint:misspell // «Descripcion» es español y lo fija el contrato; el diccionario de misspell es solo inglés (research.md D9).
	if descripcion := a.Descripcion(); descripcion != "" {
		fmt.Fprintf(&ayuda, "\n%s\n", descripcion)
	}

	ayuda.WriteString(bloque("verbos:", entradasDeVerbos(a)))
	fmt.Fprintf(&ayuda, "\nLas banderas de cada verbo, en «%s <verbo> --help».\n", nombre)

	return ayuda.String()
}

// listaDeApplets es la enumeración que acompaña a un applet que no se ha podido
// resolver, en una sola línea porque va dentro del mensaje del fallo: el mismo
// que llega a la salida de error y al contenido del sobre de fallo (FR-006).
func listaDeApplets(r *Registro) string {
	nombres := r.Nombres()
	if len(nombres) == 0 {
		return "este binario no registra ningún applet"
	}

	return "applets disponibles: " + strings.Join(nombres, ", ")
}

// listaDeVerbos es la enumeración que acompaña a una invocación que no nombra
// verbo sobre un applet que no declara ninguno por omisión
// (contracts/registro-y-describe.md §2 bis).
func listaDeVerbos(a Applet) string {
	verbos := a.Verbos()
	nombres := make([]string, 0, len(verbos))

	for _, verbo := range verbos {
		nombres = append(nombres, verbo.Nombre)
	}

	return "verbos de " + a.Nombre() + ": " + strings.Join(nombres, ", ")
}

// entradasDeApplets enumera el registro en el orden estable de Nombres, que es
// lo que hace que dos invocaciones iguales produzcan la misma ayuda.
func entradasDeApplets(r *Registro) []entrada {
	nombres := r.Nombres()
	entradas := make([]entrada, 0, len(nombres))

	for _, nombre := range nombres {
		// Nombres sale del mismo registro que Buscar consulta, así que el applet
		// está siempre: no hay ausencia que tratar.
		applet, _ := r.Buscar(nombre)

		//nolint:misspell // «Descripcion» es español y lo fija el contrato; el diccionario de misspell es solo inglés (research.md D9).
		entradas = append(entradas, entrada{nombre: nombre, detalle: applet.Descripcion()})
	}

	return entradas
}

// entradasDeVerbos conserva el orden en que el applet declara sus verbos: es el
// orden que eligió quien lo escribió, y el registro ya ha comprobado que no hay
// dos con el mismo nombre.
func entradasDeVerbos(a Applet) []entrada {
	verbos := a.Verbos()
	entradas := make([]entrada, 0, len(verbos))

	for _, verbo := range verbos {
		//nolint:misspell // «Descripcion» es español y lo fija el contrato; el diccionario de misspell es solo inglés (research.md D9).
		detalle := verbo.Descripcion

		if verbo.PorOmision {
			detalle = strings.TrimSpace(detalle + " (por omisión)")
		}

		entradas = append(entradas, entrada{nombre: verbo.Nombre, detalle: detalle})
	}

	return entradas
}

// bloque escribe una lista de la ayuda con sus nombres alineados. El ancho sale
// del nombre más largo y se mide en runas, no en bytes, porque un nombre con
// tilde ocupa más bytes que posiciones y la lista saldría descuadrada.
func bloque(titulo string, entradas []entrada) string {
	var lista strings.Builder

	ancho := 0
	for _, e := range entradas {
		ancho = max(ancho, utf8.RuneCountInString(e.nombre))
	}

	fmt.Fprintf(&lista, "\n%s\n", titulo)

	for _, e := range entradas {
		// La línea se recorta por la derecha: un applet o un verbo sin
		// descripción dejaría si no una línea acabada en espacios.
		linea := fmt.Sprintf("  %-*s  %s", ancho, e.nombre, e.detalle)
		lista.WriteString(strings.TrimRight(linea, " ") + "\n")
	}

	return lista.String()
}
