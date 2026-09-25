package territorio

import (
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// alfabetoDelPliegue es a lo que el pliegue lleva toda letra que conoce:
// minúsculas y cifras ASCII y el espacio que separa palabras. Lo que queda
// fuera es una runa que la tabla no cubre, y el control del corpus congelado
// la detecta precisamente por eso (research.md D10, D27).
const alfabetoDelPliegue = "abcdefghijklmnopqrstuvwxyz0123456789 "

// Dos caracteres que no se ven escritos tal cual, así que van por sus bytes
// en UTF-8.
const (
	// acentoAgudoCombinante es U+0301: el acento de una «é» escrita
	// descompuesta, como la dejan algunos teclados y algunos copiados.
	acentoAgudoCombinante = "\xcc\x81"
	// espacioDuro es U+00A0, el espacio de no separación.
	espacioDuro = "\xc2\xa0"
)

// TestPlegar fija el pliegue de nombres de data-model §2.7: minúsculas, cada
// letra con diacrítico a su letra base, los signos separadores y los espacios
// colapsados en un espacio, y nada delante ni detrás. Lo que la tabla no
// cubre no se inventa: queda en minúscula, fuera del alfabeto del pliegue
// (FR-015, research.md D10).
func TestPlegar(t *testing.T) {
	t.Parallel()

	t.Run("mayúsculas", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "villaprueba", Plegar("VILLAPRUEBA"))
		assert.Equal(t, "villaprueba", Plegar("VillaPrueba"))
		assert.Equal(t, "28991", Plegar("28991"), "las cifras no cambian")
	})

	t.Run("diacríticos", func(t *testing.T) {
		t.Parallel()

		// Cada letra de la tabla del pliegue, en minúscula y en mayúscula, y
		// la letra base a la que tiene que llegar (research.md D10).
		tabla := map[string]string{
			"áàäâ": "a", "éèëê": "e", "íìïî": "i", "óòöô": "o", "úùüû": "u", "ñ": "n", "ç": "c",
		}

		for letras, base := range tabla {
			for _, letra := range letras {
				minuscula, mayuscula := string(letra), string(unicode.ToUpper(letra))
				assert.Equal(t, base, Plegar(minuscula), "%q no pliega a su letra base", minuscula)
				assert.Equal(t, base, Plegar(mayuscula), "%q no pliega a su letra base", mayuscula)
			}
		}

		assert.Equal(t, "peniscola del rio", Plegar("Peñíscola del Río"))
		assert.Equal(t, "iruneta", Plegar("IRUÑETA"))
		assert.Equal(t, "collegats", Plegar("Col·legats"), "el punto volado une la ele de cada lado, no la separa")
		assert.Equal(t, "leganes", Plegar("Legane"+acentoAgudoCombinante+"s"),
			"una marca combinante es un diacrítico más")
	})

	t.Run("espacios y separadores", func(t *testing.T) {
		t.Parallel()

		casos := map[string]string{
			"  Rozas   de  Prueba  ":         "rozas de prueba",
			"Rozas\tde\nPrueba":              "rozas de prueba",
			"Villa" + espacioDuro + "Prueba": "villa prueba",
			"Rozas de Prueba, Las":           "rozas de prueba las",
			"Alegría-Dulantzi":               "alegria dulantzi",
			"Iruñeta/Pamploneta":             "iruneta pamploneta",
			"Hospitalet, L'":                 "hospitalet l",
			"L'Hospitalet":                   "l hospitalet",
			"L’Hospitalet":                   "l hospitalet",
			"«Villa» (Prueba).":              "villa prueba",
			"Villa - / - Prueba":             "villa prueba",
			"":                               "",
			" ,.-/ ":                         "",
			"Castell d'Aro, Platja d'":       "castell d aro platja d",
		}

		for entrada, plegada := range casos {
			assert.Equal(t, plegada, Plegar(entrada), "el pliegue de %q", entrada)
		}
	})

	t.Run("lo que la tabla no cubre", func(t *testing.T) {
		t.Parallel()

		casos := map[string]string{
			"Ø":        "ø",
			"Åland":    "åland",
			"Straße":   "straße",
			"2807\xff": "2807" + string(utf8.RuneError),
			"٢٨":       "٢٨",
		}

		for entrada, plegada := range casos {
			assert.Equal(t, plegada, Plegar(entrada), "el pliegue de %q", entrada)
			assert.True(t, strings.ContainsFunc(Plegar(entrada), fueraDelAlfabeto),
				"%q no sale del alfabeto del pliegue: el control del corpus no la vería", entrada)
		}
	})

	t.Run("lo que la tabla cubre no sale del alfabeto", func(t *testing.T) {
		t.Parallel()

		plegado := Plegar("ÁÀÄÂ ÉÈËÊ ÍÌÏÎ ÓÒÖÔ ÚÙÜÛ Ñ Ç áàäâ éèëê íìïî óòöô úùüû ñ ç l·l ,.-/'’()«»")
		assert.False(t, strings.ContainsFunc(plegado, fueraDelAlfabeto),
			"%q tiene algo fuera del alfabeto del pliegue", plegado)
	})
}

// fueraDelAlfabeto dice si una runa no es del alfabeto del pliegue.
func fueraDelAlfabeto(runa rune) bool {
	return !strings.ContainsRune(alfabetoDelPliegue, runa)
}

// TestFormasDelNombre fija las formas conocidas de un municipio, derivadas
// solo de su nombre oficial y sin ningún caso especial (FR-015, FR-024,
// data-model §2.7): la del INE, la del artículo pospuesto antepuesto y cada
// lado de un nombre bilingüe con la suya, todas plegadas y sin repetir.
func TestFormasDelNombre(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		oficial string
		formas  []string
	}{
		{"sin artículo ni lados", "Villaprueba", []string{"villaprueba"}},
		{"artículo pospuesto", "Coruña, A", []string{"coruna a", "a coruna"}},
		{"artículo con apóstrofo", "Hospitalet de Llobregat, L'", []string{
			"hospitalet de llobregat l", "l hospitalet de llobregat",
		}},
		{"coma que no pospone un artículo", "Saus, Camallera i Llampaies", []string{"saus camallera i llampaies"}},
		{"nombre bilingüe", "Donostia/San Sebastián", []string{
			"donostia san sebastian", "donostia", "san sebastian",
		}},
		{"las dos cosas", "Vila Joiosa, la/Villajoyosa", []string{
			"vila joiosa la villajoyosa", "vila joiosa la", "la vila joiosa", "villajoyosa",
		}},
		{"lados que pliegan igual", "Pamploneta/PAMPLONETA", []string{"pamploneta pamploneta", "pamploneta"}},
		{"lado vacío", "Pamploneta/", []string{"pamploneta"}},
		{"artículo sin nombre", ", La", []string{"la"}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.formas, formasDelNombre(caso.oficial))
		})
	}
}

// TestPlegarEsIdempotente exige que plegar lo plegado no lo cambie, de modo
// que la entrada de quien pregunta y las formas del índice se comparan en el
// mismo espacio por muchas veces que se pliegue una u otra. Se comprueba runa
// a runa en todo el plano básico de Unicode, en las formas del territorio
// sintético y en cadenas que mezclan letras, separadores y bytes que no son
// UTF-8.
func TestPlegarEsIdempotente(t *testing.T) {
	t.Parallel()

	t.Run("cada runa del plano básico", func(t *testing.T) {
		t.Parallel()

		for runa := rune(0); runa <= 0xFFFF; runa++ {
			if !utf8.ValidRune(runa) {
				continue
			}

			entrada := "a" + string(runa) + "b"
			plegada := Plegar(entrada)
			if plegada != Plegar(plegada) {
				assert.Failf(t, "el pliegue no es idempotente", "U+%04X: %q pliega a %q y después a %q",
					runa, entrada, plegada, Plegar(plegada))
			}
		}
	})

	t.Run("nombres y formas", func(t *testing.T) {
		t.Parallel()

		entradas := []string{
			"", " ", "Villaprueba", "  Rozas   de  Prueba, Las ", "Iruñeta/Pamploneta", "Col·legats",
			"Legane" + acentoAgudoCombinante + "s", "L’Hospitalet", "2807\xff", "Straße ẞ İ K Å",
			"- , Villa -- Prueba ,-", "Villa" + espacioDuro + "Prueba",
		}
		for _, fila := range ficherosSinteticos().Municipios.Municipios {
			entradas = append(entradas, fila.Nombre)
			entradas = append(entradas, formasDelNombre(fila.Nombre)...)
		}

		for _, entrada := range entradas {
			plegada := Plegar(entrada)
			assert.Equal(t, plegada, Plegar(plegada), "plegar %q dos veces no da lo mismo que una", entrada)
		}
	})
}
