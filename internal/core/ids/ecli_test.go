package ids_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/ids"
)

// formaDelECLI es la forma de FR-005 de H23, escrita aparte de la del
// analizador, que comprueba byte a byte: el test contrasta las dos
// formulaciones. En RE2 `[A-Z]` y `[0-9]` solo casan ASCII y `$` sin la
// bandera m es el final del texto, de modo que ni una letra de otra escritura
// ni un salto de línea final pasan.
var formaDelECLI = regexp.MustCompile(`^ECLI:ES:[A-Z][A-Z0-9]{0,6}:[0-9]{4}:[A-Z0-9.]{1,25}$`)

// eclisDeLaPareja son, con la misma segunda formulación, los ECLI de los que
// FR-012 deduce un ROJ: los de órgano TS con el número solo de cifras.
var eclisDeLaPareja = regexp.MustCompile(`^ECLI:ES:TS:[0-9]{4}:[0-9]{1,25}$`)

// semillasDelECLI son las formas que nombran FR-005, FR-006 y FR-012 de H23:
// el de la sentencia conocida, uno del Tribunal Constitucional, los tres sin
// equivalente —otro órgano, un número que termina en letra y otro con un
// punto—, el mal formado, el de minúsculas, el de otro país, uno con un blanco
// delante, otro con el blanco detrás y la entrada vacía. Cada una tiene su
// fichero en testdata/fuzz/FuzzECLI.
var semillasDelECLI = []string{
	"ECLI:ES:TS:2023:3144", "ECLI:ES:TC:2024:79", "ECLI:ES:AN:2023:3144", "ECLI:ES:TS:2023:3144A",
	"ECLI:ES:TS:2023:31.44", "ECLI:ES:TS:2023", "ecli:es:ts:2023:3144", "ECLI:FR:CC:2023:1",
	" ECLI:ES:TS:2023:3144", "ECLI:ES:TS:2023:3144 ", "",
}

// TestAnalizarECLI fija la forma del ECLI de una resolución española: cinco
// partes separadas por dos puntos, sin blancos y en mayúsculas. Lo que casa se
// acepta tal cual, sin recortar ni pasar a mayúsculas y sin buscar el órgano en
// ninguna lista; lo demás es un error de argumentos que nombra la entrada y
// dice qué tiene de malo, con el valor cero al lado, y el de un ECLI de otro
// país dice que no es español (H23, FR-005, FR-006, FR-014).
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
			{"órgano que no es de ninguna lista", "ECLI:ES:ZZZ:2024:1", "ZZZ"},
			{"número de 25 caracteres", "ECLI:ES:TS:2023:" + strings.Repeat("9", 25), "TS"},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				ecli, err := ids.AnalizarECLI(caso.entrada)
				require.NoError(t, err)
				assert.Equal(t, caso.entrada, ecli.String())
				assert.Equal(t, caso.organo, ecli.Organo())
				assert.True(t, formaDelECLI.MatchString(caso.entrada), "el caso no tiene la forma de FR-005")
			})
		}
	})

	t.Run("rechazados", func(t *testing.T) {
		t.Parallel()

		const (
			forma      = "la forma ECLI:ES:<órgano>:<año>:<número>"
			noEsOrgano = ", que no es una letra mayúscula de la A a la Z ni una cifra"
			noEsNumero = ", que no es una letra mayúscula de la A a la Z, una cifra ni un punto"
		)

		casos := []casoDeAnalisis{
			{nombre: "vacía", entrada: "", motivo: "está vacío y la forma es ECLI:ES:<órgano>:<año>:<número>"},
			{nombre: "una parte de menos", entrada: "ECLI:ES:TS:2023", motivo: "tiene 4 partes separadas por dos puntos y " + forma + " tiene 5"},
			{nombre: "una parte de más", entrada: "ECLI:ES:TS:2023:3144:1", motivo: "tiene 6 partes separadas por dos puntos y " + forma + " tiene 5"},
			{nombre: "sin dos puntos", entrada: "STS 3144/2023", motivo: "tiene 1 partes separadas por dos puntos y " + forma + " tiene 5"},
			{nombre: "en minúsculas", entrada: "ecli:es:ts:2023:3144", motivo: `empieza por "ecli" y ` + forma + " empieza por ECLI"},
			{nombre: "un blanco delante", entrada: " ECLI:ES:TS:2023:3144", motivo: `empieza por " ECLI" y ` + forma + " empieza por ECLI"},
			{nombre: "de otro país", entrada: "ECLI:FR:CC:2023:1", motivo: `el código de país es "FR" y no ES: no es español`},
			{nombre: "de la Unión Europea", entrada: "ECLI:EU:C:2019:123", motivo: `el código de país es "EU" y no ES: no es español`},
			{nombre: "país en minúsculas", entrada: "ECLI:es:TS:2023:3144", motivo: `el código de país es "es" y ` + forma + " lleva ES"},
			{nombre: "país de tres letras", entrada: "ECLI:ESP:TS:2023:3144", motivo: `el código de país es "ESP" y ` + forma + " lleva ES"},
			{nombre: "país vacío", entrada: "ECLI::TS:2023:3144", motivo: `el código de país es "" y ` + forma + " lleva ES"},
			{nombre: "órgano vacío", entrada: "ECLI:ES::2023:3144", motivo: `el órgano "" tiene 0 caracteres y son de 1 a 7`},
			{nombre: "órgano de 8 caracteres", entrada: "ECLI:ES:TSJMADRI:2023:1", motivo: `el órgano "TSJMADRI" tiene 8 caracteres y son de 1 a 7`},
			{nombre: "órgano que empieza por cifra", entrada: "ECLI:ES:1TS:2023:1", motivo: `el órgano "1TS" empieza por una cifra y no por una letra`},
			{nombre: "órgano en minúsculas", entrada: "ECLI:ES:ts:2023:3144", motivo: `el órgano "ts" lleva "t"` + noEsOrgano},
			{nombre: "órgano con una letra que no es ASCII", entrada: "ECLI:ES:TÑ:2023:1", motivo: `el órgano "TÑ" lleva "Ñ"` + noEsOrgano},
			{nombre: "un blanco dentro", entrada: "ECLI:ES:T S:2023:3144", motivo: `el órgano "T S" lleva " "` + noEsOrgano},
			{nombre: "año de tres cifras", entrada: "ECLI:ES:TS:202:3144", motivo: `el año "202" tiene 3 cifras y son 4`},
			{nombre: "año de cinco cifras", entrada: "ECLI:ES:TS:20233:3144", motivo: `el año "20233" tiene 5 cifras y son 4`},
			{nombre: "año con una letra", entrada: "ECLI:ES:TS:20A3:3144", motivo: `el año "20A3" lleva "A", que no es una cifra`},
			{nombre: "año con cifras que no son ASCII", entrada: "ECLI:ES:TS:٢٠٢٣:3144", motivo: `el año "٢٠٢٣" lleva "٢", que no es una cifra`},
			{nombre: "número vacío", entrada: "ECLI:ES:TS:2023:", motivo: `el número "" tiene 0 caracteres y son de 1 a 25`},
			{
				nombre: "número de 26 caracteres", entrada: "ECLI:ES:TS:2023:" + strings.Repeat("9", 26),
				motivo: `el número "` + strings.Repeat("9", 26) + `" tiene 26 caracteres y son de 1 a 25`,
			},
			{nombre: "número en minúsculas", entrada: "ECLI:ES:TS:2023:3144a", motivo: `el número "3144a" lleva "a"` + noEsNumero},
			{nombre: "número con un guion", entrada: "ECLI:ES:TS:2023:31-44", motivo: `el número "31-44" lleva "-"` + noEsNumero},
			{nombre: "un blanco detrás", entrada: "ECLI:ES:TS:2023:3144 ", motivo: `el número "3144 " lleva " "` + noEsNumero},
			{nombre: "salto de línea detrás", entrada: "ECLI:ES:TS:2023:3144\n", motivo: `el número "3144\n" lleva "\n"` + noEsNumero},
			{nombre: "byte que no es UTF-8", entrada: "ECLI:ES:TS:2023:3144\xff", motivo: `el número "3144\xff" lleva "\xff"` + noEsNumero},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				ecli, err := ids.AnalizarECLI(caso.entrada)
				compruebaRechazo(t, err, caso.entrada, caso.motivo)
				assert.Equal(t, ids.ECLI{}, ecli, "un ECLI rechazado no puede acompañarse de un valor")
				assert.False(t, formaDelECLI.MatchString(caso.entrada), "el caso tiene la forma de FR-005")
			})
		}
	})

	t.Run("valor cero", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, ids.ECLI{}.String(), "el valor cero no es ningún ECLI")
		assert.Empty(t, ids.ECLI{}.Organo())
	})
}

// TestEquivalencia fija la única pareja en la que un ECLI y un ROJ se deducen
// el uno del otro sin consultar nada: el ECLI de órgano TS con el número solo
// de cifras y el ROJ de siglas STS, con el número y el año trasladados carácter
// a carácter. En cualquier otro caso —la tabla de FR-012— ninguno de los dos
// métodos deduce nada, y lo que acompaña al «no» es el valor cero, nunca un
// identificador compuesto a medias (H23, FR-012, FR-023).
func TestEquivalencia(t *testing.T) {
	t.Parallel()

	t.Run("la pareja, en los dos sentidos", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre, ecli, roj string
		}{
			{"la sentencia conocida", "ECLI:ES:TS:2023:3144", "STS 3144/2023"},
			{"la de docs/JURISPRUDENCIA.md", "ECLI:ES:TS:2026:3505", "STS 3505/2026"},
			{"número de una cifra", "ECLI:ES:TS:1999:1", "STS 1/1999"},
			{"ceros por delante, sin normalizar", "ECLI:ES:TS:2023:0007", "STS 0007/2023"},
			{
				"número de 25 cifras",
				"ECLI:ES:TS:2023:" + strings.Repeat("9", 25), "STS " + strings.Repeat("9", 25) + "/2023",
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				ecli, err := ids.AnalizarECLI(caso.ecli)
				require.NoError(t, err)

				roj, err := ids.AnalizarROJ(caso.roj)
				require.NoError(t, err)

				delECLI, seDeduce := ecli.ROJ()
				require.True(t, seDeduce, "de %q no se deduce su ROJ", caso.ecli)
				assert.Equal(t, roj, delECLI)
				assert.Equal(t, caso.roj, delECLI.String())

				delROJ, seDeduce := roj.ECLI()
				require.True(t, seDeduce, "de %q no se deduce su ECLI", caso.roj)
				assert.Equal(t, ecli, delROJ)
				assert.Equal(t, caso.ecli, delROJ.String())
			})
		}
	})

	t.Run("un ECLI sin equivalente", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre, entrada string
		}{
			{"de otro órgano", "ECLI:ES:AN:2023:3144"},
			{"del Tribunal Constitucional", "ECLI:ES:TC:2024:79"},
			{"de un órgano que empieza por TS", "ECLI:ES:TSJM:2023:3144"},
			{"de un órgano de una letra", "ECLI:ES:T:2023:3144"},
			{"número final con letras, el de un auto", "ECLI:ES:TS:2023:3144A"},
			{"número final con puntos", "ECLI:ES:TS:2023:31.44"},
			{"número final solo de letras", "ECLI:ES:TS:2023:A"},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				ecli, err := ids.AnalizarECLI(caso.entrada)
				require.NoError(t, err)

				roj, seDeduce := ecli.ROJ()
				assert.False(t, seDeduce, "de %q se deduce %q", caso.entrada, roj.String())
				assert.Equal(t, ids.ROJ{}, roj, "lo que no se deduce no puede acompañarse de un ROJ")
				assert.False(t, eclisDeLaPareja.MatchString(caso.entrada), "el caso es de la pareja de FR-012")
			})
		}
	})

	t.Run("un ROJ sin equivalente", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre, entrada string
		}{
			{"de otras siglas", "SAN 3144/2023"},
			{"de un auto del mismo órgano", "ATS 3144/2023"},
			{"de siglas de dos palabras", "SAP M 1234/2020"},
			{"de siglas que empiezan por STS", "STSJ M 3144/2023"},
			{"de siglas que terminan en STS", "M STS 3144/2023"},
			{"número que no cabe en un ECLI", "STS " + strings.Repeat("9", 26) + "/2023"},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				roj, err := ids.AnalizarROJ(caso.entrada)
				require.NoError(t, err)

				ecli, seDeduce := roj.ECLI()
				assert.False(t, seDeduce, "de %q se deduce %q", caso.entrada, ecli.String())
				assert.Equal(t, ids.ECLI{}, ecli, "lo que no se deduce no puede acompañarse de un ECLI")
				assert.False(t, rojDeLaPareja.MatchString(caso.entrada), "el caso es de la pareja de FR-012")
			})
		}
	})

	t.Run("valor cero", func(t *testing.T) {
		t.Parallel()

		roj, seDeduce := ids.ECLI{}.ROJ()
		assert.False(t, seDeduce, "del valor cero no se deduce nada")
		assert.Equal(t, ids.ROJ{}, roj)

		ecli, seDeduce := ids.ROJ{}.ECLI()
		assert.False(t, seDeduce, "del valor cero no se deduce nada")
		assert.Equal(t, ids.ECLI{}, ecli)
	})
}

// FuzzECLI contrasta el analizador del ECLI con la forma de FR-005 sobre
// cualquier entrada: ningún panic, lo aceptado casa con la forma y lo que casa
// se acepta, lo aceptado vuelve igual por String() y con su órgano, su ROJ se
// deduce solo en la pareja de FR-012, y todo rechazo es de argumentos, nombra
// la entrada y va con el valor cero (H23, FR-083).
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

		roj, seDeduce := ecli.ROJ()
		require.Equal(t, eclisDeLaPareja.MatchString(entrada), seDeduce,
			"el ROJ de %q se deduce fuera de la pareja de FR-012, o no se deduce dentro", entrada)
		compruebaLaPareja(t, ecli, roj, seDeduce)
	})
}

// compruebaLaPareja exige de un ECLI y de un ROJ, uno analizado y el otro
// deducido de él, lo que FR-012 promete de la equivalencia. Si no se deduce,
// uno de los dos es el valor cero: lo que no se deduce no se acompaña de un
// identificador. Si se deduce, los dos tienen su forma —lo que escriben vuelve
// a analizarse al mismo valor— y cada uno da el otro. La comparten los dos
// objetivos de fuzz, cada uno desde su lado.
func compruebaLaPareja(t *testing.T, ecli ids.ECLI, roj ids.ROJ, seDeduce bool) {
	t.Helper()

	if !seDeduce {
		require.True(t, ecli == ids.ECLI{} || roj == ids.ROJ{},
			"sin equivalencia entre %q y %q, uno de los dos tiene que ser el valor cero", ecli.String(), roj.String())

		return
	}

	elECLI, err := ids.AnalizarECLI(ecli.String())
	require.NoError(t, err, "el ECLI de la pareja no tiene su forma")
	require.Equal(t, ecli, elECLI)

	elROJ, err := ids.AnalizarROJ(roj.String())
	require.NoError(t, err, "el ROJ de la pareja no tiene su forma")
	require.Equal(t, roj, elROJ)

	delECLI, seDeduce := ecli.ROJ()
	require.True(t, seDeduce, "de %q no se deduce su ROJ", ecli.String())
	require.Equal(t, roj, delECLI, "el ROJ de %q no es el de la pareja", ecli.String())

	delROJ, seDeduce := roj.ECLI()
	require.True(t, seDeduce, "de %q no se deduce su ECLI", roj.String())
	require.Equal(t, ecli, delROJ, "el ECLI de %q no es el de la pareja", roj.String())
}
