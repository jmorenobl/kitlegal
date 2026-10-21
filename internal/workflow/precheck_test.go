package workflow

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// rutaDeLaConstitucion es la de la constitución, en el proyecto y en el
	// árbol temporal.
	rutaDeLaConstitucion = ".specify/memory/constitution.md"

	// defectoDelConstitutionCheck es el comienzo del defecto que precheck.sh
	// plan da cuando un principio no tiene su fila.
	defectoDelConstitutionCheck = "Constitution Check sin fila para los principios: "
)

// encabezadoDePrincipio es el de cada principio de la constitución,
// «### <numeral>. <nombre>»; su primer grupo es el numeral.
var encabezadoDePrincipio = regexp.MustCompile(`(?m)^### ([IVXLC]+)\. `)

// TestElConstitutionCheckLlevaUnaFilaPorPrincipio: precheck.sh plan da como
// defecto cada principio de la constitución que no tiene su fila en el
// Constitution Check del plan. Los principios salen de la constitución, no de
// una lista del guion: con el principio X (ADR 0044) el prompt del plan seguía
// pidiendo «un ítem por principio (I-IX)». Cuentan las tres formas de fila de
// los planes de H0 a H26.1 («**I**», «I · …» e «I»); ni «II» ni una fila que
// empieza por una palabra, como «Implementación…» o «Validación…», cuentan como
// el principio I o el V.
func TestElConstitutionCheckLlevaUnaFilaPorPrincipio(t *testing.T) {
	t.Parallel()

	textoDeLaConstitucion := leerDelRepositorio(t, rutaDeLaConstitucion)

	var principios []string
	for _, encabezado := range encabezadoDePrincipio.FindAllStringSubmatch(textoDeLaConstitucion, -1) {
		principios = append(principios, encabezado[1])
	}

	require.GreaterOrEqual(t, len(principios), 2, "la constitución tiene sus principios con encabezado «### <numeral>.»")

	primero, ultimo := principios[0], principios[len(principios)-1]

	casos := []struct {
		nombre string
		filas  []string
		faltan string
	}{
		{nombre: "todas", filas: principios},
		{nombre: "sin la última", filas: principios[:len(principios)-1], faltan: ultimo},
		{nombre: "sin la primera, que ni II ni «Implementación» suplen", filas: principios[1:], faltan: primero},
		{nombre: "sin ninguna", faltan: strings.Join(principios, " ")},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			a := nuevoArbol(t)
			a.escribir(rutaDeLaConstitucion, textoDeLaConstitucion)
			a.escribir(directorioDelHito+"/plan.md", planConFilas(caso.filas))

			a.debeSalirBien(nil, "scripts/workflow/precheck.sh", "plan")

			var defectos []string
			for linea := range strings.SplitSeq(a.leer(gatesDelHito+"/plan-precheck.txt"), "\n") {
				if strings.HasPrefix(linea, defectoDelConstitutionCheck) {
					defectos = append(defectos, linea)
				}
			}

			if caso.faltan == "" {
				assert.Empty(t, defectos, "con una fila por principio no hay defecto")

				return
			}

			require.Len(t, defectos, 1, "un solo defecto, con todos los principios que faltan")
			assert.True(t, strings.HasPrefix(defectos[0], defectoDelConstitutionCheck+caso.faltan+" ("),
				"nombra exactamente los que faltan: %q", defectos[0])
		})
	}
}

// planConFilas es un plan cuyo Constitution Check tiene una fila por cada
// principio dado, en las tres formas de los planes anteriores por turno, y dos
// filas que no son de ningún principio: una que empieza por una palabra con las
// letras de un numeral, y otra con un numeral en la sección siguiente.
func planConFilas(principios []string) string {
	formas := []func(string) string{
		func(p string) string { return "| **" + p + "** | Principio | Cómo lo cumple | Cumple |" },
		func(p string) string { return "| " + p + " · Principio | Cómo lo cumple |" },
		func(p string) string { return "| " + p + " | Cómo lo cumple |" },
	}

	lineas := []string{"# Plan", "", "## Constitution Check", "", "| # | Principio | Cómo lo cumple |", "|---|---|---|"}
	for i, p := range principios {
		lineas = append(lineas, formas[i%len(formas)](p))
	}

	lineas = append(lineas, "| Implementación por fases, Validación y Ley | No es ningún principio |")

	lineas = append(lineas, "", "## Otra sección", "", "| I | No es del Constitution Check |", "")

	return strings.Join(lineas, "\n")
}
