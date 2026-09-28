package graph

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	sqlite3 "modernc.org/sqlite/lib"
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

	// Los sufijos con los que SQLite nombra los auxiliares de world.db: el
	// registro de escritura (WAL), su índice en memoria compartida y el diario
	// de un escritor en modo rollback.
	sufijoWAL               = "-wal"
	sufijoMemoriaCompartida = "-shm"
	sufijoDiario            = "-journal"

	// sistemaWindows es el valor de runtime.GOOS en el que SQLite nombra los
	// auxiliares sin seguir los enlaces simbólicos (research.md V47).
	sistemaWindows = "windows"

	// pragmaSoloConsultas impide dentro de la conexión cualquier escritura de
	// SQL, sea cual sea el modo con el que se abrió.
	pragmaSoloConsultas = "_pragma=query_only(1)"
)

// modoDeLectura es cómo se abre world.db para leerlo, según sus auxiliares y
// según que el proceso pueda escribirlo (contracts/almacen-world-db.md §3, paso
// 4; research.md D10). Los tres leen lo mismo —todo lo confirmado— y ninguno
// escribe en world.db; difieren en lo que dejan en su directorio.
type modoDeLectura int

const (
	// modoSinAuxiliares es el estado normal: el proceso puede escribir
	// world.db y no hay ningún auxiliar. mode=rw con query_only lee sin crear
	// el fichero, sin poder escribirlo y sin dejar ningún auxiliar al cerrar
	// (research.md V9, V36 E).
	modoSinAuxiliares modoDeLectura = iota + 1
	// modoInmutable es world.db que el proceso no puede escribir, sin -wal ni
	// -journal: nada confirmado vive fuera de world.db, y mode=ro con
	// immutable=1 lo lee tal cual sin crear ningún fichero (research.md V46).
	modoInmutable
	// modoConAuxiliares es cualquier otro caso: algún auxiliar con permiso de
	// escritura, o -wal o -journal sin él. mode=ro lee también lo confirmado en
	// el WAL, respeta los bloqueos de otra conexión y deja world.db, el -wal y
	// el diario como estaban (research.md V36 D y F, V43, V45 c).
	modoConAuxiliares
)

// cadenaDeLectura es la cadena de conexión de cada modo de lectura, carácter a
// carácter la de contracts/almacen-world-db.md §3, paso 4. Un modo que no es de
// los tres recibe la de modoConAuxiliares, la que nunca escribe world.db ni hace
// un checkpoint al cerrar.
//
// El modo inmutable no lleva busy_timeout: no toma bloqueos, así que nunca
// espera a nadie.
func cadenaDeLectura(ruta string, modo modoDeLectura) string {
	camino := "file:" + rutaParaURI(ruta)

	switch modo {
	case modoSinAuxiliares:
		return camino + "?mode=rw&" + pragmaDelTramo + "&" + pragmaSoloConsultas
	case modoInmutable:
		return camino + "?mode=ro&immutable=1&" + pragmaSoloConsultas
	default:
		return camino + "?mode=ro&" + pragmaDelTramo + "&" + pragmaSoloConsultas
	}
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

// rutaDeLosAuxiliares es la ruta a partir de la cual SQLite nombra los
// auxiliares de world.db (contracts/almacen-world-db.md §3, paso 2; research.md
// V47): fuera de Windows resuelve cada componente de la ruta, enlaces
// simbólicos incluidos, y los crea junto al destino y con su nombre; en Windows
// no sigue los enlaces y los nombra por la ruta tal cual. Sin enlaces, las dos
// son la misma. El sistema es un argumento para poder fijar las dos ramas desde
// cualquiera; quien lee pasa runtime.GOOS.
//
// Un enlace sin destino da el error de filepath.EvalSymlinks, que es un fichero
// ausente.
func rutaDeLosAuxiliares(ruta, sistema string) (string, error) {
	if sistema == sistemaWindows {
		return ruta, nil
	}

	return filepath.EvalSymlinks(ruta)
}

// auxiliares son los ficheros que SQLite deja junto a world.db y que están
// ahí: el -wal y el -shm de una conexión en WAL —abierta, o interrumpida sin
// cerrar— y el diario de un escritor en modo rollback.
type auxiliares struct {
	wal, memoriaCompartida, diario bool
}

// alguno dice si hay alguno de los tres.
func (a auxiliares) alguno() bool {
	return a.wal || a.memoriaCompartida || a.diario
}

// buscarAuxiliares mira cuáles de los tres existen junto a la ruta con la que
// SQLite los nombra. Uno de 0 bytes es un fichero que existe. Un fallo que no
// dice si un auxiliar está —distinto de que no exista— se devuelve tal cual:
// suponer que no está podría llevar a un modo que escribiera world.db.
func buscarAuxiliares(base string) (auxiliares, error) {
	var encontrados auxiliares

	for _, cada := range []struct {
		sufijo   string
		presente *bool
	}{
		{sufijo: sufijoWAL, presente: &encontrados.wal},
		{sufijo: sufijoMemoriaCompartida, presente: &encontrados.memoriaCompartida},
		{sufijo: sufijoDiario, presente: &encontrados.diario},
	} {
		_, err := os.Stat(base + cada.sufijo)

		switch {
		case err == nil:
			*cada.presente = true
		case !ausente(err):
			return auxiliares{}, err
		}
	}

	return encontrados, nil
}

// comprobarEscritura dice si el proceso puede escribir world.db: lo abre para
// leer y escribir y lo cierra, sin leer, escribir ni truncar, lo que no cambia
// sus bytes ni su fecha de modificación (contracts/almacen-world-db.md §3, paso
// 3, y §4, paso 4; research.md V46). Es lo mismo que mirará SQLite al abrirlo,
// con el usuario efectivo, y en Windows ve el atributo de solo lectura.
//
// Devuelve el error de abrirlo o de cerrarlo tal cual, y quien llama lo
// clasifica: nil, puede; fs.ErrNotExist —lo retiraron, o es un enlace sin
// destino—, ya no está; cualquier otro, no puede.
func comprobarEscritura(ruta string) error {
	fichero, err := os.OpenFile(filepath.Clean(ruta), os.O_RDWR, 0)
	if err != nil {
		return err
	}

	return fichero.Close()
}

// comprobarLectura dice si el proceso puede leer world.db: lo abre para leer y
// lo cierra. Separa un fichero que no se deja leer de un directorio que no
// admite los auxiliares cuando SQLite contesta con el mismo código a los dos.
func comprobarLectura(ruta string) error {
	fichero, err := os.Open(filepath.Clean(ruta))
	if err != nil {
		return err
	}

	return fichero.Close()
}

// modoPara elige el modo de lectura por lo que se encontró (research.md D10):
// con permiso de escritura y sin ningún auxiliar, el que no deja ninguno; sin
// permiso y sin -wal ni -journal, inmutable —un -shm suelto no guarda nada
// confirmado—; en cualquier otro caso, solo lectura.
func modoPara(puedeEscribir bool, encontrados auxiliares) modoDeLectura {
	switch {
	case puedeEscribir && !encontrados.alguno():
		return modoSinAuxiliares
	case !puedeEscribir && !encontrados.wal && !encontrados.diario:
		return modoInmutable
	default:
		return modoConAuxiliares
	}
}

// apertura es lo que la lectura decide antes de abrir SQLite: o que el grafo
// es el vacío y no hay nada que abrir, o el modo con que se abre y los
// auxiliares que había.
type apertura struct {
	// vacia dice que se lee el grafo vacío sin abrir nada.
	vacia bool
	// modo es el de la cadena de conexión; sin valor si vacia.
	modo modoDeLectura
	// auxiliares son los que había junto a la ruta con que SQLite los nombra.
	auxiliares auxiliares
}

// decidirApertura hace los pasos 1 a 4 de la lectura
// (contracts/almacen-world-db.md §3) sin abrir SQLite ni crear nada:
//
//   - world.db que no existe —también bajo un componente que no es directorio
//     o como enlace sin destino— es el grafo vacío (FR-004);
//   - un directorio es inutilizable (FR-010), y cualquier otro fallo al
//     mirarlo, inesperado;
//   - un fichero de 0 bytes es siempre una base sin esquema, y se lee como el
//     grafo vacío sin abrir SQLite, haya los auxiliares que haya: abrirlo en
//     cualquier modo que no sea inmutable borraría un -wal no vacío junto a él
//     (research.md V49);
//   - en otro caso, el modo por los auxiliares, buscados donde SQLite los
//     nombra, y por el permiso de escritura. Si world.db desaparece entre medias,
//     es el grafo vacío.
func decidirApertura(ruta string) (apertura, error) {
	info, err := os.Stat(ruta)

	switch {
	case ausente(err):
		return apertura{vacia: true}, nil
	case err != nil:
		return apertura{}, errorDeEntradaSalida(operacionLeer, ruta, err)
	case info.IsDir():
		return apertura{}, errorEsDirectorio(operacionLeer, ruta)
	case info.Size() == 0:
		return apertura{vacia: true}, nil
	}

	base, err := rutaDeLosAuxiliares(ruta, runtime.GOOS)

	switch {
	case ausente(err):
		return apertura{vacia: true}, nil
	case err != nil:
		return apertura{}, errorDeEntradaSalida(operacionLeer, ruta, err)
	}

	encontrados, err := buscarAuxiliares(base)
	if err != nil {
		return apertura{}, errorDeEntradaSalida(operacionLeer, ruta, err)
	}

	escritura := comprobarEscritura(ruta)
	if ausente(escritura) {
		return apertura{vacia: true}, nil
	}

	return apertura{modo: modoPara(escritura == nil, encontrados), auxiliares: encontrados}, nil
}

// abrirParaLeer hace el paso 5 de la lectura (contracts/almacen-world-db.md
// §3): abre world.db con el modo decidido y lee la versión de su esquema. Con
// la versión que este binario conoce, devuelve la conexión abierta; sin
// esquema —versión 0—, la cierra sin consultar ninguna tabla y devuelve nil,
// que es el grafo vacío (FR-004); con una posterior, inutilizable «de otra
// versión» sin tocarlo (FR-012). Una versión negativa no es de ningún esquema:
// inutilizable.
func abrirParaLeer(ctx context.Context, ruta string, decidida apertura) (*sql.DB, error) {
	conocida, err := versionConocida()
	if err != nil {
		fallo := errorDeMigracionesIlegibles(ruta, err)
		fallo.Operacion = operacionLeer

		return nil, fallo
	}

	base, version, err := conexionDeLectura(ctx, ruta, decidida)
	if err != nil {
		return nil, err
	}

	switch {
	case version == conocida:
		return base, nil
	case version == 0:
		return nil, cerrarTrasElFallo(base, operacionLeer, ruta)
	case version > conocida:
		return nil, errors.Join(errorDeVersionPosterior(operacionLeer, ruta, version, conocida),
			cerrarTrasElFallo(base, operacionLeer, ruta))
	default:
		return nil, errors.Join(errorInutilizable(operacionLeer, ruta, nil),
			cerrarTrasElFallo(base, operacionLeer, ruta))
	}
}

// conexionDeLectura abre la conexión con el modo decidido y lee la versión. Si
// SQLite no puede abrir lo que necesita para leer y el recurso inmutable cabe
// —falloAlLeer lo decide—, cierra y reabre inmutable (research.md D10, V9).
func conexionDeLectura(ctx context.Context, ruta string, decidida apertura) (*sql.DB, int64, error) {
	base, err := abrirConexion(operacionLeer, ruta, cadenaDeLectura(ruta, decidida.modo))
	if err != nil {
		return nil, 0, err
	}

	version, err := leerVersion(ctx, base)
	if err == nil {
		return base, version, nil
	}

	reabrir, fallo := falloAlLeer(ctx, ruta, decidida, err)
	cierre := cerrarTrasElFallo(base, operacionLeer, ruta)

	if !reabrir || cierre != nil {
		return nil, 0, errors.Join(fallo, cierre)
	}

	inmutable := apertura{modo: modoInmutable}

	base, err = abrirConexion(operacionLeer, ruta, cadenaDeLectura(ruta, inmutable.modo))
	if err != nil {
		return nil, 0, err
	}

	version, err = leerVersion(ctx, base)
	if err != nil {
		_, fallo = falloAlLeer(ctx, ruta, inmutable, err)

		return nil, 0, errors.Join(fallo, cerrarTrasElFallo(base, operacionLeer, ruta))
	}

	return base, version, nil
}

// falloAlLeer clasifica el fallo de la primera lectura de world.db, la de su
// versión (contracts/almacen-world-db.md §3, paso 5, y §6), y dice si cabe
// reabrirlo inmutable. En el orden de lo que explica el fallo:
//
//   - el contexto terminado es el plazo agotado, y SQLITE_BUSY con el contexto
//     vivo, la espera propia agotada (FR-014);
//   - 776, SQLITE_READONLY_ROLLBACK, es un diario caliente que el modo de solo
//     lectura no puede deshacer: inutilizable «con una transacción interrumpida
//     sin deshacer», sin probar otro modo, porque deshacerla sería escribir
//     world.db e inmutable leería páginas sin confirmar (research.md V43);
//   - 1544 o 14 son SQLite que no puede abrir lo que necesita para leer: el
//     propio fichero, o los auxiliares en un directorio que no los admite. Si el
//     fichero no se deja leer, es él el inutilizable; si se deja y no hay -wal
//     ni -journal, se reabre inmutable (research.md V9), salvo que ya lo
//     estuviera; con alguno, inutilizable, porque inmutable no leería lo
//     confirmado en el WAL ni desharía un diario (research.md V36 D, V43);
//   - cualquier otro —26 el primero, 11 de una base dañada— es un fichero que no
//     sirve como base de datos (FR-010).
func falloAlLeer(ctx context.Context, ruta string, decidida apertura, causa error) (reabrir bool, fallo error) {
	if deLaEspera := esperaDeLaInvocacion().falloDeLaEspera(ctx, operacionLeer, ruta, causa); deLaEspera != nil {
		return false, deLaEspera
	}

	switch codigoDeSQLite(causa) {
	case sqlite3.SQLITE_READONLY_ROLLBACK:
		return false, errorDeTransaccionInterrumpida(ruta, causa)
	case sqlite3.SQLITE_READONLY_DIRECTORY, sqlite3.SQLITE_CANTOPEN:
	default:
		return false, errorInutilizable(operacionLeer, ruta, causa)
	}

	if err := comprobarLectura(ruta); err != nil {
		return false, errorInutilizable(operacionLeer, ruta, errors.Join(err, causa))
	}

	if decidida.modo == modoInmutable || decidida.auxiliares.wal || decidida.auxiliares.diario {
		return false, errorInutilizable(operacionLeer, ruta, causa)
	}

	return true, nil
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
