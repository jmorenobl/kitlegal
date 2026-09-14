package boe

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// Lo que comparten las pruebas de indice. Como en metadatos_test.go, las
// direcciones y las claves van escritas enteras, y no con las funciones de
// direcciones.go y entradas.go, para que un cambio en ellas no pase por aquí en
// silencio.
const (
	// indiceVigente e indiceInexistente son las direcciones del índice de la Ley
	// 39/2015 y del de la norma que la fuente responde con 404 (S2).
	indiceVigente     = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/indice"
	indiceInexistente = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2099-99999/texto/indice"
	// paginaDeLaNormaVigente es la dirección pública de la Ley 39/2015, la url
	// del data de su índice (refs/boe.py 384).
	paginaDeLaNormaVigente = "https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565"
	// sinteticoDeLaFuenteCaida es el índice de la Ley 39/2015 respondido con un
	// 503 y el cuerpo vacío (contrato esquemas-fixtures-y-controles §4).
	sinteticoDeLaFuenteCaida = "testdata/sintetico/fuente-caida/" + NombreDeLaFuente
	// vigenciaDelIndice es la de su entrada, siete días (FR-091).
	vigenciaDelIndice = 604_800 * time.Second
	// bloquesDelIndiceGrabado son los que trae el índice grabado de la Ley
	// 39/2015.
	bloquesDelIndiceGrabado = 195
)

// TestIndice fija el verbo indice (contrato verbos-y-salidas §2; data-model.md
// §2.4, §3.1 y §7.3; FR-040, FR-041, FR-070, FR-090 a FR-094, FR-096 y FR-101)
// sobre la fuente compuesta con el cliente real de httpx en reproducción de las
// grabaciones y los sintéticos, y con la caché real en una carpeta temporal, las
// dos gobernadas por el mismo reloj de prueba:
//
//   - la norma se valida antes de abrir la caché o construir el cliente;
//   - el índice se pide en JSON y sus bloques se leen en el orden de la fuente,
//     anidados en el primer elemento de data o planos, en lista o sueltos, sin
//     los que no son objeto, cada uno con su id y su título tal como llegan y el
//     tipo que TipoDesdeID infiere del id (US4, escenario 1);
//   - lo leído se escribe con el instante de la petición, se sirve de su entrada
//     con su url y su fecha de consulta sin pedir nada, también con --offline y
//     con --dry-run, y deja de servirse a los siete días;
//   - sin entrada, con --offline, «fuente no disponible» sin pedir nada, y con
//     --dry-run, la línea de la petición que se habría emitido, sin crear la
//     caché;
//   - y los fallos —el 404 y data vacío, que son «no encontrado» (US4,
//     escenario 4), el 503 y la respuesta que no se interpreta— llevan su clase,
//     la dirección del índice y el instante de la petición, y no se escriben.
//
// Todo resultado, de éxito o de fallo, lleva la fuente del BOE y una dirección
// de www.boe.es (FR-002, FR-101).
func TestIndice(t *testing.T) {
	t.Parallel()

	t.Run("lpac", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
		consulta := ConsultaIndice{Norma: normaVigente}
		esperado := resultadoResuelto(indiceVigente, banco.reloj.ahora(), indiceGrabadoDeLaNormaVigente(t))

		resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

		compruebaResuelta(t, resultado, err, esperado)
		compruebaBloquesDeLaNormaVigente(t, resultado.Datos)
		banco.compruebaPedidas(t, indiceVigente)

		banco.reloj.adelanta(vigenciaDelIndice - time.Nanosecond)

		for _, ec := range []schema.Contexto{{}, {Offline: true}, {DryRun: true}, {Offline: true, DryRun: true}} {
			servido, err := banco.resuelve(t, ec, consulta)

			require.NoError(t, err, "con %+v", ec)
			assert.Equal(t, esperado, servido, "con %+v", ec)
		}

		banco.compruebaPedidas(t, indiceVigente)
		assert.Equal(t, 1, banco.construcciones, "lo servido de la caché no construye ningún cliente")

		banco.reloj.adelanta(time.Nanosecond)

		caducado, err := banco.resuelve(t, schema.Contexto{Offline: true}, consulta)

		compruebaFalloDeLaConsulta(t, caducado, err, falloSinIndiceConOffline(indiceVigente))
		banco.compruebaPedidas(t, indiceVigente)

		esperado.Procedencia.FechaConsulta = banco.reloj.ahora()

		pedidoOtraVez, err := banco.resuelve(t, schema.Contexto{}, consulta)

		compruebaResuelta(t, pedidoOtraVez, err, esperado)
		banco.compruebaPedidas(t, indiceVigente, indiceVigente)
	})

	leidos := []struct {
		nombre  string
		cuerpo  string
		bloques []EntradaDeIndice
	}{
		{
			// Planos en data, con los tres ids del escenario 1 de US4.
			nombre: "planos",
			cuerpo: `{"data": [{"id": "a21", "titulo": "Artículo 21"}, ` +
				`{"id": "da3", "titulo": "Disposición adicional tercera"}, {"id": "dt1", "titulo": "Disposición transitoria primera"}]}`,
			bloques: []EntradaDeIndice{
				{ID: "a21", Titulo: "Artículo 21", Tipo: "articulo"},
				{ID: "da3", Titulo: "Disposición adicional tercera", Tipo: "disposicion_adicional"},
				{ID: "dt1", Titulo: "Disposición transitoria primera", Tipo: "disposicion_transitoria"},
			},
		},
		{
			// El elemento que anida y su bloque, sueltos: una lista de uno, y no
			// cada clave del objeto como un bloque, como en refs/boe.py 385
			// (FR-070).
			nombre:  "anidado-suelto-con-un-bloque-suelto",
			cuerpo:  `{"data": {"bloque": {"id": "cv3", "titulo": "CAPÍTULO V"}}}`,
			bloques: []EntradaDeIndice{{ID: "cv3", Titulo: "CAPÍTULO V", Tipo: "capitulo"}},
		},
		{
			// Lo que no es objeto se salta, y el id o el título ausentes o nulos son
			// la cadena vacía, nunca el marcador ? de refs/boe.py 387 (FR-016).
			nombre: "sin-objetos-ni-campos",
			cuerpo: `{"data": [{"bloque": ["a1", {"titulo": "Preámbulo"}, null, {"id": "a2", "titulo": null}, 3]}]}`,
			bloques: []EntradaDeIndice{
				{ID: "", Titulo: "Preámbulo", Tipo: ""},
				{ID: "a2", Titulo: "", Tipo: "articulo"},
			},
		},
		{
			// data no está vacío aunque no anide ningún bloque: no es «no
			// encontrado» (J3; refs/boe.py 362).
			nombre:  "anidado-sin-bloques",
			cuerpo:  `{"data": [{"bloque": []}]}`,
			bloques: []EntradaDeIndice{},
		},
	}

	for _, caso := range leidos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, respondeConElCuerpo(caso.cuerpo))
			consulta := ConsultaIndice{Norma: normaVigente}
			esperado := resultadoResuelto(indiceVigente, banco.reloj.ahora(),
				Indice{Norma: normaVigente, URL: paginaDeLaNormaVigente, Bloques: caso.bloques})

			resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

			compruebaResuelta(t, resultado, err, esperado)

			banco.reloj.adelanta(time.Second)

			servido, err := banco.resuelve(t, schema.Contexto{Offline: true}, consulta)

			compruebaResuelta(t, servido, err, esperado)
			banco.compruebaPedidas(t, indiceVigente)
		})
	}

	sinIndice := []struct {
		nombre  string
		pedidor func(*testing.T, *relojDePrueba) *pedidorDePrueba
		norma   string
		// fallo es el esperado salvo el instante, que es el de la petición.
		fallo falloDeLaConsulta
		// deHTTPX es la clase del error de httpx que el fallo envuelve, cuyo
		// texto sigue en el mensaje detrás del de la fuente (contrato
		// errores-y-codigos §2); vacía si no envuelve ninguno.
		deHTTPX schema.Clase
	}{
		{
			nombre:  "inexistente",
			pedidor: reproduce(carpetaDeLasGrabaciones),
			norma:   normaInexistente,
			fallo: falloDeLaConsulta{
				direccion: indiceInexistente,
				clase:     schema.ClaseNoEncontrado,
				codigo:    3,
				mensaje:   "la norma BOE-A-2099-99999 no tiene índice (" + indiceInexistente + ")",
			},
		},
		{
			nombre:  "fuente-caida",
			pedidor: reproduce(sinteticoDeLaFuenteCaida),
			norma:   normaVigente,
			fallo: falloDeLaConsulta{
				direccion: indiceVigente,
				clase:     schema.ClaseFuenteNoDisponible,
				codigo:    4,
				mensaje:   "ha fallado la petición del índice de la norma BOE-A-2015-10565 (" + indiceVigente + ")",
			},
			deHTTPX: schema.ClaseFuenteNoDisponible,
		},
		{
			nombre:  "ilegible",
			pedidor: respondeConElCuerpo(`{"data": [{"bloque": [`),
			norma:   normaVigente,
			fallo:   falloAlInterpretarElIndice("el cuerpo no es JSON legible"),
		},
		{
			nombre:  "id-que-no-es-texto",
			pedidor: respondeConElCuerpo(`{"data": [{"bloque": [{"id": "a1", "titulo": "Artículo 1"}, {"id": 21}]}]}`),
			norma:   normaVigente,
			fallo:   falloAlInterpretarElIndice(`el campo "id" no es texto, sino un número`),
		},
		{
			nombre:  "titulo-que-no-es-texto",
			pedidor: respondeConElCuerpo(`{"data": [{"id": "a21", "titulo": ["Artículo 21"]}]}`),
			norma:   normaVigente,
			fallo:   falloAlInterpretarElIndice(`el campo "titulo" no es texto, sino una lista`),
		},
		{nombre: "data-vacio", pedidor: respondeConElCuerpo(`{"status": {"code": "200"}, "data": []}`), norma: normaVigente, fallo: falloSinIndice()},
		{nombre: "data-nulo", pedidor: respondeConElCuerpo(`{"data": null}`), norma: normaVigente, fallo: falloSinIndice()},
		{nombre: "sin-data", pedidor: respondeConElCuerpo(`{"status": {}}`), norma: normaVigente, fallo: falloSinIndice()},
		{nombre: "data-cadena-vacia", pedidor: respondeConElCuerpo(`{"data": ""}`), norma: normaVigente, fallo: falloSinIndice()},
	}

	for _, caso := range sinIndice {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, caso.pedidor)
			esperado := caso.fallo
			esperado.instante = banco.reloj.ahora()

			resultado, err := banco.resuelve(t, schema.Contexto{}, ConsultaIndice{Norma: caso.norma})

			if caso.deHTTPX != "" {
				var deHTTPX *httpx.Error
				require.ErrorAs(t, err, &deHTTPX)
				assert.Equal(t, caso.deHTTPX, deHTTPX.Clase())

				esperado.mensaje += ": " + deHTTPX.Error()
			}

			compruebaFalloDeLaConsulta(t, resultado, err, esperado)
			compruebaIndiceSinEscribir(t, banco, caso.norma, esperado.direccion)
		})
	}

	t.Run("ensayo", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(t.TempDir()))

		resultado, err := banco.resuelve(t, schema.Contexto{DryRun: true}, ConsultaIndice{Norma: normaVigente})

		require.NoError(t, err)
		assert.Equal(t, schema.Resultado{
			Procedencia: schema.Procedencia{Fuente: NombreDeLaFuente, URL: indiceVigente},
			Ensayo:      []string{"GET " + indiceVigente},
		}, resultado)
		banco.compruebaPedidas(t, indiceVigente)
		require.Len(t, banco.pedidor.respuestas, 1)
		assert.True(t, banco.pedidor.respuestas[0].Ensayo, "bajo --dry-run no se emite la petición")
		banco.compruebaCacheSinCrear(t)
	})

	for nombre, ec := range map[string]schema.Contexto{
		"offline-sin-entrada":          {Offline: true},
		"offline-y-ensayo-sin-entrada": {Offline: true, DryRun: true},
	} {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, reproduce(t.TempDir()))

			resultado, err := banco.resuelve(t, ec, ConsultaIndice{Norma: normaVigente})

			compruebaFalloDeLaConsulta(t, resultado, err, falloSinIndiceConOffline(indiceVigente))
			assert.Zero(t, banco.construcciones, "con --offline no se construye ningún cliente")
			banco.compruebaCacheSinCrear(t)
		})
	}

	for sufijo, ec := range ejecucionesDeLosArgumentos() {
		t.Run("norma-invalida"+sufijo, func(t *testing.T) {
			t.Parallel()

			compruebaArgumentosSinAbrirNada(t, ec, ConsultaIndice{Norma: "BOE-A-15-10565"},
				`la norma "BOE-A-15-10565" no tiene la forma BOE-A-<año>-<número>, `+
					"con cuatro dígitos en el año y de uno a nueve en el número")
		})
	}
}

// indiceGrabadoDeLaNormaVigente es el data de indice de la Ley 39/2015 compuesto
// a partir de su grabación sin el código de lectura: el cuerpo que sirve la
// reproducción, decodificado sobre una estructura fija con la biblioteca
// estándar, con cada bloque en su orden, su id y su título tal como llegan y el
// tipo que TipoDesdeID infiere del id, que TestTipoDesdeID fija regla a regla y
// compruebaBloquesDeLaNormaVigente, bloque a bloque.
func indiceGrabadoDeLaNormaVigente(t *testing.T) Indice {
	t.Helper()

	cliente := clienteDeReproduccion(t, carpetaDeLasGrabaciones, time.Now)

	respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{}, httpx.Peticion{
		Metodo: "GET",
		URL:    indiceVigente,
		Acepta: "application/json",
	})
	require.NoError(t, err)

	var grabado struct {
		Data []struct {
			Bloque []struct {
				ID     string `json:"id"`
				Titulo string `json:"titulo"`
			} `json:"bloque"`
		} `json:"data"`
	}

	require.NoError(t, json.Unmarshal(respuesta.Cuerpo, &grabado))
	require.Len(t, grabado.Data, 1, "el índice grabado de la Ley 39/2015 llega anidado en un único elemento")

	indice := Indice{Norma: normaVigente, URL: paginaDeLaNormaVigente, Bloques: []EntradaDeIndice{}}
	for _, bloque := range grabado.Data[0].Bloque {
		indice.Bloques = append(indice.Bloques, EntradaDeIndice{ID: bloque.ID, Titulo: bloque.Titulo, Tipo: TipoDesdeID(bloque.ID)})
	}

	return indice
}

// compruebaBloquesDeLaNormaVigente fija a mano, en el data de indice de la Ley
// 39/2015, cuántos bloques trae y, en su posición, el id, el título tal como
// llega —con el espacio de no separación de los artículos— y el tipo de uno de
// cada clase: uno por cada regla de TipoDesdeID que casa en la norma, sus
// rarezas —ti es titulo y cv, capitulo— y el bloque sin regla, con tipo vacío
// (US4, escenario 1; data-model.md §4).
func compruebaBloquesDeLaNormaVigente(t *testing.T, datos any) {
	t.Helper()

	indice, esIndice := datos.(Indice)
	require.Truef(t, esIndice, "el data de indice no es un Indice, sino %T", datos)
	require.Len(t, indice.Bloques, bloquesDelIndiceGrabado)

	esperados := map[int]EntradaDeIndice{
		0:   {ID: "preambulo", Titulo: "", Tipo: "preambulo"},
		1:   {ID: "tpreliminar", Titulo: "TÍTULO PRELIMINAR", Tipo: "titulo"},
		2:   {ID: "a1", Titulo: "Artículo\u00a01", Tipo: "articulo"},
		4:   {ID: "ti", Titulo: "TÍTULO I", Tipo: "titulo"},
		5:   {ID: "ci", Titulo: "CAPÍTULO I", Tipo: "capitulo"},
		27:  {ID: "a21", Titulo: "Artículo\u00a021", Tipo: "articulo"},
		68:  {ID: "s1", Titulo: "Sección 1", Tipo: "seccion"},
		107: {ID: "cv", Titulo: "CAPÍTULO V", Tipo: "capitulo"},
		179: {ID: "da-3", Titulo: "Disposición adicional octava", Tipo: "disposicion_adicional"},
		181: {ID: "dtprimera", Titulo: "Disposición transitoria primera", Tipo: "disposicion_transitoria"},
		186: {ID: "ddunica", Titulo: "Disposición derogatoria única", Tipo: "disposicion_derogatoria"},
		193: {ID: "dfseptima", Titulo: "Disposición final séptima", Tipo: "disposicion_final"},
		194: {ID: "firma", Titulo: "", Tipo: ""},
	}

	for posicion, esperado := range esperados {
		assert.Equal(t, esperado, indice.Bloques[posicion], "el bloque en la posición %d", posicion)
	}
}

// compruebaIndiceSinEscribir exige, tras un fallo al resolver el índice de la
// norma, que no se haya pedido nada más que su dirección y que no se haya
// escrito: con --offline no hay ninguna entrada que servir (FR-093).
func compruebaIndiceSinEscribir(t *testing.T, banco *bancoDeLaFuente, norma, direccion string) {
	t.Helper()

	banco.compruebaPedidas(t, direccion)

	resultado, err := banco.resuelve(t, schema.Contexto{Offline: true}, ConsultaIndice{Norma: norma})

	compruebaFalloDeLaConsulta(t, resultado, err, falloSinIndiceConOffline(direccion))
	banco.compruebaPedidas(t, direccion)
}

// falloSinIndiceConOffline es el fallo del índice sin entrada vigente con
// --offline: «fuente no disponible», código 4, con la dirección del índice, sin
// instante —el sobre lo fecha el montaje— y un mensaje que nombra --offline, el
// verbo y la clave (contrato errores-y-codigos, fila 5).
func falloSinIndiceConOffline(direccion string) falloDeLaConsulta {
	return falloDeLaConsulta{
		direccion: direccion,
		clase:     schema.ClaseFuenteNoDisponible,
		codigo:    4,
		mensaje: `con --offline no se pide nada a la fuente y no hay ninguna entrada vigente de indice con la clave ` +
			`"boe.legislacion-consolidada|1|indice|` + direccion + `" (` + direccion + ")",
	}
}

// falloSinIndice es el del índice de la Ley 39/2015 con data vacío: «no
// encontrado», código 3, con el mensaje del 404 (contrato errores-y-codigos,
// fila 8; FR-041).
func falloSinIndice() falloDeLaConsulta {
	return falloDeLaConsulta{
		direccion: indiceVigente,
		clase:     schema.ClaseNoEncontrado,
		codigo:    3,
		mensaje:   "la norma BOE-A-2015-10565 no tiene índice (" + indiceVigente + ")",
	}
}

// falloAlInterpretarElIndice es el del índice de la Ley 39/2015 cuya respuesta
// no se puede interpretar: «fuente no disponible», código 4, con lo que no se
// pudo interpretar detrás de la dirección (contrato errores-y-codigos, fila 17).
func falloAlInterpretarElIndice(motivo string) falloDeLaConsulta {
	return falloDeLaConsulta{
		direccion: indiceVigente,
		clase:     schema.ClaseFuenteNoDisponible,
		codigo:    4,
		mensaje: "no se puede interpretar la respuesta a la petición del índice de la norma BOE-A-2015-10565 (" +
			indiceVigente + "): " + motivo,
	}
}
