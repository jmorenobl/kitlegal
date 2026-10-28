## Specification Analysis Report — H31.1 (029-h31-1-boe-legislacion-habla)

Análisis de solo lectura de `spec.md`, `plan.md` y `tasks.md` (con `research.md`, `contracts/` y la constitución como contexto). Lo que no he podido medir en esta sesión (sin modelo, sin `make ci`, sin ejecutar los guiones) se dice como tal.

| ID | Category | Severity | Location(s) | Summary | Recommendation |
|----|----------|----------|-------------|---------|----------------|
| C1 | Coverage | LOW | tasks.md «Historias sin tarea» (US5, US7) | US5 (SC-007, SC-009 a SC-011) y US7 (SC-012) no tienen tarea. Está justificado: los mide el cierre del workflow y decide la persona antes de fusionar (ADR 0018), y ningún umbral cambia. | Ninguna acción. Queda explícito en tasks.md. |
| I1 | Inconsistency | LOW | tasks.md T005/T006 `[P]` frente a «Dependencias y orden» | T005 y T006 llevan `[P]` pero el fichero dice que se ejecutan igualmente en orden. Los ficheros no se solapan (`docs/SKILLS.md`, `CHANGELOG.md`), así que no hay riesgo. | Ninguna acción; el marcador solo indica independencia. |
| U1 | Underspecification | LOW | tasks.md T004 | T004 agrupa cuatro cambios: `SKILL.md`, dos referencias y `CONTRIBUTING.md`. Es consecuencia de FR-001 y de research D10, y las tres primeras son copias byte a byte; el cuarto se cubre solo con `make ci`. | Ninguna acción. Es una rebanada coherente y todas las rutas están declaradas. |
| U2 | Underspecification | LOW | tasks.md T004, `CONTRIBUTING.md:742-743` | El cambio «la frase de la regla 7» → «la frase de `references/lecturas_que_fallan.md`» no tiene un control propio. Ningún guion ni test lo ve. | Ninguna acción obligatoria. La revisión final puede cotejarlo con research D10. |

### Verificado en esta sesión

- La cuenta de caracteres del cuerpo de `contracts/skill-md-v0.3.1.md` (lo que sigue al segundo `---`, contado con `wc -m`) es **10 996**, igual que dice tasks.md T004 y por debajo de 11 000 (FR-005). Contado a mano, no con el gate `TestCuerpoDeLaSkill`, que no he ejecutado.
- D10 es correcta. En `skills/boe-legislacion/SKILL.md` de `main` la regla 7 es «Si `kitlegal graph check` no termina con `0`» (hay 8 reglas). En el contrato de la skill medida es «Sin herramienta y sin binario». La frase «No se ha podido comprobar si la redacción…» está hoy en `references/lecturas_que_fallan.md:46`. El cambio en CONTRIBUTING es necesario.
- Los ficheros que T001 a T006 citan existen: `contracts/{aceptacion,lecturas-y-grabaciones,skill-md-v0.3.1,skill-boe-legislacion-v0-3-1,guia-y-changelog}.md`; el ancla `Las dos fuentes` en `docs/SKILLS.md:22`; las líneas 742-743 de `CONTRIBUTING.md`.
- Las etiquetas de tarea `[aceptacion]` y `[datos]` aparecen una sola vez cada una en tasks.md (líneas 27 y 39), en su propia tarea. Ninguna otra tarea las contiene en el texto.
- Los tres guiones de T001 coinciden con los nombres de `contracts/aceptacion.md`.

### Coverage Summary Table

| Requirement Key | Has Task? | Task IDs | Notes |
|---|---|---|---|
| FR-001 (parte de la skill medida) | Sí | T004 | Copia byte a byte de `SKILL.md` y de las dos referencias. |
| FR-002 (pasajes medidos) | Sí | T001, T004 | Guion `h31-1-responder` y `h31-1-lee-lo-que-nombra`. |
| FR-003 (formas fijas) | Sí | T001, T004 | Guion `sitio-en-el-cuerpo` y los gates `linea-sin-consulta` y similares. |
| FR-004 (tabla de comandos) | Sí | T001, T004 | `make skills-check`. |
| FR-005 (≤ 11 000) | Sí | T004 | `TestCuerpoDeLaSkill`. |
| FR-006 (lo que toda respuesta necesita) | Sí | T001, T004 | Encabezados y reglas contra la medida. |
| FR-007 (salida de texto a referencia) | Sí | T001, T004 | Gate `referencias`. |
| FR-008 (controles de contenido) | Sí | T004 y todas por `make ci` | |
| FR-009 (genericidad) | Sí | T001, T004 | Guion `sitio-en-el-cuerpo`. |
| FR-010 (frontmatter) | Sí | T001, T004 | |
| FR-011 (nueve lecturas) | Sí | T003 | |
| FR-012 (manifiesto y grabación) | Sí | T002, `grabar_datos` | Ocho bloques; `a24` ya está grabado. |
| FR-013 (caché sirve los bloques) | Sí | T003 | |
| FR-014 (guía) | Sí | T005 | |
| FR-015 (tope del job) | Sí | T003 | Plan: 21 120 s frente a 20 117 s; sin cambio de tope. |
| FR-016 (CHANGELOG) | Sí | T006 | |
| SC-001 a SC-006, SC-008, SC-013 | Sí | T001 a T006 | Según la tabla «Controles de umbral» de tasks.md. |
| SC-007, SC-009, SC-010, SC-011 | No (por diseño) | — | Los mide el cierre del workflow. |
| SC-012 | No (por diseño) | — | La decide una persona antes de fusionar. |

### Constitution Alignment Issues

Ninguno. Contrastado con los principios III (aceptación primero, congelada, offline, fixtures por `grabar_datos`), V, VIII y X (genericidad por materia: FR-009 lo hace comprobable), y con «Flujo de trabajo y gates» (suite de aceptación escrita primero; `make ci` por tarea). La checklist `checklists/requirements.md` está completa: 16 casillas marcadas y 0 pendientes.

### Unmapped Tasks

Ninguna. Cada tarea cita FR, US y SC.

### Metrics

- Total Requirements: 16 FR + 13 SC (de ellos, 5 SC de cierre o de decisión humana sin tarea por diseño)
- Total Tasks: 6
- Coverage % (FR con ≥ 1 tarea): 100 % (16/16)
- Ambiguity Count: 0 (sin `TODO`, `TKTK` ni `NEEDS CLARIFICATION` en spec.md; no he buscado adjetivos vagos más allá de una lectura de FR/SC, todos con criterio medible)
- Duplication Count: 0
- Critical Issues Count: 0 (HIGH: 0 · MEDIUM: 0 · LOW: 4)

### Next Actions

- No hay CRITICAL ni HIGH: se puede pasar a `/speckit-implement`.
- Los cuatro hallazgos son LOW y no exigen cambios. Ninguno se aplica ni se deja para después.
