package cache

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Las tres consultas con las que los tests miran el esquema por dentro. Van
// aquí, como constantes, porque son las mismas en todas las tablas y porque así
// ninguna se compone a partir de una cadena de fuera.
const (
	tablasDeLaVersion = `SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'schema_version'`
	filasDeLaVersion  = `SELECT count(*) FROM schema_version`
	versionAplicada   = `SELECT COALESCE(MAX(version), 0) FROM schema_version`
)

// TestMigracionesEmbebidasBienFormadas fija lo que el resto de las migraciones
// da por hecho: que el orden por nombre es el orden por versión y que la
// versión que este binario conoce es el número de ficheros que trae dentro
// (D7).
//
// La numeración contigua desde 0001 no es un gusto de nombres: el código deriva
// la versión de cada migración de su posición en la lista, así que un hueco o un
// número repetido aplicaría una migración con la versión de otra. Ese es el
// reparto del contrato: la posición la usa el código, el nombre lo vigila este
// test.
func TestMigracionesEmbebidasBienFormadas(t *testing.T) {
	t.Parallel()

	nombres, err := migracionesEmbebidas()
	require.NoError(t, err)
	require.NotEmpty(t, nombres, "el binario trae al menos el esquema v1 (FR-024)")

	formato := regexp.MustCompile(`^(\d{4})_[a-z][a-z0-9_]*\.sql$`)

	for indice, nombre := range nombres {
		coincidencia := formato.FindStringSubmatch(nombre)
		require.NotNil(t, coincidencia, "%q no se llama NNNN_<nombre>.sql", nombre)

		numero, err := strconv.Atoi(coincidencia[1])
		require.NoError(t, err)
		assert.Equal(t, indice+1, numero,
			"la numeración va de 0001 en adelante, sin huecos ni repetidos: %q está en la "+
				"posición %d y el código le daría esa versión", nombre, indice+1)
	}

	conocida, err := versionConocida()
	require.NoError(t, err)
	assert.Equal(t, int64(len(nombres)), conocida,
		"la versión conocida es el número de migraciones embebidas")
}

// TestMigracionesIdempotentes fija FR-025: volver a abrir una base ya migrada
// no aplica nada ni falla. La comprobación no es que la segunda apertura
// funcione —eso pasaría igual aplicando la migración dos veces y tragándose el
// error—, sino que en schema_version haya exactamente una fila por migración
// aplicada.
func TestMigracionesIdempotentes(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	ruta := filepath.Join(directorio, ficheroDeLaBase)

	conocida, err := versionConocida()
	require.NoError(t, err)

	primero, err := New(t.Context(), ConDirectorio(directorio))
	require.NoError(t, err)
	require.NoError(t, primero.Close())

	segundo, err := New(t.Context(), ConDirectorio(directorio))
	require.NoError(t, err)
	assert.Equal(t, conocida, segundo.versionEsquema,
		"la base queda en la versión que este binario conoce")
	require.NoError(t, segundo.Close())

	assert.Equal(t, conocida, consultaEntero(t, ruta, filasDeLaVersion),
		"una fila por migración aplicada, y ni una más: la segunda apertura no vuelve a aplicar nada")
	assert.Equal(t, conocida, consultaEntero(t, ruta, versionAplicada))
}

// TestVersionMayorQueLaConocida fija FR-027 y la fila 9 del contrato de
// errores: el fichero trae un esquema que este binario no conoce —lo escribió
// uno más nuevo—, y la caché no lo usa, no lo degrada y no lo toca.
//
// Decirlo con las dos versiones es lo que permite actuar: quien lee el fallo
// sabe que tiene que volver con el binario que corresponde y que sus datos
// siguen ahí. El SHA-256 es lo que fija que «no se modifica» es verdad.
//
// El 99 lo escribe el test por debajo de la caché, con database/sql: es la
// única forma de imitar un binario más nuevo sin tener uno.
func TestVersionMayorQueLaConocida(t *testing.T) {
	t.Parallel()

	conocida, err := versionConocida()
	require.NoError(t, err)

	for _, modo := range modosDeApertura {
		t.Run(modo.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ruta := filepath.Join(directorio, ficheroDeLaBase)

			creada, err := New(t.Context(), ConDirectorio(directorio))
			require.NoError(t, err)
			require.NoError(t, creada.Close())

			ejecutaEnLaBase(t, ruta,
				`INSERT INTO schema_version(version, aplicada_en) VALUES (99, '2026-01-01T00:00:00Z')`)

			antes := huella(t, ruta)

			cliente, err := New(t.Context(),
				append(slices.Clone(modo.opciones), ConDirectorio(directorio))...)

			require.Error(t, err)
			assert.Nil(t, cliente)
			assert.Equal(t, schema.ClaseInesperado, cli.Clasificar(err))
			assert.Equal(t, 1, cli.CodigoSalida(err))
			assert.Contains(t, err.Error(), ruta, "el mensaje nombra el fichero")
			assert.Contains(t, err.Error(), "versión 99", "…la versión que encontró")
			assert.Contains(t, err.Error(), "conoce la "+strconv.FormatInt(conocida, 10),
				"…y la que este binario conoce")

			assert.Equal(t, antes, huella(t, ruta),
				"un esquema de otra versión no se modifica ni se borra (FR-027)")
			assert.Equal(t, int64(99), consultaEntero(t, ruta, versionAplicada),
				"la versión que había sigue siendo la que hay")
		})
	}
}

// TestMigracionAtomica fija FR-026: una migración entra entera o no entra. La
// interrupción se provoca sin matar ningún proceso, que es lo que un test puede
// medir de verdad: una tabla «entradas» ajena, con otra forma, hace que la
// migración v1 falle en su segunda sentencia **después** de haber creado
// schema_version.
//
// Si la migración no fuera atómica, tras el fallo quedaría una base con
// schema_version creada y sin entradas, y la apertura siguiente la tomaría por
// buena. Lo que se mide es lo contrario: ni rastro de schema_version, y la tabla
// ajena tal como estaba.
func TestMigracionAtomica(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	ruta := filepath.Join(directorio, ficheroDeLaBase)

	const tablaAjena = `CREATE TABLE entradas (ajena TEXT NOT NULL) STRICT`

	ejecutaEnLaBase(t, ruta, tablaAjena)

	cliente, err := New(t.Context(), ConDirectorio(directorio))

	require.Error(t, err)
	assert.Nil(t, cliente)
	assert.Equal(t, schema.ClaseInesperado, cli.Clasificar(err))
	assert.Equal(t, 1, cli.CodigoSalida(err))

	assert.Equal(t, int64(0), consultaEntero(t, ruta, tablasDeLaVersion),
		"la transacción se deshizo entera: schema_version no se queda creada (FR-026)")
	assert.Equal(t, tablaAjena, consultaTexto(t, ruta,
		`SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'entradas'`),
		"lo que había en el fichero sigue como estaba")
}

// TestContextoCanceladoDuranteLaMigracion fija FR-003 y la fila 15 del contrato
// de errores dentro de la migración, que es la única operación de New que
// escribe: el contexto de quien llama se cancela cuando la transacción de la
// migración ya está abierta —el reloj inyectado, al que la migración pide el
// instante de aplicación desde dentro de ella, es quien lo cancela— y New
// termina con «fuente no disponible» (4), con la causa del contexto alcanzable,
// sin devolver ningún cliente y sin registrar ninguna versión: la transacción se
// deshace entera (FR-026) y la base queda como estaba, lista para que la
// siguiente invocación la migre.
//
// Sin la rama del contexto en la clasificación de ese fallo, la misma situación
// saldría como una migración que no se pudo aplicar (1), que es un fichero al
// que culpar de un plazo que no era suyo.
func TestContextoCanceladoDuranteLaMigracion(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	ruta := filepath.Join(directorio, ficheroDeLaBase)

	conocida, err := versionConocida()
	require.NoError(t, err)

	ctx, cancela := context.WithCancel(t.Context())
	t.Cleanup(cancela)

	cliente, err := New(ctx, ConDirectorio(directorio), ConReloj(relojQueCancela(cancela)))

	require.Error(t, err)
	assert.Nil(t, cliente, "un fallo al construir no devuelve ningún cliente")
	assert.Equal(t, schema.ClaseFuenteNoDisponible, cli.Clasificar(err))
	assert.Equal(t, 4, cli.CodigoSalida(err),
		"el contexto terminado es la fila 15, no una migración que no se pudo aplicar: %v", err)
	require.ErrorIs(t, err, context.Canceled, "la causa del contexto sigue alcanzable")
	assert.Contains(t, err.Error(), ruta, "el mensaje nombra el fichero")

	assert.Equal(t, int64(0), consultaEntero(t, ruta, tablasDeLaVersion),
		"la transacción se deshizo entera: no queda ninguna versión registrada (FR-026)")

	migrado, err := New(t.Context(), ConDirectorio(directorio))
	require.NoError(t, err, "la base quedó como estaba y la siguiente invocación la migra")
	assert.Equal(t, conocida, migrado.versionEsquema)
	require.NoError(t, migrado.Close())

	assert.Equal(t, conocida, consultaEntero(t, ruta, filasDeLaVersion),
		"una fila por migración aplicada: la cancelada no dejó ninguna")
}

// relojQueCancela es el reloj inyectado con el que se termina el contexto desde
// dentro de una operación: cada vez que la caché pregunta la hora, cancela y
// devuelve la de la máquina. Es lo que permite que el contexto termine en un
// punto interior conocido —dentro de la transacción de la migración, que le
// pide el instante de aplicación— sin esperar tiempo real ni depender de que el
// motor mire el contexto. Cancelar dos veces no hace nada.
func relojQueCancela(cancela context.CancelFunc) func() time.Time {
	return func() time.Time {
		cancela()

		return time.Now()
	}
}

// TestSoloLecturaNoMigra fija la mitad de FR-015 que se decide al abrir: una
// invocación de solo lectura no crea el esquema ni siquiera cuando no hay
// ninguno. Un cache.db de cero bytes —lo que deja una ejecución normal
// interrumpida antes de migrar— es una base válida y sin tablas: se abre, la
// versión registrada es 0 y sigue siendo 0 al cerrar, porque migrar es escribir.
func TestSoloLecturaNoMigra(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	ruta := filepath.Join(directorio, ficheroDeLaBase)

	require.NoError(t, os.WriteFile(ruta, nil, 0o600))
	antes := huella(t, ruta)

	cliente, err := New(t.Context(), ConDirectorio(directorio), SoloLectura())
	require.NoError(t, err, "una base sin esquema se abre para leerla: no hay nada que servir, no un fallo")
	assert.Equal(t, int64(0), cliente.versionEsquema, "sin esquema, la versión registrada es 0")
	require.NoError(t, cliente.Close())

	assert.Equal(t, antes, huella(t, ruta),
		"una invocación de solo lectura no cambia ni un byte de la base (SC-003)")
	assert.Equal(t, int64(0), consultaEntero(t, ruta, tablasDeLaVersion),
		"la base sigue en la versión 0: en solo lectura no se aplica ninguna migración")
}

// TestDosClientesMigranUnaVez fija el escenario 3 de US5 y FR-029: dos
// invocaciones que arrancan a la vez sobre una caché que todavía no existe se
// serializan y la migración se aplica **una** vez.
//
// Las dos ganan, no una: la transacción inmediata hace que la segunda espere el
// bloqueo de escritura en vez de fallar, y al entrar relee la versión dentro de
// la transacción y encuentra el trabajo hecho. Una sola fila en schema_version
// es lo que distingue eso de aplicarla dos veces y tragarse el error.
func TestDosClientesMigranUnaVez(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	ruta := filepath.Join(directorio, ficheroDeLaBase)

	conocida, err := versionConocida()
	require.NoError(t, err)

	// Dos invocaciones que arrancan a la vez, que es el escenario 3 de US5.
	const aLaVez = 2

	var (
		grupo    sync.WaitGroup
		fallos   = make(chan error, aLaVez)
		abiertos = make(chan *Cliente, aLaVez)
	)

	for range aLaVez {
		grupo.Add(1)

		go func() {
			defer grupo.Done()

			cliente, err := New(t.Context(), ConDirectorio(directorio))
			if err != nil {
				fallos <- err

				return
			}

			abiertos <- cliente
		}()
	}

	grupo.Wait()
	close(fallos)
	close(abiertos)

	for err := range fallos {
		require.NoError(t, err,
			"dos invocaciones a la vez sobre una caché nueva se turnan: ninguna falla (FR-029, FR-032)")
	}

	for cliente := range abiertos {
		assert.Equal(t, conocida, cliente.versionEsquema)
		require.NoError(t, cliente.Close())
	}

	assert.Equal(t, conocida, consultaEntero(t, ruta, filasDeLaVersion),
		"la migración se aplicó una sola vez")
	assert.Equal(t, conocida, consultaEntero(t, ruta, versionAplicada))
}

// consultaEntero y consultaTexto miran el esquema de cache.db sin pasar por el
// cliente, que no deja ejecutar ninguna sentencia desde fuera (FR-005). Abren
// en modo de solo lectura a propósito: comprobar el fichero no puede cambiarlo,
// o las huellas que estas tablas comparan no medirían nada.
func consultaEntero(t *testing.T, ruta, consulta string) int64 {
	t.Helper()

	var valor int64
	require.NoError(t, paraLeer(t, ruta).QueryRowContext(t.Context(), consulta).Scan(&valor))

	return valor
}

func consultaTexto(t *testing.T, ruta, consulta string) string {
	t.Helper()

	var valor string
	require.NoError(t, paraLeer(t, ruta).QueryRowContext(t.Context(), consulta).Scan(&valor))

	return valor
}

// ejecutaEnLaBase escribe directamente en cache.db, que es lo que permite
// dejarla como la habría dejado otro binario —con un esquema más nuevo, o con
// una tabla ajena— sin tener ese otro binario.
//
// Cierra la conexión antes de volver: al cerrarse la última, SQLite consolida
// el registro de escritura en cache.db y lo retira, de modo que la huella del
// fichero no dependa de que quede una conexión abierta. Cerrar dos veces —aquí
// y en la limpieza— no es un fallo.
func ejecutaEnLaBase(t *testing.T, ruta, sentencia string) {
	t.Helper()

	base := abreDirectamente(t, "file:"+filepath.Clean(ruta))

	_, err := base.ExecContext(t.Context(), sentencia)
	require.NoError(t, err)
	require.NoError(t, base.Close())
}

// paraLeer abre cache.db para leerla y sin tocarla.
func paraLeer(t *testing.T, ruta string) *sql.DB {
	t.Helper()

	return abreDirectamente(t, "file:"+filepath.Clean(ruta)+"?mode=ro")
}

func abreDirectamente(t *testing.T, dsn string) *sql.DB {
	t.Helper()

	base, err := sql.Open(controladorSQLite, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, base.Close()) })

	return base
}
