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
// §5): la base es el repositorio; la cabeza, el repositorio con el mutante en
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
	referencia := ficheroDelRepositorio(t, repositorio, "skills/legal-core/references", esMarkdown)

	cli := ficheroDelRepositorio(t, repositorio, "internal/cli", esCodigoSinTests)
	docs := ficheroDelRepositorio(t, repositorio, "docs", esMarkdown)

	casos := []struct {
		nombre    string
		mutante   mutanteDeLaPropuesta
		esperadas []string
		motivo    motivoDeLaEleccion
	}{
		{
			nombre:    "referencia-de-legal-core",
			mutante:   conUnaLineaMas(t, repositorio, referencia),
			esperadas: evalsDeLasSkills(evals, "legal-core"),
			motivo:    motivoConjunto,
		},
		{
			nombre:    "cuerpo-de-boe-legislacion",
			mutante:   conUnaLineaMas(t, repositorio, "skills/boe-legislacion/SKILL.md"),
			esperadas: evalsDeLasSkills(evals, "boe-legislacion", "legal-core", "jurisprudencia"),
			motivo:    motivoConjunto,
		},
		{
			nombre:    "cuerpo-de-jurisprudencia",
			mutante:   conUnaLineaMas(t, repositorio, "skills/jurisprudencia/SKILL.md"),
			esperadas: evalsDeLasSkills(evals, "jurisprudencia"),
			motivo:    motivoConjunto,
		},
		{
			nombre:    "internal-cli",
			mutante:   conUnaLineaMas(t, repositorio, cli),
			esperadas: evals,
			motivo:    motivoComun,
		},
		{
			nombre:    "sqlite-en-go-mod",
			mutante:   conOtraVersionDeSqlite(t, repositorio, base),
			esperadas: evals,
			motivo:    motivoComun,
		},
		{
			nombre:    "descripcion-de-legal-core",
			mutante:   conOtraDescripcionDeLegalCore(t, repositorio),
			esperadas: evals,
			motivo:    motivoDescripcion,
		},
		{
			nombre:    "paquete-sin-asignar",
			mutante:   conUnPaqueteNuevo(t, repositorio),
			esperadas: evals,
			motivo:    motivoSinAsignar,
		},
		{
			nombre:  "solo-docs",
			mutante: conUnaLineaMas(t, repositorio, docs),
		},
		{
			nombre:    "una-eval",
			mutante:   conUnaLineaMas(t, repositorio, "evals/jurisprudencia/03-por-materia.yaml"),
			esperadas: []string{"jurisprudencia/03-por-materia"},
			motivo:    motivoConjunto,
		},
		{
			nombre:    "referencia-renombrada",
			mutante:   conLaReferenciaRenombrada(t, repositorio, referencia),
			esperadas: evalsDeLasSkills(evals, "legal-core"),
			motivo:    motivoConjunto,
		},
	}

	for _, caso := range casos {
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
			assert.Equal(t, caso.esperadas, nombresDeLasElegidas(eleccion), "las evals elegidas, eval por eval")

			var rutas []string

			for _, cambio := range mutante.cambios {
				rutas = append(rutas, cambio.ruta)
				if cambio.anterior != "" {
					rutas = append(rutas, cambio.anterior)
				}
			}

			slices.Sort(rutas)

			for _, elegida := range eleccion.elegidas {
				assert.Equalf(t, caso.motivo, elegida.motivo, "el motivo de %s", elegida.nombre())
				assert.Equalf(t, rutas, elegida.rutas, "las rutas que eligen %s", elegida.nombre())
			}
		})
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
