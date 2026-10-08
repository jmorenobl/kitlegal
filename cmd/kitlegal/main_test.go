package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// appletsDelBinario es la lista de applets que enumera el kernel ante una
// invocación que no resuelve ninguno: la evidencia observable de que el binario
// que se publica registra boe, cita, graph, mcp, skills y territorio, y solo esos,
// desde H23 (FR-001, contracts/registro-y-describe.md §3 de H1; contrato del
// applet territorio §7; contracts/applet-skills.md §1 de H19;
// contracts/applet-graph.md §1 de H7; contracts/servidor-mcp.md §1 de H21;
// contracts/applet-cita.md §1 de H23).
const appletsDelBinario = "applets disponibles: boe, cita, graph, mcp, skills, territorio"

// TestPuntoDeEntrada ejerce el contrato observable del binario distribuido con la
// **misma composición que main()** —app.Arrancar con el registro de producción y
// los datos de construcción de este paquete— y dos escritores en memoria en lugar
// de los descriptores del sistema, que es lo que hace comprobable el contrato
// entero sin lanzar ningún subproceso (FR-035; contrato puerto-y-applet §5 de H4).
//
// De contracts/cli-version.md de H0 se conservan las tres líneas de version, la
// salida de error vacía y el código 0 (D16), y también el código 2 de cualquier
// otra invocación: lo que cambia es el mensaje. Un nombre que no es ningún
// applet lo resuelve el despacho, que **nombra lo desconocido** y enumera lo
// disponible —boe, cita, graph, mcp, skills y territorio, en este binario—
// (FR-006, contracts/registro-y-describe.md §2 y §3); lo que sobra tras
// «version», que no admite argumentos ni banderas, se nombra en el mensaje en
// lugar de descartarse (FR-027); y boe, cita, graph, mcp, skills y territorio sin
// verbo se corrigen igual, porque ninguno declara verbo por omisión (FR-001,
// FR-050; contracts/applet-skills.md §1 de H19; contracts/servidor-mcp.md §1 de
// H21; H23 FR-001).
// Ningún caso ejecuta un verbo, así que nada de esta tabla pide nada, abre la
// caché o world.db, examina el disco ni lee de la entrada estándar.
//
// No es paralelo, y no es un descuido: fija KITLEGAL_LOG —en el test y en cada
// subcaso, que es lo que lo deja hermético por separado— para que el nivel del
// registro de eventos no dependa del entorno de quien ejecuta los tests, y eso
// es estado del proceso entero.
func TestPuntoDeEntrada(t *testing.T) {
	t.Setenv(cli.VariableNivel, "")

	// La salida de version se compone a partir de las variables del paquete y no
	// de un literal, de modo que el caso siga siendo válido cuando -ldflags las
	// inyecte (FR-004).
	salidaVersion := "kitlegal " + version + "\ncommit: " + commit + "\nfecha:  " + fecha + "\n"

	casos := []struct {
		nombre string
		argv   []string
		salida string
		// errores son los fragmentos que el mensaje para la persona tiene que
		// llevar; vacío significa que la salida de error queda vacía.
		errores []string
		codigo  int
	}{
		{
			nombre: "version escribe los tres datos y termina con 0",
			argv:   []string{"kitlegal", "version"},
			salida: salidaVersion,
			codigo: 0,
		},
		{
			// El verbo reservado se reconoce **antes** que el registro y no
			// admite nada detrás: «version» no tiene sobre ni banderas, y lo que
			// sobra es una invocación que hay que corregir —código 2, como en
			// H0— con un mensaje que nombra lo que sobra (FR-027, D16,
			// contracts/cli-version.md de H0).
			nombre:  "un argumento de más tras version termina con 2 y se nombra",
			argv:    []string{"kitlegal", "version", "extra"},
			errores: []string{`"version"`, `"extra"`},
			codigo:  2,
		},
		{
			nombre:  "una bandera tras version tampoco se admite",
			argv:    []string{"kitlegal", "version", "--jsno"},
			errores: []string{`"version"`, `"--jsno"`},
			codigo:  2,
		},
		{
			nombre:  "sin applet, el fallo enumera lo que hay y termina con 2",
			argv:    []string{"kitlegal"},
			errores: []string{"no se ha indicado ningún applet", appletsDelBinario},
			codigo:  2,
		},
		{
			nombre:  "un applet que no existe se nombra en el fallo y termina con 2",
			argv:    []string{"kitlegal", "inventado"},
			errores: []string{`"inventado"`, appletsDelBinario},
			codigo:  2,
		},
		{
			// El applet de ejemplo vive en el binario que compila el test e2e y
			// nunca en el que se publica: sobre el distribuido es un nombre
			// desconocido como cualquier otro
			// (contracts/registro-y-describe.md §3).
			nombre:  "echo no es del binario distribuido, sino del binario del e2e",
			argv:    []string{"kitlegal", "echo", "hola"},
			errores: []string{`"echo"`, appletsDelBinario},
			codigo:  2,
		},
		{
			// boe no declara ningún verbo por omisión, así que nombrarlo es
			// obligatorio: sin él la invocación se corrige —código 2— con un
			// mensaje que nombra el applet y enumera sus seis verbos (FR-001).
			nombre:  "boe sin verbo termina con 2 y enumera sus verbos",
			argv:    []string{"kitlegal", "boe"},
			errores: []string{`"boe"`, "verbos de boe: buscar, indice, articulo, articulos, metadatos, analisis"},
			codigo:  2,
		},
		{
			// cita tampoco declara verbo por omisión: preparar y cotejar se
			// nombran siempre (contracts/applet-cita.md §1 de H23).
			nombre:  "cita sin verbo termina con 2 y enumera sus verbos",
			argv:    []string{"kitlegal", "cita"},
			errores: []string{`"cita"`, "verbos de cita: preparar, cotejar"},
			codigo:  2,
		},
		{
			// graph tampoco declara verbo por omisión: show, stats y check se
			// nombran siempre (contracts/applet-graph.md §1).
			nombre:  "graph sin verbo termina con 2 y enumera sus verbos",
			argv:    []string{"kitlegal", "graph"},
			errores: []string{`"graph"`, "verbos de graph: show, stats, check"},
			codigo:  2,
		},
		{
			// mcp tampoco declara verbo por omisión: serve se nombra siempre, y
			// sin él el servidor no arranca (contracts/servidor-mcp.md §1 de
			// H21).
			nombre:  "mcp sin verbo termina con 2 y enumera su verbo",
			argv:    []string{"kitlegal", "mcp"},
			errores: []string{`"mcp"`, "verbos de mcp: serve"},
			codigo:  2,
		},
		{
			// skills tampoco declara verbo por omisión: install, list y doctor
			// se nombran siempre (contracts/applet-skills.md §1 de H19).
			nombre:  "skills sin verbo termina con 2 y enumera sus verbos",
			argv:    []string{"kitlegal", "skills"},
			errores: []string{`"skills"`, "verbos de skills: install, list, doctor"},
			codigo:  2,
		},
		{
			// territorio tampoco declara verbo por omisión: nombrar resolver es
			// obligatorio (contrato del applet territorio §1).
			nombre:  "territorio sin verbo termina con 2 y enumera su verbo",
			argv:    []string{"kitlegal", "territorio"},
			errores: []string{`"territorio"`, "verbos de territorio: resolver"},
			codigo:  2,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Setenv(cli.VariableNivel, "")

			var salida, errores bytes.Buffer

			codigo := app.Arrancar(caso.argv, app.RegistroDeProduccion,
				&salida, &errores, version, commit, fecha)

			assert.Equal(t, caso.codigo, codigo, "código de salida")
			assert.Equal(t, caso.salida, salida.String(), "salida estándar")

			if len(caso.errores) == 0 {
				assert.Empty(t, errores.String(), "la salida de error queda vacía")
			}

			for _, fragmento := range caso.errores {
				assert.Contains(t, errores.String(), fragmento,
					"el mensaje para la persona dice qué ha pasado")
			}
		})
	}
}

// TestValoresPorDefecto fija los valores que el código lleva cuando nadie los
// inyecta: son los que ven `go run` y `go test`, y los que un binario de
// make build o make install nunca debe mostrar (FR-004, data-model.md R1.2).
func TestValoresPorDefecto(t *testing.T) {
	t.Parallel()

	if version != "dev" || commit != "none" || fecha != "unknown" {
		t.Errorf("valores por defecto = %q, %q, %q; se esperaban %q, %q, %q",
			version, commit, fecha, "dev", "none", "unknown")
	}
}

// TestVersionDeLaIdentificacion ata la versión con la que kitlegal se presenta
// en la red a la que imprime su verbo version: son dos variables de paquete
// distintas —main.version y la de internal/httpx—, y el Makefile inyecta en
// las dos la misma VERSION, así que lo único que podría separarlas es que sus
// valores por omisión divergieran. Este caso lo impide: sin -ldflags los dos
// valen «dev», y cambiar uno solo deja make ci en rojo (FR-007, research.md D5).
//
// Desde H4 el binario distribuido enlaza internal/httpx a través del applet boe,
// y la identificación que aquí se comprueba es la que acompaña a cada petición
// que ese binario emite. Qué módulos arrastra consigo lo fija
// TestDependenciasDelBinario (FR-124).
func TestVersionDeLaIdentificacion(t *testing.T) {
	t.Parallel()

	// El espacio final acota la versión: sin él, «kitlegal/dev» también casaría
	// dentro de «kitlegal/dev-2», y la divergencia pasaría inadvertida.
	assert.Contains(t, httpx.AgenteDeUsuario(), "kitlegal/"+version+" ",
		"la identificación lleva la misma versión que el verbo version de este binario (FR-007)")
}
