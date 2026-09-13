# Data model: H3 · `internal/cache`

Entidades del hito, sus campos, invariantes y transiciones. El puerto vive en `internal/core`; todo lo
demás en `internal/cache`. Las decisiones que lo motivan están en [research.md](./research.md) y la
API exacta en [`contracts/puerto-y-cliente.md`](./contracts/puerto-y-cliente.md).

## 1. Puerto de caché (`core.Cache`)

La idea de caché tal como la ve el dominio (Key Entities «Puerto de caché»). Una interfaz de dos
métodos, sin `Close` (D1).

| Método | Firma | Semántica |
|---|---|---|
| `Get` | `(ctx context.Context, clave string) (contenido []byte, presente bool, err error)` | presente y vigente → `contenido, true, nil`; ausente o expirada → `nil, false, nil`; fallo → `nil, false, err` con `schema.ConClase`. Una implementación construida para no ir a la fuente (solo lectura, `--offline`) informa la ausencia como fallo de clase `fuente-no-disponible`, que quien llama propaga tal cual (FR-016; contrato de errores §5); fuera de ese modo la ausencia nunca es un error |
| `Put` | `(ctx context.Context, clave string, contenido []byte, vigencia time.Duration) error` | guarda o sustituye; `nil` si quedó guardada |

**Invariantes**: `presente == true ⇒ err == nil`; `err != nil ⇒ presente == false`; quien llama no
compara errores para saber si había entrada (FR-013). El paquete `core` importa solo `context` y `time`
(R1).

## 2. Cliente de caché (`cache.Cliente`)

Lo que devuelve `New(ctx, opciones...)`. Único objeto del módulo que abre la base de datos de la caché
(FR-002); implementa `core.Cache` y añade `Close`.

| Campo (privado) | Tipo | Invariante |
|---|---|---|
| `ruta` | `string` | `<directorio>/cache.db`, saneada con `filepath.Clean`; fijada en `New` (FR-023). En los DSN va codificada para el camino de un URI de SQLite (`%` → `%25`, `?` → `%3F`, `#` → `%23`), de modo que la base esté en ese fichero y en ningún otro (FR-020; contrato de apertura §3) |
| `esperaAnteBloqueo` | `time.Duration` | 5 s: el tiempo total que cada operación espera a que otra invocación suelte el bloqueo de escritura, por tramos de 100 ms que miran el contexto (FR-003, FR-031; contrato de apertura §4). Se mide con el reloj real, no con `reloj`; solo las pruebas del paquete la acortan |
| `origen` | enum privado: `opcion`, `variable`, `omision` | De dónde salió el directorio; solo para los mensajes de error (FR-022, FR-035) |
| `soloLectura` | `bool` | Fijado por `SoloLectura()`; nunca cambia después (FR-015) |
| `reloj` | `func() time.Time` | Nunca `nil`: `time.Now` por omisión (FR-009) |
| `registrador` | `*slog.Logger` | Nunca `nil`: descarta por omisión (D14) |
| `db` | `*sql.DB` | `nil` cuando el cliente está **sin base** (solo lectura y directorio o `cache.db` «inexistentes» según la regla de §2) o cerrado; con una sola conexión (`SetMaxOpenConns(1)`) |
| `versionEsquema` | `int` | Versión registrada al abrir: en normal, siempre la conocida tras migrar; en solo lectura, `0` (sin esquema) o la conocida |
| `cerrado` | `bool` + `sync.Mutex` | `Close` lo pone a `true` una vez; `Get`/`Put` lo comprueban antes de tocar `db` |

**Estados** del cliente:

```
            New (normal)                      New (solo lectura)
                │                                    │
   Stat(dir): existe y no es dir → [fallido] (2)   Stat(dir): existe y no es dir → [fallido] (2)
   crea dir 0700, fichero 0600,               ├─ inexistente ──▶ [sin base]  Get→ausencia(→4)  Put→1
   abre, migra ─── error ──▶ [fallido]        ├─ otro error ──▶ [fallido] (1)
                │                             └─ es dir: Stat(cache.db)
                ▼                                 ├─ inexistente ──▶ [sin base]  Get→ausencia(→4)  Put→1
             [abierto]                            ├─ otro error ──▶ [fallido] (1)
        Get/Put/Close                             └─ existe: abre ro, comprueba esquema
                │                                      ├─ 1544|14 y sin -wal → reabre immutable → [abierto ro]
              Close                                    ├─ 1544|14 con -wal, NOTADB, versión ajena → [fallido] (1)
                ▼                                      └─ ok → [abierto ro]  Get→valor|4  Put→1
             [cerrado]  Get/Put → 1 ; Close → nil               │
                                                              Close
                                                                ▼
                                                            [cerrado]
```

**Regla «inexistente»** (la misma en las dos flechas que llevan a `[sin base]`): un fallo de `os.Stat`
es «inexistente» si y solo si `errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR)`.
`ENOTDIR` es el error de `Stat` cuando un componente de la ruta es un fichero (el «padre que es un
fichero» de `TestDirectorioNoCreable`); en Go 1.27 no es `fs.ErrNotExist` y sin nombrarlo ese caso caería
en «otro error» (1) contra FR-015 y SC-011, que exigen 4. Todo lo que no es «inexistente» es «otro error»
(`fs.ErrPermission`, E/S) → «inesperado» (1): no hay una tercera rama. Definida en `ruta.go`
(research D3), la aplican los contratos de apertura §6 y de errores §3 (filas 8 y 12).

`[fallido]` no es un estado del cliente: `New` devuelve `nil, err` y no queda nada abierto (la conexión
que llegó a abrirse se cierra antes de devolver el error).

## 3. Opciones (`cache.Opcion`)

`func(*ajustes) error`, validadas en `New` en el orden en que se pasan; la última repetida gana. El tipo
privado se llama `ajustes` y no `configuracion` porque `misspell` marca la segunda como errata inglesa
(research D15); la superficie exportada no cambia.

| Opción | Efecto | Valor inválido → clase |
|---|---|---|
| `ConDirectorio(dir)` | Directorio efectivo con precedencia máxima (FR-023) | `""` → «argumentos» (2); ruta existente que no es directorio → «argumentos» (2) al construir |
| `SoloLectura()` | Modo de solo lectura (FR-015) | — |
| `ConReloj(ahora)` | Reloj que decide la vigencia (FR-009) | `nil` → «argumentos» (2) |
| `ConRegistrador(l)` | Destino de los eventos `debug` (D14) | `nil` → «argumentos» (2) |

## 4. Ruta efectiva

Se resuelve una vez en `New` (D3). No es un tipo exportado; es el procedimiento:

| Prioridad | Fuente | Valor | Origen registrado |
|---|---|---|---|
| 1 | `ConDirectorio(dir)` | `dir` | `opcion` |
| 2 | `KITLEGAL_CACHE_DIR` **presente** (`os.LookupEnv`) | su valor, aunque sea `""` | `variable` |
| 3 | por omisión | `filepath.Join(os.UserHomeDir(), ".cache", "kitlegal")` | `omision` |

Fichero: siempre `cache.db` dentro de ese directorio (FR-019, FR-020). Ficheros auxiliares que SQLite
puede crear junto a él y que **no** son de este modelo: `cache.db-wal`, `cache.db-shm` (SC-003 los
excluye).

## 5. Entrada de caché (tabla `entradas`)

Lo que se guarda bajo una clave (Key Entities «Entrada de caché»). Esquema v1 (D6):

| Columna | Tipo SQLite (`STRICT`) | Origen | Invariante |
|---|---|---|---|
| `clave` | `TEXT PRIMARY KEY NOT NULL` | `Put(clave)` | Opaca, no vacía (FR-011); una fila por clave (FR-007) |
| `contenido` | `BLOB NOT NULL` | `Put(contenido)` | Opaco, byte a byte (FR-006, FR-012); cero bytes es un valor válido; un `nil` se guarda como BLOB vacío |
| `expira_en` | `INTEGER NOT NULL` | `reloj().Add(vigencia)` en `Put`, en nanosegundos Unix y saturado al intervalo representable (`instanteDeExpiracion`) | Nanosegundos Unix entre `math.MinInt64` (1677) y `math.MaxInt64` (2262), donde `UnixNano` está definido; vigente mientras `reloj().Before(time.Unix(0, expira_en))` (FR-008); una expiración posterior a 2262 se guarda como `math.MaxInt64` y la entrada vale hasta ese instante |

**Transiciones**: `Put` inserta o sustituye (upsert); `Get` no modifica nada; una entrada caducada sigue
en la tabla hasta que un `Put` de la misma clave la sustituya (D6). No hay borrado.

## 6. Vigencia y reloj

- **Vigencia** (`time.Duration`): se declara al escribir (FR-006); `<= 0` es inválida (FR-010). Sin valor
  por omisión: cada fuente traerá el suyo (H4).
- **Reloj** (`func() time.Time`): única fuente de «ahora» para escribir `expira_en` y para decidir en
  `Get`; se inyecta con `ConReloj` (FR-009). El borde es exacto: `ahora == expira_en` es ausencia.
- **Intervalo representable**: `expira_en` solo representa instantes entre el 21 de septiembre de 1677 y
  el 11 de abril de 2262. `Put` satura al extremo más cercano lo que caiga fuera: ninguna vigencia mayor
  que cero se rechaza por larga (FR-010), una expiración posterior a 2262 vale hasta el último instante
  representable y, para cualquier reloj dentro del intervalo, `Get` decide igual que con el instante
  exacto (contrato de apertura §1).

## 7. Versión de esquema (tabla `schema_version`)

| Columna | Tipo | Invariante |
|---|---|---|
| `version` | `INTEGER PRIMARY KEY` | Una fila por migración aplicada; la versión actual es `MAX(version)` |
| `aplicada_en` | `TEXT NOT NULL` | Instante RFC 3339 UTC de la aplicación |

**Versión conocida** por el binario: número de ficheros `migraciones/NNNN_*.sql` embebidos (en H3, 1).

| Registrada | Modo normal | Modo de solo lectura |
|---|---|---|
| tabla ausente (0) | aplica 1..conocida | sin esquema: toda lectura ausencia |
| `< conocida` | aplica las pendientes | «inesperado» (1) (inalcanzable con una sola versión) |
| `== conocida` | nada | lee |
| `> conocida` | «inesperado» (1), fichero intacto | «inesperado» (1) |

Cada migración se aplica en **una transacción inmediata** que relee la versión antes de ejecutar (D7).

## 8. Fallo (`cache.Error`)

| Campo | Tipo | Cuándo se rellena |
|---|---|---|
| `Operacion` | `string` | siempre: `construir`, `migrar`, `leer`, `escribir`, `cerrar` |
| `Ruta` | `string` | cuando hay fichero o directorio implicado |
| `Origen` | `string` | cuando el fallo es de ruta: `opción ConDirectorio`, `variable KITLEGAL_CACHE_DIR`, `ruta por omisión` |
| `Clave` | `string` | cuando hay clave implicada (`leer`, `escribir`) |
| `Causa` | `error` | el error de origen (driver, sistema de ficheros, contexto); `Unwrap` lo expone |
| `clase` (privada) | `schema.Clase` | una de `argumentos`, `fuente-no-disponible`, `inesperado`; nunca otra (FR-033) |

`Error()` nunca devuelve `""` ni entra en `panic`, tampoco sobre un valor cero (patrón de H2). La tabla
cerrada situación → clase → código está en [`contracts/errores-y-codigos.md`](./contracts/errores-y-codigos.md).

## 9. Adaptador de prueba (solo en `cache_test`)

| Tipo | Papel |
|---|---|
| `adaptadorDePrueba{directorio, reloj, cliente *httpx.Cliente}` | `app.Applet` de nombre `prueba`; sus verbos son `consultar <url>` y `guardar <clave> <contenido>` |
| `cuerpoDeLaFuente{Cuerpo string; Origen string}` | contenido de `data` de `consultar`; `Origen` ∈ {`fuente`, `cache`} |
| `argumentosConsultar{URL}` / `argumentosGuardar{Clave, Contenido}` | `app.Argumentos` de cada verbo |

Clave de caché del adaptador: `"prueba:" + url`. Vigencia fija: una hora. El modo lo decide
`ejecucion.Offline` (FR-047). Ningún binario lo registra (FR-044). Su único fixture es la grabación
escrita a mano `internal/cache/testdata/reproduccion/prueba/GET_http_fuente.prueba_norma.json`
(`GET http://fuente.prueba/norma` → 200, `text/plain; charset=utf-8`, cuerpo `<norma>contenido</norma>`),
con el formato y el nombre del contrato de grabación de H2 (D12).

## 10. Flujo de una consulta con caché (lo que H4 repetirá)

```
Ejecutar(ctx, ejecucion, registrador)
  ├─ cache.New(ctx, ConDirectorio(d), ConRegistrador(registrador), [SoloLectura() si ejecucion.Offline])
  │      └─ error → devolver (clase 2 o 1)
  ├─ Get(ctx, clave)
  │      ├─ presente → devolver contenido (origen: cache); Close
  │      ├─ error → devolver (en --offline la ausencia llega como clase 4); Close
  │      └─ ausente → seguir
  ├─ httpx.Pedir(ctx, ejecucion, GET url) → error → devolver; Close
  ├─ Put(ctx, clave, cuerpo, ttl) → error → devolver; Close
  └─ devolver contenido (origen: fuente); Close
```
