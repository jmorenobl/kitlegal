package cache

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
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

	err := c.db.QueryRowContext(ctx, leeLaEntrada, clave).Scan(&contenido, &expiraEn)

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

	expiraEn := c.reloj().Add(vigencia).UnixNano()

	if _, err := c.db.ExecContext(ctx, guardaLaEntrada, clave, contenido, expiraEn); err != nil {
		return c.falloAlOperar(ctx, "escribir", clave, err)
	}

	c.registrador.DebugContext(ctx, "caché: entrada guardada",
		slog.String("ruta", c.ruta),
		slog.String("clave", clave))

	return nil
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
// terminó —antes de empezar, o mientras el controlador trabajaba y lo
// interrumpió con su propio código—, es «fuente no disponible» (4), la clase con
// la que el kernel trata el plazo agotado (fila 15). Cualquier otra cosa del
// controlador o del sistema de ficheros es «inesperado» (1) con la causa
// envuelta (fila 14). Los dos nombran la operación y la clave, y llevan la ruta.
func (c *Cliente) falloAlOperar(ctx context.Context, operacion, clave string, causa error) *Error {
	var fallo *Error

	if esDelContexto(causa) || ctx.Err() != nil {
		fallo = errorDeFuenteNoDisponible(operacion, fmt.Sprintf(
			"el contexto terminó antes de %s %q", operacion, clave), causa)
	} else {
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
