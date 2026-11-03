# Specification Analysis Report — H39 (`031-h39-c-omo-se`)

Artefactos analizados: `spec.md` (409 líneas), `plan.md` (220), `tasks.md` (190), con `contracts/tabla.md`,
`contracts/laboratorio.md` y `.specify/memory/constitution.md` consultados en los puntos que lo pedían. Análisis de
solo lectura; el único fichero escrito es este.

Comprobado en el repositorio durante esta sesión (grep y lectura): los símbolos que citan las tareas existen
(`NombresDeHerramientas`, `UnionDeGrabaciones`, `rutaQueMide`, `ejecutarSesiones`, `InterpretarInvocacion`,
`LeerConjunto`, `LeerSesion`, `anotarBanderas`, `EsquemasDeHerramienta`, `ValidarDocumentoYAML`, `skills.Regenerar`,
`skills.Comparar`); las ayudas de argumento suman 28 (7 + 6 + 3 + 6 + 5 + 1 en `internal/app/{boe,cita,grafo,instalacion,plazos,territorio}.go`),
y las 6 de `instalacion.go` son de un applet sin herramienta, lo que cuadra con las 22 en verbos con herramienta de
T003; el texto de la ayuda de `bloque` en `boe.go:193` y el de `global` en `instalacion.go:165` coinciden con los de
T001; `.golangci.yml` exime a `^cmd/` de `forbidigo` (R4/R5), así que `cmd/laboratorio` cabe; `internal/evals` ya
importa `internal/app` (`preparar.go`), así que T005 y T009 no crean un ciclo; `laboratorio/` no existe todavía.
No medido (no se ejecutó nada del hito): que los tests y los guiones de las tareas pasen, que las 28 ayudas se
reflejen hoy en `--describe`, y los números de la evidencia de partida (440 sesiones, ejecución 41474010633).

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| U1 | Underspecification | MEDIUM | spec.md FR-021, «Fuera de alcance» (512 bytes); tasks.md T005; `internal/mcp/instrucciones.go:11-14` | La variante `instrucciones` añade su frase tras los 485 bytes de `mcp.Instrucciones` (queda en torno a los 565 bytes), es decir, detrás del byte 512 que `TestInstrucciones` y el comentario de `instrucciones.go` fijan como la parte que debe bastarse a sí misma. El spec no dice si la variante debe respetar ese tope ni qué se concluye si el efecto solo aparece pasado el byte 512. | Dejarlo anotado como supuesto: la variante mide «instrucciones del servidor» con un texto que el producto hoy no podría publicar, y el ADR de FR-051 tiene que decir qué haría falta cambiar (el tope) si esa es la variante que gana. No cambia ninguna tarea ni el alcance. |
| C1 | Coverage | LOW | spec.md FR-052; tasks.md «User Story 4» | FR-052 («el informe final lo dice») no tiene tarea; se cubre solo con `gates/supuestos.md`. Es deliberado y está dicho en tasks.md, pero depende de que el cierre del workflow lo recoja. | Sin acción: coherente con ADR 0032 y con el trato de H32. |
| C2 | Coverage | LOW | spec.md FR-033; tasks.md T010 | FR-033 (conservar la transcripción con variante, modo y eval) no está nombrado como comprobación en `TestArnesDelLaboratorio`; la ruta `DIRECTORIO/<variante>/<modo>/<eval>-<nn>/` sí. | Sin acción: la comprobación de esa ruta en T010 lo cubre. |
| I1 | Inconsistency | LOW | tasks.md líneas 147-148 y 169-171 | Las etiquetas `[datos]` y `[aceptacion]` aparecen también fuera de las líneas de tarea (ejemplo de paralelismo y notas). Según la memoria del proyecto, el workflow casa esas etiquetas en cualquier parte de la línea. Las líneas no empiezan por `- [ ] T…`, por lo que lo probable es que no se lean como tareas, pero no se ha medido. | Sin acción salvo que el analizador del workflow las trate como tareas; si lo hiciera, la causa es del proceso y no del hito. |
| A1 | Ambiguity | LOW | tasks.md T002 | T002 junta en una sola tarea el código de `describirArgumentos`, su test, siete esquemas, dos golden, una sección de un testscript y el CHANGELOG. Está justificado en su línea (`make ci` tiene que quedar en verde), pero es la más grande y la de mayor riesgo de cuarentena. | Sin acción: dividirla dejaría `make ci` en rojo entre las dos partes. |
| A2 | Ambiguity | LOW | tasks.md T009 | «uno o dos sondeos» de fixtures deja el número abierto. | Sin acción: el contrato de tabla (§6) fija qué debe cubrirse; el número es detalle del ejecutor. |

Sin hallazgos CRITICAL ni HIGH. Sin duplicaciones. Ningún marcador sin resolver (TODO, TKTK, `???`) en spec, plan ni tasks.

## Coverage Summary

| Requisito | ¿Tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001, FR-002, FR-003 | Sí | T001, T002, T003 | FR-002 lo comprueba T003 (research D12) |
| FR-004, FR-005 | Sí | T002 | CHANGELOG en T002; `skills/` sin tocar |
| FR-010 | Sí | T007 | |
| FR-011 | Sí | T005 | |
| FR-012 | Sí | T006 | |
| FR-013, FR-014 | Sí | T004 | |
| FR-020 a FR-027, FR-028 | Sí | T005 | |
| FR-029 | Sí | T008 | `TestInstrucciones` sin cambios; ver U1 |
| FR-030, FR-031, FR-033 | Sí | T010 | |
| FR-032 | Sí | T011 | |
| FR-040 a FR-043 | Sí | T009 | |
| FR-050, FR-051 | No (persona) | — | Fuera del run por diseño (FR-052) |
| FR-052 | No (informe) | — | `gates/supuestos.md` |
| FR-053 | No (cierre) | — | Job de evals del cierre; filas `evals:` en plan.md |
| SC-001 a SC-008 | Sí | T003, T002, T004, T005, T007, T009, T010, T001 y cierre | Cada uno con su control en «Controles de umbral» |
| SC-009 | No (persona) | — | Capa 3 |

## Constitution Alignment Issues

Ninguno. El «Constitution Check» de plan.md recorre los diez principios y las siete reglas; he comprobado contra el
repositorio las dos que más dependen del código: la exención de `^cmd/` para R4/R5 y la ausencia de dependencias
nuevas. Principio III respetado (T001 escribe la suite de aceptación antes del código y debe verla fallar); VIII (ninguna
skill del producto cambia).

## Unmapped Tasks

Ninguna. Las doce tareas se trazan a requisitos: T001 → US1; T002/T003 → FR-001 a FR-005; T004 → FR-013/FR-014; T005 →
FR-011/FR-020 a FR-028; T006 → FR-012; T007 → FR-010; T008 → FR-021/FR-024/FR-029; T009 → FR-040 a FR-043; T010 →
FR-030/FR-031/FR-033; T011 → FR-032; T012 → documentación (DoD §1.6).

## Metrics

- Requisitos funcionales: 32 (FR-001 a FR-053); criterios de éxito: 9
- Tareas: 12
- Cobertura de lo construible por el run: 100 % (28 de 32 FR con tarea; los otros cuatro —FR-050, FR-051, FR-052, FR-053— los hace una persona, el informe o el cierre del workflow, por diseño)
- Ambigüedades: 3 (U1, A1, A2)
- Duplicaciones: 0
- Incoherencias: 1 (I1, de proceso)
- Problemas CRITICAL: 0

## Next Actions

- No hay nada CRITICAL ni HIGH: se puede pasar a `implement`.
- U1 queda como supuesto para el ADR de FR-051 (el texto de `instrucciones.txt` supera el tope de 512 bytes de las
  instrucciones del producto); no exige cambiar tareas.
- El resto son LOW y no piden acción.
