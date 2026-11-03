# Specification Quality Checklist: H39 · Cómo se organiza una skill para que se active y se lea en los dos modos

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-11-03
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué llega a cada esquema (FR-001 a FR-004), qué es una variante (FR-020 a FR-029) y qué cuenta el guion (FR-040 a FR-043), sin fijar paquete, formato de fichero ni mecanismo; cómo existen las instrucciones y los parámetros de las sesiones del sondeo sin tocar el binario distribuido (FR-029) es del plan.
- [x] Focused on user value and business needs — historias 1 a 4: lo que gana el agente en modo herramienta (FR-001, FR-002) y la medida que decide el ADR (FR-050, FR-051).
- [x] Written for non-technical stakeholders — cada historia explica su porqué; los nombres de herramientas son los del producto.
- [x] All mandatory sections completed — User Scenarios, Requirements, Success Criteria, Assumptions y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — ninguno; las lecturas tomadas van en Assumptions.
- [x] Requirements are testable and unambiguous — cada FR nombra su comprobación (FR-004, FR-010 a FR-014, FR-028, FR-043) o su escenario Dado/Cuando/Entonces; «antes de la primera herramienta» se define por turnos en FR-041, y el turno es una respuesta del modelo (los bloques que emite antes de recibir el resultado de ninguna herramienta, en uno o varios mensajes `assistant`), numerada desde 1; «las que activan las dos» también es antes de la primera herramienta; FR-030 fija qué skills tiene instaladas cada sesión.
- [x] Success criteria are measurable — SC-001 a SC-008 con recuento o porcentaje y su control en `make ci` o en el job de evals; SC-009 lo comprueba la persona.
- [x] Success criteria are technology-agnostic (no implementation details) — miden lo que ve la persona o el agente (descripciones, ficheros instalados, recuentos, sesiones abiertas).
- [x] All acceptance scenarios are defined — escenarios en las cuatro historias, con los valores de Wilson de la historia 3.
- [x] Edge cases are identified — sesión sin herramienta, sin terminar, referencia abierta dos veces, variante sin referencia, variante con referencia que ninguna sesión abre, fila con n = 0 (FR-042, FR-043), transcripción ilegible (regla genérica, ADR 0023).
- [x] Scope is clearly bounded — «Fuera de alcance» con la reorganización, los topes, las instrucciones del producto, las banderas globales, el artefacto del job y el sondeo dentro del run.
- [x] Dependencies and assumptions identified — dependencia de H32 (evidencia de partida, `consulta-juridica`, grabaciones de las evals) y Assumptions.

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — FR-001 a FR-005 (historia 1, SC-001, SC-002), FR-010 a FR-029 (historia 2, SC-003 a SC-005), FR-030 a FR-043 (historia 3, SC-006, SC-007), FR-050 a FR-053 (historia 4, SC-008, SC-009).
- [x] User scenarios cover primary flows — esquemas, laboratorio, arnés y guion, aceptación de la persona.
- [x] Feature meets measurable outcomes defined in Success Criteria — cada SC se traza a sus FR.
- [x] No implementation details leak into specification — ver el primer ítem.

## Criterios del hito y del juez

- [x] Los criterios de «Aceptación» y «Controles» del hito están literalmente en el spec — sección «Criterios del hito (literales)», trazados a FR-010, FR-014, FR-028, FR-001, FR-002, FR-050 a FR-053.
- [x] Cada umbral que se mide sin persona nombra la comprobación que sale en rojo — SC-001 a SC-007 (tests de `make ci`; SC-007 cubre también el reparto n/4 por eval y las 0 sesiones de una petición inválida de FR-030), FR-011 (forma y topes de la skill de laboratorio en `make ci`), SC-008 (job de evals del cierre); SC-009 y FR-050 los mide la persona.
- [x] Uso de fuera adentro — sección «Uso, de fuera adentro»: quién pide cada salida, cuántas veces, su tamaño (2373 caracteres como cota de las descripciones; 14 filas y 560 sesiones) y cuándo deja de leerse.
- [x] Ninguna decisión cerrada de CLAUDE.md se reabre — topes del ADR 0046, sesiones sin modelo en el run (ADR 0032), evals del job sin cambios, sin telemetría.

## Notes

- Validado en una iteración. `scripts/workflow/precheck.sh spec` no se pudo ejecutar en esta sesión (pidió aprobación);
  sus comprobaciones se repitieron con grep: sin marcas pendientes, «## Fuera de alcance» presente, 41 ids definidos una
  vez, ninguno citado sin definir, escenarios Dado/Cuando/Entonces y ningún ítem sin marcar.
