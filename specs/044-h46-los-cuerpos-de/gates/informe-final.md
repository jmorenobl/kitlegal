# Informe del hito H46 · Los cuerpos de las skills de capacidad, con sitio para las materias (ADR 0063)

Generado por el workflow `hito` el 2026-11-25T15:44:58Z, sobre `e35ece97` de `044-h46-los-cuerpos-de`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/044-h46-los-cuerpos-de/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige el fichero del hito en docs/roadmap/in-progress/ y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `e35ece9` (medición 2): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: consulta-juridica aprobado, jurisprudencia aprobado, normativa aprobado, plazos-y-territorio aprobado (tasas en la sección 3).
- **Umbrales del job**: consulta-juridica: 8 umbrales, los 4 que deciden se cumplen y 4 solo se publican (4 por encima de su umbral: `fuera_de_lo_grabado:claude-sonnet-5-5:orden`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); jurisprudencia: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (2 por encima de su umbral: `afirma_que_existe:claude-sonnet-5-5:orden`, `afirma_que_existe:claude-sonnet-5-5:herramienta`); normativa: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (6 por encima de su umbral: `cuenta_su_proceso:claude-sonnet-5-5:orden`, `cuenta_su_proceso:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-sonnet-5-5:orden`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); plazos-y-territorio: 18 umbrales, los 12 que deciden se cumplen y 6 solo se publican (2 por encima de su umbral: `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`, `lectura_sin_intentar:claude-haiku-4-5-20251001:herramienta`) (sección 3).
- **Revisión final**: juez A aprobado (con `claude-opus-5-5`), juez B aprobado (con `claude-fable-5-2`), 1 ronda.
- **Sesiones con otro modelo que el de su rol**: ninguna (21 sesiones registradas, todas con el modelo que `workflow.yml` da a su rol).
- **Cambios que ningún juez vio: 1** (sección 7). El producto que se fusionaría no lo ha juzgado entero ningún juez.
- **Tareas**: 6 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 14 commits; 82 files changed, 52051 insertions(+), 267 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- corrector_spec: [h] FR-019 y SC-007 no decían qué tests ni qué applets cubren → se acotan a `--describe`, esquema de salida, sobre, códigos y forma de `data` de cada applet; lo que instala `kitlegal skills install` (y sus huellas) es el contenido de `skills/` de FR-010, con el applet `skills` sin cambios; quedan como excepción nombrada los tests y guiones que fijan el texto de las skills (FR-003, FR-008, FR-009, FR-014, FR-015, FR-016) y la suite nueva de FR-021, porque el hito los cambia y sin … (entera en `specs/044-h46-los-cuerpos-de/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run


- Observaciones de los jueces, por debajo del umbral y sin corregir: 11 en `spec-r2.json`, 9 en `plan-r1.json`, 8 en `tasks-r1.json`, 6 en `revision-a-r1.json`, 8 en `revision-b-r1.json`.

### Internos (1)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T006 (1). Enteras en `specs/044-h46-los-cuerpos-de/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**consulta-juridica**: aprobado sobre `e35ece9`; decide `claude-sonnet-5-5` con 2 de 3; 4 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-poliza-tras-el-fallecimiento.yaml` | 2/3 | 0/3 | 3/3 | 0/3 |  |
| `02-una-sola-capacidad.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `03-seguro-prorroga.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `04-despido-plazo.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cobertura_sin_declarar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cobertura_sin_declarar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 11 de 12 (91,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 18 de 12 (150,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 14 de 12 (116,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 19 de 12 (158,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

**jurisprudencia**: aprobado sobre `e35ece9`; decide `claude-sonnet-5-5` con 2 de 3; 10 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-resumen-sin-documento.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `03-por-materia.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-documento-pegado.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-documento-que-no-es-el-pedido.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `07-resumen-de-una-conocida.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `08-doctrina-con-el-fallo-delante.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `09-de-que-trata-con-la-ficha-sola.yaml` | 3/3 | 0/3 | 3/3 | 1/3 |  |
| `10-doctrina-dada-por-hecha.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |

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
| `duracion_del_juez:orden` | 54 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 55 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |

**normativa**: aprobado sobre `e35ece9`; decide `claude-sonnet-5-5` con 2 de 3; 23 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 0/3 | 3/3 | 1/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 1/3 | 3/3 | 2/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `11-no-activa-programacion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `12-no-activa-acuerdo-entre-amigos.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `13-lrbrl-atribuciones-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `14-trlrhl-impuestos-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `15-irpf-rendimientos-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `16-lrjsp-legalidad-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `17-ltaibg-plazo-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `18-lrjpac-norma-derogada.yaml` | 3/3 | — | 2/3 | — | informativa |
| `19-lcsp-contrato-menor-redaccion-cambiada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `20-lcsp-dos-bloques-redaccion-cambiada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `21-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |
| `22-lcsp-contrato-menor-en-2019.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `23-lpac-articulo-21-en-2016.yaml` | 3/3 | 3/3 | 3/3 | 0/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 6 de 60 (10,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 2 de 60 (3,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 120 (0,0 %) | ≤ 2,8 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 450 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 433 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 114 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 106 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 5 de 66 (7,6 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 20 de 42 (47,6 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 5 de 66 (7,6 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 21 de 42 (50,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

**plazos-y-territorio**: aprobado sobre `e35ece9`; decide `claude-sonnet-5-5` con 2 de 3; 12 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |
| `05-plazo-reposicion-leganes.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-plazo-reposicion-tordesillas.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `07-cobertura-liquidacion-tributaria.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
| `08-cobertura-multa-de-trafico.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `09-plazo-contencioso-agosto.yaml` | 3/3 | 1/3 | 3/3 | 2/3 |  |
| `10-plazo-reposicion-anterior-a-la-lpac.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `11-plazo-reposicion-sin-leer-el-precepto.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |

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
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 2 de 33 (6,1 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-haiku-4-5-20251001:herramienta` | 1 de 3 (33,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

**Serie de cierres de jurisprudencia** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `74ea0b7` (árbol de `skills/jurisprudencia`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

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
| H41.4-1 | 2026-11-17 | `4b073f8` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H33-1 | 2026-11-20 | `85f986f` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H38-1 | 2026-11-21 | `3473dd4` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H38-2 | 2026-11-21 | `c392ef1` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H36-1 | 2026-11-22 | `0c8f8de` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H45-1 | 2026-11-24 | `23db28c` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H45.1-1 | 2026-11-24 | `52af770` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H46-1 | 2026-11-25 | `e5611f6` | `74ea0b7` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H46-2 | 2026-11-25 | `e35ece9` | `74ea0b7` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |

Tasa acumulada con la skill como está: 0 de 120 (0,00 %), de H46-1 y H46-2.
Insuficiente: 120 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

**Serie de cierres de normativa** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `bbc1a43` (árbol de `skills/normativa`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

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
| H38-2 | 2026-11-21 | `c392ef1` | `8358dd9` | 3/60 | 0/60 | 0/90 | 3/30 | 3/120 | no: otra skill |
| H36-1 | 2026-11-22 | `0c8f8de` | `8358dd9` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H45-1 | 2026-11-24 | `23db28c` | `f0a4da8` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H45.1-1 | 2026-11-24 | `52af770` | `f0a4da8` | 1/60 | 0/60 | 0/90 | 1/30 | 1/120 | no: otra skill |
| H46-1 | 2026-11-25 | `e5611f6` | `bbc1a43` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | sí |
| H46-2 | 2026-11-25 | `e35ece9` | `bbc1a43` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | sí |

Tasa acumulada con la skill como está: 0 de 240 (0,00 %), de H46-1 y H46-2.
Insuficiente: 240 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

## 4. Revisión que la constitución reserva a la persona

- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h41-1-skills-instaladas.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h41-skills-instaladas.txtar`.
- Suite de aceptación activada y congelada (2 guiones, escritos desde el spec en la primera tarea): `h46-tabla-de-tres-columnas.txtar`, `h46-textos-en-su-referencia.txtar`.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T001(hecha) T002(hecha) T006(hecha) | tabla-de-tres-columnas.txtar  | comprobado por su control | `ci:internal/skills/comandos_test.go:TestRenderizarTabla`, en make ci (verde); `ci:internal/skills/sincronia_test.go:TestRegenerarYComparar`, en make ci (verde); `ci:internal/skills/sincronia_test.go:TestCuerpoDeLaSkill`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-002 | T002(hecha) T006(hecha) | tabla-de-tres-columnas.txtar  | comprobado por su control | `ci:internal/skills/comandos_test.go:TestRenderizarTabla`, en make ci (verde); `ci:internal/skills/comandos_test.go:TestDescripcionDeVerbo`, en make ci (verde) |
| FR-003 | T002(hecha) T006(hecha) | — | comprobado por su control | `ci:internal/skills/comandos_test.go:TestRenderizarTabla`, en make ci (verde); `ci:internal/skills/sincronia_test.go:TestRegenerarYComparar`, en make ci (verde); `ci:internal/skills/sincronia_test.go:TestCuerpoDeLaSkill`, en make ci (verde) |
| FR-004 | T002(hecha) | tabla-de-tres-columnas.txtar  | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-005 | T004(hecha) T006(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-006 | T004(hecha) T006(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-007 | T004(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-008 | T004(hecha) | — | comprobado por su control | `ci:Makefile:ci`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-009 | T004(hecha) | — | tareas hechas | — |
| FR-010 | T001(hecha) T003(hecha) | textos-en-su-referencia.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-011 | T001(hecha) T003(hecha) | textos-en-su-referencia.txtar  | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-012 | T001(hecha) T003(hecha) | textos-en-su-referencia.txtar  | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-013 | T003(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde) |
| FR-014 | T003(hecha) T006(hecha) | textos-en-su-referencia.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde) |
| FR-015 | T003(hecha) T006(hecha) | textos-en-su-referencia.txtar  | comprobado por su control | `ci:internal/skills/remisiones_test.go:TestRemisionesDelRepositorio`, en make ci (verde) |
| FR-016 | T003(hecha) T006(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-017 | T002(hecha) T004(hecha) | — | tareas hechas | — |
| FR-018 | T005(hecha) | — | tareas hechas | — |
| FR-019 | T005(hecha) | — | comprobado por su control | `ci:Makefile:test`, en make ci (verde); `ci:Makefile:schema-check`, en make ci (verde) |
| FR-020 | T003(hecha) | — | tareas hechas | — |
| FR-021 | T001(hecha) | tabla-de-tres-columnas.txtar textos-en-su-referencia.txtar  | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-001 | T004(hecha) | — | comprobado por su control | `ci:Makefile:ci`, en make ci (verde) |
| SC-002 | T002(hecha) T003(hecha) T004(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| SC-003 | T004(hecha) T006(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| SC-004 | T003(hecha) T006(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde) |
| SC-005 | T003(hecha) T006(hecha) | — | comprobado por su control | `ci:internal/skills/remisiones_test.go:TestRemisionesDelRepositorio`, en make ci (verde) |
| SC-006 | T001(hecha) T006(hecha) | tabla-de-tres-columnas.txtar textos-en-su-referencia.txtar  | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-007 | T006(hecha) | — | comprobado por su control | `ci:Makefile:test`, en make ci (verde); `ci:Makefile:schema-check`, en make ci (verde) |
| SC-008 | T006(hecha) | — | comprobado por su control | `evals:normativa:sin_activar:claude-sonnet-5-5:orden`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 120 (0,0 %), ≤ 2,8 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:normativa:duracion_de_las_sesiones:orden`: 450, ≤ 900; `evals:normativa:duracion_de_las_sesiones:herramienta`: 433, ≤ 900; `evals:normativa:duracion_del_juez:orden`: 114, ≤ 900; `evals:normativa:duracion_del_juez:herramienta`: 106, ≤ 900; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 54, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 55, ≤ 900; `evals:consulta-juridica:sin_activar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:consulta-juridica:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:consulta-juridica:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:consulta-juridica:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 % |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`1dd843f`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `e5611f67` docs(H46): veredictos de la revisión final: solo registros de `gates/`.
- `bf450baa` docs(H46): informe final del run: solo registros de `gates/`.
- `e35ece97` test(H46): la regla de los 9600, en rojo desde internal/skills: **ningún juez lo vio**.

### Cambios que ningún juez vio

**Anomalía**: cambian el producto después del último veredicto y ningún juez los ha juzgado. Desde el workflow 2.3.0 el propio bucle de revisión y cierre los lleva a los jueces antes de volver a medir: si están aquí, el run es anterior, se interrumpió o alguien los añadió a mano.

- **`e35ece97` test(H46): la regla de los 9600, en rojo desde internal/skills**: `internal/skills/paquetes_test.go`.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/044-h46-los-cuerpos-de/quickstart.md`. Suite de aceptación congelada: `specs/044-h46-los-cuerpos-de/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `bacd89bd`: 5 h 16 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh bacd89bd`.
