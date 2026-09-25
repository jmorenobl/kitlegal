package ids_test

import (
	"embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/ids"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestClaseDeLosErrores pasa por cada camino de rechazo del paquete y exige de
// todos lo mismo: que el error declare la clase «argumentos», que es la que el
// kernel traduce a código 2 sin que el dominio importe internal/cli; que la
// conserve envuelto con %w, que es como llega al kernel desde quien resuelve;
// que empiece nombrando el identificador que se esperaba, y que nombre la
// entrada con %q (FR-033, FR-096, contrato de identificadores §5).
func TestClaseDeLosErrores(t *testing.T) {
	t.Parallel()

	const (
		codigoINE          = "el código INE"
		codigoINEConDigito = "el código INE con dígito de control"
		codigoDIR3         = "el código DIR3"
	)

	leganes, err := ids.AnalizarCodigoINE("28074")
	require.NoError(t, err)

	analizarINE := func(entrada string) error {
		_, err := ids.AnalizarCodigoINE(entrada)

		return err
	}
	analizarINEConDigito := func(entrada string) error {
		_, _, err := ids.AnalizarCodigoINEConDigito(entrada)

		return err
	}
	analizarDIR3 := func(entrada string) error {
		_, err := ids.AnalizarDIR3(entrada)

		return err
	}

	casos := []struct {
		nombre        string
		identificador string
		entrada       string
		err           error
	}{
		{"INE de otra longitud", codigoINE, "2807", analizarINE("2807")},
		{"INE con algo que no es cifra", codigoINE, "2807a", analizarINE("2807a")},
		{"INE de provincia inexistente", codigoINE, "00074", analizarINE("00074")},
		{"INE de municipio 000", codigoINE, "28000", analizarINE("28000")},
		{"INE con dígito de otra longitud", codigoINEConDigito, "28074", analizarINEConDigito("28074")},
		{"INE con dígito con algo que no es cifra", codigoINEConDigito, "28074a", analizarINEConDigito("28074a")},
		{"INE con dígito de provincia inexistente", codigoINEConDigito, "530018", analizarINEConDigito("530018")},
		{"INE con dígito de municipio 000", codigoINEConDigito, "280008", analizarINEConDigito("280008")},
		{"INE con dígito distinto del oficial", codigoINEConDigito, "280749", leganes.ComprobarDigito('9', '5')},
		{"DIR3 vacío", codigoDIR3, "", analizarDIR3("")},
		{"DIR3 con otra letra", codigoDIR3, "X01280748", analizarDIR3("X01280748")},
		{"DIR3 con algo que no es cifra", codigoDIR3, "L0128074a", analizarDIR3("L0128074a")},
		{"DIR3 de otra longitud", codigoDIR3, "L0128074", analizarDIR3("L0128074")},
		{"DIR3 de otra entidad local", codigoDIR3, "L02280748", analizarDIR3("L02280748")},
		{"DIR3 de provincia inexistente", codigoDIR3, "L01000748", analizarDIR3("L01000748")},
		{"DIR3 de municipio 000", codigoDIR3, "L01280008", analizarDIR3("L01280008")},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			compruebaErrorDeArgumentos(t, caso.err, caso.entrada)
			compruebaErrorDeArgumentos(t, fmt.Errorf("resolver el territorio: %w", caso.err), caso.entrada)
			assert.True(t, strings.HasPrefix(caso.err.Error(), fmt.Sprintf("%s %q ", caso.identificador, caso.entrada)),
				"el mensaje %q no empieza nombrando %s y la entrada", caso.err.Error(), caso.identificador)
		})
	}
}

// compruebaErrorDeArgumentos exige que err sea un rechazo de entrada: un
// error, de clase «argumentos» y con la entrada entrecomillada con %q en el
// mensaje. La clase se busca como la busca el kernel, con errors.As a
// schema.ConClase siguiendo la cadena; el código 2 que le corresponde no se
// puede comprobar aquí, porque el dominio no importa internal/cli ni en sus
// tests (contrato de identificadores §5).
func compruebaErrorDeArgumentos(t *testing.T, err error, entrada string) {
	t.Helper()

	require.Error(t, err, "se acepta %q, que no es válida", entrada)

	var conClase schema.ConClase
	require.ErrorAs(t, err, &conClase, "el rechazo de %q no declara su clase", entrada)
	assert.Equal(t, schema.ClaseArgumentos, conClase.Clase(), "el rechazo de %q no es de argumentos", entrada)
	assert.Contains(t, err.Error(), fmt.Sprintf("%q", entrada), "el mensaje no nombra la entrada")
}

// compruebaRechazo es compruebaErrorDeArgumentos más el motivo: el mensaje
// dice qué tiene de malo la entrada.
func compruebaRechazo(t *testing.T, err error, entrada, motivo string) {
	t.Helper()

	compruebaErrorDeArgumentos(t, err, entrada)
	assert.Contains(t, err.Error(), motivo, "el mensaje no dice qué tiene de malo %q", entrada)
}

// fuentesDelPaquete son los ficheros Go del paquete, embebidos al compilar el
// test. El dominio no hace entrada ni salida tampoco en sus tests —la lista
// core de depguard le deniega os e io, y con ellos io/fs—, así que el test de
// superficie no lee el directorio: el compilador le entrega el fuente, que es
// exactamente el que se está compilando.
//
//go:embed *.go
var fuentesDelPaquete embed.FS

// superficieDelContrato es lo que el paquete exporta, copiado del contrato de
// identificadores §1: dos tipos con sus campos privados, tres analizadores, un
// constructor y seis métodos. Nada más: ELI, ECLI, CELEX y NIF entran con sus
// hitos (FR-034).
var superficieDelContrato = []string{
	"type CodigoINE struct{ /* campos privados */ }",
	"func AnalizarCodigoINE(entrada string) (CodigoINE, error)",
	"func AnalizarCodigoINEConDigito(entrada string) (codigo CodigoINE, digito byte, err error)",
	"func (c CodigoINE) String() string",
	"func (c CodigoINE) Provincia() string",
	"func (c CodigoINE) ComprobarDigito(declarado, oficial byte) error",
	"type DIR3 struct{ /* campos privados */ }",
	"func AnalizarDIR3(entrada string) (DIR3, error)",
	"func DIR3DeAyuntamiento(c CodigoINE, digito byte) DIR3",
	"func (d DIR3) String() string",
	"func (d DIR3) CodigoINE() CodigoINE",
	"func (d DIR3) Digito() byte",
}

// camposPrivados es como se escribe en superficieDelContrato un tipo
// estructura exportado cuyos campos no lo están: su interior no es
// superficie, y tiene que seguir sin serlo para que el valor solo se construya
// analizando.
const camposPrivados = "struct{ /* campos privados */ }"

// TestSuperficieDeIds exige que lo exportado sea exactamente el contrato §1,
// firma a firma, y se comprueba a sí mismo: la segunda subprueba demuestra
// sobre fuentes sintéticas que la revisión ve cada forma de exportar algo,
// para que un ELI, un ECLI, un CELEX o un NIF no pudieran entrar sin que el
// test lo notara (FR-034).
func TestSuperficieDeIds(t *testing.T) {
	t.Parallel()

	t.Run("lo exportado es el contrato", func(t *testing.T) {
		t.Parallel()

		entradas, err := fuentesDelPaquete.ReadDir(".")
		require.NoError(t, err)

		fuentes := map[string][]byte{}

		for _, entrada := range entradas {
			nombre := entrada.Name()
			if path.Ext(nombre) != ".go" || strings.HasSuffix(nombre, "_test.go") {
				continue
			}

			fuente, err := fuentesDelPaquete.ReadFile(nombre)
			require.NoError(t, err)

			fuentes[nombre] = fuente
		}

		require.NotEmpty(t, fuentes, "el paquete no tiene ningún fichero de código que revisar")
		assert.ElementsMatch(t, superficieDelContrato, superficieDe(t, fuentes),
			"lo exportado por internal/core/ids no es el contrato de identificadores §1: "+
				"ni ELI, ni ECLI, ni CELEX, ni NIF, que entran con sus hitos (FR-034)")
	})

	t.Run("la revisión ve toda forma de exportar", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre     string
			fragmento  string
			superficie []string
		}{
			{"función exportada", "func Nueva(a, b byte) error { return nil }", []string{"func Nueva(a, b byte) error"}},
			{"función sin exportar", "func nueva() {}", nil},
			{
				"método de un tipo exportado", "type T struct{ a int }\n\nfunc (t T) Valor() int { return t.a }",
				[]string{"type T " + camposPrivados, "func (t T) Valor() int"},
			},
			{
				"método con puntero", "type T struct{}\n\nfunc (t *T) Valor() {}",
				[]string{"type T " + camposPrivados, "func (t *T) Valor()"},
			},
			{"método sin exportar", "type T struct{}\n\nfunc (T) valor() {}", []string{"type T " + camposPrivados}},
			{"método de un tipo sin exportar", "type t struct{}\n\nfunc (x t) Valor() {}", nil},
			{"campo exportado", "type T struct{ A int }", []string{"type T struct{ A int }"}},
			{"campo embebido", "type T struct{ fmt.Stringer }", []string{"type T struct{ fmt.Stringer }"}},
			{"tipo que no es estructura", "type ECLI string", []string{"type ECLI string"}},
			{"variable exportada", "var ELI = \"\"", []string{"var ELI"}},
			{"constante exportada", "const NIF = \"\"", []string{"const NIF"}},
			{"grupo con una exportada", "var (\n\tcelex = 1\n\tCELEX = 2\n)", []string{"var CELEX"}},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				fuente := []byte("package ids\n\n" + caso.fragmento + "\n")
				assert.ElementsMatch(t, caso.superficie, superficieDe(t, map[string][]byte{"sintetico.go": fuente}))
			})
		}
	})
}

// superficieDe analiza los fuentes y devuelve sus declaraciones exportadas,
// escritas como las escribe superficieDelContrato: las funciones y los métodos
// de tipos exportados por su firma, sin cuerpo; los tipos estructura sin
// campos exportados ni embebidos, con camposPrivados; los demás tipos, enteros,
// y las variables y constantes, por su nombre. Un método de un tipo sin
// exportar no es superficie: nadie de fuera puede tener un valor de ese tipo.
func superficieDe(t *testing.T, fuentes map[string][]byte) []string {
	t.Helper()

	conjunto := token.NewFileSet()

	var superficie []string

	for nombre, fuente := range fuentes {
		arbol, err := parser.ParseFile(conjunto, nombre, fuente, parser.SkipObjectResolution)
		require.NoError(t, err, "%s no se pudo analizar", nombre)

		for _, declarada := range arbol.Decls {
			switch d := declarada.(type) {
			case *ast.FuncDecl:
				superficie = append(superficie, funcionExportada(t, conjunto, d)...)
			case *ast.GenDecl:
				superficie = append(superficie, grupoExportado(t, conjunto, d)...)
			}
		}
	}

	return superficie
}

// funcionExportada devuelve la firma de una función o de un método exportado
// de un tipo exportado, o nada.
func funcionExportada(t *testing.T, conjunto *token.FileSet, d *ast.FuncDecl) []string {
	t.Helper()

	if !d.Name.IsExported() || (d.Recv != nil && !ast.IsExported(nombreDelReceptor(d.Recv))) {
		return nil
	}

	firma := *d
	firma.Doc, firma.Body = nil, nil

	return []string{escrito(t, conjunto, &firma)}
}

// grupoExportado devuelve los tipos, las variables y las constantes
// exportados de una declaración. Las importaciones no llevan nada.
func grupoExportado(t *testing.T, conjunto *token.FileSet, d *ast.GenDecl) []string {
	t.Helper()

	var exportado []string

	for _, especificacion := range d.Specs {
		switch s := especificacion.(type) {
		case *ast.TypeSpec:
			if !s.Name.IsExported() {
				continue
			}

			if estructura, esEstructura := s.Type.(*ast.StructType); esEstructura && sinCamposExportados(estructura) {
				exportado = append(exportado, "type "+s.Name.Name+" "+camposPrivados)

				continue
			}

			exportado = append(exportado, "type "+escrito(t, conjunto, s))
		case *ast.ValueSpec:
			for _, nombre := range s.Names {
				if nombre.IsExported() {
					exportado = append(exportado, d.Tok.String()+" "+nombre.Name)
				}
			}
		}
	}

	return exportado
}

// sinCamposExportados dice si ningún campo de la estructura es superficie: ni
// exportado ni embebido, porque uno embebido promociona hacia fuera lo que
// exporte.
func sinCamposExportados(estructura *ast.StructType) bool {
	for _, campo := range estructura.Fields.List {
		if len(campo.Names) == 0 {
			return false
		}

		for _, nombre := range campo.Names {
			if nombre.IsExported() {
				return false
			}
		}
	}

	return true
}

// nombreDelReceptor devuelve el nombre del tipo sobre el que se declara un
// método, sin puntero y sin parámetros de tipo.
func nombreDelReceptor(receptor *ast.FieldList) string {
	if len(receptor.List) == 0 {
		return ""
	}

	tipo := receptor.List[0].Type
	if puntero, esPuntero := tipo.(*ast.StarExpr); esPuntero {
		tipo = puntero.X
	}

	switch generico := tipo.(type) {
	case *ast.IndexExpr:
		tipo = generico.X
	case *ast.IndexListExpr:
		tipo = generico.X
	}

	if identificador, esIdentificador := tipo.(*ast.Ident); esIdentificador {
		return identificador.Name
	}

	return ""
}

// escrito devuelve el nodo tal como lo escribe go/printer, en una sola línea
// si en el fuente ocupa una.
func escrito(t *testing.T, conjunto *token.FileSet, nodo ast.Node) string {
	t.Helper()

	var texto strings.Builder
	require.NoError(t, printer.Fprint(&texto, conjunto, nodo))

	return texto.String()
}
