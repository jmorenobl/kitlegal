package graph

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestErrores fija la tabla de contracts/almacen-world-db.md §6: cada situación
// con su clase, su operación, su ruta y su mensaje literal, que empieza por
// «grafo: » y nombra world.db; y la causa alcanzable con errors.Is a través del
// error del paquete, también envuelto.
func TestErrores(t *testing.T) {
	t.Parallel()

	directorio := filepath.Join("cache", "kitlegal")
	ruta := filepath.Join(directorio, "world.db")
	citada := strconv.Quote(ruta)
	causa := errors.New("causa de prueba")
	rechazo := &grafo.Rechazo{Operacion: schema.Nodo{ID: "ine:28074", Tipo: "Municipio"}, Motivo: "motivo de prueba"}

	casos := []struct {
		nombre    string
		fallo     *Error
		mensaje   string
		clase     schema.Clase
		operacion string
		ruta      string
		causa     error
	}{
		{
			nombre:    "ruta no resoluble",
			fallo:     errorDeUbicacion(operacionLeer, causa),
			mensaje:   "grafo: no se puede ubicar world.db: causa de prueba",
			clase:     schema.ClaseArgumentos,
			operacion: operacionLeer,
			causa:     causa,
		},
		{
			nombre:    "no es una base de datos utilizable",
			fallo:     errorInutilizable(operacionLeer, ruta, causa),
			mensaje:   "grafo: " + citada + " no es una base de datos utilizable: causa de prueba",
			clase:     schema.ClaseInesperado,
			operacion: operacionLeer,
			ruta:      ruta,
			causa:     causa,
		},
		{
			nombre:    "esquema posterior",
			fallo:     errorDeVersionPosterior(operacionLeer, ruta, 2, 1),
			mensaje:   "grafo: " + citada + " tiene el esquema en la versión 2 y este binario conoce la 1: no se modifica",
			clase:     schema.ClaseInesperado,
			operacion: operacionLeer,
			ruta:      ruta,
		},
		{
			nombre:    "bloqueo más largo que la espera propia",
			fallo:     errorDeBloqueo(operacionEscribir, ruta, esperaPropia, causa),
			mensaje:   "grafo: " + citada + " está bloqueada por otra invocación y la espera de 5s se agotó",
			clase:     schema.ClaseInesperado,
			operacion: operacionEscribir,
			ruta:      ruta,
			causa:     causa,
		},
		{
			nombre:    "plazo agotado al leer",
			fallo:     errorDePlazo(operacionLeer, ruta, causa),
			mensaje:   "grafo: el plazo terminó antes de leer " + citada,
			clase:     schema.ClaseFuenteNoDisponible,
			operacion: operacionLeer,
			ruta:      ruta,
			causa:     causa,
		},
		{
			nombre:    "plazo agotado al escribir antes de ubicar world.db",
			fallo:     errorDePlazo(operacionEscribir, "", causa),
			mensaje:   "grafo: el plazo terminó antes de escribir world.db",
			clase:     schema.ClaseFuenteNoDisponible,
			operacion: operacionEscribir,
			causa:     causa,
		},
		{
			nombre:    "lote rechazado antes de ubicar world.db",
			fallo:     errorDeLoteRechazado("", rechazo),
			mensaje:   `grafo: el lote no entra en world.db: el nodo "ine:28074": motivo de prueba`,
			clase:     schema.ClaseInesperado,
			operacion: operacionEscribir,
			causa:     rechazo,
		},
		{
			nombre:    "lote rechazado con la ruta ya conocida",
			fallo:     errorDeLoteRechazado(ruta, rechazo),
			mensaje:   "grafo: el lote no entra en " + citada + `: el nodo "ine:28074": motivo de prueba`,
			clase:     schema.ClaseInesperado,
			operacion: operacionEscribir,
			ruta:      ruta,
			causa:     rechazo,
		},
		{
			nombre:    "directorio que no se puede crear",
			fallo:     errorDeDirectorioNoEscribible(directorio, causa),
			mensaje:   "grafo: no se puede escribir world.db en " + strconv.Quote(directorio) + ": causa de prueba",
			clase:     schema.ClaseInesperado,
			operacion: operacionEscribir,
			ruta:      directorio,
			causa:     causa,
		},
		{
			nombre:    "modo WAL que no se fija",
			fallo:     errorDeModoWAL(ruta, "delete"),
			mensaje:   "grafo: no se puede poner " + citada + " en modo WAL: el modo sigue siendo delete",
			clase:     schema.ClaseInesperado,
			operacion: operacionEscribir,
			ruta:      ruta,
		},
		{
			nombre:    "fallo de entrada y salida",
			fallo:     errorDeEntradaSalida(operacionLeer, ruta, causa),
			mensaje:   "grafo: no se pudo leer " + citada + ": causa de prueba",
			clase:     schema.ClaseInesperado,
			operacion: operacionLeer,
			ruta:      ruta,
			causa:     causa,
		},
		{
			nombre:    "fallo de entrada y salida al escribir",
			fallo:     errorDeEntradaSalida(operacionEscribir, ruta, causa),
			mensaje:   "grafo: no se pudo escribir " + citada + ": causa de prueba",
			clase:     schema.ClaseInesperado,
			operacion: operacionEscribir,
			ruta:      ruta,
			causa:     causa,
		},
		{
			nombre:    "migración que falla",
			fallo:     errorDeMigracion(ruta, "0001_grafo.sql", causa),
			mensaje:   `grafo: no se pudo aplicar la migración "0001_grafo.sql" en ` + citada + ": causa de prueba",
			clase:     schema.ClaseInesperado,
			operacion: operacionEscribir,
			ruta:      ruta,
			causa:     causa,
		},
		{
			nombre:    "migraciones que el binario no puede leer",
			fallo:     errorDeMigracionesIlegibles(ruta, causa),
			mensaje:   "grafo: no se pueden leer las migraciones de world.db que el binario trae dentro",
			clase:     schema.ClaseInesperado,
			operacion: operacionEscribir,
			ruta:      ruta,
			causa:     causa,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.mensaje, caso.fallo.Error())
			assert.True(t, strings.HasPrefix(caso.fallo.Error(), "grafo: "), "todo mensaje empieza por «grafo: »")
			assert.Contains(t, caso.fallo.Error(), "world.db", "todo mensaje nombra world.db")
			assert.Equal(t, caso.clase, caso.fallo.Clase())
			assert.Equal(t, caso.operacion, caso.fallo.Operacion)
			assert.Equal(t, caso.ruta, caso.fallo.Ruta)

			envuelto := fmt.Errorf("entrega: %w", caso.fallo)

			var conClase schema.ConClase

			require.ErrorAs(t, envuelto, &conClase, "el error declara su clase también envuelto")
			assert.Equal(t, caso.clase, conClase.Clase())

			if caso.causa == nil {
				require.NoError(t, caso.fallo.Unwrap(), "sin causa, Unwrap no devuelve nada")

				return
			}

			require.ErrorIs(t, envuelto, caso.causa, "la causa sigue alcanzable")
		})
	}
}

// TestErrorSinDeclarar fija que un *Error nulo o construido a cero desde fuera
// de sus constructores no deja el mensaje vacío ni entra en panic, y que lo que
// nadie declaró es «inesperado» (FR-010).
func TestErrorSinDeclarar(t *testing.T) {
	t.Parallel()

	for nombre, fallo := range map[string]*Error{"nulo": nil, "a cero": {}} {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "grafo: world.db ha fallado sin declarar el motivo", fallo.Error())
			assert.Equal(t, schema.ClaseInesperado, fallo.Clase())
			require.NoError(t, fallo.Unwrap())
		})
	}
}

// TestNombrar fija cómo se nombra world.db en un mensaje: con su ruta entre
// comillas cuando ya se conoce, que hace visibles los espacios y los caracteres
// de control, y por su nombre cuando todavía no.
func TestNombrar(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "world.db", nombrar(""))
	assert.Equal(t, `"/tmp/a b/world.db"`, nombrar("/tmp/a b/world.db"))
	assert.Equal(t, `"/tmp/a\nb/world.db"`, nombrar("/tmp/a\nb/world.db"))
}
