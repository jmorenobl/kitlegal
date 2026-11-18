---
id: H33
fase: "1"
depende_de: [H32, H41.4]
---

# H33 · Cobertura en positivo por capacidad y materia

- **Objetivo**: que la cobertura diga lo que kitlegal cubre, por capacidad y materia, y que todo lo demás se declare no
  cubierto con una sola línea, igual en todas las materias (ADR 0044, punto 4; principio II). Hoy `data/cobertura.yaml`
  solo lista lo que la LPAC deja fuera del procedimiento común. Por eso una pregunta civil, de seguros o laboral no cae
  en ninguna fila, y `consulta-juridica` dice en prosa lo que no cubre, sin forma que un programa compruebe.
- **Evidencia de partida**:
  - **La bitácora** (`docs/USO.md`, 2026-10-21). En la póliza, kitlegal leyó y citó bien la Ley 50/1980, de Contrato de
    Seguro. Pero nada le dijo a la familia que los plazos, a quién reclamar y qué vía seguir no estaban cubiertos.
  - **Las evals de hoy.** La eval 01 de `consulta-juridica` (la póliza) dice en su comentario que lo que no cubre «lo dice
    en prosa», y no exige ninguna línea fija (research D13 de H32; ADR 0048, que describe así la skill y que el ADR 0044,
    punto 4, cambia).
  - **El umbral.** `cobertura_sin_declarar` solo existe en `plazos-y-territorio`, sobre las evals 07, 08 y 12, todas de
    lo que la ley regula aparte.
  - **La línea, en dos formas.** Las dos están en el cuerpo de `plazos-y-territorio`, y de ellas salen las nueve filas
    de `references/administrativo.md`.
- **Entrega**:
  - **`data/cobertura.yaml`, en dos listas**, validadas contra `schemas/cobertura.yaml.json`. Lo lee
    `skills.LeerCobertura`:
    - **`cubierto`, nueva**: lo que kitlegal cubre, una fila por capacidad y materia. Cada fila lleva `capacidad` (una de
      `normativa`, `jurisprudencia`, `plazos` o `territorio`, el vocabulario de las capacidades con herramienta de
      `consulta-juridica`), `alcance` (qué cubre, en una frase) y, si es de una materia, `materia` y `normas`.
      - **`materia`** es el nombre de un paquete, el `vertical:` de `data/normas.yaml`. La fila que no la lleva sirve a
        todas las materias.
      - **`normas`** son los identificadores que delimitan lo cubierto. El esquema no admite una fila de materia sin
        normas, ni normas sin materia. Que cada norma esté en `data/normas.yaml` con el `vertical` de su fila no lo
        puede comprobar un JSON Schema: lo comprueba la sincronía de las skills (`internal/skills`, en `make
        skills-sync` y en `make ci`), con su defecto, como hoy los preceptos de las regiones.
      - **Las filas** son las que el ADR 0044 declara cubiertas hoy, y ninguna más:
        - normativa, en todas las materias: leer y citar cualquier norma consolidada del BOE, estatal o autonómica, en
          cualquier fecha;
        - jurisprudencia, en todas: preparar la consulta de una sentencia, que la persona busca con su navegador, y
          cotejar el documento que trae;
        - territorio, en todas: la provincia, la comunidad, el régimen, el DIR3 del ayuntamiento y los boletines de un
          municipio;
        - plazos, en lo administrativo: los del procedimiento común y del acceso a la información ante cualquier
          administración y la interposición del contencioso, con los festivos de los municipios que cuentan. Sus normas
          son la LPAC (`BOE-A-2015-10565`), la LTAIBG (`BOE-A-2013-12887`) y la LJCA (`BOE-A-1998-16718`).
    - **`regulados_aparte`, la de hoy**, con sus nueve filas tal como están. Sigue siendo lo que la ley regula aparte del
      procedimiento común, con el precepto que lo dice y, si se cita de lo grabado, la norma que lo rige.
  - **Un único `⚠ NO CUBIERTO:`**, una sola plantilla en todas las materias, y la marca de lo leído:

    ```text
    ⚠ NO CUBIERTO: <asunto>: <motivo>. kitlegal no da sus plazos, órganos, vías ni escritos; lo que lea de su norma con kitlegal va marcado ⚠ NO VERIFICADO.
    ⚠ NO VERIFICADO: <lo encontrado>: <forma legible> [<identificador>, bloque <id>]; kitlegal no ha comprobado que se aplique a este asunto.
    ```

    - **La cola nombra la lectura, no la búsqueda.** «Lo que lea de su norma con kitlegal» son el índice y los artículos
      de `kitlegal boe`, que tienen las dos skills (paso 6 de `plazos-y-territorio`). No nombra `boe buscar`, que
      `plazos-y-territorio` no tiene en su tabla y cuyas búsquedas de texto libre las evals no sirven nunca (ADR 0043,
      puntos 2 y 3). Tampoco nombra una orden de `kitlegal` que `ordenes-de-ejemplo` vaya a ejecutar.
    - **Qué es lo no verificado.** El texto citado es el de la norma, leído con kitlegal y citado como en `normativa`. Lo
      no verificado es que se aplique a este asunto.
    - **El `<motivo>`**, según el caso:
      - **fuera de lo declarado**: `no está en lo que kitlegal cubre`, fijo;
      - **lo regulado aparte con norma**: `se rige por <norma> (<identificador de la norma>), no por el procedimiento
        común (<precepto> [<identificador>, bloque <id>])`;
      - **lo regulado aparte sin norma**: `lo regula aparte la ley, no el procedimiento común (<precepto> [<identificador>,
        bloque <id>])`.
    - **La línea generada.** La de la aplicación de los tributos, por ejemplo, queda así, en una sola línea:
      `⚠ NO CUBIERTO: aplicación de los tributos en materia tributaria y aduanera y su revisión en vía administrativa: se
      rige por Ley 58/2003, General Tributaria (BOE-A-2003-23186), no por el procedimiento común (DA 1ª.2.a LPAC
      [BOE-A-2015-10565, bloque daprimera]). kitlegal no da sus plazos, órganos, vías ni escritos; lo que lea de su norma
      con kitlegal va marcado ⚠ NO VERIFICADO.`
  - **Lo regulado aparte se queda como hoy en lo que se dice tras la línea.** Tras su línea no se expone la norma que
    rige el procedimiento —sus plazos, sus recursos— (ADR 0043, punto 7): la regla 9 de `plazos-y-territorio` y «Lo que
    la ley regula aparte» de `administrativo.md` lo siguen diciendo. Por eso, en lo regulado aparte no hay
    `⚠ NO VERIFICADO:`, y la cola de su línea no se cumple con nada. Que se pueda leer y marcar esa norma pide antes un
    ADR que sustituya ese punto, y no es de este hito.
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
    - **Probado.** Un prototipo de este contrato genera las nueve filas de `administrativo.md`.
  - **`consulta-juridica`**:
    - **El párrafo de entrada** dice que la skill declara lo que no cubre con su línea, que es suya, y que las demás
      líneas `⚠` las lleva cada capacidad.
    - **§3.** Las cuatro viñetas de las capacidades salen de `cubierto`, en una región que genera `make skills-sync`.
      Las viñetas de los paquetes de materia, ya generadas, y la de lo que no tiene herramienta, escrita a mano, se
      quedan.
    - **§4** pasa a decir lo que no cubre con su línea.
    - **El texto medido.** Del párrafo de entrada y de §4, el que dio el prototipo, palabra por palabra:

      > Esta skill no consulta nada: decide qué capacidad de kitlegal resuelve la pregunta, con qué skill, y qué cubre la
      > base, y declara lo que no cubre con su línea (paso 4). Lo que responde cada capacidad —su protocolo, sus reglas,
      > sus demás líneas `⚠` y sus órdenes de `kitlegal`, desde el `PATH` o con su herramienta— lo lleva la skill de esa
      > capacidad: actívala y síguela.

      > ### 4. Lo que no cubre, con su línea
      >
      > Toda parte de la pregunta que no resuelve ninguna capacidad de la lista del 3 —el plazo de una materia que no es
      > la del procedimiento administrativo común, a quién reclamar, qué vía seguir, un escrito— lleva su línea, con el
      > motivo «no está en lo que kitlegal cubre». Lo que leas con la skill de normativa para completarla va detrás,
      > citado, con su marca: el texto citado es el de la norma; lo que no está verificado es que se aplique a este
      > asunto. No des de memoria ningún plazo, órgano ni vía, ni apliques los de otra materia o los del procedimiento
      > común, ni pidas `kitlegal plazos calcular` para un asunto de otra materia: ese plazo no es el suyo.

      Detrás van las dos plantillas, en un bloque `text`, y el párrafo «Por qué» de hoy, sin cambios.
    - **El resto del cuerpo** y el frontmatter no cambian.
  - **`plazos-y-territorio`**: tres cambios del cuerpo, con el texto que dio el prototipo, y nada más.
    - **Paso 7.** Tras «**ni `plazos calcular`** ni ninguna fecha (regla 9)» se añade: «; si es de otra materia, igual,
      con el motivo de lo no declarado».
    - **«Cómo se presenta un plazo».** Las dos formas de hoy y su párrafo pasan a ser las dos plantillas, detrás de este
      párrafo:

      > Lo no cubierto (paso 7 y regla 9) va con esta forma: en un asunto de la cobertura del paquete, la línea de su
      > fila, entera; en otra materia, con el motivo «no está en lo que kitlegal cubre», y lo que leas de su norma,
      > detrás, con su marca:

    - **Regla 9.** Su última frase pasa a ser: «Un plazo, un órgano o una vía de otra materia tampoco: su línea, sin
      darlos de memoria ni pedir `plazos calcular`.».
    - **Lo que no cambia.** Las órdenes de leer (ADR 0052 y 0053), el frontmatter y lo demás del cuerpo.
    - **El tope.** Con los tres cambios, el cuerpo tiene 10 940 caracteres, frente a 10 883 hoy y 11 000 de máximo (ADR
      0046). Si la implementación necesita más, lo que sobra va a `administrativo.md` o a `calculo_de_plazos.md`, como
      movimiento declarado en `TestSkillsDelProductoDesdeControl`. No se acorta otro texto del cuerpo.
  - **`administrativo.md`.** Gana, en «Lo que la ley regula aparte», delante de la región, los dos motivos en un bloque
    `text`, con una frase que diga que el primero es el de la fila con norma. Lo demás escrito a mano no cambia, y la
    tabla se regenera.
  - **El paquete de materia incluye lo cubierto.** Un paquete trae ahora también sus filas de `cubierto` y, si trae
    región de cobertura, sus dos motivos (ADR 0044, punto 5; contracts/paquetes.md de H32). La vertical de ensayo de
    `TestVerticalDeEnsayo` gana una fila de `cubierto` y los dos motivos en su `references/ensayo.md`. Es `testdata/`,
    así que lo hace una tarea `[datos]`. Y sigue entrando sin un `.go` ni un cambio en un `SKILL.md` fuera de las marcas.
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
    - **La respuesta declarada** es la que lleva una línea `⚠ NO CUBIERTO:` con las partes fijas de la plantilla:
      - la cola, entera;
      - y además, en lo regulado aparte, el identificador de la norma esperada y la cita de su precepto, como hoy;
      - o, fuera de lo declarado, el motivo fijo.

      La línea con `⚠ PLAZO:` no cuenta.
    - **La tolerancia** es la de las etiquetas de hoy (`internal/evals/plazos.go`), aplicada a la línea entera: el
      énfasis de Markdown, el selector de presentación de las marcas (`⚠️` vale como `⚠`), los blancos juntados y las
      minúsculas. Nada más: una línea con otras palabras no cuenta.
    - **El umbral** se calcula en toda skill con alguna eval de cobertura. Desde H33 lo tienen
      `cobertura_sin_declarar:<modelo>:<modo>` de `consulta-juridica` (evals 01 y 04) y el de `plazos-y-territorio`
      (07, 08 y 12). Sigue con 0 de tope, porque hay forma (ADR 0038, punto 1), y su descripción en el informe dice la
      cola.
    - **`⚠ NO VERIFICADO:`** no entra en ningún umbral (ver «Fuera»).
  - **Una regla nueva del conjunto de `consulta-juridica`, «frontera en dos materias».** Hay al menos dos normas de
    `data/normas.yaml` sin `vertical` que tienen cada una:
    - una eval activa que la cita y no es de cobertura;
    - y otra activa que la cita, es de cobertura con `fuera_de_lo_declarado` y prohíbe `plazos calcular`.

    La regla la cumplen las cuatro evals de arriba.
  - **El trabajo de evals.** `consulta-juridica` pasa de 2 a 4 evals y sigue a concurrencia 1. Su peor caso pasa de
    8485 s a 15 013 s, y cabe en el tope de 352 minutos (21 120 s). La concurrencia de las demás no cambia.
  - **Documentación.**
    - `docs/SKILLS.md`: la plantilla, los motivos en la referencia del paquete y la tabla de defectos de la región, que
      hoy dice «las dos formas».
    - `docs/HISTORIAL.md`: la viñeta del hito.
    - `CHANGELOG.md`: «Cambiado».
    - `web/src/pages/como-se-evalua/index.astro`: la frase que anuncia H33 pasa a decir lo hecho.
  - **Lo generado.** Las referencias y las regiones se regeneran con `make skills-sync`.
- **Fuera**:
  - cubrir algo que hoy no se cubre. Ninguna fila de `cubierto` más allá de las cuatro de arriba, y ninguna capacidad
    nueva: destinatarios y documentos son H35 y H37;
  - cambiar las filas de `regulados_aparte` (sus procedimientos, preceptos y normas) o quitarlas;
  - exponer o marcar con `⚠ NO VERIFICADO:` la norma que rige lo regulado aparte (ADR 0043, punto 7);
  - cambiar `normativa` o `jurisprudencia`, o una `description` de cualquier skill (ADR 0055 y 0057);
  - cambiar en `consulta-juridica`, `plazos-y-territorio` o `administrativo.md` algo distinto de lo que dice «Entrega»;
  - un umbral o una clase del juez para `⚠ NO VERIFICADO:`. Saber si una frase vino de una lectura pide leer su
    significado, y una clase del juez solo decide medida contra casos etiquetados (ADR 0037);
  - un umbral para una línea `⚠ NO CUBIERTO:` de más en una eval de dentro de lo cubierto. Es un límite conocido: las
    evals 02 y 03 solo comprueban comando y cita, así que una skill que declarase todo no cubierto las pasaría. Medirlo
    con tope 0 sin una tasa de fondo arriesga el cierre por algo que nadie ha visto pasar (ADR 0038). Si el uso lo
    muestra, va a la bitácora;
  - cambiar el umbral de `cobertura_sin_declarar`, cómo cuentan los demás, o ningún otro umbral, límite ni tope (ADR
    0038, punto 1);
  - cambiar el job de evals más allá de lo que piden sus tests con las evals nuevas: ni `VERSION_DE_CLAUDE_CODE`, ni
    `MODELO_DE_EVALS`, ni lo que instala (ADR 0031);
  - que el binario o un servidor detecte la materia de una pregunta o lea una respuesta. La frontera la decide la skill
    con `cubierto`, y la comprueban las evals (principio X);
  - evals de más materias. La aceptación de la fase 1, con cinco o seis materias, no es de este hito.
- **Controles**:
  - `make ci` comprueba, sin red ni modelo:
    - **El esquema de la cobertura** rechaza, cada caso con su defecto:
      - una fila de `cubierto` sin `capacidad` o `alcance`, o con una capacidad fuera del vocabulario;
      - una fila de materia sin normas, o normas sin materia.

      La sincronía rechaza una norma que no está en `data/normas.yaml` o no tiene el `vertical` de su fila. Con
      `regulados_aparte`, igual que hoy.
    - **Las regiones.** La de `consulta-juridica` sale de `cubierto` con diff vacío. La tabla de `administrativo.md` sale
      de la plantilla y de los motivos de la referencia, entera y en una línea por fila. Cada defecto del contrato del
      generador tiene un mutante que lo da.
    - **`formas-compartidas`**: las dos plantillas son la misma línea en `consulta-juridica` y en `plazos-y-territorio`.
      `forma-de-cita` y `ordenes-de-ejemplo` siguen en verde con ellas (comprobado en el prototipo).
    - **`cobertura_sin_declarar`, con mutantes** de respuestas escritas como constantes en los tests. Cuentan como sin
      declarar, cada una con su motivo:
      - una respuesta sin la línea, en una eval fuera de lo declarado y en una de lo regulado aparte;
      - la línea dicha con otras palabras o sin la cola;
      - en lo regulado aparte, sin la norma o sin la cita;
      - la línea con `⚠ PLAZO:`.

      No cuentan la línea entera, ni la línea con negritas, con `⚠️` o en minúsculas.
    - **La regla «frontera en dos materias»**: el conjunto de la rama la cumple, y un mutante sin la eval de un lado da
      su defecto.
    - **La vertical de ensayo** entra con su fila de `cubierto` y sus motivos.
    - **El tope** del cuerpo de cada skill y el peor caso del trabajo de evals bajo su tope (`TestDefinicionDelJob`).
  - **Tests de unidad que el prototipo puso en rojo**. Se actualizan con los datos y la forma nuevos, sin cambiar lo que
    comprueban:
    - en `internal/skills`, `TestLeerCobertura` y `TestRegiones`;
    - en `internal/evals`:
      - con la línea vieja escrita como constante: `TestJuzgarPlazos`, `TestJuzgarLasLecturasSinServir`,
        `TestUmbralesDePlazosYTerritorio` y `TestUmbralesDeLasSkills`, y las de `informe_test.go`;
      - con los datos de las evals nuevas: `TestDefinicionDelJob` (el peor caso de `consulta-juridica`, de 8485 s a
        15 013 s), `TestMatrizDeLasPropuestas` (la matriz de cada ejecución) y `TestCacheDeLasSesiones` (el art. 59, que
        se graba);
      - en `conjunto_test.go`, dos: «las dos evals de la entrega» de `consulta-juridica`, y el mutante de «no
        activación de cada capacidad». La eval 03 también lleva `plazos-y-territorio` en `no_se_activan`, y el mutante
        tiene que quitarlo de las dos.

      `TestFormatoDePlazos` no tiene que cambiar (ver el `if`/`then` de arriba).
  - **Las suites congeladas siguen en verde, con cuatro cambios declarados aquí.** Son las únicas que se pusieron en rojo
    con `go test ./...`, en una copia fuera del repositorio. El prototipo llevaba lo que dice «Entrega»:
    - el generador de este contrato;
    - el formato de eval con `fuera_de_lo_declarado` y la comprobación de la cola;
    - los cambios de las dos skills y de `administrativo.md`;
    - `cubierto` en los datos;
    - y las evals 01, 03 y 04.

    No llevaba la región de §3 de `consulta-juridica` ni los motivos de la vertical de ensayo. Los cuatro cambios:
    - **`h32-paquetes.txtar`** (de H32). Su `tributos.txt`, la fila de la aplicación de los tributos byte a byte, pasa a
      ser la línea de arriba. Lo demás del guion no cambia: las 33 filas de `c991808` siguen con sus datos en la
      referencia.
    - **`h32-capacidades.txtar`** (de H32). En su `formas.txt`, la fila `plazos-y-territorio|⚠ NO CUBIERTO:|2` pasa a
      `|1`, y se añade `plazos-y-territorio|⚠ NO VERIFICADO:|1`. Hoy cae en `formas: 16 de 16`, y el total pasa a 17.
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
      - La huella pasa a ser la del `SKILL.md` que entrega H33, con sus regiones generadas reducidas a sus marcas, para
        que regenerar la lista de lo cubierto no la mueva.
      - El nombre de la base en los mensajes pasa a ser el de H33.
      - Sus mutantes siguen dando su defecto exacto.

      Lo que cambia del `SKILL.md` lo ve la persona en el diff de la propuesta.

  Las evals salen en verde en la medición del cierre, en los dos modos, con `cobertura_sin_declarar` en 0 en
  `consulta-juridica` y en `plazos-y-territorio`. Esa es la primera medida de la plantilla: nadie la ha medido antes con
  un modelo. Si caen, la reparación no toca la plantilla, el motivo, la cola, la tolerancia ni el umbral para que pase:
  lo deja dicho para la persona.
- **Aceptación**:
  - **e2e** (`testscript`) sobre `kitlegal skills install`:
    - el `SKILL.md` instalado de `consulta-juridica` y el de `plazos-y-territorio` llevan las dos plantillas, la misma
      línea en los dos;
    - la lista de lo cubierto de `consulta-juridica` nombra las cuatro filas de `cubierto`;
    - cada fila de `cobertura_por_materia` de `administrativo.md` es la plantilla con el motivo de su caso, con la de los
      tributos byte a byte;
    - y la plantilla no dice «procedimiento común» en ningún cuerpo: eso solo lo dicen los motivos de la referencia.
  - **Las evals** de las cuatro skills, en verde en el cierre.
  - El run no lanza ningún sondeo.
- **Dependencias**: H32 y H41.4.
