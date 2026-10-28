# 0047 · El sondeo de la persona decide con 80 respuestas y un límite que no suspende a la skill que cumple

- **Estado**: propuesta (2026-10-28); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-10-28
- **Hito**: H31.1, escrito tras el run `c0ccd6c3`, fuera de él, en una sesión interactiva. Cambia la regla de decisión
  del sondeo que la entrada de H31.1 deja a una persona antes de fusionar. No toca ningún umbral del job de evals
  (`evals/*/juez/clases.yaml`; ADR 0038 y 0045), ni el juez (ADR 0037), ni el modelo que decide (ADR 0031), ni el
  arnés del sondeo (`evidencias/sondeos-20261025/parches/arnes-del-sondeo.diff`).

## Contexto y problema

La entrada de H31.1 pide, antes de fusionar, un sondeo de la eval 15 de 40 respuestas, 20 por modo, en el que
`afirma_lo_no_leido` dé como mucho 2 de 40 y `cuenta_su_proceso` como mucho 4 de 40. Lo que se quería es que la skill
afirme lo no leído en torno a un 5 % de las respuestas, frente al 20 % del control (`evidencias/sondeos-20261025/`,
16 de 80), y que no vuelva a contar su proceso.

**Con 40 respuestas, el límite de 2 suspende a una skill que cumple.** El número de marcadas de una skill con una tasa
*p* sigue una binomial. Con 40 respuestas y un límite de 2, la probabilidad de pasar es:

| Tasa de la skill | 2,5 % | 5 % | 10 % | 12,5 % | 20 % |
|---|---|---|---|---|---|
| Pasa con ≤ 2 de 40 | 92 % | **68 %** | 22 % | 11 % | 1 % |

Una skill justo en el objetivo suspende 1 de cada 3 veces. Es el sorteo que el ADR 0038 quitó del job, con la misma
causa: un límite fijado sin mirar cuántas respuestas lo deciden.

**Lo medido lo enseña.** `brazo-W` dio 1 de 40 (`evidencias/sondeos-20261026/`); la skill de la #204, que conserva sus
pasos 3, 5 y 6 palabra por palabra, dio 4 de 40. La diferencia entre las dos no es significativa (Fisher unilateral,
p = 0,18), y las dos juntas dan 5 de 80, frente a 16 de 80 del control (p = 0,009). Con estos datos, la tasa de W puede
estar entre el 0,1 % y el 13 % (intervalo exacto al 95 %): el 1 de 40 que fijó la referencia tuvo, probablemente,
algo de suerte, y el criterio exige repetirla.

## Opciones consideradas

1. **Dejar el criterio y repetir el sondeo hasta que pase.** Rechazada: convierte la decisión en un sorteo, y repetir
   una medida para ver si sale en verde es lo que la regla de la persona prohíbe.
2. **Bajar la exigencia a la tasa del control.** Rechazada: dejaría pasar una skill que no mejora.
3. **Más respuestas con el mismo límite proporcional, 4 de 80.** Rechazada: una skill en el 5 % pasa el 63 % de las
   veces; el sorteo sigue.
4. **80 respuestas y un límite elegido por lo que distingue.** Elegida.

## Decisión

1. **El sondeo tiene 80 respuestas**: la eval 15, 40 repeticiones por modo, con lo demás como hoy (el arnés,
   `claude-sonnet-5-5`, Claude Code 2.1.289, concurrencia 5). Dura en torno a media hora.
2. **`afirma_lo_no_leido` pasa con como mucho 6 de 80.** Es el límite con el que una skill en el objetivo, el 5 %,
   pasa casi 9 de cada 10 veces, y una en el 12,5 % pasa 1 de cada 9, la misma protección que daba el criterio de 40
   frente a una skill mala. Con 6 de 80, la mejora sobre el control es significativa (Fisher unilateral, p = 0,019),
   así que el límite ya exige mejorar al control.
3. **`cuenta_su_proceso` pasa con como mucho 8 de 80**, la proporción de antes, 1 de cada 10. Es el mayor límite con
   el que la mejora sobre el control, 17 de 80, sigue siendo significativa (p = 0,04). A la tasa medida en los brazos
   con el paso 6, en torno al 5 %, pasa el 98 % de las veces.
4. **Si no pasa, la propuesta no se fusiona** y la persona decide con las cifras y las respuestas marcadas, como hoy.
   El sondeo no se repite para ver si sale en verde.

| Tasa de la skill | 2,5 % | 5 % | 7,5 % | 10 % | 12,5 % | 20 % |
|---|---|---|---|---|---|---|
| `afirma_lo_no_leido` pasa con ≤ 6 de 80 | 100 % | **89 %** | 61 % | 30 % | 11 % | 0 % |
| `cuenta_su_proceso` pasa con ≤ 8 de 80 | — | 98 % | 86 % | 59 % | 32 % | 1 % |

## Consecuencias

- El límite nominal de `afirma_lo_no_leido` sube del 5 % al 7,5 % de las respuestas, y una skill en el 10 % pasa algo
  más que antes (30 % frente a 22 %). El objetivo sigue en el 5 %: lo que cambia es que la skill que lo cumple deja de
  suspender 1 de cada 3 veces, y la que está en el 12,5 % o peor sigue sin pasar.
- El sondeo cuesta el doble de sesiones. Con el límite de uso compartido, no se lanza mientras corre un job de evals.
- La entrada de H31.1 remite a este ADR en lo que deja a una persona. Un hito posterior que pida un sondeo de la
  persona fija su límite con el mismo cálculo: el objetivo, el número de respuestas y la probabilidad de pasar a la
  tasa objetivo y a una claramente peor.

## Validación

Las probabilidades salen de la binomial y las p de Fisher unilateral, con la misma función que
`evidencias/sondeos-20261026/guiones/analizar.py`; las cifras de los brazos, de `evidencias/sondeos-20261026/LEEME.md`
y del sondeo de la #204 del 2026-10-28.
