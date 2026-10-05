package evals

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parejaDeCarpetas es una fila de la tabla de TestCopiasDelJuez: la carpeta del
// juez de una skill y la carpeta de donde se copian sus cuatro ficheros, las dos
// relativas a la raíz del repositorio y con la barra de separador.
type parejaDeCarpetas struct {
	copias string
	origen string
}

// copiasDelJuez es la tabla de TestCopiasDelJuez: la carpeta del juez de cada
// skill que lo tiene, con la de la evidencia que la validó
// (contracts/medida-del-juez.md §3 de H24; FR-023). Una skill con juez que no
// esté aquí hace fallar el test.
var copiasDelJuez = []parejaDeCarpetas{
	{copias: "evals/boe-legislacion/juez", origen: "evidencias/adr-0037"},
}

// ficherosCopiadosDelJuez son los cuatro ficheros de la carpeta del juez que son
// copia de su original: la rúbrica, el esquema de la respuesta, los casos y la
// medida (FR-023). La declaración de clases no lo es: se escribe en la carpeta.
var ficherosCopiadosDelJuez = []string{"rubrica.md", "esquema.json", "casos.yaml", "medida.json"}

// TestCopiasDelJuez es el control de umbral de FR-023, FR-108 y SC-008 de H24
// (contracts/medida-del-juez.md §3 y §8): las cuatro copias de la carpeta del
// juez de cada skill de la tabla son idénticas, byte a byte, a su original, y
// ninguna skill tiene juez sin estar en la tabla. Cada defecto nombra su fichero.
//
// Sobre una copia de las dos carpetas en t.TempDir(), lo ve fallar: con un byte
// cambiado en cada una de las cuatro copias, y sin cada una de ellas, el único
// defecto es el de ese fichero, cuatro de cuatro; y la carpeta del juez de una
// skill que no está en la tabla da el suyo.
func TestCopiasDelJuez(t *testing.T) {
	t.Parallel()

	t.Run("del-repositorio", func(t *testing.T) {
		t.Parallel()

		defectos := defectosDeLasCopias(t, raizDelRepositorio, copiasDelJuez)
		assert.Empty(t, defectos, "copias del juez que no son las de su original:\n%s", strings.Join(defectos, "\n"))
	})

	for _, fichero := range ficherosCopiadosDelJuez {
		t.Run("cambiada-"+fichero, func(t *testing.T) {
			t.Parallel()

			probarCopiaCambiada(t, fichero)
		})

		t.Run("sin-"+fichero, func(t *testing.T) {
			t.Parallel()

			probarCopiaQueFalta(t, fichero)
		})
	}

	t.Run("skill-fuera-de-la-tabla", func(t *testing.T) {
		t.Parallel()

		raiz := copiarLasParejas(t, copiasDelJuez)
		require.NoError(t, os.MkdirAll(filepath.Join(raiz, "evals", "otra-skill", "juez"), 0o750))

		assert.Equal(t,
			[]string{"evals/otra-skill/juez: la skill tiene juez y su carpeta no está en la tabla de las copias"},
			defectosDeLasCopias(t, raiz, copiasDelJuez))
	})
}

// probarCopiaCambiada cambia un byte, el último, de la copia de ese fichero en
// una copia de las carpetas de la tabla, que antes no tenía ningún defecto, y
// exige el defecto que la nombra y ningún otro.
func probarCopiaCambiada(t *testing.T, fichero string) {
	t.Helper()

	pareja := copiasDelJuez[0]
	raiz := copiarLasParejas(t, copiasDelJuez)
	require.Empty(t, defectosDeLasCopias(t, raiz, copiasDelJuez), "la copia sin tocar no tiene ningún defecto")

	ruta := filepath.Join(raiz, pareja.copias, fichero)
	contenido := contenidoDelFichero(t, ruta)
	require.NotEmpty(t, contenido, "%s tiene algún byte que cambiar", ruta)

	contenido[len(contenido)-1] ^= 1
	require.NoError(t, os.WriteFile(ruta, contenido, 0o600))

	assert.Equal(t,
		[]string{pareja.copias + "/" + fichero + ": no es idéntico a " + pareja.origen + "/" + fichero},
		defectosDeLasCopias(t, raiz, copiasDelJuez))
}

// probarCopiaQueFalta quita la copia de ese fichero de una copia de las carpetas
// de la tabla y exige el defecto que la nombra y ningún otro.
func probarCopiaQueFalta(t *testing.T, fichero string) {
	t.Helper()

	pareja := copiasDelJuez[0]
	raiz := copiarLasParejas(t, copiasDelJuez)
	require.NoError(t, os.Remove(filepath.Join(raiz, pareja.copias, fichero)))

	defectos := defectosDeLasCopias(t, raiz, copiasDelJuez)
	require.Len(t, defectos, 1, "solo falta %s:\n%s", fichero, strings.Join(defectos, "\n"))
	assert.True(t, strings.HasPrefix(defectos[0], pareja.copias+"/"+fichero+": falta o no se puede leer: "), defectos[0])
}

// defectosDeLasCopias compara byte a byte, bajo raiz, cada fichero de
// ficherosCopiadosDelJuez de la carpeta de copias de cada pareja con el de su
// carpeta de origen, y devuelve un defecto por cada uno que difiere o que no
// se puede leer, con el nombre del fichero delante; y, detrás, uno por cada
// carpeta evals/<skill>/juez de raiz que no es la de ninguna pareja.
func defectosDeLasCopias(t *testing.T, raiz string, parejas []parejaDeCarpetas) []string {
	t.Helper()

	var defectos []string

	for _, pareja := range parejas {
		for _, fichero := range ficherosCopiadosDelJuez {
			copia, original := pareja.copias+"/"+fichero, pareja.origen+"/"+fichero

			contenidoDelOriginal, err := leerFichero(filepath.Join(raiz, original))
			if err != nil {
				defectos = append(defectos, original+": falta o no se puede leer: "+err.Error())

				continue
			}

			contenidoDeLaCopia, err := leerFichero(filepath.Join(raiz, copia))
			if err != nil {
				defectos = append(defectos, copia+": falta o no se puede leer: "+err.Error())

				continue
			}

			if !slices.Equal(contenidoDeLaCopia, contenidoDelOriginal) {
				defectos = append(defectos, copia+": no es idéntico a "+original)
			}
		}
	}

	return append(defectos, juecesFueraDeLaTabla(t, raiz, parejas)...)
}

// juecesFueraDeLaTabla devuelve un defecto por cada skill de raiz/evals que
// tiene la entrada juez sin que su carpeta sea la de copias de ninguna pareja, en
// orden de nombre. Lo que en raiz/evals no es una carpeta no son las evals de
// ninguna skill y no se mira.
func juecesFueraDeLaTabla(t *testing.T, raiz string, parejas []parejaDeCarpetas) []string {
	t.Helper()

	entradas, err := os.ReadDir(filepath.Join(raiz, "evals"))
	require.NoError(t, err, "el directorio de evals de %s", raiz)

	var defectos []string

	for _, entrada := range entradas {
		if !entrada.IsDir() {
			continue
		}

		carpeta := "evals/" + entrada.Name() + "/juez"

		_, err := os.Lstat(filepath.Join(raiz, carpeta))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		require.NoError(t, err, "la carpeta del juez %s", carpeta)

		if !slices.ContainsFunc(parejas, func(pareja parejaDeCarpetas) bool { return pareja.copias == carpeta }) {
			defectos = append(defectos, carpeta+": la skill tiene juez y su carpeta no está en la tabla de las copias")
		}
	}

	return defectos
}

// copiarLasParejas copia en un directorio temporal del test, con su ruta desde
// la raíz del repositorio, cada fichero de ficherosCopiadosDelJuez de las dos
// carpetas de cada pareja, y devuelve ese directorio.
func copiarLasParejas(t *testing.T, parejas []parejaDeCarpetas) string {
	t.Helper()

	raiz := t.TempDir()

	for _, pareja := range parejas {
		for _, carpeta := range []string{pareja.copias, pareja.origen} {
			require.NoError(t, os.MkdirAll(filepath.Join(raiz, carpeta), 0o750))

			for _, fichero := range ficherosCopiadosDelJuez {
				contenido := contenidoDelFichero(t, filepath.Join(raizDelRepositorio, carpeta, fichero))
				require.NoError(t, os.WriteFile(filepath.Join(raiz, carpeta, fichero), contenido, 0o600))
			}
		}
	}

	return raiz
}
