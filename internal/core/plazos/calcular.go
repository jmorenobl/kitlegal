package plazos

import (
	"slices"
	"strconv"
	"time"
)

// MunicipioResuelto es un municipio de la petición tal como lo da territorio
// resolver: el del calendario, con el código de su comunidad, y el nombre
// oficial de esa comunidad, que nombran los avisos de los festivos locales y
// los de una parte de la comunidad (contracts/applet-plazos.md §5).
type MunicipioResuelto struct {
	Municipio
	// NombreDeLaComunidad es el que da el fichero de su comunidad en
	// data/territorio/.
	NombreDeLaComunidad string
}

// Peticion es lo que pide quien llama (data-model §1): el tipo, la fecha de
// inicio, la duración y su unidad tal como las escribe, cadenas cuya validez
// decide el dominio (research D2), y los municipios que cuentan, ya resueltos
// por territorio. Una duración o una unidad vacías son las que no se dan.
type Peticion struct {
	// Tipo es uno de los códigos de data/plazos.yaml.
	Tipo string
	// Inicio es la fecha AAAA-MM-DD de la notificación o la publicación o, en el
	// silencio, la del acuerdo de iniciación o la de entrada de la solicitud.
	Inicio string
	// Duracion es un entero positivo, solo en los tipos que no fijan la suya.
	Duracion string
	// Unidad es la de Duracion, una de Unidades.
	Unidad string
	// Municipios son los que cuentan, en el orden dado: en un plazo de la LPAC,
	// el del interesado y el de la sede del órgano; en el contencioso, el de la
	// sede del juzgado. Dos veces el mismo código es un municipio que cuenta.
	Municipios []MunicipioResuelto
}

// peticionComprobada es una petición cuyos argumentos valen: su tipo, su
// fecha de inicio y la duración con que se cuenta, la del tipo o la dada.
type peticionComprobada struct {
	tipo     Tipo
	inicio   time.Time
	duracion Duracion
}

// ComprobarPeticion comprueba, en el orden de research D10, lo que la petición
// pide y no depende de resolver ningún municipio: el tipo, la fecha de inicio,
// la duración, la unidad y cuántos municipios son, de los que solo cuenta el
// número. Es lo que el applet comprueba antes de resolver los municipios con
// territorio, para que un argumento que no vale salga con 2 aunque algún
// municipio no exista (FR-012). Devuelve el primer defecto, de clase
// argumentos, o nada; Calcular hace la misma comprobación.
func ComprobarPeticion(tipos []Tipo, peticion Peticion) error {
	_, err := comprobarPeticion(tipos, peticion)

	return err
}

// Calcular da el plazo que pide la petición con los festivos del calendario y
// los tipos de data/plazos.yaml (FR-001 a FR-012; contracts/applet-plazos.md
// §2 y §4). No resuelve ningún municipio: los recibe resueltos. Un argumento
// que no vale, o un plazo que acaba después del 31 de diciembre de 9999, es un
// error de clase argumentos y no da plazo. Un plazo que empieza antes del día
// desde el que rige, como lo cuenta data/plazos.yaml, alguno de sus preceptos
// es un error de clase no encontrado que los nombra, y tampoco da plazo: un
// plazo se rige por la redacción vigente el día en que empieza, y kitlegal no
// tiene otra (ADR 0042).
func Calcular(calendario *Calendario, tipos []Tipo, peticion Peticion) (Plazo, error) {
	comprobada, err := comprobarPeticion(tipos, peticion)
	if err != nil {
		return Plazo{}, err
	}

	if posteriores := preceptosPosterioresAlInicio(comprobada); len(posteriores) > 0 {
		return Plazo{}, plazoConPreceptosPosteriores(comprobada, posteriores)
	}

	plazo := nuevoCalculo(calendario, comprobada, peticion.Municipios)

	resultado, cabe := plazo.computar(comprobada.inicio, comprobada.duracion)
	if !cabe {
		return Plazo{}, plazoFueraDelCalendario(peticion)
	}

	plazo.hasta = resultado.dia

	return plazo.salida(comprobada.duracion.Unidad, resultado), nil
}

// comprobarPeticion lee los argumentos de la petición en el orden de research
// D10 y se detiene en el primero que no vale.
func comprobarPeticion(tipos []Tipo, peticion Peticion) (peticionComprobada, error) {
	indice := slices.IndexFunc(tipos, func(tipo Tipo) bool { return tipo.Codigo == peticion.Tipo })
	if indice < 0 {
		return peticionComprobada{}, tipoDesconocido(peticion.Tipo, tipos)
	}

	tipo := tipos[indice]

	inicio, err := leerInicio(peticion.Inicio)
	if err != nil {
		return peticionComprobada{}, err
	}

	duracion, err := leerDuracion(tipo, peticion.Duracion, peticion.Unidad)
	if err != nil {
		return peticionComprobada{}, err
	}

	if len(peticion.Municipios) != municipiosDelComputo(tipo) {
		return peticionComprobada{}, municipiosQueNoCuentan(tipo, len(peticion.Municipios))
	}

	return peticionComprobada{tipo: tipo, inicio: inicio, duracion: duracion}, nil
}

// preceptosPosterioresAlInicio son los preceptos del plazo —los del tipo, el de
// la unidad en el genérico y los de su cómputo, en ese orden y cada apartado
// una vez— que rigen, como los cuenta data/plazos.yaml, desde un día posterior
// al de inicio. Los días se comparan como texto: los dos son AAAA-MM-DD.
func preceptosPosterioresAlInicio(comprobada peticionComprobada) []Precepto {
	inicio := comprobada.inicio.Format(time.DateOnly)

	var del []Precepto

	del = append(del, comprobada.tipo.Preceptos...)
	if precepto, esta := comprobada.tipo.Unidades[comprobada.duracion.Unidad]; esta {
		del = append(del, precepto)
	}

	del = append(del, comprobada.tipo.PreceptosDelComputo...)

	var posteriores []Precepto

	for _, precepto := range del {
		if precepto.VigenteDesde <= inicio || slices.ContainsFunc(posteriores, func(visto Precepto) bool {
			return visto.Norma == precepto.Norma && visto.Bloque == precepto.Bloque && visto.Apartado == precepto.Apartado
		}) {
			continue
		}

		posteriores = append(posteriores, precepto)
	}

	return posteriores
}

// municipiosDelComputo es cuántos municipios cuentan en el cómputo de un
// tipo: uno, la sede del juzgado, en el contencioso (FR-010), y dos en la LPAC
// (FR-007).
func municipiosDelComputo(tipo Tipo) int {
	if tipo.Computo == ComputoContencioso {
		return 1
	}

	return 2
}

// leerInicio lee la fecha de inicio: primero su forma, AAAA-MM-DD con sus
// ceros, y después que sea un día del calendario.
func leerInicio(texto string) (time.Time, error) {
	if !escritaAAAAMMDD(texto) {
		return time.Time{}, inicioMalEscrito(texto)
	}

	dia, err := time.Parse(time.DateOnly, texto)
	if err != nil {
		return time.Time{}, inicioFueraDelCalendario(texto)
	}

	return dia, nil
}

// escritaAAAAMMDD dice si un texto tiene la forma AAAA-MM-DD: diez caracteres,
// guiones en la quinta y la octava posición y cifras en las demás.
func escritaAAAAMMDD(texto string) bool {
	if len(texto) != len(time.DateOnly) {
		return false
	}

	for indice := range len(texto) {
		caracter := texto[indice]

		switch indice {
		case 4, 7:
			if caracter != '-' {
				return false
			}
		default:
			if caracter < '0' || caracter > '9' {
				return false
			}
		}
	}

	return true
}

// leerDuracion da la duración con que se cuenta el plazo: la del tipo, si la
// fija, que no admite ni duración ni unidad dadas; o la dada, un entero
// positivo con una unidad del vocabulario (research D3 y D10).
func leerDuracion(tipo Tipo, duracion, unidad string) (Duracion, error) {
	if tipo.Duracion != nil {
		switch {
		case duracion != "":
			return Duracion{}, duracionDeMas(tipo.Codigo, duracion)
		case unidad != "":
			return Duracion{}, unidadDeMas(tipo.Codigo, unidad)
		}

		return *tipo.Duracion, nil
	}

	if duracion == "" {
		return Duracion{}, duracionQueFalta(tipo.Codigo)
	}

	cantidad, err := strconv.Atoi(duracion)
	if err != nil || cantidad < 1 {
		return Duracion{}, duracionNoPositiva(duracion)
	}

	switch {
	case unidad == "":
		return Duracion{}, unidadQueFalta(duracion)
	case !slices.Contains(Unidades(), Unidad(unidad)):
		return Duracion{}, unidadDesconocida(unidad)
	}

	return Duracion{Cantidad: cantidad, Unidad: Unidad(unidad)}, nil
}

// calculo es un plazo en cálculo: el calendario, el tipo, los municipios que
// cuentan y su tramo, del día siguiente al de inicio a la fecha que se da, que
// es lo que miran los avisos y la cobertura.
type calculo struct {
	calendario *Calendario
	tipo       Tipo
	// municipios son los que cuentan, una vez cada uno, en el orden dado.
	municipios []MunicipioResuelto
	// delCalendario son los mismos, como los consulta el calendario.
	delCalendario []Municipio
	desde         time.Time
	hasta         time.Time
}

// nuevoCalculo prepara el cálculo de una petición comprobada: los municipios,
// una vez cada uno en el orden dado, y el primer día del tramo.
func nuevoCalculo(calendario *Calendario, comprobada peticionComprobada, municipios []MunicipioResuelto) calculo {
	plazo := calculo{calendario: calendario, tipo: comprobada.tipo, desde: comprobada.inicio.AddDate(0, 0, 1)}

	for _, municipio := range municipios {
		if !slices.ContainsFunc(plazo.municipios, func(visto MunicipioResuelto) bool {
			return visto.CodigoINE == municipio.CodigoINE
		}) {
			plazo.municipios = append(plazo.municipios, municipio)
			plazo.delCalendario = append(plazo.delCalendario, municipio.Municipio)
		}
	}

	return plazo
}
