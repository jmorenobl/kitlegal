package plazos

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Los cuatro códigos de aviso (contracts/applet-plazos.md §5; research D6).
const (
	AvisoFaltanNacionalesYAutonomicos = "faltan-festivos-nacionales-y-autonomicos"
	AvisoFaltanLocales                = "faltan-festivos-locales"
	AvisoDiasDelArt183                = "dias-del-art-183-lopj"
	AvisoFestivoDeUnaParte            = "festivo-de-una-parte"
)

// CodigosDeAviso son los cuatro códigos de aviso, en el orden en que van en el
// plazo. Devuelve una lista nueva en cada llamada.
func CodigosDeAviso() []string {
	return []string{AvisoFaltanNacionalesYAutonomicos, AvisoFaltanLocales, AvisoDiasDelArt183, AvisoFestivoDeUnaParte}
}

// EtiquetaDeAviso es la etiqueta de un código de aviso, la que su texto lleva
// tras «⚠ » y antes de «:», como la forma fija que comprueban las evals
// (FR-048); con lo que no es un código, nada.
func EtiquetaDeAviso(codigo string) (string, bool) {
	switch codigo {
	case AvisoFaltanNacionalesYAutonomicos:
		return "FALTAN FESTIVOS NACIONALES Y AUTONÓMICOS", true
	case AvisoFaltanLocales:
		return "FALTAN FESTIVOS LOCALES", true
	case AvisoDiasDelArt183:
		return "DÍAS DEL ART. 183 LOPJ", true
	case AvisoFestivoDeUnaParte:
		return "FESTIVO DE UNA PARTE", true
	}

	return "", false
}

// Las frases fijas de los textos de los avisos.
const (
	sinEllos       = "; la fecha se ha calculado sin ellos y puede ser anterior a la real."
	textoDelArt183 = "el plazo pasa por días del 24 de diciembre al 6 de enero, que el art. 183 LOPJ declara " +
		"inhábiles para las actuaciones judiciales; se han contado, salvo los inhábiles del art. 182.1 LOPJ, porque " +
		"el art. 128.2 LJCA no los excluye del plazo para interponer el recurso. La fecha es la más temprana posible."
)

// avisoCon es un aviso con su código y su texto: una línea que empieza por
// «⚠ <ETIQUETA>: ».
func avisoCon(codigo, cuerpo string) Aviso {
	etiqueta, _ := EtiquetaDeAviso(codigo)

	return Aviso{Codigo: codigo, Texto: "⚠ " + etiqueta + ": " + cuerpo}
}

// segmento son años seguidos sin festivos de un ámbito, con un mismo motivo.
type segmento struct {
	desde int
	hasta int
	falta Falta
}

// segmentosSinFestivos son los años sin festivos de un ámbito y, en uno local,
// de un municipio, de los tramos de la cobertura: los tramos con falta, con los
// seguidos del mismo motivo juntos.
func segmentosSinFestivos(tramos []Tramo, ambito AmbitoDeCobertura, municipio string) []segmento {
	var segmentos []segmento

	for _, tramo := range tramos {
		if tramo.Ambito != ambito || tramo.Municipio != municipio || tramo.Falta == "" {
			continue
		}

		if ultimo := len(segmentos) - 1; ultimo >= 0 && segmentos[ultimo].falta == tramo.Falta &&
			segmentos[ultimo].hasta+1 == tramo.Desde {
			segmentos[ultimo].hasta = tramo.Hasta

			continue
		}

		segmentos = append(segmentos, segmento{desde: tramo.Desde, hasta: tramo.Hasta, falta: tramo.Falta})
	}

	return segmentos
}

// aniosDe escribe los años de un segmento como los nombra un aviso: «de 2027»,
// «de 2028 a 2030» o, antes del primer año cargado de su ámbito, «anteriores a
// 2025». La preposición va con los años, para que «los de <años>» de
// contracts/applet-plazos.md §5 no dé «los de de 2028» ni «los de anteriores».
func aniosDe(s segmento, alcance *Alcance) string {
	switch {
	case alcance != nil && s.hasta < alcance.Desde:
		return fmt.Sprintf("anteriores a %d", alcance.Desde)
	case s.desde == s.hasta:
		return fmt.Sprintf("de %d", s.desde)
	default:
		return fmt.Sprintf("de %d a %d", s.desde, s.hasta)
	}
}

// motivoDe es por qué faltan los festivos, tras los años: nada con un año sin
// cargar, y una frase con los demás motivos de la cobertura.
func motivoDe(falta Falta, comunidad string) string {
	switch falta {
	case FaltaSinFestivosLocales:
		return ": no tiene festivos locales de " + comunidad
	case FaltaNoComunicado:
		return ": el boletín los da como no comunicados por el ayuntamiento"
	case FaltaLineaSinLeer:
		return ": su línea del boletín no se lee sin interpretarla"
	case FaltaSinLinea:
		return ": el boletín no trae su línea"
	default:
		return ""
	}
}

// sinFestivos escribe los segmentos de un aviso, cada uno con sus años y su
// motivo, unidos por « ni ».
func sinFestivos(segmentos []segmento, alcance *Alcance, comunidad string) string {
	partes := make([]string, 0, len(segmentos))
	for _, s := range segmentos {
		partes = append(partes, aniosDe(s, alcance)+motivoDe(s.falta, comunidad))
	}

	return strings.Join(partes, " ni ")
}

// avisoDeLosNacionales es el aviso de los festivos nacionales y autonómicos que
// faltan, con hasta dónde llegan los cargados.
func avisoDeLosNacionales(segmentos []segmento, alcance *Alcance) Aviso {
	cuerpo := "kitlegal no tiene los " + sinFestivos(segmentos, alcance, "") + sinEllos
	if alcance != nil {
		cuerpo += fmt.Sprintf(" Los que tiene llegan a %d (%s).", alcance.Hasta, alcance.Disposicion)
	}

	return avisoCon(AvisoFaltanNacionalesYAutonomicos, cuerpo)
}

// avisoDeLosLocales es el aviso de los festivos locales que faltan de un
// municipio, con hasta dónde llegan los cargados de su comunidad, si tiene.
func avisoDeLosLocales(municipio MunicipioResuelto, segmentos []segmento, alcance *Alcance) Aviso {
	cuerpo := fmt.Sprintf("kitlegal no tiene los de %s (%s) %s%s", municipio.Nombre, municipio.CodigoINE,
		sinFestivos(segmentos, alcance, municipio.NombreDeLaComunidad), sinEllos)
	if alcance != nil {
		cuerpo += fmt.Sprintf(" Los de %s llegan a %d (%s).", municipio.NombreDeLaComunidad, alcance.Hasta,
			alcance.Disposicion)
	}

	return avisoCon(AvisoFaltanLocales, cuerpo)
}

// avisoDelArt183 es el aviso de los días del 24 de diciembre al 6 de enero en
// el contencioso.
func avisoDelArt183() Aviso {
	return avisoCon(AvisoDiasDelArt183, textoDelArt183)
}

// avisoDeUnaParte es el aviso de un festivo de una parte de la comunidad de un
// municipio que el cómputo contó como hábil (FR-011).
func avisoDeUnaParte(dia time.Time, parte FestivoDeUnaParte, municipio MunicipioResuelto) Aviso {
	return avisoCon(AvisoFestivoDeUnaParte, fmt.Sprintf("el %s es fiesta solo en %s (%s), según %s; kitlegal no sabe "+
		"si %s está ahí y lo ha contado como hábil. La fecha es la más temprana posible.", dia.Format(time.DateOnly),
		parte.Parte, municipio.NombreDeLaComunidad, parte.Source, municipio.Nombre))
}

// alcanceSi es el alcance si lo hay.
func alcanceSi(alcance Alcance, cargado bool) *Alcance {
	if !cargado {
		return nil
	}

	return &alcance
}

// avisos son los del plazo, cada uno solo si el tramo pasa por lo que lo causa,
// en el orden de CodigosDeAviso y, dentro de un código, de fecha y de municipio
// (contracts/applet-plazos.md §4 y §5): el de los nacionales y autonómicos que
// faltan, uno por municipio de los locales que faltan, el del art. 183 LOPJ en
// el contencioso y uno por festivo de una parte y municipio. Ninguno nace de
// una entidad menor, que no es un municipio que cuente (FR-024).
func (c calculo) avisos(cobertura []Tramo) []Aviso {
	avisos := []Aviso{}

	if segmentos := segmentosSinFestivos(cobertura, CoberturaNacionalYAutonomica, ""); len(segmentos) > 0 {
		avisos = append(avisos, avisoDeLosNacionales(segmentos, alcanceSi(c.calendario.AlcanceNacional())))
	}

	avisos = append(avisos, c.avisosDeLosLocales(cobertura)...)

	if c.tipo.Computo == ComputoContencioso && pasaPor(c.desde, c.hasta, time.December, 24, time.January, 6) {
		avisos = append(avisos, avisoDelArt183())
	}

	return append(avisos, c.avisosDeUnaParte()...)
}

// avisosDeLosLocales son los de los festivos locales que faltan, uno por
// municipio, por el primer año que falta y después en el orden dado.
func (c calculo) avisosDeLosLocales(cobertura []Tramo) []Aviso {
	type avisoLocal struct {
		primerAnio int
		aviso      Aviso
	}

	var locales []avisoLocal

	for _, municipio := range c.municipios {
		segmentos := segmentosSinFestivos(cobertura, CoberturaLocal, municipio.CodigoINE)
		if len(segmentos) == 0 {
			continue
		}

		alcance := alcanceSi(c.calendario.AlcanceLocal(municipio.Comunidad))
		locales = append(locales, avisoLocal{
			primerAnio: segmentos[0].desde,
			aviso:      avisoDeLosLocales(municipio, segmentos, alcance),
		})
	}

	slices.SortStableFunc(locales, func(a, b avisoLocal) int { return cmp.Compare(a.primerAnio, b.primerAnio) })

	avisos := make([]Aviso, 0, len(locales))
	for _, local := range locales {
		avisos = append(avisos, local.aviso)
	}

	return avisos
}

// avisosDeUnaParte son los de los festivos de una parte de la comunidad de
// algún municipio en los días del tramo que el cómputo contó como hábiles: uno
// por festivo y municipio, por fecha, municipio en el orden dado y parte
// (FR-011). Un día inhábil por otra causa no se contó como hábil y no avisa.
func (c calculo) avisosDeUnaParte() []Aviso {
	var avisos []Aviso

	for _, dia := range c.calendario.diasConPartes(c.desde, c.hasta) {
		if c.inhabil(dia) {
			continue
		}

		for _, municipio := range c.municipios {
			for _, parte := range c.calendario.PartesDe(dia, municipio.Comunidad) {
				avisos = append(avisos, avisoDeUnaParte(dia, parte, municipio))
			}
		}
	}

	return avisos
}

// diasConPartes son los días de desde a hasta, los dos incluidos, con algún
// festivo de una parte de una comunidad, en orden. Mira los días con festivos
// cargados y no los del tramo, que puede ser de años.
func (c *Calendario) diasConPartes(desde, hasta time.Time) []time.Time {
	primero, ultimo := fechaDe(desde), fechaDe(hasta)

	var dias []time.Time

	for clave, filas := range c.nacionales {
		if clave.comparar(primero) < 0 || clave.comparar(ultimo) > 0 ||
			!slices.ContainsFunc(filas, func(fila filaNacional) bool { return fila.parte != "" }) {
			continue
		}

		dias = append(dias, time.Date(clave.anio, clave.mes, clave.dia, 0, 0, 0, 0, time.UTC))
	}

	slices.SortFunc(dias, time.Time.Compare)

	return dias
}

// comparar ordena dos días del calendario, como cmp.Compare.
func (f fecha) comparar(otra fecha) int {
	return cmp.Or(cmp.Compare(f.anio, otra.anio), cmp.Compare(f.mes, otra.mes), cmp.Compare(f.dia, otra.dia))
}
