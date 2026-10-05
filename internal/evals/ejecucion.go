package evals

import (
	"fmt"
	"os"
	"path/filepath"
)

// EjecucionDelJob es lo que ejecutarElJob necesita para ejecutar el job de evals
// de una skill (data-model §6 de H24; contracts/job-de-evals.md §2 de H24): lo
// que su punto de entrada, TestEjecucionDelJob, reúne de sus banderas.
type EjecucionDelJob struct {
	// Skill es la skill evaluada.
	Skill string

	// Evals es el directorio de sus evals: la skill tiene juez si tiene en él la
	// carpeta del juez.
	Evals string

	// Plan es el plan de sesiones del job, con las evals bien formadas de Evals:
	// sus sesiones son las que se abren, y sus modelos, sus repeticiones y sus
	// modos, los del informe.
	Plan PlanDeEvals

	// AbrirLaTanda es quien abre las sesiones de una tanda del plan: el
	// repartidor, ejecutarSesiones, con lo del job. Se recibe como función para
	// que un test de make ci vea que no se llama con el instrumento sin medir
	// (research D11 de H24; FR-104 de H24).
	AbrirLaTanda func(tanda []SesionPlanificada) (EjecucionDeSesiones, error)

	// Sesiones es el directorio en el que quien abre las sesiones crea el de
	// cada una, y Destino, el directorio en el que se escriben informe.md e
	// informe.json.
	Sesiones string
	Destino  string

	// Umbral es cuántas sesiones de una serie tienen que pasar para que la serie
	// pase, y Commit, el commit evaluado.
	Umbral int
	Commit string

	// SinPython es la ruta de sin-python.txt.
	SinPython string

	// ObjetivoDeDuracion son los segundos que el job admite para la tanda de
	// cada modo; 0, sin objetivo.
	ObjetivoDeDuracion int

	// Votar es quien da cada voto del juez de la skill; con una skill sin juez
	// no hace falta.
	Votar Votante

	// ModeloDelJuez y VersionDelJuez son el id del modelo del juez y la versión
	// de Claude Code de sus votos, los fijados para el job: con ellos se
	// comprueba la medida versionada.
	ModeloDelJuez  string
	VersionDelJuez string

	// ConcurrenciaDelJuez es cuántas respuestas se votan a la vez como mucho: la
	// concurrencia de la skill.
	ConcurrenciaDelJuez int
}

// ejecutarElJob ejecuta el job de evals de una skill, de su plan a su informe
// (contracts/job-de-evals.md §2 y contracts/informe-del-job.md §5 de H24;
// research D11 de H24; FR-043, FR-044):
//
//  1. con una skill que tiene juez, comprueba antes de nada que su medida
//     versionada corresponde y se cumple (comprobarLaMedida), con el modelo y la
//     versión fijados. Si da alguna línea, escribe el informe del instrumento
//     sin medir, con ellas, sin llamar a quien abre las sesiones ni al votante:
//     ninguna sesión, de evals o del juez, se abre con un juez que no está
//     medido;
//  2. si no, abre las sesiones del plan tanda a tanda (ejecutarPorTandas) y
//     escribe el informe con EscribirInforme, que juzga las respuestas con el
//     votante. Ningún caso etiquetado se vota aquí: los dos umbrales de la
//     medida llevan los recuentos de la versionada;
//  3. con una skill sin juez no comprueba ninguna medida ni necesita votante.
//
// El error es el de un plan sin sentido o unas evals que no se pueden leer,
// antes de abrir ninguna sesión; el de quien abre las sesiones, sin escribir el
// informe; o el de EscribirInforme. Con el informe escrito no hay error, sea
// cual sea su veredicto: lo mira quien la llama.
func ejecutarElJob(e EjecucionDelJob) (Informe, error) {
	if err := e.Plan.Comprobar(); err != nil {
		return Informe{}, fmt.Errorf("el job no se puede ejecutar: %w", err)
	}

	conjunto, err := LeerConjunto(e.Evals)
	if err != nil {
		return Informe{}, fmt.Errorf("el job no se puede ejecutar: %w", err)
	}

	if conjunto.Juez != nil {
		if sinMedir := comprobarLaMedida(conjunto.Juez, e.ModeloDelJuez, e.VersionDelJuez); len(sinMedir) > 0 {
			return EscribirInforme(e.informe(ejecucionPorTandas{}, sinMedir))
		}
	}

	ejecucion, err := ejecutarPorTandas(e.Plan.Sesiones(), e.AbrirLaTanda)
	if err != nil {
		return Informe{}, err
	}

	return EscribirInforme(e.informe(ejecucion, nil))
}

// informe es lo que EscribirInforme recibe de la ejecución del job: lo del job,
// lo que dejó la apertura de las sesiones por tandas y, si la medida del juez no
// corresponde o no se cumple, las líneas que lo dicen.
func (e EjecucionDelJob) informe(ejecucion ejecucionPorTandas, sinMedir []string) InformeAEscribir {
	return InformeAEscribir{
		Skill:                 e.Skill,
		Evals:                 e.Evals,
		Sesiones:              e.Sesiones,
		Destino:               e.Destino,
		ModeloQueDecide:       e.Plan.ModeloQueDecide,
		ModelosInformativos:   e.Plan.ModelosInformativos,
		Repeticiones:          e.Plan.Repeticiones,
		Umbral:                e.Umbral,
		Modos:                 e.Plan.Modos,
		Commit:                e.Commit,
		SinPython:             e.SinPython,
		SinAbrir:              ejecucion.sinAbrir,
		DuracionDeLasSesiones: ejecucion.duracion,
		DuracionDeLosModos:    ejecucion.porModo,
		ObjetivoDeDuracion:    e.ObjetivoDeDuracion,
		Votar:                 e.Votar,
		ModeloDelJuez:         e.ModeloDelJuez,
		VersionDelJuez:        e.VersionDelJuez,
		ConcurrenciaDelJuez:   e.ConcurrenciaDelJuez,
		InstrumentoSinMedir:   sinMedir,
	}
}

// pathDelJuez es el PATH de los votos del juez en el job
// (contracts/job-de-evals.md §2 de H24; FR-091 de H24): el del job con el
// directorio del Claude Code del juez delante, de modo que el claude que
// encuentra el guion del voto es ese, con su versión, y no el de las sesiones.
// Sin PATH del job es solo ese directorio, sin un elemento vacío detrás, que
// sería el directorio de trabajo del voto. La ruta del claude del juez tiene que
// ser absoluta: cada voto se ejecuta en su propio directorio, y una relativa se
// buscaría desde él, sin encontrarla, hasta dar con el claude de las sesiones.
func pathDelJuez(claude, delJob string) (string, error) {
	if !filepath.IsAbs(claude) {
		return "", fmt.Errorf("el claude del juez %s no tiene ruta absoluta", claude)
	}

	directorio := filepath.Dir(claude)
	if delJob == "" {
		return directorio, nil
	}

	return directorio + string(os.PathListSeparator) + delJob, nil
}
