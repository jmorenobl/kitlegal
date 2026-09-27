package instalacion

import (
	"regexp"
	"strings"
)

// Las piezas de la gramática de SemVer 2.0.0, las de la expresión regular que
// recomienda semver.org con \d escrito [0-9]: en RE2 [0-9] solo casa cifras
// ASCII, y un número de otra escritura no tiene forma.
const (
	// numeroSemVer es un número del núcleo o un identificador numérico de la
	// pre-release: sin ceros a la izquierda.
	numeroSemVer = `(?:0|[1-9][0-9]*)`
	// identificadorDePreRelease es numérico, o alfanumérico con al menos una
	// letra o un guion; solo el numérico no admite ceros a la izquierda.
	identificadorDePreRelease = `(?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)`
	// identificadorDeConstruccion admite ceros a la izquierda.
	identificadorDeConstruccion = `[0-9a-zA-Z-]+`
)

// formaSemVer es la gramática entera, con una v inicial opcional (research.md
// D31). `$` sin la bandera m es el final del texto, de modo que un salto de
// línea final no pasa.
var formaSemVer = regexp.MustCompile(`^v?` +
	numeroSemVer + `\.` + numeroSemVer + `\.` + numeroSemVer +
	`(?:-` + identificadorDePreRelease + `(?:\.` + identificadorDePreRelease + `)*)?` +
	`(?:\+` + identificadorDeConstruccion + `(?:\.` + identificadorDeConstruccion + `)*)?$`)

// FormaSemVer dice si version tiene la forma de SemVer 2.0.0, con una v
// minúscula inicial opcional: núcleo X.Y.Z sin ceros a la izquierda,
// pre-release y metadatos de construcción (data-model §8). Una versión sin
// ella es la de un binario de desarrollo, que no compara ni avisa (FR-073).
// Solo reconoce la forma: no ordena versiones ni dice cuál es más nueva.
func FormaSemVer(version string) bool {
	return formaSemVer.MatchString(version)
}

// MismaVersion dice si a y b son la misma versión según FR-077, la regla del
// aviso y de doctor: quitada a cada una una v inicial si la lleva, son la
// misma cadena byte a byte, pre-release y metadatos de construcción incluidos.
// Así v0.1.0 y 0.1.0 son la misma, y 0.1.0 y 0.1.0+abc no. No mira la forma:
// quien necesita que la tengan la comprueba antes con FormaSemVer.
func MismaVersion(a, b string) bool {
	return strings.TrimPrefix(a, "v") == strings.TrimPrefix(b, "v")
}
