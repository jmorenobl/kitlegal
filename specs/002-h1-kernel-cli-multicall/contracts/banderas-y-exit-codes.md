# Contrato: banderas globales, ayuda y códigos de salida

Lo que un agente o una persona puede dar por supuesto en **cualquier** applet de `kitlegal`, sin leer su
documentación.

**Requisitos que lo definen**: FR-018 … FR-037, FR-045, FR-049. **Criterios**: SC-002, SC-006, SC-010,
SC-011, SC-014.

---

## 1. Las ocho banderas globales

Las ofrece el kernel a **todos** los applets, con idéntica sintaxis y semántica, **sin que ningún applet
las declare**. Un applet nuevo que solo declara su nombre, sus verbos y su `data` las hereda todas.

| Bandera | Valor | Por omisión | Semántica en H1 |
|---|---|---|---|
| `--json` | — | falso | Salida legible por máquina: el sobre en JSON. En su ausencia, tabla mínima legible por una persona |
| `--timeout` | duración (`30s`, `1m`, `500ms`) | **`30s`** | Límite de **toda** la operación. Agotarlo → código **4**. Valor no positivo o con formato inválido → código **2** |
| `--offline` | — | falso | Declara que la operación no puede acceder a la red. Se acepta y se propaga al applet. **Su semántica completa —responder solo desde caché— no existe hasta H3** |
| `--dry-run` | — | falso | Describe la operación sin realizarla. Ver §3 |
| `--describe` | — | falso | Emite el esquema JSON de entrada y salida y **excluye la ejecución**. Código 0 |
| `--no-graph` | — | falso | Declara que la ejecución no altera el grafo. Se acepta y se propaga. **Sin efecto observable hasta H12** |
| `--asunto` | cadena | vacía | Declara sobre qué asunto se trabaja. Se acepta y se propaga. **No abre ni crea nada hasta H14** |
| `--verbose` | — | falso | Aumenta el detalle del registro de eventos en la salida de error. **No altera la salida estándar en absoluto** |

**Sobre las tres banderas sin objeto todavía** (`--offline`, `--no-graph`, `--asunto`): H1 fija su
sintaxis y su presencia en el juego común y las hace llegar al applet en el Contexto de ejecución. **No se
les inventa una semántica que ningún hito ha definido**, y su comportamiento actual —aceptar y propagar—
es explícito, no un silencio engañoso.

---

## 2. Ayuda

- `--help` funciona en el binario y en cada applet, y termina con código **0**. La forma corta `-h`,
  que la ayuda de un verbo anuncia como `-h, --help`, pide la ayuda en las mismas tres posiciones —el
  binario, el applet y el verbo— con el mismo resultado; es la única forma corta del juego común.
- La lista de applets y de verbos que muestra **procede del registro de applets**; no existe ninguna lista
  paralela mantenida a mano.
- La ayuda se emite **siempre como texto para personas en la salida estándar**. `--json` **no la altera**:
  `--json --help` se comporta exactamente igual que `--help` solo, en cualquier orden.
- La contraparte legible por máquina de la ayuda es `--describe`, no `--json`.
- Pedir la ayuda de un applet **no** resuelve ningún verbo: `kitlegal echo --help` enumera los verbos del
  applet en lugar de describir uno solo
  ([`registro-y-describe.md`](./registro-y-describe.md) §2 bis).
- La ayuda es un mensaje para la persona: se escribe siempre, **sin depender del nivel de registro** (§6).

**Precedencia entre banderas excluyentes**, fija e independiente del orden de escritura:

1. `--help` gana sobre todo lo demás.
2. `--describe` gana sobre la ejecución.
3. En otro caso, se ejecuta.

---

## 3. `--dry-run`

- La descripción de la operación que se habría realizado va a la **salida de error**, y es **siempre
  visible**: **no es un registro de eventos**, sino un mensaje dirigido a la persona, así que no pasa por
  `slog` y **ningún nivel puede ocultarla** —ni `--verbose`, ni `KITLEGAL_LOG`, ni el umbral por omisión—.
  Un requisito de visibilidad incondicional no se implementa sobre un mecanismo filtrable (§6).
- La **salida estándar queda vacía**, también si se añade `--json`. El sobre es la cita de un contenido
  consultado, y una operación no realizada no tiene nada que citar.
- El código de salida es **0**.
- El kernel **no corta** la ejecución antes de entregar el control al applet: la bandera viaja en el
  Contexto de ejecución —igual que `--offline`, `--no-graph` y `--asunto`— y cada capa con efectos la
  honra describiendo en lugar de ejecutar. En H1, donde ninguna capa tiene efectos, la descripción
  (applet, verbo y argumentos que se habrían ejecutado) la escribe el kernel y el presentador no se
  invoca.

---

## 4. Códigos de salida

Tabla **estable**. Un agente clasifica el fallo por el código, sin analizar el mensaje.

| Código | Clase | Significado | Qué debería hacer un agente |
|---|---|---|---|
| **0** | — | Correcto | Continuar |
| **1** | `inesperado` | Fallo que no encaja en ninguna clase prevista | Detenerse y reportar: es un defecto, no una condición esperable |
| **2** | `argumentos` | Bandera desconocida, valor con formato inválido, argumento obligatorio ausente, applet no registrado o no deducible | Corregir la invocación. Reintentar igual no sirve |
| **3** | `no-encontrado` | Lo pedido no existe | Cambiar de estrategia; no reintentar |
| **4** | `fuente-no-disponible` | La fuente no responde, o se agotó `--timeout` | Reintentar más tarde |
| **5** | `limite-o-tos` | Límite de peticiones, o restricción de términos de uso | Esperar; **nunca** rodear el límite |
| **6** | `identidad-humana` | La acción requiere identidad humana | **Detenerse y pedir a una persona que actúe.** No se ha realizado ninguna acción con efectos externos |

**0, 2, 3, 4, 5 y 6 son una decisión cerrada** del proyecto (`CLAUDE.md`) y no se reabren. El **1** lo
introduce H1 para lo que FR-031 llama «fallo inesperado»: es el «error general» de la convención Unix, no
colisiona con ningún reservado, y **queda reservado para eso**; ningún hito posterior debe reutilizarlo
para otra cosa.

**Garantías de la traducción:**

- Vive en **un único lugar**, y es **exhaustiva**: una clase de error nueva sin código asignado hace
  fallar el análisis estático, no pasa inadvertida.
- **Envolver un error no cambia su clase**: añadir contexto con `%w` deja intacto el código resultante.
- Un error que no corresponda a ninguna clase prevista sale con **1**, nunca con 0 ni con un reservado: un
  fallo inesperado no puede confundirse con un éxito ni con un fallo clasificado.
- **Ningún camino de usuario termina en pánico.** Los fallos se propagan como error y salen por el único
  punto que termina el proceso.
- Escribir en la salida estándar puede fallar (tubería cerrada): eso también se traduce a un código de
  salida, nunca a un pánico. En concreto: el presentador **comprueba y propaga** el error de toda
  escritura; el fallo se clasifica como `inesperado` y el proceso termina con **1**; y **no se intenta
  emitir un segundo sobre** por el descriptor que acaba de fallar —solo el mensaje para la persona en la
  salida de error, y si esa también falla, únicamente el código—. Esta garantía tiene control mecánico
  propio: la exclusión de `errcheck` para `fmt.Fprint*` que H0 dejó para su punto de entrada **se retira
  en H1**, porque taparía justamente este camino.

---

## 5. Separación de descriptores

| Descriptor | Qué lleva, siempre |
|---|---|
| **Salida estándar** | Exclusivamente lo que emite el presentador: el sobre (con `--json`), la tabla mínima, la ayuda o el esquema de `--describe` |
| **Salida de error** | Exclusivamente el registro de eventos y los mensajes dirigidos a la persona |

- Con `--json`, la salida estándar contiene **exactamente un documento JSON y nada más**: ni una línea de
  registro, ni un aviso, ni una cabecera, **ni siquiera con el registro de eventos al máximo detalle**.
  Esta regla rige sobre la presentación del resultado de un applet; no sobre la ayuda (texto para
  personas) ni sobre `--dry-run` (salida estándar vacía).
- El mensaje de un fallo va a la salida de error **y además**, con `--json`, en forma estructurada dentro
  del sobre de fallo.
- **Sin `--json`, un fallo deja la salida estándar vacía**: el mensaje va a la salida de error y no se
  presenta ninguna tabla. Un resultado que no existe no se cita, por la misma razón por la que `--dry-run`
  no emite sobre.
- El binario garantiza que escribe cada cosa en su descriptor. Si quien invoca redirige la salida de error
  a la estándar, la separación se pierde por decisión suya, no del binario.

**Un solo escritor, dos canales.** Todo lo que el binario escribe —el sobre, la tabla, la ayuda, el texto
de `version`, el mensaje de un fallo, la descripción de `--dry-run`— pasa por `internal/render`, que
recibe los dos descriptores de la raíz de composición. `log/slog` escribe en el mismo descriptor de error,
pero solo el registro de eventos. Los descriptores del sistema (`os.Stdout`, `os.Stderr`) no se nombran en
ningún otro sitio del árbol.

---

## 6. Registro de eventos

| Aspecto | Regla |
|---|---|
| Destino | **Siempre** la salida de error, nunca la estándar |
| Nivel | `KITLEGAL_LOG` (`debug`, `info`, `warn`, `error`) tiene prioridad; luego `--verbose` (equivale a `debug`); por omisión, `warn` |
| Formato | Estructurado, para que un consumidor automático pueda filtrarlo |
| Privacidad | **No incluye por omisión contenido que identifique a una persona física.** Registra applet, verbo, duración y clase de error; los argumentos, solo en `debug` |

**Qué es registro y qué no.** La salida de error lleva dos cosas distintas, y la diferencia es contractual:

| | Registro de eventos | Mensaje para la persona |
|---|---|---|
| Qué | Applet, verbo, duración, clase de error; detalle técnico | Mensaje de un fallo, lista de applets disponibles, descripción de `--dry-run`, aviso de un `KITLEGAL_LOG` inválido |
| Por dónde | `log/slog` | El presentador |
| ¿Lo puede ocultar un nivel? | **Sí**: para eso está el nivel | **No, nunca** |

De ahí que el nivel por omisión pueda ser `warn` sin que se pierda nada que el contrato prometa visible:
lo que tiene que verse siempre no viaja por un canal filtrable.

Un valor inválido de `KITLEGAL_LOG` no aborta la invocación: se ignora y el binario avisa por el canal no
filtrable —precisamente porque el valor inválido no puede silenciar su propio aviso—. Una variable de
entorno mal escrita no es un error de argumentos de la línea de órdenes.
