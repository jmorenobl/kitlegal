# Contrato: registro de applets, despacho multicall y autodescripción

Dos interfaces distintas en un mismo documento porque comparten una fuente: **el registro de applets es lo
único de lo que se derivan el despacho, la ayuda y la autodescripción** (FR-001).

- Hacia **quien escribe un applet**: qué tiene que declarar, y qué recibe gratis.
- Hacia **quien consume el binario**: cómo se elige el applet y qué devuelve `--describe`.

**Requisitos que lo definen**: FR-001 … FR-009, FR-026, FR-044, FR-046 … FR-049. **Criterios**: SC-003,
SC-007, SC-010, SC-012.

---

## 1. Lo que declara un applet

```go
type Applet interface {
    Nombre() string          // "echo", y más adelante "boe", "cita", "plazos"…
    Descripcion() string     // una línea, para la ayuda
    Verbos() []Verbo
}

type Verbo struct {
    Nombre      string
    Descripcion string
    Argumentos  func() Argumentos // fábrica del struct de argumentos, con etiquetas del analizador
    Salida      any               // valor cero del tipo de `data`, solo para reflexión
    PorOmision  bool              // el verbo que se asume cuando la invocación no nombra ninguno
}

type Argumentos interface {
    // El registrador llega montado y con el nivel ya resuelto: el applet no lee
    // --verbose ni KITLEGAL_LOG, y el dominio no importa log/slog.
    Ejecutar(ctx context.Context, ec schema.Contexto, log *slog.Logger) (schema.Resultado, error)
}
```

**Lo que un applet NO declara, y recibe igualmente**: las ocho banderas globales, el sobre de salida, la
huella, `fecha_consulta`, los códigos de salida, las dos formas de presentación y el registro de eventos
—destino, formato y nivel incluidos—. Registrar un applet nuevo que solo declara nombre, verbos y
contenido de `data` basta para que lo herede todo (SC-010).

**El verbo por omisión.** `PorOmision` es la única declaración que el contrato añade sobre lo literal del
spec, y existe porque la entrega del hito lo exige: `kitlegal echo hola` no nombra verbo. Un applet puede
no marcar ninguno —entonces el verbo es obligatorio— y nunca puede marcar dos.

**Lo que un applet devuelve**: `schema.Resultado{Procedencia, Datos}`. Nunca un sobre ya montado, nunca
texto escrito en un descriptor, nunca un código de salida.

**Reglas de validación del registro** (FR-008), comprobadas **al construirse** y nunca en la invocación de
un usuario:

- Nombre no vacío, sin espacios y sin prefijo `-` (colisión con una bandera).
- Nombre distinto del de cualquier otro applet ya registrado.
- Nombre distinto de los verbos reservados del binario (`version`).
- Al menos un verbo, con nombres únicos dentro del applet.
- **Como máximo un verbo marcado `PorOmision`.**

Un registro que incumple cualquiera de esas reglas es un **defecto de compilación**: se detecta al
arrancar, no se convierte en un código de salida de usuario.

---

## 2. Despacho multicall

Sea `n` el nombre con el que se invocó al binario (`os.Args[0]`, quedándose con el último componente de la
ruta y sin el sufijo `.exe`).

| Caso | Qué ocurre |
|---|---|
| `n` está registrado | **Manda `n`.** Los argumentos se entregan **íntegros** al applet: `./echo boe` ejecuta `echo` con el argumento `boe` |
| `n` no está registrado (incluido `n == "kitlegal"`) | El applet se toma del **primer argumento**; el resto se entrega al applet |
| `n` no está registrado y no hay primer argumento | Código **2**, y la lista de applets disponibles en la salida de error |
| El primer argumento no corresponde a ningún applet | Código **2**, mensaje que **nombra el applet desconocido**, y la lista de applets disponibles en la salida de error |
| `n` no está registrado y el primer argumento es un verbo reservado (`version`) | Lo atiende el binario, **sin argumentos ni banderas**: cualquier cosa detrás de `version` —`kitlegal version extra`, `kitlegal version --json`— es código **2** con un mensaje que nombra lo que sobra, como en H0 (FR-027, [`cli-version.md`](../../001-h0-esqueleto-del-repo/contracts/cli-version.md)) |

**Garantía de indistinguibilidad** (FR-007, SC-003): invocar un applet por enlace simbólico produce la
**misma salida estándar, la misma salida de error y el mismo código de salida** que invocarlo con el
applet como primer argumento —salvo los datos que dependen del instante de la consulta—. Es lo que permite
que una skill siga escribiendo `scripts/boe articulo …` sin saber que por debajo hay un solo ejecutable.

**Un nombre de enlace desconocido no inutiliza el binario**: `kitlegal-dev -> kitlegal` se comporta como si
se hubiera invocado por su nombre propio.

---

## 2 bis. Qué verbo se ejecuta

Resuelto el applet, y **antes** de analizar la línea de órdenes, el kernel decide el verbo con una regla
de tres líneas, fija e independiente del orden de las banderas:

| Invocación | Verbo que se ejecuta |
|---|---|
| El primer argumento restante nombra un verbo del applet | ese verbo |
| No lo nombra y el applet declara verbo por omisión | el verbo por omisión; el argumento se le entrega tal cual |
| No lo nombra y el applet **no** declara verbo por omisión | ninguno: código **2**, con la lista de verbos del applet en la salida de error |
| Se pidió `--help` del applet | ninguno: la ayuda enumera los verbos del applet (§4) |

Ejemplos, con `echo` (un verbo, `repetir`, por omisión) y el segundo applet de ejemplo (dos verbos,
ninguno por omisión):

| Invocación | Resultado |
|---|---|
| `kitlegal echo hola` | verbo `repetir`, mensaje `hola` *(entrega literal del hito)* |
| `kitlegal echo repetir hola` | idéntico al anterior |
| `kitlegal echo repetir` | verbo `repetir` sin mensaje: **si el primer argumento nombra un verbo, es el verbo** |
| `kitlegal echo --json hola` | verbo `repetir`, mensaje `hola` |
| `kitlegal echo --help` | ayuda del applet: enumera sus verbos, código 0 |
| `kitlegal <segundo> hola` | código **2**: ese applet exige nombrar verbo |

Esa ambigüedad —un argumento que se llama como un verbo— está resuelta por escrito y en un solo sentido;
quien necesite lo contrario dispone del terminador `--`.

---

## 3. Los dos registros

El binario `kitlegal` que se distribuye y el binario que compila el test e2e usan **exactamente el mismo
mecanismo de registro**; lo único que cambia es qué applets se registran.

| Registro | Contiene en H1 | Consecuencia |
|---|---|---|
| **De producción** (`cmd/kitlegal`) | **Vacío**. El primer applet de producto llega en H4 | `kitlegal echo hola` sobre el binario distribuido termina con código **2**, applet desconocido, como cualquier otro nombre no registrado |
| **De ejemplo** (solo el binario del e2e) | `echo` —un verbo, `repetir`, por omisión— y un segundo applet con dos verbos y ninguno por omisión | La entrega literal del hito —`kitlegal echo hola --json`, `ln -s kitlegal echo && ./echo hola`, `--describe`— se demuestra aquí, y el segundo applet cubre la rama en que el verbo es obligatorio |

El applet `echo` es **un ejemplo del patrón, no funcionalidad del producto**: no accede a la red, no toca
disco, y su `data` se deriva de sus argumentos. Su definición vive en `internal/app/testdata`.

---

## 4. Ayuda derivada del registro

`--help` en el binario enumera los applets registrados con su descripción; `--help` en un applet enumera
sus verbos. Ambas listas salen del registro: **no existe ninguna lista duplicada mantenida a mano**, de
modo que registrar un applet lo hace aparecer en la ayuda sin tocar nada más.

---

## 5. `--describe`

Emite en la salida estándar **un** esquema JSON válido (borrador 2020-12) que describe la **entrada** y la
**salida** del applet invocado, y termina con código **0** sin ejecutar la operación: describirse y actuar
son excluyentes. Describe siempre **un verbo concreto**: el nombrado o, si no se nombró ninguno, el verbo
por omisión que resuelve §2 bis.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "kitlegal echo",
  "description": "…",
  "properties": {
    "entrada": { "…": "los argumentos del verbo, más las ocho banderas globales" },
    "salida":  { "…": "el sobre, con `data` condicionado a `ok`" }
  }
}
```

**Reglas:**

1. **La parte de salida describe el sobre completo**, con `data` **condicionado a `ok`**: el `data` propio
   del applet cuando `ok` es verdadero, y la forma común de error del kernel cuando es falso. La forma
   exacta del esqueleto está en [`sobre-de-salida.md`](./sobre-de-salida.md) §6.
2. **Una ejecución fallida de ese mismo applet valida contra el esquema que ese applet emite** (SC-015).
   Un test de contrato sobre una ejecución fallida no puede rechazar una salida conforme.
3. **El esquema se deriva de la definición del applet y de los tipos del sobre; no se mantiene a mano**: un
   cambio en los verbos o en los argumentos se refleja en el esquema sin que nadie lo edite.
4. El esquema es válido para un validador de esquemas: SC-007 lo comprueba compilándolo.

**Consumidores previstos** —y razón de que exista una sola definición—: las *tools* del servidor MCP
(H22), la tabla de comandos de cada `SKILL.md` (H11) y cualquier agente que descubra el binario. Una
definición, tres consumidores.

---

## 6. Qué NO forma parte de este contrato en H1

- **El directorio `schemas/`**, su generación desde `--describe` y la comprobación de deriva en CI:
  **H11** (borrador en H4). H1 emite el esquema por applet; **no lo persiste** ni lo compara con nada
  commiteado.
- **La generación de tools MCP** y de tablas de comandos: H22 y H11.
- **El applet `echo` en el binario distribuido**: no se registra, ni en H1 ni después.
- **`Emit(ctx) []GraphOp`**: los applets no emiten operaciones de grafo hasta H12.
- **La interfaz `Source`** (`Name`, `Fetch`, `TTL`, `Terms`): nace con el primer adaptador, en H4.
