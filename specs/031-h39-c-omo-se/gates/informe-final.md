# Informe del hito H39 · Cómo se organiza una skill para que se active y se lea en los dos modos

Generado por el workflow `hito` el 2026-11-03T19:49:20Z, sobre `a593631b` de `031-h39-c-omo-se`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/031-h39-c-omo-se/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige el fichero del hito en docs/roadmap/in-progress/ y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `a593631` (medición 1): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: consulta-juridica aprobado, jurisprudencia aprobado, normativa aprobado, plazos-y-territorio aprobado (tasas en la sección 3).
- **Umbrales del job**: consulta-juridica: 4 umbrales, ninguno decide y 4 solo se publican (4 por encima de su umbral: `fuera_de_lo_grabado:claude-sonnet-5-5:orden`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); jurisprudencia: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (2 por encima de su umbral: `afirma_que_existe:claude-sonnet-5-5:orden`, `afirma_que_existe:claude-sonnet-5-5:herramienta`); normativa: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (5 por encima de su umbral: `cuenta_su_proceso:claude-sonnet-5-5:orden`, `cuenta_su_proceso:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); plazos-y-territorio: 18 umbrales, los 12 que deciden se cumplen y 6 solo se publican, todos dentro de su umbral (sección 3).
- **Revisión final**: juez A aprobado (con `claude-opus-5-5`), juez B aprobado (con `claude-fable-5-1`), 2 rondas.
- **Sesiones con otro modelo que el de su rol**: ninguna (35 sesiones registradas, todas con el modelo que `workflow.yml` da a su rol).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 12 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 20 commits; 179 files changed, 11470 insertions(+), 51 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- corrector_spec: FR-030 fijaba el reparto n/4 por eval y las 0 sesiones de una petición inválida sin comprobación en rojo → SC-007 se amplía: el test de `make ci` con el sustituto de `claude` que cuenta invocaciones falla si una petición inválida (variante en un modo que no le corresponde, n no múltiplo de 4) abre alguna sesión o si una válida no abre n/4 por eval en cada variante y modo.
- corrector_spec: FR-041 no definía el turno → un turno es una respuesta del modelo (los bloques que emite antes de recibir el resultado de ninguna herramienta, en uno o varios mensajes `assistant`), numerado desde 1; FR-043 cubre `Skill` y la primera herramienta de una misma respuesta en mensajes separados.
- corrector_spec: FR-041 dejaba ambiguo «las que activan las dos» en `entrada` → antes de la primera herramienta, como pide el hito.
- corrector_spec: la tabla no definía las celdas sin datos → con referencia y ninguna apertura, «0 de n» con su intervalo y el turno vacío; con n = 0, la fila sale con sus sesiones sin terminar y recuentos e intervalos vacíos (Edge Cases, FR-042, FR-043).
- plan: FR-030 no decía cómo se piden varias variantes y modos → `VARIANTES` y `MODOS` como listas, con su producto, y un par que FR-020 a FR-027 no dan es un error de argumentos; las 14 filas de FR-050 salen de dos lanzamientos (seis variantes en los dos modos y dos solo en herramienta) que `laboratorio tabla` junta; descartado: una lista de pares variante/modo, más larga de escribir y con otro error que contar.
- plan: FR-030 no fijaba la concurrencia ni dónde quedan las sesiones → `CONCURRENCIA` obligatoria (no hay una del job para la skill de laboratorio) y `DIRECTORIO` nuevo, que el arnés crea con `sondeo.json`; un `DIRECTORIO` que ya existe es un error de argumentos; descartado: la concurrencia de `normativa` por omisión, que ata el laboratorio a una skill del producto.
- plan: FR-041 no decía cómo partir un turno si un resultado llega entre dos mensajes `assistant` de una misma respuesta → turno nuevo solo con un `assistant` que llega tras un `tool_result` y lleva otro `message.id` (research D10, S2); descartado: separar solo por el resultado, que partiría esa respuesta.
- T001: research V4 da `<` y `>` por literales en `--describe` por `SetEscapeHTML(false)` (`internal/cli/describe.go:545`), pero `Schema.MarshalJSON` de invopop serializa cada esquema con `json.Marshal` (`invopop/jsonschema@v0.14.0/reflect.go:1124`), que los escapa: un prototipo que solo pone la `description` emite la ayuda de `norma` con `<` y `>` como escapes de JSON (barra invertida, `u` y `003c` o `003e`) → el guion `describe.txtar` afirma lo que fijan la tarea y contracts/esquemas.md §1 y §2,… (entera en `specs/031-h39-c-omo-se/gates/supuestos.md` o `clarify-respuestas.json`)
- T010: contracts/arnes.md §2 pide `MODOS` «sin repetir» pero solo da el mensaje del modo que no existe → un modo repetido dice `MODOS: «<m>» se repite`, como `VARIANTES`, una vez por elemento y en el orden en que aparece; los pares se comprueban con las variantes y los modos que valen aunque otros no valgan (todos los errores juntos), en el orden de apertura (declaración y, en cada variante, orden antes que herramienta); un elemento vacío (`VARIANTES=`) es ««» no es ninguna variante del laborator… (entera en `specs/031-h39-c-omo-se/gates/supuestos.md` o `clarify-respuestas.json`)
- T010: el contrato no dice qué es un `DIRECTORIO` que no se puede mirar (un `lstat` que falla sin «no existe», como bajo un fichero) → es un error de uso más, `DIRECTORIO: <ruta> no se puede mirar: <error>`, porque no se sabe si existe; `DIRECTORIO` se usa como llega (el guion de T011 lo hace absoluto desde la raíz).
- T010: §2 da la forma de `CLAUDE_CODE: claude --version no se pudo ejecutar: <error>` sin fijar sus causas → `claude` se busca en las carpetas absolutas del `PATH` de quien lo lanza (una relativa o vacía no se mira) y se ejecuta con su entorno; sin él, «claude no está en ninguna carpeta del PATH»; si no sale con 0, el error del proceso y lo que escribió en su salida de error; si no escribe ningún campo, «no escribió ninguna versión». La versión se pregunta solo con los argumentos y la credencial … (entera en `specs/031-h39-c-omo-se/gates/supuestos.md` o `clarify-respuestas.json`)
- T010: §4 da la línea de un par abierto a medias tras el límite de uso, no la de los pares que ya no se abren → `- <v>, <m>: 0 abiertas, <n> sin abrir tras el límite de uso.`, y «abiertas» va siempre en plural (`1 abiertas`), una forma fija; el límite se da por alcanzado, como dice §3.5, cuando un reparto vuelve con sesiones sin abrir.
- T011: cómo resuelve scripts/evals-laboratorio.sh un `DIRECTORIO` relativo «desde la raíz» → con la raíz física (`pwd -P`, como su temporal) delante y sin normalizar el resto; uno vacío o absoluto pasa tal cual, para que el error de uno vacío o que ya existe lo dé el arnés con su mensaje de contracts/arnes.md §2. Los mensajes y la línea de `make laboratorio-tabla` de la salida nombran así la ruta absoluta.
- T011: en qué punto elige scripts/evals-sondeo-llavero.sh su primer argumento → tras las comprobaciones de hoy del paso de workflow y de la sesión única, que no cambian, y antes de tocar el llavero: con uno que no es `sondeo` ni `laboratorio`, o sin ninguno, su uso y 1 sin pedir la contraseña ni ejecutar ningún guion. Sus mensajes conservan el prefijo `evals-sondeo:` de hoy, también para el laboratorio.
- clarify Q2 (criterio c): En modo orden, la primera orden kitlegal es el primer bloque tool_use de Bash cuya orden invoca el binario `kitlegal` (como orden, o dentro de una tubería o una secuencia) con un verbo de un applet que da herramienta del servidor —los de `NombresDeHerramientas`: ni `mcp` ni `skills`— y que consulta: no pide la ayuda ni lleva `--describe` ni `--dry-run`, el criterio de `Invocacion.Consulta`. Cuenta termine como termine, igual que en modo herramienta cuenta la llamada aunq… (entera en `specs/031-h39-c-omo-se/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q4 (criterio c): Abrir la referencia es un bloque tool_use de Read cuyo file_path es el de la referencia, entero o por tramos (offset y limit), o uno de Bash cuya orden nombra su fichero (`references/<fichero>.md` de la skill de laboratorio). El turno de la primera apertura es el del primer bloque que cumple, y la referencia abierta varias veces cuenta una vez (FR-041). Grep y Glob no cuentan.
- clarify Q5 (criterio c): Cuatro evals de laboratorio, todas sobre la misma herramienta —la que lleva la regla en la descripción de su parámetro en la variante `parametro`— y con lecturas que ya sirve testdata/evals/: 10 sesiones de cada una por variante y modo con las 40 de la aceptación. FR-030 exige entonces que el número de sesiones por variante y modo que pide la persona sea múltiplo de 4; si no, sale con un código distinto de 0 sin abrir ninguna sesión.

**Alcance (lo que queda fuera o dentro del hito)**

- corrector_spec: FR-030 no decía qué skills tiene instaladas una sesión del sondeo → lectura conservadora: solo la skill de laboratorio de su variante y, en `entrada`, su skill de entrada; ninguna skill del producto, porque `normativa` instalada podría activarse en lugar de la de laboratorio y la tabla no mediría la variante.
- plan: FR-052 → la propuesta de cambio queda pendiente del sondeo de la persona (FR-050: `make evals-laboratorio` en dos lanzamientos y `make laboratorio-tabla`, quickstart.md §6) y del ADR que precisa el 0046 con sus cifras (FR-051); el run no hace ninguno de los dos, y la propuesta no se fusiona sin ellos.
- plan: FR-053 → si la medición del cierre da los rojos de partida en modo herramienta (eval 04 de `normativa`, `sin_activar`; eval 02 de `consulta-juridica`), la reparación no toca `skills/` ni reorganiza las skills del producto, y el informe da sus cifras para que decida una persona; descartado: arreglarlos con texto nuevo en las skills, que el hito deja fuera.

**Skill (lo que pide, dice o comprueba una skill)**

- T005: contracts/laboratorio.md §4 y §5 fijan la `description`, la estructura del cuerpo de control y las frases que cambian, pero no la redacción completa del título, de los pasos ni de la frase de la cita, ni el título de `consulta-legal` → se escriben con las frases del contrato literales («nada de lo que diga la respuesta sale de memoria: sale del texto que devuelve la herramienta», «cita la norma y el bloque así: [BOE-A-1978-31229, bloque a140].», los cuatro identificadores, `kitlegal boe in… (entera en `specs/031-h39-c-omo-se/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run

- 2026-11-03T19:09:37Z · La ronda 2 de la revisión final (ciclo 1) aprobó con los dos jueces lo que quedaba en revision-pendiente.md; se conserva como specs/031-h39-c-omo-se/gates/revision-pendiente-resuelto-r2.md.

- Observaciones de los jueces, por debajo del umbral y sin corregir: 11 en `spec-r2.json`, 11 en `plan-r1.json`, 9 en `tasks-r2.json`, 11 en `revision-a-r2.json`, 7 en `revision-b-r2.json`.

### Internos (23)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T002 (1), T003 (1), T004 (1), T005 (1), T007 (1), T008 (4), T009 (6), T010 (3), T011 (1), corrector_revision (3), plan (1). Enteras en `specs/031-h39-c-omo-se/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**consulta-juridica**: aprobado sobre `a593631`; decide `claude-sonnet-5-5` con 2 de 3; 2 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-poliza-tras-el-fallecimiento.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `02-una-sola-capacidad.yaml` | 3/3 | 1/3 | 2/3 | 0/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 22 de 6 (366,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 11 de 6 (183,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 19 de 6 (316,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 11 de 6 (183,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

**jurisprudencia**: aprobado sobre `a593631`; decide `claude-sonnet-5-5` con 2 de 3; 10 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-resumen-sin-documento.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `03-por-materia.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `04-documento-pegado.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-documento-que-no-es-el-pedido.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `07-resumen-de-una-conocida.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
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
| `afirma_que_existe:claude-sonnet-5-5:herramienta` | 2 de 30 (6,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `cita_sin_documento:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `documento_sin_cotejar:claude-sonnet-5-5:herramienta` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 60 (0,0 %) | ≤ 1,7 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 125 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 124 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_del_juez:orden` | 50 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 53 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |

**normativa**: aprobado sobre `a593631`; decide `claude-sonnet-5-5` con 2 de 3; 23 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 0/3 | 3/3 | 1/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
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
| `23-lpac-articulo-21-en-2016.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 1 de 60 (1,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 1 de 60 (1,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 120 (0,0 %) | ≤ 2,8 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 525 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 504 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 84 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 78 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 66 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 33 de 42 (78,6 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 8 de 66 (12,1 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 36 de 42 (85,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

**plazos-y-territorio**: aprobado sobre `a593631`; decide `claude-sonnet-5-5` con 2 de 3; 12 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |
| `05-plazo-reposicion-leganes.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `06-plazo-reposicion-tordesillas.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `07-cobertura-liquidacion-tributaria.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `08-cobertura-multa-de-trafico.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `09-plazo-contencioso-agosto.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `10-plazo-reposicion-anterior-a-la-lpac.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
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

Clave de la cabeza: skill `7dbf0c5` (árbol de `skills/jurisprudencia`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

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
| H32-1 | 2026-10-30 | `7f8ef89` | `7dbf0c5` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H39-1 | 2026-11-03 | `a593631` | `7dbf0c5` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |

Tasa acumulada con la skill como está: 0 de 120 (0,00 %), de H32-1 y H39-1.
Insuficiente: 120 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

**Serie de cierres de normativa** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `afbcae6` (árbol de `skills/normativa`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

`afirma_lo_no_leido`:

| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |
|---|---|---|---|---|---|---|---|---|---|
| H32-1 | 2026-10-30 | `7f8ef89` | `8566858` | 0/60 | 2/60 | 2/90 | 0/30 | 2/120 | no: otra skill |
| H39-1 | 2026-11-03 | `a593631` | `afbcae6` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | sí |

Tasa acumulada con la skill como está: 0 de 120 (0,00 %), de H39-1.
Insuficiente: 120 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

## 4. Revisión que la constitución reserva a la persona

- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/plazos/describe.json`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h20-fecha-main.txtar`.
- Fixture o esquema EXISTENTE modificado: `schemas/bloque.json`.
- Fixture o esquema EXISTENTE modificado: `schemas/cita.json`.
- Fixture o esquema EXISTENTE modificado: `schemas/grafo.json`.
- Fixture o esquema EXISTENTE modificado: `schemas/instalacion.json`.
- Fixture o esquema EXISTENTE modificado: `schemas/municipio.json`.
- Fixture o esquema EXISTENTE modificado: `schemas/norma.json`.
- Fixture o esquema EXISTENTE modificado: `schemas/plazo.json`.
- Fixtures nuevos (72), por directorio:
  - `internal/evals/testdata/laboratorio/respuestas/`: `cumple-con-lineas-en-blanco.txt`, `cumple.txt`, `no-cumple-sello-con-texto.txt`, `no-cumple-sello-en-negrita.txt`, `no-cumple-sello-no-al-final.txt`, `no-cumple-sin-sello.txt`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/control/herramienta/01-constitucion-articulo-140-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/control/herramienta/02-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/control/herramienta/03-lrjsp-principio-de-legalidad-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/control/herramienta/04-et-vacaciones-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/control/orden/01-constitucion-articulo-140-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/control/orden/02-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/control/orden/03-lrjsp-principio-de-legalidad-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/control/orden/04-et-vacaciones-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/entrada/herramienta/01-constitucion-articulo-140-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/entrada/herramienta/02-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/entrada/herramienta/03-lrjsp-principio-de-legalidad-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/referencia-paso/orden/01-constitucion-articulo-140-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/referencia-paso/orden/02-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/referencia-paso/orden/03-lrjsp-principio-de-legalidad-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/referencia-secundaria/herramienta/01-constitucion-articulo-140-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/referencia-secundaria/herramienta/02-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-a/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/sondeo/h39-b/control/herramienta/01-constitucion-articulo-140-02/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-b/descripcion-sin-herramientas/orden/02-lgt-prescripcion-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-b/descripcion-sin-herramientas/orden/03-lrjsp-principio-de-legalidad-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-b/parametro/herramienta/01-constitucion-articulo-140-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-b/parametro/herramienta/04-et-vacaciones-01/`: `codigo-de-la-sesion`, `sesion.err`, `sesion.jsonl`
  - `internal/evals/testdata/laboratorio/sondeo/h39-b/`: `sondeo.json`
  - `internal/evals/testdata/laboratorio/`: `tabla.md`
- Suite de aceptación activada y congelada (1 guiones, escritos desde el spec en la primera tarea): `h39-describe.txtar`.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T001(hecha) T002(hecha) T003(hecha) | describe.txtar  | comprobado por su control | `ci:internal/app/ayudas_test.go:TestAyudasDeLosArgumentos`, en make ci (verde); `ci:internal/cli/describe_test.go:TestAyudaComoDescripcion`, en make ci (verde) |
| FR-002 | T001(hecha) T002(hecha) T003(hecha) | describe.txtar  | comprobado por su control | `ci:internal/app/ayudas_test.go:TestAyudasDeLosArgumentos`, en make ci (verde) |
| FR-003 | T001(hecha) T002(hecha) T003(hecha) | describe.txtar  | comprobado por su control | `ci:internal/app/ayudas_test.go:TestAyudasDeLosArgumentos`, en make ci (verde); `ci:internal/cli/describe_test.go:TestAyudaComoDescripcion`, en make ci (verde) |
| FR-004 | T001(hecha) T002(hecha) | — | comprobado por su control | `ci:internal/app/esquemas_test.go:TestEsquemasPublicados`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-005 | T002(hecha) | — | tareas hechas | — |
| FR-010 | T007(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestLaboratorioFueraDelProducto`, en make ci (verde) |
| FR-011 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestVariantesDelLaboratorio`, en make ci (verde) |
| FR-012 | T006(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestEvalsDelLaboratorio`, en make ci (verde) |
| FR-013 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestReglaTestigo`, en make ci (verde) |
| FR-014 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestReglaTestigo`, en make ci (verde) |
| FR-020 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestVariantesDelLaboratorio`, en make ci (verde) |
| FR-021 | T005(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestVariantesDelLaboratorio`, en make ci (verde); `ci:internal/evals/laboratorio_servidor_test.go:TestServidorDelLaboratorio`, en make ci (verde) |
| FR-022 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestVariantesDelLaboratorio`, en make ci (verde) |
| FR-023 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestVariantesDelLaboratorio`, en make ci (verde) |
| FR-024 | T005(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestVariantesDelLaboratorio`, en make ci (verde); `ci:internal/evals/laboratorio_servidor_test.go:TestServidorDelLaboratorio`, en make ci (verde) |
| FR-025 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestVariantesDelLaboratorio`, en make ci (verde) |
| FR-026 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestVariantesDelLaboratorio`, en make ci (verde) |
| FR-027 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestVariantesDelLaboratorio`, en make ci (verde) |
| FR-028 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestVariantesDelLaboratorio`, en make ci (verde) |
| FR-029 | T008(hecha) | — | comprobado por su control | `ci:internal/mcp/instrucciones_test.go:TestInstrucciones`, en make ci (verde); `ci:internal/evals/laboratorio_servidor_test.go:TestServidorDelLaboratorio`, en make ci (verde) |
| FR-030 | T010(hecha) T011(hecha) T012(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDelLaboratorio`, en make ci (verde); `ci:internal/evals/laboratorio_sondeo_test.go:TestRecetasDelSondeo`, en make ci (verde) |
| FR-031 | T010(hecha) T011(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDelLaboratorio`, en make ci (verde) |
| FR-032 | T011(hecha) T012(hecha) | — | tareas hechas | — |
| FR-033 | T010(hecha) T011(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestRecetasDelSondeo`, en make ci (verde) |
| FR-040 | T009(hecha) T012(hecha) | — | tareas hechas | — |
| FR-041 | T009(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDelLaboratorio`, en make ci (verde) |
| FR-042 | T009(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDelLaboratorio`, en make ci (verde) |
| FR-043 | T009(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDelLaboratorio`, en make ci (verde) |
| FR-050 | — | — | SIN TAREA | — |
| FR-051 | — | — | SIN TAREA | — |
| FR-052 | — | — | SIN TAREA | — |
| FR-053 | — | — | SIN TAREA | `evals:normativa:sin_activar:claude-sonnet-5-5:orden`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 120 (0,0 %), ≤ 2,8 %; `evals:consulta-juridica:habla_en_primera_o_segunda_persona`: el job no publica ese umbral; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 % |
| SC-001 | T001(hecha) T002(hecha) T003(hecha) | describe.txtar  | comprobado por su control | `ci:internal/app/ayudas_test.go:TestAyudasDeLosArgumentos`, en make ci (verde); `ci:internal/cli/describe_test.go:TestAyudaComoDescripcion`, en make ci (verde) |
| SC-002 | T002(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| SC-003 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestReglaTestigo`, en make ci (verde) |
| SC-004 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestVariantesDelLaboratorio`, en make ci (verde) |
| SC-005 | T007(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_test.go:TestLaboratorioFueraDelProducto`, en make ci (verde) |
| SC-006 | T009(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_tabla_test.go:TestTablaDelLaboratorio`, en make ci (verde) |
| SC-007 | T010(hecha) T011(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDelLaboratorio`, en make ci (verde) |
| SC-008 | — | — | SIN TAREA | `evals:normativa:sin_activar:claude-sonnet-5-5:orden`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 120 (0,0 %), ≤ 2,8 %; `evals:consulta-juridica:habla_en_primera_o_segunda_persona`: el job no publica ese umbral; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 % |
| SC-009 | — | — | SIN TAREA | — |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`44e009e`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `6121c82a` fix(H39): motivos de la revisión final: lo vio la ronda 2 (juez A: aprobado; juez B: aprobado). Toca `internal/evals/laboratorio_servidor.go`, `internal/evals/laboratorio_servidor_test.go`, `internal/evals/laboratorio_tabla.go`, `internal/evals/laboratorio_tabla_test.go`.
- `8641a81e` docs(H39): veredictos de la revisión final: solo registros de `gates/`.
- `a593631b` docs(H39): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/031-h39-c-omo-se/quickstart.md`. Suite de aceptación congelada: `specs/031-h39-c-omo-se/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `50b213d0`: 11 h 41 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh 50b213d0`.
