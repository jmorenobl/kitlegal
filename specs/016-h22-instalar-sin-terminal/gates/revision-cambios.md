# Qué juzga la ronda 3 de la revisión final

Ciclo 2: cambios fuera de gates/ posteriores al último veredicto. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 2, sobre `a9c64f9` (juez A: aprobado; juez B: aprobado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `a9c64f9..HEAD` (`41ff279`):

- `41ff279` fix(H22): cierre en la plataforma — paso `reparar_cierre`, tras una medición del cierre en rojo: `CHANGELOG.md`, `skills/boe-legislacion/SKILL.md`, `specs/016-h22-instalar-sin-terminal/plan.md`, `specs/016-h22-instalar-sin-terminal/research.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/016-h22-instalar-sin-terminal/gates/revision-cambios-r3.diff` (87 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.

El rango trae una reparación del cierre: la medición de CI y evals que la motivó está en `gates/cierre.json`, `gates/cierre.log` y `gates/evals/*.json` (con sus `umbrales`), y sus decisiones, en las líneas `reparar_cierre` de `gates/supuestos.md`. Tras esta revisión, el workflow vuelve a medir la cabeza.
