package app_test

import (
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/app/ejemplo"
)

// urlDelEsquemaEmitido identifica ante el compilador el documento que cada
// verbo emite con --describe. Es una URL de recurso, no una dirección que se
// visite: el compilador resuelve contra ella las referencias `#/$defs/…` del
// propio documento y nada más.
const urlDelEsquemaEmitido = "https://ventanillalegal.es/schemas/ejemplo/describe.json"

// rutaDeLaSalida es el puntero JSON de la parte del documento que describe el
// sobre. Se compila dentro del documento y no extraída, porque sus referencias
// apuntan a los `$defs` de la raíz; un consumidor real la usa igual.
const rutaDeLaSalida = "#/properties/salida"

// argumentosDeEjemplo es con qué se invoca cada verbo de cada applet de ejemplo
// para obtener un sobre de éxito real. Se escribe por nombre de applet y verbo,
// y el test exige que **todo** verbo registrado tenga fila: un verbo nuevo sin
// caso aquí falla en lugar de quedar sin validar.
var argumentosDeEjemplo = map[string]string{
	"echo repetir":    "hola",
	"contar letras":   "canción",
	"contar palabras": "dos palabras",
}

// TestSalidaContraSuEsquema es el control de la Definition of Done §1.4 sobre
// los applets de ejemplo: la **salida real** de cada verbo —lo que el binario
// escribe en la salida estándar con --json— valida contra el esquema que ese
// mismo verbo emite con --describe, compilado con las aserciones de formato
// activadas (FR-017, FR-047, SC-015; filas 4 y 5 de la tabla de gates de
// plan.md).
//
// Es la comprobación que ninguna tabla de internal/cli puede hacer: allí se
// validan sobres montados sobre verbos declarados en el propio test. Aquí el
// esquema sale de lo que el applet declara en Verbo.Salida y el sobre de lo que
// su Ejecutar devuelve en Resultado.Datos, que son dos declaraciones distintas
// que nada relaciona en tiempo de compilación: un applet que declarara una
// salida y devolviera otra publicaría con --describe un esquema que rechaza su
// propio sobre, y solo este test lo vería.
func TestSalidaContraSuEsquema(t *testing.T) {
	t.Parallel()

	for _, applet := range ejemplo.Applets() {
		for _, verbo := range applet.Verbos() {
			nombre := applet.Nombre() + " " + verbo.Nombre

			argumento, hay := argumentosDeEjemplo[nombre]
			require.Truef(t, hay, "el verbo %q no tiene argumentos en esta tabla", nombre)

			t.Run(nombre, func(t *testing.T) {
				t.Parallel()

				esquema := esquemaEmitidoPor(t, applet.Nombre(), verbo.Nombre)
				invocacion := []string{applet.Nombre(), verbo.Nombre, argumento}

				t.Run("el sobre de éxito real valida, y su data no admite otra forma", func(t *testing.T) {
					t.Parallel()

					exigirSobreDeExitoConforme(t, esquema, invocacion)
				})

				t.Run("un sobre de fallo real valida: bandera desconocida", func(t *testing.T) {
					t.Parallel()

					exigirSobreDeFalloConforme(t, esquema, append(invocacion, "--jsno", "--json"))
				})

				t.Run("un sobre de fallo real valida: valor inválido", func(t *testing.T) {
					t.Parallel()

					exigirSobreDeFalloConforme(t, esquema, append(invocacion, "--timeout", "0s", "--json"))
				})
			})
		}
	}
}

// esquemaEmitidoPor pide al binario el esquema del verbo y lo compila. Que
// compile es la mitad de SC-007; la otra mitad es lo que se valida después
// contra él.
func esquemaEmitidoPor(t *testing.T, applet, verbo string) *jsonschema.Schema {
	t.Helper()

	res := invocarEjemplo(t, applet, verbo, "--describe")
	require.Equal(t, 0, res.codigo, res.errores)

	documento, err := jsonschema.UnmarshalJSON(strings.NewReader(res.salida))
	require.NoError(t, err, "--describe emite un único documento JSON")

	// Las aserciones de formato no están activas por omisión en el borrador
	// 2020-12: sin activarlas, `format: uri` es una anotación y una url vacía
	// pasaría (contracts/sobre-de-salida.md §6, nota para quien escriba el
	// validador).
	compilador := jsonschema.NewCompiler()
	compilador.AssertFormat()
	require.NoError(t, compilador.AddResource(urlDelEsquemaEmitido, documento))

	compilado, err := compilador.Compile(urlDelEsquemaEmitido + rutaDeLaSalida)
	require.NoError(t, err, "el esquema emitido tiene que ser compilable por un validador")

	return compilado
}

// exigirSobreDeExitoConforme comprueba que el sobre de éxito real valida contra
// el esquema y que la validación **restringe** `data`: el mismo sobre con una
// clave de más o de menos dentro de `data` deja de validar. Sin esa segunda
// mitad, un esquema que dejara `data` sin describir —o que describiera otra
// forma con `additionalProperties` abierto— pasaría el test sin comprobar nada
// (FR-047, FR-048).
func exigirSobreDeExitoConforme(t *testing.T, esquema *jsonschema.Schema, invocacion []string) {
	t.Helper()

	res := invocarEjemplo(t, append(invocacion, "--json")...)
	require.Equal(t, 0, res.codigo, res.errores)

	sobre := sobreGenerico(t, res.salida)
	require.NoError(t, esquema.Validate(sobre),
		"la salida real del verbo valida contra el esquema que el propio verbo emite")

	datos := datosDelSobre(t, sobre)
	require.NotEmpty(t, datos, "un data vacío no permitiría comprobar que el esquema lo restringe")

	conClaveDeMas := sobreGenerico(t, res.salida)
	datosDelSobre(t, conClaveDeMas)["ajena"] = "no declarada"
	require.Error(t, esquema.Validate(conClaveDeMas),
		"el esquema rechaza una clave que el tipo declarado en Verbo.Salida no tiene")

	for clave := range datos {
		conClaveDeMenos := sobreGenerico(t, res.salida)
		delete(datosDelSobre(t, conClaveDeMenos), clave)
		assert.Errorf(t, esquema.Validate(conClaveDeMenos),
			"el esquema exige la clave %q que el tipo declarado en Verbo.Salida sí tiene", clave)
	}
}

// datosDelSobre es el `data` de un sobre cuyo contenido es un objeto, que es el
// caso de los dos applets de ejemplo y de todo sobre de fallo.
func datosDelSobre(t *testing.T, sobre map[string]any) map[string]any {
	t.Helper()

	datos, esObjeto := sobre["data"].(map[string]any)
	require.True(t, esObjeto, "el data de este sobre es un objeto")

	return datos
}

// exigirSobreDeFalloConforme comprueba que un sobre de fallo real —emitido por
// el kernel con la forma común de error— valida contra la rama `else` del
// esquema del verbo, que es lo que SC-015 exige: un test de contrato sobre una
// ejecución fallida no puede rechazar una salida conforme.
func exigirSobreDeFalloConforme(t *testing.T, esquema *jsonschema.Schema, invocacion []string) {
	t.Helper()

	res := invocarEjemplo(t, invocacion...)
	require.Equal(t, 2, res.codigo, "los dos fallos de esta tabla son de argumentos: %s", res.errores)

	sobre := sobreGenerico(t, res.salida)
	require.Equal(t, false, sobre["ok"], "el sobre es de fallo")
	require.NoError(t, esquema.Validate(sobre),
		"el sobre de fallo real valida contra el esquema que el propio verbo emite")
}

// sobreGenerico lee la salida estándar en la representación que el validador
// consume: números conservados como literales y un único documento. Se llama
// una vez por alteración porque cada una necesita su propia copia del sobre.
func sobreGenerico(t *testing.T, salida string) map[string]any {
	t.Helper()

	documento, err := jsonschema.UnmarshalJSON(strings.NewReader(salida))
	require.NoError(t, err, "la salida estándar lleva un único documento JSON")

	sobre, esObjeto := documento.(map[string]any)
	require.True(t, esObjeto, "el sobre es un objeto JSON")

	return sobre
}
