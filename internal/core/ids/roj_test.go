package ids_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/ids"
)

// formaDelROJ es la forma de FR-005 de H23, escrita aparte de la del
// analizador, que comprueba byte a byte: el test contrasta las dos
// formulaciones. En RE2 `[A-Z]` y `[0-9]` solo casan ASCII, el espacio es el
// carácter U+0020 y ningún otro blanco, y `$` sin la bandera m es el final del
// texto.
var formaDelROJ = regexp.MustCompile(`^[A-Z]+( [A-Z]+)* [0-9]+/[0-9]{4}$`)

// rojDeLaPareja son, con la misma segunda formulación, los ROJ de los que
// FR-012 deduce un ECLI: los de siglas STS cuyo número cabe en un ECLI, que
// admite hasta 25 caracteres.
var rojDeLaPareja = regexp.MustCompile(`^STS [0-9]{1,25}/[0-9]{4}$`)

// semillasDelROJ son las formas que nombran FR-005, FR-006 y FR-012 de H23:
// el de la sentencia conocida, unas siglas de dos palabras, las de un auto,
// que no tienen equivalente, un número que no cabe en un ECLI, el que lleva el
// prefijo `ROJ:`, unas siglas en minúsculas, dos espacios seguidos, un año de
// dos cifras, uno sin año, un número con letras, un blanco al final y la
// entrada vacía. Cada una tiene su fichero en testdata/fuzz/FuzzROJ.
var semillasDelROJ = []string{
	"STS 3144/2023", "SAP M 1234/2020", "ATS 3144/2023", "STS " + strings.Repeat("9", 26) + "/2023",
	"ROJ: STS 3144/2023", "sts 3144/2023", "SAP  M 1234/2020", "STS 3144/23", "STS 3144",
	"STS 3144A/2023", "STS 3144/2023 ", "",
}

// TestAnalizarROJ fija la forma del ROJ: unas siglas de una o más palabras de
// letras mayúsculas separadas por un solo espacio, un espacio, el número, una
// barra y el año de cuatro cifras, sin nada delante ni detrás. Lo que casa se
// acepta tal cual, sin recortar ni pasar a mayúsculas y sin mirar las siglas en
// ninguna lista de órganos, y Numero() da el número con su año como van tras
// las siglas; lo demás es un error de argumentos que nombra la entrada y dice
// qué tiene de malo, con el valor cero al lado (H23, FR-005, FR-006, FR-025).
func TestAnalizarROJ(t *testing.T) {
	t.Parallel()

	t.Run("aceptados", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre, entrada, numero string
		}{
			{"la sentencia conocida", "STS 3144/2023", "3144/2023"},
			{"siglas de dos palabras", "SAP M 1234/2020", "1234/2020"},
			{"siglas de tres palabras", "STSJ CAT M 12/2021", "12/2021"},
			{"siglas de una letra", "A 1/2024", "1/2024"},
			{"siglas que no son de ningún órgano", "ZZZ 1/2024", "1/2024"},
			{"número con ceros por delante", "STS 0007/2023", "0007/2023"},
			{"número de 26 cifras", "STS " + strings.Repeat("9", 26) + "/2023", strings.Repeat("9", 26) + "/2023"},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				roj, err := ids.AnalizarROJ(caso.entrada)
				require.NoError(t, err)
				assert.Equal(t, caso.entrada, roj.String())
				assert.Equal(t, caso.numero, roj.Numero())
				assert.True(t, formaDelROJ.MatchString(caso.entrada), "el caso no tiene la forma de FR-005")
			})
		}
	})

	t.Run("rechazados", func(t *testing.T) {
		t.Parallel()

		const (
			forma       = "la forma <siglas> <número>/<año>"
			noEsLetra   = ", que no es una letra mayúscula de la A a la Z"
			malSeparada = " no son una o más palabras separadas por un solo espacio"
		)

		casos := []casoDeAnalisis{
			{nombre: "vacía", entrada: "", motivo: "está vacío y la forma es <siglas> <número>/<año>"},
			{nombre: "sin espacio", entrada: "STS3144/2023", motivo: "no lleva ningún espacio y " + forma + " separa con uno las siglas del número"},
			{nombre: "un tabulador por espacio", entrada: "STS\t3144/2023", motivo: "no lleva ningún espacio y " + forma + " separa con uno las siglas del número"},
			{nombre: "sin año", entrada: "STS 3144", motivo: `tras el último espacio va "3144" y ` + forma + " lleva ahí el número, una barra y el año"},
			{nombre: "un blanco detrás", entrada: "STS 3144/2023 ", motivo: `tras el último espacio va "" y ` + forma + " lleva ahí el número, una barra y el año"},
			{nombre: "con el prefijo ROJ:", entrada: "ROJ: STS 3144/2023", motivo: `las siglas "ROJ: STS" llevan ":"` + noEsLetra},
			{nombre: "siglas en minúsculas", entrada: "sts 3144/2023", motivo: `las siglas "sts" llevan "s"` + noEsLetra},
			{nombre: "siglas con una cifra", entrada: "ST5 3144/2023", motivo: `las siglas "ST5" llevan "5"` + noEsLetra},
			{nombre: "siglas con una letra que no es ASCII", entrada: "STÑ 3144/2023", motivo: `las siglas "STÑ" llevan "Ñ"` + noEsLetra},
			{nombre: "sin siglas", entrada: " 3144/2023", motivo: `las siglas ""` + malSeparada},
			{nombre: "un blanco delante", entrada: " STS 3144/2023", motivo: `las siglas " STS"` + malSeparada},
			{nombre: "dos espacios seguidos en las siglas", entrada: "SAP  M 1234/2020", motivo: `las siglas "SAP  M"` + malSeparada},
			{nombre: "dos espacios ante el número", entrada: "STS  3144/2023", motivo: `las siglas "STS "` + malSeparada},
			{nombre: "número vacío", entrada: "STS /2023", motivo: "el número está vacío y son una o más cifras"},
			{nombre: "número con letras", entrada: "STS 3144A/2023", motivo: `el número "3144A" lleva "A", que no es una cifra`},
			{nombre: "número con cifras que no son ASCII", entrada: "STS ٣١٤٤/2023", motivo: `el número "٣١٤٤" lleva "٣", que no es una cifra`},
			{nombre: "año de dos cifras", entrada: "STS 3144/23", motivo: `el año "23" tiene 2 cifras y son 4`},
			{nombre: "año de cinco cifras", entrada: "STS 3144/20233", motivo: `el año "20233" tiene 5 cifras y son 4`},
			{nombre: "año vacío", entrada: "STS 3144/", motivo: `el año "" tiene 0 cifras y son 4`},
			{nombre: "dos barras", entrada: "STS 3144/20/23", motivo: `el año "20/23" lleva "/", que no es una cifra`},
			{nombre: "salto de línea detrás", entrada: "STS 3144/2023\n", motivo: `el año "2023\n" lleva "\n", que no es una cifra`},
			{nombre: "byte que no es UTF-8", entrada: "STS 3144/2023\xff", motivo: `el año "2023\xff" lleva "\xff", que no es una cifra`},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				roj, err := ids.AnalizarROJ(caso.entrada)
				compruebaRechazo(t, err, caso.entrada, caso.motivo)
				assert.Equal(t, ids.ROJ{}, roj, "un ROJ rechazado no puede acompañarse de un valor")
				assert.False(t, formaDelROJ.MatchString(caso.entrada), "el caso tiene la forma de FR-005")
			})
		}
	})

	t.Run("valor cero", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, ids.ROJ{}.String(), "el valor cero no es ningún ROJ")
		assert.Empty(t, ids.ROJ{}.Numero(), "el valor cero no tiene número")
	})
}

// FuzzROJ contrasta el analizador del ROJ con la forma de FR-005 sobre
// cualquier entrada: ningún panic, lo aceptado casa con la forma y lo que casa
// se acepta, lo aceptado vuelve igual por String() y con su número, su ECLI se
// deduce solo en la pareja de FR-012, y todo rechazo es de argumentos, nombra
// la entrada y va con el valor cero (H23, FR-083).
func FuzzROJ(f *testing.F) {
	for _, semilla := range semillasDelROJ {
		f.Add(semilla)
	}

	f.Fuzz(func(t *testing.T, entrada string) {
		roj, err := ids.AnalizarROJ(entrada)
		if !formaDelROJ.MatchString(entrada) {
			compruebaErrorDeArgumentos(t, err, entrada)
			require.Equal(t, ids.ROJ{}, roj, "un ROJ rechazado no puede acompañarse de un valor")

			return
		}

		require.NoError(t, err, "se rechaza %q, que casa con la forma", entrada)
		require.Equal(t, entrada, roj.String(), "escribir lo analizado no da la entrada")
		require.Equal(t, entrada[strings.LastIndex(entrada, " ")+1:], roj.Numero())

		otraVez, err := ids.AnalizarROJ(roj.String())
		require.NoError(t, err, "se rechaza lo que el propio paquete escribe")
		require.Equal(t, roj, otraVez, "analizar lo ya escrito cambia el valor")

		ecli, seDeduce := roj.ECLI()
		require.Equal(t, rojDeLaPareja.MatchString(entrada), seDeduce,
			"el ECLI de %q se deduce fuera de la pareja de FR-012, o no se deduce dentro", entrada)
		compruebaLaPareja(t, ecli, roj, seDeduce)
	})
}
