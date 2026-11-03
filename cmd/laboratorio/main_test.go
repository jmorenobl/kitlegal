package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// usoDelLaboratorio es la línea de uso de laboratorio (contracts/arnes.md §5
// de H39).
const usoDelLaboratorio = "laboratorio: uso: laboratorio servidor -kitlegal <ruta> [-instrucciones <texto>] " +
	"[-herramienta <h> -parametro <p> -descripcion <texto>]\n"

// TestOrdenesDelLaboratorio fija el despacho de laboratorio: sin orden, o con
// una que no es suya, su línea de uso en la salida de error, nada en la
// estándar y 1; con servidor, lo que diga el filtro con los argumentos que
// siguen a la orden, aquí unos sin su forma, que dan la misma línea y 1 sin
// arrancar nada. Lo que hace el filtro lo fija TestServidorDelLaboratorio.
func TestOrdenesDelLaboratorio(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		args   []string
	}{
		{nombre: "sin-orden", args: []string{"laboratorio"}},
		{nombre: "sin-argumentos"},
		{nombre: "orden-desconocida", args: []string{"laboratorio", "otra", "-kitlegal", "/bin/kitlegal"}},
		{nombre: "servidor-sin-su-forma", args: []string{"laboratorio", "servidor", "-instrucciones", "x"}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			var salida, errores bytes.Buffer

			codigo := ejecutar(t.Context(), caso.args, strings.NewReader(""), &salida, &errores)

			assert.Equal(t, 1, codigo)
			assert.Equal(t, usoDelLaboratorio, errores.String())
			assert.Empty(t, salida.String())
		})
	}
}
