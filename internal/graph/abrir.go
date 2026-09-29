package graph

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	// controladorSQLite es el nombre con el que modernc.org/sqlite se registra
	// en database/sql.
	controladorSQLite = "sqlite"

	// conexionesPorOperacion es una: cada lectura y cada entrega abre su propio
	// grupo de conexiones y lo deja en una, que se comporta como la conexión de
	// otra invocación y no compite con ninguna otra dentro del proceso
	// (research.md D9).
	conexionesPorOperacion = 1

	// sufijoWAL es el sufijo con el que SQLite nombra el registro de escritura
	// (WAL) de world.db.
	sufijoWAL = "-wal"

	// pragmaSoloConsultas impide dentro de la conexión cualquier escritura de
	// SQL, sea cual sea el modo con el que se abrió.
	pragmaSoloConsultas = "_pragma=query_only(1)"
)

// modoDeLectura es cómo se abre world.db para leerlo, según haya o no un
// world.db-wal junto a él (contracts/almacen-world-db.md §3, paso 2; H7.1
// research.md D17). Los dos leen todo lo confirmado y ninguno escribe SQL en
// world.db; difieren en lo que dejan en su directorio.
type modoDeLectura int

const (
	// modoSinWAL es el estado normal, sin -wal: mode=rw con query_only lee sin
	// crear el fichero, sin poder escribirlo y sin dejar ningún auxiliar al
	// cerrar (H7.1 research.md V8).
	modoSinWAL modoDeLectura = iota + 1
	// modoConWAL es world.db con su -wal, el de una escritura propia
	// interrumpida o el de otra invocación abierta: mode=ro lee también lo
	// confirmado en el WAL y deja world.db y el -wal como estaban (research.md
	// V36 D y F de H7).
	modoConWAL
)

// cadenaDeLectura es la cadena de conexión de cada modo de lectura, carácter a
// carácter la de contracts/almacen-world-db.md §3, paso 2. Un modo que no es de
// los dos recibe la de modoConWAL, la que nunca hace un checkpoint al cerrar.
func cadenaDeLectura(ruta string, modo modoDeLectura) string {
	camino := "file:" + rutaParaURI(ruta)

	if modo == modoSinWAL {
		return camino + "?mode=rw&" + pragmaDelTramo + "&" + pragmaSoloConsultas
	}

	return camino + "?mode=ro&" + pragmaDelTramo + "&" + pragmaSoloConsultas
}

// rutaParaURI convierte la ruta del fichero en el camino de un URI «file:» de
// SQLite. El motor decodifica en ese camino toda secuencia %HH y lo corta en el
// primer «?» o «#», así que esos tres caracteres son los únicos que hay que
// proteger: sin protegerlos, un directorio con «?» o «#» en el nombre abriría
// otra base, y uno con «%41», otro nombre. Las barras del sistema pasan a «/».
func rutaParaURI(ruta string) string {
	return escapadorDeURI.Replace(filepath.ToSlash(filepath.Clean(ruta)))
}

// escapadorDeURI escapa lo que la sintaxis de URI de SQLite interpreta dentro
// del camino: el «%» primero, para que los otros dos escapes no se vuelvan a
// escapar.
var escapadorDeURI = strings.NewReplacer("%", "%25", "?", "%3F", "#", "%23")

// apertura es lo que la lectura decide antes de abrir SQLite: o que el grafo
// es el vacío y no hay nada que abrir, o el modo con que se abre.
type apertura struct {
	// vacia dice que se lee el grafo vacío sin abrir nada.
	vacia bool
	// modo es el de la cadena de conexión; sin valor si vacia.
	modo modoDeLectura
}

// decidirApertura hace los pasos 1 y 2 de la lectura
// (contracts/almacen-world-db.md §3) sin abrir SQLite ni crear nada:
//
//   - world.db que no existe —también bajo un componente que no es
//     directorio— es el grafo vacío (H7 FR-004);
//   - un fichero de 0 bytes es una base sin esquema, y se lee como el grafo
//     vacío sin abrir SQLite: abrirlo borraría el -wal no vacío que deja junto
//     a él una escritura propia interrumpida (research.md V49 de H7);
//   - en otro caso, con world.db-wal junto a él, el modo que lee lo confirmado
//     en el WAL, y sin él, el que no deja ningún auxiliar.
//
// Cualquier otro fallo al mirar si están es «no se pudo leer»: suponer que no
// están podría llevar a un modo que no leyera lo confirmado.
func decidirApertura(ruta string) (apertura, error) {
	info, err := os.Stat(ruta)

	switch {
	case ausente(err):
		return apertura{vacia: true}, nil
	case err != nil:
		return apertura{}, errorDeEntradaSalida(operacionLeer, ruta, err)
	case info.Size() == 0:
		return apertura{vacia: true}, nil
	}

	_, err = os.Stat(ruta + sufijoWAL)

	switch {
	case err == nil:
		return apertura{modo: modoConWAL}, nil
	case ausente(err):
		return apertura{modo: modoSinWAL}, nil
	default:
		return apertura{}, errorDeEntradaSalida(operacionLeer, ruta, err)
	}
}

// abrirParaLeer hace el paso 3 de la lectura (contracts/almacen-world-db.md
// §3): abre world.db con el modo decidido y lee la versión de su esquema,
// esperando por tramos mientras otra invocación lo retiene (H7 FR-014). Con
// cualquier versión de la 1 a la que este binario conoce, devuelve la conexión
// abierta y la versión leída: la 1, la que escribe H7, se lee sin migrarla
// —la lectura no escribe— y sin la tabla lecturas (H7.1 research.md D4). Sin
// esquema —versión 0—, la cierra sin consultar ninguna tabla y devuelve nil, que
// es el grafo vacío (H7 FR-004); con una posterior, «de otra versión» sin
// tocarlo (H7 FR-012). Cualquier otro resultado es la regla genérica:
// inutilizable, con la ruta y la causa (H7.1 FR-070).
func abrirParaLeer(ctx context.Context, ruta string, modo modoDeLectura) (*sql.DB, int64, error) {
	conocida, err := versionConocida()
	if err != nil {
		fallo := errorDeMigracionesIlegibles(ruta, err)
		fallo.Operacion = operacionLeer

		return nil, 0, fallo
	}

	base, err := abrirConexion(operacionLeer, ruta, cadenaDeLectura(ruta, modo))
	if err != nil {
		return nil, 0, err
	}

	version, err := leerVersion(ctx, base)
	if err != nil {
		return nil, 0, errors.Join(falloAlLeer(ctx, ruta, err), cerrarTrasElFallo(base, operacionLeer, ruta))
	}

	switch {
	case version == 0:
		return nil, 0, cerrarTrasElFallo(base, operacionLeer, ruta)
	case version > conocida:
		return nil, 0, errors.Join(errorDeVersionPosterior(operacionLeer, ruta, version, conocida),
			cerrarTrasElFallo(base, operacionLeer, ruta))
	case version > 0:
		return base, version, nil
	default:
		return nil, 0, errors.Join(errorInutilizable(operacionLeer, ruta, nil),
			cerrarTrasElFallo(base, operacionLeer, ruta))
	}
}

// falloAlLeer clasifica el fallo de la primera lectura de world.db, la de su
// versión (contracts/almacen-world-db.md §3, paso 3, y §6): el contexto
// terminado es el plazo agotado, y SQLITE_BUSY con el contexto vivo, la espera
// propia agotada (H7 FR-014); cualquier otro es la regla genérica, un world.db
// que el binario no puede usar, con la causa (H7.1 FR-070).
func falloAlLeer(ctx context.Context, ruta string, causa error) error {
	if deLaEspera := esperaDeLaInvocacion().falloDeLaEspera(ctx, operacionLeer, ruta, causa); deLaEspera != nil {
		return deLaEspera
	}

	return errorInutilizable(operacionLeer, ruta, causa)
}

// abrirConexion abre el grupo de conexiones sobre una cadena ya compuesta y lo
// deja en una sola conexión. database/sql no abre nada todavía: la primera
// consulta es la que abre el fichero.
func abrirConexion(operacion, ruta, cadena string) (*sql.DB, error) {
	base, err := sql.Open(controladorSQLite, cadena)
	if err != nil {
		return nil, errorDeEntradaSalida(operacion, ruta, err)
	}

	base.SetMaxOpenConns(conexionesPorOperacion)

	return base, nil
}

// leerVersion es la primera consulta de toda apertura —la versión que el
// esquema dice tener—, esperando por tramos que miran el contexto mientras
// otra invocación retiene world.db (FR-014). Devuelve el error tal cual; lo
// clasifica quien llama.
func leerVersion(ctx context.Context, base consultante) (int64, error) {
	var version int64

	err := esperaDeLaInvocacion().reintentar(ctx, func() (err error) {
		version, err = versionRegistrada(ctx, base)

		return err
	})

	return version, err
}

// cerrarTrasElFallo cierra la conexión que no llegó a servir. Quien llama une
// lo que devuelva a la causa principal con errors.Join, de modo que el fallo
// que importa siga siendo el primero y el del cierre no se pierda.
func cerrarTrasElFallo(base *sql.DB, operacion, ruta string) error {
	if err := base.Close(); err != nil {
		return errorDeEntradaSalida(operacion, ruta, err)
	}

	return nil
}

// ausente dice si el fallo de mirar un fichero es que no está: no existe, o un
// componente de su ruta no es un directorio, que en Unix es syscall.ENOTDIR y
// no fs.ErrNotExist (en Windows ese caso ya es fs.ErrNotExist).
func ausente(err error) bool {
	return errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR)
}
