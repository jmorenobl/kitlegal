# Qué juzga la ronda 2 de la revisión final

Ciclo 1: revisión final del hito. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 1, sobre `a53bf9a` (juez A: rechazado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `a53bf9a..HEAD` (`b0d7e159`):

- `b0d7e159` fix(H41.3): motivos de la revisión final — paso `corrector_revision`: `internal/evals/laboratorio_tabla.go`, `internal/evals/laboratorio_tabla_test.go`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/037-h41-3-por-que-normativa/gates/revision-cambios-r2.diff` (193 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
