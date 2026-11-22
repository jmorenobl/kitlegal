# Qué juzga la ronda 2 de la revisión final

Ciclo 1: revisión final del hito. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 1, sobre `424297c` (juez A: aprobado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `424297c..HEAD` (`10172e21`):

- `10172e21` fix(H36): motivos de la revisión final — paso `corrector_revision`: `internal/core/plazos/avisos_test.go`, `internal/core/plazos/computo.go`, `internal/core/plazos/tipos.go`, `specs/041-h36-el-motor-de/mutantes.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/041-h36-el-motor-de/gates/revision-cambios-r2.diff` (57 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
