# H7.1 · Cierre de la Definition of Done

Cierre de T025 (FR-081, FR-083, FR-090, FR-095, FR-096; SC-013, SC-014) sobre `049753e` (T024, el último commit antes de
T025). La rama está al día con `main`: `git merge-base HEAD main` es la cabeza de `main`, `43fb293`, así que
`git diff --name-status main` da exactamente lo que el hito cambia. Las cuatro partes son las de la tarea: lo tocado en
esquemas y datos de prueba, la cobertura, el quickstart y los puntos de la Definition of Done que no aplican.

## 1. Esquemas y datos de prueba tocados en el hito

Lista para la capa 3 del informe final: cada fichero bajo `schemas/` o bajo cualquier `testdata/` que el hito crea o
modifica, con la tarea que lo tocó y su motivo. Sale de `git diff --name-status main`, filtrado a esos dos árboles, y la
tarea de cada fila, de `git log --format=%s main..HEAD -- <fichero>`. Son once filas: dos esquemas, una derivada y los
ocho guiones de H7 que nombra FR-080. Ningún fichero de esos árboles se retira ni se renombra (ninguna línea `D` ni
`R`), y ninguna grabación cambia: la derivada es un fichero nuevo, fuera de `internal/source/boe/testdata/`.

### Esquemas (`schemas/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| M | `schemas/grafo.json` | T014 `[datos]` | El contrato publicado de `graph check` (FR-012, FR-082): los argumentos de posición y opcionales `norma` y `bloques`, fuera de `required`; la descripción del verbo con «como mucho 50 hallazgos», y `data` pasa de una lista de hallazgos al objeto `{norma, bloques, version-obsoleta, fuente-caducada, omitidos, hallazgos}`, con las listas nunca `null` (contracts/applet-graph.md §1 y §3). Regenerado con `go test ./internal/app -run '^TestEsquemasPublicados$' -actualizar-esquemas`, no escrito a mano; `TestSalidaDelGrafoContraSchemas` valida contra él la salida real. `show` y `stats` no cambian. |
| M | `schemas/eval.yaml.json` | T018 `[datos]` | El formato común de eval gana lo que pide FR-054: `$defs.clase-de-hallazgo` (`{"enum": ["version-obsoleta"]}`), la propiedad `hallazgos` (`array`, `minItems: 1`), prohibida en el `else` de una eval que no activa la skill, como `avisos`, y `norma` opcional en `$defs.comando-comprobacion` (`required` sigue siendo `applet` y `verbo`). `avisos` y su enumerado no cambian. Los casos de `internal/evals/formato_test.go` lo fijan: «comprobacion-con-norma» se acepta; una clase fuera del enumerado, una lista vacía y `hallazgos` en una eval con `activa: false` se rechazan. |

Los demás esquemas de `main` no cambian.

### Datos de prueba (`testdata/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| A | `internal/app/testdata/derivadas/version-ulterior/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json` | T010 `[datos]` | La redacción C del e2e (research D23; contracts/arnes-e2e.md §1): la grabación de H4 del bloque `a21` de la Ley 39/2015 con `fecha_vigencia` `20260101` en lugar de `20161002` y, al final del texto, el párrafo `[Redacción sintética de prueba: versión ulterior derivada de la grabación de H4.]`, que lo marca como sintético. Solo cambia el `cuerpo`: la grabación del bloque no declara su longitud. La produjo un programa de un solo uso fuera del repositorio a partir de esa grabación, como `version-posterior` de H7; ningún valor de la fuente se escribió a mano. `TestGrabacionesDerivadas` (T010, T011) fija que, leída con `boe`, da el mismo `Articulo` que la grabación de H4 salvo la fecha y el párrafo. La usa la suite congelada para el paso 5 de FR-025 (FR-091). |
| M | `internal/app/testdata/script/h7-grafo-codigos.txtar` | T002 y T014 `[datos]` | T002: en «world.db que no es una base de datos», fuera los `cksum` y los `! exists` de los auxiliares, y el mensaje pierde `; no se modifica` (se quedan el 1, la clase `inesperado` y la ruta); fuera la sección «world.db que es un directorio» y lo que la cabecera decía de ella (FR-070, FR-071). T014: la `data` de `graph check` sin argumentos con la forma nueva; la sección de `graph check ine:28074`, que sigue saliendo con 2, no cambia. |
| M | `internal/app/testdata/script/h7-grafo-entrega-fallida.txtar` | T002 `[datos]` | En «world.db que no es una base de datos», la regex pierde `; no se modifica$` (se quedan el 0, `cmp stdout` y `una-linea-mas.sh`), y fuera la sección «world.db que es un directorio» (FR-070, FR-071). |
| M | `internal/app/testdata/script/h7-grafo-concurrencia.txtar` | T002 `[datos]` | Fuera `! stdout 'world\.db-nuevo'` y lo que la cabecera y los comentarios decían del temporal de publicación, que T003 retira (FR-071). |
| M | `internal/app/testdata/script/h7-grafo-version-obsoleta.txtar` | T002 y T014 `[datos]` | T002: fuera la sección del orden inverso (H7 US2.3) y lo que la cabecera decía de él (FR-027). T014: la forma nueva de `data`. |
| M | `internal/app/testdata/script/h7-grafo-fuente-caducada.txtar` | T014 `[datos]` | La forma nueva de `data` en sus cuatro `graph check --json` (con `"fuente-caducada":3` donde hay tres hallazgos), y `! stdout 'version-obsoleta'` pasa a `! stdout '"clase":"version-obsoleta"'`, porque la clave `version-obsoleta` de los totales está siempre en `data`. |
| M | `internal/app/testdata/script/h7-grafo-no-emiten.txtar` | T014 `[datos]` | La forma nueva de `data` del `graph check --json` sobre el grafo vacío. |
| M | `internal/app/testdata/script/h7-grafo-applet.txtar` | T014 `[datos]` | La descripción nueva de `check` en la ayuda de `graph`, con `\s+` entre palabras, y la forma nueva de `data`. |
| M | `internal/app/testdata/script/h7-grafo-show.txtar` | T002 y T017 `[datos]` | T002: en «Sin --json, la tabla mínima del kernel», fuera las aserciones de la tabla, con cada `exec`, `! stderr .` y `cp stdout *.txt` de `sin-texto.sh` intactos. T017: esa sección afirma la salida legible de contracts/applet-graph.md §5 —la primera línea de `show` del `Bloque`, del `BloqueVersion` y de la `Norma`, las cabeceras de `stats` y de `check` con T8 y el grupo `fuente-caducada`— y `! stdout '^(fuente|url|fecha_consulta|hash)\s'`, y la cabecera dice la salida legible (FR-060 a FR-063). |

Cada cambio de un guion de H7 quita o adapta aserciones de lo que el hito retira o cambia (FR-080), en tareas `[datos]`
separadas del código (T002 antes de T003, T004 y T008; T014, indivisible con el contrato de `check`; T017 tras T016).

### Lo que no está en la lista

- **La eval 19**: `evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml` (M, T022) no está bajo
  `schemas/` ni `testdata/`: gana `norma: BOE-A-2015-10565` en su `graph check` y `hallazgos: [version-obsoleta]`, con
  `grafo_previo` sin cambios e `informativa: true`. Valida contra el esquema ampliado en `TestEvalsDelRepositorio`.
- **El esquema de `world.db`**: `internal/graph/migraciones/0002_lecturas.sql` (A, T007) es la migración embebida en
  el binario (data-model §1), código del adaptador y no un contrato de `schemas/`.
- **Material de los tests nuevos**: el dominio se prueba en memoria, y el almacén, la medida (`TestMedidaDelGrafo`,
  `TestLoQueLeeLaSkill`) y la preparación de la eval 19 siembran sus grafos en `t.TempDir()`; los casos nuevos del
  informe de evals son constantes escritas en `t.TempDir()`. Nada de eso es un fichero bajo `testdata/`.
- **Corpus de fuzz**: el hito no añade ninguna función `Fuzz` ni ningún fichero en un `testdata/fuzz/`.
- **Copias momentáneas**: las `zz-` de la suite congelada (T001, la medida del avance y el §6 de abajo), los mutantes
  de T011 y T023 y las sondas se crearon y se retiraron dentro de su tarea: ningún commit de la rama las contiene
  (`git log --name-only --format= main..HEAD` no nombra ningún `zz-`) y el directorio de guiones no tiene ninguna.
- **`data/` y `evidencias/`** no cambian (`git diff --quiet main -- data/ evidencias/` sale con 0).

## 2. Cobertura (Definition of Done, punto 9; FR-096)

Medida con `go tool cover -func` sobre los dos perfiles del `make ci` de esta tarea: `coverage.out` (`make test`) y
`coverage-integration.out` (`make test-integration`). Cada subtotal sale de la misma orden sobre el perfil filtrado a
las líneas del árbol, con su cabecera `mode:`; el recuento de sentencias, de ese mismo perfil filtrado. «Unión» es la
concatenación de los dos perfiles, que `go tool cover` funde bloque a bloque, como hace Codecov con los dos que publica
CI.

| Árbol | `coverage.out` | `coverage-integration.out` | Unión | Umbral (`codecov.yml`) |
|---|---|---|---|---|
| Global | 97,2 % | 97,4 % | 97,4 % | ≥ 70 % |
| `internal/core/**` | 98,6 % (2545 de 2581 sentencias) | 98,6 % (2545 de 2581) | 98,6 % | ≥ 85 % |
| `internal/cli/**` | 98,7 % (372 de 377) | 98,7 % | 98,7 % | ≥ 90 % |
| `internal/core/grafo` | 94,4 % (424 de 449) | 94,4 % | 94,4 % | — |
| `internal/graph` | 90,1 % (591 de 656) | 90,4 % (593 de 656) | 90,4 % | — |
| `internal/app` | 94,0 % | 94,0 % | 94,0 % | — |
| `internal/evals` | 99,0 % | 99,0 % | 99,0 % | — |
| `internal/skills` | 98,5 % | 98,5 % | 98,5 % | — |

Los tres umbrales de proyecto se cumplen: no hace falta ningún test y T025 no toca ningún fichero de test.

Frente al cierre de H7 (`specs/010-h7-internal-graph-grafo/datos-tocados.md`), `internal/core/**` baja de 99,6 % a
98,6 %, `internal/core/grafo` de 100,0 % a 94,4 % e `internal/graph` de 92,2 % a 90,1 %. Es lo que pide el hito: T003,
T004 y T005 retiran los tests de los estados a los que el binario no llega, y el código que los rechaza se queda como
regla genérica, sin test propio (FR-073, FR-078; research D20, D21). Lo que `go tool cover -func` da por debajo del 100 %
en esos dos paquetes son ramas de error de ese tipo: fallos de SQLite, de `os.OpenFile` y de las migraciones, fechas de
consulta que no son RFC 3339, un `fecha_vigencia` que no es cadena y la validación del lote contra lo que ningún emisor
produce.

**El diff (`patch`, `target: auto`).** FR-096 pide además el diff sin retroceder respecto de la base, y eso lo mide
Codecov en la propuesta de cambio (quickstart.md §9), por líneas. Una estimación local con el mismo criterio —cada
línea de un bloque de la unión de los dos perfiles, sobre las líneas añadidas o cambiadas de los `.go` de producto de
`git diff -U0 main`— da **91,20 %: 508 de 557 líneas**, sin líneas parciales. La base es la cobertura del proyecto en
`main`, que al cierre de H7 era del 97,4 % en sentencias, así que el estado `codecov/patch` de la propuesta
probablemente salga por debajo de su objetivo. Las 49 líneas que ningún perfil ejecuta son todas ramas de error de la
regla genérica, y ninguna tarea puede añadirles un test (FR-073, constitución, «Gates»):

| Fichero | Líneas sin ejecutar | Qué son |
|---|---|---|
| `internal/graph/aplicar.go` | 16 | `os.OpenFile` y `Close` de `world.db` sobre lo que no es un fichero (un directorio en su sitio: regla genérica, T003), y los errores de SQLite al leer y escribir las filas de `lecturas` y al buscar la redacción de partida, que la entrega propaga |
| `internal/app/grafo.go` | 8 | los errores de la lectura y de la salida legible que los tres verbos propagan, y el de `check` sobre lo que ninguna entrega guarda |
| `internal/core/grafo/comprobar.go` | 7 | una fecha de consulta que no es RFC 3339 y un `fecha_vigencia` que no es cadena |
| `internal/graph/lectura.go` | 6 | una lista de bloques que `encoding/json` no puede codificar y un error de SQLite al leer las filas de `lecturas` |
| `internal/graph/abrir.go` | 4 | las migraciones embebidas ilegibles, un fallo al abrir la conexión y una versión de esquema negativa |
| `internal/app/grafo_legible.go` | 4 | un dato o un valor que `encoding/json` no puede codificar, venido de un JSON ya leído |
| `internal/core/grafo/lecturas.go` | 2 | una fecha de consulta que no es RFC 3339 |
| `internal/evals/avisos.go` | 2 | una etiqueta sin palabras, que ninguna de las fijas es |

Queda como supuesto `[alcance]` de T025 en `gates/supuestos.md`: si `codecov/patch` sale rojo, lo decide quien lea el
informe; ningún umbral se toca.

## 3. Quickstart (escenarios 0 a 8)

Ejecutados sobre `049753e` tal como los escribe `quickstart.md`, con el binario de §0 en un directorio de `mktemp -d`
y `KITLEGAL_CACHE_DIR` dentro de él. El escenario 9 es el `make ci` de la verificación de esta tarea, en verde. Todos
dan lo esperado.

| Escenario | Resultado |
|---|---|
| 0. Binario y caché temporales | `go build` sale con 0. |
| 1. Ayuda de `check` | `Usage: graph check [<norma> [<bloques> ...]] [flags]` y la descripción de contracts/applet-graph.md §1, con «como mucho 50 hallazgos»; las ayudas de `[<norma>]` y `[<bloques> ...]`; `código 0`. |
| 2. Norma desconocida sin grafo | `código 0`, `"ok":true` y `"data":{"norma":"BOE-A-2099-99999","bloques":[],"version-obsoleta":0,"fuente-caducada":0,"omitidos":0,"hallazgos":[]}`; después, `no se ha creado nada`. |
| 3. Argumentos sin forma | Las tres con `código 2` y, en la salida de error, `argumentos inválidos: la norma "a21" no tiene la forma BOE-A-<año>-<número>, con cuatro dígitos en el año y de uno a nueve en el número`, lo mismo con `""`, y `argumentos inválidos: el bloque " " está vacío o solo tiene espacio en blanco`. |
| 4. Legible sobre el grafo vacío | `El grafo del mundo tiene 0 nodos, 0 aristas y 0 textos.`; `No hay nada que volver a comprobar en todo lo consultado.`, una línea en blanco y `Para acotar la comprobación a una norma y a sus bloques: kitlegal graph check <norma> [<bloque>...]`. Ninguna línea empieza por `fuente`, `url`, `fecha_consulta` ni `hash`, y el directorio de la caché sigue sin existir. |
| 5. Regla genérica | `código 1` las dos veces, con `grafo: "<ruta>/world.db" no es una base de datos utilizable: file is not a database (26)` en la salida de error y, con `--json`, `"data":{"clase":"inesperado",…}` con ese mismo mensaje. |
| 6. Entrega con las grabaciones | Ver abajo. |
| 7. Lo retirado no está | `retirado:` de `internal/graph/publicar.go`, `internal/graph/publicar_test.go` e `internal/graph/integracion_enlace_test.go`, y `coincidencias: 1`: la búsqueda no encuentra nada en `internal/graph` ni en `internal/core/grafo` (FR-095, SC-013). |
| 8. Skill, eval y esquemas | `make skills-check` y `make schema-check` salen con 0, y los subtests `hallazgos-del-esquema` y `hallazgos-de-la-skill` de `TestEvalsDelRepositorio` pasan (`--- PASS` los dos, contados); `SKILL.md` de `boe-legislacion` tiene 250 líneas (< 300); la forma fija `⚠ REDACCIÓN MODIFICADA:` está en dos líneas (177, la regla, y 182, el ejemplo); la eval 19 lleva `hallazgos:` (l. 27) y `- version-obsoleta` (l. 28). |

### Escenario 6

- **La suite congelada, todavía sin activar.** `go test -count=1 -run '^TestEntregaDelHito$/^h7-1-' ./internal/app/`
  sale con `ok … [no tests to run]`: `TestEntregaDelHito` solo lee `internal/app/testdata/script/`, y los `h7-1-*` no
  existen hasta que el workflow activa la suite tras el bucle. No es un resultado. Por eso los cuatro guiones de
  `aceptacion/` se copiaron un momento, sin cambiarlos (`cmp` sin diferencias), como `zz-grafo-lecturas.txtar`,
  `zz-grafo-check-acotado.txtar`, `zz-grafo-legible.txtar` y `zz-grafo-regla-generica.txtar`, y se ejecutaron con
  `rtk proxy go test -count=1 -v -run '^TestEntregaDelHito$/^zz-' ./internal/app/`: **cuatro subtests
  `--- PASS: TestEntregaDelHito/zz-…`, contados, ninguno `FAIL`, y `ok`** (`zz-grafo-regla-generica` 0,61 s,
  `zz-grafo-lecturas` 1,66 s, `zz-grafo-legible` 1,73 s, `zz-grafo-check-acotado` 1,79 s). Las copias se retiraron
  antes de `make ci`. Las huellas SHA-256 de los cuatro guiones de `aceptacion/` son las de
  `gates/aceptacion-congelada.json` (T001, `0da488f`). La ejecución de los `h7-1-*` activados la hace el workflow tras
  el bucle, con su `make ci`.
- **Los guiones de H7.** `^h7-grafo-`: once subtests `--- PASS`, ninguno `FAIL`, y `ok` —los ocho de FR-080 más
  `h7-grafo-matriz-territorial`, `h7-grafo-memoria` y `h7-grafo-no-interferencia`, sin cambios—.
- **La medida** (`-tags=integration`, `-v`). `TestMedidaDelGrafo` pasa y publica `graph check --json` sin argumentos
  en **34 591 bytes** (máximo 40 000; SC-001), con la norma en 10 813 bytes (máximo 40 000; FR-013) y con la norma y
  un bloque en 2 759 bytes (SC-002). Fuera de la orden del quickstart, `TestLoQueLeeLaSkill` pasa y publica los cinco
  bloques cambiados en **3 795 bytes** (máximo 3 800; SC-005) y, tras leerlos otra vez, 336 bytes: cinco bytes de
  margen.
- **`TestIntegracionGrafoDeH7`** (`-tags=integration`): `ok` (SC-012).
- **`internal/evals`**: `TestEtiquetasDeHallazgo`, `TestComprobarFormasDeHallazgo` y `TestJuzgar`, `--- PASS` los
  tres, y `ok` (SC-008).

## 4. Lo que la Definition of Done no pide aquí, comprobado

- **ADR nuevo: no hay** (FR-083). `docs/ADR/` tiene los mismos 28 ficheros que `main` y
  `git diff --quiet main -- docs/` sale con 0. El hito aplica ADR 0014 (el grafo en tres piezas), ADR 0026 (la salida
  legible en `Resultado.Legible`), ADR 0023 (los códigos: hallazgos con 0 en `data`, argumentos con 2, la regla
  genérica con 1) y ADR 0028 (el umbral de materialidad con el que retira de H7 lo que no lo pasa), sin reabrir
  ninguno.
- **`specs/010-h7-internal-graph-grafo/` intacto** (FR-083): `git diff --quiet main -- specs/010-h7-internal-graph-grafo`
  sale con 0.
- **Ningún otro guion cambia** (FR-081): `git diff --name-status main -- internal/app/testdata/script/` nombra
  exactamente los ocho guiones de FR-080 (ocho líneas `M`, ninguna más), ninguna copia `zz-`. Con ello la matriz
  territorial de H7 (`h7-grafo-matriz-territorial.txtar`, Definition of Done, punto 11) no cambia.
- **`docs/SOURCES.md`: no aplica.** El hito no toca ninguna fuente: `graph` es un applet calculado (procedencia
  `kitlegal.graph`, `kitlegal:applet/graph`) que solo lee `world.db` y no abre red, y `boe` y `territorio` emiten lo
  mismo que en H7, con su fuente de siempre. No hay manifiesto `grabaciones.json` nuevo, ni test `//go:build
  grabacion`, ni nada que grabar para el paso `grabar_datos`, ni material en `evidencias/`; la única derivada sale de
  una grabación de H4 ya versionada, sin red. `git diff --quiet main -- docs/SOURCES.md` sale con 0.
- **SC-014.** `make ci` en verde con `schema-check` y `skills-check` sin drift; la cobertura del proyecto dentro de los
  umbrales (§2; el diff, en el supuesto de T025); `SKILL.md` de `boe-legislacion` con 250 líneas; y la entrada de H7.1
  en `CHANGELOG.md`, bajo *Unreleased* (T024).
