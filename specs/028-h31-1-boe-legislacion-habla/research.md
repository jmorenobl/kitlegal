# Research: H31.1 · `boe-legislacion` habla de la norma, no de la consulta, y lee lo que nombra

Decisiones del plan (D), con la alternativa descartada y por qué; verificaciones hechas en esta sesión y vistas
terminar (V), con la orden y lo que dio; y supuestos que no se pudieron verificar sin red ni modelo (S). El spec no deja
ninguna marca de aclaración pendiente: sus dos preguntas están resueltas en «Clarifications» (Q1, el tope de una sesión;
Q2, qué lecturas son el salto).

## Decisiones

### D1 · Las reglas de los dos pasajes de `27aeccc` vuelven regla a regla, con la redacción de `27aeccc` donde H31 las comprimió

- **Decisión**: cada regla de la última viñeta del paso 4 y del paso 5 de `27aeccc` (B47 a B69 de la tabla de H31)
  tiene su frase en el cuerpo de v0.3.1 (contracts/skill-boe-legislacion-v0-3-1.md §2). Las que el cuerpo de hoy ya
  dice con el mismo contenido se quedan como están (FR-001); las que H31 comprimió —B48, B50, B52 y B53— vuelven con
  las palabras de `27aeccc` y sus dos ejemplos; las de lo remitido y del precepto no leído —B58 a B61— vuelven como las
  dejan FR-002 y FR-003, con los dos ejemplos que H31 quitó.
- **Por qué**: es lo que mide el sondeo (`brazo-E`, la skill de antes de H31, 0 de 40 en `cuenta_su_proceso`; la de
  hoy, 17 de 80) y lo que pide el spec regla a regla. Lo que vuelve es lo que el sondeo atribuye al texto: la frase de
  qué es la respuesta, desde su primera palabra, quién la lee y qué no ve, y lo que no le sirve.
- **Alternativas**: (a) el paso 5 de `27aeccc` entero, palabra por palabra: unos 1 220 caracteres más que hoy (medido
  sobre un borrador, V6), que no caben en 11 000 sin sacar del cuerpo reglas invariantes o formas fijas, que FR-006
  prohíbe; (b) una referencia `responder.md` con la guía: `brazo-D` la puso así y ninguna de sus 40 sesiones la abrió,
  con `afirma_lo_no_leido` peor (15 de 40); descartada por la evidencia y por FR-004.

### D2 · B48, en la misma frase que dice qué no ve quien pregunta

- **Decisión**: «quien pregunta la lee entera: no ve las órdenes, las llamadas ni lo que devuelven, y lo que decides al
  completar la lectura no va en ella» (paso 4). B53 se junta con B52 en «sin decir de dónde sale, qué has hecho o vas a
  hacer ni qué ha devuelto o no una orden: no le sirve».
- **Por qué**: US1, escenario 2, pide las cuatro cosas en el paso «Responder», y «al completar la lectura» nombra el
  paso 3 de hoy, que es el «Evaluar si falta contexto» de `27aeccc`. Dicho una vez, cuesta unos 15 caracteres menos que
  en dos frases.
- **Alternativa**: la viñeta de `27aeccc` seguida de una frase propia para B48 («Lo que decidas al completar la
  lectura tampoco va en la respuesta: quien pregunta no lo ve»): unos 80 caracteres más, que el presupuesto no da.

### D3 · FR-002 en el paso 3, detrás de su regla de hoy, que no cambia

- **Decisión**: el paso 3 conserva su frase de hoy y gana una segunda: «Nombra un precepto solo si
  `kitlegal boe articulo` o `articulos` ha devuelto su texto en esta conversación, salvo lo del paso 5: si lo remite un
  bloque leído en el paso 2, o su id solo salió en un índice, léelo antes con los pasos 1 y 2, o no lo nombres; leído,
  di de qué trata, con su cita.»
- **Por qué**: FR-002 pide en el paso 3 las tres cosas —solo se nombra lo que esas órdenes han devuelto en la
  conversación y el índice no basta; lo remitido por el paso 2 o salido de un índice se lee antes con los pasos 1 y 2, o
  no se nombra; y leído, se dice de qué trata con su cita— y FR-003 exige que la regla de hoy siga igual. «Precepto»
  es la palabra del título del paso 5, que lo define como artículo, apartado o disposición: así la regla no alcanza a
  la norma que se nombra, que no se lee con `articulo`.
- **Alternativa**: reescribir la frase de hoy para que las dos lecturas compartan «con los pasos 1 y 2»: ahorra unos
  20 caracteres, pero cambia la regla que FR-003 manda conservar tal cual.

### D4 · El salto, solo de lo leído para nombrarlo

- **Decisión**: paso 5, «Lo leído para nombrarlo (paso 3) es un salto: lo que remite no se lee para nombrarlo y, como lo
  que no se pudo leer, se nombra solo con las palabras del texto leído.»
- **Por qué**: la Clarificación Q2 fija que el salto son solo las lecturas de FR-002 y que lo remitido que cambia la
  respuesta se sigue leyendo sin límite de saltos. «Para nombrarlo» separa las dos: lo que remite una lectura de FR-002
  no se lee para nombrarlo, y si cambia la respuesta lo lee la regla de hoy del paso 3. La lectura que falla se nombra
  igual, como dice hoy el paso 5 (FR-003), y la regla 2 sigue diciendo qué bloques no se pudieron consultar.
- **Alternativa**: «lo que remite no se lee», sin más: contradice la regla del paso 3 cuando lo remitido cambia la
  respuesta, que el spec deja fuera de alcance limitar.

### D5 · B59, el índice que no basta, va con FR-002 en el paso 3

- **Decisión**: el paso 5 deja «y el índice solo da números»; el paso 3 dice que se nombra un precepto solo si
  `kitlegal boe articulo` o `articulos` ha devuelto su texto en esta conversación, y que el que solo salió en un índice
  se lee antes con los pasos 1 y 2, o no se nombra.
- **Por qué**: FR-001 manda que las reglas del precepto no leído vuelvan como las deja FR-002, y FR-002 dice que el id
  en un índice no basta: el índice da el id, no el texto. Decirlo también en el paso 5 cuesta 29 caracteres (V6).
- **Alternativa**: las dos frases: 10 993 + 29 = 11 022, 22 más que el máximo.

### D6 · El sitio: siete textos de una sola clase de pregunta a referencias que ya existen

- **Decisión**: los siete de contracts/skill-boe-legislacion-v0-3-1.md §3 —la regla 7 de la comprobación que falla, el
  detalle de la lectura con fecha, el de la línea de la redacción modificada, lo demás de `graph check`, la segunda
  lectura de un bloque, la frase de la norma que la búsqueda no da y «(«Comandos»)»—. Cuatro ya estaban en su
  referencia y salen sin tocarla; dos (filas 4 y 5) van a apartados nuevos de `redaccion_modificada.md`, y la 6, a una
  frase de `buscar_la_norma.md`. El enlace de `redaccion_en_una_fecha.md` pasa del apartado «Redacción en una fecha» al
  paso 2, que es el paso que lee. Cuerpo resultante: 10 993 caracteres (V6).
- **Por qué**: FR-006. Cada texto lo necesita una clase de pregunta —con fecha, con `version-obsoleta`, con la
  comprobación que falla, con herramientas, con una norma que no está en la tabla— y su referencia la pide un paso que
  ya la enlaza con la condición de esa clase; varias son referencias que las sesiones abren porque traen un dato para
  seguir (la orden con `--fecha`, la frase de la comprobación que falla, los nombres de las herramientas), que es lo que
  mide la evidencia (`LEEME.md`).
- **Alternativas descartadas, por texto**: «si faltó, pídela ya», de la comprobación en el paso 2: es del protocolo y
  vale en los dos modos (B36 de la tabla de H31, `specs/027-h31-las-skills-siguen/research.md`), y `con_herramientas.md`
  solo se lee con herramientas, así que una sesión en modo orden que leyera un bloque sin su comprobación no sabría
  pedirla ni llevaría la línea REDACCIÓN MODIFICADA; la definición de fecha («un día, un mes, un año o un hecho fechado»): sin
  ella, un año puede no leerse como fecha y `redaccion_sin_fechar` decide con 0 en las evals 22 y 23; la explicación de
  la regla 8 (sin binario): la eval 21 la mide y la referencia no traería ningún dato para seguir; la regla 5: el spec
  dice que no cambia; la cola de la regla 4: no hay referencia que la reciba y su enlace costaría lo mismo; la forma para
  PowerShell: `ordenes-para-powershell` la exige detrás de su orden de Bash en el mismo fichero; «aunque acabe en punto»:
  `como_se_cita.md` lo dice, pero su enlace («para otra forma») no dice que se lea por un id que acaba en punto, y
  ajustar la condición cuesta más que el texto; «si importan la vigencia o una modificación, los metadatos y el
  análisis»: su enlace en el paso 3 costaría lo mismo; la tabla de comandos: FR-006 dice que no cambia.

### D7 · La regla 7 fuera: la 8 pasa a ser la 7

- **Decisión**: la regla 8 de hoy se renumera 7, la regla 2 remite a «la regla 7» y `CONTRIBUTING.md` deja de llamar
  «la frase de la regla 7» a la de las lecturas que fallan (contracts/guia-y-changelog.md §3).
- **Por qué**: un hueco en la numeración se renderiza como 7 y deja el texto fuente con un 8; y `CONTRIBUTING.md` diría
  una regla que ya no es. Ninguna referencia de la skill, ningún test y ningún guion dependen del número (V9): los
  comentarios de `internal/evals/prohibidas_test.go` citan el contrato de H7.4, en el que la frase era de la regla 7, y
  `docs/USO.md` es una bitácora; no se tocan.
- **Alternativa**: dejar una regla 7 de una línea con el enlace: unos 110 caracteres que repiten el enlace del paso 2 y
  dejan el cuerpo por encima de 11 000.

### D8 · Las lecturas, detrás de la grabación

- **Decisión**: una tarea `[datos]` añade los ocho bloques al manifiesto y nada más; `grabar_datos` los graba; la tarea
  siguiente añade las nueve lecturas a las evals y los tests de §4 de contracts/lecturas-y-grabaciones.md.
- **Por qué**: con una lectura sin grabar, la caché de cualquier sesión de la skill no se prepara y salen en rojo
  `grabado`, `ordenes-de-ejemplo` y `lo-que-lleva-a-leer` (V7); con el manifiesto solo, `make ci` sigue en verde
  porque en un build normal el manifiesto solo se lee (`TestManifiestoDeGrabaciones`). Es el orden de H26.1.
- **Alternativa**: lecturas y manifiesto en la misma tarea: una tarea `[datos]` no mezcla otro trabajo (workflow
  `tasks`), y la verificación de la tarea correría `make ci` antes de grabar.

### D9 · `TestCacheDeLasSesiones/lecturas`, con la sesión de cada eval

- **Decisión**: cada fila de la tabla lleva la eval de cuya sesión se prepara la caché: las cuatro de hoy, la 01; las
  nueve nuevas, la suya (14, 15 o 17).
- **Por qué**: US3, escenarios 1 y 2, y SC-004 dicen «la caché de su eval». La caché es la misma unión en todas las
  sesiones de la skill salvo el `sin_servir` de la eval (`prepararLaCache`), y las tres no tienen `sin_servir`, así que
  la sesión de la 01 también las serviría; probarlas desde la suya es la lectura literal y cuesta tres preparaciones más,
  de menos de medio segundo cada una (V6).
- **Alternativa**: las nueve desde la sesión de la 01, como las de hoy: más corto, pero no prueba lo que dice el
  escenario.

### D10 · El tope del trabajo no cambia (FR-013)

- **Decisión**: ni `timeout-minutes` de `.github/workflows/evals.yml` ni `internal/evals/definicion.go` cambian.
- **Por qué**: el peor caso del trabajo de `boe-legislacion` es 20 117 s (V3): 485 s fuera de las sesiones, 900 s de la
  decisión, 46 tandas de 272 s (⌈109/5⌉ + ⌈108/5⌉ + ⌈6/5⌉; cada sesión en su tope de 240 s, con 10 s hasta KILL y 22 s de
  preparación), 60 s + 148 tandas de 40 s del juez y 240 s del juez de la descripción. Las lecturas no entran en él:
  `peorCaso` cuenta sesiones y respuestas juzgadas, no lo que hace cada sesión (`definicion.go`, 752 a 797), y el hito
  no añade evals ni repeticiones. 352 minutos son 21 120 s y lo cubren; `TestDefinicionDelJob` lo comprueba y sale en
  rojo si no (`fallosDelTope`). El objetivo de duración de cada modo (900 s, umbral que decide) midió 410 y 348 s en el
  cierre de H31 (V10); la regla nueva añade una lectura por precepto remitido que la respuesta nombra (cinco en la
  eval 15) a las sesiones que lo nombran, que corren de cinco en cinco: un cálculo, no una medida (S3), que deja margen.
- **Alternativa**: subir el tope de una sesión: fuera de alcance (Clarificación Q1).

### D11 · El párrafo de `docs/SKILLS.md` §1, junto a la regla del reparto

- **Decisión**: detrás de «**En una referencia, el detalle que solo pide una clase de pregunta**», con las cifras de
  `LEEME.md` y su ruta (contracts/guia-y-changelog.md §1).
- **Por qué**: es donde la guía dice qué va al cuerpo y qué a una referencia; lo medido lo justifica, y las dos
  consecuencias de FR-012 son la regla de ese apartado.
- **Alternativa**: un apartado propio: separaría la medida de la regla que la usa.

### D12 · `CHANGELOG.md`: una entrada nueva que sustituye a v0.3

- **Decisión**: en «### Cambiado» de *Unreleased*, detrás de v0.3, una entrada v0.3.1 con «Sustituye a
  `boe-legislacion` v0.3».
- **Por qué**: es como la v0.2.2 sigue a la v0.2.1 dentro de 0.8.0 (líneas 455 a 461 de `CHANGELOG.md`): cada versión
  de la skill tiene su entrada.
- **Alternativa**: reescribir la entrada de v0.3: perdería lo que H31 hizo y lo que esta versión corrige.

### D13 · Aceptación e2e: no aplica

- **Decisión**: ningún guion `testscript`.
- **Por qué**: el binario no cambia —ningún applet, verbo, bandera, código ni salida—; `kitlegal skills install` instala
  lo empotrado y `TestSkillsEmpotradas` comprueba sus bytes. La aceptación sin persona es el job de evals sobre la cabeza
  (SC-007 a SC-011) y la de la persona, el sondeo (SC-012). Es la de H24, H26 y H31.
- **Alternativa**: un guion que instale la skill y busque una frase del cuerpo: fijaría en la suite congelada un texto
  que diseña el hito y repetiría `TestSkillsEmpotradas`.

### D14 · Ningún control nuevo para FR-004 y FR-008

- **Decisión**: las líneas que la rama añade al cuerpo sin las normas de las evals por materia (FR-004) y la
  descripción igual a la de `main` (FR-008) se comprueban con `git diff main` (quickstart, escenarios 3 y 4) y las coteja
  la revisión final; `sin-instrucciones-de-evals` y `normas-nombradas` siguen en `make ci`.
- **Por qué**: son criterios sobre el diff de esta rama contra `main`, que deja de tener sentido al fusionarla; un test
  permanente que vetara esas normas en el cuerpo sería una regla nueva que ningún FR pide (proporcionalidad).
- **Alternativa**: un subtest de `TestSkillsDelRepositorio` con la lista: ataría el cuerpo a tres normas de tres evals
  para siempre.

### D15 · El prototipo, fuera del repositorio

- **Decisión**: el cuerpo y las referencias se validaron en una copia de `HEAD` bajo `$TMPDIR`
  (`git archive -o` y un extractor de `archive/tar` de la biblioteca estándar), sin tocar el árbol, el índice ni el
  historial; el texto de los contratos es el de esa copia.
- **Por qué**: medir el presupuesto con el gate y no con una cuenta a mano, y ver en verde los controles de contenido
  antes de fijar el texto; `git worktree` deja un registro en `.git` y la memoria del proyecto lo desaconseja.

## Verificaciones (hechas en esta sesión)

- **V1 · El cuerpo de hoy**: `awk 'NR>1 && /^---$/ {f=1; next} f' skills/boe-legislacion/SKILL.md | wc -m` → 10 992,
  la cifra del spec; `head -15 … | wc -m` → 1 012 del frontmatter, 12 004 el fichero.
- **V2 · El árbol de antes de H31**: `git rev-parse d14bb86^:skills/boe-legislacion` →
  `27aeccc6ce6df6e73d5b0c043ed5375743e310ff`; `git show d14bb86^:skills/boe-legislacion/SKILL.md | sed -n '92,131p'`
  da el paso 4 (viñeta final en la línea 98) y el paso 5 (100 a 130); las filas B44 a B69 de
  `specs/027-h31-las-skills-siguen/research.md` citan esas líneas.
- **V3 · El peor caso**: un test temporal en `internal/evals` (borrado en el acto; `git status` limpio) que llama a
  `leerDefinicionDelJob(rutaDeLaDefinicionDelJob).peorCaso("../../evals", skill)` dio, con `go test -run`:
  `boe-legislacion: 20117 s = 485 s + 900 s + (⌈109 / 5⌉ + ⌈108 / 5⌉ + ⌈6 / 5⌉) × (22 s + 240 s + 10 s) + 60 s +
  (⌈60 × 6 / 5⌉ + ⌈60 × 6 / 5⌉ + ⌈3 × 6 / 5⌉) × (35 s + 5 s) + 6 × (35 s + 5 s)`; `jurisprudencia: 13717 s`;
  `legal-core: 20725 s`; tope 352 min. `TestDefinicionDelJob` en verde (`go test -run '^TestDefinicionDelJob$'`).
- **V4 · Los ids en los índices grabados**: `jq` sobre `.respuesta.cuerpo | fromjson | .data[0].bloque[]` de los
  `…_texto_indice.json` de `testdata/evals/boe.legislacion-consolidada/`: `a7` «Artículo 7», `a19`, `a20`,
  `dadecimoctava` «Disposición adicional decimoctava» (y `a17`, `a18`) en `BOE-A-2006-20764`; `a60`, `a92`, `a100`,
  `a104` (y `a15`, `a59`) en `BOE-A-2004-4214`; `a24` (y `a20`) en `BOE-A-2013-12887`. El art. 17 grabado remite al
  «artículo 7 de esta Ley» y a la «disposición adicional decimoctava de esta Ley» (`grep` sobre su cuerpo).
- **V5 · El manifiesto y las grabaciones de hoy**: `jq` sobre `internal/evals/testdata/grabaciones.json`:
  `"fuente": "boe.legislacion-consolidada"`, `Ley 35/2006,` con `a18`, `Real Decreto Legislativo 2/2004,` con `a15`,
  `Ley 19/2013,` con `a24`; `ls` de `internal/evals/testdata/boe.legislacion-consolidada/`: grabado `a24` (3 844 bytes),
  `a18` (18 728), `a15` (2 697). `docs/SOURCES.md` de `main`: la fila de `boe.legislacion-consolidada` con «Revisado»
  2026-09-13. `TestGrabarEvals` (`internal/evals/grabacion_test.go`, `//go:build grabacion`) sirve primero de lo grabado
  y no escribe en `evidencias/`.
- **V6 · El prototipo del cuerpo**: clon de la rama (`HEAD`, `3dec6e3`) en `$TMPDIR`, con el cuerpo montado con `sed`
  desde las líneas de hoy que no cambian y los bloques de §1.1 y §1.2 de contracts/skill-boe-legislacion-v0-3-1.md, y
  `redaccion_modificada.md` y `buscar_la_norma.md` con los de §4:
  `go -C <copia> test -count=1 -v -run '^(TestSkillsDelRepositorio|TestOrdenesDeLasSkillsEmpotradas|TestNormasDelRepositorio|TestTerritorioDelRepositorio|TestJerarquiaDelRepositorio|TestEvalsDelRepositorio|TestIdentificadoresDeLasNormas|TestSkillsEmpotradas|TestCuerpoDeLaSkill|TestControlesSobreLasReferenciasAMano)$' ./internal/app/ ./internal/skills/ ./internal/evals/ .`
  → `--- PASS` de los diez y `ok` en los cuatro paquetes; `wc -m` del fichero, 12 005, menos 1 012 del frontmatter:
  10 993 (hoy, 12 004: 10 992). Un mutante de 8 caracteres más (11 001) dio
  `cuerpo de 11001 caracteres (máximo 11000)` en `TestSkillsDelRepositorio`. Los seis ejemplos de FR-001, cada uno en
  una línea (`grep -oF -f` con los seis), y ninguna de las normas de FR-004 en las líneas añadidas (`git diff` de la
  copia, `grep -cE` → 0).
- **V7 · Una lectura sin grabar**: en la copia, con `a7` en las `lecturas` de la 15 y `a24` en las de la 17,
  `TestEvalsDelRepositorio/grabado` dio «la lectura BOE-A-2006-20764 a7 sin su bloque… no hay grabación de GET …/bloque/a7»
  y ninguna falta de `a24`; `lecturas` falló por su mapa escrito a mano, y `ordenes-de-ejemplo` y
  `lo-que-lleva-a-leer/boe-legislacion`, porque la caché de la sesión no se preparó. Sin esas dos lecturas, todo en verde.
- **V8 · La evidencia**: `evidencias/sondeos-20261025/LEEME.md`, «Los sondeos»: `brazo-E` 0 de 40 y `brazo-A` +
  `brazo-A2` 17 de 80 en `cuenta_su_proceso`; ninguna de las 40 sesiones de D abrió `responder.md`; `normas` de 23 a 40
  sesiones por brazo y `con_herramientas` de 6 a 9 en los brazos con la skill de H31.
- **V9 · Nada depende del texto que sale ni del número de la regla 8**: `grep -rlE` de las frases que salen y de «la
  regla 8» en `internal/`, `testdata/`, `evals/`, `web/src` y `scripts/` → nada; «regla 7» de `boe-legislacion` solo en
  comentarios de `internal/evals/prohibidas_test.go`, en `docs/USO.md` (bitácora) y en `CONTRIBUTING.md:743`.
- **V10 · Los umbrales del job**: `jq` sobre `specs/027-h31-las-skills-siguen/gates/evals/*.json` (el cierre de H31):
  veredicto `aprobado` en las tres; los que deciden de cada skill, los de plan.md, «Controles de umbral», más
  `habla_en_primera_o_segunda_persona`, que solo va cuando la propuesta cambia una descripción; `duracion_de_las_sesiones`
  de `boe-legislacion`, 410 s (orden) y 348 s (herramienta) frente a 900. `evals/boe-legislacion/juez/clases.yaml`:
  `afirma_lo_no_leido` decide con 0.028 y `cuenta_su_proceso` no decide. `internal/evals/umbrales.go`:
  `fuera_de_lo_grabado` no decide nunca; `internal/evals/sesion.go:40`: el motivo «tope de 240 s agotado (código 124)».

- **V11 · El tamaño de un sobre de lectura**: en la copia del prototipo, un test temporal (borrado) preparó la caché de
  una sesión de la 01 con `cacheDeUnaSesion` y ejecutó `app.Main` con `boe articulo <norma> <bloque> --offline --json`:
  `a17` de la Ley 35/2006, 9 345 bytes; `a18`, 4 803; `a59` del texto refundido de Haciendas Locales, 1 302; `a15`,
  1 708; `a20` de la Ley 19/2013, 2 432; `a24`, 2 690. Todos con código 0. Los bloques nuevos no estaban grabados y su
  tamaño no se conocía hasta `grabar_datos`. Tras grabarlos, el barrido repitió la medida en una copia de la rama bajo
  `$TMPDIR`, desde la caché de una sesión de su eval: `a60`, `a92`, `a100` y `a104` del texto refundido, 893, 1 704,
  2 061 y 7 779 bytes; `a7`, `a19`, `a20` y `dadecimoctava` de la Ley 35/2006, 22 304, 2 819, 2 921 y 3 061 (con `a18`,
  los cinco de la eval 15 suman 35 908); todos con código 0, y `a18` y `a24` dan los mismos bytes que arriba. Las
  grabaciones van de 1 766 bytes (`a60`) a 402 526 (`a7`, con veinte redacciones del bloque).

- **V12 · Lo que deja el quickstart**: con `LC_ALL=C`, la misma cuenta del cuerpo de hoy da 11 168 (bytes), no 10 992:
  el escenario 1 lo advierte. `make ci` escribe `coverage.out` y `coverage-integration.out` (`Makefile`, líneas 87 y 91,
  `-coverprofile`), que `.gitignore` ignora (`git check-ignore -v`: `/coverage.*` y `/*.out`). Los escenarios 2, 3, 4,
  6 (`git ls-files`), 7 (`grep`) y 8, ejecutados hoy, dan los valores «de hoy» que citan.
- **V13 · Lo que SC-008 cuenta y lo que no**: `jq -c '.fuera_de_lo_grabado[] | select(.eval | test("^1[457]-"))'`
  sobre `specs/027-h31-las-skills-siguen/gates/evals/boe-legislacion.json` (el cierre de H31) da dos invocaciones, las
  dos de la sesión `14-trlrhl-impuestos-por-materia-herramienta-claude-sonnet-5-5-03` y con código 5:
  `boe_articulos BOE-A-2004-4214 a57 a59` y `boe_articulo BOE-A-2004-4214 a57`. `a59` es el comando de la eval 14 y
  está grabado (`testdata/evals/boe.legislacion-consolidada/`, `…_BOE-A-2004-4214_texto_bloque_a59.json`); `a57` no está
  en ninguna de las dos carpetas de grabaciones (`ls`). La orden de dos bloques salió entera con 5 por el que no estaba
  grabado. `internal/evals/juzgar.go` pone en `FueraDeLoGrabado` toda invocación con 4 o 5 que no consulte una lectura
  del `sin_servir` de su eval, y `internal/evals/umbrales.go` dice que su umbral no decide. `internal/evals/consultas.go`:
  la caché preparada sirve cada bloque de las `lecturas` sin fecha y con `--fecha 9999-12-31`, «así el bloque en
  cualquier fecha».

Herramientas, comprobadas en local: `go version` → go1.27.3 (`go.mod`: `go 1.27.0`, `toolchain go1.27.3`); `-run`
parte su patrón por las barras sin corchetes y cada parte casa con un nivel del nombre del test (`go help testflag`),
por eso los escenarios usan `'^Test…$/^subtest$'`; `git archive -o <fichero>` (`git archive -h`); `git diff -U0 main --
<ruta>` y `git diff main -- <ruta>`, ejecutadas en esta sesión sobre la rama (sin cambios en `skills/` aún: 0 hunks y
0 líneas); `jq` lee las grabaciones (`.respuesta.cuerpo` es una cadena JSON: `fromjson`).

## Supuestos no verificados

- **S1 · El efecto en las respuestas.** Que el cuerpo de v0.3.1 baje `cuenta_su_proceso` cerca de 0 de 40 y
  `afirma_lo_no_leido` a 2 de 40 en la eval 15 no se puede medir sin modelo: lo mide el job del cierre (SC-007 a SC-011)
  y, con potencia, el sondeo de una persona (SC-012). El texto no es el de `27aeccc` palabra por palabra (D1), y el
  sondeo dirá si basta.
- **S2 · Lo que se graba.** Que `TestGrabarEvals` solo pida al BOE los ocho bloques —la búsqueda, los metadatos y el
  índice de las dos normas están grabados desde H5, como dice el `para` de H26.1— no se puede ver sin red; lo registra
  `gates/grabaciones.md` del run. Tras grabar, registra los ocho bloques y, además, el `robots.txt` del BOE, que
  `internal/httpx` pide antes de las peticiones y queda regrabado con el mismo cuerpo (otro `grabado_en` y otras
  cabeceras `Age` y `Date`); ni búsqueda, ni metadatos, ni índice.
- **S3 · La duración de las sesiones.** Lo que tarda una sesión de la eval 15 con cinco lecturas más no está medido
  (la evidencia no guarda duraciones). Si una no cabe en 240 s, el informe lo dice con su motivo (Clarificación Q1).
- **S4 · Los umbrales del cierre.** Los nombres de plan.md, «Controles de umbral», son los del cierre de H31 sin el de
  la descripción; si la decisión del job reutiliza medidas de `main` o mide otra cosa, el informe final lee los que el
  job publique.

## Fuera de alcance, anotado

Ninguna fuente nueva: la única que se graba, `boe.legislacion-consolidada`, tiene fila revisada en `main`. Nada de
`data/` cambia ni se deriva. El tope de una sesión, los umbrales, la rúbrica y los casos del juez, `jurisprudencia`,
`legal-core`, la tabla de comandos, el binario, el arnés de evals (salvo los tests que enumeran las lecturas y el tope
de dos casos de `TestTopeDeLaSesion`, tasks.md T002),
`scripts/workflow/`, `.specify/`, `internal/workflow/` y `web/` no se tocan (spec, «Fuera de alcance»).
