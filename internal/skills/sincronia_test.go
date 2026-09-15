package skills_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// Las fuentes del repositorio temporal de TestRegenerarYComparar, sin nada
// generado: la tabla de normas y el SKILL.md de dos skills. alfa declara dos
// applets, en otro orden que el del registro, y la referencia de las normas, y su
// tabla de comandos está escrita a mano; beta no declara nada, así que no lleva
// tabla, ni referencias, ni enlaces.
const (
	normasDeSincronia = "normas:\n" +
		"  BOE-A-2015-10565:\n" +
		"    titulo: Norma de prueba de la sincronía.\n" +
		"    rango: Ley\n" +
		"    materias:\n" +
		"      - procedimiento\n" +
		"  BOE-A-1985-5392:\n" +
		"    titulo: Otra norma de prueba de la sincronía.\n" +
		"    rango: Ley\n" +
		"    abreviatura: ONP\n" +
		"    materias:\n" +
		"      - régimen local\n"

	// cabeceraDeAlfa es el SKILL.md de alfa hasta la marca de inicio de su tabla,
	// incluida: la marca está en la línea 10.
	cabeceraDeAlfa = "---\n" +
		"name: alfa\n" +
		"description: Skill de prueba con dos applets y una referencia.\n" +
		"metadata:\n" +
		"  kitlegal-applets: dos uno\n" +
		"  kitlegal-referencias: normas\n" +
		"---\n" +
		"# Alfa\n" +
		"\n" +
		inicioDeLaTabla + "\n"

	// pieDeAlfa es el SKILL.md de alfa desde la marca de fin de su tabla.
	pieDeAlfa = finDeLaTabla + "\n" +
		"\n" +
		"## Reglas\n"

	skillMdDeAlfaSinSincronizar = cabeceraDeAlfa + "tabla escrita a mano\n" + pieDeAlfa

	skillMdDeBetaSinNadaQueGenerar = "---\n" +
		"name: beta\n" +
		"description: Skill de prueba sin nada que generar.\n" +
		"---\n" +
		"# Beta\n"
)

// Lo que la sincronía genera para alfa desde esas fuentes y desde
// descripcionesDeSincronia, byte a byte: la región de su tabla de comandos, con
// una sección por applet en el orden de su declaración (contrato
// sincronizacion-y-comprobacion §3), y la referencia de las normas, por año y
// número del identificador (contrato normas-y-referencias §5).
const (
	regionDeAlfa = "\n" +
		"### `scripts/dos`\n" +
		"\n" +
		columnasDeLaTabla +
		"| `scripts/dos listar` | Lista los bloques. | lista de objetos con `bloque` |\n" +
		"\n" +
		"### `scripts/uno`\n" +
		"\n" +
		columnasDeLaTabla +
		"| `scripts/uno leer <bloque>` | Lee un bloque. | objeto con `texto` |\n" +
		"\n" +
		"Todas devuelven el sobre `ok`, `data`; con `ok` falso, `data` lleva `clase` y `mensaje`.\n" +
		"\n" +
		"Banderas comunes: `--json`, `--timeout <valor>`.\n" +
		"\n"

	skillMdDeAlfaSincronizado = cabeceraDeAlfa + regionDeAlfa + pieDeAlfa

	referenciaDeAlfa = comienzoDeLaReferencia +
		"| Otra norma de prueba de la sincronía. | ONP | `BOE-A-1985-5392` | Ley | régimen local |\n" +
		"| Norma de prueba de la sincronía. |  | `BOE-A-2015-10565` | Ley | procedimiento |\n"
)

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
// research.md D4, D5 y D7; FR-034, FR-035, FR-036, FR-042, SC-005): lo generado
// sale de la declaración de cada skill y de data/, sin nada escrito para una skill
// concreta; Regenerar y Comparar no escriben nada; Escribir deja el árbol sin
// derivas, y un segundo Escribir no cambia ningún byte, ningún enlace ni ningún
// tiempo de modificación; cada clase de deriva nombra la skill y el fichero o el
// enlace, y Escribir la deshace; y cada defecto de una skill nombra la skill, la
// deja sin nada que comparar e impide que Escribir escriba nada.
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
				Referencias: []skills.Referencia{{Fichero: "normas.md", Contenido: []byte(referenciaDeAlfa)}},
				Enlaces: []skills.Enlace{
					{Nombre: "dos", Destino: destinoDeLosEnlacesDePrueba},
					{Nombre: "uno", Destino: destinoDeLosEnlacesDePrueba},
				},
			},
			{Nombre: "beta", Contenido: []byte(skillMdDeBetaSinNadaQueGenerar)},
		}}, regenerado)

		derivas, err := skills.Comparar(raiz, regenerado)
		require.NoError(t, err)
		assert.Equal(t, []*skills.Deriva{
			{Skill: "alfa", Ruta: "SKILL.md", Clase: skills.DerivaContenidoDistinto},
			{Skill: "alfa", Ruta: "references/normas.md", Clase: skills.DerivaFicheroAusente},
			{Skill: "alfa", Ruta: "scripts/dos", Clase: skills.DerivaEnlaceAusente},
			{Skill: "alfa", Ruta: "scripts/uno", Clase: skills.DerivaEnlaceAusente},
		}, derivas)
		assert.Equal(t, antes, fotografiar(t, raiz), "Regenerar y Comparar no escriben nada")

		require.NoError(t, skills.Escribir(raiz, regenerado))
		assert.Empty(t, compararSinDefectos(t, raiz, descripcionesDeSincronia()))

		assert.Equal(t, skillMdDeAlfaSincronizado, leerFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "SKILL.md")))
		assert.Equal(t, referenciaDeAlfa, leerFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "normas.md")))

		for _, applet := range []string{"dos", "uno"} {
			destino, err := os.Readlink(rutaDeSkill(raiz, "alfa", "scripts", applet))
			require.NoError(t, err)
			assert.Equal(t, destinoDeLosEnlacesDePrueba, destino, "el enlace de %s lleva el destino literal", applet)
		}

		assert.Equal(t, skillMdDeBetaSinNadaQueGenerar, leerFicheroDePrueba(t, rutaDeSkill(raiz, "beta", "SKILL.md")))
		assert.NoDirExists(t, rutaDeSkill(raiz, "beta", "references"), "beta no declara referencias")
		assert.NoDirExists(t, rutaDeSkill(raiz, "beta", "scripts"), "beta no declara applets")
	})

	t.Run("segundo-escribir-no-cambia-nada", func(t *testing.T) {
		t.Parallel()

		raiz := arbolDeFuentes(t)
		escribirFicheroDePrueba(t, rutaDeSkill(raiz, "beta", "references", "sobrante.md"), "# Sobrante\n")
		crearEnlaceDePrueba(t, "../../../bin/kitlegal", rutaDeSkill(raiz, "alfa", "scripts", "uno"))
		escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "scripts", "dos"), "no es un enlace\n")
		crearEnlaceDePrueba(t, destinoDeLosEnlacesDePrueba, rutaDeSkill(raiz, "beta", "scripts", "tres"))

		require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, descripcionesDeSincronia())))
		require.Empty(t, compararSinDefectos(t, raiz, descripcionesDeSincronia()))

		envejecer(t, raiz)
		antes := fotografiar(t, raiz)

		require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, descripcionesDeSincronia())))
		assert.Equal(t, antes, fotografiar(t, raiz),
			"el segundo Escribir no cambia ningún byte, ningún enlace ni ningún tiempo de modificación")
	})

	t.Run("doscientas-noventa-y-nueve-lineas", func(t *testing.T) {
		t.Parallel()

		raiz := arbolSincronizado(t)
		alargarSkillMd(t, rutaDeSkill(raiz, "alfa", "SKILL.md"), 299)

		assert.Empty(t, compararSinDefectos(t, raiz, descripcionesDeSincronia()))
	})

	probarDerivas(t)
	probarDefectos(t)
}

// probarDerivas fija cada clase de deriva de data-model §5: sobre un árbol
// sincronizado con un solo cambio, Comparar da exactamente esa deriva, que nombra
// la skill y el fichero o el enlace, y Escribir la deshace sin dejar ningún
// defecto.
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
			nombre: "datos-sin-regenerar",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, filepath.Join(raiz, "data", "normas.yaml"),
					"titulo: Norma de prueba de la sincronía.", "titulo: Norma de prueba con otro título.")
			},
			deriva:  skills.Deriva{Skill: "alfa", Ruta: "references/normas.md", Clase: skills.DerivaContenidoDistinto},
			mensaje: "alfa: references/normas.md: contenido-distinto",
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
			nombre: "referencia-que-no-es-un-fichero-regular",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				ruta := rutaDeSkill(raiz, "alfa", "references", "normas.md")
				retirarDePrueba(t, ruta)
				crearEnlaceDePrueba(t, "../../../data/normas.yaml", ruta)
			},
			deriva: skills.Deriva{
				Skill:   "alfa",
				Ruta:    "references/normas.md",
				Clase:   skills.DerivaContenidoDistinto,
				Detalle: "no es un fichero regular",
			},
			mensaje: "alfa: references/normas.md: contenido-distinto (no es un fichero regular)",
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

				escribirFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "references", "antigua.md"), "# Antigua\n")
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
		{
			nombre: "enlace-ausente",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				retirarDePrueba(t, rutaDeSkill(raiz, "alfa", "scripts", "uno"))
			},
			deriva:  skills.Deriva{Skill: "alfa", Ruta: "scripts/uno", Clase: skills.DerivaEnlaceAusente},
			mensaje: "alfa: scripts/uno: enlace-ausente",
		},
		{
			nombre: "enlace-sobrante",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				crearEnlaceDePrueba(t, destinoDeLosEnlacesDePrueba, rutaDeSkill(raiz, "alfa", "scripts", "tres"))
			},
			deriva:  skills.Deriva{Skill: "alfa", Ruta: "scripts/tres", Clase: skills.DerivaEnlaceSobrante},
			mensaje: "alfa: scripts/tres: enlace-sobrante",
		},
		{
			nombre: "enlace-en-una-skill-que-no-declara-applets",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				crearEnlaceDePrueba(t, destinoDeLosEnlacesDePrueba, rutaDeSkill(raiz, "beta", "scripts", "uno"))
			},
			deriva:  skills.Deriva{Skill: "beta", Ruta: "scripts/uno", Clase: skills.DerivaEnlaceSobrante},
			mensaje: "beta: scripts/uno: enlace-sobrante",
		},
		{
			nombre: "enlace-con-otro-destino",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				ruta := rutaDeSkill(raiz, "alfa", "scripts", "uno")
				retirarDePrueba(t, ruta)
				crearEnlaceDePrueba(t, "../../../bin/kitlegal", ruta)
			},
			deriva: skills.Deriva{
				Skill:   "alfa",
				Ruta:    "scripts/uno",
				Clase:   skills.DerivaEnlaceConOtroDestino,
				Detalle: "apunta a ../../../bin/kitlegal",
			},
			mensaje: "alfa: scripts/uno: enlace-con-otro-destino (apunta a ../../../bin/kitlegal)",
		},
		{
			nombre: "fichero-regular-en-lugar-de-enlace",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				ruta := rutaDeSkill(raiz, "alfa", "scripts", "dos")
				retirarDePrueba(t, ruta)
				escribirFicheroDePrueba(t, ruta, "#!/bin/sh\n")
			},
			deriva: skills.Deriva{
				Skill:   "alfa",
				Ruta:    "scripts/dos",
				Clase:   skills.DerivaEnlaceConOtroDestino,
				Detalle: "no es un enlace simbólico",
			},
			mensaje: "alfa: scripts/dos: enlace-con-otro-destino (no es un enlace simbólico)",
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

// probarDefectos fija los defectos de una skill (data-model §5): cada uno nombra
// la skill y van en el orden frontmatter, líneas, región y datos; una skill con
// defectos no tiene nada regenerado con que comparar, aunque las demás sí; y
// Escribir no escribe nada, tampoco lo de las skills sin defectos.
func probarDefectos(t *testing.T) {
	t.Helper()

	casos := []struct {
		nombre               string
		alterar              func(t *testing.T, raiz string)
		cambiarDescripciones func(descripciones []skills.DescripcionDeVerbo)
		defectos             []string
	}{
		{
			nombre: "frontmatter",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "SKILL.md"), "name: alfa\n", "name: otra\n")
			},
			defectos: []string{`alfa: name "otra" distinto del nombre del directorio`},
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
			defectos: []string{`alfa: metadata/kitlegal-applets: applet "tres" no registrado`},
		},
		{
			nombre: "referencia-sin-generador",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "alfa", "SKILL.md"),
					"kitlegal-referencias: normas\n", "kitlegal-referencias: normas otras\n")
				escribirFicheroDePrueba(t, filepath.Join(raiz, "data", "otras.yaml"), "otras: []\n")
			},
			defectos: []string{`alfa: metadata/kitlegal-referencias: "otras" sin generador conocido`},
		},
		{
			nombre: "trescientas-lineas",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				alargarSkillMd(t, rutaDeSkill(raiz, "alfa", "SKILL.md"), 300)
			},
			defectos: []string{"alfa: SKILL.md tiene 300 líneas (máximo 299)"},
		},
		{
			// El límite es el del SKILL.md regenerado: con la región de la tabla
			// vacía, el del árbol tiene menos de 300 líneas, tantas menos como la
			// tabla, y la tabla que se regenera lo lleva a 300.
			nombre: "trescientas-lineas-al-regenerar",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				ruta := rutaDeSkill(raiz, "alfa", "SKILL.md")
				cambiarFicheroDePrueba(t, ruta, regionDeAlfa, "")
				alargarSkillMd(t, ruta, 300-strings.Count(regionDeAlfa, "\n"))
			},
			defectos: []string{"alfa: SKILL.md tiene 300 líneas (máximo 299)"},
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
			nombre: "norma-con-vertical",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarFicheroDePrueba(t, filepath.Join(raiz, "data", "normas.yaml"),
					"      - procedimiento\n", "      - procedimiento\n    vertical: fiscal\n")
			},
			defectos: []string{"alfa: data/normas.yaml: BOE-A-2015-10565: campo no declarado: vertical"},
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
			nombre: "frontmatter-y-lineas-en-su-orden",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				ruta := rutaDeSkill(raiz, "alfa", "SKILL.md")
				cambiarFicheroDePrueba(t, ruta, "name: alfa\n", "name: otra\n")
				alargarSkillMd(t, ruta, 300)
			},
			defectos: []string{
				`alfa: name "otra" distinto del nombre del directorio`,
				"alfa: SKILL.md tiene 300 líneas (máximo 299)",
			},
		},
		{
			nombre: "lineas-y-region-en-su-orden",
			alterar: func(t *testing.T, raiz string) {
				t.Helper()

				ruta := rutaDeSkill(raiz, "alfa", "SKILL.md")
				retirarLasMarcas(t, ruta)
				alargarSkillMd(t, ruta, 300)
			},
			defectos: []string{
				"alfa: SKILL.md tiene 300 líneas (máximo 299)",
				"alfa: SKILL.md: sin las marcas de la tabla de comandos",
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := arbolSincronizado(t)
			if caso.alterar != nil {
				caso.alterar(t, raiz)
			}

			// beta no tiene defectos y sí una deriva, que Comparar sigue viendo y que
			// Escribir tampoco deshace.
			crearEnlaceDePrueba(t, destinoDeLosEnlacesDePrueba, rutaDeSkill(raiz, "beta", "scripts", "uno"))
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
			assert.Equal(t, []*skills.Deriva{{Skill: "beta", Ruta: "scripts/uno", Clase: skills.DerivaEnlaceSobrante}},
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
// temporal sincronizado: un fichero donde va references/ o scripts/ de alfa, y un
// fichero donde va el propio directorio de alfa. Comparar y Escribir reciben lo
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
		{nombre: "scripts-que-es-un-fichero", partes: []string{"scripts"}, error: "alfa: scripts no es un directorio"},
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

// skillMdDeBetaConElAppletUno es el SKILL.md de beta cuando declara el applet
// uno, como alfa, con la región de su tabla vacía.
const skillMdDeBetaConElAppletUno = "---\n" +
	"name: beta\n" +
	"description: Skill de prueba con uno de los applets de alfa.\n" +
	"metadata:\n" +
	"  kitlegal-applets: uno\n" +
	"---\n" +
	"# Beta\n" +
	"\n" +
	inicioDeLaTabla + "\n" +
	finDeLaTabla + "\n"

// TestEscribirSinAplicarUnArreglo fija los errores de Escribir al deshacer una
// deriva, cada uno sobre la estructura de un repositorio temporal y con la skill,
// la ruta y lo que no se puede hacer:
//
//   - retirar: references/normas.md es un directorio con un fichero dentro;
//   - crear el directorio: el directorio de alfa es ahora un enlace colgante, así
//     que ninguna de sus rutas existe y el directorio no se puede crear donde
//     está el enlace;
//   - enlazar: alfa y beta declaran el applet uno, y el directorio de beta es
//     ahora un enlace al de alfa; Escribir busca las derivas de las dos antes de
//     deshacer ninguna, y cuando llega a las de beta, alfa ya ha creado
//     scripts/uno.
//
// En los dos últimos, lo regenerado es de antes del cambio: Regenerar no lista un
// enlace como skill.
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

				ruta := rutaDeSkill(raiz, "alfa", "references", "normas.md")
				retirarDePrueba(t, ruta)
				escribirFicheroDePrueba(t, filepath.Join(ruta, "nota.md"), "# Nota\n")

				return regenerar(t, raiz, descripcionesDeSincronia())
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
		{
			nombre: "enlazar-lo-que-otra-skill-ya-enlazo",
			preparar: func(t *testing.T, raiz string) skills.Regenerado {
				t.Helper()

				escribirFicheroDePrueba(t, rutaDeSkill(raiz, "beta", "SKILL.md"), skillMdDeBetaConElAppletUno)
				regenerado := regenerar(t, raiz, descripcionesDeSincronia())

				ruta := rutaDeSkill(raiz, "beta")
				require.NoError(t, os.RemoveAll(ruta))
				crearEnlaceDePrueba(t, "alfa", ruta)

				return regenerado
			},
			prefijo: "beta: scripts/uno no se puede enlazar: ",
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

// alargarSkillMd añade líneas en blanco al final del SKILL.md de la ruta, que
// termina en salto de línea, hasta que tiene las líneas dadas.
func alargarSkillMd(t *testing.T, ruta string, lineas int) {
	t.Helper()

	contenido := leerFicheroDePrueba(t, ruta)
	require.True(t, strings.HasSuffix(contenido, "\n"))
	require.LessOrEqual(t, skills.ContarLineas([]byte(contenido)), lineas)

	escribirFicheroDePrueba(t, ruta, contenido+strings.Repeat("\n", lineas-skills.ContarLineas([]byte(contenido))))
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
