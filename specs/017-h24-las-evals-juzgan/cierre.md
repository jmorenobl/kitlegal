# H24 · Cierre de la Definition of Done

Cierre de T017 (FR-036, FR-045, FR-096, FR-100, FR-101; SC-013) sobre `b1c5156`, que es T016, el último commit antes de
T017. La rama está al día con `main`: `git merge-base main HEAD` da la cabeza de `main`, `6d9d947`, así que
`git diff --name-status main` da exactamente lo que cambia el hito. Las cinco partes son las de la tarea: lo creado y lo
modificado, la cobertura, el quickstart, los controles de umbral, y los umbrales y los atajos.

Todo lo que se cita aquí se ejecutó en la sesión de T017 —o en la de T018, donde se dice—, en primer plano, y se vio
terminar, o está en el repositorio. Lo que no se pudo medir se dice tal cual (§3, escenarios 7 a 9; §4, las diez filas
`evals:`).

**La constancia que T017 no pudo dar la da T018: el hito no añade ningún `//nolint`.** T017 encontró uno nuevo, el
`//nolint:misspell` que T015 dejó en `internal/evals/conjunto_test.go:2105`, y no lo retiró: los dos ficheros que había
que tocar no estaban entre sus rutas. T018 lo retira y neutraliza la palabra donde el repositorio neutraliza el
español, en `ignore-rules` de `misspell`, en `.golangci.yml`: los ficheros Go llevan 7 directivas `//nolint:` en `main`
y 7 en el árbol de T018 (§5, «Atajos»). Lo demás de (1) a (5) ya cuadraba en T017, que no tocó código ni añadió ningún
test.

De este fichero, T018 cambia solo lo que eso mueve, medido en su sesión sobre `d5ae2e7`, que es T017, con sus cambios
encima: lo que esta cabecera dice de T018, los recuentos de §1 con `.golangci.yml`, «Atajos» de §5 y el primer punto
que tenía §6. La cobertura, el quickstart y los controles siguen como los midió T017, sobre `b1c5156`.

El corrector de la revisión final (ronda 1) cambia solo lo que mueve su corrección del motivo `[k]`, medido en su
sesión sobre `81d025d` con sus cambios encima: en §3, el recuento de la primera orden de quickstart §4, que pasa de 49
a 47 porque `TestMedidaVersionada` pierde las dos subpruebas del fichero que deja de poder leerse (de 14 a 12;
`TestEjecucionDeLaMedida` sigue en 18, con una que sale y otra que entra); y en §4, las líneas de los seis tests de
`medida_test.go` y `conjunto_test.go` que cambian de sitio. El motivo `[l][f]` queda sin arreglar
(`gates/revision-pendiente.md`): los topes de 352 y 269 minutos y sus cifras siguen como estaban.

## 1. Lo creado y lo modificado en el hito

`git diff --name-status main`, en T018 y antes de tocar este fichero, da **90 ficheros: 54 `A` y 36 `M`**, ninguna `D`
ni `R` (21 897 líneas añadidas y 2 861 quitadas). 38 están en `specs/017-h24-las-evals-juzgan/`, todos `A`, con este
`cierre.md`. Los otros 52 son 16 `A` y 36 `M` (17 702 líneas añadidas y 2 861 quitadas).

T017 midió 88, 53 `A` y 35 `M`, antes de escribir este fichero: los dos de más son este y `.golangci.yml`, que T018
modifica. La orden no cuenta `gates/converge-hecho`, vacío y sin seguimiento: `git ls-files --others
--exclude-standard` da en T018 ese fichero y ningún otro, y T017 no vio ninguno.

### Los esquemas, la carpeta del juez, los dos ficheros restaurados, la lista y la skill

Son once de esos 52, con la tarea que los tocó (`git log --format=%s main..HEAD -- <ruta>`) y sus líneas
(`git diff --numstat main`):

| Estado | Fichero | Tarea | Líneas | Qué es |
|---|---|---|---|---|
| A | `schemas/juez-clases.yaml.json` | T001 `[datos]` | +23 | El esquema de la declaración de clases del juez. |
| M | `schemas/expresiones-prohibidas.yaml.json` | T015 `[datos]` | +7 −1 | La clave `salida_de_las_herramientas`, que entra en `required`; `additionalProperties: false` se queda. |
| A | `evals/boe-legislacion/juez/clases.yaml` | T001 | +11 | `afirma_lo_no_leido` (`decide: true`, `umbral: 0`) y `cuenta_su_proceso` (`decide: false`, `umbral: 0`). |
| A | `evals/boe-legislacion/juez/rubrica.md` | T001 | +53 | Copia de la evidencia del ADR 0037. |
| A | `evals/boe-legislacion/juez/esquema.json` | T001 | +1 | Copia. |
| A | `evals/boe-legislacion/juez/casos.yaml` | T001 | +1 559 | Copia: los 259 casos etiquetados. |
| A | `evals/boe-legislacion/juez/medida.json` | T001 | +28 | Copia: 0 de 212 y 0 de 47, `claude-opus-5-5`, `2.1.289`. |
| A | `testdata/evals/retiradas/19-lpac-articulo-21-redaccion-cambiada.yaml` | T008 `[datos]` | +28 | La eval 19 de H7.1, restaurada. |
| A | `testdata/evals/retiradas/grafo-previo/lpac-a21-version-anterior/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json` | T008 `[datos]` | +52 | Su grafo previo: una derivada de la grabación de H4, restaurada. |
| M | `evals/boe-legislacion/expresiones-prohibidas.yaml` | T015 | +14 −7 | La cabecera, que dice que ya no juzga ninguna respuesta, y la familia nueva con tres expresiones: `el sobre`, `fecha_vigencia` y `norma_modificadora`. |
| M | `skills/boe-legislacion/SKILL.md` | T015 | +19 −19 | v0.1.7: 298 líneas, las mismas que en `main`. |

Tres comprobaciones sobre ellos:

- **Los dos restaurados son los de la historia, byte a byte.** `git rev-parse` da el mismo objeto para
  `c4819d1^:evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml` y para el fichero de la cabeza
  (`6ce6f957…`), y el mismo para la derivada de `c4819d1^:testdata/evals/grafo-previo/lpac-a21-version-anterior/` y la
  de la cabeza (`410437a2…`). `c4819d1` es el commit de H7.2 que los retiró. La derivada la comprueba además
  `TestGrabacionesDerivadas/lpac-a21-version-anterior`, de `internal/app` (§3, escenario 4).
- **La skill es el diff del contrato y nada más.** `git apply --check -R` de
  `contracts/skill-boe-legislacion-v0.1.7.diff` termina con 0, y su `--numstat` da `19 19`, lo mismo que
  `git diff --numstat main -- skills/boe-legislacion/SKILL.md`.
- **La lista no pierde ninguna expresión.** En la cabeza tiene `maquinaria` 22, `otra_conversacion` 16, `anuncio` 39 y
  `redaccion_no_leida` 10 —las 87 del cierre de H7.4—, más las 3 de `salida_de_las_herramientas` y 2 formas fijas. El
  diff no quita ninguna línea de expresión: las 7 que quita son de la cabecera.

### Los otros 41

- `.golangci.yml` (T018): la entrada `informativo` de `ignore-rules` de `misspell`, con su comentario; +6 líneas y
  ninguna quitada (§5, «Atajos»).
- `internal/evals`, 31 ficheros: los nuevos `juez.go`, `medida.go` y `ejecucion.go` con sus tres `_test.go`, y 25
  modificados (`conjunto`, `consulta_repetida`, `definicion`, `formato`, `informe`, `juzgar`, `prohibidas`, `sesion`,
  `sondeo` y `umbrales` con sus tests, y `doc.go`, `grabaciones.go`, `preparar.go`, `job_test.go` y
  `sustitutos_test.go`).
- `internal/app/grafo_test.go` (T008): la entrada de la derivada restaurada en `TestGrabacionesDerivadas`. Es el único
  fichero de `cmd/` o de `internal/` fuera del paquete de evals que cambia.
- `scripts/evals-voto.sh` y `scripts/evals-medir-juez.sh`, nuevos, y `scripts/evals.sh` (T003, T011, T012).
- `Makefile` (T012), `.github/workflows/evals.yml` (T004, T013, T016), `CHANGELOG.md`, `CONTRIBUTING.md` y
  `docs/WORKFLOW.md` (T016).

### Lo que no cambia (FR-045, FR-096)

Cada comprobación es un `git diff --name-status main -- <rutas>`, ejecutado en esta sesión, que **no imprime nada**.
Entre paréntesis, los ficheros versionados que hay en esas rutas (`git ls-files`), para que el cero no sea el de una
ruta vacía:

- **Ninguna grabación**: `internal/source` (95), los dos manifiestos `grabaciones.json`
  (`internal/source/boe/testdata/` y `testdata/evals/`) y los dos `grabacion_test.go`. Bajo `testdata/` y `*/testdata/*`
  el diff solo nombra los dos ficheros restaurados de la tabla. El paso `grabar_datos` no tenía nada que grabar.
- **Ninguna fuente**: `internal/source`, `internal/httpx`, `internal/cache`, `docs/SOURCES.md` y
  `scripts/verify-sources.sh`.
- **Ningún manifiesto**: `'*grabaciones.json' '*manifiesto*' '*manifest*'` (11: los dos de las grabaciones, los de
  `evidencias/h6/` y `evidencias/adr-0037/`, los tres de `.specify/integrations/`, el esquema de `testdata/mcpb/`, el
  contrato de H19 y los dos `.go` del manifiesto de la instalación).
- **Ningún ADR**: `docs/ADR` (37). De `docs/`, el diff solo nombra `docs/WORKFLOW.md`.
- **Ningún spec, plan ni informe de H7.1 a H22**: las seis carpetas de `specs/011-…` a `specs/016-…` (356). Tampoco
  ningún otro spec: `-- specs ':!specs/017-h24-las-evals-juzgan'` (865).
- **El directorio de los guiones del workflow**: `scripts/workflow` (16); y con `.specify`, `scripts/hito.sh` y
  `scripts/paso.sh` (63), y con `.claude`, `.agents` y `skills-lock.json`. De `scripts/`, el diff solo nombra los tres
  guiones de las evals.
- **La carpeta de la evidencia del ADR 0037**: `evidencias` (28), con `evidencias/adr-0037` (13).
- **El binario y lo demás**: `data`, `web`, `cmd`, `pkg`, `mcp`, `plugin`, `README.md`, `go.mod`, `go.sum` y `tools`.
  `go list -deps ./cmd/kitlegal` no nombra `internal/evals` (0 líneas): lo único del producto que cambia es la skill
  que el binario empotra.

Las mismas órdenes sí imprimen donde hay cambios: con `-- internal/evals` salen 31 líneas; con `-- '*_test.go'`, 16; y
con `-- '*.json' ':!specs'`, 5. `git ls-files --others --exclude-standard` no da ningún fichero sin seguimiento.

### Las cuatro copias siguen idénticas a sus originales (FR-023, FR-045)

`cmp` de cada original de la evidencia del ADR 0037 con su copia de `evals/boe-legislacion/juez/` termina con 0 y sin
salida, las cuatro veces, y `shasum -a 256` da la misma huella a cada par:

| Fichero | SHA-256 del original y de la copia |
|---|---|
| `rubrica.md` | `5f1e2115b107d942ccdb4dd9b55f8f9606b1b56b52a2cc3e5a8fdef9692d06ee` |
| `esquema.json` | `b3a622159b449ae8fa64367083fada1d1a0bc73bb7cc56bb85f4a9da5c9df427` |
| `casos.yaml` | `4827894aefc4133f99d0af93672585f70c9d3f2b8803af36eb2cec4f1811401a` |
| `medida.json` | `eb1eff770b4a27137460fb00772faaedb635cec7be1d8ba7282e2e773d15c96f` |

`git log main..HEAD -- evals/boe-legislacion/juez` solo nombra T001 (`a63e750`): ninguna tarea posterior tocó la
carpeta. En `make ci` lo fija `TestCopiasDelJuez/del-repositorio` (§3, escenario 4).

### La Definition of Done, punto a punto (`ROADMAP.md` §1)

| Punto | Aplica | Qué se ve en la cabeza |
|---|---|---|
| 1 · `make ci` en verde | sí | §3, escenario 1: `ci: todos los controles en verde`, dos veces. |
| 2 · tests offline; fixtures si toca red | tests sí; fixtures no | Los tests de §3 y §4. El hito no toca la red: ninguna grabación (arriba). |
| 3 · sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | sí | En las líneas añadidas a los `.go` de `internal/evals` que no son tests, 0 coincidencias de `"net/http"`, `os.Exit`, `fmt.Print`, `os.Stdout`, `os.Stderr`, `panic(`, `"database/sql"` y `modernc.org` (la misma búsqueda da 55 con `fmt.Errorf`). `golangci-lint`: `0 issues.` |
| 4 · esquemas y `schema-check` | los dos esquemas tocados no salen de `--describe` | `make schema-check` en verde dentro de `make ci` (`ok` en `TestEsquemasPublicados`). |
| 5 · errores con código estable | ningún código de `kitlegal` cambia | `cmd` e `internal` fuera de `internal/evals`: solo `grafo_test.go`. |
| 6 · e2e y `CHANGELOG.md` | e2e no aplica (plan.md); `CHANGELOG.md` sí | `CHANGELOG.md` +69 (T016), con `boe-legislacion` v0.1.7 en *Unreleased* (línea 25). |
| 7 · ADR | no | `docs/ADR` sin cambios. |
| 8 · `SOURCES.md` | no | `docs/SOURCES.md` sin cambios. |
| 9 · cobertura | sí | §2. |
| 10 · skill: `SKILL.md` < 300, sin drift | sí | 298 líneas; `make skills-check` en verde (§3, escenario 5). Ninguna eval cambia: 21 de `boe-legislacion` en `main` y en la cabeza. |
| 11, 12, 13 · territorio, grafo, plazos | no | — |

## 2. Cobertura (Definition of Done §1.9)

Se mide sobre los dos perfiles que deja el segundo `make ci` de §3, escenario 1, en verde y con la caché de tests
vacía (`go clean -testcache` justo antes): `coverage.out` (`make test`) y `coverage-integration.out` (`make
test-integration`). El global sale de `go tool cover -func` sobre cada perfil; el de cada árbol o fichero, de `go tool
cover -func` sobre el perfil filtrado a sus líneas, con su cabecera `mode:`, escrito en el directorio temporal. Los
recuentos de sentencias salen de los bloques del perfil, contando cada bloque una vez. «Unión» son los dos perfiles
fundidos bloque a bloque, como hace Codecov con los dos que publica la CI.

| Árbol | `coverage.out` | `coverage-integration.out` | Unión | Umbral (`codecov.yml`) |
|---|---|---|---|---|
| Global | **97,1 %** (13 912 de 14 335 sentencias) | **97,5 %** (13 978 de 14 335) | 97,5 % | ≥ 70 % |
| `internal/core/**` (dominio) | **98,6 %** (2 545 de 2 581) | **98,6 %** (2 545 de 2 581) | 98,6 % | ≥ 85 % |
| `internal/cli/**` | 98,5 % (528 de 536) | 98,5 % (528 de 536) | 98,5 % | ≥ 90 % |
| `internal/evals` | 97,6 % (4 805 de 4 923) | 98,2 % (4 836 de 4 923) | 98,2 % | — |
| `internal/evals/juez.go` | **99,5 %** (379 de 381) | **99,5 %** (379 de 381) | 99,5 % | — |
| `internal/evals/medida.go` | **98,0 %** (451 de 460) | **98,0 %** (451 de 460) | 98,0 % | — |
| `internal/evals/ejecucion.go` | **100 %** (19 de 19) | **100 %** (19 de 19) | 100 % | — |

**Los umbrales se cumplen y no hace falta ningún test**: la tarea solo los pide si una cifra queda bajo su umbral, y
T017 no toca ningún `_test.go`. El hito no toca el dominio: `internal/core` no está en la lista de §1 y da las mismas
2 545 de 2 581 sentencias que los cierres de H21 y H22. En el perfil de integración, `-func` da 98,3 % al paquete de
evals y el recuento de bloques, 98,23 %; el cierre de H7.4 atribuyó esa diferencia a las funciones anónimas de nivel de
paquete, y aquí no se ha vuelto a comprobar.

**Por qué el segundo `make ci`.** El primero de la sesión, sin vaciar la caché, también terminó en verde, pero su
`test-integration` sacó 14 paquetes de la caché de tests —entre ellos `internal/cli` y los cinco de `internal/core`—,
así que su perfil no salía entero de tests ejecutados en la sesión. En el segundo, el registro no trae ningún
`(cached)`. Las cifras de `internal/core`, `internal/cli`, `internal/evals` y los tres ficheros fueron las mismas en
los dos; el global de la integración dio 13 974 en el primero y 13 978 en el segundo, cuatro sentencias de diferencia
fuera de esos árboles, que no se han buscado.

**Los tres ficheros, función a función.** Dan lo mismo en los dos perfiles:

| Fichero | Funciones | Las que no llegan al 100 % |
|---|---|---|
| `juez.go` | 34 | `leerClasesDelJuez` (91,7 %) y `dichosDelJuicio` (90,9 %) |
| `medida.go` | 41 | `baseVigente` (90,9 %), `textosDeLaSesion` (86,4 %), `prepararLaCacheDeLaSesion` (88,9 %), `registroDeLaSesion` (87,5 %) y `sobreSinElBloque` (96,0 %) |
| `ejecucion.go` | 3 (`ejecutarElJob`, `informe`, `pathDelJuez`) | ninguna |

Las once sentencias sin cubrir son, una a una, el `return` de una rama de error:

- `juez.go:243`: `esquemaDeClasesDelJuez()` devuelve un error.
- `juez.go:1051`: `json.Unmarshal` del juicio falla después de que el mismo juicio haya validado contra el esquema.
- `medida.go:794` y `798`: `retirarLaBase()` y `leerLasEvalsDeHoy()` fallan al preparar la base.
- `medida.go:860` y `870`: `os.MkdirTemp` y `os.Mkdir` fallan en el temporal de la sesión reconstruida.
- `medida.go:880`: `registroDeLaSesion` devuelve un error; dentro de ella, `922` (`registroDeBoe`) y `930`
  (`registro.Registrar`).
- `medida.go:906`: `copiarGrabaciones` no llega a la caché de la sesión.
- `medida.go:1096`: `json.Marshal` del sobre falla.

Ninguna tiene un test, y la tarea no pide añadirlo: son los estados que plan.md, «Trazabilidad», deja a la regla
genérica.

**El diff.** `codecov/patch` es bloqueante con `target: auto`, la cobertura de la base, y `main` no se midió en esta
sesión, así que no se estima. Lo que sí se midió: los tres ficheros nuevos suman 849 de 860 sentencias cubiertas
(98,7 %). Lo que diga `codecov/patch` solo se sabe en la propuesta de cambio.

## 3. Quickstart (§1 a §6)

Los seis escenarios se ejecutaron en primer plano, con sus órdenes tal cual, sobre `b1c5156` y con el árbol limpio
salvo los dos ficheros de `gates/` que el workflow lleva modificados (`tarea-actual.json` y `tareas-intentos.json`).
Fueron en su orden, y §1 dos veces: antes de §2 y, con la caché de tests vacía, después de §6 (§2, «Por qué el segundo
`make ci`»). **Todas las órdenes dan lo esperado**; una frase del quickstart no describe lo que la orden imprime
(abajo, «La medida no se imprime»).

| § | Orden | Esperado | Obtenido | ¿Cuadra? |
|---|---|---|---|---|
| 1 | `make ci` | termina con 0 | dos veces código 0 y `ci: todos los controles en verde`. En el segundo, con la caché vacía: registro de 100 líneas, 51 `ok`, ningún `FAIL` ni `(cached)`; `0 issues.` del lint; `No vulnerabilities found.` y `Your code is affected by 0 vulnerabilities.`; `ok` en `TestEsquemasPublicados` y en los tres paquetes de `skills-check`; `1 configuration file(s) validated`; `no leaks found`; `all modules verified`; `go mod tidy -diff` sin salida | sí |
| 2 | `go test -count=1 -run '^(TestTextosDeLaSesion\|TestMensajeDelVoto\|TestOrdenDelVoto\|TestFraseEnLaRespuesta\|TestVotoDelJuez\|TestReglaDeLosVotos)$' ./internal/evals/` | `ok  	github.com/jmorenobl/kitlegal/internal/evals` | `ok  	github.com/jmorenobl/kitlegal/internal/evals	6.332s` | sí |
| 3 | `go test -count=1 -run '^(TestInformeConElJuez\|TestUmbralesDelInforme\|TestEjecucionSinMedir\|TestJuzgarSinLaLista\|TestJuicioDelSondeo\|TestSalidaDelSondeo)$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	2.648s` | sí |
| 4 | `go test -count=1 -run '^(TestMedidaVersionada\|TestCopiasDelJuez\|TestGrabacionesDerivadas\|TestEjecucionDeLaMedida)$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	2.269s` | sí |
| 4 | `go test -count=1 -run '^TestGrabacionesDerivadas$' ./internal/app/` | `ok` | `ok  	…/internal/app	1.955s` | sí |
| 4 | los cuatro `cmp` | ninguna escribe nada y las cuatro terminan con 0 | las cuatro, con 0 y sin salida | sí |
| 4 | `shasum -a 256` de `rubrica.md` y `casos.yaml` de la carpeta del juez | `5f1e2115…06ee` y `4827894a…401a` | `5f1e2115b107d942ccdb4dd9b55f8f9606b1b56b52a2cc3e5a8fdef9692d06ee` y `4827894aefc4133f99d0af93672585f70c9d3f2b8803af36eb2cec4f1811401a` | sí |
| 4 | `jq -r '.rubrica.sha256, .casos.sha256, .modelo_del_juez, .version_de_claude_code, .defectos.sin_marcar, .correctos.marcados'` de `medida.json` | esas dos huellas, `claude-opus-5-5`, `2.1.289`, `0` y `0` | las seis líneas, en ese orden y con esos valores | sí |
| 5 | `make skills-check` | termina con 0 | código 0; `ok` en `internal/app`, `internal/skills` e `internal/evals` | sí |
| 5 | `wc -l skills/boe-legislacion/SKILL.md` | 298 o menos | `298` | sí |
| 5 | `go test -count=1 -run '^(TestProsaDeLaSkill\|TestEvalsDelRepositorio)$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	0.498s` | sí |
| 5 | `grep -n -o -e 'el sobre' -e 'fecha_vigencia' -e 'norma_modificadora' skills/boe-legislacion/SKILL.md \| sort -u` | `fecha_vigencia` en las líneas 240, 241 y 242, `norma_modificadora` en la 240 y la 241, y ninguna vez «el sobre» | `240:fecha_vigencia`, `240:norma_modificadora`, `241:fecha_vigencia`, `241:norma_modificadora`, `242:fecha_vigencia`; las tres líneas están en la tabla generada, que va de la 232 a la 257 | sí |
| 6 | `go test -count=1 -run '^TestDefinicionDelJob$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	0.355s` | sí |
| 6 | `grep -n -e 'MODELO_DEL_JUEZ' -e 'VERSION_DE_CLAUDE_CODE_DEL_JUEZ' -e 'evals-medir-juez' -e 'medir_al_juez' -e 'timeout-minutes' .github/workflows/evals.yml` | el modelo y la versión del juez en `evals` y `medida`; la etiqueta y la entrada en la condición de `medida`; los topes `352` y `269`; los de `cambios` y `tanda`, como están | `MODELO_DEL_JUEZ: claude-opus-5-5` en las líneas 197 (`evals`) y 372 (`medida`); `VERSION_DE_CLAUDE_CODE_DEL_JUEZ: 2.1.289` en la 213 y la 373; `if: github.event.label.name == 'evals-medir-juez' \|\| inputs.medir_al_juez == true` en la 359; `timeout-minutes: 352` en la 186 y `269` en la 368; `5` en `cambios` (44) y `15` en `tanda` (98), que son los de `main` (allí `evals` tenía 240) | sí |

**Los tests, por su nombre.** Cada orden de `go test` se ejecutó además con `-v`, para contar sus tests y subpruebas y
que ninguna pasara en vacío. Ninguna trae un `--- FAIL` ni un `--- SKIP`:

- §2, 75 `--- PASS`: `TestTextosDeLaSesion` (8 subpruebas), `TestMensajeDelVoto` (4), `TestOrdenDelVoto` (15),
  `TestFraseEnLaRespuesta` (14), `TestVotoDelJuez` (19) y `TestReglaDeLosVotos` (9).
- §3, 90: `TestInformeConElJuez` (13), `TestUmbralesDelInforme` (20), `TestEjecucionSinMedir` (11),
  `TestJuzgarSinLaLista` (2: `sin-binario-ni-servidor` y `calibrado`), `TestJuicioDelSondeo` (30) y
  `TestSalidaDelSondeo` (8).
- §4, primera orden, 47: `TestMedidaVersionada` (12), `TestCopiasDelJuez` (10), `TestGrabacionesDerivadas` (3:
  `respuestas`, `derivados` y `preguntas`) y `TestEjecucionDeLaMedida` (18). Segunda, 9: `TestGrabacionesDerivadas` de
  `internal/app` y sus ocho subpruebas, entre ellas `lpac-a21-version-anterior`, la derivada restaurada.
- §5, 42: `TestProsaDeLaSkill` (26) y `TestEvalsDelRepositorio` (14). Entre las catorce no están
  `expresiones-calibradas`, `expresiones-en-los-bloques` ni `expresiones-de-la-skill`, que el hito retira (FR-071).
- §6, 81: `TestDefinicionDelJob`, con seis subpruebas de primer nivel: `del-repositorio`, `sinteticas`, `peor-caso`,
  `errores`, `segundo-disparo` y `estado-de-la-tanda`.

**La medida no se imprime.** quickstart.md §4 dice que la primera orden «imprime en el test la medida con 0 de 212 y 0
de 47». La orden imprime solo su línea `ok`, y con `-v` tampoco sale ninguna línea que no sea de `go test`: la medida
no se imprime, se exige. Lo hace `TestEjecucionDeLaMedida/los-259-bien-con-una-medida-que-no-corresponde`
(`probarLos259Bien`, en `medida_test.go`), que compara el texto de la medida dada con el de 0 de 212 y 0 de 47 y exige
683 votos. Lo «esperado» de la orden, `ok`, se cumple; quickstart.md no se ha tocado. (El barrido global, posterior a
T018, corrigió esa frase de quickstart.md §4: ahora dice que el test exige la medida y que la orden no la imprime.)

**Cómo se ejecutó.** Tres cosas difieren de teclear el quickstart en una terminal, y ninguna cambia una orden:

- Las órdenes llevan delante `rtk proxy`, que da la salida sin resumir, y detrás un `&& echo …` que dice su código.
- Las salidas de `make ci` y de las ejecuciones con `-v` se guardaron en ficheros del directorio temporal, fuera del
  repositorio, para leerlas enteras.
- Los cuatro `cmp` de §4 fueron una sola línea, encadenados con `&&`.

`git status --porcelain` da lo mismo antes de §2 y después de §6: los dos ficheros de `gates/`.

**§7, §8 y §9, no ejecutados.** Ninguna sesión del run abre una sesión con modelo ni mide en la plataforma:

- §7, el sondeo (`make evals-sondeo`): abre sesiones con la suscripción de quien lo lanza. Es de una persona.
- §8, el job de cierre: lo lanza el workflow tras la revisión final, y sus órdenes leen `gates/evals/*.json`, que hoy no
  existe.
- §9, la medida con el código del job (SC-014): la lanza Jorge antes de fusionar, con la etiqueta `evals-medir-juez` o
  con la entrada `medir_al_juez`.

## 4. Controles de umbral (plan.md)

Las 22 filas de plan.md, «Controles de umbral»: 10 con un control `evals:` y 12 con uno o más `ci:`.

### Las doce filas `ci:`

Cada test se busca como lo busca el informe final (`estado_control`, en el guion del informe): `git grep -nE
'^(func <Test>\(|<Test>:)' HEAD -- <ruta>`. **Los quince controles distintos están en su ruta.** La misma búsqueda con
un nombre que no existe no da nada. Ninguno de sus siete ficheros lleva `//go:build`, el `-skip` de `make test` y de
`make test-integration` solo aparta `TestMedidasDeTiempo` y `TestCosteDelGrafo`, y el registro de `make ci` trae `ok`
para `internal/evals` en las dos órdenes y en `skills-check`.

| Fila de plan.md | Control `ci:` | Dónde está | Lo que el árbol da hoy (§3) |
|---|---|---|---|
| FR-007, SC-001 (sin juzgar) | `TestInformeConElJuez` | `internal/evals/informe_test.go:3842` | `PASS`, 13 subpruebas, entre ellas `un-voto-que-no-llega` y `dos-sin-juzgar` |
| FR-102, SC-002 | `TestVotoDelJuez` | `internal/evals/juez_test.go:1032` | `PASS`, 19 subpruebas |
| FR-103, SC-003 | `TestReglaDeLosVotos` | `internal/evals/juez_test.go:1337` | `PASS`, 9 subpruebas |
| FR-104, SC-004 | `TestInformeConElJuez` y `TestEjecucionSinMedir` | `internal/evals/informe_test.go:3842` y `internal/evals/ejecucion_test.go:102` | `PASS`; la segunda, 11 subpruebas, entre ellas `medida-que-corresponde` y las mutaciones |
| FR-042, FR-105, SC-005 | `TestMedidaVersionada` | `internal/evals/medida_test.go:259` | `PASS`, 12 subpruebas, entre ellas `del-repositorio` |
| FR-106, SC-006 | `TestEjecucionDeLaMedida` | `internal/evals/medida_test.go:2323` | `PASS`, 18 subpruebas, entre ellas `un-defecto-sin-marcar` y `un-correcto-marcado` |
| FR-107, SC-007 | `TestMensajeDelVoto` y `TestOrdenDelVoto` | `internal/evals/juez_test.go:280` y `:598` | `PASS`, 4 y 15 subpruebas |
| FR-023, FR-108, SC-008 | `TestCopiasDelJuez` | `internal/evals/medida_test.go:628` | `PASS`, 10 subpruebas: `del-repositorio`, las cuatro `cambiada-…`, las cuatro `sin-…` y `skill-fuera-de-la-tabla` |
| FR-109, SC-009 | `TestGrabacionesDerivadas` | `internal/evals/medida_test.go:1979` | `PASS`: `respuestas`, `derivados` y `preguntas` |
| FR-086, FR-110, SC-010 | `skills-check`, `TestProsaDeLaSkill` y `TestEvalsDelRepositorio` | `Makefile:126` (`skills-check: check-tools`); `internal/evals/conjunto_test.go:2071` y `:1395` | `make skills-check`, código 0; `PASS`, 26 y 14 subpruebas; `SKILL.md`, 298 líneas |
| FR-111, SC-011 | `TestJuzgarSinLaLista` | `internal/evals/juzgar_test.go:1053` | `PASS`: `sin-binario-ni-servidor` y `calibrado` |
| FR-092, FR-112, SC-012 | `TestDefinicionDelJob` | `internal/evals/definicion_test.go:173` | `PASS`, con `del-repositorio`, `sinteticas` y `peor-caso`; topes de 352 y 269 minutos en `evals.yml` |

Que cada control se pone en rojo lo vio su tarea, con sus casos negativos o con mutantes momentáneos que no quedan en
el diff (tasks.md, «Controles de umbral»): esta tarea no los repite ni añade ninguno.

### Las diez filas `evals:`

Su control es el informe del job de evals de `boe-legislacion`, que el workflow lanza en el cierre: **ninguna se mide
en el run**, y aquí no hay ninguna cifra suya. Lo que se ve en el árbol es que el código las construye con esos
nombres y con `decide: true` (§5), y que los tests de `make ci` las nombran:

| Filas | Control | Dónde se construye |
|---|---|---|
| FR-030 (dos modos) | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5:<modo>` | `umbralesDeLasRespuestas`, con el umbral y el `decide` de `clases.yaml` |
| FR-032 (defectos y correctos) | `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` y `…:correctos_marcados` | `umbralesDeLaMedida` |
| FR-033 (dos modos) | `evals:boe-legislacion:duracion_del_juez:<modo>` | `umbralesDeLaDuracion` |
| FR-034, sin activar (dos modos) | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:<modo>` | `umbralDeSinActivar` |
| FR-034, sesiones (dos modos) | `evals:boe-legislacion:duracion_de_las_sesiones:<modo>` | `umbralesDelInforme` |

Las cinco funciones están en `internal/evals/umbrales.go`. Los diez nombres están escritos, literales, en
`informe_test.go` y en `umbrales_test.go` (`grep -c -F` de cada uno da al menos 1 en los dos ficheros).

## 5. Umbrales y atajos (FR-036)

### Ningún umbral se rebajó, pasó a `decide: false` ni perdió respuestas o casos de su total

- **Los valores.** En `internal/evals/umbrales.go`: `umbralDeRespuestasSinActivar = 0` (el de `main`),
  `umbralDeLaMedidaDelJuez = 0` y `segundosDelJuezPorModo = 900`. El de `afirma_lo_no_leido` sale de `clases.yaml`:
  `umbral: 0`. `git log -G` de esas tres constantes solo nombra T005, que deja la primera sola con su valor, y T006,
  que añade las otras dos: ningún commit posterior las cambia. Y `clases.yaml` solo lo toca T001.
- **`decide`.** Los cuatro constructores de umbral que no dependen de una clase llevan `Decide: true` escrito
  (`umbrales.go:174`, `243`, `271`, `279` y `300`); el de cada clase lleva el de `clases.yaml`, donde
  `afirma_lo_no_leido` es `decide: true`. `cuenta_su_proceso` es `decide: false` desde T001, como pide FR-031: no es un
  umbral que haya dejado de decidir.
- **El total de las respuestas** es `len(grupo.respuestas)`, las respuestas medidas del modelo que decide en ese modo en
  las evals que activan la skill, el mismo en `sin_activar` y en cada clase: la respuesta sin juzgar sigue en él
  (`umbralesDeLasRespuestas`). Ninguna eval cambia ni sale: 21 de `boe-legislacion` en `main` y en la cabeza, y el diff
  de `evals/` solo nombra la lista y la carpeta del juez.
- **El total de los casos** son los 212 y los 47 de `medida.json`, y los casos, los 259 de `casos.yaml`: los dos son
  idénticos a sus originales (§1). La rúbrica y el esquema, también (FR-045).
- **La definición del job.** `git log -G 'objetivo_de_duracion|MODELO_DE_EVALS:|REPETICIONES_DE_EVALS:|UMBRAL_DE_EVALS:'
  main..HEAD -- .github/workflows/evals.yml` no nombra ningún commit: siguen `objetivo_de_duracion: 900`,
  `MODELO_DE_EVALS: claude-sonnet-5-5`, `REPETICIONES_DE_EVALS: 3`, `UMBRAL_DE_EVALS: 2` y `VERSION_DE_CLAUDE_CODE:
  2.1.284`. El modelo del juez y su versión los fijan T004 y T013 con los valores de la medida. El tope de `evals` sube
  de 240 a 352 minutos, y no baja ninguno.
- **Los dos que salen.** `umbrales` ya no lleva `expresiones_prohibidas:…` ni `redaccion_no_leida:…` (T005). No es una
  rebaja de FR-030, FR-032 ni FR-033: lo pide FR-034, que lo atribuye a la entrada del hito. En `umbrales.go` solo
  quedan nombrados en un comentario.
- **Esta sesión no lanzó la medida ni un sondeo** (FR-045): no ha ejecutado `claude`, `make evals`, `make evals-sondeo`,
  `make evals-medir-juez`, `TestEjecucionDelJob`, `TestMedidaDelJuez` ni `TestSondeo`. Lo que hicieron las sesiones de
  T001 a T016 no se puede constatar desde aquí; lo que sí se ve es que la medida versionada no ha cambiado desde T001 y
  es la de la evidencia (§1).

### Atajos: ningún `//nolint` nuevo, ningún `t.Skip` y ningún TODO

En las 15 524 líneas añadidas fuera de `specs/` (`git diff main -- . ':!specs'`, en T018; T017 contó 15 518, y las
seis de más son las de `.golangci.yml`), la búsqueda de `nolint`, `t.Skip`, `.Skip(`, `SkipNow`, `TODO`, `FIXME` y
`XXX` da tres líneas, y las tres son la plantilla `kitlegal-medida-del-juez.XXXXXX` de `mktemp`, en
`scripts/evals-medir-juez.sh` y en dos comentarios. T017 encontró una cuarta, que T018 retira:

```go
"carácter informativo.\n" //nolint:misspell // «informativo» es español: el párrafo va tal cual.
```

- **Dónde estaba y de quién era.** `internal/evals/conjunto_test.go:2105`, en `TestProsaDeLaSkill`: la última línea de
  la constante `vigenciaDeLaVersionAnterior`, que lleva byte a byte el párrafo del paso 5 de `SKILL.md` v0.1.6 y
  termina en «carácter informativo». `misspell` lee «informativo» como una errata de «information». La añadió T015
  (`816a43a`), con su supuesto en `gates/supuestos.md`, porque `.golangci.yml` no estaba entre sus rutas; T017 la dejó
  como supuesto `[alcance]`, por lo mismo. Contradecía plan.md, «Constraints» («ningún `//nolint`»), y la batería de
  tasks.md, «Sin atajos».
- **Primero, en rojo.** T018 quitó de esa línea la directiva y su motivo, y nada más: `git diff --numstat` da `1 1` a
  ese fichero, y lo que queda en la línea es el literal, `"carácter informativo.\n"`. `make lint` terminó entonces con
  error y una sola incidencia: ``internal/evals/conjunto_test.go:2105:15: `informativo` is a misspelling of
  `information` (misspell)``.
- **Después, donde el repositorio neutraliza el español.** `informativo` entra en `ignore-rules` de `misspell`, en
  `.golangci.yml`, entre `disposicion` e `inventario`, con un comentario como el de sus vecinas: qué palabra es, dónde
  está y cómo la lee `misspell`. `git diff --numstat` da `6 0` a ese fichero: la entrada y las cinco líneas de su
  comentario. Ninguna otra regla, exclusión ni entrada cambia. `make lint` terminó con 0 y `0 issues.`
- **El recuento.** `git grep -c '//nolint:' main -- '*.go'` y la misma orden sin `main`, sobre el árbol de T018, dan
  lo mismo: **7 directivas en `main` y 7 en el árbol**, en los mismos seis ficheros de test y en las mismas líneas
  (`coste_test.go`, `e2e_test.go`, con dos, y `tuberia_unix_test.go`, de `internal/app`; `internal/arch_test.go`;
  `internal/core/territorio/coste_test.go`; e `internal/evals/cierre_test.go`). `git grep -n nolint --
  internal/evals/conjunto_test.go` no da nada. T017 contó 7 y 8. **El hito no añade ningún `//nolint`.**
- **Lo que no cambia.** El párrafo sigue siendo el de v0.1.6: ni se recorta ni se parte la palabra en dos literales, y
  ningún caso del test cambia. `go test -count=1 -v -run '^TestProsaDeLaSkill$' ./internal/evals/` da `ok` y 27
  `--- PASS`, el test y sus 26 subpruebas de §3, sin ningún `--- FAIL` ni `--- SKIP`.
- **`make ci`.** Sobre el árbol de T018, antes de tocar este fichero: código 0 y `ci: todos los controles en verde`,
  con `0 issues.` del lint, 51 `ok` y ningún `FAIL`; 12 de esos `ok`, todos de `test-integration`, salieron de la caché
  de tests, y `internal/evals` no es uno de ellos. La verificación de T018 es otro `make ci`, posterior a este
  fichero: su resultado no está aquí.

`t.Skip`: los dos de la cabeza son los dos de `main` (`internal/cache/integracion_test.go:451` e
`internal/evals/cierre_test.go:70`). TODO y FIXME: ninguno en la cabeza ni en `main`, en `*.go`, `Makefile`,
`scripts/*.sh` y `.github`.

## 6. Lo que queda fuera del run

- **Las diez filas `evals:`** (§4) y SC-001: las mide el job de cierre que lanza el workflow.
- **`codecov/patch`** (§2): solo se sabe en la propuesta de cambio.
- **quickstart.md §7 y §9**, y SC-014: de una persona. Antes de fusionar, Jorge lanza la medida del juez con el código
  nuevo, y después lee las respuestas con algún voto afirmativo.
- **La verificación de esta tarea** es otro `make ci`, posterior a este fichero: su resultado no está aquí.
