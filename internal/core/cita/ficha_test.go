package cita_test

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/cita"
)

// carpetaDelCorpus es la del corpus versionado de FuzzLeerFicha.
const carpetaDelCorpus = "testdata/fuzz/FuzzLeerFicha"

// corpusDeLaFicha es ese corpus, embebido al compilar el test. El dominio no
// hace entrada ni salida tampoco en sus tests —la lista core de depguard le
// deniega os e io—, así que el test no abre el fragmento de la evidencia del
// ADR 0036, que vive fuera del paquete: lo recibe del compilador en la semilla
// «fragmento», escrita con una orden desde ese fichero y nunca tecleada, y
// comprueba con su huella que es él (fragmento).
//
//go:embed testdata/fuzz/FuzzLeerFicha
var corpusDeLaFicha embed.FS

// Lo que el manifiesto de la evidencia del ADR 0036 y research.md M3 dicen del
// fragmento: su huella SHA-256, y que su ficha son sus once primeras líneas,
// 316 bytes.
const (
	huellaDelFragmento = "4886e0c8ca9836527ec08d8732b50315640a76803033374af5287b2a8321ee27"
	lineasDeLaFicha    = 11
	bytesDeLaFicha     = 316
)

// fichaSerializada son los ocho datos de la ficha del fragmento, con las
// claves de contracts/applet-cita.md §4 en su orden y la fecha AAAA-MM-DD.
const fichaSerializada = `{"roj":"STS 3144/2023","ecli":"ECLI:ES:TS:2023:3144",` +
	`"organo":"Tribunal Supremo. Sala de lo Civil","fecha":"2023-07-04","recurso":"4703/2019",` +
	`"resolucion":"1088/2023","ponente":"PEDRO JOSE VELA TORRES","tipo":"Sentencia"}`

// lineaDeOtroOrgano es la primera línea de una ficha sintética de otro
// órgano, cuyo ROJ y cuyo ECLI no son de la pareja de FR-012.
const lineaDeOtroOrgano = "Roj: SAP M 1234/2020 - ECLI:ES:APM:2020:1234"

// Las formas de los cuatro datos que se comparan (FR-005, FR-021), escritas
// aparte de las del paquete, que comprueba byte a byte: el fuzz contrasta las
// dos formulaciones. En RE2 `[A-Z]` y `[0-9]` solo casan ASCII y `$` sin la
// bandera m es el final del texto.
var (
	patronDelROJ    = regexp.MustCompile(`^[A-Z]+( [A-Z]+)* [0-9]+/[0-9]{4}$`)
	patronDelECLI   = regexp.MustCompile(`^ECLI:ES:[A-Z][A-Z0-9]{0,6}:[0-9]{4}:[A-Z0-9.]{1,25}$`)
	patronDeLaFecha = regexp.MustCompile(`^([0-9]{4})-([0-9]{2})-([0-9]{2})$`)
	patronDelNumero = regexp.MustCompile(`^[0-9]+/[0-9]{4}$`)
)

// semillaDelCorpus devuelve el texto de una semilla del corpus versionado,
// que `go test` guarda como una línea de versión y un literal de Go.
func semillaDelCorpus(tb testing.TB, nombre string) string {
	tb.Helper()

	contenido, err := corpusDeLaFicha.ReadFile(carpetaDelCorpus + "/" + nombre)
	require.NoError(tb, err, "la semilla %s no está en el corpus", nombre)

	resto, conCabecera := strings.CutPrefix(string(contenido), "go test fuzz v1\nstring(")
	require.True(tb, conCabecera, "la semilla %s no es una cadena del corpus de go test", nombre)

	literal, conCierre := strings.CutSuffix(resto, ")\n")
	require.True(tb, conCierre, "la semilla %s no cierra su valor", nombre)

	texto, err := strconv.Unquote(literal)
	require.NoError(tb, err, "la semilla %s no es un literal de Go", nombre)

	return texto
}

// fragmento es el fragmento de la evidencia del ADR 0036 —el encabezamiento y
// el fallo de la sentencia 1088/2023 de la Sala de lo Civil del Tribunal
// Supremo—, leído de su semilla del corpus. Su huella es la del manifiesto de
// la evidencia: si la semilla dejara de ser el fragmento, byte a byte, todo
// test que lo usa fallaría aquí.
func fragmento(tb testing.TB) string {
	tb.Helper()

	texto := semillaDelCorpus(tb, "fragmento")
	huella := sha256.Sum256([]byte(texto))
	require.Equal(tb, huellaDelFragmento, hex.EncodeToString(huella[:]),
		"la semilla «fragmento» no es el fragmento de la evidencia del ADR 0036")

	return texto
}

// fichaDelFragmento es la ficha sola: las once primeras líneas del fragmento,
// de la de «Roj:» a la de «Tipo de Resolución:», que es lo que la skill pasa.
func fichaDelFragmento(tb testing.TB) string {
	tb.Helper()

	lineas := strings.SplitAfter(fragmento(tb), "\n")
	require.Greater(tb, len(lineas), lineasDeLaFicha)

	ficha := strings.Join(lineas[:lineasDeLaFicha], "")
	require.Len(tb, ficha, bytesDeLaFicha, "la ficha del fragmento no son sus once primeras líneas")

	return ficha
}

// conLaLinea devuelve el texto con la primera línea que lleva la etiqueta
// cambiada por otra; con la cadena vacía, sin ella.
func conLaLinea(tb testing.TB, texto, etiqueta, nueva string) string {
	tb.Helper()

	lineas := strings.SplitAfter(texto, "\n")
	for indice, linea := range lineas {
		if !strings.HasPrefix(linea, etiqueta+":") {
			continue
		}

		lineas[indice] = ""
		if nueva != "" {
			lineas[indice] = nueva + "\n"
		}

		return strings.Join(lineas, "")
	}

	require.FailNow(tb, "el texto no tiene ninguna línea con la etiqueta", etiqueta)

	return ""
}

// sinLaLinea devuelve el texto sin la primera línea que lleva la etiqueta.
func sinLaLinea(tb testing.TB, texto, etiqueta string) string {
	tb.Helper()

	return conLaLinea(tb, texto, etiqueta, "")
}

// leida es la ficha de un texto que la lleva.
func leida(tb testing.TB, texto string) cita.Ficha {
	tb.Helper()

	ficha, err := cita.LeerFicha(texto)
	require.NoError(tb, err)

	return ficha
}

// TestLeerFicha fija la lectura de la ficha con la que el CENDOJ encabeza un
// documento, con las reglas de FR-021 y de contracts/applet-cita.md §5 y
// ninguna más: la primera línea «Roj:» del texto, sus dos partes, y cada uno
// de los otros seis datos de la primera línea posterior que lleva su etiqueta
// entera; los finales de línea de los dos tipos y los blancos de los extremos
// no cuentan, y lo que hay antes de la ficha no se lee. Con uno de los ocho
// datos de menos, o con uno de los cuatro que se comparan sin su forma, no hay
// ficha: un error de argumentos que nombra el dato con su etiqueta, o que
// falta la ficha, con el valor cero al lado (H23, FR-021, FR-022, FR-026).
func TestLeerFicha(t *testing.T) {
	t.Parallel()

	t.Run("el fragmento de la evidencia", func(t *testing.T) {
		t.Parallel()

		ficha := leida(t, fragmento(t))
		compruebaSerializacion(t, fichaSerializada, ficha)
		compruebaEtiquetas(t, ficha)
	})

	t.Run("la misma ficha", func(t *testing.T) {
		t.Parallel()
		compruebaLaMismaFicha(t)
	})

	t.Run("la primera ficha del texto", func(t *testing.T) {
		t.Parallel()

		ficha := fichaDelFragmento(t)
		otra := conLaLinea(t, ficha, "Roj", lineaDeOtroOrgano)

		assert.Equal(t, leida(t, ficha), leida(t, ficha+"\n"+otra))
		assert.Equal(t, leida(t, otra), leida(t, otra+"\n"+ficha))
		assert.Equal(t, "SAP M 1234/2020", leida(t, otra+"\n"+ficha).ROJ)
	})

	t.Run("sin ficha", func(t *testing.T) {
		t.Parallel()
		compruebaTextosSinFicha(t)
	})

	t.Run("con un dato de menos", func(t *testing.T) {
		t.Parallel()
		compruebaDatosQueFaltan(t)
	})

	t.Run("con un dato sin su forma", func(t *testing.T) {
		t.Parallel()
		compruebaDatosSinSuForma(t)
	})
}

// compruebaLaMismaFicha fija lo que no cambia la lectura: lo que sigue a la
// ficha y lo que la precede, también cuando lleva las etiquetas de sus datos;
// el final de línea \r\n; los blancos de los extremos de cada línea, que son
// los de Unicode; el salto final; y una etiqueta repetida más abajo, de la que
// se lee la primera.
func compruebaLaMismaFicha(t *testing.T) {
	t.Helper()

	// espacioDuro es U+00A0, el blanco que deja un PDF copiado.
	const espacioDuro = "\xc2\xa0"

	// otrosDatos son las seis etiquetas que siguen a la línea «Roj:», con
	// valores que no son los de la ficha del fragmento.
	const otrosDatos = "Órgano: Otro Órgano\nFecha: 01/01/2000\nNº de Recurso: 1/1999\n" +
		"Nº de Resolución: 1/2000\nPonente: OTRO PONENTE\nTipo de Resolución: Auto\n"

	ficha := fichaDelFragmento(t)
	esperada := leida(t, fragmento(t))

	sinSaltoFinal := strings.TrimSuffix(ficha, "\n")
	conBlancos := "  \t" + strings.ReplaceAll(sinSaltoFinal, "\n", " \t\n  \t") + " \t\n"
	conEspaciosDuros := espacioDuro + strings.ReplaceAll(sinSaltoFinal, "\n", espacioDuro+"\n"+espacioDuro) + espacioDuro + "\n"

	casos := []struct{ nombre, texto string }{
		{"la ficha sola", ficha},
		{"con texto delante", "JURISPRUDENCIA\n\n" + ficha},
		{"con texto delante y detrás, sin línea en blanco", "JURISPRUDENCIA\n" + ficha + "TRIBUNAL SUPREMO\n"},
		{"con las etiquetas de sus datos delante, con otros valores", otrosDatos + ficha},
		{"con los finales de línea \\r\\n", strings.ReplaceAll(ficha, "\n", "\r\n")},
		{"con el fragmento entero y sus finales \\r\\n", strings.ReplaceAll(fragmento(t), "\n", "\r\n")},
		{"con blancos en los extremos de cada línea", conBlancos},
		{"con espacios duros en los extremos de cada línea", conEspaciosDuros},
		{"sin el salto final", sinSaltoFinal},
		{"con una etiqueta repetida más abajo", ficha + "Fecha: 01/01/2000\nPonente: OTRO PONENTE\n"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, esperada, leida(t, caso.texto))
		})
	}
}

// compruebaTextosSinFicha fija que un texto sin ninguna línea que empiece por
// «Roj:» no lleva ficha, y que el error lo dice. Las etiquetas se escriben
// como en el documento.
func compruebaTextosSinFicha(t *testing.T) {
	t.Helper()

	ficha := fichaDelFragmento(t)

	casos := []struct{ nombre, texto string }{
		{"vacío", ""},
		{"una frase cualquiera", "Una frase cualquiera, que no es el encabezamiento de ningún documento."},
		{"el fragmento sin su encabezamiento", strings.TrimPrefix(fragmento(t), ficha)},
		{"la etiqueta en minúsculas", strings.Replace(ficha, "Roj:", "roj:", 1)},
		{"la etiqueta sin sus dos puntos", strings.Replace(ficha, "Roj:", "Roj", 1)},
		{"un blanco entre la etiqueta y sus dos puntos", strings.Replace(ficha, "Roj:", "Roj :", 1)},
		{"la etiqueta en mitad de la línea", "Su " + ficha},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			leido, err := cita.LeerFicha(caso.texto)
			compruebaErrorDeArgumentos(t, err, "no lleva la ficha", "«Roj:»")
			assert.Equal(t, cita.Ficha{}, leido, "un texto sin ficha no puede acompañarse de datos")
		})
	}
}

// compruebaDatosQueFaltan fija que con cada uno de los ocho datos quitado no
// hay ficha y el error nombra el que falta con su etiqueta. «Nº de Recurso» y
// «Nº de Resolución» no se confunden: sin la línea de una, la de la otra no
// da su dato. Y un dato cuya primera línea va sin valor falta, aunque más
// abajo haya otra línea con su etiqueta.
func compruebaDatosQueFaltan(t *testing.T) {
	t.Helper()

	ficha := fichaDelFragmento(t)
	conElPonenteVacio := conLaLinea(t, ficha, "Ponente", "Ponente:")

	casos := []struct{ nombre, etiqueta, texto string }{
		{"sin el ROJ", "Roj", conLaLinea(t, ficha, "Roj", "Roj: - ECLI:ES:TS:2023:3144")},
		{"sin el ROJ ni el ECLI", "Roj", conLaLinea(t, ficha, "Roj", "Roj:")},
		{"sin el ECLI", "ECLI", conLaLinea(t, ficha, "Roj", "Roj: STS 3144/2023")},
		{"sin el ECLI, con su guion", "ECLI", conLaLinea(t, ficha, "Roj", "Roj: STS 3144/2023 - ")},
		{"sin la línea del órgano", "Órgano", sinLaLinea(t, ficha, "Órgano")},
		{"sin la línea de la fecha", "Fecha", sinLaLinea(t, ficha, "Fecha")},
		{"sin la línea del recurso", "Nº de Recurso", sinLaLinea(t, ficha, "Nº de Recurso")},
		{"sin la línea de la resolución", "Nº de Resolución", sinLaLinea(t, ficha, "Nº de Resolución")},
		{"sin la línea del ponente", "Ponente", sinLaLinea(t, ficha, "Ponente")},
		{"sin la línea del tipo", "Tipo de Resolución", sinLaLinea(t, ficha, "Tipo de Resolución")},
		{"con la línea del ponente vacía", "Ponente", conElPonenteVacio},
		{"con la línea del ponente vacía y otra con valor más abajo", "Ponente", conElPonenteVacio + "Ponente: OTRO PONENTE\n"},
		{"con un blanco entre la etiqueta del ponente y sus dos puntos", "Ponente", strings.Replace(ficha, "Ponente:", "Ponente :", 1)},
		{"con la línea de la fecha solo de blancos", "Fecha", conLaLinea(t, ficha, "Fecha", "Fecha:  \t")},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			leido, err := cita.LeerFicha(caso.texto)
			compruebaErrorDeArgumentos(t, err, "falta el dato «"+caso.etiqueta+"»")
			assert.Equal(t, cita.Ficha{}, leido, "una ficha sin un dato no puede acompañarse de los demás")
		})
	}
}

// compruebaDatosSinSuForma fija que con uno de los cuatro datos que se
// comparan sin su forma —el ROJ, el ECLI español, la fecha dd/mm/aaaa de un
// día que existe y el número de resolución— no hay ficha, y el error nombra
// el dato con su etiqueta y lo que el documento lleva.
func compruebaDatosSinSuForma(t *testing.T) {
	t.Helper()

	ficha := fichaDelFragmento(t)

	casos := []struct {
		nombre, etiqueta, linea string
		nombra                  []string
	}{
		{"el ROJ en minúsculas", "Roj", "Roj: sts 3144/2023 - ECLI:ES:TS:2023:3144", []string{`"sts 3144/2023"`}},
		{"el ECLI de otro país", "ECLI", "Roj: STS 3144/2023 - ECLI:FR:CC:2023:1", []string{`"ECLI:FR:CC:2023:1"`, "no es español"}},
		{"el ECLI con algo detrás", "ECLI", "Roj: STS 3144/2023 - ECLI:ES:TS:2023:3144 - STS", []string{`"ECLI:ES:TS:2023:3144 - STS"`}},
		{"el 31 de febrero", "Fecha", "Fecha: 31/02/2023", []string{`"31/02/2023"`, "dd/mm/aaaa"}},
		{"la fecha escrita AAAA-MM-DD", "Fecha", "Fecha: 2023-07-04", []string{`"2023-07-04"`, "dd/mm/aaaa"}},
		{"la fecha sin sus ceros", "Fecha", "Fecha: 4/7/2023", []string{`"4/7/2023"`, "dd/mm/aaaa"}},
		{"el número con un guion", "Nº de Resolución", "Nº de Resolución: 1088-2023", []string{`"1088-2023"`, "<número>/<año>"}},
		{"el número con el año de dos cifras", "Nº de Resolución", "Nº de Resolución: 1088/23", []string{`"1088/23"`, "<número>/<año>"}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			lineaDelDato := caso.etiqueta
			if caso.etiqueta == "ECLI" {
				lineaDelDato = "Roj"
			}

			leido, err := cita.LeerFicha(conLaLinea(t, ficha, lineaDelDato, caso.linea))
			compruebaErrorDeArgumentos(t, err, append(caso.nombra, "el dato «"+caso.etiqueta+"»")...)
			assert.Equal(t, cita.Ficha{}, leido, "una ficha con un dato sin su forma no puede acompañarse de los demás")
		})
	}
}

// entradaDelCorpus es una semilla del corpus versionado de FuzzLeerFicha: el
// nombre de su fichero y su texto.
type entradaDelCorpus struct {
	nombre, texto string
}

// semillasDeLaFicha son las del corpus, cada una con su fichero en
// testdata/fuzz/FuzzLeerFicha. Todas salen del fragmento de la evidencia con
// las mismas operaciones que usan los casos de TestLeerFicha, sin teclear el
// texto de ninguna sentencia: el fragmento y su ficha sola, la ficha con los
// finales \r\n, con texto delante y con blancos en los extremos, dos fichas
// seguidas, el fragmento sin su ficha, la ficha sin un dato, con una fecha que
// no existe y con otra primera línea, y la entrada vacía.
func semillasDeLaFicha(tb testing.TB) []entradaDelCorpus {
	tb.Helper()

	ficha := fichaDelFragmento(tb)
	otra := conLaLinea(tb, ficha, "Roj", lineaDeOtroOrgano)

	return []entradaDelCorpus{
		{"fragmento", fragmento(tb)},
		{"ficha-sola", ficha},
		{"finales-crlf", strings.ReplaceAll(ficha, "\n", "\r\n")},
		{"texto-delante", "JURISPRUDENCIA\n\n" + ficha},
		{"blancos-en-los-extremos", " \t" + strings.ReplaceAll(ficha, "\n", " \t\n \t")},
		{"dos-fichas", otra + "\n" + ficha},
		{"otro-organo", otra},
		{"sin-ficha", strings.TrimPrefix(fragmento(tb), ficha)},
		{"sin-ponente", sinLaLinea(tb, ficha, "Ponente")},
		{"fecha-imposible", conLaLinea(tb, ficha, "Fecha", "Fecha: 31/02/2023")},
		{"cadena-vacia", ""},
	}
}

// FuzzLeerFicha ejerce la lectura de la ficha sobre cualquier texto: ningún
// panic; todo rechazo es de argumentos y va con el valor cero; y lo aceptado
// tiene sus ocho datos con su forma, sale de un texto con una línea «Roj:» y,
// escrito de nuevo como una ficha, se lee igual. Sus semillas son las del
// corpus versionado, ni una más ni una menos (H23, FR-083).
func FuzzLeerFicha(f *testing.F) {
	semillas := semillasDeLaFicha(f)

	versionadas, err := corpusDeLaFicha.ReadDir(carpetaDelCorpus)
	require.NoError(f, err)
	require.Len(f, versionadas, len(semillas), "el corpus versionado no tiene un fichero por semilla")

	for _, semilla := range semillas {
		require.Equal(f, semilla.texto, semillaDelCorpus(f, semilla.nombre),
			"el fichero de la semilla %s no lleva su texto", semilla.nombre)
		f.Add(semilla.texto)
	}

	f.Fuzz(func(t *testing.T, texto string) {
		ficha, err := cita.LeerFicha(texto)
		if err != nil {
			compruebaErrorDeArgumentos(t, err)
			require.Equal(t, cita.Ficha{}, ficha, "una ficha rechazada no puede acompañarse de datos")

			return
		}

		require.Contains(t, texto, "Roj:", "se lee una ficha de un texto sin su primera línea")
		compruebaLaFormaDeLaFicha(t, ficha)

		otraVez, err := cita.LeerFicha(escrita(t, ficha))
		require.NoError(t, err, "no se lee la ficha escrita con los datos leídos")
		require.Equal(t, ficha, otraVez, "leer la ficha ya leída cambia sus datos")
	})
}

// compruebaLaFormaDeLaFicha exige de una ficha leída lo que FR-021 llama
// reconocible: el ROJ, el ECLI español, la fecha de un día que existe y el
// número de resolución con su forma, y los otros cuatro datos con texto, sin
// blancos alrededor y en una sola línea.
func compruebaLaFormaDeLaFicha(t *testing.T, ficha cita.Ficha) {
	t.Helper()

	require.Regexp(t, patronDelROJ, ficha.ROJ)
	require.Regexp(t, patronDelECLI, ficha.ECLI)
	require.Regexp(t, patronDelNumero, ficha.Resolucion)
	require.True(t, esUnDiaQueExiste(ficha.Fecha), "la fecha %q no es un día que existe, AAAA-MM-DD", ficha.Fecha)

	for _, texto := range []string{ficha.Organo, ficha.Recurso, ficha.Ponente, ficha.Tipo} {
		require.NotEmpty(t, texto, "un dato de texto va vacío")
		require.Equal(t, strings.TrimSpace(texto), texto, "un dato lleva blancos alrededor")
		require.NotContains(t, texto, "\n", "un dato ocupa más de una línea")
	}

	compruebaEtiquetas(t, ficha)
}

// esUnDiaQueExiste dice si la fecha es AAAA-MM-DD y nombra un día del
// calendario. No usa el análisis de fechas con el que el paquete las lee: da
// el día al calendario y mira que no lo haya corrido a otro.
func esUnDiaQueExiste(fecha string) bool {
	partes := patronDeLaFecha.FindStringSubmatch(fecha)
	if partes == nil {
		return false
	}

	anio, _ := strconv.Atoi(partes[1])
	mes, _ := strconv.Atoi(partes[2])
	dia, _ := strconv.Atoi(partes[3])

	enElCalendario := time.Date(anio, time.Month(mes), dia, 0, 0, 0, 0, time.UTC)

	return enElCalendario.Year() == anio && int(enElCalendario.Month()) == mes && enElCalendario.Day() == dia
}

// escrita es la ficha con sus ocho datos, como la escribe el CENDOJ: la línea
// «Roj:» con el ROJ y el ECLI y una línea por cada uno de los otros seis, con
// la fecha dd/mm/aaaa.
func escrita(t *testing.T, ficha cita.Ficha) string {
	t.Helper()

	partes := patronDeLaFecha.FindStringSubmatch(ficha.Fecha)
	require.NotNil(t, partes, "la fecha %q no es AAAA-MM-DD", ficha.Fecha)

	return strings.Join([]string{
		"Roj: " + ficha.ROJ + " - " + ficha.ECLI,
		"Órgano: " + ficha.Organo,
		"Fecha: " + partes[3] + "/" + partes[2] + "/" + partes[1],
		"Nº de Recurso: " + ficha.Recurso,
		"Nº de Resolución: " + ficha.Resolucion,
		"Ponente: " + ficha.Ponente,
		"Tipo de Resolución: " + ficha.Tipo,
		"",
	}, "\n")
}
