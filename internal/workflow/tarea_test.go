package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnIntentoCortadoPorElLimiteNoCuenta: cuando el límite de uso corta una
// sesión del intento en curso, el supervisor llama a `tarea.sh sin-modelo`, y
// el intento siguiente de esa tarea lleva el mismo número y no la acerca a la
// cuarentena. En H9, los tres intentos de T015 chocaron con el límite y la tarea
// entró en cuarentena sin haber fallado (scripts/workflow/tarea.sh).
func TestUnIntentoCortadoPorElLimiteNoCuenta(t *testing.T) {
	t.Parallel()

	const (
		guion   = "scripts/workflow/tarea.sh"
		marca   = gatesDelHito + "/intento-sin-modelo"
		tarea   = "T001"
		tareas  = "- [ ] T001 La tarea del test, en `" + ficheroDelProducto + "`\n"
		ajena   = "T999"
		maximo  = 3
		supuest = gatesDelHito + "/supuestos.md"
	)

	a := nuevoArbol(t)
	a.escribir(directorioDelHito+"/tasks.md", tareas)
	a.commit("tasks.md del test")

	siguiente := func() (string, int) {
		t.Helper()

		a.debeSalirBien(nil, guion, "siguiente", hitoDelTest)

		var actual struct {
			ID      string `json:"id"`
			Intento int    `json:"intento"`
		}

		a.leerJSON(gatesDelHito+"/tarea-actual.json", &actual)

		return actual.ID, actual.Intento
	}

	id, intento := siguiente()
	require.Equal(t, tarea, id)
	require.Equal(t, 1, intento, "el primer intento")

	a.debeSalirBien(nil, guion, "sin-modelo")
	require.True(t, a.existe(marca), "sin-modelo deja la marca de la tarea en curso")

	_, intento = siguiente()
	assert.Equal(t, 1, intento, "el intento cortado se repite con su número")
	assert.True(t, a.noExiste(marca), "y la marca se consume")
	assert.Contains(t, a.leer(supuest), "Intento 1 de T001 cortado por el límite de uso", "el informe lo cuenta")

	_, intento = siguiente()
	assert.Equal(t, 2, intento, "un intento sin marca cuenta")

	a.escribir(marca, ajena+"\n")

	_, intento = siguiente()
	assert.Equal(t, 3, intento, "la marca de otra tarea no descuenta nada")
	assert.True(t, a.noExiste(marca), "y también se consume")

	a.debeSalirBien(nil, guion, "sin-modelo")

	_, intento = siguiente()
	assert.Equal(t, maximo, intento, "el último intento cortado no manda la tarea a cuarentena")
	assert.NotContains(t, a.leer(gatesDelHito+"/cuarentena.json"), tarea)

	a.debeSalirBien(nil, guion, "siguiente", hitoDelTest)
	assert.Contains(t, a.leer(gatesDelHito+"/cuarentena.json"), tarea, "un cuarto intento que cuenta, sí")
}
