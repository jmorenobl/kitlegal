# Qué juzga la ronda 2 de la revisión final

Ciclo 1: revisión final del hito. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 1, sobre `4c23eed` (juez A: rechazado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `4c23eed..HEAD` (`007c7fe`):

- `007c7fe` fix(H26.1): motivos de la revisión final — paso `corrector_revision`: `CHANGELOG.md`, `CONTRIBUTING.md`, `internal/evals/consultas.go`, `internal/evals/consultas_test.go`, `internal/evals/informe.go`, `internal/evals/medida.go`, `internal/evals/medida_test.go`, `specs/025-h26-1-las-evals-sirven/cierre.md`, `specs/025-h26-1-las-evals-sirven/contracts/cache-de-la-sesion.md`, `specs/025-h26-1-las-evals-sirven/data-model.md`, `specs/025-h26-1-las-evals-sirven/plan.md`, `specs/025-h26-1-las-evals-sirven/research.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/025-h26-1-las-evals-sirven/gates/revision-cambios-r2.diff` (248 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
