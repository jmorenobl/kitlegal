# Data model: H26

El hito no añade ningún tipo, ningún campo ni ninguna clave. Cambia qué elementos compone el informe con los tipos que
ya hay, y dos números de datos. Los bytes de cada elemento están en
[contracts/umbrales-del-informe.md](./contracts/umbrales-del-informe.md).

## 1. La clase del juez (`ClaseDelJuez`, sin cambios)

Lo que declara cada entrada de `evals/<skill>/juez/clases.yaml`, contra `schemas/juez-clases.yaml.json`, que no cambia.

| Campo | Qué es | Qué gobierna desde H26 |
|---|---|---|
| `nombre` | El de la clase | El nombre de sus umbrales |
| `decide` | Si la clase hace fallar el job | Que la clase tenga **umbral conjunto** y los dos de la medida, y que una respuesta quede marcada con tres votos y no con el primero. Ya no es el `decide` de su recuento por modo, que es siempre `false` |
| `umbral` | Una proporción de 0 a 1 | El `umbral` del conjunto y el de cada recuento por modo. En una clase que decide es una tasa sobre las respuestas de todos los modos juntas |

Los valores (FR-020), que son los del ADR 0038:

| Skill | Clase | `decide` | `umbral` |
|---|---|---|---|
| `boe-legislacion` | `afirma_lo_no_leido` | `true` | **0.028** (era 0) |
| `boe-legislacion` | `cuenta_su_proceso` | `false` | 0 |
| `jurisprudencia` | `afirma_lo_no_leido` | `true` | **0.017** (era 0) |
| `jurisprudencia` | `afirma_que_existe` | `false` | 0 |

El diff de los dos ficheros: [contracts/clases-yaml.diff](./contracts/clases-yaml.diff).

## 2. Las respuestas que se miden (sin cambios)

Los grupos de `respuestasQueSeJuzgan`: uno por modo del plan, el del modo orden delante, y uno sin modo, el de la eval
sin binario ni servidor. En cada uno, las respuestas del modelo que decide en las evals que activan la skill, de
sesiones planificadas, legibles, medidas y terminadas, en orden de sesión.

- **Juzgada**: está en su grupo. Es el `total`.
- **Marcada** en una clase: `marcadasEn` la devuelve; con tres síes con su frase comprobada si la clase decide, y con
  el primero si solo se publica. Es la `medida`.
- **Sin juzgar**: un voto no llegó. Sigue en el grupo y `marcadasEn` no la devuelve.

El grupo sin modo no entra en ningún umbral.

## 3. Los dos elementos de una clase que decide

Los dos son un `Umbral`, el tipo de hoy, con el contrato del ADR 0029.

| | Recuento por modo | Umbral conjunto |
|---|---|---|
| Cuántos | Uno por modo del plan | Uno |
| `nombre` | `<clase>:<modelo>:<modo>` | `<clase>:<modelo>` |
| `medida` | Las marcadas del grupo de ese modo | La suma de las de todos los modos |
| `total` | Las respuestas del grupo de ese modo | La suma de las de todos los modos |
| `umbral` | El de la clase | El de la clase |
| `cumple` | `medida/total <= umbral`; `true` con `total` 0 | Lo mismo |
| `decide` | `false` | `true` |
| Motivo | Nunca | Si no se cumple |

Una clase que solo se publica tiene el recuento por modo, como hoy, y no tiene umbral conjunto.

**Invariantes**, que valen para cualquier skill con juez:

1. `conjunto.medida` es la suma de `recuento.medida` de sus modos, y `conjunto.total`, la de `recuento.total`.
2. Los tres llevan el mismo `umbral`.
3. El número de elementos de `umbrales` es, con *m* modos, *c* clases, *d* de ellas que deciden, *s* = 1 si alguna
   eval declara sentencias y *o* = 1 si el job da objetivo de duración: `m·(1 + c + s) + d·3 + m·o + m`. En
   `boe-legislacion`, `2·3 + 3 + 2 + 2 = 13`; en `jurisprudencia`, `2·4 + 3 + 0 + 2 = 13`. `legal-core`, sin juez, sin
   sentencias y sin objetivo, no tiene ninguno.
4. Deciden `m·(1 + s) + d·3 + m·o + m`: `2 + 3 + 2 + 2 = 9` y `4 + 3 + 0 + 2 = 9`.
5. El veredicto es `fallo` si y solo si hay algún motivo. El conjunto da uno si y solo si no se cumple.

## 4. Lo que cambia en `internal/evals/umbrales.go`

Nada exportado cambia: `Umbral` conserva sus campos y sus claves.

| Pieza | Cambio | Requisito |
|---|---|---|
| `umbralesDelInforme` | Añade los conjuntos detrás de los umbrales de las respuestas y delante de los de la medida | FR-005 |
| `umbralesConjuntos` (nueva, de `juicioDelJuez`) | Por cada clase que decide, el elemento de §3 | FR-001 a FR-003 |
| `nombreDelUmbralConjunto` (nueva) | `<clase>:<modelo>`; `nombreDelUmbralDeClase` le añade el modo | FR-001 |
| La descripción del conjunto (constante nueva) | El texto de contracts §2 | FR-001 |
| `umbralesDeLasClases` | `Decide: false` | FR-010 |
| `marcadasDelUmbral` | Casa el nombre del conjunto y nombra las marcadas de todos los modos, cada una con ` (modo <modo>)` | FR-004 |

## 5. Lo que cambia en los tests

| Pieza | Cambio |
|---|---|
| `ejecucionConUmbrales.clasesDelJuez` (campo nuevo del arnés) | El contenido de `clases.yaml` de la carpeta del juez de la ejecución sintética; vacío, la constante de hoy, con `umbral: 0`. Como `medidaDelJuez` |
| Los ayudantes que escriben a mano los umbrales esperados (`umbralesConJuez`, `umbralesDeUnModo`, `umbralesDeJurisprudencia`, `umbralesDeLasClasesDeJurisprudencia`) | El recuento por modo con `Decide: false` y el umbral de la clase; el conjunto, en su sitio |

## 6. El registro de los mutantes

`specs/021-h26-el-umbral-de/mutantes.md`, que escribe la tarea de FR-078. Por cada uno de los dos mutantes: la línea
que cambia, la orden ejecutada y los subtests de `TestUmbralConjunto` y de `TestUmbralesDeJurisprudencia` que fallan.
Es un registro del run: nada lo lee en `make ci`.
