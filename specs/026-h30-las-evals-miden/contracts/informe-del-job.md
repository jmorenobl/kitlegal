# Contrato: el informe de cada skill con lo medido y lo reutilizado

`informe.json` (y su resumen `informe.md`) que escribe `TestEjecucionDelJob` y que `scripts/evals.sh` imprime entre
sus marcas. Entidades en [../data-model.md](../data-model.md) §8 y §9; decisiones en [../research.md](../research.md)
D9, D10 y D12.

## 1. Lo que no cambia

- Todos los campos de hoy (`tasas`, `evals`, `juez`, `sesiones_sin_medir`, `fuera_de_lo_grabado`, `red`, `umbrales`,
  `duracion_de_las_sesiones`…) dicen solo lo medido en esta ejecución (FR-023). Lo que lee `scripts/workflow/serie.sh
  fila` de la 2.11.0 —las sesiones de `.evals[]`, `.juez.respuestas`, `.juez.sin_juzgar`, `.sesiones_sin_medir`,
  `.motivos` y la `medida`/`total` de `<clase>:<modelo>` y `<clase>:<modelo>:<modo>`— no ve nada reutilizado.
- En la suite completa no hay nada reutilizado: el informe es el de antes del hito más el bloque `origen` (FR-026). El
  umbral `<clase>:<modelo>` publica el mismo `umbral` y da el mismo `cumple` que hoy con todas las respuestas medidas
  (§4).
- Nombres de los trabajos, marcas `--- inicio de informe.json ---` / `--- fin de informe.json ---` y código de salida
  (0 aprobado, distinto de 0 fallo), como hoy: `scripts/workflow/cierre.sh` lo recoge igual (FR-027).

## 2. El bloque `origen` (FR-022, FR-011)

Siempre presente. Ejemplo de una ejecución con reutilización de `jurisprudencia` (dos de sus diez evals; recortado):

```json
"origen": {
  "modo": "reutilizacion",
  "evals_medidas": 1,
  "evals_reutilizadas": 9,
  "sesiones_abiertas": 12,
  "evals": [
    {
      "eval": "01-existe-con-numero-y-fecha",
      "huella": "5e0b9c41d2a7f3e86b1f04c9a2d5e7b3c8f9a0d1e2b3c4d5e6f708192a3b4c5d",
      "medida": false,
      "procedencia": {"ejecucion": 18734029311, "trabajo": 53392011870,
        "commit": "0f3c2a9e8d7b6c5a4f3e2d1c0b9a8f7e6d5c4b3a", "propuesta": 187, "fecha": "2026-10-24T09:12:03Z"},
      "aportes": [{"umbral": "afirma_lo_no_leido:claude-sonnet-5-5", "medida": 0, "total": 6},
        {"umbral": "sin_activar:claude-sonnet-5-5:orden", "medida": 0, "total": 3}],
      "tasas": [{"eval": "01-existe-con-numero-y-fecha", "modelo": "claude-sonnet-5-5", "modo": "orden", "…": "…"}]
    },
    {
      "eval": "03-por-materia",
      "huella": "a1d4…(64)",
      "medida": true,
      "motivo": "conjunto",
      "rutas": ["evals/jurisprudencia/03-por-materia.yaml"],
      "aportes": [{"umbral": "afirma_lo_no_leido:claude-sonnet-5-5", "medida": 0, "total": 6}]
    }
  ]
}
```

| Campo | Valores |
|---|---|
| `modo` | `completa` o `reutilizacion` |
| `sin_conjunto` | la causa de FR-005 si el cálculo falló en el job; ausente si no |
| `error_de_la_plataforma` | si una orden de `gh` de la decisión falló o se agotó el plazo de la búsqueda: `{"orden": "<la orden>", "error": "<su salida de error>"}`, o `"error": "plazo de 10 minutos agotado"` con la orden en curso; ausente si no. Con él, cada eval con `sin_resultado_de_main` se mide porque la búsqueda paró ahí, y cada una con `plataforma`, porque `gh` no dio los ficheros de la propuesta (research D17) |
| `evals_medidas`, `evals_reutilizadas` | cuántas; suman las evals de la skill en el commit |
| `sesiones_abiertas` | sesiones que abrió esta ejecución |
| `evals[].eval` | el mismo nombre que `eval` en `tasas`; una entrada por eval de la skill, en orden de fichero |
| `evals[].huella` | 64 caracteres hexadecimales; ausente solo con `sin_conjunto` |
| `evals[].medida` | `true` si se midió aquí |
| `evals[].motivo`, `evals[].rutas` | de una medida: `conjunto`, `comun`, `descripcion`, `sin_asignar`, `sin_conjunto` (con las rutas que la eligen), `plataforma` (sin rutas: la plataforma no dio los ficheros de la propuesta y se eligen todas), `sin_resultado_de_main` o `suite_completa` |
| `evals[].procedencia` | de una reutilizada: ejecución, trabajo, commit, propuesta y fecha de la medida en que se abrieron sus sesiones, de menos de 90 días antes de la decisión (data-model §8) |
| `evals[].aportes` | medida y total que dan sus respuestas a cada umbral de las respuestas (todos menos duración y medida del juez), los de total mayor que 0 |
| `evals[].tasas` | de una reutilizada: sus series, también las de los modelos informativos, como en su medida |

El fichero de la decisión (`-decision`) es este mismo bloque sin `sesiones_abiertas`.

## 3. Umbrales `<nombre>:reutilizadas` (FR-015)

Con alguna eval reutilizada, junto a cada umbral de las respuestas al que esas evals aportan, uno con la suma de sus
aportes:

```json
{"nombre": "sin_activar:claude-sonnet-5-5:orden:reutilizadas",
 "descripcion": "Respuestas reutilizadas de claude-sonnet-5-5 en el modo orden sin la skill activada",
 "medida": 0, "total": 27, "comparacion": "<=", "umbral": 0, "cumple": true, "decide": true}
```

`decide` es el del umbral original, salvo `<clase>:<modelo>:reutilizadas`, que no decide: es «las marcadas y el total»
de lo reutilizado que FR-015 y FR-024 piden aparte, sin decidir. Con umbral 0, que se cumplan los dos (el medido y el
reutilizado) es lo mismo que se cumpla el de la unión: el veredicto es el que daría todo medido.

## 4. El umbral en tasa `<clase>:<modelo>` (FR-024)

`medida` y `total` son las marcadas y las respuestas medidas del modelo que decide en los dos modos (*n*). Se cumple si
`medida ≤ k(n)`, con k(*n*) de data-model §9. Publica como `umbral` el de `clases.yaml` si el mayor *m* con
`m/n ≤ umbral` es k(*n*); si no, `k(n)/n`. Así `scripts/workflow/informe.sh`, que rehace `medida/total ≤ umbral`, nunca
lo marca INCOHERENTE. Ejemplo, `boe-legislacion` con 60 respuestas medidas (k(60) = 2, la proporción daría 1):

```json
{"nombre": "afirma_lo_no_leido:claude-sonnet-5-5", "descripcion": "…", "medida": 1, "total": 60,
 "comparacion": "<=", "umbral": 0.03333333333333333, "cumple": true, "decide": true}
```

Con las 120: `"umbral": 0.028`, como hoy.

## 5. Veredicto (FR-015)

`fallo` si, además de lo de hoy sobre lo medido: alguna serie que decide de una eval reutilizada tiene `pasa: false`, o
algún umbral `:reutilizadas` que decide no se cumple. Cada uno es un motivo de la raíz, con la eval o el umbral. La
duración de las sesiones y de los votos es solo la de lo medido.

## 6. `informe.md`

Una sección «Medidas y reutilizadas» con los recuentos y una línea por eval, con el motivo en palabras: «medida: rutas
de su conjunto (<rutas>)», «medida: común (<rutas>)», «medida: descripción (<rutas>)», «medida: sin asignar
(<rutas>)», «medida: sin conjunto calculable», «medida: la plataforma no dio los ficheros de la propuesta», «medida: sin
resultado reutilizable de `main`» (con «, la búsqueda paró por un error de la plataforma» si lo hubo) o «medida: suite
completa»; o «reutilizada de la ejecución <n> (propuesta #<p>, <commit corto>, <fecha>)». Con un error de la
plataforma, una línea antes de la lista: «Error de la plataforma: `<orden>`: <error>». Las series reutilizadas van en
su propia tabla.

## 7. Tamaño y vida (criterio de uso)

Lo lee quien revisa el informe, una vez por ejecución; el cierre lo copia a `gates/evals/<skill>.json`. Por eval medida,
unos 250 bytes de entrada más unos 70 por aporte (hoy, de 6 a 20 aportes según la skill y sus umbrales): de 0,7 a
1,7 KB; una reutilizada añade sus series, unos 350 bytes por serie (2 modelos × 2 modos): unos 1,4 KB más. Con las 45
evals de hoy, entre 30 y 140 KB en los tres informes, frente a los 1,8 MiB que ya tienen (922, 349 y 534 KiB en
`specs/025-h26-1-las-evals-sirven/gates/evals/`); con 300 evals, del orden de 1 MB. No crece con la historia: cada eval
lleva un solo resultado, con la procedencia de la medida en que se abrieron sus sesiones. La marca «reutilizada» deja
de darse en cuanto cambia su huella, en cuanto esa procedencia cumple 90 días —la vida de su registro—, aunque una
medida posterior la haya reutilizado, o en cuanto su medida de `main` sale de las 10 más recientes (FR-012): entonces
se mide. `error_de_la_plataforma` son unas líneas, uno por informe como mucho (la búsqueda para en el primer error), y
deja de darse en la primera ejecución en que la plataforma responde.
