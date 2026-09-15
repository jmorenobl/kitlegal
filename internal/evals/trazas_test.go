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
// es el principal; el argv decodificado; el código de salida, la muerte por señal
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
