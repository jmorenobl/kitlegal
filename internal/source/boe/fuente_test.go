package boe

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"maps"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

// actualizarGolden es la bandera con la que TestGolden escribe, antes de
// compararlos, los golden de todos los casos de la lista cerrada. Solo la usa la
// tarea [datos] de los golden, con la orden del contrato
// esquemas-fixtures-y-controles §2, y lo escrito lo revisa una persona (FR-112).
var actualizarGolden = flag.Bool("actualizar-golden", false,
	"escribe en testdata/golden el golden de cada caso de la lista cerrada antes de compararlos")

// Lo que comparten las pruebas de este fichero.
const (
	// codigoDeLoInesperado es el código de salida con el que el kernel termina
	// un fallo de clase «inesperado».
	codigoDeLoInesperado = 1
	// deQuienCompone es el principio del mensaje de un defecto al componer la
	// fuente (contrato errores-y-codigos, fila 22).
	deQuienCompone = "la fuente del BOE no se puede componer: "
	// direccionInvocada es la del recurso de las invocaciones de prueba: la que
	// nombra cualquier fallo de la caché (fila 21).
	direccionInvocada = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/metadatos"
	// motivoAlAbrir y motivoAlCerrar son los mensajes de la caché que no se deja
	// abrir o cerrar, delante de la dirección.
	motivoAlAbrir  = "no se ha podido abrir la caché"
	motivoAlCerrar = "no se ha podido cerrar la caché"
)

// El cliente de la caché es la CacheAbierta de producción, y que lo siga siendo
// no depende de que alguien lo recuerde.
var _ CacheAbierta = (*cache.Cliente)(nil)

// TestNuevaRechazaDependenciasAusentes fija la composición de la fuente (contrato
// puerto-y-applet §3.1; errores-y-codigos, fila 22): sin ConCliente o sin
// ConCache, con una función, un registrador o una opción nulos, Nueva no
// devuelve ninguna fuente y falla con «inesperado», porque es un defecto de quien
// la compone y no de quien invoca; el error no lleva dirección ni instante, de
// modo que el sobre lo firma y lo fecha el kernel. Componerla no construye ningún
// cliente ni abre ninguna caché, y sin ConRegistrador los eventos se descartan.
func TestNuevaRechazaDependenciasAusentes(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		opciones func(dependencias *dependenciasDePrueba) []Opcion
		mensaje  string
	}{
		{
			nombre:   "sin-ninguna-opcion",
			opciones: func(*dependenciasDePrueba) []Opcion { return nil },
			mensaje:  deQuienCompone + "faltan las dependencias ConCliente y ConCache",
		},
		{
			nombre: "solo-el-registrador",
			opciones: func(*dependenciasDePrueba) []Opcion {
				return []Opcion{ConRegistrador(slog.New(slog.DiscardHandler))}
			},
			mensaje: deQuienCompone + "faltan las dependencias ConCliente y ConCache",
		},
		{
			nombre: "sin-el-cliente",
			opciones: func(dependencias *dependenciasDePrueba) []Opcion {
				return []Opcion{ConCache(dependencias.abrir)}
			},
			mensaje: deQuienCompone + "falta la dependencia ConCliente",
		},
		{
			nombre: "sin-la-cache",
			opciones: func(dependencias *dependenciasDePrueba) []Opcion {
				return []Opcion{ConCliente(dependencias.construir)}
			},
			mensaje: deQuienCompone + "falta la dependencia ConCache",
		},
		{
			nombre: "cliente-nulo",
			opciones: func(dependencias *dependenciasDePrueba) []Opcion {
				return []Opcion{ConCliente(nil), ConCache(dependencias.abrir)}
			},
			mensaje: deQuienCompone + "la opción ConCliente no lleva ninguna función que construya el cliente",
		},
		{
			nombre: "cache-nula",
			opciones: func(dependencias *dependenciasDePrueba) []Opcion {
				return []Opcion{ConCliente(dependencias.construir), ConCache(nil)}
			},
			mensaje: deQuienCompone + "la opción ConCache no lleva ninguna función que abra la caché",
		},
		{
			nombre: "registrador-nulo",
			opciones: func(dependencias *dependenciasDePrueba) []Opcion {
				return []Opcion{ConCliente(dependencias.construir), ConCache(dependencias.abrir), ConRegistrador(nil)}
			},
			mensaje: deQuienCompone + "la opción ConRegistrador no lleva ningún registrador",
		},
		{
			nombre: "opcion-nula",
			opciones: func(dependencias *dependenciasDePrueba) []Opcion {
				return []Opcion{ConCliente(dependencias.construir), nil, ConCache(dependencias.abrir)}
			},
			mensaje: deQuienCompone + "la opción 2 de 3 es nula",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			dependencias := &dependenciasDePrueba{}

			fuente, err := Nueva(caso.opciones(dependencias)...)

			assert.Nil(t, fuente)
			compruebaDefectoAlComponer(t, err, caso.mensaje)
			dependencias.compruebaSinUso(t)
		})
	}

	t.Run("con-las-dos-dependencias-y-sin-registrador", func(t *testing.T) {
		t.Parallel()

		dependencias := &dependenciasDePrueba{}

		fuente, err := Nueva(ConCliente(dependencias.construir), ConCache(dependencias.abrir))

		require.NoError(t, err)
		require.NotNil(t, fuente)
		assert.Equal(t, slog.DiscardHandler, fuente.registrador.Handler())
		dependencias.compruebaSinUso(t)
	})

	t.Run("con-registrador", func(t *testing.T) {
		t.Parallel()

		dependencias := &dependenciasDePrueba{}
		registrador := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

		fuente, err := Nueva(ConRegistrador(registrador), ConCliente(dependencias.construir), ConCache(dependencias.abrir))

		require.NoError(t, err)
		require.NotNil(t, fuente)
		assert.Same(t, registrador, fuente.registrador)
		dependencias.compruebaSinUso(t)
	})
}

// TestFuenteNombreVigenciasYTerminos fija lo que la fuente declara de sí misma
// por el puerto core.Source (FR-091, FR-121; data-model.md §8): su nombre, la
// vigencia de lo que responde cada consulta —la de CACHE_TTL en refs/boe.py
// 37-43, que no depende de los argumentos— y los términos de uso de terminos.go.
// Una consulta que la fuente no declara no tiene vigencia, porque nunca se guarda
// nada de ella. Las vigencias van escritas en segundos, como en el contrato, para
// que un cambio de las constantes no pase por aquí en silencio.
func TestFuenteNombreVigenciasYTerminos(t *testing.T) {
	t.Parallel()

	const (
		vigenciaDeCincoMinutos = 300 * time.Second
		vigenciaDeSieteDias    = 604_800 * time.Second
		norma                  = "BOE-A-2015-10565"
	)

	fuente := fuenteDePrueba(t, &dependenciasDePrueba{})

	assert.Equal(t, "boe.legislacion-consolidada", fuente.Name())
	assert.Equal(t, terminosDeUso, fuente.Terms())

	casos := []struct {
		nombre   string
		consulta core.Consulta
		vigencia time.Duration
	}{
		{nombre: "buscar", consulta: ConsultaBuscar{Texto: []string{"procedimiento"}}, vigencia: vigenciaDeCincoMinutos},
		{nombre: "buscar-sin-texto", consulta: ConsultaBuscar{}, vigencia: vigenciaDeCincoMinutos},
		{nombre: "metadatos", consulta: ConsultaMetadatos{Norma: norma}, vigencia: vigenciaDeCincoMinutos},
		{nombre: "indice", consulta: ConsultaIndice{Norma: norma}, vigencia: vigenciaDeSieteDias},
		{nombre: "articulo", consulta: ConsultaArticulo{Norma: norma, Bloque: "a21"}, vigencia: vigenciaDeSieteDias},
		{
			nombre:   "articulos",
			consulta: ConsultaArticulos{Norma: norma, Bloques: []string{"a21", "a22"}},
			vigencia: vigenciaDeSieteDias,
		},
		{nombre: "analisis", consulta: ConsultaAnalisis{Norma: norma}, vigencia: vigenciaDeSieteDias},
		{nombre: "de-otro-tipo-con-un-verbo-de-la-fuente", consulta: consultaAjena{verbo: verboMetadatos}},
		{nombre: "sin-consulta", consulta: nil},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.vigencia, fuente.TTL(caso.consulta))
		})
	}
}

// TestConsultaDeOtroTipo fija que Fetch no resuelve una consulta que la fuente no
// declara, aunque su verbo se llame como uno de los suyos, ni la ausencia de
// consulta: «inesperado», código 1, sin procedencia —la firma el kernel— y sin
// abrir la caché ni construir el cliente, también con --offline y con --dry-run,
// que conservan el fallo (contrato errores-y-codigos, fila 22 y nota de la tabla).
func TestConsultaDeOtroTipo(t *testing.T) {
	t.Parallel()

	consultas := []struct {
		nombre   string
		consulta core.Consulta
		mensaje  string
	}{
		{
			nombre:   "de-otro-tipo-con-un-verbo-de-la-fuente",
			consulta: consultaAjena{verbo: verboArticulo},
			mensaje:  "la fuente del BOE no resuelve consultas de tipo boe.consultaAjena",
		},
		{
			nombre:   "de-otro-tipo-con-un-verbo-ajeno",
			consulta: consultaAjena{verbo: "sumario"},
			mensaje:  "la fuente del BOE no resuelve consultas de tipo boe.consultaAjena",
		},
		{
			nombre:   "sin-consulta",
			consulta: nil,
			mensaje:  "la fuente del BOE no ha recibido ninguna consulta",
		},
	}

	ejecuciones := []struct {
		nombre string
		ec     schema.Contexto
	}{
		{nombre: "normal", ec: schema.Contexto{}},
		{nombre: "offline", ec: schema.Contexto{Offline: true}},
		{nombre: "ensayo", ec: schema.Contexto{DryRun: true}},
	}

	for _, consulta := range consultas {
		for _, ejecucion := range ejecuciones {
			t.Run(consulta.nombre+"-"+ejecucion.nombre, func(t *testing.T) {
				t.Parallel()

				dependencias := &dependenciasDePrueba{cacheAbierta: &cacheDePrueba{}}
				fuente := fuenteDePrueba(t, dependencias)

				resultado, err := fuente.Fetch(t.Context(), ejecucion.ec, consulta.consulta)

				assert.Zero(t, resultado)
				compruebaDefectoAlComponer(t, err, consulta.mensaje)
				dependencias.compruebaSinUso(t)
			})
		}
	}
}

// TestInvocarAbreYCierraLaCache fija lo que cada verbo obtiene de invocar una
// vez validada su consulta (contrato puerto-y-applet §3.1; research.md D5): la
// caché se abre en solo lectura con --offline o con --dry-run y en modo normal en
// otro caso, sigue abierta mientras se resuelve y se cierra antes de volver; el
// cliente se construye solo si la resolución lo pide, y una sola vez; y cualquier
// dependencia que no se deja usar es un defecto al componer, sin que nada entre
// en pánico.
func TestInvocarAbreYCierraLaCache(t *testing.T) {
	t.Parallel()

	resuelto := schema.Resultado{
		Procedencia: schema.Procedencia{
			Fuente:        NombreDeLaFuente,
			URL:           direccionInvocada,
			FechaConsulta: time.Date(2026, time.September, 13, 10, 30, 0, 123456789, time.UTC),
		},
		Datos: "los datos de la consulta",
	}

	t.Run("abre-en-el-modo-de-la-ejecucion-y-cierra-al-volver", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre      string
			ec          schema.Contexto
			soloLectura bool
		}{
			{nombre: "normal", ec: schema.Contexto{Timeout: time.Minute, Asunto: "asunto-de-prueba"}},
			{nombre: "offline", ec: schema.Contexto{Offline: true}, soloLectura: true},
			{nombre: "ensayo", ec: schema.Contexto{DryRun: true}, soloLectura: true},
			{nombre: "offline-y-ensayo", ec: schema.Contexto{Offline: true, DryRun: true}, soloLectura: true},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				dependencias := &dependenciasDePrueba{cacheAbierta: &cacheDePrueba{}}
				fuente := fuenteDePrueba(t, dependencias)

				var (
					recibida          *invocacion
					cierresAlResolver int
				)

				resultado, err := fuente.invocar(t.Context(), caso.ec, direccionInvocada,
					func(_ context.Context, en *invocacion) (schema.Resultado, error) {
						recibida = en
						cierresAlResolver = dependencias.cacheAbierta.cierres

						return resuelto, nil
					})

				require.NoError(t, err)
				assert.Equal(t, resuelto, resultado)
				assert.Equal(t, []bool{caso.soloLectura}, dependencias.aperturas)
				require.NotNil(t, recibida)
				assert.Equal(t, caso.ec, recibida.ec)
				assert.Same(t, dependencias.cacheAbierta, recibida.cache)
				assert.Zero(t, cierresAlResolver, "la caché tiene que seguir abierta mientras se resuelve")
				assert.Equal(t, 1, dependencias.cacheAbierta.cierres)
				assert.Zero(t, dependencias.construcciones, "sin pedir nada no se construye ningún cliente")
			})
		}
	})

	t.Run("construye-el-cliente-solo-si-se-pide-y-una-vez", func(t *testing.T) {
		t.Parallel()

		pedidor := &pedidorDePrueba{responde: fallaCon(errors.New("la prueba no pide nada"))}
		dependencias := &dependenciasDePrueba{cliente: pedidor, cacheAbierta: &cacheDePrueba{}}
		fuente := fuenteDePrueba(t, dependencias)

		var (
			construccionesAlEmpezar int
			pedidores               []Pedidor
		)

		_, err := fuente.invocar(t.Context(), schema.Contexto{}, direccionInvocada,
			func(_ context.Context, en *invocacion) (schema.Resultado, error) {
				construccionesAlEmpezar = dependencias.construcciones

				for range 2 {
					obtenido, err := en.pedidor()
					if err != nil {
						return schema.Resultado{}, err
					}

					pedidores = append(pedidores, obtenido)
				}

				return resuelto, nil
			})

		require.NoError(t, err)
		assert.Zero(t, construccionesAlEmpezar, "abrir la caché no construye el cliente")
		assert.Equal(t, 1, dependencias.construcciones)
		assert.Equal(t, []Pedidor{pedidor, pedidor}, pedidores)
		assert.Empty(t, pedidor.peticiones)
	})

	t.Run("el-cliente-que-no-se-construye-es-un-defecto-al-componer", func(t *testing.T) {
		t.Parallel()

		causa := errorConClaseDePrueba{clase: schema.ClaseArgumentos}

		casos := []struct {
			nombre       string
			dependencias *dependenciasDePrueba
			mensaje      string
		}{
			{
				nombre:       "con-error",
				dependencias: &dependenciasDePrueba{errDelCliente: causa, cacheAbierta: &cacheDePrueba{}},
				mensaje:      deQuienCompone + "no se ha podido construir el cliente HTTP: " + causa.Error(),
			},
			{
				nombre:       "sin-cliente-ni-error",
				dependencias: &dependenciasDePrueba{cacheAbierta: &cacheDePrueba{}},
				mensaje:      deQuienCompone + "la función de ConCliente no ha devuelto ni cliente ni error",
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				fuente := fuenteDePrueba(t, caso.dependencias)

				_, err := fuente.invocar(t.Context(), schema.Contexto{}, direccionInvocada,
					func(_ context.Context, en *invocacion) (schema.Resultado, error) {
						_, err := en.pedidor()

						return schema.Resultado{}, err
					})

				compruebaDefectoAlComponer(t, err, caso.mensaje)
				assert.Equal(t, 1, caso.dependencias.construcciones)
				assert.Equal(t, 1, caso.dependencias.cacheAbierta.cierres, "la caché se cierra también si la invocación falla")
			})
		}
	})

	t.Run("la-apertura-que-no-devuelve-cache-ni-error-es-un-defecto-al-componer", func(t *testing.T) {
		t.Parallel()

		dependencias := &dependenciasDePrueba{}
		fuente := fuenteDePrueba(t, dependencias)

		resultado, err := fuente.invocar(t.Context(), schema.Contexto{}, direccionInvocada, resolverQueNoSeLlama(t))

		assert.Zero(t, resultado)
		compruebaDefectoAlComponer(t, err, deQuienCompone+"la función de ConCache no ha devuelto ni caché ni error")
		assert.Equal(t, []bool{false}, dependencias.aperturas)
	})

	t.Run("la-fuente-que-no-se-construyo-con-nueva-es-un-defecto-al-componer", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre string
			fuente *Fuente
		}{
			{nombre: "nula", fuente: nil},
			{nombre: "a-cero", fuente: &Fuente{}},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				resultado, err := caso.fuente.invocar(t.Context(), schema.Contexto{}, direccionInvocada,
					resolverQueNoSeLlama(t))

				assert.Zero(t, resultado)
				compruebaDefectoAlComponer(t, err, deQuienCompone+"la fuente no se ha construido con Nueva")
			})
		}
	})
}

// TestInvocarFallosDeLaCache fija los fallos de la caché que invocar abre y
// cierra (contrato errores-y-codigos, fila 21; puerto-y-applet §3.1): llevan la
// clase que declara la caché si es una de las tres de cache.Error, e
// «inesperado» si no, nunca la de identidad humana (FR-100); la dirección del
// recurso consultado, que firma el sobre con la fuente; ningún instante, porque
// la caché no pide nada y el sobre lo fecha el montaje; y la causa, alcanzable
// con errors.Is. Si falla la apertura no se resuelve ni se cierra nada; si falla
// solo el cierre, la invocación falla con su clase y sin datos ni fecha; y si
// fallan la resolución y el cierre, prevalece la resolución y los dos errores
// quedan unidos.
func TestInvocarFallosDeLaCache(t *testing.T) {
	t.Parallel()

	_, deLaCacheReal := cache.New(t.Context(), cache.ConDirectorio(""))
	require.Error(t, deLaCacheReal)

	soloLaProcedencia := schema.Procedencia{Fuente: NombreDeLaFuente, URL: direccionInvocada}
	clases := []struct {
		nombre string
		causa  error
		clase  schema.Clase
		// enElMensaje dice si el texto de la causa entra en el mensaje, que es
		// cuando declara su clase.
		enElMensaje bool
	}{
		{nombre: "la-de-la-cache-real", causa: deLaCacheReal, clase: schema.ClaseArgumentos, enElMensaje: true},
		{
			nombre:      "fuente-no-disponible",
			causa:       errorConClaseDePrueba{clase: schema.ClaseFuenteNoDisponible},
			clase:       schema.ClaseFuenteNoDisponible,
			enElMensaje: true,
		},
		{
			nombre:      "inesperado",
			causa:       errorConClaseDePrueba{clase: schema.ClaseInesperado},
			clase:       schema.ClaseInesperado,
			enElMensaje: true,
		},
		{
			nombre:      "no-encontrado-que-la-cache-no-declara",
			causa:       errorConClaseDePrueba{clase: schema.ClaseNoEncontrado},
			clase:       schema.ClaseInesperado,
			enElMensaje: true,
		},
		{
			nombre:      "identidad-humana",
			causa:       errorConClaseDePrueba{clase: schema.ClaseIdentidadHumana},
			clase:       schema.ClaseInesperado,
			enElMensaje: true,
		},
		{nombre: "sin-clase", causa: errors.New("disk I/O error"), clase: schema.ClaseInesperado},
	}

	for _, deLaCache := range clases {
		t.Run("la-apertura-falla-"+deLaCache.nombre, func(t *testing.T) {
			t.Parallel()

			dependencias := &dependenciasDePrueba{errDeLaApertura: deLaCache.causa}
			fuente := fuenteDePrueba(t, dependencias)

			resultado, err := fuente.invocar(t.Context(), schema.Contexto{Offline: true}, direccionInvocada,
				resolverQueNoSeLlama(t))

			assert.Equal(t, schema.Resultado{Procedencia: soloLaProcedencia}, resultado)
			compruebaFalloDeLaCache(t, err, deLaCache.causa, deLaCache.clase,
				mensajeDeLaCache(motivoAlAbrir, deLaCache.causa, deLaCache.enElMensaje))
			assert.Equal(t, []bool{true}, dependencias.aperturas)
			assert.Zero(t, dependencias.construcciones)
		})

		t.Run("el-cierre-falla-tras-resolver-"+deLaCache.nombre, func(t *testing.T) {
			t.Parallel()

			dependencias := &dependenciasDePrueba{cacheAbierta: &cacheDePrueba{errDelCierre: deLaCache.causa}}
			fuente := fuenteDePrueba(t, dependencias)
			resuelto := schema.Resultado{
				Procedencia: schema.Procedencia{
					Fuente:        NombreDeLaFuente,
					URL:           direccionInvocada,
					FechaConsulta: time.Date(2026, time.September, 13, 10, 30, 0, 0, time.UTC),
				},
				Datos: "los datos de la consulta",
			}

			resultado, err := fuente.invocar(t.Context(), schema.Contexto{}, direccionInvocada,
				func(context.Context, *invocacion) (schema.Resultado, error) { return resuelto, nil })

			assert.Equal(t, schema.Resultado{Procedencia: soloLaProcedencia}, resultado)
			compruebaFalloDeLaCache(t, err, deLaCache.causa, deLaCache.clase,
				mensajeDeLaCache(motivoAlCerrar, deLaCache.causa, deLaCache.enElMensaje))
			assert.Equal(t, 1, dependencias.cacheAbierta.cierres)
		})
	}

	t.Run("el-cierre-falla-tras-un-ensayo-que-lo-describe", func(t *testing.T) {
		t.Parallel()

		causa := errorConClaseDePrueba{clase: schema.ClaseInesperado}
		dependencias := &dependenciasDePrueba{cacheAbierta: &cacheDePrueba{errDelCierre: causa}}
		fuente := fuenteDePrueba(t, dependencias)
		ensayado := schema.Resultado{Procedencia: soloLaProcedencia, Ensayo: []string{"GET " + direccionInvocada}}

		resultado, err := fuente.invocar(t.Context(), schema.Contexto{DryRun: true}, direccionInvocada,
			func(context.Context, *invocacion) (schema.Resultado, error) { return ensayado, nil })

		assert.Equal(t, ensayado, resultado, "el fallo conserva la descripción de lo que no se hizo")
		compruebaFalloDeLaCache(t, err, causa, schema.ClaseInesperado, mensajeDeLaCache(motivoAlCerrar, causa, true))
	})

	t.Run("la-resolucion-que-falla-prevalece-sobre-el-cierre", func(t *testing.T) {
		t.Parallel()

		instante := time.Date(2026, time.September, 13, 10, 30, 0, 123456789, time.UTC)
		fallo := errorDeNoEncontrado(direccionInvocada, instante, nil, "la norma BOE-A-2015-10565 no tiene metadatos")
		delFallo := resultadoDelFallo(fallo)
		cierre := errorConClaseDePrueba{clase: schema.ClaseInesperado}

		require.Equal(t, schema.Resultado{Procedencia: schema.Procedencia{
			Fuente:        NombreDeLaFuente,
			URL:           direccionInvocada,
			FechaConsulta: instante,
		}}, delFallo)

		casos := []struct {
			nombre  string
			cierre  error
			mensaje string
		}{
			{nombre: "el-cierre-no-falla", mensaje: fallo.Error()},
			{
				nombre:  "el-cierre-falla",
				cierre:  cierre,
				mensaje: fallo.Error() + "\n" + mensajeDeLaCache(motivoAlCerrar, cierre, true),
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				dependencias := &dependenciasDePrueba{cacheAbierta: &cacheDePrueba{errDelCierre: caso.cierre}}
				fuente := fuenteDePrueba(t, dependencias)

				resultado, err := fuente.invocar(t.Context(), schema.Contexto{}, direccionInvocada,
					func(context.Context, *invocacion) (schema.Resultado, error) { return delFallo, fallo })

				assert.Equal(t, delFallo, resultado)
				require.ErrorIs(t, err, fallo)
				assert.Equal(t, schema.ClaseNoEncontrado, cli.Clasificar(err))
				assert.Equal(t, caso.mensaje, err.Error())
				assert.Equal(t, 1, dependencias.cacheAbierta.cierres)

				if caso.cierre != nil {
					require.ErrorIs(t, err, caso.cierre)
				}
			})
		}
	})
}

// TestGolden compara byte a byte cada golden de testdata/golden con lo que da su
// caso de la lista cerrada (FR-112; contrato esquemas-fixtures-y-controles §2): la
// fuente sobre la reproducción de las grabaciones, con una caché vacía para cada
// caso, resuelve su consulta, y su data y su url se escriben en la forma canónica
// —claves ordenadas, dos espacios de sangrado, sin escapar HTML y con salto
// final—, sin la fecha, que no forma parte de lo que se fija. Un fichero que no es
// el golden de ningún caso hace fallar la prueba, y mientras no hay ninguno no se
// compara nada. Con -actualizar-golden escribe antes el de cada caso.
//
// El primer subtest compara los golden reales; el resto fija la lista cerrada y
// la forma canónica, y demuestra sobre carpetas temporales que el comparador no
// pasa en vacío.
func TestGolden(t *testing.T) {
	t.Parallel()

	if *actualizarGolden {
		for _, caso := range casosDeGolden {
			contenido, err := goldenDelCaso(t, caso)
			require.NoError(t, err)

			escribeGolden(t, carpetaDeLosGolden, caso, contenido)
		}
	}

	t.Run("testdata-golden", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, comprobarGolden(carpetaDeLosGolden, casosDeGolden, func(caso casoDeGolden) ([]byte, error) {
			return goldenDelCaso(t, caso)
		}))
	})

	t.Run("lista-cerrada", func(t *testing.T) {
		t.Parallel()

		const casosDelContrato = 13

		require.Len(t, casosDeGolden, casosDelContrato)

		nombres := make(map[string]bool, len(casosDeGolden))
		verbos := make(map[string]bool)

		for _, caso := range casosDeGolden {
			assert.False(t, nombres[caso.nombre], "el caso %s está repetido", caso.nombre)
			nombres[caso.nombre] = true

			verbo := caso.consulta.Verbo()
			assert.True(t, strings.HasPrefix(caso.nombre, verbo+"-"), "el caso %s no empieza por su verbo, %s", caso.nombre, verbo)
			verbos[verbo] = true
		}

		assert.ElementsMatch(t,
			[]string{verboBuscar, verboIndice, verboArticulo, verboArticulos, verboMetadatos, verboAnalisis},
			slices.Collect(maps.Keys(verbos)))
	})

	t.Run("forma-canonica", func(t *testing.T) {
		t.Parallel()

		type datosDePrueba struct {
			Zeta   string   `json:"zeta"`
			Alfa   []string `json:"alfa"`
			Numero uint64   `json:"numero"`
		}

		const esperada = `{
  "data": {
    "alfa": [
      "<p>Artículo 21 & «siguiente»</p>"
    ],
    "numero": 18446744073709551615,
    "zeta": ""
  },
  "url": "https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565#a21"
}
`

		contenido, err := formaCanonica(contenidoDelGolden{
			Data: datosDePrueba{Alfa: []string{"<p>Artículo 21 & «siguiente»</p>"}, Numero: math.MaxUint64},
			URL:  "https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565#a21",
		})

		// Línea a línea y no como JSON equivalente: lo que se fija es la forma
		// exacta —orden de claves, sangrado y salto final—, no solo el valor.
		require.NoError(t, err)
		assert.Equal(t, strings.Split(esperada, "\n"), strings.Split(string(contenido), "\n"))
	})

	t.Run("escribe-cada-golden-solo-para-su-propietario", func(t *testing.T) {
		t.Parallel()

		caso := casoDeGolden{nombre: "metadatos-de-prueba", consulta: ConsultaMetadatos{Norma: "BOE-A-2015-10565"}}
		carpeta := filepath.Join(t.TempDir(), "golden")

		escribeGolden(t, carpeta, caso, goldenDePrueba(caso))

		fichero, err := os.Stat(filepath.Join(carpeta, "metadatos-de-prueba.json"))
		require.NoError(t, err)
		assert.Equal(t, fs.FileMode(0o600), fichero.Mode().Perm())
		require.NoError(t, comprobarGolden(carpeta, []casoDeGolden{caso}, func(casoDeGolden) ([]byte, error) {
			return goldenDePrueba(caso), nil
		}))
	})

	compruebaElComparador(t)
}

// compruebaElComparador demuestra sobre carpetas temporales, con dos casos de
// prueba y un generador que anota los casos que genera, que comprobarGolden solo
// genera los casos que tienen golden, que no compara nada sin golden y que falla
// nombrando el fichero ante un byte distinto, un fichero sin caso o un caso que no
// da su golden.
func compruebaElComparador(t *testing.T) {
	t.Helper()

	metadatos := casoDeGolden{nombre: "metadatos-de-prueba", consulta: ConsultaMetadatos{Norma: "BOE-A-2015-10565"}}
	indice := casoDeGolden{nombre: "indice-de-prueba", consulta: ConsultaIndice{Norma: "BOE-A-2015-10565"}}
	casosDePrueba := []casoDeGolden{metadatos, indice}

	casos := []struct {
		nombre string
		// prepara deja en la carpeta de los golden, que no existe hasta que
		// prepara la crea, lo que el caso necesita.
		prepara func(t *testing.T, carpeta string)
		// fallaElCaso hace que el generador no dé ningún golden.
		fallaElCaso bool
		// generados son los casos que el comparador tiene que haber generado.
		generados []string
		// mensaje es un fragmento del error esperado; vacío, los golden valen.
		mensaje string
	}{
		{nombre: "sin-carpeta", prepara: func(t *testing.T, _ string) { t.Helper() }},
		{
			nombre:  "carpeta-vacia",
			prepara: func(t *testing.T, carpeta string) { t.Helper(); require.NoError(t, os.Mkdir(carpeta, 0o750)) },
		},
		{
			nombre: "golden-igual-a-su-caso",
			prepara: func(t *testing.T, carpeta string) {
				t.Helper()
				escribeGolden(t, carpeta, metadatos, goldenDePrueba(metadatos))
			},
			generados: []string{metadatos.nombre},
		},
		{
			nombre: "un-byte-distinto",
			prepara: func(t *testing.T, carpeta string) {
				t.Helper()
				escribeGolden(t, carpeta, metadatos, bytes.Replace(goldenDePrueba(metadatos), []byte("es/"), []byte("es/X"), 1))
			},
			generados: []string{metadatos.nombre},
			mensaje: "metadatos-de-prueba.json no coincide byte a byte con lo que da su caso: la línea 3 es " +
				`«  "url": "https://www.boe.es/X"» en el golden y «  "url": "https://www.boe.es/"» en su caso`,
		},
		{
			nombre: "un-salto-final-de-menos",
			prepara: func(t *testing.T, carpeta string) {
				t.Helper()
				escribeGolden(t, carpeta, metadatos, bytes.TrimSuffix(goldenDePrueba(metadatos), []byte("\n")))
			},
			generados: []string{metadatos.nombre},
			mensaje:   "la línea 5 es el final del fichero en el golden y «» en su caso",
		},
		{
			nombre: "fichero-sin-caso",
			prepara: func(t *testing.T, carpeta string) {
				t.Helper()
				escribeGolden(t, carpeta, metadatos, goldenDePrueba(metadatos))
				escribeDatoDePrueba(t, filepath.Join(carpeta, "otro-caso.json"), []byte("{}\n"))
			},
			generados: []string{metadatos.nombre},
			mensaje:   "otro-caso.json no es el golden de ningún caso",
		},
		{
			nombre: "fichero-que-no-es-un-golden",
			prepara: func(t *testing.T, carpeta string) {
				t.Helper()
				escribeDatoDePrueba(t, filepath.Join(carpeta, "LEEME.md"), []byte("golden\n"))
			},
			mensaje: "LEEME.md no es el golden de ningún caso",
		},
		{
			nombre: "carpeta-con-el-nombre-de-un-golden",
			prepara: func(t *testing.T, carpeta string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(filepath.Join(carpeta, "indice-de-prueba.json"), 0o750))
			},
			mensaje: "indice-de-prueba.json no es el golden de ningún caso",
		},
		{
			nombre: "caso-que-no-da-su-golden",
			prepara: func(t *testing.T, carpeta string) {
				t.Helper()
				escribeGolden(t, carpeta, indice, goldenDePrueba(indice))
			},
			fallaElCaso: true,
			generados:   []string{indice.nombre},
			mensaje:     "el caso indice-de-prueba no da su golden: la consulta de prueba ha fallado",
		},
		{
			nombre: "carpeta-que-es-un-fichero",
			prepara: func(t *testing.T, carpeta string) {
				t.Helper()
				escribeDatoDePrueba(t, carpeta, []byte("{}\n"))
			},
			mensaje: "la carpeta de los golden no se puede leer",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			carpeta := filepath.Join(t.TempDir(), "golden")
			caso.prepara(t, carpeta)

			var generados []string

			err := comprobarGolden(carpeta, casosDePrueba, func(generado casoDeGolden) ([]byte, error) {
				generados = append(generados, generado.nombre)
				if caso.fallaElCaso {
					return nil, errors.New("la consulta de prueba ha fallado")
				}

				return goldenDePrueba(generado), nil
			})

			assert.Equal(t, caso.generados, generados)

			if caso.mensaje == "" {
				require.NoError(t, err)

				return
			}

			require.ErrorContains(t, err, caso.mensaje)
		})
	}
}

// contenidoDelGolden es lo que fija el golden de un caso: el data y la url del
// resultado, sin la fecha de consulta (contrato §2).
type contenidoDelGolden struct {
	Data any    `json:"data"`
	URL  string `json:"url"`
}

// goldenDelCaso es el golden que da hoy el caso: la fuente compuesta sobre la
// reproducción de las grabaciones y una caché vacía para él solo resuelve su
// consulta, y su data y su url se escriben en la forma canónica.
func goldenDelCaso(t *testing.T, caso casoDeGolden) ([]byte, error) {
	t.Helper()

	fuente, err := Nueva(
		ConCliente(construirReproduccion(carpetaDeLasGrabaciones)),
		ConCache(abrirCacheEn(t.TempDir())),
	)
	if err != nil {
		return nil, err
	}

	resultado, err := fuente.Fetch(t.Context(), schema.Contexto{}, caso.consulta)
	if err != nil {
		return nil, err
	}

	return formaCanonica(contenidoDelGolden{Data: resultado.Datos, URL: resultado.Procedencia.URL})
}

// formaCanonica escribe el valor como lo fija un golden: claves ordenadas, dos
// espacios de sangrado, sin escapar HTML y con salto final. Pasa por un valor
// genérico para ordenar también las claves de los structs, con los números como
// json.Number para que ninguno pierda precisión por el camino.
func formaCanonica(valor any) ([]byte, error) {
	serializado, err := json.Marshal(valor)
	if err != nil {
		return nil, err
	}

	decodificador := json.NewDecoder(bytes.NewReader(serializado))
	decodificador.UseNumber()

	var generico any
	if err := decodificador.Decode(&generico); err != nil {
		return nil, err
	}

	var canonica bytes.Buffer

	codificador := json.NewEncoder(&canonica)
	codificador.SetEscapeHTML(false)
	codificador.SetIndent("", "  ")

	if err := codificador.Encode(generico); err != nil {
		return nil, err
	}

	return canonica.Bytes(), nil
}

// comprobarGolden compara byte a byte cada golden de la carpeta con lo que genera
// su caso, y solo esos: un caso sin golden no se genera, y una carpeta que no
// existe o que está vacía no compara nada, que es lo que hay antes de la tarea
// [datos] de los golden (research.md D12). Lo que hay en la carpeta y no es el
// golden de ningún caso es un error que lo nombra.
func comprobarGolden(carpeta string, casos []casoDeGolden, generar func(casoDeGolden) ([]byte, error)) error {
	entradas, err := os.ReadDir(filepath.Clean(carpeta))

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return fmt.Errorf("la carpeta de los golden no se puede leer: %w", err)
	}

	fallos := make([]error, 0, len(entradas))

	for _, entrada := range entradas {
		indice := slices.IndexFunc(casos, func(caso casoDeGolden) bool {
			return caso.nombre+extensionDeLosGolden == entrada.Name()
		})
		if indice < 0 || !entrada.Type().IsRegular() {
			fallos = append(fallos, fmt.Errorf("%s: %s no es el golden de ningún caso", carpeta, entrada.Name()))

			continue
		}

		fallos = append(fallos, comprobarUnGolden(ficheroDelGolden(carpeta, casos[indice]), casos[indice], generar))
	}

	return errors.Join(fallos...)
}

// comprobarUnGolden compara el golden de la ruta con lo que genera su caso y,
// si no coinciden, nombra la primera línea distinta.
func comprobarUnGolden(ruta string, caso casoDeGolden, generar func(casoDeGolden) ([]byte, error)) error {
	esperado, err := leerGolden(ruta)
	if err != nil {
		return err
	}

	generado, err := generar(caso)
	if err != nil {
		return fmt.Errorf("%s: el caso %s no da su golden: %w", ruta, caso.nombre, err)
	}

	if bytes.Equal(esperado, generado) {
		return nil
	}

	linea, delGolden, delCaso := primeraLineaDistinta(esperado, generado)

	return fmt.Errorf("%s no coincide byte a byte con lo que da su caso: la línea %d es %s en el golden y %s en su caso",
		ruta, linea, delGolden, delCaso)
}

// primeraLineaDistinta es el número de la primera línea en que difieren dos
// contenidos distintos y lo que dice cada uno en ella: la línea entre comillas
// latinas, o el final del fichero si ese contenido ya no tiene esa línea.
func primeraLineaDistinta(esperado, generado []byte) (int, string, string) {
	delGolden := strings.Split(string(esperado), "\n")
	delCaso := strings.Split(string(generado), "\n")

	for indice := range max(len(delGolden), len(delCaso)) {
		enElGolden, enElCaso := lineaDe(delGolden, indice), lineaDe(delCaso, indice)
		if enElGolden != enElCaso {
			return indice + 1, enElGolden, enElCaso
		}
	}

	return 0, "", ""
}

// lineaDe es la línea del índice entre comillas latinas, o el final del fichero
// si no hay tantas.
func lineaDe(lineas []string, indice int) string {
	if indice >= len(lineas) {
		return "el final del fichero"
	}

	return "«" + lineas[indice] + "»"
}

// ficheroDelGolden es la ruta del golden del caso en la carpeta.
func ficheroDelGolden(carpeta string, caso casoDeGolden) string {
	return filepath.Join(carpeta, caso.nombre+extensionDeLosGolden)
}

// leerGolden lee un golden por filepath.Clean (gosec G304), desde un auxiliar
// distinto del que escribe (G703).
func leerGolden(ruta string) ([]byte, error) {
	contenido, err := os.ReadFile(filepath.Clean(ruta))
	if err != nil {
		return nil, fmt.Errorf("%s no se puede leer: %w", ruta, err)
	}

	return contenido, nil
}

// escribeGolden deja el golden del caso en la carpeta, y la carpeta, solo para su
// propietario (gosec G301 y G306), desde un auxiliar distinto del que lee (G703).
func escribeGolden(t *testing.T, carpeta string, caso casoDeGolden, contenido []byte) {
	t.Helper()

	escribeDatoDePrueba(t, ficheroDelGolden(carpeta, caso), contenido)
}

// goldenDePrueba es un golden en forma canónica que solo depende del nombre del
// caso, para las pruebas del comparador.
func goldenDePrueba(caso casoDeGolden) []byte {
	return []byte("{\n  \"data\": \"" + caso.nombre + "\",\n  \"url\": \"https://www.boe.es/\"\n}\n")
}

// construirReproduccion es la función de ConCliente que construye el cliente real
// de httpx en reproducción sobre la carpeta de grabaciones, sin abrir ninguna
// conexión.
func construirReproduccion(carpeta string) func() (Pedidor, error) {
	return func() (Pedidor, error) {
		cliente, err := httpx.Replay(carpeta, httpx.ConFuente(NombreDeLaFuente))
		if err != nil {
			return nil, err
		}

		return cliente, nil
	}
}

// abrirCacheEn es la AperturaDeCache de la caché real en la carpeta, en solo
// lectura cuando la fuente lo pide.
func abrirCacheEn(carpeta string) AperturaDeCache {
	return func(ctx context.Context, soloLectura bool) (CacheAbierta, error) {
		opciones := []cache.Opcion{cache.ConDirectorio(carpeta)}
		if soloLectura {
			opciones = append(opciones, cache.SoloLectura())
		}

		abierta, err := cache.New(ctx, opciones...)
		if err != nil {
			return nil, err
		}

		return abierta, nil
	}
}

// consultaAjena es una consulta que la fuente no declara, con el verbo que se le
// dé, también uno de los de la fuente.
type consultaAjena struct {
	verbo string
}

// Verbo es el que se le dio.
func (c consultaAjena) Verbo() string {
	return c.verbo
}

// dependenciasDePrueba son las dos dependencias de la fuente: construir y abrir
// devuelven lo que se les da y anotan cada llamada. No son seguras para usarlas
// desde varias goroutines a la vez: cada prueba construye las suyas.
type dependenciasDePrueba struct {
	// cliente es lo que devuelve construir, y errDelCliente, su error.
	cliente       Pedidor
	errDelCliente error
	// cacheAbierta es lo que devuelve abrir si no es nula, y errDeLaApertura, su
	// error; sin ninguna de las dos, abrir no devuelve ni caché ni error.
	cacheAbierta    *cacheDePrueba
	errDeLaApertura error

	// construcciones cuenta las llamadas a construir, y aperturas anota el modo
	// de cada llamada a abrir: true en solo lectura.
	construcciones int
	aperturas      []bool
}

// construir es la función de ConCliente.
func (d *dependenciasDePrueba) construir() (Pedidor, error) {
	d.construcciones++

	if d.errDelCliente != nil {
		return nil, d.errDelCliente
	}

	return d.cliente, nil
}

// abrir es la AperturaDeCache.
func (d *dependenciasDePrueba) abrir(_ context.Context, soloLectura bool) (CacheAbierta, error) {
	d.aperturas = append(d.aperturas, soloLectura)

	switch {
	case d.errDeLaApertura != nil:
		return nil, d.errDeLaApertura
	case d.cacheAbierta == nil:
		return nil, nil
	default:
		return d.cacheAbierta, nil
	}
}

// compruebaSinUso exige que no se haya construido ningún cliente ni abierto
// ninguna caché.
func (d *dependenciasDePrueba) compruebaSinUso(t *testing.T) {
	t.Helper()

	assert.Zero(t, d.construcciones, "no se construye ningún cliente")
	assert.Empty(t, d.aperturas, "no se abre ninguna caché")
}

// cacheDePrueba es la CacheAbierta de las pruebas de la invocación: no guarda
// nada, cuenta los cierres y falla al cerrar con el error que se le dé.
type cacheDePrueba struct {
	errDelCierre error
	cierres      int
}

// Get no encuentra nada.
func (*cacheDePrueba) Get(context.Context, string) ([]byte, bool, error) {
	return nil, false, nil
}

// Put no guarda nada.
func (*cacheDePrueba) Put(context.Context, string, []byte, time.Duration) error {
	return nil
}

// Close cuenta el cierre y devuelve el error que se le dio.
func (c *cacheDePrueba) Close() error {
	c.cierres++

	return c.errDelCierre
}

// fuenteDePrueba compone la fuente con las dependencias de prueba.
func fuenteDePrueba(t *testing.T, dependencias *dependenciasDePrueba) *Fuente {
	t.Helper()

	fuente, err := Nueva(ConCliente(dependencias.construir), ConCache(dependencias.abrir))
	require.NoError(t, err)

	return fuente
}

// resolverQueNoSeLlama es la resolución de una invocación que no debe llegar a
// resolver nada.
func resolverQueNoSeLlama(t *testing.T) func(context.Context, *invocacion) (schema.Resultado, error) {
	t.Helper()

	return func(context.Context, *invocacion) (schema.Resultado, error) {
		t.Error("la invocación ha llegado a resolver aunque no debía resolver nada")

		return schema.Resultado{}, nil
	}
}

// compruebaDefectoAlComponer exige el fallo de quien compone la fuente: un
// *Error sin dirección ni instante, de modo que el sobre lo firma y lo fecha el
// kernel, de clase «inesperado», código 1, y el mensaje entero.
func compruebaDefectoAlComponer(t *testing.T, err error, mensaje string) {
	t.Helper()

	var fallo *Error
	require.ErrorAs(t, err, &fallo)
	assert.Empty(t, fallo.URL)
	assert.Zero(t, fallo.Instante)
	assert.Equal(t, schema.ClaseInesperado, cli.Clasificar(err))
	assert.Equal(t, codigoDeLoInesperado, cli.CodigoSalida(err))
	assert.Equal(t, mensaje, err.Error())
}

// compruebaFalloDeLaCache exige el fallo de la caché de una invocación: un *Error
// con la dirección del recurso consultado y sin instante, la clase esperada
// —nunca la de identidad humana—, la causa alcanzable con errors.Is y el mensaje
// entero.
func compruebaFalloDeLaCache(t *testing.T, err, causa error, clase schema.Clase, mensaje string) {
	t.Helper()

	var fallo *Error
	require.ErrorAs(t, err, &fallo)
	require.ErrorIs(t, err, causa)
	assert.Equal(t, direccionInvocada, fallo.URL)
	assert.Zero(t, fallo.Instante)
	assert.Equal(t, clase, cli.Clasificar(err))
	assert.NotEqual(t, codigoDeIdentidadHumana, cli.CodigoSalida(err))
	assert.Equal(t, mensaje, err.Error())
}

// mensajeDeLaCache es el mensaje del fallo de la caché con ese motivo: la
// dirección detrás y, si la causa declara su clase, su texto.
func mensajeDeLaCache(motivo string, causa error, enElMensaje bool) string {
	mensaje := motivo + " (" + direccionInvocada + ")"
	if enElMensaje {
		mensaje += ": " + causa.Error()
	}

	return mensaje
}
