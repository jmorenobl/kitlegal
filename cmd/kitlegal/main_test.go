package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cli"
)

// registroVacio es la lista de applets que escribe el kernel cuando no hay
// ninguno registrado: la evidencia observable de que el binario que se publica
// no registra ningún applet en H1 —el primer applet de kitlegal llega en H4—
// (contracts/registro-y-describe.md §3).
const registroVacio = "este binario no registra ningún applet"

// TestPuntoDeEntrada ejerce el contrato observable del binario distribuido con la
// **misma composición que main()** —el registro de producción y los datos de
// construcción de este paquete— y dos escritores en memoria en lugar de los
// descriptores del sistema, que es lo que hace comprobable el contrato entero sin
// lanzar ningún subproceso (FR-035).
//
// De contracts/cli-version.md de H0 se conserva exactamente la parte que no
// cambia: las tres líneas de version, la salida de error vacía y el código 0
// (D16). Lo que H0 resolvía con la línea «uso: kitlegal version» y el código 2
// dejó de ser contrato de version: un nombre que no es ningún applet lo resuelve
// el despacho, que **nombra lo desconocido** y enumera lo disponible —nada, en
// este binario— con código 2 (FR-006, contracts/registro-y-describe.md §2 y §3).
//
// No es paralelo, y no es un descuido: fija KITLEGAL_LOG —en el test y en cada
// subcaso, que es lo que lo deja hermético por separado— para que el nivel del
// registro de eventos no dependa del entorno de quien ejecuta los tests, y eso
// es estado del proceso entero.
func TestPuntoDeEntrada(t *testing.T) {
	t.Setenv(cli.VariableNivel, "")

	// La salida de version se compone a partir de las variables del paquete y no
	// de un literal, de modo que el caso siga siendo válido cuando -ldflags las
	// inyecte (FR-004).
	salidaVersion := "kitlegal " + version + "\ncommit: " + commit + "\nfecha:  " + fecha + "\n"

	casos := []struct {
		nombre string
		argv   []string
		salida string
		// errores son los fragmentos que el mensaje para la persona tiene que
		// llevar; vacío significa que la salida de error queda vacía.
		errores []string
		codigo  int
	}{
		{
			nombre: "version escribe los tres datos y termina con 0",
			argv:   []string{"kitlegal", "version"},
			salida: salidaVersion,
			codigo: 0,
		},
		{
			// Lo que H0 resolvía con el código 2. El verbo reservado se reconoce
			// **antes** que el registro, así que aquí no hay ningún applet
			// desconocido que nombrar y version se atiende sin mirar lo que
			// sobra: sin sobre y sin banderas (contracts/registro-y-describe.md
			// §2, D16). Este caso fija ese comportamiento, que el contrato no
			// nombraba (gates/tasks.json, observación no bloqueante).
			nombre: "un argumento de más no altera version: el verbo reservado se atiende igual",
			argv:   []string{"kitlegal", "version", "extra"},
			salida: salidaVersion,
			codigo: 0,
		},
		{
			nombre:  "sin applet, el fallo enumera lo que hay y termina con 2",
			argv:    []string{"kitlegal"},
			errores: []string{"no se ha indicado ningún applet", registroVacio},
			codigo:  2,
		},
		{
			nombre:  "un applet que no existe se nombra en el fallo y termina con 2",
			argv:    []string{"kitlegal", "inventado"},
			errores: []string{`"inventado"`, registroVacio},
			codigo:  2,
		},
		{
			// El applet de ejemplo vive en el binario que compila el test e2e y
			// nunca en el que se publica: sobre el distribuido es un nombre
			// desconocido como cualquier otro
			// (contracts/registro-y-describe.md §3).
			nombre:  "echo no es del binario distribuido, sino del binario del e2e",
			argv:    []string{"kitlegal", "echo", "hola"},
			errores: []string{`"echo"`, registroVacio},
			codigo:  2,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Setenv(cli.VariableNivel, "")

			var salida, errores bytes.Buffer

			codigo := app.Main(caso.argv, app.RegistroDeProduccion(),
				&salida, &errores, version, commit, fecha)

			assert.Equal(t, caso.codigo, codigo, "código de salida")
			assert.Equal(t, caso.salida, salida.String(), "salida estándar")

			if len(caso.errores) == 0 {
				assert.Empty(t, errores.String(), "la salida de error queda vacía")
			}

			for _, fragmento := range caso.errores {
				assert.Contains(t, errores.String(), fragmento,
					"el mensaje para la persona dice qué ha pasado")
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
