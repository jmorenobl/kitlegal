package app

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/render"
	"github.com/jmorenobl/kitlegal/internal/skills"
)

// regenerarSkills es la bandera con la que TestSkillsDelRepositorio escribe en
// skills/, antes de comparar, lo que regenera: las referencias, la tabla de
// comandos de SKILL.md y los enlaces de scripts/. Solo la pasa
// scripts/skills-sync.sh, que es lo que ejecuta make skills-sync; make
// skills-check nunca la pasa (contrato sincronizacion-y-comprobacion §1 y §2).
var regenerarSkills = flag.Bool("regenerar-skills", false,
	"escribe en skills/ las referencias, la tabla de comandos de SKILL.md y los enlaces de scripts/ antes de compararlos")

const (
	// raizDelRepositorio es la raíz del repositorio, relativa al directorio de
	// este paquete, que es donde go test ejecuta sus tests.
	raizDelRepositorio = "../.."

	// skillDelHito es la skill que el repositorio tiene desde H5: sin ella, las
	// comprobaciones sobre skills/ pasarían en vacío (plan, obligación 12).
	skillDelHito = "boe-legislacion"

	// destinoDeLosEnlacesDeScripts es el destino literal de todo enlace de
	// scripts/ de una skill (data-model §3).
	destinoDeLosEnlacesDeScripts = "../../../bin/instalado/kitlegal"

	// inicioDeLaTablaDeComandos y finDeLaTablaDeComandos son las dos líneas que
	// delimitan la región generada de SKILL.md (data-model §2).
	inicioDeLaTablaDeComandos = "<!-- inicio de la tabla de comandos: generada desde --describe con make skills-sync, " +
		"no editar -->"
	finDeLaTablaDeComandos = "<!-- fin de la tabla de comandos -->"
)

// TestSkillsDelRepositorio es lo que vigila make skills-check sobre skills/ y lo
// que escribe make skills-sync (contrato sincronizacion-y-comprobacion §2;
// research.md D2, D6 y D7; FR-030 a FR-036, FR-040 a FR-042): describe cada
// verbo de cada applet que declara alguna skill con la misma función que atiende
// --describe, regenera en memoria lo generado de cada skill y lo compara con su
// árbol. Cada defecto de una skill falla nombrando la skill y el defecto, y cada
// deriva, la skill y el fichero o el enlace. Con -regenerar-skills escribe antes
// lo regenerado, salvo si alguna skill tiene defectos, que se presentan igual.
//
// El primer subtest compara el árbol real y exige la skill del hito; los
// siguientes rompen, de uno en uno, copias temporales del árbol real y exigen los
// fallos exactos, de modo que ninguna comprobación pasa en vacío; y los dos
// últimos comprueban las normas que nombra SKILL.md (FR-020) y lo que la skill
// no puede decir (FR-077).
func TestSkillsDelRepositorio(t *testing.T) {
	t.Parallel()

	descripciones := descripcionesDeLasSkills(t, raizDelRepositorio)

	// Antes de los subtests, que copian el árbol real: tienen que copiar lo ya
	// escrito.
	if *regenerarSkills {
		escribirSkills(t, raizDelRepositorio, descripciones)
	}

	t.Run("skills", func(t *testing.T) {
		t.Parallel()

		nombres, err := skills.Listar(raizDelRepositorio)
		require.NoError(t, err)
		require.Contains(t, nombres, skillDelHito, "skills/ tiene la skill del hito")

		assert.Empty(t, fallosDeLasSkills(t, raizDelRepositorio, descripciones))
	})

	probarSobreCopias(t, descripciones, casosDeDeriva())

	t.Run("trescientas-lineas", func(t *testing.T) {
		t.Parallel()

		probarSobreCopias(t, descripciones, casosDeLineas())
	})

	t.Run("frontmatter", func(t *testing.T) {
		t.Parallel()

		probarSobreCopias(t, descripciones, casosDeFrontmatter())
	})

	t.Run("enlaces", func(t *testing.T) {
		t.Parallel()

		probarSobreCopias(t, descripciones, casosDeEnlaces())
	})

	t.Run("region", func(t *testing.T) {
		t.Parallel()

		skillMd := leerFicheroDelArbol(t, rutaEnLaSkill(raizDelRepositorio, skillDelHito, "SKILL.md"))
		probarSobreCopias(t, descripciones, casosDeRegion(
			lineaDeLaMarca(t, skillMd, inicioDeLaTablaDeComandos), lineaDeLaMarca(t, skillMd, finDeLaTablaDeComandos)))
	})

	t.Run("regenerar-dos-veces", func(t *testing.T) {
		t.Parallel()

		probarRegenerarDosVeces(t, descripciones)
	})

	t.Run("normas-nombradas", func(t *testing.T) {
		t.Parallel()

		probarNormasNombradas(t)
	})

	t.Run("sin-instrucciones-de-evals", func(t *testing.T) {
		t.Parallel()

		probarSinInstruccionesDeEvals(t)
	})
}

// TestTablaDeComandosCoincideConLaGramatica vigila la única inferencia de la
// tabla de comandos que --describe no declara: que un argumento obligatorio va
// por su posición (contrato sincronizacion-y-comprobacion §3; data-model §2.2;
// research.md D6). Para cada verbo del registro de producción, la sintaxis de su
// fila en la tabla generada se convierte en la invocación mínima que la cumple,
// seguida de --describe, y esa invocación termina en 0 describiendo ese verbo.
// El último subtest demuestra que la comprobación no pasa en vacío: con un
// argumento obligatorio presentado como opcional, la gramática rechaza la
// invocación que sale de la tabla.
func TestTablaDeComandosCoincideConLaGramatica(t *testing.T) {
	t.Parallel()

	registro, err := RegistroDeProduccion()
	require.NoError(t, err)

	descripciones := describirApplets(t, registro, registro.Nombres())
	verbos := verbosDeProduccion(t)
	ordenes := sintaxisDeLaTabla(t, registro.Nombres(), descripciones)
	require.Len(t, ordenes, len(verbos), "la tabla tiene una fila por verbo del registro")

	for indice, orden := range ordenes {
		t.Run(verbos[indice], func(t *testing.T) {
			t.Parallel()

			argv := invocacionDeLaSintaxis(t, orden)

			res := invocar(t, registro, argv...)
			require.Equal(t, 0, res.codigo, "%s: %s", strings.Join(argv, " "), res.errores)

			documento, err := objetoJSON([]byte(res.salida))
			require.NoError(t, err)
			assert.Equal(t, verbos[indice], documento["title"], "%s describe su verbo", strings.Join(argv, " "))
		})
	}

	t.Run("obligatorio-presentado-como-opcional", func(t *testing.T) {
		t.Parallel()

		cambiadas := slices.Clone(descripciones)
		articulo := indiceDelVerbo(t, cambiadas, "boe", "articulo")
		cambiadas[articulo].Argumentos = slices.Clone(cambiadas[articulo].Argumentos)

		bloque := slices.IndexFunc(cambiadas[articulo].Argumentos, func(argumento skills.Argumento) bool {
			return argumento.Nombre == "bloque"
		})
		require.GreaterOrEqual(t, bloque, 0, "boe articulo declara el argumento bloque")

		cambiadas[articulo].Argumentos[bloque].Obligatorio = false

		orden := sintaxisDeLaTabla(t, registro.Nombres(), cambiadas)[articulo]
		require.Equal(t, "scripts/boe articulo <norma> [--bloque]", orden)

		res := invocar(t, registro, invocacionDeLaSintaxis(t, orden)...)
		assert.Equal(t, 2, res.codigo, "la gramática exige el bloque por su posición: %s", res.errores)
	})
}

// casoSobreUnaCopia es un cambio sobre una copia temporal del árbol real —en sus
// ficheros, en las descripciones de los verbos o en los dos— y los fallos exactos
// que da, en su orden; ninguno si la copia sigue sin defectos ni derivas.
type casoSobreUnaCopia struct {
	nombre               string
	alterar              func(t *testing.T, copia string)
	cambiarDescripciones func(t *testing.T, descripciones []skills.DescripcionDeVerbo)
	fallos               []string
}

// probarSobreCopias ejecuta cada caso como un subtest sobre su propia copia del
// árbol real y con su propia copia de las descripciones.
func probarSobreCopias(t *testing.T, descripciones []skills.DescripcionDeVerbo, casos []casoSobreUnaCopia) {
	t.Helper()

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			copia := copiaDelRepositorio(t)
			usadas := slices.Clone(descripciones)

			if caso.alterar != nil {
				caso.alterar(t, copia)
			}

			if caso.cambiarDescripciones != nil {
				caso.cambiarDescripciones(t, usadas)
			}

			assert.Equal(t, caso.fallos, fallosDeLasSkills(t, copia, usadas))
		})
	}
}

// casosDeDeriva son las derivas de lo generado desde sus tres fuentes: la
// referencia editada a mano, los datos cambiados sin regenerar y lo que declara
// --describe cambiado sin regenerar (US5 escenarios 2 y 3, SC-005).
func casosDeDeriva() []casoSobreUnaCopia {
	referenciaDistinta := []string{"boe-legislacion: references/normas.md: contenido-distinto"}

	return []casoSobreUnaCopia{
		{
			nombre: "referencia-editada",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				ruta := rutaEnLaSkill(copia, skillDelHito, "references", "normas.md")
				escribirFicheroDeLaCopia(t, ruta, leerFicheroDelArbol(t, ruta)+"| una fila escrita a mano | | | | |\n")
			},
			fallos: referenciaDistinta,
		},
		{
			nombre: "datos-sin-regenerar",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				cambiarFicheroDeLaCopia(t, filepath.Join(copia, "data", "normas.yaml"),
					"Ley 39/2015, de 1 de octubre,", "Ley 39/2015, de 2 de octubre,")
			},
			fallos: referenciaDistinta,
		},
		{
			nombre: "describe-cambiado",
			cambiarDescripciones: func(t *testing.T, descripciones []skills.DescripcionDeVerbo) {
				t.Helper()

				articulo := indiceDelVerbo(t, descripciones, "boe", "articulo")
				descripciones[articulo].Hace = "Devuelve el texto de un bloque de una norma."
			},
			fallos: []string{"boe-legislacion: SKILL.md: contenido-distinto"},
		},
	}
}

// casosDeLineas son el límite de líneas de SKILL.md (FR-041; data-model §1.2):
// 299 no es un defecto y 300 sí.
func casosDeLineas() []casoSobreUnaCopia {
	return []casoSobreUnaCopia{
		{nombre: "doscientas-noventa-y-nueve", alterar: alargarSkillMd(299)},
		{
			nombre:  "trescientas",
			alterar: alargarSkillMd(300),
			fallos:  []string{"boe-legislacion: SKILL.md tiene 300 líneas (máximo 299)"},
		},
	}
}

// alargarSkillMd añade líneas en blanco al final del SKILL.md de la skill del
// hito, que termina en salto de línea, hasta que tiene las líneas dadas: fuera de
// la región generada, así que no cambia lo regenerado.
func alargarSkillMd(lineas int) func(t *testing.T, copia string) {
	return func(t *testing.T, copia string) {
		t.Helper()

		ruta := rutaEnLaSkill(copia, skillDelHito, "SKILL.md")
		contenido := leerFicheroDelArbol(t, ruta)
		actuales := skills.ContarLineas([]byte(contenido))

		require.True(t, strings.HasSuffix(contenido, "\n"), "SKILL.md termina en salto de línea")
		require.LessOrEqual(t, actuales, lineas, "SKILL.md no tiene ya más de %d líneas", lineas)

		escribirFicheroDeLaCopia(t, ruta, contenido+strings.Repeat("\n", lineas-actuales))
	}
}

// casosDeFrontmatter son las reglas del frontmatter (data-model §1.1 y §1.3;
// research.md D3 y D4; FR-040): una fila por defecto, cada una con un solo
// defecto que nombra la skill —el directorio de la skill se renombra cuando el
// caso necesita otro nombre—, y los límites exactos válidos, 64 caracteres en
// name y 1024 en description, contados como caracteres y no como bytes.
func casosDeFrontmatter() []casoSobreUnaCopia {
	const (
		name        = "name: boe-legislacion\n"
		description = "description: Consulta y cita normativa consolidada del BOE.\n"
		metadata    = "metadata:\n  kitlegal-applets: boe\n  kitlegal-referencias: normas\n"
	)

	nombreDe64, nombreDe65 := strings.Repeat("a", 64), strings.Repeat("a", 65)

	return []casoSobreUnaCopia{
		{
			nombre:  "sin-name",
			alterar: conOtroFrontmatter(skillDelHito, description+metadata),
			fallos:  []string{"boe-legislacion: falta name"},
		},
		{
			nombre:  "name-distinto-del-directorio",
			alterar: conOtroFrontmatter(skillDelHito, "name: otra-skill\n"+description+metadata),
			fallos:  []string{`boe-legislacion: name "otra-skill" distinto del nombre del directorio`},
		},
		{
			nombre:  "name-con-mayuscula",
			alterar: conOtroFrontmatter("Boe-legislacion", "name: Boe-legislacion\n"+description+metadata),
			fallos:  []string{`Boe-legislacion: name "Boe-legislacion" con caracteres que no son a-z, 0-9 ni -`},
		},
		{
			nombre:  "name-con-dos-guiones-seguidos",
			alterar: conOtroFrontmatter("boe--legislacion", "name: boe--legislacion\n"+description+metadata),
			fallos: []string{
				`boe--legislacion: name "boe--legislacion" con un guion al principio, al final o dos seguidos`,
			},
		},
		{
			nombre:  "name-que-empieza-por-guion",
			alterar: conOtroFrontmatter("-boe-legislacion", "name: \"-boe-legislacion\"\n"+description+metadata),
			fallos: []string{
				`-boe-legislacion: name "-boe-legislacion" con un guion al principio, al final o dos seguidos`,
			},
		},
		{
			nombre:  "name-de-65-caracteres",
			alterar: conOtroFrontmatter(nombreDe65, "name: "+nombreDe65+"\n"+description+metadata),
			fallos:  []string{nombreDe65 + ": name de 65 caracteres (máximo 64)"},
		},
		{
			nombre:  "sin-description",
			alterar: conOtroFrontmatter(skillDelHito, name+metadata),
			fallos:  []string{"boe-legislacion: falta description"},
		},
		{
			nombre:  "description-vacia",
			alterar: conOtroFrontmatter(skillDelHito, name+"description: \"\"\n"+metadata),
			fallos:  []string{"boe-legislacion: description vacía"},
		},
		{
			nombre:  "description-de-1025-caracteres",
			alterar: conOtroFrontmatter(skillDelHito, name+"description: "+strings.Repeat("ñ", 1025)+"\n"+metadata),
			fallos:  []string{"boe-legislacion: description de 1025 caracteres (máximo 1024)"},
		},
		{
			nombre:  "description-con-menor-que",
			alterar: conOtroFrontmatter(skillDelHito, name+"description: \"Consulta normas del BOE < 1978.\"\n"+metadata),
			fallos:  []string{"boe-legislacion: description con < o >"},
		},
		{
			nombre:  "clave-no-admitida",
			alterar: conOtroFrontmatter(skillDelHito, name+description+"version: \"1\"\n"+metadata),
			fallos:  []string{"boe-legislacion: clave no admitida: version"},
		},
		{
			nombre: "applet-que-no-existe",
			alterar: conOtroFrontmatter(skillDelHito,
				name+description+"metadata:\n  kitlegal-applets: boe inexistente\n  kitlegal-referencias: normas\n"),
			fallos: []string{`boe-legislacion: metadata/kitlegal-applets: applet "inexistente" no registrado`},
		},
		{
			nombre: "referencia-sin-sus-datos",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				retirarDeLaCopia(t, filepath.Join(copia, "data", "normas.yaml"))
			},
			fallos: []string{`boe-legislacion: metadata/kitlegal-referencias: "normas" sin data/normas.yaml`},
		},
		{
			nombre: "limites-validos",
			alterar: conOtroFrontmatter(nombreDe64,
				"name: "+nombreDe64+"\ndescription: "+strings.Repeat("ñ", 1024)+"\n"+metadata),
		},
	}
}

// conOtroFrontmatter deja la skill del hito en el directorio de nombre skill,
// renombrándolo si hace falta, con el frontmatter dado en su SKILL.md y el resto
// del fichero intacto.
func conOtroFrontmatter(skill, frontmatter string) func(t *testing.T, copia string) {
	return func(t *testing.T, copia string) {
		t.Helper()

		if skill != skillDelHito {
			require.NoError(t, os.Rename(rutaEnLaSkill(copia, skillDelHito), rutaEnLaSkill(copia, skill)))
		}

		ruta := rutaEnLaSkill(copia, skill, "SKILL.md")

		cuerpo, abre := strings.CutPrefix(leerFicheroDelArbol(t, ruta), "---\n")
		require.True(t, abre, "SKILL.md empieza por su frontmatter")

		_, resto, cierra := strings.Cut(cuerpo, "\n---\n")
		require.True(t, cierra, "el frontmatter de SKILL.md se cierra")

		escribirFicheroDeLaCopia(t, ruta, "---\n"+frontmatter+"---\n"+resto)
	}
}

// casosDeEnlaces son las clases de deriva de los enlaces de scripts/ (data-model
// §3 y §5; FR-036; US5 escenario 5).
func casosDeEnlaces() []casoSobreUnaCopia {
	return []casoSobreUnaCopia{
		{
			nombre: "enlace-ausente",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				retirarDeLaCopia(t, rutaEnLaSkill(copia, skillDelHito, "scripts", "boe"))
			},
			fallos: []string{"boe-legislacion: scripts/boe: enlace-ausente"},
		},
		{
			nombre: "enlace-sobrante",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				enlazarEnLaCopia(t, destinoDeLosEnlacesDeScripts, rutaEnLaSkill(copia, skillDelHito, "scripts", "cita"))
			},
			fallos: []string{"boe-legislacion: scripts/cita: enlace-sobrante"},
		},
		{
			nombre: "enlace-con-otro-destino",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				ruta := rutaEnLaSkill(copia, skillDelHito, "scripts", "boe")
				retirarDeLaCopia(t, ruta)
				enlazarEnLaCopia(t, "../../../bin/kitlegal", ruta)
			},
			fallos: []string{"boe-legislacion: scripts/boe: enlace-con-otro-destino (apunta a ../../../bin/kitlegal)"},
		},
		{
			nombre: "fichero-regular-en-lugar-de-enlace",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				ruta := rutaEnLaSkill(copia, skillDelHito, "scripts", "boe")
				retirarDeLaCopia(t, ruta)
				escribirFicheroDeLaCopia(t, ruta, "no es un enlace\n")
			},
			fallos: []string{"boe-legislacion: scripts/boe: enlace-con-otro-destino (no es un enlace simbólico)"},
		},
	}
}

// casosDeRegion son los defectos de las marcas de la región generada de SKILL.md
// (data-model §2; FR-032), con las líneas inicio y fin que ocupan las dos marcas
// en el SKILL.md real.
func casosDeRegion(inicio, fin int) []casoSobreUnaCopia {
	return []casoSobreUnaCopia{
		{
			nombre: "sin-marcas",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				ruta := rutaEnLaSkill(copia, skillDelHito, "SKILL.md")
				cambiarFicheroDeLaCopia(t, ruta, inicioDeLaTablaDeComandos+"\n", "")
				cambiarFicheroDeLaCopia(t, ruta, finDeLaTablaDeComandos+"\n", "")
			},
			fallos: []string{"boe-legislacion: SKILL.md: sin las marcas de la tabla de comandos"},
		},
		{
			// La marca de inicio repetida ocupa la línea de la de fin, que baja una.
			nombre: "dos-inicios",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				cambiarFicheroDeLaCopia(t, rutaEnLaSkill(copia, skillDelHito, "SKILL.md"), finDeLaTablaDeComandos+"\n",
					inicioDeLaTablaDeComandos+"\n"+finDeLaTablaDeComandos+"\n")
			},
			fallos: []string{fmt.Sprintf("boe-legislacion: SKILL.md: la marca de inicio de la tabla de comandos "+
				"aparece 2 veces, en las líneas %d y %d", inicio, fin)},
		},
		{
			nombre: "fin-antes-del-inicio",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				ruta := rutaEnLaSkill(copia, skillDelHito, "SKILL.md")
				contenido := leerFicheroDelArbol(t, ruta)

				antes, tras, hayInicio := strings.Cut(contenido, inicioDeLaTablaDeComandos+"\n")
				tabla, despues, hayFin := strings.Cut(tras, finDeLaTablaDeComandos+"\n")
				require.True(t, hayInicio && hayFin, "SKILL.md tiene las dos marcas, en su orden")

				escribirFicheroDeLaCopia(t, ruta, antes+finDeLaTablaDeComandos+"\n"+tabla+inicioDeLaTablaDeComandos+"\n"+despues)
			},
			fallos: []string{fmt.Sprintf("boe-legislacion: SKILL.md: la marca de fin de la tabla de comandos, en la "+
				"línea %d, está antes que la de inicio, en la línea %d", inicio, fin)},
		},
	}
}

// lineaDeLaMarca es el número de la línea del contenido que es exactamente la
// marca, que tiene que estar en una sola línea.
func lineaDeLaMarca(t *testing.T, contenido, marca string) int {
	t.Helper()

	require.Equal(t, 1, strings.Count(contenido, marca+"\n"), "la marca %q está en una sola línea", marca)

	antes, _, _ := strings.Cut(contenido, marca+"\n")

	return strings.Count(antes, "\n") + 1
}

// probarRegenerarDosVeces fija que la regeneración es idempotente (FR-034,
// FR-036): sobre una copia con una deriva de cada parte generada, Escribir la
// deja sin derivas, y un segundo Escribir no cambia ningún byte, ningún enlace ni
// ningún tiempo de modificación.
func probarRegenerarDosVeces(t *testing.T, descripciones []skills.DescripcionDeVerbo) {
	t.Helper()

	copia := copiaDelRepositorio(t)

	retirarDeLaCopia(t, rutaEnLaSkill(copia, skillDelHito, "references", "normas.md"))
	escribirFicheroDeLaCopia(t, rutaEnLaSkill(copia, skillDelHito, "references", "antigua.md"), "# Antigua\n")
	cambiarFicheroDeLaCopia(t, rutaEnLaSkill(copia, skillDelHito, "SKILL.md"), finDeLaTablaDeComandos+"\n",
		"tabla escrita a mano\n"+finDeLaTablaDeComandos+"\n")

	enlace := rutaEnLaSkill(copia, skillDelHito, "scripts", "boe")
	retirarDeLaCopia(t, enlace)
	enlazarEnLaCopia(t, "../../../bin/kitlegal", enlace)
	enlazarEnLaCopia(t, destinoDeLosEnlacesDeScripts, rutaEnLaSkill(copia, skillDelHito, "scripts", "cita"))

	require.Equal(t, []string{
		"boe-legislacion: SKILL.md: contenido-distinto",
		"boe-legislacion: references/normas.md: fichero-ausente",
		"boe-legislacion: references/antigua.md: fichero-sobrante",
		"boe-legislacion: scripts/boe: enlace-con-otro-destino (apunta a ../../../bin/kitlegal)",
		"boe-legislacion: scripts/cita: enlace-sobrante",
	}, fallosDeLasSkills(t, copia, descripciones))

	escribirSkills(t, copia, descripciones)
	require.Empty(t, fallosDeLasSkills(t, copia, descripciones), "Escribir deja la copia sin defectos ni derivas")

	envejecerArbol(t, copia)
	antes := fotografiarArbol(t, copia)

	escribirSkills(t, copia, descripciones)
	assert.Equal(t, antes, fotografiarArbol(t, copia),
		"el segundo Escribir no cambia ningún byte, ningún enlace ni ningún tiempo de modificación")
	assert.Empty(t, fallosDeLasSkills(t, copia, descripciones))
}

// normaNombrada es la forma en que SKILL.md nombra una norma por su rango, su
// número y su año: «Ley N/AAAA», «Ley Orgánica N/AAAA», «Real Decreto N/AAAA»,
// «Real Decreto-ley N/AAAA» o «Real Decreto Legislativo N/AAAA» (contrato
// sincronizacion-y-comprobacion §2). Los rangos más largos van antes que los que
// empiezan igual.
var normaNombrada = regexp.MustCompile(
	`\b(?:Ley Orgánica|Ley|Real Decreto Legislativo|Real Decreto-ley|Real Decreto) [0-9]+/[0-9]{4}\b`)

// probarNormasNombradas fija que toda norma que nombra el SKILL.md de una skill
// empieza el título de una norma de data/normas.yaml (FR-020): en el árbol real,
// que nombra alguna, y en una copia con cinco normas nombradas que la tabla no
// tiene, una de cada forma, y una que sí tiene, que no falla. Las cinco llevan el
// número 0, que no lleva ninguna norma: la tabla crece con cada hito —en H6, con
// la LEC 1/2000 y la LOPDGDD 3/2018, que eran dos de estos ejemplos— y un ejemplo
// que pudiera entrar en ella dejaría el caso sin fallo.
func probarNormasNombradas(t *testing.T) {
	t.Helper()

	skillMd := leerFicheroDelArbol(t, rutaEnLaSkill(raizDelRepositorio, skillDelHito, "SKILL.md"))
	require.NotEmpty(t, normaNombrada.FindAllString(skillMd, -1), "SKILL.md nombra alguna norma por su número y año")
	assert.Empty(t, fallosDeNormasNombradas(t, raizDelRepositorio))

	copia := copiaDelRepositorio(t)
	ruta := rutaEnLaSkill(copia, skillDelHito, "SKILL.md")
	escribirFicheroDeLaCopia(t, ruta, leerFicheroDelArbol(t, ruta)+"Ver la Ley 0/2000, la Ley Orgánica 0/2018, "+
		"el Real Decreto 0/2001, el Real Decreto-ley 0/2020, el Real Decreto Legislativo 0/2015 y el "+
		"Real Decreto Legislativo 2/2004.\n")

	const sinEntrada = "boe-legislacion: SKILL.md nombra «%s», que no empieza el título de ninguna norma de " +
		"data/normas.yaml"

	assert.Equal(t, []string{
		fmt.Sprintf(sinEntrada, "Ley 0/2000"),
		fmt.Sprintf(sinEntrada, "Ley Orgánica 0/2018"),
		fmt.Sprintf(sinEntrada, "Real Decreto 0/2001"),
		fmt.Sprintf(sinEntrada, "Real Decreto-ley 0/2020"),
		fmt.Sprintf(sinEntrada, "Real Decreto Legislativo 0/2015"),
	}, fallosDeNormasNombradas(t, copia))
}

// fallosDeNormasNombradas es un fallo por cada norma que nombra el SKILL.md de
// una skill del árbol de raíz sin que ninguna norma de su data/normas.yaml tenga
// un título que empiece así, en el orden de las skills y del texto.
func fallosDeNormasNombradas(t *testing.T, raiz string) []string {
	t.Helper()

	normas, err := skills.LeerNormas([]byte(leerFicheroDelArbol(t, filepath.Join(raiz, "data", "normas.yaml"))))
	require.NoError(t, err)

	var fallos []string

	for _, skill := range skillsDelArbol(t, raiz) {
		for _, nombrada := range normaNombrada.FindAllString(leerFicheroDelArbol(t, rutaEnLaSkill(raiz, skill, "SKILL.md")), -1) {
			empieza := slices.ContainsFunc(normas, func(norma skills.Norma) bool {
				return empiezaElTitulo(norma.Titulo, nombrada)
			})
			if !empieza {
				fallos = append(fallos, fmt.Sprintf("%s: SKILL.md nombra «%s», que no empieza el título de ninguna "+
					"norma de data/normas.yaml", skill, nombrada))
			}
		}
	}

	return fallos
}

// empiezaElTitulo dice si el título empieza por la norma nombrada y lo que sigue
// no es otra cifra: «Ley 1/2000» no empieza «Ley 1/20001».
func empiezaElTitulo(titulo, nombrada string) bool {
	resto, empieza := strings.CutPrefix(titulo, nombrada)

	return empieza && (resto == "" || resto[0] < '0' || resto[0] > '9')
}

// instruccionDeEvals es lo que ni SKILL.md, fuera de la región generada, ni
// references/ pueden decir (contrato skill-boe-legislacion §2.5; FR-077, SC-012):
// sin distinguir mayúsculas y como palabra completa, delimitada por lo que no es
// letra ni cifra, de modo que «evalúa» no cuenta. El modelo lleva [0-9]+ para que
// un número de más de una cifra también sea una palabra completa.
var instruccionDeEvals = regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])` +
	`(?:KITLEGAL_CACHE_DIR|evals?|job|GitHub Actions|claude-[a-z]+-[0-9]+|siempre[^.\n]*--offline)` +
	`(?:[^\p{L}\p{N}]|$)`)

// probarSinInstruccionesDeEvals fija que ninguna skill del árbol real dice cómo
// se mide (FR-077), y sobre una copia que cada término cuenta en SKILL.md y en
// references/, y que no cuentan «evalúa», el «siempre» de otra frase ni lo que va
// dentro de la región generada.
func probarSinInstruccionesDeEvals(t *testing.T) {
	t.Helper()

	assert.Empty(t, fallosDeInstruccionesDeEvals(t, raizDelRepositorio))

	copia := copiaDelRepositorio(t)

	skillMd := rutaEnLaSkill(copia, skillDelHito, "SKILL.md")
	cambiarFicheroDeLaCopia(t, skillMd, finDeLaTablaDeComandos+"\n",
		"Un job dentro de la tabla generada no cuenta.\n"+finDeLaTablaDeComandos+"\n")

	contenido := leerFicheroDelArbol(t, skillMd)
	primera := skills.ContarLineas([]byte(contenido)) + 1
	prohibidas := []string{
		"Fija KITLEGAL_CACHE_DIR antes de consultar.",
		"Esta skill se mide con una eval.",
		"Las EVALS se lanzan aparte.",
		"Lo lanza un job semanal.",
		"Corre en GitHub Actions.",
		"Responde como claude-haiku-4-5.",
		"Consulta siempre con `--offline`.",
	}
	admitidas := []string{
		"Evalúa si falta contexto.",
		"Cita siempre. Sin caché con `--offline`, dilo.",
	}
	escribirFicheroDeLaCopia(t, skillMd, contenido+strings.Join(slices.Concat(prohibidas, admitidas), "\n")+"\n")

	referencia := rutaEnLaSkill(copia, skillDelHito, "references", "normas.md")
	tabla := leerFicheroDelArbol(t, referencia)
	escribirFicheroDeLaCopia(t, referencia, tabla+"| evals | | | | |\n")

	esperados := make([]string, 0, len(prohibidas)+1)
	for indice, linea := range prohibidas {
		esperados = append(esperados, fmt.Sprintf("boe-legislacion: SKILL.md:%d: %s", primera+indice, linea))
	}

	esperados = append(esperados, fmt.Sprintf("boe-legislacion: references/normas.md:%d: | evals | | | | |",
		skills.ContarLineas([]byte(tabla))+1))

	assert.Equal(t, esperados, fallosDeInstruccionesDeEvals(t, copia))
}

// fallosDeInstruccionesDeEvals es un fallo por cada línea de las skills del
// árbol de raíz que dice lo que instruccionDeEvals describe: las del SKILL.md de
// cada una fuera de la región generada y las de cada fichero .md de su
// references/, en ese orden.
func fallosDeInstruccionesDeEvals(t *testing.T, raiz string) []string {
	t.Helper()

	var fallos []string

	for _, skill := range skillsDelArbol(t, raiz) {
		skillMd := leerFicheroDelArbol(t, rutaEnLaSkill(raiz, skill, "SKILL.md"))
		fallos = append(fallos, lineasConInstruccionesDeEvals(skill, "SKILL.md", skillMd, true)...)

		for _, referencia := range referenciasDeLaSkill(t, raiz, skill) {
			contenido := leerFicheroDelArbol(t, rutaEnLaSkill(raiz, skill, "references", referencia))
			fallos = append(fallos, lineasConInstruccionesDeEvals(skill, "references/"+referencia, contenido, false)...)
		}
	}

	return fallos
}

// lineasConInstruccionesDeEvals son las líneas del contenido del fichero de la
// skill que dicen lo que instruccionDeEvals describe, cada una con la skill, el
// fichero y su número; con fueraDeLaTabla, sin las de la región generada.
func lineasConInstruccionesDeEvals(skill, fichero, contenido string, fueraDeLaTabla bool) []string {
	var (
		lineas    []string
		enLaTabla bool
	)

	for indice, linea := range strings.Split(contenido, "\n") {
		switch {
		case fueraDeLaTabla && linea == inicioDeLaTablaDeComandos:
			enLaTabla = true
		case fueraDeLaTabla && linea == finDeLaTablaDeComandos:
			enLaTabla = false
		case !enLaTabla && instruccionDeEvals.MatchString(linea):
			lineas = append(lineas, fmt.Sprintf("%s: %s:%d: %s", skill, fichero, indice+1, linea))
		}
	}

	return lineas
}

// referenciasDeLaSkill son los nombres de los ficheros .md del references/ de
// la skill, en orden de nombre, o ninguno si no lo tiene.
func referenciasDeLaSkill(t *testing.T, raiz, skill string) []string {
	t.Helper()

	entradas, err := os.ReadDir(rutaEnLaSkill(raiz, skill, "references"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	require.NoError(t, err)

	var referencias []string

	for _, entrada := range entradas {
		if strings.HasSuffix(entrada.Name(), ".md") {
			referencias = append(referencias, entrada.Name())
		}
	}

	return referencias
}

// descripcionesDeLasSkills son las descripciones de los verbos de los applets
// que declaran las skills del árbol de raíz, desde el registro de producción.
func descripcionesDeLasSkills(t *testing.T, raiz string) []skills.DescripcionDeVerbo {
	t.Helper()

	registro, err := RegistroDeProduccion()
	require.NoError(t, err)

	return describirApplets(t, registro, appletsDeclarados(t, raiz))
}

// appletsDeclarados son los applets que declara en kitlegal-applets alguna skill
// del árbol de raíz, cada uno una vez y en el orden en que aparece por primera
// vez (contrato sincronizacion-y-comprobacion §2, paso 1). Una skill sin SKILL.md
// o con un frontmatter que no se puede leer no declara ninguno: Regenerar lo
// presenta como defecto de esa skill.
func appletsDeclarados(t *testing.T, raiz string) []string {
	t.Helper()

	var applets []string

	for _, nombre := range skillsDelArbol(t, raiz) {
		skill, err := skills.Cargar(raiz, nombre)

		var defecto *skills.DefectoDeSkill
		if errors.As(err, &defecto) {
			continue
		}

		require.NoError(t, err)

		frontmatter, err := skills.LeerFrontmatter(skill.Contenido)
		if err != nil {
			continue
		}

		for _, applet := range frontmatter.DeclaracionDeKitlegal().Applets {
			if !slices.Contains(applets, applet) {
				applets = append(applets, applet)
			}
		}
	}

	return applets
}

// describirApplets son las descripciones de cada verbo de los applets dados, en
// su orden y en el de sus verbos, leídas de lo que emite para cada uno describir
// —la función que atiende --describe— sobre un presentador con búferes, y de lo
// que emite cli.Describir para un verbo sin argumentos, que da las banderas
// globales. Un applet que no está registrado no tiene verbos que describir:
// Regenerar lo presenta como defecto de la skill que lo declara.
func describirApplets(t *testing.T, registro *Registro, applets []string) []skills.DescripcionDeVerbo {
	t.Helper()

	var globales, erroresDeGlobales bytes.Buffer
	require.NoError(t, cli.Describir(render.Nuevo(&globales, &erroresDeGlobales), cli.Verbo{}),
		erroresDeGlobales.String())

	var descripciones []skills.DescripcionDeVerbo

	for _, nombre := range applets {
		applet, registrado := registro.Buscar(nombre)
		if !registrado {
			continue
		}

		for _, verbo := range applet.Verbos() {
			var salida, errores bytes.Buffer
			require.NoError(t, describir(render.Nuevo(&salida, &errores), applet, verbo.Nombre),
				"%s %s --describe: %s", nombre, verbo.Nombre, errores.String())

			descripcion, err := skills.LeerDescripcionDeVerbo(salida.Bytes(), globales.Bytes())
			require.NoError(t, err)

			descripciones = append(descripciones, descripcion)
		}
	}

	return descripciones
}

// indiceDelVerbo es la posición en las descripciones de la del verbo del applet,
// que tiene que estar.
func indiceDelVerbo(t *testing.T, descripciones []skills.DescripcionDeVerbo, applet, verbo string) int {
	t.Helper()

	indice := slices.IndexFunc(descripciones, func(descripcion skills.DescripcionDeVerbo) bool {
		return descripcion.Applet == applet && descripcion.Verbo == verbo
	})
	require.GreaterOrEqual(t, indice, 0, "hay descripción de %s %s", applet, verbo)

	return indice
}

// sintaxisDeLaTabla son las órdenes de las filas de la tabla de comandos que se
// genera para los applets, en su orden: el texto de la primera celda, sin las
// comillas de código.
func sintaxisDeLaTabla(t *testing.T, applets []string, descripciones []skills.DescripcionDeVerbo) []string {
	t.Helper()

	tabla, err := skills.RenderizarTabla(applets, descripciones)
	require.NoError(t, err)

	var ordenes []string

	for linea := range strings.Lines(string(tabla)) {
		celda, esFila := strings.CutPrefix(linea, "| `")
		if !esFila {
			continue
		}

		orden, _, cerrada := strings.Cut(celda, "` |")
		require.True(t, cerrada, "la primera celda de %q se cierra", linea)

		ordenes = append(ordenes, orden)
	}

	return ordenes
}

// invocacionDeLaSintaxis es la invocación mínima que cumple una sintaxis de la
// tabla, seguida de --describe (contrato sincronizacion-y-comprobacion §3): el
// enlace del applet como nombre del programa, que es como lo invoca la skill, el
// verbo, x por cada argumento obligatorio, x y por cada uno de varios valores y
// ninguno de los opcionales.
func invocacionDeLaSintaxis(t *testing.T, orden string) []string {
	t.Helper()

	partes := strings.Fields(orden)
	require.GreaterOrEqual(t, len(partes), 2, "la sintaxis %q nombra el applet y el verbo", orden)

	argv := []string{partes[0], partes[1]}

	for _, parte := range partes[2:] {
		switch {
		case strings.HasPrefix(parte, "[--"):
		case strings.HasSuffix(parte, ">..."):
			argv = append(argv, "x", "y")
		default:
			argv = append(argv, "x")
		}
	}

	return append(argv, "--describe")
}

// fallosDeLasSkills regenera en memoria lo generado de las skills del árbol de
// raíz y lo compara con él, sin escribir nada: un texto por cada defecto de una
// skill, que nombra la skill y el defecto, y después uno por cada deriva, que
// nombra la skill y el fichero o el enlace; nil si no hay ninguno.
func fallosDeLasSkills(t *testing.T, raiz string, descripciones []skills.DescripcionDeVerbo) []string {
	t.Helper()

	regenerado, err := skills.Regenerar(raiz, descripciones)
	require.NoError(t, err)

	derivas, err := skills.Comparar(raiz, regenerado)
	require.NoError(t, err)

	var fallos []string

	for _, defecto := range regenerado.Defectos() {
		fallos = append(fallos, defecto.Error())
	}

	for _, deriva := range derivas {
		fallos = append(fallos, deriva.Error())
	}

	return fallos
}

// escribirSkills deja el árbol de raíz como lo regenerado desde él. Si alguna
// skill tiene defectos no escribe nada —Escribir tampoco lo haría— y los presenta
// fallosDeLasSkills.
func escribirSkills(t *testing.T, raiz string, descripciones []skills.DescripcionDeVerbo) {
	t.Helper()

	regenerado, err := skills.Regenerar(raiz, descripciones)
	require.NoError(t, err)

	if len(regenerado.Defectos()) == 0 {
		require.NoError(t, skills.Escribir(raiz, regenerado))
	}
}

// skillsDelArbol son los nombres de las skills del árbol de raíz.
func skillsDelArbol(t *testing.T, raiz string) []string {
	t.Helper()

	nombres, err := skills.Listar(raiz)
	require.NoError(t, err)

	return nombres
}

// rutaEnLaSkill es la ruta de las partes dentro del directorio de la skill del
// árbol de raíz.
func rutaEnLaSkill(raiz, skill string, partes ...string) string {
	return filepath.Join(append([]string{raiz, "skills", skill}, partes...)...)
}

// copiaDelRepositorio copia en un directorio temporal lo que Regenerar lee del
// árbol real —skills/ y data/— y devuelve su raíz. Los enlaces simbólicos se
// recrean con su destino literal, sin seguirlos: el de scripts/ no resuelve
// hasta que se instala. Los dos árboles se abren como os.Root, que no deja salir
// de ellos por un enlace mientras se copia.
func copiaDelRepositorio(t *testing.T) string {
	t.Helper()

	copia := t.TempDir()
	origen, destino := abrirArbol(t, raizDelRepositorio), abrirArbol(t, copia)

	for _, carpeta := range []string{"skills", "data"} {
		err := fs.WalkDir(origen.FS(), carpeta, func(ruta string, entrada fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			nativa := filepath.FromSlash(ruta)

			switch tipo := entrada.Type(); {
			case tipo.IsDir():
				return destino.MkdirAll(nativa, 0o750)
			case tipo&fs.ModeSymlink != 0:
				enlazado, err := origen.Readlink(nativa)
				if err != nil {
					return err
				}

				return destino.Symlink(enlazado, nativa)
			case tipo.IsRegular():
				contenido, err := origen.ReadFile(nativa)
				if err != nil {
					return err
				}

				return destino.WriteFile(nativa, contenido, 0o600)
			default:
				return fmt.Errorf("%s no es un directorio, un fichero regular ni un enlace simbólico", ruta)
			}
		})
		require.NoError(t, err)
	}

	return copia
}

// abrirArbol abre el árbol de raíz como un os.Root y lo cierra al terminar el
// test.
func abrirArbol(t *testing.T, raiz string) *os.Root {
	t.Helper()

	arbol, err := os.OpenRoot(raiz)
	require.NoError(t, err)

	t.Cleanup(func() { assert.NoError(t, arbol.Close()) })

	return arbol
}

// leerFicheroDelArbol es el contenido del fichero de la ruta, leído por un
// auxiliar distinto del que escribe (gosec G703) y con filepath.Clean (G304).
func leerFicheroDelArbol(t *testing.T, ruta string) string {
	t.Helper()

	contenido, err := os.ReadFile(filepath.Clean(ruta))
	require.NoError(t, err)

	return string(contenido)
}

// escribirFicheroDeLaCopia deja el contenido en el fichero de la ruta, dentro de
// una copia temporal, solo para su propietario (gosec G306).
func escribirFicheroDeLaCopia(t *testing.T, ruta, contenido string) {
	t.Helper()

	require.NoError(t, os.WriteFile(ruta, []byte(contenido), 0o600))
}

// cambiarFicheroDeLaCopia sustituye en el fichero de la ruta, dentro de una copia
// temporal, la primera aparición de viejo, que tiene que estar, por nuevo.
func cambiarFicheroDeLaCopia(t *testing.T, ruta, viejo, nuevo string) {
	t.Helper()

	escribirFicheroDeLaCopia(t, ruta, string(reemplazaUnaVez(t, []byte(leerFicheroDelArbol(t, ruta)), viejo, nuevo)))
}

// retirarDeLaCopia retira el fichero o el enlace de la ruta, dentro de una copia
// temporal.
func retirarDeLaCopia(t *testing.T, ruta string) {
	t.Helper()

	require.NoError(t, os.Remove(ruta))
}

// enlazarEnLaCopia crea en la ruta, dentro de una copia temporal, un enlace
// simbólico con el destino literal dado.
func enlazarEnLaCopia(t *testing.T, destino, ruta string) {
	t.Helper()

	require.NoError(t, os.Symlink(destino, ruta))
}

// entradaDeLaFoto es lo que fotografiarArbol guarda de una entrada: su tipo, sus
// bytes si es un fichero, su destino si es un enlace y su tiempo de modificación.
type entradaDeLaFoto struct {
	tipo       fs.FileMode
	contenido  string
	destino    string
	modificada time.Time
}

// fotografiarArbol es cada entrada del árbol de raíz, sin seguir los enlaces, por
// su ruta dentro de él.
func fotografiarArbol(t *testing.T, raiz string) map[string]entradaDeLaFoto {
	t.Helper()

	arbol := abrirArbol(t, raiz)
	foto := map[string]entradaDeLaFoto{}

	err := fs.WalkDir(arbol.FS(), ".", func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		estado, err := entrada.Info()
		if err != nil {
			return err
		}

		guardada := entradaDeLaFoto{tipo: estado.Mode().Type(), modificada: estado.ModTime()}
		nativa := filepath.FromSlash(ruta)

		switch {
		case estado.Mode().IsRegular():
			var contenido []byte
			contenido, err = arbol.ReadFile(nativa)
			guardada.contenido = string(contenido)
		case estado.Mode()&fs.ModeSymlink != 0:
			guardada.destino, err = arbol.Readlink(nativa)
		}

		foto[ruta] = guardada

		return err
	})
	require.NoError(t, err)

	return foto
}

// envejecerArbol pone a cada fichero y directorio del árbol de raíz un tiempo de
// modificación del pasado, salvo a los enlaces, que no se pueden fechar sin
// seguirlos: así una reescritura, o una entrada creada o retirada en un
// directorio, cambia lo que fotografiarArbol guarda.
func envejecerArbol(t *testing.T, raiz string) {
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
