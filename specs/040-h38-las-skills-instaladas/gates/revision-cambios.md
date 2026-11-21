# Qué juzga la ronda 2 de la revisión final

Ciclo 1: revisión final del hito. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 1, sobre `2190df5` (juez A: rechazado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `2190df5..HEAD` (`be6e7483`):

- `be6e7483` fix(H38): motivos de la revisión final — paso `corrector_revision`: `internal/cache/candado.go`, `internal/core/instalacion/reinstalacion.go`, `internal/core/instalacion/reinstalacion_test.go`, `specs/040-h38-las-skills-instaladas/contracts/aviso-y-reinstalacion.md`, `specs/040-h38-las-skills-instaladas/data-model.md`, `specs/040-h38-las-skills-instaladas/plan.md`, `specs/040-h38-las-skills-instaladas/research.md`, `specs/040-h38-las-skills-instaladas/tasks.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/040-h38-las-skills-instaladas/gates/revision-cambios-r2.diff` (156 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
