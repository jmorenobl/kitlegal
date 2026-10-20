# Informe del hito H20 · `boe articulo --fecha`: la redacción vigente en una fecha + `boe-legislacion` v0.2 (ADR 0027)

Generado por el workflow `hito` el 2026-10-19T19:38:38Z, sobre `77211a9` de `024-h20-boe-articulo-fecha`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/024-h20-boe-articulo-fecha/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige la sección del hito en docs/ROADMAP.md y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `77211a9` (medición 1): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: boe-legislacion aprobado, jurisprudencia aprobado, legal-core aprobado (tasas en la sección 3).
- **Umbrales del job**: boe-legislacion: 15 umbrales, los 11 que deciden se cumplen y 4 solo se publican (3 por encima de su umbral: `cuenta_su_proceso:claude-sonnet-5-5:orden`, `afirma_lo_no_leido:claude-sonnet-5-5:herramienta`, `cuenta_su_proceso:claude-sonnet-5-5:herramienta`); jurisprudencia: 15 umbrales, los 11 que deciden se cumplen y 4 solo se publican (2 por encima de su umbral: `afirma_que_existe:claude-sonnet-5-5:orden`, `afirma_que_existe:claude-sonnet-5-5:herramienta`); legal-core: 6 umbrales, los 6 que deciden se cumplen (sección 3).
- **Revisión final**: juez A aprobado (con `claude-opus-5-5`), juez B aprobado (con `claude-fable-5-1`), 2 rondas.
- **Sesiones con otro modelo que el de su rol**: ninguna (40 sesiones registradas, todas con el modelo que `workflow.yml` da a su rol).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 15 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 22 commits; 125 files changed, 11831 insertions(+), 465 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- corrector_plan: [e] FR-007 y FR-082 sin test de la lectura fechada con la fuente caída, con el límite de ritmo o con los metadatos caídos → `TestFallosNoSeGuardan` gana tres casos con `Fecha` (bloque ilegible, 4; metadatos con el 503, 4; 429 de los metadatos, 5) sobre los sintéticos de H4, cada uno con la misma clase, código y mensaje que su consulta sin `Fecha`, nada guardado y el 4 de la entrada de las versiones ausente con `--offline`; el diseño dice que la entrada solo se escribe con el bloq… (entera en `specs/024-h20-boe-articulo-fecha/gates/supuestos.md` o `clarify-respuestas.json`)
- T005: «la lectura fechada se sirve solo de [la entrada de las versiones]» no dice si, sin ella, los metadatos de la norma se resuelven con su entrada de 300 s, como en la lectura sin fecha, o se piden siempre → se resuelven como en la lectura sin fecha (la entrada de los metadatos vigente se sirve; si no, se piden y se escriben), por el mismo código (`resolverElBloque`, compartido con `articuloDelBloque`): es la lectura que no añade comportamiento, pide los mismos recursos que la de hoy (FR-006)… (entera en `specs/024-h20-boe-articulo-fecha/gates/supuestos.md` o `clarify-respuestas.json`)
- T006: contracts/articulo-fecha.md §1 dice que la propiedad `fecha` de la entrada de `--describe` va «con la ayuda de la bandera», pero el kernel no pone en la entrada la ayuda de ningún argumento (ni la de `norma` y `bloque`, ni la de las banderas de `plazos`) → `fecha` sale como el kernel refleja todo argumento, `{"type":"string"}`, y su ayuda llega a `--help`: ponerla en el esquema exigiría cambiar `internal/cli`, que la tarea no declara, y con él el `--describe` de los demás verbos, que FR-08… (entera en `specs/024-h20-boe-articulo-fecha/gates/supuestos.md` o `clarify-respuestas.json`)
- T006: la `Descripcion` del verbo `articulo` («Devuelve el texto vigente de un bloque…») no la nombra la tarea, y la fila de la tabla que fija contracts/articulo-fecha.md §1 solo cambia la orden y los campos → se deja como está: cambiarla cambiaría también la descripción de la herramienta `boe_articulo` y la tercera columna de la tabla de la skill, que ni FR-001 ni el contrato piden.
- clarify Q3 (criterio c): La redacción posterior es la redacción de hoy de la nota [5], la última versión del bloque, la misma que devuelve la lectura sin fecha, cuando no es la elegida, con su `fecha_vigencia` y su `norma_modificadora` aunque esa fecha sea posterior al día de la consulta. El binario no mira el reloj para elegirla: la lectura fechada depende solo de la respuesta del BOE y de la fecha pedida. El texto fijo que fije el plan para la línea de FR-041 da la fecha de vigencia y la norma… (entera en `specs/024-h20-boe-articulo-fecha/gates/supuestos.md` o `clarify-respuestas.json`)

**Alcance (lo que queda fuera o dentro del hito)**

- corrector_plan: [f][i] FR-061 y SC-012 piden corregir `docs/WORKFLOW.md`, cuya frase de `medir_cierre` (línea 139) dice que en `boe-legislacion` deciden nueve, y el plan lo dejaba sin cambiar con una comprobación que no se sostenía → la tarea de los documentos declara `docs/WORKFLOW.md` y solo cambia esa frase: deciden once, con `redaccion_sin_fechar:<modelo>:<modo>` en cada modo detrás de `sin_activar`, sin reescribir lo demás (FR-062; contracts/documentos.md, research D21, plan.md).

**Skill (lo que pide, dice o comprueba una skill)**

- corrector_spec: FR-040 y FR-046 daban dos lecturas a una pregunta que sitúa el momento por un hecho sin fecha → lectura conservadora: si ni la pregunta ni la conversación dan la fecha del hecho, la skill pide el día y no lee ni cita ninguna redacción como la de ese momento (FR-040, Edge Cases); FR-046 queda para las preguntas que no nombran otro momento que hoy.
- corrector_spec: FR-049 restringía a la lectura sin fecha que ninguna lectura trae el fin de una redacción, que sigue siendo cierto con la fechada (la posterior no es el fin de la citada si hay versiones entre medias, como el art. 22 de la LRBRL el 2005-06-15) → se quita de FR-049, y FR-041 prohíbe decir hasta cuándo rigió la redacción citada y presentar la vigencia de la posterior como su fin.
- corrector_spec: US1.2 pedía «el Real Decreto-ley 3/2020» en la línea de la posterior, que el binario no da y lo grabado no trae → US1.2, la historia de US1 y la eval (a) de FR-052 piden `BOE-A-2020-1651` tal como lo da el binario, como FR-041 y FR-054; SC-003, humano, queda como lo escribe el hito.
- corrector_spec: [i] la tabla «Uso, de fuera adentro» no contaba las dos lecturas fechadas de FR-045 → sus filas primera y tercera dicen que, con un año o un mes sin día, son dos lecturas fechadas del bloque (como mucho tres salidas por bloque con la de hoy de FR-042) y como mucho dos líneas de FR-041 o FR-043 y una de la posterior por bloque.
- corrector_spec: FR-046 («no nombra otro momento que hoy») se solapaba con FR-040 en una pregunta con la fecha de hoy y dejaba sin regla «qué decía antes» sin fecha, periodo ni hecho → FR-046 pasa a ser el complemento de FR-040 y FR-045 (sin fecha, sin año o mes y sin hecho): la fecha de hoy se lee con `--fecha` y «qué decía antes» sin fecha se lee sin ella, como antes de la corrección r1; sustituye en eso a la última frase de la primera línea de este fichero.
- clarify Q4 (criterio c): La skill hace dos lecturas fechadas del bloque, una con el primer día y otra con el último del periodo (del año: 01-01 y 12-31; del mes: el día 1 y su último día). Dos versiones son la misma si las dos salidas dan la misma `fecha_vigencia` y la misma `norma_modificadora`. Si es la misma versión, esa redacción rigió todo el periodo, porque con la regla de FR-002 ninguna versión entró en vigor entre medias, y la skill la cita con la línea de FR-041 de cada lectura, que dic… (entera en `specs/024-h20-boe-articulo-fecha/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run


- Observaciones de los jueces, por debajo del umbral y sin corregir: 9 en `spec-r3.json`, 9 en `plan-r2.json`, 7 en `tasks-r2.json`, 5 en `revision-a-r2.json`, 6 en `revision-b-r2.json`.

### Internos (11)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T001 (3), T007 (1), T008 (2), T009 (1), T010 (1), T013 (2), corrector_revision (1). Enteras en `specs/024-h20-boe-articulo-fecha/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**boe-legislacion**: aprobado sobre `77211a9`; decide `claude-sonnet-5-5` con 2 de 3; 23 evals, 2 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `22-lcsp-contrato-menor-en-2019.yaml` | 3/3 | 3/3 | 3/3 | 3/3 | **nueva** |
| `23-lpac-articulo-21-en-2016.yaml` | 3/3 | 3/3 | 3/3 | 3/3 | **nueva** |
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
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
| `21-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 1 de 60 (1,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 2 de 60 (3,3 %) | ≤ 2,8 % | por encima | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 1 de 60 (1,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 2 de 120 (1,7 %) | ≤ 2,8 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 472 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 507 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 89 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 100 | ≤ 900 | sí | sí |

**jurisprudencia**: aprobado sobre `77211a9`; decide `claude-sonnet-5-5` con 2 de 3; 10 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `02-resumen-sin-documento.yaml` | 3/3 | 1/3 | 3/3 | 3/3 |  |
| `03-por-materia.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `04-documento-pegado.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-documento-que-no-es-el-pedido.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `07-resumen-de-una-conocida.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `08-doctrina-con-el-fallo-delante.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `09-de-que-trata-con-la-ficha-sola.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `10-doctrina-dada-por-hecha.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |

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
| `duracion_del_juez:orden` | 45 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 40 | ≤ 900 | sí | sí |

**legal-core**: aprobado sobre `77211a9`; decide `claude-sonnet-5-5` con 2 de 3; 9 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 1/3 | — | — |  |
| `05-plazo-reposicion-leganes.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `06-plazo-reposicion-tordesillas.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `07-cobertura-liquidacion-tributaria.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `08-cobertura-multa-de-trafico.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `09-plazo-contencioso-agosto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 21 (0,0 %) | ≤ 0,0 % | sí | sí |
| `plazo_sin_fecha:claude-sonnet-5-5:orden` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cobertura_sin_declarar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 21 (0,0 %) | ≤ 0,0 % | sí | sí |
| `plazo_sin_fecha:claude-sonnet-5-5:herramienta` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cobertura_sin_declarar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

**Serie de cierres de boe-legislacion** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `49b59ac` (árbol de `skills/boe-legislacion`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

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
| H20-1 | 2026-10-19 | `77211a9` | `49b59ac` | 0/60 | 2/60 | 0/90 | 2/30 | 2/120 | sí |

Tasa acumulada con la skill como está: 2 de 120 (1,67 %), de H20-1.
Insuficiente: 120 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

**Serie de cierres de jurisprudencia** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `4759aa4` (árbol de `skills/jurisprudencia`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

`afirma_lo_no_leido`:

| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |
|---|---|---|---|---|---|---|---|---|---|
| H25-1 | 2026-10-10 | `3f99446` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H25-2 | 2026-10-10 | `e380f0a` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: 1 respuesta sin juzgar |
| H25-3 | 2026-10-10 | `d51f3d7` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H26-1 | 2026-10-13 | `b29fe6c` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H23.1-1 | 2026-10-16 | `60e8678` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H9-1 | 2026-10-18 | `3f14e3a` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H9-2 | 2026-10-18 | `92fc232` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H20-1 | 2026-10-19 | `77211a9` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |

Tasa acumulada con la skill como está: 0 de 240 (0,00 %), de H23.1-1, H9-1, H9-2 y H20-1.
Insuficiente: 240 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

## 4. Revisión que la constitución reserva a la persona

- Fixture o esquema EXISTENTE modificado: `schemas/bloque.json`.
- Fixture o esquema EXISTENTE modificado: `schemas/eval.yaml.json`.
- Suite de aceptación activada y congelada (6 guiones, escritos desde el spec en la primera tarea): `h20-fecha-codigos.txtar`, `h20-fecha-grafo.txtar`, `h20-fecha-herramienta.txtar`, `h20-fecha-main.txtar`, `h20-fecha-offline.txtar`, `h20-fecha-version.txtar`.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T001(hecha) T005(hecha) T006(hecha) | fecha-codigos.txtar fecha-grafo.txtar fecha-herramienta.txtar fecha-main.txtar fecha-offline.txtar fecha-version.txtar  | comprobado por su control | `ci:internal/app/testdata/script/h20-fecha-herramienta.txtar`, en make ci (verde); `ci:internal/app/herramientas_test.go:TestHerramientasDelServidor`, en make ci (verde) |
| FR-002 | T001(hecha) T003(hecha) T005(hecha) | fecha-version.txtar  | comprobado por su control | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFecha`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-version.txtar`, en make ci (verde) |
| FR-003 | T001(hecha) T005(hecha) T006(hecha) T011(hecha) | fecha-codigos.txtar fecha-version.txtar  | comprobado por su control | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFecha`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-version.txtar`, en make ci (verde) |
| FR-004 | T005(hecha) T011(hecha) | — | comprobado por su control | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFecha`, en make ci (verde) |
| FR-005 | T001(hecha) T004(hecha) T006(hecha) | fecha-codigos.txtar  | comprobado por su control | `ci:internal/source/boe/ids_test.go:TestValidarFecha`, en make ci (verde); `ci:internal/source/boe/fuente_test.go:TestFallosNoSeGuardan`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-codigos.txtar`, en make ci (verde) |
| FR-006 | T001(hecha) T005(hecha) T014(hecha) | fecha-codigos.txtar  | tareas hechas | — |
| FR-007 | T001(hecha) T005(hecha) | fecha-codigos.txtar  | comprobado por su control | `ci:internal/source/boe/ids_test.go:TestValidarFecha`, en make ci (verde); `ci:internal/source/boe/fuente_test.go:TestFallosNoSeGuardan`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-codigos.txtar`, en make ci (verde) |
| FR-010 | T001(hecha) T003(hecha) T005(hecha) T006(hecha) | fecha-version.txtar  | tareas hechas | — |
| FR-011 | T005(hecha) | — | comprobado por su control | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFecha`, en make ci (verde) |
| FR-012 | T001(hecha) T003(hecha) T005(hecha) T006(hecha) | fecha-main.txtar  | comprobado por su control | `ci:internal/app/testdata/script/h20-fecha-main.txtar`, en make ci (verde); `ci:internal/source/boe/fuente_test.go:TestGolden`, en make ci (verde) |
| FR-013 | T001(hecha) T005(hecha) | fecha-version.txtar  | comprobado por su control | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFecha`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-version.txtar`, en make ci (verde) |
| FR-014 | T006(hecha) | — | comprobado por su control | `ci:Makefile:schema-check`, en make ci (verde); `ci:internal/app/boe_test.go:TestSalidaDeBoeContraSchemas`, en make ci (verde) |
| FR-020 | T001(hecha) T005(hecha) | fecha-offline.txtar  | comprobado por su control | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFechaYLaCache`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-offline.txtar`, en make ci (verde) |
| FR-021 | T001(hecha) T005(hecha) | fecha-offline.txtar  | comprobado por su control | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFechaYLaCache`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-offline.txtar`, en make ci (verde) |
| FR-022 | T001(hecha) T005(hecha) | fecha-offline.txtar  | comprobado por su control | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFechaYLaCache`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-offline.txtar`, en make ci (verde) |
| FR-030 | T001(hecha) T002(hecha) T007(hecha) | fecha-grafo.txtar  | comprobado por su control | `ci:internal/graph/almacen_test.go:TestApplySoloLoNuevo`, en make ci (verde); `ci:internal/graph/integracion_test.go:TestIntegracionLecturaFechada`, en make ci (verde); `ci:internal/app/grafo_test.go:TestLecturaFechadaEnElGrafo`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-grafo.txtar`, en make ci (verde) |
| FR-031 | T001(hecha) T002(hecha) T007(hecha) | fecha-grafo.txtar  | comprobado por su control | `ci:internal/graph/almacen_test.go:TestApplySoloLoNuevo`, en make ci (verde); `ci:internal/graph/integracion_test.go:TestIntegracionLecturaFechada`, en make ci (verde); `ci:internal/app/grafo_test.go:TestLecturaFechadaEnElGrafo`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-grafo.txtar`, en make ci (verde) |
| FR-032 | T001(hecha) T007(hecha) | fecha-grafo.txtar  | comprobado por su control | `ci:internal/graph/almacen_test.go:TestApplySoloLoNuevo`, en make ci (verde); `ci:internal/graph/integracion_test.go:TestIntegracionLecturaFechada`, en make ci (verde); `ci:internal/app/grafo_test.go:TestLecturaFechadaEnElGrafo`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-grafo.txtar`, en make ci (verde) |
| FR-040 | T011(hecha) | — | tareas hechas | — |
| FR-041 | T011(hecha) | — | tareas hechas | — |
| FR-042 | T011(hecha) | — | tareas hechas | — |
| FR-043 | T011(hecha) | — | tareas hechas | — |
| FR-044 | T011(hecha) | — | tareas hechas | — |
| FR-045 | T011(hecha) | — | tareas hechas | — |
| FR-046 | T011(hecha) | — | tareas hechas | — |
| FR-047 | T010(hecha) T011(hecha) | — | comprobado por su control | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 60 (0,0 %), ≤ 0,0 % |
| FR-048 | T006(hecha) T011(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-049 | T011(hecha) | — | tareas hechas | — |
| FR-050 | T010(hecha) T015(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-051 | T008(hecha) T010(hecha) | — | comprobado por su control | `ci:internal/evals/formato_test.go:TestFormatoDeLasRedaccionesFechadas`, en make ci (verde); `ci:internal/evals/juzgar_test.go:TestJuzgarLecturasFechadas`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-052 | T008(hecha) T010(hecha) T015(hecha) | — | comprobado por su control | `ci:internal/evals/formato_test.go:TestFormatoDeLasRedaccionesFechadas`, en make ci (verde); `ci:internal/evals/juzgar_test.go:TestJuzgarLecturasFechadas`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-053 | T009(hecha) T010(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-054 | T010(hecha) T015(hecha) | — | comprobado por su control | `evals:boe-legislacion:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 % |
| FR-055 | T010(hecha) | — | comprobado por su control | `evals:boe-legislacion:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `ci:internal/evals/informe_test.go:TestUmbralesDeBoeLegislacion`, en make ci (verde); `ci:internal/evals/juzgar_test.go:TestJuzgarLecturasFechadas`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills`, en make ci (verde) |
| FR-056 | T010(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills`, en make ci (verde) |
| FR-057 | T010(hecha) | — | tareas hechas | — |
| FR-058 | T009(hecha) T010(hecha) | — | comprobado por su control | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5`: 2 de 120 (1,7 %), ≤ 2,8 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 % |
| FR-059 | T010(hecha) T014(hecha) | — | comprobado por su control | `evals:boe-legislacion:duracion_de_las_sesiones:orden`: 472, ≤ 900; `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`: 507, ≤ 900; `evals:boe-legislacion:duracion_del_juez:orden`: 89, ≤ 900; `evals:boe-legislacion:duracion_del_juez:herramienta`: 100, ≤ 900 |
| FR-060 | T013(hecha) | — | tareas hechas | — |
| FR-061 | T013(hecha) | — | tareas hechas | — |
| FR-062 | T013(hecha) | — | tareas hechas | — |
| FR-063 | T013(hecha) T014(hecha) | — | tareas hechas | — |
| FR-070 | T001(hecha) T006(hecha) T014(hecha) | fecha-main.txtar  | tareas hechas | — |
| FR-071 | T010(hecha) | — | comprobado por su control | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 45, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 40, ≤ 900; `evals:legal-core:sin_activar:claude-sonnet-5-5:orden`: 0 de 21 (0,0 %), ≤ 0,0 %; `evals:legal-core:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 21 (0,0 %), ≤ 0,0 %; `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 % |
| FR-072 | T010(hecha) | — | tareas hechas | — |
| FR-080 | T014(hecha) | — | comprobado por su control | `ci:Makefile:ci`, en make ci (verde) |
| FR-081 | T005(hecha) T015(hecha) | — | comprobado por su control | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFecha`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-version.txtar`, en make ci (verde) |
| FR-082 | T001(hecha) T005(hecha) | fecha-offline.txtar  | comprobado por su control | `ci:internal/source/boe/ids_test.go:TestValidarFecha`, en make ci (verde); `ci:internal/source/boe/fuente_test.go:TestFallosNoSeGuardan`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-codigos.txtar`, en make ci (verde); `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFechaYLaCache`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-offline.txtar`, en make ci (verde) |
| FR-083 | T001(hecha) | fecha-main.txtar  | comprobado por su control | `ci:internal/app/testdata/script/h20-fecha-main.txtar`, en make ci (verde); `ci:internal/source/boe/fuente_test.go:TestGolden`, en make ci (verde) |
| FR-084 | T001(hecha) T006(hecha) | fecha-herramienta.txtar  | comprobado por su control | `ci:internal/app/testdata/script/h20-fecha-herramienta.txtar`, en make ci (verde); `ci:internal/app/herramientas_test.go:TestHerramientasDelServidor`, en make ci (verde) |
| FR-085 | T006(hecha) | — | comprobado por su control | `ci:Makefile:schema-check`, en make ci (verde); `ci:internal/app/boe_test.go:TestSalidaDeBoeContraSchemas`, en make ci (verde) |
| FR-086 | T002(hecha) T007(hecha) | — | comprobado por su control | `ci:internal/graph/almacen_test.go:TestApplySoloLoNuevo`, en make ci (verde); `ci:internal/graph/integracion_test.go:TestIntegracionLecturaFechada`, en make ci (verde); `ci:internal/app/grafo_test.go:TestLecturaFechadaEnElGrafo`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-grafo.txtar`, en make ci (verde) |
| FR-087 | T005(hecha) | — | comprobado por su control | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFechaYLaCache`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-offline.txtar`, en make ci (verde) |
| FR-088 | T010(hecha) | — | comprobado por su control | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 45, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 40, ≤ 900; `evals:legal-core:sin_activar:claude-sonnet-5-5:orden`: 0 de 21 (0,0 %), ≤ 0,0 %; `evals:legal-core:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 21 (0,0 %), ≤ 0,0 %; `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 % |
| FR-089 | T008(hecha) T010(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeBoeLegislacion`, en make ci (verde); `ci:internal/evals/juzgar_test.go:TestJuzgarLecturasFechadas`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills`, en make ci (verde) |
| FR-090 | T005(hecha) T007(hecha) T010(hecha) T012(hecha) | — | comprobado por su control | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFecha`, en make ci (verde); `ci:internal/source/boe/grafo_test.go:TestObservadoDeBoe`, en make ci (verde); `ci:internal/app/grafo_test.go:TestLecturaFechadaEnElGrafo`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeBoeLegislacion`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills`, en make ci (verde) |
| FR-091 | T013(hecha) | — | tareas hechas | — |
| FR-092 | T014(hecha) | — | tareas hechas | — |
| SC-001 | T010(hecha) T011(hecha) | — | comprobado por su control | `evals:boe-legislacion:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5`: 2 de 120 (1,7 %), ≤ 2,8 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:duracion_de_las_sesiones:orden`: 472, ≤ 900; `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`: 507, ≤ 900; `evals:boe-legislacion:duracion_del_juez:orden`: 89, ≤ 900; `evals:boe-legislacion:duracion_del_juez:herramienta`: 100, ≤ 900 |
| SC-002 | T014(hecha) | — | tareas hechas | — |
| SC-003 | — | — | SIN TAREA | — |
| SC-004 | T001(hecha) T005(hecha) T014(hecha) | fecha-version.txtar  | comprobado por su control | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFecha`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-version.txtar`, en make ci (verde) |
| SC-005 | T001(hecha) T014(hecha) | fecha-main.txtar  | comprobado por su control | `ci:internal/app/testdata/script/h20-fecha-main.txtar`, en make ci (verde); `ci:internal/source/boe/fuente_test.go:TestGolden`, en make ci (verde) |
| SC-006 | T002(hecha) T007(hecha) T014(hecha) | — | comprobado por su control | `ci:internal/graph/almacen_test.go:TestApplySoloLoNuevo`, en make ci (verde); `ci:internal/graph/integracion_test.go:TestIntegracionLecturaFechada`, en make ci (verde); `ci:internal/app/grafo_test.go:TestLecturaFechadaEnElGrafo`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-grafo.txtar`, en make ci (verde) |
| SC-007 | T005(hecha) T014(hecha) | — | comprobado por su control | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFechaYLaCache`, en make ci (verde); `ci:internal/app/testdata/script/h20-fecha-offline.txtar`, en make ci (verde) |
| SC-008 | T012(hecha) T014(hecha) | — | comprobado por su control | `ci:internal/source/boe/fecha_test.go:TestArticuloEnUnaFecha`, en make ci (verde); `ci:internal/source/boe/grafo_test.go:TestObservadoDeBoe`, en make ci (verde); `ci:internal/app/grafo_test.go:TestLecturaFechadaEnElGrafo`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeBoeLegislacion`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills`, en make ci (verde) |
| SC-009 | T010(hecha) T014(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills`, en make ci (verde) |
| SC-010 | T008(hecha) T010(hecha) T014(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeBoeLegislacion`, en make ci (verde); `ci:internal/evals/juzgar_test.go:TestJuzgarLecturasFechadas`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills`, en make ci (verde) |
| SC-011 | T006(hecha) T011(hecha) T014(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| SC-012 | T013(hecha) | — | tareas hechas | — |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`5e3c689`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `95f4949` fix(H20): motivos de la revisión final: lo vio la ronda 2 (juez A: aprobado; juez B: aprobado). Toca `internal/source/boe/articulo.go`, `internal/source/boe/fecha.go`, `internal/source/boe/ids.go`, `internal/source/boe/ids_test.go`, `cierre.md`, `tasks.md`.
- `093aa22` docs(H20): veredictos de la revisión final: solo registros de `gates/`.
- `77211a9` docs(H20): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/024-h20-boe-articulo-fecha/quickstart.md`. Suite de aceptación congelada: `specs/024-h20-boe-articulo-fecha/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `884eaf7f`: 10 h 6 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh 884eaf7f`.
