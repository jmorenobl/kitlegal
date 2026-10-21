# Research: H26.1 · Las evals sirven lo que la skill lleva a leer y hacen fallar a propósito la lectura que quieren probar

Fuentes de este documento: el spec (`spec.md`), el ADR 0043, la constitución 2.13.0, `docs/ROADMAP.md` §2, §3 y la
sección de H26.1, y el código del repositorio en `6bf2da3`. Tres clases de entrada: **V** (verificado en esta sesión,
contra el código, la documentación local o una ejecución que se vio terminar), **D** (decisión, con la alternativa
rechazada y por qué) y **S** (supuesto que no se ha podido verificar sin red ni sesión con modelo).

Ninguna incógnita del contexto técnico queda abierta.

## Verificaciones (V)

**V1 · Lo que `data/` lleva a leer está grabado hoy (FR-005).** Ejecutado en esta sesión con un test temporal en
`internal/evals/` (borrado al terminar; `git status` limpio después): leídos `data/plazos.yaml` con `plazos.LeerTipos`,
`data/cobertura.yaml` con `skills.LeerCobertura` y `data/normas.yaml` con `skills.LeerNormas`, salen **17 bloques**
(los preceptos de los tipos, de sus unidades y de los cómputos, el precepto de cada entrada de la cobertura y el
aplicable de su norma) y **19 normas** (las de esos bloques, las de la cobertura y las de `data/normas.yaml`): 72
consultas, `boe articulo <norma> <bloque>` sin fecha y con `--fecha 9999-12-31` por bloque e `indice`/`metadatos` por
norma. `Preparar(dir, UnionDeGrabaciones(), consultas)` dio **0 faltas**. Los 17 bloques: `BOE-A-2015-10565` `a21`,
`a24`, `a30`, `a112`, `a122`, `a124`, `daprimera`; `BOE-A-1998-16718` `a46`, `a128`; `BOE-A-1985-12666`
`acientoochentaydos`, `acientoochentaytres`, `acientoochentaycinco`; `BOE-A-1889-4763` `art5`; `BOE-A-2013-12887` `a20`,
`a24`; `BOE-A-2003-23186` `a97`; `BOE-A-2015-11722` `a1`. El ADR 0043 dice lo mismo («17 bloques, y el índice y los
metadatos de cada norma»).

**V2 · Los cuatro bloques de `lecturas` no están grabados y sí están en los índices grabados (FR-010).** En la misma
ejecución: `Preparar` con `boe articulo BOE-A-2015-10566 a26`, `… a27`, `BOE-A-2006-20764 a18` y `BOE-A-2004-4214 a15`
dio 4 faltas, cada una «no hay grabación de GET https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/<norma>/texto/bloque/<bloque>».
`boe indice <norma> --offline --json` de las tres normas, desde la caché de V1, terminó con 0 y su salida lleva
`"a26"`, `"a27"`, `"a18"` y `"a15"`. `BOE-A-2004-4214`, `BOE-A-2006-20764` y `BOE-A-2015-10566` son entradas de
`data/normas.yaml` (líneas 34, 42 y 49).

**V3 · Una caché preparada sin las dos lecturas de un bloque no lo sirve con ninguna fecha (FR-013).** En la misma
ejecución, con las 72 consultas de V1 menos `articulo BOE-A-2015-10565 a124` y `… a124 --fecha 9999-12-31`:
`boe articulo BOE-A-2015-10565 a124 --offline --json` terminó con **4** sin fecha, con `--fecha 2026-09-09` y con
`--fecha 9999-12-31`; sin `--offline`, sobre la reproducción de un directorio vacío (`registroDeBoe`), con **1** (la
reproducción sin grabación es un defecto inesperado); y con las 72 consultas enteras, con **0** en los seis casos. La
lectura sin fecha y la fechada se guardan bajo claves distintas del mismo bloque, `claveDelArticulo` y
`claveDeLasVersiones` (`internal/source/boe/entradas.go:135` y `:143`), y la de `9999-12-31` sirve después cualquier
fecha (`internal/evals/consultas.go:20-25`).

**V4 · La línea con la que el binario dice que no pudo escribir su salida estándar (FR-024, FR-025).** En la misma
ejecución, `app.Main` con `boe indice BOE-A-2015-10565 --json` y una salida estándar que falla con `syscall.EPIPE`
terminó con **1** y escribió en la salida de error exactamente `render: no se pudo escribir en la salida estándar: broken pipe\n`.
La compone `enDescriptor` (`internal/render/render.go:91-100`, `"render: no se pudo escribir en %s: %w"`, con
`nombreSalida = "la salida estándar"` en la línea 29); la lleva a la salida de error `Montador.emitirFallo`, que escribe
`err.Error()` con `Aviso` (`internal/cli/sobre.go:104-110` y `:156`), y `escribirTexto` le pone el salto de línea
(`internal/render/texto.go:21-33`). `TestTuberiaCerrada` (`internal/app/tuberia_unix_test.go:23` y `:68`) ya exige al
binario real, con una tubería sin lector, el código 1 y «no se pudo escribir en la salida estándar» en su salida de
error.

**V5 · El binario no enlaza `internal/evals`.** `go list -deps ./cmd/kitlegal | grep -c kitlegal/internal/evals` dio 0.
Los guiones de H19 comparan el `SKILL.md` instalado con el del árbol con `cmp`, sea cual sea su texto
(`internal/app/testdata/script/h19-skills-ambito-dir.txtar:40`).

**V6 · La caché de una sesión es la de las consultas necesarias de las evals de su skill.** `prepararLaCache` llama a
`Preparar(dirCache, s.Grabaciones, ConsultasNecesarias(evals))` (`internal/evals/preparar.go:319-330`), y
`ConsultasNecesarias` recorre comandos, normas de la eval y citas, sin repetir invocaciones y sumando orígenes
(`internal/evals/consultas.go:76-181`). La base de la medida del juez es otra preparación con las mismas consultas,
compartida por las sesiones de la skill (`internal/evals/medida.go:1053-1074` y `:1114-1146`). Las preparaciones solo
corren en proceso, con `httpx.Replay` (`internal/evals/preparar.go:401-413`), sin red.

**V7 · Dónde se comprueba que lo que necesitan las evals está grabado.** `TestEvalsDelRepositorio`, subtests `grabado`
(`boe-legislacion`) y `conjunto-legal-core`, prepara con `Preparar` y comprueba con `ComprobarSinRed` las consultas de
cada skill (`internal/evals/conjunto_test.go:2044-2084` y `:2169-2191`); el de `jurisprudencia` exige que no necesite
ninguna (`:2107-2110`). Lo ejecuta `make ci` por `skills-check` (`Makefile:125-127`) y el job antes de abrir ninguna
sesión (`scripts/evals.sh:142-146`).

**V8 · `schema-check` no valida las evals.** `make schema-check` ejecuta `TestEsquemasPublicados` de `internal/app`, que
compara `schemas/` con lo que emite `--describe` (`Makefile:121-123`), y `schemas/eval.yaml.json` solo lo leen
`internal/evals/formato.go`, `formato_test.go` y `doc.go` (`grep -rln`). Cada eval se valida contra ese esquema en
`LeerEval` (`internal/evals/formato.go:561-575`), y una que no lo cumple es un fichero mal formado de `LeerConjunto`, que
`TestEvalsDelRepositorio` no admite. FR-009 se hace cumplir ahí (D9).

**V9 · El límite de líneas de una skill.** `maximoDeLineas = 299` en `internal/skills/sincronia.go:25-27`: 300 líneas o
más es un defecto de `make skills-check`. `skills/legal-core/SKILL.md` tiene hoy **299** (`wc -l`). La versión de la
skill solo se nombra en `CHANGELOG.md` (`grep -rn "1\.2\.2"`: una sola línea, la 28).

**V10 · El peor caso de `legal-core` y el tope del job (FR-020).** Fórmula de `peorCasoDelTrabajo`
(`internal/evals/definicion.go:56-73` y `:421-536`): 485 s más, por cada grupo (modo orden con la prueba de red, modo
herramienta, evals sin binario ni servidor), ⌈sesiones / concurrencia⌉ tandas de 22 + 240 + 10 s. Hoy, con 10 evals
(9 de modo y la 04 sin binario ni servidor), 2 modelos, 3 repeticiones y concurrencia 2
(`.github/workflows/evals.yml:205-207`): `16261 s = 485 s + (⌈55 / 2⌉ + ⌈54 / 2⌉ + ⌈6 / 2⌉) × (22 s + 240 s + 10 s)`,
el literal de `internal/evals/definicion_test.go:1896-1897`. Con la 11 y la 12, de modo y no informativas: 11 evals de
modo × 2 modelos × 3 repeticiones = 66 sesiones por modo, 67 con la prueba de red: `19525 s = 485 s + (⌈67 / 2⌉ +
⌈66 / 2⌉ + ⌈6 / 2⌉) × 272 s`, 70 tandas, **325,4 minutos**, dentro del tope de **352** minutos (21 120 s,
`.github/workflows/evals.yml:226`). Pasa a ser el mayor de los tres (`boe-legislacion`, 18 977 s; `jurisprudencia`,
12 577 s).

**V11 · Grabar.** `scripts/workflow/grabar-datos.sh` graba, tras el commit de una tarea `[datos]`, cada manifiesto
`<paquete>/testdata/grabaciones.json` que la tarea cambió, con su `fuente` revisada en `docs/SOURCES.md` de `main`, con
`KITLEGAL_RECORD=1 KITLEGAL_EVIDENCIAS=… go test -tags=grabacion -count=1 -run '^TestGrabar' ./<paquete>/`
(`grabar-datos.sh:7-33`, `:70-78` y `:103-119`). La fila de `boe.legislacion-consolidada` en `main` está revisada el
2026-09-13 (`git show main:docs/SOURCES.md`, línea 42). `TestGrabarEvals` (`internal/evals/grabacion_test.go`, etiqueta
`grabacion`) sirve primero cada consulta desde H4 y H5 y solo pide a la fuente lo que falta, que escribe en la raíz de
H9, `internal/evals/testdata/` (`grabacion_test.go:36-60`; `internal/evals/grabaciones.go:27-35`). Las búsquedas de las
tres normas de FR-010 («régimen jurídico del sector público» con `Ley 40/2015,`, «impuesto renta personas físicas» con
`Ley 35/2006,` y «reguladora de las haciendas locales» con `Real Decreto Legislativo 2/2004,`) ya están en el manifiesto
de H5, `testdata/evals/grabaciones.json`: solo los bloques son nuevos. `verificarLasNormasDeLaCobertura` resuelve cada
entrada del manifiesto de H9 por su búsqueda grabada (`grabaciones_test.go:1450-1494`), y las tres nuevas se resuelven
igual y no duplican ninguna norma de la cobertura.

**V12 · Cómo se lee hoy una llamada y su código (FR-023).** `leerUser` marca la llamada con resultado como error si su
`is_error` es verdadero o su contenido es un sobre con `ok` falso, y toma la clase de ese sobre (`sobreDeFallo`,
`internal/evals/sesion.go:715-783`); `codigoDeLaLlamada` da 1, el de lo inesperado, al error sin clase
(`internal/evals/juzgar.go:622-639`). Un error sin sobre —el aviso del agente de que la herramienta no existe— sale con 1.
Los tests construyen las llamadas a mano (`llamadaFallida`, `internal/evals/juzgar_test.go:3555-3559`).

**V13 · Lo que se lee de las órdenes de Bash (FR-025, FR-026).** `leerOrdenDeBash` anota cada `tool_use` de Bash cuyo
`input.command` lleva la palabra `kitlegal` (`contieneComoPalabra`, `internal/evals/juzgar.go:1180-1209`), y
`anotarLaSalida` le pone el texto de su `tool_result` (`internal/evals/sesion.go:640-699`). Van a `Sesion.Textos` junto
con las llamadas, sin nada que distinga unas de otras (`sesion.go:155-179`).

**V14 · Lo que del juicio depende del código de una invocación.** Solo `informarLaJuzgada`, que reparte entre fuera de
lo grabado y otras fallidas (`juzgar.go:961-1003`), y `consultoComoPide`, que decide si satisface un comando
(`juzgar.go:1052-1070`). `ejecutaElProhibido` no lo mira (`juzgar.go:876-878`).

**V15 · Los umbrales de las respuestas son solo del modelo que decide.** `respuestasQueSeJuzgan` deja fuera los modelos
informativos (`internal/evals/informe.go:2024-2062`) y `respuestasMedidas.umbrales` publica, por modo,
`sin_activar`, las clases del juez y los de sentencias, cotejo, plazo, cobertura, vigencia y redacción fechada
(`internal/evals/umbrales.go:392-425`). Los modelos del plan son `ModeloQueDecide` y `ModelosInformativos`
(`internal/evals/plan.go:50-76`); en el job, `claude-sonnet-5-5` y `claude-haiku-4-5-20251001`
(`.github/workflows/evals.yml`). Hoy `legal-core` publica 10 umbrales, los 10 deciden; `boe-legislacion`, 15 y 11;
`jurisprudencia`, 15 y 11 (`TestUmbralesDeLasTresSkills`, `internal/evals/informe_test.go:6681-6779`;
`umbralesDeBoeLegislacionSobre`, `:7481-7513`).

**V16 · `encoding/json/v2` y las claves que no van.** El informe se escribe con `encoding/json/v2`
(`internal/evals/informe.go:5-6`). `go doc encoding/json/v2`: `omitempty` omite el campo que se codificaría como `null`,
cadena vacía, objeto vacío o lista vacía; un puntero nil sin opciones se codifica `null`.

**V17 · Reglas de dependencia.** `depguard` solo prohíbe `internal/render` a `internal/core/**` (R1) y a
`internal/graph/**` (R6) (`.golangci.yml:137` y `:208`); `internal/arch_test.go` lleva R1, R2, R3, R6 y R7. Hoy
`internal/evals` importa `internal/app`, `cache`, `cli`, `core/grafo`, `core/plazos`, `core/schema`, `core/territorio`,
`graph`, `httpx`, `skills` y `source/boe` (`go list -f '{{join .Imports "\n"}}' ./internal/evals`).

**V18 · Lo que `CLAUDE.md` y `docs/WORKFLOW.md` dicen de la caché de las sesiones (FR-032).** Nada: `grep -n caché`
en los dos solo da la caché SQLite del binario, `~/.cache/kitlegal/` y la caché de contexto de las sesiones del run.
`CONTRIBUTING.md` documenta el formato de las evals en su tabla (líneas 612-634) y el job en «Job de evals» (desde la
línea 850).

**V19 · Las rutas de `data/` desde el paquete.** Los tests de `internal/evals` leen `../../data/plazos.yaml`,
`../../data/cobertura.yaml` y `../../data/normas.yaml` (`grabaciones_test.go:330`, `:765`, `:768`), y `LeerCobertura`
compila su esquema desde `../../schemas/cobertura.yaml.json` (`internal/skills/cobertura.go:13`): la misma profundidad.
El job ejecuta `go test` en `./internal/evals/` (`scripts/evals.sh:146` y `:173`).

**V20 · Lo que comprueba el conjunto de `legal-core`.** `ReglasDeLegalCore` (`internal/evals/conjunto.go:298-308`) pide
al menos un mínimo de evals, municipio cubierto y no cubierto, no activación, algo verificable en cada eval activa (una
de plazo o de cobertura lo es: `conjunto.go:736-752`), una de plazo y una de cobertura, y una sola sin binario ni
servidor. Ninguna regla compara preguntas: la 11 puede llevar la pregunta de la 05. El esquema exige `razonamiento` en
`plazo` (`schemas/eval.yaml.json`, `properties.plazo.required`). `TestEvalsDelRepositorio` fija en 10 las evals de
`legal-core` (`conjunto_test.go:2064`, `evalsDeLegalCoreConLaDeVigencia`).

**V21 · `pattern` en el esquema.** `santhosh-tekuri/jsonschema` v6.0.3, la del módulo, compila `pattern` con
`regexp.Compile` y lo comprueba con `MatchString`, que dice si la cadena **contiene** una coincidencia
(`$GOMODCACHE/github.com/santhosh-tekuri/jsonschema/v6@v6.0.3/compiler.go:311-331` y `validator.go:457`); el repositorio
no cambia el motor (`grep -rn RegexpEngine internal/`: nada). `"pattern": "\\S"` rechaza la cadena vacía y la que solo
tiene blancos.

## Decisiones (D)

**D1 · Lo que la skill lleva a leer se deriva en `internal/evals/consultas.go` y se suma a las consultas necesarias con
el mismo acumulador (FR-001, FR-002).** Una función sin exportar, `lecturasDeLosDatos(dir)`, lee los tres ficheros de
`dir` con `plazos.LeerTipos`, `skills.LeerCobertura` y `skills.LeerNormas` y anota, con el acumulador de
`ConsultasNecesarias`, cada bloque sin fecha y con `--fecha 9999-12-31` y el `indice` y los `metadatos` de cada norma,
con un punto nuevo, `PuntoLoQueLlevaALeer`, y el fichero de `data/` como origen. `consultasDeLaCache(evals, dir)` da las
necesarias de las evals y, detrás, las de los datos, sin repetir ninguna invocación: la que ya estaba suma su origen.
`dir` es `datosDelRepositorio = "../../data"` en el job y en la medida, y una copia en el test de la falta (FR-003).
*Rechazadas*: meterlas dentro de `ConsultasNecesarias`, que cambiaría su contrato —el de las evals— y el de sus siete
llamadas y sus casos, y haría falso que `jurisprudencia` no necesite ninguna; concatenar dos listas, que repetiría la
misma invocación y daría dos faltas de la misma lectura; y leer `data.Plazos`, empotrado en el binario, que no lleva ni
la cobertura ni las normas: las tres lecturas tienen que salir del mismo sitio, como en los tests de hoy (V19).

**D2 · `lecturas` es el cuarto punto de las consultas necesarias (FR-006).** `ConsultasNecesarias` anota cada entrada de
`lecturas` con `anotarBloque` —sin fecha y con `9999-12-31`— y un punto nuevo, `PuntoLecturaDeLaEval`. Así la caché de
toda sesión de la skill las sirve, `TestEvalsDelRepositorio/grabado` exige su grabación (FR-010) y la base de la medida
las tiene sin tocarla. *Rechazada*: una lista aparte, que habría que añadir en cada sitio que prepara.

**D3 · `sin_servir` se aplica al preparar la caché de la sesión, quitando sus consultas antes de `Preparar` (FR-013).**
`sinLoQueNoSeSirve(consultas, eval)` quita de las de `consultasDeLaCache` cada una que consulta una lectura de
`eval.SinServir`, con el mismo predicado con el que el juicio reconoce una invocación que la consulta (D6): un bloque,
`articulo` o `articulos` de esa norma con ese bloque, con o sin `--fecha`; una consulta de norma, ese verbo con esa
norma. V3 muestra que, sin esas dos consultas, la caché no sirve el bloque con ninguna fecha. *Rechazadas*: borrar
entradas de la caché ya preparada, que pediría SQLite en `internal/evals` (R3) y conocer las claves de `internal/source/boe`;
y una reproducción que falle a propósito, que es la opción 2 del ADR 0043, rechazada.

**D4 · La medida del juez sirve lo mismo que FR-001 y las `lecturas`, sin `sin_servir` (FR-004).** `leerLasEvalsDeHoy`
guarda `consultasDeLaCache(evals, datosDelRepositorio)` en lugar de `ConsultasNecesarias(evals)`. La base es una por
skill y la comparten las sesiones de todas sus evals (V6), y solo tienen medida las skills con juez, `boe-legislacion` y
`jurisprudencia`, cuyas evals no llevan `sin_servir`: aplicarlo en la medida no lo alcanza ninguna vía real (umbral de
materialidad) y FR-004 no lo pide. *Rechazada*: una base por eval, que multiplicaría la preparación sin cambiar nada
que vea una persona.

**D5 · Un comando de bloque cuya lectura está en `sin_servir` admite también 4 y 5 (FR-016).** `satisface` recibe si la
lectura del comando está en el `sin_servir` de su eval, y `consultoComoPide` admite entonces el 4 y el 5 además del 0 y
del 3 de hoy; `leeElBloque` no cambia, así que la lectura tiene que ser en la fecha del comando. *Rechazada*: una forma
de comando nueva, que repetiría `leeElBloque`.

**D6 · El reparto de cada invocación, en este orden (FR-015, FR-023, FR-024).** Ver `data-model.md` §4: la que no
consulta, solo en `invocaciones`; la llamada sin sobre, en `otras_fallidas` con `codigo: null` y `motivo: "sin_sobre"`,
consulte lo que consulte; la que consulta una lectura de `sin_servir`, en `lecturas_sin_servir`, termine como termine,
con `motivo` y `bash` si es una orden de salida no escrita; y las demás, como hoy, con `motivo` y `bash` en
`otras_fallidas` si son una orden de salida no escrita. El orden es el del spec (FR-015 deja fuera de lo intentado solo
la llamada sin sobre; `gates/supuestos.md`, líneas 1 y 4).

**D7 · «Sin sobre» es una llamada cuyo resultado es un error y ningún texto de su contenido es un objeto JSON con `ok`
(FR-023, FR-028).** `leerUser` marca `Llamada.SinSobre` con `ConResultado`, `Error` y ningún sobre leído;
`invocacionDeLaLlamada` le da `Codigo: nil` y el motivo. Una llamada sin resultado sigue sin código y fuera de las
fallidas, como hoy. Un resultado que no es un error y no trae sobre sigue con 0, como hoy: con `null` cambiaría si
satisface un comando, y FR-028 lo prohíbe; el servidor de kitlegal devuelve siempre un sobre (V12). *Rechazada*: todo
resultado sin sobre, que cambiaría el juicio.

**D8 · La línea de la escritura fallida se reconoce con el error del presentador del binario, sin tocar el binario
(FR-025).** `internal/evals` pide una vez a `render.Nuevo(<escritor que falla con una causa conocida>, io.Discard)` su
`Texto` y toma como prefijo lo que su error lleva delante de esa causa: `render: no se pudo escribir en la salida estándar: `
(V4). Una línea de la salida de una orden de Bash es la de FR-025 si empieza por ese prefijo. Es «su error» de FR-025:
la misma pieza que la escribe, sin ninguna copia del mensaje. El test que los ata ejecuta `app.Main` con una salida
estándar que falla y exige que el arnés reconozca la línea que el kernel deja en la salida de error; y un mutante que
cambie el mensaje en `internal/render/render.go:99` hace fallar `TestTuberiaCerrada`, que lo exige al binario real (V4).
*Rechazadas*: una constante exportada en `internal/render` que el binario y el arnés compartan, que cambia un paquete
del binario —y su `enDescriptor`— para algo que el arnés puede pedirle sin cambiarlo; y una copia literal, que FR-025
prohíbe.

**D9 · `lecturas` y `sin_servir` se validan en dos sitios (FR-008, FR-009, FR-012, FR-014).** La forma, en
`schemas/eval.yaml.json`: `lecturas`, lista no vacía de `{norma, bloque, motivo}` sin más claves y con `motivo` de al
menos un carácter que no sea blanco (`pattern: "\\S"`); `sin_servir`, lista no vacía de `{norma, bloque}` o de
`{norma, verbo}` con `verbo` `indice` o `metadatos`, sin más claves (`oneOf`). Lo que cruza campos, en `LeerEval`, detrás
del esquema: un bloque de `lecturas` que ya es un comando de bloque de la eval, con o sin fecha, o una de sus citas, y un
bloque de `sin_servir` que es una de sus citas; cada error nombra el fichero y la entrada, y la eval queda mal formada en
`LeerConjunto`. El control de FR-009 en `make ci` es `TestEvalsDelRepositorio`, que no admite ficheros mal formados, y
los casos de `TestFormatoDeLecturas` (V8). *Rechazada*: validarlo todo en Go, que repetiría el esquema; o todo en el
esquema, que no compara entre campos.

**D10 · Las órdenes de Bash, aparte (FR-026).** `Sesion.OrdenesDeBash` son los usos de Bash que nombran `kitlegal`, con
su orden y su salida, en su orden, leídos en la misma pasada que `Textos` (un indicador en el uso que ya se anota). El
enlace de FR-026 recorre solo esa lista. *Rechazada*: filtrar `Textos`, que mezcla llamadas sin marca (V13).

**D11 · Las sesiones que cuentan en los dos umbrales nuevos (FR-021, FR-033).** Las «sesiones medidas» de un modelo en
un modo son las de las series que pide el plan con ese modelo y ese modo que se pudieron leer y no quedaron sin medir
por un límite de uso (`sin_medir` vacío): las de cualquier eval de ese modo, active o no la skill y termine o no la
sesión. Es lo que el spec llama «sesiones medidas», sin más exclusión que la que nombra: las de las evals sin binario
ni servidor, que no tienen modo («Edge Cases»). `fuera_de_lo_grabado` suma las lecturas de su `fuera_de_lo_grabado`;
`lectura_sin_intentar` cuenta, de las de las evals con un comando de bloque en su `sin_servir`, las que no tienen en
`lecturas_sin_servir` una invocación que consulte el bloque de cada uno de esos comandos: también la que el tope cortó
antes de intentarla, que es una de las seis de SC-013 sin la lectura intentada. *Rechazada*: el filtro de las
«respuestas medidas» de los umbrales de hoy (terminadas y de una eval que activa la skill; `respuestasQueSeJuzgan`,
V15), que dejaría fuera del recuento la lectura fallida de una sesión cortada o de una eval de no activación que el
spec cuenta.

**D12 · Dónde van los umbrales nuevos.** Detrás de todos los de hoy, por cada modo del plan en su orden y, dentro de
cada uno, por cada modelo del plan —el que decide delante y los informativos en su orden—:
`fuera_de_lo_grabado:<modelo>:<modo>` y, solo en la skill con una eval con un comando de bloque en su `sin_servir`,
`lectura_sin_intentar:<modelo>:<modo>`, como hoy `plazo_sin_fecha` solo en la que tiene evals de plazo
(`respuestasMedidas.umbrales`, `internal/evals/umbrales.go:392-425`). Con el plan del job son 8 en `legal-core`, que
queda con 18 (deciden 12), y 4 en `boe-legislacion` y en `jurisprudencia`, que quedan con 19 (deciden 11, los de hoy).
El motivo de un
`lectura_sin_intentar` que decide y no se cumple nombra sus sesiones como el de `plazo_sin_fecha`
(`nombradasPorSuSesion`, `internal/evals/umbrales.go:879-923`). *Rechazada*: intercalarlos entre los de cada modo, que
no los separa de los que miden respuestas del modelo que decide y obliga a reordenar en cada test lo que hoy no cambia.

**D13 · El tope del trabajo no cambia (FR-020).** 19 525 s caben en 352 minutos (V10). `TestDefinicionDelJob` ya pone
en rojo un tope que no cubre el peor caso; cambian sus literales: `peorCasoDeLegalCore` (de 16 261 s a 19 525 s) y el
caso con el tope en 288 minutos (17 280 s), que ahora también deja fuera a `legal-core`. El comentario de
`.github/workflows/evals.yml` sobre los peores casos no lo pide ningún requisito y no se toca.

**D14 · `legal-core` v1.2.3 cabe en 299 líneas (FR-029 a FR-031).** El paso 7 ya dice que la línea `⚠ NO CUBIERTO:` va
aunque falle una lectura (`skills/legal-core/SKILL.md:109-112`); se le añade, en el mismo párrafo, que la respuesta no
expone la norma que rige el procedimiento —sus plazos, sus recursos—, que nombra la línea. El paso 10 gana un punto: si
la lectura de un precepto en la fecha de inicio sale con 4 o 5, la línea `⚠ PLAZO:` se mantiene con la fecha de
`plazos` y la cita de cada precepto de `data.preceptos`, y en lugar de su línea `⚠ REDACCIÓN A FECHA:` la respuesta dice
aparte que no pudo leer esa redacción, como `boe-legislacion` dice cualquier lectura fallida, sin afirmar lo que dice el
precepto. Lo que entra se compensa reescribiendo con menos líneas los dos párrafos que ya tratan la lectura fallida —el
del paso 7 y el primer punto del paso 10—, sin quitar ninguna regla ni ningún caso: el fichero tiene hoy 299 líneas
(V9) y `make skills-check` lo pone en rojo con 300.

**D15 · Aceptación e2e: no aplica (principio III).** El hito no cambia ningún paquete del binario: `cmd/kitlegal` no
enlaza `internal/evals` (V5), `internal/render` no se toca (D8) y FR-028 prohíbe cambiar el código con que termina. La
skill que cambia viaja empotrada y los guiones de H19 ya comparan lo instalado con el árbol, sea cual sea su texto (V5).
La aceptación es la del instrumento y la de la skill: el job de evals de cierre (SC-011 a SC-020), con cada pieza fijada
en `make ci`. Precedentes: H24, H25, H26 y H23.1.

**D16 · Los mutantes se aplican en una copia fuera del árbol y quedan en `mutantes.md` (SC-007, SC-008, FR-013,
FR-025, FR-033).** La tarea que cierra cada pieza aplica su mutante en `git worktree add --detach "$TMPDIR/…"` (nunca
dentro del repositorio: gitleaks y `git add -A` lo recorrerían), ejecuta los tests de la pieza, deja en
`specs/025-h26-1-las-evals-sirven/mutantes.md` la orden exacta y el test que falla, y retira el worktree. El de FR-025
se puede escribir ya, sobre código que existe (`quickstart.md`, escenario 7).

**D17 · Ni `docs/SOURCES.md` ni `scripts/verify-sources.sh` cambian (Definition of Done §1.8).** La grabación de FR-010
pide a la misma fuente la misma clase de dirección que ya graban H5 y H9, `…/id/<norma>/texto/bloque/<bloque>` (V2,
V11), al ritmo de su fila; y `scripts/verify-sources.sh` ya verifica cada noche esa clase con `boe articulo`
(`scripts/verify-sources.sh:2-6`). Cambiar la fila es una decisión de una persona (constitución, capa 3), y nada en el
uso de la fuente la pide. *Rechazada*: tocar la fila para fechar de nuevo la verificación, que no cambia nada de lo que
la fila revisa.

## Supuestos sin verificar (S)

**S1 · En una sesión del job, la lectura que la caché no sirve sale con 5, o con 4.** No se puede reproducir sin una
sesión del job: no hay red ni `strace` en esta. El ADR 0043 lo observó en los jobs de la #175 y la #178 (5, porque la
sesión no alcanza el `robots.txt` del BOE). En `make ci` se comprueba con `--offline`, que da 4 (V3). FR-013 y FR-016
cuentan los dos.

**S2 · El `tool_result` de una orden de Bash trae su salida de error.** Es la suposición del spec («Assumptions») y de
H24, que ya lee esa salida para el juez; no se puede ver sin una sesión con modelo. Si una versión de Claude Code
dejara de traerla, FR-024 no enlazaría ninguna orden y quedarían como hoy (FR-027): nada decide por ello (FR-028).

**S3 · La causa de los `1` de `boe indice` en `boe-legislacion`.** «Con toda probabilidad» una salida estándar cerrada
por un `| python3` (ADR 0043). El job de cierre la confirma o la descarta con lo que publique (SC-017).

## Fuentes y alcance

La única fuente es `boe.legislacion-consolidada`, con fila revisada en `main` (V11). Ninguna fuente sin fila revisada
queda en el plan.
