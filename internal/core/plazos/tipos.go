package plazos

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

// rutaDeLosTipos es la del fichero de los tipos, relativa a la raíz del
// repositorio, la que nombran sus defectos.
const rutaDeLosTipos = "data/plazos.yaml"

// Computo es el código de un cómputo de la sección computos de
// data/plazos.yaml, la regla con que se cuenta un plazo. Los códigos de los
// tipos y de los cómputos son datos: ningún Go los nombra (FR-010 de H36).
type Computo string

// Unidad es la unidad en que se expresa una duración.
type Unidad string

// Las tres unidades de una duración (art. 30 LPAC).
const (
	UnidadMeses         Unidad = "meses"
	UnidadDiasHabiles   Unidad = "dias-habiles"
	UnidadDiasNaturales Unidad = "dias-naturales"
)

// Unidades es el vocabulario de unidades. Devuelve una lista nueva en cada
// llamada.
func Unidades() []Unidad {
	return []Unidad{UnidadMeses, UnidadDiasHabiles, UnidadDiasNaturales}
}

// FijaDuracion es lo que fija el precepto que da la duración de un tipo.
const FijaDuracion = "duracion"

// Lo que da un tipo (FR-005): el día del vencimiento o el siguiente.
const (
	daVencimiento  = "vencimiento"
	daDiaSiguiente = "dia-siguiente"
)

// Los efectos de un periodo del año en el cómputo (spec §C.1 y §C.3): sus días
// son inhábiles, no corren o solo dan un aviso.
const (
	efectoInhabil = "inhabil"
	efectoNoCorre = "no-corre"
	efectoAviso   = "aviso"
)

// gramaticaDeUnCodigo es la de un código de tipo, de cómputo o de aviso de un
// periodo (spec §C.8): una o más palabras de minúsculas ASCII y cifras unidas
// por guiones, la de schemas/eval.yaml.json.
var gramaticaDeUnCodigo = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// noEsUnCodigo es lo que dice el defecto de un código que no sigue la
// gramática.
const noEsUnCodigo = "no es una o más palabras de minúsculas y cifras unidas por guiones"

// anioDeLosPeriodos es un año que no es bisiesto: un MM-DD de un periodo es un
// día que tienen todos los años si es un día de este.
const anioDeLosPeriodos = 2001

// Tipo es un tipo de plazo de data/plazos.yaml (data-model §4).
type Tipo struct {
	// Codigo es el que lo nombra, único entre los tipos.
	Codigo string `yaml:"codigo"`
	// Nombre es el que lo describe.
	Nombre string `yaml:"nombre"`
	// Computo es el código del cómputo con que se cuenta.
	Computo Computo `yaml:"computo"`
	// Duracion es la que fija el tipo, o nil si la recibe con su unidad.
	Duracion *Duracion `yaml:"duracion"`
	// Da es la fecha que da: vencimiento, la del vencimiento, o dia-siguiente,
	// la del día siguiente (FR-005).
	Da string `yaml:"da"`
	// Preceptos son los que lo fundan, cada uno con lo que fija.
	Preceptos []Precepto `yaml:"preceptos"`
	// Unidades lleva, en un tipo sin duración que la lleve, de cada unidad de
	// su cómputo al precepto que da su cómputo.
	Unidades map[Unidad]Precepto `yaml:"unidades"`
	// PreceptosDelComputo son los de su cómputo, los de la sección computos del
	// fichero, que LeerTipos copia en cada tipo.
	PreceptosDelComputo []Precepto `yaml:"-"`
	// computo es su cómputo entero, el de la sección computos del fichero que
	// nombra Computo, que LeerTipos copia en cada tipo (research D1).
	computo computoDescrito
}

// Duracion es la duración que fija un tipo.
type Duracion struct {
	// Cantidad es un entero de 1 o más.
	Cantidad int `yaml:"cantidad"`
	// Unidad es una de las de su cómputo.
	Unidad Unidad `yaml:"unidad"`
}

// Precepto es un apartado de una norma consolidada del BOE y lo que fija.
type Precepto struct {
	// Norma es el identificador BOE-A-… de la norma.
	Norma string `yaml:"norma"`
	// Bloque es el id del bloque en el índice consolidado de la norma.
	Bloque string `yaml:"bloque"`
	// Apartado es el que se cita, como «122.1».
	Apartado string `yaml:"apartado"`
	// Fija es lo que fija el precepto en el tipo: su duración (FijaDuracion),
	// su inicio, su cómputo…
	Fija string `yaml:"fija"`
	// VigenteDesde es el primer día, AAAA-MM-DD, en que el apartado dice lo que
	// dice en la redacción con que está escrito data/plazos.yaml: un plazo que
	// empieza antes no se calcula con él (ADR 0042).
	VigenteDesde string `yaml:"vigente_desde"`
}

// computoDescrito es un cómputo de la sección computos de data/plazos.yaml: lo
// que cuenta cualquier plazo suyo (data-model §1).
type computoDescrito struct {
	// Codigo es el que lo nombra, único entre los cómputos.
	Codigo Computo `yaml:"codigo"`
	// Municipios son los textos que nombran a cada municipio que cuenta, en el
	// orden en que se dan.
	Municipios []string `yaml:"municipios"`
	// Unidades son las de Unidades que admite.
	Unidades []Unidad `yaml:"unidades"`
	// Periodos son sus periodos del año, en el orden en que dan sus avisos.
	Periodos []periodo `yaml:"periodos"`
	// Preceptos son los que cuentan cualquier plazo suyo, en el orden del
	// fichero.
	Preceptos []Precepto `yaml:"preceptos"`
	// posicion es su índice en la sección computos, que ordena sus avisos
	// entre los de los demás (research D3).
	posicion int
}

// periodo es un periodo del año de un cómputo (data-model §2).
type periodo struct {
	// Desde y Hasta son su primer día y el último, MM-DD; si Hasta es anterior a
	// Desde, el periodo cruza el año.
	Desde string `yaml:"desde"`
	Hasta string `yaml:"hasta"`
	// Efecto es el que tienen sus días en el cómputo: inhabil, no-corre o aviso.
	Efecto string `yaml:"efecto"`
	// Precepto nombra, por su norma y su apartado, el que lo funda, que es uno
	// de los del cómputo o de los de cada uno de sus tipos.
	Precepto *referenciaDeUnPrecepto `yaml:"precepto"`
	// Aviso es el que da un plazo que pasa por él, solo en un periodo de aviso.
	Aviso *avisoDeUnPeriodo `yaml:"aviso"`
}

// referenciaDeUnPrecepto nombra un precepto por su norma y su apartado.
type referenciaDeUnPrecepto struct {
	Norma    string `yaml:"norma"`
	Apartado string `yaml:"apartado"`
}

// nombra dice si la referencia es la de un precepto.
func (r referenciaDeUnPrecepto) nombra(precepto Precepto) bool {
	return precepto.Norma == r.Norma && precepto.Apartado == r.Apartado
}

// avisoDeUnPeriodo es el aviso de un periodo de aviso (data-model §3): en la
// salida, «⚠ <Etiqueta>: <Texto>» con su código.
type avisoDeUnPeriodo struct {
	Codigo   string `yaml:"codigo"`
	Etiqueta string `yaml:"etiqueta"`
	Texto    string `yaml:"texto"`
}

// ficheroDeTipos es data/plazos.yaml.
type ficheroDeTipos struct {
	Tipos    []Tipo            `yaml:"tipos"`
	Computos []computoDescrito `yaml:"computos"`
}

// LeerTipos lee data/plazos.yaml (spec §C.1 y §C.2 de H36). Exige que cada tipo
// lleve un código de la gramática de spec §C.8, único entre los tipos, su
// nombre, un cómputo que está en computos y lo que da (vencimiento o
// dia-siguiente); con duración, una de 1 o más en una unidad que su cómputo
// admite y un precepto que la fija; con unidades, exactamente las de su
// cómputo, cada una con su precepto, y no también duración; y algún precepto.
// Exige que cada cómputo lleve un código de la gramática, único entre los
// cómputos, uno o dos municipios no vacíos, una o más unidades de Unidades sin
// repetir, algún precepto y sus periodos bien formados: con su primer día y su
// último, MM-DD de un día que tienen todos los años; su efecto; un precepto que
// es de los del cómputo o de los de cada uno de sus tipos; un aviso completo, de
// una línea, con un código de la gramática y una etiqueta sin dos puntos, solo
// y siempre en un periodo de aviso, y cuyo código y etiqueta no son los de otro
// aviso de periodo ni de un aviso del motor; un periodo que no corre, solo un
// mes entero de 30 o 31 días y en un cómputo de solo meses; y ningún día en dos
// periodos del cómputo. Cada precepto lleva su norma, su bloque, su apartado, lo
// que fija y el día AAAA-MM-DD desde el que rige como lo cuenta el fichero. Un
// cómputo puede no tener tipos.
//
// Los tipos se devuelven en el orden del fichero, cada uno con su cómputo y los
// preceptos de este. Con cualquier defecto no hay tipos: el error los dice
// todos, en el orden del fichero y cada uno con su tipo o su cómputo, y no lleva
// clase de usuario; el kernel lo da como inesperado.
func LeerTipos(contenido []byte) ([]Tipo, error) {
	var fichero ficheroDeTipos
	if err := decodificar(rutaDeLosTipos, contenido, &fichero); err != nil {
		return nil, err
	}

	for indice := range fichero.Computos {
		fichero.Computos[indice].posicion = indice
	}

	c := comprobacionDeTipos{computos: fichero.Computos, posiciones: posicionesDeLosComputos(fichero.Computos)}

	c.tipos(fichero.Tipos)
	c.comprobarComputos(fichero.Tipos)

	if err := errors.Join(c.defectos...); err != nil {
		return nil, err
	}

	for indice := range fichero.Tipos {
		tipo := &fichero.Tipos[indice]
		tipo.computo = *c.computoDe(tipo.Computo)
		tipo.PreceptosDelComputo = tipo.computo.Preceptos
	}

	return fichero.Tipos, nil
}

// posicionesDeLosComputos da de cada código de cómputo la posición del primero
// que lo lleva.
func posicionesDeLosComputos(computos []computoDescrito) map[Computo]int {
	posiciones := make(map[Computo]int, len(computos))

	for indice, computo := range computos {
		if _, visto := posiciones[computo.Codigo]; !visto && computo.Codigo != "" {
			posiciones[computo.Codigo] = indice
		}
	}

	return posiciones
}

// comprobacionDeTipos acumula los defectos de data/plazos.yaml en el orden del
// fichero: los de los tipos y, después, los de los cómputos.
type comprobacionDeTipos struct {
	defectos []error
	// computos son los del fichero, cada uno con su posición.
	computos []computoDescrito
	// posiciones da de cada código de cómputo la del primero que lo lleva.
	posiciones map[Computo]int
}

// defecto anota un defecto de data/plazos.yaml.
func (c *comprobacionDeTipos) defecto(formato string, argumentos ...any) {
	c.defectos = append(c.defectos, &defectoDeDatos{fichero: rutaDeLosTipos, motivo: fmt.Sprintf(formato, argumentos...)})
}

// computoDe es el cómputo de un código, o nil si no está en computos.
func (c *comprobacionDeTipos) computoDe(codigo Computo) *computoDescrito {
	posicion, esta := c.posiciones[codigo]
	if !esta {
		return nil
	}

	return &c.computos[posicion]
}

// tipos comprueba cada tipo y que su código siga la gramática y no se repita.
func (c *comprobacionDeTipos) tipos(tipos []Tipo) {
	vistos := make(map[string]bool, len(tipos))

	for indice, tipo := range tipos {
		switch {
		case tipo.Codigo == "":
			c.defecto("el tipo %d no tiene código", indice+1)
		case !gramaticaDeUnCodigo.MatchString(tipo.Codigo):
			c.defecto("el código %q del tipo %d %s", tipo.Codigo, indice+1, noEsUnCodigo)
		case vistos[tipo.Codigo]:
			c.defecto("el tipo %q está repetido", tipo.Codigo)
		}

		vistos[tipo.Codigo] = true

		c.tipo(tipo)
	}
}

// tipo comprueba el nombre, el cómputo, la duración, lo que da, las unidades y
// los preceptos de un tipo.
func (c *comprobacionDeTipos) tipo(tipo Tipo) {
	if tipo.Nombre == "" {
		c.defecto("el tipo %q no tiene nombre", tipo.Codigo)
	}

	computo := c.computoDe(tipo.Computo)
	if computo == nil {
		c.defecto("el tipo %q nombra el cómputo %q, que no está en computos", tipo.Codigo, tipo.Computo)
	}

	c.duracion(tipo, computo)
	c.da(tipo)
	c.unidades(tipo, computo)

	if len(tipo.Preceptos) == 0 && len(tipo.Unidades) == 0 {
		c.defecto("el tipo %q no nombra ningún precepto", tipo.Codigo)
	}

	for indice, precepto := range tipo.Preceptos {
		c.precepto(fmt.Sprintf("el precepto %d del tipo %q", indice+1, tipo.Codigo), precepto)
	}
}

// duracion comprueba que la duración de un tipo, si la lleva, es de 1 o más,
// en una unidad que su cómputo admite, y que la fija uno de sus preceptos.
func (c *comprobacionDeTipos) duracion(tipo Tipo, computo *computoDescrito) {
	duracion := tipo.Duracion
	if duracion == nil {
		return
	}

	if duracion.Cantidad < 1 {
		c.defecto("la duración del tipo %q no es un entero de 1 o más", tipo.Codigo)
	}

	switch {
	case !slices.Contains(Unidades(), duracion.Unidad):
		c.defecto("la unidad %q del tipo %q no es %s", duracion.Unidad, tipo.Codigo, enumeracion(Unidades()))
	case computo != nil && !slices.Contains(computo.Unidades, duracion.Unidad):
		c.defecto("la unidad %q de la duración del tipo %q no la admite su cómputo, %q", duracion.Unidad, tipo.Codigo,
			computo.Codigo)
	}

	fijada := slices.ContainsFunc(tipo.Preceptos, func(p Precepto) bool { return p.Fija == FijaDuracion })
	if !fijada {
		c.defecto("ningún precepto del tipo %q fija su duración", tipo.Codigo)
	}
}

// da comprueba que el tipo dice qué fecha da, con uno de sus dos valores.
func (c *comprobacionDeTipos) da(tipo Tipo) {
	switch tipo.Da {
	case daVencimiento, daDiaSiguiente:
	case "":
		c.defecto("el tipo %q no lleva da", tipo.Codigo)
	default:
		c.defecto("el tipo %q da %q, que no es %s ni %s", tipo.Codigo, tipo.Da, daVencimiento, daDiaSiguiente)
	}
}

// unidades comprueba que un tipo que lleva unidades no lleva también duración y
// lleva exactamente las de su cómputo, cada una con su precepto.
func (c *comprobacionDeTipos) unidades(tipo Tipo, computo *computoDescrito) {
	if tipo.Unidades == nil {
		return
	}

	if tipo.Duracion != nil {
		c.defecto("el tipo %q lleva unidades y duración", tipo.Codigo)
	}

	if computo == nil {
		for _, unidad := range slices.Sorted(maps.Keys(tipo.Unidades)) {
			c.precepto(preceptoDeLaUnidad(tipo, unidad), tipo.Unidades[unidad])
		}

		return
	}

	c.unidadesDelComputo(tipo, *computo)
}

// unidadesDelComputo comprueba que un tipo lleva cada unidad de su cómputo, con
// su precepto, y ninguna más.
func (c *comprobacionDeTipos) unidadesDelComputo(tipo Tipo, computo computoDescrito) {
	vistas := make(map[Unidad]bool, len(computo.Unidades))

	for _, unidad := range computo.Unidades {
		if vistas[unidad] {
			continue
		}

		vistas[unidad] = true

		precepto, esta := tipo.Unidades[unidad]
		if !esta {
			c.defecto("el tipo %q no lleva la unidad %q de su cómputo, %q", tipo.Codigo, unidad, computo.Codigo)

			continue
		}

		c.precepto(preceptoDeLaUnidad(tipo, unidad), precepto)
	}

	for _, unidad := range slices.Sorted(maps.Keys(tipo.Unidades)) {
		if !vistas[unidad] {
			c.defecto("el tipo %q lleva la unidad %q, que su cómputo, %q, no admite", tipo.Codigo, unidad, computo.Codigo)
		}
	}
}

// preceptoDeLaUnidad es como nombran sus defectos el precepto de una unidad de
// un tipo.
func preceptoDeLaUnidad(tipo Tipo, unidad Unidad) string {
	return fmt.Sprintf("el precepto de la unidad %q del tipo %q", unidad, tipo.Codigo)
}

// comprobarComputos comprueba cada cómputo, en el orden del fichero, y que su
// código siga la gramática y no se repita.
func (c *comprobacionDeTipos) comprobarComputos(tipos []Tipo) {
	vistos := make(map[Computo]bool, len(c.computos))
	avisos := avisosDelMotor()

	for indice, computo := range c.computos {
		switch {
		case computo.Codigo == "":
			c.defecto("el cómputo %d no tiene código", indice+1)
		case !gramaticaDeUnCodigo.MatchString(string(computo.Codigo)):
			c.defecto("el código %q del cómputo %d %s", computo.Codigo, indice+1, noEsUnCodigo)
		case vistos[computo.Codigo]:
			c.defecto("el cómputo %q está repetido en computos", computo.Codigo)
		}

		vistos[computo.Codigo] = true

		c.computo(computo, tiposDelComputo(tipos, computo.Codigo), avisos)
	}
}

// tiposDelComputo son los tipos que cuentan con un cómputo.
func tiposDelComputo(tipos []Tipo, codigo Computo) []Tipo {
	var suyos []Tipo

	for _, tipo := range tipos {
		if tipo.Computo == codigo {
			suyos = append(suyos, tipo)
		}
	}

	return suyos
}

// computo comprueba los municipios, las unidades, los periodos y los preceptos
// de un cómputo.
func (c *comprobacionDeTipos) computo(computo computoDescrito, suyos []Tipo, avisos *avisosVistos) {
	if cuantos := len(computo.Municipios); cuantos < 1 || cuantos > 2 {
		c.defecto("el cómputo %q no lleva uno ni dos municipios", computo.Codigo)
	}

	if slices.Contains(computo.Municipios, "") {
		c.defecto("el cómputo %q lleva un municipio vacío", computo.Codigo)
	}

	c.unidadesDeUnComputo(computo)
	c.periodos(computo, suyos, avisos)

	if len(computo.Preceptos) == 0 {
		c.defecto("el cómputo %q no nombra ningún precepto", computo.Codigo)
	}

	for indice, precepto := range computo.Preceptos {
		c.precepto(fmt.Sprintf("el precepto %d del cómputo %q", indice+1, computo.Codigo), precepto)
	}
}

// unidadesDeUnComputo comprueba que un cómputo lleva una o más unidades de
// Unidades, sin repetir.
func (c *comprobacionDeTipos) unidadesDeUnComputo(computo computoDescrito) {
	if len(computo.Unidades) == 0 {
		c.defecto("el cómputo %q no lleva unidades", computo.Codigo)

		return
	}

	vistas := make(map[Unidad]bool, len(computo.Unidades))

	for _, unidad := range computo.Unidades {
		switch {
		case !slices.Contains(Unidades(), unidad):
			c.defecto("el cómputo %q lleva la unidad %q, que no es %s", computo.Codigo, unidad, enumeracion(Unidades()))
		case vistas[unidad]:
			c.defecto("el cómputo %q repite la unidad %q", computo.Codigo, unidad)
		}

		vistas[unidad] = true
	}
}

// tramoDelAnio son los días de un periodo por su número en un año que no es
// bisiesto, de 1 a 365; si hasta es menor que desde, el tramo cruza el año.
type tramoDelAnio struct {
	// numero es el del periodo en su cómputo, desde 1.
	numero       int
	desde, hasta int
}

// contiene dice si un día del año es del tramo.
func (t tramoDelAnio) contiene(dia int) bool {
	if t.desde <= t.hasta {
		return dia >= t.desde && dia <= t.hasta
	}

	return dia >= t.desde || dia <= t.hasta
}

// periodos comprueba cada periodo de un cómputo y que dos periodos no
// compartan ningún día.
func (c *comprobacionDeTipos) periodos(computo computoDescrito, suyos []Tipo, avisos *avisosVistos) {
	tramos := make([]tramoDelAnio, 0, len(computo.Periodos))

	for indice, periodo := range computo.Periodos {
		numero := indice + 1
		quien := fmt.Sprintf("el periodo %d del cómputo %q", numero, computo.Codigo)

		desde, conDesde := c.diaDelPeriodo(quien, "desde", periodo.Desde)
		hasta, conHasta := c.diaDelPeriodo(quien, "hasta", periodo.Hasta)

		c.efecto(quien, periodo.Efecto)
		c.preceptoDelPeriodo(numero, computo, periodo.Precepto, suyos)
		c.avisoDelPeriodo(numero, computo.Codigo, periodo, avisos)

		if !conDesde || !conHasta {
			continue
		}

		if periodo.Efecto == efectoNoCorre {
			c.noCorre(quien, computo, desde, hasta)
		}

		tramos = append(tramos, tramoDelAnio{numero: numero, desde: desde.YearDay(), hasta: hasta.YearDay()})
	}

	c.solapes(computo.Codigo, tramos)
}

// diaDelPeriodo lee el primer día o el último de un periodo, que es MM-DD de un
// día que tienen todos los años.
func (c *comprobacionDeTipos) diaDelPeriodo(quien, nombre, texto string) (time.Time, bool) {
	if texto == "" {
		c.defecto("%s no tiene %s", quien, nombre)

		return time.Time{}, false
	}

	dia, err := time.Parse(time.DateOnly, fmt.Sprintf("%04d-%s", anioDeLosPeriodos, texto))
	if err != nil {
		c.defecto("%s tiene un %s, %q, que no es MM-DD de un día de todos los años", quien, nombre, texto)

		return time.Time{}, false
	}

	return dia, true
}

// efecto comprueba que el periodo lleva uno de los tres efectos.
func (c *comprobacionDeTipos) efecto(quien, efecto string) {
	switch efecto {
	case efectoInhabil, efectoNoCorre, efectoAviso:
	case "":
		c.defecto("%s no tiene efecto", quien)
	default:
		c.defecto("%s tiene un efecto, %q, que no es %s, %s ni %s", quien, efecto, efectoInhabil, efectoNoCorre,
			efectoAviso)
	}
}

// preceptoDelPeriodo comprueba que el periodo nombra, por su norma y su
// apartado, un precepto del cómputo o, si el cómputo tiene tipos, uno de cada
// uno de ellos.
func (c *comprobacionDeTipos) preceptoDelPeriodo(numero int, computo computoDescrito,
	referencia *referenciaDeUnPrecepto, suyos []Tipo,
) {
	if referencia == nil {
		c.defecto("el periodo %d del cómputo %q no tiene precepto", numero, computo.Codigo)

		return
	}

	completo := true

	for _, dato := range []struct{ nombre, valor string }{
		{"norma", referencia.Norma},
		{"apartado", referencia.Apartado},
	} {
		if dato.valor == "" {
			c.defecto("el precepto del periodo %d del cómputo %q no tiene %s", numero, computo.Codigo, dato.nombre)

			completo = false
		}
	}

	if !completo || slices.ContainsFunc(computo.Preceptos, referencia.nombra) {
		return
	}

	deCadaTipo := len(suyos) > 0 && !slices.ContainsFunc(suyos, func(tipo Tipo) bool {
		return !slices.ContainsFunc(tipo.Preceptos, referencia.nombra)
	})
	if !deCadaTipo {
		c.defecto("el precepto del periodo %d del cómputo %q, %s %s, no está entre los del cómputo ni entre los de "+
			"cada uno de sus tipos", numero, computo.Codigo, referencia.Norma, referencia.Apartado)
	}
}

// noCorre comprueba que un periodo que no corre es un mes entero de 30 o 31
// días, del día 1 al último, en un cómputo que solo admite meses.
func (c *comprobacionDeTipos) noCorre(quien string, computo computoDescrito, desde, hasta time.Time) {
	elUltimoDelMes := hasta.AddDate(0, 0, 1).Day() == 1
	if desde.Day() != 1 || hasta.Month() != desde.Month() || !elUltimoDelMes || hasta.Day() < 30 {
		c.defecto("%s no corre y no es un mes entero de 30 o 31 días", quien)
	}

	if slices.ContainsFunc(computo.Unidades, func(unidad Unidad) bool { return unidad != UnidadMeses }) {
		c.defecto("%s no corre y el cómputo admite otras unidades que %s", quien, UnidadMeses)
	}
}

// solapes comprueba que dos periodos de un cómputo no comparten ningún día, y
// nombra el primero del año que comparten.
func (c *comprobacionDeTipos) solapes(computo Computo, tramos []tramoDelAnio) {
	for indice, uno := range tramos {
		for _, otro := range tramos[indice+1:] {
			if dia, comparten := primerDiaEnComun(uno, otro); comparten {
				c.defecto("los periodos %d y %d del cómputo %q comparten el %s", uno.numero, otro.numero, computo, dia)
			}
		}
	}
}

// primerDiaEnComun es el primer día del año, MM-DD, que es de los dos tramos.
func primerDiaEnComun(uno, otro tramoDelAnio) (string, bool) {
	primero := time.Date(anioDeLosPeriodos, time.January, 1, 0, 0, 0, 0, time.UTC)

	for dia := primero; dia.Year() == anioDeLosPeriodos; dia = dia.AddDate(0, 0, 1) {
		if uno.contiene(dia.YearDay()) && otro.contiene(dia.YearDay()) {
			return dia.Format("01-02"), true
		}
	}

	return "", false
}

// avisosVistos son los códigos y las etiquetas de los avisos del motor y de los
// avisos de periodo ya leídos, en el orden del fichero.
type avisosVistos struct {
	delMotor  map[string]bool
	codigos   map[string]bool
	etiquetas map[string]bool
}

// avisosDelMotor son los códigos y las etiquetas de los tres avisos que da el
// motor, que no son de ningún cómputo (data-model §5).
func avisosDelMotor() *avisosVistos {
	vistos := &avisosVistos{delMotor: map[string]bool{}, codigos: map[string]bool{}, etiquetas: map[string]bool{}}

	for _, codigo := range CodigosDeAviso(nil) {
		etiqueta, _ := etiquetaDelMotor(codigo)

		vistos.delMotor[codigo] = true
		vistos.etiquetas[etiqueta] = true
	}

	return vistos
}

// avisoDelPeriodo comprueba que un periodo lleva su aviso si y solo si es de
// aviso, y que ese aviso está completo y no repite otro.
func (c *comprobacionDeTipos) avisoDelPeriodo(numero int, computo Computo, periodo periodo, avisos *avisosVistos) {
	quien := fmt.Sprintf("el periodo %d del cómputo %q", numero, computo)

	switch {
	case periodo.Efecto == efectoAviso && periodo.Aviso == nil:
		c.defecto("%s es de aviso y no lleva su aviso", quien)

		return
	case periodo.Efecto == efectoAviso:
	case periodo.Aviso != nil && (periodo.Efecto == efectoInhabil || periodo.Efecto == efectoNoCorre):
		c.defecto("%s no es de aviso y lleva un aviso", quien)

		return
	default:
		return
	}

	c.formaDelAviso(numero, computo, *periodo.Aviso)
	c.unicidadDelAviso(numero, computo, *periodo.Aviso, avisos)
}

// formaDelAviso comprueba que el aviso de un periodo lleva un código de la
// gramática, y una etiqueta sin dos puntos y un texto, los dos no vacíos y de
// una línea.
func (c *comprobacionDeTipos) formaDelAviso(numero int, computo Computo, aviso avisoDeUnPeriodo) {
	delAviso := fmt.Sprintf("el aviso del periodo %d del cómputo %q", numero, computo)

	switch {
	case aviso.Codigo == "":
		c.defecto("%s no tiene código", delAviso)
	case !gramaticaDeUnCodigo.MatchString(aviso.Codigo):
		c.defecto("el código %q del aviso del periodo %d del cómputo %q %s", aviso.Codigo, numero, computo, noEsUnCodigo)
	}

	c.lineaDelAviso(delAviso, "etiqueta", aviso.Etiqueta)

	if strings.Contains(aviso.Etiqueta, ":") {
		c.defecto("%s tiene una etiqueta con dos puntos", delAviso)
	}

	c.lineaDelAviso(delAviso, "texto", aviso.Texto)
}

// lineaDelAviso comprueba que la etiqueta o el texto de un aviso no está vacío
// y es de una línea.
func (c *comprobacionDeTipos) lineaDelAviso(delAviso, nombre, valor string) {
	switch {
	case valor == "":
		c.defecto("%s no tiene %s", delAviso, nombre)
	case strings.ContainsAny(valor, "\n\r"):
		c.defecto("%s tiene un salto de línea en su %s", delAviso, nombre)
	}
}

// unicidadDelAviso comprueba que el código y la etiqueta del aviso de un
// periodo no son los de un aviso del motor ni los de otro aviso de periodo, y
// los anota.
func (c *comprobacionDeTipos) unicidadDelAviso(numero int, computo Computo, aviso avisoDeUnPeriodo,
	avisos *avisosVistos,
) {
	quien := fmt.Sprintf("el aviso %q del periodo %d del cómputo %q", aviso.Codigo, numero, computo)

	switch {
	case aviso.Codigo == "":
	case avisos.delMotor[aviso.Codigo]:
		c.defecto("%s es un aviso del motor", quien)
	case avisos.codigos[aviso.Codigo]:
		c.defecto("%s repite el código de otro aviso", quien)
	default:
		avisos.codigos[aviso.Codigo] = true
	}

	switch {
	case aviso.Etiqueta == "":
	case avisos.etiquetas[aviso.Etiqueta]:
		c.defecto("%s repite la etiqueta de otro aviso", quien)
	default:
		avisos.etiquetas[aviso.Etiqueta] = true
	}
}

// precepto comprueba que un precepto lleva sus cinco datos, y el día desde el
// que rige como una fecha AAAA-MM-DD.
func (c *comprobacionDeTipos) precepto(quien string, precepto Precepto) {
	for _, dato := range []struct{ nombre, valor string }{
		{"norma", precepto.Norma},
		{"bloque", precepto.Bloque},
		{"apartado", precepto.Apartado},
		{"fija", precepto.Fija},
		{"vigente_desde", precepto.VigenteDesde},
	} {
		if dato.valor == "" {
			c.defecto("%s no tiene %s", quien, dato.nombre)
		}
	}

	if desde := precepto.VigenteDesde; desde != "" && !esUnDia(desde) {
		c.defecto("%s tiene un vigente_desde, %q, que no es una fecha AAAA-MM-DD", quien, desde)
	}
}

// esUnDia dice si un texto es un día del calendario escrito AAAA-MM-DD, con sus
// ceros: el mismo que leerInicio admite como fecha de inicio, de modo que los
// dos se comparan como texto.
func esUnDia(texto string) bool {
	_, err := leerInicio(texto)

	return err == nil
}

// defectoDeDatos es un defecto de un fichero de data/ empotrado: uno que no se
// lee o un dato que no vale. No declara ninguna clase a propósito —el kernel lo
// trata como inesperado, código 1—: un fichero empotrado que no se lee es un
// defecto del binario y no algo que quien pregunta pueda corregir.
type defectoDeDatos struct {
	// fichero es la ruta del fichero, relativa a la raíz del repositorio.
	fichero string
	// motivo dice qué falla y nombra el dato.
	motivo string
}

// Error nombra el fichero y el defecto, como «data/plazos.yaml: el tipo
// "alzada" no lleva da».
func (d *defectoDeDatos) Error() string {
	return d.fichero + ": " + d.motivo
}

// decodificar lee un fichero YAML en su tipo. Una clave que el tipo no declara
// es un defecto, no algo que se descarta en silencio, y la clave repetida la
// rechaza el propio lector.
func decodificar(ruta string, contenido []byte, destino any) error {
	if len(bytes.TrimSpace(contenido)) == 0 {
		return &defectoDeDatos{fichero: ruta, motivo: "está vacío"}
	}

	lector := yaml.NewDecoder(bytes.NewReader(contenido))
	lector.KnownFields(true)

	if err := lector.Decode(destino); err != nil {
		return &defectoDeDatos{fichero: ruta, motivo: "no se puede leer: " + err.Error()}
	}

	return nil
}

// enumeracion escribe un vocabulario como «a, b ni c».
func enumeracion[T ~string](valores []T) string {
	var texto string

	for indice, valor := range valores {
		switch {
		case indice == 0:
		case indice == len(valores)-1:
			texto += " ni "
		default:
			texto += ", "
		}

		texto += string(valor)
	}

	return texto
}
