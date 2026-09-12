# Contrato: el puerto `core.Cache` y el cliente `cache.Cliente`

La superficie exportada de los dos paquetes que H3 crea, sus opciones, sus garantías y sus valores por
omisión. Decisiones en [research.md](../research.md) D1, D2, D3, D8, D10 y D14.

## 1. `internal/core` (paquete `core`): el puerto

```go
// internal/core/cache.go
package core

import (
    "context"
    "time"
)

// Cache es el puerto de caché del dominio: guardar bajo una clave con una vigencia y recuperar
// mientras esté vigente. Quien lo implementa vive en un adaptador; el dominio no sabe qué hay detrás.
type Cache interface {
    // Get devuelve el contenido guardado bajo la clave si está vigente. Ausencia (incluida la entrada
    // expirada) es (nil, false, nil); un fallo es err != nil y declara su clase (schema.ConClase).
    // Una implementación construida para no ir a la fuente (solo lectura, --offline) informa la
    // ausencia como fallo de clase «fuente no disponible», que quien llama propaga tal cual (§8);
    // fuera de ese modo la ausencia nunca es un error (FR-013, FR-016).
    Get(ctx context.Context, clave string) (contenido []byte, presente bool, err error)
    // Put guarda el contenido bajo la clave con la vigencia dada, sustituyendo por completo lo que hubiera.
    Put(ctx context.Context, clave string, contenido []byte, vigencia time.Duration) error
}
```

Garantías: el paquete importa solo `context` y `time`; no nombra SQLite, `database/sql`, ficheros ni
`internal/cache` (R1, FR-001, FR-038). No hay `Close` en el puerto (D1).

## 2. `internal/cache`: la superficie exportada, y ninguna más

```go
package cache

const VariableDirectorio = "KITLEGAL_CACHE_DIR"

type Cliente struct { /* privado */ }

func New(ctx context.Context, opciones ...Opcion) (*Cliente, error)

type Opcion func(*ajustes) error // ajustes es privado; se llama así y no «configuracion» por misspell (D15)
func ConDirectorio(dir string) Opcion
func SoloLectura() Opcion
func ConReloj(ahora func() time.Time) Opcion
func ConRegistrador(registrador *slog.Logger) Opcion

func (c *Cliente) Get(ctx context.Context, clave string) ([]byte, bool, error)
func (c *Cliente) Put(ctx context.Context, clave string, contenido []byte, vigencia time.Duration) error
func (c *Cliente) Close() error

type Error struct { … }            // contrato de errores
func (e *Error) Error() string
func (e *Error) Unwrap() error
func (e *Error) Clase() schema.Clase
```

`var _ core.Cache = (*Cliente)(nil)` en el paquete. `TestSuperficieExportada` (patrón de H2) recorre
las declaraciones exportadas con `go/parser` y falla si aparece un selector de `sql` o de `sqlite`, o una
declaración exportada fuera de esta lista (FR-005).

## 3. `New`: qué hace en cada modo

| Paso | Modo normal | Modo de solo lectura (`SoloLectura()`) |
|---|---|---|
| Opciones | se validan en orden; la primera inválida → «argumentos» (2) | igual |
| Ruta | opción > `KITLEGAL_CACHE_DIR` > `~/.cache/kitlegal`; `""` → 2; `Stat(dir)` **existe** y no es directorio → 2 | igual |
| Directorio | `Stat(dir)` **inexistente** u otro error → `MkdirAll(dir, 0o700)`; fallo → 2 | no se crea; `Stat(dir)` **inexistente** → cliente **sin base**; **otro** error (`ErrPermission`, E/S) → 1 |
| Fichero | `OpenFile(cache.db, O_RDWR\|O_CREATE, 0o600)` y cierre; fallo → 2 nombrando el fichero si `cache.db` ya existía (son sus permisos) o el directorio si no (es él quien no deja crearlo) | `Stat(cache.db)`: **inexistente** → sin base; **otro** error (`ErrPermission`…) → 1 |
| Apertura | `file:<ruta>?_pragma=busy_timeout(100)&_pragma=journal_mode(WAL)&_pragma=synchronous(FULL)&_txlock=immediate`; `SetMaxOpenConns(1)`; `<ruta>` codificada para el camino del URI (`%`, `?`, `#`; contrato de apertura §3) | `file:<ruta>?mode=ro&_pragma=busy_timeout(100)&_pragma=query_only(1)`; `SetMaxOpenConns(1)`; ante `SQLITE_READONLY_DIRECTORY` (1544) o `SQLITE_CANTOPEN` (14) en la primera consulta: si `cache.db` no se deja leer (`os.Open`), «inesperado» (1) nombrando el fichero; si se deja leer y no hay `-wal`, reapertura con `&immutable=1`; con `-wal`, «inesperado» (1) |
| Esquema | lee la versión; `> conocida` → 1; aplica las pendientes en transacciones inmediatas | lee la versión; `0` → sin esquema; `== conocida` → lee; otra → 1 |
| Resultado | cliente abierto, base migrada | cliente abierto (o sin base) y `cache.db` idéntico byte a byte |

**Regla «inexistente»** (la única que decide qué fallo de `os.Stat` es una ausencia): un error de `Stat`
es «inexistente» si y solo si `errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR)`. La
segunda condición cubre el directorio cuyo padre es un fichero (`Stat("<fichero>/sub")` devuelve
`ENOTDIR`, que en Go 1.27 **no** es `fs.ErrNotExist`): ese directorio no existe como directorio, y FR-015
manda tratarlo en solo lectura «exactamente como un `cache.db` inexistente» (sin base → 4, `TestDirectorioNoCreable`),
no como un fallo de lectura (1). Todo lo demás (`fs.ErrPermission`, E/S) es «otro» → «inesperado» (1) en
solo lectura. Se define una vez (`ruta.go`) y la usan los dos `Stat` de la tabla; el contrato de apertura §6
y el de errores §3 (filas 8 y 12) dicen lo mismo. Es lo único para lo que `internal/cache` importa `syscall`.

`New` respeta `ctx` en todas las operaciones de base de datos, también mientras espera a que otra
invocación suelte el bloqueo de escritura: esa espera va por tramos de 100 ms que miran el contexto,
hasta 5 s en total (contrato de apertura §4). Si `ctx` vence, el fallo es «fuente no disponible» (4),
en cuanto acaba el tramo en curso, y no queda nada abierto; si el bloqueo dura más que la espera, el
fallo es «inesperado» (1) diciendo que la base está bloqueada.

## 4. `Get` y `Put`: garantías

| | `Get(ctx, clave)` | `Put(ctx, clave, contenido, vigencia)` |
|---|---|---|
| `clave == ""` | 2 | 2 |
| `vigencia <= 0` | — | 2, sin escribir |
| vigencia mayor que cero que lleva la expiración más allá del 11 de abril de 2262 | — | se guarda el último instante representable y la entrada es vigente hasta él; nunca 2 (contrato de apertura §1) |
| cerrado | 1 | 1 |
| solo lectura | ausencia (sin base, sin esquema, sin fila o expirada) → **4** nombrando la clave; presente → valor | **1** nombrando la clave, sin escribir |
| normal, ausente | `nil, false, nil` | — |
| normal, expirada (`!reloj().Before(expira)`) | `nil, false, nil`; la fila se conserva | — |
| normal, presente | `contenido, true, nil`; cero bytes → `[]byte{}` no nulo | upsert en una sentencia; `nil` → BLOB vacío |
| error del driver o de E/S | 1 con la causa | 1 con la causa |
| base bloqueada por otra invocación | espera hasta 5 s por tramos que miran `ctx` (leer en WAL no suele esperar); si persiste, 1 nombrando la ruta, la clave y la espera | igual; el upsert es una sentencia y repetirla tras `SQLITE_BUSY` es seguro |
| `ctx` cancelado o vencido, antes o durante la espera | 4 | 4 |
| tipo de sentencia | `QueryRowContext(...).Scan` (una fila; nada que cerrar) | `ExecContext` (autocommit) |

Lo guardado se devuelve byte a byte (FR-012). Dos claves distintas nunca se responden la una por la otra
(SC-010). La misma clave escrita por dos clientes a la vez termina con la última escritura completa
(FR-007, FR-032): el upsert es una sentencia atómica.

## 5. `Close`

Cierra la conexión (o no hace nada si el cliente está sin base). Devuelve el error de `(*sql.DB).Close`
la primera vez y `nil` las siguientes (FR-004). Tras `Close`, `Get` y `Put` devuelven «inesperado» (1)
sin tocar el disco. Al cerrar el último cliente de un proceso, SQLite retira `-wal` y `-shm` si puede;
un cliente de solo lectura no los retira.

## 6. Valores por omisión y constantes

| Qué | Valor | Dónde |
|---|---|---|
| Directorio por omisión | `~/.cache/kitlegal` | FR-019 |
| Nombre del fichero | `cache.db` | FR-019 (constante privada) |
| Permisos | directorio `0700`, fichero `0600`, auxiliares heredan `0600` | FR-021 |
| `busy_timeout` (tramo del motor por intento) | 100 ms | D4, contrato de apertura §4 |
| Espera ante bloqueo (total, en el cliente) | 5 s | FR-031, `espera.go` |
| `journal_mode` / `synchronous` | `WAL` / `FULL` | FR-030, FR-031, D4 |
| Conexiones por cliente | 1 | D4 |
| Reloj | `time.Now` | D2 |
| Registrador | descarta (`slog.DiscardHandler`) | D14 |
| Versión de esquema conocida | 1 | D6, D7 |

## 7. Concurrencia

`*Cliente` es seguro para uso concurrente desde varias goroutines (la conexión única serializa; el estado
de cierre va con mutex). Dos clientes —mismo proceso o dos procesos— sobre la misma base no fallan por
bloqueo (espera total de 5 s, por tramos de 100 ms que miran el contexto), no corrompen el fichero (WAL) y un lector ve una entrada completa o ninguna y
toda entrada confirmada antes de leer, también si el lector es de solo lectura (FR-032, SC-007).

## 8. Lo que un adaptador de fuente escribirá (patrón, H4)

```go
opciones := []cache.Opcion{cache.ConDirectorio(dir), cache.ConRegistrador(registrador)}
if ejecucion.Offline {
    opciones = append(opciones, cache.SoloLectura())
}
c, err := cache.New(ctx, opciones...)
if err != nil {
    return schema.Resultado{}, err // ya lleva su clase
}
defer func() { err = errors.Join(err, c.Close()) }()

contenido, presente, err := c.Get(ctx, clave)
if err != nil {
    return schema.Resultado{}, err // bajo --offline la ausencia llega como clase 4
}
if !presente {
    respuesta, err := cliente.Pedir(ctx, ejecucion, httpx.Peticion{Metodo: "GET", URL: url}) // sin importar net/http (R2)
    …
    if err := c.Put(ctx, clave, respuesta.Cuerpo, ttl); err != nil { … }
}
```

El adaptador de prueba de H3 (`internal/cache/adaptador_test.go`) es exactamente este patrón con
`ConReloj` añadido para sembrar entradas caducadas.
