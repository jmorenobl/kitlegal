# Specification Quality Checklist: H32 · Arquitectura de skills: una entrada genérica, skills por capacidad y las materias como referencias

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-29
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué lleva cada skill, qué hace la
  instalación y qué control sale en rojo; los nombres de las skills, dónde vive un paquete, cómo llega cada referencia a
  su skill, la redacción de los arreglos y el peor caso de cada trabajo los decide el plan (FR-007, FR-015, FR-016,
  FR-022, FR-039).
- [x] Focused on user value and business needs — US1 parte de la pregunta de cualquier materia (`docs/USO.md`,
  2026-10-21); US2 y US3, de los defectos medidos que ve quien lee la respuesta; US5, de quien ya instaló las skills.
- [x] Written for non-technical stakeholders — cada historia dice por qué importa, con las cifras del hito; los nombres
  de controles son los de `docs/SKILLS.md` §1.
- [x] All mandatory sections completed — escenarios, requisitos, «Uso, de fuera adentro», criterios de éxito, supuestos y
  «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — cerrados en clarify (2026-10-29): lo que la base no cubre se dice en
  prosa, sin forma fija nueva (FR-001); «de un paquete» es lo que `data/normas.yaml` marca con la vertical de un paquete
  (FR-021); una instalación parcial retira la skill de nombre retirado si las huellas coinciden (FR-026).
- [x] Requirements are testable and unambiguous — cada FR lleva su «Criterio»; el Contexto fija «skill de entrada»,
  «skill de capacidad», «paquete de materia», «nombre retirado», «forma única de cita», «cita de un bloque» y las evals
  por su fichero en `c991808`; FR-006 y FR-032 definen qué es una remisión o una aparición de un nombre para que la
  búsqueda no confunda un nombre con una palabra. SC-010 y US6 escenario 1 dan el 0.028 a la skill de normativa y el
  0.017 a la heredera de `jurisprudencia`, como FR-038; US7 escenario 2 es la búsqueda de FR-032, con sus directorios y
  sus excepciones, sin contradecir «Fuera de alcance».
- [x] Success criteria are measurable — SC-002 (11 000 caracteres), SC-007 (900 caracteres y umbral 0), SC-010 (0.028),
  SC-011 (0 por modo), SC-012 (0 sesiones), SC-016 (la tabla del ADR 0047), y cada umbral que se mide sin persona nombra
  la comprobación que sale en rojo (SC-001 a SC-015, SC-017, SC-018; FR-014, FR-021, FR-039).
- [x] Success criteria are technology-agnostic (no implementation details) — miden la skill, la instalación y el cierre
  por sus resultados; las comprobaciones que nombran son las de hoy o las que pide el hito.
- [x] All acceptance scenarios are defined — US1 a US8, en Dado/Cuando/Entonces.
- [x] Edge cases are identified — dos capacidades en una pregunta, una materia sin paquete, un paquete que no cabe, un
  nombre retirado pedido por su nombre o en dos ámbitos, una skill no empotrada de otra rama, y la regla genérica para lo
  manipulado por algo que no es el binario.
- [x] Scope is clearly bounded — «Fuera de alcance» recoge el del hito y lo que no está en él (juez de la skill de entrada,
  otras materias, `doctor` y `list`, `CLAUDE.md` y `CONTRIBUTING.md`, capturas de la web, el workflow). La búsqueda de
  FR-032 y SC-018 cuenta solo rutas que operan con la skill (`skills/`, `evals/` del repositorio o `.agents/skills/` de
  un ámbito, sin otro segmento delante) y deja fuera los ficheros de `evals/<skill>/juez/`, lo grabado bajo
  `internal/evals/testdata/` y las rutas a `specs/`: no obliga a editar los casos de los jueces ni un fixture; ningún
  umbral se ensancha (SC-010).
- [x] Dependencies and assumptions identified — «Assumptions»: nombres del plan, la serie que empieza de cero, la falta de
  red para las guías, la grabación de la Ley 50/1980, el arnés del sondeo y los documentos que nombran las skills de hoy.

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — FR-001 a FR-043, cada uno con su «Criterio».
- [x] User scenarios cover primary flows — la entrada (US1), las skills de capacidad y la cita (US2), los arreglos de
  normativa (US3), el molde (US4), las instalaciones (US5), las evals y el cierre (US6), la guía (US7) y la decisión de la
  persona (US8).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 a SC-018 se trazan a los FR y a los criterios
  del hito, copiados en «Criterios del hito, tal como los escribe».
- [x] No implementation details leak into specification — ver «Content Quality».

## Notes

- Comprobado contra la rúbrica de `juez_spec` (criterios a a j) y contra `scripts/workflow/precheck.sh spec`: el único
  defecto mecánico que queda son los tres marcadores y su ítem, que cierra la integración de clarify.
- Criterio j: SC-016 lo mide una persona (fuera del criterio); SC-014 no tiene umbral; los demás umbrales nombran su
  comprobación en rojo.
- Criterio i: «Uso, de fuera adentro» da, por salida, quién la pide, cuántas veces por pregunta, su tamaño medido o
  acotado y cuándo deja de darse cada señal.
