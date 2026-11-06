# Informe del hito H41 · Las skills del producto mandan leer sus referencias

Generado por el workflow `hito` el 2026-11-06T07:02:50Z, sobre `1e4ef0fd` de `033-h41-las-skills-del`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/033-h41-las-skills-del/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige el fichero del hito en docs/roadmap/in-progress/ y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `1e4ef0f` (medición 1): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: consulta-juridica aprobado, jurisprudencia aprobado, normativa aprobado, plazos-y-territorio aprobado (tasas en la sección 3).
- **Umbrales del job**: consulta-juridica: 4 umbrales, ninguno decide y 4 solo se publican (4 por encima de su umbral: `fuera_de_lo_grabado:claude-sonnet-5-5:orden`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); jurisprudencia: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (2 por encima de su umbral: `afirma_que_existe:claude-sonnet-5-5:orden`, `afirma_que_existe:claude-sonnet-5-5:herramienta`); normativa: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (5 por encima de su umbral: `cuenta_su_proceso:claude-sonnet-5-5:orden`, `cuenta_su_proceso:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); plazos-y-territorio: 18 umbrales, los 12 que deciden se cumplen y 6 solo se publican, todos dentro de su umbral (sección 3).
- **Revisión final**: juez A aprobado (con `claude-opus-5-5`), juez B aprobado (con `claude-fable-5-1`), 1 ronda.
- **Sesiones con otro modelo que el de su rol**: ninguna (29 sesiones registradas, todas con el modelo que `workflow.yml` da a su rol).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 9 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 15 commits; 164 files changed, 7812 insertions(+), 426 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- corrector_spec: `make laboratorio-tabla` con un `DIRECTORIO` que junta sondeos de más de una familia (FR-040) → lectura conservadora de H40 (contracts/tabla.md §1): una tabla cuenta una sola familia, la de la primera variante que lee en orden de nombre, y un directorio de variante de otra familia, también de H39, es un defecto que nombra su ruta, con código 1 y sin tabla; la aceptación da una tabla por familia, con trece filas entre las tres (US5-1, SC-010, FR-052, «Uso»). No se implementa una t… (entera en `specs/033-h41-las-skills-del/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_spec: `SESIONES=0` era un múltiplo de 4 válido (FR-030, US3-4) → error de argumentos, como en H40: un múltiplo de 4 mayor que 0; los demás argumentos y la versión de Claude Code se comprueban como en H40 (contracts/arnes.md §2).
- corrector_plan: `laboratorio tabla` tenía un mensaje y un caso de `TestTablaDeLasFamilias` propios para un `sondeo.json` sin `referencias_del_producto`, que solo se alcanza editándolo a mano porque el arnés la escribe siempre que instala `<skill>-producto` → se retiran el mensaje y el caso; queda la regla genérica de lo que el guion no sabe leer, como en H39: defecto `inesperado`, código 1 y sin tabla (ADR 0023; contracts/tabla.md §1 y §5; research D14; plan.md, «Trazabilidad» y fila V).
- clarify Q2 (criterio c): A, con la lista de nombres y no con la copia. Al instalar `<skill>-producto` desde `skills/` en el temporal, el arnés escribe en el directorio del sondeo (`DIRECTORIO`) los nombres de los ficheros de `references/` de la skill que instaló, en orden de nombre. `make laboratorio-tabla` toma de ahí las columnas de `<skill>-producto`. El plan fija el fichero y su forma, sin cambiar lo que los guiones de H39 y H40 leen de `sondeo.json`, para que sus suites congeladas sigan en … (entera en `specs/033-h41-las-skills-del/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q3 (criterio c): B. Tras una orden simple `cd` a un directorio cuya ruta termina en `<skill>/references`, con o sin barra final, una orden simple posterior de la misma `input.command` abre `<f>` si alguna de sus palabras es exactamente `<f>` o `./<f>`. Las palabras se parten como en sh, con las comillas simples y dobles y la barra invertida, y las órdenes simples por `|`, `||`, `&&`, `;`, `&` y saltos de línea: es el mismo lector que el guion ya usa para la primera herramienta en modo or… (entera en `specs/033-h41-las-skills-del/gates/supuestos.md` o `clarify-respuestas.json`)

**Alcance (lo que queda fuera o dentro del hito)**

- corrector_spec: las instrucciones de `kitlegal mcp serve` y las descripciones de sus herramientas, que en modo herramienta comparten `control` y `producto` → a «Fuera de alcance»: el hito no pide tocarlas y cambiarlas movería las dos variantes a la vez.
- corrector_spec: cambiar las evals que ya existen, también las cuatro de cada familia que elige el sondeo (FR-030) → a «Fuera de alcance»: el hito solo excluía las nuevas; el sondeo las usa como están.
- tasks: FR-052 —la aceptación de la persona con el sondeo de FR-032, trece filas y 520 sesiones con `claude-sonnet-5-5`— abre sesiones con modelo y la credencial del sondeo, que un run no tiene ni puede usar (FR-031; ADR 0032) → el run no lanza el sondeo de las trece filas ni escribe su ADR. La fusión de la propuesta de cambio queda pendiente de las tres tablas, una por familia, en `evidencias/` con el modelo y la versión de Claude Code anotados, y de su lectura en la propuesta (las órdenes, en q… (entera en `specs/033-h41-las-skills-del/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q1 (criterio d, conservadora): B. La regla ampliada de FR-041 se aplica solo a las familias de las skills del producto (`normativa`, `plazos-y-territorio` y `jurisprudencia`), cuyo contrato es el de H40 (`specs/032-h40-por-qu-e/contracts/tabla.md` §2), que es el que amplía el hito. La familia del laboratorio de H39 (`articulos-del-boe`) se sigue contando exactamente con su contrato (`specs/031-h39-c-omo-se/contracts/tabla.md` §2) y con su comportamiento de hoy, byte a byte, como ya fijó … (entera en `specs/033-h41-las-skills-del/gates/supuestos.md` o `clarify-respuestas.json`)

**Skill (lo que pide, dice o comprueba una skill)**

- corrector_plan: el control de SC-001 (`TestRemisionesDelRepositorio`) buscaba «si la tienes», «antes de escribir la respuesta» y «antes de dar la consulta» sin decir nada de mayúsculas, y no ponía en rojo «Antes de dar la consulta, lee […]» al principio de la frase de `jurisprudencia` → las tres expresiones se buscan sin distinguir mayúsculas de minúsculas, como «lee» o «Lee» en el verbo, y el mensaje nombra la expresión en minúscula; el subtest «control» la ve fallar en `jurisprudencia-control`… (entera en `specs/033-h41-las-skills-del/gates/supuestos.md` o `clarify-respuestas.json`)
- T003: contracts/remisiones.md da a `normativa` un cuerpo de 10 981 caracteres, pero la vertical de ensayo de H32 (`TestVerticalDeEnsayo`, suite activa que protege el principio X) añade su línea de 70 caracteres al bloque de los paquetes de las tres skills de capacidad, y el cuerpo pasa a 11 051 (`normativa: SKILL.md: cuerpo de 11051 caracteres (máximo 11000)`): sin paquetes cabe; con el primero, no → un segundo movimiento de FR-009, el que menos cambia: « Si una lectura falla, di qué bloques fal… (entera en `specs/033-h41-las-skills-del/gates/supuestos.md` o `clarify-respuestas.json`)
- T004: contracts/remisiones.md §4 y la línea de T004 dan a `plazos-y-territorio` un cuerpo de 10 982 caracteres, sin movimientos y sin cambiar ninguna referencia, pero con la línea de 70 caracteres de la vertical de ensayo de H32 (`TestVerticalDeEnsayo`, suite activa) el cuerpo pasa a 11 052 (`plazos-y-territorio: SKILL.md: cuerpo de 11052 caracteres (máximo 11000)`, medido en T004), como ya anunciaba la corrección de T003 → un movimiento de FR-009, como el segundo de `normativa`: « Si sale con 4… (entera en `specs/033-h41-las-skills-del/gates/supuestos.md` o `clarify-respuestas.json`)
- T009: P4, la orden de leer `references/calculo_de_plazos.md` en el paso 10 si la lectura sale con 4 o 5 (111 caracteres), no cabe: el cuerpo de `plazos-y-territorio` pasa a 10 984 y, con la línea de 70 de la vertical de ensayo de H32, a 11 054 (tope 10 930 sin paquetes); y en el cuerpo no queda nada que solo necesite el caso de P4, así que el movimiento que hace sitio es de otro caso, a una de las referencias que la tarea deja elegir «que el cuerpo manda leer en el paso que lo usa». `jerarquia_n… (entera en `specs/033-h41-las-skills-del/gates/supuestos.md` o `clarify-respuestas.json`)
- barrido: FR-003, «en el paso que usa la referencia y antes de la orden o llamada de ese paso» → cuatro órdenes van en su paso pero detrás de la orden o llamada cuyo resultado trae su condición: en `normativa`, las de `redaccion_modificada.md`, `como_se_cita.md` y `lecturas_que_fallan.md` en el paso 2, y en `plazos-y-territorio`, P4 (`calculo_de_plazos.md`) en el paso 10. Research D1 y docs/SKILLS.md §1 («Cómo se remite») lo leen como admitido por FR-003 («en el paso que usa la referencia»); docs… (entera en `specs/033-h41-las-skills-del/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run


- Observaciones de los jueces, por debajo del umbral y sin corregir: 17 en `spec-r2.json`, 7 en `plan-r2.json`, 4 en `tasks-r2.json`, 7 en `revision-a-r1.json`, 8 en `revision-b-r1.json`.

### Internos (7)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T002 (3), T003 (1), T008 (2), barrido (1). Enteras en `specs/033-h41-las-skills-del/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**consulta-juridica**: aprobado sobre `1e4ef0f`; decide `claude-sonnet-5-5` con 2 de 3; 2 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-poliza-tras-el-fallecimiento.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `02-una-sola-capacidad.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 25 de 6 (416,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 11 de 6 (183,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 20 de 6 (333,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 4 de 6 (66,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

**jurisprudencia**: aprobado sobre `1e4ef0f`; decide `claude-sonnet-5-5` con 2 de 3; 10 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `02-resumen-sin-documento.yaml` | 3/3 | 1/3 | 3/3 | 3/3 |  |
| `03-por-materia.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
| `04-documento-pegado.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-documento-que-no-es-el-pedido.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `07-resumen-de-una-conocida.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `08-doctrina-con-el-fallo-delante.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `09-de-que-trata-con-la-ficha-sola.yaml` | 3/3 | 0/3 | 3/3 | 1/3 |  |
| `10-doctrina-dada-por-hecha.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |

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
| `duracion_del_juez:orden` | 53 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 52 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |

**normativa**: aprobado sobre `1e4ef0f`; decide `claude-sonnet-5-5` con 2 de 3; 23 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 3/3 | 3/3 | 0/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
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
| `22-lcsp-contrato-menor-en-2019.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
| `23-lpac-articulo-21-en-2016.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 1 de 60 (1,7 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 3 de 60 (5,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 3 de 60 (5,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 1 de 120 (0,8 %) | ≤ 2,8 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 549 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 548 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 93 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 89 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 66 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 31 de 42 (73,8 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 6 de 66 (9,1 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 34 de 42 (81,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

**plazos-y-territorio**: aprobado sobre `1e4ef0f`; decide `claude-sonnet-5-5` con 2 de 3; 12 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 0/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 2/3 | — | — |  |
| `05-plazo-reposicion-leganes.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `06-plazo-reposicion-tordesillas.yaml` | 3/3 | 1/3 | 3/3 | 3/3 |  |
| `07-cobertura-liquidacion-tributaria.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `08-cobertura-multa-de-trafico.yaml` | 3/3 | 1/3 | 3/3 | 2/3 |  |
| `09-plazo-contencioso-agosto.yaml` | 3/3 | 1/3 | 3/3 | 2/3 |  |
| `10-plazo-reposicion-anterior-a-la-lpac.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `11-plazo-reposicion-sin-leer-el-precepto.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
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
| H41-1 | 2026-11-06 | `1e4ef0f` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |

Tasa acumulada con la skill como está: 0 de 60 (0,00 %), de H41-1.
Insuficiente: 60 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

**Serie de cierres de normativa** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `7a0ccb8` (árbol de `skills/normativa`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

`afirma_lo_no_leido`:

| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |
|---|---|---|---|---|---|---|---|---|---|
| H32-1 | 2026-10-30 | `7f8ef89` | `8566858` | 0/60 | 2/60 | 2/90 | 0/30 | 2/120 | no: otra skill |
| H39-1 | 2026-11-03 | `a593631` | `afbcae6` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H40-1 | 2026-11-04 | `179bd21` | `afbcae6` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H41-1 | 2026-11-06 | `1e4ef0f` | `7a0ccb8` | 1/60 | 0/60 | 0/90 | 1/30 | 1/120 | sí |

Tasa acumulada con la skill como está: 1 de 120 (0,83 %), de H41-1.
Insuficiente: 120 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

## 4. Revisión que la constitución reserva a la persona

- Fixtures nuevos (77), por directorio:
  - `internal/evals/testdata/laboratorio/familias/jurisprudencia/sondeo/h41-a/jurisprudencia-control/herramienta/03-por-materia-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/jurisprudencia/sondeo/h41-a/jurisprudencia-control/orden/01-existe-con-numero-y-fecha-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/jurisprudencia/sondeo/h41-a/jurisprudencia-producto/herramienta/01-existe-con-numero-y-fecha-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/jurisprudencia/sondeo/h41-a/jurisprudencia-producto/herramienta/06-documento-que-no-es-el-pedido-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/jurisprudencia/sondeo/h41-a/jurisprudencia-producto/orden/04-documento-pegado-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/jurisprudencia/sondeo/h41-a/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/jurisprudencia/`: `tabla.md`
  - `internal/evals/testdata/laboratorio/familias/mezcla/sondeo/h41/normativa-control/orden/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/mezcla/sondeo/h41/plazos-y-territorio-control/orden/02-territorio-municipio-no-cubierto-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/mezcla/sondeo/h41/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-control/herramienta/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-control/orden/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-producto/herramienta/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-producto/herramienta/05-trlrhl-impuestos-municipales-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-producto/orden/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-producto/orden/22-lcsp-contrato-menor-en-2019-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-remisiones-en-orden/herramienta/15-irpf-rendimientos-por-materia-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-testigo-en-referencia/herramienta/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/normativa-testigo-en-referencia/herramienta/05-trlrhl-impuestos-municipales-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-a/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-b/normativa-control/orden/05-trlrhl-impuestos-municipales-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/normativa/sondeo/h41-b/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/normativa/`: `tabla.md`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-control/herramienta/10-plazo-reposicion-anterior-a-la-lpac-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-control/orden/02-territorio-municipio-no-cubierto-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-control/orden/05-plazo-reposicion-leganes-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-producto/herramienta/02-territorio-municipio-no-cubierto-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-producto/herramienta/05-plazo-reposicion-leganes-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/plazos-y-territorio-producto/orden/11-plazo-reposicion-sin-leer-el-precepto-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo/h41-a/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/plazos-y-territorio/`: `tabla.md`
- Suite de aceptación activada y congelada (2 guiones, escritos desde el spec en la primera tarea): `h41-laboratorio-tabla.txtar`, `h41-remisiones.txtar`.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T001(hecha) T004(hecha) T009(hecha) | remisiones.txtar  | comprobado por su control | `ci:internal/skills/remisiones_test.go:TestRemisionesDelRepositorio`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-002 | T001(hecha) T004(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/skills/remisiones_test.go:TestRemisionesDelRepositorio`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-003 | T001(hecha) T004(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/skills/remisiones_test.go:TestRemisionesDelRepositorio`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-004 | T001(hecha) T003(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-005 | T001(hecha) T004(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-006 | T001(hecha) T004(hecha) | remisiones.txtar  | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-007 | T003(hecha) T004(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde) |
| FR-008 | T003(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-009 | T001(hecha) T003(hecha) T009(hecha) | remisiones.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-010 | T003(hecha) T004(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-020 | T002(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestVariantesDeControl`, en make ci (verde) |
| FR-021 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_normativa_test.go:TestVariantesDeNormativa`, en make ci (verde) |
| FR-022 | T002(hecha) T005(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_familias_test.go:TestVariantesDeControl`, en make ci (verde); `ci:internal/evals/laboratorio_test.go:TestLaboratorioFueraDelProducto`, en make ci (verde) |
| FR-023 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestLaboratorioFueraDelProducto`, en make ci (verde) |
| FR-030 | T005(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde) |
| FR-031 | — | — | SIN TAREA | — |
| FR-033 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_familias_test.go:TestVariantesDeControl`, en make ci (verde); `ci:internal/evals/laboratorio_test.go:TestLaboratorioFueraDelProducto`, en make ci (verde) |
| FR-032 | T005(hecha) | — | tareas hechas | — |
| FR-040 | T001(hecha) T006(hecha) T007(hecha) T008(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestLecturaConCd`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-041 | T001(hecha) T006(hecha) T007(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestLecturaConCd`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-042 | T006(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-050 | T008(hecha) | — | tareas hechas | — |
| FR-051 | T001(hecha) T008(hecha) | — | comprobado por su control | `evals:normativa:sin_activar:claude-sonnet-5-5:orden`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5`: 1 de 120 (0,8 %), ≤ 2,8 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:normativa:duracion_de_las_sesiones:orden`: 549, ≤ 900; `evals:normativa:duracion_de_las_sesiones:herramienta`: 548, ≤ 900; `evals:normativa:duracion_del_juez:orden`: 93, ≤ 900; `evals:normativa:duracion_del_juez:herramienta`: 89, ≤ 900; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 53, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 52, ≤ 900; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 % |
| FR-052 | T008(hecha) | — | tareas hechas | — |
| SC-001 | T001(hecha) T004(hecha) | remisiones.txtar  | comprobado por su control | `ci:internal/skills/remisiones_test.go:TestRemisionesDelRepositorio`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-002 | T003(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| SC-003 | T003(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| SC-004 | T001(hecha) T003(hecha) T004(hecha) | remisiones.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-005 | T001(hecha) T006(hecha) T007(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestLecturaConCd`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-006 | T002(hecha) T005(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_familias_test.go:TestVariantesDeControl`, en make ci (verde); `ci:internal/evals/laboratorio_test.go:TestLaboratorioFueraDelProducto`, en make ci (verde) |
| SC-007 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde) |
| SC-008 | T002(hecha) T006(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestVariantesDeControl`, en make ci (verde); `ci:internal/evals/laboratorio_normativa_test.go:TestVariantesDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_test.go:TestLaboratorioFueraDelProducto`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-009 | — | — | SIN TAREA | `evals:normativa:sin_activar:claude-sonnet-5-5:orden`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5`: 1 de 120 (0,8 %), ≤ 2,8 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:normativa:duracion_de_las_sesiones:orden`: 549, ≤ 900; `evals:normativa:duracion_de_las_sesiones:herramienta`: 548, ≤ 900; `evals:normativa:duracion_del_juez:orden`: 93, ≤ 900; `evals:normativa:duracion_del_juez:herramienta`: 89, ≤ 900; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 53, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 52, ≤ 900; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 % |
| SC-010 | — | — | SIN TAREA | — |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`58a4c2e`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `1e4ef0fd` docs(H41): veredictos de la revisión final: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/033-h41-las-skills-del/quickstart.md`. Suite de aceptación congelada: `specs/033-h41-las-skills-del/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `0f548c7d`: 8 h 22 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh 0f548c7d`.
