# 0055 · La caída de la eval 04 es de la instalación completa, y las descripciones cortas se aplican con un criterio fijado antes de medirlas

- **Estado**: propuesta (2026-11-17); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-11-17
- **Hito**: H41.3. Se escribe después del run `8da230a5` (#270) y de su fusión, fuera del run, en una sesión
  interactiva. Usa las cifras del sondeo de la persona (`evidencias/sondeos-20261117/`). Es el ADR que pide la entrada
  de H41.3 en «Aceptación»: la tabla «va a un ADR que dice» qué pasa con la instalación, el presupuesto y el remedio.
  No cambia ningún umbral, límite ni tope, ni el job de evals, ni ninguna skill.

## Contexto y problema

El ADR 0054 atribuyó a la instalación completa la caída de la eval 04 de `normativa` en modo herramienta, como
hipótesis de trabajo. En esa caída, la sesión pide `boe_articulo` y responde sin ningún `tool_use` de `Skill`. El ADR
dejó a H41.3 tres comparaciones en un sondeo de cuatro filas de 160 sesiones:
- **la instalación**: `sola` frente a `control`;
- **el presupuesto del listado de skills**: `diagnostico`, con `SLASH_COMMAND_TOOL_CHAR_BUDGET=1000000`, frente a
  `control`;
- **un remedio que solo toca las cuatro `description` del producto**: `remedio`, con la variante
  `producto-descripciones-cortas`, frente a `control`.

La entrada fijó antes de medir cómo se lee cada comparación. El remedio lo aplica el hito siguiente si cumple dos
condiciones frente al control:
- queda por debajo en «No activan ninguna», con intervalos que no se tocan;
- su «Activan otra skill» no queda por encima.

La segunda condición está porque la investigación del plan (D2) temía que el remedio funcionase llevando la
activación a otra skill en lugar de a `normativa`. Por eso los textos cortos conservan «antes que las demás skills de
kitlegal» en `consulta-juridica`.

**La versión de Claude Code es 2.1.326, no la 2.1.284 de la entrada.** Lo decidió Jorge el 2026-11-16, para medir con
la versión del uso real. Desde la #271, el job mide con esa misma versión (ADR 0031; `docs/USO.md`, entrada de ese
día). `SLASH_COMMAND_TOOL_CHAR_BUDGET` está en el binario de las dos versiones.

## Lo medido

Las condiciones del sondeo:
- `claude-sonnet-5-5` y Claude Code 2.1.326;
- la eval 04 en modo herramienta, 160 sesiones por fila, sobre `e12146f1`;
- 640 sesiones, todas terminadas y sin ningún `sesion.err` con contenido.

Los intervalos son de Wilson al 95 % y las p, de Fisher unilateral.

| Fila | Activan `normativa` antes | Activan otra skill | No activan ninguna | Abren `con_herramientas.md` | Abren `normas.md` |
|---|---|---|---|---|---|
| `sola` | 159 | 0 [0,0–2,3 %] | 1 [0,1–3,5 %] | 116 | 117 |
| `control` | 146 | 0 [0,0–2,3 %] | **14** [5,3–14,2 %] | 124 | 130 |
| `diagnostico` | 153 | 0 [0,0–2,3 %] | **7** [2,1–8,8 %] | 137 | 142 |
| `remedio` | 160 [97,7–100,0 %] | **18** [7,2–17,1 %] | 0 [0,0–2,3 %] | 145 | 148 |

- **Las 22 caídas son del tipo de las del job.** Ocurren en `sola` 82, en 14 de `control` y en 7 de `diagnostico`.
  Cada una pide `boe_articulo` con `ToolSearch`, lo llama y responde sin activar ninguna skill. La de `sola` es la
  primera caída con `normativa` sola instalada en los sondeos de H40 a H41.3.
- **La instalación.** `control` queda por encima de `sola`, y los intervalos no se tocan (p ≈ 0,0004). El control
  reproduce la caída, así que las otras dos comparaciones deciden.
- **El presupuesto.** `diagnostico` cae a la mitad que `control`, pero los intervalos se tocan (p ≈ 0,087). Con un
  listado que no se puede recortar siguen cayendo 7 de 160. Tampoco se separa de `sola` ([2,1–8,8 %] frente a
  [0,1–3,5 %]; p ≈ 0,033), así que el sondeo no demuestra que el presupuesto sea la causa, pero tampoco descarta que
  contribuya.
- **El remedio, en «No activan ninguna».** `remedio` queda por debajo de `control`, y los intervalos no se tocan
  (p ≈ 0,00005).
- **El remedio, en «Activan otra skill».** Queda por encima de `control`, y los intervalos no se tocan (p ≈ 0,000002).
  Las 18 sesiones activan `consulta-juridica` y `normativa`, las dos antes de la primera herramienta. Abren
  `con_herramientas.md` y `normas.md` en el turno 3, uno más tarde que casi todas las demás, porque la skill de
  entrada ocupa un turno. En las otras 142 sesiones de `remedio`, `normativa` se activa sola y no cae ninguna. El remedio no funciona
  llevando la activación a otra skill: `normativa` se activa antes en las 160.
- **Lo que mide el control.** En `control` no activa `consulta-juridica` ninguna sesión, aunque su descripción dice que
  se usa «antes que las demás skills de kitlegal» y la entrada de una pregunta es suya (ADR 0048). Con la descripción
  corta, la activan 18. En las 18, la skill hace lo que su descripción y el ADR 0048 piden.
- **La descripción corta de `consulta-juridica` cambia más que la longitud.** Pierde varias cosas
  (`specs/037-h41-3-por-que-normativa/contracts/variante.md` §2):
  - la lista de ejemplos, que incluía «un artículo de una ley»;
  - «qué tipo de pregunta es» y «qué cubre la base y qué no»;
  - «también cuando el agente cree saber la respuesta», que no cabía en 450 caracteres. D2 conserva esa frase en
    `normativa` («aunque el agente crea saberlo») porque es lo que mide la eval 04.

  Además, «cualquier pregunta sobre una situación legal» pasa a «toda pregunta legal». Quitarla haría pensar en menos activaciones de `consulta-juridica`,
  no en más, así que el texto no explica las 18 por sí solo.

## Opciones consideradas

1. **Aplicar el remedio ya, leyendo las 18 sesiones como la skill de entrada que hace su trabajo.** Rechazada. La
   entrada de H41.3 fijó la regla antes de medir, y una regla fijada así se aplica como está. La condición de «Activan
   otra skill» estaba para detectar justo esto. Que esta vez no señale un defecto solo se sabe mirando las sesiones, y
   esa lectura se hace ahora, después de medir. Una regla nueva solo puede decidir sobre datos que todavía no existen.
   Es lo que hizo el ADR 0047: cambió una regla después de un sondeo y dejó la decisión a un sondeo nuevo, que no se
   repite para ver si sale en verde (punto 4).
2. **Fijar `SLASH_COMMAND_TOOL_CHAR_BUDGET` al instalar kitlegal.** Rechazada. No se demuestra que el presupuesto sea
   la causa, es un ajuste de Claude Code en el equipo de quien instala (fuera, por la entrada de H41.3), y el remedio
   quita la caída sin él.
3. **Acortar las descripciones de otra forma**, por ejemplo, devolviendo a `consulta-juridica` la frase que perdió o
   quitándole la de «antes que las demás». Rechazada: sería un remedio sin medir, y lo medido es la variante tal como
   está.
4. **Otro sondeo de laboratorio con un criterio nuevo, antes de tocar el producto.** Rechazada. Costaría lo mismo que el
   sondeo de aceptación del hito que aplica el remedio, que mide las mismas dos filas y además valida las evals de las
   cuatro skills en el job.
5. **Aplicar el remedio en un hito propio, con un criterio de aceptación que este ADR fija antes de medir y un sondeo
   nuevo que lo decide.** Elegida.

## Decisión

1. **La caída de la eval 04 es de la instalación completa.** Lo demuestra el criterio que fijó la entrada de H41.3, y
   el punto 1 del ADR 0054 deja de ser una hipótesis. Con las cuatro skills instaladas y Claude Code 2.1.326,
   `normativa` no se activa en 14 de 160 sesiones de la eval 04 en modo herramienta [5,3–14,2 %].
2. **No se demuestra que la causa sea el presupuesto del listado.** Nada se fija en `SLASH_COMMAND_TOOL_CHAR_BUDGET`,
   ni en el producto ni en `kitlegal skills install`. El mecanismo queda sin explicar, y el remedio no depende de él.
3. **El remedio no se aplica con la regla de H41.3, porque no la cumple.** La lectura de las 18 sesiones que da «Lo
   medido» no cambia la regla de este sondeo.
4. **Lo aplica H41.4**, el hito siguiente. Pone en `skills/` las cuatro `description` de
   `laboratorio/producto-descripciones-cortas/`, byte a byte, y no hace ningún otro cambio en las skills. Lo generado
   que dependa de ellas lo regenera el hito. Su aceptación es un sondeo nuevo en las condiciones del de H41.3
   (`claude-sonnet-5-5`, la versión de Claude Code del job y la eval 04 en modo herramienta), con **320 sesiones por
   fila** y dos filas:
   - **control**: `INSTALACION=producto`, lanzado desde un árbol en la base de la rama, para que instale `skills/` de
     `main`;
   - **producto**: `INSTALACION=producto`, lanzado desde la cabeza de la rama.

   Las dos filas se lanzan el mismo día, una detrás de otra.

   El criterio queda fijado aquí, antes de medir. **«No activan `normativa` antes de la primera herramienta»**, que es
   *n* menos «Activan antes», queda en `producto` por debajo de `control`, con intervalos que no se tocan. Ese recuento
   cuenta como fallo tres cosas:
   - la sesión que no activa ninguna skill;
   - la que activa solo otra;
   - la que activa `normativa` después de la primera herramienta.

   Es el desplazamiento que temía D2. «Activan otra skill» se da en la tabla y no decide: una skill que se activa junto
   a `normativa`, antes de la primera herramienta, no hace que `normativa` falle. Con los datos de H41.3, el recuento
   es 14 de 160 en `control` y 0 de 160 en `remedio`.

   Además, las evals de las cuatro skills salen en verde en el job, que las mide todas porque cambia una descripción
   (ADR 0045), y el job juzga la tercera persona de los cuatro textos (ADR 0046, punto 4).

   **Si no se cumple el criterio del sondeo, o si fallan las evals o el juicio de la tercera persona, H41.4 no se
   fusiona.** El sondeo no se repite para ver si sale en verde (ADR 0047, punto 4), y el ADR que lo lea dice qué se
   mide después.

## Consecuencias

- **Hasta que entre H41.4, el cierre de un hito puede salir en rojo por esta caída.** El job mide 3 sesiones de la
  eval 04 en modo herramienta en cada medición. Con la tasa del control, en torno al 9 %, sale en rojo en torno a
  1 de cada 4 mediciones. Se acepta, como en el ADR 0054, solo hasta el arreglo, y ningún hito la esquiva.
- **H41.4 va primero en la cola**, por delante de H33, porque es el que trata la caída. Su entrada sigue el punto 4 y
  llega en una propuesta de cambio aparte. Hasta que esté en `docs/roadmap/todo/`, el siguiente hito del repositorio
  sigue siendo H33.
- **El sondeo de H41.4 solo mide la activación de `normativa`, pero cambian cuatro descripciones.** Las de
  `jurisprudencia` y `plazos-y-territorio` también pierden disparadores, y su activación solo la protege el job. El job
  mide 3 sesiones por eval y medición, así que no ve una caída del 5 al 10 %: la de la eval 04 tardó cinco cierres en
  verse por eso. Es herencia de la entrada de H41.3, y se acepta como riesgo.
- **La activación de `consulta-juridica` pasa a ser algo que se mira.** En las evals de `normativa` con la instalación
  de hoy no la activa ninguna sesión, aunque su descripción lo pide. Si eso es un defecto de la skill de entrada no se
  decide aquí. Las evals de `consulta-juridica` del job ya miden su activación en las preguntas que son suyas.
- **`evidencias/sondeos-20261117/`** queda como la medida de partida de H41.4.

## Validación

- **Las cifras** salen de `evidencias/sondeos-20261117/tabla-normativa.md`, que imprime `make laboratorio-tabla`. La
  tabla regenerada desde las sesiones coincide byte a byte con la del lanzamiento.
- **El tipo de las 22 caídas** sale de los `tool_use` de su `sesion.jsonl`: `ToolSearch` (`select:` de
  `mcp__kitlegal__boe_articulo`, sola o con `mcp__kitlegal__boe_buscar`), `mcp__kitlegal__boe_articulo` y ningún
  `Skill`. Las sesiones se conservan fuera del repositorio, en `~/code/kitlegal-sondeos/h41-3/` de quien lanzó el
  sondeo.
- **Lo que abren las 18 sesiones de `remedio` que activan `consulta-juridica`** sale de las filas de detalle de la
  tabla: activaciones `consulta-juridica (antes), normativa (antes)`, y `con_herramientas.md` y `normas.md` en el
  turno 3.
- **La estadística.** Los intervalos son de Wilson al 95 %, y las p, de la prueba exacta de Fisher unilateral sobre
  las dos filas que se comparan. La potencia del criterio del punto 4 es la probabilidad binomial de que los
  intervalos de las dos filas no se toquen. Las dos son filas medidas: `producto` no se supone en 0.

  | Tasa real de `producto` | 160 por fila, control al 7,5 % | 160, control al 8,75 % | **320, control al 7,5 %** | **320, control al 8,75 %** |
  |---|---|---|---|---|
  | 0 % | 0,92 | 0,97 | 1,00 | 1,00 |
  | 0,6 %, la de `sola` | 0,68 | 0,82 | 0,98 | 1,00 |
  | 1,25 % | 0,47 | 0,64 | 0,91 | 0,97 |
  | la del control, sin mejora | 0,002 | — | 0,003 | — |

  Con 160 por fila, un remedio que funcionase como `sola` suspendería entre 1 de cada 3 y 1 de cada 5 sondeos. Es el
  sorteo que corrigió el ADR 0047. Con 320 por fila pasa 49 de cada 50 veces, y un remedio sin efecto, 3 de cada 1000.
  Con el control en el límite inferior de su intervalo de hoy, el 5,3 %, y `producto` en 0, pasa con probabilidad
  1,00 con 320 por fila, frente a 0,62 con 160.
