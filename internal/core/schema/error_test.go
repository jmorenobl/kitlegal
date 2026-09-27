package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errSinClase es un error corriente, de los que no declaran clase: la mayoría.
var errSinClase = errors.New("el índice 7 queda fuera del segmento")

// errorDeclarado es un error de test que declara su propia clase, como hará el
// error tipado del cliente HTTP: implementa ConClase sin conocer el kernel, sus
// sentinelas ni sus códigos de salida (research.md D4).
type errorDeclarado struct {
	clase Clase
}

func (e errorDeclarado) Error() string {
	return "el sitio fuente.prueba no respondió a tiempo"
}

func (e errorDeclarado) Clase() Clase {
	return e.clase
}

// La comprobación en tiempo de compilación de que un tipo así satisface la
// interfaz: si ConClase dejara de embeber error, o si cambiara la firma de
// Clase, esto no compilaría.
var _ ConClase = errorDeclarado{}

// TestClase comprueba el vocabulario de clases de error que aparece en el sobre
// y en el esquema que emite --describe: las siete clases, con el valor exacto que
// declara contracts/sobre-de-salida.md §6, y los datos de error con sus dos
// únicas claves (FR-029, FR-045).
func TestClase(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		clase  Clase
		valor  string
	}{
		{nombre: "argumentos", clase: ClaseArgumentos, valor: "argumentos"},
		{nombre: "no encontrado", clase: ClaseNoEncontrado, valor: "no-encontrado"},
		{nombre: "fuente no disponible", clase: ClaseFuenteNoDisponible, valor: "fuente-no-disponible"},
		{nombre: "límite o términos de uso", clase: ClaseLimiteOTos, valor: "limite-o-tos"},
		{nombre: "identidad humana", clase: ClaseIdentidadHumana, valor: "identidad-humana"},
		{nombre: "conflicto", clase: ClaseConflicto, valor: "conflicto"},
		{nombre: "inesperado", clase: ClaseInesperado, valor: "inesperado"},
	}

	for _, caso := range casos {
		t.Run("clase: "+caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.valor, string(caso.clase))

			emitido, err := json.Marshal(DatosError{Clase: caso.clase, Mensaje: "mensaje para la persona"})
			require.NoError(t, err)
			assert.JSONEq(t,
				`{"clase":"`+caso.valor+`","mensaje":"mensaje para la persona"}`,
				string(emitido))
		})
	}

	t.Run("las siete clases son distintas entre sí", func(t *testing.T) {
		t.Parallel()

		vistas := make(map[Clase]struct{}, len(casos))
		for _, caso := range casos {
			vistas[caso.clase] = struct{}{}
		}
		assert.Len(t, vistas, len(casos))
	})

	t.Run("el vocabulario enumera las siete del contrato, en su orden, y ninguna más", func(t *testing.T) {
		t.Parallel()

		// La lista esperada sale de la tabla de arriba, que escribe los valores
		// del contrato a mano: si alguien añadiera una octava constante y se
		// olvidara de Clases —o al revés—, el esquema de --describe y el sobre
		// dejarían de hablar del mismo vocabulario, y esto lo señalaría.
		esperadas := make([]Clase, 0, len(casos))
		for _, caso := range casos {
			esperadas = append(esperadas, Clase(caso.valor))
		}

		assert.Equal(t, esperadas, Clases())

		primera, segunda := Clases(), Clases()
		primera[0] = "alterada"
		assert.NotEqual(t, primera[0], segunda[0],
			"cada llamada devuelve una lista propia: el vocabulario no se altera desde fuera")
	})

	t.Run("los datos de error llevan exactamente clase y mensaje", func(t *testing.T) {
		t.Parallel()

		emitido, err := json.Marshal(DatosError{Clase: ClaseNoEncontrado, Mensaje: "no existe el bloque a99"})
		require.NoError(t, err)

		var claves map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(emitido, &claves))

		emitidas := make([]string, 0, len(claves))
		for clave := range claves {
			emitidas = append(emitidas, clave)
		}
		assert.ElementsMatch(t, []string{"clase", "mensaje"}, emitidas)
	})
}

// TestConClase comprueba el puerto por el que un error declara la clase con la
// que debe traducirse a código de salida: que es un error —se devuelve como tal
// y se envuelve con %w—, que quien lo recibe lo encuentra en la cadena con
// errors.As aunque vaya envuelto, y que la clase que declara no la altera la
// envoltura (FR-031, FR-063, research.md D4).
//
// Aquí acaba lo que el dominio puede decir: quién traduce esa clase a un código
// de salida —y qué hace con una clase ajena al vocabulario— es cosa del kernel,
// y se comprueba en internal/cli.
func TestConClase(t *testing.T) {
	t.Parallel()

	for _, clase := range Clases() {
		t.Run("declara la clase "+string(clase), func(t *testing.T) {
			t.Parallel()

			// Se guarda en un error, que es lo que un adaptador devuelve: la
			// interfaz embebe error, así que no hace falta conversión ninguna.
			var declarado error = errorDeclarado{clase: clase}

			var conClase ConClase
			require.ErrorAs(t, declarado, &conClase)
			assert.Equal(t, clase, conClase.Clase())
		})

		t.Run("envuelta, la clase "+string(clase)+" sigue siendo la misma", func(t *testing.T) {
			t.Parallel()

			envuelto := fmt.Errorf("consultando el artículo a21: %w",
				fmt.Errorf("el sitio fuente.prueba: %w", errorDeclarado{clase: clase}))

			var conClase ConClase
			require.ErrorAs(t, envuelto, &conClase)
			assert.Equal(t, clase, conClase.Clase())
		})
	}

	t.Run("un error sin clase declarada no está en la cadena", func(t *testing.T) {
		t.Parallel()

		var conClase ConClase
		assert.NotErrorAs(t, fmt.Errorf("al montar la gramática: %w", errSinClase), &conClase)
	})
}
