package schema

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// metodoDeOperacion es el método sin exportar que sella Operacion: solo un tipo
// de este paquete puede declararlo.
const metodoDeOperacion = "operacionDeGrafo"

// TestOperacionSellada fija que una operación de grafo es un Nodo, una Arista o
// un Texto y nada más (FR-021, research.md D2). La interfaz tiene un único
// método y no es exportado, así que ningún tipo de otro paquete puede
// declararlo; dentro de este, solo lo declaran esos tres tipos, con receptor de
// valor, y ningún tipo lo hereda incrustando uno de ellos. Una operación no
// puede ser así dos cosas a la vez ni ninguna, que es lo que admitiría un
// struct con un campo de clase y tres punteros. Y ninguna lleva fuente, url ni
// fecha: la procedencia de cada una la pone el kernel al entregarla.
func TestOperacionSellada(t *testing.T) {
	t.Parallel()

	t.Run("un solo metodo, sin exportar", func(t *testing.T) {
		t.Parallel()

		operacion := reflect.TypeFor[Operacion]()
		require.Equal(t, 1, operacion.NumMethod(),
			"la interfaz no declara nada que un applet tenga que implementar: solo la sella")

		metodo := operacion.Method(0)
		assert.Equal(t, metodoDeOperacion, metodo.Name)
		assert.False(t, metodo.IsExported(),
			"un metodo exportado dejaria que cualquier tipo de otro paquete fuera una operacion")
	})

	t.Run("nodo, arista y texto son operaciones", func(t *testing.T) {
		t.Parallel()

		operacion := reflect.TypeFor[Operacion]()
		for _, tipo := range []reflect.Type{reflect.TypeFor[Nodo](), reflect.TypeFor[Arista](), reflect.TypeFor[Texto]()} {
			assert.True(t, tipo.Implements(operacion), "%s es una operacion por su valor", tipo.Name())
		}
	})

	t.Run("solo ellos la declaran, con receptor de valor, y nadie la hereda", func(t *testing.T) {
		t.Parallel()

		conjunto := token.NewFileSet()
		nombres, err := filepath.Glob("*.go")
		require.NoError(t, err)

		var ficheros []*ast.File

		for _, nombre := range nombres {
			if strings.HasSuffix(nombre, "_test.go") {
				continue
			}

			fichero, err := parser.ParseFile(conjunto, nombre, nil, parser.SkipObjectResolution)
			require.NoError(t, err, "%s no se pudo analizar", nombre)

			ficheros = append(ficheros, fichero)
		}

		require.NotEmpty(t, ficheros, "el paquete no tiene ningun fichero de codigo que revisar")

		revisado := revisaSellado(conjunto, ficheros)
		assert.ElementsMatch(t, []string{"Nodo", "Arista", "Texto"}, revisado.declaran,
			"los tipos que declaran %s, con su receptor", metodoDeOperacion)
		assert.Empty(t, revisado.herencias,
			"un tipo que incrusta una operacion o la interfaz la hereda y seria una cuarta clase")
	})

	t.Run("la revision del sellado distingue lo que lo abre", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre    string
			fragmento string
			declaran  []string
			herencias []string
		}{
			{
				nombre:    "los tres con receptor de valor",
				fragmento: "func (Nodo) operacionDeGrafo() {}\nfunc (Arista) operacionDeGrafo() {}\nfunc (t Texto) operacionDeGrafo() {}",
				declaran:  []string{"Nodo", "Arista", "Texto"},
			},
			{
				nombre:    "receptor de puntero",
				fragmento: "func (*Nodo) operacionDeGrafo() {}",
				declaran:  []string{"*Nodo"},
			},
			{
				nombre:    "funcion sin receptor con el mismo nombre",
				fragmento: "func operacionDeGrafo() {}",
			},
			{
				nombre:    "struct que incrusta una operacion",
				fragmento: "func (Nodo) operacionDeGrafo() {}\ntype Envoltorio struct{ Nodo }",
				declaran:  []string{"Nodo"},
				herencias: []string{"sintetico.go:4:25: incrusta Nodo"},
			},
			{
				nombre:    "puntero incrustado en un struct anonimo",
				fragmento: "func (Nodo) operacionDeGrafo() {}\nvar suelto struct{ *Nodo }",
				declaran:  []string{"Nodo"},
				herencias: []string{"sintetico.go:4:20: incrusta Nodo"},
			},
			{
				nombre:    "interfaz que incrusta la interfaz",
				fragmento: "type Otra interface{ Operacion }",
				herencias: []string{"sintetico.go:3:22: incrusta Operacion"},
			},
			{
				nombre:    "campo con nombre del mismo tipo",
				fragmento: "func (Nodo) operacionDeGrafo() {}\ntype Contenedor struct{ nodo Nodo }",
				declaran:  []string{"Nodo"},
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				conjunto := token.NewFileSet()
				fichero, err := parser.ParseFile(conjunto, "sintetico.go",
					"package schema\n\n"+caso.fragmento+"\n", parser.SkipObjectResolution)
				require.NoError(t, err)

				revisado := revisaSellado(conjunto, []*ast.File{fichero})
				assert.Equal(t, caso.declaran, revisado.declaran)
				assert.Equal(t, caso.herencias, revisado.herencias)
			})
		}
	})

	t.Run("ninguna operacion lleva fuente, url ni fecha", func(t *testing.T) {
		t.Parallel()

		// FR-021: la fuente, la dirección y la fecha de consulta de lo que se
		// observa son las de la Procedencia del Resultado, y las pone el kernel
		// al montar el lote; un applet no tiene dónde escribir otras.
		assert.Equal(t, []string{"ID", "Tipo", "Datos"}, camposDe(reflect.TypeFor[Nodo]()))
		assert.Equal(t, []string{"Origen", "Relacion", "Destino"}, camposDe(reflect.TypeFor[Arista]()))
		assert.Equal(t, []string{"Huella", "Cuerpo"}, camposDe(reflect.TypeFor[Texto]()))

		assert.Equal(t, reflect.TypeFor[map[string]any](), reflect.TypeFor[Nodo]().Field(2).Type,
			"los datos identificativos de un nodo son un objeto JSON")

		for _, tipo := range []reflect.Type{reflect.TypeFor[Arista](), reflect.TypeFor[Texto]()} {
			for i := range tipo.NumField() {
				assert.Equal(t, reflect.TypeFor[string](), tipo.Field(i).Type,
					"%s.%s es texto", tipo.Name(), tipo.Field(i).Name)
			}
		}
	})
}

// TestObservadoCero fija que el valor cero de lo observado es válido y no emite
// nada (FR-020, FR-047): un applet que no observa el mundo —skills, los de
// ejemplo, los verbos que solo buscan— devuelve su Resultado como antes, sin
// nombrar el campo, y lleva así un Observado sin vigencia declarada y sin
// ninguna operación.
func TestObservadoCero(t *testing.T) {
	t.Parallel()

	t.Run("sin vigencia declarada y sin operaciones", func(t *testing.T) {
		t.Parallel()

		var observado Observado
		assert.Zero(t, observado.Vigencia, "cero significa que la fuente no declara vigencia (FR-065)")
		assert.Empty(t, observado.Operaciones)
	})

	t.Run("un resultado que no nombra el grafo no emite", func(t *testing.T) {
		t.Parallel()

		resultado := Resultado{
			Procedencia: Procedencia{Fuente: "kitlegal.echo", URL: "kitlegal:applet/echo"},
			Datos:       map[string]any{"mensaje": "hola"},
		}
		assert.Equal(t, Observado{}, resultado.Grafo)
		assert.Empty(t, resultado.Grafo.Operaciones)
	})

	t.Run("la vigencia es una duracion y las operaciones una lista de operaciones", func(t *testing.T) {
		t.Parallel()

		// La vigencia viaja aquí y no en la Procedencia, que es lo que hace
		// citable un sobre y comparten todos los applets (research.md D3).
		assert.Equal(t, []string{"Vigencia", "Operaciones"}, camposDe(reflect.TypeFor[Observado]()))

		vigencia, existe := reflect.TypeFor[Observado]().FieldByName("Vigencia")
		require.True(t, existe)
		assert.Equal(t, reflect.TypeFor[time.Duration](), vigencia.Type)

		operaciones, existe := reflect.TypeFor[Observado]().FieldByName("Operaciones")
		require.True(t, existe)
		assert.Equal(t, reflect.TypeFor[[]Operacion](), operaciones.Type)

		grafo, existe := reflect.TypeFor[Resultado]().FieldByName("Grafo")
		require.True(t, existe, "el resultado declara el campo Grafo")
		assert.Equal(t, reflect.TypeFor[Observado](), grafo.Type)
	})
}

// sellado es lo que la revisión encuentra en unos ficheros del paquete: quién
// declara el método que sella Operacion y dónde se incrusta un tipo que la
// satisface.
type sellado struct {
	// declaran son los tipos que declaran el método, en el orden del fuente,
	// con «*» delante si el receptor es un puntero.
	declaran []string

	// herencias son las incrustaciones de la interfaz o de uno de esos tipos,
	// con su posición: cada una promociona el método a un tipo que no lo
	// declara.
	herencias []string
}

// revisaSellado pasa dos veces por los ficheros: la primera anota los tipos
// que declaran el método que sella Operacion; la segunda, cada campo
// incrustado —de un struct o de una interfaz, con nombre o anónimos, en
// cualquier declaración— cuyo tipo es la interfaz o uno de esos tipos, que es
// la única otra forma de que un tipo del paquete tenga el método.
func revisaSellado(conjunto *token.FileSet, ficheros []*ast.File) sellado {
	var revisado sellado

	satisfacen := map[string]bool{"Operacion": true}

	for _, fichero := range ficheros {
		for _, declarada := range fichero.Decls {
			funcion, esFuncion := declarada.(*ast.FuncDecl)
			if !esFuncion || funcion.Recv == nil || funcion.Name.Name != metodoDeOperacion {
				continue
			}

			receptor := funcion.Recv.List[0].Type
			nombre := nombreLocal(receptor)

			if _, esPuntero := receptor.(*ast.StarExpr); esPuntero {
				revisado.declaran = append(revisado.declaran, "*"+nombre)
			} else {
				revisado.declaran = append(revisado.declaran, nombre)
			}

			satisfacen[nombre] = true
		}
	}

	for _, fichero := range ficheros {
		ast.Inspect(fichero, func(nodo ast.Node) bool {
			var campos *ast.FieldList

			switch tipo := nodo.(type) {
			case *ast.StructType:
				campos = tipo.Fields
			case *ast.InterfaceType:
				campos = tipo.Methods
			default:
				return true
			}

			for _, campo := range campos.List {
				if nombre := nombreLocal(campo.Type); len(campo.Names) == 0 && satisfacen[nombre] {
					revisado.herencias = append(revisado.herencias,
						fmt.Sprintf("%s: incrusta %s", conjunto.Position(campo.Pos()), nombre))
				}
			}

			return true
		})
	}

	return revisado
}

// nombreLocal devuelve el nombre de un tipo de este paquete, con o sin
// puntero, y la cadena vacía para cualquier otra expresión: un tipo de otro
// paquete no puede declarar el método sin exportar de este.
func nombreLocal(expresion ast.Expr) string {
	if puntero, esPuntero := expresion.(*ast.StarExpr); esPuntero {
		expresion = puntero.X
	}

	if identificador, esIdentificador := expresion.(*ast.Ident); esIdentificador {
		return identificador.Name
	}

	return ""
}
