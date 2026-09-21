package ids_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/ids"
)

// semillasDelContrato son las del contrato de identificadores §6. Las reciben
// los dos objetivos de fuzz enteras, no repartidas: lo que tiene forma de
// código INE ejercita el rechazo del analizador de DIR3, y al revés.
var semillasDelContrato = []string{
	"28074", "280748", "01001", "52001", "", "2807", "2807a", "00074", "28000",
	"L01280748", "l01280748", "L0128074", "X01280748",
}

// cincoCifrasASCII y seisCifrasASCII son la forma de la gramática de
// referencia del contrato §2. Van como expresión regular a propósito: el
// analizador comprueba byte a byte, y el test lo contrasta con otra
// formulación de la misma gramática. En RE2 `[0-9]` solo casa cifras ASCII y
// `$` sin la bandera m es el final del texto, de modo que un salto de línea
// final no pasa.
var (
	cincoCifrasASCII = regexp.MustCompile(`^[0-9]{5}$`)
	seisCifrasASCII  = regexp.MustCompile(`^[0-9]{6}$`)
)

// cumpleLaGramaticaINE es la gramática de referencia del código INE: cinco
// cifras ASCII, provincia de 1 a 52 y municipio de 1 a 999, comprobados como
// números y no como texto.
func cumpleLaGramaticaINE(texto string) bool {
	if !cincoCifrasASCII.MatchString(texto) {
		return false
	}

	provincia, errDeProvincia := strconv.Atoi(texto[:2])
	municipio, errDeMunicipio := strconv.Atoi(texto[2:])

	return errDeProvincia == nil && errDeMunicipio == nil &&
		1 <= provincia && provincia <= 52 && 1 <= municipio && municipio <= 999
}

// cumpleLaGramaticaINEConDigito es la del código con su dígito: seis cifras
// ASCII, y las cinco primeras, un código INE.
func cumpleLaGramaticaINEConDigito(texto string) bool {
	return seisCifrasASCII.MatchString(texto) && cumpleLaGramaticaINE(texto[:5])
}

// casoDeAnalisis es una entrada con lo que su análisis tiene que dar: nada
// que objetar, o un error de argumentos cuyo mensaje contiene motivo.
type casoDeAnalisis struct {
	nombre  string
	entrada string
	motivo  string
}

// TestAnalizarCodigoINE fija las dos gramáticas del código INE del contrato
// §2: cinco cifras, y seis con el dígito de control declarado. Lo que casa se
// acepta tal cual, con sus ceros por delante; lo demás es un error de
// argumentos que nombra la entrada y dice qué tiene de malo, y el valor que lo
// acompaña es el cero, nunca un código plausible (FR-030, FR-033).
func TestAnalizarCodigoINE(t *testing.T) {
	t.Parallel()

	t.Run("cinco cifras", func(t *testing.T) {
		t.Parallel()

		casos := []casoDeAnalisis{
			{nombre: "Leganés", entrada: "28074"},
			{nombre: "ceros por delante", entrada: "01001"},
			{nombre: "la última provincia", entrada: "52001"},
			{nombre: "el último municipio", entrada: "28999"},
			{nombre: "vacía", entrada: "", motivo: "tiene 0 cifras y la forma PPMMM tiene 5"},
			{nombre: "cuatro cifras", entrada: "2807", motivo: "tiene 4 cifras y la forma PPMMM tiene 5"},
			{nombre: "seis cifras", entrada: "280748", motivo: "tiene 6 cifras y la forma PPMMM tiene 5"},
			{nombre: "letra", entrada: "2807a", motivo: `"a" no es una cifra`},
			{nombre: "espacio delante", entrada: " 28074", motivo: `" " no es una cifra`},
			{nombre: "salto de línea detrás", entrada: "28074\n", motivo: `"\n" no es una cifra`},
			{nombre: "signo", entrada: "+2807", motivo: `"+" no es una cifra`},
			{nombre: "cifras que no son ASCII", entrada: "٢٨٠٧٤", motivo: `"٢" no es una cifra`},
			{nombre: "byte que no es UTF-8", entrada: "2807\xff", motivo: `"\xff" no es una cifra`},
			{nombre: "provincia 00", entrada: "00074", motivo: `la provincia "00" no está entre 01 y 52`},
			{nombre: "provincia 53", entrada: "53001", motivo: `la provincia "53" no está entre 01 y 52`},
			{nombre: "provincia 99", entrada: "99001", motivo: `la provincia "99" no está entre 01 y 52`},
			{nombre: "municipio 000", entrada: "28000", motivo: `el municipio "000" no está entre 001 y 999`},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				codigo, err := ids.AnalizarCodigoINE(caso.entrada)
				if caso.motivo != "" {
					compruebaRechazo(t, err, caso.entrada, caso.motivo)
					assert.Equal(t, ids.CodigoINE{}, codigo, "un código rechazado no puede acompañarse de un valor")

					return
				}

				require.NoError(t, err)
				assert.Equal(t, caso.entrada, codigo.String())
				assert.Equal(t, caso.entrada[:2], codigo.Provincia())
			})
		}
	})

	t.Run("seis cifras", func(t *testing.T) {
		t.Parallel()

		casos := []casoDeAnalisis{
			{nombre: "Leganés", entrada: "280748"},
			{nombre: "ceros por delante", entrada: "010014"},
			{nombre: "dígito cero", entrada: "010040"},
			{nombre: "vacía", entrada: "", motivo: "tiene 0 cifras y la forma PPMMMD tiene 6"},
			{nombre: "cinco cifras", entrada: "28074", motivo: "tiene 5 cifras y la forma PPMMMD tiene 6"},
			{nombre: "siete cifras", entrada: "2807480", motivo: "tiene 7 cifras y la forma PPMMMD tiene 6"},
			{nombre: "letra en el dígito", entrada: "28074a", motivo: `"a" no es una cifra`},
			{nombre: "provincia 00", entrada: "000748", motivo: `la provincia "00" no está entre 01 y 52`},
			{nombre: "municipio 000", entrada: "280008", motivo: `el municipio "000" no está entre 001 y 999`},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				codigo, digito, err := ids.AnalizarCodigoINEConDigito(caso.entrada)
				if caso.motivo != "" {
					compruebaRechazo(t, err, caso.entrada, caso.motivo)
					assert.Equal(t, ids.CodigoINE{}, codigo, "un código rechazado no puede acompañarse de un valor")
					assert.Zero(t, digito, "un código rechazado no puede acompañarse de un dígito")

					return
				}

				require.NoError(t, err)
				assert.Equal(t, caso.entrada[:5], codigo.String())
				assert.Equal(t, caso.entrada[:2], codigo.Provincia())
				assert.Equal(t, caso.entrada[5], digito)
			})
		}
	})
}

// TestComprobarDigito prueba los cien pares de dígitos: el código se acepta
// solo cuando el declarado es el oficial, y si no, el error es de argumentos,
// nombra el código con el dígito recibido y dice cuál es cada uno de los dos
// (contrato de identificadores §3).
func TestComprobarDigito(t *testing.T) {
	t.Parallel()

	leganes, err := ids.AnalizarCodigoINE("28074")
	require.NoError(t, err)

	for declarado := byte('0'); declarado <= '9'; declarado++ {
		for oficial := byte('0'); oficial <= '9'; oficial++ {
			err := leganes.ComprobarDigito(declarado, oficial)
			if declarado == oficial {
				require.NoError(t, err, "el dígito %c es el oficial", declarado)

				continue
			}

			compruebaRechazo(t, err, "28074"+string(declarado), fmt.Sprintf(
				"el dígito de control recibido es %q y el oficial es %q", string(declarado), string(oficial)))
		}
	}
}

// TestIdaYVuelta es la propiedad de FR-032 sobre los dos identificadores:
// analizar y volver a escribir da la misma cadena normalizada, y analizar lo
// ya normalizado da el mismo valor. La parte del DIR3 vive con su analizador,
// en dir3_test.go.
func TestIdaYVuelta(t *testing.T) {
	t.Parallel()

	t.Run("ine", func(t *testing.T) {
		t.Parallel()

		for _, entrada := range []string{"28074", "01001", "52001", "28999", "09001"} {
			compruebaIdaYVueltaINE(t, entrada)
		}

		for _, entrada := range []string{"280748", "010014", "520017", "010040"} {
			compruebaIdaYVueltaINEConDigito(t, entrada)
		}
	})

	t.Run("dir3", func(t *testing.T) {
		t.Parallel()

		for _, entrada := range dir3Aceptados {
			compruebaIdaYVueltaDIR3(t, entrada)
		}
	})
}

// FuzzCodigoINE contrasta los dos analizadores del código INE con la
// gramática de referencia sobre cualquier entrada: ningún panic, lo aceptado
// casa con la gramática y lo rechazado no, lo aceptado tiene ida y vuelta
// estable, y todo rechazo es de argumentos y nombra la entrada (FR-035,
// contrato de identificadores §6).
func FuzzCodigoINE(f *testing.F) {
	for _, semilla := range semillasDelContrato {
		f.Add(semilla)
	}

	f.Fuzz(func(t *testing.T, entrada string) {
		_, err := ids.AnalizarCodigoINE(entrada)
		if cumpleLaGramaticaINE(entrada) {
			require.NoError(t, err, "se rechaza %q, que casa con la gramática", entrada)
			compruebaIdaYVueltaINE(t, entrada)
		} else {
			compruebaErrorDeArgumentos(t, err, entrada)
		}

		_, _, err = ids.AnalizarCodigoINEConDigito(entrada)
		if cumpleLaGramaticaINEConDigito(entrada) {
			require.NoError(t, err, "se rechaza %q, que casa con la gramática", entrada)
			compruebaIdaYVueltaINEConDigito(t, entrada)
		} else {
			compruebaErrorDeArgumentos(t, err, entrada)
		}
	})
}

// compruebaIdaYVueltaINE analiza un código de cinco cifras que tiene que
// aceptarse, y exige que escribirlo dé la misma cadena y que volver a
// analizarla dé el mismo valor.
func compruebaIdaYVueltaINE(t *testing.T, entrada string) {
	t.Helper()

	codigo, err := ids.AnalizarCodigoINE(entrada)
	require.NoError(t, err, "se rechaza %q", entrada)
	require.Equal(t, entrada, codigo.String(), "escribir lo analizado cambia la cadena")
	require.Equal(t, entrada[:2], codigo.Provincia())

	otraVez, err := ids.AnalizarCodigoINE(codigo.String())
	require.NoError(t, err, "se rechaza lo que el propio paquete escribe")
	require.Equal(t, codigo, otraVez, "analizar lo ya normalizado cambia el valor")
	require.Equal(t, codigo.String(), otraVez.String())
}

// compruebaIdaYVueltaINEConDigito hace lo mismo con un código de seis cifras:
// el código y el dígito que salen, escritos uno tras otro, son la entrada, y
// el dígito declarado se comprueba como el oficial.
func compruebaIdaYVueltaINEConDigito(t *testing.T, entrada string) {
	t.Helper()

	codigo, digito, err := ids.AnalizarCodigoINEConDigito(entrada)
	require.NoError(t, err, "se rechaza %q", entrada)
	require.Equal(t, entrada, codigo.String()+string(digito), "escribir lo analizado cambia la cadena")
	require.NoError(t, codigo.ComprobarDigito(digito, entrada[5]))

	compruebaIdaYVueltaINE(t, codigo.String())
}
