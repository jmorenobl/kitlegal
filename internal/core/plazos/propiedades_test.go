package plazos

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		Tipo: TipoGenerico, Inicio: inicio.Format(time.DateOnly), Duracion: strconv.Itoa(duracion),
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
// es inhábil.
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
