# Contrato: `umbrales` del informe del job, con el umbral conjunto

Lo que cambia en `informe.json` y en `informe.md` de una skill con juez. Todo lo demás del informe —las tasas, la
clave `juez`, las sesiones, la duración— es lo de H24 y H25. El contrato de cada elemento sigue siendo el del ADR 0029:
`nombre`, `descripcion`, `medida`, `total`, `comparacion`, `umbral`, `cumple`, `decide`, con sus claves en ese orden.

`<modelo>` es el que decide (hoy `claude-sonnet-5-5`), `<modo>` es `orden` u `herramienta`, y `<clase>`, una clase de
`evals/<skill>/juez/clases.yaml`. Los ejemplos son los bytes que escribió el prototipo de research.md (M4 a M8).

## 1. Los elementos y su orden

Con el plan del job, trece en cada una de las dos skills con juez, nueve con `decide: true`.

**`boe-legislacion`**

| # | `nombre` | `total` | `umbral` | `decide` |
|---|---|---|---|---|
| 1 | `sin_activar:<modelo>:orden` | 54 | 0 | sí |
| 2 | `afirma_lo_no_leido:<modelo>:orden` | 54 | 0.028 | **no** |
| 3 | `cuenta_su_proceso:<modelo>:orden` | 54 | 0 | no |
| 4 | `sin_activar:<modelo>:herramienta` | 54 | 0 | sí |
| 5 | `afirma_lo_no_leido:<modelo>:herramienta` | 54 | 0.028 | **no** |
| 6 | `cuenta_su_proceso:<modelo>:herramienta` | 54 | 0 | no |
| 7 | **`afirma_lo_no_leido:<modelo>`** | 108 | 0.028 | **sí** |
| 8 | `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 212 | 0 | sí |
| 9 | `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 47 | 0 | sí |
| 10 | `duracion_de_las_sesiones:orden` | — | 900 | sí |
| 11 | `duracion_de_las_sesiones:herramienta` | — | 900 | sí |
| 12 | `duracion_del_juez:orden` | — | 900 | sí |
| 13 | `duracion_del_juez:herramienta` | — | 900 | sí |

**`jurisprudencia`**

| # | `nombre` | `total` | `umbral` | `decide` |
|---|---|---|---|---|
| 1 | `sin_activar:<modelo>:orden` | 30 | 0 | sí |
| 2 | `afirma_lo_no_leido:<modelo>:orden` | 30 | 0.017 | **no** |
| 3 | `afirma_que_existe:<modelo>:orden` | 30 | 0 | no |
| 4 | `cita_sin_documento:<modelo>:orden` | 30 | 0 | sí |
| 5 | `sin_activar:<modelo>:herramienta` | 30 | 0 | sí |
| 6 | `afirma_lo_no_leido:<modelo>:herramienta` | 30 | 0.017 | **no** |
| 7 | `afirma_que_existe:<modelo>:herramienta` | 30 | 0 | no |
| 8 | `cita_sin_documento:<modelo>:herramienta` | 30 | 0 | sí |
| 9 | **`afirma_lo_no_leido:<modelo>`** | 60 | 0.017 | **sí** |
| 10 | `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 125 | 0 | sí |
| 11 | `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 124 | 0 | sí |
| 12 | `duracion_del_juez:orden` | — | 900 | sí |
| 13 | `duracion_del_juez:herramienta` | — | 900 | sí |

Los totales son los del conjunto de evals de hoy con todas las sesiones medidas. **`legal-core`**, sin juez: `[]`.

La regla, sin ningún caso por skill: detrás de los umbrales de las respuestas de todos los modos, y delante de los de
la medida, un umbral conjunto por cada clase con `decide: true`, en el orden de `clases.yaml`. Con un plan de un solo
modo, el conjunto lleva la medida y el total de ese modo.

## 2. El umbral conjunto

| Campo | Valor |
|---|---|
| `nombre` | `<clase>:<modelo>`, sin modo |
| `descripcion` | `Respuestas de <modelo>, con todos los modos juntos, que el juez marca en <clase> con sus tres votos, sobre sus respuestas juzgadas en las evals que activan la skill` |
| `medida` | La suma de las `medida` de los recuentos por modo de la clase |
| `total` | La suma de sus `total` |
| `comparacion` | `"<="` |
| `umbral` | El de la clase en `clases.yaml` |
| `cumple` | `medida / total <= umbral` en `float64`, sin redondear; `true` si `total` es 0 |
| `decide` | `true` |

Ejemplo, 335 bytes, con 3 marcadas de 108:

```json
{"nombre":"afirma_lo_no_leido:claude-sonnet-5-5","descripcion":"Respuestas de claude-sonnet-5-5, con todos los modos juntos, que el juez marca en afirma_lo_no_leido con sus tres votos, sobre sus respuestas juzgadas en las evals que activan la skill","medida":3,"total":108,"comparacion":"<=","umbral":0.028,"cumple":true,"decide":true}
```

Y en `jurisprudencia`, 334 bytes, con 1 marcada de 60:

```json
{"nombre":"afirma_lo_no_leido:claude-sonnet-5-5","descripcion":"Respuestas de claude-sonnet-5-5, con todos los modos juntos, que el juez marca en afirma_lo_no_leido con sus tres votos, sobre sus respuestas juzgadas en las evals que activan la skill","medida":1,"total":60,"comparacion":"<=","umbral":0.017,"cumple":true,"decide":true}
```

Cuándo se cumple, con los valores de `clases.yaml`:

| Skill | Juzgadas | Se cumple con | No se cumple con |
|---|---|---|---|
| `boe-legislacion` | 108 | 3 (0,0278) | 4 (0,0370) |
| `boe-legislacion` | 107 | 2 | 3 (0,0280, por encima de 0.028) |
| `boe-legislacion` | 104 | 2 (0,0192) | 3 (0,0288) |
| `jurisprudencia` | 60 | 1 (0,0167) | 2 (0,0333) |
| `jurisprudencia` | 58 | 0 | 1 (0,0172) |

No depende de cómo se reparten las marcadas entre los modos.

## 3. El recuento por modo

El elemento `<clase>:<modelo>:<modo>` de hoy, con dos cambios cuando la clase decide: `umbral` es el de la clase —el
mismo que el del conjunto— y `decide` es `false`. Su `cumple` es la comparación del contrato, y solo informa.

Ejemplo, 342 bytes: las 3 marcadas de un job de `boe-legislacion` están en el modo herramienta.

```json
{"nombre":"afirma_lo_no_leido:claude-sonnet-5-5:herramienta","descripcion":"Respuestas de claude-sonnet-5-5 en el modo herramienta que el juez marca en afirma_lo_no_leido con sus tres votos, sobre sus respuestas juzgadas en las evals que activan la skill","medida":3,"total":54,"comparacion":"<=","umbral":0.028,"cumple":false,"decide":false}
```

El conjunto de ese mismo informe es el primer ejemplo de §2: se cumple, y el veredicto es `aprobado`, sin motivos.

Y en `jurisprudencia`, con su única marcada en el modo herramienta: el conjunto es el segundo ejemplo de §2, y el
veredicto, `aprobado`.

```json
{"nombre":"afirma_lo_no_leido:claude-sonnet-5-5:herramienta","descripcion":"Respuestas de claude-sonnet-5-5 en el modo herramienta que el juez marca en afirma_lo_no_leido con sus tres votos, sobre sus respuestas juzgadas en las evals que activan la skill","medida":1,"total":30,"comparacion":"<=","umbral":0.017,"cumple":false,"decide":false}
```

El de una clase con `decide: false` (`cuenta_su_proceso`, `afirma_que_existe`) no cambia en un byte, y esa clase no
tiene umbral conjunto.

## 4. El motivo

Un umbral conjunto que no se cumple da un motivo de la raíz, donde hoy va el del recuento por modo: con los de los
umbrales, en su orden. Uno que se cumple no da ninguno. Un recuento por modo no da ninguno nunca.

```text
umbral <clase>:<modelo>: <medida> de <total> (<p> %), y tiene que ser ≤ <u> %: <sesión> (modo <modo>): «<frase 1>» · «<frase 2>» · «<frase 3>»; <sesión> (modo <modo>): «…» · «…» · «…»
```

- `<p>` y `<u>`, con un decimal y coma, como en todo umbral con `total`. El resultado lo dice `cumple`, no el
  porcentaje escrito: 3 de 107 se escribe «3 de 107 (2,8 %), y tiene que ser ≤ 2,8 %».
- Las respuestas son las marcadas de todos los modos: las del modo orden delante y, dentro de un modo, en orden de
  sesión. Cada una lleva el nombre de su modo entre paréntesis, detrás de su sesión.
- Las frases son las de los tres votos que la marcan; de un voto nulo cuenta la de su repetición, como hoy.

Ejemplo, con 2 marcadas en cada modo de `boe-legislacion` (las frases, recortadas aquí con `…`):

```text
umbral afirma_lo_no_leido:claude-sonnet-5-5: 4 de 108 (3,7 %), y tiene que ser ≤ 2,8 %: 01-sintetica-claude-sonnet-5-5-01 (modo orden): «…» · «…» · «…»; 02-sintetica-claude-sonnet-5-5-01 (modo orden): «…» · «…» · «…»; 01-sintetica-herramienta-claude-sonnet-5-5-01 (modo herramienta): «…» · «…» · «…»; 02-sintetica-herramienta-claude-sonnet-5-5-01 (modo herramienta): «…» · «…» · «…»
```

Y en `jurisprudencia`, con una en cada modo:

```text
umbral afirma_lo_no_leido:claude-sonnet-5-5: 2 de 60 (3,3 %), y tiene que ser ≤ 1,7 %: 07-resumen-de-una-conocida-claude-sonnet-5-5-01 (modo orden): «…» · «…» · «…»; 07-resumen-de-una-conocida-herramienta-claude-sonnet-5-5-01 (modo herramienta): «…» · «…» · «…»
```

## 5. `informe.md`

La tabla «Umbrales» gana la fila del conjunto, en su sitio, y las del recuento por modo cambian su condición y su
última celda:

```text
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 2,8 % | sí | no: solo se publica |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 3 de 54 (5,6 %) | ≤ 2,8 % | no | no: solo se publica |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 3 de 108 (2,8 %) | ≤ 2,8 % | sí | sí |
```

## 6. Lo que no cambia

- `sin_activar`, `cita_sin_documento`, los dos de la medida y las duraciones: su nombre, su medida, su umbral y su
  `decide`.
- La respuesta sin juzgar: sigue en el `total` de su modo y en el del conjunto, y fuera de toda `medida`; da `fallo`
  con su motivo «de la ejecución, no de la skill», también con el conjunto cumplido.
- Las respuestas de la eval sin binario ni servidor: no son de ningún modo y no entran en ningún umbral.
- El instrumento sin medir: el informe no lleva ningún umbral del juez, tampoco el conjunto.
- La clave `juez`, los votos y la regla de los tres votos.
- La salida de `make evals-sondeo`: una línea por clase con las marcadas sobre todas sus respuestas juzgadas, sin
  umbral ni veredicto.

## 7. Uso, de fuera adentro

Ninguna skill pide estas salidas ni las lee un `SKILL.md`: 0 invocaciones por pregunta. Nada crece con lo que se haya
consultado con el kit: cada job parte de las evals del repositorio.

| Salida | Quién la pide y cuántas veces | Tamaño | Cuándo deja de darse su señal |
|---|---|---|---|
| `umbrales` | El job, una vez; el informe final, una vez por medición del cierre; la persona | 13 elementos: 3 822 bytes en `boe-legislacion` y 4 099 en `jurisprudencia` (eran 12, con 3 482 y 3 757). Fijo: depende de las clases y de los modos | Se miden de nuevo en cada job. El `cumple: false` del conjunto, en el primer job en que las marcadas de los dos modos caben en la tasa; el de un recuento por modo, cuando caben las de ese modo |
| El motivo del conjunto | La persona y la reparación del cierre, que no cambia por él ninguna skill. Como mucho uno por skill y por job | 90 bytes de cabecera y unos 440 por respuesta marcada: 1,9 KB con las 4 de `boe-legislacion` y 1 KB con las 2 de `jurisprudencia`, que son los más cortos que pueden darse con las 108 y las 60 juzgadas; con menos juzgadas pueden llevar menos: 3, de 107 o de 104, y 1, de 58. Como mucho, las juzgadas: 108 y 60 respuestas, unos 48 KB y 26 KB | En el primer job en que el conjunto se cumple. No se acumula de un job a otro |
| La fila de `informe.md` | La persona que lee el informe | Una fila más, de 85 bytes en el ejemplo de §5 | Con cada job |

## 8. Los dos mutantes

Sobre el método que compone el umbral conjunto (`umbralesConjuntos`, en `internal/evals/umbrales.go`):

| Mutante | Qué cambia | Qué casos lo ven por el veredicto |
|---|---|---|
| Un solo modo | El conjunto suma solo el primer modo del plan | `TestUmbralConjunto`: `3-y-0` y `2-y-1` dan `fallo`. `TestUmbralesDeJurisprudencia`: `una-marcada-en-el-modo-orden` da `fallo`. Todos los casos ven además otra `medida` y otro `total` |
| Sin decidir | `decide` del conjunto en `false` | `TestUmbralConjunto`: `4-y-0`, `2-y-2` y `104-con-3` dan `aprobado`. `TestUmbralesDeJurisprudencia`: `dos-marcadas-en-el-modo-orden` y `una-marcada-en-cada-modo` dan `aprobado`. Deciden 8 y no 9 |

Cómo se aplican, en una copia fuera del árbol: [quickstart](../quickstart.md), escenario 6.
