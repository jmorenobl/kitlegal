# 0031 · El modelo que decide las evals es el id al que resuelve `sonnet` en la versión de Claude Code del job

- **Estado**: aceptada
- **Fecha**: 2026-09-30
- **Hito**: transversal, tras H7.3 y el ADR 0030 y antes de H7.4. Sustituye al ADR 0016 en el punto 1 de su decisión
  (qué modelo decide y qué modelos informan); el resto del ADR 0016 sigue vigente: repeticiones y umbral por serie,
  evals informativas, extracción de la cita y disparador por cambios.

## Contexto y problema

El ADR 0016 decidió que el job de evals lo decide «el modelo del uso real» de la skill, fijado por su id:
`MODELO_DE_EVALS: claude-sonnet-5`, con `claude-haiku-4-5-20251001` como límite inferior informativo. Quien usa las
skills en Claude Code no escribe un id: elige el alias `sonnet`. Y el alias ya no resuelve a Sonnet 5:

- En los transcripts del proyecto, la última respuesta de `claude-sonnet-5` es del 2026-09-28 a las 12:33 (Claude Code
  2.1.283) y la primera de `claude-sonnet-5-5`, de las 22:15 del mismo día (2.1.284): 1 188 turnos del primero, todos
  con 2.1.283 o anteriores, y 206 del segundo, todos con 2.1.284. El run de H7.3 (`f41e85d1`) ya ejecutó sus pasos de
  análisis, cuyo rol pide `sonnet@high`, con `sonnet-5-5` (`scripts/coste-run.sh f41e85d1`).
- **La resolución del alias es del cliente, no del servidor.** La misma orden,
  `npx -y @anthropic-ai/claude-code@<versión> -p --model sonnet --output-format json 'Responde solo: ok'`, con la
  misma cuenta y en el mismo minuto, da en `modelUsage` `claude-sonnet-5` con 2.1.270 y con 2.1.283, y
  `claude-sonnet-5-5` con 2.1.284.
- **El job fija `VERSION_DE_CLAUDE_CODE: 2.1.270`**, que no conoce el modelo nuevo: con `--model claude-sonnet-5-5` la
  sesión corre con él, pero el cliente avisa con `[claude-code:unrecognized_model]`.

Así que hoy el veredicto del job habla de un modelo con el que ya no se usa la skill, que es justo lo contrario de la
premisa del ADR 0016. La última medida con el modelo anterior es el cierre de H7.3
(`specs/013-h7-3-el-umbral-de/gates/evals/boe-legislacion.json`, Claude Code 2.1.270): Sonnet 5, todas las series que
deciden 3 de 3 y 1 de 51 respuestas con una expresión prohibida; Haiku 4.5, 0 de 30; 492 s; `aprobado`.

**Un defecto del control del modelo.** `exigirElModeloPedido` (`internal/evals/juzgar.go`) da por buena una sesión si el
id que declara *empieza* por el pedido, para admitir la fecha de la versión (`claude-haiku-4-5` →
`claude-haiku-4-5-20251001`). `claude-sonnet-5` es prefijo de `claude-sonnet-5-5`, así que una sesión que pidió Sonnet 5 y
corrió con Sonnet 5.5 pasaba el control, y el informe habría publicado como medida de un modelo lo que hizo el otro.

## Medidas

Sondeo comparado (`TestSondeo`, el de `make evals-sondeo`, con la misma preparación y el mismo juez que el job, sin la
traza de strace), el 2026-09-30, sobre `941b24f` (`main`), con Claude Code 2.1.284 en macOS, las 19 evals de
`boe-legislacion` y las 3 de `legal-core`, 3 repeticiones y la concurrencia del job (4 y 1). Las 66 sesiones de cada
modelo declaran el id pedido. «Estado de lo comprobado» cuenta las respuestas, de las 51 que activan la skill, cuya
primera línea anuncia lo comprobado o cómo se leyó —«no trae avisos de vigencia», «el texto no ha cambiado desde una
consulta anterior», «Lo he leído en el BOE consolidado», «Con esto ya tengo suficiente para contestar»— y que la lista
de expresiones prohibidas no ve.

| Eval (`boe-legislacion`) | Decide | Sonnet 5 | Sonnet 5.5 |
|---|---|---|---|
| 01 a 03, 05 a 12 | sí | 3 de 3 cada una | 3 de 3 cada una |
| 04 LGT, prescripción | sí | 3 de 3 | **1 de 3** |
| 13, 14, 16, 17, 18 | no | 3 de 3 cada una | 3 de 3 cada una |
| 15 IRPF, por materia | no | 2 de 3 | 3 de 3 |
| 19 LCSP, redacción cambiada | no | 1 de 3 | 3 de 3 |
| Respuestas con expresión prohibida | | 3 de 51 (5,9 %) | 0 de 51 (0,0 %) |
| Primera línea con el estado de lo comprobado, fuera de la lista | | 3 de 51 | 6 de 51 |
| Duración (preparación y 57 sesiones) | | 296 s | 217 s |
| `legal-core`, 01 a 03 | sí | 3 de 3 cada una; 99 s | 3 de 3 cada una; 70 s |

- **Sonnet 5.5 no activa la skill en la 04.** En dos de sus tres sesiones no llama a `Skill` y responde de memoria
  («Contesto de memoria y no he consultado el texto consolidado en el BOE»). Son las dos únicas sesiones, de las 102
  de los dos modelos que deben activar la skill, que no la activan. La pregunta nombra la norma y el artículo.
- **Las expresiones prohibidas desaparecen, el anuncio cambia de forma.** Sonnet 5 lleva «conversación anterior» en dos
  respuestas de la 19 y «respondo con el texto» en una de la 15. Sonnet 5.5 no lleva ninguna. Pero seis de sus primeras
  líneas dicen el estado de lo comprobado con palabras que la lista no tiene: cinco con «no trae avisos de vigencia» o
  equivalentes (03, 07 y las tres de la 13) y una con «Lo he leído en el BOE consolidado» (16). Las tres de Sonnet 5 son
  «No hay avisos de vigencia ni cambios de redacción. Con esto ya tengo suficiente…» (14), «I'll responder con el texto
  ya obtenido…» (02) y «…lo confirmo con la sección…» (14).
- **Sonnet 5 con Claude Code 2.1.284 no reproduce el cierre de H7.3**, hecho con 2.1.270: 3 de 51 frente a 1 de 51, y la
  19 con 1 de 3 frente a 3 de 3. Con tres repeticiones, un cambio de 2 respuestas de 51 cabe en la variación de una
  tirada, así que no se atribuye a la versión del cliente.

## Opciones consideradas

Qué modelo decide:

1. **Seguir con `claude-sonnet-5`.** Rechazada: el veredicto dejaría de hablar del modelo con el que se usa la skill.
   Que Sonnet 5 pase hoy todas las series que deciden y Sonnet 5.5 no, no cambia qué se mide: un veredicto verde sobre
   un modelo que nadie recibe con `sonnet` no dice nada de la skill.
2. **El alias `sonnet` en `MODELO_DE_EVALS`.** Rechazada. Es reproducible, porque la resolución es del cliente y el job
   fija su versión. Pero con 2.1.270 el alias daría Sonnet 5, así que no bastaría. Además, subir Claude Code por
   cualquier otro motivo cambiaría el modelo que decide sin que el cambio lo dijera. Y el informe nombra por id los
   umbrales (`expresiones_prohibidas:<modelo>`, contrato del ADR 0029) y el modelo que exige a cada sesión.
3. **`claude-sonnet-5-5` por su id, con Claude Code 2.1.270.** Rechazada: esa versión no reconoce el modelo, y el uso
   real lo ejecuta un cliente que sí lo reconoce.
4. **`claude-sonnet-5-5` por su id y Claude Code 2.1.284**, la primera versión en la que `sonnet` resuelve a ese id.
   Elegida: es el modelo y el cliente del uso real, reproducible, y un cambio de modelo es siempre una línea del diff.

Sonnet 5:

5. **Como modelo informativo.** Rechazada. No es un límite inferior: los dos Sonnet fallan en cosas distintas (Sonnet 5
   en expresiones y en la 19; Sonnet 5.5 en la activación de la 04), y el límite inferior ya lo da Haiku 4.5. Claude
   Code se actualiza solo, así que quien recibe Sonnet 5 con el alias desaparece en días. Quien lo fija por su id
   eligió otro modelo, y ese uso no lo mide el job.
6. **Fuera del job.** Elegida.

Cómo se detecta el próximo cambio del alias:

7. **Una sesión con el alias en cada ejecución del job, comparada con `MODELO_DE_EVALS`.** Rechazada por
   proporcionalidad: dentro del job el alias solo cambia si cambia `VERSION_DE_CLAUDE_CODE`, que es un cambio del
   propio fichero, y ningún requisito pide detectarlo sin persona.
8. **Una regla con una orden que la comprueba**, abajo. Elegida.

## Decisión

1. **Decide `claude-sonnet-5-5`**, fijado por su id en `MODELO_DE_EVALS`, con `VERSION_DE_CLAUDE_CODE: 2.1.284`.
   **Sonnet 5 sale del job.** `claude-haiku-4-5-20251001` sigue en `MODELOS_INFORMATIVOS_DE_EVALS` como límite
   inferior publicado.
2. **Invariante.** `MODELO_DE_EVALS` es el id al que resuelve el alias `sonnet` en `VERSION_DE_CLAUDE_CODE`, y los dos
   cambian juntos. Se comprueba con la clave `claude-sonnet-*` de `modelUsage` en
   `npx -y @anthropic-ai/claude-code@<VERSION_DE_CLAUDE_CODE> -p --model sonnet --output-format json 'Responde solo: ok'`.
   Toda propuesta de cambio que toque `VERSION_DE_CLAUDE_CODE` la ejecuta y lo dice.
3. **Detección.** Antes de lanzar cada hito, quien lo lanza ejecuta
   `claude -p --model sonnet --output-format json 'Responde solo: ok'` con su Claude Code, que es el del uso real. Si la
   clave `claude-sonnet-*` de `modelUsage` no es `MODELO_DE_EVALS`, el hito espera a una propuesta de cambio aparte que
   suba el par a esa versión y a ese id, con un sondeo comparado como el de arriba y su entrada en `docs/USO.md`. Seguir
   el alias es aplicar este ADR, no decidir otra vez: solo hace falta un ADR nuevo si la decisión cambia. Una segunda
   señal ya existe: la tabla de `scripts/coste-run.sh` de cada run muestra a qué resuelve `sonnet` en los pasos de
   análisis.
4. **El control del modelo de la sesión exige el id pedido**, tal cual o seguido de `-` y la fecha de la versión
   (`AAAAMMDD`), y nada más (`esElModeloPedido`, con su test). Así el informe distingue Sonnet 5 de Sonnet 5.5. Los
   transcripts sintéticos de los tests del informe y del sondeo declaraban `claude-haiku-4-5` para un pedido
   `claude-haiku`, y con eso se daba por probado que se admite la fecha. Era la misma forma que el defecto, y ahora
   declaran `claude-haiku-20251001`.
5. **Si el modelo nuevo no cumple, la elección no cambia.** Es un defecto de la skill, y lo arregla H7.4. Con las
   medidas de arriba, el job con Sonnet 5.5 da previsiblemente `fallo` por la eval 04, que decide (1 de 3 en el
   sondeo: dos sesiones sin activar la skill), y la ejecución de apertura lo confirma (abajo). Además, las seis
   primeras líneas con el estado de lo comprobado son el ruido que H7.4 tiene que quitar aunque la lista no las vea.
   La medida del job de esta propuesta de cambio es la línea de base de H7.4.

## Medición del job en la propuesta de cambio

Ejecución de apertura 36712391835 de la PR #88, sobre `ee8e7ee` (el producto de esta decisión; los commits posteriores
solo tocan `docs/`), con Claude Code 2.1.284:

- **`boe-legislacion`: veredicto `fallo`**, por un solo motivo de la skill: la 04 con `claude-sonnet-5-5`, 1 de 3.
  Las sesiones 01 y 03 no activan la skill, así que no leen el bloque `a66` ni lo citan. Todas las demás series de
  Sonnet 5.5 dan 3 de 3, también la 19 y las informativas.
- **`umbrales`, todos cumplidos**: `expresiones_prohibidas:claude-sonnet-5-5` 0 de 51 (`decide: true`);
  `expresiones_prohibidas:claude-haiku-4-5-20251001` 0 de 30 (`decide: false`); `duracion_de_las_sesiones` 421 s de
  900 (`decide: true`). `red` vacío, ninguna sesión sin medir y ningún reintento por límite de ritmo.
- **Modelos de sesión**: `modelos_de_sesion` es `claude-haiku-4-5-20251001` y `claude-sonnet-5-5`. Las 57 sesiones
  pedidas con `claude-sonnet-5-5` lo declaran en `modelo_de_la_sesion`, y las 36 de Haiku, su id.
- **Haiku 4.5**: la 06, 2 de 3 (una sesión no lee ni cita el bloque `a17`); las demás, 3 de 3.
- **Primera línea con el estado de lo comprobado**, fuera de la lista: 6 de 51 de Sonnet 5.5 (01, dos; 03, una; 13,
  las tres), lo mismo que en el sondeo.
- **`legal-core`: `aprobado`**, 198 s.

Es la línea de base de H7.4: la activación de la 04 hace fallar el veredicto, y las seis primeras líneas son el ruido
que no ve ningún control.

## Consecuencias

**A favor.** El veredicto vuelve a hablar del modelo con el que se usa la skill, y cada sesión lo demuestra: el control
del modelo ya no confunde un id con otro que empieza igual. Cambiar de modelo sigue siendo un diff explícito de
`evals.yml`, ahora atado a la versión del cliente que lo resuelve, con una orden que cualquiera ejecuta para
comprobarlo. H7.4 mide desde el principio con el modelo que va a recibir quien use la release.

**En contra, y asumido.**

- **`main` queda con el job de evals en rojo hasta H7.4**: la ejecución de apertura confirma lo que midió el sondeo. Es lo mismo que el ADR 0029 asumió con H7.2: el defecto es de la skill con el modelo del uso real, y taparlo
  midiendo con el modelo anterior sería peor. No hay release hasta H7.4.
- **La detección depende de una persona** que ejecuta una orden antes de cada hito. Si no la ejecuta, el job sigue
  midiendo con un modelo reproducible, que solo ha dejado de ser el del uso real, y la señal de `coste-run.sh` lo
  delata en el siguiente run.
- **Subir Claude Code en el job** cambia a la vez la versión que el research de H5 verificó (V1-V13, con 2.1.270). El
  sondeo ejecuta las mismas banderas, el mismo `stream-json` y el mismo cargador de skills con 2.1.284; la traza de
  strace solo la comprueba la ejecución del job en la plataforma.
- **La tabla de precios de `scripts/coste-run.sh`** casa `claude-sonnet-5-5` con los precios de Sonnet 5. Es informativa
  y no se toca aquí.
- **Las paráfrasis del estado de lo comprobado** siguen sin control mecánico. Qué hacer con ellas lo decide H7.4, no
  este ADR: la lista de expresiones y los umbrales son de la skill.
