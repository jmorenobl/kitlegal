package boe

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Lo que comparten los tres tests de la consulta guardada: la base de la API, las
// normas con las que se construyen las claves y los mensajes de sus fallos, que
// se fijan por su texto (contrato errores-y-codigos §2, fila 20).
const (
	apiDeLasEntradas            = "https://www.boe.es/datosabiertos/api/legislacion-consolidada"
	normaDeLasEntradas          = "BOE-A-2015-10565"
	otraNormaDeLasEntradas      = "BOE-A-2015-10566"
	motivoDeLaEntradaQueNoSeLee = "la entrada de la caché con la clave %q no se puede leer"
	motivoDeLaEntradaSinGuardar = "la entrada de la caché con la clave %q no se puede componer"
)

// TestClaveDeEntrada fija la clave de cada consulta guardada de data-model.md
// §6, boe.legislacion-consolidada|1|<verbo>|<dirección de la API del recurso>:
// dos consultas distintas nunca comparten clave, dos textos de búsqueda que
// construyen la misma consulta sí, y los metadatos de una norma tienen una sola,
// la misma para metadatos, articulo y articulos (FR-090, research.md D5).
func TestClaveDeEntrada(t *testing.T) {
	t.Parallel()

	t.Run("la forma de cada clave", func(t *testing.T) {
		t.Parallel()

		direccionDeLaBusqueda, err := direccionDeBusqueda("procedimiento administrativo común")
		require.NoError(t, err)

		formas := []struct {
			nombre   string
			clave    string
			esperada string
		}{
			{
				nombre:   "buscar, con la dirección de la búsqueda",
				clave:    claveDeLaBusqueda(direccionDeLaBusqueda).String(),
				esperada: "boe.legislacion-consolidada|1|buscar|" + direccionDeLaBusqueda,
			},
			{
				nombre:   "indice, con la dirección del índice",
				clave:    claveDelIndice(normaDeLasEntradas).String(),
				esperada: "boe.legislacion-consolidada|1|indice|" + apiDeLasEntradas + "/id/BOE-A-2015-10565/texto/indice",
			},
			{
				nombre:   "metadatos, con la dirección de los metadatos",
				clave:    claveDeLosMetadatos(normaDeLasEntradas).String(),
				esperada: "boe.legislacion-consolidada|1|metadatos|" + apiDeLasEntradas + "/id/BOE-A-2015-10565/metadatos",
			},
			{
				nombre:   "articulo, con la dirección del bloque",
				clave:    claveDelArticulo(normaDeLasEntradas, "a21").String(),
				esperada: "boe.legislacion-consolidada|1|articulo|" + apiDeLasEntradas + "/id/BOE-A-2015-10565/texto/bloque/a21",
			},
			{
				nombre:   "analisis, con la dirección del análisis",
				clave:    claveDelAnalisis(normaDeLasEntradas).String(),
				esperada: "boe.legislacion-consolidada|1|analisis|" + apiDeLasEntradas + "/id/BOE-A-2015-10565/analisis",
			},
		}

		for _, forma := range formas {
			t.Run(forma.nombre, func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, forma.esperada, forma.clave)
			})
		}
	})

	t.Run("dos consultas distintas nunca comparten clave", func(t *testing.T) {
		t.Parallel()

		consultas := clavesDeConsultasDistintas(t)

		consultaDeCadaClave := make(map[string]string, len(consultas))
		for _, consulta := range consultas {
			if otra, repetida := consultaDeCadaClave[consulta.clave]; repetida {
				t.Errorf("%s y %s comparten la clave %q", otra, consulta.nombre, consulta.clave)
			}
			consultaDeCadaClave[consulta.clave] = consulta.nombre

			// La clave se parte en sus cuatro segmentos de una sola forma: el
			// separador no aparece en ninguna dirección, tampoco en la de una
			// búsqueda cuyo texto lo lleva, y articulos no tiene entrada propia.
			segmentos := strings.Split(consulta.clave, "|")
			if assert.Len(t, segmentos, 4, "los segmentos de la clave de %s", consulta.nombre) {
				assert.Equal(t, NombreDeLaFuente, segmentos[0], "la fuente de la clave de %s", consulta.nombre)
				assert.Equal(t, "1", segmentos[1], "la versión de la clave de %s", consulta.nombre)
				assert.Contains(t, []string{"buscar", "indice", "metadatos", "articulo", "analisis"}, segmentos[2],
					"el verbo de la clave de %s", consulta.nombre)
				assert.True(t, strings.HasPrefix(segmentos[3], apiDeLasEntradas),
					"la dirección de la clave de %s es de la API: %q", consulta.nombre, segmentos[3])
			}
		}
	})

	t.Run("dos textos que construyen la misma consulta comparten clave", func(t *testing.T) {
		t.Parallel()

		conPalabras, err := direccionDeBusqueda("ley  general")
		require.NoError(t, err)
		conOperadores, err := direccionDeBusqueda("titulo:ley AND titulo:general")
		require.NoError(t, err)

		assert.Equal(t, claveDeLaBusqueda(conOperadores).String(), claveDeLaBusqueda(conPalabras).String())
	})

	t.Run("los metadatos de una norma tienen una sola clave para metadatos, articulo y articulos", func(t *testing.T) {
		t.Parallel()

		const esperada = "boe.legislacion-consolidada|1|metadatos|" + apiDeLasEntradas + "/id/BOE-A-2015-10565/metadatos"

		// claveDeLosMetadatos solo recibe la norma: ni el verbo que la lee o la
		// escribe ni el bloque que lo lleva a ella cambian la clave, así que
		// metadatos, articulo y articulos usan la misma entrada, y la de cada
		// bloque es otra.
		consultas := []struct {
			verbo     string
			metadatos string
			bloques   []string
		}{
			{verbo: "metadatos", metadatos: claveDeLosMetadatos(ConsultaMetadatos{Norma: normaDeLasEntradas}.Norma).String()},
			{
				verbo:     "articulo",
				metadatos: claveDeLosMetadatos(ConsultaArticulo{Norma: normaDeLasEntradas, Bloque: "a21"}.Norma).String(),
				bloques:   []string{claveDelArticulo(normaDeLasEntradas, "a21").String()},
			},
			{
				verbo: "articulos",
				metadatos: claveDeLosMetadatos(
					ConsultaArticulos{Norma: normaDeLasEntradas, Bloques: []string{"a21", "da3"}}.Norma).String(),
				bloques: []string{
					claveDelArticulo(normaDeLasEntradas, "a21").String(),
					claveDelArticulo(normaDeLasEntradas, "da3").String(),
				},
			},
		}

		for _, consulta := range consultas {
			t.Run(consulta.verbo, func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, esperada, consulta.metadatos)
				assert.NotContains(t, consulta.bloques, esperada)
			})
		}
	})
}

// claveDeUnaConsulta es la clave de una consulta con el nombre con el que la
// nombra el fallo.
type claveDeUnaConsulta struct {
	nombre string
	clave  string
}

// clavesDeConsultasDistintas son las claves de consultas que piden recursos
// distintos, de los cinco verbos que guardan entradas: textos de búsqueda que
// construyen consultas distintas —con palabras, con otras mayúsculas, con
// operadores y con el separador de la clave—, el índice, los metadatos y el
// análisis de tres normas, y bloques que solo se distinguen en un carácter o en
// la norma.
func clavesDeConsultasDistintas(t *testing.T) []claveDeUnaConsulta {
	t.Helper()

	consultas := []claveDeUnaConsulta{}

	for _, texto := range []string{
		"ley general tributaria", "ley general", "ley", "Ley", `"ley general"`, "ley OR general", "ley|general",
	} {
		direccion, err := direccionDeBusqueda(texto)
		require.NoError(t, err, "la dirección de la búsqueda %q", texto)

		consultas = append(consultas,
			claveDeUnaConsulta{nombre: "buscar " + texto, clave: claveDeLaBusqueda(direccion).String()})
	}

	for _, norma := range []string{normaDeLasEntradas, otraNormaDeLasEntradas, "BOE-A-1978-31229", "BOE-A-2015-1"} {
		consultas = append(consultas,
			claveDeUnaConsulta{nombre: "indice " + norma, clave: claveDelIndice(norma).String()},
			claveDeUnaConsulta{nombre: "metadatos " + norma, clave: claveDeLosMetadatos(norma).String()},
			claveDeUnaConsulta{nombre: "analisis " + norma, clave: claveDelAnalisis(norma).String()},
		)
	}

	for _, pedido := range []struct{ norma, bloque string }{
		{norma: normaDeLasEntradas, bloque: "a21"},
		{norma: normaDeLasEntradas, bloque: "a2"},
		{norma: normaDeLasEntradas, bloque: "A21"},
		{norma: normaDeLasEntradas, bloque: "da3"},
		{norma: normaDeLasEntradas, bloque: "a1-30"},
		{norma: normaDeLasEntradas, bloque: "a85bis."},
		{norma: otraNormaDeLasEntradas, bloque: "a21"},
		{norma: "BOE-A-2015-1", bloque: "a21"},
	} {
		consultas = append(consultas, claveDeUnaConsulta{
			nombre: "articulo " + pedido.norma + " " + pedido.bloque,
			clave:  claveDelArticulo(pedido.norma, pedido.bloque).String(),
		})
	}

	return consultas
}

// TestEntradaIdaYVuelta fija que lo que se guarda vuelve igual: la fecha de
// consulta en RFC 3339 con nanosegundos y desplazamiento, la url, que es la
// dirección de la clave, y los datos, de modo que el sobre que se monta con lo
// leído lleva la misma url, la misma fecha_consulta y el mismo data, byte a byte,
// que el de la consulta que lo guardó (FR-096, invariante 3 de data-model.md §9);
// y que lo que no podría volver así no se compone (research.md D5).
func TestEntradaIdaYVuelta(t *testing.T) {
	t.Parallel()

	madrid := time.FixedZone("CEST", 2*60*60)
	terranova := time.FixedZone("NDT", -(2*60*60 + 30*60))

	t.Run("cada fecha, con sus nanosegundos y su desplazamiento", func(t *testing.T) {
		t.Parallel()

		fechas := []struct {
			nombre string
			fecha  time.Time
			texto  string
		}{
			{
				nombre: "UTC con nanosegundos",
				fecha:  time.Date(2026, time.September, 13, 10, 30, 0, 123456789, time.UTC),
				texto:  "2026-09-13T10:30:00.123456789Z",
			},
			{
				nombre: "sin los ceros finales de la fracción",
				fecha:  time.Date(2026, time.September, 13, 10, 30, 0, 120000000, time.UTC),
				texto:  "2026-09-13T10:30:00.12Z",
			},
			{
				nombre: "sin fracción de segundo",
				fecha:  time.Date(2026, time.September, 13, 10, 30, 0, 0, time.UTC),
				texto:  "2026-09-13T10:30:00Z",
			},
			{
				nombre: "desplazamiento positivo y un nanosegundo",
				fecha:  time.Date(2026, time.September, 13, 12, 30, 0, 1, madrid),
				texto:  "2026-09-13T12:30:00.000000001+02:00",
			},
			{
				nombre: "desplazamiento negativo con minutos",
				fecha:  time.Date(2026, time.September, 13, 8, 0, 0, 987654321, terranova),
				texto:  "2026-09-13T08:00:00.987654321-02:30",
			},
		}

		for _, caso := range fechas {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				compruebaIdaYVuelta(t, claveDeLosMetadatos(normaDeLasEntradas), caso.fecha, caso.texto, metadatosGuardados())
			})
		}
	})

	t.Run("cada clase de entrada, con sus datos", func(t *testing.T) {
		t.Parallel()

		fecha := time.Date(2026, time.September, 13, 12, 30, 0, 123456789, madrid)
		const texto = "2026-09-13T12:30:00.123456789+02:00"

		sinResultados, err := direccionDeBusqueda("ley inexistente")
		require.NoError(t, err)
		conResultados, err := direccionDeBusqueda("procedimiento administrativo común")
		require.NoError(t, err)

		clases := []struct {
			nombre    string
			comprobar func(t *testing.T)
		}{
			{nombre: "buscar sin resultados, con la lista vacía", comprobar: func(t *testing.T) {
				t.Helper()
				compruebaIdaYVuelta(t, claveDeLaBusqueda(sinResultados), fecha, texto, []ResultadoDeBusqueda{})
			}},
			{nombre: "buscar con resultados y una dirección con &", comprobar: func(t *testing.T) {
				t.Helper()
				compruebaIdaYVuelta(t, claveDeLaBusqueda(conResultados), fecha, texto, []ResultadoDeBusqueda{
					{
						Identificador: normaDeLasEntradas, Titulo: "Ley 39/2015, de 1 de octubre", Rango: "Ley",
						VigenciaAgotada: "N", EstadoConsolidacion: "Finalizado",
						URL: direccionPublicaDeLaNorma(normaDeLasEntradas),
					},
					{Identificador: "", URL: direccionPublicaDeLaNorma("")},
				})
			}},
			{nombre: "indice", comprobar: func(t *testing.T) {
				t.Helper()
				compruebaIdaYVuelta(t, claveDelIndice(normaDeLasEntradas), fecha, texto, Indice{
					Norma: normaDeLasEntradas,
					URL:   direccionPublicaDeLaNorma(normaDeLasEntradas),
					Bloques: []EntradaDeIndice{
						{ID: "a21", Titulo: "Artículo 21. Obligación de resolver", Tipo: tipoArticulo},
						{ID: "da3", Titulo: "Disposición adicional tercera", Tipo: tipoDisposicionAdicional},
					},
				})
			}},
			{nombre: "metadatos, con sus avisos", comprobar: func(t *testing.T) {
				t.Helper()
				compruebaIdaYVuelta(t, claveDeLosMetadatos(normaDeLasEntradas), fecha, texto, metadatosGuardados())
			}},
			{nombre: "articulo, con un texto que JSON escapa y sin avisos", comprobar: func(t *testing.T) {
				t.Helper()

				const textoDelBloque = "Artículo 21.\n1. <La norma> & \"su texto\" \\   € 𝄞.\n2. Segundo."

				compruebaIdaYVuelta(t, claveDelArticulo(normaDeLasEntradas, "a21"), fecha, texto, Articulo{
					Norma: normaDeLasEntradas, Bloque: "a21", Titulo: "Artículo 21", Tipo: tipoArticulo,
					FechaVersion: "20151002", Texto: textoDelBloque, HashTexto: huellaDelTexto(textoDelBloque),
					Avisos: []Aviso{}, URL: direccionPublicaDelBloque(normaDeLasEntradas, "a21"),
				})
			}},
			{nombre: "analisis, con una lista vacía", comprobar: func(t *testing.T) {
				t.Helper()
				compruebaIdaYVuelta(t, claveDelAnalisis(normaDeLasEntradas), fecha, texto, Analisis{
					Norma:    normaDeLasEntradas,
					Materias: []Materia{{Codigo: "1234", Texto: "Procedimiento"}, {Texto: "suelta"}},
					Notas:    []string{"Entrada en vigor"},
					Referencias: Referencias{
						Anteriores:  []ReferenciaAnterior{{Relacion: "DEROGA", Norma: "BOE-A-1992-26318", Texto: "la Ley 30/1992"}},
						Posteriores: []ReferenciaPosterior{},
					},
				})
			}},
		}

		for _, clase := range clases {
			t.Run(clase.nombre, func(t *testing.T) {
				t.Parallel()

				clase.comprobar(t)
			})
		}
	})

	t.Run("lo que no podría volver igual no se compone", func(t *testing.T) {
		t.Parallel()

		clave := claveDeLosMetadatos(normaDeLasEntradas)
		mensaje := fmt.Sprintf(motivoDeLaEntradaSinGuardar, clave.String()) +
			" (" + apiDeLasEntradas + "/id/BOE-A-2015-10565/metadatos)"

		fechas := []struct {
			nombre string
			fecha  time.Time
		}{
			{nombre: "sin fecha de consulta", fecha: time.Time{}},
			{nombre: "una fecha que RFC 3339 no puede escribir", fecha: time.Date(10000, time.January, 1, 0, 0, 0, 0, time.UTC)},
		}

		for _, caso := range fechas {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				contenido, err := contenidoDeEntrada(clave, caso.fecha, metadatosGuardados())

				assert.Nil(t, contenido)
				compruebaFalloDeEntrada(t, err, apiDeLasEntradas+"/id/BOE-A-2015-10565/metadatos", mensaje)
			})
		}
	})
}

// compruebaIdaYVuelta guarda los datos bajo la clave con la fecha, exige la forma
// de lo guardado —exactamente las tres claves, la fecha como su texto RFC 3339 y
// la url, la dirección de la clave— y exige que lo leído monte el mismo sobre,
// byte a byte, que la consulta que lo guardó.
func compruebaIdaYVuelta[T datosDeEntrada](t *testing.T, clave claveDeEntrada[T], fecha time.Time, texto string, datos T) {
	t.Helper()

	contenido, err := contenidoDeEntrada(clave, fecha, datos)
	require.NoError(t, err)

	var guardado map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(contenido, &guardado))
	assert.ElementsMatch(t, []string{"fecha_consulta", "url", "datos"}, slices.Collect(maps.Keys(guardado)))

	var fechaGuardada, urlGuardada string
	require.NoError(t, json.Unmarshal(guardado["fecha_consulta"], &fechaGuardada))
	require.NoError(t, json.Unmarshal(guardado["url"], &urlGuardada))
	assert.Equal(t, texto, fechaGuardada)
	assert.Equal(t, clave.direccion, urlGuardada)

	leida, err := leerEntrada(clave, contenido)
	require.NoError(t, err)

	assert.Equal(t, texto, leida.FechaConsulta.Format(time.RFC3339Nano))
	assert.True(t, fecha.Equal(leida.FechaConsulta), "%s no es el instante %s", leida.FechaConsulta, fecha)
	assert.Equal(t, clave.direccion, leida.URL)
	assert.Equal(t, datos, leida.Datos)

	deLaConsulta, err := json.Marshal(schema.Sobre{
		Ok: true, Fuente: NombreDeLaFuente, URL: clave.direccion, FechaConsulta: fecha, Data: datos,
	})
	require.NoError(t, err)
	deLoGuardado, err := json.Marshal(schema.Sobre{
		Ok: true, Fuente: NombreDeLaFuente, URL: leida.URL, FechaConsulta: leida.FechaConsulta, Data: leida.Datos,
	})
	require.NoError(t, err)
	assert.Equal(t, string(deLaConsulta), string(deLoGuardado))
}

// metadatosGuardados son unos metadatos con todos sus campos y los tres avisos.
func metadatosGuardados() Metadatos {
	return Metadatos{
		Norma:               normaDeLasEntradas,
		Titulo:              "Ley 39/2015, de 1 de octubre",
		Rango:               "Ley",
		NumeroOficial:       "39/2015",
		FechaDisposicion:    "20151001",
		FechaPublicacion:    "20151002",
		FechaVigencia:       "20161002",
		EstatusDerogacion:   "S",
		VigenciaAgotada:     "S",
		EstadoConsolidacion: EstadoDeConsolidacion{Codigo: "4", Texto: "En proceso"},
		URLELI:              "https://www.boe.es/eli/es/l/2015/10/01/39",
		Avisos: avisosDe(map[string]any{
			"estado_consolidacion": map[string]any{"codigo": "4"},
			"estatus_derogacion":   "S",
			"vigencia_agotada":     "S",
		}),
	}
}

// TestEntradaIlegibleEsInesperado fija que una entrada con la clave correcta que
// no tiene la forma de la consulta guardada no se sirve ni se da por ausente: es
// un fallo «inesperado», código 1, con la dirección del recurso de la clave, sin
// instante y con un mensaje que nombra la clave y no el detalle técnico (contrato
// errores-y-codigos, fila 20; research.md D5). La forma es la que escribe
// contenidoDeEntrada: las tres claves, ninguna nula ni de más —tampoco dentro de
// los datos—, la fecha de consulta en RFC 3339 con desplazamiento y distinta del
// instante cero, y la url igual a la dirección de la clave.
func TestEntradaIlegibleEsInesperado(t *testing.T) {
	t.Parallel()

	const (
		direccion = apiDeLasEntradas + "/id/BOE-A-2015-10565/metadatos"
		fecha     = `"fecha_consulta":"2026-09-13T12:30:00.123456789+02:00"`
		url       = `"url":"` + direccion + `"`
		datos     = `"datos":{"norma":"BOE-A-2015-10565","avisos":[{"codigo":"derogada","texto":"derogada"}]}`
	)

	clave := claveDeLosMetadatos(normaDeLasEntradas)
	mensaje := fmt.Sprintf(motivoDeLaEntradaQueNoSeLee, clave.String()) + " (" + direccion + ")"

	t.Run("la entrada de la que parten los casos se lee", func(t *testing.T) {
		t.Parallel()

		leida, err := leerEntrada(clave, []byte("{"+fecha+","+url+","+datos+"}"))
		require.NoError(t, err)
		assert.Equal(t, direccion, leida.URL)
		assert.Equal(t, normaDeLasEntradas, leida.Datos.Norma)
	})

	casos := []struct {
		nombre    string
		contenido string
	}{
		{nombre: "vacía", contenido: ""},
		{nombre: "no es JSON", contenido: "{" + fecha + ","},
		{nombre: "la raíz es una lista", contenido: "[{" + fecha + "," + url + "," + datos + "}]"},
		{nombre: "la raíz es nula", contenido: "null"},
		{nombre: "hay otro valor detrás", contenido: "{" + fecha + "," + url + "," + datos + "} {}"},
		{nombre: "hay basura detrás", contenido: "{" + fecha + "," + url + "," + datos + "}x"},
		{nombre: "no es UTF-8", contenido: "{" + fecha + "," + url + `,"datos":{"norma":"BOE-A-2015-10565` + "\xff" + `"}}`},
		{nombre: "una clave de más", contenido: "{" + fecha + "," + url + "," + datos + `,"hash":"sha256:0"}`},
		{nombre: "una clave de más en los datos", contenido: "{" + fecha + "," + url + `,"datos":{"norma":"","extra":""}}`},
		{
			nombre:    "una clave de más dentro de un aviso",
			contenido: "{" + fecha + "," + url + `,"datos":{"avisos":[{"codigo":"derogada","nivel":"alto"}]}}`,
		},
		{nombre: "sin fecha_consulta", contenido: "{" + url + "," + datos + "}"},
		{nombre: "sin url", contenido: "{" + fecha + "," + datos + "}"},
		{nombre: "sin datos", contenido: "{" + fecha + "," + url + "}"},
		{nombre: "fecha_consulta nula", contenido: `{"fecha_consulta":null,` + url + "," + datos + "}"},
		{nombre: "url nula", contenido: "{" + fecha + `,"url":null,` + datos + "}"},
		{nombre: "datos nulos", contenido: "{" + fecha + "," + url + `,"datos":null}`},
		{nombre: "fecha sin desplazamiento", contenido: `{"fecha_consulta":"2026-09-13T12:30:00.123456789",` + url + "," + datos + "}"},
		{nombre: "fecha que no es RFC 3339", contenido: `{"fecha_consulta":"13/09/2026 12:30",` + url + "," + datos + "}"},
		{nombre: "fecha que no es texto", contenido: `{"fecha_consulta":1789000000,` + url + "," + datos + "}"},
		{nombre: "fecha en el instante cero", contenido: `{"fecha_consulta":"0001-01-01T00:00:00Z",` + url + "," + datos + "}"},
		{
			nombre:    "url de los metadatos de otra norma",
			contenido: "{" + fecha + `,"url":"` + apiDeLasEntradas + `/id/BOE-A-2015-10566/metadatos",` + datos + "}",
		},
		{nombre: "url que no es texto", contenido: "{" + fecha + `,"url":3,` + datos + "}"},
		{nombre: "datos de otra forma", contenido: "{" + fecha + "," + url + `,"datos":[]}`},
		{nombre: "un campo de los datos de otro tipo", contenido: "{" + fecha + "," + url + `,"datos":{"titulo":3}}`},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			_, err := leerEntrada(clave, []byte(caso.contenido))

			compruebaFalloDeEntrada(t, err, direccion, mensaje)
		})
	}
}

// compruebaFalloDeEntrada exige que el error sea el *Error «inesperado» de una
// entrada: código de salida 1, la dirección del recurso de la clave, ningún
// instante —la fecha del sobre la pone el montaje— y exactamente el mensaje.
func compruebaFalloDeEntrada(t *testing.T, err error, direccion, mensaje string) {
	t.Helper()

	var fallo *Error
	require.ErrorAs(t, err, &fallo)
	assert.Equal(t, schema.ClaseInesperado, fallo.Clase())
	assert.Equal(t, schema.ClaseInesperado, cli.Clasificar(err))
	assert.Equal(t, 1, cli.CodigoSalida(err))
	assert.Equal(t, direccion, fallo.URL)
	assert.True(t, fallo.Instante.IsZero(), "el instante del fallo de una entrada: %s", fallo.Instante)
	assert.Equal(t, mensaje, err.Error())
}
