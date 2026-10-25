# Qué juzga la ronda 2 de la revisión final

Ciclo 1: revisión final del hito. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 1, sobre `22d7962` (juez A: rechazado; juez B: aprobado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `22d7962..HEAD` (`3b987fd`):

- `3b987fd` fix(H31): motivos de la revisión final — paso `corrector_revision`: `CHANGELOG.md`, `CONTRIBUTING.md`, `docs/SKILLS.md`, `skills/boe-legislacion/SKILL.md`, `skills/boe-legislacion/references/lecturas_que_fallan.md`, `skills/legal-core/SKILL.md`, `specs/027-h31-las-skills-siguen/contracts/skills-reescritas.md`, `specs/027-h31-las-skills-siguen/plan.md`, `specs/027-h31-las-skills-siguen/research.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/027-h31-las-skills-siguen/gates/revision-cambios-r2.diff` (141 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
