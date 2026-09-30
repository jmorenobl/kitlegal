//go:build evals

package evals

import (
	"flag"
	"os"
	"os/signal"
	"path/filepath"
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
// (contracts/ejecucion-del-job.md §1 y §5 de H7.3).
var (
	banderaSkill        = flag.String("skill", "", "skill evaluada")
	banderaQueDecide    = flag.String("modelo-que-decide", "", "modelo del job cuyas sesiones deciden el veredicto")
	banderaInformativos = flag.String("modelos-informativos", "",
		"modelos informativos del job, separados por comas; vacío si no hay ninguno")
	banderaRepeticiones = flag.Int("repeticiones", 0, "sesiones que se abren de cada eval con cada modelo")
	banderaUmbral       = flag.Int("umbral", 0, "sesiones de una serie que tienen que pasar para que la serie pase")
	banderaConcurrencia = flag.Int("concurrencia", 0, "sesiones que se abren a la vez como mucho")
	banderaPruebaDeRed  = flag.Bool("prueba-de-red", false,
		"añade la sesión de la primera eval con el texto de la prueba de red")
	banderaObjetivo = flag.Int("objetivo-de-duracion", 0,
		"segundos que el job admite para sus sesiones; 0, sin objetivo")
	banderaSkills    = flag.String("skills", "", "directorio de las skills instaladas, el que deja make install")
	banderaCommit    = flag.String("commit", "", "commit evaluado")
	banderaSinPython = flag.String("sin-python", "", "ruta de sin-python.txt")
	banderaSesiones  = flag.String("sesiones", "", "directorio en el que se crea el de cada sesión")
	banderaInforme   = flag.String("informe", "", "directorio en el que se escriben informe.md e informe.json")
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

	evals := filepath.Join(directorioDeEvalsDeLasSkills, *banderaSkill)

	conjunto, err := LeerConjunto(evals)
	require.NoError(t, err)
	require.Empty(t, conjunto.MalFormados, "el directorio de evals no tiene ficheros mal formados")

	plan := PlanDeEvals{
		Evals:               conjunto.Evals,
		ModeloQueDecide:     *banderaQueDecide,
		ModelosInformativos: separarLosModelos(*banderaInformativos),
		Repeticiones:        *banderaRepeticiones,
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
		Concurrencia:  *banderaConcurrencia,
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
		Repeticiones:          *banderaRepeticiones,
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
