package evals

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// rutaDeLaDefinicionDelJob es la definición del job de evals, relativa al
// directorio de este paquete, que es donde go test ejecuta los tests desde los
// que se lee (research.md V46 de H5).
const rutaDeLaDefinicionDelJob = "../../.github/workflows/evals.yml"

// Las variables del env del trabajo evals de las que sale el plan de sesiones de
// cada skill: el modelo que decide, los informativos, separados por comas, y las
// repeticiones (contrato job-de-evals §3 de H5).
const (
	variableDelModeloQueDecide       = "MODELO_DE_EVALS"
	variableDeLosModelosInformativos = "MODELOS_INFORMATIVOS_DE_EVALS"
	variableDeLasRepeticiones        = "REPETICIONES_DE_EVALS"
)

// Las variables del env del trabajo evals que fijan el juez con modelo, junto a
// la del modelo que decide (contracts/job-de-evals.md §1 de H24; FR-090 y FR-091
// de H24; ADR 0037): el id completo de su modelo y la versión de Claude Code de
// sus votos, que va aparte de la de las sesiones, VERSION_DE_CLAUDE_CODE.
const (
	variableDelModeloDelJuez   = "MODELO_DEL_JUEZ"
	variableDeLaVersionDelJuez = "VERSION_DE_CLAUDE_CODE_DEL_JUEZ"
)

// paqueteDeClaudeCode es el nombre con el que npm instala Claude Code: el paso
// de instalación del trabajo evals es el que lo nombra en su run
// (contracts/job-de-evals.md §2 de H24).
const paqueteDeClaudeCode = "@anthropic-ai/claude-code"

// Los términos que el juez con modelo añade al peor caso de un trabajo
// (contracts/job-de-evals.md §4 de H24; research.md D6 y S3 de H24). El tope de
// un voto y su margen son topeDelVoto y margenDelVoto, los del votante.
const (
	// instalacionDelClaudeDelJuez es la cota, sin medir, de lo que tarda en
	// instalarse el segundo Claude Code, el de los votos.
	instalacionDelClaudeDelJuez = 60 * time.Second

	// votosPorRespuestaComoMucho son los votos que una respuesta puede pedir
	// como mucho: los que hacen falta para marcarla, cada uno con su repetición
	// por nulo.
	votosPorRespuestaComoMucho = 2 * votosParaMarcar
)

// Los términos del peor caso del trabajo de una skill (research.md D13 de H7.3):
// dos medidos en el cierre de H7.2 y el tope de una sesión con su margen.
const (
	// fueraDeLasSesiones es lo más que tardó en el cierre de H7.2 un trabajo
	// fuera de sus sesiones: la preparación del runner, las comprobaciones
	// previas, el plan y el informe.
	fueraDeLasSesiones = 485 * time.Second

	// preparacionDeUnaSesion es lo que tardó de media en el cierre de H7.2 una
	// sesión entera con su preparación, cota de la preparación sola.
	preparacionDeUnaSesion = 22 * time.Second

	// topeDeUnaSesion es lo que dura una sesión del job como mucho antes de
	// recibir TERM, y margenDelTopeDeUnaSesion lo que se espera después antes de
	// enviar KILL (contracts/ejecucion-del-job.md §3 de H7.3).
	topeDeUnaSesion          = 240 * time.Second
	margenDelTopeDeUnaSesion = 10 * time.Second
)

// DefinicionDelJob es lo que se lee de la definición del job de evals —el
// trabajo evals, el trabajo tanda, el trabajo medida y las entradas del
// despacho— para comprobarla, para dar al sondeo la concurrencia de cada skill
// y para dar a la comprobación de la medida del juez el modelo y la versión de
// Claude Code fijados para él (data-model §6 de H7.3; research.md D15 de H7.3;
// contracts/tanda-del-job.md §4 de H7.4; data-model §6 y contracts/job-de-evals.md
// §5 de H24).
type DefinicionDelJob struct {
	// Nombre es el name del trabajo, o vacío si no lo tiene.
	Nombre string

	// Grupo es el group de su concurrency, o vacío si no lo tiene.
	Grupo string

	// CancelaLaEnCurso es el cancel-in-progress de su concurrency, o nil si no
	// lo tiene.
	CancelaLaEnCurso *bool

	// ConcurrenciaDeFlujo dice si el flujo tiene concurrency de nivel de flujo.
	ConcurrenciaDeFlujo bool

	// TopeEnMinutos es su timeout-minutes, o 0 si no lo tiene.
	TopeEnMinutos int

	// Env son las variables de su env, con su valor como texto.
	Env map[string]string

	// Skills son las de su matriz, en su orden.
	Skills []string

	// PorSkill son los ajustes que include da a cada skill.
	PorSkill map[string]AjustesDeSkill

	// ModeloQueDecide, ModelosInformativos y Repeticiones son los del plan de
	// sesiones de cada skill, leídos del env.
	ModeloQueDecide     string
	ModelosInformativos []string
	Repeticiones        int

	// ModeloDelJuez y VersionDelJuez son el id del modelo del juez y la versión
	// de Claude Code de sus votos, leídos del env; vacíos si no los fija. Son los
	// que tiene que llevar la medida versionada del juez para corresponder
	// (contracts/medida-del-juez.md §2 de H24).
	ModeloDelJuez  string
	VersionDelJuez string

	// Dependencias es su needs, o nil si no lo tiene.
	Dependencias []string

	// Condicion es su if, o vacío si no lo tiene.
	Condicion string

	// Tanda es el trabajo tanda del flujo, el que decide si la ejecución mide
	// el commit, o nil si el flujo no lo tiene (contracts/tanda-del-job.md §1 y
	// §4 de H7.4).
	Tanda *TrabajoDeLaTanda

	// OrdenDeInstalacion es el run del paso que instala Claude Code, el primero
	// de los suyos que nombra paqueteDeClaudeCode, o vacío si ninguno lo nombra
	// (contracts/job-de-evals.md §2 de H24).
	OrdenDeInstalacion string

	// Medida es el trabajo medida del flujo, el que ejecuta la medida del juez,
	// o nil si el flujo no lo tiene (contracts/job-de-evals.md §3 de H24).
	Medida *TrabajoDeLaMedida

	// EntradasDelDespacho son las de on.workflow_dispatch.inputs, por su nombre.
	EntradasDelDespacho map[string]EntradaDelDespacho
}

// TrabajoDeLaMedida es lo que se lee del trabajo medida de la definición del
// job, el que ejecuta la medida del juez de una skill fuera de la ejecución de
// sus evals (contracts/job-de-evals.md §3 de H24; data-model §6 de H24).
type TrabajoDeLaMedida struct {
	// Nombre es su name, o vacío si no lo tiene.
	Nombre string

	// Condicion es su if, o vacío si no lo tiene.
	Condicion string

	// Env son las variables de su env, con su valor como texto.
	Env map[string]string

	// Skills son las de su matriz, en su orden.
	Skills []string

	// ConcurrenciaPorSkill es la que su include da a cada skill: cuántos casos
	// se votan a la vez como mucho.
	ConcurrenciaPorSkill map[string]int

	// TopeEnMinutos es su timeout-minutes, o 0 si no lo tiene.
	TopeEnMinutos int

	// ConDependencias dice si tiene needs.
	ConDependencias bool
}

// EntradaDelDespacho es lo que se lee de una entrada de
// on.workflow_dispatch.inputs.
type EntradaDelDespacho struct {
	// PorOmision es su default, con el tipo con el que está escrito, o nil si no
	// lo tiene.
	PorOmision any `yaml:"default"`
}

// TrabajoDeLaTanda es lo que se lee del trabajo tanda de la definición del job
// para comprobarlo (contracts/tanda-del-job.md §4 de H7.4).
type TrabajoDeLaTanda struct {
	// Nombre es su name, o vacío si no lo tiene: gh da el trabajo por su name
	// y, sin él, por su id.
	Nombre string

	// Condicion es su if, o vacío si no lo tiene.
	Condicion string

	// Permisos son los de su permissions, por ámbito.
	Permisos map[string]string

	// ConConcurrencia dice si tiene concurrency.
	ConConcurrencia bool

	// Salidas son las de su outputs, con su expresión.
	Salidas map[string]string

	// TopeEnMinutos es su timeout-minutes, o 0 si no lo tiene.
	TopeEnMinutos int

	// Pasos son los suyos, en su orden.
	Pasos []PasoDelTrabajo
}

// PasoDelTrabajo es lo que se lee de un paso de un trabajo: su id, su name, su
// if y su run, cada uno vacío si no lo tiene.
type PasoDelTrabajo struct {
	ID        string `yaml:"id"`
	Nombre    string `yaml:"name"`
	Condicion string `yaml:"if"`
	Orden     string `yaml:"run"`
}

// AjustesDeSkill son lo que la entrada de include de una skill añade a su
// combinación de la matriz (FR-030 y FR-051 de H7.3).
type AjustesDeSkill struct {
	// Concurrencia es cuántas sesiones abre a la vez su trabajo como mucho.
	Concurrencia int `yaml:"concurrencia"`

	// ObjetivoDeDuracion es el de la duración de sus sesiones, en segundos; 0,
	// sin objetivo.
	ObjetivoDeDuracion int `yaml:"objetivo_de_duracion"`
}

// flujoDelJob son las claves de la definición del job que lee
// leerDefinicionDelJob; las demás no se leen.
type flujoDelJob struct {
	// Concurrencia es el concurrency de nivel de flujo: un texto o un mapa si
	// está, nil si no.
	Concurrencia any `yaml:"concurrency"`

	// Eventos son los del on del flujo, de los que solo se leen las entradas del
	// despacho.
	Eventos struct {
		Despacho struct {
			Entradas map[string]EntradaDelDespacho `yaml:"inputs"`
		} `yaml:"workflow_dispatch"`
	} `yaml:"on"`

	Trabajos struct {
		// Tanda es el trabajo de id trabajoDeLaTanda: la etiqueta no admite la
		// constante.
		Tanda  *clavesDeLaTanda        `yaml:"tanda"`
		Evals  trabajoDeEvals          `yaml:"evals"`
		Medida *clavesDelTrabajoMedida `yaml:"medida"`
	} `yaml:"jobs"`
}

// clavesDeLaTanda son las claves del trabajo tanda que lee
// leerDefinicionDelJob.
type clavesDeLaTanda struct {
	Nombre    string            `yaml:"name"`
	Condicion string            `yaml:"if"`
	Permisos  map[string]string `yaml:"permissions"`

	// Concurrencia es su concurrency: un texto o un mapa si está, nil si no.
	Concurrencia any `yaml:"concurrency"`

	Salidas       map[string]string `yaml:"outputs"`
	TopeEnMinutos int               `yaml:"timeout-minutes"`
	Pasos         []PasoDelTrabajo  `yaml:"steps"`
}

// trabajoDeEvals son las claves del trabajo evals que lee leerDefinicionDelJob.
type trabajoDeEvals struct {
	Nombre       string   `yaml:"name"`
	Dependencias []string `yaml:"needs"`
	Condicion    string   `yaml:"if"`

	Concurrencia struct {
		Grupo            string `yaml:"group"`
		CancelaLaEnCurso *bool  `yaml:"cancel-in-progress"`
	} `yaml:"concurrency"`

	Estrategia    estrategiaDeSkills `yaml:"strategy"`
	TopeEnMinutos int                `yaml:"timeout-minutes"`
	Env           map[string]string  `yaml:"env"`
	Pasos         []PasoDelTrabajo   `yaml:"steps"`
}

// clavesDelTrabajoMedida son las claves del trabajo medida que lee
// leerDefinicionDelJob.
type clavesDelTrabajoMedida struct {
	Nombre    string `yaml:"name"`
	Condicion string `yaml:"if"`

	// Dependencias es su needs: un texto o una lista si está, nil si no.
	Dependencias any `yaml:"needs"`

	Estrategia    estrategiaDeSkills `yaml:"strategy"`
	TopeEnMinutos int                `yaml:"timeout-minutes"`
	Env           map[string]string  `yaml:"env"`
}

// estrategiaDeSkills es el strategy de un trabajo con una combinación por
// skill: las skills de su matriz y lo que include añade a cada una.
type estrategiaDeSkills struct {
	Matriz struct {
		Skill   []string `yaml:"skill"`
		Include []struct {
			Skill          string `yaml:"skill"`
			AjustesDeSkill `yaml:",inline"`
		} `yaml:"include"`
	} `yaml:"matrix"`
}

// leerDefinicionDelJob lee la definición del job de la ruta con el lector común
// de documentos YAML de internal/skills —un único documento y ninguna clave
// repetida—, sin esquema: una definición de GitHub Actions tiene muchas claves
// que aquí no se miran. Del trabajo evals lee además el run de su paso de
// instalación; del flujo, el trabajo tanda, el trabajo medida y las entradas del
// despacho, si están. Las repeticiones del env tienen que ser un entero. Todo
// error nombra la ruta.
func leerDefinicionDelJob(ruta string) (DefinicionDelJob, error) {
	contenido, err := leerFichero(ruta)
	if err != nil {
		return DefinicionDelJob{}, fmt.Errorf("la definición del job no se puede leer: %w", err)
	}

	flujo, err := skills.ValidarDocumentoYAML[flujoDelJob](contenido, nil)
	if err != nil {
		return DefinicionDelJob{}, fmt.Errorf("%s: %w", ruta, err)
	}

	trabajo := flujo.Trabajos.Evals

	leida := DefinicionDelJob{
		Nombre:              trabajo.Nombre,
		Grupo:               trabajo.Concurrencia.Grupo,
		CancelaLaEnCurso:    trabajo.Concurrencia.CancelaLaEnCurso,
		ConcurrenciaDeFlujo: flujo.Concurrencia != nil,
		TopeEnMinutos:       trabajo.TopeEnMinutos,
		Env:                 trabajo.Env,
		Skills:              trabajo.Estrategia.Matriz.Skill,
		PorSkill:            map[string]AjustesDeSkill{},
		ModeloQueDecide:     trabajo.Env[variableDelModeloQueDecide],
		ModelosInformativos: separarLosModelos(trabajo.Env[variableDeLosModelosInformativos]),
		ModeloDelJuez:       trabajo.Env[variableDelModeloDelJuez],
		VersionDelJuez:      trabajo.Env[variableDeLaVersionDelJuez],
		Dependencias:        trabajo.Dependencias,
		Condicion:           trabajo.Condicion,
		OrdenDeInstalacion:  ordenDeInstalacion(trabajo.Pasos),
		Medida:              trabajoDeLaMedida(flujo.Trabajos.Medida),
		EntradasDelDespacho: flujo.Eventos.Despacho.Entradas,
	}

	if tanda := flujo.Trabajos.Tanda; tanda != nil {
		leida.Tanda = &TrabajoDeLaTanda{
			Nombre:          tanda.Nombre,
			Condicion:       tanda.Condicion,
			Permisos:        tanda.Permisos,
			ConConcurrencia: tanda.Concurrencia != nil,
			Salidas:         tanda.Salidas,
			TopeEnMinutos:   tanda.TopeEnMinutos,
			Pasos:           tanda.Pasos,
		}
	}

	for _, entrada := range trabajo.Estrategia.Matriz.Include {
		leida.PorSkill[entrada.Skill] = entrada.AjustesDeSkill
	}

	repeticiones := trabajo.Env[variableDeLasRepeticiones]

	leida.Repeticiones, err = strconv.Atoi(repeticiones)
	if err != nil {
		return DefinicionDelJob{}, fmt.Errorf("%s: jobs.evals.env.%s vale %q y no es un entero", ruta,
			variableDeLasRepeticiones, repeticiones)
	}

	return leida, nil
}

// ordenDeInstalacion es el run del primero de los pasos que nombra
// paqueteDeClaudeCode, o vacío si ninguno lo nombra.
func ordenDeInstalacion(pasos []PasoDelTrabajo) string {
	for _, paso := range pasos {
		if strings.Contains(paso.Orden, paqueteDeClaudeCode) {
			return paso.Orden
		}
	}

	return ""
}

// trabajoDeLaMedida es lo leído del trabajo medida, o nil si el flujo no lo
// tiene.
func trabajoDeLaMedida(claves *clavesDelTrabajoMedida) *TrabajoDeLaMedida {
	if claves == nil {
		return nil
	}

	medida := &TrabajoDeLaMedida{
		Nombre:               claves.Nombre,
		Condicion:            claves.Condicion,
		Env:                  claves.Env,
		Skills:               claves.Estrategia.Matriz.Skill,
		ConcurrenciaPorSkill: map[string]int{},
		TopeEnMinutos:        claves.TopeEnMinutos,
		ConDependencias:      claves.Dependencias != nil,
	}

	for _, entrada := range claves.Estrategia.Matriz.Include {
		medida.ConcurrenciaPorSkill[entrada.Skill] = entrada.Concurrencia
	}

	return medida
}

// separarLosModelos separa por comas la lista de modelos informativos del env;
// la vacía no es ningún modelo, y no uno con el nombre vacío.
func separarLosModelos(lista string) []string {
	if lista == "" {
		return nil
	}

	return strings.Split(lista, ",")
}

// peorCasoDelTrabajo es lo más que puede tardar el trabajo de una skill:
// fueraDeLasSesiones más, por cada grupo de sesiones que el trabajo reparte por
// separado, una tanda de Concurrencia sesiones por cada ⌈sesiones del grupo /
// Concurrencia⌉, cada una con su preparación, su tope y el margen del tope
// (research.md D13 de H7.3; contracts/evals-en-dos-modos.md §6 de H21;
// research.md D20 de H21). Con una skill con juez, más lo que el juez añade
// (contracts/job-de-evals.md §4 de H24).
type peorCasoDelTrabajo struct {
	// SesionesPorGrupo son las sesiones de cada grupo del plan de la skill, en
	// el orden en que se abren: las del modo orden con la prueba de red, las del
	// modo herramienta y las de las evals sin binario ni servidor.
	SesionesPorGrupo []int

	// Concurrencia es la de la skill, al menos 1.
	Concurrencia int

	// Juez es lo que el juez de la skill añade al peor caso, o nil si la skill
	// no tiene juez: entonces el peor caso es el de sus sesiones.
	Juez *peorCasoDelJuez
}

// peorCasoDelJuez es lo más que el juez con modelo puede añadir a un trabajo
// (contracts/job-de-evals.md §4 de H24; research.md D6 y S3 de H24):
// instalacionDelClaudeDelJuez más, por cada grupo de respuestas que se vota por
// separado, una tanda de Concurrencia votos por cada ⌈respuestas del grupo ×
// votosPorRespuestaComoMucho / Concurrencia⌉, cada una con el tope de un voto y
// su margen.
type peorCasoDelJuez struct {
	// RespuestasPorGrupo son las respuestas que el juez juzga de cada grupo, en
	// el orden en que se votan.
	RespuestasPorGrupo []int

	// Concurrencia es cuántas respuestas se votan a la vez como mucho, al menos
	// 1.
	Concurrencia int
}

// tandas son las veces que el juez pide sus Concurrencia votos: la suma, por
// grupo, de ⌈respuestas del grupo × votosPorRespuestaComoMucho / Concurrencia⌉.
// Un grupo no llena la última tanda con los votos del siguiente.
func (p peorCasoDelJuez) tandas() int {
	tandas := 0
	for _, respuestas := range p.RespuestasPorGrupo {
		tandas += (respuestas*votosPorRespuestaComoMucho + p.Concurrencia - 1) / p.Concurrencia
	}

	return tandas
}

// Duracion es lo que el juez añade en el peor caso.
func (p peorCasoDelJuez) Duracion() time.Duration {
	return instalacionDelClaudeDelJuez + time.Duration(p.tandas())*(topeDelVoto+margenDelVoto)
}

// String presenta en segundos los términos de lo que el juez añade, como
// «60 s + (⌈54 × 6 / 4⌉ + ⌈54 × 6 / 4⌉ + ⌈3 × 6 / 4⌉) × (35 s + 5 s)».
func (p peorCasoDelJuez) String() string {
	porGrupo := make([]string, 0, len(p.RespuestasPorGrupo))
	for _, respuestas := range p.RespuestasPorGrupo {
		porGrupo = append(porGrupo, fmt.Sprintf("⌈%d × %d / %d⌉", respuestas, votosPorRespuestaComoMucho, p.Concurrencia))
	}

	return fmt.Sprintf("%d s + (%s) × (%d s + %d s)", segundos(instalacionDelClaudeDelJuez),
		strings.Join(porGrupo, " + "), segundos(topeDelVoto), segundos(margenDelVoto))
}

// peorCasoDeLaMedida es lo más que puede tardar el trabajo de la medida del
// juez de una skill (contracts/job-de-evals.md §4 de H24): fueraDeLasSesiones,
// que no abre ninguna, más lo que el juez añade con sus casos etiquetados, que
// vota en un solo grupo.
type peorCasoDeLaMedida struct {
	// Casos son los casos etiquetados del juez de la skill.
	Casos int

	// Concurrencia es la de la skill en el trabajo de la medida, al menos 1.
	Concurrencia int
}

// delJuez es lo que el juez añade con los casos de la medida.
func (p peorCasoDeLaMedida) delJuez() peorCasoDelJuez {
	return peorCasoDelJuez{RespuestasPorGrupo: []int{p.Casos}, Concurrencia: p.Concurrencia}
}

// Duracion es el peor caso.
func (p peorCasoDeLaMedida) Duracion() time.Duration {
	return fueraDeLasSesiones + p.delJuez().Duracion()
}

// String presenta el peor caso en segundos con sus términos, como
// «16105 s = 485 s + 60 s + (⌈259 × 6 / 4⌉) × (35 s + 5 s)».
func (p peorCasoDeLaMedida) String() string {
	return fmt.Sprintf("%d s = %d s + %s", segundos(p.Duracion()), segundos(fueraDeLasSesiones), p.delJuez())
}

// tandas son las veces que el trabajo abre sus Concurrencia sesiones: la suma,
// por grupo, de ⌈sesiones del grupo / Concurrencia⌉. Un grupo no llena la
// última tanda con sesiones del siguiente.
func (p peorCasoDelTrabajo) tandas() int {
	tandas := 0
	for _, sesiones := range p.SesionesPorGrupo {
		tandas += (sesiones + p.Concurrencia - 1) / p.Concurrencia
	}

	return tandas
}

// Duracion es el peor caso.
func (p peorCasoDelTrabajo) Duracion() time.Duration {
	duracion := fueraDeLasSesiones +
		time.Duration(p.tandas())*(preparacionDeUnaSesion+topeDeUnaSesion+margenDelTopeDeUnaSesion)

	if p.Juez != nil {
		duracion += p.Juez.Duracion()
	}

	return duracion
}

// String presenta el peor caso en segundos con sus términos, como
// «14357 s = 485 s + (⌈97 / 4⌉ + ⌈96 / 4⌉ + ⌈6 / 4⌉) × (22 s + 240 s + 10 s)», y,
// con una skill con juez, con los del juez detrás.
func (p peorCasoDelTrabajo) String() string {
	porGrupo := make([]string, 0, len(p.SesionesPorGrupo))
	for _, sesiones := range p.SesionesPorGrupo {
		porGrupo = append(porGrupo, fmt.Sprintf("⌈%d / %d⌉", sesiones, p.Concurrencia))
	}

	texto := fmt.Sprintf("%d s = %d s + (%s) × (%d s + %d s + %d s)", segundos(p.Duracion()),
		segundos(fueraDeLasSesiones), strings.Join(porGrupo, " + "), segundos(preparacionDeUnaSesion),
		segundos(topeDeUnaSesion), segundos(margenDelTopeDeUnaSesion))

	if p.Juez != nil {
		texto += " + " + p.Juez.String()
	}

	return texto
}

// segundos son los segundos enteros de una duración.
func segundos(duracion time.Duration) int {
	return int(duracion / time.Second)
}

// peorCaso es el peor caso del trabajo de la skill (research.md D13 de H7.3;
// contracts/evals-en-dos-modos.md §6 de H21): sus sesiones son las del
// PlanDeEvals de las evals bien formadas de su carpeta dentro de evals, con los
// modelos y las repeticiones del env, los dos modos y la prueba de red, que se
// cuenta aunque el trabajo no la lleve, agrupadas como el trabajo las reparte;
// su concurrencia, la que le da include. Es un error una concurrencia menor que
// 1, una carpeta que no se puede leer o un plan que no se puede componer.
//
// Si la skill tiene juez, la carpeta juez de sus evals, lleva además lo que el
// juez añade (contracts/job-de-evals.md §4 de H24): sus respuestas son las que
// el informe le da a juzgar (respuestasQueSeJuzgan) cuando todas las sesiones
// terminan y se miden, y se votan con la concurrencia de la skill.
func (d DefinicionDelJob) peorCaso(evals, skill string) (peorCasoDelTrabajo, error) {
	concurrencia, err := concurrenciaDeLaSkill(d.PorSkill[skill].Concurrencia, skill)
	if err != nil {
		return peorCasoDelTrabajo{}, err
	}

	conjunto, err := LeerConjunto(filepath.Join(evals, skill))
	if err != nil {
		return peorCasoDelTrabajo{}, err
	}

	plan := PlanDeEvals{
		Evals:               conjunto.Evals,
		ModeloQueDecide:     d.ModeloQueDecide,
		ModelosInformativos: d.ModelosInformativos,
		Repeticiones:        d.Repeticiones,
		Modos:               []Modo{ModoOrden, ModoHerramienta},
		PruebaDeRed:         true,
	}
	if err := plan.Comprobar(); err != nil {
		return peorCasoDelTrabajo{}, err
	}

	sesiones := plan.Sesiones()

	peor := peorCasoDelTrabajo{
		SesionesPorGrupo: contarPorGrupo(sesiones, func(SesionPlanificada) bool { return true }),
		Concurrencia:     concurrencia,
	}

	if conjunto.Juez != nil {
		peor.Juez = &peorCasoDelJuez{
			RespuestasPorGrupo: contarPorGrupo(sesiones, func(sesion SesionPlanificada) bool { return seJuzga(plan, sesion) }),
			Concurrencia:       concurrencia,
		}
	}

	return peor, nil
}

// peorCaso es el peor caso del trabajo de la medida del juez de la skill
// (contracts/job-de-evals.md §4 de H24): sus casos son los casos etiquetados de
// la carpeta juez de sus evals, dentro de evals, y su concurrencia, la que le da
// el include del trabajo. Es un error una concurrencia menor que 1, una carpeta
// que no se puede leer, una skill sin juez o unos casos que no tienen su forma.
func (m TrabajoDeLaMedida) peorCaso(evals, skill string) (peorCasoDeLaMedida, error) {
	concurrencia, err := concurrenciaDeLaSkill(m.ConcurrenciaPorSkill[skill], skill)
	if err != nil {
		return peorCasoDeLaMedida{}, err
	}

	conjunto, err := LeerConjunto(filepath.Join(evals, skill))
	if err != nil {
		return peorCasoDeLaMedida{}, err
	}

	if conjunto.Juez == nil {
		return peorCasoDeLaMedida{}, fmt.Errorf("%s no tiene juez que medir", skill)
	}

	etiquetados, err := leerCasosEtiquetados([]byte(conjunto.Juez.Casos))
	if err != nil {
		return peorCasoDeLaMedida{}, fmt.Errorf("los casos etiquetados del juez %s: %w",
			filepath.Join(evals, skill, carpetaDelJuez, ficheroDeCasosDelJuez), err)
	}

	return peorCasoDeLaMedida{Casos: len(etiquetados.Casos), Concurrencia: concurrencia}, nil
}

// concurrenciaDeLaSkill es la concurrencia que el include de un trabajo da a la
// skill, que tiene que ser al menos 1: con menos no hay tandas que contar.
func concurrenciaDeLaSkill(concurrencia int, skill string) (int, error) {
	if concurrencia < 1 {
		return 0, fmt.Errorf("la concurrencia de %s es %d y tiene que ser un entero mayor o igual que 1",
			skill, concurrencia)
	}

	return concurrencia, nil
}

// seJuzga dice si la respuesta de la sesión del plan es de las que el informe
// da a juzgar al juez (respuestasQueSeJuzgan): las de las series que pide el
// plan con el modelo que decide cuya eval espera que la skill se active. La de
// la prueba de red no es de ninguna serie.
func seJuzga(plan PlanDeEvals, sesion SesionPlanificada) bool {
	if sesion.PruebaDeRed || sesion.Modelo != plan.ModeloQueDecide {
		return false
	}

	return slices.ContainsFunc(plan.Evals, func(eval Eval) bool { return eval.Fichero == sesion.Fichero && eval.Activa })
}

// contarPorGrupo es cuántas sesiones de cada grupo de las sesiones de un plan
// cuentan, en su orden: el plan las da agrupadas por modo, así que un grupo son
// las sesiones seguidas del mismo modo. Un grupo sin ninguna que cuente sale
// con 0.
func contarPorGrupo(sesiones []SesionPlanificada, cuenta func(SesionPlanificada) bool) []int {
	var porGrupo []int

	for posicion, sesion := range sesiones {
		if posicion == 0 || sesion.Modo != sesiones[posicion-1].Modo {
			porGrupo = append(porGrupo, 0)
		}

		if cuenta(sesion) {
			porGrupo[len(porGrupo)-1]++
		}
	}

	return porGrupo
}
