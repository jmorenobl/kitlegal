package data

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// arbolConEntrada es el subárbol de comunidades con una sola entrada, para
// ejercer cada rama de comunidadesDe sin tocar el embebido.
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

// TestComunidadesDe fija las tres ramas de error del recorrido, que el
// subárbol embebido no da nunca: el directorio que no se lista, la entrada que
// no es un fichero <código>.yaml y el fichero que no se lee.
func TestComunidadesDe(t *testing.T) {
	t.Parallel()

	t.Run("valida", func(t *testing.T) {
		t.Parallel()

		ficheros, err := comunidadesDe(arbolConEntrada("13.yaml", []byte("codigo: \"13\"\n")))
		require.NoError(t, err)
		assert.Equal(t, map[string][]byte{"13": []byte("codigo: \"13\"\n")}, ficheros)
	})

	t.Run("sin-el-subarbol", func(t *testing.T) {
		t.Parallel()

		ficheros, err := comunidadesDe(fstest.MapFS{})
		require.ErrorContains(t, err, "no se puede listar data/"+carpetaDeComunidades+": ")
		assert.Nil(t, ficheros)
	})

	t.Run("una-entrada-que-es-un-directorio", func(t *testing.T) {
		t.Parallel()

		arbol := fstest.MapFS{carpetaDeComunidades + "/13.yaml/dentro": {Data: []byte("x")}}

		ficheros, err := comunidadesDe(arbol)
		require.ErrorContains(t, err,
			"data/"+carpetaDeComunidades+"/13.yaml no es el fichero de una comunidad, <código>"+extensionDeComunidad)
		assert.Nil(t, ficheros)
	})

	t.Run("una-entrada-sin-la-extension", func(t *testing.T) {
		t.Parallel()

		ficheros, err := comunidadesDe(arbolConEntrada("LEEME.md", []byte("x")))
		require.ErrorContains(t, err,
			"data/"+carpetaDeComunidades+"/LEEME.md no es el fichero de una comunidad, <código>"+extensionDeComunidad)
		assert.Nil(t, ficheros)
	})

	t.Run("un-fichero-que-no-se-lee", func(t *testing.T) {
		t.Parallel()

		ficheros, err := comunidadesDe(arbolQueFallaAlLeer{arbolConEntrada("13.yaml", []byte("x"))})
		require.ErrorContains(t, err, "no se puede leer data/"+carpetaDeComunidades+"/13.yaml: ")
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
