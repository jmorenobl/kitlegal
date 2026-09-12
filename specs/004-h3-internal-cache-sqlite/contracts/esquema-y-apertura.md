# Contrato: esquema, migraciones, apertura y `PRAGMA`

La forma de `cache.db`, cómo se crea y actualiza, con qué cadena de conexión se abre en cada modo y qué
garantiza cada `PRAGMA`. Decisiones en [research.md](../research.md) D4, D5, D6 y D7; todo lo que
sigue está verificado con las sondas de la tabla de verificación de research.md.

## 1. Esquema v1 (`internal/cache/migraciones/0001_entradas.sql`)

```sql
CREATE TABLE schema_version (
    version     INTEGER PRIMARY KEY,
    aplicada_en TEXT    NOT NULL
) STRICT;

CREATE TABLE entradas (
    clave     TEXT    PRIMARY KEY NOT NULL,
    contenido BLOB    NOT NULL,
    expira_en INTEGER NOT NULL
) STRICT;
```

Sin índices adicionales (D6). `expira_en` son nanosegundos Unix (`time.Time.UnixNano()`); la vigencia
se decide en Go: vigente ⇔ `reloj().Before(time.Unix(0, expira_en))` (FR-008).

## 2. Migraciones embebidas

- `//go:embed migraciones/*.sql` en un `embed.FS`; ficheros `NNNN_<nombre>.sql`, numeración contigua
  desde `0001`; la **versión conocida** es el número de ficheros (`TestMigracionesEmbebidasBienFormadas`
  lo exige).
- Versión registrada: `0` si no existe la tabla `schema_version` (consulta a `sqlite_master`);
  `COALESCE(MAX(version), 0)` si existe.
- Aplicación (modo normal), por cada versión pendiente en orden: transacción **inmediata**
  (`_txlock=immediate`) → releer la versión dentro de la transacción (si ya no está pendiente, deshacer y
  seguir) → `ExecContext` con el fichero entero → `INSERT INTO schema_version(version, aplicada_en)` →
  `Commit`. Una interrupción deja la base en la versión anterior (FR-026).
- Idempotencia: abrir una base ya en la versión conocida no ejecuta nada (FR-025).
- Versión mayor que la conocida: «inesperado» (1) con esperada y encontrada; el fichero no se toca
  (FR-027).
- Fichero que no es una base: «inesperado» (1) sin borrar ni rehacer (FR-028).
- Solo lectura: nunca se aplica nada; se comprueba la versión (`0` → sin esquema; `== conocida` →
  lectura; otra → 1).

## 3. Cadenas de conexión

| Modo | DSN |
|---|---|
| normal | `file:<ruta>?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(FULL)&_txlock=immediate` |
| solo lectura | `file:<ruta>?mode=ro&_pragma=busy_timeout(5000)&_pragma=query_only(1)` |
| solo lectura, directorio que no admite crear `-shm` (1544 o 14 en la primera consulta) y sin `-wal` | `file:<ruta>?mode=ro&immutable=1&_pragma=busy_timeout(5000)&_pragma=query_only(1)` |

`<ruta>` es el resultado de `filepath.Clean`. Los `_pragma` se aplican en cada conexión nueva por el
driver (`applyQueryParams`), en el orden `busy_timeout` primero y el resto alfabético. Cada cliente
usa `SetMaxOpenConns(1)`.

## 4. `PRAGMA` y garantías (FR-030, FR-031, SC-006)

| `PRAGMA` | Valor | Garantía | Verificación |
|---|---|---|---|
| `journal_mode` | `WAL` | Un lector no bloquea a un escritor ni al revés; lo confirmado se ve aunque viva en el WAL | `PRAGMA journal_mode` → `wal`; sonda 1 D y E |
| `synchronous` | `FULL` (2) | Cada confirmación se sincroniza con el disco: sin pérdida de lo confirmado ni corrupción | `PRAGMA synchronous` → `2`; es el valor por omisión de SQLite |
| `busy_timeout` | `5000` | Ante un bloqueo, esperar hasta 5 s en vez de fallar de inmediato | sonda 1 E frente a sonda 2 G |
| `query_only` | `1` (solo lectura) | Toda escritura falla en la conexión aunque el modo `ro` no bastara | sonda 1 D |
| `_txlock` | `immediate` | Las transacciones (solo las de migración) toman el bloqueo de escritura al empezar | sonda 4 |

`TestAbrirAplicaLosPragma` consulta los cuatro `PRAGMA` en la propia base (SC-006).

## 5. Apertura en modo normal

0. `os.Stat(dir)`: si existe y **no** es un directorio → «argumentos» (2) (FR-022, en cualquier modo).
   Si es un directorio → paso 1. Si falla —**inexistente** según la regla de §6, o cualquier otro error—
   → paso 1, que es el que decide.
1. `os.MkdirAll(dir, 0o700)`; fallo (con `ENOTDIR` cuando el padre es un fichero, sonda 5 de research;
   o por permisos) → «argumentos» (2).
2. `os.OpenFile(ruta, os.O_RDWR|os.O_CREATE, 0o600)` + cierre: crea `cache.db` vacío con acceso
   reservado a la cuenta; los auxiliares `-wal`/`-shm` heredan `0600` (FR-021).
3. `sql.Open` con el DSN normal; `SetMaxOpenConns(1)`.
4. Migrar (§2).
5. Un fallo en 0, 1 o 2 por permisos o por ruta → «argumentos» (2); `SQLITE_READONLY_DIRECTORY` en 4 →
   «argumentos» (2); cualquier otro fallo en 3 o 4 → «inesperado» (1), sin dejar conexión abierta.

## 6. Apertura en modo de solo lectura (FR-015)

**Regla «inexistente»**, común a los dos `Stat` de este modo y definida una sola vez en `ruta.go`
(research D3, sonda 5): un error de `os.Stat` cuenta como «inexistente» si y solo si
`errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR)`; cualquier otro error
(`fs.ErrPermission`, E/S) es «otro». `ENOTDIR` es lo que devuelve `Stat` cuando un componente de la ruta
es un fichero (`"<fichero>/sub/cache.db"`), y en Go 1.27 `syscall.Errno.Is` **no** lo reconoce como
`fs.ErrNotExist` (`syscall_unix.go` 126-127): sin la segunda condición, el «padre que es un fichero» de
`TestDirectorioNoCreable` terminaría en 1 donde FR-015 y SC-011 exigen 4. En Windows `syscall.ENOTDIR`
es `ERROR_PATH_NOT_FOUND` y la regla da el mismo resultado. No hay una tercera rama: todo fallo de
`Stat` es inexistente u otro.

0. `os.Stat(dir)`: existe y no es un directorio → «argumentos» (2). **Inexistente** → cliente **sin
   base** (el directorio, con su padre fichero o sin nada en la ruta, no existe como directorio: es «un
   directorio de caché inexistente» de FR-015; no se crea nada, no se hace ningún otro acceso, toda
   lectura es ausencia → 4). **Otro** → «inesperado» (1). Existe y es un directorio → paso 1.
1. `os.Stat(ruta)` (`<dir>/cache.db`): **inexistente** → cliente **sin base** (ni el directorio ni el
   fichero ni los auxiliares se crean; toda lectura es ausencia → 4). **Otro** (`os.ErrPermission` en un
   directorio `0000`, E/S) → «inesperado» (1), nunca una ausencia falsa.
2. `sql.Open` con el DSN de solo lectura; `SetMaxOpenConns(1)`.
3. Comprobación de esquema (§2) como primera consulta:
   - `SQLITE_READONLY_DIRECTORY` (1544) **o** `SQLITE_CANTOPEN` (14): el directorio no admite crear
     `-shm`; cuál de los dos códigos llega depende de qué auxiliares existan (1544 sin `-wal`; 14 con
     `-wal` y sin `-shm`, research D5 y sonda 3 B, D y E), así que los dos disparan la **misma**
     comprobación y lo que decide es la presencia de `<ruta>-wal`, no el código. Si **no** existe
     `<ruta>-wal` → cerrar y reabrir con `immutable=1` (no hay WAL, nada que perder; nadie puede escribir
     donde no se pueden crear los auxiliares); si la reapertura o su primera consulta vuelven a fallar →
     «inesperado» (1) nombrando la ruta. Si **existe** → «inesperado» (1) nombrando la ruta, `-wal` y
     `-shm` (fila 13 del contrato de errores y su forma de mensaje en §6).
   - `SQLITE_NOTADB` (26) o cualquier otro código → «inesperado» (1).
4. Versión: `0` → sin esquema; `== conocida` → lectura; otra → 1.

Garantía observable (SC-003): `cache.db` es idéntico byte a byte antes y después de cualquier
invocación de solo lectura; `-wal` y `-shm` pueden aparecer (directorio escribible) y no cuentan.
Garantía observable (FR-015, SC-007): un lector de solo lectura ve toda entrada confirmada por otra
invocación, esté o no consolidada en `cache.db`, y no falla por bloqueo mientras aquella escribe.

## 7. Ficheros en disco

| Fichero | Quién lo crea | Permisos | Cuándo desaparece |
|---|---|---|---|
| `<dir>/` | `New` en modo normal | `0700` | nunca (el binario no borra) |
| `cache.db` | `New` en modo normal | `0600` | nunca |
| `cache.db-wal`, `cache.db-shm` | SQLite, al abrir en WAL (también un lector `ro` en directorio escribible) | heredan `0600` | al cerrar el último cliente **normal** del proceso; un lector `ro` no los retira |

Ningún otro fichero: sin temporales, sin copias, sin `.bak`.
