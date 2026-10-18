package app

import (
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/plazos"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/core/territorio"
)

// actualizarGoldenDePlazos es la bandera con la que TestAppletPlazos escribe,
// antes de compararlos, los golden de testdata/plazos/ desde lo que da el
// verbo: nunca los escribe una mano (contracts/applet-plazos.md §7 de H9). Solo
// la usa la tarea [datos] que los crea o los cambia, y lo escrito lo revisa una
// persona; make ci nunca la pasa.
var actualizarGoldenDePlazos = flag.Bool("actualizar-golden-de-plazos", false,
	"escribe en testdata/plazos/ los golden de plazos calcular desde su salida antes de compararlos")

// Lo que estos tests esperan del sobre de plazos, escrito aquí otra vez y no
// tomado del código que se prueba (contracts/applet-plazos.md §3 de H9).
const (
	// fuenteDePlazosEsperada es la del espacio reservado a lo que el binario
	// calcula: dice que no se ha consultado ninguna fuente.
	fuenteDePlazosEsperada = "kitlegal.plazos"
	// urlDePlazosEsperada es la de todo sobre del applet, de éxito o de fallo.
	urlDePlazosEsperada = "kitlegal:applet/plazos"
)

// Los golden (contracts/applet-plazos.md §7 de H9).
const (
	// carpetaDeLosGoldenDePlazos es la suya, relativa al directorio de este
	// paquete.
	carpetaDeLosGoldenDePlazos = "testdata/plazos"
	// goldenDelDescribeDePlazos es el de --describe de calcular.
	goldenDelDescribeDePlazos = "describe.json"
)

// procedenciaDeUnFalloDePlazos es la firma de todo fallo que decide el applet.
var procedenciaDeUnFalloDePlazos = schema.Procedencia{Fuente: fuenteDePlazosEsperada, URL: urlDePlazosEsperada}

// goldenDePlazos es un caso de la tabla de contracts/casos-de-plazos.md de H9
// con su golden: el fichero y la invocación que lo da, con los municipios por su
// nombre, como los pide la skill.
type goldenDePlazos struct {
	fichero string
	orden   []string
}

// goldenDePlazosDelContrato son los de los casos (a), (b) y (e): Leganés, con
// festivos locales; Tordesillas, sin ellos y con su aviso; y el contencioso
// que pasa por agosto (H9 FR-085; contracts/casos-de-plazos.md, casos 1 a 3).
func goldenDePlazosDelContrato() []goldenDePlazos {
	return []goldenDePlazos{
		{fichero: "caso-a.json", orden: []string{"reposicion", "2026-09-09", "Leganés", "Leganés"}},
		{fichero: "caso-b.json", orden: []string{"reposicion", "2026-09-09", "Tordesillas", "Tordesillas"}},
		{fichero: "caso-e.json", orden: []string{"contencioso", "2026-06-15", "Madrid"}},
	}
}

// verbosDelContratoDePlazos es el único verbo de plazos con la invocación con
// que se describe: el tipo, la fecha y un municipio son obligatorios para la
// gramática, que va antes que la descripción, así que sin ellos --describe
// termina en 2 (contracts/applet-plazos.md §1 de H9). Describir no comprueba ni
// resuelve nada, de modo que valen tres argumentos cualesquiera.
func verbosDelContratoDePlazos() []verboDelContrato {
	return []verboDelContrato{{nombre: "calcular", argumento: []string{"calcular", "x", "x", "x"}}}
}

// fuentesDePlazos son los ficheros del binario, como los recibe el applet en
// las dos raíces de composición. Cada llamada da unas nuevas, que se pueden
// alterar sin tocar las de otro test.
func fuentesDePlazos(t *testing.T) FicherosDePlazos {
	t.Helper()

	fuentes, err := FuentesDePlazos()
	require.NoError(t, err, "los ficheros de plazos del binario se leen")

	return fuentes
}

// registroDePlazos es el registro de estos tests: el applet plazos compuesto
// con esas fuentes. Es un registro local, sin almacén al que entregar.
func registroDePlazos(t *testing.T, fuentes FicherosDePlazos) *Registro {
	t.Helper()

	var registro Registro

	require.NoError(t, registro.Registrar(AppletPlazos(fuentes)))

	return &registro
}

// ordenDePlazos ejecuta calcular con esos argumentos y los que se añadan, con
// --json, por la raíz de composición entera.
func ordenDePlazos(t *testing.T, registro *Registro, argumentos ...string) invocacionDePrueba {
	t.Helper()

	return invocar(t, registro, slices.Concat([]string{"kitlegal", "plazos", "calcular"}, argumentos,
		[]string{"--json"})...)
}

// TestAppletPlazos fija lo que declara y lo que da el applet plazos (H9 FR-001,
// FR-007, FR-010, FR-012, FR-013, FR-085; contracts/applet-plazos.md §1 a §7):
//
//   - un solo verbo, calcular, sin verbo por omisión, con sus argumentos de
//     cadena —tres de posición y dos banderas— y el valor cero del plazo;
//   - los golden de los casos (a), (b) y (e), byte a byte salvo la fecha del
//     reloj del kernel, que validan contra schemas/plazo.json, con el mismo data
//     y el mismo hash con --offline;
//   - ninguna operación de grafo, ni del éxito ni de un fallo;
//   - el golden de --describe, con sus dos banderas propias;
//   - los enumerados de la salida publicada, iguales a los vocabularios del
//     dominio;
//   - los fallos de los municipios, en el orden de §6, con el mensaje de
//     territorio de cada uno;
//   - unas fuentes que no cargan, el defecto de composición, que no impide la
//     ayuda ni --describe, y un análisis de las fuentes por applet, en su
//     primera ejecución.
func TestAppletPlazos(t *testing.T) {
	t.Parallel()

	t.Run("declara-el-verbo-calcular", func(t *testing.T) {
		t.Parallel()

		compruebaDeclaracionDePlazos(t)
	})

	t.Run("sin-verbo", func(t *testing.T) {
		t.Parallel()

		res := invocar(t, registroDePlazos(t, fuentesDePlazos(t)), "kitlegal", "plazos", "--json")

		exigirSobreDeFallo(t, res, schema.ClaseArgumentos, 2, cli.ProcedenciaKernel())
		assert.Contains(t, res.errores, "verbos de plazos: calcular")
	})

	t.Run("los-golden-de-los-casos", func(t *testing.T) {
		t.Parallel()

		compruebaGoldenDePlazos(t)
	})

	t.Run("sin-operaciones-de-grafo", func(t *testing.T) {
		t.Parallel()

		compruebaPlazosSinGrafo(t)
	})

	t.Run("el-golden-de-describe", func(t *testing.T) {
		t.Parallel()

		compruebaDescribeDePlazos(t)
	})

	t.Run("los-enumerados-publicados", func(t *testing.T) {
		t.Parallel()

		compruebaEnumeradosDePlazos(t)
	})

	for _, caso := range casosDeLosMunicipios() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			caso.comprueba(t)
		})
	}

	for _, caso := range casosDeFuentesDePlazosQueNoCargan() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			compruebaDefectoDeComposicionDePlazos(t, caso)
		})
	}

	t.Run("analiza-sus-fuentes-una-sola-vez", func(t *testing.T) {
		t.Parallel()

		compruebaAnalisisUnicoDePlazos(t)
	})
}

// argumentoDeCalcular es un campo de los argumentos de calcular: su nombre en
// Go, si va por su posición y su tipo.
type argumentoDeCalcular struct {
	campo      string
	dePosicion bool
	tipo       reflect.Type
}

// compruebaDeclaracionDePlazos exige la declaración del applet: su nombre, su
// único verbo, sin verbo por omisión, y sus argumentos, todos de cadena y sin
// enumerado de la gramática, que decide el dominio (research D2 y V12 de H9),
// con una ayuda que nombra cada tipo y cada unidad.
func compruebaDeclaracionDePlazos(t *testing.T) {
	t.Helper()

	applet := AppletPlazos(fuentesDePlazos(t))

	assert.Equal(t, "plazos", applet.Nombre())
	assert.NotEmpty(t, applet.Descripcion(), "la ayuda del binario lo describe")

	require.Equal(t, []string{"calcular"}, nombresDeLosVerbos(applet), "calcular es el único verbo")

	verbo := applet.Verbos()[0]

	assert.NotEmpty(t, verbo.Descripcion, "la ayuda del applet y la herramienta lo describen")
	assert.False(t, verbo.PorOmision, "ningún verbo por omisión: nombrar calcular es obligatorio")
	assert.IsType(t, plazos.Plazo{}, verbo.Salida)
	assert.True(t, reflect.ValueOf(verbo.Salida).IsZero(), "Salida es el valor cero de su tipo")

	primero, segundo := verbo.Argumentos(), verbo.Argumentos()
	assert.NotSame(t, primero, segundo, "la fábrica da un valor nuevo en cada invocación")

	tipo := reflect.TypeOf(primero)
	require.Equal(t, reflect.Pointer, tipo.Kind())

	var (
		declarados []argumentoDeCalcular
		ayudas     = map[string]string{}
	)

	for _, campo := range reflect.VisibleFields(tipo.Elem()) {
		if !campo.IsExported() {
			continue
		}

		_, dePosicion := campo.Tag.Lookup("arg")
		_, conEnumerado := campo.Tag.Lookup("enum")
		assert.False(t, conEnumerado, "%s: el dominio decide qué vale, no la gramática", campo.Name)

		declarados = append(declarados, argumentoDeCalcular{campo: campo.Name, dePosicion: dePosicion, tipo: campo.Type})
		ayudas[campo.Name] = campo.Tag.Get("help")
	}

	texto, textos := reflect.TypeFor[string](), reflect.TypeFor[[]string]()

	assert.Equal(t, []argumentoDeCalcular{
		{campo: "Tipo", dePosicion: true, tipo: texto},
		{campo: "Inicio", dePosicion: true, tipo: texto},
		{campo: "Municipios", dePosicion: true, tipo: textos},
		{campo: "Duracion", tipo: texto},
		{campo: "Unidad", tipo: texto},
	}, declarados, "los argumentos del contrato, y ninguno más")

	for _, codigo := range plazos.CodigosDeTipo() {
		assert.Contains(t, ayudas["Tipo"], codigo, "la ayuda del tipo nombra cada uno")
	}

	for _, unidad := range plazos.Unidades() {
		assert.Contains(t, ayudas["Unidad"], string(unidad), "la ayuda de la unidad nombra cada una")
	}
}

// compruebaGoldenDePlazos ejecuta cada caso con golden y exige: código 0 y el
// sobre de §3, con la fecha del reloj del kernel; que valide contra su parte de
// schemas/plazo.json, que restringe data; que, con la fecha de los golden, sea
// el fichero byte a byte; y que con --offline dé el mismo data y el mismo hash.
// Con -actualizar-golden-de-plazos escribe antes cada fichero desde la salida.
// Los golden son los únicos de la carpeta, con el de --describe: uno de más
// sería un golden que nadie compara.
func compruebaGoldenDePlazos(t *testing.T) {
	t.Helper()

	publicado, id := ficheroPublicadoDelVerbo(t, "calcular")
	require.Equal(t, raizDeLosEsquemas+"plazo.json", id, "la parte de calcular está en schemas/plazo.json")

	esquema := salidaPublicada(t, publicado, id, "calcular")
	registro := registroDePlazos(t, fuentesDePlazos(t))

	ficheros := []string{filepath.Join(carpetaDeLosGoldenDePlazos, goldenDelDescribeDePlazos)}

	for _, golden := range goldenDePlazosDelContrato() {
		ficheros = append(ficheros, filepath.Join(carpetaDeLosGoldenDePlazos, golden.fichero))

		antes := time.Now()
		res := ordenDePlazos(t, registro, golden.orden...)
		despues := time.Now()

		require.Equal(t, 0, res.codigo, "%s: %s", golden.fichero, res.errores)
		assert.Empty(t, res.errores, "una salida correcta no escribe nada en la salida de error")

		sobre := sobreDelJSON(t, res.salida)
		assert.Equal(t, true, sobre["ok"])
		assert.Equal(t, fuenteDePlazosEsperada, sobre["fuente"], "la fuente dice que no se ha consultado ninguna")
		assert.Equal(t, urlDePlazosEsperada, sobre["url"])
		assert.WithinRange(t, fechaDelSobre(t, sobre), antes, despues, "el sobre lo fecha el reloj del kernel")

		exigirSalidaPublicada(t, esquema, res.salida)

		emitido := conLaFechaDeLosGolden(t, res.salida)
		ruta := filepath.Join(carpetaDeLosGoldenDePlazos, golden.fichero)

		if *actualizarGoldenDePlazos {
			require.NoError(t, os.MkdirAll(carpetaDeLosGoldenDePlazos, 0o750))
			require.NoError(t, os.WriteFile(ruta, []byte(emitido), 0o600))
		}

		guardado, err := os.ReadFile(filepath.Clean(ruta))
		require.NoError(t, err, "el golden lo escribe el verbo, con -actualizar-golden-de-plazos")
		assert.Equal(t, string(guardado), emitido, "%s: el sobre, byte a byte salvo fecha_consulta", golden.fichero)

		sinRed := sobreDelJSON(t, ordenDePlazos(t, registro, append(slices.Clone(golden.orden), "--offline")...).salida)
		assert.Equal(t, sobre["data"], sinRed["data"], "%s: --offline no cambia data", golden.fichero)
		assert.Equal(t, sobre["hash"], sinRed["hash"], "%s: --offline no cambia el hash", golden.fichero)
	}

	if !*actualizarGoldenDePlazos {
		enLaCarpeta, err := filepath.Glob(filepath.Join(carpetaDeLosGoldenDePlazos, "*"))
		require.NoError(t, err)
		assert.ElementsMatch(t, ficheros, enLaCarpeta, "los golden de %s", carpetaDeLosGoldenDePlazos)
	}
}

// compruebaPlazosSinGrafo ejecuta calcular sin el kernel, que no entrega nada
// de un fallo y no dejaría ver si el applet lo trae, y exige que ni el éxito ni
// un fallo lleven ninguna operación de grafo: plazos no observa identidades
// del mundo y lo que nombra sale de data/ (H9 FR-013; research D11).
func compruebaPlazosSinGrafo(t *testing.T) {
	t.Helper()

	verbo := AppletPlazos(fuentesDePlazos(t)).Verbos()[0]

	for _, peticion := range []struct {
		argumentos []string
		falla      bool
	}{
		{argumentos: []string{"reposicion", "2026-09-09", "Leganés", "Leganés"}},
		{argumentos: []string{"contencioso", "2026-06-15", "Madrid"}},
		{argumentos: []string{"reposicion", "2026-09-09", "Inexistente", "Leganés"}, falla: true},
	} {
		argumentos, esDeCalcular := verbo.Argumentos().(*argumentosDeCalcular)
		require.True(t, esDeCalcular, "la fábrica da los argumentos de calcular")

		argumentos.Tipo, argumentos.Inicio = peticion.argumentos[0], peticion.argumentos[1]
		argumentos.Municipios = peticion.argumentos[2:]

		res, err := argumentos.Ejecutar(t.Context(), schema.Contexto{}, slog.New(slog.DiscardHandler))

		if peticion.falla {
			require.Error(t, err, "%q", peticion.argumentos)
		} else {
			require.NoError(t, err, "%q", peticion.argumentos)
			assert.IsType(t, plazos.Plazo{}, res.Datos)
		}

		assert.Zero(t, res.Grafo, "%q: ninguna operación de grafo", peticion.argumentos)
	}
}

// compruebaDescribeDePlazos exige que --describe de calcular sea su golden,
// byte a byte, y que su entrada declare como banderas propias la duración y la
// unidad, en ese orden, que es de donde la herramienta saca las suyas (H9
// FR-085; research V13). Con -actualizar-golden-de-plazos escribe antes el
// fichero desde la salida.
func compruebaDescribeDePlazos(t *testing.T) {
	t.Helper()

	invocacion := slices.Concat([]string{"kitlegal", "plazos"}, argumentoDelContrato(t, "plazos", "calcular"),
		[]string{"--describe"})

	res := invocar(t, registroDePlazos(t, fuentesDePlazos(t)), invocacion...)
	require.Equal(t, 0, res.codigo, res.errores)

	ruta := filepath.Join(carpetaDeLosGoldenDePlazos, goldenDelDescribeDePlazos)

	if *actualizarGoldenDePlazos {
		require.NoError(t, os.MkdirAll(carpetaDeLosGoldenDePlazos, 0o750))
		require.NoError(t, os.WriteFile(ruta, []byte(res.salida), 0o600))
	}

	guardado, err := os.ReadFile(filepath.Clean(ruta))
	require.NoError(t, err, "el golden lo escribe --describe, con -actualizar-golden-de-plazos")
	assert.Equal(t, string(guardado), res.salida, "--describe de calcular, byte a byte")

	documento, err := objetoJSON([]byte(res.salida))
	require.NoError(t, err, "--describe emite un único objeto JSON")

	assert.Equal(t, []any{"duracion", "unidad"}, valorDelEsquema(t, documento, "properties", "entrada", "x-banderas"),
		"las dos banderas propias de calcular")
}

// compruebaEnumeradosDePlazos exige que los enumerados de la salida publicada
// en schemas/plazo.json, que salen de las etiquetas de los tipos de salida,
// digan lo mismo que los vocabularios del dominio (contracts/applet-plazos.md
// §7 de H9).
func compruebaEnumeradosDePlazos(t *testing.T) {
	t.Helper()

	publicado, _ := ficheroPublicadoDelVerbo(t, "calcular")

	enumerado := func(tipo, propiedad string) any {
		return valorDelEsquema(t, publicado, "$defs", "calcular", "$defs", tipo, "properties", propiedad, "enum")
	}

	assert.Equal(t, comoLista(plazos.CodigosDeTipo()), enumerado("plazos.Plazo", "tipo"))
	assert.Equal(t, comoLista(plazos.Ambitos()), enumerado("plazos.FestivoSaltado", "ambito"))
	assert.Equal(t, comoLista(plazos.CodigosDeAviso()), enumerado("plazos.Aviso", "codigo"))
	assert.Equal(t, comoLista(plazos.AmbitosDeCobertura()), enumerado("plazos.TramoDeCobertura", "ambito"))
	assert.Equal(t, comoLista(plazos.FaltasDeCobertura()), enumerado("plazos.TramoDeCobertura", "falta"))
}

// comoLista es un vocabulario como lo da el validador al leer un enumerado.
func comoLista[T ~string](valores []T) []any {
	lista := make([]any, 0, len(valores))
	for _, valor := range valores {
		lista = append(lista, string(valor))
	}

	return lista
}

// casoDeLosMunicipios es una fila de los fallos de contracts/applet-plazos.md
// §6 de H9 que dependen de los municipios: la invocación, el código y la clase
// que da, y su mensaje, compuesto con lo que da territorio de cada municipio
// pedido.
type casoDeLosMunicipios struct {
	nombre  string
	orden   []string
	codigo  int
	clase   schema.Clase
	mensaje func(t *testing.T, registro *territorio.Registro) string
}

// comprueba ejecuta la invocación y exige el sobre de fallo del applet, con su
// código, su clase y su mensaje exacto.
func (caso casoDeLosMunicipios) comprueba(t *testing.T) {
	t.Helper()

	fuentes := fuentesDePlazos(t)

	registro, err := territorio.Cargar(fuentes.Territorio)
	require.NoError(t, err)

	res := ordenDePlazos(t, registroDePlazos(t, fuentes), caso.orden...)

	exigirSobreDeFallo(t, res, caso.clase, caso.codigo, procedenciaDeUnFalloDePlazos)
	assert.Equal(t, caso.mensaje(t, registro), datosDelSobre(t, sobreDelJSON(t, res.salida))["mensaje"])
}

// casosDeLosMunicipios son los fallos de los municipios y su orden: el de un
// nombre de varios municipios, con los candidatos tal como los da territorio;
// el de uno que no está en la relación, con 3; el 2 antes que el 3, aunque el
// que daría el 3 vaya antes; los de varios, separados por « | »; y lo que
// decide el dominio antes que ningún municipio (H9 FR-007, FR-010, FR-012;
// research D10, V17).
func casosDeLosMunicipios() []casoDeLosMunicipios {
	return []casoDeLosMunicipios{
		{
			nombre: "un-nombre-de-varios-municipios",
			orden:  []string{"reposicion", "2026-09-09", "Arroyomolinos", "Leganés"},
			codigo: 2, clase: schema.ClaseArgumentos,
			mensaje: func(t *testing.T, registro *territorio.Registro) string {
				t.Helper()

				return comoLoDiceTerritorio(t, registro, "Arroyomolinos")
			},
		},
		{
			nombre: "un-municipio-que-no-esta-en-la-relacion",
			orden:  []string{"reposicion", "2026-09-09", "Leganés", "Inexistente"},
			codigo: 3, clase: schema.ClaseNoEncontrado,
			mensaje: func(t *testing.T, registro *territorio.Registro) string {
				t.Helper()

				return comoLoDiceTerritorio(t, registro, "Inexistente")
			},
		},
		{
			nombre: "el-dos-antes-que-el-tres",
			orden:  []string{"reposicion", "2026-09-09", "Inexistente", "Arroyomolinos"},
			codigo: 2, clase: schema.ClaseArgumentos,
			mensaje: func(t *testing.T, registro *territorio.Registro) string {
				t.Helper()

				return comoLoDiceTerritorio(t, registro, "Arroyomolinos")
			},
		},
		{
			nombre: "dos-municipios-que-no-valen",
			orden:  []string{"reposicion", "2026-09-09", "Arroyomolinos", "00000"},
			codigo: 2, clase: schema.ClaseArgumentos,
			mensaje: func(t *testing.T, registro *territorio.Registro) string {
				t.Helper()

				return comoLoDiceTerritorio(t, registro, "Arroyomolinos") + " | " +
					comoLoDiceTerritorio(t, registro, "00000")
			},
		},
		{
			nombre: "el-tipo-antes-que-los-municipios",
			orden:  []string{"inventado", "2026-09-09", "Inexistente", "Leganés"},
			codigo: 2, clase: schema.ClaseArgumentos,
			mensaje: comoLoDiceElDominio(plazos.Peticion{Tipo: "inventado", Inicio: "2026-09-09"}, 2),
		},
		{
			nombre: "el-numero-antes-que-los-municipios",
			orden:  []string{"reposicion", "2026-09-09", "Inexistente"},
			codigo: 2, clase: schema.ClaseArgumentos,
			mensaje: comoLoDiceElDominio(plazos.Peticion{Tipo: "reposicion", Inicio: "2026-09-09"}, 1),
		},
	}
}

// comoLoDiceTerritorio es el mensaje de un municipio pedido que territorio no
// resuelve: el suyo tras «municipio «<pedido>»: » (contracts/applet-plazos.md
// §6 de H9).
func comoLoDiceTerritorio(t *testing.T, registro *territorio.Registro, pedido string) string {
	t.Helper()

	_, err := registro.Resolver(pedido)
	require.Error(t, err, "territorio no resuelve %q", pedido)

	return "municipio «" + pedido + "»: " + err.Error()
}

// comoLoDiceElDominio es el mensaje con el que el dominio rechaza la petición
// con ese número de municipios, sin resolver ninguno.
func comoLoDiceElDominio(
	peticion plazos.Peticion, municipios int,
) func(t *testing.T, registro *territorio.Registro) string {
	return func(t *testing.T, _ *territorio.Registro) string {
		t.Helper()

		tipos, err := plazos.LeerTipos(fuentesDePlazos(t).Tipos)
		require.NoError(t, err)

		peticion.Municipios = make([]plazos.MunicipioResuelto, municipios)

		err = plazos.ComprobarPeticion(tipos, peticion)
		require.Error(t, err)

		return err.Error()
	}
}

// fuentesDePlazosQueNoCargan es un caso de fuentes con un fichero que no carga:
// las del binario con ese fichero vaciado, y el defecto que da su lector, que
// el mensaje tiene que llevar.
type fuentesDePlazosQueNoCargan struct {
	nombre  string
	vaciar  func(fuentes *FicherosDePlazos)
	defecto func(fuentes FicherosDePlazos) error
}

// casosDeFuentesDePlazosQueNoCargan son los tres: la relación de territorio,
// los tipos y los festivos nacionales y autonómicos.
func casosDeFuentesDePlazosQueNoCargan() []fuentesDePlazosQueNoCargan {
	return []fuentesDePlazosQueNoCargan{
		{
			nombre: "sin-el-territorio",
			vaciar: func(fuentes *FicherosDePlazos) { fuentes.Territorio = territorio.Fuentes{} },
			defecto: func(fuentes FicherosDePlazos) error {
				_, err := territorio.Cargar(fuentes.Territorio)

				return err
			},
		},
		{
			nombre: "sin-los-tipos",
			vaciar: func(fuentes *FicherosDePlazos) { fuentes.Tipos = nil },
			defecto: func(fuentes FicherosDePlazos) error {
				_, err := plazos.LeerTipos(fuentes.Tipos)

				return err
			},
		},
		{
			nombre: "sin-los-festivos-nacionales",
			vaciar: func(fuentes *FicherosDePlazos) { fuentes.Nacionales = nil },
			defecto: func(fuentes FicherosDePlazos) error {
				_, err := plazos.CargarCalendario(fuentes.Nacionales, fuentes.Locales)

				return err
			},
		},
	}
}

// compruebaDefectoDeComposicionDePlazos exige que la ayuda y --describe
// respondan con unas fuentes que no cargan, y que ejecutar el verbo sea el
// defecto de composición, las dos veces —inesperado y firmado por el kernel—,
// con el defecto del fichero en el mensaje (contracts/applet-plazos.md §6 de
// H9).
func compruebaDefectoDeComposicionDePlazos(t *testing.T, caso fuentesDePlazosQueNoCargan) {
	t.Helper()

	fuentes := fuentesDePlazos(t)
	caso.vaciar(&fuentes)

	errDeCarga := caso.defecto(fuentes)
	require.Error(t, errDeCarga, "las fuentes del caso no cargan")

	registro := registroDePlazos(t, fuentes)

	for _, argv := range [][]string{
		{"kitlegal", "plazos", "--help"},
		{"kitlegal", "plazos", "calcular", "--help"},
		{"kitlegal", "plazos", "calcular", "x", "x", "x", "--describe"},
	} {
		res := invocar(t, registro, argv...)

		assert.Equal(t, 0, res.codigo, "%q no analiza las fuentes: %s", argv, res.errores)
		assert.NotEmpty(t, res.salida, "%q responde", argv)
	}

	for range 2 {
		res := ordenDePlazos(t, registro, "reposicion", "2026-09-09", "Leganés", "Leganés")

		exigirSobreDeFallo(t, res, schema.ClaseInesperado, 1, cli.ProcedenciaKernel())
		assert.Contains(t, res.errores, errDeCarga.Error(), "el mensaje dice qué fichero y qué dato fallan")
	}
}

// compruebaAnalisisUnicoDePlazos vacía los festivos locales de las fuentes
// después de la primera ejecución: si el applet volviera a analizarlas, la
// segunda no saltaría el 9 de octubre de Leganés y daría otra fecha.
func compruebaAnalisisUnicoDePlazos(t *testing.T) {
	t.Helper()

	fuentes := fuentesDePlazos(t)
	registro := registroDePlazos(t, fuentes)

	primera := ordenDePlazos(t, registro, "reposicion", "2026-09-09", "Leganés", "Leganés")
	require.Equal(t, 0, primera.codigo, primera.errores)

	clear(fuentes.Locales)

	segunda := ordenDePlazos(t, registro, "reposicion", "2026-09-09", "Leganés", "Leganés")
	require.Equal(t, 0, segunda.codigo, segunda.errores)
	assert.Equal(t, conLaFechaDeLosGolden(t, primera.salida), conLaFechaDeLosGolden(t, segunda.salida),
		"la segunda responde con lo analizado en la primera")
	assert.Contains(t, primera.salida, `"fecha":"2026-10-13"`, "la primera salta los festivos de Leganés")
}

// TestFestivoDeUnaParte comprueba con los ficheros del binario que un festivo
// que la resolución del BOE limita a una parte de una comunidad cuenta como
// hábil y lleva su aviso (H9 FR-011; contracts/casos-de-plazos.md, «Festivo de
// una parte»; research.md S1 de H9): toma el primero que cae en día laborable y
// un municipio de su comunidad sin fiesta local ese día, y un plazo genérico de
// un día hábil desde la víspera vence ese mismo día, con el aviso
// festivo-de-una-parte que nombra el día y la parte. Si lo empotrado no trae
// ninguno, falla: el hito lo da por hecho.
func TestFestivoDeUnaParte(t *testing.T) {
	t.Parallel()

	fuentes := fuentesDePlazos(t)
	festivo, comunidad := primerFestivoDeUnaParte(t, fuentes.Nacionales)
	municipio := municipioSinFiestaLocal(t, fuentes, comunidad, festivo.Fecha)

	dia, err := time.Parse(time.DateOnly, festivo.Fecha)
	require.NoError(t, err)

	vispera := dia.AddDate(0, 0, -1).Format(time.DateOnly)

	plazo := plazoDeLaSalida(t, ordenDePlazos(t, registroDePlazos(t, fuentes),
		"generico", vispera, municipio, municipio, "--duracion", "1", "--unidad", "dias-habiles"))

	assert.Equal(t, festivo.Fecha, plazo.Fecha, "el festivo de %s cuenta como hábil en %s", festivo.Parte, municipio)

	conElDiaYLaParte := "el " + festivo.Fecha + " es fiesta solo en " + festivo.Parte + " ("
	assert.True(t, slices.ContainsFunc(plazo.Avisos, func(aviso avisoDeLaTabla) bool {
		return aviso.Codigo == deUnaParteDeLaTabla && strings.Contains(aviso.Texto, conElDiaYLaParte)
	}), "ningún aviso %s dice «%s…»: %v", deUnaParteDeLaTabla, conElDiaYLaParte, plazo.Avisos)
}

// primerFestivoDeUnaParte es, en el orden del fichero, el primer festivo de
// data/festivos/nacionales-y-autonomicos.yaml limitado a una parte de una
// comunidad que cae en día laborable para ella —de lunes a viernes y sin un
// festivo de toda la comunidad ese día—, con esa comunidad.
func primerFestivoDeUnaParte(t *testing.T, nacionales []byte) (plazos.FestivoNacional, string) {
	t.Helper()

	var fichero plazos.FestivosNacionales
	require.NoError(t, yaml.Unmarshal(nacionales, &fichero))

	for _, festivo := range fichero.Festivos {
		if festivo.Parte == "" || len(festivo.Comunidades) == 0 {
			continue
		}

		dia, err := time.Parse(time.DateOnly, festivo.Fecha)
		require.NoError(t, err)

		if semana := dia.Weekday(); semana == time.Saturday || semana == time.Sunday {
			continue
		}

		comunidad := festivo.Comunidades[0]
		if !festivoDeToda(fichero.Festivos, festivo.Fecha, comunidad) {
			return festivo, comunidad
		}
	}

	require.FailNow(t, "lo empotrado no trae ningún festivo de una parte en día laborable: el hito lo da por hecho (S1)")

	return plazos.FestivoNacional{}, ""
}

// festivoDeToda dice si una fecha es festivo de toda una comunidad.
func festivoDeToda(festivos []plazos.FestivoNacional, fecha, comunidad string) bool {
	return slices.ContainsFunc(festivos, func(festivo plazos.FestivoNacional) bool {
		return festivo.Fecha == fecha && festivo.Parte == "" && slices.Contains(festivo.Comunidades, comunidad)
	})
}

// municipioSinFiestaLocal es el código INE del primer municipio de la
// comunidad, en el orden de Registro.MunicipiosDe, que no tiene fiesta local
// esa fecha en los festivos locales del binario, si la comunidad los tiene.
func municipioSinFiestaLocal(t *testing.T, fuentes FicherosDePlazos, comunidad, fecha string) string {
	t.Helper()

	registro, err := territorio.Cargar(fuentes.Territorio)
	require.NoError(t, err)

	municipios := registro.MunicipiosDe(comunidad)
	require.NotEmpty(t, municipios, "data/territorio/ tiene municipios de la comunidad %s", comunidad)

	var locales plazos.FestivosLocales
	if contenido, hay := fuentes.Locales[comunidad]; hay {
		require.NoError(t, yaml.Unmarshal(contenido, &locales))
	}

	for _, municipio := range municipios {
		codigo := municipio.Codigo.String()

		if !slices.ContainsFunc(locales.Festivos, func(festivo plazos.FestivoLocal) bool {
			return festivo.Fecha == fecha && festivo.Municipio == codigo
		}) {
			return codigo
		}
	}

	require.FailNow(t, "todos los municipios de la comunidad tienen fiesta local ese día",
		"la comunidad %s, el %s", comunidad, fecha)

	return ""
}
