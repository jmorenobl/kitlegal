# Contrato: `kitlegal graph` tras H7.1

Lo que cambia del contrato de H7 (`specs/010-h7-internal-graph-grafo/contracts/applet-graph.md`, que no se edita);
todo lo que no se dice aquí sigue igual: procedencia del sobre (`kitlegal.graph`, `kitlegal:applet/graph`, el instante
del reloj), `show` y `stats` con `--json`, las plantillas de las explicaciones y los códigos de `show` y `stats`.
FR-001 a FR-015, FR-020 a FR-032, FR-060 a FR-064. Modelo: [../data-model.md](../data-model.md). Decisiones:
[../research.md](../research.md) D5-D11.

## 1. Invocación de `check`

```text
kitlegal graph check [<norma> [<bloques>...]]
```

- **Descripción** (verbo, ayuda, `--describe` y tabla de comandos de `SKILL.md`; FR-007), con el `50` escrito desde
  `grafo.MaximoDeHallazgos`:
  `Comprueba lo consultado de una norma, de algunos de sus bloques o, sin argumentos, todo lo consultado, y lista como mucho 50 hallazgos: redacciones que han cambiado desde la lectura anterior y consultas caducadas.`
- Argumentos de posición, los dos opcionales (research D6): `norma` —ayuda
  `Identificador BOE de la norma, BOE-A-<año>-<número>; sin él, todo lo consultado.`— y `bloques` —ayuda
  `Ids de bloque de esa norma, como a21; sin ellos, todos los suyos.`—. La línea de uso de la ayuda de Kong es
  `Usage: graph check [<norma> [<bloques> ...]] [flags]` (research V12); la tabla de comandos escribe
  `kitlegal graph check [<norma> [<bloques>...]]` (research D7).
- `--describe`: la entrada gana `norma` (`string`) y `bloques` (`array` de `string`), ninguna en `required`; la
  salida, `data` con `$ref` a `grafo.Comprobacion` (§3). `schemas/grafo.json` se regenera (FR-082).

## 2. Validación y códigos de `check`

Antes de abrir nada (FR-004, FR-006):

| Entrada | Código | Clase | Mensaje (salida de error; con `--json`, también en `data.mensaje`) |
|---|---|---|---|
| norma dada fuera de `^BOE-A-[0-9]{4}-[0-9]{1,9}$` (`boe.ValidarNorma`), p. ej. `a21`, `ine:28074`, `BOE-B-2015-10565` o `""` | 2 | `argumentos` | `argumentos inválidos: la norma "<valor>" no tiene la forma BOE-A-<año>-<número>, con cuatro dígitos en el año y de uno a nueve en el número` |
| un bloque vacío o de solo espacio en blanco (`unicode.IsSpace` en todos sus caracteres) | 2 | `argumentos` | `argumentos inválidos: el bloque "<valor>" está vacío o solo tiene espacio en blanco` |
| norma bien formada que el grafo no conoce; bloque que la norma no tiene; `world.db` ausente, de 0 bytes o sin esquema | 0 | — | — (ningún hallazgo de lo que no conoce; FR-003) |
| `world.db` que no se puede usar (cualquier causa, p. ej. un fichero que no es una base SQLite) | 1 | `inesperado` | `grafo: "<ruta>" no es una base de datos utilizable: <causa>` (FR-070) |
| esquema de una versión posterior | 1 | `inesperado` | `grafo: "<ruta>" tiene el esquema en la versión N y este binario conoce la 2: no se modifica` (H7 FR 012) |
| plazo de `--timeout` agotado esperando la base / espera propia agotada | 4 / 1 | como en H7 FR 014 | como en H7 |

Los errores de argumentos los firma el applet (su procedencia), no el kernel.

## 3. `data` de `check` (FR-012)

Claves, en este orden, siempre presentes y nunca `null`:

| Clave | Tipo | Qué es |
|---|---|---|
| `norma` | cadena | la norma pedida; `""` sin argumentos (todo lo consultado) |
| `bloques` | lista de cadenas | los bloques pedidos, en su orden; `[]` si ninguno |
| `version-obsoleta` | entero | total de hallazgos de esa clase en el ámbito, contando los omitidos |
| `fuente-caducada` | entero | ídem |
| `omitidos` | entero | `version-obsoleta` + `fuente-caducada` − número de listados |
| `hallazgos` | lista de hallazgos | los listados, como mucho 50, en el orden de §4; `[]` si ninguno. Cada uno, la forma de H7 FR 061 sin cambios |

Ejemplo (paso 2 de la secuencia de FR-025, `graph check BOE-A-2015-10565 a21 --json` con el reloj T1; una línea):

```json
{"ok":true,"fuente":"kitlegal.graph","url":"kitlegal:applet/graph","fecha_consulta":"2026-09-29T12:00:00Z","hash":"sha256:…","data":{"norma":"BOE-A-2015-10565","bloques":["a21"],"version-obsoleta":1,"fuente-caducada":0,"omitidos":0,"hallazgos":[{"clase":"version-obsoleta","id":"eli/es/l/2015/10/01/39#a21@20161002:sha256:98d3b9d4686a3155f48641841df7023e2c17a693fcd0e11d6beef2615abcec7c","explicacion":"La versión de [BOE-A-2015-10565, bloque a21] con fecha de vigencia 20161002 está superada por la de fecha de vigencia 20250101, observada en https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21 el 2026-09-28T12:00:00Z.","procedencia":{"fuente":"boe.legislacion-consolidada","url":"https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21","fecha_consulta":"2026-09-28T12:00:00Z"},"fecha_vigencia":"20161002","fecha_vigencia_reciente":"20250101"}]}}
```

**Bytes** (medidos con `wc -c`, research D9): sobre sin `data` y con el salto final, 197; `data` sin hallazgos, 115 con
un bloque, 139 con cinco, 102 sin argumentos; un `version-obsoleta` del art. 21 de la LPAC, 691; un `fuente-caducada`
de su redacción, 655.

## 4. Reglas, orden y cota

- `version-obsoleta` y `fuente-caducada`: data-model §5. Plantillas de `explicacion` sin cambios (H7):
  `La versión de <cita> con fecha de vigencia <fecha> está superada por la de fecha de vigencia <fecha>, observada en <url> el <fecha_consulta>.`
  y
  `La consulta de <cita> a <fuente> en <url> del <fecha_consulta> tenía una vigencia de <N> s y caducó el <instante>.`
  En `version-obsoleta`, la «más reciente» es la redacción que vio la última lectura y su `procedencia`, la última
  observación de esa redacción.
- Orden: todos los `version-obsoleta` antes que todos los `fuente-caducada`; dentro de cada clase, por `id` comparando
  bytes (FR-011). Cota: los 50 primeros (FR-010). Mismo grafo, mismos argumentos y mismo instante, misma salida.
- Ámbito (FR-002): con norma, solo su `Norma` (por su `identificador`), sus `Bloque` —los nombrados, o todos— y las
  `BloqueVersion` de esos bloques. Sin argumentos, todo lo consultado (FR-005).
- `check` no escribe nada, tampoco en `lecturas` (FR-014).

## 5. Salida legible (sin `--json`; FR-060 a FR-063)

Reglas comunes: texto en `Resultado.Legible` compuesto a partir de `data` en el mismo instante; sin las líneas
`fuente`, `url`, `fecha_consulta` y `hash` del sobre ni el `kitlegal.graph` / `kitlegal:applet/graph` del sobre, sin
pares ruta/valor aplanados (`nodo.id`, `0.clase`…), sin tabuladores ni secuencias de escape; columnas alineadas con
espacios por el número de runas; termina en `\n`; con `--json`, nada cambia. Los fallos no tienen forma legible: su
mensaje va a la salida de error, como hasta ahora.

### 5.1 `stats`

```text
El grafo del mundo tiene 3 nodos, 2 aristas y 1 texto.

Nodos por tipo y fuente:
  Bloque         boe.legislacion-consolidada  1
  BloqueVersion  boe.legislacion-consolidada  1
  Norma          boe.legislacion-consolidada  1

Aristas por relación y fuente:
  eli:has_part     boe.legislacion-consolidada  1
  eli:has_version  boe.legislacion-consolidada  1
```

Singular y plural por el número («1 nodo», «0 nodos»). Una sección sin pares no se escribe; el grafo vacío es solo la
primera línea con tres ceros.

### 5.2 `show <id>`

```text
Bloque eli/es/l/2015/10/01/39#a21
  bloque: a21
Primera observación: 2026-09-28T12:00:00Z
Última observación: 2026-09-28T12:00:00Z · boe.legislacion-consolidada · https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21

Aristas salientes:
  eli:has_version → eli/es/l/2015/10/01/39#a21@20161002:sha256:98d3b9d4686a3155f48641841df7023e2c17a693fcd0e11d6beef2615abcec7c
    última observación: 2026-09-28T12:00:00Z · boe.legislacion-consolidada · https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21

Aristas entrantes:
  eli:has_part ← eli/es/l/2015/10/01/39
    última observación: 2026-09-28T12:00:00Z · boe.legislacion-consolidada · https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21
```

Primera línea: tipo e id. Debajo, un dato por línea (`  <clave>: <valor>`), por clave comparando bytes; un valor que
no es cadena, en JSON. Sin aristas en un sentido: `Aristas salientes: ninguna.` / `Aristas entrantes: ninguna.`.

### 5.3 `check`

Con hallazgos (`<ámbito>`: `en todo lo consultado`, `de BOE-A-2015-10565`, `de BOE-A-2015-10565, bloque a21`,
`de BOE-A-2015-10565, bloques a21 y a22`, `…, bloques a21, a22 y a23`):

```text
Hallazgos en todo lo consultado: 1 version-obsoleta y 3 fuente-caducada; se listan 4 y se omiten 0.

version-obsoleta (1):
  - La versión de [BOE-A-2015-10565, bloque a21] con fecha de vigencia 20161002 está superada por … .
    eli/es/l/2015/10/01/39#a21@20161002:sha256:98d3…
fuente-caducada (3):
  - La consulta de BOE-A-2015-10565 a boe.legislacion-consolidada en … caducó el 2026-10-05T12:00:00Z.
    eli/es/l/2015/10/01/39
  - …

Para acotar la comprobación a una norma y a sus bloques: kitlegal graph check <norma> [<bloque>...]
```

- La cabecera dice siempre el total de cada clase y cuántos se listan y se omiten, también cuando no se omite
  ninguno (`se omiten 0`; FR-063); con la medida, `Hallazgos en todo lo consultado: 240 version-obsoleta y 4590
  fuente-caducada; se listan 50 y se omiten 4780.`
- Un grupo por clase con algún hallazgo listado, `version-obsoleta` primero, con el total de la clase entre
  paréntesis; cada hallazgo, su explicación y, debajo, su id.
- Sin hallazgos: una sola frase, `No hay nada que volver a comprobar en todo lo consultado.` o
  `No hay nada que volver a comprobar de BOE-A-2015-10565, bloque a21.`
- Sin argumentos, con hallazgos o sin ellos, termina con la línea en blanco y
  `Para acotar la comprobación a una norma y a sus bloques: kitlegal graph check <norma> [<bloque>...]`.

## 6. Uso, de fuera adentro

Volumen de referencia: la medida de la bitácora (300 normas, 2 400 bloques, 240 con dos lecturas, el 90 % hace más de
una semana). Bytes con `--json`, de §3.

| Salida | Quién la pide y cuántas veces por pregunta | Tamaño | Cuándo deja de darse cada señal |
|---|---|---|---|
| `graph check <norma> <bloques leídos> --json` | `boe-legislacion`, paso 5: una vez por norma citada (una en la mayoría de las preguntas, una más por remisión); traslada cada `version-obsoleta` con la forma fija; no le llega ningún `fuente-caducada` de lo recién leído | como mucho k hallazgos con k bloques leídos: 312 bytes sin cambios (lo habitual); 197 + 139 + 5 × 691 + 4 = 3 795 con cinco bloques cambiados (a21-a25 de la LPAC); no depende de lo acumulado | `version-obsoleta`: la da la lectura que ve la redacción nueva y la apaga la lectura siguiente de ese bloque (FR-024) |
| `graph check [--json]` sin argumentos | una persona que repasa su memoria; ninguna skill | como mucho 50 hallazgos: con la medida, 4 830 contados (240 y 4 590), 50 listados, todos `version-obsoleta`: 197 + 102 + 50 × 685 + 49 = 34 598 bytes con los ids de la siembra de TestMedidaDelGrafo (un `version-obsoleta` de `BOE-A-2020-1299`, bloque `a1`, pesa 685; SC-001: ≤ 40 000) | `version-obsoleta`, como arriba; `fuente-caducada`: sobre la `Norma`, el `Bloque` y la redacción vista cuya consulta pasó su vigencia (7 días); la apaga la lectura siguiente del bloque, que ya no puede servir la caché (FR-031); nunca sobre una redacción superada |
| `graph check <norma> [--json]` | una persona que acota (H8 y H10, fuera de este hito) | como mucho 50; con 8 bloques por norma, como mucho 8 + 17 = 25 hallazgos, unos 17 KB | como arriba |
| `graph stats [--json]` | una persona | una línea por par (tipo, fuente) y (relación, fuente): hoy 5 + 3, ≈ 1 KB con cualquier volumen | no da señales |
| `graph show <id> [--json]` | una persona que depura | un nodo y sus aristas: una `Norma`, una por bloque consultado de ella (8 de media) | no da señales |

Sin `--json`, la salida legible de cada verbo lleva los mismos elementos que su `data` y ocupa menos: por hallazgo, su
explicación (≈ 300 bytes) y su id (≈ 100), así que `check` sin argumentos sobre la medida da como mucho unos
50 × 400 ≈ 20 KB y con la norma y los bloques de una pregunta, unos cientos de bytes; `stats`, una línea por par; `show`
de una `Norma`, dos líneas por arista (≈ 250 bytes cada una).
