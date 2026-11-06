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

// familiaDeNormativa es la familia de normativa con sus variantes versionadas,
// las de H40, sin la del producto, con la que los tests de H40 miran las
// variantes de normativa del repositorio, cada una con su directorio
// (contracts/variantes.md de H40).
var familiaDeNormativa = familiasDelProducto[0].versionada()

// huellasDeLosControles son las huellas SHA-256 de los ficheros de la variante de
// control de cada familia del producto, por su skill y por su ruta desde el
// directorio de la variante, escritas al copiarlos de skills/ en v0.9.0 (research
// D8 y V1; data-model §2): 9 de normativa, 5 de plazos-y-territorio y 3 de
// jurisprudencia. No se comparan con skills/<skill>, que cambia con el producto
// (FR-020), ni con la etiqueta, que el clon superficial del make ci de la entrega
// no tiene.
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
// familias del producto (FR-020, FR-022, SC-006 y SC-008; contracts/variantes.md
// §1 y §2; data-model §1 y §2; research D8 y D9): son normativa,
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
// contracts/variantes.md §2).
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
// control de la familia son, ruta a ruta, los de sus huellas (FR-020, SC-008;
// research D8); falla con la ruta de cada uno que difiere, falta o sobra.
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

	rutas := slices.Sorted(maps.Keys(esperadas))
	rutas = append(rutas, slices.Collect(maps.Keys(obtenidas))...)
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

// Lo que genera make skills-sync en una skill, que la comparación con control
// deja fuera (research D7; contracts/remisiones.md §7): las referencias que
// empiezan por la cabecera de lo generado y, en el SKILL.md, la región de los
// paquetes y la de la tabla de comandos (inicioDeLaRegionGenerada y
// finDeLaRegionGenerada), cada una de la línea que empieza por su marca de inicio
// a la que es su marca de fin.
const (
	cabeceraDeLoGenerado = "<!-- generado desde "
	inicioDeLosPaquetes  = "<!-- paquetes: "
	finDeLosPaquetes     = "<!-- fin de los paquetes -->"
)

// cambioDelCuerpo es una diferencia declarada del SKILL.md de una skill del
// producto sobre el de su control, sin lo generado (data-model §5): antes está una
// sola vez en el de control; despues son los bytes que lo sustituyen, vacíos si
// el texto sale del cuerpo.
type cambioDelCuerpo struct {
	antes   string
	despues string
}

// movimientoALaReferencia es un texto que sale del SKILL.md de control —el
// «antes» de un cambio con «después» vacío— y llega al final de una referencia
// (data-model §5): referencia es el nombre de su fichero en references/, y
// enLaReferencia, los bytes que se le añaden, el mismo texto con los blancos
// juntados y los enlaces relativos a la referencia.
type movimientoALaReferencia struct {
	texto          string
	referencia     string
	enLaReferencia string
}

// reorganizacionDelProducto es lo que se aplica a <skill>-control para calcular
// skills/<skill> sin lo generado (data-model §5); lo que no declara es el de
// control.
type reorganizacionDelProducto struct {
	cambios     []cambioDelCuerpo
	movimientos []movimientoALaReferencia
}

// Lo que sale del SKILL.md de normativa-control a sus referencias
// (contracts/remisiones.md §3; research D2): la viñeta del paso 4 que pasa al
// final de references/redaccion_modificada.md, y la frase de la regla 2 que pasa
// al final de references/lecturas_que_fallan.md (gates/supuestos.md, T003).
const (
	nadaDeOtraConversacion = "- **Nada de otra conversación**, ni para afirmar, confirmar o desmentir: de ellas " +
		"solo va la línea\n" +
		"  `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae.\n"
	siUnaLecturaFalla = " Si una lectura falla, di qué bloques faltan y por qué."
)

// siLaLecturaSaleCon4o5 es la frase del paso 10 de plazos-y-territorio-control que
// pasa al final de references/calculo_de_plazos.md, que ya lo dice en «La lectura en
// la fecha de inicio» (gates/supuestos.md, T004).
const siLaLecturaSaleCon4o5 = " Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: di aparte que no has podido leerlo, " +
	"sin\n" +
	"    afirmar qué dice."

// reorganizacionesDelProducto son las de las familias del producto, por su skill,
// con los bytes de contracts/remisiones.md §2 a §5: la de normativa, sus
// remisiones R1 a R6 y sus dos movimientos; la de plazos-y-territorio, sus
// remisiones P1 a P3 y su movimiento; y la de jurisprudencia, sus remisiones J1 y
// J2.
var reorganizacionesDelProducto = map[string]reorganizacionDelProducto{
	skillDeNormativa: {
		cambios: []cambioDelCuerpo{
			{ // R1
				antes: "Sigue los pasos en orden, cada operación con `kitlegal … --json` desde el `PATH` o, si la " +
					"tienes, con su\n" +
					"herramienta, tras leer [con_herramientas](references/con_herramientas.md).\n",
				despues: "Antes de la primera orden o llamada, lee [con_herramientas](references/con_herramientas.md): " +
					"cómo se pide cada\n" +
					"una con herramienta. Sigue los pasos en orden, cada operación con `kitlegal … --json` desde el " +
					"`PATH` o, si la\n" +
					"tienes, con su herramienta.\n",
			},
			{ // R2
				antes: "Toma el `BOE-A-…` de la tabla de [normas](references/normas.md); si no está, búscala con\n" +
					"[buscar_la_norma](references/buscar_la_norma.md).\n",
				despues: "Lee [normas](references/normas.md) y toma el `BOE-A-…` de su tabla; si no está, lee\n" +
					"[buscar_la_norma](references/buscar_la_norma.md) y búscala como dice.\n",
			},
			{ // R3
				antes: "- Sin fecha —con ella, [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md)—, lee cada " +
					"bloque una vez, con\n" +
					"  su comprobación detrás, en la misma orden:\n",
				despues: "- Con una fecha en la pregunta, lee [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md) " +
					"antes de leer el\n" +
					"  bloque: cómo se lee. Sin fecha, lee cada bloque una vez, con su comprobación detrás, en la misma " +
					"orden:\n",
			},
			{ // R4, R5 y R6, en el paso 2
				antes: "  fecha van siempre.\n" +
					"- Varios bloques a la vez: [varios_bloques](references/varios_bloques.md). Si una orden no termina " +
					"con `0`, antes de\n" +
					"  seguir: [lecturas_que_fallan](references/lecturas_que_fallan.md).\n",
				despues: "  fecha van siempre. Por cada entrada `version-obsoleta`, y si se pregunta qué cambió, lee\n" +
					"  [redaccion_modificada](references/redaccion_modificada.md): qué se dice.\n" +
					"- Con avisos o un id que termina en punto, lee [como_se_cita](references/como_se_cita.md): cómo se " +
					"citan.\n" +
					"- Para varios bloques a la vez, lee antes [varios_bloques](references/varios_bloques.md): su orden. " +
					"Si una orden no\n" +
					"  termina con `0`, lee [lecturas_que_fallan](references/lecturas_que_fallan.md) antes de seguir: " +
					"qué hacer.\n",
			},
			{ // el movimiento de §3, que sale del paso 4
				antes:   nadaDeOtraConversacion,
				despues: "",
			},
			{ // R5, que sale de «Cómo se cita»
				antes: "más: `art. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]`. Una cita por " +
					"bloque, también del remitido;\n" +
					"repásalas, y las dudas, en [como_se_cita](references/como_se_cita.md). Cada línea `⚠` va entera y " +
					"con los datos de la lectura.\n",
				despues: "más: `art. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]`. Una cita por " +
					"bloque, también del remitido.\n" +
					"Cada línea `⚠` va entera y con los datos de la lectura.\n",
			},
			{ // R4, que sale de «Redacción modificada»
				antes: "Por cada entrada `version-obsoleta` de `kitlegal graph check`, y si se pregunta qué cambió, lee\n" +
					"[redaccion_modificada](references/redaccion_modificada.md):\n" +
					"\n",
				despues: "",
			},
			{ // el segundo movimiento, que sale de la regla 2
				antes:   siUnaLecturaFalla,
				despues: "",
			},
		},
		movimientos: []movimientoALaReferencia{
			{
				// Partida tras «la» y no tras «línea»: ninguna línea de una
				// referencia a mano empieza por ⚠ sin ser una de SKILL.md (control
				// 7 de H31; FR-009 deja cambiar los saltos de línea).
				texto:      nadaDeOtraConversacion,
				referencia: "redaccion_modificada.md",
				enLaReferencia: "\n" +
					"- **Nada de otra conversación**, ni para afirmar, confirmar o desmentir: de ellas solo va la\n" +
					"  línea `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae.\n",
			},
			{
				texto:      siUnaLecturaFalla,
				referencia: "lecturas_que_fallan.md",
				enLaReferencia: "\n" +
					"Si una lectura falla, di qué bloques faltan y por qué.\n",
			},
		},
	},
	skillDeLaFamiliaDePlazos: {
		cambios: []cambioDelCuerpo{
			{ // P1
				antes: "- [jerarquia_normativa.md](references/jerarquia_normativa.md): para el nivel que regula algo, " +
					"con el autonómico de la\n" +
					"  `comunidad` de `data`, y la norma que prevalece.\n",
				despues: "- Para el nivel que regula algo, con el autonómico de la `comunidad` de `data`, y la norma que " +
					"prevalece, lee\n" +
					"  [jerarquia_normativa.md](references/jerarquia_normativa.md): rangos y reglas.\n",
			},
			{ // P2
				antes: "- [municipios.md](references/municipios.md): al presentar un territorio, si falta el municipio o " +
					"no sale con 0, y\n" +
					"  para explicar su cobertura.\n",
				despues: "- Al presentar un territorio, si falta el municipio o no sale con 0, y para explicar su " +
					"cobertura, lee\n" +
					"  [municipios.md](references/municipios.md): qué hacer.\n",
			},
			{ // P3
				antes: "- [calculo_de_plazos.md](references/calculo_de_plazos.md): en los municipios que cuentan, un " +
					"plazo sin calcular y una\n" +
					"  lectura fallida o con `redaccion_posterior`.\n",
				despues: "- En los municipios que cuentan, un plazo sin calcular y una lectura fallida o con " +
					"`redaccion_posterior`, lee\n" +
					"  [calculo_de_plazos.md](references/calculo_de_plazos.md): su detalle.\n",
			},
			{ // el movimiento, que sale del paso 10
				antes:   siLaLecturaSaleCon4o5,
				despues: "",
			},
		},
		movimientos: []movimientoALaReferencia{
			{
				texto:      siLaLecturaSaleCon4o5,
				referencia: "calculo_de_plazos.md",
				enLaReferencia: "\n" +
					"Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: di aparte que no has podido leerlo, sin afirmar " +
					"qué dice.\n",
			},
		},
	},
	skillDeLaFamiliaDeJurisprudencia: {
		cambios: []cambioDelCuerpo{
			{ // J1, que entra en el paso 2 antes de la tabla de lo que se pide
				antes: "Y debajo, la consulta para encontrarla, que sale de `cita_preparar` o de su orden. Lo que se " +
					"pide depende de cómo se\n" +
					"dio la referencia:\n",
				despues: "Y debajo, la consulta para encontrarla, que sale de `cita_preparar` o de su orden. Antes de " +
					"pedir una u otra, lee\n" +
					"[references/consulta_preparada.md](references/consulta_preparada.md): la orden de cada forma, " +
					"cómo se dan las casillas\n" +
					"y qué hacer con el equivalente que devuelva la operación. Lo que se pide depende de cómo se dio " +
					"la referencia:\n",
			},
			{ // J1, que sale del final del paso 2
				antes: "Antes de dar la consulta, lee [references/consulta_preparada.md](references/consulta_preparada.md): " +
					"la orden de cada\n" +
					"forma, cómo se dan las casillas y qué hacer con el equivalente que devuelva la operación.\n" +
					"\n",
				despues: "",
			},
			{ // J2, que entra al principio del paso 3
				antes: "### 3. Con el documento delante: cotejar su ficha\n" +
					"\n",
				despues: "### 3. Con el documento delante: cotejar su ficha\n" +
					"\n" +
					"Con cada documento, antes del cotejo, lee " +
					"[references/documento_aportado.md](references/documento_aportado.md): qué\n" +
					"se pasa del documento y cómo, y qué se dice si no es la pedida, si su ROJ y su ECLI no se " +
					"corresponden o si no lleva\n" +
					"ficha.\n" +
					"\n",
			},
			{ // J2, que sale del final del paso 3
				antes: "Con cada documento, antes del cotejo, lee " +
					"[references/documento_aportado.md](references/documento_aportado.md): qué\n" +
					"se pasa del documento y cómo, y qué se dice si no es la pedida, si su ROJ y su ECLI no se " +
					"corresponden o si no lleva\n" +
					"ficha.\n" +
					"\n",
				despues: "",
			},
		},
	},
}

// TestSkillsDelProductoDesdeControl es el control de la reorganización de las
// skills del producto (FR-007, FR-009 y FR-010; SC-003 y SC-004;
// contracts/remisiones.md §7; data-model §5; research D7): por familia del
// producto, skills/<skill>, sin lo que genera make skills-sync, es, ruta a ruta y
// byte a byte, lo que da su declaración sobre <skill>-control, sin lo generado;
// cada texto movido es el que llega a su referencia y ya no está en el SKILL.md;
// y el frontmatter del SKILL.md es el de control. Que lo generado sea lo que da
// make skills-sync lo comprueba TestSkillsDelRepositorio.
func TestSkillsDelProductoDesdeControl(t *testing.T) {
	t.Parallel()

	repositorio := os.DirFS(raizDelRepositorio)

	t.Run("declaraciones", func(t *testing.T) {
		t.Parallel()

		skills := make([]string, 0, len(familiasDelProducto))
		for _, familia := range familiasDelProducto {
			skills = append(skills, familia.skill)
		}

		assert.ElementsMatch(t, skills, slices.Collect(maps.Keys(reorganizacionesDelProducto)),
			"hay una declaración de la reorganización de cada familia del producto y de ninguna otra")
	})

	for _, familia := range familiasDelProducto {
		t.Run(familia.skill, func(t *testing.T) {
			t.Parallel()

			declarada, existe := reorganizacionesDelProducto[familia.skill]
			require.True(t, existe, "hay una declaración de la reorganización de %s", familia.skill)

			deControl := familia.nombres()[0]

			control, err := leerLosFicheros(repositorio, familia.directorioDeLaSkill(deControl))
			require.NoError(t, err)

			producto := path.Join(directorioDeLasSkills, familia.skill)

			delProducto, err := leerLosFicheros(repositorio, producto)
			require.NoError(t, err)

			t.Run("calculado desde control", func(t *testing.T) {
				t.Parallel()

				comprobarElProductoCalculado(t, producto, control, delProducto, declarada)
			})

			t.Run("movimientos", func(t *testing.T) {
				t.Parallel()

				comprobarLosMovimientos(t, producto, delProducto, declarada)
			})

			t.Run("frontmatter", func(t *testing.T) {
				t.Parallel()

				esperado, _, err := frontmatterYCuerpo(string(control[ficheroDeLaSkillDeLaVariante]))
				require.NoError(t, err, "el %s de %s", ficheroDeLaSkillDeLaVariante, deControl)

				ruta := path.Join(producto, ficheroDeLaSkillDeLaVariante)

				obtenido, _, err := frontmatterYCuerpo(string(delProducto[ficheroDeLaSkillDeLaVariante]))
				require.NoError(t, err, ruta)

				assert.Equal(t, esperado, obtenido, "%s: el frontmatter es el de %s, byte a byte", ruta, deControl)
			})
		})
	}
}

// comprobarElProductoCalculado: los ficheros de la skill del producto, sin lo
// generado, son, ruta a ruta y byte a byte, los que da la declaración sobre los
// de control, sin lo generado; falla con el cambio cuyo «antes» ya no está una
// vez en el SKILL.md de control, o con la ruta del primer fichero, en orden, que
// difiere, falta o sobra.
func comprobarElProductoCalculado(t *testing.T, producto string, control, delProducto map[string][]byte,
	declarada reorganizacionDelProducto,
) {
	t.Helper()

	calculados, err := calcularElProducto(control, declarada)
	require.NoError(t, err, "%s: su declaración ya no se aplica sobre su control", producto)

	aMano, err := sinLoGenerado(delProducto)
	require.NoError(t, err, producto)

	esperados := make(map[string][]byte, len(calculados))
	for relativa, contenido := range calculados {
		esperados[path.Join(producto, relativa)] = contenido
	}

	obtenidos := make(map[string][]byte, len(aMano))
	for relativa, contenido := range aMano {
		obtenidos[path.Join(producto, relativa)] = contenido
	}

	ruta, motivo := primeraDiferencia(esperados, obtenidos)
	assert.Empty(t, motivo, "%s %s: la skill, sin lo que genera make skills-sync, es la que da su declaración sobre "+
		"su control, ruta a ruta y byte a byte", ruta, motivo)
}

// comprobarLosMovimientos: por cada movimiento de la declaración, su texto es el
// «antes» de un cambio con «después» vacío; su texto y lo que llega a la
// referencia, con los blancos juntados y los enlaces relativos a la referencia,
// son el mismo; y el texto, con los blancos juntados, no está en el SKILL.md de la
// skill del producto (FR-009).
func comprobarLosMovimientos(t *testing.T, producto string, delProducto map[string][]byte,
	declarada reorganizacionDelProducto,
) {
	t.Helper()

	ruta := path.Join(producto, ficheroDeLaSkillDeLaVariante)
	skill := juntarLosBlancos(string(delProducto[ficheroDeLaSkillDeLaVariante]))

	for indice, movimiento := range declarada.movimientos {
		assert.Contains(t, declarada.cambios, cambioDelCuerpo{antes: movimiento.texto},
			"%s: el texto del movimiento %d es el «antes» de un cambio con «después» vacío", producto, indice+1)
		assert.Equal(t, textoDelMovimiento(movimiento.texto), textoDelMovimiento(movimiento.enLaReferencia),
			"%s: el movimiento %d lleva a references/%s su mismo texto, con los blancos juntados y los enlaces "+
				"relativos a la referencia", producto, indice+1, movimiento.referencia)
		assert.NotContains(t, skill, juntarLosBlancos(movimiento.texto),
			"%s: el texto del movimiento %d, que va a references/%s, sigue en el SKILL.md", ruta, indice+1,
			movimiento.referencia)
	}
}

// calcularElProducto son los ficheros de la skill del producto, sin lo generado,
// que da la declaración sobre los de control, sin lo generado, por su ruta
// relativa a la skill (contracts/remisiones.md §7): cada cambio, en el SKILL.md, y
// lo que llega con cada movimiento, al final de su referencia. Un cambio cuyo
// «antes» no está una sola vez en el SKILL.md de control, o un movimiento a una
// referencia que control no escribe a mano, es un error con su número.
func calcularElProducto(control map[string][]byte, declarada reorganizacionDelProducto) (map[string][]byte, error) {
	calculados, err := sinLoGenerado(control)
	if err != nil {
		return nil, err
	}

	skill := string(calculados[ficheroDeLaSkillDeLaVariante])

	sustituciones := make([]sustitucionDeLaVariante, 0, len(declarada.cambios))

	for indice, cambio := range declarada.cambios {
		if veces := strings.Count(skill, cambio.antes); veces != 1 {
			return nil, fmt.Errorf("el cambio %d ya no se aplica: su «antes» %q está %d veces en el %s de control, y "+
				"tiene que estar una", indice+1, cambio.antes, veces, ficheroDeLaSkillDeLaVariante)
		}

		sustituciones = append(sustituciones, sustitucionDeLaVariante{
			fichero: ficheroDeLaSkillDeLaVariante, antes: cambio.antes, despues: cambio.despues,
		})
	}

	cambiado, err := aplicarLasSustituciones(skill, sustituciones)
	if err != nil {
		return nil, fmt.Errorf("los cambios del %s de control: %w", ficheroDeLaSkillDeLaVariante, err)
	}

	calculados[ficheroDeLaSkillDeLaVariante] = []byte(cambiado)

	for indice, movimiento := range declarada.movimientos {
		referencia := path.Join(directorioDeLasReferencias, movimiento.referencia)

		contenido, aMano := calculados[referencia]
		if !aMano {
			return nil, fmt.Errorf("el movimiento %d va a %s, que no es una referencia de control escrita a mano",
				indice+1, referencia)
		}

		calculados[referencia] = append(contenido, movimiento.enLaReferencia...)
	}

	return calculados, nil
}

// sinLoGenerado son los ficheros de una skill sin lo que genera make skills-sync
// (research D7): sin las referencias cuya primera línea empieza por la cabecera de
// lo generado, y con cada región generada del SKILL.md reducida a sus marcas.
// Devuelve copias: los ficheros recibidos no cambian.
func sinLoGenerado(ficheros map[string][]byte) (map[string][]byte, error) {
	aMano := make(map[string][]byte, len(ficheros))

	for relativa, contenido := range ficheros {
		switch {
		case path.Dir(relativa) == directorioDeLasReferencias &&
			strings.HasPrefix(string(contenido), cabeceraDeLoGenerado):
			continue
		case relativa == ficheroDeLaSkillDeLaVariante:
			reducido, err := reducirLasRegionesGeneradas(string(contenido))
			if err != nil {
				return nil, fmt.Errorf("%s: %w", relativa, err)
			}

			aMano[relativa] = []byte(reducido)
		default:
			aMano[relativa] = slices.Clone(contenido)
		}
	}

	return aMano, nil
}

// reducirLasRegionesGeneradas es el SKILL.md con la región de los paquetes y la de
// la tabla de comandos reducidas a sus marcas: de la línea que empieza por la
// marca de inicio a la que es su marca de fin, se quedan esas dos. Una región que
// no se cierra es un error con su marca de fin.
func reducirLasRegionesGeneradas(contenido string) (string, error) {
	var reducido strings.Builder

	fin := ""

	for linea := range strings.SplitAfterSeq(contenido, "\n") {
		if fin != "" {
			if strings.TrimSuffix(linea, "\n") == fin {
				reducido.WriteString(linea)

				fin = ""
			}

			continue
		}

		reducido.WriteString(linea)

		switch {
		case strings.HasPrefix(linea, inicioDeLosPaquetes):
			fin = finDeLosPaquetes
		case strings.HasPrefix(linea, inicioDeLaRegionGenerada):
			fin = finDeLaRegionGenerada
		}
	}

	if fin != "" {
		return "", fmt.Errorf("una región generada no se cierra con la línea %q", fin)
	}

	return reducido.String(), nil
}

// textoDelMovimiento es un texto movido como se compara con lo que llega a su
// referencia (data-model §5): con los blancos juntados y cada
// `references/<x>` escrito `<x>`, como se enlaza desde references/.
func textoDelMovimiento(texto string) string {
	return juntarLosBlancos(strings.ReplaceAll(texto, directorioDeLasReferencias+"/", ""))
}

// juntarLosBlancos es el texto con cada tramo de blancos —espacios, tabuladores y
// saltos de línea— escrito como un espacio, y sin los del principio y el final.
func juntarLosBlancos(texto string) string {
	return strings.Join(strings.Fields(texto), " ")
}
