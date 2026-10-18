package plazos

import "time"

// Los tipos del plazo calculado, el data del verbo calcular, con las claves
// de contracts/applet-plazos.md §4 en español y ninguna más: ni la fecha de
// inicio, ni la duración, ni el cómputo, ni la comunidad de un festivo (FR-001;
// research D5, D29). Solo llevan omitempty las claves que §4 da como
// opcionales —el municipio de un festivo o de un tramo local, y la disposición
// o la falta de un tramo—, y las listas nunca van a null: los festivos y los
// avisos pueden ir vacíos y van como [].
//
// Las etiquetas jsonschema son la descripción formal de los valores, de la que
// --describe genera el esquema del verbo. Una etiqueta no puede nombrar una
// constante, así que los enumerados repiten los vocabularios del dominio
// —CodigosDeTipo, Ambitos, CodigosDeAviso, AmbitosDeCobertura y
// FaltasDeCobertura— y TestSalidaDePlazos exige que digan lo mismo.

// Plazo es el plazo calculado: la fecha que se da y lo que la sostiene.
type Plazo struct {
	// Fecha es la que se da, AAAA-MM-DD.
	Fecha string `json:"fecha" jsonschema:"pattern=^[0-9]{4}-[0-9]{2}-[0-9]{2}$"`
	// Tipo es el pedido.
	Tipo string `json:"tipo" jsonschema:"enum=alzada,enum=reposicion,enum=contencioso,enum=acceso-resolver,enum=acceso-reclamar,enum=silencio,enum=generico"`
	// Preceptos son los que fijan el plazo; en el contencioso, también el art.
	// 128.2 LJCA.
	Preceptos []PreceptoDelPlazo `json:"preceptos" jsonschema:"minItems=1"`
	// Municipios son los que cuentan, una vez cada uno, en el orden dado.
	Municipios []MunicipioDelPlazo `json:"municipios" jsonschema:"minItems=1"`
	// Festivos son los que el cómputo saltó, en orden de fecha.
	Festivos []FestivoSaltado `json:"festivos"`
	// Avisos son los de contracts/applet-plazos.md §5, en el orden de sus
	// códigos.
	Avisos []Aviso `json:"avisos"`
	// Cobertura dice, por ámbito y por tramos de años del tramo del plazo,
	// hasta qué disposición llegan los datos o por qué faltan.
	Cobertura []TramoDeCobertura `json:"cobertura" jsonschema:"minItems=1"`
}

// PreceptoDelPlazo es un apartado de una norma consolidada del BOE que funda
// el plazo.
type PreceptoDelPlazo struct {
	// Norma es el identificador BOE-A-… de la norma.
	Norma string `json:"norma" jsonschema:"pattern=^BOE-A-[0-9]{4}-[0-9]+$"`
	// Bloque es el id del bloque en el índice consolidado de la norma.
	Bloque string `json:"bloque" jsonschema:"minLength=1"`
	// Apartado es el que se cita, como «124.1».
	Apartado string `json:"apartado" jsonschema:"minLength=1"`
}

// MunicipioDelPlazo es un municipio que cuenta, como lo da territorio
// resolver.
type MunicipioDelPlazo struct {
	// CodigoINE son sus cinco cifras.
	CodigoINE string `json:"codigo_ine" jsonschema:"pattern=^[0-9]{5}$"`
	// Nombre es el oficial.
	Nombre string `json:"nombre" jsonschema:"minLength=1"`
}

// FestivoSaltado es un festivo que el cómputo saltó: nunca un sábado ni un
// domingo.
type FestivoSaltado struct {
	// Fecha es la del festivo, AAAA-MM-DD.
	Fecha string `json:"fecha" jsonschema:"pattern=^[0-9]{4}-[0-9]{2}-[0-9]{2}$"`
	// Ambito es nacional, autonomico o local.
	Ambito string `json:"ambito" jsonschema:"enum=nacional,enum=autonomico,enum=local"`
	// Municipio es el código INE del municipio, solo en uno local.
	Municipio string `json:"municipio,omitempty" jsonschema:"pattern=^[0-9]{5}$"`
	// Source es la disposición que lo fija.
	Source string `json:"source" jsonschema:"minLength=1"`
}

// Aviso es un aviso del plazo: su código y su texto, una línea con su forma
// fija.
type Aviso struct {
	// Codigo es uno de CodigosDeAviso.
	Codigo string `json:"codigo" jsonschema:"enum=faltan-festivos-nacionales-y-autonomicos,enum=faltan-festivos-locales,enum=dias-del-art-183-lopj,enum=festivo-de-una-parte"`
	// Texto empieza por «⚠ <ETIQUETA>: ».
	Texto string `json:"texto" jsonschema:"minLength=1"`
}

// TramoDeCobertura es un tramo de años de la cobertura de un ámbito: un año
// cargado, con la última disposición aplicada o la falta del municipio, o un
// hueco de años sin datos, con su falta.
type TramoDeCobertura struct {
	// Ambito es nacional-y-autonomico o local.
	Ambito string `json:"ambito" jsonschema:"enum=nacional-y-autonomico,enum=local"`
	// Municipio es el código INE, solo en uno local.
	Municipio string `json:"municipio,omitempty" jsonschema:"pattern=^[0-9]{5}$"`
	// Desde es el primer año del tramo.
	Desde int `json:"desde"`
	// Hasta es el último año del tramo.
	Hasta int `json:"hasta"`
	// Disposicion es la última aplicada del año, si hay festivos.
	Disposicion string `json:"disposicion,omitempty"`
	// Falta es por qué no hay festivos, si no los hay.
	Falta string `json:"falta,omitempty" jsonschema:"enum=sin-cargar,enum=sin-festivos-locales,enum=no-comunicado,enum=linea-sin-leer,enum=sin-linea"`
}

// salida es el plazo calculado con su resultado: la cobertura de los años del
// tramo, los preceptos del tipo y, en el genérico, el de su unidad, los
// municipios, los festivos saltados y los avisos.
func (c calculo) salida(unidad Unidad, resultado computo) Plazo {
	cobertura := c.calendario.Cobertura(c.desde.Year(), c.hasta.Year(), c.delCalendario)

	return Plazo{
		Fecha:      resultado.dia.Format(time.DateOnly),
		Tipo:       c.tipo.Codigo,
		Preceptos:  preceptosDe(c.tipo, unidad),
		Municipios: c.municipiosDelPlazo(),
		Festivos:   festivosDelPlazo(resultado.festivos),
		Avisos:     c.avisos(cobertura),
		Cobertura:  coberturaDelPlazo(cobertura),
	}
}

// preceptosDe son los preceptos de un tipo y, en el genérico, el que da el
// cómputo de su unidad (contracts/applet-plazos.md §1).
func preceptosDe(tipo Tipo, unidad Unidad) []PreceptoDelPlazo {
	preceptos := make([]PreceptoDelPlazo, 0, len(tipo.Preceptos)+1)
	for _, precepto := range tipo.Preceptos {
		preceptos = append(preceptos, PreceptoDelPlazo{
			Norma: precepto.Norma, Bloque: precepto.Bloque,
			Apartado: precepto.Apartado,
		})
	}

	if precepto, esta := tipo.Unidades[unidad]; esta {
		preceptos = append(preceptos, PreceptoDelPlazo{
			Norma: precepto.Norma, Bloque: precepto.Bloque,
			Apartado: precepto.Apartado,
		})
	}

	return preceptos
}

// municipiosDelPlazo son los municipios que cuentan, una vez cada uno.
func (c calculo) municipiosDelPlazo() []MunicipioDelPlazo {
	municipios := make([]MunicipioDelPlazo, 0, len(c.municipios))
	for _, municipio := range c.municipios {
		municipios = append(municipios, MunicipioDelPlazo{CodigoINE: municipio.CodigoINE, Nombre: municipio.Nombre})
	}

	return municipios
}

// festivosDelPlazo son los festivos saltados, en orden de fecha.
func festivosDelPlazo(saltados []festivoDelDia) []FestivoSaltado {
	festivos := make([]FestivoSaltado, 0, len(saltados))
	for _, saltado := range saltados {
		festivos = append(festivos, FestivoSaltado{
			Fecha:     saltado.dia.Format(time.DateOnly),
			Ambito:    string(saltado.Ambito),
			Municipio: saltado.Municipio,
			Source:    saltado.Source,
		})
	}

	return festivos
}

// coberturaDelPlazo son los tramos de la cobertura del calendario, por ámbito
// y por tramos de años: como mucho cinco por ámbito, porque cada año cargado es
// un tramo y cada hueco otro (research D5).
func coberturaDelPlazo(tramos []Tramo) []TramoDeCobertura {
	cobertura := make([]TramoDeCobertura, 0, len(tramos))
	for _, tramo := range tramos {
		cobertura = append(cobertura, TramoDeCobertura{
			Ambito:      string(tramo.Ambito),
			Municipio:   tramo.Municipio,
			Desde:       tramo.Desde,
			Hasta:       tramo.Hasta,
			Disposicion: tramo.Disposicion,
			Falta:       string(tramo.Falta),
		})
	}

	return cobertura
}
