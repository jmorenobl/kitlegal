# Qué juzga la ronda 2 de la revisión final

Ciclo 1: revisión final del hito. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 1, sobre `24aaf80` (juez A: rechazado; juez B: aprobado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `24aaf80..HEAD` (`9243c55`):

- `9243c55` fix(H9): motivos de la revisión final — paso `corrector_revision`: `CHANGELOG.md`, `README.md`, `data/festivos/nacionales-y-autonomicos.yaml`, `internal/app/plazos_casos_test.go`, `internal/source/boe/fiestas/resolucion.go`, `internal/source/boe/fiestas/resolucion_test.go`, `specs/023-h9-internal-core-plazos/contracts/casos-de-plazos.md`, `specs/023-h9-internal-core-plazos/contracts/datos-de-plazos.md`, `specs/023-h9-internal-core-plazos/research.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/023-h9-internal-core-plazos/gates/revision-cambios-r2.diff` (324 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
