package cli

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// errorDeOtroPaquete es un error cualquiera, ajeno al kernel y sin relación con
// ningún sentinela: representa el fallo de programación que FR-031 obliga a no
// confundir ni con un éxito ni con un fallo clasificado.
var errorDeOtroPaquete = errors.New("el índice 7 queda fuera del segmento")

// errorConClase es un error de test que declara él mismo su clase, como hará el
// error tipado del cliente HTTP: implementa schema.ConClase sin conocer los
// sentinelas del kernel ni sus códigos de salida (FR-063, research.md D4).
type errorConClase struct {
	clase   schema.Clase
	mensaje string
}

func (e errorConClase) Error() string {
	return e.mensaje
}

func (e errorConClase) Clase() schema.Clase {
	return e.clase
}

// errorConClaseSobreSentinela declara una clase **y** envuelve un sentinela del
// kernel. No es lo que hará ningún adaptador —ninguno importa internal/cli—,
// pero fija el orden que Clasificar promete: los cinco sentinelas se comprueban
// antes que la clase declarada, de modo que lo que H1 clasificaba de una manera
// se sigue clasificando igual.
type errorConClaseSobreSentinela struct {
	errorConClase
	causa error
}

func (e errorConClaseSobreSentinela) Unwrap() error {
	return e.causa
}

// casoDeClase es una fila de la tabla: un error, la clase que le corresponde y
// el código de salida en el que acaba.
type casoDeClase struct {
	nombre string
	err    error
	clase  schema.Clase
	codigo int
}

// casosDeClase cubre las seis clases —una fila por clase, con el sentinela que
// la produce y, para la inesperada, el error que no casa con ninguno—, el error
// desconocido de otro paquete y el error envuelto en uno y en dos niveles
// (FR-029 … FR-032).
//
// Cubre además las mismas seis clases declaradas por el propio error, que es
// como las nombra un adaptador que no importa el kernel: la clase de fuera del
// vocabulario, la envuelta y la que compite con un sentinela (FR-063, D4).
func casosDeClase() []casoDeClase {
	return []casoDeClase{
		{
			nombre: "argumentos",
			err:    ErrArgumentos,
			clase:  schema.ClaseArgumentos,
			codigo: 2,
		},
		{
			nombre: "no-encontrado",
			err:    ErrNoEncontrado,
			clase:  schema.ClaseNoEncontrado,
			codigo: 3,
		},
		{
			nombre: "fuente-no-disponible",
			err:    ErrFuenteNoDisponible,
			clase:  schema.ClaseFuenteNoDisponible,
			codigo: 4,
		},
		{
			nombre: "limite-o-tos",
			err:    ErrLimiteOTos,
			clase:  schema.ClaseLimiteOTos,
			codigo: 5,
		},
		{
			nombre: "identidad-humana",
			err:    ErrIdentidadHumana,
			clase:  schema.ClaseIdentidadHumana,
			codigo: 6,
		},
		{
			nombre: "inesperado",
			err:    errors.New("el analizador no debería haber llegado aquí"),
			clase:  schema.ClaseInesperado,
			codigo: 1,
		},
		{
			nombre: "desconocido: un error de otro paquete",
			err:    errorDeOtroPaquete,
			clase:  schema.ClaseInesperado,
			codigo: 1,
		},
		{
			nombre: "desconocido: un error de la biblioteca estándar",
			err:    io.ErrUnexpectedEOF,
			clase:  schema.ClaseInesperado,
			codigo: 1,
		},
		{
			nombre: "envuelto: un nivel conserva la clase",
			err:    fmt.Errorf("el bloque a21 de BOE-A-2015-10565: %w", ErrNoEncontrado),
			clase:  schema.ClaseNoEncontrado,
			codigo: 3,
		},
		{
			nombre: "envuelto: dos niveles conservan la clase",
			err: fmt.Errorf("consultando el artículo: %w",
				fmt.Errorf("la sede no admite peticiones automáticas: %w", ErrIdentidadHumana)),
			clase:  schema.ClaseIdentidadHumana,
			codigo: 6,
		},
		{
			nombre: "envuelto: el texto de la envoltura no decide la clase",
			err:    fmt.Errorf("argumentos no encontrados en la fuente: %w", ErrLimiteOTos),
			clase:  schema.ClaseLimiteOTos,
			codigo: 5,
		},
		{
			nombre: "envuelto: envolver un error desconocido sigue siendo inesperado",
			err:    fmt.Errorf("al montar la gramática: %w", errorDeOtroPaquete),
			clase:  schema.ClaseInesperado,
			codigo: 1,
		},
		{
			nombre: "declarada: argumentos",
			err: errorConClase{
				clase:   schema.ClaseArgumentos,
				mensaje: "el método PUT no está permitido: solo GET y HEAD",
			},
			clase:  schema.ClaseArgumentos,
			codigo: 2,
		},
		{
			nombre: "declarada: no-encontrado",
			err: errorConClase{
				clase:   schema.ClaseNoEncontrado,
				mensaje: "el sitio fuente.prueba no tiene la ruta /a21",
			},
			clase:  schema.ClaseNoEncontrado,
			codigo: 3,
		},
		{
			nombre: "declarada: fuente-no-disponible",
			err: errorConClase{
				clase:   schema.ClaseFuenteNoDisponible,
				mensaje: "el sitio fuente.prueba no responde",
			},
			clase:  schema.ClaseFuenteNoDisponible,
			codigo: 4,
		},
		{
			nombre: "declarada: limite-o-tos",
			err: errorConClase{
				clase:   schema.ClaseLimiteOTos,
				mensaje: "el sitio fuente.prueba limita las peticiones",
			},
			clase:  schema.ClaseLimiteOTos,
			codigo: 5,
		},
		{
			nombre: "declarada: identidad-humana",
			err: errorConClase{
				clase:   schema.ClaseIdentidadHumana,
				mensaje: "la sede exige identidad humana",
			},
			clase:  schema.ClaseIdentidadHumana,
			codigo: 6,
		},
		{
			nombre: "declarada: inesperado",
			err: errorConClase{
				clase:   schema.ClaseInesperado,
				mensaje: "el cliente no debería haber llegado aquí",
			},
			clase:  schema.ClaseInesperado,
			codigo: 1,
		},
		{
			nombre: "declarada: una clase ajena al vocabulario es inesperada",
			err: errorConClase{
				clase:   schema.Clase("una-clase-que-nadie-ha-declarado"),
				mensaje: "un fallo que se inventa su clase",
			},
			clase:  schema.ClaseInesperado,
			codigo: 1,
		},
		{
			nombre: "declarada: una clase vacía es inesperada",
			err: errorConClase{
				clase:   schema.Clase(""),
				mensaje: "un fallo que se dejó la clase sin poner",
			},
			clase:  schema.ClaseInesperado,
			codigo: 1,
		},
		{
			nombre: "declarada: envuelta en un nivel conserva la clase",
			err: fmt.Errorf("consultando el artículo: %w", errorConClase{
				clase:   schema.ClaseLimiteOTos,
				mensaje: "el sitio fuente.prueba limita las peticiones",
			}),
			clase:  schema.ClaseLimiteOTos,
			codigo: 5,
		},
		{
			nombre: "declarada: envuelta en dos niveles conserva la clase",
			err: fmt.Errorf("consultando el artículo: %w",
				fmt.Errorf("el sitio fuente.prueba: %w", errorConClase{
					clase:   schema.ClaseFuenteNoDisponible,
					mensaje: "plazo agotado",
				})),
			clase:  schema.ClaseFuenteNoDisponible,
			codigo: 4,
		},
		{
			nombre: "declarada: el sentinela se comprueba antes que la clase declarada",
			err: errorConClaseSobreSentinela{
				errorConClase: errorConClase{
					clase:   schema.ClaseArgumentos,
					mensaje: "el sitio fuente.prueba no tiene la ruta /a21",
				},
				causa: ErrNoEncontrado,
			},
			clase:  schema.ClaseNoEncontrado,
			codigo: 3,
		},
	}
}

// TestClasificar comprueba que la clase de un error se decide con errors.Is
// sobre los cinco sentinelas, que envolver con %w no la cambia y que lo que no
// casa con ninguno es inesperado y no un éxito disfrazado (FR-029, FR-031,
// FR-032).
//
// Comprueba también la segunda vía, la de FR-063: un error que declara su clase
// con schema.ConClase la ve reconocida aunque vaya envuelto, sin que el kernel
// conozca su tipo ni el adaptador conozca los sentinelas; y que lo que declara
// una clase de fuera del vocabulario no se cuela en el sobre (research.md D4).
func TestClasificar(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeClase() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.clase, Clasificar(caso.err))
		})
	}

	t.Run("los cinco sentinelas son distintos entre sí", func(t *testing.T) {
		t.Parallel()

		sentinelas := []error{
			ErrArgumentos,
			ErrNoEncontrado,
			ErrFuenteNoDisponible,
			ErrLimiteOTos,
			ErrIdentidadHumana,
		}
		for i, uno := range sentinelas {
			for j, otro := range sentinelas {
				if i == j {
					continue
				}
				assert.NotErrorIs(t, uno, otro,
					"%v y %v deberían ser errores distintos", uno, otro)
			}
		}
	})

	t.Run("todo sentinela lleva un mensaje no vacío", func(t *testing.T) {
		t.Parallel()

		for _, caso := range casosDeClase() {
			assert.NotEmpty(t, caso.err.Error())
		}
	})

	t.Run("la clase que sale de Clasificar es siempre una de las seis", func(t *testing.T) {
		t.Parallel()

		// La clase que devuelve Clasificar es la que va al sobre de fallo, donde
		// el esquema de --describe la restringe al vocabulario: una clase de
		// fuera —declarada por un error que se la invente— haría que el sobre
		// dejara de validar contra su propio esquema.
		for _, caso := range casosDeClase() {
			assert.Contains(t, schema.Clases(), Clasificar(caso.err), "caso %q", caso.nombre)
		}
	})

	t.Run("el error nulo no tiene clase y nunca se confunde con un éxito", func(t *testing.T) {
		t.Parallel()

		// La ausencia de fallo no es ninguna de las seis clases. Quien pueda
		// tener un error nulo pregunta por el código de salida, que sí distingue
		// el éxito; la clase de lo que no es un error es la de lo no previsto.
		assert.Equal(t, schema.ClaseInesperado, Clasificar(nil))
	})
}

// TestCodigoSalida comprueba la única traducción de error a código de salida:
// los seis códigos de la tabla cerrada del proyecto, el 0 del éxito y el 1 del
// fallo inesperado, que no es ninguno de los reservados (FR-029 … FR-032,
// SC-006).
func TestCodigoSalida(t *testing.T) {
	t.Parallel()

	t.Run("correcto", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 0, CodigoSalida(nil))
	})

	for _, caso := range casosDeClase() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.codigo, CodigoSalida(caso.err))
		})
	}

	t.Run("ningún fallo sale con el código del éxito", func(t *testing.T) {
		t.Parallel()

		for _, caso := range casosDeClase() {
			assert.NotEqual(t, 0, CodigoSalida(caso.err), "caso %q", caso.nombre)
		}
	})

	t.Run("la tabla cubre las seis clases y ninguna repite código", func(t *testing.T) {
		t.Parallel()

		porClase := make(map[schema.Clase]int)
		for _, caso := range casosDeClase() {
			if anterior, visto := porClase[caso.clase]; visto {
				assert.Equal(t, anterior, caso.codigo,
					"la clase %q aparece con dos códigos distintos", caso.clase)
				continue
			}
			porClase[caso.clase] = caso.codigo
		}

		assert.Equal(t, map[schema.Clase]int{
			schema.ClaseInesperado:         1,
			schema.ClaseArgumentos:         2,
			schema.ClaseNoEncontrado:       3,
			schema.ClaseFuenteNoDisponible: 4,
			schema.ClaseLimiteOTos:         5,
			schema.ClaseIdentidadHumana:    6,
		}, porClase)
	})

	t.Run("una clase ajena al vocabulario sale con el código del fallo inesperado", func(t *testing.T) {
		t.Parallel()

		// El switch de la traducción no lleva rama por defecto, de modo que el
		// linter exhaustive falle si alguien añade una séptima clase sin código.
		// Una Clase construida a mano fuera del vocabulario no es ninguna de las
		// seis y acaba donde acaba todo lo que no se ha previsto: en el 1.
		assert.Equal(t, 1, codigoDeClase(schema.Clase("una-clase-que-nadie-ha-declarado")))
	})
}
