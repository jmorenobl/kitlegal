package app

import (
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/plazos"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/core/territorio"
)

// tiposDePruebaDelCivil es el fichero de los dos tipos de prueba que cuentan
// con el cómputo civil de data/plazos.yaml, por su ruta desde internal/app
// (FR-023; contracts/datos-de-plazos.md §4 de H36).
const tiposDePruebaDelCivil = "../core/plazos/testdata/tipos-de-prueba-civil.yaml"

// Lo que estos tests esperan del cómputo civil, escrito aquí otra vez y no
// tomado del código que se prueba (spec §C.9 de H36).
const (
	// computoCivil es el código del cómputo civil de la LEC.
	computoCivil plazos.Computo = "civil"
	// pruebaCivilEnDias es el tipo de prueba de los casos en días, con el 133.2.
	pruebaCivilEnDias = "prueba-civil-dias"
	// pruebaCivilEnMeses es el tipo de prueba de los casos en meses, con el
	// 133.3.
	pruebaCivilEnMeses = "prueba-civil-meses"
	// sedeDelCivil es la sede de los goldens, Madrid, por su código INE.
	sedeDelCivil = "28079"
	// segundoMunicipioDelCivil es el municipio de más de C8.
	segundoMunicipioDelCivil = "Leganés"
	// avisoDeLosLocalesQueFaltan es el código del aviso de C2 y C4: no hay
	// festivos locales de Madrid de 2027.
	avisoDeLosLocalesQueFaltan = "faltan-festivos-locales"
)

// Las disposiciones de los festivos que saltan los goldens del civil: las
// resoluciones de los nacionales de 2025, 2026 y 2027, y la de los locales de
// Madrid de 2026 (spec §C.9 de H36).
const (
	nacionalesDe2025      = "BOE-A-2024-21316"
	nacionalesDe2026      = "BOE-A-2025-21667"
	nacionalesDe2027      = "BOE-A-2026-21254"
	localesDeMadridDe2026 = "BOCM-20251212-34"
)

// casoCivil es una fila de la tabla de spec §C.9 de H36: la petición y, si da
// plazo, su fecha, los festivos que salta y los códigos de sus avisos en su
// orden o, si no, el rechazo.
type casoCivil struct {
	nombre     string
	tipo       string
	inicio     string
	duracion   string
	unidad     plazos.Unidad
	municipios []string
	fecha      string
	festivos   []plazos.FestivoSaltado
	avisos     []string
	rechazo    *rechazoCivil
}

// rechazoCivil es lo que da una fila de spec §C.9 que no da plazo: el código de
// salida, la clase y el mensaje entero o, en C6, el único precepto que nombra
// el mensaje, como «<norma>, bloque <bloque>, apartado <apartado>».
type rechazoCivil struct {
	codigo   int
	clase    schema.Clase
	mensaje  string
	precepto string
}

// nacional es un festivo nacional saltado, con la resolución que lo fija.
func nacional(fecha, source string) plazos.FestivoSaltado {
	return plazos.FestivoSaltado{Fecha: fecha, Ambito: "nacional", Source: source}
}

// casosCiviles son las filas C1 a C10 de spec §C.9 de H36. Sin municipios, el
// de la sede, Madrid.
func casosCiviles() []casoCivil {
	return []casoCivil{
		{
			nombre: "C1 agosto inhábil en días hábiles",
			tipo:   pruebaCivilEnDias, inicio: "2026-07-24", duracion: "10", unidad: plazos.UnidadDiasHabiles,
			fecha: "2026-09-07",
		},
		{
			nombre: "C2 del 24 de diciembre al 6 de enero inhábiles",
			tipo:   pruebaCivilEnDias, inicio: "2026-12-18", duracion: "5", unidad: plazos.UnidadDiasHabiles,
			fecha: "2027-01-08",
			festivos: []plazos.FestivoSaltado{
				nacional("2026-12-25", nacionalesDe2026),
				nacional("2027-01-01", nacionalesDe2027),
				nacional("2027-01-06", nacionalesDe2027),
			},
			avisos: []string{avisoDeLosLocalesQueFaltan},
		},
		{
			nombre: "C3 prórroga de los meses fuera de agosto",
			tipo:   pruebaCivilEnMeses, inicio: "2026-06-15", duracion: "2", unidad: plazos.UnidadMeses,
			fecha: "2026-09-01",
		},
		{
			nombre: "C4 prórroga de los meses fuera de la Navidad",
			tipo:   pruebaCivilEnMeses, inicio: "2026-10-27", duracion: "2", unidad: plazos.UnidadMeses,
			fecha: "2027-01-07",
			festivos: []plazos.FestivoSaltado{
				nacional("2027-01-01", nacionalesDe2027),
				nacional("2027-01-06", nacionalesDe2027),
			},
			avisos: []string{avisoDeLosLocalesQueFaltan},
		},
		{
			nombre: "C5 agosto no para los meses",
			tipo:   pruebaCivilEnMeses, inicio: "2026-07-10", duracion: "2", unidad: plazos.UnidadMeses,
			fecha: "2026-09-10",
		},
		{
			nombre: "C6 el 130.2 no regía",
			tipo:   pruebaCivilEnMeses, inicio: "2022-12-22", duracion: "2", unidad: plazos.UnidadMeses,
			rechazo: &rechazoCivil{
				codigo: 3, clase: schema.ClaseNoEncontrado,
				precepto: "BOE-A-2000-323, bloque a130, apartado 130.2",
			},
		},
		{
			nombre: "C7 sin días naturales",
			tipo:   pruebaCivilEnDias, inicio: "2026-07-24", duracion: "10", unidad: plazos.UnidadDiasNaturales,
			rechazo: &rechazoCivil{
				codigo: 2, clase: schema.ClaseArgumentos,
				mensaje: `el tipo "prueba-civil-dias" no admite la unidad "dias-naturales": su cómputo, "civil", ` +
					`admite meses o dias-habiles`,
			},
		},
		{
			nombre: "C8 un municipio, la sede",
			tipo:   pruebaCivilEnDias, inicio: "2026-07-24", duracion: "10", unidad: plazos.UnidadDiasHabiles,
			municipios: []string{sedeDelCivil, segundoMunicipioDelCivil},
			rechazo: &rechazoCivil{
				codigo: 2, clase: schema.ClaseArgumentos,
				mensaje: `el tipo "prueba-civil-dias" se cuenta con un municipio, el de la sede del tribunal, ` +
					`y se han dado 2`,
			},
		},
		{
			nombre: "C9 primer día de agosto",
			tipo:   pruebaCivilEnDias, inicio: "2025-07-30", duracion: "2", unidad: plazos.UnidadDiasHabiles,
			fecha:    "2025-09-01",
			festivos: []plazos.FestivoSaltado{nacional("2025-08-15", nacionalesDe2025)},
		},
		{
			nombre: "C10 fiesta local de la sede",
			tipo:   pruebaCivilEnDias, inicio: "2026-11-02", duracion: "5", unidad: plazos.UnidadDiasHabiles,
			fecha: "2026-11-10",
			festivos: []plazos.FestivoSaltado{
				{Fecha: "2026-11-09", Ambito: "local", Municipio: sedeDelCivil, Source: localesDeMadridDe2026},
			},
		},
	}
}

// tiposConLosDePruebaDelCivil son los tipos de data/plazos.yaml con los dos de
// prueba del civil: el fichero con su única línea «tipos:» sustituida por el
// de prueba, que lleva la suya y sus tipos con la misma sangría, leído con
// LeerTipos (research D7 de H36).
func tiposConLosDePruebaDelCivil(t *testing.T, fuentes FicherosDePlazos) []plazos.Tipo {
	t.Helper()

	prueba, err := os.ReadFile(tiposDePruebaDelCivil)
	require.NoError(t, err, "el fichero de los tipos de prueba del civil se lee")

	combinado := sustituida(t, string(fuentes.Tipos), "\ntipos:\n", "\n"+string(prueba))

	reales, err := plazos.LeerTipos(fuentes.Tipos)
	require.NoError(t, err, "data/plazos.yaml se lee sin defecto")

	tipos, err := plazos.LeerTipos([]byte(combinado))
	require.NoError(t, err, "data/plazos.yaml con los tipos de prueba del civil se lee sin defecto")
	require.Len(t, tipos, len(reales)+2, "la combinación da los dos tipos de prueba más que data/plazos.yaml")

	for _, codigo := range []string{pruebaCivilEnDias, pruebaCivilEnMeses} {
		indice := slices.IndexFunc(tipos, func(tipo plazos.Tipo) bool { return tipo.Codigo == codigo })
		require.GreaterOrEqual(t, indice, 0, "la combinación lleva el tipo de prueba %q", codigo)
		require.Equal(t, computoCivil, tipos[indice].Computo, "el tipo de prueba %q cuenta con el civil", codigo)
	}

	return tipos
}

// TestComputoCivil calcula con plazos.Calcular, el cómputo civil de
// data/plazos.yaml, los dos tipos de prueba que cuentan con él, los festivos de
// data/festivos/ y la sede resuelta por territorio, cada fila C1 a C10 de spec
// §C.9 de H36 (FR-023; US2.1 a US2.4; SC-002): un plazo da su fecha, los
// festivos que salta y los códigos de sus avisos en su orden; un rechazo, su
// código de salida, su clase y su mensaje. Y exige que ningún tipo de
// data/plazos.yaml cuenta con el civil, que lee sin defecto (FR-014; US2.5,
// US5.1).
func TestComputoCivil(t *testing.T) {
	t.Parallel()

	fuentes := fuentesDePlazos(t)
	tipos := tiposConLosDePruebaDelCivil(t, fuentes)

	calendario, err := plazos.CargarCalendario(fuentes.Nacionales, fuentes.Locales)
	require.NoError(t, err, "los festivos de data/festivos/ cargan")

	registro, err := territorio.Cargar(fuentes.Territorio)
	require.NoError(t, err, "el territorio del binario carga")

	t.Run("ningún tipo de data/plazos.yaml cuenta con el civil", func(t *testing.T) {
		t.Parallel()

		reales, err := plazos.LeerTipos(fuentes.Tipos)
		require.NoError(t, err, "data/plazos.yaml se lee sin defecto")
		require.NotEmpty(t, reales)

		for _, tipo := range reales {
			assert.NotEqual(t, computoCivil, tipo.Computo, "el tipo %q de data/plazos.yaml cuenta con el civil",
				tipo.Codigo)
		}
	})

	for _, caso := range casosCiviles() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			pedidos := caso.municipios
			if len(pedidos) == 0 {
				pedidos = []string{sedeDelCivil}
			}

			municipios, err := resolverMunicipios(registro, pedidos)
			require.NoError(t, err, "territorio resuelve %q", pedidos)

			plazo, err := plazos.Calcular(calendario, tipos, plazos.Peticion{
				Tipo: caso.tipo, Inicio: caso.inicio, Duracion: caso.duracion, Unidad: string(caso.unidad),
				Municipios: municipios,
			})

			if caso.rechazo != nil {
				require.Error(t, err)
				assert.Equal(t, plazos.Plazo{}, plazo)
				comprobarRechazoCivil(t, tipos, caso.tipo, *caso.rechazo, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, caso.fecha, plazo.Fecha)
			assert.Equal(t, caso.tipo, plazo.Tipo)
			assert.Equal(t, append([]plazos.FestivoSaltado{}, caso.festivos...), plazo.Festivos)
			assert.Equal(t, append([]string{}, caso.avisos...), codigosDeLosAvisos(plazo.Avisos))
		})
	}
}

// comprobarRechazoCivil compara un rechazo con el de su fila: el código de
// salida y la clase con que lo da el kernel y, o el mensaje entero, o que el
// mensaje nombra el precepto de la fila y ningún otro de los del tipo y su
// cómputo.
func comprobarRechazoCivil(t *testing.T, tipos []plazos.Tipo, codigo string, rechazo rechazoCivil, err error) {
	t.Helper()

	assert.Equal(t, rechazo.codigo, cli.CodigoSalida(err))
	assert.Equal(t, rechazo.clase, cli.Clasificar(err))

	if rechazo.mensaje != "" {
		assert.Equal(t, rechazo.mensaje, err.Error())
	}

	if rechazo.precepto == "" {
		return
	}

	assert.Contains(t, err.Error(), rechazo.precepto)

	indice := slices.IndexFunc(tipos, func(tipo plazos.Tipo) bool { return tipo.Codigo == codigo })
	require.GreaterOrEqual(t, indice, 0)

	otros := 0

	for _, precepto := range slices.Concat(tipos[indice].Preceptos, tipos[indice].PreceptosDelComputo) {
		nombrado := fmt.Sprintf("%s, bloque %s, apartado %s", precepto.Norma, precepto.Bloque, precepto.Apartado)
		if nombrado == rechazo.precepto {
			continue
		}

		otros++

		assert.NotContains(t, err.Error(), nombrado, "el mensaje solo nombra %s", rechazo.precepto)
	}

	require.Positive(t, otros, "premisa: el tipo y su cómputo tienen otros preceptos que el mensaje no nombra")
}

// codigosDeLosAvisos son los códigos de unos avisos, en su orden.
func codigosDeLosAvisos(avisos []plazos.Aviso) []string {
	codigos := make([]string, 0, len(avisos))
	for _, aviso := range avisos {
		codigos = append(codigos, aviso.Codigo)
	}

	return codigos
}
