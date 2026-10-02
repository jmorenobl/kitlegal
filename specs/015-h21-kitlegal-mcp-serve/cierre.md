# H21 · Cierre de la Definition of Done

Cierre de T018 (FR-048, FR-049, FR-051, FR-079, FR-082, FR-090; SC-013) sobre `6e4ae3a`, que es T017, el último commit
antes de T018. La rama está al día con `main`: `git merge-base HEAD main` da la cabeza de `main`, `22b5bda`, así que
`git diff --name-status main` da exactamente lo que cambia el hito: 150 ficheros (85 `A`, 64 `M` y 1 `R`, el test de la
tubería que T017 renombró a `internal/app/tuberia_unix_test.go`). Las cuatro partes son las de la tarea: los esquemas,
evals y datos de prueba que toca el hito, la cobertura, el quickstart y lo que el hito no cambia, con los supuestos de
research al final.

Todo lo que se cita aquí se ejecutó en la sesión de T018 y se vio terminar, o está en el repositorio. Lo que no se pudo
medir se dice tal cual (§3, escenarios 11 a 13; §5).

## 1. Esquemas, evals y datos de prueba tocados en el hito

Esta es la lista para la capa 3 del informe final. Recoge cada fichero de `schemas/`, de `evals/` y de cualquier
`testdata/` que el hito crea o modifica, con la tarea que lo tocó y el motivo. Sale de `git diff --name-status main --
schemas evals testdata '*/testdata/*'`, y la tarea de cada fila, de `git log --format=%s main..HEAD -- <fichero>`. Son
28 ficheros: 2 `M` (un esquema y un guion) y 26 `A`. No hay ninguna `D` ni ninguna `R`. Las tres tareas que los tocan
—T008, T010 y T014— llevan `[datos]` en `tasks.md` (FR-051).

### Esquemas (`schemas/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| A | `schemas/servidor.json` | T008 `[datos]` | El contrato publicado del verbo `serve` del applet `mcp` (`$id` `https://kitlegal.es/schemas/servidor.json`, título `mcp · servidor`; 128 líneas, 3 265 bytes). Lo escribe `TestEsquemasPublicados -actualizar-esquemas` desde `mcp serve --describe` y `schema-check` lo compara en cada `make ci` (§3, escenario 3). |
| M | `schemas/eval.yaml.json` | T010 `[datos]` | El formato de eval gana `sin_binario_ni_servidor`, que solo admite `true`. Una eval que activa la skill y la lleva no admite `comandos`, `prohibidos`, `grafo_previo`, `citas`, `avisos`, `hallazgos`, `redacciones_modificadas`, `territorio`, `informativa` ni `reproduce`; sin ella, sigue exigiendo `comandos` y `citas` o `territorio`, como en `main`. Una eval de no activación tampoco la admite. El diff son 18 líneas añadidas y 3 quitadas. No sale de `--describe`, así que `schema-check` no lo compara: lo fijan los casos de `formato_test.go`. |

Los otros doce ficheros de `schemas/` no cambian: entre ellos están los de los verbos de hoy (`bloque.json`,
`norma.json`, `municipio.json`, `grafo.json`, `instalacion.json`), que es lo que §4 lee de `schema-check`.

### Evals (`evals/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| A | `evals/boe-legislacion/21-sin-binario-ni-servidor.yaml` | T010 `[datos]` | La eval sin binario ni servidor de `boe-legislacion` (FR-046, FR-047): pregunta por el art. 53 de la Ley 39/2015 con `activa: true` y `sin_binario_ni_servidor: true`, y nada más. Son 5 líneas. La skill pasa a 21 evals. |
| A | `evals/legal-core/04-sin-binario-ni-servidor.yaml` | T010 `[datos]` | La de `legal-core`: pregunta por la comunidad, la provincia y los boletines del Ayuntamiento de Getafe, con `activa: true`, `no_se_activan: [boe-legislacion]` y `sin_binario_ni_servidor: true`. Son 6 líneas. La skill pasa a 4 evals. |

Las evals 01 a 20 de `boe-legislacion`, las 01 a 03 de `legal-core` y la lista
`evals/boe-legislacion/expresiones-prohibidas.yaml` no cambian (§4).

### Datos de prueba (`testdata/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| M | `internal/app/testdata/script/argumentos.txtar` | T008 `[datos]` | El guion de H1 que fija la lista de applets del binario de e2e. Cambian sus líneas 24 y 30, y nada más (2 añadidas y 2 quitadas): `applets disponibles: boe, contar, echo, graph, skills, territorio` pasa a `… graph, mcp, skills, territorio`. Es el único guion anterior que cambia (§4, FR-090). |
| A | `internal/evals/testdata/sesiones/leer-llamadas/` (23 ficheros en 7 carpetas) | T014 `[datos]` | Los transcripts sintéticos de `TestLeerLasLlamadas`, una carpeta por caso: `con-prefijo`, `sin-prefijo`, `modo-orden`, `herramienta-ajena`, `error-por-is-error`, `error-por-ok-falso` y `sin-resultado`. Cada una lleva `sesion.jsonl` (7 líneas, y 4 en `sin-resultado`), `codigo-de-la-sesion` y `sesion.err`; las dos del modo herramienta, `con-prefijo` y `sin-prefijo`, llevan además `servidor.json`, el fichero de configuración del servidor de su sesión (358 bytes cada uno, distintos entre sí). `sin-resultado` es la sesión que el tope cortó: código `124` y una línea en `sesion.err`; las otras seis terminan con `0` y `sesion.err` vacío. Son 32 555 bytes en total. |

### Los cinco guiones de la suite

`specs/015-h21-kitlegal-mcp-serve/aceptacion/` tiene los cinco guiones que escribió T001 y que ningún commit posterior
toca (`git log main..HEAD -- …/aceptacion` solo nombra T001): `mcp-llamadas.txtar` (182 líneas), `mcp-errores.txtar`
(92), `mcp-herramientas.txtar` (57), `mcp-protocolo.txtar` (82) y `mcp-proceso.txtar` (84). `shasum -a 256` de los
cinco da, uno a uno, los hashes de `gates/aceptacion-congelada.json`. Todavía no están en `internal/app/testdata/script/`:
los copia allí el workflow, con el prefijo `h21-`, al activar la suite tras el bucle de tareas. §3 los ejecuta con copias
momentáneas.

### Ninguna grabación ni ninguna fuente cambia

- **Ninguna grabación cambia.** `git diff --name-status main -- internal/source '*grabaciones.json'
  '*grabacion_test.go' internal/core testdata` no imprime nada. Los dos manifiestos
  (`internal/source/boe/testdata/grabaciones.json` y `testdata/evals/grabaciones.json`) y los dos tests
  `//go:build grabacion` (`internal/source/boe/grabacion_test.go` e `internal/evals/grabacion_test.go`) son los de
  `main`, y no hay ninguno nuevo. El e2e se sirve de las respuestas del BOE grabadas desde H4 y el paso `grabar_datos`
  no tenía nada que grabar (FR-051; plan.md, «Datos externos»).
- **Ninguna fuente cambia**: `git diff --name-status main -- docs/SOURCES.md evidencias data web` no imprime nada (§4).

### Lo que no está en la lista

- **`skills/boe-legislacion/SKILL.md`** y **`skills/legal-core/SKILL.md`** (M, T012): son la entrega de las skills, con
  297 y 190 líneas (§4), fuera de esos árboles.
- **`.github/workflows/evals.yml`** (M), **`scripts/evals.sh`** y **`scripts/evals-sesion.sh`** (M), que son del job de
  evals y no del workflow `hito`; **`.golangci.yml`** (M, la regla R7 de `depguard`); **`go.mod`** y **`go.sum`** (M, el
  SDK y sus seis módulos); **`README.md`**, **`CONTRIBUTING.md`** y **`CHANGELOG.md`** (M, T017); y el código y los tests
  de `cmd/` e `internal/`. Ninguno es un dato.
- **Sondas y mutantes momentáneos**: se crearon y se retiraron dentro de su tarea. `git log --name-only main..HEAD`
  nombra 150 ficheros distintos, y ninguno lleva `zz-`, `.orig`, `mutante` ni `sonda` (0 coincidencias).

## 2. Cobertura (Definition of Done §1.9)

La cobertura se mide sobre los dos perfiles que deja el primer `make ci` de esta tarea, en verde (`ci: todos los
controles en verde`) y con la caché de tests vacía (`go clean -testcache` antes): `coverage.out` (`make test`) y
`coverage-integration.out` (`make test-integration`). El global sale de `go tool cover -func` sobre cada perfil. El de
cada árbol y el de cada fichero salen de `go tool cover -func` sobre el perfil filtrado a sus líneas, con su cabecera
`mode:`, escrito en el directorio temporal, fuera del repositorio. Los recuentos de sentencias salen de los bloques del
perfil, contando cada bloque una vez. «Unión» son los dos perfiles fundidos bloque a bloque, como hace Codecov con los
dos que publica CI.

| Árbol | `coverage.out` | `coverage-integration.out` | Unión | Umbral (`codecov.yml`) |
|---|---|---|---|---|
| Global | **96,9 %** (12 504 de 12 901 sentencias) | **97,4 %** (12 566 de 12 901) | 97,4 % | ≥ 70 % |
| `internal/core/**` (dominio) | **98,6 %** (2 545 de 2 581) | **98,6 %** (2 545 de 2 581) | 98,6 % | ≥ 85 % |
| `internal/cli/**` | 98,5 % (528 de 536) | 98,5 % | 98,5 % | ≥ 90 % |
| `internal/mcp/**` | 96,6 % (288 de 298) | 96,6 % | 96,6 % | — |
| `internal/app` (sin `ejemplo/`) | 96,8 % (1 187 de 1 226) | 96,8 % | 96,8 % | — |
| `internal/evals` | 97,2 % (3 631 de 3 735) | 98,1 % (3 662 de 3 735) | 98,1 % | — |
| `internal/httpx` | 97,3 % (820 de 843) | 97,3 % | 97,3 % | — |

Los tres umbrales de proyecto se cumplen. No hace falta ningún test, así que T018 no toca ningún fichero de test ni
ningún fichero fuera de `specs/015-h21-kitlegal-mcp-serve/`. El hito no toca el dominio: `git diff --name-status main --
internal/core` no imprime nada, y `internal/core` da las mismas 2 545 de 2 581 sentencias que `main`.

**`main`, medido en esta sesión.** Para tener la base con el mismo criterio, las dos órdenes de `make test` y `make
test-integration` se ejecutaron también sobre `22b5bda`, en un worktree temporal fuera del repositorio que se retiró
después (`git worktree list` vuelve a dar solo el del repositorio). Dan 96,8 % (11 539 de 11 916) y 97,4 % (11 601 de
11 916): las mismas cifras que el cierre de H7.4. El hito añade 985 sentencias y la unión pasa de 97,36 % a 97,40 %.
`internal/cli` pasa de 372 de 377 (98,7 %) a 528 de 536 (98,5 %): las tres sentencias nuevas sin cubrir son las de
`herramienta.go`, abajo.

**Los ficheros nuevos** dan lo mismo en los dos perfiles:

| Fichero | Sentencias | Funciones (`go tool cover -func`) |
|---|---|---|
| `internal/mcp/servir.go` (el adaptador del protocolo) | 25 de 25 (100 %) | `Servir`, `manejador`, `Read`, `terminada` y los dos `Close`, al 100 % |
| `internal/mcp/instrucciones.go` e `internal/mcp/doc.go` | sin sentencias | El texto de las `instructions` es una constante y `doc.go` solo documenta el paquete: el perfil no tiene ningún bloque suyo. |
| `internal/mcp/mcptest/sesion.go` (los dos clientes de prueba) | 263 de 273 (96,3 %) | `abrirVigente` 95,8 %, `nombresDeLasCapacidades` 71,4 %, `(*vigente).herramientas` 85,7 %, `(*anterior).leer` 81,0 % y `(*anterior).enviar` 90,0 %. Las otras 29, al 100 % |
| `internal/cli/herramienta.go` (los esquemas y la línea de llamada) | 152 de 155 (98,1 %) | `EsquemasDeHerramienta` 90,3 %. Las otras once (`retirarDefiniciones`, `LineaDeLlamada`, `propiedadesDeLaLlamada`, `sinPropiedadesDeMas`, `escrituras`, `textosDelArgumento`, `deUnValor`, `deUnaLista`, `textoDeCadena`, `textoDeBooleano` y `textoDeEntero`), al 100 % |
| `internal/app/herramientas.go` (el applet) | 111 de 111 (100 %) | `verbosAnunciados`, `NombresDeHerramientas`, `herramientasDe`, `banderasDeCadaLlamada`, `atender` y `resolver`, al 100 % |
| `internal/app/mcp.go` (el applet) | 21 de 21 (100 %) | `DependenciasDeMCPDelSistema`, `AppletMCP`, `Nombre`, `Descripcion`, `Verbos`, `Ejecutar`, `servir` y `validar`, al 100 % |

Las 13 sentencias sin cubrir de los ficheros nuevos son ramas de error:

- `internal/cli/herramienta.go` (3): el error de reflejar la salida del verbo (45-46) y los dos de serializar un esquema
  (59-60 y 64-65). El verbo que no se puede describir sí tiene caso, por sus argumentos y por la colisión de nombres
  (`TestEsquemasDeHerramienta/un_verbo_que_no_se_puede_describir_no_tiene_esquemas`).
- `internal/mcp/mcptest/sesion.go` (10): en el cliente vigente, el error de leer las capacidades del saludo (312-313),
  los dos de serializarlas y leerlas (329-330 y 334-335) y los de los dos esquemas de una herramienta que no se pueden
  serializar (350-351 y 355-356); en el anterior, la línea que el lector tiene en la mano cuando la sesión se abandona
  (577-579) y el error de serializar un mensaje propio (769-770).

**El diff.** H21 no pide ninguna cifra del diff, pero Codecov la mide en la propuesta de cambio con el estado `patch`,
bloqueante y con `target: auto`, que es la cobertura de la base. La estimación local usa el criterio de los cierres de
H7.1 a H7.4: cada línea añadida de los `.go` de producto de `git diff -U0 main` que cae en un bloque de la unión de los
dos perfiles cuenta, y cuenta como cubierta si alguno de esos bloques lo está. Da **97,4 %: 1 221 de 1 254 líneas**. Con
ese mismo criterio de líneas, el proyecto entero da 95,5 % en `main` (12 591 de 13 188) y 95,6 % en la cabeza (13 688
de 14 314): el diff queda por encima de la base. Es una estimación: Codecov cuenta a su manera las líneas cubiertas a
medias, y lo que diga `codecov/patch` solo se sabe en la propuesta de cambio. Las 33 líneas que ningún perfil ejecuta:

| Fichero | Líneas | Qué son |
|---|---|---|
| `internal/mcp/mcptest/sesion.go` | 312-313, 329-330, 334-335, 350-351, 355-356, 577-579, 769-770 (15) | Las ramas de error de arriba. |
| `internal/cli/herramienta.go` | 45-46, 59-60, 64-65 (6) | Las ramas de error de arriba. |
| `internal/evals/juzgar.go` | 494-497, 591-592, 613-614 (8) | El registro de producción que no se puede construir, en el juicio de las llamadas (494-497) y en `verbosDeLasHerramientas` (613-614), y el texto de `falloDeLaLlamada.Error()` (591-592), que nadie lee: el juicio pregunta por su clase. |
| `internal/evals/sesion.go` | 236-237, 287-288 (4) | Lo mismo en `LeerSesion` y en `herramientasDelRegistro`: el registro de producción que no se puede construir. |

Queda como supuesto `[alcance]` de T018 en `gates/supuestos.md`. La tarea solo pide tests si una cifra de cobertura
queda bajo su umbral, y ninguna queda. Ningún umbral se toca.

## 3. Quickstart (§1 a §10 y §14)

Se ejecutaron sobre `6e4ae3a`, desde la raíz del repositorio. §1 es el primer `make ci` de esta tarea. Las órdenes de §3
a §10 y §14 no se reescribieron: se copiaron del propio `quickstart.md` a guiones del directorio temporal, que se
ejecutaron con `rtk proxy bash`, sin el resumen del proxy de la sesión. Cada orden de `go test` se ejecutó como está y,
además, con `-v`, para contar sus líneas `--- PASS` y `--- FAIL` y que ninguna pasara en vacío. Dos cosas se apartan de
la letra del quickstart, y las dos están en `gates/supuestos.md`:

- `T` no es `$(mktemp -d)` sino `$(mktemp -d "$TMPDIR/h21-qs.XXXXXX")`: en macOS, `mktemp -d` sin plantilla crea el
  directorio en el temporal del usuario y no en `$TMPDIR`, y la tarea pide el binario bajo `$TMPDIR`.
- En §7, el ensayo no tiene «la entrada de la terminal abierta», porque la sesión no tiene terminal: la entrada del
  servidor es una tubería que nadie cierra hasta pasados 10 s.

**Todos los escenarios ejecutados dan lo esperado: 408 `--- PASS` y ningún `--- FAIL` en las órdenes de `go test` de
§2 a §4 y §8 a §10** (6, 69, 79, 31, 14 y 209).

| Escenario | Resultado |
|---|---|
| §1. `make ci` | En primer plano: código 0 y `ci: todos los controles en verde`. `0 issues.` de `golangci-lint` (con `depguard` y su regla R7), `ok` en los 22 paquetes con tests, con `-race` y con la etiqueta `integration`, `ok` en `TestMedidasDeTiempo` y `TestCosteDelGrafo`, `No vulnerabilities found.`, `ok` en `TestEsquemasPublicados` (`schema-check`, sin drift), `ok` en `internal/app`, `internal/skills` e `internal/evals` (`skills-check`, sin drift), `no leaks found`, `all modules verified` y `go mod tidy -diff` sin salida. El registro, de 96 líneas, no lleva `truncated`. Lo que el quickstart espera y todavía no puede darse: los guiones `h21-mcp-*` dentro de `TestEntregaDelHito`, que llegan cuando el workflow activa la suite (§2). |
| §2. Los guiones de aceptación | Con copias momentáneas de los cinco guiones congelados en `internal/app/testdata/script/`, de prefijo `zz-`, y `-run '^TestEntregaDelHito$/^zz-mcp-'`: código 0 y **6 `--- PASS`, ningún `--- FAIL`**: `TestEntregaDelHito` y `zz-mcp-herramientas`, `zz-mcp-errores`, `zz-mcp-llamadas`, `zz-mcp-proceso` y `zz-mcp-protocolo`. Las copias se retiraron en la misma orden (0 ficheros `zz-` después) y `git status --porcelain` quedó como antes. Con el prefijo `h21-` no se ha ejecutado: lo pone el workflow. |
| §3. La conformidad de las herramientas | Las tres órdenes, código 0. Con `-v`: **5**, **47** y **17 `--- PASS`, ningún `--- FAIL`** (69). |
| §4. Llamadas simultáneas, cierre, ensayo y plazo | Las tres órdenes, con `-race`, código 0. Con `-v`: **8**, **68** y **3 `--- PASS`, ningún `--- FAIL`** (79), y 0 `DATA RACE` en los tres registros. |
| §5. A mano: listar y llamar | `código 0`; `5` líneas; `"2025-11-25"`, `{"tools":{}}`, `"kitlegal"` y `485`; diez líneas, de `boe_analisis true` a `territorio_resolver true` (`boe_analisis`, `boe_articulo`, `boe_articulos`, `boe_buscar`, `boe_indice`, `boe_metadatos`, `graph_check`, `graph_show`, `graph_stats` y `territorio_resolver`); `mismo sobre` (1 356 bytes); `true` y `null`; `true`, `false` y `"argumentos"`; `-32602` y `true`. El error del protocolo es `unknown tool "skills_install"`, y el sobre de fallo de `boe_articulo` lleva `fuente` `kitlegal.cli` y el mensaje `argumentos inválidos: expected "<bloque>"`. En la salida de error del servidor, solo el evento de esa llamada (`level=WARN msg=invocación applet=boe verbo=articulo … clase=argumentos`) y su mensaje. |
| §6. A mano: `boe_articulo` y el grafo | `código 0`; `mismo sobre` (4 321 bytes, `fecha_consulta` `2026-09-28T12:00:00Z`); `null` y `true`; y en la caché, `cache.db` y `world.db`. `graph_check` devuelve `version-obsoleta` 0, `fuente-caducada` 0 y `hallazgos` vacío, y la salida de error del servidor queda en 0 bytes. |
| §7. A mano: el proceso | Con la entrada cerrada, `código 0` y `0` bytes. Con `--asunto x`, `código 2`, `0` bytes en la salida estándar y, en la de error, el evento `level=WARN msg=invocación applet=mcp verbo=serve … clase=argumentos` y `argumentos inválidos: mcp serve no admite --asunto: el servidor no expone nada del asunto`. Con `--dry-run`, `código 0`, `0` bytes en la salida estándar y `--dry-run: no se ha ejecutado nada; se habría ejecutado el applet "mcp", el verbo "serve", con los argumentos ["serve" "--dry-run"]`; el servidor volvió a los 0 s con su entrada abierta hasta los 10 s. Desde `/` y desde `con espacios/`, `código 0` y `mismo sobre`, con 0 bytes en la salida de error. |
| §8. Las dos skills | `make skills-check`, código 0, con `ok` en `internal/app`, `internal/skills` e `internal/evals`. `wc -l`: **297** y **190**. La línea `⚠ SIN CONSULTA AL BOE:`, `1` en cada fichero. La cabecera de la tabla con «Herramienta», `2` y `1`. Las dos órdenes de `go test`, código 0; con `-v`, **21** y **10 `--- PASS`, ningún `--- FAIL`** (31). El calibrado en 36, 11 y 9 lo comprueba `TestEvalsDelRepositorio/expresiones-calibradas`, que pasa dentro de `skills-check` y, sola con `-v`, da 2 `--- PASS`. |
| §9. Las `instructions` y la arquitectura | Las dos órdenes de `go test`, código 0: **5** y **9 `--- PASS`, ningún `--- FAIL`** (14). `go list -deps` da `1`. Los ficheros con una línea de importación del SDK son cuatro, todos de `internal/mcp/` y de `internal/mcp/mcptest/`: `servir.go`, `servir_test.go`, `mcptest/sesion.go` y `mcptest/sesion_test.go` (SC-013). |
| §10. Las evals en dos modos, sin modelo | Las tres órdenes de `go test`, código 0. Con `-v`: **56**, **94** y **59 `--- PASS`, ningún `--- FAIL`** (209). `grep` da `160:    timeout-minutes: 240` y `ls` nombra las dos evals nuevas. Ninguna de las tres órdenes abre una sesión con modelo: usan sesiones y transcripts sintéticos y el `claude` sustituto de los tests. |
| §11. El sondeo rechaza una eval sin binario ni servidor | **No se ejecuta.** `make evals-sondeo` lo lanza una persona, fuera del run (ADR 0032). Lo que cubre ese caso en `make ci` son los tests de §10: `TestComprobarElSondeo/evals-sin-binario-ni-servidor-sola`, `…-con-otras-que-se-miden` y `…-dos-en-su-orden`, `TestSondear/una-eval-sin-binario-ni-servidor` y `…-y-otra-que-se-mide`, y `TestGuionDelSondeo/uso-con-una-eval-sin-binario-ni-servidor`. |
| §12. El job de cierre | **No se ejecuta.** Lee `gates/evals/*.json`, que deja el cierre del workflow después del run. Esa carpeta no existe todavía. |
| §13. La prueba humana | **No se ejecuta.** Es de después de fusionar y la hace una persona (SC-002). |
| §14. Limpieza | `árbol intacto`: borrado `$T`, `git status --porcelain` es el del principio, los dos ficheros de `gates/` que el workflow modifica antes de la tarea. Antes de borrar, el `HOME` temporal estaba vacío y la caché temporal tenía `cache.db` y `world.db`; en `~/.cache/kitlegal` y `~/.agents` (50 ficheros), ninguno modificado desde la preparación. |

### Los tests y las subpruebas, por su nombre

Cada nombre es el de una línea `--- PASS` de la orden con `-v`, como lo escribe `go test`.

**§3, primera orden (5).** `TestHerramientasDelServidor` y sus cuatro subpruebas:
`con_los_applets_de_producción,_el_cliente_de_la_especificación_anterior`, `…_vigente`,
`con_los_de_producción_y_los_de_ejemplo,_el_cliente_de_la_especificación_anterior` y `…_vigente`.

**§3, segunda orden (47).** `TestLineaDeLlamada` y 38 subpruebas:

- las líneas: `la_línea:_sin_arguments_cuenta_como_un_objeto_vacío`, `la_línea:_un_objeto_vacío`,
  `la_línea:_un_verbo_sin_argumentos`, `la_línea:_una_cadena_por_su_posición`, `la_línea:_un_entero_por_su_posición`,
  `la_línea:_un_entero_negativo`, `la_línea:_un_entero_escrito_con_exponente`,
  `la_línea:_un_entero_escrito_con_decimales,_que_el_esquema_admite_como_entero`,
  `la_línea:_un_entero_que_no_cabe_en_un_número_de_coma_flotante_conserva_sus_cifras`,
  `la_línea:_una_cadena_vacía_es_un_argumento`, `la_línea:_una_lista_de_cadenas,_un_argumento_por_elemento`,
  `la_línea:_una_lista_vacía_no_escribe_nada`, `la_línea:_una_bandera_booleana_verdadera`,
  `la_línea:_lo_que_no_va_por_su_posición_va_delante,_como_bandera_con_su_valor`,
  `la_línea:_lo_que_parece_una_bandera_va_detrás_del_terminador,_tal_cual`,
  `la_línea:_los_posicionales_van_en_el_orden_de_sus_campos_y_no_en_el_del_objeto` y
  `la_línea:_falta_un_obligatorio:_la_línea_se_devuelve_y_lo_dirá_el_analizador`;
- los rechazos: `el_rechazo:_una_cadena_no_es_un_objeto`, `el_rechazo:_un_número_no_es_un_objeto`,
  `el_rechazo:_una_lista_no_es_un_objeto`, `el_rechazo:_null_no_es_un_objeto`,
  `el_rechazo:_dos_documentos_no_son_un_objeto`, `el_rechazo:_lo_que_no_es_JSON_no_es_un_objeto`,
  `el_rechazo:_sobra_una_propiedad`, `el_rechazo:_de_varias_que_sobran_se_nombra_siempre_la_misma`,
  `el_rechazo:_el_nombre_de_una_bandera_global_es_una_propiedad_que_sobra`,
  `el_rechazo:_un_argumento_de_posición_sin_el_que_le_precede`, `el_rechazo:_null_donde_va_una_cadena`,
  `el_rechazo:_un_número_donde_va_una_cadena`, `el_rechazo:_una_cadena_donde_va_una_lista_de_cadenas`,
  `el_rechazo:_una_lista_con_un_elemento_que_no_es_una_cadena`, `el_rechazo:_una_cadena_donde_va_un_entero`,
  `el_rechazo:_un_número_con_decimales_donde_va_un_entero` y `el_rechazo:_una_cadena_donde_va_un_booleano`;
- y `que_falte_un_obligatorio_lo_dice_el_analizador_de_la_orden`,
  `lo_que_va_detrás_del_terminador_no_es_ninguna_bandera`,
  `un_campo_que_una_llamada_no_sabe_escribir_es_un_defecto_del_verbo` y
  `el_analizador_de_la_orden_lee_cada_tipo_de_argumento`.

`TestEsquemasDeHerramienta` y 7: `un_verbo_que_no_se_puede_describir_no_tiene_esquemas`,
`la_entrada_del_ejemplo_del_contrato,_byte_a_byte`, `la_entrada_no_lleva_ninguna_bandera_global`,
`sin_$schema_y_en_JSON_compacto`, `cada_esquema_lleva_las_definiciones_que_referencia_y_solo_esas`,
`los_dos_esquemas_son_las_partes_del_documento_de_--describe` y `cada_esquema_vale_suelto_ante_un_validador`.

**§3, tercera orden (17).** `TestEsquemasPublicados` y 16: `schemas`, `regenerados-coinciden`, `forma-del-contrato`,
`forma-canonica`, `no-es-json`, `otro-sangrado`, `sin-salto-final`, `salida-cambiada-sin-regenerar`,
`raiz-editada-a-mano`, `parte-editada-a-mano`, `parte-que-falta`, `parte-de-mas`, `compara-solo-los-que-existen`,
`el-applet-sale-de-la-tabla`, `sin-ficheros-no-compara-nada` y `escribe-solo-para-su-propietario`.

**§4, primera orden (8).** `TestLlamadasSimultaneas`, `TestServirSinEntrada`, `TestServirEnEnsayo`,
`TestPlazoDeCadaLlamada` y `TestDependenciasDeRed` con sus tres subpruebas:
`el-primer-cliente-gasta-el-turno-en-su-robots`, `el-segundo-cliente-no-abre-ninguna-conexion` y
`el-de-otras-dependencias-abre-la-suya`.

**§4, segunda orden (68).** En `internal/mcp`, 14:

- `TestInstrucciones` y 4: `el_texto_son_las_cinco_frases_y_nada_más`,
  `las_cinco_frases,_en_su_orden,_dentro_de_los_primeros_512_bytes`, `mide_485_bytes` y
  `control:_una_frase_detrás_del_byte_512,_una_que_falta_y_una_fuera_de_su_orden_se_nombran`;
- `TestServir` y 8: `anuncia_solo_herramientas,_con_las_instrucciones,_el_nombre_y_la_versión`,
  `anuncia_cada_herramienta_con_su_nombre,_su_descripción,_sus_dos_esquemas_byte_a_byte_y_solo_lectura`,
  `una_llamada_que_termina_bien:_un_bloque_de_texto_con_el_sobre_y_el_mismo_documento_estructurado`,
  `una_llamada_que_falla:_el_sobre_de_fallo,_marcado_como_error_de_herramienta`,
  `una_herramienta_que_no_está_la_responde_el_protocolo,_sin_llamar_a_ninguna`,
  `nil_al_cerrarse_la_entrada_sin_llamadas`, `nil_al_cerrarse_la_entrada_con_una_llamada_en_curso,_que_termina` y
  `el_error_de_Run_si_la_entrada_no_se_ha_cerrado`.

En `internal/mcp/mcptest`, 54: `TestSesion` y sus tres grupos.

- `el_cliente_vigente` y `el_cliente_anterior`, cada uno con las mismas 17:
  `da_el_protocolo_negociado,_las_instrucciones_y_las_capacidades`, `da_las_capacidades_por_su_nombre_y_ordenadas`,
  `da_cada_herramienta_con_su_nombre,_su_descripción,_sus_esquemas_y_si_es_de_solo_lectura`,
  `da_todas_las_herramientas_de_una_lista_que_llega_en_varias_páginas`,
  `da_el_sobre_de_una_llamada,_con_los_argumentos_enviados_y_sin_ninguno`,
  `da_el_sobre_de_fallo_de_una_llamada_que_falla,_marcado`,
  `da_el_error_del_protocolo_de_una_herramienta_que_no_está,_y_sigue`,
  `rechaza_unos_argumentos_que_no_son_JSON_sin_enviar_nada`,
  `rechaza_el_resultado_que_no_cumple_lo_que_comprueba_por_su_cuenta`,
  `entrega_a_cada_llamada_simultánea_su_resultado`, `no_llama_con_el_contexto_terminado`,
  `deja_de_esperar_una_respuesta_cuando_el_contexto_termina,_y_sigue`,
  `da_un_error_si_el_servidor_termina_con_una_llamada_en_curso`,
  `da_su_error_propio_si_el_servidor_termina_sin_leer_el_saludo`,
  `da_su_error_propio_si_el_servidor_lee_el_saludo_y_termina`, `da_el_fallo_de_cerrar_la_entrada_del_servidor` y
  `al_cerrar,_cierra_la_entrada_del_servidor,_que_termina`;
- `el_cliente_anterior,_línea_a_línea`, con 9: `da_su_error_propio_si_el_servidor_responde_al_saludo_y_termina`,
  `da_el_error_del_protocolo_con_el_que_el_servidor_rechaza_el_saludo`,
  `da_un_error_si_el_resultado_no_tiene_la_forma_de_su_petición`, `no_toma_por_suya_la_respuesta_a_otra_petición`,
  `guarda_las_intercaladaes_y_las_peticiones_del_servidor,_que_no_son_la_respuesta`,
  `al_cerrar,_deja_de_esperar_el_final_de_la_salida_con_el_contexto`,
  `rechaza_una_línea_que_no_es_del_protocolo_tras_cerrar_la_entrada`,
  `rechaza_una_línea_que_no_es_del_protocolo_durante_una_llamada,_y_la_guarda` y
  `rechaza_una_línea_que_no_es_del_protocolo_antes_del_saludo`, esta con 7 más: `un_aviso_en_texto`,
  `una_línea_vacía`, `un_valor_suelto`, `un_lote_de_mensajes`, `un_objeto_sin_jsonrpc`,
  `un_jsonrpc_que_no_es_la_cadena` y `un_mensaje_de_otra_versión`.

**§4, tercera orden (3).** `TestRitmoCompartido` y 2: `las_llegadas_de_un_cliente_y_de_otro_no_se_adelantan_a_su_turno`
y `el_robots.txt_sigue_siendo_de_cada_cliente`.

**§8, primera orden de `go test` (21).** `TestOrdenesDeLasSkillsEmpotradas` y 3: `boe-legislacion`, `legal-core` y
`control`. `TestTablaDeComandosCoincideConLaGramatica` y 16: una por verbo del registro (`boe_analisis`,
`boe_articulo`, `boe_articulos`, `boe_buscar`, `boe_indice`, `boe_metadatos`, `graph_check`, `graph_show`,
`graph_stats`, `mcp_serve`, `skills_doctor`, `skills_install`, `skills_list` y `territorio_resolver`) y las dos de
control, `obligatorio-presentado-como-opcional` y `obligatorio-y-lista-presentados-como-opcionales`.

**§8, segunda orden de `go test` (10).** `TestRenderizarTabla` y 9: `filas-del-contrato`,
`secciones-en-el-orden-declarado`, `sintaxis-y-escapes`, `opcionales-de-posicion`, `banderas-distintas`,
`sobre-distinto`, `una-sola-clave-de-fallo`, `applet-sin-verbos` y `sin-applets`.

**§9, primera orden (5).** `TestInstrucciones` y sus 4 subpruebas, las de §4.

**§9, segunda orden (9).** `TestDependenciasDelBinario`, y `TestArquitectura` con 7:
`R1_·_el_dominio_no_importa_el_kernel,_los_adaptadores_ni_entrada_y_salida`,
`R2_·_solo_internal/httpx_importa_net/http`, `R3_·_solo_internal/{cache,store,graph}_importan_SQLite_y_database/sql`,
`R6_·_internal/graph_no_importa_internal/source_ni_internal/render`,
`R7_·_solo_el_paquete_del_servidor_importa_el_SDK_de_MCP`,
`sin_red_·_instalacion,_disco,_el_paquete_raíz_y_los_ficheros_del_applet_skills_y_del_aviso_no_alcanzan_net,_net/http_ni_internal/httpx`
y `control:_el_recorrido_nombra_la_cadena_entera_y_solo_sigue_aristas_del_módulo`.

**§10, primera orden (56).**

- `TestUmbralesDelInforme` y 24. Los de dos modos: `tres-de-54-en-herramienta-y-cero-en-orden`,
  `tres-de-54-en-orden-y-cero-en-herramienta`, `dos-de-54-en-los-dos-modos`, `sin-activar-solo-en-herramienta`,
  `redaccion-no-leida-solo-en-herramienta`, `901-s-en-herramienta-y-900-en-orden`,
  `901-s-en-orden-y-900-en-herramienta`, `900-s-en-los-dos-modos` y
  `la-sin-binario-ni-servidor-fuera-de-toda-medida`. Los de antes: `tres-de-54`, `dos-de-54`,
  `tres-de-54-y-seis-sin-terminar`, `sin-activar-una`, `sin-activar-ninguna`, `una-sin-activar-con-expresion`,
  `redaccion-no-leida-una`, `redaccion-no-leida-ninguna`, `900-s-con-objetivo-900`, `901-s-con-objetivo-900`,
  `dos-de-30-de-haiku-4-5`, `sin-lista-ni-objetivo`, `objetivo-negativo`, `todas-las-de-sonnet-5-5-sin-medir` y
  `los-tres-motivos-en-su-orden`.
- `TestInformeEnDosModos` y 5: `una-serie-falla-solo-en-el-modo-herramienta`,
  `la-serie-de-la-eval-sin-binario-ni-servidor-falla`, `las-lineas-del-guion-del-informe-final`, `legal-core` y
  `sin-poder-saber-el-modo`.
- `TestJuzgarLasLlamadas` y 11: `bloque-satisfecho`, `terminos-por-boe-buscar`, `municipio-por-territorio-resolver`,
  `prefijo-del-agente`, `otra-norma-u-otro-bloque`, `resultado-de-error`, `sin-resultado`,
  `argumentos-que-no-convierten`, `prohibido-por-herramienta`, `orden-en-el-modo-herramienta` y
  `el-servidor-no-es-una-consulta`.
- `TestJuzgarSinBinarioNiServidor` y 4: `con-la-linea-y-sin-citas`, `sin-la-linea`,
  `con-la-etiqueta-y-sin-la-direccion-en-su-linea` y `con-la-linea-y-una-cita`.
- `TestLeerLasLlamadas` y 7, una por carpeta de §1: `con-prefijo`, `sin-prefijo`, `modo-orden`, `herramienta-ajena`,
  `error-por-is-error`, `error-por-ok-falso` y `sin-resultado`.

**§10, segunda orden (94).**

- `TestPlanEnDosModos` y 8: `textos`, `series`, `sesiones`, `prueba-de-red`, `sondeo` y `del-repositorio`, esta con
  `boe-legislacion` y `legal-core`.
- `TestSesionesPorModo` y 7: `cada-modo`; `binario-sin-ruta-absoluta`, con `vacio` y `relativo`; y
  `servidor-que-no-se-escribe`, con `su-ruta-es-un-directorio` y `binario-que-no-es-utf8`.
- `TestGuionDeLaSesion` y 4: `con-servidor`, `con-servidor-y-traza`, `sin-servidor` y `sin-servidor-con-traza`.
- `TestEvalsDelRepositorio` y 17: `formato`, `conjunto`, `conjunto-legal-core`, `linea-sin-consulta`,
  `normas-conocidas`, `grabado`, `grafo-previo`, `cobertura-del-esquema`, `avisos-del-esquema`, `avisos-de-la-skill`,
  `hallazgos-del-esquema`, `hallazgos-de-la-skill`, `prosa-de-la-skill`, `expresiones-de-la-skill`,
  `expresiones-en-los-bloques`, `expresiones-calibradas` y `ordenes-para-powershell`.
- `TestDefinicionDelJob` y 53. De primer nivel, 6: `del-repositorio`, `peor-caso`, `errores`, `segundo-disparo`,
  `estado-de-la-tanda` y `sinteticas`. `segundo-disparo` lleva 10: `anterior-que-mide-corre`,
  `anterior-que-mide-espera`, `anterior-sin-decidir-que-mide`, `anterior-sin-decidir-que-no-mide`,
  `anterior-que-no-mide`, `anterior-terminada`, `posterior-que-mide`, `sola`, `espera-agotada` y
  `consulta-que-falla`. `estado-de-la-tanda` lleva 5: `marca-en-success`, `marca-saltada`, `tanda-en-curso`,
  `sin-tanda` y `tanda-saltada`. `sinteticas` lleva 32: `la-del-contrato`, `cancela-la-que-corre`,
  `tope-con-los-modelos-del-env`, `tope-con-las-repeticiones-del-env`, `tope-que-cubre-el-peor-caso`,
  `tope-por-debajo-de-los-dos`, `tope-por-debajo-del-peor-caso`, `env-con-el-objetivo-escrito`,
  `env-sin-la-concurrencia`, `include-con-otra-skill`, `include-sin-legal-core`, `otro-objetivo`,
  `otra-concurrencia`, `matriz-sin-legal-core`, `sin-nombre`, `concurrency-de-flujo`, `sin-cancel-in-progress`,
  `decidir-sin-la-ejecucion`, `decidir-sin-su-id`, `grupo-sin-la-skill`, `grupo-sin-la-cabeza-del-evento`,
  `evals-con-otro-if`, `evals-sin-needs-tanda`, `marca-sin-su-if`, `marca-con-otro-nombre`,
  `tanda-sin-la-prueba-de-red`, `tanda-sin-la-salida-medir`, `tanda-sin-actions-read`, `tanda-con-concurrency`,
  `tanda-con-nombre`, `tanda-con-su-id-como-nombre` y `sin-tanda`.

**§10, tercera orden (59).**

- `TestComprobarElSondeo` y 41. Los tres de FR-084: `evals-sin-binario-ni-servidor-sola`,
  `evals-sin-binario-ni-servidor-con-otras-que-se-miden` y `evals-sin-binario-ni-servidor-dos-en-su-orden`. La
  credencial: `sin-la-variable`, `con-la-variable-vacia` y `sin-la-variable-y-con-un-argumento-que-no-vale`. Las
  repeticiones y la concurrencia, siete de cada una: `repeticiones-vacias`, `repeticiones-0`, `repeticiones--1`,
  `repeticiones-tres`, `repeticiones-2.5`, `repeticiones-1e3`, `repeticiones-99999999999999999999`, `concurrencia-0`,
  `concurrencia--1`, `concurrencia-tres`, `concurrencia-2.5`, `concurrencia-1e3`,
  `concurrencia-99999999999999999999` y `concurrencia-pedida`, más `concurrencia-del-job-en-boe-legislacion` y
  `concurrencia-del-job-en-legal-core`. Las evals: `evals-vacia`, `evals-una-cifra`, `evals-tres-cifras`,
  `evals-con-un-espacio`, `evals-coma-al-final`, `evals-otro-separador`, `evals-numero-repetido`,
  `evals-nombre-de-la-eval`, `evals-que-no-son-de-la-skill` y `evals-de-otra-skill`. La skill y el modelo:
  `skill-sin-la-forma-de-un-nombre`, `skill-sin-carpeta-de-evals`, `skill-con-un-fichero-mal-formado`,
  `skill-que-el-job-no-ejecuta`, `skill-y-evals-que-no-valen`, `modelo-vacio`, `modelo-sin-la-forma-de-un-id`,
  `varios-argumentos-a-la-vez` y `job-ilegible`.
- `TestSondear` y 7: `sin-la-credencial`, `con-la-credencial-vacia`, `un-argumento-que-no-vale`,
  `una-eval-sin-binario-ni-servidor`, `una-eval-sin-binario-ni-servidor-y-otra-que-se-mide`,
  `todas-las-sesiones-fallan` y `limite-de-uso-en-la-primera`.
- `TestGuionDelSondeo` y 8: `sin-cinco-argumentos`, `go-test-sale-con-0`, `go-test-sale-con-1`,
  `go-test-sale-con-1-y-deja-uso`, `go-test-sale-con-2`, `uso-sin-la-credencial`,
  `uso-con-dos-argumentos-que-no-valen` y `uso-con-una-eval-sin-binario-ni-servidor`.

### Una errata que ve este recuento

La subprueba `guarda_las_intercaladaes_y_las_peticiones_del_servidor,_que_no_son_la_respuesta`
(`internal/mcp/mcptest/sesion_test.go:464`) lleva una errata en su nombre: «intercaladaes». El test pasa y comprueba lo
que dice. T018 no la corrige: la tarea solo deja tocar los ficheros de test para añadir los que falten si una cifra de
cobertura queda bajo su umbral. Queda para quien lea el informe.

## 4. Lo que el hito no cambia, y los umbrales

Todo sale de `git diff --name-status main` sobre `6e4ae3a`.

- **No hay ningún ADR nuevo** (FR-090): `git diff --name-status main -- docs` no imprime nada, y `docs/ADR/` tiene los
  mismos 35 ficheros en `HEAD` y en `main`. La decisión del hito es la del ADR 0035.
- **Ningún spec, plan ni suite congelada de un hito anterior cambia** (FR-090): `git diff --quiet main -- specs
  ':!specs/015-h21-kitlegal-mcp-serve'` sale con 0, sobre 745 ficheros versionados. Dentro están los `spec.md`, los
  `plan.md`, las carpetas `aceptacion/` y los `gates/aceptacion-congelada.json` de H0 a H7.4.
- **De los guiones anteriores solo cambian las dos líneas de la lista de applets** (FR-090): `git diff --name-status
  main -- internal/app/testdata` solo nombra `M internal/app/testdata/script/argumentos.txtar`, de 48 guiones
  versionados, y su diff son las líneas 24 y 30 (§1). Ningún guion activado de H19, H7 o H7.1 cambia.
- **Los guiones del workflow no cambian** (FR-048): `git diff --name-status main -- scripts/workflow .specify
  scripts/hito.sh scripts/paso.sh scripts/claude-modelo.sh scripts/coste-run.sh` no imprime nada, sobre 65 ficheros
  versionados. De `scripts/` solo cambian `evals.sh` y `evals-sesion.sh`, que son del job. Que el
  `scripts/workflow/informe.sh` de hoy lee el informe de dos modos lo fija
  `TestInformeEnDosModos/las-lineas-del-guion-del-informe-final` (§3).
- **La tabla de fuentes no cambia**: `git diff --name-status main -- docs/SOURCES.md evidencias data web` no imprime
  nada, sobre 42 ficheros versionados de los tres primeros.
- **Los verbos de hoy dan el mismo `--describe`** (`schema-check`): de `schemas/` solo cambian `eval.yaml.json`, que no
  sale de `--describe`, y `servidor.json`, que es nuevo. Los ficheros de los verbos de `boe`, `territorio`, `graph` y
  `skills` son los de `main`, y `TestEsquemasPublicados` los compara byte a byte con lo que emite cada verbo: pasa en
  `make ci` y en §3, con `schemas`, `regenerados-coinciden` y `forma-canonica`. `internal/cli/describe.go` cambia
  (comparte su generador con los esquemas de las herramientas) sin cambiar lo que escribe.
- **Cada `SKILL.md` tiene menos de 300 líneas** (FR-036; SC-009): 297 `boe-legislacion` y 190 `legal-core`, las del
  prototipo (research V35). También lo comprueba `TestSkillsDelRepositorio` en `skills-check`, que falla con 300 o más.
- **Ningún umbral de FR-043 se ha rebajado, dejado sin decidir ni sumado entre modos** (FR-049):
  - `umbralDeExpresionesProhibidas = 0.05`, `umbralDeRespuestasSinActivar = 0` y
    `umbralDeRespuestasConRedaccionNoLeida = 0` (`internal/evals/umbrales.go:16-18`) son los mismos valores, en las
    mismas líneas, que en `main`. `git log -G 'umbralDe[A-Za-z]+ += ' main..HEAD` no nombra ningún commit.
  - La definición del job no mueve nada que decida: `git log -G
    'objetivo_de_duracion|MODELO_DE_EVALS:|REPETICIONES_DE_EVALS:|UMBRAL_DE_EVALS:|MODELOS_INFORMATIVOS' main..HEAD --
    .github/workflows/evals.yml` no nombra ningún commit. Siguen `objetivo_de_duracion: 900` en `boe-legislacion` y `0`
    en `legal-core`, `MODELO_DE_EVALS: claude-sonnet-5-5`, `MODELOS_INFORMATIVOS_DE_EVALS: claude-haiku-4-5-20251001`,
    `REPETICIONES_DE_EVALS: 3` y `UMBRAL_DE_EVALS: 2`. El diff de `evals.yml` (5 líneas añadidas y 3 quitadas, de T010)
    solo sube el tope de 122 a 240 minutos, con su comentario (FR-083).
  - En `umbralesDelInforme`, cada umbral es de un modo y lo nombra: `expresiones_prohibidas:<modelo>:<modo>` decide en
    el modelo que decide (`delModelo.modelo == e.ModeloQueDecide`); detrás de él y solo en ese modelo van
    `sin_activar:<modelo>:<modo>` y `redaccion_no_leida:<modelo>:<modo>`; y `duracion_de_las_sesiones:<modo>` lleva
    `Decide: true` en cada modo.
  - Los modos no se suman y el total de cada uno son sus 54 respuestas. `TestUmbralesDelInforme` lo ve fallar (§3):
    `tres-de-54-en-herramienta-y-cero-en-orden` y `tres-de-54-en-orden-y-cero-en-herramienta` dan `fallo` con el motivo
    `umbral expresiones_prohibidas:claude-sonnet-5-5:<modo>: 3 de 54 (5,6 %), y tiene que ser ≤ 5,0 %`, aunque sumados
    serían 3 de 108; `dos-de-54-en-los-dos-modos` da `aprobado`; `sin-activar-solo-en-herramienta`,
    `redaccion-no-leida-solo-en-herramienta`, `901-s-en-herramienta-y-900-en-orden` y
    `901-s-en-orden-y-900-en-herramienta` dan `fallo`; y `la-sin-binario-ni-servidor-fuera-de-toda-medida` deja sus
    sesiones fuera de toda medida, de todo total y de toda duración (FR-047).
- **La lista de expresiones no ha perdido ninguna** (FR-049): `git diff --quiet main --
  evals/boe-legislacion/expresiones-prohibidas.yaml schemas/expresiones-prohibidas.yaml.json` sale con 0, y `git log
  main..HEAD` no nombra ningún commit que toque la lista. Es la de `main`, byte a byte, y su calibrado sigue en 36, 11 y
  9 (`expresiones-calibradas`, §3).
- **plan.md tiene las ocho filas del cierre de FR-082** en «Controles de umbral», todas con `FR-043, SC-001`:
  - `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5-5:orden` y `…:herramienta` (≤ 5 %, ≤ 2 de 54);
  - `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden` y `…:herramienta` (0);
  - `evals:boe-legislacion:redaccion_no_leida:claude-sonnet-5-5:orden` y `…:herramienta` (0);
  - `evals:boe-legislacion:duracion_de_las_sesiones:orden` y `…:herramienta` (≤ 900 s).

  Cada nombre aparece una sola vez en plan.md. El umbral del modelo informativo no tiene fila, y el texto de la sección
  lo dice. Las otras 14 filas son los umbrales que se miden en `make ci`, cada una con su control (`ci:…`); entre ellas
  están las tres que FR-082 nombra: las 300 líneas de cada `SKILL.md`, los 512 bytes de las `instructions` y el conjunto
  exacto de herramientas.
- **SC-013**: `make ci` en verde, con `depguard` sin hallazgos y `schema-check` sin drift (FR-079); 0 importaciones del
  SDK de MCP fuera de `internal/mcp` (§3, escenario 9). La documentación de FR-060 y FR-061 es de T017 y la comprueba
  la revisión final.

## 5. Los supuestos S1 a S11 de research

Lo que el run ha podido medir, con su medida, y lo que no.

| # | Supuesto | Medido en el run |
|---|---|---|
| S1 | La v1.8.0 del SDK se comporta como la v1.7.0 en lo que usa el hito | **No se ha medido.** No hace falta: `go.mod` fija `github.com/modelcontextprotocol/go-sdk v1.7.0` y la v1.8.0 no entra en el hito. |
| S2 | `govulncheck` no encuentra vulnerabilidades alcanzables en `go-sdk` v1.7.0 ni en los seis módulos que arrastra | **Sí.** `make vuln`, dentro del `make ci` de esta tarea: `No vulnerabilities found.` y `Your code is affected by 0 vulnerabilities`. Dice además «0 vulnerabilities in packages you import and 1 vulnerability in modules you require»: con `-show verbose`, esa es `GO-2026-5970`, de `golang.org/x/text` v0.14.0, que no es ninguno de los siete, ya estaba en el `go.mod` de `main` y llega por `internal/evals` (`santhosh-tekuri/jsonschema`), no por el binario. |
| S3 | En las otras cinco plataformas de distribución el SDK enlaza los mismos siete módulos | **Sí.** `go list -deps ./cmd/kitlegal` con `GOOS`, `GOARCH` y `CGO_ENABLED=0` de cada una: los 7 de 7 en `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64` y `windows/arm64` (26, 26, 24, 24, 25 y 25 módulos de terceros en total). `TestDependenciasDelBinario` pasa en `make ci` y en el escenario 9 de §3. |
| S4 | Con `claude -p … --mcp-config <fichero>`, las herramientas están disponibles desde el primer turno como `mcp__kitlegal__<applet>_<verbo>`, y `--permission-mode bypassPermissions` las deja llamar sin preguntar | **No se ha medido.** Exige abrir una sesión con modelo, que un paso del run no abre (ADR 0032). Lo mide el job de cierre, en el modo herramienta. |
| S5 | La forma de una llamada a una herramienta MCP y de su resultado en el transcript `stream-json` | **No se ha medido.** Los transcripts de `leer-llamadas/` son sintéticos y traen la forma que el supuesto da por buena, con prefijo y sin él. Lo mide el job de cierre. |
| S6 | El validador del cliente de Claude Code acepta el esquema de salida de `--describe` y el sobre real | **No se ha medido** con ese validador. Lo medido es lo que cabe sin él: `TestEsquemasDeHerramienta/cada_esquema_vale_suelto_ante_un_validador` pasa, y en §5 y §6 el `structuredContent` de cada llamada es el sobre de su orden. Lo mide el job de cierre. |
| S7 | Claude Code no pasa su entorno entero al servidor que arranca | **No se ha medido.** No se depende de ello: el `servidor.json` de cada sesión lleva su `env` (§1). |
| S8 | El shell de la herramienta Bash de Claude Code, en el runner, no vuelve a poner en el `PATH` el directorio de `kitlegal` | **No se ha medido**: depende del runner. Lo medido es el juicio que lo hace inofensivo: en `TestJuzgarLasLlamadas/orden-en-el-modo-herramienta`, que pasa, la sesión que ejecuta `kitlegal` por la ruta de su binario no pasa, con el motivo `orden de kitlegal en una sesión sin kitlegal en el PATH: …`. |
| S9 | Las sesiones del modo herramienta tardan como las del modo orden | **No se ha medido.** Lo mide `duracion_de_las_sesiones:herramienta`, que decide, en el job de cierre. |
| S10 | Un trabajo de GitHub Actions en un runner alojado admite `timeout-minutes: 240` | **No se ha medido.** Lo que se mide en `make ci` es que el tope cubre el peor caso (`TestDefinicionDelJob/peor-caso` y `del-repositorio`, §3). Que la plataforma lo admite lo dirá el propio job. |
| S11 | Lo que el ADR 0035 dice de Codex, de Antigravity y de la app de Claude | **No se ha medido.** Es la prueba humana de SC-002, después de fusionar. |
