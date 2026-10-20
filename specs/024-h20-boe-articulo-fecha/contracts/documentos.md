# Contrato: los documentos (FR-060 a FR-063)

Qué pasaje cambia en cada documento y qué tiene que decir. Solo donde lo que dicen de `boe articulo`, de
`boe-legislacion`, de sus evals, de su formato o de sus umbrales queda falso o incompleto con el hito; cada cifra sale
del árbol tras el cambio (los tests de la fila de cada una) y no de memoria; ninguna promete una release, y ningún test
ata este texto al código (FR-062). No se tocan `docs/ROADMAP.md`, ningún ADR, `docs/USO.md`, `docs/HISTORIAL.md`,
`docs/SOURCES.md`, la web, `CLAUDE.md` ni los specs, planes o informes de otros hitos (FR-063; spec, «Fuera de alcance»).

## `CHANGELOG.md`, *Unreleased* (FR-060)

Una sección `### Añadido` delante de `### Cambiado`, con:

- `kitlegal boe articulo <norma> <bloque> --fecha AAAA-MM-DD` y el parámetro `fecha` de `boe_articulo`: la versión del
  bloque vigente en esa fecha, la de mayor fecha de vigencia que no es posterior a ella —nunca por la de publicación—;
  una fecha anterior a la primera vigencia sale con 3 y su mensaje nombra la fecha de vigencia de la primera
  redacción; `data` lleva además `fecha_pedida` y, si hoy rige otra, `redaccion_posterior`, con su fecha de vigencia y
  su norma modificadora; sin `--fecha`, nada cambia.
- Que una lectura fechada no cuenta como lectura del bloque en `graph check`: emite solo la `BloqueVersion` elegida y
  su texto, y no cambia lo que el grafo ya guarda.
- `boe-legislacion` v0.2: con una pregunta con fecha, cita la redacción de esa fecha con las líneas
  `⚠ REDACCIÓN A FECHA:`, `⚠ REDACCIÓN DE HOY:` y `⚠ SIN REDACCIÓN VIGENTE:`.
- Las dos evals con fecha (22 y 23), el formato que las admite (`fecha` en un comando de bloque, `redaccion_fechada`)
  y el umbral `redaccion_sin_fechar:<modelo>:<modo>`: cuenta, sobre las respuestas del modelo que decide en las evals
  con fecha, las que no leen el bloque en su fecha o no llevan las líneas esperadas; con 0 y decide.

Las secciones de las versiones publicadas no cambian.

## `README.md` (FR-061)

- **«Qué sabe hacer tu agente con kitlegal»**: la frase de `boe-legislacion` y `legal-core` gana que, si la pregunta
  lleva una fecha —la del acto, la de los hechos—, cita la redacción que regía ese día y avisa si hoy es otra. Es lo
  único que el hito deja incompleto en el README.
- «Lo que viene» («Pronto: citas verificadas») habla de resolver una cita en lenguaje natural y revisar las de un
  escrito, que no entrega este hito: no cambia.

## `CONTRIBUTING.md` (FR-061)

| Pasaje | Hoy dice | Pasa a decir |
|---|---|---|
| Tabla del formato común de eval, fila `comandos` | la forma bloque, `applet`, `norma`, `bloque` | además `fecha` opcional, AAAA-MM-DD: con ella, solo la cumple `articulo` con `--fecha` esa fecha, con 0 o 3; sin ella, solo una lectura sin `--fecha` |
| Ídem, fila nueva `redaccion_fechada` | — | sus claves (contracts/evals-redaccion-fechada.md §1.2), qué líneas pide y que la prohíben `activa: false` y `sin_binario_ni_servidor` |
| Ídem, fila `sin_binario_ni_servidor` | la lista de claves que no lleva | con `redaccion_fechada` |
| Las reglas del conjunto de `boe-legislacion` | «exactamente diez positivas que deciden, de materias distintas…» | que las evals con `redaccion_fechada` se apartan de las positivas, como la eval sin binario ni servidor |
| El juicio de una sesión y «Una llamada a una herramienta cuenta como la invocación de su orden» | un comando de bloque lo cumple con su norma y su bloque | con `fecha`, solo una lectura con esa fecha, con 0 o 3; sin ella, ninguna lectura con fecha |
| «Qué respuestas se juzgan» | 54 por modo en `boe-legislacion` | 60 |
| El trabajo del job de evals | veintiuna evals, 198 sesiones (96, 96 y 6), doce que deciden, 111 respuestas juzgadas (54, 54 y 3) | veintitrés evals, 222 sesiones (108, 108 y 6), catorce que deciden, 123 respuestas (60, 60 y 3); las cifras, las de `TestDefinicionDelJob` |
| El peor caso del trabajo de `boe-legislacion` | 97, 96 y 6 sesiones a 5, 11 909 s; 54, 54 y 3 respuestas, 5 420 s; 17 329 s, 288,8 minutos; «con dos evals nuevas serían 23 449 s» a 4 | 109, 108 y 6 sesiones a 5, 12 997 s; 60, 60 y 3 respuestas, 72, 72 y 4 tandas de votos, 5 980 s; 18 977 s, 316,3 minutos, el literal de `TestDefinicionDelJob`; lo de 4 a la vez, como lo que era hasta H20 |
| Los umbrales del informe: cuándo existen los de las respuestas | sentencias, plazo o cobertura | también una eval con `redaccion_fechada` |
| Ídem, «En `boe-legislacion` son trece, nueve de ellos decidiendo» y su tabla | trece, nueve; `(54)`; `(108)`; «los tres del modo orden y los tres del modo herramienta» | quince, once; la fila de `redaccion_sin_fechar:<modelo>:<modo>` con su medida, su `total` (6) y su motivo; `(60)`; `(120)`; los cuatro de cada modo |
| La entrega al grafo («Hoy emiten `boe articulo`…») | qué emite cada verbo | que `boe articulo --fecha` emite solo la `BloqueVersion` de la versión elegida y su texto, con `SoloLoNuevo`, y no es una lectura del bloque |

## `docs/WORKFLOW.md` (FR-061)

Un solo pasaje: el de `medir_cierre`, en «Cierre en la plataforma e informe final» (`docs/WORKFLOW.md:139`), que enumera los umbrales que deciden en
`boe-legislacion`. Con FR-054 y FR-056 deja de ser cierto. `grep -n` de `boe-legislacion`, `boe articulo`,
`sin_activar` y `duracion_de_las_sesiones` en esta sesión da solo esa línea y la 22, que nombra la carpeta
`evals/boe-legislacion/juez/` de la evidencia del ADR 0037, que el hito no toca; las demás líneas con «umbral» hablan
de los umbrales del workflow o de cualquier skill, que el hito no cambia.

| Pasaje | Hoy dice | Pasa a decir |
|---|---|---|
| `medir_cierre`: los umbrales que deciden en `boe-legislacion` | «en `boe-legislacion` deciden nueve: por cada modo, `sin_activar:<modelo>:<modo>`, `duracion_de_las_sesiones:<modo>` y `duracion_del_juez:<modo>`; …» | «deciden once: por cada modo, `sin_activar:<modelo>:<modo>`, `redaccion_sin_fechar:<modelo>:<modo>` (desde H20, sobre las respuestas de las evals con fecha), `duracion_de_las_sesiones:<modo>` y `duracion_del_juez:<modo>`; …», en el orden de `umbrales` (contracts/evals-redaccion-fechada.md §3) |

El resto de la frase —`afirma_lo_no_leido:<modelo>`, los dos de la medida del juez, los que solo se publican y los que
ya no se publican— sigue siendo cierto y no se reescribe (FR-062).

## `internal/source/boe/doc.go`

El comentario de paquete gana un párrafo: `articulo` con `--fecha`, que `refs/boe.py` no tiene, elige la versión de
una fecha y la guarda en su propia entrada (`fecha.go`). La tabla de notas no cambia.
