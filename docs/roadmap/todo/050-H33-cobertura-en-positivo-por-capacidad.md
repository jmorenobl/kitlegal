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
    en prosa», y no exige ninguna línea fija (research D13 de H32).
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
      - **`normas`** son los identificadores que delimitan lo cubierto, cada uno en `data/normas.yaml` con el `vertical`
        de la fila. El esquema no admite una fila de materia sin normas, ni normas sin materia.
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
  - **Un único `⚠ NO CUBIERTO:`**, una sola plantilla en todas las materias:

    ```text
    ⚠ NO CUBIERTO: <asunto>: <motivo>. kitlegal no da sus plazos, órganos, vías ni escritos; lo que encuentre con la búsqueda del BOE de kitlegal va marcado ⚠ NO VERIFICADO.
    ```

    Y la marca de lo encontrado:

    ```text
    ⚠ NO VERIFICADO: <lo encontrado>: <forma legible> [<identificador>, bloque <id>]; kitlegal no ha comprobado que se aplique a este asunto.
    ```

    El `<motivo>` es, según el caso:
    - **fuera de lo declarado**: `no está en lo que kitlegal cubre`, fijo;
    - **lo regulado aparte con norma**: `se rige por <norma> (<identificador de la norma>), no por el procedimiento
      común (<precepto> [<identificador>, bloque <id>])`;
    - **lo regulado aparte sin norma**: `lo regula aparte la ley, no el procedimiento común (<precepto> [<identificador>,
      bloque <id>])`.

    Cómo se reparte:
    - **Las dos plantillas**, la línea y la marca, van iguales en el cuerpo de `consulta-juridica` y en el de
      `plazos-y-territorio` (`formas-compartidas`, ADR 0048).
    - **Los dos motivos de lo regulado aparte** van en la referencia del paquete, `administrativo.md`, porque hablan del
      procedimiento común, que es de una materia (principio X).
    - **El generador.** La región `cobertura_por_materia` genera cada fila con la plantilla del cuerpo y el motivo de su
      caso. Ningún texto de la línea vive en Go.
    - **La línea generada.** La de la aplicación de los tributos, por ejemplo, queda así:
      `⚠ NO CUBIERTO: aplicación de los tributos en materia tributaria y aduanera y su revisión en vía administrativa: se
      rige por Ley 58/2003, General Tributaria (BOE-A-2003-23186), no por el procedimiento común (DA 1ª.2.a LPAC
      [BOE-A-2015-10565, bloque daprimera]). kitlegal no da sus plazos, órganos, vías ni escritos; lo que encuentre con
      la búsqueda del BOE de kitlegal va marcado ⚠ NO VERIFICADO.`, en una sola línea.
  - **`consulta-juridica`**:
    - **§3.** La lista de lo que cubre la base sale de `cubierto`, en una región que genera `make skills-sync`, como la
      de los paquetes.
    - **§4** pasa a decir lo que no cubre con la línea:
      - toda parte de la pregunta que ninguna fila de `cubierto` resuelve lleva su `⚠ NO CUBIERTO:`, con el motivo de lo
        no declarado;
      - lo que el agente lea con `normativa` para completar la respuesta, como un plazo que fija la norma, va con su
        `⚠ NO VERIFICADO:`, citado;
      - ningún plazo, órgano ni vía de memoria, ni los de otra materia o los del procedimiento común, ni
        `plazos calcular` para un asunto de otra materia.
    - **El resto del cuerpo** y el frontmatter no cambian.
  - **`plazos-y-territorio`**:
    - **La plantilla.** «Cómo se presenta un plazo» lleva la plantilla única y la marca en lugar de las dos formas de
      hoy.
    - **Regla 9 y paso 7.** Un plazo, un órgano o una vía de una materia que no es la de ningún paquete se declaran con
      la línea y el motivo de lo no declarado, sin `plazos calcular` ni ninguna fecha.
    - **Lo regulado aparte** se declara como hoy, con la línea de su fila. Tras ella, lo que el agente lea de la norma
      que lo rige va marcado con `⚠ NO VERIFICADO:`, y no se dice de memoria. Esto sustituye a la orden de hoy de no
      exponerla, en la regla 9 y en `administrativo.md`, para que lo no cubierto se trate igual en todas las materias
      (ADR 0044, punto 4).
    - **Lo que no cambia.** Las órdenes de leer (ADR 0052 y 0053), el frontmatter y lo demás del cuerpo y de las
      referencias.
    - **El tope.** El cuerpo cabe en los 11 000 caracteres del ADR 0046: un borrador con la plantilla, la marca y la regla
      9 nuevas da 10 885, frente a 10 883 hoy.
  - **El paquete de materia incluye lo cubierto.** Un paquete trae ahora también sus filas de `cubierto` (ADR 0044,
    punto 5; contracts/paquetes.md de H32). La vertical de ensayo de `TestVerticalDeEnsayo` trae una y sigue entrando
    sin un `.go` ni un cambio en un `SKILL.md` fuera de las marcas.
  - **Evals.** Cuatro en `consulta-juridica`, una a cada lado de la frontera en dos materias sin paquete, las dos con
    norma en `data/normas.yaml` y sin `vertical`:

    | Materia | Dentro de lo cubierto | Fuera de lo cubierto |
    |---|---|---|
    | Seguros (Ley 50/1980, `BOE-A-1980-22501`) | **03, nueva**: «¿Qué dice la Ley de Contrato de Seguro sobre la prórroga de la póliza?». Comando y cita del art. 22 (`aveintidos`); `se_activan: [normativa]` | **01, la de la póliza**: pide además «qué podemos hacer», que es una vía. Gana `no_cubierto: {fuera_de_lo_declarado: true}`; lo demás sigue igual, y su comentario se corrige |
    | Laboral (ET, `BOE-A-2015-11430`) | **02, la de hoy**: los días de vacaciones (art. 38, `a38`), sin cambios | **04, nueva**: «Me despidieron el 2 de noviembre de 2026. ¿Qué dice el Estatuto de los Trabajadores del plazo para reclamar y hasta qué día puedo hacerlo?». Comando y cita del art. 59 (`a59`): leerlo es lo cubierto, y la fecha es lo no cubierto. Lleva `no_cubierto: {fuera_de_lo_declarado: true}` y prohíbe `plazos calcular` |

    - **Lo grabado.** El art. 22 de la Ley 50/1980 ya está grabado; el art. 59 del ET lo graba una tarea `[datos]` con
      `grabar_datos`.
    - **Las lecturas.** Las que provoque cada eval nueva van declaradas con su motivo (ADR 0043).
    - **Las evals de cobertura de `plazos-y-territorio`.** La 07, la 08 y la 12 no cambian de fichero: su línea esperada
      pasa a ser la de la plantilla única.
  - **`no_cubierto` y `cobertura_sin_declarar`**:
    - **El formato de eval** (`schemas/eval.yaml.json`) admite, además de `{norma, precepto}`, la variante
      `{fuera_de_lo_declarado: true}`, nunca las dos a la vez. Una eval con cualquiera de las dos es de cobertura.
    - **La respuesta declarada** es la que lleva una línea `⚠ NO CUBIERTO:` con las partes fijas de la plantilla:
      - la etiqueta y la cola, `kitlegal no da sus plazos, órganos, vías ni escritos; lo que encuentre con la búsqueda
        del BOE de kitlegal va marcado ⚠ NO VERIFICADO.`;
      - y además, en lo regulado aparte, el identificador de la norma esperada y la cita de su precepto, como hoy;
      - o, fuera de lo declarado, el motivo fijo `no está en lo que kitlegal cubre`.

      Una línea con otras palabras no cuenta, y con la línea `⚠ PLAZO:` tampoco.
    - **El umbral** se calcula en toda skill con alguna eval de cobertura. Desde H33 lo tienen
      `cobertura_sin_declarar:<modelo>:<modo>` de `consulta-juridica` (evals 01 y 04) y el de `plazos-y-territorio`
      (07, 08 y 12). Sigue con 0 de tope, porque hay forma (ADR 0038), y su descripción en el informe dice la cola.
    - **`⚠ NO VERIFICADO:`** no entra en ningún umbral (ver «Fuera»).
  - **Una regla nueva del conjunto de `consulta-juridica`, «frontera en dos materias».** Hay al menos dos normas de
    `data/normas.yaml` sin `vertical` que tienen cada una:
    - una eval activa que la cita y no es de cobertura;
    - y otra activa que la cita, es de cobertura con `fuera_de_lo_declarado` y prohíbe `plazos calcular`.

    La regla la cumplen las cuatro evals de arriba.
  - **El trabajo de evals.** `consulta-juridica` pasa de 2 a 4 evals y sigue a concurrencia 1. Su peor caso pasa de
    8485 s a 15 013 s, y cabe en el tope de 352 minutos (21 120 s). La concurrencia de las demás no cambia.
  - **Documentación.**
    - `docs/SKILLS.md`: la plantilla única, los motivos en la referencia del paquete y la tabla de defectos de la región,
      que hoy dice «las dos formas».
    - `docs/HISTORIAL.md`: la viñeta del hito.
    - `CHANGELOG.md`: «Cambiado».
    - `web/src/pages/como-se-evalua/index.astro`: la frase que anuncia H33 pasa a decir lo hecho.
  - **Lo generado.** Las referencias y las regiones se regeneran con `make skills-sync`.
  - **Lo que requiere `[datos]`.** `schemas/` y `testdata/` cambian solo en tareas `[datos]` (guardián de diff,
    `docs/WORKFLOW.md`).
- **Fuera**:
  - cubrir algo que hoy no se cubre. Ninguna fila de `cubierto` más allá de las cuatro de arriba, y ninguna capacidad
    nueva: destinatarios y documentos son H35 y H37;
  - cambiar las filas de `regulados_aparte` (sus procedimientos, preceptos y normas) o quitarlas;
  - cambiar `normativa` o `jurisprudencia`, o una `description` de cualquier skill (ADR 0055 y 0057);
  - cambiar en `consulta-juridica` o `plazos-y-territorio` algo distinto de lo que dice «Entrega»;
  - un umbral o una clase del juez para `⚠ NO VERIFICADO:`. Saber si una frase vino de una búsqueda pide leer su
    significado, y una clase del juez solo decide medida contra casos etiquetados (ADR 0037);
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
      - una fila de materia sin normas, o normas sin materia;
      - una norma que no está en `data/normas.yaml` o no tiene el `vertical` de su fila.

      Con `regulados_aparte`, igual que hoy.
    - **La región de `consulta-juridica`** sale de `cubierto` con diff vacío. La tabla de `administrativo.md` sale de la
      plantilla del cuerpo y del motivo de cada fila, entera y en una línea, sin texto de la línea en Go.
    - **`formas-compartidas`**: la plantilla `⚠ NO CUBIERTO:` y la marca `⚠ NO VERIFICADO:` son la misma línea en
      `consulta-juridica` y en `plazos-y-territorio`. `forma-de-cita` y `ordenes-de-ejemplo` siguen en verde: la cola no
      nombra ninguna orden de `kitlegal` que se pueda ejecutar.
    - **`cobertura_sin_declarar`, con mutantes** de respuestas grabadas a mano. Cuentan como sin declarar, cada una con
      su motivo:
      - una respuesta sin la línea, en una eval fuera de lo declarado y en una de lo regulado aparte;
      - la línea dicha con otras palabras o sin la cola;
      - en lo regulado aparte, sin la norma o sin la cita;
      - la línea con `⚠ PLAZO:`.

      Una respuesta con la línea entera no cuenta.
    - **La regla «frontera en dos materias»**: el conjunto de la rama la cumple, y un mutante sin la eval de un lado da
      su defecto.
    - **La vertical de ensayo** entra con su fila de `cubierto`.
    - **El tope** del cuerpo de cada skill y el peor caso del trabajo de evals bajo su tope (`TestDefinicionDelJob`).
  - **Tests de unidad que el prototipo puso en rojo por sus literales**, y que se actualizan con los datos nuevos sin
    cambiar lo que comprueban:
    - `TestLeerCobertura`, `TestRegionDeLaCobertura` y `TestRegiones` (`internal/skills`);
    - `TestDefinicionDelJob` (el peor caso de `consulta-juridica`, de 8485 s a 15 013 s);
    - `TestMatrizDeLasPropuestas` (la matriz de cada ejecución, con las evals que se miden);
    - y en `conjunto_test.go`, «las dos evals de la entrega» de `consulta-juridica` y el mutante de «no activación de
      cada capacidad». La eval 03 nueva también lleva `plazos-y-territorio` en `no_se_activan`, y el mutante tiene que
      quitarlo de las dos.
  - **Las suites congeladas siguen en verde, con tres cambios declarados aquí.** Son las únicas que se pusieron en rojo
    con `go test ./...` en una copia fuera del repositorio, con un prototipo de la plantilla en los dos cuerpos, la
    tabla regenerada, `cubierto` en los datos y las evals 03 y 04. El prototipo no cubría el generador ni el formato de
    eval nuevos:
    - **`h32-paquetes.txtar`** (de H32). Su `tributos.txt`, la fila de la aplicación de los tributos byte a byte, pasa a
      ser la línea de la plantilla única de arriba. Lo demás del guion no cambia: las 33 filas de `c991808` siguen con
      sus datos en la referencia.
    - **`TestSkillsDelProductoDesdeControl`**, el subtest de `plazos-y-territorio` (de H41, H41.1 y H41.4):
      - la declaración de `plazos-y-territorio` gana los cambios de H33, los del cuerpo y los de lo escrito a mano de
        `administrativo.md`;
      - la comparación quita también las regiones generadas de las referencias escritas a mano, como ya hace con las
        del `SKILL.md`.

      Es lo que hizo H41.1 con el subtest de H41. Las de `normativa` y `jurisprudencia` no cambian.
    - **`TestConsultaJuridicaDesdeLaBase`** (de H41.4). Hoy compara `skills/consulta-juridica/` con la base fija de
      H41.4 (`ab951c22`). Pasa a compararla con esa base y los cambios de H33, declarados como en
      `TestSkillsDelProductoDesdeControl`. Sus mutantes siguen dando su defecto exacto.

  Las evals salen en verde en la medición del cierre, en los dos modos, con `cobertura_sin_declarar` en 0 en
  `consulta-juridica` y en `plazos-y-territorio`. Si caen, la reparación no toca la plantilla, el motivo, la cola ni el
  umbral para que pase: lo deja dicho para la persona.
- **Aceptación**:
  - **e2e** (`testscript`) sobre `kitlegal skills install`:
    - el `SKILL.md` instalado de `consulta-juridica` y el de `plazos-y-territorio` llevan la plantilla y la marca, la
      misma línea en los dos;
    - la lista de lo cubierto de `consulta-juridica` nombra las cuatro filas de `cubierto`;
    - cada fila de `cobertura_por_materia` de `administrativo.md` es la plantilla con el motivo de su caso, con la de los
      tributos byte a byte;
    - y ningún cuerpo de capacidad dice «procedimiento común» en la plantilla.
  - **Las evals** de las cuatro skills, en verde en el cierre.
  - El run no lanza ningún sondeo.
- **Dependencias**: H32 y H41.4.
