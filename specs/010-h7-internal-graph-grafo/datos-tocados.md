# H7 · Esquemas y datos de prueba tocados en el hito

Lista para el informe final y su revisión humana posterior (constitución, capa 3; FR-095): cada fichero bajo
`schemas/` o bajo cualquier `testdata/` que el hito crea o modifica, con la tarea que lo tocó y su motivo. Sale de
`git diff --name-status main` sobre `6d01b81` (T028, el último commit antes de esta tarea), filtrado a esos dos
árboles, más los guiones que añadirá la activación de la suite. Ningún fichero de esos árboles se retira ni se renombra
(`git diff --name-status main...HEAD` no da ninguna línea `D` ni `R`), y ninguna grabación existente cambia: las cuatro
derivadas son ficheros nuevos, fuera de `internal/source/boe/testdata/`.

## Esquemas (`schemas/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| A | `schemas/grafo.json` | T016 `[datos]` | Contrato publicado de la salida del applet `graph` (entidad `grafo`, título `graph · grafo`, un `$defs` por verbo: `check`, `show` y `stats`; FR-050, FR-051; Definition of Done, punto 4). Generado con `TestEsquemasPublicados -actualizar-esquemas`, no escrito a mano: `make schema-check` lo compara con `--describe` y la fila `grafo.json` de `ficherosDeEsquemas` lo exige en cuanto el applet está registrado. T017 valida contra él la salida real de los tres verbos (`TestSalidaDelGrafoContraSchemas`): con hallazgos de las dos clases, sin hallazgos y con el grafo ausente. |
| M | `schemas/eval.yaml.json` | T022 `[datos]` | El formato común de eval gana `comando-comprobacion` (verbo `check`) en el `oneOf` de `comandos`, `prohibidos` (`comando-prohibido`, `minItems: 1`) y `grafo_previo` (`grabaciones` con el patrón de nombre y `comandos` de bloque); el `else` de las evals de no activación los rechaza como ya rechazaba `comandos`, `citas`, `avisos` y `territorio` (FR-086; contracts/evals-y-skill.md §1). `avisos` no cambia y ninguna eval existente deja de validar. Ningún mensaje que `internal/evals/formato_test.go` compare literalmente cambió, así que T022 no tocó ese test; el Go que lee las formas nuevas llegó en T023. |

Los demás esquemas de `main` no cambian.

## Datos de prueba (`testdata/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| A | `internal/app/testdata/derivadas/version-posterior/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json` | T019 `[datos]` | La derivada B del e2e (research D22; contracts/arnes-e2e.md §3): la grabación de H4 del bloque `a21` de la Ley 39/2015 con `fecha_vigencia` `20250101` en lugar de `20161002` y un párrafo más al final, `[Redacción sintética de prueba: versión posterior derivada de la grabación de H4.]`, que marca el texto como sintético. Solo cambia el `cuerpo` (esa grabación no lleva `Content-Length`). Da la segunda versión del bloque con la que `graph check` encuentra `version-obsoleta` (US2, SC-005). |
| A | `internal/app/testdata/derivadas/sin-eli/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_metadatos.json` | T019 `[datos]` | Los metadatos de la Ley 39/2015 de H4 con `url_eli` vacío, y el `Content-Length` recalculado (1363 → 1313). Con ellos, `boe articulo` no emite nada, ni con un id alternativo (FR-040; `grafo-no-emiten`). |
| A | `internal/app/testdata/derivadas/eli-sin-segmento/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_metadatos.json` | T019 `[datos]` | Los mismos metadatos con `url_eli` `https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565`, una URL sin ningún segmento `eli`, y el `Content-Length` recalculado (1363 → 1370): el segundo caso en que `boe articulo` no emite (FR-040; `grafo-no-emiten`). |
| A | `testdata/evals/grafo-previo/lpac-a21-version-anterior/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json` | T025 `[datos]` | El grafo previo de la eval 19 de `boe-legislacion` (FR-085; contracts/evals-y-skill.md §3): la grabación de H4 del bloque `a21` con `fecha_vigencia` `20151002` y el párrafo `[Redacción sintética de prueba: versión anterior derivada de la grabación de H4.]`. Solo cambia el `cuerpo`. La preparación de la sesión la entrega al grafo antes de llenar la caché, para que la consulta de la eval se encuentre con una versión anterior del bloque. |
| M | `internal/app/testdata/script/argumentos.txtar` | T016 `[datos]` | La lista literal de applets de las líneas 24 y 30 pasa a `boe, contar, echo, graph, skills, territorio`, porque registrar `graph` la cambia. Solo esas dos líneas: ninguna aserción nueva ni retirada. |
| M | `internal/app/testdata/script/territorio-matriz.txtar` | T016 `[datos]` | La línea 206 pasa de `! exists cache` a `! exists cache/cache.db`, y su comentario dice por qué (research D29): desde H7 el kernel crea el directorio de la caché para dejar en él `world.db` con lo que observa cada `territorio resolver`. Lo que la aserción protege, que `territorio` no abre la caché en ningún camino, se sigue afirmando sobre `cache.db`. Las líneas `cronometra` no cambian. |

Las cuatro derivadas las produjo un programa de un solo uso, fuera del repositorio, a partir de las grabaciones de H4
ya versionadas en `internal/source/boe/testdata/boe.legislacion-consolidada/`. Cada una conserva la forma y el nombre
de la grabación que sustituye y cambia exactamente lo que dice su nombre. Ningún valor de la fuente se escribió a mano. `TestGrabacionesDerivadas`
(`internal/app/grafo_test.go`; T020 para las tres del e2e, T026 para la del grafo previo) fija que cada una, leída con
`boe`, da el mismo `Articulo` o los mismos metadatos que su grabación de H4 salvo eso. La subprueba `grafo-previo` de
`TestEvalsDelRepositorio` (T026) prepara la del grafo previo en temporales sin ninguna falta.

Los dos guiones modificados cambian en la misma tarea que el registro de `graph` y de la entrega porque, por separado,
`make ci` queda en rojo (*Complexity Tracking* de plan.md).

## Guiones que añade la activación de la suite

Tras la última tarea, el paso `activar_aceptacion` del workflow (`scripts/workflow/aceptacion.sh activar H7`) copia
los 11 guiones congelados de `specs/010-h7-internal-graph-grafo/aceptacion/` a
`internal/app/testdata/script/h7-<nombre>.txtar` y los añade a la huella de `gates/aceptacion-congelada.json`. Hoy no
existe ningún `h7-*` en ese directorio. Los escribió T001 `[aceptacion]` desde el spec y ninguna tarea posterior los
ha tocado. Entrarán como ficheros nuevos, con los mismos bytes que el congelado, y desde ahí los ejecutará
`TestEntregaDelHito` en `make ci`. Lo que cubre cada uno es su línea «Cubre» de cabecera.

| Estado | Fichero | Cubre |
|---|---|---|
| A | `internal/app/testdata/script/h7-grafo-memoria.txtar` | US1.1 a US1.4 y el empate de fecha de consulta de «Edge Cases» · FR-001, FR-002, FR-021, FR-022, FR-023, FR-026, FR-030, FR-035, FR-040, FR-043, FR-053, FR-054, FR-065, FR-089, FR-093 · SC-001, SC-002 |
| A | `internal/app/testdata/script/h7-grafo-no-emiten.txtar` | US1.5 y el primer caso de «Edge Cases» · FR-030, FR-032, FR-034, FR-040, FR-046 · SC-004 |
| A | `internal/app/testdata/script/h7-grafo-version-obsoleta.txtar` | US2.1, US2.2, US2.3 · FR-060 a FR-064, FR-090, FR-093 · SC-005 |
| A | `internal/app/testdata/script/h7-grafo-fuente-caducada.txtar` | US2.4, US2.5 · FR-004, FR-023, FR-061, FR-065, FR-066, FR-067 |
| A | `internal/app/testdata/script/h7-grafo-no-interferencia.txtar` | US3.1 a US3.4, US3.7 · FR-005, FR-031, FR-034, FR-042, FR-091 · SC-003, SC-004 |
| A | `internal/app/testdata/script/h7-grafo-entrega-fallida.txtar` | US3.5, US3.6 · FR-011, FR-012, FR-033 · SC-011 |
| A | `internal/app/testdata/script/h7-grafo-codigos.txtar` | US4.3, US4.4, US6.4 · FR-004, FR-010, FR-011, FR-013, FR-031, FR-052, FR-054, FR-060, FR-094 · SC-011 |
| A | `internal/app/testdata/script/h7-grafo-show.txtar` | US4.1, US4.2 · FR-041, FR-053, FR-055, FR-070 · SC-006 |
| A | `internal/app/testdata/script/h7-grafo-concurrencia.txtar` | US6.3 · FR-014, FR-022 · SC-009 |
| A | `internal/app/testdata/script/h7-grafo-matriz-territorial.txtar` | FR-043, FR-045, FR-094 (matriz territorial en e2e) · constitución IX |
| A | `internal/app/testdata/script/h7-grafo-applet.txtar` | US4 · FR-031 (ayuda de `--no-graph`), FR-050, FR-051 |

## Lo que no está en la lista

- **La eval nueva**: `evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml` (A, T026) no está bajo
  `schemas/` ni `testdata/`. Es la eval informativa de la consulta repetida (FR-085 a FR-087), que lee el grafo previo
  de arriba. Valida contra el esquema ampliado y la revisa quien revise la skill.
- **El esquema de `world.db`**: `internal/graph/migraciones/0001_grafo.sql` (A, T007) es la migración embebida en el
  binario (data-model §3), código del adaptador y no un contrato de `schemas/`. La prueban los tests de
  `internal/graph` y la matriz de integración de T010.
- **Material de los tests nuevos**: el dominio (`internal/core/grafo`) se prueba en memoria, y el almacén, la matriz
  de integración, el applet, los costes y la preparación del grafo previo construyen sus bases, sus grafos y su
  derivada sintética en `t.TempDir()`. `TestNingunVerboDelGrafoDevuelveTexto` y `TestLaSalidaDeBoeNoCambiaConElGrafo`
  leen las grabaciones de H4 y H5.1 tal como están. Nada de eso es un fichero bajo `testdata/`.
- **Corpus de fuzz**: el hito no añade ninguna función `Fuzz` ni ficheros en ningún `testdata/fuzz/`.
- **Copias momentáneas**: las `zz-` de la suite congelada (verificación de T001 y medida del avance desde T020), los
  mutantes de T017 y T018 y la sonda de R6 de T007 (plan.md, «Sondas y copias momentáneas») se crearon y se
  retiraron dentro de su tarea. Ningún commit de la rama las contiene: `git log --name-only main..HEAD` no nombra
  ningún `zz-`.
- **`data/` y `evidencias/`** no cambian.

## Puntos de la Definition of Done que no aplican

- **ADR nuevo**: no aplica. ADR 0014 (`docs/ADR/0014-grafo-en-tres-piezas.md`) ya decide el grafo del mundo con ids
  naturales en `world.db`, las operaciones que viajan en el `Resultado` y que el kernel entrega con la `Procedencia`
  del propio resultado como fuente, sin cambiar el contrato `Applet`, y deja al plan de H7 los nombres concretos del
  campo y de los tipos. El hito lo implementa sin reabrirlo y no añade nada a `docs/ADR/`.
- **Fila de `docs/SOURCES.md`**: no aplica, porque el hito no toca ninguna fuente. `graph` es un applet calculado
  (procedencia `kitlegal.graph`, `kitlegal:applet/graph`) que solo lee el `world.db` local y no abre red. `boe` y
  `territorio` declaran lo que ya observaban, con su fuente de siempre: `boe.legislacion-consolidada`, con su fila
  desde H4, y `kitlegal.territorio`, los datos congelados que viajan en el binario. `docs/SOURCES.md` no cambia.
- **Grabaciones**: no aplica. No hay manifiesto `grabaciones.json`, ni test `//go:build grabacion`, ni paso
  `grabar_datos`, ni material en `evidencias/` (plan.md, «Datos externos»; research D32). Las cuatro derivadas salen de
  grabaciones de H4 ya versionadas, sin red (ver arriba).

## Cobertura (Definition of Done, punto 9; FR-094, SC-013)

Medida con el perfil unitario `coverage.out` que deja `make test` dentro del `make ci` que verifica esta tarea y
`go tool cover -func`. La única modificación de código de T029 está en `internal/graph/espera_test.go`, un fichero
`_test.go` (el origen de la medida de dos tests, tomado antes del plazo; `gates/tarea-T029.md`), y no cambia ninguna
sentencia de producto: el código medido es el de `6d01b81`, y las cifras del intento 1, medidas sobre su perfil, son
las mismas sobre el perfil del intento 2. Cada subtotal sale de la misma orden sobre ese perfil filtrado a las líneas
del árbol, con su cabecera `mode:`.

| Árbol | Cobertura | Umbral |
|---|---|---|
| `internal/core/grafo` (paquete nuevo) | 100,0 % | — |
| `internal/core/**` | 99,6 % (2531 de 2542 sentencias) | ≥ 85 % |
| `internal/graph` (adaptador nuevo) | 92,2 % | — |
| Global | 97,4 % | ≥ 70 % |

Las cifras valen para el código de producto de `6d01b81`, que es el del commit de T029. Si la activación de la suite
o la revisión final cambian código, hay que volver a medirlas sobre la cabeza final.

## Quickstart (escenarios 2 a 12)

Ejecutados sobre `6d01b81`, tal como los escribe `quickstart.md`, en una sola sesión de `sh` con la preparación de
«Antes de empezar». El escenario 1 es el `make ci` de la verificación de esta tarea. Todos dan lo esperado:

| Escenario | Resultado |
|---|---|
| 2. El binario recuerda | Las dos salidas de `graph stats` son idénticas: 5 nodos, 3 aristas y 1 texto, con `Bloque`, `BloqueVersion` y `Norma` de `boe.legislacion-consolidada`, `Municipio` y `Organo` de `kitlegal.territorio` y las tres relaciones con su fuente. En `cache/`, solo `cache.db` y `world.db`. |
| 3. `graph show` | El `Bloque` con `eli:has_part` entrante y `eli:has_version` saliente; la versión con `fecha_vigencia` `20161002` y su `hash_texto`; `ultima_observacion` con la procedencia del sobre (`2026-09-28T12:00:00Z` para `boe`, `2026-02-04T00:00:00Z` para `territorio`); `ine:28074` con `lb:pertenece_a` entrante desde `L01280745`; `0` coincidencias del texto; `código 3`. |
| 4. `version-obsoleta` | `[]` y después un único hallazgo sobre la versión de 2016, con `[BOE-A-2015-10565, bloque a21]`, `20161002`, `20250101`, la url del bloque y `2026-09-29T12:00:00Z`. |
| 5. `fuente-caducada` | Un único `fuente-caducada` (la versión de 2016, `604800 s`, caducada el `2026-10-05T12:00:00Z`) junto al `version-obsoleta`; sobre un directorio vacío, `[]` y `sin world.db`. |
| 6. No interferencia | `intacto` e `iguales`; ni `world.db-wal` ni `world.db-shm`. |
| 7. Entrega fallida | `código 0` y `misma salida` con una sola línea de aviso que nombra `world.db`; los dos `territorio resolver` salen con 0 y esa línea; `0` bytes con `--no-graph`; `$T/roto/world.db` intacto; con `world.db` en `0400`, `código 0`, `1` línea («no se puede escribir …world.db»), solo `world.db` y `sin cambios`. |
| 8. Códigos | `código 2` en las siete invocaciones mal formadas (U+00A0 y U+007F incluidos); `código 3` con `a b`; `código 1` con una base que no lo es y con un directorio; `código 2` con la variable vacía y `--no-graph`; con 0 bytes, ceros y listas vacías y `0` bytes después; sobre `$T/ro`, `código 0`, `misma lectura`, solo `world.db` y `sin cambios`. |
| 9. Ocho a la vez | `0` bytes de error, 8 `Municipio` y 8 `Organo` de `kitlegal.territorio`, solo `world.db` en `$T/ocho`. |
| 10. Binario distribuido y skill | `Municipio` y `Organo` de Tordesillas; `1`; `5` líneas con `kitlegal graph check`; `skills-check` en verde. |
| 11. Costes | `test-tiempos` en verde; SC-007, 0,95 ms de diferencia de mediana (máximo 150 ms); SC-008, `graph check` 165 ms (máximo 3 s) y `graph stats` 20 ms (máximo 1 s). |
| 12. Limpieza | `árbol intacto`. |

Dos desajustes del propio texto del quickstart, corregidos en ese fichero:

- **Escenario 10**: la ayuda de `boe articulo` parte la frase de `--no-graph` en dos líneas cuando no cabe en una, así
  que `grep -c` de la frase entera daba `0` y no `1`. El guion congelado `grafo-applet` ya lo tiene en cuenta: busca la
  frase con `\s+` entre palabras. La orden une las líneas con `tr -s ' \n' '  '` antes de buscar, y el esperado lo
  explica. La ayuda es la de contracts/resultado-y-entrega.md §6, carácter a carácter.
- **Escenario 11**: `TestCosteDelGrafo` publica sus medianas con `t.Logf`, que `go test` solo muestra con `-v`, y
  `make test-tiempos` no las enseñaba. Se añade una orden que ejecuta ese test solo, con `-v`, y filtra sus dos
  líneas `SC-007`/`SC-008`. El esperado nombra los máximos.

Dos observaciones ajenas a H7, que no contradicen nada de lo esperado:

- En el escenario 6, `boe articulo --dry-run` deja `cache.db-wal` y `cache.db-shm` en `cache/`, igual que el binario
  de `main` (comprobado con los dos binarios de e2e sobre la misma reproducción). Es el camino de ensayo de la caché
  de H3 y H4, no el grafo, y ningún escenario afirma nada sobre esos dos ficheros. Queda fuera del alcance del hito.
- En el escenario 10, el binario distribuido avisa en la salida de error de que las skills instaladas en la cuenta son
  de otra versión (el aviso de H19). Es un efecto del entorno de quien ejecuta el quickstart, no del hito.
