package graph

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json/v2"
	"errors"
	"fmt"
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
	// conLecturas dice que la base tiene la tabla lecturas, la de la versión 2
	// del esquema. La 1, la que escribe H7, no la tiene: todos sus bloques
	// están sin fila (H7.1 data-model §3, research.md D4).
	conLecturas bool
	// cerrada dice que ya se llamó a Close.
	cerrada bool
}

// Leer abre world.db para leerlo (contracts/almacen-world-db.md §3; H7 FR-004,
// FR-005, FR-012, FR-014; H7.1 FR-070): resuelve la ruta en este instante,
// decide sin abrir SQLite si hay algo que abrir y, por la existencia de
// world.db-wal, con qué modo, y lo abre leyendo la versión de su esquema. Una
// base de la versión 1, la de H7, se lee sin migrarla y sin ninguna fila de
// lecturas (H7.1 FR-026, research.md D4). Sobre lo que dejan las entregas, sin
// -wal no cambia ni un byte de nada en el directorio; con él, lo único que
// cambia o aparece es el -shm que SQLite escribe para leer lo confirmado —la
// desviación declarada de research.md D10 de H7—, y world.db y el -wal quedan
// como estaban.
//
// Un fallo es un *Error con su clase: la ruta no resoluble, «argumentos»; el
// plazo agotado, «fuente-no-disponible»; world.db con un esquema posterior o
// que el binario no puede usar por cualquier otra causa —la regla genérica,
// sin ninguna promesa sobre sus bytes—, y la espera propia agotada,
// «inesperado».
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

	base, version, err := abrirParaLeer(ctx, ruta, decidida.modo)
	if err != nil {
		return nil, err
	}

	return &Lectura{ruta: ruta, base: base, conLecturas: version >= versionDeLasLecturas}, nil
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

// Instantanea es lo que lee `graph check` del ámbito en una sola transacción de
// lectura (data-model §5; research.md D15; H7.1 data-model §3 y §6,
// contracts/almacen-world-db.md §3, paso 4, research.md D8): cada nodo con sus
// datos, su última observación y la vigencia que declaró, cada arista por su
// terna y cada fila de lecturas. Sin norma, todo el grafo. Con una norma, lo
// lee acotado en SQL: la Norma con ese identificador BOE, los Bloque a los que
// llega eli:has_part desde ella —los nombrados, si se nombra alguno—, las
// BloqueVersion a las que llega eli:has_version desde esos bloques, esas
// aristas y las filas de lecturas de esos bloques. Una norma o un bloque que el
// grafo no conoce no es un error: la instantánea no lo trae (FR-003).
//
// Los nodos van por id, las aristas por origen, relación y destino y las filas
// por bloque, comparando bytes, para que el orden no dependa del motor. El
// grafo vacío y una base de la versión 1, sin la tabla lecturas, no tienen
// ninguna fila; el grafo vacío son tres listas vacías.
func (l *Lectura) Instantanea(ctx context.Context, ambito grafo.Ambito) (grafo.Instantanea, error) {
	consultas, argumentos, err := consultasDelAmbito(ambito)
	if err != nil {
		return grafo.Instantanea{}, errorDeEntradaSalida(operacionLeer, l.laRuta(), err)
	}

	instantanea := grafo.Instantanea{
		Nodos: []grafo.NodoDeInstantanea{}, Aristas: []schema.Arista{}, Lecturas: []grafo.LecturasDeBloque{},
	}

	err = l.enTransaccion(ctx, func(tx *sql.Tx) (err error) {
		instantanea, err = leerInstantanea(ctx, tx, consultas, l.conLecturas, argumentos)

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

// laRuta es la de world.db, para los mensajes; vacía sobre una lectura nula.
func (l *Lectura) laRuta() string {
	if l == nil {
		return ""
	}

	return l.ruta
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
		return errorDeEntradaSalida(operacionLeer, l.laRuta(), errLecturaCerrada)
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

// consultasDeInstantanea son las tres consultas de una instantánea: la de sus
// nodos, la de sus aristas y la de sus filas de lecturas, que solo se hace si
// la base tiene la tabla.
type consultasDeInstantanea struct {
	nodos, aristas, lecturas string
}

// Lo que la instantánea lee de los nodos y de las filas de lecturas, con
// ámbito o sin él.
const (
	nodosDeLaInstantanea    = `SELECT id, type, props, last_seen, source, url, ttl FROM nodes`
	lecturasDeLaInstantanea = `SELECT bloque, ultima, anterior FROM lecturas`
)

// ambitoDeUnaNorma son las tablas comunes de las consultas de la instantánea de
// una norma (contracts/almacen-world-db.md §3, paso 4; research.md D8, V7): la
// Norma, por su tipo y su identificador BOE; sus Bloque, por la clave primaria
// de edges desde ella, filtrados por su id de bloque si :bloques, una lista
// JSON, nombra alguno; y sus BloqueVersion, por la clave primaria de edges
// desde esos bloques. Los textos se comparan byte a byte, con la colación
// BINARY de SQLite.
const ambitoDeUnaNorma = `WITH normas (id) AS (
	SELECT id FROM nodes
	WHERE type = :tipo_norma AND json_extract(props, '$.` + grafo.DatoIdentificador + `') = :norma
), bloques (id) AS (
	SELECT edges.dst FROM normas
	JOIN edges ON edges.src = normas.id AND edges.rel = :tiene_parte
	JOIN nodes ON nodes.id = edges.dst
	WHERE json_array_length(:bloques) = 0
		OR json_extract(nodes.props, '$.` + grafo.DatoBloque + `') IN (SELECT value FROM json_each(:bloques))
), versiones (id) AS (
	SELECT edges.dst FROM bloques
	JOIN edges ON edges.src = bloques.id AND edges.rel = :tiene_version
)
`

var (
	// todoElGrafo son las consultas de la instantánea sin ámbito: todo, como
	// en H7.
	todoElGrafo = consultasDeInstantanea{
		nodos:    nodosDeLaInstantanea,
		aristas:  `SELECT src, rel, dst FROM edges`,
		lecturas: lecturasDeLaInstantanea,
	}

	// unaNorma son las consultas de la instantánea de una norma: los nodos de
	// ambitoDeUnaNorma, las aristas eli:has_part de la norma a esos bloques y
	// eli:has_version de esos bloques a sus versiones, y las filas de esos
	// bloques.
	unaNorma = consultasDeInstantanea{
		nodos: ambitoDeUnaNorma + nodosDeLaInstantanea + `
WHERE id IN (SELECT id FROM normas UNION ALL SELECT id FROM bloques UNION ALL SELECT id FROM versiones)`,
		aristas: ambitoDeUnaNorma + `SELECT edges.src, edges.rel, edges.dst FROM normas
JOIN edges ON edges.src = normas.id AND edges.rel = :tiene_parte
WHERE edges.dst IN (SELECT id FROM bloques)
UNION ALL
SELECT edges.src, edges.rel, edges.dst FROM bloques
JOIN edges ON edges.src = bloques.id AND edges.rel = :tiene_version`,
		lecturas: ambitoDeUnaNorma + lecturasDeLaInstantanea + `
WHERE bloque IN (SELECT id FROM bloques)`,
	}
)

// consultasDelAmbito son las consultas de la instantánea del ámbito y sus
// argumentos: sin norma, las de todo el grafo, sin ninguno; con norma, las de
// una norma, con la norma, los bloques como lista JSON —vacía si no se nombra
// ninguno— y el tipo y las relaciones que siguen.
func consultasDelAmbito(ambito grafo.Ambito) (consultasDeInstantanea, []any, error) {
	if ambito.Norma == "" {
		return todoElGrafo, nil, nil
	}

	// encoding/json/v2 escribe una lista nula como [], no como null.
	bloques, err := json.Marshal(ambito.Bloques)
	if err != nil {
		return consultasDeInstantanea{}, nil, err
	}

	return unaNorma, []any{
		sql.Named("tipo_norma", grafo.TipoNorma),
		sql.Named("norma", ambito.Norma),
		// Como texto: SQLite leería un BLOB como JSONB.
		sql.Named("bloques", string(bloques)),
		sql.Named("tiene_parte", grafo.RelacionTieneParte),
		sql.Named("tiene_version", grafo.RelacionTieneVersion),
	}, nil
}

// leerInstantanea hace las consultas dentro de la transacción —la de lecturas,
// solo si la base tiene la tabla— y ordena por bytes lo que leen.
func leerInstantanea(
	ctx context.Context, tx *sql.Tx, consultas consultasDeInstantanea, conLecturas bool, argumentos []any,
) (grafo.Instantanea, error) {
	nodos, err := consultar(ctx, tx, consultas.nodos,
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

			nodo.Vigencia = vigenciaGuardada(ttl)

			return nodo, nil
		}, argumentos...)
	if err != nil {
		return grafo.Instantanea{}, err
	}

	aristas, err := consultar(ctx, tx, consultas.aristas,
		func(filas *sql.Rows) (schema.Arista, error) {
			var arista schema.Arista

			return arista, filas.Scan(&arista.Origen, &arista.Relacion, &arista.Destino)
		}, argumentos...)
	if err != nil {
		return grafo.Instantanea{}, err
	}

	lecturas := []grafo.LecturasDeBloque{}

	if conLecturas {
		lecturas, err = consultar(ctx, tx, consultas.lecturas,
			func(filas *sql.Rows) (grafo.LecturasDeBloque, error) {
				var fila grafo.LecturasDeBloque

				return fila, filas.Scan(&fila.Bloque, &fila.Ultima, &fila.Anterior)
			}, argumentos...)
		if err != nil {
			return grafo.Instantanea{}, err
		}
	}

	slices.SortFunc(nodos, func(a, b grafo.NodoDeInstantanea) int {
		return strings.Compare(a.ID, b.ID)
	})
	slices.SortFunc(aristas, func(a, b schema.Arista) int {
		return cmp.Or(strings.Compare(a.Origen, b.Origen), strings.Compare(a.Relacion, b.Relacion),
			strings.Compare(a.Destino, b.Destino))
	})
	slices.SortFunc(lecturas, func(a, b grafo.LecturasDeBloque) int {
		return strings.Compare(a.Bloque, b.Bloque)
	})

	return grafo.Instantanea{Nodos: nodos, Aristas: aristas, Lecturas: lecturas}, nil
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
// columna ttl, en segundos: sin valor, cero —no se declaró— (FR-065). La
// columna la escribe la entrega a partir de la vigencia del lote; lo que no se
// puede leer como entero ya lo devuelve la lectura de la fila (H7.1 FR-070).
func vigenciaGuardada(ttl sql.NullInt64) time.Duration {
	if !ttl.Valid {
		return 0
	}

	return time.Duration(ttl.Int64) * time.Second
}
