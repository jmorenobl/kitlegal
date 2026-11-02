---
id: H39
fase: "1"
depende_de: [H32]
---

# H39 · Cómo se organiza una skill para que se active y se lea en los dos modos

- **Objetivo**: saber, medido, qué organización de una skill y del servidor MCP hace que el agente active la skill antes
  de la primera herramienta y lea y cumpla lo que va fuera del cuerpo, en modo orden y en modo herramienta, antes de
  reorganizar las skills del producto. Hoy los cuerpos están en el techo
  del ADR 0046 (11 000 caracteres, unos 4870 tokens, por debajo de los 5000 de la especificación; `normativa` tiene
  10 927 y el tope real es de 10 930), porque cada defecto medido se ha arreglado con una frase más en el cuerpo: lo
  que va a una referencia, en modo herramienta, no se lee (`docs/USO.md`, 2026-11-02 y 2026-10-25). El próximo arreglo
  de `normativa` no cabe.
- **Evidencia de partida**: las 440 sesiones de los sondeos de la #214 (`~/kitlegal-sondeos/20261030-pr214`,
  `20261101-pr214` y `20261102-pr214`; specs/030-h32-arquitectura-de-skills/research.md D26 a D30). En `normativa`, en
  modo herramienta, `ToolSearch` va en el mismo turno que `Skill` en 240 de 240 sesiones: el agente carga las
  herramientas antes de leer el cuerpo y va a `boe_indice` sin abrir ninguna referencia. `normas.md` se abre en 11, 11
  y 8 de 80 (en modo orden, 79, 77 y 78) y `con_herramientas.md` en 4, 3 y 2 de 80 (39, 26 y 41). `con_herramientas.md`
  lleva las reglas que fallaron en la eval 22 en modo herramienta. En `plazos-y-territorio`, cuya descripción no nombra
  ninguna herramienta, `administrativo.md` se abre en 58 de 60 en modo herramienta. La suite completa de la release
  v0.9.0 (ejecución 41474010633, sobre `cd61442`) sale en rojo por lo mismo, en modo herramienta y solo en él: en la
  eval 04 de `normativa`, una sesión va directa a `boe_articulo` sin activar ninguna skill ni leer el índice
  (`sin_activar` 1 de 60; tercera vez, tras `7f8ef89` y `75bf303`), y en la eval 02 de `consulta-juridica` dos de tres
  sesiones activan `normativa` sin pasar por la skill de entrada (1 de 3, umbral 2). Las instrucciones del servidor ya
  remiten a las skills y no bastan. Los esquemas de las herramientas no
  llevan la descripción de ningún parámetro: la ayuda de un argumento no llega a `--describe`.
- **Entrega**:
  - Una skill de laboratorio, pequeña y fuera del producto: no va en `skills/`, no se empotra ni se instala. Usa las
    herramientas de kitlegal sobre lecturas grabadas y lleva una regla testigo que el modelo no puede saber de memoria
    y que un guion comprueba sin modelo en la respuesta.
  - Sus variantes, cada una distinta de la de control en una sola cosa, y las mismas evals para todas. Lo que se
    pregunta: la regla en el cuerpo (control); con las instrucciones del servidor MCP pidiendo activar la skill antes de la
    primera herramienta; con una skill de entrada que remite a la de laboratorio, como `consulta-juridica`; en una
    referencia que el cuerpo manda leer en una frase secundaria, como
    hoy; en una referencia que se lee como paso propio antes de la primera llamada; en una referencia que se manda leer
    en el punto en que hace falta, como `administrativo.md`; en la descripción del parámetro de la herramienta; y con
    la descripción de la skill sin nombrar herramientas.
  - Que la ayuda de cada argumento de un applet llegue a la descripción de su parámetro en el esquema de `--describe`,
    y con ella a las herramientas MCP y a la tabla de comandos: es un defecto de la base, conocido desde H32, y es lo
    que necesita la variante del esquema.
  - El arnés que lanza las variantes por modo en el sondeo de la persona (ADR 0032; el run no abre sesiones con
    modelo) y un guion que da, por variante y modo, las sesiones que activan la skill (y la de entrada) antes de la
    primera herramienta, las que abren la referencia y en qué turno, y las que cumplen la regla testigo, con su
    intervalo.
- **Fuera**: reorganizar las skills del producto, que es un hito aparte cuya entrada se escribe con el resultado;
  cambiar ningún umbral, límite ni tope, también el de 11 000 caracteres (ADR 0046); y guardar las sesiones del job de
  evals como artefacto, que sigue en `backlog/medida-de-las-skills.md`.
- **Controles**: `make ci` comprueba la regla testigo con respuestas fijas que la cumplen y que no, que las variantes
  difieren de la de control solo en lo declarado, que la skill de laboratorio no se empotra ni se instala, y la
  descripción de los parámetros en `--describe` y en las herramientas MCP. Las evals de las skills del producto en
  verde: el cambio de los esquemas es común y las mide todas.
- **Aceptación**: el sondeo de la persona, con `claude-sonnet-5-5`, 40 sesiones por variante y modo, deja su tabla en
  `evidencias/`, y un ADR que precisa el 0046 dice, con esas cifras, dónde va en una skill, y en el servidor, lo que hace
  que el agente la active y cumpla lo que pide en cada modo. Si ninguna referencia se lee de forma fiable en modo herramienta, el ADR lo dice y propone
  dónde va lo imprescindible.
- **Dependencias**: H32.
