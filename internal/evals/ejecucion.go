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

	// Raiz es la raíz del repositorio evaluado: con el bloque descripcion en la
	// decisión, de ella se leen el juez de las descripciones, de la evidencia
	// del ADR 0046, y, si se juzga, la descripción de skills/<skill>/SKILL.md
	// (contracts/juez-de-las-descripciones.md §3 de H31).
	Raiz string

	// VotanteDeLasDescripciones da quien vota la descripción con el juez de las
	// descripciones que recibe: solo se le llama con la descripción juzgada en
	// la decisión y su medida que corresponde y se cumple, una vez y antes de
	// abrir las sesiones, y sus votos se piden detrás de ellas. Se recibe como
	// función para que un test de make ci vea que no se llama con el
	// instrumento sin medir, y para que el punto de entrada, que es quien abre
	// sesiones con modelo, lo prepare con el juez que el job lee (FR-026,
	// FR-033 de H31).
	VotanteDeLasDescripciones func(juez *Juez) Votante

	// decision es la del trabajo, la del fichero de -decision: qué evals del
	// plan se miden y, de las que reutiliza, su resultado en la medida de main
	// de la que viene (contracts/job-de-evals.md §3 de H30; FR-012, FR-014).
	// nil las mide todas. Desde H31, con el bloque descripcion, también si se
	// juzga la descripción de la skill o se reutiliza su juicio
	// (contracts/juez-de-las-descripciones.md §2 de H31; FR-030).
	decision *origenDeLasEvals
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
//  3. con una skill sin juez no comprueba la medida de ningún juez de las
//     respuestas ni necesita votante.
//
// Desde H30, con la decisión del trabajo abre solo las sesiones de las evals
// que mide (PlanDeEvals.deLasMedidas) y el informe compone con ellas las que
// reutiliza (FR-012, FR-014 y FR-022 de H30); si las reutiliza todas, no abre
// ninguna.
//
// Desde H31 (contracts/juez-de-las-descripciones.md §3 de H31; research D7 y
// D8 de H31; FR-026, FR-031, FR-032), con el bloque descripcion en la decisión,
// juzgada o reutilizada:
//
//  1. antes de abrir ninguna sesión, de las evals o del juez, comprueba también
//     la medida del juez de las descripciones de Raiz, con el mismo modelo y la
//     misma versión; cada una de sus líneas, precedida de «juez de las
//     descripciones: », va al informe del instrumento sin medir, detrás de las
//     del juez de la skill, si las hay, y no abre nada;
//  2. con la juzgada, antes de las sesiones, lee su descripción y pide el
//     votante a VotanteDeLasDescripciones;
//  3. tras las sesiones de las evals, con la juzgada, vota su descripción; con
//     la reutilizada, toma el juicio del bloque, sin pedir ningún voto;
//  4. y el informe la lleva, con el umbral de su clase.
//
// El error es el de un plan sin sentido, unas evals que no se pueden leer o una
// descripción que no se puede juzgar —el juez de las descripciones o el
// SKILL.md de su skill que no se pueden leer, la juzgada sin quien dé su
// votante o la reutilizada sin su juicio—, antes de abrir ninguna sesión; el de
// quien abre las sesiones, sin escribir el informe; o el de EscribirInforme.
// Con el informe escrito no hay error, sea cual sea su veredicto: lo mira quien
// la llama.
func ejecutarElJob(e EjecucionDelJob) (Informe, error) {
	if err := e.Plan.Comprobar(); err != nil {
		return Informe{}, fmt.Errorf("el job no se puede ejecutar: %w", err)
	}

	conjunto, err := LeerConjunto(e.Evals)
	if err != nil {
		return Informe{}, fmt.Errorf("el job no se puede ejecutar: %w", err)
	}

	descripcion, err := e.descripcionDelTrabajo()
	if err != nil {
		return Informe{}, fmt.Errorf("el job no se puede ejecutar: %w", err)
	}

	if sinMedir := e.sinMedir(conjunto.Juez, descripcion); len(sinMedir) > 0 {
		return EscribirInforme(e.informe(ejecucionPorTandas{}, sinMedir, conjunto.Juez == nil, nil))
	}

	if err := descripcion.prepararLaVotacion(e.Raiz, e.VotanteDeLasDescripciones); err != nil {
		return Informe{}, fmt.Errorf("el job no se puede ejecutar: %w", err)
	}

	ejecucion, err := ejecutarPorTandas(e.Plan.deLasMedidas(e.decision).Sesiones(), e.AbrirLaTanda)
	if err != nil {
		return Informe{}, err
	}

	return EscribirInforme(e.informe(ejecucion, nil, conjunto.Juez == nil, descripcion.informada()))
}

// descripcionDelTrabajo es la descripción de la skill que el trabajo juzga o
// cuyo juicio reutiliza, con el juez de las descripciones leído de Raiz, si la
// decisión lleva el bloque descripcion; nil si no lo lleva. El error es el del
// juez que no se puede leer.
func (e EjecucionDelJob) descripcionDelTrabajo() (*descripcionDelTrabajo, error) {
	if e.decision == nil || e.decision.Descripcion == nil {
		return nil, nil
	}

	return leerElJuezDeLaDescripcion(e.Raiz, *e.decision.Descripcion)
}

// sinMedir son las líneas de la comprobación de las medidas versionadas con el
// modelo y la versión fijados para el juez: las del juez de la skill, si lo
// tiene, y detrás las del juez de las descripciones, si el trabajo publica una
// descripción, cada una con su prefijo. Ninguna si las que hay corresponden y
// se cumplen.
func (e EjecucionDelJob) sinMedir(juez *Juez, descripcion *descripcionDelTrabajo) []string {
	var lineas []string
	if juez != nil {
		lineas = comprobarLaMedida(juez, e.ModeloDelJuez, e.VersionDelJuez)
	}

	return append(lineas, descripcion.sinMedir(e.ModeloDelJuez, e.VersionDelJuez)...)
}

// informe es lo que EscribirInforme recibe de la ejecución del job: lo del job,
// lo que dejó la apertura de las sesiones por tandas, si la medida de un juez no
// corresponde o no se cumple, las líneas que lo dicen, si la skill no tiene
// juez y la descripción de la skill, si el trabajo la publica.
func (e EjecucionDelJob) informe(
	ejecucion ejecucionPorTandas, sinMedir []string, sinJuez bool, descripcion *DescripcionInformada,
) InformeAEscribir {
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
		decision:              e.decision,
		descripcion:           descripcion,
		sinJuez:               sinJuez,
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
