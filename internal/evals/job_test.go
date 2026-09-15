//go:build evals

package evals

import (
	"flag"
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
	banderaSkill       = flag.String("skill", "", "skill evaluada")
	banderaEval        = flag.String("eval", "", "fichero de la eval con la que se juzga la sesión")
	banderaSesion      = flag.String("sesion", "", "directorio de la sesión que se prepara")
	banderaPruebaDeRed = flag.Bool("prueba-de-red", false, "añade a la pregunta el texto de la prueba de red")
	banderaSesiones    = flag.String("sesiones", "", "directorio con un subdirectorio por sesión")
	banderaInforme     = flag.String("informe", "", "directorio en el que se escriben informe.md e informe.json")
	banderaModelo      = flag.String("modelo", "", "modelo fijado en el job")
	banderaCommit      = flag.String("commit", "", "commit evaluado")
	banderaSinPython   = flag.String("sin-python", "", "ruta de sin-python.txt")
)

// TestPrepararSesion prepara el directorio de una sesión del job de evals con
// PrepararSesion: las evals de la skill de la raíz del repositorio, las
// grabaciones de H4 y de H5 en el orden en que las copia Preparar, y la eval, el
// directorio y la prueba de red de sus banderas. Falla con un error o con alguna
// falta de lo grabado (contrato job-de-evals §3.2). Solo lo ejecuta
// scripts/evals.sh; lo que decide lo fija TestPrepararDirectorioDeSesion.
func TestPrepararSesion(t *testing.T) {
	t.Parallel()

	exigirBanderas(t, "skill", "eval", "sesion")

	faltas, err := PrepararSesion(SesionAPreparar{
		Evals:       filepath.Join(directorioDeEvalsDeLasSkills, *banderaSkill),
		Grabaciones: UnionDeGrabaciones(),
		Fichero:     *banderaEval,
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

	exigirBanderas(t, "skill", "sesiones", "informe", "modelo", "commit", "sin-python")

	informe, err := EscribirInforme(InformeAEscribir{
		Skill:     *banderaSkill,
		Evals:     filepath.Join(directorioDeEvalsDeLasSkills, *banderaSkill),
		Sesiones:  *banderaSesiones,
		Destino:   *banderaInforme,
		Modelo:    *banderaModelo,
		Commit:    *banderaCommit,
		SinPython: *banderaSinPython,
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
