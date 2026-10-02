package mcp

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// topeDeLasInstrucciones son los bytes dentro de los que tienen que caber los
// cinco contenidos: la parte que la documentación de Codex pide que se baste a
// sí misma (FR-006). Se cuentan bytes, que acotan por arriba a los caracteres
// (research.md D10).
const topeDeLasInstrucciones = 512

// contenidosDeLasInstrucciones son los cinco contenidos de FR-006, en su
// orden, cada uno con la frase que lo dice (contracts/servidor-mcp.md §5).
var contenidosDeLasInstrucciones = []struct{ contenido, frase string }{
	{
		contenido: "qué son las herramientas",
		frase: "Herramientas de kitlegal: consultan fuentes legales públicas españolas y devuelven un sobre con " +
			"el texto vigente de una norma consolidada del BOE o con el territorio de un municipio.",
	},
	{
		contenido: "ningún contenido legal se afirma si no viene del texto devuelto",
		frase:     "No afirmes ningún contenido legal que no venga del texto devuelto en esta conversación.",
	},
	{
		contenido: "cada afirmación lleva norma y bloque",
		frase: "Cada afirmación lleva su cita, con la norma y el bloque: art. 21 de la Ley 39/2015 " +
			"[BOE-A-2015-10565, bloque a21].",
	},
	{
		contenido: "los avisos del sobre se trasladan",
		frase:     "Traslada los avisos del sobre (data.avisos).",
	},
	{
		contenido: "el protocolo completo son las skills de kitlegal",
		frase:     "El protocolo completo son las skills de kitlegal.",
	},
}

// TestInstrucciones comprueba el texto fijo que el servidor da como
// `instructions`: sus cinco contenidos, en su orden y dentro de los primeros
// 512 bytes, y su tamaño (FR-006, FR-077; SC-010).
func TestInstrucciones(t *testing.T) {
	t.Parallel()

	t.Run("mide 485 bytes", func(t *testing.T) {
		t.Parallel()

		assert.Len(t, Instrucciones, 485, "el tamaño que fija contracts/servidor-mcp.md §5")
	})

	t.Run("las cinco frases, en su orden, dentro de los primeros 512 bytes", func(t *testing.T) {
		t.Parallel()

		for _, defecto := range defectosDeLasInstrucciones(Instrucciones) {
			t.Error(defecto)
		}
	})

	t.Run("el texto son las cinco frases y nada más", func(t *testing.T) {
		t.Parallel()

		frases := make([]string, 0, len(contenidosDeLasInstrucciones))
		for _, contenido := range contenidosDeLasInstrucciones {
			frases = append(frases, contenido.frase)
		}

		if esperado := strings.Join(frases, " "); Instrucciones != esperado {
			t.Errorf("el texto no es el de contracts/servidor-mcp.md §5, carácter a carácter:\n   es: %q\nsería: %q",
				Instrucciones, esperado)
		}
	})

	t.Run("control: una frase detrás del byte 512, una que falta y una fuera de su orden se nombran", func(t *testing.T) {
		t.Parallel()

		ultima := contenidosDeLasInstrucciones[4]
		cuarta := contenidosDeLasInstrucciones[3]
		require.Contains(t, Instrucciones, ultima.frase)
		require.Contains(t, Instrucciones, cuarta.frase)

		relleno := strings.Repeat("x", topeDeLasInstrucciones)

		desplazada := strings.Replace(Instrucciones, ultima.frase, relleno+" "+ultima.frase, 1)
		assert.Equal(t, []string{
			fmt.Sprintf("la frase 5 de las instrucciones (%s) termina en el byte %d, detrás del %d: %q",
				ultima.contenido, len(desplazada), topeDeLasInstrucciones, ultima.frase),
		}, defectosDeLasInstrucciones(desplazada))

		sinLaCuarta := strings.Replace(Instrucciones, cuarta.frase, "", 1)
		assert.Equal(t, []string{
			fmt.Sprintf("la frase 4 de las instrucciones (%s) no está: %q", cuarta.contenido, cuarta.frase),
		}, defectosDeLasInstrucciones(sinLaCuarta))

		// Las dos últimas, cambiadas de sitio: la cuarta queda detrás de la quinta.
		cambiadas := strings.Replace(sinLaCuarta, ultima.frase, ultima.frase+" "+cuarta.frase, 1)
		assert.Equal(t, []string{
			fmt.Sprintf("la frase 5 de las instrucciones (%s) no va detrás de la frase 4: %q",
				ultima.contenido, ultima.frase),
		}, defectosDeLasInstrucciones(cambiadas))
	})
}

// defectosDeLasInstrucciones devuelve, en el orden de los contenidos, lo que
// le falta al texto para cumplir FR-006: cada frase que no está, que no va
// detrás de la anterior o que no termina dentro del tope, nombrada con su
// número, su contenido y su texto.
func defectosDeLasInstrucciones(texto string) []string {
	var defectos []string

	finDeLaAnterior := 0

	for i, contenido := range contenidosDeLasInstrucciones {
		numero := i + 1

		comienzo := strings.Index(texto, contenido.frase)
		if comienzo < 0 {
			defectos = append(defectos, fmt.Sprintf("la frase %d de las instrucciones (%s) no está: %q",
				numero, contenido.contenido, contenido.frase))

			continue
		}

		fin := comienzo + len(contenido.frase)

		switch {
		case comienzo < finDeLaAnterior:
			defectos = append(defectos, fmt.Sprintf("la frase %d de las instrucciones (%s) no va detrás de la frase %d: %q",
				numero, contenido.contenido, numero-1, contenido.frase))
		case fin > topeDeLasInstrucciones:
			defectos = append(defectos, fmt.Sprintf("la frase %d de las instrucciones (%s) termina en el byte %d, detrás del %d: %q",
				numero, contenido.contenido, fin, topeDeLasInstrucciones, contenido.frase))
		}

		finDeLaAnterior = max(finDeLaAnterior, fin)
	}

	return defectos
}
