# Informe del hito H38 · Las skills instaladas se actualizan con el binario (ADR 0060)

Generado por el workflow `hito` el 2026-11-21T19:27:57Z, sobre `c392ef1b` de `040-h38-las-skills-instaladas`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/040-h38-las-skills-instaladas/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige el fichero del hito en docs/roadmap/in-progress/ y se relanza.

## 1. Estado

- **make ci local**: ROJO (exit 2; gates/ci.log).
- **CI y evals remotos** sobre `c392ef1` (medición 2): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: consulta-juridica aprobado, jurisprudencia aprobado, normativa aprobado, plazos-y-territorio aprobado (tasas en la sección 3).
- **Umbrales del job**: consulta-juridica: 8 umbrales, los 4 que deciden se cumplen y 4 solo se publican (4 por encima de su umbral: `fuera_de_lo_grabado:claude-sonnet-5-5:orden`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); jurisprudencia: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (2 por encima de su umbral: `afirma_que_existe:claude-sonnet-5-5:orden`, `afirma_que_existe:claude-sonnet-5-5:herramienta`); normativa: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (7 por encima de su umbral: `afirma_lo_no_leido:claude-sonnet-5-5:orden`, `cuenta_su_proceso:claude-sonnet-5-5:orden`, `cuenta_su_proceso:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-sonnet-5-5:orden`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); plazos-y-territorio: 18 umbrales, los 12 que deciden se cumplen y 6 solo se publican, todos dentro de su umbral (sección 3).
- **Revisión final**: juez A aprobado (con `claude-opus-5-5`), juez B aprobado (con `claude-fable-5-2`), 4 rondas en 2 ciclos (el primero juzga el hito; cada uno de los siguientes, lo que cambió después del último veredicto, o el hito entero si es un rejuicio).
- **Sesiones con otro modelo que el de su rol**: ninguna (33 sesiones registradas, todas con el modelo que `workflow.yml` da a su rol).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 7 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 19 commits; 102 files changed, 57085 insertions(+), 145 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- corrector_spec: [h][f] con `--dry-run` o con un error sin `--json` la salida de error ya lleva el texto del kernel, y US3.5 y SC-003 pedían exactamente una línea → nueva definición «Una línea de aviso más»: la salida de error es la de la misma orden sin skills instaladas con exactamente una línea de aviso más (FR-006, FR-009 o la de hoy) en el sitio de la de hoy, o ninguna más; `\A…\n\z` solo donde esa orden deja vacía la salida de error. US3.5, SC-003 y la nota de los controles del hito remiten… (entera en `specs/040-h38-las-skills-instaladas/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_spec: [h][f] la propia consulta escribe el grafo del mundo en el directorio de `cache.Directorio()`, y varios escenarios y SC-006 decían «nada escrito» o «0 entradas en la caché» → FR-013 nombra lo que escribe la reinstalación (el árbol del global, el candado y el directorio de la caché que crea para tomarlo) y define «nada escrito»: el árbol sin cambios y la caché como la deja la misma orden sin skills instaladas, sin candado ni un directorio que esa orden no crearía. FR-002, FR-003, … (entera en `specs/040-h38-las-skills-instaladas/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_spec: [j][f] la cláusula de FR-010 «de dos o más invocaciones a la vez, como mucho una reinstala» no tenía control → nuevo escenario US4.6, y SC-004 añade una comprobación de `make ci` que sale en rojo si, de dos invocaciones lanzadas a la vez sobre la misma instalación global v1, las dos dan la línea de FR-006. Dónde vive esa comprobación y cómo fuerza la concurrencia lo decide el plan. El umbral no se rebaja.
- plan: FR-010 fija un candado por fecha de modificación, y comprobar la fecha y retirar el huérfano no son una sola operación del sistema de ficheros → se retira y se crea en exclusiva una vez (research.md D5); dos invocaciones que examinan el mismo huérfano en el mismo instante pueden tomarlo las dos y aplicar el mismo plan, con el árbol final de una sola reinstalación y dos líneas de FR-006. Descartado: un cierre del sistema (`flock` no existe en Windows en `syscall`) y una ficha de sustitución… (entera en `specs/040-h38-las-skills-instaladas/gates/supuestos.md` o `clarify-respuestas.json`)
- T002: el pedido de data-model §4 puede salir vacío (un global cuyas skills declaradas no empotra el binario ni son nombres retirados), y entonces `Planificar` daría un plan que solo reescribe la versión del manifiesto, mientras que `install -g` sin nombres instalaría todas las empotradas → con el pedido vacío no se reinstala y se da la línea de hoy, sin candado y sin escribir nada: es la lectura que menos hace, no calla el aviso sin haber puesto nada al día y no instala ninguna skill que el glob… (entera en `specs/040-h38-las-skills-instaladas/gates/supuestos.md` o `clarify-respuestas.json`)
- T003: si el directorio de la caché no existía, `TomarCandado` lo crea y después la creación del candado falla por otra causa que «ya existe», contracts/aviso-y-reinstalacion.md §4 dice «se crea con 0o700 y se sigue — queda el directorio» y «lo que hubiera», y FR-013 pide no dejar «un directorio que esa orden no crearía» → el directorio queda: es la lectura que menos hace (sin una retirada que habría que acotar a lo que `MkdirAll` creó, padres incluidos) y no puede romper a otra invocación que en… (entera en `specs/040-h38-las-skills-instaladas/gates/supuestos.md` o `clarify-respuestas.json`)
- T006: tasks.md pide que, tras el fallo en cada operación n, `install -g` llegue al estado final de la reinstalación sin fallos, pero si la n cae a mitad de la retirada de una skill de nombre retirado (ya se retiró una parte y no todo), lo que queda de ella ya no tiene «exactamente los ficheros que el manifiesto le declara», y FR-026 y FR-027 de H32, decisión cerrada, mandan que la instalación no la toque y la conserve con su aviso; FR-011 solo pide un estado que `install -g` «completa sin confli… (entera en `specs/040-h38-las-skills-instaladas/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_revision: [k] la rama del pedido vacío de `evaluar` (supuesto T002) no la pide ningún FR —FR-003 reinstala «si y solo si» (a) a (d)— y solo sirve a un global cuyas skills declaradas el binario ni empotra ni hereda, que ningún binario publicado produce → se retira por proporcionalidad, con su caso «ninguna skill que pedir» de `TestReinstalacion` y su descripción en contracts/aviso-y-reinstalacion.md §2 (paso 5) y data-model §4 y §6; el supuesto T002 queda retirado.

**Alcance (lo que queda fuera o dentro del hito)**

- reparar_cierre: la medición 1 del cierre sobre `3473dd4d` (ejecución 44928879479, trabajo 136921530208) sale en rojo solo por `evals (normativa)`, con un único motivo: `afirma_lo_no_leido:claude-sonnet-5-5`, 4 de 120 (3,3 %; el tope es 2,8 %), 1 de 60 en orden y 3 de 60 en herramienta, con `skills/` sin tocar por el hito y los demás umbrales que deciden cumplidos. Dos respuestas de la eval 17 (orden-01 y herramienta-02) leen solo el art. 20 de la Ley 19/2013 y dicen ante quién se presenta la rec… (entera en `specs/040-h38-las-skills-instaladas/gates/supuestos.md` o `clarify-respuestas.json`)

**Skill (lo que pide, dice o comprueba una skill)**

- corrector_revision: [i] la reparación del cierre cambia cómo se cumple FR-044 de H9 (`specs/023-h9-internal-core-plazos/spec.md`) —en la reclamación del art. 24.2 LTAIBG, señalar como variación autonómica el órgano ante el que se reclama— y no lo registraba: las dos líneas retiradas de `plazos-y-territorio/references/administrativo.md` eran su implementación → se registra, sin tocar el texto de la skill, que es la lectura conservadora (ningún contenido legal sin su lectura) y la que pide la medi… (entera en `specs/040-h38-las-skills-instaladas/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run

- 2026-11-21T18:11:43Z · La revisión final abre el ciclo 2 para juzgar lo que cambió fuera de gates/ después del veredicto sobre be6e748: 886d808 «fix(H38): cierre en la plataforma».
- 2026-11-21T18:43:43Z · «fix(H38): motivos de la revisión final»: make ci sigue en rojo; lo verán el paso siguiente y el informe final.

- Observaciones de los jueces, por debajo del umbral y sin corregir: 7 en `spec-r2.json`, 9 en `plan-r2.json`, 7 en `tasks-r1.json`, 5 en `revision-a-r4.json`, 7 en `revision-b-r4.json`.

### Internos (9)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T001 (3), T003 (1), corrector_plan (2), corrector_revision (1), plan (1), tasks (1). Enteras en `specs/040-h38-las-skills-instaladas/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**consulta-juridica**: aprobado sobre `c392ef1`; decide `claude-sonnet-5-5` con 2 de 3; 4 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-poliza-tras-el-fallecimiento.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `02-una-sola-capacidad.yaml` | 3/3 | 0/3 | 3/3 | 1/3 |  |
| `03-seguro-prorroga.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `04-despido-plazo.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cobertura_sin_declarar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cobertura_sin_declarar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 19 de 12 (158,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 16 de 12 (133,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 18 de 12 (150,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 18 de 12 (150,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

**jurisprudencia**: aprobado sobre `c392ef1`; decide `claude-sonnet-5-5` con 2 de 3; 10 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `02-resumen-sin-documento.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `03-por-materia.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
| `04-documento-pegado.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-documento-que-no-es-el-pedido.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `07-resumen-de-una-conocida.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `08-doctrina-con-el-fallo-delante.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `09-de-que-trata-con-la-ficha-sola.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `10-doctrina-dada-por-hecha.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |

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
| `afirma_que_existe:claude-sonnet-5-5:herramienta` | 3 de 30 (10,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `cita_sin_documento:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `documento_sin_cotejar:claude-sonnet-5-5:herramienta` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 60 (0,0 %) | ≤ 1,7 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 125 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 124 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_del_juez:orden` | 79 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 63 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |

**normativa**: aprobado sobre `c392ef1`; decide `claude-sonnet-5-5` con 2 de 3; 23 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 1/3 | 3/3 | 2/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 0/3 | 3/3 | 1/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 0/3 | 3/3 | 1/3 |  |
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
| `22-lcsp-contrato-menor-en-2019.yaml` | 3/3 | 3/3 | 3/3 | 0/3 |  |
| `23-lpac-articulo-21-en-2016.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 3 de 60 (5,0 %) | ≤ 2,8 % | por encima | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 7 de 60 (11,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 1 de 60 (1,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 3 de 120 (2,5 %) | ≤ 2,8 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 409 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 389 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 138 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 110 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 4 de 66 (6,1 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 22 de 42 (52,4 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 10 de 66 (15,2 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 20 de 42 (47,6 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

**plazos-y-territorio**: aprobado sobre `c392ef1`; decide `claude-sonnet-5-5` con 2 de 3; 12 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |
| `05-plazo-reposicion-leganes.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `06-plazo-reposicion-tordesillas.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `07-cobertura-liquidacion-tributaria.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `08-cobertura-multa-de-trafico.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
| `09-plazo-contencioso-agosto.yaml` | 3/3 | 1/3 | 3/3 | 2/3 |  |
| `10-plazo-reposicion-anterior-a-la-lpac.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `11-plazo-reposicion-sin-leer-el-precepto.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `plazo_sin_fecha:claude-sonnet-5-5:orden` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cobertura_sin_declarar:claude-sonnet-5-5:orden` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |
| `vigencia_sin_declarar:claude-sonnet-5-5:orden` | 0 de 3 (0,0 %) | ≤ 0,0 % | sí | sí |
| `redaccion_sin_fechar:claude-sonnet-5-5:orden` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `plazo_sin_fecha:claude-sonnet-5-5:herramienta` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cobertura_sin_declarar:claude-sonnet-5-5:herramienta` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |
| `vigencia_sin_declarar:claude-sonnet-5-5:herramienta` | 0 de 3 (0,0 %) | ≤ 0,0 % | sí | sí |
| `redaccion_sin_fechar:claude-sonnet-5-5:herramienta` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 33 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-sonnet-5-5:orden` | 0 de 3 (0,0 %) | ≤ 0,0 % | sí | sí |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 0 de 33 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-haiku-4-5-20251001:orden` | 0 de 3 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 0 de 33 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-sonnet-5-5:herramienta` | 0 de 3 (0,0 %) | ≤ 0,0 % | sí | sí |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 0 de 33 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-haiku-4-5-20251001:herramienta` | 0 de 3 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

**Serie de cierres de jurisprudencia** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `779f453` (árbol de `skills/jurisprudencia`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

`afirma_lo_no_leido`:

| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |
|---|---|---|---|---|---|---|---|---|---|
| H25-1 | 2026-10-10 | `3f99446` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H25-2 | 2026-10-10 | `e380f0a` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: 1 respuesta sin juzgar |
| H25-3 | 2026-10-10 | `d51f3d7` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H26-1 | 2026-10-13 | `b29fe6c` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H23.1-1 | 2026-10-16 | `60e8678` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H9-1 | 2026-10-18 | `3f14e3a` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H9-2 | 2026-10-18 | `92fc232` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H20-1 | 2026-10-19 | `77211a9` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H26.1-1 | 2026-10-21 | `444e674` | `3bbd9cd` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H30-1 | 2026-10-22 | `2c8dda6` | `3bbd9cd` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H31-1 | 2026-10-25 | `03cf8d9` | `5a9198f` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H31.1-1 | 2026-10-27 | `a66e9d2` | `5a9198f` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H32-1 | 2026-10-30 | `7f8ef89` | `7dbf0c5` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H39-1 | 2026-11-03 | `a593631` | `7dbf0c5` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H40-1 | 2026-11-04 | `179bd21` | `7dbf0c5` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H41-1 | 2026-11-06 | `69f5ecd` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H41-2 | 2026-11-06 | `d5e6619` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H41.1-1 | 2026-11-10 | `e4915f6` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H41.2-1 | 2026-11-11 | `a27c276` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H41.2-2 | 2026-11-11 | `899708c` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H41.2-3 | 2026-11-11 | `e8715a9` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H41.3-1 | 2026-11-16 | `6701b62` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H41.4-1 | 2026-11-17 | `4b073f8` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H33-1 | 2026-11-20 | `85f986f` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H38-1 | 2026-11-21 | `3473dd4` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H38-2 | 2026-11-21 | `c392ef1` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |

Tasa acumulada con la skill como está: 0 de 240 (0,00 %), de H41.4-1, H33-1, H38-1 y H38-2.
Insuficiente: 240 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

**Serie de cierres de normativa** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `8358dd9` (árbol de `skills/normativa`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

`afirma_lo_no_leido`:

| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |
|---|---|---|---|---|---|---|---|---|---|
| H32-1 | 2026-10-30 | `7f8ef89` | `8566858` | 0/60 | 2/60 | 2/90 | 0/30 | 2/120 | no: otra skill |
| H39-1 | 2026-11-03 | `a593631` | `afbcae6` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H40-1 | 2026-11-04 | `179bd21` | `afbcae6` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H41-1 | 2026-11-06 | `69f5ecd` | `26368b3` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H41-2 | 2026-11-06 | `d5e6619` | `26368b3` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H41.1-1 | 2026-11-10 | `e4915f6` | `26368b3` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H41.2-1 | 2026-11-11 | `a27c276` | `26368b3` | 0/60 | 1/60 | 1/90 | 0/30 | 1/120 | no: otra skill |
| H41.2-2 | 2026-11-11 | `899708c` | `26368b3` | 1/60 | 0/60 | 1/90 | 0/30 | 1/120 | no: otra skill |
| H41.2-3 | 2026-11-11 | `e8715a9` | `26368b3` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H41.3-1 | 2026-11-16 | `6701b62` | `26368b3` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H41.4-1 | 2026-11-17 | `4b073f8` | `3f4260c` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H33-1 | 2026-11-20 | `85f986f` | `3f4260c` | 1/60 | 1/60 | 0/90 | 2/30 | 2/120 | no: otra skill |
| H38-1 | 2026-11-21 | `3473dd4` | `3f4260c` | 1/60 | 3/60 | 0/90 | 4/30 | 4/120 | no: otra skill |
| H38-2 | 2026-11-21 | `c392ef1` | `8358dd9` | 3/60 | 0/60 | 0/90 | 3/30 | 3/120 | sí |

Tasa acumulada con la skill como está: 3 de 120 (2,50 %), de H38-2.
Insuficiente: 120 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

## 4. Revisión que la constitución reserva a la persona

- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-aviso-sin-aviso.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-aviso.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-conflictos-dentro.txtar`.
- Suite de aceptación activada y congelada (4 guiones, escritos desde el spec en la primera tarea): `h38-candado.txtar`, `h38-excluidas.txtar`, `h38-reinstala.txtar`, `h38-tocadas.txtar`.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T001(hecha) T004(hecha) T005(hecha) | excluidas.txtar reinstala.txtar  | tareas hechas | — |
| FR-002 | T001(hecha) T004(hecha) | excluidas.txtar  | tareas hechas | — |
| FR-003 | T001(hecha) T002(hecha) T005(hecha) | excluidas.txtar reinstala.txtar tocadas.txtar  | tareas hechas | — |
| FR-004 | T001(hecha) T002(hecha) | reinstala.txtar tocadas.txtar  | tareas hechas | — |
| FR-005 | T001(hecha) T002(hecha) T005(hecha) | reinstala.txtar tocadas.txtar  | tareas hechas | — |
| FR-006 | T001(hecha) T002(hecha) T005(hecha) | candado.txtar reinstala.txtar  | tareas hechas | — |
| FR-007 | T001(hecha) T002(hecha) | tocadas.txtar  | tareas hechas | — |
| FR-008 | T001(hecha) T004(hecha) | reinstala.txtar  | CONTROL SIN VERIFICAR | `ci:internal/app/e2e_test.go:TestEntregaDelHito`: make ci no está en verde |
| FR-009 | T001(hecha) T002(hecha) T004(hecha) | excluidas.txtar  | CONTROL SIN VERIFICAR | `ci:internal/app/e2e_test.go:TestEntregaDelHito`: make ci no está en verde |
| FR-010 | T001(hecha) T003(hecha) T005(hecha) T006(hecha) | candado.txtar reinstala.txtar  | CONTROL SIN VERIFICAR | `ci:internal/app/e2e_test.go:TestEntregaDelHito`: make ci no está en verde; `ci:internal/core/instalacion/reinstalacion_test.go:TestReinstalacionConcurrente`: make ci no está en verde; `ci:internal/cache/candado_test.go:TestTomarCandado`: make ci no está en verde |
| FR-011 | T001(hecha) T002(hecha) T006(hecha) | candado.txtar  | tareas hechas | — |
| FR-012 | T001(hecha) T005(hecha) | reinstala.txtar  | CONTROL SIN VERIFICAR | `ci:internal/app/e2e_test.go:TestEntregaDelHito`: make ci no está en verde; `ci:internal/arch_test.go:TestArquitectura`: make ci no está en verde |
| FR-013 | T001(hecha) T002(hecha) T003(hecha) T005(hecha) | candado.txtar excluidas.txtar reinstala.txtar tocadas.txtar  | CONTROL SIN VERIFICAR | `ci:internal/app/e2e_test.go:TestEntregaDelHito`: make ci no está en verde |
| FR-014 | T001(hecha) T002(hecha) | excluidas.txtar  | tareas hechas | — |
| FR-015 | T001(hecha) T004(hecha) | excluidas.txtar  | tareas hechas | — |
| FR-016 | T003(hecha) | — | tareas hechas | — |
| FR-017 | T007(hecha) | — | tareas hechas | — |
| FR-018 | T005(hecha) | — | CONTROL SIN VERIFICAR | `ci:Makefile:ci`: make ci no está en verde |
| FR-019 | T002(hecha) T006(hecha) | candado.txtar  | CONTROL SIN VERIFICAR | `ci:Makefile:ci`: make ci no está en verde |
| SC-001 | T001(hecha) T002(hecha) T005(hecha) | reinstala.txtar  | CONTROL SIN VERIFICAR | `ci:internal/app/e2e_test.go:TestEntregaDelHito`: make ci no está en verde |
| SC-002 | T001(hecha) T002(hecha) | tocadas.txtar  | CONTROL SIN VERIFICAR | `ci:internal/app/e2e_test.go:TestEntregaDelHito`: make ci no está en verde |
| SC-003 | T001(hecha) | candado.txtar excluidas.txtar reinstala.txtar tocadas.txtar  | CONTROL SIN VERIFICAR | `ci:internal/app/e2e_test.go:TestEntregaDelHito`: make ci no está en verde |
| SC-004 | T001(hecha) T003(hecha) T006(hecha) | candado.txtar  | CONTROL SIN VERIFICAR | `ci:internal/app/e2e_test.go:TestEntregaDelHito`: make ci no está en verde; `ci:internal/core/instalacion/reinstalacion_test.go:TestReinstalacionConcurrente`: make ci no está en verde |
| SC-005 | T005(hecha) | — | CONTROL SIN VERIFICAR | `ci:internal/arch_test.go:TestArquitectura`: make ci no está en verde |
| SC-006 | T001(hecha) T002(hecha) | excluidas.txtar tocadas.txtar  | CONTROL SIN VERIFICAR | `ci:internal/app/e2e_test.go:TestEntregaDelHito`: make ci no está en verde |
| SC-007 | T005(hecha) T007(hecha) | — | CONTROL SIN VERIFICAR | `ci:Makefile:ci`: make ci no está en verde |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`2190df5`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `be6e7483` fix(H38): motivos de la revisión final: lo vio la ronda 2 (juez A: aprobado; juez B: aprobado). Toca `internal/cache/candado.go`, `internal/core/instalacion/reinstalacion.go`, `internal/core/instalacion/reinstalacion_test.go`, `contracts/aviso-y-reinstalacion.md`, `data-model.md`, `plan.md`, `research.md`, `tasks.md`.
- `5bf12dd1` docs(H38): veredictos de la revisión final: solo registros de `gates/`.
- `3473dd4d` docs(H38): registros del run: solo registros de `gates/`.
- `886d8086` fix(H38): cierre en la plataforma: lo vio la ronda 3 (juez A: aprobado; juez B: rechazado). Toca `CHANGELOG.md`, `internal/evals/laboratorio_familias_test.go`, `skills/normativa/references/lecturas_que_fallan.md`, `skills/plazos-y-territorio/references/administrativo.md`, `plan.md`, `research.md`, `tasks.md`.
- `09cba100` fix(H38): motivos de la revisión final (make ci en rojo): lo vio la ronda 4 (juez A: aprobado; juez B: aprobado). Toca `CONTRIBUTING.md`, `plan.md`, `research.md`.
- `4c1a0e23` docs(H38): veredictos de la revisión final: solo registros de `gates/`.
- `c392ef1b` docs(H38): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/040-h38-las-skills-instaladas/quickstart.md`. Suite de aceptación congelada: `specs/040-h38-las-skills-instaladas/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `5eadf689`: 7 h 30 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh 5eadf689`.
