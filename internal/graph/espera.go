package graph

import (
	"context"
	"errors"
	"strconv"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

const (
	// tramoDeEspera es lo que SQLite espera por su cuenta en cada intento, con
	// busy_timeout, antes de contestar SQLITE_BUSY. Esa espera vive dentro del
	// motor y no mira el contexto, así que el contexto solo se atiende entre
	// tramo y tramo: cien milisegundos es lo más que tarda en hacerse efectivo
	// un plazo que termina (FR-014; contracts/almacen-world-db.md §4,
	// «Esperas»).
	tramoDeEspera = 100 * time.Millisecond

	// esperaPropia es lo que una operación espera, en total, a que otra
	// invocación suelte world.db antes de rendirse. Cinco segundos cubren con
	// margen cualquier entrega; un bloqueo que dura más es un fallo real.
	esperaPropia = 5 * time.Second
)

// pragmaDelTramo es el parámetro de toda cadena de conexión que fija el tramo
// dentro del motor, en milisegundos.
var pragmaDelTramo = "_pragma=busy_timeout(" + strconv.FormatInt(tramoDeEspera.Milliseconds(), 10) + ")"

// espera es la espera ante el bloqueo de otra invocación, hecha de tramos que
// miran el contexto, con su total propio. El reloj es el real: lo que se espera
// es tiempo de otra invocación.
type espera struct {
	// total es la espera propia: lo más que se reintenta una operación
	// bloqueada.
	total time.Duration
}

// esperaDeLaInvocacion es la espera de toda lectura y toda entrega: la propia
// de 5 s (§4, «Esperas»).
func esperaDeLaInvocacion() espera {
	return espera{total: esperaPropia}
}

// reintentar hace el intento y, mientras SQLite conteste que otra invocación
// tiene world.db bloqueada, lo repite tramo a tramo hasta que salga, hasta que
// termine el contexto o hasta agotar la espera propia, lo que llegue antes.
// Devuelve el error del último intento tal cual; clasificarlo es cosa de
// falloDeLaEspera y de quien llama.
//
// El intento tiene que ser una sentencia en autocommit o el comienzo de una
// transacción: un SQLITE_BUSY dice que no se hizo nada, así que repetirlo es
// seguro. Dentro de una transacción abierta no se reintenta nada.
func (e espera) reintentar(ctx context.Context, intento func() error) error {
	limite := time.Now().Add(e.total)

	for {
		inicio := time.Now()
		err := intento()

		if !esBloqueo(err) || ctx.Err() != nil || !time.Now().Before(limite) {
			return err
		}

		// SQLite ya esperó un tramo dentro del intento, salvo que contestara
		// sin esperar —lo hace cuando esperar podría interbloquear—: entonces
		// se completa aquí el tramo, mirando el contexto, para no reintentar en
		// caliente; y nunca más allá de la espera propia.
		resto := min(tramoDeEspera-time.Since(inicio), time.Until(limite))
		if resto <= 0 {
			continue
		}

		temporizador := time.NewTimer(resto)

		select {
		case <-ctx.Done():
			temporizador.Stop()

			return err
		case <-temporizador.C:
		}
	}
}

// falloDeLaEspera clasifica lo que la espera explica (§4, «Esperas», y §6): el
// contexto terminado es el plazo agotado, «fuente-no-disponible», con el error
// del contexto unido a la causa; SQLITE_BUSY con el contexto vivo es la espera
// propia agotada, «inesperado». Cualquier otro fallo lo clasifica quien llama,
// y para él devuelve nil, igual que sin fallo.
func (e espera) falloDeLaEspera(ctx context.Context, operacion, ruta string, causa error) error {
	switch {
	case causa == nil:
		return nil
	case terminoElContexto(ctx, causa):
		return errorDePlazo(operacion, ruta, conElErrorDelContexto(ctx, causa))
	case esBloqueo(causa):
		return errorDeBloqueo(operacion, ruta, e.total, causa)
	default:
		return nil
	}
}

// terminoElContexto dice si el fallo se debe a que terminó el contexto de quien
// llama, en cualquiera de las formas en que llega: el contexto ya terminado al
// clasificar —vencido durante una espera que el motor no interrumpe—, el error
// del propio contexto o SQLITE_INTERRUPT, que el controlador solo produce al
// terminar el contexto. Es lo que impide presentar un plazo agotado como un
// fichero estropeado o como un bloqueo.
func terminoElContexto(ctx context.Context, causa error) bool {
	return ctx.Err() != nil ||
		errors.Is(causa, context.Canceled) ||
		errors.Is(causa, context.DeadlineExceeded) ||
		codigoPrimario(causa) == sqlite3.SQLITE_INTERRUPT
}

// conElErrorDelContexto une a la causa el error del contexto terminado si no
// lo lleva ya, para que errors.Is alcance las dos cosas: por qué falló la
// operación y por qué se dejó de esperar.
func conElErrorDelContexto(ctx context.Context, causa error) error {
	if err := ctx.Err(); err != nil && !errors.Is(causa, err) {
		return errors.Join(err, causa)
	}

	return causa
}

// esBloqueo dice si el fallo es SQLITE_BUSY en cualquiera de sus formas
// extendidas: otra conexión tiene el bloqueo que el intento necesita.
func esBloqueo(err error) bool {
	return codigoPrimario(err) == sqlite3.SQLITE_BUSY
}

// codigoPrimario es el código de resultado de SQLite sin su parte extendida.
func codigoPrimario(err error) int {
	return codigoDeSQLite(err) & 0xff
}

// codigoDeSQLite es el código de resultado, extendido, que el controlador
// adjunta al fallo; 0 —SQLITE_OK, que ningún fallo lleva— si el fallo no viene
// de él.
func codigoDeSQLite(err error) int {
	var delMotor *sqlite.Error
	if errors.As(err, &delMotor) {
		return delMotor.Code()
	}

	return 0
}
