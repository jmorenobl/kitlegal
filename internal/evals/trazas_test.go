package evals

import (
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
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

// Resultados de connect tal como los escribe strace (research.md V53 y V54).
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
// líneas de señal, que no cuentan; y la traza ilegible, con el fichero, la línea
// y su texto, en lugar de ignorar lo que no se entiende (contrato job-de-evals
// §9; FR-072, FR-076).
func TestLeerTrazas(t *testing.T) {
	t.Parallel()

	a21 := []string{normaDeLasTrazas, "a21"}
	a9998 := []string{normaDeLasTrazas, "a9998"}
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
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			dir := filepath.Join(casosDeLeerTrazas, caso.nombre)

			invocaciones, err := LeerTrazas(dir, caso.cortada)

			if caso.defecto == nil {
				require.NoError(t, err)
				assert.Equal(t, caso.invocaciones, invocaciones)

				return
			}

			require.Error(t, err)
			assert.Nil(t, invocaciones, "una traza ilegible no devuelve ninguna invocación")

			for _, fichero := range caso.defecto.ficheros {
				require.ErrorContains(t, err, filepath.Join(dir, fichero))
			}

			if caso.defecto.linea > 0 {
				require.ErrorContains(t, err, fmt.Sprintf("línea %d", caso.defecto.linea))
				require.ErrorContains(t, err, lineaDeLaTraza(t, filepath.Join(dir, caso.defecto.ficheros[0]), caso.defecto.linea))
			}

			require.ErrorContains(t, err, caso.defecto.motivo)

			for _, ajeno := range caso.defecto.ajenos {
				assert.NotContains(t, err.Error(), ajeno)
			}
		})
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
// por el nombre de invocación multicall, el verbo y los argumentos sin las
// banderas globales, con su valor en --timeout y --asunto, y sin consulta con
// --describe o --dry-run; lo que no invoca un applet registrado se ignora.
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
// resultado es el de una llamada que el corte interrumpió.
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
		SinResultado: strings.HasPrefix(resultado, "? "),
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
