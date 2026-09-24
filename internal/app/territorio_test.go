package app

import (
	"encoding/json"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/core/territorio"
)

// El registro local de estos tests: los cuatro ficheros de data/territorio/
// escritos aquí, con la forma de los congelados, y compuestos en el applet como
// los compone la raíz de producción. Todo es inventado a propósito —nombres,
// comunidades, boletines, códigos, fechas y procedencias—, de modo que nada de
// esto pueda tomarse por un dato real escrito de memoria y ningún test pase
// porque el applet lo tenga escrito; los ficheros reales los ejerce el guion e2e
// de la matriz territorial.
//
// La comunidad 01 tiene configurados sus dos boletines, el mismo con su motivo;
// la 02 es foral, de dos provincias y sin ninguno. Tres municipios se llaman
// igual, uno en cada provincia, y la relación no los lista en el orden de su
// código. Solo uno tiene DIR3 verificado.
//
// Las fechas son todas distintas, para que cada respuesta se feche con un
// fichero distinto: el municipio cubierto, con la correspondencia DIR3, que es
// la más antigua; el no cubierto, que no la trae, con el de su comunidad; y todo
// fallo que decide el applet, con la relación (data-model §2.8).
const (
	relacionDeTerritorio = `fecha: "2026-03-10"
source: prueba.relacion
municipios:
  "22991": {dc: "8", nombre: "Fuentesanta", provincia: "22", comunidad: "02"}
  "11991": {dc: "4", nombre: "Villaconfigurada", provincia: "11", comunidad: "01"}
  "21992": {dc: "3", nombre: "Fuentesanta", provincia: "21", comunidad: "02"}
  "21991": {dc: "6", nombre: "Robledal del Río", provincia: "21", comunidad: "02"}
  "11992": {dc: "0", nombre: "Fuentesanta", provincia: "11", comunidad: "01"}
`
	correspondenciaDeTerritorio = `fecha: "2026-01-15"
source: prueba.correspondencia
correspondencia:
  "11991": "L01119914"
`
	estadoDeTerritorio = `fecha: "2026-05-01"
source: prueba.estado
boletin:
  codigo: "BOEP"
  nombre: "Boletín Oficial del Estado de Prueba"
  url: "https://estado.example/"
`
	comunidadConfiguradaDeTerritorio = `fecha: "2026-04-01"
source: prueba.nombres
codigo: "01"
nombre: "Comunidad Configurada"
regimen: comun
provincias:
  "11": "Provincia Configurada"
boletines:
  autonomico: {codigo: "BOCC", nombre: "Boletín Oficial de la Comunidad Configurada", url: "https://configurada.example/"}
  provincial: {codigo: "BOCC", nombre: "Boletín Oficial de la Comunidad Configurada", url: "https://configurada.example/", motivo: "Comunidad uniprovincial: su boletín hace también de boletín provincial."}
`
	comunidadSinBoletinesDeTerritorio = `fecha: "2026-02-20"
source: prueba.nombres
codigo: "02"
nombre: "Comunidad Foral Sin Boletines"
regimen: foral
provincias:
  "21": "Provincia Norte"
  "22": "Provincia Sur"
`
)

// Los data enteros que el applet tiene que devolver para los dos municipios
// resueltos del registro local, escritos a mano y no compuestos con el código
// del dominio (data-model §2.5). Cada dato lleva su source; nada va con
// omitempty, así que el DIR3 no verificado sale con sus dos claves vacías y el
// motivo de un boletín que no lo necesita, vacío. El no cubierto trae solo el
// boletín estatal y ningún nombre, código ni dirección de otro (FR-008, FR-021).
const (
	dataDelCubierto = `{
  "municipio":  {"nombre": "Villaconfigurada", "source": "prueba.relacion"},
  "codigo_ine": {"codigo": "11991", "digito_de_control": "4", "source": "prueba.relacion"},
  "provincia":  {"codigo": "11", "nombre": "Provincia Configurada", "source": "prueba.nombres"},
  "comunidad":  {"codigo": "01", "nombre": "Comunidad Configurada", "source": "prueba.nombres"},
  "dir3":       {"codigo": "L01119914", "source": "prueba.correspondencia"},
  "regimen":    {"valor": "comun", "source": "data/territorio/comunidades/01.yaml"},
  "boletines":  [
    {"nivel": "estatal", "codigo": "BOEP", "nombre": "Boletín Oficial del Estado de Prueba",
     "url": "https://estado.example/", "motivo": "", "source": "data/territorio/estado.yaml"},
    {"nivel": "autonomico", "codigo": "BOCC", "nombre": "Boletín Oficial de la Comunidad Configurada",
     "url": "https://configurada.example/", "motivo": "", "source": "data/territorio/comunidades/01.yaml"},
    {"nivel": "provincial", "codigo": "BOCC", "nombre": "Boletín Oficial de la Comunidad Configurada",
     "url": "https://configurada.example/",
     "motivo": "Comunidad uniprovincial: su boletín hace también de boletín provincial.",
     "source": "data/territorio/comunidades/01.yaml"}
  ],
  "cobertura":  {"boletin_autonomico": "configurado", "boletin_provincial": "configurado", "dir3": "verificado"}
}`
	dataDelNoCubierto = `{
  "municipio":  {"nombre": "Robledal del Río", "source": "prueba.relacion"},
  "codigo_ine": {"codigo": "21991", "digito_de_control": "6", "source": "prueba.relacion"},
  "provincia":  {"codigo": "21", "nombre": "Provincia Norte", "source": "prueba.nombres"},
  "comunidad":  {"codigo": "02", "nombre": "Comunidad Foral Sin Boletines", "source": "prueba.nombres"},
  "dir3":       {"codigo": "", "source": ""},
  "regimen":    {"valor": "foral", "source": "data/territorio/comunidades/02.yaml"},
  "boletines":  [
    {"nivel": "estatal", "codigo": "BOEP", "nombre": "Boletín Oficial del Estado de Prueba",
     "url": "https://estado.example/", "motivo": "", "source": "data/territorio/estado.yaml"}
  ],
  "cobertura":  {"boletin_autonomico": "no-configurado", "boletin_provincial": "no-configurado", "dir3": "no-verificado"}
}`
)

// Las fechas de consulta con las que se firma cada respuesta del registro
// local, a medianoche UTC.
const (
	fechaDelCubierto              = "2026-01-15T00:00:00Z"
	fechaDelNoCubierto            = "2026-02-20T00:00:00Z"
	fechaDeLaRelacionDeTerritorio = "2026-03-10T00:00:00Z"
)

// firmaDeTerritorio es la procedencia con la que firma el applet: la del
// espacio reservado, porque no consulta ninguna fuente en ejecución (FR-004,
// contrato del applet §2). Se escribe aquí entera, y no con las constantes del
// applet, para que un cambio en ellas no pase por aquí en silencio.
var firmaDeTerritorio = schema.Procedencia{
	Fuente: "kitlegal.territorio",
	URL:    "kitlegal:applet/territorio",
}

// clavesDelTerritorio son las ocho claves de primer nivel del data de
// resolver, las de la entrega (FR-006).
var clavesDelTerritorio = []string{
	"municipio", "codigo_ine", "provincia", "comunidad", "dir3", "regimen", "boletines", "cobertura",
}

// codigosQueNuncaDa son los que el applet no decide nunca: no consulta fuentes
// ni cruza la frontera humana, y sus datos no pueden faltar en ejecución
// (FR-016). El 4 solo lo pone el kernel, para todo applet, cuando se agota el
// plazo de --timeout (FR-020 de H1).
var codigosQueNuncaDa = []int{4, 5, 6}

// fuentesDeTerritorio devuelve el registro local, nuevo en cada llamada: cada
// test puede cambiar las suyas sin tocar las de ningún otro.
func fuentesDeTerritorio() territorio.Fuentes {
	return territorio.Fuentes{
		Municipios: []byte(relacionDeTerritorio),
		DIR3:       []byte(correspondenciaDeTerritorio),
		Estado:     []byte(estadoDeTerritorio),
		Comunidades: map[string][]byte{
			"01": []byte(comunidadConfiguradaDeTerritorio),
			"02": []byte(comunidadSinBoletinesDeTerritorio),
		},
	}
}

// registroDeTerritorio construye el registro de una invocación de prueba: el
// applet territorio compuesto con esas fuentes, y nada más.
func registroDeTerritorio(t *testing.T, fuentes territorio.Fuentes) *Registro {
	t.Helper()

	var registro Registro

	require.NoError(t, registro.Registrar(AppletTerritorio(fuentes)))

	return &registro
}

// argvDeResolver es la invocación de «kitlegal territorio resolver» con la
// consulta y las banderas.
func argvDeResolver(consulta string, banderas ...string) []string {
	return slices.Concat([]string{"kitlegal", "territorio", "resolver", consulta}, banderas)
}

// municipioResuelto es un municipio del registro local tal como tiene que
// resolverlo el applet: las consultas que lo nombran —por su nombre, con y sin
// mayúsculas y diacríticos, por su código y por su código con el dígito—, que
// dan todas la misma salida, su data entero y la fecha con que se firma.
type municipioResuelto struct {
	nombre    string
	consultas []string
	data      string
	fecha     string
}

// municipiosResueltos son el municipio cubierto y el no cubierto.
func municipiosResueltos() []municipioResuelto {
	return []municipioResuelto{
		{
			nombre:    "cubierto",
			consultas: []string{"Villaconfigurada", "11991", "119914", "VILLACONFIGURADA", "villaconfigurada"},
			data:      dataDelCubierto,
			fecha:     fechaDelCubierto,
		},
		{
			nombre:    "no-cubierto",
			consultas: []string{"Robledal del Río", "21991", "219916", "robledal del rio", "ROBLEDAL DEL RÍO"},
			data:      dataDelNoCubierto,
			fecha:     fechaDelNoCubierto,
		},
	}
}

// TestResolverDevuelveElTerritorio fija por el kernel en proceso, sobre el
// registro local, lo que resolver devuelve de un municipio cubierto y de uno
// no cubierto (FR-003 a FR-006, FR-008, FR-009, FR-021 a FR-023, SC-001,
// SC-002, contrato del applet §2, §3 y §5):
//
//   - el sobre de seis claves, firmado en el espacio reservado y fechado con la
//     fecha más antigua de los ficheros que sostienen su data;
//   - un data con las ocho claves de la entrega, ni una más ni una menos, sin
//     omitempty y con el source de cada dato;
//   - los boletines del cubierto y los del no cubierto, que solo trae el estatal;
//   - la cobertura con sus tres claves y su vocabulario, y la invariante del
//     DIR3 no verificado;
//   - la misma salida byte a byte por nombre, por código, con el dígito, con
//     --offline y por el enlace simbólico;
//   - y con --dry-run, el applet se ejecuta igual y el kernel no emite sobre.
func TestResolverDevuelveElTerritorio(t *testing.T) {
	t.Parallel()

	for _, municipio := range municipiosResueltos() {
		t.Run(municipio.nombre, func(t *testing.T) {
			t.Parallel()

			registro := registroDeTerritorio(t, fuentesDeTerritorio())

			res := invocar(t, registro, argvDeResolver(municipio.consultas[0], "--json")...)
			require.Equal(t, 0, res.codigo, res.errores)

			sobre := sobreDelJSON(t, res.salida)

			compruebaFirmaDeTerritorio(t, sobre, municipio.fecha)
			compruebaDataDeTerritorio(t, sobre["data"], municipio.data)
			compruebaMismaSalida(t, registro, municipio.consultas, res.salida)
			compruebaEnsayoDeTerritorio(t, registro, municipio.consultas[0])
		})
	}
}

// compruebaFirmaDeTerritorio exige el sobre de éxito del applet: ok, la
// procedencia del espacio reservado, la fecha de los ficheros —no la del
// reloj— y una huella.
func compruebaFirmaDeTerritorio(t *testing.T, sobre map[string]any, fecha string) {
	t.Helper()

	assert.Equal(t, true, sobre["ok"])
	assert.Equal(t, firmaDeTerritorio.Fuente, sobre["fuente"], "FR-004")
	assert.Equal(t, firmaDeTerritorio.URL, sobre["url"], "FR-004")
	assert.Equal(t, fecha, sobre["fecha_consulta"],
		"la más antigua de las fechas de los ficheros que sostienen el data (data-model §2.8)")
	assert.Regexp(t, schema.PatronHuella, sobre["hash"])
}

// compruebaDataDeTerritorio exige el data entero y, por separado, las tres
// cosas que el contrato del applet §3 fija de su forma.
func compruebaDataDeTerritorio(t *testing.T, data any, esperado string) {
	t.Helper()

	objeto := objetoDe(t, data, "data")

	assert.ElementsMatch(t, clavesDelTerritorio, slices.Collect(maps.Keys(objeto)),
		"las ocho claves de la entrega, ni una más ni una menos (FR-006)")

	contenido, err := json.Marshal(objeto)
	require.NoError(t, err)

	assert.JSONEq(t, esperado, string(contenido),
		"cada dato con su source y ninguno con omitempty: lo que no hay va vacío (FR-005, FR-008, FR-021)")

	compruebaCobertura(t, objetoDe(t, objeto["cobertura"], "cobertura"))
	compruebaDIR3NoVerificado(t, objeto)
}

// compruebaCobertura exige las tres claves de la cobertura, cada una con un
// valor de su vocabulario, que es el del dominio y no tiene ninguno que
// signifique «no existe» (FR-020, FR-022).
func compruebaCobertura(t *testing.T, cobertura map[string]any) {
	t.Helper()

	vocabulario := territorio.AspectosDeCobertura()
	claves := make([]string, 0, len(vocabulario))

	for _, aspecto := range vocabulario {
		claves = append(claves, aspecto.Clave)

		assert.Contains(t, aspecto.Valores, cobertura[aspecto.Clave],
			"cobertura.%s está en su vocabulario", aspecto.Clave)
	}

	assert.ElementsMatch(t, claves, slices.Collect(maps.Keys(cobertura)),
		"la cobertura enumera sus tres claves siempre")
}

// compruebaDIR3NoVerificado exige la invariante del DIR3 no verificado:
// dir3.codigo vacío si y solo si dir3.source vacío, si y solo si la cobertura
// lo declara no verificado. Nunca un código derivado presentado como registral
// (FR-023, data-model §2.5).
func compruebaDIR3NoVerificado(t *testing.T, objeto map[string]any) {
	t.Helper()

	dir3 := objetoDe(t, objeto["dir3"], "dir3")
	cobertura := objetoDe(t, objeto["cobertura"], "cobertura")

	sinCodigo := dir3["codigo"] == ""

	assert.Equal(t, sinCodigo, dir3["source"] == "", "dir3.codigo vacío ⟺ dir3.source vacío")
	assert.Equal(t, sinCodigo, cobertura["dir3"] == "no-verificado",
		"dir3.codigo vacío ⟺ cobertura.dir3 no-verificado")
}

// objetoDe exige que un valor del sobre decodificado sea un objeto.
func objetoDe(t *testing.T, valor any, nombre string) map[string]any {
	t.Helper()

	objeto, esObjeto := valor.(map[string]any)
	require.True(t, esObjeto, "%s es un objeto", nombre)

	return objeto
}

// compruebaMismaSalida exige que cada consulta del municipio dé exactamente la
// misma salida que la primera, con --offline y sin ella, por el primer argumento
// y por el enlace simbólico del multicall, con las banderas que no cambian nada:
// el mismo data, la misma huella y la misma fecha, que no depende del reloj
// (SC-001, FR-009, contrato del applet §2 y §5).
func compruebaMismaSalida(t *testing.T, registro *Registro, consultas []string, salida string) {
	t.Helper()

	invocaciones := [][]string{
		{"/usr/local/bin/territorio", "resolver", consultas[0], "--json", "--offline", "--no-graph", "--asunto", "demo"},
	}

	for _, consulta := range consultas {
		invocaciones = append(invocaciones,
			argvDeResolver(consulta, "--json"),
			argvDeResolver(consulta, "--json", "--offline"),
			[]string{"territorio", "resolver", consulta, "--json"},
		)
	}

	for _, argv := range invocaciones {
		res := invocar(t, registro, argv...)

		require.Equal(t, 0, res.codigo, "%q: %s", argv, res.errores)
		assert.Equal(t, salida, res.salida, "%q da la misma salida byte a byte", argv)
	}
}

// compruebaEnsayoDeTerritorio exige lo que el contrato del applet §5 dice de
// --dry-run: el applet se ejecuta igual y no describe ninguna operación, porque
// no tiene ninguna capa con efectos, y el kernel escribe su línea en la salida
// de error y no emite sobre (research.md D7).
func compruebaEnsayoDeTerritorio(t *testing.T, registro *Registro, consulta string) {
	t.Helper()

	res := invocar(t, registro, argvDeResolver(consulta, "--json", "--dry-run")...)

	assert.Equal(t, 0, res.codigo, res.errores)
	assert.Empty(t, res.salida, "con --dry-run no hay sobre")
	assert.True(t, strings.HasPrefix(res.errores, prefijoDeEnsayo), "la línea del kernel: %q", res.errores)
	assert.NotContains(t, res.errores, "se habría pedido", "el applet no pide nada")
}

// casoDeTerritorio es una invocación de resolver que no se resuelve, con el
// código y la clase con que termina.
type casoDeTerritorio struct {
	nombre string
	argv   []string
	codigo int
	clase  schema.Clase
	// consulta es la que decide el applet, y delApplet dice que la decide él:
	// entonces firma él, con la fecha de la relación, y su mensaje es el del
	// dominio y nombra la entrada. Si no, el fallo es anterior al applet y lo
	// firma el kernel (contrato del applet §2).
	consulta  string
	delApplet bool
	// mensaje comprueba lo que el caso exige además de su mensaje, o es nulo.
	mensaje func(t *testing.T, mensaje string)
}

// grupoDeCasos son los casos de uno de los subtests de TestCodigosDeTerritorio.
type grupoDeCasos struct {
	nombre string
	casos  []casoDeTerritorio
}

// candidatosDeFuentesanta son los tres municipios del registro local que se
// llaman Fuentesanta, en la forma fija «<código INE> <nombre> (<provincia>)»,
// separados por «; » y en el orden de su código INE, que no es el de la
// relación (FR-014, research.md D14).
const candidatosDeFuentesanta = "11992 Fuentesanta (Provincia Configurada); " +
	"21992 Fuentesanta (Provincia Norte); 22991 Fuentesanta (Provincia Sur)"

// gruposDeCasos son los casos de TestCodigosDeTerritorio, por subtest.
func gruposDeCasos() []grupoDeCasos {
	decidido := func(nombre, consulta string, clase schema.Clase, codigo int) casoDeTerritorio {
		return casoDeTerritorio{
			nombre: nombre, argv: argvDeResolver(consulta, "--json"), codigo: codigo, clase: clase,
			consulta: consulta, delApplet: true,
		}
	}
	argumentos := func(nombre, consulta string) casoDeTerritorio {
		return decidido(nombre, consulta, schema.ClaseArgumentos, 2)
	}
	noEncontrado := func(nombre, consulta string) casoDeTerritorio {
		return decidido(nombre, consulta, schema.ClaseNoEncontrado, 3)
	}
	ambiguo := func(nombre, consulta string) casoDeTerritorio {
		caso := argumentos(nombre, consulta)
		caso.mensaje = func(t *testing.T, mensaje string) {
			t.Helper()

			assert.True(t, strings.HasSuffix(mensaje, candidatosDeFuentesanta),
				"todos los candidatos, con su provincia y en orden de código INE: %q", mensaje)
		}

		return caso
	}

	digitoIncorrecto := argumentos("distinto-del-oficial", "119910")
	digitoIncorrecto.mensaje = func(t *testing.T, mensaje string) {
		t.Helper()

		assert.Contains(t, mensaje, `"0"`, "el dígito recibido")
		assert.Contains(t, mensaje, `"4"`, "el oficial")
	}

	kernel := func(nombre string, argv ...string) casoDeTerritorio {
		return casoDeTerritorio{nombre: nombre, argv: argv, codigo: 2, clase: schema.ClaseArgumentos}
	}

	return []grupoDeCasos{
		{nombre: "ambiguo", casos: []casoDeTerritorio{
			ambiguo("como-lo-escribe-la-relacion", "Fuentesanta"),
			ambiguo("en-minusculas", "fuentesanta"),
		}},
		{nombre: "no-encontrado", casos: []casoDeTerritorio{
			noEncontrado("por-nombre", "Villainventada"),
			noEncontrado("por-codigo", "11993"),
			noEncontrado("por-codigo-con-digito", "119935"),
			noEncontrado("por-codigo-de-una-provincia-sin-comunidad", "52991"),
		}},
		{nombre: "codigo-mal-formado", casos: []casoDeTerritorio{
			argumentos("cuatro-cifras", "1199"),
			argumentos("siete-cifras", "1199140"),
			argumentos("cifras-con-un-espacio", "11 991"),
			argumentos("provincia-00", "00991"),
			argumentos("provincia-53", "53991"),
			argumentos("provincia-99", "99991"),
			argumentos("provincia-00-con-digito", "009914"),
			argumentos("municipio-000", "11000"),
			argumentos("vacia", ""),
		}},
		{nombre: "digito-incorrecto", casos: []casoDeTerritorio{digitoIncorrecto}},
		{nombre: "sin-argumento", casos: []casoDeTerritorio{
			kernel("sin-consulta", "kitlegal", "territorio", "resolver", "--json"),
			kernel("sin-consulta-por-el-enlace", "territorio", "resolver", "--json"),
			kernel("sin-verbo", "kitlegal", "territorio", "--json"),
		}},
	}
}

// TestCodigosDeTerritorio fija por el kernel en proceso, sobre el registro
// local, el código, la clase, la procedencia y el mensaje de cada consulta que
// no se resuelve (FR-010 a FR-016, US4, SC-004, contrato del applet §4):
//
//   - un nombre de varios municipios es 2, con todos los candidatos en el
//     mensaje, en la forma fija y en orden de código INE;
//   - un nombre o un código bien formado —provincia de 01 a 52 y municipio
//     ausente del registro local— que no está en la relación es 3;
//   - unas cifras que no forman un código —de más o de menos, con la provincia
//     00 o mayor que 52, con el municipio 000— son 2, con clase argumentos y
//     nunca 3, y lo mismo el dígito que no es el oficial;
//   - la invocación sin la consulta, o sin el verbo, es 2 y la firma el kernel;
//   - ninguna invocación, con ninguna bandera y con un plazo que la carga
//     cumple, termina en 4, 5 ni 6;
//   - y un plazo que la carga no puede cumplir termina en 4, con clase fuente
//     no disponible y la firma del applet: lo decide el kernel, no el applet.
func TestCodigosDeTerritorio(t *testing.T) {
	t.Parallel()

	for _, grupo := range gruposDeCasos() {
		t.Run(grupo.nombre, func(t *testing.T) {
			t.Parallel()

			for _, caso := range grupo.casos {
				t.Run(caso.nombre, func(t *testing.T) {
					t.Parallel()

					caso.comprueba(t)
				})
			}
		})
	}

	t.Run("ninguna-invocacion-da-4-5-ni-6", func(t *testing.T) {
		t.Parallel()

		compruebaQueNuncaDaCuatroCincoNiSeis(t)
	})

	t.Run("plazo-agotado-da-4-y-lo-decide-el-kernel", func(t *testing.T) {
		t.Parallel()

		compruebaPlazoAgotado(t)
	})
}

// compruebaPlazoAgotado exige la única excepción de FR-016, la que fija H1 para
// todo applet: con un plazo de un nanosegundo, vencido antes de que el applet
// termine, la invocación es 4 con clase fuente no disponible y la firma del
// applet, aunque el applet habría resuelto la consulta (FR-020).
func compruebaPlazoAgotado(t *testing.T) {
	t.Helper()

	consulta := municipiosResueltos()[0].consultas[0]
	res := invocar(t, registroDeTerritorio(t, fuentesDeTerritorio()),
		argvDeResolver(consulta, "--json", "--timeout", "1ns")...)

	exigirSobreDeFallo(t, res, schema.ClaseFuenteNoDisponible, 4, firmaDeTerritorio)
	assert.Contains(t, res.errores, "plazo", "el mensaje dice que lo que falló fue el plazo")
}

// comprueba invoca el caso sobre el registro local y exige su sobre de fallo.
func (caso casoDeTerritorio) comprueba(t *testing.T) {
	t.Helper()

	res := invocar(t, registroDeTerritorio(t, fuentesDeTerritorio()), caso.argv...)

	procedencia := cli.ProcedenciaKernel()
	if caso.delApplet {
		procedencia = firmaDeTerritorio
	}

	exigirSobreDeFallo(t, res, caso.clase, caso.codigo, procedencia)
	assert.NotContains(t, codigosQueNuncaDa, res.codigo, "FR-016")

	sobre := sobreDelJSON(t, res.salida)

	mensaje, esTexto := datosDelSobre(t, sobre)["mensaje"].(string)
	require.True(t, esTexto, "el mensaje es un texto")
	assert.Contains(t, res.errores, mensaje, "el mismo mensaje sale para la persona")

	if caso.delApplet {
		assert.Equal(t, fechaDeLaRelacionDeTerritorio, sobre["fecha_consulta"],
			"un fallo que decide el applet se fecha con la relación, no con el reloj")
		assert.Equal(t, mensajeDelDominio(t, caso.consulta), mensaje,
			"el mensaje del dominio llega literal al sobre (research.md V17)")
		assert.Contains(t, mensaje, strconv.Quote(caso.consulta), "el mensaje nombra la entrada")
	}

	if caso.mensaje != nil {
		caso.mensaje(t, mensaje)
	}
}

// mensajeDelDominio es el mensaje con el que el dominio rechaza la consulta
// sobre el mismo registro local.
func mensajeDelDominio(t *testing.T, consulta string) string {
	t.Helper()

	registro, err := territorio.Cargar(fuentesDeTerritorio())
	require.NoError(t, err)

	_, err = registro.Resolver(consulta)
	require.Error(t, err, "el dominio no resuelve %q", consulta)

	return err.Error()
}

// banderasDeTerritorio son las combinaciones de banderas globales con las que
// se barren todas las consultas: ninguna cambia lo que el applet puede
// devolver. El plazo de --timeout es uno que la carga cumple; el que no, lo
// prueba compruebaPlazoAgotado.
var banderasDeTerritorio = [][]string{
	nil,
	{"--json"},
	{"--offline"},
	{"--json", "--offline"},
	{"--dry-run"},
	{"--json", "--dry-run"},
	{"--json", "--no-graph", "--asunto", "demo"},
	{"--json", "--timeout", "5s"},
	{"--verbose"},
}

// compruebaQueNuncaDaCuatroCincoNiSeis invoca todas las consultas de estos
// tests, las que se resuelven y las que no, con cada combinación de banderas
// globales, sobre un mismo applet, y exige que todas terminen en 0, 2 o 3
// (FR-016, SC-004): con un plazo que la carga cumple, el applet no decide
// nunca 4, 5 ni 6.
func compruebaQueNuncaDaCuatroCincoNiSeis(t *testing.T) {
	t.Helper()

	var consultas []string

	for _, municipio := range municipiosResueltos() {
		consultas = append(consultas, municipio.consultas...)
	}

	for _, grupo := range gruposDeCasos() {
		for _, caso := range grupo.casos {
			if caso.delApplet {
				consultas = append(consultas, caso.consulta)
			}
		}
	}

	registro := registroDeTerritorio(t, fuentesDeTerritorio())

	for _, consulta := range consultas {
		for _, banderas := range banderasDeTerritorio {
			argv := argvDeResolver(consulta, banderas...)
			res := invocar(t, registro, argv...)

			assert.NotContains(t, codigosQueNuncaDa, res.codigo, "%q: %s", argv, res.errores)
			assert.Contains(t, []int{0, 2, 3}, res.codigo, "%q: %s", argv, res.errores)
		}
	}
}

// fuentesQueNoCargan son unas fuentes con las que el applet se compone y que
// no cargan: un defecto de composición, que no puede darse en el binario
// publicado porque sus ficheros viajan dentro y make ci los valida (contrato
// del applet §7).
type fuentesQueNoCargan struct {
	nombre  string
	fuentes func() territorio.Fuentes
}

// casosDeFuentesQueNoCargan son unas fuentes que no se analizan y unas que no
// pasan la integridad.
func casosDeFuentesQueNoCargan() []fuentesQueNoCargan {
	return []fuentesQueNoCargan{
		{nombre: "fuentes-vacias", fuentes: func() territorio.Fuentes { return territorio.Fuentes{} }},
		{nombre: "fuentes-sin-integridad", fuentes: func() territorio.Fuentes {
			fuentes := fuentesDeTerritorio()
			delete(fuentes.Comunidades, "02")

			return fuentes
		}},
	}
}

// TestAppletTerritorio fija el applet tal como lo declara el contrato del
// applet §1 y lo que hace al componerse (§7):
//
//   - se llama territorio y declara un único verbo, resolver, que no es por
//     omisión, con un único argumento obligatorio, Consulta, que va por su
//     posición, ninguna bandera propia y el valor cero del territorio resuelto
//     como data;
//   - unas fuentes que no cargan son un defecto de composición —código 1, clase
//     inesperado y firma del kernel—, nunca un código de usuario, y ni la ayuda
//     ni --describe las analizan, así que responden igual;
//   - el applet analiza sus fuentes una sola vez, en su primera ejecución, y
//     cada applet las suyas: el análisis vive en su valor y no en el paquete.
func TestAppletTerritorio(t *testing.T) {
	t.Parallel()

	t.Run("declara-el-verbo-resolver", func(t *testing.T) {
		t.Parallel()

		compruebaDeclaracionDeTerritorio(t)
	})

	for _, caso := range casosDeFuentesQueNoCargan() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			compruebaDefectoDeComposicion(t, caso.fuentes())
		})
	}

	t.Run("analiza-sus-fuentes-una-sola-vez", func(t *testing.T) {
		t.Parallel()

		compruebaAnalisisUnico(t)
	})

	t.Run("cada-applet-analiza-las-suyas", func(t *testing.T) {
		t.Parallel()

		compruebaAnalisisPorApplet(t)
	})
}

// compruebaDeclaracionDeTerritorio exige la declaración del applet.
func compruebaDeclaracionDeTerritorio(t *testing.T) {
	t.Helper()

	applet := AppletTerritorio(fuentesDeTerritorio())

	assert.Equal(t, "territorio", applet.Nombre())
	assert.NotEmpty(t, applet.Descripcion(), "la ayuda del binario lo describe")

	verbos := applet.Verbos()
	require.Len(t, verbos, 1, "resolver es el único verbo")

	assert.Equal(t, "resolver", verbos[0].Nombre)
	assert.NotEmpty(t, verbos[0].Descripcion, "la ayuda del applet lo describe")
	assert.False(t, verbos[0].PorOmision, "ningún verbo por omisión: nombrar resolver es obligatorio")

	compruebaArgumentosDelContrato(t, verbos[0], verboDelContrato{
		campos: []campoDelContrato{{campo: "Consulta", nombre: "consulta", tipo: reflect.TypeFor[string]()}},
		salida: territorio.Territorio{},
	})

	var registro Registro

	require.NoError(t, registro.Registrar(applet), "el registro lo admite")
}

// compruebaDefectoDeComposicion exige que la ayuda y --describe respondan con
// unas fuentes que no cargan, y que ejecutar el verbo sea el defecto de
// composición, las dos veces: la segunda no vuelve a analizarlas, y sigue sin
// haber registro.
func compruebaDefectoDeComposicion(t *testing.T, fuentes territorio.Fuentes) {
	t.Helper()

	_, errDeCarga := territorio.Cargar(fuentes)
	require.Error(t, errDeCarga, "las fuentes del caso no cargan")

	registro := registroDeTerritorio(t, fuentes)

	for _, argv := range [][]string{
		{"kitlegal", "territorio", "--help"},
		{"kitlegal", "territorio", "resolver", "--help"},
		argvDeResolver("Villaconfigurada", "--describe"),
	} {
		res := invocar(t, registro, argv...)

		assert.Equal(t, 0, res.codigo, "%q no analiza las fuentes: %s", argv, res.errores)
		assert.NotEmpty(t, res.salida, "%q responde", argv)
	}

	for range 2 {
		res := invocar(t, registro, argvDeResolver("Villaconfigurada", "--json")...)

		exigirSobreDeFallo(t, res, schema.ClaseInesperado, 1, cli.ProcedenciaKernel())
		assert.Contains(t, res.errores, errDeCarga.Error(), "el mensaje dice qué fichero y qué dato fallan")
	}
}

// compruebaAnalisisUnico vacía las comunidades de las fuentes después de la
// primera ejecución, lo que las deja sin cargar: si el applet volviera a
// analizarlas, la segunda ejecución fallaría.
func compruebaAnalisisUnico(t *testing.T) {
	t.Helper()

	fuentes := fuentesDeTerritorio()
	registro := registroDeTerritorio(t, fuentes)

	primera := invocar(t, registro, argvDeResolver("Villaconfigurada", "--json")...)
	require.Equal(t, 0, primera.codigo, primera.errores)

	clear(fuentes.Comunidades)

	_, err := territorio.Cargar(fuentes)
	require.Error(t, err, "las fuentes vaciadas ya no cargan")

	segunda := invocar(t, registro, argvDeResolver("Villaconfigurada", "--json")...)
	require.Equal(t, 0, segunda.codigo, segunda.errores)
	assert.Equal(t, primera.salida, segunda.salida, "la segunda responde con el registro de la primera")
}

// compruebaAnalisisPorApplet alterna invocaciones de un applet con fuentes que
// no cargan y de otro con fuentes que sí: cada uno responde con las suyas, en
// cualquier orden.
func compruebaAnalisisPorApplet(t *testing.T) {
	t.Helper()

	sanas := registroDeTerritorio(t, fuentesDeTerritorio())
	defectuosas := registroDeTerritorio(t, territorio.Fuentes{})

	for _, paso := range []struct {
		registro *Registro
		codigo   int
	}{
		{registro: defectuosas, codigo: 1},
		{registro: sanas, codigo: 0},
		{registro: defectuosas, codigo: 1},
		{registro: sanas, codigo: 0},
	} {
		res := invocar(t, paso.registro, argvDeResolver("Villaconfigurada", "--json")...)

		assert.Equal(t, paso.codigo, res.codigo, res.errores)
	}
}

// TestSalidaDeTerritorioContraSchemas es el punto 4 de la Definition of Done
// sobre el applet territorio (FR-092, SC-002, contrato del applet §6 y §8): el
// sobre real que emite el kernel con --json, sobre el registro local, valida
// contra la parte de resolver leída de schemas/municipio.json, y no contra lo
// que emite --describe mientras se ejecuta el test. Lo hacen el del municipio
// cubierto, el del no cubierto y el de cada consulta que el applet decide no
// resolver —ambigua, no encontrada, mal formada o con otro dígito—, porque toda
// salida del applet pasa por el contrato. La validación restringe: el mismo
// sobre con una clave de más o de menos en su data no valida.
func TestSalidaDeTerritorioContraSchemas(t *testing.T) {
	t.Parallel()

	publicado, id := ficheroPublicadoDelVerbo(t, "resolver")
	esquema := salidaPublicada(t, publicado, id, "resolver")

	for _, caso := range salidasDeTerritorio() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			res := invocar(t, registroDeTerritorio(t, fuentesDeTerritorio()), argvDeResolver(caso.consulta, "--json")...)

			require.Equal(t, caso.codigo, res.codigo, res.errores)
			assert.Equal(t, caso.codigo == 0, sobreDelJSON(t, res.salida)["ok"],
				"ok decide la rama del esquema contra la que se valida data")

			exigirSalidaPublicada(t, esquema, res.salida)
		})
	}
}

// salidaDeTerritorio es una invocación de TestSalidaDeTerritorioContraSchemas:
// la consulta, que se pasa con --json, y el código con el que termina.
type salidaDeTerritorio struct {
	nombre   string
	consulta string
	codigo   int
}

// salidasDeTerritorio son las invocaciones de resolver cuyo sobre se valida: la
// de éxito de cada municipio resuelto y la de fallo de cada caso que decide el
// applet en TestCodigosDeTerritorio, con su código. Los que decide el kernel
// antes de llegar al applet no son salida suya.
func salidasDeTerritorio() []salidaDeTerritorio {
	var salidas []salidaDeTerritorio

	for _, municipio := range municipiosResueltos() {
		salidas = append(salidas, salidaDeTerritorio{nombre: "exito-" + municipio.nombre, consulta: municipio.consultas[0]})
	}

	for _, grupo := range gruposDeCasos() {
		for _, caso := range grupo.casos {
			if caso.delApplet {
				salidas = append(salidas, salidaDeTerritorio{
					nombre:   "fallo-" + strconv.Itoa(caso.codigo) + "-" + grupo.nombre + "-" + caso.nombre,
					consulta: caso.consulta,
					codigo:   caso.codigo,
				})
			}
		}
	}

	return salidas
}

// presentacionDeTerritorio es una de las dos formas en que el kernel presenta la
// salida del verbo: el sobre, con --json, o la tabla, sin ella.
type presentacionDeTerritorio struct {
	nombre   string
	banderas []string
}

// presentacionesDeTerritorio son las dos: lo que no aparece en una tampoco puede
// aparecer en la otra.
var presentacionesDeTerritorio = []presentacionDeTerritorio{
	{nombre: "json", banderas: []string{"--json"}},
	{nombre: "tabla"},
}

// TestSalidaSinBoletinesNoConfigurados fija sobre el registro local lo que la
// salida del municipio no cubierto dice de sus boletines (FR-020 a FR-022, US2
// escenarios 2, 3 y 4, SC-002):
//
//   - en ninguna de sus dos presentaciones, ni en la salida de error, aparece el
//     código, el nombre ni la dirección de ningún boletín no configurado —los
//     que los ficheros de las demás comunidades configuran—, que la misma
//     búsqueda sí encuentra en la salida del cubierto, para no pasar en vacío;
//   - boletines trae solo el estatal, el del fichero del estado;
//   - la cobertura declara no-configurado el autonómico y el provincial, y el
//     contrato publicado no admite para ellos más valores que configurado y
//     no-configurado: ninguno que diga que no existen.
func TestSalidaSinBoletinesNoConfigurados(t *testing.T) {
	t.Parallel()

	cubierto, noCubierto := municipioLlamado(t, "cubierto"), municipioLlamado(t, "no-cubierto")
	noConfigurados := boletinesConfiguradosFueraDe(t, fuentesDeTerritorio(),
		noCubierto.territorioEsperado(t).Comunidad.Codigo)

	for _, forma := range presentacionesDeTerritorio {
		t.Run("ningun-boletin-no-configurado-en-"+forma.nombre, func(t *testing.T) {
			t.Parallel()

			compruebaSinBoletinesNoConfigurados(t, cubierto, noCubierto, noConfigurados, forma.banderas)
		})
	}

	t.Run("solo-el-boletin-estatal", func(t *testing.T) {
		t.Parallel()

		compruebaSoloElBoletinEstatal(t, noCubierto)
	})

	t.Run("cobertura-sin-valor-de-no-existe", func(t *testing.T) {
		t.Parallel()

		compruebaCoberturaNoConfigurada(t, noCubierto)
	})
}

// municipioLlamado es el municipio resuelto del registro local con ese nombre,
// que tiene que estar.
func municipioLlamado(t *testing.T, nombre string) municipioResuelto {
	t.Helper()

	municipios := municipiosResueltos()

	i := slices.IndexFunc(municipios, func(municipio municipioResuelto) bool {
		return municipio.nombre == nombre
	})
	require.NotEqual(t, -1, i, "%q está entre los municipios resueltos", nombre)

	return municipios[i]
}

// territorioEsperado es el data que el municipio tiene que devolver, leído de su
// esperado escrito a mano y no de lo que emite el applet.
func (municipio municipioResuelto) territorioEsperado(t *testing.T) territorio.Territorio {
	t.Helper()

	var esperado territorio.Territorio

	require.NoError(t, json.Unmarshal([]byte(municipio.data), &esperado), "el data esperado de %s", municipio.nombre)

	return esperado
}

// territorioEmitido es el data del sobre que emite resolver con --json.
func territorioEmitido(t *testing.T, salida string) territorio.Territorio {
	t.Helper()

	var sobre struct {
		Data territorio.Territorio `json:"data"`
	}

	require.NoError(t, json.Unmarshal([]byte(salida), &sobre), "con --json la salida estándar es el sobre")

	return sobre.Data
}

// boletinesConfiguradosFueraDe son los boletines que configuran los ficheros de
// las comunidades de las fuentes distintas de la del municipio, leídos de esos
// ficheros y no de lo que emite el applet. La del municipio no configura
// ninguno, que es lo que lo deja fuera del territorio configurado, así que para
// él todos son boletines no configurados; y tiene que haber alguno, porque
// buscar en su salida una lista vacía no probaría nada.
func boletinesConfiguradosFueraDe(t *testing.T, fuentes territorio.Fuentes, comunidad string) []territorio.BoletinConfigurado {
	t.Helper()

	var boletines []territorio.BoletinConfigurado

	for _, codigo := range slices.Sorted(maps.Keys(fuentes.Comunidades)) {
		var fichero territorio.FicheroDeComunidad

		require.NoError(t, yaml.Unmarshal(fuentes.Comunidades[codigo], &fichero), "el fichero de la comunidad %s", codigo)

		if codigo == comunidad {
			require.Nil(t, fichero.Boletines, "la comunidad %s del municipio no configura ningún boletín", codigo)

			continue
		}

		if fichero.Boletines == nil {
			continue
		}

		for _, boletin := range []*territorio.BoletinConfigurado{fichero.Boletines.Autonomico, fichero.Boletines.Provincial} {
			if boletin != nil {
				boletines = append(boletines, *boletin)
			}
		}
	}

	require.NotEmpty(t, boletines, "las fuentes configuran algún boletín fuera de la comunidad %s", comunidad)

	return boletines
}

// compruebaSinBoletinesNoConfigurados busca, en una presentación, el código, el
// nombre y la dirección de cada boletín no configurado: la búsqueda los
// encuentra en la salida del cubierto, cuya comunidad los configura, y no
// encuentra ninguno en la del no cubierto, ni en su salida estándar ni en la de
// error (FR-021, US2 escenario 3, SC-002).
func compruebaSinBoletinesNoConfigurados(
	t *testing.T, cubierto, noCubierto municipioResuelto, noConfigurados []territorio.BoletinConfigurado, banderas []string,
) {
	t.Helper()

	registro := registroDeTerritorio(t, fuentesDeTerritorio())

	delCubierto := invocar(t, registro, argvDeResolver(cubierto.consultas[0], banderas...)...)
	require.Equal(t, 0, delCubierto.codigo, delCubierto.errores)

	delNoCubierto := invocar(t, registro, argvDeResolver(noCubierto.consultas[0], banderas...)...)
	require.Equal(t, 0, delNoCubierto.codigo, delNoCubierto.errores)

	for _, boletin := range noConfigurados {
		for _, dato := range []string{boletin.Codigo, boletin.Nombre, boletin.URL} {
			require.Contains(t, delCubierto.salida, dato, "la búsqueda encuentra %q donde está configurado", dato)

			assert.NotContains(t, delNoCubierto.salida, dato, "%q no aparece fuera del territorio configurado", dato)
			assert.NotContains(t, delNoCubierto.errores, dato, "%q tampoco aparece en la salida de error", dato)
		}
	}
}

// compruebaSoloElBoletinEstatal exige que los boletines del no cubierto sean
// uno solo, el estatal, con el código, el nombre y la dirección del fichero del
// estado (FR-008, FR-021, US2 escenario 2).
func compruebaSoloElBoletinEstatal(t *testing.T, noCubierto municipioResuelto) {
	t.Helper()

	var estado territorio.FicheroDeEstado

	require.NoError(t, yaml.Unmarshal(fuentesDeTerritorio().Estado, &estado), "el fichero del estado")

	res := invocar(t, registroDeTerritorio(t, fuentesDeTerritorio()), argvDeResolver(noCubierto.consultas[0], "--json")...)
	require.Equal(t, 0, res.codigo, res.errores)

	boletines := territorioEmitido(t, res.salida).Boletines
	require.Len(t, boletines, 1, "solo el estatal: %+v", boletines)

	assert.Equal(t, "estatal", boletines[0].Nivel)
	assert.Equal(t, estado.Boletin.Codigo, boletines[0].Codigo)
	assert.Equal(t, estado.Boletin.Nombre, boletines[0].Nombre)
	assert.Equal(t, estado.Boletin.URL, boletines[0].URL)
}

// compruebaCoberturaNoConfigurada exige que la cobertura del no cubierto declare
// no-configurado el boletín autonómico y el provincial, y que el contrato
// publicado no deje decir de ellos nada más: su enumerado es configurado y
// no-configurado (data-model §2.4), y el mismo sobre con cualquier otro valor
// —uno que dijera que no existen— no valida (FR-020, FR-022, US2 escenario 4).
func compruebaCoberturaNoConfigurada(t *testing.T, noCubierto municipioResuelto) {
	t.Helper()

	res := invocar(t, registroDeTerritorio(t, fuentesDeTerritorio()), argvDeResolver(noCubierto.consultas[0], "--json")...)
	require.Equal(t, 0, res.codigo, res.errores)

	cobertura := territorioEmitido(t, res.salida).Cobertura
	assert.Equal(t, "no-configurado", cobertura.BoletinAutonomico)
	assert.Equal(t, "no-configurado", cobertura.BoletinProvincial)

	publicado, id := ficheroPublicadoDelVerbo(t, "resolver")
	esquema := salidaPublicada(t, publicado, id, "resolver")

	require.NoError(t, esquema.Validate(sobreValidable(t, res.salida)), "el sobre del no cubierto valida")

	for _, aspecto := range []string{"boletin_autonomico", "boletin_provincial"} {
		assert.ElementsMatch(t, []any{"configurado", "no-configurado"},
			valorDelEsquema(t, publicado, "$defs", "resolver", "$defs", "territorio.Cobertura", "properties", aspecto, "enum"),
			"cobertura.%s no tiene en el contrato ningún valor que diga que el boletín no existe", aspecto)

		conOtroValor := sobreValidable(t, res.salida)
		objetoDe(t, objetoDeData(t, conOtroValor)["cobertura"], "cobertura")[aspecto] = "no-existe"

		assert.Errorf(t, esquema.Validate(conOtroValor), "cobertura.%s fuera de su enumerado no valida", aspecto)
	}
}
