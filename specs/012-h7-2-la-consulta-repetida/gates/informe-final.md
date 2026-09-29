# Informe del hito H7.2 · La consulta repetida con evidencia coherente, y respuestas sin la maquinaria interna

Generado por el workflow `hito` el 2026-09-29T14:59:50Z, sobre `fbdab2e` de `012-h7-2-la-consulta-repetida`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/012-h7-2-la-consulta-repetida/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige la sección del hito en docs/ROADMAP.md y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `fbdab2e`: verde.
- **Evals por skill**: boe-legislacion aprobado, legal-core aprobado (tasas en la sección 3).
- **Revisión final**: juez A aprobado, juez B aprobado, 2 rondas.
- **Tareas**: 12 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 17 commits; 73 files changed, 6661 insertions(+), 411 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- corrector_spec: FR-060 prometía «no tener efectos fuera del directorio temporal», con una lectura universal (ni cachés de Go ni estado de Claude Code en el HOME) y otra enumerada → queda la enumerada como requisito (no cambiar árbol, índice ni historial de git, la instalación de kitlegal, las skills de la persona ni dejar conversaciones guardadas fuera del temporal); las cachés de compilación y de módulos de Go y el estado propio de Claude Code pueden cambiar, porque aislarlos exigiría llevar la… (entera en `specs/012-h7-2-la-consulta-repetida/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_spec: US3.7, US3.9 y FR-055 nombraban un veredicto «rechazado» que el job no da → el veredicto es el `fallo` que ya da `internal/evals/informe.go` (ADR 0016); no se añade ningún valor de veredicto, coherente con FR-054 (la regla del veredicto no cambia).
- corrector_revision: [f] el §6 de quickstart.md confiaba en `PATH="$d/bin:$PATH"` heredado para que las conversaciones ejecutaran el binario de la rama, y el Bash de Claude Code carga antes una instantánea del shell con el `PATH` de los ficheros de arranque de la persona → las conversaciones reciben `CLAUDE_ENV_FILE` con un guion de `"$d"` que antepone `"$d/bin"` (Claude Code 2.1.284 lo ejecuta tras la instantánea; una conversación real con el guion ejecutó el de la rama), y antes de abrirlas una… (entera en `specs/012-h7-2-la-consulta-repetida/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q2 (criterio c): Solo dentro del directorio temporal del escenario. El binario se construye desde la rama del hito en ese directorio y va primero en el PATH de las órdenes del escenario, de modo que la skill invoca ese `kitlegal` desde el PATH, como lo invocan las skills (ADR 0019). La skill se instala con ese mismo binario, `kitlegal skills install --host claude`, en ámbito local, en el directorio de trabajo de las conversaciones (`.agents/skills/` y el enlace relativo en `.claude/skill… (entera en `specs/012-h7-2-la-consulta-repetida/gates/supuestos.md` o `clarify-respuestas.json`)

**Alcance (lo que queda fuera o dentro del hito)**

- T012: la tarea solo pide tests si la cobertura global (≥ 70 %) o la de `internal/core/**` (≥ 85 %) quedan bajo su umbral, y no lo hacen (97,5 % y 98,6 %), pero el estado `patch` de `codecov.yml` (`target: auto`) compara el diff con la base, y la estimación local por líneas da 95,36 % (144 de 151), por debajo del 97,4 % que midió el cierre de H7.1 → no se añade ningún test: las siete líneas sin ejecutar son ramas de la regla genérica (leer un fichero regular que `os.ReadDir` acaba de listar, la m… (entera en `specs/012-h7-2-la-consulta-repetida/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q3 (criterio c): Un punto de entrada de prueba de `internal/evals` con la etiqueta de compilación `evals`, invocado desde el quickstart con `go test -tags evals -run '^…$' ./internal/evals/ -args …`, como `TestPrepararSesion`, que el escenario ya usa para preparar el directorio con la misma preparación que el job (FR-060). Lee las dos respuestas guardadas y comprueba las tres condiciones de FR-061 con el mismo código que el job: la lista de la skill con su comparación de FR-051, y la com… (entera en `specs/012-h7-2-la-consulta-repetida/gates/supuestos.md` o `clarify-respuestas.json`)

**Skill (lo que pide, dice o comprueba una skill)**

- clarify Q4 (criterio c): Se aplica y se publica, pero no entra en el recuento por modelo. La sesión de la prueba de red se juzga con la eval 01, que activa la skill, así que `Juzgar` le aplica la lista como a cualquier sesión (FR-052) y el informe publica sus expresiones encontradas, como las de cada sesión (FR-053; Aceptación del hito: «cada sesión publica sus expresiones prohibidas»). El recuento por modelo frente al umbral (FR-053, SC-001) se hace sobre las sesiones de las series que pide el … (entera en `specs/012-h7-2-la-consulta-repetida/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run


- Observaciones de los jueces, por debajo del umbral y sin corregir: 10 en `spec-r2.json`, 7 en `plan-r2.json`, 4 en `tasks-r1.json`, 7 en `revision-a-r2.json`, 10 en `revision-b-r2.json`.

### Internos (2)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T007 (1), T010 (1). Enteras en `specs/012-h7-2-la-consulta-repetida/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**boe-legislacion**: aprobado sobre `fbdab2e`; decide `claude-sonnet-5` con 2 de 3; 19 evals, 1 nuevas, 7 informativas.

| Eval | claude-sonnet-5 | claude-haiku-4-5-20251001 | Marca |
|---|---|---|---|
| `19-lcsp-contrato-menor-redaccion-cambiada.yaml` | 2/3 | — | **nueva** · informativa |
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 3/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 2/3 | 3/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 3/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 3/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 2/3 | 3/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 3/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 3/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 3/3 |  |
| `11-no-activa-programacion.yaml` | 3/3 | 3/3 |  |
| `12-no-activa-acuerdo-entre-amigos.yaml` | 3/3 | 3/3 |  |
| `13-lrbrl-atribuciones-por-materia.yaml` | 1/3 | — | informativa |
| `14-trlrhl-impuestos-por-materia.yaml` | 1/3 | — | informativa |
| `15-irpf-rendimientos-por-materia.yaml` | 0/3 | — | informativa |
| `16-lrjsp-legalidad-por-materia.yaml` | 3/3 | — | informativa |
| `17-ltaibg-plazo-por-materia.yaml` | 2/3 | — | informativa |
| `18-lrjpac-norma-derogada.yaml` | 3/3 | — | informativa |

**legal-core**: aprobado sobre `fbdab2e`; decide `claude-sonnet-5` con 2 de 3; 3 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5 | claude-haiku-4-5-20251001 | Marca |
|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 3/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 3/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 |  |

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016).

## 4. Revisión que la constitución reserva a la persona

- Fixtures nuevos (1), por directorio:
  - `testdata/evals/grafo-previo/lcsp-a1-30-redaccion-original/`: `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2017-12902_texto_bloque_a1-30.json`
- Esquemas nuevos: `schemas/expresiones-prohibidas.yaml.json`.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

| Requisito | Tareas | Aceptación | Estado |
|---|---|---|---|
| FR-001 | T007(hecha) | — | hecho |
| FR-002 | T007(hecha) | — | hecho |
| FR-003 | T007(hecha) | — | hecho |
| FR-004 | T007(hecha) | — | hecho |
| FR-010 | T006(hecha) | — | hecho |
| FR-011 | T006(hecha) T008(hecha) | — | hecho |
| FR-012 | T006(hecha) T008(hecha) | — | hecho |
| FR-013 | T006(hecha) T008(hecha) | — | hecho |
| FR-014 | T006(hecha) | — | hecho |
| FR-020 | T003(hecha) T007(hecha) T008(hecha) | — | hecho |
| FR-021 | T007(hecha) | — | hecho |
| FR-030 | T007(hecha) | — | hecho |
| FR-040 | T009(hecha) | — | hecho |
| FR-041 | T009(hecha) | — | hecho |
| FR-042 | T009(hecha) | — | hecho |
| FR-043 | T003(hecha) T009(hecha) | — | hecho |
| FR-044 | T009(hecha) | — | hecho |
| FR-045 | T009(hecha) | — | hecho |
| FR-046 | T009(hecha) | — | hecho |
| FR-047 | T009(hecha) | — | hecho |
| FR-048 | T011(hecha) | — | hecho |
| FR-050 | T001(hecha) T002(hecha) T003(hecha) T011(hecha) | — | hecho |
| FR-051 | T002(hecha) T003(hecha) T004(hecha) | — | hecho |
| FR-052 | T004(hecha) | — | hecho |
| FR-053 | T005(hecha) T011(hecha) | — | hecho |
| FR-054 | T004(hecha) T005(hecha) | — | hecho |
| FR-055 | T001(hecha) T002(hecha) T005(hecha) | — | hecho |
| FR-056 | T002(hecha) T005(hecha) T010(hecha) T012(hecha) | — | hecho |
| FR-060 | T010(hecha) | — | hecho |
| FR-061 | T010(hecha) | — | hecho |
| FR-062 | T010(hecha) T012(hecha) | — | hecho |
| FR-070 | T012(hecha) | — | hecho |
| FR-080 | T012(hecha) | — | hecho |
| FR-081 | T006(hecha) T008(hecha) | — | hecho |
| FR-082 | T001(hecha) T002(hecha) | — | hecho |
| FR-083 | T004(hecha) | — | hecho |
| FR-084 | T003(hecha) | — | hecho |
| FR-085 | T003(hecha) T007(hecha) | — | hecho |
| FR-086 | T011(hecha) | — | hecho |
| FR-087 | — | — | SIN TAREA |
| SC-001 | T005(hecha) T009(hecha) | — | hecho |
| SC-002 | T010(hecha) | — | hecho |
| SC-003 | T003(hecha) | — | hecho |
| SC-004 | T003(hecha) T007(hecha) | — | hecho |
| SC-005 | T006(hecha) T008(hecha) | — | hecho |
| SC-006 | T001(hecha) T002(hecha) T004(hecha) | — | hecho |
| SC-007 | T007(hecha) T008(hecha) T012(hecha) | — | hecho |
| SC-008 | T009(hecha) T011(hecha) T012(hecha) | — | hecho |

## 7. Cambios posteriores a la revisión final

Commits que ningún juez vio (correcciones del cierre y este informe):

- `fbdab2e` docs(H7.2): registros del run

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/012-h7-2-la-consulta-repetida/quickstart.md`. Suite de aceptación congelada: `specs/012-h7-2-la-consulta-repetida/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `6335565c`: 6 h 34 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh 6335565c`.
