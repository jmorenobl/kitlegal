# Specification Quality Checklist: H26.1 · Las evals sirven lo que la skill lleva a leer

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-21
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec fija contratos observables (claves de la eval y
  del informe, códigos, motivos); los únicos nombres de código son los que el propio hito impone (`LeerConjunto`,
  `plazos.LeerTipos`, `skills.LeerCobertura`, `internal/evals/medida.go`) y el test que ya hace cumplir el tope del job
  (FR-020).
- [x] Focused on user value and business needs — «Contexto» y cada historia dicen qué gana `legal-core` o
  `boe-legislacion`, o quien lee el informe del job.
- [x] Written for non-technical stakeholders — en la medida en que el consumidor es quien lee el informe del job; cada
  historia explica el porqué antes del criterio.
- [x] All mandatory sections completed — escenarios, requisitos (FR-001 a FR-032), criterios de éxito (SC-001 a
  SC-021, con FR-033), supuestos y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — las dos que dejaba el spec (FR-017 y SC-013) las cerró clarify, y su
  respuesta queda en «Clarifications» junto a la de FR-029 y FR-030.
- [x] Requirements are testable and unambiguous — cada FR lleva su «Criterio:» con el test o la comprobación que lo
  distingue; FR-015 fija que la llamada sin sobre va a `otras_fallidas` aunque consulte una lectura de `sin_servir` y
  no es lectura intentada (FR-023, FR-033), y que la orden de salida no escrita, que sí llegó a la lectura, va a
  `lecturas_sin_servir` con su motivo y es lectura intentada (FR-024, FR-033); y FR-016 que la invocación
  que satisface el comando lee el bloque en su `fecha`.
- [x] Success criteria are measurable — SC-011 a SC-021 nombran su umbral y la comprobación que sale en rojo; SC-013
  la hace cumplir `lectura_sin_intentar` (FR-033).
- [x] Success criteria are technology-agnostic (no implementation details) — se miden en el informe del job y en
  `make ci`, que son la superficie del proyecto.
- [x] All acceptance scenarios are defined — seis historias en Dado/Cuando/Entonces.
- [x] Edge cases are identified — «Edge Cases», con el requisito que fija cada uno, también la lectura de `sin_servir`
  en otra fecha y la que es a la vez una llamada sin sobre (no intentada) o una orden de salida no escrita (intentada).
- [x] Scope is clearly bounded — «Fuera de alcance» lleva lo del hito y lo que el spec deja fuera.
- [x] Dependencies and assumptions identified — «Assumptions».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — «Criterio:» en cada FR.
- [x] User scenarios cover primary flows — caché derivada (US1), `sin_servir` y las evals 11 y 12 (US2), `lecturas`
  (US3), recuento (US4), motivo (US5) y la skill (US6).
- [x] Feature meets measurable outcomes defined in Success Criteria — los «Controles» y la «Aceptación» del hito, tal
  como los escribe, están en SC-001 a SC-020.
- [x] No implementation details leak into specification — ver el primer punto.

## Notes

- Clarify (2026-10-21) cerró los dos [NEEDS CLARIFICATION] y una tercera duda sobre qué mide el job de FR-029 y FR-030;
  los límites conocidos que dejan están en «Fuera de alcance» y los enseña el informe final.
