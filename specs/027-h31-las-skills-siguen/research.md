# Research: H31 · Las skills siguen la especificación Agent Skills

Fase 0 del plan. Cada decisión dice qué se elige, por qué y qué se descarta (constitución, «Criterio de decisión
autónoma», punto 3). Las verificaciones (V) se ejecutaron en la sesión del plan, en primer plano, y dicen la orden y lo
que dio; lo que no se pudo verificar sin red va como supuesto (S). La tabla de reglas de FR-021 está al final.

No queda ninguna duda abierta: las cinco del spec las cerró clarify (spec, «Clarifications»), y las del
diseño se cierran aquí.

## Decisiones

### D1 · El juez de las descripciones corre en el trabajo `evals (<skill>)` de cada skill y juzga la de su skill

- **Decisión**: cada trabajo de la matriz `evals` de `.github/workflows/evals.yml` (hoy `boe-legislacion`,
  `jurisprudencia` y `legal-core`) juzga, o reutiliza, la descripción de su propia skill y la publica en su
  `informe.json`, con su umbral (D7). Con las tres descripciones cambiadas (SC-010), cada trabajo juzga una.
- **Por qué**: el cierre del workflow (`scripts/workflow/cierre.sh`, `informe_del_trabajo` de `comun.sh`) y el informe
  final (`scripts/workflow/informe.sh`, `recuentos_y_umbrales`) leen el `informe.json` de cada trabajo
  `evals (<skill>)` y sus `umbrales` (V9); `scripts/workflow/` queda fuera de alcance (spec), así que solo un umbral en
  ese informe llega al informe final y a «Controles de umbral» como `evals:<skill>:<nombre>`. La reutilización de H30
  ya busca, por skill, el informe de ese trabajo en las medidas de `main` (`buscarEnLasMedidasDeMain`,
  `reutilizacion.go`): el juicio de la descripción se busca en el mismo recorrido (D4). El peor caso de cada trabajo
  crece en una descripción, como mucho seis votos (D10).
- **Descartado**: un trabajo aparte `descripciones` que juzgue las tres (su informe no lo leería el cierre sin tocar
  `scripts/workflow/`, y pediría otra búsqueda en `main` y otra marca de «informe escrito»); juzgar las tres en uno de
  los trabajos de la matriz (mezcla en el informe de una skill el juicio de otras y le suma su peor caso).

### D2 · El job lee la rúbrica, el esquema, los casos y la medida de `evidencias/adr-0046/`, sin copia

- **Decisión**: el código del job lee `rubrica.md`, `esquema.json`, `casos.yaml` y `medida.json` de
  `evidencias/adr-0046/`, tal cual. El control de identidad de FR-025 (SC-006) comprueba en `make ci` que la huella
  SHA-256 de cada uno de los cuatro es la que registra para él `evidencias/adr-0046/manifiesto.json`, el manifiesto de
  la evidencia (ADR 0018), y sale en rojo con un byte distinto en cualquiera; la correspondencia de `medida.json` con
  la rúbrica y los casos (sus `rubrica.sha256` y `casos.sha256`) es la comprobación de FR-026 (D8).
- **Por qué**: FR-025 pide usar los de `evidencias/adr-0046/` idénticos byte a byte; leerlos donde están no deja
  ninguna copia que pueda divergir. Una copia, como la de `evals/<skill>/juez/` de H24 y H25 (`TestCopiasDelJuez`),
  no tiene dónde vivir: `evals/<x>/` es el directorio de evals de la skill `<x>` —`calcularDependencias` lo recorre
  como tal (V8)—, y en cualquier carpeta que mide (`directoriosQueMiden`: `skills/`,
  `evals/`, `data/`, `cmd/`, `internal/`) un cambio solo del juez elegiría evals (`sin_asignar`, todas), contra FR-030,
  que pide que un cambio solo del juez no corra nada. `evidencias/` no mide (V8): cambiarla no elige ninguna eval.
- **Descartado**: copiar los cuatro a `evals/descripciones/juez/` (lo leería el arnés como las evals de una skill que
  no existe) o a `internal/evals/` (un cambio del juez elegiría las 45 evals por `comun`, y los datos no son código).

### D3 · Si la ejecución juzga descripciones lo decide la decisión de cada trabajo, por las rutas de la propuesta

- **Decisión**: la decisión del trabajo (`decidirElTrabajo`, `TestDecisionDelTrabajo`) lleva un bloque nuevo,
  `descripcion`, solo si la ejecución tiene propuesta y alguna ruta `skills/<s>/SKILL.md` cumple `cambiaLaDescripcion`
  (`eleccion.go`): su `description` difiere entre la base y la cabeza, o la skill está en un solo árbol. La comprobación
  recorre los `skills/<s>/SKILL.md` de los dos árboles, que son exactamente las rutas de la propuesta que pueden
  cambiar una descripción, con la regla de H30 tal cual. Vale en los dos modos: con reutilización y con la suite
  completa sobre una propuesta (`evals-completa`, `evals-prueba-de-red`), en los que `TestDecisionDelTrabajo` pasa a
  leer también la base cuando `-propuesta` y `-base` no están vacías. Sin propuesta (el despacho manual), no hay
  bloque. Sin bloque, el trabajo no comprueba la medida del juez de las descripciones, no abre ninguna de sus sesiones
  y no publica ni su juicio ni su umbral (FR-030; Clarificación Q2).
- **Por qué**: FR-030 decide por las rutas, no por el motivo de cada eval: en una propuesta que cambia también lo
  común, como esta, el motivo de cada eval es `comun` y el juez corre igual. En el modo completo la decisión no lee de
  la plataforma los ficheros de la propuesta («sin leer nada de la plataforma», `decidirElTrabajo`), pero la base y la
  cabeza dan la misma respuesta sin plataforma: un `SKILL.md` cuya descripción difiere entre las dos es una ruta de la
  propuesta.
- **Descartado**: decidir por el motivo `descripcion` de las evals elegidas (con `comun` delante, no aparece: lo
  descarta la Clarificación Q2); leer los ficheros de la propuesta de la plataforma también en el modo completo (más
  lecturas para la misma respuesta).

### D4 · La reutilización del juicio, por huella y como H30

- **Decisión**: la huella del juicio de una descripción es el SHA-256 de un texto con una línea por dato, en orden: la
  skill, el SHA-256 de su `description` (la de `LeerFrontmatter`), el de cada uno de los cuatro ficheros de D2, el
  modelo del juez y su versión de Claude Code. Con reutilización, la decisión busca en las mismas medidas de `main`
  que recorre H30 —10 como mucho, de menos de 90 días, de propuestas fusionadas, con el informe escrito— un bloque
  `descripcion` de la skill con la misma huella y juicio completo (juzgado, no «sin juzgar», FR-032), y si lo halla lo
  reutiliza con su procedencia (ejecución, trabajo, commit, propuesta, fecha); si no, la descripción se juzga. En la
  suite completa no busca, como H30 no busca resultados de evals: se juzga.
- **Por qué**: FR-030 pide la misma descripción, la misma skill y el mismo juez (rúbrica, esquema, casos, medida,
  modelo y versión), «buscado como H30 busca los resultados de las evals»; una sola lectura de la plataforma sirve a
  las dos búsquedas. Un juicio reutilizado no cuenta en el umbral (FR-031; ADR 0045, punto 1), esté marcado o no, como
  un resultado reutilizado de H30: el informe lo da con su marca y su procedencia.
- **Descartado**: una búsqueda propia (otra lectura de la plataforma con otra ventana); reutilizar solo los juicios
  sin marca (separaría la regla de la de H30, que reutiliza cualquier resultado completo).

### D5 · El mensaje del voto y la frase, los de `evidencias/adr-0046/guiones/juez.py`

- **Decisión**: `mensajeDeLaDescripcion(skill, descripcion)` devuelve `<descripcion skill="<skill>">`, salto de
  línea, la descripción sin blancos a los lados, salto, `</descripcion>`, dos saltos y «Responde a la pregunta de la
  rúbrica sobre esta descripción.», como `mensaje_de`. Los blancos que se quitan son los de `esBlanco` (`juez.go`), el
  conjunto con el que H24 igualó el `\s` y el `isspace` de Python (research D8, V6 de H24), que es el que usa
  `str.strip()` sin argumentos. La frase se comprueba con `fraseEsta` (`juez.go`), que es `frase_esta`: la función de
  `evidencias/adr-0046/guiones/juez.py` es, letra a letra, la de `evidencias/adr-0037/guiones/juez.py` (V5).
- **Por qué**: FR-027 y FR-028 fijan las dos funciones de la validación; reutilizar `fraseEsta` no añade nada.
- **Descartado**: `strings.TrimSpace` para el mensaje (difiere de `str.strip()` en U+001C a U+001F, research V6 de H24).

### D6 · La regla de los votos es la del juez de las evals, con una sola clase que decide

- **Decisión**: el juez de las descripciones es un `Juez` (`juez.go`) con una clase, `habla_en_primera_o_segunda_persona`,
  que decide, y la rúbrica, el esquema, los casos y la medida de D2. Lo vota la `votacion` de hoy (`nuevaVotacion`,
  `votosDelNumero`, `contar`), con `juzgar` partido en dos: lo que arma el mensaje de una respuesta
  (`mensajeDelVoto`) y lo que vota un mensaje contra el texto en que se busca la frase, que usan las respuestas y las
  descripciones. Así: tres «si» unánimes, pedidos por orden; un «si» con frase que no está, nulo y repetido una vez;
  el voto cortado por el tope de 35 s, repetido una vez; la sesión con error, sin la forma del esquema o cortada por un
  límite de uso, sin repetición y la descripción sin juzgar; como mucho seis votos (FR-028, FR-029, FR-032). El voto lo
  abre `nuevoVotanteDelGuion` con `scripts/evals-voto.sh`, que lee el modelo, la rúbrica y el esquema del directorio
  del juez (FR-027).
- **Por qué**: FR-028 a FR-032 son la regla del juez de las evals (ADR 0037) aplicada a un texto; duplicarla haría dos
  reglas que mantener. El tope de 35 s cubre los votos de la validación, de 3,9 s de mediana y 9 s como mucho
  (`evidencias/adr-0046/LEEME.md`, «Cómo se hizo», punto 4).
- **Descartado**: un votante y una regla propios.

### D7 · El umbral, el informe y el motivo de la descripción sin juzgar

- **Decisión**: con el bloque `descripcion` en la decisión, el `informe.json` del trabajo lleva un bloque `descripcion`
  (data-model §5; contracts/juez-de-las-descripciones.md §3) y un umbral `habla_en_primera_o_segunda_persona`, con
  `medida` las descripciones marcadas entre las juzgadas en la ejecución (0 o 1), `total` las juzgadas (0 si se
  reutilizó), `comparacion` `<=`, `umbral` 0 y `decide: true`; una marcada lo pone en rojo y el veredicto en `fallo`
  por `motivosDeLosUmbrales`. Una descripción sin juzgar añade el motivo «el juez de las descripciones dejó sin juzgar
  la descripción de <skill>: <motivo del voto>», distinto del de un umbral (FR-032). `informe.md` da la skill, la
  frase y el motivo de la marcada, y de cada una si se juzgó o se reutilizó y con qué procedencia (FR-031).
- **Por qué**: el umbral sin el sufijo `:<modelo>` no se cruza con las filas de la serie de cierres, que solo buscan
  `<clase>:<modelo>` y `<clase>:<modelo>:<modo>` de las clases del bloque `juez` (`serie.sh`, V9); el informe final
  recalcula `cumple` de `medida`, `comparacion` y `umbral` (`informe.sh`, `cumple_calculado`), que el umbral lleva.
- **Descartado**: meter la clase en el bloque `juez` del informe (la serie de cierres la contaría como una clase de
  las respuestas); nombrarla con el modelo del juez (no hay modelo de respuestas que nombrar).

### D8 · La comprobación de la medida, la de hoy

- **Decisión**: antes de abrir ninguna sesión, con el bloque `descripcion` en la decisión, `ejecutarElJob` comprueba
  la medida del juez de las descripciones con `comprobarLaMedida` (`medida.go`): huellas de la rúbrica y de los casos,
  modelo y versión del juez frente a los fijados, nombre de la clase y los dos recuentos a 0; si da alguna línea,
  escribe el informe sin medir y no abre ninguna sesión, ni de las evals ni del juez, como hoy con el juez de las
  respuestas (`TestEjecucionSinMedir`). En `make ci`, `TestMedidaVersionada` hace la misma comprobación con el modelo y
  la versión que fija `evals.yml` (`fijadosParaElJuez`) y sus seis mutaciones (FR-026, SC-008).
- **Por qué**: `comprobarLaMedida` ya lee esa forma (`leerMedidaDelJuez` ignora `votos` y `origen`, V8). «62 de 63» es
  `defectos.sin_marcar` 1, y «1 de 36», `correctos.marcados` 1: los dos recuentos que comprueba.

### D9 · La medida del juez de las descripciones con el código del job (FR-034)

- **Decisión**: el trabajo `medida` de `evals.yml` (etiqueta `evals-medir-juez` o entrada `medir_al_juez`) gana una
  entrada de su matriz, `descripciones`, con concurrencia 5; `scripts/evals-medir-juez.sh descripciones` ejecuta el
  punto de entrada `TestMedidaDelJuezDeLasDescripciones` (`//go:build evals`), que vota los 99 casos de
  `casos.yaml` con la votación de D6, por orden, y escribe en `-salida` la medida con las claves que lee
  `leerMedidaDelJuez` —`clase`, `fecha`, `modelo_del_juez`, `version_de_claude_code`, `rubrica`, `casos`,
  `defectos`, `correctos`— y `origen`, más una línea por caso mal etiquetado; el guion la imprime entre sus marcas,
  como hoy, sin escribir nada en el repositorio.
- **Por qué**: es el flujo de la medida de H24 y H25 con otros casos; los de las descripciones no se reconstruyen de
  informes (llevan la descripción y la etiqueta), así que no usan el `reconstructor`. El tope del trabajo `medida`
  (269 min) cubre la entrada nueva: 485 s + 60 s + ⌈99 × 6 / 5⌉ × 40 s = 5305 s (D10).
- **Descartado**: un trabajo `medida` aparte (repetiría sus pasos); escribir `votos.jsonl` (FR-034 pide la medida con
  la forma de `medida.json`, sin escribir en el repositorio).

### D10 · El tope de los trabajos (SC-015)

- **Decisión**: `peorCasoDelTrabajo` (`definicion.go`) suma, en todo trabajo `evals`, el peor caso del juez de una
  descripción: `votosPorRespuestaComoMucho` (6) × (`topeDelVoto` + `margenDelVoto`) = 240 s, y la instalación del
  Claude Code del juez (60 s) en el trabajo de una skill sin juez de las respuestas (hoy `legal-core`), que hoy no la
  cuenta. `TestDefinicionDelJob` falla si el tope del trabajo (`timeout-minutes`, 352 min = 21 120 s) es menor que el
  mayor peor caso, y su cadena esperada de cada skill cambia en esas cifras: el de `boe-legislacion` pasa de 19 877 s a
  20 117 s (V10), el de `jurisprudencia` de 13 477 s a 13 717 s y el de `legal-core`, el mayor, de 20 425 s a
  20 725 s, todos por debajo del tope, que no cambia. El trabajo `medida` gana la entrada de D9.
- **Por qué**: SC-015 pide que el tope cubra lo que suman el peor caso de las evals y el de las descripciones; en un
  trabajo es una descripción, la de su skill (D1). El paso que instala los dos Claude Code ya corre en todo trabajo
  `evals` (V10).

### D11 · La validación de la especificación: nombre, etiquetas XML y descripción

- **Decisión**: `ValidarFrontmatter` añade a las reglas de hoy dos defectos de `name` —contiene `anthropic` o
  `claude` (subcadena, en minúsculas, que es lo único que admite el nombre), y una etiqueta XML— y uno de
  `description`, una etiqueta XML; y su máximo de `description` pasa de 1024 a 900 caracteres (runas). Una etiqueta
  XML (FR-003) es una coincidencia de `</?\p{L}[^<>]*>`: `<`, una barra opcional, una letra y lo que siga sin otro `<`
  ni `>` hasta un `>`; cubre `<nombre…>`, `</nombre>` y `<nombre/>`. «< 5000» y «a<b» no casan. Sustituye a la regla
  de hoy «description con < o >», que daría como defecto «< 5000» (SC-002). Cada defecto de `ValidarFrontmatter` pasa a
  empezar por `SKILL.md: ` para nombrar el fichero (FR-011), con el texto de hoy detrás.
- **Por qué**: FR-002 y FR-003 lo fijan; con 900 como máximo, el de 1024 de la especificación ya no puede darse, y
  dejar los dos daría dos defectos por una misma descripción. Un `name` con `<` ya es hoy defecto por su juego de
  caracteres; el de etiqueta XML se añade igual, porque SC-002 pide su mutante.
- **Descartado**: mantener «< o >» además de la etiqueta (pondría en rojo «< 5000», contra US2.2); contar
  `claude` solo como palabra entre guiones (FR-002 dice «no contiene»).

### D12 · El cuerpo y sus presupuestos

- **Decisión**: el cuerpo es el `SKILL.md` regenerado (con la región de la tabla de comandos como la deja
  `make skills-sync`) desde el byte que sigue a la línea `---` que cierra el frontmatter. Se cuentan sus líneas con
  `ContarLineas` y sus caracteres con `utf8.RuneCount`; es defecto con 500 líneas o más (`SKILL.md: cuerpo de 500
  líneas (máximo 499)`) y con más de 11 000 caracteres (`SKILL.md: cuerpo de 11001 caracteres (máximo 11000)`). Lo
  comprueba `regenerarSkill` sobre el contenido regenerado, donde hoy está `defectoDeLasLineas`, que desaparece con
  `maximoDeLineas` (FR-004, FR-010).
- **Por qué**: es la definición de FR-004 y la medida de la entrada (spec, «Contexto»: 283, 283 y 193 líneas; V1); en
  `regenerarSkill`, `make skills-sync` da el defecto y no escribe nada, como hoy (FR-011).

### D13 · La regla de `references/`: clasificación, sincronía y gates

- **Decisión**: una función de `internal/skills`, `ReferenciasAMano(arbol fs.FS, skill string) ([]Referencia, error)`,
  da las referencias a mano de una skill según FR-012: los ficheros regulares `.md` de `skills/<skill>/references/` que
  `kitlegal-referencias` no declara y cuya primera línea no es la cabecera «generado desde» (`DatosDeLaReferencia`). La
  usan la sincronía, las dependencias de las evals (D15) y los controles de contenido (D16). En la sincronía:
  `sobrantes` deja de dar como `fichero-sobrante` a las referencias a mano (ni deriva ni retirada); el fichero no
  declarado que abre con la cabecera sigue siendo `fichero-sobrante`; el declarado, generado, sigue siendo
  `contenido-distinto` o `fichero-ausente` si no es lo que genera su fila, con o sin cabecera. Una entrada de
  `references/` que no es un fichero regular `.md` —un subdirectorio, un enlace, otra extensión— es un defecto
  (`references/<entrada>: no es un fichero .md`) y no deriva: `make skills-sync` no escribe nada ni la retira (FR-006).
  Sobre el cuerpo regenerado y las referencias —las generadas, como las regenera su fila; las a mano, como están—:
  cada referencia sin un enlace desde `SKILL.md` (FR-007), cada enlace de una referencia a otra de su skill (FR-008) y
  cada referencia de más de 100 líneas sin su índice (FR-009) es un defecto que nombra la skill, el fichero y la regla.
- **Por qué**: la clasificación de FR-012 la necesitan tres consumidores; una sola función evita tres lecturas de la
  misma regla (proporcionalidad: se traza a FR-012, FR-014 y FR-016). Que un no-`.md` sea defecto y no se retire es
  lo que pide FR-006 («defecto y `skills-check` sale en rojo»); hoy `skills-sync` lo borraba como sobrante.
- **Descartado**: clasificar por la cabecera sola (un fichero declarado que perdió la cabecera sería «a mano» y su
  deriva no se vería, contra FR-012 y el caso límite del spec).

### D14 · Enlaces, bloques de código e índice

- **Decisión**: un bloque de código es lo que va entre una línea que abre una valla —tras hasta tres espacios, tres o
  más acentos graves o virgulillas— y la que la cierra con la misma marca, o el final del fichero si no se cierra. Un
  enlace es un `[texto](destino)` en una línea fuera de un bloque de código; el destino va hasta el primer `)` o
  espacio. FR-007: el destino, sin `#…`, es `references/<fichero>.md`. FR-008: el destino de un enlace de una
  referencia, sin `#…` y resuelto desde `references/` (`path.Join`), es otro fichero de `references/` de la skill; el
  que va solo a `#apartado` o a la propia referencia no cuenta. FR-009: en una referencia de más de 100 líneas, los
  encabezados de nivel 2 son las líneas, fuera de los bloques, que empiezan por `## `; tiene que haber al menos uno, y
  antes del primero una lista —líneas que empiezan, tras su sangría, por `- `, `* ` o un número y `. `— en la que, por
  orden, cada encabezado tenga un elemento que contenga su texto (sin la sintaxis de enlace ni el énfasis); los
  elementos de más, como los de los apartados de nivel 3, no estorban. Es la definición del spec (FR-009 y
  Clarificación Q5).
- **Por qué**: el spec los define así («fuera de los bloques de código», «con o sin `#apartado`», «una lista, antes de
  su primer encabezado de nivel 2, que nombra cada encabezado de nivel 2 del fichero, en su orden»); no hay un
  analizador de Markdown en el repositorio (V7) ni una dependencia permitida que lo sea (constitución V), y estas tres
  formas son las que escriben las skills.
- **Descartado**: una dependencia de CommonMark (fuera de la lista de la constitución V, para tres formas); tratar
  también el código en línea o los bloques sangrados (el spec dice «bloques de código», y las skills sangran listas,
  no código).

### D15 · Las dependencias de cada eval (FR-014)

- **Decisión**: `datosDeLasReferencias` (`dependencias.go`) deja fuera de la parte (c) las referencias a mano de la
  skill (`ReferenciasAMano`), sin error; las demás siguen como hoy (la que no abre con la cabecera y no es a mano sigue
  siendo error). La parte (b) no cambia: ya lleva `skills/<s>/references/` entero como directorio, así que la huella de
  la eval cambia con una referencia a mano y una propuesta que solo la cambia elige las evals de esa skill y de las que
  la nombran (V8). `skillsDeLaEval` busca nombres de otras skills, con `nombra`, en el `SKILL.md` y en cada referencia a
  mano de cada skill que visita, sin mezclar sus bytes con los del `SKILL.md` que lee `AppletsDeLaTabla`.
- **Por qué**: FR-014 y la Clarificación Q1. Con las skills de `main`, sin referencias a mano, la elección no cambia.

### D16 · Los controles de contenido leen la skill entera

- **Decisión**: cada control de FR-016 lee, en las skills en que se aplica hoy, el `SKILL.md` y sus referencias a mano,
  cada fichero por separado (una orden de un fichero no se empareja con la forma para PowerShell de otro), y cada
  defecto nombra su fichero:
  - de ausencia, sobre los dos: la prosa sin expresiones prohibidas ni fechas AAAAMMDD con cifras
    (`boe-legislacion`); las normas nombradas que están en `data/normas.yaml` (todas); sin instrucciones de las evals
    (todas, que ya leía todo `references/`; su mutante sobre una referencia a mano es nuevo);
  - las órdenes para PowerShell (`boe-legislacion`): cada orden de lectura o de comprobación de cualquiera de los dos
    lleva detrás su forma, y la presencia de una orden de Bash por verbo de lectura se busca en la skill entera
    (FR-018);
  - las formas fijas que las evals comprueban sin modelo —la línea sin consulta (`boe-legislacion`, `legal-core`), las
    formas de aviso, de hallazgo y de la redacción fechada (`boe-legislacion`)— se siguen exigiendo en el `SKILL.md`:
    llevarlas solo a una referencia lo pone en rojo; y, en esas dos skills, cada línea de una referencia a mano que
    empieza, tras su sangría, un marcador de lista y el énfasis, por la marca `⚠` es, sin los blancos de los extremos,
    una línea del `SKILL.md`: una variante mal escrita de una forma fija es una línea `⚠` que el cuerpo no tiene
    (FR-018).
- **Por qué**: FR-016 a FR-018. La línea `⚠` es la forma fija (ADR 0046, decisión 1: «las formas fijas que las evals
  comprueban sin modelo (las líneas `⚠`)»), y las gramáticas con las que se juzga una respuesta toleran mayúsculas,
  blancos y énfasis, así que una variante con una errata ni siquiera la reconocerían: la comparación con las líneas
  del cuerpo es la que la ve.
- **Descartado**: aplicar la variante con la gramática de cada forma (no ve la errata que la gramática no reconoce);
  extender los controles a skills en que hoy no se aplican (FR-016 dice «en las skills a las que hoy se aplica»).

### D17 · Las órdenes de ejemplo que leen del BOE (FR-019)

- **Decisión**: un subtest nuevo de `TestEvalsDelRepositorio`, `ordenes-de-ejemplo`, extrae del `SKILL.md` y de las
  referencias a mano de cada skill con evals las órdenes concretas —en un bloque de código o en código en línea, los
  tramos que ya recorre `ordenesDeLectura`, cada aparición de `kitlegal boe` con lo que la sigue hasta el primer `&&`,
  `;` o `|`, sin `--json`, que lleva un verbo y al menos un argumento y ningún marcador `<…>`, `...` ni `…`— y las
  ejecuta con `ComprobarSinRed` (`preparar.go`), que lleva `--offline`, sobre la caché de una sesión de su skill
  (`cacheDeUnaSesion`): la de la primera eval de `evals/<skill>/`, por nombre de fichero, que no declara `sin_servir`.
  Sale en rojo con cada `Falta`. Tiene que encontrar al menos una orden en el repositorio, para no pasar en vacío. El
  extractor, `ordenesDeEjemplo`, lo fija `TestOrdenesDeEjemplo` (contracts/controles-de-contenido.md §10).
- **La mención del verbo no es una orden.** El texto nombra el verbo sin argumentos en la prosa —«`kitlegal boe
  articulo` o `articulos`» de la regla 6 (B123), que se queda en el cuerpo; «con `kitlegal boe articulo`» de B115, en
  references/lecturas_que_fallan.md; hoy, las l. 64, 77, 83, 271 y 282 de `skills/boe-legislacion/SKILL.md`—, y
  `boe articulo` o `boe articulos` sin norma ni bloque no pasan la gramática (argumentos obligatorios,
  `internal/app/boe.go`): ejecutarla daría una `Falta` por cada mención y `make ci` en rojo sin que ningún ejemplo
  esté mal. Es la frontera que ya fija `TestOrdenesParaPowerShell` («El texto de kitlegal boe articulo en la prosa no
  es una orden»).
- **Por qué**: FR-019. La caché de una sesión sirve, para las normas de `data/normas.yaml`, `indice`, `metadatos` y
  `articulo` con y sin `--fecha`, y ningún `buscar` ni `analisis` (V3, V4): las órdenes de ejemplo de hoy
  `kitlegal boe buscar régimen jurídico del sector público --json` y `kitlegal boe analisis BOE-A-2017-12902 --json`
  (`boe-legislacion`, paso 2 y paso 3) no se sirven, así que la reescritura las deja con marcador
  (`kitlegal boe buscar <términos> --json`) o las retira; no es una regla, es un ejemplo (contracts/skills-reescritas.md).
- **Descartado**: grabar `buscar` y `analisis` para las sesiones (datos nuevos y una fila de lo grabado que el spec no
  pide); ejecutar las órdenes con placeholders o la mención del verbo sin argumentos (no son una orden); reescribir
  las menciones del verbo para que no estén en código (cambiaría el texto de la regla 6, que FR-021 conserva).

### D18 · El reparto de cada skill y la tabla de reglas

- **Decisión**: la estructura de cada skill reescrita —el cuerpo por apartados, cada referencia a mano con sus
  apartados y la línea del cuerpo que dice cuándo leerla, las formas fijas que se quedan, los ejemplos que cambian y la
  descripción en tercera persona— está en contracts/skills-reescritas.md; la tabla de cada regla de `main` con su
  sitio, al final de este fichero (FR-021). Las cifras son de borradores completos escritos en la sesión del plan y
  medidos con `wc -lm` (V2), no versionados: `boe-legislacion` 10 881 caracteres y 130 líneas, `legal-core` 10 490 y
  161, `jurisprudencia` 10 605 y 157; sus descripciones, 878, 888 y 842. Los presupuestos los comprueban los gates
  sobre el texto que escriban las tareas.
- **Decisiones del reparto**: en `boe-legislacion`, las órdenes de `articulos` con su forma para PowerShell pasan a
  `references/varios_bloques.md` (el control de PowerShell lee la skill entera, D16), se retiran tres ejemplos que no
  llevan regla y la frase fija de la regla 7, que ningún control exige en el cuerpo, pasa con su caso a
  `references/lecturas_que_fallan.md`; la regla 7 se queda en el cuerpo. En `jurisprudencia`, lo que se le dice a la
  persona que haga (J27) se queda en el paso 2, porque lo necesita toda respuesta con una sentencia que no está
  delante. La región generada de la tabla de comandos no cambia (FR-024).
- **Por qué**: el cuerpo de `boe-legislacion` es el ajustado: la región generada (3076) y las formas fijas con sus
  órdenes son unos 4800 caracteres que no se tocan, y con `articulos` en el cuerpo el borrador daba 11 258 (V2).
- **Descartado**: llevar reglas invariantes a una referencia (ADR 0046, decisión 1); quitar filas de la tabla de
  comandos (FR-024).

### D19 · Las versiones en `CHANGELOG.md`

- **Decisión**: `boe-legislacion` pasa de v0.2.2 a v0.3, `legal-core` de v1.2.3 a v1.3 y `jurisprudencia` de v0.1.2 a
  v0.2 (versiones de hoy en `CHANGELOG.md`, V6): cambia la forma de lo que lee el modelo y no ninguna regla, que es un
  cambio menor, no un parche. La entrada dice también los gates nuevos y el juez de las descripciones (FR-023).

### D20 · FR-015 lo cumple `TestSkillsEmpotradas`, sin cambios

- **Decisión**: ningún mecanismo nuevo. `TestSkillsEmpotradas` (`skills_test.go` de la raíz, en `make test`) compara,
  ruta a ruta y byte a byte, el `SKILL.md` y todo fichero bajo `references/` de cada skill del árbol con lo empotrado,
  y falla si falta uno; `//go:embed skills/*/SKILL.md skills/*/references/*` empotra todo fichero de `references/`
  (V11). Una referencia a mano es un fichero más.

### D21 · Aceptación e2e: no aplica

- **Decisión**: «Aceptación e2e: no aplica» (plan.md, Complexity Tracking). Ningún applet, verbo, bandera, código de
  salida ni salida del binario cambia; `kitlegal skills install` copia lo empotrado sin cambios de código. Lo que
  entrega el hito son gates y controles de `make ci` (Go, offline), el job de evals y el texto de las tres skills, que
  miden sus 45 evals de hoy en el job (SC-010), que el spec no deja ampliar.
- **Descartado**: un guion `testscript` que instale las skills y mire que existe una referencia a mano: fijaría en la
  suite congelada los nombres de unas referencias que diseña el hito, y lo empotrado ya lo comprueba D20.

## Verificaciones

- **V1 · Las medidas de hoy.** `LC_ALL=en_US.UTF-8 awk 'f>=2{print} /^---/{f++}' skills/<skill>/SKILL.md | wc -lm`:
  `boe-legislacion` 283 líneas y 24 439 caracteres, `legal-core` 283 y 19 225, `jurisprudencia` 193 y 13 637. La región
  de la tabla de comandos (`awk '/inicio de la tabla de comandos/{p=1} p{print} /fin de la tabla de comandos/{p=0}'`):
  3076, 1498 y 1302 caracteres. `wc -l` de `skills/*/references/*.md`: la más larga, `normas.md`, 24 líneas.
- **V2 · Los borradores.** `wc -lm` sobre los apartados de cada borrador, escritos en la sesión del plan fuera del
  repositorio: `boe-legislacion` 130 líneas y 10 881 caracteres (11 030 antes de retirar tres ejemplos, y 11 258 con
  `articulos` en el cuerpo); `legal-core` 161 y 10 490; `jurisprudencia` 157 y 10 605; referencias a mano de
  `legal-core` 64 y 59 líneas, de `jurisprudencia` 43 y 50 (38 tras dejar J27 en el cuerpo); descripciones 878, 888 y
  842.
- **V3 · Lo que sirve la caché de una sesión, 1.** Un test temporal en `internal/evals/`, borrado en la misma sesión,
  con `cacheDeUnaSesion` y `ComprobarSinRed` sobre las órdenes de ejemplo concretas de hoy: desde la sesión de la 01 de
  `boe-legislacion`, `buscar régimen jurídico del sector público` y `analisis BOE-A-2017-12902` salen con 4 («con
  --offline no se pide nada a la fuente y no hay ninguna entrada vigente»); `indice BOE-A-2015-10565`,
  `articulo BOE-A-2015-10565 a21` y `articulo BOE-A-1985-5392 a22 --fecha 2005-06-15` salen con 0. Desde la de la 01 de
  `legal-core`, además, la de `BOE-A-1985-5392` sale con 4.
- **V4 · Lo que sirve la caché de una sesión, 2.** El mismo arnés, desde la 01 de `boe-legislacion`:
  `buscar procedimiento administrativo común` y `analisis BOE-A-2015-10565` salen con 4; `metadatos BOE-A-2015-10565`,
  `indice BOE-A-2017-12902` y `articulo BOE-A-2015-10565 a21 --fecha 9999-12-31` con 0. Ningún `buscar` ni `analisis`
  se sirve.
- **V5 · La frase.** `grep -n -A6 "^def normal\|^def frase_esta"` en `evidencias/adr-0037/guiones/juez.py` y la lectura
  de `evidencias/adr-0046/guiones/juez.py`: `normal` y `frase_esta` son el mismo código. `fraseEsta` y
  `sinEnfasisNiBlancosDeMas` (`internal/evals/juez.go`) son las de H24.
- **V6 · Versiones de las skills.** `grep -o` en `CHANGELOG.md`: las más altas son `boe-legislacion` v0.2.2,
  `legal-core` v1.2.3 y `jurisprudencia` v0.1.2.
- **V7 · Markdown en el repositorio.** `git grep` de `](` y de «enlace» en `internal/skills/*.go` e
  `internal/app/*.go`: ningún analizador de enlaces (lo que sale no es Markdown); `go.mod` no lleva `goldmark`,
  `blackfriday` ni otra dependencia de CommonMark. Los bloques `text` los recorre `bloquesDeTexto`
  (`internal/evals/conjunto_test.go`, l. 2690), en los tests.
- **V8 · Lo que mide y cómo se recorre `evals/`.** `internal/evals/dependencias.go`: `directoriosQueMiden` es
  `skills/`, `evals/`, `data/`, `cmd/`, `internal/`; `directoriosDeLoComun` incluye `internal/evals/`;
  `calcularDependencias` toma cada directorio de `evals/` como las evals de una skill. La parte (b) de
  `deUnaSkill` es `skills/<s>/SKILL.md` y el directorio `skills/<s>/references/`; `datosDeLasReferencias` da error con
  «la referencia %s no tiene la cabecera «generado desde»». `leerMedidaDelJuez` (`medida.go`) lee solo `clase`,
  `fecha`, `modelo_del_juez`, `version_de_claude_code`, `rubrica.sha256`, `casos.sha256` y los recuentos.
- **V9 · Lo que leen la serie y el informe final.** `scripts/workflow/serie.sh`, l. 163–170: cruza solo
  `.nombre == $c.clase + ":" + $m` y con `":" + $modo`, para las clases del informe; `scripts/workflow/informe.sh`
  recorre todos los `umbrales` del informe de cada skill (l. 286–310).
- **V10 · Los trabajos del job.** `.github/workflows/evals.yml`: el paso «Instalar strace y Claude Code» instala el
  Claude Code del juez en todo trabajo `evals`, sin condición (l. 354–366); `timeout-minutes: 352` del trabajo `evals` y
  269 del trabajo `medida`, con la matriz `[boe-legislacion, jurisprudencia]`. El peor caso de hoy de `boe-legislacion`
  es la constante `peorCasoDeBoeLegislacion`, «19877 s …», de `internal/evals/definicion_test.go` (l. 2804; `grep -n`).
- **V11 · Go.** `go version`: go1.27.3. `go doc embed`: un patrón como `image/*` empotra también `image/.tempfile`.
  `go doc unicode/utf8.RuneCount`: el número de runas. `go help testflag`: `-run` parte el patrón por las barras sin
  corchete, una parte por nivel de subtest.
- **V12 · El juez de las descripciones con el código de hoy.** Un test temporal en `internal/evals/`, borrado en la
  misma sesión: el SHA-256 de `rubrica.md` (`db35645c…`), `esquema.json` (`9fddb8c3…`), `casos.yaml` (`b9b2384a…`) y
  `medida.json` (`1bb9b689…`) de `evidencias/adr-0046/` es el que registra `manifiesto.json` para cada uno;
  `leerMedidaDelJuez` lee `medida.json` sin error; `casos.yaml` lleva 63 casos `defecto` y 36 `correcto`; y
  `comprobarLaMedida` sobre un `Juez` con la clase `habla_en_primera_o_segunda_persona`, esa rúbrica, esos casos y esa
  medida, con `claude-opus-5-5` y 2.1.289, no da ninguna línea (D2, D6, D8).

## Supuestos no verificados

- **S1 · Las dos fuentes de la guía.** Las citas de la especificación Agent Skills y de la guía de Anthropic (500 líneas,
  «< 5000 tokens recommended», 1024 caracteres, etiquetas XML, `anthropic` y `claude` en el nombre, tercera persona,
  índice en las referencias de más de 100 líneas) son las que registra el ADR 0046, consultadas el 2026-10-23; la sesión
  no tiene red y no las vuelve a leer (FR-001).
- **S2 · GitHub Actions.** Que una entrada nueva de `matrix.skill` con su `include` (`descripciones`, concurrencia 5)
  da un trabajo más de `medida` se apoya en la misma construcción que ya usa ese trabajo en `evals.yml`; no hay
  documentación de Actions en local.

## Datos externos

Ninguno: el hito no graba nada ni deriva ningún fichero de `data/`. La rúbrica, los casos y la medida son de
`evidencias/adr-0046/`, que está en `main` (spec, «Input») y solo se lee.

## Tabla de reglas (FR-021)

Una fila por regla de la versión de `main` de cada skill, con su sitio en la nueva. «En main» da el apartado y las
líneas de `skills/<skill>/SKILL.md` en `main`; «En la nueva», el fichero y el apartado de contracts/skills-reescritas.md.
Las líneas del cuerpo que dicen cuándo leer cada referencia son punteros, no reglas. La revisión final contrasta la
tabla con las dos versiones (SC-012).

### boe-legislacion

Fuente: `skills/boe-legislacion/SKILL.md` de `main` (299 líneas; cuerpo, l. 17–299). Una fila por regla; cuando `main`
dice la misma regla en varios sitios, la fila da todos. Los apartados nuevos son los de la estructura nueva: el
«Protocolo» con cinco pasos numerados (1 Identificar la norma, 2 Leer, 3 Completar, 4 Responder, 5 De un precepto no
leído, nada), «Cómo se cita», «Redacción en una fecha», «Redacción modificada», «Comandos» y «Reglas», con los números
de hoy, y siete referencias escritas a mano.

| Id | Regla | En main | En la nueva |
|---|---|---|---|
| B01 | Consulta y cita normativa consolidada del BOE de cualquier materia. | frontmatter, description (l. 4–6) | frontmatter, description |
| B02 | Se activa siempre que la respuesta dependa de lo que dice una norma: un artículo, una ley o un real decreto, un plazo, requisito o procedimiento, dónde se regula una materia. | frontmatter, description (l. 6–7) | frontmatter, description |
| B03 | Se activa también si depende de lo que decía una norma en una fecha, la del acto o la de los hechos. | frontmatter, description (l. 7–8) | frontmatter, description |
| B04 | También cuando se cree conocer la respuesta: sin leer la norma con kitlegal no hay cita. | frontmatter, description (l. 8) | frontmatter, description |
| B05 | Se activa antes de llamar a `boe_articulo` u otra herramienta `boe_…`: qué pedir y cómo citar lo dice la skill. | frontmatter, description (l. 9) | frontmatter, description |
| B06 | La norma puede nombrarse por número y año, abreviatura o `BOE-A-…`, y ser estatal o autonómica consolidada en el BOE, hoy o en una fecha. | frontmatter, description (l. 9–12) | frontmatter, description (la fecha, con B03) |
| B07 | Lee índice y artículos con kitlegal y cita identificador y bloque. | frontmatter, description (l. 12) | frontmatter, description |
| B08 | Alcance: normas consolidadas del BOE (Constitución, leyes, reales decretos legislativos, reales decretos y autonómicas que consolida), de cualquier ámbito. | SKILL.md, introducción (l. 20–21) | frontmatter, description (B01, B06); SKILL.md, título |
| B09 | Lo que se dice de una norma sale del texto que devolvió `kitlegal boe` en la misma conversación, con su cita. | SKILL.md, introducción (l. 22–23) | SKILL.md, «Protocolo», paso 4 |
| B10 | La skill consulta y cita: no tramita nada y no sustituye el asesoramiento de un profesional. | SKILL.md, introducción (l. 23) | SKILL.md, «Reglas», 3 y 4 |
| B11 | Seguir los pasos del protocolo en orden. | SKILL.md, «Protocolo» (l. 27) | SKILL.md, «Protocolo» (párrafo inicial) |
| B12 | Orden y herramienta devuelven el mismo sobre; sin `fuente`, `url`, `fecha_consulta` y `hash` no hay cita. | SKILL.md, «Protocolo» (l. 27–28) | SKILL.md, «Protocolo» (párrafo inicial) |
| B13 | Con la herramienta de su nombre, sola o tras el prefijo del agente (`mcp__kitlegal__boe_articulo`); si se tiene, siempre. | SKILL.md, «Protocolo» (l. 30–32) | SKILL.md, «Protocolo» (párrafo inicial: «si la tienes, siempre»); references/con_herramientas.md, «Nombres» |
| B14 | Los argumentos de una herramienta van por su nombre: `norma`, `bloque`, `bloques`, `texto`, `fecha`. | SKILL.md, «Protocolo» (l. 31–32) | references/con_herramientas.md, «Argumentos» |
| B15 | Sin la herramienta, la orden `kitlegal <applet> <verbo> … --json`. | SKILL.md, «Protocolo» (l. 33) | SKILL.md, «Protocolo» (párrafo inicial) |
| B16 | Dos órdenes unidas con `&&` son, con herramientas, dos llamadas seguidas; la segunda, solo si la primera no falló. | SKILL.md, «Protocolo» (l. 34–35) | references/con_herramientas.md, «Encadenar y fallar» |
| B17 | Donde una orden termina con un código, la herramienta da un error y su `data.clase` dice cuál. | SKILL.md, «Protocolo» (l. 35–36) | references/con_herramientas.md, «Encadenar y fallar» |
| B18 | Antes de consultar nada, localizar la norma en `references/normas.md` por nombre, número y año o abreviatura. | SKILL.md, «1. Identificar la norma» (l. 40–41) | SKILL.md, «Protocolo», paso 1 (las formas de nombrarla, en frontmatter, description, B06) |
| B19 | La tabla de normas no es exhaustiva. | SKILL.md, «1. Identificar la norma» (l. 41–42) | SKILL.md, «Protocolo», paso 1 |
| B20 | Si la norma está en la tabla, su `BOE-A-…` se toma de ahí. | SKILL.md, «2. Resolver» (l. 46) | SKILL.md, «Protocolo», paso 1 |
| B21 | Si no está, buscarla por las palabras de su título, con `boe_buscar` (en `texto`) o con su orden. | SKILL.md, «2. Resolver» (l. 47–48) | SKILL.md, «Protocolo», paso 1 (el cuándo); references/buscar_la_norma.md, «Buscar por el título» |
| B22 | Elegir entre los resultados por título y rango, y decir en la respuesta qué norma se eligió. | SKILL.md, «2. Resolver» (l. 48–49) | references/buscar_la_norma.md, «Buscar por el título» |
| B23 | Con varias posibles (ley y texto refundido, ley y reglamento, estatal y autonómica), decir cuáles y por qué se elige una. | SKILL.md, «2. Resolver» (l. 49–51) | references/buscar_la_norma.md, «Varias posibles» |
| B24 | Si la pregunta no permite elegir, preguntar o responder de ambas, distinguiéndolas. | SKILL.md, «2. Resolver» (l. 50–51) | references/buscar_la_norma.md, «Varias posibles» |
| B25 | Una norma autonómica consolidada en el BOE se resuelve igual y, si la búsqueda la encuentra, se lee y se cita como una estatal. | SKILL.md, «2. Resolver» (l. 52–53) | references/buscar_la_norma.md, «Normas autonómicas» |
| B26 | Si la búsqueda no da la norma, reformularla con otras palabras del título; que no aparezca no prueba que no exista. | SKILL.md, «2. Resolver» (l. 53–54) | references/buscar_la_norma.md, «Si no aparece»; SKILL.md, «Reglas», 1 |
| B27 | Sin el id del bloque (el número del artículo no basta), leer el índice, con `boe_indice` o su orden. | SKILL.md, «3. Leer» (l. 58–59) | SKILL.md, «Protocolo», paso 2 |
| B28 | Copiar el id de la entrada cuyo `titulo` es el artículo; nunca componerlo con el número (Ley 9/2017: «Artículo 118» es `a1-30`). | SKILL.md, «3. Leer» (l. 60–62) | SKILL.md, «Protocolo», paso 2 |
| B29 | Sin fecha, leer los bloques de uno en uno con `kitlegal boe articulo` y `kitlegal graph check` de esa norma y bloque detrás, en la misma orden. | SKILL.md, «3. Leer» (l. 63–67) | SKILL.md, «Protocolo», paso 2 |
| B30 | Con fecha en la pregunta, se lee como dice «Redacción en una fecha». | SKILL.md, «3. Leer» (l. 63) | SKILL.md, «Protocolo», paso 2 |
| B31 | En PowerShell 5.1, sin `&&`, la misma orden con `; if ($LASTEXITCODE -eq 0) { … }`. | SKILL.md, «3. Leer» (l. 68–71, 79–80) | SKILL.md, «Protocolo», paso 2 (la de `articulo`); references/varios_bloques.md, «La orden» (la de `articulos`) |
| B32 | Con herramientas, `graph_check` lleva la misma `norma` y el bloque en `bloques`. | SKILL.md, «3. Leer» (l. 72) | references/con_herramientas.md, «Las órdenes del protocolo» |
| B33 | Si `kitlegal graph check` termina con un código que no es `0`, vale la regla 7. | SKILL.md, «3. Leer» (l. 73–74) | SKILL.md, «Protocolo», paso 2 (enlace de lecturas_que_fallan) y «Reglas», 7 |
| B34 | La comprobación va siempre detrás de la lectura, en su misma orden y con su misma norma y sus mismos bloques. | SKILL.md, «3. Leer» (l. 74–75) | SKILL.md, «Protocolo», paso 2 |
| B35 | Nunca pedir `kitlegal graph check` sin argumentos ni antes de leer. | SKILL.md, «3. Leer» (l. 75) | SKILL.md, «Protocolo», paso 2 |
| B36 | Si un bloque se leyó sin la comprobación, pedirla enseguida con esa norma y ese bloque. | SKILL.md, «3. Leer» (l. 75–76) | SKILL.md, «Protocolo», paso 2 |
| B37 | `kitlegal boe articulos` solo para varios bloques a la vez, todos del índice, con la comprobación de esos bloques detrás. | SKILL.md, «3. Leer» (l. 77–80) | SKILL.md, «Protocolo», paso 2 (el cuándo); references/varios_bloques.md, «Cuándo» y «La orden» |
| B38 | Sin fecha, cada bloque se lee una sola vez por pregunta: otra lectura apagaría su línea `⚠ REDACCIÓN MODIFICADA:`. | SKILL.md, «3. Leer» (l. 81) | SKILL.md, «Protocolo», paso 2 |
| B39 | Una lectura de varios bloques falla entera; con `4` o `5`, pedir cada bloque por separado antes de dar ninguno por no consultado. | SKILL.md, «3. Leer» (l. 82–84) | references/lecturas_que_fallan.md, «Una lectura de varios bloques» |
| B40 | Seguir las remisiones necesarias: leer el bloque remitido, de esa norma o de otra, resuelta con los pasos de identificar. | SKILL.md, «3. Leer» (l. 85–86) | SKILL.md, «Protocolo», paso 3 |
| B41 | Si la pregunta depende de la vigencia o de las modificaciones, leer metadatos y análisis (`boe_metadatos`, `boe_analisis` o sus órdenes). | SKILL.md, «3. Leer» (l. 87–88) | SKILL.md, «Protocolo», paso 3; references/con_herramientas.md, «Las órdenes del protocolo» |
| B42 | Nunca pedir un id de bloque que no salga del índice o de la pregunta. | SKILL.md, «3. Leer» (l. 89) | SKILL.md, «Protocolo», paso 2 |
| B43 | Si `kitlegal boe` termina con `3`, volver al índice en vez de probar otros ids; si el artículo no existe en la norma, decirlo. | SKILL.md, «3. Leer» (l. 89–90) | references/lecturas_que_fallan.md, «Un bloque que no está» |
| B44 | Si el bloque remite a otro precepto u otra norma que cambia la respuesta, leerlo. | SKILL.md, «4. Evaluar» (l. 94) | SKILL.md, «Protocolo», paso 3 |
| B45 | Avisos, o una fecha de vigencia que no encaja con la situación preguntada: tenerlos en cuenta y trasladarlos. | SKILL.md, «4. Evaluar» (l. 95–96) | SKILL.md, «Reglas», 3 |
| B46 | Si una norma posterior cambió el bloque y eso importa, leer sus metadatos o su análisis. | SKILL.md, «4. Evaluar» (l. 97) | SKILL.md, «Protocolo», paso 3 |
| B47 | Lo que no se puede leer no se suple. | SKILL.md, «4. Evaluar» (l. 98) | SKILL.md, «Reglas», 2 |
| B48 | Lo que se decide al evaluar el contexto no va en la respuesta. | SKILL.md, «4. Evaluar» (l. 98) | SKILL.md, «Protocolo», paso 4 («de qué has decidido») |
| B49 | La respuesta empieza por lo que se pregunta y habla de la norma, no de la consulta. | SKILL.md, «5. Responder» (l. 102) | SKILL.md, «Protocolo», paso 4 |
| B50 | La respuesta es lo escrito tras la última orden o llamada; quien pregunta no ve órdenes, llamadas ni lo que devuelven. | SKILL.md, «5. Responder» (l. 102–104) | SKILL.md, «Protocolo», paso 4 |
| B51 | La respuesta la hacen la norma, su texto, su cita, sus avisos de vigencia y las líneas `⚠` de cada bloque. | SKILL.md, «5. Responder» (l. 104–105) | SKILL.md, «Protocolo», paso 4 |
| B52 | Escribir como quien tiene el BOE delante, con la norma por su nombre e identificador, sin decir de qué tabla u orden sale. | SKILL.md, «5. Responder» (l. 105–107) | SKILL.md, «Protocolo», paso 4 |
| B53 | No decir lo hecho, lo que se va a hacer ni lo que ha devuelto o no una orden. | SKILL.md, «5. Responder» (l. 107) | SKILL.md, «Protocolo», paso 4 |
| B54 | Nada de otra conversación: no hablar de ella, ni para afirmar, confirmar o desmentir. | SKILL.md, «5. Responder» (l. 108–109) | SKILL.md, «Protocolo», paso 4 |
| B55 | De otras conversaciones, la respuesta solo lleva la línea `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae. | SKILL.md, «5. Responder» (l. 109–110); «Redacción modificada» (l. 219–220) | SKILL.md, «Protocolo», paso 4; «Redacción modificada» |
| B56 | Cada afirmación sobre el contenido de una norma lleva su cita, de lo que devolvió `kitlegal boe` en esta conversación. | SKILL.md, «5. Responder» (l. 111–112) | SKILL.md, «Protocolo», paso 4 |
| B57 | La cita: forma legible de norma y bloque y, en la misma línea, `[<identificador>, bloque <id>]` tal como los da la fuente. | SKILL.md, «5. Responder» (l. 112–113); «Cómo se cita» (l. 134–135) | SKILL.md, «Cómo se cita» |
| B58 | De un precepto no leído no decir qué dice, de qué trata ni cuándo o cómo se aplica, ni en un inciso ni creyendo saberlo. | SKILL.md, «5. Responder» (l. 114–116) | SKILL.md, «Protocolo», paso 5 |
| B59 | El índice no dice qué dice un bloque: solo da su número. | SKILL.md, «5. Responder» (l. 116–117) | SKILL.md, «Protocolo», paso 5 |
| B60 | Número y materia de un precepto no leído nunca juntos en una frase: ni afirmando, ni negando, ni al ofrecer leerlo, ni al avisar. | SKILL.md, «5. Responder» (l. 117–120) | SKILL.md, «Protocolo», paso 5 |
| B61 | Para decir de qué trata lo remitido, leerlo y citarlo; si no, la remisión con las palabras del texto leído y nada detrás. | SKILL.md, «5. Responder» (l. 120–121) | SKILL.md, «Protocolo», paso 5 |
| B62 | Se puede avisar de que una materia se regula en otra parte, sin nombrar precepto ni regla. | SKILL.md, «5. Responder» (l. 121) | SKILL.md, «Protocolo», paso 5 |
| B63 | Con normas de rango distinto, decir el de cada una (de la tabla o de la búsqueda) y que la ley prevalece sobre su reglamento. | SKILL.md, «5. Responder» (l. 122–123) | SKILL.md, «Protocolo», paso 4 |
| B64 | Señalar cuándo lo preguntado puede variar por normativa autonómica. | SKILL.md, «5. Responder» (l. 124–125) | SKILL.md, «Reglas», 5 |
| B65 | Si corresponde a ordenanzas u otras normas locales, decir que no están en esta fuente. | SKILL.md, «5. Responder» (l. 125–126); «Reglas», 5 (l. 279–280) | SKILL.md, «Reglas», 5 |
| B66 | Trasladar cada aviso de vigencia de la lectura con su forma fija. | SKILL.md, «5. Responder» (l. 127); «Cómo se cita» (l. 154–157); «Reglas», 3 (l. 272–274) | SKILL.md, «Reglas», 3 |
| B67 | De la redacción leída, decir qué norma la dio y desde cuándo rige. | SKILL.md, «5. Responder» (l. 127–128); «Redacción en una fecha» (l. 188–189) | SKILL.md, «Protocolo», paso 4; references/redaccion_en_una_fecha.md, «Lo que se cita» |
| B68 | Hasta cuándo rige una redacción, nunca: ninguna lectura lo trae, tampoco en una norma derogada, cuyo aviso no lleva fecha. | SKILL.md, «5. Responder» (l. 128–130); «Redacción en una fecha» (l. 189) | SKILL.md, «Protocolo», paso 4; references/redaccion_en_una_fecha.md, «Lo que se cita» |
| B69 | `kitlegal graph check` no dice nada de la vigencia. | SKILL.md, «5. Responder» (l. 130) | SKILL.md, «Protocolo», paso 4 |
| B70 | Repasar cada cita de la respuesta con las reglas de la cita. | SKILL.md, «Cómo se cita» (l. 135) | SKILL.md, «Cómo se cita» |
| B71 | La forma recomendada pone la forma legible de la norma y del bloque delante del corchete. | SKILL.md, «Cómo se cita» (l. 135–140) | SKILL.md, «Cómo se cita» |
| B72 | Hace cita que los corchetes, en la misma línea, terminen en `<identificador>, bloque <id>]`, sin nada entre el id y el corchete. | SKILL.md, «Cómo se cita» (l. 142–144) | SKILL.md, «Cómo se cita»; references/como_se_cita.md, «Lo que hace cita» |
| B73 | La forma legible dentro de los corchetes va delante del identificador, separada por una coma. | SKILL.md, «Cómo se cita» (l. 145–146) | references/como_se_cita.md, «La forma legible dentro» |
| B74 | Detrás del id no va nada; tampoco son cita el identificador sin el id del bloque ni los dos sin corchetes. | SKILL.md, «Cómo se cita» (l. 146–148) | SKILL.md, «Cómo se cita» («sin nada tras el id»); references/como_se_cita.md, «Lo que no es cita» |
| B75 | La regla vale igual con la cita sola en una línea o bajo una cita textual en bloque. | SKILL.md, «Cómo se cita» (l. 149–150) | references/como_se_cita.md, «Sola o bajo una cita textual» |
| B76 | Identificador e id, tal como los devuelve `kitlegal boe`, también si el id termina en punto. | SKILL.md, «Cómo se cita» (l. 150–151) | SKILL.md, «Cómo se cita»; references/como_se_cita.md, «Ids que terminan en punto» |
| B77 | Una cita por bloque; un bloque remitido se cita por separado, con su norma y su id. | SKILL.md, «Cómo se cita» (l. 152) | SKILL.md, «Cómo se cita»; references/como_se_cita.md, «Bloques remitidos» |
| B78 | El aviso va con `⚠`, su etiqueta (lo que el `texto` lleva entre la marca y los dos puntos), dos puntos y la frase o una explicación. | SKILL.md, «Cómo se cita» (l. 154–156); «Reglas», 3 (l. 273–274) | SKILL.md, «Reglas», 3; references/como_se_cita.md, «Los avisos» (de dónde sale la etiqueta) |
| B79 | Etiquetas: `⚠ NORMA DEROGADA:`, `⚠ VIGENCIA AGOTADA:`, `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:`, una por código de aviso. | SKILL.md, «Cómo se cita» (l. 156–160) | SKILL.md, «Reglas», 3 |
| B80 | Cada línea `⚠` va entera, en una línea, con su forma y los datos de la lectura; dicha con otras palabras, falta. | SKILL.md, «Cómo se cita» (l. 162–165) | SKILL.md, «Cómo se cita» |
| B81 | Una fecha (día, mes, año o un hecho con día) obliga a citar la redacción vigente ese día, no la de hoy. | SKILL.md, «Redacción en una fecha» (l. 169–170) | SKILL.md, «Redacción en una fecha» (cuándo); references/redaccion_en_una_fecha.md, «Qué es una fecha» |
| B82 | Con fecha, cada bloque se lee con `--fecha AAAA-MM-DD` o con `boe_articulo` y `fecha`. | SKILL.md, «Redacción en una fecha» (l. 170–171) | references/redaccion_en_una_fecha.md, «Cómo se lee»; references/con_herramientas.md, «Argumentos» |
| B83 | La lectura con fecha va sin `kitlegal graph check`. | SKILL.md, «Redacción en una fecha» (l. 172) | SKILL.md, «Redacción en una fecha»; references/redaccion_en_una_fecha.md, «Cómo se lee» |
| B84 | Si un hecho sitúa el momento y nadie da su día, pedir el día; no leer ni citar ninguna redacción como la de entonces. | SKILL.md, «Redacción en una fecha» (l. 172–173) | references/redaccion_en_una_fecha.md, «Un hecho sin día» |
| B85 | Con un año o un mes sin día, leer con su primer y con su último día. | SKILL.md, «Redacción en una fecha» (l. 173–174) | references/redaccion_en_una_fecha.md, «Un año o un mes» |
| B86 | Si las dos dan igual fecha de vigencia y norma, o las dos el `3`, la línea de cada una; si no, ninguna cita y pedir el día. | SKILL.md, «Redacción en una fecha» (l. 174–175) | references/redaccion_en_una_fecha.md, «Un año o un mes» |
| B87 | Lo leído con fecha se cita como en «Cómo se cita», con sus avisos, que son los de la norma hoy. | SKILL.md, «Redacción en una fecha» (l. 176–177) | references/redaccion_en_una_fecha.md, «Lo que se cita» |
| B88 | Lo leído con fecha lleva la línea `⚠ REDACCIÓN A FECHA:`, con la fecha pedida y la de vigencia como las da la lectura. | SKILL.md, «Redacción en una fecha» (l. 176–179) | SKILL.md, «Redacción en una fecha» (forma); references/redaccion_en_una_fecha.md, «Lo que se cita» (cuándo y con qué datos) |
| B89 | Con `redaccion_posterior` en `data`, va también `⚠ REDACCIÓN DE HOY:`, con su norma y su fecha de vigencia. | SKILL.md, «Redacción en una fecha» (l. 181–185) | SKILL.md, «Redacción en una fecha» (forma); references/redaccion_en_una_fecha.md, «La redacción de hoy» |
| B90 | Esa línea es el único sitio para la norma y la fecha de la redacción de hoy; otra frase no la sustituye ni añade nada. | SKILL.md, «Redacción en una fecha» (l. 186–187) | references/redaccion_en_una_fecha.md, «La redacción de hoy» |
| B91 | Qué dice, qué cambió o si rige ya la redacción de hoy no se ha leído, salvo leyéndola sin fecha. | SKILL.md, «Redacción en una fecha» (l. 188) | references/redaccion_en_una_fecha.md, «La redacción de hoy» |
| B92 | Con `3` y la fecha de la primera redacción en el mensaje, no citar su texto como el de entonces ni decir qué regía. | SKILL.md, «Redacción en una fecha» (l. 190–192) | references/redaccion_en_una_fecha.md, «Sin redacción vigente» |
| B93 | En ese caso, la línea `⚠ SIN REDACCIÓN VIGENTE:`, con la fecha del mensaje. | SKILL.md, «Redacción en una fecha» (l. 192–195) | SKILL.md, «Redacción en una fecha» (forma); references/redaccion_en_una_fecha.md, «Sin redacción vigente» |
| B94 | Con otro `3` en una lectura con fecha, vale la regla 2. | SKILL.md, «Redacción en una fecha» (l. 192) | references/redaccion_en_una_fecha.md, «Sin redacción vigente» |
| B95 | Las tres líneas fechadas salen solo de una lectura con fecha, cada una con su forma fija. | SKILL.md, «Redacción en una fecha» (l. 196) | SKILL.md, «Redacción en una fecha» |
| B96 | De `kitlegal graph`, el protocolo solo usa `check`, tras cada lectura sin fecha. | SKILL.md, «Redacción modificada» (l. 203–204) | SKILL.md, «Redacción modificada»; «Protocolo», paso 2 |
| B97 | Con `version-obsoleta`, una línea `⚠ REDACCIÓN MODIFICADA:` por bloque, con su cita y las dos fechas: la superada y después la leída. | SKILL.md, «Redacción modificada» (l. 206–211) | SKILL.md, «Redacción modificada» |
| B98 | Con dos bloques en `version-obsoleta`, dos líneas, cada una con su cita y sus fechas. | SKILL.md, «Redacción modificada» (l. 212) | SKILL.md, «Redacción modificada» («por cada entrada»); references/redaccion_modificada.md, «La línea, por bloque» |
| B99 | Decirlo con otras palabras no vale, ni en lugar de la línea ni además de ella. | SKILL.md, «Redacción modificada» (l. 212–213) | references/redaccion_modificada.md, «La línea, por bloque»; SKILL.md, «Cómo se cita» (B80) |
| B100 | Una entrada `fuente-caducada` no va en la respuesta, que cita el texto leído. | SKILL.md, «Redacción modificada» (l. 214) | SKILL.md, «Redacción modificada» |
| B101 | La redacción modificada no es un aviso de vigencia: es un dato de kitlegal sobre otras conversaciones. | SKILL.md, «Redacción modificada» (l. 215–217) | references/redaccion_modificada.md, «No es un aviso de vigencia» |
| B102 | Un bloque sin entrada `version-obsoleta` no da nada a la respuesta: ni línea, ni frase junto a avisos o fecha, ni palabra al final. | SKILL.md, «Redacción modificada» (l. 217–218) | SKILL.md, «Redacción modificada» |
| B103 | La línea es todo lo que la respuesta dice de otras conversaciones y de `kitlegal graph check`. | SKILL.md, «Redacción modificada» (l. 219–220) | SKILL.md, «Redacción modificada» («Nada más de `kitlegal graph`…») |
| B104 | La redacción superada no se ha leído: no decir, resumir ni comparar lo que decía, aunque se crea saberlo. | SKILL.md, «Redacción modificada» (l. 221–224) | references/redaccion_modificada.md, «La redacción superada» |
| B105 | Si se pregunta qué cambió, decir que se cita la redacción vigente y que la de antes no se ha leído. | SKILL.md, «Redacción modificada» (l. 224) | references/redaccion_modificada.md, «Si se pregunta qué cambió»; SKILL.md, «Redacción modificada» (cuándo) |
| B106 | `kitlegal` se invoca desde el `PATH`. | SKILL.md, «Comandos» (l. 228) | SKILL.md, «Comandos» |
| B107 | Cada herramienta, por el nombre de su fila en la tabla de comandos. | SKILL.md, «Comandos» (l. 228) | SKILL.md, «Protocolo» (párrafo inicial); references/con_herramientas.md, «Nombres» |
| B108 | Cada orden termina con `0` si todo fue bien y, si no, con el código de lo que falló; la herramienta lo da en `data.clase`. | SKILL.md, «Comandos» (l. 228–230) | references/lecturas_que_fallan.md, «Los códigos»; references/con_herramientas.md, «Encadenar y fallar» |
| B109 | Códigos: `2` argumentos, `3` no-encontrado, `4` fuente-no-disponible, `5` limite-o-tos, `6` identidad-humana, `1` inesperado (`world.db` ilegible). | SKILL.md, «Comandos» (l. 230–233) | references/lecturas_que_fallan.md, «Los códigos» |
| B110 | Una búsqueda vacía o una norma fuera de la tabla no prueba que no exista: «no encontrada con esta búsqueda» y proponer reformularla. | SKILL.md, «Reglas», 1 (l. 264–265) | SKILL.md, «Reglas», 1 |
| B111 | Nunca inventar contenido legal ni suplir el texto con conocimiento propio. | SKILL.md, «Reglas», 2 (l. 266, 270) | SKILL.md, «Reglas», 2 |
| B112 | Si `kitlegal boe` falla (`3`, `4`, `5`, o sin caché con `--offline`) o no está, decir qué bloques no se consultaron, por su número. | SKILL.md, «Reglas», 2 (l. 266–268) | SKILL.md, «Reglas», 2; references/lecturas_que_fallan.md, «Los códigos» (qué es fallar) |
| B113 | Decir por qué: no están en la norma, la fuente no estaba disponible, limitó las consultas o no fecha sus redacciones. | SKILL.md, «Reglas», 2 (l. 268–269) | SKILL.md, «Reglas», 2 («y por qué»); references/lecturas_que_fallan.md, «Qué dice la respuesta» |
| B114 | Sin el código, sin decir de qué tratan los bloques ni para qué se pedían. | SKILL.md, «Reglas», 2 (l. 269–270) | SKILL.md, «Reglas», 2; references/lecturas_que_fallan.md, «Qué dice la respuesta» (para qué se pedían) |
| B115 | Si la orden que falló pedía varios bloques, decirlo solo tras pedir cada uno por separado con `kitlegal boe articulo`. | SKILL.md, «Reglas», 2 (l. 270–271) | references/lecturas_que_fallan.md, «Una lectura de varios bloques» |
| B116 | Sin la herramienta ni el binario, vale la regla 8. | SKILL.md, «Reglas», 2 (l. 271) | SKILL.md, «Reglas», 2 |
| B117 | No presentar como vigente el texto de una norma derogada. | SKILL.md, «Reglas», 3 (l. 274) | SKILL.md, «Reglas», 3 |
| B118 | Los textos consolidados del BOE tienen carácter informativo y no son asesoramiento. | SKILL.md, «Reglas», 3 (l. 275) | SKILL.md, «Reglas», 3 |
| B119 | No presentar, notificar, firmar ni tramitar nada en nombre de nadie, ni simularlo. | SKILL.md, «Reglas», 4 (l. 276–277) | SKILL.md, «Reglas», 4 |
| B120 | Si la pregunta lo pide, decir que es una acción de la persona y citar, si procede, la norma aplicable. | SKILL.md, «Reglas», 4 (l. 277) | SKILL.md, «Reglas», 4 |
| B121 | Ningún caso especial para un territorio: sin reglas propias de un municipio o una comunidad. | SKILL.md, «Reglas», 5 (l. 278–279) | SKILL.md, «Reglas», 5 |
| B122 | Por un municipio, responder con la normativa estatal o autonómica consolidada en el BOE. | SKILL.md, «Reglas», 5 (l. 279–280) | SKILL.md, «Reglas», 5 |
| B123 | El texto citado sale siempre de `kitlegal boe articulo` o `articulos`. | SKILL.md, «Reglas», 6 (l. 281–282) | SKILL.md, «Reglas», 6 |
| B124 | Nunca citar, parafrasear ni reconstruir texto con la salida de `kitlegal graph`: dice qué volver a comprobar, no qué dice el artículo. | SKILL.md, «Reglas», 6 (l. 282–283) | SKILL.md, «Reglas», 6 |
| B125 | Si la lectura sin fecha da el texto y la comprobación termina con otro código (o `graph_check` da error), citar igual lo leído. | SKILL.md, «Reglas», 7 (l. 284–286) | SKILL.md, «Reglas», 7; references/lecturas_que_fallan.md, «Si `kitlegal graph check` falla» |
| B126 | Y llevar la frase «No se ha podido comprobar…», tal cual y sin nada más sobre la redacción. | SKILL.md, «Reglas», 7 (l. 286–289) | SKILL.md, «Reglas», 7 (sin la frase); references/lecturas_que_fallan.md, «Si `kitlegal graph check` falla» (la frase, en un bloque `text`) |
| B127 | La frase es solo de ese caso: con `0` no va, ni afirmada, ni negada, ni con otras palabras. | SKILL.md, «Reglas», 7 (l. 290–291) | SKILL.md, «Reglas», 7 («y solo entonces»); references/lecturas_que_fallan.md, «Si `kitlegal graph check` falla» |
| B128 | Si lo que falla es la lectura, la orden termina ahí, sin la comprobación, y vale la regla 2. | SKILL.md, «Reglas», 7 (l. 291–292) | references/lecturas_que_fallan.md, «Si `kitlegal graph check` falla» |
| B129 | La línea `⚠ REDACCIÓN MODIFICADA:` solo va cuando la comprobación termina con `0` y trae `version-obsoleta`. | SKILL.md, «Reglas», 7 (l. 292) | SKILL.md, «Redacción modificada» («por cada entrada…; si falla, la regla 7») |
| B130 | Sin herramienta, si la orden falla porque `kitlegal` no está (shell sin él, o sin poder ejecutar órdenes), no buscarlo en otras carpetas ni en un plugin. | SKILL.md, «Reglas», 8 (l. 293–295) | SKILL.md, «Reglas», 8 |
| B131 | Entonces no se ha consultado nada: no afirmar nada de la norma, ni de memoria ni con salvedades, ni escribir ninguna cita. | SKILL.md, «Reglas», 8 (l. 295–296) | SKILL.md, «Reglas», 8 |
| B132 | Llevar la línea `⚠ SIN CONSULTA AL BOE:` con la causa en el marcador, sola en un bloque `text`. | SKILL.md, «Reglas», 8 (l. 296–299) | SKILL.md, «Reglas», 8 |

No son filas (explicación sin regla): l. 41–42 («las normas que esta skill consulta a menudo»), l. 72–74 (los dos sobres
que devuelve la lectura, salvo el reenvío a la regla 7, B33), l. 77 («los devuelve en el orden pedido», que pasa a
references/varios_bloques.md), l. 98 («quien pregunta no lo ve»), l. 162–163 (que la línea `⚠` la lee un programa,
que queda como razón de B80), l. 200–203 (qué recuerda kitlegal y qué compara la comprobación, que pasa a
references/redaccion_modificada.md, «Qué compara»), l. 296 («quien acaba de instalar suele tenerlo y sin arrancar
aún», que se retira) y la región generada, l. 235–260, que no cambia. Los ejemplos que cambian están en las notas de
la estructura.
### legal-core

Una fila por regla de `skills/legal-core/SKILL.md` en `main` (6ea709d, igual en la rama del hito; 300 líneas:
frontmatter l. 1–16, cuerpo l. 17–300). «En la nueva» remite a la estructura nueva de legal-core: `SKILL.md` con los
apartados «intro», «Referencias», «Protocolo, paso N», «Cómo se presenta el territorio», «Cómo se presenta un plazo»,
«Comandos» y «Reglas, N»; y dos referencias a mano, `references/municipios.md` y `references/calculo_de_plazos.md`.
Donde una regla queda en el cuerpo y su detalle o su porqué en una referencia, se dan los dos sitios.

| Id | Regla | En main | En la nueva |
|---|---|---|---|
| L01 | Se activa como punto de partida de las preguntas de derecho público español que dependen de un municipio o de un plazo. | frontmatter, description (l. 4) | frontmatter, description (1.ª frase) |
| L02 | Identifica con kitlegal el territorio de un municipio: código INE, provincia, comunidad, régimen, DIR3 del ayuntamiento y boletines. | frontmatter, description (l. 4–5) | frontmatter, description |
| L03 | Calcula con kitlegal el último día para recurrir, de un plazo administrativo o del contencioso y el silencio, con los festivos de los municipios que cuentan. | frontmatter, description (l. 5–8) | frontmatter, description (capacidad y lista «Se usa cuando…») |
| L04 | Dice cuándo un asunto lo regula la ley aparte: tributos, multas de tráfico, Seguridad Social, extranjería. | frontmatter, description (l. 8–9) | frontmatter, description |
| L05 | Razona con la jerarquía normativa y las leyes vertebrales (Constitución, LPAC, LRJSP, LRBRL, TRLRHL, LCSP, LGT…). | frontmatter, description (l. 9–10) | frontmatter, description («Constitución» pasa a «CE») |
| L06 | Activación: hasta cuándo recurrir o qué plazo tiene un recurso, vencimiento o silencio, comunidad, provincia o boletines de un ayuntamiento, foral, qué norma prevalece. | frontmatter, description (l. 10–12) | frontmatter, description («Se usa cuando se pregunta…») |
| L07 | Para el texto de un artículo se apoya en boe-legislacion. | frontmatter, description (l. 12) | frontmatter, description |
| L08 | Declara los applets `territorio` y `plazos` y las cuatro referencias generadas. | frontmatter, metadata (l. 13–15) | frontmatter, metadata (sin cambio, FR-024) |
| L09 | Identifica con `kitlegal territorio`, calcula con `kitlegal plazos` y razona con cuatro referencias: identificador, nivel, tipo de plazo y lo regulado aparte. | intro (l. 20–24) | SKILL.md, intro + Referencias (las cuatro líneas generadas) |
| L10 | Identifica, calcula y orienta: no tramita nada ni sustituye el asesoramiento de un profesional. | intro (l. 25) | SKILL.md, intro |
| L11 | Una pregunta de territorio sigue los pasos 1 a 4, y el 5 y el 6 si va más allá. | Protocolo (l. 29) | SKILL.md, Protocolo (párrafo inicial) |
| L12 | Una de plazos o recursos empieza por el 7, que decide antes de preguntar nada; solo si hay plazo, el 8 (con 1–4) y 9 a 11. | Protocolo (l. 29–31) | SKILL.md, Protocolo (párrafo inicial y pasos 7 y 8) |
| L13 | Las dos vías devuelven el mismo sobre: `fuente`, `url`, `fecha_consulta`, `hash` y `data`. | Protocolo (l. 31–32) | SKILL.md, Comandos (región generada, «devuelven el mismo sobre») |
| L14 | Si entre tus herramientas está la de su fila, sola o con el prefijo del agente (`mcp__kitlegal__territorio_resolver`), úsala siempre. | Protocolo (l. 34–35) | SKILL.md, Comandos (párrafo) |
| L15 | Sin esa herramienta, la orden `kitlegal … --json`. | Protocolo (l. 36) | SKILL.md, Comandos (párrafo) |
| L16 | Donde un paso dice que la orden termina con un código, la herramienta da un error con su `data.clase`. | Protocolo (l. 38) | SKILL.md, Comandos (párrafo) |
| L17 | Toma el municipio de la conversación, por nombre o código INE, tal como lo dio la persona. | 1. Identificar el territorio (l. 42) | SKILL.md, Protocolo, paso 1 (y paso 2, `<nombre o código INE>`) |
| L18 | Si la conversación no lo dice, pregúntalo y espera; no lo supongas. | 1. Identificar el territorio (l. 43) | SKILL.md, Protocolo, paso 1 |
| L19 | No lo deduzcas de otros datos ni sigas con un municipio de ejemplo ni con el de quien usa el kit. | 1. Identificar el territorio (l. 43–44) | SKILL.md, Protocolo, paso 1 (el de quien usa el kit) + references/municipios.md, El municipio que falta |
| L20 | Si solo se nombra una provincia o una comunidad, pregunta también por el municipio. | 1. Identificar el territorio (l. 44–45) | references/municipios.md, El municipio que falta |
| L21 | Resuelve con `territorio_resolver` (en `consulta`) o `kitlegal territorio resolver <nombre o código INE> --json`. | 2. Resolverlo (l. 49–50) | SKILL.md, Protocolo, paso 2 |
| L22 | Pasa el nombre o el código tal como los dio la persona; no conviertas de memoria uno en otro. | 2. Resolverlo (l. 52) | SKILL.md, Protocolo, paso 2 («tal como lo dio») + references/municipios.md, El municipio que falta |
| L23 | Ningún dato de territorio se da por sabido ni de memoria: todos salen de `data` en esta conversación. | 2. Resolverlo (l. 53–54) | SKILL.md, Reglas, 1 + references/municipios.md, Lo que trae `data` |
| L24 | Ningún municipio ni ninguna comunidad tiene un trato propio. | 2. Resolverlo (l. 54) | references/municipios.md, Lo que trae `data` |
| L25 | Con 0, `data` trae siempre municipio, código INE, provincia, comunidad, DIR3, régimen, boletines y cobertura, cada dato con su `source`. | 2. Resolverlo (l. 55–56) | references/municipios.md, Lo que trae `data` |
| L26 | Sin herramienta ni binario, vale la regla 7. | 2. Resolverlo (l. 57) | SKILL.md, Protocolo, paso 2 |
| L27 | Qué dicen los tres aspectos de la cobertura (`boletin_autonomico`, `boletin_provincial`, `dir3`) y sus valores. | 3. Leer `cobertura` (l. 61–65) | references/municipios.md, Qué dice cada aspecto de la cobertura |
| L28 | Responde con lo que dice `data` y escribe la cobertura, con sus tres aspectos, en su forma fija. | 3. Leer `cobertura` (l. 67–68) | SKILL.md, Protocolo, paso 3 + Cómo se presenta el territorio |
| L29 | Lo `no-configurado` o `no-verificado`, dicho con sus palabras, no significa que no exista. | 3. Leer `cobertura` (l. 68–69) | SKILL.md, Reglas, 5 + references/municipios.md, Qué dice cada aspecto de la cobertura |
| L30 | No nombres ningún boletín que el applet no haya devuelto, aunque creas conocerlo. | 3. Leer `cobertura` (l. 70) | SKILL.md, Protocolo, paso 3 + references/municipios.md, Qué dice cada aspecto de la cobertura («aunque creas conocerlo») |
| L31 | La clase de boletín de `jerarquia_normativa` no suple un nivel `no-configurado`. | 3. Leer `cobertura` (l. 70–71) | references/municipios.md, Qué dice cada aspecto de la cobertura |
| L32 | Con `dir3` en `no-verificado`, no compongas el DIR3 a partir del código INE. | 3. Leer `cobertura` (l. 72) | references/municipios.md, Qué dice cada aspecto de la cobertura (y la plantilla, «no verificado») |
| L33 | Si `regimen.valor` es `foral`, dilo. | 3. Leer `cobertura` (l. 72–73) | SKILL.md, Reglas, 4 + references/municipios.md, La plantilla del territorio |
| L34 | Di de cuándo son los datos: `fecha_consulta`. | 3. Leer `cobertura` (l. 73) | SKILL.md, Protocolo, paso 3 |
| L35 | Código 2 con candidatos (`<código INE> <nombre> (<provincia>)`): ofrécelos, pregunta cuál y vuelve al paso 2 con el elegido. | 4. Ambigüedad y ausencia (l. 77–78) | SKILL.md, Protocolo, paso 4 (remite) + references/municipios.md, Cuando no sale con 0 |
| L36 | Código 2 sin candidatos: código INE no válido o dígito de control no oficial; dilo, pide nombre o código, no lo corrijas. | 4. Ambigüedad y ausencia (l. 79–80) | references/municipios.md, Cuando no sale con 0 |
| L37 | Código 3: no está en la relación; dilo sin concluir que no exista y pide otra forma o el código INE. | 4. Ambigüedad y ausencia (l. 81–82) | references/municipios.md, Cuando no sale con 0 (+ SKILL.md, Reglas, 5) |
| L38 | Con cualquier otro código o `clase`, di que no se pudo resolver y no suplas sus datos. | 4. Ambigüedad y ausencia (l. 83) | SKILL.md, Protocolo, paso 4 («sin suplir lo que falte») + references/municipios.md, Cuando no sale con 0 |
| L39 | Nombra cada norma con el identificador de `leyes_vertebrales`, con su rango. | 5. Razonar con las referencias (l. 87) | SKILL.md, Referencias (línea de leyes_vertebrales.md) + Protocolo, paso 5 + Reglas, 3 |
| L40 | Usa `jerarquia_normativa` para decir en qué nivel se regula lo preguntado y qué norma prevalece. | 5. Razonar con las referencias (l. 88) | SKILL.md, Referencias (línea de jerarquia_normativa.md) + Protocolo, paso 5 |
| L41 | El nivel autonómico es el de la `comunidad` de `data`; los boletines concretos, solo los de `boletines`. | 5. Razonar con las referencias (l. 88–89) | SKILL.md, Referencias (línea de jerarquia_normativa.md) + Protocolo, paso 3 + references/municipios.md, Qué dice cada aspecto de la cobertura |
| L42 | Afirmar lo que dice un artículo exige leerlo con boe-legislacion en esta conversación (`boe articulo`/`boe_articulo`) y citarlo como ella. | 6. Delegar el texto (l. 93–94) | SKILL.md, Protocolo, paso 6 + Reglas, 2 (la orden con `--fecha`, en el paso 10) |
| L43 | No lo cites de memoria ni desde las referencias, que no llevan el texto de ninguna norma. | 6. Delegar el texto (l. 94–95) | SKILL.md, Reglas, 1 y 2 |
| L44 | El identificador de una norma que no está en las referencias se resuelve con boe-legislacion. | 6. Delegar el texto (l. 96) | SKILL.md, Protocolo, paso 6 |
| L45 | Con lo que ya dice la conversación del acto, la administración y la materia, decide antes de preguntar nada, ni municipio ni fecha. | 7. Decidir si hay plazo (l. 100–102) | SKILL.md, Protocolo, paso 7 |
| L46 | Si la materia está en `cobertura_por_materia`, declara que no está cubierto con la línea `⚠ NO CUBIERTO:`. | 7. Decidir si hay plazo (l. 104–107) | SKILL.md, Protocolo, paso 7 + Referencias (línea de cobertura_por_materia.md, que lleva la lista de materias) |
| L47 | No pidas `plazos calcular` ni des la línea del plazo ni ninguna fecha: el plazo del común no es el suyo. | 7. Decidir si hay plazo (l. 107–108) | SKILL.md, Protocolo, paso 7 + Reglas, 9 (el porqué, references/calculo_de_plazos.md, Lo que la ley regula aparte) |
| L48 | Tras la línea, no expongas la norma que rige el procedimiento —plazos, recursos—: la nombra la línea, con lo que trae la entrada. | 7. Decidir si hay plazo (l. 108–109) | SKILL.md, Reglas, 9 («ni se exponen los de la norma que lo rige») + references/calculo_de_plazos.md, Lo que la ley regula aparte |
| L49 | La línea va aunque falle una lectura: no afirma qué dice un precepto, sino qué no calcula kitlegal. | 7. Decidir si hay plazo (l. 109–110) | SKILL.md, Reglas, 9 («falle o no una lectura») + references/calculo_de_plazos.md, Lo que la ley regula aparte |
| L50 | La lectura fallida se dice aparte, como boe-legislacion, sin lo que dice el bloque ni la línea de la regla 7 (ejemplo de cuotas con `daprimera`). | 7. Decidir si hay plazo (l. 110–112) | references/calculo_de_plazos.md, Lo que la ley regula aparte (con el ejemplo) |
| L51 | Si la conversación no dice la materia, no la preguntes ni detengas la respuesta por ella: sigue con el 8. | 7. Decidir si hay plazo (l. 112–113) | SKILL.md, Protocolo, paso 7 |
| L52 | Al responder, di que la fecha no vale si el asunto es de los regulados aparte, nombrándolos solo con la referencia. | 7. Decidir si hay plazo (l. 113–114) | SKILL.md, Protocolo, paso 7 + references/calculo_de_plazos.md, Lo que la ley regula aparte («solo con la referencia») |
| L53 | Un acto presunto no tiene fecha: no pidas ningún plazo. | 7. Decidir si hay plazo (l. 115–116) | SKILL.md, Protocolo, paso 7 |
| L54 | En el acto presunto, cita los arts. 122.1 y 124.1 LPAC y 46.1 LJCA, leídos con boe-legislacion. | 7. Decidir si hay plazo (l. 115–116) | references/calculo_de_plazos.md, El acto presunto |
| L55 | La fecha de inicio es la de notificación o publicación; en el silencio, la del acuerdo de iniciación o la entrada en el registro. | 8. Situar el plazo (l. 120–121) | SKILL.md, Protocolo, paso 8 |
| L56 | Si falta la fecha, pregúntala: `plazos` no la deduce y tú tampoco. | 8. Situar el plazo (l. 121) | SKILL.md, Protocolo, paso 8 |
| L57 | Cuentan el municipio del interesado y el de la sede del órgano (pueden coincidir); en el contencioso, el de la sede del juzgado. | 8. Situar el plazo (l. 122–123) | SKILL.md, Protocolo, paso 8 (+ «pueden ser el mismo», references/calculo_de_plazos.md, Los municipios que cuentan) |
| L58 | Resuélvelos con los pasos 1 a 4; si no los dio, pregúntalos; nunca supongas el de quien usa el kit. | 8. Situar el plazo (l. 123–124) | SKILL.md, Protocolo, pasos 8 y 1 + references/calculo_de_plazos.md, Los municipios que cuentan |
| L59 | Elige el tipo en `recursos_y_plazos` y pide `plazos_calcular` o `kitlegal plazos calcular <tipo> <inicio> <municipios>... --json`, municipios en su orden. | 9. Pedir el plazo (l. 128–130) | SKILL.md, Protocolo, paso 9 + Referencias (línea de recursos_y_plazos.md) |
| L60 | Solo municipios que `territorio` haya resuelto en esta conversación, como los dio la persona: sin el paso 8 faltan territorio y cobertura. | 9. Pedir el plazo (l. 132–134) | SKILL.md, Protocolo, paso 9 + references/calculo_de_plazos.md, Los municipios que cuentan |
| L61 | Silencio: el plazo máximo lo fija la norma del procedimiento; léela con boe-legislacion y pásalo con `--duracion` y `--unidad`. | 9. Pedir el plazo (l. 135–136) | references/calculo_de_plazos.md, El silencio |
| L62 | Solo si la norma no fija ninguno, los tres meses del art. 21.3 LPAC, dicho con su cita. | 9. Pedir el plazo (l. 136) | references/calculo_de_plazos.md, El silencio |
| L63 | `plazos` no dice si el silencio es estimatorio o desestimatorio: lo dice la norma. | 9. Pedir el plazo (l. 137) | references/calculo_de_plazos.md, El silencio |
| L64 | Reclamación del art. 24.2 LTAIBG: señala como variación autonómica ante quién se reclama (CTBG u órgano autonómico, art. 24.6). | 9. Pedir el plazo (l. 138–139) | references/calculo_de_plazos.md, La reclamación del art. 24.2 LTAIBG |
| L65 | Con el 2 o el 3 de un municipio en `plazos`, como en el paso 4, con los candidatos tras el municipio. | 9. Pedir el plazo (l. 140) | SKILL.md, Protocolo, paso 4 + references/municipios.md, Cuando no sale con 0 |
| L66 | Con el 3 y un `data.mensaje` que empieza por «el plazo de tipo», no hay fecha; `plazos` no elige qué regía, y tú tampoco. | 9. Pedir el plazo (l. 141–144) | SKILL.md, Protocolo, paso 9 + references/calculo_de_plazos.md, El plazo sin calcular (el porqué) |
| L67 | Da `⚠ PLAZO SIN CALCULAR:`, sin la línea del plazo, sin fecha de vencimiento, sin la norma que regía de memoria y sin el paso 10. | 9. Pedir el plazo (l. 144–145) | SKILL.md, Protocolo, paso 9 (lo de memoria: Reglas, 1 + references/calculo_de_plazos.md, El plazo sin calcular) |
| L68 | Cada precepto de `data.preceptos`, y la norma del silencio, se lee con boe-legislacion en la fecha de inicio pasada a `plazos`. | 10. Leer el precepto (l. 149–151) | SKILL.md, Protocolo, paso 10 (la norma del silencio y el porqué: references/calculo_de_plazos.md, El silencio y La lectura en la fecha de inicio) |
| L69 | Con `boe_articulo` y `fecha` o `kitlegal boe articulo <norma> <bloque> --fecha <inicio> --json`, sin `kitlegal graph check` detrás. | 10. Leer el precepto (l. 151–152) | SKILL.md, Protocolo, paso 10 |
| L70 | Se cita como una redacción a fecha, con su línea `⚠ REDACCIÓN A FECHA:`. | 10. Leer el precepto (l. 152) | SKILL.md, Protocolo, paso 10 + Cómo se presenta un plazo |
| L71 | Si esa lectura sale con 4 o 5, `⚠ PLAZO:` se mantiene, con la fecha de `plazos` y la cita de cada precepto. | 10. Leer el precepto (l. 154–155) | SKILL.md, Protocolo, paso 10 |
| L72 | En lugar de su `⚠ REDACCIÓN A FECHA:`, di aparte que no pudiste leerla, como boe-legislacion; no afirmes qué dice. | 10. Leer el precepto (l. 155–156) | SKILL.md, Protocolo, paso 10 + references/calculo_de_plazos.md, La lectura en la fecha de inicio |
| L73 | Si la lectura trae `redaccion_posterior`, va también su línea `⚠ REDACCIÓN DE HOY:`. | 10. Leer el precepto (l. 157–158) | SKILL.md, Cómo se presenta un plazo + references/calculo_de_plazos.md, La lectura en la fecha de inicio |
| L74 | Con `redaccion_posterior`: en un precepto cambió otra parte del bloque; en la norma del silencio, la duración es la de ese día salvo transitorio no leído. Dilo. | 10. Leer el precepto (l. 158–160) | references/calculo_de_plazos.md, La lectura en la fecha de inicio |
| L75 | Da `data.fecha` en la línea del plazo, con su tipo y una cita por precepto de `data.preceptos`, y ninguna otra fecha de vencimiento. | 11. Responder (l. 164–165) | SKILL.md, Protocolo, paso 11 + Cómo se presenta un plazo |
| L76 | No calcules ni corrijas ninguna fecha, tampoco para «comprobarla». | 11. Responder (l. 165) | SKILL.md, Reglas, 8 (y paso 11, «ninguna otra (regla 8)») |
| L77 | Di qué festivos movieron la fecha, con `data.festivos`: fecha, ámbito y, si es local, municipio. | 11. Responder (l. 166) | SKILL.md, Protocolo, paso 11 |
| L78 | Traslada cada aviso de `data.avisos` con su `texto`, tal cual. | 11. Responder (l. 167) | SKILL.md, Protocolo, paso 11 + Cómo se presenta un plazo |
| L79 | Hasta qué disposición llegan los festivos lo dice el aviso que lo trae; sin aviso, no lo copies. | 11. Responder (l. 168) | SKILL.md, Protocolo, paso 11 («solo si lo dice un aviso») |
| L80 | Cada dato del territorio sale de `data`, como lo da, en la plantilla: municipio, provincia, comunidad y régimen, DIR3, un boletín por entrada, cobertura. | Cómo se presenta el territorio (l. 172–183) | SKILL.md, Cómo se presenta el territorio (las tres líneas de cobertura) + references/municipios.md, La plantilla del territorio (la plantilla entera, sin cambio) |
| L81 | La cobertura va con la clave y el valor exactos de `data`; detrás puede explicarse. | Cómo se presenta el territorio (l. 185) | SKILL.md, Cómo se presenta el territorio |
| L82 | Un boletín se nombra por su `codigo` y su `nombre` exactos. | Cómo se presenta el territorio (l. 185–186) | SKILL.md, Cómo se presenta el territorio |
| L83 | Línea `⚠ PLAZO:` con `data.fecha` tal cual, el tipo como lo nombra `recursos_y_plazos` y la cita de cada precepto leído en el paso 10. | Cómo se presenta un plazo (l. 190–195) | SKILL.md, Cómo se presenta un plazo (forma) + Protocolo, paso 11 + Referencias (recursos_y_plazos.md, «nombrarlo») |
| L84 | Cada precepto en sus propios corchetes, en la misma línea, terminados en `<identificador>, bloque <bloque>]` con un bloque (ejemplo del contencioso). | Cómo se presenta un plazo (l. 197–203) | SKILL.md, Cómo se presenta un plazo (con el ejemplo, sin cambio) |
| L85 | Dos bloques en unos mismos corchetes no son cita de ninguno. | Cómo se presenta un plazo (l. 205–206) | SKILL.md, Cómo se presenta un plazo |
| L86 | Tras la línea del plazo, `⚠ REDACCIÓN A FECHA:` por precepto leído; con `redaccion_posterior`, `⚠ REDACCIÓN DE HOY:` con su norma y vigencia. | Cómo se presenta un plazo (l. 208–215) | SKILL.md, Cómo se presenta un plazo (las dos formas, sin cambio) |
| L87 | Con el 3 de un plazo sin calcular, en lugar de la línea del plazo, `⚠ PLAZO SIN CALCULAR:` con tipo, inicio y cada línea del mensaje. | Cómo se presenta un plazo (l. 217–222) | SKILL.md, Cómo se presenta un plazo (forma, sin cambio) |
| L88 | `⚠ NO CUBIERTO:` con el procedimiento de `cobertura_por_materia`: la primera forma si la entrada trae la norma que lo rige, la segunda si no. | Cómo se presenta un plazo (l. 224–230) | SKILL.md, Cómo se presenta un plazo (las dos formas, sin cambio) |
| L89 | Cada aviso, el `texto` de `data.avisos`, empieza por una de las cuatro formas de aviso. | Cómo se presenta un plazo (l. 232–239) | SKILL.md, Cómo se presenta un plazo (las cuatro, sin cambio) |
| L90 | Cada línea `⚠` va entera, en una sola línea y con su forma, porque la lee un programa; detrás puede explicarse. | Cómo se presenta un plazo (l. 241–243) | SKILL.md, Cómo se presenta un plazo (párrafo final) |
| L91 | `kitlegal` se invoca desde el `PATH`, y la herramienta, por el nombre de su fila. | Comandos (l. 247) | SKILL.md, Comandos (párrafo) |
| L92 | Los dos applets responden con lo que lleva el binario, sin consultar ninguna fuente. | Comandos (l. 247–248) | references/municipios.md, Lo que trae `data` |
| L93 | Códigos: 0 correcto; 2 (`argumentos`), también más de un municipio; 3 (`no-encontrado`), municipio fuera de la relación o plazo con preceptos que no regían. | Comandos (l. 248–250) | SKILL.md, Comandos (párrafo) |
| L94 | La tabla de comandos se genera desde `--describe` con `make skills-sync` y no se edita. | Comandos, región generada (l. 252–270) | SKILL.md, Comandos, región generada (sin cambio, 1498 caracteres) |
| L95 | Regla 1: nunca inventar ni citar de memoria texto, identificador, dato de territorio ni fecha; lo que no salga de kitlegal, referencias o boe-legislacion no se afirma. | Reglas, 1 (l. 274–276) | SKILL.md, Reglas, 1 |
| L96 | Regla 2: cada afirmación sobre una norma, con su identificador, de las referencias o de boe-legislacion; el texto, solo de boe-legislacion. | Reglas, 2 (l. 277–278) | SKILL.md, Reglas, 2 |
| L97 | Regla 3: el rango de cada norma nombrada; un reglamento nunca puede contradecir la ley. | Reglas, 3 (l. 279–280) | SKILL.md, Reglas, 3 |
| L98 | Regla 4: señalar lo que una comunidad puede haber regulado de otro modo, más si es foral; las normas locales no están en las referencias. | Reglas, 4 (l. 281–282) | SKILL.md, Reglas, 4 |
| L99 | Regla 5: no concluir «no existe» sin cobertura completa: municipio fuera, boletín, DIR3, festivos que faltan o norma ausente de una referencia. | Reglas, 5 (l. 283–285) | SKILL.md, Reglas, 5 |
| L100 | Regla 6: ninguna acción con identidad, ni simulada; si se pide, decir que la hace la persona. | Reglas, 6 (l. 286–287) | SKILL.md, Reglas, 6 |
| L101 | Regla 7: sin herramienta y con la orden fallida porque falta kitlegal, no buscarlo en otras carpetas ni en un plugin: no se ha consultado nada. | Reglas, 7 (l. 288–289) | SKILL.md, Reglas, 7 |
| L102 | Regla 7: entonces, ningún dato de territorio, fecha, plazo ni nada de una norma, ni de memoria ni con salvedades, y ninguna cita. | Reglas, 7 (l. 289–291) | SKILL.md, Reglas, 7 («de memoria», en la regla 1) |
| L103 | Regla 7: la línea `⚠ SIN CONSULTA AL BOE:` con la causa, exacta y sola en un bloque `text`. | Reglas, 7 (l. 291–295) | SKILL.md, Reglas, 7 (línea y bloque sin cambio) |
| L104 | Regla 8: la fecha de un plazo es `data.fecha`; ni contarla ni retocarla. | Reglas, 8 (l. 297) | SKILL.md, Reglas, 8 |
| L105 | Regla 9: lo regulado aparte se declara no cubierto, falle o no una lectura, y no se pide ningún plazo. | Reglas, 9 (l. 298–299) | SKILL.md, Reglas, 9 |

105 reglas. Ninguna regla nueva: lo que la versión nueva escribe distinto (la regla 9 con «ni se exponen los de la norma
que lo rige», la regla 8 con «tampoco para «comprobarla»», el paso 11 con los festivos) es L48, L76 y L79 cambiados de
sitio, no reglas añadidas.
### jurisprudencia

Fuente: `skills/jurisprudencia/SKILL.md` de `main` (208 líneas; cuerpo de 193 líneas y 13 637 runas, con la región
generada de 1 302). Números de línea del fichero entero, frontmatter incluido. «paso N» es el apartado `### N.` del
protocolo, con el mismo título en las dos versiones. 74 reglas. No son filas los ejemplos ni las explicaciones sin
regla (el porqué del CAPTCHA, «Un ECLI que recuerdas no es un dato», «Sería jurisprudencia sin fuente», «esa línea
anuncia una consulta en el CENDOJ»…): siguen donde están, salvo las que acompañan a lo que va a las referencias (el
porqué de pasar solo la ficha, el de los ROJ correlativos y el ejemplo 1088/2023 frente a 3144/2023).

| Id | Regla | En main | En la nueva |
|---|---|---|---|
| J01 | Se activa cuando la pregunta nombra, pide, cita o resume una sentencia o un auto, o pregunta si existe o qué dice. | frontmatter, description (l. 4–6) | frontmatter, description |
| J02 | Se activa también cuando pide jurisprudencia sobre una materia o trae el texto o el PDF de una resolución. | frontmatter, description (l. 6–7) | frontmatter, description |
| J03 | Se activa también cuando el agente cree conocerla: ninguna sentencia se cita de memoria, solo con su documento delante. | frontmatter, description (l. 4, 7–8) | frontmatter, description |
| J04 | kitlegal no consulta el buscador del CENDOJ: prepara la consulta para que la persona busque con su navegador y coteja el documento. | frontmatter, description (l. 8–9) | frontmatter, description |
| J05 | Se activa antes de llamar a `cita_preparar` o a `cita_cotejar`. | frontmatter, description (l. 9–11) | frontmatter, description |
| J06 | kitlegal no consulta el buscador de jurisprudencia del CENDOJ; el CAPTCHA que opone a los programas no se sortea. | SKILL.md, entrada bajo el título (l. 18–19) | SKILL.md, entrada bajo el título |
| J07 | La respuesta no anuncia a la persona ningún CAPTCHA ni le pide que resuelva ninguno. | SKILL.md, entrada bajo el título (l. 20–21) | SKILL.md, entrada bajo el título |
| J08 | La persona busca con su navegador; la skill le da la consulta (dirección, casillas y valores) y comprueba que el documento traído es el pedido. | SKILL.md, entrada bajo el título (l. 21–23) | SKILL.md, entrada bajo el título |
| J09 | Solo se cita la sentencia cuyo documento está en la conversación. | SKILL.md, entrada bajo el título (l. 23–24) | SKILL.md, entrada bajo el título |
| J10 | La skill prepara, coteja y cita: no tramita nada ni sustituye el asesoramiento de un profesional. | SKILL.md, entrada bajo el título (l. 24–25) | SKILL.md, entrada bajo el título |
| J11 | Seguir los pasos que pida la pregunta. | SKILL.md, «Protocolo» (l. 29) | SKILL.md, «Protocolo» |
| J12 | Si el agente tiene la herramienta (`cita_preparar`, `cita_cotejar`, sola o tras el prefijo de su agente), la usa siempre. | SKILL.md, «Protocolo» (l. 31–34) | SKILL.md, «Protocolo» |
| J13 | Argumentos de la herramienta por su nombre (`ecli`, `roj`, `resolucion`, `fecha`, `texto`, `documento`), solo los usados; uno vacío es error de argumentos. | SKILL.md, «Protocolo» (l. 32–34) | SKILL.md, «Protocolo» |
| J14 | Sin herramienta, la orden `kitlegal cita <verbo> … --json`, siempre con `--json`. | SKILL.md, «Protocolo» (l. 35) | SKILL.md, «Protocolo» |
| J15 | Con la herramienta, el código de salida de la orden llega como resultado marcado como error, con su `data.clase`. | SKILL.md, «Protocolo» (l. 37–38) | SKILL.md, «Comandos», párrafo de los códigos (fundida en él: ningún paso del cuerpo nuevo nombra ya un código) |
| J16 | Ninguna de las dos operaciones pide nada a la red: no dicen si una sentencia existe. | SKILL.md, «Protocolo» (l. 38) | SKILL.md, «Protocolo» |
| J17 | Está delante si su documento está en la conversación (texto pegado o PDF leído) con su ficha: `Roj: <ROJ> - <ECLI>` y de `Órgano:` a `Tipo de Resolución:`. | SKILL.md, paso 1 (l. 42–45) | SKILL.md, paso 1 |
| J18 | No está delante la que la persona nombra sin traerla, la que el agente recuerda ni la que otro documento menciona. | SKILL.md, paso 1 (l. 46–47) | SKILL.md, paso 1 |
| J19 | La que no está delante va al paso 2; con un documento, primero el paso 3, antes de decir nada; las del TC, al paso 7. | SKILL.md, paso 1 (l. 48–49) | SKILL.md, paso 1 |
| J20 | Por cada sentencia necesaria que no está delante, al principio de una línea: `⚠ SENTENCIA NO COMPROBADA: <referencia tal como se dio>`. | SKILL.md, paso 2 (l. 53–58) | SKILL.md, paso 2, forma fija con su bloque de ejemplo |
| J21 | Debajo de la línea, la consulta para encontrarla, que sale de `cita_preparar` o de su orden. | SKILL.md, paso 2 (l. 60–61) | SKILL.md, paso 2 |
| J22 | Se pide según se dio la referencia: ECLI en `ecli`; ROJ dicho como tal en `roj`; número y fecha en `resolucion` y `fecha`; con la orden de cada forma. | SKILL.md, paso 2, tabla (l. 63–67) | SKILL.md, paso 2, tabla «Cómo se dio / Qué se pide»; la columna «Orden», en references/consulta_preparada.md, «La orden de cada forma» |
| J23 | «STS 1088/2023, de 4 de julio» lleva el número de resolución, no el ROJ: `fecha` en `AAAA-MM-DD`, con el año del número salvo que diga otro. | SKILL.md, paso 2 (l. 70–71) | SKILL.md, paso 2, primera viñeta |
| J24 | Un número sin fecha no se prepara de ninguna forma, tampoco como ROJ: la respuesta lleva la línea y pide la fecha. | SKILL.md, paso 2, tabla y viñeta (l. 68, 72–73) | SKILL.md, paso 2, fila «Número sin fecha» de la tabla; el porqué (ROJ correlativos), en references/consulta_preparada.md, «La orden de cada forma» |
| J25 | La consulta se da con lo que devuelve la operación, nada de memoria: la dirección y cada casilla con su valor. | SKILL.md, paso 2 (l. 74–75) | SKILL.md, paso 2, segunda viñeta |
| J26 | Cada casilla, con el nombre que ve la persona, su parte si tiene dos («Desde» y «Hasta») y lo que hay que escribir. | SKILL.md, paso 2 (l. 75–76) | references/consulta_preparada.md, «Las casillas» |
| J27 | Decir a la persona qué hacer: abrir la dirección, escribir cada valor, buscar, descargar el documento y traerlo, pegado o adjunto. | SKILL.md, paso 2 (l. 76–77) | SKILL.md, paso 2, segunda viñeta |
| J28 | El equivalente que devuelve la operación puede nombrarse junto a la referencia, dicho como deducido de ella y sin comprobar. | SKILL.md, paso 2 (l. 78–80) | references/consulta_preparada.md, «El equivalente» |
| J29 | El equivalente no se compone: solo se deduce para el Tribunal Supremo, y si no viene no lo hay. | SKILL.md, paso 2 (l. 79–80) | references/consulta_preparada.md, «El equivalente» |
| J30 | La respuesta no dice que la sentencia existe ni que no existe, ni de qué trata, qué resolvió o si es firme. | SKILL.md, paso 2 (l. 81–82) | SKILL.md, paso 2, tercera viñeta |
| J31 | Con un documento en la conversación (pegado o PDF leído), el cotejo es lo primero: su ficha a `cita_cotejar` antes de ninguna frase sobre él. | SKILL.md, paso 3 (l. 86–88) | SKILL.md, paso 3, primera viñeta |
| J32 | Nada del documento antes del cotejo: ni que es la pedida, ni que coincide con ella, ni de qué trata. | SKILL.md, paso 3 (l. 89) | SKILL.md, paso 3, segunda viñeta |
| J33 | Si es o no la pedida lo dice `es_la_pedida`, nunca una comparación a ojo, que falla con 1088/2023 frente al ROJ STS 3144/2023. | SKILL.md, paso 3 (l. 90–93) | SKILL.md, paso 3, tercera viñeta (la regla); el ejemplo, en references/documento_aportado.md, «El documento que no es el pedido» |
| J34 | En `documento` va la ficha y nada más, tal como está, de la línea `Roj:` a la de `Tipo de Resolución:`, con las intermedias. | SKILL.md, paso 3 (l. 95–97) | references/documento_aportado.md, «La ficha y nada más» |
| J35 | Si se pidió una sentencia concreta, se da además la referencia en la forma en que se pidió (`ecli`, `roj`, o `resolucion` con `fecha`). | SKILL.md, paso 3 (l. 97–99) | SKILL.md, paso 3, cuarta viñeta |
| J36 | Con la orden, la ficha va por la entrada estándar (`<<'DOCUMENTO'`), o en `--documento` si el shell no tiene `<<`. | SKILL.md, paso 3 (l. 99–106) | references/documento_aportado.md, «La ficha y nada más» |
| J37 | Un PDF lo lee el agente y pasa su ficha: kitlegal no abre ficheros. | SKILL.md, paso 3 (l. 108) | references/documento_aportado.md, «La ficha y nada más» |
| J38 | El cotejo devuelve `ficha`, `es_la_pedida` (solo si se dio una referencia) y, si no es el pedido, `hallazgos`. | SKILL.md, paso 3 (l. 109–110) | references/documento_aportado.md, «Lo que devuelve» |
| J39 | Si el documento no es el pedido, se dice qué difiere, con las palabras del hallazgo. | SKILL.md, paso 3 (l. 111–112) | references/documento_aportado.md, «El documento que no es el pedido» |
| J40 | Ese documento no se cita como el pedido; la pedida va por el paso 2, con su línea y su consulta, en la forma en que se pidió. | SKILL.md, paso 3 (l. 112–114) | SKILL.md, paso 3, quinta viñeta (la referencia la nombra, sin repetir la línea) |
| J41 | Con `correspondencia` `no-se-corresponden`: no citar; decirlo, dar ROJ y ECLI como están en la ficha y pedir que lo descargue de nuevo. | SKILL.md, paso 3 (l. 115–117) | references/documento_aportado.md, «El ROJ y el ECLI que no se corresponden» |
| J42 | Con cualquier otro valor de `correspondencia`, se cita sin decir nada de esto. | SKILL.md, paso 3 (l. 117–118) | references/documento_aportado.md, «El ROJ y el ECLI que no se corresponden» |
| J43 | Código 2 por ficha no reconocible: si el dato que nombra el mensaje está en el documento, repetir el cotejo con la ficha entera. | SKILL.md, paso 3 (l. 119–120) | references/documento_aportado.md, «Un texto sin ficha reconocible» |
| J44 | Si ese dato no está, no hay sentencia que citar: decir qué falta y pedir el documento completo. | SKILL.md, paso 3 (l. 120–122) | references/documento_aportado.md, «Un texto sin ficha reconocible» |
| J45 | Sobre un texto sin ficha se puede responder, sin atribuirlo a una sentencia identificada. | SKILL.md, paso 3 (l. 122) | references/documento_aportado.md, «Un texto sin ficha reconocible» |
| J46 | La cita tiene forma fija: siglas, número de resolución y fecha, y en la misma línea `[<ECLI>, ROJ: <ROJ>]`. | SKILL.md, paso 4 (l. 126–131) | SKILL.md, paso 4, forma fija con su bloque de ejemplo |
| J47 | Todos los datos de la cita son los de la ficha leída por `cita_cotejar`: siglas del ROJ, número de resolución, fecha de la ficha. | SKILL.md, paso 4 (l. 133–134) | SKILL.md, paso 4 |
| J48 | La respuesta dice que la cita sale del documento aportado: kitlegal no ha consultado el buscador. | SKILL.md, paso 4 (l. 135) | SKILL.md, paso 4 |
| J49 | Ningún ECLI que no venga de estas operaciones o de la persona, ni dentro de una cita ni suelto. | SKILL.md, paso 4 (l. 136–137) | SKILL.md, paso 4 |
| J50 | Las sentencias que menciona el documento no se citan con esa forma ni llevan ECLI: se nombran como las nombra el texto. | SKILL.md, paso 4 (l. 138–139) | SKILL.md, paso 4 |
| J51 | Con la ficha sola: la cita y los datos de la ficha; nada de su contenido. | SKILL.md, paso 5 (l. 143–144) | SKILL.md, paso 5 |
| J52 | Con su texto: leerlo y responder sobre él, con sus palabras. | SKILL.md, paso 5 (l. 145) | SKILL.md, paso 5 |
| J53 | Lo que el texto pegado no trae (tras un `[…]`, los fundamentos que faltan) no se ha leído, y se dice así. | SKILL.md, paso 5 (l. 145–146) | SKILL.md, paso 5 |
| J54 | No resumir ni caracterizar una sentencia sin su texto: ni qué resolvió, ni su doctrina, ni si otra la confirma o la corrige. | SKILL.md, paso 5 (l. 147–148) | SKILL.md, paso 5 |
| J55 | Ante «qué dice la jurisprudencia sobre…», `cita_preparar` con los términos de la pregunta en `texto`, o su orden. | SKILL.md, paso 6 (l. 152–156) | SKILL.md, paso 6; la orden de ejemplo, en references/consulta_preparada.md, «La búsqueda por materia» |
| J56 | Dar la dirección que devuelve, que abre el buscador con esa búsqueda ya hecha. | SKILL.md, paso 6 (l. 158) | SKILL.md, paso 6 (dar la dirección); «abre la búsqueda ya hecha», en references/consulta_preparada.md, «La búsqueda por materia» |
| J57 | Pedir a la persona que descargue y adjunte las sentencias que le interesen, y seguir con lo que traiga (pasos 3 a 5). | SKILL.md, paso 6 (l. 158–159) | references/consulta_preparada.md, «La búsqueda por materia» |
| J58 | Mientras tanto, ninguna sentencia de memoria, ni como ejemplo; si la respuesta necesita una, va por el paso 2. | SKILL.md, paso 6 (l. 159–160) | SKILL.md, paso 6 |
| J59 | El Tribunal Constitucional no está cubierto: decirlo y dar https://hj.tribunalconstitucional.es/ para consultarla a mano. | SKILL.md, paso 7 (l. 164–165) | SKILL.md, paso 7 |
| J60 | No citarla, no ponerle la línea del paso 2 ni pedir `cita_preparar` para ella, aunque traiga su ECLI. | SKILL.md, paso 7 (l. 165–166) | SKILL.md, paso 7 |
| J61 | `kitlegal` se invoca desde el `PATH`, y cada herramienta, por el nombre de su fila. | SKILL.md, «Comandos» (l. 170) | SKILL.md, «Comandos» |
| J62 | Código 0: correcto, también si el documento no es el pedido, que es un hallazgo y no un fallo. | SKILL.md, «Comandos» (l. 170–172) | SKILL.md, «Comandos» |
| J63 | Código 2 (`argumentos`): referencia mal formada, más de una forma, número sin fecha, texto vacío o documento sin ficha reconocible. | SKILL.md, «Comandos» (l. 172–173) | SKILL.md, «Comandos» |
| J64 | Ante un error de argumentos, corregir la llamada o pedir a la persona lo que falta, y no citar. | SKILL.md, «Comandos» (l. 173–174) | SKILL.md, «Comandos» |
| J65 | Tabla de comandos generada desde `--describe`: orden, herramienta, qué hace y qué devuelve; orden y herramienta dan el mismo sobre. | SKILL.md, «Comandos», región generada (l. 176–189) | SKILL.md, «Comandos», región generada, igual byte a byte |
| J66 | Regla 1: nunca inventar sentencia, ECLI, ROJ, número, fecha ni contenido; lo que no está en un documento u operación no se afirma. | SKILL.md, «Reglas», 1 (l. 193–195) | SKILL.md, «Reglas», 1 |
| J67 | Regla 2: ni «existe» ni «no existe»; de la que no está delante se dice que no está comprobada y cómo buscarla. | SKILL.md, «Reglas», 2 (l. 196–197) | SKILL.md, «Reglas», 2 |
| J68 | Regla 3: la respuesta habla a la persona: dirección, casillas y ficha con sus palabras, sin nombrar los campos de la salida. | SKILL.md, «Reglas», 3 (l. 198–199) | SKILL.md, «Reglas», 3 |
| J69 | Regla 4: con normas, distinguir ley de reglamento, señalar la variación autonómica y no aplicar el procedimiento común a lo regulado aparte. | SKILL.md, «Reglas», 4 (l. 200–201) | SKILL.md, «Reglas», 4 |
| J70 | Regla 4: el texto de un artículo no se cita de memoria: se consulta con `boe-legislacion`. | SKILL.md, «Reglas», 4 (l. 201–202) | SKILL.md, «Reglas», 4 |
| J71 | Regla 5: ninguna acción con identidad: no presentar, firmar ni tramitar nada en nombre de nadie, ni simularlo. | SKILL.md, «Reglas», 5 (l. 203) | SKILL.md, «Reglas», 5 |
| J72 | Regla 6: sin herramientas y sin binario, no buscar `kitlegal` en otras carpetas ni en un plugin, ni citar ninguna sentencia. | SKILL.md, «Reglas», 6 (l. 204–205) | SKILL.md, «Reglas», 6 |
| J73 | Regla 6: la respuesta lleva la línea del paso 2 por cada sentencia que necesita y dice que no pudo preparar la consulta. | SKILL.md, «Reglas», 6 (l. 205–206) | SKILL.md, «Reglas», 6 |
| J74 | Regla 6: si ya tiene kitlegal, cerrar Claude del todo, abrirlo y preguntar en un chat nuevo; si no, https://kitlegal.es/instalar/. | SKILL.md, «Reglas», 6 (l. 206–207) | SKILL.md, «Reglas», 6 |
