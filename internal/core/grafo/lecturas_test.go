package grafo_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// redaccionDeEjemplo es una redacción de un bloque de la norma de ejemplo: la
// parte que la tiene, su fecha de vigencia y el cuerpo cuya huella lleva su id.
type redaccionDeEjemplo struct {
	parte         string
	fechaVigencia string
	cuerpo        string
}

// Las redacciones del bloque a21 de la secuencia de FR-025 (data-model §4) —A,
// la de la grabación de H4; B, la derivada version-posterior; y C, la derivada
// version-ulterior—, cuyos ids quedan en ese orden comparando bytes, y una del
// bloque a22.
var (
	redaccionA   = redaccionDeEjemplo{parte: "a21", fechaVigencia: "20161002", cuerpo: cuerpoA21}
	redaccionB   = redaccionDeEjemplo{parte: "a21", fechaVigencia: "20250101", cuerpo: cuerpoA21 + " Version posterior."}
	redaccionC   = redaccionDeEjemplo{parte: "a21", fechaVigencia: "20260101", cuerpo: cuerpoA21 + " Version ulterior."}
	redaccionA22 = redaccionDeEjemplo{parte: "a22", fechaVigencia: "20161002", cuerpo: "Articulo 22."}
)

// bloque es el id del bloque de la redacción, con la forma que emite boe
// (contracts/emision.md §1 de H7).
func (r redaccionDeEjemplo) bloque() string {
	return idNorma + "#" + r.parte
}

// id es el de la redacción, con la forma que emite boe.
func (r redaccionDeEjemplo) id() string {
	return r.bloque() + "@" + r.fechaVigencia + ":" + huellaDe(r.cuerpo)
}

// datos son los identificativos de la redacción.
func (r redaccionDeEjemplo) datos() map[string]any {
	return map[string]any{grafo.DatoFechaVigencia: r.fechaVigencia, grafo.DatoHashTexto: huellaDe(r.cuerpo)}
}

// operaciones son las que emite boe del artículo que la devuelve
// (internal/source/boe/grafo.go): la norma, el bloque, la redacción, las
// aristas eli:has_part y eli:has_version, y el texto.
func (r redaccionDeEjemplo) operaciones() []schema.Operacion {
	return []schema.Operacion{
		schema.Nodo{ID: idNorma, Tipo: grafo.TipoNorma, Datos: map[string]any{grafo.DatoIdentificador: "BOE-A-2015-10565"}},
		schema.Nodo{ID: r.bloque(), Tipo: grafo.TipoBloque, Datos: map[string]any{grafo.DatoBloque: r.parte}},
		schema.Nodo{ID: r.id(), Tipo: grafo.TipoBloqueVersion, Datos: r.datos()},
		schema.Arista{Origen: idNorma, Relacion: grafo.RelacionTieneParte, Destino: r.bloque()},
		schema.Arista{Origen: r.bloque(), Relacion: grafo.RelacionTieneVersion, Destino: r.id()},
		schema.Texto{Huella: huellaDe(r.cuerpo), Cuerpo: r.cuerpo},
	}
}

// vistaEl es la redacción como nodo de una instantánea, observada por última
// vez en fecha.
func (r redaccionDeEjemplo) vistaEl(fecha string) grafo.NodoDeInstantanea {
	return grafo.NodoDeInstantanea{
		ID: r.id(), Tipo: grafo.TipoBloqueVersion, Datos: r.datos(),
		UltimaObservacion: grafo.Procedencia{Fuente: fuenteBOE, URL: urlA21, FechaConsulta: fecha},
	}
}

// loteQueDevuelve es el lote de una invocación de boe articulo o boe articulos
// que devuelve estas redacciones, en este orden, con la procedencia y la
// vigencia del lote de ejemplo.
func loteQueDevuelve(redacciones ...redaccionDeEjemplo) core.Lote {
	lote := loteDeEjemplo()
	lote.Operaciones = nil

	for _, redaccion := range redacciones {
		lote.Operaciones = append(lote.Operaciones, redaccion.operaciones()...)
	}

	return lote
}

// TestLecturas fija qué es una lectura en un lote (data-model §2; research.md
// D2; FR-020): una por arista eli:has_version del lote consolidado, del Bloque
// que la origina a la BloqueVersion a la que llega, en el orden de sus claves
// comparando bytes, sin que ningún emisor tenga que decir nada más.
func TestLecturas(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		lote      core.Lote
		esperadas []grafo.Lectura
	}{
		{
			"boe articulo: una",
			loteQueDevuelve(redaccionA),
			[]grafo.Lectura{{Bloque: redaccionA.bloque(), Version: redaccionA.id()}},
		},
		{
			"boe articulos con dos bloques, pedidos a22 y a21: dos, por su clave",
			loteQueDevuelve(redaccionA22, redaccionB),
			[]grafo.Lectura{
				{Bloque: redaccionB.bloque(), Version: redaccionB.id()},
				{Bloque: redaccionA22.bloque(), Version: redaccionA22.id()},
			},
		},
		{"un lote sin eli:has_version: ninguna", loteDeEjemplo(), nil},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			consolidado, err := grafo.Consolidar(caso.lote)
			require.NoError(t, err)
			assert.Equal(t, caso.esperadas, consolidado.Lecturas())
		})
	}
}

// TestLeida fija la fila de lecturas del bloque a21 tras cada paso de la
// secuencia de FR-025 (data-model §4), desde (A, A), la de una primera lectura
// que vio A: cada lectura desplaza la última a la anterior, una que ve lo
// mismo que las dos anteriores deja la fila como estaba, y un paso sin lectura
// —con --no-graph— no la toca (FR-021).
func TestLeida(t *testing.T) {
	t.Parallel()

	fila := func(ultima, anterior redaccionDeEjemplo) grafo.LecturasDeBloque {
		return grafo.LecturasDeBloque{Bloque: idBloque, Ultima: ultima.id(), Anterior: anterior.id()}
	}

	pasos := []struct {
		nombre   string
		lee      string // el id de la redacción que ve la lectura; "" si no hay lectura
		esperada grafo.LecturasDeBloque
	}{
		{"1: la fuente sirve A", redaccionA.id(), fila(redaccionA, redaccionA)},
		{"2: caducada la cache, la fuente sirve B", redaccionB.id(), fila(redaccionB, redaccionA)},
		{"3: con --no-graph no hay lectura", "", fila(redaccionB, redaccionA)},
		{"4: la cache sirve B", redaccionB.id(), fila(redaccionB, redaccionB)},
		{"4 otra vez: la fila no cambia", redaccionB.id(), fila(redaccionB, redaccionB)},
		{"5: caducada la cache, la fuente sirve C", redaccionC.id(), fila(redaccionC, redaccionB)},
	}

	actual := fila(redaccionA, redaccionA)

	for _, paso := range pasos {
		if paso.lee != "" {
			actual = actual.Leida(paso.lee)
		}

		assert.Equal(t, paso.esperada, actual, "paso %s", paso.nombre)
	}
}

// TestRedaccionVistaSinLecturas fija la redacción vista de un bloque sin fila
// de lecturas (data-model §3; research.md D3; FR-026): entre sus BloqueVersion,
// la de última observación más reciente, comparando los instantes de sus
// fechas de consulta y no su texto, y a igualdad de instante la de id menor
// comparando bytes. El orden en que llegan no cuenta, y sin ninguna no hay
// redacción vista.
func TestRedaccionVistaSinLecturas(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		versiones []grafo.NodoDeInstantanea
		vista     string
	}{
		{
			"gana el instante mayor, aunque su id sea mayor y el texto de su fecha menor",
			[]grafo.NodoDeInstantanea{redaccionA.vistaEl(lunesEnMadrid), redaccionB.vistaEl(lunesYMedio)},
			redaccionB.id(),
		},
		{
			"con el mismo instante, el id menor, aunque el texto de su fecha sea menor",
			[]grafo.NodoDeInstantanea{redaccionA.vistaEl(lunes), redaccionB.vistaEl(lunesEnMadrid)},
			redaccionA.id(),
		},
		{
			"entre tres, la observada la ultima",
			[]grafo.NodoDeInstantanea{redaccionB.vistaEl(martes), redaccionA.vistaEl(lunes), redaccionC.vistaEl(domingo)},
			redaccionB.id(),
		},
		{"un solo candidato", []grafo.NodoDeInstantanea{redaccionA.vistaEl(lunes)}, redaccionA.id()},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			invertidas := slices.Clone(caso.versiones)
			slices.Reverse(invertidas)

			for _, versiones := range [][]grafo.NodoDeInstantanea{caso.versiones, invertidas} {
				vista, hay, err := grafo.RedaccionVistaSinLecturas(versiones)
				require.NoError(t, err)
				assert.True(t, hay)
				assert.Equal(t, caso.vista, vista)
			}
		})
	}

	t.Run("sin ninguna BloqueVersion", func(t *testing.T) {
		t.Parallel()

		vista, hay, err := grafo.RedaccionVistaSinLecturas(nil)
		require.NoError(t, err)
		assert.False(t, hay)
		assert.Empty(t, vista)
	})
}
