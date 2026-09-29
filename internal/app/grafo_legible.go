package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
)

// Lo que show, stats y check cuentan a una persona cuando no se pide --json: el
// texto que va en Resultado.Legible y que el kernel escribe en la salida
// estándar en lugar de la tabla mínima (docs/ADR/0026; H7.1 FR-060 a FR-063;
// contracts/applet-graph.md §5). Se compone a partir del mismo data y en el
// mismo instante, así que no puede decir otra cosa que el sobre, y no lleva
// nada del sobre —ni su fuente, ni su url, ni su fecha, ni su huella— ni pares
// ruta/valor. Es determinista, sin tabuladores ni secuencias de escape, con lo
// que va en columna alineado con espacios por el número de runas y con el salto
// de línea al final.

// paraAcotarLaComprobacion es la línea con la que termina check sin
// argumentos: cómo acotarlo a una norma y a sus bloques (FR-063).
const paraAcotarLaComprobacion = "Para acotar la comprobación a una norma y a sus bloques: kitlegal graph check" +
	" <norma> [<bloque>...]"

// Las sangrías del texto: dos espacios para cada dato, par, arista o hallazgo,
// y cuatro para lo que se dice debajo de una arista o de un hallazgo.
const (
	sangriaDelGrafo  = "  "
	sangriaDeDetalle = "    "
)

// separacionEnColumna son los espacios que separan lo que stats alinea en
// columna: los mismos que deja la tabla mínima.
const separacionEnColumna = "  "

// legibleDeStats es lo que stats cuenta (FR-061; §5.1): cuántos nodos, aristas
// y textos hay, cada número con su nombre en singular o en plural, y debajo una
// sección por grupo con algún par —los nodos por tipo y fuente, las aristas
// por relación y fuente— con cada recuento. El grafo vacío es solo la primera
// frase.
func legibleDeStats(recuento grafo.Recuento) string {
	var b strings.Builder

	fmt.Fprintf(&b, "El grafo del mundo tiene %s, %s y %s.\n",
		cantidad(recuento.Nodos, "nodo", "nodos"),
		cantidad(recuento.Aristas, "arista", "aristas"),
		cantidad(recuento.Textos, "texto", "textos"))

	nodos := make([]parContado, 0, len(recuento.NodosPorTipo))
	for _, par := range recuento.NodosPorTipo {
		nodos = append(nodos, parContado{clave: presentable(par.Tipo), fuente: presentable(par.Fuente), n: par.Nodos})
	}

	aristas := make([]parContado, 0, len(recuento.AristasPorRelacion))
	for _, par := range recuento.AristasPorRelacion {
		aristas = append(aristas, parContado{
			clave: presentable(par.Relacion), fuente: presentable(par.Fuente), n: par.Aristas,
		})
	}

	seccionDeRecuentos(&b, "Nodos por tipo y fuente:", nodos)
	seccionDeRecuentos(&b, "Aristas por relación y fuente:", aristas)

	return b.String()
}

// parContado es una línea de una sección de stats: lo que se cuenta —un tipo o
// una relación— y su fuente, ya presentables, y cuántos son.
type parContado struct {
	clave  string
	fuente string
	n      int
}

// seccionDeRecuentos escribe, tras una línea en blanco, el título y un par por
// línea, con lo contado y su fuente alineados en columna por runas y el número
// al final. Sin pares no escribe nada.
func seccionDeRecuentos(b *strings.Builder, titulo string, pares []parContado) {
	if len(pares) == 0 {
		return
	}

	anchoDeClave, anchoDeFuente := 0, 0

	for _, par := range pares {
		anchoDeClave = max(anchoDeClave, utf8.RuneCountInString(par.clave))
		anchoDeFuente = max(anchoDeFuente, utf8.RuneCountInString(par.fuente))
	}

	fmt.Fprintf(b, "\n%s\n", titulo)

	for _, par := range pares {
		fmt.Fprintf(b, "%s%-*s%s%-*s%s%d\n", sangriaDelGrafo, anchoDeClave, par.clave, separacionEnColumna,
			anchoDeFuente, par.fuente, separacionEnColumna, par.n)
	}
}

// cantidad es el número con el nombre de lo contado, en singular si es uno y
// en plural si no («1 nodo», «0 nodos»).
func cantidad(n int, singular, plural string) string {
	if n == 1 {
		return "1 " + singular
	}

	return strconv.Itoa(n) + " " + plural
}

// legibleDeShow es lo que show cuenta de un nodo (FR-062; §5.2): su tipo y su
// id; un dato por línea, en el orden de sus claves comparando bytes, y el valor
// que no es una cadena, en JSON; su primera observación y la última, con su
// fuente y su url; y sus aristas salientes y entrantes, en el orden de data,
// cada una con su relación, el otro extremo y su última observación debajo, o
// «ninguna».
//
// Un dato que no se puede escribir en JSON no es algo que world.db pueda
// guardar —los datos se leen de JSON—, así que es un fallo inesperado y no un
// texto a medias.
func legibleDeShow(ficha grafo.Ficha) (string, error) {
	var b strings.Builder

	nodo := ficha.Nodo
	fmt.Fprintf(&b, "%s %s\n", presentable(nodo.Tipo), presentable(nodo.ID))

	for _, clave := range slices.Sorted(maps.Keys(nodo.Datos)) {
		valor, err := valorDeDato(nodo.Datos[clave])
		if err != nil {
			return "", fmt.Errorf("graph: el dato %q de %q no se puede escribir: %w", clave, nodo.ID, err)
		}

		fmt.Fprintf(&b, "%s%s: %s\n", sangriaDelGrafo, presentable(clave), valor)
	}

	fmt.Fprintf(&b, "Primera observación: %s\n", presentable(nodo.PrimeraObservacion))
	fmt.Fprintf(&b, "Última observación: %s\n", observacionLegible(nodo.UltimaObservacion))

	aristasLegibles(&b, "Aristas salientes", "→", ficha.Salientes)
	aristasLegibles(&b, "Aristas entrantes", "←", ficha.Entrantes)

	return b.String(), nil
}

// aristasLegibles escribe, tras una línea en blanco, las aristas de un sentido
// con la flecha de ese sentido, o que no hay ninguna.
func aristasLegibles(b *strings.Builder, titulo, flecha string, aristas []grafo.AristaDeFicha) {
	if len(aristas) == 0 {
		fmt.Fprintf(b, "\n%s: ninguna.\n", titulo)

		return
	}

	fmt.Fprintf(b, "\n%s:\n", titulo)

	for _, arista := range aristas {
		fmt.Fprintf(b, "%s%s %s %s\n", sangriaDelGrafo, presentable(arista.Relacion), flecha, presentable(arista.ID))
		fmt.Fprintf(b, "%súltima observación: %s\n", sangriaDeDetalle, observacionLegible(arista.UltimaObservacion))
	}
}

// observacionLegible es una procedencia contada en una línea: la fecha de
// consulta, la fuente y la url, separadas por un punto medio.
func observacionLegible(procedencia grafo.Procedencia) string {
	return presentable(procedencia.FechaConsulta) + " · " + presentable(procedencia.Fuente) + " · " +
		presentable(procedencia.URL)
}

// valorDeDato es el valor de un dato tal como se presenta: una cadena, tal
// cual; cualquier otro valor, en JSON, sin escapar «&», «<» ni «>», que aquí no
// van a ningún HTML.
func valorDeDato(valor any) (string, error) {
	if cadena, esCadena := valor.(string); esCadena {
		return presentable(cadena), nil
	}

	var escrito bytes.Buffer

	codificador := json.NewEncoder(&escrito)
	codificador.SetEscapeHTML(false)

	if err := codificador.Encode(valor); err != nil {
		return "", err
	}

	return presentable(strings.TrimSuffix(escrito.String(), "\n")), nil
}

// legibleDeCheck es lo que check cuenta (FR-063; §5.3). Con hallazgos, la
// cabecera con su ámbito, el total de cada clase y cuántos se listan y cuántos
// se omiten, también si no se omite ninguno; y un grupo por clase con algún
// hallazgo listado, version-obsoleta primero, con el total de la clase y cada
// hallazgo con su explicación y, debajo, su id. Sin hallazgos, una frase que
// dice que no hay nada que volver a comprobar en ese ámbito. Sin argumentos,
// con hallazgos o sin ellos, termina diciendo cómo acotar la comprobación.
func legibleDeCheck(comprobacion grafo.Comprobacion) string {
	var b strings.Builder

	ambito := ambitoLegible(comprobacion.Norma, comprobacion.Bloques)

	if comprobacion.VersionObsoleta+comprobacion.FuenteCaducada == 0 {
		fmt.Fprintf(&b, "No hay nada que volver a comprobar %s.\n", ambito)
	} else {
		fmt.Fprintf(&b, "Hallazgos %s: %d %s y %d %s; %s y %s.\n\n", ambito,
			comprobacion.VersionObsoleta, grafo.ClaseVersionObsoleta,
			comprobacion.FuenteCaducada, grafo.ClaseFuenteCaducada,
			conNumero(len(comprobacion.Hallazgos), "se lista", "se listan"),
			conNumero(comprobacion.Omitidos, "se omite", "se omiten"))

		grupoDeHallazgos(&b, grafo.ClaseVersionObsoleta, comprobacion.VersionObsoleta, comprobacion.Hallazgos)
		grupoDeHallazgos(&b, grafo.ClaseFuenteCaducada, comprobacion.FuenteCaducada, comprobacion.Hallazgos)
	}

	if comprobacion.Norma == "" {
		b.WriteString("\n" + paraAcotarLaComprobacion + "\n")
	}

	return b.String()
}

// grupoDeHallazgos escribe los hallazgos listados de una clase, en su orden,
// bajo su nombre y el total de la clase, que cuenta también los omitidos. Una
// clase sin ninguno listado no escribe nada.
func grupoDeHallazgos(b *strings.Builder, clase grafo.ClaseDeHallazgo, total int, hallazgos []grafo.Hallazgo) {
	cabecera := false

	for _, hallazgo := range hallazgos {
		if hallazgo.Clase != clase {
			continue
		}

		if !cabecera {
			fmt.Fprintf(b, "%s (%d):\n", clase, total)
			cabecera = true
		}

		fmt.Fprintf(b, "%s- %s\n", sangriaDelGrafo, presentable(hallazgo.Explicacion))
		fmt.Fprintf(b, "%s%s\n", sangriaDeDetalle, presentable(hallazgo.ID))
	}
}

// conNumero es el verbo seguido del número, concertado con él: en singular si
// es uno y en plural si no («se lista 1», «se omiten 0»).
func conNumero(n int, singular, plural string) string {
	if n == 1 {
		return singular + " 1"
	}

	return plural + " " + strconv.Itoa(n)
}

// ambitoLegible es el ámbito de la comprobación tal como se nombra detrás de
// «Hallazgos» o de «volver a comprobar»: «en todo lo consultado» sin norma;
// «de <norma>» con ella, y detrás, si se pidieron, sus bloques en su orden:
// «bloque a21», «bloques a21 y a22», «bloques a21, a22 y a23».
func ambitoLegible(norma string, bloques []string) string {
	if norma == "" {
		return "en todo lo consultado"
	}

	ambito := "de " + presentable(norma)

	nombrados := make([]string, 0, len(bloques))
	for _, bloque := range bloques {
		nombrados = append(nombrados, presentable(bloque))
	}

	switch len(nombrados) {
	case 0:
		return ambito
	case 1:
		return ambito + ", bloque " + nombrados[0]
	}

	ultimo := len(nombrados) - 1

	return ambito + ", bloques " + strings.Join(nombrados[:ultimo], ", ") + " y " + nombrados[ultimo]
}

// presentable es un valor de data tal como entra en el texto: tal cual, salvo
// que lleve algún carácter de control —la categoría Cc de Unicode,
// unicode.IsControl: el tabulador, los saltos, ESC, DEL, los C1…—, que partiría
// la línea o se tomaría por una secuencia de escape; entonces va entre comillas
// y con esos caracteres escapados, como lo nombran los mensajes de error (%q).
// Ninguno de los que escribe el propio binario los lleva; sí puede un bloque
// pedido, que solo se rechaza vacío o en blanco.
func presentable(valor string) string {
	if strings.ContainsFunc(valor, unicode.IsControl) {
		return strconv.Quote(valor)
	}

	return valor
}
