---

description: "Tareas de H46 · Los cuerpos de las skills de capacidad, con sitio para las materias (ADR 0063)"
---

# Tasks: H46 · Los cuerpos de las skills de capacidad, con sitio para las materias (ADR 0063)

**Input**: los artefactos de `specs/044-h46-los-cuerpos-de/`: [spec.md](spec.md), [plan.md](plan.md),
[research.md](research.md) (P1, D1-D14), [data-model.md](data-model.md), [contracts/](contracts/) y
[quickstart.md](quickstart.md) (Q1-Q9).

**Tests**: los pide el spec. Cada tarea lleva el test que la comprueba junto al código que lo pone en verde (FR-003,
FR-008, FR-009, FR-014, FR-015, FR-016), y T001 escribe la suite de aceptación de FR-021 antes que nada.

**Orden** (FR-020; research D9; plan, «Diseño»): T001 la suite; T002 la tabla de tres columnas (FR-001 a FR-004); T003
`[datos]`, los textos de `f89093dd` con lo que los fija (FR-010 a FR-016); T004 el sitio de las materias (FR-005 a
FR-009); T005 `CHANGELOG.md` (FR-018); T006 el registro de mutantes. `docs/SKILLS.md` (FR-017) va con lo que describe:
§1 en T002; §3, §5 y «La vertical de ensayo» en T004. Cada tarea termina con `make ci` en verde.

**Formato**: `- [ ] Tnnn [P?] [etiqueta?] [historia?] Descripción con todas las rutas que crea o modifica`. El
guardián de diff solo admite esas rutas (más `go.mod`, `go.sum`, `CHANGELOG.md`, el directorio del feature y el
`x_test.go` de un `x.go` declarado).

## Phase 1: Aceptación (congelada)

**Propósito**: la suite de FR-021, que cae hoy por una aserción y queda congelada (contracts/aceptacion.md; plan,
«Aceptación e2e»).

- [X] T001 [aceptacion] Escribir, a partir del spec y de [contracts/aceptacion.md](contracts/aceptacion.md), sin código de producto y sin evals (spec, «Fuera de alcance»: «Las evals, sus umbrales y el job»; research D13), los dos guiones testscript specs/044-h46-los-cuerpos-de/aceptacion/tabla-de-tres-columnas.txtar y specs/044-h46-los-cuerpos-de/aceptacion/textos-en-su-referencia.txtar, cada uno byte a byte el bloque `text` de su apartado: §1 (US1.1; la línea «# Cubre: US1.1 · FR-001, FR-021 · SC-006»; `exec kitlegal skills install --dir destino --json` y, sobre `destino/<skill>/SKILL.md` de `normativa`, `plazos-y-territorio` y `jurisprudencia`, la cabecera `| Orden | Herramienta | Qué hace |` con su separador `|---|---|---|` y una fila detrás, y las negativas acotadas entre las marcas de la tabla: ninguna cabecera de más columnas, ningún separador de cuatro, ninguna fila con cuarta celda, ni «Qué devuelve en», ni la línea del sobre, ni la de las banderas comunes) y §2 (US3.1 a US3.6; la línea «# Cubre: US3.1-US3.6 · FR-010, FR-011, FR-012, FR-021 · SC-006»; los `grep` y `! grep` de §2 en su orden sobre `SKILL.md`, `references/municipios.md`, `references/consulta_preparada.md` y `references/documento_aportado.md` instalados), con `\s+` entre las palabras de cada frase citada, `[^#]*` por apartado y `\z` para «al final» (research D8). Ninguna aserción sobre lo que el spec no fija: ni el número de tablas o filas, ni los bytes de las skills fuera de lo que nombran US1.1 y US3.1-US3.6 (US3.7 es de `TestSkillsDelProductoDesdeControl` y de quickstart Q4). **Rojo primero**: sobre los `skills/` de hoy, cada guion cae por una aserción (`tabla-de-tres-columnas.txtar:20` y `textos-en-su-referencia.txtar:25`, `no match for …`; research P1), lo que comprueba el workflow antes de congelarlos; el workflow los activa al terminar el bucle como los guiones `h46-tabla-de-tres-columnas` y `h46-textos-en-su-referencia` de `TestEntregaDelHito`, y desde entonces ninguna tarea los toca — FR-021; SC-006; US1 (escenario 1) y US3 (escenarios 1 a 6).

**Checkpoint**: la suite existe, cae hoy por una aserción y queda congelada.

---

## Phase 2: User Story 1 — La tabla de comandos lleva lo que se necesita antes de la primera llamada (Priority: P1)

**Goal**: la tabla que genera `RenderizarTabla` tiene tres columnas, sin la de lo que devuelve en `data` ni las líneas
del sobre y de las banderas comunes; lo que lee de `--describe` y sus errores no cambian.

**Independent Test**: quickstart Q2 y Q7 (`TestRenderizarTabla`, `TestDescripcionDeVerbo`, `TestRegenerarYComparar`,
`TestCuerpoDeLaSkill`, `TestOrdenesDeLasSkillsEmpotradas/control`) en verde y `make skills-check` sin defectos ni
deriva; el guion `h46-tabla-de-tres-columnas`, al activarse, en verde.

- [X] T002 [US1] La tabla de comandos de tres columnas, con sus tests, la sincronía de las tres skills de capacidad y su párrafo de la guía, según [contracts/tabla-de-comandos.md](contracts/tabla-de-comandos.md) y research D1 y D11 — US1; FR-001, FR-002, FR-003, FR-004, FR-017 (§1); SC-002. **Producción** (internal/skills/comandos.go): `columnasDeLaTabla` pasa a `| Orden | Herramienta | Qué hace |` con `|---|---|---|`; `filaDelVerbo` escribe tres celdas, con el mismo escape; `RenderizarTabla` cierra con una línea en blanco, sin la del sobre ni la de las banderas, sigue llamando a `mismoSobreYBanderas` por cada verbo con los mismos errores (§2), y su comentario dice que lo que devuelve en `data`, el sobre y las banderas comunes los dan el esquema de salida de la herramienta y `--describe` de la orden (ADR 0063); `LeerDescripcionDeVerbo` y los campos `Devuelve`, `Sobre` y `Banderas` de `DescripcionDeVerbo`, sin cambios; se retiran, por quedar sin uso (§3; `unused`), `Devuelve.texto`, `lineaDelSobre`, `lineaDeLasBanderas`, `comoCodigo`, `textoDeObjeto`, `textoDeListaDeObjetos`, `textoSinFormaDeclarada` y `valorDeUnaBandera`, y se quedan `enumerar` y `Bandera.escrita`. **Tests** (§4): en internal/skills/comandos_test.go, `TestRenderizarTabla` con `columnasDeLaTabla`, `filaDeBuscar`, `filaDeArticulo`, `filaDeVacio` y las filas de cada caso en tres celdas, `pieDeLaTabla` = `"\n"` y fuera `lineaDeLasBanderasDePrueba`, sin tocar los casos de error (`banderas-distintas`, `sobre-distinto`, sin applets, applet sin verbos) ni `TestDescripcionDeVerbo` ni `probarDefectosDeLaDescripcion`; en internal/skills/sincronia_test.go, `regionDeAlfa` con sus dos filas de tres celdas y sin las dos líneas; en internal/app/skills_test.go, el fixture de `TestOrdenesDeLasSkillsEmpotradas/control` con la cabecera de tres columnas y cada fila sin su cuarta celda, con lo que espera sin cambios (D11). **Sincronía** (FR-004): `make skills-sync` rehace la tabla de skills/normativa/SKILL.md, skills/plazos-y-territorio/SKILL.md y skills/jurisprudencia/SKILL.md, y ningún otro fichero; sus cuerpos quedan bajo 11 000 (9513, 10 117 y 10 336; D9). **Guía** (FR-017): docs/SKILLS.md §1, la viñeta de la tabla de comandos (línea 78), dice que lleva tres columnas —orden, herramienta y qué hace— y que lo que devuelve en `data`, el sobre y las banderas comunes los trae cada llamada, como texto nuevo (D12). **Verde**: quickstart Q7 (primera y tercera órdenes) y Q2 (`make skills-sync` no cambia nada y `make skills-check` sale con 0), `make lint` con 0 issues sin `//nolint` ni `t.Skip`, y `make ci`.

**Checkpoint**: tablas de tres columnas generadas y fijadas por sus tests; ningún `--describe` ni sobre distinto.

---

## Phase 3: User Story 3 y User Story 4 — Los textos de la referencia y lo que los fija (Priority: P1 y P2)

**Goal**: `skills/` es, byte a byte, el de `f89093dd`, y lo que fijaba el texto de hoy fija el nuevo, en una sola tarea
`[datos]` (FR-020), porque cada parte sin las otras deja `make ci` en rojo (D9; precedente: T002 de H41.1).

**Independent Test**: quickstart Q3, Q4, Q5, Q6 (`h41-skills-instaladas`, `h41-1-skills-instaladas`,
`h45-paquete-de-mas`) y Q7 (`TestSkillsDelProductoDesdeControl`, `TestRemisionesDelRepositorio`); el guion
`h46-textos-en-su-referencia`, al activarse, en verde.

- [X] T003 [datos] [US3] Los textos del prototipo, sus declaraciones del producto, las dos frases de las remisiones y los guiones de H41 y H41.1, en una sola tarea porque con los textos nuevos cada parte que falte deja `make ci` en rojo, y con los de hoy las partes nuevas no se cumplen (FR-020; research D5, D6, D7, D9; [contracts/textos-y-controles.md](contracts/textos-y-controles.md)) — US3 y US4; FR-010, FR-011, FR-012, FR-013, FR-014, FR-015, FR-016; SC-002, SC-004, SC-005. **Textos** (§1; FR-010 a FR-013): `git checkout f89093dd -- skills/`, un objeto del repositorio local, sin red, que cambia estos seis ficheros y ningún otro: skills/normativa/SKILL.md, skills/plazos-y-territorio/SKILL.md, skills/plazos-y-territorio/references/municipios.md, skills/jurisprudencia/SKILL.md, skills/jurisprudencia/references/consulta_preparada.md y skills/jurisprudencia/references/documento_aportado.md; después, `git diff --quiet f89093dd -- skills/` (quickstart Q4) y `make skills-sync` sin cambios (las tablas del prototipo son las que da T002; D5). Ninguna línea `⚠`, forma fija, regla invariante, orden de leer, descripción ni frontmatter sale ni cambia, y ninguna referencia nueva (FR-013), por construcción: son los bytes del prototipo. **Declaraciones del producto** (§2; FR-014): en `declaracionesDelProducto` de internal/evals/laboratorio_familias_test.go, las 15 entradas de la tabla de §2, en el orden del fichero, como `cambioDelCuerpo` o `cambioEnUnaReferencia` y no como `movimientoALaReferencia` (D6) —en `normativa` 1 del cuerpo; en `jurisprudencia` 5 del cuerpo, 2 en `consulta_preparada.md` y 2 en `documento_aportado.md`; en `plazos-y-territorio` 3 del cuerpo y 2 en `municipios.md`—, con los bytes de `antes` y `despues` de `skills/` de `main` y de `f89093dd` (la referencia de sus bytes, `git diff main f89093dd -- internal/evals/laboratorio_familias_test.go`); lo que el test comprueba no cambia. **Remisiones** (§3; FR-015): en `remisionesDelPlan` de internal/skills/remisiones_test.go, las dos frases de `main` pasan a las dos de §3, completas (`jurisprudencia`: la orden de leer `consulta_preparada.md` con «qué se pide según cómo se dio la referencia…»; `plazos-y-territorio`: la de leer `municipios.md` con «cómo se pide cada operación…»). **Guiones** (§4; FR-016): en internal/app/testdata/script/h41-skills-instaladas.txtar, la línea del `grep` de la orden de `consulta_preparada.md` con la tabla detrás pasa a las dos líneas de §4 (la orden con su frase nueva en el `SKILL.md`, y `^\| Cómo se dio \| Qué se pide \|$` en `references/consulta_preparada.md`); en internal/app/testdata/script/h41-1-skills-instaladas.txtar y en su copia congelada specs/035-h41-1-plazos-y-territorio/aceptacion/skills-instaladas.txtar, el `grep` del «Protocolo» con la orden de `main` pasa a la línea de §4, partida en sus líneas de 120 columnas, la misma en los dos ficheros, que quedan iguales byte a byte (quickstart Q5, `cmp` sin diferencias); la copia congelada de la suite de H41 no se toca (spec, «Fuera de alcance»). **Verde**: quickstart Q3 (`normativa 9498 9498`, `plazos-y-territorio 9577 9483`, `jurisprudencia 9473 9473`), Q4, Q5, Q6 (los tres guiones de H41, H41.1 y H45 en `--- PASS`), Q7 (`TestSkillsDelProductoDesdeControl` y `TestRemisionesDelRepositorio` en `ok`) y `make ci`.

**Checkpoint**: lo que el agente carga es el texto del prototipo, y los controles que fijaban el de hoy fijan el nuevo.

---

## Phase 4: User Story 2 — El sitio de las materias lo comprueba `make ci` (Priority: P1)

**Goal**: un `SKILL.md` con el par de marcas de los paquetes cuyo cuerpo, sin las líneas entre ellas, pasa de 9600
caracteres da un defecto de `skills-check`; sin las marcas, ninguno nuevo.

**Independent Test**: quickstart Q7 (`TestSkillsDelRepositorio/presupuestos`, `TestVerticalDeEnsayo`) y Q6
(`h45-paquete-de-mas`) en verde, `make skills-check` sin defectos con los textos de T003 (FR-007), y los mutantes de
SC-003 en rojo (T006).

- [X] T004 [US2] El sitio de las materias en `defectosDelCuerpo`, con sus casos en `casosDePresupuestos`, `TestVerticalDeEnsayo` sin la línea de más y su parte de la guía, según [contracts/sitio-de-las-materias.md](contracts/sitio-de-las-materias.md), data-model.md y research D2, D3, D4 y D12 (va detrás de T003: con los cuerpos de hoy, `plazos-y-territorio` y `jurisprudencia` pasarían de 9600; D9) — US2; FR-005, FR-006, FR-007, FR-008, FR-009, FR-017 (§3, §5 y «La vertical de ensayo»); SC-001, SC-002, SC-003. **Producción** (internal/skills/sincronia.go): la constante `maximoSinLosPaquetes = 9600` junto a `maximoDeLineasDelCuerpo` y `maximoDeCaracteresDelCuerpo`; en `defectosDelCuerpo`, detrás del defecto de 11 000, si `regionDeLosPaquetes.buscar(cuerpo)` da la región sin error, `n := utf8.RuneCount(cuerpo) - utf8.RuneCount(cuerpo[region.desde:region.hasta])` y, con `n > maximoSinLosPaquetes`, el defecto `<skill>: SKILL.md: cuerpo sin las líneas de los paquetes de <n> caracteres (máximo 9600)` (§1: las marcas, la tabla y todo lo demás del cuerpo cuentan); sin las marcas, o con unas que no van en pareja, ningún defecto nuevo (FR-006; el de las marcas ya sale de `paquetes.regenerar`); los comentarios de `defectosDelCuerpo` y de `Regenerar` dicen la regla; ningún import nuevo. **Test del umbral** (FR-008; §2): en internal/app/skills_test.go, las dos marcas como constantes del test junto a `inicioDeLaTablaDeComandos`, y `casosDePresupuestos` lee el `SKILL.md` de cada skill en el árbol (`rutaEnLaSkill(raizDelRepositorio, …)`), busca la línea de la marca de inicio y la de fin y cuenta en runas `p`, lo que hay entre ellas; con las marcas, el cuerpo ensanchado a `9600 + p` sin defectos, a `9601 + p` con solo el de FR-005 (9601) y a 11 001 con el de 11 000 y, detrás, el de FR-005 con `11001 − p`, comparados con `assert.Equal` en su orden; sin ellas, 11 000 sin defectos y 11 001 con el suyo, como hoy; 499 y 500 líneas, como hoy; `ensancharElCuerpo` sigue añadiendo su línea fuera de las marcas. **FR-009** (§3): en internal/skills/vertical_de_ensayo_test.go, fuera `comprobarElSitioDeOtroPaquete`, su llamada, el import `unicode/utf8` y la frase del comentario que la describe, que pasa a decir que el sitio de los demás paquetes lo comprueba el presupuesto del cuerpo sin sus líneas; lo demás del test, igual. **Guía** (FR-017; D12): en docs/SKILLS.md, §3 gana la fila «Caracteres del cuerpo sin las líneas de los paquetes, en una skill con sus marcas | Como mucho 9600» con su porqué (1400 para las líneas, unas veinte materias de nombre corto a 52 + 3n) y las cifras de los tres cuerpos al cerrar medidas como el gate con quickstart Q3 (9498, 9483 y 9473 sin las líneas; 9498, 9577 y 9473 con ellas); §5, la fila del defecto con su mensaje y los 9600 en «En los límites, en verde»; «La vertical de ensayo», sin «sitio para otra línea de paquete de su misma longitud». **Verde**: quickstart Q7 (`TestSkillsDelRepositorio/presupuestos`, `TestVerticalDeEnsayo`), Q6 (`h45-paquete-de-mas`, cuyo paso de sincronía corre el subtest en su copia con el paquete de prueba en `normativa`; SC-001), `make skills-check` sin defectos (FR-007), `make lint` con 0 issues sin `//nolint` ni `t.Skip`, y `make ci`.

**Checkpoint**: el presupuesto de 9600 sin las líneas de los paquetes lo comprueba `make ci`, y lo fijan sus casos en los límites.

---

## Phase 5: Cierre del hito (registro y comprobaciones sin modelo)

**Propósito**: FR-018 y el registro de SC-003 a SC-006 (research D10, D12). US5 (las respuestas no cambian, SC-008) la
deciden las evals del cierre, que lanza el workflow tras la revisión final y no una tarea (D13; plan, «Autonomía»).

- [ ] T005 [P] La entrada del hito en CHANGELOG.md, bajo *Unreleased* → «Cambiado», como la de H41.1 (`CHANGELOG.md:246`; research D12) — FR-018; US1 a US4. Una entrada (H46; ADR 0063) que nombra `normativa`, `plazos-y-territorio` y `jurisprudencia` y lo que cambia en cada una (la tabla de comandos de tres columnas, sin lo que devuelve en `data` ni el sobre ni las banderas comunes, en las tres; en `normativa`, «o con su herramienta»; en `plazos-y-territorio`, cómo se presenta cada dato y la herramienta u orden en `municipios.md`, con las tres líneas de la cobertura en el cuerpo; en `jurisprudencia`, lo que se pide en `consulta_preparada.md` y la referencia pedida y las sentencias que menciona en `documento_aportado.md`), y la regla nueva de `skills-check`: un cuerpo con las marcas de los paquetes no pasa de 9600 caracteres sin sus líneas (1400 para las materias); ningún applet devuelve otra cosa (FR-019). **Verde**: `grep -n 'H46' CHANGELOG.md` da la entrada bajo *Unreleased*, y `make ci`.
- [ ] T006 [P] El registro de mutantes en specs/044-h46-los-cuerpos-de/mutantes.md, con cada mutante corrido en un clon del temporal (`git clone -q --no-hardlinks . "$TMPDIR/h46-mut"`, donde `main` es `origin/main`; research D10, V14) y nunca en el árbol de trabajo, según quickstart Q9 y la tabla «Mutantes» de plan.md — SC-003, SC-004, SC-005, SC-006; FR-001, FR-002, FR-003, FR-005, FR-006, FR-014, FR-015, FR-016. Por cada fila, la orden, lo esperado y lo visto, tal cual salga: `maximoSinLosPaquetes = 9601` (6 casos de `presupuestos` en rojo) y `= 9599` (9); la regla que no se aplica, con el `sincronia.go` de `main` y los tests del hito (6); sin restar las líneas de los paquetes (3 casos de `presupuestos/plazos-y-territorio`); el defecto de FR-005 delante del de 11 000 (los `once-mil-y-un-caracteres` de las skills con marcas); la regla también sin las marcas (`presupuestos/consulta-juridica`: `once-mil-caracteres` y `once-mil-y-un-caracteres`); la tabla de cuatro columnas, con el `comandos.go` de `main` (`TestRenderizarTabla`, `TestRegenerarYComparar`, `TestCuerpoDeLaSkill`); `RenderizarTabla` sin la llamada a `mismoSobreYBanderas` (`TestRenderizarTabla`, `banderas-distintas` y `sobre-distinto`); cada una de las 15 entradas de FR-014 deshecha en su fichero de skill, de una en una (`TestSkillsDelProductoDesdeControl`, 15 rojos); la frase de `main` en `remisionesDelPlan`, en cada una de las dos por separado (`TestRemisionesDelRepositorio`, 2 rojos); la orden de `main` en `h41-1-skills-instaladas.txtar` y la línea de `main` en `h41-skills-instaladas.txtar` (cada guion en `--- FAIL`); y los dos guiones de la suite sobre los ficheros de skill de `main`, copiados en el clon como los guiones `h46-tabla-de-tres-columnas` y `h46-textos-en-su-referencia` de `TestEntregaDelHito`, porque el workflow aún no los ha activado (los dos en `--- FAIL`). Añade, corridos en el árbol de trabajo sin cambiarlo, quickstart Q3 (las cifras de T004), Q4, Q5 y Q8 (en este punto, la primera orden da once rutas: las trece de su lista menos las dos `h46-` que añade la activación; la segunda orden, solo `comandos.go` y `sincronia.go` de `internal/skills`; la tercera, nada: ninguna eval, esquema ni dato cambia, SC-007, SC-008), y el `rm -rf` del clon al final. Un mutante que no sale en rojo donde dice el plan se registra con lo visto, sin ajustar lo esperado. **Verde**: el fichero tiene una fila por mutante de la tabla del plan (con las 15 de FR-014 y las 2 de FR-015 por separado), el árbol de trabajo solo cambia en specs/044-h46-los-cuerpos-de/mutantes.md, y `make ci`.

---

## Controles de umbral → tarea que los construye (ADR 0029)

| Fila de plan.md «Controles de umbral» | Tarea | Test que lo ve fallar pasado el umbral |
|---|---|---|
| SC-001, FR-008 (`make ci`, con `h45-paquete-de-mas`) | T004 (y cada tarea lo deja en verde) | `casosDePresupuestos` en el paso de sincronía de `h45-paquete-de-mas`; mutantes de T006 |
| SC-002, FR-005, FR-007 (0 cuerpos con marcas > 9600 sin sus líneas; 0 > 11 000) | T004 | `TestSkillsDelRepositorio/presupuestos`; mutantes 9601, 9599 y la regla desactivada (T006) |
| SC-002, FR-004 (0 ficheros que `skills-sync` cambiaría) | T002 (y T003, sin deriva) | la deriva de `TestSkillsDelRepositorio`; mutante de la tabla de cuatro columnas (T006) |
| SC-003, FR-005, FR-006, FR-008 (9600/9601/11 001 con marcas; 11 000/11 001 sin) | T004 | `casosDePresupuestos` con `assert.Equal`; mutantes de T006 |
| FR-001, FR-003 (3 columnas, 0 líneas del sobre o banderas, 0 bytes de diferencia) | T002 | `TestRenderizarTabla`, `TestRegenerarYComparar`, `TestCuerpoDeLaSkill`; mutante con `comandos.go` de `main` (T006) |
| FR-002 (0 cambios en errores y defectos) | T002 | casos de error de `TestRenderizarTabla` y `TestDescripcionDeVerbo`, sin tocar; mutante sin `mismoSobreYBanderas` (T006) |
| SC-004, FR-010, FR-013, FR-014 (0 bytes de diferencia con el control) | T003 | `TestSkillsDelProductoDesdeControl`; 15 mutantes (T006) |
| SC-005, FR-015 (las 2 frases nuevas) | T003 | `TestRemisionesDelRepositorio`; 2 mutantes (T006) |
| SC-006, FR-001, FR-010, FR-011, FR-012, FR-021 (100 % de los guiones `h46-`) | T001 | los dos guiones en `TestEntregaDelHito`, con su rojo primero; mutante sobre los ficheros de skill de `main` (T006) |
| FR-016 (100 % de `h41-skills-instaladas` y `h41-1-skills-instaladas`) | T003 | los dos guiones en `TestEntregaDelHito`; mutantes con la línea de `main` (T006) |
| SC-007, FR-019 (0 cambios en `--describe`, esquema, sobre, códigos o `data`) | ninguna tarea toca un applet ni sus tests; T006 registra Q8 | los tests de cada applet y `make schema-check`, sin cambios |
| SC-008 (evals de `normativa`, `plazos-y-territorio`, `jurisprudencia`, `consulta-juridica`) | el job de evals del cierre, que lanza el workflow, sin cambios (D13); T003 es lo que mide, y T006 registra que ninguna eval cambia | el veredicto del job, en fallo con una clase o serie que decide bajo su umbral |

## Definition of Done (docs/ROADMAP.md §1) aplicable

- **`make ci` en verde** (1): cada tarea; SC-001.
- **Tests offline del código nuevo** (2): T002 y T004; ninguna fuente nueva ni grabación (plan, «Datos externos»: no hay
  manifiesto `grabaciones.json` ni test `TestGrabar*`).
- **e2e actualizado y `CHANGELOG.md`** (6): T001 (la suite), T003 (los guiones de H41 y H41.1), T005 (`CHANGELOG.md`).
- **ADR** (7): no aplica; el ADR 0063 ya está en `main` y el spec deja fuera «Cambiar cualquier ADR o escribir otro».
- **`docs/SOURCES.md`** (8): no aplica; ninguna fuente externa.
- **Esquemas** (4): no aplica; ningún `--describe` ni esquema cambia (FR-019), y `make schema-check` lo comprueba en
  `make ci`.
- **Fixtures** (2, 6): T003, la única tarea `[datos]`, con los guiones de H41, H41.1 y su copia congelada.
- **Evals de la skill** (10): ninguna cambia (spec, «Fuera de alcance»; D13); las mide todas el cierre (SC-008).
- 3, 5, 9, 11, 12 y 13: sin código que los toque; `TestArquitectura` y la configuración de lint, sin cambios.

## Dependencies & Execution Order

- T001 → T002 → T003 → T004 → {T005, T006}. Cada una depende solo de las anteriores:
  - T002 antes de T003: las tablas del prototipo son las que da el generador de tres columnas (D5); sin él, deriva.
  - T003 antes de T004: con los cuerpos de `main`, `plazos-y-territorio` y `jurisprudencia` pasarían de 9600 (D9).
  - T005 describe el hito entero y T006 mide su estado final: van detrás de T004, y no se tocan entre sí (`[P]`).
- Historias: US1 (T002) es independiente; US3 y US4 (T003) parten de US1; US2 (T004), de US3; US5 la decide el cierre.

## Parallel Example

```text
# Tras T004, en cualquier orden:
T005 CHANGELOG.md
T006 specs/044-h46-los-cuerpos-de/mutantes.md
```

Lo demás va en secuencia: cada tarea toca ficheros de la anterior (`skills/`, `internal/app/skills_test.go`,
`docs/SKILLS.md`) o necesita su estado.

## Implementation Strategy

- **MVP**: T001 y T002 (US1): la tabla de tres columnas libera 1237, 569 y 506 caracteres de los tres cuerpos sin tocar
  su texto.
- **Incremento 2**: T003 (US3, US4): los textos del prototipo dejan los cuerpos en 9498, 9577 y 9473.
- **Incremento 3**: T004 (US2): `make ci` comprueba el sitio de las materias.
- **Cierre**: T005 y T006; las evals del cierre (US5) y la propuesta las hace el workflow.
