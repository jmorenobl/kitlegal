package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// Los datos de prueba de la fuente del BOE vistos desde este paquete: las
// grabaciones reales y los sintéticos que fija una persona (contrato
// esquemas-fixtures-y-controles §3 y §4). Estas pruebas los leen y no escriben
// nada en ellos, y httpx.Replay los sirve sin abrir ninguna conexión.
const (
	// grabacionesDeBoe es la carpeta de las grabaciones, <raíz>/<fuente>.
	grabacionesDeBoe = "../source/boe/testdata/" + boe.NombreDeLaFuente
	// sinteticosDeBoe es la de los sintéticos, un escenario por subcarpeta, cada
	// uno con la carpeta de la fuente dentro.
	sinteticosDeBoe = "../source/boe/testdata/sintetico"
)

// normaDeBoe es la Ley 39/2015, la norma de la entrega del hito, y las
// direcciones de la API que citan los sobres de estas pruebas, escritas enteras y
// no con el código de la fuente, para que un cambio en él no pase por aquí en
// silencio (contrato verbos-y-salidas §1 a §6).
const (
	normaDeBoe              = "BOE-A-2015-10565"
	direccionDeLaNorma      = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565"
	direccionDelBloqueA21   = direccionDeLaNorma + "/texto/bloque/a21"
	direccionDelBloqueA22   = direccionDeLaNorma + "/texto/bloque/a22"
	direccionDelBloqueA9999 = direccionDeLaNorma + "/texto/bloque/a9999"
	direccionDeLosMetadatos = direccionDeLaNorma + "/metadatos"
	direccionDelIndice      = direccionDeLaNorma + "/texto/indice"
	direccionDelAnalisis    = direccionDeLaNorma + "/analisis"
	direccionDeLaBusqueda   = "https://www.boe.es/datosabiertos/api/legislacion-consolidada?limit=10&query=" +
		"%7B%22query%22%3A%20%7B%22query_string%22%3A%20%7B%22query%22%3A%20%22titulo%3Aprocedimiento%20AND%20" +
		"titulo%3Aadministrativo%20AND%20titulo%3Acom%5Cu00fan%22%7D%7D%7D"
)

// eventoDePeticionEmitida es el evento con el que httpx anota, en su registrador,
// cada petición que emite —también la que sirve una grabación—, con su dirección.
// Es lo que estas pruebas cuentan para saber qué se pidió.
const eventoDePeticionEmitida = "petición emitida"

// inicioDelReloj y pasoDelReloj gobiernan la hora de estas pruebas. La caché vive
// siempre en inicioDelReloj, de modo que todo lo que se guarda sigue vigente; la
// hora de emisión de httpx avanza un paso en cada petición, de modo que cada una
// tiene su instante y el de la n-ésima es inicioDelReloj + n pasos. El inicio es
// anterior a cualquier ejecución de las pruebas, así que un sobre fechado al
// montarse nunca se confunde con uno fechado por una petición.
var inicioDelReloj = time.Date(2025, time.March, 4, 9, 15, 0, 0, time.UTC)

const pasoDelReloj = time.Second

// ahoraDeLaCache es el reloj de la caché de estas pruebas.
func ahoraDeLaCache() time.Time {
	return inicioDelReloj
}

// instanteDeLaPeticion es el instante que declara la n-ésima petición emitida
// por el cliente de un banco, contando desde 1.
func instanteDeLaPeticion(n int) time.Time {
	return inicioDelReloj.Add(time.Duration(n) * pasoDelReloj)
}

// horaQueAvanza es la hora de emisión del cliente de un banco: httpx la lee una
// vez por petición emitida, justo antes de entregarla al transporte (contrato
// httpx-acepta-e-instante §4), y cada lectura avanza un paso.
type horaQueAvanza struct {
	lecturas atomic.Int64
}

func (h *horaQueAvanza) ahora() time.Time {
	return instanteDeLaPeticion(int(h.lecturas.Add(1)))
}

// eventosDeHTTPX recoge en JSON los eventos del cliente de un banco. Es un
// io.Writer protegido, porque el manejador de slog escribe cada evento de una
// sola vez pero no promete hacerlo desde una sola goroutine.
type eventosDeHTTPX struct {
	mu     sync.Mutex
	lineas bytes.Buffer
}

func (e *eventosDeHTTPX) Write(p []byte) (int, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.lineas.Write(p)
}

// registrador es el registrador que se da al cliente, con todos los niveles.
func (e *eventosDeHTTPX) registrador() *slog.Logger {
	return slog.New(slog.NewJSONHandler(e, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// pedidas son las direcciones de las peticiones emitidas hasta ahora, en su
// orden.
func (e *eventosDeHTTPX) pedidas(t *testing.T) []string {
	t.Helper()

	e.mu.Lock()
	defer e.mu.Unlock()

	var pedidas []string

	decodificador := json.NewDecoder(bytes.NewReader(e.lineas.Bytes()))

	for {
		var evento struct {
			Mensaje string `json:"msg"`
			URL     string `json:"url"`
		}

		err := decodificador.Decode(&evento)
		if errors.Is(err, io.EOF) {
			return pedidas
		}

		require.NoError(t, err, "cada evento del cliente es un documento JSON")

		if evento.Mensaje == eventoDePeticionEmitida {
			pedidas = append(pedidas, evento.URL)
		}
	}
}

// bancoDeBoe es el applet boe de una prueba con lo que lo rodea: el cliente de
// httpx en reproducción de una carpeta, con la hora que avanza y los eventos que
// cuentan las peticiones; la caché real en una carpeta temporal con su reloj; y
// cuántas veces se construyó el cliente. El cliente de la prueba no usa el
// registrador que le entrega el kernel, sino el suyo, para poder contar lo que
// pide sin depender del nivel del registro de eventos de quien ejecuta los tests.
type bancoDeBoe struct {
	dependencias     DependenciasDeBoe
	carpetaDeLaCache string
	hora             horaQueAvanza
	eventos          eventosDeHTTPX
	construcciones   atomic.Int64
}

// nuevoBancoDeBoe compone el banco sobre la carpeta de reproducción.
func nuevoBancoDeBoe(t *testing.T, carpetaDeReproduccion string) *bancoDeBoe {
	t.Helper()

	banco := &bancoDeBoe{carpetaDeLaCache: t.TempDir()}
	banco.dependencias = DependenciasDeBoe{
		Cliente: func(*slog.Logger) (*httpx.Cliente, error) {
			banco.construcciones.Add(1)

			return httpx.Replay(carpetaDeReproduccion,
				httpx.ConFuente(boe.NombreDeLaFuente),
				httpx.ConRegistrador(banco.eventos.registrador()),
				httpx.ConHora(banco.hora.ahora))
		},
		Cache: []cache.Opcion{cache.ConDirectorio(banco.carpetaDeLaCache), cache.ConReloj(ahoraDeLaCache)},
	}

	return banco
}

// sintetico es la carpeta de reproducción del escenario sintético.
func sintetico(escenario string) string {
	return filepath.Join(sinteticosDeBoe, escenario, boe.NombreDeLaFuente)
}

// cacheEnUnFichero hace que la caché del banco se declare sobre un fichero
// regular, que cache.New rechaza al abrirla con «argumentos» (contrato de H3,
// esquema-y-apertura; errores-y-codigos de H4, fila 21).
func (b *bancoDeBoe) cacheEnUnFichero(t *testing.T) {
	t.Helper()

	fichero := filepath.Join(t.TempDir(), "cache-que-no-es-una-carpeta")
	require.NoError(t, os.WriteFile(fichero, nil, 0o600))

	b.dependencias.Cache = []cache.Opcion{cache.ConDirectorio(fichero), cache.ConReloj(ahoraDeLaCache)}
}

// invocacionDeBoe es lo observable de una invocación sobre el banco: lo de
// cualquier invocación del kernel, el reloj del sistema justo antes y justo
// después, las peticiones que emitió y cuántas veces construyó el cliente.
type invocacionDeBoe struct {
	invocacionDePrueba

	antes, despues time.Time
	pedidas        []string
	construcciones int64
}

// invocar ejecuta el kernel entero con el applet boe del banco registrado.
func (b *bancoDeBoe) invocar(t *testing.T, argv ...string) invocacionDeBoe {
	t.Helper()

	var registro Registro

	require.NoError(t, registro.Registrar(AppletBoe(b.dependencias)))

	previas, construidas := len(b.eventos.pedidas(t)), b.construcciones.Load()

	antes := time.Now()
	res := invocar(t, &registro, argv...)
	despues := time.Now()

	pedidas := b.eventos.pedidas(t)[previas:]
	if len(pedidas) == 0 {
		pedidas = nil
	}

	return invocacionDeBoe{
		invocacionDePrueba: res,
		antes:              antes,
		despues:            despues,
		pedidas:            pedidas,
		construcciones:     b.construcciones.Load() - construidas,
	}
}

// compruebaCarpetaVacia exige que la carpeta de la caché siga vacía: nada la
// abrió, o se abrió en solo lectura y no creó nada (FR-092, FR-094).
func compruebaCarpetaVacia(t *testing.T, carpeta string) {
	t.Helper()

	entradas, err := os.ReadDir(carpeta)
	require.NoError(t, err)
	assert.Empty(t, entradas, "la caché no se ha creado")
}

// fechaDelSobre es la fecha de consulta de un sobre.
func fechaDelSobre(t *testing.T, sobre map[string]any) time.Time {
	t.Helper()

	texto, esTexto := sobre["fecha_consulta"].(string)
	require.True(t, esTexto, "fecha_consulta es un texto")

	fecha, err := time.Parse(time.RFC3339Nano, texto)
	require.NoError(t, err, "fecha_consulta está en RFC 3339")

	return fecha
}

// argvDeBoe es la invocación de «kitlegal boe» con los argumentos.
func argvDeBoe(argumentos ...string) []string {
	return slices.Concat([]string{"kitlegal", "boe"}, argumentos)
}

// verboDelContrato es lo que el contrato puerto-y-applet §4 fija de un verbo, y
// una invocación suya sobre las grabaciones con la consulta que tiene que
// construir y la dirección que cita su sobre (contrato verbos-y-salidas), más
// las invocaciones con las que falla y la que devuelve avisos.
type verboDelContrato struct {
	nombre    string
	campos    []campoDelContrato
	salida    any
	argumento []string
	consulta  core.Consulta
	url       string
	// fueraDeGramatica es una invocación que termina con código 2 sin pedir
	// nada: un argumento fuera de su gramática (contrato errores-y-codigos,
	// filas 2-4).
	fueraDeGramatica []string
	// inexistente es una invocación sobre las grabaciones que termina con
	// código 3: un recurso que la fuente no tiene (filas 6 y 7). buscar no
	// tiene ninguna: una búsqueda sin resultados da una lista vacía, y un 404
	// en buscar es el código 4 (fila 9).
	inexistente []string
	// conAvisos es, en los verbos que devuelven avisos (FR-012), una
	// invocación sobre el sintético avisos cuyo data lleva los tres.
	conAvisos []string
}

// normaInexistente es la norma grabada que la fuente no tiene, y
// normaFueraDeGramatica un identificador que no llega a pedirse (contrato
// esquemas-fixtures-y-controles §3.1, filas 19, 20 y 22; contrato
// errores-y-codigos, fila 2).
const (
	normaInexistente      = "BOE-A-2099-99999"
	normaFueraDeGramatica = "BOE-A-2015"
)

// campoDelContrato es un argumento del verbo, de los que van por su posición: el
// campo, su tipo y el nombre con el que Kong lo presenta.
type campoDelContrato struct {
	campo  string
	nombre string
	tipo   reflect.Type
}

// verbosDelContrato son los seis verbos, en el orden del contrato.
func verbosDelContrato() []verboDelContrato {
	var (
		texto  = reflect.TypeFor[string]()
		textos = reflect.TypeFor[[]string]()
		norma  = campoDelContrato{campo: "Norma", nombre: "norma", tipo: texto}
	)

	return []verboDelContrato{
		{
			nombre:           "buscar",
			campos:           []campoDelContrato{{campo: "Texto", nombre: "texto", tipo: textos}},
			salida:           []boe.ResultadoDeBusqueda(nil),
			argumento:        []string{"buscar", "procedimiento", "administrativo", "común"},
			consulta:         boe.ConsultaBuscar{Texto: []string{"procedimiento", "administrativo", "común"}},
			url:              direccionDeLaBusqueda,
			fueraDeGramatica: []string{"buscar", ""},
		},
		{
			nombre:           "indice",
			campos:           []campoDelContrato{norma},
			salida:           boe.Indice{},
			argumento:        []string{"indice", normaDeBoe},
			consulta:         boe.ConsultaIndice{Norma: normaDeBoe},
			url:              direccionDelIndice,
			fueraDeGramatica: []string{"indice", normaFueraDeGramatica},
			inexistente:      []string{"indice", normaInexistente},
		},
		{
			nombre:           "articulo",
			campos:           []campoDelContrato{norma, {campo: "Bloque", nombre: "bloque", tipo: texto}},
			salida:           boe.Articulo{},
			argumento:        []string{"articulo", normaDeBoe, "a21"},
			consulta:         boe.ConsultaArticulo{Norma: normaDeBoe, Bloque: "a21"},
			url:              direccionDelBloqueA21,
			fueraDeGramatica: []string{"articulo", normaDeBoe, "../a21"},
			inexistente:      []string{"articulo", normaDeBoe, "a9999"},
			conAvisos:        []string{"articulo", normaDeBoe, "a21"},
		},
		{
			nombre:           "articulos",
			campos:           []campoDelContrato{norma, {campo: "Bloques", nombre: "bloques", tipo: textos}},
			salida:           []boe.Articulo(nil),
			argumento:        []string{"articulos", normaDeBoe, "a21", "a22"},
			consulta:         boe.ConsultaArticulos{Norma: normaDeBoe, Bloques: []string{"a21", "a22"}},
			url:              direccionDeLaNorma,
			fueraDeGramatica: []string{"articulos", normaDeBoe, "a21", "../a22"},
			inexistente:      []string{"articulos", normaDeBoe, "a9999"},
			conAvisos:        []string{"articulos", normaDeBoe, "a21"},
		},
		{
			nombre:           "metadatos",
			campos:           []campoDelContrato{norma},
			salida:           boe.Metadatos{},
			argumento:        []string{"metadatos", normaDeBoe},
			consulta:         boe.ConsultaMetadatos{Norma: normaDeBoe},
			url:              direccionDeLosMetadatos,
			fueraDeGramatica: []string{"metadatos", normaFueraDeGramatica},
			inexistente:      []string{"metadatos", normaInexistente},
			conAvisos:        []string{"metadatos", normaDeBoe},
		},
		{
			nombre:           "analisis",
			campos:           []campoDelContrato{norma},
			salida:           boe.Analisis{},
			argumento:        []string{"analisis", normaDeBoe},
			consulta:         boe.ConsultaAnalisis{Norma: normaDeBoe},
			url:              direccionDelAnalisis,
			fueraDeGramatica: []string{"analisis", normaFueraDeGramatica},
			inexistente:      []string{"analisis", normaInexistente},
		},
	}
}

// TestAppletBoe fija el applet boe tal como lo declara el contrato
// puerto-y-applet §4 y lo que hace al ejecutarse (FR-001, US1 escenario 4):
//
//   - se llama boe y declara exactamente los seis verbos, en su orden y ninguno
//     por omisión, de modo que nombrar el verbo es obligatorio;
//   - cada verbo declara sus argumentos posicionales, todos obligatorios, y el
//     valor cero del tipo de su data;
//   - cada verbo construye su consulta con lo que recibe y devuelve tal cual lo
//     que responde la fuente: el sobre que emite el kernel es el que se monta con
//     el resultado de Fetch sobre las mismas grabaciones;
//   - la fuente y su cliente se componen en cada invocación, que solo construye
//     el cliente si pide algo, y lo que una deja en la caché lo sirve la
//     siguiente;
//   - --dry-run abre la caché en solo lectura: no crea nada;
//   - y sin cliente declarado, la invocación es el defecto de composición, sin
//     ningún pánico.
func TestAppletBoe(t *testing.T) {
	t.Parallel()

	t.Run("declara los seis verbos del contrato y ninguno por omisión", func(t *testing.T) {
		t.Parallel()

		applet := AppletBoe(DependenciasDeBoe{})

		assert.Equal(t, "boe", applet.Nombre())
		assert.NotEmpty(t, applet.Descripcion(), "la ayuda del binario lo describe")

		nombres := make([]string, 0, len(applet.Verbos()))

		for _, verbo := range applet.Verbos() {
			nombres = append(nombres, verbo.Nombre)

			assert.NotEmpty(t, verbo.Descripcion, "la ayuda del applet describe %q", verbo.Nombre)
			assert.False(t, verbo.PorOmision, "%q no es el verbo por omisión: boe no tiene ninguno (FR-001)", verbo.Nombre)
		}

		esperados := make([]string, 0, len(verbosDelContrato()))
		for _, verbo := range verbosDelContrato() {
			esperados = append(esperados, verbo.nombre)
		}

		assert.Equal(t, esperados, nombres)

		var registro Registro

		require.NoError(t, registro.Registrar(applet), "el registro lo admite")
	})

	for _, contrato := range verbosDelContrato() {
		t.Run(contrato.nombre, func(t *testing.T) {
			t.Parallel()

			verbo, hay := verboDe(AppletBoe(DependenciasDeBoe{}), contrato.nombre)
			require.True(t, hay, "boe declara el verbo")

			compruebaArgumentosDelContrato(t, verbo, contrato)
			compruebaQueDevuelveLaFuenteTalCual(t, contrato)
		})
	}

	t.Run("compone la fuente en cada invocación y la caché sirve a las siguientes", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBancoDeBoe(t, grabacionesDeBoe)

		primera := banco.invocar(t, argvDeBoe("metadatos", normaDeBoe, "--json")...)
		segunda := banco.invocar(t, argvDeBoe("metadatos", normaDeBoe, "--json")...)
		tercera := banco.invocar(t, argvDeBoe("indice", normaDeBoe, "--json")...)

		for _, res := range []invocacionDeBoe{primera, segunda, tercera} {
			require.Equal(t, 0, res.codigo, res.errores)
		}

		assert.Equal(t, []string{direccionDeLosMetadatos}, primera.pedidas)
		assert.Equal(t, int64(1), primera.construcciones, "la primera construye su cliente para pedir")

		assert.Empty(t, segunda.pedidas, "la segunda sale de la caché que dejó la primera")
		assert.Zero(t, segunda.construcciones, "lo que se sirve de la caché no construye ningún cliente")
		assert.Equal(t, primera.salida, segunda.salida, "con la misma url y la misma fecha de consulta (FR-096)")

		assert.Equal(t, []string{direccionDelIndice}, tercera.pedidas)
		assert.Equal(t, int64(1), tercera.construcciones, "cada invocación construye su propio cliente")
	})

	t.Run("--dry-run abre la caché en solo lectura", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBancoDeBoe(t, grabacionesDeBoe)

		res := banco.invocar(t, argvDeBoe("articulo", normaDeBoe, "a21", "--dry-run")...)

		assert.Equal(t, 0, res.codigo, res.errores)
		assert.Empty(t, res.salida)
		assert.Contains(t, res.errores,
			"--dry-run: se habría pedido GET "+direccionDelBloqueA21+"\n"+
				"--dry-run: se habría pedido GET "+direccionDeLosMetadatos,
			"el ensayo describe las dos peticiones (contrato verbos-y-salidas §8)")
		assert.Empty(t, res.pedidas)
		compruebaCarpetaVacia(t, banco.carpetaDeLaCache)
	})

	t.Run("sin cliente declarado es un defecto de composición", func(t *testing.T) {
		t.Parallel()

		var registro Registro

		require.NoError(t, registro.Registrar(AppletBoe(DependenciasDeBoe{})))

		res := invocar(t, &registro, argvDeBoe("metadatos", normaDeBoe, "--json")...)

		exigirSobreDeFallo(t, res, schema.ClaseInesperado, 1, cli.ProcedenciaKernel())
		assert.Contains(t, res.errores, "ConCliente", "el mensaje nombra la dependencia que falta")
	})
}

// compruebaArgumentosDelContrato exige que el verbo declare exactamente los
// argumentos posicionales del contrato, obligatorios, que su fábrica dé un valor
// nuevo en cada llamada y que su data sea el valor cero del tipo del contrato.
func compruebaArgumentosDelContrato(t *testing.T, verbo Verbo, contrato verboDelContrato) {
	t.Helper()

	require.NotNil(t, verbo.Argumentos)

	primero, segundo := verbo.Argumentos(), verbo.Argumentos()
	assert.NotSame(t, primero, segundo, "la fábrica da un valor nuevo en cada invocación")

	tipo := reflect.TypeOf(primero)
	require.Equal(t, reflect.Pointer, tipo.Kind())
	require.Equal(t, reflect.Struct, tipo.Elem().Kind())

	var declarados []campoDelContrato

	for i := range tipo.Elem().NumField() {
		campo := tipo.Elem().Field(i)
		if !campo.IsExported() {
			continue
		}

		_, porPosicion := campo.Tag.Lookup("arg")
		_, opcional := campo.Tag.Lookup("optional")
		_, porOmision := campo.Tag.Lookup("default")

		assert.True(t, porPosicion, "%s es un argumento que va por su posición", campo.Name)
		assert.False(t, opcional || porOmision, "%s es obligatorio y sin valor por omisión", campo.Name)

		declarados = append(declarados, campoDelContrato{campo: campo.Name, nombre: campo.Tag.Get("name"), tipo: campo.Type})
	}

	assert.Equal(t, contrato.campos, declarados, "los argumentos del contrato, y ninguno más")
	assert.Equal(t, reflect.TypeOf(contrato.salida), reflect.TypeOf(verbo.Salida), "el tipo de data del contrato")
	assert.True(t, reflect.ValueOf(verbo.Salida).IsZero(), "Salida es el valor cero de su tipo")
}

// compruebaQueDevuelveLaFuenteTalCual invoca el verbo sobre las grabaciones y
// exige que su sobre sea el que se monta con lo que responde Fetch a la consulta
// del contrato, con la fuente compuesta a mano sobre las mismas grabaciones, una
// caché vacía y el mismo reloj.
func compruebaQueDevuelveLaFuenteTalCual(t *testing.T, contrato verboDelContrato) {
	t.Helper()

	banco := nuevoBancoDeBoe(t, grabacionesDeBoe)

	res := banco.invocar(t, argvDeBoe(append(slices.Clone(contrato.argumento), "--json")...)...)
	require.Equal(t, 0, res.codigo, res.errores)

	sobre := sobreDelJSON(t, res.salida)

	assert.Equal(t, boe.NombreDeLaFuente, sobre["fuente"], "FR-002")
	assert.Equal(t, contrato.url, sobre["url"], "la dirección que cita el verbo (FR-002)")
	assert.Equal(t, sobreDeLaFuente(t, contrato.consulta), sobre, "el sobre del resultado de Fetch, tal cual")
}

// sobreDeLaFuente es el sobre, decodificado como lo decodifica sobreDelJSON, que
// el kernel monta con el resultado de Fetch sobre las grabaciones.
func sobreDeLaFuente(t *testing.T, consulta core.Consulta) map[string]any {
	t.Helper()

	var hora horaQueAvanza

	carpetaDeLaCache := t.TempDir()

	fuente, err := boe.Nueva(
		boe.ConCliente(func() (boe.Pedidor, error) {
			cliente, err := httpx.Replay(grabacionesDeBoe, httpx.ConFuente(boe.NombreDeLaFuente), httpx.ConHora(hora.ahora))
			if err != nil {
				return nil, err
			}

			return cliente, nil
		}),
		boe.ConCache(func(ctx context.Context, _ bool) (boe.CacheAbierta, error) {
			abierta, err := cache.New(ctx, cache.ConDirectorio(carpetaDeLaCache), cache.ConReloj(ahoraDeLaCache))
			if err != nil {
				return nil, err
			}

			return abierta, nil
		}),
	)
	require.NoError(t, err)

	resultado, err := fuente.Fetch(t.Context(), schema.Contexto{JSON: true, Timeout: timeoutDelContrato}, consulta)
	require.NoError(t, err)

	var montador cli.Montador

	sobre, err := montador.Exito(resultado)
	require.NoError(t, err)

	contenido, err := json.Marshal(sobre)
	require.NoError(t, err)

	var decodificado map[string]any

	require.NoError(t, json.Unmarshal(contenido, &decodificado))

	return decodificado
}

// TestDependenciasDeRed fija, sin tocar la red, las dependencias del binario
// distribuido (contrato puerto-y-applet §4): el cliente se construye por
// invocación con httpx.New y el registrador que entrega el kernel —en ensayo no
// emite nada, y su evento llega a ese registrador; uno nulo lo rechaza httpx—, y
// la caché no lleva opciones, así que es la de la cuenta o la de
// KITLEGAL_CACHE_DIR. El nombre de la fuente y el ritmo no se observan sin red:
// los atan a SOURCES.md las constantes que usa (TestFuenteCoincideConSources).
//
// No es paralelo porque neutraliza la variable de grabación, que httpx.New lee
// del entorno del proceso entero.
func TestDependenciasDeRed(t *testing.T) {
	t.Setenv(httpx.VariableGrabacion, "")

	dependencias := DependenciasDeRed()

	require.NotNil(t, dependencias.Cliente)
	assert.Empty(t, dependencias.Cache, "sin opciones: la caché de la cuenta o la de KITLEGAL_CACHE_DIR")

	var eventos eventosDeHTTPX

	cliente, err := dependencias.Cliente(eventos.registrador())
	require.NoError(t, err)
	require.NotNil(t, cliente)

	respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{DryRun: true},
		httpx.Peticion{Metodo: "GET", URL: direccionDeLosMetadatos, Acepta: "application/json"})
	require.NoError(t, err)
	assert.True(t, respuesta.Ensayo, "en ensayo no se emite nada")
	assert.Contains(t, eventos.lineas.String(), "ensayo: la petición no se emite",
		"los eventos del cliente van al registrador que entrega el kernel")

	_, err = dependencias.Cliente(nil)
	require.Error(t, err, "el registrador llega a httpx tal cual")
	assert.Equal(t, schema.ClaseArgumentos, cli.Clasificar(err))
}

// filaDeSalida es un caso de TestCodigosDeSalidaDeBoe: una fila del contrato
// errores-y-codigos, la invocación que la provoca y el sobre de fallo que tiene
// que salir.
type filaDeSalida struct {
	fila   int
	nombre string
	// reproduccion es la carpeta que sirve el cliente: las grabaciones si va
	// vacía.
	reproduccion string
	// siembra es una invocación previa, sobre el mismo banco, que deja entradas
	// en la caché.
	siembra []string
	// cacheEnUnFichero declara la caché sobre un fichero regular.
	cacheEnUnFichero bool
	argumentos       []string

	clase  schema.Clase
	codigo int
	fuente string
	url    string
	// instante es el número de la petición cuyo instante fecha el sobre, o cero
	// si lo fecha el montaje.
	instante int
	// pedidas son las peticiones de la invocación, sin las de la siembra.
	pedidas []string
	// cacheSinCrear exige que la carpeta de la caché siga vacía.
	cacheSinCrear bool
}

// TestCodigosDeSalidaDeBoe fija por el kernel en proceso, sobre las grabaciones y
// los sintéticos de la fuente y con reloj controlado, las filas 1-6, 11, 14, 18,
// 19 y 21 del contrato errores-y-codigos: el código, la clase, la fuente, la url
// y la fecha de consulta del sobre de fallo, y lo que se pidió (FR-100, FR-101,
// SC-008, US1 escenarios 5 y 6). Ningún caso termina con 6.
//
// La fila 11 —un 5xx tras agotar los reintentos— se provoca con la grabación
// sintética del 503 que la fuente entrega cuando se agotan: la reproducción no
// reintenta, y que el instante sea el del último intento lo fija
// TestInstanteDeEmision en httpx. La fila 19 es la de SC-008: el primer bloque
// sale de la entrada que dejó una consulta anterior, el segundo se pide, no se
// encuentra, y el sobre lleva el instante de esa petición y no el de la entrada
// ni el del montaje.
func TestCodigosDeSalidaDeBoe(t *testing.T) {
	t.Parallel()

	for _, caso := range filasDeSalida() {
		t.Run(fmt.Sprintf("fila-%02d-%s", caso.fila, caso.nombre), func(t *testing.T) {
			t.Parallel()

			carpeta := caso.reproduccion
			if carpeta == "" {
				carpeta = grabacionesDeBoe
			}

			banco := nuevoBancoDeBoe(t, carpeta)
			if caso.cacheEnUnFichero {
				banco.cacheEnUnFichero(t)
			}

			if len(caso.siembra) > 0 {
				sembrada := banco.invocar(t, argvDeBoe(caso.siembra...)...)
				require.Equal(t, 0, sembrada.codigo, sembrada.errores)
			}

			caso.comprueba(t, banco, banco.invocar(t, argvDeBoe(caso.argumentos...)...))
		})
	}
}

// filasDeSalida son los casos de TestCodigosDeSalidaDeBoe, en el orden de la
// tabla.
func filasDeSalida() []filaDeSalida {
	kernel := cli.ProcedenciaKernel()

	argumentos := func(fila int, nombre string, argumentos ...string) filaDeSalida {
		return filaDeSalida{
			fila: fila, nombre: nombre, argumentos: argumentos,
			clase: schema.ClaseArgumentos, codigo: 2, fuente: kernel.Fuente, url: kernel.URL,
			cacheSinCrear: true,
		}
	}

	return []filaDeSalida{
		argumentos(1, "sin-verbo", "--json"),
		argumentos(1, "verbo-desconocido", "articul", normaDeBoe, "a21", "--json"),
		argumentos(1, "falta-un-argumento", "articulo", normaDeBoe, "--json"),
		argumentos(1, "bandera-desconocida", "articulo", normaDeBoe, "a21", "--jsno", "--json"),
		argumentos(2, "norma-fuera-de-su-gramatica", "articulo", "BOE-A-2015", "a21", "--json"),
		argumentos(3, "bloque-fuera-de-su-gramatica", "articulo", normaDeBoe, "../a21", "--json"),
		argumentos(3, "un-bloque-de-articulos-fuera-de-su-gramatica", "articulos", normaDeBoe, "a21", "../a22", "--json"),
		argumentos(4, "busqueda-vacia", "buscar", "", "--json"),
		argumentos(4, "busqueda-solo-con-espacio-en-blanco", "buscar", " \t", "--json"),
		{
			fila: 5, nombre: "offline-sin-entrada", argumentos: []string{"articulo", normaDeBoe, "a21", "--offline", "--json"},
			clase: schema.ClaseFuenteNoDisponible, codigo: 4, fuente: boe.NombreDeLaFuente, url: direccionDelBloqueA21,
			cacheSinCrear: true,
		},
		{
			fila: 6, nombre: "bloque-inexistente", argumentos: []string{"articulo", normaDeBoe, "a9999", "--json"},
			clase: schema.ClaseNoEncontrado, codigo: 3, fuente: boe.NombreDeLaFuente, url: direccionDelBloqueA9999,
			instante: 1, pedidas: []string{direccionDelBloqueA9999},
		},
		{
			fila: 11, nombre: "5xx", reproduccion: sintetico("fuente-caida"),
			argumentos: []string{"indice", normaDeBoe, "--json"},
			clase:      schema.ClaseFuenteNoDisponible, codigo: 4, fuente: boe.NombreDeLaFuente, url: direccionDelIndice,
			instante: 1, pedidas: []string{direccionDelIndice},
		},
		{
			fila: 14, nombre: "429", reproduccion: sintetico("limite"),
			argumentos: []string{"metadatos", normaDeBoe, "--json"},
			clase:      schema.ClaseLimiteOTos, codigo: 5, fuente: boe.NombreDeLaFuente, url: direccionDeLosMetadatos,
			instante: 1, pedidas: []string{direccionDeLosMetadatos},
		},
		{
			fila: 18, nombre: "bloque-obtenido-y-metadatos-caidos", reproduccion: sintetico("metadatos-caidos"),
			argumentos: []string{"articulo", normaDeBoe, "a21", "--json"},
			clase:      schema.ClaseFuenteNoDisponible, codigo: 4, fuente: boe.NombreDeLaFuente, url: direccionDeLosMetadatos,
			instante: 2, pedidas: []string{direccionDelBloqueA21, direccionDeLosMetadatos},
		},
		{
			fila: 19, nombre: "articulos-con-el-segundo-bloque-inexistente",
			siembra:    []string{"articulo", normaDeBoe, "a21", "--json"},
			argumentos: []string{"articulos", normaDeBoe, "a21", "a9999", "--json"},
			clase:      schema.ClaseNoEncontrado, codigo: 3, fuente: boe.NombreDeLaFuente, url: direccionDelBloqueA9999,
			instante: 3, pedidas: []string{direccionDelBloqueA9999},
		},
		{
			fila: 21, nombre: "la-cache-no-se-abre", cacheEnUnFichero: true,
			argumentos: []string{"articulo", normaDeBoe, "a21", "--json"},
			clase:      schema.ClaseArgumentos, codigo: 2, fuente: boe.NombreDeLaFuente, url: direccionDelBloqueA21,
		},
	}
}

// comprueba exige el sobre de fallo de la fila sobre lo que dejó la invocación.
func (caso filaDeSalida) comprueba(t *testing.T, banco *bancoDeBoe, res invocacionDeBoe) {
	t.Helper()

	assert.Equal(t, caso.codigo, res.codigo, res.errores)

	sobre := sobreDelJSON(t, res.salida)

	assert.Equal(t, false, sobre["ok"])
	assert.Equal(t, caso.fuente, sobre["fuente"])
	assert.Equal(t, caso.url, sobre["url"], "la dirección de lo que falló (FR-101)")

	datos := datosDelSobre(t, sobre)
	assert.Equal(t, string(caso.clase), datos["clase"])

	mensaje, esTexto := datos["mensaje"].(string)
	require.True(t, esTexto, "el mensaje es un texto")
	assert.NotEmpty(t, mensaje)
	assert.Contains(t, res.errores, mensaje, "el mismo mensaje sale para la persona")

	fecha := fechaDelSobre(t, sobre)

	if caso.instante == 0 {
		assert.False(t, fecha.Before(res.antes) || fecha.After(res.despues),
			"sin petición, el sobre lo fecha el montaje: %s no está entre %s y %s", fecha, res.antes, res.despues)
	} else {
		assert.True(t, fecha.Equal(instanteDeLaPeticion(caso.instante)),
			"el sobre lleva el instante de la petición %d: %s, no %s",
			caso.instante, instanteDeLaPeticion(caso.instante), fecha)
	}

	assert.Equal(t, caso.pedidas, res.pedidas, "las peticiones de la invocación")

	if len(caso.pedidas) == 0 {
		assert.Zero(t, res.construcciones, "sin nada que pedir no se construye ningún cliente")
	}

	if caso.cacheSinCrear {
		compruebaCarpetaVacia(t, banco.carpetaDeLaCache)
	}
}

// fechaDeConsultaDelSobre casa la fecha de consulta de un sobre en JSON.
var fechaDeConsultaDelSobre = regexp.MustCompile(`"fecha_consulta":\s*"[^"]*"`)

// TestSinGrafoNiAsuntoNoCambianLaSalida fija FR-128 y la mitad de SC-013 sobre
// los seis verbos: con --no-graph, con --asunto y con las dos, la salida estándar
// es la misma byte a byte que sin ellas, salvo fecha_consulta, y las peticiones
// son las mismas y en el mismo orden. Cada invocación va sobre su propio banco,
// con la caché vacía y el mismo reloj.
func TestSinGrafoNiAsuntoNoCambianLaSalida(t *testing.T) {
	t.Parallel()

	conBanderas := [][]string{{"--no-graph"}, {"--asunto", "demo"}, {"--no-graph", "--asunto", "demo"}}

	for _, contrato := range verbosDelContrato() {
		t.Run(contrato.nombre, func(t *testing.T) {
			t.Parallel()

			base := slices.Concat(contrato.argumento, []string{"--json"})
			sinBanderas := nuevoBancoDeBoe(t, grabacionesDeBoe).invocar(t, argvDeBoe(base...)...)
			require.Equal(t, 0, sinBanderas.codigo, sinBanderas.errores)

			for _, banderas := range conBanderas {
				res := nuevoBancoDeBoe(t, grabacionesDeBoe).invocar(t, argvDeBoe(slices.Concat(base, banderas)...)...)

				require.Equal(t, 0, res.codigo, res.errores)
				assert.Equal(t,
					fechaDeConsultaDelSobre.ReplaceAllString(sinBanderas.salida, `"fecha_consulta":""`),
					fechaDeConsultaDelSobre.ReplaceAllString(res.salida, `"fecha_consulta":""`),
					"%v no cambia la salida", banderas)
				assert.Equal(t, sinBanderas.pedidas, res.pedidas, "%v no cambia las peticiones", banderas)
			}
		})
	}
}

// TestSalidaDeBoeContraSchemas es el punto 4 de la Definition of Done sobre el
// applet boe (FR-111, SC-006, US6 escenarios 1 y 2): el sobre real que emite el
// kernel con --json para cada uno de los seis verbos —el de éxito y los de fallo
// con código 2, 3 y 4— valida contra la parte de su verbo leída de
// schemas/norma.json o de schemas/bloque.json, y no contra lo que emite
// --describe mientras se ejecuta el test. La validación restringe: el mismo
// sobre con una clave de más o de menos en su data no valida.
//
// En los tres verbos que devuelven avisos (FR-012), la parte publicada enumera
// Aviso.codigo con exactamente sus tres valores, y un sobre real con los tres
// avisos deja de validar si cualquiera de ellos lleva otro código.
func TestSalidaDeBoeContraSchemas(t *testing.T) {
	t.Parallel()

	for _, contrato := range verbosDelContrato() {
		t.Run(contrato.nombre, func(t *testing.T) {
			t.Parallel()

			publicado, id := ficheroPublicadoDelVerbo(t, contrato.nombre)
			esquema := salidaPublicada(t, publicado, id, contrato.nombre)

			if contrato.conAvisos != nil {
				t.Run("Aviso.codigo", func(t *testing.T) {
					t.Parallel()

					assert.ElementsMatch(t, []any{"consolidacion-no-finalizada", "derogada", "vigencia-agotada"},
						valorDelEsquema(t, publicado, "$defs", contrato.nombre, "$defs", "boe.Aviso", "properties", "codigo", "enum"),
						"exactamente los tres valores de FR-012, sin ningún otro")
				})
			}

			for _, caso := range salidasDeBoe(contrato) {
				t.Run(caso.nombre, func(t *testing.T) {
					t.Parallel()

					banco := nuevoBancoDeBoe(t, caso.reproduccion)
					res := banco.invocar(t, argvDeBoe(slices.Concat(caso.argumentos, []string{"--json"})...)...)

					require.Equal(t, caso.codigo, res.codigo, res.errores)
					assert.Equal(t, caso.codigo == 0, sobreDelJSON(t, res.salida)["ok"],
						"ok decide la rama del esquema contra la que se valida data")

					exigirSalidaPublicada(t, esquema, res.salida)

					if caso.conAvisos {
						exigirAvisosEnumerados(t, esquema, res.salida)
					}
				})
			}
		})
	}
}

// salidaDeBoe es una invocación de TestSalidaDeBoeContraSchemas: la carpeta que
// sirve el cliente, los argumentos sin --json, el código con el que termina y si
// su data lleva los tres avisos.
type salidaDeBoe struct {
	nombre       string
	reproduccion string
	argumentos   []string
	codigo       int
	conAvisos    bool
}

// salidasDeBoe son las invocaciones del verbo cuyo sobre se valida: la de éxito
// sobre las grabaciones; las de fallo con código 2, con 3 si el verbo lo tiene y
// con 4, esta con --offline y la caché vacía (contrato errores-y-codigos, fila
// 5); y, si el verbo devuelve avisos, la de éxito con los tres.
func salidasDeBoe(contrato verboDelContrato) []salidaDeBoe {
	salidas := []salidaDeBoe{
		{nombre: "exito", reproduccion: grabacionesDeBoe, argumentos: contrato.argumento},
		{
			nombre: "fallo-2-fuera-de-la-gramatica", reproduccion: grabacionesDeBoe,
			argumentos: contrato.fueraDeGramatica, codigo: 2,
		},
		{
			nombre: "fallo-4-offline-sin-entrada", reproduccion: grabacionesDeBoe,
			argumentos: slices.Concat(contrato.argumento, []string{"--offline"}), codigo: 4,
		},
	}

	if contrato.inexistente != nil {
		salidas = append(salidas, salidaDeBoe{
			nombre: "fallo-3-inexistente", reproduccion: grabacionesDeBoe,
			argumentos: contrato.inexistente, codigo: 3,
		})
	}

	if contrato.conAvisos != nil {
		salidas = append(salidas, salidaDeBoe{
			nombre: "exito-con-tres-avisos", reproduccion: sintetico("avisos"),
			argumentos: contrato.conAvisos, conAvisos: true,
		})
	}

	return salidas
}

// ficheroPublicadoDelVerbo lee de schemas/ el fichero que publica la parte del
// verbo, norma.json o bloque.json (contrato esquemas-fixtures-y-controles §1), en
// la representación que consume el validador, y devuelve con él su $id, que
// tiene que ser el del contrato.
func ficheroPublicadoDelVerbo(t *testing.T, verbo string) (any, string) {
	t.Helper()

	i := slices.IndexFunc(ficherosDeEsquemas, func(fichero ficheroDeEsquemas) bool {
		return slices.Contains(fichero.verbos, verbo)
	})
	require.NotEqual(t, -1, i, "el verbo %q tiene su parte en un fichero publicado", verbo)

	fichero := ficherosDeEsquemas[i]

	contenido, err := leerEsquema(filepath.Join(carpetaDeLosEsquemas, fichero.nombre))
	require.NoError(t, err, "la salida se valida contra el fichero publicado (FR-111)")

	documento, err := jsonschema.UnmarshalJSON(bytes.NewReader(contenido))
	require.NoError(t, err, "schemas/%s es un único documento JSON", fichero.nombre)

	id, esTexto := valorDelEsquema(t, documento, "$id").(string)
	require.True(t, esTexto, "el $id de schemas/%s es un texto", fichero.nombre)
	require.Equal(t, raizDeLosEsquemas+fichero.nombre, id, "el $id del contrato")

	return documento, id
}

// salidaPublicada compila la salida del verbo desde el fichero publicado: el
// documento entra en el compilador con su $id y la parte se compila por ese $id,
// con el puntero del contrato esquemas-fixtures-y-controles §1; sus referencias
// internas resuelven contra el $id de la parte, que es un recurso embebido. Las
// aserciones de formato van activadas: sin ellas, en el borrador 2020-12, format
// es una anotación y una url o una fecha_consulta mal formadas validarían.
func salidaPublicada(t *testing.T, documento any, id, verbo string) *jsonschema.Schema {
	t.Helper()

	compilador := jsonschema.NewCompiler()
	compilador.AssertFormat()
	require.NoError(t, compilador.AddResource(id, documento))

	esquema, err := compilador.Compile(id + "#/$defs/" + verbo + "/properties/salida")
	require.NoError(t, err, "la parte de %q en %s compila", verbo, id)

	return esquema
}

// exigirSalidaPublicada exige que el sobre real valide contra la salida
// publicada y que la validación restrinja data: el mismo sobre con una clave de
// más, o sin cualquiera de las suyas, en el objeto de data no valida.
func exigirSalidaPublicada(t *testing.T, esquema *jsonschema.Schema, salida string) {
	t.Helper()

	require.NoError(t, esquema.Validate(sobreValidable(t, salida)), "el sobre real valida contra su parte de schemas/")

	conClaveDeMas := sobreValidable(t, salida)
	objetoDeData(t, conClaveDeMas)["ajena"] = "no declarada"
	require.Error(t, esquema.Validate(conClaveDeMas), "data con una clave de más no valida")

	claves := slices.Sorted(maps.Keys(objetoDeData(t, sobreValidable(t, salida))))
	require.NotEmpty(t, claves, "un data sin claves no permitiría comprobar que el esquema las exige")

	for _, clave := range claves {
		sinLaClave := sobreValidable(t, salida)
		delete(objetoDeData(t, sinLaClave), clave)
		assert.Errorf(t, esquema.Validate(sinLaClave), "data sin la clave %q no valida", clave)
	}
}

// exigirAvisosEnumerados exige que el data del sobre real lleve los tres avisos
// y que el mismo sobre, con otro código en cualquiera de ellos, no valide: el
// enumerado de Aviso.codigo se aplica a lo que se emite (FR-012).
func exigirAvisosEnumerados(t *testing.T, esquema *jsonschema.Schema, salida string) {
	t.Helper()

	require.Len(t, avisosDelSobre(t, sobreValidable(t, salida)), 3, "el sintético avisos cumple las tres condiciones")

	for posicion := range 3 {
		conOtroCodigo := sobreValidable(t, salida)

		aviso, esObjeto := avisosDelSobre(t, conOtroCodigo)[posicion].(map[string]any)
		require.True(t, esObjeto, "cada aviso es un objeto")

		aviso["codigo"] = "otro-codigo"
		assert.Errorf(t, esquema.Validate(conOtroCodigo), "el aviso %d con un código fuera de los tres no valida", posicion)
	}
}

// avisosDelSobre es la lista de avisos del objeto de data.
func avisosDelSobre(t *testing.T, sobre map[string]any) []any {
	t.Helper()

	avisos, esLista := objetoDeData(t, sobre)["avisos"].([]any)
	require.True(t, esLista, "los avisos son una lista")

	return avisos
}

// objetoDeData es el objeto que describe data en el sobre: data mismo o, si es
// una lista —en buscar y articulos—, su primer elemento, que tiene que existir.
// Pertenece al sobre, así que cambiarlo cambia el sobre.
func objetoDeData(t *testing.T, sobre map[string]any) map[string]any {
	t.Helper()

	data := sobre["data"]

	if lista, esLista := data.([]any); esLista {
		require.NotEmpty(t, lista, "una lista vacía no permitiría comprobar que el esquema restringe sus elementos")

		data = lista[0]
	}

	objeto, esObjeto := data.(map[string]any)
	require.True(t, esObjeto, "data, o su primer elemento, es un objeto")

	return objeto
}

// sobreValidable lee la salida estándar en la representación que consume el
// validador, con las cifras como literales. Cada llamada da una copia nueva, que
// se puede alterar sin tocar las demás.
func sobreValidable(t *testing.T, salida string) map[string]any {
	t.Helper()

	documento, err := jsonschema.UnmarshalJSON(strings.NewReader(salida))
	require.NoError(t, err, "con --json la salida estándar es un único documento JSON")

	sobre, esObjeto := documento.(map[string]any)
	require.True(t, esObjeto, "el sobre es un objeto JSON")

	return sobre
}

// valorDelEsquema es el valor al que llevan las claves, una tras otra, dentro
// del documento; cada paso tiene que existir.
func valorDelEsquema(t *testing.T, documento any, claves ...string) any {
	t.Helper()

	valor := documento

	for i, clave := range claves {
		objeto, esObjeto := valor.(map[string]any)
		require.True(t, esObjeto, "/%s es un objeto", strings.Join(claves[:i], "/"))

		siguiente, existe := objeto[clave]
		require.True(t, existe, "el esquema tiene /%s", strings.Join(claves[:i+1], "/"))

		valor = siguiente
	}

	return valor
}
