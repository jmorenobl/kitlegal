package cli

import (
	"reflect"
	"testing"
	"time"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// banderaModeloAyuda es el nombre con el que la ayuda integrada de Kong aparece
// en el modelo. Es la única bandera de la gramática común que no declara el
// kernel, y por eso se nombra aquí: para poder afirmar que no hay ninguna otra.
const banderaModeloAyuda = "help"

// nombresDeLasOcho son los nombres largos exactos que el contrato promete
// (contracts/banderas-y-exit-codes.md §1), escritos aquí sin derivarlos del
// código que se comprueba: si alguien renombra un campo de Globales o cambia una
// etiqueta, el contrato sigue diciendo lo mismo y el test falla.
var nombresDeLasOcho = []string{
	"json",
	"timeout",
	"offline",
	"dry-run",
	"describe",
	"no-graph",
	"asunto",
	"verbose",
}

// modeloDeGlobales construye el modelo de Kong de una gramática que no contiene
// nada más que las ocho globales embebidas. Es la forma de comprobar lo que Kong
// entiende de la declaración —nombres, valores por omisión y textos de ayuda—
// sin analizar ninguna invocación.
//
// Se le entregan los escritores del doble y la terminación sustituida por la
// misma razón que al analizador del kernel: ningún camino de esta biblioteca
// puede escribir en los descriptores del proceso ni terminar el proceso
// (gates/supuestos-kong.md, S2 y S3).
func modeloDeGlobales(t *testing.T) map[string]*kong.Flag {
	t.Helper()

	var gramatica struct{ Globales }

	doble := &presentadorDoble{}
	analizador, err := kong.New(&gramatica,
		kong.Name("kitlegal"),
		kong.Writers(doble.Salida(), doble.Error()),
		kong.Exit(func(int) { t.Error("la construcción del modelo no puede terminar el proceso") }),
	)
	require.NoError(t, err)

	banderas := make(map[string]*kong.Flag, len(analizador.Model.Flags))
	for _, bandera := range analizador.Model.Flags {
		banderas[bandera.Name] = bandera
	}

	return banderas
}

// TestGlobales comprueba la declaración de las ocho banderas globales: que son
// ocho, que se llaman como promete el contrato, que solo las acompaña la ayuda
// integrada de Kong y que las etiquetas que necesitan valor por omisión y texto
// de ayuda lo tienen (FR-018 … FR-021, FR-023, FR-024, FR-026).
func TestGlobales(t *testing.T) {
	t.Parallel()

	t.Run("el struct declara exactamente ocho banderas", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, len(nombresDeLasOcho), reflect.TypeFor[Globales]().NumField(),
			"las globales son ocho: una novena no se declara sin ampliar el contrato")
	})

	t.Run("las ocho tienen el nombre largo que promete el contrato", func(t *testing.T) {
		t.Parallel()

		banderas := modeloDeGlobales(t)

		for _, nombre := range nombresDeLasOcho {
			assert.Contains(t, banderas, nombre)
		}
	})

	t.Run("lo único que acompaña a las ocho es la ayuda integrada de Kong", func(t *testing.T) {
		t.Parallel()

		banderas := modeloDeGlobales(t)

		esperadas := append([]string{banderaModeloAyuda}, nombresDeLasOcho...)
		nombres := make([]string, 0, len(banderas))
		for nombre := range banderas {
			nombres = append(nombres, nombre)
		}

		assert.ElementsMatch(t, esperadas, nombres,
			"una bandera de más en la gramática común la heredarían todos los applets sin que nadie la haya declarado")
	})

	t.Run("las ocho llevan texto de ayuda para la persona", func(t *testing.T) {
		t.Parallel()

		banderas := modeloDeGlobales(t)

		for _, nombre := range nombresDeLasOcho {
			require.Contains(t, banderas, nombre)
			assert.NotEmpty(t, banderas[nombre].Help,
				"la ayuda de %s es lo único que explica la bandera a quien invoca", nombre)
		}
	})

	t.Run("solo --timeout tiene valor por omisión, y es 30s", func(t *testing.T) {
		t.Parallel()

		banderas := modeloDeGlobales(t)

		for _, nombre := range nombresDeLasOcho {
			require.Contains(t, banderas, nombre)

			if nombre == "timeout" {
				assert.True(t, banderas[nombre].HasDefault)
				assert.Equal(t, timeoutPorOmision.String(), banderas[nombre].Default,
					"la etiqueta y la constante nombran el mismo plazo")

				continue
			}

			assert.False(t, banderas[nombre].HasDefault,
				"%s vale lo que valga el cero de su tipo; un valor por omisión sería otra semántica", nombre)
		}
	})
}

// casoDeContexto es una fila de la conversión: unas globales ya analizadas y el
// contexto de ejecución que el applet debe recibir de ellas.
type casoDeContexto struct {
	nombre   string
	globales Globales
	esperado schema.Contexto
}

// casosDeContexto cubre las seis opciones que viajan al applet y las dos que
// no: --describe, que excluye la ejecución y por tanto el applet nunca ve, y
// --verbose, que solo fija el nivel del registro de eventos (data-model.md §8).
func casosDeContexto() []casoDeContexto {
	return []casoDeContexto{
		{
			nombre:   "las globales en su valor cero, salvo el plazo",
			globales: Globales{Timeout: timeoutPorOmision},
			esperado: schema.Contexto{Timeout: timeoutPorOmision},
		},
		{
			nombre:   "--json",
			globales: Globales{JSON: true, Timeout: timeoutPorOmision},
			esperado: schema.Contexto{JSON: true, Timeout: timeoutPorOmision},
		},
		{
			nombre:   "--timeout",
			globales: Globales{Timeout: 90 * time.Second},
			esperado: schema.Contexto{Timeout: 90 * time.Second},
		},
		{
			nombre:   "--offline",
			globales: Globales{Offline: true, Timeout: timeoutPorOmision},
			esperado: schema.Contexto{Offline: true, Timeout: timeoutPorOmision},
		},
		{
			nombre:   "--dry-run",
			globales: Globales{DryRun: true, Timeout: timeoutPorOmision},
			esperado: schema.Contexto{DryRun: true, Timeout: timeoutPorOmision},
		},
		{
			nombre:   "--no-graph",
			globales: Globales{SinGrafo: true, Timeout: timeoutPorOmision},
			esperado: schema.Contexto{SinGrafo: true, Timeout: timeoutPorOmision},
		},
		{
			nombre:   "--asunto",
			globales: Globales{Asunto: "expediente-3", Timeout: timeoutPorOmision},
			esperado: schema.Contexto{Asunto: "expediente-3", Timeout: timeoutPorOmision},
		},
		{
			nombre:   "--describe no llega al contexto: excluye la ejecución",
			globales: Globales{Describe: true, Timeout: timeoutPorOmision},
			esperado: schema.Contexto{Timeout: timeoutPorOmision},
		},
		{
			nombre:   "--verbose tampoco: solo fija el nivel del registro",
			globales: Globales{Verbose: true, Timeout: timeoutPorOmision},
			esperado: schema.Contexto{Timeout: timeoutPorOmision},
		},
		{
			nombre: "las seis que viajan, a la vez",
			globales: Globales{
				JSON:     true,
				Timeout:  2 * time.Minute,
				Offline:  true,
				DryRun:   true,
				Describe: true,
				SinGrafo: true,
				Asunto:   "expediente-3",
				Verbose:  true,
			},
			esperado: schema.Contexto{
				JSON:     true,
				Timeout:  2 * time.Minute,
				Offline:  true,
				DryRun:   true,
				SinGrafo: true,
				Asunto:   "expediente-3",
			},
		},
	}
}

// TestGlobalesContexto comprueba la conversión de las banderas al contexto de
// ejecución: las seis opciones que viajan al applet y las dos que se quedan en
// el kernel (FR-018, FR-021, FR-023, FR-024, FR-049 parcial).
func TestGlobalesContexto(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeContexto() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.esperado, caso.globales.Contexto())
		})
	}
}
