# Contrato: los seis verbos del applet `boe`

Qué acepta cada verbo, qué pide a la fuente y en qué orden, qué lee y escribe en la caché, qué devuelve y con qué
procedencia. Los campos de `data` están en [data-model.md](../data-model.md) §2; los fallos, en
[errores-y-codigos.md](./errores-y-codigos.md). Decisiones en [research.md](../research.md) D5-D10.

## 0. Común a los seis

- **Invocación**: `kitlegal boe <verbo> <argumentos> [banderas]` o, por el enlace `boe -> kitlegal`,
  `boe <verbo> <argumentos> [banderas]`, con la misma salida byte a byte (FR-001). Sin verbo → 2.
- **Banderas**: las ocho del kernel, sin ninguna propia. `--json` elige el sobre en JSON; sin ella, la tabla mínima
  del kernel. `--timeout` acota la invocación entera (FR-095). `--offline` abre la caché en solo lectura y no pide
  nada (FR-092). `--dry-run` abre la caché en solo lectura, no pide nada, describe cada petición que habría
  emitido y termina con 0 (FR-094). `--describe` emite el esquema del verbo sin ejecutar nada. `--no-graph` y
  `--asunto` no cambian nada (FR-128). `--verbose` hace visibles los eventos de `httpx`, de la caché y de `boe`.
- **Sobre de éxito**: `fuente` = `boe.legislacion-consolidada`; `url` = la de la API que se indica en cada verbo;
  `fecha_consulta` = la de la consulta (reciente o guardada; si hay varias, la más antigua); `hash` = huella de
  `data` (FR-002, FR-096).
- **Ensayo**: tras la línea del kernel, una línea `--dry-run: se habría pedido GET <dirección>` por cada petición
  que se habría emitido, sin repetir ninguna (ADR 0011, FR-020).
- **Base de la API** = `https://www.boe.es/datosabiertos/api/legislacion-consolidada`.

## 1. `buscar <texto>…`

| Aspecto | Contrato |
|---|---|
| Argumentos | uno o más; se unen con un espacio |
| Validación | sin operadores y sin ninguna palabra → 2 sin abrir la caché (FR-030) |
| Consulta | con ` AND `, ` OR `, ` NOT `, `titulo:`, `materia:` o `"` → tal cual; si no → `titulo:<p1> AND titulo:<p2> …` ([research.md](../research.md) D8) |
| Petición | `GET <base>?limit=10&query=<quote(json.dumps({"query": {"query_string": {"query": <consulta>}}}))>`, `Accept: application/json` |
| Caché | clave `boe.legislacion-consolidada\|1\|buscar\|<dirección>`, 300 s; se escribe también sin resultados (FR-032) |
| `data` | `[]ResultadoDeBusqueda` (`[]` sin resultados) |
| `url` del sobre | la dirección de la petición |

Ejemplo: `kitlegal boe buscar procedimiento administrativo común --json` pide

```
GET https://www.boe.es/datosabiertos/api/legislacion-consolidada?limit=10&query=%7B%22query%22%3A%20%7B%22query_string%22%3A%20%7B%22query%22%3A%20%22titulo%3Aprocedimiento%20AND%20titulo%3Aadministrativo%20AND%20titulo%3Acom%5Cu00fan%22%7D%7D%7D
```

## 2. `indice <norma>`

| Aspecto | Contrato |
|---|---|
| Validación | norma (FR-081) |
| Petición | `GET <base>/id/<norma>/texto/indice`, `Accept: application/json` |
| Caché | `…\|1\|indice\|<dirección>`, 604 800 s |
| `data` | `Indice{norma, url, bloques}`; índice anidado o plano (FR-040); `data` vacío → 3 (FR-041) |
| `url` del sobre | la dirección de la petición |

## 3. `articulo <norma> <bloque>`

| Aspecto | Contrato |
|---|---|
| Validación | norma (FR-081) y bloque (FR-080) |
| Peticiones, en orden | 1) `GET <base>/id/<norma>/texto/bloque/<bloque>`, `Accept: application/xml`; 2) solo si el bloque se obtuvo **y** no hay entrada vigente de metadatos: `GET <base>/id/<norma>/metadatos`, `Accept: application/json` (FR-013, FR-090) |
| Caché que lee | `…\|1\|articulo\|<dirección del bloque>`; si falta, `…\|1\|metadatos\|<dirección de metadatos>` |
| Caché que escribe | la de metadatos (300 s) si los pidió y se obtuvieron; la del artículo (604 800 s) al terminar bien |
| `data` | `Articulo` |
| `url` del sobre | la dirección del **bloque** |
| `fecha_consulta` | mín(instante del bloque, fecha de los metadatos usados); servido de su entrada, la guardada |

Ejemplo de forma (los valores entre `<…>` los da la fuente):

```json
{"ok":true,"fuente":"boe.legislacion-consolidada","url":"https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21","fecha_consulta":"<RFC 3339>","hash":"sha256:<64 hex>","data":{"norma":"BOE-A-2015-10565","bloque":"a21","titulo":"<atributo titulo>","tipo":"<atributo tipo>","fecha_version":"<fecha_publicacion u original>","fecha_vigencia":"<atributo o vacío>","norma_modificadora":"<atributo o vacío>","texto":"<texto normalizado>","hash_texto":"sha256:<64 hex>","avisos":[],"url":"https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565#a21","url_eli":"<url_eli de los metadatos>"}}
```

## 4. `articulos <norma> <bloque> [<bloque>…]`

| Aspecto | Contrato |
|---|---|
| Validación | norma y **todos** los bloques antes de nada; al menos un bloque (Kong) |
| Peticiones | por cada id distinto, en el orden de primera aparición y sin caché de artículo: su bloque; los metadatos **una vez** por invocación, tras el primer bloque que los necesite y solo si no hay entrada vigente; se detiene en el primer fallo (FR-020, FR-021) |
| Caché | la de `articulo` por bloque y la de `metadatos` compartida |
| `data` | `[]Articulo` en el orden pedido, con repeticiones |
| `url` del sobre | `<base>/id/<norma>` |
| `fecha_consulta` | la más antigua de las de sus elementos |

## 5. `metadatos <norma>`

| Aspecto | Contrato |
|---|---|
| Validación | norma |
| Petición | `GET <base>/id/<norma>/metadatos`, `Accept: application/json` |
| Caché | `…\|1\|metadatos\|<dirección>`, 300 s (la misma entrada que leen y escriben `articulo` y `articulos`) |
| `data` | `Metadatos` con los avisos de `_check_vigencia` (FR-050); `data` vacío → 3 (FR-051) |
| `url` del sobre | la dirección de la petición |

## 6. `analisis <norma>`

| Aspecto | Contrato |
|---|---|
| Validación | norma |
| Petición | `GET <base>/id/<norma>/analisis`, `Accept: application/json` |
| Caché | `…\|1\|analisis\|<dirección>`, 604 800 s |
| `data` | `Analisis`; texto de las referencias anteriores completo (FR-060); `data` vacío → 3 (FR-061) |
| `url` del sobre | la dirección de la petición |

## 7. Peticiones emitidas por escenario (SC-003)

| Escenario | Peticiones a la fuente |
|---|---|
| Cualquier verbo con su entrada vigente | 0 |
| `articulo` sin su entrada y con la de metadatos vigente | 1 (bloque) |
| `articulo` sin su entrada ni la de metadatos | 2 (bloque, metadatos) |
| `articulo` con el bloque inexistente | 1 (bloque) |
| `articulos` de tres bloques sin ninguna entrada | 4 (b1, metadatos, b2, b3) |
| `articulos` de tres bloques con el segundo inexistente | 3 (b1, metadatos, b2) y b1 queda en caché |
| `articulos a21 a21 a21` sin entradas | 2 (a21, metadatos) |
| `--offline` o `--dry-run`, cualquier verbo | 0 |

## 8. Descripción de ensayo (FR-094)

`kitlegal boe articulo BOE-A-2015-10565 a21 --dry-run` sin entradas escribe en la salida de error, y en nada más:

```
--dry-run: no se ha ejecutado nada; se habría ejecutado el applet "boe", el verbo "articulo", con los argumentos ["articulo" "BOE-A-2015-10565" "a21" "--dry-run"]
--dry-run: se habría pedido GET https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21
--dry-run: se habría pedido GET https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/metadatos
```

y termina con 0, sin crear la caché.
