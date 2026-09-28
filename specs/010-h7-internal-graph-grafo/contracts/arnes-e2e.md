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

Lo servido desde la caché conserva la `fecha_consulta` con la que se guardó (H4). Los binarios `$KITLEGAL_V1_BIN`,
`$KITLEGAL_V2_BIN` y `$KITLEGAL_SIN_ENLACES_BIN` de H19 siguen igual. `TestBinariosDelArnes` gana los tres nuevos: cada
uno responde `version` con `dev` y su `graph stats --json` sobre un directorio vacío lleva en `fecha_consulta` su
instante.

## 3. Directorio de trabajo

- `KITLEGAL_CACHE_DIR=$WORK/cache` (vacío al empezar): ahí aparecen `cache.db` y `world.db`.
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

- Bytes de `world.db`: `[unix] exec cksum cache/world.db` + `cp stdout antes.txt`, y después `cmp stdout antes.txt`;
  auxiliares: `! exists cache/world.db-wal` y `! exists cache/world.db-shm`. Entre dos órdenes de un guion no hay
  ninguna conexión abierta, así que no quedan auxiliares y los verbos de `graph` leen en el estado normal, sin cambiar
  ningún byte (research D10, V9, V36 E); ningún guion construye un `-wal` huérfano: ese estado lo ejerce la matriz de
  integración (contracts/almacen-world-db.md §7).
- Quitar una variable: `[unix] exec env -u KITLEGAL_CACHE_DIR -u HOME kitlegal …` (`env K=` la deja presente y vacía,
  research V8).
- Códigos distintos de 0: `exec sh -c 'kitlegal …; test $? -eq N'`, como los guiones de H4-H6.
- «Exactamente una línea más en la salida de error»: un guion auxiliar dentro del propio `.txtar` que compara dos
  salidas de error guardadas con `cp stderr`, como `una-linea-mas.sh` de `h19-skills-aviso`.
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
| La ayuda de `--no-graph` | [resultado-y-entrega.md](./resultado-y-entrega.md) §6 |
| Mensajes que nombran `world.db` | [almacen-world-db.md](./almacen-world-db.md) §6 |

Ningún valor se escribe de memoria: identificadores, `hash_texto`, fechas de vigencia y DIR3 salen de las grabaciones,
de las derivadas y de `data/territorio/`.

## 6. Los once guiones

| Guion | Qué ejerce |
|---|---|
| `grafo-memoria` | US1.1-US1.4: `boe articulo` (T0) y `territorio resolver`, `graph stats` con los recuentos y las fuentes, repetición sin cambios, `graph show` de cada nodo y arista con la procedencia del sobre, `--offline` que también entrega; y el empate de FR-023: `a22` y `a21` con T0 (en ese orden, sobre una caché nueva) dejan en la `Norma` la url del bloque `a21`, la menor |
| `grafo-no-emiten` | US1.5: `url_eli` vacío y sin `eli` → misma salida y código que con `--no-graph`, stderr vacío, sin `world.db`; `indice`, `buscar`, `metadatos`, `analisis`, `skills install/list/doctor`, los tres de `graph`, una invocación fallida y `--dry-run` no crean `world.db` |
| `grafo-version-obsoleta` | US2.1-US2.3: A (grabación, T0) y B (derivada, T1) → un hallazgo `version-obsoleta` sobre A con la cita, `20161002`, `20250101`, la url del bloque y `2026-09-29T12:00:00Z`; una sola versión → ninguno; orden inverso en otro directorio de caché → el mismo hallazgo |
| `grafo-fuente-caducada` | US2.4-US2.5: A en T0, `check` con T8 → `fuente-caducada` en la `Norma`, el `Bloque` y la versión con `604800`, `604800 s` y `2026-10-05T12:00:00Z`, ninguno en `Municipio` ni `Organo`; con T1 ninguno; sin `world.db` → `[]` y sigue sin existir |
| `grafo-no-interferencia` | US3.1-US3.4, US3.7: `boe articulo` con y sin `--no-graph`, con y sin `--json`, byte a byte y mismo código; `--no-graph` y `--dry-run` dejan `world.db` con los mismos bytes o sin crear; `graph stats` con y sin `--no-graph` iguales (T0) sin cambiar `world.db`; con `KITLEGAL_CACHE_DIR` vacía, `graph stats --no-graph` sale con 2 |
| `grafo-entrega-fallida` | US3.5-US3.6: `world.db` que no es base → `boe articulo` desde la caché con la misma salida y código y una línea más que nombra `world.db`, fichero intacto; `KITLEGAL_CACHE_DIR` vacía, que nombra un fichero, y sin ella ni `HOME` → `territorio resolver` sale con 0, misma salida que con `--no-graph`, una línea más, nada creado; con `--no-graph` o `skills list`, ninguna línea |
| `grafo-codigos` | US4.3-US4.4, US6.4: códigos 0, 1, 2 y 3 de los tres verbos; no base y directorio → 1 con el fichero intacto; argumentos de más o de menos → 2; `show` con un id vacío, formado solo por espacio en blanco (un id de U+0020, uno de U+00A0 y uno de U+2003) o con un carácter de control (U+007F; U+0085, que es de las dos clases) → 2, y `show` de `a b` → 3 (applet-graph.md §4; research D35). U+0000 no puede viajar en un argumento (research V40): lo fija `TestValidarID`. Los caracteres que no son ASCII van literales dentro de las comillas de `exec sh -c '…'` (`testscript` solo parte por espacio ASCII, tabulador, `\r` y `#`, V40); variable vacía, fichero o sin `HOME` → 2 con y sin `--no-graph`; `world.db` de 0 bytes → `stats` a cero, `check` `[]`, `show` 3 y el fichero igual; después `territorio resolver` crea en él el esquema sin aviso |
| `grafo-show` | US4.1-US4.2: `show` del `Bloque` con sus dos aristas y del `BloqueVersion` con `fecha_vigencia` y `hash_texto`; ninguna línea del texto del artículo en `show`, `stats` ni `check`, con `--json` ni sin él |
| `grafo-concurrencia` | US6.3: ocho `territorio resolver` a la vez con ocho municipios distintos → todos 0, sin aviso de grafo, `graph stats` cuenta ocho `Municipio` |
| `grafo-matriz-territorial` | FR-043, FR-045: Leganés (cubierto) y Tordesillas (no cubierto) dejan cada uno su `Municipio`, su `Organo` y su arista con la misma forma |
| `grafo-applet` | FR-050, FR-051, FR-031: la ayuda de `graph` con sus tres verbos, `--describe` de cada uno, `graph` sin verbo → 2, el enlace `graph`, el sobre firmado por `kitlegal.graph`, y la ayuda de `--no-graph` en `boe articulo --help` sin la frase antigua |

Los municipios de `grafo-concurrencia` se toman de `data/territorio/municipios.yaml` (p. ej. `28074`, `28079`,
`28065`, `28007`, `28092`, `28058`, `28106`, `47165`), y sus DIR3 de `dir3.yaml`.
