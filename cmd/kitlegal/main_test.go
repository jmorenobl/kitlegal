package main

import (
	"bytes"
	"testing"
)

// TestRun ejerce el contrato de contracts/cli-version.md con escritores en
// memoria: los tres datos y el código 0 del verbo version, y el código 2 con
// línea de uso en stderr ante verbo ausente, desconocido o argumento sobrante.
func TestRun(t *testing.T) {
	t.Parallel()

	// La salida esperada del verbo version se compone aquí a partir de las
	// variables de paquete, no de la constante de formato del código, de modo
	// que el caso siga siendo válido cuando -ldflags las inyecte (FR-004).
	salidaVersion := "kitlegal " + version + "\ncommit: " + commit + "\nfecha:  " + fecha + "\n"

	casos := []struct {
		nombre string
		args   []string
		stdout string
		stderr string
		codigo int
	}{
		{
			nombre: "version imprime los tres datos y termina con 0",
			args:   []string{"version"},
			stdout: salidaVersion,
			stderr: "",
			codigo: 0,
		},
		{
			nombre: "sin argumentos imprime el uso en stderr y termina con 2",
			args:   []string{},
			stdout: "",
			stderr: "uso: kitlegal version\n",
			codigo: 2,
		},
		{
			nombre: "verbo desconocido imprime el uso en stderr y termina con 2",
			args:   []string{"inventado"},
			stdout: "",
			stderr: "uso: kitlegal version\n",
			codigo: 2,
		},
		{
			nombre: "version con argumento sobrante termina con 2",
			args:   []string{"version", "extra"},
			stdout: "",
			stderr: "uso: kitlegal version\n",
			codigo: 2,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			codigo := run(caso.args, &stdout, &stderr)

			if codigo != caso.codigo {
				t.Errorf("código de salida = %d, se esperaba %d", codigo, caso.codigo)
			}
			if stdout.String() != caso.stdout {
				t.Errorf("stdout = %q, se esperaba %q", stdout.String(), caso.stdout)
			}
			if stderr.String() != caso.stderr {
				t.Errorf("stderr = %q, se esperaba %q", stderr.String(), caso.stderr)
			}
		})
	}
}

// TestValoresPorDefecto fija los valores que el código lleva cuando nadie los
// inyecta: son los que ven `go run` y `go test`, y los que un binario de
// make build o make install nunca debe mostrar (FR-004, data-model.md R1.2).
func TestValoresPorDefecto(t *testing.T) {
	t.Parallel()

	if version != "dev" || commit != "none" || fecha != "unknown" {
		t.Errorf("valores por defecto = %q, %q, %q; se esperaban %q, %q, %q",
			version, commit, fecha, "dev", "none", "unknown")
	}
}
