package app_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cli"
)

// Las ayudas que declaran hoy los argumentos y banderas propias de los verbos del
// registro de producción: en todos los verbos y en los que el servidor anuncia
// como herramienta (SC-001 de H39; research.md V2 de H39). Escritas aquí, y no
// contadas en el código que se prueba, para que el control no pase en vacío.
//
// Son etiquetas `help` declaradas, y el recorrido cuenta una por argumento de
// cada verbo: hoy, 36 y 28, porque boe indice, metadatos y analisis, cita
// preparar y cotejar, y skills list y doctor comparten sus argumentos. Cada
// declaración llega al menos a un verbo, así que una cuenta por debajo de estas
// cifras es que el recorrido ha dejado de comprobar alguna (gates/supuestos.md
// de H39, T003).
const (
	ayudasDeHoy               = 28
	ayudasDeHoyConHerramienta = 22
)

// argumentoDelVerbo es un argumento o una bandera propia de un verbo tal como
// lo ve el modelo que Kong construye con la gramática del applet: el nombre con
// el que se escribe y su ayuda, la que imprime --help.
type argumentoDelVerbo struct {
	nombre, ayuda string
}

// verboConAyudas es un verbo del registro con sus argumentos y banderas propias,
// en el orden del modelo de Kong, y cuántos argumentos de posición exige.
type verboConAyudas struct {
	applet, verbo string
	argumentos    []argumentoDelVerbo
	obligatorios  int
}

func (v verboConAyudas) titulo() string {
	return v.applet + " " + v.verbo
}

func (v verboConAyudas) herramienta() string {
	return v.applet + "_" + v.verbo
}

// ayudas cuenta los argumentos del verbo que declaran una ayuda.
func (v verboConAyudas) ayudas() int {
	cuantas := 0

	for _, argumento := range v.argumentos {
		if argumento.ayuda != "" {
			cuantas++
		}
	}

	return cuantas
}

// entradaDescrita es lo que este control lee de un esquema de entrada: la
// `description` de cada propiedad, nula si no la lleva.
type entradaDescrita struct {
	Propiedades map[string]struct {
		Descripcion *string `json:"description"`
	} `json:"properties"`
}

// TestAyudasDeLosArgumentos es el control de SC-001 de H39 (FR-001, FR-002,
// FR-003; contracts/esquemas.md §5 de H39; research.md D2 de H39): pasa por los
// verbos del registro de producción y, de cada uno, compara la ayuda de cada
// argumento y bandera propios del modelo que Kong construye con la gramática del
// applet —la que imprime --help— con la `description` de su propiedad en la
// entrada del documento de --describe y, en los verbos con herramienta, en el
// `inputSchema` que anuncia el servidor arrancado en proceso. Falla con el
// primer argumento cuya `description` falta o difiere, con una propiedad sin
// ayuda que la lleva y con una de las ocho banderas globales que la lleva, que
// tienen ayuda y no son del verbo; y si cuenta menos ayudas de las que hay hoy.
//
// El oráculo no es el código que se prueba: la ayuda sale del modelo de Kong y no
// de leer las etiquetas, que repetiría la regla de nombres de internal/cli
// (research.md D2 y V1 de H39). Los subtests demuestran que la comparación no
// pasa en vacío con cada forma de defecto, también con las que el registro de hoy
// no puede dar: ningún argumento de sus verbos está sin ayuda.
func TestAyudasDeLosArgumentos(t *testing.T) {
	t.Parallel()

	for _, caso := range defectosDePrueba() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			var entrada entradaDescrita

			require.NoError(t, json.Unmarshal([]byte(caso.entrada), &entrada))

			err := primerDefecto(argumentosDePrueba, entrada, []string{"json"})
			if caso.defecto == "" {
				assert.NoError(t, err)

				return
			}

			assert.ErrorContains(t, err, caso.defecto)
		})
	}

	registro, err := app.RegistroDeProduccion("")
	require.NoError(t, err)

	anunciadas := entradasAnunciadas(t, registro)
	globales := nombresDeLasGlobales()

	var ayudas, conHerramienta int

	for _, verbo := range verbosConAyudas(t, registro) {
		require.NoError(t, primerDefecto(verbo.argumentos, entradaDeDescribe(t, registro, verbo), globales),
			"la entrada de %s --describe lleva como description la ayuda de cada argumento del verbo, y nada más "+
				"(FR-001, FR-003)", verbo.titulo())

		ayudas += verbo.ayudas()

		if slices.Contains(appletsSinHerramientas, verbo.applet) {
			continue
		}

		anunciada, anuncia := anunciadas[verbo.herramienta()]
		require.True(t, anuncia, "el servidor anuncia la herramienta %s", verbo.herramienta())
		require.NoError(t, primerDefecto(verbo.argumentos, anunciada, nil),
			"el inputSchema de %s en tools/list lleva la misma description que --describe (FR-002, FR-003)",
			verbo.herramienta())

		conHerramienta += verbo.ayudas()
	}

	assert.GreaterOrEqual(t, ayudas, ayudasDeHoy,
		"los argumentos con ayuda de todos los verbos del registro, que el control ha comprobado (SC-001)")
	assert.GreaterOrEqual(t, conHerramienta, ayudasDeHoyConHerramienta,
		"los argumentos con ayuda de los verbos con herramienta, que el control ha comprobado (SC-001)")
}

// argumentosDePrueba son los de un verbo de prueba: uno con una ayuda que lleva
// `<`, `>` y `$`, que se comparan sin escapar ni interpolar, y uno sin ayuda.
var argumentosDePrueba = []argumentoDelVerbo{
	{nombre: "con", ayuda: "Ayuda de <con> en $HOME."},
	{nombre: "sin"},
}

// defectoDePrueba es una entrada del verbo de prueba, con una bandera global,
// y el defecto que la comparación tiene que nombrar en ella, vacío si no hay.
type defectoDePrueba struct {
	nombre, entrada, defecto string
}

// defectosDePrueba son la entrada que cumple y una por cada forma de defecto.
func defectosDePrueba() []defectoDePrueba {
	return []defectoDePrueba{
		{
			nombre:  "cumple",
			entrada: `{"properties":{"con":{"type":"string","description":"Ayuda de <con> en $HOME."},"sin":{},"json":{}}}`,
		},
		{
			nombre:  "falta-la-propiedad",
			entrada: `{"properties":{"sin":{},"json":{}}}`,
			defecto: "el argumento con no tiene propiedad",
		},
		{
			nombre:  "falta-la-description",
			entrada: `{"properties":{"con":{"type":"string"},"sin":{},"json":{}}}`,
			defecto: "a la propiedad de con le falta la description",
		},
		{
			nombre:  "otra-description",
			entrada: `{"properties":{"con":{"description":"Ayuda de con en $HOME."},"sin":{},"json":{}}}`,
			defecto: "la description de con es",
		},
		{
			nombre:  "sin-ayuda-con-description",
			entrada: `{"properties":{"con":{"description":"Ayuda de <con> en $HOME."},"sin":{"description":""},"json":{}}}`,
			defecto: "el argumento sin no tiene ayuda y su propiedad lleva la description",
		},
		{
			nombre:  "falta-la-global",
			entrada: `{"properties":{"con":{"description":"Ayuda de <con> en $HOME."},"sin":{}}}`,
			defecto: "falta la propiedad de la bandera global --json",
		},
		{
			nombre:  "global-con-description",
			entrada: `{"properties":{"con":{"description":"Ayuda de <con> en $HOME."},"sin":{},"json":{"description":"x"}}}`,
			defecto: "la bandera global --json, que no es del verbo, lleva la description",
		},
		{
			nombre:  "propiedad-de-ningun-argumento-con-description",
			entrada: `{"properties":{"con":{"description":"Ayuda de <con> en $HOME."},"sin":{},"json":{},"otra":{"description":"x"}}}`,
			defecto: "la propiedad otra, que no es de ningún argumento del verbo, lleva la description",
		},
	}
}

// verbosConAyudas son los verbos del registro, applet a applet y en el orden de
// su catálogo, tal como los ve el modelo que Kong construye con la gramática de
// cada applet: las ocho banderas globales en la raíz y un mandato por verbo con
// los argumentos que da su fábrica. De cada mandato toma solo lo suyo: sus
// argumentos de posición y sus banderas, no las globales de la raíz.
func verbosConAyudas(t *testing.T, registro *app.Registro) []verboConAyudas {
	t.Helper()

	var verbos []verboConAyudas

	for _, nombre := range registro.Nombres() {
		applet, registrado := registro.Buscar(nombre)
		require.True(t, registrado, nombre)

		opciones := []kong.Option{kong.Name(nombre)}
		for _, verbo := range applet.Verbos() {
			opciones = append(opciones, kong.DynamicCommand(verbo.Nombre, verbo.Descripcion, "", verbo.Argumentos()))
		}

		analizador, err := kong.New(&cli.Globales{}, opciones...)
		require.NoError(t, err, "la gramática del applet %s", nombre)
		require.Len(t, analizador.Model.Children, len(applet.Verbos()),
			"el modelo de %s tiene un mandato por verbo de su catálogo", nombre)

		for _, mandato := range analizador.Model.Children {
			verbos = append(verbos, verboDelModelo(nombre, mandato))
		}
	}

	return verbos
}

// verboDelModelo lee de un mandato del modelo de Kong sus argumentos de posición
// y sus banderas, con su nombre y su ayuda, y cuántos de posición exige.
func verboDelModelo(applet string, mandato *kong.Node) verboConAyudas {
	verbo := verboConAyudas{applet: applet, verbo: mandato.Name}

	for _, dePosicion := range mandato.Positional {
		verbo.argumentos = append(verbo.argumentos, argumentoDelVerbo{nombre: dePosicion.Name, ayuda: dePosicion.Help})

		if dePosicion.Required {
			verbo.obligatorios++
		}
	}

	for _, bandera := range mandato.Flags {
		verbo.argumentos = append(verbo.argumentos, argumentoDelVerbo{nombre: bandera.Name, ayuda: bandera.Help})
	}

	return verbo
}

// entradaDeDescribe es la entrada del documento que el verbo emite con
// --describe, invocado en proceso. El análisis va antes que la descripción, así
// que la orden lleva un valor cualquiera por cada argumento de posición
// obligatorio, que --describe no lee (research.md V5 de H39).
func entradaDeDescribe(t *testing.T, registro *app.Registro, verbo verboConAyudas) entradaDescrita {
	t.Helper()

	orden := []string{verbo.applet, verbo.verbo}
	for range verbo.obligatorios {
		orden = append(orden, "x")
	}

	orden = append(orden, "--describe")

	describe := invocarCon(t, registro, orden...)
	require.Equal(t, 0, describe.codigo, "%s: %s", strings.Join(orden, " "), describe.errores)

	var documento struct {
		Titulo string `json:"title"`
		Partes struct {
			Entrada entradaDescrita `json:"entrada"`
		} `json:"properties"`
	}

	require.NoError(t, json.Unmarshal([]byte(describe.salida), &documento), verbo.titulo())
	require.Equal(t, verbo.titulo(), documento.Titulo, "el documento de --describe es el de este verbo")

	return documento.Partes.Entrada
}

// entradasAnunciadas son los `inputSchema` de las herramientas que anuncia
// `kitlegal mcp serve` sobre los applets del registro de producción, arrancado en
// proceso y cerrado antes de volver, por el nombre de cada herramienta.
func entradasAnunciadas(t *testing.T, distribuido *app.Registro) map[string]entradaDescrita {
	t.Helper()

	servidor := arrancarElServidor(t, registroParaServir(t, distribuido, false), false)

	anunciadas, err := servidor.sesion.Herramientas(t.Context())
	require.NoError(t, err)

	entradas := make(map[string]entradaDescrita, len(anunciadas))

	for _, anunciada := range anunciadas {
		var entrada entradaDescrita

		require.NoError(t, json.Unmarshal(anunciada.Entrada, &entrada), anunciada.Nombre)
		entradas[anunciada.Nombre] = entrada
	}

	require.Equal(t, 0, servidor.cerrar(t), servidor.errores.lineas())

	return entradas
}

// nombresDeLasGlobales son las propiedades de las ocho banderas globales en la
// entrada de --describe, de la lista que este paquete escribe a mano.
func nombresDeLasGlobales() []string {
	nombres := make([]string, 0, len(banderasGlobales))

	for _, global := range banderasGlobales {
		nombres = append(nombres, strings.TrimPrefix(global.nombre, "--"))
	}

	return nombres
}

// primerDefecto compara una entrada con los argumentos del verbo y devuelve el
// primero que no cumple: un argumento, en el orden del modelo de Kong, cuya
// propiedad falta, no lleva su ayuda como description o la lleva sin tener
// ayuda; después, una bandera global que falta o lleva description; y, por
// último, cualquier otra propiedad que la lleve sin ser de ningún argumento del
// verbo. Las globales son las que la entrada tiene que llevar sin description:
// ninguna en el esquema de una herramienta, que no las lleva.
func primerDefecto(argumentos []argumentoDelVerbo, entrada entradaDescrita, globales []string) error {
	revisadas := make(map[string]bool, len(argumentos)+len(globales))

	for _, argumento := range argumentos {
		revisadas[argumento.nombre] = true

		if err := defectoDelArgumento(argumento, entrada); err != nil {
			return err
		}
	}

	for _, global := range globales {
		revisadas[global] = true

		propiedad, existe := entrada.Propiedades[global]
		if !existe {
			return fmt.Errorf("falta la propiedad de la bandera global --%s", global)
		}

		if propiedad.Descripcion != nil {
			return fmt.Errorf("la bandera global --%s, que no es del verbo, lleva la description %q",
				global, *propiedad.Descripcion)
		}
	}

	for _, nombre := range slices.Sorted(maps.Keys(entrada.Propiedades)) {
		if descripcion := entrada.Propiedades[nombre].Descripcion; !revisadas[nombre] && descripcion != nil {
			return fmt.Errorf("la propiedad %s, que no es de ningún argumento del verbo, lleva la description %q",
				nombre, *descripcion)
		}
	}

	return nil
}

// defectoDelArgumento dice qué no cumple la propiedad de un argumento: que
// falte, que el argumento tenga ayuda y su description falte o no sea esa ayuda
// carácter a carácter, o que no la tenga y su propiedad lleve una.
func defectoDelArgumento(argumento argumentoDelVerbo, entrada entradaDescrita) error {
	propiedad, existe := entrada.Propiedades[argumento.nombre]

	switch {
	case !existe:
		return fmt.Errorf("el argumento %s no tiene propiedad", argumento.nombre)
	case argumento.ayuda == "" && propiedad.Descripcion != nil:
		return fmt.Errorf("el argumento %s no tiene ayuda y su propiedad lleva la description %q",
			argumento.nombre, *propiedad.Descripcion)
	case argumento.ayuda == "":
		return nil
	case propiedad.Descripcion == nil:
		return fmt.Errorf("a la propiedad de %s le falta la description %q", argumento.nombre, argumento.ayuda)
	case *propiedad.Descripcion != argumento.ayuda:
		return fmt.Errorf("la description de %s es %q y su ayuda, %q", argumento.nombre,
			*propiedad.Descripcion, argumento.ayuda)
	default:
		return nil
	}
}
