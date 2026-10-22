package evals

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Lo que gh run view --log pone delante de cada línea del registro de un
// trabajo, en los registros sintéticos de los tests de la plataforma: el
// trabajo y el paso, cada uno con un tabulador detrás, y la hora con un espacio.
const (
	trabajoDelRegistro = "evals (boe-legislacion)\t"
	pasoDelRegistro    = "Ejecutar las evals\t"
	horaDeUnRegistro   = "2026-10-22T09:41:03.0000001Z "
)

// lineaDelRegistro es una línea del registro sintético de un trabajo: el
// trabajo, el paso, la marca de orden de bytes si empieza un trozo, la hora y
// el texto.
func lineaDelRegistro(texto string, empiezaUnTrozo bool) string {
	linea := trabajoDelRegistro + pasoDelRegistro
	if empiezaUnTrozo {
		linea += marcaDeOrdenDeBytes
	}

	return linea + horaDeUnRegistro + texto + "\n"
}

// registroConElInforme es el registro sintético de un trabajo que imprimió las
// líneas dadas entre las dos marcas del informe, con líneas de otros pasos
// delante y detrás. La marca de inicio y la segunda línea del informe empiezan
// un trozo, con la marca de orden de bytes delante de la hora.
func registroConElInforme(lineas []string) string {
	var registro strings.Builder

	registro.WriteString(lineaDelRegistro("##[group]Run make evals", false))
	registro.WriteString(lineaDelRegistro("--- inicio de informe.json ---", true))

	for indice, linea := range lineas {
		registro.WriteString(lineaDelRegistro(linea, indice == 1))
	}

	registro.WriteString(lineaDelRegistro("--- fin de informe.json ---", false))
	registro.WriteString(trabajoDelRegistro + "El informe del trabajo está escrito\t" + horaDeUnRegistro + "Post job cleanup.\n")

	return registro.String()
}

// lineasDeUnInforme son las de un informe.json sintético, con la sangría con
// la que lo escribe el job.
var lineasDeUnInforme = []string{
	"{",
	`  "skill": "boe-legislacion",`,
	`  "tasas": []`,
	"}",
}

// consultaDeLaPlataforma es una de las lecturas de la plataforma, con lo que
// lee como valor.
type consultaDeLaPlataforma func(ctx context.Context, plataforma lecturasDeLaPlataforma) (any, error)

// TestOrdenesDeLaPlataforma fija las órdenes de gh con las que se lee la
// plataforma (contracts/job-de-evals.md §5; research D15, V7), con el
// sustituto de gh como único programa del PATH de las órdenes, un subtest por
// lectura: cada una da a gh los argumentos de su orden constante, con el valor
// de cada variable entero y sin interpretarlo, aunque lleve espacios, comillas o
// $( ), y ese valor sustituye al que el entorno tenga con el mismo nombre; lee
// la respuesta de gh, la de @tsv sin sus escapes; y una orden que termina con
// otro código que 0 es un *errorDeUnaOrden con la orden, sus variables y su
// salida de error, y su texto lleva además cómo terminó y su salida estándar.
func TestOrdenesDeLaPlataforma(t *testing.T) {
	t.Parallel()

	// Los valores que el entorno de la ejecución ya tiene no son los de la orden.
	base := sobreLaBase(os.Environ(), []string{
		"PATH=" + escribirElGh(t), "PROPUESTA=la-del-entorno", "COMMIT_EVALUADO=el-del-entorno",
		"EJECUCION=la-del-entorno", "TRABAJO=el-del-entorno", "EJECUCION_ANTERIOR=la-del-entorno",
	})

	const commit = `6ab3add $(printf inyectado) "con comillas"`

	casos := []struct {
		nombre     string
		consultar  consultaDeLaPlataforma
		respuesta  string
		leido      any
		argumentos []string
		orden      string
		variables  []string
	}{
		{
			nombre: "los-ficheros-de-la-propuesta",
			consultar: func(ctx context.Context, plataforma lecturasDeLaPlataforma) (any, error) {
				return plataforma.ficherosDeLaPropuesta(ctx, 184)
			},
			respuesta: "docs/ROADMAP.md\t\n" +
				"internal/evals/nuevo.go\tinternal/evals/viejo.go\n" +
				`evals/con\\barra.yaml` + "\t" + `evals/con\ttabulador.yaml` + "\n",
			leido: []cambioDeLaPropuesta{
				{ruta: "docs/ROADMAP.md"},
				{ruta: "internal/evals/nuevo.go", anterior: "internal/evals/viejo.go"},
				{ruta: `evals/con\barra.yaml`, anterior: "evals/con\ttabulador.yaml"},
			},
			argumentos: []string{
				"api", "--paginate", "repos/{owner}/{repo}/pulls/184/files",
				"--jq", `.[] | [.filename, (.previous_filename // "")] | @tsv`,
			},
			orden:     ordenDeLosFicherosDeLaPropuesta,
			variables: []string{"PROPUESTA=184"},
		},
		{
			nombre: "las-propuestas-fusionadas",
			consultar: func(ctx context.Context, plataforma lecturasDeLaPlataforma) (any, error) {
				return plataforma.propuestasFusionadas(ctx)
			},
			respuesta: "184\t61b8b5600aa1\t2026-10-22T09:15:00Z\n182\t7fac5bf00bb2\t2026-10-21T18:40:12Z\n",
			leido: []propuestaFusionada{
				{numero: 184, cabeza: "61b8b5600aa1", fusion: time.Date(2026, 10, 22, 9, 15, 0, 0, time.UTC)},
				{numero: 182, cabeza: "7fac5bf00bb2", fusion: time.Date(2026, 10, 21, 18, 40, 12, 0, time.UTC)},
			},
			argumentos: []string{
				"api", "repos/{owner}/{repo}/pulls?state=closed&base=main&sort=updated&direction=desc&per_page=100",
				"--jq", `.[] | select(.merged_at != null) | [.number, .head.sha, .merged_at] | @tsv`,
			},
			orden: ordenDeLasPropuestasFusionadas,
		},
		{
			nombre: "las-ejecuciones-del-flujo",
			consultar: func(ctx context.Context, plataforma lecturasDeLaPlataforma) (any, error) {
				return plataforma.ejecucionesDelFlujo(ctx)
			},
			respuesta: `[{"conclusion":"","createdAt":"2026-10-22T10:02:05Z","databaseId":18712399999,` +
				`"headSha":"7fac5bf00bb2","status":"in_progress"},{"conclusion":"success",` +
				`"createdAt":"2026-10-22T09:20:31Z","databaseId":18712345678,"headSha":"61b8b5600aa1",` +
				`"status":"completed"}]` + "\n",
			leido: []ejecucionDelFlujo{
				{
					ID: 18712399999, Commit: "7fac5bf00bb2", Status: "in_progress",
					Creada: time.Date(2026, 10, 22, 10, 2, 5, 0, time.UTC),
				},
				{
					ID: 18712345678, Commit: "61b8b5600aa1", Status: statusTerminado, Conclusion: conclusionDeExito,
					Creada: time.Date(2026, 10, 22, 9, 20, 31, 0, time.UTC),
				},
			},
			argumentos: []string{
				"run", "list", "--workflow", "evals.yml", "--limit", "200",
				"--json", "databaseId,headSha,status,conclusion,createdAt",
			},
			orden: ordenDeLasEjecucionesDelFlujo,
		},
		{
			nombre: "las-ejecuciones-del-commit",
			consultar: func(ctx context.Context, plataforma lecturasDeLaPlataforma) (any, error) {
				return plataforma.ejecucionesDelCommit(ctx, commit)
			},
			respuesta: `[{"conclusion":"failure","createdAt":"2026-10-22T09:20:31Z","databaseId":18712345678,` +
				`"headSha":"6ab3add","status":"completed"}]` + "\n",
			leido: []ejecucionDelFlujo{
				{
					ID: 18712345678, Commit: "6ab3add", Status: statusTerminado, Conclusion: "failure",
					Creada: time.Date(2026, 10, 22, 9, 20, 31, 0, time.UTC),
				},
			},
			argumentos: []string{
				"run", "list", "--workflow", "evals.yml", "--commit", commit, "--limit", "100",
				"--json", "databaseId,headSha,status,conclusion,createdAt",
			},
			orden:     ordenDeLasEjecucionesDeUnCommit,
			variables: []string{"COMMIT_EVALUADO=" + commit},
		},
		{
			nombre: "los-trabajos-de-una-ejecucion",
			consultar: func(ctx context.Context, plataforma lecturasDeLaPlataforma) (any, error) {
				return plataforma.trabajos(ctx, 18712345678)
			},
			respuesta: `{"jobs":[{"completedAt":"2026-10-22T11:02:44Z","conclusion":"failure",` +
				`"databaseId":53398112233,"name":"evals (boe-legislacion)","startedAt":"2026-10-22T09:21:02Z",` +
				`"status":"completed","steps":[{"conclusion":"failure","name":"Ejecutar las evals","number":7,` +
				`"status":"completed"},{"conclusion":"success","name":"El informe del trabajo está escrito",` +
				`"number":8,"status":"completed"}],"url":"https://github.com/o/r/actions/runs/1/job/2"}]}` + "\n",
			leido: []trabajoDeLaEjecucion{
				{
					ID: 53398112233, Name: "evals (boe-legislacion)", Status: statusTerminado, Conclusion: "failure",
					Steps: []pasoDeLaEjecucion{
						{Name: "Ejecutar las evals", Conclusion: "failure"},
						{Name: "El informe del trabajo está escrito", Conclusion: conclusionDeExito},
					},
				},
			},
			argumentos: []string{"run", "view", "18712345678", "--json", "jobs"},
			orden:      ordenDeLosTrabajos,
			variables:  []string{"EJECUCION_ANTERIOR=18712345678"},
		},
		{
			nombre: "el-informe-de-un-trabajo",
			consultar: func(ctx context.Context, plataforma lecturasDeLaPlataforma) (any, error) {
				return plataforma.informeDelTrabajo(ctx, 18712345678, 53398112233)
			},
			respuesta:  registroConElInforme(lineasDeUnInforme),
			leido:      []byte(strings.Join(lineasDeUnInforme, "\n") + "\n"),
			argumentos: []string{"run", "view", "18712345678", "--job", "53398112233", "--log"},
			orden:      ordenDelRegistroDeUnTrabajo,
			variables:  []string{"EJECUCION=18712345678", "TRABAJO=53398112233"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			respuesta := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(respuesta, respuestaDeGh), []byte(caso.respuesta), 0o600))

			gh := ghDeLaPlataforma{entorno: sobreLaBase(base, []string{variableDeLaRespuestaDeGh + "=" + respuesta})}

			leido, err := caso.consultar(t.Context(), gh)
			require.NoError(t, err)
			assert.Equal(t, caso.leido, leido)

			argumentos, err := leerFichero(filepath.Join(respuesta, argumentosDeGh))
			require.NoError(t, err)
			assert.Equal(t, caso.argumentos, strings.Split(strings.TrimSuffix(string(argumentos), "\n"), "\n"))

			fallido := ghDeLaPlataforma{entorno: sobreLaBase(base, []string{variableDeGhQueFalla + "=si"})}

			_, err = caso.consultar(t.Context(), fallido)

			var deLaOrden *errorDeUnaOrden

			require.ErrorAs(t, err, &deLaOrden)
			assert.Equal(t, caso.orden, deLaOrden.orden)
			assert.Equal(t, caso.variables, deLaOrden.variables)
			assert.Equal(t, "HTTP 403: Resource not accessible by integration\n", deLaOrden.salidaDeError)
			require.ErrorContains(t, err, caso.orden)
			require.ErrorContains(t, err, "exit status 4")
			require.ErrorContains(t, err, "lo escrito antes del error")
			require.ErrorContains(t, err, "HTTP 403: Resource not accessible by integration")

			for _, variable := range caso.variables {
				require.ErrorContains(t, err, variable)
			}
		})
	}
}

// TestInformeDelRegistro fija cómo se saca el informe.json del registro de un
// trabajo, como informe_del_trabajo de scripts/workflow/comun.sh: de cada línea
// se quitan el trabajo, el paso, la marca de orden de bytes y la hora, y el
// informe es lo que hay entre las dos marcas, línea a línea. Un registro sin la
// marca de inicio, sin la de fin tras ella o con dos informes es un error que
// lo dice.
func TestInformeDelRegistro(t *testing.T) {
	t.Parallel()

	registro := registroConElInforme(lineasDeUnInforme)

	casos := []struct {
		nombre   string
		registro string
		informe  string
		error    string
	}{
		{
			nombre:   "con-prefijos-y-marca-de-orden-de-bytes",
			registro: registro,
			informe:  strings.Join(lineasDeUnInforme, "\n") + "\n",
		},
		{
			nombre: "sin-trabajo-ni-paso-ni-hora",
			registro: "--- inicio de informe.json ---\n" + marcaDeOrdenDeBytes + lineasDeUnInforme[0] + "\n" +
				lineasDeUnInforme[3] + "\n--- fin de informe.json ---\n",
			informe: lineasDeUnInforme[0] + "\n" + lineasDeUnInforme[3] + "\n",
		},
		{
			nombre:   "sin-marcas",
			registro: lineaDelRegistro("##[group]Run make evals", true) + lineaDelRegistro("Post job cleanup.", false),
			error:    `el registro no trae la marca "--- inicio de informe.json ---"`,
		},
		{
			nombre:   "sin-la-marca-de-fin",
			registro: strings.Split(registro, "--- fin de informe.json ---")[0],
			error:    `el registro no trae la marca "--- fin de informe.json ---" tras la de inicio`,
		},
		{
			nombre:   "con-dos-informes",
			registro: registro + registro,
			error:    `el registro trae más de un informe: la marca "--- inicio de informe.json ---" aparece más de una vez`,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			informe, err := informeDelRegistro([]byte(caso.registro))
			if caso.error != "" {
				require.EqualError(t, err, caso.error)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, caso.informe, string(informe))
		})
	}
}
