# Informe del hito H25 · El juez de `jurisprudencia`: resumir o caracterizar una sentencia que no se ha leído decide (ADR 0037)

Generado por el workflow `hito` el 2026-10-10T21:32:42Z, sobre `d51f3d7` de `020-h25-el-juez-de`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/020-h25-el-juez-de/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige la sección del hito en docs/ROADMAP.md y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `d51f3d7` (medición 3): ROJO: evals (boe-legislacion); es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: boe-legislacion fallo, jurisprudencia aprobado, legal-core aprobado (tasas en la sección 3).
- **Umbrales del job**: boe-legislacion: 12 umbrales, **3 sin cumplir o solo publicados** (`cuenta_su_proceso:claude-sonnet-5-5:orden`, `afirma_lo_no_leido:claude-sonnet-5-5:herramienta`, `cuenta_su_proceso:claude-sonnet-5-5:herramienta`); jurisprudencia: 12 umbrales, **2 sin cumplir o solo publicados** (`afirma_que_existe:claude-sonnet-5-5:orden`, `afirma_que_existe:claude-sonnet-5-5:herramienta`); legal-core: sin umbrales (sección 3).
- **Revisión final**: juez A aprobado, juez B aprobado, 4 rondas en 2 ciclos (el primero juzga el hito; cada uno de los siguientes, lo que cambió después del último veredicto).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 9 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 20 commits; 94 files changed, 38656 insertions(+), 969 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Alcance (lo que queda fuera o dentro del hito)**

- T007: la tarea pide registrar en `CHANGELOG.md` (*Unreleased*) `jurisprudencia` v0.1 y lo del job, y no dice qué hacer con las entradas de H23 de esa misma sección, que afirman lo que H25 deja falso —«`jurisprudencia` no tiene juez», seis evals de una en una, 72 sesiones, cuatro umbrales, 20 341 s, `SKILL.md` de 195 líneas— → no se añade una entrada aparte que las contradiga: las dos entradas de H23 pasan a decir lo que llevará la release —la de la skill, como v0.1, y la del job, con el juez—, y… (entera en `specs/020-h25-el-juez-de/gates/supuestos.md` o `clarify-respuestas.json`)
- reparar_cierre: la medición 2 de la cabeza `e380f0a` sale en rojo en `evals (boe-legislacion)` por `afirma_lo_no_leido`, 1 de 54 en cada modo frente al umbral 0, con dos respuestas de la eval 15 que ponen la materia del art. 7 de la LIRPF con «sobre» en el aviso final de las remisiones no seguidas («el art. 7 sobre rentas exentas»), sin haberlo leído y con un índice que no da rúbricas (`gates/evals/boe-legislacion.json`) → `boe-legislacion` v0.1.8: la viñeta «De un precepto que no has leído, nad… (entera en `specs/020-h25-el-juez-de/gates/supuestos.md` o `clarify-respuestas.json`)
- reparar_cierre: la misma medición sale en rojo en `evals (jurisprudencia)` con sus doce umbrales cumplidos, por una respuesta sin juzgar: un voto del juez agotó su tope de 35 s, lo que H24 (research D6 y D28, contracts/juez-y-voto.md §5) dejó sin reintento y reservó a una persona, y que ya puso en rojo la medición 2 del cierre de H24 → el voto que el tope corta se pide otra vez, una sola, con su mismo número, como el nulo: dos peticiones por voto como mucho, así que el peor caso, los dos topes y… (entera en `specs/020-h25-el-juez-de/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_revision: la reparación R1 reescribió en una línea la última viñeta de «Redacción modificada» de `skills/boe-legislacion/SKILL.md`, un pasaje que la medición no toca, para pagar la línea que gana la otra viñeta (juez A, [i]) → lectura conservadora: se retira. Esa viñeta vuelve a ser, byte a byte, la de `main`, y `SKILL.md` queda en 299 líneas, su máximo, con un solo trozo de diff con `main` (+5 −4). Con 299, `make ci` salía en rojo en `TestSkillsDelRepositorio/region/boe-legislacion/do… (entera en `specs/020-h25-el-juez-de/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_revision: el registro de R1 negaba que se apartara de la sección del hito y del spec, y el de R2 llevaba impacto interno aunque decide lo que H24 reservó a una persona (juez A, [f][i]) → research.md («Reparaciones del cierre»: preámbulo, R1 y R2), plan.md («Summary» y «Decisiones») y las dos líneas `reparar_cierre` de arriba dicen de qué se aparta cada una —R1, de «Fuera de alcance» de la sección del hito y de FR-005; R2, de lo que H24 dejó sin reintento y a la decisión de una persona—… (entera en `specs/020-h25-el-juez-de/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run

- 2026-10-10T19:50:42Z · La revisión final abre el ciclo 2 para juzgar lo que cambió fuera de gates/ después del veredicto sobre 4a127c5: 0dc549c «fix(H25): cierre en la plataforma».

- Observaciones de los jueces, por debajo del umbral y sin corregir: 6 en `spec-r1.json`, 10 en `plan-r1.json`, 11 en `tasks-r1.json`, 10 en `revision-a-r4.json`, 7 en `revision-b-r4.json`.

### Internos (36)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T002 (10), T003 (5), T004 (6), T005 (6), T007 (1), T008 (2), T009 (1), barrido (3), corrector_revision (1), reparar_cierre (1). Enteras en `specs/020-h25-el-juez-de/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**boe-legislacion**: fallo sobre `d51f3d7`; decide `claude-sonnet-5-5` con 2 de 3; 21 evals, 0 nuevas, 8 informativas.

- Motivo del fallo: umbral afirma_lo_no_leido:claude-sonnet-5-5:herramienta: 2 de 54 (3,7 %), y tiene que ser ≤ 0,0 %: 15-irpf-rendimientos-por-materia-herramienta-claude-sonnet-5-5-03: «el texto añade "sin perjuicio de lo dispuesto en el artículo 7", que regula las rentas exentas.» · «el texto añade "sin perjuicio de lo dispuesto en el artículo 7", que regula las rentas exentas.» · «el texto añade "sin perjuicio de lo dispuesto en el artículo 7", que regula las rentas exentas.»; 16-lrjsp-legalidad-por-materia-herramienta-claude-sonnet-5-5-03: «No he leído los demás artículos del Capítulo (por ejemplo, el art. 27 sobre tipicidad), así que no puedo decir qué regulan.» · «No he leído los demás artículos del Capítulo (por ejemplo, el art. 27 sobre tipicidad), así que no puedo decir qué regulan.» · «No he leído los demás artículos del Capítulo (por ejemplo, el art. 27 sobre tipicidad), así que no puedo decir qué regulan.»

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
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
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 2 de 54 (3,7 %) | ≤ 0,0 % | ✗ **NO** | no: solo se publica, no es un control |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 2 de 54 (3,7 %) | ≤ 0,0 % | ✗ **NO** | sí |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | no: solo se publica, no es un control |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 487 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 459 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 95 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 112 | ≤ 900 | sí | sí |

**jurisprudencia**: aprobado sobre `d51f3d7`; decide `claude-sonnet-5-5` con 2 de 3; 10 evals, 4 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `07-resumen-de-una-conocida.yaml` | 3/3 | 2/3 | 3/3 | 3/3 | **nueva** |
| `08-doctrina-con-el-fallo-delante.yaml` | 3/3 | 3/3 | 3/3 | 3/3 | **nueva** |
| `09-de-que-trata-con-la-ficha-sola.yaml` | 3/3 | 3/3 | 3/3 | 3/3 | **nueva** |
| `10-doctrina-dada-por-hecha.yaml` | 3/3 | 1/3 | 3/3 | 1/3 | **nueva** |
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `02-resumen-sin-documento.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `03-por-materia.yaml` | 3/3 | 0/3 | 3/3 | 3/3 |  |
| `04-documento-pegado.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-documento-que-no-es-el-pedido.yaml` | 2/3 | 0/3 | 3/3 | 0/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_que_existe:claude-sonnet-5-5:orden` | 1 de 30 (3,3 %) | ≤ 0,0 % | ✗ **NO** | no: solo se publica, no es un control |
| `cita_sin_documento:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_que_existe:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | no: solo se publica, no es un control |
| `cita_sin_documento:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 125 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 124 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_del_juez:orden` | 36 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 35 | ≤ 900 | sí | sí |

**legal-core**: aprobado sobre `d51f3d7`; decide `claude-sonnet-5-5` con 2 de 3; 4 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
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
| FR-001 | T004(hecha) T008(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestCopiasDelJuez`, en make ci (verde) |
| FR-002 | T004(hecha) | — | tareas hechas | — |
| FR-003 | T004(hecha) | — | tareas hechas | — |
| FR-004 | T004(hecha) | — | tareas hechas | — |
| FR-005 | T001(hecha) T002(hecha) T003(hecha) T004(hecha) T005(hecha) T008(hecha) T009(hecha) | — | tareas hechas | — |
| FR-010 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/sesion_test.go:TestTextosDeLaSesion`, en make ci (verde); `ci:internal/evals/juez_test.go:TestMensajeDelVoto`, en make ci (verde) |
| FR-011 | T001(hecha) T004(hecha) T008(hecha) | — | tareas hechas | — |
| FR-012 | T004(hecha) | — | tareas hechas | — |
| FR-013 | T001(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/juez_test.go:TestVotoDelJuez`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| FR-020 | T004(hecha) T008(hecha) | — | comprobado por su control | `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 % |
| FR-021 | T004(hecha) | — | tareas hechas | — |
| FR-022 | T004(hecha) T008(hecha) | — | comprobado por su control | `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 % |
| FR-023 | T004(hecha) T008(hecha) | — | comprobado por su control | `evals:jurisprudencia:duracion_del_juez:orden`: 36, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 35, ≤ 900 |
| FR-024 | T003(hecha) T004(hecha) T007(hecha) T009(hecha) | — | comprobado por su control | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 % |
| FR-025 | T004(hecha) | — | tareas hechas | — |
| FR-026 | T005(hecha) T008(hecha) T009(hecha) | — | tareas hechas | — |
| FR-027 | T008(hecha) | — | tareas hechas | — |
| FR-030 | T004(hecha) T008(hecha) | — | tareas hechas | — |
| FR-031 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestMedidaVersionada`, en make ci (verde); `ci:internal/evals/ejecucion_test.go:TestEjecucionSinMedir`, en make ci (verde) |
| FR-032 | T004(hecha) | — | tareas hechas | — |
| FR-033 | T004(hecha) T008(hecha) | — | tareas hechas | — |
| FR-040 | T004(hecha) T005(hecha) | — | tareas hechas | — |
| FR-041 | T002(hecha) T005(hecha) | — | tareas hechas | — |
| FR-042 | T002(hecha) T005(hecha) | — | tareas hechas | — |
| FR-043 | T002(hecha) T005(hecha) | — | tareas hechas | — |
| FR-044 | T002(hecha) T005(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestReconstruccionDeJurisprudencia`, en make ci (verde) |
| FR-045 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestGrabacionesDerivadas`, en make ci (verde) |
| FR-050 | T004(hecha) | — | tareas hechas | — |
| FR-051 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestEjecucionDeLaMedida`, en make ci (verde) |
| FR-060 | T003(hecha) T009(hecha) | — | tareas hechas | — |
| FR-061 | T003(hecha) | — | tareas hechas | — |
| FR-062 | T003(hecha) | — | tareas hechas | — |
| FR-063 | T003(hecha) | — | tareas hechas | — |
| FR-064 | T003(hecha) | — | tareas hechas | — |
| FR-065 | T003(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestPreguntasDelSondeo`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestPreguntasConElFragmento`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestConjuntoDeEvals`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-066 | T003(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestPreguntasDelSondeo`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestPreguntasConElFragmento`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestConjuntoDeEvals`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-070 | T003(hecha) T007(hecha) | — | tareas hechas | — |
| FR-071 | T004(hecha) | — | tareas hechas | — |
| FR-072 | T003(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-073 | T008(hecha) | — | tareas hechas | — |
| FR-080 | T006(hecha) | — | tareas hechas | — |
| FR-081 | T006(hecha) | — | tareas hechas | — |
| FR-082 | T006(hecha) | — | tareas hechas | — |
| FR-083 | T006(hecha) T008(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde) |
| FR-084 | T007(hecha) | — | tareas hechas | — |
| FR-090 | T007(hecha) | — | tareas hechas | — |
| FR-091 | T007(hecha) | — | tareas hechas | — |
| FR-092 | T007(hecha) | — | tareas hechas | — |
| FR-093 | T007(hecha) | — | tareas hechas | — |
| FR-095 | T008(hecha) | — | tareas hechas | — |
| FR-096 | T007(hecha) T008(hecha) | — | tareas hechas | — |
| FR-100 | T008(hecha) T009(hecha) | — | tareas hechas | — |
| FR-101 | T008(hecha) | — | tareas hechas | — |
| FR-102 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestCopiasDelJuez`, en make ci (verde) |
| FR-103 | T003(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-104 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestMedidaVersionada`, en make ci (verde); `ci:internal/evals/ejecucion_test.go:TestEjecucionSinMedir`, en make ci (verde) |
| FR-105 | T002(hecha) T005(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestReconstruccionDeJurisprudencia`, en make ci (verde) |
| FR-106 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestGrabacionesDerivadas`, en make ci (verde) |
| FR-107 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestEjecucionDeLaMedida`, en make ci (verde) |
| FR-108 | T001(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/juez_test.go:TestVotoDelJuez`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| FR-109 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/sesion_test.go:TestTextosDeLaSesion`, en make ci (verde); `ci:internal/evals/juez_test.go:TestMensajeDelVoto`, en make ci (verde) |
| FR-110 | T003(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestPreguntasDelSondeo`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestPreguntasConElFragmento`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestConjuntoDeEvals`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-111 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| FR-112 | T007(hecha) | — | tareas hechas | — |
| SC-001 | T004(hecha) | — | comprobado por su control | `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 36, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 35, ≤ 900; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `ci:internal/evals/informe_test.go:TestInformeConElJuez`, en make ci (verde) |
| SC-002 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestCopiasDelJuez`, en make ci (verde) |
| SC-003 | T003(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| SC-004 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestMedidaVersionada`, en make ci (verde); `ci:internal/evals/ejecucion_test.go:TestEjecucionSinMedir`, en make ci (verde) |
| SC-005 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestReconstruccionDeJurisprudencia`, en make ci (verde) |
| SC-006 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestGrabacionesDerivadas`, en make ci (verde) |
| SC-007 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestEjecucionDeLaMedida`, en make ci (verde) |
| SC-008 | T001(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/juez_test.go:TestVotoDelJuez`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| SC-009 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/sesion_test.go:TestTextosDeLaSesion`, en make ci (verde); `ci:internal/evals/juez_test.go:TestMensajeDelVoto`, en make ci (verde) |
| SC-010 | T003(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestPreguntasDelSondeo`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestPreguntasConElFragmento`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestConjuntoDeEvals`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| SC-011 | T006(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde) |
| SC-012 | T007(hecha) T008(hecha) T009(hecha) | — | tareas hechas | — |
| SC-013 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| SC-014 | T006(hecha) | — | tareas hechas | — |
| SC-015 | T007(hecha) | — | tareas hechas | — |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`062c399`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `4a127c5` fix(H25): motivos de la revisión final: lo vio la ronda 2 (juez A: aprobado; juez B: aprobado). Toca `internal/evals/medida.go`, `internal/evals/medida_test.go`, `cierre.md`.
- `26e8122` docs(H25): veredictos de la revisión final: solo registros de `gates/`.
- `3f99446` docs(H25): registros del run: solo registros de `gates/`.
- `e380f0a` docs(H25): registros del run: solo registros de `gates/`.
- `0dc549c` fix(H25): cierre en la plataforma: lo vio la ronda 3 (juez A: rechazado; juez B: rechazado). Toca `CHANGELOG.md`, `CONTRIBUTING.md`, `internal/evals/doc.go`, `internal/evals/ejecucion_test.go`, `internal/evals/informe.go`, `internal/evals/informe_test.go`, `internal/evals/juez.go`, `internal/evals/juez_test.go`, `internal/evals/medida_test.go`, `internal/evals/sondeo_test.go`, `skills/boe-legislacion/SKILL.md`, `plan.md`, `research.md`.
- `660391c` fix(H25): motivos de la revisión final: lo vio la ronda 4 (juez A: aprobado; juez B: aprobado). Toca `CHANGELOG.md`, `internal/app/skills_test.go`, `skills/boe-legislacion/SKILL.md`, `cierre.md`, `contracts/juez-de-jurisprudencia.md`, `data-model.md`, `plan.md`, `research.md`.
- `8c3578d` docs(H25): veredictos de la revisión final: solo registros de `gates/`.
- `d51f3d7` docs(H25): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/020-h25-el-juez-de/quickstart.md`. Suite de aceptación congelada: `specs/020-h25-el-juez-de/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `d0eb1f35`: 11 h 21 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh d0eb1f35`.
