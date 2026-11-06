# Qué juzga la ronda 2 de la revisión final

Ciclo 2: cambios fuera de gates/ posteriores al último veredicto. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 1, sobre `0169108` (juez A: aprobado; juez B: aprobado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `0169108..HEAD` (`7ebf7dde`):

- `7ebf7dde` fix(H41): cierre en la plataforma — paso `reparar_cierre`, tras una medición del cierre en rojo: `specs/034-h41-las-skills-del/research.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/034-h41-las-skills-del/gates/revision-cambios-r2.diff` (50 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.

El rango trae una reparación del cierre: la medición de CI y evals que la motivó está en `gates/cierre.json`, `gates/cierre.log` y `gates/evals/*.json` (con sus `umbrales`), y sus decisiones, en las líneas `reparar_cierre` de `gates/supuestos.md`. Tras esta revisión, el workflow vuelve a medir la cabeza.
