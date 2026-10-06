# Pendiente de la revisión final de H24

Lo escribió el corrector de la ronda 1 (ciclo 1) y lo reescribe el de la ronda 2. **No queda nada pendiente.** De los
dos motivos del juez B en la ronda 1, el `[k]` está corregido, y el `[l][f]`, que quedó aquí sin arreglar, lo retiró el
propio juez en la ronda 2 (`gates/revision-b-r2.json`, criterio l): el peor caso del código, con 215 s por respuesta
como mucho, es 20 652 s en `evals` y 14 520 s en `medida`, bajo la cota de research D6 que calcula
`internal/evals/definicion.go` y bajo los dos topes. Nada cambia: ni `definicion.go`, ni
`.github/workflows/evals.yml`, ni ningún umbral ni tope.

## `[l][f]` El peor caso del juez y los dos topes (FR-092): retirado

**El motivo de la ronda 1.** `peorCasoDelJuez.tandas` (`internal/evals/definicion.go`) cuenta
`⌈respuestas × 6 / concurrencia⌉` tandas de 40 s, como si los votos de una respuesta se repartieran entre los hilos, y
`juzgarTodas` (`internal/evals/juez.go`) da cada respuesta entera a una gorrutina, con sus votos uno detrás de otro. De
ahí salía un peor caso de `⌈respuestas / concurrencia⌉ × 6 × 40 s`, que los dos `timeout-minutes` no cubrían, y la
primera versión de este fichero dejaba pendiente subirlos.

**Por qué se retira.** Esa cuenta pide seis votos de 40 s por respuesta, y el código no llega a ellos: un voto que
agota su tope de 35 s deja la respuesta sin juzgar y no se pide ninguno más (`juzgar` devuelve en cuanto
`votosDelNumero` da un motivo, y el del tope agotado es uno). Una respuesta ocupa como mucho cinco votos completos, de
menos de 35 s, y uno cortado a los 35 + 5 s: 215 s. Un grupo de R respuestas a C a la vez termina como mucho en
`⌈R / C⌉ × 215 s`.

**Las cifras.** El peor caso del código es 14 357 + 60 + (14 + 14 + 1) × 215 = 20 652 s en `evals` y
485 + 60 + 65 × 215 = 14 520 s en `medida`. Lo que calcula `definicion.go` es 21 097 s y 16 105 s, y los topes son
352 minutos (21 120 s) y 269 minutos (16 140 s): la cota y los dos topes quedan por encima del peor caso. El control
sigue siendo `TestDefinicionDelJob`, que compara cada `timeout-minutes` con esa cota.

Son aritmética sobre las constantes del código (`topeDelVoto` 35 s, `margenDelVoto` 5 s, `votosParaMarcar` 3 y su doble
por la repetición del nulo, `instalacionDelClaudeDelJuez` 60 s, `fueraDeLasSesiones` 485 s) y las mismas de los dos
jueces en la ronda 2; el juez A dice haberlas medido en un clon (`gates/revision-a-r2.json`, criterio l). El corrector
de la ronda 2 ha leído el código que corta la respuesta y ha repetido la aritmética; no ha ejecutado ninguna medida
propia de esos tiempos.

La cota de `definicion.go` no es la fórmula exacta del reparto por respuestas. Los dos jueces lo dejan como
observación, por debajo del umbral de materialidad: con las cifras del repositorio la cota queda por encima del peor
caso en los dos trabajos.
