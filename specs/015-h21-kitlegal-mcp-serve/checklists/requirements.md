# Specification Quality Checklist: H21 · `kitlegal mcp serve`: las herramientas del binario por MCP, con las skills y las evals en los dos modos

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-01
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué anuncia y qué devuelve el servidor, qué dicen las skills y qué mide el job. Los nombres técnicos que aparecen (`structuredContent`, `inputSchema`, `outputSchema`, `readOnlyHint`, `instructions`, el SDK, `depguard`) son los que fija la sección del hito. El paquete del servidor, el texto de las `instructions`, la forma de la tabla generada, la forma exacta de la línea `⚠ SIN CONSULTA AL BOE:`, los nombres de los umbrales por modo y el formato de la eval sin binario ni servidor quedan para el plan (Assumptions, última viñeta).
- [x] Focused on user value and business needs — Resumen y US1 a US3: lo que recibe quien pregunta desde un agente sin shell, con shell y sin nada; US4 y US7, quien declara el servidor; US5, el instrumento que lo mide.
- [x] Written for non-technical stakeholders — cada historia empieza por lo que hace y ve una persona; el detalle técnico está en los requisitos, como en H7.4.
- [x] All mandatory sections completed — User Scenarios & Testing, Requirements, Success Criteria, Assumptions y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — el de FR-050 se resolvió en clarify (sesión 2026-10-01, Q1, opción A): el sondeo sigue solo en el modo orden y pedirle una eval sin binario ni servidor es un error de uso; FR-084 lo comprueba.
- [x] Requirements are testable and unambiguous — cada requisito lleva «Comprobable» o lo cubre un control de FR-070 a FR-083. FR-001 a FR-008 los comprueban el control de conformidad (FR-070) y los e2e (FR-071, FR-073); FR-010 a FR-015, el e2e de FR-071 y el test de FR-074; FR-020 a FR-026, los e2e de FR-071, FR-072, FR-075 —que ejerce también `mcp serve --dry-run` (FR-022)— y FR-076, el control del kernel que recorre cada verbo con `--describe` y `depguard` (FR-079); FR-030 a FR-036, `make ci` (FR-077) y las evals en los dos modos (SC-001); FR-040 a FR-049, los tests del informe y de `Juzgar` (FR-080, que lee el informe de dos modos con `scripts/workflow/informe.sh`, y FR-081, con la llamada que falla y no satisface el `comando` de FR-042) y el job de cierre (SC-001); FR-060 y FR-061, la revisión final (SC-013). FR-050 lo comprueba FR-084 (tests del sondeo en `make ci`). FR-048 y FR-090 se comprueban en la revisión, que ve si el diff toca `scripts/workflow/` o los specs anteriores.
- [x] Success criteria are measurable — SC-001 a SC-013 con recuentos, proporciones, segundos o caracteres; SC-002 lo mide una persona tras fusionar, como dice el hito.
- [x] Success criteria are technology-agnostic (no implementation details) — miden herramientas anunciadas, sobres, respuestas, umbrales y líneas; cada uno nombra la comprobación que falla, como exige el ADR 0029.
- [x] All acceptance scenarios are defined — US1 a US7, en forma Dado/Cuando/Entonces.
- [x] Edge cases are identified — «Edge Cases»: herramienta no anunciada, bandera global como argumento, lectura y comprobación lanzadas a la vez, una llamada que falla o agota su plazo, la entrada que se cierra con llamadas en curso, dos procesos sobre la misma caché, herramienta y binario a la vez, la orden en el modo herramienta, el prefijo del agente y 3 de 54 en un solo modo.
- [x] Scope is clearly bounded — «Fuera de alcance»: las siete exclusiones literales del hito y lo que el hito no especifica.
- [x] Dependencies and assumptions identified — «Assumptions» y «Relación con H5, H6, H7.4 y H19».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — tabla de trazabilidad de «Criterios del hito, literales» y controles FR-070 a FR-083.
- [x] User scenarios cover primary flows — la respuesta por herramienta (US1), las dos formas en la skill (US2), sin herramienta ni binario (US3), el servidor en el agente (US4), el job en dos modos (US5), un applet nuevo (US6) y la documentación (US7).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 y SC-002 recogen la Aceptación literal, en el run y humana; SC-003 a SC-013, los Controles.
- [x] No implementation details leak into specification — ver el primer ítem.

## Notes

- Items marked incomplete require spec updates before `/speckit-clarify` or `/speckit-plan`.
- El ítem del marcador de FR-050, el único que quedaba sin marcar, lo cerró clarify con contexto limpio (ADR 0018): decisión conservadora de alcance, registrada en «Clarifications» del spec.
- Medidas citadas en «Uso, de fuera adentro»: tomadas en la sesión que escribió el spec, sobre `main` (`22b5bda`), o leídas de ficheros versionados; lo que no se midió, el spec lo dice.
