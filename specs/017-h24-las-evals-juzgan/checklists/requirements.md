# Specification Quality Checklist: H24 · Las evals juzgan el significado con un modelo: «afirma lo que no ha leído» decide, `boe-legislacion` deja de glosar lo que no ha leído, y la lista de expresiones deja de decidir

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-05
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué recibe el juez, qué regla decide y qué ve quien lee el informe. Los nombres de ficheros, tests, umbrales y etiquetas que aparecen son los que fija el hito o existen en el repositorio; el formato de la declaración de clases, las claves del informe, el tope de un voto y cómo se graban los votos quedan para el plan (Assumptions, última viñeta).
- [x] Focused on user value and business needs — Resumen y US1: lo que lee quien pregunta; US2 a US7: el instrumento que lo mide y quien lo lee.
- [x] Written for non-technical stakeholders — cada historia empieza por lo que ve una persona; el detalle técnico está en los requisitos, como en H7.4 y H21.
- [x] All mandatory sections completed — User Scenarios, Requirements, Success Criteria, Assumptions y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — la de FR-013 (el efecto de una marca del juez en la eval sin binario ni servidor) se cerró en `## Clarifications`: solo se publica, sin umbral ni motivo de fallo, y a «Fuera de alcance» va que decida.
- [x] Requirements are testable and unambiguous — cada FR lleva «Comprobable», un **Control** o remite a un control de FR-100 a FR-113; los de la skill (FR-081 a FR-084) los miden los umbrales de FR-030 y FR-031 en el job de cierre. El efecto de FR-013 quedó fijado en `## Clarifications`, con su comprobable en FR-104 y SC-004: el caso de una respuesta de esa eval marcada con tres síes.
- [x] Success criteria are measurable — SC-001 a SC-013 con recuentos, segundos o número de casos; SC-014 es la parte humana de la Aceptación.
- [x] Success criteria are technology-agnostic (no implementation details) — miden respuestas, votos, casos y umbrales; los controles nombran la comprobación que falla, como exige el ADR 0029.
- [x] All acceptance scenarios are defined — US1 a US7, en forma Dado/Cuando/Entonces (31 escenarios).
- [x] Edge cases are identified — «Edge Cases»: sí, sí y no; el nulo repetido; el voto que no llega a darse; la sesión sin textos; un modo marcado y otro no; la versión de las sesiones que sube; el aviso correcto.
- [x] Scope is clearly bounded — «Fuera de alcance», con las ocho viñetas literales del hito y lo no especificado.
- [x] Dependencies and assumptions identified — «Assumptions» y «Relación con H5.1 a H22».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — tabla de trazabilidad de «Criterios del hito, literales» y controles FR-100 a FR-113.
- [x] User scenarios cover primary flows — la respuesta (US1), el juez y su regla (US2), la medida versionada (US3), su ejecución a petición (US4), el informe (US5), la lista (US6) y el sondeo (US7).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 y SC-014 recogen la Aceptación literal; SC-002 a SC-013, los Controles.
- [x] No implementation details leak into specification — ver el primer ítem.

## Criterios del juez del spec (workflow `hito`, `juez_spec`)

- [x] a. Alcance: cubre Entrega (1) a (7) y nada más (tabla de trazabilidad). No especifica caso a caso ficheros manipulados a mano: una declaración de clases mal formada es un fichero mal formado, como una eval (FR-020), y una copia que difiere hace fallar `make ci` (FR-023).
- [x] b. Fuera de alcance: las ocho viñetas literales del hito, comprobadas contra `docs/ROADMAP.md` con `diff`, y lo no especificado (el modelo informativo, el efecto del voto en la serie, lo que el sondeo no hace, grabar respuestas nuevas).
- [x] c. Clarificaciones: las tres respuestas de `gates/clarify-respuestas.json` (Q1 a Q3) están en `## Clarifications` y aplicadas en FR-005, FR-013, FR-043, FR-051, FR-052, «Edge Cases» y «Fuera de alcance», y lo que cada una pide a los tests, en su control: Q1 en FR-104 y SC-004, Q2 en FR-102, y Q3 en FR-106, SC-006 y el escenario 4 de US4.
- [x] d. Aceptación: «Criterios del hito, literales» lleva Entrega, Controles y Aceptación transcritos, comprobados contra `docs/ROADMAP.md` con `diff`; SC-001 a SC-014.
- [x] e. Decisiones cerradas: ni el modelo que decide, ni las repeticiones, ni la regla por serie, ni el binario; el juez es el del ADR 0037 (modelo distinto, tres votos unánimes, medido, nunca en un paso del run: FR-010, FR-041, FR-090, FR-093).
- [x] f. Checklist veraz: cada marca cita dónde se cumple, y no queda ningún ítem sin marcar.
- [x] g. Sin implementación: las formas concretas son del plan; `voto_real`, `prompt_de` y `esquema.json` los nombra el hito.
- [x] h. Derivable: la regla de los votos con sus cuatro casos (FR-010, FR-011, FR-103), el voto nulo y el que no llega a darse (FR-006, FR-007), la medida que corresponde y la que se cumple (FR-041), los umbrales con medida y total (FR-030 a FR-033) y las dos frases de v0.1.6 con las que falla la comprobación de la prosa (FR-085).
- [x] i. Uso: «Uso, de fuera adentro», con quién pide cada salida, su tamaño con su cálculo y cuándo deja de darse cada señal.
- [x] j. Umbral con control: FR-030, FR-032 y FR-033 nombran el job en rojo, y FR-032 además `make ci`; FR-086 y FR-092, su comprobación de `make ci`; SC-001 a SC-013 llevan **Control**; FR-101 fija las seis filas de «Controles de umbral». `cuenta_su_proceso` no es un requisito con umbral: el hito lo deja publicado y que decida está fuera de alcance (FR-031).

## Notes

- Las tres ambigüedades del paso de clarify (efecto de una marca en la eval 21, comprobación de la frase, y qué hace la ejecución de la medida con la medida versionada) están cerradas en `## Clarifications`. Las demás lecturas que el spec fija salen del hito, del ADR 0037 o de `evidencias/adr-0037/`, y están en Assumptions con su fuente.
- En esta sesión no se ha ejecutado ningún test ni `make ci`. Lo comprobado está en la viñeta «Lo comprobado en esta sesión» de Assumptions.
