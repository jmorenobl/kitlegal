package boe

import (
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// Lo que comparten las pruebas de metadatos. Las direcciones y las claves van
// escritas enteras, y no con las funciones de direcciones.go y entradas.go, para
// que un cambio en ellas no pase por aquí en silencio.
const (
	// normaVigente es la Ley 39/2015; normaDerogada, la Ley 30/1992, que la
	// primera derogó; y normaInexistente, la que la fuente responde con 404 (S2).
	normaVigente     = "BOE-A-2015-10565"
	normaDerogada    = "BOE-A-1992-26318"
	normaInexistente = "BOE-A-2099-99999"
	// metadatosVigente, metadatosDerogada y metadatosInexistente son las
	// direcciones de sus metadatos.
	metadatosVigente     = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/metadatos"
	metadatosDerogada    = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-1992-26318/metadatos"
	metadatosInexistente = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2099-99999/metadatos"
	// Los sintéticos de los metadatos de la Ley 39/2015 (contrato
	// esquemas-fixtures-y-controles §4).
	sinteticoDeLosAvisos             = "testdata/sintetico/avisos/" + NombreDeLaFuente
	sinteticoDelLimite               = "testdata/sintetico/limite/" + NombreDeLaFuente
	sinteticoDeLosMetadatosIlegibles = "testdata/sintetico/metadatos-ilegibles/" + NombreDeLaFuente
	// vigenciaDeLosMetadatos es la de su entrada, 300 s (FR-091).
	vigenciaDeLosMetadatos = 300 * time.Second
)

// TestMetadatos fija el verbo metadatos (contrato verbos-y-salidas §5;
// data-model.md §2.5 y §7.3; FR-050, FR-051, FR-090 a FR-094, FR-096 y FR-101)
// sobre la fuente compuesta con el cliente real de httpx en reproducción de las
// grabaciones y los sintéticos, y con la caché real en una carpeta temporal, las
// dos gobernadas por el mismo reloj de prueba:
//
//   - la norma se valida antes de abrir la caché o construir el cliente;
//   - la entrada vigente se sirve con su url y su fecha de consulta sin pedir
//     nada, también con --offline y con --dry-run, y deja de servirse a los
//     300 s;
//   - sin ella, con --offline, «fuente no disponible» sin pedir nada, y con
//     --dry-run, la línea de la petición que se habría emitido, sin crear la
//     caché;
//   - si no, se piden en JSON, y lo leído se escribe con el instante de la
//     petición, que es su fecha de consulta;
//   - y los fallos —404, data vacío, 429, respuesta que no se interpreta— llevan
//     su clase, la dirección de los metadatos y el instante de la petición, y
//     no se escriben.
//
// Todo resultado, de éxito o de fallo, lleva la fuente del BOE y una dirección
// de www.boe.es (FR-002, FR-101).
func TestMetadatos(t *testing.T) {
	t.Parallel()

	t.Run("vigente", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
		consulta := ConsultaMetadatos{Norma: normaVigente}
		esperado := resultadoResuelto(metadatosVigente, banco.reloj.ahora(), metadatosDeLaNormaVigente())

		resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

		compruebaResuelta(t, resultado, err, esperado)
		banco.compruebaPedidas(t, metadatosVigente)

		banco.reloj.adelanta(vigenciaDeLosMetadatos - time.Nanosecond)

		for _, ec := range []schema.Contexto{{}, {Offline: true}, {DryRun: true}, {Offline: true, DryRun: true}} {
			resultado, err := banco.resuelve(t, ec, consulta)

			require.NoError(t, err, "con %+v", ec)
			assert.Equal(t, esperado, resultado, "con %+v", ec)
		}

		banco.compruebaPedidas(t, metadatosVigente)
		assert.Equal(t, 1, banco.construcciones, "lo servido de la caché no construye ningún cliente")
	})

	t.Run("caducada", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
		consulta := ConsultaMetadatos{Norma: normaVigente}

		_, err := banco.resuelve(t, schema.Contexto{}, consulta)
		require.NoError(t, err)

		banco.reloj.adelanta(vigenciaDeLosMetadatos)

		resultado, err := banco.resuelve(t, schema.Contexto{Offline: true}, consulta)

		compruebaFalloDeLaConsulta(t, resultado, err, falloSinEntradaConOffline(metadatosVigente))
		banco.compruebaPedidas(t, metadatosVigente)

		esperado := resultadoResuelto(metadatosVigente, banco.reloj.ahora(), metadatosDeLaNormaVigente())

		resultado, err = banco.resuelve(t, schema.Contexto{}, consulta)

		compruebaResuelta(t, resultado, err, esperado)
		banco.compruebaPedidas(t, metadatosVigente, metadatosVigente)
	})

	resueltas := []struct {
		nombre    string
		pedidor   func(*testing.T, *relojDePrueba) *pedidorDePrueba
		norma     string
		direccion string
		datos     Metadatos
	}{
		{
			nombre:    "derogada",
			pedidor:   reproduce(carpetaDeLasGrabaciones),
			norma:     normaDerogada,
			direccion: metadatosDerogada,
			datos:     metadatosDeLaNormaDerogada(),
		},
		{
			nombre:    "tres-avisos",
			pedidor:   reproduce(sinteticoDeLosAvisos),
			norma:     normaVigente,
			direccion: metadatosVigente,
			datos:     metadatosConTresAvisos(),
		},
		{
			// Un objeto suelto en data, y rango y estado_consolidacion como
			// cadenas: el str() de refs/boe.py 466 y 469, sin código de
			// consolidación, y los campos que faltan, vacíos (FR-016).
			nombre:    "objeto-suelto-con-cadenas",
			pedidor:   respondeConElCuerpo(`{"data": {"rango": "Ley", "estado_consolidacion": "Finalizado"}}`),
			norma:     normaVigente,
			direccion: metadatosVigente,
			datos: Metadatos{
				Norma:               normaVigente,
				Rango:               "Ley",
				EstadoConsolidacion: EstadoDeConsolidacion{Texto: "Finalizado"},
				Avisos:              []Aviso{},
			},
		},
	}

	for _, caso := range resueltas {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, caso.pedidor)
			consulta := ConsultaMetadatos{Norma: caso.norma}
			esperado := resultadoResuelto(caso.direccion, banco.reloj.ahora(), caso.datos)

			resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

			compruebaResuelta(t, resultado, err, esperado)
			banco.compruebaPedidas(t, caso.direccion)

			banco.reloj.adelanta(time.Second)

			resultado, err = banco.resuelve(t, schema.Contexto{Offline: true}, consulta)

			compruebaResuelta(t, resultado, err, esperado)
		})
	}

	fallidas := []struct {
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
				direccion: metadatosInexistente,
				clase:     schema.ClaseNoEncontrado,
				codigo:    3,
				mensaje:   "la norma BOE-A-2099-99999 no tiene metadatos (" + metadatosInexistente + ")",
			},
		},
		{
			nombre:  "limite",
			pedidor: reproduce(sinteticoDelLimite),
			norma:   normaVigente,
			fallo: falloDeLaConsulta{
				direccion: metadatosVigente,
				clase:     schema.ClaseLimiteOTos,
				codigo:    5,
				mensaje:   "ha fallado la petición de los metadatos de la norma BOE-A-2015-10565 (" + metadatosVigente + ")",
			},
			deHTTPX: schema.ClaseLimiteOTos,
		},
		{
			nombre:  "ilegible",
			pedidor: reproduce(sinteticoDeLosMetadatosIlegibles),
			norma:   normaVigente,
			fallo:   falloAlInterpretar("el cuerpo no es JSON legible"),
		},
		{
			nombre:  "primer-elemento-que-no-es-objeto",
			pedidor: respondeConElCuerpo(`{"data": ["BOE-A-2015-10565"]}`),
			norma:   normaVigente,
			fallo:   falloAlInterpretar("el primer elemento de data no es un objeto, sino una cadena"),
		},
		{
			nombre:  "campo-que-no-es-texto",
			pedidor: respondeConElCuerpo(`{"data": [{"titulo": "Ley 39/2015", "numero_oficial": 39}]}`),
			norma:   normaVigente,
			fallo:   falloAlInterpretar(`el campo "numero_oficial" no es texto, sino un número`),
		},
		{
			nombre:  "data-vacio",
			pedidor: respondeConElCuerpo(`{"status": {"code": "200", "text": "ok"}, "data": []}`),
			norma:   normaVigente,
			fallo:   falloSinMetadatos(),
		},
		{nombre: "data-nulo", pedidor: respondeConElCuerpo(`{"data": null}`), norma: normaVigente, fallo: falloSinMetadatos()},
		{nombre: "sin-data", pedidor: respondeConElCuerpo(`{"status": {}}`), norma: normaVigente, fallo: falloSinMetadatos()},
		{nombre: "data-objeto-vacio", pedidor: respondeConElCuerpo(`{"data": {}}`), norma: normaVigente, fallo: falloSinMetadatos()},
	}

	for _, caso := range fallidas {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, caso.pedidor)
			consulta := ConsultaMetadatos{Norma: caso.norma}
			esperado := caso.fallo
			esperado.instante = banco.reloj.ahora()

			resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

			if caso.deHTTPX != "" {
				var deHTTPX *httpx.Error
				require.ErrorAs(t, err, &deHTTPX)
				assert.Equal(t, caso.deHTTPX, deHTTPX.Clase())

				esperado.mensaje += ": " + deHTTPX.Error()
			}

			compruebaFalloDeLaConsulta(t, resultado, err, esperado)
			banco.compruebaPedidas(t, esperado.direccion)

			resultado, err = banco.resuelve(t, schema.Contexto{Offline: true}, consulta)

			compruebaFalloDeLaConsulta(t, resultado, err, falloSinEntradaConOffline(esperado.direccion))
			banco.compruebaPedidas(t, esperado.direccion)
		})
	}

	t.Run("ensayo", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(t.TempDir()))

		resultado, err := banco.resuelve(t, schema.Contexto{DryRun: true}, ConsultaMetadatos{Norma: normaVigente})

		require.NoError(t, err)
		compruebaDireccionDeLaFuente(t, resultado.Procedencia.URL)
		assert.Equal(t, schema.Resultado{
			Procedencia: schema.Procedencia{Fuente: NombreDeLaFuente, URL: metadatosVigente},
			Ensayo:      []string{"GET " + metadatosVigente},
		}, resultado)
		banco.compruebaPedidas(t, metadatosVigente)
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

			resultado, err := banco.resuelve(t, ec, ConsultaMetadatos{Norma: normaVigente})

			compruebaFalloDeLaConsulta(t, resultado, err, falloSinEntradaConOffline(metadatosVigente))
			assert.Zero(t, banco.construcciones, "con --offline no se construye ningún cliente")
			banco.compruebaCacheSinCrear(t)
		})
	}

	for nombre, ec := range map[string]schema.Contexto{
		"norma-invalida":           {},
		"norma-invalida-offline":   {Offline: true},
		"norma-invalida-en-ensayo": {DryRun: true},
	} {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			dependencias := &dependenciasDePrueba{cacheAbierta: &cacheDePrueba{}}
			fuente := fuenteDePrueba(t, dependencias)

			resultado, err := fuente.Fetch(t.Context(), ec, ConsultaMetadatos{Norma: "BOE-A-2015-1056a"})

			assert.Zero(t, resultado, "el sobre de los argumentos lo firma y lo fecha el kernel")

			var fallo *Error
			require.ErrorAs(t, err, &fallo)
			assert.Empty(t, fallo.URL)
			assert.Zero(t, fallo.Instante)
			assert.Equal(t, schema.ClaseArgumentos, cli.Clasificar(err))
			assert.Equal(t, 2, cli.CodigoSalida(err))
			assert.Equal(t, `la norma "BOE-A-2015-1056a" no tiene la forma BOE-A-<año>-<número>, `+
				"con cuatro dígitos en el año y de uno a nueve en el número", err.Error())
			dependencias.compruebaSinUso(t)
		})
	}
}

// metadatosDeLaNormaVigente son los de la Ley 39/2015 tal como los da su
// grabación, sin ningún aviso.
func metadatosDeLaNormaVigente() Metadatos {
	return Metadatos{
		Norma:               normaVigente,
		Titulo:              "Ley 39/2015, de 1 de octubre, del Procedimiento Administrativo Común de las Administraciones Públicas.",
		Rango:               "Ley",
		NumeroOficial:       "39/2015",
		FechaDisposicion:    "20151001",
		FechaPublicacion:    "20151002",
		FechaVigencia:       "20161002",
		EstatusDerogacion:   "N",
		VigenciaAgotada:     "N",
		EstadoConsolidacion: EstadoDeConsolidacion{Codigo: "3", Texto: "Finalizado"},
		URLELI:              "https://www.boe.es/eli/es/l/2015/10/01/39",
		Avisos:              []Aviso{},
	}
}

// metadatosDeLaNormaDerogada son los de la Ley 30/1992 tal como los da su
// grabación: derogada y con la vigencia agotada, con esos dos avisos en su orden
// (US5, escenario 1).
func metadatosDeLaNormaDerogada() Metadatos {
	return Metadatos{
		Norma: normaDerogada,
		Titulo: "Ley 30/1992, de 26 de noviembre, de Régimen Jurídico de las Administraciones Públicas " +
			"y del Procedimiento Administrativo Común.",
		Rango:               "Ley",
		NumeroOficial:       "30/1992",
		FechaDisposicion:    "19921126",
		FechaPublicacion:    "19921127",
		FechaVigencia:       "19930227",
		EstatusDerogacion:   "S",
		VigenciaAgotada:     "S",
		EstadoConsolidacion: EstadoDeConsolidacion{Codigo: "3", Texto: "Finalizado"},
		URLELI:              "https://www.boe.es/eli/es/l/1992/11/26/30",
		Avisos: []Aviso{
			{Codigo: "derogada", Texto: fraseEsperadaDeNormaDerogada},
			{Codigo: "vigencia-agotada", Texto: fraseEsperadaDeVigenciaAgotada},
		},
	}
}

// metadatosConTresAvisos son los del sintético de los avisos: los de la Ley
// 39/2015 con la consolidación sin finalizar, derogada y con la vigencia agotada,
// y los tres avisos en el orden de _check_vigencia, con sus frases literales
// (refs/boe.py 197-211; FR-012, FR-050).
func metadatosConTresAvisos() Metadatos {
	metadatos := metadatosDeLaNormaVigente()
	metadatos.EstatusDerogacion = "S"
	metadatos.VigenciaAgotada = "S"
	metadatos.EstadoConsolidacion.Codigo = "4"
	metadatos.Avisos = []Aviso{
		{Codigo: "consolidacion-no-finalizada", Texto: fraseEsperadaDeConsolidacionNoFinalizada},
		{Codigo: "derogada", Texto: fraseEsperadaDeNormaDerogada},
		{Codigo: "vigencia-agotada", Texto: fraseEsperadaDeVigenciaAgotada},
	}

	return metadatos
}

// falloDeLaConsulta es lo que el applet necesita de un fallo al resolver una
// consulta para montar su sobre y su código de salida.
type falloDeLaConsulta struct {
	direccion string
	instante  time.Time
	clase     schema.Clase
	codigo    int
	mensaje   string
}

// falloSinEntradaConOffline es el fallo de los metadatos sin entrada vigente con
// --offline: «fuente no disponible», código 4, con la dirección de los
// metadatos, sin instante —el sobre lo fecha el montaje— y un mensaje que nombra
// --offline, el verbo y la clave (contrato errores-y-codigos, fila 5).
func falloSinEntradaConOffline(direccion string) falloDeLaConsulta {
	return falloDeLaConsulta{
		direccion: direccion,
		clase:     schema.ClaseFuenteNoDisponible,
		codigo:    4,
		mensaje: `con --offline no se pide nada a la fuente y no hay ninguna entrada vigente de metadatos con la clave ` +
			`"boe.legislacion-consolidada|1|metadatos|` + direccion + `" (` + direccion + ")",
	}
}

// falloSinMetadatos es el de los metadatos de la Ley 39/2015 con data vacío:
// «no encontrado», código 3, con el mensaje del 404 (contrato
// errores-y-codigos, fila 8; FR-051).
func falloSinMetadatos() falloDeLaConsulta {
	return falloDeLaConsulta{
		direccion: metadatosVigente,
		clase:     schema.ClaseNoEncontrado,
		codigo:    3,
		mensaje:   "la norma BOE-A-2015-10565 no tiene metadatos (" + metadatosVigente + ")",
	}
}

// falloAlInterpretar es el de los metadatos de la Ley 39/2015 cuya respuesta no
// se puede interpretar: «fuente no disponible», código 4, con lo que no se pudo
// interpretar detrás de la dirección (contrato errores-y-codigos, fila 17).
func falloAlInterpretar(motivo string) falloDeLaConsulta {
	return falloDeLaConsulta{
		direccion: metadatosVigente,
		clase:     schema.ClaseFuenteNoDisponible,
		codigo:    4,
		mensaje: "no se puede interpretar la respuesta a la petición de los metadatos de la norma BOE-A-2015-10565 (" +
			metadatosVigente + "): " + motivo,
	}
}

// resultadoResuelto es el Resultado de una consulta resuelta: la procedencia de
// la fuente con la dirección y la fecha de consulta, y los datos.
func resultadoResuelto(direccion string, fechaConsulta time.Time, datos any) schema.Resultado {
	return schema.Resultado{
		Procedencia: schema.Procedencia{Fuente: NombreDeLaFuente, URL: direccion, FechaConsulta: fechaConsulta},
		Datos:       datos,
	}
}

// compruebaResuelta exige una consulta resuelta sin fallo, con una dirección que
// el sobre puede citar y el resultado esperado.
func compruebaResuelta(t *testing.T, resultado schema.Resultado, err error, esperado schema.Resultado) {
	t.Helper()

	require.NoError(t, err)
	compruebaDireccionDeLaFuente(t, resultado.Procedencia.URL)
	assert.Equal(t, esperado, resultado)
}

// compruebaFalloDeLaConsulta exige el fallo de una consulta tal como lo recibe el
// applet: un *Error alcanzable con errors.As con la dirección, que el sobre puede
// citar, y el instante esperados; la clase y el código que le da el kernel; el
// mensaje entero; y un resultado con la procedencia del fallo y nada más.
func compruebaFalloDeLaConsulta(t *testing.T, resultado schema.Resultado, err error, esperado falloDeLaConsulta) {
	t.Helper()

	var fallo *Error
	require.ErrorAs(t, err, &fallo)
	compruebaDireccionDeLaFuente(t, fallo.URL)
	assert.Equal(t, esperado.direccion, fallo.URL)
	assert.Equal(t, esperado.instante, fallo.Instante)
	assert.Equal(t, esperado.clase, cli.Clasificar(err))
	assert.Equal(t, esperado.codigo, cli.CodigoSalida(err))
	assert.Equal(t, esperado.mensaje, err.Error())
	assert.Equal(t, schema.Resultado{Procedencia: schema.Procedencia{
		Fuente:        NombreDeLaFuente,
		URL:           esperado.direccion,
		FechaConsulta: esperado.instante,
	}}, resultado)
}

// relojDePrueba da siempre el mismo instante hasta que la prueba lo adelanta, de
// modo que cada fecha que declara la fuente se conoce de antemano. Es seguro para
// usarlo desde varias goroutines.
type relojDePrueba struct {
	inicio       time.Time
	transcurrido atomic.Int64
}

// ahora es el instante del reloj.
func (r *relojDePrueba) ahora() time.Time {
	return r.inicio.Add(time.Duration(r.transcurrido.Load()))
}

// adelanta mueve el reloj el tramo.
func (r *relojDePrueba) adelanta(tramo time.Duration) {
	r.transcurrido.Add(int64(tramo))
}

// bancoDeLaFuente es la fuente de una prueba de verbo con lo que la rodea: el
// Pedidor que anota cada petición, la caché real en una carpeta temporal y el
// reloj que gobierna a la vez la hora de emisión y el «ahora» de la caché.
type bancoDeLaFuente struct {
	fuente           *Fuente
	pedidor          *pedidorDePrueba
	reloj            *relojDePrueba
	carpetaDeLaCache string
	// construcciones cuenta las veces que la fuente construyó su cliente.
	construcciones int
}

// nuevoBanco compone la fuente con el Pedidor que construye pedidor, la caché
// real en una carpeta temporal y un reloj de prueba para los dos.
func nuevoBanco(t *testing.T, pedidor func(*testing.T, *relojDePrueba) *pedidorDePrueba) *bancoDeLaFuente {
	t.Helper()

	reloj := &relojDePrueba{inicio: time.Date(2026, time.September, 13, 10, 30, 0, 123456789, time.UTC)}
	banco := &bancoDeLaFuente{pedidor: pedidor(t, reloj), reloj: reloj, carpetaDeLaCache: t.TempDir()}

	fuente, err := Nueva(
		ConCliente(banco.construir),
		ConCache(abrirCacheEn(banco.carpetaDeLaCache, cache.ConReloj(reloj.ahora))),
	)
	require.NoError(t, err)

	banco.fuente = fuente

	return banco
}

// reproduce es el Pedidor de prueba sobre el cliente real de httpx en
// reproducción de la carpeta, con la hora de emisión del reloj.
func reproduce(carpeta string) func(*testing.T, *relojDePrueba) *pedidorDePrueba {
	return func(t *testing.T, reloj *relojDePrueba) *pedidorDePrueba {
		t.Helper()

		return &pedidorDePrueba{responde: clienteDeReproduccion(t, carpeta, reloj.ahora).Pedir}
	}
}

// respondeConElCuerpo es el Pedidor de prueba que responde a toda petición con
// un 200, el cuerpo y el instante del reloj al componerlo.
func respondeConElCuerpo(cuerpo string) func(*testing.T, *relojDePrueba) *pedidorDePrueba {
	return func(t *testing.T, reloj *relojDePrueba) *pedidorDePrueba {
		t.Helper()

		const estadoCorrecto = 200

		return &pedidorDePrueba{responde: respondeCon(estadoCorrecto, []byte(cuerpo), reloj.ahora())}
	}
}

// construir es la función de ConCliente del banco.
func (b *bancoDeLaFuente) construir() (Pedidor, error) {
	b.construcciones++

	return b.pedidor, nil
}

// resuelve resuelve la consulta con la fuente del banco.
func (b *bancoDeLaFuente) resuelve(t *testing.T, ec schema.Contexto, consulta core.Consulta) (schema.Resultado, error) {
	t.Helper()

	return b.fuente.Fetch(t.Context(), ec, consulta)
}

// compruebaPedidas exige que la fuente haya pedido, en orden, exactamente esas
// direcciones, cada una con GET y en JSON (FR-003).
func (b *bancoDeLaFuente) compruebaPedidas(t *testing.T, direcciones ...string) {
	t.Helper()

	pedidas := make([]httpx.Peticion, 0, len(direcciones))
	for _, direccion := range direcciones {
		pedidas = append(pedidas, httpx.Peticion{Metodo: "GET", URL: direccion, Acepta: "application/json"})
	}

	assert.Equal(t, pedidas, b.pedidor.peticiones)
}

// compruebaCacheSinCrear exige que la carpeta de la caché siga vacía: con
// --offline y con --dry-run la caché se abre en solo lectura y no crea ni
// escribe nada (FR-092, FR-094).
func (b *bancoDeLaFuente) compruebaCacheSinCrear(t *testing.T) {
	t.Helper()

	entradas, err := os.ReadDir(b.carpetaDeLaCache)
	require.NoError(t, err)
	assert.Empty(t, entradas)
}
