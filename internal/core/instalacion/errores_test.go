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
// «conflicto» y no nombra ninguno (FR-144; ADR 0023).
func TestErrorDeConflictosSinNinguno(t *testing.T) {
	t.Parallel()

	var rechazo instalacion.ErrorDeConflictos

	assert.Equal(t, cabeceraDeInstall, rechazo.Error())
	assert.Equal(t, schema.ClaseConflicto, rechazo.Clase())
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

// TestAmbitoIlegibleSinMotivo exige lo mismo del valor cero del ámbito
// ilegible de list y doctor: declara la clase «conflicto», no nombra ningún
// motivo y su mensaje no provoca un panic (FR-144; ADR 0023).
func TestAmbitoIlegibleSinMotivo(t *testing.T) {
	t.Parallel()

	var rechazo instalacion.AmbitoIlegible

	assert.NotPanics(t, func() { _ = rechazo.Error() })
	assert.Equal(t, schema.ClaseConflicto, rechazo.Clase())
	assert.Equal(t, instalacion.Conflicto{}, rechazo.Motivo())
}
