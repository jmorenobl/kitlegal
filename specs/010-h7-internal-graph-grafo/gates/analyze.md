# Specification Analysis Report — H7 · `internal/graph`

Feature: `specs/010-h7-internal-graph-grafo` · Artefactos: `spec.md`, `plan.md` (con `research.md`, `data-model.md`, `contracts/`, `quickstart.md`), `tasks.md` (29 tareas) · Constitución 2.5.0.

Análisis de solo lectura: no se ha modificado ningún artefacto; este fichero es el único que se escribe (a petición expresa del modo autónomo). Ganchos `before_analyze` y `after_analyze` (`speckit.git.commit`, ambos opcionales): no se ejecutan; el commit de artefactos lo hace el workflow.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| F1 | Inconsistencia | LOW | spec.md FR-043 y *Assumptions* (l. 234, 350) frente a `data/territorio/dir3.yaml:4367`, contracts/emision.md §2, tasks.md T001 | El spec ilustra el DIR3 de Leganés (`ine:28074`) con `L01280740`; el dato congelado, y todo lo que lo cita (emision.md, arnes-e2e.md, quickstart, T001), es `L01280745`. El ejemplo del spec no se corresponde con ningún dato. | Sin cambio en tasks: T001 obliga a sacar el valor de `data/` y plan.md (obligación 11) prohíbe el valor de memoria. Si se toca el spec, poner `L01280745`. |
| F2 | Subespecificación | LOW | spec.md (Resumen, US5), tasks.md T027, `skills/boe-legislacion/SKILL.md` | «`boe-legislacion` v0.1» es la etiqueta del roadmap, pero ningún artefacto dice dónde consta esa versión: el frontmatter de `SKILL.md` no lleva campo de versión y T027 no lo añade. | Aceptable como nombre del hito; dejarlo en CHANGELOG (T028), que ya nombra «`boe-legislacion` v0.1». Sin tarea nueva. |
| F3 | Inconsistencia | LOW | tasks.md tabla de dependencias, T024 | T024 declara depender de T016 (registro), pero lo que compone —`EntregarAlGrafo` (T011) y `graph.Nuevo`/`ConDirectorio` (T009)— ya existe antes. La dependencia declarada es más fuerte que la real; no rompe el orden. | Ninguna acción; el orden secuencial del bucle es correcto. |
| F4 | Ambigüedad de proceso | LOW | plan.md l. 176-183 y 246-249 frente a gates/plan-pendiente.md (motivo «a») | El gate de plan pidió reescribir tres frases del Constitution Check («los estados… se evitan», «cotas… que valen para cualquier world.db», «sin ninguna violación»). El diseño ya corrige la causa (T008 y T009 no abren SQLite con 0 bytes; T010 lo fija en la matriz), pero las frases siguen como estaban. Con el diseño actual son veraces. | Ninguna acción; el motivo pendiente queda resuelto por el contenido, no por el texto. |
| F5 | Riesgo de ejecución | MEDIUM | tasks.md T007, T009, T011 (10 rutas o más cada una), T009 (tests con listas exactas de bytes de §4.1) | Tareas muy densas para el bucle de una tarea, un intento y cuarentena (constitución, «Reglas del modo desatendido»). No es una contradicción, pero es donde más probable es agotar los intentos. | Sin cambio: las rebanadas ya son las mínimas con test e implementación en el mismo diff (declarado en la cabecera de tasks.md). Si una entra en cuarentena, dividirla en la tarea que sigue a la que falló. |
| F6 | Cobertura | LOW | spec.md FR-035 (`--offline` no impide la entrega) | El requisito no tiene test unitario propio: lo cubren el guion `grafo-memoria` (plan.md l. 392 lo lista) y el hecho de que `--offline` no llega a la entrega (contracts/resultado-y-entrega.md §3). | Confirmar al escribir T001 que `grafo-memoria` (o `grafo-no-interferencia`) ejecuta una consulta con `--offline` desde la caché y la ve en `graph stats`. |

Ninguna violación de la constitución (I a IX), ningún marcador pendiente en spec, plan o tasks, ninguna dependencia nueva (`encoding/json/v2` y `jsontext` son de la biblioteca estándar y ya los usan tests del repositorio con Go 1.27), y ninguna contradicción entre requisitos.

## Tabla de cobertura

| Requisito | ¿Tiene tarea? | Tareas | Notas |
|-----------|---------------|--------|-------|
| FR-001 a FR-005 (almacén, ubicación, lectura sin crear) | Sí | T006, T007, T008, T009, T010, T015 | T006 exporta la regla de ubicación de la caché |
| FR-010 a FR-014 (errores, códigos, esquema, esperas) | Sí | T007, T008, T009, T010, T015 | 0 bytes con `-wal`: T008, T009, T010 |
| FR-020, FR-021 (operaciones en el `Resultado`, puerto) | Sí | T002 | El contrato `Applet` no cambia |
| FR-022 a FR-025 (transacción, fusión, rechazos, `Persona`) | Sí | T003, T004, T005, T009, T010 | 20 rechazos y 11 aceptaciones en T004 |
| FR-026 (entrega tras presentar) | Sí | T011 | |
| FR-030 a FR-035 (qué se entrega, `--no-graph`, fallos, `--dry-run`, `--offline`) | Sí | T001, T011, T016, T018 | Ayuda literal de `--no-graph` en T011 |
| FR-040 a FR-042 (emisión de `boe`) | Sí | T012, T018, T019 | Derivadas sin ELI en T019 |
| FR-043 a FR-045 (emisión de `territorio`, matriz territorial) | Sí | T013, T001 | Cubierto y no cubierto en `grafo-matriz-territorial` |
| FR-046, FR-047 (`skills` y `graph` no emiten; ejemplos del kernel) | Sí | T001, T002, T015, T016 | |
| FR-050 a FR-055 (applet `graph`, sobre, `show`, `stats`, legible) | Sí | T015, T016, T017 | Esquema publicado en T016, salida contra él en T017 |
| FR-060 a FR-067 (`check`, dos clases, explicaciones, reloj) | Sí | T014, T015, T019, T020 | Relojes T0, T1 y T8 del arnés en T020 |
| FR-070, FR-071 (texto por huella, nunca respuesta) | Sí | T012, T017 | |
| FR-080 a FR-083 (protocolo de la skill) | Sí | T027 | Después de la eval (T026), como pide DoD §1.10 |
| FR-085 a FR-087 (eval de consulta repetida, formato, informativa) | Sí | T022, T023, T024, T025, T026 | |
| FR-088 (integración con base temporal) | Sí | T010 | |
| FR-089 (procedencia igual a la del sobre) | Sí | T018 | |
| FR-090, FR-091 (e2e de cambio de versión y de no interferencia) | Sí | T001, T019, T020 | |
| FR-092 (arquitectura, R6) | Sí | T007 | |
| FR-093, FR-094 (aceptación, Definition of Done) | Sí | T001, T016, T027, T028, T029 | |
| FR-095 (régimen de `schemas/` y `testdata/`) | Sí | T016, T019, T022, T025, T029 | Las cuatro tareas `[datos]` y la lista de T029 |
| SC-001 a SC-006, SC-009 a SC-011 | Sí | T005, T009, T010, T014, T017, T018, T020 | Guiones de T001 |
| SC-007, SC-008 (costes) | Sí | T021 | `test-tiempos` en el Makefile |
| SC-012 (tasa de la eval informativa) | Sí | T026 | La mide el job en el cierre, sin tarea |
| SC-013 (`make ci`, cobertura, `skills-check`) | Sí | todas; T029 | |

## Alineación con la constitución

Sin incidencias. Comprobado: I (ninguna petición nueva; todo offline sobre grabaciones ya versionadas), II (nada entra sin fuente, FR-024, T004 y T018), III (T001 congela los 11 guiones antes de cualquier código; los tests van con cada tarea, y las excepciones de T017 y T018 están declaradas con su verificación por mutantes), IV (R6 nueva, `core` sin importar `graph`, solo `graph` abre SQLite junto a `cache` y `store`), V (ninguna dependencia nueva), VI (`--no-graph` gana efecto y ayuda), VII (`Persona` rechazada por regex dentro de `Apply`, T004), VIII (T026 antes que T027, `SKILL.md` de menos de 300 líneas), IX (matriz cubierto y no cubierto en `grafo-matriz-territorial`). Las etiquetas literales `[aceptacion]` y `[datos]` solo aparecen en T001, T016, T019, T022 y T025.

## Tareas sin requisito

Ninguna. T029 y T028 mapean a FR-094, FR-095 y SC-013; T021 a SC-007 y SC-008.

## Métricas

- Requisitos funcionales: 62 (FR-001 a FR-095, con los huecos de numeración del spec).
- Criterios de éxito: 13 (SC-001 a SC-013); todos con trabajo construible o test asignado.
- Tareas: 29 (1 de aceptación, 4 `[datos]`, 0 `[P]`).
- Cobertura (requisitos con al menos una tarea): 100 %.
- Ambigüedades: 1 (F2). Duplicaciones: 0.
- Incidencias críticas: 0. Altas: 0. Medias: 1 (F5). Bajas: 5.

## Próximas acciones

- Ninguna incidencia crítica ni alta: se puede pasar a `/speckit-implement`.
- Las cinco incidencias bajas y la media no exigen cambios en `spec.md`, `plan.md` ni `tasks.md`; F1 y F6 se resuelven al ejecutar T001 (valores tomados de `data/`, consulta con `--offline` en `grafo-memoria`).
