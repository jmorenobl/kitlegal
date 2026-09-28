# Data model: H7 · grafo del mundo

Entidades del hito, sus campos, sus reglas y sus estados. Los contratos de [contracts/](./contracts/) fijan la forma
exacta de lo que se ve desde fuera; aquí va el modelo del que salen. Decisiones en [research.md](./research.md).

## 1. Lo que emite un applet (`internal/core/schema`, fichero nuevo `grafo.go`)

Se declaran en `schema` porque viajan en `schema.Resultado` y `schema` solo puede importar `crypto/sha256`,
`encoding/hex`, `encoding/json` y `time` (research V15, D1).

```go
// Resultado gana un quinto campo; su valor cero (ninguna operación) es válido.
type Resultado struct {
	Procedencia Procedencia
	Datos       any
	Legible     string
	Ensayo      []string
	Grafo       Observado // lo que la invocación observó del mundo (ADR 0014)
}

type Observado struct {
	Vigencia    time.Duration // la que la fuente declara para la consulta; 0 = no la declara (FR-065)
	Operaciones []Operacion
}

type Operacion interface{ operacionDeGrafo() } // sellada: solo Nodo, Arista y Texto

type Nodo struct {
	ID    string         // id natural: ELI, «ine:<código>», DIR3
	Tipo  string         // Norma, Bloque, BloqueVersion, Municipio, Organo; Persona solo lo vigila Apply
	Datos map[string]any // datos identificativos; nunca el cuerpo de un bloque (FR-041)
}

type Arista struct{ Origen, Relacion, Destino string }

type Texto struct {
	Huella string // «sha256:» + 64 hexadecimales en minúscula de los bytes de Cuerpo (FR-024, FR-071)
	Cuerpo string
}
```

Ninguna operación lleva fuente: la pone el kernel (FR-021).

## 2. El lote y el puerto (`internal/core/graphstore.go`)

```go
type GraphStore interface {
	Apply(ctx context.Context, lote Lote) error // transaccional e idempotente (FR-022)
}

type Lote struct {
	Fuente        string            // «fuente» del sobre presentado
	URL           string            // «url» del sobre presentado
	FechaConsulta string            // «fecha_consulta» del sobre, tal como la escribe (RFC3339Nano; research V2, D5)
	Vigencia      time.Duration     // Observado.Vigencia; 0 = no declarada
	Operaciones   []schema.Operacion
}
```

El kernel construye el lote con `cli.LoteDe(sobre, res.Grafo)` y nadie más lo construye con una procedencia propia
(contracts/resultado-y-entrega.md).

## 3. `world.db` (migración `internal/graph/migraciones/0001_grafo.sql`)

```sql
CREATE TABLE schema_version (
    version     INTEGER PRIMARY KEY,
    aplicada_en TEXT    NOT NULL
) STRICT;

CREATE TABLE nodes (
    id           TEXT    PRIMARY KEY NOT NULL,
    type         TEXT    NOT NULL,
    props        TEXT    NOT NULL,   -- datos identificativos de la última observación, JSON canónico (RFC 8785)
    first_seen   TEXT    NOT NULL,   -- fecha_consulta de la primera observación, tal como la escribió su sobre
    first_source TEXT    NOT NULL,   -- fuente y url de la primera: solo para el desempate de FR-023
    first_url    TEXT    NOT NULL,
    last_seen    TEXT    NOT NULL,   -- fecha_consulta de la última observación
    source       TEXT    NOT NULL,   -- fuente de la última
    url          TEXT    NOT NULL,   -- url de la última
    ttl          INTEGER             -- vigencia de la última en segundos; NULL si no se declaró
) STRICT;
CREATE INDEX nodes_type_source ON nodes(type, source);

CREATE TABLE edges (
    src          TEXT    NOT NULL REFERENCES nodes(id),
    rel          TEXT    NOT NULL,
    dst          TEXT    NOT NULL REFERENCES nodes(id),
    first_seen   TEXT    NOT NULL,
    first_source TEXT    NOT NULL,
    first_url    TEXT    NOT NULL,
    last_seen    TEXT    NOT NULL,
    source       TEXT    NOT NULL,
    url          TEXT    NOT NULL,
    ttl          INTEGER,
    PRIMARY KEY (src, rel, dst)
) STRICT;
CREATE INDEX edges_dst_rel ON edges(dst, rel);

CREATE TABLE texts (
    hash       TEXT PRIMARY KEY NOT NULL,   -- sha256:<hex>, el hash_texto del sobre
    body       TEXT NOT NULL,               -- nunca sale por ningún verbo (FR-070)
    fetched_at TEXT NOT NULL,               -- fecha_consulta de la observación más antigua
    source     TEXT NOT NULL,
    url        TEXT NOT NULL
) STRICT;
```

Nombres de tabla fijados por el hito; los de columna, los de `refs/kitlegal-grafo.md` §3 adaptados (research D12).

### 3.1 Estados de `world.db`

| Estado | Cómo se reconoce | Verbos de `graph` | Entrega de un applet |
|---|---|---|---|
| Ruta no resoluble | `cache.Directorio()` falla (variable vacía, variable que nombra algo que no es directorio, sin variable y sin `HOME`) | 2, clase `argumentos` (FR-011) | una línea de aviso, código del applet, nada creado (FR-033) |
| Ausente | `Stat` → no existe (o un componente de la ruta no es directorio) | grafo vacío, sin crear nada (FR-004) | valida contra grafo vacío; crea los directorios que falten (`0700`) y, en un temporal `world.db-nuevo-*` (`0600`) del mismo directorio, WAL, esquema y lote; lo publica con `os.Link` (FR-013: nunca un `world.db` a medias). Si otra invocación lo publicó antes (`fs.ErrExist`), aplica como en «Versión 1». Un fallo no deja nada: ni `world.db`, ni el temporal, ni los directorios que creó (research D11) |
| Sin esquema | 0 bytes, o base SQLite sin `schema_version` (versión 0) | grafo vacío, bytes intactos si no hay auxiliares (FR-004) | valida contra grafo vacío; fija WAL si no lo está —fuera de toda transacción y comprobando que el pragma devuelve `wal` (research V42)— y crea el esquema y aplica en una transacción (FR-013). Si no estaba en WAL y la entrega falla **después** de fijarlo, queda lo que SQLite escribe al confirmar el paso a WAL, declarado por su causa y con sus cotas: el contenido de la base no cambia, sigue en versión 0 y no queda ningún fichero nuevo; cambian bytes de la cabecera, 0 bytes pasa a 4096, con `auto_vacuum=full` y páginas libres el fichero se vacía y se trunca, y un `world.db-journal` frío o vacío desaparece (research D11, V41, V45). Desviación declarada (contracts/almacen-world-db.md §4.1) |
| Versión 1 | `MAX(version) = 1` | lee | aplica |
| Sin permiso de escritura | `os.OpenFile(<ruta>, os.O_RDWR, 0)` falla sobre un `world.db` que existe (p. ej., `chmod a-w`, sistema de ficheros de solo lectura) | sin `-wal` ni `-journal`: lee con `mode=ro&immutable=1`, sin crear ni cambiar nada (research D10, V46); con alguno, como «Con auxiliares» | falla antes de abrir SQLite: una línea de aviso, `inesperado` «no se puede escribir», nada cambia ni se crea, ni se recupera nada (research D11, V46) |
| Con auxiliares | existe `-wal`, `-shm` o `-journal` con el nombre que les da SQLite —el de la ruta resuelta si `world.db` es un enlace simbólico, fuera de Windows (research V47)—: otra conexión abierta, un `-wal` huérfano de un escritor interrumpido (con marcos confirmados que aún no están en `world.db`), un `-shm` suelto, o el diario de un escritor en modo rollback —vivo, frío, vacío o **caliente** (una transacción interrumpida sin deshacer)— | lee con `mode=ro`, también lo confirmado en el WAL; `world.db`, `-wal` y el diario intactos; con auxiliares de WAL, `-shm` reescrito o creado, y con un `-shm` suelto un `-wal` vacío creado: desviación declarada (research D10, V48). Con un diario caliente no lee: 1, `inesperado`, «tiene una transacción interrumpida sin deshacer», nada cambia (research D10, V43) | aplica; SQLite recupera lo que dejó el escritor interrumpido —deshace el diario caliente al leer y lleva el `-wal` huérfano a `world.db` al cerrar la última conexión— **aunque la entrega falle después**: `world.db` cambia de bytes, con el contenido confirmado; desviación declarada (research D11, V43, V44) |
| Versión posterior | `MAX(version) > 1` | 1, `inesperado`, no se toca (FR-010, FR-012) | aviso, no se toca, salvo la recuperación de «Con auxiliares» |
| Inutilizable | no es base (26), dañada (p. ej., 11 con un tamaño en cabecera mayor que el fichero, sin cambiar nada: research V41), es directorio, no deja abrirse, WAL sin memoria compartida en directorio de solo lectura | 1, `inesperado`, no se toca (FR-010) | aviso, no se toca |
| Bloqueada | `SQLITE_BUSY` durante más de la espera propia (5 s) | 1, `inesperado` (FR-014) | aviso; nada cambia, salvo lo declarado en «Sin esquema» y la recuperación de «Con auxiliares» |
| Plazo agotado | el contexto de `--timeout` termina esperando | 4, `fuente-no-disponible` (FR-014) | aviso; nada creado ni cambiado, salvo lo declarado en «Sin esquema» y la recuperación de «Con auxiliares» |

Los estados «Sin esquema», «Versión 1» y «Versión posterior» se combinan con «Sin permiso de escritura» y con «Con
auxiliares»: la columna de los verbos de `graph` dice cómo se abre en cada caso, y lo que se lee es lo de su fila. Si
`world.db` es un enlace simbólico, todo lo que esta tabla dice de `world.db` vale para su destino, que es el fichero
que SQLite abre; un enlace sin destino es «Ausente» para los verbos de `graph` y «Sin permiso de escritura» para la
entrega, que no crea el destino a través de él (research D10, D11).

## 4. Lo guardado y sus reglas (`internal/core/grafo`)

### 4.1 Observación

Una observación es la llegada de un nodo, una arista o un texto en un lote: (instante de `FechaConsulta`,
`URL`, `Fuente`, `FechaConsulta` como texto, `Vigencia`, y en un nodo sus datos canónicos). Instantes iguales pueden
tener textos distintos (`…T12:00:00+02:00` y `…T10:00:00Z`).

Orden de desempate `≺` entre dos observaciones con el **mismo instante**, comparando bytes: url; después fuente;
después el texto de la fecha; después «sin vigencia» antes que «con vigencia», y entre dos vigencias la menor; y en un
nodo, después, los datos canónicos (FR-023).

| Qué se guarda | Regla |
|---|---|
| Última observación de un nodo o arista (`last_seen`, `source`, `url`, `ttl`, y `props` en un nodo) | la de mayor instante; a igualdad, la menor por `≺` |
| Primera observación de un nodo o arista (`first_seen`, `first_source`, `first_url`) | la de menor instante; a igualdad, la menor por (url, fuente, texto de la fecha) |
| Procedencia de un texto (`fetched_at`, `source`, `url`) | la de menor instante; a igualdad, la menor por (url, fuente, texto de la fecha) |
| Tipo de un nodo | nunca cambia: un lote que da otro tipo se rechaza entero (FR-024) |
| Cuerpo de un texto | nunca cambia: una huella ya guardada con otro cuerpo rechaza el lote (FR-024) |

Consecuencias: el resultado no depende del orden de llegada ni de repetir un lote (FR-022, FR-023, SC-001); una
observación idéntica a la guardada no cambia nada; la última nunca retrocede y la primera nunca avanza.

### 4.2 Validación de un lote

Antes de tocar el disco (`grafo.ValidarLote`): fuente y url no vacías, url URI absoluto, fecha RFC 3339, vigencia no
negativa; cada nodo con id y tipo; cada arista con origen, relación y destino; cada texto con la huella de su cuerpo;
un mismo id con un solo tipo en todo el lote; ningún `Persona` con forma de DNI, NIE o NIF en su id o en cualquier
cadena de sus datos, con letras, cifras y separador ASCII (contracts/almacen-world-db.md §5; research D34). Contra un grafo vacío (fichero ausente o sin esquema): cada
extremo de arista es un nodo del lote. Dentro de la transacción: cada extremo está en el lote o en el grafo; el tipo
de cada id coincide con el guardado; cada huella guardada tiene el mismo cuerpo. Un solo incumplimiento rechaza el lote
entero (FR-024, FR-025).

Dentro de un lote, dos operaciones con la misma clave (id, terna o huella) comparten observación: se consolidan en
una, y dos nodos con el mismo id y datos distintos se quedan con los datos canónicos menores (FR-023).

## 5. Lo que ve el applet `graph` (`internal/core/grafo`, con etiquetas JSON)

| Tipo | Campos (clave JSON) | Verbo |
|---|---|---|
| `Ficha` | `nodo` (`NodoDeFicha`), `salientes` y `entrantes` (`[]AristaDeFicha`, nunca nulas) | `show` |
| `NodoDeFicha` | `id`, `tipo`, `datos` (objeto), `primera_observacion` (texto RFC 3339), `ultima_observacion` (`Procedencia`) | |
| `AristaDeFicha` | `relacion`, `id` (el otro extremo), `primera_observacion`, `ultima_observacion` | |
| `Procedencia` | `fuente`, `url`, `fecha_consulta` (texto, tal como lo escribió el sobre) | `show`, `check` |
| `Recuento` | `nodos`, `aristas`, `textos` (enteros); `nodos_por_tipo` (`[]{tipo, fuente, nodos}`), `aristas_por_relacion` (`[]{relacion, fuente, aristas}`), nunca nulas | `stats` |
| `Hallazgo` | `clase`, `id`, `explicacion`, `procedencia`; `fecha_vigencia` y `fecha_vigencia_reciente` (solo `version-obsoleta`); `vigencia_segundos` (solo `fuente-caducada`) | `check` (`data` es `[]Hallazgo`, nunca nula) |

`Instantanea` (sin JSON): todos los nodos (id, tipo, datos, última observación con su vigencia) y todas las aristas
(origen, relación, destino) de una lectura consistente.

## 6. Reglas de `check` (`grafo.Comprobar(instantanea, ahora)`)

- **`version-obsoleta`** (FR-063, FR-064). Para cada `Bloque` B, sus versiones son los `BloqueVersion` V con arista
  `eli:has_version` B→V y `fecha_vigencia` válida (ocho cifras ASCII AAAAMMDD que forman una fecha:
  `time.Parse("20060102", v)` sin error; research D34, V37). La más reciente es la de
  `fecha_vigencia` mayor; a igualdad, la de última observación de mayor instante; a igualdad, la de id menor. Cada V
  con otra versión W de B de `fecha_vigencia` **estrictamente** posterior da un hallazgo: id de V, procedencia de la
  última observación de la más reciente, `fecha_vigencia` de V y de la más reciente, y la explicación de
  contracts/applet-graph.md §5.
- **`fuente-caducada`** (FR-066). Cada nodo cuya última observación declara vigencia y cuyo instante más la vigencia es
  **estrictamente anterior** a `ahora` da un hallazgo, sea del tipo que sea y aunque ya tenga `version-obsoleta`.
- **Orden** (FR-062): por clase y después por id, comparando bytes.
- **Cita de un nodo** (explicaciones): `Norma` → su `identificador`; `Bloque` → `[<identificador de su Norma>, bloque
  <bloque>]`, con la `Norma` de la arista `eli:has_part` que llega a él; `BloqueVersion` → la cita de su `Bloque`
  (arista `eli:has_version` que llega a él). Sin esos datos, el id del nodo.

## 7. Vocabulario (`internal/core/grafo`, constantes)

| Qué | Valores |
|---|---|
| Tipos | `Norma`, `Bloque`, `BloqueVersion`, `Municipio`, `Organo`, `Persona` |
| Relaciones | `eli:has_part` (Norma → Bloque), `eli:has_version` (Bloque → BloqueVersion), `lb:pertenece_a` (Organo → Municipio) |
| Datos | `Norma`: `identificador`; `Bloque`: `bloque`; `BloqueVersion`: `fecha_vigencia`, `fecha_version`, `norma_modificadora`, `hash_texto`; `Municipio`: `codigo_ine`, `nombre`; `Organo`: `dir3` |
| Clases de hallazgo | `fuente-caducada`, `version-obsoleta` |

Qué emite cada verbo con estos valores: contracts/emision.md.
