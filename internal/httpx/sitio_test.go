package httpx

import (
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClaveDeSitio fija la identidad de un sitio, que es la base común de la
// caché de robots.txt (FR-013, FR-015) y del limitador de ritmo (FR-019): dos
// direcciones son del mismo sitio si y solo si su clave coincide. Las cuatro
// primeras filas fijan la forma de la clave; las cuatro restantes, qué dos
// direcciones la comparten (plan.md «Inventario de tests», data-model.md §4).
//
// Toda dirección de la tabla es de un host ficticio o local, que es lo único
// que admite la comprobación «solo direcciones locales» del quickstart.
func TestClaveDeSitio(t *testing.T) {
	t.Parallel()

	formas := []struct {
		nombre    string
		direccion string
		clave     string
	}{
		{
			nombre:    "sin puerto, http toma el 80 de su esquema",
			direccion: "http://fuente.prueba",
			clave:     "http://fuente.prueba:80",
		},
		{
			nombre:    "sin puerto, https toma el 443 de su esquema",
			direccion: "https://fuente.prueba",
			clave:     "https://fuente.prueba:443",
		},
		{
			nombre:    "el puerto explícito se conserva",
			direccion: "http://fuente.prueba:8080",
			clave:     "http://fuente.prueba:8080",
		},
		{
			nombre:    "host en minúsculas, y ni la ruta ni la consulta entran en la clave",
			direccion: "http://Fuente.Prueba/ruta?q=1",
			clave:     "http://fuente.prueba:80",
		},
	}

	for _, caso := range formas {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.clave, claveDeSitio(direccionDePrueba(t, caso.direccion)))
		})
	}

	parejas := []struct {
		nombre     string
		una        string
		otra       string
		mismoSitio bool
	}{
		{
			nombre:     "el esquema distingue el sitio",
			una:        "http://fuente.prueba",
			otra:       "https://fuente.prueba",
			mismoSitio: false,
		},
		{
			nombre:     "el host distingue el sitio",
			una:        "http://fuente.prueba",
			otra:       "http://otra.prueba",
			mismoSitio: false,
		},
		{
			nombre:     "el puerto distingue el sitio, aunque el host sea el mismo",
			una:        "http://127.0.0.1:53211",
			otra:       "http://127.0.0.1:53212",
			mismoSitio: false,
		},
		{
			nombre:     "el puerto por omisión y el mismo puerto explícito son un solo sitio",
			una:        "http://fuente.prueba",
			otra:       "http://fuente.prueba:80",
			mismoSitio: true,
		},
	}

	for _, caso := range parejas {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			una := claveDeSitio(direccionDePrueba(t, caso.una))
			otra := claveDeSitio(direccionDePrueba(t, caso.otra))

			if caso.mismoSitio {
				assert.Equal(t, una, otra)

				return
			}

			assert.NotEqual(t, una, otra)
		})
	}
}

// TestMapaDeSitios comprueba lo que el mapa garantiza: una sola entrada por
// clave de sitio, creada la primera vez que se ve la dirección, compartida por
// todo el que la pida después y también cuando varias goroutines la piden a la
// vez. Es lo que permite que el ritmo y el robots.txt de un sitio sean uno solo
// por cliente y no uno por petición (FR-021, SC-011, D14).
func TestMapaDeSitios(t *testing.T) {
	t.Parallel()

	t.Run("dos direcciones del mismo sitio comparten entrada", func(t *testing.T) {
		t.Parallel()

		sitios := nuevosSitios()

		primero := sitios.de(direccionDePrueba(t, "http://fuente.prueba/norma?id=1"))
		segundo := sitios.de(direccionDePrueba(t, "http://fuente.prueba:80/otra"))

		assert.Same(t, primero, segundo)
		assert.Equal(t, "http://fuente.prueba:80", primero.clave)
	})

	t.Run("dos sitios distintos tienen entradas distintas", func(t *testing.T) {
		t.Parallel()

		sitios := nuevosSitios()

		uno := sitios.de(direccionDePrueba(t, "http://fuente.prueba"))
		otro := sitios.de(direccionDePrueba(t, "http://otra.prueba"))

		assert.NotSame(t, uno, otro)
		assert.Equal(t, "http://otra.prueba:80", otro.clave)
	})

	t.Run("varias goroutines a la vez obtienen la misma entrada", func(t *testing.T) {
		t.Parallel()

		const goroutines = 16

		sitios := nuevosSitios()
		// La dirección se interpreta aquí, en la goroutine del test: dentro de
		// las otras no se puede fallar el test.
		direccion := direccionDePrueba(t, "http://fuente.prueba/ruta")
		obtenidos := make([]*sitio, goroutines)

		var grupo sync.WaitGroup

		for i := range goroutines {
			grupo.Go(func() {
				obtenidos[i] = sitios.de(direccion)
			})
		}

		grupo.Wait()

		for _, obtenido := range obtenidos {
			assert.Same(t, obtenidos[0], obtenido)
		}

		assert.Equal(t, "http://fuente.prueba:80", obtenidos[0].clave)
	})
}

// direccionDePrueba interpreta una dirección de las tablas de este fichero. Un
// fallo aquí es un error de la tabla, no del código bajo prueba.
func direccionDePrueba(t *testing.T, direccion string) *url.URL {
	t.Helper()

	interpretada, err := url.Parse(direccion)
	require.NoError(t, err, "la dirección de la tabla debe ser interpretable")

	return interpretada
}
