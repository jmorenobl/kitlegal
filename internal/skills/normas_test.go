package skills_test

import (
	"encoding/json"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/evals"
	"github.com/jmorenobl/kitlegal/internal/httpx"
	"github.com/jmorenobl/kitlegal/internal/skills"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// Los títulos de las normas de las tablas sintéticas de TestLeerNormas, tal como
// los dan los metadatos grabados en H4 (contrato normas-y-referencias §1 y §5).
const (
	tituloDeLaLPAC = "Ley 39/2015, de 1 de octubre, del Procedimiento Administrativo Común " +
		"de las Administraciones Públicas."
	tituloDeLaLCSP = "Ley 9/2017, de 8 de noviembre, de Contratos del Sector Público, por la que se transponen " +
		"al ordenamiento jurídico español las Directivas del Parlamento Europeo y del Consejo 2014/23/UE " +
		"y 2014/24/UE, de 26 de febrero de 2014."
	tituloDeLaLRBRL = "Ley 7/1985, de 2 de abril, Reguladora de las Bases del Régimen Local"
)

// Trozos de las tablas sintéticas de TestLeerNormas, cada uno con sus líneas
// completas: el principio de la tabla y las normas de la LPAC y de la LCSP, de
// seis líneas cada una, y la de la LRBRL, de cinco y sin abreviatura.
const (
	inicioDeLasNormas = "normas:\n"

	normaDeLaLPAC = "  BOE-A-2015-10565:\n" +
		"    titulo: \"" + tituloDeLaLPAC + "\"\n" +
		"    rango: Ley\n" +
		"    abreviatura: LPAC\n" +
		"    materias:\n" +
		"      - procedimiento administrativo\n"

	normaDeLaLCSP = "  BOE-A-2017-12902:\n" +
		"    titulo: \"" + tituloDeLaLCSP + "\"\n" +
		"    rango: Ley\n" +
		"    abreviatura: LCSP\n" +
		"    materias:\n" +
		"      - contratación pública\n"

	normaDeLaLRBRL = "  BOE-A-1985-5392:\n" +
		"    titulo: \"" + tituloDeLaLRBRL + "\"\n" +
		"    rango: Ley\n" +
		"    materias:\n" +
		"      - régimen local\n"
)

// Líneas de normaDeLaLPAC que los casos de TestLeerNormas cambian.
const (
	claveDeLaLPAC       = "  BOE-A-2015-10565:\n"
	tituloEscritoLPAC   = "    titulo: \"" + tituloDeLaLPAC + "\"\n"
	rangoDeLaLPAC       = "    rango: Ley\n"
	abreviaturaDeLaLPAC = "    abreviatura: LPAC\n"
	materiasDeLaLPAC    = "    materias:\n      - procedimiento administrativo\n"
	materiaDeLaLPAC     = "      - procedimiento administrativo\n"
)

// TestLeerNormas fija LeerNormas (contrato normas-y-referencias §3; US6,
// escenarios 2 y 3; SC-005): una tabla válida da sus normas en el orden del
// documento, con la abreviatura vacía si no la lleva; cada defecto de una norma
// da un error que nombra la norma y el defecto —el identificador repetido con
// sus dos líneas en lugar de quedarse con la segunda entrada, también cuando la
// repetición llega por un alias—; y los defectos que no son de ninguna norma
// dicen qué falla y en qué línea.
func TestLeerNormas(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		documento string
		normas    []skills.Norma
		error     string
		fragmento string
		defecto   *skills.DefectoDeNorma
	}{
		{
			nombre:    "valida",
			documento: inicioDeLasNormas + normaDeLaLPAC + normaDeLaLCSP + normaDeLaLRBRL,
			normas: []skills.Norma{
				{
					Identificador: "BOE-A-2015-10565",
					Titulo:        tituloDeLaLPAC,
					Rango:         "Ley",
					Abreviatura:   "LPAC",
					Materias:      []string{"procedimiento administrativo"},
				},
				{
					Identificador: "BOE-A-2017-12902",
					Titulo:        tituloDeLaLCSP,
					Rango:         "Ley",
					Abreviatura:   "LCSP",
					Materias:      []string{"contratación pública"},
				},
				{
					Identificador: "BOE-A-1985-5392",
					Titulo:        tituloDeLaLRBRL,
					Rango:         "Ley",
					Materias:      []string{"régimen local"},
				},
			},
		},
		{
			// La marca es opcional: la lleva la LPAC, la LRBRL la escribe falsa y la
			// LCSP no la escribe, que es lo mismo.
			nombre: "con-vertebral",
			documento: inicioDeLasNormas +
				cambiada(t, normaDeLaLPAC, abreviaturaDeLaLPAC, abreviaturaDeLaLPAC+"    vertebral: true\n") +
				normaDeLaLCSP +
				normaDeLaLRBRL + "    vertebral: false\n",
			normas: []skills.Norma{
				{
					Identificador: "BOE-A-2015-10565",
					Titulo:        tituloDeLaLPAC,
					Rango:         "Ley",
					Abreviatura:   "LPAC",
					Materias:      []string{"procedimiento administrativo"},
					Vertebral:     true,
				},
				{
					Identificador: "BOE-A-2017-12902",
					Titulo:        tituloDeLaLCSP,
					Rango:         "Ley",
					Abreviatura:   "LCSP",
					Materias:      []string{"contratación pública"},
				},
				{
					Identificador: "BOE-A-1985-5392",
					Titulo:        tituloDeLaLRBRL,
					Rango:         "Ley",
					Materias:      []string{"régimen local"},
				},
			},
		},
		{
			nombre: "vertebral-que-no-es-booleano",
			documento: inicioDeLasNormas +
				cambiada(t, normaDeLaLPAC, abreviaturaDeLaLPAC, abreviaturaDeLaLPAC+"    vertebral: \"sí\"\n"),
			error: "BOE-A-2015-10565: vertebral: got string, want boolean",
		},
		{
			nombre: "con-vertical",
			documento: inicioDeLasNormas +
				cambiada(t, normaDeLaLPAC, abreviaturaDeLaLPAC, abreviaturaDeLaLPAC+"    vertical: fiscal\n"),
			error:   "BOE-A-2015-10565: campo no declarado: vertical",
			defecto: &skills.DefectoDeNorma{Norma: "BOE-A-2015-10565", Defecto: "campo no declarado: vertical"},
		},
		{
			nombre: "con-campo-no-declarado",
			documento: inicioDeLasNormas +
				cambiada(t, normaDeLaLPAC, abreviaturaDeLaLPAC, abreviaturaDeLaLPAC+"    estado: vigente\n"),
			error: "BOE-A-2015-10565: campo no declarado: estado",
		},
		{
			nombre:    "sin-titulo",
			documento: inicioDeLasNormas + cambiada(t, normaDeLaLPAC, tituloEscritoLPAC, ""),
			error:     "BOE-A-2015-10565: falta titulo",
			defecto:   &skills.DefectoDeNorma{Norma: "BOE-A-2015-10565", Defecto: "falta titulo"},
		},
		{
			nombre:    "sin-rango",
			documento: inicioDeLasNormas + cambiada(t, normaDeLaLPAC, rangoDeLaLPAC, ""),
			error:     "BOE-A-2015-10565: falta rango",
		},
		{
			nombre:    "sin-materias",
			documento: inicioDeLasNormas + cambiada(t, normaDeLaLPAC, materiasDeLaLPAC, ""),
			error:     "BOE-A-2015-10565: falta materias",
		},
		{
			nombre:    "materias-vacia",
			documento: inicioDeLasNormas + cambiada(t, normaDeLaLPAC, materiasDeLaLPAC, "    materias: []\n"),
			error:     "BOE-A-2015-10565: materias vacía",
			defecto:   &skills.DefectoDeNorma{Norma: "BOE-A-2015-10565", Defecto: "materias vacía"},
		},
		{
			nombre:    "identificador-con-otra-forma",
			documento: inicioDeLasNormas + cambiada(t, normaDeLaLPAC, claveDeLaLPAC, "  BOE-A-15-10565:\n"),
			error:     "BOE-A-15-10565: identificador con otra forma",
			defecto:   &skills.DefectoDeNorma{Norma: "BOE-A-15-10565", Defecto: "identificador con otra forma"},
		},
		{
			// El nombre se escribe entre comillas: sin ellas, el espacio no se vería.
			nombre:    "identificador-con-un-espacio",
			documento: inicioDeLasNormas + cambiada(t, normaDeLaLPAC, claveDeLaLPAC, "  \"BOE-A-2015-10565 \":\n"),
			error:     "\"BOE-A-2015-10565 \": identificador con otra forma",
			defecto:   &skills.DefectoDeNorma{Norma: "BOE-A-2015-10565 ", Defecto: "identificador con otra forma"},
		},
		{
			nombre: "identificador-repetido-con-titulos-distintos",
			documento: inicioDeLasNormas + normaDeLaLPAC + normaDeLaLRBRL +
				cambiada(t, normaDeLaLPAC, "de 1 de octubre", "de 2 de octubre"),
			error:   "BOE-A-2015-10565: repetido en las líneas 2 y 13",
			defecto: &skills.DefectoDeNorma{Norma: "BOE-A-2015-10565", Defecto: "repetido en las líneas 2 y 13"},
		},
		{
			// El recorrido del lector común compara las claves como las escribe el
			// documento, y un alias no se escribe como el identificador al que
			// apunta: sin la comprobación de LeerNormas, la segunda entrada
			// sustituiría a la primera.
			nombre: "identificador-repetido-por-un-alias",
			documento: inicioDeLasNormas +
				cambiada(t, normaDeLaLPAC, abreviaturaDeLaLPAC, "    abreviatura: &identificador BOE-A-2015-10565\n") +
				cambiada(t, normaDeLaLRBRL, "  BOE-A-1985-5392:\n", "  *identificador :\n"),
			error:   "BOE-A-2015-10565: repetido en las líneas 2 y 8",
			defecto: &skills.DefectoDeNorma{Norma: "BOE-A-2015-10565", Defecto: "repetido en las líneas 2 y 8"},
		},
		{
			nombre: "campo-repetido-en-una-norma",
			documento: inicioDeLasNormas +
				cambiada(t, normaDeLaLPAC, rangoDeLaLPAC, "    titulo: \"Ley 39/2015\"\n"+rangoDeLaLPAC),
			error: "BOE-A-2015-10565: titulo repetido en las líneas 3 y 4",
		},
		{
			// El recorrido del lector común no ve la repetición, porque el alias no
			// se escribe como titulo, y el esquema valida el documento convertido,
			// con el segundo título; la norma, al leerla, tiene titulo dos veces.
			nombre: "campo-repetido-en-una-norma-por-un-alias",
			documento: inicioDeLasNormas + cambiada(t,
				cambiada(t, normaDeLaLPAC, tituloEscritoLPAC, "    &titulo "+strings.TrimLeft(tituloEscritoLPAC, " ")),
				materiaDeLaLPAC, materiaDeLaLPAC+"    *titulo : \"Ley 39/2015\"\n"),
			error: "el documento YAML no se puede leer como skills.tablaDeNormas: BOE-A-2015-10565: " +
				"la norma no se puede leer: yaml: unmarshal errors:\n  line 8: field titulo already set in type skills.Norma",
		},
		{
			nombre:    "titulo-sin-texto",
			documento: inicioDeLasNormas + cambiada(t, normaDeLaLPAC, tituloEscritoLPAC, "    titulo: \"\"\n"),
			error:     "BOE-A-2015-10565: titulo sin texto",
		},
		{
			nombre:    "titulo-que-no-es-texto",
			documento: inicioDeLasNormas + cambiada(t, normaDeLaLPAC, tituloEscritoLPAC, "    titulo: 39\n"),
			error:     "BOE-A-2015-10565: titulo: got number, want string",
		},
		{
			nombre: "rango-no-admitido",
			// «Bando» es un acto del alcalde: nunca aparece como rango en la
			// legislación consolidada del BOE, así que ninguna grabación futura
			// puede meterlo en el enum y volver a dejar este caso sin error.
			documento: inicioDeLasNormas + cambiada(t, normaDeLaLPAC, rangoDeLaLPAC, "    rango: Bando\n"),
			error:     "BOE-A-2015-10565: rango no admitido: Bando",
		},
		{
			nombre:    "materia-sin-texto",
			documento: inicioDeLasNormas + cambiada(t, normaDeLaLPAC, materiaDeLaLPAC, "      - \"\"\n"),
			error:     "BOE-A-2015-10565: materias/0 sin texto",
		},
		{
			nombre:    "materias-repetidas",
			documento: inicioDeLasNormas + cambiada(t, normaDeLaLPAC, materiaDeLaLPAC, materiaDeLaLPAC+materiaDeLaLPAC),
			error:     "BOE-A-2015-10565: materias/0 y materias/1 repetidas",
		},
		{
			nombre: "defectos-en-dos-normas-en-orden-de-linea",
			documento: inicioDeLasNormas +
				cambiada(t, normaDeLaLPAC, abreviaturaDeLaLPAC, abreviaturaDeLaLPAC+"    vertical: fiscal\n") +
				cambiada(t, normaDeLaLRBRL, rangoDeLaLPAC, ""),
			error: "BOE-A-2015-10565: campo no declarado: vertical\nBOE-A-1985-5392: falta rango",
		},
		{
			// Sin línea ni campo que nombrar: el defecto va con el texto del lector.
			nombre:    "documento-vacio",
			documento: "",
			error:     "got null, want object",
		},
		{
			nombre:    "sin-normas",
			documento: "otras: 1\n",
			error:     "línea 1: campo no declarado: otras\nlínea 1: falta normas",
		},
		{
			nombre:    "normas-vacia",
			documento: "normas: {}\n",
			error:     "línea 1: normas vacía",
		},
		{
			nombre:    "normas-repetida",
			documento: inicioDeLasNormas + normaDeLaLPAC + inicioDeLasNormas + normaDeLaLRBRL,
			error:     "normas repetido en las líneas 1 y 8",
		},
		{
			nombre:    "clave-que-no-es-texto",
			documento: inicioDeLasNormas + cambiada(t, normaDeLaLPAC, claveDeLaLPAC, "  1:\n"),
			error:     "línea 2: normas: mapa con claves que no son texto",
		},
		{
			// La etiqueta de texto no hace texto de una lista: la conversión a un mapa
			// de texto falla antes de leer ninguna norma, así que a cada norma solo
			// llega una clave que es un escalar de texto, o un alias de uno.
			nombre:    "clave-de-texto-que-es-una-lista",
			documento: inicioDeLasNormas + cambiada(t, normaDeLaLPAC, claveDeLaLPAC, "  ? !!str [BOE-A-2015-10565]\n  :\n"),
			error: "el documento YAML no se puede leer: yaml: unmarshal errors:\n" +
				"  line 2: cannot unmarshal !!str `` into string",
		},
		{
			// Una fusión trae normas sin escribir su identificador en la tabla: el
			// esquema la acepta, porque valida el resultado de fusionar, y
			// LeerNormas la rechaza.
			nombre: "clave-de-fusion",
			documento: inicioDeLasNormas +
				"  <<:\n" +
				"    BOE-A-1985-5392:\n" +
				"      titulo: \"" + tituloDeLaLRBRL + "\"\n" +
				"      rango: Ley\n" +
				"      materias:\n" +
				"        - régimen local\n" +
				normaDeLaLPAC,
			error: "línea 2: normas: clave de fusión << no admitida: cada norma se escribe con su identificador",
		},
		{
			nombre:    "yaml-mal-formado",
			documento: inicioDeLasNormas + "  BOE-A-2015-10565: [\n",
			fragmento: "no es YAML válido: ",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			normas, err := skills.LeerNormas([]byte(caso.documento))
			if caso.error == "" && caso.fragmento == "" {
				require.NoError(t, err)
				assert.Equal(t, caso.normas, normas)

				return
			}

			require.Error(t, err)
			assert.Nil(t, normas, "una tabla con defectos no se entrega a medias")

			if caso.error != "" {
				require.EqualError(t, err, caso.error)
			}

			if caso.fragmento != "" {
				require.ErrorContains(t, err, caso.fragmento)
			}

			if caso.defecto != nil {
				var defecto *skills.DefectoDeNorma
				require.ErrorAs(t, err, &defecto)
				assert.Equal(t, caso.defecto, defecto)
			}
		})
	}
}

// cambiada es el texto con la primera aparición de viejo sustituida por nuevo.
// viejo tiene que estar en el texto: un caso que no cambiara nada sería la
// tabla válida con otro nombre.
func cambiada(t *testing.T, texto, viejo, nuevo string) string {
	t.Helper()

	require.Contains(t, texto, viejo)

	return strings.Replace(texto, viejo, nuevo, 1)
}

// tablaDeNormasDelRepositorio es data/normas.yaml, relativo al directorio de
// este paquete, que es donde go test ejecuta sus tests (research.md V46).
const tablaDeNormasDelRepositorio = "../../data/normas.yaml"

// leyesVertebrales son los identificadores de las quince leyes de la tabla de
// leyes vertebrales de refs/mapa-sistema-legal-skills.md §1.4, en su orden
// (FR-070): la Constitución, el Código Civil, la LPAC, la LRJSP, la LJCA, la LEC,
// la LOPJ, la LRBRL, el TRLRHL, la LCSP, la LGS, la LTAIBG, la LGT, la LOPDGDD y
// la Ley General Presupuestaria. TestIdentificadoresDeLasNormas ata cada uno a su
// búsqueda grabada del BOE (FR-071, FR-072).
var leyesVertebrales = []string{
	"BOE-A-1978-31229",
	"BOE-A-1889-4763",
	"BOE-A-2015-10565",
	"BOE-A-2015-10566",
	"BOE-A-1998-16718",
	"BOE-A-2000-323",
	"BOE-A-1985-12666",
	"BOE-A-1985-5392",
	"BOE-A-2004-4214",
	"BOE-A-2017-12902",
	"BOE-A-2003-20977",
	"BOE-A-2013-12887",
	"BOE-A-2003-23186",
	"BOE-A-2018-16673",
	"BOE-A-2003-21614",
}

// TestNormasDelRepositorio comprueba que la tabla de normas del repositorio es
// válida (US6, escenario 1; FR-021, FR-043): LeerNormas la lee sin ningún
// defecto, sin campo vertical ni ningún otro que el esquema no declare, y con
// alguna norma; y la marca vertebral la llevan exactamente las quince leyes de la
// tabla de leyes vertebrales, ninguna más ni ninguna menos (FR-067, FR-070,
// SC-010).
func TestNormasDelRepositorio(t *testing.T) {
	t.Parallel()

	contenido, err := os.ReadFile(tablaDeNormasDelRepositorio)
	require.NoError(t, err, "la tabla de normas del repositorio no se puede leer")

	normas, err := skills.LeerNormas(contenido)
	require.NoError(t, err, "la tabla de normas del repositorio, %s", tablaDeNormasDelRepositorio)
	assert.NotEmpty(t, normas, "la tabla de normas del repositorio, %s, tiene normas", tablaDeNormasDelRepositorio)

	t.Run("vertebrales", func(t *testing.T) {
		t.Parallel()

		var marcadas []string

		for _, norma := range normas {
			if norma.Vertebral {
				marcadas = append(marcadas, norma.Identificador)
			}
		}

		assert.ElementsMatch(t, leyesVertebrales, marcadas,
			"las normas de %s con vertebral: true son las quince de la tabla de leyes vertebrales",
			tablaDeNormasDelRepositorio)
	})
}

// esquemaPublicadoDeNormas es el esquema de data/normas.yaml, relativo al
// directorio de este paquete, que es donde go test ejecuta sus tests
// (research.md V46): el mismo fichero con el que valida LeerNormas.
const esquemaPublicadoDeNormas = "../../schemas/normas.yaml.json"

// TestEsquemaDeNormas comprueba el esquema publicado de data/normas.yaml
// (contrato normas-y-referencias §2): compila con las aserciones de formato
// activas, y el enum de rango es exactamente el conjunto de rangos de los
// resultados de las búsquedas grabadas de H4 y de H5, sin valores repetidos, de
// modo que no admite un rango que la fuente no dio ni le falta uno que dio.
func TestEsquemaDeNormas(t *testing.T) {
	t.Parallel()

	contenido, err := os.ReadFile(esquemaPublicadoDeNormas)
	require.NoError(t, err)

	t.Run("compila", func(t *testing.T) {
		t.Parallel()

		esquema, err := skills.CompilarEsquema(contenido)
		require.NoError(t, err)
		assert.NotNil(t, esquema)
	})

	t.Run("rangos-grabados", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, rangosDeLasBusquedasGrabadas(t, evals.UnionDeGrabaciones()), rangosDelEsquema(t, contenido),
			"el enum de rango de %s es el conjunto de rangos de las búsquedas grabadas", esquemaPublicadoDeNormas)
	})
}

// TestCompilarEsquemaDeNormasDesdeUnaRuta fija la lectura y la compilación del
// esquema de data/normas.yaml desde su ruta: la del esquema publicado da el
// esquema; una carpeta en su lugar es un error de lectura, y un JSON que no es un
// JSON Schema válido, uno de compilación que nombra la ruta.
func TestCompilarEsquemaDeNormasDesdeUnaRuta(t *testing.T) {
	t.Parallel()

	esquema, err := skills.CompilarEsquemaDeNormas(esquemaPublicadoDeNormas)
	require.NoError(t, err)
	assert.NotNil(t, esquema)

	directorio := t.TempDir()

	carpeta := filepath.Join(directorio, "carpeta.json")
	require.NoError(t, os.Mkdir(carpeta, 0o750))

	esquema, err = skills.CompilarEsquemaDeNormas(carpeta)
	require.ErrorIs(t, err, syscall.EISDIR)
	require.ErrorContains(t, err, "no se puede leer el esquema de data/normas.yaml: ")
	assert.Nil(t, esquema)

	sinCompilar := filepath.Join(directorio, "rango-sin-tipo.json")
	escribirFicheroDePrueba(t, sinCompilar,
		`{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {"rango": {"type": "texto"}}}`)

	esquema, err = skills.CompilarEsquemaDeNormas(sinCompilar)
	require.ErrorContains(t, err, "el esquema de data/normas.yaml "+sinCompilar+": el esquema no compila: ")
	assert.Nil(t, esquema)
}

// rangosDelEsquema son los valores del enum de rango del esquema de normas,
// ordenados; tienen que ser textos, al menos uno y sin repetir.
func rangosDelEsquema(t *testing.T, contenido []byte) []string {
	t.Helper()

	var esquema struct {
		Defs struct {
			Norma struct {
				Properties struct {
					Rango struct {
						Enum []string `json:"enum"`
					} `json:"rango"`
				} `json:"properties"`
			} `json:"norma"`
		} `json:"$defs"`
	}
	require.NoError(t, json.Unmarshal(contenido, &esquema), "el enum de rango del esquema de normas son textos")

	enum := esquema.Defs.Norma.Properties.Rango.Enum
	require.NotEmpty(t, enum, "el esquema de normas declara el enum de rango")

	ordenados := slices.Sorted(slices.Values(enum))
	require.Len(t, slices.Compact(slices.Clone(ordenados)), len(enum), "el enum de rango no repite valores: %q", enum)

	return ordenados
}

// Lo que rangosDeLasBusquedasGrabadas reconoce en una grabación: la ruta y el
// parámetro de la dirección de una búsqueda de la fuente, y el estado con el que
// se grabó.
const (
	rutaDeLaBusqueda           = "/datosabiertos/api/legislacion-consolidada"
	parametroDeLaConsulta      = "query"
	estadoDeUnaBusquedaGrabada = 200
)

// rangosDeLasBusquedasGrabadas es el conjunto ordenado de los rangos de todos
// los resultados de las búsquedas grabadas en los conjuntos. Lee con el
// decodificador JSON genérico, sin el código de lectura de la fuente, el cuerpo
// que httpx.Replay sirve para la dirección de cada búsqueda de cada conjunto;
// cada conjunto tiene al menos una búsqueda, y cada resultado, un rango con su
// texto: nada que no se entienda se salta.
func rangosDeLasBusquedasGrabadas(t *testing.T, conjuntos []string) []string {
	t.Helper()

	rangos := map[string]struct{}{}

	for _, conjunto := range conjuntos {
		cliente, err := httpx.Replay(conjunto, httpx.ConFuente(boe.NombreDeLaFuente))
		require.NoError(t, err)

		busquedas := direccionesDeLasBusquedas(t, conjunto)
		require.NotEmpty(t, busquedas, "el conjunto de grabaciones %s tiene alguna búsqueda", conjunto)

		for _, direccion := range busquedas {
			for _, rango := range rangosDeLaBusqueda(t, cliente, direccion) {
				rangos[rango] = struct{}{}
			}
		}
	}

	return slices.Sorted(maps.Keys(rangos))
}

// direccionesDeLasBusquedas son las direcciones grabadas de las peticiones de
// búsqueda de un conjunto de grabaciones, en el orden de sus ficheros.
func direccionesDeLasBusquedas(t *testing.T, conjunto string) []string {
	t.Helper()

	entradas, err := os.ReadDir(conjunto)
	require.NoError(t, err, "el conjunto de grabaciones %s no se puede listar", conjunto)

	var direcciones []string

	for _, entrada := range entradas {
		contenido, err := os.ReadFile(filepath.Clean(filepath.Join(conjunto, entrada.Name())))
		require.NoError(t, err)

		var grabacion struct {
			Peticion struct {
				URL string `json:"url"`
			} `json:"peticion"`
		}
		require.NoError(t, json.Unmarshal(contenido, &grabacion), "la grabación %s", entrada.Name())
		require.NotEmpty(t, grabacion.Peticion.URL, "la grabación %s lleva la dirección de su petición", entrada.Name())

		direccion, err := url.Parse(grabacion.Peticion.URL)
		require.NoError(t, err, "la grabación %s", entrada.Name())

		if direccion.Path == rutaDeLaBusqueda && direccion.Query().Has(parametroDeLaConsulta) {
			direcciones = append(direcciones, grabacion.Peticion.URL)
		}
	}

	return direcciones
}

// rangosDeLaBusqueda son los textos de rango de los resultados de una búsqueda
// grabada, en su orden: data vacío no tiene ninguno, y data puede ser la lista de
// resultados o un resultado suelto.
func rangosDeLaBusqueda(t *testing.T, cliente *httpx.Cliente, direccion string) []string {
	t.Helper()

	respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{}, httpx.Peticion{
		Metodo: "GET",
		URL:    direccion,
		Acepta: "application/json",
	})
	require.NoError(t, err, "la búsqueda grabada %s", direccion)
	require.Equal(t, estadoDeUnaBusquedaGrabada, respuesta.Estado, "la búsqueda grabada %s", direccion)

	var cuerpo struct {
		Data any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(respuesta.Cuerpo, &cuerpo), "la búsqueda grabada %s", direccion)

	var resultados []any

	switch data := cuerpo.Data.(type) {
	case string:
		require.Empty(t, data, "el data de texto de la búsqueda grabada %s es el vacío, sin resultados", direccion)
	case []any:
		resultados = data
	case map[string]any:
		resultados = []any{data}
	default:
		require.Failf(t, "data con otra forma", "el data de la búsqueda grabada %s es %T", direccion, data)
	}

	rangos := make([]string, 0, len(resultados))

	for posicion, resultado := range resultados {
		objeto, esObjeto := resultado.(map[string]any)
		require.True(t, esObjeto, "el resultado %d de la búsqueda grabada %s es un objeto", posicion, direccion)

		rango, esObjeto := objeto["rango"].(map[string]any)
		require.True(t, esObjeto, "el rango del resultado %d de la búsqueda grabada %s es un objeto", posicion, direccion)

		texto, esTexto := rango["texto"].(string)
		require.True(t, esTexto, "el rango del resultado %d de la búsqueda grabada %s tiene texto", posicion, direccion)
		require.NotEmpty(t, texto, "el rango del resultado %d de la búsqueda grabada %s tiene texto", posicion, direccion)

		rangos = append(rangos, texto)
	}

	return rangos
}
