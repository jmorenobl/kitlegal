package skills_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"unicode"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/ids"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/core/territorio"
	"github.com/jmorenobl/kitlegal/internal/skills"
)

// El territorio sintético de TestLeerTerritorio y del subtest gramaticas: dos
// municipios, uno de ellos con su DIR3, el boletín estatal y dos comunidades,
// una configurada entera y otra foral sin ningún boletín. Todo es inventado,
// como en los tests del dominio: los nombres y los boletines no son de ningún
// sitio y los códigos no están en la relación congelada.
const (
	municipiosDePrueba = "fecha: \"2026-02-04\"\n" +
		"source: prueba.relacion\n" +
		"municipios:\n" +
		"  \"28991\": {dc: \"5\", nombre: \"Villaprueba\", provincia: \"28\", comunidad: \"01\"}\n" +
		"  \"31991\": {dc: \"3\", nombre: \"Iruñeta/Pamploneta\", provincia: \"31\", comunidad: \"02\"}\n"

	dir3DePrueba = "fecha: \"2026-09-21\"\n" +
		"source: prueba.correspondencia\n" +
		"correspondencia:\n" +
		"  \"28991\": \"L01289915\"\n"

	estadoDePrueba = "fecha: \"2026-09-20\"\n" +
		"source: prueba.estado\n" +
		"boletin:\n" +
		"  codigo: BOEP\n" +
		"  nombre: \"Boletín Oficial del Estado de Prueba\"\n" +
		"  url: \"https://estado.example/\"\n"

	comunidadConfiguradaDePrueba = "fecha: \"2026-09-20\"\n" +
		"source: prueba.nombres\n" +
		"codigo: \"01\"\n" +
		"nombre: \"Comunidad Uniprovincial\"\n" +
		"regimen: comun\n" +
		"provincias:\n" +
		"  \"28\": \"Provincia Única\"\n" +
		"boletines:\n" +
		"  autonomico: {codigo: BOCU, nombre: \"Boletín de la Uniprovincial\", url: \"https://uniprovincial.example/\"}\n" +
		"  provincial: {codigo: BOCU, nombre: \"Boletín de la Uniprovincial\", url: \"https://uniprovincial.example/\", " +
		"motivo: \"Comunidad uniprovincial: su boletín hace también de provincial.\"}\n"

	comunidadForalDePrueba = "fecha: \"2026-09-20\"\n" +
		"source: prueba.nombres\n" +
		"codigo: \"02\"\n" +
		"nombre: \"Comunidad Foral\"\n" +
		"regimen: foral\n" +
		"provincias:\n" +
		"  \"31\": \"Provincia Foral\"\n"
)

// fuentesDePrueba son las fuentes del territorio sintético, nuevas en cada
// llamada, para que cada caso cambie las suyas sin tocar las de los demás.
func fuentesDePrueba() territorio.Fuentes {
	return territorio.Fuentes{
		Municipios: []byte(municipiosDePrueba),
		DIR3:       []byte(dir3DePrueba),
		Estado:     []byte(estadoDePrueba),
		Comunidades: map[string][]byte{
			"01": []byte(comunidadConfiguradaDePrueba),
			"02": []byte(comunidadForalDePrueba),
		},
	}
}

// ficherosDePrueba es lo que LeerTerritorio lee de fuentesDePrueba.
func ficherosDePrueba() territorio.Ficheros {
	boletinUniprovincial := territorio.BoletinConfigurado{
		Codigo: "BOCU",
		Nombre: "Boletín de la Uniprovincial",
		URL:    "https://uniprovincial.example/",
	}
	provincial := boletinUniprovincial
	provincial.Motivo = "Comunidad uniprovincial: su boletín hace también de provincial."

	return territorio.Ficheros{
		Municipios: territorio.FicheroDeMunicipios{
			Fecha:  "2026-02-04",
			Source: "prueba.relacion",
			Municipios: map[string]territorio.FilaDeMunicipio{
				"28991": {DC: "5", Nombre: "Villaprueba", Provincia: "28", Comunidad: "01"},
				"31991": {DC: "3", Nombre: "Iruñeta/Pamploneta", Provincia: "31", Comunidad: "02"},
			},
		},
		DIR3: territorio.FicheroDeDIR3{
			Fecha:           "2026-09-21",
			Source:          "prueba.correspondencia",
			Correspondencia: map[string]string{"28991": "L01289915"},
		},
		Estado: territorio.FicheroDeEstado{
			Fecha:  "2026-09-20",
			Source: "prueba.estado",
			Boletin: territorio.BoletinDelEstado{
				Codigo: "BOEP",
				Nombre: "Boletín Oficial del Estado de Prueba",
				URL:    "https://estado.example/",
			},
		},
		Comunidades: map[string]territorio.FicheroDeComunidad{
			"01": {
				Fecha:      "2026-09-20",
				Source:     "prueba.nombres",
				Codigo:     "01",
				Nombre:     "Comunidad Uniprovincial",
				Regimen:    "comun",
				Provincias: map[string]string{"28": "Provincia Única"},
				Boletines: &territorio.BoletinesConfigurados{
					Autonomico: &boletinUniprovincial,
					Provincial: &provincial,
				},
			},
			"02": {
				Fecha:      "2026-09-20",
				Source:     "prueba.nombres",
				Codigo:     "02",
				Nombre:     "Comunidad Foral",
				Regimen:    "foral",
				Provincias: map[string]string{"31": "Provincia Foral"},
			},
		},
	}
}

// TestLeerTerritorio fija LeerTerritorio sobre el territorio sintético
// (contrato de datos §2; FR-044): los cuatro documentos válidos se leen en los
// tipos del dominio, la comunidad sin boletines con Boletines a nil; y cada
// defecto —una clave que el esquema no declara, un código que no casa con su
// patrón, una clave repetida— es un error que nombra su fichero y el punto del
// documento, y con él no se entrega nada. Los defectos de varios ficheros van
// todos, en el orden de los ficheros.
func TestLeerTerritorio(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		cambiar func(t *testing.T, fuentes *territorio.Fuentes)
		error   string
		// repetida es la clave repetida que el error lleva dentro, si la hay.
		repetida *skills.ClaveRepetida
	}{
		{
			nombre:  "valido",
			cambiar: func(*testing.T, *territorio.Fuentes) {},
		},
		{
			nombre: "clave-desconocida",
			cambiar: func(t *testing.T, fuentes *territorio.Fuentes) {
				t.Helper()

				fuentes.Estado = []byte(cambiada(t, estadoDePrueba, "  url:", "  sede: Capital\n  url:"))
			},
			error: "data/territorio/estado.yaml: boletin, línea 4: additional properties 'sede' not allowed",
		},
		{
			nombre: "clave-desconocida-en-una-comunidad",
			cambiar: func(t *testing.T, fuentes *territorio.Fuentes) {
				t.Helper()

				fuentes.Comunidades["02"] = []byte(cambiada(t, comunidadForalDePrueba, "regimen:", "capital: Ninguna\nregimen:"))
			},
			error: "data/territorio/comunidades/02.yaml: línea 1: additional properties 'capital' not allowed",
		},
		{
			nombre: "patron",
			cambiar: func(t *testing.T, fuentes *territorio.Fuentes) {
				t.Helper()

				fuentes.Municipios = []byte(cambiada(t, municipiosDePrueba, "\"31991\"", "\"3199\""))
			},
			error: "data/territorio/municipios.yaml: municipios, línea 4: invalid propertyName '3199': " +
				"'3199' does not match pattern '^(0[1-9]|[1-4][0-9]|5[0-2])(00[1-9]|0[1-9][0-9]|[1-9][0-9]{2})$'",
		},
		{
			nombre: "clave-repetida",
			cambiar: func(t *testing.T, fuentes *territorio.Fuentes) {
				t.Helper()

				fuentes.DIR3 = []byte(dir3DePrueba + "  \"28991\": \"L01289915\"\n")
			},
			error:    "data/territorio/dir3.yaml: correspondencia: 28991 repetido en las líneas 4 y 5",
			repetida: &skills.ClaveRepetida{Mapa: []string{"correspondencia"}, Clave: "28991", Lineas: [2]int{4, 5}},
		},
		{
			nombre: "defectos-en-dos-ficheros",
			cambiar: func(t *testing.T, fuentes *territorio.Fuentes) {
				t.Helper()

				fuentes.DIR3 = []byte(cambiada(t, dir3DePrueba, "\"L01289915\"", "\"X01289915\""))
				fuentes.Comunidades["01"] = []byte(cambiada(t, comunidadConfiguradaDePrueba, "regimen: comun", "regimen: otro"))
			},
			error: "data/territorio/dir3.yaml: correspondencia/28991, línea 4: " +
				"'X01289915' does not match pattern " +
				"'^[Ll]01(0[1-9]|[1-4][0-9]|5[0-2])(00[1-9]|0[1-9][0-9]|[1-9][0-9]{2})[0-9]$'\n" +
				"data/territorio/comunidades/01.yaml: regimen, línea 5: value must be one of 'comun', 'foral'",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			fuentes := fuentesDePrueba()
			caso.cambiar(t, &fuentes)

			ficheros, err := skills.LeerTerritorio(fuentes)
			if caso.error == "" {
				require.NoError(t, err)
				assert.Equal(t, ficherosDePrueba(), ficheros)

				return
			}

			require.EqualError(t, err, caso.error)
			assert.Zero(t, ficheros, "unos ficheros con defectos no se entregan a medias")

			if caso.repetida != nil {
				var repetida *skills.ClaveRepetida
				require.ErrorAs(t, err, &repetida)
				assert.Equal(t, caso.repetida, repetida)
			}
		})
	}
}

// TestCompilarEsquemaDelTerritorioDesdeUnaRuta fija la lectura y la compilación
// del esquema de un fichero de data/territorio/ desde su ruta: la del esquema
// publicado da el esquema; una carpeta en su lugar es un error de lectura, y un
// JSON que no es un JSON Schema válido, uno de compilación que nombra la ruta.
func TestCompilarEsquemaDelTerritorioDesdeUnaRuta(t *testing.T) {
	t.Parallel()

	comprobarCompilarEsquemaDesdeUnaRuta(t, skills.CompilarEsquemaDelTerritorio,
		"../../schemas/territorio-estado.yaml.json",
		"no se puede leer un esquema de data/territorio/: ", "el esquema de data/territorio/ ", "fecha")
}

// Los ficheros congelados de data/territorio/ y la tabla de fuentes, relativos
// al directorio de este paquete, que es donde go test ejecuta sus tests
// (research.md V42). Se leen por su ruta, sin importar el paquete data, que solo
// importan el registro de applets y el binario de e2e (contrato de datos §3).
const (
	arbolDelRepositorio       = "../.."
	municipiosDelRepositorio  = "../../data/territorio/municipios.yaml"
	dir3DelRepositorio        = "../../data/territorio/dir3.yaml"
	estadoDelRepositorio      = "../../data/territorio/estado.yaml"
	comunidadesDelRepositorio = "../../data/territorio/comunidades"
	fuentesDocumentadas       = "../../docs/SOURCES.md"

	// carpetaDelTerritorio es donde viven los ficheros congelados, relativa a
	// la raíz del repositorio: el source de un dato que fija la configuración
	// es la ruta de uno de ellos (data-model §2.5).
	carpetaDelTerritorio = "data/territorio/"

	// extensionDeComunidad es la de cada fichero de comunidad, <código>.yaml.
	extensionDeComunidad = ".yaml"

	// comunidadesYCiudades son las comunidades y ciudades autónomas: cada una
	// tiene su fichero y su régimen (FR-055).
	comunidadesYCiudades = 19
)

// Lo único configurado en este hito, tal como lo escribe su fichero congelado:
// la Comunidad de Madrid, con el BOCM de boletín autonómico y también de
// provincial (FR-051, FR-052, SC-005).
const (
	codigoDeLaConfigurada  = "13"
	nombreDeLaConfigurada  = "Comunidad de Madrid"
	boletinDeLaConfigurada = "BOCM"
)

// regimenes es el vocabulario cerrado del régimen de una comunidad
// (data-model §2.2).
var regimenes = []string{"comun", "foral"}

// TestTerritorioDelRepositorio comprueba los ficheros congelados de
// data/territorio/, los que el binario lleva dentro (contrato de datos §2 y §6;
// FR-044, FR-056; research.md D27): validan contra sus esquemas, son coherentes
// entre sí, cada dato cita una procedencia que existe, solo la Comunidad de
// Madrid trae boletines, todo municipio tiene régimen, los patrones de los
// esquemas son las gramáticas de internal/core/ids, y el pliegue de nombres
// cubre el corpus sin dejar ningún municipio inalcanzable ni ningún nombre que
// se lea como un código. Todo subtest parte de los cuatro ficheros y las 19
// comunidades: sin uno de ellos el test falla antes, en vez de pasar en vacío.
func TestTerritorioDelRepositorio(t *testing.T) {
	t.Parallel()

	corpus := nuevoCorpus(fuentesDelRepositorio(t))

	subtests := []struct {
		nombre    string
		comprobar func(t *testing.T, corpus *corpusDelRepositorio)
	}{
		{"esquema", comprobarEsquema},
		{"integridad", comprobarIntegridad},
		{"fuentes", comprobarFuentes},
		{"madrid-configurada", comprobarMadridConfigurada},
		{"regimen-de-todas", comprobarRegimenDeTodas},
		{"gramaticas", func(t *testing.T, _ *corpusDelRepositorio) { t.Helper(); comprobarGramaticas(t) }},
		{"pliegue-cubre-el-corpus", comprobarPliegueCubreElCorpus},
		{"nombres-alcanzables", comprobarNombresAlcanzables},
		{"ningun-nombre-es-solo-cifras", comprobarNingunNombreEsSoloCifras},
	}

	for _, subtest := range subtests {
		t.Run(subtest.nombre, func(t *testing.T) {
			t.Parallel()

			subtest.comprobar(t, corpus)
		})
	}
}

// corpusDelRepositorio son los ficheros congelados del repositorio y lo que
// los subtests calculan con ellos, cada cosa una sola vez: leerlos contra sus
// esquemas, cargarlos y resolver por su código cada municipio de la relación.
type corpusDelRepositorio struct {
	leidos    func() (territorio.Ficheros, error)
	cargado   func() (*territorio.Registro, error)
	resueltos func() (map[string]territorio.Territorio, error)
}

// nuevoCorpus prepara los cálculos del corpus sobre las fuentes, sin hacer
// ninguno todavía.
func nuevoCorpus(fuentes territorio.Fuentes) *corpusDelRepositorio {
	corpus := &corpusDelRepositorio{
		leidos:  sync.OnceValues(func() (territorio.Ficheros, error) { return skills.LeerTerritorio(fuentes) }),
		cargado: sync.OnceValues(func() (*territorio.Registro, error) { return territorio.Cargar(fuentes) }),
	}
	corpus.resueltos = sync.OnceValues(corpus.resolverTodos)

	return corpus
}

// ficheros son los ficheros leídos contra sus esquemas; si no se leen, el
// subtest que los necesita falla con ellos.
func (c *corpusDelRepositorio) ficheros(t *testing.T) territorio.Ficheros {
	t.Helper()

	ficheros, err := c.leidos()
	require.NoError(t, err, "los ficheros congelados validan contra sus esquemas")

	return ficheros
}

// registro es el registro cargado de los ficheros; si no se carga, el subtest
// que lo necesita falla con él.
func (c *corpusDelRepositorio) registro(t *testing.T) *territorio.Registro {
	t.Helper()

	registro, err := c.cargado()
	require.NoError(t, err, "los ficheros congelados son coherentes entre sí (data-model §2.1)")

	return registro
}

// territorios son los de todos los municipios de la relación, por código.
func (c *corpusDelRepositorio) territorios(t *testing.T) map[string]territorio.Territorio {
	t.Helper()

	territorios, err := c.resueltos()
	require.NoError(t, err, "todo municipio de la relación se resuelve por su código")

	return territorios
}

// resolverTodos resuelve por su código cada municipio de la relación.
func (c *corpusDelRepositorio) resolverTodos() (map[string]territorio.Territorio, error) {
	ficheros, err := c.leidos()
	if err != nil {
		return nil, err
	}

	registro, err := c.cargado()
	if err != nil {
		return nil, err
	}

	territorios := make(map[string]territorio.Territorio, len(ficheros.Municipios.Municipios))
	for codigo := range ficheros.Municipios.Municipios {
		resuelto, err := registro.Resolver(codigo)
		if err != nil {
			return nil, fmt.Errorf("el municipio %s no se resuelve por su código: %w", codigo, err)
		}

		territorios[codigo] = resuelto
	}

	return territorios, nil
}

// comprobarEsquema exige que los ficheros validen contra sus esquemas y que
// traigan lo que tienen que traer: municipios, filas verificadas y las 19
// comunidades (FR-044).
func comprobarEsquema(t *testing.T, corpus *corpusDelRepositorio) {
	t.Helper()

	ficheros := corpus.ficheros(t)
	assert.NotEmpty(t, ficheros.Municipios.Municipios, "la relación tiene municipios")
	assert.NotEmpty(t, ficheros.DIR3.Correspondencia, "la correspondencia tiene filas verificadas")
	assert.Len(t, ficheros.Comunidades, comunidadesYCiudades, "hay un fichero por comunidad y ciudad autónoma")
}

// comprobarIntegridad delega en territorio.Cargar la integridad entre ficheros
// (data-model §2.1): sin repetir aquí ni sus tipos ni sus reglas.
func comprobarIntegridad(t *testing.T, corpus *corpusDelRepositorio) {
	t.Helper()

	assert.NotNil(t, corpus.registro(t))
}

// comprobarFuentes exige que todo source —el que declara la raíz de cada
// fichero y el que emite el territorio de cada municipio— sea una fila de
// docs/SOURCES.md o la ruta de un fichero congelado que existe (data-model
// §2.5; FR-005).
func comprobarFuentes(t *testing.T, corpus *corpusDelRepositorio) {
	t.Helper()

	ficheros := corpus.ficheros(t)
	filas := filasDeSources(t)

	// De dónde sale cada source, para nombrarlo si no existe.
	origenes := map[string]string{
		ficheros.Municipios.Source: "la raíz de " + carpetaDelTerritorio + "municipios.yaml",
		ficheros.DIR3.Source:       "la raíz de " + carpetaDelTerritorio + "dir3.yaml",
		ficheros.Estado.Source:     "la raíz de " + carpetaDelTerritorio + "estado.yaml",
	}
	for _, codigo := range slices.Sorted(maps.Keys(ficheros.Comunidades)) {
		origenes[ficheros.Comunidades[codigo].Source] = "la raíz de la comunidad " + codigo
	}

	territorios := corpus.territorios(t)
	for _, codigo := range slices.Sorted(maps.Keys(territorios)) {
		for _, source := range sourcesEmitidos(territorios[codigo]) {
			if _, anotado := origenes[source]; !anotado {
				origenes[source] = "el territorio del municipio " + codigo
			}
		}
	}

	for _, source := range slices.Sorted(maps.Keys(origenes)) {
		assert.True(t, esProcedenciaConocida(filas, source),
			"el source %q de %s no es una fila de docs/SOURCES.md ni un fichero de %s", source,
			origenes[source], carpetaDelTerritorio)
	}
}

// comprobarMadridConfigurada exige que la Comunidad de Madrid esté configurada
// y declare el BOCM como autonómico y como provincial, con la razón escrita
// (FR-052). **No** exige que sea la única: configurar otra comunidad es
// rellenar los boletines de su fichero y nada más (SC-005, CONTRIBUTING), así
// que prohibirlo aquí haría falsa esa promesa. Que hoy solo esté Madrid
// (FR-051) es el estado del hito y se comprueba en su cierre (quickstart §9),
// no en un control permanente.
func comprobarMadridConfigurada(t *testing.T, corpus *corpusDelRepositorio) {
	t.Helper()

	ficheros := corpus.ficheros(t)

	var configuradas []string
	for _, codigo := range slices.Sorted(maps.Keys(ficheros.Comunidades)) {
		if ficheros.Comunidades[codigo].Boletines != nil {
			configuradas = append(configuradas, codigo)
		}
	}
	require.Contains(t, configuradas, codigoDeLaConfigurada,
		"la comunidad %s trae boletines (FR-052)", codigoDeLaConfigurada)

	configurada := ficheros.Comunidades[codigoDeLaConfigurada]
	assert.Equal(t, nombreDeLaConfigurada, configurada.Nombre)

	autonomico, provincial := configurada.Boletines.Autonomico, configurada.Boletines.Provincial
	require.NotNil(t, autonomico, "la comunidad %s configura su boletín autonómico", codigoDeLaConfigurada)
	require.NotNil(t, provincial, "la comunidad %s configura su boletín provincial", codigoDeLaConfigurada)
	assert.Equal(t, boletinDeLaConfigurada, autonomico.Codigo)
	assert.Equal(t, []string{autonomico.Codigo, autonomico.Nombre, autonomico.URL},
		[]string{provincial.Codigo, provincial.Nombre, provincial.URL},
		"el boletín autonómico hace también de provincial (FR-052)")
	assert.NotEmpty(t, provincial.Motivo, "la razón de la equivalencia consta en la configuración (FR-052)")
}

// comprobarRegimenDeTodas exige que cada una de las 19 comunidades declare su
// régimen y que todo municipio de la relación lo tenga, esté o no configurada
// su comunidad: es dato nacional (FR-055, SC-003).
func comprobarRegimenDeTodas(t *testing.T, corpus *corpusDelRepositorio) {
	t.Helper()

	ficheros := corpus.ficheros(t)
	for _, codigo := range slices.Sorted(maps.Keys(ficheros.Comunidades)) {
		assert.Contains(t, regimenes, ficheros.Comunidades[codigo].Regimen,
			"la comunidad %s declara su régimen (FR-055)", codigo)
	}

	territorios := corpus.territorios(t)

	var sinRegimen []string
	for _, codigo := range slices.Sorted(maps.Keys(territorios)) {
		if !slices.Contains(regimenes, territorios[codigo].Regimen.Valor) {
			sinRegimen = append(sinRegimen, codigo)
		}
	}
	assert.Empty(t, sinRegimen, "todo municipio de la relación tiene régimen, esté o no configurada su comunidad")
}

// comprobarPliegueCubreElCorpus exige que toda runa de todo nombre del corpus
// —municipios, provincias y comunidades— la cubra la tabla del pliegue: una
// runa nueva hace fallar make ci en vez de pasar en silencio (research.md D10).
func comprobarPliegueCubreElCorpus(t *testing.T, corpus *corpusDelRepositorio) {
	t.Helper()

	var fuera []string
	for _, nombre := range nombresDelCorpus(corpus.ficheros(t)) {
		plegado := territorio.Plegar(nombre.texto)
		if ajenas := runasFueraDelPliegue(plegado); ajenas != "" {
			fuera = append(fuera, fmt.Sprintf("%s, %q, se pliega a %q, con %q fuera del alfabeto del pliegue",
				nombre.de, nombre.texto, plegado, ajenas))
		}
	}
	assert.Empty(t, fuera, "toda runa de todo nombre del corpus está cubierta por el pliegue")
}

// comprobarNombresAlcanzables exige que ningún municipio quede inalcanzable
// por su nombre oficial: o se resuelve él, o la ambigüedad lo nombra entre sus
// candidatos (research.md D10).
func comprobarNombresAlcanzables(t *testing.T, corpus *corpusDelRepositorio) {
	t.Helper()

	ficheros := corpus.ficheros(t)
	registro := corpus.registro(t)

	var inalcanzables []string
	for _, codigo := range slices.Sorted(maps.Keys(ficheros.Municipios.Municipios)) {
		fila := ficheros.Municipios.Municipios[codigo]
		if motivo := alcanzable(registro, ficheros, codigo, fila); motivo != "" {
			inalcanzables = append(inalcanzables, fmt.Sprintf("el municipio %s, %q, %s", codigo, fila.Nombre, motivo))
		}
	}
	assert.Empty(t, inalcanzables,
		"por su nombre oficial se resuelve cada municipio, o una ambigüedad que lo nombra entre sus candidatos")
}

// comprobarNingunNombreEsSoloCifras exige que ningún nombre de municipio se
// pliegue a algo sin letras, que la resolución leería como un código
// (data-model §2.6).
func comprobarNingunNombreEsSoloCifras(t *testing.T, corpus *corpusDelRepositorio) {
	t.Helper()

	ficheros := corpus.ficheros(t)

	var sinLetras []string
	for _, codigo := range slices.Sorted(maps.Keys(ficheros.Municipios.Municipios)) {
		nombre := ficheros.Municipios.Municipios[codigo].Nombre
		if plegado := territorio.Plegar(nombre); !strings.ContainsFunc(plegado, unicode.IsLetter) {
			sinLetras = append(sinLetras, fmt.Sprintf("el municipio %s, %q, se pliega a %q", codigo, nombre, plegado))
		}
	}
	assert.Empty(t, sinLetras, "ningún nombre plegado es solo cifras, forma que la resolución lee como un código")
}

// fuentesDelRepositorio lee los ficheros congelados de data/territorio/ tal
// como los empaqueta el binario y exige que estén todos: los tres sueltos y los
// de las 19 comunidades, cada uno con la forma <código>.yaml. Sin ellos, los
// controles del corpus no dirían nada.
func fuentesDelRepositorio(t *testing.T) territorio.Fuentes {
	t.Helper()

	leer := func(ruta string) []byte {
		contenido, err := os.ReadFile(filepath.Clean(ruta))
		require.NoError(t, err, "el fichero congelado %s se puede leer", ruta)
		require.NotEmpty(t, contenido, "el fichero congelado %s tiene contenido", ruta)

		return contenido
	}

	entradas, err := os.ReadDir(comunidadesDelRepositorio)
	require.NoError(t, err, "la carpeta de comunidades %s se puede listar", comunidadesDelRepositorio)

	comunidades := make(map[string][]byte, len(entradas))
	for _, entrada := range entradas {
		codigo, conExtension := strings.CutSuffix(entrada.Name(), extensionDeComunidad)
		require.True(t, conExtension && entrada.Type().IsRegular(),
			"%s/%s es el fichero de una comunidad, <código>%s", comunidadesDelRepositorio, entrada.Name(),
			extensionDeComunidad)

		comunidades[codigo] = leer(filepath.Join(comunidadesDelRepositorio, entrada.Name()))
	}
	require.Len(t, slices.Sorted(maps.Keys(comunidades)), comunidadesYCiudades,
		"%s tiene el fichero de cada comunidad y ciudad autónoma", comunidadesDelRepositorio)

	return territorio.Fuentes{
		Municipios:  leer(municipiosDelRepositorio),
		DIR3:        leer(dir3DelRepositorio),
		Estado:      leer(estadoDelRepositorio),
		Comunidades: comunidades,
	}
}

// sourcesEmitidos son los source de todos los datos del territorio resuelto:
// el del DIR3 cuando lo trae, y cuando no lo trae pero declara un source, también,
// para que ninguno se escape.
func sourcesEmitidos(resuelto territorio.Territorio) []string {
	sources := []string{
		resuelto.Municipio.Source,
		resuelto.CodigoINE.Source,
		resuelto.Provincia.Source,
		resuelto.Comunidad.Source,
		resuelto.Regimen.Source,
	}

	if resuelto.DIR3.Codigo != "" || resuelto.DIR3.Source != "" {
		sources = append(sources, resuelto.DIR3.Source)
	}

	for _, boletin := range resuelto.Boletines {
		sources = append(sources, boletin.Source)
	}

	return sources
}

// filasDeSources son los identificadores de las filas de las tablas de
// docs/SOURCES.md: la primera celda de cada fila, entre comillas invertidas.
func filasDeSources(t *testing.T) []string {
	t.Helper()

	contenido, err := os.ReadFile(fuentesDocumentadas)
	require.NoError(t, err, "la tabla de fuentes %s se puede leer", fuentesDocumentadas)

	const inicioDeFila = "| `"

	var filas []string
	for linea := range strings.Lines(string(contenido)) {
		resto, esFila := strings.CutPrefix(strings.TrimSpace(linea), inicioDeFila)
		if !esFila {
			continue
		}

		identificador, _, cerrado := strings.Cut(resto, "`")
		require.True(t, cerrado, "la primera celda de la fila %q cierra sus comillas invertidas", linea)

		filas = append(filas, identificador)
	}
	require.NotEmpty(t, filas, "%s tiene filas de fuentes", fuentesDocumentadas)

	return filas
}

// esProcedenciaConocida dice si un source es el identificador de una fila de
// docs/SOURCES.md o la ruta, ya limpia, de un fichero congelado de
// data/territorio/ que existe (data-model §2.5).
func esProcedenciaConocida(filas []string, source string) bool {
	if slices.Contains(filas, source) {
		return true
	}

	if !strings.HasPrefix(source, carpetaDelTerritorio) || path.Clean(source) != source {
		return false
	}

	fichero, err := os.Stat(filepath.Join(arbolDelRepositorio, filepath.FromSlash(source)))

	return err == nil && fichero.Mode().IsRegular()
}

// nombreDelCorpus es un nombre de los ficheros congelados y de quién es.
type nombreDelCorpus struct {
	de, texto string
}

// nombresDelCorpus son todos los nombres de los ficheros congelados: el de cada
// municipio, el de cada provincia y el de cada comunidad, que son los que se
// comparan plegados con lo que se pregunta.
func nombresDelCorpus(ficheros territorio.Ficheros) []nombreDelCorpus {
	var nombres []nombreDelCorpus

	for _, codigo := range slices.Sorted(maps.Keys(ficheros.Municipios.Municipios)) {
		nombres = append(nombres, nombreDelCorpus{"el municipio " + codigo, ficheros.Municipios.Municipios[codigo].Nombre})
	}

	for _, codigo := range slices.Sorted(maps.Keys(ficheros.Comunidades)) {
		comunidad := ficheros.Comunidades[codigo]
		nombres = append(nombres, nombreDelCorpus{"la comunidad " + codigo, comunidad.Nombre})

		for _, provincia := range slices.Sorted(maps.Keys(comunidad.Provincias)) {
			nombres = append(nombres, nombreDelCorpus{"la provincia " + provincia, comunidad.Provincias[provincia]})
		}
	}

	return nombres
}

// runasFueraDelPliegue son las runas de un nombre plegado que no son del
// alfabeto del pliegue —minúsculas y cifras ASCII y el espacio—, o vacío si no
// hay ninguna: una runa que la tabla del pliegue no cubre (territorio.Plegar).
func runasFueraDelPliegue(plegado string) string {
	var ajenas strings.Builder

	for _, runa := range plegado {
		if ('a' <= runa && runa <= 'z') || ('0' <= runa && runa <= '9') || runa == ' ' {
			continue
		}

		ajenas.WriteRune(runa)
	}

	return ajenas.String()
}

// alcanzable dice por qué un municipio no se alcanza por su nombre oficial, o
// vacío si se alcanza: el nombre lo resuelve a él, o es ambiguo y la ambigüedad
// lo nombra entre sus candidatos, en la forma fija «<código INE> <nombre>
// (<provincia>)» (FR-013, FR-014).
func alcanzable(registro *territorio.Registro, ficheros territorio.Ficheros, codigo string,
	fila territorio.FilaDeMunicipio,
) string {
	resuelto, err := registro.Resolver(fila.Nombre)
	if err == nil {
		if resuelto.CodigoINE.Codigo != codigo {
			return "se resuelve al municipio " + resuelto.CodigoINE.Codigo
		}

		return ""
	}

	provincia := ficheros.Comunidades[fila.Comunidad].Provincias[fila.Provincia]
	candidato := fmt.Sprintf("%s %s (%s)", codigo, fila.Nombre, provincia)

	var conClase schema.ConClase
	if errors.As(err, &conClase) && conClase.Clase() == schema.ClaseArgumentos && strings.Contains(err.Error(), candidato) {
		return ""
	}

	return "no se resuelve ni es candidato de una ambigüedad: " + err.Error()
}

// ubicacionDeGramatica es un sitio de los esquemas de data/territorio/ donde va
// un identificador de internal/core/ids.
type ubicacionDeGramatica struct {
	nombre string
	// analizar es el analizador de internal/core/ids que lee lo que va ahí.
	analizar func(string) error
	// valido es el valor que ese sitio lleva en el territorio sintético.
	valido string
	// casos son los valores que se prueban en ese sitio.
	casos []string
	// documento es el fichero sintético donde está el sitio, forma cómo se
	// escribe el sitio, con %s donde va el valor, y poner deja el documento
	// cambiado en las fuentes.
	documento, forma string
	poner            func(fuentes *territorio.Fuentes, contenido []byte)
	// rechazo es el tipo de incumplimiento con el que el esquema lo rechaza:
	// el del patrón de la clave o el del patrón del valor.
	rechazo jsonschema.ErrorKind
}

// fuentes son las del territorio sintético con el valor en el sitio, escrito
// como texto JSON, que es YAML y lo conserva tal cual.
func (u ubicacionDeGramatica) fuentes(t *testing.T, valor string) territorio.Fuentes {
	t.Helper()

	enJSON := func(texto string) string {
		codificado, err := json.Marshal(texto)
		require.NoError(t, err)

		return string(codificado)
	}

	fuentes := fuentesDePrueba()
	u.poner(&fuentes, []byte(cambiada(t, u.documento, fmt.Sprintf(u.forma, enJSON(u.valido)),
		fmt.Sprintf(u.forma, enJSON(valor)))))

	return fuentes
}

// Los casos límite de cada gramática (contrato de identificadores §2 y §6):
// la forma, la longitud, las cifras que no son ASCII y los rangos de la
// provincia y del municipio.
var (
	casosDeCodigoINE = []string{
		"28074", "01001", "52999", "", "2807", "280748", "2807a", "00074", "53001", "99999", "28000",
		" 28074", "28074 ", "28074\n", "２８０７４",
	}
	casosDeProvincia = []string{"28", "01", "52", "", "2", "280", "00", "53", "99", "2a", " 28", "２８"}
	casosDeDigito    = []string{"0", "9", "", "a", "10", " 5", "٥"}
	casosDeDIR3      = []string{
		"L01280748", "l01280748", "L01010014", "L01529990", "", "L0128074", "L012807481", "X01280748",
		"L02280748", "L01000740", "L01530011", "L01280000", "L01 280748", "L0128074a", "Ｌ01280748",
	}
)

// comprobarGramaticas exige que el patrón de cada sitio de los esquemas de
// data/territorio/ donde va un identificador acepte exactamente lo que acepta
// su analizador de internal/core/ids (research.md V38): cada valor, en un
// territorio sintético que solo puede fallar por él, se lee sin error si y
// solo si el analizador lo acepta, y cuando no se lee es por el patrón de ese
// sitio y no por otra cosa.
func comprobarGramaticas(t *testing.T) {
	t.Helper()

	codigoINE := func(valor string) error {
		_, err := ids.AnalizarCodigoINE(valor)

		return err
	}
	// La provincia son las dos primeras cifras de un código INE, con un
	// municipio que siempre está en su rango.
	provincia := func(valor string) error { return codigoINE(valor + "001") }
	digito := func(valor string) error {
		_, _, err := ids.AnalizarCodigoINEConDigito("28991" + valor)

		return err
	}
	dir3 := func(valor string) error {
		_, err := ids.AnalizarDIR3(valor)

		return err
	}

	enMunicipios := func(fuentes *territorio.Fuentes, contenido []byte) { fuentes.Municipios = contenido }
	enDIR3 := func(fuentes *territorio.Fuentes, contenido []byte) { fuentes.DIR3 = contenido }
	enLaConfigurada := func(fuentes *territorio.Fuentes, contenido []byte) { fuentes.Comunidades["01"] = contenido }

	ubicaciones := []ubicacionDeGramatica{
		{
			nombre: "clave de municipios", analizar: codigoINE, valido: "28991", casos: casosDeCodigoINE,
			documento: municipiosDePrueba, forma: "%s: {dc:", poner: enMunicipios, rechazo: &kind.PropertyNames{},
		},
		{
			nombre: "provincia de un municipio", analizar: provincia, valido: "28", casos: casosDeProvincia,
			documento: municipiosDePrueba, forma: "provincia: %s", poner: enMunicipios, rechazo: &kind.Pattern{},
		},
		{
			nombre: "dígito de control de un municipio", analizar: digito, valido: "5", casos: casosDeDigito,
			documento: municipiosDePrueba, forma: "dc: %s", poner: enMunicipios, rechazo: &kind.Pattern{},
		},
		{
			nombre: "clave de la correspondencia", analizar: codigoINE, valido: "28991", casos: casosDeCodigoINE,
			documento: dir3DePrueba, forma: "%s: \"L01289915\"", poner: enDIR3, rechazo: &kind.PropertyNames{},
		},
		{
			nombre: "DIR3 de la correspondencia", analizar: dir3, valido: "L01289915", casos: casosDeDIR3,
			documento: dir3DePrueba, forma: "\"28991\": %s", poner: enDIR3, rechazo: &kind.Pattern{},
		},
		{
			nombre: "provincia de una comunidad", analizar: provincia, valido: "28", casos: casosDeProvincia,
			documento: comunidadConfiguradaDePrueba, forma: "%s: \"Provincia Única\"", poner: enLaConfigurada,
			rechazo: &kind.PropertyNames{},
		},
	}

	for _, ubicacion := range ubicaciones {
		t.Run(ubicacion.nombre, func(t *testing.T) {
			t.Parallel()

			require.NoError(t, ubicacion.analizar(ubicacion.valido))
			_, err := skills.LeerTerritorio(ubicacion.fuentes(t, ubicacion.valido))
			require.NoError(t, err, "con %q en %s el territorio sintético es válido: sin eso, un rechazo no "+
				"diría nada del patrón", ubicacion.valido, ubicacion.nombre)

			for _, valor := range ubicacion.casos {
				t.Run(fmt.Sprintf("%q", valor), func(t *testing.T) {
					t.Parallel()

					_, err := skills.LeerTerritorio(ubicacion.fuentes(t, valor))
					if ubicacion.analizar(valor) == nil {
						require.NoError(t, err, "el esquema acepta %q en %s, como internal/core/ids", valor, ubicacion.nombre)

						return
					}

					var defecto *skills.DefectoEnElDocumento
					require.ErrorAs(t, err, &defecto, "el esquema rechaza %q en %s, como internal/core/ids",
						valor, ubicacion.nombre)
					assert.IsType(t, ubicacion.rechazo, defecto.Tipo,
						"el rechazo de %q en %s es el del patrón de ese sitio: %v", valor, ubicacion.nombre, err)
				})
			}
		})
	}
}
