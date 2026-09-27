package instalacion_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// TestFormaSemVer fija la gramática de SemVer 2.0.0 con una v inicial
// opcional (data-model §8; research.md D31): núcleo X.Y.Z sin ceros a la
// izquierda, pre-release cuyos identificadores numéricos tampoco los llevan y
// metadatos de construcción, que sí pueden llevarlos. Lo que no la tiene es un
// binario de desarrollo, que no compara ni avisa (FR-073): `dev`, un hash
// abreviado, lo que tiene ceros a la izquierda en un número y cualquier otra
// cosa que no sea exactamente la gramática.
//
// Cada fila que no empieza por v se comprueba además con una v delante, que
// no cambia la respuesta: la v es opcional y solo una.
func TestFormaSemVer(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		version string
		forma   bool
	}{
		// El núcleo.
		{nombre: "la primera release", version: "0.1.0", forma: true},
		{nombre: "la primera release con v", version: "v0.1.0", forma: true},
		{nombre: "todo ceros", version: "0.0.0", forma: true},
		{nombre: "números de varias cifras", version: "10.20.30", forma: true},
		{
			nombre:  "números sin límite de tamaño",
			version: "99999999999999999999999.999999999999999999.99999999999999999",
			forma:   true,
		},

		// La pre-release.
		{nombre: "la del snapshot", version: "0.1.1-SNAPSHOT-abc1234", forma: true},
		{nombre: "la del snapshot con v", version: "v0.1.1-SNAPSHOT-abc1234", forma: true},
		{nombre: "pre-release alfanumérica", version: "1.0.0-alpha", forma: true},
		{nombre: "pre-release con varios identificadores", version: "1.0.0-alpha.1", forma: true},
		{nombre: "pre-release numérica", version: "1.0.0-0.3.7", forma: true},
		{nombre: "pre-release mezclada", version: "1.0.0-x.7.z.92", forma: true},
		{nombre: "pre-release de guiones", version: "1.0.0-x-y-z.--", forma: true},
		{nombre: "pre-release que es un cero", version: "1.0.0-0", forma: true},
		{nombre: "alfanumérico que empieza por cero", version: "1.0.0-0A.is.legal", forma: true},
		{
			nombre:  "pseudo-versión de Go",
			version: "v0.0.0-20260926120000-abcdef123456",
			forma:   true,
		},

		// Los metadatos de construcción.
		{nombre: "construcción tras el núcleo", version: "0.1.0+abc", forma: true},
		{nombre: "construcción con ceros a la izquierda", version: "1.0.0+001", forma: true},
		{nombre: "construcción de fecha", version: "1.0.0+20130313144700", forma: true},
		{nombre: "pre-release y construcción", version: "1.0.0-beta+exp.sha.5114f85", forma: true},
		{nombre: "construcción de guiones", version: "1.0.0+21AF26D3----117B344092BD", forma: true},
		{nombre: "guiones en todas partes", version: "1.2.3----RC-SNAPSHOT.12.9.1--.12+788", forma: true},

		// Un binario de desarrollo.
		{nombre: "dev", version: "dev", forma: false},
		{nombre: "la de go install sin versión", version: "(devel)", forma: false},
		{nombre: "un hash abreviado", version: "abc1234", forma: false},
		{nombre: "un hash abreviado de cifras", version: "5114f85", forma: false},
		{nombre: "vacía", version: "", forma: false},
		{nombre: "solo la v", version: "v", forma: false},

		// Ceros a la izquierda.
		{nombre: "cero a la izquierda en la mayor", version: "01.0.0", forma: false},
		{nombre: "cero a la izquierda en la menor", version: "0.01.0", forma: false},
		{nombre: "cero a la izquierda en el parche", version: "0.0.01", forma: false},
		{nombre: "cero a la izquierda con v", version: "v01.0.0", forma: false},
		{nombre: "cero a la izquierda en la pre-release", version: "1.0.0-01", forma: false},
		{nombre: "cero a la izquierda en otro identificador", version: "1.0.0-alpha.01", forma: false},

		// El núcleo incompleto o de más.
		{nombre: "solo la mayor", version: "1", forma: false},
		{nombre: "sin parche", version: "1.0", forma: false},
		{nombre: "cuatro números", version: "1.0.0.0", forma: false},
		{nombre: "número vacío", version: "1..0", forma: false},
		{nombre: "signo delante", version: "-1.0.0", forma: false},
		{nombre: "más delante", version: "+1.0.0", forma: false},

		// La v, que es una y minúscula.
		{nombre: "dos v", version: "vv0.1.0", forma: false},
		{nombre: "V mayúscula", version: "V0.1.0", forma: false},
		{nombre: "v con punto", version: "v.0.1.0", forma: false},

		// Pre-release y construcción mal formadas.
		{nombre: "guion sin pre-release", version: "1.0.0-", forma: false},
		{nombre: "más sin construcción", version: "1.0.0+", forma: false},
		{nombre: "identificador de pre-release vacío", version: "1.0.0-alpha..1", forma: false},
		{nombre: "identificador de construcción vacío", version: "1.0.0+build..1", forma: false},
		{nombre: "construcción que empieza por punto", version: "1.1.2+.123", forma: false},
		{nombre: "dos construcciones", version: "9.8.7+meta+meta", forma: false},
		{nombre: "guion bajo en la pre-release", version: "1.0.0-alpha_1", forma: false},
		{nombre: "pre-release y más vacío", version: "1.0.0-alpha+", forma: false},

		// Fuera de ASCII y blancos.
		{nombre: "cifra arábigo-índica", version: "\xd9\xa1.0.0", forma: false},
		{nombre: "letra griega en la pre-release", version: "1.0.0-\xce\xb1", forma: false},
		{nombre: "salto de línea final", version: "0.1.0\n", forma: false},
		{nombre: "blanco delante", version: " 0.1.0", forma: false},
		{nombre: "blanco detrás", version: "0.1.0 ", forma: false},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.forma, instalacion.FormaSemVer(caso.version), "FormaSemVer(%q)", caso.version)

			if !strings.HasPrefix(caso.version, "v") {
				conV := "v" + caso.version
				assert.Equal(t, caso.forma, instalacion.FormaSemVer(conV), "FormaSemVer(%q)", conV)
			}
		})
	}
}

// TestMismaVersion fija la regla de igualdad de FR-077, la misma para el
// aviso y para doctor: quitada a cada una una v inicial si la lleva, la misma
// cadena byte a byte, pre-release y metadatos de construcción incluidos. No
// mira la forma ni ordena, y es simétrica: cada fila se comprueba en los dos
// sentidos.
func TestMismaVersion(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		a, b   string
		misma  bool
	}{
		// Iguales.
		{nombre: "solo difieren en la v", a: "v0.1.0", b: "0.1.0", misma: true},
		{nombre: "las dos sin v", a: "0.1.0", b: "0.1.0", misma: true},
		{nombre: "las dos con v", a: "v0.1.0", b: "v0.1.0", misma: true},
		{nombre: "pre-release igual y la v", a: "v0.1.1-SNAPSHOT-abc1234", b: "0.1.1-SNAPSHOT-abc1234", misma: true},
		{nombre: "construcción igual y la v", a: "v0.1.0+abc", b: "0.1.0+abc", misma: true},
		{nombre: "sin forma pero la misma cadena", a: "dev", b: "dev", misma: true},

		// Distintas.
		{nombre: "metadatos de construcción", a: "0.1.0", b: "0.1.0+abc", misma: false},
		{nombre: "el snapshot siguiente", a: "0.1.0", b: "0.1.1-SNAPSHOT-abc1234", misma: false},
		{nombre: "otra construcción", a: "0.1.0+abc", b: "0.1.0+abd", misma: false},
		{nombre: "pre-release", a: "0.1.0-rc.1", b: "0.1.0", misma: false},
		{nombre: "mayúsculas en la pre-release", a: "0.1.1-SNAPSHOT-abc1234", b: "0.1.1-snapshot-abc1234", misma: false},
		{nombre: "otro parche", a: "v0.1.0", b: "0.1.1", misma: false},
		{nombre: "cero a la izquierda", a: "0.1.0", b: "0.01.0", misma: false},
		{nombre: "blanco detrás", a: "0.1.0", b: "0.1.0 ", misma: false},
		{nombre: "desarrollo frente a release", a: "dev", b: "0.1.0", misma: false},

		// Solo se quita una v, y minúscula.
		{nombre: "dos v frente a ninguna", a: "vv0.1.0", b: "0.1.0", misma: false},
		{nombre: "dos v frente a una", a: "vv0.1.0", b: "v0.1.0", misma: false},
		{nombre: "V mayúscula", a: "V0.1.0", b: "0.1.0", misma: false},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.misma, instalacion.MismaVersion(caso.a, caso.b), "MismaVersion(%q, %q)", caso.a, caso.b)
			assert.Equal(t, caso.misma, instalacion.MismaVersion(caso.b, caso.a), "MismaVersion(%q, %q)", caso.b, caso.a)
		})
	}
}
