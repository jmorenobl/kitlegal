# Informe del hito H7.1 · `graph check` acotado a la pregunta, con señales que se apagan y salida legible; H7 sin lo que no pasa el umbral (ADR 0028)

Generado por el workflow `hito` el 2026-09-29T07:51:07Z, sobre `a8aeb0d` de `011-h7-1-graph-check-acotado`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/011-h7-1-graph-check-acotado/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige la sección del hito en docs/ROADMAP.md y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `a8aeb0d`: verde.
- **Evals por skill**: boe-legislacion aprobado, legal-core aprobado (tasas en la sección 3).
- **Revisión final**: juez A aprobado, juez B aprobado, 1 rondas.
- **Tareas**: 25 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 30 commits; 136 files changed, 12052 insertions(+), 5543 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- plan: cómo se pasan la norma y los bloques a `graph check` (FR-001 lo deja al plan) → argumentos de posición y opcionales, `kitlegal graph check [<norma> [<bloques>...]]`, en el mismo orden que `boe articulos`, con la norma como `*string` para que una norma vacía salga con 2 y no se lea como «sin argumentos»; los errores de forma salen con 2 y «argumentos inválidos: …»; se descartan las banderas `--norma`/`--bloque` (research D6).
- plan: forma de `data` de `graph check` (FR-012 deja los nombres al plan) → objeto plano `{norma, bloques, version-obsoleta, fuente-caducada, omitidos, hallazgos}`, con `norma` vacía y `bloques` `[]` sin argumentos; la forma anidada (`ambito`, `totales`) daba 3 818 bytes con cinco bloques cambiados, por encima de los 3 800 de SC-005 (research D9).
- plan: mensaje de un `world.db` que no se puede usar → `grafo: "<ruta>" no es una base de datos utilizable: <causa>`, sin «; no se modifica», porque FR-070 retira la promesa sobre sus bytes; el de esquema posterior la conserva (research D19).
- plan: creación de `world.db` en su sitio (FR-071) → directorio en 0700 y fichero en 0600 como en H7 y como la caché; una entrega que falla tras crearlos deja el directorio y, como mucho, un `world.db` sin esquema, que la lectura trata como grafo vacío (research D18).
- plan: `fuente-caducada` solo sobre nodos `Norma`, `Bloque` y la redacción vista de cada bloque, lectura literal de FR-030; con los emisores de hoy no cambia nada más, porque `territorio` no declara vigencia (research D5).
- plan: primera lectura de un bloque → se guarda `(v, v)` y la entrega escribe una fila de `lecturas` solo si cambia, para que una entrega repetida idéntica siga sin escribir nada (H7 FR 022) (research D1).
- T016: contracts/applet-graph.md §5 pide el texto «sin tabuladores ni secuencias de escape», pero un valor de `data` puede llevarlos —un bloque pedido a `check`, que solo se rechaza vacío o en blanco (`graph check BOE-A-… $'a\t1'`), o un dato de un nodo leído de una fuente— y no dice cómo se escribe entonces → el valor que lleva algún carácter de control (`unicode.IsControl`: tabulador, saltos, ESC, DEL, C1) va entre comillas y con ellos escapados, con `strconv.Quote`, como lo nombran los mensaje… (entera en `specs/011-h7-1-graph-check-acotado/gates/supuestos.md` o `clarify-respuestas.json`)
- T016: la plantilla de la cabecera de `check` (§5.3) solo muestra plurales («se listan 4 y se omiten 0») y el singular y el plural por el número solo se fijan para `stats` (§5.1) → la cabecera concierta igual el verbo con el número: «se lista 1», «se omite 1», y en plural con cualquier otro, también 0; los nombres de clase (`1 version-obsoleta`) no se pluralizan, porque son las claves de `data`.

**Alcance (lo que queda fuera o dentro del hito)**

- corrector_spec: el spec mantenía el orden inverso de H7 US2.3 como caso límite, MUST NOT propio (FR-027) y caso de FR-023, y el hito dice que no pasa el umbral → queda fuera sin comportamiento ni test propios: FR-027 solo lo retira (sale con FR-080; pedir una redacción anterior es `--fecha`, de H20) y el «si y solo si» de FR-023 decide sin enumerarlo.
- plan: `README.md` (121-128) y `CONTRIBUTING.md` (335-351) se corrigen donde el hito los deja falsos (la skill ya no comprueba antes de leer ni traslada `fuente-caducada`; la publicación con temporal y los tests retirados); `CLAUDE.md` no se toca.
- corrector_plan: [n] el plan declaraba y conservaba el test de una `schema_version` negativa, que solo deja quien edita `world.db` a mano, y la misma clase en el lote (fecha que no es RFC 3339, vigencia negativa o con fracción de segundo, datos sin forma JSON canónica: ningún emisor los produce) → salen del contrato como casos y sus subtests se retiran (`TestMigrar`, `probarRechazosDeLaProcedencia`, `TestApplyRechazaElLote`); el código que los rechaza se queda como regla genérica, porque FR-075 c… (entera en `specs/011-h7-1-graph-check-acotado/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_plan: [n] la retirada de la ronda 1 dejaba tests de la misma clase en `internal/core/grafo` y en `internal/app` (fecha imposible guardada o llegando en `observacion_test` y `comprobar_test`; datos sin forma JSON en `persona_test`, `canonico_test`, `lote_test` y `salida_test`; `baseConUnaFechaIlegible`) → salen todos, caso a caso, y su código (`FusionarNodo/Arista/Texto`, `indexar`, `validarPersona`, `DatosCanonicos`, `Ficha.MarshalJSON`) se queda como regla genérica sin test propio; lo… (entera en `specs/011-h7-1-graph-check-acotado/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_plan: [n] el inventario de la ronda 2 dejaba casos de entradas del lote que ningún emisor produce (operación nula o que no es un valor, en `lote_test` y `errores_test`; una arista delante de sus extremos; un id repetido con otros datos; el nombre de un nodo sin id o de una arista, que solo llevan rechazos que FR-075 retira) → salen caso a caso (research D20, V29); su código se queda como regla genérica salvo el caso de la arista de `nombrar`, que sale con esos rechazos; `TestLecturas` … (entera en `specs/011-h7-1-graph-check-acotado/gates/supuestos.md` o `clarify-respuestas.json`)
- T025: FR-096 pide también «el diff sin retroceder respecto de la base» (el estado `patch` de `codecov.yml`, `target: auto`), pero la tarea solo fija como cifras la global (≥ 70 %) y la de `internal/core/**` (≥ 85 %), y las 49 líneas del diff que ningún perfil ejecuta son todas ramas de error de la regla genérica (FR-073; research D20, D21: fallos de SQLite, de `os.OpenFile` sobre lo que no es un fichero, de fechas guardadas que ninguna entrega escribe, de un JSON que no se puede codificar y de u… (entera en `specs/011-h7-1-graph-check-acotado/gates/supuestos.md` o `clarify-respuestas.json`)

**Skill (lo que pide, dice o comprueba una skill)**

- plan: la tabla de comandos de `SKILL.md` escribe un argumento opcional como Kong, `[<norma> [<bloques>...]]`, en lugar de `[--nombre]`, que describiría una orden inexistente; ninguna otra tabla tiene argumentos opcionales (research D7).
- plan: la eval 19 cambia con la implementación del formato de eval y no en la tarea `[aceptacion]`, porque con `hallazgos` y la norma de la comprobación no valida contra el esquema de hoy y `make ci` quedaría en rojo (research D15).
- T023: contracts/evals-y-skill.md §4 da la introducción reescrita de «Memoria de consultas» (qué recuerda `kitlegal` y qué devuelve `graph check` con la norma y los bloques) sin decir qué pasa con la frase de v0.1 «De los verbos de `kitlegal graph`, el protocolo solo usa `check`», ni con «cada uno con su `clase`» → se conservan las dos, la primera con «una vez por norma citada, cuando ya no queda nada por leer y antes de responder (paso 5)» en lugar de «dos veces…», porque FR-047 no deja cambiar … (entera en `specs/011-h7-1-graph-check-acotado/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run


- Observaciones de los jueces, por debajo del umbral y sin corregir: 11 en `spec-r2.json`, 15 en `plan-r4.json`, 6 en `tasks-r2.json`, 5 en `revision-a-r1.json`, 9 en `revision-b-r1.json`.

### Internos (21)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T002 (1), T003 (3), T004 (3), T005 (2), T006 (1), T008 (1), T014 (1), T015 (1), T019 (1), T020 (1), T021 (1), T024 (2), tasks (3). Enteras en `specs/011-h7-1-graph-check-acotado/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**boe-legislacion**: aprobado sobre `a8aeb0d`; decide `claude-sonnet-5` con 2 de 3; 19 evals, 0 nuevas, 7 informativas.

| Eval | claude-sonnet-5 | claude-haiku-4-5-20251001 | Marca |
|---|---|---|---|
| `19-lpac-articulo-21-redaccion-cambiada.yaml` | 2/3 | — | cambiada · informativa |
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 3/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 3/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 3/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 2/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 2/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 3/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 3/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 3/3 |  |
| `11-no-activa-programacion.yaml` | 3/3 | 3/3 |  |
| `12-no-activa-acuerdo-entre-amigos.yaml` | 3/3 | 3/3 |  |
| `13-lrbrl-atribuciones-por-materia.yaml` | 3/3 | — | informativa |
| `14-trlrhl-impuestos-por-materia.yaml` | 3/3 | — | informativa |
| `15-irpf-rendimientos-por-materia.yaml` | 3/3 | — | informativa |
| `16-lrjsp-legalidad-por-materia.yaml` | 3/3 | — | informativa |
| `17-ltaibg-plazo-por-materia.yaml` | 2/3 | — | informativa |
| `18-lrjpac-norma-derogada.yaml` | 3/3 | — | informativa |

**legal-core**: aprobado sobre `a8aeb0d`; decide `claude-sonnet-5` con 2 de 3; 3 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5 | claude-haiku-4-5-20251001 | Marca |
|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 3/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 3/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 |  |

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016).

## 4. Revisión que la constitución reserva a la persona

- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h7-grafo-applet.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h7-grafo-codigos.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h7-grafo-concurrencia.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h7-grafo-entrega-fallida.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h7-grafo-fuente-caducada.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h7-grafo-no-emiten.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h7-grafo-show.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h7-grafo-version-obsoleta.txtar`.
- Fixture o esquema EXISTENTE modificado: `schemas/eval.yaml.json`.
- Fixture o esquema EXISTENTE modificado: `schemas/grafo.json`.
- Fixtures nuevos (1), por directorio:
  - `internal/app/testdata/derivadas/version-ulterior/`: `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json`
- Suite de aceptación activada y congelada (4 guiones, escritos desde el spec en la primera tarea): `h7-1-grafo-check-acotado.txtar`, `h7-1-grafo-lecturas.txtar`, `h7-1-grafo-legible.txtar`, `h7-1-grafo-regla-generica.txtar`.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

| Requisito | Tareas | Aceptación | Estado |
|---|---|---|---|
| FR-001 | T001(hecha) T014(hecha) | grafo-check-acotado.txtar  | hecho |
| FR-002 | T001(hecha) T012(hecha) T014(hecha) | grafo-check-acotado.txtar  | hecho |
| FR-003 | T001(hecha) T012(hecha) T014(hecha) | grafo-check-acotado.txtar  | hecho |
| FR-004 | T001(hecha) T014(hecha) | grafo-check-acotado.txtar  | hecho |
| FR-005 | T001(hecha) T012(hecha) T014(hecha) | grafo-check-acotado.txtar  | hecho |
| FR-006 | T001(hecha) T014(hecha) | grafo-check-acotado.txtar grafo-regla-generica.txtar  | hecho |
| FR-007 | T001(hecha) T013(hecha) T014(hecha) | grafo-check-acotado.txtar grafo-lecturas.txtar grafo-legible.txtar grafo-regla-generica.txtar  | hecho |
| FR-010 | T001(hecha) T014(hecha) | — | hecho |
| FR-011 | T001(hecha) T008(hecha) T014(hecha) | grafo-check-acotado.txtar grafo-lecturas.txtar  | hecho |
| FR-012 | T001(hecha) T014(hecha) | grafo-check-acotado.txtar  | hecho |
| FR-013 | T001(hecha) T015(hecha) | — | hecho |
| FR-014 | T001(hecha) T014(hecha) | grafo-lecturas.txtar  | hecho |
| FR-015 | T014(hecha) | — | hecho |
| FR-020 | T001(hecha) T006(hecha) T007(hecha) | grafo-lecturas.txtar  | hecho |
| FR-021 | T006(hecha) T007(hecha) | — | hecho |
| FR-022 | T007(hecha) | — | hecho |
| FR-023 | T001(hecha) T008(hecha) | grafo-lecturas.txtar  | hecho |
| FR-024 | T001(hecha) T008(hecha) | grafo-lecturas.txtar  | hecho |
| FR-025 | T001(hecha) T006(hecha) T008(hecha) T009(hecha) T010(hecha) | grafo-lecturas.txtar  | hecho |
| FR-026 | T006(hecha) T007(hecha) T008(hecha) | — | hecho |
| FR-027 | T002(hecha) | — | hecho |
| FR-030 | T001(hecha) T009(hecha) | grafo-lecturas.txtar  | hecho |
| FR-031 | T001(hecha) T009(hecha) | grafo-lecturas.txtar  | hecho |
| FR-032 | T001(hecha) T009(hecha) | grafo-lecturas.txtar  | hecho |
| FR-040 | T013(hecha) T023(hecha) | — | hecho |
| FR-041 | T023(hecha) | — | hecho |
| FR-042 | T023(hecha) | — | hecho |
| FR-043 | T023(hecha) | — | hecho |
| FR-044 | T023(hecha) | — | hecho |
| FR-045 | T019(hecha) T023(hecha) | — | hecho |
| FR-046 | T023(hecha) | — | hecho |
| FR-047 | T023(hecha) | — | hecho |
| FR-048 | T024(hecha) | — | hecho |
| FR-050 | T020(hecha) T022(hecha) | — | hecho |
| FR-051 | T019(hecha) T020(hecha) | — | hecho |
| FR-052 | T020(hecha) | — | hecho |
| FR-053 | T022(hecha) | — | hecho |
| FR-054 | T018(hecha) T019(hecha) T020(hecha) | — | hecho |
| FR-055 | T021(hecha) T022(hecha) | — | hecho |
| FR-060 | T001(hecha) T016(hecha) T017(hecha) | grafo-legible.txtar grafo-regla-generica.txtar  | hecho |
| FR-061 | T001(hecha) T016(hecha) T017(hecha) | grafo-legible.txtar  | hecho |
| FR-062 | T001(hecha) T016(hecha) T017(hecha) | grafo-legible.txtar  | hecho |
| FR-063 | T001(hecha) T016(hecha) T017(hecha) | grafo-legible.txtar  | hecho |
| FR-064 | T001(hecha) T016(hecha) | grafo-legible.txtar  | hecho |
| FR-070 | T001(hecha) T002(hecha) T003(hecha) T004(hecha) | grafo-regla-generica.txtar  | hecho |
| FR-071 | T002(hecha) T003(hecha) | — | hecho |
| FR-072 | T003(hecha) T004(hecha) | — | hecho |
| FR-073 | T003(hecha) T004(hecha) | — | hecho |
| FR-074 | T005(hecha) | — | hecho |
| FR-075 | T005(hecha) | — | hecho |
| FR-076 | T005(hecha) | — | hecho |
| FR-077 | T003(hecha) T004(hecha) | — | hecho |
| FR-078 | T005(hecha) | — | hecho |
| FR-080 | T002(hecha) T014(hecha) T017(hecha) T025(hecha) | — | hecho |
| FR-081 | T002(hecha) T017(hecha) T025(hecha) | — | hecho |
| FR-082 | T014(hecha) T018(hecha) | — | hecho |
| FR-083 | T025(hecha) | — | hecho |
| FR-090 | T025(hecha) | — | hecho |
| FR-091 | T001(hecha) T010(hecha) T011(hecha) | grafo-check-acotado.txtar grafo-lecturas.txtar grafo-legible.txtar grafo-regla-generica.txtar  | hecho |
| FR-092 | T015(hecha) | — | hecho |
| FR-093 | T023(hecha) | — | hecho |
| FR-094 | T020(hecha) | — | hecho |
| FR-095 | T003(hecha) T004(hecha) T025(hecha) | — | hecho |
| FR-096 | T025(hecha) | — | hecho |
| FR-097 | T024(hecha) | — | hecho |
| FR-098 | — | — | SIN TAREA |
| SC-001 | T015(hecha) | — | hecho |
| SC-002 | T015(hecha) | — | hecho |
| SC-003 | T001(hecha) T008(hecha) | grafo-lecturas.txtar  | hecho |
| SC-004 | T001(hecha) T009(hecha) | grafo-lecturas.txtar  | hecho |
| SC-005 | T015(hecha) | — | hecho |
| SC-006 | T021(hecha) | — | hecho |
| SC-007 | — | — | SIN TAREA |
| SC-008 | T019(hecha) T020(hecha) T023(hecha) | — | hecho |
| SC-009 | T001(hecha) T014(hecha) | grafo-check-acotado.txtar  | hecho |
| SC-010 | T001(hecha) T016(hecha) T017(hecha) | grafo-legible.txtar  | hecho |
| SC-011 | T001(hecha) T003(hecha) T004(hecha) | grafo-regla-generica.txtar  | hecho |
| SC-012 | T008(hecha) | — | hecho |
| SC-013 | T003(hecha) T004(hecha) T005(hecha) T025(hecha) | — | hecho |
| SC-014 | T025(hecha) | — | hecho |

## 7. Cambios posteriores a la revisión final

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/011-h7-1-graph-check-acotado/quickstart.md`. Suite de aceptación congelada: `specs/011-h7-1-graph-check-acotado/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `411e4b8c`: 12 h 27 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh 411e4b8c`.
