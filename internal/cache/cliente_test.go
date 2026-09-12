package cache

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestNewRechazaOpcionesInvalidas fija la fila 3 de la tabla cerrada del
// contrato de errores §3: una opción con un valor que no sirve es «argumentos»
// (2) y el mensaje nombra **la opción**, porque lo que hay que corregir está en
// el código que la pasa y no en lo que la persona escribió (D2).
//
// Fija además que las opciones se validan **en orden** y que la primera
// inválida es la que decide: la última fila pasa dos opciones inválidas y exige
// que el mensaje nombre la primera y no la segunda. Sin eso, validarlas en
// cualquier orden pasaría igual y el fallo señalaría un sitio que quien lo lee
// no tocó.
//
// El control positivo de la cabecera es lo que impide que la tabla pase con un
// New que fallara siempre: con las cuatro opciones bien construidas no falla, y
// el cliente que devuelve se cierra —dos veces, porque Close es idempotente
// (FR-004)— sin error.
func TestNewRechazaOpcionesInvalidas(t *testing.T) {
	t.Parallel()

	cliente, err := New(t.Context(),
		ConDirectorio(t.TempDir()),
		SoloLectura(),
		ConReloj(time.Now),
		ConRegistrador(slog.New(slog.DiscardHandler)))
	require.NoError(t, err)
	require.NotNil(t, cliente)
	require.NoError(t, cliente.Close())
	require.NoError(t, cliente.Close(),
		"cerrar dos veces no es un fallo: el defer de quien lo construyó y el cierre "+
			"explícito conviven (FR-004)")

	casos := []struct {
		nombre   string
		opciones []Opcion
		nombra   string
		noNombra string
	}{
		{
			nombre:   "un directorio sin ninguna ruta dentro",
			opciones: []Opcion{ConDirectorio("")},
			nombra:   "ConDirectorio",
		},
		{
			nombre:   "un reloj nulo",
			opciones: []Opcion{ConReloj(nil)},
			nombra:   "ConReloj",
		},
		{
			nombre:   "un registrador nulo",
			opciones: []Opcion{ConRegistrador(nil)},
			nombra:   "ConRegistrador",
		},
		{
			nombre:   "con dos inválidas, la que se nombra es la primera",
			opciones: []Opcion{ConReloj(nil), ConRegistrador(nil)},
			nombra:   "ConReloj",
			noNombra: "ConRegistrador",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			cliente, err := New(t.Context(), caso.opciones...)

			require.Error(t, err)
			assert.Nil(t, cliente)
			assert.Equal(t, schema.ClaseArgumentos, cli.Clasificar(err))
			assert.Equal(t, 2, cli.CodigoSalida(err))
			assert.Contains(t, err.Error(), "opción "+caso.nombra,
				"el mensaje nombra la opción que hay que corregir")

			if caso.noNombra != "" {
				assert.NotContains(t, err.Error(), caso.noNombra,
					"las opciones se validan en orden: la segunda inválida no se llega a mirar")
			}
		})
	}
}
