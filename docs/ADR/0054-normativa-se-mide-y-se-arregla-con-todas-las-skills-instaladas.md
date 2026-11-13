# 0054 · La caída de la eval 04 se atribuye a la instalación completa: se mide y se arregla con todas las skills instaladas

- **Estado**: propuesta (2026-11-13); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-11-13
- **Hito**: H41.2, escrito tras el run `f40bea97` (#256) y su fusión, fuera de él, en una sesión interactiva, con las
  cifras del sondeo de la persona (`evidencias/sondeos-20261112/`). Es el ADR que pide la entrada de H41.2
  («Aceptación»: la tabla «va a un ADR que dice qué se hace con la caída»). No cambia ningún umbral, límite ni tope,
  ni el job de evals, ni ninguna skill.

## Contexto y problema

En el job de evals, la eval 04 de `normativa` (`04-lgt-prescripcion`) cae en modo herramienta con
`claude-sonnet-5-5`. La sesión pide `boe_articulo` sin ningún `tool_use` de `Skill`, y `sin_activar`, cuyo umbral es
cero, pone el job en rojo. Hay cinco caídas registradas, todas en esa eval y en ese modo:
- dos en los cierres de H32;
- una en la suite de la release v0.9.0;
- la `-01` del primer cierre de H41;
- la `-02` de la primera medición del cierre de H41.2, la primera con la lista de activaciones: `activaciones: []`,
  ninguna skill activada, con las cuatro del producto instaladas y Claude Code 2.1.284.

En los sondeos de la persona no caía ninguna. La D27 de H32 y la D21 de H41 dejaron la serie a una persona: o un ADR
sobre el umbral, o un hito que midiera la activación en modo herramienta. H41.2 midió. Entre el job y los sondeos
había dos diferencias:
- la versión de Claude Code: 2.1.284 en el job, 2.1.317 y 2.1.319 en los sondeos;
- las skills instaladas: el job instala las cuatro del producto, como `kitlegal skills install`; el arnés instalaba
  solo la variante de la skill que se medía.

H41.2 dio al arnés la instalación del producto (`INSTALACION=producto`) y una sola eval (`EVAL=04`). Su sondeo cruza
las dos diferencias en cuatro filas.

## Lo medido

`claude-sonnet-5-5`, la eval 04, modo herramienta, 40 sesiones por fila, 160 en total, todas terminadas, sobre
`8191a90e`. Intervalos de Wilson al 95 %.

| Fila | Claude Code | Skills instaladas | Activan `normativa` antes | Activan otra | No activan ninguna |
|---|---|---|---|---|---|
| `sola-2.1.284` | 2.1.284 | `normativa` | 40 | 0 | 0 [0,0–8,8 %] |
| `todas-2.1.284` | 2.1.284 | las cuatro | 38 | 0 | **2** [1,4–16,5 %] |
| `sola-2.1.322` | 2.1.322 | `normativa` | 40 | 0 | 0 [0,0–8,8 %] |
| `todas-2.1.322` | 2.1.322 | las cuatro | 36 | 0 | **4** [4,0–23,1 %] |

- **Las seis caídas son del mismo tipo que las del job.** Las seis sesiones cargan la herramienta con `ToolSearch`
  (`select:mcp__kitlegal__boe_articulo`), leen el art. 66 de la LGT con `boe_articulo` y responden sin activar
  ninguna skill, tampoco `consulta-juridica` (`evidencias/sondeos-20261112/LEEME.md`). Dan el plazo y la cita, pero
  sin el protocolo de la skill.
- **Solo caen con las cuatro skills instaladas**: 0 de 80 con `normativa` sola [0,0–4,6 %] y 6 de 80 con las cuatro
  [3,5–15,4 %]. Es la comparación que el sondeo se diseñó para hacer, y la prueba de Fisher unilateral da p ≈ 0,014.
- **De la versión de Claude Code, el sondeo no dice nada.** Caen 2 de 80 con 2.1.284 y 4 de 80 con 2.1.322 (p ≈
  0,34), pero con tan pocas caídas esa comparación no puede detectar ninguna diferencia, y no descarta que la versión
  influya.
- **El criterio del hito no se cumple en este sondeo.** El hito pedía intervalos que no se tocaran, y se tocan: por
  fila, y también juntando las filas por instalación (3,5 % frente a 4,6 %). **El sondeo no podía cumplirlo.** Con 40
  sesiones por fila y una tasa del 7,5 %, la probabilidad de separar los intervalos de una fila era menor de 0,01, y
  juntando las filas, con 80 por grupo, de 0,25. Es el error que el ADR 0047 corrigió en el sondeo de H31.1: un
  criterio fijado sin mirar cuántas respuestas lo deciden.

**Lo que suman los sondeos anteriores.** El modelo decide si activa una skill con su nombre y su descripción, antes
de leer ningún cuerpo. Las variantes de `normativa` de los sondeos de H40 y H41 tienen el nombre y la descripción de la
skill del producto, byte a byte, en las cabezas de los cuatro sondeos. En la eval 04 en modo herramienta, con una sola
skill instalada, activaron antes en todas sus sesiones:
- 40 en `evidencias/sondeos-20261104/`, con Claude Code 2.1.317;
- 30 en `evidencias/sondeos-20261106/`, con 2.1.317;
- 30 en `evidencias/sondeos-20261109/`, con 2.1.319.

Con las 80 de H41.2 son 0 de 180 [0,0–2,1 %], frente a 6 de 80 [3,5–15,4 %] con las cuatro: los intervalos no se
tocan, y Fisher da p ≈ 0,0007. Pero la agregación se hizo después de ver los datos y junta tres versiones de Claude
Code, así que no sustituye a una medida hecha para esto. Lo que hace es apuntar en la misma dirección que la
comparación diseñada y que las cinco caídas del job, todas con las cuatro skills instaladas.

**Un mecanismo documentado que encaja, sin demostrar.** El listado de skills que Claude Code pone en el contexto tiene
un presupuesto del 1 % de la ventana del modelo. Si se pasa, Claude Code quita descripciones, empezando por las de las
skills que menos se usan, y deja siempre los nombres (https://code.claude.com/docs/en/skills, «Skill descriptions are
cut short»). El listado de las sesiones tiene 20 skills con 2.1.284 y 21 con 2.1.322, contando las integradas de Claude
Code, frente a 17 y 18 con `normativa` sola. Las cuatro descripciones del producto suman 3313 caracteres, frente a los
887 de `normativa` sola. Que el listado pase del presupuesto, y qué descripción pierde, no se ha medido.

- **Si el presupuesto es la causa**, el recorte es el mismo en todas las sesiones de una fila: cada sesión del arnés
  empieza con un directorio de configuración nuevo, sin historial de uso. La caída de 1 de cada 13 encajaría entonces
  con una `normativa` que se queda sin descripción y que el modelo activa casi siempre solo por su nombre.
- **La otra explicación** es que, entre cuatro descripciones de kitlegal, la de `normativa` pese menos.

Las dos se distinguen midiendo: si el presupuesto es la causa, la caída desaparece con un presupuesto que no recorta
(`SLASH_COMMAND_TOOL_CHAR_BUDGET`); si es la competencia, no.

## Opciones consideradas

1. **Tratarla como ruido y dar un margen a `sin_activar`.** Rechazada. `sin_activar` es un hecho de la sesión, y su
   umbral es cero (ADR 0038, punto 1); una tasa solo va donde hay significado. Además, la caída no es ruido: es lo
   que pasa con la instalación que tiene quien usa kitlegal.
2. **Medir el job con una sola skill instalada.** Rechazada: dejaría de medir el producto como se instala, y la
   caída seguiría en el equipo de quien lo usa.
3. **Repetir el sondeo de H41.2 con más sesiones antes de hacer nada.** Rechazada. La confirmación de la instalación
   como causa cabe en el hito que busca el arreglo, con una fila de `normativa` sola en el mismo sondeo que el
   control y el remedio. Un sondeo solo para confirmarla costaría lo mismo que esa fila y retrasaría el arreglo.
4. **Pedir la activación en las instrucciones del servidor o en la skill de entrada.** Rechazada: el ADR 0050, punto
   3, decide que la activación no se asegura con texto fuera de la skill. Se midió con una sola skill instalada, pero
   sigue siendo una decisión cerrada: este ADR no la reabre, y H41.3 no la mide.
5. **Atribuir la caída a la instalación completa como hipótesis de trabajo, y medir la causa y el arreglo con esa
   instalación en un hito propio, con un sondeo dimensionado para separar los intervalos.** Elegida.

## Decisión

1. **La caída de la eval 04 se atribuye a la instalación completa.** Es la hipótesis que sostienen la comparación
   diseñada (0 de 80 frente a 6 de 80, p ≈ 0,014), la agregación de las sesiones con una sola skill (0 de 180) y las
   cinco caídas del job. No está demostrada con el criterio que fijó H41.2, y H41.3 la confirma o la descarta con una
   fila propia. Si se confirma, es un defecto del producto: con las cuatro skills instaladas, que es como las deja
   `kitlegal skills install`, `normativa` no se activa en torno a 1 de cada 13 sesiones de la eval 04 en modo
   herramienta, con una tasa de entre el 3,5 % y el 15,4 %.
2. **El umbral de `sin_activar` sigue en cero, y el job sigue instalando todas las skills.** El job mide lo que el
   producto hace, y el rojo que da esta caída tiene motivo. Ninguna eval cambia.
3. **La activación se mide con todas las skills instaladas.** Un sondeo que mida la activación de una skill del
   producto en modo herramienta usa `INSTALACION=producto`. Lo que midieron con una sola skill los sondeos de H39 a
   H41.1 vale para la lectura de las referencias, que empieza después de activar la skill. Su «40 de 40» en «Activan
   antes» es de esa instalación, no de la del producto.
4. **La causa y el arreglo los mide H41.3**, el primer hito de la cola, por delante de H33. Lo hace con la eval 04 en
   modo herramienta y Claude Code 2.1.284, la versión del job (ADR 0031), en cuatro filas de 160 sesiones del mismo
   sondeo:
   - **sola**: `normativa` sola instalada, para confirmar o descartar el punto 1;
   - **control**: las cuatro skills como están;
   - **diagnóstico**: el control con un presupuesto del listado que no puede recortarlo, para saber si la causa es el
     presupuesto. No es un remedio, porque quien instala kitlegal no lo fija;
   - **remedio**: el arreglo candidato, que solo toca lo que el producto controla, las descripciones de sus skills,
     dentro del ADR 0046.

   Con 160 sesiones por fila, una tasa del 7,5 % separa sus intervalos de los de 0 de 160 en 9 de cada 10 sondeos. Si
   el remedio queda por debajo del control en «No activan ninguna», con intervalos que no se tocan, lo aplica un hito
   siguiente. Lo decide el ADR que lea ese sondeo, como el 0053 con H41.1.

## Consecuencias

- **Hasta el arreglo, el cierre de un hito puede salir en rojo por esta caída.** Con 3 sesiones de la eval 04 en modo
  herramienta por medición y una tasa del 7,5 %, sale en rojo en torno a 1 de cada 5 mediciones, y las tres de un
  cierre, en torno a 1 de cada 100. Es lo que el ADR 0038 quería evitar, un job que hay que repetir hasta que sale
  limpio, y se acepta solo hasta que el arreglo entre. H41.3 es el primero de la cola porque es el hito que trata la
  caída. Los demás no la arreglan ni la esquivan: la medición siguiente la repite, y si las tres caen, la decide la
  persona.
- **`evidencias/sondeos-20261112/`** queda como la medida de partida de H41.3, con la agregación de este ADR.
- **`docs/SKILLS.md` §1** dice que la activación se mide con la instalación del producto (punto 3).
- **El arnés no cambia por este ADR.** Dar a una fila un presupuesto propio y otra instalación es entrega de H41.3.

## Validación

- **Las cifras de H41.2** salen de `evidencias/sondeos-20261112/tabla-normativa.md`, que imprime
  `make laboratorio-tabla`.
- **Las de los sondeos anteriores** salen de las filas `…/herramienta/04-lgt-prescripcion-*` de
  `evidencias/sondeos-20261104/tabla.md`, `evidencias/sondeos-20261106/tabla-normativa.md` y
  `evidencias/sondeos-20261109/tabla-normativa.md`.
- **La huella del nombre y la descripción** es el SHA-256 del `name` y la `description` del frontmatter de
  `skills/normativa/SKILL.md` y de `laboratorio/normativa/*/skills/normativa/SKILL.md`, en `be1dac4a`, `fb369637`,
  `7bca2cf8` y `8191a90e`, las cabezas de los cuatro sondeos.
- **El número de skills del listado** es la longitud de `skills` en el mensaje `system/init` de `sesion.jsonl`.
- **La estadística.** Los intervalos son de Wilson al 95 %, y las p, de Fisher unilateral. La potencia es la
  probabilidad binomial de que la fila dé las caídas con las que su intervalo deja de tocar el de 0 de *n*:
  - con 160 por fila hacen falta 8 de 160, cuyo límite inferior, 2,6 %, supera el 2,3 % del de 0 de 160;
  - con una tasa real del 7,5 %, la potencia es 0,92; con el 5 %, 0,55; y con el 3,5 %, el límite inferior de lo que
    midió H41.2, 0,20;
  - con una tasa del 7,5 %, el control no llega a 8 de 160 en 1 de cada 12 sondeos.
