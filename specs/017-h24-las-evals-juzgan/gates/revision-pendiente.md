# Pendiente de la revisión final de H24

Lo escribe el corrector de la ronda 1 (ciclo 1). De los dos motivos del juez B, el `[k]` está corregido; el `[l][f]`
queda sin arreglar, entero, porque solo se arregla en la configuración de CI.

## `[l][f]` El peor caso del juez y los dos topes (FR-092)

**El motivo.** `peorCasoDelJuez.tandas` (`internal/evals/definicion.go`) cuenta `⌈respuestas × 6 / concurrencia⌉` tandas
de 40 s, como si los votos de una respuesta se repartieran entre los hilos, y `juzgarTodas` (`internal/evals/juez.go`)
da cada respuesta entera a una gorrutina, con sus hasta 6 votos uno detrás de otro. El peor caso del código es
`⌈respuestas / concurrencia⌉ × 6 × 40 s`, y los dos `timeout-minutes` de `.github/workflows/evals.yml` no lo cubren.

**Por qué no se arregla aquí.** El arreglo tiene dos mitades que no se pueden separar:

1. Corregir la fórmula en `definicion.go`. Con ella, `TestDefinicionDelJob` compara el `timeout-minutes` real de
   `evals.yml` con el peor caso corregido y sale en rojo: es lo que tiene que hacer, porque los topes de hoy quedan por
   debajo.
2. Subir los dos topes en `.github/workflows/evals.yml`, que es configuración de CI: un corrector no la toca, y el
   guardián global (`scripts/workflow/guardian-diff.sh global`, `.github/*`) aparta entera la corrección que lo haga,
   con el arreglo del motivo `[k]` dentro.

Con la primera mitad sola, `make ci` queda en rojo sin que ningún paso del run pueda ponerlo en verde. Sin ninguna de
las dos, el control sigue en verde con una medida que no es la del código, que es lo que dice el motivo. No hay un
cambio de producto que lo resuelva sin tocar los topes: bajar el tope de un voto o su margen para que la cuenta quepa
sería ajustar a la baja, sin medirlo, un tope que ya está sin medir (research S6), y repartir votos en lugar de
respuestas cambia la regla de los votos, que van en serie porque cada uno decide si se pide el siguiente.

**Las cifras.** Son aritmética sobre las constantes del código (`topeDelVoto` 35 s, `margenDelVoto` 5 s,
`votosParaMarcar` 3 y su doble por la repetición del nulo, `instalacionDelClaudeDelJuez` 60 s, `fueraDeLasSesiones`
485 s) y sobre las cifras que ya llevan los comentarios de `evals.yml`; coinciden con las del motivo del juez B y con
la primera observación del juez A. En esta sesión no se ha ejecutado ningún test con la fórmula corregida.

| Trabajo | Hoy calcula | Peor caso del código | Tope de hoy | Tope que lo cubre |
|---|---|---|---|---|
| `evals` (`boe-legislacion`) | 14 357 + 60 + (81 + 81 + 5) × 40 = 21 097 s | 14 357 + 60 + (14 + 14 + 1) × 240 = 21 377 s (356,3 min) | 352 min (21 120 s) | 357 min |
| `medida` (`boe-legislacion`) | 485 + 60 + 389 × 40 = 16 105 s | 485 + 60 + 65 × 240 = 16 145 s (269,1 min) | 269 min (16 140 s) | 270 min |

Los dos topes nuevos siguen por debajo de las 6 horas de un trabajo de un runner de GitHub. Ningún umbral baja.

**Qué cambiar, en una sola propuesta, quien pueda tocar `evals.yml`:**

- `internal/evals/definicion.go`: en `peorCasoDelJuez`, que la duración sea, por grupo,
  `⌈respuestas / Concurrencia⌉ × votosPorRespuestaComoMucho × (topeDelVoto + margenDelVoto)`, y que `String` presente
  `⌈54 / 4⌉ × 6` en lugar de `⌈54 × 6 / 4⌉`; sus comentarios y los de `peorCasoDeLaMedida`.
- `.github/workflows/evals.yml`: `timeout-minutes: 357` en `evals` y `270` en `medida`, con sus dos comentarios
  (21 377 s, 356,3 minutos, 29 oleadas de 240 s: 14, 14 y 1; 16 145 s, 269,1 minutos, 65 oleadas).
- `internal/evals/definicion_test.go`: los casos sintéticos de `TestDefinicionDelJob` que llevan la cuenta antigua
  (el peor caso de la medida sintética, los topes «un minuto por debajo» y las duraciones de `peorCasoDelJuez`), con
  sus comentarios.
- Los textos que repiten las cifras: `contracts/job-de-evals.md` §4, `plan.md` («Performance Goals» y la fila FR-092
  de «Controles de umbral»), `cierre.md` §3 (quickstart §6), `CONTRIBUTING.md` («Job de evals») y `CHANGELOG.md`
  (*Unreleased*, la viñeta del tope).

Hasta entonces, esos textos dicen lo que el código calcula hoy, y el código calcula de menos: el trabajo `evals` puede
cortarse por su tope 257 s antes de lo que tardaría su peor caso, y el de `medida`, 5 s antes. Ese peor caso pide que
cada uno de los 6 votos de cada respuesta agote sus 40 s sin dejarla sin juzgar.
