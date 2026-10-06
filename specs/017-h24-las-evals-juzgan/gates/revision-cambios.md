# Qué juzga la ronda 2 de la revisión final

Ciclo 1: revisión final del hito. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 1, sobre `81d025d` (juez A: aprobado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `81d025d..HEAD` (`10484b5`):

- `10484b5` fix(H24): motivos de la revisión final — paso `corrector_revision`: `internal/evals/conjunto_test.go`, `internal/evals/definicion.go`, `internal/evals/definicion_test.go`, `internal/evals/informe.go`, `internal/evals/informe_test.go`, `internal/evals/juez.go`, `internal/evals/medida.go`, `internal/evals/medida_test.go`, `internal/evals/sondeo.go`, `specs/017-h24-las-evals-juzgan/cierre.md`, `specs/017-h24-las-evals-juzgan/data-model.md`, `specs/017-h24-las-evals-juzgan/tasks.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/017-h24-las-evals-juzgan/gates/revision-cambios-r2.diff` (372 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
