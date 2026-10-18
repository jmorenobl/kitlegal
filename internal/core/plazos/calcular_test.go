package plazos

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// El calendario sintético de los cálculos. Como el de TestCargarCalendario, es
// inventado a propósito —las comunidades 71 y 72, sus municipios, los años 2031
// y 2032, las disposiciones—, para que ningún test pase porque el dominio
// tenga escrito un dato real.
//
// Los nacionales y autonómicos llegan a 2031 y 2032; los locales de la
// comunidad 71, solo a 2031, y la 72 no tiene fichero local. En 2031: el 1 y el
// 6 de enero, el 15 de agosto, el 13 de octubre (lunes), el 8 de diciembre
// (lunes) y el 25 de diciembre (jueves), nacionales; el 19 de marzo
// (miércoles), autonómico de la 72; el 2 de mayo (viernes), autonómico de la
// 71; y el 24 (martes) y el 28 (sábado) de junio, fiestas solo de la Isla
// Sintética, una parte de la 72. En 2032: el 1 de enero (jueves), el 6 y el 25
// de diciembre (sábado), nacionales. Los locales de 2031: Villa Primera
// (71001) el 15 de mayo (jueves) y el 10 de octubre (viernes), y Villa Segunda
// (71002) el 10 de noviembre (lunes). Villa Tercera no comunicó los suyos, la
// línea de Villa Cuarta no se lee y el boletín no trae la de Villa Quinta; la
// Aldea Sintética, una entidad menor, tiene su línea sin municipio.
const (
	nacionalesDeLosCalculos = `# generado a mano para los tests del cómputo
fuente: prueba.fiestas
anios:
  - {anio: 2031, disposicion: PRUEBA-N-2031}
  - {anio: 2032, disposicion: PRUEBA-N-2032}
festivos:
  - {fecha: "2031-01-01", ambito: nacional, comunidades: ["71", "72"], source: PRUEBA-N-2031}
  - {fecha: "2031-01-06", ambito: nacional, comunidades: ["71", "72"], source: PRUEBA-N-2031}
  - {fecha: "2031-03-19", ambito: autonomico, comunidades: ["72"], source: PRUEBA-N-2031}
  - {fecha: "2031-05-02", ambito: autonomico, comunidades: ["71"], source: PRUEBA-N-2031}
  - {fecha: "2031-06-24", ambito: autonomico, comunidades: ["72"], parte: Isla Sintética, source: PRUEBA-N-2031}
  - {fecha: "2031-06-28", ambito: autonomico, comunidades: ["72"], parte: Isla Sintética, source: PRUEBA-N-2031}
  - {fecha: "2031-08-15", ambito: nacional, comunidades: ["71", "72"], source: PRUEBA-N-2031}
  - {fecha: "2031-10-13", ambito: nacional, comunidades: ["71", "72"], source: PRUEBA-N-2031}
  - {fecha: "2031-12-08", ambito: nacional, comunidades: ["71", "72"], source: PRUEBA-N-2031}
  - {fecha: "2031-12-25", ambito: nacional, comunidades: ["71", "72"], source: PRUEBA-N-2031}
  - {fecha: "2032-01-01", ambito: nacional, comunidades: ["71", "72"], source: PRUEBA-N-2032}
  - {fecha: "2032-01-06", ambito: nacional, comunidades: ["71", "72"], source: PRUEBA-N-2032}
  - {fecha: "2032-12-25", ambito: nacional, comunidades: ["71", "72"], source: PRUEBA-N-2032}
`
	localesDeLosCalculos = `# generado a mano para los tests del cómputo
fuente: prueba.fiestas-locales
comunidad: "71"
anios:
  - {anio: 2031, disposiciones: [PRUEBA-L-1, PRUEBA-L-2]}
festivos:
  - {fecha: "2031-05-15", municipio: "71001", source: PRUEBA-L-1}
  - {fecha: "2031-10-10", municipio: "71001", source: PRUEBA-L-1}
  - {fecha: "2031-11-10", municipio: "71002", source: PRUEBA-L-2}
faltas:
  - {anio: 2031, municipio: "71003", motivo: no-comunicado, source: PRUEBA-L-1}
  - {anio: 2031, municipio: "71004", motivo: linea-sin-leer, source: PRUEBA-L-1}
  - {anio: 2031, municipio: "71005", motivo: sin-linea, source: PRUEBA-L-2}
  - {anio: 2031, linea: Aldea Sintética, motivo: sin-municipio, source: PRUEBA-L-1}
`
)

// Los municipios de los cálculos, ya resueltos como los recibe Calcular.
var (
	villaPrimera = resuelto("71001", "Villa Primera", "71", "Comunidad Sintética")
	villaSegunda = resuelto("71002", "Villa Segunda", "71", "Comunidad Sintética")
	villaTercera = resuelto("71003", "Villa Tercera", "71", "Comunidad Sintética")
	villaCuarta  = resuelto("71004", "Villa Cuarta", "71", "Comunidad Sintética")
	villaQuinta  = resuelto("71005", "Villa Quinta", "71", "Comunidad Sintética")
	puertoMarino = resuelto("72001", "Puerto Marino", "72", "Comunidad Marina")
)

// Los preceptos de data/plazos.yaml, como los da el plazo.
var (
	precepto122      = PreceptoDelPlazo{Norma: "BOE-A-2015-10565", Bloque: "a122", Apartado: "122.1"}
	precepto124      = PreceptoDelPlazo{Norma: "BOE-A-2015-10565", Bloque: "a124", Apartado: "124.1"}
	precepto21       = PreceptoDelPlazo{Norma: "BOE-A-2015-10565", Bloque: "a21", Apartado: "21.3"}
	precepto24LPAC   = PreceptoDelPlazo{Norma: "BOE-A-2015-10565", Bloque: "a24", Apartado: "24.4"}
	precepto30Dias   = PreceptoDelPlazo{Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.2"}
	precepto30Meses  = PreceptoDelPlazo{Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.4"}
	precepto20       = PreceptoDelPlazo{Norma: "BOE-A-2013-12887", Bloque: "a20", Apartado: "20.1"}
	precepto24LTAIBG = PreceptoDelPlazo{Norma: "BOE-A-2013-12887", Bloque: "a24", Apartado: "24.2"}
	precepto46       = PreceptoDelPlazo{Norma: "BOE-A-1998-16718", Bloque: "a46", Apartado: "46.1"}
	precepto128      = PreceptoDelPlazo{Norma: "BOE-A-1998-16718", Bloque: "a128", Apartado: "128.2"}
)

// Los textos de los avisos que dan los cálculos, escritos a mano desde
// contracts/applet-plazos.md §5.
const (
	textoLocalesDePuerto = "⚠ FALTAN FESTIVOS LOCALES: kitlegal no tiene los de Puerto Marino (72001) de 2031: no " +
		"tiene festivos locales de Comunidad Marina; la fecha se ha calculado sin ellos y puede ser anterior a la real."
	textoDeLaIsla = "⚠ FESTIVO DE UNA PARTE: el 2031-06-24 es fiesta solo en Isla Sintética (Comunidad Marina), " +
		"según PRUEBA-N-2031; kitlegal no sabe si Puerto Marino está ahí y lo ha contado como hábil. La fecha es la más " +
		"temprana posible."
	llegaElLocal = " Los de Comunidad Sintética llegan a 2031 (PRUEBA-L-2)."
)

// resuelto es un municipio como lo da territorio resolver.
func resuelto(codigo, nombre, comunidad, nombreDeLaComunidad string) MunicipioResuelto {
	return MunicipioResuelto{
		Municipio:           Municipio{CodigoINE: codigo, Nombre: nombre, Comunidad: comunidad},
		NombreDeLaComunidad: nombreDeLaComunidad,
	}
}

// los son los municipios de una petición, en el orden dado.
func los(municipios ...MunicipioResuelto) []MunicipioResuelto {
	return municipios
}

// festivoNacional, festivoAutonomico y festivoLocal son festivos saltados.
func festivoNacional(fecha, source string) FestivoSaltado {
	return FestivoSaltado{Fecha: fecha, Ambito: "nacional", Source: source}
}

func festivoAutonomico(fecha, source string) FestivoSaltado {
	return FestivoSaltado{Fecha: fecha, Ambito: "autonomico", Source: source}
}

func festivoLocal(fecha, municipio, source string) FestivoSaltado {
	return FestivoSaltado{Fecha: fecha, Ambito: "local", Municipio: municipio, Source: source}
}

// calendarioDeLosCalculos carga el calendario sintético de los cálculos.
func calendarioDeLosCalculos(t *testing.T) *Calendario {
	t.Helper()

	calendario, err := CargarCalendario([]byte(nacionalesDeLosCalculos),
		map[string][]byte{"71": []byte(localesDeLosCalculos)})
	require.NoError(t, err)

	return calendario
}

// tiposDelContrato son los siete tipos de data/plazos.yaml del contrato.
func tiposDelContrato(t *testing.T) []Tipo {
	t.Helper()

	tipos, err := LeerTipos([]byte(plazosDelContrato))
	require.NoError(t, err)

	return tipos
}

// TestCalcular calcula sobre el calendario sintético cada regla del cómputo de
// la LPAC, del silencio y del contencioso, cada aviso y cada fila de los
// errores de argumentos, con su orden (FR-002 a FR-012, FR-024, FR-026;
// contracts/applet-plazos.md §2, §4 a §6; research D7 a D10).
func TestCalcular(t *testing.T) {
	t.Parallel()

	calendario := calendarioDeLosCalculos(t)
	tipos := tiposDelContrato(t)

	t.Run("el plazo entero", func(t *testing.T) {
		t.Parallel()
		comprobarPlazoEntero(t, calendario, tipos)
	})
	t.Run("el cómputo", func(t *testing.T) {
		t.Parallel()
		comprobarComputos(t, calendario, tipos, casosDelComputo())
	})
	t.Run("el contencioso", func(t *testing.T) {
		t.Parallel()
		comprobarComputos(t, calendario, tipos, casosDelContencioso())
	})
	t.Run("los avisos", func(t *testing.T) {
		t.Parallel()
		comprobarAvisos(t, calendario, tipos)
	})
	t.Run("los errores de argumentos", func(t *testing.T) {
		t.Parallel()
		comprobarErrores(t, calendario, tipos, erroresDeArgumentos())
	})
	t.Run("el orden de los errores", func(t *testing.T) {
		t.Parallel()
		comprobarErrores(t, calendario, tipos, ordenDeLosErrores())
	})
	t.Run("el último día que se escribe AAAA-MM-DD", func(t *testing.T) {
		t.Parallel()
		comprobarElUltimoDia(t, calendario, tipos)
	})
}

// comprobarPlazoEntero calcula el caso (a) del calendario sintético —un mes que
// vence en un festivo local de viernes y se prorroga sobre el fin de semana y
// un nacional— y lo compara entero: con el mismo municipio dos veces, cuenta
// una vez (research D4).
func comprobarPlazoEntero(t *testing.T, calendario *Calendario, tipos []Tipo) {
	t.Helper()

	plazo, err := Calcular(calendario, tipos, Peticion{
		Tipo: TipoReposicion, Inicio: "2031-09-10", Municipios: los(villaPrimera, villaPrimera),
	})
	require.NoError(t, err)

	assert.Equal(t, Plazo{
		Fecha:      "2031-10-14",
		Tipo:       TipoReposicion,
		Preceptos:  []PreceptoDelPlazo{precepto124},
		Municipios: []MunicipioDelPlazo{{CodigoINE: "71001", Nombre: "Villa Primera"}},
		Festivos: []FestivoSaltado{
			festivoLocal("2031-10-10", "71001", "PRUEBA-L-1"),
			festivoNacional("2031-10-13", "PRUEBA-N-2031"),
		},
		Avisos: []Aviso{},
		Cobertura: []TramoDeCobertura{
			{Ambito: "nacional-y-autonomico", Desde: 2031, Hasta: 2031, Disposicion: "PRUEBA-N-2031"},
			{Ambito: "local", Municipio: "71001", Desde: 2031, Hasta: 2031, Disposicion: "PRUEBA-L-2"},
		},
	}, plazo)

	otraVez, err := Calcular(calendario, tipos, Peticion{
		Tipo: TipoReposicion, Inicio: "2031-09-10", Municipios: los(villaPrimera, villaPrimera),
	})
	require.NoError(t, err)
	assert.Equal(t, plazo, otraVez, "la misma petición da el mismo plazo")
}

// casoDeComputo es un plazo calculado con su fecha, sus festivos saltados, los
// códigos de sus avisos y, si no es nil, sus preceptos.
type casoDeComputo struct {
	nombre    string
	peticion  Peticion
	fecha     string
	festivos  []FestivoSaltado
	avisos    []string
	preceptos []PreceptoDelPlazo
}

// casosDelComputo son los de la LPAC y del silencio (arts. 30.2 a 30.6 LPAC;
// research D7 y D8).
func casosDelComputo() []casoDeComputo {
	octubre := []FestivoSaltado{
		festivoLocal("2031-10-10", "71001", "PRUEBA-L-1"), festivoNacional("2031-10-13", "PRUEBA-N-2031"),
	}

	return []casoDeComputo{
		{
			nombre:    "art. 30.6: el festivo local del interesado es inhábil para el cómputo",
			peticion:  Peticion{Tipo: TipoAlzada, Inicio: "2031-09-10", Municipios: los(villaPrimera, villaSegunda)},
			fecha:     "2031-10-14",
			festivos:  octubre,
			preceptos: []PreceptoDelPlazo{precepto122},
		},
		{
			nombre:   "art. 30.6: el festivo local de la sede es inhábil para el cómputo",
			peticion: Peticion{Tipo: TipoReposicion, Inicio: "2031-10-10", Municipios: los(villaPrimera, villaSegunda)},
			fecha:    "2031-11-11",
			festivos: []FestivoSaltado{festivoLocal("2031-11-10", "71002", "PRUEBA-L-2")},
		},
		{
			nombre:    "art. 30.6: el autonómico de la comunidad de la sede es inhábil",
			peticion:  Peticion{Tipo: TipoAccesoResolver, Inicio: "2031-02-19", Municipios: los(villaPrimera, puertoMarino)},
			fecha:     "2031-03-20",
			festivos:  []FestivoSaltado{festivoAutonomico("2031-03-19", "PRUEBA-N-2031")},
			avisos:    []string{AvisoFaltanLocales},
			preceptos: []PreceptoDelPlazo{precepto20},
		},
		{
			nombre:    "el autonómico de otra comunidad es hábil",
			peticion:  Peticion{Tipo: TipoAccesoReclamar, Inicio: "2031-02-19", Municipios: los(villaPrimera, villaPrimera)},
			fecha:     "2031-03-19",
			preceptos: []PreceptoDelPlazo{precepto24LTAIBG},
		},
		{
			nombre:   "art. 30.5: el último día en sábado se prorroga al lunes",
			peticion: Peticion{Tipo: TipoReposicion, Inicio: "2031-09-04", Municipios: los(villaPrimera, villaPrimera)},
			fecha:    "2031-10-06",
		},
		{
			nombre: "art. 30.4: el mes sin el día equivalente vence el último",
			peticion: Peticion{
				Tipo: TipoGenerico, Inicio: "2031-01-31", Duracion: "1", Unidad: "meses",
				Municipios: los(villaPrimera, villaPrimera),
			},
			fecha:     "2031-02-28",
			preceptos: []PreceptoDelPlazo{precepto30Meses},
		},
		{
			nombre: "art. 30.4 y 30.5: el último día del mes, en domingo, se prorroga",
			peticion: Peticion{
				Tipo: TipoGenerico, Inicio: "2031-10-31", Duracion: "1", Unidad: "meses",
				Municipios: los(villaPrimera, villaPrimera),
			},
			fecha: "2031-12-01",
		},
		{
			nombre: "art. 30.2: los días naturales cuentan los inhábiles y el último se prorroga",
			peticion: Peticion{
				Tipo: TipoGenerico, Inicio: "2031-09-25", Duracion: "15", Unidad: "dias-naturales",
				Municipios: los(villaPrimera, villaPrimera),
			},
			fecha:     "2031-10-14",
			festivos:  octubre,
			preceptos: []PreceptoDelPlazo{precepto30Dias},
		},
		{
			nombre: "art. 30.2: los días hábiles saltan los festivos del tramo",
			peticion: Peticion{
				Tipo: TipoGenerico, Inicio: "2031-10-08", Duracion: "5", Unidad: "dias-habiles",
				Municipios: los(villaPrimera, villaPrimera),
			},
			fecha:     "2031-10-17",
			festivos:  octubre,
			preceptos: []PreceptoDelPlazo{precepto30Dias},
		},
		{
			nombre: "FR-002: se cuenta desde el día siguiente al de inicio",
			peticion: Peticion{
				Tipo: TipoGenerico, Inicio: "2031-10-16", Duracion: "1", Unidad: "dias-naturales",
				Municipios: los(villaPrimera, villaPrimera),
			},
			fecha: "2031-10-17",
		},
		{
			nombre: "un festivo en sábado no es un festivo saltado",
			peticion: Peticion{
				Tipo: TipoGenerico, Inicio: "2032-12-23", Duracion: "2", Unidad: "dias-habiles",
				Municipios: los(villaPrimera, villaPrimera),
			},
			fecha:  "2032-12-27",
			avisos: []string{AvisoFaltanLocales},
		},
		{
			nombre: "el silencio en meses: el día siguiente al vencimiento prorrogado",
			peticion: Peticion{
				Tipo: TipoSilencio, Inicio: "2031-07-10", Duracion: "3", Unidad: "meses",
				Municipios: los(villaPrimera, villaPrimera),
			},
			fecha:     "2031-10-15",
			festivos:  octubre,
			preceptos: []PreceptoDelPlazo{precepto21, precepto24LPAC},
		},
		{
			nombre: "el silencio en días hábiles: el día siguiente, aunque sea sábado",
			peticion: Peticion{
				Tipo: TipoSilencio, Inicio: "2031-10-08", Duracion: "5", Unidad: "dias-habiles",
				Municipios: los(villaPrimera, villaPrimera),
			},
			fecha:     "2031-10-18",
			festivos:  octubre,
			preceptos: []PreceptoDelPlazo{precepto21, precepto24LPAC},
		},
		{
			nombre: "el silencio en días naturales",
			peticion: Peticion{
				Tipo: TipoSilencio, Inicio: "2031-09-25", Duracion: "15", Unidad: "dias-naturales",
				Municipios: los(villaPrimera, villaPrimera),
			},
			fecha:    "2031-10-15",
			festivos: octubre,
		},
	}
}

// casosDelContencioso son los del recurso contencioso-administrativo: agosto,
// el art. 182.1 LOPJ y la prórroga del 185.2, y los días del art. 183 contados
// (research D9).
func casosDelContencioso() []casoDeComputo {
	return []casoDeComputo{
		{
			nombre:    "agosto no corre: el vencimiento pasa al mismo día del mes siguiente",
			peticion:  Peticion{Tipo: TipoContencioso, Inicio: "2031-06-16", Municipios: los(villaPrimera)},
			fecha:     "2031-09-16",
			preceptos: []PreceptoDelPlazo{precepto46, precepto128},
		},
		{
			nombre:   "una notificación de agosto cuenta desde el 31 de agosto",
			peticion: Peticion{Tipo: TipoContencioso, Inicio: "2031-08-05", Municipios: los(villaPrimera)},
			fecha:    "2031-10-31",
		},
		{
			nombre:   "si el día siguiente es de agosto, también desde el 31",
			peticion: Peticion{Tipo: TipoContencioso, Inicio: "2031-07-31", Municipios: los(villaPrimera)},
			fecha:    "2031-10-31",
		},
		{
			nombre:   "art. 182.1: el 24 de diciembre es inhábil y la prórroga salta la Navidad",
			peticion: Peticion{Tipo: TipoContencioso, Inicio: "2031-10-24", Municipios: los(villaPrimera)},
			fecha:    "2031-12-26",
			festivos: []FestivoSaltado{festivoNacional("2031-12-25", "PRUEBA-N-2031")},
			avisos:   []string{AvisoDiasDelArt183},
		},
		{
			nombre:   "art. 182.1: el 31 de diciembre es inhábil y la prórroga cambia de año",
			peticion: Peticion{Tipo: TipoContencioso, Inicio: "2031-10-31", Municipios: los(villaPrimera)},
			fecha:    "2032-01-02",
			festivos: []FestivoSaltado{festivoNacional("2032-01-01", "PRUEBA-N-2032")},
			avisos:   []string{AvisoFaltanLocales, AvisoDiasDelArt183},
		},
		{
			nombre:   "art. 183: los días del 24 de diciembre al 6 de enero se cuentan",
			peticion: Peticion{Tipo: TipoContencioso, Inicio: "2031-11-20", Municipios: los(villaPrimera)},
			fecha:    "2032-01-20",
			avisos:   []string{AvisoFaltanLocales, AvisoDiasDelArt183},
		},
		{
			nombre:   "art. 185.2: el sábado, el domingo y la fiesta nacional",
			peticion: Peticion{Tipo: TipoContencioso, Inicio: "2031-10-06", Municipios: los(villaPrimera)},
			fecha:    "2031-12-09",
			festivos: []FestivoSaltado{festivoNacional("2031-12-08", "PRUEBA-N-2031")},
		},
		{
			nombre:   "art. 182.1: el festivo local de la sede",
			peticion: Peticion{Tipo: TipoContencioso, Inicio: "2031-03-15", Municipios: los(villaPrimera)},
			fecha:    "2031-05-16",
			festivos: []FestivoSaltado{festivoLocal("2031-05-15", "71001", "PRUEBA-L-1")},
		},
		{
			nombre:   "art. 182.1: el autonómico de la comunidad de la sede",
			peticion: Peticion{Tipo: TipoContencioso, Inicio: "2031-03-02", Municipios: los(villaPrimera)},
			fecha:    "2031-05-05",
			festivos: []FestivoSaltado{festivoAutonomico("2031-05-02", "PRUEBA-N-2031")},
		},
	}
}

// comprobarComputos calcula cada caso y compara su fecha, sus festivos
// saltados, los códigos de sus avisos y, si los fija, sus preceptos.
func comprobarComputos(t *testing.T, calendario *Calendario, tipos []Tipo, casos []casoDeComputo) {
	t.Helper()

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			plazo, err := Calcular(calendario, tipos, caso.peticion)
			require.NoError(t, err)

			assert.Equal(t, caso.fecha, plazo.Fecha)
			assert.Equal(t, caso.peticion.Tipo, plazo.Tipo)
			assert.Equal(t, append([]FestivoSaltado{}, caso.festivos...), plazo.Festivos)
			assert.Equal(t, append([]string{}, caso.avisos...), codigosDe(plazo.Avisos))

			if caso.preceptos != nil {
				assert.Equal(t, caso.preceptos, plazo.Preceptos)
			}
		})
	}
}

// codigosDe son los códigos de unos avisos, en su orden.
func codigosDe(avisos []Aviso) []string {
	codigos := []string{}
	for _, aviso := range avisos {
		codigos = append(codigos, aviso.Codigo)
	}

	return codigos
}

// casoDeAvisos es un plazo calculado con sus avisos enteros y, si no es nil,
// su cobertura.
type casoDeAvisos struct {
	nombre    string
	peticion  Peticion
	fecha     string
	avisos    []Aviso
	cobertura []TramoDeCobertura
}

// comprobarAvisos calcula cada caso de avisos y compara sus avisos enteros,
// con su código, su texto y su orden, y su cobertura (FR-011, FR-024, FR-026;
// contracts/applet-plazos.md §4 y §5). Ningún aviso nombra la entidad menor.
func comprobarAvisos(t *testing.T, calendario *Calendario, tipos []Tipo) {
	t.Helper()

	for _, caso := range casosDeAvisos() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			plazo, err := Calcular(calendario, tipos, caso.peticion)
			require.NoError(t, err)

			assert.Equal(t, caso.fecha, plazo.Fecha)
			assert.Equal(t, append([]Aviso{}, caso.avisos...), plazo.Avisos)

			if caso.cobertura != nil {
				assert.Equal(t, caso.cobertura, plazo.Cobertura)
			}

			for _, aviso := range plazo.Avisos {
				assert.NotContains(t, aviso.Texto, "Aldea Sintética", "ningún aviso nombra una entidad menor (FR-024)")
				assert.NotContains(t, aviso.Texto, "\n", "cada aviso es una línea")
			}
		})
	}
}

// casosDeAvisos son los de los cuatro códigos de aviso: cada uno solo si el
// tramo pasa por lo que lo causa, en el orden de los códigos y, dentro de uno,
// de fecha y de municipio.
func casosDeAvisos() []casoDeAvisos {
	casos := casosDeCobertura()

	return append(casos,
		casoDeAvisos{
			nombre: "un festivo de una parte de la comunidad cuenta como hábil y avisa",
			peticion: Peticion{
				Tipo: TipoGenerico, Inicio: "2031-06-23", Duracion: "1", Unidad: "dias-habiles",
				Municipios: los(puertoMarino, puertoMarino),
			},
			fecha: "2031-06-24",
			avisos: []Aviso{
				{Codigo: AvisoFaltanLocales, Texto: textoLocalesDePuerto},
				{Codigo: AvisoFestivoDeUnaParte, Texto: textoDeLaIsla},
			},
		},
		casoDeAvisos{
			nombre: "el aviso de una parte nombra solo el municipio de esa comunidad",
			peticion: Peticion{
				Tipo: TipoGenerico, Inicio: "2031-06-23", Duracion: "1", Unidad: "dias-habiles",
				Municipios: los(villaPrimera, puertoMarino),
			},
			fecha: "2031-06-24",
			avisos: []Aviso{
				{Codigo: AvisoFaltanLocales, Texto: textoLocalesDePuerto},
				{Codigo: AvisoFestivoDeUnaParte, Texto: textoDeLaIsla},
			},
		},
		casoDeAvisos{
			nombre: "un festivo de una parte de otra comunidad no avisa",
			peticion: Peticion{
				Tipo: TipoGenerico, Inicio: "2031-06-23", Duracion: "1", Unidad: "dias-habiles",
				Municipios: los(villaPrimera, villaPrimera),
			},
			fecha: "2031-06-24",
		},
		casoDeAvisos{
			nombre: "un festivo de una parte en sábado no se ha contado como hábil y no avisa",
			peticion: Peticion{
				Tipo: TipoGenerico, Inicio: "2031-06-26", Duracion: "2", Unidad: "dias-habiles",
				Municipios: los(puertoMarino, puertoMarino),
			},
			fecha:  "2031-06-30",
			avisos: []Aviso{{Codigo: AvisoFaltanLocales, Texto: textoLocalesDePuerto}},
		},
		casoDeAvisos{
			nombre:   "el art. 183 solo avisa en el contencioso",
			peticion: Peticion{Tipo: TipoContencioso, Inicio: "2031-10-24", Municipios: los(villaPrimera)},
			fecha:    "2031-12-26",
			avisos: []Aviso{{Codigo: AvisoDiasDelArt183, Texto: "⚠ DÍAS DEL ART. 183 LOPJ: el plazo pasa por días " +
				"del 24 de diciembre al 6 de enero, que el art. 183 LOPJ declara inhábiles para las actuaciones " +
				"judiciales; se han contado, salvo los inhábiles del art. 182.1 LOPJ, porque el art. 128.2 LJCA no los " +
				"excluye del plazo para interponer el recurso. La fecha es la más temprana posible."}},
		},
	)
}

// casosDeCobertura son los de los avisos de los festivos que faltan: los
// años antes y después de los cargados y cada motivo de falta.
func casosDeCobertura() []casoDeAvisos {
	return []casoDeAvisos{
		{
			nombre: "los años anteriores a los cargados, de los dos ámbitos",
			peticion: Peticion{
				Tipo: TipoGenerico, Inicio: "2030-12-30", Duracion: "3", Unidad: "dias-habiles",
				Municipios: los(villaPrimera, villaPrimera),
			},
			fecha: "2031-01-03",
			avisos: []Aviso{
				{Codigo: AvisoFaltanNacionalesYAutonomicos, Texto: "⚠ FALTAN FESTIVOS NACIONALES Y AUTONÓMICOS: " +
					"kitlegal no tiene los anteriores a 2031; la fecha se ha calculado sin ellos y puede ser anterior a " +
					"la real. Los que tiene llegan a 2032 (PRUEBA-N-2032)."},
				{Codigo: AvisoFaltanLocales, Texto: "⚠ FALTAN FESTIVOS LOCALES: kitlegal no tiene los de Villa " +
					"Primera (71001) anteriores a 2031; la fecha se ha calculado sin ellos y puede ser anterior a la " +
					"real." + llegaElLocal},
			},
			cobertura: []TramoDeCobertura{
				{Ambito: "nacional-y-autonomico", Desde: 2030, Hasta: 2030, Falta: "sin-cargar"},
				{Ambito: "nacional-y-autonomico", Desde: 2031, Hasta: 2031, Disposicion: "PRUEBA-N-2031"},
				{Ambito: "local", Municipio: "71001", Desde: 2030, Hasta: 2030, Falta: "sin-cargar"},
				{Ambito: "local", Municipio: "71001", Desde: 2031, Hasta: 2031, Disposicion: "PRUEBA-L-2"},
			},
		},
		{
			nombre: "un tramo de años después de los cargados",
			peticion: Peticion{
				Tipo: TipoGenerico, Inicio: "2032-12-20", Duracion: "14", Unidad: "meses",
				Municipios: los(villaPrimera, villaPrimera),
			},
			fecha: "2034-02-20",
			avisos: []Aviso{
				{Codigo: AvisoFaltanNacionalesYAutonomicos, Texto: "⚠ FALTAN FESTIVOS NACIONALES Y AUTONÓMICOS: " +
					"kitlegal no tiene los de 2033 a 2034; la fecha se ha calculado sin ellos y puede ser anterior a la " +
					"real. Los que tiene llegan a 2032 (PRUEBA-N-2032)."},
				{Codigo: AvisoFaltanLocales, Texto: "⚠ FALTAN FESTIVOS LOCALES: kitlegal no tiene los de Villa " +
					"Primera (71001) de 2032 a 2034; la fecha se ha calculado sin ellos y puede ser anterior a la real." +
					llegaElLocal},
			},
			cobertura: []TramoDeCobertura{
				{Ambito: "nacional-y-autonomico", Desde: 2032, Hasta: 2032, Disposicion: "PRUEBA-N-2032"},
				{Ambito: "nacional-y-autonomico", Desde: 2033, Hasta: 2034, Falta: "sin-cargar"},
				{Ambito: "local", Municipio: "71001", Desde: 2032, Hasta: 2034, Falta: "sin-cargar"},
			},
		},
		{
			nombre:   "un municipio de una comunidad sin festivos locales",
			peticion: Peticion{Tipo: TipoReposicion, Inicio: "2031-09-10", Municipios: los(puertoMarino, puertoMarino)},
			fecha:    "2031-10-10",
			avisos:   []Aviso{{Codigo: AvisoFaltanLocales, Texto: textoLocalesDePuerto}},
			cobertura: []TramoDeCobertura{
				{Ambito: "nacional-y-autonomico", Desde: 2031, Hasta: 2031, Disposicion: "PRUEBA-N-2031"},
				{Ambito: "local", Municipio: "72001", Desde: 2031, Hasta: 2031, Falta: "sin-festivos-locales"},
			},
		},
		{
			nombre:   "la línea que no se lee y la que no está, una por municipio",
			peticion: Peticion{Tipo: TipoReposicion, Inicio: "2031-09-10", Municipios: los(villaCuarta, villaQuinta)},
			fecha:    "2031-10-10",
			avisos: []Aviso{
				{Codigo: AvisoFaltanLocales, Texto: "⚠ FALTAN FESTIVOS LOCALES: kitlegal no tiene los de Villa " +
					"Cuarta (71004) de 2031: su línea del boletín no se lee sin interpretarla; la fecha se ha calculado " +
					"sin ellos y puede ser anterior a la real." + llegaElLocal},
				{Codigo: AvisoFaltanLocales, Texto: "⚠ FALTAN FESTIVOS LOCALES: kitlegal no tiene los de Villa " +
					"Quinta (71005) de 2031: el boletín no trae su línea; la fecha se ha calculado sin ellos y puede ser " +
					"anterior a la real." + llegaElLocal},
			},
			cobertura: []TramoDeCobertura{
				{Ambito: "nacional-y-autonomico", Desde: 2031, Hasta: 2031, Disposicion: "PRUEBA-N-2031"},
				{Ambito: "local", Municipio: "71004", Desde: 2031, Hasta: 2031, Falta: "linea-sin-leer"},
				{Ambito: "local", Municipio: "71005", Desde: 2031, Hasta: 2031, Falta: "sin-linea"},
			},
		},
		{
			nombre:   "no comunicado y sin cargar, en un aviso; los locales por fecha y después por municipio",
			peticion: Peticion{Tipo: TipoReposicion, Inicio: "2031-12-15", Municipios: los(villaPrimera, villaTercera)},
			fecha:    "2032-01-15",
			avisos: []Aviso{
				{Codigo: AvisoFaltanLocales, Texto: "⚠ FALTAN FESTIVOS LOCALES: kitlegal no tiene los de Villa " +
					"Tercera (71003) de 2031: el boletín los da como no comunicados por el ayuntamiento ni de 2032; la " +
					"fecha se ha calculado sin ellos y puede ser anterior a la real." + llegaElLocal},
				{Codigo: AvisoFaltanLocales, Texto: "⚠ FALTAN FESTIVOS LOCALES: kitlegal no tiene los de Villa " +
					"Primera (71001) de 2032; la fecha se ha calculado sin ellos y puede ser anterior a la real." +
					llegaElLocal},
			},
			cobertura: []TramoDeCobertura{
				{Ambito: "nacional-y-autonomico", Desde: 2031, Hasta: 2031, Disposicion: "PRUEBA-N-2031"},
				{Ambito: "nacional-y-autonomico", Desde: 2032, Hasta: 2032, Disposicion: "PRUEBA-N-2032"},
				{Ambito: "local", Municipio: "71001", Desde: 2031, Hasta: 2031, Disposicion: "PRUEBA-L-2"},
				{Ambito: "local", Municipio: "71001", Desde: 2032, Hasta: 2032, Falta: "sin-cargar"},
				{Ambito: "local", Municipio: "71003", Desde: 2031, Hasta: 2031, Falta: "no-comunicado"},
				{Ambito: "local", Municipio: "71003", Desde: 2032, Hasta: 2032, Falta: "sin-cargar"},
			},
		},
	}
}

// casoDeError es una petición que no vale: Calcular y ComprobarPeticion la
// rechazan con el mismo error de argumentos, cuyo mensaje nombra lo que dice
// nombra y no lo que dice calla.
type casoDeError struct {
	nombre   string
	peticion Peticion
	nombra   []string
	calla    []string
}

// erroresDeArgumentos es una fila por cada defecto de contracts/applet-plazos.md
// §6 que decide el dominio (FR-012; research D10).
func erroresDeArgumentos() []casoDeError {
	dos := los(villaPrimera, villaSegunda)

	return []casoDeError{
		{nombre: "el tipo no es de los siete", peticion: Peticion{
			Tipo: "inventado", Inicio: "2031-09-10",
			Municipios: dos,
		}, nombra: []string{`"inventado"`, "alzada", "generico"}},
		{nombre: "la fecha no es AAAA-MM-DD", peticion: Peticion{
			Tipo: TipoAlzada, Inicio: "10/09/2031",
			Municipios: dos,
		}, nombra: []string{`"10/09/2031"`, "AAAA-MM-DD"}},
		{nombre: "la fecha sin ceros", peticion: Peticion{
			Tipo: TipoAlzada, Inicio: "2031-9-10",
			Municipios: dos,
		}, nombra: []string{`"2031-9-10"`, "AAAA-MM-DD"}},
		{nombre: "la fecha con barras", peticion: Peticion{
			Tipo: TipoAlzada, Inicio: "2031/09/10",
			Municipios: dos,
		}, nombra: []string{`"2031/09/10"`, "AAAA-MM-DD"}},
		{nombre: "la fecha no es un día del calendario", peticion: Peticion{
			Tipo: TipoAlzada, Inicio: "2031-02-29",
			Municipios: dos,
		}, nombra: []string{`"2031-02-29"`, "día del calendario"}},
		{nombre: "falta la duración", peticion: Peticion{
			Tipo: TipoGenerico, Inicio: "2031-09-10",
			Unidad: "dias-habiles", Municipios: dos,
		}, nombra: []string{`"generico"`, "duración"}},
		{nombre: "la duración no es entera", peticion: Peticion{
			Tipo: TipoGenerico, Inicio: "2031-09-10",
			Duracion: "1.5", Unidad: "dias-habiles", Municipios: dos,
		}, nombra: []string{`"1.5"`, "entero positivo"}},
		{nombre: "la duración es cero", peticion: Peticion{
			Tipo: TipoSilencio, Inicio: "2031-09-10",
			Duracion: "0", Unidad: "meses", Municipios: dos,
		}, nombra: []string{`"0"`, "entero positivo"}},
		{nombre: "la duración es negativa", peticion: Peticion{
			Tipo: TipoGenerico, Inicio: "2031-09-10",
			Duracion: "-3", Unidad: "meses", Municipios: dos,
		}, nombra: []string{`"-3"`, "entero positivo"}},
		{
			nombre: "la duración no cabe en un entero", peticion: Peticion{
				Tipo: TipoGenerico, Inicio: "2031-09-10",
				Duracion: "99999999999999999999", Unidad: "meses", Municipios: dos,
			},
			nombra: []string{`"99999999999999999999"`, "entero positivo"},
		},
		{nombre: "la duración a un tipo que la fija", peticion: Peticion{
			Tipo: TipoReposicion, Inicio: "2031-09-10",
			Duracion: "2", Municipios: dos,
		}, nombra: []string{`"reposicion"`, `"2"`}},
		{nombre: "falta la unidad", peticion: Peticion{
			Tipo: TipoSilencio, Inicio: "2031-09-10",
			Duracion: "3", Municipios: dos,
		}, nombra: []string{`"3"`, "unidad", "dias-habiles"}},
		{nombre: "la unidad no es del vocabulario", peticion: Peticion{
			Tipo: TipoGenerico, Inicio: "2031-09-10",
			Duracion: "3", Unidad: "semanas", Municipios: dos,
		}, nombra: []string{`"semanas"`, "dias-naturales"}},
		{nombre: "la unidad a un tipo que la fija", peticion: Peticion{
			Tipo: TipoReposicion, Inicio: "2031-09-10",
			Unidad: "meses", Municipios: dos,
		}, nombra: []string{`"reposicion"`, `"meses"`}},
		{nombre: "un municipio en un plazo de la LPAC", peticion: Peticion{
			Tipo:   TipoReposicion,
			Inicio: "2031-09-10", Municipios: los(villaPrimera),
		}, nombra: []string{`"reposicion"`, "dos", "1"}},
		{
			nombre: "tres municipios en un plazo de la LPAC", peticion: Peticion{
				Tipo:   TipoAlzada,
				Inicio: "2031-09-10", Municipios: los(villaPrimera, villaSegunda, villaTercera),
			},
			nombra: []string{`"alzada"`, "dos", "3"},
		},
		{nombre: "ningún municipio en un plazo de la LPAC", peticion: Peticion{
			Tipo:   TipoGenerico,
			Inicio: "2031-09-10", Duracion: "3", Unidad: "meses",
		}, nombra: []string{`"generico"`, "dos", "0"}},
		{nombre: "dos municipios en el contencioso", peticion: Peticion{
			Tipo:   TipoContencioso,
			Inicio: "2031-09-10", Municipios: dos,
		}, nombra: []string{`"contencioso"`, "un municipio", "2"}},
		{nombre: "ningún municipio en el contencioso", peticion: Peticion{
			Tipo:   TipoContencioso,
			Inicio: "2031-09-10",
		}, nombra: []string{`"contencioso"`, "un municipio", "0"}},
	}
}

// ordenDeLosErrores son peticiones con dos defectos: el mensaje nombra el
// primero en el orden de research D10 —tipo, fecha, duración, unidad, número
// de municipios— y calla el segundo.
func ordenDeLosErrores() []casoDeError {
	return []casoDeError{
		{nombre: "el tipo antes que la fecha", peticion: Peticion{
			Tipo: "inventado", Inicio: "2031-02-30",
			Municipios: los(villaPrimera),
		}, nombra: []string{`"inventado"`}, calla: []string{"2031-02-30"}},
		{nombre: "la fecha antes que la duración", peticion: Peticion{
			Tipo: TipoGenerico, Inicio: "2031-02-30",
			Duracion: "x", Unidad: "meses",
		}, nombra: []string{`"2031-02-30"`}, calla: []string{`"x"`}},
		{nombre: "la duración antes que la unidad", peticion: Peticion{
			Tipo: TipoGenerico, Inicio: "2031-09-10",
			Duracion: "x", Unidad: "semanas",
		}, nombra: []string{`"x"`}, calla: []string{"semanas"}},
		{nombre: "la duración de más antes que la unidad de más", peticion: Peticion{
			Tipo:   TipoReposicion,
			Inicio: "2031-09-10", Duracion: "2", Unidad: "meses",
		}, nombra: []string{`"2"`}, calla: []string{"meses"}},
		{
			nombre: "la unidad antes que el número de municipios", peticion: Peticion{
				Tipo:   TipoGenerico,
				Inicio: "2031-09-10", Duracion: "3", Unidad: "semanas", Municipios: los(villaPrimera),
			},
			nombra: []string{`"semanas"`}, calla: []string{"dos municipios"},
		},
	}
}

// comprobarErrores comprueba que cada petición no da plazo y da un error de
// clase argumentos, el mismo con Calcular y con ComprobarPeticion, que es lo
// que el applet comprueba antes de resolver los municipios.
func comprobarErrores(t *testing.T, calendario *Calendario, tipos []Tipo, casos []casoDeError) {
	t.Helper()

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			plazo, err := Calcular(calendario, tipos, caso.peticion)
			require.Error(t, err)
			assert.Equal(t, Plazo{}, plazo)
			comprobarClaseArgumentos(t, err)

			for _, fragmento := range caso.nombra {
				assert.Contains(t, err.Error(), fragmento)
			}

			for _, fragmento := range caso.calla {
				assert.NotContains(t, err.Error(), fragmento)
			}

			comprobado := ComprobarPeticion(tipos, caso.peticion)
			require.Error(t, comprobado)
			assert.Equal(t, err.Error(), comprobado.Error())
		})
	}
}

// comprobarClaseArgumentos exige que el error declare la clase argumentos,
// como la reconoce el kernel.
func comprobarClaseArgumentos(t *testing.T, err error) {
	t.Helper()

	var conClase schema.ConClase
	require.ErrorAs(t, err, &conClase)
	assert.Equal(t, schema.ClaseArgumentos, conClase.Clase())
}

// comprobarElUltimoDia calcula los plazos que llegan al 31 de diciembre de 9999,
// el último día que se escribe AAAA-MM-DD, y rechaza con un error de
// argumentos los que pasan de él, sea cual sea la regla que los lleva ahí; la
// comprobación previa a resolver los municipios no lo mira, porque la prórroga
// depende de los festivos.
func comprobarElUltimoDia(t *testing.T, calendario *Calendario, tipos []Tipo) {
	t.Helper()

	dos := los(villaPrimera, villaPrimera)

	plazo, err := Calcular(calendario, tipos, Peticion{
		Tipo: TipoGenerico, Inicio: "9999-12-30", Duracion: "1",
		Unidad: "dias-naturales", Municipios: dos,
	})
	require.NoError(t, err)
	assert.Equal(t, "9999-12-31", plazo.Fecha)

	for nombre, peticion := range map[string]Peticion{
		"días naturales":      {Tipo: TipoGenerico, Inicio: "2031-01-01", Duracion: "999999999", Unidad: "dias-naturales", Municipios: dos},
		"meses":               {Tipo: TipoGenerico, Inicio: "2031-01-01", Duracion: "100000", Unidad: "meses", Municipios: dos},
		"días hábiles de más": {Tipo: TipoGenerico, Inicio: "2031-01-01", Duracion: "4000000", Unidad: "dias-habiles", Municipios: dos},
		"días hábiles que el fin de semana lleva al final": {Tipo: TipoGenerico, Inicio: "9999-12-24", Duracion: "7", Unidad: "dias-habiles", Municipios: dos},
		"el día siguiente del silencio": {
			Tipo: TipoSilencio, Inicio: "9999-12-30", Duracion: "1", Unidad: "dias-naturales", Municipios: dos,
		},
		"el contencioso":              {Tipo: TipoContencioso, Inicio: "9999-11-15", Municipios: los(villaPrimera)},
		"la prórroga del contencioso": {Tipo: TipoContencioso, Inicio: "9999-10-31", Municipios: los(villaPrimera)},
	} {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			plazo, err := Calcular(calendario, tipos, peticion)
			require.Error(t, err)
			assert.Equal(t, Plazo{}, plazo)
			comprobarClaseArgumentos(t, err)
			assert.Contains(t, err.Error(), "9999-12-31")
			assert.Contains(t, err.Error(), peticion.Inicio)
			assert.NoError(t, ComprobarPeticion(tipos, peticion))
		})
	}
}

// TestComprobarPeticion comprueba que, de los municipios, la comprobación previa
// solo cuenta cuántos son: el applet la llama antes de resolverlos (FR-012).
func TestComprobarPeticion(t *testing.T) {
	t.Parallel()

	tipos := tiposDelContrato(t)

	require.NoError(t, ComprobarPeticion(tipos, Peticion{
		Tipo: TipoReposicion, Inicio: "2026-09-09", Municipios: make([]MunicipioResuelto, 2),
	}))
	require.NoError(t, ComprobarPeticion(tipos, Peticion{
		Tipo: TipoContencioso, Inicio: "2026-06-15", Municipios: make([]MunicipioResuelto, 1),
	}))

	err := ComprobarPeticion(tipos, Peticion{Tipo: TipoContencioso, Inicio: "2026-06-15"})
	require.Error(t, err)
	comprobarClaseArgumentos(t, err)
	assert.Contains(t, err.Error(), "un municipio")
}
