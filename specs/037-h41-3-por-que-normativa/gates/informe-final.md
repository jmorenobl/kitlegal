# Informe del hito H41.3 · Por qué `normativa` no se activa con todas las skills instaladas, y si unas descripciones cortas lo arreglan

Generado por el workflow `hito` el 2026-11-16T16:42:56Z, sobre `6701b623` de `037-h41-3-por-que-normativa`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/037-h41-3-por-que-normativa/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige el fichero del hito en docs/roadmap/in-progress/ y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `6701b62` (medición 1): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: consulta-juridica aprobado, jurisprudencia aprobado, normativa aprobado, plazos-y-territorio aprobado (tasas en la sección 3).
- **Umbrales del job**: consulta-juridica: 4 umbrales, ninguno decide y 4 solo se publican (4 por encima de su umbral: `fuera_de_lo_grabado:claude-sonnet-5-5:orden`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); jurisprudencia: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (2 por encima de su umbral: `afirma_que_existe:claude-sonnet-5-5:orden`, `afirma_que_existe:claude-sonnet-5-5:herramienta`); normativa: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (5 por encima de su umbral: `cuenta_su_proceso:claude-sonnet-5-5:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:orden`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); plazos-y-territorio: 18 umbrales, los 12 que deciden se cumplen y 6 solo se publican (2 por encima de su umbral: `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `lectura_sin_intentar:claude-haiku-4-5-20251001:herramienta`) (sección 3).
- **Revisión final**: juez A aprobado (con `claude-opus-5-5`), juez B aprobado (con `claude-fable-5-2`), 2 rondas.
- **Sesiones con otro modelo que el de su rol**: ninguna (36 sesiones registradas, todas con el modelo que `workflow.yml` da a su rol).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 11 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 19 commits; 117 files changed, 6480 insertions(+), 206 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- corrector_spec: [a] FR-016 y el supuesto cambiaban el mensaje de hoy de un `INSTALACION` que no vale («INSTALACION: «todas» no es producto») en una petición sin las dos opciones → lectura conservadora (constitución, «Criterio de decisión autónoma», puntos 2 y 4): ese mensaje queda como hoy, byte a byte, y FR-015, FR-018, SC-004 y el supuesto dicen que una petición sin las dos opciones que no vale da los mensajes de error de hoy con los valores esperados de los tests de hoy sin cambiar; el hito p… (entera en `specs/037-h41-3-por-que-normativa/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_spec: [h][f] la forma de `PRESUPUESTO` tenía dos lecturas (ceros delante, signo) → FR-013 fija solo cifras, sin signo ni ceros delante, la forma de los demás enteros de argumento del arnés (`^[1-9][0-9]*$`, internal/evals/sondeo.go:34), y FR-016 y US2 escenario 5 hacen error de argumentos lo demás (`0`, `0100`, `+1000000`); así «tal como se pidió» y «el valor pedido» tienen una sola lectura.
- corrector_spec: [h][f] FR-013 no decía si vale un `PRESUPUESTO` de solo cifras que no cabe en el entero de los demás argumentos (`99999999999999999999`) → lectura conservadora (constitución, «Criterio de decisión autónoma», puntos 2 y 4), la regla de hoy: `PRESUPUESTO` tiene la forma y el rango de `SESIONES` (cabe en un `int`; `sesionesDeCadaEval`, internal/evals/laboratorio_sondeo.go:607-613), y un valor que no cabe es error de argumentos de FR-016, con un código distinto de 0 y ninguna sesión;… (entera en `specs/037-h41-3-por-que-normativa/gates/supuestos.md` o `clarify-respuestas.json`)
- plan: FR-001 deja al plan dónde vive la variante y con qué forma → `laboratorio/producto-descripciones-cortas/<skill>.txt`, un fichero de una línea por skill de `skills/`, cuyo texto es el fichero sin su salto de línea final, la regla de los textos de las variantes de H39 (research D1). Descartados: un YAML con una clave por skill, que obliga a tratar claves repetidas y a citar `:` y `#` dos veces; una copia de `skills/`, que guarda cuerpos y referencias que FR-001 excluye.
- plan: FR-010 y FR-016 dejan al plan los mensajes de `INSTALACION=producto-descripciones-cortas` → `INSTALACION: producto-descripciones-cortas solo vale con VARIANTES=normativa-producto` con cualquier variante que vale y no es `normativa-producto`, también en una mezcla; con la familia de H39, el mensaje de hoy; y un `INSTALACION: <defecto>` por cada defecto de la variante, que solo se lee con ese valor (research D6). Descartado: dar el primero solo con familia, como `producto`, con lo que la mez… (entera en `specs/037-h41-3-por-que-normativa/gates/supuestos.md` o `clarify-respuestas.json`)
- plan: FR-013 deja al plan cómo llega `PRESUPUESTO` y sus mensajes → décimo argumento de la receta y de `scripts/evals-laboratorio.sh`, `-presupuesto` en `go test`; `PRESUPUESTO: «<v>» no es un entero mayor que 0`, el texto de `SESIONES` con `EVAL`, y con la familia de H39 el mismo de `INSTALACION`; una línea, la última, detrás de la de `EVAL` (research D7). Descartado: tomarlo del entorno de quien lanza, que FR-014 prohíbe.
- plan: FR-012 deja al plan las claves de la constancia → `instalacion` (cadena) y `presupuesto` (número), con `omitempty`, detrás de `skills_instaladas`, cada una solo con su opción; sin las dos, `sondeo.json` como hoy (research D9).
- plan: FR-020 no fija dónde van las dos columnas ni el orden de las filas → detrás de «Skills instaladas»; filas por variante, modo, modelo, versión, instalación, presupuesto y skills instaladas, con la instalación y el presupuesto en la clave solo si algún sondeo lleva la constancia: las cuatro filas del sondeo salen en el orden de FR-021 y las tablas de hoy, byte a byte (research D10). Descartado: ordenarlas detrás de las skills instaladas, que daría la fila `sola` la última.
- plan: FR-022 dice «si sus `skills_instaladas` son las cuatro del producto» → `producto` si tienen alguna skill además de la de la familia, que es lo que deja `INSTALACION=producto` también en un árbol con otra skill más (research D11). Descartado: compararlas con una lista fija de cuatro nombres.
- T009: para saber si la tabla lleva «Instalación» y «Presupuesto», `algunoConLaConstancia` lee los `sondeo.json` de la ruta, en su orden, hasta tener las dos constancias, y no ya solo hasta el primero con `skills_instaladas`; con dos defectos a la vez —una sesión que no sabe leer en un sondeo anterior y un `sondeo.json` que no sabe leer en uno posterior—, el que se da puede ser el del `sondeo.json` → se acepta: la cabecera depende de todos los `sondeo.json` antes de contar ninguna sesión, el cont… (entera en `specs/037-h41-3-por-que-normativa/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q1 (criterio d, conservadora): `INSTALACION=producto-descripciones-cortas` solo vale con `VARIANTES=normativa-producto`, la variante que nombra el hito y la única que usa el sondeo de H41.3. Cualquier otra variante en `VARIANTES` es un error de argumentos de FR-016, sola o junto a `normativa-producto`: las versionadas de `normativa` (`normativa-control`…), las `<skill>-producto` y las versionadas de las otras dos familias del producto, y las de H39. El mensaje nombra `INSTALACION` y dice… (entera en `specs/037-h41-3-por-que-normativa/gates/supuestos.md` o `clarify-respuestas.json`)

**Alcance (lo que queda fuera o dentro del hito)**

- plan: FR-030 y FR-031 piden el sondeo de cuatro filas y 640 sesiones y el ADR que lee su tabla, y el run no puede abrir sesiones con modelo ni usar la credencial del sondeo (FR-017; ADR 0032) → el run no lanza el sondeo ni escribe el ADR: la propuesta de cambio queda pendiente de la tabla en `evidencias/`, con el modelo, la versión de Claude Code, la instalación y el presupuesto de cada fila, y de su lectura en la propuesta, las dos de mano de la persona, que no deciden la fusión (quickstart.md … (entera en `specs/037-h41-3-por-que-normativa/gates/supuestos.md` o `clarify-respuestas.json`)
- reparador T001: la verificación de T001 falla en `TestTopeDeLaSesion/muere-por-otra-senal` (`0` en lugar de `129`), un test de internal/evals que la tarea no toca; el fallo se reproduce byte a byte con el proceso de test lanzado con SIGHUP ignorada, que la sesión hereda, y no se ha podido medir la disposición de SIGHUP del verificador (gates/tarea-T001.md) → T001 queda redelimitada y declara además internal/evals/sesiones_test.go, solo para que el test escuche SIGHUP mientras corre y la sesión l… (entera en `specs/037-h41-3-por-que-normativa/gates/supuestos.md` o `clarify-respuestas.json`)

**Skill (lo que pide, dice o comprueba una skill)**

- plan: FR-001 deja al plan los cuatro textos → los de contracts/variante.md §2, de 417, 420, 426 y 427 runas (1690 de 1700), cada uno con lo que conserva y lo que deja de su `description` de hoy; se conservan siempre la frase que ata la skill a su herramienta, «aunque el agente crea saberlo» y, en `consulta-juridica`, «antes que las demás skills de kitlegal», para que el remedio mida la longitud y no otro cambio (research D2). Descartado: recortar cada texto por igual, que cortaría frases a media… (entera en `specs/037-h41-3-por-que-normativa/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run


- Observaciones de los jueces, por debajo del umbral y sin corregir: 11 en `spec-r3.json`, 3 en `plan-r2.json`, 6 en `tasks-r1.json`, 6 en `revision-a-r2.json`, 11 en `revision-b-r2.json`.

### Internos (13)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T003 (1), T004 (3), T005 (2), T006 (1), T008 (1), corrector_plan (2), corrector_spec (1), plan (2). Enteras en `specs/037-h41-3-por-que-normativa/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**consulta-juridica**: aprobado sobre `6701b62`; decide `claude-sonnet-5-5` con 2 de 3; 2 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-poliza-tras-el-fallecimiento.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `02-una-sola-capacidad.yaml` | 3/3 | 0/3 | 3/3 | 1/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 42 de 6 (700,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 10 de 6 (166,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 16 de 6 (266,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 7 de 6 (116,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

**jurisprudencia**: aprobado sobre `6701b62`; decide `claude-sonnet-5-5` con 2 de 3; 10 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `02-resumen-sin-documento.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `03-por-materia.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
| `04-documento-pegado.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-documento-que-no-es-el-pedido.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `07-resumen-de-una-conocida.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `08-doctrina-con-el-fallo-delante.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `09-de-que-trata-con-la-ficha-sola.yaml` | 3/3 | 0/3 | 3/3 | 2/3 |  |
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
| `duracion_del_juez:orden` | 56 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 53 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |

**normativa**: aprobado sobre `6701b62`; decide `claude-sonnet-5-5` con 2 de 3; 23 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 1/3 | 3/3 | 2/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 0/3 | 3/3 | 1/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
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
| `22-lcsp-contrato-menor-en-2019.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `23-lpac-articulo-21-en-2016.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 4 de 60 (6,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 120 (0,0 %) | ≤ 2,8 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 535 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 527 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 109 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 91 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 3 de 66 (4,5 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 27 de 42 (64,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 4 de 66 (6,1 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 25 de 42 (59,5 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

**plazos-y-territorio**: aprobado sobre `6701b62`; decide `claude-sonnet-5-5` con 2 de 3; 12 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 0/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |
| `05-plazo-reposicion-leganes.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `06-plazo-reposicion-tordesillas.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `07-cobertura-liquidacion-tributaria.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `08-cobertura-multa-de-trafico.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `09-plazo-contencioso-agosto.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `10-plazo-reposicion-anterior-a-la-lpac.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `11-plazo-reposicion-sin-leer-el-precepto.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |

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
| `lectura_sin_intentar:claude-haiku-4-5-20251001:herramienta` | 2 de 3 (66,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

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
| H41.1-1 | 2026-11-10 | `e4915f6` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H41.2-1 | 2026-11-11 | `a27c276` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H41.2-2 | 2026-11-11 | `899708c` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H41.2-3 | 2026-11-11 | `e8715a9` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H41.3-1 | 2026-11-16 | `6701b62` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |

Tasa acumulada con la skill como está: 0 de 420 (0,00 %), de H41-1, H41-2, H41.1-1, H41.2-1, H41.2-2, H41.2-3 y H41.3-1.
Insuficiente: 420 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

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
| H41.1-1 | 2026-11-10 | `e4915f6` | `26368b3` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | sí |
| H41.2-1 | 2026-11-11 | `a27c276` | `26368b3` | 0/60 | 1/60 | 1/90 | 0/30 | 1/120 | sí |
| H41.2-2 | 2026-11-11 | `899708c` | `26368b3` | 1/60 | 0/60 | 1/90 | 0/30 | 1/120 | sí |
| H41.2-3 | 2026-11-11 | `e8715a9` | `26368b3` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | sí |
| H41.3-1 | 2026-11-16 | `6701b62` | `26368b3` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | sí |

Tasa acumulada con la skill como está: 2 de 840 (0,24 %), de H41-1, H41-2, H41.1-1, H41.2-1, H41.2-2, H41.2-3 y H41.3-1.
Escalón de la regla 3 con 120 respuestas, el total de hoy: como mucho 2 marcadas (vale hasta una tasa del 0,48 %). El umbral de hoy, 0.028 en `evals/normativa/juez/clases.yaml`, admite 3 de 120.
Permite estrechar el umbral a 2 de 120. Lo baja una persona, con un diff de `evals/normativa/juez/clases.yaml` y estas cifras (ADR 0038, punto 4); ningún run lo cambia.

## 4. Revisión que la constitución reserva a la persona

- Fixtures nuevos (39), por directorio:
  - `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/h41-2/normativa-producto/herramienta/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/h41-2/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/h41-3-control/normativa-producto/herramienta/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/h41-3-control/normativa-producto/herramienta/04-lgt-prescripcion-02/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/h41-3-control/normativa-producto/herramienta/04-lgt-prescripcion-03/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/h41-3-control/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/h41-3-diagnostico/normativa-producto/herramienta/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/h41-3-diagnostico/normativa-producto/herramienta/04-lgt-prescripcion-02/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/h41-3-diagnostico/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/h41-3-remedio/normativa-producto/herramienta/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/h41-3-remedio/normativa-producto/herramienta/04-lgt-prescripcion-02/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/h41-3-remedio/normativa-producto/herramienta/04-lgt-prescripcion-03/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/h41-3-remedio/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/h41-3-sola/normativa-producto/herramienta/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/h41-3-sola/normativa-producto/herramienta/04-lgt-prescripcion-02/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/h41-3-sola/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/instalacion/`: `tabla.md`
- Suite de aceptación activada y congelada (1 guiones, escritos desde el spec en la primera tarea): `h41-3-laboratorio-tabla.txtar`.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T003(hecha) T011(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_descripciones_test.go:TestDescripcionesCortas`, en make ci (verde) |
| FR-002 | T002(hecha) T003(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_descripciones_test.go:TestDescripcionesCortas`, en make ci (verde) |
| FR-010 | T006(hecha) T011(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLasDescripcionesCortas`, en make ci (verde); `ci:internal/evals/laboratorio_descripciones_test.go:TestConLaDescripcion`, en make ci (verde) |
| FR-011 | T004(hecha) T006(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLasDescripcionesCortas`, en make ci (verde); `ci:internal/evals/laboratorio_descripciones_test.go:TestConLaDescripcion`, en make ci (verde) |
| FR-012 | T006(hecha) T007(hecha) T011(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDelLaboratorio`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLaInstalacionDelProducto`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConUnaEval`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConElPresupuesto`, en make ci (verde) |
| FR-013 | T007(hecha) T008(hecha) T011(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConElPresupuesto`, en make ci (verde) |
| FR-014 | T005(hecha) T007(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConElPresupuesto`, en make ci (verde) |
| FR-015 | T005(hecha) T006(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDelLaboratorio`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLaInstalacionDelProducto`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConUnaEval`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConElPresupuesto`, en make ci (verde) |
| FR-016 | T006(hecha) T007(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLasDescripcionesCortas`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConElPresupuesto`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLaInstalacionDelProducto`, en make ci (verde) |
| FR-017 | T008(hecha) | — | tareas hechas | — |
| FR-018 | T005(hecha) T006(hecha) T007(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLasDescripcionesCortas`, en make ci (verde); `ci:internal/evals/laboratorio_descripciones_test.go:TestConLaDescripcion`, en make ci (verde) |
| FR-020 | T001(hecha) T009(hecha) T010(hecha) T011(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaConLaInstalacion`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-021 | T001(hecha) T009(hecha) T010(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaConLaInstalacion`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-022 | T001(hecha) T009(hecha) T010(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaConLasActivaciones`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDelLaboratorio`, en make ci (verde) |
| FR-023 | T009(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaConLasActivaciones`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDelLaboratorio`, en make ci (verde) |
| FR-024 | T001(hecha) T010(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaConLaInstalacion`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-030 | — | — | SIN TAREA | — |
| FR-031 | T011(hecha) | — | tareas hechas | — |
| FR-032 | T009(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaConLasActivaciones`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDelLaboratorio`, en make ci (verde) |
| FR-033 | T001(hecha) T003(hecha) | — | tareas hechas | — |
| FR-034 | — | — | SIN TAREA | `evals:normativa:sin_activar:claude-sonnet-5-5:orden`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 120 (0,0 %), ≤ 2,8 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:normativa:duracion_de_las_sesiones:orden`: 535, ≤ 900; `evals:normativa:duracion_de_las_sesiones:herramienta`: 527, ≤ 900; `evals:normativa:duracion_del_juez:orden`: 109, ≤ 900; `evals:normativa:duracion_del_juez:herramienta`: 91, ≤ 900; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 56, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 53, ≤ 900; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 % |
| SC-001 | T003(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_descripciones_test.go:TestDescripcionesCortas`, en make ci (verde) |
| SC-002 | T004(hecha) T006(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLasDescripcionesCortas`, en make ci (verde); `ci:internal/evals/laboratorio_descripciones_test.go:TestConLaDescripcion`, en make ci (verde) |
| SC-003 | T005(hecha) T007(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConElPresupuesto`, en make ci (verde) |
| SC-004 | T005(hecha) T006(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDelLaboratorio`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLaInstalacionDelProducto`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConUnaEval`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConElPresupuesto`, en make ci (verde) |
| SC-005 | T006(hecha) T007(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLasDescripcionesCortas`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConElPresupuesto`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLaInstalacionDelProducto`, en make ci (verde) |
| SC-006 | T001(hecha) T009(hecha) T010(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaConLaInstalacion`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-007 | T009(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaConLasActivaciones`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDelLaboratorio`, en make ci (verde) |
| SC-008 | — | — | SIN TAREA | `evals:normativa:sin_activar:claude-sonnet-5-5:orden`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 120 (0,0 %), ≤ 2,8 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:normativa:duracion_de_las_sesiones:orden`: 535, ≤ 900; `evals:normativa:duracion_de_las_sesiones:herramienta`: 527, ≤ 900; `evals:normativa:duracion_del_juez:orden`: 109, ≤ 900; `evals:normativa:duracion_del_juez:herramienta`: 91, ≤ 900; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 56, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 53, ≤ 900; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 % |
| SC-009 | — | — | SIN TAREA | — |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`a53bf9a`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `b0d7e159` fix(H41.3): motivos de la revisión final: lo vio la ronda 2 (juez A: aprobado; juez B: aprobado). Toca `internal/evals/laboratorio_tabla.go`, `internal/evals/laboratorio_tabla_test.go`.
- `58397ad9` docs(H41.3): veredictos de la revisión final: solo registros de `gates/`.
- `6701b623` docs(H41.3): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/037-h41-3-por-que-normativa/quickstart.md`. Suite de aceptación congelada: `specs/037-h41-3-por-que-normativa/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `8da230a5`: 6 h 53 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh 8da230a5`.
