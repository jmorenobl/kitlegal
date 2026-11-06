# Specification Quality Checklist: H41 · Las skills del producto mandan leer sus referencias

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-11-05
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué debe tener cada remisión (qué es una
  frase, FR-001, como comportamiento), qué cuenta la tabla y qué comprueba `make ci`, no cómo se declaran las variantes,
  cómo se instala `<skill>-producto` ni cómo se organiza el código de la regla de lectura: FR-041 remite a la partición
  de `specs/031-h39-c-omo-se/contracts/tabla.md` §2, sin elegir función ni lector. Los nombres `make evals-laboratorio`
  y `make laboratorio-tabla` y las rutas de `laboratorio/` los da el hito.
- [x] Focused on user value and business needs — que en modo herramienta `normativa` abra las referencias que necesita
  como en modo orden, sin que ninguna skill abra menos que hoy (Contexto; FR-052).
- [x] Written for non-technical stakeholders — cada historia explica su porqué; «remisión», «familia», `control` y
  `producto` se definen en «Key Entities».
- [x] All mandatory sections completed — User Scenarios, Requirements, Success Criteria, Assumptions y Fuera de alcance.

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — la de FR-041 se cerró en clarify: la regla ampliada rige solo en las
  familias de las skills del producto y la tabla de H39 se queda como está (conservadora, a «Fuera de alcance»).
- [x] Requirements are testable and unambiguous — cada FR nombra su comprobación: FR-001 a FR-003, FR-007 a FR-010,
  FR-020 a FR-023, FR-030, FR-040 a FR-042 en `make ci`; FR-004 a FR-006, con la lista de remisiones del plan que
  comprueba el test de FR-001; FR-051, el job de evals del cierre; FR-052, la persona. Sin lecturas abiertas en el
  arnés y la tabla: `SESIONES` es un múltiplo de 4 mayor que 0 (FR-030); la aceptación de `normativa` son dos
  peticiones, 160 y 40 sesiones (US3-1, conforme a FR-030); una tabla cuenta una sola familia y un directorio de otra es
  un defecto con código 1 (FR-040), así que la aceptación da una tabla por familia (US5-1, SC-010).
- [x] Success criteria are measurable — SC-001 a SC-009 con recuento, porcentaje o bytes y su control; SC-010, capa 3.
- [x] Success criteria are technology-agnostic (no implementation details) — miden frases, caracteres, bytes, ficheros,
  sesiones abiertas y cifras de la tabla, no cómo se calculan.
- [x] All acceptance scenarios are defined — escenarios Dado/Cuando/Entonces en las cinco historias.
- [x] Edge cases are identified — cuerpo que pasa de 11 000, cambio en una `<skill>-control`, `skills/` que cambia tras
  el lanzamiento, referencia nueva, referencia a cuya condición no llega ninguna eval, sesión sin terminar, transcripción
  ilegible (regla genérica), `DIRECTORIO` con sondeos de más de una familia (FR-040), y en los errores de argumentos de
  US3-4, `SESIONES` que no es un múltiplo de 4 mayor que 0.
- [x] Scope is clearly bounded — «Fuera de alcance» con los topes, las descripciones, el generador, `consulta-juridica`,
  las instrucciones y descripciones de herramientas de `kitlegal mcp serve`, cambiar las evals que ya existen, evals
  nuevas, el juez, el artefacto del job, otras variantes, modos, modelos y hosts, la huella o copia de
  `<skill>-producto`, la regla con `cd` en la tabla de H39 y el sondeo dentro del run.
- [x] Dependencies and assumptions identified — dependencia de H40 (variantes, arnés, tabla; FR-021, FR-030, FR-040) y
  supuestos sobre las cifras de partida, `v0.9.0`, la regla ampliada y las lecturas grabadas (Assumptions).

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — escenarios de las historias 1 a 5 y SC-001 a
  SC-010, cada uno con los FR que cubre.
- [x] User scenarios cover primary flows — remisiones, sitio en el cuerpo, arnés, tabla con la regla ampliada y la
  aceptación de la persona.
- [x] Feature meets measurable outcomes defined in Success Criteria — cada SC remite a su FR y a su control.
- [x] No implementation details leak into specification — ver el primer ítem: FR-041 ya no fija la estructura del
  código de la regla de lectura.

## Criterios del hito y del juez

- [x] Los criterios de «Aceptación» y «Controles» del hito están literalmente en el spec — sección «Criterios del hito
  (literales)», con los FR y SC de cada uno.
- [x] Cada umbral que se mide sin persona nombra la comprobación que sale en rojo — 11 000 caracteres (FR-008, SC-002),
  un test de `make ci`; bytes y ficheros de las variantes (SC-006, SC-008), recuentos de la tabla (SC-005) y sesiones del
  arnés (SC-007), tests de `make ci`; evals del cierre (SC-009), el job de evals; los intervalos de la aceptación
  (SC-010) solo los mide la persona y el spec lo dice.
- [x] Uso de fuera adentro — sección «Uso, de fuera adentro»: quién lee el cuerpo, las referencias y la tabla, cuántas
  veces por pregunta, su tamaño (cuerpo de 11 000 como mucho, unos 5,8 KB de referencias por pregunta en `normativa`,
  una tabla por familia, trece filas entre las tres y 520 sesiones) y que no crecen con lo consultado.
- [x] Ninguna decisión cerrada de CLAUDE.md se reabre — el tope de 11 000 del ADR 0046 sigue (FR-008, Fuera de alcance);
  las descripciones no cambian (FR-010); las variantes no se distribuyen (FR-023); el juez no mide el sondeo y la
  persona lo lanza con su credencial (FR-031; ADR 0032).

## Notes

- Clarify (2026-11-05) cerró tres preguntas: alcance de la regla de FR-041 (solo familias del producto), origen de las
  referencias de `<skill>-producto` (lista de nombres que deja el arnés, FR-033) y qué es nombrar `<f>` por su nombre
  solo (`<f>` o `./<f>` como palabra de una orden simple).
