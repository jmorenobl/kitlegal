package evals

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"maps"
	"os"
	"path"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// huellasDeLosControles son las huellas SHA-256 de los ficheros de la variante de
// control de cada familia del producto, por su skill y por su ruta desde el
// directorio de la variante, escritas al copiarlos de skills/ en v0.9.0 (research
// D9 y V1; data-model §2), calculadas con shasum -a 256 sobre las copias: 9 de
// normativa, 5 de plazos-y-territorio y 3 de jurisprudencia. No se comparan con
// skills/<skill>, que cambia con el producto (FR-020), ni con la etiqueta, que el
// clon superficial del make ci del cierre no tiene.
var huellasDeLosControles = map[string]map[string]string{
	skillDeNormativa: {
		"skills/normativa/SKILL.md":                             "9678c3c1657d26c854a7fbe915c0cb94346af4368e75996bef52695de7478f62",
		"skills/normativa/references/buscar_la_norma.md":        "7bd42668c82fd3ee12b9995fa31862676317450baeb1686beb7f02161f1a005e",
		"skills/normativa/references/como_se_cita.md":           "d61e44c2b0e9ed98894872d1bb45d9f9d70c25232d4eb4acdffcd64de749a5bc",
		"skills/normativa/references/con_herramientas.md":       "66da56acc657881183e5413f11ea4d28c17034a8e6c07498e475b8121e7dfb84",
		"skills/normativa/references/lecturas_que_fallan.md":    "e7b775aaf95c693d6afe01fbf9f8011b2142cb69fbccebf4c1b64b94ce6a28f7",
		"skills/normativa/references/normas.md":                 "3e6d1fbc22b086df5430cfd533fd8b97cba3ead10161c291a8b0f6e2168cc5ef",
		"skills/normativa/references/redaccion_en_una_fecha.md": "b9b867bc4fd9f812e12415f789a7fba37d78828eb1b5760b731dfc04b6b96632",
		"skills/normativa/references/redaccion_modificada.md":   "a90adb4335f6dfe979293cb650f0645de476a2614d28c8abadebe0b34c0a583f",
		"skills/normativa/references/varios_bloques.md":         "413e9d730134d4f804b41c06f8897f32c777b522cb669442518effa5b79c7fcc",
	},
	skillDeLaFamiliaDePlazos: {
		"skills/plazos-y-territorio/SKILL.md":                          "9fdc08c6d0680f43ab7976d591740d47b49465b32ab167f4aed6c95d640688d0",
		"skills/plazos-y-territorio/references/administrativo.md":      "54c032caceea6ae96390a434a8f429818268a502772103a385faf635a774cfe8",
		"skills/plazos-y-territorio/references/calculo_de_plazos.md":   "f85af1c493b9332ae788c850f534fc7b69fea0318f83f2769885aaf60bf47eb5",
		"skills/plazos-y-territorio/references/jerarquia_normativa.md": "339e60d49cf36e1cb713a22078d890e1d1300364662b58141dfcb4658f5cfe70",
		"skills/plazos-y-territorio/references/municipios.md":          "0afac8314160be8c99d499554c3abd9db8978b0f67901c9e2c63fa8e6fec5a0e",
	},
	skillDeLaFamiliaDeJurisprudencia: {
		"skills/jurisprudencia/SKILL.md":                         "b93e13e82b19fda334413dd791c6dc57a3f6eab44f9e7d6303d7f1a88f5ef7fc",
		"skills/jurisprudencia/references/consulta_preparada.md": "f4dcb6dd409406c1429c871779c39244864a945839c80d3c9ac8c15bbb4426e8",
		"skills/jurisprudencia/references/documento_aportado.md": "b2f0137c4d0540587ff6925d9be2cefd55648b1d52bbbbca454c0c1f7bcff912",
	},
}

// TestVariantesDeControl es el control de las variantes versionadas de las
// familias del producto (FR-020, FR-022; SC-007, SC-009; contracts/variantes.md
// §1 y §4; data-model §1 y §2; research D9, D10 y V1): son normativa,
// plazos-y-territorio y jurisprudencia, en ese orden; en cada una, la primera
// variante es <skill>-control y la del producto, <skill>-producto; la carpeta de
// la familia tiene un directorio por variante versionada y nada más, y ninguno es
// el de la del producto; los ficheros de <skill>-control son, ruta a ruta, los de
// sus huellas; sus referencias son los ficheros de references/ de su control, en
// orden de nombre; su directorio de evals tiene las evals de su sondeo; y los
// nombres de las variantes de las cuatro familias, la de H39 y las tres, no se
// repiten.
func TestVariantesDeControl(t *testing.T) {
	t.Parallel()

	repositorio := os.DirFS(raizDelRepositorio)

	t.Run("familias del producto", func(t *testing.T) {
		t.Parallel()

		skills := make([]string, 0, len(familiasDelProducto))
		for _, familia := range familiasDelProducto {
			skills = append(skills, familia.skill)
		}

		assert.Equal(t, []string{skillDeNormativa, skillDeLaFamiliaDePlazos, skillDeLaFamiliaDeJurisprudencia}, skills,
			"las familias del producto son las de normativa, plazos-y-territorio y jurisprudencia, en ese orden")
		assert.ElementsMatch(t, skills, slices.Collect(maps.Keys(huellasDeLosControles)),
			"hay huellas del control de cada familia del producto y de ninguna otra")
		assert.Equal(t, familiasDelProducto[0].versionada(), familiaDeNormativa,
			"la familia de normativa es la primera del producto, con sus variantes versionadas")
	})

	for _, familia := range familiasDelProducto {
		t.Run(familia.skill, func(t *testing.T) {
			t.Parallel()

			t.Run("control y producto", func(t *testing.T) {
				t.Parallel()

				comprobarElControlYElProducto(t, familia)
			})

			t.Run("directorios de la familia", func(t *testing.T) {
				t.Parallel()

				comprobarLosDirectoriosDeLaFamilia(t, repositorio, familia)
			})

			t.Run("huellas del control", func(t *testing.T) {
				t.Parallel()

				comprobarLasHuellasDelControl(t, repositorio, familia)
			})

			t.Run("referencias de la familia", func(t *testing.T) {
				t.Parallel()

				comprobarLasReferenciasDeLaFamilia(t, repositorio, familia)
			})

			t.Run("evals del sondeo", func(t *testing.T) {
				t.Parallel()

				for _, fichero := range familia.evalsDelSondeo {
					ruta := path.Join(familia.evals, fichero)

					eval, err := fs.Stat(repositorio, ruta)
					if assert.NoError(t, err, "%s es una eval del sondeo de %s", ruta, familia.skill) {
						assert.True(t, eval.Mode().IsRegular(), "%s es un fichero", ruta)
					}
				}
			})
		})
	}

	t.Run("nombres de las cuatro familias", func(t *testing.T) {
		t.Parallel()

		declarada, err := leerLaDeclaracionDelRepositorio(raizDelRepositorio)
		require.NoError(t, err)
		require.NotEmpty(t, declarada.Variantes, "la declaración del laboratorio de H39 tiene variantes")

		deLaFamilia := map[string]string{}

		for _, familia := range slices.Concat([]familiaDelLaboratorio{declarada.familia()}, familiasDelProducto) {
			for _, nombre := range familia.nombres() {
				otra, repetido := deLaFamilia[nombre]
				assert.False(t, repetido, "%s es el nombre de una variante de %s y de una de %s", nombre, otra,
					familia.carpeta)

				deLaFamilia[nombre] = familia.carpeta
			}
		}
	})
}

// comprobarElControlYElProducto: la primera variante de la familia es
// <skill>-control, en los dos modos; la del producto es <skill>-producto, una de
// sus variantes, en los dos modos; y la familia versionada es la misma sin ella
// (data-model §1).
func comprobarElControlYElProducto(t *testing.T, familia familiaDelLaboratorio) {
	t.Helper()

	losDosModos := []Modo{ModoOrden, ModoHerramienta}

	require.NotEmpty(t, familia.variantes, "la familia de %s tiene variantes", familia.skill)
	assert.Equal(t, varianteDelLaboratorio{Nombre: familia.skill + "-control", Modos: losDosModos},
		familia.variantes[0], "la primera variante de la familia de %s es su control, en los dos modos", familia.skill)

	assert.Equal(t, familia.skill+"-producto", familia.producto, "la variante del producto de la familia de %s",
		familia.skill)

	posicion := slices.IndexFunc(familia.variantes, func(variante varianteDelLaboratorio) bool {
		return variante.Nombre == familia.producto
	})
	if assert.GreaterOrEqual(t, posicion, 0, "la variante del producto es una de la familia de %s", familia.skill) {
		assert.Equal(t, losDosModos, familia.variantes[posicion].Modos,
			"la variante del producto de la familia de %s se mide en los dos modos", familia.skill)
	}

	versionada := familia.versionada()
	assert.Equal(t, slices.DeleteFunc(familia.nombres(), func(nombre string) bool { return nombre == familia.producto }),
		versionada.nombres(), "las variantes versionadas de %s son las suyas menos la del producto, en su orden",
		familia.skill)
	assert.Empty(t, versionada.producto, "la familia versionada de %s no tiene variante del producto", familia.skill)
}

// comprobarLosDirectoriosDeLaFamilia: la carpeta de la familia tiene un
// directorio por variante versionada y nada más, y ninguno es el de la variante
// del producto, que la familia declara y no se versiona (FR-022;
// contracts/variantes.md §1 y §4; data-model §1).
func comprobarLosDirectoriosDeLaFamilia(t *testing.T, repositorio fs.FS, familia familiaDelLaboratorio) {
	t.Helper()

	entradas, err := fs.ReadDir(repositorio, familia.carpeta)
	require.NoError(t, err)

	nombres := make([]string, 0, len(entradas))

	for _, entrada := range entradas {
		ruta := path.Join(familia.carpeta, entrada.Name())

		assert.True(t, entrada.IsDir(), "%s no es el directorio de una variante", ruta)
		assert.NotEqual(t, familia.producto, entrada.Name(),
			"%s es la variante del producto, que es skills/%s y no se versiona", ruta, familia.skill)

		nombres = append(nombres, entrada.Name())
	}

	assert.ElementsMatch(t, familia.versionada().nombres(), nombres,
		"%s tiene un directorio por variante versionada de la familia de %s y nada más", familia.carpeta, familia.skill)
}

// comprobarLasHuellasDelControl: los ficheros del directorio de la variante de
// control de la familia son, ruta a ruta, los de sus huellas (FR-020, SC-009;
// research D9); falla con la ruta de cada uno que difiere, falta o sobra.
func comprobarLasHuellasDelControl(t *testing.T, repositorio fs.FS, familia familiaDelLaboratorio) {
	t.Helper()

	esperadas, declaradas := huellasDeLosControles[familia.skill]
	require.True(t, declaradas, "hay huellas del control de la familia de %s", familia.skill)

	control := familia.directorioDeLaVariante(familia.nombres()[0])

	ficheros, err := leerLosFicheros(repositorio, control)
	require.NoError(t, err)

	obtenidas := make(map[string]string, len(ficheros))
	for relativa, contenido := range ficheros {
		obtenidas[relativa] = huellaDelFichero(contenido)
	}

	rutas := slices.Concat(slices.Collect(maps.Keys(esperadas)), slices.Collect(maps.Keys(obtenidas)))
	slices.Sort(rutas)

	for _, relativa := range slices.Compact(rutas) {
		esperada, laEspera := esperadas[relativa]
		obtenida, laTiene := obtenidas[relativa]
		ruta := path.Join(control, relativa)

		switch {
		case !laTiene:
			assert.Fail(t, "falta", "%s falta en la variante de control de %s", ruta, familia.skill)
		case !laEspera:
			assert.Fail(t, "sobra", "%s sobra en la variante de control de %s", ruta, familia.skill)
		case esperada != obtenida:
			assert.Fail(t, "difiere", "%s difiere: su huella es %s y la de la copia de v0.9.0, %s", ruta, obtenida,
				esperada)
		}
	}
}

// huellaDelFichero es la huella SHA-256 de esos bytes, en hexadecimal, la de
// shasum -a 256.
func huellaDelFichero(contenido []byte) string {
	suma := sha256.Sum256(contenido)

	return hex.EncodeToString(suma[:])
}

// comprobarLasReferenciasDeLaFamilia: las referencias de la familia son los
// ficheros de references/ de su variante de control, en orden de nombre
// (data-model §1).
func comprobarLasReferenciasDeLaFamilia(t *testing.T, repositorio fs.FS, familia familiaDelLaboratorio) {
	t.Helper()

	referencias := path.Join(familia.directorioDeLaSkill(familia.nombres()[0]), directorioDeLasReferencias)

	entradas, err := fs.ReadDir(repositorio, referencias)
	require.NoError(t, err)

	ficheros := make([]string, 0, len(entradas))
	for _, entrada := range entradas {
		ficheros = append(ficheros, entrada.Name())
	}

	assert.Equal(t, ficheros, familia.referencias,
		"las referencias de la familia de %s son los ficheros de %s, en orden de nombre", familia.skill, referencias)
}
