# Qué juzga la ronda 3 de la revisión final

Ciclo 1: revisión final del hito. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 2, sobre `10484b5` (juez A: aprobado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `10484b5..HEAD` (`ae250eb`):

- `ae250eb` fix(H24): motivos de la revisión final — paso `corrector_revision`: `internal/evals/conjunto_test.go`, `internal/evals/medida.go`, `specs/017-h24-las-evals-juzgan/cierre.md`, `specs/017-h24-las-evals-juzgan/tasks.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/017-h24-las-evals-juzgan/gates/revision-cambios-r3.diff` (14 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
