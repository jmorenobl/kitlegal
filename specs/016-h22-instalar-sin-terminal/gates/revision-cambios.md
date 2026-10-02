# Qué juzga la ronda 2 de la revisión final

Ciclo 1: revisión final del hito. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 1, sobre `874d502` (juez A: aprobado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `874d502..HEAD` (`a9c64f9`):

- `a9c64f9` fix(H22): motivos de la revisión final (make ci en rojo) — paso `corrector_revision`: `internal/empaquetado/catalogo.go`, `internal/empaquetado/catalogo_test.go`, `internal/empaquetado/export_test.go`, `specs/016-h22-instalar-sin-terminal/contracts/paso.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/016-h22-instalar-sin-terminal/gates/revision-cambios-r2.diff` (23 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
