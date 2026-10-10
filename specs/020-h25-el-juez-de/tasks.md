# Tasks: H25 · El juez de `jurisprudencia`: resumir o caracterizar una sentencia que no se ha leído decide

**Input**: `specs/020-h25-el-juez-de/` (spec.md, plan.md, research.md, data-model.md, contracts/, quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` 2.13.0 aplicada (umbral de materialidad,
criterio de uso, proporcionalidad y convergencia en «Gates»; controles de umbral del ADR 0029; ADR 0030, 0032, 0036 y
0037); H24 y H23 en `main` (`internal/evals` con el juez con modelo, sus umbrales, la medida versionada y su
ejecución, el job en dos modos con su definición comprobada en `make ci`; el applet `cita`; `jurisprudencia` v0 con sus
seis evals y sin juez), y la evidencia de la validación del juez de `jurisprudencia` versionada (la rúbrica, el
esquema de la respuesta, los 249 casos, la medida, las preguntas y los dos informes de sondeo). **Ninguna tarea crea
el esqueleto ni los gates: `make ci` existe y pasa**, y cada tarea lo deja en verde al terminar.

**Aceptación**: el plan dice **«Aceptación e2e: no aplica»** (plan.md, «Aceptación e2e»; research D18): el hito no
cambia el binario ni el applet `cita` —cambiarlos está fuera de alcance (FR-096) y `cmd/kitlegal` no enlaza el paquete
de evals—, así que no hay comportamiento de `kitlegal` que un guion `testscript` describa, y **no hay tarea de
aceptación**. El hito cambia una skill y le añade cuatro evals: van en T003, antes que `SKILL.md` v0.1, que llega en
T006 (Definition of Done §1.10). No son la primera tarea porque con las reglas del conjunto de hoy, que exigen
exactamente seis, y con la skill de una en una, cuatro evals más dejan `make ci` en rojo (research M5; plan.md,
«Orden de implementación», paso 3): entran con las reglas y con la concurrencia que las admiten. La aceptación del
hito es la del job de evals que el workflow lanza en el cierre, tras la revisión final (SC-001, FR-112). En `make ci`
la fijan, sin modelo, los tests de la tabla de plan.md, «Aceptación e2e», que traen las tareas. SC-014 lo mide una
persona después del run.

**Tests**: obligatorios (constitución §III; spec FR-100 a FR-112; plan.md, «Controles mecánicos»). Cada tarea de código
trae su test y la implementación mínima que lo hace pasar, en el mismo diff; los nombres de test y sus ficheros son los
de plan.md, porque el informe final busca cada control `ci:<ruta>:<Test>` como `^func <Test>\(` en esa ruta.

**Rebanadas verticales y excepciones declaradas.** Son rebanadas completas —su test y su implementación en el mismo
diff, `make ci` en verde por sí solas— T001 a T004 y T006. Las demás, cada una por su razón y con su verificación:

- **T005** son tests sobre código que ya existe (el de H24 y la reconstrucción de T002), con los casos del repositorio,
  que no están en el árbol hasta T004. Si uno descubre un defecto de la reconstrucción, su corrección entra en la
  misma tarea.
- **T007** es documentación: su verificación es que cada fichero, test, clave y umbral que nombra existe en el árbol, y
  `make ci`.
- **T008** es el cierre de la Definition of Done, sin código de producto: escribe la lista de lo tocado, mide la
  cobertura y ejecuta quickstart.md §0 a §9; solo añade tests si una cifra de cobertura queda bajo su umbral.

**Por qué T003 y T004 son una tarea cada una** (plan.md, «Obligaciones para `tasks.md`»). Con diez evals y la skill de
una en una, el tope del trabajo no cubre el peor caso (33 397 s, con la fórmula de `definicion.go`): las evals, las
reglas del conjunto y la concurrencia van juntas en T003. Y la carpeta del juez sola deja cuatro tests en rojo
(research M4): va en T004 con su fila en la tabla de las copias, los doce umbrales y la matriz de `medida`.

**Lo que este fichero añade al plan, comprobado en el árbol en esta sesión.** plan.md, «Tests existentes que cambian»,
no nombra `TestJuzgarSentencias`. En `internal/evals/sentencias_test.go`, `juiciosDeLasSeisEvals` exige que la carpeta
de evals de la skill tenga seis (`require.Len(t, conjunto.Evals, 6, …)`, línea 822), y `sesionesModeloDeLasSeisEvals`
da la sesión que pasa cada una, que es además la que `armarSesionesDeJurisprudencia` escribe para
`TestUmbralesDeJurisprudencia` (`internal/evals/informe_test.go`, líneas 4528 y 4529). Con diez evals, los dos fallan.
T003 lleva ese fichero entre sus rutas y le da las sesiones de las cuatro evals nuevas, que son además la prueba, sin
modelo, de los escenarios 3 a 5 de US5.

**Lo que no se ha medido en esta sesión.** No se ha ejecutado ningún test ni `make ci`. Las cifras de las tareas son
las de research.md (tablas V y M) y las de los contratos; la única calculada aquí, a mano y con la fórmula de
contracts/job-de-evals.md §3, es el peor caso intermedio de T003 (8 917 s), que el test toma de `peorCaso` y no de
este fichero.

**Modo**: desatendido (ADR 0018). El workflow `hito` ejecuta y verifica las tareas **una a una**, con guardián de diff:
cada tarea declara en su propia línea, tras «Rutas:», **todas** las rutas que crea, modifica o retira, y solo toca esas
(más `go.mod`, `go.sum`, `CHANGELOG.md`, el directorio del feature y el `x_test.go` de cada `x.go` declarado). Los
demás nombres que aparecen en una línea son contexto: referencias a los documentos del feature (contracts/, plan.md…)
y nombres de ficheros que la tarea solo lee, que se nombran sin su directorio. El cierre en la plataforma —publicar la
rama, abrir la propuesta de cambio y medir CI y el job de evals— lo hace el workflow tras la revisión final;
quickstart.md §10 queda para el workflow y §11 para una persona.

## Formato: `- [ ] Tnnn [etiqueta?] [Story?] Descripción — FR/SC. Rutas: …`

- **[P]**: ninguna tarea lo lleva; en el bucle del workflow todo va en secuencia.
- **Aceptación**: ninguna tarea lleva esa etiqueta (plan.md, «Aceptación e2e: no aplica»).
- **Datos**: ninguna tarea lleva esa etiqueta (plan.md, «Fixtures»; research D17, D19). Nada cambia en el directorio de
  esquemas ni en ningún directorio de fixtures: el esquema de la declaración de clases y el formato de eval ya admiten
  lo que el hito escribe. H25 no graba nada de ninguna fuente: no hay manifiesto `grabaciones.json` nuevo, ni test de
  grabación, ni nada que grabar para el paso `grabar_datos`. Las cuatro copias de T004 son ficheros ya versionados; el
  fragmento, los informes y las preguntas se leen de donde están; y ningún fichero de `data/` cambia. Los transcripts,
  las sesiones, los informes sintéticos y las salidas del juez de los tests son constantes del test o ficheros que el
  test escribe en `t.TempDir()`.
- **[Story]**: US1 quien pregunta por una sentencia no lee nada de una que nadie tiene delante · US2 el job juzga las
  respuestas por lo que dicen · US3 el juez decide porque está medido · US4 una persona repite la medida, y el job
  reconstruye los 249 casos · US5 cuatro evals ponen a la skill donde puede fallar · US6 el trabajo del job cabe en su
  tope · US7 la skill no anuncia un CAPTCHA y dice que el equivalente es deducido. La documentación y el cierre no
  llevan historia.

## Batería de verificación por tarea

- **`make ci` en primer plano** y la tarea marcada `[X]` en el mismo turno en que termina en verde: formato · lint
  (`depguard`, `forbidigo`, `gosec`, `misspell`, `revive`, `dupl`, `gocyclo`, `paralleltest`…, también sobre los
  ficheros con la etiqueta `evals`) · tests con `-race` · `test-integration` · `test-tiempos` · `vuln` · `schema-check`
  · `skills-check` · `goreleaser-check` · secretos · módulos. Ningún registro de la verificación en la raíz del
  repositorio. Las salidas de `go test` y de `make` se leen con `rtk proxy` o con sondas positivas, porque el proxy de
  la sesión las resume.
- **Sin red ni modelo** (FR-095; ADR 0032): ninguna tarea usa la red salvo la de las herramientas de Go (`make vuln`);
  ninguna ejecuta `claude`, `make evals`, `make evals-sondeo`, `make evals-medir-juez`, los guiones `evals*.sh`,
  `TestEjecucionDelJob`, `TestMedidaDelJuez`, `TestSondeo` ni los guiones de Python de la evidencia. Los tests de
  `make ci` usan votantes de pega y no abren ninguna sesión con modelo.
- **Sin personas a mitad**: ninguna tarea la hace, la revisa ni la desbloquea una persona. La medida del juez y el
  sondeo no los lanza el run (FR-050, FR-095).
- **El instrumento no se cambia** (FR-003, FR-026, FR-030, FR-095): la rúbrica, el esquema de la respuesta, los casos y
  la medida llegan validados. Sus copias se escriben una vez, en T004, y ninguna tarea posterior las toca; ninguna
  tarea escribe en la carpeta de la evidencia. Ninguna tarea ni corrección cumple un umbral de plan.md, «Controles de
  umbral», rebajándolo, dejándolo en `decide: false`, sacando respuestas o casos de su total, retirando su test o
  quitando una fila de una tabla de casos. Cada control nuevo se ve en rojo antes de darlo por bueno: con sus casos
  negativos o con un mutante temporal de lo que fija, que no queda en el diff.
- **La orden y el mensaje del voto no cambian** (FR-011): ni `mensajeDelVoto` ni el guion que abre el voto; los tests
  de H24 de los dos siguen pasando sin cambiar lo que esperan.
- **`boe-legislacion` y `legal-core` no cambian** (FR-005): sus evals, sus `SKILL.md`, la carpeta del juez de la
  primera, su concurrencia y sus umbrales. Lo que los tests de H24 esperan de ellas no cambia en ninguna tarea.
- **Ningún test escribe fuera de `t.TempDir()`**; lo temporal de una tarea va bajo `$TMPDIR`, nunca bajo `/tmp`. Los
  fixtures nuevos de los tests son constantes del `_test.go` o ficheros escritos en `t.TempDir()`.
- **Sin atajos** (plan.md, «Comprobación contra la rúbrica», g): ningún `//nolint`, ningún `t.Skip`, TODO ni error
  silenciado; ningún test se desactiva. Si `dupl`, `gocyclo`, `paralleltest`, `gosec` o `misspell` marcan algo, se
  reestructura o se renombra (una función por subprueba desde una tabla; `t.Context()` solo en la llamada).
- **Antes de tocar un tipo, una constante o una lista que construyen los tests**, `grep` de sus usos en todos los
  `_test.go` del paquete: lo que haya que alinear entra en la misma tarea, y sus ficheros están en sus rutas.
- **Proporcionalidad** (constitución, «Gates»; plan.md, «Trazabilidad», último párrafo): ninguna tarea añade un caso,
  un mensaje ni un test para un estado sin vía real —un informe, unas preguntas o unos casos versionados que no se
  pueden leer, y un `quitado` que no dice qué quitar, siguen la regla que ya hay: el caso no se resuelve, con un error
  que lo nombra—.
- **Lo que no cambia** (spec, «Fuera de alcance»; FR-096): el binario, el applet `cita` y `--describe`; las seis evals
  de H23, el formato de eval y lo que `sentencias` comprueba sin modelo; el modelo que decide, el del juez y sus
  versiones de Claude Code; los topes de los dos trabajos; el sondeo; el directorio de esquemas; los ADR; los specs,
  los planes y los informes de otros hitos; la constitución; y el workflow `hito` con sus guiones.

---

## Phase 1: Setup

No hay tareas: el repositorio, los objetivos de `make` y los gates existen desde H0, y el hito no añade ninguna
dependencia (plan.md, «Primary Dependencies»; research V2).

---

## Phase 2: Fundación

No hay tareas sin historia: lo que el hito añade al código compartido —el campo del voto y la reconstrucción— es ya de
US2 y de US4, y va en las dos fases siguientes.

---

## Phase 3: User Story 2 (primera parte) — El voto con `sentencia` (P1)

**Objetivo**: el voto de una clase puede llevar `sentencia`, y el informe lo publica donde el de `boe-legislacion`
lleva `precepto` (plan.md, paso 1).

**Prueba independiente**: `go test -count=1 -run '^(TestVotoDelJuez|TestInformeConElJuez|TestInformeMarkdownDeLosUmbrales)$' ./internal/evals/`.

- [X] T001 [US2] El campo `sentencia` del voto, junto a `precepto`, y la columna de la tabla «Votos» que toma su nombre de los votos que lleva (contracts/juez-de-jurisprudencia.md §3 y §4; data-model §2; research D2, V7; plan.md, paso 1). **`juez.go`**: `VotoDeClase` y lo que se lee de cada clase de un voto ganan `Sentencia *string`, con la clave `sentencia`, al lado de `Precepto`: es `nil`, y va sin clave en `informe.json`, donde el esquema de la clase no la tiene; en un voto que dice no va vacía, y su clave va igual; ningún voto lleva los dos. Las claves de un voto publicado van en el orden de contracts §4: `voto`, `nulo`, `motivo`, `respuesta`, `frase`, el campo propio de la clase y `frase_en_la_respuesta`. **`informe.go`**: en `informe.md`, la tabla «Votos» conserva sus diez columnas; la octava se llama «Sentencia» si algún voto de la tabla lleva `sentencia` y «Precepto» en otro caso, y la celda de la clase que no lo tiene es «—». Lo que se publica de un juez cuyo esquema lleva `precepto` no cambia en un byte (FR-005), y ni `mensajeDelVoto` ni la orden del voto cambian (FR-011). **Tests**, con un juez sintético cuyo esquema es una constante del test, con las dos clases de esta skill y `sentencia` obligatoria en `afirma_lo_no_leido` —la carpeta del juez de la skill llega en T004—: `TestVotoDelJuez` gana el voto del ejemplo de contracts §3, que vale y deja `Sentencia` con su valor y `Precepto` en `nil`; el mismo voto sin `sentencia`, que no tiene la forma del esquema y no llega a darse; y un no, con `sentencia` vacía. `TestInformeConElJuez` gana el caso de ese juez con una respuesta marcada: cada voto de `afirma_lo_no_leido` se publica con `sentencia` y sin `precepto`, comparado con `JSONEq` con la forma de contracts §4, y los de `afirma_que_existe`, sin ninguno de los dos. `TestInformeMarkdownDeLosUmbrales` gana la tabla con la columna «Sentencia» y conserva, con el juez de siempre, la de «Precepto» byte a byte. Los casos de H24 de los tres tests no cambian lo que esperan. Antes de tocar el tipo, `grep` de `Precepto` en los `_test.go` del paquete. — FR-005, FR-011, FR-013, FR-108; SC-008. Rutas: internal/evals/juez.go, internal/evals/informe.go

---

## Phase 4: User Story 4 (primera parte) — La reconstrucción de un caso de esta skill (P2)

**Objetivo**: el reconstructor sabe hacer lo que los casos de `jurisprudencia` piden y el de hoy no: el informe de un
sondeo, el derivado que quita texto de la pregunta y las órdenes del applet `cita` (plan.md, paso 2). No necesita la
carpeta del juez: sus tests usan informes sintéticos.

**Prueba independiente**: `go test -count=1 -run '^(TestLeerCasosEtiquetados|TestArgumentosDeLaInvocacion|TestTextoQuitado|TestResolverCasos)$' ./internal/evals/`.

- [X] T002 [US4] La reconstrucción de los casos de esta skill: `quitado` con `texto`, el informe de un sondeo con sus preguntas, el texto pegado con sus tres recortes y las órdenes `cita` repetidas en proceso con su entrada estándar y su código (contracts/medida-y-casos.md §1 a §7; data-model §3 a §6; research D3 a D10, V3, V8, V9, V11, V12, M1 a M3; plan.md, paso 2). **`medida.go`**: (1) el tipo de `quitado` pasa de `BloqueQuitado` a `Quitado{Norma, Bloque, Texto}`, con la clave `texto` leída —hoy se pierde sin decir nada (research M1)—; `regla` no se lee; y el nombre de un derivado en un error es `<informe> <sesión> sin <norma> <bloque>` o `<informe> <sesión> sin <texto>`. (2) Un informe es de un sondeo si lleva la clave `sondeo` (research D8): sus sesiones están en `sesiones[]`, y de cada una se leen `sesion`, `pregunta` —un id de `preguntas.json`, que está junto al informe—, `respuesta` e `invocaciones[]`. Su pregunta es la de la eval de hoy que nombra esa entrada, o su plantilla con `{fragmento}` sustituido por el fichero que nombra `fragmento`, byte a byte, y `{ficha}`, por ese contenido hasta su primera línea en blanco, con un salto de línea al final (contracts §2; research D9). Sus textos no repiten ninguna orden: de cada invocación de `Bash` cuya orden empieza por `kitlegal `, la orden tal cual y su `salida` (contracts §5). (3) El texto pegado de una pregunta, su entrada y lo que queda con `documento`, `fallo` y `apartado-2`, con las reglas de contracts §3; la pregunta de un derivado es la entrada y, si queda algo, una línea en blanco y lo que queda. (4) En un informe del job, una orden que empieza por `cita ` o por `cita_` se repite en proceso con `AppletCita` en el registro de la sesión: sus argumentos, con la regla de contracts §4 —se parte en cada blanco seguido de `--` y una letra minúscula; el valor de una bandera puede llevar espacios y, en `--documento`, saltos de línea; `--json` una sola vez, al final—; la entrada estándar, por `Registro.LeerDe`, solo para `cotejar` sin `--documento`, con lo que queda del texto pegado; su texto es su orden tal cual y su salida estándar (research D6); y si termina con un código distinto del que el informe da a esa invocación, el caso no se resuelve, con el error de contracts §7. Sin texto pegado, ninguna orden `cotejar` da texto, en los dos tipos de informe. (5) Lo reconstruido de una sesión se recuerda por informe, sesión y texto quitado (research D10). Las órdenes de `boe` y de `graph`, la invocación del servidor, que no da texto, y los derivados por bloque siguen como en H24 (FR-005; research D5, D7). Nada pide nada a la red ni deja nada en la caché, en el grafo o en el directorio temporal de quien lo ejecuta (FR-044). **Tests**, con informes, preguntas y evals sintéticos escritos en `t.TempDir()`: `TestLeerCasosEtiquetados` gana un derivado con `texto` y con `regla`, que se lee con su `Texto`. `TestArgumentosDeLaInvocacion` gana las dos órdenes de la tabla de contracts §4 —la referencia sin bandera, valores con espacios, `--documento=` con saltos de línea, `--json` una vez—. `TestTextoQuitado`, nuevo, sobre el fragmento del repositorio (`fragmentoDelRepositorio`, que el test solo lee) y sobre su ficha: el texto pegado de una pregunta y lo que queda con cada uno de los tres recortes; una pregunta sin texto pegado, y `fallo` o `apartado-2` sobre un texto sin la línea `F A L L O`, dan error. `TestResolverCasos` gana: un informe de un sondeo con sus preguntas —una con `eval`, una con `{fragmento}` y una con `{ficha}`—, con las órdenes y las salidas de su informe; una sesión sin invocaciones, sin textos; una orden `cotejar` del modo `orden`, que recibe el texto pegado por la entrada estándar, y una del modo `herramienta`, con su `--documento`; un código distinto del del informe, que no se resuelve y nombra el caso; y un derivado por cada recorte: con `documento`, su pregunta sin el texto pegado y ninguna orden `cotejar` entre sus textos; con `fallo` y con `apartado-2`, su pregunta sin esa parte y `cotejar` repetida con lo que queda. Los casos de H24 de los cuatro tests, y `TestSinElBloque`, no cambian lo que esperan. — FR-005, FR-041, FR-042, FR-043, FR-044, FR-105. Rutas: internal/evals/medida.go

---

## Phase 5: User Story 5 y User Story 6 (primera parte) — Las diez evals (P2)

**Objetivo**: cuatro evals con preguntas del sondeo, las reglas del conjunto con las diez y el trabajo `evals` de la
skill a cuatro sesiones a la vez (plan.md, paso 3). La skill sigue sin juez hasta T004.

**Prueba independiente**: `go test -count=1 -run '^(TestPreguntasDelSondeo|TestPreguntasConElFragmento|TestConjuntoDeEvals|TestEvalsDelRepositorio|TestJuzgarSentencias|TestUmbralesDeJurisprudencia|TestDefinicionDelJob)$' ./internal/evals/`.

- [X] T003 [US5] Las cuatro evals nuevas, las reglas del conjunto con las diez y `jurisprudencia` a cuatro sesiones a la vez en el trabajo `evals`, en una sola tarea porque con diez evals las reglas de hoy dan cuatro defectos y, de una en una, el tope no cubre el peor caso (contracts/evals-jurisprudencia.md §1 a §4; contracts/job-de-evals.md §1, §3 y §4; data-model §8 y §9; research D13 a D15, M5 a M8, M12, S9; plan.md, paso 3). **Las evals**: los cuatro ficheros de contracts/evals-jurisprudencia.md §1, con su contenido carácter a carácter y, delante, un comentario que dice qué pone a prueba cada una, como los seis de H23. El fragmento y su ficha llegan a las evals (h) e (i) con las órdenes de contracts §2 y de research D15, nunca tecleados ni abiertos con un editor; el fragmento solo se lee. Las seis evals de H23 no cambian en un byte. **`conjunto.go`**: `ReglasDeJurisprudencia` con la tabla de contracts §3 —tamaño, exactamente 10; «número y fecha», 3, y exige además al menos una dirección; «materia», 2; «documento», 3; «no cubierta» y «documento distinto», 1—, sin fijar los valores de ninguna eval. **`evals.yml`**: en el `include` del trabajo `evals`, la skill pasa a `concurrencia: 4`, con `objetivo_de_duracion: 0`, y nada más: la matriz de `medida` llega en T004 y los comentarios, en T007. **Tests**: `TestPreguntasDelSondeo`, nuevo, en `conjunto_test.go`: la pregunta de cada una de las cuatro evals, tal como la lee `LeerEval`, es igual, byte a byte, a la de su entrada de `preguntas.json` compuesta con el fragmento y su ficha, con la tabla que une cada eval con su pregunta del sondeo (research D14); con un byte cambiado en una copia de una eval, deja de serlo. `TestPreguntasConElFragmento` gana la eval (h), que lleva el fragmento byte a byte, y la (i), que lleva su ficha byte a byte y nada más del fragmento. `TestConjuntoDeEvals`, en su caso de esta skill, con evals sintéticas: las diez cumplen; con nueve y con once, falla la de tamaño; con una de «número y fecha» sin dirección, y con cada una de las cuatro nuevas sin la línea, sin la dirección, sin una casilla, sin `ninguna_cita`, sin la cita, sin la dirección de búsqueda o sin su comando, falla la regla de su clase. `TestEvalsDelRepositorio`, en el conjunto de esta skill: las diez cumplen las reglas. `TestJuzgarSentencias`, en `sentencias_test.go`: las sesiones modelo pasan a ser las de las diez evals —la de (g), con `cita preparar`, la línea, la dirección, las dos casillas con `241/2013` y `09/05/2013` y ninguna cita; las de (h) y (i), con `cita cotejar` y la cita con su forma, con `ECLI:ES:TS:2023:3144` y `STS 3144/2023`; la de (j), con `cita preparar` con texto, la dirección de búsqueda que devolvió y ninguna cita—, cada una pasa en los dos modos, y no pasan, con su motivo: la de (g) sin la línea y la de (g) con una cita; la de (h) sin `cita cotejar` y la de (i) sin la cita; y la de (j) sin la dirección de búsqueda y la de (j) con una cita (US5, escenarios 3 a 5). `TestUmbralesDeJurisprudencia` sigue con cuatro umbrales —el juez llega en T004—, ahora sobre las 30 respuestas de cada modo, con sus motivos recalculados sobre 30. `TestDefinicionDelJob`: la skill con `{concurrencia: 4, objetivo_de_duracion: 0}`; `del-repositorio` sin ninguna línea; `sinteticas`, con la definición del contrato del test a cuatro y sus casos recalculados con la misma fórmula; y `peor-caso` y `tope-con-las-evals-del-repositorio` con las diez evals y cuatro a la vez, todavía sin juez —lo que da `peorCaso`: 485 s y 31 tandas de 272 s, 8 917 s—, y la mutación «de una en una en `evals`», que da la línea del `include` y la del tope de contracts/job-de-evals.md §4 con el peor caso que `peorCaso` da sin juez. Los peores casos de `boe-legislacion` y de `legal-core` no cambian. Antes de escribir: `grep` de `tamanioDeJurisprudencia`, `porNumeroYFechaDelConjunto`, `unaDeCadaClase`, `sesionesModeloDeLasSeisEvals` y `respuestasDeJurisprudenciaPorModo` en los `_test.go` del paquete. — FR-005, FR-024, FR-060, FR-061, FR-062, FR-063, FR-064, FR-065, FR-066, FR-070, FR-072, FR-103, FR-110; SC-003, SC-010. Rutas: evals/jurisprudencia/07-resumen-de-una-conocida.yaml, evals/jurisprudencia/08-doctrina-con-el-fallo-delante.yaml, evals/jurisprudencia/09-de-que-trata-con-la-ficha-sola.yaml, evals/jurisprudencia/10-doctrina-dada-por-hecha.yaml, internal/evals/conjunto.go, internal/evals/sentencias_test.go, internal/evals/informe_test.go, internal/evals/definicion_test.go, .github/workflows/evals.yml

---

## Phase 6: User Story 3, User Story 2 (segunda parte), User Story 1 y User Story 6 (segunda parte) — El juez de la skill (P1)

**Objetivo**: la carpeta del juez de `jurisprudencia`, con lo que una persona validó, y con ella el juez sobre sus
sesiones, sus doce umbrales, la comprobación de su medida y la skill en el trabajo `medida` (plan.md, paso 4).

**Prueba independiente**: `go test -count=1 -run '^(TestCopiasDelJuez|TestMedidaVersionada|TestEjecucionSinMedir|TestEvalsDelRepositorio|TestUmbralesDeJurisprudencia|TestInformeConElJuez|TestTextosDeLaSesion|TestMensajeDelVoto|TestDefinicionDelJob)$' ./internal/evals/`.

- [X] T004 [US3] La carpeta del juez de `jurisprudencia`, su fila en la tabla de las copias, sus doce umbrales y la skill en la matriz de `medida`, en una sola tarea porque la carpeta sola deja cuatro tests en rojo (contracts/juez-de-jurisprudencia.md §1, §2, §5 a §7 y §9; contracts/job-de-evals.md §1 a §4; data-model §1, §7 y §8; research D12, D20, V4 a V6, V13, V14, M4, M6; plan.md, paso 4). **La carpeta**, nueva, con las dos órdenes de contracts/juez-de-jurisprudencia.md §1: `rubrica.md`, `esquema.json`, `casos.yaml` y `medida.json`, copiados una sola vez con `cp` desde la carpeta de la evidencia de la validación del juez de esta skill, sin abrirlos con un editor y comparados con `cmp` antes de terminar; y `clases.yaml`, con el contenido de contracts §1 carácter a carácter: `afirma_lo_no_leido`, que decide con 0, y `afirma_que_existe`, que solo se publica, sin `cuenta_su_proceso`. La carpeta de la evidencia no se escribe, y ninguna tarea posterior toca las copias. **`evals.yml`**: la matriz de `medida` pasa a `[boe-legislacion, jurisprudencia]`, con la entrada de `include` de la skill con `concurrencia: 4`, y la descripción de la entrada `medir_al_juez` es la de contracts/job-de-evals.md §1; no cambian los dos topes, la condición de `medida`, el modelo del juez ni su versión de Claude Code (FR-033, FR-072). **Tests**: `TestCopiasDelJuez` gana la fila de la skill: sus cuatro copias son idénticas, y con un byte cambiado en cada una, o sin ella, el defecto nombra el fichero: 4 de 4. `TestMedidaVersionada` y `TestEjecucionSinMedir` hacen sus seis mutaciones por cada fila de la tabla, y no solo por la primera (research V13): para esta skill, la medida del repositorio corresponde y se cumple; con cada una de sus cuatro claves cambiada, o con cada recuento distinto de 0, la comprobación da su línea y el job termina con `fallo` sin abrir ninguna sesión, con el motivo de contracts §7 y sin ningún umbral del juez en `umbrales`: 6 de 6. `TestEvalsDelRepositorio`, en el conjunto de esta skill, deja de exigir que no tenga juez y exige sus dos clases, cuál decide y sus umbrales. `TestUmbralesDeJurisprudencia`, con un votante de pega y las 30 respuestas de cada modo, deja de exigir que el informe no lleve `juez` y fija: los doce elementos en el orden de contracts §5, diez con `decide: true`, y los dos de la medida con 0 de 125 y 0 de 124; con 1 respuesta marcada en `afirma_lo_no_leido` en un modo y 0 en el otro, `cumple: false` en ese modo y `fallo`, con el motivo de contracts §6, que nombra la sesión y sus tres frases; con 0, se cumple y el veredicto no cambia; con 3 respuestas con sí en `afirma_que_existe`, `medida` 3, `decide: false`, el mismo veredicto y sus frases en `juez`; una respuesta marcada con sí en las dos clases cuenta una vez en cada umbral; con 901 s de votos en un modo, `fallo` con el motivo de la ejecución; el voto publicado lleva `sentencia`; se juzgan las respuestas del modelo que decide y ninguna del informativo, con 60 votos si ninguna se marca y dos más por cada una que el primer voto marca; y los casos de `cita_sin_documento` y de `sin_activar` siguen dando su `fallo`, ahora entre los doce. Si este test descubre que el código de H24 no da los doce como dice contracts §5 —research V6 lo leyó y no lo ejecutó con esta skill—, la corrección entra aquí, en `umbrales.go`, sin cambiar ningún resultado de `boe-legislacion` ni de `legal-core` (FR-005). `TestInformeConElJuez`, de H24, no cambia: su caso del voto que no llega sigue dando la respuesta sin juzgar y `fallo`. `TestTextosDeLaSesion` y `TestMensajeDelVoto` ganan una sesión de cada modo, con una orden `kitlegal cita cotejar` y con una llamada `cita_cotejar`, sobre la pregunta con el fragmento pegado: el mensaje lleva la pregunta entera, la respuesta y cada texto, en su orden, también el de una orden que falla, y ningún byte de `SKILL.md`, de lo que la eval espera ni del juicio sin modelo; lo que ya esperan de la orden y del mensaje no cambia (FR-011). `TestDefinicionDelJob`: la matriz de `medida` que se espera se deriva de las skills de la matriz de `evals` con carpeta de juez, en su orden, cada una con la concurrencia de su trabajo `evals` (research D12); `peor-caso`, con el de la skill en `{[61, 60], 4}` y su juez `{[30, 30], 4}`, 12 577 s, y el de su medida, `{249, 4}`, 15 505 s; `tope-con-las-evals-del-repositorio`, con los topes de 352 y 269 minutos cubriéndolos; las tres mutaciones de contracts/job-de-evals.md §4 —de una en una en `evals`, de una en una en `medida` y sin la skill en `medida`—, cada una con sus líneas; y `sinteticas`, con la skill en la matriz de `medida` y con carpeta de juez en las evals sintéticas. Sigue exigiendo que `medida` solo corra con su etiqueta o con su entrada, que no tenga `needs` y que ni `tanda` ni `evals` nombren esa etiqueta (FR-050). — FR-001, FR-002, FR-003, FR-004, FR-005, FR-010, FR-011, FR-012, FR-013, FR-020, FR-021, FR-022, FR-023, FR-024, FR-025, FR-030, FR-031, FR-032, FR-033, FR-040, FR-050, FR-071, FR-072, FR-102, FR-103, FR-104, FR-108, FR-109, FR-111; SC-001, SC-002, SC-003, SC-004, SC-008, SC-009, SC-013. Rutas: evals/jurisprudencia/juez/, internal/evals/medida_test.go, internal/evals/ejecucion_test.go, internal/evals/conjunto_test.go, internal/evals/informe_test.go, internal/evals/umbrales.go, internal/evals/sesion_test.go, internal/evals/juez_test.go, internal/evals/definicion_test.go, .github/workflows/evals.yml

---

## Phase 7: User Story 4 (segunda parte) — Los 249 casos del repositorio (P2)

**Objetivo**: los casos de la skill se resuelven todos, cada derivado solo pierde lo quitado, y la ejecución de la
medida da la medida de lo que hay o dice qué caso juzgó mal (plan.md, paso 5).

**Prueba independiente**: `go test -count=1 -run '^(TestReconstruccionDeJurisprudencia|TestGrabacionesDerivadas|TestEjecucionDeLaMedida)$' ./internal/evals/`.

- [X] T005 [US4] Los tests con los 249 casos del repositorio: la reconstrucción, el control de derivaciones y la ejecución de la medida con un votante que responde según la etiqueta de cada caso (contracts/medida-y-casos.md §7 a §9 y §11; data-model §6; research D11, M2, S1; plan.md, paso 5). Son tests de código que ya existe —la reconstrucción de T002 y la medida de H24—, con los casos que T004 dejó en la carpeta del juez; ninguno lee los votos de la evidencia, y ninguno abre un voto de verdad. `TestReconstruccionDeJurisprudencia`, nuevo, en `medida_test.go`: se resuelven 249, con 138, 39 y 72 por informe; 122 sin quitar nada y 127 derivados, 45, 41 y 41 por lo quitado. Un caso del informe del cierre de H23 lleva el sobre `ok: true` de cada orden `cita`, y el de un `cotejar` del modo `orden`, la huella del texto pegado en su `url`; uno de `sondeo-con-skill.json` lleva las órdenes y las salidas de su informe, y uno de `sondeo-sin-skill.json`, ningún texto; cada derivado lleva su pregunta sin lo quitado, y los 45 sin el documento, ninguna orden `cotejar`. No usa la red, ningún modelo ni Python, y no deja nada en el directorio temporal, en la caché ni en el grafo (FR-044). `TestGrabacionesDerivadas` pasa a recorrer los casos de cada skill con juez; para esta, lo de contracts §8: la respuesta de cada uno de los 249 es, byte a byte, la de su sesión leída aparte de su informe; la pregunta de cada caso sin `quitado` es la de su sesión, y la de cada derivado, esa sin la parte quitada y nada más; las órdenes de los textos de cada derivado son las de su sesión, en su orden, y sin el documento, las mismas menos las de `cotejar`; y en un caso de un sondeo, cada salida es la de su informe. `TestEjecucionDeLaMedida` gana los casos de esta skill, con un votante que cuenta sus llamadas: con los 249 bien, 499 votos y la medida con sus cuatro claves —las huellas de la rúbrica y de los casos, el id del modelo del juez y la versión de Claude Code de sus votos— y sus dos recuentos, 0 de 125 y 0 de 124, con la forma de contracts §9; con un defecto sin marcar, y con un correcto marcado, error con el caso y sus frases, con la forma de la línea de contracts §9; y con un caso que no se resuelve, sobre una copia de la carpeta del juez en `t.TempDir()`, error que lo nombra y ningún voto. Lo que los dos tests esperan de `boe-legislacion` no cambia. Si un test descubre un caso que la reconstrucción de T002 no resuelve como el prototipo de research M2, la corrección entra aquí, en `medida.go`; ningún caso, etiqueta ni total cambia para que un test pase (FR-026). — FR-005, FR-026, FR-040, FR-041, FR-042, FR-043, FR-044, FR-045, FR-051, FR-105, FR-106, FR-107; SC-005, SC-006, SC-007. Rutas: internal/evals/medida.go

---

## Phase 8: User Story 7 — `jurisprudencia` v0.1 (P3)

**Objetivo**: los dos pasajes de `SKILL.md` que llevan a la respuesta a anunciar un CAPTCHA y a dar el equivalente
como un dato de la sentencia (plan.md, paso 6, y «Cambios de `SKILL.md` trazados a la causa»). Va detrás de las evals
(T003) y del instrumento que la mide (T004).

**Prueba independiente**: `make skills-check`.

- [X] T006 [US7] `jurisprudencia` v0.1: los dos pasajes de `SKILL.md`, C1 y C2, aplicados con `git apply specs/020-h25-el-juez-de/contracts/skill-jurisprudencia-v0.1.diff` y nada más (contracts/skill-jurisprudencia.md §1 a §3 y §5; research D16, M11, «Traza de las dos frases»; plan.md, paso 6). **C1**, el párrafo inicial: el obstáculo del buscador es de los programas y no de la persona, a quien con su navegador no le sale, y por eso la respuesta no le anuncia un CAPTCHA ni le pide que resuelva ninguno (FR-081). **C2**, la viñeta del equivalente del paso 2: si la respuesta nombra el ECLI o el ROJ que `cita preparar` deduce, lo da como lo que es, deducido de esa referencia y sin que nadie lo haya comprobado (FR-082). La causa de cada uno, con sus respuestas por sesión y sus frases, ya está en research.md y en plan.md (FR-080): esta tarea no la vuelve a escribir. No cambian la descripción del frontmatter, el protocolo, la forma de la cita ni la de la línea `⚠ SENTENCIA NO COMPROBADA:`, las reglas ni la tabla de comandos (FR-083). **Verificación**: el fichero tiene 197 líneas; `grep -c CAPTCHA` da 2, las del párrafo inicial; `git diff --stat main` del fichero da 9 líneas añadidas y 7 quitadas; y `make skills-check` pasa, con menos de 300 líneas, frontmatter válido y la región generada sin drift, que es el control de SC-011 —sus casos negativos, las 300 líneas y el drift, ya están en `sincronia_test.go` y no cambian—. El efecto en las respuestas no lo mide ninguna tarea ni ningún guion (ADR 0037): lo lee una persona tras el cierre (SC-014; research S5). — FR-080, FR-081, FR-082, FR-083; SC-011. Rutas: skills/jurisprudencia/SKILL.md

---

## Phase 9: Cierre — documentación y Definition of Done

- [X] T007 La documentación que el hito deja falsa o incompleta (contracts/juez-de-jurisprudencia.md; contracts/medida-y-casos.md; contracts/evals-jurisprudencia.md; contracts/job-de-evals.md §3; contracts/skill-jurisprudencia.md §3; plan.md, paso 7). CHANGELOG.md, *Unreleased*: `jurisprudencia` v0.1 con lo que cambia para quien la usa —no anuncia un CAPTCHA a la persona y da el equivalente deducido como deducido y sin comprobar (FR-084)—, y lo del job: juzga `jurisprudencia` con el juez con modelo, sus dos clases y cuál decide, sus umbrales, las cuatro evals nuevas, su trabajo a cuatro sesiones a la vez y la etiqueta `evals-medir-juez`, que mide a las dos skills (FR-090, FR-112). CONTRIBUTING.md, en lo que dice del formato de eval y del job de evals, los seis puntos de FR-093, con cada cifra tomada del árbol —la definición del job, las reglas del conjunto, la medida versionada y los tests— y no de memoria: (1) la skill tiene carpeta de juez, con `afirma_lo_no_leido`, que decide, y `afirma_que_existe`, que solo se publica; (2) diez evals, la regla del conjunto que las exige con lo que espera cada una de las cuatro nuevas, y las sesiones que abre su trabajo; (3) cuatro sesiones a la vez en el trabajo `evals` y en el de la medida, sus peores casos, 12 577 s y 15 505 s, bajo los topes de 352 y 269 minutos, y que sigue sin objetivo de duración de las sesiones; (4) doce elementos en `umbrales`, diez que deciden, en su orden; (5) la medida de las dos skills, que la etiqueta y la entrada del flujo lanzan sin selector por skill, con los recuentos de la de esta (0 de 125 y 0 de 124) y lo que sus casos tienen de distinto, el informe de un sondeo y el derivado que quita texto de la pregunta; y (6) el campo `sentencia` del voto, donde el texto describe `precepto`. Los pasajes están hoy en torno a sus líneas 625, 702, 922, 942, 955, 1042, 1084, 1107 y 1129. No reescribe lo que sigue siendo cierto, no promete ninguna release y no añade ningún test que ate el texto a la definición del job. docs/JURISPRUDENCIA.md (FR-091): lo que H25 deja hecho —resumir o caracterizar una sentencia que no se tenía delante lo decide el juez con `afirma_lo_no_leido`; `afirma_que_existe` solo se publica; y sigue sin medir nadie lo que la respuesta dice de una norma y su fidelidad al texto que sí leyó—, y deja de decir que no lo decide ningún control. docs/WORKFLOW.md (FR-092): la fila «Evidencia de un ADR» nombra también la carpeta de la validación del juez de una skill, con la ruta que da FR-092, y los cuatro ficheros que la carpeta del juez de `jurisprudencia` copia de ella. internal/evals/doc.go: lo que el comentario del paquete dice de `jurisprudencia` —que no tiene juez y que sus umbrales son cuatro— pasa a decir lo de contracts/juez-de-jurisprudencia.md §5, y la reconstrucción nombra el informe de un sondeo, el derivado por texto y las órdenes `cita` (FR-024). .github/workflows/evals.yml: solo comentarios —la cabecera y los de los trabajos que dicen que la skill no tiene juez, que va de una en una y que su peor caso es de 20 341 s pasan a decir lo de contracts/job-de-evals.md §3 (FR-070)—. **Verificación**: cada fichero, test, clave y umbral que se nombra existe en el árbol (`grep` uno a uno); ninguna de las seis afirmaciones falsas de SC-015 queda en CONTRIBUTING.md; y `make ci`, que con `TestDefinicionDelJob` ve que los comentarios no han cambiado ninguna clave del flujo. No se editan los ADR, el roadmap, la bitácora de uso, la constitución ni los specs de otros hitos (FR-096). — FR-024, FR-070, FR-084, FR-090, FR-091, FR-092, FR-093, FR-096, FR-112; SC-012, SC-015. Rutas: CHANGELOG.md, CONTRIBUTING.md, docs/JURISPRUDENCIA.md, docs/WORKFLOW.md, internal/evals/doc.go, .github/workflows/evals.yml
- [ ] T008 Cierre de la Definition of Done, sin código de producto (`ROADMAP.md` §1; plan.md, «Obligaciones para `tasks.md`»): specs/020-h25-el-juez-de/cierre.md con (1) la lista, sacada de `git diff --name-status main`, de la carpeta del juez, las evals, la skill, el flujo y los ficheros de Go y de documentación creados o modificados, y la comprobación de que no cambia nada de lo que el hito deja como está: la carpeta de la evidencia, el directorio de esquemas, ningún fixture ni grabación, las seis evals de H23, las evals y las skills de `boe-legislacion` y de `legal-core` con la carpeta del juez de la primera, el applet `cita` y su dominio, el guion del voto y `mensajeDelVoto`, los dos topes y el modelo y la versión del juez en el flujo, los ADR, la constitución, los guiones del workflow y los specs de otros hitos (FR-005, FR-011, FR-033, FR-073, FR-095, FR-096); (2) las cuatro copias de la carpeta del juez comparadas con `cmp` con sus originales, sin salida (FR-001, FR-030), y la constancia de que las diferencias de `SKILL.md` con el de `main` son solo los dos pasajes de T006 (FR-083); (3) la cobertura sobre `coverage.out` y `coverage-integration.out` de `make ci` con `go tool cover -func` —global ≥ 70 % y el dominio ≥ 85 % (Definition of Done §1.9)— y la del paquete de evals; (4) el resultado de quickstart.md §0 a §9, orden a orden, con lo esperado de cada una y con la caché en un directorio bajo `$TMPDIR`, y la constancia de que §10 lo hace el workflow y §11 queda para después del run; (5) la tabla de plan.md, «Controles de umbral», con el test de cada fila `ci:` encontrado como `^func <Test>\(` en su ruta, que es como lo busca el informe final, y el objetivo `skills-check` en los de `make`; (6) la constancia de que ningún umbral de FR-020, FR-022 o FR-023 se rebajó, pasó a `decide: false` ni perdió respuestas o casos de su total, de que la rúbrica, los casos y la medida son los de sus originales (FR-026), y de que el diff del hito no añade ningún `//nolint`, `t.Skip` ni TODO, no abrió ninguna sesión con modelo y no lanzó la medida ni un sondeo (FR-095); y (7) la constancia de que plan.md, research.md y contracts/skill-jurisprudencia.md §6 conservan su apartado de las reparaciones del cierre, con «ninguna todavía», que es donde una reparación del cierre deja su traza (FR-027). Solo si una cifra de cobertura queda bajo su umbral se añaden los tests que faltan, en los `_test.go` de los ficheros declarados. — FR-001, FR-005, FR-011, FR-026, FR-027, FR-030, FR-033, FR-073, FR-083, FR-095, FR-096, FR-100, FR-101; SC-012. Rutas: specs/020-h25-el-juez-de/cierre.md, internal/evals/juez.go, internal/evals/informe.go, internal/evals/medida.go, internal/evals/conjunto.go

---

## Dependencias y orden de ejecución

El orden es el de plan.md, «Orden de implementación (de dentro afuera)», y es estrictamente secuencial: T001 → T008.
Ninguna tarea depende de una posterior.

| Tarea | Paso del plan | Necesita |
|---|---|---|
| T001 | 1 | — |
| T002 | 2 | — (sus tests usan informes sintéticos; no necesita la carpeta del juez) |
| T003 | 3 | — (va detrás de T001 y T002 por el orden del plan; la skill sigue sin juez) |
| T004 | 4 | T001 (el voto con `sentencia`, que el informe de la skill publica), T003 (las diez evals: 30 respuestas por modo y el peor caso con el juez) |
| T005 | 5 | T002 (la reconstrucción), T004 (los casos en la carpeta del juez) |
| T006 | 6 | T003 (las evals van antes que la skill); va detrás de T004 porque el instrumento precede a la skill |
| T007 | 7 | T001 a T006 |
| T008 | — | T001 a T007 |

### Historias

- **US2** (T001, T004) y **US3** (T004) son el instrumento que ve el defecto y que solo decide porque está medido;
  **US1** se sostiene en él —el umbral que decide, en T004— y en las evals que provocan el defecto (T003); **US5**
  (T003) pone a la skill donde puede fallar; **US6** (T003, T004) hace que el trabajo quepa en su tope; **US4** (T002,
  T005) deja que una persona repita la medida; y **US7** (T006) corrige las dos frases.
- Las historias no son independientes en su construcción —comparten `internal/evals`— y sí en su prueba: cada fase
  nombra la orden que la comprueba sola.

### Oportunidades de paralelismo

Ninguna dentro del run: el workflow ejecuta las tareas una a una, y casi todas tocan los mismos ficheros de
`internal/evals`. Ninguna tarea lleva `[P]`. Fuera de un run, T001, T002 y T003 no dependen entre sí, y T006 solo de
T003.

---

## Trazabilidad

### Requisitos → tareas

| Requisito | Tareas |
|---|---|
| FR-001, FR-002 | T004; T008 (las copias, constatadas) |
| FR-003 | T004 (la copia de la rúbrica); restricción de todas las tareas (batería) |
| FR-004 | T004 (`decide: false` en `clases.yaml`; cuenta con el primer voto y no cambia el veredicto) |
| FR-005 | T001, T002, T003, T004, T005 (lo que los tests de H24 esperan no cambia); T008 |
| FR-010, FR-012 | T004 |
| FR-011 | T001, T004 (los tests de H24 de la orden y del mensaje, sin cambiar); T008 |
| FR-013 | T001 (el voto y la columna), T004 (el informe de la skill) |
| FR-020, FR-021, FR-022, FR-023, FR-025 | T004 |
| FR-024 | T003 (`cita_sin_documento` y `sin_activar` sobre las diez evals), T004 (los doce), T007 (`doc.go`) |
| FR-026 | Restricción de todas las tareas (batería); T005; T008 la deja constatada |
| FR-027 | Rige las reparaciones del cierre, que hace el workflow tras la revisión final y no una tarea; T008 deja constancia de que los apartados donde dejan su traza siguen en su sitio |
| FR-030 | T004 (la copia; ninguna tarea repite ni escribe la medida); T008 |
| FR-031, FR-032 | T004 |
| FR-033 | T004 (el flujo no los cambia y `TestDefinicionDelJob` los sigue exigiendo); T008 |
| FR-040 | T004 (la copia de los casos), T005 |
| FR-041, FR-042, FR-043, FR-044 | T002 (con informes sintéticos), T005 (con los casos del repositorio) |
| FR-045 | T005 |
| FR-050 | T004 |
| FR-051 | T005 |
| FR-060, FR-061, FR-062, FR-063, FR-064, FR-065, FR-066 | T003 |
| FR-070 | T003 (la concurrencia), T007 (los comentarios del flujo) |
| FR-071 | T004 |
| FR-072 | T003 (el tope de `evals` con las diez), T004 (los dos topes con el juez y con los 249 casos) |
| FR-073 | Nada cambia; T008 lo deja constatado. Las nueve sesiones a la vez las mide el cierre |
| FR-080 | research.md y plan.md (la traza, ya escrita); T006 aplica C1 y C2 |
| FR-081, FR-082, FR-083 | T006; T008 (FR-083, constatado) |
| FR-084 | T007 |
| FR-090, FR-091, FR-092, FR-093 | T007 |
| FR-095, FR-096 | Restricción de todas las tareas (batería); T007; T008 |
| FR-100 | Todas (`make ci` por tarea); T008 |
| FR-101 | plan.md, «Controles de umbral»; tabla siguiente; T008 |
| FR-102 | T004 |
| FR-103 | T003, T004 |
| FR-104 | T004 |
| FR-105 | T002, T005 |
| FR-106, FR-107 | T005 |
| FR-108 | T001, T004 |
| FR-109 | T004 |
| FR-110 | T003 |
| FR-111 | T004 |
| FR-112 | T007; el job de evals lo lanza el workflow en el cierre |
| SC-001 | Lo mide el cierre del workflow, que no es una tarea; T004 construye los controles que lo hacen fallar (con T003, los de las diez evals) y T006 trae la skill |
| SC-002 | T004 |
| SC-003 | T003 (de una en una en `evals`), T004 (las dos mutaciones, con el juez) |
| SC-004 | T004 |
| SC-005, SC-006, SC-007 | T005 |
| SC-008 | T001, T004 |
| SC-009 | T004 |
| SC-010 | T003 |
| SC-011 | T006 |
| SC-012 | T007, T008 |
| SC-013 | T004 |
| SC-014 | Después del run, una persona; no es una tarea |
| SC-015 | T007; lo comprueba la revisión final por lectura |

### Controles de umbral (plan.md) → tarea que construye el control

| Fila de plan.md | Control | Tarea | Test que lo ve fallar |
|---|---|---|---|
| FR-020, SC-001 (los dos modos) | `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5:<modo>` | T004 | `TestUmbralesDeJurisprudencia`: 1 marcada en un modo, `cumple: false` y `fallo` |
| FR-022, SC-001 (defectos y correctos) | `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:…` | T004 | `TestMedidaVersionada` y `TestEjecucionSinMedir`: cada recuento distinto de 0 |
| FR-023, SC-001 (los dos modos) | `evals:jurisprudencia:duracion_del_juez:<modo>` | T004 | `TestUmbralesDeJurisprudencia`: 901 s, `fallo` |
| FR-024, SC-001 (cita sin documento, los dos modos) | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:<modo>` | T003, T004 | `TestUmbralesDeJurisprudencia`: 1 de 30 y 2 de 30, `fallo` |
| FR-024, SC-001 (sin activar, los dos modos) | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:<modo>` | T003, T004 | `TestUmbralesDeJurisprudencia`: 1 de 30, `fallo` |
| SC-001 (sin juzgar) | `ci:internal/evals/informe_test.go:TestInformeConElJuez` | T001 y T004, que tocan el informe y lo dejan en verde; el test es de H24 | Su caso del voto que no llega: `fallo` |
| FR-001, FR-102, SC-002 | `ci:internal/evals/medida_test.go:TestCopiasDelJuez` | T004 | Un byte cambiado en cada copia, o sin ella: 4 de 4 |
| FR-072, FR-103, SC-003 | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` | T003, T004 | De una en una en `evals` y en `medida`: 2 de 2 |
| FR-031, FR-104, SC-004 | `TestMedidaVersionada`, `TestEjecucionSinMedir` | T004 | Cuatro claves y dos recuentos: 6 de 6 |
| FR-044, FR-105, SC-005 | `ci:internal/evals/medida_test.go:TestReconstruccionDeJurisprudencia` | T005 (el mecanismo, T002) | 249 resueltos y 0 `cotejar` en los 45; un código distinto del informe no se resuelve (`TestResolverCasos`, T002) |
| FR-045, FR-106, SC-006 | `ci:internal/evals/medida_test.go:TestGrabacionesDerivadas` | T005 | Una respuesta o un derivado que no coincide |
| FR-051, FR-107, SC-007 | `ci:internal/evals/medida_test.go:TestEjecucionDeLaMedida` | T005 | Un defecto sin marcar y un correcto marcado: error con el caso |
| FR-013, FR-108, SC-008 | `TestVotoDelJuez`, `TestUmbralesDeJurisprudencia` | T001, T004 | El voto sin `sentencia` no tiene la forma; el publicado la lleva |
| FR-010, FR-109, SC-009 | `TestTextosDeLaSesion`, `TestMensajeDelVoto` | T004 | El mensaje, byte a byte, en los dos modos |
| FR-065, FR-066, FR-110, SC-010 | `TestPreguntasDelSondeo`, `TestPreguntasConElFragmento`, `TestConjuntoDeEvals`, `TestEvalsDelRepositorio` | T003 | Un byte cambiado en una eval; nueve y once evals |
| FR-083, SC-011 | `ci:Makefile:skills-check` | T006 | Los casos de las 300 líneas y del drift de `TestSkillsDelRepositorio`, que no cambian |
| FR-111, SC-013 | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia` | T004 | Los cuatro casos de FR-111 y los doce elementos |

### Definition of Done (`ROADMAP.md` §1) → tareas

| Punto | Aplica | Tareas |
|---|---|---|
| 1 · `make ci` en verde | Sí | Todas; T008 |
| 2 · Tests offline; fixtures si toca red | Sí; no toca red | T001 a T005; ningún fixture grabado |
| 3 · Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | Sí | T001, T002 (la salida de cada orden repetida va a un búfer) |
| 4 · Esquemas y `schema-check` | Ningún esquema cambia; `schema-check` sigue en verde | T008 lo constata |
| 5 · Errores tipados y códigos estables | Ningún código de `kitlegal` cambia | — |
| 6 · e2e y `CHANGELOG.md` | «Aceptación e2e: no aplica»; `CHANGELOG.md`, sí | T007 |
| 7 · ADR | No: la decisión es el ADR 0037, ya aceptado (FR-096) | — |
| 8 · `SOURCES.md` | No: ninguna fuente nueva ni cambiada (research V24, D19) | — |
| 9 · Cobertura | Sí | T008 |
| 10 · Skill: evals antes, `SKILL.md` < 300, sin drift | Sí | T003 (las evals), T004 (el instrumento), T006 (la skill) |
| 11 · Dimensión territorial | No | — |
| 12 · Operaciones de grafo | No: ningún applet cambia | — |
| 13 · Recursos, plazos o escritos | No | — |

---

## Estrategia de implementación

- **De dentro afuera** (plan.md): lo que lleva el voto, lo que sabe hacer el reconstructor, las evals con sus reglas y
  su concurrencia, la carpeta del juez con todo lo que su llegada cambia, los tests con los casos del repositorio y,
  al final, la skill y la documentación.
- **Lo mínimo que ya da valor** son US2, US3 y US5 con el job (T001, T003 y T004): un job que marca la respuesta que
  resume una sentencia que no tenía delante, sobre preguntas que provocan el defecto, y que no decide sin estar
  medido. No es una entrega parcial: el run ejecuta las ocho, y SC-001 pide el cierre entero.
- **Lo que el run no puede medir** (research S1 a S9): que el juez vota los casos reconstruidos en Go como votó los de
  la validación; que nueve sesiones a la vez no chocan con el límite de ritmo; cuánto tarda un voto; que las series de
  las cuatro evals nuevas pasan en el modo herramienta; y que v0.1 quita el CAPTCHA de las respuestas. Lo primero lo
  mide una persona con la etiqueta de la medida (SC-014), y lo demás, el job del cierre. Si una respuesta queda
  marcada, la reparación sigue FR-027 y no toca la rúbrica, los casos, la medida ni el umbral (FR-026).
- **Después del run** (SC-014): una persona lanza una vez la medida con el código nuevo y lee las respuestas con algún
  voto afirmativo y las de las cuatro evals nuevas.

---

## Comprobación contra la rúbrica de `juez_tasks` y `precheck.sh tasks`

- `precheck.sh tasks`: ocho líneas con el formato `- [ ] Tnnn`, ids únicos y correlativos; todas con rutas; ninguna
  nombra el directorio de esquemas ni uno de fixtures, y ninguna lleva la etiqueta de datos; ninguna nombra la carpeta
  de la evidencia por su ruta, ni la variable de grabación, ni una acción de plataforma, ni pide nada a una persona;
  ninguna tarea de aceptación, como pide «Aceptación e2e: no aplica». El único defecto que da hoy es que falta
  `gates/analyze.md`, que escribe el paso de análisis después de este fichero.
- a · `analisis_critico`: lo escribe el paso de análisis después de este fichero.
- b · `trazabilidad`: cada tarea cita sus FR y SC; «Requisitos → tareas» cubre FR-001 a FR-112 y SC-001 a SC-015; y
  «Controles de umbral» da la tarea y el test de cada fila de plan.md.
- c · `rebanadas_verdes`: cada tarea trae su test y su implementación; las excepciones están declaradas arriba; el
  orden es el del plan y la tabla de dependencias no tiene ninguna hacia delante.
- d · `rutas_declaradas`: tras «Rutas:», ficheros concretos o la carpeta nueva del juez; ninguna ruta genérica.
- e · `datos_separados`: ninguna tarea toca el directorio de esquemas ni uno de fixtures; ningún dato externo, ninguna
  grabación.
- f · `dod`: tabla «Definition of Done → tareas»; `CHANGELOG.md` en T007; sin ADR ni `SOURCES.md`, con su motivo.
- g · `checklist_veraz`: `checklists/requirements.md` no cambia.
- h · `aceptacion_primero`: el plan justifica «no aplica»; las evals de la skill van en T003, antes que `SKILL.md`.
- i · `autonomia`: ninguna tarea exige a una persona, publica, ni mide en la plataforma.
