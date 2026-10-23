package graph

import (
	"context"
	"path/filepath"
	"sync"
)

// turnos son los turnos de escritura de este proceso, uno por world.db: las
// entregas de un mismo proceso —las llamadas simultáneas del servidor MCP, por
// ejemplo— esperan aquí la suya, en orden y mirando su contexto, en lugar de
// competir por el bloqueo de SQLite. Ese bloqueo se reintenta por tramos de
// 100 ms con una espera propia de 5 s (espera.go): entre muchas entregas a la
// vez y con la máquina cargada, alguna la agotaba esperando a otra del mismo
// proceso, y lo observado no llegaba al grafo. Con el turno, la espera propia
// solo cuenta lo que se espera a otra invocación.
//
// El turno de un world.db se crea la primera vez que se pide y no se retira:
// hay uno por cada world.db en que escribe el proceso, uno en producción.
var turnos = struct {
	sync.Mutex

	porRuta map[string]chan struct{}
}{porRuta: map[string]chan struct{}{}}

// tomarTurno espera el turno de escritura de world.db en ruta y devuelve con
// qué soltarlo. Si el contexto termina antes, no lo toma y devuelve el plazo
// agotado, «fuente-no-disponible», como cualquier otra espera de la entrega
// (FR-014).
func tomarTurno(ctx context.Context, ruta string) (func(), error) {
	turno := turnoDe(ruta)

	select {
	case turno <- struct{}{}:
		return func() { <-turno }, nil
	case <-ctx.Done():
		return nil, errorDePlazo(operacionEscribir, ruta, ctx.Err())
	}
}

// turnoDe es el turno de world.db en ruta, que crea si no lo había. Dos
// escrituras de la ruta, limpia, comparten turno; dos que la escriben distinto
// —por un enlace, por ejemplo— no, y entonces compiten en SQLite como dos
// procesos.
func turnoDe(ruta string) chan struct{} {
	clave := filepath.Clean(ruta)

	turnos.Lock()
	defer turnos.Unlock()

	turno, existe := turnos.porRuta[clave]
	if !existe {
		turno = make(chan struct{}, 1)
		turnos.porRuta[clave] = turno
	}

	return turno
}
