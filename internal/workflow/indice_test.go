package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestElSupervisorEsperaAQueSeLibereElIndice: un paso que vence su timeout a
// mitad de un commit deja vivo un rato el git y su gancho de pre-commit, con
// el candado del índice puesto; en H41.2 el supervisor reanudó `hito_a_done`
// al segundo y gastó sus dos reanudaciones en «index.lock: File exists».
// Antes de reanudar, `esperar_indice` (scripts/workflow/comun.sh) espera a
// que el candado desaparezca, como mucho el plazo que se le da, y si sigue
// ahí sale con 1 sin borrarlo: puede ser de un git vivo.
func TestElSupervisorEsperaAQueSeLibereElIndice(t *testing.T) {
	t.Parallel()

	const (
		guion   = "scripts/esperar-indice.sh"
		candado = ".git/index.lock"
	)

	a := nuevoArbol(t)
	// El git que termina: borra el candado al cabo de un segundo, mientras
	// esperar_indice espera.
	a.escribirEjecutable(guion, `#!/usr/bin/env bash
set -euo pipefail
. scripts/workflow/comun.sh
if [ "${LIBERAR:-}" = 1 ]; then ( sleep 1; rm -f "$(git rev-parse --git-path index.lock)" ) & fi
esperar_indice "$1"
`)

	a.debeSalirBien(nil, guion, "0")

	a.escribir(candado, "")
	a.debeSalirBien([]string{"LIBERAR=1"}, guion, "10")
	assert.True(t, a.noExiste(candado), "espera hasta que el git que lo tenía lo suelta")

	a.escribir(candado, "")

	tomado := a.ejecutar(nil, guion, "1")
	assert.Equal(t, 1, tomado.codigo, "con el candado puesto al vencer el plazo, sale con 1")
	assert.True(t, a.existe(candado), "y no lo borra")
}
