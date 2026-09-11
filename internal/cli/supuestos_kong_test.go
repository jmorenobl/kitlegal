package cli

import (
	"bytes"
	"testing"
	"time"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// gramaticaDeSupuestos es una gramática mínima que existe solo para ejercer los
// cinco supuestos de research.md D24 sobre la biblioteca de análisis de la línea
// de órdenes. No adelanta nada de las banderas globales del kernel, que son
// alcance de otra tarea: sus cuatro banderas están elegidas para que cada una
// cubra una capacidad de etiqueta distinta del supuesto S4 —ayuda, valor por
// omisión sobre una duración, nombre explícito con variable de entorno y forma
// corta— y el mandato con argumento obligatorio para que el análisis tenga algo
// que resolver.
type gramaticaDeSupuestos struct {
	JSON    bool          `help:"Emite el sobre en JSON."`
	Timeout time.Duration `help:"Plazo total de la operación." default:"30s"`
	Asunto  string        `help:"Directorio del asunto." name:"asunto" env:"KITLEGAL_ASUNTO_DE_PRUEBA"`
	Verbose bool          `short:"v" help:"Sube el detalle del registro de eventos."`

	Repetir struct {
		Texto string `arg:"" help:"Lo que se repite."`
	} `cmd:"" help:"Repite lo que se le da."`
}

// analisis recoge todo lo observable de una llamada al análisis: lo que escribió
// en cada uno de los dos escritores que se le inyectaron, los códigos con los
// que pidió terminar el proceso —que nadie atiende— y el error que devolvió.
type analisis struct {
	analizador    *kong.Kong
	gramatica     gramaticaDeSupuestos
	salida        *bytes.Buffer
	errores       *bytes.Buffer
	terminaciones []int
	err           error
}

// analizar monta el analizador con los escritores y la función de terminación
// sustituidos —las dos sustituciones que afirman S2 y S3— y le entrega los
// argumentos ya separados del nombre del programa, que es lo que afirma S1.
//
// Que esta función retorne es, por sí misma, la comprobación de que el análisis
// devuelve el control: si algún camino de la biblioteca terminara el proceso por
// su cuenta, el test no llegaría a ninguna aserción.
func analizar(t *testing.T, args ...string) *analisis {
	t.Helper()

	resultado := &analisis{salida: &bytes.Buffer{}, errores: &bytes.Buffer{}}

	analizador, err := kong.New(&resultado.gramatica,
		kong.Name("kitlegal"),
		kong.Writers(resultado.salida, resultado.errores),
		kong.Exit(func(codigo int) {
			resultado.terminaciones = append(resultado.terminaciones, codigo)
		}),
	)
	require.NoError(t, err)

	resultado.analizador = analizador
	_, resultado.err = analizador.Parse(args)

	return resultado
}

// TestSupuestosDeKong comprueba uno por uno los cinco supuestos que el plan no
// pudo verificar sin red (research.md D24). El resultado, la versión resuelta y
// la contingencia aplicada —ninguna— quedan por escrito en
// specs/002-h1-kernel-cli-multicall/gates/supuestos-kong.md.
func TestSupuestosDeKong(t *testing.T) {
	t.Parallel()

	t.Run("S1: se construye con una gramática y se analiza una lista de argumentos", func(t *testing.T) {
		t.Parallel()

		res := analizar(t, "repetir", "hola", "--timeout=1500ms")

		require.NoError(t, res.err)
		assert.Equal(t, "hola", res.gramatica.Repetir.Texto)
		assert.Equal(t, 1500*time.Millisecond, res.gramatica.Timeout)
		assert.Empty(t, res.terminaciones,
			"un análisis correcto no pide terminar el proceso")
	})

	t.Run("S2: los dos escritores son sustituibles y reciben lo que se escribe", func(t *testing.T) {
		t.Parallel()

		res := analizar(t, "--help")

		assert.Same(t, res.salida, res.analizador.Stdout)
		assert.Same(t, res.errores, res.analizador.Stderr)
		assert.Contains(t, res.salida.String(), "Usage: kitlegal",
			"la ayuda integrada debe salir por el escritor inyectado")
		assert.Empty(t, res.errores.String())

		// Sin la opción que los sustituye, los escritores son los descriptores
		// del proceso: que los inyectados sean otros es lo que demuestra que
		// nada de lo anterior llegó a ellos.
		predeterminado, err := kong.New(&gramaticaDeSupuestos{}, kong.Name("kitlegal"))
		require.NoError(t, err)
		assert.NotSame(t, predeterminado.Stdout, res.analizador.Stdout)
		assert.NotSame(t, predeterminado.Stderr, res.analizador.Stderr)
	})

	t.Run("S3: la ayuda integrada devuelve el control en lugar de terminar el proceso", func(t *testing.T) {
		t.Parallel()

		res := analizar(t, "--help")

		// El supuesto bloqueante: la biblioteca pide terminar con 0, la función
		// sustituida lo anota y el análisis retorna. Ningún camino llama a la
		// terminación del proceso por su cuenta.
		assert.Equal(t, []int{0}, res.terminaciones)
		assert.NotEmpty(t, res.salida.String())

		// Consecuencia observada que el kernel tendrá que atender: tras imprimir
		// la ayuda el análisis sigue y acaba devolviendo el error de la
		// gramática incompleta. La ayuda se pidió de todos modos, así que la
		// precedencia de --help no puede leerse del error devuelto.
		require.Error(t, res.err)
		var fallo *kong.ParseError
		require.ErrorAs(t, res.err, &fallo)
	})

	t.Run("S3: la ayuda de un mandato tampoco termina el proceso", func(t *testing.T) {
		t.Parallel()

		res := analizar(t, "repetir", "--help")

		assert.Equal(t, []int{0}, res.terminaciones)
		assert.Contains(t, res.salida.String(), "repetir")
		assert.Empty(t, res.errores.String())
	})

	t.Run("S4: las etiquetas cubren ayuda, omisión, nombre, forma corta y entorno", func(t *testing.T) {
		t.Parallel()

		res := analizar(t, "repetir", "hola")
		require.NoError(t, res.err)

		banderas := make(map[string]*kong.Flag, len(res.analizador.Model.Flags))
		for _, bandera := range res.analizador.Model.Flags {
			banderas[bandera.Name] = bandera
		}

		require.Contains(t, banderas, "json")
		assert.Equal(t, "Emite el sobre en JSON.", banderas["json"].Help)

		require.Contains(t, banderas, "timeout")
		assert.True(t, banderas["timeout"].HasDefault)
		assert.Equal(t, "30s", banderas["timeout"].Default)

		require.Contains(t, banderas, "asunto")
		assert.Equal(t, []string{"KITLEGAL_ASUNTO_DE_PRUEBA"}, banderas["asunto"].Envs)

		require.Contains(t, banderas, "verbose")
		assert.Equal(t, 'v', banderas["verbose"].Short)
	})

	t.Run("S4: una duración se analiza nativamente y un valor inválido es un error", func(t *testing.T) {
		t.Parallel()

		conOmision := analizar(t, "repetir", "hola")
		require.NoError(t, conOmision.err)
		assert.Equal(t, 30*time.Second, conOmision.gramatica.Timeout,
			"el valor por omisión de la etiqueta se aplica sin convertir a mano")

		invalido := analizar(t, "repetir", "hola", "--timeout=abc")
		require.Error(t, invalido.err)
		assert.Contains(t, invalido.err.Error(), "--timeout")
		assert.Empty(t, invalido.terminaciones,
			"un valor inválido se devuelve como error, no termina el proceso")
	})

	t.Run("S5: una bandera desconocida es un error distinguible y no escribe nada", func(t *testing.T) {
		t.Parallel()

		res := analizar(t, "--bandera-que-nadie-ha-declarado")

		require.Error(t, res.err)

		var fallo *kong.ParseError
		require.ErrorAs(t, res.err, &fallo,
			"el fallo del análisis debe ser distinguible por tipo para traducirlo a ErrArgumentos")
		assert.Contains(t, res.err.Error(), "--bandera-que-nadie-ha-declarado",
			"el mensaje debe nombrar el problema concreto (FR-027)")

		assert.Empty(t, res.terminaciones,
			"un error de argumentos se devuelve; quien decide el código de salida es el kernel")
		assert.Empty(t, res.salida.String())
		assert.Empty(t, res.errores.String(),
			"la biblioteca no imprime el error por su cuenta: lo imprime el presentador")
	})

	t.Run("S5: un mandato desconocido también es un error devuelto", func(t *testing.T) {
		t.Parallel()

		res := analizar(t, "mandato-que-nadie-ha-declarado")

		require.Error(t, res.err)
		var fallo *kong.ParseError
		require.ErrorAs(t, res.err, &fallo)
		assert.Empty(t, res.terminaciones)
		assert.Empty(t, res.salida.String())
		assert.Empty(t, res.errores.String())
	})
}
