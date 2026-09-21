package territorio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestResolver fija cómo se lee una consulta (data-model §2.6): una entrada
// sin ninguna letra es un código INE de cinco cifras, o de seis con su dígito
// de control, y cualquier otra, un nombre. Un código se busca primero en la
// relación y solo después se compara su dígito con el oficial; un nombre se
// busca por su forma plegada. Nunca se elige entre varios candidatos, y una
// entrada que no llega a ser un código nunca es «no encontrado» (FR-010 a
// FR-015, US1, US4).
func TestResolver(t *testing.T) {
	t.Parallel()

	registro := cargarSintetico(t)

	porNombre, err := registro.Resolver("Villaprueba")
	require.NoError(t, err)

	t.Run("por-nombre", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "28991", porNombre.CodigoINE.Codigo)
		assert.Equal(t, "Villaprueba", porNombre.Municipio.Nombre)
	})

	// Lo que resuelve al mismo municipio devuelve exactamente lo mismo: los
	// mismos datos y la misma fecha (SC-001).
	mismoMunicipio := []struct {
		nombre  string
		entrada string
	}{
		{"por-codigo", "28991"},
		{"por-codigo-con-digito", "289915"},
		{"nombre-con-mayusculas-y-sin-tildes", "VILLAPRUEBA"},
		{"nombre-con-espacios-alrededor", "  villaprueba "},
	}

	for _, caso := range mismoMunicipio {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			resuelto, err := registro.Resolver(caso.entrada)
			require.NoError(t, err)
			assert.Equal(t, porNombre, resuelto)
		})
	}

	t.Run("forma-alternativa", func(t *testing.T) {
		t.Parallel()

		casos := map[string]string{
			"Las Rozas de Prueba":  "28992",
			"rozas de prueba, las": "28992",
			"Pamploneta":           "31991",
			"IRUÑETA":              "31991",
			"Iruñeta/Pamploneta":   "31991",
			"peniscola del rio":    "47992",
			"Castro, El":           "05992",
		}

		for entrada, codigo := range casos {
			resuelto, err := registro.Resolver(entrada)
			require.NoError(t, err, "la consulta %q", entrada)
			assert.Equal(t, codigo, resuelto.CodigoINE.Codigo, "la consulta %q", entrada)
		}
	})

	const (
		ambiguo      = "es el de 2 municipios de la relación; consulta uno por su código INE: "
		fueraDeRango = "no está entre 01 y 52"
	)

	fallos := []struct {
		nombre string
		clase  schema.Clase
		// consultas son las entradas del caso, cada una con lo que su mensaje
		// tiene que decir además de nombrarla.
		consultas map[string]string
	}{
		{
			nombre: "digito-incorrecto",
			clase:  schema.ClaseArgumentos,
			consultas: map[string]string{
				"289916": `el dígito de control recibido es "6" y el oficial es "5"`,
			},
		},
		{
			nombre: "codigo-inexistente",
			clase:  schema.ClaseNoEncontrado,
			consultas: map[string]string{
				"28999":  "ningún municipio de la relación tiene el código INE",
				"289990": "ningún municipio de la relación tiene el código INE",
				"52001":  "ningún municipio de la relación tiene el código INE",
			},
		},
		{
			nombre: "nombre-inexistente",
			clase:  schema.ClaseNoEncontrado,
			consultas: map[string]string{
				"Pueblo Que No Está": "ningún municipio de la relación se llama",
				"Villa":              "ningún municipio de la relación se llama",
			},
		},
		{
			nombre: "nombre-ambiguo",
			clase:  schema.ClaseArgumentos,
			consultas: map[string]string{
				"Villanueva": ambiguo + "05991 Villanueva (Provincia Primera); 47991 Villanueva (Provincia Segunda)",
				"VILLANUEVA": ambiguo + "05991 Villanueva (Provincia Primera); 47991 Villanueva (Provincia Segunda)",
				"El Castro":  ambiguo + "05992 Castro, El (Provincia Primera); 47993 El Castro (Provincia Segunda)",
			},
		},
		{
			nombre: "entrada-de-solo-cifras-invalida",
			clase:  schema.ClaseArgumentos,
			consultas: map[string]string{
				"00991":   `la provincia "00" ` + fueraDeRango,
				"009915":  `la provincia "00" ` + fueraDeRango,
				"53001":   `la provincia "53" ` + fueraDeRango,
				"99999":   `la provincia "99" ` + fueraDeRango,
				"990015":  `la provincia "99" ` + fueraDeRango,
				"28000":   `el municipio "000" no está entre 001 y 999`,
				"280005":  `el municipio "000" no está entre 001 y 999`,
				"2899":    "tiene 4 cifras y la forma PPMMM tiene 5",
				"2899151": "tiene 7 cifras y la forma PPMMM tiene 5",
			},
		},
		{
			nombre: "entrada-sin-letras-con-algo-que-no-es-cifra",
			clase:  schema.ClaseArgumentos,
			consultas: map[string]string{
				" 28991": `" " no es una cifra`,
				"28-991": `"-" no es una cifra`,
				"2899 1": `" " no es una cifra`,
			},
		},
		{
			nombre: "entrada-vacia",
			clase:  schema.ClaseArgumentos,
			consultas: map[string]string{
				"":       "no nombra ningún municipio: no tiene ninguna letra ni ninguna cifra",
				"   ":    "no nombra ningún municipio: no tiene ninguna letra ni ninguna cifra",
				" ,.-/ ": "no nombra ningún municipio: no tiene ninguna letra ni ninguna cifra",
				"·":      "no nombra ningún municipio: no tiene ninguna letra ni ninguna cifra",
			},
		},
	}

	for _, caso := range fallos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			for entrada, motivo := range caso.consultas {
				resuelto, err := registro.Resolver(entrada)
				compruebaFalloDeConsulta(t, err, entrada, caso.clase)
				require.ErrorContains(t, err, motivo, "la consulta %q", entrada)
				assert.Equal(t, Territorio{}, resuelto, "un fallo no puede acompañarse de un territorio")
			}
		})
	}
}
