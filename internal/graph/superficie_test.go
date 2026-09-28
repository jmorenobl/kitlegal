// superficie_test.go comprueba que ninguna declaración exportada de
// internal/graph nombra la biblioteca de base de datos ni el controlador de
// SQLite (contracts/almacen-world-db.md §1 y §7; como FR-005 de H3 en la
// caché). Quien usa el grafo desde fuera —el kernel, el applet graph, la
// preparación de evals— solo puede entregar un lote o leer por los métodos del
// contrato: no puede recibir la conexión, construirla ni pasarle una sentencia.
// La regla R3 de `depguard` e internal/arch_test.go son la red secundaria.
//
// Lo comprueba sobre el árbol sintáctico del fuente, que es donde se escribiría
// la fuga, con una pasada que mira solo lo exportado —firmas de funciones y
// métodos, campos y métodos exportados de los tipos exportados, y el tipo y el
// valor de variables y constantes exportadas— y nunca los cuerpos, que es donde
// el paquete usa la base, ni lo que no se exporta. A diferencia de la caché, no
// cierra la lista de lo exportado: la superficie del grafo crece tarea a tarea
// y lo que se comprueba es que ninguna parte de ella nombra la base.
//
// Va en el paquete externo graph_test porque mira el paquete desde donde lo
// mira quien lo usa.
package graph_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// vetadas son las rutas de importación que la superficie no puede nombrar, con
// sus subpaquetes: database/sql/driver y el lib del controlador también.
var vetadas = []string{"database/sql", "modernc.org/sqlite"}

// TestSuperficieExportada revisa el paquete entero y, en su segunda subprueba,
// se revisa a sí misma sobre fuentes sintéticas: que ve cada forma en que la
// base puede asomar y que no confunde con una fuga lo que se queda dentro.
func TestSuperficieExportada(t *testing.T) {
	t.Parallel()

	t.Run("ninguna declaración exportada nombra database/sql ni modernc.org/sqlite", func(t *testing.T) {
		t.Parallel()

		entradas, err := os.ReadDir(".")
		require.NoError(t, err)

		pasada := pasadaDeSuperficie{ficheros: token.NewFileSet()}

		for _, entrada := range entradas {
			nombre := entrada.Name()
			if entrada.IsDir() || filepath.Ext(nombre) != ".go" || strings.HasSuffix(nombre, "_test.go") {
				continue
			}

			arbol, err := parser.ParseFile(pasada.ficheros, nombre, nil, parser.SkipObjectResolution)
			require.NoError(t, err, "%s no se pudo analizar", nombre)

			pasada.revisa(arbol)
		}

		// Sin esto la prueba pasaría en vacío el día que dejara de ver la
		// superficie: lo que el contrato exporta ya en esta tarea tiene que
		// aparecer.
		for _, esperada := range []string{"Opcion", "ConDirectorio", "Error", "Error.Clase", "Nulo", "Nulo.Apply"} {
			assert.Contains(t, pasada.exportadas, esperada, "la revisión no ve %s", esperada)
		}

		for _, fuga := range pasada.fugas {
			t.Errorf("%s: la superficie exportada del grafo no nombra %s; la conexión se queda dentro del "+
				"paquete (contracts/almacen-world-db.md §1, regla R3)", fuga, strings.Join(vetadas, " ni "))
		}
	})

	t.Run("control: la revisión distingue la superficie del interior", func(t *testing.T) {
		t.Parallel()

		const cabecera = "import (\n\t\"database/sql\"\n\t\"time\"\n\n\t\"modernc.org/sqlite\"\n" +
			"\tsqlite3 \"modernc.org/sqlite/lib\"\n)\n"

		casos := []struct {
			nombre, importaciones, fragmento string
			fuga                             bool
		}{
			{nombre: "resultado de una función", fragmento: "func Base() *sql.DB { return nil }", fuga: true},
			{nombre: "parámetro de un método", fragmento: "func (*T) Ejecuta(r sql.Result) {}", fuga: true},
			{nombre: "parámetro de tipo", fragmento: "func F[T sql.Scanner]() {}", fuga: true},
			{nombre: "campo exportado", fragmento: "type T struct{ Base *sql.DB }", fuga: true},
			{nombre: "campo embebido", fragmento: "type T struct{ *sql.DB }", fuga: true},
			{nombre: "método de una interfaz", fragmento: "type I interface{ Q(string) (*sql.Rows, error) }", fuga: true},
			{nombre: "tipo definido sobre otro", fragmento: "type Base sql.DB", fuga: true},
			{nombre: "alias", fragmento: "type Base = sql.DB", fuga: true},
			{nombre: "variable sin tipo declarado", fragmento: "var Base = &sql.DB{}", fuga: true},
			{nombre: "error del controlador", fragmento: "type E struct{ Motor *sqlite.Error }", fuga: true},
			{nombre: "constante del controlador", fragmento: "const Ocupado = sqlite3.SQLITE_BUSY", fuga: true},
			{
				nombre:        "biblioteca con otro nombre",
				importaciones: "import base \"database/sql\"\n",
				fragmento:     "func Base() *base.DB { return nil }",
				fuga:          true,
			},
			{
				nombre:        "biblioteca importada con punto",
				importaciones: "import . \"database/sql\"\n",
				fragmento:     "func Base() *DB { return nil }",
				fuga:          true,
			},
			{
				nombre:        "paquete que ninguna importación explica",
				importaciones: "import \"modernc.org/sqlite/lib\"\n",
				fragmento:     "const Ocupado = sqlite3.SQLITE_BUSY",
				fuga:          true,
			},
			{nombre: "campo sin exportar", fragmento: "type T struct{ base *sql.DB }"},
			{nombre: "cuerpo de una función", fragmento: `func Abre() { _, _ = sql.Open("sqlite", "") }`},
			{nombre: "función literal en un valor", fragmento: "var F = func() { _ = sql.ErrNoRows }"},
			{nombre: "función sin exportar", fragmento: "func abre() *sql.DB { return nil }"},
			{nombre: "método de un tipo sin exportar", fragmento: "func (*base) Base() *sql.DB { return nil }"},
			{nombre: "tipo sin exportar", fragmento: "type base struct{ Base *sql.DB }"},
			{nombre: "otro paquete", fragmento: "func Plazo() time.Duration { return 0 }"},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				importaciones := caso.importaciones
				if importaciones == "" {
					importaciones = cabecera
				}

				pasada := pasadaDeSuperficie{ficheros: token.NewFileSet()}
				arbol, err := parser.ParseFile(pasada.ficheros, "sintetico.go",
					"package graph\n\n"+importaciones+"\n"+caso.fragmento+"\n", parser.SkipObjectResolution)
				require.NoError(t, err)

				pasada.revisa(arbol)

				if caso.fuga {
					assert.NotEmpty(t, pasada.fugas, "la revisión deja pasar %q", caso.fragmento)
				} else {
					assert.Empty(t, pasada.fugas, "la revisión denuncia lo que no es superficie: %q", caso.fragmento)
				}
			})
		}
	})
}

// pasadaDeSuperficie mira las declaraciones exportadas de uno o varios
// ficheros y acumula lo que ve —para comprobar que ve algo— y las fugas, cada
// una con su posición.
type pasadaDeSuperficie struct {
	ficheros *token.FileSet
	// nombres es la ruta de cada importación del fichero en curso por el nombre
	// con que el fichero la menciona.
	nombres map[string]string
	// exportadas son las declaraciones exportadas vistas: «Tipo», «Tipo.Campo»,
	// «Tipo.Método» o «Función».
	exportadas []string
	fugas      []string
}

// revisa mira las importaciones y las declaraciones de un fichero.
func (p *pasadaDeSuperficie) revisa(arbol *ast.File) {
	p.nombres = map[string]string{}

	for _, importacion := range arbol.Imports {
		ruta, err := strconv.Unquote(importacion.Path.Value)
		if err != nil {
			p.anota(importacion.Pos(), "la importación %s no se puede leer", importacion.Path.Value)

			continue
		}

		nombre := path.Base(ruta)
		if importacion.Name != nil {
			nombre = importacion.Name.Name
		}

		if nombre == "." && esVetada(ruta) {
			p.anota(importacion.Pos(), "importa %q con «.», y sus tipos se nombran sin prefijo", ruta)
		}

		p.nombres[nombre] = ruta
	}

	for _, cada := range arbol.Decls {
		switch d := cada.(type) {
		case *ast.FuncDecl:
			p.funcion(d)
		case *ast.GenDecl:
			for _, especificacion := range d.Specs {
				p.especificacion(especificacion)
			}
		}
	}
}

// funcion revisa la firma de una función exportada o de un método exportado de
// un tipo exportado, nunca su cuerpo.
func (p *pasadaDeSuperficie) funcion(d *ast.FuncDecl) {
	nombre := d.Name.Name

	if d.Recv != nil && len(d.Recv.List) > 0 {
		receptor := nombreDeTipo(d.Recv.List[0].Type)
		if !ast.IsExported(receptor) {
			return
		}

		nombre = receptor + "." + nombre
	}

	if !d.Name.IsExported() {
		return
	}

	p.exportadas = append(p.exportadas, nombre)
	p.inspecciona(d.Type)
}

// especificacion revisa un tipo, una variable o una constante exportados.
func (p *pasadaDeSuperficie) especificacion(especificacion ast.Spec) {
	switch s := especificacion.(type) {
	case *ast.TypeSpec:
		if !s.Name.IsExported() {
			return
		}

		p.exportadas = append(p.exportadas, s.Name.Name)
		p.miembros(s.Name.Name, s.Type)

		if s.TypeParams != nil {
			p.inspecciona(s.TypeParams)
		}

		p.inspecciona(s.Type)
	case *ast.ValueSpec:
		if !slices.ContainsFunc(s.Names, (*ast.Ident).IsExported) {
			return
		}

		for _, identificador := range s.Names {
			p.exportadas = append(p.exportadas, identificador.Name)
		}

		if s.Type != nil {
			p.inspecciona(s.Type)
		}

		for _, valor := range s.Values {
			p.inspecciona(valor)
		}
	}
}

// miembros anota los campos y los métodos exportados de un tipo exportado
// estructura o interfaz. Uno embebido se llama como su tipo.
func (p *pasadaDeSuperficie) miembros(tipo string, cuerpo ast.Expr) {
	var lista *ast.FieldList

	switch c := cuerpo.(type) {
	case *ast.StructType:
		lista = c.Fields
	case *ast.InterfaceType:
		lista = c.Methods
	default:
		return
	}

	for _, miembro := range lista.List {
		for _, nombre := range nombresDe(miembro) {
			if ast.IsExported(nombre) {
				p.exportadas = append(p.exportadas, tipo+"."+nombre)
			}
		}
	}
}

// inspecciona baja por un nodo de la superficie y anota cada mención de un
// paquete vetado, o de uno que ninguna importación explica. Poda lo que no es
// superficie: los campos y métodos sin exportar de estructuras e interfaces
// —los parámetros de una firma sí lo son, se llamen como se llamen— y el cuerpo
// de las funciones literales.
func (p *pasadaDeSuperficie) inspecciona(nodo ast.Node) {
	ast.Inspect(nodo, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.FuncLit:
			p.inspecciona(t.Type)

			return false
		case *ast.StructType:
			p.exportados(t.Fields)

			return false
		case *ast.InterfaceType:
			p.exportados(t.Methods)

			return false
		case *ast.SelectorExpr:
			p.mencion(t)
		}

		return true
	})
}

// exportados inspecciona los campos o los métodos exportados de una estructura
// o de una interfaz; uno embebido cuenta si su tipo es exportado, porque
// promociona hacia fuera lo que ese tipo exporta.
func (p *pasadaDeSuperficie) exportados(lista *ast.FieldList) {
	for _, miembro := range lista.List {
		if slices.ContainsFunc(nombresDe(miembro), ast.IsExported) {
			p.inspecciona(miembro.Type)
		}
	}
}

// mencion anota un selector sobre un paquete vetado o desconocido.
func (p *pasadaDeSuperficie) mencion(selector *ast.SelectorExpr) {
	paquete, esIdentificador := selector.X.(*ast.Ident)
	if !esIdentificador {
		return
	}

	ruta, conocido := p.nombres[paquete.Name]

	switch {
	case !conocido:
		p.anota(selector.Pos(), "nombra %s.%s y ninguna importación se llama %s", paquete.Name, selector.Sel.Name,
			paquete.Name)
	case esVetada(ruta):
		p.anota(selector.Pos(), "nombra %s.%s, de %q", paquete.Name, selector.Sel.Name, ruta)
	}
}

// anota añade una fuga con su posición.
func (p *pasadaDeSuperficie) anota(posicion token.Pos, formato string, argumentos ...any) {
	p.fugas = append(p.fugas, p.ficheros.Position(posicion).String()+": "+fmt.Sprintf(formato, argumentos...))
}

// nombresDe devuelve los nombres de un campo o de un método; uno embebido, o un
// parámetro sin nombre, se llama como su tipo.
func nombresDe(miembro *ast.Field) []string {
	if len(miembro.Names) == 0 {
		return []string{nombreDeTipo(miembro.Type)}
	}

	nombres := make([]string, 0, len(miembro.Names))
	for _, identificador := range miembro.Names {
		nombres = append(nombres, identificador.Name)
	}

	return nombres
}

// nombreDeTipo desnuda una expresión de tipo hasta el identificador que la
// nombra: sin puntero, sin parámetros de tipo y sin el paquete.
func nombreDeTipo(expresion ast.Expr) string {
	switch t := expresion.(type) {
	case *ast.StarExpr:
		return nombreDeTipo(t.X)
	case *ast.IndexExpr:
		return nombreDeTipo(t.X)
	case *ast.IndexListExpr:
		return nombreDeTipo(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	case *ast.Ident:
		return t.Name
	default:
		return ""
	}
}

// esVetada dice si la ruta es una de las vetadas o un subpaquete suyo.
func esVetada(ruta string) bool {
	return slices.ContainsFunc(vetadas, func(vetada string) bool {
		return ruta == vetada || strings.HasPrefix(ruta, vetada+"/")
	})
}
