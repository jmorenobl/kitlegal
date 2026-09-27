package cli

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"

	// El alias distingue las dos bibliotecas de esquemas que este test necesita a
	// la vez y que se llaman igual: `invopop` es la que **genera** el documento
	// —la misma que usa describe.go— y `jsonschema`, sin alias, la que lo
	// **compila** para comprobar que un validador lo acepta.
	invopop "github.com/invopop/jsonschema"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// urlDelEsquemaGenerado identifica el documento que emite --describe ante el
// compilador. No es la del contrato: lo que se compila aquí es el esquema
// **generado**, y confundir los dos recursos ocultaría justo la diferencia que
// este test existe para comprobar.
const urlDelEsquemaGenerado = "https://kitlegal.es/schemas/describe.json"

// rutaDeLaSalida es el puntero JSON de la parte del documento que describe el
// sobre. Se compila por separado —y no extrayéndola a otro documento— porque sus
// referencias apuntan a los `$defs` de la raíz: sacarla de ahí las rompería, y un
// consumidor real (una tool MCP, H28) la usa igual, dentro del documento.
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

// Los nombres con los que esos dos tipos aparecen en los `$defs` del
// documento: cualificados con el paquete, porque no son tipos del sobre sino
// del applet, y así no pueden pisar los del contrato.
const (
	definicionDelEco    = "cli.datosDelEco"
	definicionDeCuenta  = "cli.datosDeCuenta"
	definicionDeError   = "DatosError"
	definicionHomonima  = "cli.DatosError"
	definicionDeSobre   = "Sobre"
	prefijoDeDefinicion = "#/$defs/"
)

// DatosError es un tipo de `data` que un applet podría declarar con el mismo
// nombre que el del kernel y otra forma. Existe para comprobar que el documento
// no los confunde: sin cualificar los nombres, la biblioteca haría que las dos
// ramas de la condición apuntaran al mismo `$defs` y el esquema rechazaría el
// sobre de éxito correcto de ese verbo.
type DatosError struct {
	Codigo int `json:"codigo"`
}

// homonimoUno y homonimoDos son dos tipos **distintos** que se llaman igual y
// viven en el mismo paquete —solo un tipo local puede hacerlo—, de modo que ni
// cualificar el nombre los separa. Son la única forma de provocar, sin inventar
// paquetes, el choque de nombres que el documento tiene que rechazar en lugar de
// describir otra cosa.
func homonimoUno() reflect.Type {
	type Homonimo struct {
		Uno int `json:"uno"`
	}

	return reflect.TypeOf(Homonimo{})
}

func homonimoDos() reflect.Type {
	type Homonimo struct {
		Dos int `json:"dos"`
	}

	return reflect.TypeOf(Homonimo{})
}

// salidaConHomonimos es un `data` que contiene un campo de cada homónimo.
func salidaConHomonimos() any {
	tipo := reflect.StructOf([]reflect.StructField{
		{Name: "Uno", Type: homonimoUno(), Tag: `json:"uno"`},
		{Name: "Dos", Type: homonimoDos(), Tag: `json:"dos"`},
	})

	return reflect.New(tipo).Elem().Interface()
}

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

// datosDe devuelve el `data` de un sobre como objeto, para poder alterarlo.
func datosDe(t *testing.T, sobre map[string]any) map[string]any {
	t.Helper()

	datos, esObjeto := sobre["data"].(map[string]any)
	require.True(t, esObjeto, "el data de este sobre es un objeto")

	return datos
}

// referencia devuelve el `$ref` con el que una rama de la condición describe
// `data`.
func referencia(t *testing.T, salida map[string]any, rama string) any {
	t.Helper()

	return bajar(t, salida, rama, "properties", "data")["$ref"]
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
		assert.Equal(t, prefijoDeDefinicion+definicionDelEco, referencia(t, salida, "then"),
			"con ok verdadero, `data` es el del applet, nombrado con su paquete")
		assert.Equal(t, prefijoDeDefinicion+definicionDeError, referencia(t, salida, "else"),
			"con ok falso, `data` es la forma común de error del kernel, nombrada como en el contrato")

		definiciones := bajar(t, documento, "$defs")
		assert.Contains(t, definiciones, definicionDeError)
		assert.Contains(t, definiciones, definicionDelEco)
		assert.NotContains(t, definiciones, definicionDeSobre,
			"el sobre va escrito en su sitio y no como una definición referenciada")
	})

	t.Run("la salida lleva las restricciones del contrato, derivadas de los tipos", func(t *testing.T) {
		t.Parallel()

		documento, _ := describirDePrueba(t, verboDePrueba())
		propiedades := bajar(t, documento, "properties", "salida", "properties")

		// Las restricciones salen de las etiquetas del sobre y del vocabulario de
		// clases del dominio, no de una lista escrita en el generador: son las de
		// contracts/sobre-de-salida.md §6, clave por clave (FR-017, FR-048).
		assert.Equal(t, json.Number("1"), bajar(t, propiedades, "fuente")["minLength"],
			"fuente no va vacía")
		assert.Equal(t, json.Number("1"), bajar(t, propiedades, "url")["minLength"],
			"url no va vacía")
		assert.Equal(t, "uri", bajar(t, propiedades, "url")["format"],
			"url es un URI")
		assert.Equal(t, "date-time", bajar(t, propiedades, "fecha_consulta")["format"],
			"fecha_consulta es RFC 3339")
		assert.Equal(t, schema.PatronHuella, bajar(t, propiedades, "hash")["pattern"],
			"hash lleva el prefijo del algoritmo y 64 dígitos hexadecimales")

		datosDeError := bajar(t, documento, "$defs", definicionDeError)
		assert.Equal(t, false, datosDeError["additionalProperties"])
		assert.Equal(t, []any{"clase", "mensaje"}, datosDeError["required"])

		clases := make([]any, 0, len(schema.Clases()))
		for _, clase := range schema.Clases() {
			clases = append(clases, string(clase))
		}

		assert.Equal(t, clases, bajar(t, datosDeError, "properties", "clase")["enum"],
			"clase está dentro del vocabulario del dominio, y en su orden")
		assert.Equal(t, json.Number("1"), bajar(t, datosDeError, "properties", "mensaje")["minLength"],
			"mensaje no va vacío")
	})

	t.Run("el esquema emitido rechaza lo que el contrato prohíbe", func(t *testing.T) {
		t.Parallel()

		documento, _ := describirDePrueba(t, verboDePrueba())
		compilado := compilarGenerado(t, documento, rutaDeLaSalida)

		deExito := func() map[string]any { return sobreEmitido(t, resultadoDePrueba(), nil) }
		deFallo := func() map[string]any {
			return sobreEmitido(t, schema.Resultado{}, fmt.Errorf("el bloque a99: %w", ErrNoEncontrado))
		}

		// Ninguno de estos sobres lo produce el kernel —el dominio y el montador
		// los rechazan antes—, así que se alteran a mano sobre un sobre real: lo
		// que se comprueba es que la descripción formal que el binario publica
		// los rechazaría igualmente, que es la mitad de FR-017 que no depende del
		// código que los emite.
		casos := []struct {
			nombre  string
			sobre   func() map[string]any
			alterar func(sobre map[string]any)
		}{
			{"url vacía", deExito, func(s map[string]any) { s["url"] = "" }},
			{"url que no es un URI", deExito, func(s map[string]any) { s["url"] = "no-es-un-uri" }},
			{"url sin esquema", deExito, func(s map[string]any) { s["url"] = "www.boe.es/buscar" }},
			{"fuente vacía", deExito, func(s map[string]any) { s["fuente"] = "" }},
			{"hash sin el prefijo del algoritmo", deExito, func(s map[string]any) { s["hash"] = "x" }},
			{"hash con el prefijo pero sin los 64 dígitos", deExito, func(s map[string]any) {
				s["hash"] = schema.PrefijoHuella + "abc"
			}},
			{"fecha_consulta que no es RFC 3339", deExito, func(s map[string]any) { s["fecha_consulta"] = "ayer" }},
			{"una séptima clave", deExito, func(s map[string]any) { s["traza"] = "cli.Emitir(...)" }},
			{"clase fuera del vocabulario", deFallo, func(s map[string]any) { datosDe(t, s)["clase"] = "inventada" }},
			{"mensaje vacío", deFallo, func(s map[string]any) { datosDe(t, s)["mensaje"] = "" }},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				sobre := caso.sobre()
				require.NoError(t, compilado.Validate(sobre),
					"el sobre real vale antes de alterarlo: si no, el rechazo no probaría nada")

				caso.alterar(sobre)
				require.Error(t, compilado.Validate(sobre))
			})
		}
	})

	t.Run("el esquema emitido y la copia del contrato describen el mismo sobre", func(t *testing.T) {
		t.Parallel()

		// La copia a mano del contrato (TestContratoSobre) y el esquema generado
		// son dos descripciones del mismo sobre, y tienen que decir lo mismo de
		// sus seis claves y del `data` de fallo: si el generador dejara de
		// derivar una restricción, o el contrato cambiara sin que el sobre lo
		// siguiera, se verían aquí. Lo único que difiere es la rama `then`, que
		// el contrato deja abierta porque el `data` de éxito es de cada applet.
		documento, _ := describirDePrueba(t, verboDePrueba())
		salida := bajar(t, documento, "properties", "salida")
		contrato := unicoDocumento(t, esquemaDelContrato)

		assert.Equal(t, contrato["properties"], salida["properties"])
		assert.Equal(t, contrato["required"], salida["required"])
		assert.Equal(t, contrato["additionalProperties"], salida["additionalProperties"])
		assert.Equal(t, contrato["if"], salida["if"])
		assert.Equal(t, contrato["else"], salida["else"])
		assert.Equal(t, bajar(t, contrato, "$defs")[definicionDeError],
			bajar(t, documento, "$defs")[definicionDeError])
	})

	t.Run("un tipo del applet que se llama como uno del kernel no lo pisa", func(t *testing.T) {
		t.Parallel()

		def := verboDePrueba()
		def.Salida = DatosError{}

		documento, _ := describirDePrueba(t, def)
		salida := bajar(t, documento, "properties", "salida")

		assert.Equal(t, prefijoDeDefinicion+definicionHomonima, referencia(t, salida, "then"))
		assert.Equal(t, prefijoDeDefinicion+definicionDeError, referencia(t, salida, "else"))

		compilado := compilarGenerado(t, documento, rutaDeLaSalida)

		delApplet := schema.Resultado{Procedencia: procedenciaDelApplet, Datos: DatosError{Codigo: 7}}
		require.NoError(t, compilado.Validate(sobreEmitido(t, delApplet, nil)),
			"el sobre de éxito de ese verbo valida contra su propio esquema")

		for _, caso := range casosDeFallo(t) {
			require.NoError(t, compilado.Validate(
				sobreEmitido(t, schema.Resultado{Procedencia: caso.delApplet}, caso.err)),
				"el sobre de fallo sigue describiéndose con la forma del kernel: %s", caso.nombre)
		}

		delKernel := schema.Resultado{
			Procedencia: procedenciaDelApplet,
			Datos:       schema.DatosError{Clase: schema.ClaseArgumentos, Mensaje: "no soy el data del applet"},
		}
		require.Error(t, compilado.Validate(sobreEmitido(t, delKernel, nil)),
			"con ok verdadero, la forma de error del kernel no es el data de ese applet")
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
		assert.Equal(t, prefijoDeDefinicion+definicionDeCuenta,
			referencia(t, bajar(t, documento, "properties", "salida"), "then"))
		assert.NotContains(t, bajar(t, documento, "$defs"), definicionDelEco)
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

// ArgumentosComunes es el trozo de gramática que varios verbos de un mismo applet
// comparten embebiéndolo, igual que toda gramática embebe Globales. Va exportado
// porque es así como Kong puede rellenar sus campos promovidos, que es justo lo que
// hace pertinente la regla que TestDescribeGramatica comprueba.
type ArgumentosComunes struct {
	Compartido string `required:""`
}

// argumentosConMarcas ejerce de una vez las cuatro marcas con las que la gramática
// decide si la invocación tiene que escribir un campo, más las dos formas que
// camposVisibles descarta: el propio campo embebido y uno no exportado.
type argumentosConMarcas struct {
	ArgumentosComunes

	Opcional       string `optional:""`
	Predeterminado string `default:"7"`
	DePosicion     string `arg:""`

	// descartado no nombra ninguna bandera: la gramática no puede rellenar un
	// campo que no se exporta, así que describirlo describiría algo que la
	// invocación no puede escribir.
	descartado string
}

// verboConMarcas es la definición del verbo cuyos argumentos son los de arriba. No
// declara salida a propósito: lo que este verbo existe para comprobar está entero
// en la parte de entrada del documento.
func verboConMarcas() Verbo {
	return Verbo{
		Applet:     nombreDePrueba,
		Verbo:      "marcar",
		Ayuda:      "Ejerce las marcas de la gramática.",
		Argumentos: &argumentosConMarcas{descartado: "ni se describe ni se escribe"},
	}
}

// TestDescribeGramatica comprueba que la entrada descrita es exactamente la que la
// gramática acepta: qué campos nombran una bandera y cuáles de ellos tiene que
// escribir quien invoca. Describir como obligatorio lo que no lo es —o describir un
// campo que la invocación no puede escribir— describiría un binario distinto del
// que se publica (FR-048, SC-010).
func TestDescribeGramatica(t *testing.T) {
	t.Parallel()

	t.Run("las cuatro marcas deciden qué es obligatorio", func(t *testing.T) {
		t.Parallel()

		documento, _ := describirDePrueba(t, verboConMarcas())
		entrada := bajar(t, documento, "properties", "entrada")

		assert.ElementsMatch(t, []any{"compartido", "de-posicion"}, entrada["required"],
			"lo marcado como obligatorio y el argumento de posición, que lo es por "+
				"serlo; lo marcado como opcional y lo que tiene valor por omisión, no")
	})

	t.Run("el campo embebido no se describe y sus hijos sí", func(t *testing.T) {
		t.Parallel()

		documento, _ := describirDePrueba(t, verboConMarcas())
		propiedades := bajar(t, documento, "properties", "entrada", "properties")

		assert.ElementsMatch(t,
			append([]string{"compartido", "opcional", "predeterminado", "de-posicion"},
				nombresDeLasOcho...),
			claves(propiedades),
			"el campo embebido no nombra ninguna bandera —sus hijos ya están en la "+
				"lista— y lo no exportado no lo puede rellenar la gramática")
	})
}

// TestDescribeDefectosDelKernel comprueba los dos fallos que no puede provocar quien
// invoca, sino quien declara el sobre o el documento: el sobre que dejara de
// declarar una de las dos claves de las que depende la condición, y el documento que
// no se puede codificar. Los dos salen con el código de lo que nadie previó y
// ninguno emite un esquema a medias (FR-031, FR-048).
func TestDescribeDefectosDelKernel(t *testing.T) {
	t.Parallel()

	t.Run("un campo que el sobre no declara no se puede describir", func(t *testing.T) {
		t.Parallel()

		clave, err := claveDelSobre("NoExiste")

		require.ErrorIs(t, err, errEsquemaImposible)
		assert.Empty(t, clave)
		assert.Contains(t, err.Error(), "NoExiste", "el fallo nombra el campo que falta")
		assert.Equal(t, 1, CodigoSalida(err),
			"que el sobre deje de declarar un campo no es un error de quien invoca")
	})

	t.Run("las dos claves de la condición salen de la etiqueta del sobre", func(t *testing.T) {
		t.Parallel()

		for nombre, esperada := range map[string]string{"Ok": "ok", "Data": "data"} {
			clave, err := claveDelSobre(nombre)

			require.NoError(t, err)
			assert.Equal(t, esperada, clave,
				"la condición sigue al sobre y no a una copia de sus claves (FR-048)")
		}
	})

	t.Run("dos tipos distintos con el mismo nombre no se describen", func(t *testing.T) {
		t.Parallel()

		def := verboDePrueba()
		def.Salida = salidaConHomonimos()

		doble := &presentadorConJSON{}
		err := Describir(doble, def)

		require.ErrorIs(t, err, errEsquemaImposible)
		assert.Contains(t, err.Error(), "Homonimo", "el fallo nombra el nombre en conflicto")
		assert.Equal(t, 1, CodigoSalida(err),
			"un choque de nombres es un defecto de quien declaró el verbo, no de quien invoca")
		assert.Empty(t, doble.salida.String(),
			"no se emite un esquema que describiría otra cosa: la biblioteca, sola, "+
				"haría que el segundo tipo referenciara la definición del primero")
	})

	t.Run("el mismo tipo dos veces no es un choque de nombres", func(t *testing.T) {
		t.Parallel()

		// El sobre de fallo se refleja en la rama `else` y puede aparecer además
		// dentro del data de un applet: es el mismo tipo y la misma definición,
		// no una colisión.
		g := nuevoGenerador()
		g.registrar(definicionDeError, tipoDatosDeFallo)
		g.registrar(definicionDeError, tipoDatosDeFallo)
		require.NoError(t, g.colision)

		g.registrar(definicionDeError, reflect.TypeOf(DatosError{}))
		require.ErrorIs(t, g.colision, errEsquemaImposible)
		assert.Contains(t, g.colision.Error(), "schema.DatosError")
		assert.Contains(t, g.colision.Error(), "cli.DatosError")
	})

	t.Run("un documento que no se puede codificar no se emite a medias", func(t *testing.T) {
		t.Parallel()

		// Extras es la única vía por la que un esquema puede acabar llevando un
		// valor que json no sabe escribir; un canal es el más corto que hay.
		documento, err := serializar(&invopop.Schema{
			Extras: map[string]any{"roto": make(chan int)},
		})

		require.ErrorIs(t, err, errEsquemaImposible)
		assert.Empty(t, documento, "un documento a medias es peor que ninguno")
	})
}
