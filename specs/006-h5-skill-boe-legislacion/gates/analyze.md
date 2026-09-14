# Specification Analysis Report

**Feature**: H5 · Skill `boe-legislacion` (genérica) + andamiaje de skills (`skills-sync`, evals)
**Artefactos analizados**: `spec.md` (FR-001–FR-085, SC-001–SC-012), `plan.md`, `tasks.md` (T001–T031), `.specify/memory/constitution.md` v1.3.0
**Modo**: desatendido, solo lectura. Verificación cruzada contra el estado real del repositorio (`.golangci.yml`, `go.mod`, `Makefile`) además de la consistencia interna entre los tres artefactos.

## Contexto del análisis

Este spec pasó ya por varias rondas de gate automático antes de esta ejecución: `spec-r1` a `spec-r4` (4 rondas del gate del spec, ver `gates/spec.json` y `checklists/requirements.md`) y `plan-r1` a `plan-r15` (15 rondas del gate del plan). `tasks.md` incluye su propia tabla de trazabilidad completa (FR-001–FR-085, SC-001–SC-012 → tareas) y una autocomprobación contra la rúbrica `juez_tasks` (criterios a-g). Por eso este análisis independiente parte de un artefacto ya muy depurado; se ha buscado deliberadamente más allá de lo que esas rondas ya cubren, incluyendo verificación empírica contra el repositorio (existencia de ficheros, contenido de `go.mod`, `Makefile` y `.golangci.yml`) para no limitarse a la coherencia textual entre documentos.

## Hallazgos

| ID | Categoría | Severidad | Ubicación(es) | Resumen | Recomendación |
|----|----------|----------|-------------|---------|----------------|
| I1 | Inconsistencia terminológica | LOW | plan.md:84-86 (Scale/Scope) | La palabra «guiones» se usa para dos conjuntos distintos en la misma lista: los 4 scripts de shell (`skills-sync.sh`, `instalar-skills.sh`, `grabar-evals.sh`, `evals.sh`) y, por separado, los «4 guiones `testscript`» (los 4 ficheros `.txtar` de T018). Ambos grupos suman casualmente 4, lo que hace más fácil leer la lista como si fuera un único conjunto de 4 en vez de dos conjuntos de 4. | Ninguna acción de código; si se retoca `plan.md`, nombrar el segundo grupo como «guiones `testscript` (`.txtar`)» o «casos de `testscript`» para diferenciarlo léxicamente del primero. No bloquea `/speckit-implement`. |
| U1 | Infraespecificación (documental, no funcional) | LOW | plan.md:66-67, 279-284 (Makefile objetivos) | `make skills-check` se describe como «una sola invocación de `go test` de los tests del repositorio» sobre `internal/skills`, `internal/evals` e `internal/app`, paquetes que ya están dentro de `./...` y por tanto ya los ejecuta `make test`. El plan no dice explícitamente que esto es intencional y redundante con `make test` (a diferencia de cómo sí lo hace para el mismo patrón en el `Complexity Tracking`, ítem «`make skills-check` y `make evals`», que solo justifica la existencia del objetivo, no la duplicación de cobertura). | Ninguna acción de código: el patrón replica el de `schema-check` (H0-H4), ya establecido y verificado contra el `Makefile` actual (`schema-check` también repite un subconjunto de `make test`). Es una convención del repositorio, no un defecto de este spec. Se deja anotado por transparencia, no como bloqueante. |

No se han encontrado hallazgos de severidad CRITICAL, HIGH o MEDIUM. En particular, no se detectó ninguna violación de un principio MUST de la constitución, ningún requisito con cobertura cero, ninguna contradicción de orden entre tareas, ningún marcador `TODO`/`NEEDS CLARIFICATION` sin resolver y ninguna divergencia numérica entre las cifras que `plan.md` declara en *Scale/Scope* (12 evals, 10 normas, 49 casos sintéticos / 195 ficheros, etc.) y lo que `tasks.md` efectivamente detalla tarea a tarea (verificado sumando los casos y ficheros de T020-T022: 15+21+13 = 49 casos, 42+32+121 = 195 ficheros).

Verificaciones empíricas adicionales contra el repositorio, todas conformes con lo que spec/plan/tasks afirman:
- `go.yaml.in/yaml/v3` y `.../v4` ya están en `go.mod` como indirectas (confirma V31/V58 y la justificación de *Complexity Tracking* de promover `v3` a directa).
- `grabacion` ya está en `run.build-tags` de `.golangci.yml` (confirma que T007 no necesita añadirlo, solo T026 añade `evals`).
- `make test` ejecuta `go test ./...` y `schema-check` ya repite ese patrón de subconjunto dedicado (contexto del hallazgo U1).

## Tabla de cobertura

Dado que `tasks.md` ya mantiene una tabla de trazabilidad requisito → tarea completa y verificada (FR-001 a FR-085, SC-001 a SC-012), este análisis no la duplica fila a fila; en su lugar se verificó por muestreo dirigido (bloques completos FR-001-015, FR-060-077, y las 12 SC) que cada requisito tiene al menos una tarea y que esa tarea efectivamente contiene el fichero o el test que lo cumple. Resumen agregado:

| Bloque de requisitos | Nº de FR | ¿Cobertura? | Tareas principales |
|---|---|---|---|
| La skill `boe-legislacion` (FR-001–FR-015) | 15 | Sí | T013, T016, T017 |
| `data/normas.yaml` y esquema (FR-020–FR-025) | 6 | Sí | T005, T007–T011 |
| Generación `skills-sync` (FR-030–FR-036) | 7 | Sí | T013–T017 |
| Comprobación mecánica en `make ci` (FR-040–FR-044) | 5 | Sí | T002, T010, T011, T017, T029 |
| `make install` (FR-050–FR-055) | 6 | Sí | T016, T018, T019 |
| Evals (FR-060–FR-066) | 7 | Sí | T001–T003, T012 |
| Job de evals (FR-070–FR-077) | 8 | Sí | T004, T006–T008, T012, T021–T027, T030 |
| Aceptación y cierre (FR-080–FR-085) | 6 | Sí | T028–T031 |
| **Total FR** | **60** | **100 %** | — |
| Success Criteria (SC-001–SC-012) | 12 | Sí (100 %) | T017, T025, T026, T029, T031 entre otras |

**Tareas sin requisito asociado**: ninguna. Las tareas sin una historia de usuario explícita (T028 documentación, T029 cierre, T030-T031 plataforma) llevan igualmente su propia fila en la tabla de trazabilidad de `tasks.md` (FR-083–FR-085, SC-011, SC-012, etc.), y el propio `tasks.md` lo declara así en su nota introductoria («Las tareas de documentación, cierre y plataforma no llevan historia»).

## Alineación con la constitución

Sin problemas. La tabla «Constitution Check» de `plan.md` (principios I-IX + reglas de dependencia + gates por capa) se revisó punto por punto contra el contenido real del spec y las tareas:
- Principio I (frontera humana): confirmado — ningún verbo nuevo con identidad, grabación solo en tarea `[datos]` con pausa (T008).
- Principio III (tests primero, offline): confirmado — cada tarea de código en `tasks.md` antepone su test; toda la batería es offline contra `httpx.Replay`.
- Principio V (simplicidad, dependencias fijadas): única divergencia declarada (`go.yaml.in/yaml/v3` en vez de `gopkg.in/yaml.v3`) justificada en *Complexity Tracking* con alternativas rechazadas y motivo — cumple el requisito constitucional de justificación explícita.
- Principio VIII (skills primero): confirmado — la skill se mide con 12 evals (10 positivas + 2 de no activación) y un job; ninguna herramienta nueva se enlaza en el binario (`internal/skills` e `internal/evals` son paquetes de test, no `package main`).
- Principio IX (genericidad territorial): confirmado — FR-012, FR-066 y la tarea T012/T017 excluyen explícitamente casos de municipio; el punto 11 de la Definition of Done se marca «solo en lo que prohíbe», correctamente, porque H5 no introduce ninguna herramienta con dimensión territorial.
- Gates por capa (mecánica/juez/humano): la capa 3 (humano) está correctamente acotada a T008 (manifiesto y grabación), T001/T009 (esquemas nuevos bajo `schemas/`) y T030/T031 (plataforma y fusión); ninguna tarea del bucle ejecuta `KITLEGAL_RECORD`, `scripts/grabar-evals.sh`, `make evals` ni `make verify-sources`, tal como exige la constitución.

No hay hallazgos que reabrir aquí.

## Tareas sin mapear

Ninguna.

## Métricas

- **Total de requisitos**: 72 (60 FR + 12 SC)
- **Total de tareas**: 31 (T001–T031; 7 `[datos]`, 2 `[plataforma]`, 22 de código/documentación/cierre)
- **Cobertura** (requisitos con ≥1 tarea): 100 %
- **Ambigüedades detectadas**: 0 (sin adjetivos vagos sin criterio medible; todas las SC son binarias o porcentuales)
- **Duplicaciones detectadas**: 0 requisitos duplicados; 1 duplicación terminológica menor (I1)
- **Incidencias CRITICAL**: 0

## Próximas acciones

No hay bloqueantes. Los dos hallazgos son LOW y puramente cosméticos/documentales sobre `plan.md`; ninguno afecta a `spec.md` ni a `tasks.md`, ninguno cambia el comportamiento de ninguna tarea, y ninguno requiere volver a `/speckit-specify`, `/speckit-clarify` o `/speckit-plan`. El feature está listo para `/speckit-implement` tal como está.

Si se quiere pulir por completitud (opcional, no bloqueante): renombrar en `plan.md:86` el segundo grupo de «guiones» a «guiones `testscript`» de forma más explícita en el propio Scale/Scope, y añadir una frase en el ítem «`make skills-check` y `make evals`» de *Complexity Tracking* aclarando que `skills-check` repite cobertura de `make test` a propósito (mismo patrón que `schema-check`).

## Modo desatendido

Por instrucción explícita del usuario, este informe no ofrece un plan de remediación para más adelante ni pregunta antes de continuar: los dos hallazgos LOW quedan registrados aquí como documentación, sin edición de ningún artefacto (análisis estrictamente de solo lectura).
