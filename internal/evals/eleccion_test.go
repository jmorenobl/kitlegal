package evals

import (
	"errors"
	"io/fs"
	"maps"
	"os"
	"path"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEleccionDeLosMutantes calcula sobre el árbol del repositorio la elección
// de cada mutante de SC-003 y de FR-002 a FR-007 (quickstart.md §2; data-model
// §5), y la del de la referencia a mano de FR-014 de H31 (SC-004), que elige
// lo mismo que un cambio de una referencia generada de su skill: la base es el
// repositorio; la cabeza, el repositorio con el mutante en
// memoria, con sus conjuntos, calculados sobre ella. Cada subtest compara las
// evals elegidas, eval por eval, con las que espera, sacadas del listado de
// evals/<skill>/ y no de los conjuntos, y comprueba que cada una lleva el motivo
// del mutante y, como rutas, las suyas.
func TestEleccionDeLosMutantes(t *testing.T) {
	t.Parallel()

	grafo, err := leerGrafoDePaquetes(t.Context(), raizDelRepositorio)
	require.NoError(t, err)

	repositorio := os.DirFS(raizDelRepositorio)
	base := arbolConLosCambios(t, repositorio, mutanteDeLaPropuesta{})

	comun, err := grafo.loComun(base.modulos)
	require.NoError(t, err)

	evals := nombresDeLasEvals(t, repositorio)

	esperadas := map[string]struct {
		evals  []string
		motivo motivoDeLaEleccion
	}{
		mutanteReferenciaDeLegalCore: {evals: evalsDeLasSkills(evals, "legal-core"), motivo: motivoConjunto},
		mutanteCuerpoDeBoeLegislacion: {
			evals: evalsDeLasSkills(evals, "boe-legislacion", "legal-core", "jurisprudencia"), motivo: motivoConjunto,
		},
		mutanteCuerpoDeJurisprudencia: {evals: evalsDeLasSkills(evals, "jurisprudencia"), motivo: motivoConjunto},
		mutanteInternalCLI:            {evals: evals, motivo: motivoComun},
		mutanteSqliteEnGoMod:          {evals: evals, motivo: motivoComun},
		mutanteDescripcionDeLegalCore: {evals: evals, motivo: motivoDescripcion},
		mutantePaqueteSinAsignar:      {evals: evals, motivo: motivoSinAsignar},
		mutanteSoloDocs:               {},
		mutanteUnaEval:                {evals: []string{"jurisprudencia/03-por-materia"}, motivo: motivoConjunto},
		mutanteReferenciaRenombrada:   {evals: evalsDeLasSkills(evals, "legal-core"), motivo: motivoConjunto},
		mutanteReferenciaAManoDeLegalCore: {
			evals: evalsDeLasSkills(evals, "legal-core"), motivo: motivoConjunto,
		},
	}

	casos := append(mutantesDeLaEleccion(t, repositorio, base), mutanteConNombre{
		nombre: mutanteReferenciaAManoDeLegalCore, mutante: conUnaReferenciaAMano(t, repositorio, "legal-core"),
	})
	require.Len(t, casos, len(esperadas), "un esperado por mutante")

	for _, caso := range casos {
		esperada, conEsperada := esperadas[caso.nombre]
		require.Truef(t, conEsperada, "el mutante %s tiene su esperado", caso.nombre)

		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			mutante := caso.mutante
			cabeza := arbolConLosCambios(t, repositorio, mutante)

			conjuntos, err := calcularDependencias(arbolConCambios{FS: repositorio, cambiados: mutante.cambiados}, grafo)
			require.NoError(t, err)

			eleccion := calculoDeLaEleccion{conjuntos: conjuntos, comun: comun, cabeza: cabeza, base: base}.
				elegir(mutante.cambios)

			require.Empty(t, eleccion.sinConjunto, "cada conjunto de la cabeza se calcula")
			assert.Nil(t, eleccion.plataforma)
			assert.Equal(t, esperada.evals, nombresDeLasElegidas(eleccion), "las evals elegidas, eval por eval")

			var rutas []string

			for _, cambio := range mutante.cambios {
				rutas = append(rutas, cambio.ruta)
				if cambio.anterior != "" {
					rutas = append(rutas, cambio.anterior)
				}
			}

			slices.Sort(rutas)

			for _, elegida := range eleccion.elegidas {
				assert.Equalf(t, esperada.motivo, elegida.motivo, "el motivo de %s", elegida.nombre())
				assert.Equalf(t, rutas, elegida.rutas, "las rutas que eligen %s", elegida.nombre())
			}
		})
	}
}

// Los nombres de los mutantes de SC-003 y de FR-002 a FR-007, los de
// quickstart.md §2.
const (
	mutanteReferenciaDeLegalCore  = "referencia-de-legal-core"
	mutanteCuerpoDeBoeLegislacion = "cuerpo-de-boe-legislacion"
	mutanteCuerpoDeJurisprudencia = "cuerpo-de-jurisprudencia"
	mutanteInternalCLI            = "internal-cli"
	mutanteSqliteEnGoMod          = "sqlite-en-go-mod"
	mutanteDescripcionDeLegalCore = "descripcion-de-legal-core"
	mutantePaqueteSinAsignar      = "paquete-sin-asignar"
	mutanteSoloDocs               = "solo-docs"
	mutanteUnaEval                = "una-eval"
	mutanteReferenciaRenombrada   = "referencia-renombrada"
)

// mutanteReferenciaAManoDeLegalCore es el mutante de FR-014 de H31 (SC-004),
// que solo elige: la matriz de TestMatrizDeLasPropuestas es la de las
// ejecuciones fijas de H30.
const mutanteReferenciaAManoDeLegalCore = "referencia-a-mano-de-legal-core"

// mutanteConNombre es un mutante de SC-003 o de FR-002 a FR-007 con su nombre.
type mutanteConNombre struct {
	nombre  string
	mutante mutanteDeLaPropuesta
}

// mutantesDeLaEleccion son los mutantes de SC-003 y de FR-002 a FR-007 sobre el
// repositorio, en el orden de quickstart.md §2, los que comparten
// TestEleccionDeLosMutantes y TestMatrizDeLasPropuestas; base es el árbol del
// repositorio sin cambios.
func mutantesDeLaEleccion(t *testing.T, repositorio fs.FS, base arbolDelRepositorio) []mutanteConNombre {
	t.Helper()

	referencia := ficheroDelRepositorio(t, repositorio, "skills/legal-core/references", esMarkdown)
	cli := ficheroDelRepositorio(t, repositorio, "internal/cli", esCodigoSinTests)
	docs := ficheroDelRepositorio(t, repositorio, "docs", esMarkdown)

	return []mutanteConNombre{
		{nombre: mutanteReferenciaDeLegalCore, mutante: conUnaLineaMas(t, repositorio, referencia)},
		{nombre: mutanteCuerpoDeBoeLegislacion, mutante: conUnaLineaMas(t, repositorio, "skills/boe-legislacion/SKILL.md")},
		{nombre: mutanteCuerpoDeJurisprudencia, mutante: conUnaLineaMas(t, repositorio, "skills/jurisprudencia/SKILL.md")},
		{nombre: mutanteInternalCLI, mutante: conUnaLineaMas(t, repositorio, cli)},
		{nombre: mutanteSqliteEnGoMod, mutante: conOtraVersionDeSqlite(t, repositorio, base)},
		{nombre: mutanteDescripcionDeLegalCore, mutante: conOtraDescripcionDeLegalCore(t, repositorio)},
		{nombre: mutantePaqueteSinAsignar, mutante: conUnPaqueteNuevo(t, repositorio)},
		{nombre: mutanteSoloDocs, mutante: conUnaLineaMas(t, repositorio, docs)},
		{
			nombre:  mutanteUnaEval,
			mutante: conUnaLineaMas(t, repositorio, "evals/jurisprudencia/03-por-materia.yaml"),
		},
		{nombre: mutanteReferenciaRenombrada, mutante: conLaReferenciaRenombrada(t, repositorio, referencia)},
	}
}

// mutanteDeLaPropuesta es una propuesta sobre el repositorio, en memoria: los
// ficheros de la cabeza que cambia o añade, con su contenido, los que retira y
// las rutas cambiadas que daría la plataforma.
type mutanteDeLaPropuesta struct {
	cambiados fstest.MapFS
	retirados []string
	cambios   []cambioDeLaPropuesta
}

// arbolConLosCambios es el árbol, sin ficheros, del repositorio con el
// mutante: el que leerArbolDeLaBase lee de un commit, con las rutas de
// skills/ y evals/, sin las retiradas y con las cambiadas, y lo que se lee de
// ellas y de go.mod, con el contenido del mutante si lo cambia.
func arbolConLosCambios(t *testing.T, repositorio fs.FS, mutante mutanteDeLaPropuesta) arbolDelRepositorio {
	t.Helper()

	var rutas []string

	for _, directorio := range []string{"skills", "evals"} {
		require.NoError(t, fs.WalkDir(repositorio, directorio, func(ruta string, entrada fs.DirEntry, err error) error {
			if err == nil && !entrada.IsDir() {
				rutas = append(rutas, ruta)
			}

			return err
		}))
	}

	rutas = slices.DeleteFunc(rutas, func(ruta string) bool { return slices.Contains(mutante.retirados, ruta) })
	rutas = append(rutas, slices.Collect(maps.Keys(mutante.cambiados))...)
	slices.Sort(rutas)

	cabeza := arbolConCambios{FS: repositorio, cambiados: mutante.cambiados}

	arbol, err := completarArbol(t.Context(), raizDelRepositorio, slices.Compact(rutas), func(ruta string) ([]byte, error) {
		return fs.ReadFile(cabeza, ruta)
	})
	require.NoError(t, err)

	return arbol
}

// conUnaLineaMas es el mutante que cambia el fichero del repositorio con una
// línea más al final, un comentario en Go y en YAML, que no nombra ninguna
// skill: la cabecera de un SKILL.md queda igual y una eval se sigue leyendo.
func conUnaLineaMas(t *testing.T, repositorio fs.FS, ruta string) mutanteDeLaPropuesta {
	t.Helper()

	contenido, err := fs.ReadFile(repositorio, ruta)
	require.NoError(t, err)

	comentario := map[string]string{".go": "// ", ".yaml": "# "}[path.Ext(ruta)]

	return mutanteDeLaPropuesta{
		cambiados: fstest.MapFS{ruta: {Data: append(slices.Clone(contenido), "\n"+comentario+"Una línea del mutante.\n"...)}},
		cambios:   []cambioDeLaPropuesta{{ruta: ruta}},
	}
}

// conOtraVersionDeSqlite es el mutante que cambia en go.mod la versión de
// modernc.org/sqlite, la de la caché, por otra válida de la misma serie.
func conOtraVersionDeSqlite(t *testing.T, repositorio fs.FS, base arbolDelRepositorio) mutanteDeLaPropuesta {
	t.Helper()

	const sqlite = "modernc.org/sqlite"

	version, requerido := base.modulos[sqlite]
	require.Truef(t, requerido, "go.mod requiere %s", sqlite)

	contenido, err := fs.ReadFile(repositorio, "go.mod")
	require.NoError(t, err)

	mutado := strings.Replace(string(contenido), sqlite+" "+version+"\n", sqlite+" "+version+"-mutante\n", 1)
	require.NotEqual(t, string(contenido), mutado, "el mutante cambia la versión en go.mod")

	return mutanteDeLaPropuesta{
		cambiados: fstest.MapFS{"go.mod": {Data: []byte(mutado)}},
		cambios:   []cambioDeLaPropuesta{{ruta: "go.mod"}},
	}
}

// conOtraDescripcionDeLegalCore es el mutante que cambia solo el description
// de la cabecera de skills/legal-core/SKILL.md, con una palabra delante.
func conOtraDescripcionDeLegalCore(t *testing.T, repositorio fs.FS) mutanteDeLaPropuesta {
	t.Helper()

	const ruta, descripcion = "skills/legal-core/SKILL.md", "description: >-\n  "

	contenido, err := fs.ReadFile(repositorio, ruta)
	require.NoError(t, err)

	mutado := strings.Replace(string(contenido), descripcion, descripcion+"Mutante. ", 1)
	require.NotEqual(t, string(contenido), mutado, "el mutante cambia la descripción")

	return mutanteDeLaPropuesta{
		cambiados: fstest.MapFS{ruta: {Data: []byte(mutado)}},
		cambios:   []cambioDeLaPropuesta{{ruta: ruta}},
	}
}

// conUnPaqueteNuevo es el mutante que añade un paquete bajo internal/, que no
// está en el repositorio y del que no depende ningún applet ni lo común.
func conUnPaqueteNuevo(t *testing.T, repositorio fs.FS) mutanteDeLaPropuesta {
	t.Helper()

	const ruta = "internal/sinasignar/sinasignar.go"

	_, err := fs.Stat(repositorio, ruta)
	require.ErrorIs(t, err, fs.ErrNotExist, "el paquete del mutante es nuevo")

	return mutanteDeLaPropuesta{
		cambiados: fstest.MapFS{ruta: {Data: []byte("package sinasignar\n")}},
		cambios:   []cambioDeLaPropuesta{{ruta: ruta}},
	}
}

// conLaReferenciaRenombrada es el mutante que renombra la referencia, en su
// carpeta, sin cambiar su contenido: la plataforma da la ruta nueva con la
// anterior.
func conLaReferenciaRenombrada(t *testing.T, repositorio fs.FS, referencia string) mutanteDeLaPropuesta {
	t.Helper()

	contenido, err := fs.ReadFile(repositorio, referencia)
	require.NoError(t, err)

	nueva := strings.TrimSuffix(referencia, ".md") + "_renombrada.md"

	return mutanteDeLaPropuesta{
		cambiados: fstest.MapFS{nueva: {Data: contenido}},
		retirados: []string{referencia},
		cambios:   []cambioDeLaPropuesta{{ruta: nueva, anterior: referencia}},
	}
}

// conUnaReferenciaAMano es el mutante que añade a references/ de la skill una
// referencia a mano (FR-012 y FR-014 de H31): un .md que kitlegal-referencias
// no declara, sin la cabecera «generado desde» y que no nombra ninguna skill.
func conUnaReferenciaAMano(t *testing.T, repositorio fs.FS, skill string) mutanteDeLaPropuesta {
	t.Helper()

	ruta := "skills/" + skill + "/references/" + referenciaAManoDelMutante

	_, err := fs.Stat(repositorio, ruta)
	require.ErrorIs(t, err, fs.ErrNotExist, "la referencia a mano del mutante es nueva")

	return mutanteDeLaPropuesta{
		cambiados: fstest.MapFS{ruta: {Data: []byte("# Protocolo del mutante\n\nUna línea del mutante.\n")}},
		cambios:   []cambioDeLaPropuesta{{ruta: ruta}},
	}
}

// esMarkdown dice si el nombre es el de un fichero de Markdown.
func esMarkdown(nombre string) bool {
	return strings.HasSuffix(nombre, ".md")
}

// evalsDeLasSkills son las evals del listado que son de alguna de las skills,
// en el orden del listado.
func evalsDeLasSkills(evals []string, skillsElegidas ...string) []string {
	return slices.DeleteFunc(slices.Clone(evals), func(eval string) bool {
		skill, _, _ := strings.Cut(eval, "/")

		return !slices.Contains(skillsElegidas, skill)
	})
}

// nombresDeLasElegidas son los nombres de las evals elegidas, en su orden.
func nombresDeLasElegidas(eleccion eleccionDeLaPropuesta) []string {
	var nombres []string
	for _, elegida := range eleccion.elegidas {
		nombres = append(nombres, elegida.nombre())
	}

	return nombres
}

// TestMatrizDeLasPropuestas fija byte a byte la matriz de FR-019 (SC-004;
// FR-008; research D16 de H30) de un conjunto fijo de ejecuciones sobre el
// repositorio, con la plataforma en memoria y la medida de main sintética de
// TestDecisionDelJob, con la huella de cada eval: la apertura de cada mutante
// de TestEleccionDeLosMutantes, que solo corre los trabajos si elige alguna
// eval (FR-016); la etiqueta evals sobre una propuesta que no elige nada, que
// los corre igual (FR-017); y la suite completa (FR-018). La matriz de cada una,
// tras una línea con su nombre, dice lo que mide —con su motivo— y lo que
// reutiliza el trabajo de cada skill; todas juntas, en ese orden, se comparan
// con goldenDeLaMatriz, y el test sale en rojo con un byte de diferencia.
func TestMatrizDeLasPropuestas(t *testing.T) {
	t.Parallel()

	repositorio := leerElRepositorioDeLaDecision(t)

	var (
		casos          []casoDeLaDecision
		sinNadaElegido mutanteDeLaPropuesta
	)

	for _, mutante := range mutantesDeLaEleccion(t, repositorio.arbol, repositorio.base) {
		casos = append(casos, casoDeLaDecision{nombre: mutante.nombre, disparo: disparoDeLaApertura, mutante: mutante.mutante})

		if mutante.nombre == mutanteSoloDocs {
			sinNadaElegido = mutante.mutante
		}
	}

	require.NotEmpty(t, sinNadaElegido.cambios, "el mutante %s no elige nada", mutanteSoloDocs)

	casos = append(casos,
		casoDeLaDecision{nombre: "etiqueta-evals-sin-nada-elegido", disparo: disparoDeLaEtiqueta, mutante: sinNadaElegido},
		casoDeLaDecision{nombre: "suite-completa", disparo: disparoCompleto},
	)

	var matrices strings.Builder

	for _, caso := range casos {
		matrices.WriteString("## " + caso.nombre + "\n")
		matrices.WriteString(matrizDeLaEjecucion(repositorio.trabajosDelCaso(t, caso)))
	}

	assert.Equal(t, goldenDeLaMatriz, matrices.String(), "la matriz de cada ejecución, byte a byte")
}

// trabajosDelCaso son los trabajos de evals que corre la ejecución del caso,
// uno por skill y cada uno con su decisión, como en comprobarLaDecision, en
// orden inverso de skill, para que sea la matriz la que los ordena: ninguno si
// es la apertura de una propuesta que no elige ninguna eval.
func (r repositorioDeLaDecision) trabajosDelCaso(t *testing.T, caso casoDeLaDecision) []trabajoDecidido {
	t.Helper()

	calculo := r.calculo(t, caso)
	if !r.corren(t, caso, calculo) {
		return nil
	}

	trabajos := make([]trabajoDecidido, 0, len(r.skills))

	for _, skill := range slices.Backward(r.skills) {
		origen, _ := r.decidir(t, caso, calculo, skill)
		trabajos = append(trabajos, trabajoDecidido{skill: skill, origen: origen})
	}

	return trabajos
}

// TestReglasDeLaEleccion comprueba cada regla de data-model §5 sobre un árbol
// pequeño, con tres evals de dos skills —dos, cuyo SKILL.md nombra a uno, y
// uno—, dadas desordenadas: cada subtest compara, línea a línea, el registro de
// la elección (contracts/eleccion-y-huella.md §3), que dice qué evals elige, en
// qué orden, con qué motivo y por qué rutas.
func TestReglasDeLaEleccion(t *testing.T) {
	t.Parallel()

	todas := func(linea string) []string {
		return []string{"elegida dos/01-c: " + linea, "elegida uno/01-a: " + linea, "elegida uno/02-b: " + linea}
	}

	casos := []struct {
		nombre  string
		cabeza  func(cabeza *arbolDelRepositorio)
		cambios []cambioDeLaPropuesta
		lineas  []string
	}{
		{
			nombre:  "fichero-de-una-eval",
			cambios: []cambioDeLaPropuesta{{ruta: "evals/uno/02-b.yaml"}},
			lineas:  []string{"elegida uno/02-b: conjunto: evals/uno/02-b.yaml"},
		},
		{
			nombre:  "referencia-de-una-skill-nombrada",
			cambios: []cambioDeLaPropuesta{{ruta: "skills/uno/references/x.md"}},
			lineas:  todas("conjunto: skills/uno/references/x.md"),
		},
		{
			nombre:  "fichero-de-un-paquete",
			cambios: []cambioDeLaPropuesta{{ruta: "internal/core/dos/dos.go"}},
			lineas:  []string{"elegida dos/01-c: conjunto: internal/core/dos/dos.go"},
		},
		{
			nombre:  "renombrado-con-las-dos-rutas",
			cambios: []cambioDeLaPropuesta{{ruta: "skills/dos/references/nueva.md", anterior: "skills/dos/references/vieja.md"}},
			lineas: []string{
				"elegida dos/01-c: conjunto: skills/dos/references/nueva.md, skills/dos/references/vieja.md",
			},
		},
		{
			nombre:  "renombrado-desde-fuera-de-lo-que-mide",
			cambios: []cambioDeLaPropuesta{{ruta: "evals/uno/02-b.yaml", anterior: "docs/vieja.yaml"}},
			lineas:  []string{"elegida uno/02-b: conjunto: evals/uno/02-b.yaml"},
		},
		{
			nombre: "comun-con-una-ruta-de-un-conjunto",
			cambios: []cambioDeLaPropuesta{
				{ruta: "internal/cli/salida.go"}, {ruta: "evals/uno/01-a.yaml"}, {ruta: "internal/cli/salida.go"},
			},
			lineas: []string{
				"elegida dos/01-c: comun: internal/cli/salida.go",
				"elegida uno/01-a: comun: evals/uno/01-a.yaml, internal/cli/salida.go",
				"elegida uno/02-b: comun: internal/cli/salida.go",
			},
		},
		{
			nombre:  "fichero-comun-suelto",
			cambios: []cambioDeLaPropuesta{{ruta: ".github/workflows/evals.yml"}},
			lineas:  todas("comun: .github/workflows/evals.yml"),
		},
		{
			nombre: "rutas-que-no-miden",
			cambios: []cambioDeLaPropuesta{
				{ruta: "docs/ROADMAP.md"},
				{ruta: "internal/workflow/flujo.go"},
				{ruta: "README.md"},
				{ruta: "internal/core/uno/uno_test.go"},
				{ruta: "internal/core/uno/testdata/x.json"},
			},
		},
		{
			nombre:  "test-de-un-paquete-sin-conjunto",
			cambios: []cambioDeLaPropuesta{{ruta: "internal/render/render_test.go"}},
			lineas:  todas("sin_asignar: internal/render/render_test.go"),
		},
		{
			nombre:  "descripcion",
			cabeza:  func(cabeza *arbolDelRepositorio) { cabeza.descripciones["dos"] = "La de dos, cambiada." },
			cambios: []cambioDeLaPropuesta{{ruta: "skills/dos/SKILL.md"}},
			lineas:  todas("descripcion: skills/dos/SKILL.md"),
		},
		{
			nombre:  "cuerpo-de-un-skill-md",
			cambios: []cambioDeLaPropuesta{{ruta: "skills/dos/SKILL.md"}},
			lineas:  []string{"elegida dos/01-c: conjunto: skills/dos/SKILL.md"},
		},
		{
			nombre:  "skill-md-de-una-skill-nueva",
			cabeza:  conLaSkillNueva,
			cambios: []cambioDeLaPropuesta{{ruta: "skills/nueva/SKILL.md"}},
			lineas:  todas("descripcion: skills/nueva/SKILL.md"),
		},
		{
			nombre:  "directorio-de-skills-anadido",
			cabeza:  conLaSkillNueva,
			cambios: []cambioDeLaPropuesta{{ruta: "skills/nueva/references/x.md"}},
			lineas:  todas("sin_asignar: skills/nueva/references/x.md"),
		},
		{
			nombre: "directorio-de-evals-retirado",
			cabeza: func(cabeza *arbolDelRepositorio) {
				cabeza.directoriosDeEvals = slices.DeleteFunc(cabeza.directoriosDeEvals, func(directorio string) bool {
					return directorio == "vieja"
				})
			},
			cambios: []cambioDeLaPropuesta{{ruta: "evals/vieja/01-x.yaml"}},
			lineas:  todas("sin_asignar: evals/vieja/01-x.yaml"),
		},
		{
			nombre:  "eval-retirada",
			cambios: []cambioDeLaPropuesta{{ruta: "evals/uno/03-retirada.yaml"}},
			lineas:  todas("sin_asignar: evals/uno/03-retirada.yaml"),
		},
		{
			nombre:  "version-de-go",
			cabeza:  func(cabeza *arbolDelRepositorio) { cabeza.versionDeGo = "1.27.1" },
			cambios: []cambioDeLaPropuesta{{ruta: "go.mod"}},
			lineas:  todas("comun: go.mod"),
		},
		{
			nombre:  "toolchain",
			cabeza:  func(cabeza *arbolDelRepositorio) { cabeza.toolchain = "go1.27.4" },
			cambios: []cambioDeLaPropuesta{{ruta: "go.sum"}, {ruta: "go.mod"}},
			lineas:  todas("comun: go.mod, go.sum"),
		},
		{
			nombre:  "modulo-de-lo-comun",
			cabeza:  func(cabeza *arbolDelRepositorio) { cabeza.modulos["ejemplo.invalid/comun"] = "v1.1.0" },
			cambios: []cambioDeLaPropuesta{{ruta: "go.mod"}},
			lineas:  todas("comun: go.mod"),
		},
		{
			nombre:  "modulo-de-un-conjunto",
			cabeza:  func(cabeza *arbolDelRepositorio) { cabeza.modulos["ejemplo.invalid/de-dos"] = "v1.1.0" },
			cambios: []cambioDeLaPropuesta{{ruta: "go.mod"}},
			lineas:  []string{"elegida dos/01-c: conjunto: go.mod"},
		},
		{
			nombre:  "modulo-de-una-skill-nombrada",
			cabeza:  func(cabeza *arbolDelRepositorio) { cabeza.modulos["ejemplo.invalid/de-uno"] = "v1.1.0" },
			cambios: []cambioDeLaPropuesta{{ruta: "go.mod"}},
			lineas:  todas("conjunto: go.mod"),
		},
		{
			nombre:  "modulo-anadido-de-un-conjunto",
			cabeza:  func(cabeza *arbolDelRepositorio) { cabeza.modulos["ejemplo.invalid/solo-de-dos"] = "v1.0.0" },
			cambios: []cambioDeLaPropuesta{{ruta: "go.mod"}},
			lineas:  []string{"elegida dos/01-c: conjunto: go.mod"},
		},
		{
			nombre:  "modulo-retirado-de-un-conjunto",
			cabeza:  func(cabeza *arbolDelRepositorio) { delete(cabeza.modulos, "ejemplo.invalid/de-dos") },
			cambios: []cambioDeLaPropuesta{{ruta: "go.mod"}},
			lineas:  []string{"elegida dos/01-c: conjunto: go.mod"},
		},
		{
			nombre:  "modulo-sin-conjunto",
			cabeza:  func(cabeza *arbolDelRepositorio) { cabeza.modulos["ejemplo.invalid/otro"] = "v1.1.0" },
			cambios: []cambioDeLaPropuesta{{ruta: "go.mod"}},
			lineas:  todas("sin_asignar: go.mod"),
		},
		{
			nombre:  "go-sum-sin-cambiar-ningun-modulo",
			cambios: []cambioDeLaPropuesta{{ruta: "go.sum"}},
			lineas:  todas("sin_asignar: go.sum"),
		},
		{
			nombre: "sin_asignar-por-encima-de-descripcion-y-conjunto",
			cabeza: func(cabeza *arbolDelRepositorio) { cabeza.descripciones["uno"] = "La de uno, cambiada." },
			cambios: []cambioDeLaPropuesta{
				{ruta: "skills/uno/SKILL.md"}, {ruta: "evals/uno/01-a.yaml"}, {ruta: "internal/render/render.go"},
			},
			lineas: []string{
				"elegida dos/01-c: sin_asignar: internal/render/render.go, skills/uno/SKILL.md",
				"elegida uno/01-a: sin_asignar: evals/uno/01-a.yaml, internal/render/render.go, skills/uno/SKILL.md",
				"elegida uno/02-b: sin_asignar: internal/render/render.go, skills/uno/SKILL.md",
			},
		},
		{
			nombre:  "sin-cambios",
			cambios: nil,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			calculo := calculoDeReglas()
			if caso.cabeza != nil {
				caso.cabeza(&calculo.cabeza)
			}

			eleccion := calculo.elegir(caso.cambios)

			assert.Empty(t, eleccion.sinConjunto)
			assert.Nil(t, eleccion.plataforma)
			assert.Equal(t, caso.lineas, eleccion.lineas())
		})
	}

	t.Run("sin-conjunto", func(t *testing.T) {
		t.Parallel()

		calculo := conUnConjuntoIncalculable(calculoDeReglas())

		eleccion := calculo.elegir([]cambioDeLaPropuesta{{ruta: "evals/uno/02-b.yaml"}})

		assert.Equal(t, `dos/01-c: el applet "inexistente" no tiene fichero en internal/app`, eleccion.sinConjunto)
		assert.Equal(t, []string{
			"elegida dos/01-c: sin_conjunto",
			"elegida uno/01-a: sin_conjunto",
			"elegida uno/02-b: sin_conjunto: evals/uno/02-b.yaml",
		}, eleccion.lineas())
	})

	t.Run("comun-por-encima-de-sin-conjunto", func(t *testing.T) {
		t.Parallel()

		calculo := conUnConjuntoIncalculable(calculoDeReglas())

		eleccion := calculo.elegir([]cambioDeLaPropuesta{{ruta: "cmd/kitlegal/main.go"}})

		assert.NotEmpty(t, eleccion.sinConjunto)
		assert.Equal(t, todas("comun: cmd/kitlegal/main.go"), eleccion.lineas())
	})

	t.Run("plataforma", func(t *testing.T) {
		t.Parallel()

		plataforma := errorDeLaPlataforma{orden: "gh api --paginate repos/o/r/pulls/7/files", causa: "HTTP 502"}

		eleccion := calculoDeReglas().elegirSinLosFicheros(plataforma)

		assert.Empty(t, eleccion.sinConjunto)
		assert.Equal(t, &plataforma, eleccion.plataforma)
		assert.Equal(t, append([]string{"error de la plataforma: gh api --paginate repos/o/r/pulls/7/files: HTTP 502"},
			todas("plataforma")...), eleccion.lineas())
	})

	t.Run("sin-conjunto-por-encima-de-plataforma", func(t *testing.T) {
		t.Parallel()

		plataforma := errorDeLaPlataforma{orden: "gh api repos/o/r/pulls/7/files", causa: "plazo de 10 minutos agotado"}

		eleccion := conUnConjuntoIncalculable(calculoDeReglas()).elegirSinLosFicheros(plataforma)

		assert.NotEmpty(t, eleccion.sinConjunto)
		assert.Equal(t, append([]string{"error de la plataforma: gh api repos/o/r/pulls/7/files: plazo de 10 minutos agotado"},
			todas("sin_conjunto")...), eleccion.lineas())
	})

	t.Run("motivos-por-prioridad", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []motivoDeLaEleccion{
			motivoComun, motivoSinConjunto, motivoPlataforma, motivoSinAsignar, motivoDescripcion, motivoConjunto,
		}, prioridadDeLosMotivos)
		assert.Equal(t, []string{"comun", "sin_conjunto", "plataforma", "sin_asignar", "descripcion", "conjunto"},
			[]string{
				string(motivoComun), string(motivoSinConjunto), string(motivoPlataforma), string(motivoSinAsignar),
				string(motivoDescripcion), string(motivoConjunto),
			})
	})
}

// calculoDeReglas es el cálculo de TestReglasDeLaEleccion, nuevo en cada
// llamada para que cada subtest cambie su cabeza: tres evals —dos/01-c, que
// lleva también la skill uno, y uno/01-a y uno/02-b—, dadas desordenadas, con
// un paquete y un módulo por skill; lo común, con internal/cli y un módulo; y la
// base y la cabeza iguales, con una skill y un directorio de evals viejos sin
// evals.
func calculoDeReglas() calculoDeLaEleccion {
	deLaSkill := func(skill string) []entradaDeDependencia {
		return []entradaDeDependencia{
			{clase: entradaFichero, ruta: "skills/" + skill + "/SKILL.md", parte: parteDeLasSkills},
			{clase: entradaDirectorio, ruta: "skills/" + skill + "/references/", parte: parteDeLasSkills},
			{clase: entradaPaquete, ruta: "internal/core/" + skill, parte: parteDeLosPaquetes},
		}
	}
	deLaEval := func(skill, eval string, skillsDeLaEval ...string) dependenciasDeLaEval {
		conjunto := dependenciasDeLaEval{
			skill: skill, eval: eval, modulos: map[string]string{},
			entradas: []entradaDeDependencia{
				{clase: entradaFichero, ruta: "evals/" + skill + "/" + eval + ".yaml", parte: parteDeLaEval},
			},
		}
		for _, otra := range skillsDeLaEval {
			conjunto.entradas = append(conjunto.entradas, deLaSkill(otra)...)
			conjunto.modulos["ejemplo.invalid/de-"+otra] = "v1.0.0"
		}

		return conjunto
	}

	arbol := func() arbolDelRepositorio {
		return arbolDelRepositorio{
			descripciones:      map[string]string{"dos": "La de dos.", "uno": "La de uno.", "vieja": "La vieja."},
			skills:             []string{"dos", "uno", "vieja"},
			directoriosDeEvals: []string{"dos", "uno", "vieja"},
			versionesDeGoMod: versionesDeGoMod{
				modulos: map[string]string{
					"ejemplo.invalid/comun": "v1.0.0", "ejemplo.invalid/de-dos": "v1.0.0",
					"ejemplo.invalid/de-uno": "v1.0.0", "ejemplo.invalid/otro": "v1.0.0",
				},
				versionDeGo: "1.27.0",
				toolchain:   "go1.27.3",
			},
		}
	}

	dos := deLaEval("dos", "01-c", "dos", "uno")
	dos.modulos["ejemplo.invalid/solo-de-dos"] = "v1.0.0"

	return calculoDeLaEleccion{
		conjuntos: []dependenciasDeLaEval{deLaEval("uno", "02-b", "uno"), dos, deLaEval("uno", "01-a", "uno")},
		comun:     loComun{paquetes: []string{"internal/cli"}, modulos: map[string]string{"ejemplo.invalid/comun": "v1.0.0"}},
		cabeza:    arbol(),
		base:      arbol(),
	}
}

// conLaSkillNueva cambia la cabeza de TestReglasDeLaEleccion con una skill
// nueva, con su descripción.
func conLaSkillNueva(cabeza *arbolDelRepositorio) {
	cabeza.skills = append(cabeza.skills, "nueva")
	cabeza.descripciones["nueva"] = "La nueva."
}

// conUnConjuntoIncalculable es el cálculo con el conjunto de dos/01-c sin
// calcular, con su causa.
func conUnConjuntoIncalculable(calculo calculoDeLaEleccion) calculoDeLaEleccion {
	for indice, conjunto := range calculo.conjuntos {
		if conjunto.nombre() == "dos/01-c" {
			calculo.conjuntos[indice] = dependenciasDeLaEval{
				skill: conjunto.skill, eval: conjunto.eval,
				err: errors.New(`dos/01-c: el applet "inexistente" no tiene fichero en internal/app`),
			}
		}
	}

	return calculo
}

// goldenDeLaMatriz es la matriz de cada ejecución de TestMatrizDeLasPropuestas,
// en su orden, tras una línea con su nombre (FR-019; SC-004): las rutas de los
// mutantes son las que elige mutantesDeLaEleccion, y la procedencia de cada
// reutilizada, la de la medida de main sintética de TestDecisionDelJob.
const goldenDeLaMatriz = `## referencia-de-legal-core
evals (boe-legislacion): medidas 0, reutilizadas 23
  01-lpac-articulo-21.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  02-lcsp-contrato-menor.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  03-lrbrl-atribuciones-del-pleno.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  04-lgt-prescripcion.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  05-trlrhl-impuestos-municipales.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  06-irpf-rendimientos-del-trabajo.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  07-lrjsp-principio-de-legalidad.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  08-ltaibg-plazo-de-resolucion.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  09-constitucion-articulo-140.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  10-et-vacaciones.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  11-no-activa-programacion.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  12-no-activa-acuerdo-entre-amigos.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  13-lrbrl-atribuciones-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  14-trlrhl-impuestos-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  15-irpf-rendimientos-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  16-lrjsp-legalidad-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  17-ltaibg-plazo-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  18-lrjpac-norma-derogada.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  19-lcsp-contrato-menor-redaccion-cambiada.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  20-lcsp-dos-bloques-redaccion-cambiada.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  21-sin-binario-ni-servidor.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  22-lcsp-contrato-menor-en-2019.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  23-lpac-articulo-21-en-2016.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
evals (jurisprudencia): medidas 0, reutilizadas 10
  01-existe-con-numero-y-fecha.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  02-resumen-sin-documento.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  03-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  04-documento-pegado.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  05-tribunal-constitucional.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  06-documento-que-no-es-el-pedido.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  07-resumen-de-una-conocida.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  08-doctrina-con-el-fallo-delante.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  09-de-que-trata-con-la-ficha-sola.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  10-doctrina-dada-por-hecha.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
evals (legal-core): medidas 12, reutilizadas 0
  01-territorio-municipio-cubierto.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md)
  02-territorio-municipio-no-cubierto.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md)
  03-no-activa-receta-de-cocina.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md)
  04-sin-binario-ni-servidor.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md)
  05-plazo-reposicion-leganes.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md)
  06-plazo-reposicion-tordesillas.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md)
  07-cobertura-liquidacion-tributaria.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md)
  08-cobertura-multa-de-trafico.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md)
  09-plazo-contencioso-agosto.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md)
  10-plazo-reposicion-anterior-a-la-lpac.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md)
  11-plazo-reposicion-sin-leer-el-precepto.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md)
  12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md)
## cuerpo-de-boe-legislacion
evals (boe-legislacion): medidas 23, reutilizadas 0
  01-lpac-articulo-21.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  02-lcsp-contrato-menor.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  03-lrbrl-atribuciones-del-pleno.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  04-lgt-prescripcion.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  05-trlrhl-impuestos-municipales.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  06-irpf-rendimientos-del-trabajo.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  07-lrjsp-principio-de-legalidad.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  08-ltaibg-plazo-de-resolucion.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  09-constitucion-articulo-140.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  10-et-vacaciones.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  11-no-activa-programacion.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  12-no-activa-acuerdo-entre-amigos.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  13-lrbrl-atribuciones-por-materia.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  14-trlrhl-impuestos-por-materia.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  15-irpf-rendimientos-por-materia.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  16-lrjsp-legalidad-por-materia.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  17-ltaibg-plazo-por-materia.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  18-lrjpac-norma-derogada.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  19-lcsp-contrato-menor-redaccion-cambiada.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  20-lcsp-dos-bloques-redaccion-cambiada.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  21-sin-binario-ni-servidor.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  22-lcsp-contrato-menor-en-2019.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  23-lpac-articulo-21-en-2016.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
evals (jurisprudencia): medidas 10, reutilizadas 0
  01-existe-con-numero-y-fecha.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  02-resumen-sin-documento.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  03-por-materia.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  04-documento-pegado.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  05-tribunal-constitucional.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  06-documento-que-no-es-el-pedido.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  07-resumen-de-una-conocida.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  08-doctrina-con-el-fallo-delante.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  09-de-que-trata-con-la-ficha-sola.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  10-doctrina-dada-por-hecha.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
evals (legal-core): medidas 12, reutilizadas 0
  01-territorio-municipio-cubierto.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  02-territorio-municipio-no-cubierto.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  03-no-activa-receta-de-cocina.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  04-sin-binario-ni-servidor.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  05-plazo-reposicion-leganes.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  06-plazo-reposicion-tordesillas.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  07-cobertura-liquidacion-tributaria.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  08-cobertura-multa-de-trafico.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  09-plazo-contencioso-agosto.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  10-plazo-reposicion-anterior-a-la-lpac.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  11-plazo-reposicion-sin-leer-el-precepto.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
  12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml: medida: rutas de su conjunto (skills/boe-legislacion/SKILL.md)
## cuerpo-de-jurisprudencia
evals (boe-legislacion): medidas 0, reutilizadas 23
  01-lpac-articulo-21.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  02-lcsp-contrato-menor.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  03-lrbrl-atribuciones-del-pleno.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  04-lgt-prescripcion.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  05-trlrhl-impuestos-municipales.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  06-irpf-rendimientos-del-trabajo.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  07-lrjsp-principio-de-legalidad.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  08-ltaibg-plazo-de-resolucion.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  09-constitucion-articulo-140.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  10-et-vacaciones.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  11-no-activa-programacion.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  12-no-activa-acuerdo-entre-amigos.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  13-lrbrl-atribuciones-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  14-trlrhl-impuestos-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  15-irpf-rendimientos-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  16-lrjsp-legalidad-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  17-ltaibg-plazo-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  18-lrjpac-norma-derogada.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  19-lcsp-contrato-menor-redaccion-cambiada.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  20-lcsp-dos-bloques-redaccion-cambiada.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  21-sin-binario-ni-servidor.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  22-lcsp-contrato-menor-en-2019.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  23-lpac-articulo-21-en-2016.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
evals (jurisprudencia): medidas 10, reutilizadas 0
  01-existe-con-numero-y-fecha.yaml: medida: rutas de su conjunto (skills/jurisprudencia/SKILL.md)
  02-resumen-sin-documento.yaml: medida: rutas de su conjunto (skills/jurisprudencia/SKILL.md)
  03-por-materia.yaml: medida: rutas de su conjunto (skills/jurisprudencia/SKILL.md)
  04-documento-pegado.yaml: medida: rutas de su conjunto (skills/jurisprudencia/SKILL.md)
  05-tribunal-constitucional.yaml: medida: rutas de su conjunto (skills/jurisprudencia/SKILL.md)
  06-documento-que-no-es-el-pedido.yaml: medida: rutas de su conjunto (skills/jurisprudencia/SKILL.md)
  07-resumen-de-una-conocida.yaml: medida: rutas de su conjunto (skills/jurisprudencia/SKILL.md)
  08-doctrina-con-el-fallo-delante.yaml: medida: rutas de su conjunto (skills/jurisprudencia/SKILL.md)
  09-de-que-trata-con-la-ficha-sola.yaml: medida: rutas de su conjunto (skills/jurisprudencia/SKILL.md)
  10-doctrina-dada-por-hecha.yaml: medida: rutas de su conjunto (skills/jurisprudencia/SKILL.md)
evals (legal-core): medidas 0, reutilizadas 12
  01-territorio-municipio-cubierto.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  02-territorio-municipio-no-cubierto.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  03-no-activa-receta-de-cocina.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  04-sin-binario-ni-servidor.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  05-plazo-reposicion-leganes.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  06-plazo-reposicion-tordesillas.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  07-cobertura-liquidacion-tributaria.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  08-cobertura-multa-de-trafico.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  09-plazo-contencioso-agosto.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  10-plazo-reposicion-anterior-a-la-lpac.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  11-plazo-reposicion-sin-leer-el-precepto.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
## internal-cli
evals (boe-legislacion): medidas 23, reutilizadas 0
  01-lpac-articulo-21.yaml: medida: común (internal/cli/describe.go)
  02-lcsp-contrato-menor.yaml: medida: común (internal/cli/describe.go)
  03-lrbrl-atribuciones-del-pleno.yaml: medida: común (internal/cli/describe.go)
  04-lgt-prescripcion.yaml: medida: común (internal/cli/describe.go)
  05-trlrhl-impuestos-municipales.yaml: medida: común (internal/cli/describe.go)
  06-irpf-rendimientos-del-trabajo.yaml: medida: común (internal/cli/describe.go)
  07-lrjsp-principio-de-legalidad.yaml: medida: común (internal/cli/describe.go)
  08-ltaibg-plazo-de-resolucion.yaml: medida: común (internal/cli/describe.go)
  09-constitucion-articulo-140.yaml: medida: común (internal/cli/describe.go)
  10-et-vacaciones.yaml: medida: común (internal/cli/describe.go)
  11-no-activa-programacion.yaml: medida: común (internal/cli/describe.go)
  12-no-activa-acuerdo-entre-amigos.yaml: medida: común (internal/cli/describe.go)
  13-lrbrl-atribuciones-por-materia.yaml: medida: común (internal/cli/describe.go)
  14-trlrhl-impuestos-por-materia.yaml: medida: común (internal/cli/describe.go)
  15-irpf-rendimientos-por-materia.yaml: medida: común (internal/cli/describe.go)
  16-lrjsp-legalidad-por-materia.yaml: medida: común (internal/cli/describe.go)
  17-ltaibg-plazo-por-materia.yaml: medida: común (internal/cli/describe.go)
  18-lrjpac-norma-derogada.yaml: medida: común (internal/cli/describe.go)
  19-lcsp-contrato-menor-redaccion-cambiada.yaml: medida: común (internal/cli/describe.go)
  20-lcsp-dos-bloques-redaccion-cambiada.yaml: medida: común (internal/cli/describe.go)
  21-sin-binario-ni-servidor.yaml: medida: común (internal/cli/describe.go)
  22-lcsp-contrato-menor-en-2019.yaml: medida: común (internal/cli/describe.go)
  23-lpac-articulo-21-en-2016.yaml: medida: común (internal/cli/describe.go)
evals (jurisprudencia): medidas 10, reutilizadas 0
  01-existe-con-numero-y-fecha.yaml: medida: común (internal/cli/describe.go)
  02-resumen-sin-documento.yaml: medida: común (internal/cli/describe.go)
  03-por-materia.yaml: medida: común (internal/cli/describe.go)
  04-documento-pegado.yaml: medida: común (internal/cli/describe.go)
  05-tribunal-constitucional.yaml: medida: común (internal/cli/describe.go)
  06-documento-que-no-es-el-pedido.yaml: medida: común (internal/cli/describe.go)
  07-resumen-de-una-conocida.yaml: medida: común (internal/cli/describe.go)
  08-doctrina-con-el-fallo-delante.yaml: medida: común (internal/cli/describe.go)
  09-de-que-trata-con-la-ficha-sola.yaml: medida: común (internal/cli/describe.go)
  10-doctrina-dada-por-hecha.yaml: medida: común (internal/cli/describe.go)
evals (legal-core): medidas 12, reutilizadas 0
  01-territorio-municipio-cubierto.yaml: medida: común (internal/cli/describe.go)
  02-territorio-municipio-no-cubierto.yaml: medida: común (internal/cli/describe.go)
  03-no-activa-receta-de-cocina.yaml: medida: común (internal/cli/describe.go)
  04-sin-binario-ni-servidor.yaml: medida: común (internal/cli/describe.go)
  05-plazo-reposicion-leganes.yaml: medida: común (internal/cli/describe.go)
  06-plazo-reposicion-tordesillas.yaml: medida: común (internal/cli/describe.go)
  07-cobertura-liquidacion-tributaria.yaml: medida: común (internal/cli/describe.go)
  08-cobertura-multa-de-trafico.yaml: medida: común (internal/cli/describe.go)
  09-plazo-contencioso-agosto.yaml: medida: común (internal/cli/describe.go)
  10-plazo-reposicion-anterior-a-la-lpac.yaml: medida: común (internal/cli/describe.go)
  11-plazo-reposicion-sin-leer-el-precepto.yaml: medida: común (internal/cli/describe.go)
  12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml: medida: común (internal/cli/describe.go)
## sqlite-en-go-mod
evals (boe-legislacion): medidas 23, reutilizadas 0
  01-lpac-articulo-21.yaml: medida: común (go.mod)
  02-lcsp-contrato-menor.yaml: medida: común (go.mod)
  03-lrbrl-atribuciones-del-pleno.yaml: medida: común (go.mod)
  04-lgt-prescripcion.yaml: medida: común (go.mod)
  05-trlrhl-impuestos-municipales.yaml: medida: común (go.mod)
  06-irpf-rendimientos-del-trabajo.yaml: medida: común (go.mod)
  07-lrjsp-principio-de-legalidad.yaml: medida: común (go.mod)
  08-ltaibg-plazo-de-resolucion.yaml: medida: común (go.mod)
  09-constitucion-articulo-140.yaml: medida: común (go.mod)
  10-et-vacaciones.yaml: medida: común (go.mod)
  11-no-activa-programacion.yaml: medida: común (go.mod)
  12-no-activa-acuerdo-entre-amigos.yaml: medida: común (go.mod)
  13-lrbrl-atribuciones-por-materia.yaml: medida: común (go.mod)
  14-trlrhl-impuestos-por-materia.yaml: medida: común (go.mod)
  15-irpf-rendimientos-por-materia.yaml: medida: común (go.mod)
  16-lrjsp-legalidad-por-materia.yaml: medida: común (go.mod)
  17-ltaibg-plazo-por-materia.yaml: medida: común (go.mod)
  18-lrjpac-norma-derogada.yaml: medida: común (go.mod)
  19-lcsp-contrato-menor-redaccion-cambiada.yaml: medida: común (go.mod)
  20-lcsp-dos-bloques-redaccion-cambiada.yaml: medida: común (go.mod)
  21-sin-binario-ni-servidor.yaml: medida: común (go.mod)
  22-lcsp-contrato-menor-en-2019.yaml: medida: común (go.mod)
  23-lpac-articulo-21-en-2016.yaml: medida: común (go.mod)
evals (jurisprudencia): medidas 10, reutilizadas 0
  01-existe-con-numero-y-fecha.yaml: medida: común (go.mod)
  02-resumen-sin-documento.yaml: medida: común (go.mod)
  03-por-materia.yaml: medida: común (go.mod)
  04-documento-pegado.yaml: medida: común (go.mod)
  05-tribunal-constitucional.yaml: medida: común (go.mod)
  06-documento-que-no-es-el-pedido.yaml: medida: común (go.mod)
  07-resumen-de-una-conocida.yaml: medida: común (go.mod)
  08-doctrina-con-el-fallo-delante.yaml: medida: común (go.mod)
  09-de-que-trata-con-la-ficha-sola.yaml: medida: común (go.mod)
  10-doctrina-dada-por-hecha.yaml: medida: común (go.mod)
evals (legal-core): medidas 12, reutilizadas 0
  01-territorio-municipio-cubierto.yaml: medida: común (go.mod)
  02-territorio-municipio-no-cubierto.yaml: medida: común (go.mod)
  03-no-activa-receta-de-cocina.yaml: medida: común (go.mod)
  04-sin-binario-ni-servidor.yaml: medida: común (go.mod)
  05-plazo-reposicion-leganes.yaml: medida: común (go.mod)
  06-plazo-reposicion-tordesillas.yaml: medida: común (go.mod)
  07-cobertura-liquidacion-tributaria.yaml: medida: común (go.mod)
  08-cobertura-multa-de-trafico.yaml: medida: común (go.mod)
  09-plazo-contencioso-agosto.yaml: medida: común (go.mod)
  10-plazo-reposicion-anterior-a-la-lpac.yaml: medida: común (go.mod)
  11-plazo-reposicion-sin-leer-el-precepto.yaml: medida: común (go.mod)
  12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml: medida: común (go.mod)
## descripcion-de-legal-core
evals (boe-legislacion): medidas 23, reutilizadas 0
  01-lpac-articulo-21.yaml: medida: descripción (skills/legal-core/SKILL.md)
  02-lcsp-contrato-menor.yaml: medida: descripción (skills/legal-core/SKILL.md)
  03-lrbrl-atribuciones-del-pleno.yaml: medida: descripción (skills/legal-core/SKILL.md)
  04-lgt-prescripcion.yaml: medida: descripción (skills/legal-core/SKILL.md)
  05-trlrhl-impuestos-municipales.yaml: medida: descripción (skills/legal-core/SKILL.md)
  06-irpf-rendimientos-del-trabajo.yaml: medida: descripción (skills/legal-core/SKILL.md)
  07-lrjsp-principio-de-legalidad.yaml: medida: descripción (skills/legal-core/SKILL.md)
  08-ltaibg-plazo-de-resolucion.yaml: medida: descripción (skills/legal-core/SKILL.md)
  09-constitucion-articulo-140.yaml: medida: descripción (skills/legal-core/SKILL.md)
  10-et-vacaciones.yaml: medida: descripción (skills/legal-core/SKILL.md)
  11-no-activa-programacion.yaml: medida: descripción (skills/legal-core/SKILL.md)
  12-no-activa-acuerdo-entre-amigos.yaml: medida: descripción (skills/legal-core/SKILL.md)
  13-lrbrl-atribuciones-por-materia.yaml: medida: descripción (skills/legal-core/SKILL.md)
  14-trlrhl-impuestos-por-materia.yaml: medida: descripción (skills/legal-core/SKILL.md)
  15-irpf-rendimientos-por-materia.yaml: medida: descripción (skills/legal-core/SKILL.md)
  16-lrjsp-legalidad-por-materia.yaml: medida: descripción (skills/legal-core/SKILL.md)
  17-ltaibg-plazo-por-materia.yaml: medida: descripción (skills/legal-core/SKILL.md)
  18-lrjpac-norma-derogada.yaml: medida: descripción (skills/legal-core/SKILL.md)
  19-lcsp-contrato-menor-redaccion-cambiada.yaml: medida: descripción (skills/legal-core/SKILL.md)
  20-lcsp-dos-bloques-redaccion-cambiada.yaml: medida: descripción (skills/legal-core/SKILL.md)
  21-sin-binario-ni-servidor.yaml: medida: descripción (skills/legal-core/SKILL.md)
  22-lcsp-contrato-menor-en-2019.yaml: medida: descripción (skills/legal-core/SKILL.md)
  23-lpac-articulo-21-en-2016.yaml: medida: descripción (skills/legal-core/SKILL.md)
evals (jurisprudencia): medidas 10, reutilizadas 0
  01-existe-con-numero-y-fecha.yaml: medida: descripción (skills/legal-core/SKILL.md)
  02-resumen-sin-documento.yaml: medida: descripción (skills/legal-core/SKILL.md)
  03-por-materia.yaml: medida: descripción (skills/legal-core/SKILL.md)
  04-documento-pegado.yaml: medida: descripción (skills/legal-core/SKILL.md)
  05-tribunal-constitucional.yaml: medida: descripción (skills/legal-core/SKILL.md)
  06-documento-que-no-es-el-pedido.yaml: medida: descripción (skills/legal-core/SKILL.md)
  07-resumen-de-una-conocida.yaml: medida: descripción (skills/legal-core/SKILL.md)
  08-doctrina-con-el-fallo-delante.yaml: medida: descripción (skills/legal-core/SKILL.md)
  09-de-que-trata-con-la-ficha-sola.yaml: medida: descripción (skills/legal-core/SKILL.md)
  10-doctrina-dada-por-hecha.yaml: medida: descripción (skills/legal-core/SKILL.md)
evals (legal-core): medidas 12, reutilizadas 0
  01-territorio-municipio-cubierto.yaml: medida: descripción (skills/legal-core/SKILL.md)
  02-territorio-municipio-no-cubierto.yaml: medida: descripción (skills/legal-core/SKILL.md)
  03-no-activa-receta-de-cocina.yaml: medida: descripción (skills/legal-core/SKILL.md)
  04-sin-binario-ni-servidor.yaml: medida: descripción (skills/legal-core/SKILL.md)
  05-plazo-reposicion-leganes.yaml: medida: descripción (skills/legal-core/SKILL.md)
  06-plazo-reposicion-tordesillas.yaml: medida: descripción (skills/legal-core/SKILL.md)
  07-cobertura-liquidacion-tributaria.yaml: medida: descripción (skills/legal-core/SKILL.md)
  08-cobertura-multa-de-trafico.yaml: medida: descripción (skills/legal-core/SKILL.md)
  09-plazo-contencioso-agosto.yaml: medida: descripción (skills/legal-core/SKILL.md)
  10-plazo-reposicion-anterior-a-la-lpac.yaml: medida: descripción (skills/legal-core/SKILL.md)
  11-plazo-reposicion-sin-leer-el-precepto.yaml: medida: descripción (skills/legal-core/SKILL.md)
  12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml: medida: descripción (skills/legal-core/SKILL.md)
## paquete-sin-asignar
evals (boe-legislacion): medidas 23, reutilizadas 0
  01-lpac-articulo-21.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  02-lcsp-contrato-menor.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  03-lrbrl-atribuciones-del-pleno.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  04-lgt-prescripcion.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  05-trlrhl-impuestos-municipales.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  06-irpf-rendimientos-del-trabajo.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  07-lrjsp-principio-de-legalidad.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  08-ltaibg-plazo-de-resolucion.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  09-constitucion-articulo-140.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  10-et-vacaciones.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  11-no-activa-programacion.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  12-no-activa-acuerdo-entre-amigos.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  13-lrbrl-atribuciones-por-materia.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  14-trlrhl-impuestos-por-materia.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  15-irpf-rendimientos-por-materia.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  16-lrjsp-legalidad-por-materia.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  17-ltaibg-plazo-por-materia.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  18-lrjpac-norma-derogada.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  19-lcsp-contrato-menor-redaccion-cambiada.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  20-lcsp-dos-bloques-redaccion-cambiada.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  21-sin-binario-ni-servidor.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  22-lcsp-contrato-menor-en-2019.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  23-lpac-articulo-21-en-2016.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
evals (jurisprudencia): medidas 10, reutilizadas 0
  01-existe-con-numero-y-fecha.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  02-resumen-sin-documento.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  03-por-materia.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  04-documento-pegado.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  05-tribunal-constitucional.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  06-documento-que-no-es-el-pedido.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  07-resumen-de-una-conocida.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  08-doctrina-con-el-fallo-delante.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  09-de-que-trata-con-la-ficha-sola.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  10-doctrina-dada-por-hecha.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
evals (legal-core): medidas 12, reutilizadas 0
  01-territorio-municipio-cubierto.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  02-territorio-municipio-no-cubierto.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  03-no-activa-receta-de-cocina.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  04-sin-binario-ni-servidor.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  05-plazo-reposicion-leganes.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  06-plazo-reposicion-tordesillas.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  07-cobertura-liquidacion-tributaria.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  08-cobertura-multa-de-trafico.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  09-plazo-contencioso-agosto.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  10-plazo-reposicion-anterior-a-la-lpac.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  11-plazo-reposicion-sin-leer-el-precepto.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
  12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml: medida: sin asignar (internal/sinasignar/sinasignar.go)
## solo-docs
ningún trabajo de evals
## una-eval
evals (boe-legislacion): medidas 0, reutilizadas 23
  01-lpac-articulo-21.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  02-lcsp-contrato-menor.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  03-lrbrl-atribuciones-del-pleno.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  04-lgt-prescripcion.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  05-trlrhl-impuestos-municipales.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  06-irpf-rendimientos-del-trabajo.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  07-lrjsp-principio-de-legalidad.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  08-ltaibg-plazo-de-resolucion.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  09-constitucion-articulo-140.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  10-et-vacaciones.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  11-no-activa-programacion.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  12-no-activa-acuerdo-entre-amigos.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  13-lrbrl-atribuciones-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  14-trlrhl-impuestos-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  15-irpf-rendimientos-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  16-lrjsp-legalidad-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  17-ltaibg-plazo-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  18-lrjpac-norma-derogada.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  19-lcsp-contrato-menor-redaccion-cambiada.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  20-lcsp-dos-bloques-redaccion-cambiada.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  21-sin-binario-ni-servidor.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  22-lcsp-contrato-menor-en-2019.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  23-lpac-articulo-21-en-2016.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
evals (jurisprudencia): medidas 1, reutilizadas 9
  01-existe-con-numero-y-fecha.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  02-resumen-sin-documento.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  03-por-materia.yaml: medida: rutas de su conjunto (evals/jurisprudencia/03-por-materia.yaml)
  04-documento-pegado.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  05-tribunal-constitucional.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  06-documento-que-no-es-el-pedido.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  07-resumen-de-una-conocida.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  08-doctrina-con-el-fallo-delante.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  09-de-que-trata-con-la-ficha-sola.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  10-doctrina-dada-por-hecha.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
evals (legal-core): medidas 0, reutilizadas 12
  01-territorio-municipio-cubierto.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  02-territorio-municipio-no-cubierto.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  03-no-activa-receta-de-cocina.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  04-sin-binario-ni-servidor.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  05-plazo-reposicion-leganes.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  06-plazo-reposicion-tordesillas.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  07-cobertura-liquidacion-tributaria.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  08-cobertura-multa-de-trafico.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  09-plazo-contencioso-agosto.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  10-plazo-reposicion-anterior-a-la-lpac.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  11-plazo-reposicion-sin-leer-el-precepto.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
## referencia-renombrada
evals (boe-legislacion): medidas 0, reutilizadas 23
  01-lpac-articulo-21.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  02-lcsp-contrato-menor.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  03-lrbrl-atribuciones-del-pleno.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  04-lgt-prescripcion.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  05-trlrhl-impuestos-municipales.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  06-irpf-rendimientos-del-trabajo.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  07-lrjsp-principio-de-legalidad.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  08-ltaibg-plazo-de-resolucion.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  09-constitucion-articulo-140.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  10-et-vacaciones.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  11-no-activa-programacion.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  12-no-activa-acuerdo-entre-amigos.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  13-lrbrl-atribuciones-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  14-trlrhl-impuestos-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  15-irpf-rendimientos-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  16-lrjsp-legalidad-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  17-ltaibg-plazo-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  18-lrjpac-norma-derogada.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  19-lcsp-contrato-menor-redaccion-cambiada.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  20-lcsp-dos-bloques-redaccion-cambiada.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  21-sin-binario-ni-servidor.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  22-lcsp-contrato-menor-en-2019.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  23-lpac-articulo-21-en-2016.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
evals (jurisprudencia): medidas 0, reutilizadas 10
  01-existe-con-numero-y-fecha.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  02-resumen-sin-documento.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  03-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  04-documento-pegado.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  05-tribunal-constitucional.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  06-documento-que-no-es-el-pedido.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  07-resumen-de-una-conocida.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  08-doctrina-con-el-fallo-delante.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  09-de-que-trata-con-la-ficha-sola.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  10-doctrina-dada-por-hecha.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
evals (legal-core): medidas 12, reutilizadas 0
  01-territorio-municipio-cubierto.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md, skills/legal-core/references/cobertura_por_materia_renombrada.md)
  02-territorio-municipio-no-cubierto.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md, skills/legal-core/references/cobertura_por_materia_renombrada.md)
  03-no-activa-receta-de-cocina.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md, skills/legal-core/references/cobertura_por_materia_renombrada.md)
  04-sin-binario-ni-servidor.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md, skills/legal-core/references/cobertura_por_materia_renombrada.md)
  05-plazo-reposicion-leganes.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md, skills/legal-core/references/cobertura_por_materia_renombrada.md)
  06-plazo-reposicion-tordesillas.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md, skills/legal-core/references/cobertura_por_materia_renombrada.md)
  07-cobertura-liquidacion-tributaria.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md, skills/legal-core/references/cobertura_por_materia_renombrada.md)
  08-cobertura-multa-de-trafico.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md, skills/legal-core/references/cobertura_por_materia_renombrada.md)
  09-plazo-contencioso-agosto.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md, skills/legal-core/references/cobertura_por_materia_renombrada.md)
  10-plazo-reposicion-anterior-a-la-lpac.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md, skills/legal-core/references/cobertura_por_materia_renombrada.md)
  11-plazo-reposicion-sin-leer-el-precepto.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md, skills/legal-core/references/cobertura_por_materia_renombrada.md)
  12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md, skills/legal-core/references/cobertura_por_materia_renombrada.md)
## etiqueta-evals-sin-nada-elegido
evals (boe-legislacion): medidas 0, reutilizadas 23
  01-lpac-articulo-21.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  02-lcsp-contrato-menor.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  03-lrbrl-atribuciones-del-pleno.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  04-lgt-prescripcion.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  05-trlrhl-impuestos-municipales.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  06-irpf-rendimientos-del-trabajo.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  07-lrjsp-principio-de-legalidad.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  08-ltaibg-plazo-de-resolucion.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  09-constitucion-articulo-140.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  10-et-vacaciones.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  11-no-activa-programacion.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  12-no-activa-acuerdo-entre-amigos.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  13-lrbrl-atribuciones-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  14-trlrhl-impuestos-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  15-irpf-rendimientos-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  16-lrjsp-legalidad-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  17-ltaibg-plazo-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  18-lrjpac-norma-derogada.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  19-lcsp-contrato-menor-redaccion-cambiada.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  20-lcsp-dos-bloques-redaccion-cambiada.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  21-sin-binario-ni-servidor.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  22-lcsp-contrato-menor-en-2019.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  23-lpac-articulo-21-en-2016.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
evals (jurisprudencia): medidas 0, reutilizadas 10
  01-existe-con-numero-y-fecha.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  02-resumen-sin-documento.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  03-por-materia.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  04-documento-pegado.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  05-tribunal-constitucional.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  06-documento-que-no-es-el-pedido.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  07-resumen-de-una-conocida.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  08-doctrina-con-el-fallo-delante.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  09-de-que-trata-con-la-ficha-sola.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  10-doctrina-dada-por-hecha.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
evals (legal-core): medidas 0, reutilizadas 12
  01-territorio-municipio-cubierto.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  02-territorio-municipio-no-cubierto.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  03-no-activa-receta-de-cocina.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  04-sin-binario-ni-servidor.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  05-plazo-reposicion-leganes.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  06-plazo-reposicion-tordesillas.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  07-cobertura-liquidacion-tributaria.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  08-cobertura-multa-de-trafico.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  09-plazo-contencioso-agosto.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  10-plazo-reposicion-anterior-a-la-lpac.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  11-plazo-reposicion-sin-leer-el-precepto.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
  12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml: reutilizada de la ejecución 301 (propuesta #199, cabeza-, 2026-10-22T11:00:00Z)
## suite-completa
evals (boe-legislacion): medidas 23, reutilizadas 0
  01-lpac-articulo-21.yaml: medida: suite completa
  02-lcsp-contrato-menor.yaml: medida: suite completa
  03-lrbrl-atribuciones-del-pleno.yaml: medida: suite completa
  04-lgt-prescripcion.yaml: medida: suite completa
  05-trlrhl-impuestos-municipales.yaml: medida: suite completa
  06-irpf-rendimientos-del-trabajo.yaml: medida: suite completa
  07-lrjsp-principio-de-legalidad.yaml: medida: suite completa
  08-ltaibg-plazo-de-resolucion.yaml: medida: suite completa
  09-constitucion-articulo-140.yaml: medida: suite completa
  10-et-vacaciones.yaml: medida: suite completa
  11-no-activa-programacion.yaml: medida: suite completa
  12-no-activa-acuerdo-entre-amigos.yaml: medida: suite completa
  13-lrbrl-atribuciones-por-materia.yaml: medida: suite completa
  14-trlrhl-impuestos-por-materia.yaml: medida: suite completa
  15-irpf-rendimientos-por-materia.yaml: medida: suite completa
  16-lrjsp-legalidad-por-materia.yaml: medida: suite completa
  17-ltaibg-plazo-por-materia.yaml: medida: suite completa
  18-lrjpac-norma-derogada.yaml: medida: suite completa
  19-lcsp-contrato-menor-redaccion-cambiada.yaml: medida: suite completa
  20-lcsp-dos-bloques-redaccion-cambiada.yaml: medida: suite completa
  21-sin-binario-ni-servidor.yaml: medida: suite completa
  22-lcsp-contrato-menor-en-2019.yaml: medida: suite completa
  23-lpac-articulo-21-en-2016.yaml: medida: suite completa
evals (jurisprudencia): medidas 10, reutilizadas 0
  01-existe-con-numero-y-fecha.yaml: medida: suite completa
  02-resumen-sin-documento.yaml: medida: suite completa
  03-por-materia.yaml: medida: suite completa
  04-documento-pegado.yaml: medida: suite completa
  05-tribunal-constitucional.yaml: medida: suite completa
  06-documento-que-no-es-el-pedido.yaml: medida: suite completa
  07-resumen-de-una-conocida.yaml: medida: suite completa
  08-doctrina-con-el-fallo-delante.yaml: medida: suite completa
  09-de-que-trata-con-la-ficha-sola.yaml: medida: suite completa
  10-doctrina-dada-por-hecha.yaml: medida: suite completa
evals (legal-core): medidas 12, reutilizadas 0
  01-territorio-municipio-cubierto.yaml: medida: suite completa
  02-territorio-municipio-no-cubierto.yaml: medida: suite completa
  03-no-activa-receta-de-cocina.yaml: medida: suite completa
  04-sin-binario-ni-servidor.yaml: medida: suite completa
  05-plazo-reposicion-leganes.yaml: medida: suite completa
  06-plazo-reposicion-tordesillas.yaml: medida: suite completa
  07-cobertura-liquidacion-tributaria.yaml: medida: suite completa
  08-cobertura-multa-de-trafico.yaml: medida: suite completa
  09-plazo-contencioso-agosto.yaml: medida: suite completa
  10-plazo-reposicion-anterior-a-la-lpac.yaml: medida: suite completa
  11-plazo-reposicion-sin-leer-el-precepto.yaml: medida: suite completa
  12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml: medida: suite completa
`
