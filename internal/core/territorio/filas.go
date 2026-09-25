package territorio

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

// La relación y la correspondencia son ficheros de filas: una cabecera con su
// fecha y su source, y un mapa de filas indexado por código que ocupa el resto
// del fichero, una fila por línea y en forma de flujo (research.md D4). Así
// los escribe el repositorio, y en esa forma se leen sin pasar las filas por
// el lector de YAML: para miles de filas, el árbol de nodos que él construye y
// su decodificación fila a fila son casi todo lo que cuesta una carga, que el
// binario paga en cada invocación (research.md S3).
//
// Leerlas así no cambia nada de lo que da la carga. Un fichero está en su
// forma solo si el lector de YAML daría con él exactamente la misma cabecera y
// las mismas filas; cualquier otro —otra sangría, otro orden de los campos, un
// comentario, un escape que no es de los dos, un carácter que el lector trata
// aparte— lo lee entero el lector de YAML, como antes, con los mismos defectos.

// Lo que fija la forma de una línea de filas.
const (
	// sangriaDeLasFilas es la de cada fila bajo la clave de las filas.
	sangriaDeLasFilas = "  "
	// longitudMaximaDeLaClave es lo más que ocupa, en bytes, lo escrito entre
	// las comillas de la clave de una fila. El lector de YAML exige que los
	// dos puntos de una clave implícita lleguen como mucho 1024 caracteres
	// después de su comienzo, la comilla que la abre, y una clave más larga es
	// un error suyo; cada carácter ocupa al menos un byte.
	longitudMaximaDeLaClave = 1024 - len(`""`)
)

// Los separadores de línea y de párrafo de Unicode, U+2028 y U+2029: el
// lector de YAML los admite, pero son saltos de línea para él.
const (
	separadorDeLinea   = 0x2028
	separadorDeParrafo = 0x2029
)

// cabeceraDeFilas es la cabecera de un fichero de filas, C, que el lector de
// YAML decodifica con las filas en un nodo que no decodifica.
type cabeceraDeFilas[C any] interface {
	*C
	// nodoDeLasFilas es ese nodo.
	nodoDeLasFilas() *yaml.Node
}

// nodoDeLasFilas es el nodo de las filas de la relación.
func (r *relacionSinFilas) nodoDeLasFilas() *yaml.Node {
	return &r.Municipios
}

// nodoDeLasFilas es el nodo de las filas de la correspondencia.
func (c *correspondenciaSinFilas) nodoDeLasFilas() *yaml.Node {
	return &c.Correspondencia
}

// decodificarFicheroDeFilas lee un fichero de filas: la cabecera en C y las
// filas, cada una con leerFila, en un mapa por su clave. Si el fichero está en
// su forma, las filas se leen sin el lector de YAML; si no, el lector lo lee
// entero.
func decodificarFicheroDeFilas[C any, PC cabeceraDeFilas[C], V any](
	ruta, clave string, contenido []byte, leerFila func(string) (V, bool),
) (C, map[string]V, error) {
	if cabecera, filas, enSuForma := leerEnSuForma[C, PC](ruta, clave, contenido, leerFila); enSuForma {
		leidas, err := filas.resultado()

		return cabecera, leidas, err
	}

	return decodificarConElLector[C, PC, V](ruta, clave, contenido)
}

// decodificarConElLector lee un fichero de filas entero con el lector de YAML,
// esté o no en su forma: la cabecera en C y el nodo de las filas con
// decodificarFilas.
func decodificarConElLector[C any, PC cabeceraDeFilas[C], V any](
	ruta, clave string, contenido []byte,
) (C, map[string]V, error) {
	cabecera, err := decodificarFichero[C](ruta, contenido)
	if err != nil {
		var vacia C

		return vacia, nil, err
	}

	filas, err := decodificarFilas[V](ruta, clave, PC(&cabecera).nodoDeLasFilas())

	return cabecera, filas, err
}

// leerEnSuForma lee un fichero de filas que está en su forma: la cabecera con
// el lector de YAML y cada fila con leerFila, sin él. Si el fichero no está en
// su forma, devuelve false y nada más: es el lector de YAML el que tiene que
// leerlo entero.
//
// En su forma, la cabecera va hasta una línea que es exactamente la clave de
// las filas con sus dos puntos, y el lector de YAML la decodifica sin defectos
// y con esa clave sin valor en esa misma línea; en el fichero entero, su valor
// es entonces el mapa de filas que empieza en la línea siguiente, y la
// cabecera se lee igual con él que sin él. Cada línea que sigue, hasta el
// final, es una fila en su forma (leerLinea), y hay al menos una.
func leerEnSuForma[C any, PC cabeceraDeFilas[C], V any](
	ruta, clave string, contenido []byte, leerFila func(string) (V, bool),
) (C, *filasLeidas[V], bool) {
	var vacia C

	fichero, separado := separarFilas(contenido, clave)
	if !separado {
		return vacia, nil, false
	}

	cabecera, err := decodificarFichero[C](ruta, fichero.cabecera)
	if err != nil || !esLaClaveSinValor(PC(&cabecera).nodoDeLasFilas(), fichero.lineaDeLaClave, clave) {
		return vacia, nil, false
	}

	filas := nuevasFilasLeidas[V](ruta, clave, strings.Count(fichero.filas, "\n")+1)
	numero := fichero.lineaDeLaClave

	for resto := fichero.filas; resto != ""; {
		var linea string

		linea, resto, _ = strings.Cut(resto, "\n")
		numero++

		codigo, fila, enSuForma := leerLinea(linea, leerFila)
		if !enSuForma {
			return vacia, nil, false
		}

		if filas.nueva(codigo, numero) {
			filas.filas[codigo] = fila
		}
	}

	return cabecera, filas, true
}

// ficheroDeFilas es un fichero de filas separado por la línea de la clave de
// las filas.
type ficheroDeFilas struct {
	// cabecera es el fichero hasta esa línea, incluida.
	cabecera []byte
	// lineaDeLaClave es el número de esa línea, desde 1.
	lineaDeLaClave int
	// filas es todo lo que la sigue, en una sola copia de la que cada texto
	// leído es un tramo.
	filas string
}

// separarFilas separa un fichero de filas por la primera línea que es
// exactamente la clave de las filas con sus dos puntos. Sin esa línea, o sin
// nada detrás, devuelve false.
func separarFilas(contenido []byte, clave string) (ficheroDeFilas, bool) {
	marca := clave + ":"

	for inicio, numero := 0, 1; ; numero++ {
		fin := bytes.IndexByte(contenido[inicio:], '\n')
		if fin < 0 {
			return ficheroDeFilas{}, false
		}

		fin += inicio

		if string(contenido[inicio:fin]) == marca {
			if fin+1 == len(contenido) {
				return ficheroDeFilas{}, false
			}

			return ficheroDeFilas{
				cabecera:       contenido[:fin+1],
				lineaDeLaClave: numero,
				filas:          string(contenido[fin+1:]),
			}, true
		}

		inicio = fin + 1
	}
}

// esLaClaveSinValor dice si el nodo de las filas que el lector de YAML ve en
// la cabecera es el valor vacío de la clave de su última línea: un escalar
// justo después de sus dos puntos, que el lector sitúa en la línea y la
// columna, desde 1, del final de ese indicador. Si no lo es —la cabecera
// tiene varios documentos, o el lector cuenta como salto de línea algo que no
// es el de siempre—, la clave de las filas no es la de esa línea para él.
func esLaClaveSinValor(nodo *yaml.Node, linea int, clave string) bool {
	return nodo.Kind == yaml.ScalarNode && nodo.Line == linea && nodo.Column == len(clave)+len(":")+1
}

// leerLinea lee una línea de filas en su forma: la sangría, la clave entre
// comillas dobles, los dos puntos con un espacio y la fila, que lee leerFila
// de todo lo que queda de la línea. Si la línea no está en su forma, devuelve
// false.
func leerLinea[V any](linea string, leerFila func(string) (V, bool)) (clave string, fila V, enSuForma bool) {
	lector := nuevoLectorDeFila(linea)

	lector.literal(sangriaDeLasFilas)
	clave = lector.clave()
	lector.literal(": ")

	if !lector.enSuForma {
		return "", fila, false
	}

	fila, enSuForma = leerFila(lector.resto)

	return clave, fila, enSuForma
}

// leerFilaDeMunicipio lee la fila de un municipio de la relación: un mapa en
// forma de flujo con sus cuatro campos en su orden, cada uno con su texto
// entre comillas dobles. Si la fila no está en su forma, devuelve false.
func leerFilaDeMunicipio(escrita string) (FilaDeMunicipio, bool) {
	var fila FilaDeMunicipio

	lector := nuevoLectorDeFila(escrita)

	lector.literal("{dc: ")
	fila.DC = lector.texto()
	lector.literal(", nombre: ")
	fila.Nombre = lector.texto()
	lector.literal(", provincia: ")
	fila.Provincia = lector.texto()
	lector.literal(", comunidad: ")
	fila.Comunidad = lector.texto()
	lector.literal("}")

	return fila, lector.terminada()
}

// leerFilaDeDIR3 lee la fila de un municipio de la correspondencia: su DIR3,
// entre comillas dobles. Si la fila no está en su forma, devuelve false.
func leerFilaDeDIR3(escrita string) (string, bool) {
	lector := nuevoLectorDeFila(escrita)
	dir3 := lector.texto()

	return dir3, lector.terminada()
}

// lectorDeFila lee de izquierda a derecha una línea de filas y comprueba que
// está en su forma. En cuanto algo no lo está, deja de leer: enSuForma
// queda en false, y nada de lo que devuelve desde ahí vale.
type lectorDeFila struct {
	resto     string
	enSuForma bool
}

// nuevoLectorDeFila empieza a leer un texto, que nada ha sacado aún de su
// forma.
func nuevoLectorDeFila(texto string) lectorDeFila {
	return lectorDeFila{resto: texto, enSuForma: true}
}

// literal consume un texto fijo de la forma.
func (l *lectorDeFila) literal(texto string) {
	if !l.enSuForma || !strings.HasPrefix(l.resto, texto) {
		l.enSuForma = false

		return
	}

	l.resto = l.resto[len(texto):]
}

// clave consume la clave de una fila, un texto entre comillas dobles que no
// pasa de longitudMaximaDeLaClave, y devuelve su valor.
func (l *lectorDeFila) clave() string {
	escrito := l.entrecomillado()
	if len(escrito) > longitudMaximaDeLaClave {
		l.enSuForma = false
	}

	return sinEscapes(escrito)
}

// texto consume un texto entre comillas dobles y devuelve su valor.
func (l *lectorDeFila) texto() string {
	return sinEscapes(l.entrecomillado())
}

// entrecomillado consume un texto entre comillas dobles y devuelve lo que hay
// entre ellas tal como está escrito, con sus escapes.
func (l *lectorDeFila) entrecomillado() string {
	l.literal(`"`)

	if !l.enSuForma {
		return ""
	}

	cierre := cierreDelTexto(l.resto)
	if cierre < 0 {
		l.enSuForma = false

		return ""
	}

	escrito := l.resto[:cierre]
	l.resto = l.resto[cierre+len(`"`):]

	return escrito
}

// terminada dice si la línea se ha leído entera y en su forma.
func (l *lectorDeFila) terminada() bool {
	return l.enSuForma && l.resto == ""
}

// cierreDelTexto devuelve la posición de la comilla que cierra un texto entre
// comillas dobles, del que resto es lo que sigue a la que lo abre, si todo lo
// que hay antes está en su forma; si no, -1. En su forma, cada carácter vale
// lo que está escrito (esLiteral), salvo los dos únicos escapes, «\"» y «\\»,
// que valen una comilla doble y una barra invertida.
func cierreDelTexto(resto string) int {
	for indice := 0; indice < len(resto); {
		switch resto[indice] {
		case '"':
			return indice
		case '\\':
			if siguiente := resto[indice+1:]; !strings.HasPrefix(siguiente, `"`) && !strings.HasPrefix(siguiente, `\`) {
				return -1
			}

			indice += len(`\"`)
		default:
			runa, medida := utf8.DecodeRuneInString(resto[indice:])
			if !esLiteral(runa, medida) {
				return -1
			}

			indice += medida
		}
	}

	return -1
}

// esLiteral dice si el lector de YAML toma tal cual una runa dentro de un
// texto entre comillas dobles: es una de las que él admite (yaml.v3,
// readerc.go) y no es de las que cuenta como salto de línea —el de siempre, el
// retorno de carro, NEL y los separadores de línea y de párrafo—, que dentro
// de un texto pliega en un espacio. Una medida de uno con utf8.RuneError es
// un byte que no es UTF-8.
func esLiteral(runa rune, medida int) bool {
	switch {
	case runa == utf8.RuneError && medida == 1:
		return false
	case runa == separadorDeLinea, runa == separadorDeParrafo:
		return false
	case runa == '\t',
		' ' <= runa && runa <= '~',
		0xA0 <= runa && runa <= 0xD7FF,
		0xE000 <= runa && runa <= 0xFFFD,
		runa >= 0x10000:
		return true
	default:
		return false
	}
}

// sinEscapes devuelve el valor de lo escrito entre las comillas de un texto en
// su forma: cada barra invertida, que siempre va seguida de una comilla doble
// o de otra barra invertida, vale lo que la sigue. Sin escapes, lo escrito es
// el valor, y no se copia.
func sinEscapes(escrito string) string {
	if !strings.Contains(escrito, `\`) {
		return escrito
	}

	var valor strings.Builder

	valor.Grow(len(escrito))

	for indice := 0; indice < len(escrito); indice++ {
		if escrito[indice] == '\\' {
			indice++
		}

		valor.WriteByte(escrito[indice])
	}

	return valor.String()
}

// filasLeidas reúne las filas de un fichero de filas en el orden en que las
// escribe, con la línea de cada clave, y los defectos de las que no se leen,
// cada uno con su línea o con su clave. La clave repetida se busca con un
// mapa, fila a fila: el lector de YAML compara cada clave de un mapa con todas
// las demás, y con las miles de filas de estos ficheros esa comparación, ella
// sola, pasa del tiempo que tiene una invocación entera (research.md S3).
type filasLeidas[V any] struct {
	ruta, clave string
	filas       map[string]V
	lineas      map[string]int
	defectos    []error
}

// nuevasFilasLeidas prepara las filas de un fichero de filas, con sitio para
// las que se esperan.
func nuevasFilasLeidas[V any](ruta, clave string, esperadas int) *filasLeidas[V] {
	return &filasLeidas[V]{
		ruta:   ruta,
		clave:  clave,
		filas:  make(map[string]V, esperadas),
		lineas: make(map[string]int, esperadas),
	}
}

// nueva anota la clave de una fila y su línea, y dice si es la primera vez que
// aparece; si no, anota el defecto de la clave repetida.
func (f *filasLeidas[V]) nueva(clave string, linea int) bool {
	if primera, repetida := f.lineas[clave]; repetida {
		f.defecto("la clave %q está repetida en las líneas %d y %d", clave, primera, linea)

		return false
	}

	f.lineas[clave] = linea

	return true
}

// defecto anota un defecto de las filas.
func (f *filasLeidas[V]) defecto(formato string, argumentos ...any) {
	f.defectos = append(f.defectos, defectoDeFilas(f.ruta, f.clave, formato, argumentos...))
}

// resultado devuelve las filas leídas y, juntos, sus defectos.
func (f *filasLeidas[V]) resultado() (map[string]V, error) {
	return f.filas, errors.Join(f.defectos...)
}

// defectoDeFilas es un defecto de las filas de un fichero, que nombra la clave
// de las filas delante.
func defectoDeFilas(ruta, clave, formato string, argumentos ...any) error {
	return &defectoDeCarga{fichero: ruta, motivo: clave + ": " + fmt.Sprintf(formato, argumentos...)}
}
