# Contrato: el informe del job de evals (umbrales y recuentos)

`internal/evals/informe.go` y `umbrales.go`; `informe.json` e `informe.md` de cada trabajo del flujo `evals`. Contrato de
`umbrales` del ADR 0029, sin cambios en sus claves ni en sus invariantes (H7.3 FR 001). Research D13, D14.

## 1. Qué se cuenta (FR-045, FR-061)

Una **respuesta medida** de un modelo es una sesión de una serie que pide el plan (no la de la prueba de red), de una
eval que activa la skill, legible, no sin medir y **terminada** (`sesion_terminada`). Las demás se siguen publicando en
`evals`, con sus motivos, y la sin terminar sigue sin pasar en su serie; no cuentan en ningún recuento ni umbral, ni en
la medida ni en el total.

Por modelo del job (el que decide y los informativos, en su orden): `respuestas` (las medidas), `con_alguna` (con alguna
expresión de la lista), `con_redaccion_no_leida` (con alguna de `redaccion_no_leida`) y `sin_activar` (sin la skill
activada). `expresiones_prohibidas_por_modelo` publica, con sus claves de hoy, `modelo`, `con_alguna` y `respuestas`; los
otros dos solo llegan a los umbrales. Sin lista, `expresiones_prohibidas_por_modelo` es `[]` como hoy.

## 2. Los umbrales (FR-040 a FR-048)

En una skill con lista, en este orden: los tres del modelo que decide, los de las expresiones de cada informativo y el
de la duración si hay objetivo. En una sin lista, solo el de la duración si hay objetivo (`legal-core`: `[]`, porque su
objetivo es 0).

| `nombre` | `descripcion` | `medida` | `total` | `umbral` | `decide` |
|---|---|---|---|---|---|
| `expresiones_prohibidas:<que decide>` | «Respuestas de <modelo> con alguna expresión prohibida, sobre sus respuestas medidas en las evals que activan la skill» (la de hoy) | `con_alguna` | `respuestas` | 0.05 | `true` |
| `sin_activar:<que decide>` | «Respuestas de <modelo> sin la skill activada, sobre sus respuestas medidas en las evals que la activan» | `sin_activar` | `respuestas` | 0 | `true` |
| `redaccion_no_leida:<que decide>` | «Respuestas de <modelo> con alguna expresión de redaccion_no_leida (una redacción que ninguna orden devolvió), sobre sus respuestas medidas en las evals que activan la skill» | `con_redaccion_no_leida` | `respuestas` | 0 | `true` |
| `expresiones_prohibidas:<informativo>` | la de hoy | `con_alguna` | `respuestas` | 0.05 | `false` |
| `duracion_de_las_sesiones` | la de hoy | segundos | — | objetivo (900) | `true` |

Todos con `comparacion` `"<="`; `cumple` es `medida/total <= umbral` (0 si `total` es 0) o `medida <= umbral` sin total,
en `float64` y sin redondeos (ADR 0029). Con 54 respuestas medidas, el de las expresiones admite 2 (2/54 ≈ 0,037) y no 3
(3/54 ≈ 0,056); los de umbral 0 admiten 0.

Ejemplo de un elemento, 250 bytes (los cinco de `boe-legislacion`, ≈ 1,2 KB, fijos: no crecen con el uso ni con lo
consultado, porque se miden de nuevo en cada job sobre su commit):

```json
{"nombre":"sin_activar:claude-sonnet-5-5","descripcion":"Respuestas de claude-sonnet-5-5 sin la skill activada, sobre sus respuestas medidas en las evals que la activan","medida":0,"total":54,"comparacion":"<=","umbral":0,"cumple":true,"decide":true}
```

## 3. Los motivos y el veredicto (FR-046)

Cada umbral que decide y no se cumple, salvo el de la duración, da el motivo de la raíz de hoy:
`umbral <nombre>: <medida> de <total> (<p> %), y tiene que ser ≤ <u> %`. Por ejemplo, «umbral
sin_activar:claude-sonnet-5-5: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %» y «umbral
redaccion_no_leida:claude-sonnet-5-5: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %». El de la duración, con el prefijo de la
ejecución, como hoy. Con algún motivo, el veredicto es `fallo` y `scripts/evals.sh` sale con 1 (el trabajo en rojo, que el
cierre cuenta). Un umbral que se cumple no da motivo ni cambia el veredicto. Ninguna tarea cumple un umbral rebajándolo,
dejándolo en `decide: false`, sacando evals del total o recortando la lista (FR-047).

## 4. `informe.md`

- «Umbrales»: las filas de §2, con las columnas de hoy.
- «Expresiones prohibidas por modelo»: la tabla de hoy, con las respuestas medidas de §1.
- «Tasas por eval»: la columna «Formas exigidas» lleva, detrás de las de los hallazgos, `⚠ REDACCIÓN MODIFICADA: <texto>`
  por cada redacción esperada de la eval (la 20: dos).
- «Sesiones»: dos columnas detrás de «Hallazgos ausentes»: «Redacciones modificadas encontradas» y «Redacciones
  modificadas ausentes» (`ninguna` si no hay).

## 5. El sondeo (FR-082)

El recuento de la línea «Respuestas con alguna expresión prohibida en las evals que activan la skill: …» es el de §1
(solo las terminadas). Nada más de su salida cambia (fuera de alcance: `sin_activar`, `redaccion_no_leida` o recuentos
por clase en el sondeo).

## 6. Tests (FR-097; SC-006)

`TestUmbralesDelInforme` (`umbrales_test.go`) gana casos con sesiones sintéticas del modelo que decide en 18 evals que
activan la skill (54 respuestas), cada uno con su veredicto y sus motivos esperados:

| Caso | Medida | Esperado |
|---|---|---|
| `sin-activar-una` | 1 de 54 sin activar | `sin_activar:<m>` con `cumple: false`; veredicto `fallo`, con su motivo |
| `sin-activar-ninguna` | 0 de 54 | `cumple: true`; el veredicto es el de las series |
| `redaccion-no-leida-una` | 1 de 54 con una expresión de `redaccion_no_leida` | `redaccion_no_leida:<m>` y `expresiones_prohibidas:<m>` las cuentan (1); `redaccion_no_leida` no se cumple; `fallo` con su motivo |
| `redaccion-no-leida-ninguna` | 0 | `cumple: true` |
| `tres-de-54` | 3 de 54 con alguna expresión | `expresiones_prohibidas:<m>` 3 de 54, no se cumple, `fallo` |
| `dos-de-54` | 2 de 54 | se cumple |
| `tres-de-54-y-seis-sin-terminar` | 3 de 54 y 6 sesiones más sin terminar (tope, código 124) | `total` 54, no 60; no se cumple; las seis, publicadas con su motivo, no cuentan en `expresiones_prohibidas_por_modelo` |
| `una-sin-activar-con-expresion` | la misma sesión sin activar y con una expresión | cuenta una vez en cada umbral |

Y el de `legal-core` sigue dando `[]` (`TestUmbralesDelInforme` ya lo fija). `TestInformeMarkdownDeLosUmbrales` gana las
columnas y las formas exigidas de §4. Los informes esperados de `informe_test.go` y `juzgar_test.go` ganan
`redacciones_modificadas_encontradas` y `redacciones_modificadas_ausentes` (vacías) y, donde los arman con la lista del
repositorio, los umbrales nuevos. Ninguno se desactiva ni se salta.
