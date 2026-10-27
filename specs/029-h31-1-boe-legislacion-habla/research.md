# Research: H31.1 · `boe-legislacion` habla de la norma, no de la consulta, y lee lo que nombra

Decisiones (D), verificaciones hechas en esta sesión (V) y supuestos que no se han podido verificar sin red o sin
persona (S). Nada de lo que aquí se da como hecho se ha medido fuera de esta sesión, salvo lo que está en el
repositorio y se cita con su ruta.

## Decisiones

### D1. La skill de la entrega es la medida con el cuerpo reescrito; ningún texto sale a una referencia

- **Decisión**: `skills/boe-legislacion/` de v0.3.1 = `evidencias/sondeos-20261026/brazo-W/skill/` (las ocho
  referencias, tal cual) con el `SKILL.md` de `contracts/skill-md-v0.3.1.md`, cuyo cuerpo tiene 10 995 caracteres. El
  sitio sale entero de reescribir más corto lo que no es pasaje medido, forma fija ni tabla (FR-006; contrato §2,
  T1 a T15). Ningún texto sale a una referencia y no hay referencia nueva (contrato §3).
- **Por qué**: FR-007 permite sacar un texto de una clase, no lo exige, y lo acota (Clarificación 3): el enlace solo
  en el paso 1, el 2 o el 4 fuera de su primera viñeta, y nada sale de «Protocolo», «Cómo se cita», «Redacción en una
  fecha», «Redacción modificada» ni «Reglas». Con esa frontera, los textos de una clase de la skill medida o ya son
  enlaces (paso 1, `redaccion_en_una_fecha`, `varios_bloques`, `lecturas_que_fallan`), o los ata un control
  (PowerShell: `ordenes-para-powershell` la exige en el mismo fichero que la orden), o son reglas invariantes (ley y
  reglamento), o el enlace que pedirían ocupa casi lo mismo que el texto («Nada de otra conversación»). El único grande
  que queda, la viñeta del índice del paso 2, decide qué id se pide: es guía, no un dato que el agente necesite para
  seguir, y lo medido dice que esa referencia no se abriría (`evidencias/sondeos-20261025/LEEME.md`, 57-59; V16).
  Reescribir deja en el cuerpo todo lo que hoy cambia una respuesta y cabe (V1, V3).
- **Alternativas**: (a) sacar los textos de una clase que la entrada del hito suma en «unos 886 caracteres»:
  rechazada por lo anterior; de esos, los que caen fuera de los pasos no pueden salir (FR-007), y los que caen dentro
  no darían el sitio sin perder lo que el agente usa. (b) Una referencia nueva para el índice con el enlace en el
  paso 2: rechazada, el agente compondría el id sin abrirla. (c) Rebajar el gate: fuera de alcance (ADR 0046).

### D2. Cómo se reescribe, y qué no se toca aunque quepa

- **Decisión**: cada texto reescrito quita solo (1) lo que el cuerpo ya dice en otro sitio —casi siempre en los pasos
  5 y 6 medidos, que repiten lo que antes decían el paso 4 y las reglas 2, 3 y 6—, o (2) lo que el agente no haría mal
  sin ello (el encabezado de una regla repetido en su texto, de dónde sale un dato que ya tiene). Cada fila del
  contrato §2 nombra dónde está lo quitado. La regla 7 no se toca: la eval 21, que decide, depende de ella, y lleva su
  porqué («no has consultado nada») y prohibiciones concretas («de memoria o con salvedades»), que `brazo-X` mostró que
  importan (quitarlas de la regla 5 la devolvió a 6 de 40; `evidencias/sondeos-20261026/LEEME.md`). Tampoco las dos
  prohibiciones de la viñeta del índice ni el aviso de que el texto consolidado es informativo.
- **Por qué**: FR-006 pide las prácticas de las guías (quitar lo que no hace falta y lo repetido, el porqué en vez de
  prohibiciones), y la evidencia del hito dice que las prohibiciones concretas de lo medido no sobran. Lo que se quita
  es lo repetido, no lo que solo está aquí.
- **Alternativas**: reescribir la regla 7 (22 caracteres menos) y quitar el aviso de la regla 3 (103): rechazadas, la
  primera toca lo que decide la eval 21 y la segunda quita contenido de la respuesta que ningún otro sitio da.

### D3. El título `#` del cuerpo sale

- **Decisión**: la línea 17 de la skill medida, `# Consultar y citar legislación consolidada del BOE`, y su línea en
  blanco salen (T1, 53 caracteres).
- **Por qué**: no pide nada al agente, que ya tiene el nombre y la descripción; no es un encabezado `##` ni `###` y
  nada remite a él (FR-006); ningún gate lo exige (V3). Era lo único que quedaba por quitar sin quitar instrucción.
- **Alternativas**: conservarlo y quitar a cambio la viñeta «Nada de otra conversación» (75) o el aviso de la regla 3
  (103): rechazadas, las dos son instrucción. Que `legal-core` y `jurisprudencia` sí lo lleven es forma, no
  comportamiento, y ningún control lo mira.

### D4. La suite de aceptación compara con la skill medida desde testscript

- **Decisión**: tres guiones (`contracts/aceptacion.md`), `h31-1-responder`, `h31-1-lee-lo-que-nombra` y
  `h31-1-sitio-en-el-cuerpo`, que leen la skill medida en `$KITLEGAL_SKILLS/../evidencias/sondeos-20261026/brazo-W/skill/`
  y comparan con `exec awk -f <programa del propio guion>`; cada comparación sale con 0 e imprime su veredicto, y la
  aserción es `stdout`, `cmp` o `! grep`.
- **Por qué**: SC-003 pide que la suite compare con ese fichero y salga en rojo con una diferencia. Desde testscript se
  llega a él como ya hace H23 (V5), y `awk` está en el runner y en macOS y ya lo usan los guiones. El veredicto por
  `stdout` es forzoso: `scripts/workflow/aceptacion.sh` (línea 55) da por inválido, no por rojo, un guion cuya salida
  dice `unexpected command`, y eso imprime un `exec` que sale con 1 (V5).
- **Alternativas**: (a) copiar los pasajes esperados dentro del guion: rechazada, duplica la evidencia y deja de
  comparar con ella. (b) Comparar el `SKILL.md` entero con `contracts/skill-md-v0.3.1.md`: rechazada, congelaría en la
  suite el texto reescrito, que el spec da a la revisión final (FR-006). (c) Una orden nueva del arnés e2e en
  `internal/app/e2e_test.go`: rechazada, cambia el arnés para un hito cuando `awk` basta. (d) Un test Go en
  `internal/skills`: no es la suite congelada que nombra SC-003.
- **Consecuencia**: la suite fija los pasajes, las formas, la tabla, los encabezados y las reglas a los de la skill
  medida; el hito que los cambie (H32) cambia también estos guiones, como cualquier guion activado cuyo comportamiento
  cambia.

### D5. US4 se prueba con los tests de Go, no con un guion

- **Decisión**: las lecturas las comprueban `TestEvalsDelRepositorio` (`lecturas` y `grabado`),
  `TestCacheDeLasSesiones/lecturas` y `TestLecturaDeLaEvalSinGrabar` (`contracts/lecturas-y-grabaciones.md` §4).
- **Por qué**: es la prueba independiente que da el spec para US4 y la que nombran SC-004 y SC-008. Preparar la caché
  de una sesión es `PrepararSesion`, de `internal/evals`; el arnés e2e reproduce las grabaciones de
  `internal/source/boe/testdata/` (`e2e_test.go:72`), no las de las evals, y el binario no cambia.
- **Alternativa**: un guion que copie las grabaciones de las evals a su reproducción y lea los nueve bloques:
  rechazada, repite `TestCacheDeLasSesiones` por otro camino sin requisito que lo pida (constitución V).

### D6. Un mismo `motivo` para las nueve lecturas

- **Decisión**: «Lo nombran, sin haberlo leído, respuestas de los sondeos de evidencias/sondeos-20261025/ o de los
  jobs de las evals, y la regla de leer lo que se nombra lo pide (H31.1; ADR 0043, clase b).»
- **Por qué**: FR-011 pide que el motivo diga que respuestas de los sondeos y de los jobs lo nombran sin haberlo leído
  y que la regla lo pide. Los sondeos son solo de la eval 15, así que «sondeos y jobs» para cada lectura sería falso
  en la 14 y la 17; con «o» es cierto para las nueve (V16; S2). Una sola constante en el mapa del test.
- **Alternativa**: un motivo por eval (sondeos en la 15, jobs en la 14 y la 17): igual de cierto y con tres textos;
  rechazada por más texto sin diferencia para nadie.

### D7. La tarea `[datos]` amplía dos entradas del manifiesto

- **Decisión**: los bloques nuevos van en las entradas de `Ley 35/2006,` y `Real Decreto Legislativo 2/2004,` de
  `internal/evals/testdata/grabaciones.json`, detrás de los de hoy, y su `para` dice por qué; la de `Ley 19/2013,` no
  cambia porque `a24` ya está (contrato §2).
- **Por qué**: `TestGrabarEvals` resuelve cada entrada por su búsqueda y pide sus bloques (V20); una entrada nueva para
  la misma norma repetiría búsqueda, metadatos e índice. Que `grabar_datos` solo pida lo que falta lo dice el mismo
  test.

### D8. El peor caso del trabajo de evals de `boe-legislacion` no cambia, ni su tope

- **Decisión**: `timeout-minutes` de `.github/workflows/evals.yml` sigue en 352 (21 120 s), y `internal/evals/definicion.go`
  no cambia. FR-015 queda hecho con este recálculo; `TestDefinicionDelJob` ya pone en rojo un tope menor que el peor
  caso.
- **Cálculo** (V13): peor caso = 485 s fuera de las sesiones + 900 s de la decisión + (⌈109/5⌉ + ⌈108/5⌉ + ⌈6/5⌉) ×
  (22 s + 240 s + 10 s) + 60 s + (⌈60×6/5⌉ + ⌈60×6/5⌉ + ⌈3×6/5⌉) × (35 s + 5 s) + 6 × 40 s = 485 + 900 + 46 × 272 + 60
  + 148 × 40 + 240 = **20 117 s**, el de hoy (`definicion_test.go:2997` y `:3112`). Leer lo que se nombra y repasar no
  añaden sesiones, evals ni respuestas que juzgar: alargan cada sesión, que corta su tope de 240 s (`topeDeUnaSesion`,
  fuera de alcance), con 10 s de margen. Lo único que crece fuera de la sesión es su preparación, con nueve lecturas
  más; el término de 22 s es la media de una sesión entera en H7.2, cota de la preparación sola, y la preparación con
  las nueve tarda 0,54 s frente a 0,18 s sin ellas (V6). El mayor de los tres trabajos sigue siendo `legal-core`,
  20 725 s, bajo el mismo tope.
- **Lo que no se mide aquí**: cuánto alargan las sesiones en el job. En el sondeo, la mediana pasó de 19 s (`brazo-H`)
  a 23 s (`brazo-W`) (S3); si un modo pasa de su objetivo, `duracion_de_las_sesiones` decide (SC-009).

### D9. El párrafo de `docs/SKILLS.md` §1

- **Decisión**: un párrafo, «Lo que se lee y lo que no», detrás del que dice cómo el cuerpo dice cuándo leer cada
  referencia (`contracts/guia-y-changelog.md` §1), con las cifras de `LEEME.md` y dos enlaces: `https://agentskills.io/`
  y `#las-dos-fuentes`.
- **Por qué**: ahí es donde la guía explica cómo se pide una referencia, que es la consecuencia de lo medido. La única
  URL de una guía que registra el repositorio es la de `web/src/components/Pie.astro:30` (V9); la de Anthropic no está
  en ningún fichero y no se escribe de memoria (FR-014).

### D10. `CONTRIBUTING.md` deja de nombrar «la regla 7» por la frase de `lecturas_que_fallan`

- **Decisión**: en `CONTRIBUTING.md:743`, «la frase de la regla 7» pasa a «la frase de `references/lecturas_que_fallan.md`».
- **Por qué**: es consecuencia de FR-001. En la skill medida la regla 7 es «Sin herramienta y sin binario»; la frase
  que esa línea de la guía de contribución describe («No se ha podido comprobar si la redacción ha cambiado desde una
  consulta anterior», forma de `formas_fijas` de `expresiones-prohibidas.yaml`) está en `lecturas_que_fallan.md:46`
  (V17). Sin el cambio, quien contribuye busca en la regla 7 una frase que no está.
- **Alternativa**: dejarla: rechazada, la entrega haría falsa una frase del repositorio.

### D11. La entrada de `CHANGELOG.md`

- **Decisión**: la primera de «### Cambiado» en *Unreleased*, con el texto de `contracts/guia-y-changelog.md` §2. La de
  v0.3 no se toca: describe esa versión.

### D12. Nada más cambia

Ni el binario, ni el arnés de evals salvo los dos tests de §4 del contrato de lecturas, ni `.github/workflows/`, ni
`docs/SOURCES.md` (la fuente no cambia), ni `schemas/`, ni ningún ADR (ninguna decisión de arquitectura cambia), ni
`jurisprudencia` ni `legal-core`. La versión de la skill solo va en `CHANGELOG.md` (spec, «Assumptions»).

## Verificaciones de esta sesión

- **V1. Tamaños.** `tail -n +16 <SKILL.md> | wc -m`: skill medida 12 081, `main` 10 992, `contracts/skill-md-v0.3.1.md`
  10 995; `wc -l`, 169 y 156. El gate cuenta runas (`utf8.RuneCount`, `internal/skills/sincronia.go:442`) frente a
  `maximoDeCaracteresDelCuerpo = 11000` (`:30`); `TestCuerpoDeLaSkill` (`internal/skills/sincronia_test.go`) lo prueba
  con un cuerpo de 11 000 caracteres y otro de 11 001, que da «cuerpo de 11001 caracteres (máximo 11000)» y con el que
  la sincronía no escribe nada (casos `once-mil-caracteres` y `once-mil-y-un-caracteres`).
- **V2. Lo que difiere de `main`.** `diff -rq evidencias/sondeos-20261026/brazo-W/skill skills/boe-legislacion`: solo
  `SKILL.md`, `buscar_la_norma.md` y `redaccion_modificada.md`; el diff de `SKILL.md` empieza en la línea 22, así que
  el frontmatter (1 a 15) es igual. La regla 7 de `main` («Si `kitlegal graph check` no termina con `0`») no está en
  la medida; `lecturas_que_fallan.md:39-46` lo dice.
- **V3. Los gates y controles con el contrato.** En un `git worktree` de `HEAD` en `$TMPDIR` con
  `contracts/skill-md-v0.3.1.md` como `SKILL.md` y las referencias de la skill medida, `go -C <copia> test -count=1 -run
  '^(TestSkillsDelRepositorio|TestOrdenesDeLasSkillsEmpotradas|TestNormasDelRepositorio|TestTerritorioDelRepositorio|TestJerarquiaDelRepositorio|TestEvalsDelRepositorio|TestIdentificadoresDeLasNormas|TestCuerpoDeLaSkill|TestSkillsEmpotradas)$'
  ./internal/app/ ./internal/skills/ ./internal/evals/` sale `ok` en los tres, y con `-v` pasan `referencias`,
  `sin-instrucciones-de-evals`, `normas-nombradas`, `prosa-de-la-skill`, `ordenes-para-powershell`,
  `linea-sin-consulta`, `formas-de-la-redaccion-fechada`, `avisos-de-la-skill`, `hallazgos-de-la-skill`,
  `lineas-de-aviso-de-las-referencias` y `ordenes-de-ejemplo`; `TestSkillsEmpotradas` (`skills_test.go`) también. Con
  el `SKILL.md` de la skill medida, `TestSkillsDelRepositorio` sale en rojo: «cuerpo de 12081 caracteres (máximo
  11000)». Ningún código de `internal/skills` exige un título `#` (`grep -rn '"# '`, sin resultados).
- **V4. Lo que no cambia.** Con awk sobre los dos ficheros: los pasajes medidos son idénticos (12, 6, 9 y 13 líneas),
  y los encabezados `##`/`###` (14), el número y el encabezado de las reglas (7), el frontmatter (15 líneas) y la tabla
  (26 líneas) también; las cinco líneas de los bloques `text` están en el contrato; el cuerpo no nombra ninguna de las
  normas de FR-009 (`grep -c`, 0).
- **V5. La suite.** Los tres guiones de `contracts/aceptacion.md`, como `zz-h31-1-*.txtar` en la copia, con
  `go -C <copia> test -count=1 -v -run '^TestEntregaDelHito$/^zz-h31-1-' ./internal/app/`: pasan con el contrato;
  con la skill de `main` fallan los tres por `no match for … found in stdout` («distinto: 6 líneas frente a 4»,
  «distinto: 12 líneas frente a 4», «encabezados: distinto: 14 frente a 13»), sin `unexpected command`; un mutante
  con otro encabezado en la regla 5 da «reglas: distinto en 5», y otro con «TRLRHL» en el cuerpo, `unexpected match`.
  Un `exec awk` que sale con 1 imprime `unexpected command failure`, que `aceptacion.sh:55` cuenta como inválido.
  testscript no expande variables entre comillas simples (`testscript/doc.go:82-84` de `rogpeppe/go-internal`
  v1.16.0, la de `go.mod`); `KITLEGAL_SKILLS` es la ruta absoluta de `skills/` (`internal/app/e2e_test.go:356-367`) y
  `h23-cita-herramientas.txtar:44` ya lee `evidencias/` con ella.
- **V6. La preparación con las lecturas.** En la copia, con las grabaciones, el manifiesto y las evals 14, 15 y 17 de
  la rama del primer run y la tabla de `probarLasLecturasDeLasEvals` con los nueve bloques:
  `TestCacheDeLasSesiones/lecturas` pasa en 0,54 s, frente a 0,18 s con los cuatro de hoy.
- **V7. Genericidad de hoy.** El cuerpo de `main` no nombra ninguna de las normas de FR-009 (`grep -c`, 0); entre las
  referencias, solo `como_se_cita.md` y `normas.md` nombran alguna, iguales en `main` y en la medida (`grep -l`).
- **V8. Los ids.** Los nueve están en el índice grabado de su norma
  (`testdata/evals/boe.legislacion-consolidada/…_<norma>_texto_indice.json`, `grep -F '\"id\": \"<id>\",'`).
- **V9. La fuente y la URL.** `git show main:docs/SOURCES.md`, línea 42: `boe.legislacion-consolidada`, ritmo `1s`,
  «Revisado» `2026-09-13`. `web/src/components/Pie.astro:30`: `https://agentskills.io/`.
- **V10. Las grabaciones del primer run.** `git ls-tree -l origin/028-h31-1-boe-legislacion-habla`: de 1 766 bytes
  (`a60`) a 402 526 (`a7`); `a19` 5 501, `a20` 25 889, `dadecimoctava` 4 240, `a92` 2 761, `a100` 5 897, `a104` 27 604.
- **V11. Los controles de las lecturas.** En la copia de V6: `TestEvalsDelRepositorio/lecturas` sale en rojo con el
  mapa de hoy (le faltan las nueve) y `grabado` pasa; sin la grabación de `a7`, `grabado` y
  `TestCacheDeLasSesiones/lecturas` salen en rojo con la falta «la lectura BOE-A-2006-20764 a7 sin su bloque»;
  `TestLecturaDeLaEvalSinGrabar` pasa. Sin la entrada del manifiesto pero con la grabación, `grabado` pasa: lo que
  comprueba es lo grabado, y el manifiesto es lo que `grabar_datos` graba.
- **V12. Los sobres.** Con un test temporal en la copia, no versionado, que lee cada bloque con `boe articulo …
  --offline --json` desde la caché de una sesión de la 01: `a17` 9 345 bytes, `a18` 4 802, `a7` 22 303, `a19` 2 818,
  `a20` 2 920, `dadecimoctava` 3 060 (los cuatro nuevos de la 15, 31 101), `a60` 893, `a92` 1 704, `a100` 2 061,
  `a104` 7 779, `a24` 2 690. El de `a60`, entero, en plan.md («Uso»). Y con un guion e2e temporal en la copia, `exec
  kitlegal boe articulo BOE-A-2015-10565 a21 --json` y después `exec kitlegal graph check BOE-A-2015-10565 a21 --json`:
  el sobre de la comprobación tiene 324 bytes (`wc -c`), con `bloques: ["a21"]` y sin hallazgos.
- **V13. El peor caso.** `internal/evals/definicion.go:74-88` (485 s, 22 s, 240 s, 10 s), `definicion_test.go:2997` y
  `:3112` (boe-legislacion, 20 117 s con sus términos), `.github/workflows/evals.yml:276` (`timeout-minutes: 352`).
- **V14. El umbral de la descripción no va.** `internal/evals/decision.go:85`: sin una descripción que cambie, el
  cálculo no tiene huellas de descripciones y no se juzga ninguna; `umbrales.go:420-430`: sin ella,
  `habla_en_primera_o_segunda_persona` no va en el informe. FR-010 no cambia la descripción.
- **V15. Los umbrales que deciden.** `jq '.umbrales[] | select(.decide == true) | .nombre'` sobre
  `specs/027-h31-las-skills-siguen/gates/evals/<skill>.json`: los nombres de la tabla «Controles de umbral»
  de plan.md, más `habla_en_primera_o_segunda_persona` (V14); con `decide: false`, `afirma_lo_no_leido` por modo,
  `cuenta_su_proceso` por modo y `fuera_de_lo_grabado` por modelo y modo.
- **V16. Lo que nombran las respuestas.** `evidencias/sondeos-20261025/LEEME.md`: control, `brazo-A` y `brazo-A2`
  (42, 47); `brazo-D` y `responder.md` (45, 57); las referencias que se abren (57-59). En `votos.json` de `brazo-A` y
  `brazo-A2`, 68 de 80 respuestas nombran el art. 7 y 52 la disposición adicional decimoctava. En los informes
  versionados de los cierres, la eval 17 nombra el art. 24 en las 6 respuestas de H31 y en las 6 del primer run, y la
  14, el art. 60 en una de H30.
- **V17. `CONTRIBUTING.md`.** Líneas 741-743; la frase está en `skills/boe-legislacion/references/lecturas_que_fallan.md:46`
  y en la misma referencia de la skill medida.
- **V18. Caracteres por token.** ADR 0046, línea 24: 2,26 caracteres por token en el español de las tres skills.
- **V19. Go.** `go version`: go1.27.3; `go.mod`: `go 1.27.0`, `toolchain go1.27.3`.
- **V20. La grabación.** `internal/evals/grabacion_test.go:38-75`: `TestGrabarEvals` sirve primero cada consulta desde
  lo grabado y solo pide lo que da una falta, por entrada del manifiesto (búsqueda, metadatos, índice y cada bloque).

La copia de `$TMPDIR` y el test temporal de V12 no son parte de la entrega; el worktree se retira al acabar esta
sesión.

## Supuestos

- **S1. Lo reescrito no se midió.** La skill de la entrega ya no es la réplica de `brazo-W`: decide el sondeo de una
  persona (SC-012), no el cierre.
- **S2. Lo que nombran las respuestas de los jobs.** Que las respuestas nombran sin leer `a19`, `a20`, `a92`, `a100` y
  `a104` lo dice la entrada del hito; en lo versionado que se ha buscado (V16) no aparecen con número. El motivo de D6
  no lo afirma de cada bloque.
- **S3. La duración en el job.** No se ha medido cuánto alargan las sesiones la lectura y el repaso en el job; el
  sondeo da 23 s de mediana con la skill medida, frente a 19 s con la de la #200.
- **S5. El awk del runner.** Los programas de la suite solo se han ejecutado con el awk de macOS (`awk version
  20200816`; no hay gawk ni mawk en esta máquina). Usan lo que POSIX define y ninguna ancla dentro de una alternancia,
  y los guiones de H23 ya ejecutan `awk` en el runner; que den lo mismo con el awk de `ubuntu-24.04` lo comprueba el CI
  del cierre.
- **S4. El `robots.txt`.** Que `grabar_datos` regrabe `GET_https_www.boe.es_robots.txt.json` con otras cabeceras es lo
  que hizo en el primer run (su diff con `main`, `git diff --stat`); sin red no se ha comprobado aquí.

## Fuentes fuera de alcance

Ninguna: la única fuente que el hito graba tiene su fila revisada en `main` (V9).
