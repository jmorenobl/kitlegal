# Qué juzga la ronda 2 de la revisión final

Ciclo 1: revisión final del hito. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 1, sobre `5cfda6e` (juez A: aprobado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `5cfda6e..HEAD` (`437d4931`):

- `437d4931` fix(H41.2): motivos de la revisión final — paso `corrector_revision`: `internal/evals/activaciones_test.go`, `internal/evals/informe_test.go`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/036-h41-2-por-qu-e/gates/revision-cambios-r2.diff` (49 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
