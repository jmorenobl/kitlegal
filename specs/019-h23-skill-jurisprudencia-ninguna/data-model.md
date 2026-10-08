# Data model: H23 · skill `jurisprudencia` + `cita preparar` y `cita cotejar`

Las entidades del spec («Key Entities») con los campos que fija el plan. Las formas JSON, con ejemplos, están en
contracts/applet-cita.md; las de las evals, en contracts/evals-jurisprudencia.md. Nada de esto se guarda: ni en la
caché, ni en el grafo, ni en ningún fichero.

## 1. Identificadores (`internal/core/ids`)

| Tipo | Forma | Reglas | Requisito |
|---|---|---|---|
| `ECLI` | `ECLI:ES:<órgano>:<año>:<número>` | Cinco partes separadas por dos puntos; `ECLI` y `ES` literales; órgano, 1 a 7 letras mayúsculas ASCII o cifras, la primera una letra; año, cuatro cifras; número, 1 a 25 letras mayúsculas ASCII, cifras o puntos. No se recorta ni se pasa a mayúsculas. El de otro país se rechaza diciendo que no es español | FR-005, FR-006 |
| `ROJ` | `<siglas> <número>/<año>` | Siglas, una o más palabras de letras mayúsculas ASCII separadas por un solo espacio; número, cifras; año, cuatro cifras | FR-005, FR-006 |

Los dos son valores inmutables que solo se construyen analizando, como `CodigoINE` y `DIR3`; su rechazo es un error
de clase `argumentos`. Lo que exportan: `AnalizarECLI`, `ECLI.String`, `ECLI.Organo` (el Tribunal Constitucional,
FR-014) y `ECLI.ROJ`; `AnalizarROJ`, `ROJ.String`, `ROJ.ECLI` y `ROJ.Numero` (FR-012, FR-025).

**Equivalencia** (FR-012): `ECLI.ROJ()` da `STS <número>/<año>` si el órgano es `TS` y el número es solo de cifras;
`ROJ.ECLI()` da `ECLI:ES:TS:<año>:<número>` si las siglas son `STS`. En cualquier otro caso, no se deduce. El número
y el año se trasladan carácter a carácter.

## 2. Referencia (`internal/core/cita`)

| Campo | Valores | Regla |
|---|---|---|
| `forma` | `ecli`, `roj`, `resolucion` | Una sola forma por invocación |
| `valor` | Tal como se dio | Con `resolucion`, `<número>/<año>` en cifras y con el año de cuatro |
| `fecha` | `AAAA-MM-DD` | Solo con `resolucion`, y obligatoria con ella; un día que existe |

Se construye de los cuatro argumentos —ECLI, ROJ, número y fecha—, cada uno dado o no dado: el que se escribe con
valor vacío está dado, y se comprueba con su forma como cualquier otro (research D3). Errores: más de una forma
dada; `fecha` sin `resolucion`; `resolucion` sin `fecha`; cualquier forma mal escrita, y la vacía lo es (FR-006). Sin
ninguna forma dada no hay referencia, que es un error en `preparar` sin texto y lo normal en `cotejar`.

## 3. Consulta preparada

| Campo | Regla |
|---|---|
| `referencia` o `texto` | Uno de los dos, nunca los dos (FR-015). El texto, ni vacío ni solo de blancos |
| `equivalente` | `{forma, valor}`; solo si §1 lo deduce de la referencia |
| `cobertura` | Solo con un ECLI de órgano `TC` (FR-014): `{cendoj: no-cubierto, motivo}`. En cualquier otra consulta no está: el binario no afirma que el CENDOJ tenga lo que se busca (research D4) |
| `direccion` | La del buscador con una referencia; la de la búsqueda con un texto; ausente con la referencia fuera de cobertura |
| `casillas` | Las de la forma dada, en el orden de FR-011; ninguna con un texto o fuera de cobertura |

**Casilla**: `{nombre, campo?, valor}`. Nombres: «ECLI», «Nº ROJ», «Nº Resolución», «Fecha resolución» con `campo`
«Desde» y «Hasta». La fecha de una casilla va `dd/mm/aaaa`.

**Dirección de búsqueda** (FR-013): `https://www.poderjudicial.es/search/sentencias/<texto>/1/AN`, con el texto en
UTF-8 y cada octeto que no es letra ASCII, cifra, `-`, `.`, `_` o `~` escrito `%XX` en mayúsculas.

## 4. Ficha y documento

El **documento** es el texto que llega: por `--documento` —dado, también vacío, es el texto, y la entrada no se
lee— o, en la orden y sin él, por la entrada estándar. Un texto vacío no es un documento (FR-026). Su identidad en
el sobre es el SHA-256 de sus bytes: la skill pasa la ficha sola, y la `url` es entonces la de la ficha.

| Dato de la ficha | Etiqueta en el documento | Clave | Forma exigida |
|---|---|---|---|
| ROJ | `Roj` (primera parte) | `roj` | La de un ROJ |
| ECLI | `Roj` (segunda parte, tras ` - `) | `ecli` | La de un ECLI español |
| Órgano | `Órgano` | `organo` | Texto no vacío |
| Fecha | `Fecha` | `fecha` | `dd/mm/aaaa`, un día que existe; en `data`, `AAAA-MM-DD` |
| Número de recurso | `Nº de Recurso` | `recurso` | Texto no vacío |
| Número de resolución | `Nº de Resolución` | `resolucion` | `<número>/<año>` |
| Ponente | `Ponente` | `ponente` | Texto no vacío |
| Tipo de resolución | `Tipo de Resolución` | `tipo` | Texto no vacío |

Se lee la primera ficha del texto, desde la primera línea `Roj:`; con uno de los ocho de menos, o con uno de los
cuatro que tienen forma exigida —ROJ, ECLI, fecha y número de resolución— sin la suya, no hay ficha (FR-021, FR-026).
El error nombra el primero que falta o que no tiene su forma, en el orden de la tabla.

## 5. Cotejo

| Campo | Regla |
|---|---|
| `ficha` | §4 |
| `correspondencia` | `se-corresponden` o `no-se-corresponden` si el ECLI y el ROJ de la ficha son los dos de la pareja de §1; `no-se-deduce` en otro caso (FR-023) |
| `pedida`, `es_la_pedida` | Solo con una referencia. Por ECLI o por ROJ, igualdad carácter a carácter con el de la ficha; por número y fecha, los dos iguales (FR-024) |
| `hallazgos` | Vacía, o un hallazgo si `es_la_pedida` es falso (FR-025) |

**Hallazgo de documento distinto**: `clase` `documento-distinto`; `difiere`, una `{dato, pedido, documento}` por dato
distinto; `cruce` `numero-de-resolucion` si el `<número>/<año>` del ROJ pedido es el número de resolución de la
ficha, o `numero-del-roj` si el número de resolución pedido es el `<número>/<año>` del ROJ de la ficha;
`explicacion`. No cambia el código de salida: 0, con `ok` verdadero.

## 6. La cita y la línea (la skill)

| Forma | Texto | De dónde salen sus datos |
|---|---|---|
| Cita de una sentencia | `<siglas> <resolución>, de <día> de <mes> [<ECLI>, ROJ: <ROJ>]` | De `ficha` de un `cita cotejar` de la conversación |
| Línea de sentencia no comprobada | `⚠ SENTENCIA NO COMPROBADA: <la referencia, como se dio>` | De la pregunta |

## 7. Eval y umbral

- **Eval**: gana `sentencias` y dos formas de comando (contracts/evals-jurisprudencia.md §1).
- **Hecho de la sesión**, por respuesta: los ECLI que leyó un `cita cotejar`; los que están en algún sobre de la
  sesión; los de la pregunta; las citas de la respuesta; y los ECLI sueltos de la respuesta fuera de una cita y de
  una línea que empieza por `⚠`. De ahí, si la respuesta cuenta (§4 del mismo contrato).
- **Umbral `cita_sin_documento:<modelo>:<modo>`**: un elemento de `umbrales` con el contrato del ADR 0029: `medida`,
  `total`, `comparacion` `<=`, `umbral` 0, `cumple`, `decide` verdadero. Uno por modo.

## 8. Estados

No hay ninguno. Cada invocación es independiente de las anteriores; no hay transiciones, ni concurrencia que
coordinar, ni escritura que pueda quedar a medias.
