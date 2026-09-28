# Contrato: el arnés e2e y los guiones de aceptación de H7

Lo que los guiones de aceptación pueden usar del arnés y de qué contrato sale cada formato que afirman. T001 los
escribe contra **este** contrato en `specs/010-h7-internal-graph-grafo/aceptacion/*.txtar`; quedan congelados; el
workflow los activa al final como `internal/app/testdata/script/h7-<nombre>.txtar` y los ejecuta
`TestEntregaDelHito` (o `TestMedidasDeTiempo` si cronometran; ninguno lo hace). Las tareas que implementan el arnés lo
cumplen tal cual. Decisiones: [../research.md](../research.md) D22-D24.

## 1. Regla de todos los guiones: la precondición en rojo

Cada guion empieza, antes de cualquier otra orden, por:

```text
# Precondición: el binario registra el applet graph (FR-050).
exec kitlegal --help
stdout '^  graph +\S'
```

Sin el applet registrado falla por esa aserción, que es lo que el rojo-primero exige; ningún guion puede fallar antes
por una orden desconocida o de uso. Detrás, cada guion puede usar todo lo de este contrato, que existirá al activarse.

## 2. Binarios

Todos son el binario de e2e (`internal/app/ejemplo/kitlegal-e2e`), con `boe` sobre la reproducción, `territorio`,
`skills`, los de ejemplo y **`graph`**, y con la entrega al grafo en el `world.db` de `KITLEGAL_CACHE_DIR`:

| En el guion | Reloj (`-X main.reloj`) | Qué fija |
|---|---|---|
| `kitlegal` (en el `PATH`), `$KITLEGAL_BIN` | ninguno: el del sistema | — |
| `$KITLEGAL_T0_BIN` | `2026-09-28T12:00:00Z` | la `fecha_consulta` de lo que `boe` pide a la reproducción y el instante de `graph` |
| `$KITLEGAL_T1_BIN` | `2026-09-29T12:00:00Z` | ídem |
| `$KITLEGAL_T8_BIN` | `2026-10-06T12:00:00Z` | ídem |

`reloj` es una variable de cadena de `main`, como `enlazador`: con un instante RFC 3339, la reproducción de `boe`
lleva `httpx.ConHora` en ese instante y `graph` recibe `DependenciasDelGrafoDelSistema()` con el `Reloj` sustituido;
sin valor, el binario se compone como el distribuido; con un valor que no es RFC 3339, el registro no se construye
(inesperado). Lo servido desde la caché conserva la `fecha_consulta` con la que se guardó (H4). Los binarios
`$KITLEGAL_V1_BIN`, `$KITLEGAL_V2_BIN` y `$KITLEGAL_SIN_ENLACES_BIN` de H19 siguen igual. `TestBinariosDelArnes` gana
los tres nuevos: cada uno responde `version` con `dev`, su `graph stats --json` sobre un directorio vacío lleva en
`fecha_consulta` su instante y deja el directorio vacío, y su `boe articulo BOE-A-2015-10565 a21 --json` servido
desde una copia de las grabaciones, también; los demás fechan con el reloj del sistema.

## 3. Directorio de trabajo

- `KITLEGAL_CACHE_DIR=$WORK/cache`, que no existe al empezar (los guiones lo afirman con `! exists cache`): ahí
  aparecen `cache.db` y `world.db`.
- `$WORK/reproduccion/boe.legislacion-consolidada/`: copia de las grabaciones de H4 (como hasta ahora).
- **`$WORK/derivadas/`** (nuevo): copia de `internal/app/testdata/derivadas/`, con las grabaciones derivadas y el
  nombre de la que sustituyen (research D22):
  - `derivadas/version-posterior/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json`
    — `fecha_vigencia` `20250101`;
  - `derivadas/sin-eli/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_metadatos.json`
    — `url_eli` vacío;
  - `derivadas/eli-sin-segmento/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_metadatos.json`
    — `url_eli` `https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565`.

  Un guion las pone en juego con `cp derivadas/<caso>/<fichero> reproduccion/boe.legislacion-consolidada/`, y para que
  se pidan de nuevo retira antes la caché con `rm cache/cache.db` (conservando `world.db`).

## 4. Formas que los guiones usan

- Bytes de `world.db`: `[unix] exec cksum cache/world.db` + `[unix] cp stdout antes.txt`, y después
  `[unix] cmp stdout antes.txt` (lo que depende de la salida de una orden `[unix]` lleva la misma condición);
  auxiliares: `! exists cache/world.db-wal`, `! exists cache/world.db-shm` y `! exists cache/world.db-journal`.
  Lo que hay en un directorio: `exec ls -A cache` y `stdout`/`! stdout` sobre sus nombres. Entre dos órdenes de un guion no hay
  ninguna conexión abierta, así que no quedan auxiliares y los verbos de `graph` leen en el estado normal, sin cambiar
  ningún byte (research D10, V9, V36 E); ningún guion construye un `-wal` huérfano: ese estado lo ejerce la matriz de
  integración (contracts/almacen-world-db.md §7).
- Quitar una variable: `[unix] exec env -u KITLEGAL_CACHE_DIR -u HOME kitlegal …` (`env K=` la deja presente y vacía,
  research V8).
- Códigos distintos de 0: `exec sh -c 'kitlegal …; test $? -eq N'`, como los guiones de H4-H6.
- «Exactamente una línea más en la salida de error»: un guion auxiliar dentro del propio `.txtar` que compara dos
  salidas de error guardadas con `cp stderr`, como `una-linea-mas.sh` de `h19-skills-aviso` (en H7, el de
  `grafo-entrega-fallida`, invocado con `exec sh $WORK/una-linea-mas.sh <sin> <con>`).
- Los caracteres de un id que no son ASCII van literales dentro de las comillas de `exec sh -c '…'`; el tabulador y
  U+007F, que son ASCII, los escribe `printf` (`"$(printf "\t")"`, `"$(printf "\177")"`).
- Concurrencia: `exec kitlegal territorio resolver <código> --json &` ocho veces y `wait`, que falla si alguna falló
  (research V8).
- Enlace multicall: `symlink graph -> <ruta del binario>` y `exec ./graph stats --json`, como `multicall.txtar`.
- `--describe` de `show` lleva su argumento (`kitlegal graph show x --describe`): sin el argumento obligatorio el
  análisis termina en 2 antes de describir, como `territorio resolver --describe` (`territorio-matriz.txtar:184`).

## 5. Formatos que los guiones afirman, y de dónde salen

| Qué | Contrato |
|---|---|
| Sobre de `graph`: `fuente` `kitlegal.graph`, `url` `kitlegal:applet/graph`; claves de `data` de `show`, `stats` y `check` | [applet-graph.md](./applet-graph.md) §2-§3 |
| Plantillas de las explicaciones y el instante de caducidad | [applet-graph.md](./applet-graph.md) §5 |
| Códigos de `graph` | [applet-graph.md](./applet-graph.md) §4 |
| Ids, tipos, datos y aristas emitidos; DIR3 de Leganés `L01280745`; fecha de `territorio` `2026-02-04T00:00:00Z` | [emision.md](./emision.md); `data/territorio/` |
| La línea de aviso de una entrega fallida | [resultado-y-entrega.md](./resultado-y-entrega.md) §4 |
| La ayuda de `--no-graph` (buscada con `\s+` entre sus palabras, porque la ayuda la parte en dos líneas) | [resultado-y-entrega.md](./resultado-y-entrega.md) §6 |
| Mensajes que nombran `world.db` | [almacen-world-db.md](./almacen-world-db.md) §6 |

Ningún valor se escribe de memoria: identificadores, `hash_texto`, fechas de vigencia y DIR3 salen de las grabaciones,
de las derivadas y de `data/territorio/`.

## 6. Los once guiones

| Guion | Qué ejerce |
|---|---|
| `grafo-memoria` | US1.1-US1.4: `boe articulo` (T0) y `territorio resolver`, `graph stats` con los recuentos y las fuentes, repetición sin cambios, `graph show` de cada nodo y arista con la procedencia del sobre, `--offline` que también entrega (`a23` guardado antes con `--no-graph`); el empate de FR-023 en los dos órdenes: `a21` y `a23` con T0 dejan en la `Norma` la url del bloque `a21`, la menor, y `a22` y `a21` con T0 (en ese orden, sobre una caché nueva), también |
| `grafo-no-emiten` | US1.5: `url_eli` vacío y sin `eli` → misma salida y código que con `--no-graph`, con y sin `--json`, stderr vacío, sin `world.db`, también con `articulos`; `indice`, `buscar`, `metadatos`, `analisis`, `skills install/list/doctor`, los tres de `graph`, una invocación fallida (`boe`, 3; `territorio`, 2) y `--dry-run` de `articulo`, `articulos` y `territorio resolver` no crean `world.db` ni ningún auxiliar |
| `grafo-version-obsoleta` | US2.1-US2.3: A (grabación, T0) y B (derivada, T1) → un hallazgo `version-obsoleta` sobre A con la cita, `20161002`, `20250101`, la url del bloque y `2026-09-29T12:00:00Z`; una sola versión → ninguno; orden inverso en otro directorio de caché (B con T1 y A con T0) → el mismo hallazgo byte a byte; y B con T0 y A con T1 en un tercero → el hallazgo sigue sobre A, con la procedencia de B (`2026-09-28T12:00:00Z`) |
| `grafo-fuente-caducada` | US2.4-US2.5: sin `world.db` → `[]` y sigue sin existir; A en T0, `check` con T1 → ninguno; con T8 → `fuente-caducada` en la `Norma`, el `Bloque` y la versión con `604800`, `604800 s` y `2026-10-05T12:00:00Z`, ninguno en `Municipio` ni `Organo`; una lectura nueva con T8 lo renueva y `check` vuelve a `[]` |
| `grafo-no-interferencia` | US3.1-US3.4, US3.7: `boe articulo` con y sin `--no-graph`, con y sin `--json`, byte a byte y mismo código; `--no-graph` y `--dry-run` dejan `world.db` con los mismos bytes o sin crear; `graph stats`, `check` y `show` con y sin `--no-graph` iguales (T0) sin cambiar `world.db`; con `KITLEGAL_CACHE_DIR` vacía, `graph stats` sale con 2 con `--no-graph` y sin ella |
| `grafo-entrega-fallida` | US3.5-US3.6: `world.db` que no es base, o que es un directorio → `boe articulo` desde la caché con la misma salida y código y una línea más que nombra `world.db`, fichero intacto; `KITLEGAL_CACHE_DIR` vacía (con la que `boe articulo` sale con 2 sin línea de grafo), que nombra un fichero, y sin ella con `HOME` ausente o vacía → `territorio resolver` sale con 0, misma salida que con `--no-graph`, una línea más, nada creado; con `--no-graph` o `skills list`, ninguna línea |
| `grafo-codigos` | US4.3-US4.4, US6.4: códigos 0, 1, 2 y 3 de los tres verbos; no base y directorio → 1 con el fichero intacto; argumentos de más o de menos → 2; `show` con un id vacío, formado solo por espacio en blanco (un id de U+0020, uno de U+00A0 y uno de U+2003) o con un carácter de control (U+007F; U+0085 y el tabulador, que son de las dos clases) → 2, y `show` de `a b`, de ` ine:28074` y de `ine:28074` + U+200B → 3 (applet-graph.md §4; research D35). U+0000 no puede viajar en un argumento (research V40): lo fija `TestValidarID`. Los caracteres que no son ASCII van literales dentro de las comillas de `exec sh -c '…'` (`testscript` solo parte por espacio ASCII, tabulador, `\r` y `#`, V40); variable vacía, fichero o sin `HOME` → 2 con y sin `--no-graph`; `world.db` de 0 bytes → `stats` a cero, `check` `[]`, `show` 3 y el fichero igual; después `territorio resolver` crea en él el esquema sin aviso |
| `grafo-show` | US4.1-US4.2: `show` del `Bloque` con sus dos aristas y del `BloqueVersion` con `fecha_vigencia` y `hash_texto`; sin `--json`, las cuatro líneas de procedencia de la tabla mínima y `nodo.id`, `nodo.tipo`; ninguna línea del texto del artículo en `show`, `stats` ni `check` (este con T8, con hallazgos), con `--json` ni sin él, con un control positivo sobre la tabla de `boe articulo` (`sin-texto.sh`) |
| `grafo-concurrencia` | US6.3: ocho `territorio resolver` a la vez con ocho municipios distintos sobre un directorio de caché que no existe → todos 0, sin aviso de grafo; en el directorio, `world.db` y ningún `world.db-nuevo-*`; `graph stats` cuenta ocho `Municipio`, ocho `Organo` y ocho aristas, y `show` de cada municipio da su órgano |
| `grafo-matriz-territorial` | FR-043, FR-045: Leganés (cubierto) y Tordesillas (no cubierto) dejan cada uno su `Municipio`, su `Organo` y su arista con la misma forma |
| `grafo-applet` | FR-050, FR-051, FR-031: la ayuda del binario y la de `graph` con sus tres verbos, las ocho banderas globales, `--describe` de cada uno (el de `show` sin su argumento → 2), `graph` sin verbo → 2, el enlace `graph`, el sobre firmado por `kitlegal.graph` (con T0 y con el reloj del sistema; `--offline` igual), y la ayuda de `--no-graph` en `boe articulo --help` y en `graph stats --help` sin la frase antigua |

Los municipios de `grafo-concurrencia` se toman de `data/territorio/municipios.yaml` (p. ej. `28074`, `28079`,
`28065`, `28007`, `28092`, `28058`, `28106`, `47165`), y sus DIR3 de `dir3.yaml`.
