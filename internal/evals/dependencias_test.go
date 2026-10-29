package evals

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// TestConjuntoDeCadaEval calcula sobre el árbol del repositorio el conjunto de
// dependencias de cada eval (FR-001 a FR-005, SC-002; data-model §1 y §3;
// contracts/eleccion-y-huella.md §2) y comprueba, cada cosa en su subtest: que
// cada eval de evals/*/ tiene un conjunto calculable, no vacío y con sus seis
// partes; lo que FR-001 pide y excluye para una eval de plazos-y-territorio y
// una de jurisprudencia; que un applet inexistente en la tabla de una skill es
// un error de cada eval que la usa, nombrada; que una referencia a mano entra
// por la parte (b), sin (c) ni error, y que una skill nombrada solo desde una
// de ellas entra con sus partes (FR-014 de H31, SC-004); y lo que da el
// predicado de las rutas que miden con una ruta de cada clase de data-model §1.
func TestConjuntoDeCadaEval(t *testing.T) {
	t.Parallel()

	grafo, err := leerGrafoDePaquetes(t.Context(), raizDelRepositorio)
	require.NoError(t, err)

	repositorio := os.DirFS(raizDelRepositorio)

	conjuntos, err := calcularDependencias(repositorio, grafo)
	require.NoError(t, err)

	casos := []struct {
		nombre    string
		comprobar func(t *testing.T, calculo calculoDelRepositorio)
	}{
		{nombre: "cada-eval-con-sus-seis-partes", comprobar: comprobarLasSeisPartes},
		{nombre: "lo-que-pide-fr-001-a-plazos-y-territorio", comprobar: comprobarFR001DePlazosYTerritorio},
		{nombre: "lo-que-pide-fr-001-a-jurisprudencia", comprobar: comprobarFR001DeJurisprudencia},
		{nombre: "applet-inexistente", comprobar: comprobarElAppletInexistente},
		{nombre: "referencias-a-mano", comprobar: comprobarLasReferenciasAMano},
		{nombre: "rutas-que-miden", comprobar: comprobarLasRutasQueMiden},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			caso.comprobar(t, calculoDelRepositorio{repositorio: repositorio, grafo: grafo, conjuntos: conjuntos})
		})
	}
}

// calculoDelRepositorio es lo que comparten los subtests de
// TestConjuntoDeCadaEval: el árbol del repositorio, su grafo de paquetes y los
// conjuntos calculados sobre los dos, que ningún subtest cambia.
type calculoDelRepositorio struct {
	repositorio fs.FS
	grafo       grafoDePaquetes
	conjuntos   []dependenciasDeLaEval
}

// comprobarLasSeisPartes comprueba que hay un conjunto por eval del repositorio,
// en su orden, y que cada uno se calcula sin error y trae sus seis partes, cada
// una con lo que dice el repositorio, leído aquí por su lado: (a) su fichero;
// (b) el SKILL.md y las referencias de su skill; (c) los data/*.yaml de las
// cabeceras de las referencias de las skills de (b); (d) algún paquete, de un
// directorio que existe; (e) las raíces de grabaciones y la de su grafo previo;
// (f) la carpeta del juez de su skill, si la tiene, y los demás ficheros de su
// directorio que no son evals.
func comprobarLasSeisPartes(t *testing.T, calculo calculoDelRepositorio) {
	t.Helper()

	evals := nombresDeLasEvals(t, calculo.repositorio)
	require.NotEmpty(t, evals)

	calculados := make([]string, 0, len(calculo.conjuntos))
	for _, conjunto := range calculo.conjuntos {
		calculados = append(calculados, conjunto.nombre())
	}

	require.Equal(t, evals, calculados, "un conjunto por eval del repositorio, en el orden de skill y de fichero")

	for _, conjunto := range calculo.conjuntos {
		require.NoErrorf(t, conjunto.err, "%s: el conjunto no se puede calcular", conjunto.nombre())
		require.NotEmptyf(t, conjunto.entradas, "%s: el conjunto está vacío", conjunto.nombre())

		partes := entradasPorParte(conjunto)
		ficheroDeLaSkill := "skills/" + conjunto.skill + "/SKILL.md"

		assert.Equalf(t, []string{"fichero evals/" + conjunto.nombre() + ".yaml"}, partes[parteDeLaEval],
			"%s: (a)", conjunto.nombre())
		assert.Subsetf(t, partes[parteDeLasSkills],
			[]string{"fichero " + ficheroDeLaSkill, "directorio skills/" + conjunto.skill + "/references/"},
			"%s: (b)", conjunto.nombre())

		datos := datosDeLasCabeceras(t, calculo.repositorio, partes[parteDeLasSkills])
		assert.NotEmptyf(t, datos, "%s: (c) sin ningún data/*.yaml", conjunto.nombre())
		assert.Equalf(t, datos, partes[parteDeLosDatos], "%s: (c)", conjunto.nombre())

		comprobarLosPaquetes(t, calculo.repositorio, conjunto.nombre(), partes[parteDeLosPaquetes])

		assert.ElementsMatchf(t, grabacionesDeLaEval(t, calculo.repositorio, conjunto), partes[parteDeLasGrabaciones],
			"%s: (e)", conjunto.nombre())
		assert.ElementsMatchf(t, juezDeLaSkill(t, calculo.repositorio, conjunto.skill), partes[parteDelJuez],
			"%s: (f)", conjunto.nombre())

		t.Logf("%s: %d entradas y %d módulos", conjunto.nombre(), len(conjunto.entradas), len(conjunto.modulos))
	}
}

// comprobarLosPaquetes comprueba la parte (d) de un conjunto: algún paquete, y
// el directorio de cada uno existe en el repositorio.
func comprobarLosPaquetes(t *testing.T, repositorio fs.FS, eval string, paquetes []string) {
	t.Helper()

	assert.True(t, slices.ContainsFunc(paquetes, func(entrada string) bool {
		return strings.HasPrefix(entrada, string(entradaPaquete)+" ")
	}), "%s: (d) sin ningún paquete", eval)

	for _, entrada := range paquetes {
		if directorio, esPaquete := strings.CutPrefix(entrada, string(entradaPaquete)+" "); esPaquete {
			estado, err := fs.Stat(repositorio, directorio)
			require.NoErrorf(t, err, "%s: (d) el paquete %s no está en el repositorio", eval, directorio)
			assert.Truef(t, estado.IsDir(), "%s: (d) %s no es un directorio", eval, directorio)
		}
	}
}

// comprobarFR001DePlazosYTerritorio comprueba el criterio de FR-001 en la
// primera eval de plazos-y-territorio: pide skills/normativa/SKILL.md,
// data/plazos.yaml y un fichero de internal/core/plazos, y excluye otra eval
// suya, una eval de otra skill y skills/jurisprudencia/SKILL.md
// (contracts/eleccion-y-huella.md §2).
func comprobarFR001DePlazosYTerritorio(t *testing.T, calculo calculoDelRepositorio) {
	t.Helper()

	comprobarLoQuePideFR001(t, calculo, "plazos-y-territorio",
		[]string{
			"skills/normativa/SKILL.md",
			"data/plazos.yaml",
			ficheroDelRepositorio(t, calculo.repositorio, "internal/core/plazos", esCodigoSinTests),
		},
		[]string{"skills/jurisprudencia/SKILL.md"})
}

// comprobarFR001DeJurisprudencia comprueba el criterio de FR-001 en la primera
// eval de jurisprudencia: pide skills/normativa/SKILL.md, un fichero de
// internal/core/cita y uno de la carpeta de su juez, y excluye otra eval suya,
// una eval de otra skill y skills/plazos-y-territorio/SKILL.md.
func comprobarFR001DeJurisprudencia(t *testing.T, calculo calculoDelRepositorio) {
	t.Helper()

	comprobarLoQuePideFR001(t, calculo, "jurisprudencia",
		[]string{
			"skills/normativa/SKILL.md",
			ficheroDelRepositorio(t, calculo.repositorio, "internal/core/cita", esCodigoSinTests),
			ficheroDelRepositorio(t, calculo.repositorio, "evals/jurisprudencia/juez", func(string) bool { return true }),
		},
		[]string{"skills/plazos-y-territorio/SKILL.md"})
}

// comprobarLoQuePideFR001 comprueba que el conjunto de la primera eval de la
// skill contiene cada ruta pedida y ninguna excluida, además de la segunda eval
// de la skill y la primera de otra skill, y que cada una de esas rutas existe en
// el repositorio, para que una ruta que ya no está no pase sin comprobar nada.
func comprobarLoQuePideFR001(t *testing.T, calculo calculoDelRepositorio, skill string, pide, excluye []string) {
	t.Helper()

	evals := nombresDeLasEvals(t, calculo.repositorio)
	deLaSkill := slices.DeleteFunc(slices.Clone(evals), func(eval string) bool {
		return !strings.HasPrefix(eval, skill+"/")
	})
	require.GreaterOrEqual(t, len(deLaSkill), 2, "la skill %s tiene al menos dos evals", skill)

	indiceDeOtra := slices.IndexFunc(evals, func(eval string) bool { return !strings.HasPrefix(eval, skill+"/") })
	require.GreaterOrEqual(t, indiceDeOtra, 0, "hay evals de otra skill")

	excluye = append(slices.Clone(excluye), "evals/"+deLaSkill[1]+".yaml", "evals/"+evals[indiceDeOtra]+".yaml")

	indice := slices.IndexFunc(calculo.conjuntos, func(conjunto dependenciasDeLaEval) bool {
		return conjunto.nombre() == deLaSkill[0]
	})
	require.GreaterOrEqual(t, indice, 0, "%s tiene conjunto", deLaSkill[0])

	conjunto := calculo.conjuntos[indice]
	require.NoError(t, conjunto.err)

	for _, ruta := range slices.Concat(pide, excluye) {
		_, err := fs.Stat(calculo.repositorio, ruta)
		require.NoErrorf(t, err, "%s está en el repositorio", ruta)
	}

	for _, ruta := range pide {
		assert.Truef(t, conjunto.pertenece(ruta), "%s: el conjunto no contiene %s", conjunto.nombre(), ruta)
	}

	for _, ruta := range excluye {
		assert.Falsef(t, conjunto.pertenece(ruta), "%s: el conjunto contiene %s", conjunto.nombre(), ruta)
	}
}

// appletInexistente es el applet que el mutante añade a la tabla de comandos de
// una skill: ningún fichero de internal/app lo declara.
const appletInexistente = "inexistente"

// comprobarElAppletInexistente añade en memoria, a la tabla de comandos del
// SKILL.md de jurisprudencia, la sección de un applet inexistente, y comprueba
// que el conjunto de cada eval que usa esa skill —la de sus evals y la de las
// que la nombran, según el cálculo sin el mutante— es un error que empieza por
// el nombre de la eval y nombra el applet, y que el de las demás se calcula. Si
// el cálculo acepta el applet, sale en rojo (FR-005).
func comprobarElAppletInexistente(t *testing.T, calculo calculoDelRepositorio) {
	t.Helper()

	const ficheroMutado = "skills/jurisprudencia/SKILL.md"

	original, err := fs.ReadFile(calculo.repositorio, ficheroMutado)
	require.NoError(t, err)

	applets, err := skills.AppletsDeLaTabla(original)
	require.NoError(t, err)

	titulo := "### `kitlegal " + applets[0] + "`\n"
	require.Equal(t, 1, strings.Count(string(original), titulo), "la tabla titula una vez el applet %s", applets[0])

	mutado := strings.Replace(string(original), titulo, titulo+"\n### `kitlegal "+appletInexistente+"`\n", 1)

	conMutante, err := skills.AppletsDeLaTabla([]byte(mutado))
	require.NoError(t, err)
	require.Contains(t, conMutante, appletInexistente, "el mutante añade el applet a la tabla")

	arbol := arbolConCambios{FS: calculo.repositorio, cambiados: fstest.MapFS{ficheroMutado: {Data: []byte(mutado)}}}

	conjuntos, err := calcularDependencias(arbol, calculo.grafo)
	require.NoError(t, err)
	require.Len(t, conjuntos, len(calculo.conjuntos))

	usanLaSkill := 0

	for indice, conjunto := range conjuntos {
		require.Equal(t, calculo.conjuntos[indice].nombre(), conjunto.nombre())

		if !calculo.conjuntos[indice].pertenece(ficheroMutado) {
			require.NoErrorf(t, conjunto.err, "%s no usa la skill del mutante", conjunto.nombre())

			continue
		}

		usanLaSkill++

		require.Errorf(t, conjunto.err, "%s: el cálculo acepta el applet %q", conjunto.nombre(), appletInexistente)
		assert.Truef(t, strings.HasPrefix(conjunto.err.Error(), conjunto.nombre()+": "),
			"el error nombra la eval: %v", conjunto.err)
		assert.Contains(t, conjunto.err.Error(), `"`+appletInexistente+`"`)
	}

	assert.Positive(t, usanLaSkill, "alguna eval usa la skill del mutante")
}

// arbolConCambios es un árbol con unos ficheros cambiados o añadidos en
// memoria: abre los de cambiados y todo lo demás del árbol de debajo, y la
// lista de un directorio es la de debajo con los añadidos que están en él.
type arbolConCambios struct {
	fs.FS

	cambiados fstest.MapFS
}

// Open abre el fichero cambiado, si lo es, y si no el del árbol de debajo.
func (a arbolConCambios) Open(nombre string) (fs.File, error) {
	if _, cambiado := a.cambiados[nombre]; cambiado {
		return a.cambiados.Open(nombre)
	}

	return a.FS.Open(nombre)
}

// ReadDir lista el directorio de debajo, si está, con las entradas de
// cambiados que no tiene, en orden de nombre: un fichero añadido, y el
// directorio que solo está en cambiados, con los suyos.
func (a arbolConCambios) ReadDir(nombre string) ([]fs.DirEntry, error) {
	deDebajo, errDeDebajo := fs.ReadDir(a.FS, nombre)

	anadidas, errDeLosCambiados := a.cambiados.ReadDir(nombre)
	if errDeLosCambiados != nil {
		return deDebajo, errDeDebajo
	}

	if errDeDebajo != nil && !errors.Is(errDeDebajo, fs.ErrNotExist) {
		return nil, errDeDebajo
	}

	entradas := map[string]fs.DirEntry{}
	for _, entrada := range slices.Concat(anadidas, deDebajo) {
		entradas[entrada.Name()] = entrada
	}

	return slices.SortedFunc(maps.Values(entradas), func(una, otra fs.DirEntry) int {
		return strings.Compare(una.Name(), otra.Name())
	}), nil
}

// referenciaAManoDelMutante es el nombre, en references/ de una skill, de la
// referencia a mano que añaden los mutantes de FR-014 de H31: ninguna skill lo
// declara en kitlegal-referencias.
const referenciaAManoDelMutante = "protocolo-del-mutante.md"

// comprobarLasReferenciasAMano añade en memoria una referencia a mano a
// normativa, que nombra jurisprudencia, y otra a jurisprudencia, que
// nombra plazos-y-territorio, y comprueba sobre el cálculo con las dos (FR-014
// y Clarificación Q1 de H31; SC-004): que cada conjunto se calcula, sin el
// error de la referencia sin cabecera; que cada referencia a mano entra solo
// por la parte (b), con su directorio de references/, y no da nada en (c); y
// que en el conjunto de cada eval de normativa entra jurisprudencia, nombrada
// solo desde una referencia a mano, con su SKILL.md, sus references/ y los
// paquetes de su tabla, y plazos-y-territorio, a la que se llega siguiendo el
// cierre desde jurisprudencia. Sin las dos referencias, ni una ni otra son de
// esos conjuntos: el test lo comprueba antes, para no pasar sin comprobar nada.
func comprobarLasReferenciasAMano(t *testing.T, calculo calculoDelRepositorio) {
	t.Helper()

	deNormativa := "skills/normativa/references/" + referenciaAManoDelMutante
	deJurisprudencia := "skills/jurisprudencia/references/" + referenciaAManoDelMutante

	nombradas := []string{
		"fichero skills/jurisprudencia/SKILL.md", "directorio skills/jurisprudencia/references/",
		"fichero skills/plazos-y-territorio/SKILL.md", "directorio skills/plazos-y-territorio/references/",
	}
	paquetes := []string{
		ficheroDelRepositorio(t, calculo.repositorio, "internal/core/cita", esCodigoSinTests),
		ficheroDelRepositorio(t, calculo.repositorio, "internal/core/plazos", esCodigoSinTests),
	}

	sinLasReferencias := 0

	for _, conjunto := range calculo.conjuntos {
		if conjunto.skill != "normativa" {
			continue
		}

		sinLasReferencias++

		for _, ruta := range []string{"skills/jurisprudencia/SKILL.md", "skills/plazos-y-territorio/SKILL.md"} {
			require.Falsef(t, conjunto.pertenece(ruta),
				"%s: sin las referencias a mano, el conjunto ya contiene %s", conjunto.nombre(), ruta)
		}
	}

	require.Positive(t, sinLasReferencias, "normativa tiene evals")

	for _, ruta := range []string{deNormativa, deJurisprudencia} {
		_, err := fs.Stat(calculo.repositorio, ruta)
		require.ErrorIsf(t, err, fs.ErrNotExist, "la referencia a mano %s es nueva", ruta)
	}

	arbol := arbolConCambios{FS: calculo.repositorio, cambiados: fstest.MapFS{
		deNormativa:      {Data: []byte("# Protocolo de detalle\n\nUna sentencia la lleva la skill jurisprudencia.\n")},
		deJurisprudencia: {Data: []byte("# Protocolo de detalle\n\nUn plazo lo calcula la skill plazos-y-territorio.\n")},
	}}

	conjuntos, err := calcularDependencias(arbol, calculo.grafo)
	require.NoError(t, err)
	require.Len(t, conjuntos, len(calculo.conjuntos))

	for _, conjunto := range conjuntos {
		require.NoErrorf(t, conjunto.err, "%s: el conjunto no se puede calcular", conjunto.nombre())

		partes := entradasPorParte(conjunto)
		assert.Equalf(t, datosDeLasCabeceras(t, arbol, partes[parteDeLasSkills]), partes[parteDeLosDatos],
			"%s: (c)", conjunto.nombre())

		if conjunto.skill != "normativa" {
			continue
		}

		for _, ruta := range []string{deNormativa, deJurisprudencia} {
			assert.Equalf(t, []parteDelConjunto{parteDeLasSkills}, partesQueContienen(conjunto, ruta),
				"%s: la referencia a mano %s entra solo por (b)", conjunto.nombre(), ruta)
		}

		assert.Subsetf(t, partes[parteDeLasSkills], nombradas, "%s: (b)", conjunto.nombre())

		for _, ruta := range paquetes {
			assert.Truef(t, conjunto.pertenece(ruta), "%s: el conjunto no contiene %s", conjunto.nombre(), ruta)
		}
	}
}

// partesQueContienen son las partes de las entradas del conjunto que contienen
// la ruta, en el orden de las entradas.
func partesQueContienen(conjunto dependenciasDeLaEval, ruta string) []parteDelConjunto {
	var partes []parteDelConjunto

	for _, entrada := range conjunto.entradas {
		if entrada.contiene(ruta) {
			partes = append(partes, entrada.parte)
		}
	}

	return partes
}

// comprobarLasRutasQueMiden comprueba el predicado de data-model §1 con una
// ruta del repositorio de cada clase, con los paquetes de los conjuntos: un
// fichero de un paquete de un applet mide; uno de internal/workflow, un
// _test.go y un fixture de un paquete de un applet y uno de docs/ no miden; y
// uno de una raíz de grabaciones dentro de los fixtures de un paquete de un
// applet mide.
func comprobarLasRutasQueMiden(t *testing.T, calculo calculoDelRepositorio) {
	t.Helper()

	paquetes := paquetesDeLosConjuntos(calculo.conjuntos)
	require.NotEmpty(t, paquetes)

	conTests := slices.IndexFunc(paquetes, func(paquete string) bool {
		entradas, err := fs.ReadDir(calculo.repositorio, paquete)
		require.NoError(t, err)

		return slices.ContainsFunc(entradas, func(entrada fs.DirEntry) bool {
			return strings.HasSuffix(entrada.Name(), "_test.go")
		})
	})
	require.GreaterOrEqual(t, conTests, 0, "algún paquete de un applet tiene tests")

	casos := []struct {
		clase string
		ruta  string
		mide  bool
	}{
		{
			clase: "fichero de un paquete de un applet",
			ruta:  ficheroDelRepositorio(t, calculo.repositorio, paquetes[conTests], esCodigoSinTests),
			mide:  true,
		},
		{
			clase: "internal/workflow",
			ruta:  ficheroDelRepositorio(t, calculo.repositorio, "internal/workflow", esCodigoSinTests),
		},
		{
			clase: "_test.go de un paquete de un applet",
			ruta: ficheroDelRepositorio(t, calculo.repositorio, paquetes[conTests], func(nombre string) bool {
				return strings.HasSuffix(nombre, "_test.go")
			}),
		},
		{clase: "fixture de un paquete de un applet", ruta: fixtureDeUnPaquete(t, calculo.repositorio, paquetes)},
		{clase: "raíz de grabaciones", ruta: grabacionDeUnPaquete(t, calculo.repositorio, paquetes), mide: true},
		{
			clase: "docs/",
			ruta: ficheroDelRepositorio(t, calculo.repositorio, "docs", func(nombre string) bool {
				return strings.HasSuffix(nombre, ".md")
			}),
		},
	}

	for _, caso := range casos {
		assert.Equalf(t, caso.mide, rutaQueMide(caso.ruta, paquetes), "%s: %s", caso.clase, caso.ruta)
	}
}

// fixtureDeUnPaquete es el primer fichero de los fixtures de un paquete de los
// dados que no está en una raíz de grabaciones.
func fixtureDeUnPaquete(t *testing.T, repositorio fs.FS, paquetes []string) string {
	t.Helper()

	for _, paquete := range paquetes {
		if ruta := primerFicheroBajo(t, repositorio, paquete+"/testdata", func(ruta string) bool {
			return !empiezaPorAlguno(ruta, raicesDeGrabaciones())
		}); ruta != "" {
			return ruta
		}
	}

	require.FailNow(t, "ningún paquete de un applet tiene fixtures fuera de las grabaciones")

	return ""
}

// grabacionDeUnPaquete es el primer fichero de la primera raíz de grabaciones
// que está entre los fixtures de un paquete de los dados.
func grabacionDeUnPaquete(t *testing.T, repositorio fs.FS, paquetes []string) string {
	t.Helper()

	for _, raiz := range raicesDeGrabaciones() {
		enUnPaquete := slices.ContainsFunc(paquetes, func(paquete string) bool {
			return strings.HasPrefix(raiz, paquete+"/testdata/")
		})
		if !enUnPaquete {
			continue
		}

		if ruta := primerFicheroBajo(t, repositorio, strings.TrimSuffix(raiz, "/"), func(string) bool {
			return true
		}); ruta != "" {
			return ruta
		}
	}

	require.FailNow(t, "ninguna raíz de grabaciones está entre los fixtures de un paquete de un applet")

	return ""
}

// primerFicheroBajo es la ruta del primer fichero bajo el directorio, en el
// orden del recorrido, que cumple la condición; vacía si no hay ninguno o el
// directorio no existe.
func primerFicheroBajo(t *testing.T, repositorio fs.FS, directorio string, cumple func(ruta string) bool) string {
	t.Helper()

	if _, err := fs.Stat(repositorio, directorio); err != nil {
		require.ErrorIs(t, err, fs.ErrNotExist)

		return ""
	}

	var encontrada string

	require.NoError(t, fs.WalkDir(repositorio, directorio, func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil || encontrada != "" || entrada.IsDir() {
			return err
		}

		if cumple(ruta) {
			encontrada = ruta
		}

		return nil
	}))

	return encontrada
}

// TestLoComun comprueba lo común del repositorio (FR-002; data-model §2): sus
// rutas, con una de cada clase; sus paquetes, los del módulo que están en
// ellas; y sus módulos: el de cada importación externa de un paquete de
// lo común y de los tests del arnés, con la versión de go.mod.
func TestLoComun(t *testing.T) {
	t.Parallel()

	grafo, err := leerGrafoDePaquetes(t.Context(), raizDelRepositorio)
	require.NoError(t, err)

	repositorio := os.DirFS(raizDelRepositorio)

	goMod, err := fs.ReadFile(repositorio, "go.mod")
	require.NoError(t, err)

	versiones, err := modulosDeGoMod(t.Context(), raizDelRepositorio, goMod)
	require.NoError(t, err)

	requeridos := versiones.modulos

	comun, err := grafo.loComun(requeridos)
	require.NoError(t, err)

	t.Run("rutas", func(t *testing.T) {
		t.Parallel()

		for ruta, esComun := range map[string]bool{
			ficheroDelRepositorio(t, repositorio, "internal/cli", esCodigoSinTests):   true,
			ficheroDelRepositorio(t, repositorio, "internal/evals", esCodigoSinTests): true,
			"skills.go":                     true,
			"Makefile":                      true,
			".github/workflows/evals.yml":   true,
			"scripts/evals-sesion.sh":       true,
			"schemas/juez-clases.yaml.json": true,
			ficheroDelRepositorio(t, repositorio, "internal/core/plazos", esCodigoSinTests): false,
			"skills/plazos-y-territorio/SKILL.md":                                           false,
			"go.mod":                                                                        false,
		} {
			assert.Equalf(t, esComun, esDeLoComun(ruta), "%s", ruta)
		}
	})

	t.Run("paquetes", func(t *testing.T) {
		t.Parallel()

		assert.Subset(t, comun.paquetes, []string{"cmd/kitlegal", "internal/app", "internal/cli", "internal/evals"})
		assert.NotContains(t, comun.paquetes, "internal/core/plazos")
		assert.NotContains(t, comun.paquetes, "internal/workflow")
	})

	t.Run("modulos", func(t *testing.T) {
		t.Parallel()

		comprobarLosModulosDeLoComun(t, grafo, comun, requeridos)
	})
}

// comprobarLosModulosDeLoComun comprueba que cada importación externa de un
// paquete de lo común y de los tests de internal/evals es de un módulo de lo
// común, con la versión que fija go.mod.
func comprobarLosModulosDeLoComun(t *testing.T, grafo grafoDePaquetes, comun loComun, requeridos map[string]string) {
	t.Helper()

	arnes, listado := grafo.paquetes[grafo.modulo+"/internal/evals"]
	require.True(t, listado)

	importaciones := slices.Concat(arnes.TestImports, arnes.XTestImports)

	for _, paquete := range comun.paquetes {
		importaciones = append(importaciones, grafo.paquetes[grafo.modulo+"/"+paquete].Imports...)
	}

	externas := 0

	for _, importacion := range importaciones {
		primero, _, _ := strings.Cut(importacion, "/")
		if !strings.Contains(primero, ".") || strings.HasPrefix(importacion, grafo.modulo) {
			continue
		}

		externas++

		modulo := ""

		for candidato := range comun.modulos {
			if (importacion == candidato || strings.HasPrefix(importacion, candidato+"/")) && len(candidato) > len(modulo) {
				modulo = candidato
			}
		}

		if assert.NotEmptyf(t, modulo, "el módulo de %s es de lo común", importacion) {
			assert.Equalf(t, requeridos[modulo], comun.modulos[modulo], "la versión de %s es la de go.mod", modulo)
		}
	}

	assert.Positive(t, externas, "lo común y los tests del arnés importan algún paquete externo")
}

// TestArbolDeLaCabezaYDeLaBase comprueba el árbol de la cabeza y el de la base
// de fusión (data-model §4; research D5) sobre un repositorio de git en un
// temporal: main con un commit y una propuesta que cambia la descripción de una
// skill, añade otra y sube un módulo, con cambios sin commitear encima. La
// cabeza da las rutas que miden, sin las borradas del árbol de trabajo ni las
// ignoradas, con el sha256 de su contenido; la base, lo de su commit, sin
// ficheros.
func TestArbolDeLaCabezaYDeLaBase(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()

	raiz, err := os.OpenRoot(directorio)
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, raiz.Close()) })

	escribirEnLaRaiz(t, raiz, map[string]string{
		".gitignore":                         "internal/ignorado.go\n",
		"go.mod":                             goModDelArbol("v1.0.0"),
		"skills/uno/SKILL.md":                skillDelRepositorioDePrueba("uno", "La de la base."),
		"evals/uno/01-una.yaml":              "pregunta: una\n",
		"evals/uno/02-otra.yaml":             "pregunta: otra\n",
		"docs/nota.md":                       "# Nota\n",
		"internal/workflow/flujo.go":         "package workflow\n",
		"internal/paquete/paquete.go":        "package paquete\n",
		"internal/paquete/paquete_test.go":   "package paquete\n",
		"internal/paquete/testdata/uno.json": "{}\n",
	})
	ejecutarGit(t, directorio, "init", "-q", "-b", "main")
	ejecutarGit(t, directorio, "add", "-A")
	ejecutarGit(t, directorio, "commit", "-q", "-m", "la base")

	base := ejecutarGit(t, directorio, "rev-parse", "HEAD")

	ejecutarGit(t, directorio, "checkout", "-q", "-b", "propuesta")
	escribirEnLaRaiz(t, raiz, map[string]string{
		"go.mod":              goModDelArbol("v1.1.0"),
		"skills/uno/SKILL.md": skillDelRepositorioDePrueba("uno", "La de la cabeza."),
		"skills/dos/SKILL.md": skillDelRepositorioDePrueba("dos", "La nueva."),
	})
	ejecutarGit(t, directorio, "add", "-A")
	ejecutarGit(t, directorio, "commit", "-q", "-m", "la propuesta")
	escribirEnLaRaiz(t, raiz, map[string]string{
		"internal/nuevo/nuevo.go": "package nuevo\n",
		"internal/ignorado.go":    "package internal\n",
	})
	require.NoError(t, raiz.Remove("evals/uno/01-una.yaml"))

	t.Run("cabeza", func(t *testing.T) {
		t.Parallel()

		cabeza, err := leerArbolDeLaCabeza(t.Context(), directorio, []string{"internal/paquete"})
		require.NoError(t, err)

		medidos := []string{
			"evals/uno/02-otra.yaml", "go.mod", "internal/nuevo/nuevo.go", "internal/paquete/paquete.go",
			"skills/dos/SKILL.md", "skills/uno/SKILL.md",
		}
		assert.Equal(t, medidos, slices.Sorted(maps.Keys(cabeza.ficheros)))

		for _, ruta := range medidos {
			contenido, err := raiz.ReadFile(ruta)
			require.NoError(t, err)

			suma := sha256.Sum256(contenido)
			assert.Equalf(t, hex.EncodeToString(suma[:]), cabeza.ficheros[ruta], "el sha256 de %s", ruta)
		}

		assert.Equal(t, map[string]string{"uno": "La de la cabeza.", "dos": "La nueva."}, cabeza.descripciones)
		assert.Equal(t, []string{"dos", "uno"}, cabeza.skills)
		assert.Equal(t, []string{"uno"}, cabeza.directoriosDeEvals)
		assert.Equal(t, map[string]string{"ejemplo.invalid/dependencia": "v1.1.0"}, cabeza.modulos)
		assert.Equal(t, "1.21", cabeza.versionDeGo)
		assert.Equal(t, "go1.21.0", cabeza.toolchain)
	})

	t.Run("base", func(t *testing.T) {
		t.Parallel()

		deFusion, err := baseDeFusion(t.Context(), directorio, "main")
		require.NoError(t, err)
		assert.Equal(t, base, deFusion)

		arbol, err := leerArbolDeLaBase(t.Context(), directorio, deFusion)
		require.NoError(t, err)

		assert.Empty(t, arbol.ficheros)
		assert.Equal(t, map[string]string{"uno": "La de la base."}, arbol.descripciones)
		assert.Equal(t, []string{"uno"}, arbol.skills)
		assert.Equal(t, []string{"uno"}, arbol.directoriosDeEvals)
		assert.Equal(t, map[string]string{"ejemplo.invalid/dependencia": "v1.0.0"}, arbol.modulos)
		assert.Equal(t, "1.21", arbol.versionDeGo)
		assert.Equal(t, "go1.21.0", arbol.toolchain)
	})
}

// goModDelArbol es el go.mod del repositorio de TestArbolDeLaCabezaYDeLaBase,
// con la versión dada de su única dependencia. La versión de Go es anterior a
// la del repositorio, para que go mod edit no cambie de toolchain.
func goModDelArbol(version string) string {
	return "module ejemplo.invalid/arbol\n\ngo 1.21\n\ntoolchain go1.21.0\n\nrequire ejemplo.invalid/dependencia " +
		version + "\n"
}

// skillDelRepositorioDePrueba es el SKILL.md de una skill del repositorio de
// TestArbolDeLaCabezaYDeLaBase, con su nombre y su descripción.
func skillDelRepositorioDePrueba(nombre, descripcion string) string {
	return "---\nname: " + nombre + "\ndescription: " + descripcion + "\n---\n\n# " + nombre + "\n"
}

// escribirEnLaRaiz escribe cada fichero, con las carpetas que le faltan.
func escribirEnLaRaiz(t *testing.T, raiz *os.Root, ficheros map[string]string) {
	t.Helper()

	for ruta, contenido := range ficheros {
		require.NoError(t, raiz.MkdirAll(path.Dir(ruta), 0o750))
		require.NoError(t, raiz.WriteFile(ruta, []byte(contenido), 0o600))
	}
}

// nombresDeLasEvals son las evals de evals/*/, como <skill>/<fichero sin
// .yaml>, en orden de skill y de fichero, leídas del listado del repositorio.
func nombresDeLasEvals(t *testing.T, repositorio fs.FS) []string {
	t.Helper()

	ficheros, err := fs.Glob(repositorio, "evals/*/[0-9][0-9]-*.yaml")
	require.NoError(t, err)

	evals := make([]string, 0, len(ficheros))
	for _, fichero := range ficheros {
		evals = append(evals, strings.TrimSuffix(strings.TrimPrefix(fichero, "evals/"), ".yaml"))
	}

	return evals
}

// entradasPorParte son las entradas del conjunto por su parte, cada una como
// «<clase> <ruta>», ordenadas.
func entradasPorParte(conjunto dependenciasDeLaEval) map[parteDelConjunto][]string {
	partes := map[parteDelConjunto][]string{}
	for _, entrada := range conjunto.entradas {
		partes[entrada.parte] = append(partes[entrada.parte], string(entrada.clase)+" "+entrada.ruta)
	}

	for parte := range partes {
		slices.Sort(partes[parte])
	}

	return partes
}

// cabeceraDeLaReferencia es la primera línea de una referencia generada, con el
// data/*.yaml del que sale: la sonda del test, independiente de la lectura de
// internal/skills.
var cabeceraDeLaReferencia = regexp.MustCompile(`\A<!-- generado desde (data/[^,\n]+), no editar -->\n`)

// datosDeLasCabeceras son los data/*.yaml de las cabeceras de las referencias
// de las skills cuyo SKILL.md está entre las entradas dadas, como entradas
// fichero, ordenadas y sin repetir. La referencia sin cabecera es a mano y no
// da ninguno (FR-014 de H31): si no lo fuera, el conjunto sería un error, y
// quien llama lo comprueba antes.
func datosDeLasCabeceras(t *testing.T, repositorio fs.FS, entradasDeLasSkills []string) []string {
	t.Helper()

	var datos []string

	for _, entrada := range entradasDeLasSkills {
		ruta, esFichero := strings.CutPrefix(entrada, string(entradaFichero)+" ")
		if !esFichero {
			continue
		}

		referencias, err := fs.Glob(repositorio, path.Dir(ruta)+"/references/*")
		require.NoError(t, err)

		for _, referencia := range referencias {
			contenido, err := fs.ReadFile(repositorio, referencia)
			require.NoError(t, err)

			if cabecera := cabeceraDeLaReferencia.FindSubmatch(contenido); cabecera != nil {
				datos = append(datos, string(entradaFichero)+" "+string(cabecera[1]))
			}
		}
	}

	slices.Sort(datos)

	return slices.Compact(datos)
}

// grabacionesDeLaEval son las entradas de la parte (e) del conjunto de la eval:
// las raíces de UnionDeGrabaciones y, si la eval lleva grafo previo, su
// directorio de GrafosPrevios.
func grabacionesDeLaEval(t *testing.T, repositorio fs.FS, conjunto dependenciasDeLaEval) []string {
	t.Helper()

	var grabaciones []string
	for _, raiz := range UnionDeGrabaciones() {
		grabaciones = append(grabaciones, string(entradaDirectorio)+" "+strings.TrimPrefix(raiz, "../../")+"/")
	}

	contenido, err := fs.ReadFile(repositorio, "evals/"+conjunto.nombre()+".yaml")
	require.NoError(t, err)

	eval, err := LeerEval(path.Base(conjunto.nombre())+".yaml", contenido)
	require.NoError(t, err)

	if eval.GrafoPrevio.Grabaciones != "" {
		grabaciones = append(grabaciones, string(entradaDirectorio)+" "+strings.TrimPrefix(GrafosPrevios, "../../")+
			"/"+eval.GrafoPrevio.Grabaciones+"/")
	}

	return grabaciones
}

// juezDeLaSkill son las entradas de la parte (f) de las evals de la skill: la
// carpeta de su juez, si la tiene, y cada fichero de su directorio de evals que
// no es una eval.
func juezDeLaSkill(t *testing.T, repositorio fs.FS, skill string) []string {
	t.Helper()

	entradas, err := fs.ReadDir(repositorio, "evals/"+skill)
	require.NoError(t, err)

	var juez []string

	for _, entrada := range entradas {
		esEval, err := path.Match("[0-9][0-9]-*.yaml", entrada.Name())
		require.NoError(t, err)

		switch {
		case entrada.IsDir() && entrada.Name() == "juez":
			juez = append(juez, string(entradaDirectorio)+" evals/"+skill+"/juez/")
		case !entrada.IsDir() && !esEval:
			juez = append(juez, string(entradaFichero)+" evals/"+skill+"/"+entrada.Name())
		}
	}

	return juez
}

// esCodigoSinTests dice si el nombre es el de un fichero de Go que no es de
// tests.
func esCodigoSinTests(nombre string) bool {
	return strings.HasSuffix(nombre, ".go") && !strings.HasSuffix(nombre, "_test.go")
}

// ficheroDelRepositorio es la ruta del primer fichero del directorio, sin
// entrar en sus subdirectorios, cuyo nombre cumple la condición. Sale en rojo
// si no hay ninguno.
func ficheroDelRepositorio(t *testing.T, repositorio fs.FS, directorio string, cumple func(nombre string) bool) string {
	t.Helper()

	entradas, err := fs.ReadDir(repositorio, directorio)
	require.NoError(t, err)

	indice := slices.IndexFunc(entradas, func(entrada fs.DirEntry) bool {
		return !entrada.IsDir() && cumple(entrada.Name())
	})
	require.GreaterOrEqualf(t, indice, 0, "%s tiene el fichero que busca el test", directorio)

	return directorio + "/" + entradas[indice].Name()
}
