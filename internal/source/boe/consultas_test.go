package boe_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// TestVerbosDeLasConsultas fija que cada una de las seis consultas del adaptador
// implementa core.Consulta, por valor y por puntero, y declara el verbo del
// applet boe que resuelve, el mismo esté vacía o lleve lo que se pide; y que los
// seis verbos son distintos, de modo que ninguna consulta pueda pasar por otra
// (contrato puerto-y-applet §3.2, research.md D2).
func TestVerbosDeLasConsultas(t *testing.T) {
	t.Parallel()

	const norma = "BOE-A-2015-10565"

	casos := []struct {
		verbo    string
		vacia    core.Consulta
		conDatos core.Consulta
	}{
		{
			verbo:    "buscar",
			vacia:    boe.ConsultaBuscar{},
			conDatos: &boe.ConsultaBuscar{Texto: []string{"ley", "general", "tributaria"}},
		},
		{verbo: "indice", vacia: boe.ConsultaIndice{}, conDatos: &boe.ConsultaIndice{Norma: norma}},
		{verbo: "articulo", vacia: boe.ConsultaArticulo{}, conDatos: &boe.ConsultaArticulo{Norma: norma, Bloque: "a21"}},
		{
			verbo:    "articulos",
			vacia:    boe.ConsultaArticulos{},
			conDatos: &boe.ConsultaArticulos{Norma: norma, Bloques: []string{"a21", "da3", "a21"}},
		},
		{verbo: "metadatos", vacia: boe.ConsultaMetadatos{}, conDatos: &boe.ConsultaMetadatos{Norma: norma}},
		{verbo: "analisis", vacia: boe.ConsultaAnalisis{}, conDatos: &boe.ConsultaAnalisis{Norma: norma}},
	}

	declarados := make([]string, 0, len(casos))

	for _, caso := range casos {
		declarados = append(declarados, caso.vacia.Verbo())

		t.Run(caso.verbo, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.verbo, caso.vacia.Verbo(), "el verbo de %T", caso.vacia)
			assert.Equal(t, caso.verbo, caso.conDatos.Verbo(), "el verbo de %T", caso.conDatos)
		})
	}

	slices.Sort(declarados)
	assert.Len(t, slices.Compact(declarados), len(casos), "dos consultas declaran el mismo verbo")
}
