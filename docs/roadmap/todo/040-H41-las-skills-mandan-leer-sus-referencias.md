---
id: H41
fase: "1"
depende_de: [H40]
---

# H41 · Las skills del producto mandan leer sus referencias

- **Objetivo**: que cada skill del producto con referencias las mande leer con una orden, como decide el ADR 0051, y
  que en modo herramienta `normativa` abra las que necesita como las abre en modo orden. Es la reorganización que
  dejaron pendiente H39 y H40 (ADR 0050, punto 5; ADR 0051, punto 4): hasta que un sondeo la mida, el defecto de
  `normativa` en modo herramienta sigue abierto.
- **Evidencia de partida**: el sondeo de H40 (`evidencias/sondeos-20261104/`; ADR 0051), con `claude-sonnet-5-5`,
  Claude Code 2.1.317, las evals 04, 05, 15 y 22 de `normativa` y 40 sesiones por variante y modo:
  - con `normativa` tal como está, en modo herramienta `con_herramientas.md` se abre en 3 de 40 [2,6–19,9 %] y
    `normas.md` en 7 [8,7–32,0 %], frente a 21 y 40 de 40 en modo orden;
  - con cada remisión escrita como una orden de leer (`laboratorio/normativa/normativa-remisiones-en-orden`), en 38
    [83,5–98,6 %] y 39 [87,1–99,6 %], y `redaccion_en_una_fecha.md` en 9 de las 10 sesiones de la eval 22, la que trae
    una fecha;
  - la orden de leer `como_se_cita.md` «antes de escribir la respuesta» se sigue en 20 de 40 (ADR 0051, punto 3);
  - esa variante mide 11 084 caracteres de cuerpo, 155 más que `normativa` (10 929) y 84 por encima del tope de
    11 000 (ADR 0046, punto 2).

  Las otras dos skills con referencias tienen remisiones que el ADR 0051 no cuenta como órdenes:
  - `plazos-y-territorio` (10 930 caracteres de cuerpo) nombra en una lista, antes del «Protocolo», tres de sus cuatro
    referencias con su propósito y sin el verbo que manda leer («[jerarquia_normativa.md](…): para el nivel que regula
    algo…», y lo mismo `municipios.md` y `calculo_de_plazos.md`); la cuarta, la del paquete (`administrativo.md`), ya es
    una orden, la genera `make skills-sync` y se abre en 58 de 60 sesiones en modo herramienta (#214);
  - `jurisprudencia` (10 842) manda leer sus dos referencias con una orden, pero la de `consulta_preparada.md` va
    «antes de dar la consulta», al escribir la respuesta, y no antes de la llamada que la usa.

  `consulta-juridica` no tiene referencias.
- **Entrega**:
  - **Las remisiones de `normativa`, `plazos-y-territorio` y `jurisprudencia`**, cada una con el ADR 0051:
    - el verbo «lee», el enlace, cuándo y lo que trae (punto 1);
    - la condición que pone el caso (una fecha en la pregunta, varios bloques, una orden que no termina con `0`, una
      entrada `version-obsoleta`, un municipio que no sale con `0`) se queda; la que pone el agente («si la tienes»)
      se quita, y lo que siempre hace falta se manda sin condición (punto 2);
    - la orden va antes de la primera orden o llamada, o en el paso que usa la referencia, nunca como orden de leer
      antes de escribir la respuesta (punto 3). Lo que toda respuesta necesita sigue en el cuerpo (ADR 0046, punto 1;
      ADR 0050, punto 1).

    En `normativa`, el punto de partida son las cinco remisiones de `normativa-remisiones-en-orden`, salvo la de
    `como_se_cita.md`, que es la que el punto 3 descarta. También se escriben con el punto 1 las remisiones cuya
    condición no llegó en el sondeo de H40 (`buscar_la_norma.md`, `lecturas_que_fallan.md`, `redaccion_modificada.md`
    y `varios_bloques.md`), como pide el ADR 0051, punto 4.
  - **Sitio en el cuerpo de `normativa`**, y en el de cualquier otra skill que lo necesite, sin tocar el tope de 11 000
    caracteres: lo que no necesita toda respuesta pasa a una referencia que el cuerpo manda leer con una orden en el
    paso que lo usa (ADR 0050, consecuencias; ADR 0051, consecuencias). Ninguna regla se quita ni cambia de sentido al
    moverse, y lo que se mueve está entero en su referencia. El porqué de cada movimiento queda en el plan.
  - **El arnés de H39 y H40 mide la reorganización**. Cada familia (`normativa`, `plazos-y-territorio` y
    `jurisprudencia`) tiene dos variantes:
    - `<skill>-control`: la skill tal como está en `main` al lanzar el hito, que es la de v0.9.0, congelada en
      `laboratorio/<skill>/`. La de `normativa` ya existe y deja de compararse con `skills/normativa`: `make ci`
      comprueba que no cambia. Las variantes de H40 siguen comparándose con ella, como hasta ahora;
    - `<skill>-producto`: `skills/<skill>` tal como queda tras el hito. El arnés la instala desde `skills/`, sin
      copia versionada que pueda desviarse.

    En `normativa`, el sondeo usa sus cuatro evals de H40 y mide también `normativa-remisiones-en-orden`, para
    comparar con ella en el mismo sondeo. En las otras dos familias, el sondeo usa cuatro evals de la skill que ya
    tienen sus lecturas grabadas y que entre las cuatro llegan a la condición del mayor número de remisiones; la
    elección y su porqué van en el research del hito. Diez sesiones por eval, 40 por variante y modo.
  - **Las filas del sondeo**: `control` y `producto` de cada familia, en los dos modos, y
    `normativa-remisiones-en-orden` en modo herramienta. Son 520 sesiones.
  - **La regla de lectura de la tabla, ampliada** (`specs/032-h40-por-qu-e/contracts/tabla.md` §2; ADR 0051,
    consecuencias): una orden de `Bash` también abre la referencia `<f>` si hace `cd` a un directorio que termina en
    `<skill>/references` y después, en la misma orden, lee `<f>` por su nombre solo. Era la lectura que no se contaba
    en 9 de las 200 sesiones de H40. Las suites congeladas de H39 y H40 siguen en verde.
  - `CHANGELOG.md` (*Unreleased*), por el cambio de las tres skills, y `docs/SKILLS.md`, si la reorganización da algo
    que la guía no recoja.
- **Fuera**:
  - cambiar ningún umbral, límite ni tope, tampoco los 11 000 caracteres del cuerpo (ADR 0046, punto 2);
  - cambiar la descripción de una skill: la activación ya se mide en 40 de 40 en los dos modos (ADR 0050 y 0051);
  - cambiar el generador de la tabla de comandos o de las referencias generadas desde `data/`, que es común a todas
    las skills;
  - `consulta-juridica`, que no tiene referencias;
  - evals nuevas;
  - medir con el juez en el sondeo, que cuenta lecturas y activación, sin modelo;
  - guardar las sesiones del job de evals como artefacto, que sigue en `backlog/medida-de-las-skills.md`.
- **Controles**:
  - `make ci` comprueba que cada `<skill>-control` no cambia;
  - que cada variante de H40 difiere de `normativa-control` solo en lo que declara;
  - que ninguna variante se empotra ni se instala;
  - el cuerpo de cada skill, como hasta ahora, con 11 000 caracteres como mucho;
  - la regla de lectura ampliada, con sesiones grabadas que leen con `cd … && cat <f>` y que no leen.

  Las evals de las skills del producto salen en verde en la medición del cierre, en los dos modos.
- **Aceptación**: el sondeo de la persona, con `claude-sonnet-5-5` y 40 sesiones por variante y modo, deja su tabla en
  `evidencias/`, y la reorganización vale si cumple todo esto:
  - en modo herramienta, `normativa-producto` abre `con_herramientas.md` y `normas.md` con un intervalo que no se toca
    con el de `normativa-control` y sí con el de `normativa-remisiones-en-orden` (ADR 0051, punto 4);
  - en cada familia y modo, `producto` no queda por debajo de `control` en las sesiones que activan la skill antes de
    la primera herramienta, ni en ninguna referencia a cuya condición llegan las evals. «Por debajo» quiere decir con
    un intervalo que no se toca con el de `control`.

  El run no lanza el sondeo. La tabla y su lectura se añaden a la propuesta de cambio antes de fusionarla, y la
  propuesta no se fusiona sin ellas. Si el sondeo no cumple, la propuesta no se fusiona, y lo que se mida va a un ADR
  que dice qué medir después.
- **Dependencias**: H40.
