package cache

import (
	"context"
	"errors"
	"time"

	sqlite3 "modernc.org/sqlite/lib"
)

const (
	// esperaAnteBloqueo es el tiempo total que una operación espera a que otra
	// invocación suelte el bloqueo de escritura antes de declarar el fallo
	// (FR-031, D4): cinco segundos cubren cualquier transacción de la caché —una
	// migración o un Put— con margen, y un bloqueo que dura más es un fallo real
	// (fila 14 del contrato de errores).
	esperaAnteBloqueo = 5 * time.Second

	// tramoDeEspera es lo que SQLite espera por su cuenta en cada intento, con
	// busy_timeout, antes de devolver SQLITE_BUSY. Es corto a propósito: esa
	// espera vive dentro del motor y no mira el contexto de quien llama —ni
	// sqlite3_interrupt la corta—, así que el contexto solo se puede atender
	// entre tramo y tramo. Cien milisegundos es el retardo que una cancelación
	// puede tardar en hacerse efectiva, y es lo que hace cierto que toda
	// operación termina cuando termina su contexto (FR-003).
	tramoDeEspera = 100 * time.Millisecond
)

// reintentaMientrasBloqueada ejecuta la operación y, mientras SQLite conteste
// que otra invocación tiene la base bloqueada, vuelve a intentarla hasta agotar
// el presupuesto del cliente o hasta que el contexto termine: es la espera ante
// bloqueo de FR-031 hecha de tramos cortos para que respete el contexto (FR-003).
//
// Devuelve el error del último intento tal cual —SQLITE_BUSY si el bloqueo no
// se soltó a tiempo—, y es quien clasifica ese fallo quien mira el contexto para
// decidir si fue el plazo de quien llama (fila 15) o la espera agotada (fila 14).
//
// La operación tiene que ser una sola sentencia en autocommit o el comienzo de
// una transacción: un SQLITE_BUSY dice que no se hizo nada, así que repetirla es
// seguro. Dentro de una transacción ya abierta no se reintenta nada.
//
// El presupuesto se mide con el reloj real y no con el inyectado: lo que se
// espera es tiempo de otra invocación, no la vigencia de ninguna entrada, y un
// reloj de prueba detenido no puede convertir la espera en infinita.
func (c *Cliente) reintentaMientrasBloqueada(ctx context.Context, operacion func() error) error {
	limite := time.Now().Add(c.esperaAnteBloqueo)

	for {
		inicio := time.Now()
		err := operacion()

		if !esBloqueo(err) || ctx.Err() != nil || !time.Now().Before(limite) {
			return err
		}

		// SQLite ya durmió un tramo dentro del intento, salvo que contestara sin
		// invocar su manejador de espera —lo hace cuando esperar podría
		// interbloquear—; entonces se espera aquí lo que falta del tramo, mirando
		// el contexto, para no reintentar en caliente.
		if resto := tramoDeEspera - time.Since(inicio); resto > 0 {
			select {
			case <-ctx.Done():
				return err
			case <-time.After(resto):
			}
		}
	}
}

// esBloqueo dice si el fallo es SQLITE_BUSY, en cualquiera de sus formas
// extendidas: otra conexión tiene el bloqueo que la operación necesita.
func esBloqueo(err error) bool {
	return codigoPrimario(err) == sqlite3.SQLITE_BUSY
}

// terminoElContexto dice si el fallo se debe a que el contexto de quien llama
// terminó, cancelado o vencido, en cualquiera de las tres formas en que eso
// llega: el propio error del contexto, el contexto ya terminado cuando se
// clasifica —lo que ocurre cuando venció durante una espera que el motor no
// interrumpe— y SQLITE_INTERRUPT, que en este paquete solo produce el
// controlador al terminar el contexto. Es lo que impide que un plazo agotado se
// presente como un fichero estropeado o como un bloqueo (fila 15).
func terminoElContexto(ctx context.Context, causa error) bool {
	return ctx.Err() != nil ||
		errors.Is(causa, context.Canceled) ||
		errors.Is(causa, context.DeadlineExceeded) ||
		codigoPrimario(causa) == sqlite3.SQLITE_INTERRUPT
}

// conElErrorDelContexto añade a la causa el error del contexto cuando este
// terminó y la causa no lo lleva ya —el SQLITE_BUSY de la última espera, por
// ejemplo—, de modo que errors.Is alcance las dos cosas: por qué falló la
// operación y por qué se dejó de esperar.
func conElErrorDelContexto(ctx context.Context, causa error) error {
	if err := ctx.Err(); err != nil && !errors.Is(causa, err) {
		return errors.Join(err, causa)
	}

	return causa
}

// codigoPrimario es el código de resultado de SQLite sin su parte extendida: el
// controlador activa los códigos extendidos, así que un SQLITE_BUSY puede llegar
// como SQLITE_BUSY_SNAPSHOT o SQLITE_BUSY_RECOVERY y los tres son un bloqueo.
func codigoPrimario(err error) int {
	return codigoDeSQLite(err) & 0xff
}
