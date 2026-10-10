# Specification Quality Checklist: H25 · El juez de `jurisprudencia`: resumir o caracterizar una sentencia que no se ha leído decide

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-10
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué recibe el juez de una sesión de `jurisprudencia`, qué umbrales deciden, cómo se resuelve un caso y qué espera cada eval. Los nombres de ficheros, tests, umbrales y etiquetas que aparecen son los que fija el hito o existen en el repositorio; los nombres de los ficheros de las evals, cómo conoce la reconstrucción el applet `cita`, las claves del voto en el informe y el texto de las dos frases de `SKILL.md` quedan para el plan (Assumptions, última viñeta).
- [x] Focused on user value and business needs — Resumen y US1: lo que lee quien pregunta por una sentencia; US2 a US6: el instrumento que lo mide; US7: las dos frases de la skill.
- [x] Written for non-technical stakeholders — cada historia empieza por lo que ve una persona; el detalle técnico está en los requisitos, como en H23 y H24.
- [x] All mandatory sections completed — User Scenarios, Requirements, Success Criteria, Assumptions y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — la de FR-093 (si entra corregir `CONTRIBUTING.md`) la cerró el paso de clarify el 2026-10-10: entra, con SC-015.
- [x] Requirements are testable and unambiguous — cada FR lleva «Comprobable», un **Control** o remite a un control de FR-100 a FR-112; FR-081 y FR-082, que no tienen control automático, dicen quién los mide (SC-014) y por qué; FR-093 lo comprueba la revisión final por lectura contra el árbol (SC-015), y FR-027 los jueces de la revisión final.
- [x] Success criteria are measurable — SC-001 a SC-013 con recuentos, segundos o número de casos; SC-014 es la parte humana de la Aceptación.
- [x] Success criteria are technology-agnostic (no implementation details) — miden respuestas, votos, casos y umbrales; los controles nombran la comprobación que falla, como exige el ADR 0029.
- [x] All acceptance scenarios are defined — US1 a US7, en forma Dado/Cuando/Entonces (31 escenarios).
- [x] Edge cases are identified — «Edge Cases»: el equivalente deducido, el fallo solo y la ficha sola, la sesión sin textos, la orden que falla, sí-sí-no, el voto que no llega, un modo marcado y otro no, el derivado sin documento del modo herramienta, el límite de ritmo con nueve sesiones.
- [x] Scope is clearly bounded — «Fuera de alcance», con las nueve viñetas literales del hito y lo no especificado.
- [x] Dependencies and assumptions identified — «Assumptions» y «Relación con H23 y H24».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — tabla de trazabilidad de «Criterios del hito, literales» y controles FR-100 a FR-112; FR-093 ya no es una marca abierta (SC-015).
- [x] User scenarios cover primary flows — la respuesta (US1), el juez sobre las sesiones y sus umbrales (US2), la carpeta y la medida versionada (US3), la reconstrucción de los casos y la ejecución de la medida (US4), las cuatro evals (US5), el trabajo del job (US6) y la skill v0.1 (US7).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 y SC-014 recogen la Aceptación literal; SC-002 a SC-013, los Controles.
- [x] No implementation details leak into specification — ver el primer ítem.

## Criterios del juez del spec (workflow `hito`, `juez_spec`)

- [x] a. Alcance: cubre Entrega 1 a 7 y nada más (tabla de trazabilidad). No especifica caso a caso ficheros manipulados a mano: una copia que difiere hace fallar `make ci` (FR-001), una declaración de clases mal formada es un fichero mal formado (FR-002) y un caso que no se resuelve es un error de la reconstrucción (FR-044).
- [x] b. Fuera de alcance: las nueve viñetas literales del hito y lo no especificado (otra clase, un control automático del CAPTCHA, más evals, el formato de eval, un selector por skill, el sondeo, la orden y el mensaje del voto).
- [x] c. Clarificaciones: dos, en «Clarifications» (2026-10-10): la reparación del cierre puede cambiar `SKILL.md` fuera de los dos pasajes con traza (FR-027, FR-083) y `CONTRIBUTING.md` entra en la documentación (FR-093).
- [x] d. Aceptación: «Criterios del hito, literales» lleva Entrega, Controles y Aceptación transcritos; comprobado sin modelo en la sesión que escribió el spec que los 30 fragmentos —7 de la Entrega, 12 de los Controles, la Aceptación y 10 de «Fuera de alcance»— están en `spec.md` tal como están en `docs/ROADMAP.md`. SC-001 a SC-014.
- [x] e. Decisiones cerradas: el juez es el del ADR 0037 (modelo distinto del que decide, tres votos unánimes, medido, nunca en un paso del run: FR-011, FR-031, FR-033, FR-095); no se consulta el CENDOJ ni cambia el applet `cita` (FR-096); ni el modelo que decide, ni las repeticiones, ni la regla por serie.
- [x] f. Checklist veraz: cada marca cita dónde se cumple; ningún ítem queda sin marcar.
- [x] g. Sin implementación: las formas concretas son del plan; `TestCopiasDelJuez`, `TestDefinicionDelJob`, `mensajeDelVoto`, `scripts/evals-voto.sh` y `guiones/casos.py` los nombra el hito.
- [x] h. Derivable: lo que el juez recibe (FR-010), los umbrales con su medida y su total (FR-020 a FR-024), la reconstrucción de cada origen de caso y de cada derivado (FR-041 a FR-044), lo que espera sin modelo cada eval (FR-061 a FR-064) y los peores casos del job con sus cifras (FR-072).
- [x] i. Uso: «Uso, de fuera adentro», con quién pide cada salida, su tamaño con su cálculo y cuándo deja de darse cada señal.
- [x] j. Umbral con control: FR-020, FR-022 y FR-023 nombran el job en rojo, y FR-022 además `make ci`; FR-001, FR-045, FR-065, FR-072 y FR-083, su comprobación de `make ci`; SC-001 a SC-013 llevan **Control**; FR-101 fija las seis filas de «Controles de umbral»; FR-111 pide la prueba de que los controles del informe fallan por encima de su umbral. `afirma_que_existe` no es un requisito con umbral: el hito lo deja publicado y que decida está fuera de alcance (FR-021).

## Notes

- No queda ninguna marca [NEEDS CLARIFICATION]: la de FR-093 la cerró el paso de clarify. Las demás lecturas que el spec fija salen del hito, del ADR 0037, de H23 y H24 o de `evidencias/adr-0037-jurisprudencia/`, y están en Assumptions con su fuente.
- En esta sesión no se ha ejecutado ningún test ni `make ci`. Lo comprobado está en la viñeta «Lo comprobado en esta sesión» de Assumptions: los recuentos de los casos, la medida, la definición del job, una orden `kitlegal cita preparar` con el binario compilado de la rama y los peores casos del job recalculados a mano.
