# Qué juzga la ronda 4 de la revisión final

Ciclo 2: cambios fuera de gates/ posteriores al último veredicto. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 3, sobre `886d808` (juez A: aprobado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `886d808..HEAD` (`09cba100`):

- `09cba100` fix(H38): motivos de la revisión final (make ci en rojo) — paso `corrector_revision`: `CONTRIBUTING.md`, `specs/040-h38-las-skills-instaladas/plan.md`, `specs/040-h38-las-skills-instaladas/research.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/040-h38-las-skills-instaladas/gates/revision-cambios-r4.diff` (27 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
