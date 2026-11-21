package instalacion

import (
	"cmp"
	"regexp"
	"slices"
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

// precedencia compara a y b, dos versiones con la forma de FormaSemVer, en la
// precedencia de SemVer 2.0.0 §11, la regla de versiones del aviso que
// reinstala (FR-004 de H38; research.md D6): negativa si a va antes que b,
// cero si tienen la misma y positiva si a va después. La v inicial y los
// metadatos de construcción no cuentan; el núcleo se compara número a número;
// una versión sin pre-release va después de la misma con pre-release; y dos
// pre-releases, identificador a identificador, con la que tiene más
// identificadores después si todos los anteriores son iguales. Quien la llama
// ha comprobado antes la forma de las dos: con otra cosa, el resultado no
// significa nada.
func precedencia(a, b string) int {
	nucleoDeA, preReleaseDeA, conPreReleaseA := partesDePrecedencia(a)
	nucleoDeB, preReleaseDeB, conPreReleaseB := partesDePrecedencia(b)

	if orden := slices.CompareFunc(nucleoDeA, nucleoDeB, compararNumeros); orden != 0 {
		return orden
	}

	switch {
	case conPreReleaseA && conPreReleaseB:
		return slices.CompareFunc(preReleaseDeA, preReleaseDeB, compararIdentificadores)
	case conPreReleaseA:
		return -1
	case conPreReleaseB:
		return 1
	default:
		return 0
	}
}

// partesDePrecedencia son, de una versión con la forma de FormaSemVer, los
// tres números de su núcleo y los identificadores de su pre-release, y si la
// tiene; sin la v inicial y sin los metadatos de construcción. El primer
// guion separa la pre-release, porque el núcleo solo lleva cifras y puntos; los
// de después son de sus identificadores.
func partesDePrecedencia(version string) (nucleo, preRelease []string, conPreRelease bool) {
	sinMetadatos, _, _ := strings.Cut(strings.TrimPrefix(version, "v"), "+")
	numeros, identificadores, conPreRelease := strings.Cut(sinMetadatos, "-")

	nucleo = strings.Split(numeros, ".")
	if conPreRelease {
		preRelease = strings.Split(identificadores, ".")
	}

	return nucleo, preRelease, conPreRelease
}

// compararNumeros compara dos números de una versión por su valor: la
// gramática no admite ceros a la izquierda (numeroSemVer), así que el más
// largo es el mayor y, con la misma longitud, lo dicen sus bytes. Es exacto y
// no desborda con números de cualquier longitud.
func compararNumeros(a, b string) int {
	if orden := cmp.Compare(len(a), len(b)); orden != 0 {
		return orden
	}

	return strings.Compare(a, b)
}

// compararIdentificadores compara dos identificadores de una pre-release: los
// numéricos por su valor, los alfanuméricos por sus bytes ASCII y un numérico
// antes que un alfanumérico.
func compararIdentificadores(a, b string) int {
	numericoA, numericoB := esNumerico(a), esNumerico(b)

	switch {
	case numericoA && numericoB:
		return compararNumeros(a, b)
	case numericoA:
		return -1
	case numericoB:
		return 1
	default:
		return strings.Compare(a, b)
	}
}

// esNumerico dice si un identificador de pre-release, que la gramática no
// deja vacío, es numérico: solo cifras ASCII.
func esNumerico(identificador string) bool {
	return strings.Trim(identificador, "0123456789") == ""
}
