# Qué juzga la ronda 2 de la revisión final

Ciclo 1: revisión final del hito. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 1, sobre `062c399` (juez A: aprobado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `062c399..HEAD` (`4a127c5`):

- `4a127c5` fix(H25): motivos de la revisión final — paso `corrector_revision`: `internal/evals/medida.go`, `internal/evals/medida_test.go`, `specs/020-h25-el-juez-de/cierre.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/020-h25-el-juez-de/gates/revision-cambios-r2.diff` (116 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
