# Qué juzga la ronda 4 de la revisión final

Ciclo 3: cambios fuera de gates/ posteriores al último veredicto. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 3, sobre `53dfec5` (juez A: rechazado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `53dfec5..HEAD` (`1a94ddd`):

- `1a94ddd` fix(H7.4): motivos de la revisión final — paso `corrector_revision`: `specs/014-h7-4-boe-legislacion-sin/research.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/014-h7-4-boe-legislacion-sin/gates/revision-cambios-r4.diff` (3 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
