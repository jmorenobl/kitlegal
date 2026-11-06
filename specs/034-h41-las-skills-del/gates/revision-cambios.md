# Qué juzga la ronda 3 de la revisión final

Ciclo 2: cambios fuera de gates/ posteriores al último veredicto. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 2, sobre `7ebf7dd` (juez A: rechazado; juez B: aprobado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `7ebf7dd..HEAD` (`a2268bb7`):

- `a2268bb7` fix(H41): motivos de la revisión final — paso `corrector_revision`: `specs/034-h41-las-skills-del/research.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/034-h41-las-skills-del/gates/revision-cambios-r3.diff` (70 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
