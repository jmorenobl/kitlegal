# Research: H23 · skill `jurisprudencia` + `cita preparar` y `cita cotejar`

**Fecha**: 2026-10-07 · **Modo**: desatendido. Las decisiones se tomaron con el «Criterio de decisión autónoma» de la
constitución; cada una lleva su alternativa rechazada.

Cómo se lee: **V** es lo comprobado en local en esta sesión, con dónde; **M**, lo medido; **S**, lo que no se pudo
comprobar y se declara supuesto; **D**, las decisiones. Nada de la tabla V es de memoria.

## El prototipo

Para no afirmar nada de Kong, de los generadores de esquemas, del servidor MCP ni de los tests que rompe un applet o
una skill nuevos, se construyó un prototipo **fuera del repositorio**: el árbol de `HEAD` exportado con
`git archive` a `$TMPDIR/kl-proto`, y encima el applet `cita` con sus dos verbos, el reconocimiento de ECLI y de ROJ
leído de la rama `018-h23-cita-resolver-comprobar` (`git show 018-…:internal/core/ids/ecli.go` y `roj.go`), la entrada
estándar entregada por el kernel, la anotación de las banderas en `--describe`, la tabla de comandos con banderas y
un borrador de `skills/jurisprudencia/SKILL.md`. Se ejecutó con `go -C $TMPDIR/kl-proto test …`. No tiene evals, ni
juicio, ni umbrales: esa parte se diseña leyendo `internal/evals`. Del prototipo sale solo un fichero del
repositorio, `contracts/SKILL-jurisprudencia.prototipo.md`; el resto no se versiona. No se ejecutó `golangci-lint`
sobre él (S7).

Tras el rechazo del plan, la corrección llevó al mismo prototipo sus tres cambios —los argumentos de los dos verbos
como campos puntero (D3), `cobertura` solo fuera de cobertura (D4) y el paso 3 del borrador, que pasa la ficha
(D17)— y volvió a ejecutarlo: las filas V38 a V41, y las medidas de M1 a M3, son de esa segunda pasada. Con los
tres cambios, `go test -count=1 ./...` deja en rojo en el prototipo tests y guiones de la lista de M4, y ninguno
que no esté en ella.

Tras el rechazo de las tareas —el guion `cita-herramientas` comparaba con su orden solo el sobre de `cita_cotejar`,
y US4.3 lo pide de una llamada a cada una—, la corrección ejecutó en el mismo prototipo, sin cambiarlo, un guion de
un solo uso con las llamadas que el guion gana: la fila V42 es de esa tercera pasada. El guion y el mutante con que
se le vio fallar se retiraron, y el prototipo quedó como estaba.

## V · Verificado en local

| # | Afirmación | Dónde se comprobó |
|---|---|---|
| V1 | Go del equipo, `go1.27.1`; el árbol exportado compila sin red | `go version`; `go -C $TMPDIR/kl-proto build ./cmd/kitlegal` |
| V2 | Kong v1.16.1 admite en un mismo verbo un argumento de posición opcional y banderas propias con valor; `--<bandera>=<valor>` acepta un valor de varias líneas | `go list -m github.com/alecthomas/kong` → v1.16.1; prototipo: `cita cotejar --help` da `Usage: cita cotejar [<ecli>] [flags]`; llamada de herramienta con `documento` de siete líneas, que el servidor convierte en `--documento=<texto>` |
| V3 | Un campo `string` pasa por una codificación JSON que cambia cada byte no UTF-8 por U+FFFD; `cli.Literal` lo conserva, también detrás de un puntero, y `--describe` lo presenta como cadena | `internal/cli/literal.go` (comentario y `Decode`); prototipo: `"documento": {"type": "string"}` en `--describe`, y el fragmento con dos bytes que no son UTF-8 detrás da la misma `url` por `--documento` y por la entrada estándar |
| V4 | `invopop/jsonschema` v0.14.0 tiene `Schema.Extras map[string]any` para claves fuera del vocabulario, y las serializa | `go doc github.com/invopop/jsonschema Schema` (línea `Extras`); prototipo: `"x-banderas": ["roj","resolucion","fecha","texto"]` en `schemas/cita.json` regenerado |
| V5 | `santhosh-tekuri/jsonschema` v6.0.3 compila y valida un esquema con esa clave desconocida | prototipo: `go test ./...` con la anotación puesta: `TestSalidaContraSuEsquema`, `TestEsquemasPublicados/schemas` e `internal/skills` en verde |
| V6 | Un campo con `omitempty` sale del `required` del esquema generado, y las etiquetas `jsonschema:"enum=…"` de un tipo de salida llegan al esquema | prototipo: `cita preparar --describe` (`cita.Consulta.required` = `casillas`; `cita.Cobertura.required` = `cendoj`, `motivo`; `cita.Cobertura.cendoj.enum` = `no-cubierto`); precedente: `internal/core/territorio/salida.go` y `schemas/municipio.json` |
| V7 | El servidor MCP (SDK v1.8.0) anuncia las dos herramientas sin tocar `mcp`; la llamada con `documento` devuelve el sobre; la llamada sin él devuelve el error `argumentos` y el servidor sigue atendiendo | prototipo, guion con la orden `mcp`: `1 cita_cotejar resultado`, `2 cita_cotejar error argumentos`, `3 cita_preparar resultado`, `4 cita_preparar resultado`, `salida 0` |
| V8 | En `testscript` (go-internal v1.16.0), `stdin <fichero>` da el fichero a la orden siguiente como entrada estándar, también por una ruta absoluta con `..` | prototipo: `stdin $KITLEGAL_SKILLS/../evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt` y `exec $KITLEGAL_T0_BIN cita cotejar --json` → 0 con la ficha |
| V9 | El kernel fecha con su reloj el sobre del applet que no declara `FechaConsulta`, también en el binario `$KITLEGAL_T0_BIN`, que solo fija el reloj de `boe` y de `graph` | `internal/core/schema/sobre.go` (`Procedencia.FechaConsulta`); prototipo: `"fecha_consulta":"2026-10-07T11:09:02.919508+02:00"` con el binario T0 |
| V10 | El sobre de la llamada y el de la orden, con el mismo texto, son iguales byte a byte salvo `fecha_consulta` | prototipo: `sed -E 's/"fecha_consulta":"[^"]*"/"fecha_consulta":""/'` sobre los dos y `cmp` |
| V11 | Tras los dos verbos, la caché y el grafo no cambian | prototipo: con `cache.db` y `world.db` ya creados por `boe articulo`, `arbol cache` antes y después de cuatro llamadas de herramienta, cinco órdenes —una, fallida; otra, con `--offline --no-graph`— y un `--dry-run` da el mismo listado, con las mismas huellas (`cmp`); y con un `KITLEGAL_CACHE_DIR` vacío, 0 entradas tras las 60 invocaciones en proceso de la segunda pasada |
| V12 | El SHA-256 del fragmento que da el sobre de `cita cotejar` es el del manifiesto de la evidencia | prototipo: `url` = `kitlegal:documento/sha256:4886e0c8…ee27`; `evidencias/adr-0036/manifiesto.json`: `"sha256": "4886e0c8ca9836527ec08d8732b50315640a76803033374af5287b2a8321ee27"`, 2 353 bytes |
| V13 | El fragmento no tiene tabuladores, espacios al principio o al final de una línea, comillas dobles ni barras invertidas | `grep -c -E` con esas cinco clases → 0 |
| V14 | La tabla de comandos escribe hoy como argumentos de posición las banderas propias de un verbo | `internal/skills/comandos.go`, `sintaxisDeLaOrden`: «Todos los argumentos propios de un verbo de una tabla son de posición»; prototipo, primera generación: `kitlegal cita preparar [<ecli> [<roj> [<resolucion> [<fecha> [<texto>]]]]]` |
| V15 | Con la anotación, la tabla da `kitlegal cita preparar [<ecli>] [--roj <roj>] [--resolucion <resolucion>] [--fecha <fecha>] [--texto <texto>]` y la de `cotejar`, igual con `[--documento <documento>]` | prototipo: `go test -run '^TestSkillsDelRepositorio$' ./internal/app/ -args -regenerar-skills` y lectura de la región |
| V16 | Los verbos de `skills` ya tienen banderas propias, y `TestTablaDeComandosCoincideConLaGramatica` deja de pasar para `skills doctor`, `skills list` y los dos de `cita` en cuanto la tabla las escribe como banderas: su auxiliar `invocacionDeLaSintaxis` solo entiende argumentos de posición | prototipo: `go test ./internal/app/` |
| V17 | `TestEsquemasDeHerramienta/los dos esquemas son las partes del documento de --describe` (`internal/cli`) deja de pasar si la anotación va en `--describe` y no en el esquema de la herramienta | prototipo: `go test ./internal/cli/` |
| V18 | El arnés de `TestSkillsDelRepositorio` supone que toda skill declara referencias: `metadataDeKitlegal` escribe siempre `kitlegal-referencias:` y `probarRegenerarDosVeces` escribe `references/antigua.md` sin crear la carpeta | `internal/app/skills_test.go:850-858` y `:1046`; prototipo: 13 subtests de `frontmatter/jurisprudencia` y `regenerar-dos-veces/jurisprudencia` en rojo, y en verde con las dos líneas cambiadas |
| V19 | Una skill con `kitlegal-applets` y sin `kitlegal-referencias` ni carpeta `references/` pasa la comprobación de producción | `internal/skills/frontmatter.go` (`defectosDeLasReferencias`: sin declaradas no lista nada) y `sincronia.go` (`referenciasGeneradas`); prototipo: `TestSkillsDelRepositorio` en verde tras V18, con `SKILL.md` de 195 líneas |
| V20 | `TestEvalsDelRepositorio/linea-sin-consulta` recorre todas las skills de `skills/` y exige en cada una la línea `⚠ SIN CONSULTA AL BOE:` | `internal/evals/conjunto_test.go:1577-1604`; prototipo: en rojo con la skill nueva |
| V21 | Lo que rompe registrar el applet y empotrar la skill: la lista de «Tests y guiones existentes que cambian» de `plan.md` | prototipo: `go test -count=1 ./...` con el applet en los dos registros y la skill en `skills/` |
| V22 | El applet `cita` y su dominio pasan la garantía sin red tal como está escrita | prototipo: `internal/arch_test.go` con `internal/core/cita` e `internal/core/ids` entre los orígenes y `app/cita.go` en `ficherosSinRed`; `TestArquitectura` en verde, también R1 |
| V23 | `os.Stdin` no está entre los símbolos que veta `forbidigo`, y `internal/app` ya lo nombra | `.golangci.yml:258-290` (solo `os.Exit`, `os.Stdout`, `os.Stderr`); `internal/app/mcp.go` (`DependenciasDeMCPDelSistema`) |
| V24 | Una llamada de herramienta construye su `Despacho` con un literal de tres campos y no pasa por `Despachar` | `internal/app/herramientas.go`, `llamadaDeHerramienta.resolver`: `Despacho{Destino: DestinoApplet, Applet: l.applet, Args: invocacion}` |
| V25 | `--timeout` es el plazo de toda la operación: si vence, el desenlace es `fuente-no-disponible` aunque el verbo devuelva un resultado | `internal/app/main.go`, `conPlazoAgotado`; `internal/cli/globales.go` (`default:"30s"`) |
| V26 | Los umbrales de las respuestas solo existen hoy con juez, y los grupos de respuestas por modo se componen sin él | `internal/evals/umbrales.go` (`umbralesDelInforme`, `umbralDeSinActivar`) e `informe.go` (`respuestasQueSeJuzgan`) |
| V27 | Una sesión guarda, de cada orden de Bash que nombra `kitlegal` y de cada llamada a una herramienta del registro, la orden y su salida, también cuando falla | `internal/evals/sesion.go` (`Texto`, `leerOrdenDeBash`, `anotarLaSalida`) |
| V28 | La invocación por la que cuenta una llamada lleva sus banderas como `--<nombre>=<valor>` | `internal/cli/herramienta.go` (`LineaDeLlamada`); `internal/evals/juzgar.go` (`argumentosDeLaLlamada`) |
| V29 | El peor caso de un trabajo es 485 s + tandas × (22 + 240 + 10) s, con tandas = Σ ⌈sesiones del grupo / concurrencia⌉ y la prueba de red contada en el modo orden | `internal/evals/definicion.go` (`peorCasoDelTrabajo`, `peorCaso`) |
| V30 | El informe final busca un control `ci:<ruta>:<Test>` con `git grep -qE "^(func <Test>\(\|<Test>:)" HEAD -- <ruta>`, uno `ci:<ruta>` con `git cat-file -e HEAD:<ruta>`, y uno `evals:<skill>:<nombre>` por `.umbrales[].nombre` del informe de esa skill; una celda admite varios | `scripts/workflow/informe.sh` (`estado_control`); `scripts/workflow/comun.sh` (`controles_de_celda`, `ids_de_linea`) |
| V31 | Los guiones de aceptación se activan como `internal/app/testdata/script/h23-<nombre>.txtar` | `scripts/workflow/aceptacion.sh` (`activar`: `prefijo`) |
| V32 | El cierre no enumera las skills: lee los trabajos `evals (<skill>)` que haya | `grep` de `evals (`, `boe-legislacion` y `legal-core` en `scripts/workflow/*.sh`: ninguna lista |
| V33 | La dirección del buscador del Tribunal Constitucional no está escrita en el repositorio | `grep -rn tribunalconstitucional` en `*.md`, `*.yaml`, `*.go`, `*.ts`, `*.astro`: sin resultados fuera de este feature |
| V34 | `make ci` lee las evals de **toda** carpeta de `evals/`, no solo las de las dos skills de hoy, y da por mal formado el fichero que su esquema no admite | `internal/evals/conjunto_test.go`, `TestFormatoDeLasEvalsDeCadaSkill` («el subtest formato de TestEvalsDelRepositorio lee las evals de cada skill») y `malFormadosDeCadaSkill`; `Makefile`, `skills-check` |
| V35 | En H21, las evals que usaban una clave nueva no las escribió la primera tarea: entraron «con el formato que las admite» | `specs/015-h21-kitlegal-mcp-serve/tasks.md`, T001 |
| V36 | Con el applet en los dos registros y la skill en `skills/`, el empaquetado de la extensión y del plugin, el paquete del servidor y la instalación de skills siguen en verde sin tocarlos | prototipo, `go test ./...`: `ok` en `internal/empaquetado`, `internal/mcp`, `internal/mcp/mcptest`, `internal/core/instalacion`, `internal/disco`, `internal/skills` y el paquete raíz |
| V37 | Preparar la sesión de una eval no pide ninguna respuesta grabada para un comando que no consulta una fuente: los de `territorio` y la comprobación del grafo ya son así | `internal/evals/consultas.go`, `ConsultasNecesarias` (ramas `formaTerritorio` y `formaComprobacion`) |
| V38 | Kong v1.16.1 deja en `nil` el campo puntero del argumento que no se escribe y lo rellena cuando se escribe, también con valor vacío, sea una bandera o el argumento de posición, y sea `*string` o `*cli.Literal` | `kong@v1.16.1/model.go:364` (`Value.Parse` crea el puntero antes de decodificar y marca `Set`, `model.go:271`) y `mapper.go:754` y `:776` (`ptrMapper`); prototipo, con los campos de los dos verbos como punteros: las invocaciones de V40 |
| V39 | Con los campos como punteros, `--describe` sigue dando cada argumento como cadena, con la misma anotación `x-banderas`, y la tabla de comandos no cambia; una llamada de herramienta lo sigue recibiendo como cadena y lo escribe `--<nombre>=<valor>`, también vacío | prototipo: `schemas/cita.json` regenerado con `-actualizar-esquemas` y comparado con el anterior con `diff`: cuatro líneas, las cuatro de `cobertura` (D4), y ninguna de `entrada`; `-regenerar-skills` deja `SKILL.md` idéntico (`cmp`); las llamadas de V40 |
| V40 | Un argumento escrito con valor vacío termina con 2 y la clase `argumentos`, como orden y como llamada de herramienta | prototipo, como orden: `cita preparar ECLI:ES:TS:2023:3144 --texto ""`, `… --roj ""`, `… --fecha ""`, `--roj ""`, `--roj=`, `""`, `--texto ""`, `--resolucion "" --fecha 2023-07-04` y `--resolucion 1088/2023 --fecha ""`; y `cita cotejar --documento ""`, `--documento=`, `--roj ""` y `""`, las cuatro con el fragmento en la entrada estándar. Como llamada, guion con la orden `mcp`: `cita_preparar` con `{"ecli":"ECLI:ES:TS:2023:3144","texto":""}`, con `{"ecli":"ECLI:ES:TS:2023:3144","roj":""}`, con `{"ecli":""}` y con `{"texto":""}`, y `cita_cotejar` con `{"documento":""}` y con `{"roj":"","documento":<la ficha>}`: las seis, `error argumentos`, y el servidor sigue (`salida 0`) |
| V41 | La ficha sola da el mismo cotejo que el fragmento entero: los mismos `data` y `hash`, y otra `url`, la de su texto | prototipo: las once primeras líneas del fragmento (316 bytes) con `--roj "STS 1088/2023"`, 949 bytes, como el fragmento entero; y la ficha por la entrada estándar y en `documento` de una llamada, el mismo sobre salvo `fecha_consulta` (`cmp`) |
| V42 | Una llamada con resultado a cada una de las dos herramientas da el sobre de su orden, byte a byte salvo `fecha_consulta`, también cuando sus argumentos son banderas en la orden; y ninguna cambia la caché ni el grafo | prototipo, guion de un solo uso con la orden `mcp` y ocho llamadas en un proceso del servidor: `cita_cotejar` con `{"roj":"STS 1088/2023","documento":<la ficha>}`, con `{"roj":"STS 1088/2023"}`, con `{"roj":"STS 1088/2023","documento":""}` y con `{"documento":<la ficha>}`, y `cita_preparar` con `{"resolucion":"1088/2023","fecha":"2023-07-04"}`, con `{"texto":"cláusula suelo"}`, con `{"roj":"STS 1088/2023"}` y con `{"ecli":"ECLI:ES:TS:2023:3144"}` → `resultado` la primera, la cuarta y las cuatro de `cita_preparar`, y `error argumentos` la segunda y la tercera, `salida 0`; cada sobre `mcp-<n>.json` con resultado, pasado por el `sed` de V10 y comparado con `cmp` con el de su orden (`cita cotejar --roj 'STS 1088/2023' --json` y `cita cotejar --json` con la ficha por `stdin`; `cita preparar --resolucion 1088/2023 --fecha 2023-07-04 --json`, `--texto 'cláusula suelo'`, `--roj 'STS 1088/2023'` y `ECLI:ES:TS:2023:3144`): sin diferencias; en los de las llamadas, las tres casillas del número con su fecha, la `direccion` codificada del texto y la casilla «Nº ROJ»; `arbol cache` antes de las llamadas y después de todo, el mismo listado (`cmp`), con `cache.db` y `world.db` ya escritos por una lectura de `boe articulo`. Mutante, puesto y retirado: con `LineaDeLlamada` sin escribir la propiedad `fecha`, la llamada con `resolucion` y `fecha` da `error argumentos` y el guion falla en la comparación de la salida de `mcp` |

## M · Medido

**M1.** Con el prototipo corregido (V1, V38), el sobre entero en la salida estándar, con su salto final y con
`fecha_consulta` de seis decimales; cada cero en que acaben los microsegundos es un byte menos:

| Invocación | Bytes |
|---|---|
| `cita preparar ECLI:ES:TS:2023:3144 --json` (con equivalente) | 474 |
| `cita preparar --roj "STS 3144/2023" --json` | 470 |
| `cita preparar --resolucion 1088/2023 --fecha 2023-07-04 --json` (tres casillas) | 572 |
| `cita preparar --texto "cláusula suelo" --json` | 389 |
| `cita preparar ECLI:ES:TC:2024:79 --json` (fuera de cobertura) | 424 |
| `cita preparar` sin equivalente: `--roj "STC 79/2024"`, `--roj "SAP M 1234/2020"`, `ECLI:ES:AN:2019:1.2.A` | 404, 412, 422 |
| `cita cotejar --json`, sin referencia, con el fragmento o con su ficha sola | 559 |
| `cita cotejar` con el documento pedido, por sus tres formas | 628 a 652 |
| `cita cotejar` con hallazgo: otra fecha, ROJ cruzado, número cruzado | 869, 949, 1 005 |
| Un error de argumentos | 294 a 379 en el prototipo; 294 a 418 en el binario del hito, donde los tres que da el applet empiezan por «argumentos inválidos:» (contracts/applet-cita.md §6 y §10) |
| `--describe` de `preparar` y de `cotejar` | 5 385 y 6 472 |
| La tabla mínima de `cotejar` con hallazgo, sin `--json` | 1 345 |

- **M2**. El borrador de `SKILL.md` con su tabla generada: 195 líneas y 13 648 bytes; su `description`, por debajo de
  los 1 024 caracteres que admite `TestSkillsDelRepositorio` (V19). El borrador de `contracts/`, copiado al prototipo,
  pasa `TestSkillsDelRepositorio` y `TestOrdenesDeLasSkillsEmpotradas`, y regenerar su tabla lo deja igual (V39).
- **M3**. La ficha del fragmento, de la línea `Roj:` a la de `Tipo de Resolución:`: 11 líneas y 316 bytes
  (`head -n 11 … | wc -c`); el fragmento, 39 líneas y 2 353 bytes (V12). El documento entero de esa sentencia es un
  PDF de 175 955 bytes (`evidencias/adr-0036/manifiesto.json`) que no está en el repositorio: el tamaño de su texto
  no se ha medido.
- **M4**. Lo que rompe (V21): 7 tests de Go por el applet (`TestPuntoDeEntrada`, `TestRegistroDeProduccion`,
  `TestRegistroDeE2E`, `TestHerramientasDelServidor`, `TestEsquemasPublicados`, `TestEsquemasCubrenTodosLosVerbos`,
  `TestSuperficieDeIds`), 2 por la anotación y la tabla (`TestTablaDeComandosCoincideConLaGramatica`,
  `TestEsquemasDeHerramienta`), 2 por la skill (`TestSkillsDelRepositorio`, `TestEvalsDelRepositorio`), y 21 guiones
  e2e (4 por el applet, 17 por la skill). Es la medida sin la etiqueta `integration`: con ella, que `make ci` ejecuta,
  la skill rompe además dos guiones de la instalación de desarrollo, `instalar` e `instalar-sin-gobin`, y los guiones
  son 23 (tasks.md, «Diecinueve guiones, y no diecisiete»).
- **M5**, cálculo y no medida: el peor caso del trabajo de `jurisprudencia` con concurrencia 1 es
  485 + (⌈37/1⌉ + ⌈36/1⌉) × 272 = 20 341 s, 339,0 minutos (V29: 6 evals × 2 modelos × 3 repeticiones = 36 sesiones por
  modo, más la prueba de red contada en el modo orden; ninguna sin binario ni servidor). Cabe en los 352 del trabajo.

## S · Supuestos no verificados

| # | Supuesto | Qué pasa si no se cumple |
|---|---|---|
| S1 | La dirección del buscador del Tribunal Constitucional es `https://hj.tribunalconstitucional.es/`. No está en el repositorio (V33) y esta sesión no tiene red | La skill daría una dirección que no es. La eval (e) no lo detecta: compara esa misma cadena. La persona la comprueba al leer el informe final; cambiarla son `SKILL.md` y la eval 05 |
| S2 | Con el borrador de `SKILL.md`, el modelo que decide activa la skill, pide las operaciones y escribe la línea y la cita con su forma. Esta sesión no abre sesiones con modelo | Lo mide el job de cierre; un fallo ahí es de la skill y lo repara el cierre sin tocar los umbrales (FR-064) |
| S3 | En las sesiones del modo orden, Bash acepta la ficha por la entrada estándar con `<<'DOCUMENTO'`, y el modelo copia sus once líneas sin cambiarlas | La orden tiene además `--documento`; con una línea de menos, `cita cotejar` nombra el dato que falta y la skill repite el cotejo. Lo mide el job de cierre con las evals (d) y (f) |
| S4 | El trabajo `evals (jurisprudencia)` termina muy por debajo de su peor caso. No hay medida | El tope de 352 minutos lo cubre (M5) |
| S5 | La ficha de todo documento del CENDOJ lleva las ocho etiquetas de la del fragmento. El repositorio solo tiene una, del Tribunal Supremo (spec, Assumptions) | `cita cotejar` daría «falta el dato …» y la skill no citaría: el fallo va del lado que no cita |
| S6 | Las cotas del ECLI —órgano de hasta 7 caracteres, número de hasta 25— son las de la rama de lectura, sin contrastar con el estándar europeo (spec, Assumptions) | Un ECLI válido más largo sería un error de argumentos |
| S7 | El código del diseño pasa `golangci-lint` tal como está configurado: el prototipo no se linteó. Riesgos conocidos por la memoria del proyecto: `dupl` con el esqueleto repetido de los dos verbos, `gocyclo` en la lectura de la ficha | Se resuelve en su tarea partiendo funciones, sin `//nolint` |

## D · Decisiones

### El applet y su dominio

**D1 · Paquetes.** El reconocimiento de ECLI y de ROJ va a `internal/core/ids`, donde lo dejó la rama de lectura y
donde `docs/ROADMAP.md` §2 pone los identificadores; la referencia, la consulta preparada, la ficha y el cotejo, a un
paquete nuevo, `internal/core/cita`, que es el que nombra §2; el applet, a `internal/app/cita.go`. *Rechazado*:
`internal/source/cendoj`, donde estaba en la rama: no hay fuente —el applet no consulta nada— y un adaptador de fuente
nuevo es de revisión humana. *Rechazado*: todo en `internal/app`: el dominio no se probaría sin el kernel y no
contaría para la cobertura de `internal/core/**`.

**D2 · La equivalencia vive en `ids`.** `ECLI.ROJ()` y `ROJ.ECLI()` devuelven el equivalente y si se deduce; la
única pareja es la de FR-012. `ROJ.Numero()` da `<número>/<año>`, que el cotejo compara con el número de resolución
(FR-025). *Rechazado*: exportar órgano, año y número de cada tipo y componer fuera: cinco métodos en lugar de tres, y
la regla repartida.

**D3 · Un argumento escrito con valor vacío no es un argumento no dado: es el error de su requisito.** Los
argumentos propios de los dos verbos son campos puntero —`*string`, y `*cli.Literal` el documento—: `nil` es «no se
escribió», y cualquier otro valor, «se escribió», también vacío (V38). El dominio recibe cada uno así y lo comprueba
con su forma en cuanto está dado:

- un ECLI, un ROJ, un número o una fecha vacíos no tienen su forma (FR-006): `cita preparar ""`, `--roj ""`,
  `--resolucion "" --fecha 2023-07-04` y `--resolucion 1088/2023 --fecha ""` terminan con 2;
- el vacío cuenta como una forma dada: `ECLI:ES:TS:2023:3144 --roj ""` es «más de una forma», y
  `ECLI:ES:TS:2023:3144 --fecha ""`, una fecha sin su número (FR-006);
- `--texto ""` solo es «un `--texto` vacío», y junto a una referencia, «una referencia junto a `--texto`» (FR-015);
- `--documento ""` es el texto del documento, vacío: la entrada estándar no se lee (FR-020), y no hay texto (FR-026).

Vale igual en una llamada de herramienta, que escribe cada propiedad que recibe como `--<nombre>=<valor>` (V40): un
agente que rellene con `""` un argumento que no usa recibe el error de argumentos con su mensaje y corrige la
llamada, y la skill le dice que dé solo los que use (D17). Nada cambia en `--describe`, en el esquema de las
herramientas ni en la tabla de comandos: un puntero a cadena se describe como una cadena (V39). *Rechazado*: dar
por no dado el argumento vacío, que era la decisión anterior de este plan: `ECLI:ES:TS:2023:3144 --texto ""`
terminaba con 0 donde FR-015 pide 2, y `cita cotejar --documento ""` leía la entrada donde FR-020 dice que no se
lee. *Rechazado*: un tipo propio que recuerde si se escribió, o que el kernel entregue al verbo qué valores marcó
el analizador como escritos: el puntero ya lo dice, sin tocar el kernel ni el generador de `--describe`.

**D4 · Las claves de `data` de `preparar`.** `referencia` o `texto`, `equivalente` si se deduce, `cobertura` si la
referencia queda fuera de cobertura, `direccion` si hay algo que abrir y `casillas`, siempre una lista. Lo que no
aplica **no está**: ni cadena vacía ni `null`. Así «`data` MUST NOT llevar equivalente» (FR-012) es literal, y quien
lee no confunde una dirección vacía con una dirección. `cobertura` sigue esa misma regla: solo está con el ECLI de
órgano `TC`, que es la única cobertura que el binario declara (FR-014), con `cendoj` `no-cubierto` y su `motivo`.
En las demás salidas no está, porque el binario no sabe si el CENDOJ tiene lo que se busca: de
`--roj "STC 79/2024"`, de un número con su fecha o de un texto no afirma nada (spec, Edge Cases), y ninguna skill
lee esa clave. *Rechazado*: `cobertura` en toda salida, con `cubierto`, que era la decisión anterior de este plan:
afirmaba cubierta una sentencia del Tribunal Constitucional nombrada por sus siglas. *Rechazado*: todas las claves
siempre, con valores vacíos: un `equivalente` con sus dos campos vacíos es «un ROJ compuesto a medias». Ejemplos
con sus bytes, en contracts/applet-cita.md §3.

**D5 · La casilla de dos partes.** «Fecha resolución» son dos entradas, con `campo` «Desde» y «Hasta». *Rechazado*:
un nombre compuesto («Fecha resolución · Desde»): no es lo que la persona ve.

**D6 · El sobre.** `fuente` es `kitlegal.cita`, del espacio reservado a lo calculado, que es lo que dice que no hubo
consulta (como `kitlegal.territorio`, ADR 0017). `url`: en `preparar`, la dirección que abre la persona, y
`kitlegal:applet/cita` cuando no hay ninguna —fuera de cobertura— y en todo fallo; en `cotejar`,
`kitlegal:documento/sha256:<64 hex>`, la huella SHA-256 del texto recibido, byte a byte, que es donde va «la huella
de su texto» de FR-004 (V12). `hash` sigue siendo la huella de `data`, como en todo applet. `fecha_consulta` no la
declara el applet: la pone el reloj del kernel (V9), y por eso los golden la excluyen (SC-004). *Rechazado*: la
huella del texto dentro de `data`: `data` no lleva nada del documento más que su ficha (FR-022). *Rechazado*: una
fecha fija: no hay fichero de datos que la dé, como en `territorio`.

**D7 · Las claves de `data` de `cotejar`.** `ficha` con sus ocho datos; `correspondencia`, con tres valores
—`se-corresponden`, `no-se-corresponden`, `no-se-deduce`—; `pedida` y `es_la_pedida` solo si se dio una referencia;
y `hallazgos`, siempre una lista, con un elemento como mucho, de clase `documento-distinto`, con `difiere` (cada dato
con lo pedido y lo del documento), `cruce` si lo hay (`numero-de-resolucion`, `numero-del-roj`) y `explicacion`. La
lista sigue a `graph check` (ADR 0023): quien ya lee `hallazgos` allí lee igual aquí. *Rechazado*: un objeto
`hallazgo` suelto: otra forma para lo mismo.

**D8 · Las fechas.** En `data`, `AAAA-MM-DD`, la escritura de `--fecha`, también la de la ficha, que en el documento
va `dd/mm/aaaa`: así lo pedido y lo del documento se comparan y se nombran igual. En el valor de una casilla,
`dd/mm/aaaa`, que es lo que la persona escribe.

**D9 · Errores.** Todo error de argumentos del applet lo devuelve el dominio con la clase `argumentos`
(`schema.ConClase`, como los de `ids`), con un mensaje en español que nombra la entrada y lo que falla; el kernel lo
traduce a 2. El applet firma también el fallo (`kitlegal.cita`, `kitlegal:applet/cita`). Orden: primero la
referencia (FR-006, «antes de hacer nada más»), después el texto. Nada devuelve otra clase: no hay fuente que falle.
En el producto, tres de esos errores no los puede decidir el dominio y los da el applet con `cli.ErrArgumentos`, con
el mismo código y la misma clase: `preparar` sin referencia ni `--texto`, la referencia junto a `--texto`, y
`cotejar` sin ningún texto (contracts/applet-cita.md §6).
Lo demás —una entrada estándar que no se puede leer— es `inesperado`, 1, por la regla genérica.

**D10 · Sin salida legible propia.** Sin `--json` vale la tabla mínima del kernel (spec, Fuera de alcance; M1).

**D11 · El esquema publicado** es `schemas/cita.json`, con las partes `cotejar` y `preparar`, un fichero por applet
como `grafo.json`, `instalacion.json` y `servidor.json`.

### La entrada estándar

**D12 · El texto llega por `--documento` o por la entrada estándar, y la entrada la entrega el kernel.** El
argumento se llama `documento` y es un `cli.Literal` —tras un puntero, como los demás (D3)—, para que el texto dado
por la bandera y el dado por la entrada sean los mismos bytes y den la misma huella (V3). Escrito, también vacío, es
el texto, y la entrada no se lee; solo sin él lee el verbo la entrada. La entrada estándar no es una dependencia del
applet: la raíz de
composición la registra en el registro (`Registro.LeerDe`), el despacho de una orden la pone en `Despacho.Entrada`, y
el kernel se la da al verbo que la lee, por una interfaz sin exportar, `lector`, como la de `servidor`. Una llamada
de herramienta construye su `Despacho` sin ella (V24): el verbo recibe `nil`, no tiene nada que leer y responde con
el error de argumentos de FR-026. Así «la entrada estándar no se lee nunca» es cierto por construcción y sin tocar
`mcp.go`, `herramientas.go` ni `internal/mcp` (V7). *Rechazado*: darle `os.Stdin` al applet como a `mcp`: en el
servidor, esa entrada es el protocolo, y una llamada sin `documento` lo leería. *Rechazado*: un lector compartido
que recuerde si el servidor ya leyó: estado oculto entre dos applets. *Rechazado*: una marca en la llamada: es
cambiar el camino de las herramientas, que FR-030 deja como está.

**D13 · `--timeout` vale lo que en cualquier verbo.** El plazo es el de toda la operación (V25), también la lectura
de la entrada: quien escriba el documento a mano en una terminal durante más de 30 s recibe `fuente-no-disponible`,
y puede dar `--timeout` o un fichero. Ningún mecanismo nuevo.

### La tabla de comandos

**D14 · `--describe` dice qué argumentos propios son banderas, y la tabla las escribe como tales.** Los dos verbos
son los primeros con banderas propias que entran en la tabla de una skill, y la tabla las escribiría como argumentos
de posición (V14): el modelo, en el modo orden, pediría una orden que no existe. `--describe` gana en `entrada` una
anotación, `x-banderas`, con los nombres de las banderas propias del verbo, **solo cuando las tiene**: el documento
de un verbo sin ellas no cambia un byte, ni la tabla de `boe-legislacion` ni la de `legal-core` (FR-049). La tabla
escribe cada una detrás de los argumentos de posición, `[--<nombre> <nombre>]`, o `[--<nombre>]` si es booleana
(V15). La anotación no va en el esquema de entrada de la herramienta: describe cómo se escribe la orden, y una
herramienta recibe un objeto. *Consecuencias medidas*: `schemas/instalacion.json` gana la anotación en sus tres
partes, porque los verbos de `skills` ya tenían banderas (V16); y dos tests pasan a decir «sin las banderas globales
ni la anotación» (V16, V17). *Rechazado*: la anotación también en el esquema de la herramienta: evita tocar esos dos
tests, pero mete una clave desconocida en lo que lee cada cliente MCP, sin haber podido probar ninguno.
*Rechazado*: que la tabla deduzca las banderas: `--describe` no distingue un argumento de posición opcional
(`graph check [<norma>]`) de una bandera. *Rechazado*: marcar los de posición: cambia los siete esquemas
publicados.

### La skill

**D15 · `jurisprudencia` no tiene `references/`.** El hito no le da datos de `data/`, y la comprobación de producción
lo admite (V19). Lo que lo impedía era el arnés del test, que suponía referencias en toda skill (V18): se corrige el
arnés, en dos líneas. *Rechazado*: declararle una referencia que ya existe para que el arnés pase: la skill llevaría
un fichero que no usa.

**D16 · `linea-sin-consulta` comprueba las skills que llevan esa regla.** La subprueba recorre hoy todas las de
`skills/` (V20). `jurisprudencia` no consulta el BOE y, sin herramienta ni binario, lleva su propia línea
(spec, Edge Cases); pasa a recorrer `skillsConLineaSinConsulta`, las dos que la llevan, con la misma exigencia.
*Rechazado*: poner en `jurisprudencia` la línea `⚠ SIN CONSULTA AL BOE:`: diría de una sentencia que no se consultó
el BOE.

**D17 · La redacción de `SKILL.md`** es contracts/SKILL-jurisprudencia.prototipo.md, con su tabla generada (V15, M2).
Sigue las lecciones de H7.2 a H24: cada forma fija, con su ejemplo en un bloque; cada prohibición, con su razón; la
respuesta habla a la persona y no nombra los campos de lo que devuelve una orden (regla 3); las órdenes, siempre con
`--json`. La dirección del Tribunal Constitucional está en un solo sitio de la skill, el paso 7 (S1). **El paso 3
manda pasar a `cita cotejar` la ficha y nada más** —sus líneas, de la de `Roj:` a la de `Tipo de Resolución:`—, que
es lo que dicen la entrada del hito y Edge Cases del spec («pasa la ficha a `cita cotejar`») y lo único que la
operación lee (V41): con la herramienta, `documento` lo escribe el modelo, y el texto entero haría crecer cada
cotejo con el tamaño de la sentencia —2 353 bytes ya en el fragmento, frente a los 316 de su ficha (M3)—. Si el
error de argumentos nombra un dato que sí está en el documento, la skill repite el cotejo con la ficha completa. Y
la cabecera del protocolo dice que una herramienta recibe solo los argumentos que se usan (D3). *Rechazado*: «mejor
el texto entero, tal como llegó», que era la redacción anterior: no da nada que la ficha no dé.

### Las evals

**D18 · El formato gana dos formas de comando y una clave, `sentencias`.** Los comandos de `cita` son los primeros
con banderas: `comandos` admite `{applet: cita, verbo: preparar}` y `{applet: cita, verbo: cotejar}`, cada uno con
`roj` opcional, y `preparar` con `con_texto: true`. Es lo que piden las seis evals y nada más: ninguna exige un ECLI,
un número o una fecha como argumento. Lo que se espera de la respuesta va en `sentencias`: `citas`, `ninguna_cita`,
`sin_cita_del_roj`, `no_comprobada`, `direcciones`, `casillas` y `direccion_de_busqueda`. Con `sentencias`,
`comandos` deja de ser obligatorio —la eval (e) no exige ninguno— y nada cambia para una eval sin ella.
*Rechazado*: una lista genérica `contiene` de textos: no diría qué es cada uno ni dejaría comprobar que una
dirección es la que devolvió la sesión. *Rechazado*: campos nuevos en cada resultado de `informe.json`: lo que falta
ya lo dicen sus motivos, y el informe de las otras dos skills no cambia un byte.

**D19 · Qué es una cita, una línea y un ECLI, para quien compara sin modelo.** La cita es el corchete
`[<ECLI>, ROJ: <ROJ>]`, exacto: una coma y un espacio, `ROJ:` y un espacio. La línea usa las tolerancias de las
formas fijas que ya hay (`patronDeEtiqueta`, al principio de una línea). El ECLI, la regla de FR-061. La salida de
una operación es **cada línea de la salida de una orden o de una llamada que es un sobre de kitlegal**: así no cuenta
lo que una orden encadenada lea de otro sitio (FR-061), y sí cuentan dos órdenes en una misma línea de Bash. Una
orden sin `--json` no da un sobre, y lo que escriba no cuenta: la skill las pide siempre con `--json`. *Rechazado*:
toda la salida de la orden: `cat SKILL.md; kitlegal …` dejaría pasar el ECLI del ejemplo.

**D20 · Qué da umbrales de respuestas a una skill.** Hoy, tener juez. Pasa a ser: tener juez **o** que alguna de sus
evals declare `sentencias`. `sin_activar` se publica en los dos casos; las clases, con juez; `cita_sin_documento`,
con `sentencias`. `legal-core` no tiene ni lo uno ni lo otro y sigue con `[]`; `boe-legislacion` sigue con sus doce.
Ningún nombre de skill en el código. *Rechazado*: una declaración aparte por skill, o una variable del job: otro
fichero u otro parámetro que hay que mantener al lado de las evals, que ya lo dicen. *Rechazado*: publicar
`cita_sin_documento` en toda skill: cambia los umbrales de `boe-legislacion` (FR-063).

**D21 · El motivo del umbral nombra las respuestas.** Detrás de su medida y su condición, cada sesión que cuenta con
los ECLI que la hacen contar, como el de una clase del juez nombra sus frases. Quien lee el informe, y la reparación
del cierre, necesitan saber qué respuesta es.

**D22 · El trabajo del job.** `jurisprudencia` entra en la matriz con concurrencia 1 y objetivo 0, como `legal-core`:
no hay medida de la que sacar otra cifra, y los tres trabajos corren a la vez contra la misma suscripción. Su peor
caso cabe en el tope de hoy (M5). *Rechazado*: concurrencia 4: acorta el cierre, y lleva a nueve las sesiones
simultáneas sin haber medido el límite de ritmo.

**D23 · Las preguntas de (d), (e) y (f)** y los seis ficheros, literales, en contracts/evals-jurisprudencia.md §5. El
texto pegado va en un bloque literal de YAML, que conserva sus bytes porque el fragmento no tiene nada que un bloque
literal altere (V13). **Los seis entran con la tarea del formato, no con la primera**: `make ci` lee las evals de
toda carpeta de `evals/` (V34), y seis ficheros con claves que el esquema aún no admite lo dejarían en rojo. Es lo
que hizo H21 (V35). Se escriben antes que la skill, que llega dos tareas después. *Rechazado*: que la primera tarea
toque también el esquema y el lector: dejaría de ser una tarea sin código de producto.

### Los controles

**D24 · La aceptación e2e** son cuatro guiones (plan.md). Toman el fragmento del repositorio por
`$KITLEGAL_SKILLS/../evidencias/…` (V8), sin copiarlo; la llamada de herramienta, que necesita el texto dentro de un
JSON, lleva la ficha —las once primeras líneas del fragmento, que es lo que la skill manda pasar (D17)—, y el guion
comprueba con `cmp` que son las del fragmento (V41). La orden y la llamada se comparan sin `fecha_consulta` (V10),
y se compara una llamada con resultado de cada una de las dos herramientas, no solo de `cita_cotejar`: las que la
skill hace en modo herramienta en las seis evals y la del ECLI, de modo que cada argumento que en la orden es una
bandera viaja en alguna (V42). Son los primeros verbos cuyos argumentos de herramienta son banderas, y sin esas
llamadas el camino con éxito de `cita_preparar` como herramienta solo lo vería el job del cierre.
Los dos guiones de los verbos llevan además los casos de D3 que un campo de cadena dejaría pasar: una referencia
junto a `--texto ""`, y `--documento ""` con el fragmento en la entrada (V40). *Rechazado*: una variable nueva en el
arnés: no hace falta.

**D25 · Que nada cambia en la caché ni en el grafo lo comprueba un guion** (V11), con `arbol`, que ya da la huella de
cada fichero. *Rechazado*: un test de Go que abra `cache.db` y `world.db`: R3 lo impide fuera de sus paquetes.

**D26 · Los golden** viven en `internal/app/testdata/cita/`, uno por caso, con el sobre entero y `fecha_consulta`
puesta a `0001-01-01T00:00:00Z`; el test pone ese valor en lo que da el verbo y compara los bytes.

**D27 · La garantía sin red** se hace cumplir donde ya está: `internal/arch_test.go` suma `internal/core/cita` e
`internal/core/ids` a sus orígenes y `app/cita.go` a `ficherosSinRed` (V22). *Rechazado*: un test aparte: sería una
segunda copia del recorrido.

**D28 · Fuzz.** `FuzzECLI` y `FuzzROJ` en `internal/core/ids`, `FuzzLeerFicha` en `internal/core/cita`, cada uno con
su corpus en `testdata/fuzz/` de su paquete, como `FuzzCodigoINE`.

### Alcance y datos

**D29 · Datos externos: ninguno.** El applet no pide nada y las evals no necesitan grabaciones (FR-054). El único
dato de fuera es el fragmento de `evidencias/adr-0036/`, que está en `main` y ninguna tarea escribe. El CENDOJ no
tiene fila que permita grabar nada, y nada se graba.

**D30 · Lo que el hito no toca.** `cmd/empaquetar`, `internal/empaquetado`, `internal/mcp`, `internal/app/mcp.go`,
`internal/app/herramientas.go`, `internal/httpx`, `scripts/workflow/`, `evidencias/`, `skills/boe-legislacion/`,
`skills/legal-core/`, `docs/SOURCES.md`, `CLAUDE.md`, `docs/ROADMAP.md` y `web/`.
