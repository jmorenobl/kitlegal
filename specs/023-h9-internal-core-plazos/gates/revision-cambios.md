# Qué juzga la ronda 5 de la revisión final

Ciclo 2: cambios fuera de gates/ posteriores al último veredicto. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 4, sobre `e3d6426` (juez A: aprobado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `e3d6426..HEAD` (`a172b3f`):

- `a172b3f` fix(H9): motivos de la revisión final — paso `corrector_revision`: `specs/023-h9-internal-core-plazos/research.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/023-h9-internal-core-plazos/gates/revision-cambios-r5.diff` (4 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
