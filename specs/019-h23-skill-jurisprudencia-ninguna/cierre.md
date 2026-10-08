# H23 · Cierre de la Definition of Done

Cierre de T012 (FR-031, FR-049, FR-055, FR-064, FR-065, FR-088; SC-009) sobre `19f332f`, que es T011, el último commit
antes de T012. La rama está al día con `main`: `git merge-base main HEAD` da la cabeza de `main`, `8a85cab`, así que
`git diff --name-status main` da exactamente lo que cambia el hito. Las cinco partes son las de la tarea: lo creado y
lo modificado, la cobertura, el quickstart, los controles de umbral, y los umbrales, los supuestos y los atajos.

Todo lo que se cita aquí se ejecutó en la sesión de T012, en primer plano, y se vio terminar, o está en el
repositorio. Lo que no se pudo medir se dice tal cual (§2, «El diff»; §3, escenarios 10 y 11; §4, las cinco filas
`evals:`; §6).

**T012 no toca código ni añade ningún test.** Las cifras de cobertura quedan sobre sus umbrales (§2), y la tarea solo
pide tests si alguna queda debajo. Sus únicos cambios son este fichero y su marca en tasks.md. Las cinco partes
cuadran con lo que la tarea pide comprobar, y las cuarenta órdenes de quickstart.md §0 a §9 dan lo esperado (§3).

## 1. Lo creado y lo modificado en el hito

`git diff --name-status main`, antes de escribir este fichero, da **176 ficheros: 112 `A` y 64 `M`**, ninguna `D` ni
`R` (`git diff --numstat main`: 16 484 líneas añadidas y 576 quitadas, ningún binario). 42 están en
`specs/019-h23-skill-jurisprudencia-ninguna/`, todos `A`; este `cierre.md` será el 43. Los otros 134 son 70 `A` y 64
`M` (11 032 líneas añadidas y 576 quitadas). `git ls-files --others --exclude-standard` no da ningún fichero sin
seguimiento.

### Los esquemas, los golden, los corpus de fuzz, los guiones, las evals y la skill

Son 78 de esos 134, más los cuatro guiones de la suite congelada, que están en el directorio del feature. Cada uno con
la tarea que lo tocó (`git log --format=%s main..HEAD -- <ruta>`) y sus líneas (`git diff --numstat main`).

**Los esquemas** (3):

| Estado | Fichero | Tarea | Líneas | Qué es |
|---|---|---|---|---|
| A | `schemas/cita.json` | T006 `[datos]` | +584 | El esquema del applet `cita`, con sus dos verbos, de `--describe`. |
| M | `schemas/instalacion.json` | T005 `[datos]` | +16 −3 | La anotación `x-banderas` en sus tres partes. |
| M | `schemas/eval.yaml.json` | T007 `[datos]` | +67 −5 | `sentencias` y las dos formas de comando de `cita` en el formato de eval. |

**Los golden** (11, todos `A`, de T006 `[datos]`, una línea cada uno), en `internal/app/testdata/cita/`:
`preparar-ecli.json`, `preparar-roj.json`, `preparar-resolucion.json`, `preparar-texto.json`,
`cotejar-sin-referencia.json`, `cotejar-ecli.json`, `cotejar-roj.json`, `cotejar-resolucion.json`,
`cotejar-roj-cruzado.json`, `cotejar-resolucion-cruzada.json` y `cotejar-otra-fecha.json`.

**Los corpus de fuzz** (34, todos `A`, dos líneas cada uno):

- `internal/core/ids/testdata/fuzz/FuzzECLI/` (11, T002 `[datos]`): `blanco-delante`, `blanco-detras`, `cadena-vacia`,
  `mal-formado`, `minusculas`, `numero-con-punto`, `otro-organo`, `otro-pais`, `sentencia-conocida`,
  `termina-en-letra` y `tribunal-constitucional`.
- `internal/core/ids/testdata/fuzz/FuzzROJ/` (12, T002 `[datos]`): `anio-de-dos-cifras`, `blanco-detras`,
  `cadena-vacia`, `dos-espacios`, `minusculas`, `numero-con-letras`, `numero-de-26-cifras`, `otras-siglas`,
  `prefijo-roj`, `sentencia-conocida`, `siglas-de-dos-palabras` y `sin-anio`.
- `internal/core/cita/testdata/fuzz/FuzzLeerFicha/` (11, T003 `[datos]`): `blancos-en-los-extremos`, `cadena-vacia`,
  `dos-fichas`, `fecha-imposible`, `ficha-sola`, `finales-crlf`, `fragmento`, `otro-organo`, `sin-ficha`,
  `sin-ponente` y `texto-delante`.

**Los guiones** (23 modificados y 4 nuevos):

- Por el applet, cuatro, todos `M` y de T006 `[datos]`, en `internal/app/testdata/script/`: `argumentos.txtar`
  (+2 −2), `h21-mcp-herramientas.txtar` (+10 −6), `h21-mcp-proceso.txtar` (+1 −1) y `h21-mcp-protocolo.txtar` (+3 −3).
- Por la skill, diecinueve, todos `M` y de T009 `[datos]`. Diecisiete en `internal/app/testdata/script/`:
  `skills-salida-legible.txtar` (+19 −4), `skills-host-antigravity.txtar` (+4 −4) y los `h19-skills-`
  `ambito-dir` (+8 −7), `ambito-global` (+8 −6), `aviso` (+4 −4), `aviso-sin-aviso` (+2 −2), `conflictos-dentro`
  (+1 −1), `conflictos-entradas` (+16 −11), `conflictos-rutas` (+3 −3), `doctor-copia` (+18 −13), `doctor-hallazgos`
  (+20 −17), `dry-run` (+9 −9), `idempotencia` (+14 −12), `install-hosts` (+14 −8), `install-local` (+23 −13),
  `list-doctor` (+1 −1) y `no-empotrada` (+9 −9). Y dos en `internal/skills/testdata/script/`: `instalar.txtar`
  (+7 −4) e `instalar-sin-gobin.txtar` (+3 −1).
- La suite de aceptación, cuatro, todos `A` y de T001 `[aceptacion]`, en
  `specs/019-h23-skill-jurisprudencia-ninguna/aceptacion/`: `cita-preparar.txtar` (+97), `cita-cotejar.txtar` (+101),
  `cita-herramientas.txtar` (+147) y `cita-sin-efectos.txtar` (+72). Ningún commit posterior a T001 los toca, y
  `shasum -a 256` de cada uno da la huella que guarda `gates/aceptacion-congelada.json`: `1246494b…c8815`,
  `1680962b…008ad`, `f0ae5a36…a394e` y `bc3885fb…251c7`. En `internal/app/testdata/script/` no hay hoy ningún
  `h23-`: los activa el workflow tras el bucle (§3 y §4).

**Las evals** (6, todas `A`, de T007), en `evals/jurisprudencia/`: `01-existe-con-numero-y-fecha.yaml` (+18),
`02-resumen-sin-documento.yaml` (+18), `03-por-materia.yaml` (+12), `04-documento-pegado.yaml` (+54),
`05-tribunal-constitucional.yaml` (+9) y `06-documento-que-no-es-el-pedido.yaml` (+64).

**La skill** (1, `A`, de T009 `[datos]`): `skills/jurisprudencia/SKILL.md`, 195 líneas. Es el único fichero de su
carpeta: sin `references/` ni `scripts/`.

Todo lo que el diff nombra bajo un `testdata/` o bajo `schemas/` son esos 71 ficheros (3 + 11 + 34 + 23), y cada uno
lo tocó solo una tarea `[datos]`: T002, T003, T005, T006, T007 o T009.

### Los otros 56

17 `A` y 39 `M`, con su tarea:

- `internal/core/ids` (7): `ecli.go`, `roj.go` y sus dos `_test.go`, nuevos; `doc.go`, `errores.go` y
  `errores_test.go` (T002).
- `internal/core/cita` (9, el paquete nuevo): `doc.go`, `referencia.go`, `consulta.go`, `ficha.go`, `cotejo.go` y
  los cuatro `_test.go` (T003).
- `internal/app` (12): `cita.go` y `cita_test.go`, nuevos (T006); `despacho.go`, `main.go` y `main_test.go` (T004);
  `registro.go` (T004, T006); `registro_test.go`, `esquemas_test.go` y los dos ficheros de
  `ejemplo/kitlegal-e2e/` (T006); `herramientas_test.go` (T005, T006) y `skills_test.go` (T005, T009).
- `internal/cli` (3): `describe.go`, `describe_test.go` y `herramienta_test.go` (T005).
- `internal/skills` (2): `comandos.go` y `comandos_test.go` (T005).
- `internal/evals` (15): `sentencias.go` y `sentencias_test.go`, nuevos (T007, T008); `formato.go`, `juzgar.go`,
  `consultas.go`, `conjunto.go` y los tests de formato y de consultas (T007); `conjunto_test.go` (T007, T009);
  `umbrales.go`, `informe.go` y sus dos tests (T008); `definicion_test.go` (T010) y `doc.go` (T011).
- `internal/arch_test.go` y `cmd/kitlegal/main_test.go` (T006).
- `.github/workflows/evals.yml` (T010): `jurisprudencia` en la matriz, con `concurrencia: 1` y
  `objetivo_de_duracion: 0`, y comentarios; +18 −5.
- `.golangci.yml` (T002): la entrada `constitucional` de `ignore-rules` de `misspell`, con su comentario; +5 líneas y
  ninguna quitada.
- `README.md`, `docs/JURISPRUDENCIA.md` y `CHANGELOG.md` (T011), y `CONTRIBUTING.md` (T006, T011).

### Lo que no cambia (FR-031, FR-049)

Cada comprobación es un `git diff --name-status main -- <rutas>`, ejecutado en esta sesión, que **no imprime nada**.
Entre paréntesis, los ficheros versionados que hay en esas rutas (`git ls-files`), para que el cero no sea el de una
ruta vacía:

- **Ninguna grabación**: `internal/source` (95), los dos manifiestos `grabaciones.json` con sus dos
  `grabacion_test.go` (4) y el `testdata/` de la raíz (62). Bajo `testdata/` el diff solo nombra los golden, los
  corpus y los guiones de arriba. El paso `grabar_datos` no tenía nada que grabar.
- **Ninguna fuente**: `internal/source`, `internal/httpx`, `internal/cache`, `docs/SOURCES.md` y
  `scripts/verify-sources.sh` (160).
- **Ningún manifiesto**: `'*grabaciones.json' '*manifiesto*' '*manifest*'` (12).
- **`SOURCES.md`**: `docs/SOURCES.md` (1).
- **Ningún ADR**: `docs/ADR` (37). De `docs/`, el diff solo nombra `docs/JURISPRUDENCIA.md`.
- **Ningún spec ni plan de otro hito**: `-- specs ':!specs/019-h23-skill-jurisprudencia-ninguna'` (936).
- **El directorio de los guiones del workflow**: `scripts/workflow` (16); tampoco nada de `scripts` entero (31) ni de
  `.specify` (45).
- **La carpeta de evidencias**: `evidencias` (31), con `evidencias/adr-0036` (3) y `evidencias/adr-0037` (13).
  `git rev-parse main:evidencias` y `HEAD:evidencias` dan el mismo árbol, `7da5b4ac…`, y `git log main..HEAD --
  evidencias` no nombra ningún commit.
- **El paso de empaquetado y su paquete**: `cmd/empaquetar` e `internal/empaquetado` (12), y `.goreleaser.yaml` (1).
- **El paquete del servidor MCP, `mcp.go` y `herramientas.go`**: `internal/mcp` (7), e `internal/app/mcp.go` e
  `internal/app/herramientas.go` (2).
- **El cliente HTTP del kit**: `internal/httpx` (42).
- **Las otras dos skills, byte a byte**: `skills/boe-legislacion` y `skills/legal-core` (5), y sus evals,
  `evals/boe-legislacion` y `evals/legal-core` (31). `git rev-parse` da el mismo árbol en `main` y en la cabeza a cada
  una de las cuatro carpetas: `60633173…`, `01a63a0a…`, `799416c8…` y `20d74b03…`.
- **Lo demás que tasks.md nombra**: `CLAUDE.md`, `docs/ROADMAP.md`, `docs/USO.md` y `web` (69); `data` (26); `go.mod`,
  `go.sum`, `tools` y `Makefile` (13); y `.github` sin `evals.yml` (6).

Las mismas órdenes sí imprimen donde hay cambios: con `-- internal/evals` salen 15 líneas; con `-- internal/core`, 50;
con `-- docs`, 1; y con `-- evals/jurisprudencia`, 6.

### La carpeta de evals de `jurisprudencia` no tiene `juez` (FR-055)

`ls -A evals/jurisprudencia` da los seis `.yaml` y nada más, y `git ls-files evals/jurisprudencia/juez` no da nada. La
única carpeta `juez` del repositorio es `evals/boe-legislacion/juez`. `internal/evals/medida.go` no nombra
`jurisprudencia`, e `internal/evals/medida_test.go`, donde la tabla de las copias del juez lleva la fila de
`boe-legislacion`, no está en el diff. En lo añadido fuera de `specs/`, `adr-0037` aparece una vez: la frase de
`CONTRIBUTING.md` que dice que `jurisprudencia` no tiene juez hasta H25 y por qué. Nada apunta a esa evidencia.

### La huella del fragmento es la de la cabecera de tasks.md

`shasum -a 256 evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt` da
`4886e0c8ca9836527ec08d8732b50315640a76803033374af5287b2a8321ee27`: la de la línea 15 de tasks.md y la de
`evidencias/adr-0036/manifiesto.json`. `wc` da 39 líneas y 2 353 bytes, y sus once primeras líneas, la ficha, 316
bytes. Es también la huella que `cita cotejar` pone en la `url` del sobre cuando recibe el fragmento entero (§3,
escenario 3).

### La Definition of Done, punto a punto (`ROADMAP.md` §1)

| Punto | Aplica | Qué se ve en la cabeza |
|---|---|---|
| 1 · `make ci` en verde | sí | §3, escenario 9: código 0 y `ci: todos los controles en verde`, con la caché de tests vacía. |
| 2 · tests offline; fixtures si toca red | tests sí; fixtures no | Los tests de §3 y §4. El hito no toca la red: ninguna grabación (arriba), y `TestArquitectura` exige que `cita`, `ids` y el fichero del applet no alcancen `net`, `net/http` ni `internal/httpx`. |
| 3 · sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | sí | En las 2 720 líneas no vacías añadidas a los `.go` que no son tests, 0 coincidencias de `"net/http"`, `"net"`, `os.Exit`, `fmt.Print`, `os.Stdout`, `os.Stderr`, `panic(`, `"database/sql"`, `modernc.org` e `internal/httpx`. `os.Stdin` aparece en dos, `registro.LeerDe(os.Stdin)` en `internal/app/registro.go` y en el binario de e2e: la entrada que el kernel entrega al verbo (plan.md, *Complexity Tracking*). `golangci-lint`: `0 issues.` |
| 4 · esquemas y `schema-check` | sí | `schemas/cita.json` e `instalacion.json`, de `--describe`; `make schema-check` en verde, solo (§3, escenario 6) y dentro de `make ci`. `TestCitaPreparar` y `TestCitaCotejar` validan cada salida contra su parte de `schemas/cita.json` (`exigirSalidaDeCita`, en `internal/app/cita_test.go`). |
| 5 · errores con código estable | sí | §3, escenarios 2 y 3: `argumentos` sale con 2 y el hallazgo `documento-distinto`, con 0 y `"ok":true` (ADR 0023). |
| 6 · e2e y `CHANGELOG.md` | sí | Los 23 guiones alineados y los 4 de la suite, que pasan como copias `zz-` (§3); `CHANGELOG.md` +107, en *Unreleased* (T011). |
| 7 · ADR | no | `docs/ADR` sin cambios: la decisión es el ADR 0036, con su enmienda. |
| 8 · `SOURCES.md` | no | `docs/SOURCES.md` sin cambios: ninguna fuente. |
| 9 · cobertura | sí | §2. |
| 10 · skill: evals antes, `SKILL.md` < 300, sin drift | sí | Las evals son de T007 (`5d2ce9a`) y la skill, de T009 (`510a790`); 195 líneas; `make skills-check` en verde (§3, escenario 7). |
| 11 · dimensión territorial | no | — |
| 12 · operaciones de grafo | no | `cita` no observa ninguna fuente; la copia `zz-cita-sin-efectos` pasa: ni la caché ni el grafo cambian (§3, escenario 5). |
| 13 · recursos, plazos o escritos | no | — |

## 2. Cobertura (Definition of Done §1.9)

Se mide sobre los dos perfiles que deja el `make ci` de §3, escenario 9, en verde y con la caché de tests vacía
(`go clean -testcache` justo antes; el registro no trae ningún `(cached)`): `coverage.out` (`make test`) y
`coverage-integration.out` (`make test-integration`). El porcentaje sale de `go tool cover -func`: el global, sobre
cada perfil; el de cada árbol o fichero, sobre el perfil filtrado a sus líneas, con su cabecera `mode:`, escrito en el
directorio temporal. Los recuentos de sentencias salen de los bloques del perfil, contando cada bloque una vez.
«Unión» son los dos perfiles fundidos bloque a bloque, como hace Codecov con los dos que publica la CI.

| Árbol | `coverage.out` | `coverage-integration.out` | Unión | Umbral (`codecov.yml`) |
|---|---|---|---|---|
| Global | **97,2 %** (14 589 de 15 015 sentencias) | **97,6 %** (14 651 de 15 015) | 97,6 % | ≥ 70 % |
| `internal/core/**` (dominio) | **98,8 %** (2 874 de 2 910) | **98,8 %** (2 874 de 2 910) | 98,8 % | ≥ 85 % |
| `internal/core/ids` (ampliado) | **100 %** (160 de 160) | **100 %** (160 de 160) | 100 % | — |
| `internal/core/cita` (nuevo) | **100 %** (233 de 233) | **100 %** (233 de 233) | 100 % | — |
| `internal/cli/**` | 98,5 % (540 de 548) | 98,5 % (540 de 548) | 98,5 % | ≥ 90 % |
| `internal/app` | 94,8 % (1 297 de 1 368) | 94,8 % (1 297 de 1 368) | 94,8 % | — |
| `internal/app/cita.go` | 96,0 % (48 de 50) | 96,0 % (48 de 50) | 96,0 % | — |
| `internal/evals` | 97,7 % (5 077 de 5 196) | 98,3 % (5 108 de 5 196) | 98,3 % | — |
| `internal/evals/sentencias.go` | 99,4 % (162 de 163) | 99,4 % (162 de 163) | 99,4 % | — |

**Los umbrales se cumplen y no hace falta ningún test**: T012 no toca ningún `_test.go`.

**Los dos paquetes del dominio, fichero a fichero.** Dan lo mismo en los dos perfiles, y ninguna función queda bajo
el 100 %:

| Fichero | Sentencias | Funciones |
|---|---|---|
| `internal/core/ids/ecli.go` | 57 de 57 | 12 |
| `internal/core/ids/roj.go` | 39 de 39 | 6 |
| `internal/core/cita/referencia.go` | 54 de 54 | 10 |
| `internal/core/cita/consulta.go` | 51 de 51 | 6 |
| `internal/core/cita/ficha.go` | 72 de 72 | 9 |
| `internal/core/cita/cotejo.go` | 56 de 56 | 5 |

`internal/core/cita/doc.go` no tiene sentencias. El dominio gana 329 sentencias, las 233 de `cita` y las 96 de
`ecli.go` y `roj.go`, todas cubiertas: pasa de las 2 545 de 2 581 que anotó el cierre de H24 a 2 874 de 2 910, con
36 sin cubrir, como entonces.

**Lo que no llega al 100 % fuera del dominio**, en los dos ficheros nuevos que la tarea declara. Son tres sentencias,
las mismas en los dos perfiles:

- `internal/app/cita.go:209` y `:253` (`Ejecutar`, 94,4 %, y `texto`, 87,5 %): `io.ReadAll` de la entrada estándar
  falla, y el verbo devuelve ese error. Es la entrada que no se puede leer, que tasks.md, «Proporcionalidad», deja a
  la regla genérica.
- `internal/evals/sentencias.go:537` (`valorDeLaBandera`, 92,3 %): el `break` de una bandera sin valor que es el
  último argumento de la invocación.

Ninguna tiene un test, y la tarea no pide añadirlo: ningún umbral depende de ellas.

**Dos ejecuciones, las mismas cifras.** Con este fichero ya escrito se ejecutó otro `make ci`, también con la caché de
tests vacía: código 0, `ci: todos los controles en verde`, 53 `ok` y ningún `FAIL` ni `(cached)`. Sus dos perfiles dan
las mismas cifras que los de arriba, sentencia a sentencia, en todas las filas de las dos tablas y en las tres
funciones que no llegan al 100 %.

**El perfil de integración repite bloques.** `coverage.out` trae 9 433 líneas de bloque, una por bloque;
`coverage-integration.out`, 14 647 para los mismos 9 433 bloques: catorce paquetes —los de `internal/core`,
`internal/cli` y `data`, entre ellos— salen con cada bloque tres veces. En la segunda ejecución son 25 199 líneas para
los mismos 9 433 bloques, y entre los paquetes repetidos está también `internal/evals`: cuáles se repiten cambia de
una ejecución a otra. No se ha buscado por qué. No cambia ninguna cifra: los recuentos cuentan cada bloque una vez, y
el porcentaje de `go tool cover -func` coincide con el de los bloques en todas las filas de la tabla.

**El diff.** `codecov/patch` es bloqueante con `target: auto`, la cobertura de la base, y `main` no se midió en esta
sesión, así que no se estima. Lo que diga solo se sabe en la propuesta de cambio.

## 3. Quickstart (§0 a §9)

Los diez escenarios se ejecutaron en su orden, en primer plano, sobre `19f332f` y con el árbol limpio salvo los dos
ficheros de `gates/` que el workflow lleva modificados (`tarea-actual.json` y `tareas-intentos.json`). Son cuarenta
órdenes, y **las cuarenta dan lo esperado**.

### §0 a §3: el binario

Dieciocho de las diecinueve órdenes que llaman al binario escriben además en la salida de error un aviso que no es
de este hito (abajo, «Cómo se ejecutó»); lo que sigue es su salida estándar y su código.

| § | Orden | Esperado | Obtenido | ¿Cuadra? |
|---|---|---|---|---|
| 0 | `go version` | Go 1.27 | `go version go1.27.1 darwin/arm64` | sí |
| 0 | `shasum -a 256 evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt` | `4886e0c8…ee27`, la del manifiesto | `4886e0c8ca9836527ec08d8732b50315640a76803033374af5287b2a8321ee27` | sí |
| 0 | `make build` | `bin/kitlegal` | código 0; `bin/kitlegal`, de 25 326 594 bytes | sí |
| 0 | `export KITLEGAL_CACHE_DIR="$(mktemp -d)"` | — | código 0; un directorio nuevo y vacío en el temporal | sí |
| 1 | `bin/kitlegal cita preparar ECLI:ES:TS:2023:3144 --json` | código 0, `"fuente":"kitlegal.cita"`; `"direccion":"https://www.poderjudicial.es/search/indexAN.jsp"`, la casilla `{"nombre":"ECLI","valor":"ECLI:ES:TS:2023:3144"}` y `"equivalente":{"forma":"roj","valor":"STS 3144/2023"}`; sin `cobertura` | código 0; esa `fuente`, esa `direccion`, esa casilla —la única— y ese `equivalente`, literales; sin `cobertura` | sí |
| 1 | `bin/kitlegal cita preparar --roj "STS 3144/2023" --json` | código 0; una casilla «Nº ROJ» y `"equivalente":{"forma":"ecli","valor":"ECLI:ES:TS:2023:3144"}`; sin `cobertura` | código 0; `{"nombre":"Nº ROJ","valor":"STS 3144/2023"}`, sola, y ese `equivalente`; sin `cobertura` | sí |
| 1 | `bin/kitlegal cita preparar --resolucion 1088/2023 --fecha 2023-07-04 --json` | código 0; tres casillas: «Nº Resolución» con `1088/2023` y «Fecha resolución», «Desde» y «Hasta», con `04/07/2023`; sin `equivalente` ni `cobertura` | código 0; `{"nombre":"Nº Resolución","valor":"1088/2023"}` y las dos de «Fecha resolución», con `"campo":"Desde"` y `"campo":"Hasta"` y `"valor":"04/07/2023"`; sin `equivalente` ni `cobertura` | sí |
| 1 | `bin/kitlegal cita preparar --texto "cláusula suelo" --json` | código 0; `"direccion":"https://www.poderjudicial.es/search/sentencias/cl%C3%A1usula%20suelo/1/AN"` y `"casillas":[]`; sin `cobertura` | código 0; las dos, literales; sin `cobertura` | sí |
| 1 | `bin/kitlegal cita preparar ECLI:ES:TC:2024:79 --json` | código 0; `"cobertura":{"cendoj":"no-cubierto",…}`, sin `direccion` y con `"casillas":[]` | código 0; `"cobertura":{"cendoj":"no-cubierto","motivo":"Las resoluciones del Tribunal Constitucional no están en el CENDOJ: su buscador no las tiene."}`, `data` sin `direccion` y `"casillas":[]` | sí |
| 2 | `bin/kitlegal cita preparar ECLI:ES:TS:2023 --json; echo "código $?"` | `código 2`, `"ok":false`, `"clase":"argumentos"` | los tres; el mensaje dice que tiene 4 partes y la forma tiene 5 | sí |
| 2 | `bin/kitlegal cita preparar ECLI:FR:CC:2023:1 --json; echo "código $?"` | los tres; el mensaje dice que el ECLI no es español | los tres; «el código de país es "FR" y no ES: no es español» | sí |
| 2 | `bin/kitlegal cita preparar --resolucion 1088/2023 --json; echo "código $?"` | los tres | los tres; el mensaje dice que el número de resolución necesita su fecha | sí |
| 2 | `bin/kitlegal cita preparar ECLI:ES:TS:2023:3144 --texto "cláusula suelo" --json; echo "código $?"` | los tres | los tres; «una referencia y --texto no se pueden dar a la vez» | sí |
| 2 | `bin/kitlegal cita preparar ECLI:ES:TS:2023:3144 --texto "" --json; echo "código $?"` | los tres: una referencia junto a `--texto` | los tres, con el mismo mensaje que la anterior | sí |
| 2 | `bin/kitlegal cita preparar ECLI:ES:TS:2023:3144 --roj "" --json; echo "código $?"` | los tres: más de una forma de referencia | los tres; «la referencia se da de una sola forma y se han dado un ECLI y un ROJ» | sí |
| 2 | `bin/kitlegal cita --json; echo "código $?"` | los tres | los tres; el mensaje pide nombrar un verbo y da los dos, `preparar` y `cotejar`; aquí la `fuente` es `kitlegal.cli` | sí |
| 3 | `F=evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt` | — | la asignación | sí |
| 3 | `bin/kitlegal cita cotejar --json < "$F"` | los ocho datos en `ficha`, `"correspondencia":"se-corresponden"`, `"hallazgos":[]` y `"url":"kitlegal:documento/sha256:4886e0c8…ee27"` | código 0; `roj`, `ecli`, `organo`, `fecha`, `recurso`, `resolucion`, `ponente` y `tipo` en `ficha`; esa correspondencia, `"hallazgos":[]` y la `url` con la huella entera de §0 | sí |
| 3 | `bin/kitlegal cita cotejar ECLI:ES:TS:2023:3144 --json < "$F"` | `"es_la_pedida":true` | código 0; `"es_la_pedida":true` y `"hallazgos":[]` | sí |
| 3 | `bin/kitlegal cita cotejar --roj "STS 1088/2023" --json < "$F"; echo "código $?"` | `código 0`, `"ok":true`, `"es_la_pedida":false` y un hallazgo `documento-distinto` con `"cruce":"numero-de-resolucion"` | los cuatro; el hallazgo es el único | sí |
| 3 | `bin/kitlegal cita cotejar --resolucion 3144/2023 --fecha 2023-07-04 --json < "$F"` | `"cruce":"numero-del-roj"` | código 0; `"es_la_pedida":false` y un hallazgo `documento-distinto` con `"cruce":"numero-del-roj"` | sí |
| 3 | `tail -n 12 "$F" \| bin/kitlegal cita cotejar --json; echo "código $?"` | `código 2`, `argumentos`: el final del fragmento no lleva ficha | `código 2`, `"clase":"argumentos"`; el mensaje dice que falta la línea «Roj:» | sí |
| 3 | `bin/kitlegal cita cotejar --documento "" --json < "$F"; echo "código $?"` | `código 2`, `argumentos`: con `--documento`, también vacío, la entrada no se lee | `código 2`, `"clase":"argumentos"`; el mensaje dice que no ha recibido el texto del documento | sí |
| 3 | `head -n 11 "$F" \| bin/kitlegal cita cotejar --roj "STS 1088/2023" --json` | el mismo `data` que la tercera, con otra `url` | código 0; el mismo `data`, carácter a carácter, y el mismo `hash` (`sha256:fc3a73e9…3142a`); la `url` lleva otra huella, `sha256:e2fbe650…8ab1e`, la de la ficha | sí |

### §4 a §9: los tests y `make ci`

En las tres órdenes de `TestEntregaDelHito`, quickstart.md escribe `h23-`; aquí va `zz-`, el prefijo de las copias
momentáneas (abajo).

| § | Orden | Esperado | Obtenido | ¿Cuadra? |
|---|---|---|---|---|
| 4 | `go test -count=1 -run '^TestHerramientasDelServidor$' ./internal/app/` | `ok` | `ok  	github.com/jmorenobl/kitlegal/internal/app	1.980s` | sí |
| 4 | `go test -count=1 -run '^TestEntregaDelHito$/^zz-cita-herramientas$' ./internal/app/` | `ok` | `ok  	…/internal/app	2.863s` | sí |
| 5 | `go test -count=1 -run '^TestArquitectura$' ./internal/` | `ok` | `ok  	…/internal	0.577s` | sí |
| 5 | `go test -count=1 -run '^TestEntregaDelHito$/^zz-cita-sin-efectos$' ./internal/app/` | `ok` | `ok  	…/internal/app	2.446s` | sí |
| 5 | `ls -A "$KITLEGAL_CACHE_DIR"` | el directorio de los escenarios 1 a 3, vacío | sin salida, código 0 | sí |
| 6 | `go test -count=1 -run '^(TestCitaPreparar\|TestCitaCotejar)$' ./internal/app/` | `ok` | `ok  	…/internal/app	1.798s` | sí |
| 6 | `go test -count=1 -run '^TestEntregaDelHito$/^zz-cita-(preparar\|cotejar)$' ./internal/app/` | `ok` | `ok  	…/internal/app	2.236s` | sí |
| 6 | `go test -count=1 -run '^(FuzzECLI\|FuzzROJ\|TestEquivalencia)$' ./internal/core/ids/` | `ok`; cada `Fuzz…` ejecuta su corpus versionado | `ok  	…/internal/core/ids	0.328s` | sí |
| 6 | `go test -count=1 -run '^(FuzzLeerFicha\|TestLeerFicha\|TestCotejar\|TestPreparar)$' ./internal/core/cita/` | `ok` | `ok  	…/internal/core/cita	0.320s` | sí |
| 6 | `make schema-check` | `ok` | código 0; `ok  	…/internal/app	1.738s` | sí |
| 7 | `make skills-check` | `ok` | código 0; `ok` en `internal/app`, `internal/skills` e `internal/evals` | sí |
| 7 | `wc -l skills/jurisprudencia/SKILL.md` | menos de 300 líneas | `195` | sí |
| 7 | `git diff --stat main -- skills/boe-legislacion skills/legal-core` | ninguna línea de diff | sin salida, código 0 | sí |
| 8 | `go test -count=1 -run '^(TestFormatoDeSentencias\|TestJuzgarSentencias\|TestCitaSinDocumento\|TestPreguntasConElFragmento)$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	0.335s` | sí |
| 8 | `go test -count=1 -run '^(TestUmbralesDeJurisprudencia\|TestUmbralesDelInforme\|TestDefinicionDelJob\|TestEvalsDelRepositorio)$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	1.732s` | sí |
| 9 | `make ci` | verde, con `schema-check` y `skills-check` sin diferencias | código 0 y `ci: todos los controles en verde`, con la caché de tests vacía. Registro de 102 líneas: 53 `ok`, ningún `FAIL` ni `(cached)`; `0 issues.` del lint; `No vulnerabilities found.` y `Your code is affected by 0 vulnerabilities.`; `ok` en `TestEsquemasPublicados` y en los tres paquetes de `skills-check`; `1 configuration file(s) validated`; `no leaks found`; `all modules verified`, seis veces; `go mod tidy -diff` sin salida | sí |

`govulncheck` dice además que hay una vulnerabilidad en un módulo requerido que el código no llama. Es la misma línea
que trae el registro de la verificación de T011 (`gates/ci.log`), y no hace fallar el control.

**Los tests, por su nombre.** Cada orden de `go test` se ejecutó además con `-v`, para contar sus tests y subpruebas y
que ninguna pasara en vacío. Ninguna trae un `--- FAIL` ni un `--- SKIP`:

- §4, 5 `--- PASS`: `TestHerramientasDelServidor` y sus cuatro subpruebas, los applets de producción —12
  herramientas— y los de producción con los de ejemplo —15—, cada uno con el cliente de cada especificación. Y 2:
  `TestEntregaDelHito` y `zz-cita-herramientas`.
- §5, 8: `TestArquitectura` y sus siete subpruebas, entre ellas «sin red · instalacion, disco, cita, ids, el paquete
  raíz y los ficheros de los applets skills y cita y del aviso no alcanzan net, net/http ni internal/httpx». Y 2:
  `TestEntregaDelHito` y `zz-cita-sin-efectos`.
- §6, 73: `TestCitaPreparar` (33 subpruebas, contando todos sus niveles) y `TestCitaCotejar` (38). 3:
  `TestEntregaDelHito`, `zz-cita-preparar` y `zz-cita-cotejar`. 71: `TestEquivalencia` (22), `FuzzECLI` (22: los 11
  ficheros de su corpus versionado y 11 semillas) y `FuzzROJ` (24: 12 y 12). Y 125: `TestPreparar` (28),
  `TestCotejar` (29), `TestLeerFicha` (42) y `FuzzLeerFicha` (22: 11 y 11).
- §8, 155: `TestFormatoDeSentencias` (26), `TestJuzgarSentencias` (102), `TestCitaSinDocumento` (21) y
  `TestPreguntasConElFragmento` (2: `con-un-byte-cambiado-en-la-04` y `-en-la-06`). Y 132:
  `TestUmbralesDeJurisprudencia` (5: `el-elemento-del-contrato`, `ninguna-cuenta`, `una-en-el-modo-orden`,
  `dos-en-el-modo-herramienta` y `una-sin-activar`), `TestUmbralesDelInforme` (20), `TestDefinicionDelJob` (88 en
  todos sus niveles; 7 de primer nivel, entre ellas `del-repositorio`, `peor-caso` y
  `tope-con-las-evals-del-repositorio`) y `TestEvalsDelRepositorio` (15, entre ellas `conjunto-jurisprudencia`).

**Las copias `zz-`.** Los cuatro guiones de la suite congelada se copiaron a `internal/app/testdata/script/` como
`zz-cita-herramientas.txtar`, `zz-cita-sin-efectos.txtar`, `zz-cita-preparar.txtar` y `zz-cita-cotejar.txtar` —`cmp`
de cada copia con su original, sin diferencias— antes de §4, y se borraron después de §6. Los cuatro pasan. Con las
copias ya borradas, `ls` de ese directorio no da ningún `zz-`, `git status --porcelain` da los dos ficheros de
`gates/` y `git ls-files --others --exclude-standard`, nada. El `make ci` de §9 corrió sin ellas.

**Cómo se ejecutó.** Cuatro cosas difieren de teclear el quickstart en una terminal, y ninguna cambia una orden:

- Las órdenes de cada sección se sacaron de sus bloques de código de quickstart.md con `awk` y las ejecutó, una a
  una y diciendo el código de cada una, un guion de un solo uso escrito en el directorio temporal, fuera del
  repositorio (`rtk proxy bash <guion>`, que da la salida sin resumir). La única sustitución es la de `h23-` por
  `zz-` en §4 a §6.
- La salida de error va junto a la estándar. Las dieciocho órdenes de §1 a §3 que llegan a un verbo traen en ella la
  línea `aviso: las skills instaladas son de kitlegal v0.3.1 y este binario es kitlegal v0.5.0-20-g19f332f-dirty;
  ejecuta: kitlegal skills install -g`: es el aviso de H19, que depende de las skills instaladas en la cuenta de
  esta máquina y no de este hito. `cita --json`, que no llega a ninguno, no la trae. Las nueve que fallan repiten en
  ella su mensaje, y las ocho que llegan a un verbo, además, la línea `level=WARN msg=invocación … clase=argumentos`
  del kernel.
- El registro de `make ci` y los de las ejecuciones con `-v` se guardaron en ficheros del directorio temporal, para
  leerlos enteros.
- El directorio de `KITLEGAL_CACHE_DIR` lo creó la orden de §0, `mktemp -d`, y es el mismo de §1 a §5: tiene 0
  entradas al crearlo, después de §3 y en el `ls -A` de §5.

`git status --porcelain` da lo mismo antes de §0 y después de §3, de §6, de §8 y de §9: los dos ficheros de `gates/`.
`bin/kitlegal`, `coverage.out` y `coverage-integration.out` los ignora git (`git check-ignore`).

**§10 y §11, no ejecutados.** Ninguna sesión del run abre una sesión con modelo ni mide en la plataforma:

- §10, el cierre: lo hace el workflow tras la revisión final. Publica la rama, abre la propuesta de cambio y espera a
  `evals (jurisprudencia)`, `evals (boe-legislacion)` y `evals (legal-core)` (SC-001).
- §11, después del run: una persona lee las respuestas a las evals 02, 04 y 06 en los dos modos y lo anota en
  `docs/USO.md`, y comprueba la dirección del buscador del Tribunal Constitucional (SC-002). Ninguna release lleva
  `jurisprudencia` hasta que H25 esté en `main`.

## 4. Controles de umbral (plan.md)

Las 15 filas de plan.md, «Controles de umbral»: 5 con controles `evals:` —14 en total— y 10 con uno o más `ci:`.

### Las diez filas `ci:`

Cada control se busca como lo busca el informe final (`estado_control`, en el guion del informe): la ruta, con `git
cat-file -e HEAD:<ruta>`, y el test, con `git grep -nE '^(func <Test>\(|<Test>:)' HEAD -- <ruta>`. Son 14 controles
distintos —`TestUmbralesDeJurisprudencia` está en dos filas—: **los 11 que nombran un test o un objetivo están en su
ruta, y los tres guiones `h23-` quedan pendientes de la activación**. La misma búsqueda con un nombre que no existe
no da nada. Ninguno de los ocho ficheros de test lleva `//go:build`, el `-skip` de `make test` y de `make
test-integration` solo aparta `TestMedidasDeTiempo` y `TestCosteDelGrafo`, y el registro de `make ci` trae `ok` para
`internal`, `internal/app` e `internal/evals` en esas dos órdenes.

| Fila de plan.md | Control `ci:` | Dónde está | Lo que el árbol da hoy (§3) |
|---|---|---|---|
| FR-040, SC-009 (líneas) | `Makefile:skills-check` y `TestSkillsDelRepositorio` | `Makefile:126` (`skills-check: check-tools`) e `internal/app/skills_test.go:83` | `make skills-check`, código 0; `SKILL.md`, 195 líneas |
| FR-030, SC-009 (herramientas) | `TestHerramientasDelServidor` | `internal/app/herramientas_test.go:137` | `PASS`, 4 subpruebas, con 12 y 15 |
| FR-082, SC-003 | `h23-cita-preparar.txtar` y `h23-cita-cotejar.txtar` | `internal/app/testdata/script/`: **no existen en la cabeza; pendientes de la activación** | las copias `zz-cita-preparar` y `zz-cita-cotejar`, `PASS` |
| FR-080, FR-081, SC-004 | `TestCitaPreparar` y `TestCitaCotejar` | `internal/app/cita_test.go:580` y `:790` | `PASS`, 33 y 38 subpruebas |
| FR-002, FR-084, SC-005 | `TestArquitectura` | `internal/arch_test.go:72` | `PASS`, 7 subpruebas, con la de «sin red» |
| FR-003, FR-085, SC-006 | `h23-cita-sin-efectos.txtar` | `internal/app/testdata/script/`: **no existe en la cabeza; pendiente de la activación** | la copia `zz-cita-sin-efectos`, `PASS` |
| FR-061, FR-086, SC-007 | `TestCitaSinDocumento` y `TestUmbralesDeJurisprudencia` | `internal/evals/sentencias_test.go:1067` e `internal/evals/informe_test.go:4353` | `PASS`, 21 y 5 subpruebas |
| FR-055, FR-063 | `TestUmbralesDeJurisprudencia` y `TestDefinicionDelJob` | `internal/evals/informe_test.go:4353` e `internal/evals/definicion_test.go:182` | `PASS`; la segunda, con `del-repositorio`, `peor-caso` y `tope-con-las-evals-del-repositorio` |
| FR-053, SC-008 | `TestPreguntasConElFragmento` | `internal/evals/conjunto_test.go:3357` | `PASS`, con sus dos subpruebas de un byte cambiado |
| FR-056 | `TestEvalsDelRepositorio` | `internal/evals/conjunto_test.go:1748` | `PASS`, 15 subpruebas, con `conjunto-jurisprudencia` |

**Los tres guiones pendientes.** El informe final da un control `ci:<ruta>` por verificable solo si la ruta existe en
la cabeza, y hoy las tres dan «no existe en la cabeza». Las crea el paso de activación del workflow, que copia cada
guion de `aceptacion/` a `internal/app/testdata/script/` con el prefijo `h23-`: `cita-preparar.txtar`,
`cita-cotejar.txtar` y `cita-sin-efectos.txtar` son tres de los cuatro, y el cuarto, `cita-herramientas.txtar`, se
activa con ellos aunque ninguna fila lo nombre. Lo que esta sesión vio es que los cuatro pasan contra el árbol de hoy
como copias `zz-` (§3).

Que cada control se pone en rojo lo vio su tarea, con sus casos negativos o con mutantes momentáneos que no quedan en
el diff (tasks.md, «Controles de umbral»): esta tarea no los repite ni añade ninguno.

### Las cinco filas `evals:`

Su control es el informe del job de evals, que el workflow lanza en el cierre: **ninguna se mide en el run**, y aquí
no hay ninguna cifra suya.

| Filas | Control | Lo que se ve en el árbol |
|---|---|---|
| FR-060, SC-001 (dos modos) | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:<modo>` | Lo construye `umbralDeCitaSinDocumento` (`internal/evals/umbrales.go`), con umbral 0 y `Decide: true` (§5) |
| FR-062, SC-001 (dos modos) | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:<modo>` | Lo construye `umbralDeSinActivar`, con umbral 0 y `Decide: true` (§5) |
| SC-001 (`boe-legislacion` sigue aprobada) | Sus diez `evals:boe-legislacion:…` | `TestUmbralesDelInforme`, `PASS` con sus 20 subpruebas; `evals/boe-legislacion` y su skill, sin cambios (§1) |

Los cuatro nombres de `jurisprudencia` están escritos, literales, en `internal/evals/informe_test.go` (`grep -c -F`
de cada uno da al menos 1).

## 5. Umbrales, supuestos y atajos (FR-064, FR-065)

### Ningún umbral de FR-060 o de FR-062 se rebajó, pasó a `decide: false` ni perdió evals o un modo de su total

- **Los valores.** En `internal/evals/umbrales.go`: `umbralDeRespuestasSinActivar = 0`, que es el de `main`, en la
  misma línea 13, y `umbralDeRespuestasSinDocumento = 0`, que añade T008. `git log -G` de las dos constantes solo
  nombra T008 (`47268a3`).
- **`decide`.** `umbralDeSinActivar` y `umbralDeCitaSinDocumento` llevan `Decide: true` escrito (`umbrales.go:299` y
  `:319`).
- **El total** es en los dos `len(delModo.respuestas)`, las respuestas medidas del modelo que decide en ese modo en
  las evals que activan la skill (`umbrales.go:291` y `:311`). Los modos los define `internal/evals/plan.go`, que el
  diff no nombra. `TestUmbralesDeJurisprudencia` fija los cuatro umbrales con las evals del repositorio y sesiones
  sintéticas, cada modo por separado: sus motivos esperados dicen «1 de 18» en `una-en-el-modo-orden` y «2 de 18» en
  `dos-en-el-modo-herramienta`, y el modo que no cuenta nada sigue cumpliendo.
- **Las evals.** Las seis de `evals/jurisprudencia/` llevan `activa: true` y `sentencias`, y solo las toca T007:
  ninguna sale ni cambia después. Con las 3 repeticiones del job son las 18 respuestas por modo de plan.md y de ese
  test; el total de las sesiones reales lo da el job.
- **Nada se corrigió después de crearlo.** `umbrales.go` e `informe.go` solo los toca T008; `sentencias.go`, T007 y
  T008; `juzgar.go`, `formato.go` y `conjunto.go`, T007; `evals.yml`, T010; y `SKILL.md`, T009. Lo que cuenta como
  salida de una operación está en `sentencias.go`, que ningún commit posterior a T008 toca. Las doce tareas van por
  su primer intento (`gates/tareas-intentos.json`), la cuarentena está vacía y no hay ninguna nota `gates/tarea-T*.md`.
- **La definición del job.** `git log -G 'MODELO_DE_EVALS:|REPETICIONES_DE_EVALS:|UMBRAL_DE_EVALS:|VERSION_DE_CLAUDE_CODE:|timeout-minutes:'
  main..HEAD -- .github/workflows/evals.yml` no nombra ningún commit: siguen `MODELO_DE_EVALS: claude-sonnet-5-5`,
  `REPETICIONES_DE_EVALS: 3`, `UMBRAL_DE_EVALS: 2`, `VERSION_DE_CLAUDE_CODE: 2.1.284` y el tope de 352 minutos. T010
  añade `jurisprudencia` a la matriz, con `concurrencia: 1` y `objetivo_de_duracion: 0`, y no toca las filas de las
  otras dos skills.
- **Esta sesión no lanzó el job ni un sondeo**: no ha ejecutado `claude`, `make evals`, `make evals-sondeo`,
  `make evals-medir-juez`, `TestEjecucionDelJob`, `TestMedidaDelJuez` ni `TestSondeo`.

### Los dos supuestos de alcance de FR-065 siguen como no comprobados

Están en `gates/supuestos.md`, líneas 7 y 8, con la etiqueta `[alcance]`, que es por la que el informe final los
enseña:

- «Resume»: «que la respuesta no resuma ni caracterice una sentencia cuyo texto no tenía delante (FR-045) no lo decide
  ningún control de este hito […] El informe final no puede darlo por comprobado (FR-065).»
- `no-se-corresponden`: «lo que la respuesta hace cuando `cita cotejar` dice que el ROJ y el ECLI de la ficha no se
  corresponden (FR-048) no tiene eval ni control […] Lo lee una persona en el cierre (FR-065).»

`git blame` da a las dos el commit del plan, `303cd60`: ninguna tarea las ha cambiado. Nada de este fichero las da
por comprobadas: FR-045 y FR-048 no tienen fila en §4.

### Atajos: ningún `//nolint`, `t.Skip` ni TODO nuevos

En las 11 032 líneas añadidas fuera de `specs/` (`git diff main -- . ':!specs'`), la búsqueda de `nolint`, `t.Skip`,
`.Skip(`, `SkipNow`, `TODO`, `FIXME` y `XXX` no da ninguna; la misma búsqueda con `fmt.Errorf` da 6. Y en el árbol
entero, contra `main`:

- **`//nolint:`**: `git grep -c '//nolint:' main -- '*.go'` y la misma orden sobre el árbol dan lo mismo, **7
  directivas en `main` y 7 en el árbol**, en los mismos seis ficheros de test y con el mismo recuento por fichero
  (`coste_test.go`, `e2e_test.go`, con dos, y `tuberia_unix_test.go`, de `internal/app`; `internal/arch_test.go`;
  `internal/core/territorio/coste_test.go`; e `internal/evals/cierre_test.go`).
- **`t.Skip`**: los dos de la cabeza son los dos de `main`, en las mismas líneas
  (`internal/cache/integracion_test.go:451` e `internal/evals/cierre_test.go:70`).
- **TODO, FIXME y XXX**: 12 líneas en `main` y 12 en el árbol, en `*.go`, `Makefile`, `scripts`, `.github`,
  `skills`, `evals` y `schemas`. Las 12 del árbol están en ficheros que el diff no nombra, y ninguna es una tarea
  pendiente: son plantillas `XXXXXX` de `mktemp`, `\uXXXX` y «MÉTODO».

La palabra española que `misspell` tomaba por errata, `constitucional`, está en `ignore-rules` de `.golangci.yml`
(T002), que es donde el repositorio neutraliza el español, y no en un `//nolint`.

## 6. Lo que queda fuera del run

- **Las cinco filas `evals:`** (§4) y SC-001: las mide el job de cierre que lanza el workflow.
- **Los tres controles `ci:` de los guiones `h23-`** (§4): existen en la cabeza cuando el workflow activa la suite,
  tras el bucle de tareas.
- **`codecov/patch`** (§2): solo se sabe en la propuesta de cambio.
- **quickstart.md §10 y §11**, SC-002 y SC-010: del workflow, de una persona y de la revisión final.
- **La verificación de esta tarea** es otro `make ci`, posterior a la última edición de este fichero: su resultado no
  está aquí.
