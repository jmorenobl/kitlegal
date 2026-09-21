package territorio

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// compruebaFalloDeConsulta exige que err sea la respuesta negativa a una
// consulta: un error que declara la clase esperada —buscada como la busca el
// kernel, con errors.As a schema.ConClase siguiendo la cadena—, que es una de
// las dos que una consulta puede tener, y que nombra la entrada con %q. El
// código de salida que corresponde a cada clase no se puede comprobar aquí: el
// dominio no importa internal/cli ni en sus tests (research.md V7, V18).
func compruebaFalloDeConsulta(t *testing.T, err error, entrada string, clase schema.Clase) {
	t.Helper()

	require.Error(t, err, "la consulta %q no puede dar un territorio", entrada)

	var conClase schema.ConClase
	require.ErrorAs(t, err, &conClase, "el fallo de la consulta %q no declara su clase", entrada)
	assert.Equal(t, clase, conClase.Clase(), "la clase del fallo de la consulta %q", entrada)
	assert.Contains(t, []schema.Clase{schema.ClaseArgumentos, schema.ClaseNoEncontrado}, conClase.Clase(),
		"una consulta solo puede fallar por argumentos o por no encontrado: nunca 4, 5 ni 6 (FR-016)")
	assert.Contains(t, err.Error(), fmt.Sprintf("%q", entrada), "el mensaje no nombra la entrada")
}

// TestClaseDeLosErroresDeTerritorio pasa por cada camino por el que una
// consulta falla y exige de todos que declaren schema.ClaseArgumentos —entrada
// mal formada, dígito distinto del oficial, nombre ambiguo— o
// schema.ClaseNoEncontrado —código o nombre que no está en la relación—,
// ninguna otra, también envueltos con %w, y que su mensaje nombre la entrada.
// Los errores de carga, en cambio, no llevan ninguna clase de usuario: unas
// fuentes que no cargan son un defecto de composición, que el kernel trata
// como inesperado, y no algo que quien pregunta pueda corregir; tampoco
// cuando el defecto es un código que el analizador de identificadores rechaza
// con su propia clase (FR-096, contrato del applet §7).
func TestClaseDeLosErroresDeTerritorio(t *testing.T) {
	t.Parallel()

	t.Run("fallos de consulta", func(t *testing.T) {
		t.Parallel()

		registro := cargarSintetico(t)

		consultas := []struct {
			nombre  string
			entrada string
			clase   schema.Clase
		}{
			{"vacía", "", schema.ClaseArgumentos},
			{"solo separadores", " - ", schema.ClaseArgumentos},
			{"código de otra longitud", "2899", schema.ClaseArgumentos},
			{"código con algo que no es cifra", "28 991", schema.ClaseArgumentos},
			{"provincia fuera de rango", "99991", schema.ClaseArgumentos},
			{"municipio 000", "28000", schema.ClaseArgumentos},
			{"dígito distinto del oficial", "289916", schema.ClaseArgumentos},
			{"nombre ambiguo", "Villanueva", schema.ClaseArgumentos},
			{"código fuera de la relación", "28999", schema.ClaseNoEncontrado},
			{"código con dígito fuera de la relación", "289990", schema.ClaseNoEncontrado},
			{"nombre fuera de la relación", "Pueblo Que No Está", schema.ClaseNoEncontrado},
		}

		for _, consulta := range consultas {
			t.Run(consulta.nombre, func(t *testing.T) {
				t.Parallel()

				_, err := registro.Resolver(consulta.entrada)
				compruebaFalloDeConsulta(t, err, consulta.entrada, consulta.clase)
				compruebaFalloDeConsulta(t, fmt.Errorf("resolver el territorio: %w", err), consulta.entrada,
					consulta.clase)
			})
		}
	})

	t.Run("errores de carga", func(t *testing.T) {
		t.Parallel()

		defectos := []struct {
			nombre  string
			fuentes func(t *testing.T) Fuentes
		}{
			{"yaml ilegible", func(t *testing.T) Fuentes {
				t.Helper()

				fuentes := fuentesDe(t, ficherosSinteticos())
				fuentes.Municipios = []byte("municipios: [\n")

				return fuentes
			}},
			{"código que el analizador rechaza", func(t *testing.T) Fuentes {
				t.Helper()

				ficheros := ficherosSinteticos()
				ficheros.Municipios.Municipios["99991"] = FilaDeMunicipio{
					DC: "1", Nombre: "Pueblo Fuera de Rango", Provincia: "99", Comunidad: "01",
				}

				return fuentesDe(t, ficheros)
			}},
			{"DIR3 que el analizador rechaza", func(t *testing.T) Fuentes {
				t.Helper()

				ficheros := ficherosSinteticos()
				ficheros.DIR3.Correspondencia["28991"] = "L0128991"

				return fuentesDe(t, ficheros)
			}},
			{"integridad", func(t *testing.T) Fuentes {
				t.Helper()

				ficheros := ficherosSinteticos()
				cambiarMunicipio(&ficheros, "28991", func(m *FilaDeMunicipio) { m.Comunidad = "02" })

				return fuentesDe(t, ficheros)
			}},
		}

		for _, defecto := range defectos {
			t.Run(defecto.nombre, func(t *testing.T) {
				t.Parallel()

				_, err := Cargar(defecto.fuentes(t))
				require.Error(t, err)

				var conClase schema.ConClase
				require.NotErrorAs(t, err, &conClase,
					"un error de carga no puede llevar clase de usuario: es un defecto de composición")
			})
		}
	})
}
