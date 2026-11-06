package evals

import (
	"crypto/sha256"
	"encoding/hex"
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

// Lo que genera make skills-sync en una skill del producto y su control, que
// TestSkillsDelProductoDesdeControl no compara (contracts/remisiones.md §8.1):
// las referencias cuya primera línea empieza por la cabecera de lo generado, y,
// en el SKILL.md, lo que hay entre las marcas de la región de los paquetes y
// entre las de la tabla de comandos. Que sea lo que da make skills-sync lo
// comprueba TestSkillsDelRepositorio (FR-009).
const (
	cabeceraDeLoGenerado = "<!-- generado desde "
	inicioDeLosPaquetes  = "<!-- paquetes: "
	finDeLosPaquetes     = "<!-- fin de los paquetes -->"
)

// regionesGeneradas: cada región del SKILL.md que genera make skills-sync, de la
// línea que empieza por su marca de inicio a la que es su marca de fin.
var regionesGeneradas = []struct{ inicio, fin string }{
	{inicio: inicioDeLosPaquetes, fin: finDeLosPaquetes},
	{inicio: inicioDeLaRegionGenerada, fin: finDeLaRegionGenerada},
}

// cambioDelCuerpo es un cambio del cuerpo del SKILL.md de una skill del producto
// respecto al de su control: las líneas enteras de control que cambian, antes,
// que están en él una sola vez, sin lo generado, y las que las sustituyen tras el
// hito, despues (data-model §5).
type cambioDelCuerpo struct {
	antes, despues string
}

// movimientoALaReferencia es un texto que sale del SKILL.md de control, texto, y
// llega al final de una de sus referencias, referencia, con los bytes que se le
// añaden, enLaReferencia: el mismo texto con los blancos juntados, detrás de una
// línea en blanco (data-model §5; contracts/remisiones.md §3).
type movimientoALaReferencia struct {
	texto, referencia, enLaReferencia string
}

// declaracionDelProducto es lo que una skill del producto cambia respecto a su
// control: los cambios de su cuerpo, en el orden del fichero, y sus movimientos
// (data-model §5). Nada más cambia: ni el frontmatter, ni lo generado, ni
// ninguna otra referencia (FR-007, FR-009).
type declaracionDelProducto struct {
	cambios     []cambioDelCuerpo
	movimientos []movimientoALaReferencia
}

// nadaDeOtraConversacion es la segunda viñeta del paso 4 del SKILL.md de control
// de normativa, que sale del cuerpo a redaccion_modificada.md (M-a;
// contracts/remisiones.md §3).
const nadaDeOtraConversacion = "- **Nada de otra conversación**, ni para afirmar, confirmar o desmentir: de ellas " +
	"solo va la línea\n" +
	"  `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae.\n"

// siUnaLecturaFalla es la frase de la regla 2 del SKILL.md de control de
// normativa que sale del cuerpo a lecturas_que_fallan.md (M-b;
// contracts/remisiones.md §3).
const siUnaLecturaFalla = "Si una lectura falla, di qué bloques faltan y por qué."

// declaracionesDelProducto son, por skill, los cambios y los movimientos de las
// skills del producto respecto a sus controles, con los bytes de
// contracts/remisiones.md §2 a §4. plazos-y-territorio no tiene: es su control,
// también lo generado (FR-010).
var declaracionesDelProducto = map[string]declaracionDelProducto{
	// Las siete remisiones de R1 a R6, con la «o» de R4 (research D1), y la
	// salida de M-a y M-b (contracts/remisiones.md §2 y §3).
	skillDeNormativa: {
		cambios: []cambioDelCuerpo{
			{
				antes: "Sigue los pasos en orden, cada operación con `kitlegal … --json` desde el `PATH` o, si la " +
					"tienes, con su\n" +
					"herramienta, tras leer [con_herramientas](references/con_herramientas.md).\n",
				despues: "Antes de la primera orden o llamada, lee [con_herramientas](references/con_herramientas.md): " +
					"cómo se pide cada\n" +
					"una con herramienta. Sigue los pasos en orden, cada operación con `kitlegal … --json` desde el " +
					"`PATH` o, si la\n" +
					"tienes, con su herramienta.\n",
			},
			{
				antes: "Toma el `BOE-A-…` de la tabla de [normas](references/normas.md); si no está, búscala con\n" +
					"[buscar_la_norma](references/buscar_la_norma.md).\n",
				despues: "Lee [normas](references/normas.md) y toma el `BOE-A-…` de su tabla; si no está, lee\n" +
					"[buscar_la_norma](references/buscar_la_norma.md) y búscala como dice.\n",
			},
			{
				antes: "- Sin fecha —con ella, [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md)—, lee " +
					"cada bloque una vez, con\n" +
					"  su comprobación detrás, en la misma orden:\n",
				despues: "- Con una fecha en la pregunta, lee " +
					"[redaccion_en_una_fecha](references/redaccion_en_una_fecha.md) antes de leer el\n" +
					"  bloque: cómo se lee. Sin fecha, lee cada bloque una vez, con su comprobación detrás, en la " +
					"misma orden:\n",
			},
			{
				antes: "  fecha van siempre.\n" +
					"- Varios bloques a la vez: [varios_bloques](references/varios_bloques.md). Si una orden no " +
					"termina con `0`, antes de\n" +
					"  seguir: [lecturas_que_fallan](references/lecturas_que_fallan.md).\n",
				despues: "  fecha van siempre. Por cada entrada `version-obsoleta`, o si se pregunta qué cambió, lee\n" +
					"  [redaccion_modificada](references/redaccion_modificada.md): qué se dice.\n" +
					"- Con avisos o un id que termina en punto, lee [como_se_cita](references/como_se_cita.md): " +
					"cómo se citan.\n" +
					"- Para varios bloques a la vez, lee antes [varios_bloques](references/varios_bloques.md): su " +
					"orden. Si una orden no\n" +
					"  termina con `0`, lee [lecturas_que_fallan](references/lecturas_que_fallan.md) antes de " +
					"seguir: qué hacer.\n",
			},
			{antes: nadaDeOtraConversacion},
			{
				antes: "más: `art. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]`. Una cita por " +
					"bloque, también del remitido;\n" +
					"repásalas, y las dudas, en [como_se_cita](references/como_se_cita.md). Cada línea `⚠` va " +
					"entera y con los datos de la lectura.\n",
				despues: "más: `art. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]`. Una cita por " +
					"bloque, también del remitido.\n" +
					"Cada línea `⚠` va entera y con los datos de la lectura.\n",
			},
			{
				antes: "Por cada entrada `version-obsoleta` de `kitlegal graph check`, y si se pregunta qué " +
					"cambió, lee\n" +
					"[redaccion_modificada](references/redaccion_modificada.md):\n" +
					"\n",
			},
			{
				antes:   "2. **Nunca inventar contenido legal.** " + siUnaLecturaFalla + " Lo que lo leído no da\n",
				despues: "2. **Nunca inventar contenido legal.** Lo que lo leído no da\n",
			},
		},
		movimientos: []movimientoALaReferencia{
			{
				texto:      nadaDeOtraConversacion,
				referencia: "redaccion_modificada.md",
				enLaReferencia: "\n" +
					"- **Nada de otra conversación**, ni para afirmar, confirmar o desmentir: de ellas solo va " +
					"la\n" +
					"  línea `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae.\n",
			},
			{
				texto:          siUnaLecturaFalla,
				referencia:     "lecturas_que_fallan.md",
				enLaReferencia: "\nSi una lectura falla, di qué bloques faltan y por qué.\n",
			},
		},
	},
	// jurisprudencia, sin cambios ni movimientos: hoy es su control.
	skillDeLaFamiliaDeJurisprudencia: {},
}

// TestSkillsDelProductoDesdeControl es el control de lo que cambian las skills
// del producto respecto a sus controles (FR-006, FR-007, FR-009, FR-010; SC-003
// a SC-005; contracts/remisiones.md §8; data-model §5): en las de una
// declaración, su control sin lo generado, con sus cambios y sus movimientos,
// es skills/<skill> sin lo generado, ruta a ruta y byte a byte; cada texto
// movido está en su referencia y no en el SKILL.md; y el frontmatter es el de
// control. plazos-y-territorio es su control, también lo generado.
func TestSkillsDelProductoDesdeControl(t *testing.T) {
	t.Parallel()

	repositorio := os.DirFS(raizDelRepositorio)

	for _, skill := range slices.Sorted(maps.Keys(declaracionesDelProducto)) {
		t.Run(skill, func(t *testing.T) {
			t.Parallel()

			comprobarElProductoDesdeControl(t, repositorio, skill, declaracionesDelProducto[skill])
		})
	}

	// Es así hasta H41.1, el hito que cambia plazos-y-territorio: hasta
	// entonces, skills/plazos-y-territorio es su control, también lo generado
	// (FR-010; contracts/remisiones.md §8.4; research D4).
	t.Run(skillDeLaFamiliaDePlazos, func(t *testing.T) {
		t.Parallel()

		producto := path.Join(directorioDeLasSkills, skillDeLaFamiliaDePlazos)

		deControl, err := leerLosFicheros(repositorio, directorioDelControl(t, skillDeLaFamiliaDePlazos))
		require.NoError(t, err)

		delProducto, err := leerLosFicheros(repositorio, producto)
		require.NoError(t, err)

		compararLosFicheros(t, deControl, delProducto, producto, "su control")
	})
}

// comprobarElProductoDesdeControl comprueba skills/<skill> contra su control y
// su declaración (contracts/remisiones.md §8.1 a §8.3).
func comprobarElProductoDesdeControl(t *testing.T, repositorio fs.FS, skill string,
	declarada declaracionDelProducto,
) {
	t.Helper()

	producto := path.Join(directorioDeLasSkills, skill)

	deControl, err := leerLosFicheros(repositorio, directorioDelControl(t, skill))
	require.NoError(t, err)

	delProducto, err := leerLosFicheros(repositorio, producto)
	require.NoError(t, err)

	t.Run("desde control", func(t *testing.T) {
		t.Parallel()

		controlSinLoGenerado, err := sinLoGenerado(deControl)
		require.NoError(t, err, "el control de %s", skill)

		esperados, err := conLaDeclaracion(controlSinLoGenerado, declarada)
		require.NoError(t, err, "la declaración de %s", skill)

		obtenidos, err := sinLoGenerado(delProducto)
		require.NoError(t, err, "%s", producto)

		compararLosFicheros(t, esperados, obtenidos, producto, "lo que dan su control y su declaración")
	})

	t.Run("movimientos", func(t *testing.T) {
		t.Parallel()

		cuerpoDeControl := juntarLosBlancos(string(deControl[ficheroDeLaSkill]))
		cuerpoDelProducto := juntarLosBlancos(string(delProducto[ficheroDeLaSkill]))

		for _, movimiento := range declarada.movimientos {
			texto := juntarLosBlancos(movimiento.texto)
			relativa := path.Join(directorioDeLasReferencias, movimiento.referencia)
			referencia := path.Join(producto, relativa)

			assert.Equal(t, texto, juntarLosBlancos(movimiento.enLaReferencia),
				"lo que llega a %s es, con los blancos juntados, lo que sale del SKILL.md", referencia)
			assert.Contains(t, cuerpoDeControl, texto, "lo que llega a %s sale del SKILL.md de control", referencia)
			assert.NotContains(t, cuerpoDelProducto, texto, "lo que llega a %s no está en %s/%s", referencia,
				producto, ficheroDeLaSkill)
			assert.Contains(t, string(delProducto[relativa]), movimiento.enLaReferencia,
				"lo que sale del SKILL.md está en %s", referencia)
		}
	})

	t.Run("frontmatter", func(t *testing.T) {
		t.Parallel()

		deLaSkillDeControl, _, err := frontmatterYCuerpo(string(deControl[ficheroDeLaSkill]))
		require.NoError(t, err, "el SKILL.md de control de %s", skill)

		deLaSkillDelProducto, _, err := frontmatterYCuerpo(string(delProducto[ficheroDeLaSkill]))
		require.NoError(t, err, "%s/%s", producto, ficheroDeLaSkill)

		assert.Equal(t, deLaSkillDeControl, deLaSkillDelProducto, "el frontmatter de %s/%s es el de su control",
			producto, ficheroDeLaSkill)
	})
}

// directorioDelControl es el directorio de la skill de la variante de control de
// la familia del producto de esa skill, desde la raíz del repositorio.
func directorioDelControl(t *testing.T, skill string) string {
	t.Helper()

	posicion := slices.IndexFunc(familiasDelProducto, func(familia familiaDelLaboratorio) bool {
		return familia.skill == skill
	})
	require.GreaterOrEqual(t, posicion, 0, "%s es la skill de una familia del producto", skill)

	familia := familiasDelProducto[posicion]

	return familia.directorioDeLaSkill(familia.nombres()[0])
}

// sinLoGenerado son los ficheros de una skill, por su ruta relativa a ella, sin
// lo que genera make skills-sync: sin las referencias cuya primera línea empieza
// por la cabecera de lo generado y con cada región generada del SKILL.md
// reducida a sus marcas (contracts/remisiones.md §8.1).
func sinLoGenerado(ficheros map[string][]byte) (map[string][]byte, error) {
	sinGenerar := make(map[string][]byte, len(ficheros))

	for relativa, contenido := range ficheros {
		switch {
		case relativa == ficheroDeLaSkill:
			reducido, err := conLasRegionesGeneradasReducidas(string(contenido))
			if err != nil {
				return nil, fmt.Errorf("%s: %w", relativa, err)
			}

			sinGenerar[relativa] = []byte(reducido)
		case strings.HasPrefix(relativa, directorioDeLasReferencias+"/") &&
			strings.HasPrefix(string(contenido), cabeceraDeLoGenerado):
			// Una referencia generada: no se compara.
		default:
			sinGenerar[relativa] = contenido
		}
	}

	return sinGenerar, nil
}

// conLasRegionesGeneradasReducidas es el SKILL.md con cada región generada
// reducida a sus marcas: la línea que empieza por la de inicio y la que es la de
// fin, sin nada entre ellas.
func conLasRegionesGeneradasReducidas(skill string) (string, error) {
	var reducido strings.Builder

	// fin es la marca de fin de la región abierta, o nada fuera de una.
	fin := ""

	for linea := range strings.Lines(skill) {
		recortada := strings.TrimSuffix(linea, "\n")

		switch {
		case fin == "":
			for _, region := range regionesGeneradas {
				if strings.HasPrefix(recortada, region.inicio) {
					fin = region.fin
				}
			}

			reducido.WriteString(linea)
		case recortada == fin:
			fin = ""

			reducido.WriteString(linea)
		}
	}

	if fin != "" {
		return "", fmt.Errorf("no tiene la línea %q que cierra una región generada", fin)
	}

	return reducido.String(), nil
}

// conLaDeclaracion son los ficheros de control, sin lo generado, con los cambios
// de la declaración aplicados a su SKILL.md, en su orden, y los bytes de cada
// movimiento añadidos al final de su referencia. El antes de cada cambio está una
// sola vez en el SKILL.md, con los anteriores ya aplicados; si no, es un error
// con su número.
func conLaDeclaracion(control map[string][]byte, declarada declaracionDelProducto) (map[string][]byte, error) {
	esperados := maps.Clone(control)

	skill, laTiene := esperados[ficheroDeLaSkill]
	if !laTiene {
		return nil, fmt.Errorf("el control no tiene %s", ficheroDeLaSkill)
	}

	cuerpo := string(skill)

	for numero, cambio := range declarada.cambios {
		if veces := strings.Count(cuerpo, cambio.antes); veces != 1 {
			return nil, fmt.Errorf("el antes del cambio %d está %d veces en el %s de control, y tiene que estar "+
				"una: %q", numero+1, veces, ficheroDeLaSkill, cambio.antes)
		}

		cuerpo = strings.Replace(cuerpo, cambio.antes, cambio.despues, 1)
	}

	esperados[ficheroDeLaSkill] = []byte(cuerpo)

	for _, movimiento := range declarada.movimientos {
		relativa := path.Join(directorioDeLasReferencias, movimiento.referencia)

		referencia, laTiene := esperados[relativa]
		if !laTiene {
			return nil, fmt.Errorf("el control no tiene %s, a la que llega un movimiento, entre sus referencias "+
				"a mano", relativa)
		}

		esperados[relativa] = slices.Concat(referencia, []byte(movimiento.enLaReferencia))
	}

	return esperados, nil
}

// compararLosFicheros exige que los ficheros obtenidos de ese directorio sean,
// ruta a ruta y byte a byte, los esperados, que son lo que dice queEs; falla con
// la ruta del primero, en orden de ruta, que difiere, falta o sobra.
func compararLosFicheros(t *testing.T, esperados, obtenidos map[string][]byte, directorio, queEs string) {
	t.Helper()

	rutas := slices.Concat(slices.Collect(maps.Keys(esperados)), slices.Collect(maps.Keys(obtenidos)))
	slices.Sort(rutas)

	for _, relativa := range slices.Compact(rutas) {
		esperado, loEspera := esperados[relativa]
		obtenido, loTiene := obtenidos[relativa]
		ruta := path.Join(directorio, relativa)

		switch {
		case !loTiene:
			assert.Fail(t, "falta", "%s falta: está en %s", ruta, queEs)

			return
		case !loEspera:
			assert.Fail(t, "sobra", "%s sobra: no está en %s", ruta, queEs)

			return
		case !assert.Equal(t, string(esperado), string(obtenido), "%s difiere de %s", ruta, queEs):
			return
		}
	}
}

// juntarLosBlancos es el texto con cada tramo de blancos —espacios, saltos de
// línea— cambiado por un espacio, y sin blancos al principio ni al final.
func juntarLosBlancos(texto string) string {
	return strings.Join(strings.Fields(texto), " ")
}
