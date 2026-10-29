---
id: H32
fase: "1"
depende_de: [H31.1]
---

# H32 · Arquitectura de skills: una entrada genérica, skills por capacidad y las materias como referencias

- **Objetivo**: que una pregunta de cualquier materia entre por una skill que sabe qué capacidad la resuelve y qué
  cubre la base, y que una materia nueva sea un paquete de ficheros, no una skill ni código (ADR 0044, puntos 3 y 5).
  Al repartir las skills, que ninguna arrastre los tres defectos que quedan tras H31.1, medidos en el sondeo de la
  persona (`evidencias/sondeos-20261028/`, `brazo-P80`) y en los jobs de evals de la #207:
  - **`cuenta_su_proceso`: la respuesta da el resultado de `graph check`.** Es casi la única marca que queda. En el
    sondeo, las 4 de 80 son la línea de la comprobación («El bloque no lleva avisos de vigencia y no consta ninguna
    redacción posterior»). En el job de la cabeza fusionada de la #207 (`cd21c7b`, ejecución 40700995075),
    `claude-sonnet-5-5` marca 7 de 60 en el modo orden y 14 de 60 en el modo herramienta, y 17 de esas 21 son esa
    frase («La comprobación posterior no detectó que haya sido sustituida»); en la ejecución anterior (`0f9ef2f`,
    40682099061), las 16 de 120 salen de la comprobación. Se concentra en la eval 05: 6 de sus 12 respuestas en las
    dos ejecuciones, 5 por la comprobación. Antes de H31 eran el 1,8 % y el 2,8 %, y el cuerpo decía que un bloque sin `version-obsoleta` «no
    da nada a la respuesta: ni una línea, ni una frase junto a los avisos o a la fecha de vigencia, ni una palabra al
    final» (`e1bfb77`, `skills/boe-legislacion/SKILL.md`, líneas 215 a 220). H31 lo llevó a la referencia
    `redaccion_modificada`, que solo se lee con un `version-obsoleta`: es lo medido en `docs/SKILLS.md` §1, lo que va a
    una referencia como guía no se lee.
  - **`afirma_lo_no_leido`: reglas de memoria y materia pegada a una remisión.** Las dos marcadas de `brazo-P80` son
    «antes de tributar se restan los gastos deducibles y las reducciones que correspondan», una regla del rendimiento
    neto (los arts. 19 y 20 de la Ley 35/2006, según el voto del juez) para explicar «íntegros», y «están dichas "sin
    perjuicio" de otra regla de la misma Ley sobre rentas exentas», que describe el art. 7 sin su número. Otras cuatro
    respuestas llevan un voto «sí» por lo mismo, sin unanimidad (sesiones 23 y 29 del modo orden, 30 y 38 del modo
    herramienta). La segunda forma la permite el paso 5 tal como está:
    «o el número, o la materia» y «puedes avisar de que una materia se regula en otra parte»; la sesión quita el
    número y deja la materia pegada a la remisión. El art. 7 está grabado en las `lecturas` de la eval 15 desde H31.1,
    57 de las 80 respuestas lo nombran y ninguna sesión lo lee.
  - **Dos formas para la misma cita entre `legal-core` y `boe-legislacion`.** `boe-legislacion` cita un bloque entre
    corchetes, `[BOE-A-2015-10565, bloque daprimera]`; la línea `⚠ NO CUBIERTO:` de `legal-core` pide el identificador
    entre paréntesis, `(BOE-A-2015-10565)`, y la cobertura por materia lo escribe como `(BOE-A-2015-10565, bloque
    daprimera)`. Y el paso 6 de `legal-core` manda citar el texto de un artículo «como ella cita», la de
    `boe-legislacion`, que es lo que hacen las evals de cobertura al leer la DA 1ª. En
    la ejecución 40682099061 de la #207, la sesión `07-cobertura-liquidacion-tributaria-herramienta-claude-sonnet-5-5-02`
    escribió la cita de `boe-legislacion` dentro de la línea, con la segunda forma y sin la norma que lo rige, aunque
    su entrada la trae: «⚠ NO CUBIERTO: aplicación de los tributos … lo regula aparte la DA 1ª.2.a de la Ley 39/2015
    [BOE-A-2015-10565, bloque daprimera]: kitlegal no calcula sus plazos ni sus recursos». `cobertura_sin_declarar`
    dio 1 de 9 en el modo herramienta con umbral 0 y tumbó el job. En las dos ejecuciones de la #207 es 1 de 36
    respuestas de `claude-sonnet-5-5` en las evals 07, 08 y 12; a esa tasa, un umbral de 0 sobre 18 respuestas pasa el
    60 % de las veces. Las que pasaron usan una tercera forma, `(DA 1ª.2.a de la LPAC, BOE-A-2015-10565, bloque
    daprimera)`. Es el defecto que la guía de agentskills.io atribuye a skills demasiado estrechas que se cargan juntas
    («conflicting instructions»; «Design coherent units»), y el de terminología inconsistente de la de Anthropic («Use
    consistent terminology»).
- **Entrega**: (1) una skill de entrada con el método jurídico genérico: qué tipo de pregunta es, qué capacidad la
  resuelve, qué cubre la base y, si no lo cubre, cómo lo dice; (2) las skills por capacidad que hoy tienen herramienta
  —normativa, jurisprudencia, plazos y territorio—, con los nombres y el reparto que fije el plan, y su ADR (ADR 0044,
  punto 7); (3) cada materia como un fichero de referencia en cada skill de capacidad que la necesite, la organización
  por dominio de la guía de Anthropic («Pattern 2: Domain-specific organization»); (4) el **molde de vertical** —qué
  trae un paquete de materia: normas, referencias, cobertura y evals—, probado con una vertical de ensayo en
  `testdata/` que entra sin tocar Go ni ningún `SKILL.md`; (5) lo administrativo de hoy como primer paquete; (6) una
  sola forma de cita en todas las skills, con su control; (7) los arreglos de los tres defectos; (8) las evals, sus
  tests y el arnés de activación de las skills nuevas o renombradas; (9) `docs/SKILLS.md` con el reparto, el molde y lo
  medido; (10) `CHANGELOG.md` (*Unreleased*).
- **Alcance**:
  - **Las capacidades que se reparten son las que tienen herramienta.** Normativa (`kitlegal boe`, `graph check`),
    jurisprudencia (`kitlegal cita`), plazos (`kitlegal plazos`) y territorio (`kitlegal territorio`). Documentos y
    destinatarios no tienen herramienta genérica todavía: llegan con H35 y H37, y hasta entonces la skill de entrada los
    declara no cubiertos; el cotejo de una sentencia sigue en la de jurisprudencia. Una skill de capacidad sin
    herramienta solo podría decir de memoria lo que la base no cubre.
  - **El reparto, por unidades coherentes.** El plan decide qué skills hay, y cómo se reparten hoy `boe-legislacion`,
    `legal-core` y `jurisprudencia`, con el criterio de la guía de agentskills.io: una skill tan estrecha que una
    pregunta habitual necesite cargar otra con instrucciones que choquen es un defecto, y una tan ancha que no se active
    con precisión, también. Lo dice por cada pregunta de las evals de hoy: qué skills se activan y qué formas fijas da
    cada una.
  - **Cada skill de capacidad responde sola.** Su cuerpo lleva el protocolo, las reglas invariantes de `CLAUDE.md` y
    las formas fijas que dan sus respuestas, aunque otra skill las repita: lo que va a otra skill o a una referencia
    como guía no se lee (`docs/SKILLS.md` §1). La skill de entrada decide qué capacidad resuelve la pregunta y qué cubre
    la base; no lleva ninguna regla que otra skill necesite para responder.
  - **Los nombres.** Los fija el plan y los registra el ADR del hito: de capacidad, no de fuente —H34 añade EUR-Lex a
    la misma skill de normativa—; con el mismo patrón en todas, en español y como sustantivo, que la guía de Anthropic
    admite («Noun phrases»), y sin «anthropic» ni «claude». `legal-core` deja de presentarse como «derecho público»:
    la skill de entrada habla de cualquier materia, y la de plazos dice qué plazos cuenta hoy, los de la LPAC y la
    interposición del contencioso, hasta H36.
  - **Una sola forma de cita.** Un bloque se cita en todas las skills, en sus formas fijas y en sus referencias, como
    lo cita hoy `boe-legislacion`: `<forma legible> [<identificador>, bloque <id>]`. La línea `⚠ NO CUBIERTO:` cita así
    el precepto de la LPAC que lo regula aparte, y nombra la norma que lo rige por su nombre y su identificador, como
    hoy. La cobertura por materia generada da, por cada fila, la línea `⚠ NO CUBIERTO:` entera, lista para copiar, de
    modo que la sesión no elige entre dos plantillas («Templates for output format»); el cuerpo conserva la forma con
    sus marcadores. La medida de `cobertura_sin_declarar` exige, además de la norma esperada, el precepto de su fila
    con esa cita. Lo que H33 cambie de la línea —una sola en todas las materias, con la herramienta con que el agente
    busca— lo cambia sobre esta forma.
  - **La comprobación no da nada a la respuesta sin `version-obsoleta`.** Lo que la skill de antes de H31 decía de un
    bloque sin `version-obsoleta` —ni una línea, ni una frase junto a los avisos o la fecha de vigencia— vuelve al
    cuerpo de la skill de normativa, en el paso que pide la comprobación, como un *gotcha* de la guía de agentskills.io
    («Keep gotchas in `SKILL.md` where the agent reads them before encountering the situation»). La redacción la decide
    el plan, partiendo de la medida (`e1bfb77`), y la juzga el sondeo.
  - **Lo no leído no se explica ni se pega a una remisión.** Dos cambios en el paso 5 de la skill de normativa, con su
    porqué: una remisión no leída se repite con las palabras del texto leído, sin materia —sale el permiso de avisar de
    que una materia se regula en otra parte—, y explicar una palabra del texto leído con lo que regula otro precepto es
    afirmar ese precepto. Los pasos 3 y 6 y la primera viñeta del paso 4 no cambian, salvo lo que el plan muestre que
    pide la comprobación; los ejemplos no nombran ninguna norma de un paquete de materia.
  - **Las materias como referencias.** Lo que hoy es de lo administrativo en el cuerpo y las referencias de
    `legal-core` —la cobertura por materia, los recursos y plazos, las leyes vertebrales y lo que solo pide el
    procedimiento común— pasa al paquete de lo administrativo, con las referencias que hoy se generan desde `data/`
    generadas igual. Lo que vale para cualquier materia —la jerarquía normativa, el método— queda en la capacidad o en
    la entrada. Ningún cuerpo de una skill de capacidad nombra una norma de un paquete (principio X).
  - **El molde de vertical.** El plan fija dónde vive un paquete y cómo llega cada referencia a su skill. El enlace
    desde cada `SKILL.md` lo genera `make skills-sync` entre marcas, como la tabla de comandos, con la condición de su
    clase («si la pregunta es de <materia>»): una referencia sin el paso que la pide no se lee (`docs/SKILLS.md` §1), y
    así un paquete no edita ningún `SKILL.md` a mano. La vertical de ensayo de `testdata/` trae dos normas, una
    referencia por capacidad, una fila de cobertura y una eval; un test la instala sobre una copia de `skills/` y
    comprueba que pasa `skills-check` y que sus evals se cargan, y que para entrar no escribe ningún `.go` ni, fuera de
    lo que genera `make skills-sync`, ningún `SKILL.md`.
  - **El sitio en el cuerpo.** Cada cuerpo sigue en 11 000 caracteres como mucho (ADR 0046), y el de
    `boe-legislacion` tiene hoy 10 995. El sitio para los dos arreglos sale del reparto y, si no basta, como en H31.1:
    lo que solo pide una clase de pregunta, a la referencia de esa clase con el paso que la pide, y lo demás reescrito
    más corto con las guías de autoría. No salen del cuerpo las formas fijas, la tabla de comandos ni lo que toda
    respuesta necesita. El plan da, por cada texto movido o reescrito, el de hoy y el de la entrega.
  - **Las instalaciones existentes.** Una skill que cambia de nombre deja la del nombre anterior en `.agents/skills/` y
    `~/.agents/skills/` de quien la instaló, y dos skills con el mismo protocolo se activarían a la vez. El plan
    resuelve, con el manifiesto de `kitlegal skills install` (H19), que la instalación retire lo que instaló con el
    nombre anterior si sus huellas coinciden, y avise si no, con su e2e; y el plugin, el servidor MCP, la página
    `/instalar/` de la web, `README.md` y lo que nombra las skills en `internal/`.
  - **Las evals y su serie.** Las evals de cada skill pasan a la skill que hereda su pregunta, con su historia en git.
    El plan dice cómo siguen la serie de cierres y la reutilización de las medidas de `main` con un nombre nuevo (ADR
    0045), sin que ningún umbral cambie. Las evals nuevas: la activación de la skill de entrada con una pregunta de una
    materia que la base no cubre —la de la póliza de `docs/USO.md`, 2026-10-21, que lee la Ley 50/1980 y no da plazos,
    órganos ni vías de memoria— y con una que resuelve una sola capacidad; y la no activación de las de capacidad con
    las preguntas de otra.
  - **La suite de H31.1.** `internal/app/testdata/script/h31-1-h31-1-*.txtar` fija, palabra por palabra, el cuerpo y
    las referencias de `boe-legislacion` frente a `brazo-W`. Este hito cambia esos ficheros, así que retira los tres
    guiones en el mismo cambio y su suite de aceptación lleva lo que sigue valiendo: las formas fijas, la tabla de
    comandos, los pasos 3 y 6 y la primera viñeta del paso 4 iguales a los de `main` salvo los cambios de este hito, y
    ningún cuerpo de una skill de capacidad que nombre una norma de las evals por materia. `h19-skills-invocacion`
    sigue exigiendo `PATH` en cada `SKILL.md`, con los nombres nuevos.
  - **La duración.** Con skills nuevas y renombradas, el job mide la suite completa: el plan recalcula el peor caso de
    cada trabajo de evals y su tope antes de cerrarse.
- **Fuera de alcance**:
  - cualquier umbral del job, la rúbrica y los casos de los jueces, y que `cuenta_su_proceso` decida en el job: subir
    un umbral pide un ADR, y hacer decidir una clase que solo se publica, otro (ADR 0037 y 0038);
  - que `kitlegal boe articulo` dé la comprobación en su propio sobre, en lugar de que la sesión pida `graph check`:
    cambia el contrato de un applet y pide su ADR; antes de H31, el texto del cuerpo bastaba (1,8 % y 2,8 %);
  - que `kitlegal boe` devuelva la rúbrica o el id de los preceptos remitidos: la rúbrica en el índice no bajó las
    marcas (`evidencias/adr-0038/`); queda en el backlog;
  - las demás marcas de `cuenta_su_proceso` del job, que no son de la comprobación: dos de la eval 23 («antes de esa
    fecha el bloque no tenía redacción vigente»), una «no hace falta leer nada más» y una «el BOE no dice aquí qué norma
    la sustituye»;
  - la cobertura en positivo (H33), el motor de cómputo (H36), los ejes de territorio y los destinatarios (H37), los
    documentos (H35) y la actualización de las skills instaladas con el binario (H38);
  - las skills de documentos y destinatarios, y los packs (`packs/`), que llegan con la segunda vertical (ADR 0044,
    punto 5).
- **Controles**:
  - `make ci` en verde, con `make skills-check`, `ordenes-de-ejemplo`, `normas-nombradas` y los demás controles de
    contenido de `docs/SKILLS.md` §1, sobre las skills nuevas;
  - un control nuevo de una sola forma de cita, con su mutante en rojo: ningún `SKILL.md` ni ninguna referencia de
    ninguna skill cita un bloque con otra forma que `[<identificador>, bloque <id>]`, y cada forma fija que dos skills
    comparten es la misma línea en las dos;
  - el test de la vertical de ensayo y el e2e de la instalación con un nombre retirado;
  - las descripciones nuevas, juzgadas en tercera persona en el job (ADR 0046), de como mucho 900 caracteres;
  - el tope de los trabajos de evals, recalculado; `CHANGELOG.md` (*Unreleased*); y el job de evals en la propuesta de
    cambio, con todas las skills.
- **Aceptación**:
  - en el informe del job de cierre, los umbrales que deciden se cumplen en todas las skills, con los de hoy:
    `afirma_lo_no_leido:claude-sonnet-5-5` como mucho 0.028, y `cobertura_sin_declarar` 0 en los dos modos, con la
    medida que exige también la cita del precepto; ninguna sesión de las evals 14, 15 y 17 tiene en
    `fuera_de_lo_grabado` un bloque de sus `lecturas`. `cuenta_su_proceso` se publica por modo, sin cifra de aceptación
    en el cierre;
  - las evals de activación nuevas pasan, y las de hoy siguen pasando en la skill que hereda su pregunta;
  - **lo que la entrada deja a una persona antes de fusionar**: un sondeo con la skill de la propuesta, como los de
    `evidencias/sondeos-20261028/` —el mismo arnés, `claude-sonnet-5-5`, Claude Code 2.1.289, concurrencia 5—, con el
    límite de cada parte calculado como pide el ADR 0047: el objetivo, el número de respuestas y la probabilidad de pasar
    a la tasa objetivo y a una claramente peor. Son 280 sesiones, unas dos horas, y no se lanza con un job de evals en
    marcha. Si alguna parte no se cumple, la propuesta no se fusiona y se decide con sus cifras y sus respuestas
    marcadas; el sondeo no se repite para ver si sale en verde.

    | Parte | Skill y evals | Respuestas | Pasa con | A la tasa objetivo | A una peor |
    |---|---|---|---|---|---|
    | `afirma_lo_no_leido` | normativa, eval 15 | 80, 40 por modo | como mucho 6 | 5 %: 89 % | 12,5 %: 11 % |
    | `cuenta_su_proceso` | normativa, eval 15 | las mismas 80 | como mucho 8 | 5 %: 98 % | 12,5 %: 32 % |
    | `cuenta_su_proceso` | normativa, eval 05 | 80, 40 por modo | como mucho 6 | 5 %: 89 % | 15 %: 3 % |
    | la línea `⚠ NO CUBIERTO:` con la norma esperada y la cita del precepto | la que la lleve, evals 07, 08 y 12 | 120, 20 por eval y modo | como mucho 1 sin ella | 0,5 %: 88 % | 2,8 %: 15 % |

    - **La eval 15** repite el criterio del ADR 0047, con Fisher unilateral p < 0,05 frente al control de
      `evidencias/sondeos-20261025/`, 16 y 17 de 80. Protege que el reparto no pierda lo que ganó H31.1: con 80
      respuestas no se distingue el 2,5 % de `brazo-P80` de una tasa menor, y el arreglo del paso 5 lo deciden las
      marcas, no la cifra.
    - **La eval 05** es la que concentra la línea de la comprobación en el job. El objetivo, el 5 %, es la tasa de la
      eval 15 en `brazo-P80` y está por encima de la de antes de H31; la peor, el 15 %, está por debajo de la del job de
      `cd21c7b`, 21 de 120 (17,5 %). No tiene control en un sondeo: el límite es absoluto, y 6 de 80 frente a 21 de 120
      da p = 0,03 (Fisher unilateral).
    - **Las evals 07, 08 y 12** miden la línea sin el juez: el sondeo la juzga sin la traza, con el motivo de la
      declaración ausente. El objetivo, el 0,5 %, es el de una forma que la sesión copia de la referencia; la peor, el
      2,8 %, la de hoy, 1 de 36.

    Las probabilidades salen de la binomial, con la misma función que `evidencias/sondeos-20261026/guiones/analizar.py`.
