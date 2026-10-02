# Qué juzga la ronda 4 de la revisión final

Ciclo 2: cambios fuera de gates/ posteriores al último veredicto. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 3, sobre `41ff279` (juez A: aprobado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `41ff279..HEAD` (`b960a9b`):

- `b960a9b` fix(H22): motivos de la revisión final — paso `corrector_revision`: `CHANGELOG.md`, `skills/boe-legislacion/SKILL.md`, `specs/016-h22-instalar-sin-terminal/plan.md`, `specs/016-h22-instalar-sin-terminal/research.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/016-h22-instalar-sin-terminal/gates/revision-cambios-r4.diff` (81 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
