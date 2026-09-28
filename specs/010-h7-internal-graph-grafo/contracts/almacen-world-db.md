# Contrato: el almacén `world.db` (`internal/graph`) y la lógica del grafo (`internal/core/grafo`)

FR-001 a FR-014, FR-022 a FR-025, FR-070, FR-071, FR-088. Decisiones: [../research.md](../research.md) D8-D18, D33,
D34. Esquema y estados: [../data-model.md](../data-model.md) §3-§4.

## 1. API pública de `internal/graph`

```go
type Opcion func(*ajustes) error
func ConDirectorio(directorio string) Opcion     // sustituye a la regla de ubicación; "" es «argumentos» al resolver; con varias, gana la última

type Almacen struct{ /* opciones y, solo en pruebas, la costura que sustituye a os.Link; nada abierto */ }
func Nuevo(opciones ...Opcion) *Almacen          // guarda una copia de las opciones; no resuelve la ruta ni abre nada
func (a *Almacen) Apply(ctx context.Context, lote core.Lote) error // valor cero y *Almacen nulo = Nuevo()

type Nulo struct{}
func (Nulo) Apply(context.Context, core.Lote) error // nil, sin tocar nada

func Leer(ctx context.Context, opciones ...Opcion) (*Lectura, error)
func (l *Lectura) Ficha(ctx context.Context, id string) (grafo.Ficha, bool, error) // false sin error si no está
func (l *Lectura) Recuento(ctx context.Context) (grafo.Recuento, error)
func (l *Lectura) Instantanea(ctx context.Context) (grafo.Instantanea, error)
func (l *Lectura) Close() error                  // idempotente, también sobre el grafo vacío y una lectura nula

type Error struct {
	Operacion string // «leer» (verbos de graph) o «escribir» (entrega)
	Ruta      string // world.db o el directorio implicado; vacía si aún no se conoce
	Causa     error
	/* clase y motivo privados */
}
func (e *Error) Error() string                   // empieza por «grafo: » y nombra world.db; nunca vacío, ni con *Error nulo
func (e *Error) Unwrap() error
func (e *Error) Clase() schema.Clase             // schema.ConClase; sin clase declarada, «inesperado»
```

Cada verbo de una `Lectura` lee en su propia transacción de lectura; sobre una `Lectura` ya cerrada o nula falla con
`inesperado` («la lectura ya está cerrada»), nunca con un grafo vacío.

`ConDirectorio` es la única opción: la usan los tests (bajo `t.TempDir()`), `DependenciasDeGrafo.Almacen` y la
preparación de evals (contracts/evals-y-skill.md). El directorio de la opción se usa tal cual; lo que haya en él lo
decide la lectura o la entrega. No hay opción de registrador: `internal/graph` no registra con
`slog` ni escribe en ningún descriptor; lo único que una entrega fallida deja en la salida de error es la línea que el
kernel escribe con el error de `Apply` (FR-033, contracts/resultado-y-entrega.md §4), así que, con `--verbose` o sin
ella, una entrega que falla añade exactamente esa línea y una que no falla, ninguna.

Ninguna declaración exportada nombra `database/sql` ni `modernc.org/sqlite` (como FR-005 de H3). `Apply` abre, aplica
y cierra en cada llamada; `Leer` abre para leer y `Close` cierra. `internal/graph` importa `internal/cache` solo por
`cache.Directorio` (research D8) y **no** importa `internal/source/**` ni `internal/render` (FR-092; research D30).

## 2. Ubicación (FR-001, FR-011)

`<directorio>/world.db`, con `<directorio>` = la opción `ConDirectorio` o, sin ella, `cache.Directorio()`: la misma
regla y los mismos errores de clase `argumentos` que la caché. Sin `KITLEGAL_CACHE_DIR` ni `HOME`, el mensaje pide
declarar `HOME` o `KITLEGAL_CACHE_DIR`. La ruta se resuelve **al leer o al entregar**, nunca al construir.

Si `<dir>/world.db` es un enlace simbólico, SQLite abre su destino y, fuera de Windows, nombra y crea los auxiliares
junto a él y con su nombre (research V47): todo lo que este contrato dice de los bytes de `world.db` y de sus
auxiliares vale para ese destino. Un enlace sin destino se lee como un grafo ausente (§3) y la entrega no crea el
destino a través de él (§4, paso 4).

## 3. Lectura (`Leer`)

1. Resolver el directorio (§2). `os.Stat(<dir>/world.db)`: no existe (o un componente no es directorio, o es un
   enlace simbólico sin destino) → `Lectura` de un grafo vacío, sin abrir ni crear nada (FR-004); es un directorio →
   inutilizable (variante «es un directorio», §6); otro fallo → inesperado («no se pudo leer», §6). Un fichero de **0 bytes** → `Lectura` de un grafo vacío **sin abrir
   SQLite**, haya los auxiliares que haya: con un `world.db-wal` no vacío junto a una base de 0 páginas, SQLite borra
   ese `-wal` al abrirla en cualquier modo que no sea `immutable` (`_pagerOpenWalIfPresent`,
   `modernc.org/sqlite@v1.59.0/lib/sqlite_g_000000000001ffff.go:4461-4490`; research V49), y
   un fichero de 0 bytes es siempre una base sin esquema (versión 0).
2. `os.Stat` de `<base>-wal`, `<base>-shm` y `<base>-journal`, con `<base>` = la ruta que usa SQLite para nombrarlos:
   fuera de Windows, `filepath.EvalSymlinks(<dir>/world.db)`; en Windows, `<dir>/world.db` (research D10, V47). Si
   `EvalSymlinks` falla con `fs.ErrNotExist` o `ENOTDIR` (lo retiraron después del `Stat`), grafo vacío, como en 1;
   con otro error, o un `Stat` de un auxiliar que falla sin decir que no existe, inesperado.
3. Comprobar si el proceso puede escribir `world.db`: `os.OpenFile(<dir>/world.db, os.O_RDWR, 0)` y `Close`, sin leer,
   escribir ni truncar (V46). `nil` → puede; `fs.ErrNotExist` → grafo vacío, como en 1; otro error → no puede.
4. Abrir con una conexión (research D10):
   - **puede escribir y no existe ningún auxiliar** (estado normal) →
     `file:<ruta>?mode=rw&_pragma=busy_timeout(100)&_pragma=query_only(1)`, que no deja ningún auxiliar ni cambia
     `world.db` (V9, V36 E);
   - **no puede escribir y no existe `-wal` ni `-journal`** → `file:<ruta>?mode=ro&immutable=1&_pragma=query_only(1)`,
     que lee `world.db` tal cual sin crear ningún fichero ni cambiarlo (V46);
   - **en otro caso**: algún auxiliar con permiso de escritura, o `-wal` o `-journal` sin él (otra conexión abierta, un
     `-wal` huérfano, un `-shm` suelto o el diario de un escritor en modo rollback) →
     `file:<ruta>?mode=ro&_pragma=busy_timeout(100)&_pragma=query_only(1)`, que lee también lo confirmado en el WAL y
     deja `world.db`, `world.db-wal` y `world.db-journal` byte a byte iguales (V36 D, F, V43, V45 c).

   En `<ruta>` (la de `<dir>/world.db` con `/` como separador) se escapan `%`, `?` y `#` (`%25`, `%3F`, `%23`), los
   únicos caracteres que el URI `file:` de SQLite interpreta en el camino; la cadena de escritura (§4) hace lo mismo.
5. Leer la versión con reintento por tramos (0 sin la tabla `schema_version`; si no, `COALESCE(MAX(version), 0)`):
   - contexto terminado → plazo agotado; `SQLITE_BUSY` con el contexto vivo tras la espera propia → bloqueada (§4.1,
     «Esperas»);
   - 26 u otro fallo del fichero → inutilizable; 776 (`SQLITE_READONLY_ROLLBACK`: un diario caliente, que `mode=ro`
     no puede deshacer, V43) → inutilizable «con una transacción interrumpida sin deshacer», sin reintentar otro modo;
     1544/14 (en los modos que no son `immutable`) → si el fichero no se deja leer, inutilizable («acceso denegado» si
     es por permisos); sin `world.db-wal` ni `world.db-journal`, reabrir `mode=ro&immutable=1&_pragma=query_only(1)`;
     con alguno, inutilizable (`immutable` no lee el WAL ni deshace un diario caliente, V36 D, V43); en el modo
     inmutable no se reabre nunca;
   - versión 0 → grafo vacío, sin consultar tablas; versión > 1 → «tiene el esquema en la versión N» (FR-012);
     versión negativa → inutilizable.
6. Cada operación de lectura va en una transacción de lectura (instantánea coherente) con reintento por tramos.
   `query_only` impide cualquier escritura de SQL. Una fila que ninguna entrega escribe (`props` que no es un objeto
   JSON, `ttl` negativo o que no cabe en un `time.Duration`) es la base inutilizable de §6; `ttl` `NULL` es vigencia 0. Sin auxiliares no cambia ni un byte de `world.db` ni de nada en su
   directorio, pueda el proceso escribirlo (V9, V36 E) o no (V46), y tampoco si `world.db` es un enlace simbólico
   (V47). Con un `world.db-journal` frío, vacío o de un escritor vivo, tampoco (V43, V45 c). **Con un auxiliar de
   WAL**, `world.db` y `world.db-wal` quedan iguales, pero SQLite reescribe `world.db-shm` (marcas de lectura y, tras
   un escritor interrumpido, el índice del WAL) o lo crea si faltaba, y con un `-shm` suelto crea un `world.db-wal`
   de 0 bytes —el `-shm` suelto de un lector sobre una base limpia queda igual; el de un escritor con marcos se
   reescribe— (V48): es la desviación declarada de FR-004, FR-031 y SC-004, por su causa —lo que SQLite escribe en los
   auxiliares de WAL para leer lo confirmado en ellos— y con su cota: `world.db`, un `-wal` que ya existía, el diario
   y el contenido del grafo no cambian (research D10; plan, *Complexity Tracking*). Y si una entrega concurrente abre
   `world.db` entre el `Stat` del paso 2 y la apertura y cierra antes que el lector, el cierre del lector lleva a
   `world.db` lo que esa entrega confirmó: los bytes los cambia la entrega, no la lectura.

## 4. Escritura (`Apply`)

Cadena de escritura (research D11, V10): `file:<ruta>?_pragma=busy_timeout(100)&_pragma=synchronous(FULL)&_pragma=foreign_keys(1)&_txlock=immediate`,
**sin** `journal_mode`, una conexión.

1. Consolidar el lote sin tocar el disco: `grafo.Consolidar(lote)`, que valida con `grafo.ValidarLote` (§5) y reduce
   el lote a un registro por clave, y rechazar una `Vigencia` que no es un número entero de segundos (la columna `ttl`
   guarda segundos: «la vigencia … no es un número entero de segundos, que es como se guarda»); un rechazo es un
   `*Error` «el lote no entra en world.db» (§6). Contexto ya terminado → plazo agotado, sin tocar nada.
2. Resolver el directorio (§2) y `os.Stat(<dir>/world.db)`: un directorio → inutilizable (variante «es un
   directorio»), sin tocar nada; ausente (también `ENOTDIR` o un enlace sin destino) → paso 3; otro fallo → «no se
   pudo escribir» (§6); un fichero → paso 4.
3. **Ausente: construir en un temporal y publicar** (research D11, V36 C).
   1. `grafo.ValidarContraGrafoVacio(lote)` (cada extremo de arista es un nodo del lote); un rechazo no crea nada.
   2. Crear los directorios que falten, de uno en uno desde el primer antecesor que no existe (`0o700`), anotando los
      que crea esta llamada (`EEXIST` no se anota).
   3. `os.CreateTemp(<dir>, "world.db-nuevo-*")` (`0o600`); si 3.2 o 3.3 fallan porque algo ya no existe (otra
      invocación retiró el directorio vacío) y el antecesor más cercano que existe (`os.Lstat`), `<dir>` incluido, es
      un directorio (`os.Stat`), volver a 3.2 mientras el contexto siga vivo (terminado → plazo agotado); cualquier
      otro fallo —también «no existe» bajo un enlace sin destino o algo que no es directorio, donde repetir daría lo
      mismo— → «no se puede escribir world.db en "<dir>"» (§6).
   4. Abrir el temporal con la cadena de escritura, `PRAGMA journal_mode=WAL` fuera de toda transacción (tiene que
      devolver `wal`: dentro de una transacción, sobre un fichero de 0 bytes, devuelve `delete` sin error, V42), `BeginTx`,
      `migraciones/0001_grafo.sql`, `schema_version(1, <UTC RFC 3339>)`, el lote por los pasos 6 y 7 sobre un grafo
      vacío, confirmar y cerrar; el cierre no da error y no queda `<temporal>-wal`.
   5. Contexto terminado → plazo agotado, sin publicar.
   6. `os.Link(<temporal>, <dir>/world.db)`: nil → publicado; `errors.Is(err, fs.ErrExist)` → otra invocación lo
      publicó antes (o el nombre lo ocupa un enlace sin destino): retirar el temporal y seguir por el paso 4 con el
      mismo contexto; otro error → la entrega falla («no se puede publicar», §6).
   7. Siempre, borrar el temporal y su `-wal`, `-shm` o `-journal` si quedaran; si la entrega falla (también en el
      paso 4 al que lleva 3.6), `os.Remove` de los directorios anotados en 3.2, del más profundo al menos (`ENOTEMPTY`
      y `ENOENT` no son fallo: otra invocación los usa o ya no están). Cualquier otro fallo de la limpieza se une a la
      causa con `errors.Join`; si `world.db` ya se publicó y lo que falla es retirar el temporal, la entrega devuelve
      ese fallo.
4. **Existe: aplicar en su sitio.** Antes de abrir SQLite, `os.OpenFile(<dir>/world.db, os.O_RDWR, 0)` y `Close`
   (V46): con cualquier error —sin permiso, sistema de ficheros de solo lectura, o `fs.ErrNotExist` de un enlace sin
   destino (al que se llega desde 3.6, porque el nombre está ocupado) o de un fichero retirado entre medias—, la
   entrega falla con «no se puede escribir» (§6) sin abrir SQLite: nada cambia, nada se crea y nada se recupera; lo
   mismo si después falla `os.Stat(<dir>/world.db)`.
   Si `world.db` tiene 0 bytes, `ValidarContraGrafoVacio` va aquí, **antes** de abrir SQLite (la versión 0 se sabe
   sin abrir): un lote que el grafo vacío rechaza no toca ningún fichero, tampoco un `world.db-wal` junto a él.
   Si no hay error, abrir `<dir>/world.db` con la cadena de escritura y leer la versión (reintento por
   tramos): plazo o espera propia agotados → §4.1, «Esperas»; no es base → inutilizable; > 1 → «tiene el esquema en
   la versión N», sin modificar nada salvo la recuperación del final de este paso; negativa → inutilizable; 1 →
   paso 5; 0 → repetir
   `ValidarContraGrafoVacio` y, si `PRAGMA journal_mode` no es `wal`, fijar `PRAGMA journal_mode=WAL` (reintento por
   tramos; fuera de toda transacción, porque dentro SQLite no lo fija: en rollback falla y sobre 0 bytes devuelve
   `delete` sin error, V42; tiene que devolver `wal`, y si no, la entrega falla antes de abrir la transacción). Desde
   aquí, un fallo antes de confirmar deja lo que declara §4.1 para una base sin esquema fuera de WAL. Si hay un diario caliente o un `-wal`
   huérfano, esta conexión los recupera (el diario, en su primera lectura; el `-wal`, en el checkpoint del cierre si
   es la última conexión) aunque la entrega falle después: la otra fila declarada de §4.1 (V43, V44).
5. `BeginTx(context.WithoutCancel(ctx), nil)` con reintento por tramos (inmediata por `_txlock`); releer la versión
   **dentro**; si es 0, ejecutar `migraciones/0001_grafo.sql` e insertar `schema_version(1, <UTC RFC 3339>)`; si es
   posterior, «tiene el esquema en la versión N»; si una sentencia de la migración falla, «no se pudo aplicar la
   migración» (§6).
6. Con el lote ya consolidado en el paso 1, para cada nodo, luego cada texto, luego cada arista (en el orden de sus
   claves): leer lo guardado; en una arista, comprobar antes su origen y después su destino contra el grafo, que ya
   tiene los nodos del lote («su origen/destino no está ni en el lote ni en el grafo»); fusionar con
   `grafo.FusionarNodo`, `FusionarTexto` o `FusionarArista` (data-model §4.1), que rechazan un tipo o un cuerpo
   distintos; y escribir solo si el fusionado difiere de lo guardado (el tipo de un nodo y el cuerpo de un texto no se
   reescriben nunca). Lo guardado que ninguna entrega escribe (una fecha que no es RFC 3339, una vigencia negativa) es
   la base inutilizable de §6. Cualquier rechazo o fallo deshace la transacción entera.
7. Confirmar; cerrar (un fallo al cerrar se une al de la entrega, o es él el fallo). Una transacción deshecha deja
   `world.db` con los mismos bytes si no había un `-wal` huérfano (V11); con él, lo que dice §4.1.

### 4.1 Lo que deja una entrega que falla (FR-033)

| Estado de `world.db` antes | Lo que queda |
|---|---|
| Ausente (con o sin su directorio) | nada: ni `world.db`, ni `world.db-nuevo-*`, ni los directorios que creó la entrega |
| Existe y el proceso no puede escribirlo (sin permiso, sistema de ficheros de solo lectura, enlace sin destino), en cualquier estado | nada: la entrega falla en el paso 4 antes de abrir SQLite; ningún byte cambia, ningún auxiliar aparece y no se recupera nada (V46) |
| Versión 1, o base ya en WAL sin esquema, que el proceso puede escribir, sin `-wal` huérfano | los mismos bytes (V11); sin auxiliares si no hay otra conexión (V12) |
| No es base, esquema posterior, inutilizable, sin `-wal` huérfano ni diario caliente | los mismos bytes (FR-010, FR-012); con ellos, en una base de esquema posterior, la fila siguiente (V44) |
| **Con un `-wal` huérfano o un diario de rollback caliente** (un escritor interrumpido), en cualquier fallo tras abrir para escribir (rechazo, esquema posterior, plazo o espera agotados, E/S) | **desviación declarada**: SQLite recupera lo que dejó ese escritor aunque la entrega falle. El diario caliente se deshace en la primera lectura: `world.db` vuelve a lo que confirmó la última transacción y `world.db-journal` desaparece (V43). El `-wal` huérfano se lleva a `world.db` en el checkpoint del cierre si es la última conexión: `world.db` cambia de bytes y de tamaño y `world.db-wal` y `world.db-shm` desaparecen (V44). El contenido del grafo es el confirmado; con otra conexión abierta no hay checkpoint. Junto a un `world.db` de **0 bytes**, un `-wal` no vacío no se recupera sino que se descarta: la primera lectura de la conexión de escritura lo borra sin llevar nada a `world.db` (`_pagerOpenWalIfPresent`), aunque la entrega falle después (plazo, espera o E/S); un lote que el grafo vacío rechaza no llega a abrir SQLite (§4, paso 4) y no lo toca |
| **Existe sin esquema y no está en WAL** (0 bytes, o base en rollback sin el esquema, escrita por cualquier programa con cualquier versión y configuración de SQLite), fallo **después** de fijar WAL y antes de confirmar (plazo o espera propia agotados en el paso 5, o E/S) | **desviación declarada por su causa y con sus cotas** (research D11): queda lo que SQLite escribe al confirmar el paso a WAL de esa base, una confirmación en modo rollback que hace todo lo que hace cualquier confirmación. **Cotas**, para cualquier base de fuera: el contenido (tablas, filas y esquema de esa base) no cambia; sigue en versión 0 (grafo vacío), ahora en WAL; cerrada la última conexión, no queda ningún fichero que no estuviera antes. **Qué cambia**: bytes de la cabecera de la página 1 —siempre 18 y 19 (1 → 2), el contador de cambios (24-27) y `version-valid-for` (92-95), y, si no coinciden con lo que la confirmación calcula, el tamaño en páginas (28-31), `SQLITE_VERSION_NUMBER` (96-99) y la lista de páginas libres (32-39)—; con `auto_vacuum=full` y páginas libres, el vaciado de todas ellas, que reubica las páginas en uso que están detrás (sus bytes, los del mapa de punteros y los de las páginas que las apuntan) y trunca el fichero (V45 a); un `world.db-journal` frío o vacío desaparece (V45 b); 0 bytes pasa a 4096 (V36 A). Los conjuntos exactos medidos son ejemplos, no una lista cerrada para cualquier base: los fija `aplicar_test.go` (§7; V36 B, V41, V45) |
| Cualquiera, con el proceso terminado por una señal | sin limpieza (V39): pueden quedar `world.db-nuevo-*` y sus auxiliares y los directorios creados; fuera de FR-033 (research D11) |

Esperas (FR-014): cada intento espera como mucho un tramo de 100 ms dentro del motor; entre tramos se mira el
contexto; el total propio es 5 s. Contexto terminado → clase `fuente-no-disponible`; espera propia agotada → clase
`inesperado` «bloqueada por otra invocación». El reloj de la espera es el real.

## 5. Validación del lote (`internal/core/grafo`, FR-024, FR-025)

`ValidarLote` rechaza el lote entero con el primer incumplimiento que encuentra —la procedencia, después todos los
nodos, después el resto de las operaciones en su orden (aristas, textos, nulas o punteros)—, como un `*grafo.Rechazo{Operacion, Motivo}` de clase `inesperado` cuyo mensaje es
`<operación>: <motivo>`: `el lote`, `el nodo "<id>"` (o `el nodo de tipo "Persona"`: un nodo `Persona` no se nombra
por su id), `la arista de "<origen>" a "<destino>" por "<relación>"` (por los ids de sus extremos, sean del tipo que
sean), `el texto "<huella>"` (nunca el cuerpo) o `la operación de tipo <tipo de Go>`, si:

- `Fuente` o `URL` vacías, `URL` que no es URI absoluto (mismo criterio que `schema.Procedencia.Validar`),
  `FechaConsulta` vacía o que no es RFC 3339, `Vigencia` negativa;
- una operación nula («trae una operación nula») o que no es un valor `schema.Nodo`, `schema.Arista` o `schema.Texto`
  (un puntero a uno de ellos: «solo entran los valores schema.Nodo, schema.Arista y schema.Texto»);
- un nodo sin id o sin tipo; una arista sin origen, relación o destino (solo la cadena vacía es «sin»); un texto cuya
  huella no es exactamente `sha256:` + los 64 hexadecimales en minúscula de la SHA-256 de los bytes de su cuerpo;
- el mismo id con dos tipos en el lote;
- un nodo de tipo `Persona` cuyos datos no tienen forma JSON (sin ella no se puede comprobar que no llevan un
  documento);
- un nodo de tipo `Persona` cuyo id, o cualquier **clave o valor de cadena** de sus datos —en su forma JSON canónica,
  la que se guarda— a cualquier profundidad de objetos y listas, casa con:

  ```text
  (?:^|[^A-Za-z0-9])(?:[0-9]{2}[.\- ]?[0-9]{3}[.\- ]?[0-9]{3}[.\- ]?[A-Za-z]|[XYZxyz][.\- ]?[0-9][.\- ]?[0-9]{3}[.\- ]?[0-9]{3}[.\- ]?[A-Za-z]|[A-Za-z][.\- ]?[0-9]{2}[.\- ]?[0-9]{3}[.\- ]?[0-9]{2}[.\- ]?[0-9A-Za-z])(?:[^A-Za-z0-9]|$)
  ```

  (DNI, NIE, NIF de persona jurídica). **Sin** la bandera `(?i)`, cuyo plegado en RE2 es Unicode (research V6): las
  clases son ASCII y explícitas (research D34). Una letra es `A`-`Z` o `a`-`z`; una cifra, `0`-`9`; un separador es
  exactamente `.`, `-` o el espacio U+0020, y solo en esas posiciones; un límite es el principio o el final de la
  cadena o cualquier carácter que no es letra ni cifra ASCII (`á`, `í`, `ñ`, `ſ`, U+00A0, `１`… son límites). Los
  valores que no son cadenas no se examinan.

Casos que el test fija (SC-010), comprobados contra la expresión (research V37), en el id, en un valor de primer
nivel, dentro de una lista de un objeto anidado y como clave:

- **rechazan** (20): `12345678Z`, `12.345.678-Z`, `12 345 678 z`, `12 345 678 Z`, `12345678-Z`, `x1234567l`,
  `X-1.234.567-L`, `X-1234567-L`, `y 1234567 l`, `B12345678`, `b-12.345.678`, `B-12.345.678`, `B1234567J`,
  `b 1234567 j`, `Ana 12345678Z`, `2026-09-28T12:00:00.123456Z`, `2026-09-28T12:00:00.12345678Z`, `ſ12345678Z`
  (U+017F delante), `Martí12345678Z` y `12345678Zá`;
- **entran** (11): `ana-garcia-lopez`, `Ana García López`, `1990-01-01`, `1990-01-01 y 2000-02-02`,
  `2026-09-28T12:00:00Z`, `2026-09-28T12:00:00+02:00`, `A1234567`, `12345678Ñ`, `12345678` + U+212A (signo Kelvin),
  `12` + U+00A0 + `345 678 Z` y `１２３４５６７８Z` (cifras de anchura completa, U+FF11-U+FF18).

Los caracteres que no son ASCII se escriben en el test con su escape de Go (barra invertida, `u` y el código
hexadecimal: U+017F, U+212A, U+00A0…) para que se vean en el diff.

`grafo.Consolidar(lote)` llama a `ValidarLote` y rechaza además un nodo cuyos datos no tienen forma JSON canónica
(RFC 8785: un número que no es finito, una cadena o una clave que no es UTF-8, un valor que no es de JSON), con el
error del codificador como motivo. `ValidarContraGrafoVacio(lote)` rechaza si un extremo de arista no es un nodo del
lote («su origen/destino no es un nodo del lote y el grafo está vacío»). Dentro de la transacción: un extremo que no
está ni en el lote ni en el grafo, un id con otro tipo que el guardado (lo rechaza `FusionarNodo`), una huella
guardada con otro cuerpo (lo rechaza `FusionarTexto`) (FR-024).

## 6. Errores y códigos

Todo error es un `*graph.Error`. Su mensaje empieza siempre por `grafo: ` y nombra `world.db`: por su ruta entre
comillas (`%q`) cuando ya se conoce, y por su nombre, `world.db`, cuando no (la ruta no se pudo ubicar, o el fallo
llegó antes de resolverla: plazo agotado o lote rechazado en el paso 1 de §4).

| Situación | Clase (verbos de `graph`) | Mensaje |
|---|---|---|
| Ruta no resoluble (§2) | `argumentos` (2) | `grafo: no se puede ubicar world.db: <mensaje de la caché, o el de ConDirectorio("")>` |
| No es base, dañada, fila que ninguna entrega escribe, versión negativa, no deja abrirse, WAL sin memoria compartida | `inesperado` (1) | `grafo: "<ruta>" no es una base de datos utilizable; no se modifica` |
| Variante: el fichero no se deja leer por permisos | `inesperado` (1) | `grafo: "<ruta>" no es una base de datos utilizable: acceso denegado; no se modifica` |
| Variante: `world.db` es un directorio | `inesperado` (1) | `grafo: "<ruta>" es un directorio y no una base de datos utilizable; no se modifica` |
| Diario de rollback caliente al leer (776 en `mode=ro`, §3; research D10, V43) | `inesperado` (1) | `grafo: "<ruta>" tiene una transacción interrumpida sin deshacer; no se modifica` |
| Esquema posterior | `inesperado` (1) | `grafo: "<ruta>" tiene el esquema en la versión N y este binario conoce la 1: no se modifica` |
| Bloqueo más largo que la espera propia | `inesperado` (1) | `grafo: "<ruta>" está bloqueada por otra invocación y la espera de 5s se agotó` |
| Plazo agotado | `fuente-no-disponible` (4) | `grafo: el plazo terminó antes de <leer|escribir> "<ruta>"` (o `… escribir world.db` en el paso 1 de §4) |
| Otro fallo de entrada y salida (un `Stat` que falla, cerrar, una lectura ya cerrada, el temporal que queda incompleto) | `inesperado` (1) | `grafo: no se pudo <leer|escribir> "<ruta>": <causa>` |
| Lote rechazado | `inesperado` (solo en la entrega) | `grafo: el lote no entra en world.db: <rechazo>` en el paso 1 de §4; `grafo: el lote no entra en "<ruta>": <rechazo>` contra el grafo vacío o dentro de la transacción; `<rechazo>` es el mensaje del `grafo.Rechazo` (§5) |
| Directorio o temporal que no se pueden crear | `inesperado` (solo en la entrega) | `grafo: no se puede escribir world.db en "<directorio>": <causa>` |
| `world.db` existe y el proceso no puede abrirlo para escribir (§4, paso 4; research V46) | `inesperado` (solo en la entrega; los verbos de `graph` lo leen, §3) | `grafo: no se puede escribir "<ruta>": <causa>; no se modifica` |
| `os.Link` falla con un error distinto de `fs.ErrExist` (research S7) | `inesperado` (solo en la entrega) | `grafo: no se puede publicar world.db en "<directorio>": <causa>` |
| `PRAGMA journal_mode=WAL` no devuelve `wal` (§4, pasos 3.4 y 4; research V42) | `inesperado` (solo en la entrega) | `grafo: no se puede poner "<ruta>" en modo WAL: el modo sigue siendo <modo>` |
| Una sentencia de la migración falla dentro de su transacción (FR-013) | `inesperado` (solo en la entrega) | `grafo: no se pudo aplicar la migración "0001_grafo.sql" en "<ruta>": <causa>` |

Un fallo al cerrar tras otro fallo se une a él con `errors.Join`, y manda la clase del primero. En la entrega de un
applet toda clase se convierte en la línea de aviso de contracts/resultado-y-entrega.md §4.

## 7. Qué lo vigila

| Control | Test |
|---|---|
| Validación, `Persona` (los 20 rechazos y las 11 aceptaciones de §5, en id, valor, anidado y clave), huella, tipos en el lote | `TestValidarLote`, `TestPersonaSinDocumento` (`internal/core/grafo`) |
| FR-023 en los dos órdenes, empates, datos, texto más antiguo, idempotencia | `TestFusionar…`, `TestConsolidar` (`internal/core/grafo`) |
| Ruta (opción, variable, `HOME`), estados de §3 y de data-model §3.1, modo de apertura según los auxiliares y el permiso de escritura, nombre de los auxiliares de un enlace (ruta resuelta fuera de Windows), comprobación de escritura (`nil`, `fs.ErrNotExist`, otro error), esperas | tests de `internal/graph` sobre `t.TempDir()` |
| **Publicación del temporal** (§4 paso 3), con una costura no exportada que sustituye a `os.Link` en el test: `fs.ErrExist` (la costura publica antes el `world.db` de otro almacén) → el lote se aplica sobre ese `world.db` y quedan las dos observaciones; cualquier otro error → la entrega falla y no queda `world.db`, ni ningún `world.db-nuevo-*` ni sus auxiliares, ni los directorios que creó (sí el que ya existía o creó otra invocación: con un fichero dentro, `os.Remove` no lo retira); lote rechazado contra el grafo vacío y contexto ya terminado → no se crea nada | `internal/graph/publicar_test.go` |
| **Lo declarado en §4.1 para una base sin esquema fuera de WAL**, por los pasos internos (por la API no se alcanza de forma determinista): el paso 4 fija WAL; después, otra conexión del test abre una transacción inmediata y la mantiene, y el paso 5 con un plazo corto sale con plazo agotado (`fuente-no-disponible`); cerradas las dos conexiones, en **cada** caso se afirman las cotas —mismas tablas y filas, `wal`, versión 0, ningún fichero que no estuviera antes— y el resultado exacto de esa base, con una lista literal de bytes distintos y el tamaño (V36, V41, V45). Bases que el test construye en `t.TempDir()` con SQL y, donde se dice, retoca con `encoding/binary`: **0 bytes** → 4096 bytes, sin tablas; **tabla ajena, cabecera de este controlador** (contador 2, tamaño en cabecera válido, 3053004 en 96-99) → 18, 19, 27, 95, mismo tamaño; **otra versión de SQLite** (3043002 en 96-99) → 18, 19, 27, 95, 98, 99; **`version-valid-for` distinto del contador y tamaño 0 en 28-31** → 18, 19, 27, 31, 95; **acarreo** (contador y `version-valid-for` 511) → 18, 19, 26, 27, 94, 95; **`auto_vacuum=full` con páginas libres al final** (`auto_vacuum=INCREMENTAL`, tabla `a` con 40 filas de `zeroblob(3000)` borradas, bytes 64-67 a 0) → de 176 128 a 12 288 bytes y, en los que quedan, 18, 19, 27, 31, 35, 39, 95; **`auto_vacuum=full` con una página en uso detrás de las libres** (tablas `a` y `otra`, 40 filas en `a`, después 2 en `otra`, `a` vaciada, 64-67 a 0) → de 188 416 a 24 576 bytes y los 59 de V45: 18, 19, 27, 31, 35, 39, 95, 4106, 4110, 4111, 4115, 12 299, 16 382, 16 384, 16 388-16 393, 16 395 y de 16 399 a 16 547 de 4 en 4; **`auto_vacuum=INCREMENTAL` con páginas libres** → 18, 19, 27, 95 y el mismo tamaño (no se vacía); **diario frío** (tabla ajena escrita con `journal_mode=PERSIST`, diario de 8720 bytes) y **diario vacío** (`TRUNCATE`, 0 bytes) → 18, 19, 27, 95, mismo tamaño y el diario ya no existe; **0 bytes con un diario vacío** → 4096 bytes y sin diario. Y sobre una base ya en WAL sin esquema → los mismos bytes. Si una versión futura del controlador cambia `SQLITE_VERSION_NUMBER` o su forma de confirmar, el caso que cambie deja de valer y el test lo dice: se actualiza su lista con la sonda, nunca se relaja a un intervalo | `internal/graph/aplicar_test.go` |
| Matriz de FR-088 por la API pública: esquema creado y atómico, 0 bytes y base sin tablas leídos sin cambios y creados al entregar, idempotencia (2 y 10 veces, SC-001), fuera de orden, empates en los dos órdenes, reobservación con otros datos, rechazos (sin fuente, tipo cambiado en grafo y en lote, huella ajena, huella guardada con otro cuerpo, extremo ausente, `Persona`) con el grafo intacto, concurrencia de ocho almacenes sobre un directorio que no existe (SC-009: uno publica, los otros siete aplican, las ocho observaciones quedan), inutilizables (no base, directorio, versión posterior, sin permiso de lectura) sin modificar el fichero, plazo agotado (4) y bloqueo (1) | `internal/graph/integracion_test.go` (`//go:build integration`) |
| **Sin permiso de escritura**, en la matriz de integración por la API: con el `world.db` que dejó una entrega (en WAL), en `0400`, el directorio en `0700` y sin auxiliares, `Recuento`, `Ficha` e `Instantanea` leen su contenido y, al terminar, ningún fichero del directorio cambia de huella ni aparece; lo mismo con una base en rollback; `Apply` falla con `inesperado` y el mensaje de §6 y ningún fichero cambia ni aparece; y con un diario caliente junto a ese `world.db` sin permiso, `Apply` falla igual y `world.db` y el diario quedan con los mismos bytes: no se recupera nada (V46). Los permisos se restauran al terminar, como en la caché | `internal/graph/integracion_test.go` |
| **Enlace simbólico**, en la matriz de integración por la API: `world.db` es un enlace a un fichero de otro directorio con un `-wal` huérfano junto a él (copia de los ficheros de un escritor abierto); los tres verbos leen lo confirmado en el WAL y el destino y su `-wal` quedan con los mismos bytes (su `-shm`, la desviación declarada de §3). Un enlace sin destino: los verbos leen un grafo vacío y `Apply` falla con «no se puede escribir», sin crear el destino, ni `world.db-nuevo-*`, ni nada en el directorio (V47) | `internal/graph/integracion_enlace_test.go`, con `//go:build integration && unix` (el nombre de los auxiliares de un enlace es el de Unix; precedente de la restricción: `internal/disco/noregular_unix_test.go`) |
| **Sin residuo con `world.db` ausente**, en la matriz de integración por la API: sobre un directorio de caché que no existe, un lote que el grafo vacío rechaza, un contexto ya terminado y un antecesor sin permiso de escritura (el directorio no se puede crear) → no existe ni `world.db`, ni ningún `world.db-nuevo-*`, ni el directorio. Los fallos **después** de crear el directorio y el temporal los ejerce la fila «Publicación del temporal» | `internal/graph/integracion_test.go` |
| **Lectura con auxiliares**, en la matriz de integración: con un `-wal` huérfano con marcos (copia de los ficheros de un escritor abierto, sin otra conexión) y con un escritor abierto, `Recuento`, `Ficha` e `Instantanea` ven lo confirmado en el WAL y `world.db` y `world.db-wal` quedan con los mismos bytes; se afirma también la desviación declarada: `world.db-shm` existe al terminar (creado si faltaba). Con un `world.db-journal` **caliente** (copia de los ficheros de un escritor en rollback con la transacción a medias y páginas volcadas, `cache_size` pequeño), los tres verbos salen con la clase `inesperado` y el mensaje de §6, y `world.db` y el diario quedan con los mismos bytes; con uno **frío** (cabecera a cero) o de un escritor en rollback **vivo**, se lee lo confirmado y nada cambia (V43); lo mismo con el **vacío** de `TRUNCATE` (0 bytes) y con un diario vacío junto a un `world.db` de 0 bytes (V45 c). Con un `-shm` suelto, sin `-wal`, en sus dos orígenes: el de un lector sobre una base limpia, sin marcos en el índice, se lee, `world.db` y el `-shm` quedan iguales y aparece un `world.db-wal` de 0 bytes (V48); el de un escritor con marcos en el WAL, se lee, `world.db` queda igual, el `-shm` se reescribe (SQLite reconstruye el índice del `-wal` que falta) y aparece un `world.db-wal` de 0 bytes: las dos, la otra parte de la desviación declarada (`gates/plan-pendiente.md`, motivo 2). Con un `world.db` de **0 bytes** y un `world.db-wal` no vacío —sin `-shm`, con `-shm` y con `world.db` en `0400`—, los tres verbos leen un grafo vacío sin abrir SQLite y ningún fichero cambia (§3, paso 1); `Apply` con un lote que el grafo vacío rechaza no cambia ningún fichero, y con un fallo posterior (plazo agotado) deja lo declarado en §4.1: el `-wal` descartado. Sin auxiliares, ningún fichero del directorio cambia | `internal/graph/integracion_test.go` |
| **Recuperación declarada de §4.1** en una entrega que falla, por la API: con un `-wal` huérfano (copia de los ficheros de un escritor abierto) y un lote que se rechaza contra lo guardado (tipo cambiado), `Apply` falla, el grafo tiene lo mismo que antes (lo confirmado en el WAL) y `world.db-wal` y `world.db-shm` ya no existen; lo mismo con una versión posterior confirmada en el WAL; con un diario caliente sobre una base en rollback sin el esquema y un lote que el grafo vacío rechaza (una arista sin sus extremos, así que la entrega falla antes de fijar WAL), `world.db` queda con lo que confirmó la última transacción, sin `world.db-journal` y todavía fuera de WAL (V43, V44) | `internal/graph/integracion_test.go` |
| Sin `database/sql` ni SQLite en la superficie | `TestSuperficieExportada` (`internal/graph/superficie_test.go`) |
| `internal/graph` no importa `source/*` ni `render`; `core` no importa `graph`; SQLite solo en `cache`, `store`, `graph` | `TestArquitectura` (R1, R3, R6) y `depguard` (`core`, `sql`, `grafo`) |
