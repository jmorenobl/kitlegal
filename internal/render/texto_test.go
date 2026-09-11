package render

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// versionDeEjemplo son las tres líneas con las que responde el verbo reservado
// `version` desde H0. El presentador las escribe tal cual: quien compone el
// texto es el kernel, y quien lo hace llegar a la salida estándar es este
// paquete (research.md D16, contracts/cli-version.md de H0).
const versionDeEjemplo = "kitlegal 1.2.3\ncommit: 3ae7549\nfecha:  2026-09-11T08:12:00Z\n"

// TestTexto comprueba el texto dirigido a una persona: por qué descriptor sale
// cada clase y que llega entero y sin adornos (FR-026, FR-042,
// contracts/banderas-y-exit-codes.md §5 y §6).
func TestTexto(t *testing.T) {
	t.Parallel()

	t.Run("las tres líneas de version salen tal cual por la salida estándar", func(t *testing.T) {
		t.Parallel()

		var salida, errores bytes.Buffer
		require.NoError(t, Nuevo(&salida, &errores).Texto(versionDeEjemplo))

		assert.Equal(t, versionDeEjemplo, salida.String())
		assert.Empty(t, errores.String())
	})

	t.Run("la ayuda derivada del registro conserva sus líneas", func(t *testing.T) {
		t.Parallel()

		ayuda := "uso: kitlegal <applet> [verbo] [banderas]\n\napplets:\n  echo  repite lo que se le da"

		var salida, errores bytes.Buffer
		require.NoError(t, Nuevo(&salida, &errores).Texto(ayuda))

		assert.Equal(t, ayuda+"\n", salida.String(),
			"se cierra la última línea y no se toca nada más")
		assert.Empty(t, errores.String())
	})

	casos := []struct {
		nombre   string
		texto    string
		esperado string
	}{
		{
			nombre:   "se cierra la línea cuando el texto no la cierra",
			texto:    "applet desconocido: fiscal",
			esperado: "applet desconocido: fiscal\n",
		},
		{
			nombre:   "y no se duplica el salto cuando ya la cierra",
			texto:    "applet desconocido: fiscal\n",
			esperado: "applet desconocido: fiscal\n",
		},
		{
			nombre:   "un texto vacío no escribe ninguna línea",
			texto:    "",
			esperado: "",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			var salida, errores bytes.Buffer
			presentador := Nuevo(&salida, &errores)

			require.NoError(t, presentador.Texto(caso.texto))
			require.NoError(t, presentador.Aviso(caso.texto))

			assert.Equal(t, caso.esperado, salida.String())
			assert.Equal(t, caso.esperado, errores.String())
		})
	}
}

// TestAviso comprueba los tres mensajes que el contrato manda por la salida de
// error sin pasar por el registro de eventos —el fallo, la lista de applets y
// la descripción de --dry-run—, que es lo que hace que ningún nivel pueda
// ocultarlos (FR-022, FR-040, contracts/banderas-y-exit-codes.md §6).
func TestAviso(t *testing.T) {
	t.Parallel()

	avisos := []string{
		"no existe el artículo 99 de BOE-A-2015-10565",
		"applets disponibles: boe, placsp",
		"dry-run: echo repetir [hola] (no se ha ejecutado nada)",
	}

	for _, aviso := range avisos {
		t.Run(aviso, func(t *testing.T) {
			t.Parallel()

			var salida, errores bytes.Buffer
			require.NoError(t, Nuevo(&salida, &errores).Aviso(aviso))

			// La igualdad exacta es la comprobación: un registro de slog llevaría
			// siempre `time=`, `level=` y `msg=` delante. Que no estén es lo que
			// demuestra que el aviso no viaja por un canal que un nivel filtra.
			assert.Equal(t, aviso+"\n", errores.String())
			assert.Empty(t, salida.String(), "un aviso nunca toca la salida estándar")
		})
	}
}
