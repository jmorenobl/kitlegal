---
id: H40
fase: "1"
depende_de: [H39]
---

# H40 · Por qué `normativa` no lee sus referencias en modo herramienta

- **Objetivo**: saber, medido, cuál de las diferencias entre `normativa` y la skill de laboratorio de H39 hace que en
  modo herramienta no abra sus referencias, antes de reorganizar las skills del producto (ADR 0050, punto 5). Sin la
  causa, la reorganización movería texto sin saber si se va a leer, y el cuerpo de `normativa` no tiene sitio para otro
  arreglo (10 927 caracteres, con el tope real en 10 930).
- **Evidencia de partida**: en los sondeos de la #214 (evals 05 y 15 de `normativa`;
  `specs/030-h32-arquitectura-de-skills/research.md` D26 a D30), en modo herramienta `normas.md` se abre en 8 a 11 de 80
  sesiones, frente a 77 a 79 en modo orden, y `con_herramientas.md` en 2 a 4 de 80. En el sondeo de H39
  (`evidencias/sondeos-20261103/`, ADR 0050), la skill de laboratorio abre su referencia en 39 o 40 de 40 en modo
  herramienta con cualquiera de las tres formas de remitir, también cuando `ToolSearch` va en el mismo turno que `Skill`.
  El ADR 0050 deja tres diferencias sin medir:
  - el tamaño: 12 214 caracteres y ocho referencias, frente a 1776 y una;
  - las remisiones condicionales («si no está», «con ella») frente a una orden;
  - referencias con lo que el agente cree saber, como la tabla de identificadores del BOE, frente a una línea que solo
    está en la referencia.

  El sondeo de la eval 15 ya apuntaba a la tercera (`evidencias/sondeos-20261025/`, `brazo-D`): 0 de 40 sesiones
  abrieron una referencia de guía que el cuerpo mandaba leer.
- **Entrega**:
  - Variantes de `normativa` fuera de `skills/`, como el laboratorio: no se empotran, no se instalan y no entran en el
    job de evals. Cada una difiere de la de control en una sola de las tres diferencias:
    - `control`: `normativa` tal como está en `skills/`;
    - `remisiones-en-orden`: cada remisión condicional del cuerpo pasa a ser una orden, «lee» y el enlace, en el punto
      en que hace falta. Donde la condición la pone el caso (una fecha, varios bloques, una orden que falla, una
      redacción modificada), la condición se queda y lo que cambia es la orden; donde no la pone, se quita. Cambian
      solo el párrafo o la viñeta de cada remisión, con este texto; lo demás, byte a byte como en `control`:
      - el del «Protocolo» empieza por `Antes de la primera orden o llamada, lee
        [con_herramientas](references/con_herramientas.md).` y pierde «, tras leer [con_herramientas](…)»: el resto,
        «Sigue los pasos en orden, … o, si la tienes, con su herramienta.», queda igual;
      - el del paso 1 pasa a `Lee [normas](references/normas.md) y toma el BOE-A-… de su tabla; si no está, lee
        [buscar_la_norma](references/buscar_la_norma.md) y búscala como dice.`, con `BOE-A-…` entre comillas
        invertidas como en `control`;
      - en el paso 2, «Sin fecha —con ella, [redaccion_en_una_fecha](…)—, lee cada bloque…» pasa a `Con una fecha en
        la pregunta, lee [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md) antes de leer el bloque. Sin
        fecha, lee cada bloque…`, con el resto de la viñeta igual;
      - en el paso 2, la viñeta de varios bloques pasa a `Para leer varios bloques a la vez, lee antes
        [varios_bloques](references/varios_bloques.md). Si una orden no termina con 0, lee
        [lecturas_que_fallan](references/lecturas_que_fallan.md) antes de seguir.`, con `0` entre comillas
        invertidas como en `control`;
      - en «Cómo se cita», «repásalas, y las dudas, en [como_se_cita](…)» pasa a `. Antes de escribir la respuesta,
        lee [como_se_cita](references/como_se_cita.md) y repasa con ella las citas.`, tras «también del remitido»;
      - la de «Redacción modificada» ya es una orden en su punto («lee [redaccion_modificada](…)») y no cambia.
    - `cuerpo-corto`: el cuerpo, con las mismas palabras y las mismas remisiones, sin lo que no hace falta para leer
      y citar las cuatro evals del sondeo. Se queda, en el orden de `control`: el párrafo del «Protocolo»; los pasos
      1 y 2 enteros; del paso 4, el encabezado y la viñeta «Cada afirmación sobre el contenido de una norma lleva su
      cita…»; y las secciones «Cómo se cita», «Redacción en una fecha» y «Redacción modificada». Se quitan los pasos
      3, 5 y 6, las demás viñetas del paso 4, la tabla de comandos con su sección y las «Reglas». Las líneas del
      cuerpo son las de `control`, en su orden, sin añadir ni reescribir ninguna, salvo el número del encabezado del
      paso 4, que pasa a 3. Quedan las ocho remisiones, el mismo frontmatter y las mismas referencias, byte a byte:
      unos 3300 caracteres de cuerpo, frente a 10 929 en `control` y 1776 en el laboratorio;
    - `testigo-en-referencia`: la regla testigo de H39, la misma frase de su `cierre.md` (última línea
      `Sello de lectura: KL-7Q4`), al final de `con_herramientas.md`, la referencia con las reglas que fallaron en modo
      herramienta, en una sección «Cómo termina la respuesta». La remisión del «Protocolo», en el mismo sitio y con la
      misma condición, dice además lo que la referencia trae y el agente no puede saber: «…, tras leer
      [con_herramientas](references/con_herramientas.md), que dice también cómo termina la respuesta.». Es la
      diferencia que se mide: una referencia de la que el cuerpo solo dice lo que el agente cree saber (cómo se
      llaman las herramientas) frente a una que anuncia lo que solo está en ella, como en el laboratorio. Sin ese
      anuncio la variante no mide nada: el agente decide si abre la referencia antes de ver lo que contiene. Lo demás,
      byte a byte como en `control`.
  - El arnés de H39 (`make evals-laboratorio`, `make laboratorio-tabla`) amplía lo que lanza a variantes de una skill
    del producto con sus evals reales. Las del sondeo son las evals 04, 05, 15 y 22 de `normativa`, que ya tienen sus
    lecturas grabadas: diez sesiones por eval, 40 por variante y modo. La tabla da, por variante y modo:
    - las sesiones que activan la skill antes de la primera herramienta;
    - las que abren cada referencia, y en qué turno;
    - las que cumplen la regla testigo, donde la hay.

    Cada cifra va con su intervalo de Wilson al 95 %. Lo que H39 ya mide no cambia.
  - Las filas del sondeo:
    - `control` en los dos modos, para comprobar que el defecto de partida se reproduce;
    - las otras tres solo en modo herramienta, que es donde falla.

    Son 200 sesiones.
- **Fuera**: arreglar o reorganizar `normativa` u otra skill del producto, que es el hito siguiente y se escribe con el
  resultado; cambiar ningún umbral, límite ni tope, también el de 11 000 caracteres (ADR 0046); ampliar el
  laboratorio de H39; y medir con el juez (`afirma_lo_no_leido` y demás clases): el sondeo cuenta lecturas y forma, sin
  modelo.
- **Controles**:
  - `make ci` comprueba que la variante `control` es idéntica byte a byte a `skills/normativa`, para que no se desvíe
    cuando cambie la skill;
  - que cada variante difiere de la de control solo en lo que declara;
  - que ninguna se empotra ni se instala;
  - la regla testigo con respuestas fijas que la cumplen y que no;
  - la tabla con sesiones grabadas.

  Las evals de las skills del producto salen en verde en la medición del cierre.
- **Aceptación**: el sondeo de la persona, con `claude-sonnet-5-5` y 40 sesiones por variante y modo, deja su tabla en
  `evidencias/`. Un ADR dice, con esas cifras, cuál de las tres diferencias explica que `normativa` no lea sus
  referencias en modo herramienta, y qué debe cumplir la reorganización. Si `control` no reproduce el defecto, o
  ninguna variante lo quita, el ADR lo dice y propone qué medir después. El run no lanza el sondeo ni escribe el ADR:
  los dos se añaden a la propuesta de cambio antes de fusionarla, y la propuesta no se fusiona sin ellos.
- **Dependencias**: H39.
