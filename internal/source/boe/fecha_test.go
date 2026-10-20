package boe

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Lo que comparten las pruebas de la lectura fechada: las dos normas de lo
// grabado de H4 que tienen bloques con varias versiones, sus bloques y sus
// direcciones (contracts/articulo-fecha.md §2; research.md M1). Como en
// articulo_test.go, las direcciones y las claves van escritas enteras.
const (
	// normaDeLaLrbrl es la Ley 7/1985, de Bases del Régimen Local, y
	// bloqueDelArticulo22DeLaLrbrl, su artículo 22, con nueve versiones.
	normaDeLaLrbrl               = "BOE-A-1985-5392"
	bloqueDelArticulo22DeLaLrbrl = "a22"
	direccionDelA22DeLaLrbrl     = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-1985-5392/texto/bloque/a22"
	metadatosDeLaLrbrl           = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-1985-5392/metadatos"
	// normaDeLaLcsp es la Ley 9/2017, de Contratos del Sector Público, y
	// bloqueDelArticulo118DeLaLcsp, su artículo 118, con dos versiones (US1).
	normaDeLaLcsp                = "BOE-A-2017-12902"
	bloqueDelArticulo118DeLaLcsp = "a1-30"
	direccionDelA130DeLaLcsp     = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2017-12902/texto/bloque/a1-30"
	metadatosDeLaLcsp            = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2017-12902/metadatos"
	// goldenDelA22DeLaLrbrl es el golden de hoy del artículo 22 de la Ley 7/1985,
	// el de su última versión (FR-013).
	goldenDelA22DeLaLrbrl = "articulo-BOE-A-1985-5392-a22"
	// La redacción posterior de los bloques de lo grabado: la última versión del
	// art. 22 de la Ley 7/1985 y la del art. 118 de la Ley 9/2017.
	vigenciaDeLaUltimaDelA22     = "20151031"
	modificadoraDeLaUltimaDelA22 = "BOE-A-2015-11723"
)

// TestArticuloEnUnaFecha fija la lectura fechada de articulo sobre lo grabado de
// H4 (contracts/articulo-fecha.md §2 a §4; data-model.md §3 y §4; research.md D4
// a D7): la versión elegida es la de mayor fecha_vigencia que no es posterior al
// día pedido, a igualdad la última en el orden del bloque, y nunca por
// fecha_publicacion, también con las versiones desordenadas; antes de la menor
// vigencia, 3 con el mensaje de FR-003; la redacción posterior es la de mayor
// fecha_vigencia del bloque si no es la elegida; data lleva los doce
// campos de hoy de la versión elegida, con el título, el tipo, los avisos y las
// direcciones de la lectura sin fecha, más fecha_pedida; y la lectura pide los
// mismos dos recursos que la de hoy (FR-006). Un bloque sin versiones, o con
// alguna sin fecha_vigencia, da 3 con el mensaje de FR-004 sea cual sea el día, y
// los avisos son los de la norma hoy (FR-011). Ningún fallo de la elección lleva
// texto.
func TestArticuloEnUnaFecha(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre                  string
		norma, bloque           string
		direccion, metadatos    string
		fecha                   string
		fechaVersion, vigencia  string
		modificadora, hashTexto string
		posterior               *RedaccionPosterior
		// golden es el golden de hoy cuyo texto y huella da la versión elegida.
		golden string
		// mensaje es el del 3 de la elección; vacío, la lectura sale con 0.
		mensaje string
	}{
		{
			nombre: "anterior a la primera vigencia",
			norma:  normaDeLaLrbrl, bloque: bloqueDelArticulo22DeLaLrbrl,
			direccion: direccionDelA22DeLaLrbrl, metadatos: metadatosDeLaLrbrl,
			fecha: "1985-04-22",
			mensaje: "el bloque a22 de BOE-A-1985-5392 no tenía redacción vigente el 1985-04-22: " +
				"su primera redacción tiene fecha de vigencia 19850423 (" + direccionDelA22DeLaLrbrl + ")",
		},
		{
			nombre: "intermedia",
			norma:  normaDeLaLrbrl, bloque: bloqueDelArticulo22DeLaLrbrl,
			direccion: direccionDelA22DeLaLrbrl, metadatos: metadatosDeLaLrbrl,
			fecha:        "2005-06-15",
			fechaVersion: "20031217", vigencia: "20040101", modificadora: "BOE-A-2003-23103",
			hashTexto: "sha256:ce4210549efa0b8a103e99ccfd4137f6b66a5f300f661a94e2d2545b08847183",
			posterior: &RedaccionPosterior{FechaVigencia: vigenciaDeLaUltimaDelA22, NormaModificadora: modificadoraDeLaUltimaDelA22},
		},
		{
			nombre: "el día que entra en vigor",
			norma:  normaDeLaLrbrl, bloque: bloqueDelArticulo22DeLaLrbrl,
			direccion: direccionDelA22DeLaLrbrl, metadatos: metadatosDeLaLrbrl,
			fecha:        "2008-04-30",
			fechaVersion: "20071031", vigencia: "20080430", modificadora: "BOE-A-2007-18874",
			hashTexto: "sha256:4c8982b1fee3d8414a0b3fdf7d71e4b269b5a1ec43197edeae71a171e1d151fe",
			posterior: &RedaccionPosterior{FechaVigencia: vigenciaDeLaUltimaDelA22, NormaModificadora: modificadoraDeLaUltimaDelA22},
		},
		{
			// La séptima versión ya está publicada (20071031), pero entra en vigor
			// al día siguiente: rige la sexta.
			nombre: "la víspera, ya publicada la de ese día",
			norma:  normaDeLaLrbrl, bloque: bloqueDelArticulo22DeLaLrbrl,
			direccion: direccionDelA22DeLaLrbrl, metadatos: metadatosDeLaLrbrl,
			fecha:        "2008-04-29",
			fechaVersion: "20070529", vigencia: "20070701", modificadora: "BOE-A-2007-10701",
			hashTexto: "sha256:50a1cc5d30109aeb4747a003df8b6855ab23631c7f3fcf81b6e068f7a571343e",
			posterior: &RedaccionPosterior{FechaVigencia: vigenciaDeLaUltimaDelA22, NormaModificadora: modificadoraDeLaUltimaDelA22},
		},
		{
			nombre: "posterior a la última",
			norma:  normaDeLaLrbrl, bloque: bloqueDelArticulo22DeLaLrbrl,
			direccion: direccionDelA22DeLaLrbrl, metadatos: metadatosDeLaLrbrl,
			fecha:        "2019-06-01",
			fechaVersion: "20151031", vigencia: "20151031", modificadora: "BOE-A-2015-11723",
			hashTexto: "sha256:2a29d264c6a3749916785fdd84c817b6b0d83925e78048380f3903d469aab4ea",
			golden:    goldenDelA22DeLaLrbrl,
		},
		{
			// Publicado el 20151002, el art. 21 de la Ley 39/2015 entra en vigor el
			// 20161002: entre las dos fechas no tenía redacción vigente.
			nombre: "entre publicación y vigencia",
			norma:  normaVigente, bloque: bloqueDelArticulo21,
			direccion: direccionDelArticulo21, metadatos: metadatosVigente,
			fecha: "2016-06-01",
			mensaje: "el bloque a21 de BOE-A-2015-10565 no tenía redacción vigente el 2016-06-01: " +
				"su primera redacción tiene fecha de vigencia 20161002 (" + direccionDelArticulo21 + ")",
		},
		{
			nombre: "US1",
			norma:  normaDeLaLcsp, bloque: bloqueDelArticulo118DeLaLcsp,
			direccion: direccionDelA130DeLaLcsp, metadatos: metadatosDeLaLcsp,
			fecha:        "2019-06-01",
			fechaVersion: "20171109", vigencia: "20180309", modificadora: "BOE-A-2017-12902",
			hashTexto: "sha256:a461184434a16f37601a3e6b53e89042ce31682e45a9dd35fecd97733d55b8e5",
			posterior: &RedaccionPosterior{FechaVigencia: "20200206", NormaModificadora: "BOE-A-2020-1651"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
			fecha := caso.fecha

			resultado, err := banco.resuelve(t, schema.Contexto{},
				ConsultaArticulo{Norma: caso.norma, Bloque: caso.bloque, Fecha: &fecha})

			banco.compruebaPeticiones(t, peticionDelBloque(caso.direccion), peticionDeLosMetadatos(caso.metadatos))

			if caso.mensaje != "" {
				compruebaFalloDeLaConsulta(t, resultado, err, falloDeLaEleccion(caso.direccion, banco.reloj.ahora(), caso.mensaje))

				return
			}

			require.NoError(t, err)
			assert.Equal(t, schema.Procedencia{
				Fuente: NombreDeLaFuente, URL: caso.direccion, FechaConsulta: banco.reloj.ahora(),
			}, resultado.Procedencia)

			datos, esFechado := resultado.Datos.(ArticuloFechado)
			require.Truef(t, esFechado, "el data de la lectura fechada es de tipo %T", resultado.Datos)

			sinFecha := articuloSinFecha(t, caso.norma, caso.bloque)
			assert.Equal(t, ArticuloFechado{
				Articulo: Articulo{
					Norma:             caso.norma,
					Bloque:            caso.bloque,
					Titulo:            sinFecha.Titulo,
					Tipo:              sinFecha.Tipo,
					FechaVersion:      caso.fechaVersion,
					FechaVigencia:     caso.vigencia,
					NormaModificadora: caso.modificadora,
					Texto:             datos.Texto,
					HashTexto:         caso.hashTexto,
					Avisos:            sinFecha.Avisos,
					URL:               sinFecha.URL,
					URLELI:            sinFecha.URLELI,
				},
				FechaPedida:        caso.fecha,
				RedaccionPosterior: caso.posterior,
			}, datos)
			assert.Equal(t, huellaEsperada(datos.Texto), datos.HashTexto, "hash_texto es la huella de su texto")

			if caso.golden != "" {
				deHoy := articuloDelGolden(t, caso.golden)
				assert.Equal(t, deHoy.Texto, datos.Texto, "el texto es el del golden de hoy (FR-013)")
				assert.Equal(t, deHoy.HashTexto, datos.HashTexto, "la huella es la del golden de hoy (FR-013)")
			}
		})
	}

	t.Run("un bloque que no se puede situar", func(t *testing.T) {
		t.Parallel()

		// Los bloques construidos sustituyen al art. 21 de la Ley 39/2015, con
		// los metadatos grabados de su norma. Las versiones con fecha son del
		// 20100101 y del 20200101: los días pedidos van antes, entre y después.
		const (
			conFecha     = `<version id_norma="BOE-A-2010-1" fecha_publicacion="20091231" fecha_vigencia="20100101"><p>Primera.</p></version>`
			ultima       = `<version id_norma="BOE-A-2019-9" fecha_publicacion="20191231" fecha_vigencia="20200101"><p>Última.</p></version>`
			inicioDelA21 = `<bloque id="a21" tipo="precepto" titulo="Artículo 21">`
			finDelBloque = `</bloque>`
		)

		bloques := []struct {
			nombre string
			xml    string
		}{
			{nombre: "sin versiones", xml: inicioDelA21 + `<p>Todo el bloque.</p>` + finDelBloque},
			{
				nombre: "una versión sin fecha_vigencia",
				xml: inicioDelA21 + conFecha +
					`<version id_norma="BOE-A-2015-5" fecha_publicacion="20150101"><p>Sin vigencia.</p></version>` +
					ultima + finDelBloque,
			},
			{
				nombre: "una versión con fecha_vigencia vacía",
				xml: inicioDelA21 + conFecha +
					`<version id_norma="BOE-A-2015-5" fecha_publicacion="20150101" fecha_vigencia=""><p>Vacía.</p></version>` +
					ultima + finDelBloque,
			},
			{
				nombre: "una versión con una fecha_vigencia que no es un día",
				xml: inicioDelA21 + conFecha +
					`<version id_norma="BOE-A-2015-5" fecha_publicacion="20150101" fecha_vigencia="20150230"><p>Sin día.</p></version>` +
					ultima + finDelBloque,
			},
			{
				nombre: "la primera versión sin fecha_vigencia",
				xml: inicioDelA21 +
					`<version id_norma="BOE-A-2005-5" fecha_publicacion="20050101"><p>Sin vigencia.</p></version>` +
					conFecha + ultima + finDelBloque,
			},
		}

		for _, bloque := range bloques {
			t.Run(bloque.nombre, func(t *testing.T) {
				t.Parallel()

				banco := nuevoBanco(t,
					salvoEn(direccionDelArticulo21, respondeConElCuerpo(respuestaConBloque(bloque.xml))))

				for _, fecha := range []string{"2000-01-01", "2010-01-01", "2015-06-01", "2020-01-01", "2025-01-01"} {
					resultado, err := banco.resuelve(t, schema.Contexto{},
						ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21, Fecha: &fecha})

					compruebaFalloDeLaConsulta(t, resultado, err, falloDeLaEleccion(direccionDelArticulo21, banco.reloj.ahora(),
						"el BOE no da la fecha de vigencia de cada redacción del bloque a21 de BOE-A-2015-10565: "+
							"no se puede saber cuál tenía el "+fecha+" ("+direccionDelArticulo21+")"))
				}
			})
		}
	})

	t.Run("un bloque con las versiones desordenadas", func(t *testing.T) {
		t.Parallel()

		// El BOE da las versiones en orden creciente, pero la elección no depende
		// de ese orden (FR-002): la del 20200101 va repetida, con la de
		// BOE-A-2019-9 después en el orden, y la última en el orden es la del
		// 20100101, la que da la lectura sin fecha (nota [5]).
		const bloque = `<bloque id="a21" tipo="precepto" titulo="Artículo 21">` +
			`<version id_norma="BOE-A-2015-1" fecha_publicacion="20141231" fecha_vigencia="20150101"><p>De 2015.</p></version>` +
			`<version id_norma="BOE-A-2019-8" fecha_publicacion="20191230" fecha_vigencia="20200101"><p>De 2020, antes.</p></version>` +
			`<version id_norma="BOE-A-2019-9" fecha_publicacion="20191231" fecha_vigencia="20200101"><p>De 2020, después.</p></version>` +
			`<version id_norma="BOE-A-2010-1" fecha_publicacion="20091231" fecha_vigencia="20100101"><p>De 2010.</p></version>` +
			`</bloque>`

		deMayorVigencia := &RedaccionPosterior{FechaVigencia: "20200101", NormaModificadora: "BOE-A-2019-9"}

		casos := []struct {
			fecha, vigencia, modificadora string
			posterior                     *RedaccionPosterior
		}{
			{fecha: "2012-06-01", vigencia: "20100101", modificadora: "BOE-A-2010-1", posterior: deMayorVigencia},
			{fecha: "2016-06-01", vigencia: "20150101", modificadora: "BOE-A-2015-1", posterior: deMayorVigencia},
			{fecha: "2020-01-01", vigencia: "20200101", modificadora: "BOE-A-2019-9"},
			{fecha: "2025-01-01", vigencia: "20200101", modificadora: "BOE-A-2019-9"},
		}

		for _, caso := range casos {
			t.Run(caso.fecha, func(t *testing.T) {
				t.Parallel()

				banco := nuevoBanco(t, salvoEn(direccionDelArticulo21, respondeConElCuerpo(respuestaConBloque(bloque))))
				fecha := caso.fecha

				resultado, err := banco.resuelve(t, schema.Contexto{},
					ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21, Fecha: &fecha})
				require.NoError(t, err)

				datos, esFechado := resultado.Datos.(ArticuloFechado)
				require.Truef(t, esFechado, "el data de la lectura fechada es de tipo %T", resultado.Datos)
				assert.Equal(t, caso.vigencia, datos.FechaVigencia)
				assert.Equal(t, caso.modificadora, datos.NormaModificadora)
				assert.Equal(t, caso.posterior, datos.RedaccionPosterior)
			})
		}

		t.Run("anterior a la menor vigencia", func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, salvoEn(direccionDelArticulo21, respondeConElCuerpo(respuestaConBloque(bloque))))
			fecha := "2005-06-01"

			resultado, err := banco.resuelve(t, schema.Contexto{},
				ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21, Fecha: &fecha})

			compruebaFalloDeLaConsulta(t, resultado, err, falloDeLaEleccion(direccionDelArticulo21, banco.reloj.ahora(),
				"el bloque a21 de BOE-A-2015-10565 no tenía redacción vigente el 2005-06-01: "+
					"su primera redacción tiene fecha de vigencia 20100101 ("+direccionDelArticulo21+")"))
		})

		t.Run("la lectura sin fecha sigue dando la última en el orden", func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, salvoEn(direccionDelArticulo21, respondeConElCuerpo(respuestaConBloque(bloque))))

			resultado, err := banco.resuelve(t, schema.Contexto{},
				ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21})
			require.NoError(t, err)

			datos, esArticulo := resultado.Datos.(Articulo)
			require.Truef(t, esArticulo, "el data de la lectura sin fecha es de tipo %T", resultado.Datos)
			assert.Equal(t, "BOE-A-2010-1", datos.NormaModificadora)
		})
	})

	t.Run("los avisos son los de la norma hoy", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(sinteticoDeLosAvisos))
		fecha := "2019-06-01"

		resultado, err := banco.resuelve(t, schema.Contexto{},
			ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21, Fecha: &fecha})

		require.NoError(t, err)
		assert.Equal(t, ArticuloFechado{Articulo: articuloConTresAvisos(t), FechaPedida: fecha}, resultado.Datos)
		banco.compruebaPeticiones(t, peticionDelBloque(direccionDelArticulo21), peticionDeLosMetadatos(metadatosVigente))
	})
}

// TestArticuloEnUnaFechaYLaCache fija la caché de la lectura fechada
// (contracts/articulo-fecha.md §5; data-model.md §5; research.md D8): se sirve
// solo de la entrada de las versiones del bloque y nunca de la de hoy, que no lee
// ni escribe, de modo que, con solo la de hoy, la fechada con --offline sale con 4
// (FR-020) y, tras una fechada de una versión que no es la de hoy, la entrada de
// hoy queda igual byte a byte y la lectura sin fecha con --offline da lo mismo
// (FR-021); y lo que guarda una fechada sirve con --offline, con el mismo data, el
// mismo hash y la misma fecha de consulta, esa fecha y cualquier otra del bloque
// (FR-022, FR-087; SC-007). Bajo --dry-run, las líneas de las peticiones de la
// lectura sin fecha, sin escribir nada.
func TestArticuloEnUnaFechaYLaCache(t *testing.T) {
	t.Parallel()

	const (
		claveDeHoy = "boe.legislacion-consolidada|1|articulo|" + direccionDelA22DeLaLrbrl
		intermedia = "2005-06-15"
		otraFecha  = "2008-04-30"
	)

	fechada := func(fecha string) ConsultaArticulo {
		return ConsultaArticulo{Norma: normaDeLaLrbrl, Bloque: bloqueDelArticulo22DeLaLrbrl, Fecha: &fecha}
	}
	sinFecha := ConsultaArticulo{Norma: normaDeLaLrbrl, Bloque: bloqueDelArticulo22DeLaLrbrl}

	t.Run("la entrada de hoy no sirve ni cambia con la lectura fechada", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))

		_, err := banco.resuelve(t, schema.Contexto{}, sinFecha)
		require.NoError(t, err)

		deHoy := banco.contenidoGuardado(t, claveDeHoy)
		hoySinRed, err := banco.resuelve(t, schema.Contexto{Offline: true}, sinFecha)
		require.NoError(t, err)

		// Con solo la entrada de hoy y la de los metadatos, la fechada con
		// --offline no tiene de qué servirse (FR-020).
		resultado, err := banco.resuelve(t, schema.Contexto{Offline: true}, fechada(intermedia))
		compruebaFalloDeLaConsulta(t, resultado, err, falloSinEntradaDeLasVersiones(direccionDelA22DeLaLrbrl))

		// La fechada vuelve a pedir el bloque: no se sirve de la entrada de hoy.
		guardada, err := banco.resuelve(t, schema.Contexto{}, fechada(intermedia))
		require.NoError(t, err)
		assert.Equal(t, "20040101", vigenciaElegida(t, guardada), "la versión elegida no es la de hoy, la del 20151031")
		banco.compruebaPeticiones(t, peticionDelBloque(direccionDelA22DeLaLrbrl), peticionDeLosMetadatos(metadatosDeLaLrbrl),
			peticionDelBloque(direccionDelA22DeLaLrbrl))

		// La entrada de hoy sigue igual byte a byte, y la lectura sin fecha con
		// --offline da lo mismo que antes (FR-021).
		assert.Equal(t, string(deHoy), string(banco.contenidoGuardado(t, claveDeHoy)))

		resultado, err = banco.resuelve(t, schema.Contexto{Offline: true}, sinFecha)
		require.NoError(t, err)
		assert.Equal(t, hoySinRed, resultado)

		// Lo guardado por la fechada la sirve con --offline, con el mismo data, el
		// mismo hash y la misma fecha de consulta (FR-022).
		banco.reloj.adelanta(time.Second)

		resultado, err = banco.resuelve(t, schema.Contexto{Offline: true}, fechada(intermedia))
		require.NoError(t, err)
		assert.Equal(t, guardada, resultado)
		compruebaMismoHash(t, guardada, resultado)

		// Y la de otra fecha del mismo bloque: lo mismo que esa lectura pedida a
		// la fuente con la fecha de consulta de lo guardado.
		otra := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
		pedida, err := otra.resuelve(t, schema.Contexto{}, fechada(otraFecha))
		require.NoError(t, err)

		resultado, err = banco.resuelve(t, schema.Contexto{Offline: true}, fechada(otraFecha))
		require.NoError(t, err)
		assert.Equal(t, pedida, resultado)
		assert.Equal(t, guardada.Procedencia.FechaConsulta, resultado.Procedencia.FechaConsulta)
		compruebaMismoHash(t, pedida, resultado)
		assert.Equal(t, "20080430", vigenciaElegida(t, resultado))

		banco.compruebaPeticiones(t, peticionDelBloque(direccionDelA22DeLaLrbrl), peticionDeLosMetadatos(metadatosDeLaLrbrl),
			peticionDelBloque(direccionDelA22DeLaLrbrl))
	})

	t.Run("la lectura fechada no escribe la entrada de hoy", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))

		_, err := banco.resuelve(t, schema.Contexto{}, fechada(intermedia))
		require.NoError(t, err)

		resultado, err := banco.resuelve(t, schema.Contexto{Offline: true}, sinFecha)
		compruebaFalloDeLaConsulta(t, resultado, err, falloSinEntradaDelArticulo(direccionDelA22DeLaLrbrl))
	})

	t.Run("bajo --dry-run, las peticiones de la lectura sin fecha y nada escrito", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))

		resultado, err := banco.resuelve(t, schema.Contexto{DryRun: true}, fechada(intermedia))

		require.NoError(t, err)
		assert.Equal(t, schema.Resultado{
			Procedencia: schema.Procedencia{Fuente: NombreDeLaFuente, URL: direccionDelA22DeLaLrbrl},
			Ensayo:      []string{"GET " + direccionDelA22DeLaLrbrl, "GET " + metadatosDeLaLrbrl},
		}, resultado)
		banco.compruebaCacheSinCrear(t)
	})
}

// falloDeLaEleccion es el 3 de la elección de la lectura fechada, el de FR-003 o
// el de FR-004: «no encontrado», con la dirección del bloque y la fecha de
// consulta de sus versiones, y sin texto (contracts/articulo-fecha.md §3;
// research.md D7).
func falloDeLaEleccion(direccion string, instante time.Time, mensaje string) falloDeLaConsulta {
	return falloDeLaConsulta{
		direccion: direccion,
		instante:  instante,
		clase:     schema.ClaseNoEncontrado,
		codigo:    3,
		mensaje:   mensaje,
	}
}

// falloSinEntradaDeLasVersiones es el de la lectura fechada sin la entrada de las
// versiones del bloque con --offline: «fuente no disponible», código 4, con la
// dirección del bloque, sin instante y con un mensaje que nombra --offline, la
// entrada de las versiones y su clave (FR-020).
func falloSinEntradaDeLasVersiones(direccion string) falloDeLaConsulta {
	return falloDeLaConsulta{
		direccion: direccion,
		clase:     schema.ClaseFuenteNoDisponible,
		codigo:    4,
		mensaje: `con --offline no se pide nada a la fuente y no hay ninguna entrada vigente de versiones con la clave ` +
			`"boe.legislacion-consolidada|1|versiones|` + direccion + `" (` + direccion + ")",
	}
}

// articuloSinFecha es el data de la lectura sin fecha del bloque, resuelta con
// su propia fuente sobre las grabaciones y con la caché vacía.
func articuloSinFecha(t *testing.T, norma, bloque string) Articulo {
	t.Helper()

	banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))

	resultado, err := banco.resuelve(t, schema.Contexto{}, ConsultaArticulo{Norma: norma, Bloque: bloque})
	require.NoError(t, err)

	datos, esArticulo := resultado.Datos.(Articulo)
	require.Truef(t, esArticulo, "el data de la lectura sin fecha es de tipo %T", resultado.Datos)

	return datos
}

// articuloDelGolden es el data del golden de hoy con ese nombre.
func articuloDelGolden(t *testing.T, nombre string) Articulo {
	t.Helper()

	contenido, err := leerGolden(filepath.Join(carpetaDeLosGolden, nombre+extensionDeLosGolden))
	require.NoError(t, err)

	var golden struct {
		Data Articulo `json:"data"`
	}
	require.NoError(t, json.Unmarshal(contenido, &golden))
	require.NotEmpty(t, golden.Data.Texto)

	return golden.Data
}

// vigenciaElegida es la fecha_vigencia de la versión que eligió una lectura
// fechada resuelta.
func vigenciaElegida(t *testing.T, resultado schema.Resultado) string {
	t.Helper()

	datos, esFechado := resultado.Datos.(ArticuloFechado)
	require.Truef(t, esFechado, "el data de la lectura fechada es de tipo %T", resultado.Datos)

	return datos.FechaVigencia
}

// compruebaMismoHash exige que los dos resultados den el mismo hash del sobre.
func compruebaMismoHash(t *testing.T, esperado, obtenido schema.Resultado) {
	t.Helper()

	deEsperado, err := schema.Huella(esperado.Datos)
	require.NoError(t, err)
	deObtenido, err := schema.Huella(obtenido.Datos)
	require.NoError(t, err)

	assert.Equal(t, deEsperado, deObtenido)
}

// contenidoGuardado es el contenido de la entrada vigente con la clave en la
// caché del banco, que tiene que estar.
func (b *bancoDeLaFuente) contenidoGuardado(t *testing.T, clave string) []byte {
	t.Helper()

	lectora, err := cache.New(t.Context(), cache.ConDirectorio(b.carpetaDeLaCache), cache.ConReloj(b.reloj.ahora))
	require.NoError(t, err)

	contenido, presente, err := lectora.Get(t.Context(), clave)
	require.NoError(t, lectora.Close())
	require.NoError(t, err)
	require.True(t, presente, "la entrada %q está en la caché", clave)

	return contenido
}
