package plazos

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// El calendario sintético de los tests del dominio. Todo es inventado a
// propósito —las comunidades 71 y 72, sus municipios, los años 2031 y 2032, las
// disposiciones—, para que ningún test pase porque el dominio tenga escrito un
// dato real: los ficheros reales los comprueban los tests de derivación y los
// del applet.
//
// La comunidad 71 tiene fichero local y la 72 no. El 1 de enero de 2031 es
// fiesta nacional en las dos; el 19 de marzo, nacional en la 71 y autonómica en
// la 72; el 2 de mayo, autonómica en la 71; el 24 de junio, autonómica en la 72
// y en una parte de ella, la Isla Sintética, a la vez, y el 25 de junio solo en
// esa parte. Los locales: el 15 de mayo en 71001 y 71002, cada uno de su
// disposición, y el 8 de septiembre de 2032 en 71001. Las faltas: 71003 no
// comunicado y 71004 con la línea sin leer en 2031, 71002 sin línea en 2032, y
// una línea, Aldea Sintética, sin municipio.
const (
	nacionalesSinteticos = `# generado a mano para los tests del dominio
fuente: prueba.fiestas
anios:
  - {anio: 2031, disposicion: PRUEBA-2030-1}
  - {anio: 2032, disposicion: PRUEBA-2031-1}
festivos:
  - {fecha: "2031-01-01", ambito: nacional, comunidades: ["71", "72"], source: PRUEBA-2030-1}
  - {fecha: "2031-03-19", ambito: nacional, comunidades: ["71"], source: PRUEBA-2030-1}
  - {fecha: "2031-03-19", ambito: autonomico, comunidades: ["72"], source: PRUEBA-2030-1}
  - {fecha: "2031-05-02", ambito: autonomico, comunidades: ["71"], source: PRUEBA-2030-1}
  - {fecha: "2031-06-24", ambito: autonomico, comunidades: ["72"], source: PRUEBA-2030-1}
  - {fecha: "2031-06-24", ambito: autonomico, comunidades: ["72"], parte: Isla Sintética, source: PRUEBA-2030-1}
  - {fecha: "2031-06-25", ambito: autonomico, comunidades: ["72"], parte: Isla Sintética, source: PRUEBA-2030-1}
  - {fecha: "2032-01-01", ambito: nacional, comunidades: ["71", "72"], source: PRUEBA-2031-1}
`
	localesSinteticos = `# generado a mano para los tests del dominio
fuente: prueba.fiestas-locales
comunidad: "71"
anios:
  - {anio: 2031, disposiciones: [PRUEBA-L-1, PRUEBA-L-2]}
  - {anio: 2032, disposiciones: [PRUEBA-L-3]}
festivos:
  - {fecha: "2031-05-15", municipio: "71001", source: PRUEBA-L-1}
  - {fecha: "2031-05-15", municipio: "71002", source: PRUEBA-L-2}
  - {fecha: "2032-09-08", municipio: "71001", source: PRUEBA-L-3}
faltas:
  - {anio: 2031, municipio: "71003", motivo: no-comunicado, source: PRUEBA-L-1}
  - {anio: 2031, municipio: "71004", motivo: linea-sin-leer, source: PRUEBA-L-2}
  - {anio: 2031, linea: Aldea Sintética, motivo: sin-municipio, source: PRUEBA-L-1}
  - {anio: 2032, municipio: "71002", motivo: sin-linea, source: PRUEBA-L-3}
`
)

// Los municipios sintéticos, ya resueltos como los recibe el dominio.
var (
	conFestivos    = Municipio{CodigoINE: "71001", Nombre: "Villa Primera", Comunidad: "71"}
	sinLineaEn2032 = Municipio{CodigoINE: "71002", Nombre: "Villa Segunda", Comunidad: "71"}
	noComunicado   = Municipio{CodigoINE: "71003", Nombre: "Villa Tercera", Comunidad: "71"}
	lineaSinLeer   = Municipio{CodigoINE: "71004", Nombre: "Villa Cuarta", Comunidad: "71"}
	sinLocales     = Municipio{CodigoINE: "72001", Nombre: "Puerto Sintético", Comunidad: "72"}
	deOtraParte    = Municipio{CodigoINE: "99001", Nombre: "Lugar Ajeno", Comunidad: "99"}
)

// dia es una fecha AAAA-MM-DD de un test.
func dia(t *testing.T, fecha string) time.Time {
	t.Helper()

	analizada, err := time.Parse(time.DateOnly, fecha)
	require.NoError(t, err)

	return analizada
}

// calendarioSintetico carga el calendario sintético.
func calendarioSintetico(t *testing.T) *Calendario {
	t.Helper()

	calendario, err := CargarCalendario([]byte(nacionalesSinteticos), map[string][]byte{"71": []byte(localesSinteticos)})
	require.NoError(t, err)

	return calendario
}

// TestCargarCalendario carga el calendario sintético y comprueba cada consulta,
// desde varias goroutines a la vez, y que lo que no se puede leer es un error
// sin clase de usuario (FR-003, FR-011, FR-020, FR-022, FR-025, FR-026).
func TestCargarCalendario(t *testing.T) {
	t.Parallel()

	calendario := calendarioSintetico(t)

	t.Run("inhábil", func(t *testing.T) {
		t.Parallel()
		comprobarInhabiles(t, calendario)
	})
	t.Run("festivos de un día", func(t *testing.T) {
		t.Parallel()
		comprobarFestivosDeUnDia(t, calendario)
	})
	t.Run("partes de un día", func(t *testing.T) {
		t.Parallel()
		comprobarPartes(t, calendario)
	})
	t.Run("cobertura", func(t *testing.T) {
		t.Parallel()
		comprobarCobertura(t, calendario)
	})
	t.Run("alcance", func(t *testing.T) {
		t.Parallel()
		comprobarAlcance(t, calendario)
	})
	t.Run("vocabularios", func(t *testing.T) {
		t.Parallel()
		comprobarVocabularios(t)
	})
	t.Run("lo que no se puede leer", func(t *testing.T) {
		t.Parallel()
		comprobarLoQueNoSeLee(t)
	})
}

// comprobarInhabiles comprueba Inhabil: sábado, domingo, nacional o autonómico
// sin parte de la comunidad de alguno de los municipios, o local de alguno.
func comprobarInhabiles(t *testing.T, calendario *Calendario) {
	t.Helper()

	casos := []struct {
		nombre     string
		fecha      string
		municipios []Municipio
		inhabil    bool
	}{
		{"sábado", "2031-01-04", nil, true},
		{"domingo", "2031-01-05", []Municipio{conFestivos}, true},
		{"martes sin fiesta", "2031-01-07", []Municipio{conFestivos, sinLocales}, false},
		{"nacional en la 71", "2031-01-01", []Municipio{conFestivos}, true},
		{"nacional en la 72", "2031-01-01", []Municipio{sinLocales}, true},
		{"nacional en una y autonómico en otra", "2031-03-19", []Municipio{sinLocales}, true},
		{"autonómico de la 71", "2031-05-02", []Municipio{conFestivos}, true},
		{"autonómico de la 71 en la 72", "2031-05-02", []Municipio{sinLocales}, false},
		{"autonómico de la comunidad de uno de los dos", "2031-05-02", []Municipio{sinLocales, noComunicado}, true},
		{"de una parte y de toda la comunidad", "2031-06-24", []Municipio{sinLocales}, true},
		{"solo de una parte", "2031-06-25", []Municipio{sinLocales}, false},
		{"local del municipio", "2031-05-15", []Municipio{conFestivos}, true},
		{"local de otro municipio", "2031-05-15", []Municipio{noComunicado}, false},
		{"local de uno de los dos", "2031-05-15", []Municipio{noComunicado, sinLineaEn2032}, true},
		{"local del segundo año", "2032-09-08", []Municipio{conFestivos}, true},
		{"comunidad sin festivos cargados", "2031-01-01", []Municipio{deOtraParte}, false},
		{"año sin cargar", "2033-01-03", []Municipio{conFestivos}, false},
	}

	for _, caso := range casos {
		assert.Equal(t, caso.inhabil, calendario.Inhabil(dia(t, caso.fecha), caso.municipios), caso.nombre)
	}
}

// comprobarFestivosDeUnDia comprueba FestivosDe: los que hacen inhábil el día
// para los municipios, cada uno una vez, con su ámbito, su municipio si es
// local y su source, nacionales antes que autonómicos y locales por municipio.
func comprobarFestivosDeUnDia(t *testing.T, calendario *Calendario) {
	t.Helper()

	nacional2031 := Festivo{Ambito: AmbitoNacional, Source: "PRUEBA-2030-1"}

	casos := []struct {
		nombre     string
		fecha      string
		municipios []Municipio
		festivos   []Festivo
	}{
		{"uno para las dos comunidades", "2031-01-01", []Municipio{conFestivos, sinLocales}, []Festivo{nacional2031}},
		{
			"nacional en una y autonómico en otra", "2031-03-19",
			[]Municipio{sinLocales, conFestivos},
			[]Festivo{nacional2031, {Ambito: AmbitoAutonomico, Source: "PRUEBA-2030-1"}},
		},
		{"solo el de su comunidad", "2031-03-19", []Municipio{sinLocales}, []Festivo{{Ambito: AmbitoAutonomico, Source: "PRUEBA-2030-1"}}},
		{
			"locales por municipio", "2031-05-15",
			[]Municipio{sinLineaEn2032, conFestivos, conFestivos},
			[]Festivo{
				{Ambito: AmbitoLocal, Municipio: "71001", Source: "PRUEBA-L-1"},
				{Ambito: AmbitoLocal, Municipio: "71002", Source: "PRUEBA-L-2"},
			},
		},
		{"el de toda la comunidad y no el de la parte", "2031-06-24", []Municipio{sinLocales}, []Festivo{{Ambito: AmbitoAutonomico, Source: "PRUEBA-2030-1"}}},
		{"ninguno por una parte", "2031-06-25", []Municipio{sinLocales}, nil},
		{"ninguno en sábado", "2031-01-04", []Municipio{conFestivos}, nil},
	}

	for _, caso := range casos {
		assert.Equal(t, caso.festivos, calendario.FestivosDe(dia(t, caso.fecha), caso.municipios), caso.nombre)
	}
}

// comprobarPartes comprueba PartesDe: los festivos de una parte de la
// comunidad, que no hacen inhábil el día (FR-011, FR-025).
func comprobarPartes(t *testing.T, calendario *Calendario) {
	t.Helper()

	isla := []FestivoDeUnaParte{{Ambito: AmbitoAutonomico, Parte: "Isla Sintética", Source: "PRUEBA-2030-1"}}

	assert.Equal(t, isla, calendario.PartesDe(dia(t, "2031-06-24"), "72"))
	assert.Equal(t, isla, calendario.PartesDe(dia(t, "2031-06-25"), "72"))
	assert.Empty(t, calendario.PartesDe(dia(t, "2031-06-25"), "71"), "la parte es de otra comunidad")
	assert.Empty(t, calendario.PartesDe(dia(t, "2031-05-02"), "71"), "un festivo de toda la comunidad no es de una parte")
	assert.Empty(t, calendario.PartesDe(dia(t, "2033-06-24"), "72"), "año sin cargar")
}

// comprobarCobertura comprueba Cobertura: por ámbito y por tramos de años, un
// tramo por año cargado, con su disposición o su falta, y uno por hueco; los
// municipios una vez cada uno, en el orden dado (FR-022, FR-026).
func comprobarCobertura(t *testing.T, calendario *Calendario) {
	t.Helper()

	nacional := []Tramo{
		{Ambito: CoberturaNacionalYAutonomica, Desde: 2029, Hasta: 2030, Falta: FaltaSinCargar},
		{Ambito: CoberturaNacionalYAutonomica, Desde: 2031, Hasta: 2031, Disposicion: "PRUEBA-2030-1"},
		{Ambito: CoberturaNacionalYAutonomica, Desde: 2032, Hasta: 2032, Disposicion: "PRUEBA-2031-1"},
		{Ambito: CoberturaNacionalYAutonomica, Desde: 2033, Hasta: 2034, Falta: FaltaSinCargar},
	}

	local := func(municipio string, en2031, en2032 Tramo) []Tramo {
		en2031.Ambito, en2031.Municipio, en2031.Desde, en2031.Hasta = CoberturaLocal, municipio, 2031, 2031
		en2032.Ambito, en2032.Municipio, en2032.Desde, en2032.Hasta = CoberturaLocal, municipio, 2032, 2032

		return []Tramo{
			{Ambito: CoberturaLocal, Municipio: municipio, Desde: 2029, Hasta: 2030, Falta: FaltaSinCargar},
			en2031, en2032,
			{Ambito: CoberturaLocal, Municipio: municipio, Desde: 2033, Hasta: 2034, Falta: FaltaSinCargar},
		}
	}

	hasta2031, hasta2032 := Tramo{Disposicion: "PRUEBA-L-2"}, Tramo{Disposicion: "PRUEBA-L-3"}

	var esperada []Tramo
	esperada = append(esperada, nacional...)
	esperada = append(esperada, local("71003", Tramo{Falta: FaltaNoComunicado}, hasta2032)...)
	esperada = append(esperada, local("71001", hasta2031, hasta2032)...)
	esperada = append(esperada, local("71004", Tramo{Falta: FaltaLineaSinLeer}, hasta2032)...)
	esperada = append(esperada, local("71002", hasta2031, Tramo{Falta: FaltaSinLinea})...)
	esperada = append(esperada, Tramo{Ambito: CoberturaLocal, Municipio: "72001", Desde: 2029, Hasta: 2034, Falta: FaltaSinFestivosLocales})

	municipios := []Municipio{noComunicado, conFestivos, lineaSinLeer, noComunicado, sinLineaEn2032, sinLocales}
	assert.Equal(t, esperada, calendario.Cobertura(2029, 2034, municipios))

	assert.Equal(t, []Tramo{
		{Ambito: CoberturaNacionalYAutonomica, Desde: 2032, Hasta: 2032, Disposicion: "PRUEBA-2031-1"},
		{Ambito: CoberturaLocal, Municipio: "72001", Desde: 2032, Hasta: 2032, Falta: FaltaSinFestivosLocales},
	}, calendario.Cobertura(2032, 2032, []Municipio{sinLocales}), "un solo año")

	assert.Empty(t, calendario.Cobertura(2032, 2031, municipios), "sin años no hay tramos")
}

// comprobarAlcance comprueba hasta dónde llegan los festivos cargados de cada
// ámbito: el primer y el último año y la última disposición del último.
func comprobarAlcance(t *testing.T, calendario *Calendario) {
	t.Helper()

	nacional, cargado := calendario.AlcanceNacional()
	assert.True(t, cargado)
	assert.Equal(t, Alcance{Desde: 2031, Hasta: 2032, Disposicion: "PRUEBA-2031-1"}, nacional)

	local, cargado := calendario.AlcanceLocal("71")
	assert.True(t, cargado)
	assert.Equal(t, Alcance{Desde: 2031, Hasta: 2032, Disposicion: "PRUEBA-L-3"}, local)

	local, cargado = calendario.AlcanceLocal("72")
	assert.False(t, cargado, "la comunidad 72 no tiene fichero local")
	assert.Equal(t, Alcance{}, local)

	vacio, err := CargarCalendario([]byte("fuente: prueba.fiestas\nanios: []\nfestivos: []\n"), nil)
	require.NoError(t, err)

	nacional, cargado = vacio.AlcanceNacional()
	assert.False(t, cargado, "sin años cargados")
	assert.Equal(t, Alcance{}, nacional)
}

// comprobarVocabularios comprueba los vocabularios cerrados de ámbitos y de
// faltas, cada uno en una lista nueva en cada llamada.
func comprobarVocabularios(t *testing.T) {
	t.Helper()

	assert.Equal(t, []Ambito{"nacional", "autonomico", "local"}, Ambitos())
	assert.Equal(t, []AmbitoDeCobertura{"nacional-y-autonomico", "local"}, AmbitosDeCobertura())
	assert.Equal(t, []Falta{"sin-cargar", "sin-festivos-locales", "no-comunicado", "linea-sin-leer", "sin-linea"},
		FaltasDeCobertura())
	assert.Equal(t, []Falta{"no-comunicado", "linea-sin-leer", "sin-linea", "sin-municipio"}, MotivosDeFalta())

	ambitos := Ambitos()
	ambitos[0] = "otro"
	assert.Equal(t, AmbitoNacional, Ambitos()[0], "cada llamada da una lista nueva")
}

// comprobarLoQueNoSeLee comprueba que un fichero que no se puede leer no da
// calendario y da un error que nombra el fichero y no lleva clase de usuario:
// el kernel lo da como inesperado (research D25).
func comprobarLoQueNoSeLee(t *testing.T) {
	t.Helper()

	const rutaNacional = "data/festivos/nacionales-y-autonomicos.yaml: "

	const rutaLocal = "data/festivos/locales/71.yaml: "

	casos := []struct {
		nombre     string
		nacionales string
		locales    string
		ruta       string
		motivo     string
	}{
		{"nacionales vacío", "", localesSinteticos, rutaNacional, "está vacío"},
		{"nacionales que no son YAML", "festivos: [", localesSinteticos, rutaNacional, "no se puede leer"},
		{"clave que no declara", nacionalesSinteticos + "notas: ninguna\n", localesSinteticos, rutaNacional, "no se puede leer"},
		{
			"fecha que no es AAAA-MM-DD",
			strings.Replace(nacionalesSinteticos, `"2031-05-02"`, `"2031-02-30"`, 1), localesSinteticos,
			rutaNacional, `la fecha "2031-02-30" no es una fecha AAAA-MM-DD`,
		},
		{
			"ámbito fuera del vocabulario",
			strings.Replace(nacionalesSinteticos, "ambito: autonomico", "ambito: regional", 1), localesSinteticos,
			rutaNacional, `el ámbito "regional" no es nacional ni autonomico`,
		},
		{"locales vacío", nacionalesSinteticos, "\n", rutaLocal, "está vacío"},
		{
			"fecha local que no se lee",
			nacionalesSinteticos, strings.Replace(localesSinteticos, `"2032-09-08"`, `"8 de septiembre"`, 1),
			rutaLocal, `la fecha "8 de septiembre" no es una fecha AAAA-MM-DD`,
		},
		{
			"motivo de falta que no es del fichero",
			nacionalesSinteticos, strings.Replace(localesSinteticos, "motivo: sin-linea", "motivo: sin-cargar", 1),
			rutaLocal, `el motivo de falta "sin-cargar" no es no-comunicado, linea-sin-leer, sin-linea ni sin-municipio`,
		},
	}

	for _, caso := range casos {
		require.NotEqual(t, nacionalesSinteticos+localesSinteticos, caso.nacionales+caso.locales, caso.nombre)

		calendario, err := CargarCalendario([]byte(caso.nacionales), map[string][]byte{"71": []byte(caso.locales)})
		require.Error(t, err, caso.nombre)
		assert.Nil(t, calendario, caso.nombre)
		assert.Contains(t, err.Error(), caso.ruta+caso.motivo, caso.nombre)

		var conClase schema.ConClase
		assert.NotErrorAs(t, err, &conClase, caso.nombre)
	}
}
