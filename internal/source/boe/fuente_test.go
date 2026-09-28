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

// TestCacheDeLosSeisVerbos fija la caché de los seis verbos (FR-090, FR-091,
// FR-096; SC-003; US2, escenarios 1, 6, 7 y 8) sobre la fuente compuesta con el
// cliente real de httpx en reproducción y la caché real en una carpeta temporal,
// las dos gobernadas por el mismo reloj de prueba:
//
//   - la segunda consulta idéntica, con el reloj adelantado hasta el último
//     nanosegundo de la vigencia de su verbo, la resuelve otra invocación sobre la
//     misma caché cuyo cliente reproduce una carpeta vacía, de modo que cualquier
//     petición emitida fallaría: no construye ningún cliente ni entrega ninguna
//     petición, y da el mismo data, la misma url y la misma fecha de consulta que
//     la primera;
//   - al cumplirse la vigencia, una tercera invocación sobre las grabaciones
//     vuelve a pedir las mismas peticiones, en su orden, y lleva la fecha de su
//     propia consulta;
//   - la búsqueda sin resultados es un resultado y se sirve igual (FR-032);
//   - articulo con la entrada de metadatos de su norma vigente pide solo el
//     bloque, y lleva la fecha de los metadatos, que es la más antigua;
//   - y metadatos tras un articulo no pide nada.
func TestCacheDeLosSeisVerbos(t *testing.T) {
	t.Parallel()

	for _, verbo := range consultasDeLosSeisVerbos() {
		t.Run(verbo.nombre, func(t *testing.T) {
			t.Parallel()

			guardado, sinGrabaciones := verbo.trasGuardar(t, verbo.vigencia-time.Nanosecond)

			servido, err := sinGrabaciones.resuelve(t, schema.Contexto{}, verbo.consulta)

			require.NoError(t, err)
			assert.Equal(t, guardado, servido, "la segunda consulta da el mismo data, la misma url y la misma fecha")
			sinGrabaciones.compruebaSinPedirNada(t)

			sinGrabaciones.reloj.adelanta(time.Nanosecond)

			pedido := verbo.resuelvePidiendo(t, sinGrabaciones.otroSobreLaMismaCache(t, reproduce(carpetaDeLasGrabaciones)))

			assert.Equal(t, guardado.Datos, pedido.Datos)
		})
	}

	t.Run("articulo-con-los-metadatos-en-cache", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
		t0 := banco.reloj.ahora()

		_, err := banco.resuelve(t, schema.Contexto{}, ConsultaMetadatos{Norma: normaVigente})
		require.NoError(t, err)

		banco.reloj.adelanta(vigenciaDeLosMetadatos - time.Nanosecond)

		delArticulo := banco.otroSobreLaMismaCache(t, reproduce(carpetaDeLasGrabaciones))

		resultado, err := delArticulo.resuelve(t, schema.Contexto{},
			ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21})

		compruebaResuelta(t, resultado, err,
			resultadoDelArticulo(direccionDelArticulo21, t0, eliDeLaLey39, articuloDelArticulo21(t)))
		delArticulo.compruebaPeticiones(t, peticionDelBloque(direccionDelArticulo21))
	})

	t.Run("metadatos-tras-un-articulo", func(t *testing.T) {
		t.Parallel()

		articulo := consultaDeUnVerbo{
			consulta:  ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21},
			direccion: direccionDelArticulo21,
			pedidas:   []httpx.Peticion{peticionDelBloque(direccionDelArticulo21), peticionDeLosMetadatos(metadatosVigente)},
		}

		guardado, sinGrabaciones := articulo.trasGuardar(t, vigenciaDeLosMetadatos-time.Nanosecond)

		resultado, err := sinGrabaciones.resuelve(t, schema.Contexto{}, ConsultaMetadatos{Norma: normaVigente})

		compruebaResuelta(t, resultado, err, resultadoResuelto(metadatosVigente, guardado.Procedencia.FechaConsulta,
			metadatosDeLaNormaVigente()))
		sinGrabaciones.compruebaSinPedirNada(t)
	})
}

// TestOfflineDeLosSeisVerbos fija --offline en los seis verbos (FR-092, FR-096;
// SC-004; US2, escenarios 3 y 4; contrato errores-y-codigos, fila 5), también
// junto con --dry-run, que conserva lo que da --offline, con cada consulta
// resuelta por una invocación cuyo cliente reproduce una carpeta vacía:
//
//   - con la entrada vigente, código 0 con el data, la url y la fecha de consulta
//     de la invocación que la guardó, aunque el reloj haya avanzado;
//   - con la entrada ausente —sin ninguna caché— o caducada, código 4 con la
//     dirección del recurso que falta —la del primer bloque en articulos— y sin
//     instante, porque el sobre lo fecha el montaje;
//   - y en todos los casos, sin construir ningún cliente ni entregar ninguna
//     petición, y con la caché intacta: la carpeta sin crear si no la había, y
//     la base con los mismos bytes si la había.
func TestOfflineDeLosSeisVerbos(t *testing.T) {
	t.Parallel()

	ejecuciones := map[string]schema.Contexto{
		"offline":          {Offline: true},
		"offline-y-ensayo": {Offline: true, DryRun: true},
	}

	for _, verbo := range consultasDeLosSeisVerbos() {
		t.Run(verbo.nombre, func(t *testing.T) {
			t.Parallel()

			for modo, ec := range ejecuciones {
				t.Run(modo+"-vigente", func(t *testing.T) {
					t.Parallel()

					compruebaOfflineConLaEntradaGuardada(t, verbo, ec, verbo.vigencia-time.Nanosecond)
				})

				t.Run(modo+"-caducada", func(t *testing.T) {
					t.Parallel()

					compruebaOfflineConLaEntradaGuardada(t, verbo, ec, verbo.vigencia)
				})

				t.Run(modo+"-ausente", func(t *testing.T) {
					t.Parallel()

					banco := nuevoBanco(t, reproduce(t.TempDir()))

					resultado, err := banco.resuelve(t, ec, verbo.consulta)

					compruebaFalloDeLaConsulta(t, resultado, err, verbo.sinEntrada)
					banco.compruebaSinPedirNada(t)
					banco.compruebaCacheSinCrear(t)
				})
			}
		})
	}
}

// compruebaOfflineConLaEntradaGuardada guarda la consulta del verbo, adelanta el
// reloj el tramo y la resuelve con la ejecución, que lleva --offline, en otra
// invocación sobre la misma caché: dentro de la vigencia, lo guardado tal cual;
// al cumplirse, el fallo del verbo sin entrada vigente. En los dos casos, sin
// pedir nada y con la caché intacta (FR-092).
func compruebaOfflineConLaEntradaGuardada(t *testing.T, verbo consultaDeUnVerbo, ec schema.Contexto, tramo time.Duration) {
	t.Helper()

	guardado, sinRed := verbo.trasGuardar(t, tramo)
	antes := sinRed.baseDeLaCache(t)

	resultado, err := sinRed.resuelve(t, ec, verbo.consulta)

	if tramo < verbo.vigencia {
		require.NoError(t, err)
		assert.Equal(t, guardado, resultado, "con la entrada vigente se sirve lo guardado, con su url y su fecha")
	} else {
		compruebaFalloDeLaConsulta(t, resultado, err, verbo.sinEntrada)
	}

	sinRed.compruebaSinPedirNada(t)
	assert.Equal(t, antes, sinRed.baseDeLaCache(t), "la caché queda intacta")
}

// TestEnsayoDeLosSeisVerbos fija --dry-run en los seis verbos (FR-094; ADR 0011;
// contrato errores-y-codigos, fila 23), sin caché y con la entrada caducada, con
// cada consulta resuelta por una invocación cuyo cliente reproduce una carpeta
// vacía: código 0; la procedencia con la url del verbo y sin fecha, sin datos; una
// línea por cada petición que se habría emitido, en su orden y sin repetir
// ninguna —la de los metadatos, una sola vez para todos los bloques de articulos,
// y la del bloque repetido, ninguna vez más—; el cliente las recibe todas en
// ensayo y no emite ninguna; y nada queda escrito: sin caché, la carpeta sigue
// sin crear, y con la entrada caducada, la base conserva sus bytes y la entrada
// sigue sin servirse con --offline.
func TestEnsayoDeLosSeisVerbos(t *testing.T) {
	t.Parallel()

	for _, verbo := range consultasDeLosSeisVerbos() {
		t.Run(verbo.nombre, func(t *testing.T) {
			t.Parallel()

			t.Run("sin-cache", func(t *testing.T) {
				t.Parallel()

				banco := nuevoBanco(t, reproduce(t.TempDir()))

				verbo.compruebaEnsayo(t, banco)
				banco.compruebaCacheSinCrear(t)
			})

			t.Run("con-la-entrada-caducada", func(t *testing.T) {
				t.Parallel()

				_, sinRed := verbo.trasGuardar(t, verbo.vigencia)
				antes := sinRed.baseDeLaCache(t)

				verbo.compruebaEnsayo(t, sinRed)
				assert.Equal(t, antes, sinRed.baseDeLaCache(t), "el ensayo no escribe nada")

				resultado, err := sinRed.resuelve(t, schema.Contexto{Offline: true}, verbo.consulta)

				compruebaFalloDeLaConsulta(t, resultado, err, verbo.sinEntrada)
			})
		})
	}
}

// TestFallosNoSeGuardan fija que ningún fallo queda en la caché (FR-093; SC-012;
// US2, escenario 5): el bloque inexistente, la fuente caída, el bloque ilegible y
// los metadatos caídos de un bloque obtenido fallan con su clase, con la
// dirección y el instante de la petición que falla y sin ningún dato —tampoco el
// texto del bloque obtenido—, y repetir la consulta vuelve a entregar al cliente
// las mismas peticiones, en su orden, y falla igual con el instante de la nueva
// petición: nada de lo que falló se escribió, ni la consulta ni los metadatos de
// los que dependía, y con --offline no hay ninguna entrada que servir.
func TestFallosNoSeGuardan(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		pedidor  func(*testing.T, *relojDePrueba) *pedidorDePrueba
		consulta core.Consulta
		// pedidas son las peticiones que la consulta entrega al cliente hasta
		// fallar, en su orden.
		pedidas []httpx.Peticion
		// direccion es la de la petición que falla, y clase, la de su fallo.
		direccion string
		clase     schema.Clase
		// sinEntrada es el fallo de la consulta con --offline.
		sinEntrada falloDeLaConsulta
	}{
		{
			nombre:     "bloque-inexistente",
			pedidor:    reproduce(carpetaDeLasGrabaciones),
			consulta:   ConsultaArticulo{Norma: normaVigente, Bloque: bloqueInexistente},
			pedidas:    []httpx.Peticion{peticionDelBloque(direccionDelBloqueInexistente)},
			direccion:  direccionDelBloqueInexistente,
			clase:      schema.ClaseNoEncontrado,
			sinEntrada: falloSinEntradaDelArticulo(direccionDelBloqueInexistente),
		},
		{
			nombre:     "fuente-caida",
			pedidor:    reproduce(sinteticoDeLaFuenteCaida),
			consulta:   ConsultaIndice{Norma: normaVigente},
			pedidas:    []httpx.Peticion{{Metodo: "GET", URL: indiceVigente, Acepta: aceptaDelResto}},
			direccion:  indiceVigente,
			clase:      schema.ClaseFuenteNoDisponible,
			sinEntrada: falloSinIndiceConOffline(indiceVigente),
		},
		{
			nombre:     "bloque-ilegible",
			pedidor:    reproduce(sinteticoDelBloqueIlegible),
			consulta:   ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21},
			pedidas:    []httpx.Peticion{peticionDelBloque(direccionDelArticulo21)},
			direccion:  direccionDelArticulo21,
			clase:      schema.ClaseFuenteNoDisponible,
			sinEntrada: falloSinEntradaDelArticulo(direccionDelArticulo21),
		},
		{
			nombre:     "metadatos-caidos",
			pedidor:    reproduce(sinteticoDeLosMetadatosCaidos),
			consulta:   ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21},
			pedidas:    []httpx.Peticion{peticionDelBloque(direccionDelArticulo21), peticionDeLosMetadatos(metadatosVigente)},
			direccion:  metadatosVigente,
			clase:      schema.ClaseFuenteNoDisponible,
			sinEntrada: falloSinEntradaDelArticulo(direccionDelArticulo21),
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, caso.pedidor)
			pedidas := make([]httpx.Peticion, 0, 2*len(caso.pedidas))

			for vez := 1; vez <= 2; vez++ {
				resultado, err := banco.resuelve(t, schema.Contexto{}, caso.consulta)

				require.Error(t, err, "la consulta %d", vez)
				assert.Equal(t, caso.clase, cli.Clasificar(err), "la consulta %d", vez)
				assert.Equal(t, schema.Resultado{Procedencia: schema.Procedencia{
					Fuente:        NombreDeLaFuente,
					URL:           caso.direccion,
					FechaConsulta: banco.reloj.ahora(),
				}}, resultado, "la consulta %d falla sin datos y con el instante de su propia petición", vez)

				pedidas = append(pedidas, caso.pedidas...)
				banco.compruebaPeticiones(t, pedidas...)

				resultado, err = banco.resuelve(t, schema.Contexto{Offline: true}, caso.consulta)

				compruebaFalloDeLaConsulta(t, resultado, err, caso.sinEntrada)

				banco.reloj.adelanta(time.Second)
			}
		})
	}
}

// consultaDeUnVerbo es la consulta de uno de los seis verbos con la que se fijan
// la caché, --offline y el ensayo, con lo que da sobre las grabaciones.
type consultaDeUnVerbo struct {
	nombre   string
	consulta core.Consulta
	// direccion es la url del sobre del verbo, y vigencia, la de su entrada.
	direccion string
	vigencia  time.Duration
	// pedidas son las peticiones que la consulta entrega al cliente sin nada en
	// la caché, en su orden y sin repetir ninguna.
	pedidas []httpx.Peticion
	// sinEntrada es su fallo con --offline sin ninguna entrada vigente.
	sinEntrada falloDeLaConsulta
}

// consultasDeLosSeisVerbos son las consultas de los seis verbos: la búsqueda de
// «procedimiento administrativo común» y, de la Ley 39/2015, el índice, el
// artículo 21, varios artículos, los metadatos y el análisis; y además la
// búsqueda sin resultados, que se guarda y se sirve como cualquier otro resultado
// (FR-032). La de articulos repite un bloque, que se pide una sola vez, y sus
// tres bloques distintos comparten una sola petición de metadatos (FR-020).
func consultasDeLosSeisVerbos() []consultaDeUnVerbo {
	enJSON := func(direccion string) httpx.Peticion {
		return httpx.Peticion{Metodo: "GET", URL: direccion, Acepta: aceptaDelResto}
	}

	sinEntradaDeArticulos := falloSinEntradaDelArticulo(direccionDelArticulo21)
	sinEntradaDeArticulos.mensaje = "no se ha podido resolver el bloque a21, en la posición 1 de 4: " +
		sinEntradaDeArticulos.mensaje

	return []consultaDeUnVerbo{
		{
			nombre:     "buscar",
			consulta:   ConsultaBuscar{Texto: []string{"procedimiento", "administrativo", "común"}},
			direccion:  busquedaConResultados,
			vigencia:   vigenciaDeLaBusqueda,
			pedidas:    []httpx.Peticion{enJSON(busquedaConResultados)},
			sinEntrada: falloSinBusquedaConOffline(busquedaConResultados),
		},
		{
			nombre:     "buscar-sin-resultados",
			consulta:   ConsultaBuscar{Texto: []string{"zzqxkwvjh"}},
			direccion:  busquedaSinResultados,
			vigencia:   vigenciaDeLaBusqueda,
			pedidas:    []httpx.Peticion{enJSON(busquedaSinResultados)},
			sinEntrada: falloSinBusquedaConOffline(busquedaSinResultados),
		},
		{
			nombre:     "indice",
			consulta:   ConsultaIndice{Norma: normaVigente},
			direccion:  indiceVigente,
			vigencia:   vigenciaDelIndice,
			pedidas:    []httpx.Peticion{enJSON(indiceVigente)},
			sinEntrada: falloSinIndiceConOffline(indiceVigente),
		},
		{
			nombre:     "articulo",
			consulta:   ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21},
			direccion:  direccionDelArticulo21,
			vigencia:   vigenciaDeLosArticulos,
			pedidas:    []httpx.Peticion{peticionDelBloque(direccionDelArticulo21), enJSON(metadatosVigente)},
			sinEntrada: falloSinEntradaDelArticulo(direccionDelArticulo21),
		},
		{
			nombre: "articulos",
			consulta: ConsultaArticulos{
				Norma:   normaVigente,
				Bloques: []string{bloqueDelArticulo21, bloqueDelArticulo22, bloqueDelArticulo21, bloqueDelArticulo23},
			},
			direccion: direccionDeLaNormaVigente,
			vigencia:  vigenciaDeLosArticulos,
			pedidas: []httpx.Peticion{
				peticionDelBloque(direccionDelArticulo21),
				enJSON(metadatosVigente),
				peticionDelBloque(direccionDelArticulo22),
				peticionDelBloque(direccionDelArticulo23),
			},
			sinEntrada: sinEntradaDeArticulos,
		},
		{
			nombre:     "metadatos",
			consulta:   ConsultaMetadatos{Norma: normaVigente},
			direccion:  metadatosVigente,
			vigencia:   vigenciaDeLosMetadatos,
			pedidas:    []httpx.Peticion{enJSON(metadatosVigente)},
			sinEntrada: falloSinEntradaConOffline(metadatosVigente),
		},
		{
			nombre:     "analisis",
			consulta:   ConsultaAnalisis{Norma: normaVigente},
			direccion:  analisisVigente,
			vigencia:   vigenciaDelAnalisis,
			pedidas:    []httpx.Peticion{enJSON(analisisVigente)},
			sinEntrada: falloSinAnalisisConOffline(analisisVigente),
		},
	}
}

// resuelvePidiendo resuelve la consulta del verbo con el banco, sin nada vigente
// en su caché, y exige que la resuelva entregando al cliente exactamente sus
// peticiones, en su orden, con la url del verbo, la fecha del reloj —que no se
// mueve mientras se resuelve— y datos; devuelve lo resuelto.
func (v consultaDeUnVerbo) resuelvePidiendo(t *testing.T, banco *bancoDeLaFuente) schema.Resultado {
	t.Helper()

	resultado, err := banco.resuelve(t, schema.Contexto{}, v.consulta)

	require.NoError(t, err)
	compruebaDireccionDeLaFuente(t, resultado.Procedencia.URL)
	assert.Equal(t, schema.Procedencia{Fuente: NombreDeLaFuente, URL: v.direccion, FechaConsulta: banco.reloj.ahora()},
		resultado.Procedencia)
	assert.NotNil(t, resultado.Datos)
	assert.Empty(t, resultado.Ensayo)
	banco.compruebaPeticiones(t, v.pedidas...)

	return resultado
}

// trasGuardar resuelve la consulta del verbo con resuelvePidiendo en un banco
// nuevo sobre las grabaciones, adelanta el reloj el tramo y devuelve lo resuelto
// y otra invocación sobre la misma caché y el mismo reloj cuyo cliente reproduce
// una carpeta vacía, de modo que cualquier petición que emitiera fallaría.
func (v consultaDeUnVerbo) trasGuardar(t *testing.T, tramo time.Duration) (schema.Resultado, *bancoDeLaFuente) {
	t.Helper()

	banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
	guardado := v.resuelvePidiendo(t, banco)

	banco.reloj.adelanta(tramo)

	return guardado, banco.otroSobreLaMismaCache(t, reproduce(t.TempDir()))
}

// compruebaEnsayo resuelve la consulta del verbo con --dry-run y exige el código
// 0 y el resultado del ensayo —la procedencia con la url del verbo y sin fecha,
// sin datos, y la línea de cada petición del verbo, en su orden y sin repetir
// ninguna—, con cada petición entregada al cliente en ensayo, sin emitirla.
func (v consultaDeUnVerbo) compruebaEnsayo(t *testing.T, banco *bancoDeLaFuente) {
	t.Helper()

	lineas := make([]string, 0, len(v.pedidas))
	for _, pedida := range v.pedidas {
		lineas = append(lineas, pedida.Metodo+" "+pedida.URL)
	}

	resultado, err := banco.resuelve(t, schema.Contexto{DryRun: true}, v.consulta)

	require.NoError(t, err)
	assert.Equal(t, schema.Resultado{
		Procedencia: schema.Procedencia{Fuente: NombreDeLaFuente, URL: v.direccion},
		Ensayo:      lineas,
	}, resultado)
	assert.Len(t, slices.Compact(slices.Sorted(slices.Values(resultado.Ensayo))), len(resultado.Ensayo),
		"ninguna línea del ensayo se repite: %v", resultado.Ensayo)
	banco.compruebaPeticiones(t, v.pedidas...)

	for indice, respuesta := range banco.pedidor.respuestas {
		assert.True(t, respuesta.Ensayo, "bajo --dry-run no se emite la petición %d", indice+1)
	}
}

// otroSobreLaMismaCache es otro banco con la misma caché y el mismo reloj que
// este, y con el Pedidor que construye pedidor: la fuente de otra invocación, que
// encuentra lo que dejaron escrito las anteriores.
func (b *bancoDeLaFuente) otroSobreLaMismaCache(t *testing.T,
	pedidor func(*testing.T, *relojDePrueba) *pedidorDePrueba,
) *bancoDeLaFuente {
	t.Helper()

	otro := &bancoDeLaFuente{pedidor: pedidor(t, b.reloj), reloj: b.reloj, carpetaDeLaCache: b.carpetaDeLaCache}

	fuente, err := Nueva(
		ConCliente(otro.construir),
		ConCache(abrirCacheEn(otro.carpetaDeLaCache, cache.ConReloj(otro.reloj.ahora))),
	)
	require.NoError(t, err)

	otro.fuente = fuente

	return otro
}

// compruebaSinPedirNada exige que la fuente del banco no haya construido ningún
// cliente ni entregado ninguna petición.
func (b *bancoDeLaFuente) compruebaSinPedirNada(t *testing.T) {
	t.Helper()

	assert.Zero(t, b.construcciones, "no se construye ningún cliente")
	assert.Empty(t, b.pedidor.peticiones, "no se entrega ninguna petición")
}

// ficheroDeLaBase es el fichero de la base en la carpeta de la caché. Una
// invocación de solo lectura lo deja idéntico byte a byte, y los auxiliares
// cache.db-wal y cache.db-shm, que SQLite crea también al leer en un directorio
// escribible y que solo retira el cierre de un cliente normal, no cuentan
// (contrato de H3 esquema-y-apertura §6, garantía de SC-003, y §7).
const ficheroDeLaBase = "cache.db"

// baseDeLaCache son los bytes de la base de la caché del banco, que una
// invocación que no escribe nada deja iguales. Exige que la base exista y no esté
// vacía: sin base no hay nada que comparar.
func (b *bancoDeLaFuente) baseDeLaCache(t *testing.T) []byte {
	t.Helper()

	contenido, err := os.ReadFile(filepath.Clean(filepath.Join(b.carpetaDeLaCache, ficheroDeLaBase)))
	require.NoError(t, err)
	require.NotEmpty(t, contenido, "la base de la caché está vacía")

	return contenido
}

// verbosDeLaFuente son los seis verbos de la fuente, los que la lista cerrada de
// golden tiene que cubrir (FR-112, SC-005).
var verbosDeLaFuente = []string{verboBuscar, verboIndice, verboArticulo, verboArticulos, verboMetadatos, verboAnalisis}

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

		assert.ElementsMatch(t, verbosDeLaFuente, slices.Collect(maps.Keys(verbos)))
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

// TestGoldenCubreTodosLosCasos exige en testdata/golden exactamente los trece
// golden de la lista cerrada, que entre todos cubren los seis verbos (FR-112,
// SC-005; contrato esquemas-fixtures-y-controles §2): el golden que falta, lo que
// no es el golden de ningún caso y el verbo que ningún golden cubre hacen fallar
// la prueba nombrándolos. No compara ningún contenido: eso lo hace TestGolden, que
// sin esta prueba pasaría con la carpeta vacía.
//
// El primer subtest comprueba los golden reales; el resto demuestra sobre
// carpetas temporales que la comprobación no pasa en vacío.
func TestGoldenCubreTodosLosCasos(t *testing.T) {
	t.Parallel()

	t.Run("testdata-golden", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, comprobarGoldenCompletos(carpetaDeLosGolden, casosDeGolden, verbosDeLaFuente))
	})

	const (
		analisis  = "analisis-BOE-A-2015-10565"
		indice    = "indice-BOE-A-2015-10565"
		metadatos = "metadatos-BOE-A-1992-26318"
		falta     = "falta el golden "
		sinCubrir = "ningún golden cubre el verbo "
		ajeno     = " no es el golden de ningún caso"
	)

	todos := make([]string, 0, len(casosDeGolden))
	for _, caso := range casosDeGolden {
		todos = append(todos, caso.nombre)
	}

	sinElAnalisis := slices.DeleteFunc(slices.Clone(casosDeGolden), func(caso casoDeGolden) bool {
		return caso.consulta.Verbo() == verboAnalisis
	})

	casos := []struct {
		nombre string
		// lista es la de los casos contra la que se comprueba; nula, la cerrada.
		lista []casoDeGolden
		// sinCarpeta no crea la carpeta de los golden, y comoFichero deja un
		// fichero en su lugar.
		sinCarpeta, comoFichero bool
		// sinGolden son los casos de la lista cuyo golden no se escribe; ajenos,
		// los ficheros que se dejan además en la carpeta, y carpetas, las carpetas.
		sinGolden, ajenos, carpetas []string
		// mensajes son fragmentos del error esperado, y ausentes, fragmentos que no
		// puede llevar; sin mensajes, los golden valen.
		mensajes, ausentes []string
	}{
		{nombre: "los-trece"},
		{nombre: "sin-carpeta", sinCarpeta: true, mensajes: []string{"la carpeta de los golden no se puede leer"}},
		{nombre: "carpeta-que-es-un-fichero", comoFichero: true, mensajes: []string{"la carpeta de los golden no se puede leer"}},
		{
			nombre:    "carpeta-vacia",
			sinGolden: todos,
			mensajes: []string{
				falta + "buscar-procedimiento-administrativo-comun.json", falta + analisis + ".json",
				sinCubrir + verboBuscar, sinCubrir + verboAnalisis,
			},
		},
		{
			nombre:    "uno-de-menos",
			sinGolden: []string{metadatos},
			mensajes:  []string{falta + metadatos + ".json"},
			ausentes:  []string{sinCubrir, ajeno},
		},
		{
			nombre:    "uno-de-menos-que-deja-su-verbo-sin-cubrir",
			sinGolden: []string{analisis},
			mensajes:  []string{falta + analisis + ".json", sinCubrir + verboAnalisis},
			ausentes:  []string{ajeno},
		},
		{
			nombre:   "uno-de-mas",
			ajenos:   []string{"articulo-BOE-A-2015-10565-a5.json"},
			mensajes: []string{"articulo-BOE-A-2015-10565-a5.json" + ajeno},
			ausentes: []string{falta, sinCubrir},
		},
		{
			nombre:   "fichero-que-no-es-un-golden",
			ajenos:   []string{"LEEME.md"},
			mensajes: []string{"LEEME.md" + ajeno},
			ausentes: []string{falta, sinCubrir},
		},
		{
			nombre:    "carpeta-con-el-nombre-de-un-golden",
			sinGolden: []string{indice},
			carpetas:  []string{indice + ".json"},
			mensajes:  []string{indice + ".json" + ajeno, falta + indice + ".json", sinCubrir + verboIndice},
		},
		{
			nombre:   "lista-que-no-cubre-un-verbo",
			lista:    sinElAnalisis,
			mensajes: []string{sinCubrir + verboAnalisis},
			ausentes: []string{falta, ajeno},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			lista := caso.lista
			if lista == nil {
				lista = casosDeGolden
			}

			carpeta := filepath.Join(t.TempDir(), "golden")

			switch {
			case caso.comoFichero:
				escribeDatoDePrueba(t, carpeta, []byte("{}\n"))
			case !caso.sinCarpeta:
				escribeCarpetaDeGoldenDePrueba(t, carpeta, lista, caso.sinGolden, caso.ajenos, caso.carpetas)
			}

			err := comprobarGoldenCompletos(carpeta, lista, verbosDeLaFuente)
			if len(caso.mensajes) == 0 {
				require.NoError(t, err)

				return
			}

			for _, mensaje := range caso.mensajes {
				require.ErrorContains(t, err, mensaje)
			}

			for _, ausente := range caso.ausentes {
				assert.NotContains(t, err.Error(), ausente)
			}
		})
	}
}

// escribeCarpetaDeGoldenDePrueba crea la carpeta de los golden con el golden de
// prueba de cada caso de la lista salvo los de sinGolden, un fichero por cada
// nombre de ajenos y una carpeta por cada nombre de carpetas.
func escribeCarpetaDeGoldenDePrueba(t *testing.T, carpeta string, lista []casoDeGolden,
	sinGolden, ajenos, carpetas []string,
) {
	t.Helper()

	require.NoError(t, os.MkdirAll(carpeta, 0o750))

	for _, caso := range lista {
		if !slices.Contains(sinGolden, caso.nombre) {
			escribeGolden(t, carpeta, caso, goldenDePrueba(caso))
		}
	}

	for _, nombre := range ajenos {
		escribeDatoDePrueba(t, filepath.Join(carpeta, nombre), []byte("{}\n"))
	}

	for _, nombre := range carpetas {
		require.NoError(t, os.Mkdir(filepath.Join(carpeta, nombre), 0o750))
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
	conGolden, fallos, err := goldenDeLaCarpeta(carpeta, casos)

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return err
	}

	for _, caso := range conGolden {
		fallos = append(fallos, comprobarUnGolden(ficheroDelGolden(carpeta, caso), caso, generar))
	}

	return errors.Join(fallos...)
}

// comprobarGoldenCompletos exige en la carpeta de los golden exactamente el de
// cada caso y nada más, y que entre todos cubran los verbos: el golden que falta,
// lo que no es el golden de ningún caso y el verbo que ningún golden cubre son
// errores que los nombran, y una carpeta que no se puede leer, también la que no
// existe, no vale. No compara ningún contenido: eso lo hace comprobarGolden.
func comprobarGoldenCompletos(carpeta string, casos []casoDeGolden, verbos []string) error {
	conGolden, fallos, err := goldenDeLaCarpeta(carpeta, casos)
	if err != nil {
		return err
	}

	cubiertos := make(map[string]bool, len(verbos))
	for _, caso := range conGolden {
		cubiertos[caso.consulta.Verbo()] = true
	}

	for _, caso := range casos {
		if !slices.ContainsFunc(conGolden, func(enLaCarpeta casoDeGolden) bool { return enLaCarpeta.nombre == caso.nombre }) {
			fallos = append(fallos, fmt.Errorf("%s: falta el golden %s", carpeta, caso.nombre+extensionDeLosGolden))
		}
	}

	for _, verbo := range verbos {
		if !cubiertos[verbo] {
			fallos = append(fallos, fmt.Errorf("%s: ningún golden cubre el verbo %s", carpeta, verbo))
		}
	}

	return errors.Join(fallos...)
}

// goldenDeLaCarpeta lee la carpeta de los golden y devuelve, en el orden de la
// carpeta, los casos cuyo golden está en ella y un error por cada entrada que no
// es el golden de ningún caso, que la nombra: un fichero sin caso, o una carpeta
// aunque lleve el nombre de un golden. Si la carpeta no se puede leer, el error
// lleva la causa, alcanzable con errors.Is.
func goldenDeLaCarpeta(carpeta string, casos []casoDeGolden) ([]casoDeGolden, []error, error) {
	entradas, err := os.ReadDir(filepath.Clean(carpeta))
	if err != nil {
		return nil, nil, fmt.Errorf("la carpeta de los golden no se puede leer: %w", err)
	}

	conGolden := make([]casoDeGolden, 0, len(entradas))

	var ajenos []error

	for _, entrada := range entradas {
		indice := slices.IndexFunc(casos, func(caso casoDeGolden) bool {
			return caso.nombre+extensionDeLosGolden == entrada.Name()
		})
		if indice < 0 || !entrada.Type().IsRegular() {
			ajenos = append(ajenos, fmt.Errorf("%s: %s no es el golden de ningún caso", carpeta, entrada.Name()))

			continue
		}

		conGolden = append(conGolden, casos[indice])
	}

	return conGolden, ajenos, nil
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

// abrirCacheEn es la AperturaDeCache de la caché real en la carpeta, con las
// opciones que se le den —el reloj de la prueba, por ejemplo— y en solo lectura
// cuando la fuente lo pide.
func abrirCacheEn(carpeta string, adicionales ...cache.Opcion) AperturaDeCache {
	return func(ctx context.Context, soloLectura bool) (CacheAbierta, error) {
		opciones := slices.Concat([]cache.Opcion{cache.ConDirectorio(carpeta)}, adicionales)
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
