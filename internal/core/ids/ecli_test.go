package ids_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/ids"
)

// formaDelECLI es la forma de FR-003 de H23, escrita aparte de la del
// analizador, que comprueba byte a byte: el test contrasta las dos
// formulaciones. En RE2 `[A-Z]` y `[0-9]` solo casan ASCII y `$` sin la
// bandera m es el final del texto, de modo que ni una letra de otra escritura
// ni un salto de línea final pasan.
var formaDelECLI = regexp.MustCompile(`^ECLI:ES:[A-Z][A-Z0-9]{0,6}:[0-9]{4}:[A-Z0-9.]{1,25}$`)

// semillasDelECLI son las formas que nombran FR-003 y los casos límite del
// spec de H23: el de la sentencia conocida, uno del Tribunal Constitucional,
// uno que termina en letra, el mal formado, el de minúsculas, el de otro país
// y uno con un blanco.
var semillasDelECLI = []string{
	"ECLI:ES:TS:2023:3144", "ECLI:ES:TC:2024:79", "ECLI:ES:TS:2023:3144A",
	"ECLI:ES:TS:2023", "ecli:es:ts:2023:3144", "ECLI:FR:CC:2023:1", " ECLI:ES:TS:2023:3144", "",
}

// TestAnalizarECLI fija la forma del ECLI de una resolución española: cinco
// partes separadas por dos puntos, sin blancos y en mayúsculas. Lo que casa se
// acepta tal cual, sin recortar ni pasar a mayúsculas; lo demás es un error de
// argumentos que nombra la entrada y dice qué tiene de malo, con el valor cero
// al lado, y el de un ECLI de otro país dice que no es español (H23, FR-003).
func TestAnalizarECLI(t *testing.T) {
	t.Parallel()

	t.Run("aceptados", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre, entrada, organo string
		}{
			{"la sentencia conocida", "ECLI:ES:TS:2023:3144", "TS"},
			{"del Tribunal Constitucional", "ECLI:ES:TC:2024:79", "TC"},
			{"número que termina en letra", "ECLI:ES:TS:2023:3144A", "TS"},
			{"número con puntos", "ECLI:ES:AN:2019:1.2.A", "AN"},
			{"órgano de un carácter", "ECLI:ES:T:2023:1", "T"},
			{"órgano de 7 caracteres, con cifras", "ECLI:ES:TSJM123:2020:1234", "TSJM123"},
			{"número de 25 caracteres", "ECLI:ES:TS:2023:" + strings.Repeat("9", 25), "TS"},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				ecli, err := ids.AnalizarECLI(caso.entrada)
				require.NoError(t, err)
				assert.Equal(t, caso.entrada, ecli.String())
				assert.Equal(t, caso.organo, ecli.Organo())
				assert.True(t, formaDelECLI.MatchString(caso.entrada), "el caso no tiene la forma de FR-003")
			})
		}
	})

	t.Run("rechazados", func(t *testing.T) {
		t.Parallel()

		const forma = "la forma ECLI:ES:<órgano>:<año>:<número>"

		casos := []casoDeAnalisis{
			{nombre: "vacía", entrada: "", motivo: "está vacío y la forma es ECLI:ES:<órgano>:<año>:<número>"},
			{nombre: "cuatro partes", entrada: "ECLI:ES:TS:2023", motivo: "tiene 4 partes separadas por dos puntos y " + forma + " tiene 5"},
			{nombre: "seis partes", entrada: "ECLI:ES:TS:2023:3144:1", motivo: "tiene 6 partes separadas por dos puntos y " + forma + " tiene 5"},
			{nombre: "sin dos puntos", entrada: "STS 3144/2023", motivo: "tiene 1 partes separadas por dos puntos y " + forma + " tiene 5"},
			{nombre: "en minúsculas", entrada: "ecli:es:ts:2023:3144", motivo: `empieza por "ecli" y ` + forma + " empieza por ECLI"},
			{nombre: "un blanco delante", entrada: " ECLI:ES:TS:2023:3144", motivo: `empieza por " ECLI" y ` + forma + " empieza por ECLI"},
			{nombre: "de otro país", entrada: "ECLI:FR:CC:2023:1", motivo: `el código de país es "FR" y no ES: no es español`},
			{nombre: "de la Unión Europea", entrada: "ECLI:EU:C:2019:123", motivo: `el código de país es "EU" y no ES: no es español`},
			{nombre: "país en minúsculas", entrada: "ECLI:es:TS:2023:3144", motivo: `el código de país es "es" y ` + forma + " lleva ES"},
			{nombre: "país de tres letras", entrada: "ECLI:ESP:TS:2023:3144", motivo: `el código de país es "ESP" y ` + forma + " lleva ES"},
			{nombre: "órgano vacío", entrada: "ECLI:ES::2023:3144", motivo: `el órgano "" tiene 0 caracteres y son de 1 a 7`},
			{nombre: "órgano de 8 caracteres", entrada: "ECLI:ES:TSJMADRI:2023:1", motivo: `el órgano "TSJMADRI" tiene 8 caracteres y son de 1 a 7`},
			{nombre: "órgano que empieza por cifra", entrada: "ECLI:ES:1TS:2023:1", motivo: `el órgano "1TS" empieza por una cifra y no por una letra`},
			{nombre: "órgano en minúsculas", entrada: "ECLI:ES:ts:2023:3144", motivo: `el órgano "ts" lleva "t", que no es una letra mayúscula de la A a la Z ni una cifra`},
			{nombre: "órgano con una letra que no es ASCII", entrada: "ECLI:ES:TÑ:2023:1", motivo: `el órgano "TÑ" lleva "Ñ", que no es una letra mayúscula de la A a la Z ni una cifra`},
			{nombre: "un blanco dentro", entrada: "ECLI:ES:T S:2023:3144", motivo: `el órgano "T S" lleva " ", que no es una letra mayúscula de la A a la Z ni una cifra`},
			{nombre: "año de tres cifras", entrada: "ECLI:ES:TS:202:3144", motivo: `el año "202" tiene 3 cifras y son 4`},
			{nombre: "año de cinco cifras", entrada: "ECLI:ES:TS:20233:3144", motivo: `el año "20233" tiene 5 cifras y son 4`},
			{nombre: "año con una letra", entrada: "ECLI:ES:TS:20A3:3144", motivo: `el año "20A3" lleva "A", que no es una cifra`},
			{nombre: "año con cifras que no son ASCII", entrada: "ECLI:ES:TS:٢٠٢٣:3144", motivo: `el año "٢٠٢٣" lleva "٢", que no es una cifra`},
			{nombre: "número vacío", entrada: "ECLI:ES:TS:2023:", motivo: `el número "" tiene 0 caracteres y son de 1 a 25`},
			{
				nombre: "número de 26 caracteres", entrada: "ECLI:ES:TS:2023:" + strings.Repeat("9", 26),
				motivo: `el número "` + strings.Repeat("9", 26) + `" tiene 26 caracteres y son de 1 a 25`,
			},
			{nombre: "número en minúsculas", entrada: "ECLI:ES:TS:2023:3144a", motivo: `el número "3144a" lleva "a", que no es una letra mayúscula de la A a la Z, una cifra ni un punto`},
			{nombre: "número con un guion", entrada: "ECLI:ES:TS:2023:31-44", motivo: `el número "31-44" lleva "-", que no es una letra mayúscula de la A a la Z, una cifra ni un punto`},
			{nombre: "un blanco detrás", entrada: "ECLI:ES:TS:2023:3144 ", motivo: `el número "3144 " lleva " ", que no es una letra mayúscula de la A a la Z, una cifra ni un punto`},
			{nombre: "salto de línea detrás", entrada: "ECLI:ES:TS:2023:3144\n", motivo: `el número "3144\n" lleva "\n", que no es una letra mayúscula de la A a la Z, una cifra ni un punto`},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				ecli, err := ids.AnalizarECLI(caso.entrada)
				compruebaRechazo(t, err, caso.entrada, caso.motivo)
				assert.Equal(t, ids.ECLI{}, ecli, "un ECLI rechazado no puede acompañarse de un valor")
				assert.False(t, formaDelECLI.MatchString(caso.entrada), "el caso tiene la forma de FR-003")
			})
		}
	})

	t.Run("valor cero", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, ids.ECLI{}.String(), "el valor cero no es ningún ECLI")
		assert.Empty(t, ids.ECLI{}.Organo())
	})
}

// FuzzECLI contrasta el analizador del ECLI con la forma de FR-003 sobre
// cualquier entrada: ningún panic, lo aceptado casa con la forma y lo que casa
// se acepta, lo aceptado vuelve igual por String() y con su órgano, y todo
// rechazo es de argumentos, nombra la entrada y va con el valor cero (H23,
// FR-113).
func FuzzECLI(f *testing.F) {
	for _, semilla := range semillasDelECLI {
		f.Add(semilla)
	}

	f.Fuzz(func(t *testing.T, entrada string) {
		ecli, err := ids.AnalizarECLI(entrada)
		if !formaDelECLI.MatchString(entrada) {
			compruebaErrorDeArgumentos(t, err, entrada)
			require.Equal(t, ids.ECLI{}, ecli, "un ECLI rechazado no puede acompañarse de un valor")

			return
		}

		require.NoError(t, err, "se rechaza %q, que casa con la forma", entrada)
		require.Equal(t, entrada, ecli.String(), "escribir lo analizado no da la entrada")
		require.Equal(t, strings.Split(entrada, ":")[2], ecli.Organo())

		otraVez, err := ids.AnalizarECLI(ecli.String())
		require.NoError(t, err, "se rechaza lo que el propio paquete escribe")
		require.Equal(t, ecli, otraVez, "analizar lo ya escrito cambia el valor")
	})
}
