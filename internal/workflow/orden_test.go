package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLosGuionesOrdenanPorBytes: en el run de H33 (adf73515), el primero en
// Linux con una tarea [datos], `grabar_datos` falló tres veces con la
// grabación ya hecha, porque con `en_US.UTF-8` GNU sort ordena las líneas de
// `git status --porcelain` sin mirar su prefijo y GNU comm da esa salida por
// desordenada y sale con 1. scripts/workflow/comun.sh fija la intercalación
// por bytes para todo guion que lo carga, también si quien lanza el run fija
// el idioma con LC_ALL, y deja el resto del idioma como estaba. La orden es la
// de scripts/workflow/grabar-datos.sh, que separa lo que deja la grabación de
// lo que ya había en el árbol.
func TestLosGuionesOrdenanPorBytes(t *testing.T) {
	t.Parallel()

	const guion = "scripts/nuevos.sh"

	a := nuevoArbol(t)
	a.escribir("internal/fuente/testdata/grabada.json", "antes\n")
	a.commit("la grabación anterior")
	a.escribirEjecutable(guion, `#!/usr/bin/env bash
set -euo pipefail
. scripts/workflow/comun.sh
antes=$(mktemp); git status --porcelain --untracked-files=all | sort > "$antes"
printf 'después\n' > internal/fuente/testdata/grabada.json
printf 'nueva\n' > internal/fuente/testdata/nueva.json
mkdir -p `+directorioDelHito+`/gates
printf 'registro\n' > `+directorioDelHito+`/gates/grabaciones.md
git status --porcelain --untracked-files=all | sort | comm -13 "$antes" - | cut -c4-
rm -f "$antes"
printf 'LANG=%s LC_ALL=%s\n' "${LANG:-}" "${LC_ALL:-}"
`)

	nuevos := "internal/fuente/testdata/grabada.json\ninternal/fuente/testdata/nueva.json\n" + directorioDelHito + "/gates/grabaciones.md\n"

	// Los dos casos escriben en el mismo árbol, uno detrás de otro: el árbol
	// es del test, y sus guiones fallan con el `t` del test, no con el de un
	// subtest.
	for _, idioma := range [][]string{
		{"LANG=en_US.UTF-8", "LC_ALL="},
		{"LANG=", "LC_ALL=en_US.UTF-8"},
	} {
		salida := a.debeSalirBien(idioma, guion)
		assert.Equalf(t, nuevos+"LANG=en_US.UTF-8 LC_ALL=\n", salida.estandar,
			"con %v, separa lo que deja la grabación, en orden de bytes, y conserva el idioma", idioma)
		assert.Emptyf(t, salida.errores, "con %v, comm no da ninguna entrada por desordenada", idioma)

		a.git("checkout", "-q", "--", "internal/fuente")
		a.git("clean", "-qfd", "--", "internal/fuente", directorioDelHito+"/gates")
	}
}
