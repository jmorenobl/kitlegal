---

description: "Tareas de H31 · Las skills siguen la especificación Agent Skills: guía de autoría y gates"
---

# Tasks: H31 · Las skills siguen la especificación Agent Skills: guía de autoría y gates

**Input**: `specs/027-h31-las-skills-siguen/` — spec.md, plan.md, research.md (D1–D21, V1–V12 y la tabla de reglas de
FR-021), data-model.md, contracts/ (`gates-de-skills-check.md`, `controles-de-contenido.md`,
`juez-de-las-descripciones.md`, `skills-reescritas.md`) y quickstart.md.

**Aceptación e2e: no aplica** (plan.md, «Aceptación e2e» y Complexity Tracking; research D21): no hay tarea de
aceptación ni guiones `testscript`. El hito no cambia ningún applet, verbo, bandera, código de salida ni salida del
binario; la aceptación es `make ci` en verde con los gates sobre las tres skills (SC-001) y el job de evals sobre la
cabeza, que mide las 45 evals de hoy en los dos modos y juzga las tres descripciones (SC-010), que mide el cierre del
workflow. El spec deja fuera evals nuevas («Fuera de alcance»), así que ninguna tarea escribe en `evals/<skill>/`.

**Tests**: cada tarea lleva su test y la implementación mínima que lo hace pasar, y deja `make ci` en verde al terminar.
Todo offline y sin modelo (ADR 0032): mutantes sobre copias en `t.TempDir()`, votos simulados con un `Votante` de prueba
y el `claude` sustituto de `internal/evals/sustitutos_test.go`. Ningún fixture nuevo en `testdata/` ni en `schemas/`
(plan, «Fixtures»), ningún dato externo, ningún manifiesto de grabación ni fila de `docs/SOURCES.md` (plan, «Datos
externos»), y ningún fichero de `data/` cambia. La rúbrica, el esquema, los casos, la medida y el manifiesto del juez
de las descripciones son los del directorio de evidencia del ADR 0046, en `main` desde la entrada: las tareas solo los
leen.

**Organización**: las fases siguen el «Orden de implementación (de dentro afuera)» de plan.md, que es el único orden en
que cada paso deja `make ci` en verde: la regla de `references/` (US4) antes de que exista ninguna referencia a mano;
los controles de contenido sobre la skill entera (US5) antes de mover texto a referencias; la reescritura (US1); el
control de las órdenes de ejemplo (US5) cuando los ejemplos que la caché no sirve ya han cambiado; los gates de la
forma (US2) cuando las tres skills ya los cumplen; el juez de las descripciones (US3); y la guía (US6).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: puede ir en paralelo (ficheros distintos, sin dependencias de tareas pendientes).
- **[Story]**: la historia de usuario del spec (US1…US6).
- Cada tarea nombra todas las rutas que crea o modifica; declarar `x.go` permite `x_test.go`. `go.mod`, `go.sum`,
  `CHANGELOG.md` y `specs/027-h31-las-skills-siguen/` están siempre permitidos.

---

## Phase 1: Setup

Sin tareas: el módulo, las herramientas, los objetivos de `make` y los paquetes `internal/skills` e `internal/evals` ya
existen; el hito no añade objetivos de `make` ni dependencias (plan, «Objetivos de make: ninguno nuevo»; «Primary
Dependencies»).

---

## Phase 2: US4 — La regla de `references/` (fundacional, bloquea a US1, US5 y US3) (Priority: P2)

**Goal**: un `.md` de `references/` no declarado y sin la cabecera «generado desde» es una referencia a mano, que
`make skills-sync` no toca y que el conjunto de dependencias de cada eval cuenta en su parte (b).

**Independent Test**: `go test -count=1 ./internal/skills/`, `go test -count=1 -run '^TestSkillsDelRepositorio$'
./internal/app/` y `go test -count=1 -run '^(TestConjuntoDeCadaEval|TestEleccionDeLosMutantes)$' ./internal/evals/`
(quickstart, escenarios 2, 3 y 5).

- [X] T001 [US4] Implementar la clasificación de FR-012 (contracts/gates-de-skills-check.md §3; research D13): nueva función exportada `ReferenciasAMano(arbol fs.FS, skill string) ([]Referencia, error)` en `internal/skills/referencias.go`, que da los ficheros regulares `.md` del directorio `references` de la skill que `kitlegal-referencias` no declara y cuya primera línea no es la cabecera «generado desde» (`DatosDeLaReferencia`); en `internal/skills/sincronia.go`, `sobrantes` deja de dar `fichero-sobrante` a la referencia a mano (ni deriva ni retirada), el no declarado que abre con la cabecera sigue siendo `fichero-sobrante` y el declarado sigue dando `contenido-distinto` o `fichero-ausente` si no es lo que genera su fila, con o sin cabecera; la entrada que no es un `.md` regular sigue como hoy hasta T011. Tests: `TestReferenciasAMano` en `internal/skills/referencias_test.go` (las cinco clases de la tabla del §3, sobre un `fstest.MapFS`), `TestRegenerarYComparar` en `internal/skills/sincronia_test.go` (la clasificación en el árbol sintético) y, en `TestSkillsDelRepositorio` de `internal/app/skills_test.go`, el subtest `deriva` con las clases de FR-012 (declarada distinta con y sin cabecera, no declarada con cabecera) y el nuevo `sincronia-con-referencias-a-mano` (SC-003: sobre `copiaConLaSkill`, una referencia a mano, una generada con un byte cambiado y otra sin su cabecera; tras `skills.Regenerar` y `skills.Escribir`, la a mano idéntica byte a byte y las dos generadas rehechas; en rojo si toca la primera o no rehace una de las otras). Ajustar `internal/skills/export_test.go` si un test necesita algo no exportado. Con las skills de `main`, sin referencias a mano, nada cambia. FR-012, FR-013, SC-003. Rutas: internal/skills/referencias.go, internal/skills/referencias_test.go, internal/skills/sincronia.go, internal/skills/sincronia_test.go, internal/skills/export_test.go, internal/app/skills_test.go
- [X] T002 [US4] Contar las referencias a mano en el conjunto de dependencias de cada eval (FR-014, Clarificación Q1; research D15): en `internal/evals/dependencias.go`, `datosDeLasReferencias` deja fuera de la parte (c) las referencias a mano de la skill (`skills.ReferenciasAMano`), sin error (la que no abre con la cabecera y no es a mano sigue siendo error), y `skillsDeLaEval` busca con `nombra` los nombres de otras skills en el `SKILL.md` y en cada referencia a mano de cada skill que visita, sin mezclar sus bytes con los que lee `AppletsDeLaTabla`; las referencias generadas no se leen para nombres; actualizar la descripción del conjunto en `internal/evals/doc.go`. Tests: `TestConjuntoDeCadaEval` en `internal/evals/dependencias_test.go` (con el árbol del repositorio y una referencia a mano superpuesta: entra en (b), no en (c), sin error; una skill nombrada solo en una referencia a mano entra en (b) con su `SKILL.md` y sus `references/` y el cierre sigue desde ella; en rojo si no la incluye o da error) y `TestEleccionDeLosMutantes` en `internal/evals/eleccion_test.go` (mutante nuevo `referencia-a-mano-de-legal-core`: 0 evals de diferencia con las de `legal-core` y las de las skills que la nombran; y el mutante de la descripción de hoy sigue eligiendo las 45 evals con 0 reutilizadas, en rojo si difiere en una sola). Con las skills de `main` la elección no cambia. FR-014, FR-030, SC-004, SC-010. Rutas: internal/evals/dependencias.go, internal/evals/dependencias_test.go, internal/evals/eleccion_test.go, internal/evals/doc.go

**Checkpoint**: una referencia a mano ya puede existir en `references/` sin que `skills-sync` la retire ni el arnés de
evals dé error.

---

## Phase 3: US5 — Los controles de contenido leen la skill entera (Priority: P2)

**Goal**: mover un texto a una referencia a mano no lo saca de ningún control de contenido de `make ci`.

**Independent Test**: `go test -count=1 -run '^(TestEvalsDelRepositorio|TestControlesSobreLasReferenciasAMano)$'
./internal/evals/` y `go test -count=1 -run '^TestSkillsDelRepositorio$' ./internal/app/` (quickstart, escenario 4).

- [X] T003 [US5] Llevar los controles 1 a 7 de contracts/controles-de-contenido.md a la skill entera (FR-016 a FR-018; research D16) en `TestEvalsDelRepositorio` de `internal/evals/conjunto_test.go`, leyendo el `SKILL.md` y cada referencia a mano (`skills.ReferenciasAMano`) por separado y con su ruta en cada defecto, en las mismas skills que hoy: `prosa-de-la-skill` (`defectosDeLaProsa` sobre cada fichero: ninguna expresión de `expresiones-prohibidas.yaml`, el de las evals de `boe-legislacion`, fuera del código ni fecha AAAAMMDD con cifras; la prosa no vacía, al `SKILL.md`); `ordenes-para-powershell` (cada orden de lectura o de comprobación de cada fichero lleva detrás, en ese mismo fichero, su forma para PowerShell; una orden de Bash por verbo de lectura, `articulo` y `articulos`, en la skill entera); `linea-sin-consulta`, `formas-de-la-redaccion-fechada`, `avisos-de-la-skill` y `hallazgos-de-la-skill` siguen exigiendo su forma en el `SKILL.md`; y el subtest nuevo `lineas-de-aviso-de-las-referencias` (`boe-legislacion`, `legal-core`: cada línea de una referencia a mano que empieza, tras blancos, marcador de lista y énfasis, por `⚠` es, sin los blancos de los extremos, una línea del `SKILL.md`, con el defecto `<skill>: references/<f>.md:<n>: la línea ⚠ no es ninguna de SKILL.md: «…»`). Test nuevo `TestControlesSobreLasReferenciasAMano` en el mismo fichero: copia la skill a `t.TempDir()` y, por cada control 1 a 7, mueve su regla del `SKILL.md` a una referencia a mano de la copia o la escribe en ella (expresión prohibida, fecha AAAAMMDD, orden sin su forma para PowerShell, línea sin consulta movida, cada forma de aviso, de hallazgo y de la redacción fechada movida, una línea `⚠` mal escrita) y exige el defecto; en rojo si alguno no se detecta. Las tres skills de `main` siguen sin defectos. FR-016, FR-017, FR-018, SC-005. Rutas: internal/evals/conjunto_test.go, internal/evals/conjunto.go
- [X] T004 [US5] Llevar los controles 8 y 9 de contracts/controles-de-contenido.md a las referencias a mano (FR-016, FR-017; research D16) en `TestSkillsDelRepositorio` de `internal/app/skills_test.go`: `normas-nombradas` comprueba cada «Ley N/AAAA», «LO…» y «RD…» del `SKILL.md` y de cada referencia a mano (`skills.ReferenciasAMano`) contra los títulos de `normas.yaml`, el registro de normas, con el fichero en el defecto; `sin-instrucciones-de-evals` sigue leyendo el `SKILL.md` sin la región y todo `references/*.md`. Mutantes sobre `copiaConLaSkill`: una norma que no está en `normas.yaml` escrita en una referencia a mano y una instrucción de las evals escrita en una referencia a mano, cada uno con su defecto; en rojo si alguno no se detecta. FR-016, FR-017, SC-005. Rutas: internal/app/skills_test.go

**Checkpoint**: la reescritura puede mover texto a referencias sin apagar ningún control.

---

## Phase 4: US1 — Las tres skills reescritas (Priority: P1) 🎯 MVP

**Goal**: cada skill carga un cuerpo corto con lo que toda respuesta necesita y lee el detalle en referencias a mano
cuando la clase de pregunta lo pide, con su descripción en tercera persona; ninguna regla de `main` se pierde ni se
añade.

**Independent Test**: `make skills-check`, `go test -count=1 -run '^TestEvalsDelRepositorio$' ./internal/evals/`,
`go test -count=1 -run '^TestSkillsEmpotradas$' .` y las medidas del escenario 1 de quickstart (menos de 500 líneas y
como mucho 11 000 caracteres en el cuerpo, como mucho 900 en la descripción); en el cierre, el job de evals (SC-010).

Reglas comunes a las tres tareas (contracts/skills-reescritas.md, «Reglas comunes a las tres»): el reparto es el del
contrato y cada regla va donde dice la tabla de research.md, «Tabla de reglas (FR-021)», sin quitar ni añadir ninguna;
el `name`, `kitlegal-applets`, `kitlegal-referencias`, la región generada de la tabla de comandos y el contenido de las
referencias generadas no cambian (FR-024), y `make skills-sync` deja el árbol sin cambios; las referencias a mano van sin
cabecera ni declaración, abren con un párrafo que dice lo que llevan, no enlazan otra referencia, no repiten lo que sale
de `data/` (FR-013), no llevan ninguna línea `⚠` y tienen menos de 100 líneas; el cuerpo enlaza cada referencia, a mano
o generada, con un enlace de Markdown `[…](references/<fichero>.md)` fuera de los bloques de código y dice cuándo
leerla; la descripción va en tercera persona, sin `<` ni `>`, con como mucho 900 caracteres; y cada tarea mide el cuerpo
con la orden del escenario 1 de quickstart mientras escribe. `TestSkillsEmpotradas` (`skills_test.go` de la raíz, sin
cambios) comprueba que cada referencia nueva va en el binario con sus bytes y falla si falta una (FR-015, research D20).

- [X] T005 [US1] Reescribir `jurisprudencia` según contracts/skills-reescritas.md, «jurisprudencia»: el cuerpo de `skills/jurisprudencia/SKILL.md` con los apartados de su tabla (entrada condensada con J06–J10; J15 en el párrafo de «Comandos»; paso 2 con J20, J22, J23, J25 y J27 y la línea que manda leer la consulta preparada; paso 3 con J31–J33, J35 y J40 y la línea que manda leer el documento aportado; paso 6 sin el bloque de la orden ni J57; pasos 1, 4, 5 y 7, «Reglas» y la región generada iguales; `boe-legislacion` sigue nombrada, J70) y las dos referencias a mano nuevas `skills/jurisprudencia/references/consulta_preparada.md` (La orden de cada forma · Las casillas · El equivalente · La búsqueda por materia) y `skills/jurisprudencia/references/documento_aportado.md` (La ficha y nada más · Lo que devuelve · El documento que no es el pedido · El ROJ y el ECLI que no se corresponden · Un texto sin ficha reconocible); la línea `⚠ SENTENCIA NO COMPROBADA:` solo en el cuerpo; todas las reglas del CENDOJ (J04, J06–J08, J16, J48, J67) en el cuerpo o en la descripción; descripción en tercera persona de como mucho 900 caracteres que cubra lo que hoy dice la de `main`. Entrada en `CHANGELOG.md`: `jurisprudencia` v0.2 (de v0.1.2), con lo que cambia (research D19). Comprobar con `make skills-check`, `make skills-sync` sin cambios en el árbol, `TestEvalsDelRepositorio`, `TestSkillsEmpotradas` y `make ci`. FR-013, FR-015, FR-020, FR-021, FR-022, FR-023, FR-024, SC-009, SC-012. Rutas: skills/jurisprudencia/SKILL.md, skills/jurisprudencia/references/consulta_preparada.md, skills/jurisprudencia/references/documento_aportado.md, CHANGELOG.md
- [X] T006 [US1] Reescribir `legal-core` según contracts/skills-reescritas.md, «legal-core»: el cuerpo de `skills/legal-core/SKILL.md` con los apartados de su tabla (título y frase; `## Referencias` con una línea y un enlace por cada una de las seis referencias, que son el único enlace de cada una; `## Protocolo` con los pasos 1 a 11 y todo lo que sus evals comprueban sin modelo, que el contrato enumera en «Riesgos: lo que tiene que quedarse en el cuerpo»; `## Cómo se presenta el territorio` con las tres líneas de cobertura; `## Cómo se presenta un plazo` con las nueve formas `⚠` byte a byte; `## Comandos` con su párrafo y la región generada; `## Reglas`, las nueve, con la línea `⚠ SIN CONSULTA AL BOE: …` sola en su bloque `text` en la regla 7) y las dos referencias a mano nuevas `skills/legal-core/references/municipios.md` (El municipio que falta · Lo que trae `data` · La plantilla del territorio · Qué dice cada aspecto de la cobertura · Cuando no sale con 0) y `skills/legal-core/references/calculo_de_plazos.md` (Lo que la ley regula aparte · El acto presunto · El silencio · La reclamación del art. 24.2 LTAIBG · Los municipios que cuentan · El plazo sin calcular · La lectura en la fecha de inicio), sin ninguna norma escrita como «Ley N/AAAA» ni orden con un identificador real; los cambios de ejemplos y textos de «Notas» del contrato; descripción en tercera persona de como mucho 900 caracteres (el borrador del contrato o uno equivalente). Las cuatro referencias generadas (`leyes_vertebrales.md`, `jerarquia_normativa.md`, `cobertura_por_materia.md`, `recursos_y_plazos.md`) no cambian. Entrada en `CHANGELOG.md`: `legal-core` v1.3 (de v1.2.3). Comprobar con `make skills-check`, `make skills-sync` sin cambios en el árbol, `TestEvalsDelRepositorio`, `TestSkillsEmpotradas` y `make ci`. FR-013, FR-015, FR-020, FR-021, FR-022, FR-023, FR-024, SC-009, SC-012. Redelimitada en el intento 1: con las dos referencias a mano, el mutante `referencia-de-legal-core` (y el de la referencia renombrada) de `TestMatrizDeLasPropuestas` toma el primer fichero de Markdown de las referencias de legal-core por orden de nombre, que pasa a ser una a mano, y el golden de la matriz cambia de ruta; `mutantesDeLaEleccion` elige en su lugar la primera referencia generada (`referenciaGenerada`, con `skills.DatosDeLaReferencia`), como en H30, y el golden no cambia. Rutas: skills/legal-core/SKILL.md, skills/legal-core/references/municipios.md, skills/legal-core/references/calculo_de_plazos.md, CHANGELOG.md, internal/evals/eleccion_test.go
- [X] T007 [US1] Reescribir `boe-legislacion` según contracts/skills-reescritas.md, «boe-legislacion»: el cuerpo de `skills/boe-legislacion/SKILL.md` con los apartados de su tabla (Protocolo; pasos 1 a 5; `## Cómo se cita`; `## Redacción en una fecha` con las tres formas fechadas juntas en un bloque `text`; `## Redacción modificada` con `⚠ REDACCIÓN MODIFICADA:`; `## Comandos` y la región generada; `## Reglas`, las ocho con su número, la 3 con las tres etiquetas de aviso, la 7 sin su frase y la 8 con la línea sin consulta sola en su bloque `text`; las órdenes `kitlegal boe indice BOE-A-2015-10565` y la de `a21` con su forma para PowerShell en el cuerpo; el enlace a la referencia generada `normas.md` desde el paso 1) y las siete referencias a mano nuevas `skills/boe-legislacion/references/con_herramientas.md`, `buscar_la_norma.md`, `varios_bloques.md` (con la orden de Bash de `articulos` y su forma para PowerShell), `lecturas_que_fallan.md` (con la frase de la regla 7 sola en un bloque `text`), `como_se_cita.md`, `redaccion_en_una_fecha.md` (con `kitlegal boe articulo BOE-A-1985-5392 a22 --fecha 2005-06-15 --json`) y `redaccion_modificada.md`, con los apartados y el contenido (B…) del contrato; los «Ejemplos que cambian» del contrato, entre ellos `kitlegal boe buscar <términos> --json` en lugar del ejemplo de `buscar` y la retirada del ejemplo de `analisis`, que la caché de una sesión no sirve (research D17); la prosa de cuerpo y referencias sin ninguna expresión de `expresiones-prohibidas.yaml`, el de sus evals, ni fecha AAAAMMDD con cifras (contrato, «Riesgos»); descripción en tercera persona de como mucho 900 caracteres (la «Descripción propuesta» del contrato o una equivalente). `normas.md` no cambia. Entrada en `CHANGELOG.md`: `boe-legislacion` v0.3 (de v0.2.2). Comprobar con `make skills-check`, `make skills-sync` sin cambios en el árbol, `TestEvalsDelRepositorio`, `TestOrdenesParaPowerShell`, `TestSkillsEmpotradas` y `make ci`. Redelimitada en el intento 1: con las órdenes de `articulos` en `varios_bloques.md`, los mutantes `ordenes-de-articulos-en-una-referencia` y `sin-las-ordenes-de-articulos` de `TestControlesSobreLasReferenciasAMano` quitan de `SKILL.md` unas líneas que ya no lleva y salen en rojo; en `internal/evals/conjunto_test.go`, el primero pasa a llevarlas de `varios_bloques.md` al final de `SKILL.md` (`ordenes-de-articulos-en-skill-md`, sin defecto) y el segundo a quitarlas de `varios_bloques.md`, con el defecto de hoy (`quitarDelFichero`, `sinLasLineas` con el fichero, `conLasLineasLlevadasAlSkillMd`; `conLasLineasMovidas` sin el parámetro que ya solo recibe verdadero; el diff, en gates/tarea-T007.md). FR-013, FR-015, FR-019, FR-020, FR-021, FR-022, FR-023, FR-024, SC-009, SC-012. Rutas: skills/boe-legislacion/SKILL.md, skills/boe-legislacion/references/con_herramientas.md, skills/boe-legislacion/references/buscar_la_norma.md, skills/boe-legislacion/references/varios_bloques.md, skills/boe-legislacion/references/lecturas_que_fallan.md, skills/boe-legislacion/references/como_se_cita.md, skills/boe-legislacion/references/redaccion_en_una_fecha.md, skills/boe-legislacion/references/redaccion_modificada.md, CHANGELOG.md, internal/evals/conjunto_test.go

**Checkpoint**: las tres skills tienen su forma nueva y `make ci` sigue en verde con los controles de US5.

---

## Phase 5: US5 — Las órdenes de ejemplo que leen del BOE (Priority: P2)

**Goal**: toda orden de ejemplo concreta de `kitlegal boe` del cuerpo o de una referencia a mano es de lo que la caché
de una sesión de evals sirve sin red.

**Independent Test**: `go test -count=1 -run '^(TestEvalsDelRepositorio|TestOrdenesDeEjemplo|TestControlesSobreLasReferenciasAMano)$' ./internal/evals/` (quickstart, escenario 4).

- [X] T008 [US5] Construir el control 10 de contracts/controles-de-contenido.md (FR-019; research D17) en `internal/evals/conjunto_test.go`: el extractor `ordenesDeEjemplo`, junto a `ordenesDeLectura`, que de cada tramo de código que ya recorre `ordenesDeLectura` toma cada aparición de `kitlegal boe` con lo que la sigue hasta el primer `&&`, `;` o `|`, sin `--json`, y la da como orden concreta si lleva un verbo y al menos un argumento y ningún marcador `<…>`, `...` ni `…`; `TestOrdenesDeEjemplo` sobre Markdown escrito en el test (la orden de un bloque en su forma de Bash y en la de PowerShell, la de código en línea, cada una con su línea y su `--fecha <día>`; ninguna de la que lleva un marcador, de la mención del verbo sin argumentos ni de la prosa); el subtest `ordenes-de-ejemplo` de `TestEvalsDelRepositorio`, que en cada skill con evals ejecuta cada orden concreta del `SKILL.md` y de sus referencias a mano con `ComprobarSinRed` (`--offline`) sobre `cacheDeUnaSesion` de la primera eval de la carpeta de evals de la skill, por nombre, que no declara `sin_servir`, da una `Falta` por cada una que no sale con 0 con el defecto `<skill>: <fichero>: «…» terminó con código N: …` y exige al menos una orden en el repositorio; y el mutante del control 10 en `TestControlesSobreLasReferenciasAMano` (una orden de `buscar` sin marcador en una referencia a mano de una copia, en rojo). Ajustar `cacheDeUnaSesion` en `internal/evals/preparar_test.go` o `internal/evals/preparar.go` solo si el control lo necesita. En verde con las skills de T005–T007. FR-019, SC-005. Rutas: internal/evals/conjunto_test.go, internal/evals/preparar.go, internal/evals/preparar_test.go

---

## Phase 6: US2 — `make ci` hace cumplir la forma de la especificación y de la guía (Priority: P1)

**Goal**: `skills-check` sale en rojo, con un defecto que nombra la skill, el fichero y la regla, con la medida y el
límite, cuando una skill se sale de la forma; las tres del repositorio lo dejan en verde. Sustituye al límite de menos
de 300 líneas.

**Independent Test**: `make skills-check`, `go test -count=1 -run '^TestSkillsDelRepositorio$' ./internal/app/` y
`go test -count=1 ./internal/skills/` (quickstart, escenario 2).

- [ ] T009 [US2] Ampliar la validación de la especificación (FR-002, FR-003, FR-005, FR-011; research D11; contracts/gates-de-skills-check.md §2 y §2.1) en `internal/skills/frontmatter.go`: `ValidarFrontmatter` da como defecto un `name` que contiene `claude` o `anthropic`, una etiqueta XML (coincidencia de `</?\p{L}[^<>]*>`) en `name` y en `description`, y una `description` de más de 900 caracteres contados como runas (sustituye al máximo de 1024 y a la regla «description con < o >»); cada defecto de `ValidarFrontmatter` empieza por `SKILL.md: ` con el texto de hoy detrás. Tests: `TestValidarFrontmatter` (los defectos nuevos, el prefijo y la descripción de 900 sin defecto) y `TestEtiquetaXML` (casan `<ejemplo>`, `</b>`, `<br/>`, `<descripcion skill="x">`; no casan «< 5000», `a<b`, `a < b > c`, `<>`, `<1>`) en `internal/skills/frontmatter_test.go`; los textos esperados de `internal/skills/sincronia_test.go` con el prefijo; y en `TestSkillsDelRepositorio` de `internal/app/skills_test.go`, el subtest `frontmatter` con los mutantes de SC-002 sobre `copiaConLaSkill` (`name: claude-legal`, `name: x-anthropic`, etiqueta en `name`, `<ejemplo>` en `description`, «menos de 5000 tokens (< 5000)» sin defecto, descripción de 901 en rojo y de 900 en verde), cada uno con el defecto que nombra skill, fichero y regla; en rojo si alguno no se detecta. Las tres skills reescritas pasan. FR-002, FR-003, FR-005, FR-011, SC-002, SC-009. Rutas: internal/skills/frontmatter.go, internal/skills/frontmatter_test.go, internal/skills/sincronia.go, internal/skills/sincronia_test.go, internal/app/skills_test.go
- [ ] T010 [US2] Sustituir el límite de líneas por los presupuestos del cuerpo (FR-004, FR-010; research D12; contracts/gates-de-skills-check.md §2.2) en `internal/skills/sincronia.go`: `regenerarSkill` mide el cuerpo —el `SKILL.md` con la región de la tabla de comandos regenerada, desde el byte que sigue a la línea `---` que cierra el frontmatter— con `ContarLineas` y `utf8.RuneCount`, y da `SKILL.md: cuerpo de N líneas (máximo 499)` con 500 o más y `SKILL.md: cuerpo de N caracteres (máximo 11000)` con más de 11 000; desaparecen `maximoDeLineas`, `defectoDeLasLineas` y el defecto «SKILL.md tiene N líneas (máximo 299)»; `make skills-sync` da los defectos y no escribe nada, como hoy. Tests: `TestCuerpoDeLaSkill` en `internal/skills/sincronia_test.go` (499 líneas en verde y 500 en rojo; 11 000 runas con más de 11 000 bytes en verde y 11 001 en rojo, con caracteres de varios bytes), retirando los casos `trescientas-lineas` y `trescientas-lineas-al-regenerar`; en `TestSkillsDelRepositorio` de `internal/app/skills_test.go`, el subtest `presupuestos` con esos cuatro mutantes sobre `copiaConLaSkill` y sin el subtest `trescientas-lineas`; `internal/skills/export_test.go` si hace falta. Las tres skills reescritas pasan. FR-004, FR-010, FR-011, SC-002, SC-009. Rutas: internal/skills/sincronia.go, internal/skills/sincronia_test.go, internal/skills/export_test.go, internal/app/skills_test.go
- [ ] T011 [US2] Añadir los gates de `references/` (FR-006 a FR-009, FR-011; research D13, D14; contracts/gates-de-skills-check.md §2.3 y §3): nuevo `internal/skills/markdown.go` con los bloques de código (vallas de tres o más `` ` `` o `~` tras hasta tres espacios, hasta la misma marca o el final), los enlaces `[texto](destino)` fuera de bloques (destino hasta el primer `)` o espacio, sin `#…`), los apartados `## ` fuera de bloques y el índice (una lista antes del primer apartado que nombra, por orden, cada apartado); en `internal/skills/sincronia.go` y `internal/skills/referencias.go`, el defecto `references/<entrada>: no es un fichero .md` para un subdirectorio, un enlace o un fichero de otra extensión, que deja de ser deriva y no se retira (FR-006), y, sobre el cuerpo regenerado y las referencias (las generadas como las regenera su fila, las a mano como están), los defectos de la referencia sin enlace desde `SKILL.md` (FR-007), de la referencia que enlaza otra de su skill resuelta con `path.Join` desde `references/` (FR-008) y de la de más de 100 líneas sin índice o sin ningún `## ` (FR-009), con los textos del §2. Tests: `TestEnlacesDeMarkdown` (fuera y dentro de bloques, `#apartado`, resolución desde `references/`) y `TestIndiceDeLaReferencia` (100 y 101 líneas, con y sin índice, sin `## `) en `internal/skills/markdown_test.go`; `TestReferenciasAMano` y `TestRegenerarYComparar` con la entrada no `.md`; y en `TestSkillsDelRepositorio` de `internal/app/skills_test.go`, el subtest `referencias` con los mutantes de SC-002 sobre `copiaConLaSkill` (subdirectorio y `.txt` en `references/`, referencia sin enlace, referencia enlazada desde otra, 101 líneas sin índice en rojo y 100 en verde, 101 sin `## ` en rojo y 100 en verde), cada uno con su defecto; en rojo si alguno no se detecta. Las tres skills reescritas pasan sin defectos. Entrada en `CHANGELOG.md` con los gates nuevos de `skills-check` y la retirada del límite de 300 líneas (FR-023). FR-006, FR-007, FR-008, FR-009, FR-011, SC-002, SC-009. Rutas: internal/skills/markdown.go, internal/skills/markdown_test.go, internal/skills/sincronia.go, internal/skills/sincronia_test.go, internal/skills/referencias.go, internal/skills/referencias_test.go, internal/skills/export_test.go, internal/app/skills_test.go, CHANGELOG.md

**Checkpoint**: la forma de las tres skills ya no se puede perder sin que `make ci` lo diga.

---

## Phase 7: US3 — El job de evals juzga que la descripción está en tercera persona (Priority: P1)

**Goal**: en una ejecución del job cuya elección incluye el cambio de una descripción, el trabajo `evals (<skill>)` de
cada skill juzga (o reutiliza por huella) la descripción de su skill con la clase `habla_en_primera_o_segunda_persona`,
con umbral 0 y `decide: true`; nada de ello corre en `make ci` ni en un paso del run (FR-033).

**Independent Test**: `go test -count=1 -run '^(TestJuezDeLasDescripciones|TestMedidaVersionada|TestEjecucionSinMedir|TestDecisionDelJob|TestBusquedaDeMedidasDeMain|TestInformeConLaDescripcion|TestDefinicionDelJob|TestGuionDeLaMedida)$' ./internal/evals/` (quickstart, escenario 6), con votos simulados y el `claude` sustituto; los puntos de entrada que abren sesiones llevan `//go:build evals` y no se compilan en `make ci`.

- [ ] T012 [US3] Construir el juez de las descripciones y su identidad (FR-025, FR-026, FR-027; research D2, D4, D5, D8; contracts/juez-de-las-descripciones.md §1, §2 y §4): en `internal/evals/juez.go`, partir `juzgar` en lo que arma el mensaje de una respuesta (`mensajeDelVoto`) y lo que vota un mensaje contra el texto en que se busca la frase, sin cambiar el comportamiento del juez de las respuestas; nuevo `internal/evals/descripciones.go` con `juezDeLasDescripciones(raiz)`, que lee `rubrica.md`, `esquema.json`, `casos.yaml` y `medida.json` del directorio de evidencia del ADR 0046 y devuelve un `*Juez` con la clase `{Nombre: "habla_en_primera_o_segunda_persona", Decide: true}`; el control de identidad, que compara el SHA-256 de cada uno de los cuatro con el que registra su `manifiesto.json`; `mensajeDeLaDescripcion(skill, descripcion)`, el de `mensaje_de` (descripción de `LeerFrontmatter` sin los blancos de `esBlanco`, entre `<descripcion skill="<skill>">` y `</descripcion>`, una línea en blanco y «Responde a la pregunta de la rúbrica sobre esta descripción.», sin salto final); y la huella del juicio (SHA-256 del texto de líneas `skill`, `descripcion`, `rubrica`, `esquema`, `casos`, `medida` y `juez <modelo> <versión>`). Tests: `TestJuezDeLasDescripciones` en `internal/evals/descripciones_test.go` (identidad con el manifiesto y un byte cambiado en cualquiera de los cuatro, sobre una copia en `t.TempDir()`, en rojo; el mensaje igual, byte a byte, al construido como `mensaje_de` para un caso de `casos.yaml`; la huella cambia con cada uno de sus datos) y, en `TestMedidaVersionada` de `internal/evals/medida_test.go`, la medida de las descripciones comprobada con `comprobarLaMedida` frente al modelo y la versión que fija el workflow de las evals (`fijadosParaElJuez`) y sus seis mutaciones (otra huella de la rúbrica, otra de los casos, otro modelo, otra versión, 62 de 63 y 1 de 36), cada una con su línea; `TestJuezDeLaDefinicionDelJob` y `TestCopiasDelJuez` siguen en verde. Sin abrir ninguna sesión. FR-025, FR-026, FR-027, FR-033, SC-006, SC-008, SC-014. Rutas: internal/evals/juez.go, internal/evals/juez_test.go, internal/evals/descripciones.go, internal/evals/descripciones_test.go, internal/evals/medida.go, internal/evals/medida_test.go
- [ ] T013 [US3] Votar una descripción con la regla de la `votacion` del juez de las evals (FR-028, FR-029, FR-032, FR-033; research D6; contracts/juez-de-las-descripciones.md §5): en `internal/evals/descripciones.go`, el juicio de una descripción que pide los votos por orden con el votante de `nuevoVotanteDelGuion` (el guion de voto de las evals, sin cambios) y la `votacion` de `internal/evals/juez.go` con la clase única: marcada con tres «si» unánimes con su frase comprobada con `fraseEsta`; el segundo voto solo tras un primer «si» y el tercero solo tras el segundo; un «si» con una frase que no está, nulo y repetido una sola vez (si vuelve a ser nulo, su «si» no cuenta); el voto cortado por el tope (35 s más 5), repetido una vez, y cortado otra vez, sin juzgar; la sesión con error, sin la forma del esquema o cortada por un límite de uso, sin repetición y sin juzgar, con el motivo del voto; como mucho seis votos; `frase` y `motivo` del primer voto. Tests en `TestJuezDeLasDescripciones` de `internal/evals/descripciones_test.go` con un `Votante` simulado (tres «si» → marcada con su frase y su motivo; «si», «si», «no» → sin marcar; primer «no» → sin marcar y 0 votos pedidos después; frase que no está → nulo y una repetición; cortado dos veces → sin juzgar; error → sin juzgar y sin repetición; ≤ 6 votos en el peor caso) y con el `claude` sustituto de `internal/evals/sustitutos_test.go` (el mensaje llega por la entrada estándar y no se abre ninguna sesión con modelo); en rojo con cualquiera distinto. FR-028, FR-029, FR-032, FR-033, SC-007, SC-014. Rutas: internal/evals/descripciones.go, internal/evals/descripciones_test.go, internal/evals/juez.go, internal/evals/juez_test.go, internal/evals/sustitutos_test.go
- [ ] T014 [US3] Decidir en cada trabajo si se juzga la descripción y reutilizar su juicio por huella (FR-030, Clarificación Q2; research D3, D4; contracts/juez-de-las-descripciones.md §2): en `internal/evals/decision.go` y `internal/evals/origen.go`, `decidirElTrabajo` añade el bloque `descripcion` (`skill`, `huella`, `juzgada` y, si se reutiliza, `procedencia` con la forma de H30 y `juicio`) solo cuando la ejecución tiene `-propuesta` y `-base` no vacías y el `SKILL.md` de alguna skill cumple `cambiaLaDescripcion` entre la base y la cabeza, en los dos modos (con reutilización y en la suite completa sobre una propuesta, en la que se lee también la base; `internal/evals/suite_completa.go` si hace falta); sin esa condición, sin bloque; en `internal/evals/reutilizacion.go`, `buscarEnLasMedidasDeMain` busca en el mismo recorrido de H30 (10 medidas como mucho, de menos de 90 días, de propuestas fusionadas, con el informe escrito) un bloque `descripcion` de la skill con la misma huella y juicio completo (no «sin juzgar»), y en la suite completa no busca; `TestDecisionDelTrabajo` de `internal/evals/job_test.go` (`//go:build evals`) escribe el bloque en el fichero de la decisión, y `scripts/evals.sh` y `.github/workflows/evals.yml` le pasan lo que necesite. Tests: `TestDecisionDelJob` en `internal/evals/decision_test.go` (bloque con una descripción cambiada en los dos modos y con lo común cambiado a la vez; sin bloque sin propuesta, con la etiqueta `evals` o la suite completa sin descripciones cambiadas y con un cambio solo de la rúbrica, el esquema, los casos, la medida, el modelo o la versión del juez) y `TestBusquedaDeMedidasDeMain` en `internal/evals/reutilizacion_test.go` (reutiliza con la misma huella y su procedencia; 0 juicios reutilizados con otra huella, de una medida no válida o sin juzgar), con respuestas sintéticas de la plataforma en memoria. FR-030, SC-010. Rutas: internal/evals/decision.go, internal/evals/decision_test.go, internal/evals/origen.go, internal/evals/reutilizacion.go, internal/evals/reutilizacion_test.go, internal/evals/suite_completa.go, internal/evals/job_test.go, scripts/evals.sh, .github/workflows/evals.yml
- [ ] T015 [US3] Juzgar en la ejecución del trabajo y publicar el juicio, el umbral y el motivo (FR-026, FR-031, FR-032, FR-033; research D7, D8; contracts/juez-de-las-descripciones.md §3 y §6): en `internal/evals/ejecucion.go`, con el bloque `descripcion` en la decisión, `ejecutarElJob` llama a `comprobarLaMedida(juezDeLasDescripciones, modelo, versión)` antes de abrir ninguna sesión, de las evals o del juez, y con alguna línea escribe el informe sin medir con cada línea como motivo precedida de `juez de las descripciones: ` y no abre nada; después de las sesiones de las evals, si `juzgada`, vota la descripción de su skill (T013) y, si no, toma el juicio del bloque; en `internal/evals/informe.go` y `internal/evals/umbrales.go`, `informe.json` gana el bloque `descripcion` (marcada, frase, motivo, votos, `sin_juzgar`, o `juzgada: false` con `procedencia`) y el umbral `habla_en_primera_o_segunda_persona` (`medida` las marcadas entre las juzgadas en la ejecución, `total` 1 si se juzgó y 0 si se reutilizó o quedó sin juzgar, `comparacion` `<=`, `umbral` 0, `decide: true`), cuyo incumplimiento da el motivo de hoy y el veredicto `fallo`; la descripción sin juzgar añade el motivo propio «el juez de las descripciones dejó sin juzgar la descripción de <skill>: <motivo del voto>», que no cita ningún umbral; `informe.md` gana el apartado «La descripción»; sin el bloque, ni bloque, ni umbral, ni apartado; `scripts/evals.sh` pasa `-modelo-del-juez`, `-version-del-juez` y `-claude-del-juez` en todo trabajo, no solo cuando la carpeta de evals de la skill tiene su carpeta `juez`, y el punto de entrada del job en `internal/evals/job_test.go` los recibe. Tests: `TestEjecucionSinMedir` en `internal/evals/ejecucion_test.go` (con la medida que no corresponde o no se cumple y un abridor de sesiones que falla si lo llaman: 0 sesiones y el informe sin medir), `TestInformeConLaDescripcion` en `internal/evals/informe_test.go` (marcada → umbral en rojo, `fallo`, con la frase y el motivo; sin marcar; sin juzgar → `fallo` con su motivo propio, distinguible sin modelo del de una marcada; reutilizada → 0 en el umbral, con su procedencia; sin bloque → sin umbral ni apartado), y los tests de `internal/evals/umbrales_test.go` y de `internal/evals/sustitutos_test.go` que fijan las banderas de `scripts/evals.sh`. FR-026, FR-031, FR-032, FR-033, SC-007, SC-008, SC-010, SC-014. Rutas: internal/evals/ejecucion.go, internal/evals/ejecucion_test.go, internal/evals/informe.go, internal/evals/informe_test.go, internal/evals/umbrales.go, internal/evals/umbrales_test.go, internal/evals/descripciones.go, internal/evals/job_test.go, internal/evals/sustitutos_test.go, scripts/evals.sh
- [ ] T016 [US3] Cubrir con el tope de cada trabajo el peor caso de la descripción (SC-015, FR-030; research D10): en `internal/evals/definicion.go`, `peorCasoDelTrabajo` suma en todo trabajo `evals` el peor caso de una descripción, `votosPorRespuestaComoMucho` (6) × (`topeDelVoto` + `margenDelVoto`) = 240 s, y la instalación del Claude Code del juez (60 s) en el trabajo de una skill sin juez de las respuestas (hoy `legal-core`); `TestDefinicionDelJob` en `internal/evals/definicion_test.go` con las cadenas esperadas de cada skill con esas cifras (el de `boe-legislacion` pasa de 19 877 s a 20 117 s, bajo el tope de 352 min de `.github/workflows/evals.yml`, que no cambia) y en rojo si el tope de un trabajo es menor que la suma del peor caso de sus evals y el de la descripción. FR-030, SC-015. Rutas: internal/evals/definicion.go, internal/evals/definicion_test.go, .github/workflows/evals.yml
- [ ] T017 [US3] Medir la clase con el código del job (FR-034; research D9; contracts/juez-de-las-descripciones.md §7): punto de entrada `TestMedidaDelJuezDeLasDescripciones` en `internal/evals/job_test.go` (`//go:build evals`), que vota los 99 casos de `casos.yaml` con el mensaje de T012 y la votación de T013, por orden, y escribe en `-salida` la medida con las claves que lee `leerMedidaDelJuez` (`clase`, `fecha`, `modelo_del_juez`, `version_de_claude_code`, `rubrica`, `casos`, `defectos`, `correctos`) y `origen`, con una línea en el registro por caso mal etiquetado y por caso sin juzgar, sin escribir nada en el repositorio, apoyado en lo que necesite de `internal/evals/descripciones.go` y `internal/evals/medida.go`; `scripts/evals-medir-juez.sh` admite `descripciones` y ejecuta ese punto de entrada, imprimiendo la medida entre `--- inicio de medida.json ---` y `--- fin de medida.json ---`; el trabajo `medida` de `.github/workflows/evals.yml` gana la entrada `descripciones` en su matriz, con concurrencia 5. Tests: `TestGuionDeLaMedida` en `internal/evals/medida_test.go` (el guion con `descripciones`, con el `claude` sustituto) y `TestDefinicionDelJob` en `internal/evals/definicion_test.go` (la entrada `descripciones` de `medida` y su peor caso, 485 s + 60 s + ⌈99 × 6 / 5⌉ × 40 s = 5305 s, bajo su tope de 269 min). Entrada en `CHANGELOG.md` con el juez de las descripciones (FR-023). La medida la lanza fuera del run quien tiene la etiqueta `evals-medir-juez` (SC-011); ninguna orden de esta tarea abre una sesión. FR-023, FR-033, FR-034, SC-011, SC-014, SC-015. Rutas: internal/evals/job_test.go, internal/evals/descripciones.go, internal/evals/medida.go, internal/evals/medida_test.go, internal/evals/definicion.go, internal/evals/definicion_test.go, scripts/evals-medir-juez.sh, .github/workflows/evals.yml, CHANGELOG.md

**Checkpoint**: el job juzga las descripciones solo cuando una cambia, y el umbral lo hace fallar si alguna habla en
primera o segunda persona.

---

## Phase 8: US6 — La guía de autoría y lo que la acompaña (Priority: P3)

**Goal**: quien escribe o revisa una skill sabe qué va en el cuerpo y qué en una referencia, cómo se escribe la
descripción, qué presupuestos hay y por qué, y la regla de `references/`.

**Independent Test**: la revisión final contrasta `docs/SKILLS.md` con el ADR 0046 y con las cifras de los gates
(SC-013); `make ci` en verde (quickstart, escenario 8).

- [ ] T018 [US6] Escribir la guía de autoría `docs/SKILLS.md` (FR-001): las dos fuentes, la especificación Agent Skills (agentskills.io, «Specification») y la guía de Anthropic («Skill authoring best practices»), con la fecha de consulta (2026-10-23) y las citas que registra el ADR 0046, sin volver a consultarlas; (a) qué va en el cuerpo —el protocolo, las reglas invariantes, las formas fijas que las evals comprueban sin modelo (las líneas `⚠`), la tabla de comandos y, por cada referencia, cuándo leerla— y qué en una referencia; (b) la descripción en tercera persona, qué hace la skill y cuándo se usa, con los dos casos correctos que una lista no distingue (la pregunta citada y el verbo con la skill de sujeto) y el juez de las descripciones del job; (c) los presupuestos con su porqué —menos de 500 líneas, como mucho 11 000 caracteres (2,26 caracteres por token, unos 4870 tokens), como mucho 900 en la descripción (el 88 % de 1024), contados como runas, y que la razón se vuelve a medir cuando cambia el modelo que decide (ADR 0031)—; (d) la regla de `references/` de FR-006 a FR-009 y FR-012, con las cifras y los textos de los defectos de T009–T011. Actualizar `CONTRIBUTING.md` para que deje de pedir menos de 300 líneas y diga los gates de `skills-check`, la guía y la regla de las referencias a mano (FR-023), y completar la entrada de `CHANGELOG.md` si le falta algo de FR-023. Ejecutar los escenarios 1 a 8 de quickstart.md (sin red ni modelo) y `make ci`, que tiene que terminar con 0 (SC-001). FR-001, FR-023, SC-001, SC-013. Rutas: docs/SKILLS.md, CONTRIBUTING.md, CHANGELOG.md

---

## Dependencies & Execution Order

### Orden de las fases

- **Fase 2 (US4, T001–T002)**: sin dependencias; bloquea todo lo demás: sin ella, `make skills-sync` retiraría las
  referencias a mano de US1 y el conjunto de dependencias daría error con ellas.
- **Fase 3 (US5, T003–T004)**: tras T001 (los controles leen `skills.ReferenciasAMano`); antes de US1, para que la
  reescritura no mueva ningún texto fuera de un control.
- **Fase 4 (US1, T005–T007)**: tras T001–T004. El orden jurisprudencia → legal-core → boe-legislacion es el del plan
  (de la que tiene más margen a la más ajustada).
- **Fase 5 (US5, T008)**: tras T007, que cambia los dos ejemplos de `boe-legislacion` que la caché no sirve (`buscar` y
  `analisis`): antes, el control saldría en rojo.
- **Fase 6 (US2, T009–T011)**: tras T005–T007, porque los gates tienen que salir en verde con las tres skills del
  repositorio (las de `main` pasan de 11 000 caracteres y la descripción de `boe-legislacion` de 900).
- **Fase 7 (US3, T012–T017)**: tras T001 (y en la práctica tras US1, porque las descripciones nuevas son las que juzga
  el job); en orden: T012 (juez, identidad, mensaje, huella, medida versionada) → T013 (votos) → T014 (decisión y
  reutilización) → T015 (ejecución, informe y umbral) → T016 (tope) → T017 (medida del juez).
- **Fase 8 (US6, T018)**: al final, porque la guía da las cifras y los textos de los gates ya construidos.

### Dentro de cada historia

- Cada tarea es una rebanada vertical con su test y deja `make ci` en verde; ninguna depende de una posterior.
- T009, T010 y T011 tocan `internal/skills/sincronia.go` e `internal/app/skills_test.go`: van en secuencia.
- T012 a T017 tocan `internal/evals/descripciones.go`, `juez.go`, `job_test.go` o `.github/workflows/evals.yml`
  compartidos: van en secuencia.

### Oportunidades de paralelismo

El bucle del workflow ejecuta las tareas una a una, así que ninguna lleva `[P]`. Por ficheros, serían independientes:

- T003 (`internal/evals/conjunto_test.go`) y T004 (`internal/app/skills_test.go`).
- T005, T006 y T007 (cada una en su `skills/<skill>/`), salvo su entrada de `CHANGELOG.md`.
- T016 (`definicion.go`) frente a T014 y T015, salvo `.github/workflows/evals.yml`.

## Implementation Strategy

### MVP (US1 sobre US4 y US5)

1. T001–T002: la regla de `references/` y las dependencias.
2. T003–T004: los controles leen la skill entera.
3. T005–T007: las tres skills reescritas. **Parar y validar**: `make ci` en verde y las medidas del escenario 1 de
   quickstart; es la entrega que mide el job de evals (SC-010).

### Entrega incremental

4. T008: las órdenes de ejemplo.
5. T009–T011: los gates de la forma (US2), que sustituyen al límite de 300 líneas.
6. T012–T017: el juez de las descripciones en el job (US3).
7. T018: la guía (US6) y `make ci` final.

## Trazabilidad

### Cada requisito y sus tareas

| Requisito | Tareas |
|---|---|
| FR-001 | T018 |
| FR-002, FR-003, FR-005 | T009 |
| FR-004, FR-010 | T010 |
| FR-006, FR-007, FR-008, FR-009 | T011 |
| FR-011 | T009, T010, T011 |
| FR-012 | T001 (clasificación), T011 (entrada no `.md`) |
| FR-013 | T001 (la regla), T005–T007 (el contenido) |
| FR-014 | T002 |
| FR-015 | T005–T007 (las referencias nuevas, que comprueba `TestSkillsEmpotradas` sin cambios) |
| FR-016, FR-017, FR-018 | T003, T004 |
| FR-019 | T007 (los ejemplos), T008 (el control) |
| FR-020, FR-021, FR-022, FR-024 | T005, T006, T007 |
| FR-023 | T005, T006, T007, T011, T017, T018 |
| FR-025, FR-027 | T012 |
| FR-026 | T012 (`make ci`), T015 (el job) |
| FR-028, FR-029 | T013 |
| FR-030 | T002 (elección), T014 (decisión y reutilización), T016 (tope) |
| FR-031 | T015 |
| FR-032 | T013 (el voto), T015 (el informe) |
| FR-033 | T012, T013, T015, T017 |
| FR-034 | T017 |
| SC-001 | todas; T018 ejecuta `make ci` al final |
| SC-002 | T009, T010, T011 |
| SC-003 | T001 |
| SC-004 | T002 |
| SC-005 | T003, T004, T008 |
| SC-006 | T012 |
| SC-007 | T013, T015 |
| SC-008 | T012, T015 |
| SC-009 | T005–T007 (las skills), T009–T011 (los gates sobre el repositorio) |
| SC-010 | T002, T014, T015 (en `make ci`); T005–T007 (las skills que mide el job en el cierre) |
| SC-011 | T017 construye la medida; la lanza fuera del run quien tiene la etiqueta `evals-medir-juez` |
| SC-012 | T005–T007 (cada regla donde dice la tabla de research.md); la contrasta la revisión final |
| SC-013 | T018; la contrasta la revisión final |
| SC-014 | T012, T013, T015, T017 |
| SC-015 | T016, T017 |

### Cada fila de «## Controles de umbral» de plan.md y la tarea que construye su control

| Requisito de la fila | Control | Tarea |
|---|---|---|
| SC-001 | `ci:Makefile:ci` | todas; T018 lo ejecuta al final |
| FR-002, FR-003, FR-005, FR-011, SC-002, SC-009 | `TestSkillsDelRepositorio` (frontmatter) | T009 |
| FR-004, FR-010, SC-002, SC-009 | `TestSkillsDelRepositorio` (presupuestos) | T010 |
| FR-006 a FR-009, FR-011, SC-002 | `TestSkillsDelRepositorio` (referencias) | T011 |
| FR-012, SC-003 | `TestSkillsDelRepositorio` (deriva, sincronia-con-referencias-a-mano) | T001 |
| FR-014, SC-004 | `TestConjuntoDeCadaEval` | T002 |
| FR-014, SC-004 | `TestEleccionDeLosMutantes` | T002 |
| FR-015 | `TestSkillsEmpotradas`, sin cambios (falla hoy si falta un fichero de `references/`) | T005–T007 lo ejercen con las referencias nuevas |
| FR-016 a FR-018, SC-005 | `TestControlesSobreLasReferenciasAMano` | T003 (controles 1 a 7), T008 (control 10) |
| FR-016 a FR-018 | `TestEvalsDelRepositorio` sobre las tres skills | T003 (el control), T005–T007 (las skills) |
| FR-016, SC-005 | `TestSkillsDelRepositorio` (normas-nombradas, sin-instrucciones-de-evals) | T004 |
| FR-019, SC-005 | `TestEvalsDelRepositorio` (ordenes-de-ejemplo) | T008 |
| FR-019 | `TestOrdenesDeEjemplo` | T008 |
| FR-025, SC-006 | `TestJuezDeLasDescripciones` (identidad) | T012 |
| FR-026, SC-008 | `TestMedidaVersionada`, `TestEjecucionSinMedir` | T012, T015 |
| FR-027 a FR-029, FR-032, FR-033, SC-007, SC-014 | `TestJuezDeLasDescripciones` (mensaje y votos) | T012 (mensaje), T013 (votos) |
| FR-030, SC-010 | `TestEleccionDeLosMutantes`, `TestDecisionDelJob` | T002, T014 |
| FR-030 | `TestBusquedaDeMedidasDeMain` | T014 |
| FR-031, FR-032 | `TestInformeConLaDescripcion` | T015 |
| FR-031, SC-010 | `evals:<skill>:habla_en_primera_o_segunda_persona` (las tres skills) | T015 construye el umbral con `decide: true`; lo mide el job en el cierre |
| SC-010 | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5`, `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5` | los de hoy, sin cambios; T005–T007 entregan lo que miden en el cierre |
| SC-015, FR-030 | `TestDefinicionDelJob` | T016 (trabajos `evals`), T017 (trabajo `medida`) |

## Definition of Done (§1) que aplica

- **CHANGELOG.md**: T005, T006 y T007 (la versión de cada skill, research D19), T011 (los gates) y T017 (el juez de las
  descripciones); T018 la completa si falta algo de FR-023.
- **ADR**: ninguno nuevo; el ADR 0046 está en `main` desde la entrada («Fuera de alcance»).
- **docs/SOURCES.md**: no aplica, ninguna fuente nueva (plan, «Datos externos»).
- **Schemas**: no aplica, ningún esquema cambia (plan, «Fixtures»).
- **Fixtures**: no aplica; las copias, los casos y los votos se escriben en `t.TempDir()` o son constantes del test.
- **Skill y evals** (§1.10): las tres skills reescritas (T005–T007), sin evals nuevas («Fuera de alcance»); las mide el
  job en el cierre (SC-010).
- **Documentación**: `docs/SKILLS.md` y `CONTRIBUTING.md` (T018).

## Notes

- Ninguna tarea abre una sesión con modelo, usa la red, graba datos, escribe en el directorio de evidencia ni en
  `data/`, toca `scripts/workflow/`, `.specify/` o `internal/workflow/`, publica la rama ni mide el job remoto: el cierre
  lo hace el workflow tras la revisión final (ADR 0018).
- SC-011 (la medida del juez de las descripciones, 63 de 63 y 0 de 36) la lanza fuera del run quien tiene la etiqueta
  `evals-medir-juez`, antes de fusionar; T017 deja el código que la mide.
- Lo que no se construye, por el umbral de materialidad (plan, «Trazabilidad»): casos propios para un `references/`
  enlazado, truncado o con permisos cambiados (regla genérica, defecto `inesperado`), un analizador de Markdown general
  y una copia del juez.
