package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestDryRun comprueba las tres reglas de FR-022 sobre la misma invocación: la
// salida estándar queda vacía —también con --json—, la descripción del applet,
// el verbo y los argumentos que se habrían ejecutado van a la salida de error, y
// el código es 0.
//
// El segundo caso es el que da sentido a la decisión de research.md D10: con el
// registro de eventos apagado hasta el máximo la descripción **sigue estando**,
// porque no viaja por un canal filtrable sino por el presentador. Un requisito
// de visibilidad incondicional no se implementa sobre un canal que se filtra.
//
// No es paralelo, y no es un descuido: el nivel del registro es estado del
// proceso entero, así que fijarlo obliga a que este test tenga el proceso para
// él solo.
func TestDryRun(t *testing.T) {
	t.Setenv(cli.VariableNivel, "")

	casos := []struct {
		nombre string
		nivel  string
		argv   []string
	}{
		{
			nombre: "la salida estándar queda vacía también con --json",
			nivel:  "",
			argv:   []string{"kitlegal", "prueba", "hola", "--dry-run", "--json"},
		},
		{
			nombre: "la descripción se ve con el registro de eventos apagado",
			nivel:  "error",
			argv:   []string{"kitlegal", "prueba", "hola", "--dry-run"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Setenv(cli.VariableNivel, caso.nivel)

			res := invocar(t, registroDeCodigos(t, resultadoCorrecto), caso.argv...)

			assert.Equal(t, 0, res.codigo, "--dry-run termina con código 0")
			assert.Empty(t, res.salida,
				"una operación no realizada no tiene nada que citar")

			assert.Contains(t, res.errores, "prueba", "la descripción nombra el applet")
			assert.Contains(t, res.errores, "probar", "la descripción nombra el verbo")
			assert.Contains(t, res.errores, "hola",
				"la descripción lleva los argumentos que se habrían ejecutado")
		})
	}
}

// TestPlazoAgotado comprueba la mitad de FR-020 que ninguna otra tabla verifica:
// que vencer --timeout produce el código 4 **sin depender de que el applet
// devuelva el error tipado**.
//
// El applet de este caso espera a que el plazo venza y devuelve entonces un
// resultado correcto: si el kernel se fiara de lo que el applet devuelve, la
// invocación terminaría en 0. Lo que clasifica es el plazo agotado, que es
// indistinguible desde fuera de una fuente que no responde (research.md D9).
func TestPlazoAgotado(t *testing.T) {
	t.Parallel()

	registro := registroDeCodigos(t, func(ctx context.Context) (schema.Resultado, error) {
		<-ctx.Done()

		return schema.Resultado{
			Procedencia: procedenciaDePrueba,
			Datos:       map[string]any{"mensaje": "tarde"},
		}, nil
	})

	res := invocar(t, registro, "kitlegal", "prueba", "hola", "--json", "--timeout", "10ms")

	exigirSobreDeFallo(t, res, schema.ClaseFuenteNoDisponible, 4, procedenciaDePrueba)
	assert.Contains(t, res.errores, "plazo",
		"el mensaje para la persona dice que lo que falló fue el plazo")
}
