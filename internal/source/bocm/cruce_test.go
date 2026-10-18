package bocm

import (
	"io/fs"
	"maps"
	"os"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/ids"
	"github.com/jmorenobl/kitlegal/internal/core/territorio"
)

// Dónde está, relativo al directorio del paquete, el territorio con el que se
// cruzan las líneas, y cuántos municipios tiene la comunidad del boletín
// (FR-022).
const (
	raizDelTerritorio  = "../../../data/territorio"
	numeroDeMunicipios = 179
)

// registroDelRepositorio es el territorio de data/territorio/, leído del árbol y
// cargado con territorio.Cargar, como lo carga el binario.
func registroDelRepositorio(t *testing.T) *territorio.Registro {
	t.Helper()

	arbol := os.DirFS(raizDelTerritorio)
	leer := func(ruta string) []byte {
		contenido, err := fs.ReadFile(arbol, ruta)
		require.NoError(t, err)

		return contenido
	}

	comunidades, err := fs.Glob(arbol, "comunidades/*.yaml")
	require.NoError(t, err)
	require.NotEmpty(t, comunidades, "%s/comunidades/ no tiene ningún fichero de comunidad", raizDelTerritorio)

	fuentes := territorio.Fuentes{
		Municipios:  leer("municipios.yaml"),
		DIR3:        leer("dir3.yaml"),
		Estado:      leer("estado.yaml"),
		Comunidades: make(map[string][]byte, len(comunidades)),
	}

	for _, fichero := range comunidades {
		fuentes.Comunidades[strings.TrimSuffix(path.Base(fichero), ".yaml")] = leer(fichero)
	}

	registro, err := territorio.Cargar(fuentes)
	require.NoError(t, err)

	return registro
}

// codigoDe es el código INE del municipio con ese nombre oficial, el de la
// relación del INE, entre los de la comunidad.
func codigoDe(t *testing.T, municipios []territorio.MunicipioConFormas, oficial string) string {
	t.Helper()

	indice := slices.IndexFunc(municipios, func(m territorio.MunicipioConFormas) bool { return m.Nombre == oficial })
	require.GreaterOrEqual(t, indice, 0, "data/territorio/ no tiene en la comunidad el municipio «%s»", oficial)

	return municipios[indice].Codigo.String()
}

// municipioSintetico es un municipio con ese código y esas formas, ya plegadas
// como las da formasDelNombre.
func municipioSintetico(t *testing.T, codigo string, formas ...string) territorio.MunicipioConFormas {
	t.Helper()

	ine, err := ids.AnalizarCodigoINE(codigo)
	require.NoError(t, err)

	return territorio.MunicipioConFormas{Codigo: ine, Nombre: formas[0], Formas: formas}
}

// probarCruce comprueba que cada nombre de la tabla casa con el código que se
// espera de él, o con ninguno si el código es vacío.
func probarCruce(t *testing.T, c *cruce, esperados map[string]string) {
	t.Helper()

	for _, nombre := range slices.Sorted(maps.Keys(esperados)) {
		codigo, casa := c.municipioDe(nombre)
		assert.Equal(t, esperados[nombre] != "", casa, "«%s»", nombre)
		assert.Equal(t, esperados[nombre], codigo, "«%s»", nombre)
	}
}

// TestCruceConElTerritorio fija el cruce de los nombres que escribe el BOCM con
// los municipios de la Comunidad de Madrid de data/territorio/, sin ninguna
// lista de alias (research D13, V7; FR-022, FR-024): los seis nombres de V7
// que el pliegue solo no alcanza, por la forma compacta o por el prefijo; los
// que llevan el artículo delante o pierden un diacrítico; las tres entidades
// menores, que no casan con ningún municipio; la Cerceda de A Coruña, que no
// entra porque el cruce se limita a la comunidad del boletín; y, con municipios
// sintéticos, que con más de un candidato no hay municipio.
func TestCruceConElTerritorio(t *testing.T) {
	t.Parallel()

	registro := registroDelRepositorio(t)
	municipios := registro.MunicipiosDe(ComunidadDelBoletin)
	require.Len(t, municipios, numeroDeMunicipios)

	c := nuevoCruce(municipios)

	t.Run("los-seis-nombres-de-v7", func(t *testing.T) {
		t.Parallel()

		delBoletinAlINE := map[string]string{
			"Berzosa de Lozoya":      "Berzosa del Lozoya",
			"Braojos de la Sierra":   "Braojos",
			"Buitrago de Lozoya":     "Buitrago del Lozoya",
			"Talamanca del Jarama":   "Talamanca de Jarama",
			"Moraleja de En medio":   "Moraleja de Enmedio",
			"Valdetorres del Jarama": "Valdetorres de Jarama",
		}

		esperados := map[string]string{}
		for delBoletin, oficial := range delBoletinAlINE {
			esperados[delBoletin] = codigoDe(t, municipios, oficial)
		}

		probarCruce(t, c, esperados)
	})

	t.Run("el-articulo-y-los-diacriticos", func(t *testing.T) {
		t.Parallel()

		delBoletinAlINE := map[string]string{
			"El Álamo":                    "Álamo, El",
			"Las Rozas de Madrid":         "Rozas de Madrid, Las",
			"Los Molinos":                 "Molinos, Los",
			"La Hiruela":                  "Hiruela, La",
			"San Martin de Valdeiglesias": "San Martín de Valdeiglesias",
			"Fuente El Saz de Jarama":     "Fuente el Saz de Jarama",
			"Valdepielagos":               "Valdepiélagos",
			"Manzanares El Real":          "Manzanares el Real",
			"Leganés":                     "Leganés",
		}

		esperados := map[string]string{}
		for delBoletin, oficial := range delBoletinAlINE {
			esperados[delBoletin] = codigoDe(t, municipios, oficial)
		}

		probarCruce(t, c, esperados)
	})

	t.Run("las-entidades-menores", func(t *testing.T) {
		t.Parallel()

		probarCruce(t, c, map[string]string{"Cerceda": "", "Mataelpino": "", "El Espartal": ""})

		const cercedaDeACoruna = "15024"

		assert.False(t, slices.ContainsFunc(municipios, func(m territorio.MunicipioConFormas) bool {
			return m.Codigo.String() == cercedaDeACoruna
		}), "la Cerceda de A Coruña no es un municipio de la comunidad del boletín")

		galicia := nuevoCruce(registro.MunicipiosDe("12"))
		probarCruce(t, galicia, map[string]string{"Cerceda": cercedaDeACoruna})
	})

	t.Run("sin-un-unico-candidato", func(t *testing.T) {
		t.Parallel()

		homonimos := nuevoCruce([]territorio.MunicipioConFormas{
			municipioSintetico(t, "28001", "villa"),
			municipioSintetico(t, "28002", "villa"),
			municipioSintetico(t, "28003", "villa alta"),
			municipioSintetico(t, "28004", "aldea, la", "la aldea"),
			municipioSintetico(t, "28005", "aldea de arriba"),
		})

		probarCruce(t, homonimos, map[string]string{
			"Villa":                "",
			"Villa de Arriba":      "",
			"Villa Alta":           "28003",
			"Villa Alta de Abajo":  "28003",
			"Villa Alta del Abajo": "28003",
			"Villa Alta Abajo":     "",
			"Alta":                 "",
			"La Aldea":             "28004",
			"Aldea":                "28004",
			"Aldea de Arriba":      "28005",
			"Aldea de Abajo":       "",
			"":                     "",
			"de la":                "",
		})
	})
}
