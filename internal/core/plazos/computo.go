package plazos

import "time"

// ultimoDia es el 31 de diciembre de 9999, el último que se escribe
// AAAA-MM-DD: un plazo que acaba después no tiene fecha que dar.
var ultimoDia = time.Date(9999, time.December, 31, 0, 0, 0, 0, time.UTC)

// segundosPorDia son los de un día del calendario, sin cambios de hora: las
// fechas del cómputo son medianoches UTC.
const segundosPorDia = 24 * 60 * 60

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

// computar da el día del plazo según su cómputo (contracts/applet-plazos.md
// §2): el contencioso, el de la LPAC y, en el silencio, el día siguiente al
// vencimiento del plazo máximo así calculado (research D8). cabe es falso si
// el día pasa del 31 de diciembre de 9999.
func (c calculo) computar(inicio time.Time, duracion Duracion) (computo, bool) {
	if c.tipo.Computo == ComputoContencioso {
		return c.contencioso(inicio, duracion.Cantidad)
	}

	vencimiento, cabe := c.lpac(inicio, duracion)
	if !cabe || c.tipo.Codigo != TipoSilencio {
		return vencimiento, cabe
	}

	vencimiento.dia = vencimiento.dia.AddDate(0, 0, 1)

	return vencimiento, !vencimiento.dia.After(ultimoDia)
}

// lpac cuenta un plazo de la LPAC desde el día siguiente al de inicio (arts.
// 30.3 y 30.4; FR-002): en días hábiles, el enésimo hábil (30.2); en días
// naturales, el enésimo día, y en meses, el mismo día del mes de vencimiento o
// el último si no lo tiene (30.4); en los dos, si es inhábil, el primer hábil
// siguiente (30.5; research D7).
func (c calculo) lpac(inicio time.Time, duracion Duracion) (computo, bool) {
	if duracion.Unidad == UnidadDiasHabiles {
		return c.enesimoHabil(inicio, duracion.Cantidad)
	}

	vencimiento, cabe := deFechaAFecha(inicio, duracion)
	if !cabe {
		return computo{}, false
	}

	return c.prorrogar(vencimiento)
}

// contencioso cuenta el plazo del recurso contencioso-administrativo (research
// D9): en meses de fecha a fecha (art. 46.1 LJCA; art. 185.1 LOPJ); si el
// inicio o el día siguiente es de agosto, desde el 31 de agosto; si el tramo
// pasa por agosto, que no corre (art. 128.2 LJCA), hasta el mismo día del mes
// siguiente; y con el último día inhábil del art. 182.1 LOPJ prorrogado al
// primer hábil (art. 185.2). Los días del 24 de diciembre al 6 de enero se
// cuentan, y ninguno es inhábil por serlo: el art. 182.1 grabado, en la
// redacción de la LO 14/2022, ya no enumera el 24 ni el 31 de diciembre
// (research S4).
func (c calculo) contencioso(inicio time.Time, meses int) (computo, bool) {
	partida := inicio
	if inicio.Month() == time.August || inicio.AddDate(0, 0, 1).Month() == time.August {
		partida = time.Date(inicio.Year(), time.August, 31, 0, 0, 0, 0, time.UTC)
	}

	vencimiento, cabe := sumarMeses(partida, meses)
	if cabe && pasaPor(partida.AddDate(0, 0, 1), vencimiento, time.August, 1, time.August, 31) {
		vencimiento, cabe = sumarMeses(partida, meses+1)
	}

	if !cabe {
		return computo{}, false
	}

	return c.prorrogar(vencimiento)
}

// inhabil dice si un día es inhábil en el cómputo del plazo: el sábado, el
// domingo y los festivos de alguno de los municipios, en la LPAC (arts. 30.2 y
// 30.6), y los mismos de la sede en el contencioso, que son los que enumera el
// art. 182.1 LOPJ grabado.
func (c calculo) inhabil(dia time.Time) bool {
	return c.calendario.Inhabil(dia, c.delCalendario)
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

// festivosSaltados son los festivos de un día inhábil que el cómputo saltó:
// ninguno en sábado ni en domingo, que no son festivos aunque lo sean
// (contracts/applet-plazos.md §2).
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

// deFechaAFecha es el último día de un plazo en días naturales o en meses,
// antes de la prórroga; cabe es falso si pasa del 31 de diciembre de 9999.
func deFechaAFecha(inicio time.Time, duracion Duracion) (time.Time, bool) {
	if duracion.Unidad == UnidadDiasNaturales {
		if duracion.Cantidad > diasEntre(inicio, ultimoDia) {
			return time.Time{}, false
		}

		return inicio.AddDate(0, 0, duracion.Cantidad), true
	}

	return sumarMeses(inicio, duracion.Cantidad)
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

// pasaPor dice si algún día de desde a hasta, los dos incluidos, cae entre el
// día inicial de un año y el final, que puede ser del año siguiente, como del
// 24 de diciembre al 6 de enero.
func pasaPor(desde, hasta time.Time, mesInicial time.Month, diaInicial int, mesFinal time.Month, diaFinal int) bool {
	saltoDeAnio := 0
	if mesFinal < mesInicial {
		saltoDeAnio = 1
	}

	for anio := desde.Year() - saltoDeAnio; anio <= hasta.Year(); anio++ {
		primero := time.Date(anio, mesInicial, diaInicial, 0, 0, 0, 0, time.UTC)
		ultimo := time.Date(anio+saltoDeAnio, mesFinal, diaFinal, 0, 0, 0, 0, time.UTC)

		if !hasta.Before(primero) && !desde.After(ultimo) {
			return true
		}
	}

	return false
}
