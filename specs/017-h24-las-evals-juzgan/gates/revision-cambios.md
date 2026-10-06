# Qué juzga la ronda 5 de la revisión final

Ciclo 2: cambios fuera de gates/ posteriores al último veredicto. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 4, sobre `b50e05b` (juez A: rechazado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `b50e05b..HEAD` (`51db15e`):

- `51db15e` fix(H24): motivos de la revisión final — paso `corrector_revision`: `specs/017-h24-las-evals-juzgan/cierre.md`, `specs/017-h24-las-evals-juzgan/contracts/skill-boe-legislacion.md`, `specs/017-h24-las-evals-juzgan/plan.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/017-h24-las-evals-juzgan/gates/revision-cambios-r5.diff` (45 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
