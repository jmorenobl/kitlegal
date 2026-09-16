//go:build evals

package evals

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// directorioDeEvalsDeLasSkills es el de las evals de las skills en la raíz del
// repositorio, relativo al directorio de este paquete, que es donde go test
// ejecuta los tests (research.md V46): las de cada skill están en su
// subdirectorio.
const directorioDeEvalsDeLasSkills = "../../evals"

// Banderas con las que scripts/evals.sh invoca, tras -args, los dos arneses del
// job (contrato job-de-evals §3.2 y §3.3).
var (
	banderaSkill        = flag.String("skill", "", "skill evaluada")
	banderaEval         = flag.String("eval", "", "fichero de la eval con la que se juzga la sesión")
	banderaSesion       = flag.String("sesion", "", "directorio de la sesión que se prepara")
	banderaPruebaDeRed  = flag.Bool("prueba-de-red", false, "añade a la pregunta el texto de la prueba de red")
	banderaSesiones     = flag.String("sesiones", "", "directorio con un subdirectorio por sesión")
	banderaInforme      = flag.String("informe", "", "directorio en el que se escriben informe.md e informe.json")
	banderaModelo       = flag.String("modelo", "", "modelo con el que se abre la sesión")
	banderaQueDecide    = flag.String("modelo-que-decide", "", "modelo del job cuyas sesiones deciden el veredicto")
	banderaInformativos = flag.String("modelos-informativos", "",
		"modelos informativos del job, separados por comas; vacío si no hay ninguno")
	banderaRepeticiones = flag.Int("repeticiones", 0, "sesiones que se abren de cada eval con cada modelo")
	banderaUmbral       = flag.Int("umbral", 0, "sesiones de una serie que tienen que pasar para que la serie pase")
	banderaPlan         = flag.String("plan", "", "fichero en el que se escribe el plan de sesiones")
	banderaCommit       = flag.String("commit", "", "commit evaluado")
	banderaSinPython    = flag.String("sin-python", "", "ruta de sin-python.txt")
)

// TestPlanDeSesiones escribe en el fichero de -plan las sesiones que el job tiene
// que abrir, una por línea y con sus campos separados por tabuladores: nombre,
// fichero de eval, modelo y si la pregunta lleva el texto de la prueba de red
// (contrato job-de-evals §3.2). El guion lo lee y abre cada una; EscribirInforme
// vuelve a componer las series del mismo plan y exige que estén todas, de modo que
// el plan se escribe una sola vez. Falla si el directorio de evals tiene algún
// fichero mal formado: el guion ya lo ha comprobado antes, y planificar sobre un
// conjunto incompleto abriría menos sesiones sin decirlo. Solo lo ejecuta
// scripts/evals.sh; lo que decide lo fija TestPlan.
func TestPlanDeSesiones(t *testing.T) {
	t.Parallel()

	exigirBanderas(t, "skill", "modelo-que-decide", "repeticiones", "plan")

	conjunto, err := LeerConjunto(filepath.Join(directorioDeEvalsDeLasSkills, *banderaSkill))
	require.NoError(t, err)
	require.Empty(t, conjunto.MalFormados, "el directorio de evals no tiene ficheros mal formados")

	plan := PlanDeEvals{
		Evals:               conjunto.Evals,
		ModeloQueDecide:     *banderaQueDecide,
		ModelosInformativos: modelosInformativos(*banderaInformativos),
		Repeticiones:        *banderaRepeticiones,
		PruebaDeRed:         *banderaPruebaDeRed,
	}
	require.NoError(t, plan.Comprobar())

	var tabla strings.Builder
	for _, sesion := range plan.Sesiones() {
		fmt.Fprintf(&tabla, "%s\t%s\t%s\t%s\n", sesion.Nombre, sesion.Fichero, sesion.Modelo,
			siONo(sesion.PruebaDeRed))
	}

	require.NoError(t, os.WriteFile(*banderaPlan, []byte(tabla.String()), 0o600))
}

// modelosInformativos separa por comas el valor de -modelos-informativos; el valor
// vacío no es ningún modelo, y no uno con el nombre vacío.
func modelosInformativos(valor string) []string {
	if valor == "" {
		return nil
	}

	return strings.Split(valor, ",")
}

// TestPrepararSesion prepara el directorio de una sesión del job de evals con
// PrepararSesion: las evals de la skill de la raíz del repositorio, las
// grabaciones de H4 y de H5 en el orden en que las copia Preparar, y la eval, el
// directorio y la prueba de red de sus banderas. Falla con un error o con alguna
// falta de lo grabado (contrato job-de-evals §3.2). Solo lo ejecuta
// scripts/evals.sh; lo que decide lo fija TestPrepararDirectorioDeSesion.
func TestPrepararSesion(t *testing.T) {
	t.Parallel()

	exigirBanderas(t, "skill", "eval", "sesion", "modelo")

	faltas, err := PrepararSesion(SesionAPreparar{
		Evals:       filepath.Join(directorioDeEvalsDeLasSkills, *banderaSkill),
		Grabaciones: UnionDeGrabaciones(),
		Fichero:     *banderaEval,
		Modelo:      *banderaModelo,
		Directorio:  *banderaSesion,
		PruebaDeRed: *banderaPruebaDeRed,
	})
	require.NoError(t, err)

	textos := make([]string, 0, len(faltas))
	for _, falta := range faltas {
		textos = append(textos, falta.String())
	}

	require.Emptyf(t, faltas, "lo grabado no sirve estas consultas:\n%s", strings.Join(textos, "\n"))
}

// TestInformeDelJob escribe el informe de una ejecución del job de evals con
// EscribirInforme: las evals de la skill de la raíz del repositorio y las
// sesiones, el destino, el modelo, el commit y la comprobación sin Python de sus
// banderas. Falla con un error o con el veredicto fallo, nombrando sus motivos
// (contrato job-de-evals §3.3). Solo lo ejecuta scripts/evals.sh; lo que decide
// lo fijan TestInforme y TestEscribirInformeSinSusEntradas.
func TestInformeDelJob(t *testing.T) {
	t.Parallel()

	exigirBanderas(t, "skill", "sesiones", "informe", "modelo-que-decide", "repeticiones", "umbral", "commit",
		"sin-python")

	informe, err := EscribirInforme(InformeAEscribir{
		Skill:               *banderaSkill,
		Evals:               filepath.Join(directorioDeEvalsDeLasSkills, *banderaSkill),
		Sesiones:            *banderaSesiones,
		Destino:             *banderaInforme,
		ModeloQueDecide:     *banderaQueDecide,
		ModelosInformativos: modelosInformativos(*banderaInformativos),
		Repeticiones:        *banderaRepeticiones,
		Umbral:              *banderaUmbral,
		Commit:              *banderaCommit,
		SinPython:           *banderaSinPython,
	})
	require.NoError(t, err)
	require.Equalf(t, VeredictoAprobado, informe.Veredicto, "motivos del veredicto:\n%s", strings.Join(informe.Motivos, "\n"))
}

// exigirBanderas falla, nombrándola, si alguna de las banderas no tiene valor:
// con una ruta vacía, PrepararSesion o EscribirInforme leerían y escribirían en
// el directorio de este paquete en lugar de en el de la ejecución.
func exigirBanderas(t *testing.T, nombres ...string) {
	t.Helper()

	for _, nombre := range nombres {
		require.NotEmptyf(t, flag.Lookup(nombre).Value.String(), "falta la bandera -%s", nombre)
	}
}
