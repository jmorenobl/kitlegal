# Contrato: el trabajo de `jurisprudencia` en el flujo `evals`

Lo que cambia en `.github/workflows/evals.yml` y en lo que `TestDefinicionDelJob` exige de él. Lo demás es lo de
`specs/017-h24-las-evals-juzgan/contracts/job-de-evals.md` y `specs/013-h7-3-el-umbral-de/contracts/ejecucion-del-job.md`.
Las cifras de §3 las da `peorCaso` del repositorio con las diez evals y los 249 casos (research M6).

## 1. Lo que cambia en la definición (FR-050, FR-070, FR-071)

Tres valores y una descripción:

```yaml
on:
  workflow_dispatch:
    inputs:
      medir_al_juez:
        description: "Ejecuta la medida del juez de las skills que lo tienen en lugar de sus evals (ADR 0037)"

jobs:
  evals:
    strategy:
      matrix:
        skill: [boe-legislacion, jurisprudencia, legal-core]
        include:
          - skill: boe-legislacion
            concurrencia: 4
            objetivo_de_duracion: 900
          - skill: jurisprudencia
            concurrencia: 4          # antes, 1
            objetivo_de_duracion: 0
          - skill: legal-core
            concurrencia: 1
            objetivo_de_duracion: 0
  medida:
    strategy:
      matrix:
        skill: [boe-legislacion, jurisprudencia]   # antes, [boe-legislacion]
        include:
          - skill: boe-legislacion
            concurrencia: 4
          - skill: jurisprudencia                   # nueva
            concurrencia: 4
```

Y los comentarios del fichero que hoy dicen que `jurisprudencia` no tiene juez, que va de una en una y que su peor
caso es de 20 341 s, que pasan a decir lo de §3.

## 2. Lo que no cambia

- Los dos topes: `timeout-minutes: 352` en `evals` y `269` en `medida` (FR-072).
- `MODELO_DEL_JUEZ` (`claude-opus-5-5`) y `VERSION_DE_CLAUDE_CODE_DEL_JUEZ` (`2.1.289`), fijados una vez para las dos
  skills, y repetidos en `medida` (FR-033). El modelo que decide y la versión de las sesiones.
- La condición de `medida`, que solo corre con la etiqueta `evals-medir-juez` o con la entrada `medir_al_juez`, y la de
  `tanda`, que no corre con ninguna de las dos. Sin selector por skill: con la etiqueta o con la entrada corren los
  dos trabajos de la matriz, `medida del juez (boe-legislacion)` y `medida del juez (jurisprudencia)`, cada uno con su
  resultado: la matriz ya lleva `fail-fast: false` (research S8).
- El objetivo de duración de `jurisprudencia`, 0: no publica `duracion_de_las_sesiones:<modo>`.
- Los pasos de los dos trabajos, `scripts/evals.sh`, `scripts/evals-medir-juez.sh`, `scripts/evals-voto.sh` y el
  `Makefile`: reciben la skill y no nombran ninguna (research V23).

## 3. Los peores casos (FR-072)

| Trabajo | Con cuatro a la vez | De una en una | Tope |
|---|---|---|---|
| `evals (jurisprudencia)` | 12 577 s (209,6 min) = 485 + (⌈61 / 4⌉ + ⌈60 / 4⌉) × 272 + 60 + (⌈30 × 6 / 4⌉ + ⌈30 × 6 / 4⌉) × 40 | 47 857 s | 352 min = 21 120 s |
| `medida del juez (jurisprudencia)` | 15 505 s (258,4 min) = 485 + 60 + ⌈249 × 6 / 4⌉ × 40 | 60 305 s | 269 min = 16 140 s |

- 61 y 60 son las sesiones de cada modo: diez evals, dos modelos y tres repeticiones, y la prueba de red, que el
  cálculo cuenta en el modo orden aunque el trabajo no la lleve. 30 son las respuestas que el juez juzga en cada modo.
- Los de `boe-legislacion` (21 097 s y 16 105 s) y el de `legal-core` (12 181 s) no cambian.
- Es un cálculo del peor caso, no una medida de lo que tarda.

## 4. `TestDefinicionDelJob` (FR-103)

Lo que exige de más o de otro modo, con una línea por clave que no coincide, como hoy:

| Clave | Esperado |
|---|---|
| `jobs.evals.strategy.matrix.include`, skill `jurisprudencia` | `{concurrencia: 4, objetivo_de_duracion: 0}` |
| `jobs.medida.strategy.matrix.skill` | Las skills de la matriz de `evals` que tienen carpeta de juez, en su orden: con las evals del repositorio, `[boe-legislacion, jurisprudencia]` |
| `jobs.medida.strategy.matrix.include`, cada skill | `{concurrencia: n}`, la de su trabajo `evals` |
| `jobs.evals.timeout-minutes` | Cubre el peor caso de cada skill, con las evals del repositorio |
| `jobs.medida.timeout-minutes` | Cubre el peor caso de la medida de cada skill de su matriz, con los casos del repositorio |

Sigue exigiendo lo de hoy: la condición de `medida` es exactamente la de su etiqueta o su entrada, no tiene `needs`,
repite el modelo y la versión del juez, y ni `tanda` ni `evals` nombran la etiqueta de la medida.

Con las evals y los casos del repositorio y los topes de hoy, las dos mutaciones de SC-003 dan estas líneas:

- **`jurisprudencia` de una en una en `evals`**:
  `jobs.evals.strategy.matrix.include, skill jurisprudencia: vale {concurrencia: 1, objetivo_de_duracion: 0}, y lo esperado es {concurrencia: 4, objetivo_de_duracion: 0}`
  y
  `jobs.evals.timeout-minutes: vale 352 (21120 s), y lo esperado es al menos el peor caso de jurisprudencia, 47857 s = 485 s + (⌈61 / 1⌉ + ⌈60 / 1⌉) × (22 s + 240 s + 10 s) + 60 s + (⌈30 × 6 / 1⌉ + ⌈30 × 6 / 1⌉) × (35 s + 5 s)`.
- **`jurisprudencia` de una en una en `medida`**:
  `jobs.medida.strategy.matrix.include, skill jurisprudencia: vale {concurrencia: 1}, y lo esperado es {concurrencia: 4}`
  y
  `jobs.medida.timeout-minutes: vale 269 (16140 s), y lo esperado es al menos el peor caso de jurisprudencia, 60305 s = 485 s + 60 s + (⌈249 × 6 / 1⌉) × (35 s + 5 s)`.

Y una más, de la matriz: sin `jurisprudencia` en `medida`,
`jobs.medida.strategy.matrix.skill: vale [boe-legislacion], y lo esperado es [boe-legislacion, jurisprudencia]`.

Los subtests que cambian:

- `del-repositorio`: la definición del repositorio no da ninguna línea.
- `peor-caso`: el de `jurisprudencia` es `{[61, 60], 4}` con su juez `{[30, 30], 4}`, 12 577 s, y el de su medida,
  `{249, 4}`, 15 505 s, cada uno con sus términos.
- `tope-con-las-evals-del-repositorio`: los topes que dejan de cubrir a cada skill, con el orden de hoy —el de
  `boe-legislacion`, 21 097 s; el de `jurisprudencia`, 12 577 s; el de `legal-core`, 12 181 s—, y las dos mutaciones de
  arriba.
- `sinteticas`: la definición del contrato del test lleva `jurisprudencia` a cuatro y en la matriz de `medida`, y las
  evals sintéticas le dan carpeta de juez; los casos que hoy la dan por una skill sin juez y de una en una se
  recalculan con la misma fórmula.

## 5. El límite de ritmo (FR-073)

Con las tres skills a la vez son nueve sesiones contra la misma suscripción: cuatro, cuatro y una. No se han medido
(research S2). Lo que el job hace con el límite no cambia: una sesión que corta queda sin medir, con su motivo, que no
es de la skill, y el informe publica `reintentos_por_limite_de_ritmo`. Si pasa en el cierre, el veredicto es `fallo`
por la ejecución y llega al informe final: el run no cambia la concurrencia ni los topes.

## 6. Uso, de fuera adentro

| Salida | Quién la consume y cuántas veces | Tamaño | Cuándo deja de darse |
|---|---|---|---|
| Las líneas de `TestDefinicionDelJob` | Quien cambia la definición del job, las evals o los casos, en `make ci` | Una línea por clave, de 120 a 260 bytes | Cuando la definición vuelve a ser la del contrato y cada tope cubre su peor caso |
| El registro de cada trabajo `medida del juez (<skill>)` | La persona que puso la etiqueta, una vez por lanzamiento | La medida y, si no se cumple, sus casos (contracts/medida-y-casos.md §9) | Una por lanzamiento |
