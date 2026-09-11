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
	}
}

// TestClasificar comprueba que la clase de un error se decide con errors.Is
// sobre los cinco sentinelas, que envolver con %w no la cambia y que lo que no
// casa con ninguno es inesperado y no un éxito disfrazado (FR-029, FR-031,
// FR-032).
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
