# 0056 · El sondeo de H41.4 no cumple su criterio por un control que deriva, y el siguiente se decide con un tope y filas intercaladas

- **Estado**: propuesta (2026-11-18); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-11-18
- **Hito**: H41.4. Se escribe después del run `20123bb8` (#274), sin fusionarla, fuera del run, en una sesión
  interactiva. Usa las cifras del sondeo de aceptación que pide el ADR 0055, punto 4 (`evidencias/sondeos-20261118/`),
  y una segunda lectura independiente de los datos, hecha con `claude-fable-5-2`. Es el ADR que, según el punto 4 del
  0055, «lo lea» y diga qué se mide después. No cambia ningún umbral, límite ni tope del job de evals, ni ninguna
  skill.

## Contexto y problema

El ADR 0055 decidió que H41.4 pusiera en `skills/` las cuatro descripciones cortas de
`laboratorio/producto-descripciones-cortas/`. Antes de medir, fijó como aceptación un sondeo de dos filas de 320
sesiones de la eval 04 en modo herramienta, las dos con las cuatro skills instaladas:
- **control**, lanzado desde la base de la rama, con las skills de `main`;
- **producto**, lanzado desde la cabeza de la rama, con las descripciones cortas.

El criterio: en «No activan `normativa` antes de la primera herramienta», `producto` por debajo de `control`, con
intervalos de Wilson al 95 % que no se tocan. Si no se cumplía, H41.4 no se fusionaba, el sondeo no se repetía para ver
si salía en verde (ADR 0047, punto 4) y un ADR decía qué medir después.

El run `20123bb8` entregó la #274 en verde: CI, `codecov/patch` y las evals de las cuatro skills en el job, con el
juicio de la tercera persona de los cuatro textos sin fallos. Faltaba el sondeo.

## Lo medido

Las condiciones del sondeo:
- `claude-sonnet-5-5` y Claude Code 2.1.326, la versión del job;
- la eval 04 en modo herramienta, `INSTALACION=producto`, 320 sesiones por fila y concurrencia 5;
- `control` desde `ab951c22` y `producto` desde `77d6ed7c`, el mismo día, una fila detrás de otra;
- 640 sesiones, todas terminadas y sin ningún `sesion.err` con contenido.

| Fila | No activan `normativa` antes | Activan otra skill | No activan ninguna | Abren `con_herramientas.md` | Abren `normas.md` |
|---|---|---|---|---|---|
| `control` | **7** [1,1–4,4 %] | 0 [0,0–1,2 %] | 7 | 283 | 299 |
| `producto` | **0** [0,0–1,2 %] | 26 [5,6–11,6 %] | 0 | 291 | 299 |

- **El criterio no se cumple, por una sesión.** El límite inferior de `control` es el 1,06 % y el superior de
  `producto`, el 1,19 %: los intervalos se tocan. Con `producto` en 0 hacían falta 8 caídas en `control`.
- **Las 7 caídas son del tipo de las del job**: las sesiones 70, 198, 215, 225, 233, 252 y 270 de `control` piden
  `boe_articulo` con `ToolSearch`, lo llaman y responden sin ningún `tool_use` de `Skill`.
- **Las 26 de `producto` que activan otra skill** activan `consulta-juridica` y después `normativa`, las dos antes de la
  primera herramienta de kitlegal, abren `con_herramientas.md` y `normas.md` y responden. Ninguna esconde un fallo que
  la tabla no cuente. Es lo que el ADR 0055 dejó fuera del criterio.
- **Otras lecturas de los mismos datos dan la diferencia por real.** La prueba exacta de Fisher unilateral da
  p ≈ 0,0076, y la diferencia, 2,2 puntos con un intervalo de Newcombe al 95 % de [0,55–4,45], que no toca el cero. No
  deciden: el criterio fijado era el de los intervalos, y aceptar ahora con otra prueba es la opción 1 que rechazó el
  0055.

### El control derivó de un día a otro

En H41.3, el `control` se quedó sin activar `normativa` en 14 de 160 sesiones (8,75 %); hoy, en 7 de 320 (2,2 %). La
diferencia no es azar: p ≈ 0,0016. Si la tasa siguiera en el 8,75 %, ver 7 o menos de 320 tendría una probabilidad del
orden de una en un millón.

Lo que cambió es cuántas sesiones empiezan por `ToolSearch`: 71 de 160 (44 %) en H41.3, frente a 53 de 320 (17 %) hoy.
Dentro de ellas, la caída es parecida: 14 de 71 frente a 7 de 53 (p ≈ 0,47). En `producto` no empieza ninguna por
`ToolSearch`. La configuración de las sesiones es la misma los dos días (versión, modelo, herramientas y skills del
`init`), y entre `e12146f1` y `ab951c22` no cambia nada de lo que instala el control. La causa no está identificada; lo
más probable es una deriva del lado del modelo o del servicio.

### Por qué el criterio no podía decidir con este control

El ADR 0055 calculó la potencia con el control entre el 5,3 % y el 8,75 %, el intervalo de un solo día. Con la tasa de
hoy, el diseño de 320 + 320 aprobaba un remedio perfecto (`producto` en 0 %) 4 de cada 10 veces, y uno como `sola`
(0,6 %), 1 de cada 10. «Intervalos que no se tocan» equivale además a exigir una p en torno a 0,003, unas diez
veces más estricto que el 95 % que anuncia.

| Tasa real del control | `producto` en 0 % | `producto` en 0,6 % |
|---|---|---|
| 8,75 % (H41.3) | 1,00 | 1,00 |
| 5,3 % | 0,995 | 0,83 |
| 2,2 % (hoy) | **0,41** | 0,10 |

## Opciones consideradas

1. **Fusionar ya, leyendo los datos con Fisher o con la diferencia de proporciones.** Rechazada. Es cambiar la regla
   después de medir, lo que el ADR 0055 rechazó en su opción 1 y el 0047 en su punto 4, aunque esta vez la nueva
   lectura favorezca al remedio.
2. **Repetir el mismo sondeo.** Rechazada: es repetir para ver si sale en verde (ADR 0047, punto 4). Con el control de
   hoy aprobaría 4 de cada 10 veces aunque el remedio fuera perfecto.
3. **Abandonar H41.4.** Rechazada. Las descripciones cortas llevan 480 sesiones de la eval 04 sin ninguna caída (160 en
   H41.3 y 320 hoy), las evals de las cuatro skills salen en verde en el job, y el sondeo no ha encontrado nada contra
   ellas: no ha podido decidir.
4. **Fisher unilateral con p < 0,025, un tope en `producto` y, si el control cae poco, decidir solo con el tope.** Lo
   propuso la segunda lectura. Rechazada porque no es monótona: con `producto` en 1, un control de 6 caídas suspende
   (p ≈ 0,06) y uno de 5 aprueba, por pasar a decidir solo el tope. Un control que cae más no puede hacer más difícil
   aprobar.
5. **Un sondeo nuevo con otro diseño y otro criterio, fijados aquí antes de medir: filas intercaladas, un tope en
   `producto` y que `producto` no quede por encima de `control`.** Elegida. No es repetir el sondeo para ver si sale en
   verde, porque cambian el diseño y el criterio por un defecto que este sondeo ha demostrado: un control que deriva y
   una potencia de 0,41 con su tasa.

## Decisión

1. **El sondeo del 2026-11-18 no cumple el criterio del ADR 0055, punto 4, y H41.4 no se fusiona con él.** Se lee
   como está; ninguna otra prueba lo convierte en un aprobado.
2. **H41.4 se acepta con un sondeo nuevo**, en las condiciones del anterior (`normativa-producto`, eval 04, modo
   herramienta, `claude-sonnet-5-5`, Claude Code 2.1.326, `INSTALACION=producto`, concurrencia 5), con estos cambios:
   - **Filas intercaladas.** Cada fila son 320 sesiones en 8 tandas de 40, lanzadas el mismo día alternando `control`
     y `producto`, empezando por `control`, para que la deriva caiga por igual en las dos.
   - **Una carpeta por tanda**: `../kitlegal-sondeos/h41-4b/<fila>/<tanda>`, con `<tanda>` de 1 a 8. `make
     laboratorio-tabla` sobre `../kitlegal-sondeos/h41-4b/<fila>` junta las 8 tandas de la fila en una sola de 320.
     La tabla se saca por fila, no sobre `h41-4b`: las dos filas tienen la misma configuración, y la tabla de la
     carpeta común las suma en una sola de 640.
   - `control` se lanza desde un árbol fuera del repositorio en la base de la rama (`ab951c22`), y `producto`, desde
     la cabeza de la rama. Lo lanza la persona, o una sesión interactiva a petición suya.
3. **El criterio, fijado aquí antes de medir.** Con «No activan `normativa` antes de la primera herramienta» como el
   ADR 0055 lo define (*n* menos «Activan antes»), H41.4 pasa si se cumplen las dos condiciones:
   - **un tope en `producto`: como mucho 2 de 320.** Es lo que protege a quien usa el producto, y no depende del
     control;
   - **`producto` no queda por encima de `control`** en ese recuento.

   La tabla da además, por fila, el intervalo de Wilson, y para las dos, la p de Fisher unilateral y la diferencia con
   su intervalo de Newcombe. Esas cifras no deciden.
4. **Lo que el sondeo puede decir depende del control, y el ADR que lo lea lo dice.** Con 6 caídas o más en
   `control`, Fisher puede dar p < 0,025 y el sondeo puede mostrar que el remedio quita la caída. Con menos de 6, ni
   con `producto` en 0 se llega ahí: un aprobado solo dice que las descripciones no dañan y que la caída queda por
   debajo del tope. El ADR que lea la tabla lo escribe así, sin presentarlo como prueba del efecto.
5. **Si no pasa, H41.4 no se fusiona**, el sondeo no se repite para ver si sale en verde, y el ADR que lo lea dice qué
   se mide después.
6. **Para cualquier sondeo futuro de dos filas**, el criterio se fija sobre la diferencia o con un tope absoluto, no
   con intervalos que no se tocan; la potencia se calcula con tasas de control por debajo de la medida, porque un solo
   día no acota la deriva; las filas se intercalan, y la tabla se saca por fila.

La probabilidad de aprobar con este criterio:

| Tasa real del control | `producto` en 0 % | `producto` en 0,3 % | `producto` en 0,6 % | Remedio sin efecto (`producto` = control) |
|---|---|---|---|---|
| 8,75 % | 1,00 | 0,93 | 0,70 | 0,00 |
| 5,3 % | 1,00 | 0,93 | 0,70 | 0,00 |
| 2,2 % | 1,00 | 0,93 | 0,70 | 0,03 |
| 1 % | 1,00 | 0,88 | 0,64 | 0,34 |
| 0,5 % | 1,00 | 0,76 | 0,50 | 0,58 |

Un remedio perfecto aprueba siempre, y uno sin efecto, como mucho 3 de cada 100 veces mientras el control caiga como
hoy o más. Si la caída casi desaparece por sí sola, un remedio sin efecto aprueba más a menudo, y eso se acepta: ya no
hay caída que arreglar, y el tope garantiza que el producto no cae más que en 2 de 320.

Las órdenes, desde `~/code/kitlegal` en la cabeza de la rama de la #274, con `claude` 2.1.326 en el `PATH` y la
credencial en el entorno solo de cada orden:

```sh
git -C ~/code/kitlegal worktree add --detach ~/code/kitlegal-h41-4-control ab951c22
for t in 1 2 3 4 5 6 7 8; do
  for fila in control producto; do
    if [ "$fila" = control ]; then cd ~/code/kitlegal-h41-4-control; else cd ~/code/kitlegal; fi
    CLAUDE_CODE_OAUTH_TOKEN=$T make evals-laboratorio VARIANTES=normativa-producto MODOS=herramienta SESIONES=40 \
      MODELO=claude-sonnet-5-5 CLAUDE_CODE=2.1.326 CONCURRENCIA=5 EVAL=04 INSTALACION=producto \
      DIRECTORIO=../kitlegal-sondeos/h41-4b/$fila/$t
  done
done
cd ~/code/kitlegal
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h41-4b/control
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h41-4b/producto
git -C ~/code/kitlegal worktree remove ~/code/kitlegal-h41-4-control
```

## Consecuencias

- **H41.4 sigue abierta, en la #274, hasta el sondeo nuevo.** La rama no cambia: su spec cita el criterio del ADR 0055,
  y este ADR lo sustituye para la aceptación.
- **Hasta que entre H41.4, el cierre de un hito puede salir en rojo por esta caída.** Con la tasa de hoy, el job (3
  sesiones de la eval 04 en modo herramienta por medición) sale en rojo en torno al 6 % de las mediciones, no 1 de cada
  4 como con la de H41.3. Se acepta hasta el arreglo, como en los ADR 0054 y 0055.
- **El sondeo nuevo cuesta lo mismo que el anterior**, 640 sesiones, unos 25 minutos a concurrencia 5.
- **El arnés no cambia.** Intercalar se hace con 16 lanzamientos de 40 sesiones; la tabla ya junta las tandas de una
  carpeta.
- **La deriva del control queda sin explicar.** Si el siguiente sondeo vuelve a dar un control muy por debajo del de
  H41.3, la caída pierde peso como problema del producto, y eso lo dice el ADR que lo lea.
- **`evidencias/sondeos-20261118/`** queda como la medida de este sondeo.

## Validación

- **Las cifras** salen de `evidencias/sondeos-20261118/tabla-control.md` y `tabla-producto.md`, que imprime `make
  laboratorio-tabla` sobre cada fila.
- **El tipo de las 7 caídas y el recuento de las sesiones que empiezan por `ToolSearch`** salen de los `tool_use` de
  cada `sesion.jsonl`, aquí y en H41.3. Las sesiones se conservan fuera del repositorio, en
  `~/code/kitlegal-sondeos/h41-4/` y `~/code/kitlegal-sondeos/h41-3/` de quien lanzó los sondeos.
- **La estadística.** Los intervalos son de Wilson al 95 %; las p, de la prueba exacta de Fisher (unilateral para las
  dos filas de un sondeo, bilateral entre los controles de dos días); la diferencia, con el intervalo de Newcombe. Las
  probabilidades de aprobar son binomiales, con las dos filas medidas. Que la tabla junta las tandas de una carpeta se
  comprobó con una copia de las dos filas de este sondeo como dos tandas de una misma carpeta: dio una fila de 640.
