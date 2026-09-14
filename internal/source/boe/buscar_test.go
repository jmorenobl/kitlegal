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

// Lo que comparten las pruebas de buscar. Como en metadatos_test.go, las
// direcciones y las claves van escritas enteras, y no con las funciones de
// busqueda.go y entradas.go, para que un cambio en ellas no pase por aquí en
// silencio.
const (
	// busquedaConResultados es la dirección de la búsqueda de «procedimiento
	// administrativo común», la del contrato verbos-y-salidas §1, que la fuente
	// responde con diez resultados.
	busquedaConResultados = "https://www.boe.es/datosabiertos/api/legislacion-consolidada?limit=10&query=" +
		"%7B%22query%22%3A%20%7B%22query_string%22%3A%20%7B%22query%22%3A%20%22titulo%3Aprocedimiento%20AND%20" +
		"titulo%3Aadministrativo%20AND%20titulo%3Acom%5Cu00fan%22%7D%7D%7D"
	// busquedaSinResultados es la de «zzqxkwvjh», que la fuente responde con data
	// vacío (FR-032).
	busquedaSinResultados = "https://www.boe.es/datosabiertos/api/legislacion-consolidada?limit=10&query=" +
		"%7B%22query%22%3A%20%7B%22query_string%22%3A%20%7B%22query%22%3A%20%22titulo%3Azzqxkwvjh%22%7D%7D%7D"
	// vigenciaDeLaBusqueda es la de su entrada, 300 s (FR-091).
	vigenciaDeLaBusqueda = 300 * time.Second
	// resultadosDeLaBusquedaGrabada son los que trae la búsqueda grabada con
	// resultados, el límite de la petición.
	resultadosDeLaBusquedaGrabada = 10
	// motivoDeLaBusquedaSinPalabras es el mensaje del texto sin ninguna palabra
	// (contrato errores-y-codigos §2, fila 4).
	motivoDeLaBusquedaSinPalabras = "la búsqueda no tiene ninguna palabra"
)

// TestBuscar fija el verbo buscar (contrato verbos-y-salidas §1; data-model.md
// §2.3, §3.1, §6 y §7.3; FR-030, FR-031, FR-032, FR-070, FR-090 a FR-094, FR-096
// y FR-101) sobre la fuente compuesta con el cliente real de httpx en
// reproducción de las grabaciones, y con la caché real en una carpeta temporal,
// las dos gobernadas por el mismo reloj de prueba:
//
//   - los argumentos se unen con un espacio y el texto sin operadores y sin
//     ninguna palabra es «argumentos» antes de abrir la caché o construir el
//     cliente (fila 4);
//   - la búsqueda se pide en JSON con la dirección de busqueda.go, que es la
//     misma para dos textos que construyen la misma consulta, y sus resultados se
//     leen en el orden de la fuente, en lista o uno suelto, sin los que no son
//     objeto, cada uno con la dirección pública de su identificador;
//   - lo leído se escribe con el instante de la petición, también la lista vacía
//     de la búsqueda sin resultados (FR-032), se sirve de su entrada con su url y
//     su fecha de consulta sin pedir nada, también con --offline y con
//     --dry-run, y deja de servirse a los 300 s;
//   - sin entrada, con --offline, «fuente no disponible» sin pedir nada, y con
//     --dry-run, la línea de la petición que se habría emitido, sin crear la
//     caché;
//   - y los fallos —el 404, que en una búsqueda no dice que nada no exista
//     (fila 9), cualquier otro estado, el error del cliente y la respuesta que no
//     se interpreta— llevan su clase y la dirección de la búsqueda, y no se
//     escriben.
//
// Todo resultado, de éxito o de fallo, lleva la fuente del BOE y una dirección
// de www.boe.es (FR-002, FR-101).
func TestBuscar(t *testing.T) {
	t.Parallel()

	t.Run("con-resultados", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
		consulta := ConsultaBuscar{Texto: []string{"procedimiento", "administrativo", "común"}}
		esperado := resultadoResuelto(busquedaConResultados, banco.reloj.ahora(), resultadosGrabadosConResultados(t))

		resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

		compruebaResuelta(t, resultado, err, esperado)
		compruebaResultadosConResultados(t, resultado.Datos)
		banco.compruebaPedidas(t, busquedaConResultados)

		banco.reloj.adelanta(vigenciaDeLaBusqueda - time.Nanosecond)

		// Unir los argumentos de otro modo no cambia las palabras: la misma
		// consulta, la misma dirección y la misma entrada (FR-090).
		otraUnion := ConsultaBuscar{Texto: []string{"procedimiento  administrativo", "\tcomún "}}

		for _, ec := range []schema.Contexto{{}, {Offline: true}, {DryRun: true}, {Offline: true, DryRun: true}} {
			for _, servida := range []ConsultaBuscar{consulta, otraUnion} {
				servido, err := banco.resuelve(t, ec, servida)

				require.NoError(t, err, "con %+v y %q", ec, servida.Texto)
				assert.Equal(t, esperado, servido, "con %+v y %q", ec, servida.Texto)
			}
		}

		banco.compruebaPedidas(t, busquedaConResultados)
		assert.Equal(t, 1, banco.construcciones, "lo servido de la caché no construye ningún cliente")

		banco.reloj.adelanta(time.Nanosecond)

		caducada, err := banco.resuelve(t, schema.Contexto{Offline: true}, consulta)

		compruebaFalloDeLaConsulta(t, caducada, err, falloSinBusquedaConOffline(busquedaConResultados))
		banco.compruebaPedidas(t, busquedaConResultados)

		esperado.Procedencia.FechaConsulta = banco.reloj.ahora()

		pedidaOtraVez, err := banco.resuelve(t, schema.Contexto{}, consulta)

		compruebaResuelta(t, pedidaOtraVez, err, esperado)
		banco.compruebaPedidas(t, busquedaConResultados, busquedaConResultados)
	})

	t.Run("sin-resultados-guardada", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
		consulta := ConsultaBuscar{Texto: []string{"zzqxkwvjh"}}
		esperado := resultadoResuelto(busquedaSinResultados, banco.reloj.ahora(), []ResultadoDeBusqueda{})

		resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

		compruebaResuelta(t, resultado, err, esperado)
		compruebaListaVaciaEnJSON(t, resultado.Datos)
		banco.compruebaPedidas(t, busquedaSinResultados)

		banco.reloj.adelanta(vigenciaDeLaBusqueda - time.Nanosecond)

		servido, err := banco.resuelve(t, schema.Contexto{Offline: true}, consulta)

		compruebaResuelta(t, servido, err, esperado)
		compruebaListaVaciaEnJSON(t, servido.Datos)
		banco.compruebaPedidas(t, busquedaSinResultados)
		assert.Equal(t, 1, banco.construcciones, "lo servido de la caché no construye ningún cliente")

		banco.reloj.adelanta(time.Nanosecond)

		caducada, err := banco.resuelve(t, schema.Contexto{Offline: true}, consulta)

		compruebaFalloDeLaConsulta(t, caducada, err, falloSinBusquedaConOffline(busquedaSinResultados))
		banco.compruebaPedidas(t, busquedaSinResultados)
	})

	t.Run("con-operadores", func(t *testing.T) {
		t.Parallel()

		// Con titulo: el texto unido va tal cual y construye la misma consulta que
		// las tres palabras sueltas: la misma dirección, que responde la misma
		// grabación (FR-030; research.md D8).
		banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
		consulta := ConsultaBuscar{
			Texto: []string{"titulo:procedimiento", "AND", "titulo:administrativo", "AND", "titulo:común"},
		}
		esperado := resultadoResuelto(busquedaConResultados, banco.reloj.ahora(), resultadosGrabadosConResultados(t))

		resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

		compruebaResuelta(t, resultado, err, esperado)
		banco.compruebaPedidas(t, busquedaConResultados)
	})

	leidos := []struct {
		nombre     string
		cuerpo     string
		resultados []ResultadoDeBusqueda
	}{
		{
			// Un resultado suelto es una lista de uno, y no cada clave del objeto
			// como un resultado, como en refs/boe.py 286 (FR-070).
			nombre: "resultado-suelto",
			cuerpo: `{"data": {"identificador": "BOE-A-2015-10565", "titulo": "Ley 39/2015", "rango": {"codigo": "1300", ` +
				`"texto": "Ley"}, "vigencia_agotada": "N", "estado_consolidacion": {"codigo": "3", "texto": "Finalizado"}}}`,
			resultados: []ResultadoDeBusqueda{{
				Identificador:       "BOE-A-2015-10565",
				Titulo:              "Ley 39/2015",
				Rango:               "Ley",
				VigenciaAgotada:     "N",
				EstadoConsolidacion: "Finalizado",
				URL:                 "https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565",
			}},
		},
		{
			// Lo que no es objeto se salta; los campos ausentes o nulos son la cadena
			// vacía, nunca el marcador ? de refs/boe.py 289-297 (FR-016); rango y
			// estado_consolidacion que no son objetos también, y la dirección
			// pública lleva el identificador tal como llega, vacío incluido
			// (refs/boe.py 298).
			nombre: "sin-objetos-ni-campos",
			cuerpo: `{"data": [3, "BOE-A-2015-10565", null, {"identificador": null, "rango": "Ley", ` +
				`"estado_consolidacion": ["Finalizado"]}, {"identificador": "BOE-A-1992-26318", "rango": {}}]}`,
			resultados: []ResultadoDeBusqueda{
				{URL: "https://www.boe.es/buscar/act.php?id="},
				{Identificador: "BOE-A-1992-26318", URL: "https://www.boe.es/buscar/act.php?id=BOE-A-1992-26318"},
			},
		},
		{
			// data no está vacío aunque ningún elemento sea un objeto: no hay ningún
			// resultado.
			nombre:     "ningun-objeto",
			cuerpo:     `{"data": ["BOE-A-2015-10565", 7]}`,
			resultados: []ResultadoDeBusqueda{},
		},
		// data vacío es la lista vacía, que también se guarda (J2, J3; FR-032).
		{nombre: "data-vacio", cuerpo: `{"status": {"code": "200"}, "data": []}`, resultados: []ResultadoDeBusqueda{}},
		{nombre: "data-nulo", cuerpo: `{"data": null}`, resultados: []ResultadoDeBusqueda{}},
		{nombre: "sin-data", cuerpo: `{"status": {}}`, resultados: []ResultadoDeBusqueda{}},
		{nombre: "data-cadena-vacia", cuerpo: `{"data": ""}`, resultados: []ResultadoDeBusqueda{}},
		{nombre: "data-objeto-vacio", cuerpo: `{"data": {}}`, resultados: []ResultadoDeBusqueda{}},
	}

	for _, caso := range leidos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, respondeConElCuerpo(caso.cuerpo))
			consulta := ConsultaBuscar{Texto: []string{"procedimiento", "administrativo", "común"}}
			esperado := resultadoResuelto(busquedaConResultados, banco.reloj.ahora(), caso.resultados)

			resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

			compruebaResuelta(t, resultado, err, esperado)

			banco.reloj.adelanta(time.Second)

			servido, err := banco.resuelve(t, schema.Contexto{Offline: true}, consulta)

			compruebaResuelta(t, servido, err, esperado)
			banco.compruebaPedidas(t, busquedaConResultados)
		})
	}

	causaDelCliente := errorConClaseDePrueba{clase: schema.ClaseLimiteOTos}

	sinBusqueda := []struct {
		nombre  string
		pedidor func(*testing.T, *relojDePrueba) *pedidorDePrueba
		// fallo es el esperado salvo el instante, que es el del reloj si conInstante.
		fallo       falloDeLaConsulta
		conInstante bool
	}{
		{
			// Una búsqueda nunca es «no encontrado»: su 404 dice que la fuente no
			// responde como se espera (fila 9).
			nombre:  "estado-404",
			pedidor: respondeConElEstado(404),
			fallo: falloDeLaConsulta{
				direccion: busquedaConResultados,
				clase:     schema.ClaseFuenteNoDisponible,
				codigo:    4,
				mensaje: "la fuente ha respondido con el estado 404 a la petición de la búsqueda (" +
					busquedaConResultados + ")",
			},
			conInstante: true,
		},
		{
			nombre:  "estado-403",
			pedidor: respondeConElEstado(403),
			fallo: falloDeLaConsulta{
				direccion: busquedaConResultados,
				clase:     schema.ClaseFuenteNoDisponible,
				codigo:    4,
				mensaje: "la fuente ha respondido con el estado 403 a la petición de la búsqueda (" +
					busquedaConResultados + ")",
			},
			conInstante: true,
		},
		{
			// El error del cliente, con su clase y la dirección de la búsqueda; uno
			// que no es de httpx no declara instante (filas 11 a 15).
			nombre: "fallo-del-cliente",
			pedidor: func(*testing.T, *relojDePrueba) *pedidorDePrueba {
				return &pedidorDePrueba{responde: fallaCon(causaDelCliente)}
			},
			fallo: falloDeLaConsulta{
				direccion: busquedaConResultados,
				clase:     schema.ClaseLimiteOTos,
				codigo:    5,
				mensaje:   "ha fallado la petición de la búsqueda (" + busquedaConResultados + "): " + causaDelCliente.Error(),
			},
		},
		{
			nombre:      "ilegible",
			pedidor:     respondeConElCuerpo(`{"data": [`),
			fallo:       falloAlInterpretarLaBusqueda("el cuerpo no es JSON legible"),
			conInstante: true,
		},
		{
			nombre:      "raiz-que-no-es-objeto",
			pedidor:     respondeConElCuerpo(`[{"identificador": "BOE-A-2015-10565"}]`),
			fallo:       falloAlInterpretarLaBusqueda("la raíz del JSON no es un objeto, sino una lista"),
			conInstante: true,
		},
		{
			nombre: "titulo-que-no-es-texto",
			pedidor: respondeConElCuerpo(`{"data": [{"identificador": "BOE-A-1992-26318", "titulo": "Ley 30/1992"}, ` +
				`{"identificador": "BOE-A-2015-10565", "titulo": 39}]}`),
			fallo:       falloAlInterpretarLaBusqueda(`el campo "titulo" no es texto, sino un número`),
			conInstante: true,
		},
		{
			nombre:      "texto-del-rango-que-no-es-texto",
			pedidor:     respondeConElCuerpo(`{"data": [{"identificador": "BOE-A-2015-10565", "rango": {"texto": true}}]}`),
			fallo:       falloAlInterpretarLaBusqueda(`el campo "rango.texto" no es texto, sino un booleano`),
			conInstante: true,
		},
		{
			nombre: "texto-del-estado-que-no-es-texto",
			pedidor: respondeConElCuerpo(`{"data": [{"identificador": "BOE-A-2015-10565", ` +
				`"estado_consolidacion": {"texto": ["Finalizado"]}}]}`),
			fallo:       falloAlInterpretarLaBusqueda(`el campo "estado_consolidacion.texto" no es texto, sino una lista`),
			conInstante: true,
		},
	}

	for _, caso := range sinBusqueda {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, caso.pedidor)
			consulta := ConsultaBuscar{Texto: []string{"procedimiento", "administrativo", "común"}}
			esperado := caso.fallo

			if caso.conInstante {
				esperado.instante = banco.reloj.ahora()
			}

			resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

			compruebaFalloDeLaConsulta(t, resultado, err, esperado)
			compruebaBusquedaSinEscribir(t, banco, consulta, esperado.direccion)
		})
	}

	t.Run("ensayo", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(t.TempDir()))
		consulta := ConsultaBuscar{Texto: []string{"procedimiento", "administrativo", "común"}}

		resultado, err := banco.resuelve(t, schema.Contexto{DryRun: true}, consulta)

		require.NoError(t, err)
		assert.Equal(t, schema.Resultado{
			Procedencia: schema.Procedencia{Fuente: NombreDeLaFuente, URL: busquedaConResultados},
			Ensayo:      []string{"GET " + busquedaConResultados},
		}, resultado)
		banco.compruebaPedidas(t, busquedaConResultados)
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
			consulta := ConsultaBuscar{Texto: []string{"procedimiento", "administrativo", "común"}}

			resultado, err := banco.resuelve(t, ec, consulta)

			compruebaFalloDeLaConsulta(t, resultado, err, falloSinBusquedaConOffline(busquedaConResultados))
			assert.Zero(t, banco.construcciones, "con --offline no se construye ningún cliente")
			banco.compruebaCacheSinCrear(t)
		})
	}

	sinPalabras := []struct {
		nombre string
		texto  []string
	}{
		{nombre: "sin-argumentos", texto: nil},
		{nombre: "un-argumento-vacio", texto: []string{""}},
		{nombre: "argumentos-vacios", texto: []string{"", ""}},
		{nombre: "solo-espacio-en-blanco", texto: []string{" ", "\t\n"}},
		{nombre: "espacio-en-blanco-de-unicode-y-de-python", texto: []string{"\u00a0", "\u3000\x1c"}},
	}

	for sufijo, ec := range ejecucionesDeLosArgumentos() {
		t.Run("texto-vacio"+sufijo, func(t *testing.T) {
			t.Parallel()

			for _, caso := range sinPalabras {
				t.Run(caso.nombre, func(t *testing.T) {
					t.Parallel()

					compruebaArgumentosSinAbrirNada(t, ec, ConsultaBuscar{Texto: caso.texto}, motivoDeLaBusquedaSinPalabras)
				})
			}
		})
	}
}

// resultadosGrabadosConResultados es el data de buscar de «procedimiento
// administrativo común» compuesto a partir de su grabación sin el código de
// lectura: el cuerpo que sirve la reproducción, decodificado sobre una estructura
// fija con la biblioteca estándar, con cada resultado en su orden y la dirección
// pública de su identificador, que compruebaResultadosConResultados fija a mano
// en varias posiciones.
func resultadosGrabadosConResultados(t *testing.T) []ResultadoDeBusqueda {
	t.Helper()

	cliente := clienteDeReproduccion(t, carpetaDeLasGrabaciones, time.Now)

	respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{}, httpx.Peticion{
		Metodo: "GET",
		URL:    busquedaConResultados,
		Acepta: "application/json",
	})
	require.NoError(t, err)

	type conTexto struct {
		Texto string `json:"texto"`
	}

	var grabada struct {
		Data []struct {
			Identificador       string   `json:"identificador"`
			Titulo              string   `json:"titulo"`
			Rango               conTexto `json:"rango"`
			VigenciaAgotada     string   `json:"vigencia_agotada"`
			EstadoConsolidacion conTexto `json:"estado_consolidacion"`
		} `json:"data"`
	}

	require.NoError(t, json.Unmarshal(respuesta.Cuerpo, &grabada))

	resultados := []ResultadoDeBusqueda{}
	for _, resultado := range grabada.Data {
		resultados = append(resultados, ResultadoDeBusqueda{
			Identificador:       resultado.Identificador,
			Titulo:              resultado.Titulo,
			Rango:               resultado.Rango.Texto,
			VigenciaAgotada:     resultado.VigenciaAgotada,
			EstadoConsolidacion: resultado.EstadoConsolidacion.Texto,
			URL:                 "https://www.boe.es/buscar/act.php?id=" + resultado.Identificador,
		})
	}

	return resultados
}

// compruebaResultadosConResultados fija a mano, en el data de buscar de
// «procedimiento administrativo común», cuántos resultados trae y, en su
// posición, el primero, el segundo —la Ley 39/2015, con vigencia— y el último,
// con los títulos ya decodificados de sus escapes y en el orden de la fuente
// (US3; data-model.md §2.3).
func compruebaResultadosConResultados(t *testing.T, datos any) {
	t.Helper()

	resultados, sonResultados := datos.([]ResultadoDeBusqueda)
	require.Truef(t, sonResultados, "el data de buscar no es []ResultadoDeBusqueda, sino %T", datos)
	require.Len(t, resultados, resultadosDeLaBusquedaGrabada)

	esperados := map[int]ResultadoDeBusqueda{
		0: {
			Identificador: "BOE-A-1992-26318",
			Titulo: "Ley 30/1992, de 26 de noviembre, de Régimen Jurídico de las Administraciones Públicas y del " +
				"Procedimiento Administrativo Común.",
			Rango:               "Ley",
			VigenciaAgotada:     "S",
			EstadoConsolidacion: "Finalizado",
			URL:                 "https://www.boe.es/buscar/act.php?id=BOE-A-1992-26318",
		},
		1: {
			Identificador:       "BOE-A-2015-10565",
			Titulo:              "Ley 39/2015, de 1 de octubre, del Procedimiento Administrativo Común de las Administraciones Públicas.",
			Rango:               "Ley",
			VigenciaAgotada:     "N",
			EstadoConsolidacion: "Finalizado",
			URL:                 "https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565",
		},
		9: {
			Identificador: "BOE-A-1994-19269",
			Titulo: "Real Decreto 1773/1994, de 5 de agosto, por el que se adecuan determinados procedimientos " +
				"administrativos en materia de telecomunicaciones a la Ley 30/1992, de 26 de noviembre, de Régimen " +
				"Jurídico de las Administraciones Públicas y del Procedimiento Administrativo Común.",
			Rango:               "Real Decreto",
			VigenciaAgotada:     "S",
			EstadoConsolidacion: "Finalizado",
			URL:                 "https://www.boe.es/buscar/act.php?id=BOE-A-1994-19269",
		},
	}

	for posicion, esperado := range esperados {
		assert.Equal(t, esperado, resultados[posicion], "el resultado en la posición %d", posicion)
	}
}

// compruebaListaVaciaEnJSON exige que el data de la búsqueda sin resultados sea
// la lista vacía también al escribirlo, nunca nulo: [] en el sobre y en la
// entrada, que no se podría volver a leer con datos nulos (FR-032).
func compruebaListaVaciaEnJSON(t *testing.T, datos any) {
	t.Helper()

	escrito, err := json.Marshal(datos)
	require.NoError(t, err)
	assert.JSONEq(t, `[]`, string(escrito))
}

// compruebaBusquedaSinEscribir exige, tras un fallo al resolver la búsqueda, que
// no se haya pedido nada más que su dirección y que no se haya escrito: con
// --offline no hay ninguna entrada que servir (FR-093).
func compruebaBusquedaSinEscribir(t *testing.T, banco *bancoDeLaFuente, consulta ConsultaBuscar, direccion string) {
	t.Helper()

	banco.compruebaPedidas(t, direccion)

	resultado, err := banco.resuelve(t, schema.Contexto{Offline: true}, consulta)

	compruebaFalloDeLaConsulta(t, resultado, err, falloSinBusquedaConOffline(direccion))
	banco.compruebaPedidas(t, direccion)
}

// respondeConElEstado es el Pedidor de prueba que responde a toda petición con
// el estado, sin cuerpo y con el instante del reloj al componerlo.
func respondeConElEstado(estado int) func(*testing.T, *relojDePrueba) *pedidorDePrueba {
	return func(t *testing.T, reloj *relojDePrueba) *pedidorDePrueba {
		t.Helper()

		return &pedidorDePrueba{responde: respondeCon(estado, nil, reloj.ahora())}
	}
}

// falloSinBusquedaConOffline es el fallo de la búsqueda sin entrada vigente con
// --offline: «fuente no disponible», código 4, con la dirección de la búsqueda,
// sin instante —el sobre lo fecha el montaje— y un mensaje que nombra --offline,
// el verbo y la clave (contrato errores-y-codigos, fila 5).
func falloSinBusquedaConOffline(direccion string) falloDeLaConsulta {
	return falloDeLaConsulta{
		direccion: direccion,
		clase:     schema.ClaseFuenteNoDisponible,
		codigo:    4,
		mensaje: `con --offline no se pide nada a la fuente y no hay ninguna entrada vigente de buscar con la clave ` +
			`"boe.legislacion-consolidada|1|buscar|` + direccion + `" (` + direccion + ")",
	}
}

// falloAlInterpretarLaBusqueda es el de la búsqueda de «procedimiento
// administrativo común» cuya respuesta no se puede interpretar: «fuente no
// disponible», código 4, con lo que no se pudo interpretar detrás de la
// dirección (contrato errores-y-codigos, fila 17).
func falloAlInterpretarLaBusqueda(motivo string) falloDeLaConsulta {
	return falloDeLaConsulta{
		direccion: busquedaConResultados,
		clase:     schema.ClaseFuenteNoDisponible,
		codigo:    4,
		mensaje:   "no se puede interpretar la respuesta a la petición de la búsqueda (" + busquedaConResultados + "): " + motivo,
	}
}
