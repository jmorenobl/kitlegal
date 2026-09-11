# Data Model: H1 · Kernel CLI

Entidades del kernel, con sus campos, sus invariantes y las reglas de validación que salen de los
requisitos funcionales. **No hay modelo persistido en H1**: no hay base de datos, ni grafo, ni caché. Lo
que sigue son tipos en memoria y el documento JSON que el binario emite.

Las cinco *Key Entities* del spec se corresponden con los tipos de abajo así: *Sobre de salida* → `Sobre`;
*Error tipado* → `Clase` + sentinelas; *Contexto de ejecución* → `Contexto` **más el `*slog.Logger` que
acompaña a la llamada** (§4: el «nivel de detalle» llega al applet ya resuelto, pero no como campo del
dominio); *Applet* → `Applet`/`Verbo`; *Registro de applets* → `Registro`.

---

## 1. `Sobre` — `internal/core/schema`

La estructura que envuelve el resultado de **cualquier** applet, con éxito o con fallo (FR-010, FR-045).

| Campo | Clave JSON | Tipo | Invariante |
|---|---|---|---|
| `Ok` | `ok` | `bool` | Verdadero **si y solo si** el código de salida del proceso es 0 (FR-014, FR-045) |
| `Fuente` | `fuente` | `string` | Nunca vacía (FR-016). `kitlegal.<applet>` o `kitlegal.cli` si no hay fuente externa; el nombre de la fuente pública en un adaptador |
| `URL` | `url` | `string` | Nunca vacía y siempre un URI absoluto (FR-016, FR-017). `kitlegal:applet/<nombre>` o `kitlegal:cli` si no hay fuente externa |
| `FechaConsulta` | `fecha_consulta` | `time.Time` | RFC 3339 con desplazamiento horario explícito (FR-013) |
| `Hash` | `hash` | `string` | `sha256:` + 64 dígitos hexadecimales en minúscula (FR-011, FR-012) |
| `Data` | `data` | `any` | El contenido del applet cuando `Ok`; `DatosError` cuando no (FR-045, FR-047) |

**Invariantes del tipo, comprobadas en test:**

- **Seis claves exactas en el nivel superior**, ni una más ni una menos, en éxito y en fallo (FR-010,
  FR-045). Se garantiza porque `Sobre` es un `struct` cerrado sin `omitempty` y el esquema declara
  `additionalProperties: false`.
- `Hash` depende **solo** del contenido de `Data`, nunca de `FechaConsulta` ni del orden de serialización
  (FR-011, SC-005).
- Un `Sobre` se construye en **un único sitio** (`internal/cli`), nunca por un applet (FR-015).
- Si `Data` no es serializable, la construcción **falla antes** de que nada llegue a la salida estándar
  (caso límite del spec); el fallo se clasifica como `inesperado`.
- Si **la escritura** del sobre falla (tubería cerrada), el error se propaga y se traduce a código de
  salida —`inesperado`, 1— y **no** se intenta emitir un segundo sobre por el mismo descriptor roto
  ([research.md D15](./research.md#d15--internalrender-único-escritor-de-stdout-dos-formas-y-la-tabla-mínima)).

**Transiciones.** Un sobre no tiene estado ni ciclo de vida: se construye una vez, se escribe una vez y se
descarta. La única «transición» es la del proceso: `Resultado` del applet → `Sobre` con `Ok` verdadero, o
`error` del applet → `Clase` → código de salida **y** `Sobre` con `Ok` falso, ambos desde el mismo punto.

---

## 2. `Procedencia` — `internal/core/schema`

Par `{Fuente, URL}`. Es lo que un applet devuelve junto a su `data`, y lo que hace que el sobre sea
citable. No es un campo libre: FR-016 lo reserva.

| Emisor | `Fuente` | `URL` |
|---|---|---|
| Applet sin fuente externa | `kitlegal.<nombre>` | `kitlegal:applet/<nombre>` |
| El kernel, en un fallo anterior al applet | `kitlegal.cli` | `kitlegal:cli` |
| El kernel, en un fallo del applet | la del applet, si se conoce; `kitlegal.cli` en otro caso | ídem |
| Adaptador de fuente (H4 en adelante) | nombre de la fuente | URL `http(s)` comprobable |

**Validación:** ninguna de las dos puede ir vacía; `URL` debe ser un URI absoluto. El prefijo `kitlegal.`
y el esquema `kitlegal:` están **reservados** y ningún adaptador de `internal/source/<fuente>` puede
usarlos; esa prohibición se hace mecánica en H4, cuando exista el primer adaptador.

---

## 3. `Resultado` — `internal/core/schema`

Lo que un applet devuelve. `{Procedencia, Datos}` y nada más: el applet no conoce `Ok`, ni `Hash`, ni
`FechaConsulta`, ni el formato de salida, ni los códigos de salida.

---

## 4. `Contexto` — `internal/core/schema`

El *Contexto de ejecución* del spec: las decisiones globales que el kernel entrega al applet ya
interpretadas, para que ningún applet tenga que leer una bandera (FR-018).

| Campo | De qué bandera sale | Qué significa en H1 |
|---|---|---|
| `JSON` | `--json` | Forma de presentación. El applet no la usa; la usa `render` |
| `Timeout` | `--timeout` | Ya aplicado como plazo del `context.Context` que acompaña al `Contexto` |
| `Offline` | `--offline` | Se propaga; **sin semántica** hasta H3 (FR-021) |
| `DryRun` | `--dry-run` | Se propaga; cada capa con efectos la honra. En H1 ninguna la tiene (FR-022) |
| `SinGrafo` | `--no-graph` | Se propaga; **sin efecto observable** hasta H12 (FR-023) |
| `Asunto` | `--asunto` | Se propaga; **no abre ni crea nada** hasta H14 (FR-024) |

`--describe` **no** viaja en el `Contexto`: excluye la ejecución, así que el applet nunca llega a verlo
(FR-049).

**`--verbose` y `KITLEGAL_LOG` tampoco viajan en el `Contexto`**, y el `*slog.Logger` **no es un campo de
este tipo**: `internal/core/schema` es dominio puro y un registrador es un puerto con I/O detrás. El
applet recibe el registrador **ya montado y con el nivel ya resuelto** como parámetro explícito de su
método de ejecución (§6), de modo que sigue sin interpretar ninguna bandera —que es lo que la *Key
Entity* «Contexto de ejecución» exige— y el dominio sigue sin importar `log/slog`
([research.md D2](./research.md#d2--reparto-entre-internalcli-internalapp-e-internalcoreschema-sin-ciclos)).
`depguard` y el test de arquitectura lo hacen mecánico: bajo `internal/core/**` están denegados `log`,
`log/slog`, `os`, `io`, `net/http` y `database/sql`.

**Invariantes:**

- El `Contexto` es de solo lectura para el applet; se construye una vez por invocación y no se comparte
  entre invocaciones.
- `internal/core/schema` importa **solo** `crypto/sha256`, `encoding/hex`, `encoding/json` y `time`.

---

## 5. `Clase` y `DatosError` — `internal/core/schema`

`Clase` es un `string` con constantes, y es el vocabulario que aparece en el JSON y en el esquema.

| `Clase` | Código de salida | Significado |
|---|---|---|
| `argumentos` | 2 | Bandera desconocida, valor inválido, argumento obligatorio ausente, applet no registrado |
| `no-encontrado` | 3 | Lo pedido no existe en la fuente |
| `fuente-no-disponible` | 4 | La fuente no responde, o se agotó el plazo de `--timeout` |
| `limite-o-tos` | 5 | Límite de peticiones alcanzado, o restricción de términos de uso |
| `identidad-humana` | 6 | La acción requiere identidad humana; **no se ha realizado ninguna acción con efectos externos** |
| `inesperado` | 1 | Cualquier fallo que no encaje en los anteriores (FR-031) |

`DatosError` es lo que ocupa `data` cuando `ok` es falso:

| Campo | Clave JSON | Tipo | Invariante |
|---|---|---|---|
| `Clase` | `clase` | `Clase` | Una de las seis. Corresponde al código de salida emitido (SC-014) |
| `Mensaje` | `mensaje` | `string` | No vacío. Es el mensaje dirigido a la persona, el mismo que va a la salida de error |

`additionalProperties: false`: ni traza, ni código numérico, ni error envuelto. El detalle técnico va al
registro de eventos, no al sobre.

**Invariante de exhaustividad:** la traducción `Clase` → código de salida vive en **un único** `switch`
sobre el tipo `Clase`, sin rama `default`, de modo que el linter `exhaustive` —activo desde H0— falla si
alguien añade una clase sin darle código (FR-030).

**Invariante de envoltura:** envolver un error con `fmt.Errorf("…: %w", err)` **no** cambia su clase,
porque la clasificación usa `errors.Is` contra los sentinelas (FR-032).

---

## 6. `Applet`, `Verbo` y `Argumentos` — `internal/app`

Lo que un applet declara, y **solo** lo que declara.

| Tipo | Campos / métodos | Para qué |
|---|---|---|
| `Applet` | `Nombre() string`, `Descripcion() string`, `Verbos() []Verbo` | Identidad y catálogo |
| `Verbo` | `Nombre`, `Descripcion`, `Argumentos func() Argumentos`, `Salida any`, `PorOmision bool` | Gramática del verbo, tipo de su `data` y si es el verbo que se asume cuando no se nombra ninguno |
| `Argumentos` | `Ejecutar(ctx context.Context, ec schema.Contexto, log *slog.Logger) (schema.Resultado, error)` | La operación |

**Lo que un applet NO declara** (FR-018, FR-044, SC-010): banderas globales, sobre, huella,
`fecha_consulta`, códigos de salida, formas de presentación, nivel ni destino del registro de eventos —el
`*slog.Logger` le llega montado.

**Invariantes:**

- `Nombre()` no vacío, sin espacios, sin prefijo `-`, distinto de los verbos reservados del binario
  (`version`) y distinto del de cualquier otro applet registrado (FR-008).
- `Verbos()` no vacío, con nombres únicos dentro del applet.
- **Como máximo un verbo con `PorOmision`**. Ninguno es válido: entonces nombrar el verbo es obligatorio y
  omitirlo es un error de argumentos (código 2). Dos es un defecto de compilación, y el registro lo
  rechaza al construirse igual que un nombre duplicado (FR-008).
- `Argumentos` es una **fábrica**: devuelve un valor nuevo en cada invocación, nunca una instancia
  compartida (sin estado global, tests en paralelo).
- `Salida` es un valor cero del tipo de `data`, usado **solo** por reflexión en `--describe`; nunca se
  serializa ni se devuelve al usuario.

**Cómo se usa `PorOmision`** (research.md [D26](./research.md#d26--verbo-por-omisión-cómo-funciona-kitlegal-echo-hola-sin-nombrar-verbo)):
resuelto el applet y antes de construir la gramática, si el primer argumento restante no nombra un verbo
de ese applet y hay verbo por omisión, el kernel lo inserta a la cabeza de la lista. Eso es lo que hace
válida la entrega literal del hito, `kitlegal echo hola --json`, sin relajar la invariante de arriba. Si
se pidió la ayuda del applet (`--help`), no se inserta nada: `kitlegal echo --help` enumera los verbos.

Los dos applets de ejemplo cubren las dos ramas: `echo` declara verbo por omisión; el segundo applet de
SC-010 declara dos verbos y ninguno por omisión, de modo que invocarlo sin verbo termina en código 2.

---

## 7. `Registro` — `internal/app`

La colección de applets del binario. **Única** fuente de la que se derivan el despacho, la ayuda y la
autodescripción; no existe ninguna lista paralela mantenida a mano (FR-001).

| Operación | Qué hace | Cuándo falla |
|---|---|---|
| `Registrar(a Applet)` | Añade un applet | Nombre duplicado, vacío, con prefijo `-` o igual a un verbo reservado; sin verbos, con verbos de nombre repetido o con más de uno marcado `PorOmision` (FR-008) |
| `Buscar(nombre string)` | Devuelve el applet y si existe | — |
| `Nombres() []string` | Los nombres, ordenados, para la ayuda y los mensajes de error | — |

**Invariante de construcción:** un registro inválido se detecta **al construirse**, no en la invocación de
un usuario (FR-008). La raíz de composición convierte ese error en un fallo de arranque; nunca en un
código de salida de usuario.

**Dos registros, un mismo mecanismo** (FR-009):

| Registro | Dónde se construye | Qué contiene en H1 |
|---|---|---|
| De producción | `app.RegistroDeProduccion()`, usado por `cmd/kitlegal` | **vacío**; el primer applet de producto llega en H4 |
| De ejemplo | `internal/app/testdata/kitlegal-e2e`, usado solo por el binario del e2e | `echo` y el segundo applet de ejemplo de SC-010 |

Lo único que cambia entre ambos es **qué** se registra; el mecanismo de registro es exactamente el mismo.

---

## 8. `Globales` — `internal/cli`

Las ocho banderas, declaradas **una sola vez** y embebidas en la gramática de toda invocación. Su contrato
completo —sintaxis, valor por omisión y semántica— está en
[`contracts/banderas-y-exit-codes.md`](./contracts/banderas-y-exit-codes.md).

| Bandera | Tipo | Por omisión | Se convierte en |
|---|---|---|---|
| `--json` | `bool` | falso | `Contexto.JSON` y la elección de forma en `render` |
| `--timeout` | `time.Duration` | `30s` | el plazo del `context.Context` |
| `--offline` | `bool` | falso | `Contexto.Offline` |
| `--dry-run` | `bool` | falso | `Contexto.DryRun` |
| `--describe` | `bool` | falso | rama de autodescripción; **no** llega al `Contexto` |
| `--no-graph` | `bool` | falso | `Contexto.SinGrafo` |
| `--asunto` | `string` | vacío | `Contexto.Asunto` |
| `--verbose` | `bool` | falso | nivel del `*slog.Logger` |

**Invariante:** ningún applet declara ninguna de las ocho, y ninguna puede ser redefinida por un applet
—el registro rechaza un nombre de applet que colisione con una bandera (FR-008) y las globales se embeben
en la gramática, no se copian.

**Tres de ellas se conocen antes de existir la gramática.** Un fallo puede ocurrir antes de que Kong
pueble `Globales` —bandera desconocida— o incluso antes de que exista gramática —applet no registrado—, y
aun así hay que elegir la forma del sobre de fallo y el nivel del registro. Eso lo resuelve un
**pre-escaneo acotado de `argv`** limitado a `--json`, `--verbose` y `--help`, que se detiene en `--` y
cuyo resultado se descarta en cuanto el análisis de Kong entrega el valor definitivo
([research.md D25](./research.md#d25--pre-escaneo-acotado-de-argv-antes-de-la-gramática)).
Es el **único** mecanismo que lee banderas fuera de la gramática, y un test comprueba que coincide con el
análisis para toda invocación bien formada.

---

## Trazabilidad

| Entidad | Requisitos que la definen |
|---|---|
| `Sobre` | FR-010 … FR-017, FR-045, SC-001, SC-005, SC-014, SC-015 |
| `Procedencia` | FR-016, FR-045 |
| `Resultado` | FR-015, FR-044, SC-010 |
| `Contexto` | FR-018, FR-021 … FR-025, SC-010 |
| `Clase`, `DatosError` | FR-029 … FR-034, FR-045, SC-006, SC-014 |
| `Applet`, `Verbo`, `Argumentos` | FR-001, FR-009, FR-018, FR-044, FR-046 … FR-048, SC-010, SC-012 |
| `Registro` | FR-001 … FR-009, FR-026, SC-003 |
| `Globales` | FR-018 … FR-028, SC-010 |
