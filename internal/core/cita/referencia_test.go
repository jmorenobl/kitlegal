package cita_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/cita"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Lo que se pide en casi todos los casos: la sentencia de la evidencia del
// ADR 0036, por sus tres formas (docs/JURISPRUDENCIA.md §3).
const (
	ecliConocido       = "ECLI:ES:TS:2023:3144"
	rojConocido        = "STS 3144/2023"
	resolucionConocida = "1088/2023"
	fechaConocida      = "2023-07-04"
)

// argumentos son los cuatro con los que se da una referencia, cada uno dado o
// no dado: nil es el que no se escribió, y el que se escribió con valor vacío
// apunta a la cadena vacía (research.md D3).
type argumentos struct {
	ecli, roj, resolucion, fecha *string
}

// referencia construye la referencia de esos argumentos.
func (a argumentos) referencia() (cita.Referencia, bool, error) {
	return cita.NuevaReferencia(a.ecli, a.roj, a.resolucion, a.fecha)
}

// dado es un argumento escrito, también con valor vacío.
func dado(valor string) *string {
	return &valor
}

// TestReferencia fija cómo se da una referencia: de una sola forma de tres
// —un ECLI español, un ROJ o un número de resolución con su fecha—, desde
// cuatro argumentos que están dados o no dados. Sin ninguno no hay referencia
// y no es un error; el que llega escrito con valor vacío está dado, se
// comprueba con su forma y cuenta como una forma dada; y todo lo demás es un
// error de argumentos que nombra lo que dice su fila de contracts/applet-cita.md
// §6, con el valor cero al lado (H23, FR-005, FR-006; research.md D3).
func TestReferencia(t *testing.T) {
	t.Parallel()

	t.Run("sin ninguna forma no hay referencia", func(t *testing.T) {
		t.Parallel()

		referencia, hay, err := argumentos{}.referencia()
		require.NoError(t, err)
		assert.False(t, hay)
		assert.Equal(t, cita.Referencia{}, referencia)
	})

	t.Run("aceptadas", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre      string
			dados       argumentos
			serializada string
		}{
			{"un ECLI", argumentos{ecli: dado(ecliConocido)}, `{"forma":"ecli","valor":"ECLI:ES:TS:2023:3144"}`},
			{"un ECLI del Tribunal Constitucional", argumentos{ecli: dado("ECLI:ES:TC:2024:79")}, `{"forma":"ecli","valor":"ECLI:ES:TC:2024:79"}`},
			{"un ROJ", argumentos{roj: dado(rojConocido)}, `{"forma":"roj","valor":"STS 3144/2023"}`},
			{"un ROJ de otras siglas", argumentos{roj: dado("SAP M 1234/2020")}, `{"forma":"roj","valor":"SAP M 1234/2020"}`},
			{
				"un número con su fecha",
				argumentos{resolucion: dado(resolucionConocida), fecha: dado(fechaConocida)},
				`{"forma":"resolucion","valor":"1088/2023","fecha":"2023-07-04"}`,
			},
			{
				"un número con ceros por delante, tal como se dio",
				argumentos{resolucion: dado("0007/2023"), fecha: dado("2023-01-09")},
				`{"forma":"resolucion","valor":"0007/2023","fecha":"2023-01-09"}`,
			},
			{
				"el 29 de febrero de un año bisiesto",
				argumentos{resolucion: dado("1/2024"), fecha: dado("2024-02-29")},
				`{"forma":"resolucion","valor":"1/2024","fecha":"2024-02-29"}`,
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				referencia, hay, err := caso.dados.referencia()
				require.NoError(t, err)
				assert.True(t, hay)
				compruebaSerializacion(t, caso.serializada, referencia)
				compruebaEtiquetas(t, referencia)
			})
		}
	})

	t.Run("rechazadas", func(t *testing.T) {
		t.Parallel()

		for _, caso := range referenciasRechazadas() {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				referencia, hay, err := caso.dados.referencia()
				compruebaErrorDeArgumentos(t, err, caso.nombra...)
				assert.False(t, hay, "una referencia rechazada no es una referencia")
				assert.Equal(t, cita.Referencia{}, referencia, "una referencia rechazada no puede acompañarse de un valor")
			})
		}
	})

	t.Run("lo que la etiqueta enumera", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"ecli", "roj", "resolucion"}, enumeradoDe(t, reflect.TypeFor[cita.Referencia](), "Forma"))
	})
}

// referenciaRechazada es una referencia mal dada, con lo que su mensaje tiene
// que nombrar.
type referenciaRechazada struct {
	nombre string
	dados  argumentos
	nombra []string
}

// referenciasRechazadas son las filas de contracts/applet-cita.md §6 que
// decide la referencia, cada una con sus argumentos dados, no dados y dados
// vacíos.
func referenciasRechazadas() []referenciaRechazada {
	return slices.Concat(formasMalEscritas(), numerosYFechasSueltos(), fechasMalEscritas(), masDeUnaForma())
}

// formasMalEscritas son un ECLI, un ROJ y un número que no tienen su forma, y
// la vacía no la tiene: el mensaje nombra la entrada entre comillas.
func formasMalEscritas() []referenciaRechazada {
	const (
		formaDelROJ    = "<siglas> <número>/<año>"
		formaDelNumero = "<número>/<año>"
	)

	delNumero := func(nombre, numero string) referenciaRechazada {
		return referenciaRechazada{
			nombre: nombre,
			dados:  argumentos{resolucion: dado(numero), fecha: dado(fechaConocida)},
			nombra: []string{"el número de resolución " + strconv.Quote(numero), formaDelNumero},
		}
	}

	return []referenciaRechazada{
		{"un ECLI vacío", argumentos{ecli: dado("")}, []string{`el ECLI ""`, "ECLI:ES:<órgano>:<año>:<número>"}},
		{"un ECLI mal formado", argumentos{ecli: dado("ECLI:ES:TS:2023")}, []string{`el ECLI "ECLI:ES:TS:2023"`, "ECLI:ES:<órgano>:<año>:<número>"}},
		{"un ECLI en minúsculas", argumentos{ecli: dado("ecli:es:ts:2023:3144")}, []string{`el ECLI "ecli:es:ts:2023:3144"`, "ECLI:ES:<órgano>:<año>:<número>"}},
		{"un ECLI con un blanco", argumentos{ecli: dado(" " + ecliConocido)}, []string{`el ECLI " ECLI:ES:TS:2023:3144"`}},
		{"un ECLI de otro país", argumentos{ecli: dado("ECLI:FR:CC:2023:1")}, []string{`el ECLI "ECLI:FR:CC:2023:1"`, "no es español"}},
		{"un ROJ vacío", argumentos{roj: dado("")}, []string{`el ROJ ""`, formaDelROJ}},
		{"un ROJ sin su año", argumentos{roj: dado("STS 3144")}, []string{`el ROJ "STS 3144"`, formaDelROJ}},
		{"un ROJ en minúsculas", argumentos{roj: dado("sts 3144/2023")}, []string{`el ROJ "sts 3144/2023"`}},
		delNumero("un número vacío", ""),
		delNumero("un número sin barra", "1088-2023"),
		delNumero("un número sin su número", "/2023"),
		delNumero("un número con letras", "1088A/2023"),
		delNumero("un número con el año de dos cifras", "1088/23"),
		delNumero("un número con el año de cinco cifras", "1088/20233"),
		delNumero("un número con dos barras", "1088/20/23"),
		delNumero("un número con un blanco detrás", "1088/2023 "),
		delNumero("un número con cifras que no son ASCII", "١٠٨٨/2023"),
		{
			"un número mal escrito, también sin su fecha",
			argumentos{resolucion: dado("1088")},
			[]string{`el número de resolución "1088"`, formaDelNumero},
		},
	}
}

// numerosYFechasSueltos son el número sin su fecha y la fecha sin su número:
// la fecha es obligatoria con el número y solo vale con él, también vacía y
// también junto a otra forma.
func numerosYFechasSueltos() []referenciaRechazada {
	const soloConElNumero = "solo vale con un número de resolución"

	return []referenciaRechazada{
		{"un número sin su fecha", argumentos{resolucion: dado(resolucionConocida)}, []string{"el número de resolución 1088/2023", "necesita su fecha"}},
		{"una fecha sola", argumentos{fecha: dado(fechaConocida)}, []string{`la fecha "2023-07-04"`, soloConElNumero}},
		{"una fecha vacía sola", argumentos{fecha: dado("")}, []string{`la fecha ""`, soloConElNumero}},
		{"una fecha con un ECLI", argumentos{ecli: dado(ecliConocido), fecha: dado(fechaConocida)}, []string{`la fecha "2023-07-04"`, soloConElNumero}},
		{"una fecha vacía con un ECLI", argumentos{ecli: dado(ecliConocido), fecha: dado("")}, []string{`la fecha ""`, soloConElNumero}},
		{"una fecha con un ROJ", argumentos{roj: dado(rojConocido), fecha: dado(fechaConocida)}, []string{`la fecha "2023-07-04"`, soloConElNumero}},
	}
}

// fechasMalEscritas son las fechas que no tienen su forma o que no son un día
// que existe, con un número bien escrito: el mensaje nombra la fecha entre
// comillas y su forma.
func fechasMalEscritas() []referenciaRechazada {
	fechas := []struct{ nombre, fecha string }{
		{"una fecha vacía", ""},
		{"una fecha como la escribe el buscador", "04/07/2023"},
		{"una fecha sin sus ceros", "2023-7-4"},
		{"una fecha con el año de dos cifras", "23-07-04"},
		{"una fecha con hora", "2023-07-04T00:00:00"},
		{"una fecha con un blanco detrás", "2023-07-04 "},
		{"el 31 de febrero", "2023-02-31"},
		{"el 29 de febrero de un año que no es bisiesto", "2023-02-29"},
		{"el mes 13", "2023-13-01"},
		{"el día 0", "2023-07-00"},
	}

	casos := make([]referenciaRechazada, 0, len(fechas))
	for _, fecha := range fechas {
		casos = append(casos, referenciaRechazada{
			nombre: fecha.nombre,
			dados:  argumentos{resolucion: dado(resolucionConocida), fecha: dado(fecha.fecha)},
			nombra: []string{"la fecha " + strconv.Quote(fecha.fecha), "AAAA-MM-DD"},
		})
	}

	return casos
}

// masDeUnaForma son las referencias dadas de más de una forma, también si
// alguna va vacía: el mensaje nombra las formas dadas.
func masDeUnaForma() []referenciaRechazada {
	const unaSola = "de una sola forma"

	return []referenciaRechazada{
		{"un ECLI y un ROJ", argumentos{ecli: dado(ecliConocido), roj: dado(rojConocido)}, []string{unaSola, "un ECLI y un ROJ"}},
		{"un ECLI y un ROJ vacío", argumentos{ecli: dado(ecliConocido), roj: dado("")}, []string{unaSola, "un ECLI y un ROJ"}},
		{"un ECLI vacío y un ROJ", argumentos{ecli: dado(""), roj: dado(rojConocido)}, []string{unaSola, "un ECLI y un ROJ"}},
		{
			"un ECLI y un número con su fecha",
			argumentos{ecli: dado(ecliConocido), resolucion: dado(resolucionConocida), fecha: dado(fechaConocida)},
			[]string{unaSola, "un ECLI y un número de resolución"},
		},
		{
			"un ROJ y un número vacío",
			argumentos{roj: dado(rojConocido), resolucion: dado("")},
			[]string{unaSola, "un ROJ y un número de resolución"},
		},
		{
			"las tres",
			argumentos{ecli: dado(ecliConocido), roj: dado(rojConocido), resolucion: dado(resolucionConocida), fecha: dado(fechaConocida)},
			[]string{unaSola, "un ECLI, un ROJ y un número de resolución"},
		},
	}
}

// compruebaErrorDeArgumentos exige que el error exista, que declare la clase
// «argumentos» —la que el kernel traduce a código 2: el dominio no importa
// internal/cli ni en sus tests, así que el código no se comprueba aquí— y que
// su mensaje nombre cada tramo.
func compruebaErrorDeArgumentos(tb testing.TB, err error, nombra ...string) {
	tb.Helper()

	require.Error(tb, err)

	var conClase schema.ConClase
	require.ErrorAs(tb, err, &conClase, "el rechazo no declara su clase: %v", err)
	assert.Equal(tb, schema.ClaseArgumentos, conClase.Clase(), "el rechazo no es de argumentos: %v", err)

	for _, tramo := range nombra {
		assert.Contains(tb, err.Error(), tramo, "el mensaje no nombra lo que dice su fila")
	}
}

// compruebaSerializacion exige que el valor se escriba con esos bytes: las
// claves de contracts/applet-cita.md §3 y §4 en su orden, sin la que no
// aplica. Compara los bytes y no el documento, porque el orden de las claves
// es parte de lo que la suite de aceptación afirma.
func compruebaSerializacion(tb testing.TB, esperada string, valor any) {
	tb.Helper()

	obtenida, err := json.Marshal(valor)
	require.NoError(tb, err)
	assert.Equal(tb, esperada, string(obtenida))
}

// enumeradoDe son los valores enum de la etiqueta jsonschema de un campo, en
// su orden: los que --describe lleva al esquema del verbo (research.md V6).
func enumeradoDe(tb testing.TB, tipo reflect.Type, nombre string) []string {
	tb.Helper()

	campo, existe := tipo.FieldByName(nombre)
	require.True(tb, existe, "%s no tiene el campo %s", tipo, nombre)

	var valores []string

	for parte := range strings.SplitSeq(campo.Tag.Get("jsonschema"), ",") {
		if valor, esEnum := strings.CutPrefix(parte, "enum="); esEnum {
			valores = append(valores, valor)
		}
	}

	return valores
}

// compruebaEtiquetas exige que lo que el dominio devuelve cumpla lo que
// declaran sus propias etiquetas jsonschema, que son de las que --describe
// genera el esquema publicado del verbo: sin esto, una salida que su esquema
// rechaza no se vería hasta componer el applet.
func compruebaEtiquetas(tb testing.TB, valor any) {
	tb.Helper()

	assert.Empty(tb, incumplimientos(reflect.ValueOf(valor), ""), "la salida no cumple sus etiquetas jsonschema")
}

// incumplimientos revisa un valor y devuelve, con su ruta, cada dato que no
// cumple la restricción de su etiqueta jsonschema. Lo que no se serializa
// —un campo sin exportar, o uno con omitempty que va vacío— no se mira.
func incumplimientos(valor reflect.Value, ruta string) []string {
	var fallos []string

	switch valor.Kind() {
	case reflect.Pointer:
		if !valor.IsNil() {
			fallos = incumplimientos(valor.Elem(), ruta)
		}
	case reflect.Slice:
		for indice := range valor.Len() {
			fallos = append(fallos, incumplimientos(valor.Index(indice), ruta+"/"+strconv.Itoa(indice))...)
		}
	case reflect.Struct:
		for indice := range valor.NumField() {
			fallos = append(fallos, incumplimientosDelCampo(valor.Type().Field(indice), valor.Field(indice), ruta)...)
		}
	default:
		// Un escalar suelto no tiene etiqueta: la lleva el campo que lo contiene.
	}

	return fallos
}

// incumplimientosDelCampo comprueba el valor de un campo contra su etiqueta
// —enum, pattern, minLength y minItems, que son las que el paquete usa— y
// sigue por dentro de él.
func incumplimientosDelCampo(campo reflect.StructField, valor reflect.Value, ruta string) []string {
	clave, opciones, _ := strings.Cut(campo.Tag.Get("json"), ",")
	if !campo.IsExported() || (opciones == "omitempty" && valor.IsZero()) {
		return nil
	}

	ruta += "/" + clave

	var fallos, enumerado []string

	for parte := range strings.SplitSeq(campo.Tag.Get("jsonschema"), ",") {
		regla, argumento, _ := strings.Cut(parte, "=")

		switch regla {
		case "enum":
			enumerado = append(enumerado, argumento)
		case "pattern":
			if !regexp.MustCompile(argumento).MatchString(valor.String()) {
				fallos = append(fallos, fmt.Sprintf("%s: %q no casa con %s", ruta, valor.String(), argumento))
			}
		case "minLength", "minItems":
			if minimo, _ := strconv.Atoi(argumento); valor.Len() < minimo {
				fallos = append(fallos, fmt.Sprintf("%s: tiene %d y el mínimo es %d", ruta, valor.Len(), minimo))
			}
		}
	}

	if len(enumerado) > 0 && !slices.Contains(enumerado, valor.String()) {
		fallos = append(fallos, fmt.Sprintf("%s: %q no es ninguno de %q", ruta, valor.String(), enumerado))
	}

	return append(fallos, incumplimientos(valor, ruta)...)
}
