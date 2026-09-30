package evals

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Lo que el trabajo tanda de la definición del job deja a las ejecuciones
// posteriores del flujo sobre el mismo commit (contracts/tanda-del-job.md §1 y
// §2 de H7.4; research.md D15 de H7.4): gh run view da el trabajo con su id,
// porque no lleva name, y la marca es su último paso, que solo corre en la
// ejecución que mide el commit. TestDefinicionDelJob exige los dos en la
// definición.
const (
	trabajoDeLaTanda = "tanda"
	marcaDeLaTanda   = "Esta ejecución mide el commit"
)

// Los valores de gh con los que se decide la tanda: el status de una ejecución
// o de un trabajo que ha terminado y la conclusion de un paso que ha terminado
// bien.
const (
	statusTerminado   = "completed"
	conclusionDeExito = "success"
)

// Las órdenes de gh con las que TestTandaDelCommit consulta las ejecuciones del
// flujo sobre el commit (contracts/tanda-del-job.md §2 de H7.4; research.md D16
// y V12 de H7.4): constantes, con cada valor en una variable del entorno de la
// orden y entre comillas, de modo que ningún dato entra en su texto (gosec
// G204). exec sustituye la shell por gh, que termina con su código.
const (
	// ordenDeLasEjecuciones lista las ejecuciones del flujo sobre el commit de
	// variableDelCommitEvaluado, acotadas al commit y no al historial.
	ordenDeLasEjecuciones = `exec gh run list --workflow evals.yml --commit "$COMMIT_EVALUADO" --limit 100 ` +
		`--json databaseId,status`

	// ordenDeLosTrabajos da los trabajos de la ejecución de
	// variableDeLaEjecucionAnterior, con sus pasos.
	ordenDeLosTrabajos = `exec gh run view "$EJECUCION_ANTERIOR" --json jobs`
)

// Las variables del entorno con las que cada orden de gh recibe su valor.
const (
	variableDelCommitEvaluado     = "COMMIT_EVALUADO"
	variableDeLaEjecucionAnterior = "EJECUCION_ANTERIOR"
)

// Las esperas de esperarLaDecision (contracts/tanda-del-job.md §2 de H7.4): la
// que hay entre una consulta y la siguiente mientras alguna ejecución anterior
// no ha decidido, y lo que se espera como mucho a que decida.
const (
	esperaEntreConsultas     = 10 * time.Second
	esperaMaximaDeLaDecision = 10 * time.Minute
)

// ejecucionDelCommit es una ejecución del flujo evals sobre el commit evaluado
// (data-model §7 de H7.4).
type ejecucionDelCommit struct {
	// id es su databaseId, que crece con su creación (research.md S2 de H7.4).
	id int64

	// terminada dice si su status es completed.
	terminada bool

	// tanda es lo que decidió su trabajo tanda; sin decidir si no se ha
	// leído, porque la ejecución no cuenta para la decisión.
	tanda estadoDeLaTanda
}

// estadoDeLaTanda es lo que decidió el trabajo tanda de una ejecución.
type estadoDeLaTanda int

const (
	// tandaSinDecidir: su trabajo tanda no está o no ha terminado.
	tandaSinDecidir estadoDeLaTanda = iota

	// tandaQueMide: su trabajo tanda terminó con la marca en success.
	tandaQueMide

	// tandaQueNoMide: su trabajo tanda terminó sin la marca en success:
	// saltado, fallido, cancelado o con la marca saltada.
	tandaQueNoMide
)

// decisionDeLaTanda es lo que decide una ejecución con una consulta: si mide el
// commit y, si mide, las ejecuciones anteriores sin terminar que aún no han
// decidido, por las que vuelve a consultar.
type decisionDeLaTanda struct {
	mide       bool
	pendientes []int64
}

// decisionTrasLaEspera es lo que devuelve esperarLaDecision: la decisión de su
// última consulta y si mide porque pasó esperaMaximaDeLaDecision con alguna
// ejecución anterior sin decidir, que TestTandaDelCommit dice en su registro.
type decisionTrasLaEspera struct {
	decisionDeLaTanda

	agotada bool
}

// cuentaPara dice si la ejecución cuenta para la decisión de la ejecución
// propia: solo las anteriores, de id menor, y sin terminar. Las posteriores
// deciden respecto a la propia, y la tanda de una anterior terminada ya
// terminó: la etiqueta vuelve a medir (FR-071 de H7.4).
func (e ejecucionDelCommit) cuentaPara(propia int64) bool {
	return e.id < propia && !e.terminada
}

// decidirLaTanda decide si la ejecución propia mide el commit
// (contracts/tanda-del-job.md §2 de H7.4; FR-070 de H7.4): entre las
// ejecuciones que cuentan, con alguna que mide, no mide; si no, con alguna sin
// decidir, mide con ellas como pendientes; y si no, mide.
func decidirLaTanda(propia int64, ejecuciones []ejecucionDelCommit) decisionDeLaTanda {
	var pendientes []int64

	for _, ejecucion := range ejecuciones {
		if !ejecucion.cuentaPara(propia) {
			continue
		}

		switch ejecucion.tanda {
		case tandaQueMide:
			return decisionDeLaTanda{mide: false}
		case tandaSinDecidir:
			pendientes = append(pendientes, ejecucion.id)
		case tandaQueNoMide:
		}
	}

	return decisionDeLaTanda{mide: true, pendientes: pendientes}
}

// esperarLaDecision decide si la ejecución propia mide el commit
// (contracts/tanda-del-job.md §2 de H7.4): consulta las ejecuciones con
// consultar y decide con decidirLaTanda; mientras alguna anterior sin terminar
// no ha decidido, espera esperaEntreConsultas con esperar y vuelve a consultar.
// Pasada esperaMaximaDeLaDecision desde la primera consulta, según ahora, con
// alguna aún sin decidir, mide y lo dice en agotada: la concurrency del trabajo
// evals la pondría detrás de la otra. Un error de una consulta es un error que
// la nombra, como el de una espera.
func esperarLaDecision(ctx context.Context, propia int64,
	consultar func(context.Context) ([]ejecucionDelCommit, error), ahora func() time.Time,
	esperar func(context.Context, time.Duration) error,
) (decisionTrasLaEspera, error) {
	inicio := ahora()

	for consulta := 1; ; consulta++ {
		ejecuciones, err := consultar(ctx)
		if err != nil {
			return decisionTrasLaEspera{}, fmt.Errorf("la consulta %d de las ejecuciones del commit: %w", consulta, err)
		}

		decision := decidirLaTanda(propia, ejecuciones)
		if len(decision.pendientes) == 0 {
			return decisionTrasLaEspera{decisionDeLaTanda: decision}, nil
		}

		if ahora().Sub(inicio) >= esperaMaximaDeLaDecision {
			return decisionTrasLaEspera{decisionDeLaTanda: decision, agotada: true}, nil
		}

		if err := esperar(ctx, esperaEntreConsultas); err != nil {
			return decisionTrasLaEspera{}, fmt.Errorf("la espera tras la consulta %d de las ejecuciones del commit: %w",
				consulta, err)
		}
	}
}

// consultarLasEjecuciones hace una consulta de las ejecuciones del flujo sobre
// el commit (contracts/tanda-del-job.md §2 y §5 de H7.4): lee la lista que da
// listar, la de gh run list --json databaseId,status, y, de cada ejecución que
// cuenta para la propia, la tanda de los trabajos que da verLosTrabajos, los de
// gh run view --json jobs. La tanda de las demás no se lee. El error de listar
// se devuelve tal cual; el de los trabajos de una ejecución, con su id.
func consultarLasEjecuciones(ctx context.Context, propia int64, listar func(context.Context) ([]byte, error),
	verLosTrabajos func(context.Context, int64) ([]byte, error),
) ([]ejecucionDelCommit, error) {
	lista, err := listar(ctx)
	if err != nil {
		return nil, err
	}

	ejecuciones, err := leerLasEjecucionesDelCommit(lista)
	if err != nil {
		return nil, err
	}

	for i, ejecucion := range ejecuciones {
		if !ejecucion.cuentaPara(propia) {
			continue
		}

		trabajos, err := verLosTrabajos(ctx, ejecucion.id)
		if err != nil {
			return nil, fmt.Errorf("los trabajos de la ejecución %d: %w", ejecucion.id, err)
		}

		ejecuciones[i].tanda, err = leerLaTandaDeLaEjecucion(trabajos)
		if err != nil {
			return nil, fmt.Errorf("los trabajos de la ejecución %d: %w", ejecucion.id, err)
		}
	}

	return ejecuciones, nil
}

// ejecucionListada es una ejecución de la lista de gh run list --json
// databaseId,status.
type ejecucionListada struct {
	ID     int64  `json:"databaseId"`
	Status string `json:"status"`
}

// leerLasEjecucionesDelCommit lee la lista de gh run list --json
// databaseId,status: de cada ejecución, su id y si ha terminado, con la tanda
// sin decidir.
func leerLasEjecucionesDelCommit(lista []byte) ([]ejecucionDelCommit, error) {
	var listadas []ejecucionListada
	if err := json.Unmarshal(lista, &listadas); err != nil {
		return nil, fmt.Errorf("la lista de ejecuciones del commit no se puede leer: %w", err)
	}

	ejecuciones := make([]ejecucionDelCommit, 0, len(listadas))
	for _, listada := range listadas {
		ejecuciones = append(ejecuciones, ejecucionDelCommit{id: listada.ID, terminada: listada.Status == statusTerminado})
	}

	return ejecuciones, nil
}

// trabajosDeLaEjecucion es el JSON de gh run view --json jobs, con lo que lee
// leerLaTandaDeLaEjecucion de cada trabajo: su nombre, su status y sus pasos.
// La conclusion de un trabajo no cambia nada: saltado, fallido o cancelado, su
// marca no está en success.
type trabajosDeLaEjecucion struct {
	Jobs []trabajoDeLaEjecucion `json:"jobs"`
}

// trabajoDeLaEjecucion es un trabajo de gh run view --json jobs.
type trabajoDeLaEjecucion struct {
	Name   string              `json:"name"`
	Status string              `json:"status"`
	Steps  []pasoDeLaEjecucion `json:"steps"`
}

// pasoDeLaEjecucion es un paso de un trabajo de gh run view --json jobs.
type pasoDeLaEjecucion struct {
	Name       string `json:"name"`
	Conclusion string `json:"conclusion"`
}

// leerLaTandaDeLaEjecucion lee lo que decidió la tanda de una ejecución del
// JSON de sus trabajos, el de gh run view --json jobs
// (contracts/tanda-del-job.md §2 de H7.4; research.md S4 de H7.4): mide si su
// trabajo tanda terminó y su marca tiene la conclusion success; no ha decidido
// si no tiene trabajo tanda o no ha terminado; y no mide en cualquier otro caso.
func leerLaTandaDeLaEjecucion(contenido []byte) (estadoDeLaTanda, error) {
	var trabajos trabajosDeLaEjecucion
	if err := json.Unmarshal(contenido, &trabajos); err != nil {
		return tandaSinDecidir, fmt.Errorf("no se pueden leer: %w", err)
	}

	indice := slices.IndexFunc(trabajos.Jobs, func(trabajo trabajoDeLaEjecucion) bool {
		return trabajo.Name == trabajoDeLaTanda
	})
	if indice < 0 || trabajos.Jobs[indice].Status != statusTerminado {
		return tandaSinDecidir, nil
	}

	marcada := slices.ContainsFunc(trabajos.Jobs[indice].Steps, func(paso pasoDeLaEjecucion) bool {
		return paso.Name == marcaDeLaTanda && paso.Conclusion == conclusionDeExito
	})
	if marcada {
		return tandaQueMide, nil
	}

	return tandaQueNoMide, nil
}

// ghDeLaTanda ejecuta las órdenes de gh de la tanda con el entorno dado —en
// TestTandaDelCommit, el de la ejecución, con GH_TOKEN y GH_REPO— y el commit
// evaluado (contracts/tanda-del-job.md §2 de H7.4).
type ghDeLaTanda struct {
	entorno []string
	commit  string
}

// listar da la lista de ejecuciones del flujo sobre el commit, la de
// ordenDeLasEjecuciones: la función listar de consultarLasEjecuciones.
func (g ghDeLaTanda) listar(ctx context.Context) ([]byte, error) {
	orden := exec.CommandContext(ctx, "sh", "-c", ordenDeLasEjecuciones)

	return g.ejecutar(orden, variableDelCommitEvaluado+"="+g.commit)
}

// verLosTrabajos da los trabajos de la ejecución id, los de
// ordenDeLosTrabajos: la función verLosTrabajos de consultarLasEjecuciones.
func (g ghDeLaTanda) verLosTrabajos(ctx context.Context, id int64) ([]byte, error) {
	orden := exec.CommandContext(ctx, "sh", "-c", ordenDeLosTrabajos)

	return g.ejecutar(orden, variableDeLaEjecucionAnterior+"="+strconv.FormatInt(id, 10))
}

// ejecutar ejecuta la orden con el entorno de g y, encima, la variable dada, y
// devuelve lo que escribió en su salida estándar. Si no termina con 0, devuelve
// un error con la orden, la variable, cómo terminó y lo que escribió en sus dos
// salidas.
func (g ghDeLaTanda) ejecutar(orden *exec.Cmd, variable string) ([]byte, error) {
	var salida, errores bytes.Buffer

	orden.Env = sobreLaBase(g.entorno, []string{variable})
	orden.Stdout = &salida
	orden.Stderr = &errores

	if err := orden.Run(); err != nil {
		return nil, fmt.Errorf("%s, con %s: %w\n%s%s", orden, variable, err, salida.Bytes(), errores.Bytes())
	}

	return salida.Bytes(), nil
}

// String dice lo que decidió la tanda de una ejecución, para el registro de
// TestTandaDelCommit.
func (e estadoDeLaTanda) String() string {
	switch e {
	case tandaQueMide:
		return "su tanda mide el commit"
	case tandaQueNoMide:
		return "su tanda no mide el commit"
	case tandaSinDecidir:
	}

	return "su tanda aún no ha decidido"
}

// registroDeLaConsulta dice, para el registro de TestTandaDelCommit, qué
// ejecuciones anteriores a la propia miró la consulta, en el orden de la lista,
// y qué vio en cada una: la tanda de las que no han terminado, y que las
// terminadas no cuentan. Las posteriores no se nombran: no cuentan.
func registroDeLaConsulta(consulta int, propia int64, ejecuciones []ejecucionDelCommit) string {
	var lineas []string

	for _, ejecucion := range ejecuciones {
		switch {
		case ejecucion.id >= propia:
		case ejecucion.terminada:
			lineas = append(lineas, fmt.Sprintf("  %d, terminada: no cuenta, su tanda ya terminó", ejecucion.id))
		default:
			lineas = append(lineas, fmt.Sprintf("  %d, sin terminar: %s", ejecucion.id, ejecucion.tanda))
		}
	}

	if len(lineas) == 0 {
		return fmt.Sprintf("consulta %d: ninguna ejecución anterior a la %d sobre el commit", consulta, propia)
	}

	return fmt.Sprintf("consulta %d, ejecuciones anteriores a la %d sobre el commit:\n%s", consulta, propia,
		strings.Join(lineas, "\n"))
}

// motivoDeLaDecision dice, para el registro de TestTandaDelCommit, por qué la
// ejecución propia mide el commit o no, con las ejecuciones de la última
// consulta: no mide porque lo mide la tanda de alguna anterior sin terminar,
// que nombra; y mide porque ninguna anterior sin terminar lo mide o porque
// pasó esperaMaximaDeLaDecision con alguna sin decidir, que nombra.
func motivoDeLaDecision(propia int64, ejecuciones []ejecucionDelCommit, decision decisionTrasLaEspera) string {
	if !decision.mide {
		var queMiden []int64

		for _, ejecucion := range ejecuciones {
			if ejecucion.cuentaPara(propia) && ejecucion.tanda == tandaQueMide {
				queMiden = append(queMiden, ejecucion.id)
			}
		}

		return "no mide el commit: lo mide la tanda de estas ejecuciones anteriores sin terminar: " +
			nombrarLasEjecuciones(queMiden)
	}

	if decision.agotada {
		return fmt.Sprintf("mide el commit: tras %s, la tanda de estas ejecuciones anteriores sin terminar aún no "+
			"ha decidido: %s; la concurrency del trabajo evals pone la de esta detrás de la que mida",
			esperaMaximaDeLaDecision, nombrarLasEjecuciones(decision.pendientes))
	}

	return "mide el commit: ninguna ejecución anterior sin terminar lo mide"
}

// nombrarLasEjecuciones da los ids de las ejecuciones, separados por comas.
func nombrarLasEjecuciones(ids []int64) string {
	nombres := make([]string, 0, len(ids))
	for _, id := range ids {
		nombres = append(nombres, strconv.FormatInt(id, 10))
	}

	return strings.Join(nombres, ", ")
}

// escribirLaDecision añade a la ruta —en TestTandaDelCommit, el GITHUB_OUTPUT
// del paso— la línea medir=si o medir=no, detrás de lo que ya tenga y
// creándola si no existe, como >> en el paso de una definición
// (contracts/tanda-del-job.md §3 de H7.4). El error nombra la ruta.
func escribirLaDecision(ruta string, mide bool) error {
	linea := "medir=no\n"
	if mide {
		linea = "medir=si\n"
	}

	fichero, err := os.OpenFile(filepath.Clean(ruta), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("la decisión de la tanda no se puede escribir en %s: %w", ruta, err)
	}

	if _, err := fichero.WriteString(linea); err != nil {
		return errors.Join(fmt.Errorf("la decisión de la tanda no se puede escribir en %s: %w", ruta, err),
			fichero.Close())
	}

	if err := fichero.Close(); err != nil {
		return fmt.Errorf("la decisión de la tanda no se puede escribir en %s: %w", ruta, err)
	}

	return nil
}

// dormir espera la duración y vuelve sin error o, si el contexto termina
// antes, con su error: la espera de verdad de TestTandaDelCommit entre consulta
// y consulta.
func dormir(ctx context.Context, duracion time.Duration) error {
	temporizador := time.NewTimer(duracion)
	defer temporizador.Stop()

	select {
	case <-temporizador.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
