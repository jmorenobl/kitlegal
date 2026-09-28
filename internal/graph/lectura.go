package graph

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// errLecturaCerrada es la causa del fallo de un verbo sobre una lectura ya
// cerrada: no se contesta con un grafo vacío que no lo es.
var errLecturaCerrada = errors.New("la lectura ya está cerrada")

// Lectura es una lectura del grafo del mundo abierta por Leer: lo que usan los
// verbos show, stats y check del applet graph. Cada verbo lee en su propia
// transacción de lectura, que ve una instantánea coherente de lo confirmado, y
// ninguno escribe nada ni devuelve el cuerpo de un texto (FR-005, FR-070).
//
// El grafo ausente, el de 0 bytes y el que no tiene esquema son el grafo vacío:
// la lectura no tiene nada abierto y los verbos lo contestan sin tocar el disco
// (FR-004). Close la cierra y es idempotente; cerrada, los verbos fallan.
//
// Una Lectura es de una sola goroutine, la del verbo que la usa.
type Lectura struct {
	// ruta es la de world.db, para los mensajes.
	ruta string
	// base es la conexión abierta; nil en el grafo vacío.
	base *sql.DB
	// cerrada dice que ya se llamó a Close.
	cerrada bool
}

// Leer abre world.db para leerlo sin dejar rastro (contracts/almacen-world-db.md
// §3; FR-004, FR-005, FR-010, FR-012, FR-014): resuelve la ruta en este
// instante, decide sin abrir SQLite si hay algo que abrir y con qué modo, y lo
// abre leyendo la versión de su esquema. Sin auxiliares no cambia ni un byte de
// nada en el directorio, pueda el proceso escribir world.db o no y también si es
// un enlace simbólico; con auxiliares de WAL, lo único que cambia o aparece es
// lo que SQLite escribe en ellos para leer lo confirmado —la desviación
// declarada de research.md D10—, y world.db, el -wal y el diario quedan como
// estaban.
//
// Un fallo es un *Error con su clase: la ruta no resoluble, «argumentos»; el
// plazo agotado, «fuente-no-disponible»; world.db que es un directorio, que no
// es una base utilizable, que tiene una transacción interrumpida sin deshacer o
// un esquema posterior, y la espera propia agotada, «inesperado».
func Leer(ctx context.Context, opciones ...Opcion) (*Lectura, error) {
	ruta, err := ubicar(operacionLeer, opciones)
	if err != nil {
		return nil, err
	}

	decidida, err := decidirApertura(ruta)
	if err != nil {
		return nil, err
	}

	if decidida.vacia {
		return &Lectura{ruta: ruta}, nil
	}

	base, err := abrirParaLeer(ctx, ruta, decidida)
	if err != nil {
		return nil, err
	}

	return &Lectura{ruta: ruta, base: base}, nil
}

// Ficha es lo que `graph show` devuelve del nodo del id (FR-053;
// contracts/applet-graph.md §3.1): sus datos guardados, su primera y su última
// observación con las fechas tal como las escribió su sobre, y sus aristas
// salientes y entrantes ordenadas por relación y después por el id del otro
// extremo, comparando bytes. El id se compara byte a byte. Si no está —o el
// grafo es el vacío— devuelve false sin error.
func (l *Lectura) Ficha(ctx context.Context, id string) (grafo.Ficha, bool, error) {
	var (
		ficha      grafo.Ficha
		encontrado bool
	)

	err := l.enTransaccion(ctx, func(tx *sql.Tx) (err error) {
		ficha, encontrado, err = leerFicha(ctx, tx, id)

		return err
	})
	if err != nil || !encontrado {
		return grafo.Ficha{}, false, err
	}

	return ficha, true, nil
}

// Recuento es lo que `graph stats` devuelve (FR-054; contracts/applet-graph.md
// §3.2; research.md D16): los totales de nodos, aristas y textos, y los pares
// (tipo, fuente) y (relación, fuente) por la fuente de la última observación,
// contados por el motor y ordenados aquí por bytes, sin ningún par a cero. El
// grafo vacío son tres ceros y dos listas vacías.
func (l *Lectura) Recuento(ctx context.Context) (grafo.Recuento, error) {
	recuento := grafo.Recuento{NodosPorTipo: []grafo.RecuentoDeNodos{}, AristasPorRelacion: []grafo.RecuentoDeAristas{}}

	err := l.enTransaccion(ctx, func(tx *sql.Tx) (err error) {
		recuento, err = leerRecuento(ctx, tx)

		return err
	})
	if err != nil {
		return grafo.Recuento{}, err
	}

	return recuento, nil
}

// Instantanea es todo el grafo de una sola transacción de lectura, lo que lee
// `graph check` (data-model §5; research.md D15): cada nodo con sus datos, su
// última observación y la vigencia que declaró, y cada arista por su terna. Los
// nodos van por id y las aristas por origen, relación y destino, comparando
// bytes, para que el orden no dependa del motor. El grafo vacío son dos listas
// vacías.
func (l *Lectura) Instantanea(ctx context.Context) (grafo.Instantanea, error) {
	instantanea := grafo.Instantanea{Nodos: []grafo.NodoDeInstantanea{}, Aristas: []schema.Arista{}}

	err := l.enTransaccion(ctx, func(tx *sql.Tx) (err error) {
		instantanea, err = leerInstantanea(ctx, tx)

		return err
	})
	if err != nil {
		return grafo.Instantanea{}, err
	}

	return instantanea, nil
}

// Close cierra la lectura. Es idempotente, también sobre el grafo vacío y sobre
// una lectura nula.
func (l *Lectura) Close() error {
	if l == nil || l.cerrada {
		return nil
	}

	l.cerrada = true

	if l.base == nil {
		return nil
	}

	base := l.base
	l.base = nil

	if err := base.Close(); err != nil {
		return errorDeEntradaSalida(operacionLeer, l.ruta, err)
	}

	return nil
}

// enTransaccion hace la lectura dentro de una transacción de lectura, que ve una
// instantánea coherente de lo confirmado, esperando por tramos que miran el
// contexto mientras otra invocación retiene world.db (contracts/almacen-world-db.md
// §3, paso 6; FR-014). Un intento bloqueado se deshace entero y se repite: no
// ha escrito nada. Sobre el grafo vacío no hace nada y la lectura se queda con
// lo que ya tenía, que es el resultado vacío; sobre una lectura cerrada falla.
//
// La transacción empieza sin la cancelación del contexto, para que
// database/sql no la deshaga desde otra goroutine al terminar el plazo, y cada
// consulta lleva el contexto, que es lo que la interrumpe.
func (l *Lectura) enTransaccion(ctx context.Context, leer func(*sql.Tx) error) error {
	if l == nil || l.cerrada {
		ruta := ""
		if l != nil {
			ruta = l.ruta
		}

		return errorDeEntradaSalida(operacionLeer, ruta, errLecturaCerrada)
	}

	if l.base == nil {
		return nil
	}

	espera := esperaDeLaInvocacion()

	err := espera.reintentar(ctx, func() error {
		tx, err := l.base.BeginTx(context.WithoutCancel(ctx), &sql.TxOptions{ReadOnly: true})
		if err != nil {
			return err
		}

		err = leer(tx)
		if deshacer := tx.Rollback(); deshacer != nil && !errors.Is(deshacer, sql.ErrTxDone) {
			err = errors.Join(err, deshacer)
		}

		return err
	})

	if err == nil {
		return nil
	}

	if deLaEspera := espera.falloDeLaEspera(ctx, operacionLeer, l.ruta, err); deLaEspera != nil {
		return deLaEspera
	}

	return errorInutilizable(operacionLeer, l.ruta, err)
}

// leerFicha lee el nodo y sus aristas dentro de la transacción.
func leerFicha(ctx context.Context, tx *sql.Tx, id string) (grafo.Ficha, bool, error) {
	var (
		nodo  grafo.NodoDeFicha
		datos string
	)

	err := tx.QueryRowContext(ctx,
		`SELECT id, type, props, first_seen, last_seen, source, url FROM nodes WHERE id = ?`, id,
	).Scan(&nodo.ID, &nodo.Tipo, &datos, &nodo.PrimeraObservacion,
		&nodo.UltimaObservacion.FechaConsulta, &nodo.UltimaObservacion.Fuente, &nodo.UltimaObservacion.URL)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return grafo.Ficha{}, false, nil
	case err != nil:
		return grafo.Ficha{}, false, err
	}

	if nodo.Datos, err = datosGuardados(datos); err != nil {
		return grafo.Ficha{}, false, err
	}

	salientes, err := aristasDeFicha(ctx, tx,
		`SELECT rel, dst, first_seen, last_seen, source, url FROM edges WHERE src = ?`, id)
	if err != nil {
		return grafo.Ficha{}, false, err
	}

	entrantes, err := aristasDeFicha(ctx, tx,
		`SELECT rel, src, first_seen, last_seen, source, url FROM edges WHERE dst = ?`, id)
	if err != nil {
		return grafo.Ficha{}, false, err
	}

	return grafo.Ficha{Nodo: nodo, Salientes: salientes, Entrantes: entrantes}, true, nil
}

// aristasDeFicha lee las aristas de un extremo —la consulta elige cuál— y las
// ordena por relación y después por el id del otro extremo, comparando bytes
// (FR-053).
func aristasDeFicha(ctx context.Context, tx *sql.Tx, consulta, id string) ([]grafo.AristaDeFicha, error) {
	aristas, err := consultar(ctx, tx, consulta, func(filas *sql.Rows) (grafo.AristaDeFicha, error) {
		var arista grafo.AristaDeFicha

		err := filas.Scan(&arista.Relacion, &arista.ID, &arista.PrimeraObservacion,
			&arista.UltimaObservacion.FechaConsulta, &arista.UltimaObservacion.Fuente, &arista.UltimaObservacion.URL)

		return arista, err
	}, id)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(aristas, func(a, b grafo.AristaDeFicha) int {
		return cmp.Or(strings.Compare(a.Relacion, b.Relacion), strings.Compare(a.ID, b.ID))
	})

	return aristas, nil
}

// leerRecuento cuenta dentro de la transacción: los pares por GROUP BY, sus
// sumas como totales, y el orden por bytes en Go, sin depender de la colación
// del motor (research.md D16).
func leerRecuento(ctx context.Context, tx *sql.Tx) (grafo.Recuento, error) {
	nodos, err := consultar(ctx, tx, `SELECT type, source, count(*) FROM nodes GROUP BY type, source`,
		func(filas *sql.Rows) (grafo.RecuentoDeNodos, error) {
			var par grafo.RecuentoDeNodos

			return par, filas.Scan(&par.Tipo, &par.Fuente, &par.Nodos)
		})
	if err != nil {
		return grafo.Recuento{}, err
	}

	aristas, err := consultar(ctx, tx, `SELECT rel, source, count(*) FROM edges GROUP BY rel, source`,
		func(filas *sql.Rows) (grafo.RecuentoDeAristas, error) {
			var par grafo.RecuentoDeAristas

			return par, filas.Scan(&par.Relacion, &par.Fuente, &par.Aristas)
		})
	if err != nil {
		return grafo.Recuento{}, err
	}

	var textos int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM texts`).Scan(&textos); err != nil {
		return grafo.Recuento{}, err
	}

	recuento := grafo.Recuento{Textos: textos, NodosPorTipo: nodos, AristasPorRelacion: aristas}

	for _, par := range nodos {
		recuento.Nodos += par.Nodos
	}

	for _, par := range aristas {
		recuento.Aristas += par.Aristas
	}

	slices.SortFunc(recuento.NodosPorTipo, func(a, b grafo.RecuentoDeNodos) int {
		return cmp.Or(strings.Compare(a.Tipo, b.Tipo), strings.Compare(a.Fuente, b.Fuente))
	})
	slices.SortFunc(recuento.AristasPorRelacion, func(a, b grafo.RecuentoDeAristas) int {
		return cmp.Or(strings.Compare(a.Relacion, b.Relacion), strings.Compare(a.Fuente, b.Fuente))
	})

	return recuento, nil
}

// leerInstantanea lee todos los nodos y todas las aristas dentro de la
// transacción y los ordena por bytes.
func leerInstantanea(ctx context.Context, tx *sql.Tx) (grafo.Instantanea, error) {
	nodos, err := consultar(ctx, tx, `SELECT id, type, props, last_seen, source, url, ttl FROM nodes`,
		func(filas *sql.Rows) (grafo.NodoDeInstantanea, error) {
			var (
				nodo  grafo.NodoDeInstantanea
				datos string
				ttl   sql.NullInt64
			)

			err := filas.Scan(&nodo.ID, &nodo.Tipo, &datos, &nodo.UltimaObservacion.FechaConsulta,
				&nodo.UltimaObservacion.Fuente, &nodo.UltimaObservacion.URL, &ttl)
			if err != nil {
				return nodo, err
			}

			if nodo.Datos, err = datosGuardados(datos); err != nil {
				return nodo, err
			}

			nodo.Vigencia, err = vigenciaGuardada(ttl)

			return nodo, err
		})
	if err != nil {
		return grafo.Instantanea{}, err
	}

	aristas, err := consultar(ctx, tx, `SELECT src, rel, dst FROM edges`,
		func(filas *sql.Rows) (schema.Arista, error) {
			var arista schema.Arista

			return arista, filas.Scan(&arista.Origen, &arista.Relacion, &arista.Destino)
		})
	if err != nil {
		return grafo.Instantanea{}, err
	}

	slices.SortFunc(nodos, func(a, b grafo.NodoDeInstantanea) int {
		return strings.Compare(a.ID, b.ID)
	})
	slices.SortFunc(aristas, func(a, b schema.Arista) int {
		return cmp.Or(strings.Compare(a.Origen, b.Origen), strings.Compare(a.Relacion, b.Relacion),
			strings.Compare(a.Destino, b.Destino))
	})

	return grafo.Instantanea{Nodos: nodos, Aristas: aristas}, nil
}

// consultar hace la consulta dentro de la transacción y lee cada fila con leer.
// Devuelve una lista vacía, nunca nula, si no hay filas, y el fallo de leer una
// fila, de recorrerlas o de cerrarlas.
func consultar[T any](
	ctx context.Context, tx *sql.Tx, consulta string, leer func(*sql.Rows) (T, error), argumentos ...any,
) (lista []T, err error) {
	filas, err := tx.QueryContext(ctx, consulta, argumentos...)
	if err != nil {
		return nil, err
	}

	defer func() { err = errors.Join(err, filas.Close()) }()

	lista = []T{}

	for filas.Next() {
		cada, err := leer(filas)
		if err != nil {
			return nil, err
		}

		lista = append(lista, cada)
	}

	if err := filas.Err(); err != nil {
		return nil, err
	}

	return lista, nil
}

// datosGuardados son los datos identificativos de un nodo a partir de su forma
// canónica guardada. Lo que no es un objeto JSON no lo escribe ninguna entrega:
// es una base dañada, y no se inventa ningún dato.
func datosGuardados(canonicos string) (map[string]any, error) {
	var datos map[string]any
	if err := json.Unmarshal([]byte(canonicos), &datos); err != nil {
		return nil, fmt.Errorf("los datos guardados de un nodo no son un objeto JSON: %w", err)
	}

	return datos, nil
}

// vigenciaGuardada es la vigencia de una última observación a partir de su
// columna ttl, en segundos: sin valor, cero —no se declaró— (FR-065). Una
// negativa, o una que no cabe en una duración, no la escribe ninguna entrega: es
// una base dañada.
func vigenciaGuardada(ttl sql.NullInt64) (time.Duration, error) {
	if !ttl.Valid {
		return 0, nil
	}

	if ttl.Int64 < 0 || ttl.Int64 > math.MaxInt64/int64(time.Second) {
		return 0, fmt.Errorf("la vigencia guardada de %d s no es una vigencia", ttl.Int64)
	}

	return time.Duration(ttl.Int64) * time.Second, nil
}
