# Informe del hito H41 · Las skills del producto mandan leer sus referencias

Generado por el workflow `hito` el 2026-11-06T22:57:32Z, sobre `d5e66197` de `034-h41-las-skills-del`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/034-h41-las-skills-del/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige el fichero del hito en docs/roadmap/in-progress/ y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `d5e6619` (medición 2): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: consulta-juridica aprobado, jurisprudencia aprobado, normativa aprobado, plazos-y-territorio aprobado (tasas en la sección 3).
- **Umbrales del job**: consulta-juridica: 4 umbrales, ninguno decide y 4 solo se publican (4 por encima de su umbral: `fuera_de_lo_grabado:claude-sonnet-5-5:orden`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); jurisprudencia: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (2 por encima de su umbral: `afirma_que_existe:claude-sonnet-5-5:orden`, `afirma_que_existe:claude-sonnet-5-5:herramienta`); normativa: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (5 por encima de su umbral: `cuenta_su_proceso:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-sonnet-5-5:orden`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); plazos-y-territorio: 18 umbrales, los 12 que deciden se cumplen y 6 solo se publican (1 por encima de su umbral: `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`) (sección 3).
- **Revisión final**: juez A aprobado (con `claude-opus-5-5`), juez B aprobado (con `claude-fable-5-1`), 3 rondas en 2 ciclos (el primero juzga el hito; cada uno de los siguientes, lo que cambió después del último veredicto, o el hito entero si es un rejuicio).
- **Sesiones con otro modelo que el de su rol**: ninguna (32 sesiones registradas, todas con el modelo que `workflow.yml` da a su rol).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 9 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 18 commits; 220 files changed, 56985 insertions(+), 501 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- corrector_plan: contracts/tabla.md §1, research D17 y la fila V del Constitution Check prometían código 1 y ninguna tabla a un `sondeo.json` sin `referencias_del_producto` en un sondeo con `<skill>-producto`, y D20 traía tal cual la comprobación y el mensaje propios de la #245 (`referenciasDeLasVariantes`, «… no tiene la lista referencias_del_producto de la variante …»), un estado al que solo llega una edición a mano porque el arnés siempre escribe la lista (FR-032) → se quita ese caso de los tr… (entera en `specs/034-h41-las-skills-del/gates/supuestos.md` o `clarify-respuestas.json`)
- T006: contracts/arnes.md §2 da la forma del mensaje de la mezcla como «se mezclan variantes de <a>, … y variantes de <z>», una por familia pedida, sin escribir el caso de tres o más → cada familia pedida con su «variantes de <skill>» (la de H39, «variantes del laboratorio de H39»), unidas con comas y la última con «y» (`VARIANTES: se mezclan variantes del laboratorio de H39, variantes de normativa, variantes de plazos-y-territorio y variantes de jurisprudencia`): es la lectura de «una por famili… (entera en `specs/034-h41-las-skills-del/gates/supuestos.md` o `clarify-respuestas.json`)

**Alcance (lo que queda fuera o dentro del hito)**

- tasks: FR-052 y FR-053 piden el sondeo de las diecinueve filas de FR-033 y los ADR que leen sus tablas, y el run no puede lanzar sesiones con modelo ni usar la credencial del sondeo (FR-031; ADR 0032) → el run no lanza el sondeo ni escribe los ADR de FR-052 y FR-053: la fusión queda pendiente de las tres tablas, una por familia, en `evidencias/` con el modelo y la versión de Claude Code anotados, y de su lectura en la propuesta de cambio, las dos fuera del run y de mano de la persona; si `normat… (entera en `specs/034-h41-las-skills-del/gates/supuestos.md` o `clarify-respuestas.json`)
- reparar_cierre: `sin_activar:claude-sonnet-5-5:herramienta` 1 de 60 en el cierre de `69f5ecd1` es la cuarta caída de la eval 04 en modo herramienta (antes, `7f8ef89` y `75bf303` en H32 y la suite de v0.9.0; Claude Code 2.1.284) → sin cambiar skills ni binario, va a la persona como serie de cuatro, como decidió una persona en D27 de H32, aunque la medición siguiente salga en verde: decide la persona, con un ADR sobre el umbral o con un hito que mida la activación en modo herramienta. Por qué no s… (entera en `specs/034-h41-las-skills-del/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_revision: research D21 y la línea `reparar_cierre` de este fichero daban «dos de doce de los cuatro cierres» y «2 de 12 frente a 0 de 70» cuando el repositorio registra cuatro caídas de la eval 04 en modo herramienta, y tomaban «lo mide el job en la medición siguiente», la opción que D25 y D27 de H32 rechazaron, contra la decisión de una persona en D27 de H32 para esta misma caída → lectura conservadora de una decisión cerrada, sin tocar código, skills, evals, umbrales ni `evals.yml`: … (entera en `specs/034-h41-las-skills-del/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run

- 2026-11-06T21:40:14Z · La revisión final abre el ciclo 2 para juzgar lo que cambió fuera de gates/ después del veredicto sobre 0169108: 7ebf7dd «fix(H41): cierre en la plataforma».

- Observaciones de los jueces, por debajo del umbral y sin corregir: 12 en `spec-r2.json`, 9 en `plan-r2.json`, 7 en `tasks-r1.json`, 5 en `revision-a-r3.json`, 6 en `revision-b-r3.json`.

### Internos (9)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T002 (1), T003 (2), T004 (2), T005 (3), T009 (1). Enteras en `specs/034-h41-las-skills-del/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**consulta-juridica**: aprobado sobre `d5e6619`; decide `claude-sonnet-5-5` con 2 de 3; 2 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-poliza-tras-el-fallecimiento.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `02-una-sola-capacidad.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 21 de 6 (350,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 10 de 6 (166,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 18 de 6 (300,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 6 de 6 (100,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

**jurisprudencia**: aprobado sobre `d5e6619`; decide `claude-sonnet-5-5` con 2 de 3; 10 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-resumen-sin-documento.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `03-por-materia.yaml` | 3/3 | 1/3 | 3/3 | 3/3 |  |
| `04-documento-pegado.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-documento-que-no-es-el-pedido.yaml` | 2/3 | 1/3 | 3/3 | 0/3 |  |
| `07-resumen-de-una-conocida.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `08-doctrina-con-el-fallo-delante.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `09-de-que-trata-con-la-ficha-sola.yaml` | 3/3 | 0/3 | 3/3 | 2/3 |  |
| `10-doctrina-dada-por-hecha.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 1,7 % | dentro | no: solo se publica, no es un control |
| `afirma_que_existe:claude-sonnet-5-5:orden` | 2 de 30 (6,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
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
| `duracion_del_juez:orden` | 55 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 52 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |

**normativa**: aprobado sobre `d5e6619`; decide `claude-sonnet-5-5` con 2 de 3; 23 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 0/3 | 3/3 | 2/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
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
| `22-lcsp-contrato-menor-en-2019.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `23-lpac-articulo-21-en-2016.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 5 de 60 (8,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 120 (0,0 %) | ≤ 2,8 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 465 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 453 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 93 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 87 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 5 de 66 (7,6 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 26 de 42 (61,9 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 10 de 66 (15,2 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 23 de 42 (54,8 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

**plazos-y-territorio**: aprobado sobre `d5e6619`; decide `claude-sonnet-5-5` con 2 de 3; 12 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 0/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |
| `05-plazo-reposicion-leganes.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-plazo-reposicion-tordesillas.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `07-cobertura-liquidacion-tributaria.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `08-cobertura-multa-de-trafico.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `09-plazo-contencioso-agosto.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `10-plazo-reposicion-anterior-a-la-lpac.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `11-plazo-reposicion-sin-leer-el-precepto.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml` | 3/3 | 1/3 | 3/3 | 2/3 |  |

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
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 1 de 33 (3,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-haiku-4-5-20251001:orden` | 0 de 3 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 0 de 33 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-sonnet-5-5:herramienta` | 0 de 3 (0,0 %) | ≤ 0,0 % | sí | sí |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 0 de 33 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-haiku-4-5-20251001:herramienta` | 0 de 3 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

**Serie de cierres de jurisprudencia** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `b5b62e3` (árbol de `skills/jurisprudencia`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

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
| H41-1 | 2026-11-06 | `69f5ecd` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H41-2 | 2026-11-06 | `d5e6619` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |

Tasa acumulada con la skill como está: 0 de 120 (0,00 %), de H41-1 y H41-2.
Insuficiente: 120 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

**Serie de cierres de normativa** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `26368b3` (árbol de `skills/normativa`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

`afirma_lo_no_leido`:

| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |
|---|---|---|---|---|---|---|---|---|---|
| H32-1 | 2026-10-30 | `7f8ef89` | `8566858` | 0/60 | 2/60 | 2/90 | 0/30 | 2/120 | no: otra skill |
| H39-1 | 2026-11-03 | `a593631` | `afbcae6` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H40-1 | 2026-11-04 | `179bd21` | `afbcae6` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H41-1 | 2026-11-06 | `69f5ecd` | `26368b3` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | sí |
| H41-2 | 2026-11-06 | `d5e6619` | `26368b3` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | sí |

Tasa acumulada con la skill como está: 0 de 240 (0,00 %), de H41-1 y H41-2.
Insuficiente: 240 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

## 4. Revisión que la constitución reserva a la persona

- Fixtures nuevos (105), por directorio:
  - `internal/evals/testdata/laboratorio/familias/jurisprudencia/sondeo/h41-a/jurisprudencia-control/herramienta/03-por-materia-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/jurisprudencia/sondeo/h41-a/jurisprudencia-control/orden/01-existe-con-numero-y-fecha-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/jurisprudencia/sondeo/h41-a/jurisprudencia-producto/herramienta/06-documento-que-no-es-el-pedido-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/jurisprudencia/sondeo/h41-a/jurisprudencia-producto/orden/04-documento-pegado-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/jurisprudencia/sondeo/h41-a/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/jurisprudencia/sondeo/h41-b/jurisprudencia-control/herramienta/01-existe-con-numero-y-fecha-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/jurisprudencia/sondeo/h41-b/jurisprudencia-producto/herramienta/04-documento-pegado-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/jurisprudencia/sondeo/h41-b/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/jurisprudencia/`: `tabla.md`
  - `internal/evals/testdata/laboratorio/familias/mezcla/sondeo/h41/jurisprudencia-control/orden/01-existe-con-numero-y-fecha-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/mezcla/sondeo/h41/normativa-control/orden/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/mezcla/sondeo/h41/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-control/herramienta/05-trlrhl-impuestos-municipales-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-control/orden/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-control/orden/22-lcsp-contrato-menor-en-2019-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-producto/herramienta/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-producto/herramienta/22-lcsp-contrato-menor-en-2019-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-producto/orden/15-irpf-rendimientos-por-materia-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-remisiones-en-orden/herramienta/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-remisiones-en-orden/herramienta/05-trlrhl-impuestos-municipales-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-testigo-en-referencia/herramienta/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-testigo-en-referencia/herramienta/05-trlrhl-impuestos-municipales-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-b/normativa-control/orden/05-trlrhl-impuestos-municipales-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-b/normativa-cuerpo-corto/herramienta/15-irpf-rendimientos-por-materia-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-b/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/normativa/`: `tabla.md`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-condicion-por-orden/herramienta/05-plazo-reposicion-leganes-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-condicion-por-orden/herramienta/11-plazo-reposicion-sin-leer-el-precepto-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-condicion-por-orden/orden/11-plazo-reposicion-sin-leer-el-precepto-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-control/herramienta/05-plazo-reposicion-leganes-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-control/orden/02-territorio-municipio-no-cubierto-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-lista-con-verbo/herramienta/02-territorio-municipio-no-cubierto-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-lista-con-verbo/orden/10-plazo-reposicion-anterior-a-la-lpac-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-producto/herramienta/10-plazo-reposicion-anterior-a-la-lpac-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-producto/herramienta/11-plazo-reposicion-sin-leer-el-precepto-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-producto/orden/05-plazo-reposicion-leganes-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-sin-condicion/herramienta/02-territorio-municipio-no-cubierto-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-sin-condicion/orden/05-plazo-reposicion-leganes-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/`: `tabla.md`
- Suite de aceptación activada y congelada (2 guiones, escritos desde el spec en la primera tarea): `h41-laboratorio-tabla.txtar`, `h41-skills-instaladas.txtar`.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T001(hecha) T004(hecha) | skills-instaladas.txtar  | comprobado por su control | `ci:internal/skills/remisiones_test.go:TestRemisionesDelRepositorio`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-002 | T001(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/skills/remisiones_test.go:TestRemisionesDelRepositorio`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-003 | T001(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/skills/remisiones_test.go:TestRemisionesDelRepositorio`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-004 | T001(hecha) T003(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-005 | T001(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-006 | T001(hecha) T003(hecha) | skills-instaladas.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-007 | T003(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde) |
| FR-008 | T003(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-009 | T003(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-010 | T001(hecha) T003(hecha) T004(hecha) | skills-instaladas.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-020 | T002(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestVariantesDeControl`, en make ci (verde) |
| FR-021 | T002(hecha) T005(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_normativa_test.go:TestVariantesDeNormativa`, en make ci (verde) |
| FR-022 | T002(hecha) T006(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_familias_test.go:TestVariantesDeControl`, en make ci (verde) |
| FR-023 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestVariantesDePlazosYTerritorio`, en make ci (verde) |
| FR-024 | T005(hecha) | — | tareas hechas | — |
| FR-025 | T001(hecha) T002(hecha) T005(hecha) | skills-instaladas.txtar  | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde); `ci:internal/evals/laboratorio_test.go:TestLaboratorioFueraDelProducto`, en make ci (verde) |
| FR-030 | T006(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde) |
| FR-031 | — | — | SIN TAREA | — |
| FR-032 | T001(hecha) T006(hecha) T007(hecha) T008(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_familias_test.go:TestVariantesDeControl`, en make ci (verde) |
| FR-033 | T006(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde) |
| FR-040 | T001(hecha) T007(hecha) T008(hecha) T009(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestLecturaConCd`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-041 | T001(hecha) T007(hecha) T008(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestLecturaConCd`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-042 | T007(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-050 | T009(hecha) | — | tareas hechas | — |
| FR-051 | T009(hecha) | — | comprobado por su control | `evals:normativa:sin_activar:claude-sonnet-5-5:orden`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 120 (0,0 %), ≤ 2,8 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:normativa:duracion_de_las_sesiones:orden`: 465, ≤ 900; `evals:normativa:duracion_de_las_sesiones:herramienta`: 453, ≤ 900; `evals:normativa:duracion_del_juez:orden`: 93, ≤ 900; `evals:normativa:duracion_del_juez:herramienta`: 87, ≤ 900; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 55, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 52, ≤ 900; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 % |
| FR-052 | T009(hecha) | — | tareas hechas | — |
| FR-053 | T009(hecha) | — | tareas hechas | — |
| SC-001 | T001(hecha) T004(hecha) | skills-instaladas.txtar  | comprobado por su control | `ci:internal/skills/remisiones_test.go:TestRemisionesDelRepositorio`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-002 | T003(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| SC-003 | T003(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| SC-004 | T001(hecha) T003(hecha) T004(hecha) | skills-instaladas.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-005 | T001(hecha) T003(hecha) | skills-instaladas.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-006 | T001(hecha) T007(hecha) T008(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestLecturaConCd`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-007 | T002(hecha) T006(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_familias_test.go:TestVariantesDeControl`, en make ci (verde) |
| SC-008 | T006(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde) |
| SC-009 | T001(hecha) T002(hecha) T005(hecha) T007(hecha) | skills-instaladas.txtar  | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde); `ci:internal/evals/laboratorio_familias_test.go:TestVariantesDeControl`, en make ci (verde); `ci:internal/evals/laboratorio_normativa_test.go:TestVariantesDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_familias_test.go:TestVariantesDePlazosYTerritorio`, en make ci (verde); `ci:internal/evals/laboratorio_test.go:TestLaboratorioFueraDelProducto`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-010 | — | — | SIN TAREA | `evals:normativa:sin_activar:claude-sonnet-5-5:orden`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 120 (0,0 %), ≤ 2,8 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:normativa:duracion_de_las_sesiones:orden`: 465, ≤ 900; `evals:normativa:duracion_de_las_sesiones:herramienta`: 453, ≤ 900; `evals:normativa:duracion_del_juez:orden`: 93, ≤ 900; `evals:normativa:duracion_del_juez:herramienta`: 87, ≤ 900; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 55, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 52, ≤ 900; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 % |
| SC-011 | — | — | SIN TAREA | — |
| SC-012 | — | — | SIN TAREA | — |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`0169108`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `69f5ecd1` docs(H41): veredictos de la revisión final: solo registros de `gates/`.
- `7ebf7dde` fix(H41): cierre en la plataforma: lo vio la ronda 2 (juez A: rechazado; juez B: aprobado). Toca `research.md`.
- `a2268bb7` fix(H41): motivos de la revisión final: lo vio la ronda 3 (juez A: aprobado; juez B: aprobado). Toca `research.md`.
- `7ebded19` docs(H41): veredictos de la revisión final: solo registros de `gates/`.
- `d5e66197` docs(H41): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/034-h41-las-skills-del/quickstart.md`. Suite de aceptación congelada: `specs/034-h41-las-skills-del/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `c285050e`: 9 h 25 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh c285050e`.
