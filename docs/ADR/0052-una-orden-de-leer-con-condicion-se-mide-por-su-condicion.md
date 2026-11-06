# 0052 · Una orden de leer con una condición se mide por su condición: la reorganización de H41 no se fusiona

- **Estado**: propuesta (2026-11-06); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-11-06
- **Hito**: H41, escrito tras el run `0f548c7d`, fuera de él, en una sesión interactiva, con las cifras del sondeo de la
  persona (`evidencias/sondeos-20261106/`), como pide la entrada del hito («Aceptación»: «si el sondeo no cumple, la
  propuesta no se fusiona, y lo que se mida va a un ADR que dice qué medir después»). Precisa el punto 2 del ADR 0051.
  No cambia ningún umbral, límite ni tope, tampoco los 11 000 caracteres del cuerpo (ADR 0046, punto 2).

## Contexto y problema

El ADR 0051 decidió que una remisión a una referencia es una orden de leerla («lee», el enlace, cuándo y lo que trae),
que la condición que pone el caso se queda y la que pone el agente se quita, y que la orden va antes de la primera
llamada o en el paso que la usa. Lo midió en `normativa`, en modo herramienta, con remisiones sin condición o con una
condición que trae la pregunta («con una fecha en la pregunta»). H41 (#245) reescribió así las remisiones de
`normativa`, `plazos-y-territorio` y `jurisprudencia`, y su entrada pidió medirlo con el arnés de H39 y H40 antes de
fusionar: en cada familia, `<skill>-control` (la de v0.9.0) frente a `<skill>-producto` (la de la #245).

En `plazos-y-territorio`, el control nombraba sus tres referencias escritas a mano en una lista, antes del
«Protocolo», con el enlace delante y su propósito detrás:

```text
- [municipios.md](references/municipios.md): al presentar un territorio, si falta el municipio o no sale con 0, y
  para explicar su cobertura.
- [calculo_de_plazos.md](references/calculo_de_plazos.md): en los municipios que cuentan, un plazo sin calcular y una
  lectura fallida o con `redaccion_posterior`.
```

La #245 las dejó en el mismo sitio, con las mismas condiciones, delante del verbo:

```text
- Al presentar un territorio, si falta el municipio o no sale con 0, y para explicar su cobertura, lee
  [municipios.md](references/municipios.md): qué hacer.
- En los municipios que cuentan, un plazo sin calcular y una lectura fallida o con `redaccion_posterior`, lee
  [calculo_de_plazos.md](references/calculo_de_plazos.md): su detalle.
```

Además, para que el cuerpo cupiera con la línea de 70 caracteres de la vertical de ensayo de H32, sacó del paso 1 «Si
no lo dice, **pregúntalo** y espera: no lo supongas, tampoco el de quien usa el kit.» y la llevó al final de
`municipios.md`, que ya decía lo mismo (T009; `specs/033-h41-las-skills-del/gates/supuestos.md`).

## Lo medido

`claude-sonnet-5-5`, Claude Code 2.1.317, cuatro evals por familia (research D12 de H41), diez sesiones por eval y 40
por variante y modo, 520 en total, todas terminadas. En cada sesión solo está instalada la variante. Intervalos de
Wilson al 95 %. Las cifras son las de las tablas del arnés, que ya cuentan la lectura con `cd …/references && cat <f>`
(H41, FR-041); el desglose por eval, de `recuento.md`.

| Familia | Modo | Referencia | `control` | `producto` |
|---|---|---|---|---|
| `normativa` | herramienta | `con_herramientas.md` | 0 [0,0–8,8 %] | **38** [83,5–98,6 %] |
| `normativa` | herramienta | `normas.md` | 9 [12,3–37,5 %] | **38** [83,5–98,6 %] |
| `normativa` | orden | `con_herramientas.md` | 14 [22,1–50,5 %] | 40 [91,2–100,0 %] |
| `plazos-y-territorio` | orden | `municipios.md` | 17 [28,5–57,8 %] | **1** [0,4–12,9 %] |
| `plazos-y-territorio` | herramienta | `municipios.md` | 12 [18,1–45,4 %] | 9 [12,3–37,5 %] |
| `plazos-y-territorio` | orden | `calculo_de_plazos.md` | 27 [52,0–79,9 %] | 22 [39,8–69,3 %] |
| `plazos-y-territorio` | herramienta | `calculo_de_plazos.md` | 26 [49,5–77,9 %] | 15 [24,2–53,0 %] |
| `jurisprudencia` | orden | `consulta_preparada.md` | 30 | 30 |
| `jurisprudencia` | herramienta | `consulta_preparada.md` | 27 | 26 |

`normativa-remisiones-en-orden`, medida en el mismo sondeo, abre `con_herramientas.md` y `normas.md` en 34 y 35 de 40
en modo herramienta [70,9–92,9 % y 73,9–94,5 %]. `administrativo.md`, cuya orden genera `make skills-sync` y la #245 no
toca, se abre en 29 o 30 de 40 en las dos variantes y modos. Todas las sesiones activan la skill antes de la primera
herramienta, salvo una de `plazos-y-territorio-control` en modo orden.

- **La orden sin condición se cumple, otra vez.** En `normativa`, «Antes de la primera orden o llamada, lee
  [con_herramientas]» y «Lee [normas] y toma el `BOE-A-…` de su tabla» se siguen en 38 a 40 de 40 en los dos modos,
  como en `remisiones-en-orden` y en H40. La orden con la condición que trae la pregunta, «Con una fecha en la
  pregunta, lee [redaccion_en_una_fecha]», en las 10 sesiones de la eval 22 en los dos modos (5 en `control` en modo
  herramienta). Se cumple la primera condición de la aceptación de H41.
- **La orden con una condición que el agente juzga se sigue menos que la lista que nombraba la referencia.** En
  `plazos-y-territorio`, `municipios.md` cae de 17 a 1 de 40 en modo orden, con intervalos que no se tocan: no se
  cumple la segunda condición de la aceptación. Lo que pierde viene de las evals 05 y 11, que presentan un territorio
  sin que falte el municipio (de 9 y 5 de 10 a 0 y 0); en la 02, la del municipio no cubierto, de 3 a 1. En
  `calculo_de_plazos.md`, el total no cae por debajo con intervalos que no se tocan, pero en la eval 05, la de los
  municipios que cuentan sin plazo sin calcular ni lectura fallida, cae de 10 a 3 de 10 en modo orden y de 6 a 0 en
  modo herramienta; en las evals 10 y 11, que traen un plazo sin calcular o una lectura fallida, se mantiene.
- **Con la lista, el agente leía las referencias juntas al empezar.** En modo orden, 14 sesiones de `control` abren
  tres referencias en el turno 2 o 3, y en las sesiones de la eval 05 que se han mirado, seguidas tras activar la
  skill y antes de `kitlegal territorio resolver`; ninguna de `producto` abre más de dos. Con la condición delante del
  verbo, el agente decide si la condición se da antes de leer.
- **Las dos condiciones que caen no son de las que trae la pregunta.** «Al presentar un territorio» y «para explicar
  su cobertura» describen lo que el agente va a hacer, y su respuesta depende de su juicio (la propia research D12 lo
  dice de las evals 05, 10 y 11). «En los municipios que cuentan, un plazo sin calcular y una lectura fallida…» junta
  tres casos con «y», y la eval 05, que solo trae el primero, es la que cae; la revisión del plan ya señaló que esa
  orden «puede leerse como una condición conjunta». Ninguna de las dos lecturas está medida por separado.
- **`pregúntalo y espera` queda sin medir.** Ninguna eval del sondeo, ni del job de evals, deja el municipio sin
  decir. Con `municipios.md` abierta en 1 de 40 en modo orden, la regla que salió del paso 1 llega al agente casi solo
  en el caso que la referencia cubre.
- En `jurisprudencia`, las dos remisiones se siguen como en `control` en los dos modos; `jerarquia_normativa.md` no se
  abre en ninguna variante, porque ninguna eval llega a su condición.

## Opciones consideradas

1. **Fusionar la #245 porque `normativa`, el objeto del hito, cumple.** Rechazada: la entrada exige las dos
   condiciones y dice que, si no se cumplen, no se fusiona.
2. **Fusionar `normativa` y `jurisprudencia` y devolver `plazos-y-territorio` al control.** Rechazada como decisión de
   este ADR: es otra entrada, con otro alcance, y la escribe una persona. Queda como una de las salidas de la medida
   siguiente.
3. **No fusionar la #245 y medir, en `plazos-y-territorio`, la forma de la orden con condición antes de relanzar.**
   Elegida.

## Decisión

1. **La #245 no se fusiona.** Sus cifras quedan en `evidencias/sondeos-20261106/`, y la rama
   `033-h41-las-skills-del` se conserva hasta que la entrada corregida se relance.
2. **El punto 2 del ADR 0051 se precisa: una condición se queda si la trae el caso, y solo si el agente puede saber que
   se da sin juzgar lo que va a hacer.** «Con una fecha en la pregunta» se queda; «al presentar un territorio» o «para
   explicar su cobertura» no son del caso, sino del agente, y una lista de casos unidos con «y» no es una condición,
   sino varias. Esta precisión es la hipótesis que lo medido sostiene; se da por buena cuando la mida el sondeo del
   punto 3.
3. **Antes de relanzar H41, la entrada corregida mide en `plazos-y-territorio`, con el arnés de H41 y las mismas cuatro
   evals, variantes que difieren de `plazos-y-territorio-control` solo en la forma de las órdenes de `municipios.md` y
   `calculo_de_plazos.md`**, al menos:
   - **lo que siempre hace falta, sin condición**: la orden de leer `municipios.md` antes de la primera llamada, como
     `con_herramientas.md` en `normativa`, y la de `calculo_de_plazos.md` antes del cálculo;
   - **una condición del caso por orden**: «si falta el municipio o `territorio resolver` no sale con `0`, lee
     [municipios.md]…», «si alguno de los municipios cuenta, lee [calculo_de_plazos.md]…», y una orden aparte por cada
     uno de los otros casos;
   - **la lista de control con el verbo delante del enlace**: «- Lee [municipios.md](…): al presentar un territorio…».

   Cada variante se compara con `control` en el mismo sondeo, con la misma regla de la aceptación de H41: ninguna
   referencia a cuya condición llegan las evals por debajo de `control` con intervalos que no se tocan. En el
   relanzamiento, `skills/plazos-y-territorio` no cambia, y `plazos-y-territorio-producto` queda igual que
   `plazos-y-territorio-control`: las variantes viven solo en el laboratorio, y la forma que gane la aplica al producto
   un hito aparte (H41.1).
4. **Una regla que el cuerpo da como orden no sale del cuerpo para hacer sitio sin una eval que llegue a su caso.**
   «Si no lo dice, **pregúntalo** y espera» vuelve al paso 1, o la entrada corregida trae la eval que deja el
   municipio sin decir. Lo que no necesita toda respuesta pasa a una referencia (ADR 0050 y 0051, consecuencias) solo
   si alguna eval llega al caso que la manda leer.

## Consecuencias

- La entrada de H41 (`docs/roadmap/in-progress/H41-las-skills-mandan-leer-sus-referencias.md`) se corrige con este ADR
  y se relanza; mientras tanto sigue en `in-progress/` en `main`. Lo que la #245 hizo en `normativa` y
  `jurisprudencia`, y el arnés con las tres familias, están medidos y pueden volver tal cual en el relanzamiento.
- `docs/SKILLS.md` §1 («Cómo se remite») recoge el punto 2 cuando lo mida el sondeo del punto 3, no antes: hasta
  entonces es una hipótesis.
- El margen del cuerpo de `plazos-y-territorio` es estrecho: 10 930 caracteres como mucho sin paquetes, por la línea de
  70 de la vertical de ensayo de H32. Las variantes del punto 3 que no quepan con la regla del punto 4 lo dicen en la
  entrada corregida, con lo que movería cada una.
- El sondeo usó Claude Code 2.1.317, como los de H39 y H40: las cifras de `normativa-remisiones-en-orden` y de
  `normativa-control` se pueden comparar con las del ADR 0051, y casan (34 y 35 de 40 frente a 38 y 39; 0 y 9 frente
  a 3 y 7).

## Validación

Las cifras salen de `evidencias/sondeos-20261106/tabla-*.md`, que imprime `make laboratorio-tabla` sobre las sesiones,
y de `evidencias/sondeos-20261106/recuento.md`, de su guion `recuento.py`, que da el desglose por eval y el número de
referencias por sesión desde el detalle de las tablas. Lo que se dice de sesiones concretas, de su `sesion.jsonl`, que
se conserva fuera del repositorio (`evidencias/sondeos-20261106/LEEME.md`).
