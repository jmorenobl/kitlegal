package app

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"maps"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/render"
	"github.com/jmorenobl/kitlegal/internal/skills"
)

// regenerarSkills es la bandera con la que TestSkillsDelRepositorio escribe en
// skills/, antes de comparar, lo que regenera: las referencias y la tabla de
// comandos de SKILL.md. Solo la pasa scripts/skills-sync.sh, que es lo que
// ejecuta make skills-sync; make skills-check nunca la pasa (contrato
// sincronizacion-y-comprobacion §1 y §2).
var regenerarSkills = flag.Bool("regenerar-skills", false,
	"escribe en skills/ las referencias y la tabla de comandos de SKILL.md antes de compararlas")

const (
	// raizDelRepositorio es la raíz del repositorio, relativa al directorio de
	// este paquete, que es donde go test ejecuta sus tests.
	raizDelRepositorio = "../.."

	// programaDeLasOrdenes es con lo que empieza cada orden de la tabla de
	// comandos: el binario, invocado desde el PATH (contracts/skills-e-invocacion.md
	// §2 de H19; FR-080).
	programaDeLasOrdenes = "kitlegal"

	// inicioDeLaTablaDeComandos y finDeLaTablaDeComandos son las dos líneas que
	// delimitan la región generada de SKILL.md (data-model §2).
	inicioDeLaTablaDeComandos = "<!-- inicio de la tabla de comandos: generada desde --describe con make skills-sync, " +
		"no editar -->"
	finDeLaTablaDeComandos = "<!-- fin de la tabla de comandos -->"
)

// skillsExigidas son las skills que skills/ tiene que tener: sin una de ellas,
// las comprobaciones sobre skills/ pasarían en vacío para ella (plan, obligación
// 12). boe-legislacion la trajo H5; legal-core, H6, y jurisprudencia, H23. Los
// casos negativos no salen de esta lista, sino de las skills que hay
// (skillsDelRecorrido).
var skillsExigidas = []string{"boe-legislacion", "jurisprudencia", "legal-core"}

// cabeceraDeUnaReferencia es la primera línea de una referencia generada, que
// nombra el YAML de datos del que sale (FR-031, FR-065, FR-066).
var cabeceraDeUnaReferencia = regexp.MustCompile(`\A<!-- generado desde (data/[^,\n]+), no editar -->\n`)

// TestSkillsDelRepositorio es lo que vigila make skills-check sobre skills/ y lo
// que escribe make skills-sync (contrato sincronizacion-y-comprobacion §2;
// research.md D2, D6 y D7; FR-030 a FR-036, FR-040 a FR-042): describe cada
// verbo de cada applet que declara alguna skill con la misma función que atiende
// --describe, regenera en memoria lo generado de cada skill y lo compara con su
// árbol. Cada defecto de una skill —también llevar scripts/ (ADR 0019; FR-082 de
// H19)— falla nombrando la skill y el defecto, y cada deriva, la skill y el
// fichero. Con -regenerar-skills escribe antes lo regenerado, salvo si alguna
// skill tiene defectos, que se presentan igual.
//
// El primer subtest compara el árbol real y exige las skills exigidas; los
// siguientes rompen, de uno en uno y para cada skill del árbol, copias temporales
// con solo esa skill y exigen los fallos exactos, de modo que ninguna
// comprobación pasa en vacío para ninguna skill (research.md D26); y los dos
// últimos comprueban las normas que nombra la skill entera —SKILL.md y sus
// referencias a mano— (FR-020; FR-016 de H31) y lo que la skill no puede decir
// (FR-077), en el árbol real y sobre cada skill.
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
		require.Subset(t, nombres, skillsExigidas, "skills/ tiene las skills exigidas")

		assert.Empty(t, fallosDeLasSkills(t, raizDelRepositorio, descripciones))
	})

	t.Run("deriva", func(t *testing.T) {
		t.Parallel()

		probarSobreCopias(t, descripciones, casosDeDeriva)
	})

	t.Run("presupuestos", func(t *testing.T) {
		t.Parallel()

		probarSobreCopias(t, descripciones, casosDePresupuestos)
	})

	t.Run("frontmatter", func(t *testing.T) {
		t.Parallel()

		probarSobreCopias(t, descripciones, casosDeFrontmatter)
	})

	t.Run("referencias", func(t *testing.T) {
		t.Parallel()

		probarSobreCopias(t, descripciones, casosDeReferencias)
	})

	t.Run("scripts", func(t *testing.T) {
		t.Parallel()

		probarSobreCopias(t, descripciones, casosDeScripts)
	})

	t.Run("region", func(t *testing.T) {
		t.Parallel()

		probarSobreCopias(t, descripciones, casosDeRegion)
	})

	t.Run("regenerar-dos-veces", func(t *testing.T) {
		t.Parallel()

		probarCadaSkill(t, func(t *testing.T, skill skillDelRecorrido) {
			t.Helper()

			probarRegenerarDosVeces(t, descripciones, skill)
		})
	})

	t.Run("sincronia-con-referencias-a-mano", func(t *testing.T) {
		t.Parallel()

		probarCadaSkill(t, func(t *testing.T, skill skillDelRecorrido) {
			t.Helper()

			probarSincroniaConReferenciasAMano(t, descripciones, skill)
		})
	})

	t.Run("normas-nombradas", func(t *testing.T) {
		t.Parallel()

		probarNormasNombradasDelArbol(t)
		probarCadaSkill(t, probarNormasNombradas)
	})

	t.Run("sin-instrucciones-de-evals", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, fallosDeInstruccionesDeEvals(t, raizDelRepositorio))
		probarCadaSkill(t, probarSinInstruccionesDeEvals)
	})
}

// TestTablaDeComandosCoincideConLaGramatica vigila la única inferencia de la
// tabla de comandos que --describe no declara: que un argumento obligatorio va
// por su posición (contrato sincronizacion-y-comprobacion §3; data-model §2.2;
// research.md D6). Para cada verbo del registro de producción, la sintaxis de su
// fila en la tabla generada se convierte en la invocación mínima que la cumple,
// seguida de --describe, y esa invocación termina en 0 describiendo ese verbo.
// Desde H21, la fila lleva además la herramienta de ese mismo verbo, su título
// con un guion bajo en lugar del espacio (contracts/skills.md §4 de H21;
// FR-030). Desde H23, la fila escribe las banderas propias del verbo como
// banderas (research.md D14 de H23), y termina en 0 describiendo ese verbo
// también la invocación completa, la que escribe además todo lo opcional: es la
// que distingue una bandera de un argumento de posición opcional, que la mínima
// no escribe. Los tres últimos subtests demuestran que la comprobación no pasa
// en vacío: con argumentos obligatorios presentados como opcionales —uno solo, y
// uno seguido de otro de varios valores, anidados como los escribe Kong—, la
// gramática rechaza la invocación mínima que sale de la tabla; y con las banderas
// de un verbo presentadas como argumentos de posición opcionales, que es como
// las daría un documento sin la anotación, rechaza la completa.
func TestTablaDeComandosCoincideConLaGramatica(t *testing.T) {
	t.Parallel()

	registro, err := RegistroDeProduccion("")
	require.NoError(t, err)

	descripciones := describirApplets(t, registro, registro.Nombres())
	verbos := verbosDeProduccion(t)
	filas := filasDeLaTablaGenerada(t, registro.Nombres(), descripciones)
	require.Len(t, filas, len(verbos), "la tabla tiene una fila por verbo del registro")

	for indice, fila := range filas {
		t.Run(verbos[indice], func(t *testing.T) {
			t.Parallel()

			invocaciones := [][]string{
				invocacionDeLaSintaxis(t, fila.orden),
				invocacionCompletaDeLaSintaxis(t, fila.orden),
			}

			for _, argv := range invocaciones {
				res := invocar(t, registro, argv...)
				require.Equal(t, 0, res.codigo, "%s: %s", strings.Join(argv, " "), res.errores)

				documento, err := objetoJSON([]byte(res.salida))
				require.NoError(t, err)
				assert.Equal(t, verbos[indice], documento["title"], "%s describe su verbo", strings.Join(argv, " "))
			}

			assert.Equal(t, strings.ReplaceAll(verbos[indice], " ", separadorDeHerramienta), fila.herramienta,
				"la fila de %s nombra la herramienta de su verbo", verbos[indice])
		})
	}

	opcionales := []struct {
		nombre, verbo string
		argumentos    []string
		orden         string
	}{
		{
			nombre:     "obligatorio-presentado-como-opcional",
			verbo:      "articulo",
			argumentos: []string{"bloque"},
			// Desde H20, la fila de articulo escribe además su bandera --fecha.
			orden: "kitlegal boe articulo <norma> [<bloque>] [--fecha <fecha>]",
		},
		{
			nombre:     "obligatorio-y-lista-presentados-como-opcionales",
			verbo:      "articulos",
			argumentos: []string{"norma", "bloques"},
			orden:      "kitlegal boe articulos [<norma> [<bloques>...]]",
		},
	}

	for _, caso := range opcionales {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			probarObligatoriosComoOpcionales(t, registro, descripciones, caso.verbo, caso.argumentos, caso.orden)
		})
	}

	t.Run("banderas-presentadas-como-argumentos-de-posicion", func(t *testing.T) {
		t.Parallel()

		probarBanderasComoArgumentos(t, registro, descripciones)
	})
}

// probarBanderasComoArgumentos presenta las banderas propias de skills doctor
// como argumentos de posición opcionales, que es como las da la lectura de un
// documento de --describe sin la anotación de las banderas, comprueba que la
// fila de la tabla las escribe con la sintaxis de antes de H23 y que la gramática
// rechaza la invocación completa que sale de ella. La mínima, que no escribe
// nada opcional, termina en 0 con esa fila: por eso no basta.
func probarBanderasComoArgumentos(t *testing.T, registro *Registro, descripciones []skills.DescripcionDeVerbo) {
	t.Helper()

	cambiadas := slices.Clone(descripciones)
	indice := indiceDelVerbo(t, cambiadas, "skills", "doctor")
	require.NotEmpty(t, cambiadas[indice].BanderasPropias, "skills doctor declara banderas propias")

	cambiadas[indice].Argumentos = slices.Clone(cambiadas[indice].Argumentos)
	for _, bandera := range cambiadas[indice].BanderasPropias {
		cambiadas[indice].Argumentos = append(cambiadas[indice].Argumentos, skills.Argumento{Nombre: bandera.Nombre})
	}

	cambiadas[indice].BanderasPropias = nil

	orden := sintaxisDeLaTabla(t, registro.Nombres(), cambiadas)[indice]
	require.Equal(t, "kitlegal skills doctor [<global> [<dir>]]", orden)

	res := invocar(t, registro, invocacionDeLaSintaxis(t, orden)...)
	assert.Equal(t, 0, res.codigo, "la invocación mínima no escribe nada opcional: %s", res.errores)

	res = invocar(t, registro, invocacionCompletaDeLaSintaxis(t, orden)...)
	assert.Equal(t, 2, res.codigo, "skills doctor no tiene argumentos de posición: %s", res.errores)
}

// probarObligatoriosComoOpcionales presenta como opcionales los argumentos de un
// verbo de boe que la gramática exige, comprueba que la fila de la tabla los
// escribe como la sintaxis esperada y que la gramática rechaza la invocación que
// sale de ella.
func probarObligatoriosComoOpcionales(t *testing.T, registro *Registro, descripciones []skills.DescripcionDeVerbo,
	verbo string, argumentos []string, esperada string,
) {
	t.Helper()

	cambiadas := slices.Clone(descripciones)
	indice := indiceDelVerbo(t, cambiadas, "boe", verbo)
	cambiadas[indice].Argumentos = slices.Clone(cambiadas[indice].Argumentos)

	for _, nombre := range argumentos {
		argumento := slices.IndexFunc(cambiadas[indice].Argumentos, func(argumento skills.Argumento) bool {
			return argumento.Nombre == nombre
		})
		require.GreaterOrEqual(t, argumento, 0, "boe %s declara el argumento %s", verbo, nombre)
		require.True(t, cambiadas[indice].Argumentos[argumento].Obligatorio, "boe %s exige %s", verbo, nombre)

		cambiadas[indice].Argumentos[argumento].Obligatorio = false
	}

	orden := sintaxisDeLaTabla(t, registro.Nombres(), cambiadas)[indice]
	require.Equal(t, esperada, orden)

	res := invocar(t, registro, invocacionDeLaSintaxis(t, orden)...)
	assert.Equal(t, 2, res.codigo, "la gramática exige %s por su posición: %s", strings.Join(argumentos, " y "),
		res.errores)
}

// TestOrdenesDeLasSkillsEmpotradas fija que las skills que lleva dentro el
// binario solo mandan ejecutar lo que ese binario sabe hacer
// (contracts/skills-e-invocacion.md §3 de H19; FR-084, SC-014): cada orden de la
// región generada del SKILL.md de cada skill empotrada empieza por kitlegal, un
// applet del registro de producción y un verbo de ese applet. Desde H21, cada
// fila nombra además una herramienta, y esa herramienta es una de las que
// anuncia el servidor: las que da herramientasDe con el registro de producción
// (contracts/skills.md §5 de H21; FR-031, FR-077, SC-009). Lee lo empotrado y no
// el árbol, porque es lo que se instala. Para que no pase en vacío, las skills
// exigidas están empotradas, cada skill que declara applets tiene alguna orden
// y el servidor anuncia alguna herramienta; y el subtest de control demuestra
// que la comprobación falla con una orden por el enlace de la instalación
// anterior, sin verbo, con un applet sin registrar y con un verbo sin
// registrar —una orden inventada—, y con una herramienta inventada, con la de
// un applet registrado que el servidor no anuncia y con una fila que no nombra
// ninguna, y que no cuenta lo que queda fuera de la región.
func TestOrdenesDeLasSkillsEmpotradas(t *testing.T) {
	t.Parallel()

	registro, err := RegistroDeProduccion("")
	require.NoError(t, err)

	anunciadas := herramientasDelServidor(t, registro)

	empotradas, err := skillsEmpotradas()
	require.NoError(t, err)

	nombres := make([]string, 0, len(empotradas))

	for _, skill := range empotradas {
		nombres = append(nombres, skill.Nombre)

		t.Run(skill.Nombre, func(t *testing.T) {
			t.Parallel()

			skillMd := ficheroEmpotrado(t, skill, "SKILL.md")

			frontmatter, err := skills.LeerFrontmatter(skillMd)
			require.NoError(t, err)

			filas := filasDeLaRegion(t, string(skillMd))
			if len(frontmatter.DeclaracionDeKitlegal().Applets) > 0 {
				require.NotEmpty(t, filas, "la tabla de comandos de %s tiene alguna orden", skill.Nombre)
			}

			assert.Empty(t, ordenesSinRegistrar(registro, ordenesDe(filas)))
			assert.Empty(t, herramientasSinAnunciar(anunciadas, filas))
		})
	}

	require.Subset(t, nombres, skillsExigidas, "el binario lleva dentro las skills exigidas")

	t.Run("control", func(t *testing.T) {
		t.Parallel()

		skillMd := "| `kitlegal boe inventado` | `boe_inventado` | Fuera de la región: no cuenta. | nada |\n" +
			inicioDeLaTablaDeComandos + "\n" +
			"\n### `kitlegal boe`\n\n" +
			"| Orden | Herramienta | Qué hace | Qué devuelve en `data` |\n|---|---|---|---|\n" +
			"| `kitlegal boe articulo <norma> <bloque>` | `boe_articulo` | Registrada y anunciada. | objeto |\n" +
			"| `scripts/boe articulo <norma> <bloque>` | `boe_articulo` | Por el enlace de la instalación " +
			"anterior. | objeto |\n" +
			"| `kitlegal boe` | `boe_articulo` | Sin verbo. | objeto |\n" +
			"| `kitlegal inventado leer <bloque>` | `inventado_leer` | Applet sin registrar. | objeto |\n" +
			"| `kitlegal boe inventado <norma>` | `boe_inventado` | Verbo sin registrar. | objeto |\n" +
			"| `kitlegal boe indice <norma>` | `boe_inventada` | Herramienta inventada. | objeto |\n" +
			"| `kitlegal skills list` | `skills_list` | Registrada, y el servidor no la anuncia. | objeto |\n" +
			"| `kitlegal boe indice <norma>` | Sin herramienta. | objeto |\n" +
			finDeLaTablaDeComandos + "\n"

		filas := filasDeLaRegion(t, skillMd)
		require.Len(t, filas, 8, "solo cuentan las filas de la región")

		assert.Equal(t, []string{
			"«scripts/boe articulo <norma> <bloque>» no empieza por kitlegal <applet> <verbo>",
			"«kitlegal boe» no empieza por kitlegal <applet> <verbo>",
			"«kitlegal inventado leer <bloque>»: el applet inventado no está registrado",
			"«kitlegal boe inventado <norma>»: el applet boe no tiene el verbo inventado",
		}, ordenesSinRegistrar(registro, ordenesDe(filas)))

		assert.Equal(t, []string{
			"«kitlegal inventado leer <bloque>»: la herramienta inventado_leer no está entre las del servidor",
			"«kitlegal boe inventado <norma>»: la herramienta boe_inventado no está entre las del servidor",
			"«kitlegal boe indice <norma>»: la herramienta boe_inventada no está entre las del servidor",
			"«kitlegal skills list»: la herramienta skills_list no está entre las del servidor",
			"«kitlegal boe indice <norma>» no nombra ninguna herramienta",
		}, herramientasSinAnunciar(anunciadas, filas))
	})
}

// herramientasDelServidor son los nombres de las herramientas que anuncia el
// servidor MCP con el registro: las de herramientasDe, la función con la que
// `mcp serve` las construye (H21 FR-031). Anuncia alguna: sin ninguna, la
// comprobación de las tablas no podría pasar, pero tampoco diría por qué.
func herramientasDelServidor(t *testing.T, registro *Registro) []string {
	t.Helper()

	herramientas, err := herramientasDe(registro, schema.Contexto{}, slog.New(slog.DiscardHandler), io.Discard)
	require.NoError(t, err)
	require.NotEmpty(t, herramientas, "el servidor anuncia alguna herramienta")

	nombres := make([]string, 0, len(herramientas))
	for _, herramienta := range herramientas {
		nombres = append(nombres, herramienta.Nombre)
	}

	return nombres
}

// ficheroEmpotrado son los bytes del fichero de la ruta, relativa a su carpeta,
// de la skill empotrada, que tiene que estar.
func ficheroEmpotrado(t *testing.T, skill instalacion.SkillEmpotrada, ruta string) []byte {
	t.Helper()

	indice := slices.IndexFunc(skill.Ficheros, func(fichero instalacion.FicheroEmpotrado) bool {
		return fichero.Ruta == ruta
	})
	require.GreaterOrEqual(t, indice, 0, "la skill empotrada %s lleva %s", skill.Nombre, ruta)

	return skill.Ficheros[indice].Contenido
}

// filasDeLaRegion son las filas de la región generada del SKILL.md, entre sus dos
// marcas, en su orden; ninguna si no tiene las marcas en su orden.
func filasDeLaRegion(t *testing.T, skillMd string) []filaDeLaTabla {
	t.Helper()

	_, tras, conInicio := strings.Cut(skillMd, inicioDeLaTablaDeComandos+"\n")
	region, _, conFin := strings.Cut(tras, finDeLaTablaDeComandos+"\n")

	if !conInicio || !conFin {
		return nil
	}

	return filasDeLaTabla(t, region)
}

// herramientasSinAnunciar es un fallo por cada fila que no nombra ninguna
// herramienta o cuya herramienta no está entre las anunciadas, en su orden; nil
// si no hay ninguno.
func herramientasSinAnunciar(anunciadas []string, filas []filaDeLaTabla) []string {
	var fallos []string

	for _, fila := range filas {
		switch {
		case fila.herramienta == "":
			fallos = append(fallos, fmt.Sprintf("«%s» no nombra ninguna herramienta", fila.orden))
		case !slices.Contains(anunciadas, fila.herramienta):
			fallos = append(fallos, fmt.Sprintf("«%s»: la herramienta %s no está entre las del servidor", fila.orden,
				fila.herramienta))
		}
	}

	return fallos
}

// ordenesSinRegistrar es un fallo por cada orden que no empieza por kitlegal, un
// applet del registro y un verbo de ese applet, en su orden; nil si no hay
// ninguno.
func ordenesSinRegistrar(registro *Registro, ordenes []string) []string {
	var fallos []string

	for _, orden := range ordenes {
		partes := strings.Fields(orden)
		if len(partes) < 3 || partes[0] != programaDeLasOrdenes {
			fallos = append(fallos, fmt.Sprintf("«%s» no empieza por kitlegal <applet> <verbo>", orden))

			continue
		}

		applet, registrado := registro.Buscar(partes[1])
		if !registrado {
			fallos = append(fallos, fmt.Sprintf("«%s»: el applet %s no está registrado", orden, partes[1]))

			continue
		}

		conVerbo := slices.ContainsFunc(applet.Verbos(), func(verbo Verbo) bool {
			return verbo.Nombre == partes[2]
		})
		if !conVerbo {
			fallos = append(fallos, fmt.Sprintf("«%s»: el applet %s no tiene el verbo %s", orden, partes[1], partes[2]))
		}
	}

	return fallos
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

// skillDelRecorrido es una skill del árbol real con lo que declara su
// frontmatter: los casos negativos de cada skill se construyen desde aquí, y no
// desde una skill concreta (research.md D26).
type skillDelRecorrido struct {
	// nombre es el de su directorio.
	nombre string

	// applets son los de kitlegal-applets, en su orden.
	applets []string

	// referencias son las de kitlegal-referencias, en su orden.
	referencias []referenciaDelRecorrido
}

// referenciaDelRecorrido es una referencia generada de una skill del árbol real.
type referenciaDelRecorrido struct {
	// nombre es el que declara kitlegal-referencias.
	nombre string

	// datos es la ruta del YAML de datos del que sale, relativa a la raíz y con
	// barra, como la nombra su cabecera: data/normas.yaml.
	datos string

	// contenido son sus bytes en el árbol real.
	contenido string
}

// fichero es la ruta de la referencia dentro del directorio de su skill, con
// barra, como la nombran las derivas: references/<nombre>.md.
func (r referenciaDelRecorrido) fichero() string {
	return "references/" + r.nombre + ".md"
}

// referenciasDe son las referencias de la skill que salen del YAML de datos
// dado, en el orden de la declaración.
func (s skillDelRecorrido) referenciasDe(datos string) []referenciaDelRecorrido {
	var referencias []referenciaDelRecorrido

	for _, referencia := range s.referencias {
		if referencia.datos == datos {
			referencias = append(referencias, referencia)
		}
	}

	return referencias
}

// skillsDelRecorrido son las skills del árbol real, en el orden de sus nombres,
// cada una con lo que declara su frontmatter y, de cada referencia, el YAML de
// datos que nombra su cabecera. Cada subtest las lee por su cuenta, de modo que
// una skill que no se puede leer no impide que /skills presente sus defectos.
func skillsDelRecorrido(t *testing.T) []skillDelRecorrido {
	t.Helper()

	var recorrido []skillDelRecorrido

	for _, nombre := range skillsDelArbol(t, raizDelRepositorio) {
		cargada, err := skills.Cargar(raizDelRepositorio, nombre)
		require.NoError(t, err)

		frontmatter, err := skills.LeerFrontmatter(cargada.Contenido)
		require.NoError(t, err, "%s: el frontmatter de SKILL.md se lee", nombre)

		deKitlegal := frontmatter.DeclaracionDeKitlegal()
		recorrida := skillDelRecorrido{nombre: nombre, applets: deKitlegal.Applets}

		for _, declarada := range deKitlegal.Referencias {
			referencia := referenciaDelRecorrido{nombre: declarada}
			referencia.contenido = leerFicheroDelArbol(t, rutaEnLaSkill(raizDelRepositorio, nombre, referencia.fichero()))

			cabecera := cabeceraDeUnaReferencia.FindStringSubmatch(referencia.contenido)
			require.NotNil(t, cabecera, "%s: %s empieza por su cabecera", nombre, referencia.fichero())

			referencia.datos = cabecera[1]
			recorrida.referencias = append(recorrida.referencias, referencia)
		}

		recorrido = append(recorrido, recorrida)
	}

	return recorrido
}

// probarCadaSkill ejecuta probar sobre cada skill del recorrido, en un subtest
// con su nombre.
func probarCadaSkill(t *testing.T, probar func(t *testing.T, skill skillDelRecorrido)) {
	t.Helper()

	for _, skill := range skillsDelRecorrido(t) {
		t.Run(skill.nombre, func(t *testing.T) {
			t.Parallel()

			probar(t, skill)
		})
	}
}

// probarSobreCopias ejecuta, para cada skill del recorrido, cada uno de sus casos
// como un subtest sobre su propia copia del árbol real con solo esa skill y con
// su propia copia de las descripciones.
func probarSobreCopias(t *testing.T, descripciones []skills.DescripcionDeVerbo,
	casosDeLaSkill func(t *testing.T, skill skillDelRecorrido) []casoSobreUnaCopia,
) {
	t.Helper()

	probarCadaSkill(t, func(t *testing.T, skill skillDelRecorrido) {
		t.Helper()

		for _, caso := range casosDeLaSkill(t, skill) {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				copia := copiaConLaSkill(t, skill.nombre)
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
	})
}

// casosDeDeriva son las derivas de lo generado de la skill desde sus tres
// fuentes (US5 escenarios 2 y 3, SC-005): cada referencia editada a mano, los
// datos de cada referencia cambiados sin regenerar y lo que declara --describe de
// cada applet cambiado sin regenerar. Y las clases de FR-012 que son deriva
// (contracts/gates-de-skills-check.md §3): cada referencia declarada que no es lo
// que genera su fila, con su cabecera —la editada— o sin ella, y un .md que la
// skill no declara y que abre con la cabecera, sobrante.
func casosDeDeriva(t *testing.T, skill skillDelRecorrido) []casoSobreUnaCopia {
	t.Helper()

	var casos []casoSobreUnaCopia

	for _, referencia := range skill.referencias {
		casos = append(casos, casoSobreUnaCopia{
			nombre: "referencia-editada-" + referencia.nombre,
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				ruta := rutaEnLaSkill(copia, skill.nombre, referencia.fichero())
				escribirFicheroDeLaCopia(t, ruta, leerFicheroDelArbol(t, ruta)+"Una línea escrita a mano.\n")
			},
			fallos: []string{skill.nombre + ": " + referencia.fichero() + ": contenido-distinto"},
		}, casoSobreUnaCopia{
			nombre: "referencia-sin-cabecera-" + referencia.nombre,
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				ruta := rutaEnLaSkill(copia, skill.nombre, referencia.fichero())
				escribirFicheroDeLaCopia(t, ruta, sinSuCabecera(t, leerFicheroDelArbol(t, ruta)))
			},
			fallos: []string{skill.nombre + ": " + referencia.fichero() + ": contenido-distinto"},
		}, datosSinRegenerar(t, skill, referencia))
	}

	casos = append(casos, casoSobreUnaCopia{
		nombre: "referencia-sobrante",
		alterar: func(t *testing.T, copia string) {
			t.Helper()

			dejarSobrante(t, copia, skill.nombre)
		},
		fallos: []string{skill.nombre + ": references/" + ficheroSobrante + ": fichero-sobrante"},
	})

	for _, applet := range skill.applets {
		casos = append(casos, casoSobreUnaCopia{
			nombre: "describe-cambiado-" + applet,
			cambiarDescripciones: func(t *testing.T, descripciones []skills.DescripcionDeVerbo) {
				t.Helper()

				verbo := slices.IndexFunc(descripciones, func(descripcion skills.DescripcionDeVerbo) bool {
					return descripcion.Applet == applet
				})
				require.GreaterOrEqual(t, verbo, 0, "hay descripción de algún verbo de %s", applet)

				descripciones[verbo].Hace = "Hace otra cosa, que no se ha regenerado."
			},
			fallos: []string{skill.nombre + ": SKILL.md: contenido-distinto"},
		})
	}

	return casos
}

// datosSinRegenerar es el caso de los datos de la referencia cambiados sin
// regenerar: en su YAML de datos, el valor que da la primera celda de su tabla
// gana un añadido. Derivan la referencia y cualquier otra de la skill que salga
// del mismo YAML y muestre ese valor en una celda.
func datosSinRegenerar(t *testing.T, skill skillDelRecorrido, referencia referenciaDelRecorrido) casoSobreUnaCopia {
	t.Helper()

	valor := primeraCelda(t, referencia)

	var fallos []string

	for _, otra := range skill.referenciasDe(referencia.datos) {
		if strings.Contains(otra.contenido, "| "+valor+" |") {
			fallos = append(fallos, skill.nombre+": "+otra.fichero()+": contenido-distinto")
		}
	}

	return casoSobreUnaCopia{
		nombre: "datos-sin-regenerar-" + referencia.nombre,
		alterar: func(t *testing.T, copia string) {
			t.Helper()

			cambiarElDato(t, filepath.Join(copia, filepath.FromSlash(referencia.datos)), valor)
		},
		fallos: fallos,
	}
}

// cambiarElDato añade un texto al valor del escalar del documento YAML de la
// ruta, dentro de una copia temporal, que vale exactamente valor, y que tiene que
// ser uno solo: el valor puede aparecer también dentro de otros, como «Unión
// Europea» en el nombre de su diario oficial.
func cambiarElDato(t *testing.T, ruta, valor string) {
	t.Helper()

	var documento yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(leerFicheroDelArbol(t, ruta)), &documento))

	escalares := escalaresConValor(&documento, valor)
	require.Len(t, escalares, 1, "un solo escalar de %s vale %q", ruta, valor)

	escalares[0].Value += " (cambiado sin regenerar)"

	cambiado, err := yaml.Marshal(&documento)
	require.NoError(t, err)

	escribirFicheroDeLaCopia(t, ruta, string(cambiado))
}

// escalaresConValor son los escalares del árbol del nodo que valen exactamente
// valor, en el orden del documento.
func escalaresConValor(nodo *yaml.Node, valor string) []*yaml.Node {
	if nodo.Kind == yaml.ScalarNode && nodo.Value == valor {
		return []*yaml.Node{nodo}
	}

	var escalares []*yaml.Node

	for _, hijo := range nodo.Content {
		escalares = append(escalares, escalaresConValor(hijo, valor)...)
	}

	return escalares
}

// primeraCelda es el texto de la primera celda de la primera fila de la tabla de
// la referencia: la línea que sigue a su fila de separación.
func primeraCelda(t *testing.T, referencia referenciaDelRecorrido) string {
	t.Helper()

	_, tras, conTabla := strings.Cut(referencia.contenido, "\n|---|")
	require.True(t, conTabla, "%s tiene una tabla", referencia.fichero())

	_, filas, _ := strings.Cut(tras, "\n")
	fila, conFila := strings.CutPrefix(filas, "| ")
	celda, _, cerrada := strings.Cut(fila, " | ")
	require.True(t, conFila && cerrada, "la tabla de %s tiene alguna fila", referencia.fichero())

	return celda
}

// casosDePresupuestos son los presupuestos del cuerpo de SKILL.md de la skill
// (contracts/gates-de-skills-check.md §2.2; FR-004, FR-010, FR-011 y SC-002 de
// H31): 499 líneas y 11 000 caracteres, con más de 11 000 bytes, no son un
// defecto, y 500 líneas y 11 001 caracteres sí. Cada mutante cambia una sola de
// las dos medidas y nada de lo regenerado: las líneas, sin cambiar el número de
// caracteres (partirElCuerpo); los caracteres, con una línea más
// (ensancharElCuerpo), lejos de las 499.
func casosDePresupuestos(t *testing.T, skill skillDelRecorrido) []casoSobreUnaCopia {
	t.Helper()

	return []casoSobreUnaCopia{
		{nombre: "cuatrocientas-noventa-y-nueve-lineas", alterar: partirElCuerpo(skill.nombre, 499)},
		{
			nombre:  "quinientas-lineas",
			alterar: partirElCuerpo(skill.nombre, 500),
			fallos:  []string{skill.nombre + ": SKILL.md: cuerpo de 500 líneas (máximo 499)"},
		},
		{nombre: "once-mil-caracteres", alterar: ensancharElCuerpo(skill.nombre, 11000)},
		{
			nombre:  "once-mil-y-un-caracteres",
			alterar: ensancharElCuerpo(skill.nombre, 11001),
			fallos:  []string{skill.nombre + ": SKILL.md: cuerpo de 11001 caracteres (máximo 11000)"},
		},
	}
}

// partirElCuerpo convierte en saltos de línea, de uno en uno y desde el
// principio, los espacios de las líneas de prosa del cuerpo del SKILL.md de la
// skill hasta que el cuerpo tiene las líneas dadas. Una línea de prosa está
// fuera de la región generada y no lleva corchete, comilla invertida, virgulilla
// ni barra vertical, con los que podría ser parte de un enlace, de un código o de
// un bloque de código, o de una tabla; las marcas de
// la región tampoco lo son. Cada espacio es un carácter y cada salto otro, así
// que el cuerpo conserva sus caracteres, y lo regenerado no cambia.
func partirElCuerpo(skill string, lineas int) func(t *testing.T, copia string) {
	return func(t *testing.T, copia string) {
		t.Helper()

		ruta := rutaEnLaSkill(copia, skill, "SKILL.md")
		frontmatter, cuerpo := partesDelSkillMd(t, leerFicheroDelArbol(t, ruta))
		faltan := lineas - skills.ContarLineas([]byte(cuerpo))

		require.GreaterOrEqual(t, faltan, 0, "el cuerpo de SKILL.md no tiene ya más de %d líneas", lineas)

		var (
			partido   strings.Builder
			enLaTabla bool
		)

		for linea := range strings.Lines(cuerpo) {
			switch strings.TrimSuffix(linea, "\n") {
			case inicioDeLaTablaDeComandos:
				enLaTabla = true
			case finDeLaTablaDeComandos:
				enLaTabla = false
			default:
				for !enLaTabla && !strings.ContainsAny(linea, "[`~|") && faltan > 0 && strings.Contains(linea, " ") {
					linea = strings.Replace(linea, " ", "\n", 1)
					faltan--
				}
			}

			partido.WriteString(linea)
		}

		require.Zero(t, faltan, "el cuerpo de SKILL.md tiene espacios de prosa para llegar a %d líneas", lineas)

		escribirFicheroDeLaCopia(t, ruta, frontmatter+partido.String())
	}
}

// ensancharElCuerpo añade al final del SKILL.md de la skill, que termina en
// salto de línea y fuera de la región generada, una línea de eñes, de dos bytes
// cada una, hasta que su cuerpo tiene los caracteres dados, y exige que entonces
// tenga más bytes que caracteres.
func ensancharElCuerpo(skill string, caracteres int) func(t *testing.T, copia string) {
	return func(t *testing.T, copia string) {
		t.Helper()

		ruta := rutaEnLaSkill(copia, skill, "SKILL.md")
		contenido := leerFicheroDelArbol(t, ruta)
		_, cuerpo := partesDelSkillMd(t, contenido)
		actuales := utf8.RuneCountInString(cuerpo)

		require.True(t, strings.HasSuffix(contenido, "\n"), "SKILL.md termina en salto de línea")
		require.Less(t, actuales, caracteres, "el cuerpo de SKILL.md tiene menos de %d caracteres", caracteres)

		linea := strings.Repeat("ñ", caracteres-actuales-1) + "\n"
		require.Greater(t, len(cuerpo+linea), caracteres, "el cuerpo tiene más bytes que caracteres")

		escribirFicheroDeLaCopia(t, ruta, contenido+linea)
	}
}

// partesDelSkillMd son el frontmatter del SKILL.md, con la línea --- que lo
// cierra, y su cuerpo, lo que sigue a esa línea.
func partesDelSkillMd(t *testing.T, contenido string) (string, string) {
	t.Helper()

	tras, abierto := strings.CutPrefix(contenido, "---")
	bloque, cuerpo, cerrado := strings.Cut(tras, "\n---\n")
	require.True(t, abierto && cerrado, "SKILL.md abre con su frontmatter y lo cierra")

	return "---" + bloque + "\n---\n", cuerpo
}

// casosDeFrontmatter son las reglas del frontmatter de la skill (data-model §1.1
// y §1.3; research.md D3, D4 y D11; FR-040; FR-002, FR-003, FR-005, FR-011 y
// SC-002 de H31): una fila por defecto, cada una con un solo defecto que nombra
// la skill, el fichero —SKILL.md— y la regla —el directorio de la skill se
// renombra cuando el caso necesita otro nombre—, salvo la etiqueta XML en name,
// que es además un carácter que name no admite; uno por cada referencia sin su
// YAML de datos; y los límites exactos válidos, 64 caracteres en name y 900 en
// description, contados como caracteres y no como bytes, y un < que no abre una
// etiqueta. Los nombres que no valen salen del de la skill: todo en mayúsculas,
// con dos guiones tras su primer carácter, con un guion delante y con una
// etiqueta detrás; los que contienen claude y anthropic, de SC-002.
func casosDeFrontmatter(t *testing.T, skill skillDelRecorrido) []casoSobreUnaCopia {
	t.Helper()

	const (
		description              = "description: Una descripción válida de la skill.\n"
		conClaude, conAnthropic  = "claude-legal", "x-anthropic"
		menorQueSinEtiqueta      = "Un cuerpo de menos de 5000 tokens (< 5000)."
		descripcionConLaEtiqueta = "Consulta normas del BOE con un <ejemplo>."
	)

	var (
		nombre                 = skill.nombre
		name                   = "name: " + nombre + "\n"
		metadata               = metadataDeKitlegal(skill.applets, skill.referencias)
		enMayusculas           = strings.ToUpper(nombre)
		conDosGuiones          = nombre[:1] + "--" + nombre[1:]
		conGuionDelante        = "-" + nombre
		conEtiqueta            = nombre + "<x>"
		nombreDe64, nombreDe65 = strings.Repeat("a", 64), strings.Repeat("a", 65)
	)

	casos := []casoSobreUnaCopia{
		{
			nombre:  "sin-name",
			alterar: conOtroFrontmatter(nombre, nombre, description+metadata),
			fallos:  []string{nombre + ": SKILL.md: falta name"},
		},
		{
			nombre:  "name-distinto-del-directorio",
			alterar: conOtroFrontmatter(nombre, nombre, "name: otra-skill\n"+description+metadata),
			fallos:  []string{nombre + `: SKILL.md: name "otra-skill" distinto del nombre del directorio`},
		},
		{
			nombre:  "name-con-mayuscula",
			alterar: conOtroFrontmatter(nombre, enMayusculas, "name: "+enMayusculas+"\n"+description+metadata),
			fallos: []string{
				fmt.Sprintf("%s: SKILL.md: name %q con caracteres que no son a-z, 0-9 ni -", enMayusculas, enMayusculas),
			},
		},
		{
			nombre:  "name-con-dos-guiones-seguidos",
			alterar: conOtroFrontmatter(nombre, conDosGuiones, "name: "+conDosGuiones+"\n"+description+metadata),
			fallos: []string{
				fmt.Sprintf("%s: SKILL.md: name %q con un guion al principio, al final o dos seguidos", conDosGuiones,
					conDosGuiones),
			},
		},
		{
			nombre: "name-que-empieza-por-guion",
			alterar: conOtroFrontmatter(nombre, conGuionDelante,
				fmt.Sprintf("name: %q\n", conGuionDelante)+description+metadata),
			fallos: []string{
				fmt.Sprintf("%s: SKILL.md: name %q con un guion al principio, al final o dos seguidos", conGuionDelante,
					conGuionDelante),
			},
		},
		{
			nombre:  "name-de-65-caracteres",
			alterar: conOtroFrontmatter(nombre, nombreDe65, "name: "+nombreDe65+"\n"+description+metadata),
			fallos:  []string{nombreDe65 + ": SKILL.md: name de 65 caracteres (máximo 64)"},
		},
		{
			nombre:  "name-con-claude",
			alterar: conOtroFrontmatter(nombre, conClaude, "name: "+conClaude+"\n"+description+metadata),
			fallos:  []string{conClaude + `: SKILL.md: name "claude-legal" contiene «claude»`},
		},
		{
			nombre:  "name-con-anthropic",
			alterar: conOtroFrontmatter(nombre, conAnthropic, "name: "+conAnthropic+"\n"+description+metadata),
			fallos:  []string{conAnthropic + `: SKILL.md: name "x-anthropic" contiene «anthropic»`},
		},
		{
			nombre: "name-con-etiqueta-xml",
			alterar: conOtroFrontmatter(nombre, conEtiqueta,
				fmt.Sprintf("name: %q\n", conEtiqueta)+description+metadata),
			fallos: []string{
				fmt.Sprintf("%s: SKILL.md: name %q con caracteres que no son a-z, 0-9 ni -", conEtiqueta, conEtiqueta),
				conEtiqueta + ": SKILL.md: name con la etiqueta XML «<x>»",
			},
		},
		{
			nombre:  "sin-description",
			alterar: conOtroFrontmatter(nombre, nombre, name+metadata),
			fallos:  []string{nombre + ": SKILL.md: falta description"},
		},
		{
			nombre:  "description-vacia",
			alterar: conOtroFrontmatter(nombre, nombre, name+"description: \"\"\n"+metadata),
			fallos:  []string{nombre + ": SKILL.md: description vacía"},
		},
		{
			nombre:  "description-de-901-caracteres",
			alterar: conOtroFrontmatter(nombre, nombre, name+"description: "+strings.Repeat("ñ", 901)+"\n"+metadata),
			fallos:  []string{nombre + ": SKILL.md: description de 901 caracteres (máximo 900)"},
		},
		{
			nombre: "description-con-etiqueta-xml",
			alterar: conOtroFrontmatter(nombre, nombre,
				fmt.Sprintf("%sdescription: %q\n", name, descripcionConLaEtiqueta)+metadata),
			fallos: []string{nombre + ": SKILL.md: description con la etiqueta XML «<ejemplo>»"},
		},
		{
			nombre: "description-con-menor-que-sin-etiqueta",
			alterar: conOtroFrontmatter(nombre, nombre,
				fmt.Sprintf("%sdescription: %q\n", name, menorQueSinEtiqueta)+metadata),
		},
		{
			nombre:  "clave-no-admitida",
			alterar: conOtroFrontmatter(nombre, nombre, name+description+"version: \"1\"\n"+metadata),
			fallos:  []string{nombre + ": SKILL.md: clave no admitida: version"},
		},
		{
			nombre: "applet-que-no-existe",
			alterar: conOtroFrontmatter(nombre, nombre,
				name+description+metadataDeKitlegal(slices.Concat(skill.applets, []string{"inexistente"}), skill.referencias)),
			fallos: []string{nombre + `: SKILL.md: metadata/kitlegal-applets: applet "inexistente" no registrado`},
		},
		{
			nombre: "limites-validos",
			alterar: conOtroFrontmatter(nombre, nombreDe64,
				"name: "+nombreDe64+"\ndescription: "+strings.Repeat("ñ", 900)+"\n"+metadata),
		},
	}

	for _, referencia := range skill.referencias {
		casos = append(casos, referenciaSinSusDatos(skill, referencia))
	}

	return casos
}

// metadataDeKitlegal es el bloque metadata de un frontmatter que declara los
// applets y las referencias dados, en su orden. Sin referencias no lleva la
// clave kitlegal-referencias, como el frontmatter de la skill que no declara
// ninguna (research.md D15 de H23).
func metadataDeKitlegal(applets []string, referencias []referenciaDelRecorrido) string {
	metadata := "metadata:\n  kitlegal-applets: " + strings.Join(applets, " ") + "\n"
	if len(referencias) == 0 {
		return metadata
	}

	nombres := make([]string, 0, len(referencias))
	for _, referencia := range referencias {
		nombres = append(nombres, referencia.nombre)
	}

	return metadata + "  kitlegal-referencias: " + strings.Join(nombres, " ") + "\n"
}

// referenciaSinSusDatos es el caso de la referencia sin el YAML de datos del que
// sale, que se retira: es un defecto de cada referencia de la skill que sale de
// él, en el orden de la declaración.
func referenciaSinSusDatos(skill skillDelRecorrido, referencia referenciaDelRecorrido) casoSobreUnaCopia {
	var fallos []string

	for _, otra := range skill.referenciasDe(referencia.datos) {
		fallos = append(fallos, fmt.Sprintf("%s: SKILL.md: metadata/kitlegal-referencias: %q sin %s", skill.nombre,
			otra.nombre, otra.datos))
	}

	return casoSobreUnaCopia{
		nombre: "referencia-sin-sus-datos-" + referencia.nombre,
		alterar: func(t *testing.T, copia string) {
			t.Helper()

			retirarDeLaCopia(t, filepath.Join(copia, filepath.FromSlash(referencia.datos)))
		},
		fallos: fallos,
	}
}

// conOtroFrontmatter deja la skill en el directorio de nombre directorio,
// renombrándolo si hace falta, con el frontmatter dado en su SKILL.md y el resto
// del fichero intacto.
func conOtroFrontmatter(skill, directorio, frontmatter string) func(t *testing.T, copia string) {
	return func(t *testing.T, copia string) {
		t.Helper()

		if directorio != skill {
			require.NoError(t, os.Rename(rutaEnLaSkill(copia, skill), rutaEnLaSkill(copia, directorio)))
		}

		ruta := rutaEnLaSkill(copia, directorio, "SKILL.md")

		cuerpo, abre := strings.CutPrefix(leerFicheroDelArbol(t, ruta), "---\n")
		require.True(t, abre, "SKILL.md empieza por su frontmatter")

		_, resto, cierra := strings.Cut(cuerpo, "\n---\n")
		require.True(t, cierra, "el frontmatter de SKILL.md se cierra")

		escribirFicheroDeLaCopia(t, ruta, "---\n"+frontmatter+"---\n"+resto)
	}
}

// casosDeReferencias son los gates de references/ en la skill
// (contracts/gates-de-skills-check.md §2.3 y §3; research D13 y D14; FR-006 a
// FR-009, FR-011 y SC-002 de H31): un subdirectorio y un fichero .txt en
// references/; una referencia sin enlace desde SKILL.md; una enlazada desde otra;
// y una de 101 líneas sin índice y otra sin ningún apartado ##, con las mismas de
// 100 en verde y la de 101 con su índice también. Cada referencia de prueba es a
// mano, la única clase que cualquier skill puede ganar sin cambiar data/ ni su
// declaración, y la que el caso no quiere sin enlace va enlazada desde SKILL.md
// sin que el cuerpo gane caracteres (enlazarDesdeElSkillMd), aunque el real esté
// en sus presupuestos. Cada caso en rojo da un solo defecto, que nombra la skill,
// el fichero y la regla.
func casosDeReferencias(t *testing.T, skill skillDelRecorrido) []casoSobreUnaCopia {
	t.Helper()

	const otraAMano = "otra_a_mano_de_prueba.md"

	referencia := "references/" + ficheroAMano
	fallo := func(regla string) []string { return []string{skill.nombre + ": " + referencia + ": " + regla} }

	// enlazada deja en references/ la referencia a mano de prueba con el
	// contenido y la enlaza desde SKILL.md.
	enlazada := func(contenido string) func(t *testing.T, copia string) {
		return func(t *testing.T, copia string) {
			t.Helper()

			escribirReferenciaAMano(t, copia, skill.nombre, contenido)
			enlazarDesdeElSkillMd(t, copia, skill.nombre, ficheroAMano)
		}
	}

	return []casoSobreUnaCopia{
		{
			nombre: "subdirectorio",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				require.NoError(t, os.MkdirAll(rutaEnLaSkill(copia, skill.nombre, "references", "anexos"), 0o750))
				escribirFicheroDeLaCopia(t, rutaEnLaSkill(copia, skill.nombre, "references", "anexos", "nota.md"),
					referenciaAMano)
			},
			fallos: []string{skill.nombre + ": references/anexos: no es un fichero .md"},
		},
		{
			nombre: "fichero-txt",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				require.NoError(t, os.MkdirAll(rutaEnLaSkill(copia, skill.nombre, "references"), 0o750))
				escribirFicheroDeLaCopia(t, rutaEnLaSkill(copia, skill.nombre, "references", "notas.txt"), referenciaAMano)
			},
			fallos: []string{skill.nombre + ": references/notas.txt: no es un fichero .md"},
		},
		{
			nombre: "sin-enlace",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				escribirReferenciaAMano(t, copia, skill.nombre, referenciaAMano)
			},
			fallos: fallo("SKILL.md no la enlaza con [](" + referencia + ")"),
		},
		{
			nombre: "enlazada-desde-otra",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				escribirReferenciaAMano(t, copia, skill.nombre, referenciaAMano+"Ver [la otra]("+otraAMano+").\n")
				escribirFicheroDeLaCopia(t, rutaEnLaSkill(copia, skill.nombre, "references", otraAMano), referenciaAMano)
				enlazarDesdeElSkillMd(t, copia, skill.nombre, ficheroAMano, otraAMano)
			},
			fallos: fallo("enlaza otra referencia, references/" + otraAMano),
		},
		{nombre: "cien-lineas-sin-indice", alterar: enlazada(referenciaLargaAMano(t, 100, true, false))},
		{
			nombre:  "ciento-una-lineas-sin-indice",
			alterar: enlazada(referenciaLargaAMano(t, 101, true, false)),
			fallos:  fallo("101 líneas sin un índice que nombre sus apartados ## antes del primero (más de 100)"),
		},
		{nombre: "ciento-una-lineas-con-indice", alterar: enlazada(referenciaLargaAMano(t, 101, true, true))},
		{nombre: "cien-lineas-sin-apartados", alterar: enlazada(referenciaLargaAMano(t, 100, false, false))},
		{
			nombre:  "ciento-una-lineas-sin-apartados",
			alterar: enlazada(referenciaLargaAMano(t, 101, false, false)),
			fallos:  fallo("101 líneas sin ningún apartado ## (más de 100)"),
		},
	}
}

// referenciaLargaAMano es una referencia a mano de prueba con las líneas dadas:
// su título y su párrafo; con apartados, dos «## » y, con índice, delante de
// ellos la lista que los nombra; y líneas de relleno hasta las dadas.
func referenciaLargaAMano(t *testing.T, lineas int, conApartados, conIndice bool) string {
	t.Helper()

	fijas := []string{"# Referencia de prueba", "", "Una referencia escrita a mano, sin la cabecera de las generadas."}

	if conIndice {
		fijas = append(fijas, "", "- [Lo que lleva](#lo-que-lleva)", "- [Cuándo se lee](#cuando-se-lee)")
	}

	if conApartados {
		fijas = append(fijas, "", "## Lo que lleva", "", "## Cuándo se lee")
	}

	require.Less(t, len(fijas), lineas, "las líneas fijas caben en %d", lineas)

	referencia := strings.Join(fijas, "\n") + "\n" + strings.Repeat("Una línea de relleno.\n", lineas-len(fijas))
	require.Equal(t, lineas, skills.ContarLineas([]byte(referencia)))

	return referencia
}

// vallaDeUnBloque es la línea que abre o cierra un bloque de código en las
// skills: hasta tres espacios y tres acentos graves o virgulillas.
var vallaDeUnBloque = regexp.MustCompile("^ {0,3}(```|~~~)")

// enlazarDesdeElSkillMd enlaza desde el SKILL.md de la skill, dentro de una
// copia temporal, cada fichero de references/ dado con «[](references/<fichero>)»
// (FR-007 de H31): sustituye por el enlace de cada uno, desde el principio del
// cuerpo, una línea de prosa que tenga al menos tantos caracteres como él. Una
// línea de prosa está fuera de la región generada y de los bloques de código y
// no lleva corchete, comilla invertida, virgulilla ni barra vertical
// (partirElCuerpo), así que no enlazaba nada; y como el enlace no es más largo,
// el cuerpo no gana líneas ni caracteres, y sigue en sus presupuestos aunque el
// del SKILL.md real esté en ellos (FR-004 de H31). Lo regenerado no cambia.
func enlazarDesdeElSkillMd(t *testing.T, copia, skill string, ficheros ...string) {
	t.Helper()

	ruta := rutaEnLaSkill(copia, skill, "SKILL.md")
	frontmatter, cuerpo := partesDelSkillMd(t, leerFicheroDelArbol(t, ruta))
	pendientes := slices.Clone(ficheros)

	var (
		enlazado              strings.Builder
		enLaTabla, enUnBloque bool
	)

	for linea := range strings.Lines(cuerpo) {
		texto := strings.TrimSuffix(linea, "\n")

		switch {
		case texto == inicioDeLaTablaDeComandos:
			enLaTabla = true
		case texto == finDeLaTablaDeComandos:
			enLaTabla = false
		case vallaDeUnBloque.MatchString(texto):
			enUnBloque = !enUnBloque
		case !enLaTabla && !enUnBloque && len(pendientes) > 0 && !strings.ContainsAny(texto, "[`~|"):
			enlace := "[](references/" + pendientes[0] + ")"
			if utf8.RuneCountInString(texto) >= utf8.RuneCountInString(enlace) {
				linea = enlace + strings.TrimPrefix(linea, texto)
				pendientes = pendientes[1:]
			}
		}

		enlazado.WriteString(linea)
	}

	require.Empty(t, pendientes, "el cuerpo de SKILL.md tiene líneas de prosa para enlazar %v", ficheros)

	escribirFicheroDeLaCopia(t, ruta, frontmatter+enlazado.String())
}

// casosDeScripts son el defecto de FR-082 en la skill (ADR 0019;
// contracts/skills-e-invocacion.md §3 de H19): una entrada scripts en su
// directorio, sea lo que sea —el directorio con un enlace por applet declarado
// que dejaba la instalación anterior, uno vacío, un fichero o un enlace
// colgante, que solo se ve sin seguirlo—, hace fallar la comprobación nombrando
// la skill y el defecto, y nada más.
func casosDeScripts(t *testing.T, skill skillDelRecorrido) []casoSobreUnaCopia {
	t.Helper()

	fallos := []string{skill.nombre + ": una skill no lleva scripts/ (ADR 0019)"}

	return []casoSobreUnaCopia{
		{
			nombre: "con-los-enlaces-de-la-instalacion-anterior",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				require.NoError(t, os.Mkdir(rutaEnLaSkill(copia, skill.nombre, "scripts"), 0o750))

				for _, applet := range skill.applets {
					enlazarEnLaCopia(t, "../../../bin/instalado/kitlegal",
						rutaEnLaSkill(copia, skill.nombre, "scripts", applet))
				}
			},
			fallos: fallos,
		},
		{
			nombre: "vacio",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				require.NoError(t, os.Mkdir(rutaEnLaSkill(copia, skill.nombre, "scripts"), 0o750))
			},
			fallos: fallos,
		},
		{
			nombre: "fichero",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				escribirFicheroDeLaCopia(t, rutaEnLaSkill(copia, skill.nombre, "scripts"), "#!/bin/sh\n")
			},
			fallos: fallos,
		},
		{
			nombre: "enlace-colgante",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				enlazarEnLaCopia(t, "no-existe", rutaEnLaSkill(copia, skill.nombre, "scripts"))
			},
			fallos: fallos,
		},
	}
}

// casosDeRegion son los defectos de las marcas de la región generada del
// SKILL.md de la skill (data-model §2; FR-032), con las líneas que ocupan las dos
// marcas en su SKILL.md real.
func casosDeRegion(t *testing.T, skill skillDelRecorrido) []casoSobreUnaCopia {
	t.Helper()

	skillMd := leerFicheroDelArbol(t, rutaEnLaSkill(raizDelRepositorio, skill.nombre, "SKILL.md"))
	inicio, fin := lineaDeLaMarca(t, skillMd, inicioDeLaTablaDeComandos), lineaDeLaMarca(t, skillMd, finDeLaTablaDeComandos)

	return []casoSobreUnaCopia{
		{
			nombre: "sin-marcas",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				ruta := rutaEnLaSkill(copia, skill.nombre, "SKILL.md")
				cambiarFicheroDeLaCopia(t, ruta, inicioDeLaTablaDeComandos+"\n", "")
				cambiarFicheroDeLaCopia(t, ruta, finDeLaTablaDeComandos+"\n", "")
			},
			fallos: []string{skill.nombre + ": SKILL.md: sin las marcas de la tabla de comandos"},
		},
		{
			// La marca de inicio repetida sustituye a la tabla, entre las dos marcas:
			// el cuerpo de la copia tiene menos líneas y menos caracteres que el del
			// SKILL.md real, que puede estar en sus máximos (FR-004 de H31), y su
			// único defecto es el de la marca.
			nombre: "dos-inicios",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				ruta := rutaEnLaSkill(copia, skill.nombre, "SKILL.md")
				contenido := leerFicheroDelArbol(t, ruta)

				antes, tras, hayInicio := strings.Cut(contenido, inicioDeLaTablaDeComandos+"\n")
				_, despues, hayFin := strings.Cut(tras, finDeLaTablaDeComandos+"\n")
				require.True(t, hayInicio && hayFin, "SKILL.md tiene las dos marcas, en su orden")

				escribirFicheroDeLaCopia(t, ruta,
					antes+inicioDeLaTablaDeComandos+"\n"+inicioDeLaTablaDeComandos+"\n"+finDeLaTablaDeComandos+"\n"+despues)
			},
			fallos: []string{fmt.Sprintf("%s: SKILL.md: la marca de inicio de la tabla de comandos "+
				"aparece 2 veces, en las líneas %d y %d", skill.nombre, inicio, inicio+1)},
		},
		{
			nombre: "fin-antes-del-inicio",
			alterar: func(t *testing.T, copia string) {
				t.Helper()

				ruta := rutaEnLaSkill(copia, skill.nombre, "SKILL.md")
				contenido := leerFicheroDelArbol(t, ruta)

				antes, tras, hayInicio := strings.Cut(contenido, inicioDeLaTablaDeComandos+"\n")
				tabla, despues, hayFin := strings.Cut(tras, finDeLaTablaDeComandos+"\n")
				require.True(t, hayInicio && hayFin, "SKILL.md tiene las dos marcas, en su orden")

				escribirFicheroDeLaCopia(t, ruta, antes+finDeLaTablaDeComandos+"\n"+tabla+inicioDeLaTablaDeComandos+"\n"+despues)
			},
			fallos: []string{fmt.Sprintf("%s: SKILL.md: la marca de fin de la tabla de comandos, en la "+
				"línea %d, está antes que la de inicio, en la línea %d", skill.nombre, inicio, fin)},
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

// probarRegenerarDosVeces fija que la regeneración de la skill es idempotente
// (FR-034, FR-036): sobre una copia con una deriva de cada parte generada —la
// tabla, cada referencia y una referencia sobrante—, Escribir la deja sin
// derivas y sin ningún scripts/ (ADR 0019; FR-080 de H19), y un segundo Escribir
// no cambia ningún byte ni ningún tiempo de modificación.
func probarRegenerarDosVeces(t *testing.T, descripciones []skills.DescripcionDeVerbo, skill skillDelRecorrido) {
	t.Helper()

	copia := copiaConLaSkill(t, skill.nombre)

	cambiarFicheroDeLaCopia(t, rutaEnLaSkill(copia, skill.nombre, "SKILL.md"), finDeLaTablaDeComandos+"\n",
		"tabla escrita a mano\n"+finDeLaTablaDeComandos+"\n")

	fallos := []string{skill.nombre + ": SKILL.md: contenido-distinto"}

	for _, referencia := range skill.referencias {
		retirarDeLaCopia(t, rutaEnLaSkill(copia, skill.nombre, referencia.fichero()))
		fallos = append(fallos, skill.nombre+": "+referencia.fichero()+": fichero-ausente")
	}

	dejarSobrante(t, copia, skill.nombre)
	fallos = append(fallos, skill.nombre+": references/"+ficheroSobrante+": fichero-sobrante")

	require.Equal(t, fallos, fallosDeLasSkills(t, copia, descripciones))

	escribirSkills(t, copia, descripciones)
	require.Empty(t, fallosDeLasSkills(t, copia, descripciones), "Escribir deja la copia sin defectos ni derivas")

	_, err := os.Lstat(rutaEnLaSkill(copia, skill.nombre, "scripts"))
	require.ErrorIs(t, err, fs.ErrNotExist, "Escribir no crea scripts/")

	envejecerArbol(t, copia)
	antes := fotografiarArbol(t, copia)

	escribirSkills(t, copia, descripciones)
	assert.Equal(t, antes, fotografiarArbol(t, copia),
		"el segundo Escribir no cambia ningún byte ni ningún tiempo de modificación")
	assert.Empty(t, fallosDeLasSkills(t, copia, descripciones))
}

// Lo que las pruebas sobre copias dejan en references/ además de las referencias
// declaradas (FR-012): un fichero sobrante, que abre con la cabecera de una
// referencia generada sin que la skill lo declare, y una referencia a mano, que no
// abre con ella. Ninguno de los dos nombres es el de una fila de la tabla de
// generadores.
const (
	ficheroSobrante     = "antigua.md"
	sobranteConCabecera = "<!-- generado desde data/normas.yaml, no editar -->\n\n# Antigua\n"

	ficheroAMano    = "a_mano_de_prueba.md"
	referenciaAMano = "# Protocolo de prueba\n\nUna referencia escrita a mano, sin la cabecera de las generadas.\n"
)

// dejarSobrante deja en references/ de la skill, dentro de una copia temporal, el
// fichero sobrante. La skill que no declara referencias no tiene references/
// (research.md D15 de H23): la carpeta se crea antes.
func dejarSobrante(t *testing.T, copia, skill string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(rutaEnLaSkill(copia, skill, "references"), 0o750))
	escribirFicheroDeLaCopia(t, rutaEnLaSkill(copia, skill, "references", ficheroSobrante), sobranteConCabecera)
}

// escribirReferenciaAMano deja en references/ de la skill, dentro de una copia
// temporal, la referencia a mano de prueba con el contenido, y devuelve su ruta
// dentro del directorio de la skill, como la nombran los defectos:
// references/<fichero>. Exige que skills.ReferenciasAMano la clasifique como a
// mano (contracts/gates-de-skills-check.md §3 de H31): un mutante sobre otra
// clase de referencia no probaría que el control lee las referencias a mano.
func escribirReferenciaAMano(t *testing.T, copia, skill, contenido string) string {
	t.Helper()

	require.NoError(t, os.MkdirAll(rutaEnLaSkill(copia, skill, "references"), 0o750))
	escribirFicheroDeLaCopia(t, rutaEnLaSkill(copia, skill, "references", ficheroAMano), contenido)

	aMano, err := skills.ReferenciasAMano(os.DirFS(copia), skill)
	require.NoError(t, err)
	require.True(t, slices.ContainsFunc(aMano, func(referencia skills.Referencia) bool {
		return referencia.Fichero == ficheroAMano
	}), "%s: %s es una referencia a mano", skill, ficheroAMano)

	return path.Join("references", ficheroAMano)
}

// sinSuCabecera es el contenido de una referencia generada sin su primera línea,
// la cabecera «generado desde».
func sinSuCabecera(t *testing.T, contenido string) string {
	t.Helper()

	cabecera := cabeceraDeUnaReferencia.FindString(contenido)
	require.NotEmpty(t, cabecera, "la referencia abre con su cabecera")

	return strings.TrimPrefix(contenido, cabecera)
}

// conUnByteCambiado es el contenido de una referencia generada con un solo byte
// cambiado: la última barra de su tabla, la que cierra su última fila, es una
// exclamación.
func conUnByteCambiado(t *testing.T, contenido string) string {
	t.Helper()

	barra := strings.LastIndexByte(contenido, '|')
	require.GreaterOrEqual(t, barra, 0, "la referencia tiene una tabla")

	return contenido[:barra] + "!" + contenido[barra+1:]
}

// probarSincroniaConReferenciasAMano fija SC-003 (FR-012; contracts/
// gates-de-skills-check.md §3) sobre dos copias con solo la skill. Cada una lleva
// en references/ una referencia a mano, enlazada desde su SKILL.md (FR-007 de
// H31), y cada referencia declarada alterada, una
// con un byte cambiado y la siguiente sin su cabecera, alternando: la primera
// copia empieza por el byte y la segunda por la cabecera, de modo que cada
// referencia declarada pasa por las dos alteraciones, también en la skill que
// declara una sola. skills-check da la deriva de cada referencia declarada y
// ninguna de la referencia a mano; y skills-sync —Regenerar y Escribir— deja la
// referencia a mano idéntica, sin tocar sus bytes ni su tiempo de modificación, y
// rehace las declaradas.
func probarSincroniaConReferenciasAMano(t *testing.T, descripciones []skills.DescripcionDeVerbo,
	skill skillDelRecorrido,
) {
	t.Helper()

	alteraciones := []func(t *testing.T, contenido string) string{conUnByteCambiado, sinSuCabecera}
	aMano := path.Join("skills", skill.nombre, "references", ficheroAMano)

	for primera, nombre := range []string{"desde-el-byte-cambiado", "desde-la-cabecera-retirada"} {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			copia := copiaConLaSkill(t, skill.nombre)
			require.NoError(t, os.MkdirAll(rutaEnLaSkill(copia, skill.nombre, "references"), 0o750))
			escribirFicheroDeLaCopia(t, filepath.Join(copia, filepath.FromSlash(aMano)), referenciaAMano)
			enlazarDesdeElSkillMd(t, copia, skill.nombre, ficheroAMano)

			var derivas []string

			for indice, referencia := range skill.referencias {
				alterar := alteraciones[(primera+indice)%len(alteraciones)]
				escribirFicheroDeLaCopia(t, rutaEnLaSkill(copia, skill.nombre, referencia.fichero()),
					alterar(t, referencia.contenido))
				derivas = append(derivas, skill.nombre+": "+referencia.fichero()+": contenido-distinto")
			}

			require.Equal(t, derivas, fallosDeLasSkills(t, copia, descripciones),
				"skills-check da la deriva de cada referencia declarada y ninguna de la referencia a mano")

			envejecerArbol(t, copia)
			antes := fotografiarArbol(t, copia)

			escribirSkills(t, copia, descripciones)

			despues := fotografiarArbol(t, copia)
			assert.Equal(t, antes[aMano], despues[aMano], "skills-sync no toca la referencia a mano")
			assert.Equal(t, referenciaAMano, despues[aMano].contenido)

			for _, referencia := range skill.referencias {
				assert.Equal(t, referencia.contenido,
					leerFicheroDelArbol(t, rutaEnLaSkill(copia, skill.nombre, referencia.fichero())),
					"skills-sync rehace %s", referencia.fichero())
			}

			assert.Empty(t, fallosDeLasSkills(t, copia, descripciones))
		})
	}
}

// normaNombrada es la forma en que una skill nombra una norma por su rango, su
// número y su año: «Ley N/AAAA», «Ley Orgánica N/AAAA», «Real Decreto N/AAAA»,
// «Real Decreto-ley N/AAAA» o «Real Decreto Legislativo N/AAAA» (contrato
// sincronizacion-y-comprobacion §2). Los rangos más largos van antes que los que
// empiezan igual.
var normaNombrada = regexp.MustCompile(
	`\b(?:Ley Orgánica|Ley|Real Decreto Legislativo|Real Decreto-ley|Real Decreto) [0-9]+/[0-9]{4}\b`)

// probarNormasNombradasDelArbol fija que toda norma que nombra una skill entera
// del árbol real empieza el título de una norma de data/normas.yaml (FR-020; FR-016
// de H31), y que el árbol nombra alguna: sin ninguna, la comprobación pasaría en
// vacío.
func probarNormasNombradasDelArbol(t *testing.T) {
	t.Helper()

	var nombradas []string

	for _, skill := range skillsDelArbol(t, raizDelRepositorio) {
		for _, fichero := range ficherosDeLaSkillEntera(t, raizDelRepositorio, skill) {
			nombradas = append(nombradas, normaNombrada.FindAllString(fichero.texto, -1)...)
		}
	}

	require.NotEmpty(t, nombradas, "alguna skill nombra alguna norma por su número y año")
	assert.Empty(t, fallosDeNormasNombradas(t, raizDelRepositorio))
}

// textoDeUnFichero es un fichero de una skill que leen los controles de
// contenido: su ruta dentro del directorio de la skill, con barra, como la
// nombran los defectos, y su texto.
type textoDeUnFichero struct {
	ruta  string
	texto string
}

// ficherosDeLaSkillEntera son los ficheros de la skill del árbol de raíz que lee
// un control de contenido desde H31 (contracts/controles-de-contenido.md de H31;
// research D16 de H31; FR-016): su SKILL.md y, detrás, sus referencias a mano
// (skills.ReferenciasAMano), en orden de nombre, cada uno por separado. Las
// referencias generadas no se leen.
func ficherosDeLaSkillEntera(t *testing.T, raiz, skill string) []textoDeUnFichero {
	t.Helper()

	aMano, err := skills.ReferenciasAMano(os.DirFS(raiz), skill)
	require.NoError(t, err, "las referencias a mano de %s", skill)

	ficheros := []textoDeUnFichero{{
		ruta:  ficheroDeLaSkill,
		texto: leerFicheroDelArbol(t, rutaEnLaSkill(raiz, skill, ficheroDeLaSkill)),
	}}

	for _, referencia := range aMano {
		ficheros = append(ficheros, textoDeUnFichero{
			ruta:  path.Join("references", referencia.Fichero),
			texto: string(referencia.Contenido),
		})
	}

	return ficheros
}

// probarNormasNombradas fija, sobre una copia con solo la skill, que su SKILL.md
// falla por cada una de cinco normas nombradas que la tabla no tiene, una de cada
// forma, y no por una que sí tiene (FR-020); y que una referencia a mano falla
// igual por la que la tabla no tiene, con su fichero en el defecto, y no por la
// que tiene (FR-016, FR-017; contracts/controles-de-contenido.md, control 8, de
// H31). Las que no tiene llevan el número 0, que no lleva ninguna norma: la tabla
// crece con cada hito —en H6, con la LEC 1/2000 y la LOPDGDD 3/2018, que eran dos
// de estos ejemplos— y un ejemplo que pudiera entrar en ella dejaría el caso sin
// fallo.
func probarNormasNombradas(t *testing.T, skill skillDelRecorrido) {
	t.Helper()

	copia := copiaConLaSkill(t, skill.nombre)
	ruta := rutaEnLaSkill(copia, skill.nombre, ficheroDeLaSkill)
	escribirFicheroDeLaCopia(t, ruta, leerFicheroDelArbol(t, ruta)+"Ver la Ley 0/2000, la Ley Orgánica 0/2018, "+
		"el Real Decreto 0/2001, el Real Decreto-ley 0/2020, el Real Decreto Legislativo 0/2015 y el "+
		"Real Decreto Legislativo 2/2004.\n")

	aMano := escribirReferenciaAMano(t, copia, skill.nombre,
		referenciaAMano+"Ver la Ley 0/1999 y el Real Decreto Legislativo 2/2004.\n")

	const sinEntrada = "%s: %s nombra «%s», que no empieza el título de ninguna norma de data/normas.yaml"

	assert.Equal(t, []string{
		fmt.Sprintf(sinEntrada, skill.nombre, ficheroDeLaSkill, "Ley 0/2000"),
		fmt.Sprintf(sinEntrada, skill.nombre, ficheroDeLaSkill, "Ley Orgánica 0/2018"),
		fmt.Sprintf(sinEntrada, skill.nombre, ficheroDeLaSkill, "Real Decreto 0/2001"),
		fmt.Sprintf(sinEntrada, skill.nombre, ficheroDeLaSkill, "Real Decreto-ley 0/2020"),
		fmt.Sprintf(sinEntrada, skill.nombre, ficheroDeLaSkill, "Real Decreto Legislativo 0/2015"),
		fmt.Sprintf(sinEntrada, skill.nombre, aMano, "Ley 0/1999"),
	}, fallosDeNormasNombradas(t, copia))
}

// fallosDeNormasNombradas es un fallo por cada norma que nombra un fichero de
// una skill entera del árbol de raíz (ficherosDeLaSkillEntera) sin que ninguna
// norma de su data/normas.yaml tenga un título que empiece así, con el fichero,
// en el orden de las skills, de sus ficheros y del texto.
func fallosDeNormasNombradas(t *testing.T, raiz string) []string {
	t.Helper()

	normas, err := skills.LeerNormas([]byte(leerFicheroDelArbol(t, filepath.Join(raiz, "data", "normas.yaml"))))
	require.NoError(t, err)

	var fallos []string

	for _, skill := range skillsDelArbol(t, raiz) {
		for _, fichero := range ficherosDeLaSkillEntera(t, raiz, skill) {
			for _, nombrada := range normasSinTitulo(normas, fichero.texto) {
				fallos = append(fallos, fmt.Sprintf("%s: %s nombra «%s», que no empieza el título de ninguna "+
					"norma de data/normas.yaml", skill, fichero.ruta, nombrada))
			}
		}
	}

	return fallos
}

// normasSinTitulo son las normas que nombra el texto sin que ninguna de las
// normas tenga un título que empiece así, en el orden del texto.
func normasSinTitulo(normas []skills.Norma, texto string) []string {
	var sinTitulo []string

	for _, nombrada := range normaNombrada.FindAllString(texto, -1) {
		empieza := slices.ContainsFunc(normas, func(norma skills.Norma) bool {
			return empiezaElTitulo(norma.Titulo, nombrada)
		})
		if !empieza {
			sinTitulo = append(sinTitulo, nombrada)
		}
	}

	return sinTitulo
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

// probarSinInstruccionesDeEvals fija, sobre una copia con solo la skill, que
// cada término de lo que no puede decir (FR-077) cuenta en su SKILL.md y en cada
// una de sus referencias, también en una referencia a mano (FR-016, FR-017 de
// H31), y que no cuentan «evalúa», el «siempre» de otra frase ni lo que va dentro
// de la región generada.
func probarSinInstruccionesDeEvals(t *testing.T, skill skillDelRecorrido) {
	t.Helper()

	copia := copiaConLaSkill(t, skill.nombre)

	skillMd := rutaEnLaSkill(copia, skill.nombre, "SKILL.md")
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

	esperados := make([]string, 0, len(prohibidas)+len(skill.referencias)+1)
	for indice, linea := range prohibidas {
		esperados = append(esperados, fmt.Sprintf("%s: SKILL.md:%d: %s", skill.nombre, primera+indice, linea))
	}

	// Cada referencia declarada lleva una fila que lo dice, y la referencia a
	// mano, una línea (FR-016, FR-017; contracts/controles-de-contenido.md,
	// control 9, de H31).
	porFichero := make(map[string]string, len(skill.referencias)+1)

	for _, referencia := range skill.referencias {
		ruta := rutaEnLaSkill(copia, skill.nombre, referencia.fichero())
		tabla := leerFicheroDelArbol(t, ruta)
		escribirFicheroDeLaCopia(t, ruta, tabla+"| evals |\n")

		porFichero[referencia.fichero()] = fmt.Sprintf("%s: %s:%d: | evals |", skill.nombre, referencia.fichero(),
			skills.ContarLineas([]byte(tabla))+1)
	}

	const enLaReferenciaAMano = "Lo lanza un job semanal."

	aMano := escribirReferenciaAMano(t, copia, skill.nombre, referenciaAMano+enLaReferenciaAMano+"\n")
	porFichero[aMano] = fmt.Sprintf("%s: %s:%d: %s", skill.nombre, aMano,
		skills.ContarLineas([]byte(referenciaAMano))+1, enLaReferenciaAMano)

	// Las referencias se recorren en el orden de sus ficheros, que es el de
	// references/.
	for _, fichero := range slices.Sorted(maps.Keys(porFichero)) {
		esperados = append(esperados, porFichero[fichero])
	}

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

	registro, err := RegistroDeProduccion("")
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
// genera para los applets, en su orden.
func sintaxisDeLaTabla(t *testing.T, applets []string, descripciones []skills.DescripcionDeVerbo) []string {
	t.Helper()

	return ordenesDe(filasDeLaTablaGenerada(t, applets, descripciones))
}

// filasDeLaTablaGenerada son las filas de la tabla de comandos que se genera
// para los applets, en su orden.
func filasDeLaTablaGenerada(t *testing.T, applets []string, descripciones []skills.DescripcionDeVerbo,
) []filaDeLaTabla {
	t.Helper()

	tabla, err := skills.RenderizarTabla(applets, descripciones)
	require.NoError(t, err)

	return filasDeLaTabla(t, string(tabla))
}

// filaDeLaTabla es lo que una fila de una tabla de comandos nombra de dos
// formas: la orden y la herramienta (contracts/skills.md §4 de H21).
type filaDeLaTabla struct {
	// orden es el texto de la primera celda, sin las comillas de código.
	orden string

	// herramienta es el texto de la segunda celda, sin las comillas de código, o
	// vacío si esa celda no es solo un tramo de código.
	herramienta string
}

// ordenesDe son las órdenes de las filas, en su orden.
func ordenesDe(filas []filaDeLaTabla) []string {
	ordenes := make([]string, 0, len(filas))
	for _, fila := range filas {
		ordenes = append(ordenes, fila.orden)
	}

	return ordenes
}

// filasDeLaTabla son las filas de una tabla de comandos, en su orden: de cada
// línea que empieza por una celda de código, el texto de esa celda, que es la
// orden, y el de la celda siguiente si es también solo un tramo de código, que es
// la herramienta.
func filasDeLaTabla(t *testing.T, tabla string) []filaDeLaTabla {
	t.Helper()

	var filas []filaDeLaTabla

	for linea := range strings.Lines(tabla) {
		celda, esFila := strings.CutPrefix(linea, "| `")
		if !esFila {
			continue
		}

		orden, resto, cerrada := strings.Cut(celda, "` |")
		require.True(t, cerrada, "la primera celda de %q se cierra", linea)

		segunda, _, _ := strings.Cut(resto, " |")

		filas = append(filas, filaDeLaTabla{orden: orden, herramienta: tramoDeCodigo(strings.TrimSpace(segunda))})
	}

	return filas
}

// tramoDeCodigo es el texto de una celda que es solo un tramo de código, sin sus
// comillas; vacío si la celda lleva algo más, o nada.
func tramoDeCodigo(celda string) string {
	codigo, abierto := strings.CutPrefix(celda, "`")
	codigo, cerrado := strings.CutSuffix(codigo, "`")

	if !abierto || !cerrado || strings.Contains(codigo, "`") {
		return ""
	}

	return codigo
}

// invocacionDeLaSintaxis es la invocación mínima que cumple una sintaxis de la
// tabla, seguida de --describe (contrato sincronizacion-y-comprobacion §3;
// contracts/skills-e-invocacion.md §2 de H19): kitlegal como nombre del programa,
// el applet y el verbo, que es como lo invoca la skill desde el PATH, x por cada
// argumento obligatorio, x y por cada uno de varios valores y nada de lo
// opcional, que es lo que va entre corchetes: los argumentos de posición
// opcionales y las banderas propias, con su valor.
func invocacionDeLaSintaxis(t *testing.T, orden string) []string {
	t.Helper()

	return invocacionDe(t, orden, false)
}

// invocacionCompletaDeLaSintaxis es la invocación que escribe todo lo que admite
// una sintaxis de la tabla, seguida de --describe: la mínima y, además, x por
// cada argumento de posición opcional, x y si es de varios valores, y cada
// bandera propia como la escribe la fila, con x detrás la que lleva valor
// (research.md D14 de H23).
func invocacionCompletaDeLaSintaxis(t *testing.T, orden string) []string {
	t.Helper()

	return invocacionDe(t, orden, true)
}

// invocacionDe convierte una sintaxis de la tabla en una invocación, seguida de
// --describe, con lo que va entre corchetes o sin ello. Es opcional la parte que
// abre un corchete y toda la que va dentro de uno abierto, que es como queda el
// valor de una bandera: `[--dir <dir>]` son dos partes. De cada parte, sin sus
// corchetes, una bandera se escribe tal cual, un argumento de varios valores da x
// y, y cualquier otro —también el valor de una bandera— da x.
func invocacionDe(t *testing.T, orden string, conLoOpcional bool) []string {
	t.Helper()

	partes := strings.Fields(orden)
	require.GreaterOrEqual(t, len(partes), 3, "la sintaxis %q nombra el programa, el applet y el verbo", orden)
	require.Equal(t, programaDeLasOrdenes, partes[0], "la sintaxis %q invoca kitlegal", orden)

	argv := []string{partes[0], partes[1], partes[2]}
	abiertos := 0

	for _, parte := range partes[3:] {
		opcional := abiertos > 0 || strings.HasPrefix(parte, "[")
		abiertos += strings.Count(parte, "[") - strings.Count(parte, "]")

		if opcional && !conLoOpcional {
			continue
		}

		switch escrita := strings.Trim(parte, "[]"); {
		case strings.HasPrefix(escrita, "--"):
			argv = append(argv, escrita)
		case strings.HasSuffix(escrita, ">..."):
			argv = append(argv, "x", "y")
		default:
			argv = append(argv, "x")
		}
	}

	require.Zero(t, abiertos, "la sintaxis %q cierra cada corchete que abre", orden)

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

// copiaConLaSkill copia en un directorio temporal lo que Regenerar lee del árbol
// real para la skill —su directorio de skills/ y data/— y devuelve su raíz: sin
// las demás skills, un cambio en data/ o en lo que declara --describe solo da los
// fallos de ella. Los enlaces simbólicos se recrean con su destino literal, sin
// seguirlos, de modo que la copia tiene lo mismo que el árbol aunque no resuelvan.
// Los dos árboles se abren como os.Root, que no deja salir de ellos por un enlace
// mientras se copia.
func copiaConLaSkill(t *testing.T, skill string) string {
	t.Helper()

	copia := t.TempDir()
	origen, destino := abrirArbol(t, raizDelRepositorio), abrirArbol(t, copia)

	for _, carpeta := range []string{path.Join("skills", skill), "data"} {
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
