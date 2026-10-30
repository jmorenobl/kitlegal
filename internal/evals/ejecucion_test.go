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

	"github.com/jmorenobl/kitlegal/internal/skills"
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

	t.Run("descripciones", func(t *testing.T) {
		t.Parallel()

		probarLaEjecucionSinMedirDeLasDescripciones(t, modelo, version)
	})
}

// prefijoDeLasDescripcionesSinMedir es, escrito a mano, lo que lleva delante de
// cada línea de la comprobación de la medida del juez de las descripciones el
// motivo del informe del instrumento sin medir
// (contracts/juez-de-las-descripciones.md §3 de H31): el del instrumento sin
// medir y el del juez de las descripciones, que distingue sus líneas de las del
// juez de las respuestas.
const prefijoDeLasDescripcionesSinMedir = prefijoDelInstrumentoSinMedir + "juez de las descripciones: "

// probarLaEjecucionSinMedirDeLasDescripciones fija con la medida del juez de las
// descripciones lo que TestEjecucionSinMedir fija con la de cada juez de las
// respuestas (contracts/juez-de-las-descripciones.md §3 de H31; research D8 de
// H31; FR-026; SC-008): con el bloque descripcion en la decisión, juzgada en la
// ejecución o reutilizada, cada una de las seis mutaciones de su medida, sobre
// una copia de la evidencia, deja el job de una skill sin juez en fallo sin
// llamar a quien abre las sesiones ni a quien da el votante de las
// descripciones, con un motivo por línea, cada una con su prefijo, sin ningún
// umbral, ni bloque descripcion, ni juez. Con el juez de la skill medido, solo
// dan motivo las líneas de las descripciones, y el informe lleva el juez de la
// skill; y con los dos sin medir, las del juez de la skill van delante.
func probarLaEjecucionSinMedirDeLasDescripciones(t *testing.T, modelo, version string) {
	t.Helper()

	mutaciones := mutacionesDeLaMedidaDeLasDescripciones(modelo, version)
	require.Len(t, mutaciones, 6, "las seis mutaciones de SC-008 de H31")

	bloques := map[string]descripcionDeLaDecision{
		"juzgada":     descripcionJuzgadaEnLaDecision(),
		"reutilizada": descripcionReutilizadaEnLaDecision(),
	}

	for _, mutacion := range mutaciones {
		for deLaDecision, bloque := range bloques {
			t.Run(mutacion.nombre+"-"+deLaDecision, func(t *testing.T) {
				t.Parallel()

				_, leido := probarElJobSinMedirLaDescripcion(t, t.TempDir(), bloque, mutacion)

				exigirMotivosDeLaRaiz(t, leido, conElPrefijo(prefijoDeLasDescripcionesSinMedir, mutacion.lineas)...)
				assert.Nil(t, leido.informe.Juez, "la skill no tiene juez, tampoco en el informe del instrumento sin medir")
				exigirUmbrales(t, leido, nil)
				assert.Empty(t, leido.informe.Tasas)
				assert.Empty(t, leido.informe.Evals)
			})
		}
	}

	t.Run("con-el-juez-de-la-skill-medido", func(t *testing.T) {
		t.Parallel()

		mutacion := mutaciones[0]
		job, leido := probarElJobSinMedirLaDescripcion(t, copiarLaCarpetaDelJuez(t), descripcionJuzgadaEnLaDecision(),
			mutacion)

		exigirMotivosDeLaRaiz(t, leido, conElPrefijo(prefijoDeLasDescripcionesSinMedir, mutacion.lineas)...)
		exigirNadaMedido(t, leido, job.ejecucion)
	})

	t.Run("con-los-dos-sin-medir", func(t *testing.T) {
		t.Parallel()

		evals := copiarLaCarpetaDelJuez(t)
		cambiarUnByteDe(ficheroDeRubricaDelJuez)(t, filepath.Join(evals, carpetaDelJuez))

		mutacion := mutaciones[1]
		job, leido := probarElJobSinMedirLaDescripcion(t, evals, descripcionJuzgadaEnLaDecision(), mutacion)

		exigirMotivosDeLaRaiz(t, leido, slices.Concat(
			conElPrefijo(prefijoDelInstrumentoSinMedir, []string{lineaDeLaRubrica}),
			conElPrefijo(prefijoDeLasDescripcionesSinMedir, mutacion.lineas))...)
		exigirNadaMedido(t, leido, job.ejecucion)
	})
}

// conElPrefijo son las líneas, cada una con el prefijo delante.
func conElPrefijo(prefijo string, lineas []string) []string {
	conPrefijo := make([]string, 0, len(lineas))
	for _, linea := range lineas {
		conPrefijo = append(conPrefijo, prefijo+linea)
	}

	return conPrefijo
}

// probarElJobSinMedirLaDescripcion copia la evidencia del juez de las
// descripciones, exige que la copia sin tocar corresponda y se cumpla con lo que
// fija la definición del job, hace los cambios de la mutación y ejecuta el job
// sobre ese directorio de evals con el modelo y la versión de la mutación, la
// copia como raíz y la decisión con ese bloque descripcion: exige que no se
// llame a quien abre las sesiones ni a quien da el votante de las descripciones,
// ni se pida ningún voto de las respuestas, que no se cree el directorio de
// sesiones y que el informe sea el del instrumento sin medir, sin bloque
// descripcion ni su sección.
func probarElJobSinMedirLaDescripcion(
	t *testing.T, evals string, bloque descripcionDeLaDecision, mutacion mutacionDeLaMedida,
) (*jobDePrueba, informeLeido) {
	t.Helper()

	modelo, version := fijadosParaElJuez(t)

	raiz := copiarLaEvidenciaDeLasDescripciones(t)
	require.Empty(t, comprobarLaMedida(juezDeLasDescripcionesEn(t, raiz), modelo, version),
		"premisa: la medida de la copia sin tocar corresponde y se cumple")

	for _, cambio := range mutacion.cambios {
		cambio(t, carpetaDeLaEvidenciaEn(raiz))
	}

	job := nuevoJobDePrueba(t, evals, mutacion.modelo, mutacion.version)
	job.conLaDescripcion(raiz, bloque)
	job.ejecucion.AbrirLaTanda = abridorQueNoSeLlama(t)
	job.ejecucion.VotanteDeLasDescripciones = votanteDeLasDescripcionesQueNoSePide(t)

	informe, err := ejecutarElJob(job.ejecucion)
	require.NoError(t, err)

	assert.Empty(t, job.votante.pedidos(), "ni se pide ningún voto de las respuestas")
	assert.NoDirExists(t, job.ejecucion.Sesiones, "no se ha abierto ninguna sesión")

	leido := leerInformeEscrito(t, job.ejecucion.Destino, informe)

	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
	exigirSinLaDescripcion(t, leido)

	return job, leido
}

// abridorQueNoSeLlama es quien abre las sesiones de un job que no tiene que
// abrir ninguna: si se le llama, el test falla, y no abre nada.
func abridorQueNoSeLlama(t *testing.T) func(tanda []SesionPlanificada) (EjecucionDeSesiones, error) {
	t.Helper()

	return func(tanda []SesionPlanificada) (EjecucionDeSesiones, error) {
		t.Errorf("se llama a quien abre las sesiones con una tanda de %d", len(tanda))

		return EjecucionDeSesiones{}, errors.New("no se abre ninguna sesión")
	}
}

// votanteDeLasDescripcionesQueNoSePide es quien da el votante de las
// descripciones de un job que no tiene que votar ninguna: si se le pide, el
// test falla, y el votante que da no vota.
func votanteDeLasDescripcionesQueNoSePide(t *testing.T) func(juez *Juez) Votante {
	t.Helper()

	return func(*Juez) Votante {
		t.Error("se pide el votante de las descripciones")

		return func(string) ([]byte, error) { return nil, errors.New("ningún voto") }
	}
}

// huellaDeLaDescripcionDeLaDecision es la huella, sintética, del bloque
// descripcion de la decisión en los tests de la ejecución.
func huellaDeLaDescripcionDeLaDecision() string {
	return huellaSintetica("descripcion de " + skillDeLasSesiones)
}

// descripcionJuzgadaEnLaDecision y descripcionReutilizadaEnLaDecision son los
// dos bloques descripcion de la decisión de los tests de la ejecución, los de
// contracts/juez-de-las-descripciones.md §2 de H31: el de la descripción de la
// skill de las sesiones que se juzga en la ejecución y el de la que reutiliza
// el juicio de una medida de main, sin marcar, con su procedencia.
func descripcionJuzgadaEnLaDecision() descripcionDeLaDecision {
	return descripcionDeLaDecision{Skill: skillDeLasSesiones, Huella: huellaDeLaDescripcionDeLaDecision(), Juzgada: true}
}

func descripcionReutilizadaEnLaDecision() descripcionDeLaDecision {
	return descripcionDeLaDecision{
		Skill: skillDeLasSesiones, Huella: huellaDeLaDescripcionDeLaDecision(), Procedencia: procedenciaDeMainDelJob(),
		Juicio: &juicioReutilizado{Motivo: motivoDelVotoQueNo},
	}
}

// conLaDescripcion pone en la ejecución del job la raíz de la que se leen el
// juez de las descripciones y la descripción de la skill, y la decisión de la
// suite completa que mide todas las evals de su plan, con ese bloque
// descripcion.
func (j *jobDePrueba) conLaDescripcion(raiz string, bloque descripcionDeLaDecision) {
	j.ejecucion.Raiz = raiz
	j.ejecucion.decision = &origenDeLasEvals{Modo: origenDeLaSuiteCompleta, Descripcion: &bloque}

	for _, eval := range j.ejecucion.Plan.Evals {
		j.ejecucion.decision.Evals = append(j.ejecucion.decision.Evals,
			origenDeEval{Eval: eval.Fichero, Medida: true, Motivo: motivoSuiteCompleta})
	}
}

// exigirSinLaDescripcion exige que el informe no lleve nada de la descripción:
// ni el bloque descripcion en informe.json, ni el umbral de su clase, ni la
// sección «La descripción» en informe.md (contracts/juez-de-las-descripciones.md
// §6 de H31; FR-030).
func exigirSinLaDescripcion(t *testing.T, leido informeLeido) {
	t.Helper()

	assert.Nil(t, leido.informe.Descripcion)

	var raiz map[string]jsontext.Value

	require.NoError(t, json.Unmarshal([]byte(leido.escrito), &raiz))
	assert.NotContains(t, raiz, "descripcion", "informe.json no lleva el bloque descripcion")

	for _, umbral := range leido.informe.Umbrales {
		assert.NotEqual(t, claseHablaEnPrimeraOSegundaPersona, umbral.Nombre, "ni el umbral de su clase")
	}

	assert.NotContains(t, leido.md, "\n## La descripción\n", "ni informe.md su sección")
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
// normativa, con los 212 y los 47 casos de su medida (FR-044 de H24) y el
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
		// Desde H26.1, solo el fuera_de_lo_grabado de su único modelo, sobre sus
		// dos sesiones medidas, que lleva el informe de toda skill
		// (contracts/juicio-e-informe.md §4 de H26.1).
		exigirUmbrales(t, leido, []Umbral{umbralDeFueraDeLoGrabado(modeloQueDecide, ModoOrden, 0, 2)})
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

// TestEjecutarElJobConLaDecision fija que la ejecución del job con la decisión
// de su trabajo abre solo las sesiones de las evals que la decisión mide y
// compone el informe con las que reutiliza (FR-012, FR-014 y FR-022 de H30;
// contracts/job-de-evals.md §3 de H30): con la skill sintética del caso
// aprobado, cuya segunda eval la decisión reutiliza, quien abre las sesiones
// recibe una sola tanda, con las sesiones del plan de la primera, y deja solo
// la suya; el informe la cuenta medida, con una sesión abierta, y la otra
// reutilizada con su procedencia, y aprueba. Una sola sesión más en la tanda
// la deja en rojo.
func TestEjecutarElJobConLaDecision(t *testing.T) {
	t.Parallel()

	job := nuevoJobDePrueba(t, t.TempDir(), "", "")
	job.ejecucion.Votar = nil
	job.ejecucion.ConcurrenciaDelJuez = 0

	require.Len(t, job.ejecucion.Plan.Evals, 2, "la skill del caso aprobado tiene dos evals")

	medida, reutilizada := job.ejecucion.Plan.Evals[0], job.ejecucion.Plan.Evals[1]
	sesion := strings.TrimSuffix(medida.Fichero, ".yaml")

	job.ejecucion.decision = &origenDeLasEvals{Modo: origenConReutilizacion, Evals: []origenDeEval{
		{
			Eval: medida.Fichero, Huella: huellaSintetica(medida.Fichero), Medida: true, Motivo: motivoConjunto,
			Rutas: []string{"evals/" + skillDeLasSesiones + "/" + medida.Fichero},
		},
		{
			Eval: reutilizada.Fichero, Huella: huellaSintetica(reutilizada.Fichero), Procedencia: procedenciaDeMainDelJob(),
			Tasas: []TasaDelInforme{{
				Eval: reutilizada.Fichero, Modelo: modeloQueDecide, Modo: ModoOrden, Planificada: true, Decide: true,
				Formas: []string{}, Sesiones: 1, Pasan: 1, Pasa: true,
			}},
		},
	}}
	job.ejecucion.AbrirLaTanda = func(tanda []SesionPlanificada) (EjecucionDeSesiones, error) {
		job.tandas = append(job.tandas, tanda)

		return EjecucionDeSesiones{}, os.CopyFS(filepath.Join(job.ejecucion.Sesiones, sesion),
			os.DirFS(filepath.Join(casosDeInforme, casoAprobado, "sesiones", sesion)))
	}

	informe, err := ejecutarElJob(job.ejecucion)
	require.NoError(t, err)

	soloLaMedida := job.ejecucion.Plan
	soloLaMedida.Evals = []Eval{medida}

	require.Len(t, job.tandas, 1, "una sola tanda, la del modo orden")
	assert.Equal(t, soloLaMedida.Sesiones(), job.tandas[0], "solo las sesiones de la eval medida")

	leido := leerInformeEscrito(t, job.ejecucion.Destino, informe)

	assert.Equalf(t, VeredictoAprobado, leido.informe.Veredicto, "motivos: %v", leido.informe.Motivos)
	require.NotNil(t, leido.informe.Origen)
	assert.Equal(t, 1, leido.informe.Origen.EvalsMedidas)
	assert.Equal(t, 1, leido.informe.Origen.EvalsReutilizadas)
	require.NotNil(t, leido.informe.Origen.SesionesAbiertas)
	assert.Equal(t, 1, *leido.informe.Origen.SesionesAbiertas)
	require.Len(t, leido.informe.Origen.Evals, 2)
	assert.Equal(t, procedenciaDeMainDelJob(), leido.informe.Origen.Evals[1].Procedencia,
		"la reutilizada lleva la procedencia de la decisión")
	assert.Len(t, leido.informe.Evals, 1, "el informe juzga solo la sesión medida")
}

// TestEjecutarElJobConLaDescripcion fija el recorrido del job con el bloque
// descripcion en la decisión y la medida del juez de las descripciones que
// corresponde y se cumple (contracts/juez-de-las-descripciones.md §3 y §6 de
// H31; research D7 de H31; FR-026, FR-031, FR-032), con quien abre las sesiones
// de pega y un votante simulado, sin ninguna sesión con modelo (FR-033):
//
//   - juzgada: se pide el votante una vez, con el juez de las descripciones de
//     la raíz, antes de abrir las sesiones de las evals, y los votos de la
//     descripción de la skill, cada uno con su mensaje, detrás de ellas; el
//     informe lleva su juicio, marcada con sus tres «si», y el umbral en rojo;
//   - reutilizada: no se pide ningún votante, y el informe lleva el juicio del
//     bloque con su procedencia y el umbral con 0 de 0;
//   - y, antes de abrir ninguna sesión y sin escribir el informe, el error de
//     una raíz sin el juez de las descripciones, de una juzgada sin quien dé su
//     votante o sin el SKILL.md de su skill, y de una reutilizada sin el juicio.
func TestEjecutarElJobConLaDescripcion(t *testing.T) {
	t.Parallel()

	t.Run("juzgada", func(t *testing.T) {
		t.Parallel()

		probarElJobQueJuzgaLaDescripcion(t)
	})

	t.Run("reutilizada", func(t *testing.T) {
		t.Parallel()

		probarElJobQueReutilizaLaDescripcion(t)
	})

	antesDeAbrir := map[string]struct {
		raiz      func(t *testing.T) string
		bloque    descripcionDeLaDecision
		sinQuien  bool
		fragmento string
	}{
		"sin-el-juez-de-las-descripciones": {
			raiz:      func(t *testing.T) string { t.Helper(); return t.TempDir() },
			bloque:    descripcionJuzgadaEnLaDecision(),
			fragmento: "evidencias/adr-0046/rubrica.md",
		},
		"juzgada-sin-quien-de-su-votante": {
			raiz:      raizConLaDescripcion,
			bloque:    descripcionJuzgadaEnLaDecision(),
			sinQuien:  true,
			fragmento: "votante",
		},
		"juzgada-sin-el-skill-md": {
			raiz:      copiarLaEvidenciaDeLasDescripciones,
			bloque:    descripcionJuzgadaEnLaDecision(),
			fragmento: "skills/" + skillDeLasSesiones + "/SKILL.md",
		},
		"reutilizada-sin-el-juicio": {
			raiz: copiarLaEvidenciaDeLasDescripciones,
			bloque: descripcionDeLaDecision{
				Skill: skillDeLasSesiones, Huella: huellaDeLaDescripcionDeLaDecision(), Procedencia: procedenciaDeMainDelJob(),
			},
			fragmento: "sin su juicio",
		},
	}

	for nombre, caso := range antesDeAbrir {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			modelo, version := fijadosParaElJuez(t)

			job := nuevoJobDeLaDescripcion(t, modelo, version, caso.raiz(t), caso.bloque)
			job.ejecucion.AbrirLaTanda = abridorQueNoSeLlama(t)
			job.ejecucion.VotanteDeLasDescripciones = votanteDeLasDescripcionesQueNoSePide(t)

			if caso.sinQuien {
				job.ejecucion.VotanteDeLasDescripciones = nil
			}

			informe, err := ejecutarElJob(job.ejecucion)
			require.Error(t, err)
			require.ErrorContains(t, err, "el job no se puede ejecutar")
			require.ErrorContains(t, err, caso.fragmento)

			assert.NoDirExists(t, job.ejecucion.Sesiones, "no se ha abierto ninguna sesión")
			job.exigirSinInforme(t, informe)
		})
	}
}

// nuevoJobDeLaDescripcion arma la ejecución del job de la skill sintética del
// caso aprobado, que no tiene juez, con ese modelo y esa versión del juez, esa
// raíz y la decisión con ese bloque descripcion.
func nuevoJobDeLaDescripcion(t *testing.T, modelo, version, raiz string, bloque descripcionDeLaDecision) *jobDePrueba {
	t.Helper()

	job := nuevoJobDePrueba(t, t.TempDir(), modelo, version)
	job.ejecucion.Votar = nil
	job.ejecucion.ConcurrenciaDelJuez = 0
	job.conLaDescripcion(raiz, bloque)

	return job
}

// raizConLaDescripcion es la raíz de un repositorio en un directorio temporal
// del test, con la evidencia del juez de las descripciones copiada y el
// SKILL.md de la skill de las sesiones con la descripción de los tests de los
// votos, descripcionDeLosVotos, en un escalar literal que la deja tal cual.
func raizConLaDescripcion(t *testing.T) string {
	t.Helper()

	raiz := copiarLaEvidenciaDeLasDescripciones(t)
	skill := filepath.Join(raiz, "skills", skillDeLasSesiones)
	require.NoError(t, os.MkdirAll(skill, 0o750))

	contenido := "---\nname: " + skillDeLasSesiones + "\ndescription: |-\n  " +
		strings.ReplaceAll(descripcionDeLosVotos, "\n", "\n  ") + "\n---\n\n# " + skillDeLasSesiones + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte(contenido), 0o600))

	frontmatter, err := skills.LeerFrontmatter([]byte(contenido))
	require.NoError(t, err)
	require.Equal(t, descripcionDeLosVotos, frontmatter.Description, "premisa: la descripción del SKILL.md, tal cual")

	return raiz
}

// probarElJobQueJuzgaLaDescripcion ejecuta el job con la descripción juzgada en
// la decisión y un votante que dice tres «si» con su frase: exige que el votante
// se pida una vez, con el juez de las descripciones de la raíz, antes de abrir
// las sesiones, y los votos detrás de ellas, cada uno con el mensaje de la
// descripción; y que el informe lleve su juicio, marcada, con la frase y el
// motivo de su primer voto, y el umbral de su clase en rojo, con su motivo.
func probarElJobQueJuzgaLaDescripcion(t *testing.T) {
	t.Helper()

	modelo, version := fijadosParaElJuez(t)

	job := nuevoJobDeLaDescripcion(t, modelo, version, raizConLaDescripcion(t), descripcionJuzgadaEnLaDecision())

	var sucesos []string

	abrir := job.ejecucion.AbrirLaTanda
	job.ejecucion.AbrirLaTanda = func(tanda []SesionPlanificada) (EjecucionDeSesiones, error) {
		sucesos = append(sucesos, "tanda")

		return abrir(tanda)
	}

	votante := &votanteGrabado{t: t, grabaciones: []grabacion{
		votoDeLaDescripcion(t, "si", fraseDelPrimerVoto, motivoDelPrimerVoto),
		votoDeLaDescripcion(t, "si", fraseDelSegundoVoto, motivoDelSegundoVoto),
		votoDeLaDescripcion(t, "si", fraseDelTercerVoto, motivoDelTercerVoto),
	}}

	var jueces []*Juez

	job.ejecucion.VotanteDeLasDescripciones = func(juez *Juez) Votante {
		jueces = append(jueces, juez)
		sucesos = append(sucesos, "votante")

		return func(mensaje string) ([]byte, error) {
			sucesos = append(sucesos, "voto")

			return votante.votar(mensaje)
		}
	}

	informe, err := ejecutarElJob(job.ejecucion)
	require.NoError(t, err)

	require.Len(t, jueces, 1, "el votante se pide una vez")
	assert.Equal(t, juezDeLasDescripcionesEn(t, job.ejecucion.Raiz), jueces[0], "con el juez de las descripciones")
	assert.Equal(t, []string{"votante", "tanda", "voto", "voto", "voto"}, sucesos,
		"el votante, antes de abrir las sesiones, y los tres votos de una descripción marcada, detrás de ellas")

	for _, mensaje := range votante.mensajes {
		assert.Equal(t, mensajeDeLaDescripcion(skillDeLasSesiones, descripcionDeLosVotos), mensaje)
	}

	leido := leerInformeEscrito(t, job.ejecucion.Destino, informe)

	assert.Equal(t, &DescripcionInformada{
		Skill: skillDeLasSesiones, Clase: claseHablaEnPrimeraOSegundaPersona, Huella: huellaDeLaDescripcionDeLaDecision(),
		Juzgada: true, Marcada: true, Frase: fraseDelPrimerVoto, Motivo: motivoDelPrimerVoto,
		Votos: []VotoDeLaDescripcion{
			votoDado("si", fraseDelPrimerVoto, motivoDelPrimerVoto),
			votoDado("si", fraseDelSegundoVoto, motivoDelSegundoVoto),
			votoDado("si", fraseDelTercerVoto, motivoDelTercerVoto),
		},
	}, leido.informe.Descripcion)
	exigirUmbrales(t, leido, []Umbral{umbralDeFueraDeLoGrabado(modeloQueDecide, ModoOrden, 0, 2), umbralDeLaDescripcionEsperado(1, 1)})
	exigirMotivosDeLaRaiz(t, leido, motivoDeLaDescripcionMarcada)
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
}

// probarElJobQueReutilizaLaDescripcion ejecuta el job con la descripción
// reutilizada en la decisión: exige que no se pida ningún votante ni ningún
// voto, que se abran las sesiones de las evals y que el informe lleve el juicio
// del bloque, con su procedencia y sin votos, y el umbral de su clase con 0 de
// 0, aprobado.
func probarElJobQueReutilizaLaDescripcion(t *testing.T) {
	t.Helper()

	modelo, version := fijadosParaElJuez(t)

	job := nuevoJobDeLaDescripcion(t, modelo, version, copiarLaEvidenciaDeLasDescripciones(t),
		descripcionReutilizadaEnLaDecision())
	job.ejecucion.VotanteDeLasDescripciones = votanteDeLasDescripcionesQueNoSePide(t)

	informe, err := ejecutarElJob(job.ejecucion)
	require.NoError(t, err)

	require.Len(t, job.tandas, 1, "las sesiones de las evals se abren")

	leido := leerInformeEscrito(t, job.ejecucion.Destino, informe)

	assert.Equal(t, &DescripcionInformada{
		Skill: skillDeLasSesiones, Clase: claseHablaEnPrimeraOSegundaPersona, Huella: huellaDeLaDescripcionDeLaDecision(),
		Procedencia: procedenciaDeMainDelJob(), Motivo: motivoDelVotoQueNo,
	}, leido.informe.Descripcion)
	exigirUmbrales(t, leido, []Umbral{umbralDeFueraDeLoGrabado(modeloQueDecide, ModoOrden, 0, 2), umbralDeLaDescripcionEsperado(0, 0)})
	exigirMotivosDeLaRaiz(t, leido)
	assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)
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
