# Specification Analysis Report — H0 · Esqueleto del repo y gates de CI

**Modo**: desatendido (sin preguntas, sin remediación diferida). Generado por `/speckit-analyze`.
**Artefactos analizados**: `spec.md`, `plan.md`, `tasks.md` (+ `research.md`, `data-model.md`, `contracts/`, `quickstart.md`, `checklists/requirements.md`), `.specify/memory/constitution.md`.

## Hallazgos

| ID | Categoría | Severidad | Ubicación(es) | Resumen | Recomendación |
|----|----------|----------|-------------|---------|----------------|
| F1 | Inconsistency | HIGH | `plan.md:187,267`; `data-model.md:230`; `tasks.md:34,162-174` | `plan.md` (Project Structure y «Orden de implementación» paso 3) y `data-model.md` (inventario de artefactos de SC-008) prescriben crear `.gitleaksignore` vacío con cabecera como artefacto de H0 mapeado a FR-018. `tasks.md` decide explícitamente lo contrario en su sección «Notas»: *«No se crea en H0 un fichero de exclusiones vacío: sin ningún falso positivo que excluir sería exactamente el marcador de posición sin contenido que SC-008 prohíbe»*. Ninguna tarea (T001–T014) declara esa ruta. La enumeración literal de SC-008 en `spec.md:191` tampoco incluye `.gitleaksignore` entre los 18 ficheros exigidos ni entre las excepciones permitidas (FR-005, FR-040, FR-042). Como el guardián de diff del workflow `hito` rechaza cualquier fichero fuera de las rutas que la tarea declara, si el ejecutor sigue el «Orden de implementación» de `plan.md` al pie de la letra crearía un fichero que el guardián bloquearía; si sigue `tasks.md` (la fuente que ejecuta el workflow), el fichero nunca se crea. La contradicción debe resolverse en el artefacto, no en tiempo de ejecución. | Corregir `plan.md` (eliminar `.gitleaksignore` de «Project Structure» y del paso 3 de «Orden de implementación», o marcarlo explícitamente como diferido) y la fila de `data-model.md:230` para que coincidan con la decisión ya tomada en `tasks.md`: el fichero nace con la primera huella real que haya que excluir, no en H0. |
| F2 | Coverage / Inconsistency | MEDIUM | `tasks.md:111-128` (tabla «Requisitos → tareas») | La tabla de trazabilidad de `tasks.md` no menciona **SC-004** (equivalencia local↔CI) ni **SC-008** (inventario completo de artefactos) en ninguna fila, a diferencia del resto de criterios de éxito (SC-001 a SC-003, SC-005 a SC-007, SC-009 a SC-012, todos listados). Ambos criterios sí están funcionalmente cubiertos: `quickstart.md:281-299` los asigna a los escenarios 3, 9, 10 y 12, y T013 ejecuta exactamente esos escenarios (T013 dice «escenarios 1 a 10 y 12»). No hay una brecha de cobertura real, pero la tabla de trazabilidad —el artefacto pensado para auditar cobertura de un vistazo— queda incompleta y podría inducir a pensar que SC-004/SC-008 no tienen tarea asociada. | Añadir una fila (o ampliar la existente de T013) que indique explícitamente «SC-004, SC-008 → T013 (vía escenarios 3, 9, 10, 12 de quickstart.md)», para que la tabla sea autocontenida. |
| F3 | Inconsistency (menor) | LOW | `plan.md:286` («Complexity Tracking») vs. tarea T001 en `tasks.md:34` | La fila de *Complexity Tracking* sobre las órdenes adicionales del `Makefile` dice *«Siete órdenes del Makefile más allá de las doce de FR-006 (`fmt-check`, `lint-fast`, `secrets`, `mod-verify`, `mod-tidy-check`, `check-tools`, `hooks`)»* — siete nombres. La descripción de T001 añade además `help` como objetivo del `Makefile`, no contabilizado en esa lista ni justificado aparte. El objetivo en sí no es problemático (es un target de ayuda habitual, no un control), pero el recuento «siete» queda desactualizado frente a las ocho órdenes añadidas realmente. | Ajustar el texto de `plan.md` a «ocho órdenes» e incluir `help` en la enumeración, o anotar explícitamente que `help` queda fuera del recuento por no ser un control sujeto a justificación de complejidad. |

No se han encontrado violaciones de principios `MUST` de la constitución sin justificar: las seis divergencias registradas en *Complexity Tracking* de `plan.md` están motivadas con alternativa rechazada, tal y como exige el «Criterio de decisión autónoma» §3, y la desviación de `misspell` (sin diccionario español, pese a exigirlo FR-013 literalmente) está registrada como límite conocido de la herramienta con mitigación (`misspell.ignore-rules` caso a caso), lo que cumple la vía de registro que la constitución permite en `plan.md → Complexity Tracking`.

## Tabla de cobertura (resumen)

| Bloque de requisitos | ¿Tiene tarea? | Tarea(s) |
|---|---|---|
| FR-001 a FR-014, FR-017, FR-037 a FR-042 | Sí | T001 |
| FR-015, FR-016, FR-018 a FR-020 | Sí (ejecutados por T001, ejercidos por T004) | T001, T004 |
| FR-021, FR-032 | Sí | T010 |
| FR-022 | Sí | T003 |
| FR-023 | Sí | T012 |
| FR-024, FR-033 a FR-036 | Sí | T011 |
| FR-025 | Sí | T004 |
| FR-026 | Sí | T006 |
| FR-027 | Sí | T005 |
| FR-028 | Sí | T007 |
| FR-029 | Sí | T002 |
| FR-030 | Sí | T008 |
| FR-031 | Sí | T009 |
| SC-001, SC-002, SC-003, SC-007, SC-010, SC-011, SC-012 | Sí | T013 |
| SC-004, SC-008 | Sí (no reflejado en la tabla de `tasks.md`, ver F2) | T013 |
| SC-005, SC-006, SC-009 | Sí | T014 |

**Todas las 42 FR y los 12 SC tienen al menos una tarea asociada.** No hay requisitos huérfanos ni tareas sin requisito asociado (T001–T014 mapean todas a al menos un FR o SC).

## Hallazgos de alineación con la constitución

Ninguno bloqueante. Las siete divergencias registradas (`tools/` multi-módulo, ausencia de `go.sum` raíz, `gofumpt` sin módulo propio, siete/ocho órdenes extra del `Makefile`, `misspell` sin español, `go mod tidy -diff` solo en el módulo raíz, y la de F1 en cuanto se corrija) están o deberían estar en *Complexity Tracking* con alternativa rechazada y motivo, conforme exige «Criterio de decisión autónoma» §3 de la constitución.

## Tareas no mapeadas

Ninguna. Las 14 tareas (T001–T014) tienen requisito o criterio de éxito asociado en la tabla de trazabilidad de `tasks.md` (con la salvedad documental de F2).

## Métricas

- **Total Functional Requirements**: 42 (FR-001–FR-042, sin huecos)
- **Total Success Criteria**: 12 (SC-001–SC-012)
- **Total Tasks**: 14 (T001–T014)
- **Cobertura de requisitos (≥1 tarea)**: 100 % (42/42 FR, 12/12 SC; 2 de los SC solo por ejecución transitiva vía escenarios de quickstart, no reflejada en la tabla de `tasks.md` — F2)
- **Ambigüedad**: 0 hallazgos (adjetivos vagos ya acotados con SC medibles; sin marcadores `TODO`/`TBD` pendientes)
- **Duplicación**: 0 hallazgos
- **Incoherencias (Inconsistency)**: 2 (F1 HIGH, F3 LOW) + 1 de cobertura documental (F2 MEDIUM)
- **Issues críticos (CRITICAL)**: 0
- **Issues HIGH**: 1 (F1)
- **Issues MEDIUM**: 1 (F2)
- **Issues LOW**: 1 (F3)

## Próximas acciones

No hay hallazgos CRITICAL que bloqueen `/speckit-implement`. Antes de ejecutar el hito con el workflow `hito`, corregir F1 es recomendable pero no bloqueante por sí solo, **porque `tasks.md` —la fuente real que ejecuta el workflow— ya tiene la decisión correcta** (no crear `.gitleaksignore` en H0); el riesgo de F1 es que quien lea o edite `plan.md` reintroduzca el fichero fuera del guardián de diff. Dado el modo desatendido de este comando, no se ofrece plan de remediación diferido ni se pregunta: se deja constancia aquí para que la siguiente pasada de `/speckit-plan` o una edición manual de `plan.md`/`data-model.md` alinee esas dos líneas con `tasks.md`.

- Editar `plan.md:187` y `plan.md:267` para retirar `.gitleaksignore` de H0 (F1).
- Editar `data-model.md:230` para retirar esa fila del inventario de SC-008, o anotarla como «diferido a la primera huella real» (F1).
- Ampliar la tabla «Requisitos → tareas» de `tasks.md` con SC-004 y SC-008 → T013 (F2).
- Corregir el recuento «siete» → «ocho» (o justificar la exclusión de `help`) en `plan.md:286` (F3).

Ninguna de estas correcciones toca alcance, frontera humana, privacidad, TOS ni una decisión ya cerrada; son ajustes de coherencia documental dentro del criterio §3 («decisiones con rastro»), no requieren escalar a humano.
