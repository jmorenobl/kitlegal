package ids_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/ids"
)

// formaDelDIR3ASCII es la forma de la gramática de referencia del DIR3 de
// ayuntamiento del contrato §2, que el test contrasta con el análisis byte a
// byte del paquete.
var formaDelDIR3ASCII = regexp.MustCompile(`^[Ll]01[0-9]{6}$`)

// cumpleLaGramaticaDIR3 es la gramática de referencia del DIR3: la forma, y
// que las cinco cifras que siguen a L01 sean un código INE.
func cumpleLaGramaticaDIR3(texto string) bool {
	return formaDelDIR3ASCII.MatchString(texto) && cumpleLaGramaticaINE(texto[3:8])
}

// dir3Aceptados son DIR3 de ayuntamiento bien formados, con la letra en las
// dos cajas.
var dir3Aceptados = []string{"L01280748", "l01280748", "L01010014", "L01520017", "L01289990"}

// TestAnalizarDIR3 fija la gramática del DIR3 de ayuntamiento del contrato
// §2: la letra se acepta en las dos cajas y se escribe en mayúscula, el resto
// son cifras, y lo que sigue a L01 es un código INE y su dígito. Todo lo demás
// es un error de argumentos que nombra la entrada y dice qué tiene de malo, con
// el valor cero al lado (FR-031, FR-033, FR-034).
func TestAnalizarDIR3(t *testing.T) {
	t.Parallel()

	t.Run("aceptados", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre, entrada, normalizado, codigoINE string
			digito                                  byte
		}{
			{"Leganés", "L01280745", "L01280745", "28074", '5'},
			{"letra en minúscula", "l01280745", "L01280745", "28074", '5'},
			{"ceros por delante", "L01010014", "L01010014", "01001", '4'},
			{"dígito cero", "L01289990", "L01289990", "28999", '0'},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				dir3, err := ids.AnalizarDIR3(caso.entrada)
				require.NoError(t, err)
				assert.Equal(t, caso.normalizado, dir3.String())
				assert.Equal(t, caso.codigoINE, dir3.CodigoINE().String())
				assert.Equal(t, caso.digito, dir3.Digito())
			})
		}
	})

	t.Run("rechazados", func(t *testing.T) {
		t.Parallel()

		const forma = "la forma L01PPMMMD"

		casos := []casoDeAnalisis{
			{nombre: "vacía", entrada: "", motivo: "está vacío y la forma es L01PPMMMD"},
			{nombre: "otra letra", entrada: "X01280748", motivo: `empieza por "X" y ` + forma + " empieza por L"},
			{nombre: "letra del Estado", entrada: "E01280748", motivo: `empieza por "E" y ` + forma + " empieza por L"},
			{nombre: "letra que no es ASCII", entrada: "Ł01280748", motivo: `empieza por "Ł" y ` + forma + " empieza por L"},
			{nombre: "espacio delante", entrada: " L01280748", motivo: `empieza por " " y ` + forma + " empieza por L"},
			{nombre: "sin el dígito", entrada: "L0128074", motivo: "tras la L tiene 7 cifras y " + forma + " tiene 8"},
			{nombre: "una cifra de más", entrada: "L012807480", motivo: "tras la L tiene 9 cifras y " + forma + " tiene 8"},
			{nombre: "letra en el dígito", entrada: "L0128074a", motivo: `"a" no es una cifra`},
			{nombre: "salto de línea detrás", entrada: "L01280748\n", motivo: `"\n" no es una cifra`},
			{nombre: "otra entidad local", entrada: "L02280748", motivo: `tras la L va "02" y ` + forma + " lleva 01"},
			{nombre: "provincia 00", entrada: "L01000748", motivo: `la provincia "00" no está entre 01 y 52`},
			{nombre: "provincia 53", entrada: "L01530018", motivo: `la provincia "53" no está entre 01 y 52`},
			{nombre: "municipio 000", entrada: "L01280008", motivo: `el municipio "000" no está entre 001 y 999`},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				dir3, err := ids.AnalizarDIR3(caso.entrada)
				compruebaRechazo(t, err, caso.entrada, caso.motivo)
				assert.Equal(t, ids.DIR3{}, dir3, "un DIR3 rechazado no puede acompañarse de un valor")
			})
		}
	})
}

// TestDIR3DeAyuntamiento es la relación INE ↔ DIR3 del contrato §4: componer
// un DIR3 con un código y un dígito los devuelve intactos y escribe L01 seguido
// de los dos, y descomponer un DIR3 analizado y volver a componerlo da el mismo
// valor (FR-031).
func TestDIR3DeAyuntamiento(t *testing.T) {
	t.Parallel()

	t.Run("componer y descomponer", func(t *testing.T) {
		t.Parallel()

		for _, entrada := range []string{"28074", "01001", "52001"} {
			codigo, err := ids.AnalizarCodigoINE(entrada)
			require.NoError(t, err)

			for digito := byte('0'); digito <= '9'; digito++ {
				dir3 := ids.DIR3DeAyuntamiento(codigo, digito)

				assert.Equal(t, codigo, dir3.CodigoINE())
				assert.Equal(t, digito, dir3.Digito())
				assert.Equal(t, "L01"+entrada+string(digito), dir3.String())

				analizado, err := ids.AnalizarDIR3(dir3.String())
				require.NoError(t, err, "se rechaza el DIR3 compuesto %q", dir3.String())
				assert.Equal(t, dir3, analizado)
			}
		}
	})

	t.Run("descomponer y componer", func(t *testing.T) {
		t.Parallel()

		for _, entrada := range dir3Aceptados {
			compruebaRelacionConElINE(t, entrada)
		}
	})

	t.Run("valor cero", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, ids.CodigoINE{}.String(), "el valor cero no es ningún código INE")
		assert.Empty(t, ids.CodigoINE{}.Provincia())
		assert.Empty(t, ids.DIR3{}.String(), "el valor cero no es ningún DIR3")
		assert.Equal(t, ids.CodigoINE{}, ids.DIR3{}.CodigoINE())
		assert.Zero(t, ids.DIR3{}.Digito())
	})
}

// FuzzCodigoDIR3 contrasta el analizador del DIR3 con la gramática de
// referencia sobre cualquier entrada: ningún panic, lo aceptado casa con la
// gramática y lo rechazado no, lo aceptado tiene ida y vuelta estable y
// relación con su código INE, y todo rechazo es de argumentos y nombra la
// entrada (FR-035, contrato de identificadores §6).
func FuzzCodigoDIR3(f *testing.F) {
	for _, semilla := range semillasDelContrato {
		f.Add(semilla)
	}

	f.Fuzz(func(t *testing.T, entrada string) {
		_, err := ids.AnalizarDIR3(entrada)
		if !cumpleLaGramaticaDIR3(entrada) {
			compruebaErrorDeArgumentos(t, err, entrada)

			return
		}

		require.NoError(t, err, "se rechaza %q, que casa con la gramática", entrada)
		compruebaIdaYVueltaDIR3(t, entrada)
		compruebaRelacionConElINE(t, entrada)
	})
}

// compruebaIdaYVueltaDIR3 analiza un DIR3 que tiene que aceptarse, y exige que
// escribirlo dé la entrada con la letra en mayúscula, que volver a analizarla
// dé el mismo valor y que escribirlo otra vez no cambie nada.
func compruebaIdaYVueltaDIR3(t *testing.T, entrada string) {
	t.Helper()

	dir3, err := ids.AnalizarDIR3(entrada)
	require.NoError(t, err, "se rechaza %q", entrada)
	require.Equal(t, strings.ToUpper(entrada[:1])+entrada[1:], dir3.String(),
		"escribir lo analizado no da la entrada con la letra en mayúscula")

	otraVez, err := ids.AnalizarDIR3(dir3.String())
	require.NoError(t, err, "se rechaza lo que el propio paquete escribe")
	require.Equal(t, dir3, otraVez, "analizar lo ya normalizado cambia el valor")
	require.Equal(t, dir3.String(), otraVez.String())
}

// compruebaRelacionConElINE analiza un DIR3 que tiene que aceptarse y exige
// que su código INE y su dígito sean las cifras que siguen a L01, y que
// componer un DIR3 con ellos dé el mismo valor y la misma cadena.
func compruebaRelacionConElINE(t *testing.T, entrada string) {
	t.Helper()

	dir3, err := ids.AnalizarDIR3(entrada)
	require.NoError(t, err, "se rechaza %q", entrada)
	require.Equal(t, entrada[3:8], dir3.CodigoINE().String())
	require.Equal(t, entrada[8], dir3.Digito())

	compuesto := ids.DIR3DeAyuntamiento(dir3.CodigoINE(), dir3.Digito())
	require.Equal(t, dir3, compuesto)
	require.Equal(t, dir3.String(), compuesto.String())
}
