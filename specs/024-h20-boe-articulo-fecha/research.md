# Research: H20 · `boe articulo --fecha` + `boe-legislacion` v0.2

Fase 0 del plan, en modo desatendido. Las decisiones siguen el «Criterio de decisión autónoma» de la constitución y
cada una lleva su alternativa rechazada. Cuatro bloques:

- **V**: lo que se afirma del repositorio y de las dependencias, con el fichero y la línea, o la orden, donde se
  comprobó en esta sesión.
- **M**: lo medido en esta sesión. El prototipo se hizo en un `git worktree` desacoplado de `0b2a21d` bajo `$TMPDIR`,
  fuera del repositorio, con un test temporal y una copia de `SKILL.md`; ninguna tarea parte de él, y se retiró con
  `git worktree remove --force` antes de terminar.
- **S**: supuestos que no se pudieron comprobar sin red o sin el binario del hito.
- **D**: decisiones.

## V · Comprobado en el repositorio y en las dependencias

| # | Afirmación | Dónde |
|---|---|---|
| V1 | Go `1.27.0` con `toolchain go1.27.2`; `go version` da `go1.27.2 darwin/arm64` | `go.mod:3,5`; `go version` |
| V2 | Las dependencias que toca el hito ya están: `invopop/jsonschema v0.14.0`, `santhosh-tekuri/jsonschema/v6 v6.0.3`, `modelcontextprotocol/go-sdk v1.8.0` | `go.mod:9-12` |
| V3 | La nota [5]: «La versión vigente es la última del bloque: su fecha es fecha_publicacion, aunque venga vacía, u «original» si falta el atributo, y fecha_vigencia e id_norma quedan vacías si faltan; un bloque sin versiones da el texto de todo el bloque, la fecha «original» y ni fecha de vigencia ni norma modificadora» | `internal/source/boe/doc.go:26` |
| V4 | El lector del bloque solo guarda la última versión: al abrir cada `<version>` sustituye sus atributos y vacía su texto | `internal/source/boe/bloque.go:55` (`bloqueLeido`), `:223-226` (`abrir`) |
| V5 | `articulo` valida la norma y el bloque antes de abrir nada y resuelve dentro de `invocar` con `articuloDelBloque`: entrada, petición del bloque, metadatos, `componerArticulo`, `guardar` | `internal/source/boe/articulo.go:37-56`, `:234-267`, `:311-326` |
| V6 | La entrada de hoy: clave `boe.legislacion-consolidada\|1\|articulo\|<dirección del bloque>`, datos `Articulo` ya compuesto, vigencia de 7 días; la unión de tipos de las entradas | `internal/source/boe/entradas.go:62-64` (`datosDeEntrada`), `:98-100` (`claveDelArticulo`), `:108-111` (`String`); `fuente.go:31` (`vigenciaLarga`) |
| V7 | Con `--offline` y sin entrada, «fuente no disponible» con la clave en el mensaje | `internal/source/boe/fuente.go:505-519` (`resolverSinPedir`), `:565-568` (`ausenteConOffline`) |
| V8 | La caché no purga: una entrada se sobrescribe por su clave | `internal/cache/entradas.go:25-32` |
| V9 | `world.db` vive en el directorio de la caché, el de `KITLEGAL_CACHE_DIR` si está | `internal/graph/ruta.go:46-70`; `internal/cache/cliente.go:18` |
| V10 | `argumentosDeArticulo` solo tiene `Norma` y `Bloque`; `cita` ya usa `Fecha *string` con `placeholder:"AAAA-MM-DD"` | `internal/app/boe.go:184-197`; `internal/app/cita.go:94` |
| V11 | El nombre de una definición de `--describe` es `<paquete>.<Tipo>`; la de `data` de `articulo` es `boe.Articulo` | `internal/cli/describe.go:345-378`; `schemas/bloque.json` (`$defs.articulo.$defs`, leído con `jq`) |
| V12 | `invopop/jsonschema` hereda los campos de un struct embebido sin etiqueta y no exige los que llevan `omitempty` | `$(go list -m -f '{{.Dir}}' github.com/invopop/jsonschema)/reflect.go:4, 512-515, 952, 1061-1065` |
| V13 | La tabla de comandos de una skill nombra cada bandera propia como `[--<nombre> <nombre>]` y, en «Qué devuelve», todas las propiedades de la definición en su orden | `internal/skills/comandos.go:147-157`; `skills/jurisprudencia/SKILL.md:182` |
| V14 | Cada herramienta MCP sale del verbo; un campo sin `arg` se pasa como `--<nombre>=<valor>`, delante de `--` y de los argumentos de posición | `internal/app/herramientas.go:134`; `internal/cli/herramienta.go:32, 118` |
| V15 | `boe articulo` emite `Norma`, `Bloque`, `BloqueVersion` (`<eli>#<bloque>@<fecha_vigencia>:<hash_texto>`), `eli:has_part`, `eli:has_version` y el `Texto` | `internal/source/boe/grafo.go:55-77` |
| V16 | Una lectura del grafo es cada arista `eli:has_version` de un lote | `internal/core/grafo/lecturas.go:31-41` |
| V17 | La entrega fusiona cada registro con lo guardado y escribe el fusionado si cambia; la fila de lecturas se calcula antes de escribir, y sin fila parte de las `BloqueVersion` que el bloque enlaza por `eli:has_version` | `internal/graph/aplicar.go:274-293, 302-328, 348-369, 413-439, 474-475` |
| V18 | Al fusionar un nodo, sus datos y su vigencia siguen a su última observación | `internal/core/grafo/observacion.go:97-113` |
| V39 | La salida legible de `boe` es la tabla mínima del kernel, con `data` aplanado por claves | `internal/render/tabla.go:44, 115`; `internal/cli/sobre.go:125` |
| V19 | `fuente-caducada` mira toda `Norma`, todo `Bloque` y la `BloqueVersion` vista de cada bloque; `version-obsoleta` mira solo las filas de lecturas | `internal/core/grafo/comprobar.go:232-258, 285-323` |
| V20 | `cli.LoteDe` es el único sitio que construye un lote con procedencia | `internal/cli/entrega.go:25-36` |
| V21 | `schema.Observado` es `{Vigencia, Operaciones}` y su valor cero no emite nada | `internal/core/schema/grafo.go:10-19` |
| V22 | `ComandoEsperado` no tiene fecha; la forma bloque es `applet`, `norma` y `bloque` sin verbo | `internal/evals/formato.go:242-277, 370` |
| V23 | Hoy un comando de bloque casa con `articulo` o `articulos` cuyo primer argumento es la norma y entre cuyos argumentos está el bloque, con código 0 | `internal/evals/juzgar.go:1003, 1038-1056` |
| V24 | La llamada `boe_articulo` con una bandera se juzga con argumentos `--<bandera>=<valor>, <norma>, <bloque>`: el primero no es la norma | `internal/evals/juzgar.go:571-600` |
| V25 | El código de una llamada es 0 si tiene resultado y no es error, y el de su `clase` si lo es | `CONTRIBUTING.md:781-783`; `internal/evals/juzgar.go:565-584` |
| V26 | La gramática de las formas fijas (`⚠`, la etiqueta sin distinguir mayúsculas y con énfasis tolerado, y `:`) y su reutilización en `⚠ PLAZO:` | `internal/evals/avisos.go:47-95`; `internal/evals/plazos.go:18-48` |
| V27 | Los umbrales de las respuestas de cada modo se montan en `respuestasMedidas.umbrales`: `sin_activar`, las clases del juez y, si alguna eval lo pide, los de sentencias, cotejo, plazo y cobertura | `internal/evals/umbrales.go:349-372`; `informe.go:1446-1474` |
| V28 | El patrón de un umbral sin modelo con 0 que decide: `umbralDePlazoSinFecha` y su medida en `medirLosPlazos` | `internal/evals/umbrales.go:559-577`; `informe.go:1558-1581` |
| V29 | Las reglas del conjunto: entre 10 y 21 evals, exactamente 10 positivas que deciden, cada positiva con una norma que no cita otra | `internal/evals/conjunto.go:222-234, 420-487, 519-549` |
| V30 | `comando-bloque` de `schemas/eval.yaml.json` vale para `comandos` y para `grafo_previo.comandos`; el esquema tiene ya `dia` (AAAA-MM-DD) y `fecha` (AAAAMMDD) | `schemas/eval.yaml.json:26, 49, 184-193, 299-300` |
| V31 | La caché de las evals se prepara con `ConsultasNecesarias`, y `Preparar` y `ComprobarSinRed` toman por falta toda consulta que no termina con 0 | `internal/evals/consultas.go:90-131`; `preparar.go:147-196, 419-423` |
| V32 | `TestDefinicionDelJob` compara con las evals del repositorio el literal `peorCasoDeBoeLegislacion`, hoy de 17 329 s con los grupos 97, 96 y 6 y 54, 54 y 3 respuestas; H9 cambió su literal de `legal-core` por lo mismo | `internal/evals/definicion_test.go:1036-1160, 1866-1877, 1933-1953`; `specs/023-h9-internal-core-plazos/plan.md:244-245` |
| V33 | `TestUmbralesDeLasTresSkills` monta el informe de `boe-legislacion` con sus evals del repositorio y espera los trece de H26 | `internal/evals/informe_test.go:6546-6623` |
| V34 | La prosa de `SKILL.md` de `boe-legislacion` no puede usar «redacción posterior», «vigente hasta», «fecha_vigencia», «norma_modificadora»…, ni una fecha de ocho cifras; cada orden con `&&` va seguida de su forma para PowerShell | `evals/boe-legislacion/expresiones-prohibidas.yaml`; `internal/evals/conjunto_test.go:2559-2596, 3040-3108` |
| V35 | La `description` de una skill admite como mucho 1024 caracteres; la de `boe-legislacion` tiene hoy 1018 | `internal/skills/frontmatter.go:53, 438`; M8 |
| V36 | El arnés e2e copia lo grabado de H4 en `reproduccion/`, da `KITLEGAL_T0_BIN` con el reloj fijo y la orden `mcp`; la suite se activa como `internal/app/testdata/script/h20-<nombre>.txtar` | `internal/app/e2e_test.go:56, 147-158, 552-581, 660, 2010`; `scripts/workflow/aceptacion.sh:74-89` |
| V37 | La fila de `boe.legislacion-consolidada` en `docs/SOURCES.md` de `main` está revisada (2026-09-13) | `docs/SOURCES.md:42` |
| V38 | `make schema-check` es `TestEsquemasPublicados`; `make skills-check`, los siete tests de skills, evals y normas | `Makefile:121-127` |

## M · Medido en esta sesión

| # | Qué | Resultado |
|---|---|---|
| M1 | Las versiones de lo grabado, leídas con el `leerBloque` de `main` dejando en el XML una versión cada vez (test temporal en el prototipo) | `a1-30`: 2 versiones, `20171109`/`20180309`/`BOE-A-2017-12902` (texto de 1 688 bytes) y `20200205`/`20200206`/`BOE-A-2020-1651` (1 805). `a22`: 9, la quinta `20031217`/`20040101`/`BOE-A-2003-23103`, la sexta `20070529`/`20070701`/`BOE-A-2007-10701`, la séptima `20071031`/`20080430`/`BOE-A-2007-18874`, la novena `20151031`/`20151031`/`BOE-A-2015-11723`; textos de 2 415 a 4 969 bytes. `a21`: 1, `20151002`/`20161002`/`BOE-A-2015-10565` (3 611). La última de cada uno da el texto y la huella de su golden de hoy |
| M2 | El `data` de la lectura fechada del `a1-30` el 2019-06-01, compuesto con el golden de hoy y la primera versión | 2 226 bytes; los dos campos nuevos, 119 bytes (contracts/articulo-fecha.md §4) |
| M3 | Los datos de la entrada de las versiones (D8) de los tres bloques | `a1-30`, 3 843 bytes; `a22`, 37 561; `a21`, 3 856 |
| M4 | Cuántos bloques grabados tienen alguna versión sin `fecha_vigencia` o ninguna versión (`jq` sobre las 31 grabaciones de `texto/bloque` de las tres raíces de `UnionDeGrabaciones`) | ninguno; solo `a9999`, que es el 404 |
| M5 | `time.Parse(time.DateOnly, …)` con go1.27.2, en un programa fuera del repositorio | acepta `2019-06-01`, `9999-12-31` y `0000-01-01`; rechaza `2019-02-30`, `2019-6-1`, `+201-06-01`, `-201-06-01`, `2019-06-01T00:00:00Z`, `20190601` y la cadena vacía. `time.Parse("20060102", …)` lee `20180309` y rechaza la vacía |
| M6 | El prototipo de `SKILL.md` (contracts/SKILL-boe-legislacion.prototipo.md, con la fila de `main` en la tabla) contra los tests de `make skills-check` en el árbol del prototipo | 298 líneas; los siete tests en verde (`internal/app`, `internal/skills`, `internal/evals`) |
| M7 | Un guion e2e suelto: `go test -count=1 -run '^TestEntregaDelHito$/^boe-codigos$' ./internal/app/` | `ok`, 2,7 s |
| M8 | La `description` de hoy, con las líneas unidas como las une `>-` | 1 018 caracteres |
| M9 | `go build -o "$TMPDIR/…/kitlegal" ./internal/app/ejemplo/kitlegal-e2e` | se construye |
| M10 | Las órdenes de quickstart.md §1 a §6, en `main` | todas `ok` (las de tests que aún no existen, `[no tests to run]`); `make skills-check` y `make schema-check`, `ok`; `go run ./cmd/kitlegal boe articulo BOE-A-2015-10565 a21 --describe` da el esquema con `boe.Articulo` y sin `fecha`, y el árbol de trabajo queda igual |
| M11 | `go run ./cmd/kitlegal boe articulo --help` | sale bien y lista las banderas globales, sin `--fecha`: la precondición de los guiones falla hoy por su aserción |

## S · Supuestos no verificados

- **S1**: el peor caso del trabajo de `boe-legislacion` con veintitrés evals a cinco a la vez, 18 977 s, sale de las
  fórmulas de `definicion.go` (V32) con 109, 108 y 6 sesiones y 60, 60 y 3 respuestas; no se ha ejecutado
  `TestDefinicionDelJob` con las evals nuevas, que no existen aún. Lo comprueba ese test en `make ci` (FR-050).
- **S2**: lo que respondan las sesiones del modelo que decide; lo mide el job de evals de cierre (SC-001).
- **S3**: qué fecha de consulta pone el binario de pruebas `KITLEGAL_T0_BIN` en un sobre de fallo sin procedencia
  (código 2, o 4 con `--offline`); los guiones la igualan con `exec sed` donde la comparan (precedente:
  `h9-plazos-herramienta.txtar`).

## D · Decisiones

**D1 · La lectura fechada vive en el adaptador `internal/source/boe`.** Elegir la versión es leer y elegir dentro de
la respuesta del BOE, como hoy la nota [5] elige la última (V3, V4): el mismo recurso, el mismo lector y la misma
caché. Va a un fichero nuevo, `fecha.go`, junto a `articulo.go`. *Rechazada*: un paquete de dominio en `internal/core`
para la elección. Ningún otro adaptador ni ninguna otra skill lo usa (YAGNI, principio V), y separaría del lector
del bloque lo que solo tiene sentido con él.

**D2 · `--fecha` es un `*string`.** `argumentosDeArticulo` gana `Fecha *string` con `placeholder:"AAAA-MM-DD"`, como
`cita` (V10), y `boe.ConsultaArticulo` gana `Fecha *string`. Sin la bandera es `nil` y la lectura es la de hoy, por
el mismo camino de código que hoy (FR-012). *Rechazada*: `string`, con el que `--fecha ""` sería una lectura sin
fecha y no el código 2 que pide FR-005.

**D3 · La fecha se valida con `time.Parse(time.DateOnly, …)` y nada más.** `ValidarFecha` entra en
`internal/source/boe/ids.go`, detrás de `ValidarNorma` y `ValidarBloque` y antes de abrir nada (nota [28]); su fallo
es `errorDeArgumentos`, código 2, con el mensaje de contracts/articulo-fecha.md §3. `time.Parse` ya rechaza toda forma
que no es AAAA-MM-DD con sus ceros y todo día que no existe (M5). *Rechazadas*: reutilizar el lector de fechas de
`internal/core/plazos` (ataría el adaptador del BOE a un dominio de plazos para una línea), o una expresión regular
delante de `time.Parse` (no rechaza nada más, M5).

**D4 · La regla de elección.** El lector del bloque guarda todas las versiones, con los mismos campos que hoy da la
última y su texto normalizado igual (D9 de abajo no lo cambia). Con la fecha pedida *D*:

1. si el bloque no tiene versiones, o alguna no tiene una `fecha_vigencia` AAAAMMDD, la lectura falla con 3 y el
   mensaje de FR-004 (D7), sea cual sea *D*;
2. si no, la elegida es la última, en el orden del bloque, cuya `fecha_vigencia` no es posterior a *D*, comparadas como
   días del calendario; si no hay ninguna, 3 con el mensaje de FR-003;
3. la redacción posterior es la última versión del bloque cuando no es la elegida.

Comparar días y no textos es FR-002; recorrer en orden y quedarse con la última que cumple da, el día en que entra en
vigor una versión, esa versión. *Rechazada*: comparar las cadenas AAAAMMDD, que daría lo mismo con fechas bien
formadas pero dejaría la comparación al formato y no al calendario.

**D5 · La salida: un tipo nuevo para `articulo`, con los campos de hoy embebidos.** `boe.ArticuloFechado` es `Articulo`
embebido más `FechaPedida string` (`fecha_pedida`, `omitempty`) y `RedaccionPosterior *RedaccionPosterior`
(`redaccion_posterior`, `omitempty`), y `RedaccionPosterior` es `{FechaVigencia, NormaModificadora}`. Es la `Salida`
que el registro declara para `articulo`; la lectura sin fecha sigue devolviendo `Articulo`, por el camino de hoy, y su
JSON es el de hoy (FR-012). El esquema de `articulo` gana las dos propiedades, opcionales (V12), y su definición pasa
a llamarse `boe.ArticuloFechado` (V11); el de `articulos` no cambia (FR-070). *Rechazadas*: añadir los dos campos a
`Articulo`, que cambiaría el `--describe`, la herramienta y la fila de la tabla de `articulos`, que nunca los da
(FR-070); y un caso especial en el nombrador de definiciones del kernel para conservar el nombre `boe.Articulo`, que
sería un caso por tipo en `internal/cli`.

**D6 · Los nombres.** `fecha_pedida` lleva la fecha tal como se pidió, AAAA-MM-DD. `redaccion_posterior` lleva
`fecha_vigencia` y `norma_modificadora`, con los nombres y la forma AAAAMMDD de los mismos campos del artículo, de la
misma respuesta (FR-010). Sin redacción posterior, la clave no está (FR-010, «nada que diga que la hay»). La salida
legible, la tabla mínima del kernel, las lleva como filas `fecha_pedida` y `redaccion_posterior.fecha_vigencia` y
`.norma_modificadora` sin código nuevo (V39).

**D7 · Las filas y los mensajes del 3.** Los dos casos de 3 nuevos usan `errorDeNoEncontrado` con la dirección del
bloque y el instante de la consulta, y su mensaje, que la skill lee, los distingue entre sí y del 3 de un bloque que
no está en la norma («la norma … no tiene el bloque …»):

- FR-003: `el bloque <bloque> de <norma> no tenía redacción vigente el <AAAA-MM-DD>: su primera redacción tiene
  fecha de vigencia <AAAAMMDD>`;
- FR-004: `el BOE no da la fecha de vigencia de cada redacción del bloque <bloque> de <norma>: no se puede saber cuál
  tenía el <AAAA-MM-DD>`.

FR-004 va con 3 porque lo pedido, la redacción de esa fecha, no está en la fuente de una forma que se pueda situar:
es la fila «no encontrado» del ADR 0023. *Rechazadas*: 4, que dice que la fuente no responde o no está en caché y
empujaría a reintentar lo que no cambia; y 1, que es un defecto del programa o del entorno.

**D8 · La caché: una entrada nueva con las versiones del bloque.** La lectura fechada guarda, con la clave
`boe.legislacion-consolidada|1|versiones|<dirección del bloque>` y la vigencia de 7 días del artículo, el título, el
tipo, todas las versiones, los avisos y el ELI (contracts/articulo-fecha.md §5), y elige la versión desde ahí. La
escribe solo con el bloque y sus metadatos obtenidos e interpretados —un fallo de la fuente o al interpretar no deja
nada escrito, como hoy (FR-007; `TestFallosNoSeGuardan`)— y antes de elegir, de modo que la misma fecha, o cualquier otra, se sirve después con `--offline` con el mismo
`data` y la misma fecha de consulta (FR-022). No lee ni escribe la entrada de hoy (FR-020, FR-021), que no cambia de
forma ni de clave, ni la versión de las entradas (V6). Los avisos y el ELI se guardan con las versiones por lo mismo
que la nota [2] los guarda con el artículo. *Rechazadas*: una entrada por norma, bloque y fecha, que crece con cada
fecha distinta y no serviría con `--offline` la segunda lectura de un año o un mes (FR-045) ni la caché de las evals
(D18); y que la lectura sin fecha guarde también las versiones, que cambiaría lo que hace hoy (FR-012, «Se queda»).

**D9 · El grafo: solo la `BloqueVersion` y su texto, y una entrega que no toca lo que ya guarda.** La lectura fechada
emite el nodo `BloqueVersion` de la versión elegida, con el id y los datos de V15, y su `Texto`, y nada más: ni la
`Norma`, ni el `Bloque`, ni las dos aristas. Sin `eli:has_version` no es una lectura (V16), así que la fila de lecturas
no cambia y la siguiente lectura sin fecha parte de lo mismo (V17). Pero si el nodo ya está en `world.db` —una lectura
sin fecha posterior a la última vigencia, o el único bloque de una versión—, fusionarlo le daría la última
observación, la vigencia y los datos de la lectura fechada (V18): apagaría su `fuente-caducada` y, en un bloque sin
fila de lecturas, lo haría su redacción vista (V17, V19), y la siguiente lectura sin fecha daría un `version-obsoleta`
falso. Por eso el resultado lo marca: `schema.Observado` gana `SoloLoNuevo bool`, `cli.LoteDe` lo copia a
`core.Lote.SoloLoNuevo`, `grafo.Consolidar` a `Consolidado.SoloLoNuevo`, y la entrega escribe solo los registros que
`world.db` no tiene y deja como están los que ya tiene. El contrato `Applet` no cambia (ADR 0005): es un campo más del
`Resultado`, como `Grafo` lo fue en H7. *Rechazadas*: emitir sin marca (falla en los dos casos de arriba, que alcanza
cualquier lectura fechada de un bloque ya leído); emitir con `eli:has_version` y excluirla de `Lecturas` (la arista
sigue en lo que miran la redacción vista y el ámbito de `graph check`); una relación nueva (exige el `Bloque`, que es
siempre candidato de `fuente-caducada`); y basar la redacción vista y la caducidad en la historia de la arista
(cambia el cálculo de H7.1 entero).

**D10 · `--no-graph` y `--dry-run`** no necesitan nada nuevo: el kernel entrega a `graph.Nulo` o no entrega (FR-032),
y bajo ensayo la fuente no emite, como hoy (`articulo.go:105-115`).

**D11 · El formato de las evals.** `ComandoEsperado` gana `Fecha` (`fecha`, AAAA-MM-DD) en la forma bloque, y el
esquema la admite solo en `comandos`, con una definición nueva, `comando-bloque-esperado`; `grafo_previo` sigue con
`comando-bloque`, porque nadie siembra el grafo con una lectura fechada y su lectura no la ejecutaría. `Eval` gana
`redaccion_fechada`, con `norma`, `bloque`, `fecha` y, o bien `fecha_vigencia` y, si la hay, `posterior`
(`fecha_vigencia` y `norma_modificadora`), o bien `vigente_desde` (contracts/evals-redaccion-fechada.md §1). Lleva
norma, bloque y fecha propios, como `redacciones_modificadas`: la línea que se busca lleva la cita del bloque, y la
eval no depende de cuál de sus comandos es el fechado. El esquema la prohíbe donde prohíbe lo esperado de una
respuesta: en una eval que no activa la skill y en la eval sin binario ni servidor. *Rechazada*: sacar la fecha y el
bloque del comando con fecha de la eval, que obliga a una regla más para el caso de varios.

**D12 · Qué casa con un comando de bloque.** Los argumentos de una invocación de `articulo` se leen separando la
bandera `--fecha` (con `=` o con su valor detrás, en cualquier sitio) de los de posición. Un comando sin fecha casa
con `articulo` o `articulos` de la norma con el bloque entre sus bloques, **sin** `--fecha`, y código 0, como hoy; uno
con fecha casa solo con `articulo` de la norma y el bloque **con esa fecha** y código 0 o 3: con fecha, el 3 de FR-003
es la respuesta del binario sobre esa fecha, no un fallo de la fuente (FR-051). *Rechazada*: exigir 0 también con
fecha, con la que la eval (b), cuya lectura esperada sale con 3 (US2), no se cumpliría nunca.

**D13 · Las formas fijas.** Tres, con la gramática de las etiquetas (V26), cada una en una línea con la cita del
bloque:

```text
⚠ REDACCIÓN A FECHA: <forma legible> [<identificador>, bloque <id>]: se cita la redacción vigente el AAAA-MM-DD, la de fecha de vigencia AAAAMMDD.
⚠ REDACCIÓN DE HOY: <forma legible> [<identificador>, bloque <id>]: hoy la redacción es otra, la de <identificador>, con fecha de vigencia AAAAMMDD.
⚠ SIN REDACCIÓN VIGENTE: <forma legible> [<identificador>, bloque <id>]: el AAAA-MM-DD no tenía redacción vigente; la tiene desde el AAAAMMDD, la fecha de vigencia de su primera redacción.
```

La cita ata cada línea a su bloque, como en `⚠ REDACCIÓN MODIFICADA:`; la fecha pedida va como se pidió y las de
vigencia como las da el binario, así que un guion las distingue por su forma. «REDACCIÓN DE HOY» dice lo que pide
FR-041 —la redacción de hoy del bloque, sin decir que rija ese día— y evita «redacción posterior», que la prosa de la
skill no puede usar (V34). *Rechazadas*: una sola línea para la redacción citada y la de hoy (FR-041 pide dos);
líneas sin cita (no se podrían atar a un bloque con dos lecturas en la respuesta); y etiquetas sin `⚠`, que obligarían
a una gramática nueva.

**D14 · El umbral: `redaccion_sin_fechar:<modelo>:<modo>`.** Por cada modo, sobre las respuestas del `total` de
`sin_activar` de ese modo en las evals con `redaccion_fechada`; cuenta la sesión que no ejecutó la lectura fechada de
la eval (D12) o cuya respuesta, para el bloque de la eval, no lleva exactamente las líneas esperadas
(contracts/evals-redaccion-fechada.md §3); 0, `decide: true`. Va detrás de las clases del juez de su modo y de los
demás umbrales de las respuestas, como `documento_sin_cotejar` y `plazo_sin_fecha` (V27): `boe-legislacion` pasa a
quince, once que deciden (FR-056). La medida se calcula en el informe, junto a la de los cotejos, y no entra en el
juicio de cada sesión: la serie de 2 de 3 sigue con sus comandos, sus citas y sus avisos (FR-054, FR-057). El nombre
sigue a los de hoy, `<lo medido>_sin_<lo que falta>`. *Rechazadas*: medirlo en el juicio de la sesión y sumarlo a la
serie, que cambiaría la regla de 2 de 3 que el spec deja como está; y un nombre por modelo sin modo, que no es el
contrato de FR-054.

**D15 · Las evals con fecha son una clase propia en las reglas del conjunto.** Las reglas de H5 cuentan diez positivas
que deciden, cada una con una norma que no cita otra (V29). Las dos evals con fecha citan la LCSP y la LPAC, que ya
citan la 02 y la 01: contadas como positivas, cuatro positivas se quedarían sin norma propia. Una eval con
`redaccion_fechada` mide la lectura a una fecha de una norma que otra positiva ya cubre, como la eval sin binario ni
servidor no cita ninguna: las reglas que cuentan positivas y materias no la miran, y el máximo pasa de 21 a 23. El
mínimo, las diez positivas, las materias distintas, las normas del hito y el art. 21 no cambian; las normas conocidas
y el esquema, sí la miran. *Rechazada*: subir a doce las positivas y quitar la regla de materias distintas, que
dejaría de medir lo que H5 fijó.

**D16 · `TestDefinicionDelJob` cambia su literal, no su configuración.** El peor caso de `boe-legislacion` con las
veintitrés evals del repositorio pasa a 18 977 s (S1); el literal `peorCasoDeBoeLegislacion` y el caso de
`probarElPeorCasoDelTrabajo` que lo lleva cambian con él, como cambió el de `legal-core` en H9 (V32). La concurrencia
(5), el objetivo de 900 s, el tope de 352 minutos, la definición del contrato y `.github/workflows/evals.yml` no
cambian (FR-059). Es el control de FR-050. Por la misma razón, `TestPlanEnDosModos` (`plan_test.go`,
`probarElPlanDelRepositorio`), que fija con las evals del repositorio las series y sesiones del plan de cada skill,
cambia solo las de `boe-legislacion`: de 32, 32 y 2 series y 96, 96 y 6 sesiones a 36, 36 y 2 y 108, 108 y 6, con los
modelos y las repeticiones del job, iguales.

**D17 · La skill no pide `graph check` tras una lectura fechada.** Una lectura fechada no es lectura del bloque
(FR-031), así que `graph check` no diría nada de ella, y lo que dijera de las lecturas sin fecha no va en esa
respuesta (FR-046). Si la respuesta dice además qué dice hoy el bloque, lo lee sin fecha con su `graph check` detrás,
como siempre (FR-042).

**D18 · La caché preparada de las evals sirve toda lectura fechada.** `ConsultasNecesarias` añade, por cada consulta
`articulo <norma> <bloque>` que anota, la misma con `--fecha 9999-12-31`, la última fecha que admite `--fecha` (M5):
posterior a toda vigencia, sale con 0 en todo bloque de lo grabado (M4) y deja guardada la entrada de sus versiones
(D8), de la que se sirve después la lectura del bloque en cualquier fecha, también las dos de un año o un mes. El
comando fechado de una eval anota además el bloque sin fecha, como hoy, por si la respuesta dice qué dice hoy
(FR-042). `Preparar` y `ComprobarSinRed` siguen tomando por falta todo lo que no sale con 0 (V31). *Rechazada*:
preparar la fecha de cada eval, que en la eval (b) sale con 3 y sería una falta, y no serviría otra fecha.

**D19 · Los golden de FR-081 y FR-085.** FR-081 lo fija `TestArticuloEnUnaFecha` (`internal/source/boe/fecha_test.go`),
una tabla con los seis casos del spec y el de US1, cada uno con su código, su clase, su versión —`fecha_version`,
`fecha_vigencia`, `norma_modificadora`—, su redacción posterior o su ausencia y su mensaje escritos a mano en el test, y
la huella de cada texto igual a la de su texto; el del 2019-06-01 del `a22`, con el `texto` y el `hash_texto` del golden
de hoy, `testdata/golden/articulo-BOE-A-1985-5392-a22.json` (FR-013). El golden de `--describe` (FR-085) es
`schemas/bloque.json`: su parte `articulo` es el `--describe` de `boe articulo` con su `$id`, y `make schema-check` lo
compara byte a byte (V38). *Rechazadas*: un fichero golden por caso generado con el propio código (fijaría lo que el
código da, no lo que el spec pide) y un segundo golden de `--describe` que repetiría `schemas/bloque.json`.

**D20 · La suite de aceptación.** Seis guiones testscript, con el arnés de hoy (V36) y sin órdenes nuevas (plan.md,
«Aceptación e2e»). La salida de `main` de FR-083 se escribe dentro del guion, como sección del txtar, desde la salida
del binario sin el hito en la primera tarea; las fechas de consulta que no fija el reloj del arnés se igualan con
`exec sed` (S3). Cada guion empieza por una precondición que hoy falla por su aserción: que `boe articulo --help`
nombre `--fecha`.

**D21 · Los documentos.** `CHANGELOG.md` (*Unreleased*), `README.md`, `CONTRIBUTING.md` y `docs/WORKFLOW.md`
cambian donde el hito deja algo falso o incompleto (contracts/documentos.md). De `docs/WORKFLOW.md`, solo la frase de
`medir_cierre` que enumera los umbrales que deciden en `boe-legislacion` (`docs/WORKFLOW.md:139`): pasa de nueve a once,
con `redaccion_sin_fechar:<modelo>:<modo>` en cada modo, sin reescribir lo demás (FR-062). El comentario de paquete de `internal/source/boe/doc.go` gana un párrafo sobre la lectura fechada, que
`refs/boe.py` no tiene; la tabla de notas no cambia (su test la fija entrada a entrada).

**D22 · Ningún dato externo.** Lo grabado de H4 ya tiene los tres bloques con todas sus versiones y sus metadatos
(M1), y la lectura fechada pide los mismos recursos que la de hoy (FR-006). No hay manifiesto, test de grabación ni
evidencia nuevos, y `grabar_datos` no tiene nada que grabar.
