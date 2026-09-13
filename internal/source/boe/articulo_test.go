package boe

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// Lo que comparten las pruebas de articulo, además de las normas, las
// direcciones de sus metadatos y el banco de metadatos_test.go. Las direcciones
// y las claves van escritas enteras, y no con las funciones de direcciones.go y
// entradas.go, para que un cambio en ellas no pase por aquí en silencio.
const (
	// bloqueDelArticulo21 y bloqueDelArticulo42 son el artículo 21 de la Ley
	// 39/2015 y el 42 de la Ley 30/1992, dos artículos del diff de aceptación; y
	// bloqueInexistente, el de la Ley 39/2015 que la fuente responde con 404 (S2).
	bloqueDelArticulo21 = "a21"
	bloqueDelArticulo42 = "a42"
	bloqueInexistente   = "a9999"
	// direccionDelArticulo21, direccionDelArticulo42 y
	// direccionDelBloqueInexistente son las direcciones de sus bloques en la API.
	direccionDelArticulo21        = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21"
	direccionDelArticulo42        = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-1992-26318/texto/bloque/a42"
	direccionDelBloqueInexistente = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565" +
		"/texto/bloque/a9999"
	// Los sintéticos del bloque a21 y de los metadatos de la Ley 39/2015 que solo
	// usa articulo (contrato esquemas-fixtures-y-controles §4); los de los avisos
	// y de los metadatos ilegibles son los de metadatos_test.go.
	sinteticoDelBloqueIlegible    = "testdata/sintetico/bloque-ilegible/" + NombreDeLaFuente
	sinteticoDelBloqueSinElemento = "testdata/sintetico/bloque-sin-elemento/" + NombreDeLaFuente
	sinteticoDeLosMetadatosCaidos = "testdata/sintetico/metadatos-caidos/" + NombreDeLaFuente
	// vigenciaDeLosArticulos es la de la entrada de articulo, siete días (FR-091).
	vigenciaDeLosArticulos = 604_800 * time.Second
)

// TestArticulo fija el verbo articulo (contrato verbos-y-salidas §3;
// data-model.md §2.1 y §7.1; FR-010 a FR-016, FR-090, FR-093, FR-096 y FR-101)
// sobre la fuente compuesta con el cliente real de httpx en reproducción de las
// grabaciones y los sintéticos, y con la caché real en una carpeta temporal, las
// dos gobernadas por el mismo reloj de prueba. Cada subtest cuenta las peticiones
// que la fuente entrega al cliente y mira cada una:
//
//   - la norma y el bloque se validan antes de abrir la caché o construir el
//     cliente;
//   - la entrada vigente del artículo se sirve con su url y su fecha sin pedir
//     nada ni mirar los metadatos, también con --offline y con --dry-run, y deja
//     de servirse a los siete días;
//   - sin ella, con --offline, «fuente no disponible» con la dirección del
//     bloque, aunque los metadatos estén guardados;
//   - si no, se pide el bloque en XML y después, solo si su entrada no está
//     vigente, los metadatos en JSON, que se escriben (FR-013, FR-090); bajo
//     --dry-run, una línea por cada petición que se habría emitido y nada
//     escrito;
//   - el artículo lleva los datos de la última versión del bloque, los avisos y
//     el ELI de los metadatos, la huella de su texto y la dirección pública del
//     bloque (US1, escenarios 1 a 3);
//   - el fallo del bloque —404, cuerpo ilegible o sin bloque— termina sin pedir
//     los metadatos, con la dirección y el instante del bloque (US1, escenario 5;
//     FR-014);
//   - el de los metadatos, con su clase, su dirección y su instante, sin emitir
//     el texto (US1, escenario 6; contrato errores-y-codigos, fila 18; SC-012);
//   - y nada de lo que falla se escribe (FR-093).
//
// Todo resultado, de éxito o de fallo, lleva la fuente del BOE y una dirección
// de www.boe.es (FR-002, FR-101).
func TestArticulo(t *testing.T) {
	t.Parallel()

	t.Run("bloque-vigente", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
		consulta := ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21}
		esperado := resultadoResuelto(direccionDelArticulo21, banco.reloj.ahora(), articuloDelArticulo21(t))

		resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

		compruebaResuelta(t, resultado, err, esperado)
		banco.compruebaPeticiones(t, peticionDelBloque(direccionDelArticulo21), peticionDeLosMetadatos(metadatosVigente))

		// Los metadatos que pidió quedan en su entrada: metadatos de la misma
		// norma no pide nada (FR-090; US2, escenario 7).
		resultado, err = banco.resuelve(t, schema.Contexto{Offline: true}, ConsultaMetadatos{Norma: normaVigente})

		compruebaResuelta(t, resultado, err,
			resultadoResuelto(metadatosVigente, banco.reloj.ahora(), metadatosDeLaNormaVigente()))

		// Caducados ya los metadatos, la entrada del artículo se sigue sirviendo
		// hasta el final de su propia vigencia.
		banco.reloj.adelanta(vigenciaDeLosArticulos - time.Nanosecond)

		for _, ec := range []schema.Contexto{{}, {Offline: true}, {DryRun: true}, {Offline: true, DryRun: true}} {
			resultado, err := banco.resuelve(t, ec, consulta)

			require.NoError(t, err, "con %+v", ec)
			assert.Equal(t, esperado, resultado, "con %+v", ec)
		}

		banco.compruebaPeticiones(t, peticionDelBloque(direccionDelArticulo21), peticionDeLosMetadatos(metadatosVigente))
		assert.Equal(t, 1, banco.construcciones, "lo servido de la caché no construye ningún cliente")

		banco.reloj.adelanta(time.Nanosecond)

		resultado, err = banco.resuelve(t, schema.Contexto{Offline: true}, consulta)

		compruebaFalloDeLaConsulta(t, resultado, err, falloSinEntradaDelArticulo(direccionDelArticulo21))
	})

	resueltos := []struct {
		nombre    string
		pedidor   func(*testing.T, *relojDePrueba) *pedidorDePrueba
		consulta  ConsultaArticulo
		direccion string
		metadatos string
		datos     func(*testing.T) Articulo
	}{
		{
			// La Ley 30/1992: derogada y con la vigencia agotada, y un artículo
			// con varias versiones, del que se toma la última (US1, escenarios 2
			// y 3).
			nombre:    "derogada",
			pedidor:   reproduce(carpetaDeLasGrabaciones),
			consulta:  ConsultaArticulo{Norma: normaDerogada, Bloque: bloqueDelArticulo42},
			direccion: direccionDelArticulo42,
			metadatos: metadatosDerogada,
			datos:     articuloDelArticulo42,
		},
		{
			nombre:    "tres-avisos",
			pedidor:   reproduce(sinteticoDeLosAvisos),
			consulta:  ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21},
			direccion: direccionDelArticulo21,
			metadatos: metadatosVigente,
			datos:     articuloConTresAvisos,
		},
	}

	for _, caso := range resueltos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, caso.pedidor)
			esperado := resultadoResuelto(caso.direccion, banco.reloj.ahora(), caso.datos(t))

			resultado, err := banco.resuelve(t, schema.Contexto{}, caso.consulta)

			compruebaResuelta(t, resultado, err, esperado)
			banco.compruebaPeticiones(t, peticionDelBloque(caso.direccion), peticionDeLosMetadatos(caso.metadatos))

			banco.reloj.adelanta(time.Second)

			resultado, err = banco.resuelve(t, schema.Contexto{Offline: true}, caso.consulta)

			compruebaResuelta(t, resultado, err, esperado)
		})
	}

	t.Run("metadatos-en-cache", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))

		_, err := banco.resuelve(t, schema.Contexto{}, ConsultaMetadatos{Norma: normaVigente})
		require.NoError(t, err)

		esperado := resultadoResuelto(direccionDelArticulo21, banco.reloj.ahora(), articuloDelArticulo21(t))

		resultado, err := banco.resuelve(t, schema.Contexto{}, ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21})

		compruebaResuelta(t, resultado, err, esperado)
		banco.compruebaPeticiones(t, peticionDeLosMetadatos(metadatosVigente), peticionDelBloque(direccionDelArticulo21))
	})

	fallidos := []struct {
		nombre    string
		pedidor   func(*testing.T, *relojDePrueba) *pedidorDePrueba
		bloque    string
		direccion string
		// pedidas son las peticiones que la fuente entrega al cliente antes de
		// fallar.
		pedidas []httpx.Peticion
		// fallo es el esperado salvo el instante, que es el de la petición que
		// falla.
		fallo falloDeLaConsulta
		// deHTTPX es la clase del error de httpx que el fallo envuelve, cuyo
		// texto sigue en el mensaje detrás del de la fuente (contrato
		// errores-y-codigos §2); vacía si no envuelve ninguno.
		deHTTPX schema.Clase
	}{
		{
			nombre:    "bloque-inexistente",
			pedidor:   reproduce(carpetaDeLasGrabaciones),
			bloque:    bloqueInexistente,
			direccion: direccionDelBloqueInexistente,
			pedidas:   []httpx.Peticion{peticionDelBloque(direccionDelBloqueInexistente)},
			fallo: falloDeLaConsulta{
				direccion: direccionDelBloqueInexistente,
				clase:     schema.ClaseNoEncontrado,
				codigo:    3,
				mensaje:   "la norma BOE-A-2015-10565 no tiene el bloque a9999 (" + direccionDelBloqueInexistente + ")",
			},
		},
		{
			nombre:    "bloque-ilegible",
			pedidor:   reproduce(sinteticoDelBloqueIlegible),
			bloque:    bloqueDelArticulo21,
			direccion: direccionDelArticulo21,
			pedidas:   []httpx.Peticion{peticionDelBloque(direccionDelArticulo21)},
			fallo:     falloAlInterpretarElBloque("el cuerpo no es XML legible"),
		},
		{
			nombre:    "bloque-sin-elemento",
			pedidor:   reproduce(sinteticoDelBloqueSinElemento),
			bloque:    bloqueDelArticulo21,
			direccion: direccionDelArticulo21,
			pedidas:   []httpx.Peticion{peticionDelBloque(direccionDelArticulo21)},
			fallo:     falloAlInterpretarElBloque("el XML no tiene ningún elemento bloque por debajo de la raíz"),
		},
		{
			nombre:    "metadatos-caidos",
			pedidor:   reproduce(sinteticoDeLosMetadatosCaidos),
			bloque:    bloqueDelArticulo21,
			direccion: direccionDelArticulo21,
			pedidas:   []httpx.Peticion{peticionDelBloque(direccionDelArticulo21), peticionDeLosMetadatos(metadatosVigente)},
			fallo: falloSinVigenciaDelArticulo21(falloDeLaConsulta{
				direccion: metadatosVigente,
				clase:     schema.ClaseFuenteNoDisponible,
				codigo:    4,
				mensaje:   "ha fallado la petición de los metadatos de la norma BOE-A-2015-10565 (" + metadatosVigente + ")",
			}),
			deHTTPX: schema.ClaseFuenteNoDisponible,
		},
		{
			nombre:    "metadatos-ilegibles",
			pedidor:   reproduce(sinteticoDeLosMetadatosIlegibles),
			bloque:    bloqueDelArticulo21,
			direccion: direccionDelArticulo21,
			pedidas:   []httpx.Peticion{peticionDelBloque(direccionDelArticulo21), peticionDeLosMetadatos(metadatosVigente)},
			fallo:     falloSinVigenciaDelArticulo21(falloAlInterpretar("el cuerpo no es JSON legible")),
		},
	}

	for _, caso := range fallidos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, caso.pedidor)
			consulta := ConsultaArticulo{Norma: normaVigente, Bloque: caso.bloque}
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
			banco.compruebaPeticiones(t, caso.pedidas...)

			// Nada de lo que falló quedó escrito: ni el artículo ni los metadatos.
			resultado, err = banco.resuelve(t, schema.Contexto{Offline: true}, consulta)

			compruebaFalloDeLaConsulta(t, resultado, err, falloSinEntradaDelArticulo(caso.direccion))

			resultado, err = banco.resuelve(t, schema.Contexto{Offline: true}, ConsultaMetadatos{Norma: normaVigente})

			compruebaFalloDeLaConsulta(t, resultado, err, falloSinEntradaConOffline(metadatosVigente))
			banco.compruebaPeticiones(t, caso.pedidas...)
		})
	}

	t.Run("ensayo-sin-entradas", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(t.TempDir()))

		resultado, err := banco.resuelve(t, schema.Contexto{DryRun: true},
			ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21})

		require.NoError(t, err)
		compruebaDireccionDeLaFuente(t, resultado.Procedencia.URL)
		assert.Equal(t, schema.Resultado{
			Procedencia: schema.Procedencia{Fuente: NombreDeLaFuente, URL: direccionDelArticulo21},
			Ensayo:      []string{"GET " + direccionDelArticulo21, "GET " + metadatosVigente},
		}, resultado)
		banco.compruebaPeticiones(t, peticionDelBloque(direccionDelArticulo21), peticionDeLosMetadatos(metadatosVigente))

		for indice, respuesta := range banco.pedidor.respuestas {
			assert.True(t, respuesta.Ensayo, "bajo --dry-run no se emite la petición %d", indice+1)
		}

		banco.compruebaCacheSinCrear(t)
	})

	t.Run("ensayo-con-los-metadatos-en-cache", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
		consulta := ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21}

		_, err := banco.resuelve(t, schema.Contexto{}, ConsultaMetadatos{Norma: normaVigente})
		require.NoError(t, err)

		resultado, err := banco.resuelve(t, schema.Contexto{DryRun: true}, consulta)

		require.NoError(t, err)
		assert.Equal(t, schema.Resultado{
			Procedencia: schema.Procedencia{Fuente: NombreDeLaFuente, URL: direccionDelArticulo21},
			Ensayo:      []string{"GET " + direccionDelArticulo21},
		}, resultado)
		banco.compruebaPeticiones(t, peticionDeLosMetadatos(metadatosVigente), peticionDelBloque(direccionDelArticulo21))
		assert.True(t, banco.pedidor.respuestas[1].Ensayo, "bajo --dry-run no se emite la petición del bloque")

		// El ensayo no escribió el artículo, y con --offline falta su entrada,
		// aunque la de los metadatos esté (FR-101).
		resultado, err = banco.resuelve(t, schema.Contexto{Offline: true}, consulta)

		compruebaFalloDeLaConsulta(t, resultado, err, falloSinEntradaDelArticulo(direccionDelArticulo21))
	})

	for nombre, ec := range map[string]schema.Contexto{
		"offline-sin-entradas":          {Offline: true},
		"offline-y-ensayo-sin-entradas": {Offline: true, DryRun: true},
	} {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, reproduce(t.TempDir()))

			resultado, err := banco.resuelve(t, ec, ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21})

			compruebaFalloDeLaConsulta(t, resultado, err, falloSinEntradaDelArticulo(direccionDelArticulo21))
			assert.Zero(t, banco.construcciones, "con --offline no se construye ningún cliente")
			banco.compruebaCacheSinCrear(t)
		})
	}

	malFormadas := []struct {
		nombre   string
		consulta ConsultaArticulo
		mensaje  string
	}{
		{
			nombre:   "norma-invalida",
			consulta: ConsultaArticulo{Norma: "BOE-A-2015-1056a", Bloque: bloqueDelArticulo21},
			mensaje: `la norma "BOE-A-2015-1056a" no tiene la forma BOE-A-<año>-<número>, ` +
				"con cuatro dígitos en el año y de uno a nueve en el número",
		},
		{
			nombre:   "bloque-invalido",
			consulta: ConsultaArticulo{Norma: normaVigente, Bloque: "../a21"},
			mensaje: `el bloque "../a21" no tiene la forma de un id de bloque: de 1 a 64 caracteres, ` +
				"el primero letra o dígito ASCII y los demás letras, dígitos, guiones o puntos, " +
				"como a21, da3, preambulo, a1-30 o a85bis.",
		},
		{
			// La norma se valida antes que el bloque.
			nombre:   "norma-y-bloque-invalidos",
			consulta: ConsultaArticulo{Norma: "BOE-A-2015", Bloque: "../a21"},
			mensaje: `la norma "BOE-A-2015" no tiene la forma BOE-A-<año>-<número>, ` +
				"con cuatro dígitos en el año y de uno a nueve en el número",
		},
	}

	for _, caso := range malFormadas {
		for sufijo, ec := range map[string]schema.Contexto{"": {}, "-offline": {Offline: true}, "-en-ensayo": {DryRun: true}} {
			t.Run(caso.nombre+sufijo, func(t *testing.T) {
				t.Parallel()

				dependencias := &dependenciasDePrueba{cacheAbierta: &cacheDePrueba{}}
				fuente := fuenteDePrueba(t, dependencias)

				resultado, err := fuente.Fetch(t.Context(), ec, caso.consulta)

				assert.Zero(t, resultado, "el sobre de los argumentos lo firma y lo fecha el kernel")

				var fallo *Error
				require.ErrorAs(t, err, &fallo)
				assert.Empty(t, fallo.URL)
				assert.Zero(t, fallo.Instante)
				assert.Equal(t, schema.ClaseArgumentos, cli.Clasificar(err))
				assert.Equal(t, 2, cli.CodigoSalida(err))
				assert.Equal(t, caso.mensaje, err.Error())
				dependencias.compruebaSinUso(t)
			})
		}
	}
}

// TestFechaDeConsultaDeArticulo fija la fecha de consulta de articulo (FR-096;
// US2, escenario 8; data-model.md §6): la más antigua de las dos consultas que
// sostienen el artículo, la de su bloque y la de los metadatos de los que salen
// sus avisos, con el reloj de la prueba gobernando la hora de emisión de cada
// petición y el «ahora» de la caché. La entrada del artículo la guarda, y
// servirla da la misma; la de los metadatos guarda la de su propia petición.
func TestFechaDeConsultaDeArticulo(t *testing.T) {
	t.Parallel()

	consulta := ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21}

	t.Run("metadatos-guardados-antes-que-el-bloque", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
		t0 := banco.reloj.ahora()

		_, err := banco.resuelve(t, schema.Contexto{}, ConsultaMetadatos{Norma: normaVigente})
		require.NoError(t, err)

		banco.reloj.adelanta(vigenciaDeLosMetadatos / 2)
		t1 := banco.reloj.ahora()
		esperado := resultadoResuelto(direccionDelArticulo21, t0, articuloDelArticulo21(t))

		resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

		compruebaResuelta(t, resultado, err, esperado)
		banco.compruebaPeticiones(t, peticionDeLosMetadatos(metadatosVigente), peticionDelBloque(direccionDelArticulo21))
		assert.Equal(t, t1, banco.pedidor.respuestas[1].Instante, "el bloque se pide en t1")

		// Servido de su entrada, ya caducados los metadatos, sigue siendo t0.
		banco.reloj.adelanta(vigenciaDeLosMetadatos)

		for _, ec := range []schema.Contexto{{}, {Offline: true}} {
			resultado, err := banco.resuelve(t, ec, consulta)

			require.NoError(t, err, "con %+v", ec)
			assert.Equal(t, esperado, resultado, "con %+v", ec)
		}

		banco.compruebaPeticiones(t, peticionDeLosMetadatos(metadatosVigente), peticionDelBloque(direccionDelArticulo21))
	})

	t.Run("bloque-pedido-antes-que-los-metadatos", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduceYAdelanta(carpetaDeLasGrabaciones, time.Second))
		t0 := banco.reloj.ahora()
		esperado := resultadoResuelto(direccionDelArticulo21, t0, articuloDelArticulo21(t))

		resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

		compruebaResuelta(t, resultado, err, esperado)
		banco.compruebaPeticiones(t, peticionDelBloque(direccionDelArticulo21), peticionDeLosMetadatos(metadatosVigente))
		require.Len(t, banco.pedidor.respuestas, 2)
		assert.Equal(t, t0, banco.pedidor.respuestas[0].Instante, "el bloque se pide en t0")
		assert.Equal(t, t0.Add(time.Second), banco.pedidor.respuestas[1].Instante, "los metadatos se piden en t0 + 1 s")

		// La entrada de los metadatos guarda el instante de su petición, y la del
		// artículo, t0.
		resultado, err = banco.resuelve(t, schema.Contexto{Offline: true}, ConsultaMetadatos{Norma: normaVigente})

		compruebaResuelta(t, resultado, err,
			resultadoResuelto(metadatosVigente, t0.Add(time.Second), metadatosDeLaNormaVigente()))

		resultado, err = banco.resuelve(t, schema.Contexto{Offline: true}, consulta)

		compruebaResuelta(t, resultado, err, esperado)
	})
}

// TestArticuloCoincideConBoePy es el diff de aceptación del hito (FR-116, SC-001;
// contrato esquemas-fixtures-y-controles §5): para cada referencia escrita desde
// refs/boe.py sobre las grabaciones, la fuente, sobre la reproducción de esas
// mismas grabaciones y con una caché vacía, resuelve articulo, y su data,
// proyectado sobre los ocho campos del diff con camposComparados, tiene que
// coincidir con la referencia campo a campo. Un campo distinto hace fallar el
// subtest nombrándolo. Ante una discrepancia se corrige el código, nunca la
// referencia, que el ejecutor no escribe.
func TestArticuloCoincideConBoePy(t *testing.T) {
	t.Parallel()

	articulos, err := articulosDeLasReferencias(carpetaDeLasReferencias)
	require.NoError(t, err)
	require.Len(t, articulos, referenciasDelDiff)

	for _, articulo := range articulos {
		t.Run(strings.TrimSuffix(articulo.fichero(), ".json"), func(t *testing.T) {
			t.Parallel()

			referencia := leerReferencia(t, articulo)

			fuente, err := Nueva(
				ConCliente(construirReproduccion(carpetaDeLasGrabaciones)),
				ConCache(abrirCacheEn(t.TempDir())),
			)
			require.NoError(t, err)

			resultado, err := fuente.Fetch(t.Context(), schema.Contexto{},
				ConsultaArticulo{Norma: articulo.norma, Bloque: articulo.bloque})
			require.NoError(t, err)

			datos, esArticulo := resultado.Datos.(Articulo)
			require.Truef(t, esArticulo, "el data de articulo es de tipo %T", resultado.Datos)

			for _, campo := range camposComparados(*referencia.Campos, datos) {
				assert.Equalf(t, campo.referencia, campo.data,
					"el campo %s de data no coincide con su referencia (refs/boe.py: %s)", campo.nombre, campo.boePy)
			}
		})
	}
}

// TestCamposComparados demuestra que la proyección de
// TestArticuloCoincideConBoePy no compara en vacío (FR-116): da los ocho campos
// del diff en el orden de la referencia; cambiar en data uno de ellos hace
// distinto ese campo y solo ese; y cambiar lo que queda fuera del diff —la norma,
// el bloque, la huella, el ELI o el código de un aviso— no hace distinto ninguno.
func TestCamposComparados(t *testing.T) {
	t.Parallel()

	referencia := referenciaDeLosCampos(articuloDeLaProyeccion())

	t.Run("los-ocho-campos", func(t *testing.T) {
		t.Parallel()

		nombres := make([]string, 0, len(camposComparados(referencia, articuloDeLaProyeccion())))
		for _, campo := range camposComparados(referencia, articuloDeLaProyeccion()) {
			nombres = append(nombres, campo.nombre)
		}

		assert.Equal(t, []string{
			"titulo", "tipo", "fecha_version", "fecha_vigencia", "norma_modificadora", "texto", "avisos", "url",
		}, nombres)
	})

	casos := []struct {
		nombre string
		cambia func(datos *Articulo)
		// distinto es el campo que el cambio hace distinto; vacío, ninguno.
		distinto string
	}{
		{nombre: "sin-cambios"},
		{nombre: "titulo", cambia: func(datos *Articulo) { datos.Titulo += "." }, distinto: "titulo"},
		{nombre: "tipo", cambia: func(datos *Articulo) { datos.Tipo = "articulo" }, distinto: "tipo"},
		{nombre: "fecha-version", cambia: func(datos *Articulo) { datos.FechaVersion = "original" }, distinto: "fecha_version"},
		{nombre: "fecha-vigencia", cambia: func(datos *Articulo) { datos.FechaVigencia = "" }, distinto: "fecha_vigencia"},
		{
			nombre:   "norma-modificadora",
			cambia:   func(datos *Articulo) { datos.NormaModificadora = normaDerogada },
			distinto: "norma_modificadora",
		},
		{nombre: "texto", cambia: func(datos *Articulo) { datos.Texto += "\n" }, distinto: "texto"},
		{nombre: "un-aviso-de-menos", cambia: func(datos *Articulo) { datos.Avisos = datos.Avisos[:1] }, distinto: "avisos"},
		{
			nombre: "avisos-en-otro-orden",
			cambia: func(datos *Articulo) {
				datos.Avisos[0], datos.Avisos[1] = datos.Avisos[1], datos.Avisos[0]
			},
			distinto: "avisos",
		},
		{nombre: "url", cambia: func(datos *Articulo) { datos.URL += "x" }, distinto: "url"},
		{nombre: "norma", cambia: func(datos *Articulo) { datos.Norma = normaVigente }},
		{nombre: "bloque", cambia: func(datos *Articulo) { datos.Bloque = bloqueDelArticulo21 }},
		{nombre: "hash-texto", cambia: func(datos *Articulo) { datos.HashTexto = huellaEsperada("") }},
		{nombre: "url-eli", cambia: func(datos *Articulo) { datos.URLELI = "" }},
		{nombre: "codigo-de-un-aviso", cambia: func(datos *Articulo) { datos.Avisos[0].Codigo = "vigencia-agotada" }},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			datos := articuloDeLaProyeccion()
			if caso.cambia != nil {
				caso.cambia(&datos)
			}

			var distintos []string

			for _, campo := range camposComparados(referencia, datos) {
				if !assert.ObjectsAreEqual(campo.referencia, campo.data) {
					distintos = append(distintos, campo.nombre)
				}
			}

			if caso.distinto == "" {
				assert.Empty(t, distintos)

				return
			}

			assert.Equal(t, []string{caso.distinto}, distintos)
		})
	}
}

// campoComparado es un campo del diff de aceptación: su nombre, su valor en la
// referencia, el que da data y las líneas de refs/boe.py que lo justifican.
type campoComparado struct {
	nombre     string
	referencia any
	data       any
	boePy      string
}

// camposComparados proyecta el data de articulo sobre los ocho campos de FR-116,
// en el orden de la referencia, cada uno junto a su valor en ella: los avisos,
// como la lista ordenada de sus textos y sin su código, que refs/boe.py no
// produce, igual que la norma, el bloque, la huella y el ELI, que quedan fuera.
// La referencia tiene que traer los ocho campos con valor
// (camposDelDiff.comprobar).
func camposComparados(referencia camposDelDiff, datos Articulo) []campoComparado {
	textos := make([]string, 0, len(datos.Avisos))
	for _, aviso := range datos.Avisos {
		textos = append(textos, aviso.Texto)
	}

	return []campoComparado{
		{nombre: "titulo", referencia: *referencia.Titulo.Valor, data: datos.Titulo, boePy: referencia.Titulo.BoePy},
		{nombre: "tipo", referencia: *referencia.Tipo.Valor, data: datos.Tipo, boePy: referencia.Tipo.BoePy},
		{
			nombre:     "fecha_version",
			referencia: *referencia.FechaVersion.Valor,
			data:       datos.FechaVersion,
			boePy:      referencia.FechaVersion.BoePy,
		},
		{
			nombre:     "fecha_vigencia",
			referencia: *referencia.FechaVigencia.Valor,
			data:       datos.FechaVigencia,
			boePy:      referencia.FechaVigencia.BoePy,
		},
		{
			nombre:     "norma_modificadora",
			referencia: *referencia.NormaModificadora.Valor,
			data:       datos.NormaModificadora,
			boePy:      referencia.NormaModificadora.BoePy,
		},
		{nombre: "texto", referencia: *referencia.Texto.Valor, data: datos.Texto, boePy: referencia.Texto.BoePy},
		{nombre: "avisos", referencia: *referencia.Avisos.Valor, data: textos, boePy: referencia.Avisos.BoePy},
		{nombre: "url", referencia: *referencia.URL.Valor, data: datos.URL, boePy: referencia.URL.BoePy},
	}
}

// articuloDeLaProyeccion es un artículo con los doce campos distintos entre sí y
// dos avisos, para las pruebas de la proyección. Cada llamada da uno nuevo, cuyos
// avisos se pueden cambiar sin cambiar los de nadie más.
func articuloDeLaProyeccion() Articulo {
	const texto = "Artículo 42. Obligación de resolver."

	return Articulo{
		Norma:             normaDerogada,
		Bloque:            bloqueDelArticulo42,
		Titulo:            "Artículo 42",
		Tipo:              "precepto",
		FechaVersion:      "19990114",
		FechaVigencia:     "19990414",
		NormaModificadora: "BOE-A-1999-847",
		Texto:             texto,
		HashTexto:         huellaEsperada(texto),
		Avisos: []Aviso{
			{Codigo: "derogada", Texto: fraseEsperadaDeNormaDerogada},
			{Codigo: "vigencia-agotada", Texto: fraseEsperadaDeVigenciaAgotada},
		},
		URL:    "https://www.boe.es/buscar/act.php?id=BOE-A-1992-26318#a42",
		URLELI: "https://www.boe.es/eli/es/l/1992/11/26/30",
	}
}

// referenciaDeLosCampos son los ocho campos de una referencia que coincide con el
// artículo.
func referenciaDeLosCampos(datos Articulo) camposDelDiff {
	campo := func(valor string) *campoDelDiff[string] {
		return &campoDelDiff[string]{Valor: &valor, BoePy: "prueba"}
	}

	textos := make([]string, 0, len(datos.Avisos))
	for _, aviso := range datos.Avisos {
		textos = append(textos, aviso.Texto)
	}

	return camposDelDiff{
		Titulo:            campo(datos.Titulo),
		Tipo:              campo(datos.Tipo),
		FechaVersion:      campo(datos.FechaVersion),
		FechaVigencia:     campo(datos.FechaVigencia),
		NormaModificadora: campo(datos.NormaModificadora),
		Texto:             campo(datos.Texto),
		Avisos:            &campoDelDiff[[]string]{Valor: &textos, BoePy: "prueba"},
		URL:               campo(datos.URL),
	}
}

// articuloDelArticulo21 es el data de articulo BOE-A-2015-10565 a21 tal como lo
// dan sus grabaciones: la última versión del bloque, ningún aviso y el ELI de la
// Ley 39/2015. El título lleva, entre «Artículo» y el número, el espacio de no
// separación (U+00A0) con el que lo escribe la fuente, que la lectura conserva.
// El texto, que es largo, es el de su referencia, y su huella se calcula aquí
// sobre él.
func articuloDelArticulo21(t *testing.T) Articulo {
	t.Helper()

	texto := textoDeLaReferencia(t, articuloDelDiff{norma: normaVigente, bloque: bloqueDelArticulo21})

	return Articulo{
		Norma:             normaVigente,
		Bloque:            bloqueDelArticulo21,
		Titulo:            "Artículo\u00a021",
		Tipo:              "precepto",
		FechaVersion:      "20151002",
		FechaVigencia:     "20161002",
		NormaModificadora: "BOE-A-2015-10565",
		Texto:             texto,
		HashTexto:         huellaEsperada(texto),
		Avisos:            []Aviso{},
		URL:               "https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565#a21",
		URLELI:            "https://www.boe.es/eli/es/l/2015/10/01/39",
	}
}

// articuloDelArticulo42 es el data de articulo BOE-A-1992-26318 a42 tal como lo
// dan sus grabaciones: la última de sus versiones, la de la Ley 4/1999, los avisos
// de norma derogada y de vigencia agotada en su orden, y el ELI de la Ley
// 30/1992. El texto es el de su referencia.
func articuloDelArticulo42(t *testing.T) Articulo {
	t.Helper()

	articulo := articuloDeLaProyeccion()
	articulo.Texto = textoDeLaReferencia(t, articuloDelDiff{norma: normaDerogada, bloque: bloqueDelArticulo42})
	articulo.HashTexto = huellaEsperada(articulo.Texto)

	return articulo
}

// articuloConTresAvisos es el de BOE-A-2015-10565 a21 con los metadatos del
// sintético de los avisos: los tres, en el orden de _check_vigencia y con sus
// frases literales (refs/boe.py 197-211; FR-012).
func articuloConTresAvisos(t *testing.T) Articulo {
	t.Helper()

	articulo := articuloDelArticulo21(t)
	articulo.Avisos = metadatosConTresAvisos().Avisos

	return articulo
}

// leerReferencia lee la referencia del artículo, después de comprobar que tiene
// la forma del contrato, que es de ese artículo y que trae los ocho campos.
func leerReferencia(t *testing.T, articulo articuloDelDiff) referenciaDelDiff {
	t.Helper()

	require.NoError(t, comprobarReferencia(carpetaDeLasReferencias, carpetaDeLasGrabaciones, articulo))

	var referencia referenciaDelDiff
	require.NoError(t, leerDatoDePrueba(filepath.Join(carpetaDeLasReferencias, articulo.fichero()), &referencia))

	return referencia
}

// textoDeLaReferencia es el texto de la referencia del artículo, que no puede
// venir vacío.
func textoDeLaReferencia(t *testing.T, articulo articuloDelDiff) string {
	t.Helper()

	texto := *leerReferencia(t, articulo).Campos.Texto.Valor
	require.NotEmpty(t, texto)

	return texto
}

// huellaEsperada es el hash_texto de un texto, calculado aquí y no con
// huellaDelTexto: sha256: y el hexadecimal en minúsculas del SHA-256 de sus bytes
// UTF-8 (FR-015).
func huellaEsperada(texto string) string {
	suma := sha256.Sum256([]byte(texto))

	return "sha256:" + hex.EncodeToString(suma[:])
}

// falloSinEntradaDelArticulo es el fallo del artículo sin entrada vigente con
// --offline: «fuente no disponible», código 4, con la dirección del bloque —aunque
// la entrada de los metadatos esté—, sin instante y con un mensaje que nombra
// --offline, el verbo y la clave (contrato errores-y-codigos, fila 5; FR-101).
func falloSinEntradaDelArticulo(direccion string) falloDeLaConsulta {
	return falloDeLaConsulta{
		direccion: direccion,
		clase:     schema.ClaseFuenteNoDisponible,
		codigo:    4,
		mensaje: `con --offline no se pide nada a la fuente y no hay ninguna entrada vigente de articulo con la clave ` +
			`"boe.legislacion-consolidada|1|articulo|` + direccion + `" (` + direccion + ")",
	}
}

// falloAlInterpretarElBloque es el del bloque a21 cuya respuesta no se puede
// interpretar: «fuente no disponible», código 4, con lo que no se pudo
// interpretar detrás de la dirección y nunca el cuerpo (contrato
// errores-y-codigos, fila 16; FR-014).
func falloAlInterpretarElBloque(motivo string) falloDeLaConsulta {
	return falloDeLaConsulta{
		direccion: direccionDelArticulo21,
		clase:     schema.ClaseFuenteNoDisponible,
		codigo:    4,
		mensaje: "no se puede interpretar la respuesta a la petición del bloque a21 de la norma BOE-A-2015-10565 (" +
			direccionDelArticulo21 + "): " + motivo,
	}
}

// falloSinVigenciaDelArticulo21 es el de articulo BOE-A-2015-10565 a21 cuando el
// bloque se obtuvo y los metadatos fallan: el fallo de los metadatos, con su
// clase, su dirección y su instante, detrás del bloque obtenido y de que la
// vigencia no se pudo comprobar (contrato errores-y-codigos, fila 18; FR-013).
func falloSinVigenciaDelArticulo21(deLosMetadatos falloDeLaConsulta) falloDeLaConsulta {
	deLosMetadatos.mensaje = "el bloque a21 se obtuvo, pero no se pudo comprobar su vigencia: " + deLosMetadatos.mensaje

	return deLosMetadatos
}

// peticionDelBloque y peticionDeLosMetadatos son las peticiones con las que la
// fuente pide un bloque, en XML, y los metadatos de una norma, en JSON (FR-003;
// contrato verbos-y-salidas §3).
func peticionDelBloque(direccion string) httpx.Peticion {
	return httpx.Peticion{Metodo: "GET", URL: direccion, Acepta: aceptaDelBloque}
}

func peticionDeLosMetadatos(direccion string) httpx.Peticion {
	return httpx.Peticion{Metodo: "GET", URL: direccion, Acepta: aceptaDelResto}
}

// compruebaPeticiones cuenta las peticiones que la fuente ha entregado al
// cliente —también las de ensayo, que el cliente no emite— y exige que sean, en
// orden, exactamente esas.
func (b *bancoDeLaFuente) compruebaPeticiones(t *testing.T, pedidas ...httpx.Peticion) {
	t.Helper()

	require.Len(t, b.pedidor.peticiones, len(pedidas), "peticiones entregadas: %v", b.pedidor.peticiones)

	for indice, pedida := range pedidas {
		assert.Equal(t, pedida, b.pedidor.peticiones[indice], "la petición %d", indice+1)
	}
}

// reproduceYAdelanta es el Pedidor de prueba de reproduce que, al volver de cada
// petición, adelanta el reloj el tramo: dos peticiones de la misma invocación
// tienen así instantes de emisión distintos.
func reproduceYAdelanta(carpeta string, tramo time.Duration) func(*testing.T, *relojDePrueba) *pedidorDePrueba {
	return func(t *testing.T, reloj *relojDePrueba) *pedidorDePrueba {
		t.Helper()

		cliente := clienteDeReproduccion(t, carpeta, reloj.ahora)

		return &pedidorDePrueba{
			responde: func(ctx context.Context, ec schema.Contexto, peticion httpx.Peticion) (httpx.Respuesta, error) {
				respuesta, err := cliente.Pedir(ctx, ec, peticion)
				reloj.adelanta(tramo)

				return respuesta, err
			},
		}
	}
}
