package fiestas

import (
	"encoding/json/v2"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// Dónde están, relativos al directorio del paquete, lo que grabó grabar_datos y
// el manifiesto con el que lo grabó (contracts/datos-de-plazos.md §1.3 y §4.1).
const (
	raizDeLoGrabado     = "../../../../evidencias/h9/" + NombreDeLaFuente
	rutaDelManifiesto   = "testdata/grabaciones.json"
	numeroDeComunidades = 19
)

// manifiestoDelPaquete es testdata/grabaciones.json, leído como lo lee la
// derivación: sin miembros que no sean campos del manifiesto.
func manifiestoDelPaquete(t *testing.T) Manifiesto {
	t.Helper()

	contenido, err := os.ReadFile(rutaDelManifiesto)
	require.NoError(t, err)

	var manifiesto Manifiesto
	require.NoError(t, json.Unmarshal(contenido, &manifiesto, json.RejectUnknownMembers(true)))

	return manifiesto
}

// clienteDeLoGrabado reproduce, sin red, lo que grabó grabar_datos.
func clienteDeLoGrabado(t *testing.T) *httpx.Cliente {
	t.Helper()

	cliente, err := httpx.Replay(raizDeLoGrabado, httpx.ConFuente(NombreDeLaFuente))
	require.NoError(t, err)

	return cliente
}

// resolucionGrabada es la resolución grabada del año, leída con LeerResolucion.
func resolucionGrabada(t *testing.T, anio int) Resolucion {
	t.Helper()

	manifiesto := manifiestoDelPaquete(t)
	indice := slices.IndexFunc(manifiesto.Disposiciones, func(d Disposicion) bool { return d.Anio == anio })
	require.GreaterOrEqual(t, indice, 0, "el manifiesto no lista la resolución de %d", anio)

	respuesta, err := clienteDeLoGrabado(t).Pedir(t.Context(), schema.Contexto{},
		httpx.Peticion{Metodo: metodoDeLaLectura, URL: manifiesto.Disposiciones[indice].URL})
	require.NoError(t, err)

	resolucion, err := LeerResolucion(respuesta.Cuerpo)
	require.NoError(t, err)

	return resolucion
}

// columnaDe es la posición en la cabecera de la comunidad que la tabla escribe
// con ese nombre.
func columnaDe(t *testing.T, resolucion Resolucion, nombre string) int {
	t.Helper()

	indice := slices.IndexFunc(resolucion.Cabecera, func(c Columna) bool { return c.Nombre == nombre })
	require.GreaterOrEqual(t, indice, 0, "la cabecera de %d no tiene «%s»", resolucion.Anio, nombre)

	return indice
}

// filaDe es la fila leída de una fecha cuya primera celda empieza por el
// principio dado.
func filaDe(t *testing.T, resolucion Resolucion, fecha, principio string) Fila {
	t.Helper()

	indice := slices.IndexFunc(resolucion.Filas, func(f Fila) bool {
		return f.Fecha == fecha && strings.HasPrefix(f.Texto, principio)
	})
	require.GreaterOrEqual(t, indice, 0, "no se ha leído la fila «%s» del %s", principio, fecha)

	return resolucion.Filas[indice]
}

// numerosDe son los números de las notas que quedan fuera.
func numerosDe(notas []NotaSinLeer) []int {
	numeros := make([]int, 0, len(notas))
	for _, nota := range notas {
		numeros = append(numeros, nota.Numero)
	}

	return numeros
}

// TestLeerResolucion lee el anexo de las tres resoluciones grabadas y fija lo
// que la derivación necesita de cada una (FR-011, FR-020, FR-025; research D14):
// el año del anexo, una columna por comunidad o ciudad autónoma, todas sus filas
// leídas, el 12 de octubre de 2026 nacional en todas, una fiesta de una
// comunidad, las fiestas que una nota limita a una parte de la comunidad y las
// notas que quedan fuera por no leerse sin interpretarlas. Con documentos
// sintéticos de la misma forma fija cada regla de la lectura: lo que no se lee
// del documento es un error, y la fila o la nota que no se lee queda fuera con
// su motivo (ADR 0039, punto 3).
func TestLeerResolucion(t *testing.T) {
	t.Parallel()

	t.Run("cada-año-grabado", func(t *testing.T) {
		t.Parallel()

		// Las citas que no limitan una fiesta a una parte, solo en 2025: la de la
		// fiesta recuperable de Cataluña, en la nota 2, y la de la Comunitat
		// Valenciana, la nota 3.
		notasFuera := map[int][]int{2025: {2, 3}, 2026: {}, 2027: {}}
		// La fiesta que la nota 2, la de Cataluña, sustituye cada año en el
		// territorio de Arán por el 17 de junio.
		sustituida := map[int]string{2025: "2025-12-26", 2026: "2026-12-26", 2027: "2027-03-29"}

		for _, anio := range []int{2025, 2026, 2027} {
			resolucion := resolucionGrabada(t, anio)

			assert.Equal(t, anio, resolucion.Anio)
			assert.Len(t, resolucion.Cabecera, numeroDeComunidades)
			assert.Empty(t, resolucion.FilasSinLeer, "%d: ninguna fila de lo grabado queda fuera", anio)
			assert.Equal(t, notasFuera[anio], numerosDe(resolucion.NotasSinLeer), "%d", anio)

			// La nota 1, la de Canarias, da cada año una fiesta por isla o grupo
			// de islas: siete, entre ellas la de Gran Canaria, el 8 de
			// septiembre; y la 2, la de Cataluña, dos (FR-025).
			canarias := columnaDe(t, resolucion, "Canarias")
			assert.Equal(t, []int{1}, resolucion.Cabecera[canarias].Notas, "%d", anio)
			assert.Len(t, resolucion.Partes, 9, "%d", anio)
			assert.Contains(t, resolucion.Partes, FiestaDeUnaParte{
				Fecha: fmt.Sprintf("%d-09-08", anio), Parte: "Gran Canaria", Columna: canarias, Nota: 1,
			}, "%d", anio)

			// La sustitución de la nota 2: la fiesta sustituida deja de ser de
			// toda Cataluña y queda como del resto de ella, y el 17 de junio,
			// como del territorio de Arán (FR-011).
			cataluna := columnaDe(t, resolucion, "Cataluña")
			assert.Equal(t, []int{2}, resolucion.Cabecera[cataluna].Notas, "%d", anio)
			assert.Equal(t, []FiestaDeUnaParte{
				{Fecha: sustituida[anio], Parte: "la comunidad salvo el territorio de Arán", Columna: cataluna, Nota: 2},
				{Fecha: fmt.Sprintf("%d-06-17", anio), Parte: "el territorio de Arán", Columna: cataluna, Nota: 2},
			}, resolucion.Partes[7:], "%d", anio)

			fila := slices.IndexFunc(resolucion.Filas, func(f Fila) bool { return f.Fecha == sustituida[anio] })
			require.GreaterOrEqual(t, fila, 0, "%d: la fila de la fiesta sustituida se lee", anio)
			assert.Empty(t, resolucion.Filas[fila].Ambitos[cataluna], "%d: no es fiesta de toda Cataluña", anio)
			assert.Contains(t, resolucion.Filas[fila].Ambitos, plazos.AmbitoAutonomico,
				"%d: sigue siéndolo de otra comunidad", anio)
		}
	})

	t.Run("2026", func(t *testing.T) {
		t.Parallel()

		resolucion := resolucionGrabada(t, 2026)

		nacional := filaDe(t, resolucion, "2026-10-12", "12 Fiesta Nacional de España")
		assert.Equal(t, slices.Repeat([]plazos.Ambito{plazos.AmbitoNacional}, numeroDeComunidades), nacional.Ambitos,
			"el 12 de octubre de 2026 es nacional en las %d", numeroDeComunidades)

		madrid := filaDe(t, resolucion, "2026-05-02", "2 Fiesta de la Comunidad de Madrid")
		for indice, ambito := range madrid.Ambitos {
			if indice == columnaDe(t, resolucion, "Com. Madrid") {
				assert.Equal(t, plazos.AmbitoAutonomico, ambito)
			} else {
				assert.Empty(t, ambito, "el 2 de mayo solo es de la Comunidad de Madrid")
			}
		}

		assert.Contains(t, resolucion.Partes, FiestaDeUnaParte{
			Fecha: "2026-09-24", Parte: "El Hierro", Columna: columnaDe(t, resolucion, "Canarias"), Nota: 1,
		})
		assert.Contains(t, resolucion.Partes, FiestaDeUnaParte{
			Fecha: "2026-09-15", Parte: "Lanzarote y La Graciosa", Columna: columnaDe(t, resolucion, "Canarias"), Nota: 1,
		})
	})

	t.Run("forma-sintetica", probarFormaSintetica)
	t.Run("documento-que-no-se-lee", probarDocumentoQueNoSeLee)
	t.Run("fila-fuera", probarFilaFuera)
	t.Run("nota-fuera", probarNotaFuera)
	t.Run("cita-fuera", probarCitaFuera)
	t.Run("sustitucion", probarSustitucion)
}

// Los párrafos que siguen a la tabla en las tres resoluciones grabadas: la
// leyenda de los códigos y el principio de las notas.
const (
	leyendaGrabada = `<p class="parrafo">Códigos de las fiestas:</p>` +
		`<p class="parrafo_2">Fiesta Nacional no sustituible (*).</p>` +
		`<p class="parrafo">Fiesta Nacional respecto de la que no se ha ejercido la facultad de sustitución (**).</p>` +
		`<p class="parrafo">Fiesta de Comunidad Autónoma (***).</p>`
	parrafoDeLasNotas = `<p class="parrafo_2">Notas aclaratorias:</p>`
	// notaDeCanarias tiene la forma de la nota 1 de las tres grabadas, con
	// dos islas.
	notaDeCanarias = `<p class="parrafo_2">1. En la Comunidad Autónoma de Canarias, el Decreto 1/2026 ` +
		`(<abbr title="Boletín Oficial de Canarias">BOC</abbr> de 5 de mayo de 2025) dispone que: «En las islas de El ` +
		`Hierro y Tenerife, las fiestas laborales serán, además, las siguientes: en El Hierro: el 24 de septiembre, ` +
		`festividad de Nuestra Señora de los Reyes; en Tenerife: el 2 de febrero, festividad de la Virgen de la ` +
		`Candelaria».</p>`
)

// sintetica es una resolución con la forma de las grabadas: el título del
// anexo, el nombre de cada comunidad de la cabecera, las celdas de cada fila y
// lo que va tras la tabla.
type sintetica struct {
	titulo   string
	cabecera []string
	filas    [][]string
	tras     string
}

// basica es una resolución sintética que se lee entera: dos comunidades, una de
// ellas con una nota que da dos fiestas de una parte, y una fila de cada código.
func basica() sintetica {
	return sintetica{
		titulo:   "Año 2026",
		cabecera: []string{"Aragón", "\n<p class=\"cabeza_tabla\">Canarias</p>\n<p class=\"cabeza_tabla\">(1)</p>\n"},
		filas: [][]string{
			{"Enero", "&nbsp;", " "},
			{"1 Año Nuevo.", "\n<abbr title=\"Fiesta Nacional no sustituible\">*</abbr>\n", "*"},
			{"6 Epifanía del Señor.", "**", "**"},
			{"Abril", "", ""},
			{"23 San Jorge/Día de Aragón.", "***", " "},
			{"Mayo", "", ""},
			{"30 Día de Canarias.", "", "***"},
		},
		tras: leyendaGrabada + parrafoDeLasNotas + notaDeCanarias,
	}
}

// documento es la página del diario con la resolución: lo que precede al texto
// —una cabecera con un guion que no es XML—, el texto con el anexo y lo que lo
// sigue.
func (s sintetica) documento() []byte {
	var html strings.Builder

	html.WriteString(`<!DOCTYPE html><html><head><script>if (a < b && c) {}</script></head><body>`)
	html.WriteString(`<div id="DOdocText" class="panel"><div id="textoxslt">`)
	html.WriteString(`<p class="parrafo">Esta Dirección General de Trabajo ha resuelto:</p>`)
	html.WriteString(`<h5 class="anexo_num">ANEXO</h5><h5 class="anexo_tit">` + s.titulo + `</h5>`)
	html.WriteString(`<table class="tabla_girada_condensada"><colgroup><col width="17%"/></colgroup><thead><tr>`)
	html.WriteString(`<th rowspan="2">Fecha de las fiestas</th><th colspan="2">Comunidades Autónomas</th></tr><tr>`)

	for _, nombre := range s.cabecera {
		html.WriteString("<th>" + nombre + "</th>")
	}

	html.WriteString("</tr></thead><tbody>")

	for _, fila := range s.filas {
		html.WriteString("<tr>")

		for _, celda := range fila {
			html.WriteString(`<td class="cuerpo_tabla_centro">` + celda + "</td>")
		}

		html.WriteString("</tr>")
	}

	html.WriteString("</tbody></table>" + s.tras + `</div><!-- #textoxslt --></div></body></html>`)

	return []byte(html.String())
}

// probarFormaSintetica lee la resolución básica entera: el año, la cabecera
// sin las llamadas a las notas, cada fila con el ámbito de su código y las
// fiestas que la nota limita a una parte de la comunidad que la llama.
func probarFormaSintetica(t *testing.T) {
	t.Parallel()

	resolucion, err := LeerResolucion(basica().documento())
	require.NoError(t, err)

	nacional, autonomico := plazos.AmbitoNacional, plazos.AmbitoAutonomico

	assert.Equal(t, Resolucion{
		Anio:     2026,
		Cabecera: []Columna{{Nombre: "Aragón"}, {Nombre: "Canarias", Notas: []int{1}}},
		Filas: []Fila{
			{Fecha: "2026-01-01", Texto: "1 Año Nuevo.", Ambitos: []plazos.Ambito{nacional, nacional}},
			{Fecha: "2026-01-06", Texto: "6 Epifanía del Señor.", Ambitos: []plazos.Ambito{nacional, nacional}},
			{Fecha: "2026-04-23", Texto: "23 San Jorge/Día de Aragón.", Ambitos: []plazos.Ambito{autonomico, ""}},
			{Fecha: "2026-05-30", Texto: "30 Día de Canarias.", Ambitos: []plazos.Ambito{"", autonomico}},
		},
		Partes: []FiestaDeUnaParte{
			{Fecha: "2026-09-24", Parte: "El Hierro", Columna: 1, Nota: 1},
			{Fecha: "2026-02-02", Parte: "Tenerife", Columna: 1, Nota: 1},
		},
	}, resolucion)
}

// probarDocumentoQueNoSeLee fija cada forma del documento que no es la de una
// resolución de fiestas laborales: sin ella no hay nada que leer, y es un error
// que dice qué falta, nunca una resolución vacía.
func probarDocumentoQueNoSeLee(t *testing.T) {
	t.Parallel()

	documento := string(basica().documento())
	sinCuerpo, _, _ := strings.Cut(documento, "<tbody>")
	conTras := func(tras string) string {
		s := basica()
		s.tras = tras

		return string(s.documento())
	}
	sinComunidades := basica()
	sinComunidades.cabecera = nil

	casos := []struct {
		nombre, documento, motivo string
	}{
		{"sin-texto", "<html><body><p>Resolución</p></body></html>", "textoxslt"},
		{"texto-sin-terminar", sinCuerpo, "no se lee entero"},
		{"sin-año", strings.Replace(documento, "Año 2026", "Año", 1), "título del anexo"},
		{"sin-tabla", strings.ReplaceAll(strings.ReplaceAll(documento, "<table", "<div"), "</table>", "</div>"), "tabla"},
		{"dos-tablas", strings.Replace(documento, "</table>", "</table><table></table>", 1), "tabla"},
		{"sin-cabecera", strings.NewReplacer("<thead>", "", "</thead>", "").Replace(documento), "thead"},
		{"sin-cuerpo", strings.NewReplacer("<tbody>", "", "</tbody>", "").Replace(documento), "tbody"},
		{"cabecera-vacia", string(sinComunidades.documento()), "ninguna comunidad"},
		{"sin-leyenda", conTras(parrafoDeLasNotas + notaDeCanarias), "Códigos de las fiestas"},
		{
			"leyenda-cambiada", conTras(strings.NewReplacer("(*).", "(***).", "(***).", "(*).").Replace(leyendaGrabada)),
			"significado",
		},
		{"leyenda-con-otro-codigo", conTras(leyendaGrabada + "<p>Fiesta local (****).</p>"), "leyenda"},
		{"codigo-repetido", conTras(leyendaGrabada + "<p>Fiesta Nacional no sustituible (*).</p>"), "dos veces"},
		{"parrafo-tras-la-leyenda", conTras(leyendaGrabada + "<p>Madrid, 17 de octubre de 2025.</p>"), "leyenda"},
		{"elemento-tras-la-tabla", conTras(leyendaGrabada + "<h5>ANEXO II</h5>"), "<h5>"},
		{"nota-sin-numero", conTras(leyendaGrabada + parrafoDeLasNotas + "<p>Sin número.</p>"), "número"},
		{"nota-repetida", conTras(leyendaGrabada + parrafoDeLasNotas + notaDeCanarias + notaDeCanarias), "dos veces"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			resolucion, err := LeerResolucion([]byte(caso.documento))
			require.ErrorContains(t, err, caso.motivo)
			assert.Zero(t, resolucion)
		})
	}
}

// probarFilaFuera fija cada fila que no se lee sin interpretarla: queda fuera
// con su primera celda y su motivo, y las demás se siguen leyendo (ADR 0039,
// punto 3).
func probarFilaFuera(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		filas  [][]string
		texto  string
		motivo string
	}{
		{"mes-con-marcas", [][]string{{"Enero", "*", ""}}, "Enero", "mes"},
		{"sin-mes", [][]string{{"1 Año Nuevo.", "*", "*"}}, "1 Año Nuevo.", "mes"},
		{"dia-que-no-existe", [][]string{{"Abril", "", ""}, {"31 Día de nadie.", "*", ""}}, "31 Día de nadie.", "no existe"},
		{"celdas-de-menos", [][]string{{"Enero", "", ""}, {"1 Año Nuevo.", "*"}}, "1 Año Nuevo.", "celdas"},
		{"celda-sin-codigo", [][]string{{"Enero", "", ""}, {"1 Año Nuevo.", "*", "*(1)"}}, "1 Año Nuevo.", "código"},
		{"sin-marcas", [][]string{{"Enero", "", ""}, {"1 Año Nuevo.", "", ""}}, "1 Año Nuevo.", "ninguna comunidad"},
		{"ni-dia-ni-mes", [][]string{{"Enero", "", ""}, {"Año Nuevo.", "*", "*"}}, "Año Nuevo.", "día"},
		{"sin-celdas", [][]string{{}}, "", "celdas"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			s := basica()
			s.filas = slices.Concat(caso.filas, [][]string{{"Mayo", "", ""}, {"30 Día de Canarias.", "", "***"}})

			resolucion, err := LeerResolucion(s.documento())
			require.NoError(t, err)

			require.Len(t, resolucion.FilasSinLeer, 1)
			assert.Equal(t, caso.texto, resolucion.FilasSinLeer[0].Texto)
			assert.Contains(t, resolucion.FilasSinLeer[0].Motivo, caso.motivo)
			assert.Equal(t, []Fila{{
				Fecha: "2026-05-30", Texto: "30 Día de Canarias.",
				Ambitos: []plazos.Ambito{"", plazos.AmbitoAutonomico},
			}}, resolucion.Filas,
				"la fila que queda fuera no cambia las demás")
		})
	}
}

// parrafoDeNota es un párrafo de las notas aclaratorias.
func parrafoDeNota(texto string) string { return "<p>" + texto + "</p>" }

// notaDeCanariasCon es la nota 1, de Canarias, con una sola cita.
func notaDeCanariasCon(cita string) string {
	return parrafoDeNota("1. En la Comunidad Autónoma de Canarias, dispone que: «" + cita + "».")
}

// probarNotaFuera fija cada nota, o cita de una nota, que no se lee sin
// interpretarla: queda fuera con el número de la nota y su motivo, y no da
// ninguna fiesta de una parte ni cambia la tabla (FR-025; ADR 0039, punto 3).
func probarNotaFuera(t *testing.T) {
	t.Parallel()

	leida, err := LeerResolucion(basica().documento())
	require.NoError(t, err)

	casos := []struct {
		nombre   string
		cabecera []string
		notas    string
		motivo   string
	}{
		{
			"sustitucion-de-un-dia-sin-fiesta", nil, notaDeCanariasCon("En el territorio de El Hierro, la fiesta del día 26 de " +
				"diciembre queda sustituida por la de 17 de junio, miércoles"), "no marca como fiesta de Canarias",
		},
		{
			"sustitucion-de-una-fiesta-nacional", nil, notaDeCanariasCon("En el territorio de El Hierro, la fiesta del día 6 de " +
				"enero (Epifanía del Señor) queda sustituida por la de 17 de junio"), "no marca como fiesta de Canarias",
		},
		{
			"sustitucion-con-un-dia-que-no-existe", nil, notaDeCanariasCon("En el territorio de El Hierro, la fiesta del día 30 de " +
				"mayo queda sustituida por la de 31 de junio"), "no da dos días de 2026",
		},
		{
			"sin-parte", nil, notaDeCanariasCon("Con carácter recuperable se establece la siguiente fiesta: 24 de junio, San Juan"),
			"fiesta de una parte",
		},
		{"dia-que-no-existe", nil, notaDeCanariasCon("Serán, además: en El Hierro: el 31 de septiembre, festividad"), "día de 2026"},
		{"mes-que-no-es-un-mes", nil, notaDeCanariasCon("Serán, además: en El Hierro: el 3 de brumario"), "día de 2026"},
		{"sin-cita", nil, parrafoDeNota("1. En la Comunidad Autónoma de Canarias rige su decreto."), "comillas"},
		{"cita-sin-cerrar", nil, parrafoDeNota("1. Dispone que: «Serán, además: en El Hierro: el 24 de septiembre."), "comillas"},
		{"nadie-la-llama", []string{"Aragón", "Canarias"}, notaDeCanarias, "ninguna comunidad"},
		{"la-llaman-dos", []string{"Aragón (1)", "Canarias (1)"}, notaDeCanarias, "2 comunidades"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			s := basica()
			s.tras = leyendaGrabada + parrafoDeLasNotas + caso.notas

			if caso.cabecera != nil {
				s.cabecera = caso.cabecera
			}

			resolucion, err := LeerResolucion(s.documento())
			require.NoError(t, err)

			require.Len(t, resolucion.NotasSinLeer, 1)
			assert.Equal(t, 1, resolucion.NotasSinLeer[0].Numero)
			assert.Contains(t, resolucion.NotasSinLeer[0].Motivo, caso.motivo)
			assert.Empty(t, resolucion.Partes)
			assert.Equal(t, leida.Filas, resolucion.Filas, "la nota que queda fuera no cambia la tabla")
		})
	}
}

// probarCitaFuera fija que una cita que no se lee queda fuera sola: las otras
// citas de su nota se leen, porque cada una es una disposición distinta.
func probarCitaFuera(t *testing.T) {
	t.Parallel()

	s := basica()
	s.tras = leyendaGrabada + parrafoDeLasNotas + notaDeCanarias + parrafoDeNota("Asimismo, dispone que: «De las trece "+
		"fiestas habrá una, a escoger entre el 6 de enero y el 26 de diciembre, que tendrá el carácter de recuperable».")

	resolucion, err := LeerResolucion(s.documento())
	require.NoError(t, err)

	require.Len(t, resolucion.NotasSinLeer, 1)
	assert.Equal(t, 1, resolucion.NotasSinLeer[0].Numero)
	assert.Contains(t, resolucion.NotasSinLeer[0].Motivo, "«De las trece fiestas")
	assert.Equal(t, []FiestaDeUnaParte{
		{Fecha: "2026-09-24", Parte: "El Hierro", Columna: 1, Nota: 1},
		{Fecha: "2026-02-02", Parte: "Tenerife", Columna: 1, Nota: 1},
	}, resolucion.Partes, "la relación de la misma nota se lee")
}

// probarSustitucion lee una sustitución en una parte con la forma de la nota de
// Arán de las tres resoluciones grabadas: la fiesta sustituida deja de ser de
// toda la comunidad que llama a la nota y queda como fiesta del resto de ella,
// y la que la sustituye, como fiesta de la parte (FR-011, FR-025).
func probarSustitucion(t *testing.T) {
	t.Parallel()

	s := basica()
	s.cabecera = []string{"Aragón", "Cataluña (1)"}
	s.filas = [][]string{{"Diciembre", "", ""}, {"26 San Esteban.", "***", "***"}}
	s.tras = leyendaGrabada + parrafoDeLasNotas + parrafoDeNota("1. En la Comunidad Autónoma de Cataluña, la Orden EMT/66/2025 "+
		"dispone que: «En el territorio de Arán, la fiesta del día 26 de diciembre (Sant Esteve) queda sustituida por "+
		"la de 17 de junio (Fiesta de Arán), miércoles».")

	resolucion, err := LeerResolucion(s.documento())
	require.NoError(t, err)

	autonomico := plazos.AmbitoAutonomico

	assert.Equal(t, Resolucion{
		Anio:     2026,
		Cabecera: []Columna{{Nombre: "Aragón"}, {Nombre: "Cataluña", Notas: []int{1}}},
		Filas:    []Fila{{Fecha: "2026-12-26", Texto: "26 San Esteban.", Ambitos: []plazos.Ambito{autonomico, ""}}},
		Partes: []FiestaDeUnaParte{
			{Fecha: "2026-12-26", Parte: "la comunidad salvo el territorio de Arán", Columna: 1, Nota: 1},
			{Fecha: "2026-06-17", Parte: "el territorio de Arán", Columna: 1, Nota: 1},
		},
	}, resolucion)
}
