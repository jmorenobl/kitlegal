# 0057 · Las descripciones cortas quitan la caída de la eval 04, y H41.4 se fusiona

- **Estado**: propuesta (2026-11-18); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-11-18
- **Hito**: H41.4. Se escribe después del run `20123bb8` (#274), en la propia rama y fuera del run, en una sesión
  interactiva. Es el ADR que, según el punto 5 del ADR 0056, lee el sondeo que ese ADR fijó. Usa sus cifras
  (`evidencias/sondeos-20261118-intercalado/`). No cambia ningún umbral, límite ni tope, ni el job de evals, ni
  ninguna skill.

## Contexto y problema

El ADR 0055 decidió que H41.4 pusiera en `skills/` las cuatro descripciones cortas, con un sondeo de aceptación. Ese
sondeo (`evidencias/sondeos-20261118/`) no cumplió su criterio: el control cayó en 7 de 320 sesiones, y con esa tasa
el diseño no podía decidir. El ADR 0056 fijó, antes de medir, un sondeo nuevo y su criterio:
- en las condiciones del anterior, con 8 tandas de 40 sesiones por fila, alternando `control` y `producto`;
- H41.4 pasa si, en «No activan `normativa` antes de la primera herramienta», `producto` tiene como mucho 2 de 320 y
  no queda por encima de `control`;
- con 6 caídas o más en `control`, el sondeo puede mostrar que el remedio quita la caída; con menos, un aprobado solo
  dice que las descripciones no dañan.

## Lo medido

Las condiciones del sondeo:
- `claude-sonnet-5-5` y Claude Code 2.1.326, la versión del job;
- `normativa-producto`, la eval 04 en modo herramienta, `INSTALACION=producto` y concurrencia 5;
- `control` desde un árbol en `ab951c22`, la base de la rama, y `producto` desde su cabeza (`1ec12710`), con las mismas
  skills que `77d6ed7c`;
- 16 tandas de 40, de 09:28 a 09:53 UTC del 2026-11-18, alternas y empezando por `control`;
- 640 sesiones, todas terminadas y sin ningún `sesion.err` con contenido.

| Fila | No activan `normativa` antes | Activan otra skill | No activan ninguna |
|---|---|---|---|
| `control` | **33** [7,4–14,1 %] | 0 [0,0–1,2 %] | 33 |
| `producto` | **0** [0,0–1,2 %] | 0 [0,0–1,2 %] | 0 |

| Tanda | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|
| Caídas en `control` (de 40) | 4 | 3 | 1 | 4 | 2 | 5 | 8 | 6 |
| Caídas en `producto` (de 40) | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |

- **El criterio se cumple**: `producto` tiene 0, por debajo del tope de 2, y no queda por encima de `control`, que
  tiene 33.
- **El sondeo muestra el efecto, no solo que no dañan**: `control` tiene más de 6 caídas. Las cifras que el ADR 0056
  publica sin decidir lo confirman: Fisher unilateral, p ≈ 5·10⁻¹¹, y la diferencia, 10,3 puntos con un intervalo de
  Newcombe de [7,2–14,1].
- **Las 33 caídas son del tipo de las del job**: no tienen ningún `tool_use` de `Skill`. En `control` empiezan por
  `ToolSearch` 76 sesiones y en `producto`, 12; en `producto`, todas activan `normativa` antes de la primera
  herramienta de kitlegal.
- **El control deriva en horas, no en días.** Dos horas antes, el sondeo del ADR 0056 dio 7 de 320 (2,2 %); este, 33
  de 320 (10,3 %), en la línea de los 14 de 160 de H41.3. Es la deriva que justificó intercalar las filas, y aquí las
  dos la vivieron a la vez.
- **Ninguna sesión de `producto` activa `consulta-juridica` esta vez**, frente a 26 de 320 dos horas antes y 18 de 160
  en H41.3. No decide (ADR 0055), pero también varía de una hora a otra.

## Decisión

1. **Las descripciones cortas quitan la caída de la eval 04** con las cuatro skills instaladas y Claude Code 2.1.326,
   en las 800 sesiones de la eval 04 medidas con ellas (160 en H41.3, con la variante de `laboratorio/`, y 640 en los
   dos sondeos de H41.4), ninguna con la caída.
2. **H41.4 se fusiona.** Cumple el criterio del ADR 0056, y su CI, `codecov/patch` y las evals de las cuatro skills
   están en verde.
3. **La caída deja de ser un riesgo aceptado del cierre**: desde la fusión, el job mide las skills con las
   descripciones cortas, y ningún cierre debería salir en rojo por ella. Si vuelve a aparecer en el job, es un defecto
   de la skill, no esta caída conocida.

## Consecuencias

- **Lo que el ADR 0055 dejó como riesgo sigue en pie**: el sondeo solo mide la activación de `normativa`, y las
  descripciones de `jurisprudencia` y `plazos-y-territorio` también se acortan. Su activación la protege el job.
- **La deriva del control sigue sin explicar.** Varía entre el 2 % y el 10 % en dos horas, con la misma
  configuración. Un sondeo de dos filas que no las intercale no es fiable con esta caída (ADR 0056, punto 6).
- **`evidencias/sondeos-20261118-intercalado/`** queda como la medida de aceptación de H41.4.

## Validación

- **Las cifras** salen de `evidencias/sondeos-20261118-intercalado/tabla-control.md` y `tabla-producto.md`, que
  imprime `make laboratorio-tabla` sobre cada fila. Regeneradas desde las sesiones, coinciden byte a byte con las del
  lanzamiento.
- **Las caídas por tanda y las sesiones que empiezan por `ToolSearch`** salen de los `tool_use` de cada `sesion.jsonl`.
  Las sesiones se conservan fuera del repositorio, en `~/code/kitlegal-sondeos/h41-4b/` de quien lanzó el sondeo.
- **La estadística**, como en el ADR 0056: intervalos de Wilson al 95 %, Fisher exacta unilateral y el intervalo de
  Newcombe para la diferencia.
