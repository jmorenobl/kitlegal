# Tasks: H23 · skill `jurisprudencia`: ninguna sentencia citada sin el documento que trae la persona + `cita preparar` y `cita cotejar`

**Input**: `specs/019-h23-skill-jurisprudencia-ninguna/` (spec.md, plan.md, research.md, data-model.md, contracts/, quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` 2.13.0 aplicada (umbral de materialidad,
criterio de uso, proporcionalidad y convergencia en «Gates»; controles de umbral del ADR 0029; ADR 0030, 0032, 0035,
0036 con su enmienda del 2026-10-07, y 0037); H0-H7.4, H19, H21, H22 y H24 en `main` (el kernel con `--describe`, el
servidor MCP con una herramienta por verbo del registro, `internal/evals` con el formato común, `umbrales` con el
contrato del ADR 0029 y el job en dos modos; el modelo que decide, `claude-sonnet-5-5`, ADR 0031), y la evidencia del
ADR 0036 versionada, con el fragmento. **Ninguna tarea crea el esqueleto ni los gates: `make ci` existe y pasa**, y
cada tarea lo deja en verde al terminar.

**El fragmento.** Es `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt`: el encabezamiento y el fallo de la STS
1088/2023, de 4 de julio, 39 líneas y 2 353 bytes, con SHA-256
`4886e0c8ca9836527ec08d8732b50315640a76803033374af5287b2a8321ee27` (el de `evidencias/adr-0036/manifiesto.json`). Su
**ficha** son sus once primeras líneas, 316 bytes. Ninguna tarea lo escribe ni lo cambia, y ninguna escribe de memoria
el texto de una sentencia: los tests lo leen de su fichero, un guion e2e lo toma por
`$KITLEGAL_SKILLS/../evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt`, y donde hace falta dentro de otro
fichero —las preguntas de dos evals, la ficha de una llamada de herramienta— se lleva con una orden (`sed`, `head`),
nunca tecleado. Las líneas de las tareas lo nombran «el fragmento» y no por su ruta, porque el extractor de rutas del
workflow declara toda ruta que aparece en una línea y `precheck.sh` rechaza la que nombra esa carpeta.

**Aceptación**: cuatro guiones `testscript` (plan.md, «Aceptación e2e»), que escribe T001 en
`specs/019-h23-skill-jurisprudencia-ninguna/aceptacion/` y quedan congelados. El workflow comprueba que cada uno falla
hoy por una aserción y, tras el bucle de tareas, los activa con el prefijo `h23-` en el directorio de guiones de e2e,
donde `make ci` los ejecuta dentro de `TestEntregaDelHito`. T001 **no escribe evals**: las seis de `jurisprudencia`
llevan claves que el formato de eval no tiene hasta T007, y `make ci` lee las evals de toda carpeta de `evals/`, así
que antes lo dejarían en rojo (plan.md, «Aceptación e2e», último párrafo; research V34, D23; precedente, H21 T001).
Entran con T007, antes que el texto de la skill (T009; Definition of Done §1.10). US3 y US5 no son comportamiento del
binario y no tienen guion: las fijan en `make ci` los tests con sesiones sintéticas de T007 y T008 y, en el cierre, el
job de evals (SC-001), que lanza el workflow y no una tarea. US6 la comprueba la revisión final (SC-010).

**Tests**: obligatorios (constitución §III; spec FR-080 a FR-088; plan.md, «Controles mecánicos»). Cada tarea de
código trae su test y la implementación mínima que lo hace pasar, en el mismo diff; los nombres de test y sus ficheros
son los de plan.md, porque el informe final busca cada control `ci:<ruta>:<Test>` como `^func <Test>\(` en esa ruta.

**Rebanadas verticales y excepciones declaradas.** Son rebanadas completas —su test y su implementación en el mismo
diff, `make ci` en verde por sí solas— T004, T008 y T010. Las demás, cada una por su razón y con su verificación:

- **T001** es la suite de aceptación: solo guiones, que hoy fallan por su precondición. Se valida contra un prototipo
  de un solo uso fuera del repositorio, porque un guion congelado que no describe bien la entrega no se puede corregir
  después.
- **T002, T003, T005, T006, T007 y T009** son las de datos: tocan `schemas/` o un `testdata/`, y llevan además solo lo
  que el dato obliga a cambiar a la vez (plan.md, *Complexity Tracking*, cuarta fila, y «Obligaciones para tasks.md»;
  precedente, H7.2 a H24). T002 y T003: el corpus de cada fuzz con el código que lo ejecuta —el corpus sin su `Fuzz…`
  no lo lee nadie—. T005: `schemas/instalacion.json` cambia en el mismo instante en que `--describe` anota las
  banderas, o `schema-check` queda en rojo. T006: en cuanto el applet está en el registro, `schema-check` exige
  `schemas/cita.json`, las listas literales de applets y de herramientas de cuatro guiones cambian y los golden fijan
  lo que el verbo da. T007: el esquema de eval, el tipo que lo lee —el lector estricto rechaza una clave que el tipo
  no tiene— y las seis evals que lo usan. T009: la tercera skill empotrada cambia la lista literal de diecinueve
  guiones.
- **T011** es documentación: su verificación es que cada orden, fichero, test y clave que nombra existe, y `make ci`.
- **T012** es el cierre de la Definition of Done, sin código de producto: escribe la lista de lo tocado, mide la
  cobertura y ejecuta quickstart.md §0 a §9; solo añade tests si una cifra de cobertura queda bajo su umbral.

**Diecinueve guiones, y no diecisiete, por la skill (medido en esta sesión).** plan.md cuenta los guiones que rompe la
tercera skill con `go test ./...` sobre su prototipo, sin la etiqueta `integration`, y da diecisiete, todos del
directorio de guiones de e2e. `make ci` ejecuta además `test-integration`. En ese mismo prototipo, que sigue en
`$TMPDIR/kl-proto`, `go test -count=1 -tags=integration ./...` da en rojo, además de lo que lista research M4,
`TestInstalacion/instalar` y `TestInstalacion/instalar-sin-gobin`: los dos guiones de `internal/skills/testdata/script/`
exigen `\Aboe-legislacion\nkitlegal\.json\nlegal-core\n\z` y `make install` instala cada skill del directorio.
`instalar-de-nuevo` sigue en verde. T009 declara los dos, y los guiones que el hito alinea son veintitrés (cuatro por
el applet y diecinueve por la skill). Anotado en `gates/supuestos.md`.

**Modo**: desatendido (ADR 0018). El workflow `hito` ejecuta y verifica las tareas **una a una**, con guardián de
diff: cada tarea declara en su propia línea, tras «Rutas:», **todas** las rutas que crea, modifica o retira, y solo
toca esas (más `go.mod`, `go.sum`, `CHANGELOG.md`, el directorio del feature y el `x_test.go` de cada `x.go`
declarado). Los demás nombres que aparecen en una línea son contexto: referencias a los documentos del feature
(contracts/, plan.md…) y nombres de ficheros que la tarea solo lee, que se nombran sin su directorio. El cierre en la
plataforma —publicar la rama, abrir la propuesta de cambio y medir CI y el job de evals de SC-001— lo hace el workflow
tras la revisión final; quickstart.md §10 queda para el workflow y §11 para una persona.

## Formato: `- [ ] Tnnn [etiqueta?] [Story?] Descripción — FR/SC. Rutas: …`

- **[aceptacion]**: solo T001.
- **[datos]**: la tarea crea o modifica ficheros bajo `schemas/` o bajo un `testdata/`. Son seis: T002 y T003 (el
  corpus de los tres fuzz), T005 (`schemas/instalacion.json`), T006 (`schemas/cita.json`, los once golden y cuatro
  guiones e2e), T007 (`schemas/eval.yaml.json`) y T009 (diecinueve guiones). H23 no graba nada de ninguna fuente
  (plan.md, «Datos externos»; research D29): no hay manifiesto `grabaciones.json` nuevo, ni test de grabación, ni nada
  que grabar para el paso `grabar_datos`. Los golden los escribe el verbo, no una mano; los corpus de fuzz son
  semillas sintéticas; los esquemas se regeneran desde `--describe` o se editan como contrato del formato de eval; y
  ningún fichero de `data/` cambia. Las sesiones y los sobres de los tests del juicio y del informe son constantes del
  `_test.go` o ficheros escritos en `t.TempDir()`.
- **[P]**: ninguna tarea lo lleva; en el bucle del workflow todo va en secuencia.
- **[Story]**: US1 saber qué buscar y dónde · US2 saber si el documento traído es el pedido · US3 una respuesta solo
  cita la sentencia cuyo documento tiene delante · US4 el applet no toca la red, la caché ni el grafo, y llega como
  herramienta · US5 el job decide con un hecho de la sesión · US6 el README dice qué hace kitlegal con las sentencias.
  La aceptación, la fundación y el cierre no llevan historia. T006 lleva `[US1]` y entrega también US2 y US4: los dos
  verbos y sus herramientas entran con el applet en una sola tarea, porque registrar el applet a medias obligaría a
  regenerar dos veces su esquema y las listas de los guiones.

## Batería de verificación por tarea

- **`make ci` en primer plano** y la tarea marcada `[X]` en el mismo turno en que termina en verde: formato · lint
  (`depguard`, `forbidigo`, `gosec`, `misspell`, `revive`, `dupl`, `gocyclo`, `paralleltest`…, también sobre los
  ficheros con las etiquetas `evals` e `integration`) · tests con `-race` · `test-integration` · `test-tiempos` ·
  `vuln` · `schema-check` · `skills-check` · `goreleaser-check` · secretos · módulos. Ningún registro de la
  verificación en la raíz del repositorio. Las salidas de `go test` y de `make` se leen con `rtk proxy` o con sondas
  positivas, porque el proxy de la sesión las resume.
- **Sin red ni modelo**: ninguna tarea usa la red salvo la de las herramientas de Go (`make vuln`); ninguna ejecuta
  `claude`, `make evals`, `make evals-sondeo`, `make evals-medir-juez`, los guiones `evals*.sh` de `scripts`,
  `TestEjecucionDelJob`, `TestMedidaDelJuez` ni `TestSondeo`. El applet `cita` no pide nada a ninguna fuente, y ningún
  test ni guion consulta el CENDOJ ni comprueba que una dirección responde.
- **Sin personas a mitad**: ninguna tarea la hace, la revisa ni la desbloquea una persona.
- **Los umbrales no se rebajan** (FR-064, ADR 0029): ninguna tarea ni corrección cumple un umbral de FR-060 o de
  FR-062 rebajándolo, dejándolo en `decide: false`, sacando evals o un modo de su total, sumando los modos ni
  cambiando lo que cuenta como salida de una operación. Cada control nuevo se ve en rojo antes de darlo por bueno: con
  sus casos negativos o con un mutante temporal de lo que fija, que no queda en el diff.
- **Sin juez** (FR-055): ninguna tarea crea la carpeta `juez` de las evals de `jurisprudencia`, declara una clase de
  juez para ella ni apunta nada a la evidencia del ADR 0037.
- **Ningún test escribe fuera de `t.TempDir()`**; lo temporal de una tarea va bajo `$TMPDIR`, nunca bajo `/tmp`, y
  cualquier copia o worktree del repositorio, fuera del árbol.
- **Sin atajos** (plan.md, «Comprobación contra la rúbrica», g): ningún `//nolint`, ningún `t.Skip`, TODO ni error
  silenciado; ningún test se desactiva. Si `dupl` marca el esqueleto repetido de los dos verbos o `gocyclo` la lectura
  de la ficha (research S7), se parte la función; una palabra española que `misspell` tome por errata va a
  `ignore-rules`, que el guardián admite sin declararlo.
- **Antes de añadir un valor a una lista o un campo a un tipo que construyen los tests**, `grep` de sus usos en todos
  los `_test.go` del paquete, con la etiqueta `integration` incluida, y en los guiones: lo que haya que alinear entra
  en la misma tarea, y sus ficheros están en sus rutas.
- **Proporcionalidad** (constitución, «Gates»; plan.md, «Trazabilidad», último párrafo): ninguna tarea añade un caso,
  un mensaje ni un test para un estado sin vía real —un documento falso, una caché o un `world.db` tocados a mano, una
  entrada estándar que no se puede leer siguen la regla genérica, código 1—.
- **Lo que no cambia** (plan.md, «Constraints» y «Obligaciones para tasks.md»; research D30): `mcp.go`,
  `herramientas.go` y el paquete del servidor MCP; el paso de empaquetado y su paquete (FR-031); `internal/httpx`; los
  guiones del workflow `hito`; la carpeta de evidencias; las skills `boe-legislacion` y `legal-core`, byte a byte
  (FR-049), y sus evals; `SOURCES.md`, `CLAUDE.md`, el roadmap, la bitácora de uso y la web; ningún ADR nuevo; los
  specs, los planes y las suites congeladas de los hitos anteriores, de las que solo cambian las listas literales que
  T006 y T009 nombran. Ningún `os.Exit`, `fmt.Print*`, `os.Stdout`, `os.Stderr` ni `net/http` nuevo en Go (R2, R4, R5).
- **Vocabulario**: los comentarios llevan tilde; los caracteres que no son ASCII de los tests se escriben con su escape
  de bytes de Go (`\xe2\x9a\xa0` para `⚠`…), no con `\u`.
- **Sondas, mutantes y copias momentáneas**: se crean y se retiran dentro de la tarea, antes de `make ci` y del commit.
  Un guion congelado se ejerce copiándolo un momento al directorio de guiones de e2e con el prefijo `zz-`, como hace el
  «rojo primero», y la copia se borra.

---

## Fase 1: Aceptación — la suite congelada de la entrega

**Objetivo**: los cuatro guiones que describen los dos verbos vistos desde la terminal y desde un cliente MCP, en rojo
por una aserción.

**Prueba independiente**: el «rojo primero» del workflow sobre cada guion; después del bucle, `TestEntregaDelHito` con
los cuatro `h23-cita-*` (quickstart.md §4 a §6).

- [X] T001 [aceptacion] Los cuatro guiones `testscript` de la entrega, escritos desde spec.md (US1, US2 y US4 con sus escenarios, y FR-001 a FR-030) y desde contracts/applet-cita.md §1 a §7, sin código de producto y sin ninguna eval (plan.md, «Aceptación e2e»: las seis evals entran con el formato que las admite, en T007). Son `cita-preparar.txtar`, `cita-cotejar.txtar`, `cita-herramientas.txtar` y `cita-sin-efectos.txtar`, con esos nombres exactos, porque plan.md, «Controles de umbral», nombra sus copias activadas. **Común a los cuatro**: un comentario de cabecera que nombra los FR, SC y escenarios que cubre y que no contiene los textos `usage: `, `unknown command`, `unexpected command` ni `cannot parse` (el «rojo primero» los toma por un guion ilegible); antes de cualquier otra orden, la precondición `exec kitlegal --help` seguida de `stdout '^\s+cita\s'`, que hoy se ejecuta bien y falla en su aserción (FR-001; el precedente, los guiones `h21-mcp-`); después, solo las órdenes y variables que el arnés ya tiene (`exec`, `stdin`, `stdout`, `stderr`, `cmp`, `grep`, `arbol`, `mcp`, `KITLEGAL_T0_BIN`, `KITLEGAL_SKILLS`), sin ninguna nueva (research D24, V8); el fragmento se toma del repositorio por la ruta que da la cabecera de este fichero («El fragmento»), sin copiarlo al guion. Afirman lo que el spec y el contrato fijan —códigos, `ok`, la clase `argumentos`, `fuente`, `url`, claves y valores de vocabulario de `data`, nombres y valores de casilla, direcciones— y nada del texto libre: ni `motivo`, ni `explicacion`, ni el mensaje de un error, ni `hash`, ni `fecha_consulta`. **cita-preparar** (US1.1 a US1.6; FR-001, FR-004, FR-005, FR-006, FR-010 a FR-015, FR-082; SC-003): con `--json`, el ECLI, `--roj`, `--resolucion` con `--fecha` y `--texto` del spec terminan con 0 y dan `"fuente":"kitlegal.cita"`, la `url` de contracts §2 y el `data` de contracts §3 —`referencia` o `texto`, `equivalente` en las dos primeras y ausente en las otras, `direccion`, y `casillas` en su orden con su `nombre`, su `campo` y su `valor`; el texto con espacio y tilde, codificado— y ninguna lleva `cobertura`; el ECLI del Tribunal Constitucional termina con 0, con `cobertura` y su `cendoj` `no-cubierto`, sin `direccion`, con `casillas` vacía y `url` `kitlegal:applet/cita`; y terminan con 2, `"ok":false` y la clase `argumentos` el ECLI mal formado, el de otro país, `--resolucion` sin `--fecha`, la referencia junto a `--texto` —con un texto y con `--texto ""`—, la referencia junto a `--roj ""` y `kitlegal cita` sin verbo. **cita-cotejar** (US2.1 a US2.6; FR-004, FR-020 a FR-026, FR-082; SC-003): con el fragmento por `stdin`, sin referencia da 0, los ocho datos de `ficha` de contracts §4, `"correspondencia":"se-corresponden"`, `"hallazgos":[]`, ni `pedida` ni `es_la_pedida`, y la `url` con la huella del fragmento; pedido por su ECLI, por su ROJ y por su número con su fecha, 0 y `"es_la_pedida":true` sin hallazgo; pedido como ROJ `STS 1088/2023`, 0, `"ok":true`, `"es_la_pedida":false` y un hallazgo `documento-distinto` cuyo `difiere` nombra el ROJ pedido y el del documento, con `cruce` `numero-de-resolucion`; con la fecha `2023-01-01`, el hallazgo con `fecha` en `difiere` y sin `cruce`; con el número `3144/2023`, `cruce` `numero-del-roj`; un texto sin ficha (el final del fragmento, sacado con una orden, y una frase cualquiera), 2 y `argumentos`; y `--documento ""` con el fragmento en la entrada, que no se lee, 2 y `argumentos`. **cita-herramientas** (US4.3, US4.4; FR-020, FR-025, FR-026, FR-030): con la orden `mcp` y `KITLEGAL_T0_BIN`, el servidor anuncia `cita_cotejar` y `cita_preparar`, las dos de solo lectura (el total de herramientas no se afirma aquí: lo fija el control de conformidad); `cita_cotejar` con la ficha en `documento` y el ROJ `STS 1088/2023` da `resultado`, no error, con su hallazgo; sin `documento`, y con él vacío, `error argumentos`, y la llamada siguiente del mismo proceso da `resultado`; y **una llamada a cada una de las dos herramientas devuelve el sobre de su orden** (US4.3, entero): dan `resultado`, y el sobre de cada una se compara con `cmp` con el de la orden de los mismos argumentos, quitada de los dos `fecha_consulta` (research V10), seis llamadas, que son las que la skill hace en modo herramienta en las seis evals (contracts/evals-jurisprudencia.md §5) más la del ECLI, de modo que cada uno de los seis argumentos de contracts §1 —`ecli`, y los cinco que en la orden son banderas, `roj`, `resolucion`, `fecha`, `texto` y `documento`— viaja en una llamada con resultado: `cita_cotejar` con la ficha sola en `documento`, frente a `cita cotejar --json` con la ficha por `stdin`, y con la ficha y `roj` `STS 1088/2023`, frente a `cita cotejar --roj "STS 1088/2023" --json` con la ficha por `stdin`; y `cita_preparar` con `resolucion` `1088/2023` y `fecha` `2023-07-04`, frente a `cita preparar --resolucion 1088/2023 --fecha 2023-07-04 --json`, con `texto` `cláusula suelo`, frente a `cita preparar --texto "cláusula suelo" --json`, con `roj` `STS 1088/2023`, frente a `cita preparar --roj "STS 1088/2023" --json`, y con `ecli` `ECLI:ES:TS:2023:3144`, frente a `cita preparar ECLI:ES:TS:2023:3144 --json`; del sobre de la llamada con `resolucion` y `fecha` se afirman además sus tres casillas —«Nº Resolución» con `1088/2023` y «Fecha resolución», «Desde» y «Hasta», con `04/07/2023`—, del de `texto`, su `direccion` codificada, y del de `roj`, la casilla «Nº ROJ» con `STS 1088/2023`, que es lo que las evals 01, 03 y 06 esperan en la respuesta. Ejecutado en el prototipo en la sesión del corrector de tasks, con las ocho llamadas en un solo proceso del servidor: las seis, `resultado`, y cada comparación, sin diferencias (research V42). Las once líneas de la ficha llegan al fichero de llamadas con una orden desde el fragmento, nunca tecleadas, y el propio guion comprueba con `cmp` que son sus once primeras líneas. **cita-sin-efectos** (US4.2; FR-003, FR-085; SC-006): con la caché y el grafo ya con contenido —una lectura de `boe articulo` sobre la reproducción, como en los guiones de `graph` ya activos—, `arbol` del directorio de la caché antes; después, cada uno de los dos verbos como orden y cada uno como herramienta (US4.2 nombra los dos): `cita preparar` y `cita cotejar` con una orden que termina con 0 cada uno, y entre las órdenes de los dos, una que falla, una con `--offline --no-graph` y una con `--dry-run`; y `cita_preparar` y `cita_cotejar` con una llamada que da `resultado` cada una, afirmado en la salida de la orden `mcp`; `arbol` otra vez, y `cmp` de los dos listados. **Verificación**: los cuatro se validan contra un prototipo de un solo uso bajo `$TMPDIR`, fuera del repositorio —el de research («El prototipo») si sigue allí, o una copia de `HEAD` exportada con `git archive` con el applet mínimo de contracts—, ejecutado con `go -C`: allí los cuatro pasan, y con un mutante por guion (una casilla con otro valor, `es_la_pedida` invertido, el error de `documento` vacío quitado, un fichero escrito en la caché por el verbo) el suyo falla, y `cita-herramientas` falla también con un segundo mutante, el de lo que US4.3 pide a `cita_preparar`: la propiedad `fecha` de una llamada, que deja de llegar al verbo como bandera; el prototipo no se versiona; en el repositorio, el «rojo primero» da los cuatro en rojo por la aserción de su precondición; `make ci` en verde, sin ningún fichero fuera del directorio de la suite — FR-001, FR-003, FR-004, FR-005, FR-006, FR-010, FR-011, FR-012, FR-013, FR-014, FR-015, FR-020, FR-021, FR-022, FR-023, FR-024, FR-025, FR-026, FR-030, FR-082, FR-085; SC-003, SC-006. Rutas: specs/019-h23-skill-jurisprudencia-ninguna/aceptacion/

**Checkpoint**: la entrega está descrita y congelada; nada del producto ha cambiado.

---

## Fase 2: Fundación — los identificadores, el dominio de la cita, la entrada estándar y las banderas en `--describe`

**Objetivo**: las piezas de dentro que el applet compone (plan.md, pasos 2 a 5): el ECLI y el ROJ con su equivalencia,
la referencia, la consulta preparada, la ficha y el cotejo, sin el kernel; la entrada estándar entregada por el kernel;
y `--describe` diciendo qué argumentos son banderas, para que la tabla de comandos de la skill dé órdenes que existen.

**Prueba independiente**: `go test -count=1 -run '^(TestAnalizarECLI|TestAnalizarROJ|TestEquivalencia|FuzzECLI|FuzzROJ|TestSuperficieDeIds)$' ./internal/core/ids/`,
`go test -count=1 ./internal/core/cita/`, `go test -count=1 -run '^TestEntradaDeLaOrden$' ./internal/app/` y
`make schema-check`.

- [X] T002 [datos] El ECLI y el ROJ de una sentencia y su equivalencia, con sus dos fuzz y el corpus que los fija, y **nada más** (data-model §1; research D1, D2, D28, S6; plan.md paso 2). **`ecli.go`**, nuevo: el tipo `ECLI`, valor inmutable que solo se construye con `AnalizarECLI`, con la forma de FR-005 —cinco partes separadas por dos puntos, `ECLI` y `ES` literales, órgano de 1 a 7 letras mayúsculas ASCII o cifras con la primera letra, año de cuatro cifras y número de 1 a 25 letras mayúsculas ASCII, cifras o puntos—, sin recortar la entrada ni pasarla a mayúsculas y sin buscar el órgano en ninguna lista; el de otro país se rechaza diciendo que no es español; y `ECLI.String`, `ECLI.Organo` y `ECLI.ROJ`. **`roj.go`**, nuevo: `ROJ` y `AnalizarROJ` —siglas en una o más palabras de mayúsculas ASCII separadas por un solo espacio, número en cifras y año de cuatro—, `ROJ.String`, `ROJ.ECLI` y `ROJ.Numero`, que da el número con su año tal como van tras las siglas. **La equivalencia** (FR-012) se deduce en una sola pareja, con el número y el año trasladados carácter a carácter y sin normalizar: un ECLI de órgano `TS` con número solo de cifras da el ROJ de siglas `STS`, y un ROJ de siglas `STS` da ese ECLI; en cualquier otro caso los dos métodos dicen que no se deduce y no devuelven un identificador a medias. Todo rechazo es un error de la clase `argumentos`, construido como los de `CodigoINE` y `DIR3`, con un mensaje en español que nombra la entrada entre comillas y lo que le falla. El material de lectura son esos dos ficheros en la rama `018-h23-cita-resolver-comprobar`, leídos con `git show`; de esa rama no se trae nada del formulario ni ninguna grabación. **Tests**: `TestAnalizarECLI` y `TestAnalizarROJ`, tablas con lo que se acepta y con cada rechazo (vacío, minúsculas, un blanco delante o detrás, una parte de más o de menos, órgano de ocho caracteres o que empieza por cifra, año de tres cifras, número de veintiséis caracteres, otro país; siglas con dos espacios seguidos, año de dos cifras, número con letras); `TestEquivalencia`, con la pareja en los dos sentidos y la tabla de FR-012 sin equivalente —un ECLI de otro órgano, un ROJ de otras siglas, un número final con letras y otro con puntos—; `FuzzECLI` y `FuzzROJ`, que comprueban que ninguna entrada hace fallar al programa y que lo aceptado tiene su forma y vuelve a su texto con `String`, con sus semillas en `f.Add` y su corpus mínimo versionado; y `TestSuperficieDeIds`, que gana los nombres exportados y ninguno más. Verificación: con la comprobación del país quitada un momento, el caso del ECLI de otro país falla, y con la condición del órgano `TS` quitada, falla la tabla sin equivalente (los dos mutantes se retiran antes de `make ci`); `make ci` en verde — FR-005, FR-006, FR-012, FR-014, FR-023, FR-025, FR-083. Rutas: internal/core/ids/ecli.go, internal/core/ids/roj.go, internal/core/ids/errores.go, internal/core/ids/doc.go, internal/core/ids/testdata/fuzz/FuzzECLI/, internal/core/ids/testdata/fuzz/FuzzROJ/
- [X] T003 [datos] El dominio de la cita —la referencia, la consulta preparada, la ficha y el cotejo—, puro y sin el kernel, con el fuzz de la lectura de la ficha y su corpus, y **nada más** (data-model §2 a §5; contracts/applet-cita.md §3 a §6; research D1, D3 a D9, V6; plan.md paso 3). Paquete nuevo, que solo importa el de identificadores, el del sobre y la biblioteca estándar, sin `net` (R1, FR-002). **`doc.go`**: el comentario del paquete. **`referencia.go`**: `Referencia` (`forma`, `valor`, `fecha`) y su construcción desde los cuatro argumentos —ECLI, ROJ, número y fecha—, cada uno dado o no dado como puntero a cadena: el que llega escrito con valor vacío está dado y se comprueba con su forma (research D3); errores de FR-006, todos de la clase `argumentos` y con lo que nombra cada fila de contracts §6: una forma mal escrita, la fecha que no es un día que existe, el número sin su fecha, la fecha sin su número y más de una forma dada. **`consulta.go`**: `Consulta`, `Equivalente`, `Casilla` y `Cobertura`, con sus claves JSON de contracts §3 —lo que no aplica no está, y `casillas` es siempre una lista— y las etiquetas de enumeración que `--describe` lleva al esquema; `Preparar`, que da la dirección del buscador y las casillas de la tabla de FR-011 con la fecha escrita como la escribe quien rellena la casilla, el equivalente solo si T002 lo deduce y, con un ECLI de órgano `TC`, `cobertura` sin dirección ni casillas (FR-014); y `PrepararTexto`, que rechaza el texto vacío o solo de blancos y compone la dirección de búsqueda de FR-013 codificando el texto como un segmento de ruta, octeto a octeto y con el hexadecimal en mayúsculas. **`ficha.go`**: `Ficha` con sus ocho claves de contracts §4 y `LeerFicha`, con las reglas de FR-021 y de contracts §5 y ninguna más: la primera línea `Roj:` del texto, sus dos partes, y cada uno de los otros seis datos de la primera línea posterior que lleva su etiqueta entera; finales de línea de los dos tipos y blancos en los extremos; el error nombra el dato que falta o que no tiene su forma, o que no hay ficha (FR-026). **`cotejo.go`**: `Cotejo`, `Hallazgo` y `Diferencia`, y `Cotejar`, con la correspondencia de tres valores (FR-023), `pedida` y `es_la_pedida` solo con una referencia (FR-024), y un hallazgo `documento-distinto` como mucho, con cada dato que difiere, el cruce `numero-de-resolucion` o `numero-del-roj` cuando lo hay y una explicación en español (FR-025). **Tests**: `TestReferencia` y `TestPreparar`, con cada argumento no dado, dado y dado vacío, las cuatro salidas del spec, los casos sin equivalente y `cobertura` solo con el ECLI de órgano `TC`; `TestLeerFicha`, con el fragmento leído de su fichero, con texto delante, con finales de los dos tipos, con cada uno de los ocho datos quitado y con cada uno de los cuatro que se comparan sin su forma; `TestCotejar`, con los siete casos de FR-081 y una ficha de otro órgano, que da `no-se-deduce`; y `FuzzLeerFicha`, con su corpus mínimo versionado: ninguna entrada hace fallar, y lo aceptado tiene sus ocho datos con su forma. Verificación: con la etiqueta de «Nº de Resolución» tomada un momento por prefijo, el caso que la distingue de «Nº de Recurso» falla; con el cruce quitado, fallan sus dos casos (se restaura antes de `make ci`); `make ci` en verde — FR-002, FR-005, FR-006, FR-010, FR-011, FR-012, FR-013, FR-014, FR-015, FR-021, FR-022, FR-023, FR-024, FR-025, FR-026, FR-083. Rutas: internal/core/cita/doc.go, internal/core/cita/referencia.go, internal/core/cita/consulta.go, internal/core/cita/ficha.go, internal/core/cita/cotejo.go, internal/core/cita/testdata/fuzz/FuzzLeerFicha/
- [X] T004 La entrada estándar la entrega el kernel a la orden que la lee, y nunca a una llamada de herramienta (research D12, D13, V23, V24; contracts/applet-cita.md §7; plan.md paso 4 y *Complexity Tracking*, primera fila). **`registro.go`**: `Registro.LeerDe`, que registra la entrada estándar de las órdenes; todavía no la registra ninguna raíz de composición, que llega con el applet en T006. **`despacho.go`**: el campo `Despacho.Entrada`. **`main.go`**: el despacho de una orden recibe la entrada del registro, y el kernel se la da, antes de ejecutarlo, al verbo cuyos argumentos cumplen una interfaz sin exportar, como la que ya tiene para el verbo que sirve; una llamada de herramienta construye su `Despacho` sin entrada, así que ese verbo recibe una entrada nula sin tocar el camino de las herramientas. Ni `mcp.go` ni `herramientas.go` cambian. **Test**: `TestEntradaDeLaOrden` (`main_test.go`), con un applet del propio test cuyo verbo lee su entrada: en una orden recibe los bytes registrados con `LeerDe`; con un `Despacho` como el de una llamada de herramienta, ninguna; con un registro sin entrada registrada, ninguna; y un verbo que no la lee no recibe nada. Verificación: con la entrega al verbo quitada un momento, el caso de la orden falla (se restaura antes de `make ci`); los tests que comparan un `Despacho` entero siguen en verde sin tocarlos; `make ci` en verde — FR-020, FR-026, FR-030. Rutas: internal/app/registro.go, internal/app/despacho.go, internal/app/main.go
- [ ] T005 [datos] `--describe` dice qué argumentos propios de un verbo son banderas y la tabla de comandos las escribe como tales, con el esquema publicado que cambia a la vez, y **nada más** (research D14, V4, V5, V14 a V17; contracts/applet-cita.md §8; plan.md paso 5 y *Complexity Tracking*, segunda fila). **`describe.go`**: la parte `entrada` del documento de un verbo gana la anotación `x-banderas`, con los nombres de sus banderas propias en su orden, **solo si las tiene**: el documento de un verbo sin ellas no cambia un byte; la anotación no va en el esquema de entrada de una herramienta, que sigue siendo el de hoy. **`comandos.go`**: la orden de una fila de la tabla escribe los argumentos de posición como hoy y, detrás, cada bandera propia entre corchetes, con su nombre precedido de dos guiones y el de su valor, o sola si es booleana. **El esquema** schemas/instalacion.json, regenerado con `-actualizar-esquemas` y no editado: gana la anotación en sus tres partes, porque los tres verbos de `skills` ya tenían banderas propias; ningún otro esquema publicado cambia. **Tests**: los de `describe_test.go`, con un verbo con banderas y otro sin ellas; los de `comandos_test.go`, con un verbo con argumentos de posición y banderas, de cadena y booleana; `TestTablaDeComandosCoincideConLaGramatica`, cuyo auxiliar que convierte la orden escrita en una invocación pasa a entender una bandera, de modo que la fila de cada verbo de `skills` vuelve a casar con su gramática; y `TestEsquemasDeHerramienta` y el auxiliar que quita las banderas globales en el test de conformidad, que comparan el esquema de la herramienta con el de `--describe` sin la anotación. Las tablas de `boe-legislacion` y de `legal-core` no cambian: sus verbos no tienen banderas propias, y `make skills-check` lo ve. Verificación: con la condición «solo si las tiene» quitada un momento, `schema-check` da diferencias en los esquemas de los verbos sin banderas; con la anotación quitada, la tabla de un verbo de `skills` deja de casar con su gramática (se restaura antes de `make ci`); `make ci` en verde, con `schema-check` y `skills-check` sin diferencias — FR-030, FR-040, FR-049, FR-088. Rutas: internal/cli/describe.go, internal/skills/comandos.go, schemas/instalacion.json, internal/app/skills_test.go, internal/cli/herramienta_test.go, internal/app/herramientas_test.go

**Checkpoint**: el dominio se prueba solo, el kernel sabe entregar la entrada y la tabla de comandos sabe escribir una
bandera; el binario todavía no tiene el applet.

---

## Fase 3: US1, US2 y US4 — El applet `cita`: preparar la consulta, cotejar el documento, sin red y como herramienta (P1) 🎯 MVP

**Objetivo**: `kitlegal cita preparar` y `kitlegal cita cotejar` en el binario y en el de e2e, con su sobre, su
esquema, sus golden y sus dos herramientas, sin alcanzar la red y sin tocar la caché ni el grafo (plan.md, paso 6).

**Prueba independiente**: quickstart.md §1 a §6, con los cuatro guiones congelados como copias momentáneas.

- [ ] T006 [datos] [US1] El applet `cita` con sus dos verbos en los dos registros, con su esquema, sus once golden y las listas literales que su llegada cambia, en una sola tarea porque con el applet registrado `schema-check` exige su esquema y cuatro guiones enumeran los applets o cuentan las herramientas (contracts/applet-cita.md §1 a §9; research D3, D6, D9 a D12, D26, D27, V7, V9 a V12, V21, V22, V38 a V42; plan.md paso 6 y *Complexity Tracking*, tercera y cuarta filas). **`cita.go`**, nuevo: `AppletCita`, sin dependencias, con `preparar` y `cotejar` y ningún verbo por omisión (FR-001); los argumentos propios de los dos verbos son campos puntero —`ecli` de posición y `roj`, `resolucion`, `fecha` y `texto` o `documento` como banderas, el último un literal que conserva sus bytes—, de modo que uno escrito con valor vacío llega dado y da el error de su fila de contracts §6 (research D3); `preparar` y `cotejar` llaman al dominio de T003 y firman el sobre de contracts §2: `fuente` `kitlegal.cita`; `url`, la dirección que abre quien pregunta o `kitlegal:applet/cita` si no hay ninguna y en todo fallo, y en `cotejar` la huella SHA-256 del texto recibido con el prefijo de contracts §2; `fecha_consulta`, la del reloj del kernel; ninguna operación de grafo, ninguna salida legible propia. `cotejar` toma el texto de `documento` si se escribió, también vacío, y solo sin él lee la entrada que le da el kernel (T004); sin texto, y siempre en una llamada de herramienta sin `documento`, el error de argumentos (FR-020, FR-026); la referencia se comprueba antes que el texto. Que el documento no sea el pedido termina con 0 y `ok` verdadero (FR-025). **Los dos registros**: `AppletCita` en `RegistroDeProduccion` y en el del binario de e2e, y la entrada estándar del proceso registrada con `LeerDe` en las dos raíces de composición. **El esquema** schemas/cita.json, nuevo, regenerado con `-actualizar-esquemas`, con las partes `cotejar` y `preparar` y su anotación `x-banderas` (T005), y su fila y su contrato de verbos en la tabla de esquemas de `esquemas_test.go`. **Los golden**, en internal/app/testdata/cita/: los once de contracts §9, con el sobre entero y `fecha_consulta` puesta a `0001-01-01T00:00:00Z`; los escribe el verbo —con la bandera de actualización que el test tenga o defina—, nunca una mano, y los de `cotejar` salen del fragmento leído de su fichero. **La garantía sin red** (`arch_test.go`): los dos paquetes del dominio entre los orígenes que no alcanzan `net`, `net/http` ni el cliente HTTP del kit, y `cita.go` entre los ficheros sin red del paquete de aplicación (research D27). **Tests**: `TestCitaPreparar` y `TestCitaCotejar` (`cita_test.go`): los once golden iguales byte a byte salvo `fecha_consulta` y las once salidas válidas contra su esquema (FR-080, FR-081; SC-004); el sobre de FR-004; cada fila de errores de contracts §6, también con el argumento escrito vacío, por la línea de órdenes y por la línea que compone una llamada de herramienta; y `--documento ""` con un documento en la entrada, que no se lee. `TestHerramientasDelServidor` pasa de 10 y 13 herramientas a 12 y 15, con `cita_cotejar` y `cita_preparar` (FR-030; SC-009). `TestPuntoDeEntrada`, `TestRegistroDeProduccion` y `TestRegistroDeE2E` ganan `cita` en su lista de applets; `TestEsquemasPublicados` y `TestEsquemasCubrenTodosLosVerbos`, la fila nueva. **Los cuatro guiones**: internal/app/testdata/script/argumentos.txtar gana `cita` en sus dos listas de applets, en su orden; internal/app/testdata/script/h21-mcp-herramientas.txtar pasa de trece a quince herramientas, con sus dos líneas nuevas en el orden del servidor; e internal/app/testdata/script/h21-mcp-proceso.txtar e internal/app/testdata/script/h21-mcp-protocolo.txtar, de trece a quince; ninguno pierde nada de lo que comprobaba. **`CONTRIBUTING.md`**: la línea que reproduce el fallo de un applet desconocido, con la lista nueva. **Verificación**: los cuatro guiones congelados de T001, copiados un momento al directorio de guiones de e2e con el prefijo `zz-`, pasan con `go test -count=1 -run '^TestEntregaDelHito$'` filtrado por ese prefijo, y las copias se borran; con `AppletCita` fuera del registro un momento, `TestHerramientasDelServidor` falla; con un `import` de `net` añadido un momento a `cita.go`, `TestArquitectura` falla; con un byte cambiado en un golden, su caso falla (los tres mutantes se retiran antes de `make ci`); el paso de empaquetado, el paquete del servidor MCP, `mcp.go` y `herramientas.go` no cambian (FR-030, FR-031); `make ci` en verde, con `schema-check` sin diferencias — FR-001, FR-002, FR-003, FR-004, FR-006, FR-010, FR-011, FR-012, FR-013, FR-014, FR-015, FR-020, FR-022, FR-024, FR-025, FR-026, FR-030, FR-031, FR-080, FR-081, FR-082, FR-084, FR-085; SC-003, SC-004, SC-005, SC-006, SC-009. Rutas: internal/app/cita.go, internal/app/registro.go, internal/app/ejemplo/kitlegal-e2e/main.go, cmd/kitlegal/main_test.go, internal/app/herramientas_test.go, internal/app/esquemas_test.go, internal/arch_test.go, schemas/cita.json, internal/app/testdata/cita/, internal/app/testdata/script/argumentos.txtar, internal/app/testdata/script/h21-mcp-herramientas.txtar, internal/app/testdata/script/h21-mcp-proceso.txtar, internal/app/testdata/script/h21-mcp-protocolo.txtar, CONTRIBUTING.md

**Checkpoint**: quien tiene el binario prepara una consulta y coteja un documento, desde la terminal o como
herramienta; US1, US2 y US4 se comprueban solas con quickstart.md §1 a §6.

---

## Fase 4: US3 (primera parte) — El formato de eval, su juicio sin modelo y las seis evals (P1)

**Objetivo**: lo que mide a la skill, antes que la skill (Definition of Done §1.10): `sentencias` y los comandos de
`cita` en el formato común, su juicio por la forma y las seis evals de `jurisprudencia` (plan.md, paso 7).

**Prueba independiente**: `go test -count=1 -run '^(TestFormatoDeSentencias|TestJuzgarSentencias|TestPreguntasConElFragmento|TestConjuntoDeEvals|TestEvalsDelRepositorio)$' ./internal/evals/`.

- [ ] T007 [datos] [US3] El formato de eval gana `sentencias` y los dos comandos de `cita`, con su juicio sin modelo, las reglas del conjunto y las seis evals que los usan, en una sola tarea porque `make ci` lee las evals de toda carpeta y el lector estricto rechaza una clave que el tipo no tiene (contracts/evals-jurisprudencia.md §1, §2, §3, §5 y §7; data-model §7; research D18, D19, D23, V13, V27, V28, V34, V37; plan.md paso 7). **El esquema** schemas/eval.yaml.json: las dos formas de comando y la clave `sentencias` de contracts §1, con sus definiciones del ECLI y del ROJ y la regla de la raíz —`sentencias` solo en una eval que activa la skill y que no es sin binario ni servidor, y con ella `comandos` deja de ser obligatorio—; ninguna clave de hoy cambia de significado. **`formato.go`**: `Eval.Sentencias` con sus siete claves y los comandos de `cita` con `roj` y `con_texto`. **`sentencias.go`**, nuevo: lo que se reconoce en un texto, con las reglas de contracts §3 y de FR-061 —la cita por su corchete exacto, la línea con las tolerancias de las formas fijas que ya hay, la línea que empieza por la marca de aviso, el ECLI del país que sea sin los puntos en que termine y comparado sin distinguir mayúsculas, cada sobre de la sesión línea a línea, el ECLI que leyó un `cita cotejar` y la dirección que devolvió un `cita preparar` con texto—, y el juicio de `sentencias` con el motivo de cada fila de contracts §2. **`juzgar.go`**: el juicio de una sesión aplica `sentencias` y reconoce el comando de `cita` pedido como orden y como herramienta, con su `--roj` en sus dos escrituras y su `--texto` no vacío. **`consultas.go`**: un comando de `cita` no pide ninguna respuesta grabada al preparar la sesión (FR-054). **`conjunto.go`**: `ReglasDeJurisprudencia`, las de contracts §5 (FR-056). **Las seis evals**, con el contenido literal de contracts §5 y un comentario de cabecera que dice qué mide cada una; en la 04 y en la 06, el fragmento va en el bloque literal de la pregunta, llevado con una orden desde su fichero —cada línea con la sangría del bloque y las líneas en blanco vacías—, nunca tecleado. Ninguna carpeta `juez` (FR-055). **Tests**: `TestFormatoDeSentencias` (`formato_test.go`): las claves nuevas se leen, una eval de hoy se lee igual y `sentencias` en una eval que no activa la skill se rechaza; `TestJuzgarSentencias` (`sentencias_test.go`): cada fila de contracts §2 con una respuesta que pasa y otra que no, y el comando por orden y por herramienta (FR-087); `TestPreguntasConElFragmento` (`conjunto_test.go`): la pregunta de la 04 y la de la 06, tal como las lee `LeerEval`, contienen el fragmento byte a byte, y con un byte cambiado en una copia en `t.TempDir()` falla (FR-053; SC-008); `TestConjuntoDeEvals` gana las reglas de `jurisprudencia` sobre evals sintéticas, con cada regla incumplida nombrada, y `TestEvalsDelRepositorio`, la subprueba de su carpeta; los tests del formato, del juicio y de las consultas de las otras dos skills siguen en verde sin cambiar lo que comprueban (FR-052). Verificación: con una de las seis evals apartada un momento, la subprueba de la carpeta falla nombrando la regla; con `ninguna_cita` quitada de la 01, igual (se restaura antes de `make ci`); `shasum -a 256` del fragmento da la huella de la cabecera de este fichero; `make ci` en verde, con `skills-check` sin diferencias — FR-050, FR-051, FR-052, FR-053, FR-054, FR-055, FR-056, FR-087; SC-008. Rutas: schemas/eval.yaml.json, internal/evals/formato.go, internal/evals/sentencias.go, internal/evals/juzgar.go, internal/evals/consultas.go, internal/evals/conjunto.go, evals/jurisprudencia/01-existe-con-numero-y-fecha.yaml, evals/jurisprudencia/02-resumen-sin-documento.yaml, evals/jurisprudencia/03-por-materia.yaml, evals/jurisprudencia/04-documento-pegado.yaml, evals/jurisprudencia/05-tribunal-constitucional.yaml, evals/jurisprudencia/06-documento-que-no-es-el-pedido.yaml

**Checkpoint**: las seis evals existen y `make ci` sabe leerlas y juzgar una sesión contra ellas; todavía no hay skill
que las pase ni umbral que decida.

---

## Fase 5: US5 (primera parte) — El umbral que decide: `cita_sin_documento` (P1)

**Objetivo**: el informe de una skill cuyas evals declaran `sentencias` publica, por modo, `sin_activar` y
`cita_sin_documento`, con `decide: true`, y una sola respuesta que cuenta da `fallo` (plan.md, paso 8).

**Prueba independiente**: `go test -count=1 -run '^(TestCitaSinDocumento|TestUmbralesDeJurisprudencia|TestUmbralesDelInforme)$' ./internal/evals/`.

- [ ] T008 [US5] El hecho de la sesión y sus umbrales: ninguna respuesta lleva una cita sin documento cotejado ni un ECLI sin origen (contracts/evals-jurisprudencia.md §3, §4 y §7; data-model §7; research D19 a D21, V26; plan.md paso 8). **`sentencias.go`**: si una respuesta cuenta, con las dos condiciones de FR-060 y las definiciones de FR-061 —(1) una cita cuyo ECLI no leyó ningún `cita cotejar` de su sesión que terminara con `ok` verdadero, donde el ECLI que la salida repite por habérsele dado como referencia pedida no vale; (2) un ECLI, fuera de una cita y fuera de una línea que empieza por la marca de aviso, que no está en ningún sobre de su sesión, terminara como terminara, ni en la pregunta de su eval—, y qué ECLI la hacen contar y por cuál de las dos. **`umbrales.go`**: una skill tiene umbrales de sus respuestas si tiene juez o si alguna de sus evals declara `sentencias`, sin ningún nombre de skill en el código; `sin_activar` se publica en los dos casos, las clases solo con juez y `cita_sin_documento` solo con `sentencias`, por cada modo del plan y con el del modo orden delante, los dos con comparación menor o igual, umbral 0 y `decide: true`, y con el mismo total: las respuestas del modelo que decide en ese modo en las evals que activan la skill. **`informe.go`**: el motivo de un `cita_sin_documento` incumplido nombra detrás de su medida cada sesión que cuenta, con sus ECLI y con `(cita sin documento cotejado)` o `(sin origen)`, como en contracts §4, y con un motivo el veredicto es `fallo`. La firma de lo que el sondeo llama no cambia, y el sondeo tampoco (spec, «Fuera de alcance»). **Tests**: `TestCitaSinDocumento` (`sentencias_test.go`), los casos de FR-086 sobre una respuesta: una cita cuyo ECLI no leyó ningún cotejo, que cuenta; la misma con ese ECLI leído por un cotejo de la sesión, que no; una cita cuyo ECLI solo se dio al cotejo como referencia pedida, que cuenta; un ECLI suelto sin origen, que cuenta; el mismo en la pregunta, en el sobre de un `cita preparar` o dentro de una línea de aviso, que no; y un ECLI suelto escrito delante del punto que cierra la frase, en la respuesta o en la pregunta, que no. `TestUmbralesDeJurisprudencia` (`informe_test.go`), con las evals del repositorio y sesiones sintéticas: `umbrales` son exactamente cuatro, en el orden de contracts §4, ninguno de una clase de juez ni de duración (FR-055, FR-063); con 0 respuestas que cuentan, los cuatro se cumplen y el veredicto no falla por ellos; con una en un modo y ninguna en el otro, `cumple: false` en ese modo, veredicto `fallo` y el motivo con el umbral, su modo, su medida y la sesión; y con una respuesta sin la skill activada, `fallo` por `sin_activar`. Los tests de hoy sobre el informe de `boe-legislacion`, con sus doce umbrales, y sobre el de `legal-core`, con ninguno, siguen en verde sin cambiar lo que afirman (FR-063). Verificación: con `decide` puesto un momento a falso en el umbral nuevo, el caso de la respuesta que cuenta deja de dar `fallo` y el test lo ve; con la condición (2) quitada, falla su caso (los dos mutantes se retiran antes de `make ci`); `make ci` en verde — FR-055, FR-060, FR-061, FR-062, FR-063, FR-064, FR-086; SC-007. Rutas: internal/evals/sentencias.go, internal/evals/umbrales.go, internal/evals/informe.go

**Checkpoint**: el instrumento está entero en `make ci`; falta la skill que mide.

---

## Fase 6: US3 (segunda parte) — La skill `jurisprudencia` v0 (P1)

**Objetivo**: la skill empotrada en el binario, con su tabla de comandos generada, y todo lo que enumera las skills
alineado (plan.md, paso 9). Va detrás de sus evals y de su instrumento.

**Prueba independiente**: quickstart.md §7.

- [ ] T009 [datos] [US3] La skill `jurisprudencia` v0, con el arnés de skills que deja de suponer referencias y los diecinueve guiones que enumeran las skills empotradas, en una sola tarea porque en cuanto la carpeta existe el binario la empotra y la instala (contracts/skill-jurisprudencia.md §1, §2 y §4; contracts/SKILL-jurisprudencia.prototipo.md; research D15 a D17, V15, V18 a V21, M2, S1; plan.md paso 9; la cabecera de este fichero, «Diecinueve guiones»). **`SKILL.md`**, nuevo, sin `references` ni `scripts`: el borrador de contracts copiado tal cual, con su frontmatter (`kitlegal-applets: cita`, sin `kitlegal-referencias`) y sus siete pasos, y la tabla de comandos regenerada con `make skills-sync`; si la regenerada difiere de la del borrador, manda la regenerada. Lleva lo que contracts §1 traza a su requisito: la sentencia se cita solo con su documento delante y su ficha cotejada (FR-041); la línea `⚠ SENTENCIA NO COMPROBADA:` con su forma, una por referencia, con la consulta tomada de lo que devuelve `cita preparar`, sin decir que existe ni que no existe, y sin preparar como ROJ una cita sin fecha (FR-042); el cotejo de la ficha y nada más, el documento que no es el pedido y `no-se-corresponden` (FR-043, FR-048); la cita con su corchete y los datos de la ficha (FR-044); ningún resumen de lo que no está delante (FR-045); la pregunta por materia (FR-046); el Tribunal Constitucional, no cubierto, con la dirección de su buscador en un solo sitio, el paso 7, que es carácter a carácter la de la eval 05 (FR-047; research S1); y las reglas invariantes (FR-049). Menos de 300 líneas; no nombra evals, el job ni modelos. **El arnés** (internal/app/skills_test.go): `TestSkillsDelRepositorio` deja de suponer que toda skill declara referencias —el frontmatter sintético sin la clave cuando no hay ninguna, y la carpeta creada antes de escribir en ella la referencia sobrante (research V18)— y exige también la skill nueva donde enumera las exigidas. **`linea-sin-consulta`** (internal/evals/conjunto_test.go): la subprueba recorre las dos skills que llevan la línea del BOE, las de su lista, con la misma exigencia, y no toda skill del directorio (research D16). **Los diecinueve guiones** ganan la tercera skill en su orden —`boe-legislacion`, `jurisprudencia`, `legal-core`— en cada listado, recuento y comparación que las enumera, sin perder nada de lo que comprobaban: los diecisiete del directorio de e2e y, con la etiqueta `integration`, los dos de la instalación de desarrollo. **Verificación**: `make skills-check` sin diferencias; `wc -l` del fichero por debajo de 300, y alargado un momento hasta 300 líneas, `skills-check` falla (FR-040; SC-009); `TestOrdenesDeLasSkillsEmpotradas` en verde, con las órdenes de la tabla como verbos del registro; `go test -count=1 -tags=integration -run '^TestInstalacion$'` sobre el paquete de skills, en verde; `git diff --stat` sin ninguna línea en las otras dos skills (FR-049); el paso de empaquetado no cambia (FR-031); `make ci` en verde — FR-031, FR-040, FR-041, FR-042, FR-043, FR-044, FR-045, FR-046, FR-047, FR-048, FR-049, FR-088; SC-009. Rutas: skills/jurisprudencia/SKILL.md, internal/app/skills_test.go, internal/evals/conjunto_test.go, internal/app/testdata/script/skills-salida-legible.txtar, internal/app/testdata/script/skills-host-antigravity.txtar, internal/app/testdata/script/h19-skills-ambito-dir.txtar, internal/app/testdata/script/h19-skills-ambito-global.txtar, internal/app/testdata/script/h19-skills-aviso.txtar, internal/app/testdata/script/h19-skills-aviso-sin-aviso.txtar, internal/app/testdata/script/h19-skills-conflictos-dentro.txtar, internal/app/testdata/script/h19-skills-conflictos-entradas.txtar, internal/app/testdata/script/h19-skills-conflictos-rutas.txtar, internal/app/testdata/script/h19-skills-doctor-copia.txtar, internal/app/testdata/script/h19-skills-doctor-hallazgos.txtar, internal/app/testdata/script/h19-skills-dry-run.txtar, internal/app/testdata/script/h19-skills-idempotencia.txtar, internal/app/testdata/script/h19-skills-install-hosts.txtar, internal/app/testdata/script/h19-skills-install-local.txtar, internal/app/testdata/script/h19-skills-list-doctor.txtar, internal/app/testdata/script/h19-skills-no-empotrada.txtar, internal/skills/testdata/script/instalar.txtar, internal/skills/testdata/script/instalar-sin-gobin.txtar

**Checkpoint**: la skill viaja en el binario y se instala como las demás; si el modelo la sigue lo mide el job del
cierre (`cita_sin_documento` y `sin_activar` con 0 en los dos modos, SC-001).

---

## Fase 7: US5 (segunda parte) — El trabajo `evals (jurisprudencia)` (P1)

**Objetivo**: el job de evals mide `jurisprudencia` como una skill más de su matriz, sin objetivo de duración
(plan.md, paso 10).

**Prueba independiente**: `go test -count=1 -run '^TestDefinicionDelJob$' ./internal/evals/`.

- [ ] T010 [US5] `jurisprudencia` en la matriz del job de evals, comprobada en `make ci` (contracts/evals-jurisprudencia.md §6; research D22, M5, V29, V32, S4; plan.md paso 10). **El flujo**: la matriz del trabajo de evals pasa a `boe-legislacion`, `jurisprudencia` y `legal-core`, con la entrada de `include` de la nueva —concurrencia 1 y objetivo de duración 0, como `legal-core`—; el trabajo se llama `evals (jurisprudencia)`; el tope del trabajo no cambia, porque su peor caso calculado, 20 341 s, cabe en él; ni prueba de red, ni juez, ni grabaciones, ni otra clave del flujo; el trabajo de la medida del juez sigue solo con `boe-legislacion`. **`TestDefinicionDelJob`** (`definicion_test.go`): la lista de skills del trabajo y sus ajustes ganan la tercera, y con definiciones sintéticas falla si la skill no está en la matriz, si su objetivo de duración no es 0 o si el tope queda por debajo del peor caso de alguna de las tres, calculado con la carpeta de evals del repositorio (FR-063). Verificación: con la skill quitada un momento de la matriz del flujo, el test falla nombrándola (se restaura antes de `make ci`); `make ci` en verde. El job no se ejecuta en esta tarea: lo lanza el workflow en el cierre — FR-050, FR-054, FR-063; SC-001. Rutas: .github/workflows/evals.yml, internal/evals/definicion.go

**Checkpoint**: el cierre tendrá un trabajo `evals (jurisprudencia)` que contar.

---

## Fase 8: US6 — La documentación (P3)

**Objetivo**: quien lee el README sabe qué hace kitlegal con las sentencias y qué no (plan.md, paso 11).

**Prueba independiente**: lectura de los cinco ficheros contra FR-070 (SC-010), y `make ci`.

- [ ] T011 [US6] La documentación que el hito deja falsa o incompleta (contracts/applet-cita.md; contracts/evals-jurisprudencia.md; contracts/skill-jurisprudencia.md; plan.md paso 11). README.md, en «¿Y las sentencias?»: kitlegal sigue sin consultar el CENDOJ, y lo que hace —prepara la consulta exacta, coteja el documento que trae quien pregunta y solo cita la sentencia cuyo documento tiene delante—, con el ADR 0036 como decisión y sin remitir a la sustituida; fuera de esa sección, solo la frase o la lista que el hito deja falsa (las skills que hay, las herramientas del servidor), sin prometer ninguna release. docs/JURISPRUDENCIA.md: lo que pasa a estar hecho —el applet `cita` con sus dos verbos y la skill—, en su apartado del hito y en lo pendiente, sin tocar lo que documenta lo probado a mano. CHANGELOG.md, *Unreleased*: `kitlegal cita preparar` y `kitlegal cita cotejar` con lo que dan y sus códigos, las herramientas `cita_preparar` y `cita_cotejar`, la skill `jurisprudencia` con lo que hace para quien la usa, la anotación `x-banderas` de `--describe` y las banderas en la tabla de comandos, y lo del job: `sentencias` y los comandos de `cita` en el formato de eval, el umbral `cita_sin_documento` y el trabajo `evals (jurisprudencia)`. CONTRIBUTING.md, en «Skills y evals» («Formato común de eval» y «Job de evals»): las claves nuevas y su juicio sin modelo, qué cuenta como salida de una operación, los cuatro umbrales de `jurisprudencia`, que no tiene juez hasta H25 y la tercera skill de la matriz; y, donde describe el servidor MCP o `--describe`, las dos herramientas nuevas y la anotación. internal/evals/doc.go: el comentario del paquete con `sentencias.go` y el umbral nuevo. Verificación: cada orden, fichero, test, clave y umbral que se nombra existe en el árbol (`grep` uno a uno); ningún texto dice que una sentencia se comprueba sin su documento ni que kitlegal consulta el CENDOJ; `make ci` en verde. No se editan los ADR, el roadmap, la bitácora de uso, la memoria del proyecto, la web ni los specs de otros hitos — FR-070, FR-088; SC-010. Rutas: README.md, docs/JURISPRUDENCIA.md, CHANGELOG.md, CONTRIBUTING.md, internal/evals/doc.go

---

## Fase 9: Cierre — Definition of Done

- [ ] T012 Cierre de la Definition of Done, sin código de producto (`ROADMAP.md` §1; plan.md, «Obligaciones para tasks.md»): specs/019-h23-skill-jurisprudencia-ninguna/cierre.md con (1) la lista, sacada de `git diff --name-status main`, de los esquemas, los golden, los corpus de fuzz, los guiones, las evals y la skill creados o modificados, y la comprobación de que no cambia ninguna grabación, ninguna fuente, ningún manifiesto, `SOURCES.md`, ningún ADR, ningún spec o plan de otro hito, el directorio de los guiones del workflow, la carpeta de evidencias, el paso de empaquetado y su paquete, el paquete del servidor MCP, `mcp.go`, `herramientas.go`, el cliente HTTP del kit ni las otras dos skills (FR-031, FR-049), de que la carpeta de evals de `jurisprudencia` no tiene `juez` (FR-055) y de que la huella del fragmento es la de la cabecera de este fichero; (2) la cobertura sobre `coverage.out` y `coverage-integration.out` de `make ci` con `go tool cover -func` —global ≥ 70 % y el dominio ≥ 85 % (Definition of Done §1.9)— y la de los dos paquetes del dominio nuevos o ampliados; (3) el resultado de quickstart.md §0 a §9, orden a orden, con lo esperado de cada una —los guiones `h23-` de §4 a §6, que el workflow activa tras el bucle, ejercidos con copias momentáneas de prefijo `zz-`, que se borran—, y la constancia de que §10 lo hace el workflow y §11 queda para después del run; (4) la tabla de plan.md, «Controles de umbral», con el test de cada fila `ci:` encontrado como `^func <Test>\(` en su ruta, que es como lo busca el informe final, y los tres guiones `h23-` como pendientes de la activación; (5) la constancia de que ningún umbral de FR-060 o de FR-062 se rebajó, pasó a `decide: false` ni perdió evals o un modo de su total (FR-064), de que los dos supuestos de alcance de FR-065 —«resume» y `no-se-corresponden`— siguen en `gates/supuestos.md` como no comprobados, y de que el árbol no lleva ningún `//nolint`, `t.Skip` ni TODO nuevos. Solo si una cifra de cobertura queda bajo su umbral se añaden los tests que faltan, en los `_test.go` de los ficheros declarados — FR-031, FR-049, FR-055, FR-064, FR-065, FR-088; SC-009. Rutas: specs/019-h23-skill-jurisprudencia-ninguna/cierre.md, internal/core/ids/ecli.go, internal/core/ids/roj.go, internal/core/cita/referencia.go, internal/core/cita/consulta.go, internal/core/cita/ficha.go, internal/core/cita/cotejo.go, internal/app/cita.go, internal/evals/sentencias.go

---

## Dependencias y orden de ejecución

El orden es el de plan.md, «Orden de implementación (de dentro afuera)», y es estrictamente secuencial: T001 → T012.
Ninguna tarea depende de una posterior.

| Tarea | Paso del plan | Necesita |
|---|---|---|
| T001 | 1 | — |
| T002 | 2 | — |
| T003 | 3 | T002 (`ECLI`, `ROJ` y su equivalencia) |
| T004 | 4 | — |
| T005 | 5 | — |
| T006 | 6 | T003 (el dominio), T004 (la entrada que el kernel da al verbo), T005 (la anotación de su esquema); pasa los guiones de T001 |
| T007 | 7 | T006 (el applet `cita` existe para los comandos de las evals y da el sobre que el juicio reconoce) |
| T008 | 8 | T007 (`sentencias`, el reconocimiento y las evals del repositorio) |
| T009 | 9 | T005 (la tabla con banderas), T006 (los verbos de la tabla), T007 (sus evals, antes que ella) |
| T010 | 10 | T007 (la carpeta de evals, para el peor caso), T009 (la skill que el trabajo mide) |
| T011 | 11 | T001 a T010 |
| T012 | — | T001 a T011 |

### Historias

- **US1 y US2** (T002, T003, T006) son las dos mitades de lo que gana la skill, y **US4** (T004, T006) lo que las deja
  llegar como herramienta sin red, sin caché y sin grafo. **US3** es el instrumento que la mide (T007) y la skill
  (T009); **US5**, el umbral (T008) y su trabajo (T010); **US6**, la documentación (T011).
- Las historias no son independientes en su construcción —US1, US2 y US4 comparten el applet, y US3 y US5, el paquete
  de evals— y sí en su prueba: cada fase nombra la orden que la comprueba sola.

### Oportunidades de paralelismo

Ninguna dentro del run: el workflow ejecuta las tareas una a una. Ninguna tarea lleva `[P]`. Fuera de un run, T002,
T004 y T005 no dependen de ninguna anterior.

---

## Trazabilidad

### Requisitos → tareas

| Requisito | Tareas |
|---|---|
| FR-001 | T001 (guion), T006 |
| FR-002 | T003 (el dominio no importa red), T006 (`TestArquitectura`) |
| FR-003 | T001 (guion `cita-sin-efectos`), T006 |
| FR-004 | T001, T006 (el sobre, el esquema y los golden) |
| FR-005 | T002 (ECLI y ROJ), T003 (la referencia) |
| FR-006 | T002, T003, T006 (también con el argumento vacío) |
| FR-010, FR-011, FR-013, FR-015 | T003, T006; T001 |
| FR-012 | T002 (la equivalencia y su tabla), T003, T006 |
| FR-014 | T002 (`ECLI.Organo`), T003 (`cobertura`), T006 |
| FR-020 | T004 (la entrada), T006 (`documento`); T001 |
| FR-021 | T003 |
| FR-022, FR-024, FR-025 | T003, T006; T001 |
| FR-023 | T002 (la regla), T003 (la correspondencia) |
| FR-026 | T003 (la ficha), T004 (la llamada sin entrada), T006 |
| FR-030 | T004, T005 (el esquema de la herramienta sin la anotación), T006 (`TestHerramientasDelServidor`); T001 (guion `cita-herramientas`: una llamada con resultado a cada una de las dos, con el sobre de su orden) |
| FR-031 | Restricción de todas las tareas (batería); T006 y T009 la nombran; T012 la deja constatada |
| FR-040 | T005 (la tabla con cada orden), T009 |
| FR-041 a FR-048 | T009 (FR-041 y FR-044 los decide el umbral de T008; FR-045 y FR-048, sin control: FR-065) |
| FR-049 | T005 (las tablas de las otras dos no cambian), T009; T012 la deja constatada |
| FR-050 | T007 (las seis evals), T010 (los dos modos del job) |
| FR-051, FR-052, FR-053, FR-056 | T007 |
| FR-054 | T007 (sin grabaciones), T010 |
| FR-055 | T007 (ninguna carpeta `juez`), T008 (ningún umbral de juez); T012 |
| FR-060, FR-061, FR-062 | T008 |
| FR-063 | T008 (los cuatro umbrales y los de las otras dos skills), T010 (la matriz) |
| FR-064 | Restricción de todas las tareas (batería); T008; T012 la deja constatada |
| FR-065 | Los dos supuestos ya están en `gates/supuestos.md` (los anotó el plan); T012 constata que siguen |
| FR-066 | plan.md, «Controles de umbral»; la tabla siguiente |
| FR-070 | T011 |
| FR-080, FR-081 | T006 |
| FR-082 | T001 (los guiones), T006 (los pasa) |
| FR-083 | T002 (`FuzzECLI`, `FuzzROJ`), T003 (`FuzzLeerFicha`) |
| FR-084 | T006 |
| FR-085 | T001 (guion), T006 |
| FR-086 | T008 |
| FR-087 | T007 |
| FR-088 | Todas (`make ci` por tarea); T005 y T009 (`skills-check`), T006 (`schema-check`); T012 |
| SC-001 | El cierre del workflow, con los controles de T008 y T010 y la skill de T009; no es una tarea |
| SC-002 | Una persona, después del run; no es una tarea |
| SC-003 | T001, T006 |
| SC-004, SC-005 | T006 |
| SC-006 | T001, T006 |
| SC-007 | T008 |
| SC-008 | T007 |
| SC-009 | T006 (12 herramientas), T009 (menos de 300 líneas), T012 |
| SC-010 | T011; lo comprueba la revisión final |

### Controles de umbral (plan.md) → tarea que construye el control

| Fila de plan.md | Control | Tarea | Test que lo ve fallar |
|---|---|---|---|
| FR-060, SC-001 (los dos modos) | `cita_sin_documento:claude-sonnet-5-5:<modo>` | T008 | `TestUmbralesDeJurisprudencia`: 1 respuesta que cuenta en un modo da `cumple: false` y `fallo`; `TestCitaSinDocumento`, por respuesta |
| FR-062, SC-001 (sin activar, los dos modos) | `sin_activar:claude-sonnet-5-5:<modo>` | T008 | `TestUmbralesDeJurisprudencia`: 1 respuesta sin la skill activada da `fallo` |
| SC-001 (`boe-legislacion` sigue aprobada) | Sus diez umbrales que deciden, los de hoy | T008 (no los cambia) | `TestUmbralesDelInforme` y los del informe de H24, sin cambiar lo que afirman: fallan si los doce dejan de ser esos |
| FR-040, SC-009 (líneas) | `ci:Makefile:skills-check`, `TestSkillsDelRepositorio` | T009 | El fichero alargado un momento a 300 líneas |
| FR-030, SC-009 (herramientas) | `TestHerramientasDelServidor` | T006 | Con 12 y 15; el applet fuera del registro un momento |
| FR-082, SC-003 | Los guiones `h23-cita-preparar` y `h23-cita-cotejar` | T001 (los escribe), T006 (los pasa) | Los once casos con su código; en el prototipo, un mutante por guion |
| FR-080, FR-081, SC-004 | `TestCitaPreparar`, `TestCitaCotejar` | T006 | Un byte cambiado en un golden; cada salida contra su esquema |
| FR-002, FR-084, SC-005 | `TestArquitectura` | T006 | Un `import` de `net` un momento en `cita.go` |
| FR-003, FR-085, SC-006 | El guion `h23-cita-sin-efectos` | T001, T006 | En el prototipo, un fichero escrito en la caché por el verbo |
| FR-061, FR-086, SC-007 | `TestCitaSinDocumento`, `TestUmbralesDeJurisprudencia` | T008 | 1 respuesta que cuenta da `fallo`; 0, se cumple; 1 en un modo y 0 en el otro, `fallo` |
| FR-055, FR-063 | `TestUmbralesDeJurisprudencia`, `TestDefinicionDelJob` | T008, T010 | Otros umbrales que los cuatro; la skill fuera de la matriz, con objetivo, o el tope bajo el peor caso |
| FR-053, SC-008 | `TestPreguntasConElFragmento` | T007 | Un byte cambiado en la pregunta de una copia |
| FR-056 | `TestEvalsDelRepositorio` | T007 | Una eval apartada, o sin una clave que espera; `TestConjuntoDeEvals`, regla a regla |

### Definition of Done (`ROADMAP.md` §1) → tareas

| Punto | Aplica | Tareas |
|---|---|---|
| 1 · `make ci` en verde | Sí | Todas; T012 |
| 2 · Tests offline; fixtures si toca red | Sí; no toca red | T002 a T010; ningún fixture grabado |
| 3 · Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | Sí | T003, T004, T006 (y más: el applet no alcanza `net`) |
| 4 · `schemas/` y `schema-check` | Sí | T005, T006 (de `--describe`), T007 (el del formato de eval) |
| 5 · Errores tipados y códigos estables | Sí: `argumentos` → 2; hallazgo → 0 (ADR 0023) | T002, T003, T006 |
| 6 · e2e y `CHANGELOG.md` | Sí | T001, T006, T009 (guiones); T011 |
| 7 · ADR | No: la decisión es el ADR 0036 con su enmienda, ya aceptado | — |
| 8 · `SOURCES.md` | No: ninguna fuente (FR-088) | — |
| 9 · Cobertura | Sí | T012 |
| 10 · Skill: evals antes, `SKILL.md` < 300, sin drift | Sí | T007 (antes), T009 |
| 11 · Dimensión territorial | No (FR-088) | — |
| 12 · Operaciones de grafo | No: lo que lee `cita` no viene de una fuente (FR-003, FR-088) | T006 comprueba que no devuelve ninguna |
| 13 · Recursos, plazos o escritos | No | — |

---

## Estrategia de implementación

- **De dentro afuera** (plan.md): los identificadores, el dominio, la entrada del kernel y las banderas; el applet;
  las evals y su juicio; el umbral; la skill; el job; y la documentación.
- **Lo mínimo que ya da valor** es T001 a T006: quien tiene el binario prepara la consulta y coteja su documento,
  también como herramienta. No es una entrega parcial del hito: el run ejecuta las doce, y SC-001 pide la skill con
  sus umbrales.
- **Lo que el run no puede medir** (research S1 a S5): que el modelo que decide activa la skill, pide las operaciones
  y escribe la línea y la cita con su forma; que Bash acepta la ficha por la entrada estándar; cuánto tarda el trabajo;
  que la dirección del buscador del Tribunal Constitucional es la que da la skill; y que toda ficha del CENDOJ lleva
  las ocho etiquetas. Lo mide el job del cierre o lo comprueba una persona; si una respuesta cuenta en
  `cita_sin_documento`, la reparación corrige `SKILL.md` con la sesión y sus ECLI delante y no toca el umbral, las
  evals ni lo que cuenta como salida de una operación (FR-064).
- **Después del run** (SC-002): una persona lee las respuestas a las evals 02, 04 y 06 en los dos modos y anota si
  alguna resume una sentencia cuyo texto no tenía delante. Ninguna release lleva `jurisprudencia` hasta H25.

---

## Comprobación contra la rúbrica de `juez_tasks` y `precheck.sh tasks`

- `precheck.sh tasks`: doce líneas con el formato `- [ ] Tnnn`, ids únicos y correlativos; todas con rutas; las seis
  que tocan `schemas/` o `testdata/` llevan `[datos]`, y ninguna otra línea nombra esos directorios; ninguna nombra la
  carpeta de evidencias por su ruta, ni la variable de grabación, ni una acción de plataforma, ni pide nada a una
  persona; una sola tarea de aceptación, la primera. Comprobado en esta sesión con las expresiones del guion sobre las
  doce líneas; el guion entero no se ejecutó, porque su primera comprobación es `gates/analyze.md`, que escribe el
  paso siguiente.
- a · `analisis_critico`: lo escribe el paso de análisis después de este fichero.
- b · `trazabilidad`: cada tarea cita sus FR y SC; «Requisitos → tareas» cubre FR-001 a FR-088 y SC-001 a SC-010; y
  «Controles de umbral» da la tarea y el test de cada fila de plan.md.
- c · `rebanadas_verdes`: cada tarea trae su test y su implementación; las excepciones están declaradas arriba; el
  orden es el del plan y la tabla de dependencias no tiene ninguna hacia delante. Lo que rompe cada llegada —el
  applet, la anotación, la skill— va en su misma tarea, con la lista medida en el prototipo, también con la etiqueta
  `integration`.
- d · `rutas_declaradas`: tras «Rutas:», ficheros concretos, la carpeta de la suite, la de los golden y las tres de
  corpus; ninguna ruta genérica.
- e · `datos_separados`: T002, T003, T005, T006, T007 y T009, con lo que el dato obliga a cambiar a la vez (plan.md,
  *Complexity Tracking*); ningún dato externo, ninguna grabación, ningún fichero de `data/`.
- f · `dod`: tabla «Definition of Done → tareas»; `CHANGELOG.md` en T011; esquemas en T005 a T007; sin ADR ni
  `SOURCES.md`, con su motivo.
- g · `checklist_veraz`: `checklists/requirements.md` no cambia.
- h · `aceptacion_primero`: T001, con los cuatro guiones de plan.md y los escenarios de US1, US2 y US4 que cubre cada
  uno; US4.3 entero, con una llamada con resultado a cada una de las dos herramientas comparada con su orden, y US4.2
  con cada verbo como orden y como herramienta (research V42); US3, US5 y US6 no tienen guion, con su motivo.
- i · `autonomia`: ninguna tarea exige a una persona, publica, ni mide en la plataforma.
