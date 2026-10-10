# Changelog

Todo cambio de comportamiento visible de `kitlegal` se registra aquí.

El formato sigue [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/) y el proyecto se adhiere al
[versionado semántico](https://semver.org/lang/es/). Mientras el mayor sea `0` —la primera release publicada,
`v0.1.1`, es la de H19 (ADR 0019)— un cambio incompatible sube el **menor**. Este fichero se mantiene **a mano**, también
desde esa release: cada propuesta de cambio añade su entrada bajo *Unreleased* en el mismo cambio que
introduce el comportamiento, y al publicar una versión esa sección se cierra bajo su número y su fecha y
se abre una nueva vacía. Las notas de cada release las genera goreleaser desde los Conventional Commits, y no
sustituyen a este fichero.

## [Unreleased]

### Añadido

- **`kitlegal cita preparar` y `kitlegal cita cotejar`: el applet `cita`, para las sentencias** (ADR 0036). kitlegal
  no consulta el buscador del CENDOJ, que responde con un CAPTCHA al programa que se identifica como programa: la
  búsqueda la hace la persona con su navegador, y el applet hace lo de alrededor. Sus dos verbos no piden nada a
  ninguna fuente ni escriben en la caché o en el grafo del mundo —`--offline` y `--no-graph` no cambian nada—, y
  ninguno es el verbo por omisión: `kitlegal cita` termina con `2` nombrando los dos. Su sobre lleva `fuente`
  `kitlegal.cita`, que dice que no se ha consultado ninguna fuente.
  - **`kitlegal cita preparar`** dice qué tiene que hacer la persona para encontrar una sentencia; no dice si
    existe. Con una referencia —un ECLI español como argumento (`ECLI:ES:TS:2023:3144`), `--roj "STS 3144/2023"` o
    `--resolucion 1088/2023 --fecha 2023-07-04`—, `data` lleva la `referencia` reconocida, la `direccion` del buscador
    (`https://www.poderjudicial.es/search/indexAN.jsp`) y sus `casillas`, cada una con el `nombre` que la persona ve
    y el `valor` que tiene que escribir: «ECLI», «Nº ROJ», o «Nº Resolución» y, dos veces, «Fecha resolución», con
    `campo` «Desde» y «Hasta» y la fecha `dd/mm/aaaa`. Lleva además `equivalente` cuando se deduce sin consultar nada:
    el ROJ de un ECLI del Tribunal Supremo, o el ECLI de un ROJ de siglas `STS`. Con `--texto "cláusula suelo"`, para
    una búsqueda por materia, `data` lleva el `texto` y la `direccion` que abre el buscador con esa búsqueda hecha
    (`https://www.poderjudicial.es/search/sentencias/<texto codificado>/1/AN`), con `casillas` vacía. Un ECLI del
    Tribunal Constitucional, que no está en el CENDOJ, se reconoce y se declara fuera de cobertura dentro de `data`
    —`cobertura`, con `cendoj` `no-cubierto` y su `motivo`—, sin `direccion`, y termina con `0`. La `url` del sobre es
    la `direccion`, o `kitlegal:applet/cita` si `data` no lleva ninguna.
  - **`kitlegal cita cotejar`** lee la ficha con la que el CENDOJ encabeza cada documento —la línea
    `Roj: <ROJ> - <ECLI>` y, debajo, «Órgano», «Fecha», «Nº de Recurso», «Nº de Resolución», «Ponente» y «Tipo de
    Resolución»— del texto que le llega por la entrada estándar o en `--documento`, y, si se le da la referencia que
    se había pedido, con las tres formas de `preparar`, dice si el documento es ese. `data` lleva la `ficha`, con sus
    ocho datos; `correspondencia`, si el ROJ y el ECLI de la ficha son de la misma resolución (`se-corresponden`,
    `no-se-corresponden` o `no-se-deduce`); y, con una referencia, `pedida` y `es_la_pedida`. **Que el documento no
    sea el pedido es un hallazgo, con salida `0`** (ADR 0023): `hallazgos` lleva uno de la clase
    `documento-distinto`, que nombra cada dato que difiere, con el pedido y el del documento, y, en `cruce`, si el
    número pedido como ROJ es el número de resolución del documento (`numero-de-resolucion`) o al revés
    (`numero-del-roj`). Del documento no sale nada más que la ficha, y la `url` del sobre es la huella del texto
    recibido, `kitlegal:documento/sha256:<64 hex>`. No abre ficheros: un PDF lo lee el agente, que le pasa la ficha.
  - **Códigos.** `0`, también con un hallazgo o con una sentencia fuera de cobertura; y `2`, `argumentos`, con un
    ECLI mal formado o de otro país, un ROJ o un número de resolución sin su forma, `--resolucion` sin `--fecha` o
    `--fecha` sin `--resolucion`, una fecha que no es un día, más de una forma de referencia, `preparar` sin
    referencia ni `--texto` o con las dos cosas, un `--texto` vacío, y `cotejar` sin ningún texto, con un texto sin
    la línea `Roj:` o con una ficha a la que le falta un dato o que lleva sin su forma el ROJ, el ECLI, la fecha o el
    número de resolución. Un argumento escrito con valor vacío (`--roj ""`) está dado, y es el error de su argumento.
- **Las herramientas `cita_preparar` y `cita_cotejar`** en `kitlegal mcp serve`, que pasa a anunciar doce. Salen del
  registro, como las demás, y cada llamada devuelve el sobre de su orden; sus argumentos van por su nombre (`ecli`,
  `roj`, `resolucion`, `fecha`, `texto`, `documento`). Una llamada no tiene entrada estándar: `cita_cotejar` recibe
  el texto en `documento`, y sin él, o con él vacío, es un error de la clase `argumentos`.
- **Esquema publicado de los dos verbos**: `schemas/cita.json`, el que emiten `kitlegal cita preparar --describe` y
  `kitlegal cita cotejar --describe`, que `make schema-check` compara como los demás.
- **La skill `jurisprudencia` v0.1: ninguna sentencia citada de memoria.** Se activa cuando la pregunta nombra, pide,
  cita o resume una sentencia o un auto, pregunta si existe, pide jurisprudencia sobre una materia o trae el texto o
  el PDF de una resolución. Va empotrada en el binario con las otras dos, y `kitlegal skills install` la instala.
  - **Una sentencia solo se cita con su documento en la conversación** —pegado o adjunto— y su ficha cotejada con
    `cita cotejar`, con una forma fija y los datos de la ficha:
    `STS 1088/2023, de 4 de julio [ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023]`. La respuesta dice que la cita sale del
    documento aportado, y no escribe ningún ECLI que no venga de una operación o de la persona.
  - **De la que no está delante**, la respuesta lleva una línea de forma fija,
    `⚠ SENTENCIA NO COMPROBADA: <la referencia, como se dio>`, y debajo la consulta que prepara `cita preparar`: la
    dirección del buscador y cada casilla con su valor. Una cita escrita como «STS 1088/2023, de 4 de julio» se
    prepara como número de resolución con su fecha; sin fecha no se prepara, tampoco como ROJ, y se pide la fecha.
    No dice que la sentencia existe ni que no existe: nadie lo ha comprobado.
  - **No le anuncia un CAPTCHA a la persona.** El buscador del CENDOJ cierra el paso con un CAPTCHA a los
    programas, y por eso kitlegal no lo consulta; a la persona, con su navegador, no le sale. La respuesta no le
    dice que le saldrá uno ni le pide que resuelva ninguno.
  - **El equivalente deducido se da como deducido.** Si la respuesta nombra el ROJ o el ECLI que `cita preparar`
    deduce de la referencia (`equivalente`), lo dice como lo que es: deducido de ella, sin que nadie lo haya
    comprobado, y no como un dato de la sentencia.
  - **Si el documento traído no es el pedido**, lo dice con lo que difiere y no lo cita como si lo fuera.
  - **Ante una pregunta por materia**, da la dirección de la búsqueda por texto, y no cita ninguna sentencia de
    memoria: sigue con las que traiga la persona.
  - **Una sentencia del Tribunal Constitucional** la declara no cubierta, con la dirección del buscador del
    tribunal, sin citarla.
  - **No resume ni caracteriza una sentencia cuyo texto no está en la conversación**: con la ficha sola da la cita y
    sus datos. En el job de evals lo decide el juez con modelo de la skill, con `afirma_lo_no_leido` (abajo, en
    «Cambiado»).
  - Pide cada operación con su herramienta, si el agente la tiene, y si no, con su orden, siempre con `--json`. Sin
    herramienta y sin binario no cita ninguna sentencia. `SKILL.md` tiene 197 líneas y no lleva `references/`.
  - **Sustituye a la v0 de H23, que no llegó a publicarse**, y solo cambia de ella dos pasajes de `SKILL.md`. Su
    párrafo inicial decía que «el buscador pide un CAPTCHA a los programas, y no se sortea», sin decir a quién
    afecta, y 8 de las 35 respuestas del modelo que decide en el cierre de H23 nombraban un CAPTCHA, algunas para
    anunciárselo a la persona o pedirle que lo resolviera. Y dejaba nombrar el equivalente junto a la referencia sin
    decir qué es: de las 6 respuestas de ese cierre que lo daban, solo 1 decía que es deducido. La `description`, el
    protocolo, la forma de la cita, la de la línea `⚠ SENTENCIA NO COMPROBADA:`, las reglas y la tabla de comandos
    no cambian. Que las respuestas dejen de hacerlo no lo mide ningún control del job: lo lee una persona en las
    del cierre.

### Cambiado

- **La web y el README explican la instalación con una sola pieza, el plugin.** https://kitlegal.es/instalar/, la
  portada, `/llms.txt` y el README pasan de dos pasos —la extensión y el plugin— a uno: en la app de escritorio de
  Claude, añadir el marketplace `jmorenobl/kitlegal-plugins`, instalar el plugin `kitlegal` y cerrar y abrir la app.
  Ya no ofrecen `kitlegal.mcpb`, que sigue en cada release; `kitlegal-plugin.zip` queda para quien prefiera no añadir
  un marketplace. A quien tenía la extensión le dicen que la desinstale, y que tras actualizar el plugin las
  herramientas pasan solas a la versión nueva. Probado el 2026-10-04 con la v0.5.0 y el marketplace real, en la app
  de escritorio de Claude en macOS: desinstalada la extensión y actualizado el plugin, la app arranca al abrirse un
  solo servidor, el del plugin, con sus diez herramientas, y la respuesta lleva la cita con su forma (`docs/USO.md`).
- **`boe-legislacion` v0.1.7: de un precepto que no ha leído, nada.** De un artículo, un apartado o una disposición
  que ninguna orden ni herramienta le ha devuelto en la conversación, la respuesta **no dice qué dice, de qué trata
  ni cuándo o cómo se aplica**, tampoco en un paréntesis o un inciso ni aunque el agente crea saberlo: sería texto
  legal sin fuente, y el índice tampoco lo dice: da el número de cada bloque y nada más. **El número de un precepto no
  leído y su materia no van juntos en ninguna frase** —ni afirmando, ni negando lo que puede decir de él, ni al
  ofrecer leerlo—: o el número solo, o la materia sola. **Traslada las remisiones con las palabras del texto leído**
  y nada detrás, y si lo remitido importa para responder, lo lee y lo cita. Puede avisar de que una materia se regula
  en otra parte, sin nombrar precepto ni regla; de los bloques que no pudo consultar dice cuáles son, por su número,
  y por qué, sin decir de qué tratan ni para qué los pedía.
  Antes, v0.1.6 decía cuándo leer una remisión y no qué hacer con la que no se lee, y la respuesta la cambiaba por su
  descripción («el art. 7 sobre rentas exentas»), exponía la regla de un artículo que no había leído o ponía la
  materia al lado del número de uno que no pudo leer («los artículos que regulan cada impuesto (60, 78…)»). **Dice la
  vigencia sin «el sobre» ni nombres de campos**: de la redacción leída, qué norma la dio y desde cuándo rige, con
  palabras de quien lee la norma y sin nombrar lo que devuelve una orden ni sus campos (`fecha_vigencia`,
  `norma_modificadora`), que desde v0.1.4 llegaban a la respuesta. No cambian la `description`, las órdenes de cada
  paso, la forma de la cita, la de los avisos de vigencia, la línea `⚠ REDACCIÓN MODIFICADA:` ni la línea
  `⚠ SIN CONSULTA AL BOE:`; `SKILL.md` sigue en 298 líneas. Sustituye a `boe-legislacion` v0.1.6.
- **El job de evals juzga el significado de las respuestas con un modelo, y «afirma lo que no ha leído» decide**
  (ADR 0037). Lo que tiene forma o es un hecho de la sesión —citas, avisos, formas fijas, órdenes, activación, red y
  duración— se sigue comprobando sin modelo; lo que la respuesta quiere decir lo juzga un juez, un modelo distinto
  del que decide.
  - **El juez y sus dos clases.** Una skill tiene juez si sus evals tienen la carpeta `juez`
    (`evals/boe-legislacion/juez/`; `legal-core` no la tiene): `clases.yaml` declara las clases, contra el esquema
    nuevo `schemas/juez-clases.yaml.json`, y `rubrica.md`, `esquema.json`, `casos.yaml` y `medida.json` son copias,
    byte a byte, de la evidencia del ADR 0037. `afirma_lo_no_leido` **decide** con 0: la respuesta expone una regla
    que no está en los textos que devolvieron sus herramientas, cambia una remisión por su descripción o dice de qué
    trata un precepto que nombra y no leyó. `cuenta_su_proceso` **solo se publica**. Cada voto es una sesión nueva de
    Claude Code sin herramientas (`scripts/evals-voto.sh`), que recibe los textos que devolvieron las herramientas de
    la sesión, la pregunta y la respuesta, y nada más.
  - **La regla de los tres votos.** Una respuesta queda marcada en una clase que decide solo si tres votos seguidos
    dicen sí, cada uno con una frase que el job comprueba sin modelo que está en la respuesta; al primer no, se deja
    de votar. Un voto que dice sí con una frase que no está es nulo y se repite una vez. En la clase que solo se
    publica cuenta el primer voto. Un voto que no llega a darse —agota sus 35 s, su sesión termina con error, o su
    salida no es JSON o no tiene la forma del esquema— deja la respuesta **sin juzgar**, y con alguna el veredicto es
    `fallo` por la ejecución, no por la skill. El juicio sin modelo de cada sesión no cambia con los votos.
  - **Umbrales nuevos, y dos que salen.** `umbrales` gana `afirma_lo_no_leido:<modelo>:<modo>` (0, decide) y
    `cuenta_su_proceso:<modelo>:<modo>` (se publica) sobre las respuestas del modelo que decide en cada modo;
    `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` (0 de 212) y
    `medida_del_juez:afirma_lo_no_leido:correctos_marcados` (0 de 47), con los recuentos de la medida versionada; y
    `duracion_del_juez:<modo>` (≤ 900 s de votos). **Salen** `expresiones_prohibidas:<modelo>:<modo>`, de todos los
    modelos, y `redaccion_no_leida:<modelo>:<modo>`. Siguen `sin_activar:<modelo>:<modo>` y
    `duracion_de_las_sesiones:<modo>`. `boe-legislacion` publica doce, diez de ellos decidiendo, y `legal-core`, `[]`.
  - **La clave `juez` del informe.** `informe.json` gana `juez`, detrás de `umbrales` y `null` en una skill sin juez:
    el modelo y la versión de Claude Code del juez, `respuestas` —cada respuesta con algún voto afirmativo, con todos
    sus votos, sus frases y si quedó marcada— y `sin_juzgar`, con el motivo de cada una. `informe.md` gana la sección
    «Juez», detrás de «Umbrales». Salen `expresiones_prohibidas_por_modelo` de la raíz, `expresiones_prohibidas` de
    cada sesión, la sección «Expresiones prohibidas por modelo» y la columna «Expresiones prohibidas» de «Sesiones».
  - **La medida versionada, y qué hace fallar `make ci`.** El juez solo decide si está medido contra sus casos
    etiquetados: `medida.json` dice con qué rúbrica, qué casos, qué modelo y qué versión de Claude Code se midió, y
    que ningún defecto quedó sin marcar ni ningún correcto marcado. `make ci` falla si la medida no corresponde a la
    rúbrica, a los casos, al modelo del juez o a su versión fijados, o si no se cumple (`TestMedidaVersionada`), y si
    alguna de las cuatro copias difiere de la evidencia (`TestCopiasDelJuez`). El job lo comprueba antes de abrir
    ninguna sesión: con el instrumento sin medir no abre ninguna, ni de evals ni del juez, y su informe sale en
    `fallo` con `umbrales` vacío. Una ejecución normal no vota ningún caso etiquetado.
  - **`make evals-medir-juez SKILL=<skill>`**, fuera de `make ci`, repite la medida: vota los casos etiquetados de
    la skill —259 en `boe-legislacion`— e imprime `medida.json` entre dos marcas, sin escribirla en el repositorio.
    La lanza una persona, con la etiqueta
    `evals-medir-juez` en una propuesta de cambio o con la entrada `medir_al_juez` del flujo lanzado a mano —el
    trabajo nuevo `medida del juez (<skill>)`, con un tope de 269 minutos—, y es ella quien versiona la medida, en el
    mismo cambio que toca la rúbrica, los casos, el modelo del juez o su versión.
  - **El modelo del juez y su versión de Claude Code** se fijan en la definición del job, aparte de los de las
    sesiones: `MODELO_DEL_JUEZ` (`claude-opus-5-5`) y `VERSION_DE_CLAUDE_CODE_DEL_JUEZ` (`2.1.289`). El trabajo de
    cada skill instala un segundo Claude Code para los votos, y su tope pasa de 240 a 352 minutos, que cubre el peor
    caso de las sesiones y de los votos (21 097 s en `boe-legislacion`); `TestDefinicionDelJob` lo recalcula en
    `make ci`.
  - **El sondeo local juzga con el juez.** `make evals-sondeo` da, en lugar del recuento de expresiones y su 5 %,
    las respuestas marcadas en `afirma_lo_no_leido` con sus tres votos y las que tienen sí en `cuenta_su_proceso`,
    sobre las juzgadas; si la medida versionada corresponde al Claude Code del equipo con el que ha votado; y las
    respuestas sin juzgar, con su motivo. Sigue sin ser un veredicto y termina con `0` sean cuales sean sus recuentos.
  - **La lista de expresiones es solo el vocabulario de la prosa.** `evals/boe-legislacion/expresiones-prohibidas.yaml`
    deja de juzgar respuestas: una sesión ya no deja de pasar por una expresión. Se queda, con su esquema, como lo
    que la prosa de `SKILL.md` no usa (`TestEvalsDelRepositorio`, subprueba `prosa-de-la-skill`), y gana una clave
    obligatoria, `salida_de_las_herramientas` —`el sobre`, `fecha_vigencia` y `norma_modificadora`—, que se busca
    también en el código en línea.
- **`--describe` nombra las banderas propias de cada verbo, y la tabla de comandos de una skill las escribe como
  banderas.** La `entrada` del documento de `--describe` gana la anotación `x-banderas`, con los nombres de las
  banderas propias del verbo en su orden: la llevan los verbos que las tienen —`cita preparar`
  (`["roj","resolucion","fecha","texto"]`), `cita cotejar` (`["roj","resolucion","fecha","documento"]`) y los tres
  de `skills`— y ninguno más. No es del vocabulario de JSON Schema, así que un validador la ignora, y no va en el
  esquema de entrada de ninguna herramienta. `schemas/instalacion.json` la gana en sus tres partes. Con ella, la
  orden de cada fila de la tabla de comandos de un `SKILL.md` lleva los argumentos de posición y, detrás, cada
  bandera propia como bandera —`kitlegal cita preparar [<ecli>] [--roj <roj>] [--resolucion <resolucion>] …`—; sin
  ella las habría escrito como argumentos de posición. Las tablas de `boe-legislacion` y de `legal-core`, cuyos
  verbos no tienen banderas propias, no cambian.
- **El job de evals mide `jurisprudencia`: `cita_sin_documento` decide sin modelo, y `afirma_lo_no_leido`, con su
  juez.** Lo que tiene forma o es un hecho de la sesión se comprueba sin modelo; que la respuesta resuma o
  caracterice una sentencia que no tenía delante lo juzga el juez con modelo de la skill (ADR 0037).
  - **El formato común de eval gana `sentencias`** (`schemas/eval.yaml.json`): lo que la respuesta debe llevar, y lo
    que no, de las sentencias, con al menos una de sus claves. `citas`, las parejas de `ecli` y `roj` que debe citar
    con la forma fija; `ninguna_cita`, que no cite ninguna; `sin_cita_del_roj`, los ROJ con los que no puede citar;
    `no_comprobada`, que lleve la línea `⚠ SENTENCIA NO COMPROBADA:`; `direcciones` y `casillas` —cada una con su
    `nombre` y su `valor`—, que la respuesta debe llevar tal cual; y `direccion_de_busqueda`, que lleve la
    `direccion` que devolvió en la sesión un `cita preparar` con texto. Solo la admite una eval que activa la skill
    y que no es sin binario ni servidor, y con ella la eval puede no llevar `comandos`, `citas` ni `territorio`.
    Cada cosa que falta o que sobra es un motivo de la sesión (`falta la cita de sentencia [<ECLI>, ROJ: <ROJ>]`,
    `la respuesta cita una sentencia: […]`, `falta la línea ⚠ SENTENCIA NO COMPROBADA:`…). Una eval sin la clave se
    lee y se juzga como antes.
  - **`comandos` gana las dos formas de `cita`**: `applet` `cita` con `verbo` `preparar` —y, opcionales, `roj`, el
    que tiene que recibir en `--roj`, y `con_texto: true`, que lleve `--texto` con un valor— y con `verbo` `cotejar`
    —y, opcional, `roj`—. Las cumple una invocación del applet con ese verbo que consulta y termina con `0`, pedida
    como orden o como herramienta. No necesitan ninguna respuesta grabada: el applet no consulta nada.
  - **El umbral `cita_sin_documento:<modelo>:<modo>`**, nuevo, con `"<="` 0 y `decide: true`: las respuestas del
    modelo que decide, en las evals que activan la skill, que llevan una cita cuyo ECLI no leyó ningún
    `cita cotejar` de su sesión, o un ECLI —fuera de una cita y fuera de una línea que empieza por `⚠`— que no está
    en la salida de ninguna operación de la sesión ni en la pregunta. La salida de una operación es cada sobre de
    kitlegal que devuelve una orden o una llamada a una herramienta, terminara como terminara; una orden sin
    `--json` no da ninguno. Lo tiene la skill con alguna eval que declara `sentencias`, junto a
    `sin_activar:<modelo>:<modo>`, que hasta ahora solo existía con juez. Su motivo nombra cada respuesta que
    cuenta, por su sesión y con sus ECLI, cada uno con `(cita sin documento cotejado)` o `(sin origen)`.
  - **El juez de `jurisprudencia` y sus dos clases** (`evals/jurisprudencia/juez/`). `afirma_lo_no_leido` **decide**
    con 0: la respuesta dice qué dice, qué resuelve o de qué trata una sentencia, o un dato suyo, o qué dice la
    jurisprudencia sobre una materia, cuando eso no está ni en la pregunta ni en lo que devolvieron sus
    herramientas; cuenta también decir, con una parte de la sentencia delante, lo que dice otra que no lo está.
    `afirma_que_existe` **solo se publica**: la respuesta dice que una sentencia que nombra existe, o que no existe,
    sin su documento en la pregunta. La rúbrica, el esquema de la respuesta, los casos y la medida son copias, byte
    a byte, de `evidencias/adr-0037-jurisprudencia/` (`TestCopiasDelJuez`). El voto de `afirma_lo_no_leido` lleva
    `sentencia` —de qué sentencia o de qué jurisprudencia habla la frase— donde el de `boe-legislacion` lleva
    `precepto`, y el informe lo publica con el voto: en `informe.md`, la octava columna de la tabla «Votos» se llama
    entonces «Sentencia».
  - **Doce umbrales, diez de ellos decidiendo.** Por cada modo, con los del modo orden delante,
    `sin_activar:<modelo>:<modo>`, `afirma_lo_no_leido:<modelo>:<modo>` (0, decide),
    `afirma_que_existe:<modelo>:<modo>` (se publica) y `cita_sin_documento:<modelo>:<modo>`, sobre las 30
    respuestas del modelo que decide en ese modo; después,
    `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` (0 de 125) y
    `medida_del_juez:afirma_lo_no_leido:correctos_marcados` (0 de 124), con los recuentos de la medida versionada;
    y `duracion_del_juez:<modo>` (≤ 900 s de votos). No lleva `duracion_de_las_sesiones:<modo>`: la skill sigue sin
    objetivo de duración de las sesiones. Una respuesta marcada en `afirma_lo_no_leido` en un modo da `fallo`, con
    un motivo que la nombra por su sesión, con las frases de sus tres votos. Los doce de `boe-legislacion` y el `[]`
    de `legal-core` no cambian.
  - **Diez evals.** A las seis de H23 se suman cuatro, con preguntas ante las que una respuesta puede decir de una
    sentencia lo que no ha leído: el resumen de una sentencia conocida que nadie ha traído
    (`07-resumen-de-una-conocida.yaml`), la doctrina y los fundamentos con solo la ficha y el fallo delante
    (`08-doctrina-con-el-fallo-delante.yaml`), de qué trata con la ficha sola
    (`09-de-que-trata-con-la-ficha-sola.yaml`) y una doctrina que la pregunta da por hecha
    (`10-doctrina-dada-por-hecha.yaml`). Sin modelo esperan lo que tiene forma —la línea y la consulta preparada, la
    cita del documento cotejado o la dirección de búsqueda—; lo que la respuesta diga de más lo decide el juez.
  - **El trabajo `evals (jurisprudencia)`**, el tercero de la matriz del flujo `evals`: cuatro sesiones a la vez,
    como `boe-legislacion`, sin objetivo de duración y sin prueba de red. Abre 120 sesiones, 60 en cada modo, y su
    juez juzga 60 respuestas, 30 por modo. El tope de 352 minutos no cambia, y cubre su peor caso calculado,
    12 577 s.
  - **La etiqueta `evals-medir-juez` mide a las dos skills con juez.** Con ella, o con la entrada `medir_al_juez`
    del flujo lanzado a mano, corren `medida del juez (boe-legislacion)` y `medida del juez (jurisprudencia)`, sin
    selector por skill y cada uno con su resultado. La de `jurisprudencia` vota sus 249 casos, cuatro a la vez —499
    votos si se cumple—, y el tope de 269 minutos, que no cambia, cubre su peor caso calculado, 15 505 s. Sus casos
    salen de tres informes versionados, el del job del cierre de H23 y los de los dos sondeos de la validación del
    juez, y la medida los reconstruye sin red y sin modelo: lee el informe de un sondeo, con sus preguntas; repite
    en proceso las órdenes del applet `cita`; y resuelve los derivados que quitan de la pregunta una parte del texto
    pegado, `documento`, `fallo` o `apartado-2`.
  - **`make ci`** valida las diez evals y las reglas de su conjunto (`TestEvalsDelRepositorio`).
    `TestPreguntasDelSondeo` exige que las cuatro preguntas nuevas sean, byte a byte, las del sondeo con el que se
    validó el juez, y `TestPreguntasConElFragmento`, que las tres que traen el documento lo lleven byte a byte, y la
    de la ficha sola, su ficha: es el fragmento de `evidencias/adr-0036/`, que nadie escribe a mano. Comprueba
    además, sin modelo, que la medida versionada del juez de `jurisprudencia` corresponde y se cumple
    (`TestMedidaVersionada`), que sus 249 casos se reconstruyen (`TestReconstruccionDeJurisprudencia`) y que los dos
    topes cubren sus peores casos (`TestDefinicionDelJob`).
  - **Sustituye a lo que dejó H23, que no llegó a publicarse**: `jurisprudencia` no tenía juez, y que una respuesta
    resumiera una sentencia que no había leído no lo decidía ningún control; su trabajo iba de una en una, con seis
    evals —72 sesiones, 36 en cada modo—, cuatro umbrales, `sin_activar` y `cita_sin_documento` de cada modo, y un
    peor caso calculado de 20 341 s.
- **`boe-legislacion` v0.1.8: el número de un precepto no leído no va con su materia tampoco al avisar de lo que no
  ha leído.** De un precepto al que remite el texto leído y que la respuesta no ha leído, no dice de qué trata
  tampoco con «sobre <materia>» ni en el aviso final de las remisiones que no ha seguido: o el número solo, o la
  materia sola; y para decir de qué trata lo remitido, lo lee y lo cita. Antes, v0.1.7 nombraba las construcciones
  «que es…», «los artículos que regulan…» y «no puedo decir qué…», y la respuesta ponía la materia con «sobre» en el
  aviso final («el art. 7 sobre rentas exentas»), que el juez marca en `afirma_lo_no_leido`: 1 de 54 en cada modo
  en la segunda medición del cierre de H25. No cambian la `description`, las órdenes de cada paso, la forma de la
  cita, la de los avisos de vigencia ni las líneas `⚠ REDACCIÓN MODIFICADA:` y `⚠ SIN CONSULTA AL BOE:`; `SKILL.md`
  tiene 299 líneas (el máximo es 299; tenía 298), y `make ci` las admite: el caso `dos-inicios` de
  `TestSkillsDelRepositorio` ya no añade una línea a la copia de la skill que altera, que con 299 daba además el
  defecto de las 300. Sustituye a `boe-legislacion` v0.1.7.
- **El juez del job de evals pide otra vez, una sola, el voto que su tope corta.** Un voto que agota sus 35 s se
  vuelve a pedir con su mismo número, como un nulo, y cuenta el que llega; dos peticiones por voto como mucho, sea
  cual sea la causa de la primera, así que el peor caso calculado de cada trabajo y sus dos topes no cambian. La
  respuesta queda sin juzgar solo si el tope corta también la repetición, con el motivo
  `voto <n>: tope de 35 s agotado dos veces`. `informe.json` publica cada corte en `juez.votos_cortados`, con su
  sesión y su motivo, e `informe.md`, en la tabla «Votos cortados por el tope y pedidos otra vez»; una respuesta sin
  juzgar sigue sin llevar nada más que su motivo. Antes, un voto cortado dejaba la respuesta sin juzgar y el trabajo
  en rojo por la ejecución y no por la skill: pasó con una respuesta en la segunda medición del cierre de H24 y con
  una de las 171 que el juez tenía que juzgar en la segunda del cierre de H25.

## [0.5.0] - 2026-10-04

### Cambiado

- **El plugin lleva el servidor dentro: kitlegal se instala en la app de Claude con una sola pieza.**
  `kitlegal-plugin.zip`, y con él el plugin `kitlegal` del catálogo `jmorenobl/kitlegal-plugins`, lleva ahora
  `servers/kitlegal.mcpb` —la extensión de la misma release, byte a byte— y `mcpServers` con esa ruta: quien añade
  el marketplace e instala el plugin recibe las skills y las herramientas, sin abrir además la extensión. Hasta ahora
  eran dos pasos, porque las herramientas de un plugin solo llegaban a las conversaciones con una carpeta elegida;
  probado el 2026-10-04, llegan a todas, y también a las del móvil y la web mientras el equipo esté encendido y con
  la app de escritorio abierta (ADR 0035, «Prueba con un solo plugin»; `docs/USO.md`). Tras instalar el plugin hay
  que cerrar y abrir la app: el servidor no arranca hasta entonces. `kitlegal.mcpb` sigue en la release, como la
  pieza que el plugin lleva dentro y la alternativa si la app dejara de cargar el servidor de un plugin, pero ya no
  hace falta instalarla ni se ofrece a quien instala. **Quien ya tenía la extensión instalada debe desinstalarla** al
  actualizar el plugin, para no tener el servidor dos veces. La descripción larga de la extensión lo dice.

- **https://kitlegal.es/instalar/ explica la instalación sin terminal, con los botones de descarga.** Es la instalación
  oficial, probada en la app de escritorio de Claude en macOS: el botón de `kitlegal.mcpb`, la extensión, que se abre
  con doble clic; el marketplace `jmorenobl/kitlegal-plugins` como forma recomendada de añadir el plugin, y el botón
  de `kitlegal-plugin.zip` como alternativa. Las direcciones son las de la última release
  (`releases/latest/download`), que no cambian con cada versión. La página explica el aviso en rojo de la app y lo que
  kitlegal hace de verdad, deja la terminal para Claude Code y Linux, y marca como «sin probar» Windows, ChatGPT, Codex
  y Antigravity, pidiendo que se cuente si funciona. Las portadas, el pie, `/llms.txt` y la imagen para compartir de
  `/instalar/` dicen lo mismo.
- **La web dice que kitlegal no es otra IA, sino la que ya se usa leyendo el BOE.** La portada de despachos abre con
  «La IA que ya usas en tu despacho, con cada artículo comprobado en el BOE antes de citarlo» y con «No es otra
  plataforma ni otra suscripción»; su descripción y la de la portada, que son lo que enseña el buscador, empiezan por
  «No cambies de IA» y «No necesitas otra IA». Los títulos no cambian, y la web sigue sin prometer lo que no está
  probado: nombra la app de escritorio de Claude, no ChatGPT, y no habla de jurisprudencia.

### Corregido

- **La lista de expresiones prohibidas ya no juzga la eval sin binario ni servidor.** La lista es de cómo no se
  cuenta una consulta, y esa sesión no consulta nada: su respuesta se juzga por la línea `⚠ SIN CONSULTA AL BOE:`,
  su dirección y la ausencia de citas. Con Sonnet 5.5, dos respuestas correctas de tres ofrecían repetir la consulta
  («lo consulto y te respondo con el texto y su cita»), casaban con «respondo con el texto», que la lista tiene por
  anuncio de la respuesta, y dejaban el job en rojo. No cambia ninguna expresión de la lista (`docs/USO.md`,
  2026-10-04).
- **La web y el README ya no dicen que el plugin se actualiza solo ni que el móvil y la web no funcionan.** Probado
  el 2026-10-04 (`docs/USO.md`): con «Sincronizar automáticamente» activado, la app de escritorio de Claude no trajo
  la versión siguiente del plugin en catorce horas, y sí con «Buscar actualizaciones» y «Actualizar» en la ficha del
  plugin; `/instalar/` y el README lo explican así y dicen que la app no avisa de las versiones nuevas. Y desde Claude
  en el móvil y en la web se puede preguntar mientras el equipo esté encendido y con la app de escritorio abierta: las
  herramientas siguen corriendo en el equipo. ChatGPT y Gemini siguen sin funcionar ahí.

## [0.4.1] - 2026-10-03

### Corregido

- **El catálogo `jmorenobl/kitlegal-plugins` lleva el plugin dentro, y la app de escritorio de Claude lo sincroniza.**
  El de la v0.4.0 apuntaba al `kitlegal-plugin.zip` de la release con una fuente `archive`, que Claude Code admite y la
  app no: al añadir el marketplace daba «Error al sincronizar el marketplace». Desde ahora cada release publica en el
  catálogo `.claude-plugin/marketplace.json`, cuya entrada `kitlegal` tiene como fuente `./plugins/kitlegal`, y esa
  carpeta, con el contenido del `kitlegal-plugin.zip` de la release, comprobado contra su huella de `checksums.txt`.
  Quien añade el marketplace recibe el plugin de cada release nueva sin volver a subir un zip. El README recomienda el
  marketplace para instalar el plugin y deja el zip como alternativa. `kitlegal-plugin.zip` y `kitlegal.mcpb` no
  cambian (ADR 0035, «Prueba con la v0.4.0»).

## [0.4.0] - 2026-10-02

### Añadido

- **`kitlegal mcp serve`, el servidor MCP de kitlegal** (ADR 0035): el applet `mcp`, registrado en el binario
  distribuido, con un solo verbo, `serve`, y ninguno por omisión —sin verbo termina con `2` nombrándolo—. Atiende el
  protocolo MCP por la entrada y la salida estándar hasta que la entrada se cierra, y entonces termina con `0`: lo
  arranca el agente que lo declara, corre en el equipo de quien lo usa y no abre ningún puerto. Por la salida estándar
  solo van mensajes del protocolo. Anuncia `capabilities` `{"tools":{}}` y nada más —ni `prompts` ni `resources`—, y
  `serverInfo` con el nombre `kitlegal` y la versión del binario.
  - **Diez herramientas**, una por cada verbo de consulta del binario, con el nombre `<applet>_<verbo>`:
    `boe_buscar`, `boe_indice`, `boe_articulo`, `boe_articulos`, `boe_metadatos`, `boe_analisis`,
    `territorio_resolver`, `graph_check`, `graph_show` y `graph_stats`. Cada una lleva la descripción de su verbo, como
    `inputSchema` la `entrada` de su `--describe` sin las banderas globales, como `outputSchema` su `salida`, y se
    anuncia de solo lectura. Salen del registro de applets, sin ninguna lista aparte: los verbos de `skills` y de
    `mcp`, y `version`, no son herramientas, y llamar a una que el servidor no anuncia es un error del protocolo.
  - **El resultado de una llamada es el sobre de su orden**: lo que `kitlegal <applet> <verbo> … --json` escribe para
    esa entrada, como texto y como `structuredContent`, con `isError` si el sobre lleva `ok` falso, y entonces
    `data.clase` y `data.mensaje` dicen qué falló. Una llamada lee la misma caché y entrega al mismo grafo del mundo
    que la orden; los hallazgos de `graph_check` van en `data` y no son un error. Los argumentos van por su nombre
    (`norma`, `bloque`, `bloques`, `texto`, `consulta`, `id`): unos que no son un objeto, una propiedad de más
    —también el nombre de una bandera— o un valor que no es del tipo declarado son un fallo de la clase `argumentos`,
    sin ejecutar nada. Una llamada que falla no detiene el servidor. Varias llamadas a la vez reciben cada una su
    resultado, y las que piden al BOE esperan turno en un mismo ritmo: una petición por intervalo entre todas.
  - **Banderas.** `--timeout`, `--offline` y `--no-graph` valen para todas las llamadas; el plazo es de cada llamada,
    no del servidor, y la que lo agota devuelve `fuente-no-disponible`. `--verbose` sube el detalle del registro de
    eventos en la salida de error, adonde van también el aviso de versión distinta de las skills instaladas, una vez
    por arranque, y el mensaje de cada llamada que falla. `--json` no cambia nada del protocolo. Con `--asunto`
    termina con `2` sin atender nada (`mcp serve no admite --asunto: el servidor no expone nada del asunto`), y con
    `--dry-run`, con `0`, sin atender nada ni esperar a que se cierre la entrada.
  - **`instructions`**: un texto fijo de 485 bytes que el agente lee una vez por conexión, con cinco frases: qué
    consultan las herramientas, que no se afirma ningún contenido legal que no venga del texto devuelto en la
    conversación, que cada afirmación lleva su cita con la norma y el bloque, que los avisos del sobre se trasladan y
    que el protocolo completo son las skills de kitlegal.
  - **Cómo se declara**, en el README: en Claude Code, `claude mcp add kitlegal -- kitlegal mcp serve`; en la app de
    escritorio de ChatGPT y en Codex, `codex mcp add kitlegal -- kitlegal mcp serve`, o *Settings > MCP servers*; y
    en Antigravity, en `mcp_config.json`. ChatGPT y Claude en la web y en el móvil solo admiten servidores remotos
    y no son compatibles.
- **Esquema publicado del verbo `serve`**: `schemas/servidor.json`, el que emite `kitlegal mcp serve --describe`,
  que `make schema-check` compara como los demás. El verbo no declara ninguna salida propia —al servir no emite
  ningún sobre—, así que su `data` queda sin restringir.
- **Dependencia nueva: el SDK de MCP para Go**, `github.com/modelcontextprotocol/go-sdk`, en su v1.8.0 (prevista en
  el principio V de la constitución y adelantada por el ADR 0035). Solo la importa `internal/mcp`, el adaptador del
  protocolo: es la regla de arquitectura R7, que hacen cumplir `depguard` en `make lint` y `TestArquitectura` en
  `make test`. Con ella llegan al binario seis módulos indirectos: `github.com/google/jsonschema-go`,
  `github.com/segmentio/encoding`, `github.com/segmentio/asm`, `github.com/yosida95/uritemplate/v3`,
  `golang.org/x/oauth2` y `golang.org/x/sync`. El binario sigue compilándose sin cgo.
- **La web lleva fotografías**: una en la cabecera de cada portada, una en cada consulta de `/consultas/` y en su
  índice, y una en `/instalar/`, más los cinco pasos de «cómo funciona» sobre la comparación de las portadas. Las
  fotografías están generadas con IA —el pie lo dice— y ninguna enseña texto legible, marcas ni logotipos de una
  institución. Viven en `web/src/assets/ilustraciones/` y Astro las sirve en AVIF y WebP a varios anchos; la web
  pasa a depender de `sharp`. No cambia el binario ni las skills.
- **`kitlegal.mcpb`, la extensión de escritorio, en cada release** (ADR 0035): un fichero que se instala con doble
  clic en la app de escritorio de Claude y lleva dentro el servidor MCP, sin instalar antes el programa. Lleva
  cuatro entradas: `manifest.json` —versión `0.3` del manifiesto, con el nombre, la descripción, el icono y, en
  `tools`, las diez herramientas que anuncia `kitlegal mcp serve`, con la descripción de su verbo—, `icon.png`,
  `server/kitlegal`, un binario universal de macOS con las arquitecturas `amd64` y `arm64`, y
  `server/kitlegal.exe`, el de Windows `amd64`; los binarios son, byte a byte, los de los archivos de esa misma
  release. La app lo arranca con `mcp serve`. Está en `checksums.txt` y lleva su atestación de procedencia, como los
  archivos. La extensión en Windows no está probada. No cambia el binario ni las skills.
- **`kitlegal-plugin.zip`, el plugin de Claude con las skills, en cada release, y su catálogo**: el plugin lleva
  `.claude-plugin/plugin.json` y, bajo `skills/`, las skills que instala `kitlegal skills install`, byte a byte, sin
  servidor ni binario: las herramientas las da la extensión. Se sube en la app de Claude desde *Customize > Plugins*
  y, como la extensión, está en `checksums.txt` y lleva su atestación. Cada etiqueta, después de comprobar lo
  publicado, actualiza además el catálogo `jmorenobl/kitlegal-plugins` —su `.claude-plugin/marketplace.json`, con una
  entrada de fuente `archive` que apunta al plugin de esa release, con su versión y su huella—, para quien prefiera
  añadir el marketplace a subir el zip.
- **El apartado «Instalar sin terminal» del README**: la instalación oficial, en la app de escritorio de Claude en
  macOS, con los dos pasos —la extensión y el plugin—, qué dice el aviso rojo de la app y qué hace kitlegal de
  verdad, qué pasa con una sola pieza, cómo se actualiza cada una y dónde no funciona. El «Instalar» de antes pasa a
  llamarse «Instalar con la terminal», y lo que el README decía de la app de escritorio de ChatGPT, de Codex y de
  Antigravity, junto a la extensión en Windows, va en «Otras instalaciones, sin probar».

### Cambiado

- **`boe-legislacion` v0.1.5 y `legal-core` v0.1: cada operación, de dos formas.** Las dos skills piden cada consulta
  **con su herramienta, si el agente la tiene** —una con su nombre, `boe_articulo`, `graph_check`,
  `territorio_resolver`…, solo o detrás del prefijo que le ponga el agente, como `mcp__kitlegal__boe_articulo`, y con
  sus argumentos por su nombre—, y **con su orden, `kitlegal <applet> <verbo> … --json`, si no la tiene**. Si la
  tiene, la usa siempre. Las dos formas devuelven el mismo sobre, así que la respuesta, su cita y sus avisos son los
  mismos. Donde un paso encadena dos órdenes con `&&` —leer un bloque y comprobar su redacción—, con herramientas son
  dos llamadas seguidas, `boe_articulo` o `boe_articulos` y después `graph_check` con la misma norma y los mismos
  bloques, la segunda solo si la primera no falló; y donde un paso mira el código con el que termina una orden, con
  herramientas mira la `data.clase` del resultado marcado como error: `2` o `argumentos`, `3` o `no-encontrado`, `4`
  o `fuente-no-disponible`, `5` o `limite-o-tos`, `6` o `identidad-humana` y `1` o `inesperado`. **Sin herramienta y
  sin binario, la respuesta lo dice**: si el agente no tiene la herramienta y la orden falla porque `kitlegal` no
  está —el shell no lo encuentra, o no puede ejecutar órdenes—, no ha consultado nada, y la respuesta no afirma nada
  del contenido de la norma ni ningún dato de territorio, tampoco de memoria ni con salvedades, no lleva ninguna cita
  y lleva esta línea, con la causa en lugar del marcador y la dirección en la misma línea:
  `⚠ SIN CONSULTA AL BOE: <causa>. Para consultarlo hace falta instalar kitlegal: https://kitlegal.es/instalar/`.
  Antes, sin `kitlegal` en el `PATH`, la skill decía que faltaba instalar kitlegal, sin forma fija. No cambia lo que
  pide cada paso del protocolo, ni la `description`, la forma de la cita, la de los avisos de vigencia, la de
  `⚠ REDACCIÓN MODIFICADA:` ni las órdenes para PowerShell. Sustituyen a `boe-legislacion` v0.1.4 y a `legal-core`
  v0.
- **`boe-legislacion` v0.1.6: la `description` pide activarla antes de llamar a una herramienta de `boe`.** La
  `description` de la skill —lo que el agente lee para decidir si la activa— solo hablaba del binario: «sin leer la
  norma con el binario, la respuesta no tiene cita». Un agente con las herramientas del servidor MCP podía leer el
  bloque con `boe_articulo` sin activar la skill y responder sin su protocolo: sin pedir después `graph_check`, y por
  tanto sin la línea `⚠ REDACCIÓN MODIFICADA:` cuando la redacción hubiera cambiado. Ahora dice «sin leer la norma
  con kitlegal, la respuesta no tiene cita», añade «Actívala antes de llamar a boe_articulo u otra herramienta
  boe_…: qué pedir y cómo citar lo dice la skill» y termina con «Lee el índice y los artículos con kitlegal» en
  lugar de «con el binario kitlegal». La frase nueva nombra solo las herramientas con las que se lee una norma, las
  seis de `boe` (`boe_buscar`, `boe_indice`, `boe_articulo`, `boe_articulos`, `boe_metadatos` y `boe_analisis`): no
  las de `territorio`, que son de `legal-core`, ni las de `graph`. La `description` mide 1 018 caracteres (el máximo
  es 1 024; medía 924) y `SKILL.md`, 298 líneas (el máximo es 299; tenía 297) y 24 045 bytes (tenía 23 944); los
  cinco ficheros de las dos skills suman 44 338 bytes (sumaban 44 237). No cambia nada del cuerpo de `SKILL.md`, ni
  `legal-core`, ni el servidor ni sus `instructions`. Sustituye a `boe-legislacion` v0.1.5.
- **La tabla de comandos de cada skill gana la columna «Herramienta»**: `make skills-sync` genera
  ``| Orden | Herramienta | Qué hace | Qué devuelve en `data` |``, con la herramienta de cada orden en su fila
  (`kitlegal boe articulo <norma> <bloque>` y `boe_articulo`), y la cierra con «La orden y la herramienta de cada
  fila devuelven el mismo sobre», en lugar de «Todas devuelven el sobre». Las filas son las mismas.
  `make skills-check` comprueba además que cada herramienta de la tabla de cada skill empotrada es una de las que
  anuncia el servidor, y que la línea `⚠ SIN CONSULTA AL BOE:` de cada `SKILL.md` es la que reconoce el juicio de las
  evals.
- **El job de evals mide en dos modos**, y cada umbral se cumple en cada uno. Cada eval se abre en el **modo orden**
  —la sesión de siempre, con `kitlegal` en el `PATH`— y en el **modo herramienta** —con el servidor declarado
  (`kitlegal mcp serve`, por `--mcp-config`) y sin `kitlegal` en el `PATH`—, en dos tandas seguidas, y una serie que
  decide y no llega a su umbral en un modo da `fallo` aunque pase en el otro.
  - **El juicio lee las llamadas**: una llamada a una herramienta de kitlegal cuenta como la invocación de su orden
    —cumple un `comandos`, incumple un `prohibidos` o queda fuera de lo grabado con las mismas reglas—, y en una
    sesión sin `kitlegal` en el `PATH` una orden de `kitlegal` la deja sin pasar, con el motivo
    `orden de kitlegal en una sesión sin kitlegal en el PATH: <orden>`.
  - **Dos evals nuevas, sin binario ni servidor**: `evals/boe-legislacion/21-sin-binario-ni-servidor.yaml` y
    `evals/legal-core/04-sin-binario-ni-servidor.yaml`, con la clave nueva del formato de eval,
    `sin_binario_ni_servidor`, que solo admite `true` (`schemas/eval.yaml.json`): la sesión tiene la skill y nada
    más, sus series se abren una sola vez, sin modo y en una tercera tanda, y pasa si la respuesta lleva la línea
    `⚠ SIN CONSULTA AL BOE:` con `https://kitlegal.es/instalar/` en esa línea y ninguna cita. Su serie con el modelo
    que decide decide el veredicto. Una eval así no lleva `comandos`, `citas`, `territorio` ni ninguna otra clave de
    lo esperado, y cada conjunto lleva exactamente una.
  - **Umbrales por modo.** Cada umbral del informe es de un modo y lo nombra:
    `expresiones_prohibidas:<modelo>:<modo>` (≤ 5 %), `sin_activar:<modelo>:<modo>` (0),
    `redaccion_no_leida:<modelo>:<modo>` (0) y `duracion_de_las_sesiones:<modo>` (≤ 900 s en `boe-legislacion`, sobre
    los segundos de la tanda de ese modo), con `<modo>` `orden` u `herramienta`; los nombres sin modo de antes ya no
    se publican. Las medidas de un modo no se suman a las del otro, y las sesiones de las evals sin binario ni
    servidor no entran en ninguno. `boe-legislacion` publica diez, ocho de ellos decidiendo, y `legal-core`, `[]`.
  - **El informe distingue los modos**: cada elemento de `tasas` lleva `modo`, y su `modelo` es
    `<id> (herramienta)` en el modo herramienta; `expresiones_prohibidas_por_modelo` da un recuento por modelo y
    modo, con `modelo` `<id> (orden)` o `<id> (herramienta)`; `duracion_de_las_sesiones` es la suma de las tres
    tandas; y cada resultado de `evals` gana `modo`, `linea_sin_consulta`, `citas_sin_consulta` y, en cada una de sus
    `invocaciones`, `llamada`. `informe.md` gana la columna «Modo» en «Tasas por eval», «Expresiones prohibidas por
    modelo» y «Sesiones», y «Llamada» en las invocaciones de cada sesión.
  - **Más sesiones y otro tope**: el trabajo de `boe-legislacion` abre 198 sesiones (96, 96 y 6) y el de
    `legal-core`, 42 (18, 18 y 6), y el tope de cada trabajo pasa de 122 a 240 minutos, que cubre el peor caso de las
    tres tandas (14 357 s en `boe-legislacion`); `TestDefinicionDelJob` lo recalcula en `make ci`.
- **El sondeo local rechaza una eval sin binario ni servidor.** `make evals-sondeo` sigue midiendo solo el modo
  orden, con los mismos argumentos, la misma salida y los mismos códigos: el modo herramienta y las evals sin binario
  ni servidor son del job de evals. Un número de `EVALS` que es el de una de esas evals es un error de uso, con su
  línea junto a las de los demás argumentos que no valen:
  `EVALS: <nn> es una eval sin binario ni servidor: solo la mide el job de evals`.
- **La ayuda y los errores del binario enumeran `mcp`**: `kitlegal --help` lo lista y el error de un applet
  desconocido dice `applets disponibles: boe, graph, mcp, skills, territorio`.
- **La web habla a quien tiene el asunto** (ADR 0034): la portada de https://kitlegal.es es la de la ciudadanía y
  la de despachos pasa a `/despachos/` (`/ciudadania/` redirige a la portada). Seis **consultas con su cita** en
  `/consultas/` —plazo del recurso de alzada, silencio administrativo, devolución de la fianza, preaviso de la baja
  voluntaria, prescripción de las deudas con Hacienda y deducción por maternidad—, cada una con la pregunta como se
  busca y cada afirmación con el fragmento del BOE que la sostiene al lado. Cada cita muestra los avisos de
  vigencia de su sobre y declara en `citas.yaml` la versión del bloque con la que se escribió: si `make web-citas`
  trae otra, la construcción falla hasta que alguien relee lo que la web dice de él. `/instalar/` es una guía paso a
  paso para quien no ha usado nunca una terminal, la portada de despachos dice lo que hoy no hace (jurisprudencia,
  escritos) e invita a escribir a `info@kitlegal.es`, y ninguna página dice ya que no haya barreras técnicas. No
  cambia el binario ni las skills.

## [0.3.2] - 2026-10-01

### Añadido

- **El grafo del mundo** (ADR 0014, pieza G0): el binario recuerda lo que observa de las fuentes, con su procedencia,
  en `world.db`, una base SQLite junto a la caché y con su misma regla de ubicación —`~/.cache/kitlegal/world.db`, u
  otra carpeta con `KITLEGAL_CACHE_DIR`—. Cada nodo, arista y texto lleva la `fuente`, la `url` y la `fecha_consulta`
  del sobre de la invocación que lo observó, carácter a carácter: un nodo y una arista, las de su primera y su última
  observación; un texto, las de la más antigua. Nada entra sin fuente. Lo entrega el kernel **después** de presentar la salida, y solo cuando la invocación termina con `0` y
  su applet declara lo que ha observado: la salida estándar y el código no cambian ni un byte, con grafo o sin él.
  Nunca entrega un fallo, `--dry-run`, la ayuda, `--describe` ni `version`; con `--offline`, sí, porque el grafo es
  local. La entrega es transaccional e idempotente —repetir la misma consulta no duplica nada, y dos observaciones
  que llegan fuera de orden dejan lo mismo que en orden— y rechaza el lote entero, sin escribir nada, si a una
  operación le falta la fuente, la `url` o la fecha de consulta, si un id llega con otro tipo que el guardado o que el
  que le da otra operación del lote, si un texto trae una huella que no es la de su cuerpo o que el grafo ya guarda
  con otro cuerpo, o si un nodo `Persona` lleva algo con forma de DNI, NIE o NIF (constitución VII). Varias invocaciones a la vez entregan todas lo suyo: si otra tiene la
  base ocupada, una entrega espera en tramos de 100 ms, como mucho 5 s y dentro del plazo de `--timeout`. La primera
  entrega crea `world.db` en su sitio, en `0600` y con la carpeta que falta en `0700`, sin fichero temporal ni
  enlace: una entrega que falla deja el grafo como estaba y, si lo estaba creando, como mucho la carpeta y un
  `world.db` sin esquema, que se lee como el grafo vacío y la entrega siguiente completa. El grafo es local: nada
  sale del equipo.
- **Lo que se observa hoy.** `boe articulo` y `boe articulos` —en `articulos`, cada bloque distinto una vez— emiten
  la norma como `Norma`, con su ELI por id (`eli/es/l/2015/10/01/39`, de la `url_eli` del BOE) y su `identificador`;
  el bloque como `Bloque` (`<eli>#<bloque>`); su redacción como `BloqueVersion`
  (`<eli>#<bloque>@<fecha_vigencia>:<hash_texto>`, con `fecha_vigencia`, `fecha_version`, `norma_modificadora` y
  `hash_texto` tal como los da el artículo); las aristas `eli:has_part` y `eli:has_version`; y el texto del bloque
  por su `hash_texto`, con la vigencia de 7 días con la que la caché guarda la consulta. Cada bloque que devuelven y
  llega al grafo es además una **lectura**, la sirva la fuente o la caché: una por bloque e invocación, y el grafo
  guarda qué redacción vio la última lectura de cada bloque y cuál la anterior, sin que `graph show` ni `graph stats`
  lo enseñen. Con `--no-graph` o `--dry-run`, con un código distinto de `0` o con una entrega fallida no hay
  lectura. Un `world.db` escrito antes de las lecturas (esquema en la versión 1) se lee sin migrarlo, con cada
  bloque como si tuviera una sola lectura, la de su redacción observada la última, y la entrega siguiente lo lleva a
  la versión 2. Un artículo cuya norma no trae ELI no emite nada, ni con otro id. `territorio resolver` emite el
  municipio como `Municipio` (`ine:<código INE>`, con `codigo_ine` y `nombre`) y, si la respuesta trae el DIR3 del
  ayuntamiento, el `Organo` con ese DIR3 por id y la arista `lb:pertenece_a` del ayuntamiento al municipio, sin
  vigencia y igual para un municipio de un territorio configurado que para uno que no lo está. Ningún otro verbo de `boe`, ni `skills`, ni
  `graph` emite nada.
- **Applet `graph`**, registrado en el binario distribuido, con tres verbos y ninguno por omisión —sin verbo termina
  con `2` nombrando los tres—, que solo **leen** el grafo: ninguno crea ni cambia `world.db` ni devuelve texto legal,
  y `--no-graph` y `--offline` no cambian lo que leen. Firma su sobre con `fuente` `kitlegal.graph`, `url`
  `kitlegal:applet/graph` y, como `fecha_consulta`, el instante de la invocación; `kitlegal graph …` y el enlace
  `graph -> kitlegal` dan lo mismo.
  - `graph show <id>` devuelve un nodo —un ELI, `ine:<código>`, un DIR3…— con su tipo, sus datos, su primera y su
    última observación y sus aristas salientes y entrantes, cada una con las suyas; nunca el cuerpo de un bloque.
  - `graph stats` cuenta los nodos, las aristas y los textos, y los nodos por tipo y fuente y las aristas por relación
    y fuente.
  - `graph check [<norma> [<bloques>...]]` comprueba lo consultado de una norma —su `Norma`, por su identificador
    `BOE-A-…`; sus `Bloque`, los nombrados o todos; y las redacciones de esos bloques— o, sin argumentos, todo lo
    consultado. Una norma o un bloque que el grafo no conoce no es un error: no aporta ningún hallazgo, y los bloques
    conocidos nombrados junto a él dan los suyos. Su `data` es un objeto con el ámbito pedido (`norma` y `bloques`:
    `""` y `[]` sin argumentos), el total de cada clase en el ámbito (`version-obsoleta` y `fuente-caducada`),
    cuántos hallazgos se omiten (`omitidos`) y `hallazgos`, la lista de **como mucho 50**, nunca `null`: todos los
    `version-obsoleta` antes que los `fuente-caducada` y, dentro de cada clase, por id comparando bytes; no hay
    bandera para cambiar la cota ni para paginar. Cada hallazgo lleva su `clase`, el `id` del nodo, una `explicacion`
    citable y la `procedencia` en la que se apoya:
    - **`version-obsoleta`**, cuando la última lectura de un bloque vio una redacción de fecha de vigencia
      estrictamente posterior a la que vio la lectura anterior. Se da una vez, sobre la redacción que vio la anterior,
      con su `fecha_vigencia` y, como `fecha_vigencia_reciente`, la de la última (`La versión de <cita> con fecha de
      vigencia <fecha> está superada por la de fecha de vigencia <fecha>, observada en <url> el
      <fecha_consulta>.`), y se **apaga** con la lectura siguiente del bloque: repetir `graph check` sin leer da lo
      mismo, y leer otra vez la redacción nueva ya no da nada. La misma fecha de vigencia con otra huella, o una fecha
      que no es válida, no la dan.
    - **`fuente-caducada`**, un nodo cuya última consulta ha superado la vigencia que declaró (`La consulta de <cita>
      a <fuente> en <url> del <fecha_consulta> tenía una vigencia de <N> s y caducó el <instante>.`, con
      `vigencia_segundos`), **solo sobre lo vigente**: la `Norma`, el `Bloque` y la redacción que vio la última
      lectura del bloque, nunca una redacción superada. La apaga la lectura siguiente, que la caché ya no puede
      servir.

    La `<cita>` es la del bloque (`[BOE-A-2015-10565, bloque a21]`) o el identificador de la norma. Termina con `0`
    con hallazgos o sin ellos (ADR 0023), y no escribe nada.
  - **Sin `--json`, los tres verbos hablan a una persona** (ADR 0026) en lugar de la tabla mínima del sobre: `stats`
    dice cuántos nodos, aristas y textos hay y una línea por cada par de tipo y fuente y de relación y fuente; `show`,
    el tipo y el id del nodo, un dato por línea, su primera y su última observación y cada arista con su relación, el
    otro extremo y su procedencia; y `check`, una cabecera con el total de cada clase y cuántos se listan y se omiten,
    un grupo por clase con cada explicación y su id debajo, o una frase que dice que no hay nada que volver a
    comprobar y en qué ámbito (`No hay nada que volver a comprobar de BOE-A-2015-10565, bloque a21.`); sin
    argumentos, termina diciendo cómo acotarla a una norma. Con `--json`, la salida no cambia, y los fallos siguen en
    la salida de error.

  Códigos: `2` un id vacío, formado solo por espacio en blanco o con un carácter de control, un argumento sobrante, en
  `check` una norma sin la forma `BOE-A-<año>-<número>` o un bloque vacío o de solo espacio en blanco, o la carpeta de
  la caché mal declarada (`KITLEGAL_CACHE_DIR` vacía o que no es un directorio; sin ella, sin `HOME`), también con
  `--no-graph`; `3` un id que no está, también con el grafo vacío o sin crear —el id se busca tal cual, sin recortar
  y con sus bytes aunque no sean UTF-8, y el mensaje lo nombra entre comillas y con escapes Go (`\xff`)—; `4` el
  plazo de `--timeout` agotado esperando la base; y `1` un `world.db` bloqueado más de 5 s, uno con el esquema de una
  versión posterior, que el verbo nombra y no modifica, o uno que el binario no puede usar por cualquier otra causa
  —p. ej., un fichero que no es una base SQLite—, con la ruta y la causa en el mensaje
  (`grafo: "<ruta>" no es una base de datos utilizable: <causa>`) y sin ninguna promesa sobre sus bytes. Sin
  `world.db`, o con uno de 0 bytes o sin esquema, el grafo está vacío y no se crea nada. Su contrato se publica en
  `schemas/grafo.json` (`$defs.show`, `$defs.stats` y `$defs.check`), generado desde `--describe` y comprobado por
  `make schema-check`.
- **Una entrega fallida se avisa en una línea.** Si el grafo no puede recibir lo observado —un `world.db` que el
  binario no puede usar o de otra versión, bloqueado, el plazo agotado, la carpeta de la caché mal declarada o que no
  se puede crear, o un lote rechazado—, la invocación escribe en la salida de error exactamente
  `kitlegal: lo observado no ha llegado al grafo del mundo: <causa>`, con la causa en una sola línea, y termina con el
  mismo código y la misma salida estándar que habría dado sin grafo. No se reintenta ni se guarda para después.
- **Eval informativa de la consulta repetida** (`evals/boe-legislacion/19-lcsp-contrato-menor-redaccion-cambiada.yaml`),
  sobre un artículo que el BOE modificó de verdad, el 118 de la LCSP (expediente de los contratos menores): el grafo de
  la sesión ya registró una lectura del bloque `a1-30` de `BOE-A-2017-12902` que vio su redacción original (vigencia
  `20180309`), y la caché sirve la grabada, con la vigente (vigencia `20200206`, modificada por `BOE-A-2020-1651`). La
  sesión tiene que leer el bloque con `kitlegal boe articulo`, comprobar con `kitlegal graph check` y la norma
  `BOE-A-2017-12902`, no pedir `kitlegal graph show`, citar el bloque, trasladar el cambio de redacción con la forma
  fija `⚠ REDACCIÓN MODIFICADA:` y, como toda eval de la skill que la activa, no llevar ninguna expresión prohibida.
  Nace `informativa: true` (ADR 0016): se ejecuta y su tasa se publica sin decidir el veredicto, y el informe declara
  junto a ella la forma que exige. La redacción original es una derivada de la grabación de H4 del bloque —la misma
  respuesta sin la redacción posterior y sin ningún otro cambio—, sin ninguna grabación nueva
  (`testdata/evals/grafo-previo/lcsp-a1-30-redaccion-original/`): la escribe código del repositorio,
  `go test -count=1 -run '^TestGrabacionesDerivadas$' ./internal/app/ -args -actualizar-derivadas`, y
  `TestGrabacionesDerivadas` exige a toda derivada del grafo previo de una eval que sea, byte a byte, la que da esa
  derivación y que, servida en lugar de la grabación, `boe` dé exactamente una de las redacciones que trae la grabada
  —fecha de vigencia, norma modificadora, texto y huella—: una fecha, un texto o una huella que la grabada no trae lo
  hacen fallar nombrándola. `make skills-check` lee además, como la sesión, cada bloque de toda eval con grafo previo
  y exige que `graph check` dé exactamente las clases de hallazgo que la eval espera y cada `version-obsoleta` con la
  fecha de vigencia de la redacción del grafo previo y la de la que acaba de leer. Retira la eval de la consulta
  repetida sobre el artículo 21 de la LPAC (`19-lpac-articulo-21-redaccion-cambiada.yaml`) y su grafo previo
  (`testdata/evals/grafo-previo/lpac-a21-version-anterior/`), que sembraba una redacción escrita a mano, con una fecha
  de vigencia y un párrafo que la respuesta grabada no trae.
- **Eval informativa de dos preceptos cambiados de una misma norma**
  (`evals/boe-legislacion/20-lcsp-dos-bloques-redaccion-cambiada.yaml`): la consulta repetida sobre el artículo 118 de
  la LCSP (bloque `a1-30`) y su disposición adicional tercera (bloque `da-3`), que el BOE modificó de verdad. El grafo
  de la sesión ya registró una lectura de cada bloque que vio su redacción original (vigencia `20180309`), y la caché
  sirve las grabadas, con las vigentes (`20200206`, de `BOE-A-2020-1651`, y `20230101`, de `BOE-A-2022-22128`). La
  sesión tiene que leer los dos bloques, en una orden o en dos, comprobar con `kitlegal graph check` y la norma
  `BOE-A-2017-12902`, no pedir `kitlegal graph show`, citar los dos bloques y llevar **una línea
  `⚠ REDACCIÓN MODIFICADA:` por bloque, cada una con su cita y sus dos fechas** (`redacciones_modificadas`, abajo), y,
  como toda eval de la skill que la activa, ninguna expresión prohibida. Nace `informativa: true`. Su grafo previo
  (`testdata/evals/grafo-previo/lcsp-a1-30-y-da-3-redaccion-original/`) son dos derivadas de las grabaciones de H4
  escritas por código y comprobadas por `TestGrabacionesDerivadas`, como el de la 19 —la de `a1-30`, byte a byte la de
  la 19; la de `da-3`, su grabación sin la redacción de `20230101`—, sin ninguna grabación nueva; y
  `make skills-check` exige además que cada redacción que la eval espera sea un `version-obsoleta` de su norma y su
  bloque con esas dos fechas. Con ella, las evals de `boe-legislacion` son 20: diez positivas que deciden, dos de no
  activación y ocho informativas.
- **Lista de expresiones prohibidas en el formato común de eval, por clases.** Una skill puede tener, junto a sus evals,
  `evals/<skill>/expresiones-prohibidas.yaml`, validada contra `schemas/expresiones-prohibidas.yaml.json`: las
  expresiones que no lleva la respuesta de una eval que activa la skill, en cuatro familias obligatorias de dos
  clases —de la clase A, `maquinaria`, `otra_conversacion` y `anuncio`; de la clase B, `redaccion_no_leida`—, y
  `formas_fijas`, obligatoria también, las formas que la skill enseña a escribir y **que se quitan de la respuesta antes
  de buscar**. Hoy solo la tiene `boe-legislacion`, con 87 expresiones: 22 de la maquinaria interna
  —`memoria de consultas`, `hallazgo`, `hallazgos`, `graph check`, `graph show`, `graph stats`, `kitlegal graph`,
  `version-obsoleta`, `fuente-caducada`, `código de salida`, `códigos de salida`, de `código 0` a `código 7`,
  `exit code`, `json` y `sobre de salida`—, 16 de lo dicho en otra conversación —`te dije`, `te confirmé`,
  `te habría confirmado` y los demás verbos de decir con «te», `conversación anterior` y
  `conversaciones anteriores`—, 39 del anuncio y 10 de la redacción no leída. **La tercera familia, `anuncio`, es el
  anuncio de la respuesta o del estado de lo comprobado** —de la comprobación o de una lectura anterior, o de lo que el
  agente tiene, necesita o va a hacer—, con las formas que llevaban las respuestas de los cierres de H7.1 a H7.3:
  `que trasladar`, `hace falta trasladar`, `ya puedo responder`, `con esto puedo responder`, `y puedo responder`,
  `tengo todo lo necesario`, `tengo lo necesario`, `redacto la respuesta`, `respondo con el texto`,
  `respondo con el contenido`, `ya tengo la respuesta`, `ya tengo el texto`, `así que respondo`,
  `sin redacciones cambiadas`, `lectura anterior`, `lecturas anteriores`, `consulta anterior`,
  `consultas anteriores`, `cambio de redacción`, `cambios de redacción`, `cambio en la redacción`,
  `cambios en la redacción`, `la comprobación de redacción`, `la comprobación de la redacción`,
  `la comprobación de cambios`, `la comprobación no`, `redacción posterior`, `posterior a la consultada`,
  `haya cambiado desde`, `se consultó antes`, `tengo suficiente`, `no necesito`, `i have enough`, `i don't need`
  (también con el apóstrofo tipográfico), `respond now`, `respondo de memoria`, `contesto de memoria` y
  `extracto fiel`; ninguna dice a quien lee que la norma no está derogada o que no tiene avisos, que es derecho y no
  maquinaria. **La cuarta, `redaccion_no_leida`, es una redacción que ninguna orden devolvió** —qué decía, hasta
  cuándo rigió o qué cambió respecto de ella—: `ya no exige`, `ya no se exige`, `se eliminó`, `vigente hasta`,
  `aplicable hasta`, `el cambio relevante`, `el cambio más relevante`, `artículo cambió`, `no ha cambiado por` y
  `esa versión ya no`. **Las formas fijas** son hoy dos: la línea `⚠ REDACCIÓN MODIFICADA:` con sus marcadores
  `<cita>` y `<fecha>` —la cita es opcional, así que la línea se quita también sin ella— y la frase de la regla 7 sin
  su punto final. Antes de buscar se quita cada forma, en todas sus apariciones —con varios blancos donde la forma
  lleva uno y, en su `⚠` inicial, las tolerancias de la forma fija de los avisos—, y la marca, la etiqueta y los dos
  puntos de cada aviso de vigencia; cada tramo quitado deja un salto de línea. Así, «la que se consultó antes» dentro
  de la línea y «desde una consulta anterior» dentro de la frase no cuentan, y las mismas palabras fuera de ellas, sí. Se comparan por la forma, sin ningún
  modelo, con la tolerancia de las formas fijas de los avisos —sin distinguir mayúsculas, con blancos y énfasis de
  Markdown entre las palabras y alrededor— y delimitadas como palabras; lo mismo dicho con otras palabras no se
  detecta, y el informe publica cada respuesta. `Juzgar` la aplica, con sus cuatro familias, a cada sesión de las
  evals de `boe-legislacion` que activan la skill, con todos los modelos: cada expresión encontrada es un motivo
  `expresión prohibida: <expresión>` y la sesión no pasa, así que **decide en las evals que deciden**, con el umbral
  del ADR 0016 y la regla por serie de siempre, y cuenta en la tasa de las informativas sin decidir; cada respuesta
  cuenta además en los umbrales de su modelo (`umbrales`, abajo). Las evals de no activación y las de una skill sin
  lista se juzgan como antes. El informe publica, por sesión, `expresiones_prohibidas` en `informe.json` y la columna
  «Expresiones prohibidas» en la tabla de sesiones de `informe.md`, y, por modelo,
  `expresiones_prohibidas_por_modelo` —`con_alguna` sobre `respuestas`, las sesiones medidas de las series
  planificadas cuya eval activa la skill: solo las terminadas, sin las ilegibles ni las sin medir por límite de uso—
  y la sección «Expresiones prohibidas por modelo»; una sesión sin terminar se publica con su motivo y no pasa en su
  serie, pero no cuenta en ningún recuento ni umbral, y la de la prueba de red se juzga con la lista y publica sus
  expresiones, pero no entra en el recuento. Una lista mal formada —también una sin alguna de las cuatro familias o
  sin `formas_fijas`, con una vacía o con una forma de más de una línea— es un fichero mal formado: `make ci` falla
  nombrándola, y el informe la lista y da `fallo`. `make skills-check` comprueba además que la lista de
  `boe-legislacion` marca exactamente 36 de las 93 respuestas del informe de evals de H7.1, 11 de las 93 del de H7.2 y
  9 de las 93 del de H7.3, eval por eval, familia por familia y en las que llevan alguna, y ninguna otra; que no está
  en el texto de ningún bloque que leen sus evals ni en la respuesta compuesta con las formas fijas de los avisos y de
  los hallazgos y con los textos que su `SKILL.md` enseña a escribir, una vez quitadas las formas fijas de la lista, y
  que cada una de ellas quita algo de esa respuesta; y que **la prosa de su `SKILL.md`** —el fichero entero, frontmatter incluido, sin
  los bloques delimitados, los tramos de código en línea ni la tabla de comandos generada, párrafo a párrafo— no lleva
  ninguna expresión de la lista y el fichero ninguna fecha `AAAAMMDD` escrita con cifras: falla nombrando la primera
  línea de cada párrafo que lleva alguna, con sus expresiones, y cada línea con una fecha.
- **El formato común de eval gana seis piezas**, opcionales y, salvo `no_se_activan`, solo en una eval que activa la
  skill; las evals que ya había se leen y se juzgan igual. `comandos` admite una quinta forma, la comprobación (`applet` y `verbo` `check`,
  y, si se da, `norma`), que la cumple una invocación de ese applet con `check` —y con esa norma, si la eval la
  nombra— que termina con `0`. `hallazgos` lista las clases de hallazgo de `graph check` cuya forma fija tiene que
  llevar la respuesta —hoy solo `version-obsoleta`, la única a la que el binario da etiqueta—; se juzga sin ningún
  modelo, con las tolerancias de los avisos (el selector de presentación del emoji, el énfasis de Markdown, los
  espacios y las mayúsculas) y exacta en la etiqueta, y cada forma que falta hace que la sesión no pase con el motivo
  `forma de hallazgo ausente: <clase>`: cada sesión publica en `informe.json` `hallazgos_encontrados` y
  `hallazgos_ausentes`, la tasa de cada eval gana `formas`, con la forma literal que exige, y `informe.md` gana las
  columnas «Formas exigidas» en la tabla de tasas y «Hallazgos encontrados» y «Hallazgos ausentes» en la de
  sesiones. `make skills-check` comprueba además que el `SKILL.md` de `boe-legislacion` lleva la forma fija de
  `version-obsoleta` y que las clases que admite `hallazgos` son exactamente las que etiqueta el binario.
  `prohibidos` lista los `applet` y `verbo` que la sesión no puede invocar: toda invocación que consulta —no la
  ayuda, `--describe` ni `--dry-run`— de uno de ellos, termine como termine, hace que la sesión no pase con el motivo
  `comando prohibido ejecutado: <applet> <verbo>`; cada sesión publica en `informe.json` `comandos_prohibidos_ejecutados` y la tabla de sesiones de
  `informe.md` gana la columna «Comandos prohibidos ejecutados». Y `grafo_previo` nombra un directorio de
  `testdata/evals/grafo-previo/` y los bloques que, antes de la sesión, se consultan contra las grabaciones con ese
  directorio encima para dejar su observación en el grafo de la sesión; la caché de la sesión se prepara después, como
  siempre. `redacciones_modificadas` lista las líneas `⚠ REDACCIÓN MODIFICADA:` que tiene que llevar la respuesta, cada
  una con su `norma`, su `bloque` y sus dos fechas de vigencia, `AAAAMMDD`: `fecha_vigencia`, la de la redacción
  superada, y `fecha_vigencia_reciente`, la de la que cita. La cumple una línea con la forma fija de `version-obsoleta`
  cuya primera cita es de esa norma y ese bloque y cuyas dos primeras fechas son esas, en su orden; cada una que falta
  hace que la sesión no pase con el motivo
  `redacción modificada ausente: <norma> <bloque> <fecha_vigencia> <fecha_vigencia_reciente>`. Cada sesión publica en
  `informe.json` `redacciones_modificadas_encontradas` y `redacciones_modificadas_ausentes` (`[]` si la eval no las
  espera), las `formas` de la tasa de la eval ganan, por cada una,
  `⚠ REDACCIÓN MODIFICADA: <norma> <bloque> <fecha_vigencia> <fecha_vigencia_reciente>`, e `informe.md`, las columnas
  «Redacciones modificadas encontradas» y «Redacciones modificadas ausentes» en la tabla de sesiones. Y
  `no_se_activan`, en cualquier eval, nombra las skills que la sesión no puede activar: cada una que activa hace que no
  pase con el motivo `se activó la skill <nombre>, que la eval dice que no se activa`. Las tres evals de `legal-core`
  la llevan con `boe-legislacion`, porque lo que preguntan —qué comunidad, provincia y boletines corresponden a un
  municipio, o una receta— no depende de lo que dice ninguna norma. `schemas/eval.yaml.json` los valida y
  `make skills-check` comprueba que cada grafo previo nombrado existe y se prepara sin ninguna falta.
- **`umbrales` en el informe del job de evals, y los de las respuestas del modelo que decide deciden el veredicto.**
  `informe.json` lleva siempre `umbrales`, detrás de `expresiones_prohibidas_por_modelo`, con el contrato del
  ADR 0029: cada elemento con `nombre`, `descripcion`, `medida`, `total` —solo si lo que se compara es una
  proporción—, `comparacion` (`"<="`), `umbral`, `cumple` —la comparación en coma flotante de doble precisión y sin
  redondeos, de `medida` entre `total` (0 si `total` es 0) o de la propia `medida`— y `decide`. Una skill con lista de
  expresiones prohibidas tiene, con las respuestas medidas de cada modelo como `total` —las sesiones terminadas de las
  evals que activan la skill, también las informativas, y no la de la prueba de red ni las sin medir—, en este orden:
  del modelo que decide, `expresiones_prohibidas:<modelo>`, con las que llevan alguna expresión y el `umbral` `0.05`;
  **`sin_activar:<modelo>`, con las que no activan la skill, y `redaccion_no_leida:<modelo>`, con las que llevan alguna
  expresión de `redaccion_no_leida`, los dos con el `umbral` `0`**, y los tres con `decide: true`; y, de cada modelo
  informativo, `expresiones_prohibidas:<modelo>`, con `decide: false`, que solo se publica. En `boe-legislacion`, con
  54 respuestas de `claude-sonnet-5-5`, el primero admite como mucho 2 con alguna expresión, y los otros dos, ninguna;
  el de `claude-haiku-4-5-20251001` va sobre 30. Una respuesta sin la skill activada y con una expresión cuenta, una
  vez, en cada umbral que la mide. Un umbral con `decide: true` que no se cumple pone el veredicto en `fallo` con el motivo
  `umbral <nombre>: <medida> de <total> (<p> %), y tiene que ser ≤ <u> %`, con un decimal y coma
  (`umbral sin_activar:claude-sonnet-5-5: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %`), y el job sale en rojo; uno que
  se cumple, o que no decide, no cambia el veredicto, y la regla por serie del ADR 0016 no cambia.
  `informe.md` gana la sección «Umbrales», detrás de «Expresiones prohibidas por modelo», con la tabla
  `Umbral | Medida | Condición | Cumple | Hace fallar el veredicto` —`no: solo se publica` en los que no deciden— o el
  párrafo `ninguno`. `legal-core`, sin lista y sin objetivo de duración, da `umbrales: []`.
- **Sesiones sin medir por límite de uso.** Una sesión que un límite de la cuenta no dejó terminar no es una eval
  fallida: el informe la clasifica sin modelo, con lo que registra su transcript, en una de tres clases —(a) el
  mensaje del límite de uso de Claude Code en su resultado (el que empieza por `You've hit your`,
  `You've reached your` o `You're out of`), (b) reintentos por `rate_limit` hasta agotarlos o (c) reintentos por
  `rate_limit` hasta que la corta el tope— y la deja **sin medir**: no pasa ni falla, lleva como único motivo
  `sin medir por límite de uso: <clase>`, no entra en el recuento de expresiones prohibidas ni en los umbrales, y su
  serie queda sin medir, sin pasar y sin el motivo «pasan N de M». Otros reintentos, como los de la sobrecarga, no
  son un límite de la cuenta, y una sesión que se recupera de sus reintentos se mide como cualquier otra. Tras una
  sesión de la clase (a), que no se repone dentro del job, el job no abre ninguna más: las abiertas terminan y se
  juzgan, y las que faltaban del plan cuentan en su serie como sin medir, `sin abrir tras el límite de uso`, sin el
  motivo de las sesiones que faltan; tras (b) o (c), sigue abriéndolas. Con alguna sesión sin medir, el veredicto es
  `fallo` con un solo motivo propio, que se distingue sin modelo de los de la skill,
  `de la ejecución, no de la skill: límite de uso de la cuenta: <n> sesiones sin medir: <sesión> (<clase>), …`, y el
  job sale en rojo. `informe.json` gana `sesiones_sin_medir` —`sesion`, `eval`, `modelo` y `motivo`, en orden de
  sesión; `[]` si ninguna—, `sin_medir` en cada sesión (la clase, o `""` si se midió) y en cada tasa (cuántas); e
  `informe.md`, la sección «Sesiones sin medir» detrás de «Umbrales» (o `ninguna`), `sin medir (<n>)` en la columna
  «Resultado» de la serie y la columna «Sin medir» en «Sesiones». El motivo de una sesión que no terminó lleva además
  el texto del último `result` con `is_error` de su transcript, con el código 0 o con otro que no es el del tope
  (`código 1: result con is_error: Failed to authenticate. …`).
- **Reintentos por límite de ritmo.** El informe publica los reintentos por `rate_limit` que registra el transcript
  de cada sesión (`system/api_retry`): en `reintentos_por_limite_de_ritmo` de cada sesión y, sumados, en la raíz de
  `informe.json`, y en la línea `Reintentos por límite de ritmo: <n>` de la cabecera de `informe.md` y en la columna
  del mismo nombre de «Sesiones». Son el dato con el que se ajusta la concurrencia del job.
- **La duración de las sesiones, con su objetivo de 900 s en `boe-legislacion`.** El informe publica
  `duracion_de_las_sesiones`, los segundos desde que se prepara la primera sesión hasta que termina la última,
  redondeados hacia arriba y sin la preparación del runner, y la línea `Duración de las sesiones: <s> s` en la
  cabecera de `informe.md`. Con un objetivo mayor que 0 (`OBJETIVO_DE_DURACION_DE_EVALS`: 900 en `boe-legislacion`;
  `legal-core` no tiene), `umbrales` lleva `duracion_de_las_sesiones`, sin `total` y con `decide: true`: por encima,
  el veredicto es `fallo` con el motivo
  `de la ejecución, no de la skill: duracion_de_las_sesiones: <s> s, y tiene que ser ≤ 900 s`, y el job sale en rojo.
- **Sondeo local, `make evals-sondeo`, que no es un veredicto.**
  `make evals-sondeo SKILL=<skill> EVALS=<nn>[,<nn>…] MODELO=<id> REPETICIONES=<n> [CONCURRENCIA=<n>]` abre a la vez,
  en macOS o en Linux y sin strace ni `sudo`, las sesiones de unas evals de una skill con un solo modelo —con
  `CONCURRENCIA` vacía, como mucho las que abre a la vez el job para esa skill—, con el binario y la skill del árbol
  de trabajo, que construye e instala en un directorio temporal, sin la configuración de Claude Code de quien lo lanza
  y con una sola credencial, `CLAUDE_CODE_OAUTH_TOKEN`, el token de la suscripción que da `claude setup-token`. **Un
  error de uso se dice sin la traza de `go test`**: con un argumento que no vale, o sin la credencial o con ella
  vacía, la salida de error es solo su mensaje —una línea por argumento, que lo nombra, en el orden `SKILL`, `EVALS`,
  `MODELO`, `REPETICIONES` y `CONCURRENCIA`, o la línea de la credencial que falta—, sin nada de `go test` ni de
  testify, la salida estándar queda vacía, no se abre ninguna sesión y el guion termina con `1` (`make` añade detrás
  su propia línea de error y termina con `2`); un fallo al leer la definición del job, construir el binario, instalar
  las skills, crear el directorio de las sesiones o repartirlas sigue dando el registro entero de `go test`. Cada
  sesión se prepara y se juzga con el código del job, sin lo que el job lee de la traza de strace —los comandos esperados y los prohibidos de cada eval y las llegadas a la red—, y tras el mensaje
  del límite de uso no abre ninguna más. Su salida empieza por «Esto es un sondeo, no un veredicto: el veredicto de la
  skill lo da el job de evals.», dice lo que no comprueba y da solo lo agregado: la tasa de cada serie, las respuestas
  con alguna expresión prohibida en las evals que activan la skill —solo de las sesiones terminadas, como en el job—,
  con el 5 % como referencia, y las sesiones sin medir y las que no llegaron a terminar por otra causa, con su
  motivo. No escribe `informe.json` ni ningún veredicto, termina con `0` si
  ha podido abrir y juzgar sus sesiones, sean cuales sean sus tasas, y borra su directorio temporal al terminar, con
  el código que sea. Abre sesiones con modelo y consume la suscripción de quien lo lanza: ni `make ci` ni ningún
  flujo lo ejecutan.

### Cambiado

- **El job de evals decide con `claude-sonnet-5-5`** (ADR 0031, que sustituye al ADR 0016 en qué modelo decide): el id
  al que resuelve el alias `sonnet` en Claude Code 2.1.284, la versión que ahora instalan sus sesiones (antes
  `claude-sonnet-5` con 2.1.270, que no reconoce el modelo nuevo). Los dos van juntos en `.github/workflows/evals.yml`.
  Sonnet 5 sale del job y `claude-haiku-4-5-20251001` sigue como límite inferior informativo, así que el umbral de
  las expresiones prohibidas que decide pasa a llamarse `expresiones_prohibidas:claude-sonnet-5-5`. Una sesión solo
  pasa el control del modelo si declara el id pedido, tal cual o seguido de la fecha de su versión (`-AAAAMMDD`): antes bastaba con que empezara por
  él, y `claude-sonnet-5-5` pasaba por `claude-sonnet-5`.
- **`boe-legislacion` v0.1.4**: **se activa ante toda pregunta cuya respuesta dependa de lo que dice una norma** —qué
  dice un artículo, una ley o un real decreto, qué plazo, requisito o procedimiento fija, dónde se regula una
  materia—, **también si el modelo cree saberla**: su `description` dice que sin leer la norma con el binario la
  respuesta no tiene cita. La respuesta empieza por lo que se pregunta: es todo lo que el agente escribe después de la
  última orden, desde su primera palabra, y está hecha de la norma, su texto, su cita, sus avisos de vigencia y la
  línea `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae; **no cuenta la comprobación, lo que el agente ha hecho,
  lo que va a hacer ni lo que ha devuelto ninguna orden**, y lo que decide al mirar si le falta contexto no va en
  ella. Tampoco cuenta lo que una orden no ha devuelto: un `kitlegal graph check` sin entradas, que es lo habitual, no
  deja rastro en la respuesta, ni junto a los avisos de vigencia ni al final, porque lo que compara es lo leído en
  otras conversaciones, que quien pregunta no conoce, y no un dato de la norma. La prosa de la skill no enseña el
  vocabulario que la respuesta no puede decir: nombra cada orden por su nombre, en código, y no habla de «la
  comprobación» ni de una lectura anterior. **No describe una redacción que no ha
  leído**: `kitlegal boe` da solo la vigente y `kitlegal graph check`, dos fechas, así que la respuesta no dice qué
  decía la redacción superada, ni lo resume, ni lo compara, aunque el modelo crea saberlo; dice el texto vigente con
  su cita, qué norma le dio esa redacción y desde cuándo rige, nunca hasta cuándo —ningún sobre trae el fin de una
  redacción, tampoco en una norma derogada, cuyo aviso no lleva fecha—, y, si se pregunta qué cambió, **que cita la
  redacción vigente y que la anterior no la ha leído**. Lee cada bloque una sola vez por pregunta y comprueba si su
  redacción ha cambiado **en la misma orden que lo lee**, detrás de la lectura y con su misma norma y sus mismos
  bloques (`kitlegal boe articulo BOE-A-2015-10565 a21 --json && kitlegal graph check BOE-A-2015-10565 a21 --json`):
  una comprobación por cada orden de lectura —una por norma citada en la mayoría de las preguntas—, nunca antes de
  leer ni sin argumentos, de modo que la última orden antes de la respuesta es la que trae el texto y no una
  comprobación suelta. **Cada orden de lectura y comprobación, la de `articulo` y la de `articulos`, tiene detrás su
  forma para PowerShell**, que en su versión 5.1 no tiene `&&`, con la misma norma y los mismos bloques
  (`kitlegal boe articulo BOE-A-2015-10565 a21 --json; if ($LASTEXITCODE -eq 0) { kitlegal graph check BOE-A-2015-10565 a21 --json }`),
  y `make skills-check` lo comprueba. Por cada `version-obsoleta` lleva una línea con la forma fija
  `⚠ REDACCIÓN MODIFICADA:` —`⚠`, la etiqueta que da el binario y dos puntos— que **dice de qué precepto es, con su
  cita**, la del bloque como en cualquier otra cita, y dos puntos, seguida en la misma línea de las dos fechas de
  vigencia tal como las da la comprobación, la de la redacción superada y la de la que cita: con dos bloques
  cambiados, dos líneas, cada una con su cita y sus fechas. El ejemplo de la línea lleva marcadores (`AAAAMMDD`) en
  lugar de datos que copiar; decirlo con otras palabras no vale, y la etiqueta no es la de ningún aviso de vigencia.
  No lleva `fuente-caducada`, porque cita siempre lo que acaba de leer. **No habla de lo dicho en otra
  conversación**, ni para afirmarlo, ni para confirmarlo, ni para desmentirlo, y de lo leído en otras conversaciones
  solo lleva esa línea. El texto que cita sale siempre de `kitlegal boe articulo` o `articulos`, nunca de la salida
  de `graph`. Si `graph check` termina con otro código que `0`, responde igual con el texto leído y lleva, tal cual y
  sin nada más sobre la redacción, «No se ha podido comprobar si la redacción ha cambiado desde una consulta
  anterior.»; si falla `kitlegal boe`, dice qué no pudo consultar por lo que significa para quien pregunta —que el
  artículo no está en la norma, que la fuente no estaba disponible o que limitó las consultas—, sin el código, y no
  suple el texto. Su frontmatter declara `kitlegal-applets: boe graph` y su tabla de comandos gana `kitlegal graph`
  (`show`, `stats` y `check [<norma> [<bloques>...]]`), generada con `make skills-sync`, que escribe los argumentos de
  posición opcionales como la ayuda del binario. La forma de la cita y la de los avisos de vigencia no cambian.
  Sustituye a la v0.1.3 de H7.3, que no llegó a publicarse: su `description` la pedía cuando se preguntara qué dice
  un artículo, una ley o un real decreto, y el modelo que creía saber la respuesta la daba sin leer la norma; su
  prosa hablaba de «la comprobación» y de una lectura anterior y preguntaba si lo leído bastaba, y la respuesta lo
  repetía para contar que la redacción no había cambiado o que no necesitaba leer más; solo prohibía hablar de la
  redacción anterior, sin decir por qué ni qué responder si se preguntaba qué cambió, y la respuesta describía la
  redacción superada, que ninguna orden había devuelto; la línea `⚠ REDACCIÓN MODIFICADA:` no decía de qué precepto
  era; y sus órdenes, con `&&`, no tenían forma para PowerShell. A la v0.1.2 de H7.2, que tampoco llegó a publicarse:
  su prosa enumeraba lo que la respuesta no podía nombrar —la memoria de consultas, los códigos de salida, los
  hallazgos—, trataba el caso en que la comprobación no tenía nada que decir, pedía la comprobación sola, como última
  orden y unida a «antes de redactar la respuesta», y el ejemplo de la línea llevaba dos fechas que copiar; a la
  v0.1.1 de H7.1, que tampoco llegó a publicarse: contaba la comprobación en la respuesta, no fijaba cómo se escriben
  las fechas de la forma, decía que no había podido comprobar la memoria de consultas cuando `graph check` fallaba y
  no prohibía el código cuando fallaba `kitlegal boe`; y a la v0.1 de H7, que tampoco llegó a publicarse: comprobaba
  dos veces por pregunta, antes y después de leer y sin argumentos, y trasladaba también `fuente-caducada`.
- **`--no-graph` tiene efecto**: la invocación no entrega nada al grafo, ni resuelve la ruta de `world.db` ni lo
  abre, y su ayuda dice «No entrega al grafo del mundo nada de lo que observa la invocación.», en lugar de «Declara
  que la ejecución no altera el grafo.». La salida estándar y el código son los mismos con la bandera y sin ella.
  Desde H1 se aceptaba sin semántica, a la espera del grafo.
- **`territorio resolver` escribe en la carpeta de la caché**: sigue sin pedir nada a la red y sin tocar `cache.db`,
  pero lo que observa llega a `world.db`, que la primera entrega crea en esa carpeta.
- **La ayuda y los errores del binario enumeran `graph`**: `kitlegal --help` lo lista y el error de un applet
  desconocido dice `applets disponibles: boe, graph, skills, territorio`.
- **`schema.Resultado` gana `Grafo`**, lo que el applet ha observado: su vigencia y sus operaciones —`Nodo`, `Arista`
  y `Texto`, ninguna con fuente, url ni fecha, que pone el kernel desde el sobre—. Vacío, no se entrega nada, así que
  el contrato `Applet` (ADR 0005) no cambia y ningún applet que no emite implementa nada. En el kernel,
  `cli.Montador` gana el almacén al que entrega y `Emitir` recibe el contexto con el plazo de `--timeout`.
- **`make test-tiempos` mide también el coste del grafo** (`TestCosteDelGrafo`), solo y sin la caché de resultados de
  `go test`, como `TestMedidasDeTiempo`, y `make test` y `make test-integration` lo saltan: con un `world.db` que ya
  existe, la entrega no añade más de 150 ms a la mediana de 20 `boe articulo` servidos desde la caché respecto de los
  mismos con `--no-graph`, y sobre un grafo de 10 000 nodos y 10 000 aristas `graph check` tarda menos de 3 s y
  `graph stats` menos de 1 s (medianas de cinco). Las trece cotas de 200 ms de `boe articulo` desde la caché y de
  `territorio resolver`, que ahora entregan al grafo, no cambian.
- **`make test-integration` ejecuta la matriz del grafo** (`internal/graph/integracion_test.go`): esquema,
  idempotencia, orden de llegada, rechazos, ocho entregas a la vez, un `world.db` que no es una base de datos y uno de
  una versión posterior, el plazo y el bloqueo, la lectura con el `-wal` de una escritura propia interrumpida, un
  `world.db` escrito por H7 y lo que deja una entrega que falla. Y **la medida de `graph check`**
  (`internal/app/medida_test.go`): sobre un grafo sembrado de 300 normas y 2 400 bloques, 240 de ellos con dos
  redacciones leídas y el 90 % consultado hace más de una semana, `graph check --json` sin argumentos da 50 hallazgos,
  todos `version-obsoleta`, con los totales de cada clase y los omitidos, en 40 000 bytes como mucho, y con la norma o
  con la norma y un bloque, solo los de ese ámbito; y lo que lee la skill con cinco bloques cuya redacción ha
  cambiado, cinco `version-obsoleta` en 3 800 bytes como mucho.
- **`make lint` y `TestArquitectura` vigilan una regla más, R6**: `internal/graph` no importa las fuentes
  (`internal/source`) ni la presentación (`internal/render`); y `internal/graph` es, con `internal/cache`, el único
  paquete que importa `database/sql` y SQLite. El binario no enlaza ningún módulo nuevo.
- **`make test-e2e` construye además tres binarios de extremo a extremo con el reloj fijo** (el 28 y el 29 de
  septiembre y el 6 de octubre de 2026 a las 12:00 UTC), con los que `graph check` da hallazgos reproducibles, y
  copia a cada guion cuatro respuestas del BOE derivadas de las grabaciones de H4 —dos redacciones posteriores del
  artículo 21 de la Ley 39/2015, con fecha de vigencia `20250101` y `20260101`, y los metadatos de esa ley sin ELI,
  con la `url_eli` vacía y con una que no tiene ningún segmento `eli`—, sin ninguna grabación nueva.
- **Con 20 evals, el trabajo de `boe-legislacion` del job de evals abre 96 sesiones**: 60 de `claude-sonnet-5-5` —36
  sobre las doce que deciden y 24 sobre las ocho informativas— y 36 de `claude-haiku-4-5-20251001` sobre las doce que
  deciden.
- **El job de evals juzga la respuesta a la pregunta**: la de una sesión es el primer `result` de su transcript, si
  termina bien (`subtype` `success` e `is_error` falso), y no el último. El aviso de una tarea en segundo plano que
  termina después de la respuesta abre otro turno, con su propio `result`, y su réplica ya no sustituye a la
  respuesta. Si la sesión terminó, y el texto del error con el que no terminó, se siguen leyendo del último mensaje, y
  las skills activadas, de todo el transcript. Lo comparten el job y el sondeo.
- **El job de evals abre las sesiones de una skill a la vez, de cuatro en cuatro en `boe-legislacion`, con una sola
  tanda por commit y skill**, y con lo que garantiza cada sesión sin cambiar. `scripts/evals.sh` exige `CONCURRENCIA_DE_EVALS`,
  cuántas sesiones abre a la vez como mucho —un valor que no es un entero mayor o igual que 1 lo termina con `1` antes
  de la primera sesión, como las repeticiones—, lee `OBJETIVO_DE_DURACION_DE_EVALS`, opcional (`0`, sin objetivo), y
  ya no necesita `timeout`. Tras las comprobaciones de siempre ejecuta una sola orden de Go, `TestEjecucionDelJob`
  (etiqueta `evals`), que compone el plan, reparte sus sesiones, las juzga y escribe el informe con la duración, y que
  falla con un error o con el veredicto `fallo`, en lugar del plan en un fichero, el bucle de sesiones en bash y las
  tres entradas de Go que planificaban, preparaban cada sesión y escribían el informe, que se retiran. Cada sesión se
  prepara justo antes de abrirla, en su propio directorio —el de trabajo, su caché y su grafo, el estado de Claude Code (`CLAUDE_CONFIG_DIR`) con las
  skills tal como las deja `make install`, y su temporal—, sin escribir en nada de otra; la abre el guion nuevo
  `scripts/evals-sesion.sh` con la orden de `claude` de siempre, bajo `strace` y con el proxy que rechaza toda petición
  salvo la del modelo, en su propio grupo de procesos, y el tope de 240 s lo pone el repartidor, en Go: `TERM` al
  grupo y, 10 s después, `KILL`, con los códigos 124 y 137 de siempre. Ninguna sesión se reintenta ni se abre dos
  veces, `SIGINT` o `SIGTERM` cierran las abiertas y, sin límites de uso, el informe es el mismo que en serie, salvo
  los tiempos. En `.github/workflows/evals.yml`, la matriz da a `boe-legislacion` `concurrencia` 4 y
  `objetivo_de_duracion` 900, y a `legal-core` 1 y 0, que el trabajo pasa en esas dos variables; el trabajo se llama
  `evals (<skill>)`. **Una sola tanda por commit y skill**: un trabajo nuevo, `tanda`, que corre antes que ellos y con
  la condición que antes llevaban —el despacho, la etiqueta `evals` o `evals-prueba-de-red`, o la apertura que toca lo
  que las evals miden—, decide si la ejecución mide el commit: no lo mide si una ejecución anterior del flujo sobre el
  mismo commit, sin terminar, ya lo mide. Lo decide `TestTandaDelCommit` (etiqueta `evals`), que lo consulta con
  `gh run list` y `gh run view` y, mientras una anterior no ha decidido, vuelve a consultar cada 10 s, como mucho
  10 min; nunca espera a la tanda de otra, y deja `medir=si` o `medir=no` como salida del trabajo. Su último paso,
  «Esta ejecución mide el commit», solo corre en la que mide, y es lo que leen las posteriores. Los trabajos
  `evals (<skill>)` corren solo con `medir=si`: un segundo disparo sobre el mismo commit mientras el primero sigue
  —el de la apertura y el de la etiqueta— los salta enteros, sin abrir ninguna sesión ni dejar una comprobación roja
  o de más, y la etiqueta sobre un commit cuya tanda ya terminó vuelve a medir. `tanda` no lleva `concurrency`, así
  que ni espera en cola ni se cancela; la de cada trabajo de skill, por commit y skill con
  `cancel-in-progress: false`, se queda, y si la espera de `tanda` se agota con una anterior sin decidir, la ejecución
  mide detrás de la otra, sin cancelar ninguna. `timeout-minutes: 122` es el tope de un cuelgue, que cubre el peor
  caso de cada skill, y no el control de la duración; y el filtro de lo que las evals miden gana
  `scripts/evals-sesion.sh`. `TestDefinicionDelJob` comprueba esa definición en `make ci` —el trabajo `tanda`, su
  condición, sin `concurrency`, con `actions: read`, su salida `medir`, la orden de su paso `decidir` y la marca con su
  condición como último paso; que los de las skills dependen de `tanda` y solo corren con `medir=si`; el grupo,
  `cancel-in-progress`, el nombre, la concurrencia y el objetivo de cada skill, y que `timeout-minutes` cubre su peor
  caso, `485 s + ⌈N / C⌉ × (22 s + 240 s + 10 s)`, con `N` las sesiones de su plan con la de la prueba de red y `C`
  su concurrencia—, y prueba la decisión de la tanda con consultas sintéticas y sin red; y los tests del repartidor,
  con sustitutos de `claude` y `strace`, lo prueban sin modelo, también en macOS.

El contenido del grafo es solo lo que se ha dicho. De los **bytes** de `world.db` y de los ficheros auxiliares que
SQLite pone junto a él se promete solo esto, y solo de lo que deja el propio binario
(`specs/011-h7-1-graph-check-acotado/contracts/almacen-world-db.md`):

- **Leer sin `world.db-wal`** no cambia ni un byte de nada en la carpeta.
- **Leer junto al `-wal`** de una escritura propia interrumpida o de otra invocación abierta: los verbos de `graph`
  ven lo confirmado en él, dejan `world.db` y el `-wal` con los mismos bytes y dejan un `world.db-shm`.
- **Una entrega junto al `-wal` que dejó una escritura interrumpida** lo recupera aunque falle después, como cualquier
  escritor de SQLite: lleva a `world.db` lo confirmado en él y retira el `-wal` y el `-shm`.

Cualquier otro estado —un directorio en el lugar de `world.db`, un enlace, permisos cambiados, un diario de rollback o
un `-shm` ajenos, una base escrita por otra aplicación— sigue la **regla genérica** (constitución, «Gates»): si el
binario no puede usarlo, es un defecto `inesperado` —`1` con la ruta y la causa en los verbos de `graph`, una línea
de aviso en la entrega— y no se promete nada de sus bytes ni de su recuperación. Por eso H7.1 retira de H7, antes de
que llegara a publicarse y con sus tests, lo que solo servía a esos estados o a entradas que ningún emisor produce: la
publicación de `world.db` desde un temporal `world.db-nuevo-*` con un enlace duro; la comprobación del permiso de
escritura al leer y al entregar, y la reapertura en solo lectura inmutable; el trato propio de un enlace simbólico,
de un diario de rollback, de un `-shm` suelto y de una base de fuera, con sus listas de bytes; los mensajes de un
directorio, de una transacción interrumpida y de un fichero sin permiso de escritura, y el «; no se modifica» de un
`world.db` que no es una base de datos; los rechazos del lote por una `url` que no es un URI absoluto, un nodo sin id
o sin tipo o una arista sin relación o sin sus extremos —una arista con un extremo que no está la sigue impidiendo la
clave ajena de las aristas, que deshace la entrega entera—; los niveles del desempate de dos observaciones del mismo
instante más allá de la `url`; los ejemplos no ASCII de la regla de `Persona`, que se queda con los ASCII; y el orden
inverso de H7, en el que leer una redacción anterior después de otra posterior daba `version-obsoleta` sobre la
anterior, y ahora no da nada: pedir una redacción anterior a propósito llegará con `--fecha`, en H20.

## [0.3.1] - 2026-09-28

### Cambiado

- **`kitlegal skills install`, `list` y `doctor` hablan a una persona** (ADR 0026): sin `--json`, en lugar de la
  tabla mínima del sobre (`fuente kitlegal.skills`, `0.enlaces.0.host claude`…) dicen dónde han quedado las skills y
  qué agente las lee de ahí, cada skill con su estado o su versión y sus entradas por marca (Claude Code,
  Antigravity), con `~` en las rutas de la cuenta; si un agente no las verá, la orden que lo enlaza; y al terminar,
  qué hacer ahora. `doctor` dice «Todo en orden» o explica cada hallazgo en una frase, con su orden lista para copiar
  debajo. Con `--json` no cambia ni un byte: ni el sobre, ni `schemas/instalacion.json`, ni los códigos de salida.
  `--dry-run` usa el mismo vocabulario: `Claude Code .claude/skills/legal-core (enlace)` donde antes decía
  `enlace .claude/skills/legal-core (enlace)`. `boe` y `territorio` siguen con la tabla mínima.
- **`schema.Resultado` gana `Legible`** (ADR 0026): el contenido contado para una persona, que el kernel escribe sin
  `--json` en lugar de la tabla mínima. Vacío, la tabla como siempre; el presentador sigue sin conocer applets.

### Añadido

- **`kitlegal --version`** hace lo mismo que `kitlegal version`: el mismo texto y las mismas reglas, también que
  no admite nada detrás. La ayuda del binario dice cómo pedir la versión, y el error de un applet que no existe
  lo recuerda.
- **https://kitlegal.es/llms.txt**, el índice de la web para agentes de IA ([llmstxt.org](https://llmstxt.org/)):
  qué es kitlegal, cómo se instala, las skills de la última versión publicada con su descripción y los esquemas.
  Cada página lo anuncia con `<link rel="describedby">`. Sus cifras, órdenes y enlaces no se escriben a mano:
  salen de las páginas, de `skills/` en la etiqueta de esa versión y de `schemas/`.

## [0.3.0] - 2026-09-27

### Añadido

- **La web del proyecto, https://kitlegal.es** (ADR 0024), en `web/`: una portada para despachos y otra para
  ciudadanía, la guía de instalación, la página del rastreador (`/bot/`) y los esquemas de `schemas/` en la
  dirección de su `$id`. Ninguna cita de la web está escrita a mano: cada fragmento sale del sobre que devolvió
  `kitlegal`, y la construcción falla si no está literal en él. Órdenes nuevas, fuera de `make ci`: `make web`,
  `make web-dev` y `make web-citas`.
- **Antigravity como host** (ADR 0025): Antigravity lee sus skills globales en `~/.gemini/config/skills/` y no en
  `~/.agents/skills/`, así que `kitlegal skills install -g` enlaza ahí cada skill, con el destino relativo
  `../../../.agents/skills/<skill>`, cuando existe `~/.gemini/config/`. `--host antigravity` lo fuerza aunque no
  exista. En un proyecto no hace falta: Antigravity lee su `.agents/skills/`.

### Cambiado

- **kitlegal se identifica con `kitlegal/x.y (+https://kitlegal.es/bot)`**, una dirección que ahora explica qué es y
  cómo limitarlo. Hasta la 0.2.0 era `https://ventanillalegal.es/bot`.
- **El `$id` de cada esquema de `schemas/` pasa a `https://kitlegal.es/schemas/…`** (incompatible para quien los
  referencie por su `$id`), y cada esquema se puede descargar en esa dirección.
- **`--host` se puede repetir** (ADR 0025; incompatible): `--host claude --host antigravity`, con un valor por
  bandera, y admite `claude` y `antigravity`; el rechazo dice «los hosts admitidos son claude y antigravity».
  `--describe` de `skills install` lo declara como lista, y el `host` de cada entrada de la salida admite
  `antigravity`, con las de `claude` primero. La orden que da `doctor` lleva un `--host` por cada host en el que la
  skill tiene una entrada.
- Un manifiesto local que declara una entrada de `antigravity`, que solo tiene el ámbito global, hace el ámbito
  ilegible (código 7), como ya lo hacía cualquier entrada de host con `--dir`.
- Un binario anterior da por ilegible un manifiesto global que declara `antigravity`: para volver a una versión
  anterior, retira antes esas entradas.
- **El README ya no promete Claude Cowork**: Cowork y el chat de Claude cargan las skills de la cuenta, no las del
  disco. Explica también cómo evitar que Codex pida permiso en cada consulta al BOE.

## [0.2.0] - 2026-09-27

### Cambiado

- **La licencia pasa de Apache-2.0 a EUPL-1.2** (ADR 0022): `LICENSE` con el texto oficial de la Licencia Pública de
  la Unión Europea v. 1.2, y el identificador SPDX `EUPL-1.2` en el bucket de Scoop y en los paquetes `.deb` y `.rpm`
  (el cask de Homebrew no declara licencia). v0.1.1 sigue bajo Apache-2.0.
- **La descripción del cask, el bucket y los paquetes** dice lo que kitlegal hace para quien lo usa: «Tu asistente de
  IA responde con la ley vigente del BOE y la cita exacta».
- **Contrato de resultados** (ADR 0023; incompatible): cada orden devuelve un resultado de una tabla única de
  códigos, y dos cosas cambian de código.
  - **`kitlegal skills doctor` sale con 0 aunque encuentre algo**, con cada hallazgo en `data.hallazgos` (`clase`,
    `ruta` y la `orden` que lo arregla). Antes salía con 1 y el sobre de fallo, como si el programa hubiera fallado.
    Un guion que quiera fallar con hallazgos mira `data.hallazgos`.
  - **Código 7 y clase `conflicto`** para lo que el estado local impide y la persona puede resolver: `skills install`
    ante una entrada que no es suya (también con `--dry-run`), `skills list` y `doctor` ante una ruta del ámbito que
    no es un directorio o un manifiesto que no es un fichero regular, no respeta su forma o, con `--dir`, declara
    entradas de host, y `-g` sin `HOME`. Antes salían con 1, el código del fallo inesperado, que queda para los
    defectos del programa o del entorno: un error del sistema al leer el manifiesto sale ahora con 1, como ya salía
    un error al examinarlo.
  - El sobre de fallo de todo `--describe` y de `schemas/*.json` admite la clase `conflicto`.

## [0.1.1] - 2026-09-27

Primera versión publicada. La etiqueta `v0.1.0` existe, pero no tiene release: su publicación falló antes de construir
nada, porque `release.yml` fijaba `sigstore/cosign-installer@v4`, una etiqueta que esa acción no publica, y se corrigió
en esta versión. Lo que sigue es lo que aportan los hitos cerrados hasta hoy, en orden de llegada: **H0 — esqueleto del repositorio y sus controles**, un repositorio sin fuentes legales
todavía, pero blindado, para que cualquier línea de Go que entre después atraviese los mismos gates; y
**H1 — kernel de la línea de órdenes**, lo que un applet **no** tiene que declarar, de modo que cada
fuente legal de los hitos siguientes herede la misma forma de invocarse, de fallar y de citar sin volver
a escribirla. **H3 — caché local en SQLite** es un hito de fundación que no cambia nada en el binario,
así que su única entrada, en *Cambiado*, es lo que cambia en `make ci`. **H4 — applet `boe`** trae la
primera fuente legal: la legislación consolidada del BOE, consultable y citable desde el binario
distribuido, con caché, contratos de salida publicados y una verificación nocturna contra la fuente real.
**H5 — skill `boe-legislacion`** trae la primera skill del producto, que consulta y cita cualquier norma
consolidada del BOE con ese binario, y el andamiaje que comparten todas las skills: una tabla de normas como
única fuente de verdad, lo generado comprobado en `make ci`, la instalación con `make install` y el formato
común de eval con su job de evals. El binario distribuido no cambia. **H6 — applet `territorio` y skill
`legal-core`** trae el segundo applet, el primero sin fuente que consultar en red: resuelve cualquier municipio de
España a su territorio desde datos congelados que viajan dentro del binario, y declara lo que no está configurado en
lugar de inventarlo; y la skill madre `legal-core`, que empieza toda pregunta por el territorio. **H19 — instalar
sin clonar** trae la distribución (ADR 0019): el binario se instala con el gestor de paquetes de cada plataforma o con
`install.sh`, lleva las skills dentro y las instala con el applet `skills`, las skills invocan `kitlegal` desde el
`PATH` y `make install` pasa a ser el bucle de desarrollo.

### Añadido

*De H0 — el repositorio y sus controles:*

- **Binario `kitlegal`** con un único verbo, `version`, que imprime en tres líneas de la salida estándar
  la versión, el commit y la fecha de construcción, deja vacía la salida de error y termina con código
  `0`. La versión sale de `git describe --tags --always --dirty` y el commit coincide carácter a carácter
  con la revisión construida. Cualquier otra invocación —sin verbo, con un verbo desconocido o con un
  argumento sobrante— escribe una línea de uso en la salida de error y termina con código `2` («args» en
  la tabla de códigos de salida estables del proyecto).
- **`Makefile` como única superficie de invocación de los controles**, con `make ci` como veredicto del
  repositorio: la misma orden que se ejecuta en local es la que ejecutan el gancho de pre-commit y la
  integración continua, y ningún control se aplica por otra vía. `make ci` no modifica ningún fichero
  versionado. `make build` y `make install` compilan sin cgo, con `-trimpath` e inyectando los datos de
  construcción; `make help` es el objetivo por defecto.
- **Ocho controles activos dentro de `make ci`**: formato en modo verificación (`gofumpt` y `goimports`),
  análisis estático (`golangci-lint`, con `gosec` y `govet` incluidos), tests unitarios con detector de
  carreras y perfil de cobertura, vulnerabilidades conocidas (`govulncheck`), validación contra esquemas,
  detección de secretos (`gitleaks`), integridad de los módulos (`go mod verify`, raíz y herramientas) y
  dependencias saneadas (`go mod tidy -diff`). `make check-tools` comprueba los prerrequisitos como
  dependencia de las demás órdenes.
- **Cadena de herramientas reproducible y sin instalación previa**: los cuatro controles con binario
  propio —`golangci-lint`, `govulncheck`, `gitleaks` y `lefthook`— se construyen solos desde la versión
  fijada en `tools/<herramienta>/go.mod`, y `go.mod` declara la directiva `toolchain` (`go1.27.1`, la
  estable actual al cerrar el hito) que el `Makefile` exporta como `GOTOOLCHAIN`, de modo que todas las
  órdenes usan el mismo parche de Go que la integración continua. Los dos únicos prerrequisitos son una
  cadena Go 1.21 o superior y `git`.
- **Ganchos de pre-commit** (`make hooks`, con `lefthook`): cada commit corrige el formato y vuelve a
  preparar lo corregido, y ejecuta `lint-fast`, `secrets` y `mod-tidy-check`. Es un subconjunto rápido, no
  el veredicto: la autoridad final sigue siendo la integración continua.
- **Umbrales de cobertura bloqueantes** (`codecov.yml`): ≥ 70 % global y ≥ 85 % en el dominio interno,
  ambos declarados como estado que falla y no como información.
- **Flujos de la plataforma**: `ci` en cada propuesta de cambio y en cada push a la rama principal,
  `nightly` a diario sobre la rama principal —ambos se limitan a preparar el entorno y ejecutar `make
  ci`—, `codeql` semanal como segundo análisis de seguridad y Dependabot semanal sobre el módulo raíz, los
  cuatro módulos de herramienta y las acciones de los flujos.
- **Documentación fundacional**: `LICENSE` (Apache-2.0), `README.md`, `CONTRIBUTING.md` —ritual por hito,
  estructura de propuesta de cambio, Conventional Commits, versionado semántico, catálogo de controles y
  justificación obligatoria de toda dependencia nueva—, este `CHANGELOG.md` y los cuatro ADR fundacionales
  en `docs/ADR/`: multicall, SQLite sin cgo, CENDOJ no masivo y frontera humana.

*De H1 — el kernel de la línea de órdenes:*

- **Despacho multicall**: un solo ejecutable atiende a todos los applets. Manda el **nombre con el que se
  invoca**, de modo que un enlace simbólico `echo -> kitlegal` ejecuta el applet `echo` y le entrega
  íntegros los argumentos —`./echo boe` ejecuta `echo` con el argumento `boe`, y no el applet `boe`—; solo
  cuando ese nombre no está registrado —`kitlegal` entre ellos— el applet sale del **primer argumento**.
  Los verbos reservados del binario se reconocen antes que el registro, y por eso el registro rechaza al
  construirse un applet que se llame como uno de ellos. Una invocación que no resuelve ningún applet
  termina con código `2` nombrando lo que no reconoció y enumerando lo que existe; nunca con un pánico.
- **El registro es la única lista**: registrar un applet basta para que aparezca en `--help` y declarar un
  verbo basta para que se enumere, sin ninguna lista paralela mantenida a mano. Si el applet declara un
  verbo por omisión, invocarlo sin nombrar verbo lo usa; si no lo declara, nombrarlo es obligatorio y
  omitirlo termina con código `2`. La ayuda del binario, la de un applet y la de un verbo son texto para
  una persona y `--json` no las altera.
- **Ocho banderas globales idénticas en todos los applets**, declaradas una sola vez en el kernel y
  heredadas sin escribir ninguna: `--json`, `--timeout` (30 s por omisión, plazo de **toda** la operación
  y no de una petición suelta), `--offline`, `--dry-run`, `--describe`, `--no-graph`, `--asunto` y
  `--verbose`. Un applet declara su nombre, sus verbos y el contenido de `data`, y las recibe ya
  interpretadas. Tres —`--offline`, `--no-graph` y `--asunto`— fijan hoy solo su sintaxis y se propagan
  tal cual: su semántica llega con la caché (H3), con el grafo (H17) y con el asunto (H18), y no se
  inventa antes.
- **Códigos de salida estables**: `0` correcto, `2` argumentos inválidos, `3` no encontrado, `4` fuente no
  disponible, `5` límite de peticiones o términos de uso, `6` requiere identidad humana, y **`1` reservado
  al fallo inesperado** —lo que nadie declaró—, que es la convención de Unix para el error general y el
  único valor que un consumidor interpreta sin documentación. Un applet nombra la clase de su fallo con un
  sentinela y **nunca** un número; la traducción de clase a código ocurre en un único punto por binario,
  cubre las seis clases sin rama por defecto y el linter falla si alguien añade una clase y se olvida de
  darle código. Ninguna ruta de usuario termina en pánico.
- **Sobre de salida con huella reproducible**: toda invocación emite las mismas **seis** claves —`ok`,
  `fuente`, `url`, `fecha_consulta`, `hash` y `data`—, en éxito y en fallo, sin omitir ninguna aunque su
  valor sea el cero. `fecha_consulta` va en RFC 3339 con desplazamiento horario explícito; `hash` lleva
  delante el algoritmo que lo produjo —`sha256:`— seguido de los 64 dígitos hexadecimales del SHA-256 de
  la forma canónica de `data`, de modo que el mismo contenido dé siempre la misma huella con
  independencia del orden de las claves y del instante de la consulta, y que un solo byte distinto dé
  otra. Sin `fuente` y sin una `url` absoluta no se emite sobre, porque sin ellas no hay cita. Cuando `ok`
  es falso, `data` lleva exactamente `clase` y `mensaje`: el detalle técnico va al registro de eventos y
  no al sobre.
- **Dos formas de presentación y una sola salida estándar**: con `--json`, un único documento en una línea
  y nada más, ni siquiera con el registro de eventos al máximo detalle; sin `--json`, la **tabla mínima**,
  que escribe las cuatro líneas de procedencia —`fuente`, `url`, `fecha_consulta` y `hash`— y a
  continuación el contenido de `data` aplanado a pares ruta/valor, de modo que elegir la forma legible por
  una persona no pierda la cita. El registro de eventos va **siempre** a la salida de error, y su nivel se
  fija con `--verbose` o con la variable de entorno `KITLEGAL_LOG` (`debug`, `info`, `warn`, `error`); un
  valor fuera de esa lista se ignora, se avisa y no aborta la invocación.
- **Autodescripción**: `--describe` emite un esquema JSON con la entrada y la salida del verbo —las dos en
  un mismo documento, bajo `entrada` y `salida`— y excluye la ejecución, así que el applet nunca llega a
  ver esa invocación. Es la contraparte legible por máquina de `--help`, y de ella saldrán las tools MCP y
  la tabla de órdenes de cada `SKILL.md`.
- **Cinco dependencias directas nuevas**, todas de la lista cerrada de la constitución (§V): en el
  binario, `github.com/alecthomas/kong` (análisis de la línea de órdenes) e `github.com/invopop/jsonschema`
  (generación del esquema de `--describe`); solo en los tests, `github.com/stretchr/testify` (aserciones),
  `github.com/rogpeppe/go-internal` (los guiones `testscript` del e2e) y
  `github.com/santhosh-tekuri/jsonschema/v6` (validación del sobre contra su descripción formal). Con
  ellas aparece por primera vez un `go.sum` en la raíz del módulo. **El binario distribuido enlaza además
  los cuatro módulos que `invopop/jsonschema` arrastra** y que ninguna versión de esa biblioteca deja de
  traer: `github.com/pb33f/ordered-map/v2` (las propiedades del esquema conservan el orden de
  declaración), `github.com/bahlo/generic-list-go` y `github.com/buger/jsonparser` (dependencias de ese
  mapa) y `go.yaml.in/yaml/v4` (en versión candidata, `v4.0.0-rc.2`, fijada por aquel mapa; Dependabot la
  sigue). Están justificados en el plan del hito (*Complexity Tracking*) y en la propuesta de cambio, y
  un test de arquitectura (`TestDependenciasDelBinario`) fija la lista exacta de módulos que el binario
  enlaza, de modo que uno nuevo no entra sin justificarse por escrito. Los módulos que solo usan los
  tests —`go.yaml.in/yaml/v3` por `testify`, `golang.org/x/sys` y `golang.org/x/tools` por `testscript`,
  `golang.org/x/text` por el validador— no se enlazan en el binario.
- **La forma corta `-h` pide la ayuda en las mismas posiciones que `--help`**: en el binario, en un
  applet y en un verbo. La ayuda del verbo, que escribe el analizador, la anunciaba ya como `-h, --help`;
  ahora el kernel la reconoce también donde todavía no hay gramática que la lea.

*De H4 — el applet `boe`:*

- **Applet `boe`, el primero con fuente**, registrado en el binario distribuido: consulta la API de
  Legislación Consolidada del BOE (`https://www.boe.es/datosabiertos/api/legislacion-consolidada`) y firma
  cada sobre con `fuente` `boe.legislacion-consolidada` y la `url` de la API consultada. `kitlegal boe …` y
  el enlace `boe -> kitlegal` dan la misma salida byte a byte. Tiene **seis verbos**, ninguno por omisión
  —sin verbo termina con `2`— y ninguna bandera propia:
  - `buscar <texto>…` une las palabras y busca por título (`titulo:<p1> AND titulo:<p2> …`), o pasa la
    consulta tal cual si lleva operadores (` AND `, ` OR `, ` NOT `, `titulo:`, `materia:` o comillas); como
    mucho diez resultados, en el orden de la fuente, y `[]` si no hay ninguno.
  - `indice <norma>` devuelve los bloques de la norma en el orden de la fuente, con el tipo de cada uno.
  - `articulo <norma> <bloque>` devuelve el texto vigente del bloque, su `hash_texto`, su dirección pública
    y la `url_eli` de la norma, con los avisos de su vigencia (`codigo`, de un enumerado cerrado de tres
    valores, y `texto`), y nunca emite un texto cuya vigencia no ha podido comprobar. Es el porte de
    `refs/boe.py`: su `data` coincide con lo que calcula `boe.py` en los ocho campos del diff de aceptación,
    para seis artículos de cuatro normas.
  - `articulos <norma> <bloque>…` hace lo mismo con varios bloques, en el orden pedido y con repeticiones,
    pidiendo cada bloque distinto una sola vez y los metadatos como mucho una vez por invocación.
  - `metadatos <norma>` devuelve los datos de la norma y los avisos de su vigencia.
  - `analisis <norma>` devuelve sus materias, sus notas y sus referencias anteriores y posteriores, con el
    texto completo.

  Los identificadores se validan antes de abrir nada (`BOE-A-AAAA-N` para la norma; letras, dígitos, `.` y
  `-` para el bloque) y los fallos salen con los códigos estables: `2` argumentos inválidos; `3` lo que la
  fuente no tiene —un bloque o una norma inexistentes, o una respuesta vacía—; `4` la fuente caída tras los
  reintentos, un estado HTTP no previsto o una respuesta que ya no se sabe interpretar; y `5` el límite de
  peticiones o un `robots.txt` que deniega la ruta. Ninguno termina con `6`. El sobre de un fallo de la
  fuente lleva la `url` de la petición que falló.
- **Caché de las consultas y semántica de `--offline` y `--dry-run`.** `boe` guarda cada respuesta correcta
  en la caché local de H3 (`~/.cache/kitlegal/cache.db`, u otra carpeta con `KITLEGAL_CACHE_DIR`) con una
  vigencia por verbo —300 s para `buscar` y `metadatos`, 7 días para `indice`, `articulo`, `articulos` y
  `analisis`—; los metadatos con los que se comprueba la vigencia se comparten entre `articulo`, `articulos`
  y `metadatos`, y **ningún fallo se guarda**. Una consulta idéntica dentro de la vigencia no emite ninguna
  petición. `--offline` abre la caché en solo lectura y responde con lo vigente, o termina con `4` sin pedir
  nada; `--dry-run` tampoco pide nada ni crea la caché, y describe en la salida de error, una línea por
  petición y sin repetir ninguna, cada `GET` que habría emitido. `--no-graph` y `--asunto` siguen sin
  cambiar la salida.
- **Fila de la fuente en `docs/SOURCES.md`**, que nace con ella: licencia, términos de uso, resultado de la
  revisión del `robots.txt`, ritmo (`1s` entre dos peticiones al mismo sitio), formato y fecha de la revisión
  humana. El adaptador pide con ese ritmo y declara esos términos, y un test falla si la fila y sus
  constantes divergen.
- **Esquemas publicados `schemas/norma.json` y `schemas/bloque.json`**, los primeros contratos de salida
  versionados: `norma.json` describe `buscar`, `indice`, `metadatos` y `analisis`, y `bloque.json`,
  `articulo` y `articulos`, cada verbo bajo `$defs.<verbo>` con su `$id`
  (`https://ventanillalegal.es/schemas/<fichero>/<verbo>`) y la entrada y la salida que emite su
  `--describe`. Los tests validan contra el fichero —no contra lo que emite el binario en ese instante— el
  sobre real de éxito y los de fallo con códigos `2`, `3` y `4` de los seis verbos.
- **`make verify-sources` y la verificación nocturna contra la fuente real.** `make verify-sources` ejecuta
  `scripts/verify-sources.sh`, que pide a la API del BOE `boe articulo BOE-A-2015-10565 a21 --json` y exige
  código `0`, un sobre válido contra la parte `articulo` de `schemas/bloque.json` y un texto no vacío: forma
  y no contenido, porque el texto de un artículo cambia legítimamente con cada reforma. Es el único control
  que pide algo a una fuente real, así que **necesita red y no forma parte de `make ci`**; la misma
  comprobación, sin red, la ejerce `make test` sobre las grabaciones y sobre una respuesta que ya no se
  interpreta. El flujo `nightly` gana el trabajo `fuentes`, independiente del de `make ci` y con permiso para
  escribir incidencias: ejecuta `make verify-sources` y, si falla, comenta la incidencia abierta
  «verify-sources: boe articulo» o la abre. Ningún flujo graba respuestas: las grabaciones contra las que
  corren los tests las hace una persona con `scripts/grabar-fixtures.sh`.

*De H5 — la skill `boe-legislacion` y el andamiaje de skills:*

- **Skill `boe-legislacion`** (`skills/boe-legislacion/`), la primera del producto y genérica para cualquier
  materia: consulta y cita la normativa consolidada del BOE —procedimiento administrativo, contratación pública,
  régimen local, tributos, relaciones laborales…— con el applet `boe`. Su `SKILL.md`, de menos de 300 líneas, fija
  un protocolo de cinco pasos (identificar la norma en su referencia de normas, resolver `BOE-A-…` o buscarlo con
  `boe buscar`, leer el índice y los bloques, evaluar si falta contexto y responder citando), la forma de cita
  `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`, en la que lo que hace cita es que los corchetes terminen
  en el identificador y el bloque —con la forma legible delante del corchete o, dentro, delante del identificador—,
  también cuando la cita va sola, y sus reglas: ningún contenido legal que
  no salga del texto consultado, ley y reglamento distinguidos y la variación autonómica señalada. El id de cada
  bloque se copia de la entrada del índice, nunca se compone del número del artículo, y los bloques se leen de uno
  en uno; si una consulta de varios bloques falla, cada bloque se pide por separado antes de decir qué no se pudo
  consultar. Llama al binario por `scripts/boe`, un enlace al
  binario instalado, y no tiene ningún caso especial de un municipio ni de una comunidad.
- **Tabla de normas `data/normas.yaml`**, única fuente de verdad de las normas que referencian las skills: diez
  normas, cada una con su identificador `BOE-A-…`, su título, su rango, sus materias y, si la tiene, su
  abreviatura. De ella se genera `references/normas.md` de la skill, con la cabecera que prohíbe editarlo.
- **Esquemas de datos `schemas/normas.yaml.json` y `schemas/eval.yaml.json`** (JSON Schema 2020-12), contra los
  que se validan la tabla de normas y cada eval; no son contratos de salida de ningún applet. Los lee un lector
  YAML común que rechaza una clave escrita dos veces en el mismo mapa nombrando sus dos líneas, en lugar de
  quedarse en silencio con el último valor. Para ello entra `go.yaml.in/yaml/v3` como dependencia directa, que
  solo usan las herramientas de desarrollo y que el binario no enlaza.
- **Formato común de eval**: las evals de cada skill en su propio directorio, `evals/<skill>/`, con un fichero
  YAML por eval (`<nn>-<descripción>.yaml`) y los campos `pregunta`, `activa`, `comandos` y `citas` —estos dos,
  obligatorios en una eval que debe activar la skill y prohibidos en una que no— y los opcionales `reproduce` e
  `informativa`, que marca la eval que se mide y se publica sin decidir el veredicto.
- **Evals de `boe-legislacion`** (`evals/boe-legislacion/`), diecisiete, escritas en el formato común de eval y antes
  que la skill: diez preguntas de materias distintas que deben activarla, hacer las consultas esperadas y citar el
  bloque esperado —la del IRPF reproduce un uso documentado de `boe-fiscal` y lo declara con
  `reproduce: boe-fiscal`—, dos preguntas ajenas que no deben activarla y cinco preguntas por materia, sin el número
  del artículo, marcadas `informativa: true`: se ejecutan y su tasa se publica, pero no deciden el veredicto mientras
  siga en el backlog la herramienta que busca dentro de una norma el artículo que trata una materia (ADR 0016). Las
  respuestas del BOE que necesitan las graba una persona con `scripts/grabar-evals.sh`.
- **Job de evals**: el flujo `evals` (`.github/workflows/evals.yml`) ejecuta `make evals SKILL=boe-legislacion` a
  mano, al abrirse o reabrirse una propuesta de cambio que toque la skill, sus datos, sus evals, el applet `boe`, el
  kernel, el arnés de las evals, el `Makefile` o el propio job, y al poner la etiqueta `evals` o `evals-prueba-de-red`
  en cualquiera —sin ejecución programada y sin relanzarse al empujar a una propuesta abierta—, con Claude Code
  `2.1.270`, el secreto `CLAUDE_CODE_OAUTH_TOKEN` y un runner del que antes retira todo Python. Decide con
  `claude-sonnet-5`, el modelo del uso real de la skill, y ejecuta además `claude-haiku-4-5-20251001` como límite
  inferior que se publica sin decidir; cada eval se abre tres veces con `claude-sonnet-5` y, si no es informativa,
  otras tres con `claude-haiku-4-5-20251001`, y cada una de esas series pasa con dos (ADR 0016). El
  informe, con el veredicto global, la tasa de cada eval y lo que se comprobó de cada sesión, se imprime en el registro
  entre marcas.
- **`make skills-sync` real**: deja de anunciar que las skills llegan en H5 y regenera, desde `data/*.yaml` y desde
  `--describe` del binario, `references/`, la tabla de comandos de `SKILL.md` y los enlaces de `scripts/` de cada
  skill; dos ejecuciones seguidas no cambian nada.
- **`make skills-check`**, dentro de `make ci`: regenera todo eso en memoria y lo compara con el árbol sin escribir
  nada, y comprueba el frontmatter y el límite de líneas de cada `SKILL.md`, la tabla de normas contra su esquema y
  cada identificador contra la búsqueda grabada del BOE, y el formato y el conjunto de las evals y que lo que
  necesitan está grabado. Falla nombrando la skill y el fichero o el enlace, sin red y sin modelo.
- **`make evals`** (`scripts/evals.sh <skill>`), fuera de `make ci`: una sesión de Claude Code por cada eval, cada
  modelo y cada repetición que pide el plan, bajo `strace` y con la red cerrada salvo la del modelo, con la skill
  instalada por `make install` y el binario respondiendo desde una caché preparada con lo grabado; juzga cada sesión sin
  modelo, agrupa las de cada eval y cada modelo en una serie con su tasa y escribe el informe. Necesita Linux con
  `strace`, root o `sudo` y ningún Python accesible, y falla antes de la primera sesión si falta algo o si una eval está
  mal formada.

*De H5.1 — los avisos de vigencia en las evals:*

- **Forma fija de los avisos de vigencia en `boe-legislacion`**: su `SKILL.md` fija cómo traslada la respuesta cada
  aviso del sobre, con `⚠`, la etiqueta del aviso tal como la da el binario y dos puntos, seguidos de la frase del
  binario o de una explicación —`⚠ NORMA DEROGADA:` para `derogada`, `⚠ VIGENCIA AGOTADA:` para `vigencia-agotada` y
  `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:` para `consolidacion-no-finalizada`—, con la etiqueta entera y en la misma
  línea; decir con otras palabras que la norma está derogada no traslada el aviso. La etiqueta de cada código tiene una
  sola fuente de verdad, el applet `boe`, cuya salida no cambia en un byte, y `make skills-check` falla nombrando el
  código si el `SKILL.md` pierde la forma fija de alguno.
- **Campo `avisos` en el formato común de eval**: opcional y solo en una eval que activa la skill, lista los códigos
  de aviso de vigencia del binario —`consolidacion-no-finalizada`, `derogada` o `vigencia-agotada`— cuya forma fija
  tiene que llevar la respuesta. `schemas/eval.yaml.json` rechaza un código desconocido, una lista vacía y `avisos` en
  una eval de no activación, y `make skills-check` comprueba que el esquema admite exactamente los códigos del binario.
  Una eval con `avisos` solo pasa si la respuesta lleva la forma fija de cada uno: se juzga sin modelo, con el selector
  de variante tras `⚠`, el énfasis de Markdown, otros blancos y las minúsculas tolerados, y otra redacción o la negación
  dejan el aviso ausente. Las evals sin `avisos` se juzgan igual que antes.
- **Reparto de avisos en el informe**: cada sesión publica en `informe.json` `avisos_encontrados` y `avisos_ausentes`,
  en el orden de la eval —`[]` si no espera ninguno—, y cada aviso ausente da el motivo `aviso ausente: <código>`,
  detrás de los de las citas ausentes; la tabla de sesiones de `informe.md` gana las columnas «Avisos encontrados» y
  «Avisos ausentes». Un rojo por un aviso ausente se distingue así de uno por una cita ausente.
- **Eval informativa de una norma derogada** (`evals/boe-legislacion/18-lrjpac-norma-derogada.yaml`): pregunta por el
  artículo 42 de la Ley 30/1992 sin decir nada de su vigencia y exige consultar y citar su bloque `a42` y trasladar los
  avisos `derogada` y `vigencia-agotada` con su forma fija. Nace `informativa: true` (ADR 0016): se ejecuta y su tasa
  se publica sin decidir el veredicto. La Ley 30/1992 entra en `data/normas.yaml`, sin ninguna marca de derogación, y
  su índice lo graba una persona con `scripts/grabar-evals.sh`.

*De H6 — el applet `territorio` y la skill `legal-core`:*

- **Applet `territorio`, con un solo verbo, `resolver`**, registrado en el binario distribuido junto a `boe`:
  `kitlegal territorio resolver <consulta>` —o `territorio resolver <consulta>` por el enlace— resuelve un municipio
  de España, por su nombre o por su código INE de cinco cifras o de seis con el dígito de control, a su territorio.
  `data` lleva siempre las mismas ocho claves —`municipio`, `codigo_ine`, `provincia`, `comunidad`, `dir3`,
  `regimen`, `boletines` y `cobertura`—, ninguna omitida y cada dato con su `source`: la fila de `docs/SOURCES.md` o
  el fichero de `data/territorio/` del que sale. `boletines` trae siempre el BOE y, además, solo los boletines que la
  comunidad tiene configurados; `cobertura` dice si el boletín autonómico y el provincial están `configurado` o
  `no-configurado` y si el DIR3 del ayuntamiento está `verificado` o `no-verificado`, y ninguno de sus valores
  significa «no existe». Fuera del territorio configurado la salida no nombra ningún boletín que no tenga, y un DIR3
  sin verificar va vacío, nunca calculado. `regimen` marca `comun` o `foral` en todas las comunidades, estén
  configuradas o no. No pide nada a la red ni toca la caché: el sobre lleva `fuente` `kitlegal.territorio`, `url`
  `kitlegal:applet/territorio` y, como `fecha_consulta`, la fecha más antigua de los ficheros que sostienen la
  respuesta, de modo que la misma consulta da la misma salida byte a byte, con `--offline` o sin él y desde cualquier
  directorio. Códigos: `0` resuelto; `2` una entrada que no llega a ser un código —provincia fuera de `01`-`52`,
  municipio `000`— o un dígito de control que no es el oficial, y un nombre que corresponde a más de un municipio,
  cuyo mensaje enumera todos los candidatos como `<código INE> <nombre> (<provincia>)`; `3` un nombre, o un código
  bien formado, que no está en la relación. El applet no decide nunca `4`, `5` ni `6`; solo el plazo de `--timeout`,
  que pone el kernel para todo applet, termina en `4`. Su contrato se publica en `schemas/municipio.json`
  (`$defs.resolver`), generado desde `--describe` y comprobado por `make schema-check` como los de `boe`.
- **Datos congelados de territorio en `data/territorio/`**, versionados y embebidos en el binario, que no los pide
  en red en ningún momento (ADR 0017): `municipios.yaml`, la relación de municipios del INE con su dígito de control,
  su provincia y su comunidad (`source` `ine.municipios`); `dir3.yaml`, el DIR3 del ayuntamiento de cada municipio,
  `L` más su número de inscripción en el Registro de Entidades Locales, solo con los municipios cuyo número es coherente
  con su código INE y su dígito de control (`source` `mpt.rel`); la regla se verificó contra el directorio DIR3 del
  Punto de Acceso General en una muestra con municipios fusionados, forales y con entidades locales menores; `estado.yaml`, el boletín estatal; y `comunidades/`, un fichero por cada una de las 19 comunidades y
  ciudades autónomas con su régimen y sus provincias, cuyos nombres salen de las tablas de códigos del INE (`source`
  `ine.codigos-territoriales`). Solo la Comunidad de Madrid trae `boletines`: el BOCM, a la vez
  autonómico y provincial, con el motivo escrito en su fichero. Añadir un territorio es rellenar los boletines de su
  fichero, sin tocar código ni skills. Una persona los genera fuera del repositorio desde las descargas del INE y del
  REL; `docs/SOURCES.md` lleva la fila de cada origen con la fecha del fichero y `scripts/verify-sources.sh` no gana
  ningún caso, porque no se consultan. Se validan contra `schemas/territorio-municipios.yaml.json`,
  `schemas/territorio-dir3.yaml.json`, `schemas/territorio-estado.yaml.json` y
  `schemas/territorio-comunidad.yaml.json`, y `make skills-check` comprueba además su integridad —toda provincia
  declarada por una sola comunidad, todo DIR3 coherente con el código y el dígito de su municipio—, la procedencia de
  cada dato y que todo municipio se alcance por su nombre oficial, resuelto o entre los candidatos de un nombre
  ambiguo.
- **Jerarquía normativa como dato**, `data/jerarquia.yaml`, validada contra `schemas/jerarquia.yaml.json` dentro de
  `make skills-check`: los cinco niveles en su orden —Unión Europea, Estado, comunidad autónoma, provincia y
  municipio—, con la clase de boletín que publica las normas de cada uno y sus tipos de norma de mayor a menor rango,
  y las cuatro reglas de interpretación: competencia antes que jerarquía, ley posterior, ley especial y reglamento
  nunca contra ley. No lleva el identificador ni el texto de ninguna norma.
- **Las leyes vertebrales, con su identificador verificado**: `data/normas.yaml` gana siete normas, con el
  identificador, el título y el rango copiados de su búsqueda grabada del BOE, y la marca opcional `vertebral: true`
  en las quince de la tabla de leyes vertebrales del mapa del sistema legal, y en ninguna más. `schemas/normas.yaml.json`
  gana esa propiedad y, en el enumerado de rangos, los que traen las búsquedas nuevas. `references/normas.md` de
  `boe-legislacion` se regenera con ellas; su `SKILL.md` y sus evals no cambian.
- **Skill `legal-core` v0** (`skills/legal-core/`), la skill madre y el punto de partida de las preguntas de derecho
  público que dependen de un municipio. Su `SKILL.md`, de menos de 300 líneas, fija un protocolo que empieza por
  identificar el territorio —y pregunta el municipio si la conversación no lo dice, sin suponerlo—, lo resuelve con
  `scripts/territorio resolver … --json`, traslada la `cobertura` a la respuesta sin nombrar ningún boletín que el
  applet no haya devuelto, ofrece todos los candidatos de un nombre ambiguo, no concluye «no existe» de lo que no
  encuentra, razona con sus dos referencias y delega en `boe-legislacion`, en un solo sentido, el texto de cualquier
  artículo. Sus dos referencias se generan con `make skills-sync` y llevan la cabecera que prohíbe editarlas:
  `references/leyes_vertebrales.md`, desde las normas marcadas `vertebral` de `data/normas.yaml`, y
  `references/jerarquia_normativa.md`, desde `data/jerarquia.yaml`. Llama al binario por `scripts/territorio`, un
  enlace al binario instalado, y `make install` la enlaza junto a `boe-legislacion`.
- **Identificadores código INE y DIR3** (`internal/core/ids`): el código INE se analiza con cinco cifras —provincia
  `01`-`52` y municipio `001`-`999`— o con seis, la última el dígito de control, que se compara con el oficial
  nombrando los dos; el DIR3 del ayuntamiento, con la forma `L01PPMMMD`, se analiza con la letra en mayúscula o en
  minúscula y se normaliza a mayúscula, y se compone desde el código y su dígito. Cada error nombra la entrada y dice
  qué tiene de malo, con la clase de argumentos inválidos y nunca con un pánico; `make test` ejecuta sus dos objetivos
  de fuzz sobre un corpus semilla versionado. Los patrones de identificador de los esquemas de territorio aceptan
  exactamente lo mismo que los analizadores, y `make skills-check` lo comprueba.
- **Variante de territorio del formato común de eval**: `comandos` admite una cuarta forma, `applet`, `verbo`
  `resolver` y `municipio`, y la eval gana el esperado opcional `territorio` —`comunidad`, `provincia`, los códigos de
  `boletines` y los aspectos de `cobertura` en la forma `<aspecto>: <valor>` del vocabulario del applet—; una eval que
  activa la skill lleva `citas`, `territorio` o los dos. Se juzga sin modelo: la comunidad y la provincia sin
  distinguir mayúsculas ni tildes, el código de cada boletín como palabra exacta y cada aspecto de cobertura en su
  forma fija, con los blancos y las mayúsculas tolerados; un comando de territorio no necesita nada grabado. El
  informe gana las columnas «Territorio encontrado» y «Territorio ausente», y `make skills-check` comprueba que el
  enumerado de cobertura del esquema y el vocabulario del applet dicen lo mismo. Las evals de `boe-legislacion` siguen
  validando sin cambiar un byte.
- **Evals de `legal-core`** (`evals/legal-core/`), tres, escritas antes que la skill y que deciden todas: qué
  comunidad, provincia y boletines corresponden a un ayuntamiento, sobre un municipio de la Comunidad de Madrid —con
  su comunidad, su provincia y el BOCM en lo esperado— y sobre uno de una comunidad sin configuración —con su
  comunidad, su provincia y los dos boletines declarados no configurados—, y una pregunta ajena que no debe activarla.
  Ninguna lleva citas, porque la skill no afirma el contenido de ninguna norma. `make skills-check` exige de su
  conjunto al menos tres evals, la del municipio configurado, la del no configurado, una de no activación y un
  esperado verificable en toda eval que activa la skill.

*De H19 — instalar sin clonar:*

- **Las skills viajan dentro del binario**: `skills.go`, en la raíz del módulo, empotra el `SKILL.md` y todo
  `references/` de cada skill del repositorio —hoy `boe-legislacion` y `legal-core`—, y es lo único que se instala.
  También un `go install ./cmd/kitlegal` sin inyecciones las lleva y las instala igual. El binario no enlaza ningún
  módulo nuevo.
- **Applet `skills`, con tres verbos y ninguno por omisión**, registrado en el binario distribuido junto a `boe` y
  `territorio`; sin verbo termina con `2`. **`kitlegal skills install [skill…]`** instala las skills empotradas —todas,
  o solo las nombradas— en el directorio neutro del ámbito: `.agents/skills/` del directorio de trabajo por omisión,
  `~/.agents/skills/` con `-g` o la ruta de `--dir`. En los dos primeros, si existe `.claude/` o se pide
  `--host claude`, enlaza cada skill en `.claude/skills/<skill>` con el destino literal relativo
  `../../.agents/skills/<skill>`; donde el sistema no deja crear enlaces, la entrada de host es una copia anotada
  `copia`. Escribe el manifiesto del ámbito, `kitlegal.json`, determinista y sin rutas absolutas, fechas ni usuarios:
  la versión del binario y, por skill, su versión, la huella SHA-256 de cada fichero y sus entradas de host con su
  modo. Solo toca lo que ese manifiesto declara: antes de escribir nada busca los conflictos —carpeta ajena, fichero,
  enlace a otro sitio, enlace roto, fichero editado, fichero ajeno, ruta que no es directorio, manifiesto ilegible y
  manifiesto con entradas de host— y, con uno solo, no crea ni cambia nada y termina con `1`, con la cabecera
  `skills install: nada se ha creado ni cambiado; conflictos:` y una línea `<clase>: <ruta>` por conflicto, en el
  `mensaje` del sobre y en la salida de error. Nunca lee, escribe ni retira a través de un enlace simbólico por debajo
  del ámbito, ni abre lo que no es un fichero regular. Cada skill sale `instalada`, `actualizada` o `sin cambios`:
  repetir la orden con el mismo binario deja el disco byte a byte igual, y con otro sustituye lo instalado y retira lo
  que el binario nuevo ya no empotra; una skill que el manifiesto declara y el binario no empotra se conserva sin
  tocar. Si la escritura falla a mitad, repetir la orden la completa. `--dry-run` no toca el disco y termina con el
  mismo código que la orden real, con una línea por skill en la salida de error.
- **`kitlegal skills list` y `kitlegal skills doctor`**, que no cambian nada en disco. `list` enumera lo que declara el
  manifiesto del ámbito —el directorio, su versión y, por skill, su versión, si el binario la empotra y sus
  enlaces—. `doctor` comprueba lo instalado contra el manifiesto y da hallazgos de cinco clases —fichero editado,
  enlace colgando, enlace a otro sitio, copia y versión distinta—, cada uno como `<clase>: <ruta>: <orden>`, con una
  orden de shell POSIX de una línea que lo arregla (`rm -- '<ruta>' && kitlegal skills install <skill> …`, con `-r`
  solo sobre un directorio real y nunca `-f`, y la ruta de `--dir` como `--dir '<ruta>'` o, si empieza por `-`,
  `--dir='<ruta>'`, para que no se lea como otra bandera), en un orden en que ejecutarlas una tras otra deja el
  siguiente `doctor` sin hallazgos; con alguno termina con `1`. Sin manifiesto, los dos terminan con `0`, con versión nula y lista vacía;
  ante un manifiesto ilegible o una ruta del ámbito que no es un directorio, con `1`.
- **Validación de la invocación antes de mirar el disco**: `-g` con `--dir`, `--host` con `--dir`, `--host` distinto
  de `claude` y una skill que el binario no lleva —con las disponibles en el mensaje— terminan con `2`, en ese orden de
  precedencia; `-g` sin `HOME` termina con `1`. El applet no pide nada a la red, no emite operaciones de grafo y firma
  su sobre con `fuente` `kitlegal.skills` y `url` `kitlegal:applet/skills`. Su contrato se publica en
  `schemas/instalacion.json` (`$defs.install`, `$defs.list` y `$defs.doctor`), generado desde `--describe` y
  comprobado por `make schema-check`.
- **Aviso de versión, sin red.** Toda orden de otro applet resuelta con un verbo —también si termina con un error de
  argumentos del verbo, con `--dry-run` o con `--describe`— busca, sin seguir enlaces, el manifiesto del directorio de
  trabajo y, si no lo hay, el de la cuenta. Si su versión, o la de alguna skill declarada que el binario empotra, es
  distinta de la del binario —quitando a cada una una `v` inicial—, escribe exactamente una línea en la salida de error:
  `aviso: las skills instaladas son de kitlegal <instalada> y este binario es kitlegal <binario>; ejecuta: kitlegal
  skills install`, con ` -g` si el manifiesto es el de la cuenta. No cambia la salida estándar ni el código de salida,
  y un fallo al escribirla no se propaga. No avisan `skills`, `version`, la ayuda, los fallos anteriores a resolver el
  applet con su verbo ni un binario de desarrollo, cuya versión no tiene forma SemVer.
- **La release**: `.goreleaser.yaml`, con goreleaser v2.18.1 como módulo de herramienta (`tools/goreleaser/`), construye
  darwin, linux y windows en amd64 y arm64, sin cgo, con `-trimpath` y con las mismas cuatro inyecciones que el
  `Makefile` —en una etiqueta, la versión es la etiqueta tal cual—; archivos `kitlegal_<os>_<arch>.tar.gz` (`.zip` en
  Windows), `checksums.txt`, que incluye `install.sh`, un SBOM por archivo (syft), la firma sin clave de
  `checksums.txt` (cosign), notas de release agrupadas por Conventional Commits, el cask del tap
  `jmorenobl/homebrew-tap` para `brew install jmorenobl/tap/kitlegal` —con el gancho que retira la cuarentena de
  macOS—, el manifiesto del bucket `jmorenobl/scoop-bucket` para `scoop install kitlegal`, y paquetes `.deb` y `.rpm`.
  `make goreleaser-check` la valida dentro de `make ci`; `make release` construye el snapshot en `dist/` sin publicar,
  firmar ni generar SBOM; y `make snapshot-check` lo comprueba (`TestSnapshot`) y ejecuta contra él los guiones del
  instalador. `TestConfiguracionDeLaRelease` fija la configuración y los flujos.
- **Flujo `release`** (`.github/workflows/release.yml`), solo al empujar una etiqueta `v*`: publica con
  `PUBLISHER_TOKEN` —el único secreto, visible solo en el paso que publica—, atesta la procedencia de los seis
  archivos y de `checksums.txt` y comprueba lo publicado en un trabajo de humo: la huella, `gh attestation verify`,
  que `version` imprime la etiqueta, que `boe articulo … --offline` con la caché vacía termina con `4` y que
  `skills install` deja `boe-legislacion` en un directorio vacío. Y el trabajo **`snapshot`** del flujo `ci`, en cada
  propuesta de cambio y en cada push a `main`, sin `id-token` ni secretos: `make release` y `make snapshot-check`.
- **`scripts/install.sh`**, el instalador POSIX `sh` de macOS y Linux, sin Go, git ni gestor de paquetes:
  `curl -fsSL https://raw.githubusercontent.com/jmorenobl/kitlegal/main/scripts/install.sh | sh`, o con una versión
  (`sh -s -- 0.1.0`, con o sin `v`). Detecta el sistema y la arquitectura, descarga el archivo y `checksums.txt` de la
  release, verifica la huella en la línea cuyo segundo campo es exactamente el archivo, deja solo el binario en
  `$KITLEGAL_INSTALL_DIR` o `~/.local/bin` renombrándolo encima del anterior, no toca ningún fichero de arranque del
  shell, imprime la línea `export PATH=…` si el directorio no está en el `PATH` y termina con
  `kitlegal skills install`. Todo error sale por la salida de error con el prefijo `install.sh: `, un código distinto
  de `0` y nada instalado, también con el `/bin/sh` de macOS en un locale UTF-8. Todo lo que hace está en funciones y
  la llamada va en la última línea, así que un guion que llega cortado no ejecuta nada a medias y, si solo le falta
  esa llamada, termina con `1`. Se sirve desde `main` y va adjunto a cada release; los guiones `instalador-` del e2e lo
  prueban contra un origen local y contra el snapshot, sin red, y `TestInstaladorEnUTF8` y `TestInstaladorCortado`,
  sus rechazos en UTF-8 y sus cortes.

### Cambiado

*De H1 — el kernel de la línea de órdenes:*

- **El contrato observable del binario deja de ser el de H0.** De aquel se conservan las tres líneas de
  `version` y su código `0` —ahora escritas por el presentador, y un fallo al escribirlas se propaga en
  lugar de descartarse— y el código `2` de cualquier otra invocación; lo que cambia es el mensaje y lo
  que hay detrás. `version` sigue sin admitir argumentos ni banderas: `kitlegal version extra` o
  `kitlegal version --json` terminan con `2`, como en H0, pero el mensaje nombra lo que sobra en lugar de
  la línea `uso: kitlegal version`. Un applet desconocido lo resuelve el despacho multicall, que nombra lo
  que no reconoció y enumera lo registrado. `--help` responde la ayuda derivada del registro y termina
  con `0`, donde H0 terminaba con `2` por no ser `version`; y los códigos de salida posibles ya no son
  solo `0` y `2`, sino los siete de la tabla estable.
- **Un descriptor de salida roto termina siempre con `1`**, sea lo que sea lo que se estaba escribiendo:
  el sobre, la tabla, el esquema de `--describe`, `version` o cualquiera de las tres ayudas —también la
  del verbo, que escribe el analizador y que sin vigilar el escritor habría salido con el `2` de
  argumentos inválidos—. Vale también para una tubería del sistema cuyo lector ha terminado
  (`kitlegal … | head -1`): el binario ignora la señal `SIGPIPE` con la que el runtime de Go mataría el
  proceso sin código ni mensaje, de modo que la escritura fallida se traduce como cualquier otro fallo,
  con el mensaje en la salida de error y el código `1`, y nunca con la muerte por señal (`141` en el
  intérprete de órdenes).
- **La exclusión de `errcheck` desaparece.** H0 eximía `fmt.Fprint`, `fmt.Fprintf` y `fmt.Fprintln` porque
  el punto de entrada escribía él mismo y su contrato de códigos de salida no tenía dónde poner un fallo
  de escritura. Ya no escribe —inyecta los descriptores del sistema, el registro de producción y los datos
  de construcción, y termina con el código que le devuelve la raíz de composición—, así que ninguna
  función queda exenta y todo error de escritura se comprueba.
- **`make test-e2e` deja de anunciar el hito ausente** y ejecuta los guiones `testscript` que describen la
  entrega, contra el binario que el propio test construye.

*De H3 — la caché local en SQLite:*

- **`make ci` ejecuta también los tests de integración.** La lista de prerrequisitos de `ci` gana
  `test-integration` —`go test -race -tags=integration -coverprofile=coverage-integration.out ./...`, la receta de H0 más el perfil de cobertura—
  entre `test` y `vuln`, de modo que un test etiquetado `integration` en rojo, o un fichero etiquetado que
  no compile, hacen fallar el veredicto en local y en la integración continua por igual. Los primeros
  tests con esa etiqueta son los de la caché: los que dependen del entorno —permisos del sistema de
  ficheros y dos procesos— y trabajan solo dentro de directorios temporales. El análisis estático alcanza
  también los ficheros etiquetados (`run.build-tags: [integration]` en `.golangci.yml`), de modo que
  `sqlclosecheck` y `rowserrcheck` los vigilan. Con ello `make ci` encadena nueve controles: los ocho de
  H0 y este. Nada más cambia a la vista: ningún applet, verbo ni bandera nueva, y el binario distribuido
  no enlaza todavía la caché ni el controlador de SQLite.
- **La cobertura que publica CI incluye los tests de integración.** `ci.yml` sube los dos perfiles
  (`coverage.out` y `coverage-integration.out`) y Codecov los une, de modo que las ramas que solo
  ejercitan los tests etiquetados —permisos del sistema de ficheros, dos procesos— cuentan como lo que
  son: código con test. `codecov.yml` declara además el estado `patch` (`target: auto`, bloqueante), que
  hasta ahora regía sin declarar: la regla «no retroceder» aplicada al código nuevo de cada propuesta.

*De H4 — el applet `boe`:*

- **`make schema-check` deja de ser un aviso.** Ejecuta `TestEsquemasPublicados`
  (`go test -count=1 -run '^TestEsquemasPublicados$' ./internal/app/`), que regenera en memoria, desde
  `--describe` de los verbos registrados en el binario distribuido, la forma canónica de
  `schemas/norma.json` y de `schemas/bloque.json` y la compara con los ficheros versionados, sin escribir
  nada; falla nombrando el fichero y el verbo que difieren, o el fichero que no es la serialización canónica
  de sus partes. Sigue dentro de `make ci`, que encadena los mismos nueve controles. Los ficheros solo se
  regeneran a propósito, con la bandera `-actualizar-esquemas` del mismo test.
- **`fecha_consulta` la declara la fuente, también en lo servido desde la caché** (ADR 0015). Hasta H3 el
  kernel fechaba todo sobre con su reloj al montarlo, lo que habría hecho pasar por recién consultado un
  artículo guardado hace seis días. `schema.Procedencia` gana `FechaConsulta`, y el montador fecha el sobre
  con ella cuando la procedencia es válida y la fecha no es cero, en éxito y en fallo; con el valor cero, o
  con una procedencia inválida, lo fecha con su reloj como antes, así que un applet que no declara fecha no
  cambia. `boe` declara el instante en que `internal/httpx` emitió la petición —el de su último intento— o,
  si lo que responde sale de la caché, el instante guardado en la entrada; si `data` se sostiene sobre varias
  consultas, el más antiguo, y en un fallo, el de la petición que falló. La huella no cambia: se sigue
  calculando sobre `data`, que no lleva la fecha.
- **`app.Arrancar` es la raíz de arranque de los dos binarios.** Construye el registro y, si se construye,
  atiende la invocación con `Main`; un registro que no se construye es un defecto de composición y sale como
  sobre del kernel de clase «inesperado», con el mensaje en la salida de error y código `1`, nunca como un
  pánico. El binario distribuido termina con `os.Exit(app.Arrancar(os.Args, app.RegistroDeProduccion, …))`:
  su ayuda enumera `boe` y el fallo de un applet desconocido termina en `applets disponibles: boe`, en lugar
  de decir que el binario no registra ninguno. El binario de extremo a extremo deja su `panic` y arranca por
  el mismo camino con `echo`, `contar` y `boe` —este, sobre la reproducción de sus grabaciones y sin red—, y
  `make test-e2e` recorre con él los seis verbos, el enlace, `--offline`, los códigos y la respuesta servida
  desde la caché por debajo de 200 ms.
- **El binario distribuido enlaza la red y la caché.** `go.mod` no gana ninguna entrada, pero al registrar
  `boe` el binario enlaza `internal/httpx` e `internal/cache` y, con ellos, doce módulos de terceros:
  `github.com/temoto/robotstxt` y `golang.org/x/time` —de la lista cerrada de la constitución, por la red—,
  `modernc.org/sqlite` —ídem, por la caché— y los nueve que ese controlador arrastra (`modernc.org/libc`,
  `modernc.org/mathutil`, `modernc.org/memory`, `github.com/remyoudompheng/bigfft`,
  `github.com/dustin/go-humanize`, `github.com/google/uuid`, `github.com/mattn/go-isatty`,
  `github.com/ncruces/go-strftime` y `golang.org/x/sys`). `TestDependenciasDelBinario` fija la lista de
  dieciocho, cada uno justificado en su línea, y se retiran los tests que exigían que el binario no enlazara
  la red ni la caché, ciertos solo mientras ningún applet las usaba.

*De H5 — la skill `boe-legislacion` y el andamiaje de skills:*

- **`make install` enlaza las skills.** Tras el `go install` de H0, sin cambios, `scripts/instalar-skills.sh` enlaza
  cada skill de `skills/` en `~/.claude/skills/`, el directorio personal de skills de Claude Code, y deja
  `bin/instalado/kitlegal` apuntando al binario instalado, que es lo que alcanza `scripts/boe` de la skill. Repetirla
  deja el mismo estado. Antes del `go install`, el mismo guion con `--comprobar` busca las entradas del directorio
  personal con el nombre de una skill que no son su enlace —un directorio, un fichero, un enlace a otro sitio o un
  enlace roto—: escribe una línea de conflicto por cada una y falla sin crear ni cambiar nada, tampoco el binario.
  `make test-integration` lo prueba (`TestInstalacion`) sobre una copia mínima del árbol, con el directorio personal,
  el de binarios y el `GOPATH` temporales y sin red.
- **`make ci` encadena diez controles**: `skills-check` entra tras `schema-check`, de modo que una skill con una
  deriva o un defecto, una tabla de normas inválida o una eval mal formada hacen fallar el veredicto en local y en la
  integración continua por igual. El análisis estático alcanza también los ficheros con la etiqueta de compilación
  `evals` (`run.build-tags` de `.golangci.yml`), los arneses que usa el job de evals.

*De H6 — el applet `territorio` y la skill `legal-core`:*

- **El job de evals mide las dos skills en la misma ejecución.** El flujo `evals` deja de evaluar solo
  `boe-legislacion` y pasa a una matriz de skills, `boe-legislacion` y `legal-core`: un trabajo por skill, cada uno
  con su informe, y el rojo de uno no cancela el otro (`fail-fast: false`). La prueba de red sigue siendo solo de
  `boe-legislacion`, cuyo texto invoca el applet `boe`: la entrada `prueba_de_red` y la etiqueta
  `evals-prueba-de-red` la añaden a su trabajo y no al de `legal-core`, porque `territorio` no puede pedir nada a la
  red. Al abrirse o reabrirse una propuesta de cambio, el job arranca también si toca los paquetes de dominio
  (`internal/core/`); los ficheros congelados de `data/territorio/` ya los cubría `data/`. El umbral no cambia: cada
  serie de tres sesiones pasa con dos, y el informe lleva el commit evaluado y el identificador del modelo.

*De H19 — instalar sin clonar:*

- **Las skills invocan `kitlegal` desde el `PATH`.** La tabla de comandos de cada `SKILL.md`, que regenera
  `make skills-sync`, titula cada applet `kitlegal <applet>` y escribe cada orden `kitlegal <applet> <verbo> …`; en el
  texto libre de `boe-legislacion` y `legal-core`, `scripts/boe` y `scripts/territorio` pasan a `kitlegal boe` y
  `kitlegal territorio`, y las frases que decían de dónde sale el binario dicen ahora que se invoca desde el `PATH`.
  Protocolo, reglas, forma de la cita y de los avisos, y evals, sin cambios.
- **`make install` es el bucle de desarrollo**, no la forma de instalar: `go install` con las inyecciones del
  `Makefile` y, con ese binario, `kitlegal skills install -g --host claude`, que deja las skills en `~/.agents/skills/`,
  con su manifiesto, y enlazadas en `~/.claude/skills/<skill>` con destino `../../.agents/skills/<skill>`. Ya no
  comprueba nada antes de `go install`: un conflicto lo da `skills install`, que termina con `1` sin cambiar nada,
  con el binario ya instalado. Los enlaces absolutos que dejaba el `make install` anterior son un conflicto («enlace a
  otro sitio»), y `CONTRIBUTING.md` da el paso único que los retira. `TestInstalacion` y sus cinco guiones —los cuatro
  de `internal/skills/testdata/script/` y el del enlace roto, que escribe el propio test— lo comprueban con `HOME`,
  `GOBIN` y `GOPATH` temporales.
- **`make skills-sync` y `make skills-check` dejan los enlaces**: ni los generan ni los comprueban; una skill con
  `scripts/` es un defecto que hace fallar a los dos (ADR 0019), y `make skills-check` comprueba además que cada orden
  de la tabla de comandos de cada skill empotrada nombra un applet y un verbo del binario
  (`TestOrdenesDeLasSkillsEmpotradas`).
- **`make ci` encadena once controles**: `goreleaser-check` entra tras `skills-check`.
- **El job de evals instala con `make install`** y añade al `PATH` de cada trabajo el directorio donde `go install`
  deja el binario; `scripts/evals.sh` comprueba también que `kitlegal` está en el `PATH`, y la sesión de la prueba de
  red pide `kitlegal boe articulo BOE-A-2015-10565 a9998 --json` y la misma con `--offline`. Ninguna eval cambia.
- **La ayuda y los errores del binario enumeran `skills`**: `kitlegal --help` lo lista y un applet desconocido termina
  en `applets disponibles: boe, skills, territorio`.
- **`make test-e2e` construye además tres binarios de extremo a extremo** —con la versión `v0.1.0`, con `v0.2.0` y con
  `v0.1.0` y un creador de enlaces que siempre falla— y un origen de release local, que usan los guiones del aviso, del
  recurso de copia y del instalador.

### Eliminado

*De H19 — instalar sin clonar:*

- **La instalación por enlaces**: `skills/boe-legislacion/scripts/boe`, `skills/legal-core/scripts/territorio`,
  `scripts/instalar-skills.sh` con la comprobación previa de `make install`, `bin/instalado/`,
  `internal/skills/enlaces.go` con sus tests y las derivas de enlaces de `skills-sync` y `skills-check`
  (`enlace-ausente`, `enlace-sobrante` y `enlace-con-otro-destino`). `TestSinInstalacionPorEnlaces` comprueba que no
  vuelven y que ni los `SKILL.md`, ni el `Makefile`, ni los flujos nombran `scripts/boe`, `scripts/territorio` ni
  `bin/instalado`.

### Corregido

*De H5.1 — los avisos de vigencia en las evals:*

- **La traza de una sesión con la llamada desconocida cerrada se lee entera.** Un hilo que muere en la parada de
  entrada de una llamada que strace no llega a identificar deja `???( <unfinished ...>`, que la lectura de la traza ya
  admitía, o la misma llamada cerrada con el resultado de la llamada sin terminar y el relleno de alineación,
  `???()` seguido de espacios y `= ?`, que declaraba la sesión ilegible y hacía fallar el veredicto del job aunque
  todas las series pasaran. Así ocurrió con una sesión del modelo informativo en la ejecución `35156339496`. La
  segunda forma se lee ahora como la primera; con argumentos, con otro resultado o con la marca dentro sigue siendo
  ilegible.

*De H19 — instalar sin clonar:*

- **La traza de una sesión con un hilo huérfano que muere en la parada de entrada se lee entera.** Cuando el binario
  sale mientras su runtime crea un hilo, strace deja la `clone` sin terminar y, del hilo creado, un fichero sin la
  línea que lo crea. La lectura ya lo admitía si solo tenía su línea final, pero no si el núcleo lo mataba en la
  parada de entrada de una llamada que strace no llega a identificar, que deja `???()` seguido de espacios y `= ?`
  (o `???( <unfinished ...>`): lo tomaba por un segundo fichero raíz, declaraba la sesión ilegible y el veredicto
  del job fallaba aunque todas las series pasaran. Así ocurrió con una sesión del modelo que decide en la ejecución
  `36291141634`, y una sonda lo reproduce (`specs/009-h19-instalar-sin-clonar/gates/cierre-traza-huerfana.md`).
  Ese fichero es ahora un huérfano más: no pertenece a ningún proceso ni tiene nada que atribuir. La misma sonda
  dejó la otra forma con la que strace escribe esa parada, `syscall_0x<número>(<seis argumentos crudos>` sin
  terminar, cuando lee los registros del hilo pero el número no es de ninguna llamada que conozca, y se lee como la
  llamada desconocida. Un fichero sin línea de creación con cualquier llamada del filtro sigue siendo ilegible,
  también si la dejó sin terminar el fin del proceso.

Ninguna orden del `Makefile` espera ya su contenido de un hito posterior: `release`, que fallaba hasta H19, construye
el snapshot de la release. El binario que se publica registra **tres applets, `boe`, `skills` y `territorio`**: los de
las demás fuentes (`placsp`, `bdns`…) llegan en los hitos siguientes, en el orden de `docs/ROADMAP.md`.
