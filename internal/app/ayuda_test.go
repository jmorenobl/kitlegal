package app

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAyudaDelBinario comprueba lo que el binario responde a --help cuando no
// hay ningún applet resuelto: el modo de invocarlo y los applets registrados con
// su descripción, en el orden estable que da el registro (FR-026,
// contracts/registro-y-describe.md §4).
func TestAyudaDelBinario(t *testing.T) {
	t.Parallel()

	t.Run("enumera los applets registrados con su descripción", func(t *testing.T) {
		t.Parallel()

		ayuda := AyudaDelBinario("kitlegal", registroDeDespacho(t))

		assert.Contains(t, ayuda, "uso: kitlegal")
		assert.Contains(t, ayuda, "echo")
		assert.Contains(t, ayuda, "segundo")
		assert.Contains(t, ayuda, "applet de prueba", "cada applet sale con su descripción")
		assert.Less(t, strings.Index(ayuda, "echo"), strings.Index(ayuda, "segundo"),
			"los applets salen en el orden estable del registro")
		assert.True(t, strings.HasSuffix(ayuda, "\n"), "la ayuda cierra su última línea")
	})

	t.Run("nombra el binario por el nombre con el que se le invocó", func(t *testing.T) {
		t.Parallel()

		ayuda := AyudaDelBinario("kitlegal-dev", registroDeDespacho(t))

		assert.Contains(t, ayuda, "uso: kitlegal-dev",
			"un nombre de enlace desconocido se usa tal cual y no se corrige")
	})

	t.Run("un registro vacío lo dice en lugar de enumerar la nada", func(t *testing.T) {
		t.Parallel()

		ayuda := AyudaDelBinario("kitlegal", RegistroDeProduccion())

		assert.Contains(t, ayuda, "uso: kitlegal")
		assert.Contains(t, ayuda, "ningún applet",
			"el binario distribuido no registra ninguno en H1 y la ayuda no lo esconde")
	})

	t.Run("la misma ayuda dos veces es la misma ayuda", func(t *testing.T) {
		t.Parallel()

		registro := registroDeDespacho(t)

		primera := AyudaDelBinario("kitlegal", registro)
		segunda := AyudaDelBinario("kitlegal", registro)

		assert.Equal(t, primera, segunda,
			"dos invocaciones iguales producen la misma salida, byte a byte")
	})
}

// TestAyudaDerivadaDelRegistro es la comprobación de FR-001 sobre la ayuda:
// registrar un applet lo hace aparecer **sin tocar nada más**. Si existiera una
// lista paralela mantenida a mano, el applet nuevo no saldría aquí.
func TestAyudaDerivadaDelRegistro(t *testing.T) {
	t.Parallel()

	var registro Registro
	require.NoError(t, registro.Registrar(appletConVerbos("echo", verboDePrueba("repetir", true))))

	antes := AyudaDelBinario("kitlegal", &registro)
	assert.NotContains(t, antes, "plazos", "todavía no lo ha registrado nadie")

	require.NoError(t, registro.Registrar(appletConVerbos("plazos",
		verboDePrueba("vencimiento", true))))

	despues := AyudaDelBinario("kitlegal", &registro)
	assert.Contains(t, despues, "plazos",
		"registrar un applet basta para que aparezca en la ayuda")
	assert.Contains(t, despues, "echo", "y el que ya estaba sigue estando")
}

// TestAyudaDelApplet comprueba lo que un applet responde a --help: sus verbos,
// también derivados del registro, con el verbo por omisión marcado como tal.
// Pedir la ayuda de un applet no resuelve ningún verbo, así que la lista es la
// del applet entero (contracts/registro-y-describe.md §2 bis y §4).
func TestAyudaDelApplet(t *testing.T) {
	t.Parallel()

	t.Run("enumera los verbos y marca el que se toma por omisión", func(t *testing.T) {
		t.Parallel()

		applet := appletConVerbos("echo", verboDePrueba("repetir", true))

		ayuda := AyudaDelApplet(applet)

		assert.Contains(t, ayuda, "uso: echo")
		assert.Contains(t, ayuda, "repetir")
		assert.Contains(t, ayuda, "verbo de prueba", "cada verbo sale con su descripción")
		assert.Contains(t, ayuda, "por omisión", "el verbo que se toma sin nombrarlo va marcado")
		assert.True(t, strings.HasSuffix(ayuda, "\n"), "la ayuda cierra su última línea")
	})

	t.Run("un applet sin verbo por omisión no marca ninguno", func(t *testing.T) {
		t.Parallel()

		applet := appletConVerbos("segundo",
			verboDePrueba("uno", false), verboDePrueba("dos", false))

		ayuda := AyudaDelApplet(applet)

		assert.Contains(t, ayuda, "uso: segundo")
		assert.Contains(t, ayuda, "uno")
		assert.Contains(t, ayuda, "dos")
		assert.NotContains(t, ayuda, "por omisión",
			"no hay ninguno que tomar sin nombrarlo: el verbo es obligatorio")
	})

	t.Run("los verbos salen en el orden en que el applet los declara", func(t *testing.T) {
		t.Parallel()

		applet := appletConVerbos("boe",
			verboDePrueba("articulo", false), verboDePrueba("buscar", false))

		ayuda := AyudaDelApplet(applet)

		assert.Less(t, strings.Index(ayuda, "articulo"), strings.Index(ayuda, "buscar"))
	})

	t.Run("un verbo nuevo aparece sin tocar nada más", func(t *testing.T) {
		t.Parallel()

		uno := AyudaDelApplet(appletConVerbos("boe", verboDePrueba("articulo", true)))
		assert.NotContains(t, uno, "buscar")

		dos := AyudaDelApplet(appletConVerbos("boe",
			verboDePrueba("articulo", true), verboDePrueba("buscar", false)))
		assert.Contains(t, dos, "buscar",
			"la lista de verbos es la que declara el applet, no una copia mantenida a mano")
	})
}
