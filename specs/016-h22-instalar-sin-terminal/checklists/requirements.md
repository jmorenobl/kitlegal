# Specification Quality Checklist: H22 · Instalar sin terminal: la extensión de escritorio con el servidor y el plugin de Claude con las skills

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-02
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué llevan los dos ficheros, qué publica la release y qué comprueba cada control. Los nombres técnicos que aparecen (Go y sin Node, goreleaser, `release.extra_files`, `checksum.extra_files`, `universal_binaries`, los campos del manifiesto `0.3`, `claude plugin validate`, `humo`) son los que fija la sección del hito. Dónde vive el paso, los textos y la plantilla, cómo obtiene el paso las herramientas y las skills, cómo se hace reproducible el zip y qué cliente MCP usan los controles quedan para el plan (Assumptions, última viñeta).
- [x] Focused on user value and business needs — Resumen y US1: quien no usa una terminal instala con dos ficheros y pregunta; US2, quien añade el marketplace; US4, quien lee el README ante el aviso rojo.
- [x] Written for non-technical stakeholders — cada historia empieza por lo que hace y ve una persona; el detalle técnico está en los requisitos, como en H19 y H21.
- [x] All mandatory sections completed — User Scenarios & Testing, Requirements, Success Criteria, Assumptions y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — cerrados por clarify (sesión 2026-10-02): FR-006 (sin archivo nuevo: el universal se compara por arquitecturas con los dos archivos de macOS) y FR-017 (el esquema oficial no se versiona en el run: control propio del repositorio y pendiente de la persona en «Fuera de alcance»).
- [x] Requirements are testable and unambiguous — cada requisito lleva «Comprobable» o lo cubre un control de FR-060 a FR-070. FR-001 a FR-005, FR-060, FR-067 y FR-068; FR-010 a FR-017, FR-061, FR-062, FR-064 y FR-066; FR-020 a FR-023, FR-063 y FR-065; FR-030 y FR-031, FR-065 y FR-068; FR-040 a FR-043, la comprobación de la definición de la release (FR-068) y, en cada release, el propio `humo`; FR-050 a FR-053, la revisión final (SC-012). FR-032 y FR-080 se comprueban en la revisión, que ve si el diff toca lo que no debe. Lo que dependa de los dos marcadores se cierra con ellos.
- [x] Success criteria are measurable — SC-001 a SC-011 con recuentos, caracteres, píxeles o bytes; SC-002 lo mide una persona tras publicar una release, como dice el hito; SC-012 es documentación, sin umbral.
- [x] Success criteria are technology-agnostic (no implementation details) — miden ficheros publicados, huellas, entradas, herramientas listadas y caracteres; cada uno nombra la comprobación que falla, como exige el ADR 0029.
- [x] All acceptance scenarios are defined — US1 a US5, en forma Dado/Cuando/Entonces.
- [x] Edge cases are identified — «Edge Cases»: una sola pieza, el plugin junto a `skills install`, la versión en un snapshot, Windows `arm64` y Linux, la descripción de más de 120 caracteres, el directorio desde el que arranca la app y lo manipulado.
- [x] Scope is clearly bounded — «Fuera de alcance»: las nueve exclusiones literales del hito y lo que el hito no especifica.
- [x] Dependencies and assumptions identified — «Assumptions» y «Relación con H19 y H21».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — tabla de trazabilidad de «Criterios del hito, literales» y controles FR-060 a FR-070.
- [x] User scenarios cover primary flows — instalar y preguntar (US1), el marketplace (US2), publicar y probar en local (US3), el README (US4) y un verbo o una skill nuevos (US5).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 y SC-002 recogen la Aceptación literal, en el run y humana; SC-003 a SC-011, los Controles.
- [x] No implementation details leak into specification — ver el primer ítem.

## Notes

- Items marked incomplete require spec updates before `/speckit-clarify` or `/speckit-plan`.
- Los dos marcadores `[NEEDS CLARIFICATION]` los cerró clarify el 2026-10-02 (ADR 0018), con una tercera decisión, FR-031 (el catálogo se publica después de `humo`); véase `## Clarifications` del spec.
- Medidas citadas en «Uso, de fuera adentro»: tomadas en la sesión que escribió el spec, sobre `main` (`751b76e`), o leídas de ficheros versionados; lo que no se midió, el spec lo dice.
