package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// separadorRuta une las claves y los índices de la ruta de un valor:
// `articulos.0.titulo` (research.md D15).
const separadorRuta = "."

// claveData nombra el contenido cuando no es compuesto y por tanto no tiene
// ruta propia: es la clave que ocupa en el sobre.
const claveData = "data"

// relleno son los espacios que separan la ruta del valor. Dos es lo mínimo que
// deja ver la columna como columna.
const relleno = 2

// fila es un par ruta/valor de la tabla: la ruta de un dato y el dato ya
// convertido a texto.
type fila struct {
	ruta  string
	valor string
}

// escribirTabla emite la tabla mínima: las cuatro líneas de procedencia
// —fuente, url, fecha_consulta y hash— seguidas del contenido de data aplanado
// a pares ruta/valor, de modo que la cita no se pierda por elegir la forma
// legible por una persona (FR-041, FR-043, contracts/sobre-de-salida.md §7).
//
// Toda escritura comprueba su error, incluido el vaciado final: el tabwriter
// retiene lo escrito hasta conocer el ancho de la columna, así que un
// descriptor roto se manifiesta casi siempre ahí y solo ahí (research.md D27).
func escribirTabla(destino io.Writer, sobre schema.Sobre) error {
	filas, err := filasDelSobre(sobre)
	if err != nil {
		return err
	}

	tabla := tabwriter.NewWriter(destino, 0, 0, relleno, ' ', 0)
	for _, f := range filas {
		if _, err := fmt.Fprintf(tabla, "%s\t%s\n", f.ruta, f.valor); err != nil {
			return err
		}
	}

	return tabla.Flush()
}

// filasDelSobre es la tabla completa: primero la procedencia, que va siempre y
// en este orden, y después el contenido.
//
// La fecha se escribe con el mismo formato con el que viaja en el sobre —RFC
// 3339 con desplazamiento horario explícito y sin perder la precisión que
// traiga—, para que la forma legible y la legible por máquina citen el mismo
// instante.
func filasDelSobre(sobre schema.Sobre) ([]fila, error) {
	contenido, err := generico(sobre.Data)
	if err != nil {
		return nil, err
	}

	filas := []fila{
		{ruta: "fuente", valor: sobre.Fuente},
		{ruta: "url", valor: sobre.URL},
		{ruta: "fecha_consulta", valor: sobre.FechaConsulta.Format(time.RFC3339Nano)},
		{ruta: "hash", valor: sobre.Hash},
	}

	return aplanar(filas, "", contenido), nil
}

// generico lleva el contenido del applet a la forma genérica que se puede
// recorrer sin conocer su tipo: mapas, listas y valores sueltos, con los
// números conservados como literales para que un expediente grande no pase por
// la coma flotante y pierda precisión.
//
// Un contenido que no se puede serializar devuelve error y no un pánico, igual
// que el cálculo de la huella: el sobre no llega a la salida estándar a medias
// (FR-033, contracts/sobre-de-salida.md §4).
func generico(data any) (any, error) {
	crudo, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("render: el contenido de data no se puede recorrer: %w", err)
	}

	decodificador := json.NewDecoder(bytes.NewReader(crudo))
	decodificador.UseNumber()

	var valor any
	if err := decodificador.Decode(&valor); err != nil {
		return nil, fmt.Errorf("render: el contenido de data no se puede recorrer: %w", err)
	}

	return valor, nil
}

// aplanar va recorriendo el contenido y añade una fila por cada valor que no
// sea compuesto. Los objetos y las listas no se imprimen: se recorren, y lo que
// dejan es la ruta con la que se llega a sus hojas.
//
// Las claves de un objeto se recorren en orden alfabético, que es el mismo
// criterio con el que el dominio ordena la forma canónica de la que sale la
// huella: dos invocaciones con el mismo contenido presentan la misma tabla.
func aplanar(filas []fila, ruta string, valor any) []fila {
	switch v := valor.(type) {
	case map[string]any:
		for _, clave := range slices.Sorted(maps.Keys(v)) {
			filas = aplanar(filas, unirRuta(ruta, clave), v[clave])
		}
	case []any:
		for i, elemento := range v {
			filas = aplanar(filas, unirRuta(ruta, strconv.Itoa(i)), elemento)
		}
	default:
		filas = append(filas, fila{ruta: rutaDeHoja(ruta), valor: textoDeHoja(v)})
	}

	return filas
}

// unirRuta concatena la ruta con la clave o el índice que la prolonga. El
// primer tramo no lleva separador delante: el contenido no se anuncia como
// `data.`, porque la tabla presenta el contenido y no el sobre.
func unirRuta(ruta, tramo string) string {
	if ruta == "" {
		return tramo
	}

	return ruta + separadorRuta + tramo
}

// rutaDeHoja nombra la hoja que no tiene ruta, que es el contenido que no es
// compuesto: un data que sea una cadena, un número o un booleano. Se nombra por
// la clave que ocupa en el sobre, porque una fila sin ruta no diría qué es.
func rutaDeHoja(ruta string) string {
	if ruta == "" {
		return claveData
	}

	return ruta
}

// textoDeHoja convierte un valor que no es compuesto en lo que se ve en la
// tabla: la cadena tal cual —sin comillas, que aquí estorban— y el literal JSON
// de todo lo demás, de modo que un nulo se distinga de una cadena vacía.
func textoDeHoja(valor any) string {
	switch v := valor.(type) {
	case nil:
		return "null"
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case json.Number:
		return v.String()
	default:
		// El decodificador no produce ningún otro tipo. Escribirlo de todos
		// modos evita que un cambio futuro en encoding/json convierta esta fila
		// en un valor vacío sin que nadie se entere.
		return fmt.Sprint(v)
	}
}
