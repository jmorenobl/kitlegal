# Implementation Plan: H7 · `internal/graph`: grafo del mundo, operaciones en el `Resultado` y `graph check` mínimo

**Branch**: `010-h7-internal-graph-grafo` | **Date**: 2026-09-28 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/010-h7-internal-graph-grafo/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de
`.specify/memory/constitution.md` y están en [research.md](./research.md) (D1-D35) con su alternativa rechazada y su
motivo. Toda afirmación sobre una herramienta o dependencia externa remite a la tabla de verificación de
[research.md](./research.md) (V1-V48), con el `fichero:línea` del módulo o del toolchain, la salida de `go doc` o la
sonda ejecutada en local sin red; lo que no se puede comprobar sin el runner de CI o sin el job de evals está declarado
como supuesto (S1-S7) y **no se afirma como hecho** en ningún punto de este plan. Lo que el diseño no puede cumplir
del spec está medido y declarado como desviación en *Complexity Tracking* y en `gates/supuestos.md`.

## Summary

H7 entrega la pieza G0 del grafo (ADR 0014): el binario **recuerda** lo que observa, con su fuente, y **comprueba**
si algo de lo consultado ha cambiado o caducado. Cinco piezas: (1) `world.db` junto a la caché, con `nodes`, `edges` y
`texts`, y un `Apply` transaccional e idempotente que no deja entrar nada sin fuente ni un `Persona` con forma de
documento de identidad; (2) las operaciones de grafo como quinto campo de `schema.Resultado`, que el kernel entrega
**después de presentar** con la procedencia del sobre; (3) `boe articulo`/`articulos` y `territorio resolver` emiten;
(4) el applet `graph` con `show`, `stats` y `check` (`version-obsoleta`, `fuente-caducada`, salida 0 con hallazgos o
sin ellos); y (5) `boe-legislacion` v0.1, que comprueba con `graph check` antes y después de leer y lo dice, citando
siempre de `kitlegal boe articulo`.

Decisiones que sostienen el diseño:

1. **Cada pieza en su capa** (D1): tipos de operación en `internal/core/schema` (sin importaciones nuevas, V15);
   puerto `core.GraphStore` y `core.Lote` en la raíz del dominio; validación, fusión, `check` y formas de salida en un
   dominio nuevo, `internal/core/grafo`; SQLite en `internal/graph`; composición en `internal/app`; lote y entrega en el
   kernel (`internal/cli`).
2. **La entrega es del kernel y va detrás de la presentación, por construcción** (D4, D5): `Montador.Emitir` recibe el
   contexto con el plazo de `--timeout`, presenta y solo entonces entrega con la fuente, la url y el texto exacto de
   `fecha_consulta` **del sobre presentado**. Un fallo, `--dry-run` y un resultado sin operaciones no llegan a
   entregar; una entrega que falla deja una sola línea en la salida de error y el código intacto (D7).
3. **Leer sin dejar rastro** (D10, sondas V9, V36, V43 y V45-V48): los auxiliares se buscan con el nombre que les da
   SQLite (junto al destino si `world.db` es un enlace, V47) y se comprueba si el proceso puede escribir `world.db`.
   Sin auxiliares —el estado normal—, con permiso, los verbos de `graph` abren con `mode=rw` + `query_only`, que no
   crea el fichero ni deja nada (`mode=ro` dejaría `-wal` y `-shm`); sin permiso, con `mode=ro&immutable=1`, que
   tampoco (cualquier otro modo dejaría `-wal` y `-shm` que nadie puede borrar, V46). Con auxiliares (otra conexión,
   un `-wal` huérfano o un `world.db-journal`), con `mode=ro`, que no toca `world.db`, `-wal` ni el diario (`mode=rw`
   haría un checkpoint al cerrar o desharía un diario caliente). Un diario caliente no se deja leer sin deshacerlo: el
   verbo sale con 1 sin tocar nada (FR-010). Lo único inevitable, lo que SQLite escribe en los auxiliares de WAL para
   leer lo confirmado en ellos (`world.db-shm`, y un `-wal` vacío junto a un `-shm` suelto), es una desviación
   declarada.
4. **Escribir sin estropear lo que había** (D11, sondas V10, V11, V36 y V41-V46): validar el lote antes de tocar el
   disco; con `world.db` ausente, construir WAL, esquema y lote en un temporal del mismo directorio y publicarlo con
   `os.Link` —un fallo no deja nada, ni el directorio—; con un `world.db` que el proceso no puede escribir, fallar antes
   de abrir SQLite, sin tocar nada (V46); con un `world.db` sin esquema, fijar WAL fuera de toda transacción (dentro,
   SQLite no lo fija, V42) y crear el esquema y aplicar el lote en **una** transacción inmediata. Un rechazo no crea ni
   cambia nada salvo en dos casos que llegan de fuera o de una interrupción, declarados como desviaciones en bytes por
   su causa y con sus cotas: lo que SQLite escribe al confirmar el paso a WAL de un `world.db` sin esquema que llega de
   fuera, si la entrega falla después (el contenido no cambia, sigue en versión 0 y no aparece ningún fichero; los
   conjuntos de bytes medidos son ejemplos que fija un test, V41, V45), y la recuperación que SQLite hace de un `-wal`
   huérfano o de un diario caliente aunque la entrega falle después (V43, V44).
5. **FR-023 como orden total** (D13) en el dominio: el resultado no depende del orden de llegada ni de repetir.
6. **Reloj fijo en e2e por construcción, no por entorno** (D23): tres binarios de e2e con `-X main.reloj`, para que
   `fecha_consulta`, `version-obsoleta` y `fuente-caducada` se afirmen literales y no dependan del día en que corran.
7. **La eval de la consulta repetida prepara un grafo previo** con una versión anterior derivada y una caché con la
   grabada (D22, D26): el texto que la skill cita es siempre el real.

Artefactos de diseño: [data-model.md](./data-model.md) y [contracts/](./contracts/)
([resultado-y-entrega](./contracts/resultado-y-entrega.md), [almacen-world-db](./contracts/almacen-world-db.md),
[applet-graph](./contracts/applet-graph.md), [emision](./contracts/emision.md),
[arnes-e2e](./contracts/arnes-e2e.md), [evals-y-skill](./contracts/evals-y-skill.md)); validación en
[quickstart.md](./quickstart.md).

## Technical Context

**Language/Version**: Go 1.27 sin cambios (`go 1.27.0` + `toolchain go1.27.1`; V1). SQL de SQLite para la migración;
YAML y JSON Schema 2020-12 para evals y esquemas; Markdown para la skill; `testscript` para los guiones.

**Primary Dependencies**: las de `go.mod`, **ninguna nueva**. `modernc.org/sqlite` (§V, ya enlazado por la caché) en
`internal/graph`; biblioteca estándar nueva en el árbol: `encoding/json/jsontext` y `encoding/json/v2` (dominio;
precedente en `internal/core/instalacion`, V5), `regexp` (dominio, V6), `net/url` (adaptador `boe`), `embed`/`io/fs`
(migraciones). `TestDependenciasDelBinario` no cambia su lista.

**Storage**: `world.db` (SQLite, WAL) en el directorio de la caché —`KITLEGAL_CACHE_DIR` o `~/.cache/kitlegal/`— con
`nodes`, `edges`, `texts` y `schema_version` (data-model §3). Sin cambios en `cache.db`.

**Testing**: `make test` (unitarios del dominio, del adaptador sobre `t.TempDir()`, del kernel, del applet y los e2e
con `-race`), `make test-integration` (gana `internal/graph/integracion_test.go`, la matriz de FR-088, e
`integracion_enlace_test.go`, sus casos de enlace simbólico, solo en Unix),
`make test-tiempos` (gana `TestCosteDelGrafo`: SC-007 y SC-008), `make schema-check` (gana `schemas/grafo.json`),
`make skills-check` (la tabla nueva de la skill y la eval nueva). **Sin red**: ni el grafo, ni `territorio`, ni `graph`
la usan, y `boe` responde en e2e desde la reproducción de H4.

**Target Platform**: sin cambios (`CGO_ENABLED=0`, `-trimpath`; darwin, linux y windows en amd64 y arm64). `make ci` en
`ubuntu-latest`; job de evals en `ubuntu-24.04`.

**Project Type**: CLI multicall + skills del estándar Agent Skills.

**Performance Goals**: entregar lo observado por un `boe articulo` servido de la caché añade como mucho 150 ms a la
mediana de 20 invocaciones (SC-007); `graph check` < 3 s y `graph stats` < 1 s sobre 10 000 nodos y 10 000 aristas
(SC-008). Las mide `TestCosteDelGrafo` solo, en `test-tiempos` (D28); su cumplimiento en el runner es el supuesto S1-S2.

**Constraints**: la salida de ningún applet cambia ni un byte (FR-042, SC-003); `--no-graph` y `--dry-run` no llegan
a entregar y dejan `world.db` y sus auxiliares con sus bytes (SC-004); ningún verbo de `graph` modifica `world.db`,
`world.db-wal`, `world.db-journal` ni el contenido del grafo (FR-004, FR-005) ni devuelve texto legal (FR-070), y sin
auxiliares no cambia ni un byte de nada, con `--no-graph` o sin ella, pueda el proceso escribir `world.db` o no y sea
`world.db` un fichero o un enlace (D10, V46, V47); **con auxiliares de WAL** (otra conexión abierta, un `-wal`
huérfano o un `-shm` suelto), SQLite reescribe o crea `world.db-shm` al leer, y junto a un `-shm` suelto crea un
`world.db-wal` vacío: desviación declarada de FR-004, FR-031 y SC-004 (D10, V48); con un diario caliente el verbo
sale con 1 sin leer ni cambiar nada (D10, V43). Una entrega que falla no crea ni cambia nada (FR-033) —tampoco sobre
un `world.db` que el proceso no puede escribir, que falla antes de abrir SQLite (D11, V46)—, **salvo** en dos casos,
desviaciones declaradas en bytes y no en el contenido del grafo, por su causa y con sus cotas (D11): (a) un
`world.db` existente sin esquema y fuera de WAL —solo llega de fuera, de cualquier programa— que falla después de
fijar WAL: queda lo que SQLite escribe al confirmar el paso a WAL —bytes de la cabecera; con `auto_vacuum=full` y
páginas libres, el vaciado y el truncado del fichero; un `world.db-journal` frío o vacío que desaparece; 0 bytes →
4096—, con el contenido de esa base igual, la versión 0 y ningún fichero nuevo (V41, V45); (b) un `world.db` con un
`-wal` huérfano o un diario caliente, que la conexión de la entrega recupera aunque la
entrega falle después (checkpoint al cerrar o diario deshecho, V43, V44). Una señal que termina el proceso a mitad de
una entrega no es un fallo que el programa trate (V39) y puede dejar el temporal `world.db-nuevo-*` (D11). El contrato `Applet` no cambia (FR-020); `internal/graph` no importa
`source/*` ni `render` (FR-092); nada escribe en el `world.db` de la cuenta desde un test (D6); las cotas
`cronometra 200ms` existentes se cumplen con la entrega (S6).

**Scale/Scope**: 1 applet nuevo con 3 verbos · 2 paquetes nuevos (`internal/core/grafo`, `internal/graph`) · 1
fichero de puerto (`internal/core/graphstore.go`) · emisión en 2 applets · 1 esquema publicado nuevo y 1 modificado
(`schemas/eval.yaml.json`) · 4 grabaciones derivadas · 1 eval nueva · 1 `SKILL.md` · 11 guiones de aceptación y 2
guiones e2e existentes tocados · arnés con 3 binarios más · documentación.

## Constitution Check

*GATE: debe pasar antes de la fase 0 y volver a evaluarse tras la fase 1.*

### Principios

| # | Principio | Cómo lo cumple H7 | Veredicto |
|---|---|---|---|
| **I** | Fuentes públicas y frontera humana | H7 **no toca ninguna fuente nueva ni añade ninguna petición**: el grafo es local, `graph` solo lee `world.db` y la emisión sale de lo que `boe` y `territorio` ya obtienen; ninguna petición HTTP nueva y nada con identidad. `boe` sigue pidiendo solo por `internal/httpx` (R2) | ✅ Cumple |
| **II** | Nada sin cita ni fuente | **Un dato sin fuente no entra, por construcción**: el applet no puede escribir la fuente —las operaciones no la tienen (D2)— y el kernel la toma del sobre presentado (D4, D5); `Apply` rechaza además un lote sin fuente, url absoluta o fecha (FR-024). **El grafo no es fuente de contenido citado** (ADR 0014): ningún verbo de `graph` devuelve el cuerpo de un bloque (FR-070, SC-006) y `SKILL.md` prohíbe citar de `graph` (FR-081); lo vigilan `TestNingunVerboDelGrafoDevuelveTexto` y la eval con `graph show` prohibido. Los sobres de `graph` firman en el espacio reservado (`kitlegal.graph`) | ✅ Cumple |
| **III** | Tests primero y offline | T001 escribe los 11 guiones de aceptación desde el spec y **quedan congelados** antes de cualquier código (ADR 0018); la eval de la skill va antes que `SKILL.md` (DoD §1.10) y detrás del formato que la admite (D25). Cada tarea trae su test. Todo test es offline: dominio con datos sintéticos, almacén sobre `t.TempDir()`, e2e con la reproducción de H4 y derivadas versionadas. Toda salida de `graph` contra `schemas/grafo.json` en test (FR-051). Umbrales intactos: `internal/core/**` ≥ 85 % —el dominio nuevo cae ahí— y global ≥ 70 % | ✅ Cumple |
| **IV** | Arquitectura hexagonal con reglas ejecutables | Dominio puro en `internal/core/{schema,grafo}` y el puerto en `internal/core`; adaptador SQLite en `internal/graph`; composición en `internal/app`; entrega en el kernel. Errores tipados: `*graph.Error` implementa `schema.ConClase` y el kernel los traduce a 1, 2 y 4; `show` de un id ausente es `no-encontrado` (3); los hallazgos son `data` con salida 0 (ADR 0023). Ningún `panic` en rutas de usuario. Reglas de dependencia, abajo | ✅ Cumple |
| **V** | Simplicidad y dependencias fijadas | **Ninguna dependencia nueva.** Sin DI, ORM ni generador de CLI; `internal/cli` no se extrae. Lo que el spec no pide no se construye: ni `eli:cites`, ni FTS5, ni `graph query/neighbors/path/history/export`, ni grafo del asunto (spec, *Fuera de alcance*), ni opciones sin consumidor (la API de `internal/graph` tiene una sola, `ConDirectorio`, que usan los tests, el applet `graph` y la preparación de evals; no hay opción de registrador), ni atención a señales ni limpieza de temporales ajenos (D11). Las piezas nuevas que el spec no enumera están en *Complexity Tracking* | ✅ Cumple con justificación |
| **VI** | Un binario, convenciones de agente | `graph` es un applet más del mismo ejecutable, invocable como `kitlegal graph …` y por el enlace `graph`, con las ocho banderas globales y `--describe`; `--no-graph` deja de ser una bandera sin efecto y pasa a gobernar la entrega (FR-031), con su ayuda nueva | ✅ Cumple |
| **VII** | Grafo y privacidad | Ids naturales (ELI, `ine:<código>`, DIR3) en `world.db`; **rechazo por expresión regular de DNI, NIE y NIF en `Persona` dentro de `Apply`** (FR-025), con clases ASCII explícitas y sin `(?i)`, cuyo plegado Unicode dejaba pasar `ſ12345678Z` (V6, D34): la lectura de «letra», «cifra» y «espacio» que más rechaza en los límites, comprobada contra los 20 rechazos y las 11 aceptaciones del contrato (V37) y probada contra `Apply` sin ningún emisor de `Persona`. Sin grafo del asunto (H10), sin exportación ni sincronización (ADR 0027): `world.db` no sale de la máquina | ✅ Cumple |
| **VIII** | Skills primero; el binario es la herramienta | El hito entrega `boe-legislacion` v0.1, medida con una eval nueva, y el applet `graph` existe para ella (`graph check` en su protocolo; `graph show` para ver versiones). A Go va solo lo determinista: identidades, fusión de observaciones, reglas de `check`, huellas; el razonamiento sobre qué decir de un hallazgo vive en `SKILL.md`. `SKILL.md` < 300 líneas, tabla generada, sin `scripts/` | ✅ Cumple |
| **IX** | Genericidad territorial, validación local | `territorio` emite igual para cualquier municipio, cubierto o no (FR-045): ningún caso especial. Matriz territorial en e2e con Leganés (cubierto) y Tordesillas (no cubierto) (`h7-grafo-matriz-territorial`). Los municipios concretos solo aparecen en guiones, tests y evals | ✅ Cumple |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución §IV)

| Regla | Situación en H7 | Cómo se hace cumplir | Veredicto |
|---|---|---|---|
| `internal/core/**` no importa `internal/{source,httpx,cache,store,graph,render,cli,app,disco}` ni entrada y salida | `internal/core/grafo` y `internal/core/graphstore.go` importan solo `internal/core`, `internal/core/schema` y biblioteca estándar sin E/S (`context`, `time`, `regexp`, `encoding/json/v2`, `jsontext`, `crypto/sha256`…); `internal/core/territorio` gana `grafo.go`, que importa `internal/core/grafo`; `schema` no gana importaciones (V15) | `depguard` `core` (ya deniega `internal/graph`, V29) + `TestArquitectura` R1, que recorre todo paquete nuevo bajo `internal/core` + `TestContexto` | ✅ Cumple |
| Solo `internal/httpx` importa `net/http` | Ningún paquete nuevo lo importa; `boe` usa `net/url` para leer `url_eli` | `depguard` `red` + R2 | ✅ Cumple |
| Solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | `internal/graph` es el segundo dueño real; ningún otro paquete nuevo los importa | `depguard` `sql` (ya exceptúa `internal/graph/**`) + R3 (sigue sin exigir dueño porque `internal/store` no existe hasta H10; se corrige su comentario) | ✅ Cumple |
| Solo `internal/cli` y `cmd/` llaman a `os.Exit` | Ningún `package main` nuevo; el de e2e gana solo composición | `forbidigo` `^os\.Exit$`, sin excepciones nuevas | ✅ Cumple |
| Solo `internal/render` escribe en stdout; logs con `slog` a stderr | `graph` devuelve un `Resultado`; la línea de aviso de la entrega sale por `Presentador.Aviso`; el almacén no escribe en ningún descriptor | `forbidigo` `^fmt\.Print…$`, `^os\.Std(out|err)$`, sin excepciones nuevas | ✅ Cumple |
| `internal/graph` no importa `internal/source/*` ni `internal/render` | Importa `internal/core`, `internal/core/grafo`, `internal/core/schema` e `internal/cache` (solo `cache.Directorio`, D8) | **Nuevo**: lista `grafo` de `depguard` + subprueba R6 de `TestArquitectura`, transitiva y con exigencia de que `internal/graph` exista (D30) | ✅ Cumple |
| Los applets de ejemplo no se enlazan en el binario distribuido (ADR 0010) | El binario de e2e registra `graph` como el distribuido | `depguard` `ejemplo` + `TestElBinarioNoEnlazaLosEjemplos` | ✅ Cumple |
| Ningún adaptador firma en el espacio reservado (ADR 0006) | `graph` firma en él y puede: es un applet calculado en `internal/app`; `internal/graph` no firma nada | `TestLasFuentesNoFirmanComoKitlegal` | ✅ Cumple |
| El binario no enlaza módulos no justificados | Ninguno nuevo: SQLite ya llegaba por la caché | `TestDependenciasDelBinario`, sin cambios en su lista | ✅ Cumple |
| La superficie de `internal/cache` no deja asomar la base (FR-005 de H3) | Gana `func Directorio() (string, error)`, que no nombra `database/sql` ni SQLite | `TestSuperficieExportada` con la entrada nueva en su lista cerrada | ✅ Cumple con justificación |

### Gates (constitución, «Gates»)

- **Capa 1 (mecánica)**, lo que H7 añade o toca (detalle en «Controles mecánicos»): la suite de aceptación congelada
  con su rojo-primero; los tests del dominio del grafo (validación, `Persona`, fusión en los dos órdenes, `check`,
  explicaciones); la matriz de integración de `internal/graph` (FR-088); la salida de `graph` contra
  `schemas/grafo.json` y `make schema-check`; el esquema de eval ampliado y sus reglas; la tabla de comandos regenerada
  sin deriva; la regla R6 en `depguard` y `TestArquitectura`; la matriz territorial en e2e; `TestCosteDelGrafo` en
  `test-tiempos`. Guardián de diff con `[datos]` para `testdata/` y `schemas/`.
- **Capa 2 (jueces)**: que el protocolo de `SKILL.md` compruebe antes y después, agrupe por clase y no cite de `graph`
  (FR-080 a FR-082), y que las explicaciones de `graph check` sean claras (constitución, capa 2), lo juzgan los dos
  jueces de la revisión final; la eval informativa lo mide en el job (SC-012) en el cierre del workflow.
- **Capa 3 (humano)**, siempre fuera del run: el esquema nuevo `schemas/grafo.json` y el modificado
  `schemas/eval.yaml.json`; las cuatro grabaciones derivadas; los guiones e2e **existentes** que cambian
  (`argumentos.txtar`, `territorio-matriz.txtar`) y los activados; los supuestos del run; la fusión. **Ninguna pausa a
  mitad del run** (ADR 0018).

**Reglas del modo desatendido**: el ejecutor arregla el código, nunca un guion congelado ni un fixture; no usa la red,
`KITLEGAL_RECORD`, `make evals` ni `make verify-sources`; ningún identificador `BOE-A-…`, `hash_texto`, fecha de
vigencia, código INE ni DIR3 se escribe de memoria: sale de una grabación, de una derivada o de `data/territorio/`.
Las tareas `[datos]` que tocan código además de `testdata/` o `schemas/` son dos, razonadas en *Complexity Tracking*.

**Veredicto del gate: PASA.** Las piezas marcadas «con justificación» están en *Complexity Tracking*; ninguna afecta a
alcance, frontera humana, privacidad, términos de uso ni a una decisión cerrada. Las tres desviaciones del spec que
el diseño no puede evitar (abajo) son técnicas y tampoco tocan ninguna de esas materias. Se declaran **por su causa y
con sus cotas** —lo que nunca cambia: el contenido del grafo (o de la base de fuera), su versión y, al cerrar, que no
aparece ningún fichero nuevo, salvo el `world.db-shm` y el `-wal` vacío de la lectura con auxiliares de WAL—, no por
una lista de bytes que se dé por exhaustiva para cualquier fichero de fuera: los conjuntos medidos (V36, V41, V43-V45,
V48) son ejemplos, y un test fija cada uno. Los estados de `world.db` que llegan de fuera y que el diseño sí puede
evitar no se declaran: se evitan (un `world.db` que el proceso no puede escribir, V46; un enlace simbólico, V47).

### Re-evaluación tras la fase 1 (diseño)

- **`internal/core/grafo`** (§IV, estructura): dominio en su sitio, no listado en `CLAUDE.md` (como
  `internal/core/instalacion` en H19); lo que el roadmap llama «tipos de operación en `internal/core`» son los de
  `schema` y el puerto de `graphstore.go`.
- **`internal/graph` importa `internal/cache`** (§IV): dependencia entre adaptadores que ninguna regla prohíbe, por la
  única función que hace que `world.db` viva «junto a la caché y con su misma regla» (FR-001); va en *Complexity
  Tracking*.
- **`Montador.Emitir` con contexto y almacén** (kernel): cambia la API del kernel, no el contrato `Applet`; va en
  *Complexity Tracking*.
- **La línea de aviso no propaga el error de su escritura** (constitución, «sin atajos»): lo exige FR-033, está
  razonado (D7) y lo fija un test.
- **`.golangci.yml`**: gana la lista `grafo` y una entrada de `misspell` (`observacion`, D31); ninguna regla se relaja.
- **`internal/app/testdata/script/territorio-matriz.txtar`** cambia una aserción que H7 vuelve falsa por diseño
  (D29); sus tres `cronometra 200ms` y los diez de `boe-cache-rapida.txtar` no cambian y la entrega tiene que caber en
  ellos (S6).
- **Privacidad (VII), tras la corrección del plan**: la expresión de FR-025 va sin `(?i)` y con clases ASCII
  explícitas (D34); con la anterior, `ſ12345678Z` entraba (V6, V37). La fila VII es veraz con la expresión nueva.
- **Estados de fuera que el diseño evita, tras la corrección del plan**: (a) un `world.db` que el proceso no puede
  escribir (`chmod a-w` basta): SQLite lo abre en solo lectura sin avisar y, en WAL, cualquier modo que no sea
  `immutable` deja `world.db-shm` y `world.db-wal` (V46); los verbos de `graph` comprueban el permiso con
  `os.OpenFile(…, os.O_RDWR, 0)` y, sin él y sin `-wal` ni `-journal`, leen con `mode=ro&immutable=1`, y la entrega
  falla antes de abrir SQLite (D10, D11); (b) un `world.db` que es un enlace simbólico: fuera de Windows SQLite crea
  los auxiliares junto al destino, así que se buscan allí (D10, V47), y uno sin destino no se crea a través del
  enlace (D11).
- **Desviaciones del spec que quedan, declaradas** (*Complexity Tracking*, `gates/supuestos.md`), las tres en bytes y
  ninguna en el contenido del grafo, cada una por su causa y con sus cotas:
  1. **FR-033, paso a WAL de una base de fuera** (D11): una entrega que falla después de fijar WAL en un `world.db`
     que ya existía sin esquema y fuera de WAL —0 bytes o base en rollback, que el binario nunca crea: llega de
     cualquier programa, con cualquier versión y configuración de SQLite— deja lo que SQLite escribe al confirmar el
     paso a WAL. **Cotas**: el contenido de esa base no cambia, sigue en versión 0 (grafo vacío), y cerrada la última
     conexión no queda ningún fichero nuevo. **Qué cambia**: bytes de la cabecera de la página 1 (siempre 18, 19,
     24-27 y 92-95; los campos que la confirmación recalcula —28-31, 32-39, 96-99— si no coincidían); con
     `auto_vacuum=full` y páginas libres, el vaciado de todas ellas, con la reubicación de las páginas en uso que están
     detrás y el truncado del fichero; un `world.db-journal` frío o vacío desaparece; 0 bytes → 4096. Los conjuntos
     exactos medidos —cabecera de este controlador, de otra versión, con el tamaño a 0, con acarreo, con
     `auto_vacuum=full` y páginas libres al final o delante de una página en uso, con un diario frío de `PERSIST` o
     vacío de `TRUNCATE` (V36 B, V41, V45)— son ejemplos que `aplicar_test.go` fija caso a caso, no una lista cerrada
     para cualquier base. Con `world.db` ausente, en cambio, un fallo no deja nada: se construye en un temporal y se
     publica con `os.Link` (V36 C), y la primera versión de este plan, que dejaba un `world.db` de 4096 bytes y el
     directorio, queda descartada.
  2. **FR-033, recuperación de SQLite** (D11): con un `-wal` huérfano o un diario de rollback caliente, que deja un
     escritor interrumpido, la conexión de la entrega los recupera aunque la entrega falle después: el diario se
     deshace al leer (V43) y el `-wal` se lleva a `world.db` en el checkpoint del cierre si es la última conexión
     (V44). Cambian los bytes y el tamaño de `world.db` y desaparecen los auxiliares; el contenido es el confirmado.
  3. **FR-004, FR-031 y SC-004** (D10): con auxiliares de WAL (otra conexión abierta, un `-wal` huérfano o un `-shm`
     suelto), leer escribe lo que SQLite necesita en ellos para leer lo confirmado: reescribe o crea `world.db-shm`, y
     junto a un `-shm` suelto crea un `world.db-wal` de 0 bytes (V36 D, F, V48). **Cota**: `world.db`, un `-wal` que ya
     existía, el diario y el contenido del grafo no cambian. Sin auxiliares —el estado de los guiones, del quickstart y
     de toda invocación que ni coincide con otra ni sigue a un escritor interrumpido— no cambia nada, con permiso de
     escritura o sin él y con `world.db` fichero o enlace (V9, V36 E, V46, V47); con un diario de rollback, tampoco:
     se lee sin cambiar nada o, si está caliente, el verbo sale con 1 (V43, V45 c).

  Fuera de FR-033, y por eso declarado pero no desviación: una señal que termina el proceso a mitad de una entrega
  puede dejar el temporal `world.db-nuevo-*` y los directorios creados (V39, D11); atender señales no lo pide el hito
  (lectura conservadora).

**Veredicto tras el diseño: PASA**, con esas tres desviaciones justificadas en *Complexity Tracking*, cada una
declarada por su causa y con sus cotas —que valen para cualquier `world.db`, también uno de fuera— y con los estados
medidos fijados por test como ejemplos exactos; y sin ninguna violación de la constitución. Lo que se afirma de un
fichero de fuera son esas cotas, no una enumeración de bytes: una enumeración no se puede comprobar para cualquier
base que llegue, y por eso no se afirma.

## Project Structure

### Documentation (this feature)

```text
specs/010-h7-internal-graph-grafo/
├── spec.md
├── plan.md                 # este fichero
├── research.md             # V1-V48, D1-D35, S1-S7
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── resultado-y-entrega.md
│   ├── almacen-world-db.md
│   ├── applet-graph.md
│   ├── emision.md
│   ├── arnes-e2e.md
│   └── evals-y-skill.md
├── aceptacion/             # T001: 11 guiones .txtar, congelados (ADR 0018)
├── checklists/
├── gates/
└── tasks.md                # /speckit-tasks
```

### Source Code (repository root)

```text
internal/
├── core/
│   ├── doc.go                       CAMBIA  el comentario: GraphStore ya existe
│   ├── graphstore.go                NUEVO   GraphStore, Lote
│   ├── schema/
│   │   ├── sobre.go                 CAMBIA  Resultado.Grafo
│   │   ├── grafo.go                 NUEVO   Observado, Operacion (sellada), Nodo, Arista, Texto
│   │   ├── grafo_test.go            NUEVO   sellado, valor cero
│   │   └── sobre_test.go            CAMBIA  cinco campos en TestSobre (l. 195-196)
│   ├── grafo/                       NUEVO   dominio del grafo
│   │   ├── doc.go
│   │   ├── vocabulario.go           tipos, relaciones, claves de datos, clases de hallazgo
│   │   ├── lote.go                  ValidarLote, ValidarContraGrafoVacio, Consolidar
│   │   ├── persona.go               la expresión de FR-025 y el recorrido de los datos
│   │   ├── canonico.go              datos canónicos (RFC 8785, jsontext)
│   │   ├── observacion.go           instantes, desempate, fusión de nodos, aristas y textos (FR-023)
│   │   ├── salida.go                Ficha, NodoDeFicha, AristaDeFicha, Procedencia, Recuento, Hallazgo, Instantanea
│   │   ├── comprobar.go             Comprobar: version-obsoleta, fuente-caducada, orden
│   │   ├── explicacion.go           plantillas, cita de un nodo, instante de caducidad
│   │   ├── id.go                    ValidarID (show)
│   │   ├── errores.go               errores de rechazo y de argumentos (ConClase)
│   │   └── *_test.go                sintéticos, uno por fichero
│   └── territorio/
│       ├── grafo.go                 NUEVO   Territorio.Observado()
│       └── grafo_test.go            NUEVO   incluido «sin DIR3» con fuentes sintéticas
├── cache/
│   ├── ruta.go                      CAMBIA  func Directorio()
│   ├── ruta_test.go                 CAMBIA  TestDirectorio
│   └── superficie_test.go           CAMBIA  «func Directorio» en superficieDelContrato (l. 68-89)
├── graph/                           NUEVO   adaptador SQLite de world.db
│   ├── doc.go
│   ├── almacen.go                   Nuevo, opciones, Apply
│   ├── nulo.go                      Nulo
│   ├── ruta.go                      directorio (opción o cache.Directorio), world.db
│   ├── abrir.go                     cadenas de conexión, nombre de los auxiliares (ruta resuelta fuera de Windows), comprobación de escritura, apertura para leer (modo según los auxiliares y el permiso) y para escribir, recurso inmutable
│   ├── migraciones.go               versión, migración dentro de la transacción
│   ├── migraciones/0001_grafo.sql   esquema v1 (data-model §3)
│   ├── espera.go                    reintento por tramos que mira el contexto
│   ├── aplicar.go                   la transacción de Apply
│   ├── publicar.go                  world.db ausente: directorios, temporal, os.Link (costura no exportada), limpieza
│   ├── lectura.go                   Leer, Ficha, Recuento, Instantanea, Close
│   ├── errores.go                   Error (ConClase), mensajes con world.db
│   ├── *_test.go                    t.TempDir(); superficie_test.go
│   ├── integracion_test.go          //go:build integration — matriz de FR-088
│   └── integracion_enlace_test.go   //go:build integration && unix — world.db que es un enlace simbólico
├── source/boe/
│   ├── grafo.go                     NUEVO   Observado de articulo/articulos, id ELI
│   ├── grafo_test.go                NUEVO
│   └── articulo.go                  CAMBIA  el resultado de éxito lleva su Observado
├── cli/
│   ├── sobre.go                     CAMBIA  Montador.Grafo, Emitir(ctx, …), entrega tras presentar
│   ├── entrega.go                   NUEVO   LoteDe, la línea de aviso
│   ├── entrega_test.go              NUEVO   TestEntregaDelMontador
│   ├── sobre_test.go, describe_test.go  CAMBIAN  Emitir(context.Background(), …)
│   ├── globales.go                  CAMBIA  ayuda y comentario de --no-graph
│   └── globales_test.go             CAMBIA  la ayuda literal
├── app/
│   ├── grafo.go                     NUEVO   applet graph, DependenciasDeGrafo
│   ├── grafo_test.go                NUEVO   TestAppletGrafo, TestSalidaDelGrafoContraSchemas, TestCodigosDelGrafo,
│   │                                        TestLaEntregaLlevaLaProcedenciaDelSobre, TestNingunVerboDelGrafoDevuelveTexto,
│   │                                        TestLaSalidaDeBoeNoCambiaConElGrafo, TestGrabacionesDerivadas
│   ├── coste_test.go                NUEVO   TestCosteDelGrafo (SC-007, SC-008)
│   ├── main.go                      CAMBIA  desenlace con plazo y --no-graph; montador con almacén
│   ├── main_test.go                 CAMBIA  TestEntregaDelKernel
│   ├── registro.go                  CAMBIA  EntregarAlGrafo; registra graph y graph.Nuevo()
│   ├── registro_test.go             CAMBIA  lista boe, graph, skills, territorio (l. 245-246)
│   ├── territorio.go                CAMBIA  el resultado lleva resuelto.Observado()
│   ├── esquemas_test.go             CAMBIA  fila grafo.json
│   ├── e2e_test.go                  CAMBIA  3 binarios con reloj, copia de derivadas, TestBinariosDelArnes
│   ├── ejemplo/kitlegal-e2e/main.go CAMBIA  registra graph y la entrega; variable reloj (-X)
│   ├── ejemplo/kitlegal-e2e/main_test.go CAMBIA  lista de applets (l. 90), reloj válido e inválido
│   └── testdata/
│       ├── derivadas/               NUEVO   [datos] version-posterior/, sin-eli/, eli-sin-segmento/
│       └── script/
│           ├── argumentos.txtar     CAMBIA  [datos] lista con graph (l. 24, 30)
│           ├── territorio-matriz.txtar CAMBIA [datos] ! exists cache/cache.db (l. 206)
│           └── h7-*.txtar           NUEVOS  activación del workflow (copias de aceptacion/)
├── evals/
│   ├── formato.go                   CAMBIA  Prohibidos, GrafoPrevio, formaComprobacion
│   ├── juzgar.go                    CAMBIA  forma comprobación, prohibidos
│   ├── consultas.go                 CAMBIA  la comprobación no genera consulta
│   ├── preparar.go                  CAMBIA  grafo previo antes de la caché
│   ├── grabaciones.go               CAMBIA  directorio de los grafos previos
│   ├── informe.go                   CAMBIA  comandos_prohibidos_ejecutados
│   └── *_test.go                    CAMBIAN formato, juzgar, consultas, preparar, informe, conjunto
└── arch_test.go                     CAMBIA  R6; comentario de R3
cmd/kitlegal/main_test.go            CAMBIA  appletsDelBinario con graph (l. 19); graph sin verbo
schemas/
├── grafo.json                       NUEVO   [datos] desde --describe
└── eval.yaml.json                   CAMBIA  [datos] comando-comprobacion, prohibidos, grafo_previo
testdata/evals/grafo-previo/lpac-a21-version-anterior/  NUEVO [datos] la derivada anterior
evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml  NUEVO
skills/boe-legislacion/SKILL.md      CAMBIA  protocolo v0.1 y tabla regenerada
.golangci.yml                        CAMBIA  lista grafo de depguard; misspell observacion
Makefile                             CAMBIA  MEDIDAS_DE_TIEMPO con TestCosteDelGrafo
CHANGELOG.md, README.md, CONTRIBUTING.md  CAMBIAN
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`: dominio en `internal/core` (el puerto en su raíz, la
lógica en `internal/core/grafo`), adaptador SQLite en `internal/graph`, composición en `internal/app`, kernel en
`internal/cli`, contratos en `schemas/`, fixtures en `testdata/`. No se crean `internal/store`, `pkg/`, `mcp/` ni
`plugin/`.

## Aceptación e2e

**Aceptación e2e:** 11 guiones testscript, escritos por T001 desde el spec contra
[contracts/arnes-e2e.md](./contracts/arnes-e2e.md) —con los formatos de los contratos que enumera su §5— en
`specs/010-h7-internal-graph-grafo/aceptacion/`, congelados, cada uno encabezado por la precondición del applet
`graph` y un comentario con sus FR/SC, y activados al final como `internal/app/testdata/script/h7-*.txtar` (los ejecuta
`TestEntregaDelHito` en `make ci`). Por historia de usuario:

| Guion | Historia | FR y SC que cubre |
|---|---|---|
| `grafo-memoria` | US1 | US1.1-US1.4, empate de fecha de consulta (*Edge Cases*) · FR-001, FR-002, FR-021, FR-022, FR-023, FR-026, FR-030, FR-035, FR-040, FR-043, FR-053, FR-054, FR-065, FR-089, FR-093 · SC-001, SC-002 |
| `grafo-no-emiten` | US1 | US1.5 · FR-030, FR-032, FR-034, FR-040, FR-046 · SC-004 |
| `grafo-version-obsoleta` | US2 | US2.1-US2.3 · FR-060, FR-061, FR-062, FR-063, FR-064, FR-090, FR-093 · SC-005 |
| `grafo-fuente-caducada` | US2 | US2.4-US2.5 · FR-004, FR-061, FR-065, FR-066, FR-067 |
| `grafo-no-interferencia` | US3 | US3.1-US3.4, US3.7 · FR-005, FR-031, FR-034, FR-042, FR-091 · SC-003, SC-004 |
| `grafo-entrega-fallida` | US3 | US3.5-US3.6 · FR-011, FR-012, FR-033 · SC-011 |
| `grafo-codigos` | US4, US6 | US4.3, US4.4, US6.4 · FR-004, FR-010, FR-011, FR-013, FR-031, FR-052, FR-054, FR-060, FR-094 · SC-011 |
| `grafo-show` | US4 | US4.1-US4.2 · FR-041, FR-053, FR-055, FR-070 · SC-006 |
| `grafo-concurrencia` | US6 | US6.3 · FR-014, FR-022 · SC-009 |
| `grafo-matriz-territorial` | US1 | FR-043, FR-045, FR-094 (matriz territorial) |
| `grafo-applet` | US4 | FR-031 (ayuda), FR-050, FR-051 |

Lo que ningún guion del binario puede ejercer tiene su propia aceptación, también automática: **US5** (la skill), con
la eval informativa `evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml` (FR-085 a FR-087, SC-012),
que **no** escribe T001 porque exige el formato ampliado y `make ci` valida cada eval contra el esquema (D25): entra
detrás del formato y antes de `SKILL.md`, y la mide el job en el cierre del workflow; y los jueces de la revisión para
FR-080 a FR-082. **US6.1-US6.2** (rechazos y `Persona`), sin ningún emisor que los produzca, con
`internal/graph/integracion_test.go` y `TestValidarLote`/`TestPersonaSinDocumento` (FR-024, FR-025, SC-010). **FR-044**
(sin DIR3), sin municipio real que lo ejerza (V28), con `TestObservadoDeTerritorio`. **SC-007-SC-008** con
`TestCosteDelGrafo`.

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

| Objetivo | Cambio | En `ci` |
|---|---|---|
| `test`, `test-integration` | sin cambio de receta; saltan también `TestCosteDelGrafo` porque entra en `MEDIDAS_DE_TIEMPO` | sí |
| `test-tiempos` | `MEDIDAS_DE_TIEMPO := ^(TestMedidasDeTiempo|TestCosteDelGrafo)$$` (D28) | sí |
| `schema-check` | sin cambio de receta; gana `schemas/grafo.json` por el registro de producción | sí |
| `skills-check` | sin cambio de receta; gana la tabla de `graph` en `boe-legislacion` y la eval nueva con su subprueba `grafo-previo` | sí |
| `ci` | los mismos prerrequisitos | — |

Como la receta de `test-tiempos` cambia, las tablas de objetivos de `README.md` y `CONTRIBUTING.md` se alinean en la
misma rama.

### Tests

| Test | Dónde | Qué fija |
|---|---|---|
| `TestSobre` (←), `TestContexto` | `internal/core/schema` | `Resultado` con cinco campos; `schema` sin importaciones nuevas (FR-020) |
| `TestOperacionSellada`, `TestObservadoCero` | `internal/core/schema/grafo_test.go` | solo `Nodo`, `Arista` y `Texto` son operaciones; el valor cero no emite |
| `TestValidarLote` | `internal/core/grafo/lote_test.go` | cada motivo de FR-024 que no necesita la base, con el lote entero rechazado; extremos contra grafo vacío; tipos en el lote |
| `TestPersonaSinDocumento` | `internal/core/grafo/persona_test.go` | los 20 rechazos y las 11 aceptaciones de contracts/almacen-world-db.md §5 (SC-010 pide ≥ 12 y ≥ 6) en id, valor, anidado y clave, entre ellos los límites y las formas que solo la lectura ASCII decide: `ſ12345678Z`, `Martí12345678Z` y `12345678Zá` rechazan; `12345678Ñ`, `12345678` + U+212A, `12` + U+00A0 + `345 678 Z` y las cifras de anchura completa entran (D34) |
| `TestDatosCanonicos` | `internal/core/grafo/canonico_test.go` | JCS: orden de claves, números, escapes; mismo resultado con otro orden |
| `TestFusionarNodo`, `TestFusionarArista`, `TestFusionarTexto`, `TestConsolidar` | `internal/core/grafo/observacion_test.go`, `lote_test.go` | FR-023 en los dos órdenes: fuera de orden, cada criterio del desempate, datos de la última, primera que no avanza, texto más antiguo, idéntica sin cambios |
| `TestComprobar`, `TestExplicaciones`, `TestValidarID` | `internal/core/grafo` | FR-060 a FR-067: una vez por versión superada, empates de vigencia, vigencias no válidas (también con cifras que no son ASCII, D34), versión más reciente con sus desempates, caducidad estricta, futuro, sin vigencia, orden por clase e id, plantillas literales, instante con desplazamiento, cita de cada tipo; FR-052: id vacío, solo de espacio en blanco Unicode (U+0020, U+00A0, U+2003, `\t`, U+0085) o con un carácter de control Cc (U+0000, U+007F) → `argumentos`; `a b`, ` a`, U+200B y bytes que no son UTF-8 → válidos (D35) |
| `TestObservadoDeTerritorio` | `internal/core/territorio/grafo_test.go` | FR-043 a FR-045 con fuentes sintéticas, «sin DIR3» incluido |
| `TestDirectorio` (←), `TestSuperficieExportada` (←) | `internal/cache` | la regla exportada es la de la caché; la superficie gana solo esa función |
| tests de `internal/graph` | `internal/graph/*_test.go` | ruta (opción, variable, `HOME`), cadenas de conexión, nombre de los auxiliares (ruta resuelta fuera de Windows), comprobación de escritura (`nil`, `fs.ErrNotExist`, otro error), modo de lectura según los auxiliares y el permiso, estados de data-model §3.1, esperas (tramos, contexto, espera propia), mensajes con `world.db`, `Nulo`, superficie sin la base; `publicar_test.go`: publicación del temporal con la costura de `os.Link` (`fs.ErrExist`, otro error, directorios retirados); `aplicar_test.go`: lo declarado para un `world.db` sin esquema fuera de WAL, con las cotas afirmadas en cada caso (mismas tablas y filas, `wal`, versión 0, ningún fichero nuevo) y el resultado exacto de cada base que construye, con la lista literal de bytes distintos y el tamaño —0 bytes → 4096; cabecera de este controlador → 18, 19, 27, 95; con otra versión en 96-99 → además 98, 99; con `version-valid-for` distinto del contador y tamaño 0 en 28-31 → además 31; con acarreo del contador → 18, 19, 26, 27, 94, 95; `auto_vacuum=full` con páginas libres al final → de 176 128 a 12 288 bytes y 18, 19, 27, 31, 35, 39, 95; con una página en uso detrás de las libres → de 188 416 a 24 576 bytes y los 59 de V45; `auto_vacuum=INCREMENTAL` → 18, 19, 27, 95; diario frío de `PERSIST` y vacío de `TRUNCATE` → 18, 19, 27, 95 y sin diario; 0 bytes con diario vacío → 4096 y sin diario— (contracts/almacen-world-db.md §7, V41, V45) |
| matriz de FR-088 | `internal/graph/integracion_test.go` (`integration`) e `integracion_enlace_test.go` (`integration && unix`) | contracts/almacen-world-db.md §7, con la lectura con un `-wal` huérfano y con un escritor abierto (`world.db` y `-wal` iguales; `-shm` presente), con un `-shm` suelto (aparece un `-wal` vacío, V48), con un `world.db-journal` caliente (1, `inesperado`, `world.db` y el diario iguales), frío, vacío o de un escritor vivo (se lee; nada cambia), sin permiso de escritura sobre `world.db` (se lee y ningún fichero cambia ni aparece; la entrega falla sin tocar nada, tampoco con un diario caliente; V46), con `world.db` como enlace a un destino con un `-wal` huérfano y como enlace sin destino (V47), la recuperación declarada en una entrega que falla (con un `-wal` huérfano: `world.db` con lo confirmado y sin auxiliares; con un diario caliente: deshecho) y la ausencia de residuo con `world.db` ausente |
| `TestObservadoDeBoe` | `internal/source/boe/grafo_test.go` | FR-040, FR-041, FR-065, FR-071: ids, datos, aristas, texto, sin ELI, verbos que no emiten |
| `TestEntregaDelMontador` | `internal/cli/entrega_test.go` | contracts/resultado-y-entrega.md §4 |
| `TestGlobales` (←) | `internal/cli/globales_test.go` | la ayuda literal de `--no-graph` |
| `TestEntregaDelKernel` | `internal/app/main_test.go` | `--no-graph` → `graph.Nulo`; plazo de `--timeout`; `--dry-run` y fallo sin entrega; registro sin almacén sin entrega |
| `TestAppletGrafo`, `TestCodigosDelGrafo` | `internal/app/grafo_test.go` | contracts/applet-graph.md §1-§4 sobre un registro local; versión posterior en los tres verbos → 1 y fichero intacto |
| `TestSalidaDelGrafoContraSchemas` | `internal/app/grafo_test.go` | toda salida correcta de `graph` contra `schemas/grafo.json` (FR-051) |
| `TestLaEntregaLlevaLaProcedenciaDelSobre` | `internal/app/grafo_test.go` | FR-089, SC-002: fuente, url y fecha de cada nodo, arista y texto creados o actualizados = las del sobre; lo no tocado conserva la suya; ninguna operación sin fuente |
| `TestNingunVerboDelGrafoDevuelveTexto` | `internal/app/grafo_test.go` | FR-070, SC-006 sobre todas las respuestas grabadas de `articulo` y `articulos` |
| `TestLaSalidaDeBoeNoCambiaConElGrafo` | `internal/app/grafo_test.go` | SC-003: cada respuesta grabada de `boe articulo` en H4 y H5.1, con y sin `--no-graph`, con y sin `--json`, en proceso, con la hora de la reproducción fijada (`httpx.ConHora`) y el almacén en `t.TempDir()`, para que `fecha_consulta` sea la misma en las dos invocaciones |
| `TestGrabacionesDerivadas` | `internal/app/grafo_test.go` | cada derivada es su grabación salvo lo que dice su nombre (D22) |
| `TestCosteDelGrafo` | `internal/app/coste_test.go` | SC-007, SC-008, con el binario de e2e, no paralelo |
| `TestBinariosDelArnes` (←) | `internal/app/e2e_test.go` | los tres binarios con reloj (contracts/arnes-e2e.md §2) |
| `TestRegistroDeProduccion` (←), `TestRegistroDeE2E` (←), `TestPuntoDeEntrada` (←) | `internal/app`, `ejemplo/kitlegal-e2e`, `cmd/kitlegal` | registro `boe, graph, skills, territorio`; `graph` sin verbo → 2 con sus tres verbos |
| `TestEsquemasPublicados` (←) | `internal/app/esquemas_test.go` | fila `grafo.json`, entidad `grafo`, verbos `check`, `show`, `stats` |
| `TestArquitectura` (←) | `internal/arch_test.go` | R6 transitiva, con `internal/graph` en el grafo |
| `TestLeerEval`, `TestEsquemaDeEval`, `TestFormaDelComando`, `TestJuzgar`, `TestInforme`, `TestPrepararGrafoPrevio`, `TestEvalsDelRepositorio` (←) | `internal/evals` | contracts/evals-y-skill.md §6 |
| `TestSkillsDelRepositorio`, `TestOrdenesDeLasSkillsEmpotradas`, `TestTablaDeComandosCoincideConLaGramatica` | `internal/app/skills_test.go` (sin cambios) | la tabla de `graph` en `boe-legislacion`; cada fila con `--describe` sale con 0 |
| `TestEntregaDelHito` | `internal/app/e2e_test.go` | los 11 guiones `h7-*` y los existentes |

### Tests existentes que cambian, por paso

Barrido hecho con `git grep` sobre los campos de `schema.Resultado`, `.Emitir(`, las listas literales de applets,
`! exists cache` y la ayuda de `--no-graph` (V16, V17, V27). La tarea de cada paso **declara estos ficheros y los
cambia en el mismo diff**.

| Paso | Fichero (en `main` hoy) | Qué lo rompe | Cambio |
|---|---|---|---|
| 2 | `internal/core/schema/sobre_test.go:195-196` | `Resultado` gana `Grafo` | cinco campos |
| 5 | `internal/cache/superficie_test.go:68-89` | se exporta `Directorio` | una entrada más |
| 8 | `internal/cli/sobre_test.go`, `internal/cli/describe_test.go` (3 llamadas a `Emitir`) | `Emitir(ctx, …)` | `context.Background()` o `t.Context()` |
| 8 | `internal/cli/globales_test.go` | la ayuda de `--no-graph` | la frase de FR-031 |
| 12 | `internal/app/registro_test.go:245-246`; `cmd/kitlegal/main_test.go:19`; `internal/app/ejemplo/kitlegal-e2e/main_test.go:90`; `internal/app/testdata/script/argumentos.txtar:24`, `:30` | registrar `graph` | `boe, graph, skills, territorio` (y `boe, contar, echo, graph, skills, territorio` en e2e) |
| 12 | `internal/app/testdata/script/territorio-matriz.txtar:206` | la entrega crea `cache/world.db` | `! exists cache/cache.db` (D29) |
| 12 | `internal/app/esquemas_test.go:59-64` (`ficherosDeEsquemas`) | `TestEsquemasCubrenTodosLosVerbos` (l. 337) exige la parte publicada de cada verbo | fila `grafo.json` |

**Lo que no cambia**, y por qué: los guiones cronometrados `internal/app/testdata/script/boe-cache-rapida.txtar`
(diez `cronometra 200ms` de `boe articulo` desde la caché, con `cmp stdout` y `! stderr .`) y los tres
`cronometra 200ms` de `territorio-matriz.txtar:192-202`: desde el paso 12 cada una de esas invocaciones entrega, sin
cambiar nada del grafo ni de su salida estándar o de error, y tiene que caber en su cota (V38, S6); si en el runner no
cabe, se optimiza la entrega en `internal/graph`, nunca el guion ni la cota. `boe-codigos.txtar:41` y
`boe-verbos.txtar:58` afirman `! exists cache` solo tras invocaciones que fallan o no ejecutan, que no entregan (V27); `ayuda.txtar` compara por expresión y `territorio`
sigue siendo el nombre más largo; los tests de `internal/source/boe` que comparan un `Resultado` entero lo hacen bajo
`--dry-run`, sin emisión (si alguno de éxito lo hiciera, la tarea del paso 9 lo amplía con el `Observado` esperado);
los tests que ejecutan verbos con `RegistroDeProduccion` solo piden `--describe` o fallan, y no entregan.

### Fixtures, `testdata/` y `schemas/` (tareas `[datos]`, FR-095)

| Fichero | Cambio | Por qué |
|---|---|---|
| `schemas/grafo.json` | nuevo, con `TestEsquemasPublicados -actualizar-esquemas` | FR-051 |
| `schemas/eval.yaml.json` | `comando-comprobacion`, `prohibidos`, `grafo_previo` | FR-086 |
| `internal/app/testdata/derivadas/version-posterior/…_texto_bloque_a21.json` | nuevo, derivado de la grabación de H4 | FR-090 (D22) |
| `internal/app/testdata/derivadas/sin-eli/…_metadatos.json`, `…/eli-sin-segmento/…_metadatos.json` | nuevos, derivados | US1.5, FR-040 (D22) |
| `testdata/evals/grafo-previo/lpac-a21-version-anterior/…_texto_bloque_a21.json` | nuevo, derivado | FR-085 (D22) |
| `internal/app/testdata/script/argumentos.txtar`, `territorio-matriz.txtar` | aserciones (paso 12) | V27, D29 |
| `internal/app/testdata/script/h7-*.txtar` | los copia el workflow al activar la suite | ADR 0018 |

Ninguna grabación nueva de ninguna fuente.

### CI

Sin cambios en los flujos: el trabajo `ci` sigue siendo `make ci`; el job de evals ya se dispara con `evals/*`,
`internal/*` y `skills/*`.

## Datos externos

**Ninguno.** H7 no graba ni deriva nada de ninguna fuente: no hay manifiesto `grabaciones.json` nuevo, ni test
`//go:build grabacion`, ni paso `grabar_datos`, ni material en `evidencias/h7/`, ni fila nueva en `docs/SOURCES.md`
(D32). Las cuatro fixtures derivadas salen de grabaciones de H4 ya versionadas y revisadas, por tareas `[datos]` y
comprobadas por `TestGrabacionesDerivadas`; `world.db` es un almacén local.

## Orden de implementación (de dentro afuera)

1. **T001 `[aceptacion]`**: los 11 guiones contra contracts/arnes-e2e.md, cada uno con la precondición del applet; sin
   código de producto y sin evals (D25).
2. **Tipos**: `internal/core/schema/grafo.go` + `Resultado.Grafo` + `sobre_test.go`; `internal/core/graphstore.go` y
   el comentario de `internal/core/doc.go`.
3. **Dominio, validación**: `internal/core/grafo` —`vocabulario.go`, `canonico.go`, `persona.go`, `lote.go`,
   `errores.go`, `id.go`— con sus tests.
4. **Dominio, fusión y comprobación**: `observacion.go`, `salida.go`, `comprobar.go`, `explicacion.go` con sus tests;
   y, en la misma tarea, `observacion` en `misspell.ignore-rules` (`.golangci.yml`), porque `salida.go` introduce las
   claves `primera_observacion` y `ultima_observacion` (D31).
5. **`cache.Directorio`** con `TestDirectorio` y la entrada de `superficie_test.go`.
6. **`internal/graph`**: ruta, errores, espera, migración `0001_grafo.sql`, apertura (nombre de los auxiliares,
   comprobación de escritura y lectura según los auxiliares y el permiso, D10), lectura, `Apply` con la comprobación de
   escritura, la publicación del temporal por `os.Link` y su costura no exportada (D11), `Nulo`, con sus tests —entre
   ellos `publicar_test.go` y las cotas y los ejemplos exactos de `aplicar_test.go`—; y la regla R6 (`.golangci.yml`, lista `grafo`; `internal/arch_test.go`, subprueba R6 y comentario de R3),
   que exige que `internal/graph` exista. Verificación de R6: una sonda temporal que importa `internal/render` desde
   `internal/graph` pone `make lint` y `TestArquitectura` en rojo y se retira antes de `make ci`.
7. **Integración del almacén**: `internal/graph/integracion_test.go` (matriz de FR-088) e
   `internal/graph/integracion_enlace_test.go` (`integration && unix`).
8. **Kernel**: `internal/cli/entrega.go`, `sobre.go` (`Montador.Grafo`, `Emitir(ctx, …)`), `globales.go`, sus tests y
   los que la firma rompe; `internal/app/main.go` (desenlace, montador, contexto con el plazo),
   `internal/app/registro.go` (`EntregarAlGrafo`, sin registrar nada todavía) y `TestEntregaDelKernel`.
9. **Emisión de `boe`**: `internal/source/boe/grafo.go` y `articulo.go` con `TestObservadoDeBoe`.
10. **Emisión de `territorio`**: `internal/core/territorio/grafo.go` e `internal/app/territorio.go` con
    `TestObservadoDeTerritorio`.
11. **Applet `graph` sin registrar** (patrón 6a de H19): `internal/app/grafo.go` con `TestAppletGrafo` y
    `TestCodigosDelGrafo` sobre un registro local del test.
12. **`[datos]` indivisible, y nada más**: registrar `graph` y `EntregarAlGrafo(graph.Nuevo())` en
    `internal/app/registro.go` y en `internal/app/ejemplo/kitlegal-e2e/main.go`; `schemas/grafo.json` generado; la
    fila de `esquemas_test.go`; las cuatro listas literales de applets; y `territorio-matriz.txtar:206` (*Complexity
    Tracking*). Desde esta tarea el binario de e2e entrega, así que los `cronometra 200ms` de
    `internal/app/testdata/script/boe-cache-rapida.txtar` (diez) y de `territorio-matriz.txtar:192-202` (tres), con sus
    `! stderr .`, miden también la entrega: la tarea **no** los toca ni los declara, y su `make ci` (que incluye
    `make test-tiempos`) los tiene que pasar. Si no caben —en local o, después, en el runner (S6)—, el arreglo es
    optimizar la entrega que no cambia nada en `internal/graph`, en una tarea anterior que declare esos ficheros de
    código, nunca cambiar un guion ni una cota.
13. **Salida contra el esquema publicado**: `TestSalidaDelGrafoContraSchemas`, `TestLaEntregaLlevaLaProcedenciaDelSobre`,
    `TestNingunVerboDelGrafoDevuelveTexto` y `TestLaSalidaDeBoeNoCambiaConElGrafo`.
14. **`[datos]` derivadas del e2e**: los tres ficheros de `internal/app/testdata/derivadas/`, y nada más.
15. **Arnés**: `internal/app/e2e_test.go` (tres binarios con reloj, copia de `derivadas/`, `TestBinariosDelArnes`),
    la variable `reloj` en `internal/app/ejemplo/kitlegal-e2e/main.go` y su test, y `TestGrabacionesDerivadas`. Desde
    aquí la suite congelada se puede ejecutar copiándola un momento a `testdata/script/` para medir el avance (copias
    `zz-`, retiradas antes de `make ci`).
16. **Costes**: `internal/app/coste_test.go` y `MEDIDAS_DE_TIEMPO` en el `Makefile`.
17. **`[datos]` esquema de eval**: `schemas/eval.yaml.json` y, si el `else` ampliado cambia algún mensaje que
    `internal/evals/formato_test.go` compara literalmente, esa expectativa en la misma tarea (patrón D28 de H6;
    *Complexity Tracking*).
18. **Formato, juicio, preparación e informe**: `internal/evals/{formato,juzgar,consultas,preparar,grabaciones,informe}.go`
    y sus tests, incluido `TestPrepararGrafoPrevio` sobre temporales.
19. **`[datos]` derivada de la eval**: `testdata/evals/grafo-previo/lpac-a21-version-anterior/…_texto_bloque_a21.json`,
    y nada más.
20. **La eval**: `evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml` y, en la misma tarea, la subprueba
    `grafo-previo` de `TestEvalsDelRepositorio` (`internal/evals/conjunto_test.go`), que la exige.
21. **La skill**: `skills/boe-legislacion/SKILL.md` (frontmatter `kitlegal-applets: boe graph`, protocolo de FR-080 a
    FR-082) y la tabla regenerada con `make skills-sync`.
22. **Documentación**: `CHANGELOG.md` (*Unreleased*), `README.md` y `CONTRIBUTING.md` (applets, `graph`, `world.db`,
    `test-tiempos`).

## Complexity Tracking

| Violación o pieza nueva | Por qué es necesaria | Alternativa más simple rechazada porque |
|---|---|---|
| Paquete de dominio `internal/core/grafo`, no listado en `CLAUDE.md` | Validación, `Persona`, fusión de FR-023 y reglas de `check` son dominio puro que se prueba exhaustivamente sin E/S y cuenta en el umbral de `internal/core/**` (D1, D15) | La lógica en `internal/graph`: reglas de dominio en un adaptador, en SQL y fuera del umbral |
| `internal/graph` importa `internal/cache`, y `internal/cache` exporta `Directorio` (su superficie cerrada gana una entrada) | FR-001: `world.db` vive junto a la caché «con su misma regla»; una sola función la hace idéntica por construcción (D8). La función no nombra la base, así que FR-005 de H3 se mantiene | Copiar la regla: dos reglas que pueden divergir. Paquete compartido: R3 solo concede SQLite a tres paquetes. Refactorizar la caché: rompe FR-005 de H3 y no lo pide el hito |
| `cli.Montador` gana `Grafo` y `Emitir` recibe `context.Context` (API del kernel) | La entrega tiene que ir detrás de presentar, con la fecha del sobre presentado y con el plazo de `--timeout` (FR-014, FR-026, D4) | Rehacer el sobre en `Main`: otra fecha si la puso el reloj. Un gancho en `internal/app`: parte la garantía de la procedencia entre dos paquetes |
| La línea de aviso de una entrega fallida no propaga el error de su escritura | FR-033: el código y la salida ya están decididos; un test lo fija (D7) | Propagarlo cambiaría el código de un resultado ya presentado con `ok: true` |
| Andamiaje SQLite propio en `internal/graph` con el diseño de la caché y cuatro diferencias (lectura con `mode=rw` + `query_only` sin auxiliares y con permiso de escritura, con `mode=ro&immutable=1` sin permiso y sin `-wal` ni `-journal`, y con `mode=ro` con auxiliares —`-wal`, `-shm` o `world.db-journal`, buscados con el nombre que les da SQLite, junto al destino de un enlace fuera de Windows—; comprobación del permiso de escritura con `os.OpenFile(…, os.O_RDWR, 0)` antes de abrir, al leer y al entregar; creación de `world.db` en un temporal publicado con `os.Link`; WAL fijado fuera de la cadena de conexión y de toda transacción en un `world.db` que ya existe sin esquema, comprobando que el pragma devuelve `wal`) | Las sondas V9-V11, V36, V42-V43 y V46-V47 muestran que el modo de lectura y el WAL de la caché dejarían auxiliares —también, sin permiso de escritura, `-wal` y `-shm` que nadie puede borrar—, reescribirían `world.db` al cerrar sobre un `-wal` huérfano (también el que está junto al destino de un enlace) o al leer con un diario caliente, o dejarían un `world.db` de 4096 bytes al fallar, contra FR-004, FR-031 y FR-033, y que el pragma de WAL falla en silencio sobre 0 bytes dentro de una transacción (D9-D11) | Reutilizar el de la caché tal cual: ver sondas. Crear en su sitio: el residuo de la primera versión de este plan (D11). Comprobar el permiso con `access(2)`: usuario real y no efectivo, sin Windows y un import nuevo de `golang.org/x/sys` (D10) |
| Costura no exportada que sustituye a `os.Link` en los tests de `internal/graph` (`publicar_test.go`) | Las dos salidas de la publicación que no son el éxito —`fs.ErrExist` porque otra invocación publicó antes, y cualquier otro error— no se pueden provocar de forma determinista desde fuera, y son las que deciden que un fallo no deje nada (FR-033) y que la carrera de SC-009 no pierda ningún lote (D11) | No probarlas: la limpieza y la carrera quedarían sin control. Un sistema de ficheros sin enlaces duros en CI: no existe en el runner |
| **Desviación declarada de FR-033, paso a WAL de una base de fuera**: una entrega que falla después de fijar WAL en un `world.db` que ya existía **sin esquema y fuera de WAL** (0 bytes, o base en rollback sin el esquema, escrita por cualquier programa con cualquier versión y configuración de SQLite) deja lo que SQLite escribe al confirmar el paso a WAL. **Cotas**, para cualquier base: su contenido (tablas, filas, esquema) no cambia, sigue en versión 0 y se lee como grafo vacío, y cerrada la última conexión no queda ningún fichero nuevo. **Qué cambia**: bytes de la cabecera de la página 1 (siempre 18, 19, 24-27 y 92-95; 28-31, 32-39 y 96-99 si no coincidían con lo que la confirmación calcula); con `auto_vacuum=full` y páginas libres, el vaciado de todas —reubicando las páginas en uso que están detrás— y el truncado del fichero; un `world.db-journal` frío o vacío desaparece; 0 bytes → 4096 (V36 A, B, V41, V45) | Crear el esquema de forma atómica exige una transacción, y el modo WAL (FR-003) no se puede fijar dentro de ella —en rollback el pragma falla y sobre 0 bytes devuelve `delete` sin error (V42)— ni sin que SQLite confirme, y una confirmación hace lo que hace siempre. Ese `world.db` no lo crea nunca el binario (con `world.db` ausente se publica un temporal completo y un fallo no deja nada): solo llega de fuera, y lo que queda no cambia su contenido. Los conjuntos medidos (cabeceras de este controlador y de otra versión, tamaño a 0, acarreo, `auto_vacuum=full` con libres al final y delante de una página en uso, diarios de `PERSIST` y `TRUNCATE`) son ejemplos que `aplicar_test.go` fija con su resultado exacto, junto con las cotas en cada caso (contracts/almacen-world-db.md §4.1, §7); en `gates/supuestos.md` | Sustituirlo por un temporal con `os.Rename`: lo que entregue otra invocación que ya lo tenga abierto iría a un fichero desenlazado. Crear el esquema en rollback y fijar WAL tras confirmar: una base de versión 1 fuera de WAL (contra FR-003) y el lote del grafo escrito con un diario de rollback, que una interrupción dejaría caliente y los verbos de `graph` no podrían leer (D11, V43). Evitar el vaciado con `sqlite3_autovacuum_pages`: API C del controlador sin verificar (D11). Enumerar los bytes como lista cerrada para cualquier base: no se puede comprobar (V45) |
| **Desviación declarada de FR-033, recuperación de SQLite**: con un `-wal` huérfano o un diario de rollback caliente (un escritor interrumpido), la conexión de la entrega los recupera aunque la entrega falle después (rechazo, esquema posterior, plazo o espera agotados, E/S): el diario se deshace en su primera lectura y desaparece (V43); el `-wal` se lleva a `world.db` en el checkpoint del cierre, si es la última conexión, y `-wal` y `-shm` desaparecen (V44). Cambian los bytes y el tamaño de `world.db`; el contenido del grafo es el confirmado | La recuperación la hace SQLite al abrir o al cerrar una conexión de escritura, antes de saber si el lote entra, y es lo que hace cualquier escritor, también la primera entrega que sí entra. Lo fija la matriz de integración (contracts/almacen-world-db.md §4.1, §7); en `gates/supuestos.md` | Leer antes con `mode=ro` y abrir después para escribir: solo la evita en los fallos que esa lectura ve (esquema posterior, rechazo contra lo guardado), no en un plazo o una espera agotados ni con un diario caliente, que `mode=ro` no deja leer (V43), y añade una segunda apertura con su carrera. Desactivar el checkpoint al cerrar: exige la API C del controlador, sin verificar (D10) |
| **Desviación declarada de FR-004, FR-031 y SC-004**: con auxiliares de WAL (otra conexión abierta, un `-wal` huérfano de un escritor interrumpido o un `-shm` suelto), los verbos de `graph` escriben lo que SQLite necesita para leer lo confirmado en ellos: reescriben o crean `world.db-shm` y, junto a un `-shm` suelto, crean un `world.db-wal` de 0 bytes. **Cota**: `world.db`, un `world.db-wal` que ya existía, el diario y el contenido del grafo no cambian (V36 D, F, V48) | SQLite guarda en `-shm` el índice del WAL y las marcas de lectura de cada lector: ningún modo que lea lo confirmado en el WAL evita escribirlo. Sin auxiliares —el estado de los guiones, del quickstart y de toda invocación que no coincide con otra— no cambia ni un byte, pueda el proceso escribir `world.db` o no y sea un fichero o un enlace (V9, V36 E, V46, V47); con un `world.db-journal` tampoco: se lee con `mode=ro` sin cambiar nada o, si está caliente, el verbo sale con 1 (V43, V45 c). Lo fija la matriz de integración (contracts/almacen-world-db.md §7); en `gates/supuestos.md` | `mode=rw` + `query_only` siempre: sobre un `-wal` huérfano, el cierre reescribe `world.db` y borra los auxiliares (V36 D), y con un diario caliente lo deshace al leer (V43). `immutable=1`: no lee lo confirmado en el WAL y ve las páginas sin confirmar de un diario caliente (V43). Copiar a un temporal: instantánea incoherente con un escritor a la vez (D10) |
| Residuo ante una señal (fuera de FR-033, declarado): un proceso terminado por una señal a mitad de una entrega puede dejar `world.db-nuevo-*` con sus auxiliares y los directorios que creó | El binario no atiende señales (V39) y FR-033 enumera los fallos que el programa trata; ningún lector ni entrega mira el temporal (D11); en `gates/supuestos.md` | Atender `SIGINT`/`SIGTERM` para limpiar: no lo pide el hito (lectura conservadora). Borrar temporales viejos en la entrega siguiente: podría borrar el de otra invocación en curso |
| Tres binarios más en el arnés e2e, con reloj fijo por `-X` | `fecha_consulta`, `version-obsoleta` y `fuente-caducada` literales y reproducibles; el binario de e2e no lee su comportamiento del entorno (D23, V22) | Reloj real: resultados según el día; variable de entorno: contra la regla del binario de e2e |
| `[datos]` que mezcla código: registrar `graph` y la entrega (producción y e2e) + `schemas/grafo.json` + fila de `esquemas_test.go` + las cuatro listas de applets + `territorio-matriz.txtar:206` (paso 12), **y nada más** | En cuanto `graph` está registrado, `TestEsquemasCubrenTodosLosVerbos` exige su parte publicada, que se genera desde el applet registrado; las listas literales cambian a la vez; y en cuanto el e2e entrega, `territorio-matriz.txtar` deja de valer. Mismo patrón que D16 de H6 y el paso 6b de H19 | Registrar en una tarea y publicar en otra deja `make ci` en rojo; meter el applet entero en la `[datos]` mezcla código que ningún control ata al esquema (va antes, en el paso 11) |
| `[datos]` del esquema de eval con la expectativa de `formato_test.go` si su mensaje cambia (paso 17) | El esquema se compila del fichero real y un caso compara el mensaje con `EqualError`: separarlos dejaría un rojo entre dos tareas (patrón D28 de H6) | Relajar el test a `ErrorContains`: arreglar el control en lugar del contrato |
| La eval de FR-085 fuera de T001 | Exige el formato ampliado y `make ci` valida cada eval (D25, V23, V24) | Escribirla en T001: `make ci` en rojo en la tarea de aceptación |
| `TestCosteDelGrafo` en `MEDIDAS_DE_TIEMPO` | Una medida de reloj en paralelo con todo el módulo mide la carga (D28, V31) | Medir dentro de `make test` |
| Aserción de `territorio-matriz.txtar:206` cambiada | H7 hace que la entrega cree `world.db` en el directorio de la caché, que es lo que el spec pide (FR-001); `territorio` sigue sin tocar la caché (D29) | Mantenerla: exigiría no entregar lo que `territorio` observa, contra FR-043 |
| `observacion` en `misspell.ignore-rules` | Las claves JSON `primera_observacion` y `ultima_observacion` nombran lo que FR-053 llama «primera y última observación», y el `_` no protege del diccionario (V30, D31) | Otras claves («primera_vez»): peor vocabulario del contrato |

## Obligaciones que este plan traslada a `tasks.md`

1. **T001 es la única tarea `[aceptacion]`** y la primera: escribe los 11 guiones de «Aceptación e2e» en
   `specs/010-h7-internal-graph-grafo/aceptacion/`, cada uno con la precondición de contracts/arnes-e2e.md §1 como
   primeras órdenes y un comentario con sus FR/SC; del arnés usa solo lo de contracts/arnes-e2e.md, y cada formato lo
   copia del spec o del contrato que nombra su §5, sin inventar otro; ningún valor de memoria (§5 del contrato). Sin
   código de producto y **sin evals** (D25).
2. Los nombres de los binarios, sus relojes, las variables, `$WORK/derivadas/` y las formas de §4 son los de
   contracts/arnes-e2e.md, tal cual: la suite congelada depende de ellos.
3. Registrar `graph` y la entrega, publicar `schemas/grafo.json`, su fila, las cuatro listas y
   `territorio-matriz.txtar:206` van en **una** tarea `[datos]` (paso 12), y nada más; el applet y sus tests, antes y
   sin registrar (paso 11); la salida contra el esquema, después (paso 13).
4. Cada tarea deja `make ci` en verde. Ninguna tarea comprueba lo que un paso posterior crea: R6 nace con
   `internal/graph` (paso 6); `TestGrabacionesDerivadas` y la copia de `derivadas/` en el arnés, después de las
   derivadas (pasos 14-15); la subprueba `grafo-previo`, con la eval (paso 20).
5. Ninguna tarea usa la red, `KITLEGAL_RECORD`, `make evals` ni `make verify-sources`, ni escribe en `evidencias/`;
   ninguna publica, abre la propuesta de cambio ni mide CI o evals remotas: el cierre lo hace el workflow.
6. Tareas explícitas de Definition of Done: `CHANGELOG.md` (*Unreleased*), `README.md` y `CONTRIBUTING.md` alineados
   con el `Makefile` y con los applets, `schemas/grafo.json` y `schemas/eval.yaml.json`, las derivadas y la lista de
   ficheros de `testdata/` y `schemas/` tocados para el informe final (FR-095). Sin ADR nuevo (el ADR 0014 deja los
   nombres al plan) ni fila de `docs/SOURCES.md` (ninguna fuente).
7. Cada tarea declara en sus rutas, y cambia en su diff, los tests existentes que su paso rompe según «Tests
   existentes que cambian, por paso», y no toca los que esa sección da por no cambiados.
8. **Sin atajos**: ningún `//nolint` nuevo salvo el `//nolint:paralleltest` razonado de `TestCosteDelGrafo` (mide con
   el reloj de pared, como `TestMedidasDeTiempo`); ningún `t.Skip`, TODO ni error silenciado; ninguna exclusión de lint
   salvo `observacion` en `ignore-rules` (D31). Si `dupl` marca un clon con `internal/cache`, se reestructura (D9).
9. **`misspell`**: ningún identificador suelto `Observacion` ni `Emision`; los comentarios llevan tilde
   («observación», «emisión»); el resto del vocabulario del hito está comprobado (V30).
10. **Ningún test escribe en `~/.cache/kitlegal`**: todo registro de test sin almacén o con `graph.ConDirectorio` bajo
    `t.TempDir()`; todo guion con `KITLEGAL_CACHE_DIR` del arnés (D6).
11. **Valores de los datos, no de memoria**: el DIR3 de Leganés es `L01280745` (`data/territorio/dir3.yaml`), no el
    del ejemplo de `refs/`; la fecha de `territorio` es `2026-02-04T00:00:00Z`; los `hash_texto` salen de las
    grabaciones (V28).
12. **Rutas declaradas**: cada fichero por su ruta completa; lo que solo se lee o se ejecuta, por su nombre de test u
    objetivo de `make`; desde la tarea siguiente a cada `[datos]`, ninguna línea deja extraer `schemas/`,
    `testdata/`, `internal/app/testdata/` ni un directorio que contenga material protegido.
13. **Cotas de tiempo existentes intactas** (S6): ninguna tarea toca `boe-cache-rapida.txtar` ni las líneas
    `cronometra` de `territorio-matriz.txtar`; si la entrega no cabe en ellas, se optimiza `internal/graph` (p. ej.,
    decidir con una lectura, antes de `BEGIN IMMEDIATE`, que el lote no cambia nada).
14. **Clases de caracteres**: la expresión de `Persona` es la de contracts/almacen-world-db.md §5, sin `(?i)`, y
    `ValidarID` usa `unicode.IsSpace` y `unicode.IsControl` (D34, D35); los casos de los tests son los de esos
    contratos, con los caracteres que no son ASCII escritos con su escape de Go.
15. **Las desviaciones declaradas se afirman, no se esconden**: `aplicar_test.go` afirma, en cada base que construye,
    las cotas de contracts/almacen-world-db.md §4.1 (mismas tablas y filas, `wal`, versión 0, ningún fichero nuevo) y
    su resultado exacto —la lista literal de bytes distintos y el tamaño: la cabecera de este controlador, otra versión
    de SQLite en 96-99, tamaño en cabecera no válido, acarreo del contador, `auto_vacuum=full` con páginas libres al
    final y delante de una página en uso, `auto_vacuum=INCREMENTAL`, y los diarios de `PERSIST` y `TRUNCATE`, que ya
    no existen (V41, V45)—, sin darla por la lista de cualquier base; la matriz de integración, `world.db-shm` presente
    tras leer con auxiliares de WAL, el `-wal` vacío junto a un `-shm` suelto, un diario caliente que no se lee y no
    cambia, la recuperación de un `-wal` huérfano y de un diario caliente en una entrega que falla, y que sin permiso
    de escritura y con un enlace no cambia ni aparece nada (V46, V47); ninguna otra diferencia de bytes se da por
    buena en esos casos.

## Comprobación contra la rúbrica del juez (`juez_plan`, a-l) y `precheck.sh plan`

| Criterio | Dónde se cumple |
|---|---|
| a. constitution_check | «Constitution Check»: un ítem por principio I-IX (VII con la expresión corregida, D34); uno por regla de dependencia (las cinco de §IV, la del grafo, la de ejemplos, la del espacio reservado, la de módulos del binario y la de la superficie de la caché); gates por capa; reglas del modo desatendido; re-evaluación con los estados de fuera que el diseño evita (sin permiso de escritura, V46; enlace simbólico, V47) y las tres desviaciones declaradas por su causa y con sus cotas, con los estados medidos como ejemplos fijados por test y no como lista cerrada —FR-033 por el paso a WAL de una base de fuera (V41, V45); FR-033 por la recuperación de SQLite (V43, V44); y FR-004/FR-031/SC-004 por los auxiliares de WAL (V36, V48)— en *Complexity Tracking* |
| b. dependencias | Ninguna dependencia nueva; biblioteca estándar con precedente (V5); `TestDependenciasDelBinario` sin cambios |
| c. reglas_dependencia | Tabla de reglas; `internal/core/grafo` sin E/S; SQLite solo en `internal/graph`; R6 nueva en `depguard` y `TestArquitectura`; la entrega escribe por el `Presentador`; ningún `os.Exit` nuevo |
| d. errores_exit_codes | contracts/applet-graph.md §4 y almacen-world-db.md §6: `*graph.Error` con `schema.ConClase` (1, 2, 4), `no-encontrado` (3), hallazgos con 0 (ADR 0023); la entrega fallida no cambia el código (FR-033) |
| e. tests_primero | «Aceptación e2e» (11 guiones congelados con su FR/SC y lo que cubre cada otra aceptación), «Tests», «Tests existentes que cambian, por paso», «Fixtures», «Objetivos del Makefile» |
| f. alcance | Solo FR-001 a FR-095; lo que el spec deja fuera no aparece; ninguna opción sin consumidor (la API de `internal/graph` solo tiene `ConDirectorio`); lo que el diseño añade está en *Complexity Tracking* |
| g. sin_atajos | Obligaciones 8-10 y 15; el único error no propagado está razonado y probado (D7); los fallos de la limpieza del temporal se unen a la causa (D11) |
| h. mejor_alternativa | research D1-D35, cada una con su alternativa rechazada; D9-D11 apoyadas en sondas (V9-V11, V36, V41-V48), con `world.db` sin permiso de escritura (`immutable` al leer y fallo antes de abrir al entregar, frente a abrir sin comprobar, `access(2)`, leer la cabecera o borrar lo que quede) y con `world.db` como enlace (auxiliares donde SQLite los crea, frente a junto al nombre o siempre junto a la ruta resuelta); las clases de caracteres de FR-025, FR-064 y FR-052 decididas en D34 y D35 |
| i. afirmaciones_verificadas | research V1-V48 con `fichero:línea`, `go doc` o sonda (V6 y V9 corregidas por V36 y V37; el residuo de V36 B generalizado a otras cabeceras por V41 y a páginas libres y diarios fríos o vacíos por V45, con `_autoVacuumCommit`, `_pager_end_transaction` y `_hasHotJournal`; WAL dentro de una transacción en V42; diario y `-wal` huérfano en V43-V44; sin permiso de escritura en V46; enlace simbólico en V47; `-shm` suelto en V48), y las referencias de línea al repositorio releídas contra `main`; S1-S7 como supuestos, entre ellos las cotas `cronometra` existentes (S6), los enlaces duros (S7) y Windows (S5); lo que se afirma de una base de fuera son las cotas de cada desviación, y los conjuntos de bytes, ejemplos medidos |
| j. quickstart_ejecutable | quickstart.md: órdenes y rutas reales, binarios en `$T`, `KITLEGAL_CACHE_DIR` bajo `$T`, `bin/` ignorado, `$T` borrado y `git status` comparado al final |
| k. datos_externos | «Datos externos»: ninguno; derivadas de grabaciones ya versionadas por tareas `[datos]` con su control |
| l. autonomia | Ninguna pausa ni persona a mitad del run; el cierre (push, propuesta, CI, evals remotas) lo hace el workflow (obligación 5) |
| `precheck.sh plan` | `plan.md` y `research.md` existen, sin marcas pendientes de aclaración, con `## Constitution Check` y la línea `Aceptación e2e:` |
