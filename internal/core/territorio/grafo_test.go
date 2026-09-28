package territorio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// casoDeObservado es un territorio sintético resuelto y lo que tiene que
// observar del mundo.
type casoDeObservado struct {
	nombre   string
	registro *Registro
	consulta string
	// dir3 es el código DIR3 que trae la respuesta, o vacío si no trae ninguno:
	// la premisa del caso.
	dir3     string
	esperado schema.Observado
}

// TestObservadoDeTerritorio fija lo que observa del mundo un territorio
// resuelto, sobre las fuentes sintéticas del paquete (contracts/emision.md §2;
// research.md D21; FR-043, FR-044, FR-045, FR-065):
//
//   - el Municipio, por «ine:» y su código INE de cinco cifras, con ese código
//     y su nombre oficial, siempre;
//   - si la respuesta trae el DIR3 de su ayuntamiento, el Organo por ese DIR3,
//     con él en sus datos, y la arista lb:pertenece_a del Organo al Municipio;
//   - sin DIR3, solo el Municipio, sin Organo ni arista: ningún municipio de los
//     datos congelados carece de DIR3, así que solo lo prueban estas fuentes
//     (research.md V28);
//   - ninguna vigencia, porque territorio no consulta ninguna fuente en
//     ejecución;
//   - y lo mismo para un municipio cubierto que para uno no cubierto: la
//     cobertura que declara el territorio no cambia lo que se registra.
//
// Los ids, los tipos, la relación y las claves de los datos van escritos
// enteros, y no con las constantes de internal/core/grafo, para que un cambio en
// ellos no pase por aquí en silencio.
func TestObservadoDeTerritorio(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeObservado(t) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			caso.comprueba(t)
		})
	}

	t.Run("la-cobertura-no-cambia-lo-observado", func(t *testing.T) {
		t.Parallel()

		compruebaObservadoSinCobertura(t)
	})
}

// casosDeObservado son un municipio cubierto y uno no cubierto con DIR3, uno
// cubierto sin él, y el mismo cubierto con su DIR3 quitado de la
// correspondencia, que solo observa su Municipio.
func casosDeObservado(t *testing.T) []casoDeObservado {
	t.Helper()

	sinSuDIR3 := ficherosSinteticos()
	delete(sinSuDIR3.DIR3.Correspondencia, "28991")

	registro := cargarSintetico(t)

	return []casoDeObservado{
		{
			nombre:   "cubierto-con-dir3",
			registro: registro,
			consulta: "Villaprueba",
			dir3:     "L01289915",
			esperado: schema.Observado{Operaciones: []schema.Operacion{
				schema.Nodo{ID: "ine:28991", Tipo: "Municipio", Datos: map[string]any{
					"codigo_ine": "28991", "nombre": "Villaprueba",
				}},
				schema.Nodo{ID: "L01289915", Tipo: "Organo", Datos: map[string]any{"dir3": "L01289915"}},
				schema.Arista{Origen: "L01289915", Relacion: "lb:pertenece_a", Destino: "ine:28991"},
			}},
		},
		{
			nombre:   "no-cubierto-con-dir3",
			registro: registro,
			consulta: "319913",
			dir3:     "L01319913",
			esperado: schema.Observado{Operaciones: []schema.Operacion{
				schema.Nodo{ID: "ine:31991", Tipo: "Municipio", Datos: map[string]any{
					"codigo_ine": "31991", "nombre": "Iru\xc3\xb1eta/Pamploneta",
				}},
				schema.Nodo{ID: "L01319913", Tipo: "Organo", Datos: map[string]any{"dir3": "L01319913"}},
				schema.Arista{Origen: "L01319913", Relacion: "lb:pertenece_a", Destino: "ine:31991"},
			}},
		},
		{
			nombre:   "cubierto-sin-dir3",
			registro: registro,
			consulta: "28992",
			esperado: schema.Observado{Operaciones: []schema.Operacion{
				schema.Nodo{ID: "ine:28992", Tipo: "Municipio", Datos: map[string]any{
					"codigo_ine": "28992", "nombre": "Rozas de Prueba, Las",
				}},
			}},
		},
		{
			nombre:   "el-mismo-sin-su-dir3",
			registro: cargar(t, sinSuDIR3),
			consulta: "Villaprueba",
			esperado: schema.Observado{Operaciones: []schema.Operacion{
				schema.Nodo{ID: "ine:28991", Tipo: "Municipio", Datos: map[string]any{
					"codigo_ine": "28991", "nombre": "Villaprueba",
				}},
			}},
		},
	}
}

// comprueba resuelve la consulta del caso, exige su premisa —el DIR3 que trae
// la respuesta— y lo observado entero, que cada llamada devuelve nuevo: cambiar
// las operaciones o los datos que devuelve una no cambia los de la siguiente.
func (caso casoDeObservado) comprueba(t *testing.T) {
	t.Helper()

	resuelto := resolver(t, caso.registro, caso.consulta)
	require.Equal(t, caso.dir3, resuelto.DIR3.Codigo, "la premisa: el DIR3 que trae la respuesta")

	observado := resuelto.Observado()
	assert.Equal(t, caso.esperado, observado)

	municipio, esNodo := observado.Operaciones[0].(schema.Nodo)
	require.True(t, esNodo, "la primera operación es el Municipio")

	municipio.Datos["nombre"] = "Otro nombre"
	observado.Operaciones[0] = schema.Arista{}

	assert.Equal(t, caso.esperado, resuelto.Observado(), "lo que devuelve una llamada es suyo")
}

// municipiosDeComunidadesConfiguradas es cuántos municipios sintéticos son de
// una comunidad con algún boletín configurado: los dos de la 01 y los cinco de
// la 03.
const municipiosDeComunidadesConfiguradas = 7

// compruebaObservadoSinCobertura resuelve cada municipio sintético con las
// fuentes tal cual y con ninguna comunidad configurada: la cobertura de los
// municipios de las comunidades configuradas cambia, y lo observado de todos, no
// (FR-045).
func compruebaObservadoSinCobertura(t *testing.T) {
	t.Helper()

	sinConfigurar := ficherosSinteticos()
	for codigo := range sinConfigurar.Comunidades {
		cambiarComunidad(&sinConfigurar, codigo, func(comunidad *FicheroDeComunidad) { comunidad.Boletines = nil })
	}

	configurados, noConfigurados := cargarSintetico(t), cargar(t, sinConfigurar)
	coberturasCambiadas := 0

	for codigo := range ficherosSinteticos().Municipios.Municipios {
		cubierto, noCubierto := resolver(t, configurados, codigo), resolver(t, noConfigurados, codigo)

		if cubierto.Cobertura != noCubierto.Cobertura {
			coberturasCambiadas++
		}

		assert.Equal(t, cubierto.Observado(), noCubierto.Observado(),
			"%s: lo mismo con su comunidad configurada y sin configurar", codigo)
	}

	assert.Equal(t, municipiosDeComunidadesConfiguradas, coberturasCambiadas,
		"la premisa: sin configurar, la cobertura de cada municipio de una comunidad configurada cambia")
}
