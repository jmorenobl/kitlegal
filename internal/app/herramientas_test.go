package app_test

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/app/ejemplo"
	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/mcp/mcptest"
)

// appletsSinHerramientas son los dos applets cuyos verbos el servidor no
// anuncia (FR-002 de H21), escritos aquí otra vez y no tomados del código que
// se prueba: una lista derivada de él seguiría cuadrando si dejara de excluir
// uno.
var appletsSinHerramientas = []string{"mcp", "skills"}

// herramientasDeCita son las dos que el servidor anuncia desde H23, con los
// registros de los dos casos, escritas aquí y no derivadas del registro: las de
// los verbos del applet cita (H23 FR-030; contracts/applet-cita.md §7 de H23).
var herramientasDeCita = []string{"cita_cotejar", "cita_preparar"}

// urlDelEsquemaDeHerramienta identifica ante el compilador cada esquema que el
// servidor anuncia. Es una URL de recurso, no una dirección que se visite: el
// compilador resuelve contra ella las referencias `#/$defs/…` del propio
// esquema y nada más.
const urlDelEsquemaDeHerramienta = "https://kitlegal.es/schemas/herramienta.json"

// anotacionDeLasBanderas es la clave con la que la entrada del documento de
// --describe nombra las banderas propias de un verbo, escrita aquí y no tomada
// del código que se prueba (contracts/applet-cita.md §8 de H23).
const anotacionDeLasBanderas = "x-banderas"

// verboConHerramienta es un verbo del registro que el servidor tiene que
// anunciar, con el applet que lo ofrece.
type verboConHerramienta struct {
	applet app.Applet
	verbo  app.Verbo
}

func (v verboConHerramienta) herramienta() string {
	return v.applet.Nombre() + "_" + v.verbo.Nombre
}

// appletDelServidor es el applet que sirve, que el registro de producción trae
// y que estos tests componen sobre su tubería.
const appletDelServidor = "mcp"

// registroParaServir es el registro sobre el que el test arranca el servidor:
// los applets del registro de producción, los mismos valores, salvo mcp, y, si
// se piden, los de ejemplo. El mcp de producción lee de la entrada estándar del
// proceso, a la que un test en proceso no puede hablar, así que el de este
// registro lo registra arrancarElServidor sobre su tubería: es el mismo applet
// con otra entrada. Es un registro local, sin almacén al que entregar: el test
// no llama a ninguna herramienta, y ninguno escribe en el world.db de la cuenta
// de quien lo ejecuta.
func registroParaServir(t *testing.T, distribuido *app.Registro, conLosDeEjemplo bool) *app.Registro {
	t.Helper()

	require.Contains(t, distribuido.Nombres(), appletDelServidor,
		"el registro de producción trae el applet %s (FR-001)", appletDelServidor)

	var applets []app.Applet

	for _, nombre := range distribuido.Nombres() {
		if nombre == appletDelServidor {
			continue
		}

		applet, registrado := distribuido.Buscar(nombre)
		require.True(t, registrado, nombre)

		applets = append(applets, applet)
	}

	if conLosDeEjemplo {
		applets = append(applets, ejemplo.Applets()...)
	}

	return registroCon(t, applets...)
}

// verbosConHerramienta son los verbos del registro menos los de los applets
// excluidos, por el nombre de su herramienta.
func verbosConHerramienta(t *testing.T, registro *app.Registro) map[string]verboConHerramienta {
	t.Helper()

	for _, excluido := range appletsSinHerramientas {
		require.Contains(t, registro.Nombres(), excluido,
			"el registro lleva el applet %s, de modo que excluirlo se ejerce", excluido)
	}

	verbos := map[string]verboConHerramienta{}

	for _, nombre := range registro.Nombres() {
		if slices.Contains(appletsSinHerramientas, nombre) {
			continue
		}

		applet, registrado := registro.Buscar(nombre)
		require.True(t, registrado, nombre)

		for _, verbo := range applet.Verbos() {
			conHerramienta := verboConHerramienta{applet: applet, verbo: verbo}
			verbos[conHerramienta.herramienta()] = conHerramienta
		}
	}

	return verbos
}

// TestHerramientasDelServidor es el control de conformidad de FR-070 de H21,
// con el servidor arrancado en proceso y el cliente de prueba de cada
// especificación, sobre los applets del registro de producción y sobre un
// registro local que lleva además los de ejemplo (FR-004): el conjunto
// anunciado es el de los verbos del registro menos los excluidos —doce y
// quince, con cita_cotejar y cita_preparar desde H23 (H23 FR-030)—; cada
// nombre, cada descripción y cada par de esquemas es el de
// `--describe` de su verbo, el de entrada sin las ocho banderas globales; los
// nombres que da app.NombresDeHerramientas son los anunciados; cada
// `$ref` resuelve en su esquema; todas se anuncian de solo lectura; las
// capacidades son `{"tools":{}}`; y el tipo JSON que cli.LineaDeLlamada exige
// de cada argumento es el que declara su esquema (FR-002, FR-003, FR-005,
// FR-008, FR-020; SC-003).
//
// Lo esperado —los verbos, sus descripciones y sus documentos de `--describe`—
// sale del registro de producción tal cual, que ya trae el applet mcp; con los
// applets de ejemplo, que producción no tiene, del registro local.
func TestHerramientasDelServidor(t *testing.T) {
	t.Parallel()

	registros := []struct {
		nombre          string
		conLosDeEjemplo bool
		herramientas    int
	}{
		{nombre: "con los applets de producción", herramientas: 12},
		{nombre: "con los de producción y los de ejemplo", conLosDeEjemplo: true, herramientas: 15},
	}

	especificaciones := []struct {
		nombre   string
		anterior bool
	}{
		{nombre: "el cliente de la especificación vigente"},
		{nombre: "el cliente de la especificación anterior", anterior: true},
	}

	for _, caso := range registros {
		for _, cliente := range especificaciones {
			t.Run(caso.nombre+", "+cliente.nombre, func(t *testing.T) {
				t.Parallel()

				distribuido, err := app.RegistroDeProduccion("")
				require.NoError(t, err)

				servido := registroParaServir(t, distribuido, caso.conLosDeEjemplo)
				servidor := arrancarElServidor(t, servido, cliente.anterior)

				registro := servido
				if !caso.conLosDeEjemplo {
					require.Equal(t, distribuido.Nombres(), servido.Nombres(),
						"el servidor se arranca sobre los applets del registro de producción, y ninguno más")

					registro = distribuido
				}

				esperados := verbosConHerramienta(t, registro)

				anunciadas, err := servidor.sesion.Herramientas(t.Context())
				require.NoError(t, err)

				compruebaElConjuntoAnunciado(t, anunciadas, esperados, caso.herramientas)
				assert.ElementsMatch(t, slices.Collect(maps.Keys(esperados)), app.NombresDeHerramientas(registro),
					"los nombres que el paquete da a quien lee las llamadas de una sesión son los de las "+
						"herramientas que el servidor anuncia, sin las de %s", strings.Join(appletsSinHerramientas, " ni "))

				for _, anunciada := range anunciadas {
					verbo, esperada := esperados[anunciada.Nombre]
					if !esperada {
						// Ya la ha nombrado la comprobación del conjunto.
						continue
					}

					compruebaLaHerramienta(t, registro, anunciada, verbo)
				}

				assert.Equal(t, []string{"tools"}, servidor.sesion.Capacidades(), "solo herramientas (FR-008)")

				require.Equal(t, 0, servidor.cerrar(t), servidor.errores.lineas())

				if cliente.anterior {
					compruebaLasCapacidadesEnElCable(t, servidor.sesion.Lineas())
				}
			})
		}
	}
}

// compruebaElConjuntoAnunciado exige que las herramientas anunciadas sean
// exactamente las de los verbos esperados, en el orden del servidor —por
// nombre—, y nombra las que sobran y las que faltan.
func compruebaElConjuntoAnunciado(
	t *testing.T, anunciadas []mcptest.Herramienta, esperados map[string]verboConHerramienta, cuantas int,
) {
	t.Helper()

	nombres := make([]string, 0, len(anunciadas))
	for _, anunciada := range anunciadas {
		nombres = append(nombres, anunciada.Nombre)
	}

	assert.ElementsMatch(t, slices.Collect(maps.Keys(esperados)), nombres,
		"el servidor anuncia una herramienta por cada verbo del registro, salvo los de %s (FR-002)",
		strings.Join(appletsSinHerramientas, " y "))
	assert.Len(t, esperados, cuantas, "los verbos del registro menos los excluidos")
	assert.True(t, slices.IsSorted(nombres), "el orden de la lista es el del SDK, por nombre: %q", nombres)
	assert.Subset(t, nombres, herramientasDeCita,
		"los dos verbos de cita llegan como herramientas sin tocar el servidor (H23 FR-030)")
}

// descripcionDelVerbo es el documento que un verbo emite con --describe, con
// sus dos partes y las definiciones que comparten.
type descripcionDelVerbo struct {
	Titulo      string `json:"title"`
	Descripcion string `json:"description"`
	Partes      struct {
		Entrada map[string]any `json:"entrada"`
		Salida  map[string]any `json:"salida"`
	} `json:"properties"`
	Definiciones map[string]any `json:"$defs"`
}

// compruebaLaHerramienta exige que una herramienta anunciada sea la de su
// verbo: su nombre y su descripción, sus dos esquemas, que es de solo lectura
// y que una llamada exige de cada argumento el tipo que su esquema declara.
func compruebaLaHerramienta(
	t *testing.T, registro *app.Registro, anunciada mcptest.Herramienta, verbo verboConHerramienta,
) {
	t.Helper()

	nombre := anunciada.Nombre

	describe := invocarCon(t, registro, ordenDeDescribe(t, verbo, anunciada.Entrada)...)
	require.Equal(t, 0, describe.codigo, "%s --describe: %s", nombre, describe.errores)

	var descrito descripcionDelVerbo
	require.NoError(t, json.Unmarshal([]byte(describe.salida), &descrito), nombre)
	require.Equal(t, verbo.applet.Nombre()+" "+verbo.verbo.Nombre, descrito.Titulo,
		"el documento de --describe es el de este verbo")

	assert.Equal(t, descrito.Descripcion, anunciada.Descripcion, "%s: la descripción es la de su verbo (FR-003)", nombre)
	assert.True(t, anunciada.SoloLectura, "%s se anuncia de solo lectura (FR-005)", nombre)

	entrada, definicionesDeEntrada := esquemaAnunciado(t, nombre, anunciada.Entrada)
	assert.Equal(t, sinLasBanderasGlobales(t, nombre, descrito.Partes.Entrada, entrada), entrada,
		"%s: el esquema de entrada es la entrada de --describe sin las ocho banderas globales (FR-003)", nombre)

	salida, definicionesDeSalida := esquemaAnunciado(t, nombre, anunciada.Salida)
	assert.Equal(t, descrito.Partes.Salida, salida, "%s: el esquema de salida es la salida de --describe (FR-003)", nombre)

	for _, definiciones := range []map[string]any{definicionesDeEntrada, definicionesDeSalida} {
		for clave, esquema := range definiciones {
			assert.Equal(t, descrito.Definiciones[clave], esquema,
				"%s: la definición %s es la del documento de --describe", nombre, clave)
		}
	}

	compruebaLosTiposDeLaLlamada(t, cli.Verbo{
		Applet:     verbo.applet.Nombre(),
		Verbo:      verbo.verbo.Nombre,
		Ayuda:      verbo.verbo.Descripcion,
		Argumentos: verbo.verbo.Argumentos(),
		Salida:     verbo.verbo.Salida,
	}, entrada)
}

// ordenDeDescribe es la orden que describe un verbo. El análisis de la
// invocación va antes que la descripción, así que la orden lleva un valor
// cualquiera por cada argumento obligatorio, que --describe no lee: tantos
// como el esquema anunciado declara obligatorios. Si esos no fueran los de la
// orden, terminaría con 2 y el test lo diría.
func ordenDeDescribe(t *testing.T, verbo verboConHerramienta, entrada json.RawMessage) []string {
	t.Helper()

	var declarada struct {
		Obligatorios []string `json:"required"`
	}

	require.NoError(t, json.Unmarshal(entrada, &declarada), verbo.herramienta())

	orden := []string{verbo.applet.Nombre(), verbo.verbo.Nombre}
	for range declarada.Obligatorios {
		orden = append(orden, "x")
	}

	return append(orden, "--describe")
}

// esquemaAnunciado lee un esquema de los que el servidor anuncia y lo devuelve
// separado de sus `$defs`, después de comprobar que se vale solo: un validador
// lo compila, de modo que cada `$ref` resuelve en el propio esquema.
func esquemaAnunciado(t *testing.T, herramienta string, anunciado json.RawMessage) (esquema, definiciones map[string]any) {
	t.Helper()

	require.NotEmpty(t, anunciado, "%s anuncia sus dos esquemas", herramienta)

	documento, err := jsonschema.UnmarshalJSON(strings.NewReader(string(anunciado)))
	require.NoError(t, err, herramienta)

	compilador := jsonschema.NewCompiler()
	require.NoError(t, compilador.AddResource(urlDelEsquemaDeHerramienta, documento), herramienta)

	_, err = compilador.Compile(urlDelEsquemaDeHerramienta)
	require.NoError(t, err, "%s: el esquema anunciado se compila, con cada $ref resuelto en él: %s", herramienta, anunciado)

	require.NoError(t, json.Unmarshal(anunciado, &esquema), herramienta)

	if propias, lleva := esquema["$defs"]; lleva {
		var sonUnObjeto bool

		definiciones, sonUnObjeto = propias.(map[string]any)
		require.True(t, sonUnObjeto, "%s: $defs es un objeto", herramienta)
		require.NotEmpty(t, definiciones, "%s: un esquema que no referencia nada no lleva $defs", herramienta)

		delete(esquema, "$defs")
	}

	return esquema, definiciones
}

// sinLasBanderasGlobales es la entrada de --describe sin las ocho banderas
// globales, que tiene que llevar todas, ni la anotación de las banderas propias
// del verbo, si la lleva; y exige que el esquema anunciado no lleve ninguna de
// las ocho: las banderas son del servidor y nunca un parámetro de una
// herramienta (FR-020). Tampoco lleva la anotación, que dice cómo se escribe la
// orden: una herramienta recibe un objeto (research.md D14 de H23).
func sinLasBanderasGlobales(t *testing.T, herramienta string, descrita, anunciada map[string]any) map[string]any {
	t.Helper()

	assert.NotContains(t, anunciada, anotacionDeLasBanderas,
		"%s: el esquema de entrada no lleva la anotación de las banderas propias", herramienta)
	delete(descrita, anotacionDeLasBanderas)

	deLaOrden, sonUnObjeto := descrita["properties"].(map[string]any)
	require.True(t, sonUnObjeto, "%s: la entrada de --describe declara sus propiedades", herramienta)

	deLaHerramienta, sonUnObjeto := anunciada["properties"].(map[string]any)
	require.True(t, sonUnObjeto, "%s: el esquema de entrada declara sus propiedades", herramienta)

	for _, global := range banderasGlobales {
		propiedad := strings.TrimPrefix(global.nombre, "--")

		require.Contains(t, deLaOrden, propiedad, "%s: la entrada de --describe lleva %s", herramienta, global.nombre)
		assert.NotContains(t, deLaHerramienta, propiedad,
			"%s: el esquema de entrada no lleva la bandera global %s (FR-020)", herramienta, global.nombre)

		delete(deLaOrden, propiedad)
	}

	return descrita
}

// muestraDeArgumento es un valor JSON de una de las formas que una llamada
// puede dar a un argumento: el tipo de JSON Schema que lo describe y si es una
// lista de él. Las que no llevan tipo no las declara ningún esquema de entrada.
type muestraDeArgumento struct {
	tipo  string
	lista bool
	valor string
}

func muestrasDeArgumento() []muestraDeArgumento {
	return []muestraDeArgumento{
		{tipo: "string", valor: `"x"`},
		{tipo: "integer", valor: `3`},
		{tipo: "boolean", valor: `true`},
		{tipo: "string", lista: true, valor: `["x"]`},
		{tipo: "integer", lista: true, valor: `[3]`},
		{tipo: "boolean", lista: true, valor: `[true]`},
		{valor: `1.5`},
		{valor: `{"x":"y"}`},
		{valor: `null`},
	}
}

// compruebaLosTiposDeLaLlamada exige que el tipo JSON que cli.LineaDeLlamada
// admite de cada argumento sea el que declara el esquema de entrada anunciado,
// y ningún otro: con unos argumentos completos y válidos, cambia el valor de
// cada uno por una muestra de cada forma, y la conversión la acepta si y solo
// si es de la forma declarada (FR-070).
func compruebaLosTiposDeLaLlamada(t *testing.T, def cli.Verbo, entrada map[string]any) {
	t.Helper()

	herramienta := def.Applet + "_" + def.Verbo
	muestras := muestrasDeArgumento()

	propiedades, sonUnObjeto := entrada["properties"].(map[string]any)
	require.True(t, sonUnObjeto, herramienta)

	// Unos argumentos completos, con una muestra de la forma declarada en cada
	// propiedad: es lo que deja cambiar una sola sin que falte la anterior.
	validos := map[string]json.RawMessage{}
	declaradas := map[string]muestraDeArgumento{}

	for propiedad, esquema := range propiedades {
		declarada := formaDeclarada(t, herramienta, propiedad, esquema)

		indice := slices.IndexFunc(muestras, func(muestra muestraDeArgumento) bool {
			return muestra.tipo == declarada.tipo && muestra.lista == declarada.lista
		})
		require.GreaterOrEqual(t, indice, 0,
			"%s: el esquema declara para %q una forma que este test no sabe comprobar: %v", herramienta, propiedad, esquema)

		declaradas[propiedad] = muestras[indice]
		validos[propiedad] = json.RawMessage(muestras[indice].valor)
	}

	_, err := cli.LineaDeLlamada(def, argumentosDe(t, validos))
	require.NoError(t, err, "%s: unos argumentos de los tipos que declara su esquema valen", herramienta)

	for propiedad, declarada := range declaradas {
		for _, muestra := range muestras {
			cambiados := maps.Clone(validos)
			cambiados[propiedad] = json.RawMessage(muestra.valor)

			_, err := cli.LineaDeLlamada(def, argumentosDe(t, cambiados))

			if muestra == declarada {
				require.NoError(t, err,
					"%s: %q admite %s, que es lo que declara su esquema", herramienta, propiedad, muestra.valor)

				continue
			}

			if assert.ErrorIs(t, err, cli.ErrArgumentos,
				"%s: %q no admite %s, que no es lo que declara su esquema", herramienta, propiedad, muestra.valor) {
				assert.Contains(t, err.Error(), `el argumento "`+propiedad+`" de `+herramienta+" tiene que ser ")
			}
		}
	}
}

// formaDeclarada es el tipo que el esquema de entrada declara para una
// propiedad, y si es una lista de él.
func formaDeclarada(t *testing.T, herramienta, propiedad string, esquema any) muestraDeArgumento {
	t.Helper()

	declarado, esUnObjeto := esquema.(map[string]any)
	require.True(t, esUnObjeto, "%s: el esquema de %q es un objeto", herramienta, propiedad)

	tipo, _ := declarado["type"].(string)
	if tipo != "array" {
		return muestraDeArgumento{tipo: tipo}
	}

	elementos, esUnObjeto := declarado["items"].(map[string]any)
	require.True(t, esUnObjeto, "%s: la lista %q declara sus elementos", herramienta, propiedad)

	tipo, _ = elementos["type"].(string)

	return muestraDeArgumento{tipo: tipo, lista: true}
}

// argumentosDe es el objeto JSON de los argumentos de una llamada.
func argumentosDe(t *testing.T, propiedades map[string]json.RawMessage) []byte {
	t.Helper()

	argumentos, err := json.Marshal(propiedades)
	require.NoError(t, err)

	return argumentos
}

// compruebaLasCapacidadesEnElCable exige que el servidor anuncie en su saludo
// exactamente `{"tools":{}}`, leído de lo que escribe: ni `listChanged`, ni
// `logging`, ni `prompts`, ni `resources` (FR-008; contracts/servidor-mcp.md
// §6). Lo lee de las líneas que guarda el cliente anterior, cuya primera línea
// es la respuesta a `initialize`.
func compruebaLasCapacidadesEnElCable(t *testing.T, lineas []string) {
	t.Helper()

	require.NotEmpty(t, lineas, "el cliente anterior guarda lo que el servidor escribe")

	var saludo struct {
		Resultado struct {
			Capacidades json.RawMessage `json:"capabilities"`
		} `json:"result"`
	}

	require.NoError(t, json.Unmarshal([]byte(lineas[0]), &saludo))
	assert.JSONEq(t, `{"tools":{}}`, string(saludo.Resultado.Capacidades))
}

// TestHerramientasAnunciadas es el control en `make ci` de FR-014 de H22: lo
// que app.HerramientasAnunciadas da de un registro, que es lo que el paso que
// empaqueta la extensión escribe en `tools` de su manifiesto sin arrancar
// ningún servidor, es lo que el servidor lista por MCP con ese registro. Con el
// servidor arrancado en proceso y el cliente de prueba de cada especificación,
// sobre los applets del registro de producción y sobre un registro local que
// lleva además los de ejemplo: el nombre y la descripción de cada herramienta
// de la lista son los que da la función, comparados como conjunto y no por
// posición, porque el orden de la lista lo pone el SDK, por nombre, y el de la
// función es el del registro —applets por nombre, verbos en el orden de su
// catálogo—, que se comprueba aparte (FR-014, FR-061; SC-004; research.md D15
// de H22).
//
// La lista sale de lo que el servidor escribe, no del registro: una función que
// omita una herramienta o cambie una descripción deja de cuadrar con ella.
func TestHerramientasAnunciadas(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre          string
		conLosDeEjemplo bool
		anterior        bool
	}{
		{nombre: "con los applets de producción, el cliente de la especificación vigente"},
		{nombre: "con los applets de producción, el cliente de la especificación anterior", anterior: true},
		{
			nombre:          "con los de producción y los de ejemplo, el cliente de la especificación vigente",
			conLosDeEjemplo: true,
		},
		{
			nombre:          "con los de producción y los de ejemplo, el cliente de la especificación anterior",
			conLosDeEjemplo: true,
			anterior:        true,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			registro, listadas := herramientasListadas(t, caso.conLosDeEjemplo, caso.anterior)

			anunciadas := app.HerramientasAnunciadas(registro)

			assert.ElementsMatch(t, listadas, anunciadas,
				"el nombre y la descripción de cada herramienta que el servidor lista son los que da "+
					"app.HerramientasAnunciadas, ni una de más ni una de menos (FR-014)")

			nombres := make([]string, 0, len(anunciadas))
			for _, anunciada := range anunciadas {
				nombres = append(nombres, anunciada.Nombre)
			}

			assert.Equal(t, herramientasEnElOrdenDelRegistro(t, registro), nombres,
				"app.HerramientasAnunciadas las da en el orden del registro: applets por nombre, verbos en el de su catálogo")
		})
	}
}

// herramientasListadas arranca el servidor sobre los applets del registro de
// producción y, si se piden, los de ejemplo, le pide su lista con el cliente de
// esa especificación y la devuelve reducida al nombre y a la descripción de
// cada herramienta, junto al registro del que el servidor las saca: el de
// producción tal cual, que es el que lee el paso que empaqueta, o el local que
// lleva además los de ejemplo, que producción no tiene.
func herramientasListadas(
	t *testing.T, conLosDeEjemplo, anterior bool,
) (*app.Registro, []app.HerramientaAnunciada) {
	t.Helper()

	distribuido, err := app.RegistroDeProduccion("")
	require.NoError(t, err)

	servido := registroParaServir(t, distribuido, conLosDeEjemplo)
	servidor := arrancarElServidor(t, servido, anterior)

	anunciadas, err := servidor.sesion.Herramientas(t.Context())
	require.NoError(t, err)
	require.NotEmpty(t, anunciadas, "el servidor lista sus herramientas: una lista vacía cuadraría con cualquier omisión")

	require.Equal(t, 0, servidor.cerrar(t), servidor.errores.lineas())

	listadas := make([]app.HerramientaAnunciada, 0, len(anunciadas))
	for _, anunciada := range anunciadas {
		listadas = append(listadas, app.HerramientaAnunciada{
			Nombre:      anunciada.Nombre,
			Descripcion: anunciada.Descripcion,
		})
	}

	if conLosDeEjemplo {
		return servido, listadas
	}

	require.Equal(t, distribuido.Nombres(), servido.Nombres(),
		"el servidor se arranca sobre los applets del registro de producción, y ninguno más")

	return distribuido, listadas
}

// herramientasEnElOrdenDelRegistro son los nombres de las herramientas de un
// registro en el orden que el registro da: sus applets por nombre, menos los
// excluidos, y los verbos de cada uno en el orden de su catálogo. Lo saca del
// registro por su cuenta, sin el código que se prueba.
func herramientasEnElOrdenDelRegistro(t *testing.T, registro *app.Registro) []string {
	t.Helper()

	var nombres []string

	for _, nombre := range registro.Nombres() {
		if slices.Contains(appletsSinHerramientas, nombre) {
			continue
		}

		applet, registrado := registro.Buscar(nombre)
		require.True(t, registrado, nombre)

		for _, verbo := range applet.Verbos() {
			nombres = append(nombres, verboConHerramienta{applet: applet, verbo: verbo}.herramienta())
		}
	}

	return nombres
}
