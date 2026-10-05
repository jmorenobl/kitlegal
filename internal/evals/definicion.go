package evals

import (
	"fmt"
	"path/filepath"
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

// DefinicionDelJob es lo que se lee de la definición del job de evals, trabajo
// evals y trabajo tanda, para comprobarla, para dar al sondeo la concurrencia
// de cada skill y para dar a la comprobación de la medida del juez el modelo y
// la versión de Claude Code fijados para él (data-model §6 de H7.3; research.md
// D15 de H7.3; contracts/tanda-del-job.md §4 de H7.4; data-model §6 de H24).
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

	Trabajos struct {
		// Tanda es el trabajo de id trabajoDeLaTanda: la etiqueta no admite la
		// constante.
		Tanda *clavesDeLaTanda `yaml:"tanda"`
		Evals trabajoDeEvals   `yaml:"evals"`
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

	Estrategia struct {
		Matriz struct {
			Skill   []string `yaml:"skill"`
			Include []struct {
				Skill          string `yaml:"skill"`
				AjustesDeSkill `yaml:",inline"`
			} `yaml:"include"`
		} `yaml:"matrix"`
	} `yaml:"strategy"`

	TopeEnMinutos int               `yaml:"timeout-minutes"`
	Env           map[string]string `yaml:"env"`
}

// leerDefinicionDelJob lee la definición del job de la ruta con el lector común
// de documentos YAML de internal/skills —un único documento y ninguna clave
// repetida—, sin esquema: una definición de GitHub Actions tiene muchas claves
// que aquí no se miran. Las repeticiones del env tienen que ser un entero. Todo
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
// research.md D20 de H21).
type peorCasoDelTrabajo struct {
	// SesionesPorGrupo son las sesiones de cada grupo del plan de la skill, en
	// el orden en que se abren: las del modo orden con la prueba de red, las del
	// modo herramienta y las de las evals sin binario ni servidor.
	SesionesPorGrupo []int

	// Concurrencia es la de la skill, al menos 1.
	Concurrencia int
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
	return fueraDeLasSesiones + time.Duration(p.tandas())*(preparacionDeUnaSesion+topeDeUnaSesion+margenDelTopeDeUnaSesion)
}

// String presenta el peor caso en segundos con sus términos, como
// «14357 s = 485 s + (⌈97 / 4⌉ + ⌈96 / 4⌉ + ⌈6 / 4⌉) × (22 s + 240 s + 10 s)».
func (p peorCasoDelTrabajo) String() string {
	porGrupo := make([]string, 0, len(p.SesionesPorGrupo))
	for _, sesiones := range p.SesionesPorGrupo {
		porGrupo = append(porGrupo, fmt.Sprintf("⌈%d / %d⌉", sesiones, p.Concurrencia))
	}

	return fmt.Sprintf("%d s = %d s + (%s) × (%d s + %d s + %d s)", segundos(p.Duracion()),
		segundos(fueraDeLasSesiones), strings.Join(porGrupo, " + "), segundos(preparacionDeUnaSesion),
		segundos(topeDeUnaSesion), segundos(margenDelTopeDeUnaSesion))
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
func (d DefinicionDelJob) peorCaso(evals, skill string) (peorCasoDelTrabajo, error) {
	concurrencia := d.PorSkill[skill].Concurrencia
	if concurrencia < 1 {
		return peorCasoDelTrabajo{}, fmt.Errorf("la concurrencia de %s es %d y tiene que ser un entero mayor o igual que 1",
			skill, concurrencia)
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

	return peorCasoDelTrabajo{SesionesPorGrupo: sesionesPorGrupo(plan.Sesiones()), Concurrencia: concurrencia}, nil
}

// sesionesPorGrupo es cuántas sesiones tiene cada grupo de las sesiones de un
// plan, en su orden: el plan las da agrupadas por modo, así que un grupo son
// las sesiones seguidas del mismo modo.
func sesionesPorGrupo(sesiones []SesionPlanificada) []int {
	var porGrupo []int

	for posicion, sesion := range sesiones {
		if posicion == 0 || sesion.Modo != sesiones[posicion-1].Modo {
			porGrupo = append(porGrupo, 0)
		}

		porGrupo[len(porGrupo)-1]++
	}

	return porGrupo
}
