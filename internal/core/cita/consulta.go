package cita

import "strings"

// Las direcciones que abre la persona con su navegador. El paquete las
// compone y no las pide: ni consulta el buscador ni comprueba que responden
// (H23, FR-002).
const (
	// direccionDelBuscador es la del buscador del CENDOJ, donde se rellenan
	// las casillas (FR-010).
	direccionDelBuscador = "https://www.poderjudicial.es/search/indexAN.jsp"
	// prefijoDeLaBusqueda y sufijoDeLaBusqueda rodean el texto codificado en
	// la dirección que abre el buscador con la búsqueda ya hecha (FR-013).
	prefijoDeLaBusqueda = "https://www.poderjudicial.es/search/sentencias/"
	sufijoDeLaBusqueda  = "/1/AN"
)

// Las casillas del buscador, con el nombre que ve la persona (FR-011). La
// fecha son dos entradas de una misma casilla, con su campo (research.md D5).
const (
	casillaECLI       = "ECLI"
	casillaROJ        = "Nº ROJ"
	casillaResolucion = "Nº Resolución"
	casillaFecha      = "Fecha resolución"
	campoDesde        = "Desde"
	campoHasta        = "Hasta"
)

// La única cobertura que el paquete declara: la del ECLI del Tribunal
// Constitucional, que el CENDOJ no tiene (FR-014). No hay un valor para lo
// cubierto, porque de ninguna otra consulta se sabe si el CENDOJ la tiene
// (research.md D4).
const (
	// noCubierto es el único valor de `cendoj`.
	noCubierto = "no-cubierto"
	// organoDelConstitucional es el código de órgano de sus ECLI.
	organoDelConstitucional = "TC"
	// motivoDelConstitucional es lo que se le dice a la persona.
	motivoDelConstitucional = "Las resoluciones del Tribunal Constitucional no están en el CENDOJ: " +
		"su buscador no las tiene."
)

// cifrasHexadecimales son las del octeto codificado, en mayúsculas (FR-013).
const cifrasHexadecimales = "0123456789ABCDEF"

// Consulta es lo que una persona tiene que hacer para encontrar una sentencia
// en el buscador del CENDOJ: el data de cita preparar, con las claves de
// contracts/applet-cita.md §3. Lleva una referencia o un texto, nunca los dos,
// y lo que no aplica no está: ni cadena vacía ni null (research.md D4). No
// dice si la sentencia existe: nadie lo ha comprobado.
type Consulta struct {
	// Referencia es la reconocida, si la consulta es la de una referencia.
	Referencia *Referencia `json:"referencia,omitempty"`
	// Texto es el de la búsqueda por materia, tal como se dio.
	Texto string `json:"texto,omitempty"`
	// Equivalente es el otro identificador de la sentencia, solo si se deduce
	// de la referencia sin consultar nada (FR-012).
	Equivalente *Equivalente `json:"equivalente,omitempty"`
	// Cobertura está solo cuando la referencia queda fuera de lo que el CENDOJ
	// tiene (FR-014).
	Cobertura *Cobertura `json:"cobertura,omitempty"`
	// Direccion es la que hay que abrir: la del buscador con una referencia y
	// la de la búsqueda ya hecha con un texto. Fuera de cobertura no hay.
	Direccion string `json:"direccion,omitempty" jsonschema:"format=uri"`
	// Casillas son las que hay que rellenar, en su orden: siempre una lista,
	// vacía con un texto y fuera de cobertura.
	Casillas []Casilla `json:"casillas"`
}

// Equivalente es el identificador que se deduce de una referencia: el ROJ de
// un ECLI o el ECLI de un ROJ. Es un dato, no otra casilla.
type Equivalente struct {
	// Forma es roj o ecli: la otra, no la de la referencia.
	Forma string `json:"forma" jsonschema:"enum=ecli,enum=roj"`
	// Valor es el identificador deducido.
	Valor string `json:"valor" jsonschema:"minLength=1"`
}

// Cobertura declara que el CENDOJ no tiene lo que se busca.
type Cobertura struct {
	// CENDOJ es no-cubierto, su único valor.
	CENDOJ string `json:"cendoj" jsonschema:"enum=no-cubierto"`
	// Motivo dice por qué, para la persona.
	Motivo string `json:"motivo" jsonschema:"minLength=1"`
}

// Casilla es un campo del buscador con lo que hay que escribir en él.
type Casilla struct {
	// Nombre es el que la persona ve en el buscador.
	Nombre string `json:"nombre" jsonschema:"minLength=1"`
	// Campo es la parte de una casilla que tiene dos: Desde o Hasta.
	Campo string `json:"campo,omitempty"`
	// Valor es lo que se escribe; una fecha, dd/mm/aaaa.
	Valor string `json:"valor" jsonschema:"minLength=1"`
}

// Preparar da la consulta de una referencia: la dirección del buscador, las
// casillas de la forma dada y ninguna más, y el equivalente si se deduce. Con
// un ECLI del Tribunal Constitucional declara que el CENDOJ no lo cubre, sin
// dirección, sin casillas y sin equivalente; solo de un ECLI, porque de una
// referencia dada de otra forma no se sabe de qué órgano es (H23, FR-010,
// FR-011, FR-012, FR-014).
func Preparar(referencia Referencia) Consulta {
	consulta := Consulta{Referencia: &referencia, Casillas: []Casilla{}}

	if referencia.Forma == formaECLI && referencia.ecli.Organo() == organoDelConstitucional {
		consulta.Cobertura = &Cobertura{CENDOJ: noCubierto, Motivo: motivoDelConstitucional}

		return consulta
	}

	consulta.Direccion = direccionDelBuscador
	consulta.Casillas = casillasDe(referencia)
	consulta.Equivalente = equivalenteDe(referencia)

	return consulta
}

// PrepararTexto da la consulta de una búsqueda por materia: el texto tal como
// se dio y la dirección que abre el buscador con esa búsqueda ya hecha, sin
// casillas y sin añadir al texto operadores ni filtros. Un texto vacío o solo
// de blancos es un error de clase «argumentos» (H23, FR-013, FR-015).
func PrepararTexto(texto string) (Consulta, error) {
	if strings.TrimSpace(texto) == "" {
		return Consulta{}, argumentosInvalidos(
			"el texto de la búsqueda %q no es válido: no puede ir vacío ni ser solo de blancos", texto)
	}

	return Consulta{
		Texto:     texto,
		Direccion: prefijoDeLaBusqueda + codificarSegmento(texto) + sufijoDeLaBusqueda,
		Casillas:  []Casilla{},
	}, nil
}

// casillasDe son las de la forma dada, en el orden de FR-011. Las formas son
// tres y solo las construye NuevaReferencia: la que no es ni un ECLI ni un ROJ
// es el número con su fecha, que va escrita como la escribe quien rellena la
// casilla.
func casillasDe(referencia Referencia) []Casilla {
	switch referencia.Forma {
	case formaECLI:
		return []Casilla{{Nombre: casillaECLI, Valor: referencia.Valor}}
	case formaROJ:
		return []Casilla{{Nombre: casillaROJ, Valor: referencia.Valor}}
	}

	dia := referencia.dia.Format(formaDeLaFechaDelCENDOJ)

	return []Casilla{
		{Nombre: casillaResolucion, Valor: referencia.Valor},
		{Nombre: casillaFecha, Campo: campoDesde, Valor: dia},
		{Nombre: casillaFecha, Campo: campoHasta, Valor: dia},
	}
}

// equivalenteDe es el otro identificador de la referencia, o nil si no se
// deduce: la regla, con su única pareja, es de internal/core/ids. De un número
// de resolución no se deduce ninguno (FR-012).
func equivalenteDe(referencia Referencia) *Equivalente {
	if roj, seDeduce := referencia.ecli.ROJ(); seDeduce {
		return &Equivalente{Forma: formaROJ, Valor: roj.String()}
	}

	if ecli, seDeduce := referencia.roj.ECLI(); seDeduce {
		return &Equivalente{Forma: formaECLI, Valor: ecli.String()}
	}

	return nil
}

// codificarSegmento escribe el texto como un segmento de ruta de una
// dirección: octeto a octeto, cada uno que no es una letra ASCII, una cifra,
// «-», «.», «_» o «~» va como % y sus dos cifras hexadecimales en mayúsculas,
// de modo que un espacio es %20 y una letra con tilde, sus dos octetos de
// UTF-8 (FR-013).
func codificarSegmento(texto string) string {
	var codificado strings.Builder

	for indice := range len(texto) {
		octeto := texto[indice]

		if vaTalCual(octeto) {
			codificado.WriteByte(octeto)

			continue
		}

		codificado.WriteByte('%')
		codificado.WriteByte(cifrasHexadecimales[octeto>>4])
		codificado.WriteByte(cifrasHexadecimales[octeto&0x0f])
	}

	return codificado.String()
}

// vaTalCual dice si el octeto se escribe sin codificar en un segmento de
// ruta: los no reservados de una dirección.
func vaTalCual(octeto byte) bool {
	switch {
	case 'a' <= octeto && octeto <= 'z', 'A' <= octeto && octeto <= 'Z', '0' <= octeto && octeto <= '9':
		return true
	case octeto == '-', octeto == '.', octeto == '_', octeto == '~':
		return true
	}

	return false
}
