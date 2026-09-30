//go:build evals

package evals

import (
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// directorioDeEvalsDeLasSkills es el de las evals de las skills en la raíz del
// repositorio, relativo al directorio de este paquete, que es donde go test
// ejecuta los tests (research.md V46): las de cada skill están en su
// subdirectorio.
const directorioDeEvalsDeLasSkills = "../../evals"

// Banderas con las que scripts/evals.sh invoca, tras -args, la ejecución del job
// (contracts/ejecucion-del-job.md §1 y §5 de H7.3). -skill, -repeticiones y
// -concurrencia son también del sondeo, que recibe las dos últimas tal como las
// escribe quien lo lanza, sin comprobar, y la concurrencia vacía para la del
// job: por eso son texto. Con una bandera entera, un valor que no lo es haría
// fallar go test antes del sondeo, sin el error que nombra su argumento de make
// (FR-067 de H7.3); el job las convierte, porque scripts/evals.sh ya las ha
// comprobado.
var (
	banderaSkill        = flag.String("skill", "", "skill evaluada")
	banderaQueDecide    = flag.String("modelo-que-decide", "", "modelo del job cuyas sesiones deciden el veredicto")
	banderaInformativos = flag.String("modelos-informativos", "",
		"modelos informativos del job, separados por comas; vacío si no hay ninguno")
	banderaRepeticiones = flag.String("repeticiones", "", "sesiones que se abren de cada eval con cada modelo")
	banderaUmbral       = flag.Int("umbral", 0, "sesiones de una serie que tienen que pasar para que la serie pase")
	banderaConcurrencia = flag.String("concurrencia", "",
		"sesiones que se abren a la vez como mucho; en el sondeo, vacía es la del job para la skill")
	banderaPruebaDeRed = flag.Bool("prueba-de-red", false,
		"añade la sesión de la primera eval con el texto de la prueba de red")
	banderaObjetivo = flag.Int("objetivo-de-duracion", 0,
		"segundos que el job admite para sus sesiones; 0, sin objetivo")
	banderaSkills    = flag.String("skills", "", "directorio de las skills instaladas, el que deja make install")
	banderaCommit    = flag.String("commit", "", "commit evaluado")
	banderaSinPython = flag.String("sin-python", "", "ruta de sin-python.txt")
	banderaSesiones  = flag.String("sesiones", "", "directorio en el que se crea el de cada sesión")
	banderaInforme   = flag.String("informe", "", "directorio en el que se escriben informe.md e informe.json")
)

// Banderas con las que scripts/evals-sondeo.sh invoca, tras -args, el sondeo,
// además de -skill, -repeticiones y -concurrencia (contracts/sondeo.md §2 y §3
// de H7.3; data-model §8 de H7.3): las evals y el modelo, tal como los escribe
// quien lo lanza, y el temporal que crea el guion.
var (
	banderaEvals    = flag.String("evals", "", "números de las evals del sondeo, de dos cifras y separados por comas")
	banderaModelo   = flag.String("modelo", "", "modelo con el que se abren las sesiones del sondeo")
	banderaTemporal = flag.String("temporal", "", "directorio del sondeo, que crea y borra su guion")
)

// Banderas con las que el quickstart invoca, tras -args, la comprobación de la
// consulta repetida (contracts/comprobacion-del-quickstart.md §1).
var (
	banderaPrimera       = flag.String("primera", "", "directorio de la primera conversación de la consulta repetida")
	banderaSegunda       = flag.String("segunda", "", "directorio de la segunda conversación de la consulta repetida")
	banderaFechaSuperada = flag.String("fecha-superada", "",
		"fecha de vigencia de la redacción superada que la primera respuesta lleva en la línea de la forma")
	banderaFechaLeida = flag.String("fecha-leida", "",
		"fecha de vigencia de la redacción leída que la primera respuesta lleva en la línea de la forma")
)

// TestEjecucionDelJob es la ejecución del job de evals de una skill, entera y en
// una sola orden (contracts/ejecucion-del-job.md §1 y §5 de H7.3; research.md
// D19 de H7.3):
//
//  1. lee las evals de la skill de la raíz del repositorio y compone el plan de
//     sesiones con los modelos, las repeticiones y la prueba de red de sus
//     banderas. Falla si el directorio tiene algún fichero mal formado: el guion
//     ya lo ha comprobado antes, y planificar sobre un conjunto incompleto
//     abriría menos sesiones sin decirlo;
//  2. reparte las sesiones del plan con ejecutarSesiones, como mucho
//     -concurrencia a la vez, bajo strace, con scripts/evals-sesion.sh, el
//     entorno del job debajo del de cada sesión, las skills instaladas de
//     -skills y el tope de 240 s con su margen de 10 s. SIGINT y SIGTERM cierran
//     las abiertas con la secuencia del tope, y el test falla con el error que
//     nombra cada una, sin escribir el informe (FR-037 de H7.3);
//  3. escribe informe.md e informe.json con EscribirInforme: las mismas evals,
//     las sesiones, las que el repartidor no abrió tras el mensaje del límite de
//     uso y la duración que midió, en segundos redondeados hacia arriba, frente
//     al objetivo de -objetivo-de-duracion (FR-044, FR-050 y FR-051 de H7.3).
//
// Falla con un error o con el veredicto fallo, nombrando sus motivos: así el job
// sale en rojo por una serie que no pasa, por un umbral que decide y no se
// cumple, por la duración o por sesiones sin medir (FR-043 de H7.3). Solo lo
// ejecuta scripts/evals.sh, porque abre sesiones con modelo; lo que decide lo
// fijan TestPlan, los tests del repartidor, TestInforme y TestUmbralesDelInforme.
func TestEjecucionDelJob(t *testing.T) {
	t.Parallel()

	exigirBanderas(t, "skill", "modelo-que-decide", "repeticiones", "umbral", "concurrencia", "skills", "commit",
		"sin-python", "sesiones", "informe")

	repeticiones := enteroDeLaBandera(t, "repeticiones")
	concurrencia := enteroDeLaBandera(t, "concurrencia")

	evals := filepath.Join(directorioDeEvalsDeLasSkills, *banderaSkill)

	conjunto, err := LeerConjunto(evals)
	require.NoError(t, err)
	require.Empty(t, conjunto.MalFormados, "el directorio de evals no tiene ficheros mal formados")

	plan := PlanDeEvals{
		Evals:               conjunto.Evals,
		ModeloQueDecide:     *banderaQueDecide,
		ModelosInformativos: separarLosModelos(*banderaInformativos),
		Repeticiones:        repeticiones,
		PruebaDeRed:         *banderaPruebaDeRed,
	}
	require.NoError(t, plan.Comprobar())

	// El repartidor ejecuta el guion en el directorio de trabajo de cada sesión:
	// su ruta relativa a este paquete no le serviría.
	guion, err := filepath.Abs(guionDeLaSesion)
	require.NoError(t, err)

	senales, dejarDeEscuchar := signal.NotifyContext(t.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer dejarDeEscuchar()

	ejecucion, err := ejecutarSesiones(senales.Done(), SesionesAEjecutar{
		Plan:          plan.Sesiones(),
		Concurrencia:  concurrencia,
		Evals:         evals,
		Sesiones:      *banderaSesiones,
		Skills:        *banderaSkills,
		Guion:         guion,
		Entorno:       os.Environ(),
		Traza:         true,
		Tope:          topeDeUnaSesion,
		MargenDelTope: margenDelTopeDeUnaSesion,
	})
	require.NoError(t, err)

	informe, err := EscribirInforme(InformeAEscribir{
		Skill:                 *banderaSkill,
		Evals:                 evals,
		Sesiones:              *banderaSesiones,
		Destino:               *banderaInforme,
		ModeloQueDecide:       *banderaQueDecide,
		ModelosInformativos:   plan.ModelosInformativos,
		Repeticiones:          repeticiones,
		Umbral:                *banderaUmbral,
		Commit:                *banderaCommit,
		SinPython:             *banderaSinPython,
		SinAbrir:              ejecucion.SinAbrir,
		DuracionDeLasSesiones: segundosHaciaArriba(ejecucion.Duracion),
		ObjetivoDeDuracion:    *banderaObjetivo,
	})
	require.NoError(t, err)
	require.Equalf(t, VeredictoAprobado, informe.Veredicto, "motivos del veredicto:\n%s", strings.Join(informe.Motivos, "\n"))
}

// segundosHaciaArriba son los segundos enteros de una duración, redondeados
// hacia arriba: 900,4 s son 901 y no cumplen un objetivo de 900 (research.md D14
// de H7.3).
func segundosHaciaArriba(duracion time.Duration) int {
	return int((duracion + time.Second - 1) / time.Second)
}

// enteroDeLaBandera es el entero del valor de la bandera de la ejecución del
// job, que scripts/evals.sh ha comprobado antes de invocarla.
func enteroDeLaBandera(t *testing.T, nombre string) int {
	t.Helper()

	entero, err := strconv.Atoi(flag.Lookup(nombre).Value.String())
	require.NoErrorf(t, err, "la bandera -%s es un entero", nombre)

	return entero
}

// TestSondeo es el sondeo local de unas evals de una skill, entero y en una sola
// orden (contracts/sondeo.md §2, §3 y §6 de H7.3; research.md D16 de H7.3):
// con sondear,
//
//  1. comprueba, antes de construir nada, los argumentos de -skill, -evals,
//     -modelo, -repeticiones y -concurrencia, con los errores que nombran su
//     argumento de make, y CLAUDE_CODE_OAUTH_TOKEN (FR-063 y FR-067 de H7.3);
//  2. construye el binario del árbol de trabajo e instala sus skills en el
//     temporal de -temporal, con prepararElArbol (FR-062 de H7.3);
//  3. abre las sesiones de las evals pedidas con el repartidor, sin traza, con
//     scripts/evals-sesion.sh y el entorno de quien lo lanza que ven las
//     sesiones del sondeo. SIGINT y SIGTERM cierran las abiertas con la
//     secuencia del tope, y el test falla con el error que nombra cada una;
//  4. las juzga con el código del job, sin lo que el job lee de la traza.
//
// Escribe la salida en salida.txt del temporal, que su guion imprime, y ningún
// informe ni veredicto: falla solo con un error, sean cuales sean las tasas
// (FR-066 de H7.3). Sin -temporal falla antes de nada, porque el sondeo
// escribiría en el directorio de este paquete. Solo lo ejecuta
// scripts/evals-sondeo.sh, porque abre sesiones con modelo; lo que decide lo
// fijan TestComprobarElSondeo, TestSondear, TestJuicioDelSondeo y
// TestSalidaDelSondeo, y la orden que lo ejecuta, TestGuionDelSondeo.
func TestSondeo(t *testing.T) {
	t.Parallel()

	exigirBanderas(t, "temporal")

	// El repartidor ejecuta el guion en el directorio de trabajo de cada sesión:
	// su ruta relativa a este paquete no le serviría.
	guion, err := filepath.Abs(guionDeLaSesion)
	require.NoError(t, err)

	senales, dejarDeEscuchar := signal.NotifyContext(t.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer dejarDeEscuchar()

	salida, err := sondear(senales.Done(), SondeoAEjecutar{
		Argumentos: ArgumentosDelSondeo{
			Skill:        *banderaSkill,
			Evals:        *banderaEvals,
			Modelo:       *banderaModelo,
			Repeticiones: *banderaRepeticiones,
			Concurrencia: *banderaConcurrencia,
		},
		Entorno:          os.Environ(),
		EvalsDeLasSkills: directorioDeEvalsDeLasSkills,
		Temporal:         *banderaTemporal,
		Guion:            guion,
		PrepararElArbol:  prepararElArbol,
	})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(*banderaTemporal, ficheroDeLaSalidaDelSondeo), []byte(salida), 0o600))
}

// TestComprobarConsultaRepetida comprueba las dos respuestas de la consulta
// repetida del quickstart §6 con comprobarConsultaRepetida: la lista de
// expresiones prohibidas de las evals de la skill de -skill y las conversaciones
// de -primera y -segunda, leídas con LeerSesion como las lee el job, y las fechas
// de -fecha-superada y -fecha-leida. Falla con las conversaciones que no se
// pueden leer o con una línea por condición que falla, una por renglón; si no
// falla ninguna, lo registra (contracts/comprobacion-del-quickstart.md; FR-061).
// Solo lo ejecuta la persona en el quickstart, porque necesita las dos
// conversaciones con modelo (FR-062); lo que decide lo fija
// TestCondicionesDeLaConsultaRepetida.
func TestComprobarConsultaRepetida(t *testing.T) {
	t.Parallel()

	exigirBanderas(t, "skill", "primera", "segunda", "fecha-superada", "fecha-leida")

	evals := filepath.Join(directorioDeEvalsDeLasSkills, *banderaSkill)

	conjunto, err := LeerConjunto(evals)
	require.NoError(t, err)

	// Con la lista mal formada, el conjunto la deja vacía y ninguna respuesta
	// llevaría expresiones prohibidas sin haberlas mirado.
	malFormados := make([]string, 0, len(conjunto.MalFormados))
	for _, malFormado := range conjunto.MalFormados {
		malFormados = append(malFormados, malFormado.Error.Error())
	}

	require.Emptyf(t, malFormados, "el directorio de evals %s tiene ficheros mal formados:\n%s", evals,
		strings.Join(malFormados, "\n"))

	primera := leerConversacion(t, "primera", *banderaPrimera)
	segunda := leerConversacion(t, "segunda", *banderaSegunda)

	if t.Failed() {
		t.FailNow()
	}

	if lineas := comprobarConsultaRepetida(primera, segunda, conjunto.Prohibidas, *banderaFechaSuperada,
		*banderaFechaLeida); len(lineas) > 0 {
		t.Fatal(strings.Join(lineas, "\n"))
	}

	t.Logf("se cumplen las tres condiciones: la forma con %s y %s en la primera respuesta, sin ella en la segunda, "+
		"y ninguna expresión prohibida en las dos", *banderaFechaSuperada, *banderaFechaLeida)
}

// leerConversacion lee con LeerSesion la conversación del directorio dado. Si
// no se puede leer, lo anota nombrándola por su ordinal y con el error, y deja
// seguir para que se lea también la otra.
func leerConversacion(t *testing.T, ordinal, dir string) Sesion {
	t.Helper()

	sesion, err := LeerSesion(dir)
	if err != nil {
		t.Errorf("la %s conversación no se ha podido leer: %v", ordinal, err)
	}

	return sesion
}

// exigirBanderas falla, nombrando cada una, si alguna de las banderas no tiene
// valor: con una ruta vacía, PrepararSesion o EscribirInforme leerían y
// escribirían en el directorio de este paquete en lugar de en el de la
// ejecución. Las mira todas antes de fallar, para que quien lanza la orden sin
// ellas sepa de una vez las que faltan.
func exigirBanderas(t *testing.T, nombres ...string) {
	t.Helper()

	faltan := false

	for _, nombre := range nombres {
		if !assert.NotEmptyf(t, flag.Lookup(nombre).Value.String(), "falta la bandera -%s", nombre) {
			faltan = true
		}
	}

	if faltan {
		t.FailNow()
	}
}
