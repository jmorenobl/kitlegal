# Specification Quality Checklist: H26 · El umbral de las clases del juez pasa a ser una tasa sobre los dos modos

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-13
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué lleva el informe del job, qué decide y qué se publica, con qué valores y qué dicen los documentos. Los nombres que aparecen (`clases.yaml`, `umbrales`, `informe.json`, `TestCopiasDelJuez`, `TestDefinicionDelJob`, `TestEvalsDelRepositorio`, `make evals-sondeo`) los fija el hito o existen en el repositorio; los nombres de los tests nuevos, el texto de la `descripcion`, cómo escribe el motivo el modo de cada respuesta y cómo se aplican los mutantes quedan para el plan (Assumptions, última viñeta).
- [x] Focused on user value and business needs — Resumen y US1: el cierre de un hito cae cuando la skill empeora y no por una respuesta suelta; US2 a US5, lo que sigue viéndose y lo que no se afloja; US6, los documentos.
- [x] Written for non-technical stakeholders — cada historia empieza por lo que lee quien lanza un hito o mira un informe; el detalle está en los requisitos, como en H24 y H25.
- [x] All mandatory sections completed — User Scenarios, Requirements, Success Criteria, Assumptions y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — no se ha dejado ninguna: el hito fija los nombres, las medidas, los totales, los valores y los casos de cada test. Las lecturas que el spec toma por su cuenta están en Assumptions con su fuente.
- [x] Requirements are testable and unambiguous — cada FR lleva un **Control**, «Comprobable» o remite a un control de FR-070 a FR-079; FR-003 da los seis casos numéricos; FR-050 a FR-055, que no tienen test, los comprueba la revisión final por lectura contra el árbol (SC-013).
- [x] Success criteria are measurable — SC-001 a SC-011, SC-013 y SC-014 con recuentos, proporciones o número de elementos; SC-012 es la parte humana de la Aceptación.
- [x] Success criteria are technology-agnostic (no implementation details) — miden respuestas marcadas, elementos de `umbrales` y veredictos; los controles nombran la comprobación que falla, como exige el ADR 0029.
- [x] All acceptance scenarios are defined — US1 a US6, en forma Dado/Cuando/Entonces (15 escenarios).
- [x] Edge cases are identified — «Edge Cases»: las marcas en un solo modo, 2 y 2, la marca única de `jurisprudencia`, menos respuestas que las de hoy, el porcentaje escrito, la respuesta sin juzgar, sí-sí-no, la eval sin binario ni servidor, el instrumento sin medir, la skill sin juez, el total 0, el plan de un modo y el cierre en rojo por la tasa.
- [x] Scope is clearly bounded — «Fuera de alcance», con las seis viñetas literales del hito y lo no especificado.
- [x] Dependencies and assumptions identified — «Assumptions» y «Relación con H24 y H25».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — tabla de trazabilidad de «Criterios del hito, literales» y controles FR-070 a FR-079.
- [x] User scenarios cover primary flows — el umbral conjunto (US1), el recuento por modo (US2), los valores (US3), lo que sigue igual (US4), el sondeo (US5) y los documentos (US6).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001, SC-011 y SC-012 recogen la Aceptación literal; SC-002 a SC-010, SC-013 y SC-014, los Controles.
- [x] No implementation details leak into specification — ver el primer ítem.

## Criterios del juez del spec (workflow `hito`, `juez_spec`)

- [x] a. Alcance: cubre Entrega 1 a 4 y nada más (tabla de trazabilidad). No especifica caso a caso ficheros manipulados a mano: un valor distinto en `clases.yaml` hace fallar `make ci` (FR-022) y no se describe ninguna declaración mal formada.
- [x] b. Fuera de alcance: las seis viñetas literales del hito y lo no especificado (un umbral conjunto para lo que no es una clase, otro umbral para el recuento por modo, líneas nuevas en el sondeo, el contrato de `umbrales`, la tasa acumulada en el informe, la viñeta de H26 de `CLAUDE.md`).
- [x] c. Clarificaciones: el spec no deja ninguna pregunta; si el paso de clarify formula alguna, su respuesta entra en «Clarifications» con ese paso.
- [x] d. Aceptación: «Criterios del hito, literales» lleva Entrega, Controles y Aceptación transcritos; comprobado sin modelo en la sesión que escribió el spec que los 21 fragmentos —4 de la Entrega, 10 de los Controles, la Aceptación y 6 de «Fuera de alcance»— están en `spec.md` tal como están en `docs/ROADMAP.md`. SC-001 a SC-014.
- [x] e. Decisiones cerradas: el umbral es el del ADR 0038 (cero donde hay forma, tasa donde hay significado; solo se estrecha, y lo hace una persona: FR-020, FR-061); el juez, su regla, su rúbrica y su medida no cambian (FR-035); ninguna sesión del run abre el juez ni un sondeo (FR-060); ninguna respuesta se juzga con una lista de palabras.
- [x] f. Checklist veraz: cada marca cita dónde se cumple; ningún ítem queda sin marcar.
- [x] g. Sin implementación: las formas concretas son del plan; lo que el spec fija del informe —nombre, medida, total, umbral, `decide` y posición— es el contrato de una salida que leen el job, el informe final y una persona.
- [x] h. Derivable: el umbral conjunto con su medida, su total y su comparación sin redondeos, con seis casos numéricos (FR-001 a FR-003); su motivo (FR-004); el recuento por modo y su `cumple` que solo informa (FR-010); los trece elementos con los nueve que deciden (FR-012); y los valores (FR-020).
- [x] i. Uso: «Uso, de fuera adentro», con quién pide cada salida, su tamaño con su cálculo —medido en los informes del último cierre versionado— y cuándo deja de darse cada señal. Ninguna skill pide ninguna de estas salidas.
- [x] j. Umbral con control: FR-001 nombra el job en rojo; FR-012 y FR-022, su test de `make ci`; FR-031 a FR-034, el job en rojo, y FR-033 además `make ci`; SC-001 a SC-011, SC-013 y SC-014 llevan **Control**; FR-071 fija las nueve filas por skill de «Controles de umbral»; FR-078 pide la prueba, con mutantes, de que el control del umbral conjunto falla por encima de su umbral. El recuento por modo no es un requisito con umbral: el hito lo deja publicado (FR-010).

## Notes

- No queda ninguna marca [NEEDS CLARIFICATION].
- En esta sesión no se ha ejecutado ningún test ni `make ci`. Lo comprobado está en la viñeta «Lo comprobado en esta sesión» de Assumptions: los valores de hoy de las dos `clases.yaml`, el esquema, las carpetas de `evidencias/`, los `umbrales` y los tamaños de los informes del último cierre versionado, y las proporciones calculadas a mano.
