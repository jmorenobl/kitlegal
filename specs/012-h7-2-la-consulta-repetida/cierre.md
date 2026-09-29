# H7.2 · Cierre de la Definition of Done

Cierre de T012 (FR-056, FR-062, FR-070, FR-080; SC-007, SC-008) sobre `c053d77` (T011, el último commit antes de
T012). La rama está al día con `main`: `git merge-base HEAD main` es la cabeza de `main`, `e523381`, así que
`git diff --name-status main` da exactamente lo que el hito cambia. Las cuatro partes son las de la tarea: lo tocado en
esquemas, datos de prueba y evals, la cobertura, el quickstart y lo que la Definition of Done no pide aquí.

## 1. Esquemas, datos de prueba y evals tocados en el hito

Lista para la capa 3 del informe final: cada fichero bajo `schemas/`, bajo cualquier `testdata/` o bajo `evals/` que el
hito crea, modifica o retira, con la tarea que lo tocó y su motivo. Sale de `git diff --name-status main`, filtrado a
esos tres árboles, y la tarea de cada fila, de `git log --format=%s main..HEAD -- <fichero>`. Son seis filas: un esquema
nuevo, una derivada nueva y una retirada, y en `evals/boe-legislacion/` la lista nueva, la eval 19 nueva y la retirada.
Ninguna línea `M` ni `R`: lo que ya existía en esos árboles o sigue igual o se retira entero.

### Esquemas (`schemas/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| A | `schemas/expresiones-prohibidas.yaml.json` | T001 `[datos]` | El formato de la lista de expresiones prohibidas de una skill (FR-050, FR-055; contracts/lista-y-juicio.md §1), tal cual el contrato: `$schema` 2020-12, `$id` `https://kitlegal.es/schemas/expresiones-prohibidas.yaml.json`, objeto con `additionalProperties: false` y `maquinaria` y `otra_conversacion` obligatorias, cada una una lista con `minItems: 1` de cadenas con el patrón `^[^\s*_]+( [^\s*_]+)*$` (sin blancos en los extremos, ni dobles, ni `*` ni `_`). No sale de `--describe`, así que `schema-check` no lo compara; lo fijan los casos de `internal/evals/formato_test.go` (una lista con las dos familias valida; sin una de ellas, con una vacía, con una clave de más y con una expresión con un blanco en un extremo o con `*`, no) y `LeerConjunto` lo compila una sola vez para validar la lista (T002). |

`schemas/eval.yaml.json`, el formato común de eval, no cambia (FR-056; §4). Los demás esquemas de `main`, tampoco.

### Datos de prueba (`testdata/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| A | `testdata/evals/grafo-previo/lcsp-a1-30-redaccion-original/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2017-12902_texto_bloque_a1-30.json` | T006 `[datos]` | El grafo previo de la eval 19 nueva (FR-010 a FR-013, SC-005): la grabación de H4 del bloque `a1-30` de la LCSP (`internal/source/boe/testdata/boe.legislacion-consolidada/…_a1-30.json`, 5 693 bytes, con las redacciones de vigencia 20180309 y 20200206) sin la redacción posterior, 3 418 bytes. La escribe la derivación de `internal/app/grafo_test.go` con `go test -count=1 -run '^TestGrabacionesDerivadas$' ./internal/app/ -args -actualizar-derivadas`, nunca una persona. `TestGrabacionesDerivadas` exige que sea, byte a byte, esa derivación sobre la grabación (comprobación 1) y que, servida con `boe`, dé el `Articulo` de la redacción original, igual campo a campo al que da la grabada reducida a ella: `fecha_vigencia` 20180309, `fecha_version` 20171109, `norma_modificadora` `BOE-A-2017-12902`, el texto original y su `hash_texto` (comprobación 2). Ninguna grabación nueva ni red (FR-014). |
| D | `testdata/evals/grafo-previo/lpac-a21-version-anterior/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json` | T008 `[datos]` | El grafo previo de la eval 19 retirada (5 733 bytes, de H7, #77): la grabación del art. 21 de la LPAC con `fecha_vigencia` 20151002 y un párrafo escrito a mano, una redacción que la grabada no trae (FR-020; SC-007). Con la carpeta salen su entrada de `grabacionesDerivadas()` y `parrafoDeLaVersionAnterior` de `internal/app/grafo_test.go`, y `TestGrabacionesDerivadas` pasa a comparar carpeta a carpeta, en los dos sentidos (FR-081). La carpeta, que solo tenía este fichero, deja de existir. |

Ninguna grabación de una fuente cambia: `internal/source/**/testdata/` y los `testdata/` del e2e no aparecen en el diff,
y las cuatro derivadas del e2e (`version-posterior`, `version-ulterior`, `sin-eli`, `eli-sin-segmento`) siguen con su
clase y su comprobación (FR-013).

### Evals (`evals/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| A | `evals/boe-legislacion/expresiones-prohibidas.yaml` | T003 | La lista de `boe-legislacion` con el contenido de contracts/lista-y-juicio.md §2: 38 expresiones, 22 de `maquinaria` y 16 de `otra_conversacion` (FR-050). Calibrada contra las 93 respuestas del cierre de H7.1: casa en 35, con el reparto de FR-084 (subtest `expresiones-calibradas`), y en ninguno de los bloques que leen las evals y sus grafos previos (`expresiones-en-los-bloques`) ni en las formas que la skill enseña a escribir (`expresiones-de-la-skill`) (FR-043, FR-085; SC-003, SC-004). No es un fichero de eval: `LeerConjunto` la reconoce por su nombre exacto. `Juzgar` la aplica a las evals que activan la skill (T004) y el informe publica lo encontrado por sesión y el recuento por modelo (T005). |
| A | `evals/boe-legislacion/19-lcsp-contrato-menor-redaccion-cambiada.yaml` | T007 | La eval nueva de la consulta repetida (FR-001 a FR-004, FR-030): `activa: true`, `informativa: true`, la pregunta literal de FR-001 sobre el art. 118 de la LCSP, `grafo_previo` con la derivada `lcsp-a1-30-redaccion-original` y el comando `boe BOE-A-2017-12902 a1-30`, los comandos del bloque y de `graph check` con `norma: BOE-A-2017-12902`, `graph show` prohibido, la cita `BOE-A-2017-12902 a1-30` y `hallazgos: [version-obsoleta]`. El subtest `grafo-previo` ejecuta en proceso lo que hace la sesión y da exactamente un `version-obsoleta`, de 20180309 a 20200206. |
| D | `evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml` | T007 | La eval 19 anterior, sobre el art. 21 de la LPAC, cuyo grafo previo sembraba una redacción inventada (FR-020). Los tests que la nombraban comprueban lo mismo con la nueva (FR-021), sin desactivar, saltar ni retirar ninguno. |

El conjunto queda en 19 evals, 17 que activan la skill y 10 positivas que deciden (las informativas son de la 13 a la
19), como exige el subtest `conjunto` (SC-007).

### Lo que no está en la lista

- **`skills/boe-legislacion/SKILL.md`** (M, T009) no está en esos árboles: es la skill v0.1.2, con 266 líneas (§4).
- **Material de los tests nuevos**: los casos nuevos del informe son copias en `t.TempDir()` de los casos versionados
  del paquete, que no cambian; las derivadas inventadas de `TestGrabacionesDerivadasInventadas` se arman en
  `t.TempDir()`; las sesiones de `TestCondicionesDeLaConsultaRepetida` y de `TestJuzgarLasExpresionesProhibidas` se
  construyen en el test. Nada de eso es un fichero bajo `testdata/`.
- **Corpus de fuzz**: el hito no añade ninguna función `Fuzz` ni ningún fichero en un `testdata/fuzz/`.
- **Mutantes y sondas momentáneos** (T003, T008, T009): se crearon y se retiraron dentro de su tarea. Los ficheros de
  los commits de la rama son los del diff neto, y `git log --name-only main..HEAD -- '*zz-*' '*.orig' '*mutante*'
  '*sonda*'` no nombra ninguno.
- **`data/` y `evidencias/`** no cambian (`git diff --quiet main -- data/ evidencias/` sale con 0).

## 2. Cobertura (Definition of Done, punto 9)

Medida con `go tool cover -func` sobre los dos perfiles del `make ci` de esta tarea: `coverage.out` (`make test`) y
`coverage-integration.out` (`make test-integration`). Cada subtotal sale de la misma orden sobre el perfil filtrado a
las líneas del árbol, con su cabecera `mode:`, y el recuento de sentencias, del mismo perfil filtrado, con cada bloque
una vez. «Unión» es la concatenación de los dos perfiles, que `go tool cover` funde bloque a bloque, como hace Codecov
con los dos que publica CI.

| Árbol | `coverage.out` | `coverage-integration.out` | Unión | Umbral (`codecov.yml`) |
|---|---|---|---|---|
| Global | 97,2 % (10 439 de 10 743 sentencias) | 97,5 % (10 470 de 10 743) | 97,5 % | ≥ 70 % |
| `internal/core/**` | 98,6 % (2 545 de 2 581) | 98,6 % (2 545 de 2 581) | 98,6 % | ≥ 85 % |
| `internal/cli/**` | 98,7 % (372 de 377) | 98,7 % | 98,7 % | ≥ 90 % |
| `internal/evals` | 99,0 % (2 180 de 2 203) | 99,0 % | 99,0 % | — |
| `internal/app` | 94,0 % (1 092 de 1 162) | 94,0 % | 94,0 % | — |
| `internal/skills` | 98,5 % (1 229 de 1 248) | 98,5 % | 98,5 % | — |
| `internal/graph` | 90,1 % (591 de 656) | 90,4 % (593 de 656) | 90,4 % | — |
| `internal/core/grafo` | 94,4 % (424 de 449) | 94,4 % | 94,4 % | — |

Los tres umbrales de proyecto se cumplen: no hace falta ningún test y T012 no toca ningún fichero de test. Los árboles
que no son `internal/evals` dan las mismas cifras que el cierre de H7.1 (el binario no cambia, §4); la global sube del
97,4 % al 97,5 % en `coverage-integration.out` y en la unión. En `internal/evals`,
los ficheros de producto del hito, en la unión: `consulta_repetida.go` 31 de 31, `juzgar.go` 189 de 189, `informe.go`
589 de 593, `conjunto.go` 257 de 260, `prohibidas.go` 29 de 30 y `formato.go` 24 de 25.

**El diff.** Nada en H7.2 pide una cifra del diff, pero Codecov lo mide en la propuesta de cambio con el estado
`patch` (`target: auto`, la cobertura de la base). Una estimación local con el criterio del cierre de H7.1 —cada línea
añadida de los `.go` de producto de `git diff -U0 main` que cae en un bloque de la unión de los dos perfiles— da
**95,36 %: 144 de 151 líneas**, sin líneas parciales, todas en `internal/evals`. La base, que el cierre de H7.1 midió
en el 97,4 % en sentencias, queda por encima, así que `codecov/patch` probablemente salga por debajo de su objetivo. Las
siete líneas que ningún perfil ejecuta son ramas de la regla genérica que ninguna estructura de directorio alcanza:

| Fichero | Líneas | Qué son |
|---|---|---|
| `internal/evals/conjunto.go` | 133-134, 160, 172-173 | El error al leer un fichero regular que `os.ReadDir` acaba de listar: el de la lista (`leerLista`) y el de una eval (`leerEntrada`), las dos por `leerContenido`, que T002 sacó de `leerEntrada`, donde en `main` estaba la misma rama, tampoco ejecutada. Sin permisos ni carreras no se llega (como root, un fichero sin permisos se lee igual), y el resultado ya es la regla que hay: un `FicheroMalFormado` que lo nombra. |
| `internal/evals/prohibidas.go` | 62-63 | El esquema publicado de la lista que no se puede leer o no compila: sale de una ruta fija del repositorio, compilada una sola vez, y es la misma rama que la del esquema de eval en `LeerEval` (`formato.go`, 262-263), sin ejecutar desde H5; `compilarEsquemaPublicado`, que las dos comparten, tiene sus tests. |

Queda como supuesto `[alcance]` de T012 en `gates/supuestos.md`: si `codecov/patch` sale rojo, lo decide quien lea el
informe; ningún umbral se toca.

## 3. Quickstart (escenarios 1 a 7)

Ejecutados sobre `c053d77` tal como los escribe `quickstart.md`, desde la raíz del repositorio. Las salidas de
`go test` se leyeron con `rtk proxy`, sin el resumen del proxy de la sesión, y los `--- PASS` se contaron. Todos dan lo
esperado.

| Escenario | Resultado |
|---|---|
| 1. La lista, la eval nueva y la skill | `go test` sale con 0: **15 `--- PASS`, contados, ningún `--- FAIL`** —`TestEvalsDelRepositorio` y sus catorce subtests, entre ellos los que nombra el quickstart: `formato`, `conjunto`, `grafo-previo`, `expresiones-calibradas`, `expresiones-en-los-bloques` y `expresiones-de-la-skill`; además `hallazgos-del-esquema`, `cobertura-del-esquema`, `hallazgos-de-la-skill`, `normas-conocidas`, `avisos-de-la-skill`, `avisos-del-esquema`, `conjunto-legal-core` y `grabado`—, y `ok`. `make skills-check` sale con 0 (`ok` en `internal/app`, `internal/skills` e `internal/evals`). |
| 2. La derivada del grafo previo | La primera orden sale con 0: **10 `--- PASS`, contados, ningún `--- FAIL`** —`TestGrabacionesDerivadas` con `sin-eli`, `eli-sin-segmento`, `version-posterior`, `version-ulterior` y `lcsp-a1-30-redaccion-original`, y `TestGrabacionesDerivadasInventadas` con `parrafo-inventado`, `fecha-inventada` y `fecha-y-parrafo-inventados` (las tres no pasan la comprobación 2, como deben)—. La segunda, con `-actualizar-derivadas`, da `ok` y reescribe la derivada. La tercera no imprime nada: **la derivada queda sin cambios** (`git hash-object` del fichero reescrito es `bfbdb11…`, el mismo blob que `HEAD`, y `git diff --quiet -- testdata/` sale con 0). |
| 3. El juicio, el informe y la comprobación | `ok  github.com/jmorenobl/kitlegal/internal/evals`. |
| 4. Lo retirado no está | `retirada la eval`, `retirado su grafo previo` y la búsqueda sin ninguna coincidencia: `git grep` sale con 1 sin imprimir nada (`coincidencias: 1`). La misma búsqueda con `lcsp-a1-30-redaccion-original` sí encuentra la eval 19 nueva, `internal/app/grafo_test.go` y tres tests de `internal/evals`, así que no pasa en vacío. Fuera de esos árboles, los nombres retirados solo están en `CHANGELOG.md` (la entrada que dice que se retiran), en `docs/ROADMAP.md` (la sección del hito) y en `specs/` (SC-007). |
| 5. La skill y el conjunto | `266` líneas (< 300); `19` evals; y en la eval 19, `pregunta:` en la línea 7, con la pregunta literal de FR-001, `grabaciones: lcsp-a1-30-redaccion-original` en la 11 y `- version-obsoleta` en la 30. |
| 6. La consulta repetida en Claude Code | **No se ejecuta**: abre dos conversaciones con `claude-sonnet-5`, así que necesita modelo y la credencial de la persona, y queda fuera de `make ci` y del run; la ejecuta la persona al leer el informe final (FR-062). Sus órdenes nombran lo que existe, comprobado abajo. |
| 7. El veredicto del repositorio | `make ci` en primer plano: `ci: todos los controles en verde` (la verificación de esta tarea). |

### Las órdenes del escenario 6 nombran lo que existe

- **La eval 19 nueva**: `evals/boe-legislacion/19-lcsp-contrato-menor-redaccion-cambiada.yaml` existe (§1 y escenario
  5), y es el fichero que la orden de la preparación pasa en `-eval`.
- **Los dos puntos de entrada**: `go test -tags evals -list '^(TestPrepararSesion|TestComprobarConsultaRepetida)$'
  ./internal/evals/` lista `TestPrepararSesion` y `TestComprobarConsultaRepetida`, y `ok`.
- **Sus banderas**: sin ninguna, cada uno falla antes de hacer nada y nombra todas las que exige, que son las que pasa
  el escenario: `TestPrepararSesion`, `falta la bandera -skill`, `-eval`, `-sesion` y `-modelo`;
  `TestComprobarConsultaRepetida`, `falta la bandera -skill`, `-primera`, `-segunda`, `-fecha-superada` y
  `-fecha-leida`. `exigirBanderas` las busca con `flag.Lookup`, así que cada una está registrada en el binario de test.
- **Los ficheros que el escenario lee y escribe**: la preparación escribe `pregunta.txt` y `cache/` en el directorio de
  la sesión (`internal/evals/preparar.go`), y `LeerSesion` lee `sesion.jsonl`, `codigo-de-la-sesion` y `sesion.err`
  (`internal/evals/sesion.go`), los nombres que usan las dos conversaciones del escenario.
- **`kitlegal skills install` con `--host`**: `go run ./cmd/kitlegal skills install --help` da `Usage: skills install
  [<skill> ...] [flags]`, con el argumento de posición `[<skill> ...]` y `--host=<host>` («claude o antigravity, que
  solo tiene directorio propio con -g; se puede repetir»): `boe-legislacion --host claude`, en ámbito local, es una
  invocación válida.

## 4. Lo que la Definition of Done no pide aquí, comprobado

- **ADR nuevo: no hay** (FR-070; DoD, punto 7). `docs/ADR/` tiene los mismos 28 ficheros que `main` y
  `git diff --quiet main -- docs/` sale con 0. Ninguna decisión de arquitectura cambia: el hito aplica ADR 0016 (el job
  de evals con Sonnet 5 que decide, en tres repeticiones), ADR 0014 (el grafo como índice y validador) y ADR 0028 (el
  umbral de materialidad), sin reabrir ninguno.
- **`specs/010-h7-internal-graph-grafo/` y `specs/011-h7-1-graph-check-acotado/` intactos** (FR-070):
  `git diff --quiet main -- specs/010-h7-internal-graph-grafo specs/011-h7-1-graph-check-acotado` sale con 0.
- **El binario no cambia.** `git diff --name-only main -- cmd internal ':!internal/evals' ':!*_test.go'` no imprime
  nada (la misma orden sin las exclusiones nombra solo `internal/app/grafo_test.go` y ficheros de `internal/evals`), y
  `go list -deps ./cmd/kitlegal` no nombra `internal/evals` (0 coincidencias entre los 16 paquetes del módulo que
  enlaza).
- **El esquema común de eval no cambia** (FR-056): `git diff --quiet main -- '*eval.yaml.json'` sale con 0 (el único
  fichero que casa es `schemas/eval.yaml.json`). `internal/evals` cambia solo en la lista (`prohibidas.go`, y su
  lectura, juicio e informe en `conjunto.go`, `formato.go`, `juzgar.go` e `informe.go`), en el punto de entrada de la
  comprobación del quickstart (`consulta_repetida.go` y `TestComprobarConsultaRepetida` en `job_test.go`) y en `doc.go`,
  que lo dice; la derivación real del grafo previo está en `internal/app/grafo_test.go`. `Juzgar` no cambia de firma.
- **`docs/SOURCES.md`: no aplica** (FR-014; DoD, punto 8). El hito no toca ninguna fuente ni graba nada: no hay
  manifiesto `grabaciones.json` nuevo, ni test `//go:build grabacion`, ni nada para el paso `grabar_datos`, ni material
  en `evidencias/`. La eval nueva usa las grabaciones de H4 de `BOE-A-2017-12902`, y la única derivada sale de una de
  ellas, ya versionada, sin red. `git diff --quiet main -- docs/` sale con 0.
- **La skill tiene menos de 300 líneas** (SC-008): `skills/boe-legislacion/SKILL.md`, 266.
- **SC-008.** `make ci` en verde con `schema-check` y `skills-check` sin drift y con las reglas del conjunto; la skill
  con 266 líneas; y la entrada del hito en `CHANGELOG.md`, bajo *Unreleased* (T011): `boe-legislacion` v0.1.2, la eval
  de la consulta repetida sobre el art. 118 de la LCSP y la lista de expresiones prohibidas.
