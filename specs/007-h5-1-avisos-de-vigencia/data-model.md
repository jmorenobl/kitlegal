# Modelo de datos: H5.1 · Avisos de vigencia en las evals

Entidades que H5.1 añade o amplía. Lo que no se nombra aquí es lo de H5
(`specs/006-h5-skill-boe-legislacion/data-model.md`) con las enmiendas de ADR 0016, sin cambios. Las decisiones que lo
justifican están en [research.md](./research.md); los formatos exactos, en [contracts/](./contracts/).

## 1. Aviso de vigencia (`internal/source/boe`)

Lo emite el applet `boe` en `avisos` del `data` de `articulo`, de cada elemento de `articulos` y de `metadatos`
(H4, data-model §2.2). H5.1 no cambia ninguna de sus condiciones ni de sus frases (FR-012); añade la **etiqueta** como
dato exportado y compone la frase con ella (research D1).

| Código (`CodigosDeAviso()`, en este orden) | Etiqueta (`EtiquetasDeAviso()`) | Condición en los metadatos | Frase (`Aviso.Texto`) |
|---|---|---|---|
| `consolidacion-no-finalizada` | `TEXTO POSIBLEMENTE DESACTUALIZADO` | `estado_consolidacion.codigo` es `"4"` | `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO: la consolidación de esta norma no está finalizada. Puede haber modificaciones recientes aún no integradas.` |
| `derogada` | `NORMA DEROGADA` | `estatus_derogacion` es `"S"` | `⚠ NORMA DEROGADA: esta norma ha sido derogada.` |
| `vigencia-agotada` | `VIGENCIA AGOTADA` | `vigencia_agotada` es `"S"` | `⚠ VIGENCIA AGOTADA: esta norma ya no está en vigor.` |

Invariantes: `EtiquetasDeAviso()` tiene exactamente las claves de `CodigosDeAviso()`, cada una con una etiqueta no vacía
y distinta; cada frase empieza por `⚠ ` + etiqueta + `:`; cada llamada devuelve un mapa nuevo.

## 2. Forma fija de un aviso

Lo que la respuesta de la skill escribe y lo que el juicio busca: `⚠`, la etiqueta y dos puntos; lo que va detrás es
libre y no se lee. Gramática, tolerancias y ejemplos: research D2 y
[contracts/forma-fija-de-los-avisos.md](./contracts/forma-fija-de-los-avisos.md) §1.

| Parte | Regla |
|---|---|
| marca | U+26A0, seguido a lo sumo de un U+FE0F o un U+FE0E |
| separador | cero o más de: tabulador, carácter Zs de Unicode, `*`, `_` |
| etiqueta | las palabras de la etiqueta exportada, en su orden, cada una exacta salvo mayúsculas, separadas por separadores con al menos un blanco |
| final | `:` (U+003A) |
| ámbito | una sola línea: ningún salto de línea dentro de la forma |

**Avisos extraídos** de un texto (`ExtraerAvisos`): los códigos de `CodigosDeAviso()` cuya forma lleva el texto, en el
orden de `CodigosDeAviso()`, sin repetir; `nil` si ninguno.

## 3. Eval (amplía H5 data-model §6)

| Campo | Obligatorio | Regla |
|---|---|---|
| `avisos` | no | lista de ≥ 1 códigos de aviso, cada uno del enumerado `codigo-de-aviso` del esquema, que coincide con `CodigosDeAviso()`; **prohibida** si `activa: false`; repetidos admitidos (como `citas`) |

Tipo Go: `Eval.Avisos []string`, en el orden del fichero; nil si la eval no lleva `avisos`.

Validación (research V1-V3): código desconocido → `<fichero>: avisos/<i>, línea <n>: value must be one of
'consolidacion-no-finalizada', 'derogada', 'vigencia-agotada'`; lista vacía → `<fichero>: avisos, línea <n>: minItems:
got 0, want 1`; con `activa: false` → `<fichero>: línea 1: 'not' failed`. Las evals existentes, sin `avisos`, se leen
igual que antes (FR-023).

**Aviso esperado**: cada elemento de `Eval.Avisos`.

## 4. Resultado de eval (amplía H5 data-model §10.2)

| Campo | Regla |
|---|---|
| `avisos_encontrados` | los avisos esperados, en el orden de la eval y con sus repeticiones, que están entre los avisos extraídos de `respuesta` |
| `avisos_ausentes` | los demás avisos esperados, en ese orden |
| `motivos[]` | se añade, **detrás de cada cita ausente**, un motivo `aviso ausente: <código>` por cada aviso ausente. El del modelo distinto (ADR 0016) sigue yendo detrás de todos |
| `pasa` | la sesión terminó, la activación coincide y no hay comando, cita **ni aviso** ausente |

Sin `avisos` en la eval, o en una sesión ilegible, las dos listas quedan nulas y se escriben `[]` (research V8). El
**reparto de avisos** de una sesión es la pareja `avisos_encontrados` + `avisos_ausentes` con sus motivos.

## 5. Informe (amplía H5 data-model §10.3 y contrato del job §5)

- `informe.json`: cada elemento de `evals` lleva `avisos_encontrados` y `avisos_ausentes` detrás de `citas_ausentes`.
  Los motivos de la raíz llevan los de la sesión, precedidos de su nombre, como hasta ahora.
- `informe.md`: la tabla de `## Sesiones` pasa a tener las columnas Sesión · Eval · Modelo · Activa · Activada · Sesión
  terminada · Comandos ausentes · Citas ausentes · **Avisos encontrados** · **Avisos ausentes** · Resultado; cada celda
  de avisos es la lista separada por «, » o `ninguno`. La sección de cada sesión, con su respuesta, no cambia.
- Ningún recuento nuevo: la tasa de una eval es la de su serie (Clarifications, Q1).

## 6. Comprobaciones mecánicas

| Comprobación | Entrada | Defecto (un error por código, unidos con `errors.Join`) | Orden de los defectos |
|---|---|---|---|
| `ComprobarCodigosDeAviso` (FR-013) | esquema de eval compilado | `el esquema de eval no enumera el código de aviso <código> en avisos`; `el esquema de eval enumera en avisos el código <código>, que no es un código de aviso del binario` | primero los que faltan, en el orden de `CodigosDeAviso()`; después los que sobran, en el del esquema |
| `ComprobarFormasDeAviso` (FR-014) | contenido de `SKILL.md` | `falta la forma fija del aviso <código>: ⚠ <etiqueta>:` | el de `CodigosDeAviso()` |

Los códigos del esquema son los del `Enum` al que llegan `Properties["avisos"]` → `Items2020` → `Ref`…; sin él, el
esquema no enumera ninguno.

## 7. Norma nueva en `data/normas.yaml`

| Clave | Valor | Origen |
|---|---|---|
| identificador | `BOE-A-1992-26318` | búsqueda grabada `procedimiento administrativo común` (H4) |
| `titulo` | `Ley 30/1992, de 26 de noviembre, de Régimen Jurídico de las Administraciones Públicas y del Procedimiento Administrativo Común.` | la misma búsqueda |
| `rango` | `Ley` | la misma búsqueda |
| `abreviatura` | `LRJPAC` | research D13 |
| `materias` | `régimen jurídico de las administraciones públicas`, `procedimiento administrativo` | research D13 |

Sin ninguna marca de derogación (FR-056). `references/normas.md` la incluye tras `make skills-sync`.

## 8. Entrada nueva del manifiesto de grabación (`testdata/evals/grabaciones.json`)

| Campo | Valor |
|---|---|
| `busqueda` | `procedimiento administrativo común` |
| `titulo_empieza_por` | `Ley 30/1992,` |
| `bloques` | `["a42"]` |
| `para` | `eval 18-lrjpac-norma-derogada: artículo 42 de la Ley 30/1992, norma derogada con los avisos derogada y vigencia-agotada (FR-050 de H5.1); identificador de data/normas.yaml (FR-056 de H5.1); índice, metadatos y bloque de la cita (FR-074 de H5); la búsqueda, los metadatos y el bloque ya están en las grabaciones de H4, y solo el índice es nuevo` |

## 9. Eval de la norma derogada

`evals/boe-legislacion/18-lrjpac-norma-derogada.yaml`: pregunta «¿Qué dice el artículo 42 de la Ley 30/1992 sobre la
obligación de resolver?»; `activa: true`; `informativa: true`; comando esperado bloque `boe` `BOE-A-1992-26318` `a42`;
cita esperada `BOE-A-1992-26318` `a42`; `avisos` `derogada`, `vigencia-agotada`.

**Consultas necesarias** (H5 data-model §7.1) y dónde está grabada cada una:

| Consulta | Punto | Grabada en |
|---|---|---|
| `boe articulo BOE-A-1992-26318 a42` | comando esperado y cita esperada | H4 |
| `boe indice BOE-A-1992-26318` | norma de la eval | **H5.1** (grabación de la pausa) |
| `boe metadatos BOE-A-1992-26318` | norma de la eval | H4 |

Conjunto de `evals/boe-legislacion/` tras H5.1: 18 ficheros; 10 positivas que deciden, 6 informativas (13 a 17 y 18), 2
de no activación. Plan del job: 54 sesiones del modelo que decide y 36 del informativo (research V21).

## 10. Evidencia de la ejecución de aceptación (`gates/evals-aceptacion.md`)

| Dato | Contenido |
|---|---|
| Ruta de identificación | «primera ejecución de `evals` de la rama» (apertura) o «último evento `labeled` de `evals`» (repetición) |
| Ejecución | enlace, `databaseId`, `headSha`, conclusión y duración del paso «Ejecutar las evals» |
| Resumen del informe | la salida entera del programa de `jq` de quickstart §11.5: `commit`, `veredicto`, `red`, la tasa de la serie de la eval 18 con el modelo que decide y el reparto de avisos de sus 3 sesiones, y `true` |
| Commit y cabeza | la salida entera de la orden de quickstart §11.5 que compara el commit evaluado con la cabeza |
