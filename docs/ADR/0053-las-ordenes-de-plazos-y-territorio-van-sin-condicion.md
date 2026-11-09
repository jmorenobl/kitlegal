# 0053 · Las órdenes de leer de plazos-y-territorio van sin condición: H41.1 aplica plazos-y-territorio-sin-condicion

- **Estado**: propuesta (2026-11-09); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-11-09
- **Hito**: H41, escrito tras el run `c285050e` (#248), fuera de él, en una sesión interactiva, con las cifras del
  sondeo de la persona (`evidencias/sondeos-20261109/`, en la #248). Es el ADR que piden la entrada de H41
  («Aceptación»: la tabla de las variantes de `plazos-y-territorio` «va a un ADR que dice qué forma aplica H41.1») y
  FR-053. No cambia ningún umbral, límite ni tope, tampoco los 11 000 caracteres del cuerpo (ADR 0046, punto 2).

## Contexto y problema

El ADR 0052 no fusionó la #245. Las órdenes de `municipios.md` y `calculo_de_plazos.md` que escribía en
`plazos-y-territorio` llevaban la condición delante del verbo, y con ellas el agente abría `municipios.md` en 1 de 40
sesiones en modo orden, frente a 17 de la lista de control. El ADR pidió medir en el laboratorio tres formas de esas dos
órdenes, cada una frente a `plazos-y-territorio-control` en el mismo sondeo (punto 3):

- **`sin-condicion`**: «Antes de la primera orden o llamada, lee [municipios.md]…» y, al final del paso 7, «Antes del
  8, lee [calculo_de_plazos.md]…»; las dos líneas salen de la lista de «Referencias».
- **`condicion-por-orden`**: «Si falta el municipio o `territorio resolver` no sale con `0`, lee [municipios.md]…» en
  el paso 2, y una orden de leer `calculo_de_plazos.md` por cada uno de sus casos (los municipios que cuentan, el 3, el
  4 o 5, `redaccion_posterior`), cada una detrás de la llamada que trae su condición.
- **`lista-con-verbo`**: la lista de control con «Lee» delante de cada enlace, con las condiciones de control.

El texto exacto de las tres está en `specs/034-h41-las-skills-del/contracts/variantes.md` §2. Las tres difieren de
control solo en esas dos órdenes y conservan «Si no lo dice, **pregúntalo** y espera» en el paso 1 (ADR 0052, punto 4).
H41 las midió sin tocar `skills/plazos-y-territorio`, que sigue siendo la de v0.9.0.

## Lo medido

`claude-sonnet-5-5`, Claude Code 2.1.319, las evals 02, 05, 10 y 11 de `plazos-y-territorio`, diez sesiones por eval y
40 por variante y modo. Todas terminaron, y en cada sesión solo estaba instalada su variante. Los intervalos son de
Wilson al 95 %.

| Variante | Modo | Activan antes | `calculo_de_plazos.md` | `municipios.md` |
|---|---|---|---|---|
| `control` | orden | 40 | 25 [47,0–75,8 %] | 3 [2,6–19,9 %] |
| `producto` (= control) | orden | 39 | 25 [47,0–75,8 %] | 10 [14,2–40,2 %] |
| `sin-condicion` | orden | 39 | 29 [57,2–83,9 %] | **39** [87,1–99,6 %] |
| `condicion-por-orden` | orden | 40 | 30 [59,8–85,8 %] | 8 [10,5–34,8 %] |
| `lista-con-verbo` | orden | 38 | 27 [52,0–79,9 %] | 13 [20,1–48,0 %] |
| `control` | herramienta | 40 | 28 [54,6–81,9 %] | 11 [16,1–42,8 %] |
| `producto` (= control) | herramienta | 40 | 26 [49,5–77,9 %] | 14 [22,1–50,5 %] |
| `sin-condicion` | herramienta | 40 | 30 [59,8–85,8 %] | **40** [91,2–100,0 %] |
| `condicion-por-orden` | herramienta | 40 | 28 [54,6–81,9 %] | 8 [10,5–34,8 %] |
| `lista-con-verbo` | herramienta | 40 | 28 [54,6–81,9 %] | 11 [16,1–42,8 %] |

`administrativo.md`, cuya orden genera `make skills-sync` y ninguna variante toca, sale en 28 a 30 de 40 en todas las
filas: es la que piden las evals 05, 10 y 11.

- **Con la regla del ADR 0052, punto 3, ninguna forma queda por debajo de `control`, y solo `sin-condicion` queda por
  encima.** Ninguna celda de una referencia a cuya condición llegan las evals queda por debajo de la de `control` con
  intervalos que no se tocan. `sin-condicion` abre `municipios.md` en 39 y 40 de 40, frente a 3 y 11, con intervalos que
  no se tocan.
- **La orden sin condición se sigue también aquí.** `sin-condicion` abre `municipios.md` en las diez sesiones de cada
  eval y modo, salvo una sesión de la eval 10 en modo orden que no activa la skill. Abre `calculo_de_plazos.md` en 10 de
  10 en las evals 05, 10 y 11 en modo herramienta, y en 10, 9 y 10 en modo orden. Ya se cumplía en `normativa`, con
  `con_herramientas.md` y `normas.md`, en los sondeos de H40, de la #245 y de este (ADR 0051 y 0052).
- **Las órdenes con una condición del caso se siguen donde llega su caso, sin mejorar a control.** En
  `condicion-por-orden`, «si alguno de los municipios cuenta, lee [calculo_de_plazos.md]» llega a las evals 05, 10 y 11,
  y la abre en 10, 10 y 10 de 10 en modo orden y en 8, 10 y 10 en modo herramienta; `control` lo hace en entre 8 y 10.
  A la condición de su orden de `municipios.md` no llega ninguna eval, porque en las cuatro el municipio se dice y sale
  con 0 (research D15 y V6 de H41). Sus 8 de 40 son lecturas que ninguna orden pide, y el caso de esa condición queda
  sin medir.
- **El verbo delante del enlace, con las condiciones de control, no cambia nada medible.** `lista-con-verbo` sale, celda
  a celda, con intervalos que se tocan con los de `control`.
- **La lista de control no es una base estable entre medidas.** `producto` es byte a byte `control` y, en el mismo
  sondeo, abre `municipios.md` en 10 de 40 frente a 3 en modo orden: el ruido de la comparación. Entre sondeos se mueve
  más. En el de la #245 (Claude Code 2.1.317), `plazos-y-territorio-control` la abría en 17 de 40 [28,5–57,8 %] en modo
  orden, y en este en 3 [2,6–19,9 %]. Los intervalos no se tocan, y en la eval 05 pasa de 9 a 1 de 10. Lo que el agente
  hace con una lista de enlaces depende de algo que la skill no fija. La orden sin condición no depende de ello: no baja
  de 39 de 40 en ningún modo.
- **Lo que cuesta.** Con `sin-condicion`, `municipios.md` (3 944 bytes) se lee en toda pregunta de la skill, también en
  las de plazos en que el municipio sale con 0, y `calculo_de_plazos.md` (2 760 bytes) en toda pregunta que llega al
  paso 8. `jerarquia_normativa.md`, cuya línea no cambia, la abren 10 de 40 sesiones en modo orden, frente a 0 de
  `control`. Ninguna eval llega a su condición, y ninguna regla de la skill pide no leerla.

## Opciones consideradas

1. **Aplicar `sin-condicion`.** Elegida: es la única forma que queda por encima de `control`, y su resultado no depende
   de lo que se mueva el control.
2. **Aplicar `condicion-por-orden`.** Rechazada: no queda por debajo de `control`, pero tampoco por encima, y la
   condición de su orden de `municipios.md` no la mide ninguna eval. Además, su cuerpo (11 245 caracteres) pide tres
   movimientos para caber, uno de ellos (M3) con una regla del contencioso a cuyo caso solo llega la eval 09 del job.
3. **Aplicar `lista-con-verbo`.** Rechazada: es la lista de control con otro verbo, y hereda lo inestable de esa lista.
4. **No aplicar ninguna y dejar `plazos-y-territorio` como en v0.9.0.** Rechazada: con la lista, `municipios.md` se
   abre en 3 a 17 de 40 en modo orden según la medida, y en 11 a 14 en modo herramienta.

## Decisión

1. **H41.1 aplica `plazos-y-territorio-sin-condicion` a `skills/plazos-y-territorio`, con el movimiento M1 y ningún
   otro.** El texto de las dos órdenes es el de `specs/034-h41-las-skills-del/contracts/variantes.md` §2.1. M1 lleva la
   frase del paso 10 «Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: di aparte que no has podido leerlo, sin afirmar qué
   dice.» (105 caracteres) al final de `references/calculo_de_plazos.md` (§3 del mismo contrato), y el cuerpo queda en
   10 882 caracteres, por debajo de los 10 930 sin paquetes. M1 cumple la regla del ADR 0052, punto 4: la eval 11 llega
   a ese caso, y con `sin-condicion` la referencia se manda leer siempre antes del 8 (abierta en 10 de 10 en la eval 11
   en los dos modos). «Si no lo dice, **pregúntalo** y espera» sigue en el paso 1.
2. **H41.1 lo mide con su propio sondeo**, como H41: `plazos-y-territorio-control` frente a
   `plazos-y-territorio-producto` ya con la forma aplicada, en los dos modos y con las cuatro evals de H41. La regla es la de este hito: `producto` no
   queda por debajo de `control` en «Activan antes» ni en ninguna referencia a cuya condición llegan las evals, y en
   `municipios.md` queda por encima en los dos modos. La entrada de H41.1 la escribe una persona.
3. **El punto 2 del ADR 0052 sigue siendo una hipótesis y no entra en `docs/SKILLS.md`.** Este sondeo no lo contradice:
   las condiciones del caso de `condicion-por-orden` se siguen donde llegan. Pero tampoco lo prueba. La forma con
   condiciones del caso no mejora a `control`, la condición de `municipios.md` queda sin medir, y la lista con
   condiciones del agente no queda por debajo de `control` en esta medida, aunque sí en la de la #245. Lo que sí está
   medido en `normativa` y `plazos-y-territorio` es que **lo que la skill necesita en las preguntas a las que llega se
   manda leer sin condición, antes de la primera orden o llamada o del paso que lo usa**. `docs/SKILLS.md` §1 lo recoge
   en H41.1, con la forma aplicada.

## Consecuencias

- La #248 se puede fusionar por lo medido en `normativa` y `jurisprudencia` (`evidencias/sondeos-20261109/LEEME.md`).
  Este ADR no decide esa fusión: lo dice la entrada de H41.
- La entrada de H41.1 lleva, además de la forma y M1, la línea del punto 3 en `docs/SKILLS.md` §1, sobre lo que haya
  dejado la #248.
- El laboratorio conserva las tres variantes de `plazos-y-territorio` hasta que H41.1 mida la forma aplicada.
  `condicion-por-orden` es el sitio donde medir el caso de «no sale con 0», si una eval llega a él.
- Una medida que compara una forma con la lista de control solo vale dentro de un mismo sondeo. Entre sondeos con
  versiones distintas de Claude Code, la lista se mueve más que la diferencia que se busca.
- En `normativa`, modo herramienta, `producto` abrió `con_herramientas.md` y `normas.md` en 34 y 35 de 40, frente a 38 y
  38 en el sondeo de la #245, con intervalos que se tocan. Lo que falta está en las evals 04 y 05. La eval 04 es también
  la de la serie de `sin_activar` en modo herramienta del job, que la #248 lleva a la persona (research D21 de H41).

## Validación

Las cifras salen de `evidencias/sondeos-20261109/tabla-*.md`, que imprime `make laboratorio-tabla` sobre las sesiones,
y de `evidencias/sondeos-20261109/recuento.md`, de su guion `recuento.py`, que da el desglose por eval desde el detalle
de las tablas. Las del sondeo de la #245, de `evidencias/sondeos-20261106/`. Las sesiones se conservan fuera del
repositorio (`evidencias/sondeos-20261109/LEEME.md`).
