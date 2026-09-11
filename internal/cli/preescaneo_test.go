package cli

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// casoDePreEscaneo es una fila de la tabla: unos argumentos tal como llegarían
// en argv y las tres banderas que el pre-escaneo debe leer en ellos.
type casoDePreEscaneo struct {
	nombre   string
	args     []string
	esperado Preliminar
}

// casosReconocidos cubre la regla 1 de D25: las tres banderas en su forma larga
// exacta y con valor booleano explícito, con los seis literales que Kong acepta
// para un booleano y en cualquier posición de la lista. La última aparición
// manda, como en el análisis de Kong, donde cada aparición vuelve a asignar el
// campo.
func casosReconocidos() []casoDePreEscaneo {
	return []casoDePreEscaneo{
		{
			nombre:   "sin argumentos no hay ninguna bandera",
			args:     nil,
			esperado: Preliminar{},
		},
		{
			nombre:   "la lista vacía no es la lista nula",
			args:     []string{},
			esperado: Preliminar{},
		},
		{
			nombre:   "--json",
			args:     []string{"--json"},
			esperado: Preliminar{JSON: true},
		},
		{
			nombre:   "--verbose",
			args:     []string{"--verbose"},
			esperado: Preliminar{Verbose: true},
		},
		{
			nombre:   "--help",
			args:     []string{"--help"},
			esperado: Preliminar{Ayuda: true},
		},
		{
			nombre:   "las tres a la vez",
			args:     []string{"--json", "--verbose", "--help"},
			esperado: Preliminar{JSON: true, Verbose: true, Ayuda: true},
		},
		{
			nombre:   "las tres a la vez en otro orden dan lo mismo",
			args:     []string{"--help", "--json", "--verbose"},
			esperado: Preliminar{JSON: true, Verbose: true, Ayuda: true},
		},
		{
			nombre:   "entre los argumentos de una invocación real",
			args:     []string{"boe", "articulo", "BOE-A-2015-10565", "a21", "--json"},
			esperado: Preliminar{JSON: true},
		},
		{
			nombre:   "antes de los argumentos de una invocación real",
			args:     []string{"--verbose", "boe", "articulo", "BOE-A-2015-10565", "a21"},
			esperado: Preliminar{Verbose: true},
		},
		{
			nombre:   "en medio de banderas que el pre-escaneo no conoce",
			args:     []string{"--timeout=5s", "--json", "--asunto=expediente-3"},
			esperado: Preliminar{JSON: true},
		},
		{
			nombre:   "--json=true",
			args:     []string{"--json=true"},
			esperado: Preliminar{JSON: true},
		},
		{
			nombre:   "--json=1",
			args:     []string{"--json=1"},
			esperado: Preliminar{JSON: true},
		},
		{
			nombre:   "--json=yes",
			args:     []string{"--json=yes"},
			esperado: Preliminar{JSON: true},
		},
		{
			nombre:   "--json=TRUE: el literal no distingue mayúsculas, como en Kong",
			args:     []string{"--json=TRUE"},
			esperado: Preliminar{JSON: true},
		},
		{
			nombre:   "--json=Yes",
			args:     []string{"--json=Yes"},
			esperado: Preliminar{JSON: true},
		},
		{
			nombre:   "--json=false",
			args:     []string{"--json=false"},
			esperado: Preliminar{},
		},
		{
			nombre:   "--json=0",
			args:     []string{"--json=0"},
			esperado: Preliminar{},
		},
		{
			nombre:   "--json=no",
			args:     []string{"--json=no"},
			esperado: Preliminar{},
		},
		{
			nombre:   "--json=False",
			args:     []string{"--json=False"},
			esperado: Preliminar{},
		},
		{
			nombre:   "--verbose=false",
			args:     []string{"--verbose=false"},
			esperado: Preliminar{},
		},
		{
			nombre:   "--help=false",
			args:     []string{"--help=false"},
			esperado: Preliminar{},
		},
		{
			nombre:   "--verbose=true",
			args:     []string{"--verbose=true"},
			esperado: Preliminar{Verbose: true},
		},
		{
			nombre:   "--help=1",
			args:     []string{"--help=1"},
			esperado: Preliminar{Ayuda: true},
		},
		{
			nombre:   "el valor explícito de una no afecta a las otras",
			args:     []string{"--json=false", "--verbose", "--help=yes"},
			esperado: Preliminar{Verbose: true, Ayuda: true},
		},
		{
			nombre:   "repetida: la última aparición manda y apaga",
			args:     []string{"--json", "--json=false"},
			esperado: Preliminar{},
		},
		{
			nombre:   "repetida: la última aparición manda y enciende",
			args:     []string{"--json=false", "--json"},
			esperado: Preliminar{JSON: true},
		},
		{
			nombre:   "repetida tres veces: sigue mandando la última",
			args:     []string{"--json=yes", "--json=no", "--json=1"},
			esperado: Preliminar{JSON: true},
		},
	}
}

// casosIgnorados cubre la regla 3 de D25: todo token que no sea una de las tres
// formas exactas se ignora, y ninguno enciende ninguna bandera. Incluye las
// formas que el pre-escaneo declara **no** soportar —corta, abreviada, negada,
// con valor separado— y los valores que Kong rechazaría: ahí el pre-escaneo no
// valida ni falla, solo deja la bandera como estaba.
func casosIgnorados() []casoDePreEscaneo {
	return []casoDePreEscaneo{
		{nombre: "la forma corta no se reconoce", args: []string{"-j"}},
		{nombre: "la forma corta de --verbose tampoco", args: []string{"-v"}},
		{nombre: "un solo guion delante del nombre largo", args: []string{"-json"}},
		{nombre: "la abreviatura del nombre largo", args: []string{"--jso"}},
		{nombre: "un nombre más largo que empieza igual", args: []string{"--json-lines"}},
		{nombre: "un nombre pegado al de la bandera", args: []string{"--jsonx"}},
		{nombre: "el mismo nombre en mayúsculas", args: []string{"--JSON"}},
		{nombre: "la forma negada, que ninguna global declara", args: []string{"--no-json"}},
		{nombre: "otra bandera que empieza por --help", args: []string{"--help-me"}},
		{nombre: "el nombre sin guiones no es una bandera", args: []string{"json"}},
		{nombre: "el valor separado no se consume", args: []string{"--json", "false"}, esperado: Preliminar{JSON: true}},
		{nombre: "un valor que Kong rechazaría no enciende nada", args: []string{"--json=maybe"}},
		{nombre: "un valor vacío tampoco", args: []string{"--json="}},
		{nombre: "un valor con dos iguales tampoco", args: []string{"--json=true=false"}},
		{nombre: "el valor de otra bandera que se llama como un literal", args: []string{"--asunto=--json"}},
		{nombre: "una bandera del kernel que el pre-escaneo no lee", args: []string{"--offline"}},
		{nombre: "la cadena vacía", args: []string{""}},
		{nombre: "un guion suelto, que es la entrada estándar", args: []string{"-"}},
		{nombre: "solo espacios", args: []string{"   "}},
		{nombre: "el nombre con espacio delante", args: []string{" --json"}},
		{nombre: "el nombre con espacio detrás", args: []string{"--json "}},
		{
			nombre: "varios tokens desconocidos seguidos",
			args:   []string{"boe", "articulo", "--timeout", "5s", "--asunto", "expediente-3"},
		},
	}
}

// casosTrasElTerminador cubre la regla 2 de D25: el pre-escaneo se detiene en el
// terminador porque lo que va después son argumentos y no banderas, aunque se
// escriban igual que una bandera.
func casosTrasElTerminador() []casoDePreEscaneo {
	return []casoDePreEscaneo{
		{
			nombre:   "el terminador solo",
			args:     []string{"--"},
			esperado: Preliminar{},
		},
		{
			nombre:   "tras el terminador, --json es un argumento",
			args:     []string{"--", "--json"},
			esperado: Preliminar{},
		},
		{
			nombre:   "tras el terminador, las tres son argumentos",
			args:     []string{"--", "--json", "--verbose", "--help"},
			esperado: Preliminar{},
		},
		{
			nombre:   "tras el terminador, tampoco con valor explícito",
			args:     []string{"--", "--json=true"},
			esperado: Preliminar{},
		},
		{
			nombre:   "lo anterior al terminador sí cuenta",
			args:     []string{"--json", "--", "--verbose"},
			esperado: Preliminar{JSON: true},
		},
		{
			nombre:   "el terminador no apaga lo ya leído",
			args:     []string{"--json", "--", "--json=false"},
			esperado: Preliminar{JSON: true},
		},
		{
			nombre:   "un segundo terminador es un argumento como cualquier otro",
			args:     []string{"--verbose", "--", "--", "--json"},
			esperado: Preliminar{Verbose: true},
		},
		{
			nombre:   "el terminador al final no cambia nada",
			args:     []string{"--json", "--verbose", "--"},
			esperado: Preliminar{JSON: true, Verbose: true},
		},
		{
			nombre:   "el mensaje de echo que se llama como una bandera",
			args:     []string{"echo", "--help", "--", "--json"},
			esperado: Preliminar{Ayuda: true},
		},
		{
			nombre:   "un terminador con valor pegado no es el terminador",
			args:     []string{"--=x", "--json"},
			esperado: Preliminar{JSON: true},
		},
	}
}

// TestPreEscanear comprueba las tres reglas del pre-escaneo acotado: qué formas
// reconoce, que ignora todo lo demás y que se detiene en el terminador
// (FR-045, research.md D25).
//
// La coincidencia con lo que Kong analiza no se comprueba aquí sino en el
// TestPreescaneo de T005, cuando existen las globales contra las que
// contrastarla; esta tabla fija el comportamiento del pre-escaneo por sí solo.
func TestPreEscanear(t *testing.T) {
	t.Parallel()

	t.Run("formas reconocidas", func(t *testing.T) {
		t.Parallel()

		for _, caso := range casosReconocidos() {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, caso.esperado, PreEscanear(caso.args))
			})
		}
	})

	t.Run("formas ignoradas", func(t *testing.T) {
		t.Parallel()

		for _, caso := range casosIgnorados() {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, caso.esperado, PreEscanear(caso.args))
			})
		}
	})

	t.Run("tras el terminador", func(t *testing.T) {
		t.Parallel()

		for _, caso := range casosTrasElTerminador() {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, caso.esperado, PreEscanear(caso.args))
			})
		}
	})
}

// TestPreEscanearNoConsume comprueba la mitad de la regla 3 que no se ve en el
// valor devuelto: el pre-escaneo no consume argumentos. La lista que recibe
// sigue intacta después de la llamada, de modo que quien la pasa puede
// entregársela después a Kong entera y sin copiarla.
func TestPreEscanearNoConsume(t *testing.T) {
	t.Parallel()

	for _, caso := range casosReconocidos() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			intactos := slices.Clone(caso.args)

			PreEscanear(caso.args)

			assert.Equal(t, intactos, caso.args)
		})
	}
}

// TestPreEscanearNoFalla pasa entradas degeneradas y comprueba lo que la
// firma ya promete —no hay error que devolver— y lo que la firma no puede
// prometer: que ninguna de ellas provoca un pánico. El pre-escaneo ocurre antes
// de que exista nada que emita un fallo, así que no tiene forma de informar de
// uno; su única salida posible es un valor.
func TestPreEscanearNoFalla(t *testing.T) {
	t.Parallel()

	degeneradas := [][]string{
		nil,
		{},
		{""},
		{"", "", ""},
		{"="},
		{"=json"},
		{"--"},
		{"---"},
		{"----json"},
		{"--json=true", "=", "--", ""},
		{"--ñ", "--JSON=sí", "—json"},
		{"\x00", "\n", "\t--json"},
	}

	for _, args := range degeneradas {
		t.Run(fmt.Sprintf("%q", args), func(t *testing.T) {
			t.Parallel()

			assert.NotPanics(t, func() { PreEscanear(args) })
		})
	}
}
