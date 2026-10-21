# Informe del hito H26.1 · Las evals sirven lo que la skill lleva a leer y hacen fallar a propósito la lectura que quieren probar (ADR 0043)

Generado por el workflow `hito` el 2026-10-21T13:45:29Z, sobre `444e674` de `025-h26-1-las-evals-sirven`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/025-h26-1-las-evals-sirven/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige la sección del hito en docs/ROADMAP.md y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `444e674` (medición 1): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: boe-legislacion aprobado, jurisprudencia aprobado, legal-core aprobado (tasas en la sección 3).
- **Umbrales del job**: boe-legislacion: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (3 por encima de su umbral: `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); jurisprudencia: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (2 por encima de su umbral: `afirma_que_existe:claude-sonnet-5-5:orden`, `afirma_que_existe:claude-sonnet-5-5:herramienta`); legal-core: 18 umbrales, los 12 que deciden se cumplen y 6 solo se publican (5 por encima de su umbral: `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `lectura_sin_intentar:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`, `lectura_sin_intentar:claude-haiku-4-5-20251001:herramienta`) (sección 3).
- **Revisión final**: juez A aprobado (con `claude-opus-5-5`), juez B aprobado (con `claude-fable-5-1`), 2 rondas.
- **Sesiones con otro modelo que el de su rol**: ninguna (41 sesiones registradas, todas con el modelo que `workflow.yml` da a su rol).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 16 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 24 commits; 95 files changed, 10157 insertions(+), 491 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- corrector_spec: «Uso, de fuera adentro» no trataba `lectura_sin_intentar:<modelo>:<modo>` → añadida su entrada: 4 umbrales por skill y 12 en los tres trabajos, unos 300 bytes, el motivo que nombra las sesiones que cuentan (como mucho las tres de la 11 por modo del modelo que decide), recalculado en cada job, total 0 en las skills sin una eval así, y cuándo deja de darse.
- corrector_plan: FR-033 y su entrada de «Uso» publicaban `lectura_sin_intentar:<modelo>:<modo>` con total 0 en el informe de `boe-legislacion` y de `jurisprudencia`, que no tienen ninguna eval con un comando de bloque en su `sin_servir`: 4 umbrales «0 de 0» por job, 2 que deciden, que no miden nada y no dejan de darse → el umbral va solo en el informe de una skill con al menos una eval así (hoy, `legal-core`: 4 umbrales; ninguno en las otras dos), como `plazo_sin_fecha` solo en el de la que tiene… (entera en `specs/025-h26-1-las-evals-sirven/gates/supuestos.md` o `clarify-respuestas.json`)
- T010: contracts §4 pone `fuera_de_lo_grabado` «en el informe de toda skill», y el informe del instrumento sin medir (`informeDelInstrumentoSinMedir`, contracts/informe-del-job.md §5 de H24) no lleva por diseño ningún umbral, porque saldrían cumplidos sin ninguna sesión → ese informe sigue sin umbrales: no abre ninguna sesión y no hay lectura que contar, y es lo que menos comportamiento añade.

**Alcance (lo que queda fuera o dentro del hito)**

- corrector_revision: la base de la medida del juez servía las `lecturas` de las evals (research D4), que FR-004 no pide —solo lo de FR-001, `data/`—, y con ellas 9 de los 259 casos etiquetados de `boe-legislacion` (3 derivados y 6 defectos de procedencia `lectura` de los informes de H7.1, H7.2, H7.4 y H21) se reconstruían con un artículo leído con 0 que su sesión leyó con 5 → lectura conservadora: `leerLasEvalsDeHoy` prepara la base con `consultasDeLaCache(sinSusLecturas(evals), datosDelRepositor… (entera en `specs/025-h26-1-las-evals-sirven/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q1 (criterio d, conservadora): Que la respuesta lleve la línea `⚠ REDACCIÓN A FECHA:` no cambia el juicio de la eval 11. En FR-017, «sin `⚠ REDACCIÓN A FECHA:`» quiere decir solo que la eval no lleva `redaccion_fechada`, a diferencia de la 05. La eval no espera esa línea, pero ningún control del job mira si está en la 11: no hay campo ni regla nueva en el formato de eval, ni motivo propio en el juicio, ni sesión sintética que la lleve. En el job, FR-029 se mide por lo que la 11 ya compru… (entera en `specs/025-h26-1-las-evals-sirven/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q3 (criterio d, conservadora): De FR-029 y FR-030, el job comprueba solo lo que tiene forma y ya está en el spec. En la eval 11, la línea `⚠ PLAZO:` con `2026-10-13`, con `plazo_sin_fecha`, y la lectura de `a124` intentada, con FR-016 y el umbral de Q2. En las evals 07, 08 y 12, la línea `⚠ NO CUBIERTO:` con su norma y ninguna línea del plazo, con `cobertura_sin_declarar`, y `plazos calcular` en `prohibidos`. Lo demás no lo comprueba ningún control del job. En FR-029, que la respuesta di… (entera en `specs/025-h26-1-las-evals-sirven/gates/supuestos.md` o `clarify-respuestas.json`)

**Skill (lo que pide, dice o comprueba una skill)**

- corrector_spec: FR-015 contradecía a FR-023 y FR-024 en la invocación que consulta una lectura de `sin_servir` y es a la vez una llamada sin sobre o una orden de salida no escrita → prevalecen FR-023 y FR-024: va a `otras_fallidas` con su `motivo` (y `bash`), con `codigo: null` o 1, y no es una lectura intentada ni en `sesiones_con_lectura_sin_servir` ni en `lectura_sin_intentar` (FR-033), porque no llegó a la lectura que la caché no sirve: la eval quiere ver la reacción de la skill al 4 o 5 que… (entera en `specs/025-h26-1-las-evals-sirven/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_spec: FR-016 no decía si la invocación con 4 o 5 que satisface un comando de bloque con `fecha` tiene que leerlo en esa fecha → la tiene que leer como hoy lo pide el comando (en su `fecha`, si la lleva, como `leeElBloque`); solo cambia el código admitido. La intentada sin `--fecha` o con otra fecha deja el comando ausente; `lectura_sin_intentar` (FR-033) la sigue dando por intentada, porque mide otra cosa.
- corrector_spec: la ronda 2 contaba como «sin intentar» la orden de salida no escrita (FR-024) que consulta una lectura de `sin_servir`, aunque el kernel ejecuta el applet antes de presentar el sobre y el 1 sale al escribir → corrige la línea 1 de este fichero en lo que toca a esa orden: va a `lecturas_sin_servir` con su `codigo: 1`, su `motivo: "salida_no_escrita"` y su `bash`, no a `otras_fallidas`, y es lectura intentada en `sesiones_con_lectura_sin_servir` y en `lectura_sin_intentar` (FR-015,… (entera en `specs/025-h26-1-las-evals-sirven/gates/supuestos.md` o `clarify-respuestas.json`)
- T011: «el primer punto del paso 10» es una línea y no trata la lectura fallida, así que no se puede reescribir con menos líneas por sí solo → se funde con el párrafo de entrada del paso 10, que se reescribe con él (de 5 + 1 líneas a 4: «se cita como ella cita una redacción a fecha, con su línea `⚠ REDACCIÓN A FECHA:`»), y el punto nuevo de FR-029 pasa a ser el primero; ninguna regla ni ningún caso sale: dejan de estar el remite «Cómo se presenta un plazo» de los dos pasos, porque la etiqueta de … (entera en `specs/025-h26-1-las-evals-sirven/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run


- Observaciones de los jueces, por debajo del umbral y sin corregir: 12 en `spec-r3.json`, 9 en `plan-r2.json`, 8 en `tasks-r1.json`, 5 en `revision-a-r2.json`, 4 en `revision-b-r2.json`.

### Internos (12)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T003 (1), T004 (1), T006 (2), T009 (3), T010 (1), T011 (1), T013 (2), barrido (1). Enteras en `specs/025-h26-1-las-evals-sirven/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**boe-legislacion**: aprobado sobre `444e674`; decide `claude-sonnet-5-5` con 2 de 3; 23 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `14-trlrhl-impuestos-por-materia.yaml` | 3/3 | — | 3/3 | — | cambiada · informativa |
| `15-irpf-rendimientos-por-materia.yaml` | 3/3 | — | 3/3 | — | cambiada · informativa |
| `16-lrjsp-legalidad-por-materia.yaml` | 3/3 | — | 3/3 | — | cambiada · informativa |
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `11-no-activa-programacion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `12-no-activa-acuerdo-entre-amigos.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `13-lrbrl-atribuciones-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `17-ltaibg-plazo-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `18-lrjpac-norma-derogada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `19-lcsp-contrato-menor-redaccion-cambiada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `20-lcsp-dos-bloques-redaccion-cambiada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `21-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |
| `22-lcsp-contrato-menor-en-2019.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `23-lpac-articulo-21-en-2016.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 1 de 60 (1,7 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 1 de 120 (0,8 %) | ≤ 2,8 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 483 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 568 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 95 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 91 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 66 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 4 de 42 (9,5 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 2 de 66 (3,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 7 de 42 (16,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

**jurisprudencia**: aprobado sobre `444e674`; decide `claude-sonnet-5-5` con 2 de 3; 10 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-resumen-sin-documento.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `03-por-materia.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `04-documento-pegado.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-documento-que-no-es-el-pedido.yaml` | 3/3 | 0/3 | 2/3 | 0/3 |  |
| `07-resumen-de-una-conocida.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `08-doctrina-con-el-fallo-delante.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `09-de-que-trata-con-la-ficha-sola.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `10-doctrina-dada-por-hecha.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |

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
| `afirma_que_existe:claude-sonnet-5-5:herramienta` | 1 de 30 (3,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `cita_sin_documento:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `documento_sin_cotejar:claude-sonnet-5-5:herramienta` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 60 (0,0 %) | ≤ 1,7 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 125 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 124 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_del_juez:orden` | 45 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 41 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |

**legal-core**: aprobado sobre `444e674`; decide `claude-sonnet-5-5` con 2 de 3; 12 evals, 2 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `11-plazo-reposicion-sin-leer-el-precepto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 | **nueva** |
| `12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml` | 3/3 | 0/3 | 3/3 | 0/3 | **nueva** |
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 0/3 | — | — |  |
| `05-plazo-reposicion-leganes.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `06-plazo-reposicion-tordesillas.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `07-cobertura-liquidacion-tributaria.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `08-cobertura-multa-de-trafico.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `09-plazo-contencioso-agosto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `10-plazo-reposicion-anterior-a-la-lpac.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |

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
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 47 de 33 (142,4 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-haiku-4-5-20251001:orden` | 3 de 3 (100,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 1 de 33 (3,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-sonnet-5-5:herramienta` | 0 de 3 (0,0 %) | ≤ 0,0 % | sí | sí |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 41 de 33 (124,2 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-haiku-4-5-20251001:herramienta` | 2 de 3 (66,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

**Serie de cierres de boe-legislacion** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `27aeccc` (árbol de `skills/boe-legislacion`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

`afirma_lo_no_leido`:

| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |
|---|---|---|---|---|---|---|---|---|---|
| H24-1 | 2026-10-06 | `a5bda45` | `6d65aa3` | 1/54 | 0/54 | 1/78 | 0/30 | 1/108 | no: otra skill |
| H24-2 | 2026-10-06 | `b0a7a9c` | `75e1d04` | 0/54 | 3/54 | 0/78 | 3/30 | 3/108 | no: 1 respuesta sin juzgar |
| H24-3 | 2026-10-06 | `3107ae2` | `6063317` | 0/54 | 0/54 | 0/78 | 0/30 | 0/108 | no: otra skill |
| H23-1 | 2026-10-07 | `bf4befd` | `6063317` | 0/54 | 0/50 | 0/76 | 0/28 | 0/104 | no: otra skill |
| H25-1 | 2026-10-10 | `3f99446` | `6063317` | 0/54 | 1/54 | 0/78 | 1/30 | 1/108 | no: otra skill |
| H25-2 | 2026-10-10 | `e380f0a` | `6063317` | 1/54 | 1/54 | 0/78 | 2/30 | 2/108 | no: otra skill |
| H25-3 | 2026-10-10 | `d51f3d7` | `eed129f` | 0/54 | 2/54 | 0/78 | 2/30 | 2/108 | no: otra skill |
| H26-1 | 2026-10-13 | `b29fe6c` | `eed129f` | 0/54 | 2/54 | 1/78 | 1/30 | 2/108 | no: otra skill |
| H23.1-1 | 2026-10-16 | `60e8678` | `eed129f` | 0/54 | 3/54 | 1/78 | 2/30 | 3/108 | no: otra skill |
| H9-1 | 2026-10-18 | `3f14e3a` | `eed129f` | 0/54 | 1/54 | 0/78 | 1/30 | 1/108 | no: otra skill |
| H9-2 | 2026-10-18 | `92fc232` | `eed129f` | 0/54 | 0/54 | 0/78 | 0/30 | 0/108 | no: otra skill |
| H20-1 | 2026-10-19 | `77211a9` | `49b59ac` | 0/60 | 2/60 | 0/90 | 2/30 | 2/120 | no: otra skill |
| H26.1-1 | 2026-10-21 | `444e674` | `27aeccc` | 0/60 | 1/60 | 0/90 | 1/30 | 1/120 | sí |

Tasa acumulada con la skill como está: 1 de 120 (0,83 %), de H26.1-1.
Insuficiente: 120 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

**Serie de cierres de jurisprudencia** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `3bbd9cd` (árbol de `skills/jurisprudencia`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

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
| H26.1-1 | 2026-10-21 | `444e674` | `3bbd9cd` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |

Tasa acumulada con la skill como está: 0 de 60 (0,00 %), de H26.1-1.
Insuficiente: 60 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

## 4. Revisión que la constitución reserva a la persona

- Fixture o esquema EXISTENTE modificado: `internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_robots.txt.json`.
- Fixture o esquema EXISTENTE modificado: `internal/evals/testdata/grabaciones.json`.
- Fixture o esquema EXISTENTE modificado: `schemas/eval.yaml.json`.
- Fixtures nuevos (4), por directorio:
  - `internal/evals/testdata/boe.legislacion-consolidada/`: `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2004-4214_texto_bloque_a15.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2006-20764_texto_bloque_a18.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10566_texto_bloque_a26.json`, `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10566_texto_bloque_a27.json`
- Grabaciones del paso grabar_datos: `specs/025-h26-1-las-evals-sirven/gates/grabaciones.md`.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/preparar_test.go:TestLoQueLlevaALeerSinGrabar`, en make ci (verde) |
| FR-002 | T004(hecha) T015(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones`, en make ci (verde) |
| FR-003 | T004(hecha) T015(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/preparar_test.go:TestLoQueLlevaALeerSinGrabar`, en make ci (verde) |
| FR-004 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/medida_test.go:TestBaseDeLaReconstruccion`, en make ci (verde) |
| FR-005 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones`, en make ci (verde) |
| FR-006 | T009(hecha) | — | comprobado por su control | `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-007 | T006(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarLasLecturasSinServir`, en make ci (verde) |
| FR-008 | T003(hecha) | — | comprobado por su control | `ci:internal/evals/formato_test.go:TestFormatoDeLecturas`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-009 | T001(hecha) T003(hecha) | — | comprobado por su control | `ci:internal/evals/formato_test.go:TestFormatoDeLecturas`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-010 | T002(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-011 | T009(hecha) | — | comprobado por su control | `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-012 | T001(hecha) T003(hecha) | — | comprobado por su control | `ci:internal/evals/formato_test.go:TestFormatoDeLecturas`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-013 | T005(hecha) T006(hecha) T008(hecha) T015(hecha) | — | comprobado por su control | `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones`, en make ci (verde) |
| FR-014 | T003(hecha) | — | comprobado por su control | `ci:internal/evals/formato_test.go:TestFormatoDeLecturas`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-015 | T006(hecha) T007(hecha) T013(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarLasLecturasSinServir`, en make ci (verde); `ci:internal/evals/umbrales_test.go:TestUmbralesDeLasLecturas`, en make ci (verde) |
| FR-016 | T003(hecha) T006(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarLasLecturasSinServir`, en make ci (verde) |
| FR-017 | T006(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarLasLecturasSinServir`, en make ci (verde) |
| FR-018 | T006(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarLasLecturasSinServir`, en make ci (verde) |
| FR-019 | T008(hecha) | — | comprobado por su control | `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:legal-core:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 % |
| FR-020 | T008(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-033 | T007(hecha) T008(hecha) T013(hecha) T015(hecha) | — | comprobado por su control | `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `ci:internal/evals/umbrales_test.go:TestUmbralesDeLasLecturas`, en make ci (verde) |
| FR-021 | T010(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralesDeLasLecturas`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills`, en make ci (verde) |
| FR-022 | T006(hecha) T010(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralesDeLasLecturas`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills`, en make ci (verde) |
| FR-023 | T012(hecha) T013(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarElMotivoDeLasFallidas`, en make ci (verde); `ci:internal/evals/sesion_test.go:TestLeerLasLlamadas`, en make ci (verde) |
| FR-024 | T013(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarElMotivoDeLasFallidas`, en make ci (verde); `ci:internal/evals/sesion_test.go:TestLeerLasLlamadas`, en make ci (verde) |
| FR-025 | T012(hecha) T013(hecha) T015(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestLineaDeLaSalidaNoEscrita`, en make ci (verde); `ci:internal/app/tuberia_unix_test.go:TestTuberiaCerrada`, en make ci (verde) |
| FR-026 | T012(hecha) T013(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarElMotivoDeLasFallidas`, en make ci (verde); `ci:internal/evals/sesion_test.go:TestLeerLasLlamadas`, en make ci (verde) |
| FR-027 | T013(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarElMotivoDeLasFallidas`, en make ci (verde); `ci:internal/evals/sesion_test.go:TestLeerLasLlamadas`, en make ci (verde) |
| FR-028 | T013(hecha) T016(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarElMotivoDeLasFallidas`, en make ci (verde); `ci:internal/evals/sesion_test.go:TestLeerLasLlamadas`, en make ci (verde) |
| FR-029 | T011(hecha) | — | comprobado por su control | `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 % |
| FR-030 | T011(hecha) | — | comprobado por su control | `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 % |
| FR-031 | T011(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-032 | T014(hecha) T016(hecha) | — | tareas hechas | — |
| SC-001 | T001(hecha) T016(hecha) | — | tareas hechas | — |
| SC-002 | T016(hecha) | — | comprobado por su control | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:duracion_de_las_sesiones:orden`: 483, ≤ 900; `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`: 568, ≤ 900; `evals:boe-legislacion:duracion_del_juez:orden`: 95, ≤ 900; `evals:boe-legislacion:duracion_del_juez:herramienta`: 91, ≤ 900; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 45, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 41, ≤ 900 |
| SC-003 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/preparar_test.go:TestLoQueLlevaALeerSinGrabar`, en make ci (verde) |
| SC-004 | T003(hecha) T006(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/formato_test.go:TestFormatoDeLecturas`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| SC-005 | T003(hecha) T005(hecha) T006(hecha) T007(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones`, en make ci (verde); `ci:internal/evals/formato_test.go:TestFormatoDeLecturas`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/juzgar_test.go:TestJuzgarLasLecturasSinServir`, en make ci (verde) |
| SC-006 | T010(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralesDeLasLecturas`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills`, en make ci (verde) |
| SC-007 | T012(hecha) T013(hecha) T015(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarElMotivoDeLasFallidas`, en make ci (verde); `ci:internal/evals/sesion_test.go:TestLeerLasLlamadas`, en make ci (verde); `ci:internal/evals/juzgar_test.go:TestLineaDeLaSalidaNoEscrita`, en make ci (verde); `ci:internal/app/tuberia_unix_test.go:TestTuberiaCerrada`, en make ci (verde) |
| SC-008 | T004(hecha) T007(hecha) T008(hecha) T015(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones`, en make ci (verde); `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones`, en make ci (verde); `ci:internal/evals/umbrales_test.go:TestUmbralesDeLasLecturas`, en make ci (verde) |
| SC-009 | T008(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| SC-010 | T014(hecha) | — | tareas hechas | — |
| SC-011 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/preparar_test.go:TestLoQueLlevaALeerSinGrabar`, en make ci (verde) |
| SC-012 | T008(hecha) | — | comprobado por su control | `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:legal-core:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 % |
| SC-013 | T007(hecha) T008(hecha) | — | comprobado por su control | `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `ci:internal/evals/umbrales_test.go:TestUmbralesDeLasLecturas`, en make ci (verde) |
| SC-014 | T008(hecha) | — | comprobado por su control | `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:legal-core:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:legal-core:vigencia_sin_declarar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:legal-core:vigencia_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:legal-core:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 % |
| SC-015 | T010(hecha) | — | comprobado por su control | `ci:internal/evals/umbrales_test.go:TestUmbralesDeLasLecturas`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills`, en make ci (verde) |
| SC-016 | T016(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInforme`, en make ci (verde) |
| SC-017 | T013(hecha) T016(hecha) | — | comprobado por su control | `ci:internal/evals/juzgar_test.go:TestJuzgarElMotivoDeLasFallidas`, en make ci (verde); `ci:internal/evals/sesion_test.go:TestLeerLasLlamadas`, en make ci (verde) |
| SC-018 | T002(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| SC-019 | T010(hecha) T016(hecha) | — | comprobado por su control | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5`: 1 de 120 (0,8 %), ≤ 2,8 % |
| SC-020 | T010(hecha) T016(hecha) | — | comprobado por su control | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 45, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 41, ≤ 900 |
| SC-021 | T011(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`4c23eed`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `007c7fe` fix(H26.1): motivos de la revisión final: lo vio la ronda 2 (juez A: aprobado; juez B: aprobado). Toca `CHANGELOG.md`, `CONTRIBUTING.md`, `internal/evals/consultas.go`, `internal/evals/consultas_test.go`, `internal/evals/informe.go`, `internal/evals/medida.go`, `internal/evals/medida_test.go`, `cierre.md`, `contracts/cache-de-la-sesion.md`, `data-model.md`, `plan.md`, `research.md`.
- `b07dd5c` docs(H26.1): veredictos de la revisión final: solo registros de `gates/`.
- `444e674` docs(H26.1): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/025-h26-1-las-evals-sirven/quickstart.md`. Suite de aceptación congelada: `specs/025-h26-1-las-evals-sirven/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `e9872cde`: 15 h 13 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh e9872cde`.
