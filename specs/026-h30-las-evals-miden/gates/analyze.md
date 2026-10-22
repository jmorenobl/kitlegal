# Specification Analysis Report — H30 (026-h30-las-evals-miden)

Artefactos analizados: `spec.md` (27 FR, 16 SC, 5 historias, 4 aclaraciones de la sesión 2026-10-22), `plan.md`,
`tasks.md` (T001 a T018) y, para contrastar, `research.md`, `contracts/`, la constitución 2.14.0 y el árbol del
repositorio. Solo lectura, `grep` y un cálculo binomial con `awk`; no se ha ejecutado ningún test, `make ci`, `gh` ni
sesión con modelo (ADR 0032), así que nada de lo que sigue es una medición de `make ci`.

## Hallazgos

| ID | Categoría | Gravedad | Ubicación | Resumen | Recomendación |
|----|-----------|----------|-----------|---------|---------------|
| F1 | Cobertura / riesgo de entorno | MEDIUM | tasks.md T004, T005, T006 (research D5); `.github/workflows/ci.yml` trabajo `ci` | La base de la propuesta se lee «en local» con `git` (`merge-base`, `git show <base>:`, `git ls-tree`) y T006 usa `git clone`. El trabajo `ci` de `ci.yml` hace el checkout sin `fetch-depth: 0` (solo los trabajos `snapshot` y `release` lo piden), luego en la plataforma el árbol es superficial. Un test que necesite historia pasa en el equipo del run y se pone rojo en la CI remota que corre al cerrar. | Al implementar T004/T005, que los tests construyan la base de la propuesta a partir de un árbol sintético o de una copia en `t.TempDir()` con su propio `git init`, nunca del historial del repositorio de trabajo; `git clone` local de un repositorio superficial sí funciona. |
| F2 | Granularidad de tarea | MEDIUM | tasks.md T013 | T013 junta en un solo diff `evals.yml` (cinco cambios), `definicion.go` (peor caso con 900 s), `scripts/evals.sh` y un test nuevo con mutante en `sustitutos_test.go`. Si no converge, entra en cuarentena y arrastra T014 a T018 (T014 depende de T013). Es la tarea con más superficie y la única que toca el job entero. | Mantenerla (el job no se puede dejar a medias en verde), pero implementarla en el orden `scripts/evals.sh` + su test → `evals.yml` → `definicion.go` + `TestDefinicionDelJob`, con `make ci` entre pasos. |
| F3 | Inconsistencia (cifra) | LOW | spec.md aclaración 4 y FR-024 («3 de 120») frente a `CLAUDE.md` («3 de 108») | La spec y las tareas dan *N* = 120 para `boe-legislacion`; `CLAUDE.md` sigue en 108. Con 108 y con 120 el límite de hoy es 3 (`0,028 × N` da 3,02 y 3,36), pero *p*ref cambia. Con *N* = 120, mi cálculo binomial da 0,0197 en 0,85 % y 0,0205 en 0,86 %, y con *N* = 60, 0,01998 en 0,36 %: coincide con las cifras de la spec. No he contado las respuestas del plan, así que *N* no está medido aquí. | T002 deriva *N* del plan y T016 copia la salida de `TestLimiteDeLaTasa` («manda la salida»), y lo anota en `gates/supuestos.md`: suficiente. `CLAUDE.md` queda fuera del hito (la batería lo prohíbe); su cifra se corrige en una propuesta aparte. |
| F4 | Ambigüedad | LOW | spec.md SC-003 | «sus 23 más las 12 de `legal-core` (y las 10 de `jurisprudencia`)»: la coma y el paréntesis hacen dudar de si son 35 o 45. Los supuestos de la spec y el árbol dicen 45 (`legal-core` nombra `boe-legislacion` 11 veces y `jurisprudencia` 1; ninguna skill nombra `legal-core`). | T005 compara con el listado de `evals/<skill>/`, no con una cifra: sin riesgo. Solo redacción. |
| F5 | Duplicación | LOW | spec.md FR-012, FR-013, FR-014 y «Edge Cases» | «Una eval sin resultado reutilizable se mide» está dicho cuatro veces con matices (90 días, 10 medidas, informe ilegible). | No reescribir (la spec está cerrada). T009 y T010 los cubren con subtests distintos. |
| F6 | Cobertura | LOW | spec.md FR-018; tasks.md T013 | FR-018 pide que `evals-medir-juez`/`medir_al_juez` sigan lanzando solo la medida del juez y que `prueba_de_red` conserve su sesión de red. T013 enumera lo nuevo pero no nombra esas dos aserciones. `definicion_test.go` ya las cubre (constantes en las líneas 96-97 y condiciones en la 1263 en adelante). | Al cambiar la condición de la tanda, no tocar esas aserciones; si hay que cambiarlas, FR-018 manda. |

Sin hallazgos CRITICAL ni HIGH.

## Cobertura

| Requisito | ¿Tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001 a FR-005 | Sí | T003, T004, T005 | Conjunto, común, descripción, rutas sin asignar, conjunto incalculable |
| FR-006 a FR-008 | Sí | T005, T008, T011 | Elección, rutas con renombrado, determinismo y golden |
| FR-009 | Sí | T006 | Huella |
| FR-010 a FR-015 | Sí | T007, T008, T009, T010, T013 | Medida de `main`, búsqueda, 90 días, tope de 10, veredicto |
| FR-016 a FR-021 | Sí | T010, T011, T012, T013 | Disparadores, matriz, concurrencia 5/4/2, tanda |
| FR-022, FR-023 | Sí | T007 | `origen` y fila de la serie sin cambios |
| FR-024 | Sí | T002, T007, T016 | Regla k(*n*), *p*ref y ADR 0045 |
| FR-025 | Sí | T008, T014, T015 | Release con suite completa |
| FR-026, FR-027 | Sí | T001, T007; T018 | Golden del informe; el diff sin `scripts/workflow/`, `.specify/` ni `internal/workflow/` |
| SC-001 a SC-013, SC-015, SC-016 | Sí | según «Trazabilidad» de tasks.md | |
| SC-014 | No, a propósito | — | La medición de cierre la hace el workflow tras la revisión final; tasks.md lo dice |

Los nombres de test de plan.md («Tests nuevos» y «Controles de umbral») y los de tasks.md coinciden; los de tasks.md
que no están en el plan son tests que ya existen (`TestFilaDeLaSerie`, `TestOrdenesDeLaTanda`, `TestRegistroDeLaTanda`,
`TestSalidaDeLaTanda`) o los que la batería prohíbe ejecutar (`TestSondeo`, `TestMedidaDelJuez`, `TestGrabar*`).

## Alineación con la constitución

Sin conflictos. Comprobado:

- Aceptación e2e: plan.md dice «Aceptación e2e: no aplica» con el motivo y tasks.md no tiene ninguna `[aceptacion]`,
  que es lo que exige `precheck.sh tasks`.
- Controles de umbral: cada fila de plan.md tiene su tarea y su test con el nombre exacto que busca el informe final.
- Ningún umbral sube (FR-024, FR-026): T002 estrecha o iguala, nunca sube; T016 lo deja escrito.
- Guardián de diff: las rutas declaradas incluyen los `x.go` de cada `x_test.go` que se toca (`guardian-diff.sh`
  admite `x_test.go` de un `x.go` declarado); `.github/*` solo se toca en T013 y T015, con las rutas declaradas, y no
  se toca `scripts/workflow/`, `.specify/` ni `internal/workflow/`.
- Principio X: la granularidad por skill (sin marca de materia) está decidida en las aclaraciones y en «Fuera de
  alcance»; no hay caso especial por materia.
- Reglas del modo desatendido: ninguna tarea usa red, `gh`, `claude` ni credenciales; la plataforma va en memoria o
  con el `gh` sustituto.
- ADR 0045: el último ADR del árbol es el 0044, así que el número es el siguiente libre.

## Tareas sin requisito

Ninguna. T001, T016, T017 y T018 declaran por qué no son rebanada completa y apuntan a FR/SC.

## Métricas

- Requisitos funcionales: 27; criterios de éxito: 16 (15 con tarea, SC-014 por el workflow).
- Tareas: 18 (ninguna `[P]`, `[datos]` ni `[aceptacion]`).
- Cobertura de FR con al menos una tarea: 27/27 (100 %).
- Ambigüedades: 1 (F4). Duplicaciones: 1 (F5). Problemas CRITICAL: 0. HIGH: 0. MEDIUM: 2. LOW: 4.

## Siguientes pasos

No hay nada que bloquee `/speckit-implement`. F1 y F2 son avisos para quien implementa T004/T005 y T013; F3 a F6 no
piden cambios en `spec.md`, `plan.md` ni `tasks.md`. Lo que no se ha podido medir en esta sesión: que los 45 conjuntos
sean calculables, que los mutantes den las cifras de SC-003, que *N* sea 120 y 60, y que `make ci` quede en verde; lo
miden T004, T005, T002 y cada tarea con su `make ci`.
