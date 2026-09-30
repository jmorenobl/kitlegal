# Qué juzga la ronda 2 de la revisión final

Ciclo 2: cambios fuera de gates/ posteriores al último veredicto. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 1, sobre `885a688` (juez A: aprobado; juez B: aprobado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `885a688..HEAD` (`5fcdb6f`):

- `5fcdb6f` fix(H7.4): cierre en la plataforma — paso `reparar_cierre`, tras una medición del cierre en rojo: `CHANGELOG.md`, `skills/boe-legislacion/SKILL.md`, `specs/014-h7-4-boe-legislacion-sin/cierre.md`, `specs/014-h7-4-boe-legislacion-sin/contracts/skill-boe-legislacion.md`, `specs/014-h7-4-boe-legislacion-sin/plan.md`, `specs/014-h7-4-boe-legislacion-sin/research.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/014-h7-4-boe-legislacion-sin/gates/revision-cambios-r2.diff` (170 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.

El rango trae una reparación del cierre: la medición de CI y evals que la motivó está en `gates/cierre.json`, `gates/cierre.log` y `gates/evals/*.json` (con sus `umbrales`), y sus decisiones, en las líneas `reparar_cierre` de `gates/supuestos.md`. Tras esta revisión, el workflow vuelve a medir la cabeza.
