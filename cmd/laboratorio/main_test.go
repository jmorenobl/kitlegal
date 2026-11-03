package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Las líneas de uso de laboratorio (contracts/arnes.md §5 de H39;
// contracts/tabla.md §1 de H39): la del filtro del servidor y la del guion que
// cuenta un sondeo. Sin orden, o con una que no es suya, laboratorio escribe las
// dos; cada orden, ante unos argumentos sin su forma, la suya.
const (
	usoDelServidor = "laboratorio: uso: laboratorio servidor -kitlegal <ruta> [-instrucciones <texto>] " +
		"[-herramienta <h> -parametro <p> -descripcion <texto>]\n"
	usoDeLaTabla       = "laboratorio: uso: laboratorio tabla <ruta>\n"
	usoDelLaboratorio  = usoDelServidor + usoDeLaTabla
	prefijoDeUnDefecto = "laboratorio: "
)

// TestOrdenesDelLaboratorio fija el despacho de laboratorio: sin orden, o con
// una que no es suya, sus líneas de uso en la salida de error, nada en la
// estándar y 1; con servidor o con tabla, lo que diga cada orden con los
// argumentos que siguen a ella, aquí unos sin su forma, que dan su línea de uso
// y 1 sin arrancar ni leer nada; y con tabla y una ruta que no existe, su
// defecto con la ruta y 1 sin tabla. Lo que hace cada orden lo fijan
// TestServidorDelLaboratorio y TestTablaDelLaboratorio.
func TestOrdenesDelLaboratorio(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		args    []string
		errores string
	}{
		{nombre: "sin-orden", args: []string{"laboratorio"}, errores: usoDelLaboratorio},
		{nombre: "sin-argumentos", errores: usoDelLaboratorio},
		{
			nombre:  "orden-desconocida",
			args:    []string{"laboratorio", "otra", "-kitlegal", "/bin/kitlegal"},
			errores: usoDelLaboratorio,
		},
		{
			nombre:  "servidor-sin-su-forma",
			args:    []string{"laboratorio", "servidor", "-instrucciones", "x"},
			errores: usoDelServidor,
		},
		{nombre: "tabla-sin-ruta", args: []string{"laboratorio", "tabla"}, errores: usoDeLaTabla},
		{nombre: "tabla-con-dos-rutas", args: []string{"laboratorio", "tabla", "a", "b"}, errores: usoDeLaTabla},
		{nombre: "tabla-con-la-ruta-vacia", args: []string{"laboratorio", "tabla", ""}, errores: usoDeLaTabla},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			var salida, errores bytes.Buffer

			codigo := ejecutar(t.Context(), caso.args, strings.NewReader(""), &salida, &errores)

			assert.Equal(t, 1, codigo)
			assert.Equal(t, caso.errores, errores.String())
			assert.Empty(t, salida.String())
		})
	}

	t.Run("tabla-de-una-ruta-que-no-existe", func(t *testing.T) {
		t.Parallel()

		var salida, errores bytes.Buffer

		ruta := filepath.Join(t.TempDir(), "no-existe")
		codigo := ejecutar(t.Context(), []string{"laboratorio", "tabla", ruta}, strings.NewReader(""), &salida, &errores)

		assert.Equal(t, 1, codigo)
		assert.True(t, strings.HasPrefix(errores.String(), prefijoDeUnDefecto), "con el prefijo: %s", errores.String())
		assert.Contains(t, errores.String(), ruta)
		assert.Equal(t, 1, strings.Count(errores.String(), "\n"), "una sola línea: %s", errores.String())
		assert.Empty(t, salida.String(), "sin tabla")
	})
}
