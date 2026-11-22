package plazos

import (
	"cmp"
	"slices"
	"time"
)

// ultimoDia es el 31 de diciembre de 9999, el último que se escribe
// AAAA-MM-DD: un plazo que acaba después no tiene fecha que dar.
var ultimoDia = time.Date(9999, time.December, 31, 0, 0, 0, 0, time.UTC)

// segundosPorDia son los de un día del calendario, sin cambios de hora: las
// fechas del cómputo son medianoches UTC.
const segundosPorDia = 24 * 60 * 60

// formatoDelPeriodo es el de los días de un periodo del año, MM-DD, que
// ordenados como texto van en el orden de los días del año.
const formatoDelPeriodo = "01-02"

// computo es lo que da el cómputo: el día que se da y los festivos que saltó,
// en orden de fecha.
type computo struct {
	dia      time.Time
	festivos []festivoDelDia
}

// festivoDelDia es un festivo saltado con su día.
type festivoDelDia struct {
	dia time.Time
	Festivo
}

// computar da el día del plazo con su cómputo (spec §C.3 de H36; research D2):
// un solo algoritmo, que cuenta por la unidad de la duración con los periodos
// del cómputo del tipo y da el vencimiento o, si el tipo da dia-siguiente, el
// día siguiente al vencimiento, sin prórroga (FR-005). Ningún código de tipo ni
// de cómputo decide nada. cabe es falso si el día pasa del 31 de diciembre de
// 9999.
func (c calculo) computar(inicio time.Time, duracion Duracion) (computo, bool) {
	vencimiento, cabe := c.vencimiento(inicio, duracion)
	if !cabe || c.tipo.Da != daDiaSiguiente {
		return vencimiento, cabe
	}

	vencimiento.dia = vencimiento.dia.AddDate(0, 0, 1)

	return vencimiento, !vencimiento.dia.After(ultimoDia)
}

// vencimiento cuenta el plazo desde el día siguiente al de inicio: en días
// hábiles, el enésimo día que no es inhábil; en días naturales, el enésimo
// día, y en meses, el mismo día del mes de vencimiento o el último si no lo
// tiene, con los periodos que no corren; en los dos, si es inhábil, el primer
// día que no lo es.
func (c calculo) vencimiento(inicio time.Time, duracion Duracion) (computo, bool) {
	if duracion.Unidad == UnidadDiasHabiles {
		return c.enesimoHabil(inicio, duracion.Cantidad)
	}

	if duracion.Unidad == UnidadDiasNaturales {
		if duracion.Cantidad > diasEntre(inicio, ultimoDia) {
			return computo{}, false
		}

		return c.prorrogar(inicio.AddDate(0, 0, duracion.Cantidad))
	}

	return c.meses(inicio, duracion.Cantidad)
}

// meses cuenta un plazo en meses de fecha a fecha con los periodos de su
// cómputo que no corren, con la regla de research D9 de H9 escrita para
// cualquiera de ellos (contracts/api-del-dominio.md §2): si el inicio o el día
// siguiente es de uno, se parte de su último día; si del día siguiente a la
// partida al vencimiento pasa por alguno, el vencimiento es el mismo día del mes
// siguiente, o el último si no lo tiene; y el último día inhábil se prorroga al
// primero que no lo es. Con agosto, el contencioso de H9 tal cual.
func (c calculo) meses(inicio time.Time, meses int) (computo, bool) {
	noCorren := c.periodosQueNoCorren()

	partida := inicio
	for _, periodo := range noCorren {
		if pasaPor(inicio, inicio.AddDate(0, 0, 1), periodo) {
			partida = periodo.ultimoDiaDesde(inicio)
		}
	}

	vencimiento, cabe := sumarMeses(partida, meses)
	if cabe && slices.ContainsFunc(noCorren, func(periodo periodo) bool {
		return pasaPor(partida.AddDate(0, 0, 1), vencimiento, periodo)
	}) {
		vencimiento, cabe = sumarMeses(partida, meses+1)
	}

	if !cabe {
		return computo{}, false
	}

	return c.prorrogar(vencimiento)
}

// periodosQueNoCorren son los periodos del cómputo del tipo cuyos días no
// corren, en su orden.
func (c calculo) periodosQueNoCorren() []periodo {
	var noCorren []periodo

	for _, periodo := range c.tipo.computo.Periodos {
		if periodo.Efecto == efectoNoCorre {
			noCorren = append(noCorren, periodo)
		}
	}

	return noCorren
}

// inhabil dice si un día es inhábil en el cómputo del plazo (spec §C.3 y
// §C.4): el sábado, el domingo, los festivos nacionales y autonómicos de toda
// la comunidad de alguno de los municipios que cuentan y los locales de alguno
// —con los datos de hoy, en el contencioso, los que enumera el art. 182.1 LOPJ
// grabado— y los días de un periodo inhabil de su cómputo.
func (c calculo) inhabil(dia time.Time) bool {
	return c.calendario.Inhabil(dia, c.delCalendario) ||
		slices.ContainsFunc(c.tipo.computo.Periodos, func(periodo periodo) bool {
			return periodo.Efecto == efectoInhabil && periodo.contiene(dia)
		})
}

// enesimoHabil es el enésimo día hábil desde el día siguiente al de inicio,
// con los festivos que saltó. N días hábiles ocupan al menos N días: si no
// caben hasta el 31 de diciembre de 9999, no se recorren.
func (c calculo) enesimoHabil(inicio time.Time, habiles int) (computo, bool) {
	if habiles > diasEntre(inicio, ultimoDia) {
		return computo{}, false
	}

	var festivos []festivoDelDia

	dia := inicio

	for contados := 0; contados < habiles; {
		dia = dia.AddDate(0, 0, 1)
		if dia.After(ultimoDia) {
			return computo{}, false
		}

		if c.inhabil(dia) {
			festivos = append(festivos, c.festivosSaltados(dia)...)

			continue
		}

		contados++
	}

	return computo{dia: dia, festivos: festivos}, true
}

// prorrogar lleva un último día inhábil al primer hábil siguiente (art. 30.5
// LPAC; art. 185.2 LOPJ), con los festivos que saltó.
func (c calculo) prorrogar(dia time.Time) (computo, bool) {
	var festivos []festivoDelDia

	for c.inhabil(dia) {
		festivos = append(festivos, c.festivosSaltados(dia)...)

		dia = dia.AddDate(0, 0, 1)
		if dia.After(ultimoDia) {
			return computo{}, false
		}
	}

	return computo{dia: dia, festivos: festivos}, true
}

// festivosSaltados son los festivos de un día inhábil que el cómputo saltó,
// también el de un periodo inhabil: ninguno en sábado ni en domingo, que no
// son festivos aunque lo sean (spec §C.3).
func (c calculo) festivosSaltados(dia time.Time) []festivoDelDia {
	if semana := dia.Weekday(); semana == time.Saturday || semana == time.Sunday {
		return nil
	}

	festivos := c.calendario.FestivosDe(dia, c.delCalendario)
	saltados := make([]festivoDelDia, 0, len(festivos))

	for _, festivo := range festivos {
		saltados = append(saltados, festivoDelDia{dia: dia, Festivo: festivo})
	}

	return saltados
}

// sumarMeses es el mismo día que inicio tantos meses después o, si ese mes no
// lo tiene, su último día (art. 30.4 LPAC; art. 5.1 del Código Civil); cabe es
// falso si pasa del 31 de diciembre de 9999.
func sumarMeses(inicio time.Time, meses int) (time.Time, bool) {
	anio, mes, dia := inicio.Date()

	if meses > (ultimoDia.Year()-anio)*12+int(time.December-mes) {
		return time.Time{}, false
	}

	total := int(mes) - 1 + meses
	anioDeVencimiento, mesDeVencimiento := anio+total/12, time.Month(total%12+1)
	ultimoDelMes := time.Date(anioDeVencimiento, mesDeVencimiento+1, 0, 0, 0, 0, 0, time.UTC).Day()

	return time.Date(anioDeVencimiento, mesDeVencimiento, min(dia, ultimoDelMes), 0, 0, 0, 0, time.UTC), true
}

// diasEntre son los días de desde a hasta, sin contar desde. Se cuentan por
// segundos y no con time.Duration, que no llega a trescientos años.
func diasEntre(desde, hasta time.Time) int {
	return int((hasta.Unix() - desde.Unix()) / segundosPorDia)
}

// diaDelAnio es un día por su año y su MM-DD, que se ordenan como las fechas,
// también el año siguiente al 9999 de un periodo que cruza el año.
type diaDelAnio struct {
	anio    int
	mesYDia string
}

// diaDelAnioDe es un día del calendario como lo compara un periodo.
func diaDelAnioDe(dia time.Time) diaDelAnio {
	return diaDelAnio{anio: dia.Year(), mesYDia: dia.Format(formatoDelPeriodo)}
}

// comparar ordena dos días, como cmp.Compare.
func (d diaDelAnio) comparar(otro diaDelAnio) int {
	return cmp.Or(cmp.Compare(d.anio, otro.anio), cmp.Compare(d.mesYDia, otro.mesYDia))
}

// cruzaElAnio dice si el periodo acaba el año siguiente al que empieza, como
// del 24 de diciembre al 6 de enero.
func (p periodo) cruzaElAnio() bool {
	return p.Hasta < p.Desde
}

// contiene dice si un día es del periodo.
func (p periodo) contiene(dia time.Time) bool {
	mesYDia := dia.Format(formatoDelPeriodo)
	if p.cruzaElAnio() {
		return mesYDia >= p.Desde || mesYDia <= p.Hasta
	}

	return mesYDia >= p.Desde && mesYDia <= p.Hasta
}

// ultimoDiaDesde es el primer día, desde uno dado, que es del periodo y cuyo
// siguiente no lo es: desde un día del periodo o desde el anterior a su primer
// día, su último día.
func (p periodo) ultimoDiaDesde(dia time.Time) time.Time {
	for !p.contiene(dia) || p.contiene(dia.AddDate(0, 0, 1)) {
		dia = dia.AddDate(0, 0, 1)
	}

	return dia
}

// pasaPor dice si algún día de desde a hasta, los dos incluidos, es de un
// periodo del año, que puede cruzar el año, como del 24 de diciembre al 6 de
// enero (contracts/api-del-dominio.md §2): mira el periodo de cada año que toca
// el tramo.
func pasaPor(desde, hasta time.Time, p periodo) bool {
	saltoDeAnio := 0
	if p.cruzaElAnio() {
		saltoDeAnio = 1
	}

	primero, ultimo := diaDelAnioDe(desde), diaDelAnioDe(hasta)

	for anio := desde.Year() - saltoDeAnio; anio <= hasta.Year(); anio++ {
		empieza := diaDelAnio{anio: anio, mesYDia: p.Desde}
		acaba := diaDelAnio{anio: anio + saltoDeAnio, mesYDia: p.Hasta}

		if ultimo.comparar(empieza) >= 0 && primero.comparar(acaba) <= 0 {
			return true
		}
	}

	return false
}
