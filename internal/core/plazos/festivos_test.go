package plazos

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Los ficheros sintéticos de la escritura, con sus filas y sus comunidades
// desordenadas a propósito: el orden lo pone la escritura canónica.
func nacionalesDesordenados() FestivosNacionales {
	return FestivosNacionales{
		Fuente: "prueba.fiestas",
		Anios: []AnioNacional{
			{Anio: 2032, Disposicion: "PRUEBA-2031-1"},
			{Anio: 2031, Disposicion: "PRUEBA-2030-1"},
		},
		Festivos: []FestivoNacional{
			{Fecha: "2032-01-01", Ambito: AmbitoNacional, Comunidades: []string{"72", "71"}, Source: "PRUEBA-2031-1"},
			{Fecha: "2031-06-24", Ambito: AmbitoAutonomico, Comunidades: []string{"72"}, Parte: "Isla Sintética", Source: "PRUEBA-2030-1"},
			{Fecha: "2031-03-19", Ambito: AmbitoAutonomico, Comunidades: []string{"72"}, Source: "PRUEBA-2030-1"},
			{Fecha: "2031-06-24", Ambito: AmbitoAutonomico, Comunidades: []string{"73", "72"}, Source: "PRUEBA-2030-1"},
			{Fecha: "2031-03-19", Ambito: AmbitoNacional, Comunidades: []string{"73", "71"}, Source: "PRUEBA-2030-1"},
			{Fecha: "2031-06-24", Ambito: AmbitoAutonomico, Comunidades: []string{"71"}, Parte: "Comarca Sintética", Source: "PRUEBA-2030-1"},
		},
	}
}

func localesDesordenados() FestivosLocales {
	return FestivosLocales{
		Fuente:    "prueba.fiestas-locales",
		Comunidad: "71",
		Anios: []AnioLocal{
			{Anio: 2032, Disposiciones: []string{"PRUEBA-L-3"}},
			{Anio: 2031, Disposiciones: []string{"PRUEBA-L-2", "PRUEBA-L-1"}},
		},
		Festivos: []FestivoLocal{
			{Fecha: "2032-09-08", Municipio: "71001", Source: "PRUEBA-L-3"},
			{Fecha: "2031-05-15", Municipio: "71002", Source: "PRUEBA-L-2"},
			{Fecha: "2031-05-15", Municipio: "71001", Source: "PRUEBA-L-1"},
		},
		Faltas: []FaltaLocal{
			{Anio: 2032, Municipio: "71002", Motivo: FaltaSinLinea, Source: "PRUEBA-L-3"},
			{Anio: 2031, Linea: "Aldea Sintética", Motivo: FaltaSinMunicipio, Source: "PRUEBA-L-1"},
			{Anio: 2031, Municipio: "71004", Motivo: FaltaLineaSinLeer, Source: "PRUEBA-L-2"},
			{Anio: 2031, Linea: "Aldea Antigua", Motivo: FaltaSinMunicipio, Source: "PRUEBA-L-2"},
			{Anio: 2031, Municipio: "71003", Motivo: FaltaNoComunicado, Source: "PRUEBA-L-1"},
		},
	}
}

// Lo que escribe la forma canónica de los dos ficheros sintéticos: la línea de
// fichero generado, las claves en el orden de contracts/datos-de-plazos.md §1 y
// las listas ordenadas —los años por año; los nacionales por fecha, ámbito
// (nacional antes que autonómico) y parte (sin parte antes); los locales por
// fecha y municipio; las faltas por año, municipio y línea, las de un
// municipio antes que las de una línea—. Las disposiciones de un año quedan en
// el orden dado, que es el de su publicación.
const (
	nacionalesCanonicos = `# generado desde evidencias/prueba/prueba.fiestas/ por TestDePrueba (internal/source/prueba), no editar
fuente: prueba.fiestas
anios:
  - {anio: 2031, disposicion: PRUEBA-2030-1}
  - {anio: 2032, disposicion: PRUEBA-2031-1}
festivos:
  - {fecha: "2031-03-19", ambito: nacional, comunidades: ["71", "73"], source: PRUEBA-2030-1}
  - {fecha: "2031-03-19", ambito: autonomico, comunidades: ["72"], source: PRUEBA-2030-1}
  - {fecha: "2031-06-24", ambito: autonomico, comunidades: ["72", "73"], source: PRUEBA-2030-1}
  - {fecha: "2031-06-24", ambito: autonomico, comunidades: ["71"], parte: Comarca Sintética, source: PRUEBA-2030-1}
  - {fecha: "2031-06-24", ambito: autonomico, comunidades: ["72"], parte: Isla Sintética, source: PRUEBA-2030-1}
  - {fecha: "2032-01-01", ambito: nacional, comunidades: ["71", "72"], source: PRUEBA-2031-1}
`
	localesCanonicos = `# generado desde evidencias/prueba/prueba.fiestas-locales/ por TestDePruebaLocal (internal/source/pruebalocal), no editar
fuente: prueba.fiestas-locales
comunidad: "71"
anios:
  - {anio: 2031, disposiciones: [PRUEBA-L-2, PRUEBA-L-1]}
  - {anio: 2032, disposiciones: [PRUEBA-L-3]}
festivos:
  - {fecha: "2031-05-15", municipio: "71001", source: PRUEBA-L-1}
  - {fecha: "2031-05-15", municipio: "71002", source: PRUEBA-L-2}
  - {fecha: "2032-09-08", municipio: "71001", source: PRUEBA-L-3}
faltas:
  - {anio: 2031, municipio: "71003", motivo: no-comunicado, source: PRUEBA-L-1}
  - {anio: 2031, municipio: "71004", motivo: linea-sin-leer, source: PRUEBA-L-2}
  - {anio: 2031, linea: Aldea Antigua, motivo: sin-municipio, source: PRUEBA-L-2}
  - {anio: 2031, linea: Aldea Sintética, motivo: sin-municipio, source: PRUEBA-L-1}
  - {anio: 2032, municipio: "71002", motivo: sin-linea, source: PRUEBA-L-3}
`
)

var (
	cabeceraNacional = Cabecera{
		Desde: "evidencias/prueba/prueba.fiestas/", Test: "TestDePrueba", Paquete: "internal/source/prueba",
	}
	cabeceraLocal = Cabecera{
		Desde: "evidencias/prueba/prueba.fiestas-locales/", Test: "TestDePruebaLocal", Paquete: "internal/source/pruebalocal",
	}
)

// TestEscribirFestivos comprueba la forma canónica de los dos ficheros de
// data/festivos/: su línea de fichero generado, el orden de sus claves y de sus
// filas, que no toca lo que recibe, y la ida y vuelta: lo escrito se carga y se
// vuelve a escribir igual byte a byte (FR-020, FR-022, FR-025;
// contracts/datos-de-plazos.md §1).
func TestEscribirFestivos(t *testing.T) {
	t.Parallel()

	t.Run("nacionales y autonómicos", func(t *testing.T) {
		t.Parallel()

		fichero := nacionalesDesordenados()

		escrito, err := EscribirFestivosNacionales(fichero, cabeceraNacional)
		require.NoError(t, err)
		assert.Equal(t, nacionalesCanonicos, string(escrito))
		assert.Equal(t, nacionalesDesordenados(), fichero, "la escritura no reordena lo que recibe")

		leido, err := leerNacionales(escrito)
		require.NoError(t, err)

		reescrito, err := EscribirFestivosNacionales(leido, cabeceraNacional)
		require.NoError(t, err)
		assert.Equal(t, string(escrito), string(reescrito))
	})

	t.Run("locales", func(t *testing.T) {
		t.Parallel()

		fichero := localesDesordenados()

		escrito, err := EscribirFestivosLocales(fichero, cabeceraLocal)
		require.NoError(t, err)
		assert.Equal(t, localesCanonicos, string(escrito))
		assert.Equal(t, localesDesordenados(), fichero, "la escritura no reordena lo que recibe")

		leido, err := leerLocales("71", escrito)
		require.NoError(t, err)

		reescrito, err := EscribirFestivosLocales(leido, cabeceraLocal)
		require.NoError(t, err)
		assert.Equal(t, string(escrito), string(reescrito))
	})

	t.Run("sin filas", func(t *testing.T) {
		t.Parallel()

		escrito, err := EscribirFestivosLocales(FestivosLocales{Fuente: "prueba.fiestas-locales", Comunidad: "72"}, cabeceraLocal)
		require.NoError(t, err)
		assert.Equal(t, "# generado desde evidencias/prueba/prueba.fiestas-locales/ por TestDePruebaLocal "+
			"(internal/source/pruebalocal), no editar\n"+
			"fuente: prueba.fiestas-locales\ncomunidad: \"72\"\nanios: []\nfestivos: []\nfaltas: []\n", string(escrito))

		leido, err := leerLocales("72", escrito)
		require.NoError(t, err)

		reescrito, err := EscribirFestivosLocales(leido, cabeceraLocal)
		require.NoError(t, err)
		assert.Equal(t, string(escrito), string(reescrito))
	})

	t.Run("lo escrito se carga", func(t *testing.T) {
		t.Parallel()

		nacionales, err := EscribirFestivosNacionales(nacionalesDesordenados(), cabeceraNacional)
		require.NoError(t, err)

		locales, err := EscribirFestivosLocales(localesDesordenados(), cabeceraLocal)
		require.NoError(t, err)

		calendario, err := CargarCalendario(nacionales, map[string][]byte{"71": locales})
		require.NoError(t, err)

		assert.True(t, calendario.Inhabil(dia(t, "2031-03-19"), []Municipio{conFestivos}))
		assert.True(t, calendario.Inhabil(dia(t, "2031-05-15"), []Municipio{conFestivos}))
		assert.False(t, calendario.Inhabil(dia(t, "2031-06-24"), []Municipio{conFestivos}), "solo en una comarca de la 71")
		assert.Equal(t, []FestivoDeUnaParte{{Ambito: AmbitoAutonomico, Parte: "Comarca Sintética", Source: "PRUEBA-2030-1"}},
			calendario.PartesDe(dia(t, "2031-06-24"), "71"))
		assert.Equal(t, []Tramo{
			{Ambito: CoberturaNacionalYAutonomica, Desde: 2031, Hasta: 2031, Disposicion: "PRUEBA-2030-1"},
			{Ambito: CoberturaLocal, Municipio: "71001", Desde: 2031, Hasta: 2031, Disposicion: "PRUEBA-L-1"},
			{Ambito: CoberturaLocal, Municipio: "71003", Desde: 2031, Hasta: 2031, Falta: FaltaNoComunicado},
		}, calendario.Cobertura(2031, 2031, []Municipio{conFestivos, noComunicado}),
			"la disposición de un año es la última de las suyas, en el orden en que se dan")
	})
}
