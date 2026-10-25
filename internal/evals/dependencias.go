package evals

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// Lo común y las rutas que miden las evals son listas fijas en el código, sin
// leer el disco ni la definición del job (FR-002, FR-004; data-model §1 y §2;
// research D4).
var (
	// directoriosDeLoComun son las carpetas de lo común (FR-002): el kernel y
	// la composición del binario, internal/cli, la red, la caché, el servidor MCP
	// y el arnés de las evals, enteros, con sus tests y sus fixtures.
	directoriosDeLoComun = []string{
		"cmd/", "internal/app/", "internal/cli/", "internal/httpx/", "internal/cache/", "internal/mcp/",
		"internal/evals/",
	}

	// ficherosDeLoComun son los ficheros sueltos de lo común (FR-002): lo que
	// instala el binario y las skills en el job, el job, los guiones que ejecuta
	// y los esquemas de las evals y del juez que lee el arnés.
	ficherosDeLoComun = []string{
		"skills.go", "Makefile", ".github/workflows/evals.yml",
		"scripts/evals.sh", "scripts/evals-sesion.sh", "scripts/evals-voto.sh", "scripts/evals-medir-juez.sh",
		"schemas/eval.yaml.json", "schemas/expresiones-prohibidas.yaml.json", "schemas/juez-clases.yaml.json",
	}

	// directoriosQueMiden son las carpetas cuyas rutas miden las evals
	// (FR-004), además de las raíces de grabaciones.
	directoriosQueMiden = []string{"skills/", "evals/", "data/", "cmd/", "internal/"}

	// ficherosQueMiden son los ficheros sueltos que miden las evals además de
	// los de lo común: go.mod y go.sum (FR-004).
	ficherosQueMiden = []string{"go.mod", "go.sum"}

	// plataformaDelJob es la plataforma con la que se lista el grafo de
	// paquetes, la del runner del job: así el cálculo da lo mismo en macOS y en
	// el runner (research D3, V2).
	plataformaDelJob = []string{"GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0"}
)

const (
	// directorioDelWorkflow es el único de internal/ que no mide ninguna eval:
	// el del workflow, que el binario no enlaza (FR-027; data-model §1).
	directorioDelWorkflow = "internal/workflow/"

	// directorioDeLaComposicion es el del paquete que compone los applets del
	// binario: de cada uno, el fichero cuyo Nombre() devuelve su nombre
	// (research D3, V7).
	directorioDeLaComposicion = "internal/app"

	// directorioDelArnes es el del paquete de las evals, de cuyos tests salen
	// también los módulos de lo común (data-model §2).
	directorioDelArnes = "internal/evals"

	// Las carpetas de las skills y de sus evals, relativas a la raíz, y lo
	// que hay en ellas.
	directorioDeLasSkills      = "skills"
	directorioDeLasEvals       = "evals"
	directorioDeLasReferencias = "references"
	ficheroDeLaSkill           = "SKILL.md"

	// formaDeUnaEval es el patrón de path.Match del nombre de un fichero de
	// eval de evals/<skill>/: lo que no casa con él, y no es la carpeta del
	// juez, entra por (f) (data-model §3).
	formaDeUnaEval = "[0-9][0-9]-*.yaml"

	// metodoDelNombre es el método con el que un applet declara su nombre
	// (Applet en internal/app).
	metodoDelNombre = "Nombre"
)

// Las variables del entorno con las que las órdenes de la base de la propuesta
// reciben sus valores, y esas órdenes, de texto constante: van por sh -c, como
// las de la tanda (contracts/eleccion-y-huella.md §1; gosec G204).
const (
	variableDeLaBase   = "KITLEGAL_EVALS_BASE"
	variableDelFichero = "KITLEGAL_EVALS_FICHERO"
	variableDelGoMod   = "KITLEGAL_EVALS_GO_MOD"
	ordenDeLaFusion    = `git merge-base "$` + variableDeLaBase + `" HEAD`
	ordenDeLasRutas    = `git ls-tree -r -z --name-only "$` + variableDeLaBase + `" -- skills/ evals/`
	ordenDeUnFichero   = `git show "$` + variableDeLaBase + `:$` + variableDelFichero + `"`
	ordenDelGoMod      = `go mod edit -json "$` + variableDelGoMod + `"`
	nombreDelGoMod     = "go.mod"
	patronDelTemporal  = "kitlegal-evals-go-mod-*"
)

// claseDeEntrada es la clase de una entrada del conjunto de dependencias de una
// eval (data-model §3; research D2).
type claseDeEntrada string

// Las tres clases de entrada: un fichero, un directorio, con todo lo que tiene
// debajo, y un paquete de Go, con sus .go que no son de tests, sin sus
// subdirectorios.
const (
	entradaFichero    claseDeEntrada = "fichero"
	entradaDirectorio claseDeEntrada = "directorio"
	entradaPaquete    claseDeEntrada = "paquete"
)

// parteDelConjunto es la parte de FR-001 por la que entra una entrada.
type parteDelConjunto string

// Las seis partes del conjunto de una eval (FR-001; data-model §3).
const (
	// parteDeLaEval es (a), el fichero de la eval.
	parteDeLaEval parteDelConjunto = "a"
	// parteDeLasSkills es (b), el SKILL.md y las referencias de cada skill de la
	// eval, las generadas y las a mano: la que mide y las que nombran, por
	// cierre.
	parteDeLasSkills parteDelConjunto = "b"
	// parteDeLosDatos es (c), los data/*.yaml de los que salen las referencias
	// generadas de esas skills.
	parteDeLosDatos parteDelConjunto = "c"
	// parteDeLosPaquetes es (d), los paquetes de los applets de las tablas de
	// esas skills y lo que empotran.
	parteDeLosPaquetes parteDelConjunto = "d"
	// parteDeLasGrabaciones es (e), las grabaciones que el arnés sirve a sus
	// sesiones.
	parteDeLasGrabaciones parteDelConjunto = "e"
	// parteDelJuez es (f), la carpeta del juez de su skill y los demás ficheros
	// de su directorio de evals que no son evals.
	parteDelJuez parteDelConjunto = "f"
)

// entradaDeDependencia es una entrada del conjunto de una eval: su clase, su
// ruta relativa a la raíz —la de un directorio, con la barra final; la de un
// paquete, sin ella— y la parte por la que entra.
type entradaDeDependencia struct {
	clase claseDeEntrada
	ruta  string
	parte parteDelConjunto
}

// contiene dice si la ruta es de la entrada: igual a la de un fichero, debajo
// de la de un directorio, o un .go que no es de tests en el directorio de un
// paquete, sin sus subdirectorios (data-model §3, «Pertenencia»). No lee el
// disco: vale para una ruta que ya no está en la cabeza.
func (e entradaDeDependencia) contiene(ruta string) bool {
	switch e.clase {
	case entradaFichero:
		return ruta == e.ruta
	case entradaDirectorio:
		return strings.HasPrefix(ruta, e.ruta)
	case entradaPaquete:
		return path.Dir(ruta) == e.ruta && strings.HasSuffix(ruta, ".go") && !strings.HasSuffix(ruta, "_test.go")
	}

	return false
}

// dependenciasDeLaEval es el conjunto de dependencias de una eval (FR-001;
// data-model §3), calculado sobre un árbol y nunca escrito: sus entradas y la
// versión de cada módulo externo de sus paquetes, o el error que impide
// calcularlo, que empieza por el nombre de la eval (FR-005).
type dependenciasDeLaEval struct {
	// skill es la que mide la eval, el nombre de su directorio de evals/.
	skill string

	// eval es el nombre de su fichero, sin .yaml.
	eval string

	// entradas son las de sus seis partes, en orden de parte, de clase y de
	// ruta, sin repetir.
	entradas []entradaDeDependencia

	// modulos son la ruta y la versión de cada módulo externo de los paquetes
	// de (d).
	modulos map[string]string

	// err dice por qué el conjunto no se puede calcular; nil si se calcula.
	err error
}

// nombre es el de la eval, <skill>/<fichero sin .yaml>.
func (d dependenciasDeLaEval) nombre() string {
	return d.skill + "/" + d.eval
}

// pertenece dice si la ruta es del conjunto: de alguna de sus entradas.
func (d dependenciasDeLaEval) pertenece(ruta string) bool {
	return slices.ContainsFunc(d.entradas, func(entrada entradaDeDependencia) bool {
		return entrada.contiene(ruta)
	})
}

// paquetesDeLosConjuntos son las carpetas de las entradas paquete de los
// conjuntos, ordenados y sin repetir: los que el predicado de las rutas que
// miden necesita (data-model §1).
func paquetesDeLosConjuntos(conjuntos []dependenciasDeLaEval) []string {
	var paquetes []string

	for _, conjunto := range conjuntos {
		for _, entrada := range conjunto.entradas {
			if entrada.clase == entradaPaquete {
				paquetes = append(paquetes, entrada.ruta)
			}
		}
	}

	slices.Sort(paquetes)

	return slices.Compact(paquetes)
}

// rutaQueMide es el predicado de las rutas que miden las evals (FR-004;
// data-model §1), sobre una ruta relativa a la raíz y las carpetas de los
// paquetes de los conjuntos, sin leer el disco: mide una ruta de las raíces de
// grabaciones; no mide ninguna de internal/workflow; de las demás, mide la que
// es de lo común o de las listas de lo que mide, salvo un _test.go o un fixture
// de un paquete de un conjunto.
func rutaQueMide(ruta string, paquetes []string) bool {
	switch {
	case empiezaPorAlguno(ruta, raicesDeGrabaciones()):
		return true
	case strings.HasPrefix(ruta, directorioDelWorkflow):
		return false
	case !esDeLoComun(ruta) && !slices.Contains(ficherosQueMiden, ruta) && !empiezaPorAlguno(ruta, directoriosQueMiden):
		return false
	}

	return !slices.ContainsFunc(paquetes, func(paquete string) bool {
		return esDePruebaDelPaquete(ruta, paquete)
	})
}

// esDePruebaDelPaquete dice si la ruta es un _test.go del directorio del
// paquete, sin sus subdirectorios, o está en sus fixtures.
func esDePruebaDelPaquete(ruta, paquete string) bool {
	return (path.Dir(ruta) == paquete && strings.HasSuffix(ruta, "_test.go")) ||
		strings.HasPrefix(ruta, paquete+"/testdata/")
}

// esDeLoComun dice si la ruta es de lo común (FR-002; data-model §2): uno de
// sus ficheros o una ruta de una de sus carpetas.
func esDeLoComun(ruta string) bool {
	return slices.Contains(ficherosDeLoComun, ruta) || empiezaPorAlguno(ruta, directoriosDeLoComun)
}

// esPaqueteDeLoComun dice si el directorio de un paquete del módulo es de lo
// común.
func esPaqueteDeLoComun(directorio string) bool {
	return esDeLoComun(directorio + "/")
}

// empiezaPorAlguno dice si la ruta empieza por alguno de los prefijos.
func empiezaPorAlguno(ruta string, prefijos []string) bool {
	return slices.ContainsFunc(prefijos, func(prefijo string) bool {
		return strings.HasPrefix(ruta, prefijo)
	})
}

// raicesDeGrabaciones son las carpetas de las grabaciones que el arnés
// sirve a las sesiones (FR-001(e)): las de UnionDeGrabaciones y la de
// GrafosPrevios, relativas a la raíz y con la barra final.
func raicesDeGrabaciones() []string {
	raices := append(UnionDeGrabaciones(), GrafosPrevios)
	for indice, raiz := range raices {
		raices[indice] = desdeLaRaiz(raiz) + "/"
	}

	return raices
}

// desdeLaRaiz es una ruta del arnés, relativa a este paquete, relativa a la
// raíz del repositorio.
func desdeLaRaiz(ruta string) string {
	return strings.TrimPrefix(ruta, raizDelRepositorio+"/")
}

// paqueteDeGo es lo que el cálculo lee de un paquete de go list -json (research
// V1).
type paqueteDeGo struct {
	ImportPath   string
	Standard     bool
	Module       *moduloDeGo
	GoFiles      []string
	EmbedFiles   []string
	Imports      []string
	TestImports  []string
	XTestImports []string
	Error        *errorDeGo
	DepsErrors   []*errorDeGo
}

// moduloDeGo es el módulo de un paquete de go list -json.
type moduloDeGo struct {
	Path    string
	Version string
	Main    bool
}

// errorDeGo es un error de un paquete de go list -json.
type errorDeGo struct {
	Err string
}

// errores son el error del paquete y los de sus dependencias, unidos, o nil.
func (p paqueteDeGo) errores() error {
	var errores []error

	for _, deGo := range append([]*errorDeGo{p.Error}, p.DepsErrors...) {
		if deGo != nil {
			errores = append(errores, fmt.Errorf("el paquete %s: %s", p.ImportPath, deGo.Err))
		}
	}

	return errors.Join(errores...)
}

// grafoDePaquetes es el grafo de paquetes del módulo y sus dependencias, el de
// un solo go list (research D3): la ruta del módulo principal y cada paquete
// por su ruta de importación.
type grafoDePaquetes struct {
	modulo   string
	paquetes map[string]paqueteDeGo
}

// leerGrafoDePaquetes lista con go list -tags evals -deps -json -e ./..., en la
// raíz y con la plataforma del job, todos los paquetes del módulo y sus
// dependencias (research D3, V2). Es un error que la orden falle, que su salida
// no se pueda leer o que no dé ningún paquete del módulo principal.
func leerGrafoDePaquetes(ctx context.Context, raiz string) (grafoDePaquetes, error) {
	salida, err := ejecutarEn(exec.CommandContext(ctx, "go", "list", "-tags", "evals", "-deps", "-json", "-e", "./..."),
		raiz, plataformaDelJob...)
	if err != nil {
		return grafoDePaquetes{}, err
	}

	grafo := grafoDePaquetes{paquetes: map[string]paqueteDeGo{}}
	decodificador := jsontext.NewDecoder(bytes.NewReader(salida))

	for {
		var paquete paqueteDeGo

		err := json.UnmarshalDecode(decodificador, &paquete)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return grafoDePaquetes{}, fmt.Errorf("la salida de go list no se puede leer: %w", err)
		}

		grafo.paquetes[paquete.ImportPath] = paquete

		if paquete.Module != nil && paquete.Module.Main {
			grafo.modulo = paquete.Module.Path
		}
	}

	if grafo.modulo == "" {
		return grafoDePaquetes{}, errors.New("go list no da ningún paquete del módulo principal")
	}

	return grafo, nil
}

// directorio es el directorio, relativo a la raíz, del paquete del módulo con
// esa ruta de importación —«.» el del paquete raíz—, y false si la ruta no es
// del módulo.
func (g grafoDePaquetes) directorio(importacion string) (string, bool) {
	if importacion == g.modulo {
		return ".", true
	}

	return strings.CutPrefix(importacion, g.modulo+"/")
}

// cierreDePaquetes es lo que da el recorrido del grafo desde unos paquetes: las
// carpetas de los paquetes del módulo en los que entra, los ficheros que
// empotran, relativos a la raíz, y la versión de cada módulo externo.
type cierreDePaquetes struct {
	carpetas   []string
	empotrados []string
	modulos    map[string]string
}

// cierre sigue el grafo por Imports desde los paquetes de inicio: no entra en
// los de la biblioteca estándar, entra en los externos, de los que apunta el
// módulo, y en los del módulo cuyo directorio acepta entra. Es un error un
// paquete que no está en el grafo, que no es de ningún módulo o que trae Error o
// DepsErrors, entre aquellos en los que entra (FR-005; data-model §3).
func (g grafoDePaquetes) cierre(inicio []string, entra func(directorio string) bool) (cierreDePaquetes, error) {
	cierre := cierreDePaquetes{modulos: map[string]string{}}
	vistos := map[string]bool{}
	pendientes := slices.Clone(inicio)

	for len(pendientes) > 0 {
		importacion := pendientes[len(pendientes)-1]
		pendientes = pendientes[:len(pendientes)-1]

		if vistos[importacion] {
			continue
		}

		vistos[importacion] = true

		paquete, entrar, err := g.visitar(importacion, entra, &cierre)
		if err != nil {
			return cierreDePaquetes{}, err
		}

		if entrar {
			pendientes = append(pendientes, paquete.Imports...)
		}
	}

	slices.Sort(cierre.carpetas)
	slices.Sort(cierre.empotrados)

	return cierre, nil
}

// visitar apunta en el cierre el paquete con esa ruta de importación y dice si
// el recorrido entra en él, con el paquete.
func (g grafoDePaquetes) visitar(importacion string, entra func(string) bool, cierre *cierreDePaquetes) (
	paqueteDeGo, bool, error,
) {
	paquete, listado := g.paquetes[importacion]

	switch {
	case !listado:
		return paqueteDeGo{}, false, fmt.Errorf("el paquete %s no está en go list", importacion)
	case paquete.Standard:
		return paquete, false, nil
	case paquete.Module == nil:
		return paqueteDeGo{}, false, errors.Join(fmt.Errorf("el paquete %s no es de ningún módulo", importacion),
			paquete.errores())
	case !paquete.Module.Main:
		cierre.modulos[paquete.Module.Path] = paquete.Module.Version

		return paquete, true, paquete.errores()
	}

	directorio, _ := g.directorio(importacion)
	if !entra(directorio) {
		return paquete, false, nil
	}

	cierre.carpetas = append(cierre.carpetas, directorio)
	for _, empotrado := range paquete.EmbedFiles {
		cierre.empotrados = append(cierre.empotrados, path.Join(directorio, empotrado))
	}

	return paquete, true, paquete.errores()
}

// entraEnUnApplet dice si el cierre de un applet entra en el paquete del
// módulo de ese directorio: en ninguno de lo común ni en el raíz (FR-001(d):
// lo común no entra por (d), y el raíz llega a un conjunto solo por (b)).
func entraEnUnApplet(directorio string) bool {
	return directorio != "." && !esPaqueteDeLoComun(directorio)
}

// entraEnLoComun dice si el cierre de lo común entra en el paquete del módulo
// de ese directorio: solo en los de lo común y en el raíz (research D4).
func entraEnLoComun(directorio string) bool {
	return directorio == "." || esPaqueteDeLoComun(directorio)
}

// loComun es lo que el cálculo da de lo común (FR-002; data-model §2), además
// de sus rutas (esDeLoComun) y de go y toolchain, que están en el árbol: las
// carpetas de los paquetes del módulo que son de lo común y la versión de
// cada módulo externo del que dependen.
type loComun struct {
	paquetes []string
	modulos  map[string]string
}

// loComun calcula lo común sobre el grafo: sus paquetes, y sus módulos, los del
// cierre desde ellos, desde el paquete raíz y desde las importaciones de los
// tests de internal/evals, sin entrar en un paquete del módulo que no es de lo
// común (research D4). Una importación de esos tests que no está en el grafo
// —go list -deps no lista lo que solo importan tests— es de la biblioteca
// estándar o del módulo de go.mod, de los requeridos, cuya ruta la contiene; es
// un error que no sea de ninguno.
func (g grafoDePaquetes) loComun(requeridos map[string]string) (loComun, error) {
	var comun loComun

	inicio := []string{g.modulo}

	for _, importacion := range slices.Sorted(maps.Keys(g.paquetes)) {
		if directorio, delModulo := g.directorio(importacion); delModulo && esPaqueteDeLoComun(directorio) {
			comun.paquetes = append(comun.paquetes, directorio)
			inicio = append(inicio, importacion)
		}
	}

	arnes := g.paquetes[g.modulo+"/"+directorioDelArnes]
	sinListar := map[string]string{}

	for _, importacion := range slices.Concat(arnes.TestImports, arnes.XTestImports) {
		if _, listado := g.paquetes[importacion]; listado {
			inicio = append(inicio, importacion)

			continue
		}

		modulo, esDeUnModulo, err := moduloDeUnaImportacion(importacion, requeridos)
		if err != nil {
			return loComun{}, fmt.Errorf("los tests de %s: %w", directorioDelArnes, err)
		}

		if esDeUnModulo {
			sinListar[modulo] = requeridos[modulo]
		}
	}

	cierre, err := g.cierre(inicio, entraEnLoComun)
	if err != nil {
		return loComun{}, fmt.Errorf("lo común: %w", err)
	}

	comun.modulos = cierre.modulos
	maps.Copy(comun.modulos, sinListar)

	return comun, nil
}

// moduloDeUnaImportacion es el módulo de go.mod, de los requeridos, que
// contiene la importación, el de ruta más larga; false si la importación es de
// la biblioteca estándar, cuyo primer elemento no lleva punto. Es un error que
// no sea de la biblioteca estándar ni de ningún módulo requerido.
func moduloDeUnaImportacion(importacion string, requeridos map[string]string) (string, bool, error) {
	if primero, _, _ := strings.Cut(importacion, "/"); !strings.Contains(primero, ".") {
		return "", false, nil
	}

	modulo := ""

	for requerido := range requeridos {
		contiene := importacion == requerido || strings.HasPrefix(importacion, requerido+"/")
		if contiene && len(requerido) > len(modulo) {
			modulo = requerido
		}
	}

	if modulo == "" {
		return "", false, fmt.Errorf("%s no es de ningún módulo de go.mod", importacion)
	}

	return modulo, true, nil
}

// appletsDeLaComposicion son, por applet, los paquetes del módulo de los que
// parte su cierre: las importaciones del módulo, fuera de lo común y sin el
// raíz, de cada fichero de internal/app cuyo Nombre() devuelve el nombre del
// applet como literal (research D3, V7). Los ficheros son los que compila go
// list con la plataforma del job, leídos del árbol; es un error que no se
// puedan leer o analizar.
func (g grafoDePaquetes) appletsDeLaComposicion(arbol fs.FS) (map[string][]string, error) {
	deLaComposicion, listado := g.paquetes[g.modulo+"/"+directorioDeLaComposicion]
	if !listado {
		return nil, fmt.Errorf("el paquete %s no está en go list", directorioDeLaComposicion)
	}

	applets := map[string][]string{}

	for _, nombre := range deLaComposicion.GoFiles {
		ruta := directorioDeLaComposicion + "/" + nombre

		contenido, err := fs.ReadFile(arbol, ruta)
		if err != nil {
			return nil, fmt.Errorf("%s no se puede leer: %w", ruta, err)
		}

		fichero, err := parser.ParseFile(token.NewFileSet(), ruta, contenido, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("%s no se puede analizar: %w", ruta, err)
		}

		if applet, declara := appletDelFichero(fichero); declara {
			applets[applet] = append(applets[applet], g.importacionesDeUnApplet(fichero)...)
		}
	}

	return applets, nil
}

// importacionesDeUnApplet son las importaciones del fichero que son paquetes
// del módulo en los que entra el cierre de un applet.
func (g grafoDePaquetes) importacionesDeUnApplet(fichero *ast.File) []string {
	var importaciones []string

	for _, importacion := range fichero.Imports {
		ruta, err := strconv.Unquote(importacion.Path.Value)
		if err != nil {
			continue
		}

		if directorio, delModulo := g.directorio(ruta); delModulo && entraEnUnApplet(directorio) {
			importaciones = append(importaciones, ruta)
		}
	}

	return importaciones
}

// appletDelFichero es el nombre que declara el fichero con un método Nombre()
// cuyo cuerpo solo devuelve un literal de cadena, y false si no lo declara.
func appletDelFichero(fichero *ast.File) (string, bool) {
	for _, nodo := range fichero.Decls {
		if funcion, esFuncion := nodo.(*ast.FuncDecl); esFuncion && funcion.Recv != nil &&
			funcion.Name.Name == metodoDelNombre {
			if nombre, esLiteral := literalDevuelto(funcion); esLiteral {
				return nombre, true
			}
		}
	}

	return "", false
}

// literalDevuelto es la cadena que devuelve la función, si su cuerpo es solo
// un return de un literal de cadena.
func literalDevuelto(funcion *ast.FuncDecl) (string, bool) {
	if funcion.Body == nil || len(funcion.Body.List) != 1 {
		return "", false
	}

	devolucion, esDevolucion := funcion.Body.List[0].(*ast.ReturnStmt)
	if !esDevolucion || len(devolucion.Results) != 1 {
		return "", false
	}

	literal, esLiteral := devolucion.Results[0].(*ast.BasicLit)
	if !esLiteral || literal.Kind != token.STRING {
		return "", false
	}

	nombre, err := strconv.Unquote(literal.Value)

	return nombre, err == nil
}

// calculoDeDependencias es lo que comparten los conjuntos de todas las evals de
// un árbol: el árbol, del que lee sus ficheros, el grafo de paquetes, las
// skills de skills/ y, por applet, los paquetes de los que parte su cierre.
type calculoDeDependencias struct {
	arbol   fs.FS
	grafo   grafoDePaquetes
	skills  []string
	applets map[string][]string
}

// calcularDependencias calcula sobre el árbol y su grafo de paquetes el
// conjunto de dependencias de cada eval de evals/<skill>/, en orden de skill y
// de fichero (FR-001; data-model §3). Una parte que no se puede calcular es el
// error del conjunto de cada eval a la que afecta, con su nombre delante
// (FR-005), y quien llama decide: make ci falla, y el job elige todas con
// sin_conjunto. El error devuelto es el de lo que impide calcular cualquier
// conjunto: listar skills/ o evals/, o leer la composición.
func calcularDependencias(arbol fs.FS, grafo grafoDePaquetes) ([]dependenciasDeLaEval, error) {
	skillsDelArbol, err := directoriosDe(arbol, directorioDeLasSkills)
	if err != nil {
		return nil, err
	}

	applets, err := grafo.appletsDeLaComposicion(arbol)
	if err != nil {
		return nil, err
	}

	calculo := calculoDeDependencias{arbol: arbol, grafo: grafo, skills: skillsDelArbol, applets: applets}

	skillsDeLasEvals, err := directoriosDe(arbol, directorioDeLasEvals)
	if err != nil {
		return nil, err
	}

	var conjuntos []dependenciasDeLaEval

	for _, skill := range skillsDeLasEvals {
		evals, juez, err := calculo.directorioDeEvals(skill)
		if err != nil {
			return nil, err
		}

		comunes, modulos, errDeLaSkill := calculo.deLaSkill(skill)
		comunes = append(comunes, juez...)

		for _, fichero := range evals {
			conjuntos = append(conjuntos, calculo.deLaEval(skill, fichero, comunes, modulos, errDeLaSkill))
		}
	}

	return conjuntos, nil
}

// directoriosDe son los nombres de los subdirectorios del directorio del
// árbol, en orden.
func directoriosDe(arbol fs.FS, directorio string) ([]string, error) {
	entradas, err := fs.ReadDir(arbol, directorio)
	if err != nil {
		return nil, fmt.Errorf("%s/ no se puede listar: %w", directorio, err)
	}

	var carpetas []string

	for _, entrada := range entradas {
		if entrada.IsDir() {
			carpetas = append(carpetas, entrada.Name())
		}
	}

	return carpetas, nil
}

// directorioDeEvals lee evals/<skill>/: los nombres de sus evals, en orden, y
// la parte (f) de todas: la carpeta del juez, si está, y cada fichero que no es
// una eval.
func (c calculoDeDependencias) directorioDeEvals(skill string) ([]string, []entradaDeDependencia, error) {
	directorio := directorioDeLasEvals + "/" + skill

	entradas, err := fs.ReadDir(c.arbol, directorio)
	if err != nil {
		return nil, nil, fmt.Errorf("%s/ no se puede listar: %w", directorio, err)
	}

	var (
		evals []string
		juez  []entradaDeDependencia
	)

	for _, entrada := range entradas {
		esEval, err := path.Match(formaDeUnaEval, entrada.Name())
		if err != nil {
			return nil, nil, fmt.Errorf("la forma de una eval: %w", err)
		}

		switch {
		case entrada.IsDir():
			if entrada.Name() == carpetaDelJuez {
				juez = append(juez, entradaDeDependencia{
					clase: entradaDirectorio, ruta: directorio + "/" + carpetaDelJuez + "/", parte: parteDelJuez,
				})
			}
		case esEval:
			evals = append(evals, entrada.Name())
		default:
			juez = append(juez, entradaDeDependencia{
				clase: entradaFichero, ruta: directorio + "/" + entrada.Name(), parte: parteDelJuez,
			})
		}
	}

	return evals, juez, nil
}

// deLaEval es el conjunto de una eval: (a) su fichero, lo que tiene en común con las
// demás evals de su skill y (e) sus grabaciones; o el error de su skill, o el
// de su fichero, que no se puede leer como eval, con su nombre delante.
func (c calculoDeDependencias) deLaEval(skill, fichero string, comunes []entradaDeDependencia,
	modulos map[string]string, errDeLaSkill error,
) dependenciasDeLaEval {
	conjunto := dependenciasDeLaEval{skill: skill, eval: strings.TrimSuffix(fichero, ".yaml")}
	if errDeLaSkill != nil {
		conjunto.err = fmt.Errorf("%s: %w", conjunto.nombre(), errDeLaSkill)

		return conjunto
	}

	ruta := directorioDeLasEvals + "/" + skill + "/" + fichero

	grabaciones, err := c.grabacionesDeLaEval(ruta, fichero)
	if err != nil {
		conjunto.err = fmt.Errorf("%s: %w", conjunto.nombre(), err)

		return conjunto
	}

	entradas := slices.Concat(
		[]entradaDeDependencia{{clase: entradaFichero, ruta: ruta, parte: parteDeLaEval}}, comunes, grabaciones)
	conjunto.entradas = ordenadas(entradas)
	conjunto.modulos = maps.Clone(modulos)

	return conjunto
}

// grabacionesDeLaEval es la parte (e) del conjunto de la eval: las raíces de
// UnionDeGrabaciones y, si la eval lleva grafo previo, su directorio de
// GrafosPrevios. Es un error que el fichero no se pueda leer como eval.
func (c calculoDeDependencias) grabacionesDeLaEval(ruta, fichero string) ([]entradaDeDependencia, error) {
	contenido, err := fs.ReadFile(c.arbol, ruta)
	if err != nil {
		return nil, fmt.Errorf("%s no se puede leer: %w", ruta, err)
	}

	eval, err := LeerEval(fichero, contenido)
	if err != nil {
		return nil, fmt.Errorf("%s no se puede leer como eval: %w", ruta, err)
	}

	var grabaciones []entradaDeDependencia

	for _, raiz := range UnionDeGrabaciones() {
		grabaciones = append(grabaciones, entradaDeDependencia{
			clase: entradaDirectorio, ruta: desdeLaRaiz(raiz) + "/", parte: parteDeLasGrabaciones,
		})
	}

	if previo := eval.GrafoPrevio.Grabaciones; previo != "" {
		grabaciones = append(grabaciones, entradaDeDependencia{
			clase: entradaDirectorio, ruta: desdeLaRaiz(GrafosPrevios) + "/" + previo + "/", parte: parteDeLasGrabaciones,
		})
	}

	return grabaciones, nil
}

// deLaSkill son las partes (b), (c) y (d) que comparten las evals de la skill,
// y la versión de los módulos de sus paquetes; o el error que impide
// calcularlas: que no exista skills/<skill>/, que su SKILL.md o sus referencias
// a mano, o las de una skill que nombra, no se puedan leer, o el de alguna de
// sus partes.
func (c calculoDeDependencias) deLaSkill(skill string) ([]entradaDeDependencia, map[string]string, error) {
	if !slices.Contains(c.skills, skill) {
		return nil, nil, fmt.Errorf("no existe %s/%s/", directorioDeLasSkills, skill)
	}

	nombradas, leidas, err := c.skillsDeLaEval(skill)
	if err != nil {
		return nil, nil, err
	}

	var (
		entradas []entradaDeDependencia
		inicio   []string
	)

	for _, nombrada := range nombradas {
		deLaNombrada, deSusApplets, err := c.deUnaSkill(nombrada, leidas[nombrada])
		if err != nil {
			return nil, nil, err
		}

		entradas = append(entradas, deLaNombrada...)
		inicio = append(inicio, deSusApplets...)
	}

	cierre, err := c.grafo.cierre(inicio, entraEnUnApplet)
	if err != nil {
		return nil, nil, err
	}

	for _, directorio := range cierre.carpetas {
		entradas = append(entradas, entradaDeDependencia{clase: entradaPaquete, ruta: directorio, parte: parteDeLosPaquetes})
	}

	for _, empotrado := range cierre.empotrados {
		entradas = append(entradas, entradaDeDependencia{clase: entradaFichero, ruta: empotrado, parte: parteDeLosPaquetes})
	}

	return entradas, cierre.modulos, nil
}

// skillDelConjunto es lo que se lee de una skill del conjunto de una eval: su
// SKILL.md, del que sale su tabla de comandos, y sus referencias a mano
// (skills.ReferenciasAMano), cada fichero por separado.
type skillDelConjunto struct {
	skillMD []byte
	aMano   []skills.Referencia
}

// laNombra dice si el SKILL.md o alguna referencia a mano de la skill nombra la
// otra, en un mismo fichero (FR-014 y Clarificación Q1 de H31). Las referencias
// generadas no se leen: salen de data/, que entra por (c).
func (s skillDelConjunto) laNombra(otra string) bool {
	return nombra(s.skillMD, otra) || slices.ContainsFunc(s.aMano, func(referencia skills.Referencia) bool {
		return nombra(referencia.Contenido, otra)
	})
}

// skillsDeLaEval son las skills de una eval de la skill (FR-001(b)): ella y,
// por cierre, cada skill de skills/ cuyo nombre aparece en el SKILL.md o en una
// referencia a mano de otra del conjunto entre límites que no son [a-z0-9-]; en
// el orden en que se encuentran, con lo leído de cada una. Es un error que un
// SKILL.md o sus referencias a mano no se puedan leer.
func (c calculoDeDependencias) skillsDeLaEval(skill string) ([]string, map[string]skillDelConjunto, error) {
	var nombradas []string

	leidas := map[string]skillDelConjunto{}
	pendientes := []string{skill}

	for len(pendientes) > 0 {
		actual := pendientes[0]
		pendientes = pendientes[1:]

		if _, leida := leidas[actual]; leida {
			continue
		}

		ruta := directorioDeLasSkills + "/" + actual + "/" + ficheroDeLaSkill

		contenido, err := fs.ReadFile(c.arbol, ruta)
		if err != nil {
			return nil, nil, fmt.Errorf("%s no se puede leer: %w", ruta, err)
		}

		aMano, err := skills.ReferenciasAMano(c.arbol, actual)
		if err != nil {
			return nil, nil, fmt.Errorf("las referencias a mano de %s/%s: %w", directorioDeLasSkills, actual, err)
		}

		leidas[actual] = skillDelConjunto{skillMD: contenido, aMano: aMano}
		nombradas = append(nombradas, actual)

		for _, otra := range c.skills {
			if otra != actual && leidas[actual].laNombra(otra) {
				pendientes = append(pendientes, otra)
			}
		}
	}

	return nombradas, leidas, nil
}

// nombra dice si el nombre aparece en el contenido entre límites que no son
// [a-z0-9-]: el principio o el final del contenido, o un byte de otra clase.
func nombra(contenido []byte, nombre string) bool {
	for desde := 0; desde < len(contenido); {
		indice := bytes.Index(contenido[desde:], []byte(nombre))
		if indice < 0 {
			return false
		}

		inicio, fin := desde+indice, desde+indice+len(nombre)
		limiteAntes := inicio == 0 || !esDeUnNombre(contenido[inicio-1])
		limiteDespues := fin == len(contenido) || !esDeUnNombre(contenido[fin])

		if limiteAntes && limiteDespues {
			return true
		}

		desde = inicio + 1
	}

	return false
}

// esDeUnNombre dice si el byte es de [a-z0-9-], los del nombre de una skill.
func esDeUnNombre(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b == '-'
}

// deUnaSkill son, de una skill de la eval, su SKILL.md y sus referencias (b),
// las generadas y las a mano; los data/*.yaml de las cabeceras de las generadas
// (c); y los paquetes de los que parte el cierre de los applets de su tabla de
// comandos (d). Es un error una referencia que no es a mano sin la cabecera
// «generado desde», una tabla sin sus marcas o sin applets, o un applet de la
// tabla sin fichero en internal/app (FR-005).
func (c calculoDeDependencias) deUnaSkill(skill string, leida skillDelConjunto,
) ([]entradaDeDependencia, []string, error) {
	directorio := directorioDeLasSkills + "/" + skill
	referencias := directorio + "/" + directorioDeLasReferencias

	entradas := []entradaDeDependencia{
		{clase: entradaFichero, ruta: directorio + "/" + ficheroDeLaSkill, parte: parteDeLasSkills},
		{clase: entradaDirectorio, ruta: referencias + "/", parte: parteDeLasSkills},
	}

	datos, err := c.datosDeLasReferencias(referencias, leida.aMano)
	if err != nil {
		return nil, nil, err
	}

	entradas = append(entradas, datos...)

	applets, err := skills.AppletsDeLaTabla(leida.skillMD)
	if err != nil {
		return nil, nil, fmt.Errorf("la tabla de comandos de %s/%s: %w", directorio, ficheroDeLaSkill, err)
	}

	var inicio []string

	for _, applet := range applets {
		importaciones, declarado := c.applets[applet]
		if !declarado {
			return nil, nil, fmt.Errorf("el applet %q de la tabla de comandos de %s/%s no tiene fichero en %s",
				applet, directorio, ficheroDeLaSkill, directorioDeLaComposicion)
		}

		inicio = append(inicio, importaciones...)
	}

	return entradas, inicio, nil
}

// datosDeLasReferencias son los data/*.yaml de las cabeceras de las
// referencias del directorio, uno por fichero que no está entre las a mano
// dadas (FR-001(c); FR-014 de H31): una referencia a mano no da parte (c).
// Ninguno si el directorio no existe. Es un error que no se pueda listar o
// leer, o que una referencia que no es a mano no tenga la cabecera.
func (c calculoDeDependencias) datosDeLasReferencias(referencias string, aMano []skills.Referencia,
) ([]entradaDeDependencia, error) {
	ficheros, err := fs.ReadDir(c.arbol, referencias)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("%s/ no se puede listar: %w", referencias, err)
	}

	var datos []entradaDeDependencia

	for _, fichero := range ficheros {
		esAMano := slices.ContainsFunc(aMano, func(referencia skills.Referencia) bool {
			return referencia.Fichero == fichero.Name()
		})
		if fichero.IsDir() || esAMano {
			continue
		}

		ruta := referencias + "/" + fichero.Name()

		contenido, err := fs.ReadFile(c.arbol, ruta)
		if err != nil {
			return nil, fmt.Errorf("%s no se puede leer: %w", ruta, err)
		}

		deDatos, generada := skills.DatosDeLaReferencia(contenido)
		if !generada {
			return nil, fmt.Errorf("la referencia %s no tiene la cabecera «generado desde»", ruta)
		}

		datos = append(datos, entradaDeDependencia{clase: entradaFichero, ruta: deDatos, parte: parteDeLosDatos})
	}

	return datos, nil
}

// ordenadas son las entradas en orden de parte, de clase y de ruta, sin
// repetir.
func ordenadas(entradas []entradaDeDependencia) []entradaDeDependencia {
	entradas = slices.Clone(entradas)
	slices.SortFunc(entradas, func(una, otra entradaDeDependencia) int {
		return strings.Compare(string(una.parte)+" "+string(una.clase)+" "+una.ruta,
			string(otra.parte)+" "+string(otra.clase)+" "+otra.ruta)
	})

	return slices.Compact(entradas)
}

// arbolDelRepositorio es lo que la elección y la huella leen de un árbol, el de
// la cabeza o el de la base de la propuesta (data-model §4).
type arbolDelRepositorio struct {
	// ficheros son las rutas que miden, con el sha256 de su contenido en
	// hexadecimal; solo en la cabeza.
	ficheros map[string]string

	// descripciones son el description de la cabecera del SKILL.md de cada
	// skill, por su nombre.
	descripciones map[string]string

	// skills y directoriosDeEvals son lo que va justo debajo de skills/ y de
	// evals/ en las rutas, ordenado y sin repetir (primerComponente).
	skills             []string
	directoriosDeEvals []string

	// Los requeridos, go y toolchain de su go.mod.
	versionesDeGoMod
}

// versionesDeGoMod es lo que fija un go.mod, leído con go mod edit -json: la
// versión de cada módulo requerido, go y toolchain.
type versionesDeGoMod struct {
	modulos     map[string]string
	versionDeGo string
	toolchain   string
}

// leerArbolDeLaCabeza lee el árbol de trabajo de la raíz: las rutas de git
// ls-files -z --cached --others --exclude-standard que están en el disco —sin
// las ignoradas ni las borradas del árbol de trabajo— y, de las que miden con
// los paquetes de los conjuntos dados, el sha256 de su contenido (data-model
// §4; research D6).
func leerArbolDeLaCabeza(ctx context.Context, raiz string, paquetes []string) (arbolDelRepositorio, error) {
	salida, err := ejecutarEn(
		exec.CommandContext(ctx, "git", "ls-files", "-z", "--cached", "--others", "--exclude-standard"), raiz)
	if err != nil {
		return arbolDelRepositorio{}, err
	}

	disco := os.DirFS(raiz)
	rutas := slices.DeleteFunc(rutasDeGit(salida), func(ruta string) bool {
		_, err := fs.Stat(disco, ruta)

		return errors.Is(err, fs.ErrNotExist)
	})

	arbol, err := completarArbol(ctx, raiz, rutas, func(ruta string) ([]byte, error) {
		return fs.ReadFile(disco, ruta)
	})
	if err != nil {
		return arbolDelRepositorio{}, err
	}

	arbol.ficheros = map[string]string{}

	for _, ruta := range rutas {
		if !rutaQueMide(ruta, paquetes) {
			continue
		}

		contenido, err := fs.ReadFile(disco, ruta)
		if err != nil {
			return arbolDelRepositorio{}, fmt.Errorf("%s no se puede leer: %w", ruta, err)
		}

		suma := sha256.Sum256(contenido)
		arbol.ficheros[ruta] = hex.EncodeToString(suma[:])
	}

	return arbol, nil
}

// baseDeFusion es la base de fusión de la base dada y HEAD, con git merge-base
// en la raíz (research D5).
func baseDeFusion(ctx context.Context, raiz, base string) (string, error) {
	salida, err := ejecutarEn(exec.CommandContext(ctx, "sh", "-c", ordenDeLaFusion), raiz, variableDeLaBase+"="+base)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(salida)), nil
}

// leerArbolDeLaBase lee el árbol del commit dado, sin ficheros: las rutas de
// skills/ y evals/ de git ls-tree, y lo que se lee de ellas y de go.mod con git
// show (data-model §4; research D5).
func leerArbolDeLaBase(ctx context.Context, raiz, base string) (arbolDelRepositorio, error) {
	variable := variableDeLaBase + "=" + base

	salida, err := ejecutarEn(exec.CommandContext(ctx, "sh", "-c", ordenDeLasRutas), raiz, variable)
	if err != nil {
		return arbolDelRepositorio{}, err
	}

	return completarArbol(ctx, raiz, rutasDeGit(salida), func(ruta string) ([]byte, error) {
		return ejecutarEn(exec.CommandContext(ctx, "sh", "-c", ordenDeUnFichero), raiz, variable,
			variableDelFichero+"="+ruta)
	})
}

// rutasDeGit son las rutas de una salida de git con -z, ordenadas y sin
// repetir: una ruta en conflicto sale una vez por etapa.
func rutasDeGit(salida []byte) []string {
	rutas := strings.Split(string(salida), "\x00")
	rutas = slices.DeleteFunc(rutas, func(ruta string) bool { return ruta == "" })
	slices.Sort(rutas)

	return slices.Compact(rutas)
}

// completarArbol es el árbol de las rutas dadas, sin ficheros: sus skills, sus
// carpetas de evals, la descripción de cada SKILL.md de skills/<skill>/ y lo
// que fija go.mod, leídos con leer. Es un error que un SKILL.md o go.mod no se
// puedan leer.
func completarArbol(ctx context.Context, raiz string, rutas []string, leer func(ruta string) ([]byte, error)) (
	arbolDelRepositorio, error,
) {
	arbol := arbolDelRepositorio{descripciones: map[string]string{}}

	for _, ruta := range rutas {
		if skill, bajoSkills := primerComponente(ruta, directorioDeLasSkills); bajoSkills {
			arbol.skills = append(arbol.skills, skill)
		}

		if directorio, bajoEvals := primerComponente(ruta, directorioDeLasEvals); bajoEvals {
			arbol.directoriosDeEvals = append(arbol.directoriosDeEvals, directorio)
		}

		if skill, esElSuyo := skillDelFichero(ruta); esElSuyo {
			descripcion, err := descripcionDeLaSkill(ruta, leer)
			if err != nil {
				return arbolDelRepositorio{}, err
			}

			arbol.descripciones[skill] = descripcion
		}
	}

	slices.Sort(arbol.skills)
	slices.Sort(arbol.directoriosDeEvals)
	arbol.skills = slices.Compact(arbol.skills)
	arbol.directoriosDeEvals = slices.Compact(arbol.directoriosDeEvals)

	goMod, err := leer(nombreDelGoMod)
	if err != nil {
		return arbolDelRepositorio{}, fmt.Errorf("%s no se puede leer: %w", nombreDelGoMod, err)
	}

	arbol.versionesDeGoMod, err = modulosDeGoMod(ctx, raiz, goMod)
	if err != nil {
		return arbolDelRepositorio{}, err
	}

	return arbol, nil
}

// primerComponente es el primer componente de la ruta bajo el directorio, y
// false si la ruta no está bajo él.
func primerComponente(ruta, directorio string) (string, bool) {
	resto, bajo := strings.CutPrefix(ruta, directorio+"/")
	if !bajo || resto == "" {
		return "", false
	}

	componente, _, _ := strings.Cut(resto, "/")

	return componente, true
}

// skillDelFichero es la skill de la ruta, si la ruta es su SKILL.md,
// skills/<skill>/SKILL.md.
func skillDelFichero(ruta string) (string, bool) {
	resto, bajoSkills := strings.CutPrefix(ruta, directorioDeLasSkills+"/")
	skill, esElSuyo := strings.CutSuffix(resto, "/"+ficheroDeLaSkill)

	return skill, bajoSkills && esElSuyo && skill != "" && !strings.Contains(skill, "/")
}

// descripcionDeLaSkill es el description de la cabecera del SKILL.md de la
// ruta, leída con leer.
func descripcionDeLaSkill(ruta string, leer func(ruta string) ([]byte, error)) (string, error) {
	contenido, err := leer(ruta)
	if err != nil {
		return "", fmt.Errorf("%s no se puede leer: %w", ruta, err)
	}

	frontmatter, err := skills.LeerFrontmatter(contenido)
	if err != nil {
		return "", fmt.Errorf("la cabecera de %s: %w", ruta, err)
	}

	return frontmatter.Description, nil
}

// goModLeido es lo que el cálculo lee de go mod edit -json.
type goModLeido struct {
	Go        string
	Toolchain string
	Require   []struct {
		Path    string
		Version string
	}
}

// modulosDeGoMod es lo que fija el contenido de un go.mod, leído con go mod
// edit -json en la raíz (research D5, V3).
func modulosDeGoMod(ctx context.Context, raiz string, contenido []byte) (versionesDeGoMod, error) {
	salida, err := goModEnJSON(ctx, raiz, contenido)
	if err != nil {
		return versionesDeGoMod{}, err
	}

	var goMod goModLeido
	if err := json.Unmarshal(salida, &goMod); err != nil {
		return versionesDeGoMod{}, fmt.Errorf("la salida de go mod edit -json no se puede leer: %w", err)
	}

	versiones := versionesDeGoMod{
		modulos: make(map[string]string, len(goMod.Require)), versionDeGo: goMod.Go, toolchain: goMod.Toolchain,
	}
	for _, requerido := range goMod.Require {
		versiones.modulos[requerido.Path] = requerido.Version
	}

	return versiones, nil
}

// goModEnJSON es la salida de go mod edit -json, en la raíz, sobre una copia
// del contenido de un go.mod en un temporal, que se borra al terminar: así se
// lee el de la base sin tocar el del árbol.
func goModEnJSON(ctx context.Context, raiz string, contenido []byte) (_ []byte, err error) {
	temporal, err := os.MkdirTemp("", patronDelTemporal)
	if err != nil {
		return nil, fmt.Errorf("el temporal de %s no se puede crear: %w", nombreDelGoMod, err)
	}

	defer func() { err = errors.Join(err, retirarTemporal(temporal)) }()

	ruta := filepath.Join(temporal, nombreDelGoMod)
	if err := escribirFichero(ruta, contenido); err != nil {
		return nil, fmt.Errorf("%s no se puede copiar al temporal: %w", nombreDelGoMod, err)
	}

	return ejecutarEn(exec.CommandContext(ctx, "sh", "-c", ordenDelGoMod), raiz, variableDelGoMod+"="+ruta)
}

// ejecutarEn ejecuta la orden en el directorio, con el entorno del proceso y,
// encima, las variables dadas, y devuelve lo que escribió en su salida
// estándar. Si no termina con 0, devuelve un error con la orden, las variables,
// cómo terminó y lo que escribió en su salida de error.
func ejecutarEn(orden *exec.Cmd, directorio string, variables ...string) ([]byte, error) {
	var salida, errores bytes.Buffer

	orden.Dir = directorio
	orden.Env = sobreLaBase(os.Environ(), variables)
	orden.Stdout, orden.Stderr = &salida, &errores

	if err := orden.Run(); err != nil {
		return nil, fmt.Errorf("%s, con %v: %w\n%s", orden, variables, err, errores.Bytes())
	}

	return salida.Bytes(), nil
}
