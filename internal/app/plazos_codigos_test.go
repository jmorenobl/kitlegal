package app

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

// ficherosDeLosCodigosDePlazos son los dos ficheros de plazos de cuyos códigos
// el Go del binario no nombra ninguno, por su ruta desde internal/app:
// data/plazos.yaml y el del cómputo de ensayo de spec §C.10 de H36
// (contracts/applet-calcular.md §4).
var ficherosDeLosCodigosDePlazos = []string{
	"../../data/plazos.yaml",
	"../core/plazos/testdata/computo-de-ensayo.yaml",
}

// paquetesSinCodigosDePlazos son los directorios, desde internal/app y sin sus
// subdirectorios, de los cuatro paquetes cuyo Go no nombra ningún código de
// plazos: el dominio, el de este applet, el de las skills y el de las evals
// (research D6 de H36).
var paquetesSinCodigosDePlazos = []string{"../core/plazos", ".", "../skills", "../evals"}

// codigosDeUnFicheroDePlazos son los códigos de un fichero de plazos: los de sus
// tipos, los de sus cómputos y los de los avisos de sus periodos. Lo lee una
// estructura propia del test, sin KnownFields, porque solo saca códigos:
// LeerTipos ya valida los dos ficheros en sus tests (research D6 de H36).
type codigosDeUnFicheroDePlazos struct {
	Tipos []struct {
		Codigo string `yaml:"codigo"`
	} `yaml:"tipos"`
	Computos []struct {
		Codigo   string `yaml:"codigo"`
		Periodos []struct {
			Aviso *struct {
				Codigo string `yaml:"codigo"`
			} `yaml:"aviso"`
		} `yaml:"periodos"`
	} `yaml:"computos"`
}

// TestNingunCodigoDePlazosEnGo fija que el Go del binario no nombra ningún tipo,
// cómputo ni aviso de periodo de los datos de plazos (H36 FR-010; SC-003;
// contracts/applet-calcular.md §4; research D6): ningún literal de cadena de un
// .go que no es de test de los cuatro paquetes es entero un código de
// data/plazos.yaml o del cómputo de ensayo, y ninguna etiqueta de campo lo lleva
// como palabra entera, con el límite de palabra de grep -w. Cada defecto dice
// el fichero, la línea y el código.
func TestNingunCodigoDePlazosEnGo(t *testing.T) {
	t.Parallel()

	codigos := codigosDePlazos(t)
	assert.Empty(t, codigosNombradosEnElGo(t, codigos),
		"el Go del binario no nombra ningún código de tipo, de cómputo ni de aviso de periodo de los datos")
}

// codigosDePlazos son los códigos de los dos ficheros de plazos, ordenados y
// sin repetir. Exige en cada fichero al menos uno de cada clase —de tipo, de
// cómputo y de aviso de periodo—, para que un fichero que no se lee, o que se
// lee mal, no deje el test en verde, y que ninguno esté vacío.
func codigosDePlazos(t *testing.T) []string {
	t.Helper()

	var codigos []string

	for _, ruta := range ficherosDeLosCodigosDePlazos {
		contenido, err := os.ReadFile(filepath.FromSlash(ruta))
		require.NoError(t, err)

		var fichero codigosDeUnFicheroDePlazos
		require.NoError(t, yaml.Unmarshal(contenido, &fichero), ruta)

		var tipos, computos, avisos []string

		for _, tipo := range fichero.Tipos {
			tipos = append(tipos, tipo.Codigo)
		}

		for _, computo := range fichero.Computos {
			computos = append(computos, computo.Codigo)

			for _, periodo := range computo.Periodos {
				if periodo.Aviso != nil {
					avisos = append(avisos, periodo.Aviso.Codigo)
				}
			}
		}

		require.NotEmpty(t, tipos, "%s lleva algún código de tipo", ruta)
		require.NotEmpty(t, computos, "%s lleva algún código de cómputo", ruta)
		require.NotEmpty(t, avisos, "%s lleva algún código de aviso de periodo", ruta)

		codigos = slices.Concat(codigos, tipos, computos, avisos)
	}

	require.NotContains(t, codigos, "", "ningún código está vacío")

	slices.Sort(codigos)

	return slices.Compact(codigos)
}

// codigosNombradosEnElGo son los defectos del Go de paquetesSinCodigosDePlazos,
// leído con go/parser: uno por cada literal de cadena cuyo valor es entero un
// código y otro por cada etiqueta de campo de un struct que lo lleva como
// palabra entera, cada uno con su fichero, su línea y su código, en el orden de
// los paquetes y, en cada uno, de sus ficheros y sus líneas. Exige que cada
// paquete tenga algún fichero que leer.
func codigosNombradosEnElGo(t *testing.T, codigos []string) []string {
	t.Helper()

	palabras := make(map[string]*regexp.Regexp, len(codigos))
	for _, codigo := range codigos {
		palabras[codigo] = regexp.MustCompile(`\b` + regexp.QuoteMeta(codigo) + `\b`)
	}

	conjunto := conjuntoDe(codigos)
	ficheros := token.NewFileSet()

	var defectos []string

	for _, directorio := range paquetesSinCodigosDePlazos {
		rutas, err := filepath.Glob(filepath.Join(directorio, "*.go"))
		require.NoError(t, err)

		rutas = slices.DeleteFunc(rutas, func(ruta string) bool { return strings.HasSuffix(ruta, "_test.go") })
		require.NotEmpty(t, rutas, "el paquete %s tiene algún .go que no es de test", directorio)

		for _, ruta := range rutas {
			fichero, err := parser.ParseFile(ficheros, ruta, nil, parser.SkipObjectResolution)
			require.NoError(t, err)

			ast.Inspect(fichero, func(nodo ast.Node) bool {
				switch nodo := nodo.(type) {
				case *ast.Field:
					if nodo.Tag != nil {
						etiqueta := sinComillas(t, nodo.Tag)
						for _, codigo := range codigos {
							if palabras[codigo].MatchString(etiqueta) {
								defectos = append(defectos, defectoDeCodigo(ficheros, nodo.Tag, "la etiqueta lleva", codigo))
							}
						}
					}
				case *ast.BasicLit:
					if nodo.Kind == token.STRING && conjunto[sinComillas(t, nodo)] {
						defectos = append(defectos, defectoDeCodigo(ficheros, nodo, "el literal es", sinComillas(t, nodo)))
					}
				}

				return true
			})
		}
	}

	return defectos
}

// conjuntoDe es el conjunto de los códigos.
func conjuntoDe(codigos []string) map[string]bool {
	conjunto := make(map[string]bool, len(codigos))
	for _, codigo := range codigos {
		conjunto[codigo] = true
	}

	return conjunto
}

// sinComillas es el valor de un literal de cadena, sin sus comillas ni sus
// escapes.
func sinComillas(t *testing.T, literal *ast.BasicLit) string {
	t.Helper()

	valor, err := strconv.Unquote(literal.Value)
	require.NoError(t, err, "el literal %s es una cadena de Go", literal.Value)

	return valor
}

// defectoDeCodigo es un defecto de TestNingunCodigoDePlazosEnGo: el fichero y la
// línea del literal, y el código que nombra.
func defectoDeCodigo(ficheros *token.FileSet, literal *ast.BasicLit, que, codigo string) string {
	posicion := ficheros.Position(literal.Pos())

	return fmt.Sprintf("%s:%d: %s el código de plazos %q", filepath.ToSlash(posicion.Filename), posicion.Line, que,
		codigo)
}
