---
id: H33
fase: "1"
depende_de: [H32, H41.4]
---

# H33 · Cobertura en positivo por capacidad y materia

- **Objetivo**: que la cobertura diga lo que kitlegal cubre, por capacidad y materia, y que todo lo demás se declare no
  cubierto con una sola línea, igual en todas las materias (ADR 0044, punto 4; ADR 0058; principio II). Hoy
  `data/cobertura.yaml` solo lista lo que la LPAC deja fuera del procedimiento común. Por eso una pregunta civil, de
  seguros o laboral no cae en ninguna fila, y `consulta-juridica` dice en prosa lo que no cubre, sin forma que un
  programa compruebe.
- **Evidencia de partida**:
  - **La bitácora** (`docs/USO.md`, 2026-10-21). En la póliza, kitlegal leyó y citó bien la Ley 50/1980, de Contrato de
    Seguro. Pero nada le dijo a la familia que los plazos, a quién reclamar y qué vía seguir no estaban cubiertos.
  - **Las evals de hoy.** La eval 01 de `consulta-juridica` (la póliza) dice en su comentario que lo que no cubre «lo dice
    en prosa», y no exige ninguna línea fija (research D13 de H32; ADR 0048, punto 1, que el ADR 0058 sustituye en eso).
  - **El umbral.** `cobertura_sin_declarar` solo existe en `plazos-y-territorio`, sobre las evals 07, 08 y 12, todas de
    lo que la ley regula aparte.
  - **La línea, en dos formas.** Las dos están en el cuerpo de `plazos-y-territorio`, y de ellas salen las nueve filas
    de `references/administrativo.md`.
- **Decisiones que esta entrada aplica**: el ADR 0058, que va en la misma propuesta y precisa los ADR 0044 y 0048. El
  run no lo cambia ni escribe otro.
- **Antes de lanzar** (lo hace la persona, no el run): medir la plantilla con un modelo. Nadie la ha medido, y el cierre
  la congela con 0 de tope.
  - **El prototipo** es una rama que no se fusiona ni lleva propuesta de cambio. Lleva lo de «Entrega» que cambia lo
    que ve o hace una sesión:
    - los textos de las dos skills y de `administrativo.md`;
    - el generador, la región de lo cubierto y `data/cobertura.yaml`;
    - el formato de eval y la comprobación de la línea;
    - y las evals 01, 03 y 04.

    Además, dos cosas que el job comprueba antes de abrir ninguna sesión:
    - **lo grabado de la eval 04.** El job prepara la caché de cada sesión con las consultas de todas las evals de la
      skill (`PrepararSesion`, `internal/evals/preparar.go`), así que sin el art. 59 del ET grabado no se abre ninguna
      sesión de `consulta-juridica`. Lo graba la persona con `KITLEGAL_RECORD=1`, como lo haría `grabar_datos`, y
      `TestCacheDeLasSesiones` queda en verde;
    - **los tests del paso de comprobación de `scripts/evals.sh`** (`TestEvalsDelRepositorio` y
      `TestIdentificadoresDeLasNormas`), en verde con las cuatro evals. En `conjunto_test.go`, la cuenta de las evals de
      la entrega y el mutante de «no activación de cada capacidad», como dice «Controles».
  - **Cómo.** Tres despachos del job en modo completa sobre la rama, uno tras otro: `gh workflow run evals.yml --ref
    <rama>`. Miden las cuatro skills enteras, sin reutilizar nada. Cada uno tarda unos 35 minutos.
  - **Qué despachos cuentan.** Tres seguidos, desde el último cambio del prototipo, sin elegir los mejores. Si uno no
    cumple, se corrige y se vuelve a contar desde cero. El primero (ejecución 44355778768, sobre `f3b78ef4`) es
    evidencia, no uno de los tres: tras él cambió la regla del motivo (ver «`no_cubierto` y
    `cobertura_sin_declarar`»).
  - **Se lanza solo si**, en los tres despachos:
    - los cuatro trabajos salen en verde;
    - y `cobertura_sin_declarar` de `consulta-juridica` y de `plazos-y-territorio`, y `sin_activar` de
      `consulta-juridica`, salen en 0 en los dos modos.

    Son 18, 27 y 36 respuestas por modo. Un 0 no descarta una tasa de fondo baja (ADR 0058, punto 6): con 36
    respuestas en tres despachos, `cobertura_sin_declarar` de `consulta-juridica` aún deja pasar una tasa de hasta el
    8 %.
  - **Un rojo fuera de esos tres umbrales**, en otra skill o en otra clase, no se decide relanzando hasta que salga
    verde. Se compara con `main` en un sondeo con filas intercaladas (`INSTALACION=producto`, la eval que cae, `main`
    frente al prototipo), que lanza la persona. Si el prototipo cae más, es un defecto de esta entrada y se corrige
    antes de lanzar; si no, es la serie de `main`.
  - **La eval 17 de `normativa`, antes de los despachos.** Cuatro sondeos intercalados del laboratorio (eval 17,
    `INSTALACION=producto`, Claude Code 2.1.326; `~/code/kitlegal-sondeos/h33*` de quien los lanzó) y un despacho:
    - **El primer prototipo** (`65a0c850`), con un §4 que mandaba lo regulado aparte a «la línea de su fila, la de
      `plazos-y-territorio`», activaba `plazos-y-territorio` sin `normativa` en 35 de 160 sesiones; `main`, en 0.
    - **Sin esa mención** (`30aa4929`), un sondeo dio 1 de 160 con `plazos-y-territorio` delante, pero no validó el
      texto: 76 de sus 160 respuestas declaraban `⚠ NO CUBIERTO:` el cómputo del plazo de acceso, que kitlegal calcula.
      En el despacho siguiente (ejecución 44411742438) `normativa` cayó por `sin_activar` en la 17, y un sondeo de seis
      rondas dio 15 de 120 con `plazos-y-territorio` delante, 2 sin `normativa` y 34 de 120 con ese `⚠ NO CUBIERTO:`
      falso; `main`, 0 en los tres. Con el párrafo de entrada y la viñeta de los paquetes como hoy (`8a3a8f8f`), 9, 0
      y 42.
    - **El mecanismo.** La viñeta de los plazos de §3 omitía el acceso a la información, y el §4, con la línea
      obligatoria, lo volvía una declaración: o la sesión la escribía, o iba a `plazos-y-territorio` a comprobarlo y
      leía el artículo desde allí, sin `normativa`. Con `main`, la misma omisión en prosa no da ninguna. De ahí el
      alcance de los plazos que trae esta entrada y la segunda frase de §4.
    - **El laboratorio no reproduce la tasa del job**: con el mismo texto, 4 de 6 en el despacho frente a un 12,5 % en
      el sondeo. Ni el cliente (nativo o de npm) ni la hora lo explican; queda sin explicar.
    - **El sondeo que decide**, con el prototipo de hoy (`3ba19d30`) frente a `main`, en los dos modos, seis rondas de
      20 intercaladas, y cuatro criterios fijados antes de medir, cada uno en 0, como `main`: respuestas con
      `⚠ NO CUBIERTO:`; sesiones con `plazos-y-territorio` y sin `normativa`; sesiones que activan
      `plazos-y-territorio` en un mensaje anterior al de `normativa`; y sesiones sin `normativa`. Si cumple, los tres
      despachos se cuentan desde `3ba19d30`.
  - **Si no**, se corrige esta entrada con lo que muestren las respuestas, y se vuelve a medir.
- **Entrega**:
  - **`data/cobertura.yaml`**, validado contra `schemas/cobertura.yaml.json` y leído por `skills.LeerCobertura`, con dos
    listas y un texto:
    - **`motivo_de_lo_no_declarado`, nuevo**: `no está en lo que kitlegal cubre`. Es el motivo de la línea fuera de lo
      declarado, y su única fuente (ver «De dónde sale cada texto»).
    - **`cubierto`, nueva**: lo que kitlegal cubre, una fila por capacidad y materia.
      - **Cada fila** lleva `capacidad` (una de `normativa`, `jurisprudencia`, `plazos` o `territorio`, el vocabulario de
        las capacidades con herramienta de `consulta-juridica`) y `alcance` (qué cubre, el texto de su viñeta en §3 de
        `consulta-juridica`). Si es de una materia, lleva además `materia`.
      - **`materia`** es el nombre de un paquete, el `vertical:` de `data/normas.yaml`. La fila que no la lleva sirve a
        todas las materias.
      - **Que la `materia` sea un paquete**, un `vertical` de `data/normas.yaml`, no lo puede comprobar un JSON Schema.
        Lo comprueba la sincronía de las skills (`internal/skills`, en `make skills-sync` y en `make ci`), con su
        defecto, como hoy los preceptos de las regiones.
      - **Sin `normas`.** Lo que una fila cubre lo dice su `alcance`, y las normas de un paquete ya están en
        `data/normas.yaml` con su `vertical`. Una lista de normas en la fila no la leería nadie, y podría separarse de
        esas.
      - **Las filas** son las cuatro que `consulta-juridica` declara hoy en §3, con su texto, y ninguna más. El territorio
        va sin materia (ADR 0058, punto 5):

        ```yaml
        cubierto:
          - capacidad: normativa
            alcance: leer y citar cualquier norma consolidada del BOE, estatal o autonómica, de cualquier materia y en cualquier fecha.
          - capacidad: jurisprudencia
            alcance: preparar la consulta exacta de una sentencia, que la persona busca con su navegador, y cotejar el documento que trae. No consulta ningún buscador ni dice si una sentencia existe.
          - capacidad: plazos
            alcance: solo los del procedimiento administrativo común, los del acceso a la información pública y la interposición del recurso contencioso-administrativo, con los festivos de los municipios que cuentan. Lo que la ley regula aparte de ese procedimiento lo declara `plazos-y-territorio` con su línea.
            materia: administrativo
          - capacidad: territorio
            alcance: la provincia, la comunidad, el régimen, el DIR3 del ayuntamiento y los boletines de un municipio.
        ```

    - **`regulados_aparte`, la de hoy**, con sus nueve filas tal como están. Sigue siendo lo que la ley regula aparte del
      procedimiento común, con el precepto que lo dice y, si se cita de lo grabado, la norma que lo rige.
  - **Un único `⚠ NO CUBIERTO:`**, una sola plantilla en todas las materias (ADR 0058, punto 1), y la marca de lo leído,
    que solo lleva `consulta-juridica` (punto 2):

    ```text
    ⚠ NO CUBIERTO: <asunto>: <motivo>. kitlegal no da sus plazos, órganos, vías ni escritos.
    ⚠ NO VERIFICADO: <lo encontrado>: <forma legible> [<identificador>, bloque <id>]; kitlegal no ha comprobado que se aplique a este asunto.
    ```

    - **La cola** es `kitlegal no da sus plazos, órganos, vías ni escritos.` No nombra ninguna herramienta ni invita a
      leer nada. Qué se puede leer tras la línea lo dice cada skill: en `consulta-juridica`, lo que se lea con la skill
      de normativa fuera de lo declarado; en lo regulado aparte, nada (ADR 0043, punto 7).
    - **Qué es lo no verificado.** El texto citado es el de la norma, leído con kitlegal y citado como en `normativa`. Lo
      no verificado es que se aplique a este asunto.
    - **El `<motivo>`**, según el caso:
      - **fuera de lo declarado**: `no está en lo que kitlegal cubre`, el de `data/cobertura.yaml`;
      - **lo regulado aparte con norma**: `se rige por <norma> (<identificador de la norma>), no por el procedimiento
        común (<precepto> [<identificador>, bloque <id>])`;
      - **lo regulado aparte sin norma**: `lo regula aparte la ley, no el procedimiento común (<precepto> [<identificador>,
        bloque <id>])`.
    - **La línea generada.** La de la aplicación de los tributos queda así, en una sola línea (salida del prototipo):
      `⚠ NO CUBIERTO: aplicación de los tributos en materia tributaria y aduanera y su revisión en vía administrativa: se
      rige por Ley 58/2003, General Tributaria (BOE-A-2003-23186), no por el procedimiento común (DA 1ª.2.a LPAC
      [BOE-A-2015-10565, bloque daprimera]). kitlegal no da sus plazos, órganos, vías ni escritos.`
  - **Lo regulado aparte se queda como hoy en lo que se dice tras la línea.** Tras su línea no se expone la norma que
    rige el procedimiento, ni sus plazos ni sus recursos (ADR 0043, punto 7). Lo siguen diciendo la regla 9 de
    `plazos-y-territorio` y «Lo que la ley regula aparte» de `administrativo.md`. Por eso no hay `⚠ NO VERIFICADO:` en
    lo regulado aparte, y la cola no lo contradice.
  - **El generador de la región `cobertura_por_materia`** (`internal/skills/regiones.go`), con este contrato:
    - **La plantilla** es la única línea de un bloque `text` del `SKILL.md` que empieza por `⚠ NO CUBIERTO: ` y lleva
      `<motivo>`. Una skill con otra cantidad no tiene plantilla.
    - **Los motivos** van en la referencia del paquete que tiene la región. Son las líneas de sus bloques `text` que
      llevan `<precepto>` y no empiezan por `⚠`: una con `<norma>`, la de la fila con norma, y otra sin ella. Van ahí
      porque hablan del procedimiento común, que es de una materia (principio X).
    - **Cada fila** es la plantilla con `<asunto>` en lugar del procedimiento de la fila y `<motivo>` en lugar del
      motivo de su caso, con los marcadores de hoy sustituidos por sus datos. Ningún texto de la línea vive en Go.
    - **Los marcadores conocidos** son los de hoy más `<asunto>` y `<motivo>`.
    - **Los defectos**, cada uno con el fichero de su referencia:
      - una región en una skill sin plantilla;
      - una región en una referencia sin sus dos motivos, uno de cada;
      - un marcador que la tabla no conoce.
    - **Una fila sin región** sigue dando su defecto en cada skill con plantilla, que ahora son `consulta-juridica` y
      `plazos-y-territorio`.
  - **La región de lo cubierto**, nueva, en el `SKILL.md` que lleva sus marcas, hoy solo `consulta-juridica`. Es una
    región de `SKILL.md` como la de los paquetes de la base (`internal/skills/paquetes.go`), no una tabla de una
    referencia:
    - **Sus marcas**, cada una en su línea: `<!-- lo que cubre la base: lo genera make skills-sync -->` y
      `<!-- fin de lo que cubre la base -->`.
    - **Entre ellas**, una línea por fila de `cubierto`, en el orden del fichero: `- **<Capacidad>**: <alcance>`, con la
      primera letra de la capacidad en mayúscula. Dos filas de una capacidad dan dos líneas, como en la vertical de
      ensayo. `materia` no se escribe: el alcance ya dice a qué se ciñe.
    - **Con las cuatro filas de arriba**, la región repite las cuatro viñetas de hoy, cada una en una sola línea, y
      ninguna frase de hoy se pierde: «No consulta ningún buscador…» (ADR 0036), «solo los del procedimiento
      administrativo común» y «Lo que la ley regula aparte… lo declara `plazos-y-territorio` con su línea» van en sus
      alcances. Solo cambia la de los plazos, que nombra también los del acceso a la información pública:
      `data/plazos.yaml` los calcula (`acceso-resolver`, `acceso-reclamar`), y la viñeta de hoy los omitía. Con la línea
      obligatoria del §4, esa omisión hacía declarar no cubierto su cómputo (ver «Antes de lanzar»).
    - **Sin `data/cobertura.yaml`, o con defectos**, la región queda como está, y el defecto es el del fichero, como en
      la de los paquetes.
    - **El presupuesto de líneas.** `partirElCuerpo` (`internal/app/skills_test.go`) no parte las líneas de esta región,
      como ya no parte las de la tabla de comandos.
  - **De dónde sale cada texto que comprueba `internal/evals`.** Ninguno vive en Go:
    - **la cola**, de la plantilla de `skills/<skill>/SKILL.md` del árbol que se mide, con la función del generador que
      la busca: lo que sigue a `<motivo>. ` hasta el final de la línea;
    - **el motivo fijo**, de `motivo_de_lo_no_declarado` de `data/cobertura.yaml` del mismo árbol;
    - **y la norma y el precepto de lo regulado aparte**, de la eval, como hoy.

    Los casos sin esos textos:
    - **una eval de cobertura en una skill sin plantilla** es un defecto del conjunto, con su regla, y `make ci` no
      pasa. El juicio nunca corre sin plantilla. El laboratorio no juzga la cobertura, y sus variantes de
      `plazos-y-territorio` con las dos formas viejas no cambian;
    - **sin `data/cobertura.yaml`, o con defectos**, el conjunto da el defecto del fichero, y el juicio no corre.

    La sincronía comprueba que la prosa de cada skill con plantilla dice el motivo fijo tal cual, entre «». Antes junta
    los blancos seguidos, saltos de línea incluidos, porque el cuerpo va partido a 120 columnas. Si no lo dice, da su
    defecto. Sin `data/cobertura.yaml`, el defecto es el del fichero, como en la región.
  - **`consulta-juridica`**:
    - **El párrafo de entrada** no cambia. Un prototipo que decía en él que la skill declara lo que no cubre con su
      línea adelantaba `plazos-y-territorio` en la eval 17 de `normativa` (ver «Antes de lanzar»).
    - **§3.** Las cuatro viñetas de las capacidades pasan a ser la región de lo cubierto. Las viñetas de los paquetes
      de materia, ya generadas, y la de lo que no tiene herramienta, escrita a mano, se quedan.
    - **§4** pasa a decir lo que no cubre con su línea.
    - **El texto medido.** El de §4, del prototipo, palabra por palabra:

      > ### 4. Lo que no cubre, con su línea
      >
      > Lo que la pregunta pide y no resuelve ninguna capacidad de la lista del 3 —el plazo de una materia que no es la
      > del procedimiento administrativo común, a quién reclamar, qué vía seguir, un escrito— lleva su línea, con el
      > motivo «no está en lo que kitlegal cubre». Lo que no pide no lleva ninguna, tampoco el cómputo de una fecha que
      > no ha pedido. Esta línea no es para lo que la ley regula aparte del procedimiento común: eso lo declara, con su
      > propia línea, la skill de la capacidad que lo pide, y tras ella no leas su norma. Lo que leas con la skill de
      > normativa para completarla va detrás, citado, con su marca: el texto citado es el de la norma; lo que no está
      > verificado es que se aplique a este asunto. No des de memoria ningún plazo, órgano ni vía, ni apliques los de
      > otra materia o los del procedimiento común, ni pidas `kitlegal plazos calcular` para un asunto de otra materia:
      > ese plazo no es el suyo.

      La segunda frase deja sin línea lo que nadie preguntó. La tercera es la que impide que una pregunta de lo regulado
      aparte, que también es «el plazo de una materia que no es la del procedimiento común», reciba el motivo fijo y la
      lectura marcada de su norma (ADR 0043, punto 7). En una eval de lo regulado aparte, la línea con el motivo fijo
      cuenta como sin declarar.

      Detrás van la plantilla y la marca, en un bloque `text`, y el párrafo «Por qué» de hoy, sin cambios.
    - **La viñeta «Paquetes de materia»** de §3 no cambia, por la misma razón que el párrafo de entrada.
    - **El cuerpo** pasa de 3810 a 4483 caracteres. El resto del cuerpo y el frontmatter no cambian.
  - **`plazos-y-territorio`**: tres cambios del cuerpo, con el texto del prototipo, y nada más.
    - **Paso 7.** Tras «**ni `plazos calcular`** ni ninguna fecha (regla 9)» se añade: «; si es de otra materia, igual,
      con el motivo de lo no declarado».
    - **«Cómo se presenta un plazo».** Las dos formas de hoy y su párrafo pasan a ser la plantilla, sola en su bloque
      `text` y sin la marca, detrás de este párrafo:

      > Lo no cubierto (paso 7 y regla 9) va con esta forma: en un asunto de la cobertura del paquete, la línea de su
      > fila, entera; en otra materia, con el motivo «no está en lo que kitlegal cubre»:

    - **Regla 9.** Su última frase pasa a ser: «Un plazo, un órgano o una vía de otra materia tampoco: su línea, sin
      darlos de memoria ni pedir `plazos calcular`.».
    - **Lo que no cambia.** Las órdenes de leer (ADR 0052 y 0053), el frontmatter y lo demás del cuerpo.
    - **El tope.** Con los tres cambios, el cuerpo baja de 10 882 a 10 688 caracteres, frente a 11 000 de máximo (ADR
      0046). Cada paquete nuevo suma una línea de 58 caracteres más dos veces su nombre: la de la vertical de ensayo
      mide 70, y con ella el cuerpo queda en 10 758. Desde ahí caben tres más de 70. Si la implementación necesita más texto, no puede dejar menos de 140 caracteres libres. Lo que sobre
      va a `administrativo.md` o a `calculo_de_plazos.md`, como movimiento declarado en
      `TestSkillsDelProductoDesdeControl`. No se acorta otro texto del cuerpo.
  - **`administrativo.md`.** Gana, en «Lo que la ley regula aparte», delante de la región, la frase «Los motivos de la
    línea; el primero es el de la fila con norma:» y los dos motivos en un bloque `text`. Lo demás escrito a mano no
    cambia, y la tabla se regenera.
  - **El paquete de materia incluye lo cubierto** (ADR 0058, punto 4). Un paquete trae ahora también sus filas de
    `cubierto` y, si trae región de cobertura, sus dos motivos.
  - **La vertical de ensayo** de `TestVerticalDeEnsayo`, en una tarea `[datos]` porque es `testdata/`:
    - gana en su `data/cobertura.yaml` una fila de `cubierto`, `{capacidad: plazos, alcance: los de ensayo del molde de
      los paquetes., materia: ensayo}`, delante de `regulados_aparte`;
    - y en su `references/ensayo.md` de `plazos-y-territorio`, la frase y los dos motivos de `administrativo.md`.

    El test cambia en tres sitios:
    - `anadirEntradas` copia ahora las dos listas, cada una bajo su clave;
    - `partirPorLasMarcas` conoce las marcas de lo cubierto;
    - `comprobarLoSincronizado` comprueba que `consulta-juridica` lista la fila de la vertical.

    La vertical sigue entrando sin un `.go` ni un cambio en un `SKILL.md` fuera de las marcas (comprobado en el
    prototipo).
  - **Evals.** Cuatro en `consulta-juridica`, una a cada lado de la frontera en dos materias sin paquete, las dos con
    norma en `data/normas.yaml` y sin `vertical`:

    | Materia | Dentro de lo cubierto | Fuera de lo cubierto |
    |---|---|---|
    | Seguros (Ley 50/1980, `BOE-A-1980-22501`) | **03, nueva**: «¿Qué dice la Ley de Contrato de Seguro sobre la prórroga de la póliza?». Comando y cita del art. 22 (`aveintidos`), sin fecha; `se_activan: [normativa]`; `no_se_activan: [jurisprudencia, plazos-y-territorio]` | **01, la de la póliza**: pide además «qué podemos hacer», que es una vía. Gana `no_cubierto: {fuera_de_lo_declarado: true}`; lo demás sigue igual, y su comentario se corrige |
    | Laboral (ET, `BOE-A-2015-11430`) | **02, la de hoy**: los días de vacaciones (art. 38, `a38`), sin cambios | **04, nueva**: «Me despidieron el 2 de noviembre de 2026. ¿Qué dice el Estatuto de los Trabajadores del plazo para reclamar y hasta qué día puedo hacerlo?». Ver abajo |

    - **La 04, entera.** Su comando es `{applet: boe, norma: BOE-A-2015-11430, bloque: a59, fecha: "2026-11-02"}`.
      Lleva la fecha porque la pregunta trae un hecho fechado y `normativa` manda leer en esa fecha: un comando sin fecha
      solo lo cumple una lectura sin `--fecha` (`internal/evals/juzgar.go`). Lleva además:
      - la cita del art. 59 (`a59`), sin `redaccion_fechada`;
      - `prohibidos: [plazos calcular]`;
      - `no_cubierto: {fuera_de_lo_declarado: true}`;
      - y `no_se_activan: [jurisprudencia]`.

      No restringe más la activación: `plazos-y-territorio` puede activarse por «hasta qué día», y su paso 7 da la línea.
      Leer el art. 59 es lo cubierto; la fecha, lo no cubierto.
    - **Lo grabado.** El art. 22 de la Ley 50/1980 ya está grabado. El art. 59 del ET en esa fecha lo graba una tarea
      `[datos]` con `grabar_datos`.
    - **Las lecturas.** Ni la 03 ni la 04 declaran `lecturas`: el ADR 0043 (punto 2) solo admite las que se han visto en
      una sesión, y el run no abre sesiones con modelo (ADR 0032). Si en el cierre una sesión lee fuera de lo grabado,
      `fuera_de_lo_grabado` lo da y no decide. Las declara la persona después, con su motivo.
    - **Las evals de cobertura de `plazos-y-territorio`.** La 07, la 08 y la 12 no cambian de fichero: su línea esperada
      pasa a ser la de la plantilla.
  - **`no_cubierto` y `cobertura_sin_declarar`**:
    - **El formato de eval** (`schemas/eval.yaml.json`, en una tarea `[datos]`) admite, además de `{norma, precepto}`,
      la variante `{fuera_de_lo_declarado: true}`, nunca las dos a la vez. Una eval con cualquiera de las dos es de
      cobertura.
    - **Los mensajes de hoy.** Los que da el esquema a una eval `{norma, precepto}` mal escrita siguen igual: con un
      `if`/`then` sobre `fuera_de_lo_declarado`, no con `oneOf`. Un prototipo con `oneOf` cambiaba los de
      `TestFormatoDePlazos`.
    - **La respuesta declarada** es la que lleva una línea `⚠ NO CUBIERTO:`, encontrada por su etiqueta como hoy, que
      tras normalizarla contiene, también normalizadas:
      - la cola, sin su punto final;
      - y además, en lo regulado aparte, el identificador de la norma esperada y la cita de su precepto, como hoy;
      - o, fuera de lo declarado, el motivo fijo.

      La línea con `⚠ PLAZO:` no cuenta. El motivo no tiene que ir tras `: `: lo que une el asunto con el motivo es
      texto libre del modelo, como el asunto, y el cero del ADR 0038 va en lo que el modelo copia, el motivo y la cola.
      En el primer despacho del prototipo, una respuesta de la eval 04 con el motivo y la cola palabra por palabra
      cayó por llevar el asunto unido con «porque es un plazo laboral y».
    - **El motivo del juicio** de una respuesta sin declarar es `declaración de no cubierto ausente: ` seguido de la
      norma esperada, como hoy, o, en una eval `fuera_de_lo_declarado`, de `fuera de lo declarado`.
    - **La normalización**, sobre la línea y sobre lo que se busca en ella, y nada más:
      - quita `**`, `__`, `*`, `_` y las comillas invertidas;
      - quita el selector de presentación U+FE0F (`⚠️` vale como `⚠`);
      - junta los blancos seguidos en un espacio;
      - y pasa a minúsculas.
    - **Lo que cuenta y lo que no.** La comprobación es de subcadena, así que vale lo que vaya antes o después en la
      misma línea, también otro texto tras la cola o sin su punto final. No cuenta:
      - la línea partida en dos;
      - una palabra de la cola o del motivo cambiada, quitada o con otra tilde;
      - y la cola sin el motivo, o el motivo sin la cola.

      Sí cuenta cualquier unión entre el asunto y el motivo, con dos puntos, con una coma o con una frase.
    - **Los umbrales.** Se calculan en toda skill con alguna eval de cobertura:
      - `cobertura_sin_declarar:<modelo>:<modo>` lo tienen `consulta-juridica` (evals 01 y 04) y `plazos-y-territorio`
        (07, 08 y 12). Sigue con 0 de tope, porque hay forma (ADR 0038, punto 1). Su descripción en el informe, una
        constante de `internal/evals/umbrales.go` que no copia ningún texto de la línea, pasa a ser: «Respuestas de %s
        en el modo %s sin la línea ⚠ NO CUBIERTO: con la cola de su plantilla y el motivo esperado, o con una línea
        ⚠ PLAZO:, sobre sus respuestas medidas en las evals que esperan una declaración de no cubierto».
      - Con una eval de cobertura, `consulta-juridica` tiene umbrales de respuestas, y con ellos nace su
        `sin_activar:<modelo>:<modo>`, con 0 de tope, sobre las respuestas de sus cuatro evals (ADR 0058, punto 6). Su
        código no cambia: es el de las demás skills.
    - **`⚠ NO VERIFICADO:`** no entra en ningún umbral (ver «Fuera»).
  - **Una regla nueva del conjunto de `consulta-juridica`, «frontera en dos materias».** Hay al menos dos normas de
    `data/normas.yaml` sin `vertical` que tienen cada una:
    - una eval activa que la cita y no es de cobertura;
    - y otra activa que la cita, es de cobertura con `fuera_de_lo_declarado` y prohíbe `plazos calcular`.

    La regla la cumplen las cuatro evals de arriba. Pide normas sin `vertical`: el paquete que tome la Ley de Contrato
    de Seguro o el Estatuto de los Trabajadores tendrá que mover la eval de ese lado a otra norma sin paquete (ADR
    0058).
  - **El trabajo de evals.** `consulta-juridica` pasa de 2 a 4 evals y sigue a concurrencia 1. Su peor caso pasa de
    8485 s a 15 013 s, y cabe en el tope de 352 minutos (21 120 s). La concurrencia de las demás no cambia.
  - **Documentación.**
    - `docs/SKILLS.md`: la plantilla, los motivos en la referencia del paquete, la región de lo cubierto, la tabla de
      defectos de la región, que hoy dice «las dos formas», y la fila de `consulta-juridica` en la tabla de las skills,
      que dice «en prosa» y «sin… líneas `⚠`».
    - `CONTRIBUTING.md`:
      - la fila del campo `no_cubierto` en la tabla del formato de eval, con `fuera_de_lo_declarado`;
      - la fila de `cobertura_sin_declarar` en la tabla de umbrales, con su descripción nueva;
      - y los umbrales de `consulta-juridica`, con los dos que nacen.
    - `.github/workflows/evals.yml`: el comentario del tope que da 8485 s y «sus dos evals» para `consulta-juridica`.
      Solo el comentario: sus valores no cambian.
    - `docs/HISTORIAL.md`: la viñeta del hito.
    - `CHANGELOG.md`: «Cambiado».
    - `web/src/pages/como-se-evalua/index.astro`: la frase que anuncia H33 pasa a decir lo hecho.
    - El ADR 0058 ya va con esta entrada; el run no escribe ninguno.
  - **Lo generado.** Las referencias y las regiones se regeneran con `make skills-sync`.
- **Fuera**:
  - cubrir algo que hoy no se cubre, o dejar de cubrir algo. Ninguna fila de `cubierto` más allá de las cuatro de
    arriba, ninguna capacidad nueva (destinatarios y documentos son H35 y H37) y ningún cambio de texto en sus alcances;
  - cambiar las filas de `regulados_aparte` (sus procedimientos, preceptos y normas) o quitarlas;
  - exponer o marcar con `⚠ NO VERIFICADO:` la norma que rige lo regulado aparte (ADR 0043, punto 7; ADR 0058, punto
    2);
  - cambiar `normativa` o `jurisprudencia`, o una `description` de cualquier skill (ADR 0055 y 0057);
  - cambiar en `consulta-juridica`, `plazos-y-territorio` o `administrativo.md` algo distinto de lo que dice «Entrega»;
  - un umbral o una clase del juez para `⚠ NO VERIFICADO:`. Saber si una frase vino de una lectura pide leer su
    significado, y una clase del juez solo decide medida contra casos etiquetados (ADR 0037);
  - medir dos límites conocidos. Medirlos con 0 de tope sin una tasa de fondo arriesgaría el cierre por algo que nadie
    ha visto pasar (ADR 0038). Si el uso los muestra, van a la bitácora:
    - **una línea `⚠ NO CUBIERTO:` de más en una eval de dentro de lo cubierto.** Las evals 02 y 03 solo comprueban
      comando y cita, así que una skill que declarase todo no cubierto las pasaría;
    - **una fecha contada a mano en la prosa** de una respuesta que lleva la línea. `juzgarElPlazo`
      (`internal/evals/plazos.go`) solo falla por la línea ausente o por una línea `⚠ PLAZO:`, y una fecha suelta no
      cuenta (`TestUmbralesDePlazosYTerritorio/una-de-cobertura-con-una-fecha-suelta`). La eval 04 pasa con la línea y
      «puedes reclamar hasta el 30 de noviembre» en la prosa. Lo que la impide es el texto de §4, no un umbral;
  - cambiar el tope de `cobertura_sin_declarar` o de `sin_activar`, cómo cuentan los demás, o ningún otro umbral, límite
    ni tope, más allá de los dos que nacen en `consulta-juridica` (ADR 0038, punto 1);
  - cambiar el job de evals más allá de lo que piden sus tests con las evals nuevas y del comentario de su tope: ni
    `VERSION_DE_CLAUDE_CODE`, ni `MODELO_DE_EVALS`, ni lo que instala (ADR 0031);
  - que el binario o un servidor detecte la materia de una pregunta o lea una respuesta. La frontera la decide la skill
    con `cubierto`, y la comprueban las evals (principio X);
  - evals de más materias. La aceptación de la fase 1, con cinco o seis materias, no es de este hito.
- **Controles**:
  - `make ci` comprueba, sin red ni modelo:
    - **El esquema de la cobertura** rechaza, cada caso con su defecto:
      - una fila de `cubierto` sin `capacidad` o `alcance`, con una capacidad fuera del vocabulario, o con una clave
        que no declara, como `normas`;
      - un fichero sin `motivo_de_lo_no_declarado` o con él vacío.

      La sincronía rechaza una `materia` que no es el `vertical` de ninguna norma de `data/normas.yaml`, y una skill con
      plantilla cuya prosa no dice el motivo fijo. Con `regulados_aparte`, igual que hoy.
    - **Las regiones.** La de lo cubierto de `consulta-juridica` sale de `cubierto` con diff vacío. La tabla de
      `administrativo.md` sale de la plantilla y de los motivos de la referencia, entera y en una línea por fila. Cada
      defecto del contrato del generador tiene un mutante que lo da.
    - **`formas-compartidas`**: la plantilla es la misma línea en `consulta-juridica` y en `plazos-y-territorio`.
      `forma-de-cita` y `ordenes-de-ejemplo` siguen en verde (comprobado en el prototipo).
    - **`cobertura_sin_declarar`, con mutantes** de respuestas escritas como constantes en los tests. Cuentan como sin
      declarar, cada una con su motivo:
      - una respuesta sin la línea, en una eval fuera de lo declarado y en una de lo regulado aparte;
      - la línea con una palabra de la cola cambiada, o sin la cola;
      - fuera de lo declarado, con otro motivo;
      - en lo regulado aparte, sin la norma o sin la cita;
      - la línea partida en dos;
      - la línea con `⚠ PLAZO:`.

      No cuentan como sin declarar la línea entera, con negritas, con `⚠️`, en minúsculas, sin el punto final o con
      texto detrás.
    - **De dónde salen los textos**: un mutante con otra cola en la plantilla de la skill, y otro con otro motivo en
      `data/cobertura.yaml`, cambian lo que la eval exige, sin tocar ningún `.go`. Una eval de cobertura en una skill
      sin plantilla da el defecto del conjunto, y una prosa partida en dos líneas con el motivo dentro no da defecto.
    - **La regla «frontera en dos materias»**: el conjunto de la rama la cumple, y un mutante sin la eval de un lado da
      su defecto.
    - **La vertical de ensayo** entra con su fila de `cubierto` y sus motivos, y su fila sale en la región de lo cubierto
      de `consulta-juridica`.
    - **El tope** del cuerpo de cada skill, también con la línea de la vertical de ensayo (`TestVerticalDeEnsayo`), y el
      peor caso del trabajo de evals bajo su tope (`TestDefinicionDelJob`).
  - **Tests de unidad que el prototipo puso en rojo**, con `go test ./...` y con `-tags=integration` en
    `internal/{evals,skills,app,workflow}` y `data`. La lista es completa. Se actualizan con los datos y la forma nuevos;
    lo que cambia lo que comprueban va dicho:
    - en `internal/skills`:
      - `TestLeerCobertura`, `TestRegiones` y `TestRegionDeLaCobertura`, que lleva la fila de los tributos byte a byte;
      - `TestVerticalDeEnsayo`, con los tres cambios de la vertical de arriba;
    - en `internal/app`, `TestSkillsDelRepositorio` (sus presupuestos de líneas de `consulta-juridica`): `partirElCuerpo`
      no parte la región de lo cubierto;
    - en `internal/evals`:
      - con la línea vieja escrita como constante: `TestJuzgarPlazos`, `TestJuzgarLasLecturasSinServir` y
        `TestUmbralesDePlazosYTerritorio`, y las de `informe_test.go`;
      - `TestUmbralesDeLasSkills`, que cambia lo que comprueba. Su comentario fija que `consulta-juridica` va «sin
        `sin_activar`». Pasa a llevar `sin_activar` y `cobertura_sin_declarar` del modelo que decide en cada modo, los
        cuatro con `Decide: true` y delante de sus cuatro `fuera_de_lo_grabado` (ADR 0058, punto 6);
      - con los datos de las evals nuevas: `TestDefinicionDelJob` (el peor caso de `consulta-juridica`, de 8485 s a
        15 013 s), `TestMatrizDeLasPropuestas` (la matriz de cada ejecución) y `TestCacheDeLasSesiones` (el art. 59, que
        se graba);
      - en `conjunto_test.go` (`TestEvalsDelRepositorio`), dos subtests: «formato», con las dos evals de la entrega de
        `consulta-juridica`, y el mutante de «no activación de cada capacidad». La eval 03 también lleva
        `plazos-y-territorio` en `no_se_activan`, y el mutante tiene que quitarlo de las dos.

      `TestFormatoDePlazos` no tiene que cambiar (ver el `if`/`then` de arriba).
  - **Las suites congeladas siguen en verde, con cuatro cambios declarados aquí.** Son las únicas que se pusieron en rojo
    en el prototipo. Los dos guiones de H32 están en `internal/app/testdata/script/`: la tarea `[datos]` que cambia sus
    ficheros es la misma que cambia la skill, para que `make ci` quede en verde en cada tarea. Sus copias de
    `specs/030-h32-*/aceptacion/` son el registro de H32, no se ejecutan y no cambian.
    - **`h32-paquetes.txtar`** (de H32). Su `tributos.txt`, la fila de la aplicación de los tributos byte a byte, pasa a
      ser la línea de arriba en su fila de la tabla. Lo demás del guion no cambia: las 33 filas de `c991808` siguen con
      sus datos en la referencia.
    - **`h32-capacidades.txtar`** (de H32). En su `formas.txt`, la fila `plazos-y-territorio|⚠ NO CUBIERTO:|2` pasa a
      `|1`. El total sigue en `formas: 16 de 16`: no mira `consulta-juridica`, y `plazos-y-territorio` no lleva la marca.
    - **`TestSkillsDelProductoDesdeControl`**, el subtest de `plazos-y-territorio` (de H41, H41.1 y H41.4):
      - la declaración de `plazos-y-territorio` gana los tres cambios del cuerpo de arriba;
      - y gana un cambio en lo escrito a mano de `administrativo.md`, los motivos. Hoy la declaración solo admite cambios
        del `SKILL.md` y añadidos al final de una referencia, así que se amplía para admitir un cambio declarado en una
        referencia;
      - la comparación quita también las regiones generadas de las referencias escritas a mano, como ya hace con las
        del `SKILL.md`.

      Es lo que hizo H41.1 con el subtest de H41. Los de `normativa` y `jurisprudencia` no cambian.
    - **`TestConsultaJuridicaDesdeLaBase`** (de H41.4). Guarda la huella sha256 del `SKILL.md` de `consulta-juridica` en
      `ab951c22`, y no hay en el repositorio texto de esa base sobre el que declarar cambios.
      - La huella pasa a ser la del `SKILL.md` que entrega H33, con sus dos regiones generadas reducidas a sus marcas,
        para que regenerar la lista de lo cubierto o la de los paquetes no la mueva.
      - El nombre de la base en los mensajes pasa a ser el de H33.
      - Sus mutantes siguen dando su defecto exacto.

      Lo que cambia del `SKILL.md` lo ve la persona en el diff de la propuesta.

  Las evals salen en verde en la medición del cierre, en los dos modos, con `cobertura_sin_declarar` en 0 en
  `consulta-juridica` y en `plazos-y-territorio` y `sin_activar` en 0 en `consulta-juridica`. Las primeras medidas son las
  tres de «Antes de lanzar». Si en el cierre caen, la reparación no toca la plantilla, el motivo, la cola, la normalización ni
  los umbrales para que pasen: lo deja dicho para la persona.
- **Aceptación**:
  - **e2e** (`testscript`) sobre `kitlegal skills install`:
    - el `SKILL.md` instalado de `consulta-juridica` y el de `plazos-y-territorio` llevan la plantilla, la misma línea
      en los dos, y solo el de `consulta-juridica` lleva la marca;
    - la región de lo cubierto de `consulta-juridica` tiene una línea por cada una de las cuatro filas de `cubierto`;
    - cada fila de `cobertura_por_materia` de `administrativo.md` es la plantilla con el motivo de su caso, con la de los
      tributos byte a byte;
    - y la línea de la plantilla de cada `SKILL.md` no contiene «procedimiento común»: eso solo lo dicen los motivos de
      la referencia.
  - **Las evals** de las cuatro skills, en verde en el cierre.
  - El run no lanza ningún sondeo.
- **Dependencias**: H32 y H41.4.
