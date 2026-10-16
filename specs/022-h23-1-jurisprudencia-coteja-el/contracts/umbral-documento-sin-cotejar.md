# Contrato: el umbral `documento_sin_cotejar:<modelo>:<modo>` del informe del job de evals

El contrato de `umbrales` es el del ADR 0029 y no cambia. Este documento fija el elemento nuevo, su sitio entre los
de `jurisprudencia`, su motivo, su fila de `informe.md`, cómo se prueba y su uso. Las cifras medidas son las de
research M9 a M13.

## 1. Los quince elementos de `jurisprudencia`, en su orden

Con el plan del job —el modelo que decide, `claude-sonnet-5-5`, y los dos modos— (FR-006):

| # | `nombre` | `total` | `umbral` | `decide` |
|---|---|---|---|---|
| 1 | `sin_activar:claude-sonnet-5-5:orden` | 30 | `0` | sí |
| 2 | `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 30 | `0.017` | no |
| 3 | `afirma_que_existe:claude-sonnet-5-5:orden` | 30 | `0` | no |
| 4 | `cita_sin_documento:claude-sonnet-5-5:orden` | 30 | `0` | sí |
| 5 | **`documento_sin_cotejar:claude-sonnet-5-5:orden`** | **12** | **`0`** | **sí** |
| 6 | `sin_activar:claude-sonnet-5-5:herramienta` | 30 | `0` | sí |
| 7 | `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 30 | `0.017` | no |
| 8 | `afirma_que_existe:claude-sonnet-5-5:herramienta` | 30 | `0` | no |
| 9 | `cita_sin_documento:claude-sonnet-5-5:herramienta` | 30 | `0` | sí |
| 10 | **`documento_sin_cotejar:claude-sonnet-5-5:herramienta`** | **12** | **`0`** | **sí** |
| 11 | `afirma_lo_no_leido:claude-sonnet-5-5` | 60 | `0.017` | sí |
| 12 | `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 125 | `0` | sí |
| 13 | `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 124 | `0` | sí |
| 14 | `duracion_del_juez:orden` | — | `900` | sí |
| 15 | `duracion_del_juez:herramienta` | — | `900` | sí |

Quince, once que deciden; no deciden los cuatro recuentos de las clases del juez (2, 3, 7 y 8). Los trece de antes no
cambian ni de nombre, ni de valores, ni de orden entre sí (FR-030). `boe-legislacion` sigue con sus trece, nueve que
deciden, y `legal-core`, con `[]` (FR-007): ninguna de sus evals pide el cotejo.

La regla, sin ningún caso por skill: por cada modo, detrás de `cita_sin_documento` si la skill lo tiene, va
`documento_sin_cotejar` si alguna eval de la skill pide el cotejo (data-model §1); y solo en una skill que ya tiene
umbrales de sus respuestas, con juez o con alguna eval que declara `sentencias` (research D4).

## 2. El elemento

Ejemplo, el del modo orden con todas las sesiones cotejando, compacto como lo escribe `jq -c` sobre `informe.json`
(366 bytes; research M9):

```json
{"nombre":"documento_sin_cotejar:claude-sonnet-5-5:orden","descripcion":"Respuestas de claude-sonnet-5-5 en el modo orden cuya sesión no ejecutó ningún cita cotejar que consultara y terminara bien, sobre sus respuestas medidas en las evals que activan la skill y piden cita cotejar","medida":0,"total":12,"comparacion":"<=","umbral":0,"cumple":true,"decide":true}
```

- `total`: las respuestas del grupo del modo —las del `total` de `sin_activar` de ese modo— cuya eval pide el cotejo.
  Hoy, 4 evals × 3 repeticiones = 12.
- `medida`: de ellas, las de una sesión sin ningún cotejo (data-model §2). Cada sesión, una vez como mucho.
- `cumple`: `medida/total <= 0`, en coma flotante, 0 si `total` es 0. Con 1 de 12, `false`; con 0 de 12, `true`.
- `decide`: siempre `true`.

Qué sesiones están en el `total` y cuáles no:

| Sesión | ¿En el `total`? | ¿Puede contar en la `medida`? |
|---|---|---|
| Del modelo que decide, medida, de una eval que pide el cotejo | Sí | Sí, si no cotejó |
| Que además no activó la skill | Sí | Sí, si no cotejó; y cuenta también en `sin_activar` |
| De una eval que no pide el cotejo (01, 02, 03, 05, 07 y 10) | No | No |
| Sin medir (límite de uso, reintentos agotados) | No | No; da su motivo de la ejecución, como hoy |
| Que el tope cortó (sin terminar) | No | No; como en `sin_activar` |
| Del modelo informativo (`claude-haiku-4-5-20251001`) | No | No |
| De una eval sin binario ni servidor | No | No |

## 3. El motivo

Si no se cumple, la raíz del informe lleva un motivo, y el veredicto es `fallo` (FR-003):

```text
umbral documento_sin_cotejar:claude-sonnet-5-5:orden: 1 de 12 (8,3 %), y tiene que ser ≤ 0,0 %: 06-documento-que-no-es-el-pedido-claude-sonnet-5-5-01
```

- Es `motivoDeUnUmbral`, «umbral `<nombre>`: `<medida>` de `<total>` (`<p>` %), y tiene que ser ≤ 0,0 %», como el de
  todo umbral que decide; detrás, tras «: », cada sesión que cuenta, por su nombre, en orden de sesión, separadas por
  «; ». El modo va en el nombre del umbral, y no se repite detrás de cada sesión.
- Tamaño: 151 bytes con una sesión. Como mucho, con las 12 de un modo: unos 723 bytes en el modo orden y 873 en el
  herramienta (research M10, M12). Como mucho uno por modo: dos por job.
- Uno que se cumple no da motivo ni cambia el veredicto.

## 4. La fila de `informe.md`

Una fila más en la tabla «Umbrales», con la misma forma que las demás (92 bytes la del ejemplo con 1 de 12):

```text
| `documento_sin_cotejar:claude-sonnet-5-5:orden` | 1 de 12 (8,3 %) | ≤ 0,0 % | no | sí |
```

## 5. Cómo se prueba en `make ci`

En `TestUmbralesDeJurisprudencia` (`internal/evals/informe_test.go`), con las evals y la carpeta del juez del
repositorio, las sesiones sintéticas del plan del job y un votante de votos grabados. Cada caso exige los quince
umbrales con sus valores escritos a mano, los once que deciden, los motivos de la raíz y el veredicto.

| Caso | Qué cambia en las sesiones | `documento_sin_cotejar` | Motivos de la raíz | Veredicto | Cubre |
|---|---|---|---|---|---|
| `ninguna-cuenta` (ya existe) | Nada: las 12 de cada modo cotejan, por la traza en el modo orden y por la llamada a `cita_cotejar` en el herramienta | 0 de 12 en los dos modos, `cumple: true` | Ninguno | `aprobado` | FR-002, FR-062; SC-002, SC-003 |
| `una-sin-cotejar-en-el-modo-orden` (nuevo) | La 06, `claude-sonnet-5-5`, modo orden, 01, sin su cotejo; y la misma de `claude-haiku-4-5-20251001` | Modo orden: 1 de 12, `cumple: false` | El de §3 | `fallo` | FR-001, FR-003, FR-004, FR-062; SC-002, SC-004 (Haiku) |
| `una-sin-cotejar-en-el-modo-herramienta` (nuevo) | La 06, `claude-sonnet-5-5`, modo herramienta, 01, sin su llamada a `cita_cotejar` | Modo herramienta: 1 de 12, `cumple: false` | «umbral documento_sin_cotejar:claude-sonnet-5-5:herramienta: 1 de 12 (8,3 %), y tiene que ser ≤ 0,0 %: 06-documento-que-no-es-el-pedido-herramienta-claude-sonnet-5-5-01» | `fallo` | FR-001, FR-003, FR-062; SC-002 |
| `una-sin-activar-y-sin-cotejar` (nuevo) | La 06, `claude-sonnet-5-5`, modo orden, 02, sin su cotejo y sin la activación | Modo orden: 1 de 12; y `sin_activar` del modo orden, 1 de 30 | El de `sin_activar` y el de `documento_sin_cotejar`, que nombra la 02 | `fallo` | FR-001, FR-030; US3, escenario 3 |
| `una-sin-medir-y-sin-cotejar` (nuevo) | La 06, `claude-sonnet-5-5`, modo herramienta, 01, sin su cotejo y sin medir por el límite de uso | Modo herramienta: 0 de 11, `cumple: true`; los demás umbrales de ese modo, sobre 29; el conjunto, sobre 59 | Solo el de la ejecución, que la nombra | `fallo` | FR-005, FR-062; SC-004 |

En todos, `total` es 12 —u 11 con la sesión sin medir— y no 30: las 18 sesiones por modo de las evals que no piden el
cotejo, que no cotejan, no están ni en el `total` ni en la `medida` (FR-062, SC-004). En los tres casos «sin
cotejar», las otras once sesiones del modo cotejan: es lo que distingue «cada sesión» de «ninguna sesión con cotejo»
(FR-064).

Los diez casos de antes siguen dando lo mismo, ahora con los quince elementos y los dos nuevos a 0 de 12; y
`TestInformeFinalDeLosUmbrales/como-en-el-cierre-de-h26` sigue esperando trece (research D11).

En `TestEvalsDelRepositorio/piden-el-cotejo` (`internal/evals/conjunto_test.go`), con las evals de las tres skills del
repositorio: piden el cotejo la 04, la 06, la 08 y la 09 de `jurisprudencia`, y ninguna de `boe-legislacion` ni de
`legal-core` (FR-001, FR-007).

## 6. Los dos mutantes

Cada uno cambia una sola línea, en una copia fuera del árbol (quickstart §4; research D12, M8):

| Mutante | Línea | Casos que fallan |
|---|---|---|
| 1 · «ninguna sesión con cotejo» en lugar de «cada sesión» | en `sinCotejar`, `if !respuesta.cotejo {` pasa a `if !slices.ContainsFunc(g.queLePidenElCotejo(), func(otra respuestaDelJuez) bool { return otra.cotejo }) {` | `una-sin-cotejar-en-el-modo-orden`, `una-sin-cotejar-en-el-modo-herramienta` y `una-sin-activar-y-sin-cotejar`: esperan 1 de 12 y dan 0 |
| 2 · `decide` en falso | en `umbralDeDocumentoSinCotejar`, `Decide: true,` pasa a `Decide: false,` | Los catorce casos de la tabla: los dos elementos ya no son los esperados, deciden nueve y no once, y `una-sin-cotejar-en-el-modo-orden` y `una-sin-cotejar-en-el-modo-herramienta` dan `aprobado` |

La tarea que los aplica deja lo que ve en `specs/022-h23-1-jurisprudencia-coteja-el/mutantes.md`, y ningún mutante queda
en el árbol.

## 7. Uso, de fuera adentro

Ninguna skill lee `umbrales`: son 0 invocaciones por pregunta. Nada crece con el uso del kit —cientos de normas y
miles de bloques consultados en meses—: cada job parte de las evals del repositorio y mide de nuevo sobre su commit.

| Salida | Quién la pide y cuántas veces | Tamaño | Cuándo deja de darse la señal |
|---|---|---|---|
| Los dos elementos en `informe.json` | El job, una vez, que decide con ellos; el informe final, una vez por medición del cierre (SC-009); la persona | 366 bytes el del modo orden y 378 el del herramienta, con 0 de 12; los quince, unos 4,8 KB, frente a 4 100 bytes de los trece del cierre de H26 (research M9, M13). Fijo: depende de los modos y de las evals | El `cumple: false` de un modo, en el primer job en que todas sus sesiones cotejan |
| El motivo | La persona y la reparación del cierre (FR-050); como mucho dos por job | 151 bytes con una sesión; como mucho 723 y 873 | En el primer job en que el umbral se cumple. No se acumula de un job a otro |
| La fila de `informe.md` | La persona | 92 a 99 bytes | Cada job la escribe de nuevo |
| `mutantes.md` | La revisión final, una vez | Dos mutantes, con su línea y sus casos | No da señales |
