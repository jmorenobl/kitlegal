# 0014 · El grafo en tres piezas: memoria pronto, instrumento tarde, y las operaciones viajan en el `Resultado`

- **Estado**: aceptada
- **Fecha**: 2026-09-13
- **Hito**: H7 (pieza G0), H10 (pieza G1); amplía `schema.Resultado` sin tocar el contrato del ADR 0005

## Contexto y problema

El roadmap traía el grafo como un bloque (H17–H18, fase 3) que llegaba «cuando las skills del municipio
ya existen», y anotaba que los applets anteriores incorporarían `Emit` entonces. Al reordenar por tiempo
hasta el uso (ADR 0013) la pregunta era si ese bloque va antes o después. Ninguna de las dos respuestas
es buena, porque el grafo no es una cosa:

- **Memoria y validación.** Nodos con id natural, aristas con `source`, y las reglas de `graph check`
  que solo necesitan una fuente: `version-obsoleta`, `fuente-caducada`. Es lo que hace determinista la
  capa de consulta: el binario recuerda qué vio, cuándo y de dónde, y puede decir si ha cambiado.
- **Asunto.** `.kitlegal/case.db`, `Consulta`, `Plazo`, `Escrito`, `expediente`, y las reglas
  `plazo-vencido` y `plazo-sin-base`. Es la agenda de quien usa el kit.
- **Instrumento.** Anomalías, cruces entre fuentes, scoring, `vigilar run`, exportaciones. Necesita
  PLACSP, BDNS, BORME y el padrón del INE; sin ellos no aporta nada.

Y hay dos razones de fondo para que las dos primeras piezas lleguen pronto que el roadmap no recogía:

1. **El grafo es una serie temporal.** `version-obsoleta`, `norma-derogada` o el histórico de
   `BloqueVersion` solo dicen algo si el grafo llevaba grabando. Meterlo en el hito 18 no retrasa una
   funcionalidad: retrasa el reloj. Cada consulta hecha antes de que exista es historia perdida, y no hay
   forma de recuperarla.
2. **Un applet diseñado sin el grafo es un applet peor.** El contrato (ADR 0005) no conoce el grafo, así
   que cada applet escrito antes se diseña sin preguntarse qué identidades observa (DIR3, NIF, ELI, hash
   del bloque) ni si las conserva en sus tipos. Incorporarlas después no es añadir un método: es reabrir
   el parser y el modelo de datos de un applet ya cerrado y revisado, con el riesgo de que las identidades
   que no capturó se reconstruyan a medias. Los applets que nacen con el grafo se diseñan enteros.

Frente a eso, el riesgo de un grafo temprano es doble: que el modelo se diseñe antes de que existan las
fuentes que lo tensan, y que el grafo, con el texto de los bloques dentro, acabe sustituyendo a la fuente
como origen de lo que la skill cita, en contra del principio II de la constitución.

Hay además una decisión de diseño que el roadmap daba por hecha sin examinarla: *cómo* emite un applet.
`refs/kitlegal-grafo.md` §4 dice `Emit(ctx) []GraphOp`, un método aparte. Es una decisión de contrato y
hay que tomarla por méritos.

## Opciones consideradas

1. **Grafo entero en la fase 3**, como estaba. Pierde la historia de las fases 1 y 2 y obliga al retrofit
   de siete applets.
2. **Grafo entero pronto**, después de `territorio`. Trae anomalías, cruces y vigilancia a una fase en la
   que no hay fuentes que las alimenten, contra el principio V (YAGNI) y el VIII (sin skill que lo use, no
   se construye).
3. **Tres piezas**: memoria y validación (G0) en H7, asunto (G1) en H10, instrumento (G2) en el backlog.

Y, para el mecanismo de emisión, tres formas de que un applet entregue lo que observa:

- **a. Un método aparte, `Emit(ctx) []GraphOp`** (la forma de los `refs/`). `Ejecutar` observa y guarda;
  el kernel llama después a `Emit`, que lee lo guardado. Es un protocolo en dos fases con estado mutable
  entre llamadas y acoplamiento temporal oculto: `Emit` antes de `Ejecutar` devuelve vacío sin error, y
  un `Argumentos` que se reutilizara emitiría lo de la invocación anterior. El `ctx` en la firma invita
  además a hacer entrada/salida durante la emisión, que es justo lo que el grafo no debe hacer. Y el
  `source` de cada operación lo tendría que escribir el applet o reconstruirlo el kernel casando dos
  valores devueltos por separado. Rechazada por diseño.
- **b. Ampliar la firma de `Ejecutar`** a `(Resultado, []Op, error)`. Una sola fase, sin estado, pero las
  operaciones viajan separadas de la `Procedencia` que las hace citables: el kernel vuelve a tener que
  casar dos valores, y el contrato del ADR 0005 se sustituye. Funcionalmente equivalente a la siguiente y
  estrictamente peor en cohesión.
- **c. Un campo más en `schema.Resultado`.** El applet ya devuelve un valor de dominio con `Procedencia` y
  `Datos`; lo que observó del mundo es parte de ese mismo resultado y viaja con él, atómico, con la misma
  procedencia. Valor cero válido. Es el principio del ADR 0005 —el applet declara, el kernel decide—
  aplicado al grafo.

Si la opción b hubiera sido la mejor, este ADR sustituiría al 0005 sin más. No lo hace porque c es mejor
diseño, no porque sustituirlo cueste.

## Decisión

Se adopta la **opción 3**.

### Las piezas y cuándo llegan

| Pieza | Contenido | Hito |
|---|---|---|
| **G0 · Memoria y validación** | `world.db` con `nodes/edges/texts`; `Apply` transaccional e idempotente; puerto `GraphStore` y tipos de operación en `core`; `--no-graph` como Null Object; `graph show`, `graph stats`; `graph check` con `version-obsoleta` y `fuente-caducada`; rechazo por regex de NIF/DNI en `Persona` dentro de `Apply` (constitución VII). `boe` y `territorio` incorporan la emisión. | H7 |
| **G1 · Asunto** | `.kitlegal/case.db` y `config.yaml`, `Consulta`, `lb:en_asunto`, `--asunto`, `ATTACH` de `world.db`; `expediente`; `check` gana `plazo-vencido` y `plazo-sin-base`. `escrito` (H11) emite `Escrito` al asunto. | H10 |
| **G2 · Instrumento** | `graph neighbors\|path`, `data/anomalias/`, `graph anomalies`, cruces, `anomalies mark`, `vigilar run`, `graph export`. `boe analisis` → aristas ELI, `eli:cites`, `graph history`. `Afirmacion`, `lb:fundamenta` y `check` bloqueante. | backlog |

G0 va **después de `territorio` (H6) y antes de `cita` (H8)**. Con solo `boe` emitiendo, el grafo sería
un índice de una familia de ids; con `boe` y `territorio` hay dos (ELI y DIR3/INE) y una arista real
(`lb:pertenece_a`). Y `cita` es la pieza más parecida a un grafo de todo el producto —resuelve una
referencia en lenguaje natural a un id de nodo—: si el grafo llega antes, `cita` nace emitiendo y
`cita-verificada` puede apoyarse en `check` en lugar de reimplementar la validación y tirarla después. El
retrofit queda en dos applets jóvenes con fixtures ya grabados. Todo applet desde H8 nace emitiendo.

### Cómo emite un applet: las operaciones viajan en el `Resultado` (opción c)

El applet devuelve un `schema.Resultado` con `Procedencia` y `Datos`; **las operaciones de grafo son un
tercer campo de ese mismo valor**. El kernel las entrega al `GraphStore` después de presentar la salida,
con la `Procedencia` del propio `Resultado` como `source` de cada nodo y arista. Lo que esto garantiza, y
por qué es la opción correcta y no solo la cómoda:

- **Un dato sin fuente no puede entrar, por construcción y no por disciplina.** El `source` no lo escribe
  el applet: lo pone el kernel a partir de la procedencia con la que el applet ya tenía que responder.
  Un applet que no declara procedencia no tiene sobre (ADR 0006) y, por lo mismo, no tiene grafo. Con
  las opciones a y b esto sería una regla que vigilar; con c es imposible violarla.
- **Una sola fase, sin estado entre llamadas.** Lo observado y lo respondido son el mismo valor, producido
  en el mismo instante por la misma ejecución. No hay orden de llamadas que respetar ni instancia que
  pueda emitir lo de otra invocación.
- **La emisión no puede hacer entrada/salida.** No hay contexto ni cliente a mano al construir el campo:
  es dominio puro, como el resto de `Resultado`.
- **`--no-graph` es un `GraphStore` nulo**, que descarta; el applet no sabe si existe el grafo ni lo
  consulta, y su salida no cambia ni un byte por la existencia del grafo.
- **El dominio define los tipos de operación** (puerto en `internal/core`, como el roadmap §2 ya preveía
  para `GraphStore`); `internal/graph` los aplica sobre SQLite. `core` sigue sin importar adaptadores.
- **El contrato `Applet` del ADR 0005 queda intacto** como consecuencia, no como objetivo: `Resultado`
  crece con un campo cuyo valor cero es válido y ningún applet de ejemplo tiene que implementar nada en
  vacío.

Los nombres concretos del campo y de los tipos los fija el plan de H7. Lo que este ADR fija es la
dirección: emisión como parte del valor devuelto, con la procedencia, aplicada por el kernel, nunca escrita
por el applet ni producida en una llamada aparte.

### La línea que no se cruza: el grafo es índice y validador, nunca fuente de contenido citado

`refs/kitlegal-grafo.md` §5.2 sugiere que el agente reutilice el texto guardado en `texts` si la versión
sigue vigente. Eso, mal entendido, se come el principio II: una cita necesita `fuente`, `url`,
`fecha_consulta` y `hash` de *esta* consulta. La regla, desde G0:

> El grafo dice **qué hay que volver a comprobar**, nunca **qué dice el artículo**. Ninguna skill cita
> texto leído del grafo; el texto se cita del sobre de una consulta al applet de la fuente, que la caché
> (H3) sirve sin red cuando está vigente.

La tabla `texts` existe para `graph history` y para `check` (comparar versiones por hash), no para
responder. Un test de H7 comprueba que ningún verbo de `graph` devuelve el cuerpo de un bloque, y las
evals de `boe-legislacion` desde H7 comprueban que una consulta repetida vuelve a pasar por `boe articulo`
(con caché) y no por `graph show`.

### Lo que G0 deja fuera a propósito

`eli:cites` por regex sobre el texto de los bloques (cobertura infinita, precisión traicionera: entra con
`boe analisis` cuando haya un motivo), FTS5, `graph query` con SQL libre, `neighbors`, `path`,
exportaciones, promoción de props a columnas generadas. Nada de eso tiene consumidor en la fase 1 y el
principio V lo prohíbe.

## Consecuencias

**A favor**

- La historia empieza en el hito 7 y no en el 18: cuando las reglas de `check` empiecen a importar, el
  grafo llevará meses grabando.
- Ningún applet desde H8 se diseña sin preguntarse qué identidades observa; el retrofit se limita a `boe`
  y `territorio`.
- El `source` de cada nodo y arista queda garantizado por el mismo mecanismo que garantiza el sobre.
- El contrato de applet del ADR 0005 sigue intacto; `Resultado` crece con un campo cuyo valor cero es
  válido.
- La garantía de cita no depende de que nadie recuerde no leer del grafo: hay un test y una eval.
- `internal/graph` se construye con el andamiaje SQLite que `internal/cache` (H3) acaba de fijar
  —migraciones embebidas, WAL, espera ante bloqueo que respeta el contexto— y con ese conocimiento
  reciente: las dos capas de persistencia quedan coherentes entre sí. Es una consecuencia favorable del
  orden, no su motivo.

**En contra, y asumido**

- El esquema de nodos y aristas se fija con dos fuentes en vez de con siete. El riesgo es bajo porque es
  un property graph genérico (`nodes`/`edges` con `props` JSON, `refs/kitlegal-grafo.md` §3), no una
  ontología tipada: lo que PLACSP o BORME necesiten es un tipo de nodo y una arista más, no un cambio de
  esquema.
- H7 añade un hito a la fase 1 antes de `cita`. Es el precio de que `cita` nazca emitiendo.
- Las reglas de `check` de G0 son dos y dirán poco al principio; su valor crece con el uso, y eso es
  exactamente el argumento.
- `refs/kitlegal-grafo.md` §4 y §10 siguen diciendo `Emit(ctx) []GraphOp` y «el grafo llega cuando las
  skills del municipio ya existen». Es un documento semilla y no se reescribe; `refs/00-README.md` avisa
  de la lectura correcta y, ante conflicto, prevalecen la constitución, `CLAUDE.md` y el roadmap.
