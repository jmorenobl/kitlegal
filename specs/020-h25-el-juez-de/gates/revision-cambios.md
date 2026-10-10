# Qué juzga la ronda 4 de la revisión final

Ciclo 2: cambios fuera de gates/ posteriores al último veredicto. Lo escribe `scripts/workflow/revision.sh cambios` sin modelo, al preparar a los jueces (ADR 0030).

Tu último veredicto es el de la ronda 3, sobre `0dc549c` (juez A: rechazado; juez B: rechazado). Lo que ha cambiado fuera de `gates/` desde entonces es el rango `0dc549c..HEAD` (`660391c`):

- `660391c` fix(H25): motivos de la revisión final — paso `corrector_revision`: `CHANGELOG.md`, `internal/app/skills_test.go`, `skills/boe-legislacion/SKILL.md`, `specs/020-h25-el-juez-de/cierre.md`, `specs/020-h25-el-juez-de/contracts/juez-de-jurisprudencia.md`, `specs/020-h25-el-juez-de/data-model.md`, `specs/020-h25-el-juez-de/plan.md`, `specs/020-h25-el-juez-de/research.md`.

El diff completo de ese rango, fuera de `gates/`, está en `specs/020-h25-el-juez-de/gates/revision-cambios-r4.diff` (369 líneas cambiadas). Para tu regla de CONVERGENCIA, todo el rango es «la corrección», sea del corrector de la revisión o de otro paso posterior a tu veredicto.
