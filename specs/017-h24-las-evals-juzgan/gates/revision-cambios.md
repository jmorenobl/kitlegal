# Qué juzga la ronda 4 de la revisión final

Ciclo 2: cambios fuera de gates/ posteriores al último veredicto. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 3, sobre `ae250eb` (juez A: aprobado; juez B: aprobado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `ae250eb..HEAD` (`b50e05b`):

- `b50e05b` fix(H24): cierre en la plataforma — paso `reparar_cierre`, tras una medición del cierre en rojo: `CHANGELOG.md`, `skills/boe-legislacion/SKILL.md`, `specs/017-h24-las-evals-juzgan/contracts/skill-boe-legislacion.md`, `specs/017-h24-las-evals-juzgan/plan.md`, `specs/017-h24-las-evals-juzgan/research.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/017-h24-las-evals-juzgan/gates/revision-cambios-r4.diff` (86 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.

El rango trae una reparación del cierre: la medición de CI y evals que la motivó está en `gates/cierre.json`, `gates/cierre.log` y `gates/evals/*.json` (con sus `umbrales`), y sus decisiones, en las líneas `reparar_cierre` de `gates/supuestos.md`. Tras esta revisión, el workflow vuelve a medir la cabeza.
