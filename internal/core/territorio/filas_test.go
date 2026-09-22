package territorio

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

// relacionEnFilas escribe la relación como el repositorio escribe la
// congelada (research.md D4): la cabecera y una fila por línea, en orden de
// código, en forma de flujo y con cada texto entre comillas dobles.
func relacionEnFilas(relacion FicheroDeMunicipios) []byte {
	var fichero strings.Builder

	fmt.Fprintf(&fichero, "fecha: %s\nsource: %s\nmunicipios:\n",
		entrecomillar(relacion.Fecha), entrecomillar(relacion.Source))

	for _, codigo := range slices.Sorted(maps.Keys(relacion.Municipios)) {
		fila := relacion.Municipios[codigo]
		fmt.Fprintf(&fichero, "  %s: {dc: %s, nombre: %s, provincia: %s, comunidad: %s}\n", entrecomillar(codigo),
			entrecomillar(fila.DC), entrecomillar(fila.Nombre), entrecomillar(fila.Provincia),
			entrecomillar(fila.Comunidad))
	}

	return []byte(fichero.String())
}

// correspondenciaEnFilas escribe la correspondencia como el repositorio
// escribe la congelada (research.md D4): la cabecera y una fila por línea, en
// orden de código.
func correspondenciaEnFilas(correspondencia FicheroDeDIR3) []byte {
	var fichero strings.Builder

	fmt.Fprintf(&fichero, "fecha: %s\nsource: %s\ncorrespondencia:\n",
		entrecomillar(correspondencia.Fecha), entrecomillar(correspondencia.Source))

	for _, codigo := range slices.Sorted(maps.Keys(correspondencia.Correspondencia)) {
		fmt.Fprintf(&fichero, "  %s: %s\n", entrecomillar(codigo),
			entrecomillar(correspondencia.Correspondencia[codigo]))
	}

	return []byte(fichero.String())
}

// entrecomillar escribe un texto entre comillas dobles de YAML, con la barra
// invertida y la comilla doble escapadas y todo lo demás tal cual.
func entrecomillar(texto string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(texto) + `"`
}

// fuentesEnFilas codifica cada fichero como fuentesDe, salvo la relación y la
// correspondencia, que escribe como las congeladas.
func fuentesEnFilas(t *testing.T, ficheros Ficheros) Fuentes {
	t.Helper()

	fuentes := fuentesDe(t, ficheros)
	fuentes.Municipios = relacionEnFilas(ficheros.Municipios)
	fuentes.DIR3 = correspondenciaEnFilas(ficheros.DIR3)

	return fuentes
}

// lecturaDeFilas son las formas de leer uno de los dos ficheros de filas, F.
type lecturaDeFilas[F any] struct {
	// carga es la de Cargar.
	carga func([]byte) (F, error)
	// conElLector es la de la carga anterior: el fichero entero con el lector
	// de YAML, esté o no en su forma.
	conElLector func([]byte) (F, error)
	// enSuForma dice si la carga lee el fichero en su forma, sin el lector.
	enSuForma func([]byte) bool
}

// lecturaDeLaRelacion es la de data/territorio/municipios.yaml.
func lecturaDeLaRelacion() lecturaDeFilas[FicheroDeMunicipios] {
	return lecturaDeFilas[FicheroDeMunicipios]{
		carga: decodificarRelacion,
		conElLector: func(contenido []byte) (FicheroDeMunicipios, error) {
			relacion, filas, err := decodificarConElLector[relacionSinFilas, *relacionSinFilas, FilaDeMunicipio](
				rutaDeMunicipios, "municipios", contenido)

			return FicheroDeMunicipios{Fecha: relacion.Fecha, Source: relacion.Source, Municipios: filas}, err
		},
		enSuForma: func(contenido []byte) bool {
			_, _, enSuForma := leerEnSuForma[relacionSinFilas](
				rutaDeMunicipios, "municipios", contenido, leerFilaDeMunicipio)

			return enSuForma
		},
	}
}

// lecturaDeLaCorrespondencia es la de data/territorio/dir3.yaml.
func lecturaDeLaCorrespondencia() lecturaDeFilas[FicheroDeDIR3] {
	return lecturaDeFilas[FicheroDeDIR3]{
		carga: decodificarCorrespondencia,
		conElLector: func(contenido []byte) (FicheroDeDIR3, error) {
			dir3, filas, err := decodificarConElLector[correspondenciaSinFilas, *correspondenciaSinFilas, string](
				rutaDeDIR3, "correspondencia", contenido)

			return FicheroDeDIR3{Fecha: dir3.Fecha, Source: dir3.Source, Correspondencia: filas}, err
		},
		enSuForma: func(contenido []byte) bool {
			_, _, enSuForma := leerEnSuForma[correspondenciaSinFilas](
				rutaDeDIR3, "correspondencia", contenido, leerFilaDeDIR3)

			return enSuForma
		},
	}
}

// compruebaComoElLector exige que la carga decodifique un fichero de filas
// exactamente como la carga anterior, que lo leía entero con el lector de
// YAML: el mismo fichero y, si lo hay, el mismo defecto, con el mismo texto
// —su fichero y su línea o su clave—. Y lo compara además con lo que el
// lector decodifica por su cuenta, sin nada de la carga: lo que la carga lee
// sin defectos, el lector lo lee igual.
func compruebaComoElLector[F any](t *testing.T, lectura lecturaDeFilas[F], contenido []byte) {
	t.Helper()

	fichero, err := lectura.carga(contenido)
	delLector, errDelLector := lectura.conElLector(contenido)

	if errDelLector != nil {
		require.EqualError(t, err, errDelLector.Error(), "la carga no dice el defecto de la carga anterior")
	} else {
		require.NoError(t, err, "la carga ve un defecto que la carga anterior no veía")
	}

	assert.Equal(t, delLector, fichero, "la carga no da el fichero que daba la carga anterior")

	if err == nil {
		var porSuCuenta F

		require.NoError(t, yaml.Unmarshal(contenido, &porSuCuenta), "el lector no lee lo que lee la carga")
		assert.Equal(t, porSuCuenta, fichero, "la carga no da el fichero que da el lector")
	}
}

// Las cabeceras en su forma de la relación y de la correspondencia.
const (
	cabeceraDeLaRelacion        = "fecha: \"2026-02-04\"\nsource: prueba.relacion\nmunicipios:\n"
	cabeceraDeLaCorrespondencia = "fecha: \"2026-09-21\"\nsource: prueba.correspondencia\ncorrespondencia:\n"
)

// filaDeMunicipio escribe la línea de un municipio de la relación con su clave
// y su nombre tal como vienen, sin escapar nada, y los demás campos fijos.
func filaDeMunicipio(clave, nombre string) string {
	return `  "` + clave + `": {dc: "5", nombre: "` + nombre + `", provincia: "28", comunidad: "01"}` + "\n"
}

// casoDeForma es un fichero de filas escrito a mano y si está en su forma.
type casoDeForma struct {
	nombre    string
	contenido string
	enSuForma bool
}

// TestFormaDeLasFilas fija qué ficheros de filas lee la carga en su forma, sin
// el lector de YAML, y que con todos, lo lea como lo lea, da exactamente lo
// que da el lector: los que están en su forma, porque así está hecha la forma;
// los demás, porque los lee él (FR-043; research.md D4, S3).
func TestFormaDeLasFilas(t *testing.T) {
	t.Parallel()

	t.Run("relacion", func(t *testing.T) {
		t.Parallel()

		compruebaFormas(t, lecturaDeLaRelacion(), casosDeLaRelacion())
	})

	t.Run("correspondencia", func(t *testing.T) {
		t.Parallel()

		compruebaFormas(t, lecturaDeLaCorrespondencia(), casosDeLaCorrespondencia())
	})
}

// compruebaFormas comprueba cada caso de forma de uno de los dos ficheros.
func compruebaFormas[F any](t *testing.T, lectura lecturaDeFilas[F], casos []casoDeForma) {
	t.Helper()

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			contenido := []byte(caso.contenido)
			assert.Equal(t, caso.enSuForma, lectura.enSuForma(contenido), "si el fichero está en su forma")
			compruebaComoElLector(t, lectura, contenido)
		})
	}
}

// casosDeLaRelacion son los casos de forma de la relación.
func casosDeLaRelacion() []casoDeForma {
	villaprueba := filaDeMunicipio("28991", "Villaprueba")
	otra := filaDeMunicipio("28992", "Rozas de Prueba, Las")
	enBloque := "fecha: \"2026-02-04\"\nsource: prueba.relacion\nmunicipios:\n" +
		"  \"28991\":\n    dc: \"5\"\n    nombre: Villaprueba\n    provincia: \"28\"\n    comunidad: \"01\"\n"

	return []casoDeForma{
		// En su forma.
		{nombre: "una-fila", contenido: cabeceraDeLaRelacion + villaprueba, enSuForma: true},
		{nombre: "varias-filas", contenido: cabeceraDeLaRelacion + villaprueba + otra, enSuForma: true},
		{
			nombre:    "sin-salto-de-linea-al-final",
			contenido: cabeceraDeLaRelacion + villaprueba + strings.TrimSuffix(otra, "\n"),
			enSuForma: true,
		},
		{
			nombre:    "escapes",
			contenido: cabeceraDeLaRelacion + filaDeMunicipio(`28\"991\\`, `Villa \"La Prueba\" \\ C:\\Barra\\`),
			enSuForma: true,
		},
		{
			nombre: "runas-de-todos-los-tramos",
			contenido: cabeceraDeLaRelacion + filaDeMunicipio("28991",
				"Peñíscola\xc2\xa0del Río \xef\xac\x81 \xef\xbf\xbd \xf0\x9f\x8f\xb0 ~"),
			enSuForma: true,
		},
		{
			nombre:    "tabulador-y-marca-de-orden",
			contenido: cabeceraDeLaRelacion + filaDeMunicipio("\t28991\xef\xbb\xbf", "\tVilla\t\xef\xbb\xbfPrueba\t"),
			enSuForma: true,
		},
		{
			nombre:    "clave-repetida",
			contenido: cabeceraDeLaRelacion + villaprueba + otra + villaprueba,
			enSuForma: true,
		},
		{
			nombre:    "clave-de-1022-bytes",
			contenido: cabeceraDeLaRelacion + filaDeMunicipio(strings.Repeat("9", 1022), "V"),
			enSuForma: true,
		},
		{
			nombre:    "clave-de-511-enes",
			contenido: cabeceraDeLaRelacion + filaDeMunicipio(strings.Repeat("ñ", 511), "V"),
			enSuForma: true,
		},
		{
			nombre: "cabecera-con-retorno-de-carro-y-salto-de-linea",
			contenido: "fecha: \"2026-02-04\"\r\nsource: prueba.relacion\r\nmunicipios:\n" +
				villaprueba + otra + villaprueba,
			enSuForma: true,
		},
		{
			nombre: "cabecera-con-comentarios-y-lineas-en-blanco",
			contenido: "# relación\n\nfecha: \"2026-02-04\" # fecha\nsource: prueba.relacion\n\nmunicipios:\n" +
				villaprueba + otra + villaprueba,
			enSuForma: true,
		},
		// Fuera de su forma.
		{nombre: "en-bloque", contenido: enBloque},
		{
			nombre:    "filas-en-bloque-con-clave-repetida",
			contenido: enBloque + strings.TrimPrefix(enBloque, "fecha: \"2026-02-04\"\nsource: prueba.relacion\nmunicipios:\n"),
		},
		{nombre: "sangria-de-cuatro", contenido: cabeceraDeLaRelacion + "  " + villaprueba},
		{
			nombre: "campos-en-otro-orden",
			contenido: cabeceraDeLaRelacion +
				`  "28991": {nombre: "Villaprueba", dc: "5", provincia: "28", comunidad: "01"}` + "\n",
		},
		{
			nombre:    "comentario-detras-de-una-fila",
			contenido: cabeceraDeLaRelacion + strings.TrimSuffix(villaprueba, "\n") + " # comentario\n",
		},
		{nombre: "algo-detras-de-una-fila", contenido: cabeceraDeLaRelacion + strings.TrimSuffix(villaprueba, "\n") + "x\n"},
		{
			nombre: "clave-sin-comillas",
			contenido: cabeceraDeLaRelacion +
				`  28991: {dc: "5", nombre: "Villaprueba", provincia: "28", comunidad: "01"}` + "\n",
		},
		{
			nombre: "textos-sin-comillas",
			contenido: cabeceraDeLaRelacion +
				`  "28991": {dc: 5, nombre: Villaprueba, provincia: "28", comunidad: "01"}` + "\n",
		},
		{
			nombre: "comillas-simples",
			contenido: cabeceraDeLaRelacion +
				`  "28991": {dc: '5', nombre: 'Villa''prueba', provincia: "28", comunidad: "01"}` + "\n",
		},
		{nombre: "escape-de-tabulador", contenido: cabeceraDeLaRelacion + filaDeMunicipio("28991", `Villa\tPrueba`)},
		{nombre: "escape-hexadecimal", contenido: cabeceraDeLaRelacion + filaDeMunicipio("28991", `Villa\x41`)},
		{nombre: "escape-unicode", contenido: cabeceraDeLaRelacion + filaDeMunicipio("28991", `Villa\`+"u00e9")},
		{nombre: "escape-desconocido", contenido: cabeceraDeLaRelacion + filaDeMunicipio("28991", `Villa\q`)},
		{
			nombre:    "texto-sin-cerrar",
			contenido: cabeceraDeLaRelacion + `  "28991": {dc: "5` + "\n",
		},
		{nombre: "caracter-de-control", contenido: cabeceraDeLaRelacion + filaDeMunicipio("28991", "Villa\x01")},
		{nombre: "suprimir", contenido: cabeceraDeLaRelacion + filaDeMunicipio("28991", "Villa\x7f")},
		{nombre: "nel", contenido: cabeceraDeLaRelacion + filaDeMunicipio("28991", "Villa\xc2\x85Prueba")},
		{nombre: "separador-de-linea", contenido: cabeceraDeLaRelacion + filaDeMunicipio("28991", "Villa\xe2\x80\xa8")},
		{nombre: "separador-de-parrafo", contenido: cabeceraDeLaRelacion + filaDeMunicipio("28991", "Villa\xe2\x80\xa9")},
		{nombre: "no-es-utf8", contenido: cabeceraDeLaRelacion + filaDeMunicipio("28991", "Villa\xff")},
		{
			nombre:    "retorno-de-carro-en-las-filas",
			contenido: cabeceraDeLaRelacion + strings.ReplaceAll(villaprueba+otra, "\n", "\r\n"),
		},
		{nombre: "linea-en-blanco-entre-filas", contenido: cabeceraDeLaRelacion + villaprueba + "\n" + otra},
		{nombre: "linea-en-blanco-al-final", contenido: cabeceraDeLaRelacion + villaprueba + "\n"},
		{
			nombre:    "clave-de-1023-bytes",
			contenido: cabeceraDeLaRelacion + filaDeMunicipio(strings.Repeat("9", 1023), "V"),
		},
		{
			nombre:    "dos-documentos",
			contenido: "fecha: \"2026-02-04\"\nsource: prueba.relacion\n---\nmunicipios:\n" + villaprueba,
		},
		{
			nombre:    "retorno-de-carro-solo-en-la-cabecera",
			contenido: "fecha: \"2026-02-04\"\rsource: prueba.relacion\nmunicipios:\n" + villaprueba + villaprueba,
		},
		{nombre: "cabecera-ilegible", contenido: "fecha: [\nmunicipios:\n" + villaprueba},
		{nombre: "sin-filas", contenido: cabeceraDeLaRelacion},
		{nombre: "sin-la-clave-de-las-filas", contenido: "fecha: \"2026-02-04\"\nsource: prueba.relacion\n"},
		{nombre: "clave-de-las-filas-con-valor", contenido: "fecha: \"2026-02-04\"\nsource: x\nmunicipios: {}\n"},
		{
			nombre:    "clave-de-las-filas-repetida",
			contenido: "municipios:\nfecha: \"2026-02-04\"\nsource: prueba.relacion\nmunicipios:\n" + villaprueba,
		},
		{nombre: "otra-clave-detras-de-las-filas", contenido: cabeceraDeLaRelacion + villaprueba + "otra: 1\n"},
	}
}

// casosDeLaCorrespondencia son los casos de forma de la correspondencia.
func casosDeLaCorrespondencia() []casoDeForma {
	fila := `  "28991": "L01289915"` + "\n"

	return []casoDeForma{
		{nombre: "una-fila", contenido: cabeceraDeLaCorrespondencia + fila, enSuForma: true},
		{
			nombre:    "escapes",
			contenido: cabeceraDeLaCorrespondencia + `  "28991": "L01\"28991\\5"` + "\n",
			enSuForma: true,
		},
		{nombre: "clave-repetida", contenido: cabeceraDeLaCorrespondencia + fila + fila, enSuForma: true},
		{nombre: "en-bloque", contenido: cabeceraDeLaCorrespondencia + `  "28991": L01289915` + "\n"},
		{
			nombre:    "algo-detras-de-una-fila",
			contenido: cabeceraDeLaCorrespondencia + `  "28991": "L01289915" # comentario` + "\n",
		},
		{
			nombre:    "barra-invertida-al-final-de-la-linea",
			contenido: cabeceraDeLaCorrespondencia + `  "28991": "L01\` + "\n",
		},
		{nombre: "fila-de-la-relacion", contenido: cabeceraDeLaCorrespondencia + filaDeMunicipio("28991", "V")},
	}
}

// nombresDificiles son nombres que la forma tiene que leer como el lector de
// YAML: con comillas dobles y simples, tildes y otros diacríticos, barras y
// barras invertidas, apóstrofos, el punto volado, espacios alrededor y por
// dentro, los indicadores de YAML, textos que sin comillas no serían textos
// y caracteres de todos los tramos de Unicode que el lector admite.
var nombresDificiles = []string{
	`Villa "La Prueba"`,
	`"`,
	`""`,
	`\`,
	`\"`,
	`C:\Barra\Invertida\`,
	"L'Alfàs de Prueba",
	"Castell de l'Assaig",
	"Vall d'Uixó, La",
	"Peñíscola del Río",
	"Güeñes de Prueba",
	"Cel·la de Prueba",
	"Nombre/Izena",
	"Prueba/Proba, La",
	"  espacios  alrededor  ",
	"",
	"# no es un comentario",
	"a: b, c",
	"{llaves} [corchetes]",
	"&ancla *alias !etiqueta %directiva @arroba `grave` |barra >mayor",
	"- --- ... ? :",
	"yes",
	"null",
	"~",
	"123",
	"0x1F",
	"1e3",
	".inf",
	"Ñandú\xc2\xa0de Prueba",
	"\xef\xac\x81 \xef\xbf\xbd \xf0\x9f\x8f\xb0",
	"\tVilla\tde Prueba\t",
	"\xef\xbb\xbfVilla\xef\xbb\xbf",
}

// TestNombresDificiles fija que una relación y una correspondencia escritas
// como las congeladas, con cualquiera de los nombres difíciles como clave y
// como texto, se leen en su forma, dan exactamente lo que da el lector de YAML
// y devuelven cada nombre tal como se escribió (FR-043; research.md D4).
func TestNombresDificiles(t *testing.T) {
	t.Parallel()

	for _, nombre := range nombresDificiles {
		t.Run(fmt.Sprintf("%q", nombre), func(t *testing.T) {
			t.Parallel()

			relacion, dir3 := unaFila(nombre, nombre)

			assert.True(t, lecturaDeLaRelacion().enSuForma(relacion), "la relación no está en su forma")
			compruebaComoElLector(t, lecturaDeLaRelacion(), relacion)

			leida, err := decodificarRelacion(relacion)
			require.NoError(t, err)
			assert.Equal(t, nombre, leida.Municipios[nombre].Nombre)

			assert.True(t, lecturaDeLaCorrespondencia().enSuForma(dir3), "la correspondencia no está en su forma")
			compruebaComoElLector(t, lecturaDeLaCorrespondencia(), dir3)

			leido, err := decodificarCorrespondencia(dir3)
			require.NoError(t, err)
			assert.Equal(t, nombre, leido.Correspondencia[nombre])
		})
	}
}

// FuzzFilasComoElLector escribe una relación y una correspondencia de una sola
// fila, con una clave y un texto cualesquiera, como el repositorio escribe las
// congeladas, y exige que la carga las decodifique exactamente como el lector
// de YAML, estén o no en su forma (compruebaComoElLector). Las semillas son
// los nombres difíciles y los caracteres que sacan una fila de su forma.
func FuzzFilasComoElLector(f *testing.F) {
	fueraDeSuForma := []string{
		"\n", "\r", "\r\n", "\x00", "\x01", "\x7f", "\xc2\x85", "\xe2\x80\xa8", "\xe2\x80\xa9", "\xff", "\xed\xa0\x80",
		"a\n  \"28992\": \"L01289920\"", strings.Repeat("9", 1023),
	}

	for _, nombre := range slices.Concat(nombresDificiles, fueraDeSuForma) {
		f.Add("28991", nombre)
		f.Add(nombre, "Villaprueba")
	}

	f.Fuzz(func(t *testing.T, clave, texto string) {
		relacion, dir3 := unaFila(clave, texto)

		compruebaComoElLector(t, lecturaDeLaRelacion(), relacion)
		compruebaComoElLector(t, lecturaDeLaCorrespondencia(), dir3)
	})
}

// unaFila escribe como las congeladas una relación de un solo municipio, con
// esa clave y ese texto como nombre, y una correspondencia de una sola fila,
// con esa clave y ese texto como DIR3.
func unaFila(clave, texto string) (relacion, dir3 []byte) {
	relacion = relacionEnFilas(FicheroDeMunicipios{
		Fecha:  fechaDeLaRelacion,
		Source: fuenteDeLaRelacion,
		Municipios: map[string]FilaDeMunicipio{
			clave: {DC: "5", Nombre: texto, Provincia: "28", Comunidad: "01"},
		},
	})
	dir3 = correspondenciaEnFilas(FicheroDeDIR3{
		Fecha:           fechaDeLaCorrespondencia,
		Source:          fuenteDeLaCorrespondencia,
		Correspondencia: map[string]string{clave: texto},
	})

	return relacion, dir3
}

// TestCargarEnFilas repite TestCargar con la relación y la correspondencia
// escritas como las congeladas, que la carga lee en su forma: el territorio
// completo da los mismos ficheros que escritos en bloque, y cada caso, los
// mismos defectos, con su fichero y su línea o su clave (FR-044; research.md
// D4, S3).
func TestCargarEnFilas(t *testing.T) {
	t.Parallel()

	t.Run("completo", func(t *testing.T) {
		t.Parallel()

		sinteticos := ficherosSinteticos()
		fuentes := fuentesEnFilas(t, sinteticos)

		assert.True(t, lecturaDeLaRelacion().enSuForma(fuentes.Municipios), "la relación no está en su forma")
		assert.True(t, lecturaDeLaCorrespondencia().enSuForma(fuentes.DIR3), "la correspondencia no está en su forma")

		enFilas, err := decodificar(fuentes)
		require.NoError(t, err)
		assert.Equal(t, sinteticos, enFilas)

		enBloque, err := decodificar(fuentesDe(t, sinteticos))
		require.NoError(t, err)
		assert.Equal(t, enBloque, enFilas)

		cargar(t, sinteticos)
	})

	for _, caso := range casosDeCarga() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			compruebaDefectosDeCarga(t, caso, fuentesEnFilas)
		})
	}
}
