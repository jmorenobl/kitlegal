# Specification Quality Checklist: H5.1 · Avisos de vigencia en las evals

**Purpose**: Validar la completitud y calidad del spec antes de pasar a planificación
**Created**: 2026-09-16
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Success criteria are technology-agnostic (no implementation details)
- [x] All acceptance scenarios are defined
- [x] Edge cases are identified
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [x] Feature meets measurable outcomes defined in Success Criteria
- [x] No implementation details leak into specification

## Notas de validación

- **«No implementation details» y «No implementation details leak»**: el spec nombra solo las rutas, el campo `avisos`, `Juzgar`, `CodigosDeAviso()` y las claves `avisos_encontrados` y `avisos_ausentes` que nombra el propio hito (*Criterios del hito, literales*, «Alcance»). No fija cómo se normaliza la respuesta, cómo se exportan las etiquetas, dónde vive la comprobación mecánica ni el texto de los motivos: la tolerancia al detalle la fija el plan (FR-032) y la redacción de la pregunta también (FR-051).
- **«Focused on user value»**: US1 (quien consulta una norma derogada ve el aviso con forma fija) y US2 a US5 (quien mantiene las evals mide esa regla sin modelo); el *Resumen* enlaza con la bitácora que origina el hito.
- **«All mandatory sections completed»**: *User Scenarios & Testing* (US1 a US5 y *Edge Cases*), *Requirements* (FR-001 a FR-073, *Key Entities*), *Success Criteria* (SC-001 a SC-009), *Fuera de alcance* y *Assumptions*.
- **«No [NEEDS CLARIFICATION] markers remain»: resuelto.** `/speckit-clarify` (sesión 2026-09-16) fijó el marcador de FR-071: la tasa de la aceptación es la que el informe ya publica por eval (sesiones que pasan sobre las tres repeticiones, ADR 0016), junto con el reparto de avisos por sesión de FR-040 a FR-042; no se añade ningún recuento nuevo (*Fuera de alcance*). De paso precisó FR-014: la comprobación mecánica reconoce en `SKILL.md` la forma fija completa (`⚠`, la etiqueta y dos puntos) de cada aviso con la misma función tolerante que `Juzgar`. Esa lectura llega a todo lo que describe la comprobación o lo que `SKILL.md` tiene que llevar: FR-003, FR-014, US1 (Independent Test y escenario 2), US4 (narrativa, Independent Test y escenario 3), SC-001, SC-005 y la fila del control en la tabla de trazabilidad; ninguno describe ya la etiqueta suelta, y `SKILL.md` no tiene ningún caso de «etiqueta que sobra», que FR-014 no define.
- **«Requirements are testable»**: cada FR tiene su escenario o su criterio medible: FR-001 a FR-003 → US1 escenarios 1 y 2 y SC-001; FR-004 → US1 escenario 4 y SC-001 (diff contra `main`); FR-005 → US1 escenario 3 y SC-001; FR-010 → US4 escenario 1 y SC-005; FR-011 → US4 escenario 5 y SC-005 (sin literales de las etiquetas en `Juzgar` ni en la comprobación); FR-012 → US4 escenario 4 y SC-005; FR-013 → US4 escenario 2 y SC-005; FR-014 → US4 escenario 3 y SC-005; FR-020 → US2 escenario 10 y SC-002; FR-021 → US2 escenario 1 y SC-002; FR-022 → US2 escenarios 2 y 9 y SC-002; FR-023 → SC-002; FR-030 → US2 escenarios 3 y 7; FR-031 → US2 escenarios 5 a 7 y SC-003; FR-032 y FR-033 → US2 escenarios 3 a 6 y SC-003; FR-034 → US2 escenario 8 y SC-003; FR-035 → SC-003; FR-040 → US3 escenarios 1 y 3 y SC-004; FR-041 → US3 escenario 2 y SC-004; FR-042 → US3 escenario 1 y SC-004; FR-043 → SC-004; FR-050, FR-051 y FR-053 → US5 escenarios 1 y 2 y SC-006; FR-052 → US5 escenario 5 y SC-009; FR-054 y FR-057 → US5 escenario 3 y SC-006; FR-055 → SC-009; FR-056 → US5 escenario 4 y SC-006; FR-060 → SC-006; FR-061 → SC-008; FR-070 a FR-073 → SC-007.
- **«Success criteria are measurable» y «technology-agnostic»**: SC-001 a SC-009 se cuentan o se comprueban sobre el repositorio (forma fija completa de 3 avisos, 300 líneas, 100 % del diff de `SKILL.md` en sus tres partes, 17 evals válidas, 5 casos de juicio, 5 casos de la comprobación —código de menos y de más en el esquema; en `SKILL.md`, forma sin etiqueta, sin `⚠` o sin dos puntos—, 3 etiquetas exportadas y ninguna escrita a mano en `Juzgar` ni en la comprobación, 18 evals con 10 que deciden, 2 avisos sobre 3 sesiones, informe con su commit registrado en el directorio del hito, `red` vacío, las tres entradas de documentación, el orden de commits y las tareas `[datos]`) sin fijar lenguaje, biblioteca ni diseño.
- **«All acceptance scenarios are defined»**: los cinco casos de los *Controles* del hito están en US2 escenarios 3 a 6 (y 7), los de formato en US2 escenarios 1, 2, 9 y 10, los del informe en US3, los de la comprobación mecánica en US4 escenarios 2 y 3, la fuente única de la etiqueta en US4 escenario 5, el diff acotado de `SKILL.md` en US1 escenario 4, el orden de commits en US5 escenario 5, y la aceptación en FR-070 a FR-073 y SC-007.
- **«Edge cases are identified»**: forma fija con afirmación contraria (limitación declarada), aviso no esperado, etiqueta alterada, etiqueta sin `⚠` o sin dos puntos, varias apariciones o varias normas, `avisos` vacío o repetido, eval informativa con aviso ausente, sesión sin sobre con avisos e invocación fuera de lo grabado.
- **«Scope is clearly bounded»**: *Fuera de alcance* transcribe el del hito y añade lo no especificado (contenido tras los dos puntos, avisos no esperados, atribución a normas, otras evals de avisos, parámetros del job, grabaciones ajenas, kernel, ADR, `docs/SOURCES.md`, `docs/USO.md`, e2e del applet, un recuento adicional de sesiones que trasladaron cada aviso).
- **«Dependencies and assumptions identified»**: *Assumptions* recoge la tasa de ADR 0016, el bloque `a42`, los metadatos grabados en H4, que las etiquetas son el prefijo de las frases actuales, el disparador del job y las dependencias de H4 y H5.
- **«All functional requirements have clear acceptance criteria»**: todo FR, de FR-001 a FR-073, tiene al menos un criterio de éxito medible, según el mapeo de la nota «Requirements are testable»; ninguno queda solo con su propio enunciado. Los que no son de comportamiento del producto también: FR-004 (SC-001, diff de `SKILL.md`), FR-011 (SC-005, sin literales de las etiquetas), FR-052 y FR-055 (SC-009, orden de commits y tareas `[datos]`), FR-061 (SC-008, `CHANGELOG.md`, `README.md` y `CONTRIBUTING.md`) y FR-070 (SC-007, commit de la ejecución e informe registrado en el directorio del hito).
- **«User scenarios cover primary flows»** y **«Feature meets measurable outcomes»**: ver la tabla de trazabilidad de *Criterios del hito, literales*, que lleva cada criterio literal del hito a sus FR, US y SC.
