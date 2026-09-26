package instalacion_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Huellas de contenidos conocidos, los valores de referencia de SHA-256 de la
// cadena vacía, de «a», de «b», de «c» y de «abc». Los usan también los
// manifiestos de los tests: una huella del manifiesto es la de los bytes de un
// fichero.
const (
	huellaVacia = "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	huellaDeA   = "sha256:ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb"
	huellaDeB   = "sha256:3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d"
	huellaDeC   = "sha256:2e7d2c03a9507ae265ecf5b5356885a53393a2029d241394997265a1a25aefc6"
	huellaDeABC = "sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
)

// TestHuellaDe fija la huella de un fichero empotrado (data-model §1): el
// prefijo del algoritmo y los 64 hexadecimales en minúscula del SHA-256 de sus
// bytes, la misma forma que el hash del sobre (contracts/manifiesto.md §1).
func TestHuellaDe(t *testing.T) {
	t.Parallel()

	casos := []struct {
		contenido string
		huella    string
	}{
		{contenido: "", huella: huellaVacia},
		{contenido: "a", huella: huellaDeA},
		{contenido: "b", huella: huellaDeB},
		{contenido: "c", huella: huellaDeC},
		{contenido: "abc", huella: huellaDeABC},
	}

	formaDelSobre := regexp.MustCompile(schema.PatronHuella)

	for _, caso := range casos {
		huella := instalacion.HuellaDe([]byte(caso.contenido))
		assert.Equal(t, caso.huella, huella, "HuellaDe(%q)", caso.contenido)
		assert.Regexp(t, formaDelSobre, huella, "la huella de %q no tiene la forma del hash del sobre", caso.contenido)
	}

	assert.Equal(t, huellaVacia, instalacion.HuellaDe(nil), "sin contenido es la huella de la cadena vacía")
}

// TestNuevoFicheroEmpotrado exige que un fichero empotrado lleve su ruta, sus
// bytes y la huella de esos bytes, y que la huella siga siendo la de su
// contenido aunque quien lo construyó cambie después los bytes que pasó: el
// fichero guarda una copia (FR-003).
func TestNuevoFicheroEmpotrado(t *testing.T) {
	t.Parallel()

	contenido := []byte("abc")
	fichero := instalacion.NuevoFicheroEmpotrado("references/normas.md", contenido)

	assert.Equal(t, "references/normas.md", fichero.Ruta)
	assert.Equal(t, []byte("abc"), fichero.Contenido)
	assert.Equal(t, huellaDeABC, fichero.Huella)

	contenido[0] = 'x'

	assert.Equal(t, []byte("abc"), fichero.Contenido, "los bytes del fichero cambian con los de quien lo construyó")
	assert.Equal(t, instalacion.HuellaDe(fichero.Contenido), fichero.Huella)

	vacio := instalacion.NuevoFicheroEmpotrado("SKILL.md", nil)
	assert.Empty(t, vacio.Contenido)
	assert.Equal(t, huellaVacia, vacio.Huella)
}
