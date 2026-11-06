---
id: H41
fase: "1"
depende_de: [H40]
---

# H41 · Las skills del producto mandan leer sus referencias

- **Objetivo**: que `normativa` y `jurisprudencia` manden leer sus referencias con una orden, como deciden los ADR
  0051 y 0052, y que en modo herramienta `normativa` abra las que necesita como las abre en modo orden; y medir, en el
  laboratorio y sin tocar la skill del producto, qué forma de las órdenes de `plazos-y-territorio` se sigue al menos
  como la lista de hoy (ADR 0052, punto 3). Es la reorganización que dejaron pendiente H39 y H40 (ADR 0050, punto 5;
  ADR 0051, punto 4), corregida con lo que midió el primer run de H41.
- **Evidencia de partida**:
  - el sondeo de H40 (`evidencias/sondeos-20261104/`; ADR 0051), con `claude-sonnet-5-5`, Claude Code 2.1.317, las
    evals 04, 05, 15 y 22 de `normativa` y 40 sesiones por variante y modo: con `normativa` tal como está, en modo
    herramienta `con_herramientas.md` se abre en 3 de 40 y `normas.md` en 7; con cada remisión escrita como una orden
    de leer (`normativa-remisiones-en-orden`), en 38 y 39; la orden de leer `como_se_cita.md` «antes de escribir la
    respuesta», en 20 de 40;
  - el primer run de H41 (`0f548c7d`, #245, cerrada sin fusionar) y el sondeo de su aceptación
    (`evidencias/sondeos-20261106/`; ADR 0052), con el mismo modelo, la misma versión de Claude Code y 520 sesiones:
    - en `normativa`, las remisiones de la #245 abren `con_herramientas.md` y `normas.md` en 38 de 40 en modo
      herramienta, frente a 0 y 9 de `normativa-control`: cumple;
    - en `jurisprudencia`, las de la #245 se siguen como en `jurisprudencia-control` en los dos modos;
    - en `plazos-y-territorio`, la lista de «Referencias» reescrita con la condición delante del verbo («Al presentar
      un territorio, si falta el municipio o no sale con 0, y para explicar su cobertura, lee [municipios.md]…») abre
      `municipios.md` en 1 de 40 en modo orden, frente a 17 de `plazos-y-territorio-control`, con intervalos que no
      se tocan: no cumple. `calculo_de_plazos.md` cae de 10 a 3 de 10 en la eval 05 en modo orden. La #245 había
      sacado además del paso 1 «Si no lo dice, **pregúntalo** y espera», que ninguna eval mide.

  `skills/plazos-y-territorio` es la de v0.9.0 y su cuerpo mide 10 930 caracteres: el tope sin paquetes, con la línea
  de 70 de la vertical de ensayo de H32 (`TestVerticalDeEnsayo`), sin margen. `consulta-juridica` no tiene
  referencias.
- **Entrega**:
  - **Las remisiones de `normativa` y `jurisprudencia`, como en la #245** (rama `033-h41-las-skills-del`, que se
    conserva; `specs/033-h41-las-skills-del/contracts/remisiones.md` §2, §3 y §5): cada una con el verbo «lee», el
    enlace, cuándo y lo que trae (ADR 0051, punto 1); la condición que trae el caso se queda y la que pone el agente
    se quita (ADR 0051, punto 2, precisado por el ADR 0052, punto 2); la orden va antes de la primera orden o
    llamada, o en el paso que usa la referencia, detrás de la llamada si la condición la trae lo que devuelve, y nunca
    antes de escribir o dar la respuesta (ADR 0051, punto 3). En `normativa`, el sitio en el cuerpo se hace como en la
    #245: «Nada de otra conversación» pasa a `redaccion_modificada.md` y «Si una lectura falla, di qué bloques faltan
    y por qué» a `lecturas_que_fallan.md`, enteros y palabra por palabra. Lo medido en la #245 puede volver tal cual;
    lo que cambie respecto a ella, el plan lo dice con su porqué.
  - **`skills/plazos-y-territorio` no cambia**, ni su `SKILL.md` ni sus referencias: sigue siendo la de v0.9.0, y
    «Si no lo dice, **pregúntalo** y espera» sigue en el paso 1 (ADR 0052, punto 4). La forma que gane en el
    laboratorio la aplica H41.1.
  - **Tres variantes de `plazos-y-territorio` en el laboratorio** (ADR 0052, punto 3), en
    `laboratorio/plazos-y-territorio/` junto a `plazos-y-territorio-control`. Cada una difiere de control solo en la
    forma de las órdenes de `municipios.md` y `calculo_de_plazos.md`; la línea de `jerarquia_normativa.md`, la del
    paquete y el resto del cuerpo y de las referencias quedan como en control, también «Si no lo dice,
    **pregúntalo** y espera» en el paso 1:
    - **sin condición**: «lee [municipios.md]…» antes de la primera orden o llamada, como `con_herramientas.md` en
      `normativa`, y «lee [calculo_de_plazos.md]…» antes del cálculo;
    - **una condición del caso por orden**: «si falta el municipio o `territorio resolver` no sale con `0`, lee
      [municipios.md]…», «si alguno de los municipios cuenta, lee [calculo_de_plazos.md]…», y una orden aparte para
      cada uno de los otros casos de `calculo_de_plazos.md` (un plazo sin calcular; una lectura que sale con 4 o 5; una
      lectura con `redaccion_posterior`), en el paso cuya orden o llamada trae su condición. Las condiciones que pone
      el agente («al presentar un territorio», «para explicar su cobertura») no se escriben;
    - **la lista de control con el verbo delante del enlace**: «- Lee [municipios.md](references/municipios.md): al
      presentar un territorio…», y lo mismo `calculo_de_plazos.md`, con las condiciones de control.

    El texto exacto de cada variante va en el plan. Las variantes son del laboratorio y, como
    `normativa-remisiones-en-orden` en H40, no las limita el tope del cuerpo: medirlas con un movimiento de texto
    mezclaría dos cambios. El plan da el cuerpo de cada una y, si pasa de 10 930, lo que movería H41.1 para que quepa,
    con la regla del ADR 0052, punto 4: solo a una referencia a cuyo caso llega alguna eval, como la frase del paso 10
    «Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**…» (105 caracteres) a `calculo_de_plazos.md`, a cuyo caso llega la
    eval 11.
  - **El arnés de las tres familias, como en la #245**: en cada familia, `<skill>-control`, congelada en
    `laboratorio/<skill>/` (la de v0.9.0), y `<skill>-producto`, que el arnés instala desde `skills/<skill>` sin copia
    versionada. En `plazos-y-territorio`, `plazos-y-territorio-producto` es igual que `plazos-y-territorio-control`
    byte a byte, y el sondeo la mide igual, como fila de control contra control: da el ruido de la comparación con
    las mismas evals. El arnés acepta además las tres variantes del punto anterior. Las cuatro evals de cada familia
    son las de la #245 (research D12): 04, 05, 15 y 22 de `normativa`; 02, 05, 10 y 11 de `plazos-y-territorio`; 01,
    03, 04 y 06 de `jurisprudencia`. Diez sesiones por eval, 40 por variante y modo.
  - **La regla de lectura con `cd`, como en la #245** (research D13): una orden de `Bash` también abre la referencia
    `<f>` si hace `cd` a un directorio que termina en `<skill>/references` y después, en la misma orden, lee `<f>` por
    su nombre solo. Las suites congeladas de H39 y H40 siguen en verde.
  - **Las filas del sondeo**: `control` y `producto` de cada familia en los dos modos, las tres variantes de
    `plazos-y-territorio` en los dos modos y `normativa-remisiones-en-orden` en modo herramienta: 19 filas y 760
    sesiones.
  - `CHANGELOG.md` (*Unreleased*), por el cambio de `normativa` y `jurisprudencia`, y `docs/SKILLS.md` §1 («Cómo se
    remite») con lo que da la reorganización de esas dos skills. El punto 2 del ADR 0052 entra en `docs/SKILLS.md`
    cuando lo mida el sondeo de este hito, en el ADR que lo lea, no en el run.
- **Fuera**:
  - cambiar `skills/plazos-y-territorio`, su `SKILL.md` o sus referencias: es H41.1, con la forma que gane;
  - cambiar ningún umbral, límite ni tope, tampoco los 11 000 caracteres del cuerpo (ADR 0046, punto 2);
  - cambiar la descripción de una skill: la activación ya se mide en 40 de 40 en los dos modos (ADR 0050 y 0051);
  - cambiar el generador de la tabla de comandos o de las referencias generadas desde `data/`, que es común a todas
    las skills;
  - `consulta-juridica`, que no tiene referencias;
  - evals nuevas, y cambiar las que ya existen: «pregúntalo y espera» no sale del cuerpo, y no hace falta la eval que
    deja el municipio sin decir (ADR 0052, punto 4);
  - medir con el juez en el sondeo, que cuenta lecturas y activación, sin modelo;
  - guardar las sesiones del job de evals como artefacto, que sigue en `backlog/medida-de-las-skills.md`.
- **Controles**:
  - `make ci` comprueba que cada `<skill>-control` no cambia y que `skills/plazos-y-territorio` es igual que
    `plazos-y-territorio-control`;
  - que cada variante de H40 difiere de `normativa-control`, y cada variante de `plazos-y-territorio` de
    `plazos-y-territorio-control`, solo en lo que declara;
  - que ninguna variante se empotra ni se instala;
  - la frase de cada remisión de `normativa` y `jurisprudencia` (`TestRemisionesDelRepositorio` de la #245), con
    `plazos-y-territorio` fuera hasta H41.1, declarado en el test;
  - el cuerpo de cada skill del producto, con 11 000 caracteres como mucho;
  - la regla de lectura ampliada, con sesiones grabadas que leen con `cd … && cat <f>` y que no leen.

  Las evals de las skills del producto salen en verde en la medición del cierre, en los dos modos.
- **Aceptación**: el sondeo de la persona, con `claude-sonnet-5-5` y 40 sesiones por variante y modo, deja una tabla
  por familia en `evidencias/`, y el hito vale si cumple todo esto:
  - en modo herramienta, `normativa-producto` abre `con_herramientas.md` y `normas.md` con un intervalo que no se toca
    con el de `normativa-control` y sí con el de `normativa-remisiones-en-orden` (ADR 0051, punto 4);
  - en `normativa` y `jurisprudencia`, en cada modo, `producto` no queda por debajo de `control` en las sesiones que
    activan la skill antes de la primera herramienta, ni en ninguna referencia a cuya condición llegan las evals. «Por
    debajo» quiere decir con un intervalo que no se toca con el de `control`.

  Las variantes de `plazos-y-territorio` no deciden la fusión: su tabla, con la misma regla frente a
  `plazos-y-territorio-control` (ADR 0052, punto 3) y el desglose por eval, va a un ADR que dice qué forma aplica
  H41.1, o, si ninguna cumple, que no aplica ninguna y qué medir después. La fila de `plazos-y-territorio-producto`
  dice cuánto se mueve el control de una medida a otra.

  El run no lanza el sondeo. Las tablas y su lectura se añaden a la propuesta de cambio antes de fusionarla, y la
  propuesta no se fusiona sin ellas. Si `normativa` o `jurisprudencia` no cumplen, la propuesta no se fusiona, y lo que
  se mida va a un ADR que dice qué medir después.
- **Después**: H41.1 aplica a `skills/plazos-y-territorio` la forma que gane, con el movimiento que diga el plan de
  este hito para que quepa, y lo mide con su sondeo.
- **Dependencias**: H40.
