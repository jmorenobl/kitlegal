package evals

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// prefijoDelInstrumentoSinMedir es, escrito a mano, lo que lleva delante de
// cada línea de la comprobación de la medida el motivo del informe del
// instrumento sin medir (contracts/informe-del-job.md §5 de H24): el prefijo de
// la ejecución y que el instrumento no está medido.
const prefijoDelInstrumentoSinMedir = "de la ejecución, no de la skill: el instrumento no está medido: "

// jobDePrueba es una ejecución del job de los tests de ejecutarElJob, sin
// ninguna sesión con modelo: la skill sintética del caso aprobado de
// TestInforme, con sus dos evals; quien abre las sesiones, de pega, que anota
// cada tanda que se le pide y deja en el directorio de sesiones las dos del caso
// aprobado, ya terminadas; y el votante de los tests del informe, que dice no a
// todo y anota cada voto que se le pide.
type jobDePrueba struct {
	ejecucion EjecucionDelJob
	votante   *votanteDelInforme

	// tandas son las que se le han pedido a quien abre las sesiones, en su
	// orden.
	tandas [][]SesionPlanificada
}

// nuevoJobDePrueba arma la ejecución del job sobre ese directorio de evals, que
// es de un directorio temporal del test y al que añade las dos evals del caso
// aprobado, con ese modelo del juez y esa versión de Claude Code de sus votos.
// El directorio de sesiones no existe hasta que se abre la primera tanda.
func nuevoJobDePrueba(t *testing.T, evals, modeloDelJuez, versionDelJuez string) *jobDePrueba {
	t.Helper()

	aprobado := filepath.Join(casosDeInforme, casoAprobado)
	require.NoError(t, os.CopyFS(evals, os.DirFS(filepath.Join(aprobado, "evals"))))

	job := &jobDePrueba{votante: nuevoVotanteDelInforme(t)}
	sesiones := filepath.Join(t.TempDir(), "sesiones")

	job.ejecucion = EjecucionDelJob{
		Skill: skillDeLasSesiones,
		Evals: evals,
		Plan:  PlanDeEvals{Evals: evalsDe(t, evals), ModeloQueDecide: modeloQueDecide, Repeticiones: 1},
		AbrirLaTanda: func(tanda []SesionPlanificada) (EjecucionDeSesiones, error) {
			job.tandas = append(job.tandas, tanda)

			return EjecucionDeSesiones{}, os.CopyFS(sesiones, os.DirFS(filepath.Join(aprobado, "sesiones")))
		},
		Sesiones:            sesiones,
		Destino:             t.TempDir(),
		Umbral:              1,
		Commit:              commitEvaluado,
		SinPython:           filepath.Join(casosDeInforme, ficheroSinPython),
		Votar:               job.votante.votar,
		ModeloDelJuez:       modeloDelJuez,
		VersionDelJuez:      versionDelJuez,
		ConcurrenciaDelJuez: concurrenciaDelJuezDelInforme,
	}

	return job
}

// exigirSinInforme exige que la ejecución que termina con un error no deje
// ningún informe en su destino ni devuelva ningún veredicto.
func (j *jobDePrueba) exigirSinInforme(t *testing.T, informe Informe) {
	t.Helper()

	assert.Zero(t, informe, "sin informe, ejecutarElJob no devuelve ningún veredicto")
	assert.NoFileExists(t, filepath.Join(j.ejecucion.Destino, "informe.md"))
	assert.NoFileExists(t, filepath.Join(j.ejecucion.Destino, "informe.json"))
}

// TestEjecucionSinMedir es el control de umbral de FR-104 y SC-004 de H24 en lo
// que toca al instrumento, y el de los dos umbrales
// medida_del_juez:afirma_lo_no_leido:… en el job (contracts/informe-del-job.md
// §5 y §9 de H24; contracts/job-de-evals.md §2 de H24; research D11 de H24;
// FR-043, FR-044): el job de una skill con juez comprueba la medida versionada
// antes de abrir ninguna sesión, de evals o del juez.
//
// Sobre una copia de la carpeta del juez en t.TempDir(), con quien abre las
// sesiones y el votante de pega, que cuentan sus llamadas: con cada una de las
// cuatro claves de la medida que no coincide —la rúbrica, los casos, el modelo
// fijado y la versión fijada— y con cada uno de sus dos recuentos distinto de
// 0, las seis mutaciones de TestMedidaVersionada, el informe es el del
// instrumento sin medir, con el motivo que dice cuál, y no se llama ni a quien
// abre las sesiones ni al votante; lo mismo con la medida de otra clase, con
// todo a la vez —un motivo por línea, en su orden— y con la medida que no se
// puede leer. Y con la medida que corresponde y se cumple, se llama a quien abre
// las sesiones y el informe publica los dos umbrales de la medida con sus
// recuentos, sin votar ningún caso etiquetado.
//
// Desde H25 (contracts/juez-de-jurisprudencia.md §7 y §9 de H25; research V13 de
// H25; FR-022, FR-032, FR-104; SC-004), lo fija con la carpeta del juez de cada
// fila de la tabla de las copias, y no solo con la de la primera: con la de
// jurisprudencia, las seis mutaciones dejan el job en fallo sin abrir ninguna
// sesión, con el motivo que dice cuál y sin ningún umbral, y con su medida
// versionada, que corresponde y se cumple, el job juzga sin medir y publica sus
// dos umbrales con 0 de 125 y 0 de 124.
func TestEjecucionSinMedir(t *testing.T) {
	t.Parallel()

	modelo, version := fijadosParaElJuez(t)

	medidos := juecesMedidos()
	require.Len(t, medidos, len(copiasDelJuez), "un juez medido por cada fila de la tabla de las copias")

	for _, medido := range medidos {
		t.Run(medido.pareja.skill(), func(t *testing.T) {
			t.Parallel()

			probarLaEjecucionSinMedirDe(t, medido, modelo, version)
		})
	}
}

// juezMedido es el juez de una fila de la tabla de las copias con lo que los
// tests de la ejecución saben de él, escrito a mano: los casos etiquetados como
// defecto y como correctos de su medida versionada, que son los totales de sus
// dos umbrales, y el voto que dice no en sus clases, con la forma de su esquema.
type juezMedido struct {
	pareja              parejaDeCarpetas
	defectos, correctos int
	queNo               func(t *testing.T) grabacion
}

// juecesMedidos son los jueces de la tabla de las copias, en su orden: el de
// boe-legislacion, con los 212 y los 47 casos de su medida (FR-044 de H24) y el
// voto de sus dos clases, y el de jurisprudencia, con los 125 y los 124 de la
// suya (contracts/juez-de-jurisprudencia.md §5 de H25; FR-022) y el voto cuyo
// esquema lleva sentencia.
func juecesMedidos() []juezMedido {
	return []juezMedido{
		{
			pareja:   copiasDelJuez[0],
			defectos: 212, correctos: 47,
			queNo: func(t *testing.T) grabacion {
				t.Helper()

				return votoDeLasDosClases(t, afirmaQueNo(), cuentaQueNo())
			},
		},
		{
			pareja:   copiasDelJuez[1],
			defectos: defectosDeJurisprudencia, correctos: correctosDeJurisprudencia,
			queNo: func(t *testing.T) grabacion {
				t.Helper()

				return votoConSentencia(t, afirmaConSentenciaQueNo(), existeQueNo())
			},
		},
	}
}

// probarLaEjecucionSinMedirDe ve, con la carpeta del juez de esa fila, lo que
// TestEjecucionSinMedir fija: cada medida que no corresponde o no se cumple deja
// el job sin abrir ninguna sesión, y la versionada, que corresponde y se cumple,
// lo deja juzgar sin medir.
func probarLaEjecucionSinMedirDe(t *testing.T, medido juezMedido, modelo, version string) {
	t.Helper()

	mutaciones := mutacionesDeLaMedida(modelo, version)
	require.Len(t, mutaciones, 6, "las cuatro claves que no coinciden y los dos recuentos distintos de 0")

	for _, mutacion := range slices.Concat(mutaciones, otrasMedidasSinCorresponder(modelo, version)) {
		t.Run(mutacion.nombre, func(t *testing.T) {
			t.Parallel()

			probarElJobSinMedir(t, medido.pareja, modelo, version, mutacion)
		})
	}

	t.Run("medida-que-corresponde", func(t *testing.T) {
		t.Parallel()

		probarElJobConElInstrumentoMedido(t, medido, modelo, version)
	})
}

// probarElJobSinMedir copia la carpeta del juez de la pareja, exige que la
// copia sin tocar corresponda y se cumpla, hace los cambios del caso y ejecuta
// el job con el modelo y la versión del caso: exige que no se llame a quien abre
// las sesiones ni al votante, que no se cree el directorio de sesiones y que el
// informe sea el del instrumento sin medir, con un motivo por cada línea del
// caso.
func probarElJobSinMedir(t *testing.T, pareja parejaDeCarpetas, modelo, version string, mutacion mutacionDeLaMedida) {
	t.Helper()

	evals := copiarLaCarpetaDelJuezDe(t, pareja)
	require.Empty(t, comprobarLaMedida(juezDe(t, evals), modelo, version),
		"premisa: la medida de la copia sin tocar corresponde y se cumple")

	for _, cambio := range mutacion.cambios {
		cambio(t, filepath.Join(evals, carpetaDelJuez))
	}

	job := nuevoJobDePrueba(t, evals, mutacion.modelo, mutacion.version)

	informe, err := ejecutarElJob(job.ejecucion)
	require.NoError(t, err)

	assert.Empty(t, job.tandas, "sin una medida que corresponda y se cumpla no se llama a quien abre las sesiones")
	assert.Empty(t, job.votante.pedidos(), "ni se pide ningún voto")
	assert.NoDirExists(t, job.ejecucion.Sesiones, "no se ha abierto ninguna sesión")

	leido := leerInformeEscrito(t, job.ejecucion.Destino, informe)

	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)

	if mutacion.soloElPrincipio {
		require.Len(t, leido.informe.Motivos, 1, strings.Join(leido.informe.Motivos, "\n"))
		assert.True(t, strings.HasPrefix(leido.informe.Motivos[0], prefijoDelInstrumentoSinMedir+mutacion.lineas[0]),
			leido.informe.Motivos[0])
	} else {
		motivos := make([]string, 0, len(mutacion.lineas))
		for _, linea := range mutacion.lineas {
			motivos = append(motivos, prefijoDelInstrumentoSinMedir+linea)
		}

		exigirMotivosDeLaRaiz(t, leido, motivos...)
	}

	exigirNadaMedido(t, leido, job.ejecucion)
}

// exigirNadaMedido exige lo que el informe del instrumento sin medir lleva
// además de su veredicto y de sus motivos (contracts/informe-del-job.md §5 de
// H24; research D11 de H24): ningún umbral; el juez con el modelo y la versión
// recibidos y sin ninguna respuesta, ni juzgada ni sin juzgar; y las series,
// las sesiones, las sesiones sin medir, las invocaciones fuera de lo grabado y
// las llegadas a la red, vacías. Cada lista es una lista vacía en informe.json,
// nunca null, y su sección de informe.md dice que no hay ninguna.
func exigirNadaMedido(t *testing.T, leido informeLeido, ejecucion EjecucionDelJob) {
	t.Helper()

	exigirUmbrales(t, leido, nil)

	assert.Equal(t, &JuezInformado{
		Modelo:              ejecucion.ModeloDelJuez,
		VersionDeClaudeCode: ejecucion.VersionDelJuez,
		Respuestas:          []RespuestaConVotos{},
		SinJuzgar:           []RespuestaSinJuzgar{},
		VotosCortados:       []VotoCortado{},
	}, leido.informe.Juez)
	assert.Equal(t, `{"modelo":`+cadenaJSON(t, ejecucion.ModeloDelJuez)+`,"version_de_claude_code":`+
		cadenaJSON(t, ejecucion.VersionDelJuez)+`,"respuestas":[],"sin_juzgar":[],"votos_cortados":[]}`,
		compacto(t, leido.crudo.Juez))
	assert.Equal(t, "Modelo: "+ejecucion.ModeloDelJuez+"\n\nVersión de Claude Code: "+ejecucion.VersionDelJuez+
		"\n\nVotos: ninguna\n\nRespuestas sin juzgar: ninguna\n\nVotos cortados por el tope y pedidos otra vez: ninguna",
		seccionDelInforme(t, leido.md, "Juez"))

	var raiz map[string]jsontext.Value

	require.NoError(t, json.Unmarshal([]byte(contenidoDelInforme(t, ejecucion.Destino, "informe.json")), &raiz))

	for _, clave := range []string{"tasas", "evals", "sesiones_sin_medir", "fuera_de_lo_grabado", "red"} {
		require.Contains(t, raiz, clave)
		assert.Equal(t, "[]", compacto(t, raiz[clave]), "%s es una lista vacía en informe.json", clave)
	}

	for _, seccion := range []string{
		"Tasas por eval", "Sesiones sin medir", "Sesiones", "Invocaciones fuera de lo grabado",
	} {
		assert.Equal(t, "ninguna", seccionDelInforme(t, leido.md, seccion), "la sección %s de informe.md", seccion)
	}

	assert.Equal(t, "ninguno", seccionDelInforme(t, leido.md, "Umbrales"))
	assert.Equal(t, "ninguna petición llegó a la red de una fuente", seccionDelInforme(t, leido.md, "Peticiones llegadas a la red"))
	assert.Zero(t, leido.informe.DuracionDeLasSesiones, "sin sesiones no hay duración que publicar")
}

// probarElJobConElInstrumentoMedido ejecuta el job con la copia de la carpeta
// del juez medido sin tocar, cuya medida corresponde y se cumple con el modelo
// y la versión que fija la definición del job, y con el votante diciendo no con
// la forma del esquema de ese juez: exige que se llame a quien abre las
// sesiones, una vez por tanda y con las sesiones del plan; que el informe sea
// el de sus sesiones, aprobado, con los dos umbrales de la medida versionada
// con sus recuentos (FR-044 de H24; FR-022 de H25); y que al votante solo se le
// pida el voto de la respuesta que se juzga, la de la sesión de la eval que
// activa la skill, y el de ningún caso etiquetado.
func probarElJobConElInstrumentoMedido(t *testing.T, medido juezMedido, modelo, version string) {
	t.Helper()

	evals := copiarLaCarpetaDelJuezDe(t, medido.pareja)
	job := nuevoJobDePrueba(t, evals, modelo, version)
	job.votante.queNo = medido.queNo(t).salida

	informe, err := ejecutarElJob(job.ejecucion)
	require.NoError(t, err)

	require.Len(t, job.tandas, 1, "el plan de un solo modo es una tanda, y se abre")
	assert.Equal(t, job.ejecucion.Plan.Sesiones(), job.tandas[0])

	leido := leerInformeEscrito(t, job.ejecucion.Destino, informe)

	exigirMotivosDeLaRaiz(t, leido)
	assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)
	assert.Len(t, leido.informe.Evals, 2, "las dos sesiones que se abrieron")

	deLaMedida := slices.DeleteFunc(slices.Clone(leido.informe.Umbrales), func(umbral Umbral) bool {
		return !strings.HasPrefix(umbral.Nombre, "medida_del_juez:")
	})
	assert.Equal(t, umbralesDeLaMedidaDelJuez(0, medido.defectos, 0, medido.correctos), deLaMedida,
		"los dos umbrales de la medida, con los recuentos de la versionada")

	assert.Equal(t, &JuezInformado{
		Modelo: modelo, VersionDeClaudeCode: version, Respuestas: []RespuestaConVotos{}, SinJuzgar: []RespuestaSinJuzgar{},
		VotosCortados: []VotoCortado{},
	}, leido.informe.Juez)

	pedidos := job.votante.pedidos()
	require.Len(t, pedidos, 1, "un voto, el de la única respuesta que se juzga, y ninguno de un caso etiquetado")
	assert.Contains(t, pedidos[0], preguntaDeLasSinteticas)
}

// TestEjecutarElJob fija lo demás del recorrido del job (contracts/job-de-evals.md
// §2 de H24; data-model §6 de H24), con quien abre las sesiones y el votante de
// pega: con una skill sin juez no comprueba ninguna medida ni necesita votante,
// abre las sesiones y escribe su informe, sin juez; el error de quien abre las
// sesiones es el de la ejecución, sin informe; y con un plan sin sentido o unas
// evals que no se pueden leer, termina con su error antes de abrir ninguna.
func TestEjecutarElJob(t *testing.T) {
	t.Parallel()

	modelo, version := fijadosParaElJuez(t)

	t.Run("skill-sin-juez", func(t *testing.T) {
		t.Parallel()

		job := nuevoJobDePrueba(t, t.TempDir(), "", "")
		job.ejecucion.Votar = nil
		job.ejecucion.ConcurrenciaDelJuez = 0

		informe, err := ejecutarElJob(job.ejecucion)
		require.NoError(t, err)

		require.Len(t, job.tandas, 1)
		assert.Equal(t, job.ejecucion.Plan.Sesiones(), job.tandas[0])
		assert.Empty(t, job.votante.pedidos(), "una skill sin juez no pide ningún voto")

		leido := leerInformeEscrito(t, job.ejecucion.Destino, informe)

		exigirMotivosDeLaRaiz(t, leido)
		assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)
		assert.Nil(t, leido.informe.Juez, "una skill sin juez no tiene juez en el informe")
		exigirUmbrales(t, leido, nil)
		assert.Len(t, leido.informe.Evals, 2)
	})

	t.Run("error-al-abrir-las-sesiones", func(t *testing.T) {
		t.Parallel()

		errDeLaTanda := errors.New("la sesión no se puede abrir")

		job := nuevoJobDePrueba(t, copiarLaCarpetaDelJuez(t), modelo, version)
		job.ejecucion.AbrirLaTanda = func(tanda []SesionPlanificada) (EjecucionDeSesiones, error) {
			job.tandas = append(job.tandas, tanda)

			return EjecucionDeSesiones{}, errDeLaTanda
		}

		informe, err := ejecutarElJob(job.ejecucion)
		require.ErrorIs(t, err, errDeLaTanda)

		assert.Len(t, job.tandas, 1)
		assert.Empty(t, job.votante.pedidos(), "sin sesiones no se pide ningún voto")
		job.exigirSinInforme(t, informe)
	})

	antesDeAbrir := map[string]struct {
		cambiar   func(ejecucion *EjecucionDelJob)
		fragmento string
	}{
		"plan-sin-sentido": {
			cambiar:   func(ejecucion *EjecucionDelJob) { ejecucion.Plan.Repeticiones = 0 },
			fragmento: "las repeticiones son 0 y hace falta al menos una",
		},
		"evals-que-no-se-pueden-leer": {
			cambiar: func(ejecucion *EjecucionDelJob) {
				ejecucion.Evals = filepath.Join(ejecucion.Evals, "no-existe")
			},
			fragmento: "no-existe",
		},
	}

	for nombre, caso := range antesDeAbrir {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			job := nuevoJobDePrueba(t, copiarLaCarpetaDelJuez(t), modelo, version)
			caso.cambiar(&job.ejecucion)

			informe, err := ejecutarElJob(job.ejecucion)
			require.Error(t, err)
			require.ErrorContains(t, err, "el job no se puede ejecutar")
			require.ErrorContains(t, err, caso.fragmento)

			assert.Empty(t, job.tandas, "no se llama a quien abre las sesiones")
			assert.Empty(t, job.votante.pedidos())
			job.exigirSinInforme(t, informe)
		})
	}
}

// TestPathDelJuez fija el PATH de los votos del juez en el job
// (contracts/job-de-evals.md §2 de H24): el del job con el directorio del
// Claude Code del juez delante, de modo que el claude que encuentra el guion del
// voto es ese y no el de las sesiones; solo ese directorio si el job no tiene
// PATH; y un error si la ruta del claude del juez no es absoluta, porque cada
// voto se ejecuta en su propio directorio y desde él se buscaría.
func TestPathDelJuez(t *testing.T) {
	t.Parallel()

	delJuez := filepath.Join(t.TempDir(), "claude-del-juez", "node_modules", ".bin")
	separador := string(os.PathListSeparator)

	path, err := pathDelJuez(filepath.Join(delJuez, "claude"), "/usr/local/bin"+separador+"/usr/bin")
	require.NoError(t, err)
	assert.Equal(t, delJuez+separador+"/usr/local/bin"+separador+"/usr/bin", path)

	path, err = pathDelJuez(filepath.Join(delJuez, "claude"), "")
	require.NoError(t, err)
	assert.Equal(t, delJuez, path, "sin PATH del job, ningún elemento vacío, que sería el directorio del voto")

	relativa := filepath.Join("claude-del-juez", "node_modules", ".bin", "claude")

	path, err = pathDelJuez(relativa, "/usr/bin")
	require.Error(t, err)
	require.ErrorContains(t, err, "el claude del juez "+relativa+" no tiene ruta absoluta")
	assert.Empty(t, path)
}
