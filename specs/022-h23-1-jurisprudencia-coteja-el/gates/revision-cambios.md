# Qué juzga la ronda 2 de la revisión final

Ciclo 1: revisión final del hito. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 1, sobre `daa6867` (juez A: rechazado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `daa6867..HEAD` (`86d4952`):

- `86d4952` fix(H23.1): motivos de la revisión final — paso `corrector_revision`: `internal/evals/informe_test.go`, `internal/evals/sentencias_test.go`, `specs/021-h26-el-umbral-de/gates/modelos.jsonl`, `specs/022-h23-1-jurisprudencia-coteja-el/cierre.md`, `specs/022-h23-1-jurisprudencia-coteja-el/contracts/skill-jurisprudencia.md`, `specs/022-h23-1-jurisprudencia-coteja-el/contracts/umbral-documento-sin-cotejar.md`, `specs/022-h23-1-jurisprudencia-coteja-el/mutantes.md`, `specs/022-h23-1-jurisprudencia-coteja-el/plan.md`, `specs/022-h23-1-jurisprudencia-coteja-el/quickstart.md`, `specs/022-h23-1-jurisprudencia-coteja-el/research.md`, `specs/022-h23-1-jurisprudencia-coteja-el/spec.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/022-h23-1-jurisprudencia-coteja-el/gates/revision-cambios-r2.diff` (351 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
