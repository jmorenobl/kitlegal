# Qué juzga la ronda 2 de la revisión final

Ciclo 1: revisión final del hito. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 1, sobre `ef2fc33` (juez A: rechazado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `ef2fc33..HEAD` (`11f87e86`):

- `11f87e86` fix(H33): motivos de la revisión final — paso `corrector_revision`: `data/cobertura.yaml`, `internal/skills/cubierto.go`, `internal/skills/cubierto_test.go`, `internal/skills/regiones.go`, `specs/039-h33-cobertura-en-positivo/spec.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/039-h33-cobertura-en-positivo/gates/revision-cambios-r2.diff` (40 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
