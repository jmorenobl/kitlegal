// superficie_test.go cierra la garantía de FR-002 por el único sitio por el que
// podría abrirse: lo que el paquete exporta. Ninguna declaración exportada
// nombra un *http.Client, un http.RoundTripper, un *http.Request ni un
// *http.Response, de modo que un adaptador de fuente no tiene forma de recibir
// uno, construirlo ni devolverlo —y, con él, de emitir una petición sin plazo,
// sin identificación, sin robots.txt y sin ritmo—. La imposibilidad es así del
// diseño y no del control: `noctx`, `bodyclose`, la regla R2 de `depguard` y
// internal/arch_test.go siguen siendo la red secundaria (FR-053, FR-054).
//
// Lo comprueba sobre el árbol sintáctico del fuente con go/parser y go/ast,
// porque es en el fuente donde se escribiría la fuga y porque así el control
// alcanza también lo que ninguna otra capa mira: el tipo de un campo exportado
// o el de una variable de paquete, que no son ni una firma de función ni una
// importación.
//
// Lo que **no** revisa, y no por descuido:
//
//   - El cuerpo de las funciones. Usar la biblioteca HTTP por dentro es
//     justamente lo que este paquete tiene que hacer; lo prohibido es que asome.
//   - Los campos y los métodos sin exportar de un tipo exportado. Ahí vive el
//     *http.Client de Cliente, que es donde tiene que vivir: quien está fuera
//     del paquete no puede nombrarlo (D1, D2).
//
// Va en el paquete externo httpx_test por la misma razón que el adaptador de
// prueba: mira el paquete desde donde lo mira quien lo usa.
package httpx_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bibliotecaHTTP es la ruta de importación que se persigue, tal cual aparece
// entrecomillada en el árbol sintáctico. Lo que se compara en cada mención no
// es esta cadena sino el nombre local que le da la importación del fichero, que
// no tiene por qué ser `http`.
const bibliotecaHTTP = `"net/http"`

// tiposVetados son los cuatro tipos que FR-002 deja fuera de la superficie
// exportada, por su nombre dentro de la biblioteca. Se comparan sin el
// asterisco a propósito: exportar `http.Request` por valor no sería mejor que
// exportar `*http.Request`, porque en los dos casos quien está fuera acaba
// teniendo que importar la biblioteca para nombrar lo que recibe (R2).
var tiposVetados = []string{"Client", "RoundTripper", "Request", "Response"}

// superficieDelCliente son las declaraciones que la revisión tiene que haber
// recorrido de verdad: la operación de red y los dos constructores, más los dos
// tipos que van y vienen por ellos (contrato §1, D1), y el ritmo común a más de
// un cliente, con su constructor y su opción (contrato servidor-mcp §7 de H21).
// Sin esta comprobación el test pasaría en vacío el día que la revisión dejara
// de encontrar nada, que es la única forma de que la garantía quedara sin
// vigilar sin que nadie lo notara.
var superficieDelCliente = []string{
	"func ConRitmo",
	"func New",
	"func NuevoRitmo",
	"func Replay",
	"método Cliente.Pedir",
	"tipo Cliente",
	"tipo Respuesta",
	"tipo Ritmo",
}

// TestSuperficieExportada es el control de FR-002 sobre el paquete entero, y
// además el control de sí mismo: la segunda subprueba demuestra sobre fuentes
// sintéticas que la revisión reconoce cada forma en la que un tipo de la
// biblioteca puede asomar y que no confunde con una fuga lo que se queda dentro.
func TestSuperficieExportada(t *testing.T) {
	t.Parallel()

	t.Run("ninguna declaración exportada nombra la biblioteca HTTP", func(t *testing.T) {
		t.Parallel()

		revisada := revision{conjunto: token.NewFileSet()}

		for _, fichero := range ficherosDelPaquete(t) {
			arbol, err := parser.ParseFile(revisada.conjunto, fichero, nil, parser.SkipObjectResolution)
			require.NoError(t, err, "%s no se pudo analizar", fichero)

			revisada.revisa(arbol)
		}

		require.Subset(t, revisada.declaraciones, superficieDelCliente,
			"la revisión no recorrió la superficie del cliente: no estaría comprobando nada (FR-002)")

		for _, fuga := range revisada.fugas {
			t.Errorf("%s. La superficie exportada no nombra %s: lo que Pedir devuelve es Respuesta, "+
				"un tipo de este paquete con el cuerpo ya leído, para que un adaptador de fuente use "+
				"lo que recibió sin importar la biblioteca HTTP y sin poder emitir nada por su cuenta "+
				"(FR-002, regla R2 de contracts/reglas-de-arquitectura.md).",
				fuga, strings.Join(tiposVetados, ", "))
		}
	})

	t.Run("la revisión distingue la superficie del interior", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre    string
			fragmento string
			fuga      bool
		}{
			{"resultado de una función exportada", "func Nueva() *http.Client { return nil }", true},
			{"parámetro de una función exportada", "func Con(t http.RoundTripper) {}", true},
			{"campo exportado de un tipo exportado", "type T struct{ Ultima *http.Response }", true},
			{"campo exportado por valor", "type T struct{ Peticion http.Request }", true},
			{"campo embebido de un tipo exportado", "type T struct{ *http.Request }", true},
			{"método de un tipo exportado", "func (T) Ultima() *http.Response { return nil }", true},
			{"método de una interfaz exportada", "type I interface{ Pedir(*http.Request) error }", true},
			{"variable exportada sin tipo declarado", "var Cliente = &http.Client{}", true},
			{"campo sin exportar de un tipo exportado", "type T struct{ cliente *http.Client }", false},
			{"cuerpo de una función exportada", "func Pedir() { _ = new(http.Client) }", false},
			{"función sin exportar", "func nueva() *http.Client { return nil }", false},
			{"método de un tipo sin exportar", "func (t) Ultima() *http.Response { return nil }", false},
			{"selector de otro paquete", "func Plazo() time.Duration { return 0 }", false},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				revisada := revision{conjunto: token.NewFileSet()}
				arbol, err := parser.ParseFile(revisada.conjunto, "sintetico.go",
					fuenteSintetica(caso.fragmento), parser.SkipObjectResolution)
				require.NoError(t, err)

				revisada.revisa(arbol)

				if caso.fuga {
					assert.NotEmpty(t, revisada.fugas,
						"la revisión no vio la fuga: dejaría pasar %q", caso.fragmento)

					return
				}

				assert.Empty(t, revisada.fugas,
					"la revisión denuncia lo que no es superficie: %q se queda dentro del paquete", caso.fragmento)
			})
		}
	})
}

// fuenteSintetica envuelve un fragmento en el fichero más pequeño que el
// analizador acepta. No se compila ni se comprueban sus tipos: la revisión
// trabaja sobre la sintaxis, así que al fragmento le basta con estar bien
// escrito.
func fuenteSintetica(fragmento string) string {
	return "package httpx\n\nimport \"net/http\"\n\n" + fragmento + "\n"
}

// ficherosDelPaquete enumera los ficheros de código del paquete: los .go del
// directorio que no son de test. El test se ejecuta con el directorio del
// paquete como directorio de trabajo, de modo que la enumeración sale del árbol
// y no de la configuración de construcción vigente. Es más estricto a
// propósito: un fichero excluido hoy por una etiqueta de compilación seguiría
// siendo superficie del paquete el día que la etiqueta se cumpliera.
func ficherosDelPaquete(t *testing.T) []string {
	t.Helper()

	entradas, err := os.ReadDir(".")
	require.NoError(t, err, "no se pudo enumerar el directorio del paquete")

	var ficheros []string

	for _, entrada := range entradas {
		nombre := entrada.Name()
		if entrada.IsDir() || filepath.Ext(nombre) != ".go" || strings.HasSuffix(nombre, "_test.go") {
			continue
		}

		ficheros = append(ficheros, nombre)
	}

	require.NotEmpty(t, ficheros, "el paquete no tiene ningún fichero de código que revisar")

	return ficheros
}

// revision es una pasada por las declaraciones exportadas de uno o varios
// ficheros. Acumula lo que recorrió —para poder demostrar que no pasó en
// vacío— y las fugas que encontró, cada una con su posición, que es lo que hace
// accionable el fallo.
type revision struct {
	// conjunto es el de posiciones de todos los ficheros analizados.
	conjunto *token.FileSet

	// alias es el nombre local que el fichero en curso le da a la biblioteca
	// HTTP, o la cadena vacía si no la importa.
	alias string

	// declaraciones son las exportadas que la revisión recorrió, nombradas
	// igual que en superficieDelCliente.
	declaraciones []string

	// fugas son las menciones encontradas, ya redactadas con su posición.
	fugas []string
}

// revisa mira una a una las declaraciones exportadas de un fichero.
func (r *revision) revisa(arbol *ast.File) {
	r.alias = aliasDeLaBiblioteca(arbol)

	// Una importación con punto metería los tipos de la biblioteca en el ámbito
	// del fichero sin prefijo que los distinguiera de los del paquete, y la
	// revisión ya no podría afirmar nada. `revive` la prohíbe en todo el árbol;
	// si alguna vez dejara de hacerlo, esto es un fallo y no un silencio.
	if r.alias == "." {
		r.fugas = append(r.fugas, fmt.Sprintf("%s: el fichero importa la biblioteca HTTP con «.», "+
			"así que sus tipos se nombran sin prefijo y la superficie no se puede revisar",
			r.posicion(arbol.Pos())))

		return
	}

	for _, declarada := range arbol.Decls {
		switch d := declarada.(type) {
		case *ast.FuncDecl:
			r.funcion(d)
		case *ast.GenDecl:
			r.grupo(d)
		}
	}
}

// funcion revisa una función o un método exportado: sus parámetros de tipo, sus
// parámetros y sus resultados, nunca su cuerpo. Un método de un tipo sin
// exportar no es superficie, porque nadie de fuera puede tener un valor de ese
// tipo: `revive` impide devolverlo (regla unexported-return).
func (r *revision) funcion(d *ast.FuncDecl) {
	receptor := nombreDelReceptor(d.Recv)
	if !d.Name.IsExported() || (d.Recv != nil && !ast.IsExported(receptor)) {
		return
	}

	nombre := "func " + d.Name.Name
	if d.Recv != nil {
		nombre = "método " + receptor + "." + d.Name.Name
	}

	r.declaraciones = append(r.declaraciones, nombre)
	r.inspecciona(nombre, d.Type)
}

// grupo revisa una declaración de tipos, de variables o de constantes. Las
// importaciones no llevan nada que revisar.
func (r *revision) grupo(d *ast.GenDecl) {
	for _, especificacion := range d.Specs {
		switch s := especificacion.(type) {
		case *ast.TypeSpec:
			if !s.Name.IsExported() {
				continue
			}

			nombre := "tipo " + s.Name.Name
			r.declaraciones = append(r.declaraciones, nombre)
			r.inspecciona(nombre, s.Type)

			if s.TypeParams != nil {
				r.inspecciona(nombre, s.TypeParams)
			}
		case *ast.ValueSpec:
			r.valor(s)
		}
	}
}

// valor revisa una variable o una constante exportada: el tipo que declara y,
// cuando no declara ninguno, la expresión de la que sale, porque el tipo de
// `var X = &http.Client{}` está en el valor y no en la declaración.
func (r *revision) valor(s *ast.ValueSpec) {
	nombres := make([]string, 0, len(s.Names))

	for _, identificador := range s.Names {
		if identificador.IsExported() {
			nombres = append(nombres, identificador.Name)
		}
	}

	if len(nombres) == 0 {
		return
	}

	nombre := "valor " + strings.Join(nombres, ", ")
	r.declaraciones = append(r.declaraciones, nombre)

	if s.Type != nil {
		r.inspecciona(nombre, s.Type)
	}

	for _, valor := range s.Values {
		r.inspecciona(nombre, valor)
	}
}

// inspecciona baja por un nodo anotando cada mención de un tipo vetado, y poda
// por el camino lo que no es superficie: los campos y los métodos sin exportar,
// y el cuerpo de cualquier función literal.
func (r *revision) inspecciona(nombre string, nodo ast.Node) {
	ast.Inspect(nodo, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.StructType:
			r.inspeccionaLoExportado(nombre, t.Fields)

			return false
		case *ast.InterfaceType:
			r.inspeccionaLoExportado(nombre, t.Methods)

			return false
		case *ast.FuncLit:
			r.inspecciona(nombre, t.Type)

			return false
		case *ast.SelectorExpr:
			r.compruebaMencion(nombre, t)
		}

		return true
	})
}

// inspeccionaLoExportado sigue solo los campos exportados de una estructura y
// los métodos exportados de una interfaz. Un elemento sin nombre está embebido
// y sí es superficie: promociona hacia fuera lo que el tipo embebido exporte.
func (r *revision) inspeccionaLoExportado(nombre string, elementos *ast.FieldList) {
	if elementos == nil {
		return
	}

	for _, elemento := range elementos.List {
		if len(elemento.Names) == 0 || slices.ContainsFunc(elemento.Names, (*ast.Ident).IsExported) {
			r.inspecciona(nombre, elemento.Type)
		}
	}
}

// compruebaMencion anota la mención si nombra uno de los tipos vetados con el
// nombre local que el fichero le da a la biblioteca.
func (r *revision) compruebaMencion(nombre string, mencion *ast.SelectorExpr) {
	paquete, esIdentificador := mencion.X.(*ast.Ident)
	if r.alias == "" || !esIdentificador || paquete.Name != r.alias {
		return
	}

	if !slices.Contains(tiposVetados, mencion.Sel.Name) {
		return
	}

	r.fugas = append(r.fugas, fmt.Sprintf("%s: %s nombra %s.%s en la superficie exportada",
		r.posicion(mencion.Pos()), nombre, r.alias, mencion.Sel.Name))
}

// posicion traduce una posición del árbol a «fichero:línea:columna».
func (r *revision) posicion(posicion token.Pos) string {
	return r.conjunto.Position(posicion).String()
}

// aliasDeLaBiblioteca devuelve el nombre local con el que el fichero nombra la
// biblioteca HTTP, o la cadena vacía si no la importa.
func aliasDeLaBiblioteca(arbol *ast.File) string {
	for _, importacion := range arbol.Imports {
		if importacion.Path.Value != bibliotecaHTTP {
			continue
		}

		if importacion.Name != nil {
			return importacion.Name.Name
		}

		return "http"
	}

	return ""
}

// nombreDelReceptor devuelve el nombre del tipo sobre el que se declara un
// método, sin puntero y sin parámetros de tipo. Devuelve la cadena vacía si no
// hay receptor.
func nombreDelReceptor(receptor *ast.FieldList) string {
	if receptor == nil || len(receptor.List) == 0 {
		return ""
	}

	return nombreDelTipo(receptor.List[0].Type)
}

// nombreDelTipo desnuda una expresión de tipo hasta el identificador que la
// nombra.
func nombreDelTipo(expresion ast.Expr) string {
	switch t := expresion.(type) {
	case *ast.StarExpr:
		return nombreDelTipo(t.X)
	case *ast.IndexExpr:
		return nombreDelTipo(t.X)
	case *ast.IndexListExpr:
		return nombreDelTipo(t.X)
	case *ast.Ident:
		return t.Name
	}

	return ""
}
