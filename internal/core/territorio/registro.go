package territorio

import (
	"maps"
	"slices"
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

// MunicipioConFormas es un municipio del registro con las formas de su nombre,
// lo que necesita quien cruza con la relación los nombres que escribe otra
// fuente, como un boletín (research.md D13).
type MunicipioConFormas struct {
	// Codigo es su código INE.
	Codigo ids.CodigoINE
	// Nombre es el oficial, tal como lo escribe la relación.
	Nombre string
	// Formas son las formas conocidas de su nombre, plegadas y sin repetir:
	// las mismas por las que lo alcanza una consulta (formasDelNombre).
	Formas []string
}

// MunicipiosDe son los municipios de una comunidad o ciudad autónoma, dada por
// sus dos cifras, en orden de código INE y cada uno con sus formas: ninguno si
// el registro no la tiene (FR-022). Cada llamada los da nuevos, de modo que
// quien los recibe puede cambiarlos sin tocar el registro.
//
// No es parte de la carga, que paga cada invocación del binario: mira el
// registro entero en cada llamada, y es para las derivaciones de los festivos.
func (r *Registro) MunicipiosDe(comunidad string) []MunicipioConFormas {
	var codigos []string

	for codigo, municipio := range r.porCodigo {
		if municipio.provincia.comunidad.codigo == comunidad {
			codigos = append(codigos, codigo)
		}
	}

	slices.Sort(codigos)

	municipios := make([]MunicipioConFormas, 0, len(codigos))
	for _, codigo := range codigos {
		municipio := r.porCodigo[codigo]
		municipios = append(municipios, MunicipioConFormas{
			Codigo: municipio.codigo,
			Nombre: municipio.nombre,
			Formas: formasDelNombre(municipio.nombre),
		})
	}

	return municipios
}

// Comunidades son las comunidades y ciudades autónomas del registro, cada una
// con su código, su nombre y el source de su fichero, en orden de código: las
// de sus municipios, que en la relación son todas (research.md D14, S2).
//
// Como MunicipiosDe, mira el registro entero y no es parte de la carga.
func (r *Registro) Comunidades() []Comunidad {
	porCodigo := map[string]*comunidadRegistrada{}
	for _, municipio := range r.porCodigo {
		comunidad := municipio.provincia.comunidad
		porCodigo[comunidad.codigo] = comunidad
	}

	comunidades := make([]Comunidad, 0, len(porCodigo))
	for _, codigo := range slices.Sorted(maps.Keys(porCodigo)) {
		comunidad := porCodigo[codigo]
		comunidades = append(comunidades, Comunidad{
			Codigo: comunidad.codigo,
			Nombre: comunidad.nombre,
			Source: comunidad.origen.source,
		})
	}

	return comunidades
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
// por nombre en orden de código INE, venga en el orden que venga la lista, de
// modo que los candidatos de toda forma quedan en ese orden (FR-014).
//
// El orden sale de las claves del índice por código: cada código se escribe
// una sola vez, al indexarlo, y no en cada comparación de una ordenación, que
// con los miles de municipios de la relación es una buena parte de lo que
// cuesta una carga (research.md S3).
func (r *Registro) indexar(municipios []*municipioRegistrado) {
	r.porCodigo = make(map[string]*municipioRegistrado, len(municipios))
	r.porNombre = map[string][]*municipioRegistrado{}

	for _, municipio := range municipios {
		r.porCodigo[municipio.codigo.String()] = municipio
	}

	for _, codigo := range slices.Sorted(maps.Keys(r.porCodigo)) {
		municipio := r.porCodigo[codigo]

		for _, forma := range formasDelNombre(municipio.nombre) {
			r.porNombre[forma] = append(r.porNombre[forma], municipio)
		}
	}
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
