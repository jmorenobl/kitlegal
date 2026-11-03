# 0050 · Lo que una skill exige va en su cuerpo o en una referencia que manda leer, nunca solo en la descripción de un parámetro

- **Estado**: propuesta (2026-11-03); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-11-03
- **Hito**: H39, escrito tras el run `50b213d0`, fuera de él, en una sesión interactiva, con las cifras del sondeo de la
  persona (`evidencias/sondeos-20261103/`), como pide la entrada del hito («Aceptación»). Precisa el punto 1 del ADR
  0046, el reparto entre el cuerpo y las referencias, para los dos modos. No cambia ningún umbral, límite ni tope,
  tampoco los 11 000 caracteres del cuerpo (ADR 0046, punto 2), ni reorganiza las skills del producto, que es un hito
  aparte.

## Contexto y problema

Las skills se usan en dos modos: en modo orden el agente llama a `kitlegal` desde la terminal; en modo herramienta, a
las herramientas del servidor MCP local. En los sondeos de la #214 (440 sesiones; `specs/030-h32-arquitectura-de-skills/research.md`
D26 a D30), en modo herramienta `normativa` cargaba las herramientas con `ToolSearch` en el mismo turno en que se
activaba (240 de 240) y casi no abría sus referencias: `normas.md` en 8 a 11 de 80, frente a 77 a 79 en modo orden, y
`con_herramientas.md` en 2 a 4 de 80. La suite completa de la release v0.9.0 salía en rojo solo en ese modo. Cada
defecto se había arreglado con una frase más en el cuerpo, y el cuerpo de `normativa` está a 3 caracteres del tope.

Antes de reorganizar las skills, H39 midió dónde hay que poner lo que una skill exige para que el agente la active
antes de la primera herramienta y lo cumpla en los dos modos. Lo midió con una skill de laboratorio (`laboratorio/`,
fuera del producto): una regla testigo de forma que el modelo no puede saber de memoria (una última línea exacta) y
ocho variantes que difieren de la de control en una sola cosa, con las mismas cuatro evals.

## Lo medido

`claude-sonnet-5-5`, Claude Code 2.1.317, 40 sesiones por variante y modo, 560 en total, todas terminadas. Intervalos
de Wilson al 95 %; el de 40 de 40 es [91,2–100 %].

| Variante: lo que cambia respecto de la de control | Orden: cumplen | Herramienta: cumplen |
|---|---|---|
| `control`: la regla en el cuerpo; la descripción nombra las herramientas | 40 | 40 |
| `instrucciones`: además, las instrucciones del servidor piden activar la skill | — | 40 |
| `entrada`: además, una skill de entrada que remite a la de laboratorio | 40 | 40 |
| `referencia-secundaria`: la regla, en una referencia que el cuerpo nombra de pasada | 40 | 37 [80–97 %] |
| `referencia-paso`: la regla, en una referencia que se lee como paso propio antes de la primera llamada | 39 | 40 |
| `referencia-en-el-punto`: la regla, en una referencia que se manda leer donde hace falta | 40 | 40 |
| `parametro`: la regla, en la descripción del parámetro de la herramienta | — | **25 [47–76 %]** |
| `descripcion-sin-herramientas`: la descripción no nombra herramientas | 40 | 40 |

- **Todas las variantes activan la skill antes de la primera herramienta**, 40 de 40 en cada fila.
- **Las referencias se abren en los dos modos**: 39 o 40 de 40 en modo herramienta, siempre en el turno 2 o 3.
- **La skill de entrada se salta** en 13 de 40 sesiones en modo orden y en 23 de 40 en modo herramienta, y en la eval 01
  en las 20. El agente activa directamente la skill que resuelve.
- **La descripción de un parámetro no funciona como instrucción.** Las 40 sesiones de `parametro` cargan y llaman a la
  herramienta cuyo parámetro lleva la regla, y 15 terminan sin la línea.
- **`ToolSearch` en el mismo turno que `Skill` no impide leer.** Va así en 24 a 40 de 40 sesiones según la variante, y
  en ellas la referencia se abre igual.

**El laboratorio no reproduce el defecto de `normativa`.** Una skill de 1776 caracteres, con la descripción que nombra
las herramientas como la de `normativa`, se activa antes y se lee entera en los dos modos. La causa del defecto no es
dónde está el texto ni el orden de `ToolSearch`, sino algo que `normativa` tiene y el laboratorio no. Hay tres
diferencias sin medir:

- el tamaño: 12 214 caracteres y ocho referencias, frente a 1776 y una;
- remisiones condicionales («si no está», «con ella») frente a una orden incondicional;
- referencias que el agente cree poder suplir de memoria, como una tabla de identificadores del BOE que conoce, frente
  a una línea que solo puede sacar de la referencia. Es lo que ya apuntaba el sondeo de la eval 15
  (`evidencias/sondeos-20261025/`; `docs/SKILLS.md` §1, «Lo que se lee y lo que no»): ninguna de las 40 sesiones de
  `brazo-D` abrió la referencia de guía que el cuerpo mandaba leer antes de responder.

## Opciones consideradas

1. **Esperar a saber la causa del defecto de `normativa` para precisar el 0046.** Rechazada: lo medido ya decide dónde
   no va una regla y que una referencia mandada leer se lee en los dos modos. Esperar deja al próximo arreglo sin sitio.
2. **Llevar las reglas a los parámetros de las herramientas para el modo herramienta.** Rechazada: 25 de 40.
3. **Una skill de entrada o instrucciones del servidor para asegurar la activación.** Rechazada como remedio: no añaden
   nada a una skill que ya se activa sola, y la de entrada se salta.
4. **Precisar el 0046 con lo medido y dejar la causa del defecto de `normativa` a un hito propio.** Elegida.

## Decisión

1. **Lo que una skill exige a la respuesta va en su cuerpo, o en una referencia que el cuerpo manda leer con una orden:
   como paso propio antes de la primera llamada, o en el punto en que hace falta.** Las dos formas se cumplen en 40 de
   40 sesiones en modo herramienta y en 39 o 40 en modo orden. Una referencia nombrada de pasada, en una frase
   secundaria, queda por debajo (37 de 40) y no se usa para una regla. Lo medido es una referencia que trae lo que el
   agente no puede sacar de otra parte. Una referencia que solo trae guía no se abre aunque el cuerpo lo mande
   (`brazo-D`, 0 de 40), y lo que toda respuesta necesita sigue en el cuerpo (ADR 0046, punto 1).
2. **La descripción de un parámetro dice cómo se rellena ese parámetro y nada más.** Ninguna regla sobre la respuesta
   va solo ahí. Que la ayuda de cada argumento llegue a `--describe`, como entrega H39, sirve para rellenar bien el
   parámetro, no como sitio de una regla.
3. **La activación no se asegura con texto fuera de la skill.** Ni las instrucciones del servidor ni una skill de
   entrada mejoran la activación de una skill que se activa sola. La skill de entrada de H32 (ADR 0048) sigue siendo
   la que decide qué capacidad resuelve una pregunta. No es un paso que el agente siempre dé, y ninguna regla que deba
   cumplirse va solo en ella.
4. **Nombrar las herramientas en la descripción ni ayuda ni estorba.** Con ellas o sin ellas, la activación y el
   cumplimiento son los mismos. La descripción sigue el 0046 y nombra lo que ayude a elegir la skill.
5. **El defecto de `normativa` en modo herramienta queda abierto.** El hito que reorganice las skills del producto
   empieza por medir la causa: quita de `normativa` las tres diferencias del apartado anterior, una a una y con el arnés
   de H39, hasta que la skill se lea como la de laboratorio. Hasta entonces, ninguna reorganización da el defecto por
   arreglado sin un sondeo que lo mida en modo herramienta.

## Consecuencias

- `docs/SKILLS.md` §1 recoge los puntos 1 a 4 con sus cifras.
- Mover una regla del cuerpo a una referencia es una forma medida de hacerle sitio en el cuerpo, si la referencia se
  manda leer con una orden. En una skill del tamaño de `normativa` no está medido: el punto 5 lo mide antes.
- La entrada del hito que reorganice las skills se escribe con este ADR y con el arnés de H39 (`make evals-laboratorio`).
  Las variantes nuevas son de `normativa` reducida, no del laboratorio ampliado.
- El sondeo usó Claude Code 2.1.317 y no la 2.1.289 de los sondeos de partida. La comparación con la #214 es
  cualitativa; la que decide, la de las variantes con su control, va dentro del mismo sondeo.

## Validación

Las cifras salen de `evidencias/sondeos-20261103/tabla.md`, que imprime `make laboratorio-tabla` sobre las sesiones, y
de `evidencias/sondeos-20261103/turnos.md`, de su guion `turnos.py`. Lo que se dice de las sesiones marcadas, de su
`sesion.jsonl`, que se conserva fuera del repositorio (`evidencias/sondeos-20261103/LEEME.md`).
