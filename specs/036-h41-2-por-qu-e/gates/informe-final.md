# Informe del hito H41.2 · Por qué la eval 04 no activa `normativa` en modo herramienta en el job

Generado por el workflow `hito` el 2026-11-11T17:22:48Z, sobre `e8715a97` de `036-h41-2-por-qu-e`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/036-h41-2-por-qu-e/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige el fichero del hito en docs/roadmap/in-progress/ y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `e8715a9` (medición 3): ROJO: ci; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: consulta-juridica aprobado, jurisprudencia aprobado, normativa aprobado, plazos-y-territorio aprobado (tasas en la sección 3).
- **Umbrales del job**: consulta-juridica: 4 umbrales, ninguno decide y 4 solo se publican (4 por encima de su umbral: `fuera_de_lo_grabado:claude-sonnet-5-5:orden`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); jurisprudencia: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (2 por encima de su umbral: `afirma_que_existe:claude-sonnet-5-5:orden`, `afirma_que_existe:claude-sonnet-5-5:herramienta`); normativa: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (5 por encima de su umbral: `cuenta_su_proceso:claude-sonnet-5-5:orden`, `cuenta_su_proceso:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); plazos-y-territorio: 18 umbrales, los 12 que deciden se cumplen y 6 solo se publican, todos dentro de su umbral (sección 3).
- **Revisión final**: juez A aprobado (con `claude-opus-5-5`), juez B aprobado (con `claude-fable-5-2`), 3 rondas en 2 ciclos (el primero juzga el hito; cada uno de los siguientes, lo que cambió después del último veredicto, o el hito entero si es un rejuicio).
- **Sesiones con otro modelo que el de su rol**: ninguna (40 sesiones registradas, todas con el modelo que `workflow.yml` da a su rol).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 9 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 21 commits; 203 files changed, 57592 insertions(+), 425 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- corrector_spec: FR-010 (y la 4.ª viñeta de «Assumptions») decía que sin la opción el arnés abre «las mismas sesiones» que hoy, en contradicción con FR-013 en las filas sin la opción del sondeo que llevan el argumento → lo que no cambia sin la opción se acota a la instalación, byte a byte; las sesiones que abre no dependen de la opción: las de hoy sin el argumento de FR-013 y las de FR-013 con él.
- plan: FR-010 y FR-013 dejan al plan el nombre y los valores de la opción y del argumento → `INSTALACION=producto` (vacía o ausente, sin la opción) y `EVAL=<nn>`, el número de dos cifras de una eval del plan de la familia, como `EVALS` de `make evals-sondeo`; los dos detrás de los siete de hoy en la receta y en `scripts/evals-laboratorio.sh`, que pasa a nueve (research D7 a D9). Descartados: `SKILLS=…`, que se lee como el `SKILL` de `evals-sondeo`; el nombre del fichero de la eval; `EVALS` con li… (entera en `specs/036-h41-2-por-qu-e/gates/supuestos.md` o `clarify-respuestas.json`)
- plan: FR-012 pide la constancia de las skills instaladas también sin la opción → todo `sondeo.json` de una familia del producto gana `skills_instaladas` (`["<skill>"]` sin la opción, los cuatro nombres con ella); el de H39 sale como hoy; la tabla da las columnas nuevas solo si algún sondeo la tiene, y la de los sondeos de antes, igual byte a byte (research D11, D12). Descartado: un fichero aparte en `DIRECTORIO`.
- plan: FR-001 dice que en una sesión sin herramienta de kitlegal toda activación es «antes» y el caso límite pone como ejemplo una eval sin binario ni servidor → en esas sesiones ninguna orden cuenta como primera herramienta (research D2). Descartado: contar en ellas una orden `kitlegal` de `Bash` como en el modo orden.
- corrector_plan: contracts/arnes.md §2 callaba, con unas `VARIANTES` sin familia o que mezclan familias, también el error del valor de `INSTALACION`, contra FR-011 («junto a los demás errores, uno por línea») → solo callan los tres errores que dependen de la familia (`INSTALACION=producto` y `EVAL` con variantes de H39, y un `EVAL` fuera del plan de la familia); `INSTALACION: «<v>» no es producto` sale siempre que el valor no es vacío ni `producto`, con cualquier `VARIANTES`, y `TestArnesConLaIns… (entera en `specs/036-h41-2-por-qu-e/gates/supuestos.md` o `clarify-respuestas.json`)
- reparador T002: `make vuln` cae sobre go1.27.3 con GO-2026-6859 (`net/http`), corregida en go1.27.4, un hallazgo que aparece por la fecha y no por el hito → la directiva `toolchain` de `go.mod` sube a go1.27.4 dentro de la rama del hito (el binario se compila con ese parche; `CHANGELOG.md`, «Seguridad»), como manda `CONTRIBUTING.md`, «Subir el parche de Go». Descartado: esperar a una propuesta aparte sobre `main`, con el run sin ninguna tarea en verde hasta entonces. Al cambiar `toolchain`, el c… (entera en `specs/036-h41-2-por-qu-e/gates/supuestos.md` o `clarify-respuestas.json`)
- T004: la sesión cuyo modo no se puede leer (un `servidor.json` del que no se sabe si está) da su lista, como dice contracts/informe.md §2, pero la tarea no fija con qué modo se reconoce su primera herramienta de kitlegal → con el modo que la sesión tiene en el informe, ninguno (`""`), en el que ningún uso es de kitlegal y toda activación es «antes»: es la regla de data-model §1 para «cualquier otro modo», y no se inventa un modo que el directorio no dio.

**Alcance (lo que queda fuera o dentro del hito)**

- corrector_spec: FR-033, US1 y «Uso, de fuera adentro» pedían que el informe final del run diera la lista de activaciones de cada sesión, que el hito no pide y que `scripts/workflow/informe.sh` no publica → lectura conservadora: el informe final da el umbral sin cumplir con sus cifras, como hoy, y la lista está en el `informe.json` y el `informe.md` del job del cierre (FR-001, FR-002). Descartadas: cambiar `scripts/workflow/informe.sh` para que el cuerpo de la propuesta dé las activaciones por se… (entera en `specs/036-h41-2-por-qu-e/gates/supuestos.md` o `clarify-respuestas.json`)
- plan: FR-030 y FR-031 piden el sondeo de cuatro filas y 160 sesiones y el ADR que lee su tabla, y el run no puede abrir sesiones con modelo ni usar la credencial del sondeo (FR-014; ADR 0032) → el run no lanza el sondeo ni escribe el ADR: la propuesta de cambio queda pendiente de la tabla en `evidencias/`, con el modelo, la versión de Claude Code y las skills instaladas de cada fila, y de su lectura en la propuesta, las dos de mano de la persona, que no deciden la fusión (quickstart.md §8).
- T004: `ResultadoDeEval` es también el del juicio del sondeo de `make evals-sondeo` (`juzgarSesionDelSondeo`, internal/evals/sondeo.go), y el hito no dice si ese juicio lleva la lista → no la lleva: el spec, el plan y los contratos solo la piden en el informe del job y en la tabla del laboratorio; `TestJuicioDelSondeo`, que compara el juicio del sondeo con el del job sin lo que sale de la traza, la quita también del del job (gates/tarea-T004.md).
- reparar_cierre: el cierre de `a27c2760` (ejecución 43033002440) sale en rojo por un solo umbral que decide, `sin_activar:claude-sonnet-5-5:herramienta`, 1 de 60: la sesión `04-lgt-prescripcion-herramienta-claude-sonnet-5-5-02` pide `boe_articulo` sin activar ninguna skill (`activaciones: []`, con las cuatro del producto instaladas y Claude Code 2.1.284), la caída que este hito mide y la quinta registrada de la eval 04 en modo herramienta → se aplica FR-033: no se toca ninguna skill, eval, umbral… (entera en `specs/036-h41-2-por-qu-e/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run

- 2026-11-11T15:01:00Z · La revisión final abre el ciclo 2 para juzgar lo que cambió fuera de gates/ después del veredicto sobre 437d493: a68396b «fix(H41.2): cierre en la plataforma».

- Observaciones de los jueces, por debajo del umbral y sin corregir: 11 en `spec-r2.json`, 10 en `plan-r2.json`, 5 en `tasks-r2.json`, 5 en `revision-a-r3.json`, 8 en `revision-b-r3.json`.

### Internos (8)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T003 (3), T004 (1), T007 (1), T008 (1), plan (1), reparar_cierre (1). Enteras en `specs/036-h41-2-por-qu-e/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**consulta-juridica**: aprobado sobre `e8715a9`; decide `claude-sonnet-5-5` con 2 de 3; 2 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-poliza-tras-el-fallecimiento.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `02-una-sola-capacidad.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 38 de 6 (633,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 9 de 6 (150,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 14 de 6 (233,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 4 de 6 (66,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

**jurisprudencia**: aprobado sobre `e8715a9`; decide `claude-sonnet-5-5` con 2 de 3; 10 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-resumen-sin-documento.yaml` | 3/3 | 1/3 | 3/3 | 3/3 |  |
| `03-por-materia.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `04-documento-pegado.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-documento-que-no-es-el-pedido.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `07-resumen-de-una-conocida.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `08-doctrina-con-el-fallo-delante.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `09-de-que-trata-con-la-ficha-sola.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `10-doctrina-dada-por-hecha.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |

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
| `afirma_que_existe:claude-sonnet-5-5:herramienta` | 1 de 30 (3,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `cita_sin_documento:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `documento_sin_cotejar:claude-sonnet-5-5:herramienta` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 60 (0,0 %) | ≤ 1,7 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 125 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 124 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_del_juez:orden` | 56 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 65 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |

**normativa**: aprobado sobre `e8715a9`; decide `claude-sonnet-5-5` con 2 de 3; 23 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
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
| `22-lcsp-contrato-menor-en-2019.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `23-lpac-articulo-21-en-2016.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 3 de 60 (5,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 1 de 60 (1,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 120 (0,0 %) | ≤ 2,8 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 569 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 575 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 120 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 127 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 66 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 26 de 42 (61,9 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 4 de 66 (6,1 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 32 de 42 (76,2 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

**plazos-y-territorio**: aprobado sobre `e8715a9`; decide `claude-sonnet-5-5` con 2 de 3; 12 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |
| `05-plazo-reposicion-leganes.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-plazo-reposicion-tordesillas.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `07-cobertura-liquidacion-tributaria.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `08-cobertura-multa-de-trafico.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `09-plazo-contencioso-agosto.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `10-plazo-reposicion-anterior-a-la-lpac.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `11-plazo-reposicion-sin-leer-el-precepto.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
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
| H41-1 | 2026-11-06 | `69f5ecd` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H41-2 | 2026-11-06 | `d5e6619` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H41.1-1 | 2026-11-10 | `e4915f6` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H41.2-1 | 2026-11-11 | `a27c276` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H41.2-2 | 2026-11-11 | `899708c` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H41.2-3 | 2026-11-11 | `e8715a9` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |

Tasa acumulada con la skill como está: 0 de 360 (0,00 %), de H41-1, H41-2, H41.1-1, H41.2-1, H41.2-2 y H41.2-3.
Insuficiente: 360 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

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

Tasa acumulada con la skill como está: 2 de 720 (0,28 %), de H41-1, H41-2, H41.1-1, H41.2-1, H41.2-2 y H41.2-3.
Escalón de la regla 3 con 120 respuestas, el total de hoy: como mucho 2 marcadas (vale hasta una tasa del 0,48 %). El umbral de hoy, 0.028 en `evals/normativa/juez/clases.yaml`, admite 3 de 120.
Permite estrechar el umbral a 2 de 120. Lo baja una persona, con un diff de `evals/normativa/juez/clases.yaml` y estas cifras (ADR 0038, punto 4); ningún run lo cambia.

## 4. Revisión que la constitución reserva a la persona

- Fixtures nuevos (114), por directorio:
  - `internal/evals/testdata/laboratorio/familias/activaciones/sondeo/h41-2-a/normativa-producto/herramienta/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/activaciones/sondeo/h41-2-a/normativa-producto/herramienta/04-lgt-prescripcion-02/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/activaciones/sondeo/h41-2-a/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/activaciones/sondeo/h41-2-b/normativa-producto/herramienta/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/activaciones/sondeo/h41-2-b/normativa-producto/herramienta/04-lgt-prescripcion-02/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/activaciones/sondeo/h41-2-b/normativa-producto/herramienta/04-lgt-prescripcion-03/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/activaciones/sondeo/h41-2-b/normativa-producto/herramienta/04-lgt-prescripcion-04/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/activaciones/sondeo/h41-2-b/normativa-producto/herramienta/04-lgt-prescripcion-05/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/activaciones/sondeo/h41-2-b/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/activaciones/sondeo/h41/normativa-producto/herramienta/04-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/familias/activaciones/sondeo/h41/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/familias/activaciones/`: `tabla.md`
  - `internal/evals/testdata/sesiones/activaciones/evals/`: `22-lcsp-contrato-menor-en-2019.yaml`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/herramienta-antes/`: `codigo-de-la-sesion`, `eval.txt`, `modelo.txt`, `pregunta.txt`, `servidor.json`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/herramienta-antes/traza/`: `t.1000`, `t.2000`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/herramienta-las-dos/`: `codigo-de-la-sesion`, `eval.txt`, `modelo.txt`, `pregunta.txt`, `servidor.json`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/herramienta-las-dos/traza/`: `t.1000`, `t.2000`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/herramienta-mismo-turno/`: `codigo-de-la-sesion`, `eval.txt`, `modelo.txt`, `pregunta.txt`, `servidor.json`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/herramienta-mismo-turno/traza/`: `t.1000`, `t.2000`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/herramienta-ninguna/`: `codigo-de-la-sesion`, `eval.txt`, `modelo.txt`, `pregunta.txt`, `servidor.json`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/herramienta-ninguna/traza/`: `t.1000`, `t.2000`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/herramienta-otra/`: `codigo-de-la-sesion`, `eval.txt`, `modelo.txt`, `pregunta.txt`, `servidor.json`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/herramienta-otra/traza/`: `t.1000`, `t.2000`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/orden-antes/`: `codigo-de-la-sesion`, `eval.txt`, `modelo.txt`, `pregunta.txt`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/orden-antes/traza/`: `t.1000`, `t.2000`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/orden-las-dos/`: `codigo-de-la-sesion`, `eval.txt`, `modelo.txt`, `pregunta.txt`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/orden-las-dos/traza/`: `t.1000`, `t.2000`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/orden-mismo-turno/`: `codigo-de-la-sesion`, `eval.txt`, `modelo.txt`, `pregunta.txt`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/orden-mismo-turno/traza/`: `t.1000`, `t.2000`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/orden-ninguna/`: `codigo-de-la-sesion`, `eval.txt`, `modelo.txt`, `pregunta.txt`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/orden-ninguna/traza/`: `t.1000`, `t.2000`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/orden-otra/`: `codigo-de-la-sesion`, `eval.txt`, `modelo.txt`, `pregunta.txt`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/sesiones/activaciones/sesiones/orden-otra/traza/`: `t.1000`, `t.2000`
- Suite de aceptación activada y congelada (1 guiones, escritos desde el spec en la primera tarea): `h41-2-laboratorio-tabla.txtar`.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T002(hecha) T004(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConLasActivaciones`, en make ci (verde); `ci:internal/evals/activaciones_test.go:TestActivacionesDeLosUsos`, en make ci (verde) |
| FR-002 | T004(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConLasActivaciones`, en make ci (verde); `ci:internal/evals/activaciones_test.go:TestActivacionesDeLosUsos`, en make ci (verde) |
| FR-003 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConLasActivaciones`, en make ci (verde); `ci:internal/evals/activaciones_test.go:TestActivacionesDeLosUsos`, en make ci (verde) |
| FR-004 | T003(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConLasActivaciones`, en make ci (verde); `ci:internal/evals/informe_test.go:TestInformeDeLaSuiteCompletaNoCambia`, en make ci (verde) |
| FR-005 | T004(hecha) | — | tareas hechas | — |
| FR-006 | T003(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConLasActivaciones`, en make ci (verde); `ci:internal/evals/activaciones_test.go:TestActivacionesDeLosUsos`, en make ci (verde) |
| FR-010 | T006(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLaInstalacionDelProducto`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDelLaboratorio`, en make ci (verde) |
| FR-011 | T006(hecha) T007(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLaInstalacionDelProducto`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConUnaEval`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde) |
| FR-012 | T005(hecha) T006(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLaInstalacionDelProducto`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDelLaboratorio`, en make ci (verde) |
| FR-013 | T007(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConUnaEval`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde) |
| FR-014 | T006(hecha) T007(hecha) T009(hecha) | — | tareas hechas | — |
| FR-015 | T005(hecha) T006(hecha) T007(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLaInstalacionDelProducto`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDelLaboratorio`, en make ci (verde) |
| FR-020 | T001(hecha) T002(hecha) T008(hecha) T009(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaConLasActivaciones`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-021 | T001(hecha) T008(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaConLasActivaciones`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-022 | T001(hecha) T002(hecha) T008(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDelLaboratorio`, en make ci (verde) |
| FR-023 | T002(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDelLaboratorio`, en make ci (verde) |
| FR-024 | T001(hecha) T008(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaConLasActivaciones`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-030 | T007(hecha) | — | tareas hechas | — |
| FR-031 | T009(hecha) | — | tareas hechas | — |
| FR-032 | T002(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDelLaboratorio`, en make ci (verde) |
| FR-033 | — | — | SIN TAREA | `evals:normativa:sin_activar:claude-sonnet-5-5:orden`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 120 (0,0 %), ≤ 2,8 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:normativa:duracion_de_las_sesiones:orden`: 569, ≤ 900; `evals:normativa:duracion_de_las_sesiones:herramienta`: 575, ≤ 900; `evals:normativa:duracion_del_juez:orden`: 120, ≤ 900; `evals:normativa:duracion_del_juez:herramienta`: 127, ≤ 900; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 56, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 65, ≤ 900; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 % |
| SC-001 | T002(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConLasActivaciones`, en make ci (verde); `ci:internal/evals/activaciones_test.go:TestActivacionesDeLosUsos`, en make ci (verde) |
| SC-002 | T003(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConLasActivaciones`, en make ci (verde); `ci:internal/evals/informe_test.go:TestInformeDeLaSuiteCompletaNoCambia`, en make ci (verde) |
| SC-003 | T005(hecha) T006(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLaInstalacionDelProducto`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDelLaboratorio`, en make ci (verde) |
| SC-004 | T006(hecha) T007(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConLaInstalacionDelProducto`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConUnaEval`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde) |
| SC-009 | T007(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesConUnaEval`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias`, en make ci (verde) |
| SC-005 | T001(hecha) T008(hecha) | laboratorio-tabla.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaConLasActivaciones`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-006 | T002(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeLasFamilias`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDeNormativa`, en make ci (verde); `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDelLaboratorio`, en make ci (verde) |
| SC-007 | — | — | SIN TAREA | `evals:normativa:sin_activar:claude-sonnet-5-5:orden`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 120 (0,0 %), ≤ 2,8 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:normativa:duracion_de_las_sesiones:orden`: 569, ≤ 900; `evals:normativa:duracion_de_las_sesiones:herramienta`: 575, ≤ 900; `evals:normativa:duracion_del_juez:orden`: 120, ≤ 900; `evals:normativa:duracion_del_juez:herramienta`: 127, ≤ 900; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 56, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 65, ≤ 900; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 % |
| SC-008 | — | — | SIN TAREA | — |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`5cfda6e`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `437d4931` fix(H41.2): motivos de la revisión final: lo vio la ronda 2 (juez A: aprobado; juez B: aprobado). Toca `internal/evals/activaciones_test.go`, `internal/evals/informe_test.go`.
- `177dfe21` docs(H41.2): veredictos de la revisión final: solo registros de `gates/`.
- `a27c2760` docs(H41.2): registros del run: solo registros de `gates/`.
- `a68396b3` fix(H41.2): cierre en la plataforma: lo vio la ronda 3 (juez A: aprobado; juez B: aprobado). Toca `research.md`.
- `ebe3f77a` docs(H41.2): veredictos de la revisión final: solo registros de `gates/`.
- `899708cd` docs(H41.2): registros del run: solo registros de `gates/`.
- `e8715a97` docs(H41.2): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/036-h41-2-por-qu-e/quickstart.md`. Suite de aceptación congelada: `specs/036-h41-2-por-qu-e/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `f40bea97`: 17 h 22 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh f40bea97`.
