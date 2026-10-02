package app

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/core/territorio"
	"github.com/jmorenobl/kitlegal/internal/graph"
)

// appletDePrueba es lo que declara un applet y nada más: nombre, descripción y
// verbos. No accede a la red, no toca disco y no conoce ninguna bandera, que es
// precisamente lo que SC-010 exige de un applet nuevo.
type appletDePrueba struct {
	nombre string
	verbos []Verbo
}

func (a appletDePrueba) Nombre() string { return a.nombre }

func (a appletDePrueba) Descripcion() string { return "applet de prueba" }

func (a appletDePrueba) Verbos() []Verbo { return a.verbos }

// argumentosDePrueba es el struct de argumentos de un verbo: las etiquetas de la
// gramática y la operación. Devuelve un Resultado —procedencia y datos— y nunca
// un sobre, ni un código de salida, ni texto escrito en un descriptor.
type argumentosDePrueba struct {
	Mensaje string `arg:"" optional:"" help:"Mensaje que se devuelve."`
}

func (a *argumentosDePrueba) Ejecutar(
	_ context.Context, _ schema.Contexto, _ *slog.Logger,
) (schema.Resultado, error) {
	return schema.Resultado{
		Procedencia: schema.Procedencia{
			Fuente: "kitlegal.prueba",
			URL:    "kitlegal:applet/prueba",
		},
		Datos: map[string]any{"mensaje": a.Mensaje},
	}, nil
}

// Las dos comprobaciones en tiempo de compilación del contrato: un applet
// declara los tres métodos y unos argumentos saben ejecutarse.
var (
	_ Applet     = appletDePrueba{}
	_ Argumentos = (*argumentosDePrueba)(nil)
)

// verboDePrueba construye un verbo completo: la fábrica de argumentos —que
// devuelve un valor nuevo en cada invocación— y el valor cero del tipo de data,
// que solo se refleja.
func verboDePrueba(nombre string, porOmision bool) Verbo {
	return Verbo{
		Nombre:      nombre,
		Descripcion: "verbo de prueba",
		Argumentos:  func() Argumentos { return &argumentosDePrueba{} },
		Salida:      map[string]any{},
		PorOmision:  porOmision,
	}
}

// appletConVerbos es el applet válido mínimo: un nombre y al menos un verbo.
func appletConVerbos(nombre string, verbos ...Verbo) appletDePrueba {
	return appletDePrueba{nombre: nombre, verbos: verbos}
}

// TestRegistro ejerce lo que el registro ofrece a quien lo consume: registrar,
// buscar y enumerar. Es la única fuente de la que salen el despacho, la ayuda y
// la autodescripción, así que los nombres tienen que salir ordenados y no en el
// orden en que alguien los registró (FR-001, data-model.md §7).
func TestRegistro(t *testing.T) {
	t.Parallel()

	t.Run("registra, busca y enumera ordenado", func(t *testing.T) {
		t.Parallel()

		var registro Registro

		// Deliberadamente en desorden: lo que se comprueba es que Nombres
		// ordena, no que conserve el orden de registro.
		require.NoError(t, registro.Registrar(appletConVerbos("plazos",
			verboDePrueba("vencimiento", true))))
		require.NoError(t, registro.Registrar(appletConVerbos("boe",
			verboDePrueba("articulo", false), verboDePrueba("buscar", false))))
		require.NoError(t, registro.Registrar(appletConVerbos("cita",
			verboDePrueba("resolver", true))))

		assert.Equal(t, []string{"boe", "cita", "plazos"}, registro.Nombres(),
			"los nombres salen ordenados, sea cual sea el orden de registro")

		encontrado, existe := registro.Buscar("boe")
		require.True(t, existe)
		assert.Equal(t, "boe", encontrado.Nombre())
		assert.Len(t, encontrado.Verbos(), 2)

		_, existe = registro.Buscar("bde")
		assert.False(t, existe, "un nombre que nadie registró no está")
	})

	t.Run("un applet sin verbo por omisión es válido", func(t *testing.T) {
		t.Parallel()

		var registro Registro

		// El segundo applet de ejemplo de SC-010: dos verbos y ninguno por
		// omisión, de modo que nombrar el verbo es obligatorio. Ninguno marcado
		// es válido; dos, no (contracts/registro-y-describe.md §1).
		require.NoError(t, registro.Registrar(appletConVerbos("segundo",
			verboDePrueba("uno", false), verboDePrueba("dos", false))))

		assert.Equal(t, []string{"segundo"}, registro.Nombres())
	})

	t.Run("la fábrica de argumentos devuelve un valor nuevo", func(t *testing.T) {
		t.Parallel()

		verbo := verboDePrueba("repetir", true)

		primero, segundo := verbo.Argumentos(), verbo.Argumentos()

		require.NotNil(t, primero)
		require.NotNil(t, segundo)
		assert.NotSame(t, primero, segundo,
			"Argumentos es una fábrica: nunca una instancia compartida entre invocaciones")
	})
}

// TestRegistroRechaza es la tabla de las ocho causas de rechazo que recogen las
// cinco reglas de validación de contracts/registro-y-describe.md §1. Todas se
// comprueban **al construirse** el registro: un registro mal construido es un
// defecto de compilación y nunca un error que reciba quien invoca el binario
// (FR-008, research.md D17).
func TestRegistroRechaza(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		yaEstaban []Applet
		applet    Applet
		esperado  error
		mensaje   string
	}{
		{
			nombre:   "nombre vacío",
			applet:   appletConVerbos("", verboDePrueba("repetir", true)),
			esperado: ErrNombreInvalido,
			mensaje:  "sin nombre",
		},
		{
			nombre:   "nombre con espacios",
			applet:   appletConVerbos("boe fiscal", verboDePrueba("articulo", true)),
			esperado: ErrNombreInvalido,
			mensaje:  "espacios",
		},
		{
			nombre:   "nombre con prefijo de bandera",
			applet:   appletConVerbos("-json", verboDePrueba("repetir", true)),
			esperado: ErrNombreInvalido,
			mensaje:  "bandera",
		},
		{
			nombre:    "nombre duplicado",
			yaEstaban: []Applet{appletConVerbos("boe", verboDePrueba("articulo", true))},
			applet:    appletConVerbos("boe", verboDePrueba("buscar", true)),
			esperado:  ErrNombreDuplicado,
			mensaje:   "boe",
		},
		{
			nombre:   "nombre igual a un verbo reservado del binario",
			applet:   appletConVerbos("version", verboDePrueba("mostrar", true)),
			esperado: ErrNombreReservado,
			mensaje:  "version",
		},
		{
			nombre:   "applet sin verbos",
			applet:   appletConVerbos("mudo"),
			esperado: ErrVerbosInvalidos,
			mensaje:  "ningún verbo",
		},
		{
			nombre: "dos verbos con el mismo nombre",
			applet: appletConVerbos("boe",
				verboDePrueba("articulo", true), verboDePrueba("articulo", false)),
			esperado: ErrVerbosInvalidos,
			mensaje:  "dos veces",
		},
		{
			nombre: "dos verbos marcados por omisión",
			applet: appletConVerbos("boe",
				verboDePrueba("articulo", true), verboDePrueba("buscar", true)),
			esperado: ErrVerbosPorOmision,
			mensaje:  "por omisión",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			var registro Registro
			for _, yaEstaba := range caso.yaEstaban {
				require.NoError(t, registro.Registrar(yaEstaba))
			}

			err := registro.Registrar(caso.applet)

			require.ErrorIs(t, err, caso.esperado)
			require.ErrorContains(t, err, caso.mensaje,
				"el mensaje nombra la causa concreta del rechazo")
			require.ErrorContains(t, err, "app:",
				"el error nombra el paquete que rechaza el registro")
			assert.Len(t, registro.Nombres(), len(caso.yaEstaban),
				"el applet rechazado no entra en el registro")
			assert.Equal(t, schema.ClaseInesperado, cli.Clasificar(err),
				"un registro inválido no lleva ninguno de los seis sentinelas: "+
					"es un defecto de compilación y nunca un código de salida de usuario")
		})
	}
}

// TestRegistroDeProduccion comprueba lo que el binario distribuido registra desde
// H21: boe, graph, mcp, skills y territorio, y nada más, y el almacén del grafo
// del mundo al que el kernel entrega. Los applets de ejemplo no se registran nunca
// aquí, así que `kitlegal echo hola` sobre el binario que se publica termina como
// cualquier otro nombre desconocido (FR-001, FR-009,
// contracts/registro-y-describe.md §3 de H1). Construirlo no pide nada ni abre
// nada: el cliente y la caché de boe se componen en cada invocación (contrato
// puerto-y-applet §4 y §5 de H4), territorio recibe los ficheros que viajan en el
// binario (contrato del applet territorio §7), skills lee lo empotrado la
// primera vez que se ejecuta uno de sus verbos (contracts/applet-skills.md §1 de
// H19), graph y el almacén resuelven la ruta de world.db en cada invocación
// (contracts/applet-graph.md §1; contracts/almacen-world-db.md §1 de H7), y mcp
// solo lee de su entrada cuando sirve (contracts/servidor-mcp.md §1 de H21).
func TestRegistroDeProduccion(t *testing.T) {
	t.Parallel()

	registro, err := RegistroDeProduccion("")
	require.NoError(t, err, "el registro de producción es válido")
	require.NotNil(t, registro)

	assert.Equal(t, []string{"boe", "graph", "mcp", "skills", "territorio"}, registro.Nombres(),
		"el binario distribuido registra exactamente boe, graph, mcp, skills y territorio")

	applet, existe := registro.Buscar("graph")
	require.True(t, existe)
	assert.Equal(t, []string{"show", "stats", "check"}, nombresDeLosVerbos(applet),
		"con sus tres verbos, en el orden del contrato (contracts/applet-graph.md §1)")

	// El almacén de la regla de ubicación de la caché: el que construye
	// graph.Nuevo sin opciones, y no el nulo, que descartaría lo observado
	// (FR-001, FR-030).
	assert.IsType(t, &graph.Almacen{}, registro.almacen,
		"el kernel entrega lo que observa cada invocación al grafo del mundo")
	assert.NotNil(t, registro.almacen)

	applet, existe = registro.Buscar("boe")
	require.True(t, existe)
	assert.Len(t, applet.Verbos(), 6, "con sus seis verbos (FR-001)")

	applet, existe = registro.Buscar("mcp")
	require.True(t, existe)
	assert.Equal(t, []string{"serve"}, nombresDeLosVerbos(applet),
		"con su único verbo (contracts/servidor-mcp.md §1 de H21)")

	applet, existe = registro.Buscar("skills")
	require.True(t, existe)
	assert.Equal(t, []string{"install", "list", "doctor"}, nombresDeLosVerbos(applet),
		"con sus tres verbos, en el orden del contrato (contracts/applet-skills.md §1 de H19)")

	applet, existe = registro.Buscar("territorio")
	require.True(t, existe)
	require.Len(t, applet.Verbos(), 1, "con su único verbo (contrato del applet territorio §1)")
	assert.Equal(t, "resolver", applet.Verbos()[0].Nombre)

	_, existe = registro.Buscar("echo")
	assert.False(t, existe, "el applet de ejemplo no se registra en el binario distribuido")

	t.Run("territorio-sobre-los-ficheros-embebidos", func(t *testing.T) {
		t.Parallel()

		// Una entrada que no llega a ser un código solo la rechaza el applet
		// después de analizar sus fuentes: si las de producción no cargaran, el
		// fallo sería el defecto de composición, código 1 y firmado por el
		// kernel, y no el 2 firmado por territorio (contrato del applet §2 y §7).
		// La consulta no nombra ningún municipio real a propósito (FR-024).
		res := invocar(t, registro, "kitlegal", "territorio", "resolver", "00000", "--json")

		require.Equal(t, 2, res.codigo, res.errores)

		sobre := sobreDelJSON(t, res.salida)
		assert.Equal(t, "kitlegal.territorio", sobre["fuente"], "el fallo lo decide el applet, no el kernel")

		data, esObjeto := sobre["data"].(map[string]any)
		require.True(t, esObjeto, "el data del sobre de fallo es un objeto")
		assert.Equal(t, "argumentos", data["clase"])
		assert.Contains(t, data["mensaje"], "00000", "el mensaje nombra la entrada")
	})
}

// TestRegistroAvisa comprueba el aviso que lleva el registro: el valor cero no
// tiene ninguno, Avisar registra el que se le da y uno nulo lo retira, sin tocar
// los applets registrados (research.md D5 de H19).
func TestRegistroAvisa(t *testing.T) {
	t.Parallel()

	var registro Registro

	require.NoError(t, registro.Registrar(appletConVerbos("boe", verboDePrueba("articulo", true))))

	_, hay := registro.aviso()
	assert.False(t, hay, "el registro de valor cero no tiene aviso")

	registro.Avisar(func() (string, bool) { return "aviso: de prueba", true })

	linea, hay := registro.aviso()
	assert.True(t, hay)
	assert.Equal(t, "aviso: de prueba", linea, "el aviso es el del avisador registrado, tal cual")

	registro.Avisar(nil)

	_, hay = registro.aviso()
	assert.False(t, hay, "un avisador nulo deja el registro sin aviso")
	assert.Equal(t, []string{"boe"}, registro.Nombres(), "el aviso no toca los applets")
}

// TestRegistroDeProduccionAvisa comprueba de extremo a extremo, con el disco
// real, el aviso que compone el registro de producción con su versión y lo que
// el binario empotra (contracts/aviso.md; research.md D4 y D5 de H19): tras
// instalar las skills con un binario, la invocación de otro applet con un binario
// de otra versión lleva la línea del aviso, también cuando falla el verbo, y ni
// skills ni un binario de desarrollo la llevan (FR-070, FR-071, FR-073, FR-074).
//
// No es paralelo: el ámbito local es el directorio de trabajo, que se cambia con
// t.Chdir, y el global sale de HOME; y el kernel lee KITLEGAL_LOG del entorno,
// que se fija en nivelSinEventos para comparar la salida de error byte a byte.
func TestRegistroDeProduccionAvisa(t *testing.T) {
	t.Setenv(cli.VariableNivel, nivelSinEventos)
	t.Setenv(variableHome, t.TempDir())
	enUnProyecto(t)

	instalador, err := RegistroDeProduccion("v0.1.0")
	require.NoError(t, err)

	instalacion := invocar(t, instalador, "kitlegal", "skills", "install")
	require.Equal(t, 0, instalacion.codigo, instalacion.errores)

	nuevo, err := RegistroDeProduccion("v0.2.0")
	require.NoError(t, err)

	const linea = "aviso: las skills instaladas son de kitlegal v0.1.0 y este binario es kitlegal v0.2.0;" +
		" ejecuta: kitlegal skills install\n"

	desarrollo, err := RegistroDeProduccion("dev")
	require.NoError(t, err)

	sinAviso := invocar(t, desarrollo, "kitlegal", "boe", "articulo")
	assert.Equal(t, 2, sinAviso.codigo)
	assert.NotContains(t, sinAviso.errores, "aviso: ", "un binario de desarrollo no compara (FR-073)")

	sinNorma := invocar(t, nuevo, "kitlegal", "boe", "articulo")
	assert.Equal(t, 2, sinNorma.codigo, "el aviso no cambia el código del error de argumentos del verbo")
	assert.Equal(t, linea+sinAviso.errores, sinNorma.errores,
		"la salida de error es la misma con una línea más delante: la del aviso")
	assert.Empty(t, sinNorma.salida)

	lista := invocar(t, nuevo, "kitlegal", "skills", "list")
	assert.Equal(t, 0, lista.codigo, lista.errores)
	assert.NotContains(t, lista.errores, "aviso: ", "skills no avisa (FR-074)")
}

// nombresDeLosVerbos son los nombres de los verbos del applet, en el orden en que
// los declara.
func nombresDeLosVerbos(applet Applet) []string {
	var nombres []string

	for _, verbo := range applet.Verbos() {
		nombres = append(nombres, verbo.Nombre)
	}

	return nombres
}

// TestFuentesEmbebidas comprueba lo que la raíz de producción y el binario de
// e2e pasan a territorio: los cuatro ficheros de data/territorio/ que viajan en
// el binario, con las 19 comunidades y ciudades autónomas indexadas por el código
// que da nombre a cada fichero, y que cargan con toda su integridad —incluida la
// de que ese nombre es el código que el fichero declara dentro— (FR-056,
// contrato de datos §3, data-model §2.1).
func TestFuentesEmbebidas(t *testing.T) {
	t.Parallel()

	fuentes, err := FuentesEmbebidas()
	require.NoError(t, err)

	assert.NotEmpty(t, fuentes.Municipios, "la relación de municipios viaja en el binario")
	assert.NotEmpty(t, fuentes.DIR3, "la correspondencia DIR3 viaja en el binario")
	assert.NotEmpty(t, fuentes.Estado, "el boletín estatal viaja en el binario")
	assert.Len(t, fuentes.Comunidades, 19, "las 17 comunidades y las 2 ciudades autónomas")

	_, err = territorio.Cargar(fuentes)
	require.NoError(t, err, "los ficheros embebidos cargan con toda su integridad")
}
