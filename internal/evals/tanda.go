package evals

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
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
