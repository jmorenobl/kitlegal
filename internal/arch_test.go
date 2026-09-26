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
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
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
			duenos:           []string{grafo.modulo + "/internal/httpx"},
			denegados:        []string{"net/http"},
			duenoObligatorio: true,
		})
	})

	t.Run("R3 · solo internal/{cache,store,graph} importan SQLite y database/sql", func(t *testing.T) {
		t.Parallel()

		compruebaImportacionExclusiva(t, grafo, reglaExclusiva{
			nombre: "R3",
			razon: "el acceso a SQLite vive en los tres paquetes de almacenamiento —caché, almacén " +
				"y grafo—; el resto del árbol los usa a través de su interfaz " +
				"(contracts/reglas-de-arquitectura.md R3)",
			// Sin duenoObligatorio, y no por descuido. Desde H3 la regla tiene
			// su primer dueño real: internal/cache importa database/sql y el
			// controlador de SQLite, y ningún otro paquete del árbol lo hace.
			// Pero la exigencia pide los tres dueños —cada uno tiene que estar
			// en el grafo—, e internal/store e internal/graph no llegan hasta
			// H12 y H17. Activarla hoy convertiría en rojo el estado normal del
			// árbol, que es justo lo contrario de lo que la bandera sirve; se
			// activa cuando exista el último de los tres (H3 FR-037).
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
// enlaza, y ninguno más, cada uno con su justificación en su línea: los de la
// lista cerrada de la constitución §V que van en producción y los que cada uno de
// ellos arrastra consigo, que no se pueden quitar sin renunciar al que los trae.
// Los que entraron en H1 están justificados en plan.md (Complexity Tracking) y en
// gates/pr-h1.md; los que entran en H4, cuando el applet boe enlaza
// internal/httpx e internal/cache, en gates/pr-h4.md (FR-060, FR-124;
// research.md D14 de H4); el que entra en H6, cuando el applet territorio
// enlaza internal/core/territorio, en plan.md de H6 (Complexity Tracking) y en
// gates/pr-h6.md. Lo que importa cada uno lo mide `go list -deps` sobre
// el binario de cada una de plataformasDeDistribucion; el que no llega a todas
// lo dice en su línea.
//
// Es una lista escrita a mano a propósito. Cuando un hito, o una actualización
// de módulos, enlace uno nuevo, este test falla y obliga a hacer lo que la
// constitución exige: justificarlo por escrito antes de añadirlo aquí.
var modulosDelBinario = []string{
	// §V, H1: el analizador de la línea de órdenes de internal/cli.
	"github.com/alecthomas/kong",
	// H1: lo importa github.com/pb33f/ordered-map/v2, que llega con
	// github.com/invopop/jsonschema.
	"github.com/bahlo/generic-list-go",
	// H1: lo importa github.com/pb33f/ordered-map/v2.
	"github.com/buger/jsonparser",
	// H4: lo importa modernc.org/libc, el entorno de C traducido a Go sobre el
	// que corre el controlador de SQLite de internal/cache.
	"github.com/dustin/go-humanize",
	// H4: lo importa modernc.org/libc en darwin y linux; no llega a windows.
	"github.com/google/uuid",
	// §V, H1: el esquema de entrada y salida de --describe, en internal/cli.
	"github.com/invopop/jsonschema",
	// H4: lo importa modernc.org/libc en darwin y windows; no llega a linux.
	"github.com/mattn/go-isatty",
	// H4: lo importa modernc.org/libc en darwin y windows; no llega a linux.
	"github.com/ncruces/go-strftime",
	// H1: lo importa github.com/invopop/jsonschema para las propiedades en orden.
	"github.com/pb33f/ordered-map/v2",
	// H4: lo importa modernc.org/mathutil, que llega con modernc.org/libc.
	"github.com/remyoudompheng/bigfft",
	// §V, H4: internal/httpx interpreta con él el robots.txt de cada sitio antes
	// de pedirle nada.
	"github.com/temoto/robotstxt",
	// §V, H6: internal/core/territorio analiza con él los ficheros congelados de
	// data/territorio/, que son YAML como todo data/ (docs/ROADMAP.md §2). Entra
	// por internal/app, cuyas importaciones sigue `go list -deps` todas, en cuanto
	// existe el applet territorio, esté o no registrado. Es la biblioteca de YAML
	// que el repositorio fijó en H5; v4, que ya llegaba por otro camino, no tiene
	// ninguna versión estable (research.md D15 de H6).
	"go.yaml.in/yaml/v3",
	// H1: lo importa github.com/pb33f/ordered-map/v2.
	"go.yaml.in/yaml/v4",
	// H4: lo importan modernc.org/sqlite, modernc.org/libc, modernc.org/memory y
	// github.com/mattn/go-isatty, para las llamadas al sistema.
	"golang.org/x/sys",
	// §V, H4: golang.org/x/time/rate, el ritmo por sitio de internal/httpx.
	"golang.org/x/time",
	// H4: lo importa modernc.org/sqlite.
	"modernc.org/libc",
	// H4: lo importa modernc.org/libc.
	"modernc.org/mathutil",
	// H4: lo importa modernc.org/libc.
	"modernc.org/memory",
	// §V, H4: el controlador de SQLite sin cgo de internal/cache (ADR 0002).
	"modernc.org/sqlite",
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

	modulo := rutaDelModulo(t)

	for _, paquete := range paquetesDelBinario(t, modulo) {
		assert.False(t, cuelgaDe(paquete, paqueteDeEjemplo(modulo)),
			"el binario distribuido enlaza %s: los applets de ejemplo no son funcionalidad y solo los "+
				"registra el binario de e2e (ADR 0010)", paquete)
	}
}

// plataformasDeDistribucion son las plataformas para las que se entrega el
// binario, sin cgo: darwin, linux y windows sobre amd64 y arm64 (ADR 0002;
// docs/ROADMAP.md, entrega). Los módulos que enlaza dependen de la plataforma,
// porque modernc.org/libc elige sus ficheros por sistema: en linux no llegan
// github.com/mattn/go-isatty ni github.com/ncruces/go-strftime, y en windows no
// llega github.com/google/uuid. Por eso se mide en todas y no solo en la del
// ordenador que ejecuta el test, que haría depender el veredicto de dónde corre.
var plataformasDeDistribucion = []struct{ sistema, arquitectura string }{
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"windows", "amd64"},
	{"windows", "arm64"},
}

// TestDependenciasDelBinario comprueba que el binario distribuido no enlaza, en
// ninguna de sus plataformas, ningún módulo de terceros fuera de los declarados
// (FR-060, constitución §V), y que no se declara ninguno que no enlace en
// ninguna. Mira el cierre transitivo real de `go list -deps` sobre el punto de
// entrada con GOOS, GOARCH y CGO_ENABLED=0 de cada plataforma, que es lo mismo
// que acaba en `go version -m` de su ejecutable: ni los módulos que solo usan
// los tests ni los de las herramientas cuentan aquí.
func TestDependenciasDelBinario(t *testing.T) {
	t.Parallel()

	modulo := rutaDelModulo(t)
	// Cada módulo con las plataformas cuyo binario lo enlaza.
	enlazados := map[string][]string{}

	for _, plataforma := range plataformasDeDistribucion {
		nombre := plataforma.sistema + "/" + plataforma.arquitectura
		entorno := []string{"GOOS=" + plataforma.sistema, "GOARCH=" + plataforma.arquitectura, "CGO_ENABLED=0"}
		deEsta := map[string]bool{}

		for linea := range strings.SplitSeq(ejecutaGoCon(t, entorno, "list", "-deps", "-f", plantillaDeModulos, "./cmd/kitlegal"), "\n") {
			if linea = strings.TrimSpace(linea); linea != "" && linea != modulo {
				deEsta[linea] = true
			}
		}

		require.NotEmpty(t, deEsta, "en %s el binario enlaza al menos el analizador de la línea de órdenes", nombre)

		for enlazado := range deEsta {
			enlazados[enlazado] = append(enlazados[enlazado], nombre)
		}
	}

	assert.ElementsMatch(t, modulosDelBinario, slices.Sorted(maps.Keys(enlazados)),
		"el binario distribuido enlaza en alguna plataforma un módulo que no está declarado y justificado, o se "+
			"declara uno que no enlaza en ninguna (FR-060, constitución §V): justifícalo en plan.md y en la propuesta "+
			"de cambio antes de añadirlo; plataformas que enlazan cada módulo: %v", enlazados)
}

// prefijosReservados son los del espacio de nombres con el que firma lo que no
// consulta ninguna fuente pública —el kernel y los applets calculados—:
// «kitlegal.» en la fuente y «kitlegal:» en la url. Ningún adaptador de fuente
// puede usarlos (ADR 0006, «prohibición que nace en H4»; FR-002, SC-011).
var prefijosReservados = []string{"kitlegal.", "kitlegal:"}

// TestLasFuentesNoFirmanComoKitlegal comprueba que ningún adaptador de fuente
// lleva en su código de producción un literal de texto que empiece por un prefijo
// reservado, que es como podría firmar un sobre en el espacio del kernel. Lee
// con el analizador sintáctico de la biblioteca estándar los ficheros .go que no
// son de test de cada paquete bajo internal/source, que enumera `go list`, y
// nombra el fichero y la línea de cada literal (FR-002, SC-011; research.md D14
// de H4).
//
// No lo puede vigilar forbidigo, que mira identificadores y no literales, ni
// depguard, que mira importaciones. La comprobación en ejecución —la fuente y la
// url de los sobres de los seis verbos— la hacen los tests de cada verbo; esta
// alcanza también a un adaptador que no tuviera ninguno.
func TestLasFuentesNoFirmanComoKitlegal(t *testing.T) {
	t.Parallel()

	t.Run("ningún paquete de fuentes lleva un literal del espacio reservado", func(t *testing.T) {
		t.Parallel()

		carpetas := carpetasDeFuentes(t)
		require.NotEmpty(t, carpetas,
			"go list no enumera ningún paquete bajo internal/source: la comprobación pasaría en vacío")

		for _, carpeta := range carpetas {
			hallazgos, err := literalesReservados(carpeta)
			require.NoError(t, err)

			for _, hallazgo := range hallazgos {
				t.Errorf("%s: un adaptador de fuente firma en el espacio reservado (%s), que es del kernel y de "+
					"los applets calculados; la fuente y la url de un sobre de fuente son las de la fuente pública "+
					"que se consulta (ADR 0006, FR-002)", hallazgo, strings.Join(prefijosReservados, " y "))
			}
		}
	})

	t.Run("control: cada literal reservado se nombra con su fichero y su línea", func(t *testing.T) {
		t.Parallel()

		carpeta := t.TempDir()
		ficheroDeProduccion := filepath.Join(carpeta, "fuente.go")

		require.NoError(t, os.WriteFile(ficheroDeProduccion, []byte(fuenteQueFirmaComoKitlegal), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(carpeta, "fuente_test.go"),
			[]byte(testQueFirmaComoKitlegal), 0o600))

		hallazgos, err := literalesReservados(carpeta)
		require.NoError(t, err)

		assert.Equal(t, []string{
			ficheroDeProduccion + `:5: "kitlegal.fuente"`,
			ficheroDeProduccion + ":6: `kitlegal:fuente/prueba`",
		}, hallazgos, "los dos literales de producción, y ni el comentario, ni el texto que solo los contiene, "+
			"ni el fichero de test")
	})
}

// fuenteQueFirmaComoKitlegal y testQueFirmaComoKitlegal son los ficheros del
// control de TestLasFuentesNoFirmanComoKitlegal: uno de producción con los dos
// prefijos, en un literal interpretado y en uno crudo, más un comentario y un
// texto que solo los contienen; y uno de test, que la comprobación no mira.
const (
	fuenteQueFirmaComoKitlegal = "package fuente\n" +
		"\n" +
		"// Firma como «kitlegal.fuente», que un comentario sí puede decir.\n" +
		"const (\n" +
		"\tnombre = \"kitlegal.fuente\"\n" +
		"\turl    = `kitlegal:fuente/prueba`\n" +
		"\ttexto  = \"no empieza por kitlegal.fuente\"\n" +
		")\n"
	testQueFirmaComoKitlegal = "package fuente\n\nconst deTest = \"kitlegal.fuente\"\n"
)

// carpetasDeFuentes son las carpetas de los paquetes que `go list` enumera bajo
// internal/source.
func carpetasDeFuentes(t *testing.T) []string {
	t.Helper()

	var carpetas []string

	for linea := range strings.SplitSeq(ejecutaGo(t, "list", "-f", "{{.Dir}}", "./internal/source/..."), "\n") {
		if carpeta := strings.TrimSpace(linea); carpeta != "" {
			carpetas = append(carpetas, carpeta)
		}
	}

	return carpetas
}

// literalesReservados devuelve cada literal de texto de los ficheros de
// producción de la carpeta que empieza por un prefijo reservado, como
// «fichero:línea: literal» y en el orden de los ficheros y de sus líneas. Los
// ficheros de test no cuentan: los de un adaptador pueden nombrar el espacio
// reservado para comprobar que no lo usa.
func literalesReservados(carpeta string) ([]string, error) {
	entradas, err := os.ReadDir(carpeta)
	if err != nil {
		return nil, err
	}

	ficheros := token.NewFileSet()

	var hallazgos []string

	for _, entrada := range entradas {
		nombre := entrada.Name()
		if entrada.IsDir() || filepath.Ext(nombre) != ".go" || strings.HasSuffix(nombre, "_test.go") {
			continue
		}

		arbol, err := parser.ParseFile(ficheros, filepath.Join(carpeta, nombre), nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}

		for nodo := range ast.Preorder(arbol) {
			literal, esLiteral := nodo.(*ast.BasicLit)
			if !esLiteral || literal.Kind != token.STRING {
				continue
			}

			posicion := ficheros.Position(literal.Pos())

			valor, err := strconv.Unquote(literal.Value)
			if err != nil {
				return nil, fmt.Errorf("%s: el literal %s no se puede leer: %w", posicion, literal.Value, err)
			}

			reservado := slices.ContainsFunc(prefijosReservados, func(prefijo string) bool {
				return strings.HasPrefix(valor, prefijo)
			})
			if reservado {
				hallazgos = append(hallazgos, fmt.Sprintf("%s:%d: %s", posicion.Filename, posicion.Line, literal.Value))
			}
		}
	}

	return hallazgos, nil
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

	// duenoObligatorio exige que la regla tenga dueño de verdad: que sus
	// paquetes estén en el grafo y que alguno importe lo que les reserva. Una
	// regla exclusiva se cumple también cuando no hay nada que vigilar —así
	// estuvo R2 hasta que H2 trajo internal/httpx, «activa y vacía»—, y esa
	// forma de pasar no distingue «nadie la incumple» de «ya no vigila nada».
	duenoObligatorio bool
}

// compruebaDominioPuro hace cumplir R1: ningún paquete de internal/core puede
// alcanzar, siguiendo aristas del módulo, el kernel, un adaptador, lo empotrado
// en el paquete raíz del módulo o un paquete de entrada y salida de la
// biblioteca estándar.
//
// El recorrido es transitivo dentro del módulo y se detiene en el paquete
// denegado: interesa la arista que rompe la regla, no lo que ese paquete
// importe después. Así una cadena core → pkg/legalkit → internal/render se
// nombra entera, que es justo lo que `depguard` no puede ver mirando las
// importaciones declaradas de un fichero.
//
// El paquete raíz se deniega por igualdad y no por prefijo, porque del prefijo
// cuelga el módulo entero, internal/core incluido; así una arista del dominio
// a lo empotrado se nombra como tal, y no solo por el io/fs que lo empotrado
// importa (H19 research.md D32).
func compruebaDominioPuro(t *testing.T, g grafo) {
	t.Helper()

	dominio := g.modulo + "/internal/core"

	denegados := make([]string, 0, len(paquetesInternos)+len(entradaYSalidaEstandar))
	for _, interno := range paquetesInternos {
		denegados = append(denegados, g.modulo+"/internal/"+interno)
	}
	denegados = append(denegados, entradaYSalidaEstandar...)

	exactos := []string{g.modulo}

	for _, origen := range g.paquetesBajo(dominio) {
		visitados := map[string]bool{origen: true}

		for cadenas := [][]string{{origen}}; len(cadenas) > 0; {
			cadena := cadenas[0]
			cadenas = cadenas[1:]
			actual := cadena[len(cadena)-1]

			for _, importacion := range g.importa[actual] {
				if denegado, hay := denegacion(importacion, denegados, exactos); hay {
					t.Errorf("R1 · el dominio es puro: %s importa %q (lo prohíbe %q). "+
						"internal/core no depende del kernel, de los adaptadores, de lo empotrado ni de la "+
						"entrada y salida de la biblioteca estándar: el registro llega como *slog.Logger al "+
						"método Ejecutar, la presentación se inyecta desde la raíz de composición y el disco "+
						"y las skills empotradas llegan por los puertos del dominio "+
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

	if regla.duenoObligatorio {
		exigeDueno(t, g, regla)
	}

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

// exigeDueno comprueba que la regla vigila algo: sus dueños están en el grafo y
// alguno importa de verdad lo que les reserva. Sin esto, la subprueba pasaría
// igual el día que el dueño desapareciera del árbol o dejara de concentrar lo
// que concentra, que es precisamente cuando la regla deja de proteger nada
// (FR-053, research.md D18).
//
// Es el mismo cuidado que grafoDelModulo tiene con el dominio y con el paquete
// de applets de ejemplo, aplicado al otro extremo de la regla: allí se exige
// que haya a quién vigilar, aquí que haya quién sea el dueño.
func exigeDueno(t *testing.T, g grafo, regla reglaExclusiva) {
	t.Helper()

	reservado := strings.Join(regla.denegados, ", ")
	importaLoReservado := false

	for _, dueno := range regla.duenos {
		paquetes := g.paquetesBajo(dueno)

		require.NotEmpty(t, paquetes,
			"%s · el grafo no contiene %s, que es quien tiene que concentrar %s: la regla quedaría "+
				"activa y vacía, cumpliéndose porque no hay nada que vigilar",
			regla.nombre, dueno, reservado)

		for _, paquete := range paquetes {
			for _, importacion := range g.importa[paquete] {
				if _, hay := primerPrefijo(importacion, regla.denegados); hay {
					importaLoReservado = true
				}
			}
		}
	}

	assert.True(t, importaLoReservado,
		"%s · ningún paquete de %s importa %s: la regla ya no tiene dueño y volvería a pasar en vacío. "+
			"Si la concentración se ha mudado, la regla se muda con ella; si ha desaparecido, se retira "+
			"del contrato antes que de aquí.",
		regla.nombre, strings.Join(regla.duenos, ", "), reservado)
}

// paquetesInternos son los nueve paquetes de internal/ que el dominio no puede
// alcanzar: la dependencia va siempre hacia dentro, nunca al revés. disco, el
// adaptador del sistema de ficheros, entra en H19 (research.md D32).
var paquetesInternos = []string{
	"app", "cache", "cli", "disco", "graph", "httpx", "render", "source", "store",
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

	modulo := rutaDelModulo(t)
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

// paquetesDelBinario devuelve el cierre transitivo real del binario
// distribuido, que es lo que acaba enlazado en el ejecutable. Comprueba de paso
// que la lista trae su propio punto de entrada: sin eso, las dos comprobaciones
// que parten de ella pasarían en vacío el día que `go list` dejara de devolver
// lo que se le pide.
func paquetesDelBinario(t *testing.T, modulo string) []string {
	t.Helper()

	var paquetes []string

	for linea := range strings.SplitSeq(ejecutaGo(t, "list", "-deps", "./cmd/kitlegal"), "\n") {
		if paquete := strings.TrimSpace(linea); paquete != "" {
			paquetes = append(paquetes, paquete)
		}
	}

	require.Contains(t, paquetes, modulo+"/cmd/kitlegal",
		"el cierre del binario no contiene ni siquiera su propio punto de entrada")

	return paquetes
}

// rutaDelModulo lee la ruta del módulo de go.mod en lugar de escribirla a mano,
// para que renombrarlo no deje ningún control comprobando un prefijo que ya no
// existe.
func rutaDelModulo(t *testing.T) string {
	t.Helper()

	modulo := strings.TrimSpace(ejecutaGo(t, "list", "-m"))
	require.NotEmpty(t, modulo, "go list -m no devolvió la ruta del módulo")

	return modulo
}

// ejecutaGo ejecuta el go command en la raíz del módulo, con el entorno del
// proceso, y devuelve su salida estándar. No toca la red: `go list` solo
// consulta el módulo y la caché.
func ejecutaGo(t *testing.T, argumentos ...string) string {
	t.Helper()

	return ejecutaGoCon(t, nil, argumentos...)
}

// ejecutaGoCon es ejecutaGo con variables de entorno añadidas a las del
// proceso, que ganan a las heredadas: con GOOS y GOARCH, `go list`
// describe el cierre de otra plataforma sin compilar nada para ella.
func ejecutaGoCon(t *testing.T, entorno []string, argumentos ...string) string {
	t.Helper()

	// El ejecutable es constante y ni los argumentos ni las variables añadidas
	// vienen de fuera: son literales de este fichero, las rutas de paquete que
	// sale de enumerar internal/app/testdata en el propio árbol y las plataformas
	// de plataformasDeDistribucion. No hay entrada de usuario ni red en la orden,
	// de modo que G204 no tiene aquí nada que prevenir.
	//nolint:gosec // los argumentos son literales de este fichero y rutas del propio árbol; no hay entrada externa.
	orden := exec.CommandContext(t.Context(), "go", argumentos...)
	orden.Dir = raizDelModulo
	orden.Env = append(os.Environ(), entorno...)

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

// denegacion devuelve la entrada que prohíbe la importación: la propia
// importación si es uno de los exactos, que se comparan por igualdad, o el
// primer prefijo de la lista del que cuelga.
func denegacion(importacion string, prefijos, exactos []string) (string, bool) {
	if slices.Contains(exactos, importacion) {
		return importacion, true
	}

	return primerPrefijo(importacion, prefijos)
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
