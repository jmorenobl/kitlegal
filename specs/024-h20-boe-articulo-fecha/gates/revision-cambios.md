# Qué juzga la ronda 2 de la revisión final

Ciclo 1: revisión final del hito. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 1, sobre `5e3c689` (juez A: aprobado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `5e3c689..HEAD` (`95f4949`):

- `95f4949` fix(H20): motivos de la revisión final — paso `corrector_revision`: `internal/source/boe/articulo.go`, `internal/source/boe/fecha.go`, `internal/source/boe/ids.go`, `internal/source/boe/ids_test.go`, `specs/024-h20-boe-articulo-fecha/cierre.md`, `specs/024-h20-boe-articulo-fecha/tasks.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/024-h20-boe-articulo-fecha/gates/revision-cambios-r2.diff` (63 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
