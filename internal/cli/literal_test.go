package cli

import (
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// verbosConLiteral es un applet con un verbo que recibe el mismo valor dos
// veces, por su posición: una en un Literal y otra en un string, que es la
// comparación que dice qué conserva cada uno.
type verbosConLiteral struct {
	Buscar struct {
		ID     Literal `arg:"" name:"id" help:"Lo que se busca, tal cual."`
		Cadena string  `arg:"" name:"cadena" help:"Lo mismo, como cadena."`
	} `cmd:"" help:"Busca lo que se le da."`
}

// analisisConLiteral recoge lo observable de una llamada al análisis con el
// applet del Literal: lo que devolvió, lo que quedó escrito en su gramática y
// lo que salió por los descriptores.
type analisisConLiteral struct {
	analisis Analisis
	err      error
	verbos   *verbosConLiteral
	doble    *presentadorDoble
}

// analizarConLiteral analiza una invocación con el applet del Literal.
func analizarConLiteral(t *testing.T, args ...string) analisisConLiteral {
	t.Helper()

	res := analisisConLiteral{verbos: &verbosConLiteral{}, doble: &presentadorDoble{}}
	res.analisis, res.err = Analizar(res.doble, nombreDePrueba, res.verbos, args)

	return res
}

// TestAnalizarLiteral comprueba que un Literal llega al applet con los bytes que
// se escribieron, también los que no son UTF-8, que un string recibe cambiados
// por U+FFFD (FR-052 y FR-053 de H7): lo que el Literal conserva no depende de
// que el valor sea UTF-8 válido, esté vacío o parezca una bandera tras el
// terminador, y las banderas globales se analizan igual a su alrededor.
func TestAnalizarLiteral(t *testing.T) {
	t.Parallel()

	for _, caso := range []struct {
		nombre string
		args   []string
		valor  string
		cadena string
		json   bool
	}{
		{nombre: "un id UTF-8", args: []string{"buscar", "ine:28074", "ine:28074"}, valor: "ine:28074", cadena: "ine:28074"},
		{
			nombre: "un byte que no es UTF-8 en medio",
			args:   []string{"buscar", "a\xffb", "a\xffb"}, valor: "a\xffb", cadena: "a\xef\xbf\xbdb",
		},
		{nombre: "solo un byte que no es UTF-8", args: []string{"buscar", "\xff", "\xff"}, valor: "\xff", cadena: "\xef\xbf\xbd"},
		{
			nombre: "una secuencia UTF-8 cortada",
			args:   []string{"buscar", "a\xe2\x80", "a\xe2\x80"}, valor: "a\xe2\x80", cadena: "a\xef\xbf\xbd\xef\xbf\xbd",
		},
		{nombre: "la cadena vacía", args: []string{"buscar", "", ""}, valor: "", cadena: ""},
		{
			nombre: "tras el terminador, lo que parece una bandera",
			args:   []string{"buscar", "--", "--json", "--json"}, valor: "--json", cadena: "--json",
		},
		{
			nombre: "con una bandera global detrás",
			args:   []string{"buscar", "a\xffb", "a\xffb", "--json"}, valor: "a\xffb", cadena: "a\xef\xbf\xbdb", json: true,
		},
	} {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			res := analizarConLiteral(t, caso.args...)

			require.NoError(t, res.err)
			assert.Equal(t, DecisionEjecutar, res.analisis.Decision)
			assert.Equal(t, "buscar", res.analisis.Verbo)
			assert.Equal(t, caso.json, res.analisis.Globales.JSON)
			assert.Equal(t, Literal(caso.valor), res.verbos.Buscar.ID, "el Literal conserva los bytes")
			assert.Equal(t, caso.cadena, res.verbos.Buscar.Cadena, "premisa: el string los recibe cambiados")
			assert.Empty(t, res.doble.salida.String())
			assert.Empty(t, res.doble.errores.String())
		})
	}
}

// TestAnalizarLiteralAusente comprueba que un Literal obligatorio que falta es
// un error de argumentos que lo nombra, como un string (FR-027).
func TestAnalizarLiteralAusente(t *testing.T) {
	t.Parallel()

	res := analizarConLiteral(t, "buscar")

	require.ErrorIs(t, res.err, ErrArgumentos)
	assert.Equal(t, schema.ClaseArgumentos, Clasificar(res.err))
	assert.Contains(t, res.err.Error(), "<id>")
	assert.Equal(t, Analisis{}, res.analisis)
}

// TestLiteralDecode comprueba las dos formas en que Decode no guarda nada: lo
// siguiente no es un valor —una bandera, con las reglas del decodificador de
// las cadenas— o es un valor que no es un texto, que solo una gramática que
// declara otra cosa podría dar. El Literal queda como estaba.
func TestLiteralDecode(t *testing.T) {
	t.Parallel()

	for _, caso := range []struct {
		nombre  string
		escaner *kong.Scanner
		nombra  string
	}{
		{nombre: "una bandera no es un valor", escaner: kong.Scan("--json"), nombra: "--json"},
		{
			nombre:  "un valor que no es un texto",
			escaner: kong.ScanFromTokens(kong.Token{Type: kong.FlagValueToken, Value: 3}),
			nombra:  "3 (int)",
		},
	} {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			literal := Literal("previo")

			err := literal.Decode(&kong.DecodeContext{Scan: caso.escaner})

			require.Error(t, err)
			assert.Contains(t, err.Error(), caso.nombra)
			assert.Equal(t, Literal("previo"), literal)
		})
	}
}
