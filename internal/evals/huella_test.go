package evals

import (
	"errors"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fijadosDelTest son los modelos y las versiones con que el test calcula las
// huellas, con la forma de los que el job fija en evals.yml.
var fijadosDelTest = fijadosDeLaHuella{
	modeloQueDecide: "claude-sonnet-5-5", claudeCode: "2.1.284", modeloDelJuez: "claude-opus-5-5",
	claudeCodeDelJuez: "2.1.289",
}

// alcanceDeUnCambio es a cuántas evals elige un cambio de una ruta: a ninguna,
// a todas o a algunas, que no son todas.
type alcanceDeUnCambio string

// Los tres alcances.
const (
	eligeNinguna alcanceDeUnCambio = "ninguna"
	eligeTodas   alcanceDeUnCambio = "todas"
	eligeAlgunas alcanceDeUnCambio = "algunas"
)

// patronDeUnaHuella es la forma de una huella: 64 caracteres hexadecimales.
var patronDeUnaHuella = regexp.MustCompile(`\A[0-9a-f]{64}\z`)

// TestHuellaDeLasEvals calcula la huella de cada eval (FR-009, SC-008;
// data-model §6; contracts/eleccion-y-huella.md §4) sobre una copia del
// repositorio en un temporal y comprueba, cada cosa en su subtest:
//
//   - por cada clase de ruta de FR-001 a FR-004 —el fichero de una eval, una
//     referencia, un data/*.yaml, un paquete de un applet, una grabación, el
//     juez, la descripción de una skill, lo común, un paquete sin asignar, la
//     versión en go.mod de un módulo de lo común, un replace en go.mod y una
//     línea de go.sum de un módulo de ningún conjunto, y un fichero de docs/—,
//     que el cambio en el disco de la copia cambia la huella de las evals que
//     elige, según la elección de la propuesta con ese cambio (data-model §5),
//     y no la de las demás, y que elige las que dice la clase: ninguna, todas
//     o algunas;
//   - que el modelo que decide y la versión de Claude Code de las sesiones
//     cambian todas, y el modelo del juez y su versión, las de las skills con
//     juez;
//   - el texto canónico de cada eval, con la forma de contracts/eleccion-y-huella.md §4;
//   - que una eval sin conjunto o un fijado vacío no dan huella;
//   - y que dos clones de la copia, cada uno en su carpeta, dan las mismas.
//
// Cada subtest que cambia la copia lo hace en su propio árbol, de enlaces
// duros a los ficheros de la copia, y sustituye el fichero que cambia en vez
// de escribir en él, para no tocar la copia. El grafo de paquetes es el de la
// copia: ningún cambio toca una importación, y con otra versión en go.mod go
// list pediría la red.
func TestHuellaDeLasEvals(t *testing.T) {
	t.Parallel()

	copia := copiaDelRepositorio(t)

	grafo, err := leerGrafoDePaquetes(t.Context(), copia)
	require.NoError(t, err)

	calculo, err := leerElCalculoDeLaHuella(t.Context(), copia, grafo, fijadosDelTest)
	require.NoError(t, err)

	huellas, err := calculo.huellas()
	require.NoError(t, err)

	repositorio := os.DirFS(copia)
	evals := nombresDeLasEvals(t, repositorio)
	require.Equal(t, evals, nombresDeLosConjuntos(calculo.conjuntos), "una huella por eval, en su orden")
	require.Len(t, huellas, len(evals))

	for _, huella := range huellas {
		require.Regexp(t, patronDeUnaHuella, huella)
	}

	casos := []struct {
		nombre  string
		mutante mutanteDeLaPropuesta
		alcance alcanceDeUnCambio
	}{
		{
			nombre:  "fichero-de-la-eval",
			mutante: conUnaLineaMas(t, repositorio, "evals/jurisprudencia/03-por-materia.yaml"),
			alcance: eligeAlgunas,
		},
		{
			nombre:  "referencia",
			mutante: conUnaLineaMas(t, repositorio, ficheroDelRepositorio(t, repositorio, "skills/legal-core/references", esMarkdown)),
			alcance: eligeAlgunas,
		},
		{
			nombre:  "datos",
			mutante: conUnaLineaMas(t, repositorio, entradaDeAlgunas(t, calculo.conjuntos, parteDeLosDatos, entradaFichero)),
			alcance: eligeAlgunas,
		},
		{
			nombre: "paquete-de-un-applet",
			mutante: conUnaLineaMas(t, repositorio, ficheroDelRepositorio(t, repositorio,
				entradaDeAlgunas(t, calculo.conjuntos, parteDeLosPaquetes, entradaPaquete), esCodigoSinTests)),
			alcance: eligeAlgunas,
		},
		{
			nombre: "grabacion",
			mutante: conUnaLineaMas(t, repositorio, ficheroBajoLaEntrada(t, repositorio,
				entradaDeAlgunas(t, calculo.conjuntos, parteDeLasGrabaciones, entradaDirectorio))),
			alcance: eligeAlgunas,
		},
		{
			nombre: "juez",
			mutante: conUnaLineaMas(t, repositorio, ficheroBajoLaEntrada(t, repositorio,
				entradaDeAlgunas(t, calculo.conjuntos, parteDelJuez, entradaDirectorio))),
			alcance: eligeAlgunas,
		},
		{nombre: "descripcion", mutante: conOtraDescripcionDeLegalCore(t, repositorio), alcance: eligeTodas},
		{
			nombre:  "lo-comun",
			mutante: conUnaLineaMas(t, repositorio, ficheroDelRepositorio(t, repositorio, "internal/cli", esCodigoSinTests)),
			alcance: eligeTodas,
		},
		{nombre: "paquete-sin-asignar", mutante: conUnPaqueteNuevo(t, repositorio), alcance: eligeTodas},
		{
			nombre:  "go-mod-de-lo-comun",
			mutante: conOtraVersionDeSqliteDeLoComun(t, repositorio, calculo),
			alcance: eligeTodas,
		},
		{nombre: "replace-en-go-mod", mutante: conUnReplaceEnGoMod(t, repositorio, calculo), alcance: eligeTodas},
		{nombre: "go-sum-de-otro-modulo", mutante: conUnaLineaEnGoSum(t, repositorio, calculo), alcance: eligeTodas},
		{
			nombre: "docs",
			mutante: conUnaLineaMas(t, repositorio, ficheroDelRepositorio(t, repositorio, "docs", func(nombre string) bool {
				return strings.HasSuffix(nombre, ".md")
			})),
			alcance: eligeNinguna,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			arbol := enlazarLaCopia(t, copia)
			aplicarElMutante(t, arbol, caso.mutante)

			conElCambio, err := leerElCalculoDeLaHuella(t.Context(), arbol, grafo, fijadosDelTest)
			require.NoError(t, err)

			huellasConElCambio, err := conElCambio.huellas()
			require.NoError(t, err)

			eleccion := calculoDeLaEleccion{
				conjuntos: conElCambio.conjuntos, comun: conElCambio.comun, cabeza: conElCambio.cabeza,
				base: calculo.cabeza,
			}.elegir(caso.mutante.cambios)
			require.Empty(t, eleccion.sinConjunto, "cada conjunto de la copia con el cambio se calcula")

			elegidas := nombresDeLasElegidas(eleccion)
			comprobarElAlcance(t, caso.alcance, elegidas, evals)
			assert.Equal(t, elegidas, evalsConOtraHuella(evals, huellas, huellasConElCambio),
				"cambian las huellas de las evals que elige el cambio, y solo esas")
		})
	}

	conJuez := evalsConJuez(t, repositorio, evals)
	require.NotEmpty(t, conJuez, "alguna skill tiene juez")
	require.Less(t, len(conJuez), len(evals), "alguna skill no tiene juez")

	fijados := []struct {
		nombre    string
		cambiar   func(fijados *fijadosDeLaHuella)
		esperadas []string
	}{
		{
			nombre:    "modelo-que-decide",
			cambiar:   func(fijados *fijadosDeLaHuella) { fijados.modeloQueDecide += "-otro" },
			esperadas: evals,
		},
		{
			nombre:    "claude-code-de-las-sesiones",
			cambiar:   func(fijados *fijadosDeLaHuella) { fijados.claudeCode += ".1" },
			esperadas: evals,
		},
		{
			nombre:    "modelo-del-juez",
			cambiar:   func(fijados *fijadosDeLaHuella) { fijados.modeloDelJuez += "-otro" },
			esperadas: conJuez,
		},
		{
			nombre:    "claude-code-del-juez",
			cambiar:   func(fijados *fijadosDeLaHuella) { fijados.claudeCodeDelJuez += ".1" },
			esperadas: conJuez,
		},
	}

	for _, caso := range fijados {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			conOtros := calculo
			caso.cambiar(&conOtros.fijados)

			huellasConOtros, err := conOtros.huellas()
			require.NoError(t, err)

			assert.Equal(t, caso.esperadas, evalsConOtraHuella(evals, huellas, huellasConOtros))
		})
	}

	t.Run("texto-canonico", func(t *testing.T) {
		t.Parallel()

		textos, err := calculo.textos()
		require.NoError(t, err)

		for indice, conjunto := range calculo.conjuntos {
			comprobarElTexto(t, calculo, conjunto, textos[indice], huellas[indice], slices.Contains(conJuez, evals[indice]))
		}
	})

	t.Run("sin-huella", func(t *testing.T) {
		t.Parallel()

		comprobarQueNoHayHuella(t, calculo)
	})

	t.Run("dos-clones", func(t *testing.T) {
		t.Parallel()

		for range 2 {
			clon := t.TempDir()
			ejecutarGit(t, clon, "clone", "-q", copia, ".")

			grafoDelClon, err := leerGrafoDePaquetes(t.Context(), clon)
			require.NoError(t, err)

			delClon, err := leerElCalculoDeLaHuella(t.Context(), clon, grafoDelClon, fijadosDelTest)
			require.NoError(t, err)

			huellasDelClon, err := delClon.huellas()
			require.NoError(t, err)
			assert.Equal(t, huellas, huellasDelClon, "el clon de %s da las mismas huellas", clon)
		}
	})
}

// copiaDelRepositorio copia en un temporal los ficheros del árbol de trabajo
// del repositorio —los de git ls-files --cached --others --exclude-standard
// que están en el disco— que miden las evals sin descontar los tests ni los
// fixtures de ningún paquete, todo internal/, docs/ y los .gitignore, y los
// commitea en un repositorio nuevo, del que se puede clonar. Lo demás (specs/,
// evidencias/, web/…) ni lo miden las evals ni lo lista go list, y solo haría
// más lenta la copia.
func copiaDelRepositorio(t *testing.T) string {
	t.Helper()

	repositorio := os.DirFS(raizDelRepositorio)
	copia := t.TempDir()

	raiz, err := os.OpenRoot(copia)
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, raiz.Close()) })

	// El listado es el de leerArbolDeLaCabeza, con el entorno del proceso: lo
	// que ignora git en el repositorio tampoco está en la copia.
	listado, err := ejecutarEn(
		exec.CommandContext(t.Context(), "git", "ls-files", "-z", "--cached", "--others", "--exclude-standard"),
		raizDelRepositorio)
	require.NoError(t, err)

	copiados := 0

	for _, ruta := range rutasDeGit(listado) {
		if !rutaQueMide(ruta, nil) && !strings.HasPrefix(ruta, "internal/") && !strings.HasPrefix(ruta, "docs/") &&
			path.Base(ruta) != ".gitignore" {
			continue
		}

		estado, err := fs.Lstat(repositorio, ruta)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		require.NoError(t, err)
		require.Truef(t, estado.Mode().IsRegular(), "%s es un fichero", ruta)

		contenido, err := fs.ReadFile(repositorio, ruta)
		require.NoError(t, err)
		require.NoError(t, raiz.MkdirAll(path.Dir(ruta), 0o750))
		require.NoError(t, raiz.WriteFile(ruta, contenido, 0o600))

		copiados++
	}

	require.Positive(t, copiados)

	ejecutarGit(t, copia, "init", "-q", "-b", "main")
	ejecutarGit(t, copia, "add", "-A")
	ejecutarGit(t, copia, "commit", "-q", "-m", "la copia del repositorio")

	return copia
}

// enlazarLaCopia es un árbol nuevo en un temporal con los ficheros de la copia,
// los de su commit, cada uno un enlace duro al suyo, y un repositorio de git
// vacío, en el que todos son ficheros sin seguimiento: el de un subtest que
// cambia ficheros sin tocar la copia, que comparten los demás.
func enlazarLaCopia(t *testing.T, copia string) string {
	t.Helper()

	arbol := t.TempDir()

	raiz, err := os.OpenRoot(arbol)
	require.NoError(t, err)

	defer func() { require.NoError(t, raiz.Close()) }()

	for _, ruta := range rutasDeGit([]byte(ejecutarGit(t, copia, "ls-files", "-z"))) {
		require.NoError(t, raiz.MkdirAll(path.Dir(ruta), 0o750))
		require.NoError(t, os.Link(filepath.Join(copia, filepath.FromSlash(ruta)),
			filepath.Join(arbol, filepath.FromSlash(ruta))))
	}

	ejecutarGit(t, arbol, "init", "-q", "-b", "main")

	return arbol
}

// aplicarElMutante cambia el árbol con el mutante: retira sus ficheros
// retirados y escribe los cambiados, cada uno en un fichero nuevo, que
// sustituye al enlace a la copia.
func aplicarElMutante(t *testing.T, arbol string, mutante mutanteDeLaPropuesta) {
	t.Helper()

	raiz, err := os.OpenRoot(arbol)
	require.NoError(t, err)

	defer func() { require.NoError(t, raiz.Close()) }()

	for _, ruta := range slices.Concat(mutante.retirados, slices.Collect(maps.Keys(mutante.cambiados))) {
		if err := raiz.Remove(ruta); !errors.Is(err, fs.ErrNotExist) {
			require.NoError(t, err)
		}
	}

	for ruta, fichero := range mutante.cambiados {
		require.NoError(t, raiz.MkdirAll(path.Dir(ruta), 0o750))
		require.NoError(t, raiz.WriteFile(ruta, fichero.Data, 0o600))
	}
}

// entradaDeAlgunas es la ruta de la primera entrada, en orden de ruta, de la
// parte y la clase dadas que está en el conjunto de alguna eval y no en el de
// todas. Sale en rojo si no hay ninguna.
func entradaDeAlgunas(t *testing.T, conjuntos []dependenciasDeLaEval, parte parteDelConjunto,
	clase claseDeEntrada,
) string {
	t.Helper()

	cuantos := map[string]int{}

	for _, conjunto := range conjuntos {
		for _, entrada := range conjunto.entradas {
			if entrada.parte == parte && entrada.clase == clase {
				cuantos[entrada.ruta]++
			}
		}
	}

	for _, ruta := range slices.Sorted(maps.Keys(cuantos)) {
		if cuantos[ruta] < len(conjuntos) {
			return ruta
		}
	}

	require.FailNowf(t, "ninguna entrada de algunas evals", "parte %s, clase %s", parte, clase)

	return ""
}

// ficheroBajoLaEntrada es el primer fichero bajo la entrada directorio, con la
// barra final. Sale en rojo si no hay ninguno.
func ficheroBajoLaEntrada(t *testing.T, repositorio fs.FS, directorio string) string {
	t.Helper()

	ruta := primerFicheroBajo(t, repositorio, strings.TrimSuffix(directorio, "/"), func(string) bool { return true })
	require.NotEmptyf(t, ruta, "%s tiene algún fichero", directorio)

	return ruta
}

// conOtraVersionDeSqliteDeLoComun es el mutante que cambia en go.mod la
// versión de modernc.org/sqlite, que es un módulo de lo común.
func conOtraVersionDeSqliteDeLoComun(t *testing.T, repositorio fs.FS, calculo calculoDeLaHuella) mutanteDeLaPropuesta {
	t.Helper()

	require.Contains(t, calculo.comun.modulos, "modernc.org/sqlite", "sqlite es un módulo de lo común")

	return conOtraVersionDeSqlite(t, repositorio, calculo.cabeza)
}

// moduloDeNingunConjunto es un módulo que no está en go.mod, ni en ningún
// conjunto ni en lo común: el del replace y la línea de go.sum de los
// mutantes.
const moduloDeNingunConjunto = "ejemplo.invalid/ninguno"

// conUnReplaceEnGoMod es el mutante que añade al final de go.mod un replace de
// un módulo de ningún conjunto.
func conUnReplaceEnGoMod(t *testing.T, repositorio fs.FS, calculo calculoDeLaHuella) mutanteDeLaPropuesta {
	t.Helper()

	return conUnaLineaAlFinal(t, repositorio, calculo, "go.mod", "replace "+moduloDeNingunConjunto+" => ./ninguno")
}

// conUnaLineaEnGoSum es el mutante que añade al final de go.sum una línea de
// un módulo de ningún conjunto.
func conUnaLineaEnGoSum(t *testing.T, repositorio fs.FS, calculo calculoDeLaHuella) mutanteDeLaPropuesta {
	t.Helper()

	return conUnaLineaAlFinal(t, repositorio, calculo, "go.sum",
		moduloDeNingunConjunto+" v1.0.0/go.mod h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
}

// conUnaLineaAlFinal es el mutante que añade la línea al final del fichero,
// que es go.mod o go.sum, con un módulo que no es de ningún conjunto ni de lo
// común ni está en go.mod.
func conUnaLineaAlFinal(t *testing.T, repositorio fs.FS, calculo calculoDeLaHuella, ruta, linea string,
) mutanteDeLaPropuesta {
	t.Helper()

	require.NotContains(t, calculo.cabeza.modulos, moduloDeNingunConjunto)
	require.NotContains(t, calculo.comun.modulos, moduloDeNingunConjunto)

	for _, conjunto := range calculo.conjuntos {
		require.NotContains(t, conjunto.modulos, moduloDeNingunConjunto)
	}

	contenido, err := fs.ReadFile(repositorio, ruta)
	require.NoError(t, err)

	return mutanteDeLaPropuesta{
		cambiados: fstest.MapFS{ruta: {Data: append(slices.Clone(contenido), "\n"+linea+"\n"...)}},
		cambios:   []cambioDeLaPropuesta{{ruta: ruta}},
	}
}

// comprobarElAlcance comprueba que las evals elegidas son las que dice el
// alcance: ninguna, todas o algunas, que no son todas.
func comprobarElAlcance(t *testing.T, alcance alcanceDeUnCambio, elegidas, evals []string) {
	t.Helper()

	switch alcance {
	case eligeNinguna:
		require.Empty(t, elegidas, "el cambio no elige ninguna eval")
	case eligeTodas:
		require.Equal(t, evals, elegidas, "el cambio elige todas las evals")
	case eligeAlgunas:
		require.NotEmpty(t, elegidas, "el cambio elige alguna eval")
		require.Less(t, len(elegidas), len(evals), "el cambio no elige todas las evals")
	}
}

// nombresDeLosConjuntos son los nombres de las evals de los conjuntos, en su
// orden.
func nombresDeLosConjuntos(conjuntos []dependenciasDeLaEval) []string {
	nombres := make([]string, 0, len(conjuntos))
	for _, conjunto := range conjuntos {
		nombres = append(nombres, conjunto.nombre())
	}

	return nombres
}

// evalsConOtraHuella son las evals, en su orden, cuya huella difiere entre las
// dos listas.
func evalsConOtraHuella(evals, antes, despues []string) []string {
	var cambiadas []string

	for indice, eval := range evals {
		if antes[indice] != despues[indice] {
			cambiadas = append(cambiadas, eval)
		}
	}

	return cambiadas
}

// evalsConJuez son las evals del listado cuya skill tiene evals/<skill>/juez/,
// en su orden.
func evalsConJuez(t *testing.T, repositorio fs.FS, evals []string) []string {
	t.Helper()

	var conJuez []string

	for _, eval := range evals {
		skill, _, _ := strings.Cut(eval, "/")

		estado, err := fs.Stat(repositorio, "evals/"+skill+"/juez")
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		require.NoError(t, err)

		if estado.IsDir() {
			conJuez = append(conJuez, eval)
		}
	}

	return conJuez
}

// Los bloques del texto de la huella, en su orden (contracts/eleccion-y-huella.md
// §4): el primer campo de cada línea.
var bloquesDelTexto = []string{
	"huella-de-eval", "modelo_que_decide", "claude_code", "juez", "go", "toolchain", "modulo", "go_mod_resto",
	"go_sum_resto", "descripcion", "ruta",
}

// comprobarElTexto comprueba el texto canónico de la eval y su huella, con la
// forma de contracts/eleccion-y-huella.md §4: la huella es su sha256; sus
// líneas, terminadas en salto de línea, van por bloques en su orden y
// ordenadas por bytes dentro de cada uno; la cabecera, con los fijados del
// test, el juez solo en la skill que lo tiene, go y toolchain de go.mod; un
// módulo por cada uno del conjunto y de lo común, con la versión de go.mod; los
// dos restos; una descripción por skill, con el sha256 de la suya; y cada ruta
// con el sha256 de su contenido: la de la eval, la del job, que es de lo común,
// y ninguna de otra eval, de docs/, de go.mod ni de go.sum.
func comprobarElTexto(t *testing.T, calculo calculoDeLaHuella, conjunto dependenciasDeLaEval, texto, huella string,
	conJuez bool,
) {
	t.Helper()

	assert.Equal(t, huellaSHA256([]byte(texto)), huella, "la huella de %s es el sha256 de su texto", conjunto.nombre())
	require.True(t, strings.HasSuffix(texto, "\n"), "cada línea termina en un salto de línea")

	lineas := strings.Split(strings.TrimSuffix(texto, "\n"), "\n")
	porBloque := map[string][]string{}
	ultimo := 0

	for _, linea := range lineas {
		campo, valor, _ := strings.Cut(linea, " ")

		bloque := slices.Index(bloquesDelTexto, campo)
		require.GreaterOrEqualf(t, bloque, ultimo, "la línea %q va en su bloque, en orden", linea)

		ultimo = bloque
		porBloque[campo] = append(porBloque[campo], valor)
	}

	for campo, valores := range porBloque {
		assert.Truef(t, slices.IsSorted(valores), "las líneas %s de %s, ordenadas", campo, conjunto.nombre())
	}

	cabecera := map[string][]string{
		"huella-de-eval":    {"v1"},
		"modelo_que_decide": {fijadosDelTest.modeloQueDecide},
		"claude_code":       {fijadosDelTest.claudeCode},
		"go":                {calculo.cabeza.versionDeGo},
		"toolchain":         {calculo.cabeza.toolchain},
	}
	if conJuez {
		cabecera["juez"] = []string{fijadosDelTest.modeloDelJuez + " " + fijadosDelTest.claudeCodeDelJuez}
	}

	for _, campo := range bloquesDelTexto[:6] {
		assert.Equalf(t, cabecera[campo], porBloque[campo], "la línea %s de %s", campo, conjunto.nombre())
	}

	comprobarLosModulos(t, calculo, conjunto, porBloque["modulo"])

	for _, campo := range []string{"go_mod_resto", "go_sum_resto"} {
		require.Lenf(t, porBloque[campo], 1, "una línea %s", campo)
		assert.Regexp(t, patronDeUnaHuella, porBloque[campo][0])
	}

	var descripciones []string
	for _, skill := range calculo.cabeza.skills {
		descripciones = append(descripciones, skill+" "+huellaSHA256([]byte(calculo.cabeza.descripciones[skill])))
	}

	assert.Equal(t, descripciones, porBloque["descripcion"], "una descripción por skill")

	comprobarLasRutas(t, calculo, conjunto, porBloque["ruta"])
}

// comprobarLosModulos comprueba que hay una línea por módulo del conjunto y de
// lo común, con su versión en go.mod.
func comprobarLosModulos(t *testing.T, calculo calculoDeLaHuella, conjunto dependenciasDeLaEval, valores []string) {
	t.Helper()

	modulos := slices.Sorted(maps.Keys(calculo.comun.modulos))
	modulos = append(modulos, slices.Collect(maps.Keys(conjunto.modulos))...)
	slices.Sort(modulos)

	var esperados []string

	for _, modulo := range slices.Compact(modulos) {
		version, requerido := calculo.cabeza.modulos[modulo]
		require.Truef(t, requerido, "go.mod requiere %s", modulo)

		esperados = append(esperados, modulo+" "+version)
	}

	assert.Equal(t, esperados, valores, "los módulos de %s", conjunto.nombre())
}

// comprobarLasRutas comprueba las líneas ruta del texto de la eval: cada una,
// de un fichero que mide, con el sha256 de su contenido; la de la eval y la del
// job, que es de lo común; y ninguna de otra eval, de docs/, de go.mod ni de
// go.sum.
func comprobarLasRutas(t *testing.T, calculo calculoDeLaHuella, conjunto dependenciasDeLaEval, valores []string) {
	t.Helper()

	var rutas []string

	for _, valor := range valores {
		ruta, suma, _ := strings.Cut(valor, " ")
		assert.Equalf(t, calculo.cabeza.ficheros[ruta], suma, "el sha256 de %s", ruta)

		rutas = append(rutas, ruta)
	}

	propia := "evals/" + conjunto.nombre() + ".yaml"
	assert.Contains(t, rutas, propia)
	assert.Contains(t, rutas, ".github/workflows/evals.yml")

	for _, ruta := range rutas {
		assert.Falsef(t, esLaDeUnaEval(ruta) && ruta != propia, "%s no lleva la ruta de otra eval: %s",
			conjunto.nombre(), ruta)
		assert.NotContains(t, []string{"go.mod", "go.sum"}, ruta)
		assert.Falsef(t, strings.HasPrefix(ruta, "docs/"), "%s no lleva %s", conjunto.nombre(), ruta)
	}
}

// esLaDeUnaEval dice si la ruta es la del fichero de una eval,
// evals/<skill>/<NN>-*.yaml.
func esLaDeUnaEval(ruta string) bool {
	resto, bajoEvals := strings.CutPrefix(ruta, "evals/")
	_, fichero, enUnaSkill := strings.Cut(resto, "/")
	esEval, err := path.Match(formaDeUnaEval, fichero)

	return bajoEvals && enUnaSkill && err == nil && esEval
}

// comprobarQueNoHayHuella comprueba que no hay huella sin el conjunto de cada
// eval —el error nombra la eval que no lo tiene— ni con un fijado vacío o con
// espacios, que haría ambiguo el texto.
func comprobarQueNoHayHuella(t *testing.T, calculo calculoDeLaHuella) {
	t.Helper()

	sinConjunto := calculo
	sinConjunto.conjuntos = slices.Clone(calculo.conjuntos)
	sinConjunto.conjuntos[0].err = errors.New(sinConjunto.conjuntos[0].nombre() + ": un conjunto que no se calcula")

	_, err := sinConjunto.huellas()
	require.Error(t, err)
	assert.Contains(t, err.Error(), sinConjunto.conjuntos[0].nombre())

	for _, valor := range []string{"", "claude sonnet", "claude\nsonnet"} {
		conUnFijado := calculo
		conUnFijado.fijados.modeloQueDecide = valor

		_, err := conUnFijado.huellas()
		assert.Errorf(t, err, "el modelo que decide %q no da huella", valor)
	}
}
