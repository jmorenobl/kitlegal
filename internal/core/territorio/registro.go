package territorio

import (
	"slices"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/ids"
)

// Registro es el territorio cargado: los municipios de la relación con su
// provincia, su comunidad y, si está verificado, su DIR3, indexados por código
// INE y por cada forma plegada de su nombre. Solo lo construye Cargar, que
// antes comprueba su integridad, y no cambia después: puede usarse desde
// varias goroutines a la vez.
type Registro struct {
	// porCodigo lleva de las cinco cifras del código INE a su municipio.
	porCodigo map[string]*municipioRegistrado
	// porNombre lleva de cada forma plegada a los municipios que la tienen,
	// en orden de código INE.
	porNombre map[string][]*municipioRegistrado
	// estatal es el boletín del estado, el que aplica a todo municipio.
	estatal Boletin
	// relacion, correspondencia y estado son la fecha y el source de los
	// ficheros de municipios, de DIR3 y del estado.
	relacion, correspondencia, estado origen
}

// origen es la fecha, a medianoche UTC, y el source de un fichero congelado.
type origen struct {
	fecha  time.Time
	source string
}

// municipioRegistrado es un municipio de la relación, ya comprobado.
type municipioRegistrado struct {
	codigo ids.CodigoINE
	// digito es el dígito de control oficial, una cifra.
	digito string
	// nombre es el oficial, tal como lo escribe la relación.
	nombre    string
	provincia *provinciaRegistrada
	// dir3 es el de su ayuntamiento si la correspondencia lo trae verificado,
	// o nil.
	dir3 *ids.DIR3
}

// provinciaRegistrada es una provincia declarada por el fichero de su
// comunidad.
type provinciaRegistrada struct {
	codigo    string
	nombre    string
	comunidad *comunidadRegistrada
}

// comunidadRegistrada es una comunidad o ciudad autónoma con lo que declara su
// fichero.
type comunidadRegistrada struct {
	codigo  string
	nombre  string
	regimen string
	// ruta es la de su fichero, el source de lo que fija su configuración.
	ruta   string
	origen origen
	// autonomico y provincial son sus boletines configurados, o nil.
	autonomico, provincial *Boletin
}

// FechaDeLaRelacion es la fecha de la relación de municipios, a medianoche
// UTC. Es la de toda respuesta que no llega a un territorio —un nombre o un
// código que no está en la relación, un nombre que lleva a varios municipios,
// un dígito que no es el oficial o una entrada que no llega a ser consulta—,
// que decide la relación y no el fichero de ninguna comunidad (data-model
// §2.8, contrato del applet §2).
func (r *Registro) FechaDeLaRelacion() time.Time {
	return r.relacion.fecha
}

// indexar construye los dos índices del registro. Los municipios se indexan
// en orden de código INE, venga en el orden que venga la lista, de modo que
// los candidatos de toda forma quedan en ese orden (FR-014).
func (r *Registro) indexar(municipios []*municipioRegistrado) {
	ordenados := slices.SortedFunc(slices.Values(municipios), compararPorCodigo)

	r.porCodigo = make(map[string]*municipioRegistrado, len(ordenados))
	r.porNombre = map[string][]*municipioRegistrado{}

	for _, municipio := range ordenados {
		r.porCodigo[municipio.codigo.String()] = municipio

		for _, forma := range formasDelNombre(municipio.nombre) {
			r.porNombre[forma] = append(r.porNombre[forma], municipio)
		}
	}
}

// compararPorCodigo ordena dos municipios por su código INE.
func compararPorCodigo(a, b *municipioRegistrado) int {
	return strings.Compare(a.codigo.String(), b.codigo.String())
}

// municipio es el municipio de la relación con ese código INE, si lo hay.
func (r *Registro) municipio(codigo ids.CodigoINE) (*municipioRegistrado, bool) {
	municipio, esta := r.porCodigo[codigo.String()]

	return municipio, esta
}

// candidatos son los municipios que tienen una forma plegada, en orden de
// código INE: ninguno, uno o varios.
func (r *Registro) candidatos(forma string) []*municipioRegistrado {
	return r.porNombre[forma]
}
