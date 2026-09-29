# Contrato: el almacén `world.db` tras H7.1

Lo que cambia de `specs/010-h7-internal-graph-grafo/contracts/almacen-world-db.md` (no se edita). FR-020 a FR-026,
FR-070 a FR-078. Modelo: [../data-model.md](../data-model.md) §1-§4 y §9. Decisiones: [../research.md](../research.md)
D1-D4, D8, D17-D22. Verificaciones: research V1-V10, V21 y V27-V29.

## 1. API pública de `internal/graph`

Igual que en H7 salvo:

```go
func (l *Lectura) Instantanea(ctx context.Context, ambito grafo.Ambito) (grafo.Instantanea, error) // antes, sin ámbito
```

`Almacen` pierde el campo privado `enlazar`. Ninguna declaración exportada nombra `database/sql` ni el controlador
(`superficie_test.go` no cambia). `internal/graph` sigue sin importar `internal/source/**` ni `internal/render`.

## 2. Esquema

Versión conocida por el binario: **2** (dos migraciones embebidas). `migraciones/0002_lecturas.sql` crea `lecturas`
(data-model §1) y `migrar` registra `schema_version(2, <UTC RFC 3339>)`, en la transacción de la entrega que la
encuentra en la versión 0 o 1 (H7 FR 013). `0001_grafo.sql` no cambia.

## 3. Lectura (`Leer` y los verbos de `graph`)

1. Resolver el directorio (H7 §2). `os.Stat(<dir>/world.db)`: no existe (o un componente no es directorio) → grafo
   vacío sin abrir ni crear nada; 0 bytes → grafo vacío sin abrir SQLite (H7 FR 004); otro fallo → «no se pudo leer»
   (1).
2. `os.Stat(<dir>/world.db-wal)`: existe → `file:<ruta>?mode=ro&_pragma=busy_timeout(100)&_pragma=query_only(1)`, que
   lee lo confirmado en el WAL (el de una escritura propia interrumpida o el de otra invocación abierta); no existe →
   `file:<ruta>?mode=rw&_pragma=busy_timeout(100)&_pragma=query_only(1)`, que no deja ningún auxiliar (V8). Un fallo
   del `Stat` que no dice si está → «no se pudo leer» (1).
3. Leer la versión con la espera por tramos (H7 FR 014): contexto terminado → plazo (4); espera propia agotada →
   bloqueada (1). Versión 0 → grafo vacío; 1 o 2 → se lee (con la 1 no hay `lecturas`: todos los bloques son «sin
   fila», data-model §3); posterior → «tiene el esquema en la versión N y este binario conoce la 2: no se modifica»
   (1). Cualquier otro resultado es la regla genérica: `errorInutilizable` (1, la ruta y la causa en el mensaje;
   FR-070).
4. `Instantanea(ctx, ambito)` en una transacción de lectura (research D8):
   - sin norma: todos los nodos y aristas, como en H7, y todas las filas de `lecturas` (versión 2);
   - con norma: los nodos `Norma` con `json_extract(props, '$.identificador') = <norma>`; los `Bloque` a los que llega
     `eli:has_part` desde ellos, con `json_extract(props, '$.bloque')` en `json_each(<bloques en JSON>)` si se nombran;
     las `BloqueVersion` a las que llega `eli:has_version` desde esos bloques; esas aristas; y las filas de `lecturas`
     de esos bloques (versión 2). Una norma o un bloque que no está no es un error: la instantánea no lo trae.
   Nodos por id y aristas por terna, comparando bytes, como en H7.

Se retiran (FR-072): la comprobación del permiso de escritura, el modo `immutable=1` y la reapertura en él, la ruta de
los auxiliares por `filepath.EvalSymlinks`, la búsqueda de `-shm` y `-journal`, y la clasificación de
`SQLITE_READONLY_ROLLBACK` (776), `SQLITE_READONLY_DIRECTORY` (1544) y `SQLITE_CANTOPEN` (14). Un `world.db` que es un
directorio, un enlace, sin permisos, con diarios o `-shm` ajenos o escrito por otra aplicación sigue la regla genérica:
lo que dé, si falla es 1 con la ruta; no se promete nada sobre sus bytes.

## 4. Escritura (`Apply`)

1. Consolidar el lote sin tocar el disco (`grafo.Consolidar`, con la validación reducida de §5); contexto ya
   terminado → plazo, sin tocar nada.
2. Resolver el directorio; `os.MkdirAll(<dir>, 0o700)`: un fallo → «no se puede escribir world.db en "<dir>": <causa>»
   (H7 FR 011, FR 033).
3. `os.OpenFile(<dir>/world.db, O_RDWR|O_CREATE, 0o600)` y cerrar (research D18, V3): lo crea en su sitio si no está,
   sin temporal ni enlace (FR-071); un fallo → «no se pudo escribir» (1).
4. Abrir con la cadena de escritura de H7 (`busy_timeout(100)`, `synchronous(FULL)`, `foreign_keys(1)`,
   `_txlock=immediate`) y leer la versión con la espera por tramos: plazo o espera → como en H7; posterior → «tiene el
   esquema en la versión N…» sin modificar; 0 → `PRAGMA journal_mode=WAL` fuera de toda transacción, que tiene que
   devolver `wal` (V4); 1 o 2 → sigue; cualquier otro resultado → `errorInutilizable` (regla genérica, FR-070).
5. Transacción inmediata con la espera por tramos; `migrar` a la 2 dentro de ella.
6. Para cada `grafo.Lectura` del lote (data-model §2), **antes** de escribir nada del lote: la fila guardada de su
   bloque o, si no la hay, la de partida `(R, R)` con `R` la redacción vista sin lecturas de las `BloqueVersion` que el
   grafo ya guarda de ese bloque (`grafo.RedaccionVistaSinLecturas`; research D3), o la propia lectura si no guarda
   ninguna; y la fila nueva, `Leida(v)` (data-model §4).
7. Nodos, textos y aristas como en H7 (fusión y escritura solo si cambia). Una arista con un extremo que no está ni en
   el lote ni en el grafo la rechaza la clave ajena (V5) y la transacción entera se deshace: «no se pudo escribir…»
   (1).
8. Cada fila nueva que difiere de la guardada: `INSERT INTO lecturas (bloque, ultima, anterior) VALUES (?, ?, ?) ON
   CONFLICT (bloque) DO UPDATE SET ultima = excluded.ultima, anterior = excluded.anterior` (V6). Una entrega repetida
   idéntica no cambia ninguna fila ni escribe nada (H7 FR 022).
9. Confirmar y cerrar.

Una entrega que falla deja el grafo como estaba salvo en lo que dice FR-071: una creación interrumpida deja como mucho
el directorio y un `world.db` sin esquema (0 bytes, o en WAL sin tablas), que la lectura trata como grafo vacío y la
entrega siguiente completa. Se retiran `publicar.go` entero, la costura `Almacen.enlazar`, la limpieza de temporales y
de directorios creados, la comprobación de permisos antes de abrir y `ValidarContraGrafoVacio` (FR-071, FR-072,
FR-075).

## 5. Validación del lote (`internal/core/grafo`)

Rechazan el lote entero, sin escribir nada, y cada uno con un test: operación sin fuente, url o fecha de consulta; un id
que llega con otro tipo que el guardado o que el que le da otra operación del lote; un texto cuya huella no es la de su
cuerpo o cuya huella ya guarda otro cuerpo; un nodo `Persona` con la forma de un DNI, un NIE o un NIF (clases ASCII,
constitución VII) (FR-074, FR-075). Se retiran: URI no absoluto, nodo sin id o sin tipo, arista sin relación, arista sin
extremos y el grafo vacío como caso aparte. Ninguna otra entrada del lote es un caso de este contrato: lo que ningún
emisor produce (el kernel escribe la fecha del sobre con `time.RFC3339Nano`, research V27; la única vigencia declarada
es la constante de `boe`, V21; los emisores solo ponen cadenas en los datos y ninguno emite `Persona`, V28; solo ponen
valores `schema.Nodo`, `schema.Arista` y `schema.Texto`, nunca `nil` ni un puntero, V29) es la regla
genérica —el lote no entra, §6— sin test propio; las comprobaciones que hoy lo rechazan se quedan en el código
(research D20). Lo mismo vale para lo guardado que ninguna entrega escribe (research D21).

Desempate (FR-076): con el mismo instante de consulta, gana la observación de `url` menor comparando bytes; con la misma
`url`, la guardada (una observación idéntica no cambia nada). Dentro de un lote, un id repetido —`boe articulos` repite
la `Norma` en cada bloque, con los mismos datos (research V29)— se guarda una vez.

## 6. Errores del almacén

| Situación | Clase | Mensaje |
|---|---|---|
| ruta no resoluble (H7 FR 011) | `argumentos` | como en H7 |
| `world.db` que no se puede usar, por cualquier causa distinta de las de abajo | `inesperado` | `grafo: "<ruta>" no es una base de datos utilizable: <causa>` |
| esquema posterior | `inesperado` | `grafo: "<ruta>" tiene el esquema en la versión N y este binario conoce la 2: no se modifica` |
| espera propia agotada / plazo agotado | `inesperado` / `fuente-no-disponible` | como en H7 |
| lote rechazado (§5) | `inesperado` | `grafo: el lote no entra en "<ruta>": <motivo>` |
| directorio que no se puede crear | `inesperado` | `grafo: no se puede escribir world.db en "<dir>": <causa>` |
| `PRAGMA journal_mode=WAL` que no devuelve `wal` | `inesperado` | como en H7 |
| cualquier otro fallo de E/S o del controlador (también la clave ajena de §4.7) | `inesperado` | `grafo: no se pudo <leer\|escribir> "<ruta>": <causa>` |
| migración que falla | `inesperado` | como en H7 |

Se retiran «es un directorio…», «tiene una transacción interrumpida sin deshacer», «no se puede escribir "<ruta>"… no
se modifica», «no se puede publicar…» y el detalle «: acceso denegado». En la entrega de un applet, cualquiera de estos
es una línea en la salida de error y el código y la salida del applet no cambian (H7 FR 033).

## 7. Qué lo vigila

| Qué | Test |
|---|---|
| Lecturas: la secuencia de FR-025, una lectura por bloque e invocación, sin lectura con `--no-graph`/`--dry-run` | `internal/core/grafo` (`Consolidado.Lecturas`, `Comprobar`), `internal/graph` (`Apply` y `Lectura`), guion `grafo-lecturas` (contracts/arnes-e2e.md) |
| `world.db` de H7 (FR-026, SC-012) | `TestIntegracionGrafoDeH7` (`internal/graph/integracion_test.go`) |
| Ámbito acotado | `internal/graph` (`Instantanea` con y sin norma, bloques conocidos y desconocidos), guion `grafo-check-acotado` |
| Creación en su sitio, 0700/0600, en WAL con esquema 2 | `almacen_test.go` (lo que afirmaba `TestPublicar` sobre el resultado) |
| Regla genérica | «no es una base» en `TestIntegracionInutilizables`, `TestApplyNoModifica` (sin la aserción de que no cambia), `lectura_test.go` y guion `grafo-regla-generica` |
| Lo que se queda de H7 | 0 bytes y sin esquema; `-wal` de una escritura propia interrumpida (`TestIntegracionLecturaConWAL`, `TestIntegracionRecuperacionDeclarada` sin el caso del diario); concurrencia; esquema posterior; ubicación y directorio que no se puede escribir (`TestIntegracionSinResiduo` sin el lote que el grafo vacío rechaza); espera al leer con el vehículo de research D22 |
