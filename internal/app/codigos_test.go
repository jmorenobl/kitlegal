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
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Los datos de construcción con los que se invoca al kernel en los tests. Son
// fijos porque ninguna de estas tablas mira el verbo reservado «version»: lo que
// importa es que la raíz de composición los reciba y no los interprete.
const (
	versionDePrueba = "1.2.3"
	commitDePrueba  = "0123456789abcdef"
	fechaDePrueba   = "2026-01-01T00:00:00Z"
)

// Las seis claves exactas del sobre, en el nivel superior y sin ninguna más:
// es lo mismo en éxito y en fallo (FR-010, FR-045).
var clavesDelSobre = []string{"ok", "fuente", "url", "fecha_consulta", "hash", "data"}

// procedenciaDePrueba es la de la fuente que el applet declara estar
// consultando. Aparece en el sobre de fallo cuando el fallo ocurre **dentro**
// del applet; los fallos anteriores a su ejecución llevan la del kernel
// (FR-016, FR-045).
var procedenciaDePrueba = schema.Procedencia{
	Fuente: "kitlegal.prueba",
	URL:    "kitlegal:applet/prueba",
}

// desenlaceDePrueba es lo que el applet de estas tablas devuelve: cada caso lo
// fija y el kernel lo traduce. Recibe el contexto de la operación porque uno de
// los casos —el del plazo agotado— necesita esperar a que venza.
//
// Un applet no conoce ningún código de salida, ningún sobre y ninguna bandera:
// devuelve un resultado o un error y nada más (FR-015, SC-010).
type desenlaceDePrueba func(ctx context.Context) (schema.Resultado, error)

// appletDeCodigos declara exactamente lo que declara cualquier applet —nombre,
// descripción y un verbo— y nada más. Las ocho banderas globales, el sobre, la
// huella, las dos formas de presentación y la tabla de códigos de salida las
// hereda del kernel sin escribir una línea (SC-010).
type appletDeCodigos struct {
	desenlace desenlaceDePrueba
}

func (a appletDeCodigos) Nombre() string { return "prueba" }

func (a appletDeCodigos) Descripcion() string { return "applet de la tabla de códigos" }

func (a appletDeCodigos) Verbos() []Verbo {
	return []Verbo{{
		Nombre:      "probar",
		Descripcion: "devuelve el desenlace que el caso fija",
		Argumentos:  func() Argumentos { return &argumentosDeCodigos{desenlace: a.desenlace} },
		Salida:      map[string]any{},
		PorOmision:  true,
	}}
}

// argumentosDeCodigos son los argumentos del verbo: uno solo, que se escribe sin
// bandera y puede faltar, y que hace válida «prueba hola» sin nombrar verbo. El
// desenlace viaja en un campo no exportado, así que la gramática no lo ve y
// ninguna invocación puede fijarlo.
type argumentosDeCodigos struct {
	Mensaje string `arg:"" optional:"" help:"Mensaje que se devuelve."`

	desenlace desenlaceDePrueba
}

func (a *argumentosDeCodigos) Ejecutar(
	ctx context.Context, _ schema.Contexto, _ *slog.Logger,
) (schema.Resultado, error) {
	return a.desenlace(ctx)
}

// Las dos comprobaciones en tiempo de compilación del contrato del applet.
var (
	_ Applet     = appletDeCodigos{}
	_ Argumentos = (*argumentosDeCodigos)(nil)
)

// registroDeCodigos construye el registro de una invocación de prueba: un único
// applet cuyo desenlace fija el caso.
func registroDeCodigos(t *testing.T, desenlace desenlaceDePrueba) *Registro {
	t.Helper()

	var registro Registro

	require.NoError(t, registro.Registrar(appletDeCodigos{desenlace: desenlace}))

	return &registro
}

// resultadoCorrecto es el desenlace del caso que no falla: una procedencia que
// sostiene la cita y un contenido derivado de nada más.
func resultadoCorrecto(_ context.Context) (schema.Resultado, error) {
	return schema.Resultado{
		Procedencia: procedenciaDePrueba,
		Datos:       map[string]any{"mensaje": "hola"},
	}, nil
}

// falloDe construye el desenlace de un applet que falla con la clase que se le
// diga, envolviendo el sentinela con contexto: envolver no cambia la clase ni,
// por tanto, el código de salida (FR-032).
//
// El resultado que acompaña al fallo lleva la procedencia de la fuente que se
// estaba consultando, que es la que el sobre de fallo cita cuando se conoce
// (FR-045).
func falloDe(err error) desenlaceDePrueba {
	return func(_ context.Context) (schema.Resultado, error) {
		return schema.Resultado{Procedencia: procedenciaDePrueba},
			fmt.Errorf("el applet de prueba no pudo atender la invocación: %w", err)
	}
}

// invocacionDePrueba es todo lo observable de una invocación completa del
// kernel: su código de salida y lo que quedó en cada uno de los dos
// descriptores, capturados por separado, que es justo lo que SC-011 exige poder
// mirar.
type invocacionDePrueba struct {
	codigo  int
	salida  string
	errores string
}

// invocar ejecuta la raíz de composición entera contra dos buffers.
//
// Que esta función retorne —y que el test siga vivo después— es la comprobación
// de que ningún camino llama a os.Exit: el kernel devuelve un código y quien
// termina el proceso es el punto de entrada (FR-035).
func invocar(t *testing.T, registro *Registro, argv ...string) invocacionDePrueba {
	t.Helper()

	var salida, errores bytes.Buffer

	codigo := Main(argv, registro, &salida, &errores,
		versionDePrueba, commitDePrueba, fechaDePrueba)

	return invocacionDePrueba{
		codigo:  codigo,
		salida:  salida.String(),
		errores: errores.String(),
	}
}

// sobreDelJSON analiza la salida estándar de una invocación con --json y
// devuelve el sobre. Comprueba de paso las dos cosas que el contrato promete de
// ese descriptor: que lleva **un único** documento y que tiene exactamente las
// seis claves del sobre, ni una más ni una menos (FR-010, FR-042, SC-002).
func sobreDelJSON(t *testing.T, salida string) map[string]any {
	t.Helper()

	decodificador := json.NewDecoder(strings.NewReader(salida))

	var sobre map[string]any

	require.NoError(t, decodificador.Decode(&sobre))
	require.ErrorIs(t, decodificador.Decode(new(json.RawMessage)), io.EOF,
		"con --json la salida estándar lleva un único documento JSON y nada más")

	assert.ElementsMatch(t, clavesDelSobre, slices.Collect(maps.Keys(sobre)))

	return sobre
}

// datosDelSobre devuelve el contenido de `data` de un sobre de fallo, que son
// exactamente dos claves: la clase y el mensaje para la persona (FR-045).
func datosDelSobre(t *testing.T, sobre map[string]any) map[string]any {
	t.Helper()

	datos, esObjeto := sobre["data"].(map[string]any)
	require.True(t, esObjeto, "el data de un sobre de fallo es un objeto")

	return datos
}

// casoDeCodigo es una fila de la tabla de códigos de salida: qué devuelve el
// applet y con qué código termina el proceso. El nombre del subcaso es la clase,
// de modo que `-run 'TestCodigoSalida/inesperado'` seleccione uno solo
// (quickstart.md, escenario 5).
type casoDeCodigo struct {
	clase     schema.Clase
	desenlace desenlaceDePrueba
	codigo    int
}

// casosDeCodigo son las siete filas de la tabla de contracts/banderas-y-exit-codes.md
// §4: el éxito, el fallo inesperado y las cinco clases declaradas. Ninguna clase
// queda sin caso (SC-006).
func casosDeCodigo() []casoDeCodigo {
	return []casoDeCodigo{
		{clase: "correcto", desenlace: resultadoCorrecto, codigo: 0},
		{
			// Lo inesperado es lo que nadie declaró: un error que no casa con
			// ninguno de los cinco sentinelas. Es el mismo código con el que sale
			// un fallo de escritura propagado por el presentador (FR-031).
			clase:     schema.ClaseInesperado,
			desenlace: falloDe(errors.New("algo que nadie previó")),
			codigo:    1,
		},
		{clase: schema.ClaseArgumentos, desenlace: falloDe(cli.ErrArgumentos), codigo: 2},
		{clase: schema.ClaseNoEncontrado, desenlace: falloDe(cli.ErrNoEncontrado), codigo: 3},
		{
			clase:     schema.ClaseFuenteNoDisponible,
			desenlace: falloDe(cli.ErrFuenteNoDisponible),
			codigo:    4,
		},
		{clase: schema.ClaseLimiteOTos, desenlace: falloDe(cli.ErrLimiteOTos), codigo: 5},
		{
			clase:     schema.ClaseIdentidadHumana,
			desenlace: falloDe(cli.ErrIdentidadHumana),
			codigo:    6,
		},
	}
}

// TestCodigoSalida fuerza los siete desenlaces desde el applet y comprueba el
// código con el que termina el proceso y el reparto de descriptores **sin**
// --json: el mensaje del fallo va a la salida de error y la estándar queda
// vacía, porque un resultado que no existe no se cita
// (FR-030, SC-006, SC-011, contracts/banderas-y-exit-codes.md §5).
//
// No es paralelo porque fija KITLEGAL_LOG —en el test y en cada subcaso, que es
// lo que deja hermético cada uno por separado—: el caso correcto afirma que la
// salida de error queda vacía, y el kernel lee el nivel del registro del
// entorno del proceso, así que sin fijarlo el veredicto dependería del entorno
// de quien ejecuta los tests.
func TestCodigoSalida(t *testing.T) {
	t.Setenv(cli.VariableNivel, "")

	for _, caso := range casosDeCodigo() {
		t.Run(string(caso.clase), func(t *testing.T) {
			t.Setenv(cli.VariableNivel, "")

			res := invocar(t, registroDeCodigos(t, caso.desenlace),
				"kitlegal", "prueba", "hola")

			assert.Equal(t, caso.codigo, res.codigo)

			if caso.codigo == 0 {
				assert.Contains(t, res.salida, procedenciaDePrueba.Fuente,
					"la tabla mínima lleva los cuatro datos de procedencia")
				assert.Empty(t, res.errores,
					"una invocación correcta no tiene nada que decir en la salida de error")

				return
			}

			assert.Empty(t, res.salida,
				"sin --json un fallo deja la salida estándar vacía")
			assert.Contains(t, res.errores, "el applet de prueba no pudo atender la invocación",
				"el mensaje del fallo va a la salida de error")
		})
	}
}

// TestSobreDeFallo comprueba la otra mitad del contrato: con --json, la salida
// estándar lleva el sobre de seis claves con `ok` falso y la clase y el mensaje
// dentro de `data`, para cada uno de los seis códigos de fallo y también para
// los dos fallos **anteriores** a la ejecución del applet (FR-045, SC-014).
func TestSobreDeFallo(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeCodigo() {
		if caso.codigo == 0 {
			continue
		}

		t.Run(string(caso.clase), func(t *testing.T) {
			t.Parallel()

			res := invocar(t, registroDeCodigos(t, caso.desenlace),
				"kitlegal", "prueba", "hola", "--json")

			exigirSobreDeFallo(t, res, caso.clase, caso.codigo, procedenciaDePrueba)
			assert.Contains(t, res.errores, "el applet de prueba no pudo atender la invocación",
				"el sobre duplica el mensaje en forma estructurada; no lo sustituye")
		})
	}

	// Los dos fallos anteriores a la ejecución del applet. La forma de
	// presentación la decide el pre-escaneo, porque cuando ocurren todavía no
	// hay gramática analizada, y la procedencia es la del propio kernel, porque
	// no se ha llegado a consultar nada (FR-045, SC-014, research.md D25).
	antesDelApplet := []struct {
		nombre string
		argv   []string
	}{
		{nombre: "bandera desconocida", argv: []string{"kitlegal", "prueba", "hola", "--jsno", "--json"}},
		{nombre: "applet no registrado", argv: []string{"kitlegal", "noexiste", "--json"}},
		{nombre: "verbo reservado con argumentos", argv: []string{"kitlegal", "version", "extra", "--json"}},
	}

	for _, caso := range antesDelApplet {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			res := invocar(t, registroDeCodigos(t, resultadoCorrecto), caso.argv...)

			exigirSobreDeFallo(t, res, schema.ClaseArgumentos, 2, cli.ProcedenciaKernel())
			assert.NotEmpty(t, res.errores, "el mensaje para la persona va a la salida de error")
		})
	}
}

// exigirSobreDeFallo comprueba las cinco afirmaciones que SC-014 hace sobre
// cualquier fallo emitido con --json, sea del applet o anterior a él: el código,
// el único documento con las seis claves, `ok` falso, la procedencia que
// corresponde y la clase y el mensaje dentro de `data`.
func exigirSobreDeFallo(
	t *testing.T,
	res invocacionDePrueba,
	clase schema.Clase,
	codigo int,
	procedencia schema.Procedencia,
) {
	t.Helper()

	assert.Equal(t, codigo, res.codigo)

	sobre := sobreDelJSON(t, res.salida)

	assert.Equal(t, false, sobre["ok"], "ok es falso si y solo si el código no es 0")
	assert.Equal(t, procedencia.Fuente, sobre["fuente"])
	assert.Equal(t, procedencia.URL, sobre["url"])
	assert.NotEmpty(t, sobre["fecha_consulta"])
	assert.NotEmpty(t, sobre["hash"])

	datos := datosDelSobre(t, sobre)

	assert.Equal(t, string(clase), datos["clase"])
	assert.NotEmpty(t, datos["mensaje"], "el mensaje para la persona viaja dentro de data")
}
