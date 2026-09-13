package boe

import (
	"encoding/json"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// Lo que comparten las pruebas de analisis. Como en metadatos_test.go, las
// direcciones y las claves van escritas enteras, y no con las funciones de
// direcciones.go y entradas.go, para que un cambio en ellas no pase por aquí en
// silencio.
const (
	// analisisVigente y analisisInexistente son las direcciones del análisis de
	// la Ley 39/2015 y del de la norma que la fuente responde con 404 (S2).
	analisisVigente     = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/analisis"
	analisisInexistente = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2099-99999/analisis"
	// vigenciaDelAnalisis es la de su entrada, siete días (FR-091).
	vigenciaDelAnalisis = 604_800 * time.Second
	// anterioresDelAnalisisGrabado y posterioresDelAnalisisGrabado son las
	// referencias que trae el análisis grabado de la Ley 39/2015.
	anterioresDelAnalisisGrabado  = 11
	posterioresDelAnalisisGrabado = 18
	// recorteDeBoePy es el número de caracteres al que refs/boe.py 564 recorta el
	// texto de cada referencia anterior, y que el porte no recorta (FR-060).
	recorteDeBoePy = 200
	// textoLargoDeUnaReferencia es un texto de referencia de prueba, de más de 200
	// caracteres y con letras que en UTF-8 ocupan más de un octeto, compuesto con
	// tramos de las referencias del análisis grabado de la Ley 39/2015.
	textoLargoDeUnaReferencia = "los arts. 4 a 7 de la Ley 2/2011, de 4 de marzo; en la forma indicada, la Ley 30/1992, " +
		"de 26 de noviembre; los arts. 64, 69, 70, 72, 73, 85, 103 y 117 de la Ley 36/2011, de 10 de octubre; y la " +
		"disposición final 7, con las previsiones indicadas, de la Ley Orgánica 2/2012, de 27 de abril"
)

// TestAnalisis fija el verbo analisis (contrato verbos-y-salidas §6;
// data-model.md §2.6, §3.1 y §7.3; FR-060, FR-061, FR-070, FR-090 a FR-094,
// FR-096 y FR-101) sobre la fuente compuesta con el cliente real de httpx en
// reproducción de las grabaciones, y con la caché real en una carpeta temporal,
// las dos gobernadas por el mismo reloj de prueba:
//
//   - la norma se valida antes de abrir la caché o construir el cliente;
//   - el análisis se pide en JSON y se lee del primer elemento de data, o de data
//     si llega suelto: sus materias, sus notas y sus referencias anteriores y
//     posteriores, abiertas de sus envoltorios y en el orden de la fuente, con el
//     texto de cada referencia anterior completo, sin el recorte a 200
//     caracteres de refs/boe.py 564 (FR-060);
//   - un objeto suelto donde la fuente suele entregar una lista se lee como una
//     lista de un elemento, igual que si hubiera venido como lista (US5,
//     escenario 2; FR-070);
//   - lo leído se escribe con el instante de la petición, se sirve de su entrada
//     con su url y su fecha de consulta sin pedir nada, también con --offline, y
//     a los siete días se vuelve a pedir;
//   - sin entrada, con --offline, «fuente no disponible» sin pedir nada, y con
//     --dry-run, la línea de la petición que se habría emitido, sin crear la
//     caché ni dejar nada escrito;
//   - y los fallos —el 404 y data vacío, que son «no encontrado» (FR-061; US5,
//     escenario 3), cualquier otro estado y la respuesta que no se interpreta—
//     llevan su clase, la dirección del análisis y el instante de la petición, y
//     no se escriben.
//
// Todo resultado, de éxito o de fallo, lleva la fuente del BOE y una dirección
// de www.boe.es (FR-002, FR-101).
func TestAnalisis(t *testing.T) {
	t.Parallel()

	t.Run("lpac", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
		consulta := ConsultaAnalisis{Norma: normaVigente}
		esperado := resultadoResuelto(analisisVigente, banco.reloj.ahora(), analisisGrabadoDeLaNormaVigente(t))

		resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

		compruebaResuelta(t, resultado, err, esperado)
		compruebaAnalisisDeLaNormaVigente(t, resultado.Datos)

		// La entrada se sirve, con la url y la fecha de su petición, hasta el
		// último instante de sus siete días, sin pedir nada ni construir el
		// cliente aunque se pudiera pedir.
		banco.reloj.adelanta(vigenciaDelAnalisis - time.Nanosecond)

		servido, err := banco.resuelve(t, schema.Contexto{}, consulta)

		compruebaResuelta(t, servido, err, esperado)

		// A los siete días se vuelve a pedir, y la fecha es la de esa petición.
		banco.reloj.adelanta(time.Nanosecond)
		esperado.Procedencia.FechaConsulta = banco.reloj.ahora()

		pedidoOtraVez, err := banco.resuelve(t, schema.Contexto{}, consulta)

		compruebaResuelta(t, pedidoOtraVez, err, esperado)
		banco.compruebaPedidas(t, analisisVigente, analisisVigente)
		assert.Equal(t, 2, banco.construcciones, "lo servido de la caché no construye ningún cliente")
	})

	t.Run("texto-completo", func(t *testing.T) {
		t.Parallel()

		require.Greater(t, utf8.RuneCountInString(textoLargoDeUnaReferencia), recorteDeBoePy,
			"el texto de la prueba tiene que ser más largo que el recorte de refs/boe.py")

		textoEnJSON, err := json.Marshal(textoLargoDeUnaReferencia)
		require.NoError(t, err)

		banco := nuevoBanco(t, respondeConElCuerpo(`{"data": [{"referencias": {"anteriores": [{"anterior": [`+
			`{"id_norma": "BOE-A-2011-4117", "relacion": {"codigo": "210", "texto": "DEROGA"}, "texto": `+
			string(textoEnJSON)+`}]}]}}]}`))
		consulta := ConsultaAnalisis{Norma: normaVigente}
		esperado := resultadoResuelto(analisisVigente, banco.reloj.ahora(), Analisis{
			Norma:    normaVigente,
			Materias: []Materia{},
			Notas:    []string{},
			Referencias: Referencias{
				Anteriores: []ReferenciaAnterior{
					{Relacion: "DEROGA", Norma: "BOE-A-2011-4117", Texto: textoLargoDeUnaReferencia},
				},
				Posteriores: []ReferenciaPosterior{},
			},
		})

		resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

		compruebaResuelta(t, resultado, err, esperado)

		banco.reloj.adelanta(time.Second)

		// El texto completo es también el que se guarda y se vuelve a servir.
		servido, err := banco.resuelve(t, schema.Contexto{Offline: true}, consulta)

		compruebaResuelta(t, servido, err, esperado)
		banco.compruebaPedidas(t, analisisVigente)
	})

	// unaDeCada es el análisis con una materia, una nota y una referencia de cada
	// lista, que la fuente puede entregar en listas o en objetos sueltos.
	unaDeCada := Analisis{
		Norma:    normaVigente,
		Materias: []Materia{{Codigo: "6499", Texto: "Seguridad Social"}},
		Notas:    []string{"Entrada en vigor el 2 de octubre de 2016."},
		Referencias: Referencias{
			Anteriores: []ReferenciaAnterior{
				{Relacion: "DEROGA", Norma: "BOE-A-2011-4117", Texto: "los arts. 4 a 7 de la Ley 2/2011, de 4 de marzo"},
			},
			Posteriores: []ReferenciaPosterior{{Relacion: "SE MODIFICA", Norma: "BOE-A-2022-11589"}},
		},
	}

	leidos := []struct {
		nombre   string
		cuerpo   string
		analisis Analisis
	}{
		{
			// La forma de la API: cada lista con su envoltorio, y el contenido del
			// envoltorio en una lista.
			nombre: "en-listas",
			cuerpo: `{"data": [{"materias": [{"materia": [{"codigo": "6499", "texto": "Seguridad Social"}]}], ` +
				`"notas": [{"nota": ["Entrada en vigor el 2 de octubre de 2016."]}], ` +
				`"referencias": {"anteriores": [{"anterior": [{"id_norma": "BOE-A-2011-4117", ` +
				`"relacion": {"codigo": "210", "texto": "DEROGA"}, "texto": "los arts. 4 a 7 de la Ley 2/2011, de 4 de marzo"}]}], ` +
				`"posteriores": [{"posterior": [{"id_norma": "BOE-A-2022-11589", "relacion": {"codigo": "270", "texto": "SE MODIFICA"}}]}]}}]}`,
			analisis: unaDeCada,
		},
		{
			// Lo mismo con data, cada lista y el contenido de cada envoltorio
			// sueltos: listas de un elemento, y no cada clave del objeto como un
			// elemento, como en refs/boe.py 527-530, 549-555 y 569-575 (US5,
			// escenario 2; FR-070).
			nombre: "objetos-sueltos",
			cuerpo: `{"data": {"materias": {"materia": {"codigo": "6499", "texto": "Seguridad Social"}}, ` +
				`"notas": {"nota": "Entrada en vigor el 2 de octubre de 2016."}, ` +
				`"referencias": {"anteriores": {"anterior": {"id_norma": "BOE-A-2011-4117", ` +
				`"relacion": {"codigo": "210", "texto": "DEROGA"}, "texto": "los arts. 4 a 7 de la Ley 2/2011, de 4 de marzo"}}, ` +
				`"posteriores": {"posterior": {"id_norma": "BOE-A-2022-11589", "relacion": {"codigo": "270", "texto": "SE MODIFICA"}}}}}}`,
			analisis: unaDeCada,
		},
		{
			// Sin envoltorios: materias y referencias directas, notas como cadena
			// suelta. Una materia que no es objeto aporta solo su texto, y nulo es
			// la cadena vacía (refs/boe.py 535; J9); una referencia que no es objeto
			// se salta (J8); relacion puede ser una cadena (refs/boe.py 559-560) y,
			// si falta, es la cadena vacía, nunca el marcador ? (FR-016); y de las
			// posteriores no se lee su texto, que refs/boe.py 577-581 no presenta.
			nombre: "sin-envoltorios",
			cuerpo: `{"data": [{"materias": [{"codigo": "6499", "texto": "Seguridad Social"}, "Procedimiento administrativo", null], ` +
				`"notas": "Entrada en vigor el 2 de octubre de 2016.", ` +
				`"referencias": {"anteriores": ["BOE-A-1", {"id_norma": "BOE-A-1992-26318", "relacion": "DEROGA"}], ` +
				`"posteriores": [{"id_norma": "BOE-A-2022-11589", "texto": 77}, 3]}}]}`,
			analisis: Analisis{
				Norma: normaVigente,
				Materias: []Materia{
					{Codigo: "6499", Texto: "Seguridad Social"},
					{Texto: "Procedimiento administrativo"},
					{},
				},
				Notas: []string{"Entrada en vigor el 2 de octubre de 2016."},
				Referencias: Referencias{
					Anteriores:  []ReferenciaAnterior{{Relacion: "DEROGA", Norma: "BOE-A-1992-26318"}},
					Posteriores: []ReferenciaPosterior{{Norma: "BOE-A-2022-11589"}},
				},
			},
		},
		{
			// Listas vacías, referencias que no son un objeto (refs/boe.py 545) o
			// sin ninguna de las tres claves: listas vacías, y no nulas. data no
			// está vacío: no es «no encontrado» (J3; refs/boe.py 518).
			nombre: "vacios",
			cuerpo: `{"data": [{"materias": [], "notas": {}, "referencias": [{"anteriores": [{"id_norma": "BOE-A-2011-4117"}]}]}]}`,
			analisis: Analisis{
				Norma:       normaVigente,
				Materias:    []Materia{},
				Notas:       []string{},
				Referencias: Referencias{Anteriores: []ReferenciaAnterior{}, Posteriores: []ReferenciaPosterior{}},
			},
		},
	}

	for _, caso := range leidos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, respondeConElCuerpo(caso.cuerpo))
			consulta := ConsultaAnalisis{Norma: normaVigente}
			esperado := resultadoResuelto(analisisVigente, banco.reloj.ahora(), caso.analisis)

			resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

			compruebaResuelta(t, resultado, err, esperado)
			compruebaListasVaciasEnJSON(t, resultado.Datos)

			banco.reloj.adelanta(time.Second)

			servido, err := banco.resuelve(t, schema.Contexto{Offline: true}, consulta)

			compruebaResuelta(t, servido, err, esperado)
			banco.compruebaPedidas(t, analisisVigente)
		})
	}

	sinAnalisis := []struct {
		nombre  string
		pedidor func(*testing.T, *relojDePrueba) *pedidorDePrueba
		norma   string
		// fallo es el esperado salvo el instante, que es el de la petición.
		fallo falloDeLaConsulta
	}{
		{
			nombre:  "inexistente",
			pedidor: reproduce(carpetaDeLasGrabaciones),
			norma:   normaInexistente,
			fallo: falloDeLaConsulta{
				direccion: analisisInexistente,
				clase:     schema.ClaseNoEncontrado,
				codigo:    3,
				mensaje:   "la norma BOE-A-2099-99999 no tiene análisis (" + analisisInexistente + ")",
			},
		},
		{
			nombre:  "estado-403",
			pedidor: respondeConElEstado(403),
			norma:   normaVigente,
			fallo: falloDeLaConsulta{
				direccion: analisisVigente,
				clase:     schema.ClaseFuenteNoDisponible,
				codigo:    4,
				mensaje: "la fuente ha respondido con el estado 403 a la petición del análisis de la norma BOE-A-2015-10565 (" +
					analisisVigente + ")",
			},
		},
		{
			nombre:  "ilegible",
			pedidor: respondeConElCuerpo(`{"data": [{"materias": [`),
			norma:   normaVigente,
			fallo:   falloAlInterpretarElAnalisis("el cuerpo no es JSON legible"),
		},
		{
			nombre:  "primer-elemento-que-no-es-objeto",
			pedidor: respondeConElCuerpo(`{"data": [["BOE-A-2015-10565"]]}`),
			norma:   normaVigente,
			fallo:   falloAlInterpretarElAnalisis("el primer elemento de data no es un objeto, sino una lista"),
		},
		{
			nombre:  "codigo-de-materia-que-no-es-texto",
			pedidor: respondeConElCuerpo(`{"data": [{"materias": [{"materia": {"codigo": 6499, "texto": "Seguridad Social"}}]}]}`),
			norma:   normaVigente,
			fallo:   falloAlInterpretarElAnalisis(`el campo "codigo" no es texto, sino un número`),
		},
		{
			nombre:  "materia-que-no-es-texto",
			pedidor: respondeConElCuerpo(`{"data": [{"materias": ["Seguridad Social", true]}]}`),
			norma:   normaVigente,
			fallo:   falloAlInterpretarElAnalisis(`el campo "materia" no es texto, sino un booleano`),
		},
		{
			nombre:  "nota-que-no-es-texto",
			pedidor: respondeConElCuerpo(`{"data": [{"notas": [{"nota": ["Entrada en vigor", {"texto": "Efectos"}]}]}]}`),
			norma:   normaVigente,
			fallo:   falloAlInterpretarElAnalisis(`el campo "nota" no es texto, sino un objeto`),
		},
		{
			nombre: "relacion-que-no-es-texto",
			pedidor: respondeConElCuerpo(`{"data": [{"referencias": {"anteriores": [{"anterior": [` +
				`{"id_norma": "BOE-A-2011-4117", "relacion": {"texto": ["DEROGA"]}}]}]}}]}`),
			norma: normaVigente,
			fallo: falloAlInterpretarElAnalisis(`el campo "relacion.texto" no es texto, sino una lista`),
		},
		{
			nombre: "texto-de-la-referencia-que-no-es-texto",
			pedidor: respondeConElCuerpo(`{"data": [{"referencias": {"anteriores": [{"anterior": [` +
				`{"id_norma": "BOE-A-2011-4117", "relacion": "DEROGA", "texto": 200}]}]}}]}`),
			norma: normaVigente,
			fallo: falloAlInterpretarElAnalisis(`el campo "texto" no es texto, sino un número`),
		},
		{
			nombre: "norma-de-la-posterior-que-no-es-texto",
			pedidor: respondeConElCuerpo(`{"data": [{"referencias": {"posteriores": [{"posterior": [` +
				`{"id_norma": "BOE-A-2022-11589"}, {"id_norma": null}, {"id_norma": 2022}]}]}}]}`),
			norma: normaVigente,
			fallo: falloAlInterpretarElAnalisis(`el campo "id_norma" no es texto, sino un número`),
		},
		{nombre: "data-vacio", pedidor: respondeConElCuerpo(`{"status": {"code": "200"}, "data": []}`), norma: normaVigente, fallo: falloSinAnalisis()},
		{nombre: "data-nulo", pedidor: respondeConElCuerpo(`{"data": null}`), norma: normaVigente, fallo: falloSinAnalisis()},
		{nombre: "sin-data", pedidor: respondeConElCuerpo(`{"status": {}}`), norma: normaVigente, fallo: falloSinAnalisis()},
		{nombre: "data-objeto-vacio", pedidor: respondeConElCuerpo(`{"data": {}}`), norma: normaVigente, fallo: falloSinAnalisis()},
		{nombre: "data-cadena-vacia", pedidor: respondeConElCuerpo(`{"data": ""}`), norma: normaVigente, fallo: falloSinAnalisis()},
	}

	for _, caso := range sinAnalisis {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, caso.pedidor)
			esperado := caso.fallo
			esperado.instante = banco.reloj.ahora()

			resultado, err := banco.resuelve(t, schema.Contexto{}, ConsultaAnalisis{Norma: caso.norma})

			compruebaFalloDeLaConsulta(t, resultado, err, esperado)
			compruebaAnalisisSinEscribir(t, banco, caso.norma, esperado.direccion)
		})
	}

	t.Run("ensayo", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
		consulta := ConsultaAnalisis{Norma: normaVigente}

		resultado, err := banco.resuelve(t, schema.Contexto{DryRun: true}, consulta)

		require.NoError(t, err)
		assert.Equal(t, schema.Resultado{
			Procedencia: schema.Procedencia{Fuente: NombreDeLaFuente, URL: analisisVigente},
			Ensayo:      []string{"GET " + analisisVigente},
		}, resultado)
		require.Len(t, banco.pedidor.respuestas, 1)
		assert.True(t, banco.pedidor.respuestas[0].Ensayo, "bajo --dry-run no se emite la petición")
		banco.compruebaCacheSinCrear(t)

		// Lo que el ensayo describe no queda escrito en ninguna parte: la consulta
		// siguiente lo pide de verdad, con la fecha de su petición.
		pedido, err := banco.resuelve(t, schema.Contexto{}, consulta)

		compruebaResuelta(t, pedido, err,
			resultadoResuelto(analisisVigente, banco.reloj.ahora(), analisisGrabadoDeLaNormaVigente(t)))
		banco.compruebaPedidas(t, analisisVigente, analisisVigente)
		assert.False(t, banco.pedidor.respuestas[1].Ensayo, "sin --dry-run la petición se emite")
	})

	for nombre, ec := range map[string]schema.Contexto{
		"offline-sin-entrada":          {Offline: true},
		"offline-y-ensayo-sin-entrada": {Offline: true, DryRun: true},
	} {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, reproduce(t.TempDir()))

			resultado, err := banco.resuelve(t, ec, ConsultaAnalisis{Norma: normaVigente})

			compruebaFalloDeLaConsulta(t, resultado, err, falloSinAnalisisConOffline(analisisVigente))
			assert.Zero(t, banco.construcciones, "con --offline no se construye ningún cliente")
			banco.compruebaCacheSinCrear(t)
		})
	}

	for sufijo, ec := range ejecucionesDeLosArgumentos() {
		t.Run("norma-invalida"+sufijo, func(t *testing.T) {
			t.Parallel()

			compruebaArgumentosSinAbrirNada(t, ec, ConsultaAnalisis{Norma: "BOE-A-2015-10565/analisis"},
				`la norma "BOE-A-2015-10565/analisis" no tiene la forma BOE-A-<año>-<número>, `+
					"con cuatro dígitos en el año y de uno a nueve en el número")
		})
	}
}

// analisisGrabadoDeLaNormaVigente es el data de analisis de la Ley 39/2015
// compuesto a partir de su grabación sin el código de lectura: el cuerpo que
// sirve la reproducción, decodificado con la biblioteca estándar sobre una
// estructura fija con la forma que entrega la API —cada lista con su envoltorio—,
// con las materias, las notas y las referencias en su orden y el texto de cada
// referencia anterior tal como llega. compruebaAnalisisDeLaNormaVigente lo fija a
// mano en varias posiciones.
func analisisGrabadoDeLaNormaVigente(t *testing.T) Analisis {
	t.Helper()

	cliente := clienteDeReproduccion(t, carpetaDeLasGrabaciones, time.Now)

	respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{}, httpx.Peticion{
		Metodo: "GET",
		URL:    analisisVigente,
		Acepta: "application/json",
	})
	require.NoError(t, err)

	type referencia struct {
		Norma    string `json:"id_norma"`
		Relacion struct {
			Texto string `json:"texto"`
		} `json:"relacion"`
		Texto string `json:"texto"`
	}

	var grabado struct {
		Data []struct {
			Materias []struct {
				Materia struct {
					Codigo string `json:"codigo"`
					Texto  string `json:"texto"`
				} `json:"materia"`
			} `json:"materias"`
			Notas []struct {
				Nota []string `json:"nota"`
			} `json:"notas"`
			Referencias struct {
				Anteriores []struct {
					Anterior []referencia `json:"anterior"`
				} `json:"anteriores"`
				Posteriores []struct {
					Posterior []referencia `json:"posterior"`
				} `json:"posteriores"`
			} `json:"referencias"`
		} `json:"data"`
	}

	require.NoError(t, json.Unmarshal(respuesta.Cuerpo, &grabado))
	require.Len(t, grabado.Data, 1, "el análisis grabado de la Ley 39/2015 llega en un único elemento")

	elemento := grabado.Data[0]
	analisis := Analisis{
		Norma:       normaVigente,
		Materias:    []Materia{},
		Notas:       []string{},
		Referencias: Referencias{Anteriores: []ReferenciaAnterior{}, Posteriores: []ReferenciaPosterior{}},
	}

	for _, materia := range elemento.Materias {
		analisis.Materias = append(analisis.Materias, Materia{Codigo: materia.Materia.Codigo, Texto: materia.Materia.Texto})
	}

	for _, notas := range elemento.Notas {
		analisis.Notas = append(analisis.Notas, notas.Nota...)
	}

	for _, grupo := range elemento.Referencias.Anteriores {
		for _, anterior := range grupo.Anterior {
			analisis.Referencias.Anteriores = append(analisis.Referencias.Anteriores,
				ReferenciaAnterior{Relacion: anterior.Relacion.Texto, Norma: anterior.Norma, Texto: anterior.Texto})
		}
	}

	for _, grupo := range elemento.Referencias.Posteriores {
		for _, posterior := range grupo.Posterior {
			analisis.Referencias.Posteriores = append(analisis.Referencias.Posteriores,
				ReferenciaPosterior{Relacion: posterior.Relacion.Texto, Norma: posterior.Norma})
		}
	}

	return analisis
}

// compruebaAnalisisDeLaNormaVigente fija a mano, en el data de analisis de la Ley
// 39/2015, sus materias y sus notas enteras —con el doble espacio que trae una de
// ellas— y, de sus referencias, cuántas trae cada lista y, en su posición, la
// primera, la de la Ley 30/1992 que deroga y la última: la relación, la norma
// referida y, en las anteriores, el texto tal como llega (US5; data-model.md
// §2.6).
func compruebaAnalisisDeLaNormaVigente(t *testing.T, datos any) {
	t.Helper()

	analisis, esAnalisis := datos.(Analisis)
	require.Truef(t, esAnalisis, "el data de analisis no es un Analisis, sino %T", datos)

	assert.Equal(t, normaVigente, analisis.Norma)
	assert.Equal(t, []Materia{{Codigo: "6499", Texto: "Seguridad Social"}}, analisis.Materias)
	assert.Equal(t, []string{
		"Entrada en vigor, con la salvedad indicada, el 2 de octubre de 2016.",
		"Efectos  para las previsiones indicadas en la disposición final 7: 2 de abril de 2021.",
	}, analisis.Notas)

	require.Len(t, analisis.Referencias.Anteriores, anterioresDelAnalisisGrabado)

	anteriores := map[int]ReferenciaAnterior{
		0:  {Relacion: "DEROGA", Norma: "BOE-A-2011-4117", Texto: "los arts. 4 a 7 de la Ley 2/2011, de 4 de marzo"},
		6:  {Relacion: "DEROGA", Norma: "BOE-A-1992-26318", Texto: ", en la forma indicada, la  Ley 30/1992, de 26 de noviembre"},
		7:  {Relacion: "MODIFICA", Norma: "BOE-A-2011-15936", Texto: "los arts. 64, 69, 70, 72, 73, 85, 103 y 117 de  la Ley 36/2011, de 10 de octubre"},
		10: {Relacion: "CITA", Norma: "BOE-A-2003-21614", Texto: "Ley 47/2003, de 26 de noviembre"},
	}

	for posicion, esperada := range anteriores {
		assert.Equal(t, esperada, analisis.Referencias.Anteriores[posicion], "la referencia anterior en la posición %d", posicion)
	}

	require.Len(t, analisis.Referencias.Posteriores, posterioresDelAnalisisGrabado)

	posteriores := map[int]ReferenciaPosterior{
		0:  {Relacion: "SE DEJA SIN EFECTO", Norma: "BOE-A-2020-10491"},
		10: {Relacion: "SE AÑADE", Norma: "BOE-A-2024-22928"},
		17: {Relacion: "SE DECLARA", Norma: "BOE-A-2018-8574"},
	}

	for posicion, esperada := range posteriores {
		assert.Equal(t, esperada, analisis.Referencias.Posteriores[posicion], "la referencia posterior en la posición %d", posicion)
	}
}

// compruebaListasVaciasEnJSON exige que las listas del data de analisis sean
// listas también al escribirlo, nunca nulas: [] en el sobre y en la entrada
// (data-model.md §2).
func compruebaListasVaciasEnJSON(t *testing.T, datos any) {
	t.Helper()

	escrito, err := json.Marshal(datos)
	require.NoError(t, err)

	var listas struct {
		Materias    []any `json:"materias"`
		Notas       []any `json:"notas"`
		Referencias struct {
			Anteriores  []any `json:"anteriores"`
			Posteriores []any `json:"posteriores"`
		} `json:"referencias"`
	}

	require.NoError(t, json.Unmarshal(escrito, &listas))
	assert.NotNil(t, listas.Materias, "materias")
	assert.NotNil(t, listas.Notas, "notas")
	assert.NotNil(t, listas.Referencias.Anteriores, "referencias.anteriores")
	assert.NotNil(t, listas.Referencias.Posteriores, "referencias.posteriores")
}

// compruebaAnalisisSinEscribir exige, tras un fallo al resolver el análisis de la
// norma, que no se haya pedido nada más que su dirección y que no se haya
// escrito: con --offline no hay ninguna entrada que servir (FR-093).
func compruebaAnalisisSinEscribir(t *testing.T, banco *bancoDeLaFuente, norma, direccion string) {
	t.Helper()

	banco.compruebaPedidas(t, direccion)

	resultado, err := banco.resuelve(t, schema.Contexto{Offline: true}, ConsultaAnalisis{Norma: norma})

	compruebaFalloDeLaConsulta(t, resultado, err, falloSinAnalisisConOffline(direccion))
	banco.compruebaPedidas(t, direccion)
}

// falloSinAnalisisConOffline es el fallo del análisis sin entrada vigente con
// --offline: «fuente no disponible», código 4, con la dirección del análisis, sin
// instante —el sobre lo fecha el montaje— y un mensaje que nombra --offline, el
// verbo y la clave (contrato errores-y-codigos, fila 5).
func falloSinAnalisisConOffline(direccion string) falloDeLaConsulta {
	return falloDeLaConsulta{
		direccion: direccion,
		clase:     schema.ClaseFuenteNoDisponible,
		codigo:    4,
		mensaje: `con --offline no se pide nada a la fuente y no hay ninguna entrada vigente de analisis con la clave ` +
			`"boe.legislacion-consolidada|1|analisis|` + direccion + `" (` + direccion + ")",
	}
}

// falloSinAnalisis es el del análisis de la Ley 39/2015 con data vacío: «no
// encontrado», código 3, con el mensaje del 404 (contrato errores-y-codigos,
// fila 8; FR-061).
func falloSinAnalisis() falloDeLaConsulta {
	return falloDeLaConsulta{
		direccion: analisisVigente,
		clase:     schema.ClaseNoEncontrado,
		codigo:    3,
		mensaje:   "la norma BOE-A-2015-10565 no tiene análisis (" + analisisVigente + ")",
	}
}

// falloAlInterpretarElAnalisis es el del análisis de la Ley 39/2015 cuya
// respuesta no se puede interpretar: «fuente no disponible», código 4, con lo que
// no se pudo interpretar detrás de la dirección (contrato errores-y-codigos,
// fila 17).
func falloAlInterpretarElAnalisis(motivo string) falloDeLaConsulta {
	return falloDeLaConsulta{
		direccion: analisisVigente,
		clase:     schema.ClaseFuenteNoDisponible,
		codigo:    4,
		mensaje: "no se puede interpretar la respuesta a la petición del análisis de la norma BOE-A-2015-10565 (" +
			analisisVigente + "): " + motivo,
	}
}
