# Qué juzga la ronda 3 de la revisión final

Ciclo 1: revisión final del hito. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 2, sobre `9243c55` (juez A: rechazado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `9243c55..HEAD` (`0c9101a`):

- `0c9101a` fix(H9): motivos de la revisión final — paso `corrector_revision`: `specs/023-h9-internal-core-plazos/plan.md`, `specs/023-h9-internal-core-plazos/quickstart.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/023-h9-internal-core-plazos/gates/revision-cambios-r3.diff` (13 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
