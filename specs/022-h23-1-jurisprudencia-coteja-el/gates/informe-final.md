# Informe del hito H23.1 · `jurisprudencia` coteja el documento antes de decir nada de él, y un umbral sin modelo lo hace cumplir (ADR 0038, punto 1)

Generado por el workflow `hito` el 2026-10-16T13:09:11Z, sobre `60e8678` de `022-h23-1-jurisprudencia-coteja-el`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/022-h23-1-jurisprudencia-coteja-el/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige la sección del hito en docs/ROADMAP.md y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `60e8678` (medición 1): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: boe-legislacion aprobado, jurisprudencia aprobado, legal-core aprobado (tasas en la sección 3).
- **Umbrales del job**: boe-legislacion: 13 umbrales, los 9 que deciden se cumplen y 4 solo se publican (2 por encima de su umbral: `cuenta_su_proceso:claude-sonnet-5-5:orden`, `afirma_lo_no_leido:claude-sonnet-5-5:herramienta`); jurisprudencia: 15 umbrales, los 11 que deciden se cumplen y 4 solo se publican (2 por encima de su umbral: `afirma_que_existe:claude-sonnet-5-5:orden`, `afirma_que_existe:claude-sonnet-5-5:herramienta`); legal-core: sin umbrales (sección 3).
- **Revisión final**: juez A aprobado (con `claude-opus-5-5`), juez B aprobado (con `claude-fable-5-1`), 2 rondas.
- **Sesiones con otro modelo que el de su rol, o sin saber cuál: 6** de 30 (ADR 0033):
  - `revision_juez_b` (2026-10-16T10:59:51Z): modelo real desconocido; `workflow.yml` da `claude-fable-5-1` a `modelo_adversario`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_a` (2026-10-16T10:59:51Z): modelo real desconocido; `workflow.yml` da `claude-opus-5-5` a `modelo_revisor`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_b` (2026-10-16T11:00:54Z): modelo real desconocido; `workflow.yml` da `claude-fable-5-1` a `modelo_adversario`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_a` (2026-10-16T11:00:54Z): modelo real desconocido; `workflow.yml` da `claude-opus-5-5` a `modelo_revisor`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_b` (2026-10-16T11:30:56Z): modelo real desconocido; `workflow.yml` da `claude-fable-5-1` a `modelo_adversario`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
  - `revision_juez_a` (2026-10-16T11:30:56Z): modelo real desconocido; `workflow.yml` da `claude-opus-5-5` a `modelo_revisor`. No se sabe con qué modelo corrió: la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo (un error de la API o el límite de uso).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 5 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 10 commits; 59 files changed, 4795 insertions(+), 147 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Alcance (lo que queda fuera o dentro del hito)**

- T005: el diff del hito añade `specs/021-h26-el-umbral-de/gates/modelos.jsonl`, una línea del paso `specify` del workflow (commit `92b434a`, de ninguna tarea), y la tarea pide constatar que no cambia ningún artefacto de otros hitos → no se retira y queda escrito en `cierre.md` (§1 y §6) para la persona y la revisión final: su ruta no está entre las declaradas de T005, el guardián solo admite el directorio del feature, la causa está en el workflow, que no se toca dentro de un hito, y no es el spec… (entera en `specs/022-h23-1-jurisprudencia-coteja-el/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_revision: [i] el diff del hito añadía `specs/021-h26-el-umbral-de/gates/modelos.jsonl`, un registro del directorio de H26 con la línea del paso `specify` de este run, que deja el de H23.1 incompleto → lectura conservadora: se retira del diff (el fichero no existe en `main`) y su línea pasa a ser la primera de `gates/modelos.jsonl` de H23.1; el supuesto [alcance] de T005 («no se retira») queda sustituido por este; la causa, que el paso `specify` escribe su registro antes de que exista e… (entera en `specs/022-h23-1-jurisprudencia-coteja-el/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q1 (criterio d, conservadora): El sondeo no cambia: FR-033 se queda tal cual. El hito no entrega nada para que `make evals-sondeo` cuente las sesiones sin cotejo ni para que mida el modo herramienta, y el run no toca `internal/evals/sondeo.go` ni `scripts/evals-sondeo.sh`. SC-010 sigue siendo una medida humana, fuera del run y sin control en él, y el spec no fija su procedimiento. El spec lo anota en «Fuera de alcance» y como pendiente para la persona en el informe final: hoy `make evals… (entera en `specs/022-h23-1-jurisprudencia-coteja-el/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run

- 2026-10-16T10:11:26Z · «docs(H23.1): barrido de artefactos y documentación contra el producto»: make ci quedó en rojo; el cambio se apartó en specs/022-h23-1-jurisprudencia-coteja-el/gates/aparcado-1.patch y el hito vuelve a daa6867.

- Observaciones de los jueces, por debajo del umbral y sin corregir: 8 en `spec-r1.json`, 7 en `plan-r1.json`, 7 en `tasks-r1.json`, 9 en `revision-a-r2.json`, 8 en `revision-b-r2.json`.

### Internos (1)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: barrido (1). Enteras en `specs/022-h23-1-jurisprudencia-coteja-el/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**boe-legislacion**: aprobado sobre `60e8678`; decide `claude-sonnet-5-5` con 2 de 3; 21 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `11-no-activa-programacion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `12-no-activa-acuerdo-entre-amigos.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `13-lrbrl-atribuciones-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `14-trlrhl-impuestos-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `15-irpf-rendimientos-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `16-lrjsp-legalidad-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `17-ltaibg-plazo-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `18-lrjpac-norma-derogada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `19-lcsp-contrato-menor-redaccion-cambiada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `20-lcsp-dos-bloques-redaccion-cambiada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `21-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 1 de 54 (1,9 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 3 de 54 (5,6 %) | ≤ 2,8 % | por encima | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 0 de 54 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 3 de 108 (2,8 %) | ≤ 2,8 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 565 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 580 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 101 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 113 | ≤ 900 | sí | sí |

**jurisprudencia**: aprobado sobre `60e8678`; decide `claude-sonnet-5-5` con 2 de 3; 10 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `02-resumen-sin-documento.yaml` | 3/3 | 1/3 | 3/3 | 3/3 |  |
| `03-por-materia.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `04-documento-pegado.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-documento-que-no-es-el-pedido.yaml` | 3/3 | 0/3 | 3/3 | 1/3 |  |
| `07-resumen-de-una-conocida.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `08-doctrina-con-el-fallo-delante.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `09-de-que-trata-con-la-ficha-sola.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `10-doctrina-dada-por-hecha.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 1,7 % | dentro | no: solo se publica, no es un control |
| `afirma_que_existe:claude-sonnet-5-5:orden` | 3 de 30 (10,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `cita_sin_documento:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `documento_sin_cotejar:claude-sonnet-5-5:orden` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 1,7 % | dentro | no: solo se publica, no es un control |
| `afirma_que_existe:claude-sonnet-5-5:herramienta` | 2 de 30 (6,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `cita_sin_documento:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `documento_sin_cotejar:claude-sonnet-5-5:herramienta` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 60 (0,0 %) | ≤ 1,7 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 125 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 124 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_del_juez:orden` | 40 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 41 | ≤ 900 | sí | sí |

**legal-core**: aprobado sobre `60e8678`; decide `claude-sonnet-5-5` con 2 de 3; 4 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 0/3 | — | — |  |

Umbrales: el job no publica ninguno para esta skill.

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

**Serie de cierres de boe-legislacion** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `eed129f` (árbol de `skills/boe-legislacion`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

`afirma_lo_no_leido`:

| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |
|---|---|---|---|---|---|---|---|---|---|
| H24-1 | 2026-10-06 | `a5bda45` | `6d65aa3` | 1/54 | 0/54 | 1/78 | 0/30 | 1/108 | no: otra skill |
| H24-2 | 2026-10-06 | `b0a7a9c` | `75e1d04` | 0/54 | 3/54 | 0/78 | 3/30 | 3/108 | no: 1 respuesta sin juzgar |
| H24-3 | 2026-10-06 | `3107ae2` | `6063317` | 0/54 | 0/54 | 0/78 | 0/30 | 0/108 | no: otra skill |
| H23-1 | 2026-10-07 | `bf4befd` | `6063317` | 0/54 | 0/50 | 0/76 | 0/28 | 0/104 | no: otra skill |
| H25-1 | 2026-10-10 | `3f99446` | `6063317` | 0/54 | 1/54 | 0/78 | 1/30 | 1/108 | no: otra skill |
| H25-2 | 2026-10-10 | `e380f0a` | `6063317` | 1/54 | 1/54 | 0/78 | 2/30 | 2/108 | no: otra skill |
| H25-3 | 2026-10-10 | `d51f3d7` | `eed129f` | 0/54 | 2/54 | 0/78 | 2/30 | 2/108 | sí |
| H26-1 | 2026-10-13 | `b29fe6c` | `eed129f` | 0/54 | 2/54 | 1/78 | 1/30 | 2/108 | sí |
| H23.1-1 | 2026-10-16 | `60e8678` | `eed129f` | 0/54 | 3/54 | 1/78 | 2/30 | 3/108 | sí |

Tasa acumulada con la skill como está: 7 de 324 (2,16 %), de H25-3, H26-1 y H23.1-1.
Insuficiente: 324 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

**Serie de cierres de jurisprudencia** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `4759aa4` (árbol de `skills/jurisprudencia`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

`afirma_lo_no_leido`:

| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |
|---|---|---|---|---|---|---|---|---|---|
| H25-1 | 2026-10-10 | `3f99446` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H25-2 | 2026-10-10 | `e380f0a` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: 1 respuesta sin juzgar |
| H25-3 | 2026-10-10 | `d51f3d7` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H26-1 | 2026-10-13 | `b29fe6c` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H23.1-1 | 2026-10-16 | `60e8678` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |

Tasa acumulada con la skill como está: 0 de 60 (0,00 %), de H23.1-1.
Insuficiente: 60 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

## 4. Revisión que la constitución reserva a la persona

Nada: el hito no toca fuentes, anomalías, adaptadores, fixtures, esquemas, grabaciones ni evidencias.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T001(hecha) T002(hecha) | — | comprobado por su control | `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 % |
| FR-002 | T001(hecha) | — | comprobado por su control | `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 % |
| FR-003 | T001(hecha) | — | comprobado por su control | `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 % |
| FR-004 | T001(hecha) | — | comprobado por su control | `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 % |
| FR-005 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| FR-006 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| FR-007 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme`, en make ci (verde); `ci:internal/evals/informe_test.go:TestInformeEnDosModos`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-020 | T003(hecha) | — | tareas hechas | — |
| FR-021 | T003(hecha) | — | tareas hechas | — |
| FR-022 | T003(hecha) | — | tareas hechas | — |
| FR-023 | T003(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde) |
| FR-024 | T003(hecha) | — | tareas hechas | — |
| FR-030 | T001(hecha) | — | comprobado por su control | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 40, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 41, ≤ 900; `ci:internal/evals/umbrales_test.go:TestUmbralesConJuezYConSentencias`, en make ci (verde); `ci:internal/evals/informe_test.go:TestInformeConElJuez`, en make ci (verde); `ci:internal/evals/juzgar_test.go:TestJuzgar`, en make ci (verde); `ci:internal/evals/informe_test.go:TestInforme`, en make ci (verde) |
| FR-031 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralesConJuezYConSentencias`, en make ci (verde); `ci:internal/evals/informe_test.go:TestInformeConElJuez`, en make ci (verde); `ci:internal/evals/juzgar_test.go:TestJuzgar`, en make ci (verde); `ci:internal/evals/informe_test.go:TestInforme`, en make ci (verde) |
| FR-032 | T001(hecha) T003(hecha) T005(hecha) | — | comprobado por su control | `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `ci:internal/evals/medida_test.go:TestCopiasDelJuez`, en make ci (verde); `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde); `ci:Makefile:schema-check`, en make ci (verde) |
| FR-033 | T001(hecha) T005(hecha) | — | comprobado por su control | `ci:internal/evals/sondeo_test.go:TestSalidaDelSondeo`, en make ci (verde) |
| FR-034 | T003(hecha) T005(hecha) | — | tareas hechas | — |
| FR-040 | T004(hecha) | — | tareas hechas | — |
| FR-041 | T004(hecha) | — | tareas hechas | — |
| FR-042 | T004(hecha) | — | tareas hechas | — |
| FR-043 | T004(hecha) | — | tareas hechas | — |
| FR-044 | T004(hecha) | — | tareas hechas | — |
| FR-045 | T004(hecha) T005(hecha) | — | tareas hechas | — |
| FR-050 | T005(hecha) | — | tareas hechas | — |
| FR-051 | T005(hecha) | — | tareas hechas | — |
| FR-052 | T005(hecha) | — | tareas hechas | — |
| FR-060 | T001(hecha) T005(hecha) | — | comprobado por su control | `ci:Makefile:ci`, en make ci (verde) |
| FR-061 | T005(hecha) | — | tareas hechas | — |
| FR-062 | T001(hecha) T002(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| FR-063 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde); `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme`, en make ci (verde); `ci:internal/evals/informe_test.go:TestInformeEnDosModos`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-064 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| FR-065 | T004(hecha) | — | tareas hechas | — |
| FR-066 | — | — | SIN TAREA | — |
| SC-001 | — | — | SIN TAREA | `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 40, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 41, ≤ 900 |
| SC-002 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| SC-003 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| SC-004 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| SC-005 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde); `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme`, en make ci (verde); `ci:internal/evals/informe_test.go:TestInformeEnDosModos`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| SC-006 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| SC-007 | T003(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde) |
| SC-008 | T004(hecha) | — | tareas hechas | — |
| SC-009 | T005(hecha) | — | comprobado por su control | `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 % |
| SC-010 | — | — | SIN TAREA | — |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`daa6867`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `86d4952` fix(H23.1): motivos de la revisión final: lo vio la ronda 2 (juez A: aprobado; juez B: aprobado). Toca `internal/evals/informe_test.go`, `internal/evals/sentencias_test.go`, `specs/021-h26-el-umbral-de/gates/modelos.jsonl`, `cierre.md`, `contracts/skill-jurisprudencia.md`, `contracts/umbral-documento-sin-cotejar.md`, `mutantes.md`, `plan.md`, `quickstart.md`, `research.md`, `spec.md`.
- `9a85df8` docs(H23.1): veredictos de la revisión final: solo registros de `gates/`.
- `60e8678` docs(H23.1): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/022-h23-1-jurisprudencia-coteja-el/quickstart.md`. Suite de aceptación congelada: `specs/022-h23-1-jurisprudencia-coteja-el/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `34309f46`: 6 h 14 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh 34309f46`.
