# Qué juzga la ronda 4 de la revisión final

Ciclo 2: cambios fuera de gates/ posteriores al último veredicto. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 3, sobre `0c9101a` (juez A: aprobado; juez B: aprobado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `0c9101a..HEAD` (`e3d6426`):

- `e3d6426` fix(H9): cierre en la plataforma — paso `reparar_cierre`, tras una medición del cierre en rojo: `CHANGELOG.md`, `skills/legal-core/SKILL.md`, `specs/023-h9-internal-core-plazos/contracts/skill-legal-core.md`, `specs/023-h9-internal-core-plazos/quickstart.md`, `specs/023-h9-internal-core-plazos/research.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/023-h9-internal-core-plazos/gates/revision-cambios-r4.diff` (85 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.

El rango trae una reparación del cierre: la medición de CI y evals que la motivó está en `gates/cierre.json`, `gates/cierre.log` y `gates/evals/*.json` (con sus `umbrales`), y sus decisiones, en las líneas `reparar_cierre` de `gates/supuestos.md`. Tras esta revisión, el workflow vuelve a medir la cabeza.
