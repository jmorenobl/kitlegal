package evals

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/jmorenobl/kitlegal/internal/cache"
)

// Lo que el repartidor crea en el directorio de cada sesión, además de cache/ y
// traza/ (contracts/ejecucion-del-job.md §3.2 de H7.3; data-model §7 de H7.3).
const (
	// directorioDeTrabajo es el directorio vacío, fuera del repositorio, en el
	// que se ejecuta el guion de la sesión.
	directorioDeTrabajo = "trabajo"

	// directorioTemporal es el temporal propio de la sesión, TMPDIR y
	// CLAUDE_CODE_TMPDIR (research.md V4 y D10 de H7.3).
	directorioTemporal = "tmp"

	// directorioDeClaude es el directorio de configuración de Claude Code de la
	// sesión, CLAUDE_CONFIG_DIR, donde escribe todo su estado (research.md V4, V5
	// y D10 de H7.3).
	directorioDeClaude = "claude"

	// directorioDeSkills es, dentro de directorioDeClaude, el de las skills de
	// usuario, que Claude Code carga con --setting-sources user.
	directorioDeSkills = "skills"
)

// permisosDeLaSesion son los de cada directorio que el repartidor crea en el de
// la sesión: solo para quien la abre, como pide Claude Code para su temporal
// (research.md V4 de H7.3).
const permisosDeLaSesion = 0o700

// Valores del entorno de la sesión (contracts/ejecucion-del-job.md §4 de H7.3):
// el proxy apunta a un puerto cerrado, así que toda petición que lo usa se
// rechaza, salvo la del modelo, que va a destinoSinProxy (contrato job-de-evals
// §3.2 de H5).
const (
	proxyQueRechaza = "http://127.0.0.1:9"
	destinoSinProxy = "api.anthropic.com"
)

// variableDeLaTraza es la variable con la que el guion de la sesión ejecuta la
// orden bajo strace, con el valor valorDeLaTraza (contracts/ejecucion-del-job.md
// §2 de H7.3).
const (
	variableDeLaTraza = "KITLEGAL_EVALS_TRAZA"
	valorDeLaTraza    = "si"
)

// codigoDeSenal es lo que la shell suma al número de la señal que termina un
// proceso para dar su código: 137 es 128 más KILL, como lo da timeout.
const codigoDeSenal = 128

// SesionesAEjecutar es lo que el repartidor necesita para abrir las sesiones de
// un plan: el job y el sondeo lo llaman con lo suyo (data-model §7 de H7.3;
// contracts/ejecucion-del-job.md §3 de H7.3).
type SesionesAEjecutar struct {
	// Plan son las sesiones que se abren, en su orden.
	Plan []SesionPlanificada

	// Concurrencia es cuántas sesiones se abren a la vez como mucho.
	Concurrencia int

	// Evals es el directorio de las evals de la skill: cada sesión se prepara
	// con las consultas de todas.
	Evals string

	// Sesiones es el directorio en el que se crea el de cada sesión, con su
	// nombre.
	Sesiones string

	// Skills es el directorio de las skills instaladas, el que deja make
	// install: cada sesión ve un enlace a cada una de sus entradas.
	Skills string

	// Guion es la ruta de scripts/evals-sesion.sh, que abre la sesión.
	Guion string

	// Entorno es la base del entorno de cada sesión: el del job, entero, o la
	// lista del sondeo. Lo de contracts/ejecucion-del-job.md §4 de H7.3 va
	// encima.
	Entorno []string

	// Traza dice si la sesión se abre bajo strace, con la traza en su traza/.
	Traza bool

	// Tope es lo que dura una sesión como mucho antes de recibir TERM: 240 s.
	Tope time.Duration

	// MargenDelTope es lo que se espera tras TERM antes de enviar KILL: 10 s.
	MargenDelTope time.Duration
}

// EjecucionDeSesiones es lo que el repartidor devuelve de las sesiones de un
// plan (data-model §7 de H7.3).
type EjecucionDeSesiones struct {
	// Abiertas son los nombres de las sesiones que se abrieron, en el orden en
	// que se abrieron.
	Abiertas []string

	// SinAbrir son las sesiones del plan que no se abrieron tras una sesión con
	// el mensaje del límite de uso (FR-044 de H7.3), en su orden.
	SinAbrir []SesionPlanificada

	// Duracion es el tiempo desde antes de preparar la primera sesión hasta que
	// termina la última (research.md D14 de H7.3).
	Duracion time.Duration
}

// abrirSesion abre una sesión del plan (contracts/ejecucion-del-job.md §3.2 y
// §3.3 de H7.3; research.md D8, D9 y D10 de H7.3):
//
//  1. crea su directorio, con su nombre dentro de Sesiones, y en él trabajo/,
//     cache/, traza/, tmp/ y claude/skills/, en 0700, con un enlace simbólico
//     absoluto a cada entrada de Skills;
//  2. la prepara justo antes de abrirla con PrepararSesion, en proceso: las
//     evals de la skill, UnionDeGrabaciones(), su eval, su modelo y la prueba de
//     red. Una falta de lo grabado es un error, como el de preparación;
//  3. ejecuta el guion en trabajo/, en su propio grupo de procesos, con el
//     entorno de la sesión y su salida estándar y su salida de error en
//     sesion.jsonl y sesion.err; a los Tope envía TERM al grupo y, pasado
//     MargenDelTope, KILL;
//  4. escribe codigo-de-la-sesion: 124 si bastó TERM, 137 si hizo falta KILL, o
//     el del proceso si terminó antes.
//
// No recibe un contexto, sino interrupcion, el Done() del contexto que el job y
// el sondeo cancelan con SIGINT y SIGTERM (research.md D9 de H7.3): la
// preparación invoca el binario en proceso con app.Main, que abre el contexto de
// cada invocación y no admite el de quien llama, así que un contexto aquí sería
// una cancelación a medias, y quien lo tuviera no pasaría contextcheck sin un
// atajo (research.md D14 y V59 de H5). Lo que sí se interrumpe es la sesión: si
// interrupcion ya está cerrado al terminar la preparación, no se abre, y si se
// cierra con la sesión abierta, se cierra con la misma secuencia de TERM y KILL
// al grupo. En los dos casos el error es errSesionInterrumpida, no se escribe
// ningún código y la sesión no se reintenta (FR-037 de H7.3).
//
// El error nombra la sesión y queda, además, para lo que impide abrirla o
// escribir su código.
func abrirSesion(interrupcion <-chan struct{}, e SesionesAEjecutar, sesion SesionPlanificada) error {
	dir, err := filepath.Abs(filepath.Join(e.Sesiones, sesion.Nombre))
	if err != nil {
		return fmt.Errorf("la sesión %s: su directorio no tiene ruta absoluta: %w", sesion.Nombre, err)
	}

	if err := crearDirectorioDeSesion(dir, e.Skills); err != nil {
		return fmt.Errorf("la sesión %s: %w", sesion.Nombre, err)
	}

	faltas, err := PrepararSesion(SesionAPreparar{
		Evals:       e.Evals,
		Grabaciones: UnionDeGrabaciones(),
		Fichero:     sesion.Fichero,
		Modelo:      sesion.Modelo,
		Directorio:  dir,
		PruebaDeRed: sesion.PruebaDeRed,
	})
	if err != nil {
		return fmt.Errorf("la sesión %s no se pudo preparar: %w", sesion.Nombre, err)
	}

	if len(faltas) > 0 {
		textos := make([]string, 0, len(faltas))
		for _, falta := range faltas {
			textos = append(textos, falta.String())
		}

		return fmt.Errorf("la sesión %s no se pudo preparar: lo grabado no sirve estas consultas:\n%s",
			sesion.Nombre, strings.Join(textos, "\n"))
	}

	codigo, err := ejecutarElGuion(interrupcion, e.Guion, sesionEnMarcha{
		dir:     dir,
		entorno: entornoDeLaSesion(e.Entorno, dir, e.Traza),
		tope:    e.Tope,
		margen:  e.MargenDelTope,
	})
	if err != nil {
		return fmt.Errorf("la sesión %s: %w", sesion.Nombre, err)
	}

	if err := escribirFichero(filepath.Join(dir, ficheroDelCodigo), []byte(strconv.Itoa(codigo)+"\n")); err != nil {
		return fmt.Errorf("la sesión %s: %w", sesion.Nombre, err)
	}

	return nil
}

// crearDirectorioDeSesion crea el directorio de la sesión con trabajo/, cache/,
// traza/, tmp/ y claude/skills/, y en este un enlace simbólico a cada entrada
// de las skills instaladas. Si alguno ya existe, es un error: ninguna sesión
// escribe donde escribe otra (FR-032 de H7.3).
func crearDirectorioDeSesion(dir, skills string) error {
	if err := os.MkdirAll(dir, permisosDeLaSesion); err != nil {
		return fmt.Errorf("el directorio %s no se puede crear: %w", dir, err)
	}

	subdirectorios := []string{
		directorioDeTrabajo, directorioDeLaCache, directorioDeLaTraza, directorioTemporal, directorioDeClaude,
		filepath.Join(directorioDeClaude, directorioDeSkills),
	}
	for _, subdirectorio := range subdirectorios {
		ruta := filepath.Join(dir, subdirectorio)
		if err := os.Mkdir(ruta, permisosDeLaSesion); err != nil {
			return fmt.Errorf("el directorio %s no se puede crear: %w", ruta, err)
		}
	}

	return enlazarLasSkills(filepath.Join(dir, directorioDeClaude, directorioDeSkills), skills)
}

// enlazarLasSkills crea en destino un enlace simbólico absoluto a cada entrada
// del directorio de skills instaladas, con su nombre: la sesión ve la skill tal
// como la deja make install, sin copiarla (research.md D10 de H7.3).
func enlazarLasSkills(destino, skills string) error {
	origen, err := filepath.Abs(skills)
	if err != nil {
		return fmt.Errorf("las skills instaladas %s no tienen ruta absoluta: %w", skills, err)
	}

	entradas, err := os.ReadDir(origen)
	if err != nil {
		return fmt.Errorf("las skills instaladas %s no se pueden leer: %w", origen, err)
	}

	for _, entrada := range entradas {
		if err := os.Symlink(filepath.Join(origen, entrada.Name()), filepath.Join(destino, entrada.Name())); err != nil {
			return fmt.Errorf("la skill %s no se puede enlazar: %w", entrada.Name(), err)
		}
	}

	return nil
}

// entornoDeLaSesion es el entorno con el que se ejecuta el guion de la sesión
// del directorio dir: la base, sin las variables de
// contracts/ejecucion-del-job.md §4 de H7.3, y detrás esas variables con los
// valores de la sesión. KITLEGAL_EVALS_TRAZA solo va con traza; sin ella, se
// quita de la base.
func entornoDeLaSesion(base []string, dir string, traza bool) []string {
	temporal := filepath.Join(dir, directorioTemporal)

	variables := []string{
		cache.VariableDirectorio + "=" + filepath.Join(dir, directorioDeLaCache),
		"HTTP_PROXY=" + proxyQueRechaza,
		"HTTPS_PROXY=" + proxyQueRechaza,
		"http_proxy=" + proxyQueRechaza,
		"https_proxy=" + proxyQueRechaza,
		"NO_PROXY=" + destinoSinProxy,
		"no_proxy=" + destinoSinProxy,
		"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1",
		"CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0",
		"CLAUDE_CONFIG_DIR=" + filepath.Join(dir, directorioDeClaude),
		"TMPDIR=" + temporal,
		"CLAUDE_CODE_TMPDIR=" + temporal,
	}
	if traza {
		variables = append(variables, variableDeLaTraza+"="+valorDeLaTraza)
	}

	return sobreLaBase(base, variables, variableDeLaTraza)
}

// sobreLaBase es la base sin ninguna variable con el nombre de una de las
// variables dadas, nombre=valor, ni de los nombres quitados, y detrás las
// variables dadas, en su orden.
func sobreLaBase(base, variables []string, quitados ...string) []string {
	nombres := slices.Clone(quitados)
	for _, variable := range variables {
		nombre, _, _ := strings.Cut(variable, "=")
		nombres = append(nombres, nombre)
	}

	entorno := make([]string, 0, len(base)+len(variables))
	for _, variable := range base {
		if nombre, _, _ := strings.Cut(variable, "="); !slices.Contains(nombres, nombre) {
			entorno = append(entorno, variable)
		}
	}

	return append(entorno, variables...)
}

// sesionEnMarcha es lo que ejecutarElGuion necesita de una sesión preparada,
// además del guion: su directorio, su entorno, su tope y el margen tras él.
type sesionEnMarcha struct {
	dir     string
	entorno []string
	tope    time.Duration
	margen  time.Duration
}

// errSesionInterrumpida es el error de la sesión que no se abre, o que se
// cierra antes de terminar, porque se ha interrumpido el job o el sondeo.
var errSesionInterrumpida = errors.New("se ha interrumpido antes de terminar")

// ejecutarElGuion ejecuta el guion de la sesión, sin argumentos y con la
// salida estándar y la de error en sesion.jsonl y sesion.err, y devuelve el
// código que va en codigo-de-la-sesion. El guion llega como parámetro para que
// la orden no lleve nada que no sea constante (research.md D8 y V8 de H7.3). Con
// interrupcion ya cerrado, no lo ejecuta.
func ejecutarElGuion(interrupcion <-chan struct{}, guion string, s sesionEnMarcha) (_ int, err error) {
	select {
	case <-interrupcion:
		return 0, errSesionInterrumpida
	default:
	}

	transcript, err := crearFicheroDeSesion(filepath.Join(s.dir, ficheroDelTranscript))
	if err != nil {
		return 0, err
	}

	defer func() { err = errors.Join(err, cerrarFicheroDeSesion(transcript)) }()

	salidaDeError, err := crearFicheroDeSesion(filepath.Join(s.dir, ficheroDeSalidaDeError))
	if err != nil {
		return 0, err
	}

	defer func() { err = errors.Join(err, cerrarFicheroDeSesion(salidaDeError)) }()

	// El contexto de la orden es el del tope: al agotarse, exec llama a Cancel.
	// Solo lo termina el tope; la interrupción llega por su canal.
	contextoDelTope, cancelar := context.WithTimeout(context.Background(), s.tope)
	defer cancelar()

	orden := exec.CommandContext(contextoDelTope, guion)
	orden.Dir = filepath.Join(s.dir, directorioDeTrabajo)
	orden.Env = s.entorno
	orden.Stdout = transcript
	orden.Stderr = salidaDeError

	return esperarConElTope(contextoDelTope, interrupcion, orden, s.margen)
}

// esperarConElTope arranca la orden en su propio grupo de procesos y espera a
// que termine, con el tope de timeout --kill-after=10s 240s pero en Go, que vale
// igual en Linux y en macOS, donde no hay timeout (research.md D9 y V12 de
// H7.3): cuando termina tope, el contexto de la orden, envía TERM al grupo, que
// llega también a strace y a lo que la sesión haya lanzado, y si pasado el
// margen la orden no ha terminado, KILL. Devuelve 124 si bastó TERM, 137 si
// hizo falta KILL, y el código del proceso si terminó antes.
//
// Si antes se cierra interrupcion, cierra la sesión con la misma secuencia y
// devuelve errSesionInterrumpida.
func esperarConElTope(tope context.Context, interrupcion <-chan struct{}, orden *exec.Cmd, margen time.Duration) (
	int, error,
) {
	// exec solo llama a Cancel si el contexto termina antes que el proceso, así
	// que conTERM dice si el tope envió TERM. Lo envía al grupo, y no el KILL al
	// proceso que envía exec por omisión.
	var conTERM atomic.Bool

	orden.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	orden.Cancel = func() error {
		conTERM.Store(true)

		return senalAlGrupo(orden.Process.Pid, syscall.SIGTERM)
	}

	if err := orden.Start(); err != nil {
		return 0, fmt.Errorf("el guion %s no se puede ejecutar: %w", orden.Path, err)
	}

	terminada := make(chan error, 1)

	go func() { terminada <- orden.Wait() }()

	errDeLaEspera, conKILL, err := esperarConElMargen(tope, interrupcion, orden.Process.Pid, terminada, margen)
	if err != nil {
		return 0, err
	}

	switch {
	case conKILL:
		return codigoDeKillTrasElTope, nil
	case conTERM.Load():
		return codigoDelTope, nil
	default:
		return codigoDelProceso(errDeLaEspera)
	}
}

// esperarConElMargen espera a que la orden del grupo pid termine y devuelve lo
// que devolvió su Wait. Si antes termina el tope, cuando exec ya ha enviado TERM
// al grupo, o se cierra interrupcion, que se lo envía aquí, espera el margen y,
// si no ha terminado, envía KILL al grupo, espera a que termine y lo dice con
// conKILL. El error es errSesionInterrumpida tras la interrupción, y el de la
// señal que no se puede enviar.
func esperarConElMargen(tope context.Context, interrupcion <-chan struct{}, pid int, terminada <-chan error,
	margen time.Duration,
) (errDeLaEspera error, conKILL bool, err error) {
	select {
	case errDeLaEspera = <-terminada:
		return errDeLaEspera, false, nil
	case <-tope.Done():
	case <-interrupcion:
		if err := senalAlGrupo(pid, syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
			return nil, false, err
		}

		errDeLaEspera, conKILL, err = matarTrasElMargen(pid, terminada, margen)

		return errDeLaEspera, conKILL, errors.Join(err, errSesionInterrumpida)
	}

	return matarTrasElMargen(pid, terminada, margen)
}

// matarTrasElMargen espera el margen a que termine la orden del grupo pid, que
// ya ha recibido TERM, y, si no ha terminado, le envía KILL y espera a que
// termine. Devuelve lo que devolvió su Wait, si hizo falta KILL y el error del
// KILL que no se puede enviar.
func matarTrasElMargen(pid int, terminada <-chan error, margen time.Duration) (
	errDeLaEspera error, conKILL bool, err error,
) {
	temporizador := time.NewTimer(margen)
	defer temporizador.Stop()

	select {
	case errDeLaEspera = <-terminada:
		return errDeLaEspera, false, nil
	case <-temporizador.C:
	}

	if err := senalAlGrupo(pid, syscall.SIGKILL); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return nil, false, err
	}

	return <-terminada, true, nil
}

// senalAlGrupo envía la señal al grupo de procesos que encabeza el proceso pid.
// Un grupo sin ningún proceso es os.ErrProcessDone, que es lo que exec espera de
// Cancel cuando ya no queda nada que cancelar.
func senalAlGrupo(pid int, senal syscall.Signal) error {
	err := syscall.Kill(-pid, senal)

	switch {
	case errors.Is(err, syscall.ESRCH):
		return os.ErrProcessDone
	case err != nil:
		return fmt.Errorf("la señal %s no se puede enviar al grupo de procesos %d: %w", senal, pid, err)
	default:
		return nil
	}
}

// codigoDelProceso es el código con el que terminó el guion según lo que
// devolvió su Wait: el de salida o, si lo terminó una señal, 128 más su número,
// como lo da la shell. El error queda para una espera que no dice cómo terminó.
func codigoDelProceso(errDeLaEspera error) (int, error) {
	if errDeLaEspera == nil {
		return 0, nil
	}

	var salida *exec.ExitError
	if !errors.As(errDeLaEspera, &salida) {
		return 0, fmt.Errorf("la espera del guion ha fallado: %w", errDeLaEspera)
	}

	if estado, ok := salida.Sys().(syscall.WaitStatus); ok && estado.Signaled() {
		return codigoDeSenal + int(estado.Signal()), nil
	}

	return salida.ExitCode(), nil
}

// crearFicheroDeSesion crea, o vacía si ya existe, un fichero del directorio de
// la sesión que escribe el guion, con permisos 0o600. filepath.Clean es lo que el
// control de rutas reconoce como saneado (gosec G304).
func crearFicheroDeSesion(ruta string) (*os.File, error) {
	fichero, err := os.OpenFile(filepath.Clean(ruta), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, fmt.Errorf("el fichero %s no se puede crear: %w", ruta, err)
	}

	return fichero, nil
}

// cerrarFicheroDeSesion cierra un fichero de la sesión, nombrándolo si falla.
func cerrarFicheroDeSesion(fichero *os.File) error {
	if err := fichero.Close(); err != nil {
		return fmt.Errorf("el fichero %s no se puede cerrar: %w", fichero.Name(), err)
	}

	return nil
}
