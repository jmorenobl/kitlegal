# Tasks: H24 · Las evals juzgan el significado con un modelo: «afirma lo que no ha leído» decide, `boe-legislacion` deja de glosar lo que no ha leído, y la lista de expresiones deja de decidir

**Input**: `specs/017-h24-las-evals-juzgan/` (spec.md, plan.md, research.md, data-model.md, contracts/, quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` 2.11.0 aplicada (umbral de materialidad,
criterio de uso, proporcionalidad y convergencia en «Gates»; controles de umbral del ADR 0029; ADR 0030, 0032 y 0037);
H0-H7.4, H19, H21 y H22 en `main` (`internal/evals` con el formato común, la lista de expresiones, `umbrales` con el
contrato del ADR 0029, el job en dos modos con su definición comprobada en `make ci`, la tanda por commit y el sondeo;
`boe-legislacion` v0.1.6; el modelo que decide, `claude-sonnet-5-5`, ADR 0031), y la evidencia del ADR 0037 versionada
(la rúbrica, el esquema de la respuesta, los 259 casos y la medida). **Ninguna tarea crea el esqueleto ni los gates:
`make ci` existe y pasa**, y cada tarea lo deja en verde al terminar.

**Aceptación**: el plan dice **«Aceptación e2e: no aplica»** (plan.md, «Aceptación e2e»; research D24): el hito no
cambia el binario —cambiarlo está fuera de alcance y `cmd/kitlegal` no enlaza el paquete de evals—, así que no hay
comportamiento de `kitlegal` que un guion `testscript` describa, y **no hay tarea de aceptación**. El hito cambia una
skill y no añade ni cambia ninguna eval (spec, «Fuera de alcance»): lo que va antes que `SKILL.md` (Definition of Done
§1.10) es el instrumento que la mide, entero —el juez, sus umbrales, la medida y el job (T001 a T014)—, y la skill
llega en T015. La aceptación del hito es la del job de evals que el workflow lanza en el cierre tras la revisión final
(SC-001, FR-113). En `make ci` la fijan, sin modelo, los tests de la tabla de plan.md, «Aceptación e2e», que traen las
tareas. SC-014 lo mide una persona después del run.

**Tests**: obligatorios (constitución §III; spec FR-100 a FR-112; plan.md, «Controles mecánicos»). Cada tarea de código
trae su test y la implementación mínima que lo hace pasar, en el mismo diff; los nombres de test y sus ficheros son los
de plan.md, porque el informe final busca cada control `ci:<ruta>:<Test>` como `^func <Test>\(` en esa ruta.

**Rebanadas verticales y excepciones declaradas.** Son rebanadas completas —su test y su implementación en el mismo
diff, `make ci` en verde por sí solas— T002 a T007 y T009 a T014. Las demás, cada una por su razón y con su
verificación:

- **T001, T008 y T015** son las de datos: tocan `schemas/` o un `testdata/`, y llevan además solo lo que el dato obliga
  a cambiar a la vez (plan.md, *Complexity Tracking*, segunda fila; precedente, H7.2 a H7.4). T001: el esquema de las
  clases, la carpeta del juez, el tipo y la lectura que la entienden y los casos que los fijan —con la carpeta en el
  árbol y la lectura de hoy, `make ci` queda en rojo—. T008: los dos ficheros restaurados y la entrada del control de
  derivación que fija el segundo. T015: la clave nueva del esquema y de la lista, el tipo que la lee, la prosa que la
  aplica y `SKILL.md` v0.1.7 —con la clave nueva, la prosa de v0.1.6 da cinco defectos (research V14)—.
- **T016** es documentación: su verificación es que cada fichero, test, clave, variable, objetivo y etiqueta que nombra
  existe en el árbol, y `make ci`.
- **T017** es el cierre de la Definition of Done, sin código de producto: escribe la lista de lo tocado, mide la
  cobertura y ejecuta quickstart.md §1 a §6; solo añade tests si una cifra de cobertura queda bajo su umbral.

**El árbol entre T006 y T011.** Desde T006, `EscribirInforme` pide un votante para una skill con juez, y el punto de
entrada del job no se lo da hasta T011. `make ci` sigue en verde en cada tarea (el punto de entrada solo se compila y
se lintea: lleva la etiqueta `evals`), y el job no se ejecuta dentro del run: lo lanza el workflow en el cierre, con el
árbol entero.

**Modo**: desatendido (ADR 0018). El workflow `hito` ejecuta y verifica las tareas **una a una**, con guardián de diff:
cada tarea declara en su propia línea, tras «Rutas:», **todas** las rutas que crea, modifica o retira, y solo toca esas
(más `go.mod`, `go.sum`, `CHANGELOG.md`, el directorio del feature y el `x_test.go` de cada `x.go` declarado). Los
demás nombres que aparecen en una línea son contexto: referencias a los documentos del feature (contracts/, plan.md…)
y nombres de ficheros que la tarea solo lee, que se nombran sin su directorio. El cierre en la plataforma —publicar la
rama, abrir la propuesta de cambio y medir CI y el job de evals— lo hace el workflow tras la revisión final;
quickstart.md §7 y §9 quedan para una persona y §8 para el workflow.

## Formato: `- [ ] Tnnn [etiqueta?] [Story?] Descripción — FR/SC. Rutas: …`

- **[P]**: ninguna tarea lo lleva; en el bucle del workflow todo va en secuencia.
- **Aceptación**: ninguna tarea lleva esa etiqueta (plan.md, «Aceptación e2e: no aplica»).
- **[datos]**: la tarea crea o modifica ficheros bajo `schemas/` o bajo un `testdata/`. Son tres: T001 (el esquema de
  las clases del juez), T008 (la eval retirada y su grafo previo) y T015 (el esquema de la lista). H24 no graba nada de
  ninguna fuente (plan.md, «Datos externos»; research D25): no hay manifiesto `grabaciones.json` nuevo, ni test de
  grabación, ni nada que grabar para el paso `grabar_datos`. Los dos ficheros de T008 salen de la historia del
  repositorio y el segundo es una derivada de una grabación de H4, que un control comprueba; las copias de T001 son
  ficheros ya versionados; y ningún fichero de `data/` cambia. Los transcripts, las sesiones y las salidas del juez de
  los tests son constantes del test o ficheros que el test escribe en `t.TempDir()`: no son datos externos ni ficheros
  de `testdata/`.
- **[Story]**: US1 quien pregunta no lee nada de un precepto que nadie leyó · US2 el job marca la respuesta que afirma
  lo que no ha leído · US3 el juez solo decide si está medido · US4 una persona repite la medida cuando cambia el
  instrumento · US5 quien lee el informe ve por qué se marcó cada respuesta · US6 la lista de expresiones deja de
  decidir · US7 el sondeo juzga con el juez, sin veredicto. La documentación y el cierre no llevan historia.

## Batería de verificación por tarea

- **`make ci` en primer plano** y la tarea marcada `[X]` en el mismo turno en que termina en verde: formato · lint
  (`depguard`, `forbidigo`, `gosec`, `misspell`, `revive`, `dupl`, `gocyclo`, `paralleltest`…, también sobre los
  ficheros con la etiqueta `evals`) · tests con `-race` · `test-integration` · `test-tiempos` · `vuln` · `schema-check`
  · `skills-check` · `goreleaser-check` · secretos · módulos. Ningún registro de la verificación en la raíz del
  repositorio. Las salidas de `go test` y de `make` se leen con `rtk proxy` o con sondas positivas, porque el proxy de
  la sesión las resume.
- **Sin red ni modelo** (FR-093; contracts/job-de-evals.md §6): ninguna tarea usa la red salvo la de las herramientas
  de Go (`make vuln`); ninguna ejecuta `claude`, `make evals`, `make evals-sondeo`, `make evals-medir-juez`, los guiones
  `evals*.sh` de `scripts`, `TestEjecucionDelJob`, `TestMedidaDelJuez` ni `TestSondeo`. Los tests de `make ci` usan
  votantes de salidas grabadas y los sustitutos de `claude` y de `go`, que son los únicos que ejecutan los guiones.
- **Sin personas a mitad**: ninguna tarea la hace, la revisa ni la desbloquea una persona. La medida del juez y el
  sondeo no los lanza el run (FR-045, FR-050).
- **El instrumento no se cambia** (FR-036, FR-045): la rúbrica, el esquema de la respuesta, los casos y la medida
  llegan validados. Sus copias se escriben una vez, en T001, y ninguna tarea posterior las toca; ninguna tarea escribe
  en la carpeta de la evidencia del ADR 0037. Ninguna tarea ni corrección cumple un umbral de plan.md, «Controles de
  umbral», rebajándolo, dejándolo en `decide: false`, sacando respuestas o casos de su total, retirando su test o
  quitando una fila de una tabla de casos. Cada control nuevo se ve en rojo antes de darlo por bueno: con sus casos
  negativos o con un mutante temporal de lo que fija, que no queda en el diff.
- **Ningún test escribe fuera de `t.TempDir()`**; lo temporal de una tarea va bajo `$TMPDIR`, nunca bajo `/tmp`. Los
  fixtures nuevos de los tests son constantes del `_test.go` o ficheros escritos en `t.TempDir()`.
- **Sin atajos** (plan.md, «Comprobación contra la rúbrica», g): ningún `//nolint`, ningún `t.Skip`, TODO ni error
  silenciado; ningún test se desactiva. Si `dupl`, `gocyclo`, `paralleltest`, `gosec` o `misspell` marcan algo, se
  reestructura o se renombra (el guion como único argumento variable de `exec.CommandContext`; una función por
  subprueba desde una tabla; `t.Context()` solo en la llamada).
- **Antes de tocar un tipo, una clave o una lista que construyen los tests**, `grep` de sus usos en todos los
  `_test.go` del paquete y en el test de la release de la raíz, que lee los flujos: lo que haya que alinear entra en la
  misma tarea, y sus ficheros están en sus rutas.
- **Proporcionalidad** (constitución, «Gates»; plan.md, «Trazabilidad», último párrafo): ninguna tarea añade un caso,
  un mensaje ni un test para un estado sin vía real —un transcript, una medida, unos casos o un informe versionado que
  no se pueden leer siguen las reglas que ya hay—, ni reintenta o espera un voto más allá de la repetición del nulo.
- **Lo que no cambia** (spec, «Fuera de alcance»; FR-096): el binario, sus órdenes y `--describe`; `legal-core`; las
  evals, sus preguntas y su juicio sin modelo; el modelo que decide, las repeticiones y la regla por serie; las formas
  fijas de `avisos.go`; `SOURCES.md`; los ADR; los specs, los planes y los informes de H7.1 a H22; y el workflow `hito`
  con sus guiones.

---

## Phase 1: Setup

No hay tareas: el repositorio, el `Makefile` y los gates existen desde H0, y el hito no añade ninguna dependencia
(plan.md, «Primary Dependencies»).

---

## Phase 2: Fundación

No hay tareas sin historia: lo que bloquea a todas las demás es la carpeta del juez, que es ya el escenario 5 de US3 y
va en la fase siguiente.

---

## Phase 3: User Story 3 (primera parte) — La carpeta del juez (P1)

**Objetivo**: las clases del juez declaradas en datos y las cuatro copias de lo que el job lee, idénticas a sus
originales (plan.md, paso 1).

**Prueba independiente**: `go test -count=1 -run '^(TestLeerConjunto|TestEvalsDelRepositorio|TestCopiasDelJuez)$' ./internal/evals/`.

- [X] T001 [datos] [US3] La carpeta del juez de `boe-legislacion` con su declaración de clases, el esquema de esa declaración y su lectura en `LeerConjunto`, con los casos que los fijan, y **nada más** (contracts/juez-y-voto.md §1; contracts/medida-del-juez.md §3; data-model §1 y §7; research D19, D20, V17; plan.md, «Orden de implementación», paso 1, y *Complexity Tracking*, segunda fila: con la carpeta en el árbol y la lectura de hoy, la subprueba `formato` de `TestEvalsDelRepositorio` da un fichero mal formado, y sin el esquema la declaración no se puede validar, así que van juntos). **El esquema**, nuevo, con la cabecera (`$schema`, `$id`, `title`) de la forma de sus vecinos `*.yaml.json`: un objeto sin más claves que `clases`, obligatoria; `clases`, una lista de al menos un objeto con exactamente `nombre` (`^[a-z0-9_]+$`), `decide` (booleano) y `umbral` (número entre 0 y 1), los tres obligatorios. **La carpeta**, nueva: `clases.yaml`, con un comentario de cabecera que dice qué es y las dos clases de contracts/juez-y-voto.md §1 carácter a carácter (`afirma_lo_no_leido`, `decide: true`, `umbral: 0`; `cuenta_su_proceso`, `decide: false`, `umbral: 0`); y `rubrica.md`, `esquema.json`, `casos.yaml` y `medida.json`, copiados una sola vez con `cp` desde la carpeta de la evidencia del ADR 0037 (la pareja de carpetas, en contracts/medida-del-juez.md §3), sin abrirlos con un editor, y comparados con `cmp` antes de terminar; la carpeta de la evidencia no se escribe. **La lectura**: en `juez.go`, nuevo, `ClaseDelJuez{Nombre, Decide, Umbral}` y `Juez{Clases, Rubrica, Esquema, Casos, Medida}` (data-model §1: de la rúbrica y del esquema, su contenido entero; de los casos y de la medida, su ruta) y la función que lee la carpeta, que valida `clases.yaml` con el lector común de YAML y su esquema compilado con `compilarEsquemaPublicado`, como la lista; en `conjunto.go`, `Conjunto.Juez *Juez` y `LeerConjunto`, que deja de tomar la entrada `juez` por un fichero de eval: si es un directorio, lee de él el juez; si no lo es, es un fichero mal formado. Son mal formados, con su nombre delante (`juez/clases.yaml: …`), como una eval: `clases.yaml` que no cumple su esquema o repite un nombre; un `esquema.json` cuyas propiedades no son exactamente las clases declaradas; y cualquiera de los cinco que falta o no se puede leer. Una carpeta sin `juez` deja `Juez` en `nil`: esa skill no tiene juez. **Tests**: `TestLeerConjunto` gana los casos de la carpeta `juez`, escritos por el test en `t.TempDir()` —bien formada, con sus dos clases en su orden, la rúbrica y el esquema enteros y las dos rutas; cada mal formado de arriba, uno por caso, con el esquema incumplido de cinco maneras (clave de más, sin `nombre`, nombre con mayúscula, `umbral` 2, lista vacía); `juez` como fichero regular; y sin `juez`—; `formato` de `TestEvalsDelRepositorio` sigue sin ficheros mal formados y exige `Juez` con las dos clases en `boe-legislacion` y `nil` en `legal-core` (FR-021); y `TestCopiasDelJuez`, en `medida_test.go`, nuevo: compara byte a byte las cuatro copias con sus originales, con la pareja de carpetas como tabla del test, y falla nombrando cada fichero que difiere o falta; una skill con `juez` que no esté en la tabla lo hace fallar; y, sobre una copia de las dos carpetas en `t.TempDir()`, un byte cambiado en cada una de las cuatro la nombra: 4 de 4 (control de umbral de FR-023, FR-108 y SC-008). Antes de escribir: `grep` de quien recorre el directorio de evals de una skill en el paquete y en `scripts`, por si algo lee sus entradas sin `LeerConjunto`. — FR-020, FR-021, FR-022, FR-023, FR-040, FR-108; SC-008. Rutas: schemas/juez-clases.yaml.json, evals/boe-legislacion/juez/, internal/evals/juez.go, internal/evals/conjunto.go, internal/evals/medida_test.go

**Checkpoint**: `make ci` lee la carpeta del juez como parte del conjunto y falla si una copia difiere de su original.

---

## Phase 4: User Story 2 (primera parte) — El voto y la regla (P1)

**Objetivo**: todo lo que decide si una respuesta queda marcada, sin ningún proceso; y después el proceso que abre un
voto con la orden de la validación (plan.md, pasos 2 y 3).

**Prueba independiente**: quickstart.md §2.

- [X] T002 [US2] Los textos de las herramientas de una sesión, el mensaje del voto, la frase, la lectura del voto y la regla, sin ningún proceso (contracts/juez-y-voto.md §2, §3 y §5 a §9; data-model §2 y §3; research D2, D3, D5, D8, D9, V1, V3, V5, V6; plan.md paso 2). **`sesion.go`**: `Texto{Orden, Salida}` y `Sesion.Textos`, que `LeerSesion` deja en el orden de los bloques `tool_use` del transcript con la regla de contracts §2, la misma en los dos modos: de un `Bash` cuyo `input.command` lleva la palabra `kitlegal`, la orden tal cual y los textos del `content` de su `tool_result` (`textosDelContenido`) unidos por un salto de línea; de una herramienta del registro, con el prefijo que le ponga el agente, lo que el informe publica en `invocaciones[].orden` de esa llamada y lo mismo de su resultado; una orden o una llamada que falló deja su texto, con el error; un `tool_use` sin `tool_result`, ninguno; y nada de `Skill`, de `Read`, de otra herramienta ni de un Bash que no nombra `kitlegal`. Los casos de `TestLeerSesion` y de sus vecinos que comparan la sesión entera ganan sus `Textos`. **`juez.go`**: `mensajeDelVoto(pregunta, respuesta, textos)`, el de contracts §3 carácter a carácter —un `<texto orden="…">` por texto con su salida sin blancos en los extremos o `(sin salida)`, la línea `(ninguna herramienta devolvió ningún texto)` si no hay ninguno, la pregunta, la respuesta y la frase final «Responde a las dos preguntas de la rúbrica sobre esta respuesta.»—; `fraseEsta(frase, respuesta)`, la de contracts §6: fuera `*`, `_` y el acento grave, cada serie de blancos de `unicode.IsSpace` más U+001C a U+001F en un espacio, extremos recortados, y subcadena sin quedar vacía, sin usar ni tocar las piezas de `avisos.go`; `Votante`, `func(mensaje string) ([]byte, error)`, con un error propio para el tope agotado; la lectura del voto de contracts §5 —objeto JSON, `is_error`, `structured_output` o `result` sin la valla de código que lo envuelva, y validación contra el `Esquema` del juez—, con sus cuatro motivos de un voto que no llega a darse, que nombran el voto, en una línea y cortados a 300 caracteres; el voto nulo de §7, del voto entero, que se repite una vez; y la regla de §8, genérica sobre las clases del `Juez`: se vota por orden mientras alguna clase que decide tenga todos sus votos en sí con su frase, la clase que solo se publica cuenta con el primero, una respuesta queda marcada solo con tres síes, y queda «sin juzgar» con su motivo en cuanto un voto no llega. Ningún proceso todavía. **Tests**, con transcripts y salidas del juez escritos como constantes del test y un votante que las devuelve y cuenta sus llamadas: `TestTextosDeLaSesion` (`sesion_test.go`: modo orden y modo herramienta en su orden, la orden que falla, ninguno de `Skill` ni de un Bash sin `kitlegal`, y una sesión sin textos); `TestMensajeDelVoto` (byte a byte con textos de los dos modos y sin ellos, la salida vacía, y 0 bytes de `SKILL.md`, de la eval salvo su pregunta y del juicio sin modelo: se compone desde una eval y un resultado con cadenas centinela en sus comandos, sus citas, sus avisos, sus hallazgos y sus motivos, y ninguna aparece); `TestFraseEnLaRespuesta` (literal; cruza un salto de línea; pierde un acento grave; pierde `*` y `_`; lleva un espacio de no separación; vacía; y cambia una mayúscula, un acento o una coma, que no están); `TestVotoDelJuez` (los 4 casos de SC-002: sí con su frase, vale; sí con una frase que no está, nulo y repetido exactamente 1 vez; sí con una frase que solo difiere en blancos y énfasis, vale; y los cuatro motivos de un voto que no llega, con el del tope); y `TestReglaDeLosVotos` (las cinco filas de §8 con sus votos pedidos —1, 2, 3, 3 y 4—, «sí, sí y no» sin marcar con sus dos frases, y la clase que solo se publica con 1 voto). Controles de umbral de FR-102, FR-103 y FR-107: cada tabla se ve en rojo con un mutante temporal de lo que fija (la regla que marca con dos síes; la frase que no colapsa el salto de línea), que no queda en el diff. — FR-001, FR-002, FR-004, FR-005, FR-006, FR-007, FR-010, FR-011, FR-102, FR-103, FR-107; SC-002, SC-003, SC-007. Rutas: internal/evals/sesion.go, internal/evals/juez.go
- [X] T003 [US2] El guion que abre un voto y el votante que lo ejecuta, con su tope (contracts/juez-y-voto.md §4 y §9; research D4, D6, V7, V8, V22, S5; plan.md paso 3). **`evals-voto.sh`**, nuevo, ejecutable y sin nada posterior a bash 3.2: lee `../modelo.txt`, `../rubrica.md` entero con su salto de línea final y `../esquema.json` sin el suyo, relativos a su directorio de trabajo, y ejecuta con `exec`, con los argumentos de contracts §4 y en su orden: `claude -p --model "$modelo" --tools "" --strict-mcp-config --disable-slash-commands --no-session-persistence --system-prompt "$rubrica" --output-format json --json-schema "$esquema"`; no recibe argumentos ni lee ninguna variable más. **`juez.go`**: las constantes del tope de un voto, 35 s, y de su margen, 5 s; y el constructor del votante que ejecuta el guion —con el `Juez`, el id del modelo, la ruta absoluta del guion, el `PATH` y la credencial que le da quien lo llama, y un contexto—: crea el directorio del juez, un temporal por ejecución que retira al acabar, con `modelo.txt`, `rubrica.md`, `esquema.json`, `cwd` vacío y `config`; cada voto es un proceso nuevo con el guion como único argumento variable, el mensaje por la entrada estándar, `cwd` como directorio de trabajo y un entorno de solo cuatro variables (`PATH`, `HOME`, `CLAUDE_CONFIG_DIR` y `CLAUDE_CODE_OAUTH_TOKEN`); pasado el tope, el proceso se termina y a los 5 s se cierran sus tuberías (`WaitDelay`), y el votante devuelve el error del tope de T002; el contexto de cada voto deriva del recibido, con su `cancel` en todos los caminos. **Tests** en `juez_test.go`, con un `claude` sustituto de `sustitutos_test.go` que anota sus argumentos, su entrada, su directorio y su entorno y escribe una salida grabada: `TestOrdenDelVoto` exige los argumentos uno a uno y en su orden —con la rúbrica entera, su salto de línea final incluido, y el esquema sin el suyo—, la entrada estándar igual al mensaje, el directorio `cwd`, y las cuatro variables y ninguna más; y, con un tope de prueba y un sustituto que espera, que el votante devuelve el error del tope sin dejar el proceso vivo. La credencial de la prueba es un valor sin forma de secreto. El guion solo lo ejecutan los tests, con el sustituto delante en su `PATH`: la tarea no lo lanza a mano ni abre ninguna sesión con modelo. — FR-003, FR-004, FR-007, FR-093, FR-107; SC-007. Rutas: scripts/evals-voto.sh, internal/evals/juez.go, internal/evals/sustitutos_test.go

**Checkpoint**: la regla entera y la orden del voto están fijadas sin haber abierto ninguna sesión con modelo.

---

## Phase 5: User Story 3 (segunda parte) — La medida versionada corresponde, o `make ci` falla (P1)

**Objetivo**: quien cambia la rúbrica, los casos, el modelo del juez o su versión de Claude Code no pasa `make ci` sin
su medida (plan.md, paso 4).

**Prueba independiente**: `go test -count=1 -run '^TestMedidaVersionada$' ./internal/evals/`.

- [X] T004 [US3] El modelo del juez y la versión de Claude Code de sus votos fijados en la definición del job, y la comprobación sin modelo de que la medida versionada les corresponde (contracts/medida-del-juez.md §1, §2 y §8; contracts/job-de-evals.md §1; data-model §4 y §6; research D11, D19, D21; plan.md paso 4). **`evals.yml`**: en `env` del trabajo `evals`, junto a `MODELO_DE_EVALS` y `VERSION_DE_CLAUDE_CODE`, que no cambian, `MODELO_DEL_JUEZ: claude-opus-5-5` y `VERSION_DE_CLAUDE_CODE_DEL_JUEZ: 2.1.289`, cada una con su comentario (el modelo del juez, distinto del que decide y fijado por su id completo, ADR 0037; la versión de sus votos, que es la de la medida y va aparte de la de las sesiones: subir una no obliga a repetir la medida de la otra); nada más del fichero en esta tarea. **`definicion.go`**: `DefinicionDelJob.ModeloDelJuez` y `VersionDelJuez`, leídos de esas dos variables. **`medida.go`**, nuevo: `MedidaDelJuez` con las claves de contracts §1 que se leen (`clase`, `fecha`, `modelo_del_juez`, `version_de_claude_code`, `rubrica.sha256`, `casos.sha256`, `defectos.casos`, `defectos.sin_marcar`, `correctos.casos` y `correctos.marcados`; las demás no se leen), su lectura sin esquema publicado, y `comprobarLaMedida(juez, modeloFijado, versionFijada)`, que devuelve una línea por lo que falla, con los textos de contracts §2 carácter a carácter —la huella SHA-256 de `juez/rubrica.md`, la de `juez/casos.yaml`, el modelo, la versión, la clase frente a las que deciden de `clases.yaml`, y cada recuento distinto de 0—, y ninguna si corresponde y se cumple; una medida que no se puede leer, o sin alguna de esas claves, no corresponde; no usa ningún modelo ni abre ningún proceso. **`TestMedidaVersionada`** (`medida_test.go`): `del-repositorio`, con el juez de la carpeta del repositorio y los dos valores de `leerDefinicionDelJob`, ninguna línea; y, sobre una copia de la carpeta del juez en `t.TempDir()`, la rúbrica, los casos, el modelo fijado y la versión fijada cambiados, uno cada vez, y cada uno de los dos recuentos distinto de 0: seis de seis mutaciones dan la línea que las nombra (control de umbral de FR-042, FR-105 y SC-005, y de los dos de `medida_del_juez:afirma_lo_no_leido:…` en `make ci`). Las copias del repositorio no se tocan (FR-045). Antes de tocar `DefinicionDelJob`: `grep` de sus usos en los `_test.go` del paquete y de las lecturas de este flujo en el test de la release de la raíz, que solo se leen. — FR-032, FR-041, FR-042, FR-090, FR-091, FR-105; SC-005. Rutas: .github/workflows/evals.yml, internal/evals/definicion.go, internal/evals/medida.go

**Checkpoint**: con cualquiera de las cuatro claves cambiada, o con un recuento distinto de 0, `make ci` falla diciendo
cuál.

---

## Phase 6: User Story 6 — La lista de expresiones deja de decidir (P2)

**Objetivo**: ninguna respuesta deja de pasar por una expresión, y los umbrales de las respuestas dependen del juez y
no de la lista (plan.md, paso 5). Va antes que el juez en el informe porque `umbrales` cambia de condición.

**Prueba independiente**: `go test -count=1 -run '^(TestJuzgarSinLaLista|TestUmbralesDelInforme)$' ./internal/evals/`.

- [X] T005 [US6] La lista de expresiones deja de juzgar respuestas, y los umbrales de las respuestas pasan a depender de que la skill tenga juez (contracts/informe-del-job.md §2, §6, §7 y §9; data-model §5; research D13, D14, V10, V19, V20; plan.md paso 5). **Sale la lista** de `Juzgar` y de `ResultadoDeEval` (`ExpresionesProhibidas`, su parte de `Pasa` y su motivo: `juzgar.go`); de `Eval.Prohibidas` y del paso de `LeerConjunto` que la reparte a cada eval (`formato.go`, `conjunto.go`); del informe —`ExpresionesProhibidasPorModelo`, `RecuentoDeExpresiones`, la sección «Expresiones prohibidas por modelo» y la columna «Expresiones prohibidas» de `informe.md`, con el recuento por modelo y modo reducido a las respuestas y a las que no activaron la skill (`informe.go`)—; de `umbrales` —los elementos `expresiones_prohibidas:…` y `redaccion_no_leida:…`, con sus constantes y sus descripciones (`umbrales.go`)—; del sondeo —`SondeoAJuzgar.Prohibidas`, el recuento y sus dos líneas (`sondeo.go`)—; y de `comprobarConsultaRepetida`, que pierde su parámetro y su condición de las expresiones, con `TestComprobarConsultaRepetida`, que deja de leer la lista (`consulta_repetida.go`, `job_test.go`). De `prohibidas.go` sale solo lo que queda sin uso (`esDeLaClaseB`); se quedan el fichero de la lista, su esquema, su lectura en `Conjunto.Prohibidas`, `ExtraerExpresionesProhibidas` y la comprobación de la prosa. **D13**: `sin_activar:<modelo>:<modo>` existe si `Conjunto.Juez` no es `nil`, y no si la skill tiene lista; `duracion_de_las_sesiones:<modo>` sigue como hoy; tras esta tarea, `umbrales` de una skill con juez lleva esos cuatro elementos, y el de una sin juez, `[]`. **Tests**: `TestJuzgarSinLaLista` (`juzgar_test.go`), con el conjunto que `LeerConjunto` lee de la carpeta de evals de `boe-legislacion` y su lista (`listaDelRepositorio`), en dos partes. **(a)** Las dos respuestas de la eval sin binario ni servidor del 2026-10-04 —con la oferta de consultar después que ya cita ese fichero de test—, en sesiones que cumplen todo lo demás de su eval, pasan. **(b)** Las respuestas del calibrado de H7.4: de cada uno de los tres informes versionados del calibrado —los del cierre de H7.1, H7.2 y H7.3, con sus 93 entradas, que solo se leen—, las entradas cuya `respuesta` marca la lista del repositorio aplicada en el test con `ExtraerExpresionesProhibidas`; no se eligen por la clave `expresiones_prohibidas` de la entrada, que el informe de H7.1 no lleva en ninguna de sus 93 y que solo llevan no vacía 10 del de H7.2 y 1 del de H7.3. Premisas, exigidas con `require` para que el test no pase en vacío: son 36, 11 y 9, en ese orden, 56 en total; las 56 tienen `activa: true` en su entrada, porque la lista solo juzgaba las evals que activan la skill; y el fichero que nombra la clave `eval` de la entrada está en el conjunto en 54 y falta en 2, las dos del informe de H7.1 y de la eval 19 de ese informe, la que retiró H7.2, cuyo fichero no vuelve al árbol hasta T008. Cada una se juzga con `Juzgar` y una sesión terminada hecha de su entrada: su `respuesta` tal cual y la skill activada si su `activada` lo dice, sin invocaciones —lo que falte de lo que su eval espera da sus motivos de siempre, que no son de la lista—. La eval de las 54 es la del conjunto con ese nombre de fichero, como la deja `LeerConjunto`; la de las 2, una eval que solo declara lo que su entrada dice de ella —`Fichero`, el `eval` de la entrada, leído de ella y nunca escrito en el test, que nombra las evals por sus dos cifras, y `Activa`, su `activa`—: basta para lo que se comprueba, y el test no lee nada de la carpeta que restaura T008. De las 56, el resultado codificado en JSON no tiene la clave `expresiones_prohibidas`, y ningún motivo empieza por `expresión prohibida: `, el principio del motivo que sale, escrito a mano en el test como los principios vecinos; no se busca la expresión dentro de los motivos, porque `graph check` es una expresión de la lista y también parte del texto de un comando ausente: 0 sesiones marcadas por la lista (control de umbral de FR-111 y SC-011). Se ve en rojo con el árbol de antes de la tarea, donde el test compila porque no nombra el campo, la constante ni `Eval.Prohibidas`, que salen: los 56 resultados llevan la clave, y los 54 juzgados con la eval del conjunto, además, sus expresiones y sus motivos. **Del calibrado que se retira se quedan**, para este test: las rutas de los tres informes (`informeDeH71`, `informeDeH72`, `informeDeH73`) y `respuestasDeCadaInforme`, con su comentario al día, y la lectura de las entradas de un informe que hoy está dentro de `marcadasEnElInforme`, que pasa a un auxiliar que exige las 93 y da de cada entrada `sesion`, `eval`, `activa`, `activada` y `respuesta`; y `listaDelRepositorio`, que usa también `prosa-de-la-skill`. Salen las subpruebas `expresiones-calibradas`, `expresiones-en-los-bloques` y `expresiones-de-la-skill` de `TestEvalsDelRepositorio`, con lo demás que solo ellas usan —del calibrado, el reparto por eval y por familia (`informesCalibrados` con su tipo, `marcadasPorFamilia`, `columnasDelCalibrado`, `distintasDelCalibrado`, `marcadaPor`); `expresionesEn` y su tipo se quedan, que los usa la comprobación de la prosa—; `TestJuzgarLasExpresionesProhibidas`, `TestJuzgarLasClasesDeLaRespuesta`, `TestInformeConExpresionesProhibidas` y `TestInformeConExpresionesEnUnaSerie`, cuyo único objeto era ese juicio, se retiran con sus auxiliares, y lo que alguno fije de otra cosa pasa al test vecino; `TestInformeConListaMalFormada` se queda, porque la lista mal formada sigue siendo un fichero mal formado; y `TestUmbralesDelInforme`, los `TestInforme…`, `TestInformeMarkdownDeLosUmbrales`, `TestInformeEnDosModos`, `TestJuicioDelSondeo`, `TestSalidaDelSondeo`, `TestSondear` y `TestCondicionesDeLaConsultaRepetida` se alinean: sin la lista en el resultado, en los motivos, en el recuento, en el Markdown ni en las líneas del sondeo, y con una carpeta `juez` en la skill sintética donde esperan `sin_activar`: esa carpeta la escribe el test en `t.TempDir()`, junto a una copia de las evals de su caso, y ningún fichero de los datos de prueba del paquete se crea ni se cambia. Los casos en que `sin_activar` y la duración de las sesiones no se cumplen siguen dando `fallo` (control de umbral de FR-034: los cuatro umbrales que siguen). `TestExtraerExpresionesProhibidas` y `prosa-de-la-skill` no cambian. Ningún test se salta ni se desactiva. El guion del informe final del workflow no se toca: lee la clave que sale con un valor por omisión. — FR-034, FR-062, FR-070, FR-071, FR-111; SC-011. Rutas: internal/evals/juzgar.go, internal/evals/formato.go, internal/evals/conjunto.go, internal/evals/informe.go, internal/evals/umbrales.go, internal/evals/sondeo.go, internal/evals/consulta_repetida.go, internal/evals/prohibidas.go, internal/evals/job_test.go

**Checkpoint**: `Juzgar` no mira la lista; el fichero queda como el vocabulario de la prosa.

---

## Phase 7: User Story 2 (segunda parte) y User Story 5 — El juez en el informe (P1, P2)

**Objetivo**: el job juzga cada respuesta con el juez, decide por sus umbrales y publica los votos y las frases
(plan.md, paso 6).

**Prueba independiente**: quickstart.md §3, sus dos primeros tests.

- [X] T006 [US2] El juez vota dentro de `EscribirInforme` y decide por sus umbrales (contracts/informe-del-job.md §1, §2, §4 y §9; data-model §3 y §5; research D6, D7, D9, D13; plan.md paso 6). **`informe.go`**: `InformeAEscribir` gana `Votar`, `ModeloDelJuez`, `VersionDelJuez`, `ConcurrenciaDelJuez` y `Ahora` (sin él, `time.Now`); con una skill con juez y sin votante es un error que impide escribir el informe, y con una skill sin juez el votante no se llama (FR-037). Entre el juicio sin modelo de cada sesión y la composición del informe se juzgan, por grupos y en este orden —modo orden, modo herramienta, eval sin binario ni servidor—, las respuestas del modelo que decide de las series que pide el plan cuya eval activa la skill, terminadas, legibles y medidas: las del `total` de `sin_activar` de cada modo y, aparte, las de la eval sin binario ni servidor; nunca las del modelo informativo, la de la prueba de red ni las sin medir, sin terminar o ilegibles. De cada una, el mensaje de T002 con la pregunta de su eval, su respuesta y sus `Textos`, y la regla de T002; como mucho `ConcurrenciaDelJuez` respuestas a la vez, con los votos de una misma respuesta uno detrás de otro; el reloj de cada modo va de que empieza su primer voto a que termina el último. El juicio sin modelo no cambia con los votos: ni `pasa`, ni sus motivos, ni la tasa de su serie (FR-014). **`umbrales.go`**: los doce elementos de contracts §2 en su orden, con su `nombre`, su `descripcion`, `comparacion` `"<="` y su `decide` —por modo, `sin_activar`, `afirma_lo_no_leido:<modelo>:<modo>` (las marcadas del grupo sobre las del grupo, con el umbral y el `decide` de `clases.yaml`) y `cuenta_su_proceso:<modelo>:<modo>` (las que tienen sí en su primer voto, `decide: false`); `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` y `…:correctos_marcados`, por cada clase que decide, con los recuentos y los totales de la medida versionada leída con lo de T004 (hoy 0 de 212 y 0 de 47), sin votar ningún caso (FR-044); `duracion_de_las_sesiones:<modo>`; y `duracion_del_juez:<modo>`, los segundos de los votos del grupo redondeados hacia arriba, sin `total` y con umbral 900—; una respuesta sin juzgar sigue en el `total` y no en la `medida`; ninguno suma dos modos ni cuenta la eval sin binario ni servidor, que no añade ningún elemento ni entra en ninguna duración (FR-013). **Motivos** de contracts §4, carácter a carácter: el de una clase que decide con alguna marcada, con los de los umbrales, que nombra cada respuesta marcada por su sesión con sus tres frases; el de las respuestas sin juzgar, detrás del de las sesiones sin medir; y el de `duracion_del_juez:<modo>`, con los de la duración; los dos últimos con el prefijo de la ejecución. Las funciones de `informe.go` que el sondeo llama conservan su firma. **Tests**, con sesiones y una skill sintéticas que el test escribe en `t.TempDir()` —sus transcripts con textos, su carpeta `juez` y sus evals, sin crear ni cambiar ningún fichero de los datos de prueba del paquete— y un votante de salidas grabadas que cuenta sus llamadas: `TestInformeConElJuez` (`informe_test.go`) —1 marcada en un modo: `fallo`, el umbral de ese modo con `cumple: false`, el del otro con `true`, y el motivo con la sesión y sus tres frases; 0 marcadas: se cumple y el veredicto es aprobado; `cuenta_su_proceso` con 3: `decide: false` y el mismo veredicto; un voto que no llega: `fallo` con el motivo de la ejecución, y la respuesta en el `total` y no en la `medida`; la eval sin binario ni servidor marcada con tres síes: frente a las mismas sesiones sin esa marca, los mismos `umbrales`, los mismos motivos y la misma sesión; 901 s del juez en un modo, con el reloj inyectado: `fallo` de la ejecución, y 900: se cumple; las sesiones del modelo informativo y la de la prueba de red, sin ningún voto; y una skill sin juez: `umbrales` `[]` y el votante sin llamadas— y `TestUmbralesDelInforme` (`umbrales_test.go`): los doce en su orden, con sus nombres, sus descripciones y su `decide`, diez con `true`, y ninguno de los dos que salieron en T005. Los tests de `EscribirInforme` que ya había con una skill con juez reciben su votante. Controles de umbral de FR-030, FR-032, FR-033, FR-007 y FR-104: cada caso de `fallo` es el test que ve fallar su umbral. Ningún umbral se rebaja ni pasa a `decide: false` (FR-036). — FR-007, FR-012, FR-013, FR-014, FR-030, FR-031, FR-032, FR-033, FR-035, FR-037, FR-044, FR-104; SC-004. Rutas: internal/evals/informe.go, internal/evals/umbrales.go, internal/evals/juez.go, internal/evals/medida.go
- [X] T007 [US5] El informe publica los votos y las frases de cada respuesta con algún voto afirmativo, y las que quedan sin juzgar (contracts/informe-del-job.md §3, §7 y §9; data-model §3 y §5; research D10; plan.md paso 6). **`informe.go`**: `Informe.Juez`, la clave `juez` de la raíz de `informe.json`, detrás de `umbrales` y `null` si la skill no tiene juez: `modelo` y `version_de_claude_code`, tal como se recibieron; `respuestas`, una entrada por respuesta juzgada con algún voto afirmativo en cualquier clase, también nulo y también si no quedó marcada, en orden de sesión, con `sesion` y `clases[]` en el orden de `clases.yaml` (`clase`, `marcada` y `votos[]`: `voto` de 1 a 3, `nulo`, los campos de `esquema.json` para esa clase tal como los dio el juez —`motivo`, `respuesta`, `frase` y, si la clase lo tiene, `precepto`— y `frase_en_la_respuesta`), con todos los votos de la respuesta, también los nulos; y `sin_juzgar`, una entrada por respuesta sin juzgar con `sesion` y `motivo`; las listas vacías son `[]`; de una respuesta sin ningún voto afirmativo no lleva nada; y las de la eval sin binario ni servidor van como las demás (FR-013). `marcada`, en una clase que solo se publica, es que su primer voto dice sí con su frase. **`informe.md`**: detrás de «Umbrales», la sección «Juez» de contracts §7 —el modelo y la versión, la tabla de los votos (Sesión, Clase, Voto, Nulo, Respuesta, Frase, En la respuesta, Precepto, Motivo, Marcada) o «ninguna», y la de `sin_juzgar` (Sesión, Motivo) o «ninguna»; sin juez, «La skill no tiene juez.»—. **Tests**: `TestInformeConElJuez` gana, con el votante grabado de T006: «sí, sí y no», sin marcar, con sus tres votos, sus dos frases y `marcada: false`; `cuenta_su_proceso` con 3, las tres en `juez.respuestas` con su frase; todos no, sin entrada; un sí nulo y su repetición, los dos publicados y el primero con `nulo: true`; un voto que no llega, en `sin_juzgar` con su sesión y su motivo; la eval sin binario ni servidor marcada, con sus 3 votos y sus 3 frases; y una skill sin juez, `juez` `null`. La entrada de «sí, sí y no» se compara con `JSONEq` con la forma del ejemplo de contracts §3, y los casos del Markdown van con los de `TestInformeMarkdownDeLosUmbrales`. — FR-013, FR-060, FR-061, FR-104; SC-004. Rutas: internal/evals/informe.go, internal/evals/juez.go

**Checkpoint**: con una respuesta marcada, el informe da `fallo` con la sesión y sus tres frases, y enseña los votos de
todas las que tuvieron alguno afirmativo.

---

## Phase 8: User Story 4 (primera parte) — Los casos y la ejecución de la medida (P2)

**Objetivo**: los 259 casos reconstruidos sin escribir nada a mano, y la ejecución que los vota y da la medida de lo
que hay (plan.md, pasos 7 y 8).

**Prueba independiente**: quickstart.md §4, sus dos primeras órdenes.

- [X] T008 [datos] [US4] La eval retirada y su grafo previo, restaurados de la historia del repositorio, con el control que fija la derivada, y **nada más** (contracts/medida-del-juez.md §4; data-model §7; research D15, D25, M2, V16, V21; plan.md paso 7 y *Complexity Tracking*, tercera fila). testdata/evals/retiradas/, nueva, con dos ficheros escritos con `git show c4819d1^:<origen>` redirigido a su destino, sin abrirlos con un editor, y comparados después byte a byte con su origen (`git show … | cmp - <destino>`): `19-lpac-articulo-21-redaccion-cambiada.yaml`, el fichero de ese nombre de la carpeta de evals de la skill en ese commit, y `grafo-previo/lpac-a21-version-anterior/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json`, el del mismo nombre de la carpeta de grafos previos de ese commit (los dos orígenes, en la tabla de contracts §4). No es una grabación nueva: ni red, ni manifiesto, ni test de grabación; el segundo es una derivada de la grabación de H4, que tiene fila revisada en `SOURCES.md`. internal/app/grafo_test.go: `TestGrabacionesDerivadas` recupera la entrada de esa derivada como estaba en `c4819d1^` —`versionDelArticulo21` con la fecha `20151002` y su párrafo sintético, leídos de la versión de este mismo fichero en ese commit— sobre la carpeta nueva, y la carpeta nueva entra en la comprobación de que sus ficheros son los de sus entradas, en los dos sentidos (`compruebaLaCarpetaConSusEntradas`): la derivada restaurada es, leída con `boe`, la grabación de H4 salvo exactamente lo que dice su nombre, y un fichero de más o de menos en la carpeta falla nombrado. La eval restaurada no entra en ningún conjunto ni plan: `LeerConjunto` no lee esa carpeta y nada la valida contra el esquema de eval de hoy. Antes de escribir: `grep` de quien recorre el directorio de las grabaciones de las evals en `internal`, por si alguna comprobación exige que sus subcarpetas sean las conocidas. Se ve en rojo con un mutante temporal (un byte de la derivada cambiado), que no queda en el diff. — FR-024, FR-109; SC-009. Rutas: testdata/evals/retiradas/, internal/app/grafo_test.go
- [ ] T009 [US4] Los casos etiquetados, la reconstrucción en proceso de sus textos y el control de derivaciones (contracts/medida-del-juez.md §4, §5, §6 y §8; data-model §4; research D15, D16, D17, M1, M2, M6, V2, V4; plan.md paso 8). **`medida.go`**: `CasoEtiquetado` (`Informe`, `Sesion`, `Quitado{Norma, Bloque}`, `Grupo`, `Etiqueta`, `Procedencia`, `Frase`) y la lectura de `juez/casos.yaml` con el lector común de YAML, sin esquema publicado; y la resolución de cada caso sin escribir nada a mano: la respuesta, la de su sesión en su informe versionado, tal cual; la pregunta, la del fichero de eval que nombra esa sesión —el de la carpeta de evals de la skill o, si la eval se retiró, el de la carpeta que dejó T008, de la que solo se leen `pregunta` y `grafo_previo`—; y los textos, la reconstrucción de contracts §5 con las `invocaciones` de esa sesión: (1) la base, `Preparar` con `UnionDeGrabaciones()` y las consultas de las evals de hoy, rehecha si tiene más de 100 s; (2) por sesión, un directorio de caché nuevo, con `prepararGrafoPrevio` sobre la carpeta de los grafos previos, o sobre la de la eval retirada, si su eval tiene grafo previo, y después los ficheros de la base; (3) cada invocación en su orden salvo `mcp serve` —las palabras de su `orden`; en una llamada, `<applet>_<verbo>` pasa a `<applet> <verbo>` y gana `--json`; todas ganan `--offline`—, ejecutada con `app.Main` sobre un registro con el applet `boe` (reproducción de un directorio vacío y la caché de la sesión), el applet `graph` y la entrega al grafo de esa caché, con su `orden` y lo que escribe en la salida estándar como texto, termine con el código que termine; y (4) en un derivado, fuera el texto de cada `boe articulo` de la norma y el bloque `quitado` que terminó con 0, y del de cada `boe articulos` que lo pidió, los elementos de `data` con ese `bloque` (si no queda ninguno, fuera el texto), con el sobre vuelto a escribir con sus seis claves en su orden. Sin red, sin modelo y sin el binario instalado; los grafos temporales se descartan. `preparar.go` y `grabaciones.go` cambian solo en lo que la reconstrucción necesita de ellos: la carpeta del grafo previo como parámetro y la constante de la carpeta de la eval retirada. **`TestGrabacionesDerivadas`** (`medida_test.go`), con los casos de la copia del repositorio: los 259 nombran un informe versionado y una sesión que está en él, y su respuesta es, byte a byte, la de esa sesión leída aparte; la pregunta de cada caso es la de su eval; y cada uno de los 140 derivados se diferencia de su sesión solo en el texto quitado —sus textos son los de la sesión, en su orden, menos los del bloque `quitado`; en el de un `boe articulos`, los mismos valores JSON menos esos elementos; y se quita al menos uno— (control de umbral de FR-109 y SC-009). Los recuentos de la copia son los de research M2: 259 casos, 212 `defecto` y 47 `correcto`, y 5 `bitacora`, 140 `derivado` y 114 `lectura`. La reconstrucción se hace una vez por proceso de test y la comparten los tests que la usan. Se ve en rojo con un mutante temporal (el derivado que no quita su texto), que no queda en el diff. Los informes versionados de H7.1 a H22 solo se leen. — FR-024, FR-051, FR-109; SC-009. Rutas: internal/evals/medida.go, internal/evals/preparar.go, internal/evals/grabaciones.go
- [ ] T010 [US4] La ejecución de la medida: votar los casos y dar la medida de lo que hay (contracts/medida-del-juez.md §7 y §8; data-model §4; research D18, M2; plan.md paso 8). **`medida.go`**: `medirAlJuez`, que recibe el juez de la skill, el votante, el id del modelo del juez, la versión de Claude Code de sus votos, cuántos casos se votan a la vez, el commit y la fecha, y: (1) no comprueba la medida versionada ni la lee para decidir nada (FR-043, FR-051); (2) reconstruye cada caso con lo de T009, sin preparar ni abrir ninguna sesión de evals; (3) vota cada caso con la regla de T002, como mucho ese número de casos a la vez: tres votos por cada defecto que se marca y uno por cada correcto que no; (4) devuelve el texto de la medida con la forma de contracts §7 —`skill`, `clase`, `fecha`, `modelo_del_juez`, `version_de_claude_code`, `rubrica` y `casos` con su `fichero` y la huella SHA-256 de la copia que ha leído, `defectos` y `correctos` con sus casos y sus recuentos, y `origen` con el commit—, con las cuatro claves de lo que hay y no las de la medida versionada; (5) si un defecto no queda marcado o un correcto queda marcado, devuelve además un error con una línea por caso, `<informe> <sesión> [sin <norma> <bloque>]: etiquetado <etiqueta> y <marcado | sin marcar>: «<frase>» · …`; y (6) con algún caso sin juzgar, devuelve el error con esos casos y su motivo y ningún texto de medida (FR-053). No escribe en la salida estándar ni en el repositorio (FR-054). **`TestEjecucionDeLaMedida`** (`medida_test.go`), con un votante que responde según la etiqueta del caso y cuenta sus llamadas, sin leer los votos de la evidencia (research D18): los 259 bien, sobre una copia de la carpeta del juez en `t.TempDir()` con la rúbrica cambiada, de modo que la medida versionada **no** corresponde: vota los 259, pide 683 votos y da la medida con las huellas de la copia, el modelo y la versión recibidos, y 0 de 212 y 0 de 47; un defecto sin marcar y un correcto marcado: error con el caso y sus frases; un voto que no llega: error sin medida; y en todos, 0 sesiones de evals: `medirAlJuez` no recibe a quien las abre ni un directorio de sesiones, y el temporal del test no gana ninguno. Ningún caso del test depende de que la medida versionada corresponda (control de umbral de FR-106 y SC-006). — FR-050, FR-051, FR-052, FR-053, FR-054, FR-106; SC-006. Rutas: internal/evals/medida.go

**Checkpoint**: con votos grabados, la ejecución de la medida da 0 de 212 y 0 de 47 con 683 votos, corresponda o no la
medida versionada.

---

## Phase 9: User Story 3 (tercera parte) y User Story 4 (segunda parte) — El job (P1, P2)

**Objetivo**: el job no abre ninguna sesión sin una medida que corresponda; la medida tiene su punto de entrada, su
guion y su trabajo, que solo se lanza con su etiqueta o con su entrada; y los topes cubren su peor caso (plan.md,
paso 9).

**Prueba independiente**: `go test -count=1 -run '^(TestEjecucionSinMedir|TestGuionDeLaMedida|TestDefinicionDelJob)$' ./internal/evals/` y quickstart.md §6.

- [ ] T011 [US3] El recorrido del job con la comprobación de la medida delante: sin una medida que corresponda no se abre ninguna sesión (contracts/informe-del-job.md §5 y §9; contracts/job-de-evals.md §2; data-model §5 y §6; research D1, D11; plan.md paso 9). **`informe.go`**: `InformeAEscribir.InstrumentoSinMedir`, las líneas de `comprobarLaMedida`; con alguna, `EscribirInforme` escribe el informe de contracts/informe-del-job.md §5 —`veredicto` `fallo`; un motivo por línea, `de la ejecución, no de la skill: el instrumento no está medido: <qué>`; `umbrales` `[]`; `juez` con su modelo y su versión y con `respuestas` y `sin_juzgar` vacías; y `tasas`, `evals`, `sesiones_sin_medir`, `fuera_de_lo_grabado` y `red` `[]`—, sin leer ninguna sesión ni llamar al votante. **`ejecucion.go`**, nuevo: `EjecucionDelJob` (la skill, el plan, quien abre las sesiones de una tanda como función, los directorios, el votante, el modelo y la versión del juez y lo demás que hoy reúne `TestEjecucionDelJob`) y `ejecutarElJob`: con una skill con juez hace `comprobarLaMedida` antes de nada; si da alguna línea, escribe ese informe sin llamar a quien abre las sesiones ni al votante; si no, abre las sesiones por tandas como hoy (`ejecutarPorTandas`) y escribe el informe con el votante, sin votar ningún caso etiquetado (FR-044); y con una skill sin juez no comprueba nada ni pide votante. **`job_test.go`**: `TestEjecucionDelJob` pasa a llamar a `ejecutarElJob`, con las banderas nuevas `-modelo-del-juez`, `-version-del-juez` y `-claude-del-juez` y el votante de T003, cuyo `PATH` lleva delante el directorio del `claude` del juez y cuyo contexto es el de sus señales; sigue fallando con un error o con el veredicto `fallo`. **`evals.sh`**: con una skill que tiene la carpeta `juez` en sus evals, exige `MODELO_DEL_JUEZ`, `VERSION_DE_CLAUDE_CODE_DEL_JUEZ` y un `CLAUDE_DEL_JUEZ` ejecutable antes de la primera sesión, con el mensaje `evals: falta <variable>` de sus vecinas, y los pasa con esas tres banderas; con una skill sin juez no los mira; y su comentario de cabecera las nombra. **`TestEjecucionSinMedir`** (`ejecucion_test.go`): con una copia de la carpeta del juez en `t.TempDir()` cuya medida no corresponde, el informe es el de §5, con el motivo que dice cuál de las cuatro no coincide, 0 llamadas a quien abre las sesiones y 0 al votante, que el test cuenta; con un recuento distinto de 0, lo mismo; y con la medida que corresponde y se cumple, se llama a quien abre las sesiones (control de umbral de FR-104 y SC-004, y de los dos de `medida_del_juez:…` en el job). El guion del job no tiene test en `make ci`: su sintaxis se comprueba con `bash -n`, sin ejecutarlo, y lo ejercita el job del cierre. — FR-043, FR-044, FR-093, FR-104; SC-004. Rutas: internal/evals/ejecucion.go, internal/evals/informe.go, internal/evals/job_test.go, scripts/evals.sh
- [ ] T012 [US4] El punto de entrada de la medida, su guion y su objetivo de `make` (contracts/medida-del-juez.md §7 y §8; contracts/job-de-evals.md §3; research D4, V9; plan.md paso 9). **`job_test.go`**: `TestMedidaDelJuez`, con la etiqueta `evals`, que exige sus banderas (`-skill`, `-modelo-del-juez`, `-version-del-juez`, `-claude-del-juez`, `-concurrencia`, `-commit` y `-salida`), lee el juez de la skill, llama a `medirAlJuez` con el votante de T003 y el contexto de sus señales, escribe en el fichero de `-salida` el texto de la medida si lo hay y falla con el error de `medirAlJuez`, que nombra cada caso; la bandera `-salida`, que ya existe, pasa a describir sus dos usos. **`evals-medir-juez.sh`**, nuevo y ejecutable, con la forma del guion del sondeo: recibe la skill; sin ella, o sin `MODELO_DEL_JUEZ`, `VERSION_DE_CLAUDE_CODE_DEL_JUEZ`, `CONCURRENCIA_DE_EVALS`, `COMMIT_EVALUADO`, un `CLAUDE_DEL_JUEZ` ejecutable o `CLAUDE_CODE_OAUTH_TOKEN`, dice cuál falta y sale con 1 sin ejecutar nada; ejecuta `TestMedidaDelJuez` sin límite de tiempo de `go test`, con un temporal propio que borra al salir; imprime `medida.json` entre `--- inicio de medida.json ---` y `--- fin de medida.json ---` si el test lo escribió, y nada entre marcas si no; y sale con el código del test. No escribe nada en el repositorio (FR-054). **`Makefile`**: el objetivo `evals-medir-juez`, con su línea `##` de ayuda y en `.PHONY`, que ejecuta el guion con `$(SKILL)`, fuera de `ci`. **`TestGuionDeLaMedida`** (`medida_test.go`), con el `go` sustituto de `sustitutos_test.go`, como `TestGuionDelSondeo`: sin la skill o sin cada variable obligatoria, sale con 1 sin ejecutar `go`; con el sustituto que escribe la medida, la imprime entre sus dos marcas y sale con 0; con el que falla sin escribirla, nada entre marcas y el código del test; y con el que la escribe y falla, la medida y ese código. La tarea no ejecuta el objetivo ni el guion a mano: abren sesiones con modelo. — FR-050, FR-052, FR-053, FR-054, FR-093, FR-106. Rutas: scripts/evals-medir-juez.sh, Makefile, internal/evals/job_test.go, internal/evals/medida.go, internal/evals/sustitutos_test.go
- [ ] T013 [US4] La definición del job con el segundo Claude Code, el trabajo `medida`, su entrada y los topes recalculados, comprobada en `make ci` (contracts/job-de-evals.md §2 a §5; data-model §6; research D6, D21, S1, S3, S4, V11, V12; plan.md paso 9). **`evals.yml`**, con el texto de contracts §2 y §3 carácter a carácter: en `on.workflow_dispatch.inputs`, `medir_al_juez`, booleana y `false` por omisión; el paso que instala Claude Code en el trabajo `evals` instala además el del juez en su prefijo (`npm install --prefix "$RUNNER_TEMP/claude-del-juez"` con `VERSION_DE_CLAUDE_CODE_DEL_JUEZ`), pide su versión y deja `CLAUDE_DEL_JUEZ` en `$GITHUB_ENV`; el paso que retira Python añade `$CLAUDE_DEL_JUEZ` a lo que el job usa; la condición de `tanda` pasa de `github.event_name == 'workflow_dispatch'` a `(github.event_name == 'workflow_dispatch' && inputs.medir_al_juez != true)`; `timeout-minutes` de `evals` pasa de 240 a 352, con su comentario; y el trabajo `medida` de §3 —su `name`, su `if` (`github.event.label.name == 'evals-medir-juez' || inputs.medir_al_juez == true`), su matriz con `boe-legislacion` y su concurrencia 4, `timeout-minutes: 269`, `contents: read`, su `env` con el mismo modelo y la misma versión del juez que `evals`, sin `needs`, y sus pasos: el código del commit evaluado, Go, el Claude Code del juez y `make evals-medir-juez SKILL="$SKILL_EVALUADA"` con la credencial; sin `strace`, sin instalar las skills y sin retirar Python—. **`definicion.go`**: `leerDefinicionDelJob` lee además el trabajo `medida` (su nombre, su condición, su `env`, sus skills con su concurrencia, su tope y si tiene `needs`), las entradas del despacho y el `run` del paso de instalación; el peor caso de una skill con juez suma 60 s de la instalación del segundo Claude Code y, por grupo, `⌈respuestas × 6 / concurrencia⌉ × 40 s`, con las respuestas del modelo que decide en las evals que activan la skill por modo y las de la eval sin binario ni servidor (21 097 s con las evals de hoy); el de `medida` es `485 + 60 + ⌈casos × 6 / concurrencia⌉ × 40` con los casos de la copia (16 105 s); y el de una skill sin juez no cambia. Los 40 s son el tope y el margen de T003, no otro número. **`TestDefinicionDelJob`** (`definicion_test.go`): `del-repositorio` pasa con la definición nueva; `sinteticas` gana una definición por cada fila de contracts §5, cada una con su línea —`MODELO_DEL_JUEZ` que falta, que es un alias o que es igual a `MODELO_DE_EVALS`; `VERSION_DE_CLAUDE_CODE_DEL_JUEZ` que falta o no es `<n>.<n>.<n>`, el paso que no instala el del juez en su prefijo o que instala el de las sesiones con ella; `jobs.medida.env` con otro modelo u otra versión; `jobs.medida.if` distinto; `jobs.medida.needs` presente; `tanda` o `evals` que nombran `evals-medir-juez`, o `tanda` que admite el despacho con la entrada; la entrada que falta o con otro valor por omisión; y cada `timeout-minutes` un minuto por debajo de su peor caso—; y `peor-caso` fija los dos peores casos con sus términos (controles de umbral de FR-092, FR-112 y SC-012). La constante de la condición de `tanda` y la definición del contrato del test cambian con ella. El test de la release de la raíz se lee antes, y solo se toca si alguna de sus comprobaciones enumera los trabajos o los pasos de este flujo. — FR-050, FR-090, FR-091, FR-092, FR-112; SC-012. Rutas: .github/workflows/evals.yml, internal/evals/definicion.go, release_test.go

**Checkpoint**: el instrumento está entero: el job juzga con un juez medido, y la medida solo se repite a petición.

---

## Phase 10: User Story 7 — El sondeo juzga con el juez, sin veredicto (P3)

**Objetivo**: quien cambia la skill ve en el sondeo lo que el juez marca, sin que el sondeo falle por ello (plan.md,
paso 10).

**Prueba independiente**: `go test -count=1 -run '^(TestJuicioDelSondeo|TestSalidaDelSondeo|TestSondear)$' ./internal/evals/`.

- [ ] T014 [US7] El sondeo juzga con el juez y lo dice en su salida, sin veredicto (contracts/informe-del-job.md §8 y §9; contracts/medida-del-juez.md §2; research D12; plan.md paso 10). **`sondeo.go`**: con una skill con juez, el sondeo juzga las respuestas de su modelo en las evals que activan la skill, terminadas y medidas, con el mensaje y la regla de T002 y el votante que recibe en `SondeoAEjecutar`, con el modelo del juez de la definición del job (`DefinicionDelJob.ModeloDelJuez`) y la concurrencia del sondeo; `juicioDelSondeo` lleva, por clase, las marcadas o las que tienen sí, sobre las juzgadas, y las sin juzgar con su motivo; la versión de Claude Code del equipo es la que declaran los transcripts de sus sesiones, y con ella y ese modelo se llama a `comprobarLaMedida` solo para decirlo. La salida lleva, donde iba la línea del recuento de expresiones que salió en T005, las líneas de contracts §8 carácter a carácter: la del juez con su modelo; una por clase de `clases.yaml` (`- marcadas en <clase>, con sus tres votos: <n> de <t> (<p> %).` o `- con sí en <clase>, con un voto: <n> de <t> (<p> %).`); la de la medida, que corresponde con el Claude Code del equipo o que no corresponde y por qué; y `Respuestas sin juzgar: ninguna.` o una línea por cada una; si ningún transcript declara la versión, la línea lo dice y el sondeo sigue; y con una skill sin juez, `La skill no tiene juez.`. No lista votos ni frases, no escribe informe, no mide el modo herramienta ni lanza la medida, y sale con 0 sean cuales sean los recuentos y corresponda o no la medida. **`job_test.go`**: `TestSondeo` da a `sondear` el votante de T003, con la ruta absoluta del guion del voto, el `PATH` y la credencial de quien lo lanza y el contexto de sus señales; el guion del sondeo no cambia. **Tests** (`sondeo_test.go`), con el `claude` sustituto de las sesiones y un votante de salidas grabadas: `TestJuicioDelSondeo` —0 marcadas y 2 con sí de 9; 1 marcada; una sin juzgar con su motivo; y una skill sin juez—; `TestSalidaDelSondeo` —las cinco líneas del ejemplo de contracts §8 byte a byte, la variante de la medida que corresponde, la de la versión que ningún transcript declara y la de la skill sin juez—; y `TestSondear` —con una respuesta marcada y con una medida que no corresponde al Claude Code de los transcripts, termina sin error y con sus líneas—. — FR-075, FR-076; SC-004. Rutas: internal/evals/sondeo.go, internal/evals/job_test.go

**Checkpoint**: el sondeo da las marcas del juez y dice si la medida corresponde al Claude Code del equipo, y sale con 0.

---

## Phase 11: User Story 1 — Quien pregunta no lee nada de un precepto que nadie leyó (P1) 🎯

**Objetivo**: `boe-legislacion` v0.1.7, con cada cambio trazado a su causa, y la prosa sin «el sobre» ni los nombres
de campo (plan.md, paso 11). Va detrás del instrumento porque es él quien la mide (Definition of Done §1.10).

**Prueba independiente**: quickstart.md §5.

- [ ] T015 [datos] [US1] `boe-legislacion` v0.1.7 y el vocabulario que su prosa no usa, en la misma tarea porque con la clave nueva la prosa de v0.1.6 da cinco defectos (contracts/skill-boe-legislacion.md §1 a §4 y §7; contracts/skill-boe-legislacion-v0.1.7.diff; data-model §5 y §7; research D22, D23, V13, V14, V18, V19; plan.md paso 11, «Cambios de SKILL.md trazados a la causa» y *Complexity Tracking*, segunda fila). **El esquema** schemas/expresiones-prohibidas.yaml.json exige la clave nueva `salida_de_las_herramientas`: una lista de al menos un elemento, cada uno con el patrón de contracts/skill-boe-legislacion.md §4, que admite el guion bajo de un nombre de campo; las demás claves no cambian. **La lista** evals/boe-legislacion/expresiones-prohibidas.yaml gana esa clave con `el sobre`, `fecha_vigencia` y `norma_modificadora`, y su comentario de cabecera dice lo que el fichero es desde H24: el vocabulario que la prosa de `SKILL.md` no usa, que ya no juzga ninguna respuesta. **El tipo** (`prohibidas.go`): `ExpresionesProhibidas.SalidaDeLasHerramientas`, que el lector estricto necesita y que no entra en las expresiones que `ExtraerExpresionesProhibidas` busca en un texto: las 36, 11 y 9 respuestas que `TestJuzgarSinLaLista` exige como premisa siguen siendo esas. **La prosa** (`conjunto_test.go`): las tres expresiones se buscan en toda la prosa de `SKILL.md`, la de `parrafosDeLaProsa` —fuera de los bloques delimitados y de la región generada—, y para esta familia el código en línea cuenta: se le quitan los acentos graves y se deja su texto; las demás familias siguen sin mirarlo; la comparación es la de `formaDeExpresion`, por palabras y sin distinguir mayúsculas. **`SKILL.md`** skills/boe-legislacion/SKILL.md: el prototipo aplicado con `git apply` (19 líneas fuera y 19 dentro), que trae C1 a C8 de contracts §2, cada uno trazado a su causa (§1): C1, la viñeta «De un precepto que no has leído, nada.», con su razón (FR-081 a FR-083); C2, «por su número y sin decir de qué tratan» en la regla 2; C3 y C4, la vigencia con palabras de quien lee, sin «lo que trae el sobre» ni las dos glosas de campo (FR-084); C5 a C7, el vocabulario fuera del paso 4, de «Cómo se cita» y de la línea de la redacción modificada; y C8, la línea que paga C1. Queda en 298 líneas; lo demás de v0.1.6 no cambia (contracts §3), y no nombra evals, el job ni modelos. **Tests**: los del formato del esquema de la lista (`TestEsquemaDeExpresionesProhibidas`, en `formato_test.go`) y los de su lectura ganan la clave —sin ella, vacía, o con un elemento con acento grave o con asterisco, no valida; con `fecha_vigencia`, sí—; toda lista sintética de los `_test.go` del paquete (`grep` de `otra_conversacion`) gana la clave; `TestProsaDeLaSkill` gana los casos de contracts §4 con Markdown escrito en el test —los dos párrafos de v0.1.6 que señala FR-085, tal cual, dan su defecto con sus expresiones; un campo en código en línea cuenta y `graph check` en código en línea sigue sin contar; en un bloque delimitado y en la región generada no cuenta ninguna; «del sobre» no es «el sobre», y `fecha_vigencia_reciente` sí lleva `fecha_vigencia`—; y `prosa-de-la-skill` de `TestEvalsDelRepositorio` aplica lo mismo al `SKILL.md` del repositorio: 0 usos (controles de umbral de FR-086, FR-110 y SC-010, con `skills-check`, que falla con 300 líneas o más). Orden dentro de la tarea, para verlo en rojo: primero la clave, el tipo, la prosa y los casos, con los que `prosa-de-la-skill` da los cinco párrafos de v0.1.6 (líneas 111, 113, 139, 194 y 215); después el diff, con el que da cero. Antes de terminar, `make skills-check` y `wc -l` del fichero. Si `git apply` no entrara limpio, se aplican a mano los mismos C1 a C8 con el texto del diff, sin pasar de 298 líneas. — FR-080, FR-081, FR-082, FR-083, FR-084, FR-085, FR-086, FR-110; SC-010. Rutas: schemas/expresiones-prohibidas.yaml.json, evals/boe-legislacion/expresiones-prohibidas.yaml, internal/evals/prohibidas.go, internal/evals/conjunto_test.go, internal/evals/formato_test.go, internal/evals/informe_test.go, internal/evals/juzgar_test.go, internal/evals/sondeo_test.go, internal/evals/umbrales_test.go, skills/boe-legislacion/SKILL.md

**Checkpoint**: la skill ya no enseña a glosar lo que no se ha leído; si lo consigue lo mide el job del cierre
(`afirma_lo_no_leido:<modelo>:<modo>` con 0 en los dos modos, SC-001).

---

## Phase 12: Cierre — documentación y Definition of Done

- [ ] T016 Documentación que el hito deja falsa (contracts/skill-boe-legislacion.md §6; contracts/informe-del-job.md; contracts/medida-del-juez.md; contracts/job-de-evals.md; plan.md paso 12). CHANGELOG.md, *Unreleased*: `boe-legislacion` v0.1.7 con lo que cambia para quien la usa —no dice qué dice ni de qué trata un precepto que no ha leído, traslada las remisiones como el texto las da y dice la vigencia sin «el sobre» ni nombres de campos (FR-087)—, y lo del job: el juez con modelo y sus dos clases, la regla de los tres votos, los umbrales nuevos y los dos que salen, la clave `juez` del informe, la medida versionada y qué hace fallar `make ci`, `make evals-medir-juez` con su etiqueta y su entrada, el modelo del juez y su versión de Claude Code, lo que da ahora el sondeo y la lista como vocabulario de la prosa (FR-095, FR-113). CONTRIBUTING.md, en «Skills y evals» («Formato común de eval», «Job de evals» y «Sondeo local»): el juez y sus clases en la carpeta `juez` de las evals de cada skill; qué recibe un voto y qué no; la frase comprobada sin modelo, el voto nulo y la regla; qué respuestas se juzgan; los doce umbrales, con los dos que salen; `juez` y `sin_juzgar` en el informe y sus motivos; la medida versionada, sus cuatro claves, qué hace fallar `make ci` (`TestMedidaVersionada`, `TestCopiasDelJuez`) y qué hace el job con el instrumento sin medir; cómo se lanza la medida (la etiqueta `evals-medir-juez` o la entrada `medir_al_juez` del flujo lanzado a mano) y que la lanza y la versiona una persona, fuera de un run, en el mismo cambio que toca la rúbrica, los casos, el modelo del juez o su versión; `MODELO_DEL_JUEZ` y `VERSION_DE_CLAUDE_CODE_DEL_JUEZ`, aparte de las de las sesiones, y los dos topes; las líneas del juez en el sondeo; y que la lista es solo el vocabulario de la prosa, con su clave nueva. docs/WORKFLOW.md: lo que dice del job de cierre y de sus umbrales —los que deciden ahora, que el juez corre en el job y nunca en un paso de un run, y que la medida no la lanza el run, que solo pone la etiqueta `evals`—. internal/evals/doc.go: el comentario del paquete con `juez.go`, `medida.go` y `ejecucion.go`, la lista como vocabulario de la prosa, los umbrales de hoy y el sondeo con el juez. .github/workflows/evals.yml: solo comentarios —la cabecera y los de los trabajos dicen lo que el flujo hace desde H24 y remiten a los contratos de este hito—. Verificación: cada fichero, test, clave, variable, objetivo y etiqueta que se nombra existe en el árbol (`grep` uno a uno); ningún texto sigue diciendo como vigente que la lista decide o que `umbrales` lleva `expresiones_prohibidas` o `redaccion_no_leida`; y `make ci`, que con `TestDefinicionDelJob` ve que los comentarios no han cambiado ninguna clave del flujo. No se editan los ADR, el roadmap, la bitácora de uso ni los specs de otros hitos. — FR-087, FR-095, FR-096, FR-113; SC-013. Rutas: CHANGELOG.md, CONTRIBUTING.md, docs/WORKFLOW.md, internal/evals/doc.go, .github/workflows/evals.yml
- [ ] T017 Cierre de la Definition of Done, sin código de producto (`ROADMAP.md` §1; plan.md, «Obligaciones para tasks.md»): specs/017-h24-las-evals-juzgan/cierre.md con (1) la lista, sacada de `git diff --name-status main`, de los esquemas, la carpeta del juez, los dos ficheros restaurados, la lista y la skill creados o modificados, y la comprobación de que no cambia ninguna grabación, ninguna fuente, ningún manifiesto, `SOURCES.md`, ningún ADR, ningún spec, plan o informe de H7.1 a H22, el directorio de los guiones del workflow ni la carpeta de la evidencia del ADR 0037 (FR-045, FR-096), y de que las cuatro copias siguen idénticas a sus originales; (2) la cobertura sobre `coverage.out` y `coverage-integration.out` de `make ci` con `go tool cover -func` —global ≥ 70 % y el dominio ≥ 85 % (Definition of Done §1.9; el hito no toca el dominio)— y la de `juez.go`, `medida.go` y `ejecucion.go`; (3) el resultado de quickstart.md §1 a §6, orden a orden, con lo esperado de cada una, y la constancia de que §7 y §9 no se ejecutan en el run y de que §8 lo hace el workflow; (4) la tabla de plan.md, «Controles de umbral», con el test de cada fila `ci:` encontrado como `^func <Test>\(` en su ruta, que es como lo busca el informe final; y (5) la constancia de que ningún umbral se rebajó, pasó a `decide: false` ni perdió respuestas o casos de su total (FR-036), y de que el árbol no lleva ningún `//nolint`, `t.Skip` ni TODO nuevos. Solo si una cifra de cobertura queda bajo su umbral se añaden los tests que faltan, en los `_test.go` de esos tres ficheros. — FR-036, FR-045, FR-096, FR-100, FR-101; SC-013. Rutas: specs/017-h24-las-evals-juzgan/cierre.md, internal/evals/juez.go, internal/evals/medida.go, internal/evals/ejecucion.go

---

## Dependencias y orden de ejecución

El orden es el de plan.md, «Orden de implementación (de dentro afuera)», y es estrictamente secuencial: T001 → T017.
Ninguna tarea depende de una posterior.

| Tarea | Paso del plan | Necesita |
|---|---|---|
| T001 | 1 | — |
| T002 | 2 | T001 (`Juez`, sus clases y su esquema) |
| T003 | 3 | T002 (`Votante`, el error del tope) |
| T004 | 4 | T001 (la carpeta del juez y su medida) |
| T005 | 5 | T001 (`Conjunto.Juez`, de quien pasan a depender los umbrales); no necesita T008: las dos respuestas del calibrado cuya eval retiró H7.2 se juzgan sin su fichero |
| T006 | 6 | T002 (mensaje y regla), T004 (la medida leída), T005 (`umbrales` sin la lista) |
| T007 | 6 | T006 (los juicios de cada respuesta) |
| T008 | 7 | — (va aquí porque solo lo usa T009) |
| T009 | 8 | T001 (los casos), T002 (`Texto`), T008 (la eval retirada) |
| T010 | 8 | T002 (la regla), T009 (los casos resueltos) |
| T011 | 9 | T003 (el votante que ejecuta el guion), T004 (`comprobarLaMedida`), T006 y T007 (el informe con el juez) |
| T012 | 9 | T003, T010 (`medirAlJuez`), T011 (las banderas del juez) |
| T013 | 9 | T003 (el tope de un voto), T004 (las dos variables), T009 (los casos), T012 (el objetivo que el trabajo `medida` ejecuta) |
| T014 | 10 | T003, T004, T005 (el sondeo sin el recuento), T006 |
| T015 | 11 | T005 (la lista ya solo es de la prosa); va detrás de T001 a T014 porque el instrumento precede a la skill |
| T016 | 12 | T001 a T015 |
| T017 | — | T001 a T016 |

### Historias

- **US3** (T001, T004, T011) y **US2** (T002, T003, T006) son el instrumento que ve el defecto; **US5** (T007) lo hace
  legible; **US6** (T005) retira el control que fallaba por los dos lados; **US4** (T008 a T010, T012, T013) deja que
  una persona repita la medida; **US7** (T014) lo lleva al sondeo; y **US1** (T015) arregla la skill.
- Las historias no son independientes en su construcción —comparten `internal/evals`— y sí en su prueba: cada fase
  nombra la orden que la comprueba sola.

### Oportunidades de paralelismo

Ninguna dentro del run: el workflow ejecuta las tareas una a una, y casi todas tocan los mismos ficheros de
`internal/evals`. Ninguna tarea lleva `[P]`. Fuera de un run, T008 no depende de ninguna anterior, y T015 solo de T005.

---

## Trazabilidad

### Requisitos → tareas

| Requisito | Tareas |
|---|---|
| FR-001, FR-002 | T002 |
| FR-003 | T003 |
| FR-004 | T002 (mensaje y esquema), T003 (orden) |
| FR-005, FR-006 | T002 |
| FR-007 | T002 (motivos), T003 (tope), T006 (veredicto y motivo), T007 (`sin_juzgar`) |
| FR-010, FR-011 | T002 |
| FR-012, FR-014 | T006 |
| FR-013 | T006 (no cuenta ni decide), T007 (se publica) |
| FR-020, FR-021, FR-022, FR-023 | T001 |
| FR-024 | T008 (los dos ficheros restaurados), T009 (casos y reconstrucción) |
| FR-030, FR-031, FR-033, FR-035, FR-037 | T006 |
| FR-032 | T004 (`make ci`), T006 (los dos umbrales), T011 (el job sin medir) |
| FR-034 | T005 (salen dos y siguen cuatro), T006 (los doce) |
| FR-036 | Restricción de todas las tareas (batería); T017 la deja constatada |
| FR-040 | T001 (la copia; ninguna tarea la repite ni la escribe) |
| FR-041, FR-042 | T004 |
| FR-043 | T011 (el job), T010 (la ejecución de la medida no la hace) |
| FR-044 | T006, T011 |
| FR-045 | Restricción de todas las tareas (batería); T004 fija el modelo y la versión; T017 la deja constatada |
| FR-050 | T012 (objetivo y guion), T013 (etiqueta y entrada) |
| FR-051 | T009 (reconstrucción), T010 |
| FR-052, FR-053 | T010, T012 |
| FR-054 | T010, T012 (nada escribe la medida en el repositorio) |
| FR-060, FR-061 | T007 |
| FR-062, FR-070, FR-071 | T005 |
| FR-075, FR-076 | T014 |
| FR-080 | research y plan.md (la causa y la traza, ya escritas); T015 aplica C1 a C8 |
| FR-081, FR-082, FR-083, FR-084 | T015 (lo mide el cierre con los umbrales de T006) |
| FR-085, FR-086 | T015 |
| FR-087 | T016 |
| FR-090, FR-091 | T004 (se fijan), T013 (se comprueban) |
| FR-092 | T013 |
| FR-093 | Restricción de todas las tareas (batería); T003, T011 y T012 la nombran |
| FR-095 | T016 |
| FR-096 | Restricción de todas las tareas (batería); T016 y T017 |
| FR-100 | Todas (`make ci` por tarea); T017 |
| FR-101 | plan.md, «Controles de umbral»; tabla siguiente; T017 |
| FR-102, FR-103 | T002 |
| FR-104 | T006, T007, T011 |
| FR-105 | T004 |
| FR-106 | T010, T012 |
| FR-107 | T002, T003 |
| FR-108 | T001 |
| FR-109 | T008, T009 |
| FR-110 | T015 |
| FR-111 | T005 |
| FR-112 | T013 |
| FR-113 | T016; el job de evals lo lanza el workflow en el cierre |
| SC-001 | El cierre del workflow, con los controles de T004, T005, T006 y T011 y la skill de T015; no es una tarea |
| SC-002, SC-003 | T002 |
| SC-004 | T006, T007, T011, T014 |
| SC-005 | T004 |
| SC-006 | T010 |
| SC-007 | T002, T003 |
| SC-008 | T001 |
| SC-009 | T008, T009 |
| SC-010 | T015 |
| SC-011 | T005 |
| SC-012 | T013 |
| SC-013 | T016, T017 |
| SC-014 | Una persona, después del run; no es una tarea |

### Controles de umbral (plan.md) → tarea que construye el control

| Fila de plan.md | Control | Tarea | Test que lo ve fallar |
|---|---|---|---|
| FR-030, SC-001 (los dos modos) | `afirma_lo_no_leido:claude-sonnet-5-5:<modo>` | T006 | `TestInformeConElJuez`: 1 marcada en un modo da `cumple: false` y `fallo` |
| FR-032, SC-001 (defectos y correctos) | `medida_del_juez:afirma_lo_no_leido:…` | T006, con T004 y T011 | `TestMedidaVersionada` (recuento distinto de 0), `TestEjecucionSinMedir` (el job, `fallo` sin sesiones), `TestUmbralesDelInforme` (los dos elementos) |
| FR-033, SC-001 (los dos modos) | `duracion_del_juez:<modo>` | T006 | `TestInformeConElJuez`: 901 s da `fallo`, 900 se cumple |
| FR-034, SC-001 (`sin_activar`, los dos modos) | El de hoy, que sigue | T005 | Los casos de `TestUmbralesDelInforme` y de `TestInformeEnDosModos` con una respuesta sin activar |
| FR-034, SC-001 (sesiones, los dos modos) | El de hoy, que sigue | T005 | Los casos de la duración de `TestUmbralesDelInforme` y de `TestInformeEnDosModos` |
| FR-007, SC-001 (sin juzgar) | `ci:internal/evals/informe_test.go:TestInformeConElJuez` | T006 | Un voto que no llega: `fallo` con el motivo de la ejecución |
| FR-102, SC-002 | `ci:internal/evals/juez_test.go:TestVotoDelJuez` | T002 | Los 4 casos, con el nulo repetido exactamente 1 vez |
| FR-103, SC-003 | `ci:internal/evals/juez_test.go:TestReglaDeLosVotos` | T002 | Las filas de la regla con sus votos pedidos |
| FR-104, SC-004 | `ci:internal/evals/informe_test.go:TestInformeConElJuez`, `ci:internal/evals/ejecucion_test.go:TestEjecucionSinMedir` | T006, T007, T011 | Los casos de FR-104, con quien abre las sesiones contado |
| FR-042, FR-105, SC-005 | `ci:internal/evals/medida_test.go:TestMedidaVersionada` | T004 | 6 de 6 mutaciones |
| FR-106, SC-006 | `ci:internal/evals/medida_test.go:TestEjecucionDeLaMedida` | T010 | 1 defecto sin marcar, 1 correcto marcado, y los 259 bien con una medida que no corresponde |
| FR-107, SC-007 | `ci:internal/evals/juez_test.go:TestMensajeDelVoto`, `ci:internal/evals/juez_test.go:TestOrdenDelVoto` | T002, T003 | El mensaje byte a byte; los argumentos y el entorno del guion |
| FR-023, FR-108, SC-008 | `ci:internal/evals/medida_test.go:TestCopiasDelJuez` | T001 | Un byte cambiado en cada una de las cuatro |
| FR-109, SC-009 | `ci:internal/evals/medida_test.go:TestGrabacionesDerivadas` | T009, con T008 | Los 259 casos y los 140 derivados |
| FR-086, FR-110, SC-010 | `ci:Makefile:skills-check`, `ci:internal/evals/conjunto_test.go:TestProsaDeLaSkill`, `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` | T015 | Los dos párrafos de v0.1.6, y la prosa de v0.1.6 antes del diff |
| FR-111, SC-011 | `ci:internal/evals/juzgar_test.go:TestJuzgarSinLaLista` | T005 | Las dos respuestas del 2026-10-04 y las 56 del calibrado de H7.4 (36, 11 y 9, exigidas como premisa); falla con el árbol de antes de la tarea |
| FR-092, FR-112, SC-012 | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` | T013 | Las definiciones sintéticas, con cada tope un minuto por debajo |

### Definition of Done (`ROADMAP.md` §1) → tareas

| Punto | Aplica | Tareas |
|---|---|---|
| 1 · `make ci` en verde | Sí | Todas; T017 |
| 2 · Tests offline; fixtures si toca red | Sí; no toca red | T001 a T015; ningún fixture grabado |
| 3 · Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | Sí | T002 a T014 (el voto es un proceso; la medida se escribe en su fichero) |
| 4 · `schemas/` y `schema-check` | Los dos esquemas que se tocan no salen de `--describe` | T001, T015 |
| 5 · Errores tipados y códigos estables | Ningún código de `kitlegal` cambia | — |
| 6 · e2e y `CHANGELOG.md` | «Aceptación e2e: no aplica»; `CHANGELOG.md`, sí | T016 |
| 7 · ADR | No: la decisión es el ADR 0037, ya aceptado (FR-096) | — |
| 8 · `SOURCES.md` | No: ninguna fuente nueva ni cambiada | — |
| 9 · Cobertura | Sí | T017 |
| 10 · Skill: evals antes, `SKILL.md` < 300, sin drift | Sí; sin evals nuevas (fuera de alcance): el instrumento va antes | T001 a T014, T015 |
| 11 · Dimensión territorial | No | — |
| 12 · Operaciones de grafo | No: ningún applet cambia | — |
| 13 · Recursos, plazos o escritos | No | — |

---

## Estrategia de implementación

- **De dentro afuera** (plan.md): los datos del juez, lo que decide un voto, el proceso que lo abre, la comprobación de
  la medida, la lista fuera, el juez en el informe, los casos y su medida, el job, el sondeo y, al final, la skill.
- **Lo mínimo que ya da valor** son US2 y US3 con el job (T001 a T007 y T011): un job que marca por lo que la respuesta
  dice y que no decide sin estar medido. No es una entrega parcial: el run ejecuta las diecisiete, y SC-001 pide la
  skill corregida.
- **Lo que el run no puede medir** (research S1 a S8): que v0.1.7 deja `afirma_lo_no_leido` en 0, cuánto tarda un voto,
  que las banderas del voto son las de Claude Code 2.1.289 y que `npm install --prefix` deja el ejecutable donde el
  flujo lo busca. Lo mide el job del cierre; si una respuesta queda marcada, la reparación corrige `SKILL.md` con la
  frase marcada delante y no toca la rúbrica, los casos, la medida ni el umbral (FR-036).
- **Después del run** (SC-014): una persona lanza una vez la medida con el código nuevo y lee las respuestas con algún
  voto afirmativo.

---

## Comprobación contra la rúbrica de `juez_tasks` y `precheck.sh tasks`

- `precheck.sh tasks`: diecisiete líneas con el formato `- [ ] Tnnn`, ids únicos y correlativos; todas con rutas; las
  tres que tocan `schemas/` o `testdata/` llevan `[datos]`, y ninguna otra línea nombra esos directorios; ninguna nombra
  la carpeta de la evidencia por su ruta, ni la variable de grabación, ni una acción de plataforma, ni pide nada a una
  persona; ninguna tarea de aceptación, como pide «Aceptación e2e: no aplica».
- a · `analisis_critico`: lo escribe el paso de análisis después de este fichero.
- b · `trazabilidad`: cada tarea cita sus FR y SC; «Requisitos → tareas» cubre FR-001 a FR-113 y SC-001 a SC-014; y
  «Controles de umbral» da la tarea y el test de cada fila de plan.md.
- c · `rebanadas_verdes`: cada tarea trae su test y su implementación; las excepciones están declaradas arriba; el orden
  es el del plan y la tabla de dependencias no tiene ninguna hacia delante.
- d · `rutas_declaradas`: tras «Rutas:», ficheros concretos o las dos carpetas nuevas de datos; ninguna ruta genérica.
- e · `datos_separados`: T001, T008 y T015, con lo que el dato obliga a cambiar a la vez (plan.md, *Complexity
  Tracking*); ningún dato externo, ninguna grabación.
- f · `dod`: tabla «Definition of Done → tareas»; `CHANGELOG.md` en T016; sin ADR ni `SOURCES.md`, con su motivo.
- g · `checklist_veraz`: `checklists/requirements.md` no cambia.
- h · `aceptacion_primero`: el plan justifica «no aplica».
- i · `autonomia`: ninguna tarea exige a una persona, publica, ni mide en la plataforma.
