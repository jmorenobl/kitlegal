---

description: "Tareas de H31.1 · boe-legislacion habla de la norma, no de la consulta, y lee lo que nombra"
---

# Tasks: H31.1 · `boe-legislacion` habla de la norma, no de la consulta, y lee lo que nombra

**Input**: `specs/028-h31-1-boe-legislacion-habla/` — spec.md, plan.md, research.md (D1–D15, V1–V13, S1–S4),
data-model.md, contracts/ (`skill-boe-legislacion-v0-3-1.md`, `lecturas-y-grabaciones.md`, `guia-y-changelog.md`) y
quickstart.md.

**Aceptación e2e: no aplica** (plan.md, «Aceptación e2e» y Complexity Tracking; research D13): no hay tarea
`[aceptacion]` ni guiones `testscript`. El hito no cambia ningún applet, verbo, bandera, código de salida ni salida del
binario, y `kitlegal skills install` instala la skill nueva como cualquier fichero empotrado, que `TestSkillsEmpotradas`
ya compara byte a byte. La aceptación sin persona es `make ci` en verde (SC-001 a SC-005) y el job de evals sobre la
cabeza de la rama, que el workflow lanza en el cierre y que mide las tres skills con los umbrales de hoy (SC-007 a
SC-011). El spec deja fuera evals nuevas («Fuera de alcance»): ninguna tarea añade una eval; la única escritura en
`evals/boe-legislacion/` son las nueve `lecturas` de FR-009 (T002).

**Tests**: cada tarea lleva su test o los controles que la cubren y la implementación mínima, y deja `make ci` en verde al
terminar. Todo sin red ni modelo (ADR 0032). Los tests que cambian son los dos que enumeran las lecturas (T002), que se
escriben antes que las evals y se ven en rojo; el texto de la skill y la guía los cubren los gates y controles de H31 que
ya existen (`TestSkillsDelRepositorio`, `TestCuerpoDeLaSkill`, `TestEvalsDelRepositorio`, `TestSkillsEmpotradas`), sin
cambios (plan, «Controles mecánicos»), más las órdenes de quickstart.md que cada tarea nombra. Ningún paquete, función,
tipo ni fichero de código de producción cambia.

**Datos externos**: una fuente, `boe.legislacion-consolidada`, con fila revisada en `docs/SOURCES.md` de `main`. T001
(`[datos]`) escribe solo el manifiesto; justo después de su commit, el paso `grabar_datos` del workflow, sin modelo, graba
los ocho bloques con `TestGrabarEvals` (`internal/evals/grabacion_test.go`, `//go:build grabacion`, que ya existe). T002
ya parte de esas grabaciones. Ninguna tarea usa la red ni escribe material de origen, y ningún fichero de `data/` cambia.

**Organización**: las fases siguen el «Orden de implementación (de dentro afuera)» de plan.md, que es el único en que cada
paso deja `make ci` en verde: el manifiesto (T001) antes que las lecturas, porque una lectura sin grabar deja sin
preparar la caché de cualquier sesión de la skill y pone en rojo `grabado`, `ordenes-de-ejemplo` y `lo-que-lleva-a-leer`
(research D8, V7); las lecturas (T002) antes que la skill, para que las sesiones de las evals 14, 15 y 17 sirvan lo que la
regla nueva pide leer; la skill (T003); y la guía (T004). La entrada de `CHANGELOG.md` (FR-014) va en T004, la última,
y no en T003 como agrupa el paso 3 del orden del plan: la entrada describe también las lecturas (T002) y el párrafo de la
guía (T004), y escrita antes de T004 diría de `docs/SKILLS.md` §1 algo que aún no es cierto; su texto es el de
contracts/guia-y-changelog.md §2 con la cifra del cuerpo que da el gate tras T003.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: puede ir en paralelo (ficheros distintos, sin dependencias de tareas pendientes). Ninguna lo es aquí.
- **[Story]**: la historia de usuario del spec (US1…US6).
- Cada tarea nombra todas las rutas que crea o modifica; declarar `x.go` permite `x_test.go`. `go.mod`, `go.sum`,
  `CHANGELOG.md` y `specs/028-h31-1-boe-legislacion-habla/` están siempre permitidos.

---

## Phase 1: Setup

Sin tareas: el módulo, las herramientas, los objetivos de `make`, el arnés de evals, el manifiesto y su test de grabación
ya existen; el hito no añade objetivos de `make`, dependencias ni paquetes (plan, «Objetivos de make: ninguno nuevo»;
«Primary Dependencies»).

---

## Phase 2: Foundational

Sin tareas aparte: el único prerrequisito bloqueante es la grabación de los ocho bloques, que es la primera tarea de US3
(T001) y el paso `grabar_datos` que la sigue.

---

## Phase 3: US3 — Las evals por materia sirven lo que sus respuestas nombran (Priority: P1)

**Goal**: las evals 14, 15 y 17 llevan en `lecturas` los nueve preceptos que sus respuestas nombraban sin leer, y la
caché de las sesiones de su eval los sirve sin red, sin fecha y con `--fecha 9999-12-31` (FR-009 a FR-011, SC-004,
SC-008).

**Independent Test**: quickstart.md, escenario 6: el manifiesto con los bloques nuevos, las ocho grabaciones, y
`TestEvalsDelRepositorio/lecturas`, `TestEvalsDelRepositorio/grabado`, `TestCacheDeLasSesiones` y
`TestLecturaDeLaEvalSinGrabar` en verde.

- [X] T001 [datos] [US3] Añadir al manifiesto `internal/evals/testdata/grabaciones.json` (`"fuente": "boe.legislacion-consolidada"`, sin tocar ningún otro campo ni entrada) los ocho bloques de FR-010, en las dos entradas que ya existen y en el orden de contracts/lecturas-y-grabaciones.md §2: en la de `titulo_empieza_por` «Ley 35/2006,», `bloques` pasa a `a18`, `a7`, `a19`, `a20`, `dadecimoctava`; en la de «Real Decreto Legislativo 2/2004,», a `a15`, `a60`, `a92`, `a100`, `a104`; y al final del `para` de cada una, la frase literal de esa tabla. La entrada de la Ley 19/2013 no cambia (`a24` ya está grabado para `data/plazos.yaml`). Antes de escribirlos, comprobar con `jq` que cada id es el `id` de la entrada cuyo `titulo` es su precepto («Artículo 7», «Artículo 19», «Artículo 20», «Disposición adicional decimoctava», «Artículo 60», «Artículo 92», «Artículo 100», «Artículo 104») en los índices grabados de H5, `testdata/evals/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2006-20764_texto_indice.json` y `…_id_BOE-A-2004-4214_texto_indice.json` (solo lectura), nunca compuesto ni de memoria. Nada más en la tarea: ni evals, ni tests, ni grabaciones, ni red; las graba el paso `grabar_datos` tras el commit con `TestGrabarEvals`. Test: `TestManifiestoDeGrabaciones` sigue en verde, la primera orden `jq` de quickstart.md escenario 6 da `Ley 35/2006, a18 a7 a19 a20 dadecimoctava` y `Real Decreto Legislativo 2/2004, a15 a60 a92 a100 a104`, y `make ci` en verde (ninguna eval pide aún los bloques nuevos; research D8) (FR-010, SC-004)
- [X] T002 [US3] Añadir las nueve `lecturas` de FR-009 a las evals y los dos tests que las enumeran, en evals/boe-legislacion/14-trlrhl-impuestos-por-materia.yaml, evals/boe-legislacion/15-irpf-rendimientos-por-materia.yaml, evals/boe-legislacion/17-ltaibg-plazo-por-materia.yaml, internal/evals/conjunto_test.go e internal/evals/preparar_test.go, según contracts/lecturas-y-grabaciones.md §1 y §4, y dar holgura en internal/evals/sesiones_test.go a los dos casos de `TestTopeDeLaSesion` que no miran el corte. Primero los tests, vistos en rojo con las evals de hoy: (a) en `probarLasLecturasDeLasEvalsDelRepositorio` (`TestEvalsDelRepositorio/lecturas`), el mapa escrito a mano gana `a60`, `a92`, `a100`, `a104` de `BOE-A-2004-4214` en la 14 tras `a15`, `a7`, `a19`, `a20`, `dadecimoctava` de `BOE-A-2006-20764` en la 15 tras `a18`, y `a24` de `BOE-A-2013-12887` en la 17, cada una con el `motivo` literal de §1, y su comentario dice que llevan lecturas la 14, la 15, la 16 y la 17; (b) en `probarLasLecturasDeLasEvals` (`TestCacheDeLasSesiones/lecturas`), cada fila de la tabla lleva la eval de cuya sesión se prepara la caché —las cuatro de hoy, la 01; las nueve nuevas, la suya (14, 15 o 17), una caché preparada por eval distinta (research D9)— y cada bloque sale con 0 con `boe articulo … --offline` sin fecha y con `--fecha 9999-12-31`. Después, las evals: cada lectura detrás de las de hoy, en el orden de §1, con ese `motivo`; la 17 gana la clave `lecturas:`. Nada más de esas evals cambia —pregunta, `comandos`, `citas`, `informativa`, `por_materia`— ni ninguna otra eval, ni código de producción, ni `.github/workflows/evals.yml`, ni `internal/evals/definicion.go` (research D10). Y en internal/evals/sesiones_test.go, que la verificación del intento 1 vio en rojo por la carga de `make ci` (gates/tarea-T002.md: `termina-antes` y `muere-por-otra-senal` con 124 en lugar de 3 y de 129, porque el tope de 1 s envió TERM antes de que el sustituto de `claude` llegara a su `exit`), la tabla de `TestTopeDeLaSesion` gana un campo `tope` por caso, con `ejecucion.Tope = caso.tope`: `topeDeLosTests` en `basta-term` y `hace-falta-kill`, los que miran el corte, y `topeSinCorte` en `termina-antes` y `muere-por-otra-senal`, que solo miran el código de un sustituto que termina solo, por la regla que ya fija el propio fichero (un test que no mira el tope no depende de la carga de la máquina); el comentario del test lo dice. `topeDeLosTests`, `margenDeLosTests`, `topeSinCorte`, el tope y el margen de verdad (240 s y 10 s) y lo demás del fichero no cambian. Test: los cuatro de quickstart.md escenario 6, `TestTopeDeLaSesion` y `TestDefinicionDelJob` (escenario 7: el peor caso sigue en 20 117 s frente a 21 120 s; FR-013, SC-005) en verde, la orden `git ls-files` del escenario 6 da `8`, `make skills-check` y `make ci` en verde (FR-009, FR-011, FR-013, SC-004, SC-005, SC-008)

**Checkpoint**: las sesiones de las evals 14, 15 y 17 sirven sin red lo que la regla nueva pedirá leer; `make ci` en
verde.

---

## Phase 4: US1 y US2 — El cuerpo de `boe-legislacion` v0.3.1 (Priority: P1) 🎯 MVP

**Goal**: el paso «Responder» vuelve a decir cómo empieza la respuesta y de qué habla, con sus ejemplos (US1, FR-001); el
paso 3 manda leer antes lo que la respuesta nombra y el paso 5 dice el salto (US2, FR-002 a FR-004); el cuerpo cabe en
11 000 caracteres porque siete textos de una sola clase de pregunta pasan a referencias a mano que ya existen (FR-005,
FR-006). Las dos historias van en una tarea: comparten los pasos 3, 4 y 5 del mismo cuerpo y el mismo presupuesto, y el
texto validado es uno (research D15).

**Independent Test**: quickstart.md, escenarios 1 a 5; la tabla de contracts/skill-boe-legislacion-v0-3-1.md §2 la
coteja la revisión final (SC-003); y el job de evals del cierre y el sondeo de la eval 15 (SC-007, SC-012).

- [ ] T003 [US1] Escribir `boe-legislacion` v0.3.1 (US1 y US2) en skills/boe-legislacion/SKILL.md, skills/boe-legislacion/references/redaccion_modificada.md, skills/boe-legislacion/references/buscar_la_norma.md y CONTRIBUTING.md, con el texto de contracts/skill-boe-legislacion-v0-3-1.md: (a) el cuerpo, desde el título hasta antes de la tabla de comandos, pasa a ser el de §1.1 y §1.2 tal cual —pasos 2 a 5 con la primera viñeta del paso 4 (B48, B50, B52, B53) y sus dos ejemplos, los cuatro ejemplos del paso 5, la segunda frase del paso 3 (FR-002), la frase del salto del paso 5 (FR-003), «Cómo se cita», «Redacción en una fecha», «Redacción modificada» y las reglas, con la regla 7 de hoy fuera, la 8 renumerada 7 y la regla 2 remitiendo a «la regla 7»—; el frontmatter (con su `description`, FR-008) y la tabla de comandos generada no se tocan, y se regenera con `make skills-sync`; (b) las siete salidas de §3, cada una a la referencia y con el enlace del paso que dice §3; (c) en `redaccion_modificada.md`, el párrafo de entrada y los apartados «Lo demás de la comprobación» y «Una lectura por bloque» de §4.1, delante de «## La redacción superada», sin ninguna línea que empiece por `⚠`; en `buscar_la_norma.md`, «Si no aparece» como en §4.2; las demás referencias y la generada `normas.md` no cambian; (d) en CONTRIBUTING.md (línea 743), el paréntesis de contracts/guia-y-changelog.md §3, reflujando el párrafo a 120 columnas y nada más. Ninguna orden concreta nueva de `kitlegal boe` en el cuerpo ni en una referencia (las menciones nuevas de `articulo` y `articulos`, sin argumentos; FR-007). Test (los gates y controles de H31, sin cambios, sobre la skill nueva): quickstart.md escenario 1 da 11 000 o menos (10 993 con el contrato; con 11 001, el gate en rojo), escenario 2 da `6`, escenario 3 da `0`, escenario 4 da `0`, escenario 5 (`make skills-check`) en verde con `TestSkillsDelRepositorio`, `TestCuerpoDeLaSkill` y los ocho controles de contenido de `docs/SKILLS.md` §1 de `TestEvalsDelRepositorio`, `grep -c 'regla 7 sin su punto final' CONTRIBUTING.md` da `0`, `TestSkillsEmpotradas` y `make ci` en verde (FR-001 a FR-008, SC-001, SC-002, SC-003)

**Checkpoint**: la skill nueva cumple todos sus gates; lo que mide su significado (SC-007, SC-009, SC-010) lo da el job
del cierre.

---

## Phase 5: US5 — La guía dice qué referencias se leen, y el CHANGELOG (Priority: P2)

**Goal**: quien escribe o reparte una skill —H32 la primera— sabe por `docs/SKILLS.md` §1 qué referencias abren las
sesiones (FR-012); y `CHANGELOG.md` registra v0.3.1 (FR-014).

**Independent Test**: quickstart.md, escenario 8; la revisión final coteja las cifras con `LEEME.md` de la evidencia de
los sondeos (SC-013).

- [ ] T004 [US5] Escribir en docs/SKILLS.md, apartado «## 1. Qué va en el cuerpo y qué en una referencia», detrás del párrafo que empieza «**En una referencia, el detalle que solo pide una clase de pregunta**» y delante de «Cuándo leer cada referencia lo dice el cuerpo…», el párrafo literal de contracts/guia-y-changelog.md §1 —lo medido en los sondeos de la eval 15, con sus cifras (ninguna de las 40 sesiones de `brazo-D`; `normas`, de 23 a 40 sesiones por brazo; `con_herramientas`, de 6 a 9) y la ruta de su evidencia, y sus dos consecuencias— sin tocar nada más de la guía (FR-012); y en CHANGELOG.md, en «### Cambiado» de *Unreleased*, detrás de la entrada de `boe-legislacion` v0.3, la entrada v0.3.1 literal de contracts/guia-y-changelog.md §2, con la cifra del cuerpo que da quickstart.md escenario 1 tras T003 (FR-014). Test: quickstart.md escenario 8 da `1`, `1` y `0`; `make ci` en verde (FR-012, FR-014, SC-001, SC-006, SC-013)

**Checkpoint**: todo lo del hito escrito; `make ci` en verde. El cierre lo hace el workflow tras la revisión final.

---

## Phase 6: US4 y US6 — El cierre y el sondeo (Priority: P2)

Sin tareas, por diseño. US4 es el job de evals que el workflow lanza en el cierre sobre la cabeza de la rama, con los
umbrales de hoy (SC-007 a SC-011); su escenario 4, `TestDefinicionDelJob`, ya está en `make ci` y T002 lo ejecuta tras
añadir las lecturas (FR-013: el tope no cambia, research D10). US6 es el sondeo de la eval 15 con la skill del hito,
fuera del run (ADR 0032; SC-012): el informe final lo da como pendiente. Ninguna tarea publica la rama, mide CI o evals
remotas ni depende de nadie.

---

## Dependencies & Execution Order

### Orden de las fases

- **T001** (`[datos]`) no depende de nada. Tras su commit, el paso `grabar_datos` graba los ocho bloques.
- **T002** depende de T001 y de su grabación: sin ella, `grabado` y la preparación de la caché salen en rojo (research
  D8, V7).
- **T003** depende de T002: la regla nueva pide leer en las sesiones de las evals 14, 15 y 17 lo que T002 hace servir
  sin red; el `make ci` de T003 no lo necesita, pero el cierre mide esa skill sobre esa caché.
- **T004** depende de T003 (la cifra del cuerpo de la entrada de `CHANGELOG.md`) y de T002 (la entrada nombra las
  lecturas).

### Dentro de cada historia

- US3: el manifiesto (T001), la grabación del workflow, y los tests en rojo antes que las evals (T002).
- US1 y US2: una tarea con el texto entero del contrato, que el gate del presupuesto acota de una vez.
- US5: el párrafo de la guía y la entrada del CHANGELOG, documentación sin código.

### Oportunidades de paralelismo

Ninguna: las cuatro tareas forman una cadena (arriba). El guardián las ejecuta una a una.

---

## Implementation Strategy

### MVP (US3 + US1 y US2)

T001 → grabación → T002 → T003. Con eso la skill nueva está entera y las evals por materia sirven sin red lo que pide
leer: el job del cierre ya puede medir SC-007 a SC-011.

### Entrega incremental

T004 añade la guía y el CHANGELOG. Cada tarea deja `make ci` en verde.

---

## Trazabilidad

### Cada requisito y sus tareas

| Requisito | Tarea |
|---|---|
| FR-001 | T003 |
| FR-002, FR-003, FR-004 | T003 |
| FR-005, FR-006, FR-007, FR-008 | T003 |
| FR-009 | T002 |
| FR-010 | T001 (manifiesto) y el paso `grabar_datos` (grabaciones) |
| FR-011 | T002 |
| FR-012 | T004 |
| FR-013 | T002 (ejecuta `TestDefinicionDelJob` con las lecturas; ni el tope ni la definición cambian, plan y research D10) |
| FR-014 | T004 |
| SC-001 | todas |
| SC-002, SC-003 | T003 (SC-003 lo coteja la revisión final) |
| SC-004, SC-008 | T001, T002 |
| SC-005 | T002 |
| SC-006 | T004 (la entrada); el job del cierre, del workflow |
| SC-007, SC-009, SC-010, SC-011 | T003 entrega lo que mide el job del cierre |
| SC-012 | ninguna: el sondeo, fuera del run (US6) |
| SC-013 | T004 (la revisión final lo coteja con `LEEME.md`) |

### Cada fila de «## Controles de umbral» de plan.md y la tarea que construye su control

Ningún control es nuevo: el hito no añade tests ni umbrales (plan, «Controles mecánicos»), salvo las filas del mapa y de
la tabla de los dos tests de las lecturas (T002). Cada control ya sale en rojo cuando su medida pasa del umbral; la
columna dice qué tarea lo ejerce sobre la entrega.

| Requisito de la fila | Control | Tarea |
|---|---|---|
| SC-001 | `ci:Makefile:ci` | todas |
| FR-005, SC-002 | `TestSkillsDelRepositorio`, `TestCuerpoDeLaSkill` (presupuesto del cuerpo; rojo con 11 001, ya probado sobre copias de 11 000 y 11 001) | T003 |
| FR-006 | `TestSkillsDelRepositorio` (referencia sin enlace) | T003 |
| FR-004, FR-007 | `TestSkillsDelRepositorio` (`sin-instrucciones-de-evals`, `normas-nombradas`) | T003 |
| FR-006, FR-007, SC-001 | `TestEvalsDelRepositorio` (los ocho controles de `docs/SKILLS.md` §1) | T003 |
| FR-009, SC-004 | `TestEvalsDelRepositorio/lecturas` (mapa escrito a mano: rojo con una lectura de más, de menos o con otro motivo) | T002 lo cambia y lo ve en rojo antes de las evals |
| FR-010, FR-011, SC-004 | `TestEvalsDelRepositorio/grabado` (rojo con una lectura sin grabar, research V7) | T001 (manifiesto), T002 (lecturas) |
| FR-011, SC-004 | `TestCacheDeLasSesiones/lecturas` (rojo si un bloque no sale con 0 desde la caché de su eval) | T002 lo cambia y lo ve en rojo antes de las evals |
| FR-010, FR-011 | `TestLecturaDeLaEvalSinGrabar`, sin cambios | T002 lo ejecuta con las lecturas nuevas |
| SC-008 | las tres filas de arriba | T001, T002 |
| FR-013, SC-005 | `TestDefinicionDelJob` (`fallosDelTope`) | T002 lo ejecuta; sin cambios (research D10) |
| SC-007 | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5` | el de hoy, sin cambios; T003 entrega lo que mide el job del cierre |
| SC-009 | los umbrales que deciden de `boe-legislacion` en los dos modos | los de hoy; T003 |
| SC-011 | los umbrales que deciden de `jurisprudencia` y `legal-core` | los de hoy; ninguna tarea toca esas skills |

## Definition of Done (§1) que aplica

- **1 · `make ci` en verde**: cada tarea; T004, la última.
- **2 · Tests offline y fixtures grabados**: T002 (los dos tests de las lecturas, sin red); las ocho grabaciones las
  escribe el paso `grabar_datos` tras T001.
- **3, 4, 5, 12 · paquetes autorizados, esquemas, errores tipados, grafo**: no aplica; ningún código de producción,
  esquema, código de salida ni operación de grafo cambia (plan, «Reglas de dependencia»).
- **6 · Comportamiento visible y `CHANGELOG.md`**: el binario no cambia (sin e2e, research D13); la skill sí, y su entrada
  de *Unreleased* la escribe T004.
- **7 · ADR**: ninguno nuevo; ninguna decisión de arquitectura cambia (spec, «Fuera de alcance»: ni el ADR 0046 ni
  umbrales).
- **8 · `docs/SOURCES.md`**: no aplica; la fuente `boe.legislacion-consolidada` ya tiene fila revisada en `main` y su uso
  no cambia (plan, «Datos externos»).
- **9 · Cobertura**: sin código nuevo.
- **10 · Skill y evals**: la skill según `docs/SKILLS.md` (T003, con los gates); sin evals nuevas («Fuera de alcance»),
  las `lecturas` de tres evals (T002); las mide el job del cierre.
- **11, 13 · Dimensión territorial, plazos y cobertura**: no aplica; la skill no calcula plazos ni recursos, y la regla 5
  («Ningún caso especial para un territorio») no cambia.
- **Schemas**: no aplica; ningún esquema cambia.
- **Documentación**: `docs/SKILLS.md` §1 (T004) y `CONTRIBUTING.md` (T003).

## Notes

- `make ci` en una sesión de paso: si supera el tope de una orden, por objetivos, en primer plano y uno detrás de otro.
- Ninguna tarea toca `.github/workflows/evals.yml`, `internal/evals/definicion.go`, `jurisprudencia`, `legal-core`,
  `scripts/workflow/`, `.specify/`, `internal/workflow/` ni `web/` (spec, «Fuera de alcance»).
- Los comentarios de `internal/evals/prohibidas_test.go` que dicen «la frase de la regla 7» no se tocan (research D7).
