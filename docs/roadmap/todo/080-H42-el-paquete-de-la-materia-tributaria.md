---
id: H42
fase: "1"
depende_de: [H45, H45.1, H46]
---

# H42 · El paquete de la materia tributaria, la segunda materia real de la base (ADR 0059 y 0062)

- **Objetivo**: que el molde del paquete de materia, arreglado en H45, se pruebe con una segunda materia real, no con
  la vertical de ensayo, y que un paquete saque por primera vez lo suyo de lo que la ley regula aparte. Una pregunta
  por el recurso de reposición o la reclamación económico-administrativa contra un acto de aplicación de los tributos
  del Estado se responde con la fecha del binario y su precepto. Lo tributario que sigue fuera se declara con la línea
  de siempre, con la norma que lo rige (ADR 0058, punto 1).
- **Decisiones que esta entrada aplica**: el ADR 0062, punto 6, y el molde que dejan sus puntos 1 a 5, 7 y 8 tras
  H45; el ADR 0059 (punto 2) y el ADR 0061. El run no los cambia ni escribe otro.
- **Solo con datos.** El hito entra con datos, referencias, evals y `testdata/`. Ninguna tarea declara un fichero `.go`,
  un `SKILL.md` escrito a mano ni `.github/`: lo hace verdad el guardián de diff, que solo deja tocar las rutas
  declaradas. Si el plan encuentra algo que pida Go, es un defecto del molde que va a un hito del molde, como H45.1, no una tarea de este
  hito.
  Los cuatro que encontró el prototipo del 2026-11-24 los arregla H45.1, que va antes. Lo que no es Go y esta entrada sí
  declara son los dos guiones de aceptación de H32 y H33 que fijan las filas tributarias de antes, con sus copias
  congeladas (punto 9 de «Entrega»).
- **Lo leído antes de escribirla**, con `kitlegal boe articulo` el 2026-11-21 y el 2026-11-22. Cada precepto que fija
  esta entrada, con su bloque, su versión y el día desde el que rige.
  - **Ley 58/2003, General Tributaria** (`BOE-A-2003-23186`):
    - **`a223`, versión 20031218**, vigente desde el 2004-07-01, la única redacción. El 223.1 da «un mes contado a
      partir del día siguiente al de la notificación del acto recurrible o del siguiente a aquel en que se produzcan
      los efectos del silencio administrativo», y, en las deudas de vencimiento periódico y notificación colectiva,
      desde el día siguiente al de finalización del período voluntario de pago.
    - **`a235`, versión 20150922**, de la Ley 34/2015 (`BOE-A-2015-10143`), vigente desde el 2015-10-12, con la nota
      de su disposición transitoria única.7. El 235.1 da «el plazo de un mes a contar desde el día siguiente al de la
      notificación del acto impugnado», con otros inicios para las retenciones, las repercusiones y la factura; en el
      silencio, «podrá interponerse la reclamación desde el día siguiente a aquél en que produzcan sus efectos». Con
      `--fecha 2015-10-11` sale la versión 20031218, cuyo apartado 1 dice otra cosa, así que el 235.1 rige así desde el
      2015-10-12. Un prototipo lo confirmó con `TestPreceptosDeLosPlazos` contra lo grabado, igual que el 2004-07-01
      del 223.1.
    - **`a222`, versión 20031218**: el recurso de reposición es potestativo, va antes de la reclamación, y con los dos
      interpuestos contra el mismo acto se tramita el primero (222.2).
    - **`a226`, versión 20091219**, vigente desde el 2009-12-20: la reclamación es de «la aplicación de los tributos del
      Estado o de los recargos establecidos sobre ellos y la imposición de sanciones tributarias que realicen la
      Administración General del Estado […] y las Administraciones tributarias de las Comunidades Autónomas» (226.a).
      La reposición es de los actos susceptibles de reclamación (222.1), así que los dos tipos tienen el mismo ámbito:
      los tributos del Estado, los aplique la Administración del Estado o la de una comunidad autónoma.
    - **`a234`, versión 20150922**, vigente desde el 2015-10-12: las reclamaciones «se tramitarán en única o primera
      instancia con los recursos que esta Ley establece» (234.1).
    - **`a7`, versión 20031218**: el 7.2 hace supletorias «las disposiciones generales del derecho administrativo».
    - **`a97`, versión 20031218**: los procedimientos tributarios se rigen por la LGT y, supletoriamente, por las
      disposiciones generales sobre los procedimientos administrativos. Es la norma de la fila de la DA 1ª.2.a desde H9.
    - **`a207`, versión 20031218**: «El procedimiento sancionador en materia tributaria se regulará» por el título IV
      de la LGT y, en su defecto, por las normas del procedimiento sancionador administrativo. No habla de lo aduanero.
    - Para dejarlos fuera: `a212` (recursos contra sanciones, versión 20121030), `a213` (medios de revisión), `a225`
      (resolución de la reposición, versión 20150922) y `a241` (recurso de alzada ordinario, versión 20150922).
  - **Ley 39/2015, LPAC** (`BOE-A-2015-10565`), `daprimera` y `a112`, versión 20151002, vigentes desde el 2016-10-02: la
    DA 1ª.2.a rige la aplicación de los tributos y su revisión «por su normativa específica y supletoriamente por lo
    dispuesto en esta Ley»; la 2.c, el sancionador tributario y aduanero; el 112.4, las reclamaciones
    económico-administrativas.
  - **LOFCA** (`BOE-A-1980-21166`), `avigesimo`, versión 20091219: las reclamaciones contra los actos de las
    comunidades autónomas «en relación con sus tributos propios» van a «sus propios órganos económico-administrativos»
    (20.Uno). Para dejar fuera los tributos propios autonómicos.
  - **TRLRHL** (`BOE-A-2004-4214`), `a14`, versión 20040309: contra los actos de aplicación de los tributos locales
    solo cabe el recurso de reposición que regula su apartado 2, con su propio plazo (14.2.c). Y la **LRBRL**
    (`BOE-A-1985-5392`), `a137`, versión 20031217: el órgano de las reclamaciones de los municipios de gran población,
    cuyo procedimiento va «de acuerdo en todo caso con lo establecido en la Ley General Tributaria» (137.5). Las dos,
    para dejar fuera lo local.
  - **Ley 35/2006, del IRPF** (`BOE-A-2006-20764`): `a63`, versión 20201231, la escala general, y `a74`, versión
    20101223, la escala autonómica, la que «hayan sido aprobadas por la Comunidad Autónoma». Para `fiscal.md` de
    `normativa`.
- **Lo comprobado en un prototipo** sobre `main` (`c5cf9a1b`), con el molde de H45 puesto a mano:
  - con los dos tipos tributarios y el cómputo `lpac`, el binario da las fechas de la tabla de «Controles», y
    `TestPreceptosDeLosPlazos` pasa con los bloques de la LGT grabados;
  - las filas tributarias pasan a `fiscal.md` y `administrativo.md` baja a 98 líneas;
  - con `a207` en `data/cobertura.yaml` y sin grabar, fallan `TestPreceptosDeLosPlazos` y unas veinte pruebas de
    sesiones: la grabación va antes que los datos;
  - el conjunto de `normativa` está cerrado en Go (`internal/evals/conjunto.go`: como mucho 23 evals y exactamente 10
    positivas, de normas distintas), y hoy tiene 23;
  - con las cuatro evals nuevas, `plazos-y-territorio` necesita la concurrencia 3 que trae H45: 5 h 14 min de peor caso
    sobre un tope de 352 min.
- **Lo comprobado en un segundo prototipo**, el 2026-11-24, sobre `main` después de H45 (`95442cf0`), con la entrada
  entera —las dos `fiscal.md`, los datos, las grabaciones en `evidencias/h42/`, la 07 cambiada y las evals 13 a 16— y
  sin Go de producción (rama `h42-prototipo`, que lleva además los cuatro tests parcheados a mano y, en su último
  commit, un cambio de `.github/` solo para la medida):
  - el binario da las fechas de T1 a T8; la sincronía reparte los dos tipos y las tres filas tributarias a `fiscal.md`
    y deja la aduanera en `administrativo.md`, que baja a 98 líneas; la `fiscal.md` de `plazos-y-territorio` tiene 71;
  - el cuerpo de `normativa` pasa de 10 750 a 10 820 caracteres y el de `plazos-y-territorio` de 10 686 a 10 756 (el
    guion de H33 le pone un tope de 10 860); las descripciones no cambian. Así, con un paquete de más, `normativa`
    pasa de 11 000 en `h45-paquete-de-mas`; con los cuerpos de H46 (ADR 0063), sobre su prototipo
    (`cuerpos-con-fiscal`), quedan en 9568 y 9458, y el guion sale en verde;
  - el peor caso de la medida es de 223, 187 y 49 sesiones (459), y el de `plazos-y-territorio`, con 16 evals a tres,
    de 5 h 14 min sobre 352 min;
  - caen cuatro tests de Go que el molde de H45 dejó atados al estado de hoy (los arregla H45.1) y los guiones de H32 y
    H33 del punto 9 de «Entrega»; con los cuatro arreglados y los guiones al día, `make ci` sale en verde;
  - una referencia a mano no puede enlazar otra (la sincronía lo da como defecto), y cada línea con `⚠` de una
    referencia a mano tiene que ser una de las del `SKILL.md` (`lineas-de-aviso-de-las-referencias`).
- **Decisiones de esta entrada**, cada una con su motivo:
  - **El cómputo es `lpac`, sin uno nuevo** (ADR 0062, punto 6). Los arts. 223.1 y 235.1 solo fijan la duración y el día
    siguiente; el art. 7.2 LGT y la DA 1ª.2.a LPAC llevan al art. 30 LPAC. Sale lo que ya cuenta `lpac`: de fecha a
    fecha, la prórroga, los dos municipios del art. 30.6 y agosto hábil. Como el art. 30 rige así desde el 2016-10-02,
    un plazo tributario que empieza antes no se calcula (ADR 0042).
  - **El IVA y el IS no entran.** Ninguna capacidad del paquete los necesita: los dos tipos son de la LGT, ninguna eval
    los lee, y `normativa` ya lee y cita cualquier norma del BOE sin paquete. Marcarlos pediría su identificador
    verificado, su búsqueda grabada y sus evals, para que la tabla de normas los liste como del paquete, que es el
    paso hacia la vertical fiscal que el ADR 0059 deja para la fase 3.
  - **`normativa` no gana evals.** Su conjunto está cerrado en 23 evals y 10 positivas de normas distintas, y la Ley
    del IRPF ya la cita la eval 06. Su `fiscal.md` la ejercen las evals que ya leen normas del paquete (04, 06 y 15).
  - **El TRLRHL sigue en lo administrativo.** Está entre las diez del paquete administrativo, y su revisión, la del
    art. 14.2, no es la de estos dos tipos. Moverlo cambiaría el paquete administrativo sin cubrir nada más.
  - **Lo local, lo foral, los tributos propios autonómicos y las sanciones quedan fuera, por la lectura
    conservadora.** Un tributo local se recurre en reposición por el art. 14.2 TRLRHL, no por el 223 LGT, y en los
    municipios de gran población hay además la reclamación del art. 137 LRBRL. Las haciendas forales tienen su propia
    ley general. Los tributos propios autonómicos se reclaman ante los órganos de la comunidad (art. 20.Uno LOFCA). La
    sanción es revisable por los mismos recursos (art. 213.1 LGT), pero su recurso tiene efectos propios (art. 212.3) y
    el ADR 0059 deja fuera el sancionador: entrar en él es otro hito.
  - **Solo la notificación como inicio.** Las deudas periódicas de notificación colectiva, las retenciones, las
    repercusiones, la factura y el silencio tienen otros inicios en el 223.1 y el 235.1. Quedan fuera de lo declarado,
    y `fiscal.md` no dice nada de ellos más que eso.
  - **El nombre del paquete es `fiscal`**, el `vertical` que ya fija el ADR 0059, para que la vertical fiscal de la fase
    3 lo amplíe y no lo duplique.
  - **Las palabras clave del 7.2 y del 207 LGT no se comprueban.** La tabla de `palabrasClaveDeLosPreceptos`
    (`internal/evals/grabaciones_test.go`) es Go por norma. Lo que sí se comprueba de los dos es que su bloque está
    grabado y, del 7.2, su `vigente_desde`. Queda como supuesto.
- **Entrega**, en este orden:
  1. **La grabación**, en una tarea `[datos]` que solo toca `internal/evals/testdata/grabaciones.json`: los bloques
     `a7`, `a207`, `a222`, `a223`, `a226`, `a234` y `a235` de la LGT y `a63` y `a74` de la Ley del IRPF, en la fila de
     cada norma, con su `para`. `a222` y `a226` se graban porque `fiscal.md` los nombra y la skill los lee, y `a7`
     porque `plazos calcular` lo devuelve en `data.preceptos`. `grabar_datos` los deja en `evidencias/h42/`, que el
     molde de H45 suma solo. El BOE consolidado ya tiene su fila revisada en `docs/SOURCES.md`.
  2. **`data/normas.yaml`**: `vertical: fiscal` en la LGT y en la Ley del IRPF, y nada más. El comentario de la cabecera
     nombra los dos paquetes.
  3. **`data/plazos.yaml`**, dos tipos con el cómputo `lpac`, `da: vencimiento` y un mes:
     - `reposicion-tributaria`, «recurso de reposición contra un acto de aplicación de los tributos», con
       `{norma: BOE-A-2003-23186, bloque: a223, apartado: "223.1", fija: duracion, vigente_desde: "2004-07-01"}`;
     - `reclamacion-economico-administrativa`, «reclamación económico-administrativa en única o primera instancia»,
       con `{norma: BOE-A-2003-23186, bloque: a235, apartado: "235.1", fija: duracion, vigente_desde: "2015-10-12"}`;
     - los dos, con
       `{norma: BOE-A-2003-23186, bloque: a7, apartado: "7.2", fija: supletoria, vigente_desde: "2004-07-01"}`.

     Los códigos y los nombres los puede afinar el spec; los preceptos y los días, no.
  4. **`data/cobertura.yaml`**:
     - en `cubierto`, una fila de `plazos` con `materia: fiscal`, la última, detrás de la del territorio (la región de lo
       cubierto empieza por las cuatro de la base, y el guion de H33 lo comprueba): «solo el recurso de reposición y la reclamación
       económico-administrativa en única o primera instancia contra un acto de aplicación de los tributos del Estado,
       los aplique la Administración del Estado o la de una comunidad autónoma, desde su notificación»;
     - en `regulados_aparte`, la fila de la DA 1ª.2.a se estrecha a «…, salvo el recurso de reposición y la
       reclamación económico-administrativa en única o primera instancia contra un acto de aplicación de los tributos
       del Estado» y conserva su norma, la LGT con `a97`. Con el molde de H45, pasa a `fiscal.md`. Sigue cubriendo los
       tributos locales, con su línea de hoy (ver «Riesgos»). Lo foral y los tributos propios autonómicos no van por
       ella: `fiscal.md` los da como no declarados;
     - la fila del sancionador «en materia tributaria y aduanera» se parte en dos: la tributaria gana su norma, la LGT
       con `a207`, y pasa a `fiscal.md`; la aduanera sigue sin norma en `administrativo.md`, porque el 207 no habla de
       ella;
     - la fila del art. 112.4 se estrecha a «reclamaciones económico-administrativas, salvo la reclamación en única o
       primera instancia contra un acto de aplicación de los tributos del Estado», gana su norma, la LGT con `a234` y el
       apartado 234.1, y pasa a `fiscal.md`.

     Lo que sale de lo regulado aparte es solo lo que cubren los dos tipos. Las demás filas no cambian, y
     `administrativo.md` queda sin filas tributarias salvo la aduanera.
  5. **`skills/plazos-y-territorio/references/fiscal.md`**, de menos de 100 líneas o con su índice:
     - las reglas de «Lo que la ley regula aparte» de `administrativo.md`, para sus filas, con la región
       `cobertura_por_materia` y sus dos motivos;
     - la región `recursos_y_plazos`, con los dos tipos;
     - lo que solo pide esta materia:
       - el ámbito de los dos tipos (art. 226.a): un acto de aplicación de los tributos del Estado, de la
         Administración del Estado o de una comunidad autónoma;
       - el inicio, el día de la notificación, y que con otro inicio el plazo no se pide;
       - los dos municipios, el del interesado y el de la sede del órgano que dictó el acto, que es también el que recibe
         la reclamación (235.3). Si no se sabe el segundo, se pide, como dicen el paso 8 del cuerpo y
         `calculo_de_plazos.md`, sin enlazarla: una referencia no enlaza otra;
       - que se elige una vía (222.2), y que los dos tipos dan la misma fecha;
       - que lo local va por su fila, y que lo foral y los tributos propios de una comunidad son de lo no declarado,
         sin nombrar las normas que los rigen.
  6. **`skills/normativa/references/fiscal.md`**:
     - la variación autonómica de los tributos cedidos: en el IRPF, la escala del Estado (art. 63) y la autonómica que
       aprueba cada comunidad (art. 74), que la respuesta señala y no suma de memoria;
     - las haciendas forales: su norma no está en `data/normas.yaml`, y si `kitlegal boe buscar` no la encuentra, la
       respuesta lo dice y no aplica la estatal;
     - los reglamentos de las leyes del paquete, con el identificador que dé `kitlegal boe buscar`, nunca de memoria, y
       su rango.
  7. **Lo que regenera `make skills-sync`**: la línea del paquete en los `SKILL.md` de `normativa` y de
     `plazos-y-territorio`, la fila de lo cubierto y la de los paquetes de la base en `consulta-juridica`, y las
     regiones. Ningún `SKILL.md` se escribe a mano.
  8. **Las evals** de «Evals», y **`CHANGELOG.md`**: en «Añadido», el paquete tributario con los dos plazos; en
     «Cambiado», la eval 07 de `plazos-y-territorio`.
  9. **Los guiones de H32 y H33 que fijan las filas tributarias de antes**, en una tarea `[datos]` con sus copias
     congeladas, como hizo H45 (T015): `internal/app/testdata/script/h32-paquetes.txtar` y
     `specs/030-h32-arquitectura-de-skills/aceptacion/paquetes.txtar` buscan cada fila de las referencias de `c991808`
     en alguna referencia de un paquete —y la del sancionador «en materia tributaria y aduanera» deja de estar: «filas:
     32 de 33»— y comparan byte a byte la fila de la aplicación de los tributos, que se estrecha;
     `internal/app/testdata/script/h33-cobertura-instalada.txtar` y
     `specs/039-h33-cobertura-en-positivo/aceptacion/cobertura-instalada.txtar` buscan esa fila byte a byte en
     `administrativo.md`. Los dos pasan a la fila nueva, en `fiscal.md`, y el de H32 deja fuera de su cuenta, nombrándola,
     la fila que este hito parte; el spec fija la forma, sin aflojar nada más de lo que comprueban.
- **Alcance**. El riesgo es crecer hacia la vertical fiscal. Fuera, en una lista cerrada:
  - el calendario del contribuyente y los plazos de fecha fija (ADR 0061, punto 5), los modelos y las
    autoliquidaciones;
  - el sancionador tributario, incluido el recurso contra una sanción;
  - las aduanas;
  - la aplicación de los tributos fuera de los dos recursos: la gestión, la inspección, la recaudación y sus plazos;
  - los procedimientos especiales de revisión (arts. 216 a 221), los recursos contra las resoluciones de los tribunales
    económico-administrativos (arts. 241 y siguientes) y los plazos para resolver (art. 225.4);
  - los tributos locales y la reclamación del art. 137 LRBRL;
  - los tributos propios de las comunidades autónomas (art. 20.Uno LOFCA);
  - las haciendas forales;
  - los inicios del 223.1 y del 235.1 que no son la notificación, y el silencio;
  - los escritos, los órganos y los destinatarios, salvo que H37 tome de aquí su segunda implementación (ADR 0059);
  - la doctrina de la DGT y del TEAC, que llega con H44;
  - el IVA y el IS en `data/normas.yaml`, y evals nuevas de `normativa`;
  - cualquier cambio en Go, en `.github/`, en `kitlegal plazos` o en el protocolo de una skill fuera de lo que genera la
    sincronía.
- **Controles**. Cada uno con la comprobación que lo pone en rojo, que el spec hace con un mutante o contra `main` y
  registra en `mutantes.md`.
  - **Los goldens de los dos tipos**, como guiones de la suite de aceptación con el binario, los festivos de
    `data/festivos/` y los dos municipios. Leganés es `28074`, y Madrid, `28079`. Fechas que dio el binario del
    prototipo:

    | # | Tipo | Inicio | Municipios | Fecha | Festivos saltados | Razonamiento |
    |---|---|---|---|---|---|---|
    | T1 | reposición | 2026-09-09 | Leganés, Madrid | 2026-10-13 | 2026-10-09, local de Leganés (`BOCM-20251212-34`); 2026-10-12, nacional (`BOE-A-2025-21667`) | De fecha a fecha, el viernes 9 de octubre, fiesta local de Leganés; 10 y 11, fin de semana; 12, fiesta nacional; martes 13 |
    | T2 | reclamación | 2026-09-09 | Leganés, Madrid | 2026-10-13 | los de T1 | La misma fecha que la reposición: los dos son un mes con el mismo cómputo |
    | T3 | reclamación | 2026-04-15 | Leganés, Madrid | 2026-05-18 | 2026-05-15, local de Madrid (`BOCM-20251212-34`) | El viernes 15 de mayo es fiesta en la sede del órgano y no en Leganés: inhábil por el art. 30.6; 16 y 17, fin de semana |
    | T4 | reposición | 2026-07-30 | Leganés, Madrid | 2026-08-31 | — | El 30 de agosto es domingo; agosto es hábil en este cómputo, a diferencia del contencioso |
    | T5 | reposición | 2026-01-31 | Leganés, Madrid | 2026-03-02 | — | Febrero no tiene 31: el 28, sábado; prórroga al lunes 2 de marzo |
    | T6 | reposición | 2026-12-09 | Leganés, Madrid | 2027-01-11 | — | El 9 de enero es sábado; con los avisos `faltan-festivos-locales` de los dos municipios, sin locales de 2027 |
    | T7 | reclamación | 2015-10-01 | Leganés, Madrid | — | — | Código 3 (ADR 0042): el mensaje nombra el 235.1, que rige así desde el 2015-10-12, y los apartados 30.2 a 30.6 LPAC, desde el 2016-10-02 |
    | T8 | reposición | 2026-09-09 | Leganés | — | — | Código 2: el cómputo cuenta con dos municipios, con el texto de sus datos |

    En rojo: un cómputo que solo mira los festivos del primer municipio (T3 da el 15 de mayo); agosto como periodo
    `inhabil` de `lpac` (T4 da el 1 de septiembre); el 235.1 con `vigente_desde` del 2004-07-01 (T7 deja de nombrarlo).
  - **Los preceptos contra lo grabado** (`TestPreceptosDeLosPlazos`): cada bloque grabado, el «un mes» del apartado que
    fija la duración y cada `vigente_desde`. En rojo: el del 235.1 con el 2004-07-01.
  - **Lo administrativo, igual**: `administrativo.md` sale igual byte a byte salvo las filas tributarias que pasan a
    `fiscal.md` y la aduanera que se separa. Los casos de `TestCasosDePlazos` y los goldens de H9 y H36 dan lo mismo
    que en `main`, salvo el mensaje de un tipo desconocido, que nombra también los dos tipos nuevos y que el molde de
    H45 ya no fija con una lista.
  - **Sin Go**: el diff del hito no tiene ningún `.go`, ningún fichero de `.github/` ni ningún `SKILL.md` fuera de las
    regiones que genera la sincronía. En rojo: cualquier otro cambio, que el guardián de diff rechaza.
  - **Las referencias**: `make skills-sync` no deja diferencias.
- **Evals**. Ninguna eval existente cambia salvo la 07 de `plazos-y-territorio`. El cierre mide `normativa`,
  `plazos-y-territorio` y `consulta-juridica`, cuyos textos cambian. Las cuatro nuevas son de `plazos-y-territorio`:
  - **13, la de la aceptación**: «Vivo en Leganés y el 9 de septiembre de 2026 me notificaron una liquidación del
    IRPF de la Delegación de la Agencia Tributaria en Madrid. ¿Hasta cuándo puedo recurrirla en reposición?». Pide
    `plazos calcular reposicion-tributaria` y la lectura de `a223` en esa fecha, y da T1 con su línea `⚠ PLAZO:` y la
    cita del 223.1. Sirve también `a7`, que la skill lee de `data.preceptos`.
  - **14**: «Vivo en Leganés y el 15 de abril de 2026 me notificaron una liquidación del IVA de la Delegación de la
    Agencia Tributaria en Madrid. Quiero reclamar directamente ante el tribunal económico-administrativo: ¿hasta
    cuándo puedo?». Es T3, con la lectura de `a235` en esa fecha.
  - **15, la negativa de la aceptación**: «El 9 de septiembre de 2026 me notificaron en Leganés una sanción de la
    Agencia Tributaria. ¿Cuánto tengo para recurrirla?». Prohíbe `plazos calcular`, y pide la línea de lo regulado
    aparte con la LGT y el bloque `daprimera`. El formato de hoy compara la norma y el bloque, no el apartado: no
    distingue la fila del sancionador de la de la DA 1ª.2.a, y eso lo dice su comentario.
  - **16**, lo foral: «Vivo en Bilbao y el 9 de septiembre de 2026 la Hacienda Foral de Bizkaia me notificó una
    liquidación del IRPF. ¿Hasta cuándo puedo recurrirla?». Prohíbe `plazos calcular` y pide la línea de lo no declarado
    (`no_cubierto: {fuera_de_lo_declarado: true}`, sin precepto, como la 04 de `consulta-juridica`; es la primera de
    `plazos-y-territorio`, y la deja pasar H45.1). Con el formato de hoy, eso lo cumple cualquier línea `⚠ NO CUBIERTO:`: mide que no
    se calcula, no que se elija la línea de lo no declarado y no la de la fila con la LGT. Lo segundo lo ve la lectura
    de una muestra de respuestas en la medida previa.
  - **La 07 cambia de pregunta y conserva su clase**: hoy pregunta por una liquidación de la Agencia Tributaria, que pasa
    a estar cubierta. Pasa a ser un tributo local: «Vivo en Leganés y el 9 de septiembre de 2026 el Ayuntamiento de
    Leganés me notificó la liquidación de la plusvalía. ¿Hasta cuándo puedo recurrirla?», con lo que espera hoy: la
    línea de lo regulado aparte con la LGT y `daprimera`, sin `plazos calcular`. Protege del error que más daño hace:
    contar un tributo local con el 223 LGT o con el art. 124 LPAC.
  - **Coste**. Cada medida de las tres skills que cambian es de unas 460 sesiones, sin los votos del juez: 223 de
    `normativa` (23 evals), 187 de `plazos-y-territorio` (16) y 49 de `consulta-juridica` (4), según el plan de
    `peorCaso`, medido en el prototipo del 2026-11-24. Las grabaciones nuevas de `evidencias/h42/` son de una raíz que
    sirve a todas las sesiones: si la decisión del job mide también `jurisprudencia`, son unas 120 más.
  - **Si una clase cae**, se compara con `main` y con la serie (ADR 0038). La reparación no toca para que pase ningún
    texto de skill fuera de las dos `fiscal.md`, y como mucho dos sondeos por pregunta. Si no se resuelve, se decide
    con lo que hay o se abre un hito.
- **Riesgos**:
  - **«Si el plazo fuera de un mes».** En la medida previa del 2026-11-24 (run 45370954378: Sonnet a 3/3 en todas las
    series y 0 en todas las clases que deciden), dos respuestas en modo herramienta, una de la 07 y otra de la 15, apuntan
    una duración como hipótesis («si el plazo es de un mes, como en otros recursos tributarios…»), sin fecha ni
    cómputo. Ningún umbral lo ve. Lo cierra una frase en la `fiscal.md` de `plazos-y-territorio`: tras la línea de una
    fila, ninguna duración, ni como hipótesis.
  - **La línea de lo local nombra la LGT.** La fila de la DA 1ª.2.a dice que lo suyo se rige por la LGT, y un tributo
    local se recurre por el art. 14.2 TRLRHL. Es la línea de `main`, y su función es que no se calcule con el
    procedimiento común ni con el 223 LGT. Darle a lo local una fila propia, con el TRLRHL, queda para un hito con su
    paquete local.
  - **Lo foral y los tributos propios autonómicos caen en el texto de la fila de la DA 1ª.2.a**, y solo la regla de
    `fiscal.md` los lleva a lo no declarado.
  - **Lo tributario sale de `administrativo.md`.** Una pregunta por un tributo del ayuntamiento que la skill lea como
    administrativa no encuentra ya allí la fila de la DA 1ª.2.a, y podría pedir el tipo `reposicion` de la LPAC. Lo
    mide la eval 07, y antes de lanzar, la medida previa.
  - **Las evals que pasan a leer `fiscal.md`**: la 04, la 06 y la 15 de `normativa` (la LGT y el IRPF), la 04 de las
    más movidas en H41 y H41.4; y, si la skill las clasifica como fiscales aunque el TRLRHL no sea del paquete, la 05 y
    la 14, de impuestos municipales.
- **Aceptación**:
  - en la app de Claude, «me notificaron una liquidación del IRPF el día D: ¿hasta cuándo puedo recurrirla en
    reposición?» da la fecha del binario con el 223.1, y pide el municipio de la sede si no lo da la pregunta;
  - «¿cuánto tengo para recurrir una sanción tributaria?» se declara no cubierto, con la LGT y sin plazo ni vía;
  - la liquidación de un tributo local de Leganés se declara no cubierta, sin fecha;
  - el hito no tiene ningún cambio en Go;
  - `administrativo.md` y los plazos de hoy salen como en `main`, salvo las filas que pasan a `fiscal.md`.
- **Antes de lanzar** (lo hace la persona, no el run):
  - que H45, H45.1 y H46 estén hechos y fusionados, y que el prototipo de `h42-prototipo`, rebasado sobre `main` sin sus
    tests parcheados ni su cambio de `.github/`, dé `make ci` en verde sin ningún `.go`;
  - **una medida previa sobre un prototipo completo** sobre el `main` de después de H46, con el job de evals de
    `normativa`, `plazos-y-territorio` y `consulta-juridica`, unas 460 sesiones. El prototipo lleva las dos `fiscal.md`
    escritas, los datos, la 07 cambiada y las evals 13 a 16. Hace falta porque cambia lo que leen las preguntas
    tributarias. Un rojo se compara con `main`, y como mucho un segundo sondeo por pregunta; se lee siempre una
    muestra de respuestas;
  - en delorean, `df -h /tmp`, y sitio en la cuota de `/tmp` de la persona: el `TMPDIR` de las sesiones es
    `/tmp/claude-<uid>`, y el 2026-11-24 los scratchpads de sesiones viejas la llenaban (`disk quota exceeded` al
    enlazar en `go test`);
  - el sondeo del alias: `claude -p --model sonnet --output-format json 'Responde solo: ok'` da en `modelUsage` el
    `MODELO_DE_EVALS` de `.github/workflows/evals.yml` (ADR 0031);
  - mover este fichero a `in-progress/`, sin su prefijo.
- **Dependencias**: H45, que deja el molde; H45.1, que suelta sus cuatro tests; y H46, que deja sitio en los cuerpos
  (ADR 0063). H37 puede tomar de aquí el órgano tributario como segundo destinatario (ADR
  0059), y H44 pone la doctrina sobre este paquete.
