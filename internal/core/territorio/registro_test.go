package territorio

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/ids"
)

// codigoINE analiza un código INE de los tests, que tiene que ser válido.
func codigoINE(t *testing.T, codigo string) ids.CodigoINE {
	t.Helper()

	analizado, err := ids.AnalizarCodigoINE(codigo)
	require.NoError(t, err)

	return analizado
}

// codigosDe son los códigos INE de unos municipios del registro, en su orden.
func codigosDe(municipios []*municipioRegistrado) []string {
	codigos := make([]string, 0, len(municipios))
	for _, municipio := range municipios {
		codigos = append(codigos, municipio.codigo.String())
	}

	return codigos
}

// TestRegistro fija los dos índices del registro: por código INE y por forma
// plegada del nombre, con los candidatos de una forma que lleva a varios
// municipios en orden de código INE, y que entre los dos alcanzan a todos los
// municipios de las fuentes (FR-013, FR-014, FR-015, data-model §2.7).
func TestRegistro(t *testing.T) {
	t.Parallel()

	registro := cargarSintetico(t)

	t.Run("por código", func(t *testing.T) {
		t.Parallel()

		municipio, esta := registro.municipio(codigoINE(t, "28991"))
		require.True(t, esta)
		assert.Equal(t, "Villaprueba", municipio.nombre)
		assert.Equal(t, "5", municipio.digito)
		assert.Equal(t, "28", municipio.provincia.codigo)
		assert.Equal(t, "01", municipio.provincia.comunidad.codigo)

		_, esta = registro.municipio(codigoINE(t, "28999"))
		assert.False(t, esta, "un código bien formado que no está en la relación no está en el registro")
	})

	t.Run("por forma plegada", func(t *testing.T) {
		t.Parallel()

		casos := map[string][]string{
			"villaprueba":            {"28991"},
			"rozas de prueba las":    {"28992"},
			"las rozas de prueba":    {"28992"},
			"iruneta pamploneta":     {"31991"},
			"iruneta":                {"31991"},
			"pamploneta":             {"31991"},
			"peniscola del rio":      {"47992"},
			"castro el":              {"05992"},
			"el castro":              {"05992", "47993"},
			"villanueva":             {"05991", "47991"},
			"no esta en la relacion": {},
		}

		for forma, codigos := range casos {
			assert.Equal(t, codigos, codigosDe(registro.candidatos(forma)), "candidatos de %q", forma)
		}
	})

	t.Run("candidatos en orden de código", func(t *testing.T) {
		t.Parallel()

		// El orden no depende del de las fuentes: se indexan al revés y los
		// candidatos siguen saliendo por código INE.
		municipios := make([]*municipioRegistrado, 0, len(registro.porCodigo))
		for _, municipio := range registro.porCodigo {
			municipios = append(municipios, municipio)
		}

		slices.SortFunc(municipios, func(a, b *municipioRegistrado) int {
			return strings.Compare(b.codigo.String(), a.codigo.String())
		})

		invertido := &Registro{}
		invertido.indexar(municipios)

		assert.Equal(t, []string{"05991", "47991"}, codigosDe(invertido.candidatos("villanueva")))
		assert.Equal(t, []string{"05992", "47993"}, codigosDe(invertido.candidatos("el castro")))
	})

	t.Run("alcanza a todos los municipios", func(t *testing.T) {
		t.Parallel()

		filas := ficherosSinteticos().Municipios.Municipios
		require.Len(t, registro.porCodigo, len(filas))

		for codigo, fila := range filas {
			municipio, esta := registro.municipio(codigoINE(t, codigo))
			require.True(t, esta, "el municipio %s no está en el índice por código", codigo)
			assert.Contains(t, registro.candidatos(Plegar(fila.Nombre)), municipio,
				"el municipio %s no se alcanza por su nombre oficial %q", codigo, fila.Nombre)
		}
	})

	t.Run("fecha de la relación", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, fecha(t, fechaDeLaRelacion), registro.FechaDeLaRelacion())
	})
}

// TestMunicipiosDe fija los municipios de una comunidad que el registro da a
// los adaptadores de los boletines (FR-022, research.md D13): solo los de esa
// comunidad, en orden de código INE, cada uno con su nombre oficial y las
// formas plegadas por las que lo alcanza una consulta. Dos municipios que se
// llaman igual en dos comunidades no se mezclan.
func TestMunicipiosDe(t *testing.T) {
	t.Parallel()

	// «Castro, El», de la 03 y con el artículo pospuesto, tiene en la 01 un
	// municipio que se llama igual.
	ficheros := ficherosSinteticos()
	ficheros.Municipios.Municipios["28993"] = FilaDeMunicipio{
		DC: "8", Nombre: "Castro, El", Provincia: "28", Comunidad: "01",
	}
	registro := cargar(t, ficheros)

	castro := []string{"castro el", "el castro"}
	casos := map[string][]MunicipioConFormas{
		"01": {
			{Codigo: codigoINE(t, "28991"), Nombre: "Villaprueba", Formas: []string{"villaprueba"}},
			{
				Codigo: codigoINE(t, "28992"), Nombre: "Rozas de Prueba, Las",
				Formas: []string{"rozas de prueba las", "las rozas de prueba"},
			},
			{Codigo: codigoINE(t, "28993"), Nombre: "Castro, El", Formas: castro},
		},
		"02": {
			{
				Codigo: codigoINE(t, "31991"), Nombre: "Iruñeta/Pamploneta",
				Formas: []string{"iruneta pamploneta", "iruneta", "pamploneta"},
			},
		},
		"03": {
			{Codigo: codigoINE(t, "05991"), Nombre: "Villanueva", Formas: []string{"villanueva"}},
			{Codigo: codigoINE(t, "05992"), Nombre: "Castro, El", Formas: castro},
			{Codigo: codigoINE(t, "47991"), Nombre: "Villanueva", Formas: []string{"villanueva"}},
			{Codigo: codigoINE(t, "47992"), Nombre: "Peñíscola del Río", Formas: []string{"peniscola del rio"}},
			{Codigo: codigoINE(t, "47993"), Nombre: "El Castro", Formas: []string{"el castro"}},
		},
	}

	t.Run("los de cada comunidad", func(t *testing.T) {
		t.Parallel()

		for comunidad, municipios := range casos {
			assert.Equal(t, municipios, registro.MunicipiosDe(comunidad), "municipios de la comunidad %s", comunidad)
		}
	})

	t.Run("comunidad que no está", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, registro.MunicipiosDe("04"))
		assert.Empty(t, registro.MunicipiosDe(""))
	})

	t.Run("el registro no cambia", func(t *testing.T) {
		t.Parallel()

		// Lo que haga quien los recibe con los municipios y sus formas no
		// llega al registro, que se usa desde varias goroutines a la vez.
		recibidos := registro.MunicipiosDe("01")
		recibidos[0].Nombre = "Cambiado"
		recibidos[1].Formas[0] = "cambiada"

		assert.Equal(t, casos["01"], registro.MunicipiosDe("01"))
	})
}

// TestComunidades fija las comunidades que el registro da a los adaptadores de
// los boletines (research.md D14, S2): cada una con su código, su nombre y el
// source de su fichero, en orden de código y no de nombre.
func TestComunidades(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []Comunidad{
		{Codigo: "01", Nombre: "Comunidad Uniprovincial", Source: fuenteDeLosNombres},
		{Codigo: "02", Nombre: "Comunidad Foral", Source: fuenteDeLosNombres},
		{Codigo: "03", Nombre: "Comunidad Pluriprovincial", Source: fuenteDeLosNombres},
	}, cargarSintetico(t).Comunidades())
}
