package evals

import (
	"fmt"
	"maps"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// casosDeLeerTrazas es el directorio de los casos de TestLeerTrazas: cada
// subdirectorio es el traza/ de una sesión (contrato job-de-evals §9.1).
const casosDeLeerTrazas = "testdata/sesiones/leer-trazas"

// boeDeLaSkillInstalada es el nombre de invocación con el que la skill instalada
// llama al applet boe en las trazas sintéticas.
const boeDeLaSkillInstalada = "/home/runner/.claude/skills/boe-legislacion/scripts/boe"

// Lo que leen las invocaciones de las trazas sintéticas.
const (
	normaDeLasTrazas = "BOE-A-2015-10565"
	destinoPublico   = "203.0.113.7:443"
	destinoPublico6  = "[2001:db8::7]:443"
	destinoDeBucle   = "127.0.0.1:9"
)

// Resultados de connect tal como los escribe strace (research.md V53 y V54); el
// de la llamada que el fin del proceso deja sin terminar, ?, es
// resultadoSinTerminar, del propio paquete (V65).
const (
	resultadoEnCurso      = "-1 EINPROGRESS (Operation now in progress)"
	resultadoRechazada    = "-1 ECONNREFUSED (Connection refused)"
	resultadoSinRuta      = "-1 ENETUNREACH (Network is unreachable)"
	resultadoSinFichero   = "-1 ENOENT (No such file or directory)"
	resultadoInterrumpida = "? ERESTARTSYS (To be restarted if SA_RESTART is set)"
)

// defectoEsperado es lo que el error de una traza ilegible tiene que nombrar:
// los ficheros del caso, la línea con su texto (si el defecto es de una línea),
// un fragmento del motivo y lo que no puede nombrar.
type defectoEsperado struct {
	ficheros []string
	linea    int
	motivo   string
	ajenos   []string
}

// TestLeerTrazas fija la lectura de la traza de una sesión de data-model §9: la
// atribución de hilos y procesos por clone y clone3, también desde un hilo que no
// es el principal, y la del proceso que crea vfork con el relleno de alineación de
// strace, como en el runner de x86_64; el argv decodificado; el código de salida, la muerte por señal
// y la invocación sin código que deja el corte; las clases de conexión; las
// líneas de señal, que no cuentan; la llamada que el fin del proceso deja sin
// terminar, con la línea real del runner, en cualquier sesión, y el fichero del
// hilo que una clone así pudo crear; y la traza ilegible, con el fichero, la línea
// y su texto, en lugar de ignorar lo que no se entiende (contrato job-de-evals
// §9; FR-072, FR-076).
func TestLeerTrazas(t *testing.T) {
	t.Parallel()

	a21 := []string{normaDeLasTrazas, "a21"}
	a9998 := []string{normaDeLasTrazas, "a9998"}
	a140 := []string{"BOE-A-1978-31229", "a140"}
	json := []string{"--json"}

	casos := []struct {
		nombre       string
		cortada      bool
		invocaciones []Invocacion
		defecto      *defectoEsperado
	}{
		{
			nombre: "argv-escapado",
			invocaciones: []Invocacion{
				invocacionDeBoe(2000, codigoDeSalida(0), "buscar", []string{"código", "notificación"}, json),
			},
		},
		{
			nombre: "salida-con-codigo",
			invocaciones: []Invocacion{
				invocacionDeBoe(2000, codigoDeSalida(4), "articulo", a9998, []string{"--offline", "--json"}),
			},
		},
		{
			nombre:       "muerte-por-senal",
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(-1), "articulo", a21, json)},
		},
		{
			nombre: "hilo-por-clone",
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(0), "articulo", a21, json,
				conexionInet(destinoPublico, resultadoEnCurso, ConexionRed))},
		},
		{
			nombre: "hilo-por-clone3",
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(5), "articulo", a9998, json,
				conexionInet(destinoDeBucle, resultadoRechazada, ConexionLocal))},
		},
		{
			nombre: "hilo-de-un-hilo",
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(0), "articulo", a21, json,
				conexionInet(destinoPublico, resultadoEnCurso, ConexionRed))},
		},
		{
			nombre:       "connect-fuera-de-la-invocacion",
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(0), "articulo", a21, json)},
		},
		{
			nombre:       "proceso-por-vfork",
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(0), "articulo", a21, json)},
		},
		{
			nombre: "connect-local-rechazado",
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(5), "articulo", a9998, json,
				conexionInet(destinoDeBucle, resultadoRechazada, ConexionLocal))},
		},
		{
			nombre: "connect-publico-rechazado",
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(4), "articulo", a21, json,
				conexionInet(destinoPublico, resultadoSinRuta, ConexionBloqueada))},
		},
		{
			nombre: "connect-publico-aceptado",
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(0), "articulo", a21, json,
				conexionInet(destinoPublico, "0", ConexionRed))},
		},
		{
			nombre: "connect-publico-en-curso",
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(0), "articulo", a21, json,
				conexionInet(destinoPublico, resultadoEnCurso, ConexionRed))},
		},
		{
			nombre: "connect-ipv6-publico-en-curso",
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(0), "articulo", a21, json,
				conexionInet(destinoPublico6, resultadoEnCurso, ConexionRed))},
		},
		{
			nombre: "connect-af-unix",
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(0), "articulo", a21, json, Conexion{
				Familia:   "AF_UNIX",
				Ruta:      "/var/run/nscd/socket",
				Resultado: resultadoSinFichero,
				Clase:     ConexionLocal,
			})},
		},
		{
			nombre:  "fichero-sin-origen",
			defecto: &defectoEsperado{ficheros: []string{"t.2000", "t.2002"}, motivo: "que los crea"},
		},
		{
			nombre:  "fichero-ilegible",
			defecto: &defectoEsperado{ficheros: []string{"t.2000"}, linea: 2, ajenos: []string{"línea 1", "línea 3"}},
		},
		{
			nombre: "lineas-de-senal",
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(5), "articulo", a9998, json,
				conexionInet(destinoDeBucle, resultadoEnCurso, ConexionLocal))},
		},
		{
			nombre:  "cortada-por-el-tope",
			cortada: true,
			invocaciones: []Invocacion{invocacionDeBoe(2000, nil, "articulo", a9998, json,
				conexionInet(destinoPublico, resultadoEnCurso, ConexionRed))},
		},
		{
			nombre: "sin-linea-final-sin-corte",
			defecto: &defectoEsperado{
				ficheros: []string{"t.2000"},
				motivo:   "línea final",
				ajenos:   []string{"t.2001"},
			},
		},
		{
			nombre:  "cortada-con-llamada-interrumpida",
			cortada: true,
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(-1), "articulo", a9998, json,
				conexionInet(destinoPublico, resultadoInterrumpida, ConexionRed))},
		},
		{
			nombre:  "llamada-interrumpida-sin-corte",
			defecto: &defectoEsperado{ficheros: []string{"t.2001"}, linea: 1, motivo: "sin resultado"},
		},
		{
			nombre:  "llamada-interrumpida-antes-del-final",
			cortada: true,
			defecto: &defectoEsperado{
				ficheros: []string{"t.2000"},
				linea:    2,
				motivo:   "sin resultado",
				ajenos:   []string{"línea 3"},
			},
		},
		{
			// La línea real del runner: una clone que el fin del proceso dejó sin
			// terminar, en una sesión sin corte, no crea ningún hilo.
			nombre:       "clone-sin-terminar",
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(0), "articulo", a140, json)},
		},
		{
			// El fichero del hilo que esa clone pudo crear, sin la línea que lo
			// crea y sin ninguna llamada, no es un defecto ni una invocación.
			nombre:       "hilo-de-clone-sin-terminar",
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(0), "articulo", a140, json)},
		},
		{
			// La marca delante del paréntesis de cierre también es opcional en
			// connect, aunque strace no la escriba en esa llamada (V65): la conexión
			// queda sin resultado y, fuera del bucle local, es de clase red.
			nombre: "connect-sin-terminar",
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(0), "articulo", a9998, json,
				conexionInet(destinoPublico, resultadoSinTerminar, ConexionRed))},
		},
		{
			nombre:  "huerfano-sin-clone-sin-terminar",
			defecto: &defectoEsperado{ficheros: []string{"t.2002"}, motivo: "quedó sin terminar"},
		},
		{
			nombre:  "clone-sin-terminar-seguido-de-otra-llamada",
			defecto: &defectoEsperado{ficheros: []string{"t.2001"}, linea: 1, motivo: "solo pueden seguirla"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			comprobarLectura(t, filepath.Join(casosDeLeerTrazas, caso.nombre), caso.cortada, caso.invocaciones, caso.defecto)
		})
	}
}

// comprobarLectura lee la traza de dir y comprueba lo que el caso espera: sin
// defecto, exactamente esas invocaciones; con defecto, un error que nombra sus
// ficheros, la línea y su texto si el defecto es de una línea, y el motivo, y
// que no nombra lo ajeno.
func comprobarLectura(t *testing.T, dir string, cortada bool, invocaciones []Invocacion, defecto *defectoEsperado) {
	t.Helper()

	leidas, err := LeerTrazas(dir, cortada)

	if defecto == nil {
		require.NoError(t, err)
		assert.Equal(t, invocaciones, leidas)

		return
	}

	require.Error(t, err)
	assert.Nil(t, leidas, "una traza ilegible no devuelve ninguna invocación")

	for _, fichero := range defecto.ficheros {
		require.ErrorContains(t, err, filepath.Join(dir, fichero))
	}

	if defecto.linea > 0 {
		require.ErrorContains(t, err, fmt.Sprintf("línea %d", defecto.linea))
		require.ErrorContains(t, err, lineaDeLaTraza(t, filepath.Join(dir, defecto.ficheros[0]), defecto.linea))
	}

	require.ErrorContains(t, err, defecto.motivo)

	for _, ajeno := range defecto.ajenos {
		assert.NotContains(t, err.Error(), ajeno)
	}
}

// TestLeerTrazasSinFicheros fija la regla 1 de data-model §9 sin ningún fichero
// y el error de un directorio que no se puede listar, sobre t.TempDir(): un
// traza/ vacío no deja la lista de invocaciones vacía en silencio (contrato
// job-de-evals §9).
func TestLeerTrazasSinFicheros(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		crear  bool
		motivo string
	}{
		{nombre: "vacio", crear: true, motivo: "ningún fichero sin la línea"},
		{nombre: "inexistente", motivo: "no se puede leer"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			dir := filepath.Join(t.TempDir(), "traza")
			if caso.crear {
				require.NoError(t, os.Mkdir(dir, 0o700))
			}

			invocaciones, err := LeerTrazas(dir, false)

			require.Error(t, err)
			require.ErrorContains(t, err, dir)
			require.ErrorContains(t, err, caso.motivo)
			assert.Nil(t, invocaciones)
		})
	}
}

// TestInterpretarInvocacion fija cómo se lee un argv como invocación de un applet
// (contrato evals-y-grabaciones §6; data-model §9; US4, escenario 7): el applet
// por el nombre de invocación multicall; el verbo y los argumentos sin las
// banderas globales ni la ayuda, con su valor en --timeout y --asunto; sin
// consulta con la ayuda, --help o -h, también con un valor falso, y con
// --describe o --dry-run verdaderos, sin valor o con él, y con consulta si valen
// falso o su valor no se puede analizar; y lo que no invoca un applet registrado
// se ignora.
func TestInterpretarInvocacion(t *testing.T) {
	t.Parallel()

	a21 := []string{normaDeLasTrazas, "a21"}

	casos := []struct {
		nombre     string
		argv       []string
		deApplet   bool
		invocacion Invocacion
	}{
		{
			nombre:     "scripts-boe",
			argv:       []string{"scripts/boe", "articulo", normaDeLasTrazas, "a21"},
			deApplet:   true,
			invocacion: Invocacion{Applet: "boe", Verbo: "articulo", Argumentos: a21, Consulta: true},
		},
		{
			nombre:   "kitlegal-boe-articulos",
			argv:     []string{"/ruta/kitlegal", "boe", "articulos", normaDeLasTrazas, "a21", "a22"},
			deApplet: true,
			invocacion: Invocacion{
				Applet:     "boe",
				Verbo:      "articulos",
				Argumentos: []string{normaDeLasTrazas, "a21", "a22"},
				Consulta:   true,
			},
		},
		{
			nombre:     "banderas-antes-y-despues-del-verbo",
			argv:       []string{"scripts/boe", "--json", "--offline", "articulo", normaDeLasTrazas, "--no-graph", "a21", "--verbose"},
			deApplet:   true,
			invocacion: Invocacion{Applet: "boe", Verbo: "articulo", Argumentos: a21, Consulta: true},
		},
		{
			nombre:     "timeout-separado",
			argv:       []string{"scripts/boe", "--timeout", "5s", "articulo", normaDeLasTrazas, "a21"},
			deApplet:   true,
			invocacion: Invocacion{Applet: "boe", Verbo: "articulo", Argumentos: a21, Consulta: true},
		},
		{
			nombre:     "timeout-con-igual",
			argv:       []string{"scripts/boe", "articulo", normaDeLasTrazas, "a21", "--timeout=5s"},
			deApplet:   true,
			invocacion: Invocacion{Applet: "boe", Verbo: "articulo", Argumentos: a21, Consulta: true},
		},
		{
			nombre:     "asunto",
			argv:       []string{"scripts/boe", "articulo", normaDeLasTrazas, "--asunto", "x", "a21"},
			deApplet:   true,
			invocacion: Invocacion{Applet: "boe", Verbo: "articulo", Argumentos: a21, Consulta: true},
		},
		{
			nombre:     "describe-sin-consulta",
			argv:       []string{"scripts/boe", "articulo", normaDeLasTrazas, "a21", "--describe"},
			deApplet:   true,
			invocacion: Invocacion{Applet: "boe", Verbo: "articulo", Argumentos: a21},
		},
		{
			nombre:     "dry-run-sin-consulta",
			argv:       []string{"/ruta/kitlegal", "boe", "articulo", normaDeLasTrazas, "a21", "--dry-run", "--json"},
			deApplet:   true,
			invocacion: Invocacion{Applet: "boe", Verbo: "articulo", Argumentos: a21},
		},
		{
			nombre:     "describe-verdadero-sin-consulta",
			argv:       []string{"scripts/boe", "articulo", normaDeLasTrazas, "a21", "--describe=true"},
			deApplet:   true,
			invocacion: Invocacion{Applet: "boe", Verbo: "articulo", Argumentos: a21},
		},
		{
			// Con un valor falso, el binario no describe ni ensaya: consulta.
			nombre:     "describe-y-dry-run-falsos-con-consulta",
			argv:       []string{"scripts/boe", "articulo", normaDeLasTrazas, "a21", "--describe=false", "--dry-run=0"},
			deApplet:   true,
			invocacion: Invocacion{Applet: "boe", Verbo: "articulo", Argumentos: a21, Consulta: true},
		},
		{
			// Un valor que el analizador rechaza es un error de argumentos: el
			// binario no describe nada y la invocación cuenta como consulta.
			nombre:     "describe-con-valor-invalido-con-consulta",
			argv:       []string{"scripts/boe", "articulo", normaDeLasTrazas, "a21", "--describe=quizá"},
			deApplet:   true,
			invocacion: Invocacion{Applet: "boe", Verbo: "articulo", Argumentos: a21, Consulta: true},
		},
		{
			nombre:     "ayuda-sin-consulta",
			argv:       []string{"scripts/boe", "articulo", normaDeLasTrazas, "a21", "--help"},
			deApplet:   true,
			invocacion: Invocacion{Applet: "boe", Verbo: "articulo", Argumentos: a21},
		},
		{
			nombre:     "ayuda-corta-sin-consulta",
			argv:       []string{"/ruta/kitlegal", "boe", "-h", "articulo", normaDeLasTrazas, "a21"},
			deApplet:   true,
			invocacion: Invocacion{Applet: "boe", Verbo: "articulo", Argumentos: a21},
		},
		{
			// El analizador imprime la ayuda en cuanto aparece la bandera, también
			// con un valor falso.
			nombre:     "ayuda-con-valor-falso-sin-consulta",
			argv:       []string{"scripts/boe", "articulo", normaDeLasTrazas, "a21", "--help=false"},
			deApplet:   true,
			invocacion: Invocacion{Applet: "boe", Verbo: "articulo", Argumentos: a21},
		},
		{
			nombre: "kitlegal-version-ignorada",
			argv:   []string{"kitlegal", "version"},
		},
		{
			nombre: "applet-desconocido-ignorado",
			argv:   []string{"/ruta/kitlegal", "desconocido", "buscar", "obras"},
		},
		{
			nombre: "programa-que-no-es-un-applet-ignorado",
			argv:   []string{"/bin/bash", "-c", "scripts/boe articulo BOE-A-2015-10565 a21"},
		},
		{
			nombre: "kitlegal-sin-applet-ignorado",
			argv:   []string{"kitlegal"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			esperada := caso.invocacion
			if caso.deApplet {
				esperada.Argv = slices.Clone(caso.argv)
			}

			invocacion, deApplet, err := InterpretarInvocacion(caso.argv)

			require.NoError(t, err)
			assert.Equal(t, caso.deApplet, deApplet)
			assert.Equal(t, esperada, invocacion)
		})
	}
}

// invocacionDeBoe es la invocación que el caso espera del applet boe con el
// nombre de invocación de la skill instalada: argv con el verbo, los argumentos y
// las banderas detrás, en ese orden, y con consulta.
func invocacionDeBoe(proceso int, codigo *int, verbo string, argumentos, banderas []string,
	conexiones ...Conexion,
) Invocacion {
	return Invocacion{
		Proceso:    proceso,
		Argv:       slices.Concat([]string{boeDeLaSkillInstalada, verbo}, argumentos, banderas),
		Applet:     "boe",
		Verbo:      verbo,
		Argumentos: argumentos,
		Consulta:   true,
		Codigo:     codigo,
		Conexiones: conexiones,
	}
}

// conexionInet es la conexión AF_INET o AF_INET6, según la dirección, con el
// resultado de strace y la clase que el caso espera; sin resultado si el
// resultado es el de una llamada que el corte interrumpió (? ERRNO (…)) o que el
// fin del proceso dejó sin terminar (?, ? <unavailable> o, sin cerrar, vacío).
func conexionInet(destino, resultado string, clase ClaseDeConexion) Conexion {
	direccion := netip.MustParseAddrPort(destino)

	familia := "AF_INET6"
	if direccion.Addr().Is4() {
		familia = "AF_INET"
	}

	return Conexion{
		Familia:      familia,
		Direccion:    direccion,
		Resultado:    resultado,
		SinResultado: resultado == "" || strings.HasPrefix(resultado, "?"),
		Clase:        clase,
	}
}

// codigoDeSalida es el código de una invocación que terminó con él.
func codigoDeSalida(codigo int) *int {
	return &codigo
}

// lineaDeLaTraza es el texto de la línea numero (desde 1) de un fichero de traza,
// sin su salto de línea: lo que el error de una línea ilegible tiene que citar.
func lineaDeLaTraza(t *testing.T, ruta string, numero int) string {
	t.Helper()

	contenido, err := os.ReadFile(filepath.Clean(ruta))
	require.NoError(t, err)

	lineas := strings.Split(string(contenido), "\n")
	require.Greater(t, len(lineas), numero-1, "el fichero %s no tiene la línea %d", ruta, numero)

	return lineas[numero-1]
}

// Líneas de las trazas que escriben los tests sobre t.TempDir(), con las formas
// de research.md V53 y su salto de línea: la execve de la invocación de la skill
// instalada, la de bash y la línea final de un proceso que termina con 0.
const (
	lineaDeExecveDeBoe = `execve("` + boeDeLaSkillInstalada + `", ["` + boeDeLaSkillInstalada + `", "articulo", ` +
		`"BOE-A-2015-10565", "a21", "--json"], 0x7ffc3e7a1b88 /* 25 vars */) = 0` + "\n"
	lineaDeExecveDeBash = `execve("/usr/bin/bash", ["bash", "-c", "scripts/boe articulo BOE-A-2015-10565 a21"], ` +
		`0x7ffd5a1e2c40 /* 25 vars */) = 0` + "\n"
	lineaFinalConCero = "+++ exited with 0 +++\n"
)

// TestLeerTrazasConDefectos fija, sobre trazas que el propio test escribe en
// t.TempDir() con líneas literales, los defectos que hacen ilegible una traza y
// que ningún caso de §9.1 tiene (data-model §9, reglas 1, 2 y 5; FR-076): una
// entrada que no es un fichero t.<n> o cuyo número no cabe en un entero; una
// línea cortada o posterior a la final; un código final que no es de 0 a 255 y
// un resultado que no cabe en un entero; los argumentos de execve y de connect
// sin la forma de la traza o con una secuencia de escape inválida; un connect de
// otra familia en una invocación; y un hilo creado por dos líneas o que no
// desciende del fichero raíz. El error nombra el fichero y, si el defecto es de
// una línea, su número y su texto.
func TestLeerTrazasConDefectos(t *testing.T) {
	t.Parallel()

	invocacion := lineaDeExecveDeBoe + lineaFinalConCero

	casos := []struct {
		nombre   string
		hilos    map[string]string
		carpetas []string

		// fichero es el que nombra el error, dentro de la traza; linea, la del
		// defecto en ese fichero, o 0 si el defecto no es de una línea.
		fichero string
		linea   int

		// fragmentos es lo demás que dice el error.
		fragmentos []string
	}{
		{
			nombre:     "entrada-con-otro-nombre",
			hilos:      map[string]string{"t.2000": invocacion, "strace.log": lineaFinalConCero},
			fichero:    "strace.log",
			fragmentos: []string{"no es un fichero t.<n> de strace"},
		},
		{
			nombre:     "hilo-que-es-una-carpeta",
			hilos:      map[string]string{"t.2000": invocacion},
			carpetas:   []string{"t.2001"},
			fichero:    "t.2001",
			fragmentos: []string{"no es un fichero t.<n> de strace"},
		},
		{
			nombre:     "numero-de-hilo-que-no-cabe-en-un-entero",
			hilos:      map[string]string{"t.2000": invocacion, "t.99999999999999999999": lineaFinalConCero},
			fichero:    "t.99999999999999999999",
			fragmentos: []string{"el número del hilo no es un entero"},
		},
		{
			nombre:     "linea-cortada",
			hilos:      map[string]string{"t.2000": lineaDeExecveDeBoe + "+++ exited with 0 +++"},
			fichero:    "t.2000",
			linea:      2,
			fragmentos: []string{"la línea no termina en un salto de línea: está cortada"},
		},
		{
			nombre: "linea-tras-la-final",
			hilos: map[string]string{"t.2000": invocacion +
				`connect(9, {sa_family=AF_INET, sin_port=htons(443), sin_addr=inet_addr("203.0.113.7")}, 16) = 0` + "\n"},
			fichero:    "t.2000",
			linea:      3,
			fragmentos: []string{"la línea sigue a la línea final del fichero"},
		},
		{
			nombre:     "codigo-final-mayor-que-255",
			hilos:      map[string]string{"t.2000": lineaDeExecveDeBoe + "+++ exited with 256 +++\n"},
			fichero:    "t.2000",
			linea:      2,
			fragmentos: []string{"el código de la línea final, 256, no es un código de salida de 0 a 255"},
		},
		{
			nombre:     "codigo-final-que-no-cabe-en-un-entero",
			hilos:      map[string]string{"t.2000": lineaDeExecveDeBoe + "+++ exited with 99999999999999999999 +++\n"},
			fichero:    "t.2000",
			linea:      2,
			fragmentos: []string{"el código de la línea final no es un entero"},
		},
		{
			nombre: "resultado-que-no-cabe-en-un-entero",
			hilos: map[string]string{
				"t.2000": lineaDeExecveDeBash + lineaDeCloneDeBash("99999999999999999999") + lineaFinalConCero,
			},
			fichero:    "t.2000",
			linea:      2,
			fragmentos: []string{"el resultado no es un entero"},
		},
		{
			// strace escribe así los argumentos que no puede leer de la memoria
			// del proceso.
			nombre:     "execve-con-argumentos-sin-leer",
			hilos:      map[string]string{"t.2000": `execve("/usr/bin/bash", 0x55d0c2a3b4a0, 0x55d0c2a3b4c8) = 0` + "\n"},
			fichero:    "t.2000",
			linea:      1,
			fragmentos: []string{"los argumentos de execve no tienen la forma de la traza"},
		},
		{
			nombre: "ruta-de-execve-con-escape-desconocido",
			hilos: map[string]string{
				"t.2000": `execve("/usr/bin/ba\qsh", ["bash"], 0x7ffd5a1e2c40 /* 25 vars */) = 0` + "\n" + lineaFinalConCero,
			},
			fichero:    "t.2000",
			linea:      1,
			fragmentos: []string{`la ruta de execve: la secuencia de escape \q es desconocida`},
		},
		{
			// strace abrevia así un argv más largo de lo que escribe.
			nombre: "argv-abreviado",
			hilos: map[string]string{
				"t.2000": `execve("/usr/bin/bash", ["bash", "-c", ...], 0x7ffd5a1e2c40 /* 25 vars */) = 0` + "\n" + lineaFinalConCero,
			},
			fichero:    "t.2000",
			linea:      1,
			fragmentos: []string{"el argv de execve no es una lista de cadenas entre comillas"},
		},
		{
			nombre: "argv-sin-separador",
			hilos: map[string]string{
				"t.2000": `execve("/usr/bin/bash", ["bash" "-c"], 0x7ffd5a1e2c40 /* 25 vars */) = 0` + "\n" + lineaFinalConCero,
			},
			fichero:    "t.2000",
			linea:      1,
			fragmentos: []string{"el argv de execve no es una lista de cadenas entre comillas"},
		},
		{
			nombre: "argv-con-hexadecimal-de-una-cifra",
			hilos: map[string]string{
				"t.2000": `execve("` + boeDeLaSkillInstalada + `", ["boe", "buscar", "notificaci\xc3\xb"], ` +
					`0x7ffc3e7a1b88 /* 25 vars */) = 0` + "\n" + lineaFinalConCero,
			},
			fichero:    "t.2000",
			linea:      1,
			fragmentos: []string{`el argv de execve: la secuencia \xb no tiene dos dígitos hexadecimales`},
		},
		{
			nombre:     "connect-con-direccion-sin-leer",
			hilos:      map[string]string{"t.2000": lineaDeExecveDeBoe + "connect(9, 0x7ffc3e7a1b90, 16) = 0\n" + lineaFinalConCero},
			fichero:    "t.2000",
			linea:      2,
			fragmentos: []string{"los argumentos de connect no tienen la forma de la traza"},
		},
		{
			nombre: "direccion-inet-sin-sin-addr",
			hilos: map[string]string{"t.2000": lineaDeExecveDeBoe +
				"connect(9, {sa_family=AF_INET, sin_port=htons(443)}, 16) = " + resultadoEnCurso + "\n" + lineaFinalConCero},
			fichero:    "t.2000",
			linea:      2,
			fragmentos: []string{"la dirección AF_INET de connect: no tiene la forma de la traza"},
		},
		{
			nombre: "direccion-inet-fuera-de-rango",
			hilos: map[string]string{"t.2000": lineaDeExecveDeBoe +
				`connect(9, {sa_family=AF_INET, sin_port=htons(443), sin_addr=inet_addr("203.0.113.256")}, 16) = ` +
				resultadoEnCurso + "\n" + lineaFinalConCero},
			fichero:    "t.2000",
			linea:      2,
			fragmentos: []string{"la dirección AF_INET de connect: no es una dirección y un puerto"},
		},
		{
			// strace escribe así, con @ delante de la ruta, la dirección de un socket
			// sin fichero en el sistema de ficheros.
			nombre: "ruta-unix-abstracta",
			hilos: map[string]string{"t.2000": lineaDeExecveDeBoe +
				`connect(3, {sa_family=AF_UNIX, sun_path=@"/tmp/.X11-unix/X0"}, 20) = 0` + "\n" + lineaFinalConCero},
			fichero:    "t.2000",
			linea:      2,
			fragmentos: []string{"la dirección AF_UNIX de connect: no tiene la forma de la traza"},
		},
		{
			nombre: "ruta-unix-con-octal-fuera-de-octeto",
			hilos: map[string]string{"t.2000": lineaDeExecveDeBoe +
				`connect(3, {sa_family=AF_UNIX, sun_path="/var/run/\400"}, 110) = ` + resultadoSinFichero + "\n" +
				lineaFinalConCero},
			fichero:    "t.2000",
			linea:      2,
			fragmentos: []string{`la dirección AF_UNIX de connect: la secuencia \400 no cabe en un octeto`},
		},
		{
			nombre: "familia-ajena-en-una-invocacion",
			hilos: map[string]string{"t.2000": lineaDeExecveDeBoe +
				"connect(3, {sa_family=AF_NETLINK, nl_pid=0, nl_groups=00000000}, 12) = 0\n" + lineaFinalConCero},
			fichero:    "t.2000",
			linea:      2,
			fragmentos: []string{"connect de una invocación con la familia AF_NETLINK"},
		},
		{
			nombre: "hilo-creado-dos-veces",
			hilos: map[string]string{
				"t.2000": lineaDeExecveDeBash + lineaDeCloneDeBash("2001") + lineaDeCloneDeBash("2001") + lineaFinalConCero,
				"t.2001": lineaFinalConCero,
			},
			fichero:    "t.2001",
			fragmentos: []string{"dos líneas crean", "t.2000, línea 2, y ", "t.2000, línea 3"},
		},
		{
			nombre: "hilos-que-no-descienden-del-raiz",
			hilos: map[string]string{
				"t.2000": lineaDeExecveDeBash + lineaFinalConCero,
				"t.2001": lineaDeCloneDeBash("2002") + lineaFinalConCero,
				"t.2002": lineaDeCloneDeBash("2001") + lineaFinalConCero,
			},
			fichero:    "t.2001",
			fragmentos: []string{"t.2002 tienen la línea que los crea pero no descienden de ", "t.2000"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			dir := escribirTraza(t, caso.hilos, caso.carpetas)

			invocaciones, err := LeerTrazas(dir, false)

			require.Error(t, err)
			assert.Nil(t, invocaciones, "una traza ilegible no devuelve ninguna invocación")

			ruta := filepath.Join(dir, caso.fichero)
			require.ErrorContains(t, err, ruta)

			if caso.linea > 0 {
				texto := strings.Split(caso.hilos[caso.fichero], "\n")[caso.linea-1]
				require.ErrorContains(t, err, fmt.Sprintf("%s, línea %d: ", ruta, caso.linea))
				require.ErrorContains(t, err, "«"+texto+"»")
			}

			for _, fragmento := range caso.fragmentos {
				require.ErrorContains(t, err, fragmento)
			}
		})
	}
}

// TestLeerTrazasSinInvocaciones fija la regla 3 de data-model §9 sobre una traza
// que escribe el propio test: bash crea dos procesos que no son invocaciones, uno
// que no ejecuta nada y deja solo su línea final (research.md V53) y otro que
// ejecuta un programa con el argv vacío, que no invoca ningún applet. La traza no
// es ilegible y no tiene ninguna invocación.
func TestLeerTrazasSinInvocaciones(t *testing.T) {
	t.Parallel()

	dir := escribirTraza(t, map[string]string{
		"t.1000": lineaDeExecveDeBash + lineaDeCloneDeBash("2000") + lineaDeCloneDeBash("2001") + lineaFinalConCero,
		"t.2000": lineaFinalConCero,
		"t.2001": `execve("/usr/bin/env", [], 0x7ffd5a1e2c40 /* 0 vars */) = 0` + "\n" + lineaFinalConCero,
	}, nil)

	invocaciones, err := LeerTrazas(dir, false)

	require.NoError(t, err)
	assert.Empty(t, invocaciones)
}

// banderasDeHiloDeGo son las banderas con las que el runtime de Go crea sus hilos
// en arm64 (research.md V51 y V53); en amd64 añade CLONE_SETTLS.
const banderasDeHiloDeGo = "CLONE_VM|CLONE_FS|CLONE_FILES|CLONE_SIGHAND|CLONE_THREAD|CLONE_SYSVSEM"

// Líneas reales de la llamada que el fin del proceso deja sin terminar, tal como
// las escribió strace 6.8, con sus direcciones y sus números reales (data-model
// §9, regla 5; research.md V65): la del runner de x86_64, línea 1 de t.14465 de
// la sesión 09 de la prueba de red del intento 6 de T030 (ejecución 34961757559,
// gates/prueba-de-red.md §4), la forma A con la marca; y las de la sonda de V65
// en un contenedor de ubuntu:24.04 en arm64 (gates/tarea-T045.md): la clone de
// forma A sin la marca (t.4030), la de forma B (t.20425), la de forma C (t.15079),
// la forma D, la de un hilo que murió en la parada de entrada de una llamada que
// strace no llegó a identificar (t.258), y el connect en curso (t.14). La llamada
// desconocida cerrada con el resultado de la llamada sin terminar y el relleno de
// alineación es la línea 1 de t.19547 de una sesión de la eval 04 en el runner
// (ejecución 35156339496 de H5.1, gates/tarea-T015.md de su hito): 43 caracteres,
// con el igual en la columna 41.
const (
	lineaDeLlamadaDesconocidaCerradaDelRunner = "???()                                   = ?"

	lineaDeCloneSinTerminarDelRunner = "clone(child_stack=0x2a559d472000, flags=" + banderasDeHiloDeGo +
		"|CLONE_SETTLS <unfinished ...>) = ?"
	lineaDeCloneSinTerminarSinMarca = "clone(child_stack=0x2bb2e2418000, flags=" + banderasDeHiloDeGo + ") = ?"
	lineaDeCloneNoDisponible        = "clone(child_stack=0x203cb5a94000, flags=" + banderasDeHiloDeGo + ") = ? <unavailable>"
	lineaDeCloneSinCerrar           = "clone(child_stack=0x666394d64000, flags=" + banderasDeHiloDeGo + " <unfinished ...>"
	lineaDeLlamadaDesconocida       = "???( <unfinished ...>"
	lineaDeConnectSinTerminar       = `connect(9, {sa_family=AF_INET, sin_port=htons(60929), ` +
		`sin_addr=inet_addr("127.0.0.1")}, 16) = ?`
)

// Líneas de las trazas que escribe TestLeerTrazasSinTerminar: la execve con la
// que empieza el fichero raíz de cada traza de la sonda de V65 (la de t.254), que
// no es ningún applet; el connect de una invocación a una dirección pública que
// el fin del proceso deja sin terminar, con la forma real de V65 y la dirección
// de las trazas sintéticas; y el mismo connect con resultado 0.
const (
	lineaDeExecveDeLaSonda           = `execve("/sonda/sonda", ["/sonda/sonda"], 0xffffcbba8c90 /* 4 vars */) = 0` + "\n"
	lineaDeConnectPublicoSinTerminar = `connect(9, {sa_family=AF_INET, sin_port=htons(443), ` +
		`sin_addr=inet_addr("203.0.113.7")}, 16) = ?` + "\n"
	lineaDeConnectPublicoAceptado = `connect(9, {sa_family=AF_INET, sin_port=htons(443), ` +
		`sin_addr=inet_addr("203.0.113.7")}, 16) = 0` + "\n"
)

// TestLeerTrazasSinTerminar fija, sobre trazas que el propio test escribe en
// t.TempDir() con las líneas reales de V65, la lectura de la llamada que el fin
// del proceso deja sin terminar y del fichero del hilo que una creación así pudo
// crear (data-model §9, reglas 1 y 5; FR-076): las trazas enteras de la sonda de
// V65 —la de la clone de forma A con su fichero huérfano y la de forma B con el
// suyo, la de forma C y la de la llamada desconocida—, legibles y sin ninguna
// invocación, porque la sonda no es un applet; el connect en curso de un hilo de
// una invocación, atribuido a ella y de clase red a una dirección pública y local
// al bucle local; y, ilegibles con un error que nombra el fichero, un fichero
// huérfano cuando las únicas llamadas sin terminar de la traza son la desconocida
// o un connect, que no crean ningún hilo, un fichero sin línea de creación con
// alguna llamada junto a un huérfano, y una línea de forma C seguida de otra
// llamada. Las líneas literales de la nota van tal cual, con sus números de hilo;
// las líneas de creación que la nota solo describe llevan las pilas del test.
func TestLeerTrazasSinTerminar(t *testing.T) {
	t.Parallel()

	final := lineaFinalConCero

	// Traza 32269785: la clone de forma A en t.4030, que crea el raíz t.4028 en
	// su línea 3, y el huérfano t.4032, con solo su línea final.
	cloneSinTerminar := map[string]string{
		"t.4028": lineaDeExecveDeLaSonda + lineaDeCloneDeGo("0x2bb2e2410000", 4029) +
			lineaDeCloneDeGo("0x2bb2e2414000", 4030) + lineaDeCloneDeGo("0x2bb2e241c000", 4031) + final,
		"t.4029": final,
		"t.4030": lineaDeCloneSinTerminarSinMarca + "\n" + final,
		"t.4031": final,
		"t.4032": final,
	}

	// Traza 3f18ff32: la clone de forma B en t.20425, línea 2, tras la que crea
	// 20427; el raíz crea 20424, 20425 y 20426, t.20427 crea 20429, y t.20428 es
	// el huérfano.
	cloneNoDisponible := map[string]string{
		"t.20423": lineaDeExecveDeLaSonda + lineaDeCloneDeGo("0x203cb5a80000", 20424) +
			lineaDeCloneDeGo("0x203cb5a84000", 20425) + lineaDeCloneDeGo("0x203cb5a88000", 20426) + final,
		"t.20424": final,
		"t.20425": lineaDeCloneDeGo("0x203cb5a98000", 20427) + lineaDeCloneNoDisponible + "\n" + final,
		"t.20426": final,
		"t.20427": lineaDeCloneDeGo("0x203cb5a9c000", 20429) + final,
		"t.20428": final,
		"t.20429": final,
	}

	// Traza adb8222a: la clone de forma C en t.15079, línea 2, tras la que crea
	// 15081, y en t.15082, que crea el raíz, como línea 1, con la misma pila; sin
	// huérfano. El número del raíz no está en la nota.
	cloneSinCerrar := map[string]string{
		"t.15077": lineaDeExecveDeLaSonda + lineaDeCloneDeGo("0x666394d80000", 15078) +
			lineaDeCloneDeGo("0x666394d84000", 15079) + lineaDeCloneDeGo("0x666394d88000", 15082) + final,
		"t.15078": final,
		"t.15079": lineaDeCloneDeGo("0x666394d98000", 15081) + lineaDeCloneSinCerrar + "\n" + final,
		"t.15081": final,
		"t.15082": lineaDeCloneSinCerrar + "\n" + final,
	}

	// Traza de la primera pasada, repetición 22: la llamada desconocida en
	// t.258, que crea t.256 en su línea 1; el raíz t.254 empieza por la execve de
	// la sonda; sin huérfano.
	llamadaDesconocida := map[string]string{
		"t.254": lineaDeExecveDeLaSonda + lineaDeCloneDeGo("0xc413c510000", 255) +
			lineaDeCloneDeGo("0xc413c514000", 256) + lineaDeCloneDeGo("0xc413c51c000", 257) + final,
		"t.255": final,
		"t.256": lineaDeCloneDeGo("0xc413c518000", 258) + final,
		"t.257": final,
		"t.258": lineaDeLlamadaDesconocida + "\n" + final,
	}

	// La misma traza con la llamada desconocida cerrada, como la escribió el
	// runner.
	llamadaDesconocidaCerrada := maps.Clone(llamadaDesconocida)
	llamadaDesconocidaCerrada["t.258"] = lineaDeLlamadaDesconocidaCerradaDelRunner + "\n" + final

	// Una invocación de la skill instalada cuyo hilo 2001 deja un connect sin
	// terminar.
	invocacionConConnect := func(conexion string) map[string]string {
		return map[string]string{
			"t.2000": lineaDeExecveDeBoe + lineaDeCloneDeGo("0xc000100000", 2001) + final,
			"t.2001": conexion + final,
		}
	}

	huerfanoConLlamadaDesconocida := maps.Clone(llamadaDesconocida)
	huerfanoConLlamadaDesconocida["t.259"] = final

	huerfanoConConnect := invocacionConConnect(lineaDeConnectPublicoSinTerminar)
	huerfanoConConnect["t.2002"] = final

	sinOrigenJuntoAlHuerfano := maps.Clone(cloneSinTerminar)
	sinOrigenJuntoAlHuerfano["t.4033"] = lineaDeConnectPublicoAceptado + final

	sinCerrarSeguidaDeOtraLlamada := maps.Clone(cloneSinCerrar)
	sinCerrarSeguidaDeOtraLlamada["t.15079"] = lineaDeCloneDeGo("0x666394d98000", 15081) + lineaDeCloneSinCerrar + "\n" +
		lineaDeConnectPublicoAceptado + final

	a21 := []string{normaDeLasTrazas, "a21"}
	json := []string{"--json"}

	casos := []struct {
		nombre       string
		hilos        map[string]string
		invocaciones []Invocacion
		defecto      *defectoEsperado
	}{
		{nombre: "clone-sin-terminar-con-huerfano", hilos: cloneSinTerminar},
		{nombre: "clone-no-disponible-con-huerfano", hilos: cloneNoDisponible},
		{nombre: "clone-sin-cerrar", hilos: cloneSinCerrar},
		{nombre: "llamada-desconocida", hilos: llamadaDesconocida},
		{nombre: "llamada-desconocida-cerrada", hilos: llamadaDesconocidaCerrada},
		{
			nombre: "connect-publico-sin-terminar",
			hilos:  invocacionConConnect(lineaDeConnectPublicoSinTerminar),
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(0), "articulo", a21, json,
				conexionInet(destinoPublico, resultadoSinTerminar, ConexionRed))},
		},
		{
			nombre: "connect-local-sin-terminar",
			hilos:  invocacionConConnect(lineaDeConnectSinTerminar + "\n"),
			invocaciones: []Invocacion{invocacionDeBoe(2000, codigoDeSalida(0), "articulo", a21, json,
				conexionInet("127.0.0.1:60929", resultadoSinTerminar, ConexionLocal))},
		},
		{
			nombre:  "huerfano-solo-con-llamada-desconocida",
			hilos:   huerfanoConLlamadaDesconocida,
			defecto: &defectoEsperado{ficheros: []string{"t.259"}, motivo: "quedó sin terminar", ajenos: []string{"t.258"}},
		},
		{
			nombre:  "huerfano-solo-con-connect-sin-terminar",
			hilos:   huerfanoConConnect,
			defecto: &defectoEsperado{ficheros: []string{"t.2002"}, motivo: "quedó sin terminar", ajenos: []string{"t.2001"}},
		},
		{
			nombre:  "fichero-con-llamadas-sin-origen-junto-a-un-huerfano",
			hilos:   sinOrigenJuntoAlHuerfano,
			defecto: &defectoEsperado{ficheros: []string{"t.4033"}, motivo: "solo puede faltarle a uno", ajenos: []string{"t.4032"}},
		},
		{
			nombre:  "sin-cerrar-seguida-de-otra-llamada",
			hilos:   sinCerrarSeguidaDeOtraLlamada,
			defecto: &defectoEsperado{ficheros: []string{"t.15079"}, linea: 2, motivo: "solo pueden seguirla"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			comprobarLectura(t, escribirTraza(t, caso.hilos, nil), false, caso.invocaciones, caso.defecto)
		})
	}
}

// TestLeerTrazasHiloCreadoAntesDeLaEjecucion fija, sobre una traza que el propio
// test escribe en t.TempDir(), la regla 3 de data-model §9 con dos hilos del
// mismo proceso que conectan: bash crea el primero antes de reemplazarse por el
// applet con execve, y el applet crea el segundo después. Solo el connect del
// segundo es de la invocación: el primer hilo es del proceso, pero no lo creó
// ninguna línea posterior a la execve del applet (FR-076).
func TestLeerTrazasHiloCreadoAntesDeLaEjecucion(t *testing.T) {
	t.Parallel()

	const connectDelHiloDeBash = `connect(9, {sa_family=AF_INET, sin_port=htons(443), ` +
		`sin_addr=inet_addr("203.0.113.8")}, 16) = -1 EINPROGRESS (Operation now in progress)` + "\n"

	hilos := map[string]string{
		"t.2000": lineaDeExecveDeBash + lineaDeCloneDeGo("0xc000100000", 2001) + lineaDeExecveDeBoe +
			lineaDeCloneDeGo("0xc000104000", 2002) + lineaFinalConCero,
		"t.2001": connectDelHiloDeBash + lineaFinalConCero,
		"t.2002": lineaDeConnectPublicoAceptado + lineaFinalConCero,
	}

	comprobarLectura(t, escribirTraza(t, hilos, nil), false, []Invocacion{
		invocacionDeBoe(2000, codigoDeSalida(0), "articulo", []string{normaDeLasTrazas, "a21"}, []string{"--json"},
			conexionInet(destinoPublico, "0", ConexionRed)),
	}, nil)
}

// TestLeerTrazasAyudaSeguidaDeConsulta fija, sobre una traza que el propio test
// escribe en t.TempDir(), que el intérprete con el que LeerTrazas lee todas las
// invocaciones de una sesión decide la consulta de cada una solo con sus propias
// banderas (data-model §9, campo consulta; FR-072): claude crea con vfork el
// proceso 2000, que pide la ayuda, y después el 3000, que consulta el mismo
// bloque. La ayuda del primero no deja sin consulta al segundo, que es el que
// lee el bloque esperado.
func TestLeerTrazasAyudaSeguidaDeConsulta(t *testing.T) {
	t.Parallel()

	const (
		lineaDeExecveDeClaude = `execve("/usr/local/bin/claude", ["claude", "-p", "art. 21 de la Ley 39/2015"], ` +
			`0x7ffe2a4c8d10 /* 31 vars */) = 0` + "\n"
		lineaDeExecveDeBoeConAyuda = `execve("` + boeDeLaSkillInstalada + `", ["` + boeDeLaSkillInstalada + `", ` +
			`"articulo", "BOE-A-2015-10565", "a21", "--help"], 0x7ffc3e7a1b88 /* 25 vars */) = 0` + "\n"
	)

	// vforkDeClaude es la línea del runner con el número de proceso dado: el
	// igual sigue en la columna 41.
	vforkDeClaude := func(proceso int) string {
		return strings.TrimSuffix(lineaDeVforkDelRunner, "11494") + strconv.Itoa(proceso) + "\n"
	}

	hilos := map[string]string{
		"t.1000": lineaDeExecveDeClaude + vforkDeClaude(2000) + vforkDeClaude(3000) + lineaFinalConCero,
		"t.2000": lineaDeExecveDeBoeConAyuda + lineaFinalConCero,
		"t.3000": lineaDeExecveDeBoe + lineaFinalConCero,
	}

	a21 := []string{normaDeLasTrazas, "a21"}

	ayuda := invocacionDeBoe(2000, codigoDeSalida(0), "articulo", a21, []string{"--help"})
	ayuda.Consulta = false

	comprobarLectura(t, escribirTraza(t, hilos, nil), false, []Invocacion{
		ayuda,
		invocacionDeBoe(3000, codigoDeSalida(0), "articulo", a21, []string{"--json"}),
	}, nil)
}

// lineaDeCloneDeGo es la línea con la que un hilo de Go de arm64 crea el hilo de
// ese número (research.md V53), con la pila dada y su salto de línea.
func lineaDeCloneDeGo(pila string, hilo int) string {
	return "clone(child_stack=" + pila + ", flags=" + banderasDeHiloDeGo + ") = " + strconv.Itoa(hilo) + "\n"
}

// lineaDeVforkDelRunner es la línea con la que Claude Code 2.1.270 de x86_64 crea
// los procesos de sus órdenes, tal como la escribió strace 6.8 en el runner en la
// prueba de red del intento 3 de T030 (ejecución 34936425178): vfork(), el espacio
// tras el paréntesis de cierre, el relleno hasta la columna de alineación, la 40
// (-a 40, el valor por defecto), y el igual con el resultado desde la columna 41.
const lineaDeVforkDelRunner = "vfork()                                 = 11494"

// TestLeerLlamadaConRelleno fija, sobre líneas literales, que entre el paréntesis
// de cierre de una llamada y el igual caben el espacio y el relleno de alineación
// de strace (data-model §9, regla 5): la línea real del runner y la misma llamada
// con un solo espacio se leen igual, como una vfork con su resultado que crea el
// proceso de ese número (regla 2). El relleno no abre la forma a nada más: sin
// resultado, sin el espacio o con otro blanco, la línea sigue siendo ilegible.
func TestLeerLlamadaConRelleno(t *testing.T) {
	t.Parallel()

	require.Len(t, lineaDeVforkDelRunner, 47)
	require.Equal(t, strings.Repeat(" ", 33),
		strings.TrimSuffix(strings.TrimPrefix(lineaDeVforkDelRunner, "vfork()"), "= 11494"),
		"entre vfork() y el igual y su resultado van el espacio y el relleno: 33 espacios")
	require.Equal(t, 41, strings.Index(lineaDeVforkDelRunner, "=")+1, "el igual va en la columna 41")

	legibles := []struct{ nombre, texto string }{
		{nombre: "relleno-del-runner", texto: lineaDeVforkDelRunner},
		{nombre: "un-espacio", texto: "vfork() = 11494"},
	}

	for _, legible := range legibles {
		t.Run(legible.nombre, func(t *testing.T) {
			t.Parallel()

			leida, err := leerLlamada(legible.texto)

			require.NoError(t, err)
			assert.Equal(t, "vfork", leida.nombre)
			assert.Equal(t, "11494", leida.resultado)
			assert.False(t, leida.sinResultado)
			assert.Equal(t, 11494, leida.valor)
			assert.True(t, leida.creaHilo(), "una vfork con resultado crea el proceso de ese número")
		})
	}

	ilegibles := []struct{ nombre, texto string }{
		{nombre: "relleno-sin-igual", texto: strings.TrimSuffix(lineaDeVforkDelRunner, "= 11494")},
		{nombre: "relleno-sin-resultado", texto: strings.TrimSuffix(lineaDeVforkDelRunner, "11494")},
		{nombre: "sin-espacio", texto: "vfork()= 11494"},
		{nombre: "tabulador", texto: "vfork()\t= 11494"},
	}

	for _, ilegible := range ilegibles {
		t.Run(ilegible.nombre, func(t *testing.T) {
			t.Parallel()

			_, err := leerLlamada(ilegible.texto)

			require.ErrorContains(t, err, "no es ninguna de las formas de línea de la traza")
		})
	}
}

// TestLeerLlamadaSinTerminar fija, sobre líneas literales, las cuatro formas de la
// llamada que el fin del proceso deja sin terminar (data-model §9, regla 5;
// research.md V65): la línea del runner y cada línea real de la sonda se leen
// como una llamada sin terminar de su nombre, sin resultado —con el texto tras
// «= » como resultado, vacío en la forma sin cerrar—, que no crea ningún hilo, y
// con su dirección en el connect. Las formas no se abren a nada más: la marca con
// cualquier otro resultado, la marca dentro de los argumentos, la llamada
// desconocida con argumentos o con resultado y la forma sin cerrar de una llamada
// que no es del filtro siguen siendo ilegibles.
func TestLeerLlamadaSinTerminar(t *testing.T) {
	t.Parallel()

	legibles := []struct {
		nombre, texto string

		// llamada es el nombre de la llamada; resultado, el texto tras «= »; y
		// destino, la dirección y el puerto del connect.
		llamada, resultado, destino string
	}{
		{nombre: "clone-del-runner-con-la-marca", texto: lineaDeCloneSinTerminarDelRunner, llamada: "clone", resultado: "?"},
		{nombre: "clone-sin-la-marca", texto: lineaDeCloneSinTerminarSinMarca, llamada: "clone", resultado: "?"},
		{nombre: "clone-no-disponible", texto: lineaDeCloneNoDisponible, llamada: "clone", resultado: "? <unavailable>"},
		{nombre: "clone-sin-cerrar", texto: lineaDeCloneSinCerrar, llamada: "clone"},
		{nombre: "llamada-desconocida", texto: lineaDeLlamadaDesconocida, llamada: "???"},
		{
			nombre:    "llamada-desconocida-cerrada-del-runner",
			texto:     lineaDeLlamadaDesconocidaCerradaDelRunner,
			llamada:   "???",
			resultado: "?",
		},
		{nombre: "llamada-desconocida-cerrada-con-un-espacio", texto: "???() = ?", llamada: "???", resultado: "?"},
		{
			nombre:    "connect",
			texto:     lineaDeConnectSinTerminar,
			llamada:   "connect",
			resultado: "?",
			destino:   "127.0.0.1:60929",
		},
	}

	for _, legible := range legibles {
		t.Run(legible.nombre, func(t *testing.T) {
			t.Parallel()

			leida, err := leerLlamada(legible.texto)

			require.NoError(t, err)
			assert.Equal(t, legible.llamada, leida.nombre)
			assert.Equal(t, legible.resultado, leida.resultado)
			assert.True(t, leida.sinResultado, "la llamada queda sin resultado")
			assert.True(t, leida.sinTerminar, "el fin del proceso la dejó sin terminar")
			assert.False(t, leida.conValor, "no tiene resultado numérico")
			assert.False(t, leida.creaHilo(), "una creación sin terminar no crea ningún hilo con número")
			assert.Nil(t, leida.argv, "no es una execve con argv")

			if legible.destino == "" {
				return
			}

			assert.Equal(t, "AF_INET", leida.conexion.Familia)
			assert.Equal(t, legible.destino, leida.conexion.Direccion.String())
			assert.Equal(t, legible.resultado, leida.conexion.Resultado)
			assert.True(t, leida.conexion.SinResultado, "la conexión queda sin resultado")
		})
	}

	conMarca := strings.TrimSuffix(lineaDeCloneSinTerminarDelRunner, ") = ?")
	otraForma := "no es ninguna de las formas de línea de la traza"

	ilegibles := []struct{ nombre, texto, motivo string }{
		{nombre: "marca-con-resultado", texto: conMarca + ") = 2001", motivo: "solo cabe con el resultado ?"},
		{
			nombre: "marca-con-error",
			texto:  conMarca + ") = -1 EAGAIN (Resource temporarily unavailable)",
			motivo: "solo cabe con el resultado ?",
		},
		{
			nombre: "marca-con-llamada-interrumpida",
			texto:  conMarca + ") = " + resultadoInterrumpida,
			motivo: "solo cabe con el resultado ?",
		},
		{nombre: "marca-con-no-disponible", texto: conMarca + ") = ? <unavailable>", motivo: "solo cabe con el resultado ?"},
		{
			nombre: "marca-dentro-de-los-argumentos",
			texto:  "clone(child_stack=0x2bb2e2418000 <unfinished ...>, flags=" + banderasDeHiloDeGo + ") = ?",
			motivo: "solo va delante del paréntesis de cierre",
		},
		{
			nombre: "desconocida-con-argumentos",
			texto:  "???(child_stack=0x2bb2e2418000 <unfinished ...>",
			motivo: "no lleva argumentos",
		},
		{nombre: "desconocida-con-resultado-numerico", texto: "???() = 0", motivo: "solo lleva el resultado ?"},
		{
			nombre: "desconocida-con-error",
			texto:  "???() = -1 ENOSYS (Function not implemented)",
			motivo: "solo lleva el resultado ?",
		},
		{nombre: "desconocida-no-disponible", texto: "???() = ? <unavailable>", motivo: "solo lleva el resultado ?"},
		{nombre: "desconocida-cerrada-con-argumentos", texto: "???(0x1) = ?", motivo: "no lleva argumentos"},
		{
			nombre: "desconocida-cerrada-con-la-marca",
			texto:  "???( <unfinished ...>) = ?",
			motivo: "no lleva argumentos",
		},
		{nombre: "desconocida-cerrada-sin-espacio", texto: "???()= ?", motivo: otraForma},
		{
			nombre: "sin-cerrar-fuera-del-filtro",
			texto:  "futex(0xc000100148, FUTEX_WAIT_PRIVATE, 0, NULL <unfinished ...>",
			motivo: otraForma,
		},
		{
			// Los argumentos de la forma sin cerrar se leen igual que los de una
			// llamada con resultado: strace escribe así la dirección que no pudo
			// leer de la memoria del proceso.
			nombre: "sin-cerrar-con-direccion-sin-leer",
			texto:  "connect(9, 0x7ffc3e7a1b90, 16 <unfinished ...>",
			motivo: "los argumentos de connect no tienen la forma de la traza",
		},
	}

	for _, ilegible := range ilegibles {
		t.Run(ilegible.nombre, func(t *testing.T) {
			t.Parallel()

			_, err := leerLlamada(ilegible.texto)

			require.ErrorContains(t, err, ilegible.motivo)
		})
	}
}

// TestDecodificarCadena fija cómo se decodifica una cadena tal como strace la
// escribe entre comillas (research.md V53): las secuencias de una letra, el octal
// de una a tres cifras —tres cuando detrás va otra cifra— y el hexadecimal de dos;
// y el error de cada secuencia que no es ninguna de ellas. La barra invertida
// final no llega desde un fichero, porque la forma de la cadena en la traza no la
// admite, pero la función no puede leer más allá del final de lo que recibe.
func TestDecodificarCadena(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre       string
		escrita      string
		decodificada string

		// motivo es el error; vacío, sin error.
		motivo string
	}{
		{nombre: "sin-escapes", escrita: "/usr/bin/bash", decodificada: "/usr/bin/bash"},
		{nombre: "de-una-letra", escrita: `\"\\\f\n\r\t\v`, decodificada: "\"\\\f\n\r\t\v"},
		{nombre: "octal-de-una-dos-y-tres-cifras", escrita: `\0\12c\303\263digo`, decodificada: "\x00\ncódigo"},
		{nombre: "octal-seguido-de-otra-cifra", escrita: `\0001`, decodificada: "\x001"},
		{nombre: "octal-del-mayor-octeto", escrita: `\377`, decodificada: "\xff"},
		{nombre: "hexadecimal", escrita: `notificaci\xc3\xb3n`, decodificada: "notificación"},
		{nombre: "barra-invertida-final", escrita: `c\303\263digo\`, motivo: "la cadena termina en una barra invertida"},
		{nombre: "hexadecimal-de-una-cifra", escrita: `\x4`, motivo: `la secuencia \x4 no tiene dos dígitos hexadecimales`},
		{
			nombre:  "hexadecimal-con-otra-letra",
			escrita: `\xg1`,
			motivo:  `la secuencia \xg1 no tiene dos dígitos hexadecimales: encoding/hex: invalid byte`,
		},
		{nombre: "escape-desconocido", escrita: `\q`, motivo: `la secuencia de escape \q es desconocida`},
		{nombre: "octal-fuera-de-octeto", escrita: `\400`, motivo: `la secuencia \400 no cabe en un octeto`},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			decodificada, err := decodificarCadena(caso.escrita)

			if caso.motivo == "" {
				require.NoError(t, err)
				assert.Equal(t, caso.decodificada, decodificada)

				return
			}

			require.ErrorContains(t, err, caso.motivo)
			assert.Empty(t, decodificada)
		})
	}
}

// lineaDeCloneDeBash es la línea con la que bash crea un proceso (research.md
// V53), con el resultado tal como se escribe y su salto de línea.
func lineaDeCloneDeBash(resultado string) string {
	return "clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID|CLONE_CHILD_SETTID|SIGCHLD, " +
		"child_tidptr=0x7f3a9c2e5a10) = " + resultado + "\n"
}

// escribirTraza crea en un directorio temporal del test la traza de una sesión:
// un fichero por cada entrada de hilos, con su contenido, y una carpeta por cada
// nombre de carpetas. Devuelve su ruta.
func escribirTraza(t *testing.T, hilos map[string]string, carpetas []string) string {
	t.Helper()

	dir := t.TempDir()

	for nombre, contenido := range hilos {
		require.NoError(t, os.WriteFile(filepath.Join(dir, nombre), []byte(contenido), 0o600))
	}

	for _, carpeta := range carpetas {
		require.NoError(t, os.Mkdir(filepath.Join(dir, carpeta), 0o750))
	}

	return dir
}
