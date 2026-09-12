# Specification Analysis Report — H2 · `internal/httpx`

**Feature**: `specs/003-h2-internal-httpx-cliente/` | **Modo**: desatendido | **Fecha**: 2026-09-12

Artefactos analizados: `spec.md` (405 líneas, 65 FR, 16 SC, 6 US), `plan.md` (482 líneas), `tasks.md` (257
líneas, 19 tareas), `data-model.md`, `research.md` (22 decisiones D1–D22), `contracts/*.md` (4),
`.specify/memory/constitution.md`, `docs/ROADMAP.md` §1/§4-H2, `codecov.yml`, `Makefile`, `go.mod`,
`internal/arch_test.go`, `docs/PENDIENTES.md`. Lectura completa, no muestreada.

## Specification Analysis Report

| ID | Category | Severity | Location(s) | Summary | Recommendation |
|----|----------|----------|-------------|---------|----------------|
| I1 | Inconsistency | HIGH | research.md:136-162 (D3); plan.md tasks T006/T009 (`identificar.go`, `robots.go`); spec.md FR-006, FR-009 | La cadena de decoradores declarada es `identificar → robots → reintentar → ritmo → …`, con `identificar` **por encima** de `robots`. D3 dice explícitamente que para obtener `robots.txt`, el decorador `robots` «llama a **su propio siguiente** —la cadena **por debajo** de él—» (reintentar → ritmo → transporte), y `data-model.md` §12 lo repite («obtener robots.txt por la **cadena inferior**»). Como esa petición de `robots.txt` es fabricada dentro de `robots.go` y nunca vuelve a entrar por la cabecera de la cadena, la lógica de `identificar` —que solo se ejecuta para peticiones que **bajan desde arriba**— no la toca. Sin embargo, la misma sección D3 (línea 144-148) afirma que `identificar` pone el `User-Agent` «en **toda** petición que baja por la cadena, **incluida la de `robots.txt`**», y T006 promete lo mismo («… incluidas las que el propio paquete origina»). Ninguna pieza del diseño (spec, plan, research, data-model, contratos) asigna explícitamente a `robots.go` la responsabilidad de fijar la cabecera de identificación en la petición que él mismo fabrica; con la composición tal como está descrita, esa petición no pasaría por `identificar`. Esto compromete directamente FR-009, SC-001 y el criterio de aceptación literal del hito («ni sin UA»), además del ejemplo de fixture #10 de `plan.md` §Fixtures, que exige que una grabación de `robots.txt` lleve la cabecera de identificación. | No aplica remediación en este informe (modo desatendido); queda constancia de que el plan debe declarar explícitamente quién fija la identificación de la petición de `robots.txt` —lo natural es que `robots.go` invoque la misma función `AgenteDeUsuario()` que usa `identificar.go`— antes de implementar T009, y que `TestIdentificacionEnTodaPeticion` y/o `TestRobotsDeniegaLaRuta`/`TestRobotsPorSitio` verifiquen la cabecera en la petición de `robots.txt` misma, no solo en la del recurso. |
| E1 | Coverage Gap | MEDIUM | plan.md, tabla «Controles mecánicos», fila 6; tasks.md T006, T007, T008, T009 | El control #6 («Identificación en el 100 % de las peticiones») describe `TestIdentificacionEnTodaPeticion` como una prueba que cubre recurso, `robots.txt`, reintentos **y** saltos de redirección. Esa prueba se crea en **T006** (`identificar_test.go`), antes de que existan `ritmo.go` (T007), `reintentos.go` (T008) o `robots.go` (T009); ninguna tarea posterior declara volver a tocar `identificar_test.go` para ampliar esa cobertura una vez existen esos caminos. El checkpoint de la Fase 3 (tasks.md, tras T009) afirma que «los cinco controles literales del hito … tienen ya su test unitario sin red (SC-009)», dando por hecho que la cobertura de identificación en `robots.txt`/reintentos/redirecciones queda demostrada, pero ninguna tarea la asigna explícitamente a un fichero de test concreto fuera de `identificar_test.go` (que no la tiene) — depende de que `robots_test.go`, `reintentos_test.go` y `cliente_test.go` verifiquen la cabecera por su cuenta sin que ninguna tarea lo declare como obligación explícita. | Ninguna (informe de solo lectura). Se deja constancia para que, al escribir T007–T009, cada servidor de prueba compruebe también la cabecera de identificación en las peticiones que ejercita, no solo el comportamiento que la tarea nombra. |
| C1 | Inconsistency (menor) | LOW | tasks.md, tabla «Trazabilidad: requisito → tarea», fila `FR-053, FR-054, FR-055` | La fila atribuye FR-055 (regla R1: `internal/core` no importa `internal/httpx`) a **T009**, aunque R1 no cambia en H2 (plan.md, «Reglas de dependencia»: «Sin cambio») y T009 no toca ningún fichero de `internal/core`. Es una atribución de conveniencia («aquí es donde `httpx` empieza a importar algo y R1 se demuestra intacta»), no un error que bloquee el hito, pero rompe la lectura literal de la tabla de trazabilidad. | Ninguna (cosmético; no afecta al gate). |

**Sin hallazgos CRITICAL.** No se detectó ninguna violación de un MUST de la constitución, ningún artefacto mandatorio ausente, ningún requisito con cobertura cero que bloquee la funcionalidad base, ninguna dependencia fuera de la lista cerrada (constitución §V), ninguna cifra o ruta citada que no exista ya en el repositorio (`go.mod`, `Makefile` LDFLAGS, `codecov.yml`, `internal/arch_test.go`, `docs/ADR/`, `docs/PENDIENTES.md` verificados contra el estado real del árbol).

## Coverage Summary Table

Los 65 FR y 16 SC tienen ≥ 1 tarea asignada en `tasks.md` («Trazabilidad: requisito → tarea» y «SC-…»); se
verificó la tabla completa contra el texto de cada tarea (T001–T019) y no hay ningún rango sin tarea ni
ninguna tarea huérfana. Resumen por bloque (no se repite fila a fila lo que `tasks.md` ya traza con
exactitud):

| Bloque de requisitos | ¿Tiene tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001–FR-012 (construcción, plazo, UA, métodos) | Sí | T005, T006, T016 | I1 afecta a FR-009 dentro de este bloque |
| FR-013–FR-022 (`robots.txt`, ritmo) | Sí | T003, T007, T009 | — |
| FR-023–FR-035, FR-063 (reintentos, errores tipados) | Sí | T001, T004, T006, T008, T009, T014 | — |
| FR-036–FR-049, FR-064 (grabación/reproducción) | Sí | T010, T011, T012 `[datos]`, T013 | — |
| FR-050–FR-052, FR-065 (`--dry-run`) | Sí | T002, T006 | — |
| FR-053–FR-058 (arquitectura, dependencias) | Sí | T007, T009, T016 | C1 es una atribución menor dentro de este bloque |
| FR-059–FR-062 (adaptador de prueba) | Sí | T015, T016, T018 | — |
| SC-001…SC-016 | Sí | ver `tasks.md` tabla SC | SC-001 es la métrica que E1/I1 ponen en riesgo si no se corrige antes de implementar |

**Unmapped Tasks:** ninguna. Las 19 tareas (T001–T019) están todas ancladas a al menos un FR, SC o
historia de usuario.

**Constitution Alignment Issues:** ninguno. Los nueve principios y las cinco reglas de dependencia tienen
veredicto «✅ Cumple» en `plan.md` con evidencia verificable (Constitution Check + re-evaluación tras
diseño), y se contrastó contra el estado real del repositorio: `codecov.yml` ya trae los componentes
`internal_core` (85 %) e `internal_cli` (90 %) sin componente nuevo para `internal/httpx` (coherente con
«el hito no fija umbral propio»); `go.mod` no contiene todavía `golang.org/x/time` ni
`github.com/temoto/robotstxt` (coherente con «entran con el código que las usa», T007/T009); `Makefile`
ya tiene exactamente tres `-X` en `LDFLAGS` (coherente con «junto a las tres que ya existen», T005);
`internal/arch_test.go` en su estado actual **no** exige que el grafo contenga `internal/httpx` (confirma
que T016 tiene algo real que endurecer, no un placeholder); `docs/ADR/` llega hasta `0010`, sin colisión
con el `0011` que T002 crea; `docs/PENDIENTES.md` tiene la entrada «Antes de H2 · Dónde viven los fixtures
grabados» que T018 debe mover.

## Metrics

- **Total Requirements (FR)**: 65 (FR-001 a FR-065, todos presentes, incluidas las tres inserciones
  tardías FR-063/064/065; sin huecos de numeración funcional)
- **Total Success Criteria**: 16 (SC-001 a SC-016), todas de trabajo verificable (ninguna es una métrica
  de negocio post-lanzamiento; se excluyeron cero por ese motivo)
- **Total User Stories**: 6 (US1–US6), las seis con escenarios de aceptación y prueba independiente
- **Total Tasks**: 19 (T001–T019)
- **Coverage %** (requisitos con ≥ 1 tarea): 100 % (65/65 FR, 16/16 SC)
- **Ambiguity Count**: 1 (I1; sin adjetivos vagos ni marcadores sin resolver — `checklists/requirements.md`
  documenta tres rondas de juez que ya cerraron los huecos de ambigüedad detectables por texto)
- **Duplication Count**: 0 (no se encontraron requisitos casi duplicados; las referencias cruzadas
  FR-029↔FR-063, FR-015↔FR-030 son remisiones declaradas, no duplicación)
- **Critical Issues Count**: 0
- **High Issues Count**: 1 (I1)
- **Medium Issues Count**: 1 (E1)
- **Low Issues Count**: 1 (C1)

## Next Actions

- No hay CRITICAL: el hito **puede** avanzar a `/speckit-implement` en cuanto al gate de consistencia.
- Se recomienda, antes de ejecutar **T009** (el `robots.go` que introduce el hallazgo I1), que la tarea
  declare explícitamente qué componente fija la identificación de la petición de `robots.txt` y que
  `robots_test.go` verifique la cabecera en esa petición concreta (no solo el comportamiento de permitir/
  denegar). Es una precisión de implementación, no un cambio de requisito: FR-009 ya lo exige con
  claridad: lo que falta es la asignación de responsabilidad en el diseño interno.
- E1 se resuelve de la misma manera y en la misma tarea: basta con que los tests de T007–T009 aserten la
  cabecera de identificación en sus propios servidores de prueba.
- C1 no requiere acción; es una nota de lectura sobre `tasks.md`.
- Comandos sugeridos si se quiere dejar rastro formal antes de implementar: ninguno obligatorio — en modo
  desatendido, la corrección de I1/E1 se aplica directamente en el código de T009 con `make ci` en verde,
  tal como exige el modo de ejecución del workflow `hito` (sin fase de remediación separada).

## Extension Hooks

**Optional Hook**: git
Command: `/speckit-git-commit`
Description: Auto-commit after analysis

Prompt: Commit analysis results?
To execute: `/speckit-git-commit`
