# Contrato: la suite de aceptación de H7.1, los guiones de H7 que cambian y los tests de medida

FR-080, FR-081, FR-091, FR-092, FR-026. Arnés: el de H7 (`specs/010-h7-internal-graph-grafo/contracts/arnes-e2e.md`,
`internal/app/e2e_test.go`), sin binarios ni relojes nuevos (research D23). Decisiones: [../research.md](../research.md)
D23, D24.

## 1. Lo que el arnés ya da y lo que gana

- Binarios de e2e: `kitlegal` (reloj del sistema), `$KITLEGAL_T0_BIN` (`2026-09-28T12:00:00Z`), `$KITLEGAL_T1_BIN`
  (`2026-09-29T12:00:00Z`), `$KITLEGAL_T8_BIN` (`2026-10-06T12:00:00Z`), con `boe` sobre la reproducción y la entrega al
  `world.db` de `KITLEGAL_CACHE_DIR=$WORK/cache`, que no existe al empezar.
- `$WORK/reproduccion/boe.legislacion-consolidada/`: las grabaciones de H4. `$WORK/derivadas/`: copia de
  `internal/app/testdata/derivadas/`, que gana **`version-ulterior/`** (research D23): el bloque a21 de la LPAC con
  `fecha_vigencia` `20260101` y el párrafo sintético «[Redacción sintética de prueba: versión ulterior derivada de la
  grabación de H4.]», con su huella; `TestGrabacionesDerivadas` la fija. Se pone en juego como las de H7:
  `cp derivadas/<caso>/<fichero> reproduccion/boe.legislacion-consolidada/` y `rm cache/cache.db` («caducada la
  caché»), conservando `world.db`.
- Redacciones del bloque `a21` de `BOE-A-2015-10565`: **A** = la grabación de H4 (`20161002`, `hash_texto`
  `sha256:98d3b9d4686a3155f48641841df7023e2c17a693fcd0e11d6beef2615abcec7c`), **B** = `version-posterior`
  (`20250101`), **C** = `version-ulterior` (`20260101`). Ids: `eli/es/l/2015/10/01/39` (Norma),
  `eli/es/l/2015/10/01/39#a21` (Bloque), `eli/es/l/2015/10/01/39#a21@<fecha>:<hash_texto>` (BloqueVersion). La huella de
  B y C no se escribe: se afirma con `sha256:[0-9a-f]{64}` o se toma de la salida de `boe articulo`.
- Segunda norma para el ámbito: `BOE-A-2017-12902`, bloque `a1-30` (grabación de H4 con metadatos y ELI). Su ELI no se
  escribe de memoria: se afirma por su ausencia en las salidas acotadas a la LPAC y por el recuento de hallazgos.
- Formas: las de H7 (`exec sh -c '…; test $? -eq N'` para códigos distintos de 0; `cmp stdout <fichero>`;
  `stdout -count=N`).

## 2. Regla de todos los guiones: la precondición en rojo

Cada guion empieza por:

```text
# Precondición: graph check se acota y lista como mucho 50 hallazgos (FR-007).
exec kitlegal graph check --help
stdout 'como\s+mucho\s+50\s+hallazgos'
```

Con el binario de hoy la ayuda de `check` no lo dice, así que el guion falla por esa aserción y no por una orden
desconocida (V19). `\s+` entre palabras porque la ayuda parte la descripción en líneas (V13).

## 3. La suite congelada (`specs/011-h7-1-graph-check-acotado/aceptacion/`)

**Aceptación e2e**: cuatro guiones testscript, que la tarea `[aceptacion]` escribe desde el spec y este contrato, y que
el workflow activa como `internal/app/testdata/script/h7-1-<nombre>.txtar` (los ejecuta `TestEntregaDelHito`; ninguno
cronometra). Cada guion nombra en su cabecera los FR/SC que cubre.

| Guion | Qué ejerce | Cubre |
|---|---|---|
| `grafo-lecturas.txtar` | La secuencia de FR-025 paso a paso (lecturas con T0, comprobaciones con T1): (1) A → `graph check BOE-A-2015-10565 a21 --json` sin hallazgos; (2) B → exactamente un `version-obsoleta` sobre A con `20161002` y `20250101`, y el mismo byte a byte al repetir; (3) `boe articulo … --no-graph` → el mismo; (4) la caché sirve B → ninguno, dos veces; (5) C → exactamente uno sobre B con `20250101` y `20260101`. Después, sin leer y con T8, `graph check --json` sin argumentos: `version-obsoleta` sobre B primero y `fuente-caducada` sobre la `Norma`, el `Bloque` y C, 4 en total (`"version-obsoleta":1,"fuente-caducada":3,"omitidos":0`), ninguno sobre A ni sobre B de `fuente-caducada`; y una lectura nueva con T8 (caché retirada, la fuente sirve C) → `graph check` sin ningún hallazgo | US2.1-US2.5, US4.1-US4.2 · FR-011, FR-014, FR-020, FR-023, FR-024, FR-025, FR-030, FR-031, FR-032 · SC-003, SC-004 |
| `grafo-check-acotado.txtar` | Con T0 se leen `a21` y `a22` de la LPAC y `a1-30` de `BOE-A-2017-12902`; con T8 (todo caducado): sin argumentos, 8 `fuente-caducada`; con la LPAC, 5, todos de ids de `eli/es/l/2015/10/01/39`; con la LPAC y `a21`, 3 (la `Norma`, `#a21` y su redacción), ninguno de `#a22`; con la LPAC, `a21` y `a99`, los mismos 3 y `"bloques":["a21","a99"]`; con `BOE-A-2017-12902`, 3 y ninguno de la LPAC; con `BOE-A-2099-99999`, 0 y `"hallazgos":[]`, `ok` `true`; con `a21` como norma, con `BOE-B-2015-10565`, con una norma vacía (`""`) y con la LPAC y un bloque `' '`, 2 y clase `argumentos`; la forma de `data` de contracts/applet-graph.md §3; la ayuda con `[<norma> [<bloques> ...]]` | US3.1-US3.4 · FR-001-FR-007, FR-011, FR-012 · SC-009 |
| `grafo-legible.txtar` | Con A leído (T0): `graph stats`, `graph show` del `Bloque` y `graph check` sin `--json` con las plantillas de contracts/applet-graph.md §5 —sin hallazgos con la norma y el bloque (T1), sin hallazgos y sin argumentos (T1, con la línea final de cómo acotar), con hallazgos y sin argumentos (T8) y con hallazgos y la norma (T8, sin la línea final)—; ninguna línea `^(fuente\|url\|fecha_consulta\|hash)\s`, ni `kitlegal.graph`, ni `kitlegal:applet/graph`, ni pares `nodo.`; ninguna línea de 20 caracteres o más del texto del bloque (el `sin-texto.sh` de H7); con `--json`, `stats` y `show` con la forma de H7; y `boe articulo` sin `--json` con la tabla mínima (FR-064) | US5.1-US5.5 · FR-060-FR-064 · SC-010 |
| `grafo-regla-generica.txtar` | Con un `world.db` que no es una base SQLite: `graph stats`, `graph show <id>`, `graph check` y `graph check BOE-A-2015-10565` salen con 1, clase `inesperado` y la ruta de `world.db` en el mensaje; `boe articulo … --json`, servido desde la caché, sale con 0, con la misma salida estándar que con `--no-graph` y exactamente una línea en la salida de error, `kitlegal: lo observado no ha llegado al grafo del mundo: grafo: "…world.db" no es una base de datos utilizable: …` | US6.1-US6.2 · FR-070 · SC-011 |

US1 (la skill) no tiene guion: su aceptación es la eval 19 en el job de evals (SC-006) y la prueba de la persona en
Claude Code (SC-007, S2). US6.3 (lo retirado no existe) la comprueban quickstart §7 y la revisión final.

## 4. Tests de Go que describen la entrega (no son de la suite congelada)

- **`TestMedidaDelGrafo`** (`internal/app/medida_test.go`, `//go:build integration`; FR-092, SC-001, SC-002). Siembra
  con `graph.Nuevo(graph.ConDirectorio(dir)).Apply` el grafo de la medida, con la forma de lo que emite `boe`:
  - 300 normas `i = 0…299`: identificador `BOE-A-2020-1<i con tres cifras>`, id `eli/es/l/2020/01/01/<i+1>`, bloques
    `a1`…`a8` (2 400), fuente `boe.legislacion-consolidada`, vigencia 604 800 s, url
    `https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/<identificador>/texto/bloque/a<k>`, cuerpos
    inventados con su huella.
  - Primera lectura de cada norma (un lote con sus 8 bloques y la redacción A, `20200101`): normas 0-269 el
    `2026-09-01T10:00:00Z`, normas 270-299 el `2026-10-05T10:00:00Z`.
  - Segunda lectura del bloque `a1` de las normas 0-239 (240 bloques), con la redacción B (`20250101`), el
    `2026-09-15T10:00:00Z`.
  - Con el binario de reloj T8, `graph check --json`: 50 hallazgos, todos `version-obsoleta` (los 50 primeros ids de
    las A de `a1` en bytes), `"version-obsoleta":240,"fuente-caducada":4590,"omitidos":4780`, `norma` `""`, y la
    salida estándar ≤ 40 000 bytes. `graph check BOE-A-2020-1000 --json` (solo la norma): sus 18 hallazgos —1
    `version-obsoleta` y 17 `fuente-caducada`: la `Norma`, sus 8 `Bloque` y sus 8 redacciones vistas—, todos de esa
    norma, en ≤ 40 000 bytes (FR-013). Y `graph check BOE-A-2020-1000 a1 --json`: exactamente 4 —`version-obsoleta`
    sobre la A de `a1` y `fuente-caducada` sobre la `Norma`, el `Bloque` `a1` y su B—, todos de ese ámbito. El binario
    es el de reloj T8 del arnés (`KITLEGAL_T8_BIN` de `entorno.variables`, `internal/app/e2e_test.go`), con
    `KITLEGAL_CACHE_DIR` en el directorio sembrado.
- **`TestIntegracionGrafoDeH7`** (`internal/graph/integracion_test.go`; FR-026, SC-012): entrega A (fecha
  `2026-09-28T12:00:00Z`) y B (`2026-09-29T12:00:00Z`) del mismo bloque; guarda el `Recuento`; lleva la base a lo que
  dejaba H7 (sin `lecturas` y sin la fila 2 de `schema_version`). Entonces: `Leer` + `Instantanea` + `Comprobar` no dan
  ningún `version-obsoleta`, y el `Recuento` es el guardado; una entrega de B otra vez no da ninguno; una de C
  (`20260101`) da uno sobre B; y la base ha pasado a la versión 2.
- **`TestCosteDelGrafo`** (`internal/app/coste_test.go`, H7 SC 007-008, se queda): el grafo grande entrega cada una de
  sus cuatro versiones por bloque en su propia entrega, como cuatro lecturas sucesivas, y comprueba los totales de la
  `data` nueva: `version-obsoleta` = bloques (1 750), `fuente-caducada` = normas + 2 × bloques (3 570), 50 listados.

## 5. Guiones de H7 que cambian (FR-080), en tareas `[datos]`

Solo esto; ningún otro guion cambia (FR-081). Cada cambio quita o adapta una aserción de lo que el hito retira o cambia.

| Guion | Cambio |
|---|---|
| `h7-grafo-codigos.txtar` | «world.db que no es una base de datos» (165-184): fuera los `cksum` y los `! exists` de auxiliares; en las tres regex del mensaje, `no es una base de datos utilizable; no se modifica` pasa a `no es una base de datos utilizable` (el mensaje ya no promete nada, research D19); se quedan el 1, la clase `inesperado` y la ruta. Fuera la sección «world.db que es un directorio» (186-197) y lo que su cabecera dice de ella. `"data":\[\]` de `check` (32, 52, 209) pasa a la forma de contracts/applet-graph.md §3 sin argumentos: `"data":\{"norma":"","bloques":\[\],"version-obsoleta":0,"fuente-caducada":0,"omitidos":0,"hallazgos":\[\]\}`. `graph check ine:28074` (85-92) no cambia: sigue saliendo con 2 y `argumentos inválidos` (research D6) |
| `h7-grafo-entrega-fallida.txtar` | En «world.db que no es una base de datos» (36-61): fuera `cksum` y `! exists` de auxiliares; la regex de la línea 43 pierde `; no se modifica$`. Fuera la sección «world.db que es un directorio» (63-73) y lo que la cabecera dice de ella. Se quedan el 0, `cmp stdout` y la línea que nombra `world.db` (`una-linea-mas.sh`) |
| `h7-grafo-concurrencia.txtar` | Fuera `! stdout 'world\.db-nuevo'` (42) y los comentarios del temporal (5-6, 36-37) |
| `h7-grafo-version-obsoleta.txtar` | Fuera la sección del orden inverso (75-121) y lo que la cabecera dice de US2.3. Con la regla nueva, el hallazgo sobre A no cambia de forma; la `data` de `check` (33, 57) pasa a la forma nueva (`"version-obsoleta":1,"fuente-caducada":0,"omitidos":0,"hallazgos":\[…\]`) |
| `h7-grafo-fuente-caducada.txtar` | La `data` de `check` (25, 38, 46, 67) a la forma nueva; los tres `fuente-caducada` (Norma, Bloque y la redacción vista) no cambian |
| `h7-grafo-no-emiten.txtar` | La `data` de `check` (80) a la forma nueva sin argumentos |
| `h7-grafo-applet.txtar` | La descripción de `check` en la ayuda (30) —con `\s+` entre palabras— y la `data` de `check` (100) |
| `h7-grafo-show.txtar` | La sección «Sin --json, la tabla mínima del kernel» (49-91): las aserciones de la tabla (`\Afuente +kitlegal\.graph\n`, `^url …`, `^nodo\.id …`…) pasan a las de la salida legible (contracts/applet-graph.md §5), y se conservan los `cp stdout *.txt` que usa `sin-texto.sh`; la cabecera deja de decir «la tabla mínima» |

`h7-grafo-memoria.txtar`, `h7-grafo-no-interferencia.txtar` y `h7-grafo-matriz-territorial.txtar` no cambian y tienen
que pasar tal cual (FR-022: sus `cksum` rodean solo verbos de `graph`, `--no-graph` y `--dry-run`, que no leen).
