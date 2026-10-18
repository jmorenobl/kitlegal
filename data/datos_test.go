package data

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// arbolConEntrada es el subárbol de comunidades con una sola entrada, para
// ejercer cada rama de porComunidad sin tocar el embebido.
func arbolConEntrada(nombre string, contenido []byte) fstest.MapFS {
	return fstest.MapFS{carpetaDeComunidades + "/" + nombre: {Data: contenido}}
}

// TestComunidades fija que el subárbol embebido da un fichero por comunidad,
// indexado por su código.
func TestComunidades(t *testing.T) {
	t.Parallel()

	ficheros, err := Comunidades()
	require.NoError(t, err)
	require.Len(t, ficheros, 19)
	assert.Contains(t, string(ficheros["13"]), "Comunidad de Madrid")
	assert.NotContains(t, ficheros, "13.yaml", "la clave es el código, sin extensión")
}

// TestFicherosDePlazos fija los tres empotrados de plazos calcular: los tipos
// de data/plazos.yaml, los festivos nacionales y autonómicos de las
// resoluciones del BOE y los locales, uno por comunidad con su boletín leído e
// indexado por su código, que hoy es solo la Comunidad de Madrid (H9 FR-001,
// FR-020, FR-021; research.md D1 y D12 de H9). Cada uno es el fichero tal como
// está escrito: el paquete no interpreta nada.
func TestFicherosDePlazos(t *testing.T) {
	t.Parallel()

	assert.Contains(t, string(Plazos), "\ntipos:\n", "data/plazos.yaml")
	assert.Contains(t, string(FestivosNacionales), "\nfuente: boe.fiestas-laborales\n",
		"data/festivos/nacionales-y-autonomicos.yaml")

	locales, err := FestivosLocales()
	require.NoError(t, err)
	require.Len(t, locales, 1, "solo la Comunidad de Madrid tiene sus festivos locales cargados")
	assert.Contains(t, string(locales["13"]), "\nfuente: bocm.fiestas-locales\ncomunidad: \"13\"\n",
		"la clave es el código de la comunidad, sin extensión, y el fichero es el suyo")
}

// TestPorComunidad fija las tres ramas de error del recorrido, que los
// subárboles embebidos no dan nunca: el directorio que no se lista, la entrada
// que no es un fichero <código>.yaml y el fichero que no se lee.
func TestPorComunidad(t *testing.T) {
	t.Parallel()

	t.Run("valida", func(t *testing.T) {
		t.Parallel()

		ficheros, err := porComunidad(arbolConEntrada("13.yaml", []byte("codigo: \"13\"\n")), carpetaDeComunidades)
		require.NoError(t, err)
		assert.Equal(t, map[string][]byte{"13": []byte("codigo: \"13\"\n")}, ficheros)
	})

	t.Run("sin-el-subarbol", func(t *testing.T) {
		t.Parallel()

		ficheros, err := porComunidad(fstest.MapFS{}, carpetaDeComunidades)
		require.ErrorContains(t, err, "no se puede listar data/"+carpetaDeComunidades+": ")
		assert.Nil(t, ficheros)
	})

	t.Run("una-entrada-que-es-un-directorio", func(t *testing.T) {
		t.Parallel()

		arbol := fstest.MapFS{carpetaDeComunidades + "/13.yaml/dentro": {Data: []byte("x")}}

		ficheros, err := porComunidad(arbol, carpetaDeComunidades)
		require.ErrorContains(t, err,
			"data/"+carpetaDeComunidades+"/13.yaml no es el fichero de una comunidad, <código>"+extensionDeComunidad)
		assert.Nil(t, ficheros)
	})

	t.Run("una-entrada-sin-la-extension", func(t *testing.T) {
		t.Parallel()

		ficheros, err := porComunidad(arbolConEntrada("LEEME.md", []byte("x")), carpetaDeComunidades)
		require.ErrorContains(t, err,
			"data/"+carpetaDeComunidades+"/LEEME.md no es el fichero de una comunidad, <código>"+extensionDeComunidad)
		assert.Nil(t, ficheros)
	})

	t.Run("un-fichero-que-no-se-lee", func(t *testing.T) {
		t.Parallel()

		ficheros, err := porComunidad(arbolQueFallaAlLeer{arbolConEntrada("13.yaml", []byte("x"))}, carpetaDeComunidades)
		require.ErrorContains(t, err, "no se puede leer data/"+carpetaDeComunidades+"/13.yaml: ")
		assert.Nil(t, ficheros)
	})

	t.Run("la-carpeta-de-los-festivos-locales", func(t *testing.T) {
		t.Parallel()

		arbol := fstest.MapFS{carpetaDeFestivosLocales + "/LEEME.md": {Data: []byte("x")}}

		ficheros, err := porComunidad(arbol, carpetaDeFestivosLocales)
		require.ErrorContains(t, err,
			"data/"+carpetaDeFestivosLocales+"/LEEME.md no es el fichero de una comunidad, <código>"+extensionDeComunidad)
		assert.Nil(t, ficheros)
	})
}

// arbolQueFallaAlLeer lista como el árbol que envuelve, pero no deja abrir
// ningún fichero: es la única forma de llegar a la rama de lectura, porque
// ReadDir ya ha visto la entrada.
type arbolQueFallaAlLeer struct{ fs.FS }

func (a arbolQueFallaAlLeer) Open(nombre string) (fs.File, error) {
	if entrada, err := fs.Stat(a.FS, nombre); err == nil && !entrada.IsDir() {
		return nil, fs.ErrPermission
	}

	return a.FS.Open(nombre)
}
