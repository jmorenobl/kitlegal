package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// entradaDeArticulo es el esquema de entrada que contracts/servidor-mcp.md §2 da
// como ejemplo para `boe_articulo`, carácter a carácter: 144 bytes (research V25).
const entradaDeArticulo = `{"properties":{"norma":{"type":"string"},"bloque":{"type":"string"}},` +
	`"additionalProperties":false,"type":"object","required":["norma","bloque"]}`

// entradaSinArgumentos es la de un verbo que no declara ninguno, como
// `graph_stats`: los 62 bytes del mínimo de research V25.
const entradaSinArgumentos = `{"properties":{},"additionalProperties":false,"type":"object"}`

// argumentosInvalidos es el comienzo de todo mensaje de la clase `argumentos`,
// escrito aquí y no leído del sentinela: es el del contrato.
const argumentosInvalidos = "argumentos inv\xc3\xa1lidos: "

// Los argumentos de los verbos sintéticos de este test. Los dos primeros tienen
// la forma de dos verbos de hoy —`boe articulo` y `graph check`—, que son los de
// los ejemplos del contrato; el tercero reúne lo que hoy ningún verbo declara y
// la conversión tiene que saber escribir: banderas, un booleano y enteros.
type (
	argumentosDeArticulo struct {
		Norma  string `arg:"" name:"norma" help:"Identificador de la norma."`
		Bloque string `arg:"" name:"bloque" help:"Identificador del bloque."`
	}

	argumentosDeComprobacion struct {
		Norma   *string  `arg:"" optional:"" name:"norma" help:"Identificador de la norma."`
		Bloques []string `arg:"" optional:"" name:"bloques" help:"Identificadores de los bloques."`
	}

	argumentosVariados struct {
		Etiqueta string   `help:"Una bandera de cadena."`
		Forzar   bool     `help:"Una bandera booleana."`
		Limite   int      `default:"10" help:"Una bandera entera."`
		Materias []string `help:"Una bandera de lista de cadenas."`
		Norma    string   `arg:"" help:"Una cadena por su posición."`
		Veces    int      `arg:"" optional:"" help:"Un entero por su posición."`
		Bloques  []string `arg:"" optional:"" help:"Una lista de cadenas por su posición."`
	}

	// argumentosConFiltro declara un campo que es un struct: su esquema es una
	// referencia, y es lo que hace que la entrada tenga definiciones propias.
	argumentosConFiltro struct {
		Filtro filtroDePrueba `help:"Un filtro."`
	}

	filtroDePrueba struct {
		Desde string `json:"desde"`
	}

	// datosConAvisos es un `data` cuya definición referencia otra: la que solo se
	// alcanza siguiendo las referencias de una definición.
	datosConAvisos struct {
		Texto  string          `json:"texto"`
		Avisos []avisoDePrueba `json:"avisos"`
	}

	avisoDePrueba struct {
		Clase string `json:"clase"`
	}
)

// Los nombres de esas definiciones en los `$defs`.
const (
	definicionDelFiltro  = "cli.filtroDePrueba"
	definicionConAvisos  = "cli.datosConAvisos"
	definicionDeUnAviso  = "cli.avisoDePrueba"
	claveDeDefiniciones  = "$defs"
	claveDeReferencia    = "$ref"
	claveDePropiedades   = "properties"
	herramientaArticulo  = "boe_articulo"
	herramientaComprobar = "graph_check"
)

// verbosDeLlamada es la gramática de los tres verbos, que es con la que el
// analizador de la orden lee la línea que devuelve la conversión.
type verbosDeLlamada struct {
	Articulo argumentosDeArticulo     `cmd:"" help:"Lee un bloque de una norma."`
	Check    argumentosDeComprobacion `cmd:"" help:"Comprueba lo consultado."`
	Variar   argumentosVariados       `cmd:"" help:"Ejerce cada tipo de argumento."`
}

func verboDeArticulo() Verbo {
	return Verbo{
		Applet:     "boe",
		Verbo:      "articulo",
		Ayuda:      "Lee un bloque de una norma.",
		Argumentos: &argumentosDeArticulo{},
		Salida:     datosConAvisos{},
	}
}

func verboDeComprobacion() Verbo {
	return Verbo{
		Applet:     "graph",
		Verbo:      "check",
		Ayuda:      "Comprueba lo consultado.",
		Argumentos: &argumentosDeComprobacion{},
		Salida:     datosDeCuenta{},
	}
}

func verboVariado() Verbo {
	return Verbo{
		Applet:     nombreDePrueba,
		Verbo:      "variar",
		Ayuda:      "Ejerce cada tipo de argumento.",
		Argumentos: &argumentosVariados{},
	}
}

func verboConFiltro() Verbo {
	return Verbo{
		Applet:     nombreDePrueba,
		Verbo:      "filtrar",
		Ayuda:      "Filtra lo que se le da.",
		Argumentos: &argumentosConFiltro{},
		Salida:     datosConAvisos{},
	}
}

func verboSinArgumentos() Verbo {
	return Verbo{Applet: "graph", Verbo: "stats", Ayuda: "Cuenta lo que hay."}
}

// verbosDeLosEsquemas son los que recorren las comprobaciones que valen para
// cualquier verbo: con argumentos y sin ellos, con salida declarada y sin ella,
// con definiciones en la entrada y sin ellas.
func verbosDeLosEsquemas() map[string]Verbo {
	return map[string]Verbo{
		"uno por su posición":           verboDePrueba(),
		"dos posicionales":              verboDeArticulo(),
		"posicionales opcionales":       verboDeComprobacion(),
		"banderas y posicionales":       verboVariado(),
		"un argumento que es un tipo":   verboConFiltro(),
		"ningún argumento ni salida":    verboSinArgumentos(),
		"argumentos con campo embebido": verboConMarcas(),
	}
}

// esquemasDePrueba pide los dos esquemas de un verbo y los devuelve tal y como
// salen —los bytes— y ya leídos.
func esquemasDePrueba(t *testing.T, def Verbo) (entrada, salida []byte, leidos [2]map[string]any) {
	t.Helper()

	entrada, salida, err := EsquemasDeHerramienta(def)
	require.NoError(t, err)

	return entrada, salida, [2]map[string]any{
		unicoDocumento(t, string(entrada)),
		unicoDocumento(t, string(salida)),
	}
}

// sinDefiniciones es el mismo esquema sin sus `$defs`, que es la forma en que se
// compara con una parte del documento de --describe: allí las definiciones de
// las dos partes van juntas, en la raíz.
func sinDefiniciones(esquema map[string]any) map[string]any {
	copia := make(map[string]any, len(esquema))

	for clave, valor := range esquema {
		if clave != claveDeDefiniciones {
			copia[clave] = valor
		}
	}

	return copia
}

// definicionesDe son los `$defs` de un esquema; ninguno, si no los lleva.
func definicionesDe(t *testing.T, esquema map[string]any) map[string]any {
	t.Helper()

	if _, lleva := esquema[claveDeDefiniciones]; !lleva {
		return map[string]any{}
	}

	return bajar(t, esquema, claveDeDefiniciones)
}

// referenciasDe recoge los nombres de las definiciones que un trozo de esquema
// referencia, y falla si una referencia no apunta a los `$defs` de su documento.
func referenciasDe(t *testing.T, valor any, nombres map[string]bool) {
	t.Helper()

	switch nodo := valor.(type) {
	case map[string]any:
		for clave, hijo := range nodo {
			referencia, esTexto := hijo.(string)
			if clave != claveDeReferencia || !esTexto {
				referenciasDe(t, hijo, nombres)

				continue
			}

			nombre, interna := strings.CutPrefix(referencia, prefijoDeDefinicion)
			require.True(t, interna, "la referencia %s no apunta a los $defs del documento", referencia)

			nombres[nombre] = true
		}
	case []any:
		for _, hijo := range nodo {
			referenciasDe(t, hijo, nombres)
		}
	}
}

// definicionesAlcanzadas son las que un esquema referencia, directamente o desde
// otra definición: lo que el documento tiene que llevar, ni más ni menos.
func definicionesAlcanzadas(t *testing.T, esquema map[string]any) []string {
	t.Helper()

	definiciones := definicionesDe(t, esquema)
	alcanzadas := map[string]bool{}
	referenciasDe(t, sinDefiniciones(esquema), alcanzadas)

	for pendientes := claves(marcas(alcanzadas)); len(pendientes) > 0; {
		nombre := pendientes[0]
		pendientes = pendientes[1:]

		referida, existe := definiciones[nombre]
		require.True(t, existe, "el documento referencia %s y no la define", nombre)

		nuevas := map[string]bool{}
		referenciasDe(t, referida, nuevas)

		for otra := range nuevas {
			if !alcanzadas[otra] {
				alcanzadas[otra] = true
				pendientes = append(pendientes, otra)
			}
		}
	}

	return claves(marcas(alcanzadas))
}

// marcas presenta un conjunto de nombres como el objeto que `claves` ordena.
func marcas(conjunto map[string]bool) map[string]any {
	objeto := make(map[string]any, len(conjunto))
	for nombre := range conjunto {
		objeto[nombre] = true
	}

	return objeto
}

// TestEsquemasDeHerramienta comprueba los dos esquemas que una herramienta
// anuncia de su verbo: que son las dos partes del documento de --describe, la de
// entrada sin las ocho banderas globales, y que cada uno es un documento que se
// vale solo —con las definiciones que referencia, sin `$schema` y en JSON
// compacto— (FR-003, FR-020; contracts/servidor-mcp.md §2; research D6).
func TestEsquemasDeHerramienta(t *testing.T) {
	t.Parallel()

	subtests := []struct {
		nombre string
		probar func(*testing.T)
	}{
		{"la entrada del ejemplo del contrato, byte a byte", probarEntradaDelContrato},
		{"la entrada no lleva ninguna bandera global", probarEntradaSinGlobales},
		{"los dos esquemas son las partes del documento de --describe", probarPartesDeDescribe},
		{"cada esquema lleva las definiciones que referencia y solo esas", probarDefinicionesDeCadaEsquema},
		{"sin $schema y en JSON compacto", probarEsquemasCompactos},
		{"cada esquema vale suelto ante un validador", probarEsquemasSueltos},
		{"un verbo que no se puede describir no tiene esquemas", probarEsquemasImposibles},
	}

	for _, subtest := range subtests {
		t.Run(subtest.nombre, func(t *testing.T) {
			t.Parallel()

			subtest.probar(t)
		})
	}
}

// exigirByteAByte compara lo escrito con lo esperado como textos y no como
// documentos JSON: lo que se fija es cada byte, también el orden de las claves,
// porque es lo que el contrato escribe y lo que research V25 mide.
func exigirByteAByte(t *testing.T, esperado string, escrito []byte) {
	t.Helper()

	assert.Equal(t, esperado, string(escrito))
}

func probarEntradaDelContrato(t *testing.T) {
	t.Helper()

	entrada, _, _ := esquemasDePrueba(t, verboDeArticulo())
	exigirByteAByte(t, entradaDeArticulo, entrada)
	assert.Len(t, entrada, 144, "los bytes que mide research V25")

	// Un verbo sin argumentos tiene una entrada cerrada y vacía: los 62 bytes del
	// mínimo de esa misma medida.
	entrada, _, _ = esquemasDePrueba(t, verboSinArgumentos())
	exigirByteAByte(t, entradaSinArgumentos, entrada)
	assert.Len(t, entrada, 62)
}

func probarEntradaSinGlobales(t *testing.T) {
	t.Helper()

	propias := map[string][]string{
		"uno por su posición":        {"texto"},
		"dos posicionales":           {"norma", "bloque"},
		"posicionales opcionales":    {"norma", "bloques"},
		"banderas y posicionales":    {"etiqueta", "forzar", "limite", "materias", "norma", "veces", "bloques"},
		"ningún argumento ni salida": {},
	}

	for nombre, esperadas := range propias {
		def, existe := verbosDeLosEsquemas()[nombre]
		require.True(t, existe, "%s no es ningún verbo de este test", nombre)

		_, _, leidos := esquemasDePrueba(t, def)
		entrada := leidos[0]
		propiedades := bajar(t, entrada, claveDePropiedades)

		for _, global := range nombresDeLasOcho {
			assert.NotContains(t, propiedades, global,
				"%s: la bandera global %s no es un argumento de la herramienta", nombre, global)
		}

		assert.ElementsMatch(t, esperadas, claves(propiedades), "%s: los campos del verbo y nada más", nombre)
		assert.Equal(t, "object", entrada["type"], nombre)
		assert.Equal(t, false, entrada["additionalProperties"],
			"%s: una propiedad de más es un argumento inválido (FR-020)", nombre)
	}
}

func probarPartesDeDescribe(t *testing.T) {
	t.Helper()

	for nombre, def := range verbosDeLosEsquemas() {
		documento, _ := describirDePrueba(t, def)
		_, _, leidos := esquemasDePrueba(t, def)

		// La parte de entrada del documento, sin las ocho: ninguna es obligatoria,
		// así que quitarlas de las propiedades es quitarlas del todo.
		parteDeEntrada := bajar(t, documento, claveDePropiedades, "entrada")
		propiedades := bajar(t, parteDeEntrada, claveDePropiedades)

		for _, global := range nombresDeLasOcho {
			delete(propiedades, global)
		}

		assert.Equal(t, parteDeEntrada, sinDefiniciones(leidos[0]), "%s: la entrada", nombre)
		assert.Equal(t, bajar(t, documento, claveDePropiedades, "salida"), sinDefiniciones(leidos[1]),
			"%s: la salida", nombre)

		delDocumento := definicionesDe(t, documento)

		for _, esquema := range leidos {
			for referida, forma := range definicionesDe(t, esquema) {
				assert.Equal(t, delDocumento[referida], forma,
					"%s: la definición %s es la del documento", nombre, referida)
			}
		}
	}
}

func probarDefinicionesDeCadaEsquema(t *testing.T) {
	t.Helper()

	_, _, leidos := esquemasDePrueba(t, verboConFiltro())
	assert.Equal(t, []string{definicionDelFiltro}, claves(definicionesDe(t, leidos[0])),
		"la entrada lleva la suya y ninguna de la salida")
	assert.Equal(t, []string{definicionDeError, definicionDeUnAviso, definicionConAvisos},
		claves(definicionesDe(t, leidos[1])),
		"la salida lleva las suyas —también la que solo referencia otra definición— "+
			"y ninguna de la entrada")

	_, _, leidos = esquemasDePrueba(t, verboDeArticulo())
	assert.NotContains(t, leidos[0], claveDeDefiniciones,
		"una entrada que no referencia nada no lleva $defs")

	for nombre, def := range verbosDeLosEsquemas() {
		_, _, leidos := esquemasDePrueba(t, def)

		for _, esquema := range leidos {
			assert.Equal(t, definicionesAlcanzadas(t, esquema), claves(definicionesDe(t, esquema)),
				"%s: cada $ref se resuelve en su documento y ninguna definición sobra", nombre)
		}
	}
}

func probarEsquemasCompactos(t *testing.T) {
	t.Helper()

	for nombre, def := range verbosDeLosEsquemas() {
		entrada, salida, leidos := esquemasDePrueba(t, def)

		for i, escrito := range [][]byte{entrada, salida} {
			var compacto bytes.Buffer
			require.NoError(t, json.Compact(&compacto, escrito))
			assert.Equal(t, compacto.String(), string(escrito), "%s: sin espacios ni saltos", nombre)

			assert.NotContains(t, leidos[i], "$schema", "%s: el borrador lo fija quien lo anuncia", nombre)
			assert.NotContains(t, leidos[i], "$id",
				"%s: un $id rebasaría las referencias internas contra otra base", nombre)
		}
	}
}

func probarEsquemasSueltos(t *testing.T) {
	t.Helper()

	_, _, leidos := esquemasDePrueba(t, verboDePrueba())

	entrada := compilarGenerado(t, leidos[0], "")
	require.NoError(t, entrada.Validate(map[string]any{"texto": "hola"}))
	require.Error(t, entrada.Validate(map[string]any{}), "el argumento de posición es obligatorio")
	require.Error(t, entrada.Validate(map[string]any{"texto": "hola", "offline": true}),
		"el nombre de una bandera global es una propiedad de más (FR-020)")

	// La salida se compila sola, sin el documento de --describe alrededor: sus
	// referencias resuelven contra sus propios $defs, y lo que rechaza lo rechaza
	// por ellos.
	salida := compilarGenerado(t, leidos[1], "")
	require.NoError(t, salida.Validate(sobreEmitido(t, resultadoDePrueba(), nil)))

	for _, caso := range casosDeFallo(t) {
		require.NoError(t, salida.Validate(
			sobreEmitido(t, schema.Resultado{Procedencia: caso.delApplet}, caso.err)), caso.nombre)
	}

	require.Error(t, salida.Validate(sobreDeFalloConTraza(t)),
		"con ok falso, `data` es exactamente {clase, mensaje}")

	for nombre, def := range verbosDeLosEsquemas() {
		_, _, leidos := esquemasDePrueba(t, def)

		for _, esquema := range leidos {
			require.NotNil(t, compilarGenerado(t, esquema, ""), nombre)
		}
	}
}

func probarEsquemasImposibles(t *testing.T) {
	t.Helper()

	sinStruct := verboDePrueba()
	sinStruct.Argumentos = "no soy un struct"

	conHomonimos := verboDePrueba()
	conHomonimos.Salida = salidaConHomonimos()

	for nombre, def := range map[string]Verbo{
		"unos argumentos que no son un struct":    sinStruct,
		"dos tipos distintos con el mismo nombre": conHomonimos,
	} {
		entrada, salida, err := EsquemasDeHerramienta(def)

		require.ErrorIs(t, err, errEsquemaImposible, nombre)
		assert.Equal(t, 1, CodigoSalida(err), "%s: es un defecto de quien declaró el verbo", nombre)
		assert.Nil(t, entrada, "%s: no se entrega un esquema a medias", nombre)
		assert.Nil(t, salida, "%s: no se entrega un esquema a medias", nombre)
	}
}

// llamadaAnalizada es lo que el analizador de la orden entiende de la línea de
// una llamada.
type llamadaAnalizada struct {
	verbos   *verbosDeLlamada
	analisis Analisis
	err      error
}

// analizarLlamada convierte los argumentos de una llamada y entrega la línea al
// analizador como lo hace el kernel: detrás del verbo y de las banderas del
// servidor (contracts/servidor-mcp.md §3, paso 2).
func analizarLlamada(t *testing.T, def Verbo, argumentos string) llamadaAnalizada {
	t.Helper()

	linea, err := LineaDeLlamada(def, []byte(argumentos))
	require.NoError(t, err)

	llamada := llamadaAnalizada{verbos: &verbosDeLlamada{}}
	llamada.analisis, llamada.err = Analizar(&presentadorDoble{}, def.Applet, llamada.verbos,
		append([]string{def.Verbo, "--json"}, linea...))

	return llamada
}

// TestLineaDeLlamada comprueba la conversión de los argumentos de una llamada en
// la línea de órdenes de su verbo: los posicionales en el orden de sus campos
// detrás de `--`, lo que no va por su posición como bandera delante, y los cuatro
// rechazos de contracts/servidor-mcp.md §3 con su mensaje, carácter a carácter,
// y la clase `argumentos` (FR-010, FR-011, FR-020; research D5, V24).
func TestLineaDeLlamada(t *testing.T) {
	t.Parallel()

	for _, caso := range lineasEsperadas() {
		t.Run("la línea: "+caso.nombre, func(t *testing.T) {
			t.Parallel()

			linea, err := LineaDeLlamada(caso.def, []byte(caso.argumentos))

			require.NoError(t, err)
			assert.Equal(t, caso.linea, linea)
		})
	}

	for _, caso := range llamadasRechazadas() {
		t.Run("el rechazo: "+caso.nombre, func(t *testing.T) {
			t.Parallel()

			// Varias veces: el mensaje es parte del sobre, y el mismo rechazo tiene
			// que dar el mismo sobre aunque el objeto se recorra en otro orden.
			for range 32 {
				linea, err := LineaDeLlamada(caso.def, []byte(caso.argumentos))

				require.ErrorIs(t, err, ErrArgumentos)
				require.Equal(t, argumentosInvalidos+caso.mensaje, err.Error())
				assert.Equal(t, schema.ClaseArgumentos, Clasificar(err))
				assert.Nil(t, linea, "un rechazo no deja ninguna línea que ejecutar")
			}
		})
	}

	subtests := []struct {
		nombre string
		probar func(*testing.T)
	}{
		{"el analizador de la orden lee cada tipo de argumento", probarLineaAnteElAnalizador},
		{"lo que va detrás del terminador no es ninguna bandera", probarLineaTrasElTerminador},
		{"que falte un obligatorio lo dice el analizador de la orden", probarObligatorioAusente},
		{"un campo que una llamada no sabe escribir es un defecto del verbo", probarCampoImposible},
	}

	for _, subtest := range subtests {
		t.Run(subtest.nombre, func(t *testing.T) {
			t.Parallel()

			subtest.probar(t)
		})
	}
}

// lineaEsperada es una llamada que vale, con la línea que le corresponde.
type lineaEsperada struct {
	nombre     string
	def        Verbo
	argumentos string
	linea      []string
}

func lineasEsperadas() []lineaEsperada {
	return []lineaEsperada{
		{
			"sin arguments cuenta como un objeto vacío",
			verboDeComprobacion(), ``,
			[]string{"--"},
		},
		{
			"un objeto vacío",
			verboDeComprobacion(), `{}`,
			[]string{"--"},
		},
		{
			"un verbo sin argumentos",
			verboSinArgumentos(), `{}`,
			[]string{"--"},
		},
		{
			"una cadena por su posición",
			verboDeComprobacion(), `{"norma":"BOE-A-2015-10565"}`,
			[]string{"--", "BOE-A-2015-10565"},
		},
		{
			"los posicionales van en el orden de sus campos y no en el del objeto",
			verboDeArticulo(), `{"bloque":"a21","norma":"BOE-A-2015-10565"}`,
			[]string{"--", "BOE-A-2015-10565", "a21"},
		},
		{
			"una lista de cadenas, un argumento por elemento",
			verboDeComprobacion(), `{"bloques":["a21","da-3"],"norma":"BOE-A-2015-10565"}`,
			[]string{"--", "BOE-A-2015-10565", "a21", "da-3"},
		},
		{
			"una lista vacía no escribe nada",
			verboDeComprobacion(), `{"norma":"BOE-A-2015-10565","bloques":[]}`,
			[]string{"--", "BOE-A-2015-10565"},
		},
		{
			"una cadena vacía es un argumento",
			verboDeComprobacion(), `{"norma":""}`,
			[]string{"--", ""},
		},
		{
			"lo que parece una bandera va detrás del terminador, tal cual",
			verboDeArticulo(), `{"norma":"--offline","bloque":"-h"}`,
			[]string{"--", "--offline", "-h"},
		},
		{
			"falta un obligatorio: la línea se devuelve y lo dirá el analizador",
			verboDeArticulo(), `{"norma":"BOE-A-2015-10565"}`,
			[]string{"--", "BOE-A-2015-10565"},
		},
		{
			"un entero por su posición",
			verboVariado(), `{"veces":3,"norma":"N"}`,
			[]string{"--", "N", "3"},
		},
		{
			"un entero negativo",
			verboVariado(), `{"veces":-3,"norma":"N"}`,
			[]string{"--", "N", "-3"},
		},
		{
			"un entero escrito con decimales, que el esquema admite como entero",
			verboVariado(), `{"veces":3.0,"norma":"N"}`,
			[]string{"--", "N", "3"},
		},
		{
			"un entero escrito con exponente",
			verboVariado(), `{"veces":2e1,"norma":"N"}`,
			[]string{"--", "N", "20"},
		},
		{
			"un entero que no cabe en un número de coma flotante conserva sus cifras",
			verboVariado(), `{"veces":9007199254740993,"norma":"N"}`,
			[]string{"--", "N", "9007199254740993"},
		},
		{
			"lo que no va por su posición va delante, como bandera con su valor",
			verboVariado(),
			`{"bloques":["a1","a2"],"materias":["x","y"],"limite":-3,"forzar":false,` +
				`"veces":2,"norma":"N","etiqueta":"una etiqueta"}`,
			[]string{
				"--etiqueta=una etiqueta", "--forzar=false", "--limite=-3",
				"--materias=x", "--materias=y", "--", "N", "2", "a1", "a2",
			},
		},
		{
			"una bandera booleana verdadera",
			verboVariado(), `{"forzar":true,"norma":"N"}`,
			[]string{"--forzar=true", "--", "N"},
		},
	}
}

// llamadaRechazada es una llamada que no vale, con lo que dice su mensaje detrás
// de «argumentos inválidos: ».
type llamadaRechazada struct {
	nombre     string
	def        Verbo
	argumentos string
	mensaje    string
}

func llamadasRechazadas() []llamadaRechazada {
	const (
		noEsUnObjeto = "los argumentos de " + herramientaArticulo + " no son un objeto JSON"
		entero       = `el argumento "veces" de echo_variar tiene que ser un n` + "\xc3\xba" + `mero entero`
		listaCadenas = `el argumento "bloques" de ` + herramientaComprobar + ` tiene que ser una lista de cadenas`
	)

	return []llamadaRechazada{
		{"una lista no es un objeto", verboDeArticulo(), `["BOE-A-2015-10565","a21"]`, noEsUnObjeto},
		{"una cadena no es un objeto", verboDeArticulo(), `"a21"`, noEsUnObjeto},
		{"un número no es un objeto", verboDeArticulo(), `21`, noEsUnObjeto},
		{"null no es un objeto", verboDeArticulo(), `null`, noEsUnObjeto},
		{"lo que no es JSON no es un objeto", verboDeArticulo(), `{"norma":`, noEsUnObjeto},
		{"dos documentos no son un objeto", verboDeArticulo(), `{} {}`, noEsUnObjeto},
		{
			"sobra una propiedad",
			verboDeArticulo(), `{"norma":"N","bloque":"a21","rango":"ley"}`,
			herramientaArticulo + ` no tiene el argumento "rango"`,
		},
		{
			"el nombre de una bandera global es una propiedad que sobra",
			verboDeArticulo(), `{"norma":"N","bloque":"a21","offline":true}`,
			herramientaArticulo + ` no tiene el argumento "offline"`,
		},
		{
			"de varias que sobran se nombra siempre la misma",
			verboDeArticulo(), `{"zeta":1,"norma":"N","timeout":"1s","alfa":2,"json":true}`,
			herramientaArticulo + ` no tiene el argumento "alfa"`,
		},
		{
			"un número donde va una cadena",
			verboDeArticulo(), `{"norma":"N","bloque":21}`,
			`el argumento "bloque" de ` + herramientaArticulo + ` tiene que ser una cadena`,
		},
		{
			"null donde va una cadena",
			verboDeComprobacion(), `{"norma":null}`,
			`el argumento "norma" de ` + herramientaComprobar + ` tiene que ser una cadena`,
		},
		{
			"una cadena donde va una lista de cadenas",
			verboDeComprobacion(), `{"norma":"N","bloques":"a21"}`, listaCadenas,
		},
		{
			"una lista con un elemento que no es una cadena",
			verboDeComprobacion(), `{"norma":"N","bloques":["a21",3]}`, listaCadenas,
		},
		{
			"una cadena donde va un booleano",
			verboVariado(), `{"norma":"N","forzar":"true"}`,
			`el argumento "forzar" de echo_variar tiene que ser un booleano`,
		},
		{"una cadena donde va un entero", verboVariado(), `{"norma":"N","veces":"3"}`, entero},
		{"un número con decimales donde va un entero", verboVariado(), `{"norma":"N","veces":2.5}`, entero},
		{
			"un argumento de posición sin el que le precede",
			verboDeComprobacion(), `{"bloques":["a21"]}`,
			`el argumento "bloques" de ` + herramientaComprobar + ` no se puede dar sin "norma"`,
		},
	}
}

func probarLineaAnteElAnalizador(t *testing.T) {
	t.Helper()

	llamada := analizarLlamada(t, verboVariado(),
		`{"bloques":["a1","a2"],"materias":["x","y"],"limite":-3,"forzar":true,`+
			`"veces":2,"norma":"N","etiqueta":"una etiqueta"}`)

	require.NoError(t, llamada.err)
	assert.Equal(t, "variar", llamada.analisis.Verbo)
	assert.Equal(t, argumentosVariados{
		Etiqueta: "una etiqueta",
		Forzar:   true,
		Limite:   -3,
		Materias: []string{"x", "y"},
		Norma:    "N",
		Veces:    2,
		Bloques:  []string{"a1", "a2"},
	}, llamada.verbos.Variar)

	llamada = analizarLlamada(t, verboVariado(), `{"forzar":false,"norma":"N","veces":-3}`)

	require.NoError(t, llamada.err)
	assert.Equal(t, argumentosVariados{Limite: 10, Norma: "N", Veces: -3}, llamada.verbos.Variar,
		"un booleano falso y un entero negativo se leen como se dieron, y lo que no "+
			"se da conserva su valor por omisión")

	llamada = analizarLlamada(t, verboDeComprobacion(), ``)

	require.NoError(t, llamada.err, "el terminador sin nada detrás es una línea que vale")
	assert.Equal(t, argumentosDeComprobacion{}, llamada.verbos.Check)
}

func probarLineaTrasElTerminador(t *testing.T) {
	t.Helper()

	llamada := analizarLlamada(t, verboDeComprobacion(),
		`{"norma":"--offline","bloques":["--help","--no-graph"]}`)

	require.NoError(t, llamada.err)
	require.NotNil(t, llamada.verbos.Check.Norma)
	assert.Equal(t, "--offline", *llamada.verbos.Check.Norma)
	assert.Equal(t, []string{"--help", "--no-graph"}, llamada.verbos.Check.Bloques)

	assert.Equal(t, DecisionEjecutar, llamada.analisis.Decision,
		"ningún argumento de una llamada pide la ayuda")
	assert.False(t, llamada.analisis.Globales.Offline,
		"ningún argumento de una llamada es una bandera global (FR-020)")
	assert.False(t, llamada.analisis.Globales.SinGrafo)
}

func probarObligatorioAusente(t *testing.T) {
	t.Helper()

	llamada := analizarLlamada(t, verboDeArticulo(), `{"norma":"BOE-A-2015-10565"}`)

	require.ErrorIs(t, llamada.err, ErrArgumentos)
	assert.Equal(t, argumentosInvalidos+`expected "<bloque>"`, llamada.err.Error(),
		"el mensaje es el de la orden (contracts/servidor-mcp.md §3)")
}

func probarCampoImposible(t *testing.T) {
	t.Helper()

	linea, err := LineaDeLlamada(verboConFiltro(), []byte(`{"filtro":{"desde":"2015"}}`))

	require.ErrorIs(t, err, errEsquemaImposible)
	assert.Equal(t, 1, CodigoSalida(err), "no es un error de quien llama")
	assert.Nil(t, linea)

	sinStruct := verboDePrueba()
	sinStruct.Argumentos = "no soy un struct"

	linea, err = LineaDeLlamada(sinStruct, []byte(`{}`))

	require.ErrorIs(t, err, errEsquemaImposible)
	assert.Nil(t, linea)
}
