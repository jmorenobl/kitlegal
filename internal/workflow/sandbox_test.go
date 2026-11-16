package workflow

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLaSondaDelSandboxPruebaElNamespaceAnidado: en Linux, tener bwrap y socat
// no basta para que arranque el sandbox de las sesiones del run. En delorean
// (Ubuntu 26.04, 2026-11-16) el perfil de AppArmor `bwrap-userns-restrict`
// niega toda capacidad a lo que bwrap ejecuta, y el filtro seccomp de Claude
// Code, que crea dentro un user namespace anidado, hacía fallar cada orden de
// Bash con la sesión ya abierta. `sandbox_anidado_disponible`
// (scripts/workflow/comun.sh), que hito.sh ejecuta antes de tomar el candado,
// pide a bwrap lo mismo sin modelo y sin red, y falla cuando bwrap falla.
func TestLaSondaDelSandboxPruebaElNamespaceAnidado(t *testing.T) {
	t.Parallel()

	const guion = "scripts/sonda-sandbox.sh"

	a := nuevoArbol(t)
	// El bwrap sustituto apunta sus argumentos y sale con lo que se le pida.
	a.escribirEjecutable("bin/bwrap", `#!/usr/bin/env bash
printf '%s\n' "$*" > argumentos-de-bwrap
exit "${BWRAP_SALIDA:-0}"
`)
	a.escribirEjecutable(guion, `#!/usr/bin/env bash
set -euo pipefail
. scripts/workflow/comun.sh
PATH="$PWD/bin:$PATH" sandbox_anidado_disponible
`)

	a.debeSalirBien(nil, guion)

	argumentos := a.leer("argumentos-de-bwrap")

	for _, pedido := range []string{"--unshare-user", "--unshare-net", "-- unshare -U -r --mount --pid --fork --mount-proc true"} {
		assert.Contains(t, strings.TrimSpace(argumentos), pedido, "la sonda pide a bwrap lo que necesita el filtro seccomp")
	}

	falla := a.ejecutar([]string{"BWRAP_SALIDA=1"}, guion)
	assert.NotZero(t, falla.codigo, "si bwrap no puede crear el namespace anidado, la sonda falla")
}
