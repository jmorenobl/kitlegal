# Research: H7.4 · `boe-legislacion` sin el estado de la comprobación ni lo que no ha leído, y un job que juzga la respuesta a la pregunta

Modo desatendido: cada decisión se tomó con el «Criterio de decisión autónoma» de la constitución y lleva la alternativa
rechazada y por qué. Toda afirmación sobre Claude Code, GitHub Actions, `gh`, `gosec`, PowerShell, Go o el propio
repositorio remite a la tabla V, comprobada en local; lo que no se pudo comprobar así son los supuestos S1-S9, algunos
con una observación en la plataforma que los respalda sin sustituir a la comprobación local. Nada de este research lo
hace una persona a mitad del run. Las medidas se tomaron con dos herramientas de investigación de un solo uso en
`bin/h74/` (directorio ignorado por `.gitignore`, borrado al terminar): `calibrar`, que aplica una lista candidata
con `evals.ExtraerExpresionesProhibidas` a las respuestas de un informe tras quitar las formas fijas, y `prosa`, que
aplica a un `SKILL.md` la extracción de párrafos de la subprueba `prosa-de-la-skill` (copiada de
`internal/evals/conjunto_test.go:1617-1669`).

## V · Verificaciones en local

| Id | Qué | Dónde se comprobó |
|---|---|---|
| V1 | Los cinco informes de la evidencia, con 93 sesiones cada uno: H7.1 (`a8aeb0d`), H7.2 (`fbdab2e`) y H7.3 (`196ee05`) versionados en `specs/01{1,2,3}-…/gates/evals/boe-legislacion.json`; el de `6ab3add` en `eb6b4c8:specs/013-h7-3-el-umbral-de/gates/evals/boe-legislacion.json` y el de `eb6b4c8` en `196ee05:` con la misma ruta (cada uno con su `commit`). El de la ejecución 36712391835 (línea de base, `ee8e7ee`) sale de su registro con el mismo corte que `recoger_evals` (marcas `--- inicio/fin de informe.json ---`, `scripts/workflow/cierre.sh:64-84`): veredicto `fallo` por la 04 (1 de 3), `umbrales` 0 de 51, 0 de 30 y 421 s, 93 sesiones | `git show`; `gh run view 36712391835 --job 109877076393 --log` (lectura, como el cierre) |
| V2 | Las sesiones 04-01 y 04-03 de la línea de base: `activada=false`, 0 invocaciones; 04-01 termina «Respondo de memoria, sin haber consultado el texto consolidado del BOE. Si necesita citarlo con precisión, conviene comprobar la redacción vigente…»; 04-03, «Respondo de memoria y no he contrastado el texto consolidado en el BOE. Si lo necesitas para citarlo, puedo comprobarlo con la skill de legislación.»; 04-02 activa, lee `a66` y pasa | V1, `jq` sobre la línea de base |
| V3 | La 03-01 de `eb6b4c8`: el `respuesta` juzgado es «Esa tarea en segundo plano era solo una búsqueda auxiliar que ya no hace falta… No afecta a la respuesta ya entregada sobre el artículo 22 de la LRBRL.», con `fin_de_la_sesion` `result success` y el motivo «cita ausente»; sus invocaciones leen `a22` y comprueban | V1 |
| V4 | Claude Code **2.1.284** (la versión del job, `VERSION_DE_CLAUDE_CODE`; `~/.local/share/claude/versions/2.1.284`, Mach-O de un fichero) emite en `stream-json` un mensaje `system` con `subtype:"task_notification",task_id:…,tool_use_id:…` al terminar una tarea en segundo plano; el binario tiene 34 apariciones de `task_notification` y 13 de `<task-notification>` | `grep -a -o` sobre el binario |
| V5 | `LeerSesion` toma la respuesta de **cada** `result` y se queda con la del último («Cada result deja la respuesta en la suya, de modo que cuenta la del último», `internal/evals/sesion.go:414-452`); `Terminada`, `Fin` y `ErrorDelResultado` salen del último mensaje (`:199-215`); `leerSystem` no lee de un `system` que no sea `init` o `api_retry` más que su `subtype` (`:315-347`) | fuente |
| V6 | `recontarExpresiones` cuenta como respuesta toda sesión juzgada, no ilegible, no sin medir y de eval que activa la skill, **termine o no** (`internal/evals/informe.go:762-799`); el sondeo usa la misma función (`internal/evals/sondeo.go:244`) | fuente |
| V7 | `umbralesDelInforme` da uno de las expresiones por modelo del recuento (solo si la skill tiene lista) y el de la duración si hay objetivo; `motivosDeLosUmbrales` da «umbral <nombre>: <medida> de <total> (<p> %), y tiene que ser ≤ <u> %» por cada uno que decide y no se cumple, salvo el de la duración (`internal/evals/umbrales.go:74-166`) | fuente |
| V8 | El informe final lee de cada celda `evals:<skill>:<nombre>` el `nombre` con `cut -d: -f3-` y lo busca en `umbrales` (`scripts/workflow/informe.sh:259-264`); la celda la extrae `controles_de_celda` con `evals:[a-z0-9-]+:[A-Za-z0-9_.:-]+` (`scripts/workflow/comun.sh:164-166`): `sin_activar:claude-sonnet-5-5` y `redaccion_no_leida:claude-sonnet-5-5` casan | fuente |
| V9 | El cierre: `medir` espera mientras alguna comprobación de flujo esté `pending` y cuenta rojas las `fail` y `cancel` de cualquier flujo; `recoger_evals` copia el informe de cada comprobación del flujo `evals` en `pass` o `fail` cuyo nombre casa `^evals \(([a-z0-9-]+)\)$`, y escribe siempre en `gates/evals/<skill>.json` (una segunda con el mismo nombre lo sobrescribiría, y una sin informe lo borraría) | `scripts/workflow/cierre.sh:64-84, 147-157` |
| V10 | Esquema oficial de flujos de GitHub Actions (`workflow-v1.0`, extraído en H7.3 de la extensión de VS Code `github.vscode-github-actions` 0.31.5 a `/tmp/gha/chunk.txt`, research V6 de H7.3): el `if` de un trabajo (`job-if`) admite los contextos `github, inputs, vars, needs` y las funciones de estado, **no `matrix`**; el `if` de un paso (`step-if`) admite además `strategy, matrix, steps, job, runner, env`; `permissions` tiene la clave `actions` («Actions workflows, workflow runs, and artifacts»); el contexto `github` tiene `run_id` («A unique number for each workflow run within a repository. This number does not change if you re-run the workflow run.»); y el `group` de `concurrency`: el que llega con otro en curso queda `pending`, y «Any previously pending job or workflow in the concurrency group will be canceled» (research V6 de H7.3) | `grep -o` sobre el extracto |
| V11 | `gh` 2.101.0: `gh run list` filtra por `--workflow` y `--commit SHA`, con `--json` y los campos `databaseId`, `status`, `createdAt`, `event`, `headSha`…; `gh run view <run-id> --json jobs`; `GH_TOKEN` y `GH_REPO` («specify the GitHub repository in the `[HOST/]OWNER/REPO` format») los lee del entorno | `rtk proxy gh run list --help`, `gh run view --help`, `gh help environment` |
| V12 | `gosec` v2.28.0 (el de `tools/golangci-lint/go.mod:168`), G204 (`rules/subproc.go:62-94`): salta el argumento 0 si es un parámetro o un campo y exige que **todo otro argumento** se resuelva a constante; `exec.CommandContext(ctx, "sh", "-c", ordenConstante)` no da nada. G702 tiene por fuentes `os.Args`, `os.Getenv`, `*http.Request` y lectores de `bufio`; ni los valores de `flag` ni `os.Environ()` lo son (research V8 de H7.3) | `/Users/jorge/go/pkg/mod/github.com/securego/gosec/v2@v2.28.0/rules/subproc.go` |
| V13 | El peor caso del trabajo de una skill es `485 s + ⌈sesiones / concurrencia⌉ × (22 s + 240 s + 10 s)`, con las sesiones del plan y la prueba de red (`internal/evals/definicion.go:184-250`); con 20 evals (60 del modelo que decide, 36 del informativo) y la prueba de red, 97 sesiones a 4: 485 + 25 × 272 = **7 285 s (121,4 min)**, más que los 120 de hoy; `legal-core`, 19 a 1: 5 653 s | fuente, cálculo |
| V14 | El conjunto de `boe-legislacion` admite de 10 a 20 evals con exactamente 10 positivas que deciden (`internal/evals/conjunto.go:205-207`): con la eval nueva, 20 | fuente |
| V15 | Las respuestas grabadas desde H4 del bloque `a1-30` (5,6 KB; vigencias 20180309 y 20200206) y `da-3` (15,4 KB; 20180309 y 20230101) de `BOE-A-2017-12902`, su índice (192,8 KB) y sus metadatos están en `internal/source/boe/testdata/boe.legislacion-consolidada/`, que es de `UnionDeGrabaciones` (`internal/evals/grabaciones.go:38-40`). Las derivadas del grafo previo las escribe `TestGrabacionesDerivadas -actualizar-derivadas` desde su grabación (`internal/app/grafo_test.go:2549-2556, 2653-2700`) y el mismo test comprueba cada una byte a byte y que su carpeta no tiene ficheros sin entrada (`:2714-2735, 2784-2802`) | `ls`, `grep -o fecha_vigencia`, fuente |
| V16 | El grafo previo de una eval se prepara con su carpeta de derivadas copiada **encima** de la unión de grabaciones (`prepararGrafoPrevio`, `internal/evals/preparar.go:336-355`; lo mismo hace `expresiones-en-los-bloques`, `internal/evals/conjunto_test.go:1390-1391`): un bloque del grafo previo sin su derivada en la carpeta se leería con la grabación de H4, que da la redacción vigente. `compruebaLaSesion` exige que `graph check` dé en la sesión exactamente las clases de `hallazgos` de la eval (`:1981-1983`) | fuente |
| V17 | `TestLeerSesion` exige un directorio de `internal/evals/testdata/sesiones/leer-sesion/` por caso, y un caso por directorio (`internal/evals/sesion_test.go:192-198`) | fuente |
| V18 | La subprueba `prosa-de-la-skill` quita los bloques delimitados, la región generada y el código en línea antes de buscar (`internal/evals/conjunto_test.go:1545-1669`); `expresiones-de-la-skill` aplica la lista, sin quitar nada, a las formas escritas de los avisos y los hallazgos y a cada bloque `text` de `SKILL.md` (`:1460-1497`): con «consulta anterior» en la lista, el bloque de la regla 7 la haría fallar | fuente |
| V19 | `SKILL.md` v0.1.3: 270 líneas; la `description` tiene 722 caracteres (máximo 1024, `internal/skills/frontmatter.go:53, 438`); con la lista nueva, `prosa` da 9 párrafos con defecto (líneas 65, 72, 82, 120, 177, 185, 198 y 268 con «lectura anterior», y 260 con «la comprobación de la redacción»). El prototipo v0.1.4 de [contracts/skill-boe-legislacion.md](./contracts/skill-boe-legislacion.md): 294 líneas, ≈ 925 caracteres de `description` y 0 defectos (la v0.1.4 que entra: 295 líneas, con el párrafo que cierra «Redacción modificada», que se queda, supuesto T010; y 924 caracteres; 298 tras la reparación del cierre, C11, V28, D22); fuera de la prosa quedan «lectura anterior» en la región generada (línea 229 de v0.1.3, de `--describe`), «se consultó antes» en el bloque de la línea `⚠ REDACCIÓN MODIFICADA:` y «consulta anterior» en el bloque de la regla 7 | `prosa`, `wc` |
| V20 | Medidas de la lista candidata con `calibrar` (formas fijas quitadas) — ver «Calibrado» | `calibrar` sobre V1 |
| V21 | Ninguna expresión de la lista nueva aparece en las respuestas grabadas de los bloques que leen las evals (`a21`, `a1-30`, `da-3`, `a22`, `a66`, `a59`, `a17`, `a25`, `a20`, `a140`, `a38`, `a42` y la derivada de `a1-30`): `grep -i -E` de las 35 formas sobre `*texto_bloque*` de las dos carpetas y `grafo-previo/*` sale con 1 | `grep` |
| V22 | `pwsh` y `powershell` no están en esta máquina (`which` → «not found») | `which` |
| V23 | El sondeo comprueba los argumentos y la credencial antes de construir nada (`comprobarElSondeo`, `internal/evals/sondeo.go:509-534`), con los errores de los argumentos unidos, uno por línea (`:547-574`), y `errSinSuscripcion` (`:45-46`); `TestSondeo` falla con `require.NoError` (`internal/evals/job_test.go:218-232`) y `scripts/evals-sondeo.sh` imprime entero `go-test.log` a la salida de error y sale con 1 si `go test` falla (`:37-45`) | fuente |
| V24 | La definición del job hoy: `cambios` solo en `pull_request` que no es `labeled`; `evals` con `needs: [cambios]`, `concurrency` por commit y skill con `cancel-in-progress: false`, matriz de dos skills y `timeout-minutes: 120`; `TestDefinicionDelJob` la lee con `leerDefinicionDelJob` y comprueba claves (`internal/evals/definicion.go:93-172`, `definicion_test.go:61-221`). El `-commit` ya es una bandera de los puntos de entrada (`internal/evals/job_test.go:48`) | fuente |
| V25 | `CONTRIBUTING.md:592-594` dice que un segundo disparo «espera a que termine […] y abre su tanda después», y `:655` que el motivo lleva el texto «del último `result` con `is_error`»: el hito deja falsa la primera | `grep` |
| V26 | GNU Make (3.81 en esta máquina) sale con 2 si hay algún error, también el de una receta (`man make`, «EXIT STATUS»): `make evals-sondeo` da 2 cuando el guion sale con 1. `env -u <nombre>` quita una variable del entorno en el `env` de macOS (`man env`) | `man make`, `man env` |
| V28 | La medición del cierre sobre `4cb52cb` (C1-C10; `gates/evals/boe-legislacion.json`, ejecución 36769593099): `fallo`; `expresiones_prohibidas:claude-sonnet-5-5` **15 de 54** (27,8 %), `sin_activar` 0 de 54, `redaccion_no_leida` 0 de 54, 450 s, Haiku 4.5 0 de 30, `red` vacío, ninguna sin medir; las series 04 y 07 pasan 1 de 3, las dos por expresiones. Las 15: 01-01, 04-01, 04-02, 06-02, 07-01, 07-03, 08-01, 10-03, 14-01, 14-03, 16-02, 16-03, 17-02, 18-03 y 20-01. Catorce cuentan el resultado negativo de `kitlegal graph check`, trece en el párrafo de vigencia con que cierran («**Vigencia:** el bloque no trae avisos de vigencia. … No se ha detectado ninguna redacción modificada respecto de consultas anteriores.», 01-01; «El artículo no lleva avisos de vigencia ni se ha detectado cambio de redacción.», 04-01; «no consta que su redacción haya cambiado desde una consulta anterior», 08-01 y 16-02; «la comprobación de cambios de redacción no encontró ninguno», 07-03, 10-03 y 14-03; «ni cambios de redacción respecto a consultas anteriores», 14-01; 18-03 lo dice en su propio párrafo). La 20-01 repite las dos líneas `⚠ REDACCIÓN MODIFICADA:` con sus palabras («Ambos bloques tienen hoy una redacción distinta de la que se consultó antes.»). Las formas marcadas —`consulta(s) anterior(es)` en 7, `cambio(s) de redacción` en 8, `la comprobación de …` en 4, `haya cambiado desde` en 1, `se consultó antes` en 1— son las palabras de las dos formas fijas y de la etiqueta: «lectura anterior», que la prosa de v0.1.3 enseñaba, no aparece en ninguna. Las 39 respuestas limpias del modelo que decide cierran con el mismo párrafo de vigencia sin nombrar `kitlegal graph check` («El bloque no trae avisos de vigencia. Su redacción vigente es de …, dada por …», 03-01, 06-01, 09-01…) | `jq` sobre el informe del cierre |
| V29 | Sondeo local con C11 (`make evals-sondeo SKILL=boe-legislacion EVALS=01,04,06,07,08,10,14,16,17,18,20 MODELO=claude-sonnet-5-5 REPETICIONES=3`, las once evals con alguna respuesta marcada en V28): las once series **3 de 3**, **0 de 33** respuestas con alguna expresión de la lista (V28: 15 de esas 33 posiciones), ninguna sin medir ni sin terminar, código 0. No es un veredicto (no lee la traza): el veredicto es el job de cierre sobre la cabeza con C11 | `scripts/evals-sondeo.sh` |
| V30 | La medición del cierre sobre `ad68a70` (C1-C11; `gates/evals/boe-legislacion.json`, ejecución 36774880631): `fallo` solo por `redaccion_no_leida:claude-sonnet-5-5` **1 de 54** (1,9 %; umbral 0); `expresiones_prohibidas:claude-sonnet-5-5` 1 de 54 (la misma respuesta; cumple), `sin_activar` 0 de 54, 451 s, Haiku 4.5 1 de 30 (05-03, `código 5`, sin leer el bloque; informativo), `red` vacío, ninguna sin medir, todas las series que deciden 3 de 3, `legal-core` aprobado. La única es la 18-01 (norma derogada, informativa; la serie pasa 2 de 3), con la forma `vigente hasta`: «**Art. 42 de la Ley 30/1992, «Obligación de resolver»** [BOE-A-1992-26318, bloque a42]. Su redacción vigente hasta la derogación es la de la Ley 4/1999 (vigencia desde 14/04/1999).», con sus dos avisos y su cita bien. Las otras dos lo dicen sin la forma: «Fue modificado por la Ley 4/1999 (BOE-A-1999-847), con vigencia desde el 14/04/1999.» (18-02) y «El artículo quedó redactado por el art. 1.10 de la Ley 4/1999 (BOE-A-1999-847), con vigencia desde el 14/04/1999.» (18-03). El mismo razonamiento, sin forma, ya estaba en la 18-02 del informe de H7.1 («siendo esa la última redacción vigente (fecha de vigencia 1999-04-14) antes de la derogación»). En el calibrado, `vigente hasta` es la única forma de la clase B que marca la 19-02 de `196ee05` («…es distinta de la que se consultó antes (vigente hasta el 9 de marzo de 2018)…»); la 19-01 la marca `ya no exige` | `jq` sobre el informe del cierre y sobre los de H7.1 y H7.3 |
| V27 | Los textos sintéticos de los tests de hoy con alguna forma nueva de la lista: `internal/evals/juzgar_test.go:642` («La redacción ha cambiado desde la consulta anterior.») y `internal/evals/prohibidas_test.go:174` (la frase de la regla 7); `hallazgos_test.go:141` lleva la primera, pero `ExtraerHallazgos` no aplica la lista. Con `se consultó antes`: la línea de la redacción sin cita, entera, en `juzgar_test.go:111-112` y `:136-137`, `prohibidas_test.go:156-157` y `hallazgos_test.go:82-83` (la forma fija la quita: no cambian), y fuera de esa forma en dos casos de `TestCondicionesDeLaConsultaRepetida`, que aplica la lista del repositorio: `consulta_repetida_test.go:73-74` («…la redacción que se consultó antes ha sido sustituida.») y `:90-91` (la línea con fechas de nueve cifras, que la forma no casa) | `grep -i -F` de las 35 formas, escritas en UTF-8 y con escapes `\x`, sobre `internal/evals/*_test.go` |

Observaciones en la plataforma (lectura con `gh`, sin escribir nada), que respaldan supuestos sin ser comprobación local:

- **O1** · Las dos ejecuciones sobre `6ab3add`: 36672667544 (apertura, creada 05:16:40) y 36672671529 (etiqueta, 05:16:44).
  La de la etiqueta arrancó sus trabajos de `evals` a las 05:16:47; la de apertura terminó `cambios` a las 05:16:47, y
  sus trabajos esperaron por la concurrencia hasta las 05:24:06 y 05:29:14 y volvieron a medir. `gh run view --json jobs`
  da de cada trabajo `name`, `status`, `conclusion` y `steps` con `name`, `status` y `conclusion`; un trabajo saltado
  por su `if` sale con `conclusion: skipped` y sin pasos, y un paso saltado, con `conclusion: skipped`.
- **O2** · En la PR #89 (solo `docs/`), el trabajo de matriz saltado por su `if` aparece en `gh pr checks` como
  `evals (${{ matrix.skill }})`, `bucket: skipping`, `state: SKIPPED`.

Supuestos no verificados:

- **S1** · Windows PowerShell 5.1 no tiene `&&` (los operadores de cadena de tuberías llegan en PowerShell 7) y en las dos
  versiones `$LASTEXITCODE` es el código de salida de la última orden nativa, de modo que
  `<lectura>; if ($LASTEXITCODE -eq 0) { <comprobación> }` no ejecuta la comprobación si la lectura termina con otro
  código; en 5.1, `$?` tras una orden nativa puede depender de lo que escribe en la salida de error si se redirige, y
  una orden nativa que termina con otro código no lanza una excepción que atrape `try`. No se puede comprobar aquí
  (V22): ningún job ni test ejecuta PowerShell (fuera de alcance, «medir en Windows»); lo que se comprueba en `make ci`
  es la forma (FR-095).
- **S2** · Los `databaseId` de las ejecuciones de un flujo crecen con su creación (O1: 36672667544 < 36672671529, 4 s
  después). El diseño de D15 ordena por él.
- **S3** · Un trabajo de matriz saltado por el `if` del trabajo deja una comprobación con otro nombre que
  `evals (<skill>)` y en `skipping` (O2), que el cierre no cuenta como roja ni como informe (V9).
- **S4** · `gh run view <id> --json jobs` lista cada trabajo de la ejecución que ya se creó con su estado y sus pasos
  (O1); un trabajo que aún no se ha creado, o que no ha terminado, no está o no está `completed`. D15 solo decide con
  el trabajo `tanda` ya `completed`, así que no depende de cómo se listen los que aún no lo están.
- **S5** · `claude -p --output-format stream-json` emite un `result` al final de cada turno, y el aviso de una tarea en
  segundo plano (V4) abre un turno más con su propio `result` (V3: el juzgado fue la réplica). El primer `result` es el
  del turno que abre la pregunta.
- **S6** · `gh` está en el runner `ubuntu-24.04` con `GH_TOKEN` del `GITHUB_TOKEN` (lo usa hoy el paso de `cambios`), y
  un `GITHUB_TOKEN` con `actions: read` lee las ejecuciones y los trabajos del flujo.
- **S7** · El efecto de v0.1.4 en las respuestas —la activación en la 04, las dos clases— no se mide sin modelo: lo mide
  el job de cierre (SC-001), y el escenario del sondeo del quickstart (§8) lo anticipa para quien lo lance.
- **S8** · La lista mide por las formas: lo dicho con otras palabras (FR-037), y cualquier paráfrasis de la respuesta de
  FR-022 que caiga en una forma de la clase B, se verá en el informe, que publica cada respuesta.
- **S9** · El `if` de un trabajo sin función de estado lleva un `success()` implícito que, con una dependencia
  transitiva saltada (`cambios`, en la etiqueta y en el despacho), podría saltar el trabajo. `evals` lleva
  `!cancelled() && needs.tanda.outputs.medir == 'si'`, que vale lo mismo sea o no transitivo: sin `medir=si` (una `tanda`
  saltada, fallida o que decidió no medir) se salta; con él, corre. Es el motivo por el que el `if` de hoy lleva
  `!cancelled()` (`specs/006-h5-skill-boe-legislacion/contracts/job-de-evals.md:203-205`).

## Causa de raíz

### La activación (FR-001, FR-002)

**Medida.** En la línea de base, 2 de las 51 respuestas del modelo que decide en evals que activan la skill no la
activan, las dos de la 04 (V2); en los cierres de H7.2 y H7.3, con Sonnet 5, 0 de 51. El sondeo del ADR 0031 lo
reproduce con las mismas dos (de 102 que deben activarla, entre los dos modelos), con «Contesto de memoria y no he
consultado el texto consolidado en el BOE». Las 11 y 12 (no activación) pasan 6 de 6 con los dos modelos.

**Por qué.** Tres hechos de las sesiones:

1. El modelo conoce la skill y la trata como una comprobación opcional de lo que ya sabe: «puedo comprobarlo con la
   skill de legislación» (04-03), «conviene comprobar la redacción vigente» (04-01). La `description` de v0.1.3 la
   presenta así: «Úsala cuando se pregunte qué dice un artículo, una ley o un real decreto…». No dice que sea la forma
   de responder, ni que lo respondido sin leer no tiene cita.
2. La pregunta de la 04 no pide «qué dice» el artículo: pide un dato («¿En cuántos años prescribe…?»). La `description`
   enumera cuándo usarla por lo que se pide (el texto de un artículo) y no por lo que la respuesta necesita (lo que
   fija una norma: un plazo, un requisito). Las otras nueve positivas que deciden preguntan «qué dice» o «qué
   exige», o por una materia, y se activan 3 de 3.
3. Solo pasa con el modelo nuevo y con un artículo muy conocido (art. 66 LGT, cuatro años): el modelo cree saber la
   respuesta y la da con seguridad, y dice que «respondo de memoria». La decisión de activar se toma con la
   `description`, antes de cargar el cuerpo: lo que diga el cuerpo no cambia esa decisión.

**Qué cambia** (contrato de la skill, C1): la `description` dice que la skill se usa **siempre que la respuesta dependa
de lo que dice una norma**, nombra lo que se pregunta además del texto («qué plazo, requisito o procedimiento fija»,
«dónde se regula una materia», el ámbito de FR-001) y da la razón en lugar de una prohibición: «también cuando creas
conocer la respuesta: sin leer la norma con el binario, la respuesta no tiene cita». No usa «memoria» (está en la
lista, `memoria de consultas`) ni nombra evals ni modelos. No se toca la 04 (FR-002). Contra la sobreactivación (FR-003):
el ámbito sigue siendo lo que dice una norma consolidada del BOE; las preguntas de `legal-core` (qué comunidad,
provincia y boletines corresponden a un municipio; una receta) y las 11 y 12 (programación; reescribir un acuerdo
entre amigos) no dependen de lo que dice ninguna norma, y lo comprueban sus evals (D9).

### La clase A: la comprobación contada con otras palabras (FR-010, FR-020, FR-021)

**Medida** (V20, con la lista nueva): 7 de 51 en `196ee05` (la lista de hoy, 1), 11 en `6ab3add`, 4 en `eb6b4c8`, y
**27 de 51** en la línea de base, más las 2 de «Respondo de memoria» de la 04: 29 de 51 marcadas. Las formas que
dominan en la línea de base van casi todas en el párrafo final de vigencia: «la comprobación de redacción no detectó
cambios» (08-01, 16-01, 16-02, 17-01…), «ni cambios de redacción desde una lectura anterior» (15-02, 15-03), «no ha
cambiado respecto a lecturas anteriores» (09-02), «La comprobación de cambios de redacción no encontró ninguno» (02-01,
05-03, 10-02).

**Dónde lo enseña `SKILL.md` v0.1.3, línea a línea** (fuera del código y de la región generada; V19):

| Líneas | Texto de v0.1.3 | Lo que repite la respuesta |
|---|---|---|
| 65-66 (paso 3) | «La orden que lee un bloque **comprueba también**, detrás de la lectura, **si su redacción ha cambiado desde una lectura anterior**» | «La comprobación de redacción no detecta cambios desde una lectura anterior» (07-03), «no ha cambiado desde una lectura anterior» (14-03) |
| 72-77 (paso 3) | «El segundo es el de **la comprobación**… «Redacción modificada **desde una lectura anterior**»; si **la comprobación** termina…; **La comprobación** va siempre así» | «La comprobación de cambios de redacción no encontró ninguno» (02-01, 05-03, 10-02), «la comprobación no detectó cambios» (01-03, 03-01) |
| 79-81 (paso 3) | «con **la comprobación** de esos mismos bloques detrás» | lo mismo |
| 82-83 (paso 3) | «apagaría el aviso de que **su redacción ha cambiado desde una lectura anterior** (más en «Redacción modificada **desde una lectura anterior**»)» | «ni cambios de redacción desde una lectura anterior» (15-02, 15-03, 01-02) |
| 102 (paso 4) | «**Comprueba si lo leído basta**» | «tengo suficiente para responder…; no necesito el artículo 60» (14-02 de `196ee05`), «I have enough to respond now» (14-01) |
| 116-119 (paso 5) | «y, **si la redacción cambió**, la línea…» | el caso contrario, dicho en la respuesta: «no ha cambiado» |
| 120-122 (paso 5) | «Lo único que la respuesta dice **de una lectura anterior** es la línea… de un bloque **cuya redacción ha cambiado**» | «respecto a una consulta anterior» (01-02, 05-02 de `196ee05`) |
| 177 (título) | «## Redacción modificada **desde una lectura anterior**» | el nombre de la sección, que remiten 74, 83 y 270 |
| 185 | «dice que la redacción de ese bloque **ha cambiado desde la lectura anterior**» | «no ha cambiado respecto a lecturas anteriores» (09-02) |
| 198 | «Lo único que la respuesta dice **de una lectura anterior** es esa línea» | como 120-122 |
| 260-262 (regla 7) | «**Si la comprobación de la redacción** no termina con `0`… **sin afirmar que la redacción ha cambiado ni que no**» | «La comprobación de la redacción no detectó cambios…» (13-03 de la base) |
| 268-270 (regla 7) | «**sin comprobación**… solo va cuando **la comprobación** termina con `0`… («Redacción modificada **desde una lectura anterior**»)» | como 72-77 |

**Conclusión.** Como en H7.2 y H7.3, el modelo repite la palabra de la prosa: «desde una lectura anterior», «la
comprobación», «si su redacción ha cambiado» y «lo leído basta» son de `SKILL.md`, y la respuesta los usa para decir
el caso que la skill no enseña a decir (que no cambió, que no hace falta leer más). La región generada lleva también
«desde la lectura anterior» (la tabla de `--describe`), que el hito no puede cambiar (el binario está fuera de
alcance); queda como fuente residual, a la vista en cada informe. **Qué cambia** (contrato de la skill, C2-C10): la
prosa nombra la orden por su nombre, en código (`kitlegal graph check`), y no «la comprobación»; no dice «lectura
anterior» ni «consulta anterior» (la sección pasa a «Redacción modificada»); el paso 4 no enmarca la decisión como
«si basta» y dice que lo que se decide en él no va en la respuesta, porque quien pregunta no ve los pasos; el paso 5
añade «lo que vas a hacer» a lo que no se cuenta. Lo que era una prohibición («lo único que la respuesta dice…») pasa a
decir qué lleva la respuesta, sin enumerar lo que no (lección de H7.3: la enumeración enseña cada palabra).

### La clase B: una redacción que ninguna orden devolvió (FR-011, FR-020, FR-022)

**Medida.** 5 de las 9 sesiones de la 19 en las tres mediciones del cierre de H7.3 (19-01 de `6ab3add`; 19-01 y 19-02 de
`eb6b4c8`; 19-01 y 19-02 de `196ee05`), y ya en H7.1 (19-01) y H7.2 (19-01 y 19-02). Con el modelo que decide ahora, 0
de 3: la 19-02 dice «No puedo decirte qué contenía la redacción anterior.».

**Por qué.** Tres hechos:

1. La pregunta de la 19 («Hace tiempo te pregunté… ¿Qué dice ahora?») invita a comparar, y la línea
   `⚠ REDACCIÓN MODIFICADA:` confirma que hubo un cambio sin decir cuál. `SKILL.md` solo lo prohíbe («Lo único que la
   respuesta dice de una lectura anterior es esa línea»), sin la razón y sin decir qué responder si se pregunta qué
   cambió: el modelo lo rellena con lo que cree saber.
2. Se ancla en lo que sí devuelve la lectura: la `norma_modificadora` y la `fecha_vigencia`. «El artículo cambió porque
   el Real Decreto-ley 3/2020… modificó el art. 118… ya no exige tres informes» (19-01 de `196ee05`); «Esta redacción
   procede de la modificación del Real Decreto-ley 3/2020… es distinta de la que se consultó antes (vigente hasta el 9
   de marzo de 2018)» (19-02). La skill no separa lo que puede decir (norma modificadora y fecha de la redacción
   vigente) de lo que no (qué decía la anterior).
3. Dos lo hacen con datos falsos contra la grabada: la prohibición no basta, y la redacción superada es contenido
   legal sin fuente aunque acierte (principio II).

**Qué cambia** (contrato de la skill, C8): «Redacción modificada» dice por qué la respuesta no puede describir la
redacción superada (no la ha leído: `boe articulo` da la vigente y `graph check`, dos fechas), qué sí dice (el texto
vigente con su cita, `norma_modificadora` y `fecha_vigencia`) y qué responde si se pregunta qué cambió (que cita la
redacción vigente y que la anterior no la ha leído), que es lo que ya hace la 19-02 de la línea de base. Y cada línea
`⚠ REDACCIÓN MODIFICADA:` lleva la cita de su bloque (D4), para que la línea diga de qué precepto es.

## Calibrado (FR-032, FR-093)

La lista de D6, aplicada con las formas fijas quitadas (D7) a las 93 respuestas de cada informe (V20). Columnas: `maq`
(maquinaria), `otra` (otra conversación), `anu` (anuncio), `red` (redacción no leída), `alg` (alguna).

| Informe | Marcadas | Reparto por eval |
|---|---|---|
| H7.1 | **36 de 93** (las 35 de hoy y la 19-01) | 02 {maq 1, alg 1}; 03 {maq 3, anu 3, alg 3}; 04 {maq 3, anu 3, alg 3}; 05 {maq 3, anu 1, alg 3}; 06 {maq 3, anu 2, alg 3}; 07 {maq 3, anu 2, alg 3}; 08 {maq 2, anu 1, alg 2}; 09 {maq 2, anu 1, alg 2}; 13 {maq 3, anu 2, alg 3}; 14 {maq 3, anu 2, alg 3}; 15 {maq 3, anu 3, alg 3}; 16 {maq 2, anu 2, alg 2}; 17 {maq 3, anu 2, alg 3}; 19 {otra 1, anu 1, red 1, alg 2} |
| H7.2 | **11 de 93** (las 10 de hoy y la 19-01) | 03 {maq 1, anu 1, alg 1}; 06 {maq 1, anu 1, alg 1}; 13 {maq 2, anu 2, alg 2}; 14 {maq 2, anu 2, alg 2}; 15 {maq 3, anu 3, alg 3}; 19 {maq 1, anu 1, red 2, alg 2} |
| H7.3 (`196ee05`) | **9 de 93** | 01 {anu 1, alg 1}; 05 {anu 1, alg 1}; 13 {anu 2, alg 2}; 14 {maq 1, anu 3, alg 3}; 19 {anu 1, red 2, alg 2} |

Las respuestas de la clase B de los tres (19-01 de H7.1; 19-01 y 19-02 de H7.2; 19-01 y 19-02 de `196ee05`) las marca
`redaccion_no_leida`, y ninguna de la clase A la marca `redaccion_no_leida`. El `anu 1` de la 19 es `se consultó antes`
fuera de la línea de la redacción (19-01 de H7.1, 19-02 de H7.2 y 19-02 de `196ee05`), en respuestas que ya marca otra
familia; la línea misma, que en los tres informes va sin cita, la quita su forma fija (D7). Ninguna del modelo
informativo se marca. Fuera del calibrado exigible, y sin cambio con `se consultó antes`: `6ab3add` da 14 de 93 (las 11
de la clase A de la bitácora, la 19-01 de la B y la 19-02 y 19-03, que dicen «el cambio de redacción», clase A que la
bitácora no contó); `eb6b4c8`, 6 (las 4 y las 2 de la bitácora); la línea de base, 29 de 51 (27 de la clase A y las 2 de
la 04), frente a las 28 de la lectura del hito.

**La respuesta de FR-022 no se marca.** La 19-02 de la línea de base («No puedo decirte qué contenía la redacción
anterior.») y la frase que enseña v0.1.4 no llevan ninguna forma: por eso la clase B no gana «redacción anterior»,
«versión anterior», «versiones anteriores», «versión previa» ni «redacción original de», que marcarían a quien dice
honradamente que no la ha leído; cada respuesta de la B que las llevaba la marca otra forma (tabla de D6).

## Decisiones

### D1 · La activación va a la `description` (FR-001, FR-002)

Decisión: la `description` de C1 («Causa de raíz», activación). Alternativas rechazadas: (a) retocar la 04 o su
pregunta (fuera de alcance: se corrige la clase de defecto); (b) una regla en el cuerpo («no respondas de memoria»):
el cuerpo se carga solo si la skill ya se activó, así que no cambia la decisión de activar; (c) una prohibición en la
`description` («no respondas nunca sin la skill»): la razón («sin leer la norma… no tiene cita») es la que el modelo
contrapone a «puedo comprobarlo», y una prohibición no dice por qué; (d) enumerar más abreviaturas o normas: la 04 ya
nombra la Ley General Tributaria y el artículo.

### D2 · La clase A: la prosa sin el vocabulario, y la lista en `anuncio` (FR-010, FR-021, FR-030)

Decisión: C2-C10 quitan de la prosa «lectura anterior», «consulta anterior», «la comprobación» y «si lo leído basta»
(«Causa de raíz»), y las formas de la clase A entran en la familia `anuncio`, que ya es la del «estado de lo
comprobado» (`internal/evals/prohibidas.go:48-52`). Alternativas: (a) una familia `comprobacion` nueva: separa dos
grupos que el contrato de H7.3 ya une, sin que ningún umbral los distinga (solo la B tiene umbral propio), y añade un
campo, una clave del esquema y una columna del calibrado sin requisito que la pida (proporcionalidad); (b) solo lista,
sin tocar la prosa: el ruido sale de la prosa (H7.2, H7.3), y la lista sola dejaría el umbral en rojo (29 de 51).

### D3 · La clase B: por qué, qué sí y qué responder (FR-011, FR-022)

Decisión: C8 («Causa de raíz», clase B). La frase de FR-022 la enseña la prosa y la dice la respuesta con sus palabras;
ninguna forma de la lista la marca (Calibrado). Alternativa rechazada: una prohibición más, que FR-020 descarta
expresamente.

### D4 · La línea `⚠ REDACCIÓN MODIFICADA:` con la cita de su bloque (FR-023)

Decisión: `⚠ REDACCIÓN MODIFICADA: <forma legible> [<identificador>, bloque <id>]: la redacción con fecha de vigencia
AAAAMMDD, la que se consultó antes, ha sido sustituida por la de AAAAMMDD, que es la que se cita.`, una por bloque,
también con uno solo (supuesto del spec). La cita va detrás de la etiqueta porque es lo primero que lee quien pregunta
(«de qué precepto es»), y con la forma de «Cómo se cita», que `ExtraerCitas` ya reconoce. Alternativas: (a) la cita al
final de la línea: el texto fijo de H7.3 se lee sin saber de qué bloque habla hasta el final; (b) solo el id del bloque
(«bloque a1-30»): no es la forma de cita de la skill ni se lee como un precepto; (c) una cabecera con el precepto y la
línea debajo: dos líneas, y la eval no podría atar la línea a su bloque.

### D5 · La forma para PowerShell (FR-024, FR-095)

Decisión: `<lectura> --json; if ($LASTEXITCODE -eq 0) { <comprobación> --json }`, en un bloque `powershell` justo
detrás del bloque `bash` de cada una de las dos órdenes (la de `articulo`, con su ejemplo, y la de `articulos`, con sus
marcadores, que pasa de estar en línea a un bloque). Alternativas: (a) `&&`: solo PowerShell 7 (S1); (b) `if ($?)`:
`$LASTEXITCODE` es el código de la orden nativa y no depende de nada más, mientras que en 5.1 `$?` puede verse afectado
por lo que la orden escribe en la salida de error cuando se redirige (supuesto, como todo lo de PowerShell, S1); (c)
`-and` o `try/catch`: una orden nativa que termina con otro código no lanza excepción (S1); (d) dos órdenes separadas:
el agente las partiría, y FR-025 pide la comprobación en la misma orden que la lectura.

### D6 · La lista: la clase B en una familia nueva, `redaccion_no_leida` (FR-030, FR-032, FR-036)

Decisión: el fichero gana `redaccion_no_leida` (la única de la clase B; su clave es la del umbral que alimenta) y
`anuncio` gana las formas de la clase A (D2); maquinaria, otra conversación y anuncio siguen siendo de la clase A. Las
formas:

- `anuncio`, además de las 14 de H7.3: `lectura anterior`, `lecturas anteriores`, `consulta anterior`, `consultas
  anteriores`, `cambio de redacción`, `cambios de redacción`, `cambio en la redacción`, `cambios en la redacción`,
  `la comprobación de redacción`, `la comprobación de la redacción`, `la comprobación de cambios`, `la comprobación
  no`, `redacción posterior`, `posterior a la consultada`, `haya cambiado desde` (el estado de la comprobación o de una
  lectura anterior); `se consultó antes` (la lectura anterior con las palabras de la línea `⚠ REDACCIÓN MODIFICADA:`,
  fuera de ella: «es distinta de la que se consultó antes», 19-02 de `196ee05`; «lo que se consultó antes era la
  versión previa», 19-01 de H7.1; FR-031); `tengo suficiente`, `no necesito`, `i have enough`, `i don't need`, `i don’t need`, `respond
  now`, `respondo de memoria`, `contesto de memoria`, `extracto fiel` (lo que el agente tiene, necesita, va a hacer o
  anuncia).
- `redaccion_no_leida`: `ya no exige`, `ya no se exige`, `se eliminó`, `vigente hasta`, `aplicable hasta`, `el cambio
  relevante`, `el cambio más relevante`, `artículo cambió`, `no ha cambiado por`, `esa versión ya no` (qué decía la
  redacción superada, hasta cuándo rigió, o qué cambió respecto de ella).

Alternativas rechazadas: (a) anidar las familias por clase (`clase_a: {…}`, `clase_b: {…}`): cambia la forma de un
fichero y de un tipo que usan decenas de tests para ganar lo que ya da una clave; (b) las formas que nombran la
redacción anterior («redacción anterior», «versión anterior», «redacción original de»): marcarían la respuesta de
FR-022 (Calibrado); (c) las frases enteras de la bitácora: FR-030 pide formas; (d) «aquí va» (la 01-03): marca la 12-01
de H7.1 y de H7.2, fuera del calibrado; «extracto fiel» marca la misma frase sin eso.

### D7 · Las formas fijas se quitan antes de buscar, declaradas en la lista (FR-031, FR-033)

Decisión: `ExtraerExpresionesProhibidas` quita del texto, antes de buscar y en este orden, (1) cada forma de
`formas_fijas` de la lista —un texto con dos marcadores, `<cita>` (lo que va de la etiqueta al corchete de cierre de
una cita, en la misma línea) y `<fecha>` (ocho cifras)— y (2) la marca, la etiqueta y los dos puntos de cada aviso de
vigencia (`boe.EtiquetasDeAviso`, las mismas de `ExtraerAvisos`); cada tramo quitado se cambia por un salto de línea,
para que las palabras de sus lados no formen una expresión. `<cita>` es opcional, con lo que la sigue en la forma hasta
el blanco siguiente (en la línea, `(?:<cita>:[ \t]+)?`): la forma quita la línea de v0.1.4, con su cita, y la de
v0.1.1 a v0.1.3, sin ella, que es la que llevan las respuestas de los tres informes del calibrado (FR-032). La lista no
mide la cita de la línea (FR-023): la juzga la eval 20 (FR-052, D10). La lista de `boe-legislacion` declara dos formas:
la línea `⚠ REDACCIÓN MODIFICADA:` de D4 y la frase de la regla 7 sin su punto final (196ee05 14-01 la escribe seguida
de un paréntesis). Alternativas: (a) constantes en Go con el texto de la skill: pondría texto de `boe-legislacion` en el
paquete común de las evals; (b) expresiones regulares en el YAML: ilegibles y fáciles de romper; (c) leer las formas de
los bloques `text` de `SKILL.md`: acopla el juicio a la maquetación de la skill; (d) quitar solo la etiqueta de la línea
de redacción: dejaría «la que se consultó antes», que `anuncio` marca (`se consultó antes`, D6), en cada respuesta que
traslada un cambio (FR-031); (e) `<cita>` obligatoria: la línea sin cita de los informes del calibrado se buscaría
entera y `se consultó antes` marcaría una respuesta más de la 19 en cada uno, 37, 12 y 10 (FR-032, «Calibrado»); (f) la
línea sin cita como una tercera forma fija: la skill ya no la enseña, así que no quitaría nada de la respuesta de
`expresiones-de-la-skill`, que exige que cada forma quite algo (D8). El control de que la lista y la skill dicen lo
mismo es `expresiones-de-la-skill` (D8): una respuesta hecha de los bloques `text` de `SKILL.md` con datos de ejemplo no
se marca, y cada forma fija de la lista quita algo de ella.

### D8 · Las comprobaciones de `TestEvalsDelRepositorio` (FR-032, FR-033, FR-021, FR-093)

Decisión: `expresiones-calibradas` pasa a tres informes (el de H7.3 se añade) y a cinco columnas (las cuatro familias y
`alguna`), con la tabla de «Calibrado»; `listaDelRepositorio` exige también `redaccion_no_leida` y `formas_fijas` no
vacías; `expresiones-en-los-bloques` no cambia (ya lee los bloques de toda eval y de todo grafo previo, también los de la
nueva); `expresiones-de-la-skill` pasa a componer una respuesta con la forma escrita de cada aviso y cada bloque `text`
de `SKILL.md` con sus marcadores sustituidos por datos de ejemplo (en cada bloque, el primer `AAAAMMDD` por
`20180309` y los demás por `20200206`; `<forma legible> [<identificador>, bloque <id>]` por `art. 118 de la Ley 9/2017
[BOE-A-2017-12902, bloque a1-30]`; el final de una cita, `<identificador>, bloque <id>]`, por
`BOE-A-2015-10565, bloque a21]`), y exige que no se
marque y que cada forma fija de la lista quite algo de ella; `prosa-de-la-skill` no cambia (aplica la lista nueva). Y
una subprueba nueva, `ordenes-para-powershell` (D5). Alternativa rechazada: un test aparte por comprobación: las cuatro
ya viven en `TestEvalsDelRepositorio`, que es el control que nombra FR-093.

### D9 · El formato de eval: `no_se_activan` y `redacciones_modificadas` (FR-003, FR-004, FR-053)

Decisión: dos claves opcionales. `no_se_activan`, lista de nombres de skill que la sesión no puede activar, en cualquier
eval (también de no activación); una activada es un motivo que la nombra y la sesión no pasa. `redacciones_modificadas`,
solo en una eval que activa la skill: lista de `{norma, bloque, fecha_vigencia, fecha_vigencia_reciente}` (los nombres
de los campos de `graph check`), cada una una línea `⚠ REDACCIÓN MODIFICADA:` que la respuesta tiene que llevar. Las
tres de `legal-core` declaran `no_se_activan: [boe-legislacion]`. Alternativas: (a) una regla del conjunto de `legal-core`
o una opción del job: el juicio es de la sesión, y una eval es donde se declara lo esperado; (b) llamarla `sin_activar`:
es el nombre del umbral; (c) fechas en `hallazgos`: cambiaría el significado de una clave que usa la 19 (FR-053: las de
hoy conservan su juicio). Lo que cambia en `schemas/eval.yaml.json` va en una tarea `[datos]`.

### D10 · El juicio de una línea esperada (FR-052, FR-053)

Decisión: `ExtraerRedaccionesModificadas` recorre la respuesta línea a línea; en cada línea con la forma de la
etiqueta de `version-obsoleta` (las expresiones de `ExtraerHallazgos`) toma la primera cita (`ExtraerCitas`) y las dos
primeras fechas de ocho cifras, en su orden. Una línea esperada está si alguna línea da su norma, su bloque y sus dos
fechas en ese orden. Así, una sola línea, dos sin su cita, o las fechas de otro bloque, dejan una ausente. Cada una se
publica con el texto `<norma> <bloque> <fecha_vigencia> <fecha_vigencia_reciente>`; la ausente es un motivo
(«redacción modificada ausente: …») y la sesión no pasa; las formas exigidas de la serie ganan
`⚠ REDACCIÓN MODIFICADA: <texto>` por cada una. Alternativa rechazada: exigir la frase entera de D4 con sus datos:
la eval mediría la redacción de la línea, que ya mide la lista, y no de qué bloque es.

### D11 · La eval 20 y su grafo previo (FR-050, FR-051, FR-054, FR-055)

Decisión: `evals/boe-legislacion/20-lcsp-dos-bloques-redaccion-cambiada.yaml`, informativa, con la pregunta de FR-050,
el grafo previo `lcsp-a1-30-y-da-3-redaccion-original` (dos comandos de bloque), los comandos `a1-30`, `da-3` y `graph
check` con la norma, `graph show` prohibido, las dos citas, `hallazgos: [version-obsoleta]` y las dos
`redacciones_modificadas`. La carpeta de derivadas lleva las dos: la de `a1-30` (la misma derivación que la de la 19) y
la de `da-3`, con 20180309, las dos escritas por `TestGrabacionesDerivadas -actualizar-derivadas` y comprobadas por él
(dos entradas más en `derivadasDelGrafoPrevio`). `compruebaLaSesion` exige además que cada redacción esperada sea un
`version-obsoleta` de su bloque con sus dos fechas: la eval no puede pedir unas fechas que la grabación no da. No hace
falta ninguna grabación nueva (V15): `grabar_datos` no tiene nada que grabar. Alternativas: (a) `grafo_previo` con
varias carpetas: cambia el formato y la preparación para ahorrar una copia de 5,6 KB producida por código; (b) meter
`da-3` en la carpeta de la 19: su nombre dejaría de decir lo que tiene, y renombrarla toca la 19 (fuera de alcance);
(c) no declarar `hallazgos`: `compruebaLaSesion` exige que coincidan con las clases de la sesión (V16). El `timeout-minutes`
pasa a 122 (D17).

### D12 · La respuesta a la pregunta es el primer `result` (FR-060, FR-062)

Decisión: `LeerSesion` toma la respuesta del **primer** mensaje `result` del transcript (el del turno que abre la
pregunta, S5), con las mismas condiciones de hoy (subtype `success`, `is_error` falso); `Fin`, `Terminada`,
`MotivoSinTerminar` y `ErrorDelResultado` siguen saliendo del último mensaje. Lo comparten el job y el sondeo. Si el
agente cerró ese turno sin responder, se juzga eso (supuesto del spec). Alternativas: (a) el último `result` anterior a
un `system/task_notification`: dos reglas para lo mismo, y con dos tareas en segundo plano se juzgaría una réplica; (b)
ignorar los turnos posteriores por su `parent`: el transcript no ata un `result` a su turno de forma que el reader lo
necesite; (c) juzgar todas las respuestas: no es lo que se lee como respuesta (FR-060).

### D13 · Los recuentos cuentan solo las sesiones terminadas (FR-045, FR-061, FR-082)

Decisión: `recontarExpresiones` pasa a contar, por modelo, las sesiones juzgadas, no ilegibles, no sin medir, de series
que pide el plan, de evals que activan la skill **y terminadas** (`SesionTerminada`), y da además, para los umbrales,
las que no activaron la skill y las que llevan una expresión de `redaccion_no_leida`. Lo publicado
(`expresiones_prohibidas_por_modelo`: `modelo`, `con_alguna`, `respuestas`) no cambia de forma; el recuento interno
lleva los dos números más y no se publica (el sondeo no publica `sin_activar` ni recuentos por clase: fuera de alcance).
La sesión sin terminar sigue sin pasar en su serie (H5) y con su motivo.

### D14 · Los umbrales nuevos (FR-040 a FR-048)

Decisión: `umbralesDelInforme` da, si la skill tiene lista, `expresiones_prohibidas:<modelo que decide>`,
`sin_activar:<modelo que decide>` y `redaccion_no_leida:<modelo que decide>` (los tres con `total`, `"<="` y
`decide: true`; 0,05, 0 y 0), después `expresiones_prohibidas:` de cada informativo (`decide: false`) y detrás
`duracion_de_las_sesiones` si hay objetivo. Los motivos, los de hoy (V7): «umbral sin_activar:claude-sonnet-5-5: 1 de 54
(1,9 %), y tiene que ser ≤ 0,0 %». La condición «si la skill tiene lista» es la de hoy para las expresiones y deja
`legal-core` con `[]` sin ningún caso por skill (FR-048). Alternativas: (a) `sin_activar` para toda skill: `legal-core`
tendría umbrales (fuera de alcance); (b) una clave por skill en la definición del job: configuración que ningún
requisito pide; (c) comparar la medida sin total: el contrato del ADR 0029 da la proporción con total, y con umbral 0 es
lo mismo (1 de 54 no se cumple, 0 de 54 sí).

### D15 · Una tanda por commit: un trabajo `tanda` que decide antes de `evals` (FR-070, FR-071, FR-072, FR-100)

**El problema, con la plataforma.** El cierre abre la propuesta de cambio (`opened`) y 4 s después pone la etiqueta
(`labeled`) sobre el mismo commit: dos ejecuciones (O1). Hoy la segunda espera por la concurrencia y vuelve a medir. Un
paso que termine sin medir **dentro** del trabajo `evals` dejaría una comprobación `evals (<skill>)` en `pass` sin
informe, que el cierre podría leer en lugar de la buena (V9). En cambio, el trabajo saltado por su `if` deja
`evals (${{ matrix.skill }})` en `skipping`, que el cierre no lee ni cuenta como roja (S3, O2). Y el `if` de un
trabajo no ve `matrix` (V10): la decisión se toma para las dos skills a la vez, que es lo que hace cada disparo (las dos
se miden juntas: por etiqueta, por despacho o por apertura con cambios).

**Decisión.**

1. Un trabajo nuevo, `tanda`, con `needs: [cambios]` y el `if` que hoy lleva `evals` (despacho, etiqueta `evals` o
   `evals-prueba-de-red`, o apertura con `coincide == 'si'`): solo corre en una ejecución que quiere medir. Sin
   `concurrency` (no queda nunca en espera ni se cancela), con `permissions: {contents: read, actions: read}`, checkout
   del commit evaluado y Go. Su paso `decidir` ejecuta `TestTandaDelCommit` (etiqueta `evals`) y escribe `medir=si` o
   `medir=no` en `GITHUB_OUTPUT`; su último paso, «Esta ejecución mide el commit», con `if: steps.decidir.outputs.medir
   == 'si'`, no hace nada más que existir: es la marca que las demás ejecuciones leen.
2. `evals` pasa a `needs: [tanda]` con `if: ${{ !cancelled() && needs.tanda.outputs.medir == 'si' }}` (S9). Sin
   `medir=si`, el trabajo se salta entero (S3). La `concurrency` por commit y skill con `cancel-in-progress: false` se queda: es la garantía estructural
   de «una a la vez» (SC-010), aunque con la decisión no llegue a esperar nunca.
3. `TestTandaDelCommit` decide con `decidirLaTanda` (Go, puro) sobre lo que devuelve `gh`: las ejecuciones del flujo
   `evals.yml` sobre el commit (`gh run list --workflow evals.yml --commit <sha> --limit 100 --json databaseId,status`)
   y, de cada una **anterior** (`databaseId` menor, S2) que **no ha terminado**, sus trabajos (`gh run view <id> --json
   jobs`). La tanda de esa ejecución: `mide` si su trabajo `tanda` está `completed` con el paso de la marca en
   `success`; `sin decidir` si su `tanda` aún no está o no está `completed`; `no mide` en cualquier otro caso (saltado,
   fallido, marca saltada). **Mide** si ninguna anterior sin terminar `mide`; **no mide** en cuanto una `mide`. Con
   alguna `sin decidir` y ninguna que `mide`, espera 10 s y vuelve a consultar, como mucho 10 min; agotado, mide (la
   concurrencia la pondría detrás, como hoy). Las posteriores no cuentan: son ellas las que deciden respecto a esta.
   Una anterior ya terminada no cuenta: su tanda terminó, y la etiqueta vuelve a medir (FR-071).
4. Lo que el cierre ve: la ejecución que mide deja `cambios` (en `pass`, o saltado en la etiqueta y el despacho),
   `tanda` en `pass` y `evals (<skill>)`; la que no, `tanda` en `pass` y `evals (${{ matrix.skill }})` en `skipping`. Ninguna roja ni ninguna `evals (<skill>)` de más (V9). El cierre
   sigue esperando mientras las de la que mide estén `pending`.

**Por qué espera, y a qué.** La segunda ejecución no espera a la tanda de la primera: espera a que la primera **decida**
(su `tanda` completa, segundos). Sin eso, en el caso del cierre la segunda consultaría cuando la primera aún está en
`cambios` y no sabría si va a medir (una apertura sin cambios que midan no mide). Es la única espera, y acaba en
segundos.

**Alternativas rechazadas.** (a) La de H7.3, esperar por la concurrencia y volver a medir: es lo que FR-070 sustituye.
(b) `cancel-in-progress: true`: cancela la tanda que mide, y una cancelada es roja para el cierre. (c) Concurrencia de
nivel de flujo y saltarse al arrancar si otra midió mientras esperaba: la segunda espera (FR-070 elige no esperar), y un
tercer disparo cancela la que espera. (d) Un paso que termine sin medir dentro de `evals`: deja una `evals (<skill>)`
sin informe (V9). (e) Dar por medidora a toda anterior sin terminar, sin esperar su decisión: una apertura sin cambios
seguida de la etiqueta no mediría nunca. (f) Codificar la intención en `run-name` y recalcular `coincide` en la
posterior: repite la lógica de `cambios` y no ve la decisión real de la anterior. (g) La decisión en `bash` con `jq` en
el propio YAML: no se puede probar en `make ci` sin `jq`, que no es del stack. (h) Una referencia de git como cerrojo:
escribe en el repositorio (`contents: write`).

### D16 · `gh` desde Go con órdenes constantes (D15; V12)

Decisión: `TestTandaDelCommit` recibe el commit, su ejecución y la ruta de `GITHUB_OUTPUT` por banderas (`-commit`, que
ya existe; `-ejecucion`; `-salida`), y ejecuta `gh` con `exec.CommandContext(ctx, "sh", "-c", orden)`, con `orden`
constante (`exec gh run list --workflow evals.yml --commit "$COMMIT_EVALUADO" …` y `exec gh run view
"$EJECUCION_ANTERIOR" --json jobs`) y los valores en el entorno de la orden, entrecomillados: ningún dato entra en la
orden como texto (G204 no salta, V12; G702 no ve fuente: banderas y `os.Environ()`). El paquete no importa `net/http`
(R2). Alternativas: (a) `gh` con argumentos variables: G204, y un `//nolint` que la constitución solo admite justificado
cuando no hay otra forma; (b) un guion en `scripts/`: otro fichero para dos órdenes, y FR-072 sitúa el cambio en
`evals.yml` e `internal/evals`; (c) la API de GraphQL con `-F campo=@fichero`: su esquema no se puede comprobar en local.

### D17 · `timeout-minutes: 122` (FR-054; H7.3 FR 035)

Decisión: 122 min cubren los 7 285 s del peor caso con 97 sesiones (V13); `TestDefinicionDelJob` ya lo exige
(`fallosDelTope`). El cierre espera hasta 3 h (`scripts/workflow/cierre.sh:52`). El trabajo `tanda` lleva `timeout-minutes:
15` (10 min de espera como mucho y la preparación), y su paso `decidir` ejecuta `go test` con `-timeout 12m`: con los
10 min de `go test` por omisión, el test acabaría en pánico antes de medir tras la espera agotada (supuesto T008). Alternativa: bajar la concurrencia o el tope de sesión: cambia la
duración medida (FR-043) sin requisito.

### D18 · Los errores de uso del sondeo, sin traza (FR-080, FR-081)

Decisión: `comprobarElSondeo` devuelve sus errores de argumentos y de credencial envueltos en un tipo sin exportar, `errorDeUso`;
`TestSondeo`, con uno, escribe su mensaje con un salto de línea final en `uso.txt` del temporal y **termina sin fallar**;
con cualquier otro error falla como hoy. `scripts/evals-sondeo.sh`, tras `go test` y en este orden: si `go test`
falló, imprime el registro y sale con 1 (como hoy); si no y `uso.txt` no está vacío, lo escribe en la salida de error y
sale con 1; si no, imprime `salida.txt` (contracts/sondeo.md §3; `scripts/evals-sondeo.sh:45-57`). Ninguna sesión se abre (la comprobación va antes de construir nada). El código de `make` sigue
siendo 2 (§6 del contrato de H7.3). Alternativas: (a) distinguir el error por el registro de `go test`: frágil; (b) un
código de salida propio de `go test`: `go test` no lo propaga; (c) que el guion compruebe los argumentos: duplicaría la
comprobación en bash.

### D19 · La documentación que el hito deja falsa

`CONTRIBUTING.md` («Job de evals»: la tanda única en lugar de la espera, V25; la respuesta juzgada; los tres umbrales de
la respuesta; la lista con `redaccion_no_leida` y `formas_fijas`; las claves nuevas del formato de eval), los
comentarios de `.github/workflows/evals.yml` y `internal/evals/doc.go`. Y `CHANGELOG.md` (*Unreleased*, FR-027,
FR-101). No se documenta nada más.

### D20 · Aceptación e2e: no aplica

El binario no cambia (fuera de alcance), así que ningún guion `testscript` describe la entrega. La aceptación es el job
de cierre (SC-001) y, en `make ci`, los tests de «Controles de umbral».

### D21 · Datos externos: ninguno

Las respuestas grabadas de `a1-30` y `da-3` están desde H4 (V15); las derivadas las produce código del repositorio. No
hay manifiesto `grabaciones.json` ni test `TestGrabar*` nuevos, ni nada que grabe `grabar_datos`, ni fuente nueva.

### D22 · Reparación del cierre: el caso sin entrada de `kitlegal graph check`, dicho con la razón (FR-010, FR-020, FR-021)

Lo decide `reparar_cierre` con la medición de V28: 15 de 54 con C1-C10, frente al umbral de 2. **Causa.** D2 leyó la
clase A como vocabulario, porque en H7.2 y H7.3 el modelo repetía las palabras de la prosa, y C1-C10 las quitaron. Con
el modelo que decide, eso bajó la clase A de 27 de 51 (línea de base) a 14 de 54, y ahí se quedó: el modelo sigue
contando que `kitlegal graph check` no encontró nada, ahora con las palabras de las dos formas fijas, que FR-025 deja
como están, y de la etiqueta (V28). La prosa de C1-C10 decía qué lleva la respuesta por cada entrada de
`version-obsoleta` y nunca qué pasa sin ninguna, y ponía ese resultado al lado de los avisos de vigencia («Le sirven la
norma, su texto, su cita, sus avisos de vigencia y la línea…»): el modelo lo trató como un dato más de la vigencia del
bloque, y lo dijo donde dice «no trae avisos de vigencia», que sí es derecho (FR-012). Es la misma carencia que C9 le
quitó a la clase B —una prohibición sin la razón ni el caso—, y la clase B da 0 de 54 en la misma medición.

**Decisión** (contracts/skill-boe-legislacion.md C11): la prosa dice el caso sin entrada y su razón, en los tres sitios
que la respuesta recorre —el paso 3 (el sobre vacío no da nada), el paso 5 (tampoco se cuenta lo que una orden no ha
devuelto; el sobre de `kitlegal graph check` no dice nada de la vigencia del bloque) y «Redacción modificada» (una
entrada es un dato de kitlegal sobre otras conversaciones, no de la norma; sin entrada, ni una línea, ni una frase
junto a los avisos, ni una palabra al final)—, más que la línea no se repite con otras palabras (20-01) y que la frase
de la regla 7 no se lleva con `0`, ni afirmada ni negada (08-01, 16-02). Sin ninguna forma de la lista en la prosa
(`prosa-de-la-skill`), sin tocar las formas fijas ni la región generada, y en 298 líneas, que es el tope efectivo
(`dos-inicios` añade una línea al mutante). Alternativas rechazadas: (a) más formas en la lista («no se ha detectado»,
«no consta que»): no arreglan la skill, y el umbral ya las cuenta con lo que hay (FR-047); (b) cambiar las formas
fijas, de donde salen ahora las palabras: FR-025 y «Fuera de alcance»; (c) cambiar la salida de `graph check` sin
entradas: «Fuera de alcance» del hito; (d) pedir que la respuesta no diga que no hay avisos de vigencia, para quitarle
el párrafo en que lo cuenta: FR-012 lo tiene por derecho, y 39 respuestas limpias lo dicen sin contar nada más; (e) una
prohibición más sin la razón: FR-020 la descarta, y D3 muestra que la razón es lo que funciona. Lo mide el job de
cierre, sobre la cabeza que lleve C11 (SC-001); antes, el sondeo local con el modelo que decide sobre las once evals
de V28 da 0 de 33 y las once series 3 de 3 (V29).

### D23 · Segunda reparación del cierre: el fin de una redacción no lo trae ningún sobre (FR-011, FR-020, FR-022)

Lo decide `reparar_cierre` con la medición de V30: con C1-C11, la clase A queda en 0 de 54 y la B da 1 de 54, la 18-01,
frente al umbral 0 de `redaccion_no_leida` (FR-042). **Causa.** La respuesta presenta la última redacción del art. 42
de la Ley 30/1992, derogada, como «vigente hasta la derogación (vigencia desde 14/04/1999)». Es lo que la prosa
enseñaba, conciliado con la regla 3: «Redacción modificada» decía que de una redacción se dice qué norma la dio y
«desde cuándo está vigente», y la regla 3, que no se presente como vigente el texto de una norma derogada; con una norma
derogada, «vigente» y «derogada» se juntan en «vigente hasta la derogación», que convierte la fecha de inicio de la
redacción en un intervalo cuyo fin ningún sobre da. `kitlegal boe articulo` trae `fecha_vigencia` y
`norma_modificadora`; `kitlegal boe metadatos`, `estatus_derogacion` sin fecha; y el aviso `derogada` dice que lo está,
sin cuándo (V30). La prosa nunca decía que el sobre trae el principio de una redacción y no su fin, ni qué se dice de la
redacción de una norma derogada. Es la clase B por la definición de la lista («hasta cuándo rigió») y por FR-011, que
enumera lo que sí se puede decir: la norma modificadora y la fecha de vigencia, no el fin. La 18-02 del informe de H7.1
ya razonaba igual sin caer en una forma («la última redacción vigente (fecha de vigencia 1999-04-14) antes de la
derogación»): el hueco es de la skill, no de esta medición.

**Decisión** (contracts/skill-boe-legislacion.md C12): la prosa lo dice con la razón, como C9 y C11, en los dos sitios
que enseñan qué se dice de una redacción: el paso 5 (de la vigencia del bloque, sus avisos y, de la redacción leída,
qué norma la dio y desde cuándo rige; hasta cuándo, nunca, tampoco en una norma derogada, cuyo aviso no lleva fecha,
porque ningún sobre trae el fin de una redacción y darlo sería texto legal sin fuente) y «Redacción modificada» («desde
cuándo rige; hasta cuándo, nunca»). Sin ninguna forma de la lista en la prosa (`prosa-de-la-skill`), sin tocar las
formas fijas, la regla 3, la forma de los avisos ni la región generada, y en 298 líneas: las dos líneas que gana el
paso 5 las dan dos líneas en blanco del paso 3 entre un bloque de código y la viñeta siguiente, como ya iba la de
`articulos`. Alternativas rechazadas: (a) quitar `vigente hasta` de la lista o acotarla («vigente hasta el»): FR-047 y
el ADR 0029 (la lista no pierde expresiones ni se recorta para cumplir un umbral), y el calibrado la exige, porque es
la única forma de la clase B que marca la 19-02 de `196ee05` (FR-032; `expresiones-calibradas` fallaría); (b) decirlo
en la regla 3 o en la forma de los avisos de «Cómo se cita»: FR-025 los deja como en v0.1.3, y el paso 5 es donde se
compone la respuesta; (c) una prohibición más sin la razón: FR-020; (d) darlo por falso positivo y no cambiar nada:
el umbral es 0 y decide (FR-042), y no es falso positivo: el fin de una redacción es un dato que ninguna orden
devolvió, con fecha o sin ella. Lo mide el job de cierre sobre la cabeza que lleve C12 (SC-001).
