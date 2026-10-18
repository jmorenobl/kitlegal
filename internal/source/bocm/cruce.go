package bocm

import (
	"slices"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/core/territorio"
)

// cruce lleva los nombres que escribe el boletín a los municipios de la
// comunidad del boletín, por su código INE y sin ninguna lista de alias ni
// ningún municipio con trato propio (FR-022, FR-024; research D13; principio
// IX): primero por la forma compacta de territorio, igual a la de una forma de
// un único municipio; si ninguna la tiene, por el prefijo, el nombre que
// empieza por todas las palabras de una forma de un único municipio seguidas
// de «de» o «del» («Braojos de la Sierra» es «Braojos»). Con ningún candidato o
// con más de uno no hay municipio.
type cruce struct {
	// porCompacta lleva de cada forma compacta a los códigos de los municipios
	// que tienen una forma con ella, sin repetir y en orden de código.
	porCompacta map[string][]string
	// municipios son los de la comunidad, en orden de código y con sus formas,
	// para la regla del prefijo.
	municipios []territorio.MunicipioConFormas
}

// nuevoCruce prepara el cruce con los municipios de la comunidad, los de
// territorio.Registro.MunicipiosDe.
func nuevoCruce(municipios []territorio.MunicipioConFormas) *cruce {
	c := &cruce{porCompacta: map[string][]string{}, municipios: municipios}

	for _, municipio := range municipios {
		codigo := municipio.Codigo.String()

		for _, forma := range municipio.Formas {
			compacta := territorio.FormaCompacta(forma)
			if compacta != "" && !slices.Contains(c.porCompacta[compacta], codigo) {
				c.porCompacta[compacta] = append(c.porCompacta[compacta], codigo)
			}
		}
	}

	return c
}

// municipioDe es el código INE del único municipio con el que casa el nombre
// de una línea, y si lo hay.
func (c *cruce) municipioDe(nombre string) (string, bool) {
	candidatos := c.porCompacta[territorio.FormaCompacta(nombre)]
	if len(candidatos) == 0 {
		candidatos = c.porPrefijo(nombre)
	}

	if len(candidatos) != 1 {
		return "", false
	}

	return candidatos[0], true
}

// porPrefijo son los códigos de los municipios una de cuyas formas es el
// principio del nombre plegado, seguida de «de» o «del».
func (c *cruce) porPrefijo(nombre string) []string {
	palabras := strings.Fields(territorio.Plegar(nombre))

	var candidatos []string

	for _, municipio := range c.municipios {
		if slices.ContainsFunc(municipio.Formas, func(forma string) bool { return empiezaPor(palabras, forma) }) {
			candidatos = append(candidatos, municipio.Codigo.String())
		}
	}

	return candidatos
}

// empiezaPor dice si las palabras del nombre empiezan por todas las de la forma
// plegada y sigue «de» o «del».
func empiezaPor(palabras []string, forma string) bool {
	deLaForma := strings.Fields(forma)
	if len(deLaForma) == 0 || len(palabras) <= len(deLaForma) || !slices.Equal(palabras[:len(deLaForma)], deLaForma) {
		return false
	}

	switch palabras[len(deLaForma)] {
	case "de", "del":
		return true
	default:
		return false
	}
}
