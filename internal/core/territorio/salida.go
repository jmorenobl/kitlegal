package territorio

import "time"

// Los tipos del territorio resuelto, el data del verbo resolver, con las claves
// de data-model §2.5 en español. Ninguna lleva omitempty: todo campo se emite
// siempre, lo que no hay va como cadena vacía y la lista de boletines nunca va
// vacía, porque lleva siempre el estatal (research.md V13).
//
// Las etiquetas jsonschema son la descripción formal de los valores, de la que
// --describe genera el esquema del verbo y de la que sale el publicado
// (research.md V14, D16). Una etiqueta no puede nombrar una constante, así que
// los enumerados repiten los vocabularios de este fichero y TestTerritorioResuelto
// exige que digan lo mismo.

// Los valores de la cobertura (data-model §2.4). Ninguno significa «no
// existe»: un boletín que no está configurado puede existir igualmente, y un
// DIR3 que no está verificado también (FR-022).
const (
	configurado   = "configurado"
	noConfigurado = "no-configurado"
	verificado    = "verificado"
	noVerificado  = "no-verificado"
)

// Los niveles de un boletín aplicable (data-model §2.3).
const (
	nivelEstatal    = "estatal"
	nivelAutonomico = "autonomico"
	nivelProvincial = "provincial"
)

// Territorio es el territorio de un municipio resuelto: exactamente ocho
// claves, cada dato con el source que lo sostiene (FR-005, FR-006). Lo
// devuelve Resolver; su valor cero no es ningún territorio.
type Territorio struct {
	// Municipio es el municipio, con su nombre oficial.
	Municipio Municipio `json:"municipio"`
	// CodigoINE es su código INE, con el dígito de control oficial.
	CodigoINE CodigoINE `json:"codigo_ine"`
	// Provincia es la provincia del municipio.
	Provincia Provincia `json:"provincia"`
	// Comunidad es la comunidad o ciudad autónoma de su provincia.
	Comunidad Comunidad `json:"comunidad"`
	// DIR3 es el de su ayuntamiento, solo si está verificado.
	DIR3 DIR3 `json:"dir3"`
	// Regimen es el de su comunidad, dato nacional (FR-055).
	Regimen Regimen `json:"regimen"`
	// Boletines son los aplicables: siempre el estatal y después solo los
	// niveles que su comunidad tiene configurados (FR-008, FR-021).
	Boletines []Boletin `json:"boletines" jsonschema:"minItems=1"`
	// Cobertura dice de cada aspecto que el registro puede llenar si lo tiene
	// configurado o verificado (FR-020).
	Cobertura Cobertura `json:"cobertura"`

	// fecha es la de la respuesta: no es contenido, así que no se serializa.
	fecha time.Time
}

// Fecha es la de la respuesta: la más antigua de las de los ficheros que la
// sostienen —la relación de municipios, el fichero de la comunidad del
// municipio, el del estado y, solo si trae su DIR3, la correspondencia—, a
// medianoche UTC. Es la fecha de consulta del sobre, que así nunca aparenta
// más frescura que la parte más vieja de lo que cita y no depende del reloj
// (data-model §2.8, research.md D6).
func (t Territorio) Fecha() time.Time {
	return t.fecha
}

// Municipio es el municipio resuelto.
type Municipio struct {
	// Nombre es el nombre oficial, tal como lo escribe la relación del INE.
	Nombre string `json:"nombre" jsonschema:"minLength=1"`
	// Source es el de la relación.
	Source string `json:"source" jsonschema:"minLength=1"`
}

// CodigoINE es el código INE del municipio y su dígito de control oficial.
type CodigoINE struct {
	// Codigo son las cinco cifras, PPMMM.
	Codigo string `json:"codigo" jsonschema:"pattern=^[0-9]{5}$"`
	// DigitoDeControl es el oficial de la relación, una cifra.
	DigitoDeControl string `json:"digito_de_control" jsonschema:"pattern=^[0-9]$"`
	// Source es el de la relación.
	Source string `json:"source" jsonschema:"minLength=1"`
}

// Provincia es la provincia del municipio.
type Provincia struct {
	// Codigo son sus dos cifras, las primeras del código INE.
	Codigo string `json:"codigo" jsonschema:"pattern=^[0-9]{2}$"`
	// Nombre es el que da el fichero de su comunidad.
	Nombre string `json:"nombre" jsonschema:"minLength=1"`
	// Source es el de los nombres del fichero de su comunidad.
	Source string `json:"source" jsonschema:"minLength=1"`
}

// Comunidad es la comunidad o ciudad autónoma de la provincia.
type Comunidad struct {
	// Codigo son sus dos cifras.
	Codigo string `json:"codigo" jsonschema:"pattern=^[0-9]{2}$"`
	// Nombre es el que da su fichero.
	Nombre string `json:"nombre" jsonschema:"minLength=1"`
	// Source es el de los nombres de su fichero.
	Source string `json:"source" jsonschema:"minLength=1"`
}

// DIR3 es el código DIR3 del ayuntamiento. Sin correspondencia verificada va
// vacío entero, código y source, y la cobertura lo declara no verificado:
// nunca se emite un código derivado como si fuera registral (FR-023).
type DIR3 struct {
	// Codigo es L01PPMMMD, o vacío si no está verificado.
	Codigo string `json:"codigo" jsonschema:"pattern=^(L01[0-9]{6})?$"`
	// Source es el de la correspondencia, o vacío si no está verificado.
	Source string `json:"source"`
}

// Regimen es el régimen de la comunidad del municipio.
type Regimen struct {
	// Valor es comun o foral.
	Valor string `json:"valor" jsonschema:"enum=comun,enum=foral"`
	// Source es la ruta del fichero de la comunidad, que lo fija.
	Source string `json:"source" jsonschema:"minLength=1"`
}

// Boletin es un boletín aplicable al municipio en uno de sus niveles.
type Boletin struct {
	// Nivel es estatal, autonomico o provincial.
	Nivel string `json:"nivel" jsonschema:"enum=estatal,enum=autonomico,enum=provincial"`
	// Codigo es el del boletín.
	Codigo string `json:"codigo" jsonschema:"minLength=1"`
	// Nombre es el oficial.
	Nombre string `json:"nombre" jsonschema:"minLength=1"`
	// URL es la dirección pública del boletín.
	URL string `json:"url" jsonschema:"minLength=1,format=uri"`
	// Motivo va vacío salvo cuando un mismo boletín cubre dos niveles:
	// entonces es la razón que fija la configuración (FR-052).
	Motivo string `json:"motivo"`
	// Source es la ruta del fichero que lo fija.
	Source string `json:"source" jsonschema:"minLength=1"`
}

// Cobertura dice, de cada aspecto del territorio que el registro puede llenar,
// si lo tiene configurado o verificado. Sus tres claves van siempre
// (FR-020).
type Cobertura struct {
	// BoletinAutonomico es configurado o no-configurado.
	BoletinAutonomico string `json:"boletin_autonomico" jsonschema:"enum=configurado,enum=no-configurado"`
	// BoletinProvincial es configurado o no-configurado.
	BoletinProvincial string `json:"boletin_provincial" jsonschema:"enum=configurado,enum=no-configurado"`
	// DIR3 es verificado o no-verificado.
	DIR3 string `json:"dir3" jsonschema:"enum=verificado,enum=no-verificado"`
}

// AspectoDeCobertura es una clave de la cobertura con los valores que puede
// tomar.
type AspectoDeCobertura struct {
	// Clave es la de la cobertura en el territorio resuelto.
	Clave string
	// Valores son los que puede tomar, en el orden de su enumerado.
	Valores []string
}

// AspectosDeCobertura es el vocabulario cerrado de la cobertura, en el orden
// de sus claves: la única fuente de verdad de lo que un esperado de eval puede
// declarar sobre ella, como boe.CodigosDeAviso lo es de los avisos (contrato de
// evals §1.2). Cada llamada devuelve un vocabulario nuevo, que quien lo recibe
// puede cambiar sin cambiar el de nadie más.
func AspectosDeCobertura() []AspectoDeCobertura {
	return []AspectoDeCobertura{
		{Clave: "boletin_autonomico", Valores: []string{configurado, noConfigurado}},
		{Clave: "boletin_provincial", Valores: []string{configurado, noConfigurado}},
		{Clave: "dir3", Valores: []string{verificado, noVerificado}},
	}
}

// territorioDe compone el territorio resuelto de un municipio del registro.
func (r *Registro) territorioDe(municipio *municipioRegistrado) Territorio {
	provincia := municipio.provincia
	comunidad := provincia.comunidad
	fechas := []time.Time{r.relacion.fecha, comunidad.origen.fecha, r.estado.fecha}

	dir3, coberturaDelDIR3 := DIR3{}, noVerificado
	if municipio.dir3 != nil {
		dir3 = DIR3{Codigo: municipio.dir3.String(), Source: r.correspondencia.source}
		coberturaDelDIR3 = verificado
		fechas = append(fechas, r.correspondencia.fecha)
	}

	return Territorio{
		Municipio: Municipio{Nombre: municipio.nombre, Source: r.relacion.source},
		CodigoINE: CodigoINE{
			Codigo:          municipio.codigo.String(),
			DigitoDeControl: municipio.digito,
			Source:          r.relacion.source,
		},
		Provincia: Provincia{Codigo: provincia.codigo, Nombre: provincia.nombre, Source: comunidad.origen.source},
		Comunidad: Comunidad{Codigo: comunidad.codigo, Nombre: comunidad.nombre, Source: comunidad.origen.source},
		DIR3:      dir3,
		Regimen:   Regimen{Valor: comunidad.regimen, Source: comunidad.ruta},
		Boletines: r.boletinesDe(comunidad),
		Cobertura: Cobertura{
			BoletinAutonomico: coberturaDeBoletin(comunidad.autonomico),
			BoletinProvincial: coberturaDeBoletin(comunidad.provincial),
			DIR3:              coberturaDelDIR3,
		},
		fecha: fechaMasAntigua(fechas[0], fechas[1:]...),
	}
}

// boletinesDe son los boletines aplicables en una comunidad: el estatal y
// después, en su orden, los niveles que la comunidad tiene configurados, y
// ninguno más. La lista es nueva en cada respuesta: cambiar una no cambia las
// siguientes.
func (r *Registro) boletinesDe(comunidad *comunidadRegistrada) []Boletin {
	boletines := []Boletin{r.estatal}

	for _, nivel := range []*Boletin{comunidad.autonomico, comunidad.provincial} {
		if nivel != nil {
			boletines = append(boletines, *nivel)
		}
	}

	return boletines
}

// coberturaDeBoletin dice si un nivel de boletín está configurado.
func coberturaDeBoletin(boletin *Boletin) string {
	if boletin == nil {
		return noConfigurado
	}

	return configurado
}

// fechaMasAntigua devuelve la más antigua de las fechas dadas.
func fechaMasAntigua(primera time.Time, resto ...time.Time) time.Time {
	masAntigua := primera

	for _, fecha := range resto {
		if fecha.Before(masAntigua) {
			masAntigua = fecha
		}
	}

	return masAntigua
}
