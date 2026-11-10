package workflow

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestElHitoPasaADoneEnSuRama: el paso hito_a_done mueve el fichero del hito de
// docs/roadmap/in-progress/ a done/ en un commit de la rama del hito, para que
// al fusionarla main lo tenga hecho y lo haya visto cada ronda de la revisión y cada
// medición (ADR 0030). Al reanudar no lo vuelve a mover, y en main no mueve
// nada (scripts/workflow/hito-a-done.sh; docs/roadmap/README.md).
func TestElHitoPasaADoneEnSuRama(t *testing.T) {
	t.Parallel()

	const (
		guion   = "scripts/workflow/hito-a-done.sh"
		enDone  = "docs/roadmap/done/H0-hito-del-test.md"
		mensaje = "docs(H0): el hito pasa a done/ con su propuesta de cambio"
	)

	a := nuevoArbol(t)
	antes := a.cabeza()

	a.debeSalirBien(nil, guion, hitoDelTest)

	assert.True(t, a.noExiste(ficheroDelHitoDelTest), "el fichero sale de in-progress/")
	assert.Equal(t, hitoDelRoadmapDelTest, a.leer(enDone), "y llega a done/ con su contenido")
	assert.Equal(t, mensaje, a.git("log", "-1", "--format=%s"), "en un commit con el asunto que reconoce revision.sh")
	assert.Equal(t, antes, a.git("rev-parse", "HEAD~1"), "un solo commit")
	assert.Empty(t, a.git("status", "--porcelain"), "sin nada sin commitear")

	cabeza := a.cabeza()
	a.debeSalirBien(nil, guion, hitoDelTest)
	assert.Equal(t, cabeza, a.cabeza(), "al reanudar, un hito que ya está en done/ no se vuelve a mover")

	a.git("switch", "-q", "main")

	enMain := a.ejecutar(nil, guion, hitoDelTest)
	require.NotZero(t, enMain.codigo, "en main no se mueve")
	assert.Contains(t, enMain.errores, "en main el hito no pasa a done/")
	assert.True(t, a.existe(ficheroDelHitoDelTest), "y el fichero sigue en in-progress/")
}

// TestElHitoPasaADoneTrasUnCommitSinTerminar: si un intento anterior dejó el
// traslado preparado (el `git mv` hecho y el commit sin terminar, como en H41
// cuando el gancho de pre-commit pasó del timeout del paso), al reanudar el paso
// confirma ese traslado en su commit en vez de darlo por hecho porque el fichero
// ya está en done/.
func TestElHitoPasaADoneTrasUnCommitSinTerminar(t *testing.T) {
	t.Parallel()

	const (
		guion   = "scripts/workflow/hito-a-done.sh"
		enDone  = "docs/roadmap/done/H0-hito-del-test.md"
		mensaje = "docs(H0): el hito pasa a done/ con su propuesta de cambio"
	)

	a := nuevoArbol(t)
	antes := a.cabeza()
	require.NoError(t, os.MkdirAll(a.ruta("docs/roadmap/done"), 0o750))
	a.git("mv", ficheroDelHitoDelTest, enDone)

	a.debeSalirBien(nil, guion, hitoDelTest)

	assert.Equal(t, mensaje, a.git("log", "-1", "--format=%s"), "el traslado preparado se confirma con su asunto")
	assert.Equal(t, antes, a.git("rev-parse", "HEAD~1"), "en un solo commit")
	assert.Equal(t, "A\t"+enDone+"\nD\t"+ficheroDelHitoDelTest,
		a.git("show", "--no-renames", "--name-status", "--format=", "HEAD"), "con el borrado del origen y el fichero en done/")
	assert.Empty(t, a.git("status", "--porcelain"), "sin nada sin commitear")

	cabeza := a.cabeza()
	a.debeSalirBien(nil, guion, hitoDelTest)
	assert.Equal(t, cabeza, a.cabeza(), "y al reanudar otra vez no hace nada")
}
