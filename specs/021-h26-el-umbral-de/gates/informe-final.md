# Informe del hito H26 · El umbral de las clases del juez pasa a ser una tasa sobre los dos modos (ADR 0038)

Generado por el workflow `hito` el 2026-10-13T12:58:55Z, sobre `b29fe6c` de `021-h26-el-umbral-de`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/021-h26-el-umbral-de/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige la sección del hito en docs/ROADMAP.md y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `b29fe6c` (medición 1): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: boe-legislacion aprobado, jurisprudencia aprobado, legal-core aprobado (tasas en la sección 3).
- **Umbrales del job**: boe-legislacion: 13 umbrales, **4 sin cumplir o solo publicados** (`afirma_lo_no_leido:claude-sonnet-5-5:orden`, `cuenta_su_proceso:claude-sonnet-5-5:orden`, `afirma_lo_no_leido:claude-sonnet-5-5:herramienta`, `cuenta_su_proceso:claude-sonnet-5-5:herramienta`); jurisprudencia: 13 umbrales, **4 sin cumplir o solo publicados** (`afirma_lo_no_leido:claude-sonnet-5-5:orden`, `afirma_que_existe:claude-sonnet-5-5:orden`, `afirma_lo_no_leido:claude-sonnet-5-5:herramienta`, `afirma_que_existe:claude-sonnet-5-5:herramienta`); legal-core: sin umbrales (sección 3).
- **Revisión final**: juez A aprobado, juez B aprobado, 1 ronda.
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 5 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 9 commits; 55 files changed, 4854 insertions(+), 313 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Alcance (lo que queda fuera o dentro del hito)**

- clarify Q2 (criterio d, conservadora): Opción A. No se toca. El comentario de `.github/workflows/evals.yml` (hoy en sus líneas 152 a 154) sigue como está, el diff del hito no lleva nada bajo `.github/`, FR-055 y FR-062 no ganan ninguna excepción y SC-013 se queda en los cinco documentos de la Entrega 4. Va a «Fuera de alcance» como pendiente para la persona, dicho con sus datos: tras el hito, ese comentario dice que el informe de `jurisprudencia` lleva «doce umbrales, diez que deciden» y que `af… (entera en `specs/021-h26-el-umbral-de/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run


- Observaciones de los jueces, por debajo del umbral y sin corregir: 8 en `spec-r1.json`, 7 en `plan-r1.json`, 12 en `tasks-r1.json`, 5 en `revision-a-r1.json`, 7 en `revision-b-r1.json`.

### Internos (16)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T001 (4), T002 (5), T003 (2), T004 (3), T005 (2). Enteras en `specs/021-h26-el-umbral-de/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**boe-legislacion**: aprobado sobre `b29fe6c`; decide `claude-sonnet-5-5` con 2 de 3; 21 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 1/3 | 3/3 | 3/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
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
| `21-sin-binario-ni-servidor.yaml` | 3/3 | 2/3 | — | — |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 2,8 % | sí | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 1 de 54 (1,9 %) | ≤ 0,0 % | ✗ **NO** | no: solo se publica, no es un control |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 2 de 54 (3,7 %) | ≤ 2,8 % | ✗ **NO** | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 2 de 54 (3,7 %) | ≤ 0,0 % | ✗ **NO** | no: solo se publica, no es un control |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 2 de 108 (1,9 %) | ≤ 2,8 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 522 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 511 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 99 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 112 | ≤ 900 | sí | sí |

**jurisprudencia**: aprobado sobre `b29fe6c`; decide `claude-sonnet-5-5` con 2 de 3; 10 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `02-resumen-sin-documento.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
| `03-por-materia.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `04-documento-pegado.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-documento-que-no-es-el-pedido.yaml` | 2/3 | 0/3 | 3/3 | 0/3 |  |
| `07-resumen-de-una-conocida.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `08-doctrina-con-el-fallo-delante.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `09-de-que-trata-con-la-ficha-sola.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `10-doctrina-dada-por-hecha.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 1,7 % | sí | no: solo se publica, no es un control |
| `afirma_que_existe:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | no: solo se publica, no es un control |
| `cita_sin_documento:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 1,7 % | sí | no: solo se publica, no es un control |
| `afirma_que_existe:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | no: solo se publica, no es un control |
| `cita_sin_documento:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 60 (0,0 %) | ≤ 1,7 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 125 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 124 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_del_juez:orden` | 39 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 38 | ≤ 900 | sí | sí |

**legal-core**: aprobado sobre `b29fe6c`; decide `claude-sonnet-5-5` con 2 de 3; 4 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 0/3 | — | — |  |

Umbrales: el job no publica ninguno para esta skill.

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

## 4. Revisión que la constitución reserva a la persona

Nada: el hito no toca fuentes, anomalías, adaptadores, fixtures, esquemas, grabaciones ni evidencias.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T001(hecha) T002(hecha) T003(hecha) | — | comprobado por su control | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5`: 2 de 108 (1,9 %), ≤ 2,8 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 % |
| FR-002 | T001(hecha) T002(hecha) | — | tareas hechas | — |
| FR-003 | T001(hecha) T002(hecha) | — | comprobado por su control | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5`: 2 de 108 (1,9 %), ≤ 2,8 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 % |
| FR-004 | T001(hecha) T002(hecha) | — | comprobado por su control | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5`: 2 de 108 (1,9 %), ≤ 2,8 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 % |
| FR-005 | T001(hecha) | — | tareas hechas | — |
| FR-010 | T001(hecha) T002(hecha) | — | tareas hechas | — |
| FR-011 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConElJuez`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| FR-012 | T001(hecha) T002(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde); `ci:internal/evals/informe_test.go:TestInformeEnDosModos`, en make ci (verde) |
| FR-020 | T002(hecha) T005(hecha) | — | tareas hechas | — |
| FR-021 | T001(hecha) T002(hecha) T005(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestCopiasDelJuez`, en make ci (verde); `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde); `ci:Makefile:schema-check`, en make ci (verde) |
| FR-022 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-030 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConElJuez`, en make ci (verde) |
| FR-031 | T001(hecha) | — | comprobado por su control | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 % |
| FR-032 | T001(hecha) | — | comprobado por su control | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 % |
| FR-033 | T001(hecha) | — | comprobado por su control | `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `ci:internal/evals/medida_test.go:TestMedidaVersionada`, en make ci (verde); `ci:internal/evals/ejecucion_test.go:TestEjecucionSinMedir`, en make ci (verde) |
| FR-034 | T001(hecha) | — | comprobado por su control | `evals:boe-legislacion:duracion_de_las_sesiones:orden`: 522, ≤ 900; `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`: 511, ≤ 900; `evals:boe-legislacion:duracion_del_juez:orden`: 99, ≤ 900; `evals:boe-legislacion:duracion_del_juez:herramienta`: 112, ≤ 900; `evals:jurisprudencia:duracion_del_juez:orden`: 39, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 38, ≤ 900 |
| FR-035 | T001(hecha) T002(hecha) T005(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestCopiasDelJuez`, en make ci (verde); `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde); `ci:Makefile:schema-check`, en make ci (verde) |
| FR-036 | T005(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde) |
| FR-040 | T001(hecha) T002(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/sondeo_test.go:TestSalidaDelSondeo`, en make ci (verde) |
| FR-050 | T004(hecha) | — | tareas hechas | — |
| FR-051 | T004(hecha) | — | tareas hechas | — |
| FR-052 | T004(hecha) | — | tareas hechas | — |
| FR-053 | T004(hecha) | — | tareas hechas | — |
| FR-054 | T004(hecha) | — | tareas hechas | — |
| FR-055 | T004(hecha) T005(hecha) | — | tareas hechas | — |
| FR-060 | T005(hecha) | — | tareas hechas | — |
| FR-061 | T002(hecha) T005(hecha) | — | tareas hechas | — |
| FR-062 | T005(hecha) | — | tareas hechas | — |
| FR-070 | T001(hecha) T002(hecha) T005(hecha) | — | comprobado por su control | `ci:Makefile:ci`, en make ci (verde) |
| FR-071 | T005(hecha) | — | tareas hechas | — |
| FR-072 | T002(hecha) T003(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralConjunto`, en make ci (verde) |
| FR-073 | T002(hecha) T003(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| FR-074 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralConjunto`, en make ci (verde) |
| FR-075 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConElJuez`, en make ci (verde); `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde); `ci:internal/evals/informe_test.go:TestInformeEnDosModos`, en make ci (verde) |
| FR-076 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConElJuez`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| FR-077 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-078 | T003(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralConjunto`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| FR-079 | T004(hecha) | — | tareas hechas | — |
| SC-001 | — | — | SIN TAREA | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5`: 2 de 108 (1,9 %), ≤ 2,8 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:duracion_de_las_sesiones:orden`: 522, ≤ 900; `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`: 511, ≤ 900; `evals:boe-legislacion:duracion_del_juez:orden`: 99, ≤ 900; `evals:boe-legislacion:duracion_del_juez:herramienta`: 112, ≤ 900; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 39, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 38, ≤ 900 |
| SC-002 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralConjunto`, en make ci (verde) |
| SC-003 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| SC-004 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralConjunto`, en make ci (verde) |
| SC-005 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConElJuez`, en make ci (verde) |
| SC-006 | T001(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConElJuez`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| SC-007 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| SC-008 | T003(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralConjunto`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| SC-009 | T001(hecha) T002(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde); `ci:internal/evals/informe_test.go:TestInformeEnDosModos`, en make ci (verde) |
| SC-010 | T004(hecha) T005(hecha) | — | comprobado por su control | `ci:Makefile:ci`, en make ci (verde) |
| SC-011 | T005(hecha) | — | comprobado por su control | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5`: 2 de 108 (1,9 %), ≤ 2,8 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 % |
| SC-012 | — | — | SIN TAREA | — |
| SC-013 | T004(hecha) | — | tareas hechas | — |
| SC-014 | T001(hecha) T002(hecha) | — | comprobado por su control | `ci:internal/evals/sondeo_test.go:TestSalidaDelSondeo`, en make ci (verde) |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`314f78c`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `b29fe6c` docs(H26): veredictos de la revisión final: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/021-h26-el-umbral-de/quickstart.md`. Suite de aceptación congelada: `specs/021-h26-el-umbral-de/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `3a325961`: 5 h 18 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh 3a325961`.
