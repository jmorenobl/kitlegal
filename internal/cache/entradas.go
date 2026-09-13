package cache

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core"
)

// Cliente implementa el puerto del dominio, y que lo siga implementando no
// depende de que alguien lo recuerde: cambiar la firma de Get o de Put deja de
// compilar aquí, y no en el adaptador de fuente que lo use (FR-001, FR-002).
var _ core.Cache = (*Cliente)(nil)

const (
	// leeLaEntrada devuelve como mucho una fila, porque la clave es la clave
	// primaria: basta QueryRowContext y no queda ningún conjunto de filas que
	// cerrar (D8). La vigencia no se compara aquí sino en Go, para que el borde
	// quede escrito una sola vez y no dependa del motor (FR-008).
	leeLaEntrada = `SELECT contenido, expira_en FROM entradas WHERE clave = ?`

	// guardaLaEntrada es el upsert en una sola sentencia: inserta la clave o, si
	// ya está, sustituye a la vez contenido y vigencia. Así no queda rastro de la
	// anterior ni una clave repetida, y dos escrituras simultáneas de la misma
	// clave terminan con la última completa y nunca con una mezcla (FR-007, D8).
	guardaLaEntrada = `INSERT INTO entradas(clave, contenido, expira_en) VALUES (?, ?, ?) ` +
		`ON CONFLICT(clave) DO UPDATE SET contenido = excluded.contenido, expira_en = excluded.expira_en`
)

// Los tres resultados de una lectura que no falla, con el nombre con el que el
// registro los anota (D14). Para quien llama, ausencia y expirada son lo mismo
// (FR-008); el registro los separa porque es lo que explica por qué se volvió a
// pedir algo que estaba guardado.
const (
	resultadoAcierto  = "acierto"
	resultadoAusencia = "ausencia"
	resultadoExpirada = "expirada"
)

// Los dos extremos del intervalo que expira_en puede representar: la columna
// guarda nanosegundos Unix en un entero de 64 bits, y time.Time.UnixNano solo
// está definido entre ellos (del 21 de septiembre de 1677 al 11 de abril de
// 2262). Fuera de ese intervalo, la conversión da un número cualquiera.
var (
	primerInstanteRepresentable = time.Unix(0, math.MinInt64)
	ultimoInstanteRepresentable = time.Unix(0, math.MaxInt64)
)

// Get devuelve el contenido guardado bajo la clave mientras siga vigente, con el
// idioma «coma ok» del puerto: presente y vigente es (contenido, true, nil), y
// la ausencia —también la de una entrada caducada— es (nil, false, nil), el
// resultado normal que lleva a quien llama a pedirlo a la fuente (FR-013).
//
// Vigente quiere decir que el reloj del cliente va por delante del instante de
// expiración: en el instante exacto, y después, es ausencia (FR-008). Leer una
// entrada caducada no la borra; la sustituye la siguiente escritura de la misma
// clave (D6).
//
// En modo de solo lectura la ausencia no se devuelve como ausencia, sino como un
// fallo de la clase «fuente no disponible» (4) que nombra la clave: sin red no
// hay a dónde ir a buscar lo que falta, y así ningún adaptador tiene nada que
// decidir bajo --offline (FR-016, FR-018). En ese modo es ausencia también la
// base sin esquema y el cliente que no encontró ninguna base (FR-015).
//
// Lo presente se devuelve byte a byte, y un contenido de cero bytes llega como
// un []byte vacío y nunca nulo, para que «presente y vacío» no se confunda con
// la ausencia (FR-012).
//
// Antes de tocar el disco, y en este orden: la clave vacía es «argumentos» (2),
// la llamada después de Close es «inesperado» (1) y el contexto cancelado o
// vencido es «fuente no disponible» (4) (FR-003, FR-004, FR-011, D8).
func (c *Cliente) Get(ctx context.Context, clave string) ([]byte, bool, error) {
	if clave == "" {
		return nil, false, errorDeClaveVacia("leer")
	}

	if c.estaCerrado() {
		return nil, false, errorTrasCierre("leer", c.ruta, clave)
	}

	if err := ctx.Err(); err != nil {
		return nil, false, c.falloAlOperar(ctx, "leer", clave, err)
	}

	// Sin base o sin esquema no hay tabla que consultar, y preguntárselo al
	// fichero convertiría en fallo lo que es una ausencia. Solo ocurre en solo
	// lectura, que no crea ni migra nada (FR-015).
	if c.db == nil || c.versionEsquema == 0 {
		return c.ausente(ctx, clave, resultadoAusencia)
	}

	var (
		contenido []byte
		expiraEn  int64
	)

	// Leer con el diario en WAL no espera a ningún escritor, pero una base que
	// todavía va en diario clásico, o que otra invocación está consolidando,
	// puede contestar que está ocupada: se espera por tramos que miran el
	// contexto, como en toda operación (FR-003, FR-031).
	err := c.reintentaMientrasBloqueada(ctx, func() error {
		return c.db.QueryRowContext(ctx, leeLaEntrada, clave).Scan(&contenido, &expiraEn)
	})

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return c.ausente(ctx, clave, resultadoAusencia)
	case err != nil:
		return nil, false, c.falloAlOperar(ctx, "leer", clave, err)
	case !c.reloj().Before(time.Unix(0, expiraEn)):
		return c.ausente(ctx, clave, resultadoExpirada)
	}

	// El controlador entrega un BLOB de cero bytes como nil (research D8, sonda
	// 1 C), que aquí vuelve a ser lo que se guardó: un contenido presente y vacío.
	if contenido == nil {
		contenido = []byte{}
	}

	c.anotaLaLectura(ctx, clave, resultadoAcierto)

	return contenido, true, nil
}

// Put guarda el contenido bajo la clave con la vigencia dada, sustituyendo por
// completo —contenido y vigencia— lo que hubiera (FR-006, FR-007). El instante
// de expiración sale del reloj del cliente al escribir, y un contenido nulo se
// guarda como uno de cero bytes, que es lo que la columna admite y lo que Get
// devolverá (FR-012).
//
// Toda vigencia mayor que cero se acepta, por larga que sea (FR-010). Si lleva
// la expiración más allá del último instante que expira_en puede representar
// —el 11 de abril de 2262—, se guarda ese último instante, y la entrada es
// vigente hasta él: para cualquier reloj anterior, Get decide igual que si el
// instante calculado se hubiera podido guardar (instanteDeExpiracion).
//
// Antes de tocar el disco, y en este orden (D8): la clave vacía y la vigencia
// menor o igual que cero son «argumentos» (2), porque «no guardar esto» se
// expresa no escribiendo (FR-010, FR-011); escribir en una caché de solo lectura
// es «inesperado» (1) nombrando la clave, también en el cliente que no tiene
// base (FR-017); la llamada después de Close es «inesperado» (1); y el contexto
// cancelado o vencido es «fuente no disponible» (4).
func (c *Cliente) Put(ctx context.Context, clave string, contenido []byte, vigencia time.Duration) error {
	if clave == "" {
		return errorDeClaveVacia("escribir")
	}

	if vigencia <= 0 {
		return errorDeVigenciaInvalida(clave, vigencia)
	}

	if c.soloLectura {
		return errorDeEscrituraEnSoloLectura(c.ruta, clave)
	}

	if c.estaCerrado() {
		return errorTrasCierre("escribir", c.ruta, clave)
	}

	if err := ctx.Err(); err != nil {
		return c.falloAlOperar(ctx, "escribir", clave, err)
	}

	if contenido == nil {
		contenido = []byte{}
	}

	expiraEn := instanteDeExpiracion(c.reloj(), vigencia)

	// El upsert es una sentencia en autocommit: si otra invocación tiene el
	// bloqueo de escritura, SQLITE_BUSY dice que no se escribió nada y se
	// vuelve a intentar, por tramos que miran el contexto, hasta agotar la
	// espera (FR-003, FR-031).
	err := c.reintentaMientrasBloqueada(ctx, func() error {
		_, err := c.db.ExecContext(ctx, guardaLaEntrada, clave, contenido, expiraEn)

		return err
	})
	if err != nil {
		return c.falloAlOperar(ctx, "escribir", clave, err)
	}

	c.registrador.DebugContext(ctx, "caché: entrada guardada",
		slog.String("ruta", c.ruta),
		slog.String("clave", clave))

	return nil
}

// instanteDeExpiracion es lo que Put escribe en expira_en: el instante en que
// la entrada caduca, ahora más la vigencia, en nanosegundos Unix (FR-006).
//
// La suma es un time.Time y no desborda, pero UnixNano solo está definido
// dentro del intervalo representable: una vigencia larga y válida —FR-010 solo
// rechaza la menor o igual que cero— puede llevar la expiración más allá de
// 2262, y convertirla sin más daría un número cualquiera, con el que Get leería
// la entrada recién guardada como caducada, o como vigente para siempre. Por eso
// un instante fuera del intervalo se satura al extremo más cercano: posterior
// al último representable, se guarda el último; anterior al primero —solo
// alcanzable con un reloj inyectado que viva antes de 1678—, el primero. Para
// cualquier reloj dentro del intervalo la comparación de Get da lo mismo que
// daría con el instante exacto: sigue vigente hasta el último instante lo que
// caducaría después de él, y ya ha caducado lo que caducó antes del primero
// (FR-008). No es una situación de fallo y no añade ninguna fila a la tabla
// cerrada del contrato de errores.
func instanteDeExpiracion(ahora time.Time, vigencia time.Duration) int64 {
	expira := ahora.Add(vigencia)

	switch {
	case !expira.Before(ultimoInstanteRepresentable):
		return math.MaxInt64
	case !expira.After(primerInstanteRepresentable):
		return math.MinInt64
	}

	return expira.UnixNano()
}

// ausente es el final de toda lectura que no encuentra una entrada vigente, y el
// único sitio donde el modo decide qué es eso: fuera de solo lectura, el
// resultado normal; en solo lectura, el fallo de la fila 8 del contrato de
// errores (FR-013, FR-016).
func (c *Cliente) ausente(ctx context.Context, clave, resultado string) ([]byte, bool, error) {
	c.anotaLaLectura(ctx, clave, resultado)

	if c.soloLectura {
		return nil, false, errorDeAusenciaEnSoloLectura(c.ruta, clave)
	}

	return nil, false, nil
}

// anotaLaLectura deja en el registro qué resultado tuvo una lectura (D14).
func (c *Cliente) anotaLaLectura(ctx context.Context, clave, resultado string) {
	c.registrador.DebugContext(ctx, "caché: lectura",
		slog.String("ruta", c.ruta),
		slog.String("clave", clave),
		slog.String("resultado", resultado))
}

// estaCerrado lee el estado de cierre con el cerrojo que lo protege, y lo suelta
// antes de operar: el cerrojo protege ese estado y nada más (D10). Un Close que
// llegue mientras una operación está en curso la hace fallar en el controlador,
// y ese fallo es «inesperado» como cualquier otro (fila 14).
func (c *Cliente) estaCerrado() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.cerrado
}

// falloAlOperar clasifica lo que sale mal al leer o escribir una entrada, sin
// dejar ninguna forma de fallo sin clase (FR-033). Si el contexto de quien llama
// terminó —antes de empezar, mientras el controlador trabajaba y lo interrumpió
// con su propio código, o durante la espera ante un bloqueo—, es «fuente no
// disponible» (4), la clase con la que el kernel trata el plazo agotado (fila
// 15). La base bloqueada por otra invocación más tiempo que la espera es
// «inesperado» (1) diciendo justo eso, y cualquier otra cosa del controlador o
// del sistema de ficheros es «inesperado» (1) con la causa envuelta (fila 14).
// Todos nombran la operación y la clave, y llevan la ruta.
func (c *Cliente) falloAlOperar(ctx context.Context, operacion, clave string, causa error) *Error {
	var fallo *Error

	switch {
	case terminoElContexto(ctx, causa):
		fallo = errorDeFuenteNoDisponible(operacion, fmt.Sprintf(
			"el contexto terminó antes de %s %q", operacion, clave), conElErrorDelContexto(ctx, causa))
	case esBloqueo(causa):
		fallo = errorDeBloqueo(operacion, c.ruta, clave, c.esperaAnteBloqueo, causa)
	default:
		fallo = errorInesperado(operacion, fmt.Sprintf(
			"no se pudo %s %q en %q", operacion, clave, c.ruta), causa)
	}

	fallo.Ruta = c.ruta
	fallo.Clave = clave

	return fallo
}

// errorDeClaveVacia es la fila 1 de la tabla del contrato de errores §3: la
// clave no lleva nada, y una caché que no interpreta la clave no tiene con qué
// sustituirla. El mensaje nombra la operación, que es lo que sitúa la llamada
// (FR-011).
func errorDeClaveVacia(operacion string) *Error {
	return errorDeArgumentos(operacion,
		fmt.Sprintf("no se puede %s una entrada con la clave vacía", operacion), nil)
}

// errorDeVigenciaInvalida es la fila 2: la vigencia no es mayor que cero. El
// mensaje nombra la vigencia recibida y la clave, y no se escribe nada (FR-010).
func errorDeVigenciaInvalida(clave string, vigencia time.Duration) *Error {
	fallo := errorDeArgumentos("escribir", fmt.Sprintf(
		"la vigencia para escribir %q tiene que ser mayor que cero y es %s", clave, vigencia), nil)
	fallo.Clave = clave

	return fallo
}
