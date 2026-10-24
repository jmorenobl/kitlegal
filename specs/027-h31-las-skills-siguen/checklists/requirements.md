# Specification Quality Checklist: H31 · Las skills siguen la especificación Agent Skills

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-24
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec fija contratos observables (qué gate sale en
  rojo y con qué, qué deja `skills-sync`, qué decide el juez y qué publica el job); los nombres que da (`ValidarFrontmatter`,
  «en Go y sin red», `scripts/evals-voto.sh`, `mensaje_de`, `frase_esta`, `evidencias/adr-0046/`) son los que impone
  el propio hito (entrega 2 y 3) o el ADR 0046; cómo se implementa cada gate queda para el plan.
- [x] Focused on user value and business needs — «Contexto» dice qué gana cada skill (cuerpo de como mucho 11 000
  caracteres, descripción en tercera persona, materias como referencias) y «Uso, de fuera adentro» quién consume cada
  salida.
- [x] Written for non-technical stakeholders — en la medida en que los consumidores son el agente que carga la skill,
  quien la escribe o revisa y quien lee el informe del job; cada historia dice el porqué antes del criterio.
- [x] All mandatory sections completed — seis historias, requisitos FR-001 a FR-034, «Uso, de fuera adentro», criterios
  SC-001 a SC-015 con los del hito tal como los escribe, supuestos y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — los tres del paso `clarify` están resueltos en «Clarifications»: FR-014
  (el cierre por nombre lee también las referencias a mano), FR-030 (el juez solo corre si la elección incluye el cambio
  de una descripción) y FR-032 (la descripción sin juicio hace fallar el job); FR-009 precisa el caso sin encabezados.
- [x] Requirements are testable and unambiguous — cada FR lleva su «Criterio», un SC con el test de `make ci` o el job que
  sale en rojo, o la revisión final donde el hito la pide (FR-001, FR-013, FR-021, FR-024); FR-003, FR-007, FR-008,
  FR-009 y FR-012 definen etiqueta XML, enlace, índice y referencia generada o a mano (FR-012 clasifica primero por la
  declaración, como dicen «Edge Cases», US4 y SC-003); la suposición de la elección es la regla de FR-030; FR-028 y
  FR-032 dicen qué voto se repite (el nulo y el cortado por el tope) y cuál no; SC-005 lleva un mutante por cada
  control de FR-016, FR-018 y FR-019.
- [x] Success criteria are measurable — SC-002 a SC-009 y SC-014 y SC-015 nombran su cifra y el test de `make ci` que
  sale en rojo; SC-010, el job de evals con sus umbrales que deciden; SC-011, la medida que lanza una persona; SC-012 y
  SC-013, la revisión final.
- [x] Success criteria are technology-agnostic (no implementation details) — se miden en `make ci`, en el job de evals y
  en la revisión, que son la superficie del proyecto.
- [x] All acceptance scenarios are defined — seis historias en Dado/Cuando/Entonces.
- [x] Edge cases are identified — «Edge Cases»: skill sin `references/`, fichero sin cabecera con nombre declarado, enlaces con ancla y en bloques de código, referencia generada de más de 100 líneas, `<` que no es
  etiqueta, pregunta citada en la descripción y lo manipulado a mano (regla genérica).
- [x] Scope is clearly bounded — «Fuera de alcance» lleva lo que el hito deja fuera y lo que el spec no especifica,
  también lo que deja fuera la Clarificación Q4 (cambiar los generadores para que escriban su índice).
- [x] Dependencies and assumptions identified — «Assumptions»: fuentes sin red, la regla de la elección de FR-030, modelo y
  versión del juez, instalación de `references/`, serie de cierres y release tras H31.

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — «Criterio» en cada FR; los tres con
  [NEEDS CLARIFICATION] tienen ya su SC, que no cambia con la respuesta.
- [x] User scenarios cover primary flows — skills reescritas (US1), gates (US2), juez de las descripciones (US3),
  referencias a mano (US4), controles de contenido (US5) y guía (US6).
- [x] Feature meets measurable outcomes defined in Success Criteria — los «Controles» y la «Aceptación» del hito, tal
  como los escribe, están en SC-001 a SC-011.
- [x] No implementation details leak into specification — ver el primer punto.

## Notes

- Los tres [NEEDS CLARIFICATION] los cerró el paso `clarify` del workflow (sesión 2026-10-24, cinco preguntas).
