package graph

import (
	"context"
	"database/sql"
	"embed"
	"io/fs"
	"path"
	"time"
)

// migraciones son los ficheros del esquema de world.db que el binario trae
// dentro de sí: no hay nada que instalar ni que encontrar en la máquina, y la
// versión del esquema que el binario conoce viaja con él (FR-003).
//
//go:embed migraciones/*.sql
var migraciones embed.FS

// directorioDeMigraciones es el nombre del directorio embebido.
const directorioDeMigraciones = "migraciones"

// migracion es una versión del esquema: su número, el fichero que la trae y
// sus sentencias.
type migracion struct {
	version    int64
	nombre     string
	sentencias string
}

// migracionesEmbebidas devuelve las migraciones en el orden en que se aplican.
// embed.FS da los ficheros ordenados por nombre y van numerados desde 0001, así
// que la versión de cada una es su posición más uno; que el nombre diga lo
// mismo lo vigila TestMigracionesEmbebidas.
func migracionesEmbebidas() ([]migracion, error) {
	entradas, err := fs.ReadDir(migraciones, directorioDeMigraciones)
	if err != nil {
		return nil, err
	}

	lista := make([]migracion, 0, len(entradas))

	for posicion, entrada := range entradas {
		sentencias, err := fs.ReadFile(migraciones, path.Join(directorioDeMigraciones, entrada.Name()))
		if err != nil {
			return nil, err
		}

		lista = append(lista, migracion{
			version:    int64(posicion) + 1,
			nombre:     entrada.Name(),
			sentencias: string(sentencias),
		})
	}

	return lista, nil
}

// versionConocida es la versión del esquema que este binario sabe construir y
// leer: el número de migraciones que trae dentro. H7 introduce la 1 (FR-013).
func versionConocida() (int64, error) {
	lista, err := migracionesEmbebidas()
	if err != nil {
		return 0, err
	}

	return int64(len(lista)), nil
}

// consultante es lo que hace falta para leer la versión: la base, una conexión
// o una transacción. Leerla fuera y releerla dentro de la transacción de una
// entrega es la misma consulta.
type consultante interface {
	QueryRowContext(ctx context.Context, consulta string, argumentos ...any) *sql.Row
}

// versionRegistrada es la versión que world.db dice tener: 0 sin la tabla
// schema_version —una base recién creada o de 0 bytes, sin esquema (FR-004)— y
// su máximo si la hay. Preguntar primero por la tabla es lo que separa «no hay
// esquema» de «el fichero no sirve»: sobre lo que no es una base de datos es
// esta consulta la que falla.
func versionRegistrada(ctx context.Context, base consultante) (int64, error) {
	var tablas int64
	if err := base.QueryRowContext(ctx,
		`SELECT count(*) FROM sqlite_schema WHERE type = 'table' AND name = 'schema_version'`,
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

// migrar pone el esquema de world.db al día **dentro** de la transacción que
// la pide, y no la confirma ni la deshace: releer la versión dentro de ella es
// lo que hace que dos entregas simultáneas migren una sola vez, y confirmarla
// junto con el lote es lo que hace el esquema atómico —si algo falla a medias,
// quien la pidió la deshace y no queda ningún esquema incompleto— (FR-013;
// contracts/almacen-world-db.md §4, pasos 3.4 y 5).
//
// Cada migración pendiente se ejecuta y se registra en schema_version con el
// instante en UTC y RFC 3339. Sobre la versión conocida no ejecuta nada; un
// esquema posterior no se toca (FR-012); una versión negativa no es de ningún
// esquema y el fichero no es una base utilizable.
func migrar(ctx context.Context, tx *sql.Tx, ruta string) error {
	lista, err := migracionesEmbebidas()
	if err != nil {
		return errorDeMigracionesIlegibles(ruta, err)
	}

	registrada, err := versionRegistrada(ctx, tx)
	if err != nil {
		if terminoElContexto(ctx, err) {
			return errorDePlazo(operacionEscribir, ruta, conElErrorDelContexto(ctx, err))
		}

		return errorInutilizable(operacionEscribir, ruta, err)
	}

	conocida := int64(len(lista))

	switch {
	case registrada > conocida:
		return errorDeVersionPosterior(operacionEscribir, ruta, registrada, conocida)
	case registrada < 0:
		return errorInutilizable(operacionEscribir, ruta, nil)
	}

	aplicadaEn := time.Now().UTC().Format(time.RFC3339)

	for _, pendiente := range lista[registrada:] {
		if _, err := tx.ExecContext(ctx, pendiente.sentencias); err != nil {
			return falloAlMigrar(ctx, ruta, pendiente.nombre, err)
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schema_version(version, aplicada_en) VALUES (?, ?)`,
			pendiente.version, aplicadaEn,
		); err != nil {
			return falloAlMigrar(ctx, ruta, pendiente.nombre, err)
		}
	}

	return nil
}

// falloAlMigrar clasifica lo que sale mal al ejecutar una migración: el
// contexto terminado es el plazo agotado, y cualquier otra cosa —una tabla que
// ya estaba con otra forma, el disco lleno— es la migración que no se pudo
// aplicar.
func falloAlMigrar(ctx context.Context, ruta, nombre string, causa error) error {
	if terminoElContexto(ctx, causa) {
		return errorDePlazo(operacionEscribir, ruta, conElErrorDelContexto(ctx, causa))
	}

	return errorDeMigracion(ruta, nombre, causa)
}
