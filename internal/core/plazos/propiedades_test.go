package plazos

import (
	_ "embed"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tiposDePruebaDelCivil es el fichero de los dos tipos de prueba del cómputo
// civil de FR-023 de H36, sin duración y con da: vencimiento: uno en días
// hábiles, con el 133.2, y otro en meses, con el 133.3
// (contracts/datos-de-plazos.md §4 de H36).
//
//go:embed testdata/tipos-de-prueba-civil.yaml
var tiposDePruebaDelCivil []byte

// Lo que las propiedades del civil esperan de él, escrito aquí otra vez y no
// tomado del código que se prueba: sus dos tipos de prueba y sus periodos
// inhabil, los de la copia del contrato (contracts/datos-de-plazos.md §1 y §4
// de H36).
const (
	// computoCivil es el código del cómputo civil de la LEC.
	computoCivil Computo = "civil"
	// pruebaCivilEnDias es el tipo de prueba del civil en días hábiles.
	pruebaCivilEnDias = "prueba-civil-dias"
	// pruebaCivilEnMeses es el tipo de prueba del civil en meses.
	pruebaCivilEnMeses = "prueba-civil-meses"
)

// periodoDelOraculo es un periodo inhabil del civil, de un MM-DD a otro, los
// dos incluidos; si el último va antes que el primero, cruza el año.
type periodoDelOraculo struct {
	desde, hasta string
}

// periodosInhabilesDelCivil son los dos periodos inhabil del art. 130.2 LEC:
// agosto y del 24 de diciembre al 6 de enero.
var periodosInhabilesDelCivil = []periodoDelOraculo{{desde: "08-01", hasta: "08-31"}, {desde: "12-24", hasta: "01-06"}}

// La rejilla de las propiedades del cómputo (research D22): cada fecha de
// inicio de 2025 a 2027, cada duración de 1 a 40 y cada pareja de un juego de
// tres municipios sintéticos, sin muestreo ni semilla. Los festivos son
// inventados y se repiten cada año: nacionales en las dos comunidades,
// autonómicos distintos en cada una, uno de una parte de la 82, que no hace
// inhábil el día, y locales distintos en los dos municipios de la 81, la única
// con fichero local.
var (
	primerInicioDeLaRejilla = time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	ultimoInicioDeLaRejilla = time.Date(2027, time.December, 31, 0, 0, 0, 0, time.UTC)
	duracionMaximaRejilla   = 40

	nacionalesDeLaRejilla  = []string{"01-01", "01-06", "05-01", "08-15", "10-12", "11-01", "12-06", "12-08", "12-25"}
	autonomicosDeLaRejilla = map[string][]string{"81": {"03-19", "05-02"}, "82": {"04-23", "09-08"}}
	partesDeLaRejilla      = map[string][]string{"82": {"06-24"}}
	localesDeLaRejilla     = map[string][]string{"81001": {"05-15", "11-09"}, "81002": {"06-13", "10-09"}}

	municipiosDeLaRejilla = []MunicipioResuelto{
		resuelto("81001", "Lugar Uno", "81", "Comunidad de la Rejilla"),
		resuelto("81002", "Lugar Dos", "81", "Comunidad de la Rejilla"),
		resuelto("82001", "Lugar Tres", "82", "Otra Comunidad de la Rejilla"),
	}
)

// rejilla es el calendario de las propiedades, cargado de sus ficheros, y el
// oráculo que dice qué días son inhábiles sin pasar por él.
type rejilla struct {
	calendario *Calendario
	tipos      []Tipo
	// fiestas lleva de cada día AAAA-MM-DD a las comunidades y los municipios
	// en que es fiesta en todo su territorio.
	fiestas map[string]map[string]bool
}

// nuevaRejilla escribe los dos ficheros de festivos de la rejilla en su forma
// canónica, los carga y apunta las mismas fiestas en el oráculo.
func nuevaRejilla(t *testing.T) rejilla {
	t.Helper()

	r := rejilla{tipos: tiposDelContrato(t), fiestas: map[string]map[string]bool{}}
	nacionales := FestivosNacionales{Fuente: "prueba.fiestas"}
	locales := FestivosLocales{Fuente: "prueba.fiestas-locales", Comunidad: "81"}

	for anio := 2025; anio <= 2027; anio++ {
		disposicion := fmt.Sprintf("PRUEBA-REJILLA-%d", anio)
		nacionales.Anios = append(nacionales.Anios, AnioNacional{Anio: anio, Disposicion: disposicion})
		locales.Anios = append(locales.Anios, AnioLocal{Anio: anio, Disposiciones: []string{disposicion + "-L"}})

		for _, dia := range nacionalesDeLaRejilla {
			nacionales.Festivos = append(nacionales.Festivos, r.fiesta(anio, dia, AmbitoNacional, "", disposicion, "81", "82"))
		}

		for comunidad, dias := range autonomicosDeLaRejilla {
			for _, dia := range dias {
				nacionales.Festivos = append(nacionales.Festivos, r.fiesta(anio, dia, AmbitoAutonomico, "", disposicion, comunidad))
			}
		}

		for comunidad, dias := range partesDeLaRejilla {
			for _, dia := range dias {
				nacionales.Festivos = append(nacionales.Festivos,
					r.fiesta(anio, dia, AmbitoAutonomico, "Isla de la Rejilla", disposicion, comunidad))
			}
		}

		for municipio, dias := range localesDeLaRejilla {
			for _, dia := range dias {
				fecha := fmt.Sprintf("%d-%s", anio, dia)
				r.apuntar(fecha, municipio)
				locales.Festivos = append(locales.Festivos, FestivoLocal{Fecha: fecha, Municipio: municipio, Source: disposicion + "-L"})
			}
		}
	}

	cabecera := Cabecera{Desde: "evidencias/prueba/", Test: "TestPropiedadesDelComputo", Paquete: "internal/core/plazos"}
	ficheroNacional, err := EscribirFestivosNacionales(nacionales, cabecera)
	require.NoError(t, err)
	ficheroLocal, err := EscribirFestivosLocales(locales, cabecera)
	require.NoError(t, err)

	r.calendario, err = CargarCalendario(ficheroNacional, map[string][]byte{"81": ficheroLocal})
	require.NoError(t, err)

	return r
}

// fiesta es una fila del fichero nacional; si no es de una parte, la apunta en
// el oráculo para sus comunidades.
func (r rejilla) fiesta(anio int, dia string, ambito Ambito, parte, source string, comunidades ...string) FestivoNacional {
	fecha := fmt.Sprintf("%d-%s", anio, dia)
	if parte == "" {
		r.apuntar(fecha, comunidades...)
	}

	return FestivoNacional{Fecha: fecha, Ambito: ambito, Comunidades: comunidades, Parte: parte, Source: source}
}

// apuntar anota en el oráculo que un día es fiesta en unas comunidades o unos
// municipios.
func (r rejilla) apuntar(fecha string, territorios ...string) {
	if r.fiestas[fecha] == nil {
		r.fiestas[fecha] = map[string]bool{}
	}

	for _, territorio := range territorios {
		r.fiestas[fecha][territorio] = true
	}
}

// inhabil es el oráculo: sábado, domingo o fiesta de la comunidad o del
// municipio (arts. 30.2 y 30.6 LPAC).
func (r rejilla) inhabil(dia time.Time, municipio MunicipioResuelto) bool {
	if dia.Weekday() == time.Saturday || dia.Weekday() == time.Sunday {
		return true
	}

	fiestas := r.fiestas[dia.Format(time.DateOnly)]

	return fiestas[municipio.Comunidad] || fiestas[municipio.CodigoINE]
}

// inhabilParaLosDos es el art. 30.6: inhábil en alguno de los dos.
func (r rejilla) inhabilParaLosDos(dia time.Time, uno, otro MunicipioResuelto) bool {
	return r.inhabil(dia, uno) || r.inhabil(dia, otro)
}

// calcular da la fecha de un plazo genérico de la rejilla.
func (r rejilla) calcular(inicio time.Time, duracion int, unidad Unidad, uno, otro MunicipioResuelto) (time.Time, error) {
	plazo, err := Calcular(r.calendario, r.tipos, Peticion{
		Tipo: "generico", Inicio: inicio.Format(time.DateOnly), Duracion: strconv.Itoa(duracion),
		Unidad: string(unidad), Municipios: los(uno, otro),
	})
	if err != nil {
		return time.Time{}, err
	}

	return time.Parse(time.DateOnly, plazo.Fecha)
}

// TestPropiedadesDelComputo calcula la rejilla entera y exige, sin un solo
// contraejemplo, las tres propiedades de FR-081 (SC-004): (1) N días hábiles
// nunca acaban en un día inhábil de ninguno de los dos municipios; (2) entre el
// día siguiente al de inicio y la fecha hay exactamente N días hábiles, y la
// fecha es hábil; (3) un plazo de N meses acaba el mismo día del mes de
// vencimiento, o el último si no lo tiene, o el primer hábil siguiente si ese
// es inhábil. Y, sobre la misma rejilla con un municipio cada vez, la sede, y
// los dos tipos de prueba del cómputo civil, las tres de FR-024 de H36
// (SC-004): (4) la fecha nunca es inhábil para la sede ni cae en un periodo
// inhabil; (5) en días hábiles, entre el día siguiente al de inicio y la fecha
// hay exactamente N días que no son inhábiles de base ni de un periodo; (6) en
// meses, la fecha es la de fecha a fecha o, si esa es inhábil, el primer hábil
// siguiente.
func TestPropiedadesDelComputo(t *testing.T) {
	t.Parallel()

	r := nuevaRejilla(t)

	for indice, uno := range municipiosDeLaRejilla {
		for _, otro := range municipiosDeLaRejilla[indice:] {
			pareja := uno.CodigoINE + "-" + otro.CodigoINE

			t.Run("días hábiles "+pareja, func(t *testing.T) {
				t.Parallel()
				assert.Empty(t, primeros(r.recorrer(UnidadDiasHabiles, uno, otro, r.incumpleLosHabiles)))
			})
			t.Run("meses "+pareja, func(t *testing.T) {
				t.Parallel()
				assert.Empty(t, primeros(r.recorrer(UnidadMeses, uno, otro, r.incumpleLosMeses)))
			})
		}
	}

	civiles := tiposConLosDePruebaDelCivil(t)

	for _, sede := range municipiosDeLaRejilla {
		t.Run("civil días hábiles "+sede.CodigoINE, func(t *testing.T) {
			t.Parallel()
			assert.Empty(t, primeros(r.recorrerElCivil(civiles, pruebaCivilEnDias, UnidadDiasHabiles, sede,
				r.incumpleLosHabilesDelCivil)))
		})
		t.Run("civil meses "+sede.CodigoINE, func(t *testing.T) {
			t.Parallel()
			assert.Empty(t, primeros(r.recorrerElCivil(civiles, pruebaCivilEnMeses, UnidadMeses, sede,
				r.incumpleLosMesesDelCivil)))
		})
	}
}

// primeros son los diez primeros contraejemplos y cuántos hay, para que un
// fallo no vuelque la rejilla entera.
func primeros(contraejemplos []string) []string {
	if len(contraejemplos) <= 10 {
		return contraejemplos
	}

	return append(contraejemplos[:10], fmt.Sprintf("… y %d más", len(contraejemplos)-10))
}

// recorrer calcula cada entrada de la rejilla en una unidad para una pareja y
// da sus contraejemplos: las entradas que no dan fecha y las que incumple, que
// dice por qué una fecha incumple sus propiedades o nada si las cumple.
func (r rejilla) recorrer(unidad Unidad, uno, otro MunicipioResuelto,
	incumple func(inicio time.Time, duracion int, fecha time.Time, uno, otro MunicipioResuelto) string,
) []string {
	var contraejemplos []string

	for inicio := primerInicioDeLaRejilla; !inicio.After(ultimoInicioDeLaRejilla); inicio = inicio.AddDate(0, 0, 1) {
		for duracion := 1; duracion <= duracionMaximaRejilla; duracion++ {
			var motivo string

			if fecha, err := r.calcular(inicio, duracion, unidad, uno, otro); err != nil {
				motivo = err.Error()
			} else {
				motivo = incumple(inicio, duracion, fecha, uno, otro)
			}

			if motivo != "" {
				contraejemplos = append(contraejemplos, fmt.Sprintf("%s y %d %s: %s", inicio.Format(time.DateOnly),
					duracion, unidad, motivo))
			}
		}
	}

	return contraejemplos
}

// incumpleLosHabiles dice qué propiedad de los días hábiles incumple una
// fecha, (1) o (2), o nada si cumple las dos.
func (r rejilla) incumpleLosHabiles(inicio time.Time, habiles int, fecha time.Time, uno, otro MunicipioResuelto) string {
	if r.inhabil(fecha, uno) || r.inhabil(fecha, otro) {
		return "(1) acaba en " + fecha.Format(time.DateOnly) + ", inhábil"
	}

	if contados := r.habilesEntre(inicio, fecha, uno, otro); contados != habiles {
		return fmt.Sprintf("(2) hasta %s hay %d hábiles", fecha.Format(time.DateOnly), contados)
	}

	return ""
}

// habilesEntre cuenta los días hábiles para los dos desde el día siguiente a
// inicio hasta fecha, los dos incluidos.
func (r rejilla) habilesEntre(inicio, fecha time.Time, uno, otro MunicipioResuelto) int {
	habiles := 0

	for dia := inicio.AddDate(0, 0, 1); !dia.After(fecha); dia = dia.AddDate(0, 0, 1) {
		if !r.inhabilParaLosDos(dia, uno, otro) {
			habiles++
		}
	}

	return habiles
}

// incumpleLosMeses dice por qué una fecha incumple la propiedad (3) —no es el
// mismo día del mes de vencimiento, el último o el primer hábil siguiente—, o
// nada si la cumple.
func (r rejilla) incumpleLosMeses(inicio time.Time, meses int, fecha time.Time, uno, otro MunicipioResuelto) string {
	primeroDelMes := time.Date(inicio.Year(), inicio.Month()+time.Month(meses), 1, 0, 0, 0, 0, time.UTC)
	ultimoDelMes := primeroDelMes.AddDate(0, 1, -1)
	vencimiento := primeroDelMes.AddDate(0, 0, min(inicio.Day(), ultimoDelMes.Day())-1)

	if fecha.Before(vencimiento) {
		return "(3) acaba en " + fecha.Format(time.DateOnly) + ", antes del " + vencimiento.Format(time.DateOnly)
	}

	for dia := vencimiento; dia.Before(fecha); dia = dia.AddDate(0, 0, 1) {
		if !r.inhabilParaLosDos(dia, uno, otro) {
			return "(3) acaba en " + fecha.Format(time.DateOnly) + ", después del " + dia.Format(time.DateOnly) +
				", hábil"
		}
	}

	if r.inhabilParaLosDos(fecha, uno, otro) {
		return "(3) acaba en " + fecha.Format(time.DateOnly) + ", inhábil"
	}

	return ""
}

// tiposConLosDePruebaDelCivil son los tipos de la copia del contrato con los
// dos de prueba del civil, combinados como en TestComputoCivil (research D7 de
// H36): la única línea «tipos:» de la copia, sustituida por el fichero de
// prueba, que lleva la suya y sus tipos con la misma sangría, leído con
// LeerTipos.
func tiposConLosDePruebaDelCivil(t *testing.T) []Tipo {
	t.Helper()

	base := "\n" + plazosDelContrato
	require.Equal(t, 1, strings.Count(base, "\ntipos:\n"), "la copia del contrato tiene una sola línea «tipos:»")

	tipos, err := LeerTipos([]byte(strings.Replace(base, "\ntipos:\n", "\n"+string(tiposDePruebaDelCivil), 1)))
	require.NoError(t, err, "la copia del contrato con los tipos de prueba del civil se lee sin defecto")
	require.Len(t, tipos, len(tiposDelContrato(t))+2, "la combinación da los dos tipos de prueba más que la copia")

	for _, codigo := range []string{pruebaCivilEnDias, pruebaCivilEnMeses} {
		indice := slices.IndexFunc(tipos, func(tipo Tipo) bool { return tipo.Codigo == codigo })
		require.GreaterOrEqual(t, indice, 0, "la combinación lleva el tipo de prueba %q", codigo)
		require.Equal(t, computoCivil, tipos[indice].Computo, "el tipo de prueba %q cuenta con el civil", codigo)
	}

	return tipos
}

// enUnPeriodoDelCivil dice si un día cae en un periodo inhabil del civil.
func enUnPeriodoDelCivil(dia time.Time) bool {
	mesYDia := dia.Format("01-02")

	return slices.ContainsFunc(periodosInhabilesDelCivil, func(periodo periodoDelOraculo) bool {
		if periodo.desde <= periodo.hasta {
			return periodo.desde <= mesYDia && mesYDia <= periodo.hasta
		}

		return mesYDia >= periodo.desde || mesYDia <= periodo.hasta
	})
}

// inhabilEnElCivil es el oráculo del civil: inhábil de base para la sede o en
// un periodo inhabil (arts. 130.2 y 133 LEC).
func (r rejilla) inhabilEnElCivil(dia time.Time, sede MunicipioResuelto) bool {
	return r.inhabil(dia, sede) || enUnPeriodoDelCivil(dia)
}

// calcularElCivil da la fecha de un plazo de un tipo de prueba del civil, en
// su unidad y con una sede.
func (r rejilla) calcularElCivil(tipos []Tipo, tipo string, inicio time.Time, duracion int, unidad Unidad,
	sede MunicipioResuelto,
) (time.Time, error) {
	plazo, err := Calcular(r.calendario, tipos, Peticion{
		Tipo: tipo, Inicio: inicio.Format(time.DateOnly), Duracion: strconv.Itoa(duracion),
		Unidad: string(unidad), Municipios: los(sede),
	})
	if err != nil {
		return time.Time{}, err
	}

	return time.Parse(time.DateOnly, plazo.Fecha)
}

// recorrerElCivil calcula cada entrada de la rejilla con un tipo de prueba del
// civil, en su unidad y con una sede, y da sus contraejemplos: las entradas que
// no dan fecha y las que incumple, que dice por qué una fecha incumple sus
// propiedades o nada si las cumple.
func (r rejilla) recorrerElCivil(tipos []Tipo, tipo string, unidad Unidad, sede MunicipioResuelto,
	incumple func(inicio time.Time, duracion int, fecha time.Time, sede MunicipioResuelto) string,
) []string {
	var contraejemplos []string

	for inicio := primerInicioDeLaRejilla; !inicio.After(ultimoInicioDeLaRejilla); inicio = inicio.AddDate(0, 0, 1) {
		for duracion := 1; duracion <= duracionMaximaRejilla; duracion++ {
			var motivo string

			if fecha, err := r.calcularElCivil(tipos, tipo, inicio, duracion, unidad, sede); err != nil {
				motivo = err.Error()
			} else {
				motivo = incumple(inicio, duracion, fecha, sede)
			}

			if motivo != "" {
				contraejemplos = append(contraejemplos, fmt.Sprintf("%s y %d %s: %s", inicio.Format(time.DateOnly),
					duracion, unidad, motivo))
			}
		}
	}

	return contraejemplos
}

// fechaInhabilEnElCivil dice por qué una fecha incumple la propiedad (4) —es
// inhábil para la sede o cae en un periodo inhabil—, o nada si la cumple.
func (r rejilla) fechaInhabilEnElCivil(fecha time.Time, sede MunicipioResuelto) string {
	switch {
	case r.inhabil(fecha, sede):
		return "(4) acaba en " + fecha.Format(time.DateOnly) + ", inhábil para la sede"
	case enUnPeriodoDelCivil(fecha):
		return "(4) acaba en " + fecha.Format(time.DateOnly) + ", en un periodo inhabil"
	default:
		return ""
	}
}

// incumpleLosHabilesDelCivil dice qué propiedad de los días hábiles del civil
// incumple una fecha, (4) o (5), o nada si cumple las dos.
func (r rejilla) incumpleLosHabilesDelCivil(inicio time.Time, habiles int, fecha time.Time,
	sede MunicipioResuelto,
) string {
	if motivo := r.fechaInhabilEnElCivil(fecha, sede); motivo != "" {
		return motivo
	}

	contados := 0

	for dia := inicio.AddDate(0, 0, 1); !dia.After(fecha); dia = dia.AddDate(0, 0, 1) {
		if !r.inhabilEnElCivil(dia, sede) {
			contados++
		}
	}

	if contados != habiles {
		return fmt.Sprintf("(5) hasta %s hay %d días que no son inhábiles de base ni de un periodo",
			fecha.Format(time.DateOnly), contados)
	}

	return ""
}

// incumpleLosMesesDelCivil dice qué propiedad de los meses del civil incumple
// una fecha —(6), que no es la de fecha a fecha (el mismo día del mes de
// vencimiento, o el último si no lo tiene) ni, si esa es inhábil, el primer
// hábil siguiente; o (4)—, o nada si cumple las dos.
func (r rejilla) incumpleLosMesesDelCivil(inicio time.Time, meses int, fecha time.Time, sede MunicipioResuelto) string {
	primeroDelMes := time.Date(inicio.Year(), inicio.Month()+time.Month(meses), 1, 0, 0, 0, 0, time.UTC)
	ultimoDelMes := primeroDelMes.AddDate(0, 1, -1)
	deFechaAFecha := primeroDelMes.AddDate(0, 0, min(inicio.Day(), ultimoDelMes.Day())-1)

	if fecha.Before(deFechaAFecha) {
		return "(6) acaba en " + fecha.Format(time.DateOnly) + ", antes del " + deFechaAFecha.Format(time.DateOnly)
	}

	for dia := deFechaAFecha; dia.Before(fecha); dia = dia.AddDate(0, 0, 1) {
		if !r.inhabilEnElCivil(dia, sede) {
			return "(6) acaba en " + fecha.Format(time.DateOnly) + ", después del " + dia.Format(time.DateOnly) +
				", hábil"
		}
	}

	return r.fechaInhabilEnElCivil(fecha, sede)
}
