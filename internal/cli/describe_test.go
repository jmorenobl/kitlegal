package cli

import (
	"fmt"
	"sort"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// urlDelEsquemaGenerado identifica el documento que emite --describe ante el
// compilador. No es la del contrato: lo que se compila aquí es el esquema
// **generado**, y confundir los dos recursos ocultaría justo la diferencia que
// este test existe para comprobar.
const urlDelEsquemaGenerado = "https://ventanillalegal.es/schemas/describe.json"

// rutaDeLaSalida es el puntero JSON de la parte del documento que describe el
// sobre. Se compila por separado —y no extrayéndola a otro documento— porque sus
// referencias apuntan a los `$defs` de la raíz: sacarla de ahí las rompería, y un
// consumidor real (una tool MCP, H22) la usa igual, dentro del documento.
const rutaDeLaSalida = "#/properties/salida"

// borrador2020 es el del contrato, escrito aquí y no leído de la biblioteca que
// lo genera: si mañana esa biblioteca emitiera otro borrador, el contrato seguiría
// diciendo este y el test lo notaría.
const borrador2020 = "https://json-schema.org/draft/2020-12/schema"

// Los dos tipos de `data` de los verbos de las tablas. Son lo que un applet
// declara en Verbo.Salida, y están escritos con la única finalidad de que el
// esquema tenga algo concreto que describir: que sean dos distintos es lo que
// permite comprobar FR-048 —el esquema sale de la definición— sin editar nada.
type (
	datosDelEco struct {
		Mensaje string `json:"mensaje"`
		Veces   int    `json:"veces"`
	}

	datosDeCuenta struct {
		Cuenta int `json:"cuenta"`
	}
)

// verboDePrueba es la definición del verbo `repetir` del applet de las
// tablas: la misma que el paquete de composición construirá desde su registro.
//
// Sus datos de salida casan con los de resultadoDePrueba, que es lo que hace que
// el sobre que monta el kernel y el esquema que describe ese verbo hablen del
// mismo verbo y no de dos cosas parecidas (SC-015).
func verboDePrueba() Verbo {
	return Verbo{
		Applet:     nombreDePrueba,
		Verbo:      "repetir",
		Ayuda:      "Repite lo que se le da.",
		Argumentos: &verboRepetir{},
		Salida:     datosDelEco{},
	}
}

// describirDePrueba ejecuta la autodescripción y devuelve el documento emitido
// junto con el presentador, para poder afirmar además lo que **no** se escribió.
func describirDePrueba(t *testing.T, def Verbo) (map[string]any, *presentadorConJSON) {
	t.Helper()

	doble := &presentadorConJSON{}
	require.NoError(t, Describir(doble, def))

	return unicoDocumento(t, doble.salida.String()), doble
}

// compilarGenerado compila el esquema emitido —entero o una de sus partes— con
// las aserciones de formato activadas. Que compile es la mitad de SC-007: un
// esquema que un validador no acepta no describe nada.
func compilarGenerado(t *testing.T, documento map[string]any, ruta string) *jsonschema.Schema {
	t.Helper()

	compilador := jsonschema.NewCompiler()
	compilador.AssertFormat()
	require.NoError(t, compilador.AddResource(urlDelEsquemaGenerado, documento))

	compilado, err := compilador.Compile(urlDelEsquemaGenerado + ruta)
	require.NoError(t, err)

	return compilado
}

// bajar desciende por el documento hasta el objeto que nombra la ruta. Falla con
// la ruta recorrida en lugar de con un pánico de conversión, para que un esquema
// con otra forma diga cuál es la clave que no está donde se esperaba.
func bajar(t *testing.T, documento map[string]any, ruta ...string) map[string]any {
	t.Helper()

	actual := documento
	for i, clave := range ruta {
		valor, existe := actual[clave]
		require.True(t, existe, "el documento no tiene %v", ruta[:i+1])

		objeto, esObjeto := valor.(map[string]any)
		require.True(t, esObjeto, "%v no es un objeto", ruta[:i+1])

		actual = objeto
	}

	return actual
}

// claves son las de un objeto JSON, ordenadas, que es como se comparan con la
// lista que promete el contrato.
func claves(objeto map[string]any) []string {
	nombres := make([]string, 0, len(objeto))
	for nombre := range objeto {
		nombres = append(nombres, nombre)
	}
	sort.Strings(nombres)

	return nombres
}

// sobreEmitido monta y escribe un sobre real con el montador de las tablas y
// devuelve el documento. Los sobres de este test no se escriben a mano a
// propósito: lo que SC-015 exige es que lo que el kernel emite de verdad valide
// contra el esquema que el mismo verbo describe.
func sobreEmitido(t *testing.T, res schema.Resultado, err error) map[string]any {
	t.Helper()

	doble := &presentadorConJSON{}
	emitirDePrueba(doble, true, res, err)

	return unicoDocumento(t, doble.salida.String())
}

// sobreDeFalloConTraza es un sobre de fallo que incumple el contrato en lo único
// que la condición vigila: su `data` lleva una tercera clave. El detalle
// técnico va al registro de eventos y nunca al sobre (FR-045, research.md D7), y
// esta es la desviación que la rama `else` rechaza.
func sobreDeFalloConTraza(t *testing.T) map[string]any {
	t.Helper()

	documento := sobreEmitido(t, schema.Resultado{},
		fmt.Errorf("el bloque a99: %w", ErrNoEncontrado))

	datos, esObjeto := documento["data"].(map[string]any)
	require.True(t, esObjeto)
	datos["traza"] = "cli.Emitir(...)"

	return documento
}

// sinRamaElse devuelve el mismo documento sin la rama `else` de la condición. Es
// la mutación con la que se comprueba que esa rama es la que sostiene el caso del
// sobre de fallo: si el test siguiera en verde sin ella, no estaría comprobando
// nada.
func sinRamaElse(t *testing.T, documento map[string]any) map[string]any {
	t.Helper()

	salida := bajar(t, documento, "properties", "salida")
	require.Contains(t, salida, "else")
	delete(salida, "else")

	return documento
}

// TestDescribe comprueba la autodescripción del applet: que --describe emite un
// único esquema del borrador 2020-12 que un validador acepta, que describe la
// entrada y la salida, que el sobre de éxito y el de fallo de ese mismo verbo
// validan contra él y que el `data` está condicionado a `ok` de verdad
// (FR-046 … FR-049, SC-007, SC-015).
//
// Es el test que invoca el escenario 6 de quickstart.md por su nombre.
func TestDescribe(t *testing.T) {
	t.Parallel()

	t.Run("un esquema válido del borrador 2020-12 con las dos partes", func(t *testing.T) {
		t.Parallel()

		documento, _ := describirDePrueba(t, verboDePrueba())

		assert.Equal(t, borrador2020, documento["$schema"])
		assert.Equal(t, []string{"entrada", "salida"},
			claves(bajar(t, documento, "properties")))
		assert.Equal(t, nombreDePrueba+" repetir", documento["title"])
		assert.Equal(t, verboDePrueba().Ayuda, documento["description"])

		assert.NotContains(t, documento, "$id",
			"un $id rebasaría las referencias internas contra otra base")

		require.NotNil(t, compilarGenerado(t, documento, ""))
	})

	t.Run("la entrada lleva los argumentos del verbo y las ocho banderas", func(t *testing.T) {
		t.Parallel()

		documento, _ := describirDePrueba(t, verboDePrueba())
		entrada := bajar(t, documento, "properties", "entrada")

		assert.ElementsMatch(t, append([]string{"texto"}, nombresDeLasOcho...),
			claves(bajar(t, entrada, "properties")),
			"los argumentos del verbo más las ocho banderas globales, y nada más")
		assert.Equal(t, false, entrada["additionalProperties"],
			"la entrada es cerrada: una bandera que no existe no se describe")
		assert.Equal(t, []any{"texto"}, entrada["required"],
			"solo el argumento de posición es obligatorio; las banderas tienen omisión")

		propiedades := bajar(t, entrada, "properties")
		assert.Equal(t, "string", bajar(t, propiedades, "texto")["type"])
		assert.Equal(t, "boolean", bajar(t, propiedades, "json")["type"])
		assert.Equal(t, "string", bajar(t, propiedades, "timeout")["type"],
			"--timeout se escribe «30s»: describirlo como el entero de su tipo Go "+
				"describiría un valor que la invocación no acepta")
	})

	t.Run("los nombres de la entrada son los que acepta la gramática", func(t *testing.T) {
		t.Parallel()

		documento, _ := describirDePrueba(t, verboDePrueba())
		propiedades := bajar(t, documento, "properties", "entrada", "properties")

		// Kong es quien decide cómo se escribe cada bandera en la invocación. Que
		// el esquema diga exactamente eso, y no una lista paralela, es lo que
		// hace cierto que describe lo que el binario acepta (FR-048).
		for nombre := range modeloDeGlobales(t) {
			if nombre == banderaModeloAyuda {
				continue
			}

			assert.Contains(t, propiedades, nombre,
				"la bandera %s existe en la gramática y no está descrita", nombre)
		}
	})

	t.Run("la salida describe el sobre completo", func(t *testing.T) {
		t.Parallel()

		documento, _ := describirDePrueba(t, verboDePrueba())
		salida := bajar(t, documento, "properties", "salida")

		assert.Equal(t, "object", salida["type"])
		assert.Equal(t, false, salida["additionalProperties"])
		assert.NotContains(t, salida, "$schema",
			"el $schema va una sola vez, en la raíz del documento")

		require.Len(t, salida["required"], len(clavesDelSobre))
		for _, clave := range clavesDelSobre {
			assert.Contains(t, salida["required"], clave)
			assert.Contains(t, bajar(t, salida, "properties"), clave)
		}

		assert.Equal(t, true, bajar(t, salida, "if", "properties", "ok")["const"],
			"la condición se decide por `ok`")
		assert.Equal(t, "#/$defs/datosDelEco",
			bajar(t, salida, "then", "properties", "data")["$ref"],
			"con ok verdadero, `data` es el del applet")
		assert.Equal(t, "#/$defs/DatosError",
			bajar(t, salida, "else", "properties", "data")["$ref"],
			"con ok falso, `data` es la forma común de error del kernel")

		definiciones := bajar(t, documento, "$defs")
		assert.Contains(t, definiciones, "DatosError")
		assert.Contains(t, definiciones, "datosDelEco")
	})

	t.Run("el sobre de éxito de ese verbo valida", func(t *testing.T) {
		t.Parallel()

		documento, _ := describirDePrueba(t, verboDePrueba())
		compilado := compilarGenerado(t, documento, rutaDeLaSalida)

		require.NoError(t, compilado.Validate(sobreEmitido(t, resultadoDePrueba(), nil)))

		otroData := schema.Resultado{
			Procedencia: procedenciaDelApplet,
			Datos:       map[string]any{"mensaje": "hola"},
		}
		require.Error(t, compilado.Validate(sobreEmitido(t, otroData, nil)),
			"con ok verdadero, un `data` que no es el del applet no vale: "+
				"la rama `then` también comprueba algo")
	})

	t.Run("el sobre de fallo de ese mismo verbo valida", func(t *testing.T) {
		t.Parallel()

		documento, _ := describirDePrueba(t, verboDePrueba())
		compilado := compilarGenerado(t, documento, rutaDeLaSalida)

		for _, caso := range casosDeFallo(t) {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				sobre := sobreEmitido(t, schema.Resultado{Procedencia: caso.delApplet}, caso.err)
				assert.Equal(t, false, sobre["ok"])
				require.NoError(t, compilado.Validate(sobre),
					"un test de contrato sobre una ejecución fallida no puede "+
						"rechazar una salida conforme (FR-047)")
			})
		}

		require.Error(t, compilado.Validate(sobreDeFalloConTraza(t)),
			"con ok falso, `data` es exactamente {clase, mensaje}")
	})

	t.Run("sin la rama else el caso del sobre de fallo deja de comprobarse", func(t *testing.T) {
		t.Parallel()

		documento, _ := describirDePrueba(t, verboDePrueba())
		mutilado := compilarGenerado(t, sinRamaElse(t, documento), rutaDeLaSalida)

		require.NoError(t, mutilado.Validate(sobreDeFalloConTraza(t)),
			"quitada la rama `else`, el `data` de un sobre de fallo deja de "+
				"describirse y el caso que lo comprueba pasa a no comprobar nada")
	})

	t.Run("el esquema sale de la definición y no de una copia", func(t *testing.T) {
		t.Parallel()

		otro := Verbo{
			Applet:     nombreDePrueba,
			Verbo:      "contar",
			Ayuda:      "Cuenta hasta donde se le diga.",
			Argumentos: &verboContar{},
			Salida:     datosDeCuenta{},
		}

		documento, _ := describirDePrueba(t, otro)

		assert.Equal(t, nombreDePrueba+" contar", documento["title"])
		assert.ElementsMatch(t, append([]string{"veces"}, nombresDeLasOcho...),
			claves(bajar(t, documento, "properties", "entrada", "properties")),
			"otro verbo, otros argumentos, sin editar nada")
		assert.Equal(t, "integer",
			bajar(t, documento, "properties", "entrada", "properties", "veces")["type"])
		assert.Equal(t, "#/$defs/datosDeCuenta",
			bajar(t, documento, "properties", "salida", "then", "properties", "data")["$ref"])
		assert.NotContains(t, bajar(t, documento, "$defs"), "datosDelEco")
	})

	t.Run("un verbo sin argumentos describe solo las ocho banderas", func(t *testing.T) {
		t.Parallel()

		def := verboDePrueba()
		def.Argumentos = nil

		documento, _ := describirDePrueba(t, def)
		entrada := bajar(t, documento, "properties", "entrada")

		assert.ElementsMatch(t, nombresDeLasOcho, claves(bajar(t, entrada, "properties")))
		assert.NotContains(t, entrada, "required", "ninguna de las ocho es obligatoria")
	})

	t.Run("un verbo sin salida declarada deja `data` sin restringir", func(t *testing.T) {
		t.Parallel()

		def := verboDePrueba()
		def.Salida = nil

		documento, _ := describirDePrueba(t, def)
		salida := bajar(t, documento, "properties", "salida")

		assert.Equal(t, true, bajar(t, salida, "then", "properties")["data"],
			"sin tipo declarado no se inventa ninguno")
		assert.Equal(t, "#/$defs/DatosError",
			bajar(t, salida, "else", "properties", "data")["$ref"],
			"el fallo se describe igual: no depende de lo que declare el applet")
	})

	t.Run("describirse excluye ejecutar", func(t *testing.T) {
		t.Parallel()

		_, doble := describirDePrueba(t, verboDePrueba())

		assert.Empty(t, doble.sobres, "la salida no contiene ningún sobre")
		assert.Len(t, doble.textos, 1, "un solo documento en la salida estándar")
		assert.Empty(t, doble.avisos)
		assert.Empty(t, doble.errores.String(), "la salida de error queda intacta")
	})

	t.Run("una escritura fallida se propaga y no se traga", func(t *testing.T) {
		t.Parallel()

		doble := dobleRoto()
		err := Describir(doble, verboDePrueba())

		require.ErrorIs(t, err, errEscrituraRota)
		assert.Equal(t, 1, CodigoSalida(err), "un descriptor roto es un fallo inesperado")
	})

	t.Run("unos argumentos que no son un struct son un defecto del applet", func(t *testing.T) {
		t.Parallel()

		def := verboDePrueba()
		def.Argumentos = "no soy un struct"

		doble := &presentadorConJSON{}
		err := Describir(doble, def)

		require.ErrorIs(t, err, errEsquemaImposible)
		assert.Equal(t, 1, CodigoSalida(err),
			"lo que ningún usuario puede provocar no es un error de argumentos")
		assert.Empty(t, doble.salida.String(), "no se escribe un esquema a medias")
	})
}
