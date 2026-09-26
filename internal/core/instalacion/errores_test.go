package instalacion_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestManifiestoIlegibleSinCausa exige que el valor cero del error, que
// cualquiera puede construir porque el tipo se exporta para reconocerlo con
// errors.As, se pueda usar sin un panic: dice que el manifiesto es ilegible y
// no envuelve nada (FR-144).
func TestManifiestoIlegibleSinCausa(t *testing.T) {
	t.Parallel()

	var ilegible instalacion.ManifiestoIlegible

	assert.Equal(t, "manifiesto ilegible", ilegible.Error())
	assert.NoError(t, ilegible.Unwrap())
}

// TestErrorDeConflictosSinNinguno exige lo mismo del valor cero del rechazo
// por conflicto: su mensaje es la cabecera sola, declara la clase
// «inesperado» y no nombra ninguno (FR-144).
func TestErrorDeConflictosSinNinguno(t *testing.T) {
	t.Parallel()

	var rechazo instalacion.ErrorDeConflictos

	assert.Equal(t, cabeceraDeInstall, rechazo.Error())
	assert.Equal(t, schema.ClaseInesperado, rechazo.Clase())
	assert.Empty(t, rechazo.Lista())
}

// TestListaDeConflictosEsUnaCopia exige que cambiar la lista que devuelve el
// rechazo no cambie lo que nombra.
func TestListaDeConflictosEsUnaCopia(t *testing.T) {
	t.Parallel()

	d := nuevoDiscoEnMemoria(t)
	d.fichero(".agents", "no soy un directorio")

	pedido, err := instalacion.ValidarInvocacion(instalacion.Invocacion{}, "", empotradasDePrueba())
	require.NoError(t, err)

	var rechazo *instalacion.ErrorDeConflictos
	require.ErrorAs(t, instalacion.ComprobarConflictos(d, &enlazadorDePrueba{}, pedido, empotradasDePrueba()), &rechazo)

	lista := rechazo.Lista()
	lista[0].Ruta = "otra"

	assert.Equal(t, []instalacion.Conflicto{conflicto(noEsDirectorio, ".agents")}, rechazo.Lista())
	assert.Equal(t, cabeceraDeInstall+"\nruta que no es directorio: .agents", rechazo.Error())
}

// TestErrorDeHallazgosSinNinguno exige que el valor cero del error de doctor
// con hallazgos se pueda usar sin un panic: su mensaje es la cabecera sola,
// declara la clase «inesperado» y no nombra ninguno (FR-144).
func TestErrorDeHallazgosSinNinguno(t *testing.T) {
	t.Parallel()

	var rechazo instalacion.ErrorDeHallazgos

	assert.Equal(t, "skills doctor: 0 hallazgos:", rechazo.Error())
	assert.Equal(t, schema.ClaseInesperado, rechazo.Clase())
	assert.Empty(t, rechazo.Lista())
}

// TestListaDeHallazgosEsUnaCopia exige que cambiar la lista que devuelve el
// error de doctor no cambie lo que nombra.
func TestListaDeHallazgosEsUnaCopia(t *testing.T) {
	t.Parallel()

	d := nuevoDiscoEnMemoria(t)
	instalarLocal(d, "legal-core").escribir()
	d.fichero(neutroLc+"/SKILL.md", "editado a mano")

	_, err := instalacion.Diagnosticar(d, &enlazadorDePrueba{}, ambitoLocal, empotradasDePrueba(), versionDePrueba)

	var rechazo *instalacion.ErrorDeHallazgos
	require.ErrorAs(t, err, &rechazo)

	lista := rechazo.Lista()
	lista[0].Ruta = "otra"

	assert.Equal(t, neutroLc+"/SKILL.md", rechazo.Lista()[0].Ruta)
}

// TestAmbitoIlegibleSinMotivo exige lo mismo del valor cero del ámbito
// ilegible de list y doctor: declara la clase «inesperado», no nombra ningún
// motivo y su mensaje no provoca un panic (FR-144).
func TestAmbitoIlegibleSinMotivo(t *testing.T) {
	t.Parallel()

	var rechazo instalacion.AmbitoIlegible

	assert.NotPanics(t, func() { _ = rechazo.Error() })
	assert.Equal(t, schema.ClaseInesperado, rechazo.Clase())
	assert.Equal(t, instalacion.Conflicto{}, rechazo.Motivo())
}
