# Specification Analysis Report: H31 · Las skills siguen la especificación Agent Skills

Artefactos leídos: `spec.md` (555 líneas), `plan.md` (346) y `tasks.md` (310), completos; `research.md` solo D19 y V6;
`contracts/`, `data-model.md` y `quickstart.md` no se han vuelto a leer (los resume el plan y los citan las tareas).
Constitución: no releída en esta sesión; los principios I a X se contrastan con la tabla «Constitution Check» del plan,
no con el texto de la constitución. Solo lectura: ningún fichero salvo este informe.

Comprobado contra el repositorio en esta sesión (existen y tienen la forma que las tareas suponen): `evidencias/adr-0046/`
(`rubrica.md`, `esquema.json`, `casos.yaml`, `medida.json`, `manifiesto.json`, `guiones/juez.py`), `skills.ContarLineas`,
`skills.LeerFrontmatter`, `skills.ValidarFrontmatter`, `skills.DatosDeLaReferencia`, `skills.AppletsDeLaTabla`,
`maximoDeLineas = 299`, `nombra`, `calculoDeDependencias.skillsDeLaEval` y `.datosDeLasReferencias`, `cacheDeUnaSesion`,
`ComprobarSinRed`, `fraseEsta`, `esBlanco`, `nuevaVotacion` y `(*votacion).juzgar`, `nuevoVotanteDelGuion`,
`comprobarLaMedida`, `fijadosParaElJuez`, `leerMedidaDelJuez`, `cambiaLaDescripcion`, `peorCasoDelTrabajo`,
`TestOrdenesParaPowerShell`, `TestSkillsEmpotradas`, `TestJuezDeLaDefinicionDelJob`, `TestCopiasDelJuez`,
`TestDecisionDelTrabajo`, `copiaConLaSkill`; `scripts/evals-voto.sh`, `scripts/evals-medir-juez.sh`, `scripts/evals.sh`;
`MODELO_DEL_JUEZ: claude-opus-5-5`, `VERSION_DE_CLAUDE_CODE_DEL_JUEZ: 2.1.289` y `timeout-minutes: 352` en
`.github/workflows/evals.yml`; 45 evals (46 YAML en `evals/*/` menos `expresiones-prohibidas.yaml`); `medida.json` con
63 defectos y 36 correctos. `docs/SKILLS.md` no existe, como debe (lo crea T018).
No comprobado: el cuerpo de `TestSkillsEmpotradas` (se fía de research V11), los contratos y el contenido de las tablas de
reglas de `research.md`, y ninguna cifra de tiempo del job (19 877 s → 20 117 s) más allá del tope de 352 min.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| F1 | Inconsistencia | MEDIUM | tasks.md T003 (L73), T004 (L74), T007 (L103), T008 (L116), T015 (L148) | Cinco líneas de tarea citan en la prosa `evals/boe-legislacion/expresiones-prohibidas.yaml`, `data/normas.yaml`, `evals/<skill>/` y `evals/<skill>/juez/`. El extractor de rutas del workflow declara toda ruta que aparece en la línea, así que el guardián admitiría cambios en `data/` y `evals/`, cuando el spec («Fuera de alcance») y las notas de tasks.md dicen que ninguna tarea los toca. | Nombrarlos sin directorio en la prosa (`expresiones-prohibidas.yaml`, `normas.yaml`, «la carpeta de evals de la skill»), como pide la memoria `etiqueta-literal-en-texto-de-tarea`. |
| F2 | Inconsistencia | LOW | research.md D19 y V6; tasks.md T007 (L103) | Dan `boe-legislacion` v0.2.2 como la versión de hoy; `CHANGELOG.md:364` registra la v0.2.3 («`boe-legislacion` v0.2.3, `legal-core` v1.2.1 y `jurisprudencia` v0.1.2 … Sustituye a `boe-legislacion` v0.2.2»). La v0.3 sigue estando por encima, así que FR-023 se cumple igual. | Cambiar «de v0.2.2» por «de v0.2.3» en T007 y en D19/V6, para que la entrada de `CHANGELOG.md` diga «Sustituye a v0.2.3». |
| F3 | Inconsistencia | LOW | spec.md FR-002 (L247-250) frente a tasks.md T009 (L129) | FR-002 mantiene «descripción con como mucho 1024 caracteres» entre las reglas de hoy y FR-005 añade 900; T009 dice que el 900 «sustituye al máximo de 1024». Son compatibles (900 implica 1024), pero el texto del defecto que sale de `ValidarFrontmatter` queda sin decir cuál de los dos da. | Decirlo en T009: con más de 900 el defecto es el de FR-005; el de 1024 desaparece por redundante. |
| F4 | Cobertura | LOW | tasks.md T003 (L73) | El subtest nuevo `lineas-de-aviso-de-las-referencias` se aplica a `boe-legislacion` y `legal-core`, no a `jurisprudencia`, cuyo `⚠ SENTENCIA NO COMPROBADA:` también debe quedarse en el cuerpo. T005 exige «solo en el cuerpo», pero no hay un control que lo haga cumplir si una referencia de `jurisprudencia` lo repite mal. No he comprobado a qué skills se aplican hoy los controles `avisos-de-la-skill`, así que puede estar cubierto por otro. | Confirmar a qué skills aplica `avisos-de-la-skill`; si `jurisprudencia` queda fuera, ampliar el subtest nuevo a las tres o anotar por qué no. |
| F5 | Cobertura | LOW | tasks.md, trazabilidad de FR-015 | FR-015 no tiene tarea propia: se apoya en `TestSkillsEmpotradas` sin cambios (research D20, V11). Existe (`skills_test.go:50`), pero no he leído su cuerpo para ver que falla con una referencia que falta. | Aceptable por el umbral de materialidad; la revisión final puede mirar el cuerpo del test una vez. |

Sin hallazgos de severidad CRITICAL ni HIGH. Sin duplicaciones, marcas de aclaración pendientes ni `TODO` en los tres
artefactos (grep de `TODO`, `TKTK`, `???`, `NEEDS CLARIFICATION` no ejecutado: lo dicho sale de la lectura, no de una
búsqueda).

## Cobertura de requisitos

| Requisito | ¿Tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001 | Sí | T018 | |
| FR-002, FR-003, FR-005 | Sí | T009 | F3 |
| FR-004, FR-010 | Sí | T010 | |
| FR-006 a FR-009 | Sí | T011 | |
| FR-011 | Sí | T009, T010, T011 | |
| FR-012, FR-013 | Sí | T001, T011, T005–T007 | |
| FR-014 | Sí | T002 | Cierre por nombre en referencias a mano (Q1) |
| FR-015 | Sí (indirecta) | T005–T007 | F5 |
| FR-016, FR-017, FR-018 | Sí | T003, T004 | F4 |
| FR-019 | Sí | T007, T008 | |
| FR-020, FR-021, FR-022, FR-024 | Sí | T005, T006, T007 | La revisión final contrasta FR-021 y FR-024 |
| FR-023 | Sí | T005–T007, T011, T017, T018 | F2 |
| FR-025, FR-027 | Sí | T012 | |
| FR-026 | Sí | T012, T015 | |
| FR-028, FR-029 | Sí | T013 | |
| FR-030 | Sí | T002, T014, T016 | |
| FR-031, FR-032 | Sí | T013, T015 | |
| FR-033 | Sí | T012, T013, T015, T017 | |
| FR-034 | Sí | T017 | La medida la lanza una persona (SC-011) |
| SC-001 a SC-010, SC-014, SC-015 | Sí | según la tabla de trazabilidad de tasks.md | |
| SC-011, SC-012, SC-013 | Parcial por diseño | T017, T005–T007, T018 | Las contrasta una persona o la revisión final |

## Alineación con la constitución

Sin conflictos que encontrar en lo leído. «Aceptación e2e: no aplica» está declarada y justificada en Complexity Tracking
(el hito no cambia ningún comportamiento del binario); las tareas respetan el modo desatendido (sin red, sin sesiones
con modelo, el juez solo en el job y en la medida) y no tocan `scripts/workflow/`, `.specify/` ni `internal/workflow/`.
Salvedad: no he releído el texto de la constitución en esta sesión.

## Tareas sin requisito

Ninguna: las 18 nombran los FR y SC que cumplen.

## Orden y dependencias

Sin contradicciones: la regla de `references/` (T001–T002) y los controles sobre la skill entera (T003–T004) van antes de
mover texto (T005–T007); el control de órdenes de ejemplo (T008) va después de T007, que cambia los dos ejemplos que la
caché no sirve; los gates de forma (T009–T011) van cuando las tres skills ya los cumplen; el juez (T012–T017) comparte
ficheros y va en secuencia; la guía (T018) cierra. Ninguna línea de tarea lleva `[datos]`, `[plataforma]` ni `[aceptacion]`.

## Métricas

- Requisitos: 34 FR y 15 SC (49); tareas: 18.
- Cobertura (requisitos con al menos una tarea): 100 % de los FR; de los SC, los tres de revisión humana (SC-011 a SC-013)
  quedan por diseño fuera del run.
- Ambigüedades: 0 abiertas; duplicaciones: 0.
- Hallazgos: CRITICAL 0 · HIGH 0 · MEDIUM 1 · LOW 4.

## Próximos pasos

No hay CRITICAL ni HIGH: se puede pasar a `/speckit-implement`. Antes, F1 es una edición de `tasks.md` de unas pocas
palabras y evita que el guardián declare `data/` y `evals/` como escribibles; F2 es una cifra.
