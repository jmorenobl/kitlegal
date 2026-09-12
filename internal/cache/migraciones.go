package cache

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path"
	"time"
)

// migraciones son los ficheros del esquema que este binario trae dentro de sí.
// Van embebidos en el ejecutable a propósito: así no hay ninguna instalación que
// preparar ni ningún fichero que encontrar en la máquina, y la versión del
// esquema que el binario conoce viaja con él (D7).
//
//go:embed migraciones/*.sql
var migraciones embed.FS

const (
	// directorioDeMigraciones es el nombre del directorio embebido.
	directorioDeMigraciones = "migraciones"

	// motivoDeMigracionesIlegibles es lo que no debería poder pasar: las
	// migraciones van dentro del ejecutable y leerlas no toca el disco. Si aun
	// así falla, el binario está mal construido, y eso no lo arregla quien
	// invoca.
	motivoDeMigracionesIlegibles = "no se pueden leer las migraciones que el binario trae dentro"
)

// migracionesEmbebidas devuelve los nombres de las migraciones en el orden en
// que se aplican. embed.FS los devuelve ordenados por nombre y los ficheros van
// numerados desde 0001, así que el orden por nombre es el orden por versión.
//
// Que la numeración sea contigua y sin repetidos no se comprueba aquí, sino en
// TestMigracionesEmbebidasBienFormadas: el código deriva la versión de la
// posición en esta lista —la versión conocida es el número de ficheros— y el test
// es el que vigila que el nombre diga lo mismo que la posición (D7).
func migracionesEmbebidas() ([]string, error) {
	entradas, err := fs.ReadDir(migraciones, directorioDeMigraciones)
	if err != nil {
		return nil, err
	}

	nombres := make([]string, 0, len(entradas))
	for _, entrada := range entradas {
		nombres = append(nombres, entrada.Name())
	}

	return nombres, nil
}

// versionConocida es la versión del esquema que este binario sabe construir: el
// número de migraciones que trae dentro (D7).
func versionConocida() (int64, error) {
	nombres, err := migracionesEmbebidas()
	if err != nil {
		return 0, err
	}

	return int64(len(nombres)), nil
}

// consultante es lo que hace falta para leer la versión del esquema: la base o
// una transacción. Con las dos cosas detrás de la misma interfaz, releer la
// versión **dentro** de la transacción inmediata —que es lo que hace que dos
// invocaciones simultáneas migren una sola vez— no obliga a escribir dos veces
// la misma consulta (D7).
type consultante interface {
	QueryRowContext(ctx context.Context, consulta string, argumentos ...any) *sql.Row
}

// versionRegistrada es la versión que el fichero dice tener: 0 si no existe la
// tabla —una base recién creada, o un cache.db de cero bytes que dejó una
// ejecución interrumpida antes de migrar— y el máximo de la tabla si existe.
//
// Preguntar primero por la tabla es lo que separa «no hay esquema» de «el
// fichero no sirve»: sobre un fichero que no es una base de datos es esta
// consulta la que falla, y con ella la apertura, sin que nadie lo borre (FR-028).
func versionRegistrada(ctx context.Context, base consultante) (int64, error) {
	var tablas int64
	if err := base.QueryRowContext(ctx,
		`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'schema_version'`,
	).Scan(&tablas); err != nil {
		return 0, err
	}

	if tablas == 0 {
		return 0, nil
	}

	var version int64
	if err := base.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(version), 0) FROM schema_version`,
	).Scan(&version); err != nil {
		return 0, err
	}

	return version, nil
}

// migra pone el esquema al día y devuelve la versión que queda aplicada. Solo
// migra el modo normal: crear la base y actualizarla es escribir, y el modo de
// solo lectura no escribe nada (FR-015, FR-024).
//
// Volver a abrir una base ya migrada no ejecuta nada: el bucle está vacío
// (FR-025).
func (c *Cliente) migra(ctx context.Context, base *sql.DB) (int64, error) {
	nombres, err := migracionesEmbebidas()
	if err != nil {
		return 0, errorInesperado("migrar", motivoDeMigracionesIlegibles, err)
	}

	conocida := int64(len(nombres))

	registrada, err := versionRegistrada(ctx, base)
	if err != nil {
		return 0, c.falloAlLeerElEsquema("migrar", err)
	}

	// Un esquema más nuevo que este binario no se toca ni se degrada: quien lo
	// escribió sabía algo que aquí no se sabe, y sus datos siguen siendo suyos
	// (FR-027).
	if registrada > conocida {
		return 0, errorDeVersionAjena("migrar", c.ruta, registrada, conocida)
	}

	for indice, nombre := range nombres {
		version := int64(indice) + 1
		if version <= registrada {
			continue
		}

		if err := c.aplica(ctx, base, version, nombre); err != nil {
			return 0, err
		}
	}

	return conocida, nil
}

// aplica ejecuta una migración pendiente dentro de una transacción inmediata, y
// es donde vive la atomicidad de FR-026: el esquema y la fila que lo registra
// entran juntos o no entra ninguno de los dos, de modo que una interrupción deje
// la versión anterior y nunca un esquema a medias.
func (c *Cliente) aplica(
	ctx context.Context,
	base *sql.DB,
	version int64,
	nombre string,
) (err error) {
	sentencias, err := fs.ReadFile(migraciones, path.Join(directorioDeMigraciones, nombre))
	if err != nil {
		return errorInesperado("migrar", motivoDeMigracionesIlegibles, err)
	}

	// La transacción es inmediata por el _txlock del DSN: toma el bloqueo de
	// escritura al empezar y no al primer INSERT, de modo que dos invocaciones
	// que migran a la vez se turnen en vez de fallar por escalada de bloqueo
	// (D4, sonda 4).
	tx, err := base.BeginTx(ctx, nil)
	if err != nil {
		return c.falloAlLeerElEsquema("migrar", err)
	}

	// Deshacer es lo que garantiza que no quede nada a medias, y vale para las
	// tres salidas: la migración que falló, la que otra invocación ya había
	// aplicado y la que se confirmó. Una transacción ya cerrada devuelve
	// sql.ErrTxDone, que aquí no es un fallo sino la señal de que no había nada
	// que deshacer; cualquier otra cosa se añade a lo que se estuviera
	// devolviendo, sin taparlo.
	defer func() {
		if fallo := tx.Rollback(); !errors.Is(fallo, sql.ErrTxDone) {
			err = errors.Join(err, fallo)
		}
	}()

	// Releer la versión **dentro** de la transacción es lo que hace que la
	// migración se aplique una sola vez: la invocación que esperaba el bloqueo
	// entra cuando la otra ya confirmó y encuentra el trabajo hecho (FR-029, D7).
	registrada, err := versionRegistrada(ctx, tx)
	if err != nil {
		return c.falloAlLeerElEsquema("migrar", err)
	}

	if registrada >= version {
		return nil
	}

	if _, err := tx.ExecContext(ctx, string(sentencias)); err != nil {
		return c.falloAlAplicar(version, nombre, err)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_version(version, aplicada_en) VALUES (?, ?)`,
		version, c.reloj().UTC().Format(time.RFC3339),
	); err != nil {
		return c.falloAlAplicar(version, nombre, err)
	}

	if err := tx.Commit(); err != nil {
		return c.falloAlAplicar(version, nombre, err)
	}

	c.registrador.DebugContext(ctx, "caché: migración aplicada",
		slog.String("ruta", c.ruta),
		slog.String("migracion", nombre),
		slog.Int64("version", version))

	return nil
}

// falloAlAplicar es lo que sale mal mientras se aplica una migración: el
// contexto que termina es «fuente no disponible» (4) y cualquier otra cosa —una
// tabla que ya estaba con otra forma, el disco lleno, la espera agotada— es
// «inesperado» (1). El mensaje nombra la migración y el fichero, que es lo que
// sitúa el fallo (fila 14, FR-033).
func (c *Cliente) falloAlAplicar(version int64, nombre string, causa error) error {
	if esDelContexto(causa) {
		return c.falloDelContexto("migrar", causa)
	}

	return c.falloInesperadoEn("migrar", fmt.Sprintf(
		"no se pudo aplicar la migración %q, la versión %d del esquema de %q",
		nombre, version, c.ruta), causa)
}
