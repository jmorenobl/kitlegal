package cache

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

const (
	// controladorSQLite es el nombre con el que modernc.org/sqlite se registra
	// en database/sql. Es el único controlador del módulo y se escribe una vez.
	controladorSQLite = "sqlite"

	// permisosDelDirectorio y permisosDelFichero son el acceso reservado a la
	// cuenta que exige FR-021: el directorio solo para su dueño, el fichero solo
	// legible y escribible por él.
	permisosDelDirectorio = 0o700
	permisosDelFichero    = 0o600

	// conexionesPorCliente es una, y esa es la decisión: un cliente atiende una
	// invocación secuencial, con una sola conexión no hay contención dentro del
	// proceso, y cada cliente lleva su propio grupo de conexiones, que se
	// comporta como el de otro proceso (FR-032, D4).
	conexionesPorCliente = 1
)

// abre deja el cliente listo para operar, y es lo único que toca el disco al
// construirlo: en modo normal crea el directorio, hace nacer el fichero con
// acceso reservado, abre la conexión y migra; en modo de solo lectura no crea
// nada —ni el directorio, ni la base, ni sus auxiliares— y solo comprueba lo que
// hay (FR-015, D4, D5).
func (c *Cliente) abre(ctx context.Context) error {
	if c.soloLectura {
		return c.abreParaLeer(ctx)
	}

	return c.abreYMigra(ctx)
}

// abreYMigra es la apertura del modo normal (contrato de apertura §5). El orden
// de los pasos es el que hace ciertos los permisos de FR-021: el directorio
// primero, el fichero después —creado aquí, no por SQLite, que lo pondría a
// 0644— y solo entonces la conexión, de modo que los auxiliares del registro de
// escritura nazcan heredando los del fichero (D4, sonda 1 I).
func (c *Cliente) abreYMigra(ctx context.Context) error {
	if err := os.MkdirAll(c.directorio, permisosDelDirectorio); err != nil {
		return errorDeDirectorioNoEscribible("construir", string(c.origen), c.directorio, err)
	}

	if err := c.creaElFichero(); err != nil {
		return err
	}

	base, err := c.abreLaConexion(dsnNormal(c.ruta))
	if err != nil {
		return err
	}

	version, err := c.migra(ctx, base)
	if err != nil {
		return errors.Join(err, c.cierraTrasElFallo(base))
	}

	c.db = base
	c.versionEsquema = version

	c.registrador.DebugContext(ctx, "caché: base abierta",
		slog.String("ruta", c.ruta),
		slog.Int64("version_esquema", version))

	return nil
}

// creaElFichero hace nacer cache.db vacío con acceso reservado a la cuenta. Un
// fichero de cero bytes es una base de datos válida, así que crearlo no adelanta
// ningún esquema; lo que adelanta son los permisos, que es justo lo que no se
// puede pedir después (FR-021, D4).
//
// Sin O_EXCL a propósito: dos invocaciones que arrancan a la vez comparten el
// fichero y se turnan en la migración, que es lo que FR-029 pide.
//
// Si abrirlo falla, el mensaje culpa a quien toca: al fichero cuando ya existe
// —sus permisos no dejan escribirlo, o lo que hay en la ruta no es un fichero—
// y al directorio cuando el fichero no está y es el directorio el que no deja
// crearlo. Los dos son «argumentos» (2), porque una caché que no puede escribir
// no sirve y quien invoca puede declarar otro directorio (fila 6, FR-035).
func (c *Cliente) creaElFichero() error {
	fichero, err := os.OpenFile(filepath.Clean(c.ruta), os.O_RDWR|os.O_CREATE, permisosDelFichero)
	if err != nil {
		if _, existe := os.Stat(c.ruta); existe == nil {
			return errorDeFicheroNoEscribible("construir", string(c.origen), c.ruta, err)
		}

		return errorDeDirectorioNoEscribible("construir", string(c.origen), c.directorio, err)
	}

	if err := fichero.Close(); err != nil {
		return c.falloInesperadoEn("construir",
			fmt.Sprintf("no se pudo cerrar %q después de crearlo", c.ruta), err)
	}

	return nil
}

// abreParaLeer es la apertura del modo de solo lectura (contrato de apertura
// §6): no crea nada y la primera consulta es la que averigua qué hay.
func (c *Cliente) abreParaLeer(ctx context.Context) error {
	if _, err := os.Stat(c.ruta); err != nil {
		if !esInexistente(err) {
			// No se sabe si la base está —acceso denegado, un fallo de entrada y
			// salida—, y contestar «no hay nada» sería mentir (fila 12).
			return c.falloInesperadoEn("construir",
				fmt.Sprintf("no se puede saber si %q está", c.ruta), err)
		}

		// Ni la base ni su directorio están, y este modo no crea ninguno de los
		// dos: el cliente queda sin base y toda lectura será la ausencia que
		// FR-016 convierte en «fuente no disponible» (FR-015).
		c.registrador.DebugContext(ctx, "caché: sin base de datos que leer",
			slog.String("ruta", c.ruta))

		return nil
	}

	conocida, err := versionConocida()
	if err != nil {
		return errorInesperado("construir", motivoDeMigracionesIlegibles, err)
	}

	base, registrada, err := c.conexionDeLectura(ctx)
	if err != nil {
		return err
	}

	// 0 es una base sin esquema —toda lectura es ausencia— y la conocida es la
	// que este binario sabe leer. Cualquier otra sería adivinar, y aquí no se
	// migra: migrar es escribir (D7).
	if registrada != 0 && registrada != conocida {
		return errors.Join(errorDeVersionAjena("construir", c.ruta, registrada, conocida),
			c.cierraTrasElFallo(base))
	}

	c.db = base
	c.versionEsquema = registrada

	c.registrador.DebugContext(ctx, "caché: base abierta para leerla",
		slog.String("ruta", c.ruta),
		slog.Int64("version_esquema", registrada))

	return nil
}

// conexionDeLectura abre la base para leerla y devuelve la versión que su
// esquema dice tener, que es lo que la primera consulta averigua. Aquí vive la
// única rama del modo: SQLite no ha podido abrir lo que hace falta para leer.
//
// Cuando eso pasa hay dos culpables posibles y el código del controlador no los
// distingue: el propio fichero, que no se deja leer, y el directorio, que no
// admite crear la memoria compartida. Se pregunta primero por el fichero, que es
// lo más concreto, y solo si se puede leer se sigue con el directorio (FR-035):
// un cache.db sin permiso de lectura es «inesperado» (1) nombrando el fichero
// (fila 12), nunca un mensaje que culpe a un directorio que no tiene la culpa.
func (c *Cliente) conexionDeLectura(ctx context.Context) (*sql.DB, int64, error) {
	base, err := c.abreLaConexion(dsnSoloLectura(c.ruta, false))
	if err != nil {
		return nil, 0, err
	}

	registrada, err := c.leeLaVersion(ctx, base)
	if err == nil {
		return base, registrada, nil
	}

	if !esSinMemoriaCompartida(err) {
		return nil, 0, errors.Join(c.falloAlAbrir(ctx, "construir", err), c.cierraTrasElFallo(base))
	}

	if fallo := c.compruebaQueSeDejaLeer(err); fallo != nil {
		return nil, 0, errors.Join(fallo, c.cierraTrasElFallo(base))
	}

	return c.reabreInmutable(ctx, base, err)
}

// compruebaQueSeDejaLeer abre cache.db solo para leerlo y lo cierra: si ni eso se
// puede, el fichero es la causa del fallo del controlador y se dice así, con la
// ruta del fichero y el acceso denegado cuando lo es (fila 12, FR-035). Es una
// lectura que no escribe nada, así que cabe en el modo de solo lectura.
func (c *Cliente) compruebaQueSeDejaLeer(causa error) error {
	fichero, err := os.Open(filepath.Clean(c.ruta))
	if err != nil {
		return errorDeFicheroIlegible(c.ruta, errors.Join(err, causa))
	}

	if err := fichero.Close(); err != nil {
		return c.falloInesperadoEn("construir",
			fmt.Sprintf("no se pudo cerrar %q después de comprobar que se deja leer", c.ruta), err)
	}

	return nil
}

// reabreInmutable es la rama del directorio que no admite crear cache.db-shm,
// a la que se llega con el fichero ya comprobado como legible. Los dos códigos
// que produce ese caso —1544 cuando no hay registro de escritura, 14 cuando el
// registro existe y falta la memoria compartida— llevan a la misma pregunta, y
// esa pregunta no es cuál de los dos llegó, sino si hay algo en el registro
// (research D5, sonda 3 B, D y E):
//
//   - no hay registro: todo lo confirmado está en cache.db, y nadie puede estar
//     escribiendo donde no se pueden crear los auxiliares, así que el fichero no
//     va a cambiar bajo los pies del lector y se reabre como inmutable, que lee
//     sin crear nada (sonda 3 C). Esto es lo que hace que el contenedor de solo
//     lectura de US3 se lea «con normalidad» (FR-015);
//   - hay registro: SQLite no puede leerlo sin su memoria compartida, y abrirlo
//     como inmutable ignoraría lo que ese registro contiene, de modo que servir
//     una ausencia sería mentir. «Inesperado» (1) nombrando los tres ficheros
//     (fila 13 del contrato de errores).
func (c *Cliente) reabreInmutable(ctx context.Context, base *sql.DB, causa error) (*sql.DB, int64, error) {
	if err := c.cierraTrasElFallo(base); err != nil {
		return nil, 0, err
	}

	_, err := os.Stat(c.ruta + sufijoRegistroDeEscritura)

	switch {
	case err == nil:
		return nil, 0, errorDeWALSinMemoriaCompartida("construir", c.ruta, causa)
	case !esInexistente(err):
		return nil, 0, c.falloInesperadoEn("construir", fmt.Sprintf(
			"no se puede saber si %q tiene un registro de escritura", c.ruta), err)
	}

	inmutable, err := c.abreLaConexion(dsnSoloLectura(c.ruta, true))
	if err != nil {
		return nil, 0, err
	}

	registrada, err := c.leeLaVersion(ctx, inmutable)
	if err != nil {
		return nil, 0, errors.Join(c.falloAlLeerLoInmutable(ctx, err), c.cierraTrasElFallo(inmutable))
	}

	return inmutable, registrada, nil
}

// leeLaVersion es la primera consulta de toda apertura —la versión que el
// esquema dice tener— esperando el bloqueo por tramos que miran el contexto
// (FR-003, FR-031): en un fichero que todavía va en diario clásico, o mientras
// otra invocación consolida su registro, leer también puede encontrar la base
// ocupada.
func (c *Cliente) leeLaVersion(ctx context.Context, base consultante) (int64, error) {
	var registrada int64

	err := c.reintentaMientrasBloqueada(ctx, func() (err error) {
		registrada, err = versionRegistrada(ctx, base)

		return err
	})

	return registrada, err
}

// abreLaConexion abre el grupo de conexiones del cliente sobre un DSN ya
// compuesto y lo deja en una sola conexión.
func (c *Cliente) abreLaConexion(dsn string) (*sql.DB, error) {
	base, err := sql.Open(controladorSQLite, dsn)
	if err != nil {
		return nil, c.falloInesperadoEn("construir",
			fmt.Sprintf("no se pudo abrir %q", c.ruta), err)
	}

	base.SetMaxOpenConns(conexionesPorCliente)

	return base, nil
}

// dsnNormal es la cadena de conexión del modo normal (contrato de apertura §3):
// espera ante bloqueo por tramos —el motor espera un tramo y el cliente
// reintenta mirando el contexto hasta agotar los cinco segundos (FR-003,
// FR-031)—, diario en WAL —lo que permite leer mientras otra invocación
// escribe—, confirmación sincronizada con el disco y transacciones que toman el
// bloqueo de escritura al empezar, de modo que dos migraciones simultáneas se
// turnen en vez de fallar por escalada de bloqueo (FR-030, D4).
func dsnNormal(ruta string) string {
	return "file:" + rutaParaURI(ruta) +
		"?" + pragmaTramoDeEspera +
		"&_pragma=journal_mode(WAL)" +
		"&_pragma=synchronous(FULL)" +
		"&_txlock=immediate"
}

// dsnSoloLectura es la cadena del modo de solo lectura (contrato de apertura
// §3): mode=ro no escribe nada y query_only lo refuerza dentro de la propia
// conexión.
//
// Sin immutable salvo cuando se pide: un lector normal tiene que ver lo que otra
// invocación ya confirmó en el registro de escritura y respetar sus bloqueos, y
// un lector inmutable no toma ninguno. Inmutable solo cuando el directorio no
// admite los auxiliares y no hay registro que leer, que es el único caso en que
// el fichero no puede cambiar mientras se lee (clarificación Q2, D5).
func dsnSoloLectura(ruta string, inmutable bool) string {
	dsn := "file:" + rutaParaURI(ruta) + "?mode=ro"
	if inmutable {
		dsn += "&immutable=1"
	}

	return dsn + "&" + pragmaTramoDeEspera + "&_pragma=query_only(1)"
}

// pragmaTramoDeEspera es el busy_timeout de cada conexión, en milisegundos: el
// tramo que SQLite espera por su cuenta en cada intento (espera.go). El
// controlador lo aplica el primero de todos los PRAGMA al abrir cada conexión.
var pragmaTramoDeEspera = "_pragma=busy_timeout(" + strconv.FormatInt(tramoDeEspera.Milliseconds(), 10) + ")"

// rutaParaURI convierte la ruta del fichero en el tramo de camino de un URI
// «file:» de SQLite, que es lo que el controlador entrega al motor con
// SQLITE_OPEN_URI. El motor decodifica en ese camino toda secuencia %HH y lo
// corta en el primer «?» o «#», así que esos tres caracteres son los únicos
// que hay que proteger: sin protegerlos, un directorio con «?» o «#» en el
// nombre abriría la base en otro sitio y uno con «%41» en otro nombre (FR-020,
// FR-022). Las barras del sistema se convierten a «/», que es lo que el URI
// espera; todo lo demás va tal cual.
func rutaParaURI(ruta string) string {
	return escapadorDeURI.Replace(filepath.ToSlash(filepath.Clean(ruta)))
}

// escapadorDeURI escapa lo que la sintaxis de URI de SQLite interpreta dentro
// del camino: el «%» primero, para que los otros dos escapes no se vuelvan a
// escapar.
var escapadorDeURI = strings.NewReplacer("%", "%25", "?", "%3F", "#", "%23")

// cierraTrasElFallo retira la conexión que no llegó a servir para nada: una
// apertura que falla no puede dejar nada abierto. Quien llama añade lo que
// devuelva a la causa principal con errors.Join, de modo que el fallo que
// importa siga siendo el primero y el del cierre no se pierda.
func (c *Cliente) cierraTrasElFallo(base *sql.DB) error {
	if err := base.Close(); err != nil {
		return c.falloInesperadoEn("cerrar", fmt.Sprintf(
			"no se pudo cerrar %q después de un fallo al abrirla", c.ruta), err)
	}

	return nil
}

// falloAlAbrir clasifica lo que el controlador devuelve al leer la versión del
// esquema o al empezar la transacción de una migración, y no deja ninguna forma
// de fallo sin clase (FR-033). El orden de las ramas es el orden de lo que
// explica el fallo:
//
//   - el contexto terminado es «fuente no disponible» (4), la misma clase con
//     la que el kernel trata el plazo agotado (fila 15); también cuando venció
//     durante una espera ante bloqueo, en cuyo caso la causa es el SQLITE_BUSY
//     de la última espera y el contexto ya está terminado;
//   - la base bloqueada por otra invocación más tiempo que el presupuesto es
//     «inesperado» (1) diciendo justo eso (fila 14): un bloqueo no convierte el
//     fichero en inutilizable;
//   - que el directorio no admita el registro de escritura es «argumentos» (2),
//     porque una caché que no puede escribir no sirve de nada y quien invoca
//     puede declarar otro directorio (fila 6). En solo lectura este código no
//     llega aquí: lo atiende antes la rama de la memoria compartida;
//   - todo lo demás —SQLITE_NOTADB el primero— es un fichero que no sirve como
//     base de datos, y no se borra ni se rehace (fila 10, FR-028).
func (c *Cliente) falloAlAbrir(ctx context.Context, operacion string, causa error) error {
	switch {
	case terminoElContexto(ctx, causa):
		return c.falloDelContexto(ctx, operacion, causa)
	case esBloqueo(causa):
		return errorDeBloqueo(operacion, c.ruta, "", c.esperaAnteBloqueo, causa)
	case codigoDeSQLite(causa) == sqlite3.SQLITE_READONLY_DIRECTORY:
		return errorDeDirectorioNoEscribible(operacion, string(c.origen), c.directorio, causa)
	default:
		return errorDeFicheroInutilizable(operacion, c.ruta, causa)
	}
}

// falloAlLeerLoInmutable clasifica el fallo de la reapertura inmutable, que es
// el último intento de leer y llega con el fichero ya comprobado como legible y
// sin registro de escritura: si el contexto terminó, es su fila (15); si no, lo
// único honrado que queda es decir que no se puede leer y nombrar el fichero,
// sin atribuirlo a nada que no se haya comprobado (fila 12). Ninguno degrada a
// una ausencia falsa.
func (c *Cliente) falloAlLeerLoInmutable(ctx context.Context, causa error) error {
	if terminoElContexto(ctx, causa) {
		return c.falloDelContexto(ctx, "construir", causa)
	}

	return errorDeFicheroIlegible(c.ruta, causa)
}

// falloInesperadoEn es lo inesperado que ocurre sobre un fichero concreto. El
// constructor de la clase no lleva ruta —solo la llevan las situaciones que la
// nombran en su mensaje—, y aquí se rellena el campo para que quien recupere el
// error con errors.As sepa sobre qué fichero fue sin analizar ninguna cadena.
func (c *Cliente) falloInesperadoEn(operacion, motivo string, causa error) *Error {
	fallo := errorInesperado(operacion, motivo, causa)
	fallo.Ruta = c.ruta

	return fallo
}

// falloDelContexto es la fila 15 sobre esta base: el contexto de quien llama
// terminó mientras se construía o se migraba. Es «fuente no disponible» (4) y
// no «inesperado», porque el plazo agotado no es un fallo del fichero. La causa
// lleva el error del contexto además del que devolviera el controlador, para
// que errors.Is alcance los dos.
func (c *Cliente) falloDelContexto(ctx context.Context, operacion string, causa error) *Error {
	fallo := errorDeFuenteNoDisponible(operacion, fmt.Sprintf(
		"el contexto terminó antes de %s la caché en %q", operacion, c.ruta),
		conElErrorDelContexto(ctx, causa))
	fallo.Ruta = c.ruta

	return fallo
}

// esSinMemoriaCompartida dice si el fallo es uno de los dos códigos con los que
// SQLite dice que no pudo abrir lo que necesita para leer: el fichero mismo, o
// la memoria compartida en un directorio que no admite crearla. Son **dos**
// códigos y no uno, y cuál llega depende de qué auxiliares existan, así que los
// dos disparan la misma comprobación: primero si el fichero se deja leer y
// después si hay registro de escritura; lo que decide es eso, no el código (D5).
func esSinMemoriaCompartida(err error) bool {
	codigo := codigoDeSQLite(err)

	return codigo == sqlite3.SQLITE_READONLY_DIRECTORY || codigo == sqlite3.SQLITE_CANTOPEN
}

// codigoDeSQLite devuelve el código de resultado que el controlador adjunta al
// fallo, y 0 —SQLITE_OK, que ningún fallo lleva— cuando el fallo no viene de él.
func codigoDeSQLite(err error) int {
	var delControlador *sqlite.Error
	if errors.As(err, &delControlador) {
		return delControlador.Code()
	}

	return 0
}
