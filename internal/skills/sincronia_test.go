package skills_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// Las fuentes del repositorio temporal de TestRegenerarYComparar, sin nada
// generado: la tabla de normas, con una marcada vertebral, la jerarquía y el
// SKILL.md de dos skills. alfa declara dos applets, en otro orden que el del
// registro, y las tres referencias de la tabla de generadores, dos de ellas del
// mismo YAML de datos, y su tabla de comandos está escrita a mano; beta no declara
// nada, así que no lleva ni tabla ni referencias.
const (
	normasDeSincronia = "normas:\n" +
		"  BOE-A-2015-10565:\n" +
		"    titulo: Norma de prueba de la sincronía.\n" +
		"    rango: Ley\n" +
		"    materias:\n" +
		"      - procedimiento\n" +
		"    vertebral: true\n" +
		"  BOE-A-1985-5392:\n" +
		"    titulo: Otra norma de prueba de la sincronía.\n" +
		"    rango: Ley\n" +
		"    abreviatura: ONP\n" +
		"    materias:\n" +
		"      - régimen local\n"

	// jerarquiaDeSincronia tiene los cinco niveles y las cuatro reglas, cada uno
	// una vez y en su orden, como exige el esquema de data/jerarquia.yaml.
	jerarquiaDeSincronia = "niveles:\n" +
		"  - {nivel: ue, nombre: Unión, boletin: Diario de la Unión, normas: [Reglamento]}\n" +
		"  - {nivel: estado, nombre: Estado, boletin: Boletín del Estado, normas: [Ley, Real Decreto]}\n" +
		"  - {nivel: comunidad-autonoma, nombre: Comunidad, boletin: Boletín de la comunidad, normas: [Ley]}\n" +
		"  - {nivel: provincia, nombre: Provincia, boletin: Boletín de la provincia, normas: [Ordenanza]}\n" +
		"  - {nivel: municipio, nombre: Municipio, boletin: Boletín de la provincia, normas: [Ordenanza]}\n" +
		"reglas:\n" +
		"  - {regla: competencia-antes-que-jerarquia, enunciado: Primero la competencia.}\n" +
		"  - {regla: ley-posterior, enunciado: La posterior deroga a la anterior.}\n" +
		"  - {regla: ley-especial, enunciado: La especial prevalece.}\n" +
		"  - {regla: reglamento-nunca-contra-ley, enunciado: El reglamento cede ante la ley.}\n"

	// cabeceraDeAlfa es el SKILL.md de alfa hasta la marca de inicio de su tabla,
	// incluida: la marca está en la línea 10.
	cabeceraDeAlfa = "---\n" +
		"name: alfa\n" +
		"description: Skill de prueba con dos applets y tres referencias.\n" +
		"metadata:\n" +
		"  kitlegal-applets: dos uno\n" +
		"  kitlegal-referencias: " + referenciasDeAlfa + "\n" +
		"---\n" +
		"# Alfa\n" +
		"\n" +
		inicioDeLaTabla + "\n"

	// pieDeAlfa es el SKILL.md de alfa desde la marca de fin de su tabla, con un
	// enlace a cada referencia que declara (FR-007 de H31).
	pieDeAlfa = finDeLaTabla + "\n" +
		"\n" +
		"## Referencias\n" +
		"\n" +
		"- [normas](references/normas.md): las normas de la tabla.\n" +
		"- [leyes_vertebrales](references/leyes_vertebrales.md#leyes-vertebrales): las vertebrales.\n" +
		"- [jerarquia_normativa](references/jerarquia_normativa.md): los niveles.\n" +
		"\n" +
		"## Reglas\n"

	skillMdDeAlfaSinSincronizar = cabeceraDeAlfa + "tabla escrita a mano\n" + pieDeAlfa

	skillMdDeBetaSinNadaQueGenerar = "---\n" +
		"name: beta\n" +
		"description: Skill de prueba sin nada que generar.\n" +
		"---\n" +
		"# Beta\n"
)

// referenciasDeAlfa es el valor de kitlegal-referencias de alfa.
const referenciasDeAlfa = "normas leyes_vertebrales jerarquia_normativa"

// Lo que las pruebas de la sincronía dejan en references/ además de las
// referencias declaradas (FR-012): un fichero sobrante, que abre con la cabecera
// de una referencia generada sin que nadie lo declare, y una referencia a mano,
// que no abre con ella.
const (
	sobranteConCabecera = "<!-- generado desde data/normas.yaml, no editar -->\n\n# Sobrante\n"
	referenciaAMano     = "# Cómo se cita\n\nEl protocolo de detalle de la skill, escrito a mano.\n"
)

// Lo que las pruebas de la sincronía añaden a la primera norma de
// normasDeSincronia: el paquete de materia, que el esquema admite (FR-022 de
// H32), y un campo que no declara, con el defecto que da en alfa.
const (
	verticalDeNorma            = "    vertical: fiscal\n"
	campoNoDeclaradoDeNorma    = "    estado: vigente\n"
	defectoDelCampoNoDeclarado = "alfa: data/normas.yaml: BOE-A-2015-10565: campo no declarado: estado"
)

// Lo que la sincronía genera para alfa desde esas fuentes y desde
// descripcionesDeSincronia, byte a byte: la región de su tabla de comandos, con
// una sección por applet en el orden de su declaración (contrato
// sincronizacion-y-comprobacion §3); la referencia de las normas, por año y
// número del identificador (contrato normas-y-referencias §5); la de las leyes
// vertebrales, solo con la marcada; y la de la jerarquía, con sus cinco niveles y
// sus cuatro reglas (contrato de la skill legal-core §2).
const (
	regionDeAlfa = "\n" +
		"### `kitlegal dos`\n" +
		"\n" +
		columnasDeLaTabla +
		"| `kitlegal dos listar` | `dos_listar` | Lista los bloques. | lista de objetos con `bloque` |\n" +
		"\n" +
		"### `kitlegal uno`\n" +
		"\n" +
		columnasDeLaTabla +
		"| `kitlegal uno leer <bloque>` | `uno_leer` | Lee un bloque. | objeto con `texto` |\n" +
		"\n" +
		"La orden y la herramienta de cada fila devuelven el mismo sobre: `ok`, `data`; con `ok` falso, `data` " +
		"lleva `clase` y `mensaje`.\n" +
		"\n" +
		"Banderas comunes: `--json`, `--timeout <valor>`.\n" +
		"\n"

	skillMdDeAlfaSincronizado = cabeceraDeAlfa + regionDeAlfa + pieDeAlfa

	referenciaDeAlfa = comienzoDeLaReferencia +
		"| Otra norma de prueba de la sincronía. | ONP | `BOE-A-1985-5392` | Ley | régimen local |\n" +
		"| Norma de prueba de la sincronía. |  | `BOE-A-2015-10565` | Ley | procedimiento |\n"

	vertebralesDeAlfa = comienzoDeLasLeyesVertebrales +
		"| Norma de prueba de la sincronía. |  | `BOE-A-2015-10565` | Ley | procedimiento |\n"

	jerarquiaDeAlfa = "<!-- generado desde data/jerarquia.yaml, no editar -->\n" +
		"\n" +
		"# Jerarquía normativa\n" +
		"\n" +
		"| Nivel | Boletín | Tipos de norma, de mayor a menor rango |\n" +
		"|---|---|---|\n" +
		"| Unión | Diario de la Unión | Reglamento |\n" +
		"| Estado | Boletín del Estado | Ley, Real Decreto |\n" +
		"| Comunidad | Boletín de la comunidad | Ley |\n" +
		"| Provincia | Boletín de la provincia | Ordenanza |\n" +
		"| Municipio | Boletín de la provincia | Ordenanza |\n" +
		"\n" +
		"## Reglas de interpretación\n" +
		"\n" +
		"- `competencia-antes-que-jerarquia`: Primero la competencia.\n" +
		"- `ley-posterior`: La posterior deroga a la anterior.\n" +
		"- `ley-especial`: La especial prevalece.\n" +
		"- `reglamento-nunca-contra-ley`: El reglamento cede ante la ley.\n"
)

// referenciasRegeneradasDeAlfa son las tres referencias de alfa, en el orden de
// su declaración, en unas referencias nuevas en cada llamada.
func referenciasRegeneradasDeAlfa() []skills.Referencia {
	return []skills.Referencia{
		{Fichero: "normas.md", Contenido: []byte(referenciaDeAlfa)},
		{Fichero: "leyes_vertebrales.md", Contenido: []byte(vertebralesDeAlfa)},
		{Fichero: "jerarquia_normativa.md", Contenido: []byte(jerarquiaDeAlfa)},
	}
}

// descripcionesDeSincronia son las descripciones de los verbos del registro de
// prueba, en su orden —uno leer y dos listar—, con las mismas banderas y el mismo
// sobre, en unas descripciones nuevas en cada llamada.
func descripcionesDeSincronia() []skills.DescripcionDeVerbo {
	banderas := []skills.Bandera{{Nombre: "json"}, {Nombre: "timeout", ConValor: true}}
	sobre := skills.Sobre{Claves: []string{"ok", "data"}, ClavesDeFallo: []string{"clase", "mensaje"}}

	return []skills.DescripcionDeVerbo{
		{
			Applet:     "uno",
			Verbo:      "leer",
			Hace:       "Lee un bloque.",
			Argumentos: []skills.Argumento{{Nombre: "bloque", Obligatorio: true}},
			Banderas:   banderas,
			Devuelve:   skills.Devuelve{Forma: skills.FormaObjeto, Claves: []string{"texto"}},
			Sobre:      sobre,
		},
		{
			Applet:   "dos",
			Verbo:    "listar",
			Hace:     "Lista los bloques.",
			Banderas: banderas,
			Devuelve: skills.Devuelve{Forma: skills.FormaListaDeObjetos, Claves: []string{"bloque"}},
			Sobre:    sobre,
		},
	}
}

// TestRegenerarYComparar fija Regenerar, Escribir y Comparar sobre repositorios
// temporales (data-model §5; contrato sincronizacion-y-comprobacion §2;
// research.md D4, D5, D7 y D20; FR-034, FR-035, FR-036, FR-042, FR-065 a FR-067,
// SC-005, SC-009): lo generado sale de la declaración de cada skill y de data/,
// sin nada escrito para una skill concreta, y cada referencia, de la fila de la
// tabla de generadores que lleva su nombre, que no tiene por qué ser el de su
// YAML de datos; Regenerar y Comparar no escriben nada; Escribir deja el árbol sin
// derivas y sin ningún scripts/ (ADR 0019; FR-080), y un segundo Escribir no
// cambia ningún byte ni ningún tiempo de modificación; una referencia a mano no es
// deriva y Escribir no la toca, mientras que la declarada sin su cabecera y el .md
// no declarado que abre con ella sí lo son (FR-012); una norma con vertical no es
// un defecto (FR-022 de H32); cada clase de deriva nombra la skill y el fichero,
// y Escribir la deshace; y cada defecto de una skill —entre ellos, llevar
// scripts/ (FR-082) y los de references/ (FR-006 a FR-009 de H31)— nombra la
// skill, la deja sin nada que comparar e impide que Escribir escriba nada.
func TestRegenerarYComparar(t *testing.T) {
	t.Parallel()

	t.Run("escribir-sincroniza", func(t *testing.T) {
		t.Parallel()

		raiz := arbolDeFuentes(t)
		antes := fotografiar(t, raiz)

		regenerado := regenerar(t, raiz, descripcionesDeSincronia())
		assert.Equal(t, skills.Regenerado{Skills: []skills.SkillRegenerada{
			{
				Nombre:      "alfa",
				Contenido:   []byte(skillMdDeAlfaSincronizado),
				Referencias: referenciasRegeneradasDeAlfa(),
			},
			{Nombre: "beta", Contenido: []byte(skillMdDeBetaSinNadaQueGenerar)},
		}}, regenerado)

		derivas, err := skills.Comparar(raiz, regenerado)
		require.NoError(t, err)
		assert.Equal(t, []*skills.Deriva{
			{Skill: "alfa", Ruta: "SKILL.md", Clase: skills.DerivaContenidoDistinto},
			{Skill: "alfa", Ruta: "references/normas.md", Clase: skills.DerivaFicheroAusente},
			{Skill: "alfa", Ruta: "references/leyes_vertebrales.md", Clase: skills.DerivaFicheroAusente},
			{Skill: "alfa", Ruta: "references/jerarquia_normativa.md", Clase: skills.DerivaFicheroAusente},
		}, derivas)
		assert.Equal(t, antes, fotografiar(t, raiz), "Regenerar y Comparar no escriben nada")

		require.NoError(t, skills.Escribir(raiz, regenerado))
		assert.Empty(t, compararSinDefectos(t, raiz, descripcionesDeSincronia()))

		assert.Equal(t, skillMdDeAlfaSincronizado, leerFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "SKILL.md")))

		for _, referencia := range referenciasRegeneradasDeAlfa() {
			assert.Equal(t, string(referencia.Contenido),
				leerFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", referencia.Fichero)))
		}

		assert.Equal(t, skillMdDeBetaSinNadaQueGenerar, leerFicheroDePrueba(t, rutaDeSkill(raiz, "beta", "SKILL.md")))
		assert.NoDirExists(t, rutaDeSkill(raiz, "beta", "references"), "beta no declara referencias")

		for _, skill := range []string{"alfa", "beta"} {
			_, err := os.Lstat(rutaDeSkill(raiz, skill, "scripts"))
			require.ErrorIs(t, err, fs.ErrNotExist, "Escribir no le crea scripts/ a %s, declare o no applets", skill)
		}
	})

	t.Run("segundo-escribir-no-cambia-nada", func(t *testing.T) {
		t.Parallel()

		raiz := arbolDeFuentes(t)
		escribirFicheroDePrueba(t, rutaDeSkill(raiz, "beta", "references", "sobrante.md"), sobranteConCabecera)
		escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "como_se_cita.md"), referenciaAMano)
		enlazarDesdeSkillMd(t, raiz, "alfa", "como_se_cita.md")

		require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, descripcionesDeSincronia())))
		require.Empty(t, compararSinDefectos(t, raiz, descripcionesDeSincronia()))

		envejecer(t, raiz)
		antes := fotografiar(t, raiz)

		require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, descripcionesDeSincronia())))
		assert.Equal(t, antes, fotografiar(t, raiz),
			"el segundo Escribir no cambia ningún byte ni ningún tiempo de modificación")
	})

	t.Run("norma-con-vertical", func(t *testing.T) {
		t.Parallel()

		raiz := arbolDeFuentes(t)
		cambiarFicheroDePrueba(t, filepath.Join(raiz, "data", "normas.yaml"),
			"      - procedimiento\n", "      - procedimiento\n"+verticalDeNorma)

		regenerado := regenerar(t, raiz, descripcionesDeSincronia())
		assert.Empty(t, presentarDefectos(regenerado.Defectos()),
			"una norma con vertical es una norma válida, sin defecto en ninguna referencia de data/normas.yaml")
	})

	t.Run("referencias-a-mano", probarReferenciasAMano)
	t.Run("referencia-que-deja-de-ser-un-fichero-regular", probarReferenciaQueDejaDeSerUnFichero)

	probarDerivas(t)
	probarDefectos(t)
}

// probarReferenciasAMano fija la clasificación de FR-012 en la sincronía
// (contracts/gates-de-skills-check.md §3; research D13): sobre un árbol
// sincronizado con una referencia a mano en cada skill, enlazada desde su
// SKILL.md —en beta, con el nombre de una referencia que no declara—, la
// referencia declarada sin su cabecera y un .md no declarado que abre con ella,
// Comparar da la deriva de la declarada y la sobrante, y ninguna de las
// referencias a mano; Escribir rehace la declarada, retira la sobrante y no toca
// las referencias a mano, ni sus bytes ni su tiempo de modificación. Una entrada
// que no es un fichero .md ya no es sobrante, sino un defecto (FR-006 de H31,
// en probarDefectos).
func probarReferenciasAMano(t *testing.T) {
	t.Parallel()

	raiz := arbolSincronizado(t)
	aMano := []string{"skills/alfa/references/como_se_cita.md", "skills/beta/references/normas.md"}

	for _, ruta := range aMano {
		escribirFicheroDePrueba(t, filepath.Join(raiz, filepath.FromSlash(ruta)), referenciaAMano)
	}

	enlazarDesdeSkillMd(t, raiz, "alfa", "como_se_cita.md")
	enlazarDesdeSkillMd(t, raiz, "beta", "normas.md")

	cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "normas.md"),
		"<!-- generado desde data/normas.yaml, no editar -->\n", "")
	escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "antigua.md"), sobranteConCabecera)

	assert.Equal(t, []*skills.Deriva{
		{Skill: "alfa", Ruta: "references/normas.md", Clase: skills.DerivaContenidoDistinto},
		{Skill: "alfa", Ruta: "references/antigua.md", Clase: skills.DerivaFicheroSobrante},
	}, compararSinDefectos(t, raiz, descripcionesDeSincronia()))

	envejecer(t, raiz)
	antes := fotografiar(t, raiz)

	require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, descripcionesDeSincronia())))
	assert.Empty(t, compararSinDefectos(t, raiz, descripcionesDeSincronia()))

	despues := fotografiar(t, raiz)
	for _, ruta := range aMano {
		assert.Equal(t, antes[ruta], despues[ruta], "Escribir no toca la referencia a mano %s", ruta)
		assert.Equal(t, referenciaAMano, despues[ruta].Contenido)
	}

	assert.Equal(t, referenciaDeAlfa, leerFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "normas.md")))
	assert.NotContains(t, despues, "skills/alfa/references/antigua.md")
}

// probarReferenciaQueDejaDeSerUnFichero fija lo que Comparar y Escribir hacen
// con una referencia declarada que deja de ser un fichero regular después de
// Regenerar, que no la ha visto así: con ella, Regenerar daría el defecto de
// FR-006 de H31. Comparar la da como contenido distinto, con el detalle, y
// Escribir retira el enlace, sin escribir en su destino, y deja en su lugar la
// referencia regenerada.
func probarReferenciaQueDejaDeSerUnFichero(t *testing.T) {
	t.Parallel()

	raiz := arbolSincronizado(t)
	regenerado := regenerar(t, raiz, descripcionesDeSincronia())

	ruta := rutaDeSkill(raiz, "alfa", "references", "normas.md")
	retirarDePrueba(t, ruta)
	crearEnlaceDePrueba(t, "../../../data/normas.yaml", ruta)

	derivas, err := skills.Comparar(raiz, regenerado)
	require.NoError(t, err)
	require.Len(t, derivas, 1)
	assert.Equal(t, "alfa: references/normas.md: contenido-distinto (no es un fichero regular)", derivas[0].Error())

	require.NoError(t, skills.Escribir(raiz, regenerado))
	assert.Empty(t, compararSinDefectos(t, raiz, descripcionesDeSincronia()))
	assert.Equal(t, normasDeSincronia, leerFicheroDePrueba(t, filepath.Join(raiz, "data", "normas.yaml")),
		"Escribir no escribe en el destino del enlace")
}

// probarDerivas fija cada clase de deriva de data-model §5: sobre un árbol
// sincronizado con un solo cambio, Comparar da exactamente esa deriva, que nombra
// la skill y el fichero, y Escribir la deshace sin dejar ningún defecto.
func probarDerivas(t *testing.T) {
	t.Helper()

	casos := []struct {
		nombre               string
		alterar              func(t *testing.T, raiz string)
		cambiarDescripciones func(descripciones []skills.DescripcionDeVerbo)
		deriva               skills.Deriva
		mensaje              string
	}{
		{
			nombre: "referencia-editada",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				ruta := rutaDeSkill(raiz, "alfa", "references", "normas.md")
				escribirFicheroDePrueba(t, ruta, leerFicheroDePrueba(t, ruta)+"| una fila escrita a mano |\n")
			},
			deriva:  skills.Deriva{Skill: "alfa", Ruta: "references/normas.md", Clase: skills.DerivaContenidoDistinto},
			mensaje: "alfa: references/normas.md: contenido-distinto",
		},
		{
			// La norma no es vertebral: solo cambia la referencia de todas las normas.
			nombre: "datos-sin-regenerar",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, filepath.Join(raiz, "data", "normas.yaml"),
					"titulo: Otra norma de prueba de la sincronía.", "titulo: Otra norma de prueba con otro título.")
			},
			deriva:  skills.Deriva{Skill: "alfa", Ruta: "references/normas.md", Clase: skills.DerivaContenidoDistinto},
			mensaje: "alfa: references/normas.md: contenido-distinto",
		},
		{
			// La marca no es una columna: la referencia de todas las normas no cambia,
			// y la de las leyes vertebrales gana una fila.
			nombre: "marca-vertebral-sin-regenerar",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, filepath.Join(raiz, "data", "normas.yaml"),
					"      - régimen local\n", "      - régimen local\n    vertebral: true\n")
			},
			deriva: skills.Deriva{
				Skill: "alfa",
				Ruta:  "references/leyes_vertebrales.md",
				Clase: skills.DerivaContenidoDistinto,
			},
			mensaje: "alfa: references/leyes_vertebrales.md: contenido-distinto",
		},
		{
			nombre: "jerarquia-sin-regenerar",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, filepath.Join(raiz, "data", "jerarquia.yaml"),
					"enunciado: La especial prevalece.", "enunciado: La especial prevalece sobre la general.")
			},
			deriva: skills.Deriva{
				Skill: "alfa",
				Ruta:  "references/jerarquia_normativa.md",
				Clase: skills.DerivaContenidoDistinto,
			},
			mensaje: "alfa: references/jerarquia_normativa.md: contenido-distinto",
		},
		{
			nombre: "describe-cambiado",
			cambiarDescripciones: func(descripciones []skills.DescripcionDeVerbo) {
				descripciones[0].Hace = "Lee un bloque y sus notas."
			},
			deriva:  skills.Deriva{Skill: "alfa", Ruta: "SKILL.md", Clase: skills.DerivaContenidoDistinto},
			mensaje: "alfa: SKILL.md: contenido-distinto",
		},
		{
			nombre: "region-editada",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "SKILL.md"), "| Lee un bloque. |", "| Lee lo que pida. |")
			},
			deriva:  skills.Deriva{Skill: "alfa", Ruta: "SKILL.md", Clase: skills.DerivaContenidoDistinto},
			mensaje: "alfa: SKILL.md: contenido-distinto",
		},
		{
			// Declarada, es la referencia generada aunque haya perdido la cabecera
			// (FR-012).
			nombre: "referencia-declarada-sin-cabecera",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "normas.md"),
					"<!-- generado desde data/normas.yaml, no editar -->\n", "")
			},
			deriva:  skills.Deriva{Skill: "alfa", Ruta: "references/normas.md", Clase: skills.DerivaContenidoDistinto},
			mensaje: "alfa: references/normas.md: contenido-distinto",
		},
		{
			nombre: "referencia-ausente",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				retirarDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "normas.md"))
			},
			deriva:  skills.Deriva{Skill: "alfa", Ruta: "references/normas.md", Clase: skills.DerivaFicheroAusente},
			mensaje: "alfa: references/normas.md: fichero-ausente",
		},
		{
			nombre: "referencia-sobrante",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "antigua.md"), sobranteConCabecera)
			},
			deriva:  skills.Deriva{Skill: "alfa", Ruta: "references/antigua.md", Clase: skills.DerivaFicheroSobrante},
			mensaje: "alfa: references/antigua.md: fichero-sobrante",
		},
		{
			nombre: "referencia-en-una-skill-que-no-declara-ninguna",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				escribirFicheroDePrueba(t, rutaDeSkill(raiz, "beta", "references", "normas.md"), referenciaDeAlfa)
			},
			deriva:  skills.Deriva{Skill: "beta", Ruta: "references/normas.md", Clase: skills.DerivaFicheroSobrante},
			mensaje: "beta: references/normas.md: fichero-sobrante",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := arbolSincronizado(t)
			descripciones := descripcionesDeSincronia()

			if caso.alterar != nil {
				caso.alterar(t, raiz)
			}

			if caso.cambiarDescripciones != nil {
				caso.cambiarDescripciones(descripciones)
			}

			regenerado := regenerar(t, raiz, descripciones)
			require.Empty(t, regenerado.Defectos())

			derivas, err := skills.Comparar(raiz, regenerado)
			require.NoError(t, err)
			assert.Equal(t, []*skills.Deriva{&caso.deriva}, derivas)
			require.Len(t, derivas, 1)
			assert.Equal(t, caso.mensaje, derivas[0].Error())

			require.NoError(t, skills.Escribir(raiz, regenerado))
			assert.Empty(t, compararSinDefectos(t, raiz, descripciones), "Escribir deshace la deriva")
		})
	}
}

// casoDeDefecto es un cambio sobre un árbol sincronizado —en sus ficheros, en
// las descripciones de los verbos o en los dos— y los defectos exactos que da, en
// su orden.
type casoDeDefecto struct {
	nombre               string
	alterar              func(t *testing.T, raiz string)
	cambiarDescripciones func(descripciones []skills.DescripcionDeVerbo)
	defectos             []string
}

// probarDefectos fija los defectos de una skill (data-model §5): cada uno nombra
// la skill y van en el orden frontmatter, cuerpo, región, datos, references/ y
// scripts/; una skill con defectos no tiene nada regenerado con que comparar,
// aunque las demás sí; y Escribir no escribe nada, tampoco lo de las skills sin
// defectos. Una entrada scripts en el directorio de una skill, sea lo que sea —el
// directorio con los enlaces que dejaba la instalación anterior, uno vacío, un
// fichero o un enlace colgante—, es un defecto, también en una skill sin
// SKILL.md, y Escribir no la retira (ADR 0019; contracts/skills-e-invocacion.md
// §3; FR-082). Los de references/ los da casosDeReferences.
func probarDefectos(t *testing.T) {
	t.Helper()

	casos := []casoDeDefecto{
		{
			nombre: "frontmatter",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "SKILL.md"), "name: alfa\n", "name: otra\n")
			},
			defectos: []string{`alfa: SKILL.md: name "otra" distinto del nombre del directorio`},
		},
		{
			nombre: "frontmatter-ilegible",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "SKILL.md"), "# Alfa sin frontmatter\n")
			},
			defectos: []string{"alfa: SKILL.md: sin frontmatter: la primera línea no es ---"},
		},
		{
			nombre: "applet-no-registrado",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "SKILL.md"),
					"kitlegal-applets: dos uno\n", "kitlegal-applets: dos tres\n")
			},
			defectos: []string{`alfa: SKILL.md: metadata/kitlegal-applets: applet "tres" no registrado`},
		},
		{
			// Un YAML de datos con su nombre no le da generador: lo da la tabla.
			nombre: "referencia-sin-generador",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "SKILL.md"),
					"kitlegal-referencias: "+referenciasDeAlfa+"\n", "kitlegal-referencias: "+referenciasDeAlfa+" otras\n")
				escribirFicheroDePrueba(t, filepath.Join(raiz, "data", "otras.yaml"), "otras: []\n")
			},
			defectos: []string{`alfa: SKILL.md: metadata/kitlegal-referencias: "otras" sin generador conocido`},
		},
		{
			nombre: "referencia-sin-su-yaml-de-datos",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				retirarDePrueba(t, filepath.Join(raiz, "data", "jerarquia.yaml"))
			},
			defectos: []string{
				`alfa: SKILL.md: metadata/kitlegal-referencias: "jerarquia_normativa" sin data/jerarquia.yaml`,
			},
		},
		{
			nombre: "region-sin-marcas",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				retirarLasMarcas(t, rutaDeSkill(raiz, "alfa", "SKILL.md"))
			},
			defectos: []string{"alfa: SKILL.md: sin las marcas de la tabla de comandos"},
		},
		{
			nombre: "region-con-dos-inicios",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "SKILL.md"), "# Alfa\n", "# Alfa\n"+inicioDeLaTabla+"\n")
			},
			defectos: []string{
				"alfa: SKILL.md: la marca de inicio de la tabla de comandos aparece 2 veces, en las líneas 9 y 11",
			},
		},
		{
			// La tabla afirma que las banderas son las de todas sus órdenes: la
			// primera es la de dos listar, el primer applet que declara alfa.
			nombre: "tabla-con-banderas-distintas",
			cambiarDescripciones: func(descripciones []skills.DescripcionDeVerbo) {
				descripciones[0].Banderas = []skills.Bandera{{Nombre: "json"}}
			},
			defectos: []string{"alfa: SKILL.md: uno leer no declara las mismas banderas globales que dos listar"},
		},
		{
			// Dos referencias de alfa salen de data/normas.yaml, y su defecto va una
			// sola vez.
			nombre: "norma-con-campo-no-declarado",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, filepath.Join(raiz, "data", "normas.yaml"),
					"      - procedimiento\n", "      - procedimiento\n"+campoNoDeclaradoDeNorma)
			},
			defectos: []string{defectoDelCampoNoDeclarado},
		},
		{
			// Como legal-core: leyes_vertebrales sin normas delante, así que el
			// defecto de data/normas.yaml lo encuentra su propio generador.
			nombre: "vertebrales-sin-normas-con-campo-no-declarado",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "SKILL.md"),
					"kitlegal-referencias: "+referenciasDeAlfa+"\n", "kitlegal-referencias: leyes_vertebrales jerarquia_normativa\n")
				cambiarFicheroDePrueba(t, filepath.Join(raiz, "data", "normas.yaml"),
					"      - procedimiento\n", "      - procedimiento\n"+campoNoDeclaradoDeNorma)
			},
			defectos: []string{defectoDelCampoNoDeclarado},
		},
		{
			nombre: "jerarquia-con-una-clave-desconocida",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, filepath.Join(raiz, "data", "jerarquia.yaml"),
					"normas: [Reglamento]}", "normas: [Reglamento], vigencia: 2026}")
			},
			defectos: []string{"alfa: data/jerarquia.yaml: niveles/0, línea 2: additional properties 'vigencia' not allowed"},
		},
		{
			nombre: "sin-skill-md",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				require.NoError(t, os.MkdirAll(rutaDeSkill(raiz, "gamma"), 0o750))
			},
			defectos: []string{"gamma: falta SKILL.md"},
		},
		{
			nombre: "frontmatter-y-cuerpo-en-su-orden",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				ruta := rutaDeSkill(raiz, "alfa", "SKILL.md")
				cambiarFicheroDePrueba(t, ruta, "name: alfa\n", "name: otra\n")
				alargarElCuerpo(t, ruta, 500)
			},
			defectos: []string{
				`alfa: SKILL.md: name "otra" distinto del nombre del directorio`,
				"alfa: SKILL.md: cuerpo de 500 líneas (máximo 499)",
			},
		},
		{
			nombre: "cuerpo-y-region-en-su-orden",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				ruta := rutaDeSkill(raiz, "alfa", "SKILL.md")
				retirarLasMarcas(t, ruta)
				alargarElCuerpo(t, ruta, 500)
			},
			defectos: []string{
				"alfa: SKILL.md: cuerpo de 500 líneas (máximo 499)",
				"alfa: SKILL.md: sin las marcas de la tabla de comandos",
			},
		},
		{
			nombre: "scripts-con-los-enlaces-de-la-instalacion-anterior",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				for _, applet := range []string{"dos", "uno"} {
					crearEnlaceDePrueba(t, "../../../bin/instalado/kitlegal", rutaDeSkill(raiz, "alfa", "scripts", applet))
				}
			},
			defectos: []string{"alfa: una skill no lleva scripts/ (ADR 0019)"},
		},
		{
			nombre: "scripts-vacio",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				require.NoError(t, os.Mkdir(rutaDeSkill(raiz, "alfa", "scripts"), 0o750))
			},
			defectos: []string{"alfa: una skill no lleva scripts/ (ADR 0019)"},
		},
		{
			nombre: "scripts-que-es-un-fichero",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "scripts"), "#!/bin/sh\n")
			},
			defectos: []string{"alfa: una skill no lleva scripts/ (ADR 0019)"},
		},
		{
			// Lstat: un enlace que no resuelve también está.
			nombre: "scripts-que-es-un-enlace-colgante",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				crearEnlaceDePrueba(t, "no-existe", rutaDeSkill(raiz, "alfa", "scripts"))
			},
			defectos: []string{"alfa: una skill no lleva scripts/ (ADR 0019)"},
		},
		{
			nombre: "scripts-en-una-skill-sin-skill-md",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				require.NoError(t, os.MkdirAll(rutaDeSkill(raiz, "gamma", "scripts"), 0o750))
			},
			defectos: []string{"gamma: falta SKILL.md", "gamma: una skill no lleva scripts/ (ADR 0019)"},
		},
		{
			nombre: "datos-y-scripts-en-su-orden",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, filepath.Join(raiz, "data", "normas.yaml"),
					"      - procedimiento\n", "      - procedimiento\n"+campoNoDeclaradoDeNorma)
				require.NoError(t, os.Mkdir(rutaDeSkill(raiz, "alfa", "scripts"), 0o750))
			},
			defectos: []string{
				defectoDelCampoNoDeclarado,
				"alfa: una skill no lleva scripts/ (ADR 0019)",
			},
		},
	}

	casos = append(casos, casosDeReferences()...)

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := arbolSincronizado(t)
			if caso.alterar != nil {
				caso.alterar(t, raiz)
			}

			// beta no tiene defectos y sí una deriva, que Comparar sigue viendo y que
			// Escribir tampoco deshace.
			escribirFicheroDePrueba(t, rutaDeSkill(raiz, "beta", "references", "sobrante.md"), sobranteConCabecera)
			envejecer(t, raiz)
			antes := fotografiar(t, raiz)

			descripciones := descripcionesDeSincronia()
			if caso.cambiarDescripciones != nil {
				caso.cambiarDescripciones(descripciones)
			}

			regenerado := regenerar(t, raiz, descripciones)
			assert.Equal(t, caso.defectos, presentarDefectos(regenerado.Defectos()))

			derivas, err := skills.Comparar(raiz, regenerado)
			require.NoError(t, err)
			assert.Equal(t,
				[]*skills.Deriva{{Skill: "beta", Ruta: "references/sobrante.md", Clase: skills.DerivaFicheroSobrante}},
				derivas, "una skill con defectos no tiene nada regenerado con que comparar; las demás, sí")

			err = skills.Escribir(raiz, regenerado)

			var defecto *skills.DefectoDeSkill
			require.ErrorAs(t, err, &defecto)

			for _, texto := range caso.defectos {
				require.ErrorContains(t, err, texto)
			}

			assert.Equal(t, antes, fotografiar(t, raiz), "con defectos, Escribir no escribe nada")
		})
	}
}

// casosDeReferences son los defectos de references/ de alfa
// (contracts/gates-de-skills-check.md §2.3 y §3; research D13 y D14; FR-006 a
// FR-009 y FR-011 de H31), sobre su SKILL.md regenerado y sus referencias —las
// generadas, como las regenera su fila; las a mano, como están—: una entrada que
// no es un fichero regular .md, también la de una referencia declarada; una
// referencia, generada o a mano, sin un enlace desde SKILL.md fuera de un bloque
// de código; una que enlaza otra de la skill; y una de más de 100 líneas sin su
// índice, también la generada. Van tras los de los datos y antes del de
// scripts/: los de las entradas, en su orden, y después los de cada referencia,
// en el orden de sus nombres, por regla. Con los datos con defectos no hay
// referencias regeneradas y solo se miran las entradas.
func casosDeReferences() []casoDeDefecto {
	const (
		conEnlaces = "Las normas, en [la tabla](normas.md#normas-de-referencia), [el ámbito](./ambito.md), " +
			"[otra vez](../references/normas.md) y [aquí](como_se_cita.md#x).\n"
		sinEnlaceDeLeyes = "- [leyes_vertebrales](references/leyes_vertebrales.md#leyes-vertebrales): las vertebrales.\n"
		sinIndiceEn101   = "101 líneas sin un índice que nombre sus apartados ## antes del primero (más de 100)"
	)

	return []casoDeDefecto{
		{
			nombre: "subdirectorio-en-references",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "anexos", "nota.md"), referenciaAMano)
			},
			defectos: []string{"alfa: references/anexos: no es un fichero .md"},
		},
		{
			nombre: "fichero-de-otra-extension-en-references",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "notas.txt"), referenciaAMano)
			},
			defectos: []string{"alfa: references/notas.txt: no es un fichero .md"},
		},
		{
			// Declarada, la que no es un fichero regular es un defecto, no una
			// deriva, y Escribir no la retira.
			nombre: "referencia-declarada-que-es-un-enlace",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				ruta := rutaDeSkill(raiz, "alfa", "references", "normas.md")
				retirarDePrueba(t, ruta)
				crearEnlaceDePrueba(t, "../../../data/normas.yaml", ruta)
			},
			defectos: []string{"alfa: references/normas.md: no es un fichero .md"},
		},
		{
			nombre: "referencia-generada-sin-enlace",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "SKILL.md"), sinEnlaceDeLeyes, "")
			},
			defectos: []string{
				"alfa: references/leyes_vertebrales.md: SKILL.md no la enlaza con [](references/leyes_vertebrales.md)",
			},
		},
		{
			// Ausente del árbol, la referencia es la que regenera su fila: la mira
			// igual.
			nombre: "referencia-ausente-sin-enlace",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				retirarDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "leyes_vertebrales.md"))
				cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "SKILL.md"), sinEnlaceDeLeyes, "")
			},
			defectos: []string{
				"alfa: references/leyes_vertebrales.md: SKILL.md no la enlaza con [](references/leyes_vertebrales.md)",
			},
		},
		{
			nombre: "referencia-enlazada-solo-dentro-de-un-bloque",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "SKILL.md"), sinEnlaceDeLeyes,
					"```text\n"+sinEnlaceDeLeyes+"```\n")
			},
			defectos: []string{
				"alfa: references/leyes_vertebrales.md: SKILL.md no la enlaza con [](references/leyes_vertebrales.md)",
			},
		},
		{
			nombre: "referencia-a-mano-sin-enlace",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "como_se_cita.md"), referenciaAMano)
			},
			defectos: []string{
				"alfa: references/como_se_cita.md: SKILL.md no la enlaza con [](references/como_se_cita.md)",
			},
		},
		{
			// Una sola vez cada referencia enlazada, en el orden en que aparece; el
			// enlace a la propia referencia no cuenta.
			nombre: "referencia-que-enlaza-otras",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "como_se_cita.md"),
					referenciaAMano+conEnlaces)
				escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "ambito.md"), referenciaAMano)
				enlazarDesdeSkillMd(t, raiz, "alfa", "como_se_cita.md", "ambito.md")
			},
			defectos: []string{
				"alfa: references/como_se_cita.md: enlaza otra referencia, references/normas.md",
				"alfa: references/como_se_cita.md: enlaza otra referencia, references/ambito.md",
			},
		},
		{
			nombre: "referencia-a-mano-de-ciento-una-lineas-sin-indice",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "como_se_cita.md"),
					referenciaDe(t, []string{"## Lo que hace cita"}, nil, 101))
				enlazarDesdeSkillMd(t, raiz, "alfa", "como_se_cita.md")
			},
			defectos: []string{"alfa: references/como_se_cita.md: " + sinIndiceEn101},
		},
		{
			// La referencia de las normas, con 95, tiene 101 líneas y ningún
			// apartado: el generador no cambia en H31 (FR-009, Clarificación Q4).
			nombre: "referencia-generada-de-ciento-una-lineas",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				escribirFicheroDePrueba(t, filepath.Join(raiz, "data", "normas.yaml"), normasDeSincroniaConRelleno(93))
			},
			defectos: []string{"alfa: references/normas.md: 101 líneas sin ningún apartado ## (más de 100)"},
		},
		{
			nombre: "entradas-referencias-y-scripts-en-su-orden",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "notas.txt"), referenciaAMano)
				escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "como_se_cita.md"),
					referenciaDe(t, []string{strings.TrimSuffix(conEnlaces, "\n"), "## Lo que hace cita"}, nil, 101))
				cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "SKILL.md"), sinEnlaceDeLeyes, "")
				require.NoError(t, os.Mkdir(rutaDeSkill(raiz, "alfa", "scripts"), 0o750))
			},
			defectos: []string{
				"alfa: references/notas.txt: no es un fichero .md",
				"alfa: references/como_se_cita.md: SKILL.md no la enlaza con [](references/como_se_cita.md)",
				"alfa: references/como_se_cita.md: enlaza otra referencia, references/normas.md",
				"alfa: references/como_se_cita.md: " + sinIndiceEn101,
				"alfa: references/leyes_vertebrales.md: SKILL.md no la enlaza con [](references/leyes_vertebrales.md)",
				"alfa: una skill no lleva scripts/ (ADR 0019)",
			},
		},
		{
			nombre: "datos-con-defectos-y-entradas",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, filepath.Join(raiz, "data", "normas.yaml"),
					"      - procedimiento\n", "      - procedimiento\n"+campoNoDeclaradoDeNorma)
				escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "notas.txt"), referenciaAMano)
				escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "como_se_cita.md"), referenciaAMano)
			},
			defectos: []string{
				defectoDelCampoNoDeclarado,
				"alfa: references/notas.txt: no es un fichero .md",
			},
		},
	}
}

// normasDeSincroniaConRelleno es data/normas.yaml de la sincronía con tantas
// normas más como se piden, ninguna vertebral.
func normasDeSincroniaConRelleno(relleno int) string {
	var normas strings.Builder

	normas.WriteString(normasDeSincronia)

	for numero := 1; numero <= relleno; numero++ {
		fmt.Fprintf(&normas, "  BOE-A-2000-%d:\n"+
			"    titulo: Norma de relleno %d.\n"+
			"    rango: Ley\n"+
			"    materias:\n"+
			"      - procedimiento\n", numero, numero)
	}

	return normas.String()
}

// enlazarDesdeSkillMd añade al final del SKILL.md de la skill del repositorio de
// raíz, fuera de la región generada, un enlace a cada fichero de references/
// dado (FR-007 de H31).
func enlazarDesdeSkillMd(t *testing.T, raiz, skill string, ficheros ...string) {
	t.Helper()

	ruta := rutaDeSkill(raiz, skill, "SKILL.md")
	contenido := leerFicheroDePrueba(t, ruta)

	for _, fichero := range ficheros {
		contenido += "- [" + strings.TrimSuffix(fichero, ".md") + "](references/" + fichero + ")\n"
	}

	escribirFicheroDePrueba(t, ruta, contenido)
}

// TestCuerpoDeLaSkill fija los presupuestos del cuerpo de SKILL.md
// (contracts/gates-de-skills-check.md §2.2; data-model §1; research D12; FR-004,
// FR-010, FR-011): el SKILL.md regenerado desde el byte que sigue a la línea que
// cierra su frontmatter, que no cuenta, tiene menos de 500 líneas y como mucho
// 11 000 caracteres, contados como runas y no como bytes. Sobre el árbol
// sincronizado, cada caso cambia el SKILL.md de alfa y exige los defectos
// exactos: ninguno en el límite, y entonces ninguna deriva; uno por medida por
// encima, en el orden líneas y caracteres, y entonces Escribir no escribe nada.
func TestCuerpoDeLaSkill(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		alterar  func(t *testing.T, ruta string)
		defectos []string
	}{
		{
			// Con las 7 del frontmatter, el fichero tiene 506 líneas.
			nombre: "cuatrocientas-noventa-y-nueve-lineas",
			alterar: func(t *testing.T, ruta string) {
				t.Helper()

				alargarElCuerpo(t, ruta, 499)
			},
		},
		{
			nombre: "quinientas-lineas",
			alterar: func(t *testing.T, ruta string) {
				t.Helper()

				alargarElCuerpo(t, ruta, 500)
			},
			defectos: []string{"alfa: SKILL.md: cuerpo de 500 líneas (máximo 499)"},
		},
		{
			// El cuerpo que se mide es el regenerado: con la región de la tabla
			// vacía, el del árbol tiene tantas líneas menos como la tabla, y la
			// tabla que se regenera lo lleva a 500.
			nombre: "quinientas-lineas-al-regenerar",
			alterar: func(t *testing.T, ruta string) {
				t.Helper()

				cambiarFicheroDePrueba(t, ruta, regionDeAlfa, "")
				alargarElCuerpo(t, ruta, 500-strings.Count(regionDeAlfa, "\n"))
			},
			defectos: []string{"alfa: SKILL.md: cuerpo de 500 líneas (máximo 499)"},
		},
		{
			// Con el frontmatter, el fichero pasa de 11 000 caracteres, y el cuerpo,
			// de 11 000 bytes.
			nombre: "once-mil-caracteres",
			alterar: func(t *testing.T, ruta string) {
				t.Helper()

				ensancharElCuerpo(t, ruta, 11000)
			},
		},
		{
			nombre: "once-mil-y-un-caracteres",
			alterar: func(t *testing.T, ruta string) {
				t.Helper()

				ensancharElCuerpo(t, ruta, 11001)
			},
			defectos: []string{"alfa: SKILL.md: cuerpo de 11001 caracteres (máximo 11000)"},
		},
		{
			// ensancharElCuerpo añade una línea: de 499 a 500.
			nombre: "lineas-y-caracteres-en-su-orden",
			alterar: func(t *testing.T, ruta string) {
				t.Helper()

				alargarElCuerpo(t, ruta, 499)
				ensancharElCuerpo(t, ruta, 11001)
			},
			defectos: []string{
				"alfa: SKILL.md: cuerpo de 500 líneas (máximo 499)",
				"alfa: SKILL.md: cuerpo de 11001 caracteres (máximo 11000)",
			},
		},
		{
			// Sin una línea que cierre el frontmatter no hay cuerpo que medir: el
			// defecto es el del frontmatter.
			nombre: "frontmatter-sin-cierre",
			alterar: func(t *testing.T, ruta string) {
				t.Helper()

				escribirFicheroDePrueba(t, ruta, "---\nname: alfa\n"+strings.Repeat("ñ\n", 6000))
			},
			defectos: []string{"alfa: SKILL.md: frontmatter sin cierre: ninguna línea --- después de la primera"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := arbolSincronizado(t)
			caso.alterar(t, rutaDeSkill(raiz, "alfa", "SKILL.md"))

			regenerado := regenerar(t, raiz, descripcionesDeSincronia())
			require.Equal(t, caso.defectos, presentarDefectos(regenerado.Defectos()))

			if len(caso.defectos) == 0 {
				derivas, err := skills.Comparar(raiz, regenerado)
				require.NoError(t, err)
				assert.Empty(t, derivas, "en el límite, el cuerpo cambiado fuera de la región no deriva")

				return
			}

			envejecer(t, raiz)
			antes := fotografiar(t, raiz)

			err := skills.Escribir(raiz, regenerado)
			for _, texto := range caso.defectos {
				require.ErrorContains(t, err, texto)
			}

			assert.Equal(t, antes, fotografiar(t, raiz), "con defectos, Escribir no escribe nada")
		})
	}
}

// TestRegenerarSinPoderListar fija los errores con los que Regenerar no regenera
// nada, sobre la estructura de un repositorio temporal con las fuentes de la
// sincronía: un fichero donde va el directorio de skills y un fichero donde va el
// directorio de datos, que se lista porque alfa declara una referencia. Cada
// error nombra el directorio.
func TestRegenerarSinPoderListar(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre     string
		directorio string
		prefijo    string
	}{
		{nombre: "directorio-de-skills-que-es-un-fichero", directorio: "skills", prefijo: "el directorio de skills "},
		{nombre: "directorio-de-datos-que-es-un-fichero", directorio: "data", prefijo: "alfa: el directorio de datos "},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := arbolDeFuentes(t)
			ruta := filepath.Join(raiz, caso.directorio)
			sustituirPorUnFichero(t, ruta)

			regenerado, err := skills.Regenerar(raiz, descripcionesDeSincronia())
			require.ErrorIs(t, err, syscall.ENOTDIR)
			require.ErrorContains(t, err, caso.prefijo+ruta+" no se puede listar: ")
			assert.Zero(t, regenerado)
		})
	}
}

// TestCompararYEscribirSinBuscarLasDerivas fija los errores que impiden buscar
// las derivas de una skill (data-model §5), sobre la estructura de un repositorio
// temporal sincronizado: un fichero donde va references/ de alfa, y un fichero
// donde va el propio directorio de alfa. Comparar y Escribir reciben lo
// regenerado como un dato y miran el árbol tal como está cuando se las llama, así
// que en el último caso lo regenerado es de antes del cambio: con el fichero,
// Regenerar ya no listaría alfa. Las dos dan el mismo error, que nombra la skill y
// la carpeta, y Escribir no escribe nada.
func TestCompararYEscribirSinBuscarLasDerivas(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string

		// partes es la ruta del fichero dentro del directorio de alfa; vacía, es el
		// propio directorio.
		partes []string

		// error es el error entero o, con causa, su principio.
		error string
		causa error
	}{
		{nombre: "references-que-es-un-fichero", partes: []string{"references"}, error: "alfa: references no es un directorio"},
		{nombre: "skill-que-es-un-fichero", error: "alfa: references no se puede consultar: ", causa: syscall.ENOTDIR},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := arbolSincronizado(t)
			regenerado := regenerar(t, raiz, descripcionesDeSincronia())

			sustituirPorUnFichero(t, rutaDeSkill(raiz, "alfa", caso.partes...))
			envejecer(t, raiz)
			antes := fotografiar(t, raiz)

			derivas, errDeComparar := skills.Comparar(raiz, regenerado)
			assert.Nil(t, derivas)

			for _, err := range []error{errDeComparar, skills.Escribir(raiz, regenerado)} {
				if caso.causa == nil {
					require.EqualError(t, err, caso.error)

					continue
				}

				require.ErrorIs(t, err, caso.causa)
				require.ErrorContains(t, err, caso.error)
			}

			assert.Equal(t, antes, fotografiar(t, raiz), "sin poder buscar las derivas, Escribir no escribe nada")
		})
	}
}

// TestEscribirSinAplicarUnArreglo fija los errores de Escribir al deshacer una
// deriva, cada uno sobre la estructura de un repositorio temporal y con la skill,
// la ruta y lo que no se puede hacer:
//
//   - retirar: references/normas.md es un directorio con un fichero dentro;
//   - crear el directorio: el directorio de alfa es ahora un enlace colgante, así
//     que ninguna de sus rutas existe y el directorio no se puede crear donde
//     está el enlace.
//
// En los dos, lo regenerado es de antes del cambio: con el directorio, Regenerar
// daría el defecto de FR-006 de H31, y no lista un enlace como skill.
func TestEscribirSinAplicarUnArreglo(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string

		// preparar deja el repositorio de raíz listo para Escribir y devuelve lo
		// regenerado que se le da.
		preparar func(t *testing.T, raiz string) skills.Regenerado

		prefijo string
		causa   error
	}{
		{
			nombre: "retirar-un-directorio-con-contenido",
			preparar: func(t *testing.T, raiz string) skills.Regenerado {
				t.Helper()

				require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, descripcionesDeSincronia())))
				regenerado := regenerar(t, raiz, descripcionesDeSincronia())

				ruta := rutaDeSkill(raiz, "alfa", "references", "normas.md")
				retirarDePrueba(t, ruta)
				escribirFicheroDePrueba(t, filepath.Join(ruta, "nota.md"), "# Nota\n")

				return regenerado
			},
			prefijo: "alfa: references/normas.md no se puede retirar: ",
			causa:   syscall.ENOTEMPTY,
		},
		{
			nombre: "crear-el-directorio-de-un-enlace-colgante",
			preparar: func(t *testing.T, raiz string) skills.Regenerado {
				t.Helper()

				regenerado := regenerar(t, raiz, descripcionesDeSincronia())

				ruta := rutaDeSkill(raiz, "alfa")
				require.NoError(t, os.RemoveAll(ruta))
				crearEnlaceDePrueba(t, "no-existe", ruta)

				return regenerado
			},
			prefijo: "alfa: SKILL.md: el directorio ",
			causa:   fs.ErrExist,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := arbolDeFuentes(t)
			regenerado := caso.preparar(t, raiz)
			require.Empty(t, presentarDefectos(regenerado.Defectos()))

			err := skills.Escribir(raiz, regenerado)
			require.ErrorIs(t, err, caso.causa)
			assert.True(t, strings.HasPrefix(err.Error(), caso.prefijo), "error: %v", err)
		})
	}
}

// sustituirPorUnFichero retira lo que haya en la ruta, con lo que tenga dentro, y
// escribe en su lugar un fichero.
func sustituirPorUnFichero(t *testing.T, ruta string) {
	t.Helper()

	require.NoError(t, os.RemoveAll(ruta))
	escribirFicheroDePrueba(t, ruta, "no es un directorio\n")
}

// arbolDeFuentes es un repositorio temporal con las fuentes de la sincronía y sin
// nada generado.
func arbolDeFuentes(t *testing.T) string {
	t.Helper()

	raiz := t.TempDir()
	escribirFicheroDePrueba(t, filepath.Join(raiz, "data", "normas.yaml"), normasDeSincronia)
	escribirFicheroDePrueba(t, filepath.Join(raiz, "data", "jerarquia.yaml"), jerarquiaDeSincronia)
	escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "SKILL.md"), skillMdDeAlfaSinSincronizar)
	escribirFicheroDePrueba(t, rutaDeSkill(raiz, "beta", "SKILL.md"), skillMdDeBetaSinNadaQueGenerar)

	return raiz
}

// arbolSincronizado es un repositorio temporal con las fuentes de la sincronía y
// lo que Escribir genera desde ellas: sin defectos ni derivas.
func arbolSincronizado(t *testing.T) string {
	t.Helper()

	raiz := arbolDeFuentes(t)
	require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, descripcionesDeSincronia())))
	require.Empty(t, compararSinDefectos(t, raiz, descripcionesDeSincronia()))

	return raiz
}

// regenerar es lo que Regenerar construye para el repositorio de raíz.
func regenerar(t *testing.T, raiz string, descripciones []skills.DescripcionDeVerbo) skills.Regenerado {
	t.Helper()

	regenerado, err := skills.Regenerar(raiz, descripciones)
	require.NoError(t, err)

	return regenerado
}

// compararSinDefectos regenera el repositorio de raíz, exige que ninguna skill
// tenga defectos —sin ellos no se compararía nada— y devuelve sus derivas.
func compararSinDefectos(t *testing.T, raiz string, descripciones []skills.DescripcionDeVerbo) []*skills.Deriva {
	t.Helper()

	regenerado := regenerar(t, raiz, descripciones)
	require.Empty(t, presentarDefectos(regenerado.Defectos()))

	derivas, err := skills.Comparar(raiz, regenerado)
	require.NoError(t, err)

	return derivas
}

// presentarDefectos son los textos de los defectos, en su orden; nil si no hay
// ninguno.
func presentarDefectos(defectos []*skills.DefectoDeSkill) []string {
	var textos []string

	for _, defecto := range defectos {
		textos = append(textos, defecto.Error())
	}

	return textos
}

// rutaDeSkill es la ruta de las partes dentro del directorio de la skill del
// repositorio de raíz.
func rutaDeSkill(raiz, skill string, partes ...string) string {
	return filepath.Join(append([]string{raiz, "skills", skill}, partes...)...)
}

// leerFicheroDePrueba es el contenido del fichero de la ruta, dentro de un árbol
// temporal del test.
func leerFicheroDePrueba(t *testing.T, ruta string) string {
	t.Helper()

	contenido, err := os.ReadFile(filepath.Clean(ruta))
	require.NoError(t, err)

	return string(contenido)
}

// cambiarFicheroDePrueba sustituye en el fichero de la ruta la primera aparición
// de viejo por nuevo. Lo lee y lo escribe con auxiliares distintos.
func cambiarFicheroDePrueba(t *testing.T, ruta, viejo, nuevo string) {
	t.Helper()

	escribirFicheroDePrueba(t, ruta, cambiada(t, leerFicheroDePrueba(t, ruta), viejo, nuevo))
}

// retirarLasMarcas quita del SKILL.md de la ruta las líneas de las dos marcas de
// la región de la tabla de comandos.
func retirarLasMarcas(t *testing.T, ruta string) {
	t.Helper()

	cambiarFicheroDePrueba(t, ruta, inicioDeLaTabla+"\n", "")
	cambiarFicheroDePrueba(t, ruta, finDeLaTabla+"\n", "")
}

// cuerpoDePrueba es lo que sigue a la línea --- que cierra el frontmatter del
// SKILL.md de la ruta, que abre con la que lo abre.
func cuerpoDePrueba(t *testing.T, ruta string) string {
	t.Helper()

	contenido := leerFicheroDePrueba(t, ruta)
	tras, abierto := strings.CutPrefix(contenido, "---")
	_, cuerpo, cerrado := strings.Cut(tras, "\n---\n")
	require.True(t, abierto && cerrado, "%s abre con su frontmatter y lo cierra", ruta)

	return cuerpo
}

// alargarElCuerpo añade líneas en blanco al final del SKILL.md de la ruta, que
// termina en salto de línea, hasta que su cuerpo tiene las líneas dadas.
func alargarElCuerpo(t *testing.T, ruta string, lineas int) {
	t.Helper()

	actuales := skills.ContarLineas([]byte(cuerpoDePrueba(t, ruta)))
	contenido := leerFicheroDePrueba(t, ruta)

	require.True(t, strings.HasSuffix(contenido, "\n"))
	require.LessOrEqual(t, actuales, lineas)

	escribirFicheroDePrueba(t, ruta, contenido+strings.Repeat("\n", lineas-actuales))
	require.Equal(t, lineas, skills.ContarLineas([]byte(cuerpoDePrueba(t, ruta))))
}

// ensancharElCuerpo añade al final del SKILL.md de la ruta, que termina en salto
// de línea, una línea de eñes, de dos bytes cada una, hasta que su cuerpo tiene
// los caracteres dados, y exige que entonces tenga más bytes que caracteres.
func ensancharElCuerpo(t *testing.T, ruta string, caracteres int) {
	t.Helper()

	actuales := utf8.RuneCountInString(cuerpoDePrueba(t, ruta))
	contenido := leerFicheroDePrueba(t, ruta)

	require.True(t, strings.HasSuffix(contenido, "\n"))
	require.Less(t, actuales, caracteres)

	escribirFicheroDePrueba(t, ruta, contenido+strings.Repeat("ñ", caracteres-actuales-1)+"\n")

	cuerpo := cuerpoDePrueba(t, ruta)
	require.Equal(t, caracteres, utf8.RuneCountInString(cuerpo))
	require.Greater(t, len(cuerpo), caracteres)
}

// crearEnlaceDePrueba crea en la ruta un enlace simbólico con el destino literal
// dado, creando antes lo que falte del directorio que lo contiene.
func crearEnlaceDePrueba(t *testing.T, destino, ruta string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(ruta), 0o750))
	require.NoError(t, os.Symlink(destino, ruta))
}

// retirarDePrueba retira el fichero o el enlace de la ruta.
func retirarDePrueba(t *testing.T, ruta string) {
	t.Helper()

	require.NoError(t, os.Remove(ruta))
}

// entradaDelArbol es lo que fotografiar guarda de una entrada del árbol: su tipo,
// sus bytes si es un fichero, su destino si es un enlace y su tiempo de
// modificación.
type entradaDelArbol struct {
	Tipo       fs.FileMode
	Contenido  string
	Destino    string
	Modificada time.Time
}

// fotografiar es cada entrada del árbol de raíz, sin seguir los enlaces, por su
// ruta dentro de él.
func fotografiar(t *testing.T, raiz string) map[string]entradaDelArbol {
	t.Helper()

	arbol := abrirArbol(t, raiz)
	foto := map[string]entradaDelArbol{}

	err := fs.WalkDir(arbol.FS(), ".", func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		estado, err := entrada.Info()
		if err != nil {
			return err
		}

		guardada := entradaDelArbol{Tipo: estado.Mode().Type(), Modificada: estado.ModTime()}

		switch {
		case estado.Mode().IsRegular():
			var contenido []byte
			contenido, err = arbol.ReadFile(filepath.FromSlash(ruta))
			guardada.Contenido = string(contenido)
		case estado.Mode()&fs.ModeSymlink != 0:
			guardada.Destino, err = arbol.Readlink(filepath.FromSlash(ruta))
		}

		foto[ruta] = guardada

		return err
	})
	require.NoError(t, err)

	return foto
}

// envejecer pone a cada fichero y directorio del árbol de raíz un tiempo de
// modificación del pasado, salvo a los enlaces, que no se pueden fechar sin
// seguirlos: así una reescritura, o una entrada creada o retirada en un
// directorio, cambia lo que fotografiar guarda.
func envejecer(t *testing.T, raiz string) {
	t.Helper()

	arbol := abrirArbol(t, raiz)
	pasado := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)

	err := fs.WalkDir(arbol.FS(), ".", func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil || entrada.Type()&fs.ModeSymlink != 0 {
			return err
		}

		return arbol.Chtimes(filepath.FromSlash(ruta), pasado, pasado)
	})
	require.NoError(t, err)
}

// abrirArbol abre el árbol de raíz como un os.Root, que no deja que una ruta
// salga de él por un enlace mientras se lee o se fecha, y lo cierra al terminar
// el test.
func abrirArbol(t *testing.T, raiz string) *os.Root {
	t.Helper()

	arbol, err := os.OpenRoot(raiz)
	require.NoError(t, err)

	t.Cleanup(func() { assert.NoError(t, arbol.Close()) })

	return arbol
}
