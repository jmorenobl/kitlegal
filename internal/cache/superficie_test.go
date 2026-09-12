// superficie_test.go cierra la garantía de FR-005 por el único sitio por el que
// podría abrirse: lo que el paquete exporta. Ninguna declaración exportada nombra
// la biblioteca de base de datos ni el controlador de SQLite, y el paquete no
// exporta nada que no esté en la lista del contrato del puerto y el cliente §2 y
// del contrato de errores §2. Con las dos cosas, quien usa la caché desde fuera
// no tiene forma de recibir la conexión, de construirla ni de pasarle una
// sentencia: la imposibilidad de ejecutar SQL arbitrario es así del diseño y no
// del control. La regla R3 de `depguard`, `sqlclosecheck`, `rowserrcheck` e
// internal/arch_test.go siguen siendo la red secundaria (FR-005, SC-013).
//
// Hacen falta las dos comprobaciones, y ninguna basta sola. La primera ve la
// conexión cuando asoma por su tipo: un método que devuelva *sql.DB, un campo
// exportado que lo lleve, un tipo que lo embeba. La segunda ve lo que ningún
// tipo delata: un método Ejecuta(consulta string) error no nombra ningún paquete
// y ejecutaría lo que se le pasara.
//
// Lo comprueba sobre el árbol sintáctico del fuente con go/parser y go/ast, como
// el test de superficie de internal/httpx, porque es en el fuente donde se
// escribiría la fuga y porque así alcanza lo que ninguna otra capa mira: el tipo
// de un campo exportado o el valor de una constante exportada.
//
// Lo que **no** revisa, y no por descuido:
//
//   - El cuerpo de las funciones. Usar la base de datos por dentro es justamente
//     lo que este paquete tiene que hacer; lo prohibido es que asome.
//   - Los campos y los métodos sin exportar. Ahí vive el *sql.DB de Cliente, que
//     es donde tiene que vivir: quien está fuera del paquete no puede nombrarlo
//     (D2).
//
// Va en el paquete externo cache_test porque mira el paquete desde donde lo mira
// quien lo usa.
package cache_test

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

// paquetesDeLaBase son las rutas de importación que la superficie no puede
// nombrar: la biblioteca de base de datos y el controlador de SQLite, cada una
// con sus subpaquetes —database/sql/driver, el lib del controlador—, porque
// nombrar cualquiera de ellos obliga a quien está fuera a importar lo que la
// regla R3 reserva a los paquetes de almacenamiento.
var paquetesDeLaBase = []string{"database/sql", "modernc.org/sqlite"}

// superficieDelContrato es la lista cerrada de lo que el paquete exporta, con el
// nombre con el que la revisión anota cada declaración: la constante, el tipo
// del cliente, el constructor, el tipo de las opciones y las cuatro opciones; los
// tres métodos del cliente; y Error con sus tres métodos y los cinco campos que
// el contrato de errores §2 declara. Cliente no lleva ningún campo exportado: el
// contrato lo declara privado entero.
//
// Se compara en los dos sentidos. Lo que sobra es superficie que nadie aprobó; lo
// que falta quiere decir que la revisión dejó de ver algo, y sin esa mitad el
// test pasaría en vacío el día que no encontrara nada.
var superficieDelContrato = []string{
	"const VariableDirectorio",
	"tipo Cliente",
	"func New",
	"tipo Opcion",
	"func ConDirectorio",
	"func SoloLectura",
	"func ConReloj",
	"func ConRegistrador",
	"método Cliente.Get",
	"método Cliente.Put",
	"método Cliente.Close",
	"tipo Error",
	"campo Error.Operacion",
	"campo Error.Ruta",
	"campo Error.Origen",
	"campo Error.Clave",
	"campo Error.Causa",
	"método Error.Error",
	"método Error.Unwrap",
	"método Error.Clase",
}

// importacionesSinteticas es la cabecera con la que se analizan los fragmentos
// de la segunda subprueba cuando el caso no trae la suya: las dos rutas vetadas
// —el lib del controlador con el nombre que tiene su paquete— y una que no lo
// está.
const importacionesSinteticas = `import (
	"database/sql"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)`

// TestSuperficieExportada es el control de FR-005 sobre el paquete entero, y
// además el control de sí mismo: la segunda subprueba demuestra sobre fuentes
// sintéticas que la revisión anota cada declaración con el nombre que usa la
// lista, que reconoce cada forma en la que la base de datos puede asomar y que no
// confunde con una fuga lo que se queda dentro.
func TestSuperficieExportada(t *testing.T) {
	t.Parallel()

	t.Run("la superficie es la del contrato y no nombra la base de datos", func(t *testing.T) {
		t.Parallel()

		revisada := revision{conjunto: token.NewFileSet()}

		for _, fichero := range ficherosDelPaquete(t) {
			arbol, err := parser.ParseFile(revisada.conjunto, fichero, nil, parser.SkipObjectResolution)
			require.NoError(t, err, "%s no se pudo analizar", fichero)

			revisada.revisa(arbol)
		}

		for _, sobrante := range fueraDe(revisada.declaraciones, superficieDelContrato) {
			t.Errorf("%s: %s está exportado y no está en la lista del contrato. Cada declaración "+
				"exportada es algo que quien usa la caché puede llamar o nombrar, y la única forma de que "+
				"nadie ejecute SQL desde fuera es que no haya nada más que lo que el contrato aprueba: si "+
				"hace falta, se cambia primero el contrato del puerto y el cliente §2 y después esta lista "+
				"(FR-005).", revisada.posiciones[sobrante], sobrante)
		}

		for _, ausente := range fueraDe(superficieDelContrato, revisada.declaraciones) {
			t.Errorf("la revisión no encontró %s: o el contrato cambió sin cambiar esta lista, o la "+
				"revisión dejó de ver la superficie y no estaría comprobando nada (FR-005).", ausente)
		}

		for _, fuga := range revisada.fugas {
			t.Errorf("%s. La superficie exportada no nombra %s: la conexión y todo lo que la "+
				"acompaña se quedan dentro del paquete, y quien usa la caché solo puede hacer lo que el "+
				"puerto declara (FR-005, regla R3 de contracts/reglas-de-arquitectura.md).",
				fuga, strings.Join(paquetesDeLaBase, " ni "))
		}
	})

	t.Run("la revisión distingue la superficie del interior", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre    string
			cabecera  string
			fragmento string
			declara   []string
			fuga      bool
		}{
			{
				nombre:    "resultado de una función exportada",
				fragmento: "func Base() *sql.DB { return nil }",
				declara:   []string{"func Base"},
				fuga:      true,
			},
			{
				nombre:    "parámetro de un método exportado",
				fragmento: "func (*T) Ejecuta(r sql.Result) {}",
				declara:   []string{"método T.Ejecuta"},
				fuga:      true,
			},
			{
				nombre:    "campo exportado de un tipo exportado",
				fragmento: "type T struct{ Base *sql.DB }",
				declara:   []string{"tipo T", "campo T.Base"},
				fuga:      true,
			},
			{
				nombre:    "campo embebido de un tipo exportado",
				fragmento: "type T struct{ *sql.DB }",
				declara:   []string{"tipo T", "campo T.DB"},
				fuga:      true,
			},
			{
				nombre:    "método de una interfaz exportada",
				fragmento: "type I interface{ Consulta(string) (*sql.Rows, error) }",
				declara:   []string{"tipo I", "método I.Consulta"},
				fuga:      true,
			},
			{
				nombre:    "variable exportada sin tipo declarado",
				fragmento: "var Base = &sql.DB{}",
				declara:   []string{"var Base"},
				fuga:      true,
			},
			{
				nombre:    "error del controlador en un campo exportado",
				fragmento: "type E struct{ Origen *sqlite.Error }",
				declara:   []string{"tipo E", "campo E.Origen"},
				fuga:      true,
			},
			{
				nombre:    "constante exportada que sale del controlador",
				fragmento: "const Ocupado = sqlite3.SQLITE_BUSY",
				declara:   []string{"const Ocupado"},
				fuga:      true,
			},
			{
				nombre:    "biblioteca importada con otro nombre",
				cabecera:  `import base "database/sql"`,
				fragmento: "func Base() *base.DB { return nil }",
				declara:   []string{"func Base"},
				fuga:      true,
			},
			{
				nombre:    "biblioteca importada con punto",
				cabecera:  `import . "database/sql"`,
				fragmento: "func Base() *DB { return nil }",
				declara:   []string{"func Base"},
				fuga:      true,
			},
			{
				nombre:    "selector que ninguna importación explica",
				cabecera:  `import "modernc.org/sqlite/lib"`,
				fragmento: "const Ocupado = sqlite3.SQLITE_BUSY",
				declara:   []string{"const Ocupado"},
				fuga:      true,
			},
			{
				// No nombra ningún paquete: esta fuga no la ve el tipo sino la
				// lista, que no contiene «método T.Ejecuta».
				nombre:    "método que ejecuta sin nombrar la base de datos",
				fragmento: "func (*T) Ejecuta(consulta string) error { return nil }",
				declara:   []string{"método T.Ejecuta"},
			},
			{
				nombre:    "campo sin exportar de un tipo exportado",
				fragmento: "type T struct{ base *sql.DB }",
				declara:   []string{"tipo T"},
			},
			{
				nombre:    "cuerpo de una función exportada",
				fragmento: `func Abre() { _, _ = sql.Open("sqlite", "") }`,
				declara:   []string{"func Abre"},
			},
			{
				nombre:    "función sin exportar",
				fragmento: "func abre() *sql.DB { return nil }",
			},
			{
				nombre:    "método de un tipo sin exportar",
				fragmento: "func (c *cliente) Base() *sql.DB { return nil }",
			},
			{
				nombre:    "selector de otro paquete",
				fragmento: "func Plazo() time.Duration { return 0 }",
				declara:   []string{"func Plazo"},
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				revisada := revision{conjunto: token.NewFileSet()}
				arbol, err := parser.ParseFile(revisada.conjunto, "sintetico.go",
					fuenteSintetica(caso.cabecera, caso.fragmento), parser.SkipObjectResolution)
				require.NoError(t, err)

				revisada.revisa(arbol)

				assert.Equal(t, caso.declara, revisada.declaraciones,
					"la revisión anota cada declaración exportada con el nombre que usa la lista del contrato")

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
// analizador acepta, con la cabecera del caso o, si no trae ninguna, con la
// común. No se compila ni se comprueban sus tipos: la revisión trabaja sobre la
// sintaxis, así que al fragmento le basta con estar bien escrito.
func fuenteSintetica(cabecera, fragmento string) string {
	if cabecera == "" {
		cabecera = importacionesSinteticas
	}

	return "package cache\n\n" + cabecera + "\n\n" + fragmento + "\n"
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
// ficheros. Acumula lo que encontró —para compararlo con la lista del contrato— y
// las fugas, cada una con su posición, que es lo que hace accionable el fallo.
type revision struct {
	// conjunto es el de posiciones de todos los ficheros analizados.
	conjunto *token.FileSet

	// importaciones son las del fichero en curso: la ruta de cada una por el
	// nombre con el que el fichero la menciona.
	importaciones map[string]string

	// declaraciones son las exportadas, en el orden en que aparecen y nombradas
	// igual que en superficieDelContrato; posiciones dice dónde está cada una.
	declaraciones []string
	posiciones    map[string]string

	// fugas son las menciones encontradas, ya redactadas con su posición.
	fugas []string
}

// revisa mira una a una las declaraciones exportadas de un fichero.
func (r *revision) revisa(arbol *ast.File) {
	r.importaciones = r.importacionesDe(arbol)

	for _, declarada := range arbol.Decls {
		switch d := declarada.(type) {
		case *ast.FuncDecl:
			r.funcion(d)
		case *ast.GenDecl:
			r.grupo(d)
		}
	}
}

// importacionesDe devuelve la ruta de cada importación del fichero por su nombre
// local. Sin alias, el nombre local que se toma es el último tramo de la ruta;
// cuando el paquete se llama de otra forma —el lib del controlador se llama
// sqlite3—, sus menciones no se encuentran aquí y compruebaMencion las denuncia
// en vez de callarlas.
//
// Importar la base de datos con punto es una fuga por sí solo: mete sus tipos en
// el ámbito del fichero sin prefijo que los distinga de los del paquete, y la
// revisión ya no podría afirmar nada. `revive` lo prohíbe en todo el árbol; si
// alguna vez dejara de hacerlo, esto es un fallo y no un silencio.
func (r *revision) importacionesDe(arbol *ast.File) map[string]string {
	importaciones := make(map[string]string, len(arbol.Imports))

	for _, importacion := range arbol.Imports {
		ruta, err := strconv.Unquote(importacion.Path.Value)
		if err != nil {
			r.fugas = append(r.fugas, fmt.Sprintf("%s: la importación %s no se pudo leer, así que no se "+
				"sabe qué nombra la superficie", r.posicion(importacion.Pos()), importacion.Path.Value))

			continue
		}

		nombre := path.Base(ruta)
		if importacion.Name != nil {
			nombre = importacion.Name.Name
		}

		if nombre == "." && esDeLaBase(ruta) {
			r.fugas = append(r.fugas, fmt.Sprintf("%s: el fichero importa %q con «.», así que sus tipos se "+
				"nombran sin prefijo y la superficie no se puede revisar", r.posicion(importacion.Pos()), ruta))
		}

		importaciones[nombre] = ruta
	}

	return importaciones
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

	r.declara(nombre, d.Name.Pos())
	r.inspecciona(nombre, d.Type)
}

// grupo revisa una declaración de tipos, de variables o de constantes. Las
// importaciones no llevan nada que revisar.
func (r *revision) grupo(d *ast.GenDecl) {
	for _, especificacion := range d.Specs {
		switch s := especificacion.(type) {
		case *ast.TypeSpec:
			r.tipo(s)
		case *ast.ValueSpec:
			r.valor(d.Tok, s)
		}
	}
}

// tipo revisa un tipo exportado. Si es una estructura o una interfaz, anota
// además cada campo o cada método exportado, que también son superficie y
// también tienen que estar en la lista: un campo exportado de Cliente sería una
// forma de alcanzar la conexión que ninguna firma de función enseña.
func (r *revision) tipo(s *ast.TypeSpec) {
	if !s.Name.IsExported() {
		return
	}

	nombre := "tipo " + s.Name.Name
	r.declara(nombre, s.Name.Pos())

	if s.TypeParams != nil {
		r.inspecciona(nombre, s.TypeParams)
	}

	switch cuerpo := s.Type.(type) {
	case *ast.StructType:
		r.miembros(s.Name.Name, "campo", cuerpo.Fields)
	case *ast.InterfaceType:
		r.miembros(s.Name.Name, "método", cuerpo.Methods)
	default:
		r.inspecciona(nombre, s.Type)
	}
}

// miembros anota e inspecciona los campos exportados de una estructura o los
// métodos exportados de una interfaz. Un miembro sin nombre está embebido y se
// llama como su tipo, y es superficie si ese nombre es exportado, porque
// promociona hacia fuera todo lo que el tipo embebido exporte: embeber *sql.DB
// le daría ExecContext a quien tuviera un Cliente.
func (r *revision) miembros(tipo, clase string, lista *ast.FieldList) {
	if lista == nil {
		return
	}

	for _, miembro := range lista.List {
		exportados := nombresExportados(miembro)
		if len(exportados) == 0 {
			continue
		}

		for _, exportado := range exportados {
			r.declara(clase+" "+tipo+"."+exportado, miembro.Pos())
		}

		r.inspecciona(clase+" "+tipo+"."+strings.Join(exportados, ", "), miembro.Type)
	}
}

// valor revisa una variable o una constante exportada: el tipo que declara y,
// cuando no declara ninguno, la expresión de la que sale, porque el tipo de
// `var X = &sql.DB{}` está en el valor y no en la declaración.
func (r *revision) valor(clase token.Token, s *ast.ValueSpec) {
	exportados := make([]string, 0, len(s.Names))

	for _, identificador := range s.Names {
		if identificador.IsExported() {
			exportados = append(exportados, identificador.Name)
			r.declara(clase.String()+" "+identificador.Name, identificador.Pos())
		}
	}

	if len(exportados) == 0 {
		return
	}

	nombre := clase.String() + " " + strings.Join(exportados, ", ")

	if s.Type != nil {
		r.inspecciona(nombre, s.Type)
	}

	for _, valor := range s.Values {
		r.inspecciona(nombre, valor)
	}
}

// declara anota una declaración exportada y dónde está.
func (r *revision) declara(nombre string, posicion token.Pos) {
	r.declaraciones = append(r.declaraciones, nombre)

	if r.posiciones == nil {
		r.posiciones = make(map[string]string)
	}

	r.posiciones[nombre] = r.posicion(posicion)
}

// inspecciona baja por un nodo anotando cada mención que salga de la base de
// datos, y poda por el camino lo que no es superficie: los campos y los métodos
// sin exportar, y el cuerpo de cualquier función literal.
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
// los métodos exportados de una interfaz anidadas en una declaración. Un
// elemento sin nombre está embebido y sí es superficie.
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

// compruebaMencion anota la mención si sale de la base de datos, y también si no
// se puede saber de dónde sale: en una declaración, un selector sobre un
// identificador es un paquete, y un paquete que ninguna importación del fichero
// explica podría ser justamente el controlador con el nombre de su paquete.
func (r *revision) compruebaMencion(nombre string, mencion *ast.SelectorExpr) {
	paquete, esIdentificador := mencion.X.(*ast.Ident)
	if !esIdentificador {
		return
	}

	ruta, importada := r.importaciones[paquete.Name]

	switch {
	case !importada:
		r.fugas = append(r.fugas, fmt.Sprintf("%s: %s nombra %s.%s y ninguna importación del fichero se "+
			"llama %s, así que no se puede saber si sale de la base de datos",
			r.posicion(mencion.Pos()), nombre, paquete.Name, mencion.Sel.Name, paquete.Name))
	case esDeLaBase(ruta):
		r.fugas = append(r.fugas, fmt.Sprintf("%s: %s nombra %s.%s, de %q, en la superficie exportada",
			r.posicion(mencion.Pos()), nombre, paquete.Name, mencion.Sel.Name, ruta))
	}
}

// posicion traduce una posición del árbol a «fichero:línea:columna».
func (r *revision) posicion(posicion token.Pos) string {
	return r.conjunto.Position(posicion).String()
}

// esDeLaBase dice si una ruta de importación es una de las vetadas o un
// subpaquete suyo.
func esDeLaBase(ruta string) bool {
	return slices.ContainsFunc(paquetesDeLaBase, func(vetada string) bool {
		return ruta == vetada || strings.HasPrefix(ruta, vetada+"/")
	})
}

// fueraDe devuelve, en su orden, los elementos de unos que no están en otros.
func fueraDe(unos, otros []string) []string {
	var restantes []string

	for _, elemento := range unos {
		if !slices.Contains(otros, elemento) {
			restantes = append(restantes, elemento)
		}
	}

	return restantes
}

// nombresExportados devuelve los nombres exportados de un campo o de un método.
// Uno embebido se llama como su tipo.
func nombresExportados(miembro *ast.Field) []string {
	if len(miembro.Names) == 0 {
		if nombre := nombreDelTipo(miembro.Type); ast.IsExported(nombre) {
			return []string{nombre}
		}

		return nil
	}

	var exportados []string

	for _, identificador := range miembro.Names {
		if identificador.IsExported() {
			exportados = append(exportados, identificador.Name)
		}
	}

	return exportados
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
// nombra: sin puntero, sin parámetros de tipo y, si es de otro paquete, sin el
// paquete, que es el nombre que recibe un campo embebido.
func nombreDelTipo(expresion ast.Expr) string {
	switch t := expresion.(type) {
	case *ast.StarExpr:
		return nombreDelTipo(t.X)
	case *ast.IndexExpr:
		return nombreDelTipo(t.X)
	case *ast.IndexListExpr:
		return nombreDelTipo(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	case *ast.Ident:
		return t.Name
	}

	return ""
}
