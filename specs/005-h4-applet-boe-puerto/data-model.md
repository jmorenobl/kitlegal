# Data Model: H4 · Applet `boe`

Entidades del spec (*Key Entities*) llevadas a tipos y reglas. Las líneas son de `refs/boe.py`; las decisiones
que las justifican están en [research.md](./research.md). Los contratos de uso están en [contracts/](./contracts/).

## 1. Dónde vive cada entidad

| Entidad (spec) | Tipo Go | Paquete | Persistencia |
|---|---|---|---|
| Fuente `boe.legislacion-consolidada` | `Fuente` (implementa `core.Source`) | `internal/source/boe` | ninguna |
| Consulta (lo que se pide) | `ConsultaBuscar`, `ConsultaIndice`, `ConsultaArticulo`, `ConsultaArticulos`, `ConsultaMetadatos`, `ConsultaAnalisis` (implementan `core.Consulta`) | `internal/source/boe` | ninguna |
| Resultado de búsqueda | `ResultadoDeBusqueda` | `internal/source/boe` | entrada de caché de `buscar` |
| Norma (índice) | `Indice`, `EntradaDeIndice` | `internal/source/boe` | entrada de `indice` |
| Bloque y su versión vigente | `Articulo` | `internal/source/boe` | entrada de `articulo` |
| Aviso de vigencia | `Aviso` | `internal/source/boe` | dentro de `Articulo` y `Metadatos` |
| Norma (metadatos) | `Metadatos`, `EstadoDeConsolidacion` | `internal/source/boe` | entrada de `metadatos` |
| Análisis | `Analisis`, `Materia`, `Referencias`, `ReferenciaAnterior`, `ReferenciaPosterior` | `internal/source/boe` | entrada de `analisis` |
| Consulta guardada | `entrada` (sin exportar) | `internal/source/boe` | `internal/cache` (contenido opaco) |
| Procedencia con fecha | `schema.Procedencia{Fuente, URL, FechaConsulta}` | `internal/core/schema` | ninguna |
| Puerto de fuente | `core.Source`, `core.Consulta`, `core.Terminos` | `internal/core` | ninguna |

## 2. `data` de cada verbo

Todas las claves son **obligatorias** en el esquema y se emiten siempre: un campo ausente en la fuente va como
cadena vacía (FR-016) y una lista sin elementos como `[]`. Ninguna clave lleva el marcador `?`.

### 2.1 `articulo` → `Articulo`; `articulos` → `[]Articulo`

| Clave | Tipo | Valor | `boe.py` |
|---|---|---|---|
| `norma` | string | identificador pedido, ya validado | 399 |
| `bloque` | string | id de bloque pedido, ya validado | 399 |
| `titulo` | string | atributo `titulo` del bloque; vacío si falta | 106 |
| `tipo` | string | atributo `tipo` del bloque; vacío si falta (no se infiere del id) | 107 |
| `fecha_version` | string | `fecha_publicacion` de la última `version` si el atributo existe (aunque esté vacío); `original` si falta o si el bloque no tiene versiones | 112, 116 |
| `fecha_vigencia` | string | atributo de la última `version`; vacío si falta o sin versiones | 117, 110-112 |
| `norma_modificadora` | string | `id_norma` de la última `version`; vacío si falta o sin versiones | 118, 110-112 |
| `texto` | string | texto normalizado de la última `version`, o del bloque entero sin versiones (§3.2) | 111, 120, 132-136 |
| `hash_texto` | string, `^sha256:[0-9a-f]{64}$` | `sha256:` + hexadecimal en minúsculas del SHA-256 de los bytes UTF-8 de `texto` | — (FR-015) |
| `avisos` | `[]Aviso` | §2.2, derivados de los metadatos de la norma | 197-211 |
| `url` | string | `https://www.boe.es/buscar/act.php?id=<norma>#<bloque>` | 430 |
| `url_eli` | string | `url_eli` de los metadatos; vacío si falta | 482 (FR-015) |

`articulos`: lista en el orden pedido; un id repetido aparece en cada posición pedida con el mismo elemento.

### 2.2 `Aviso`

| Clave | Tipo | Valor |
|---|---|---|
| `codigo` | string, enumerado **exactamente** `consolidacion-no-finalizada`, `derogada`, `vigencia-agotada` | la condición |
| `texto` | string | la frase literal |

| Orden | Condición sobre el objeto de metadatos (comparación exacta de cadenas) | `codigo` | `texto` (carácter a carácter) | `boe.py` |
|---|---|---|---|---|
| 1 | `estado_consolidacion` es objeto y su `codigo` es la cadena `"4"` | `consolidacion-no-finalizada` | `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO: la consolidación de esta norma no está finalizada. Puede haber modificaciones recientes aún no integradas.` | 199-205 |
| 2 | `estatus_derogacion` es la cadena `"S"` | `derogada` | `⚠ NORMA DEROGADA: esta norma ha sido derogada.` | 207-208 |
| 3 | `vigencia_agotada` es la cadena `"S"` | `vigencia-agotada` | `⚠ VIGENCIA AGOTADA: esta norma ya no está en vigor.` | 210-211 |

Cero, uno, dos o tres avisos, en ese orden. La misma lista y los mismos textos en `articulo`, en cada elemento de
`articulos` y en `metadatos` (FR-050).

### 2.3 `buscar` → `[]ResultadoDeBusqueda`

| Clave | Tipo | Valor | `boe.py` |
|---|---|---|---|
| `identificador` | string | `identificador` del resultado | 293 |
| `titulo` | string | `titulo` | 294 |
| `rango` | string | `rango.texto` si `rango` es objeto; vacío en otro caso | 290-291, 295 |
| `vigencia_agotada` | string | `vigencia_agotada` | 296 |
| `estado_consolidacion` | string | `estado_consolidacion.texto` si es objeto; vacío en otro caso | 288-289, 297 |
| `url` | string | `https://www.boe.es/buscar/act.php?id=<identificador>` (con el identificador tal como llega, vacío incluido) | 298 |

Lista en el orden de la fuente; `[]` si no hay resultados (FR-032).

### 2.4 `indice` → `Indice`

| Clave | Tipo | Valor | `boe.py` |
|---|---|---|---|
| `norma` | string | identificador pedido | 349 |
| `url` | string | `https://www.boe.es/buscar/act.php?id=<norma>` | 384 |
| `bloques` | `[]EntradaDeIndice` | en el orden de la fuente (§3.1, lectura del índice) | 365-392 |

`EntradaDeIndice`: `id` (string, `id` del elemento; vacío si falta; 387), `titulo` (string, tal como llega;
388), `tipo` (string, enumerado: `articulo`, `titulo`, `capitulo`, `seccion`, `preambulo`,
`disposicion_adicional`, `disposicion_transitoria`, `disposicion_derogatoria`, `disposicion_final` o vacío;
`TipoDesdeID(id)`, §4; 389).

### 2.5 `metadatos` → `Metadatos`

| Clave | Tipo | Valor | `boe.py` |
|---|---|---|---|
| `norma` | string | identificador pedido | 446 |
| `titulo` | string | `titulo` | 473 |
| `rango` | string | `rango.texto` si objeto; la cadena si es cadena; vacío si falta | 468-469 |
| `numero_oficial` | string | `numero_oficial` | 475 |
| `fecha_disposicion` | string | `fecha_disposicion` | 476 |
| `fecha_publicacion` | string | `fecha_publicacion` | 477 |
| `fecha_vigencia` | string | `fecha_vigencia` | 478 |
| `estatus_derogacion` | string | `estatus_derogacion` | 479 |
| `vigencia_agotada` | string | `vigencia_agotada` | 480 |
| `estado_consolidacion` | `EstadoDeConsolidacion` | `codigo`: `estado.codigo` si objeto, vacío si no (467); `texto`: `estado.texto` si objeto, la cadena si es cadena, vacío si falta (466) | 465-467 |
| `url_eli` | string | `url_eli` | 482 |
| `avisos` | `[]Aviso` | §2.2 (textos de `_check_vigencia`, no los de 485-498) | 199-211 |

### 2.6 `analisis` → `Analisis`

| Clave | Tipo | Valor | `boe.py` |
|---|---|---|---|
| `norma` | string | identificador pedido | 505 |
| `materias` | `[]Materia` | §3.1; cada una `{codigo, texto}`; una materia que no es objeto aporta solo `texto` | 527-535 |
| `notas` | `[]string` | contenido de `notas` o de su envoltorio `nota`: una cadena → lista de una; una lista de cadenas → tal cual; vacío o ausente → `[]` | 538-541 |
| `referencias.anteriores` | `[]ReferenciaAnterior` | `relacion` (`relacion.texto` si objeto; la cadena si es cadena), `norma` (`id_norma`), `texto` (**completo**, sin recorte) | 544-564 (FR-060) |
| `referencias.posteriores` | `[]ReferenciaPosterior` | `relacion`, `norma` | 566-581 |

Si `referencias` no es un objeto, las dos listas van vacías (línea 545).

## 3. Reglas de lectura de la fuente

### 3.1 JSON (buscar, índice, metadatos, análisis)

| # | Regla | `boe.py` |
|---|---|---|
| J1 | El cuerpo es JSON y su raíz un objeto; si no → ilegible (4) | 274-280, 357-360, 454-457, 513-516 |
| J2 | `data` ausente o `null` equivale a lista vacía; «vacío» = lista vacía, objeto vacío, cadena vacía o `null` | 278, 358, 455, 514 |
| J3 | `data` vacío: `buscar` → `[]`; `indice`, `metadatos`, `analisis` → «no encontrado» (3) y no se guarda | 282, 362, 459, 518 |
| J4 | Objeto suelto → lista de uno (FR-070) en: resultados de `buscar`; `data` de `indice`; `data[0].bloque`; `materias`; `referencias.anteriores`; `referencias.posteriores`; contenido de los envoltorios `materia`, `anterior`, `posterior` | 218-225 (adaptado) |
| J5 | Metadatos y análisis: el primer elemento de `data` si es lista, o `data` si es objeto; si ese elemento no es objeto → ilegible (4) | 195, 463, 522 |
| J6 | Índice: si el primer elemento de `data` es objeto con clave `bloque`, los bloques son su `bloque`; si no, `data` | 365-369 |
| J7 | Envoltorios: materia `{materia: X}` o directa; notas `{nota: X}` o directas; referencia `{anterior: X}`/`{posterior: X}` o directa | 531, 540, 550-554, 570-574 |
| J8 | En listas de resultados, bloques y referencias, un elemento que no es objeto se salta | 287, 386, 556, 576 |
| J9 | Texto: cadena → su valor; ausente o `null` → vacío; `rango`/`estado_consolidacion`/`relacion` según §2; cualquier otro tipo donde se espera texto → ilegible (4) nombrando el campo | 289-297, 466-482, 533, 559-561, 579-581 (adaptado, D7) |

### 3.2 XML (bloque)

| # | Regla | Referencia |
|---|---|---|
| X1 | El cuerpo es UTF-8 válido; si no → ilegible (4) | `boe.py` 82 |
| X2 | Se ignora la declaración `encoding` | S5 (c) |
| X3 | Un único elemento raíz; fuera de él solo espacio en blanco, comentarios o instrucciones de proceso; si no → ilegible (4) | `ET.fromstring` con expat, S5 (f); `encoding/xml` no lo exige (`xml.go` 275-290, `go doc encoding/xml.Decoder.Token`), así que lo comprueba la lectura |
| X4 | El bloque es el primer descendiente `bloque` sin espacio de nombres que **no** sea la raíz; si no hay → ilegible (4) | `boe.py` 102-104; `ElementPath.py` 182-209 |
| X5 | Atributos `titulo`, `tipo` del bloque; `fecha_publicacion`, `fecha_vigencia`, `id_norma` de la última versión; valores normalizados (tabulador, salto, retorno → espacio) | `boe.py` 106-118; XML 1.0 §3.3.3 |
| X6 | Versiones = hijos **directos** `version`; vigente = la última | `boe.py` 109, 115 |
| X7 | Texto de un elemento = todo el `CharData` y `CDATA` del elemento y sus descendientes en orden **más** el `CharData` que sigue a su cierre hasta la siguiente etiqueta; comentarios e instrucciones de proceso no aportan texto ni cortan el que los rodea | `ElementTree.py` 406-423, 979-983; 1413-1415, 1486-1511 (`TreeBuilder` de Python puro: ni los inserta ni vacía el texto acumulado); el acelerador C se comporta igual, S5 (e); `CDATA` llega como `CharData`, `xml.go` 699-715 |
| X8 | Normalización: partir por `\n`; recortar cada línea con `unicode.IsSpace` ∪ U+001C–U+001F; descartar vacías; unir con `\n` | `boe.py` 132-136; S5 (a) |

## 4. Tipo inferido del id (`TipoDesdeID`, para `indice`)

Sobre `id` pasado a minúsculas (`strings.ToLower`), mirando runas y en este orden (`boe.py` 155-176); la primera
regla que casa decide:

| # | Condición | Tipo |
|---|---|---|
| 1 | primera runa `a` y segunda un dígito (`unicode.IsDigit`) | `articulo` |
| 2 | empieza por `t` | `titulo` |
| 3 | empieza por `ci` o `cv`, o primera runa `c` y segunda una de `i v x l c d m` | `capitulo` |
| 4 | empieza por `s` y la segunda runa es dígito o `e` | `seccion` |
| 5 | es exactamente `preambulo` | `preambulo` |
| 6 | empieza por `da` | `disposicion_adicional` |
| 7 | empieza por `dt` | `disposicion_transitoria` |
| 8 | empieza por `dd` | `disposicion_derogatoria` |
| 9 | empieza por `df` | `disposicion_final` |
| 10 | ninguna | vacío |

Rarezas portadas: `ti` es `titulo` y no `capitulo` (la regla 2 va antes); `cv3` es `capitulo`; `subseccion` no la
produce ninguna regla. `a21` → `articulo`, `da3` → `disposicion_adicional`, `dt1` → `disposicion_transitoria`
(SC-007).

## 5. Identificadores de entrada

| Entrada | Gramática | Fuera de la gramática |
|---|---|---|
| Norma | `^BOE-A-[0-9]{4}-[0-9]{1,9}$` | «argumentos» (2) antes de abrir la caché y de pedir |
| Bloque | `^[A-Za-z0-9][A-Za-z0-9.-]{0,63}$`, atada a los índices grabados (`TestGramaticaCubreLosIndicesGrabados`): admite el guion y el punto que la fuente usa (`a1-30`, `da-3`, `a85bis.`, `ci-2`), caracteres no reservados del RFC 3986 §2.3 que `url.PathEscape` no escapa, y exige letra o dígito inicial, que deja fuera los segmentos `.` y `..` | «argumentos» (2) antes de abrir la caché y de pedir |
| Texto de búsqueda | argumentos unidos por espacio; sin operadores, al menos una palabra | cero palabras → «argumentos» (2) antes de abrir la caché y de pedir |
| Bloques de `articulos` | al menos uno (Kong); cada uno con la gramática de bloque | ninguno → 2 (kernel); uno inválido → 2 sin pedir ninguno |

La validación de **todos** los argumentos de una consulta ocurre antes de abrir la caché y de construir el
cliente: un código 2 no deja rastro en disco ni en red.

## 6. Consulta guardada (entrada de caché)

| Verbo que escribe | Clave | Contenido `datos` | `url` guardada | `fecha_consulta` guardada | Vigencia |
|---|---|---|---|---|---|
| `buscar` | `boe.legislacion-consolidada\|1\|buscar\|<dirección de búsqueda>` | `[]ResultadoDeBusqueda` (también `[]`) | la dirección de búsqueda | instante de su petición | 300 s |
| `indice` | `…\|1\|indice\|<dirección del índice>` | `Indice` | la del índice | instante de su petición | 604 800 s |
| `metadatos`, `articulo`, `articulos` | `…\|1\|metadatos\|<dirección de metadatos>` | `Metadatos` | la de metadatos | instante de su petición | 300 s |
| `articulo`, `articulos` | `…\|1\|articulo\|<dirección del bloque>` | `Articulo` | la del bloque | mín(instante del bloque, `fecha_consulta` de los metadatos usados) | 604 800 s |
| `analisis` | `…\|1\|analisis\|<dirección del análisis>` | `Analisis` | la del análisis | instante de su petición | 604 800 s |

Contenido serializado: `{"fecha_consulta": "<RFC 3339 con nanosegundos y desplazamiento>", "url": "<…>", "datos":
<…>}`. Se lee sin admitir claves desconocidas; ilegible → «inesperado» (1) nombrando la clave. Nunca se escribe
un fallo ni se escribe nada con `--offline` o `--dry-run` (la caché se abre en solo lectura).

## 7. Flujos

### 7.1 `articulo <norma> <bloque>`

1. Validar norma y bloque (§5). Abrir la caché (solo lectura si `--offline` o `--dry-run`).
2. Leer `articulo|<bloque>`. Vigente → `Resultado{URL: dirección del bloque, FechaConsulta: guardada, Datos}`. Fin.
3. Ausente con `--offline` → fallo 4, URL del bloque, sin fecha. Fin.
4. Pedir el bloque (`Accept: application/xml`). Ensayo → anotar y seguir al 6 sin leer. Fallo de `httpx` → su clase,
   URL del bloque, `Error.Instante`. 404 → 3 (no se piden metadatos). Otro no 2xx → 4. 2xx → leer (§3.2); ilegible → 4.
5. Instante del bloque = `Respuesta.Instante`.
6. Leer `metadatos|<norma>`. Vigente → usar su `datos` y su `fecha_consulta`. Ausente → pedir metadatos (`Accept:
   application/json`); ensayo → anotar; fallo → su clase y URL de metadatos; 404 o `data` vacío → 3; ilegible → 4;
   2xx → leer (§3.1), **escribir** `metadatos|<norma>` con su instante.
7. En ensayo: `Resultado{URL: dirección del bloque, Ensayo: líneas}`. Fin.
8. Componer `Articulo` (avisos §2.2, `url_eli`, `hash_texto`); fecha = mín(bloque, metadatos); **escribir**
   `articulo|<bloque>`; devolver.

### 7.2 `articulos <norma> <b1> [<b2>…]`

1. Validar norma y todos los bloques. Abrir la caché.
2. Recorrer los ids **distintos** en el orden de primera aparición. Para cada uno:
   - `articulo|<bloque>` vigente → usarlo.
   - ausente con `--offline` → fallo 4 con URL de ese bloque. Fin.
   - pedir el bloque como en 7.1 paso 4; fallo → su clase, URL de ese bloque y su instante; **no** se piden los
     siguientes; los ya resueltos quedaron escritos. Fin.
   - metadatos: si aún no se tienen en esta invocación, 7.1 paso 6 (**una vez** por invocación).
   - componer y **escribir** `articulo|<bloque>`.
3. Ensayo: una línea por bloque sin entrada y **una** por los metadatos si hacían falta y no estaban; `Resultado{URL:
   dirección de la norma, Ensayo}`.
4. `Datos` = elementos en el orden pedido (con repeticiones); `URL` = `https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/<norma>`;
   fecha = mín de las fechas de los elementos.

### 7.3 `buscar`, `indice`, `metadatos`, `analisis`

Validar → abrir caché → entrada vigente (servir) / ausente con `--offline` (4) / pedir (ensayo: una línea) →
clasificar (§3, [contracts/errores-y-codigos.md](./contracts/errores-y-codigos.md)) → leer → escribir (salvo
fallo) → `Resultado{URL: la del recurso, FechaConsulta: instante, Datos}`.

## 8. La fuente

| Elemento | Valor |
|---|---|
| `Name()` | `boe.legislacion-consolidada` |
| `TTL(ConsultaBuscar)`, `TTL(ConsultaMetadatos)` | 300 s |
| `TTL(ConsultaIndice)`, `TTL(ConsultaArticulo)`, `TTL(ConsultaArticulos)`, `TTL(ConsultaAnalisis)` | 604 800 s |
| `Terms()` | `terminosDeUso` = `{URL: <términos de uso>, Revisados: <fecha>}` (`terminos.go`), iguales a la fila de `docs/SOURCES.md`; `Revisados` cero con la fila en `pendiente` solo hasta la pausa del manifiesto, en la que una persona fija los dos |
| `IntervaloEntrePeticiones` | igual al ritmo de la fila de `docs/SOURCES.md` (`terminos.go`; propuesto `1s`, S6, hasta que una persona lo fija en la pausa del manifiesto, antes de grabar) |
| Base de la API | `https://www.boe.es/datosabiertos/api/legislacion-consolidada` (`boe.py` 71-72) |
| Base pública | `https://www.boe.es/buscar/act.php` (`boe.py` 298, 384, 430) |

## 9. Invariantes

1. Todo sobre de éxito de `boe` lleva `fuente: "boe.legislacion-consolidada"` y una `url` que empieza por
   `https://www.boe.es/` (FR-002, SC-011).
2. Ninguna entrada de caché contiene un fallo (FR-093).
3. Servir una entrada da la misma `url`, `fecha_consulta` y `data` byte a byte que el sobre que la escribió o
   habría escrito (FR-096).
4. La `fecha_consulta` de un sobre que se apoya en varias consultas es la más antigua (FR-096).
5. `hash_texto` cambia si y solo si cambia `texto` (FR-015, SC-013).
6. Un código 2 no abre la caché ni construye cliente (§5).
7. `--offline` y `--dry-run` no crean ni modifican la caché (FR-092, FR-094).
8. `--no-graph` y `--asunto` no cambian nada de lo anterior (FR-128).
9. Ninguna ruta produce el código 6 (FR-100).
