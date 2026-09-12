// Package internal_test alberga el test de arquitectura del módulo: la segunda
// capa —independiente de la configuración del lint— que hace cumplir las tres
// reglas de importación de contracts/reglas-de-arquitectura.md (FR-051).
//
// El directorio internal/ no declara ningún paquete de producción: contiene
// solo este fichero, que es donde docs/ROADMAP.md §3 sitúa el test, y sobre la
// orden que §3 nombra, `go list -deps`.
//
// Por qué existe habiendo `depguard`: una configuración de lint se desactiva
// borrando tres líneas de YAML o con un `//nolint`; un test que falla en
// `make ci`, no. Y no comprueban lo mismo: `depguard` mira las importaciones
// declaradas en cada fichero, mientras que este test parte del cierre
// transitivo real que devuelve `go list -deps`, donde aparecen también los
// paquetes del módulo a los que no llega ningún comodín. El reparto entre las
// dos capas es limpio: `go list -deps` describe el grafo del código de
// producción, y los ficheros de test —que también están sujetos a las tres
// reglas— los cubre `depguard`, porque el lint corre con `run.tests: true`.
//
// Por qué solo las tres reglas de importación: R4 («solo las raíces de
// composición terminan el proceso») y R5 («solo internal/render escribe en la
// salida estándar») son reglas de **símbolo**. Todo paquete importa `os`; lo
// prohibido es llamar a `os.Exit` o referenciar `os.Stdout`, y eso un grafo de
// importación no lo ve. Las vigila `forbidigo` con análisis de tipos, y fingir
// aquí que están cubiertas daría una garantía falsa (FR-051, SC-008).
package internal_test

import (
	"errors"
	"maps"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// raizDelModulo es el directorio del módulo visto desde este paquete: el
// fichero vive en internal/, así que el módulo está justo encima. `go list` se
// ejecuta ahí para que las rutas de paquete sean las mismas que usan el
// Makefile y la integración continua.
const raizDelModulo = ".."

// plantillaDeListado pide a `go list` una línea por paquete del cierre con su
// ruta de importación y sus importaciones directas. Con -deps la lista incluye
// el paquete nombrado y todo aquello de lo que depende, que es el grafo
// transitivo del que parte este test.
const plantillaDeListado = "{{.ImportPath}}\t{{join .Imports \" \"}}"

// TestArquitectura comprueba las tres reglas de importación sobre el grafo
// transitivo real del módulo. Es el test que el escenario 8 de quickstart.md
// invoca por su nombre.
func TestArquitectura(t *testing.T) {
	t.Parallel()

	grafo := grafoDelModulo(t)

	t.Run("R1 · el dominio no importa el kernel, los adaptadores ni entrada y salida", func(t *testing.T) {
		t.Parallel()

		compruebaDominioPuro(t, grafo)
	})

	t.Run("R2 · solo internal/httpx importa net/http", func(t *testing.T) {
		t.Parallel()

		compruebaImportacionExclusiva(t, grafo, reglaExclusiva{
			nombre: "R2",
			razon: "la biblioteca HTTP se usa a través de internal/httpx, que es lo que concentra " +
				"los reintentos, el límite de peticiones por sitio, robots.txt y el User-Agent " +
				"identificable (contracts/reglas-de-arquitectura.md R2)",
			duenos:    []string{grafo.modulo + "/internal/httpx"},
			denegados: []string{"net/http"},
		})
	})

	t.Run("R3 · solo internal/{cache,store,graph} importan SQLite y database/sql", func(t *testing.T) {
		t.Parallel()

		compruebaImportacionExclusiva(t, grafo, reglaExclusiva{
			nombre: "R3",
			razon: "el acceso a SQLite vive en los tres paquetes de almacenamiento —caché, almacén " +
				"y grafo—; el resto del árbol los usa a través de su interfaz " +
				"(contracts/reglas-de-arquitectura.md R3)",
			duenos: []string{
				grafo.modulo + "/internal/cache",
				grafo.modulo + "/internal/store",
				grafo.modulo + "/internal/graph",
			},
			denegados: []string{"database/sql", "modernc.org/sqlite"},
		})
	})
}

// modulosDelBinario son los módulos de terceros que el binario distribuido
// enlaza, y ninguno más: los dos de la lista cerrada de la constitución §V que
// van en producción y los cuatro que el segundo arrastra consigo. Cada uno de
// los cuatro está justificado en plan.md (Complexity Tracking) y en
// gates/pr-h1.md: entran por `invopop/jsonschema`, que sí está en la lista, y
// no hay versión de esa biblioteca que no los traiga (FR-060).
//
// Es una lista escrita a mano a propósito. Cuando un hito, o una actualización
// de módulos, enlace uno nuevo, este test falla y obliga a hacer lo que la
// constitución exige: justificarlo por escrito antes de añadirlo aquí.
var modulosDelBinario = []string{
	"github.com/alecthomas/kong",
	"github.com/bahlo/generic-list-go",
	"github.com/buger/jsonparser",
	"github.com/invopop/jsonschema",
	"github.com/pb33f/ordered-map/v2",
	"go.yaml.in/yaml/v4",
}

// plantillaDeModulos pide a `go list` el módulo de cada paquete del cierre; los
// de la biblioteca estándar no tienen módulo y salen como línea vacía.
const plantillaDeModulos = "{{if .Module}}{{.Module.Path}}{{end}}"

// TestElBinarioNoEnlazaLosEjemplos comprueba que los applets de ejemplo no
// llegan al binario distribuido: son la implementación de referencia, no
// funcionalidad, y viven en un paquete normal del módulo para que todas las
// comprobaciones los alcancen. Lo que impide enlazarlos es esta comprobación sobre
// el cierre transitivo real de cmd/kitlegal, junto con la lista `ejemplo` de
// depguard (docs/ADR/0010-applets-de-ejemplo-fuera-de-testdata.md).
func TestElBinarioNoEnlazaLosEjemplos(t *testing.T) {
	t.Parallel()

	modulo := strings.TrimSpace(ejecutaGo(t, "list", "-m"))
	require.NotEmpty(t, modulo)

	for linea := range strings.SplitSeq(ejecutaGo(t, "list", "-deps", "./cmd/kitlegal"), "\n") {
		paquete := strings.TrimSpace(linea)
		assert.False(t, cuelgaDe(paquete, paqueteDeEjemplo(modulo)),
			"el binario distribuido enlaza %s: los applets de ejemplo no son funcionalidad y solo los "+
				"registra el binario de e2e (ADR 0010)", paquete)
	}
}

// TestDependenciasDelBinario comprueba que el binario distribuido no enlaza
// ningún módulo de terceros fuera de los declarados (FR-060, constitución §V).
// Mira el cierre transitivo real de `go list -deps` sobre el punto de entrada,
// que es lo mismo que acaba en `go version -m` del ejecutable: ni los módulos
// que solo usan los tests ni los de las herramientas cuentan aquí.
func TestDependenciasDelBinario(t *testing.T) {
	t.Parallel()

	modulo := strings.TrimSpace(ejecutaGo(t, "list", "-m"))
	require.NotEmpty(t, modulo)

	enlazados := map[string]bool{}

	for linea := range strings.SplitSeq(ejecutaGo(t, "list", "-deps", "-f", plantillaDeModulos, "./cmd/kitlegal"), "\n") {
		if linea = strings.TrimSpace(linea); linea != "" && linea != modulo {
			enlazados[linea] = true
		}
	}

	require.NotEmpty(t, enlazados, "el binario enlaza al menos el analizador de la línea de órdenes")

	assert.ElementsMatch(t, modulosDelBinario, slices.Sorted(maps.Keys(enlazados)),
		"el binario distribuido enlaza un módulo que no está declarado y justificado "+
			"(FR-060, constitución §V): justifícalo en plan.md y en la propuesta de cambio antes de añadirlo")
}

// grafo es el grafo de importación que devuelve `go list -deps`: el cierre
// transitivo de los paquetes pedidos, con las importaciones directas de cada
// paquete alcanzado.
type grafo struct {
	// modulo es la ruta del módulo, leída de go.mod y no escrita a mano, para
	// que renombrarlo no deje este test comprobando un prefijo que ya no existe.
	modulo string

	// importa asocia cada paquete del cierre con sus importaciones directas.
	importa map[string][]string
}

// esDelModulo dice si un paquete del cierre pertenece a este módulo. Las
// aristas hacia fuera del módulo —la biblioteca estándar y los módulos de
// terceros— no se recorren: lo que hagan por dentro no es una decisión de
// este proyecto, y seguirlas convertiría `encoding/json`, que importa `io`, en
// una violación de R1.
func (g grafo) esDelModulo(paquete string) bool {
	return cuelgaDe(paquete, g.modulo)
}

// paquetesBajo devuelve, ordenados, los paquetes del cierre que cuelgan del
// prefijo dado, el propio prefijo incluido.
func (g grafo) paquetesBajo(prefijo string) []string {
	var paquetes []string

	for paquete := range g.importa {
		if cuelgaDe(paquete, prefijo) {
			paquetes = append(paquetes, paquete)
		}
	}

	slices.Sort(paquetes)

	return paquetes
}

// reglaExclusiva describe una regla de la forma «solo estos paquetes importan
// esto»: R2 y R3.
type reglaExclusiva struct {
	// nombre es la etiqueta con la que el fallo nombra la regla violada
	// (FR-053).
	nombre string

	// razon explica, en el propio fallo, qué se ha roto y dónde va el código que
	// lo necesita.
	razon string

	// duenos son los paquetes a los que la regla concede la importación.
	duenos []string

	// denegados son las importaciones que la regla reserva a esos paquetes.
	denegados []string
}

// compruebaDominioPuro hace cumplir R1: ningún paquete de internal/core puede
// alcanzar, siguiendo aristas del módulo, el kernel, un adaptador o un paquete
// de entrada y salida de la biblioteca estándar.
//
// El recorrido es transitivo dentro del módulo y se detiene en el paquete
// denegado: interesa la arista que rompe la regla, no lo que ese paquete
// importe después. Así una cadena core → pkg/legalkit → internal/render se
// nombra entera, que es justo lo que `depguard` no puede ver mirando las
// importaciones declaradas de un fichero.
func compruebaDominioPuro(t *testing.T, g grafo) {
	t.Helper()

	dominio := g.modulo + "/internal/core"

	denegados := make([]string, 0, len(paquetesInternos)+len(entradaYSalidaEstandar))
	for _, interno := range paquetesInternos {
		denegados = append(denegados, g.modulo+"/internal/"+interno)
	}
	denegados = append(denegados, entradaYSalidaEstandar...)

	for _, origen := range g.paquetesBajo(dominio) {
		visitados := map[string]bool{origen: true}

		for cadenas := [][]string{{origen}}; len(cadenas) > 0; {
			cadena := cadenas[0]
			cadenas = cadenas[1:]
			actual := cadena[len(cadena)-1]

			for _, importacion := range g.importa[actual] {
				if denegado, hay := primerPrefijo(importacion, denegados); hay {
					t.Errorf("R1 · el dominio es puro: %s importa %q (lo prohíbe %q). "+
						"internal/core no depende del kernel, de los adaptadores ni de la entrada y "+
						"salida de la biblioteca estándar: el registro llega como *slog.Logger al método "+
						"Ejecutar y la presentación se inyecta desde la raíz de composición "+
						"(contracts/reglas-de-arquitectura.md R1).",
						strings.Join(cadena, " → "), importacion, denegado)

					continue
				}

				if g.esDelModulo(importacion) && !visitados[importacion] {
					visitados[importacion] = true
					cadenas = append(cadenas, append(slices.Clone(cadena), importacion))
				}
			}
		}
	}
}

// compruebaImportacionExclusiva hace cumplir R2 y R3: las importaciones que
// reservan pertenecen a sus dueños y a nadie más.
//
// Aquí no hace falta recorrer cadenas, y hacerlo daría falsos positivos: todo
// paquete del módulo que no sea dueño es ya un origen vigilado, de modo que un
// paquete alcanzable transitivamente se comprueba por sí mismo; en cambio
// seguir una arista hacia el dueño —internal/cli → internal/httpx → net/http—
// denunciaría precisamente el uso que la regla permite.
func compruebaImportacionExclusiva(t *testing.T, g grafo, regla reglaExclusiva) {
	t.Helper()

	for _, paquete := range g.paquetesBajo(g.modulo) {
		if _, esDueno := primerPrefijo(paquete, regla.duenos); esDueno {
			continue
		}

		for _, importacion := range g.importa[paquete] {
			if denegado, hay := primerPrefijo(importacion, regla.denegados); hay {
				t.Errorf("%s · importación reservada: %s importa %q (lo prohíbe %q). Está reservada a %s: %s.",
					regla.nombre, paquete, importacion, denegado,
					strings.Join(regla.duenos, ", "), regla.razon)
			}
		}
	}
}

// paquetesInternos son los ocho paquetes de internal/ que el dominio no puede
// alcanzar: la dependencia va siempre hacia dentro, nunca al revés.
var paquetesInternos = []string{
	"app", "cache", "cli", "graph", "httpx", "render", "source", "store",
}

// entradaYSalidaEstandar son los paquetes de entrada y salida de la biblioteca
// estándar que el dominio tampoco puede alcanzar. Sin esta mitad de R1, un
// *slog.Logger o un io.Writer podrían instalarse en internal/core sin que
// ningún control dijera nada, porque no son paquetes internos.
var entradaYSalidaEstandar = []string{
	"log", "log/slog", "os", "io", "net/http", "database/sql",
}

// grafoDelModulo construye el grafo pidiéndoselo a `go list -deps` sobre todos
// los paquetes del módulo.
func grafoDelModulo(t *testing.T) grafo {
	t.Helper()

	modulo := strings.TrimSpace(ejecutaGo(t, "list", "-m"))
	require.NotEmpty(t, modulo, "go list -m no devolvió la ruta del módulo")

	g := grafo{modulo: modulo, importa: map[string][]string{}}

	for linea := range strings.SplitSeq(ejecutaGo(t, "list", "-deps", "-f", plantillaDeListado, "./..."), "\n") {
		paquete, importaciones, _ := strings.Cut(linea, "\t")
		if paquete == "" {
			continue
		}

		g.importa[paquete] = strings.Fields(importaciones)
	}

	// Sin estas dos comprobaciones el test pasaría en vacío si `go list` dejara
	// de devolver lo que se le pide, que es la única forma en la que una de las
	// tres reglas podría quedar sin vigilar sin que nadie lo notara.
	require.NotEmpty(t, g.paquetesBajo(modulo+"/internal/core"),
		"el grafo no contiene ningún paquete del dominio: R1 quedaría sin nada que comprobar")
	require.Contains(t, g.importa, paqueteDeEjemplo(modulo),
		"el grafo no contiene el paquete de applets de ejemplo, que es la implementación de referencia")

	return g
}

// paqueteDeEjemplo es la ruta de importación del paquete de applets de ejemplo:
// la implementación de referencia que copiará cada applet posterior, sujeta a
// las mismas reglas que el resto del árbol y a una más, la de no enlazarse en
// el binario distribuido.
func paqueteDeEjemplo(modulo string) string {
	return modulo + "/internal/app/ejemplo"
}

// ejecutaGo ejecuta el go command en la raíz del módulo y devuelve su salida
// estándar. No toca la red: `go list` solo consulta el módulo y la caché.
func ejecutaGo(t *testing.T, argumentos ...string) string {
	t.Helper()

	// El ejecutable es constante y los argumentos no vienen de fuera: son
	// literales de este fichero más las rutas de paquete que sale de enumerar
	// internal/app/testdata en el propio árbol. No hay entrada de usuario, red ni
	// variable de entorno en la orden, de modo que G204 no tiene aquí nada que
	// prevenir.
	//nolint:gosec // los argumentos son literales de este fichero y rutas del propio árbol; no hay entrada externa.
	orden := exec.CommandContext(t.Context(), "go", argumentos...)
	orden.Dir = raizDelModulo

	salida, err := orden.Output()
	if err != nil {
		var fallo *exec.ExitError
		if errors.As(err, &fallo) {
			t.Fatalf("go %s falló: %v\n%s", strings.Join(argumentos, " "), err, fallo.Stderr)
		}

		t.Fatalf("go %s falló: %v", strings.Join(argumentos, " "), err)
	}

	return string(salida)
}

// cuelgaDe dice si una ruta de importación es el prefijo dado o cuelga de él,
// comparando por componente: `internal/store` no cuelga de `internal/sto`.
func cuelgaDe(paquete, prefijo string) bool {
	return paquete == prefijo || strings.HasPrefix(paquete, prefijo+"/")
}

// primerPrefijo devuelve el primer prefijo de la lista del que cuelga el
// paquete, para que el fallo pueda nombrar qué entrada concreta lo prohíbe.
func primerPrefijo(paquete string, prefijos []string) (string, bool) {
	for _, prefijo := range prefijos {
		if cuelgaDe(paquete, prefijo) {
			return prefijo, true
		}
	}

	return "", false
}
