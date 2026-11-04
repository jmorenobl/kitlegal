package evals

import (
	"bytes"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVariantesDeNormativa es el control de las variantes de normativa del
// laboratorio (FR-002 y SC-001 de H40; contracts/variantes.md §1, §2 y §7;
// data-model §1 y §2; research D1 a D4): normativa-control es skills/normativa,
// ruta a ruta y byte a byte, sin nada más en su directorio; las referencias de la
// familia son los ficheros de references/ de normativa-control, en orden de
// nombre; y ningún nombre de la familia es el de una variante de H39.
func TestVariantesDeNormativa(t *testing.T) {
	t.Parallel()

	repositorio := os.DirFS(raizDelRepositorio)

	t.Run("control es skills/normativa", func(t *testing.T) {
		t.Parallel()

		comprobarLaVarianteDeControl(t, repositorio)
	})

	t.Run("referencias de la familia", func(t *testing.T) {
		t.Parallel()

		referencias := path.Join(familiaDeNormativa.directorioDeLaSkill(varianteDeControlDeNormativa),
			directorioDeLasReferencias)

		entradas, err := fs.ReadDir(repositorio, referencias)
		require.NoError(t, err)

		ficheros := make([]string, 0, len(entradas))
		for _, entrada := range entradas {
			ficheros = append(ficheros, entrada.Name())
		}

		assert.Equal(t, ficheros, familiaDeNormativa.referencias,
			"las referencias de la familia de normativa son los ficheros de %s, en orden de nombre", referencias)
	})

	t.Run("nombres que no son de H39", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, varianteDeControlDeNormativa, familiaDeNormativa.variantes[0],
			"la primera variante de la familia de normativa es la de control")

		declarada, err := leerLaDeclaracionDelRepositorio(raizDelRepositorio)
		require.NoError(t, err)
		require.NotEmpty(t, declarada.Variantes, "la declaración del laboratorio de H39 tiene variantes")

		for _, deH39 := range declarada.Variantes {
			assert.NotContains(t, familiaDeNormativa.variantes, deH39.Nombre,
				"%s es el nombre de una variante de %s y de una de la familia de normativa", deH39.Nombre,
				path.Join(carpetaDelLaboratorio, ficheroDeLasVariantes))
		}
	})
}

// comprobarLaVarianteDeControl: el directorio de normativa-control tiene la
// skill de la familia con los ficheros de skills/normativa, las mismas rutas
// relativas y los mismos bytes, y nada más (FR-002); falla con la ruta del
// primer fichero, en orden, que difiere, falta o sobra.
func comprobarLaVarianteDeControl(t *testing.T, repositorio fs.FS) {
	t.Helper()

	deKitlegal := path.Join(directorioDeLasSkills, familiaDeNormativa.skill)

	deLaSkill, err := leerLosFicheros(repositorio, deKitlegal)
	require.NoError(t, err)
	require.NotEmpty(t, deLaSkill, "%s tiene algún fichero", deKitlegal)

	esperados := make(map[string][]byte, len(deLaSkill))
	for relativa, contenido := range deLaSkill {
		esperados[path.Join(carpetaDeLasSkillsDeLaVariante, familiaDeNormativa.skill, relativa)] = contenido
	}

	control := familiaDeNormativa.directorioDeLaVariante(varianteDeControlDeNormativa)

	obtenidos, err := leerLosFicheros(repositorio, control)
	require.NoError(t, err)

	relativa, motivo := primeraDiferencia(esperados, obtenidos)
	assert.Empty(t, motivo, "%s %s: la variante de control es %s, ruta a ruta y byte a byte",
		path.Join(control, relativa), motivo, deKitlegal)
}

// leerLosFicheros lee los ficheros del árbol bajo ese directorio, por su ruta
// relativa a él. Lo que no es un directorio ni un fichero normal —un enlace
// simbólico, que os.CopyFS copiaría como enlace (research V10)— es un error con
// su ruta.
func leerLosFicheros(arbol fs.FS, directorio string) (map[string][]byte, error) {
	ficheros := map[string][]byte{}

	err := fs.WalkDir(arbol, directorio, func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entrada.IsDir() {
			return nil
		}

		relativa, dentro := strings.CutPrefix(ruta, directorio+"/")
		if !dentro || !entrada.Type().IsRegular() {
			return fmt.Errorf("%s no es un fichero normal dentro de un directorio", ruta)
		}

		contenido, err := fs.ReadFile(arbol, ruta)
		if err != nil {
			return err
		}

		ficheros[relativa] = contenido

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("los ficheros de %s: %w", directorio, err)
	}

	return ficheros, nil
}

// primeraDiferencia es la primera ruta, en orden, en que los ficheros obtenidos
// no son los esperados, con su motivo: «difiere», si los dos la tienen con bytes
// distintos; «falta», si solo está en los esperados; y «sobra», si solo está en
// los obtenidos. Sin diferencias, las dos son vacías.
func primeraDiferencia(esperados, obtenidos map[string][]byte) (ruta, motivo string) {
	rutas := slices.Sorted(maps.Keys(esperados))
	rutas = append(rutas, slices.Collect(maps.Keys(obtenidos))...)
	slices.Sort(rutas)

	for _, ruta := range slices.Compact(rutas) {
		esperado, loEspera := esperados[ruta]
		obtenido, loTiene := obtenidos[ruta]

		switch {
		case !loTiene:
			return ruta, "falta"
		case !loEspera:
			return ruta, "sobra"
		case !bytes.Equal(esperado, obtenido):
			return ruta, "difiere"
		}
	}

	return "", ""
}
