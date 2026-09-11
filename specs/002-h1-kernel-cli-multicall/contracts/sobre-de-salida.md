# Contrato: el sobre de salida

Interfaz que H1 expone a **todo consumidor automático** del binario: una skill, el servidor MCP de H22, un
agente que descubre el binario, o el propio proyecto en H4 cuando valide una cita. Es el contrato más
estable del proyecto: si cambia, cambia todo lo que cita.

**Requisitos que lo definen**: FR-010 … FR-017, FR-045. **Criterios**: SC-001, SC-005, SC-014, SC-015.

---

## 1. Forma

Un **único** documento JSON, con **exactamente** seis claves en el nivel superior, ni una más ni una
menos, tanto en éxito como en fallo.

```json
{
  "ok": true,
  "fuente": "kitlegal.echo",
  "url": "kitlegal:applet/echo",
  "fecha_consulta": "2026-09-11T10:12:00+02:00",
  "hash": "sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
  "data": { "mensaje": "hola" }
}
```

Fallo, con la misma forma y la causa dentro de `data`:

```json
{
  "ok": false,
  "fuente": "kitlegal.cli",
  "url": "kitlegal:cli",
  "fecha_consulta": "2026-09-11T10:12:00+02:00",
  "hash": "sha256:…",
  "data": { "clase": "argumentos", "mensaje": "bandera desconocida: --jsno" }
}
```

---

## 2. Reglas de cada clave

| Clave | Tipo | Regla |
|---|---|---|
| `ok` | booleano | Verdadero **si y solo si** el código de salida del proceso es 0 |
| `fuente` | cadena | No vacía. Espacio de nombres reservado `kitlegal.…` cuando no hay fuente externa |
| `url` | cadena | No vacía y **URI absoluto**. Esquema reservado `kitlegal:` cuando no hay fuente externa |
| `fecha_consulta` | cadena | RFC 3339 **con desplazamiento horario explícito**. No es texto libre |
| `hash` | cadena | `sha256:` + 64 dígitos hexadecimales en minúscula |
| `data` | cualquiera | El contenido del applet si `ok`; `{clase, mensaje}` si no |

**Ningún consumidor debe asumir que existirá una séptima clave, ni que alguna de las seis podrá faltar.**

---

## 3. Procedencia: el espacio de nombres reservado

`fuente` y `url` identifican la procedencia comprobable del contenido y **nunca van vacías**. Un applet que
no consulta ninguna fuente externa usa el espacio reservado:

| Emisor | `fuente` | `url` |
|---|---|---|
| Applet calculado (`echo` en H1; previsiblemente `cita` en H8 y `plazos` en H9) | `kitlegal.<nombre>` | `kitlegal:applet/<nombre>` |
| El kernel, cuando falla antes de llegar al applet | `kitlegal.cli` | `kitlegal:cli` |
| Adaptador de fuente pública (H4 en adelante) | nombre de la fuente, p. ej. `boe.legislacion-consolidada` | URL `http(s)` comprobable |

**Regla para el consumidor**: un sobre cuya `url` empieza por `kitlegal:` es un **resultado calculado, no
una cita de fuente pública**. No debe presentarse como fuente ni usarse para fundamentar una afirmación
legal.

**Regla para el productor**: ningún adaptador de `internal/source/<fuente>` puede usar el prefijo
`kitlegal.` ni el esquema `kitlegal:`. En H1 no hay adaptadores; la comprobación mecánica nace con el
primero, en H4.

---

## 4. La huella

```
hash = "sha256:" + hex( sha256( canónico(data) ) )
```

**Forma canónica de `data`**, en tres pasos:

1. Serializar `data` a JSON.
2. Volver a leerlo a una representación genérica **conservando los números como literales** (sin pasar por
   coma flotante).
3. Volver a serializarlo **con las claves de todo objeto ordenadas**, sin espacios ni saltos de línea, y
   sin escapar caracteres HTML.

**Propiedades garantizadas** (SC-005, comprobadas en test):

- El mismo contenido produce siempre la misma huella, **con independencia del orden en que se serializaron
  las claves** y del momento de la consulta.
- Un contenido que difiere en un solo byte produce una huella distinta.
- La huella **no** depende de `fecha_consulta`: por eso sirve para detectar que un contenido cambió.

**El prefijo del algoritmo es obligatorio** (FR-012): permite cambiar de algoritmo sin romper a quien lee.
Un consumidor debe comprobar el prefijo antes de interpretar el resto.

**Si `data` no es serializable**, la construcción del sobre falla **antes** de escribir nada en la salida
estándar: nunca se emite un sobre a medias. El fallo se clasifica como `inesperado` (código 1).

**Si falla la escritura** del sobre ya construido (tubería cerrada), el error se propaga y sale también
como `inesperado` (código 1), y **no se intenta un segundo sobre** por el descriptor roto
([`banderas-y-exit-codes.md`](./banderas-y-exit-codes.md) §4). Vale también con una tubería del sistema
cuyo lector ha terminado: el binario ignora `SIGPIPE`, así que termina con el código y no por la señal.

---

## 5. El sobre de fallo

Lo emite **el kernel**, no el applet, desde el único punto que traduce el error a código de salida. Por
eso es idéntico para todos los applets y también para los fallos anteriores a la ejecución del applet
(bandera desconocida, applet no registrado).

- `ok` es falso **si y solo si** el código de salida no es 0.
- `data` es exactamente `{"clase": …, "mensaje": …}`, sin claves adicionales.
- `clase` es una de: `argumentos`, `no-encontrado`, `fuente-no-disponible`, `limite-o-tos`,
  `identidad-humana`, `inesperado`, y **corresponde al código de salida emitido**
  ([`banderas-y-exit-codes.md`](./banderas-y-exit-codes.md)).
- `mensaje` es el mensaje dirigido a la persona: **el mismo** que va a la salida de error.
- `fuente` y `url` son las de la fuente que se estaba consultando cuando se conocen —una URL que devolvió
  «no encontrado» es una cita negativa útil—, y las del espacio reservado del kernel cuando no.
- `fecha_consulta` y `hash` se calculan igual que en un sobre de éxito, con la huella sobre la forma
  canónica del `data` de error.

**El sobre de fallo no sustituye al código de salida**: lo duplica en forma estructurada. El código de
salida sigue siendo la vía primaria de clasificación, y el mensaje sigue yendo además a la salida de
error.

### Cómo se sabe que se pidió `--json` cuando el fallo es anterior al análisis

Dos de los casos que este contrato obliga a cubrir —**bandera desconocida** y **applet no registrado**—
ocurren antes de que la línea de órdenes esté analizada; en el segundo, ni siquiera existe la gramática,
porque el binario construye la gramática después de resolver el applet. La forma del sobre de fallo no
puede depender de un dato que todavía no existe, así que el kernel lo obtiene de un **pre-escaneo acotado
de los argumentos**, declarado y probado
([research.md D25](../research.md#d25--pre-escaneo-acotado-de-argv-antes-de-la-gramática)):

- Reconoce exactamente `--json`, `--verbose` y `--help` en su forma larga, con o sin valor booleano
  explícito, y una sola forma corta, el token exacto `-h` (la que la ayuda de un verbo anuncia como
  `-h, --help`; ni agrupada, ni con valor, ni en mayúscula); se detiene en el terminador `--`; ignora
  todo lo demás; **nunca falla ni consume argumentos**.
- **No decide qué se ejecuta**: solo la forma en que se presenta un fallo temprano, el nivel del registro
  y si se suprime el verbo por omisión.
- En cuanto el análisis termina bien, **manda lo analizado**; el pre-escaneo es un valor provisional. Un
  test de tabla comprueba que ambos coinciden para toda invocación bien formada.

Si el pre-escaneo **no** ve `--json`, el fallo se presenta como cualquier otra salida para personas: el
mensaje en la salida de error y la **salida estándar vacía**. No hay una tercera forma de presentar un
fallo.

---

## 6. Descripción formal

El sobre es validable contra un esquema JSON (borrador 2020-12) que el propio binario emite con
`--describe` ([`registro-y-describe.md`](./registro-y-describe.md)). Su esqueleto:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "additionalProperties": false,
  "required": ["ok", "fuente", "url", "fecha_consulta", "hash", "data"],
  "properties": {
    "ok":             { "type": "boolean" },
    "fuente":         { "type": "string", "minLength": 1 },
    "url":            { "type": "string", "minLength": 1, "format": "uri" },
    "fecha_consulta": { "type": "string", "format": "date-time" },
    "hash":           { "type": "string", "pattern": "^sha256:[0-9a-f]{64}$" },
    "data":           true
  },
  "if":   { "properties": { "ok": { "const": true } } },
  "then": { "properties": { "data": { "$ref": "#/$defs/<data del applet>" } } },
  "else": { "properties": { "data": { "$ref": "#/$defs/DatosError" } } },
  "$defs": {
    "DatosError": {
      "type": "object",
      "additionalProperties": false,
      "required": ["clase", "mensaje"],
      "properties": {
        "clase": { "type": "string",
                   "enum": ["argumentos", "no-encontrado", "fuente-no-disponible",
                            "limite-o-tos", "identidad-humana", "inesperado"] },
        "mensaje": { "type": "string", "minLength": 1 }
      }
    }
  }
}
```

**Nota para quien escriba el validador**: las aserciones de `format` **no** están activas por omisión en el
borrador 2020-12. Un validador que no las active explícitamente aceptará un `url` vacío o no-URI, y el
control quedará muerto aunque el test esté en verde. Los tests de H1 las activan siempre.

**Nota sobre los nombres de `$defs`**: `DatosError` se llama así porque es un tipo del sobre; el `data` del
applet se nombra con su paquete delante (`ejemplo.mensaje`, `boe.Articulo`), de modo que un applet que
declare su propio `DatosError` no pise el del kernel. Dos tipos distintos que aun así acabaran con el mismo
nombre hacen fallar la construcción del esquema (código 1): nunca se emite un esquema que describa otra
cosa.

---

## 7. Presentación legible

Sin `--json`, el mismo contenido se presenta en una tabla mínima que **conserva los cuatro datos de
procedencia** —`fuente`, `url`, `fecha_consulta` y `hash`— además del contenido, para que la cita no se
pierda por elegir el formato legible (FR-043):

```
fuente          kitlegal.echo
url             kitlegal:applet/echo
fecha_consulta  2026-09-11T10:12:00+02:00
hash            sha256:2cf24dba…
mensaje         hola
```

El contenido de `data` se aplana a pares `ruta<TAB>valor`, donde la ruta concatena claves e índices con
`.` (`articulos.0.titulo`). Añadir una forma de presentación nueva no obliga a modificar ningún applet.

**La tabla mínima es la forma de un resultado, no de un fallo**: cuando la operación falla y no se pidió
`--json`, la salida estándar queda vacía y el mensaje va a la salida de error (§5).

---

## 8. Qué NO forma parte de este contrato en H1

- El directorio `schemas/*.json` versionado y su comprobación de deriva: **H11** (borrador en H4). H1
  emite el esquema; no lo persiste ni lo compara con nada commiteado.
- La presentación en markdown: **H11**.
- Cualquier campo adicional del sobre. Las seis claves son cerradas.
