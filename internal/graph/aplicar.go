package graph

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/grafo"
)

const (
	// modoWAL es lo que contesta PRAGMA journal_mode sobre una base en WAL.
	modoWAL = "wal"

	// permisosDeDirectorio y permisosDeFichero son los de cada directorio que
	// crea la entrega y los de world.db: el grafo es de la cuenta, como la
	// caché (H7 research.md D8; H7.1 research.md D18).
	permisosDeDirectorio fs.FileMode = 0o700
	permisosDeFichero    fs.FileMode = 0o600
)

// cadenaDeEscritura es la cadena de conexión de toda entrega, carácter a
// carácter la de contracts/almacen-world-db.md §4 (research.md D11, V10): el
// tramo de espera, cada confirmación llevada al disco, las claves ajenas
// comprobadas y la transacción inmediata, que toma el bloqueo de escritura al
// empezar. No lleva journal_mode: en la cadena, un world.db de 0 bytes pasaría a
// 4096 aunque el lote se rechazara; el WAL lo fija fijarWAL, fuera de toda
// transacción.
func cadenaDeEscritura(ruta string) string {
	return "file:" + rutaParaURI(ruta) + "?" + pragmaDelTramo +
		"&_pragma=synchronous(FULL)&_pragma=foreign_keys(1)&_txlock=immediate"
}

// consolidar es el paso 1 de la entrega, sin tocar el disco: valida el lote y lo
// reduce a un registro por clave (grafo.Consolidar; FR-023, FR-024, FR-025), y
// rechaza una vigencia que no es un número entero de segundos, que es como se
// guarda (data-model §3): guardarla recortada no sería la que la fuente declara
// (FR-065), y ningún applet emite una así.
func consolidar(lote core.Lote) (grafo.Consolidado, error) {
	consolidado, err := grafo.Consolidar(lote)
	if err != nil {
		return grafo.Consolidado{}, err
	}

	if lote.Vigencia%time.Second != 0 {
		return grafo.Consolidado{}, &grafo.Rechazo{
			Motivo: fmt.Sprintf("la vigencia %s no es un número entero de segundos, que es como se guarda", lote.Vigencia),
		}
	}

	return consolidado, nil
}

// aplicarEnSuSitio hace los pasos 2 a 9 de la entrega (contracts/almacen-world-db.md
// §4): crea en su sitio lo que falte, abre world.db y lo prepara, aplica el
// lote en una transacción y lo cierra. Un fallo al cerrar se une al de la
// entrega, o es él el fallo.
func aplicarEnSuSitio(ctx context.Context, ruta string, consolidado grafo.Consolidado) error {
	if err := crearEnSuSitio(ruta); err != nil {
		return err
	}

	base, err := abrirParaEscribir(ctx, ruta)
	if err != nil {
		return err
	}

	return errors.Join(escribirLote(ctx, base, ruta, consolidado), cerrarTrasElFallo(base, operacionEscribir, ruta))
}

// crearEnSuSitio hace los pasos 2 y 3 de la entrega (contracts/almacen-world-db.md
// §4; H7.1 research.md D18): crea el directorio de world.db con los que le
// falten, en 0700, y world.db en su sitio, en 0600, si no está, sin temporal ni
// enlace (H7.1 FR-071); lo cierra sin leerlo ni escribirlo, así que uno que ya
// estaba no cambia. Una entrega que falla después deja como mucho el
// directorio y world.db sin esquema, que la siguiente completa (H7 FR-004,
// FR-013).
func crearEnSuSitio(ruta string) error {
	directorio := filepath.Dir(ruta)

	if err := os.MkdirAll(directorio, permisosDeDirectorio); err != nil {
		return errorDeDirectorioNoEscribible(directorio, err)
	}

	fichero, err := os.OpenFile(filepath.Clean(ruta), os.O_RDWR|os.O_CREATE, permisosDeFichero)
	if err != nil {
		return errorDeEntradaSalida(operacionEscribir, ruta, err)
	}

	if err := fichero.Close(); err != nil {
		return errorDeEntradaSalida(operacionEscribir, ruta, err)
	}

	return nil
}

// abrirParaEscribir es el paso 4 (contracts/almacen-world-db.md §4; research.md
// D11): abre world.db con la cadena de escritura y lo prepara (prepararLaBase).
// Devuelve la conexión abierta; si falla, la cierra.
func abrirParaEscribir(ctx context.Context, ruta string) (*sql.DB, error) {
	base, err := abrirConexion(operacionEscribir, ruta, cadenaDeEscritura(ruta))
	if err != nil {
		return nil, err
	}

	if err := prepararLaBase(ctx, base, ruta); err != nil {
		return nil, errors.Join(err, cerrarTrasElFallo(base, operacionEscribir, ruta))
	}

	return base, nil
}

// prepararLaBase lee la versión del esquema, con el reintento por tramos, y
// decide: una de la 1 a la que este binario conoce sigue, y la transacción de
// la entrega la migra si no es la conocida (H7.1 contracts/almacen-world-db.md
// §4, paso 4); una posterior no se toca (FR-012); cualquier otro resultado es
// un world.db que el binario no puede usar, la regla genérica (H7.1 FR-070); y
// sin esquema —versión 0— world.db se pone en WAL antes de la transacción que
// creará el esquema, porque dentro de ella SQLite no lo fija (research.md
// V42).
//
// Si world.db tiene el -wal de una escritura interrumpida, esta conexión lo
// recupera como cualquier escritor de SQLite, aunque la entrega falle después
// (H7.1 FR-077).
func prepararLaBase(ctx context.Context, base *sql.DB, ruta string) error {
	conocida, err := versionConocida()
	if err != nil {
		return errorDeMigracionesIlegibles(ruta, err)
	}

	version, err := leerVersion(ctx, base)
	if err != nil {
		if deLaEspera := esperaDeLaInvocacion().falloDeLaEspera(ctx, operacionEscribir, ruta, err); deLaEspera != nil {
			return deLaEspera
		}

		return errorInutilizable(operacionEscribir, ruta, err)
	}

	switch {
	case version > conocida:
		return errorDeVersionPosterior(operacionEscribir, ruta, version, conocida)
	case version < 0:
		return errorInutilizable(operacionEscribir, ruta, nil)
	case version > 0:
		return nil
	}

	return fijarWAL(ctx, base, ruta)
}

// fijarWAL pone la base en WAL si no lo está, fuera de toda transacción y con el
// reintento por tramos, y comprueba que el pragma contesta wal (contracts/almacen-world-db.md
// §4, paso 4; FR-003). Dentro de una transacción SQLite no lo fija: sobre
// una base en modo rollback falla, y sobre un fichero de 0 bytes contesta delete
// sin ningún error (research.md V42), que es lo que este paso no deja pasar.
func fijarWAL(ctx context.Context, base consultante, ruta string) error {
	modo, err := modoDeDiario(ctx, base, `PRAGMA journal_mode`)
	if err == nil && modo != modoWAL {
		modo, err = modoDeDiario(ctx, base, `PRAGMA journal_mode=WAL`)
	}

	switch {
	case err != nil:
		return falloAlEscribir(ctx, ruta, err)
	case modo != modoWAL:
		return errorDeModoWAL(ruta, modo)
	}

	return nil
}

// modoDeDiario es lo que contesta el pragma de modo de diario, consultado o
// fijado, esperando por tramos mientras otra invocación retiene world.db.
func modoDeDiario(ctx context.Context, base consultante, pragma string) (string, error) {
	var modo string

	err := esperaDeLaInvocacion().reintentar(ctx, func() error {
		return base.QueryRowContext(ctx, pragma).Scan(&modo)
	})

	return modo, err
}

// escribirLote hace los pasos 5 a 9 de la entrega (contracts/almacen-world-db.md
// §4): abre la transacción inmediata con el reintento por tramos, crea en ella
// el esquema si world.db no lo tiene o lo migra a la versión conocida (FR-013),
// aplica el lote consolidado y confirma. Cualquier rechazo o fallo la deshace
// entera, así que no entra nada del lote, tampoco sus lecturas (FR-022,
// FR-024; H7.1 FR-020).
//
// La transacción empieza sin la cancelación del contexto, para que database/sql
// no la deshaga desde otra goroutine al terminar el plazo; cada sentencia lleva
// el contexto, que es lo que la interrumpe.
func escribirLote(ctx context.Context, base *sql.DB, ruta string, consolidado grafo.Consolidado) error {
	var tx *sql.Tx

	err := esperaDeLaInvocacion().reintentar(ctx, func() (err error) {
		tx, err = base.BeginTx(context.WithoutCancel(ctx), nil)

		return err
	})
	if err != nil {
		return falloAlEscribir(ctx, ruta, err)
	}

	err = migrar(ctx, tx, ruta)
	if err == nil {
		err = aplicarConsolidado(ctx, tx, consolidado)
	}

	if err != nil {
		return errors.Join(falloAlEscribir(ctx, ruta, err), deshacer(tx, ruta))
	}

	if err := tx.Commit(); err != nil {
		return falloAlEscribir(ctx, ruta, err)
	}

	return nil
}

// deshacer deshace la transacción de una entrega que falla. Una transacción que
// ya terminó no es un fallo; cualquier otro fallo se devuelve para unirlo al de
// la entrega.
func deshacer(tx *sql.Tx, ruta string) error {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return errorDeEntradaSalida(operacionEscribir, ruta, err)
	}

	return nil
}

// falloAlEscribir clasifica un fallo de la entrega con world.db ya abierto, en
// el orden de lo que lo explica: lo que ya es un *Error —la migración, un
// esquema posterior— se devuelve tal cual; un rechazo del dominio es el lote que
// no entra (FR-024); el contexto terminado es el plazo agotado, y SQLITE_BUSY con
// el contexto vivo, la espera propia agotada (FR-014); y cualquier otra cosa
// —también fusionar con lo guardado que ninguna entrega escribe—, un fallo de
// entrada y salida que nombra world.db (H7.1 research.md D21).
func falloAlEscribir(ctx context.Context, ruta string, causa error) error {
	var (
		clasificado *Error
		rechazo     *grafo.Rechazo
	)

	switch {
	case errors.As(causa, &clasificado):
		return causa
	case errors.As(causa, &rechazo):
		return errorDeLoteRechazado(ruta, causa)
	}

	if deLaEspera := esperaDeLaInvocacion().falloDeLaEspera(ctx, operacionEscribir, ruta, causa); deLaEspera != nil {
		return deLaEspera
	}

	return errorDeEntradaSalida(operacionEscribir, ruta, causa)
}

// aplicarConsolidado es los pasos 6 a 8 (contracts/almacen-world-db.md §4):
// antes de escribir nada del lote, la fila de lecturas que deja cada lectura
// suya (lecturasQueCambian); después, cada nodo, luego cada texto y luego cada
// arista se fusiona con lo guardado y se escribe solo si cambia; y al final,
// cada fila de lecturas que cambia. Los nodos van primero para que una arista y
// una fila de lecturas encuentren en el grafo los nodos que trae el propio
// lote.
func aplicarConsolidado(ctx context.Context, tx *sql.Tx, consolidado grafo.Consolidado) error {
	lecturas, err := lecturasQueCambian(ctx, tx, consolidado.Lecturas())
	if err != nil {
		return err
	}

	if err := tablaDeNodos.aplicar(ctx, tx, consolidado.Nodos); err != nil {
		return err
	}

	if err := tablaDeTextos.aplicar(ctx, tx, consolidado.Textos); err != nil {
		return err
	}

	if err := tablaDeAristas.aplicar(ctx, tx, consolidado.Aristas); err != nil {
		return err
	}

	return escribirLecturas(ctx, tx, lecturas)
}

// lecturasQueCambian es el paso 6 (contracts/almacen-world-db.md §4; H7.1
// data-model §4): para cada lectura del lote, la fila de su bloque tras ella
// —Leida sobre la guardada o, sin fila, sobre la de partida (partidaSinFila)—,
// calculada sobre lo que el grafo guarda antes del lote. Devuelve solo las que
// no están guardadas tal cual, que son las que hay que escribir: con (v, v)
// guardada, una lectura que ve v no deja nada, y una entrega repetida idéntica
// no escribe (H7 FR-022).
func lecturasQueCambian(ctx context.Context, tx *sql.Tx, lecturas []grafo.Lectura) ([]grafo.LecturasDeBloque, error) {
	var cambian []grafo.LecturasDeBloque

	for _, lectura := range lecturas {
		guardada, estaba, err := leerFilaDeLecturas(ctx, tx, lectura.Bloque)
		if err != nil {
			return nil, err
		}

		partida := guardada

		if !estaba {
			if partida, err = partidaSinFila(ctx, tx, lectura); err != nil {
				return nil, err
			}
		}

		nueva := partida.Leida(lectura.Version)
		if estaba && nueva == guardada {
			continue
		}

		cambian = append(cambian, nueva)
	}

	return cambian, nil
}

// leerFilaDeLecturas lee la fila guardada del bloque, y false si no tiene.
func leerFilaDeLecturas(ctx context.Context, tx *sql.Tx, bloque string) (grafo.LecturasDeBloque, bool, error) {
	fila := grafo.LecturasDeBloque{Bloque: bloque}

	estaba, err := encontrada(tx.QueryRowContext(ctx, lecturaDeLecturas, bloque).Scan(&fila.Ultima, &fila.Anterior))
	if !estaba {
		return grafo.LecturasDeBloque{}, false, err
	}

	return fila, true, nil
}

// partidaSinFila es la fila de partida de un bloque sin fila de lecturas —el de
// un world.db de H7, o uno que H7 observó y nadie ha vuelto a leer—, que cuenta
// con una lectura (H7.1 FR-026; research.md D3): (R, R), con R su redacción
// vista sin lecturas entre las BloqueVersion que el grafo ya guarda de él
// (grafo.RedaccionVistaSinLecturas). Si no guarda ninguna, el bloque es nuevo y
// la partida es la propia lectura, de modo que su primera lectura deja (v, v).
func partidaSinFila(ctx context.Context, tx *sql.Tx, lectura grafo.Lectura) (grafo.LecturasDeBloque, error) {
	versiones, err := consultar(ctx, tx, lecturaDeVersionesGuardadas,
		func(filas *sql.Rows) (grafo.NodoDeInstantanea, error) {
			var version grafo.NodoDeInstantanea

			return version, filas.Scan(&version.ID, &version.UltimaObservacion.FechaConsulta)
		}, lectura.Bloque, grafo.RelacionTieneVersion, grafo.TipoBloqueVersion)
	if err != nil {
		return grafo.LecturasDeBloque{}, err
	}

	vista, hay, err := grafo.RedaccionVistaSinLecturas(versiones)
	if err != nil {
		return grafo.LecturasDeBloque{}, err
	}

	if !hay {
		vista = lectura.Version
	}

	return grafo.LecturasDeBloque{Bloque: lectura.Bloque, Ultima: vista, Anterior: vista}, nil
}

// escribirLecturas es el paso 8 (contracts/almacen-world-db.md §4): guarda cada
// fila, nueva o en lugar de la que tenía su bloque. La clave ajena exige que
// sus tres nodos estén ya en el grafo.
func escribirLecturas(ctx context.Context, tx *sql.Tx, filas []grafo.LecturasDeBloque) error {
	for _, fila := range filas {
		if _, err := tx.ExecContext(ctx, escrituraDeLecturas, fila.Bloque, fila.Ultima, fila.Anterior); err != nil {
			return err
		}
	}

	return nil
}

// tablaDelGrafo es lo que la entrega sabe hacer con los registros de una tabla
// de world.db: leer el guardado con la clave del que llega, fusionarlos
// (grafo.FusionarNodo, FusionarArista, FusionarTexto; data-model §4.1) y
// escribir el fusionado.
type tablaDelGrafo[R comparable] struct {
	// leer devuelve el registro guardado con la clave del que llega, y false si
	// no hay ninguno.
	leer func(ctx context.Context, tx *sql.Tx, llegado R) (R, bool, error)
	// fusionar junta lo guardado con lo que llega, o rechaza el lote.
	fusionar func(guardado, llegado R) (R, error)
	// escribir guarda el registro, nuevo o fusionado, con su clave.
	escribir func(ctx context.Context, tx *sql.Tx, registro R) error
}

// Las tres tablas de world.db, cada una con sus sentencias.
var (
	tablaDeNodos   = tablaDelGrafo[grafo.RegistroDeNodo]{leer: leerNodo, fusionar: grafo.FusionarNodo, escribir: escribirNodo}
	tablaDeTextos  = tablaDelGrafo[grafo.RegistroDeTexto]{leer: leerTexto, fusionar: grafo.FusionarTexto, escribir: escribirTexto}
	tablaDeAristas = tablaDelGrafo[grafo.RegistroDeArista]{
		leer: leerArista, fusionar: grafo.FusionarArista, escribir: escribirArista,
	}
)

// aplicar fusiona cada registro que llega con el guardado y escribe el resultado
// solo si cambia: uno que no estaba se escribe tal cual, y uno cuya fusión es lo
// que ya estaba guardado no se escribe, de modo que una entrega repetida no
// cambia nada (FR-022, FR-023). Un tipo o un cuerpo distintos los rechaza la
// fusión con un *grafo.Rechazo; cualquier otro fallo de la fusión se devuelve
// tal cual y lo clasifica falloAlEscribir.
func (t tablaDelGrafo[R]) aplicar(ctx context.Context, tx *sql.Tx, registros []R) error {
	for _, llegado := range registros {
		guardado, estaba, err := t.leer(ctx, tx, llegado)
		if err != nil {
			return err
		}

		fusionado := llegado

		if estaba {
			fusionado, err = t.fusionar(guardado, llegado)

			switch {
			case err != nil:
				return err
			case fusionado == guardado:
				continue
			}
		}

		if err := t.escribir(ctx, tx, fusionado); err != nil {
			return err
		}
	}

	return nil
}

// columnasDeHistoria son los siete campos que nodes y edges comparten, en este
// orden: la primera observación (first_seen, first_source, first_url), la
// última (last_seen, source, url) y su vigencia en segundos (ttl).
const columnasDeHistoria = "first_seen, first_source, first_url, last_seen, source, url, ttl"

// historiaExcluida pone los siete campos de la historia con los valores de la
// fila que no entró por el conflicto de clave: los del registro fusionado.
const historiaExcluida = "first_seen = excluded.first_seen, first_source = excluded.first_source, " +
	"first_url = excluded.first_url, last_seen = excluded.last_seen, source = excluded.source, " +
	"url = excluded.url, ttl = excluded.ttl"

// Las sentencias de cada tabla. Una escritura inserta el registro o, si su clave
// ya estaba, lo sustituye por el fusionado; el tipo de un nodo y el cuerpo de un
// texto nunca cambian (FR-024), así que no se reescriben.
const (
	lecturaDeNodo   = "SELECT type, props, " + columnasDeHistoria + " FROM nodes WHERE id = ?"
	escrituraDeNodo = "INSERT INTO nodes (id, type, props, " + columnasDeHistoria + ") " +
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT (id) DO UPDATE SET props = excluded.props, " + historiaExcluida

	lecturaDeArista   = "SELECT " + columnasDeHistoria + " FROM edges WHERE src = ? AND rel = ? AND dst = ?"
	escrituraDeArista = "INSERT INTO edges (src, rel, dst, " + columnasDeHistoria + ") " +
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT (src, rel, dst) DO UPDATE SET " + historiaExcluida

	lecturaDeTexto   = "SELECT body, fetched_at, source, url FROM texts WHERE hash = ?"
	escrituraDeTexto = "INSERT INTO texts (hash, body, fetched_at, source, url) VALUES (?, ?, ?, ?, ?) " +
		"ON CONFLICT (hash) DO UPDATE SET fetched_at = excluded.fetched_at, source = excluded.source, url = excluded.url"

	// Las de la tabla lecturas (H7.1 research.md V6), y la de las versiones que
	// el grafo guarda de un bloque: las BloqueVersion a las que llega su
	// eli:has_version, con la fecha de su última observación.
	lecturaDeLecturas   = "SELECT ultima, anterior FROM lecturas WHERE bloque = ?"
	escrituraDeLecturas = "INSERT INTO lecturas (bloque, ultima, anterior) VALUES (?, ?, ?) " +
		"ON CONFLICT (bloque) DO UPDATE SET ultima = excluded.ultima, anterior = excluded.anterior"
	lecturaDeVersionesGuardadas = "SELECT n.id, n.last_seen FROM edges AS e JOIN nodes AS n ON n.id = e.dst " +
		"WHERE e.src = ? AND e.rel = ? AND n.type = ?"
)

// historiaGuardada es la historia de un nodo o una arista tal como la guarda
// world.db: la primera y la última observación, y la vigencia en su columna.
type historiaGuardada struct {
	primera, ultima grafo.Procedencia
	ttl             sql.NullInt64
}

// destinos son dónde se leen los siete campos de la historia, en su orden.
func (h *historiaGuardada) destinos() []any {
	return []any{
		&h.primera.FechaConsulta, &h.primera.Fuente, &h.primera.URL,
		&h.ultima.FechaConsulta, &h.ultima.Fuente, &h.ultima.URL, &h.ttl,
	}
}

// valoresDeHistoria son los valores de los siete campos de la historia, en su
// orden: la vigencia en segundos, o NULL si no se declaró (FR-065).
func valoresDeHistoria(primera, ultima grafo.Procedencia, vigencia time.Duration) []any {
	var ttl any
	if vigencia != 0 {
		ttl = int64(vigencia / time.Second)
	}

	return []any{primera.FechaConsulta, primera.Fuente, primera.URL, ultima.FechaConsulta, ultima.Fuente, ultima.URL, ttl}
}

// encontrada separa la fila que no está, que no es un fallo, de un fallo al
// leerla.
func encontrada(err error) (bool, error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return false, nil
	case err != nil:
		return false, err
	}

	return true, nil
}

// leerNodo lee el nodo guardado con el id del que llega.
func leerNodo(ctx context.Context, tx *sql.Tx, llegado grafo.RegistroDeNodo) (grafo.RegistroDeNodo, bool, error) {
	guardado := grafo.RegistroDeNodo{ID: llegado.ID}

	var historia historiaGuardada

	estaba, err := encontrada(tx.QueryRowContext(ctx, lecturaDeNodo, llegado.ID).
		Scan(append([]any{&guardado.Tipo, &guardado.Datos}, historia.destinos()...)...))
	if !estaba {
		return grafo.RegistroDeNodo{}, false, err
	}

	guardado.PrimeraObservacion, guardado.UltimaObservacion = historia.primera, historia.ultima
	guardado.Vigencia = vigenciaGuardada(historia.ttl)

	return guardado, true, nil
}

// escribirNodo guarda el nodo.
func escribirNodo(ctx context.Context, tx *sql.Tx, nodo grafo.RegistroDeNodo) error {
	_, err := tx.ExecContext(ctx, escrituraDeNodo, append([]any{nodo.ID, nodo.Tipo, nodo.Datos},
		valoresDeHistoria(nodo.PrimeraObservacion, nodo.UltimaObservacion, nodo.Vigencia)...)...)

	return err
}

// leerArista lee la arista guardada con la terna de la que llega. Un extremo
// que no está ni en el lote ni en el grafo no se comprueba aquí: lo rechaza la
// clave ajena de edges al escribirla, y la transacción entera se deshace
// (H7.1 FR-075; research.md V5).
func leerArista(ctx context.Context, tx *sql.Tx, llegada grafo.RegistroDeArista) (grafo.RegistroDeArista, bool, error) {
	guardada := grafo.RegistroDeArista{Origen: llegada.Origen, Relacion: llegada.Relacion, Destino: llegada.Destino}

	var historia historiaGuardada

	estaba, err := encontrada(tx.QueryRowContext(ctx, lecturaDeArista, llegada.Origen, llegada.Relacion, llegada.Destino).
		Scan(historia.destinos()...))
	if !estaba {
		return grafo.RegistroDeArista{}, false, err
	}

	guardada.PrimeraObservacion, guardada.UltimaObservacion = historia.primera, historia.ultima
	guardada.Vigencia = vigenciaGuardada(historia.ttl)

	return guardada, true, nil
}

// escribirArista guarda la arista.
func escribirArista(ctx context.Context, tx *sql.Tx, arista grafo.RegistroDeArista) error {
	_, err := tx.ExecContext(ctx, escrituraDeArista, append([]any{arista.Origen, arista.Relacion, arista.Destino},
		valoresDeHistoria(arista.PrimeraObservacion, arista.UltimaObservacion, arista.Vigencia)...)...)

	return err
}

// leerTexto lee el texto guardado con la huella del que llega.
func leerTexto(ctx context.Context, tx *sql.Tx, llegado grafo.RegistroDeTexto) (grafo.RegistroDeTexto, bool, error) {
	guardado := grafo.RegistroDeTexto{Huella: llegado.Huella}

	estaba, err := encontrada(tx.QueryRowContext(ctx, lecturaDeTexto, llegado.Huella).Scan(&guardado.Cuerpo,
		&guardado.Procedencia.FechaConsulta, &guardado.Procedencia.Fuente, &guardado.Procedencia.URL))
	if !estaba {
		return grafo.RegistroDeTexto{}, false, err
	}

	return guardado, true, nil
}

// escribirTexto guarda el texto.
func escribirTexto(ctx context.Context, tx *sql.Tx, texto grafo.RegistroDeTexto) error {
	_, err := tx.ExecContext(ctx, escrituraDeTexto, texto.Huella, texto.Cuerpo,
		texto.Procedencia.FechaConsulta, texto.Procedencia.Fuente, texto.Procedencia.URL)

	return err
}
