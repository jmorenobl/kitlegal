# Specification Quality Checklist: H0 · Esqueleto del repo y gates de CI

**Purpose**: Validar la completitud y calidad del spec antes de pasar a planificación
**Created**: 2026-09-09
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

- **«No implementation details» / «No implementation details leak» / «technology-agnostic»** (revalidados tras la corrección del gate del spec). El reparto real, requisito a requisito, es este:
  - **En los requisitos funcionales solo hay resultados exigibles y las herramientas que el propio hito hereda.** Nombrar `gofumpt`/`goimports`, `golangci-lint` y sus 22 linters, `govulncheck`, `gosec`, CodeQL, `gitleaks`, `go mod verify`, `go mod tidy -diff`, `lefthook`, Dependabot y Codecov no es una decisión de diseño del spec: son el objeto literal del hito, fijados por `docs/ROADMAP.md` §3 y por la constitución (*Restricciones técnicas*), y el criterio de aceptación del hito se enuncia sobre ellos.
  - **Ningún requisito prescribe ya la forma del código ni el mecanismo de aprovisionamiento.** FR-040 exige un resultado (el comportamiento de `version` cubierto por tests y el umbral global de cobertura superado sin exclusiones) y no la firma de una función; FR-042 exige cinco propiedades verificables (versión fijada, idéntica en local y en CI, integridad comprobable con el fichero de sumas, actualizable por Dependabot, sin más prerrequisitos que Go y `git`, sin contaminar las dependencias del producto) y no un mecanismo. FR-012 exige que exista un modo de verificación no mutante del formato, sin decir qué invocación lo produce.
  - **El *cómo* vive en dos sitios, ninguno de ellos normativo.** En `## Clarifications` —`run(args, stdout, stderr) int`, `main_test.go` en `package main`, directivas `tool` y `go tool`, el patrón `^fmt\.Print(|f|ln)$`— porque la constitución (*Criterio de decisión autónoma* §3) obliga a dejar rastro de cada decisión automática con la alternativa rechazada; y en el plan, que es quien lo materializa. Un cambio de mecanismo en el plan no obliga a tocar ningún FR.
  - Con ese reparto, las tres marcas se sostienen: lo que queda en la especificación es *qué* debe cumplirse y *con qué controles heredados*, no *cómo* se escribe.
- **«Written for non-technical stakeholders»**: las historias de usuario y los criterios de éxito describen el resultado (veredicto reproducible, PR bloqueada o aprobada, binario identificable, documentación fundacional) sin exigir conocer las herramientas; el detalle técnico está confinado a los requisitos funcionales, cuyo destinatario es quien planifica.
- **[NEEDS CLARIFICATION] resueltos (3, el máximo permitido)** — planteados en `specify` porque todos afectaban al alcance y ninguno tenía respuesta determinada por el hito, `CLAUDE.md`, `refs/` o la constitución; resueltos en la sesión de clarificación del 2026-09-09 (ver `## Clarifications` del spec):
  1. **FR-011 / FR-027** — comportamiento de las órdenes del `Makefile` y del flujo `nightly` cuyo objeto no existe hasta hitos posteriores (`test-integration`, `test-e2e`, `schema-check`, `skills-sync`, `release`). Impacta directamente en si `make ci` puede estar en verde en H0 (Definition of Done §1.1).
  2. **FR-014** — `docs/ROADMAP.md` §3 sitúa `depguard`/`forbidigo` en H1, pero la aceptación literal de H0 exige que un `fmt.Println` en `internal/` ya falle en lint. Es un conflicto entre dos secciones del roadmap, no una omisión.
  3. **FR-029** — aplicación de los umbrales de cobertura (`internal/core/**` ≥ 85 %, global ≥ 70 %) sobre un repositorio que en H0 casi no tiene código; determina si la propia PR de H0 puede pasar.
- **Ítem «No [NEEDS CLARIFICATION] markers remain»**: quedó sin marcar a propósito tras `specify`, porque en modo desatendido las tres cuestiones las decide el paso de clarificación con contexto limpio. Ese paso las resolvió (Q1→A, Q2→A, Q3→A, más dos cuestiones adicionales sobre `internal/` y aprovisionamiento de herramientas), integró las respuestas en FR-002, FR-004, FR-008, FR-010, FR-011, FR-014, FR-027, FR-029 y los nuevos FR-040 a FR-042, y no queda ningún marcador en el spec, por lo que el ítem pasa a marcado.
- **Trazabilidad de la aceptación literal del hito**: «PR de prueba con un `fmt.Println` en `internal/` falla en lint» → SC-005 y US2 escenario 1; «PR limpia pasa en < 3 min» → SC-006 y US2 escenario 2; «todos los controles de §3 marcados H0» → FR-012 a FR-024 y SC-007. La transcripción literal está en la sección *Criterios del hito, literales* del spec.
- **Ítem «Success criteria are measurable»** (revalidado tras la corrección del gate del spec). Los tres criterios que no lo eran ya lo son: **SC-002** añade la condición observable de que la orden agregada no modifica el árbol; **SC-004** enuncia la equivalencia en los dos sentidos y la hace comprobable porque FR-007 mete en `ci` los ocho controles de FR-012 a FR-020 que operan sobre un árbol de trabajo y FR-025 prohíbe a la integración continua ejecutar controles fuera de esas órdenes, con CodeQL y Dependabot excluidos de forma explícita por no operar sobre un árbol; **SC-008** sustituye «los doce ficheros» —cifra que no correspondía al alcance del hito— por la enumeración completa más los artefactos que exigen FR-005, FR-040 y FR-042. SC-006 conserva el umbral literal del hito y traslada al plan la medida en frío (*Assumptions*).
- **Ítem «Requirements are testable and unambiguous»** (revalidado): la contradicción entre US1 escenario 6 («un fichero sin formatear hace fallar el veredicto») y el par FR-007/FR-012 («`ci` ejecuta `fmt`, que aplica los formateadores») queda resuelta al separar el modo mutante del de verificación y dejar solo el segundo dentro de `ci`; los escenarios 6 y 7 de US1 comprueban ambas caras (falla señalando el fichero, y no modifica el árbol).
- **Correcciones del gate registradas**: `## Clarifications → Correcciones tras el gate del spec (2026-09-09)` recoge las seis decisiones tomadas al corregir, con su alternativa rechazada, conforme a la constitución (*Criterio de decisión autónoma* §3). Ninguna afectaba a alcance, frontera humana, privacidad, términos de uso ni a una decisión cerrada, por lo que no hay nada pendiente de decisión humana (`gates/spec-pendiente.md` no se crea).
- No quedan ítems incompletos: el spec está listo para `/speckit-plan`.
