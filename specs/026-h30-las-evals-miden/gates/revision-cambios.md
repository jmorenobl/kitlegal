# Qué juzga la ronda 2 de la revisión final

Ciclo 1: revisión final del hito. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 1, sobre `f660303` (juez A: rechazado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `f660303..HEAD` (`cc977b6`):

- `cc977b6` fix(H30): motivos de la revisión final — paso `corrector_revision`: `CHANGELOG.md`, `CONTRIBUTING.md`, `docs/ADR/0045-lo-reutilizado-no-cuenta-en-la-tasa.md`, `docs/WORKFLOW.md`, `internal/evals/decision_test.go`, `internal/evals/reutilizacion_test.go`, `specs/026-h30-las-evals-miden/cierre.md`, `specs/026-h30-las-evals-miden/contracts/eleccion-y-huella.md`, `specs/026-h30-las-evals-miden/data-model.md`, `specs/026-h30-las-evals-miden/plan.md`, `specs/026-h30-las-evals-miden/quickstart.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/026-h30-las-evals-miden/gates/revision-cambios-r2.diff` (103 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
