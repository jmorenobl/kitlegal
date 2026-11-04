# 0051 · Una remisión a una referencia es una orden de leerla, no una condición ni una mención

- **Estado**: propuesta (2026-11-04); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-11-04
- **Hito**: H40, escrito tras el run `6ccfd6c6`, fuera de él, en una sesión interactiva, con las cifras del sondeo de la
  persona (`evidencias/sondeos-20261104/`), como pide la entrada del hito («Aceptación»). Cierra el punto 5 del ADR
  0050 y precisa su punto 1. No cambia ningún umbral, límite ni tope, tampoco los 11 000 caracteres del cuerpo (ADR
  0046, punto 2), ni reorganiza `normativa` ni otra skill del producto, que es el hito siguiente.

## Contexto y problema

En modo herramienta `normativa` casi no abre sus referencias: `normas.md` en 8 a 11 de 80 sesiones y
`con_herramientas.md` en 2 a 4, frente a 77 a 79 `normas.md` en modo orden (#214;
`specs/030-h32-arquitectura-de-skills/research.md` D26 a D30). La skill de laboratorio de H39 no lo reproduce: abre su referencia en 39 o 40 de 40 en los dos modos (ADR
0050). El ADR 0050 dejó tres diferencias sin medir entre las dos skills:

- el tamaño: un cuerpo de 10 929 caracteres y ocho referencias, frente a 1776 y una;
- las remisiones: condicionales o de pasada en `normativa` («si la tienes, con su herramienta, tras leer
  [con_herramientas]», «toma el `BOE-A-…` de la tabla de [normas]», «con ella, [redaccion_en_una_fecha]»), frente a
  una orden en el laboratorio («lee [references/cierre.md]: dice cómo termina»);
- el contenido: referencias con lo que el agente cree saber, como los identificadores del BOE, frente a una línea que
  solo está en la referencia.

H40 midió cada una con una variante de `normativa` que difiere de la de control solo en ella (`laboratorio/normativa/`;
la definición de cada variante, en la entrada del hito).

## Lo medido

`claude-sonnet-5-5`, Claude Code 2.1.317, las evals 04, 05, 15 y 22 de `normativa`, diez sesiones por eval y 40 por
variante y modo, 200 en total, todas terminadas. En cada sesión solo está instalada la variante. Intervalos de Wilson al
95 %. Las cifras son las de la tabla del arnés; entre corchetes, el recuento ampliado de `recuento.md`, que suma las
sesiones que leen la referencia con `cd …/references && cat <fichero>`, una lectura que el contrato de la tabla no
cuenta (9 de 200 sesiones; ninguna conclusión cambia).

| Variante: lo que cambia respecto de la de control | Modo | `con_herramientas.md` | `normas.md` | Cumplen el testigo |
|---|---|---|---|---|
| `control`: `skills/normativa` byte a byte | orden | 21 [37–67 %] | 40 [91–100 %] | — |
| `control` | herramienta | **3** [3–20 %] (4) | **7** [9–32 %] (8) | — |
| `remisiones-en-orden`: cada remisión, una orden «lee [ref]» donde hace falta | herramienta | **38** [84–99 %] (39) | **39** [87–100 %] (40) | — |
| `cuerpo-corto`: el cuerpo, de 10 929 a 3298 caracteres, con las mismas remisiones y palabras | herramienta | 10 [14–40 %] (11) | 13 [20–48 %] (14) | — |
| `testigo-en-referencia`: el sello de H39 en `con_herramientas.md`, anunciado en su remisión | herramienta | 10 [14–40 %] (11) | 12 [18–45 %] (13) | 11 [16–43 %] |

- **El defecto se reproduce con la skill sola.** `control` abre `normas.md` en 40 de 40 en modo orden y en 7 de 40 en
  modo herramienta, como en la #214, sin las otras skills del producto instaladas: la causa está en `normativa`, no en
  lo que tenga al lado.
- **Las remisiones lo explican.** Con cada remisión convertida en una orden de leer, `con_herramientas.md` y
  `normas.md` se abren en 38 y 39 de 40 en modo herramienta, siempre en el turno 2, como el laboratorio; los intervalos
  no se tocan con los de `control`. La remisión a `normas.md` no era condicional, sino una mención dentro de otra
  orden («toma el `BOE-A-…` de la tabla de [normas]»): lo que falta en las dos es el verbo que manda leer.
- **El tamaño no lo explica.** Con un tercio del cuerpo, `con_herramientas.md` y `normas.md` se abren en 10 y 13 de 40.
  Lo que sube viene casi entero de la eval 22, la que trae una fecha (8 y 8 de 10, frente a 1 y 3 en `control`); en
  las evals 04, 05 y 15 se abren en 2 y 5 de 30, como en `control` (2 y 4 de 30).
- **El contenido no lo explica.** Anunciar en la remisión que la referencia dice cómo termina la respuesta, algo que
  el agente no puede saber, la abre en 10 de 40 (11 con el recuento ampliado), y la regla de la referencia se cumple
  exactamente en las sesiones que la leen: 11 de 40. Las 29 que no la abren terminan sin la línea, sin aviso. En el
  laboratorio, una mención de pasada con el mismo anuncio se abrió en 39 de 40 (ADR 0050): la diferencia es la frase
  que remite, no lo que hay detrás.
- **Una orden de leer al final se sigue menos.** En `remisiones-en-orden`, «Antes de escribir la respuesta, lee
  [como_se_cita]» se sigue en 20 de 40 (21), de 1 de 10 en la eval 04 a 9 de 10 en la 22; las órdenes del comienzo,
  en 38 a 40. `como_se_cita.md` repite y matiza la forma de cita que el cuerpo ya da, y es guía, como la de `brazo-D`
  (`evidencias/sondeos-20261025/`): lo que toda respuesta necesita sigue en el cuerpo (ADR 0046, punto 1; ADR 0050,
  punto 1).
- **`ToolSearch` en el turno de `Skill`** va en 31 a 39 de 40 sesiones según la variante, también en las que leen: no
  impide leer, como midió el ADR 0050.
- `buscar_la_norma.md`, `lecturas_que_fallan.md`, `redaccion_modificada.md` y `varios_bloques.md` no se abren en
  ninguna sesión: ninguna de las cuatro evals llega a su condición (la norma está en la tabla, ninguna lectura falla,
  ninguna trae `version-obsoleta` y se lee un bloque). Sus remisiones no quedan medidas.

## Opciones consideradas

1. **Recortar el cuerpo de `normativa` para que se lea.** Rechazada: con un tercio del cuerpo, las evals sin fecha
   siguen sin abrir las referencias.
2. **Mover a las referencias reglas que el agente no pueda suplir.** Rechazada como remedio: la regla se cumple solo
   donde se lee la referencia, y con una remisión que no es una orden se lee en 11 de 40.
3. **Escribir cada remisión que la respuesta necesita como una orden de leer.** Elegida.

## Decisión

1. **Una remisión a una referencia es una orden de leerla: el verbo «lee», el enlace y cuándo, y lo que trae.** Así se
   abren en 38 a 40 de 40 sesiones en modo herramienta. No es una remisión, a efectos de que se lea, la referencia
   nombrada dentro de otra orden («toma … de la tabla de [normas]»), tras una condición sobre el agente («si la tienes,
   … tras leer [con_herramientas]»), entre rayas («con ella, [redaccion_en_una_fecha]») ni como sitio de dudas («las
   dudas, en [como_se_cita]»): se abren en 3 a 13 de 40, también con un cuerpo de un tercio o con un anuncio de lo que
   traen.
2. **La condición que pone el caso se queda, y la que pone el agente se quita.** «Con una fecha en la pregunta, lee
   [redaccion_en_una_fecha] antes de leer el bloque» conserva su condición, que depende de la pregunta. «Si la tienes»
   depende de lo que el agente cree necesitar, y deja la lectura a su juicio: lo que siempre hace falta se manda sin
   condición, antes de la primera orden o llamada.
3. **La orden va antes de la primera llamada o en el paso que la usa, no al final.** Una orden de leer antes de
   escribir la respuesta, para una referencia de guía, se sigue en la mitad de las sesiones; lo que toda respuesta
   necesita va en el cuerpo (ADR 0046, punto 1).
4. **La reorganización de `normativa` y de las demás skills cumple el punto 1 en cada remisión y se mide en modo
   herramienta con el arnés de H39 y H40 antes de darse por buena.** Su variante se compara con `normativa-control` en
   el mismo sondeo y abre `con_herramientas.md` y `normas.md` como `remisiones-en-orden`, con un intervalo que no se
   toque con el de `control`. Lo que el sondeo no midió (las cuatro remisiones cuya condición no llegó) se escribe con
   el punto 1 y se mide con evals que lleguen a ella cuando las haya.

## Consecuencias

- `docs/SKILLS.md` §1 recoge los puntos 1 a 3 con sus cifras.
- `remisiones-en-orden` mide 11 084 caracteres de cuerpo, 84 por encima del tope (ADR 0046, punto 2): la reorganización
  no puede copiarla tal cual y tiene que hacer sitio en el cuerpo de `normativa`. Mover a una referencia mandada leer lo
  que no necesita toda respuesta es una forma medida de hacerlo (ADR 0050, consecuencias).
- La entrada del hito que reorganice las skills se escribe con este ADR. El arnés de H40 ya mide variantes de una skill
  del producto con sus evals reales; si la reorganización de otra skill lo necesita, se amplía a su familia.
- El contrato de la tabla (`specs/032-h40-por-qu-e/contracts/tabla.md` §2) no cuenta una lectura con `cd` al directorio
  de las referencias y el nombre del fichero solo. Afectó a 9 de 200 sesiones y a una por celda como mucho. El hito
  que vuelva a usar el arnés amplía esa regla antes de su sondeo.
- El sondeo usó Claude Code 2.1.317, como el de H39, y no la 2.1.289 de la #214. La comparación con la #214 es
  cualitativa; la que decide, la de las variantes con su control, va dentro del mismo sondeo.

## Validación

Las cifras salen de `evidencias/sondeos-20261104/tabla.md`, que imprime `make laboratorio-tabla` sobre las sesiones, y
de `evidencias/sondeos-20261104/recuento.md`, de su guion `recuento.py`, que da el desglose por eval y el recuento
ampliado. Lo que se dice de sesiones concretas, de su `sesion.jsonl`, que se conserva fuera del repositorio
(`evidencias/sondeos-20261104/LEEME.md`).
