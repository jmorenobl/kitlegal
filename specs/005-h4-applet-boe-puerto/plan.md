# Implementation Plan: H4 · Applet `boe`: puerto de `boe.py`

**Branch**: `h4-applet-boe-puerto` | **Date**: 2026-09-13 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/005-h4-applet-boe-puerto/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de
`.specify/memory/constitution.md` y están registradas —decisión, alternativas y motivo— en
[research.md](./research.md). Toda afirmación sobre una herramienta, una dependencia, la biblioteca estándar o el
propio CPython en el que se apoya `refs/boe.py` está comprobada en local contra su código o su documentación, y
research.md cita dónde; lo que no se puede comprobar sin red está en
[research.md D20](./research.md#d20--supuestos-no-verificados) como supuesto, no como hecho.

## Summary

H4 es el primer hito que consulta derecho: el applet `boe` con seis verbos (`buscar`, `indice`, `articulo`,
`articulos`, `metadatos`, `analisis`) contra la API de Legislación Consolidada del BOE, con el comportamiento de
`refs/boe.py` sobre el contenido, las desviaciones que el spec declara, caché de H3, sobre citable y contratos
publicados en `schemas/`. No entrega skill: es la herramienta que `boe-legislacion` (H5) necesita (principio VIII).

Decisiones que sostienen el diseño:

1. **Puerto, adaptador y applet separados** ([D1](./research.md#d1--reparto-puerto-en-core-adaptador-en-internalsourceboe-applet-en-internalapp)):
   `core.Source` en el dominio; `internal/source/boe` implementa la fuente sin saber nada de la línea de órdenes;
   `internal/app/boe.go` declara la gramática y compone la fuente por invocación.
2. **El puerto concreto** ([D2](./research.md#d2--el-puerto-coresource-name-fetchctx-ec-consulta-ttlconsulta-terms)):
   `Fetch(ctx, ec, consulta)`, `TTL(consulta)`, `Terms()`; ADR 0015.
3. **La fecha de consulta la declara quien consulta** ([D3](./research.md#d3--la-fecha-de-consulta-la-declara-quien-consulta-schemaprocedenciafechaconsulta)):
   `schema.Procedencia.FechaConsulta`, cero = reloj del kernel; retrocompatible; ADR 0015.
4. **`httpx` pide un formato y dice cuándo emitió** ([D4](./research.md#d4--internalhttpx-gana-tres-cosas-peticionacepta-instante-y-conhora)):
   `Peticion.Acepta`, `Respuesta.Instante`/`Error.Instante` (último intento o abandono), `ConHora`.
5. **Caché por dirección de la API** ([D5](./research.md#d5--caché-claves-forma-de-la-entrada-vigencias-metadatos-compartidos-y-modos)):
   la entrada guarda `fecha_consulta`, `url` y `data`; los metadatos se comparten entre `metadatos`, `articulo` y
   `articulos`; `--offline` y `--dry-run` abren en solo lectura; el cliente HTTP solo se construye si hay que pedir.
6. **Porte literal de las lecturas** ([D6](./research.md#d6--lectura-del-bloque-xml-el-mismo-recorrido-que-elementtree),
   [D7](./research.md#d7--lectura-de-las-respuestas-json), [D8](./research.md#d8--buscar-la-misma-consulta-y-la-misma-dirección-byte-a-byte)):
   el recorrido de `ElementTree` (descendiente que no es la raíz, `tail`, normalización de líneas), las lecturas JSON
   de `boe.py` más FR-070, y la dirección de búsqueda byte a byte como `json.dumps` + `quote`.
7. **Contratos vigilados** ([D11](./research.md#d11--tipos-de-data-claves-json-y-schemas)): `schemas/norma.json` y
   `schemas/bloque.json` con cada verbo como recurso embebido con `$id`; `make schema-check` es un test que los
   regenera en memoria desde `--describe` y los compara.
8. **Datos protegidos por pasos** ([D12](./research.md#d12--fixtures-dónde-viven-cómo-se-graban-sintéticos-referencias-y-golden)):
   fixtures junto al paquete; en la pausa del manifiesto una persona fija la fila de `docs/SOURCES.md` y sus dos
   constantes, graba y escribe a mano las cinco referencias de FR-116; sintéticos antes del código de lectura; golden y
   esquemas generados por el código y revisados, en tres pasos.
9. **Sin red en `make ci`; red solo en `make verify-sources`** ([D13](./research.md#d13--tests-e2e-y-verificación-reproducción-en-el-binario-de-e2e-etiquetas-fuentes-y-grabacion),
   [D15](./research.md#d15--la-fuente-docssourcesmd-ritmo-términos-robotstxt-y-verificación-nocturna)): e2e sobre el
   binario de e2e con reproducción; verificación nocturna con etiqueta `fuentes` y apertura de incidencia.

Artefactos de diseño: [data-model.md](./data-model.md) y [contracts/](./contracts/)
([puerto-y-applet](./contracts/puerto-y-applet.md), [verbos-y-salidas](./contracts/verbos-y-salidas.md),
[httpx-acepta-e-instante](./contracts/httpx-acepta-e-instante.md), [errores-y-codigos](./contracts/errores-y-codigos.md),
[esquemas-fixtures-y-controles](./contracts/esquemas-fixtures-y-controles.md)); validación en
[quickstart.md](./quickstart.md).

## Technical Context

**Language/Version**: Go 1.27, sin cambios: `go.mod` con `go 1.27.0` + `toolchain go1.27.1`, que el `Makefile` exporta
como `GOTOOLCHAIN`.

**Primary Dependencies**: **ninguna nueva** (FR-125). Se usan las que `go.mod` ya fija y la constitución §V lista:
`alecthomas/kong` (gramática), `invopop/jsonschema` (`--describe` y enumerados), `santhosh-tekuri/jsonschema/v6`
(validación en test), `rogpeppe/go-internal/testscript` (e2e), `stretchr/testify` (tests), `temoto/robotstxt` y
`golang.org/x/time/rate` (por `internal/httpx`), `modernc.org/sqlite` (por `internal/cache`). Biblioteca estándar:
`encoding/xml`, `encoding/json`, `crypto/sha256`, `mime`, `net/url`, `unicode`, `go/parser` (tests). Al enlazar `boe`, el
binario distribuido pasa a incluir `internal/httpx` e `internal/cache` y sus módulos, que `TestDependenciasDelBinario`
fija y la PR justifica uno a uno (FR-124, D14).

**Herramientas de control**: las de H0-H3 sin cambio de versión. `.golangci.yml` añade las etiquetas `fuentes` y
`grabacion` a `run.build-tags` y seis palabras a `misspell.ignore-rules` (`administrativo`, `capitulo`, `dependencias`,
`disposicion`, `materias`, `regulares`; [D18](./research.md#d18--nombres-que-misspell-y-gosec-aceptan)); ninguna exclusión,
ningún `//nolint`.

**Storage**: la caché de H3 (`~/.cache/kitlegal/cache.db` o `KITLEGAL_CACHE_DIR`) con claves y contenido que fija la
fuente ([data-model.md](./data-model.md) §6). En tests, siempre en `t.TempDir()` o en el `$WORK` del guion.

**Testing**: `make test` (unitarios y e2e con `-race`), `make test-integration`, `make schema-check` (ahora real).
Todo offline: `httpx.Replay` sobre grabaciones reales y sintéticos, cachés en directorios temporales, reloj controlado
con `httpx.ConHora` y `cache.ConReloj`. Guiones `testscript` sobre el binario de e2e. Fuzz de ids de bloque con semillas en
`make test`. Red solo en `make verify-sources` (etiqueta `fuentes`, flujo nocturno) y en la grabación que hace una persona
(etiqueta `grabacion`).

**Target Platform**: sin cambios (binario sin cgo, `-trimpath`; CI `ubuntu-latest`; desarrollo darwin/arm64).

**Project Type**: CLI multicall; primer applet de producción y primer adaptador de fuente.

**Performance Goals**: `kitlegal boe articulo BOE-A-2015-10565 a21` con la entrada vigente en caché, en menos de 200 ms
por invocación completa sobre un binario sin detector de carreras (FR-117, SC-002). `make ci` de una PR limpia sin
cambios apreciables de duración (los tests nuevos no esperan tiempo real).

**Constraints**: solo GET por `internal/httpx`, con `Accept` explícito, identificación, `robots.txt` y un intervalo por
sitio tomado de `docs/SOURCES.md`; sin paralelismo (los bloques de `articulos` se piden en secuencia); validación de
argumentos antes de abrir la caché o construir el cliente; ningún fallo en caché; `--offline` y `--dry-run` sin crear ni
tocar la caché; ningún `panic`; ninguna escritura en stdout fuera del presentador; sobre con `fuente`
`boe.legislacion-consolidada` y `url` `https://www.boe.es/…`.

**Scale/Scope**: un paquete nuevo de producción (`internal/source/boe`, ~19 ficheros, del orden de 1500-2000 líneas de
Go de producto y más de test), un puerto (`internal/core/source.go`), ampliaciones acotadas en `internal/core/schema`,
`internal/cli` y `internal/httpx`, el applet en `internal/app`, dos raíces (`cmd/kitlegal`, binario de e2e), dos ficheros
en `schemas/`, 23 ficheros de grabación (22 recursos y el `robots.txt`), 7 escenarios sintéticos, 5 referencias, 13 golden, 5 guiones nuevos y 2 líneas de `argumentos.txtar`, dos scripts, un trabajo
nocturno, `docs/SOURCES.md`, ADR 0015 y la documentación del repositorio.

## Constitution Check

*GATE: debe pasar antes de la fase 0 y volver a evaluarse tras la fase 1.*

### Principios

| # | Principio | Cómo lo cumple H4 | Veredicto |
|---|---|---|---|
| **I** | Fuentes públicas y frontera humana | Una única fuente pública (API de Legislación Consolidada, 🟢), consultada solo con GET por `internal/httpx`: identificación, `robots.txt` (denegación → 5), reintentos acotados y un ritmo por sitio que sale de la fila de `docs/SOURCES.md` y no de un valor por omisión (FR-122, `TestFuenteCoincideConSources`). `articulos` pide en secuencia, sin paralelismo. Ninguna acción con identidad y ningún código 6 (FR-100). Los términos de uso y `robots.txt` los revisa una persona antes de grabar; si prohíben el acceso, el hito se detiene (FR-123, D15). | ✅ Cumple |
| **II** | Nada sin cita ni fuente | Todo sobre lleva `fuente: "boe.legislacion-consolidada"`, `url` de la API, `fecha_consulta` **real** de la consulta —también servido de la caché y en fallo— y `hash` (FR-002, FR-096, D3). Nunca se presenta como texto legal lo que no se interpreta (FR-014), ni el marcador `?` (FR-016), ni un artículo con la vigencia sin comprobar (FR-013); el texto de las referencias va completo (FR-060). La corrección del texto se comprueba contra la respuesta grabada del BOE con referencias que una persona escribe a mano y revisa en la pausa de la tarea `[datos]` que graba los fixtures, antes de todo código de lectura (FR-116). El rango de la norma viaja en `data` para que la skill distinga ley y reglamento (H5); las normas autonómicas se consultan igual por su `BOE-A-…`. | ✅ Cumple |
| **III** | Tests primero y offline | Cada tarea escribe su test antes del código. La verdad terreno de la entrega —las cinco referencias del diff de aceptación— la escribe y revisa una persona en la pausa de la tarea `[datos]` que graba los fixtures (paso 5), **antes** de todo código de lectura y de `articulo` (FR-116). Los guiones e2e que describen la entrega van en la tarea `[datos]` que sigue inmediatamente a la que registra `boe` en el binario de e2e (paso 12, justo detrás del 11): es la primera que puede dejarlos en verde, porque solo necesitan `boe` registrado, `Setup` con la reproducción y `KITLEGAL_CACHE_DIR`, y la orden `cronometra`, todo del paso 11; golden, esquemas y contratos, que no les hacen falta, van después (workflow, paso `tasks`). Todo test es offline contra `internal/source/boe/testdata/` (grabaciones y sintéticos); solo `scripts/verify-sources.sh` toca la red. Toda salida de los seis verbos se valida en test contra `schemas/*.json` leídos del fichero (FR-111). Umbrales de cobertura intactos: `internal/core` no gana sentencias (el puerto es una interfaz) y el kernel gana tests para su única rama nueva. | ✅ Cumple |
| **IV** | Arquitectura hexagonal con reglas ejecutables | Puerto `core.Source` en el dominio; adaptador en `internal/source/boe` que usa `core.Cache` y los tipos de `httpx`; composición manual en `internal/app` con opciones funcionales. Errores tipados: `boe.Error`, `httpx.Error` y `cache.Error` implementan `schema.ConClase` y `cli.Clasificar` los traduce sin cambios (tabla cerrada de 23 filas, [errores-y-codigos.md](./contracts/errores-y-codigos.md)). Ningún `panic`: el binario de e2e deja de usarlo (`app.Arrancar`, D16) y el fuzz fija que el analizador de ids no entra en pánico. Las reglas de dependencia, abajo. | ✅ Cumple |
| **V** | Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia nueva. Sin framework de DI ni ORM; Kong sin cambios; `internal/cli` no se extrae. Nada de *Fuera de alcance*: ni `sumario`, `vigilar`, `eli`, `buscar-materia` o `materias`, ni grafo, ni `--asunto`, ni versiones históricas, ni paginación, ni revalidación, ni mantenimiento de caché, ni skill, ni `data/normas.yaml`. Las piezas no enumeradas por el spec que el diseño necesita están en *Complexity Tracking* con su motivo. | ✅ Cumple |
| **VI** | Un binario, convenciones de agente | `boe` entra en el binario único y responde igual por `kitlegal boe` y por el enlace `boe` (FR-001, `boe-multicall.txtar`). Las ocho banderas globales se heredan: `--json` y `--timeout` del kernel, `--offline` (solo lectura, 4 si falta), `--dry-run` (describe sin pedir ni escribir), `--describe` (esquema de entrada y salida, origen de `schemas/` y de la futura tabla de comandos de H5), `--no-graph` y `--asunto` sin efecto (FR-128), `--verbose` (eventos de `httpx`, caché y `boe`). | ✅ Cumple |
| **VII** | Grafo y privacidad | No hay grafo (H7), pero los ids naturales que el applet observa quedan en `data` para emitirlos entonces sin reabrir el lector: `norma` (`BOE-A-…`), `url_eli`, `bloque`, `fecha_version` y `hash_texto` (FR-015). No se trata ningún dato personal: la fuente es legislación publicada y la caché vive con los permisos de H3 (`0700`/`0600`). | ✅ Cumple |
| **VIII** | Skills primero; el binario es la herramienta | H4 construye la herramienta determinista que `boe-legislacion` (H5) ejecutará: acceso a la fuente, lectura, caché, ids y huellas; nada de razonamiento ni de presentación en texto plano (la de `boe.py` no se porta). Protege a la skill con contratos: `--describe`, `schemas/` vigilados y verificación nocturna de la fuente. | ✅ Cumple |
| **IX** | Genericidad territorial, validación local | La API es nacional y el applet no filtra por territorio (spec, *Fuera de alcance*): ningún municipio, boletín ni comunidad en código, datos ni fixtures; las normas autonómicas consolidadas se consultan por su identificador sin caso especial. El punto 11 de la Definition of Done no aplica. | ✅ Cumple |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución §IV)

| Regla | Situación en H4 | Cómo se hace cumplir | Veredicto |
|---|---|---|---|
| `internal/core/**` no importa `internal/{source,httpx,cache,store,graph,render,cli,app}` ni entrada y salida | `internal/core/source.go` importa solo `context`, `time` e `internal/core/schema`; `schema` sigue importando solo lo que ya importaba (`time` para la fecha) | `depguard` lista `core` + `compruebaDominioPuro` (sin cambios) | ✅ Cumple |
| Solo `internal/httpx` importa `net/http` | `internal/source/boe` usa `httpx.Peticion`, `httpx.Respuesta` y `httpx.Error`; sus tests usan `httpx.Replay`, nunca `httptest`; `internal/app` y el binario de e2e construyen clientes con `httpx.New`/`httpx.Replay`. `mime` (validación de `Acepta`) vive en `httpx` | `depguard` lista `red` (por prefijo, tests incluidos) + subprueba R2 con dueño obligatorio | ✅ Cumple |
| Solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | `boe` usa el puerto `core.Cache`; `internal/app` importa `internal/cache` (no el controlador) para abrirla; tests de `boe` también por `internal/cache` | `depguard` lista `sql` + subprueba R3 (sin dueño obligatorio hasta `store` y `graph`) | ✅ Cumple |
| Solo `internal/cli` y `cmd/` llaman a `os.Exit` | `app.Arrancar` devuelve el código; lo terminan `cmd/kitlegal/main.go` y el `main` del binario de e2e, los dos únicos sitios con excepción de lint (forma estricta de H1) | `forbidigo` `^os\.Exit$` (sin cambios) | ✅ Cumple |
| Solo `internal/render` escribe en stdout; logs con `slog` a stderr | `boe` y el applet no escriben: devuelven `Resultado`; eventos por el `*slog.Logger` recibido; la tabla y el sobre los presenta el kernel | `forbidigo` `^fmt\.Print…$`, `^os\.Stdout$`, `^os\.Stderr$` (sin cambios) | ✅ Cumple |
| `internal/graph` no importa `internal/source/*` ni `internal/render` | `internal/graph` no existe (H7); H4 no lo crea | — (sin objeto; la regla sigue escrita en el roadmap) | ✅ Cumple |
| Los applets de ejemplo no se enlazan en el binario distribuido (ADR 0010) | El binario de e2e importa además `internal/source/boe` e `internal/httpx`, dentro de su árbol; `cmd/kitlegal` no importa `ejemplo` | `depguard` lista `ejemplo` + `TestElBinarioNoEnlazaLosEjemplos` | ✅ Cumple |
| Ningún adaptador firma en el espacio reservado (ADR 0006, nace en H4) | `boe` firma `boe.legislacion-consolidada` y URLs `https://www.boe.es/` | `TestLasFuentesNoFirmanComoKitlegal` (nuevo) + aserciones de `fuente` y `url` en los tests de los seis verbos | ✅ Cumple |

### Gates (constitución, «Gates»)

- **Capa 1 (mecánica)** que H4 activa por primera vez o toca: validación de toda salida del applet contra `schemas/*.json`
  y `make schema-check` real; corrección de citas contra la respuesta grabada del BOE (FR-116); golden de los seis verbos;
  exit codes 2, 3, 4 y 5 con fixtures (grabados y sintéticos; el 6 no se produce); tests offline contra `testdata/`
  grabado; espacio reservado; superficie del binario ampliada; fuzz de ids; guardián de diff con `[datos]` en las tareas
  que tocan `testdata/` y `schemas/`. Siguen sin objeto: frontmatter y `references/` de skills (H5) y matriz territorial
  (sin dimensión territorial).
- **Capa 2 (jueces)**: además de spec ↔ plan ↔ tasks, la revisión final juzga que el adaptador nuevo respeta las reglas de
  fuentes que el linter no ve (ritmo, identificación, términos).
- **Capa 3 (humano)**, pausas previstas y no defectos: en la pausa del manifiesto, la fila definitiva de
  `docs/SOURCES.md` con la fecha real de la revisión y sus dos constantes, la grabación y las cinco referencias; sintéticos;
  `argumentos.txtar` (dos veces) y los guiones e2e; golden; esquemas; el directorio nuevo `internal/source/boe` y
  `docs/SOURCES.md` (pausa final del workflow); la fusión.

**Reglas del modo desatendido**: el ejecutor nunca usa `KITLEGAL_RECORD` ni la red; la grabación la hace una persona en la
pausa de la tarea `[datos]` del manifiesto con `scripts/grabar-fixtures.sh`; ninguna tarea `[datos]` toca código; ningún
test ni fixture se ajusta al código (las referencias las escribe una persona antes del código y los golden se revisan); el
ejecutor nunca escribe una fecha de revisión de términos ni alinea las constantes de la fuente con `docs/SOURCES.md`:
propone la fila con «Revisado» `pendiente` y la fija una persona.

**Veredicto del gate: PASA.** Ninguna violación; las piezas que el spec no enumera están justificadas en *Complexity
Tracking* y ninguna afecta a alcance, frontera humana, privacidad, términos de uso ni a una decisión cerrada.

### Re-evaluación tras la fase 1 (diseño)

El diseño no añade capacidades sobre lo evaluado. Lo que apareció al bajar a diseño y roza un principio:

- **`httpx` crece** (§I, §IV): `Acepta` e `Instante` son necesarios para FR-003 y FR-096 y no relajan ninguna garantía de
  H2 (la identificación no es sobrescribible, la reproducción empareja igual); vigilados por `TestPedirConAcepta` y
  `TestInstanteDeEmision`.
- **`Procedencia` crece con la fecha** (§II): retrocompatible (cero = comportamiento actual), con ADR 0015 como pide el spec.
- **La grabación la hace una persona en una pausa provocada por un manifiesto** (§III, capa 3): es la única forma de que
  el workflow pause para grabar sin que el ejecutor toque la red; el manifiesto es además la lista cerrada revisable. En
  esa misma pausa, y por este orden, la persona fija la fila de `docs/SOURCES.md` y las dos constantes de
  `internal/source/boe/terminos.go`, graba con el intervalo ya revisado y escribe y revisa las referencias (FR-116, Q2).
- **`TestFuenteCoincideConSources` en dos pasos** (§I, §III): hasta la pausa, la fila es una propuesta con «Revisado»
  `pendiente` y `Revisados` cero, y el test admite exactamente esa pareja para que `make ci` siga en verde sin que el
  ejecutor afirme una revisión que no ha ocurrido; la tarea de código que sigue a la pausa retira esa admisión. Que el
  ejecutor no pueda cambiar después `terminos.go` ni la fila **depende de una regla sobre el texto entero de las líneas
  de tarea** (obligación 3.2), no de que ninguna tarea pretenda declarar esos ficheros: el extractor del workflow toma
  como declarada toda ruta completa que aparezca en cualquier parte de la línea —la de un fichero que la tarea solo lee,
  compara o dice no tocar, y la de un paquete dentro de una orden (`./internal/source/boe/`)— y el guardián admite esa
  ruta y todo lo que cuelga de ella. Desde esa tarea, ninguna línea contiene en ninguna parte la ruta completa de esos
  ficheros ni de un directorio que los contenga (`docs/`, `internal/`, `internal/source/`, `internal/source/boe/`): se
  nombran sin directorio (`SOURCES.md`, `terminos.go`) o por el test que los lee, y los ficheros del paquete que la
  tarea cambia van por su ruta completa, sin llaves ni comodines.
- **Golden y esquemas en tres pasos** (§III): comparador → `[datos]` → cobertura; ningún test queda desactivado: entre el
  primer y el tercer paso el comparador vigila lo que existe.
- **Lectura estricta de tipos inesperados** (§II): lo que `boe.py` imprimiría como `repr` de Python falla con 4 en lugar
  de presentarse como contenido; la verificación nocturna avisará de un cambio de forma.

**Veredicto tras el diseño: PASA**, sin ninguna violación.

## Project Structure

### Documentation (this feature)

```text
specs/005-h4-applet-boe-puerto/
├── plan.md              # Este fichero
├── research.md          # Fase 0: D1-D20 (decisiones con alternativas, verificación en local, supuestos)
├── data-model.md        # Fase 1: data de cada verbo, lecturas, tipo inferido, ids, entradas de caché, flujos, invariantes
├── quickstart.md        # Fase 1: guía de validación ejecutable
├── contracts/
│   ├── puerto-y-applet.md              # core.Source, Procedencia.FechaConsulta, superficie de boe, AppletBoe, Arrancar
│   ├── verbos-y-salidas.md             # los seis verbos: argumentos, peticiones, caché, data, url, ensayo
│   ├── httpx-acepta-e-instante.md      # Peticion.Acepta, Instante, ConHora, orden de la cadena
│   ├── errores-y-codigos.md            # tabla cerrada de 23 situaciones → clase → código → url → fecha
│   └── esquemas-fixtures-y-controles.md # schemas/, golden, manifiesto y grabaciones, sintéticos, referencias, e2e, verify-sources, SOURCES.md, doc.go
├── spec.md
├── checklists/
├── gates/
└── tasks.md             # Fase 2 (/speckit-tasks; no lo crea este comando)
```

### Source Code (repository root)

Estructura de `docs/ROADMAP.md` §2 y `CLAUDE.md`. **NUEVO** = lo crea H4; ← = cambia.

```text
cmd/kitlegal/
├── main.go                      ← app.Arrancar(os.Args, app.RegistroDeProduccion, …)                        (D16)
└── main_test.go                 ← TestPuntoDeEntrada llama a app.Arrancar con app.RegistroDeProduccion, como main()
internal/
├── core/
│   ├── doc.go                   ← «Source nace con H4»
│   ├── source.go                **NUEVO** puerto Source, Consulta, Terminos                                     (D2)
│   └── schema/sobre.go          ← Procedencia.FechaConsulta                                                   (D3)
├── cli/sobre.go                 ← montar usa Procedencia.FechaConsulta cuando no es cero                      (D3)
├── httpx/
│   ├── peticion.go              ← Peticion.Acepta; Respuesta.Instante                                         (D4)
│   ├── errores.go               ← Error.Instante                                                              (D4)
│   ├── cliente.go               ← ConHora; validación de Acepta; Accept en emitir; marca por salto            (D4)
│   ├── instante.go              **NUEVO** marca de emisión en el contexto y conMarcaDeEmision                  (D4)
│   └── robots.go                ← la obtención del robots.txt oculta la marca                                 (D4)
├── source/
│   └── boe/                     **NUEVO** el adaptador
│       ├── doc.go               las 32 anotaciones de FR-120                                                  (FR-120)
│       ├── terminos.go          IntervaloEntrePeticiones y terminosDeUso, atados a docs/SOURCES.md; los fija una persona (D15)
│       ├── fuente.go            NombreDeLaFuente, Fuente, Nueva, opciones, Name/Fetch/TTL/Terms                  (D2, D5)
│       ├── consultas.go         ConsultaBuscar … ConsultaAnalisis                                              (D2)
│       ├── ids.go               ValidarNorma, ValidarBloque, TipoDesdeID                                       (D9)
│       ├── direcciones.go       direcciones de la API y públicas                                               (D8)
│       ├── busqueda.go          consulta de búsqueda y codificación como json.dumps + quote                    (D8)
│       ├── peticiones.go        pedir con Accept e instante; estados → clases                                  (D10)
│       ├── lectura.go           envoltorio JSON, objeto suelto → lista, envoltorios, textos                     (D7)
│       ├── bloque.go            lectura del XML del bloque                                                     (D6)
│       ├── avisos.go            avisos de vigencia, CodigosDeAviso                                             (FR-012)
│       ├── datos.go             tipos de data con etiquetas json/jsonschema, TiposDeBloque                      (D11)
│       ├── entradas.go          claves y contenido de las entradas de caché                                    (D5)
│       ├── errores.go           Error con clase                                                                (D10)
│       ├── buscar.go, indice.go, articulo.go, metadatos.go, analisis.go   un fichero por verbo (articulos en articulo.go)
│       ├── *_test.go            un fichero de test por fichero de código; casos_test.go (casos y ayudas);
│       │                        grabacion_test.go (//go:build grabacion)
│       └── testdata/            **NUEVO**, datos protegidos
│           ├── grabaciones.json                           manifiesto (lista cerrada)                         (D12)
│           ├── boe.legislacion-consolidada/               grabaciones reales                                 (D12)
│           ├── sintetico/<escenario>/boe.legislacion-consolidada/   siete escenarios                         (D12)
│           ├── referencias/                               cinco referencias del diff de aceptación          (FR-116)
│           └── golden/                                    trece golden                                       (FR-112)
├── app/
│   ├── boe.go                   **NUEVO** AppletBoe, DependenciasDeBoe, DependenciasDeRed, argumentos         (D1)
│   ├── registro.go              ← RegistroDeProduccion() (*Registro, error) con boe                           (D16)
│   ├── main.go                  ← Arrancar                                                                    (D16)
│   ├── boe_test.go, esquemas_test.go, fuentes_test.go, fuentes_red_test.go (//go:build fuentes)  **NUEVOS**
│   ├── e2e_test.go              ← Setup (reproducción y KITLEGAL_CACHE_DIR) y orden cronometra                (D13)
│   ├── registro_test.go, main_test.go, ayuda_test.go, despacho_test.go   ← el registro de producción ya no está vacío
│   ├── ejemplo/kitlegal-e2e/main.go   ← Arrancar; registra además boe con httpx.Replay                        (D13)
│   └── testdata/script/boe-{verbos,multicall,offline,codigos,cache-rapida}.txtar   **NUEVOS**; argumentos.txtar ← lista de applets con boe               (D13)
└── arch_test.go                 ← retira dos tests; amplía modulosDelBinario; TestLasFuentesNoFirmanComoKitlegal (D14)
schemas/
├── norma.json                   **NUEVO** buscar, indice, metadatos, analisis                                 (D11)
└── bloque.json                  **NUEVO** articulo, articulos                                                 (D11)
scripts/
├── verify-sources.sh            **NUEVO**                                                                     (D13)
└── grabar-fixtures.sh           **NUEVO**                                                                     (D12)
.github/workflows/nightly.yml    ← trabajo fuentes                                                             (D15)
Makefile                         ← schema-check real; verify-sources                                           (D11, D15)
.golangci.yml                    ← run.build-tags + fuentes, grabacion; misspell.ignore-rules + 6 palabras      (D13, D18)
docs/
├── SOURCES.md                   **NUEVO** fila boe.legislacion-consolidada                                     (D15)
├── ADR/0015-puerto-source-y-fecha-de-consulta.md   **NUEVO**                                                  (D2, D3)
└── PENDIENTES.md                ← se borran las dos entradas de H4                                            (D12, D14)
CHANGELOG.md, README.md, CONTRIBUTING.md   ← applet boe, schema-check activo, verify-sources                    (D17)
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 sin desviaciones: el adaptador en `internal/source/boe`, el puerto en
el paquete raíz del dominio, el applet en la raíz de composición `internal/app`, `schemas/` en la raíz. Los fixtures van
junto al paquete de la fuente, en su `testdata/<fuente>/` ([D12](./research.md#d12--fixtures-dónde-viven-cómo-se-graban-sintéticos-referencias-y-golden)),
que es la decisión que `docs/PENDIENTES.md` dejaba para este hito. No se crean `internal/store`, `internal/graph`,
`pkg/`, `skills/` ni `data/`.

## Controles mecánicos que este hito añade o toca

Según la sección «Gates» de la constitución. «Demostración» dice qué falla si el control se retira o se viola.

| # | Control | Herramienta / forma | Orden | ¿En `make ci`? | Demostración |
|---|---|---|---|---|---|
| 1 | **Salida de los seis verbos contra `schemas/*.json`** | `TestSalidaDeBoeContraSchemas`: sobre real (éxito y fallos 2, 3, 4) validado contra la parte del verbo leída del fichero, `AssertFormat`; `data` con una clave de más o de menos no valida; `Aviso.codigo` con exactamente tres valores | `test` | Sí | Añadir un campo a `boe.Articulo` sin regenerar → falla (FR-111, SC-006) |
| 2 | **`make schema-check`** | `TestEsquemasPublicados` (parte por verbo y fichero entero) + `TestEsquemasCubrenTodosLosVerbos` | `schema-check` | Sí | Editar a mano `schemas/bloque.json` → falla nombrando fichero y verbo (quickstart 7, SC-006) |
| 3 | **Golden de `data`** | `TestGolden` + `TestGoldenCubreTodosLosCasos` (13 casos, 6 verbos) | `test` | Sí | Cambiar un byte de un golden o de la lectura → falla (quickstart 8, FR-112) |
| 4 | **Corrección de citas contra la respuesta grabada** | `TestArticuloCoincideConBoePy` (5 referencias que una persona escribe a mano y revisa en la pausa de grabación, antes del código) | `test` | Sí | No sumar el `tail` o tomar la primera versión → falla el campo `texto` o `fecha_version` (FR-116, SC-001) |
| 5 | **e2e de la entrega** | `TestEntregaDelHito/boe-{verbos,multicall,offline,codigos}` sobre el binario de e2e con reproducción | `test`, `test-e2e` | Sí | Quitar `boe` del binario de e2e o romper el enlace → fallan (FR-114, SC-005) |
| 6 | **< 200 ms en caché** | `TestEntregaDelHito/boe-cache-rapida` con `cronometra` y reproducción vacía | `test` | Sí | Construir el cliente o reabrir la fuente en cada lectura servida → código 1 o más de 200 ms (FR-117, SC-002) |
| 7 | **Códigos 2, 3, 4, 5 con fixtures y su sobre** | `TestCodigosDeSalidaDeBoe` (kernel en proceso, reloj controlado) + `TestClasesDeErrorDeBoe` + `TestPedirClasificaEstados` | `test` | Sí | Clasificar el 404 de un bloque como 4 → falla la fila 6 (SC-008) |
| 8 | **Caché, `--offline`, ensayo y fallos no guardados** | `TestCacheDeLosSeisVerbos`, `TestOfflineDeLosSeisVerbos`, `TestEnsayoDeLosSeisVerbos`, `TestFallosNoSeGuardan`, `TestArticulos` (peticiones contadas) | `test` | Sí | Escribir la entrada de `articulo` tras un fallo de metadatos → `TestFallosNoSeGuardan` falla (SC-003, SC-004, SC-012) |
| 9 | **Fecha de consulta** | `TestFechaDeConsultaDeArticulo`, `TestMontadorFechaDeConsulta`, `TestProcedenciaFechaDeConsultaOpcional`, `TestInstanteDeEmision` | `test` | Sí | Fechar con el reloj del kernel lo servido de la caché → falla (FR-096, SC-003, SC-008) |
| 10 | **Fuzz del analizador de ids de bloque** | `FuzzIDDeBloque` (semillas `a21`, `da3`, `dt1` en `make test`; fuzz ligero a mano) | `test` | Sí (semillas) | Aceptar `%` en la gramática → la propiedad «seguro como segmento» falla (FR-082, SC-007) |
| 11 | **Gramática contra los índices grabados** | `TestGramaticaCubreLosIndicesGrabados` | `test` | Sí | Una grabación con un id fuera de `[A-Za-z0-9]{1,64}` → falla (FR-080) |
| 12 | **Grabaciones y referencias completas** | `TestGrabacionesCompletas` (cada entrada del manifiesto se reproduce) + `TestReferenciasCompletas` (cinco referencias de tres normas, una `BOE-A-2015-10565-a21`, con la forma del contrato y `boe_py` en cada campo) | `test` | Sí | Borrar una grabación o una referencia, o vaciar un `boe_py` → falla nombrando el fichero (FR-113, FR-116, SC-005) |
| 13 | **Porte documentado** | `TestDocAnotaElPorte` (32 entradas de FR-120 con línea, destino y requisito) | `test` | Sí | Quitar una entrada de `doc.go` → falla (SC-010) |
| 14 | **Fuente atada a `docs/SOURCES.md`** | `TestFuenteCoincideConSources` (ritmo, términos, fecha; admite «Revisado» `pendiente` con `Revisados` cero solo hasta la pausa del manifiesto, y la tarea siguiente retira esa admisión) | `test` | Sí | Cambiar `IntervaloEntrePeticiones` sin tocar la fila, o dejar la fila en `pendiente` tras la pausa → falla (FR-122, FR-121) |
| 15 | **Espacio reservado** | `TestLasFuentesNoFirmanComoKitlegal` (AST de `internal/source/**`, no pasa en vacío) | `test` | Sí | Un literal `"kitlegal:applet/boe"` en `boe` → falla (quickstart 13, FR-002, SC-011) |
| 16 | **Superficie del binario** | `TestDependenciasDelBinario` con la lista ampliada y justificada; se retiran `TestElBinarioNoEnlazaHTTPX` y `TestElBinarioNoEnlazaCache` | `test` | Sí | Enlazar un módulo no declarado → falla (FR-124) |
| 17 | **`httpx`: formato e instante** | `TestPedirConAcepta`, `TestConHoraRechazaNula`, `TestInstanteDeEmision` (9 subtests) | `test` | Sí | Poner `Accept` en el `robots.txt` o fechar con la obtención del `robots.txt` → falla |
| 18 | **Registro y arranque** | `TestRegistroDeProduccion` (exactamente `boe`), `TestAppletBoe`, `TestArrancar` | `test` | Sí | Un verbo por omisión en `boe` o un registro inválido que no salga con 1 → falla |
| 19 | **`--no-graph` y `--asunto`** | `TestSinGrafoNiAsuntoNoCambianLaSalida` | `test` | Sí | Cualquier efecto → la salida difiere (FR-128, SC-013) |
| 20 | **Verificación contra la fuente real** | `make verify-sources` (`TestVerificarFuentes`, etiqueta `fuentes`) en el trabajo nocturno `fuentes`, que abre o actualiza una incidencia; control negativo sin red `TestVerificacionDeFuentesDetectaCambios` | `verify-sources` (nocturno); `test` | Solo el control negativo | Con el sintético ilegible, la verificación falla nombrando «boe articulo» (FR-115, SC-009) |
| 21 | **Lint de los ficheros etiquetados** | `run.build-tags: [integration, fuentes, grabacion]` | `lint` | Sí | Un error de tipos en `grabacion_test.go` → `make lint` falla |
| 22 | **R1-R5, formato, `-race`, `govulncheck`, `gosec`, `gitleaks`, `go mod verify`, `tidy -diff`, CodeQL** | los de H0-H3, sin exclusiones nuevas | varias | Sí (salvo CodeQL) | Los mismos de H3 |

### Objetivos del `Makefile`

- `schema-check`: deja de ser un aviso y pasa a `go test -count=1 -run '^TestEsquemasPublicados$$' ./internal/app/`, con
  `check-tools` como prerrequisito. Sigue en `ci`.
- `verify-sources` (**nuevo**): `scripts/verify-sources.sh`; con red; **no** entra en `ci`; lo ejecuta el flujo nocturno.
- `ci`, `test`, `test-integration`, `test-e2e`, `lint`: recetas sin cambios.

### Fixtures y datos protegidos

Lista cerrada en [contracts/esquemas-fixtures-y-controles.md](./contracts/esquemas-fixtures-y-controles.md): manifiesto y
22 grabaciones más el `robots.txt` (§3), 7 escenarios sintéticos (§4), 5 referencias (§5), 13 golden (§2), 2 esquemas
(§1) y 5 guiones (§6). Todo fichero bajo `testdata/` o `schemas/` entra en una tarea `[datos]`: las grabaciones y las
referencias, en la pausa de la del manifiesto; el resto, en el diff de la propia tarea.

### Tests de contrato

| Contrato | Tests |
|---|---|
| [puerto-y-applet](./contracts/puerto-y-applet.md) | `TestMontadorFechaDeConsulta`, `TestProcedenciaFechaDeConsultaOpcional`, `TestNuevaRechazaDependenciasAusentes`, `TestFuenteNombreVigenciasYTerminos`, `TestConsultaDeOtroTipo`, `TestAppletBoe`, `TestRegistroDeProduccion`, `TestArrancar`, `var _ core.Source = (*Fuente)(nil)` |
| [verbos-y-salidas](./contracts/verbos-y-salidas.md) | `TestBuscar`, `TestIndice`, `TestArticulo`, `TestArticulos`, `TestMetadatos`, `TestAnalisis`, `TestDirecciones`, `TestConsultaDeBusqueda`, `TestCodificacionComoPython`, `TestEnsayoDeLosSeisVerbos`, guiones e2e |
| [httpx-acepta-e-instante](./contracts/httpx-acepta-e-instante.md) | `TestPedirConAcepta`, `TestConHoraRechazaNula`, `TestInstanteDeEmision` |
| [errores-y-codigos](./contracts/errores-y-codigos.md) | `TestClasesDeErrorDeBoe`, `TestErrorDeBoeMensajes`, `TestPedirClasificaEstados`, `TestCodigosDeSalidaDeBoe` |
| [esquemas-fixtures-y-controles](./contracts/esquemas-fixtures-y-controles.md) | `TestEsquemasPublicados`, `TestEsquemasCubrenTodosLosVerbos`, `TestSalidaDeBoeContraSchemas`, `TestGolden`, `TestGoldenCubreTodosLosCasos`, `TestGrabacionesCompletas`, `TestReferenciasCompletas`, `TestArticuloCoincideConBoePy`, `TestFuenteCoincideConSources`, `TestDocAnotaElPorte`, `TestLasFuentesNoFirmanComoKitlegal`, `TestVerificacionDeFuentesDetectaCambios` |

## Inventario de tests

Nombres fijados aquí para que `tasks.md` y `quickstart.md` los usen tal cual. Un fichero de test por fichero de código, en
el orden del árbol (`golang-testing`); `t.Parallel()` en todos salvo los que usan `t.Setenv`. Excepciones declaradas:
`internal/core/source.go` no tiene test (sin sentencias); `casos_test.go` reúne los casos cerrados y las ayudas.

| Fichero | Tests | Cubre |
|---|---|---|
| `internal/core/schema/sobre_test.go` | `TestProcedenciaFechaDeConsultaOpcional` | D3 |
| `internal/cli/sobre_test.go` | `TestMontadorFechaDeConsulta` (con fecha → esa, en éxito y en fallo; sin fecha → reloj del montador; procedencia inválida → kernel y reloj) | FR-096, D3 |
| `internal/httpx/cliente_test.go` | `TestPedirConAcepta` (cabecera enviada y grabada; vacía → sin cabecera; inválida → 2 sin petición; no en el `robots.txt`), `TestConHoraRechazaNula` | FR-003, D4 |
| `internal/httpx/instante_test.go` | `TestInstanteDeEmision` (`acierto`, `acierto-tras-reintentos`, `fallo-tras-reintentos`, `denegada-por-robots`, `sin-turno`, `redireccion`, `reproduccion`, `ensayo`, `argumentos`) | FR-096, SC-008, D4 |
| `internal/source/boe/doc_test.go` | `TestDocAnotaElPorte` | FR-120, SC-010 |
| `…/terminos_test.go` (interno) | `TestFuenteCoincideConSources` | FR-121, FR-122 |
| `…/fuente_test.go` | `TestNuevaRechazaDependenciasAusentes`, `TestFuenteNombreVigenciasYTerminos` (`Terms()` devuelve `terminosDeUso`), `TestConsultaDeOtroTipo`, `TestCacheDeLosSeisVerbos`, `TestOfflineDeLosSeisVerbos`, `TestEnsayoDeLosSeisVerbos`, `TestFallosNoSeGuardan`, `TestGolden` (bandera `-actualizar-golden`), `TestGoldenCubreTodosLosCasos` | FR-090-FR-096, FR-112, FR-121, FR-122, SC-003, SC-004, SC-012 |
| `…/consultas_test.go` | `TestVerbosDeLasConsultas` | D2 |
| `…/ids_test.go` | `TestValidarNorma`, `TestValidarBloque`, `TestTipoDesdeID` (incluye `a21`, `da3`, `dt1`, `ti`, `cv3`), `FuzzIDDeBloque`, `TestGramaticaCubreLosIndicesGrabados` | FR-040, FR-080-FR-082, SC-007 |
| `…/direcciones_test.go` | `TestDirecciones` | FR-002, FR-003 |
| `…/busqueda_test.go` | `TestConsultaDeBusqueda` (operadores, varias palabras —`procedimiento administrativo común`, con la dirección de [verbos-y-salidas §1](./contracts/verbos-y-salidas.md) byte a byte—, una palabra recortada, espacio en blanco de Python, vacía → 2), `TestCodificacionComoPython` (ASCII, no ASCII, fuera del plano básico, comillas, barra invertida, controles, `/`) | FR-030, D8 |
| `…/peticiones_test.go` | `TestPedirClasificaEstados` (200, 404 por recurso, 404 en búsqueda, 400, 403, ensayo, error de `httpx`, instante) | FR-100, FR-101, D10 |
| `…/lectura_test.go` | `TestLeerEnvoltorio`, `TestComoLista`, `TestTextoDe` | FR-016, FR-070, D7 |
| `…/bloque_test.go` | `TestLeerBloque` (con versiones, sin versiones, última sin fecha, fecha vacía presente, `tail`, CRLF, CDATA, comentario, entidades, atributos con saltos, bloque anidado, raíz `bloque`, ilegible, no UTF-8, declaración ISO-8859-1 con UTF-8, dos raíces) | FR-010, FR-011, FR-014, D6 |
| `…/avisos_test.go` | `TestAvisosDe` (las 8 combinaciones; código numérico → sin aviso), `TestCodigosDeAviso` | FR-012, FR-050 |
| `…/datos_test.go` | `TestEnumeradosDeLosDatos`, `TestHashTexto` | FR-012, FR-015, SC-013 |
| `…/entradas_test.go` | `TestClaveDeEntrada`, `TestEntradaIdaYVuelta` (fecha y url byte a byte), `TestEntradaIlegibleEsInesperado` | FR-090, FR-096, D5 |
| `…/errores_test.go` | `TestClasesDeErrorDeBoe` (una subprueba por fila 2-4, 6-10 y 16-22 de [errores-y-codigos](./contracts/errores-y-codigos.md); la 21 con una `AperturaDeCache` de prueba: apertura, `Get`, `Put` tras obtener la respuesta y `Close` tras una invocación correcta y tras una fallida), `TestErrorDeBoeMensajes` | FR-100, SC-008 |
| `…/buscar_test.go` | `TestBuscar` (`con-resultados`, `sin-resultados-guardada`, `texto-vacio`) | US3, FR-030-FR-032 |
| `…/indice_test.go` | `TestIndice` (`lpac`, `inexistente`, `fuente-caida`) | US4-1, US4-4, FR-040, FR-041 |
| `…/articulo_test.go` | `TestArticulo` (`bloque-vigente`, `derogada`, `bloque-inexistente`, `metadatos-en-cache`, `metadatos-caidos`, `metadatos-ilegibles`, `bloque-ilegible`, `bloque-sin-elemento`, `tres-avisos`), `TestArticuloCoincideConBoePy` (5), `TestArticulos` (`tres-bloques`, `id-repetido`, `segundo-inexistente`, `mezcla-de-cache`), `TestFechaDeConsultaDeArticulo` | US1, US4-2, US4-3, FR-010-FR-021, FR-096, FR-116, SC-001, SC-003, SC-012 |
| `…/metadatos_test.go` | `TestMetadatos` (`vigente`, `derogada`, `inexistente`, `limite`, `tres-avisos`) | US5-1, US5-3, FR-050, FR-051 |
| `…/analisis_test.go` | `TestAnalisis` (`lpac`, `inexistente`, `texto-completo`) | US5, FR-060, FR-061, FR-070 |
| `…/casos_test.go` | casos cerrados de golden, manifiesto y ayudas; `TestGrabacionesCompletas`, `TestReferenciasCompletas` | FR-113, FR-116 |
| `…/grabacion_test.go` (`//go:build grabacion`) | `TestGrabarFixtures` (exige `KITLEGAL_RECORD=1`) | FR-113 |
| `internal/app/boe_test.go` | `TestAppletBoe`, `TestDependenciasDeRed` (sin red), `TestSalidaDeBoeContraSchemas`, `TestCodigosDeSalidaDeBoe` (filas 1-6, 11, 14, 18, 19 y 21; la 21 con `cache.ConDirectorio` sobre un fichero regular), `TestSinGrafoNiAsuntoNoCambianLaSalida` | FR-001, FR-101, FR-111, FR-128, SC-006, SC-008, SC-013 |
| `internal/app/registro_test.go` | `TestRegistroDeProduccion` (← exactamente `boe`) | FR-001 |
| `internal/app/main_test.go` | `TestArrancar` (`registro-valido`, `registro-invalido`) | D16 |
| `internal/app/esquemas_test.go` | `TestEsquemasPublicados` (bandera `-actualizar-esquemas`), `TestEsquemasCubrenTodosLosVerbos` | FR-110, SC-006 |
| `internal/app/fuentes_test.go` | `TestVerificacionDeFuentesDetectaCambios` | FR-115, SC-009 |
| `internal/app/fuentes_red_test.go` (`//go:build fuentes`) | `TestVerificarFuentes` | FR-115 |
| `internal/app/e2e_test.go` | `TestEntregaDelHito` + guiones `boe-verbos`, `boe-multicall`, `boe-offline`, `boe-codigos`, `boe-cache-rapida`; `argumentos` (← lista de applets con `boe`) | FR-114, FR-117, SC-002, SC-005 |
| `internal/app/{ayuda,despacho}_test.go` | ← solo `TestDespachoConRegistroVacio` y la subprueba «un registro vacío lo dice en lugar de enumerar la nada», cuyo sujeto **es** el registro vacío, pasan de `RegistroDeProduccion()` a `&Registro{}`, con sus comentarios de H1 reescritos sobre ese sujeto | FR-001 |
| `cmd/kitlegal/main_test.go` | ← `TestPuntoDeEntrada` sigue ejerciendo **la misma composición que `main()`**: cada caso llama a `app.Arrancar(caso.argv, app.RegistroDeProduccion, …)`, igual que `main()`; la constante `registroVacio` se sustituye por la lista de applets del binario distribuido (`applets disponibles: boe`) en los casos «sin applet», «inventado» y «echo no es del binario distribuido»; gana «`boe` sin verbo → 2»; el comentario de `TestVersionDeLaIdentificacion` deja de afirmar que el binario distribuido no enlaza `internal/httpx` | FR-001, D16 |
| `internal/arch_test.go` | `TestLasFuentesNoFirmanComoKitlegal`; `TestDependenciasDelBinario` (lista ampliada); retirados `TestElBinarioNoEnlazaHTTPX` y `TestElBinarioNoEnlazaCache` | FR-002, FR-124, SC-011 |

## Orden de implementación

`core → adaptador → applet → controles` (constitución, flujo 2), con el test antes que el código en cada tarea, `make ci`
en verde al cerrar cada una y las pausas humanas donde la constitución las pone.

1. **Dominio y kernel**: `schema.Procedencia.FechaConsulta`, `cli.Montador`, `core.Source`, `internal/core/doc.go`, ADR 0015.
2. **`httpx`**: `Peticion.Acepta`, `Instante`, `ConHora`, `instante.go`, ocultación en `robots.go`.
3. **Porte documentado y piezas puras de `boe`**: `doc.go` con las 32 anotaciones **antes** de portar (FR-120) y
   `TestDocAnotaElPorte`; `ids.go` (+ fuzz), `direcciones.go`, `busqueda.go` (`TestConsultaDeBusqueda` fija la dirección de
   `procedimiento administrativo común` byte a byte), `consultas.go`; `.golangci.yml` con `administrativo`, `capitulo`,
   `disposicion`, `materias` y `regulares`, que estas piezas escriben sueltas (D18).
4. **Arnés de grabación** (código): `grabacion_test.go`, `scripts/grabar-fixtures.sh`; `docs/SOURCES.md` como
   **propuesta** («Revisado» `pendiente`; licencia, términos y `robots.txt` por revisar; ritmo propuesto `1s`, S6);
   `terminos.go` con `IntervaloEntrePeticiones` y `terminosDeUso` iguales a la propuesta (`Revisados` cero) y
   `terminos_test.go` con `TestFuenteCoincideConSources`, que admite la pareja `pendiente`/cero; `.golangci.yml`
   (`grabacion`). El ejecutor no escribe ninguna fecha de revisión.
5. **`[datos]` Manifiesto** `internal/source/boe/testdata/grabaciones.json` → **pausa humana**, en este orden:
   1. revisa los términos de uso y el `robots.txt` del BOE; si prohíben el acceso automatizado, rechaza y el hito se
      detiene (FR-123);
   2. deja la fila **definitiva** de `docs/SOURCES.md` (licencia, términos, `robots.txt`, ritmo, formato) con la fecha
      real de la revisión en «Revisado»;
   3. alinea con ella `IntervaloEntrePeticiones` y `terminosDeUso` en `internal/source/boe/terminos.go`;
   4. comprueba `go test -count=1 -run '^TestFuenteCoincideConSources$' ./internal/source/boe/` en verde;
   5. graba con `scripts/grabar-fixtures.sh`, que ya usa el intervalo revisado;
   6. comprueba S1, S2, S4 y S7 sobre lo grabado (y aplica los suplentes en el manifiesto si hace falta);
   7. escribe a mano y revisa las cinco referencias de FR-116 sobre esas grabaciones
      ([contracts §5](./contracts/esquemas-fixtures-y-controles.md#5-referencias-del-diff-de-aceptación-fr-116));
   8. confirma en la rama la fila, `terminos.go`, las grabaciones, las referencias y, si cambió, el manifiesto, y aprueba.

   Que ninguna tarea posterior cambie lo que la persona fija aquí —la fila, `terminos.go`, el manifiesto, las grabaciones
   y las referencias— **depende de la regla de la obligación 3.2 sobre el texto entero de las líneas de tarea**, no de
   que ninguna tarea pretenda declarar esos ficheros: el extractor de rutas toma como declarada toda ruta completa que
   aparezca en cualquier parte de la línea, también la de un fichero que la tarea solo lee, compara o dice no tocar y la
   de un paquete dentro de una orden (`workflow.yml` 583 y 606; `docs/WORKFLOW.md` 201); el guardián admite esa ruta y
   todo lo que cuelga de ella (`docs/WORKFLOW.md` 71; `workflow.yml` 695-698), y la etiqueta `[datos]` basta para tocar
   `testdata/` (704). Desde el paso 6, ninguna línea de tarea contiene, en ninguna parte de su texto ni dentro de una
   orden, la ruta completa de esos ficheros o directorios (`docs/SOURCES.md`, `internal/source/boe/terminos.go`,
   `internal/source/boe/testdata/grabaciones.json`, `internal/source/boe/testdata/boe.legislacion-consolidada/`,
   `internal/source/boe/testdata/referencias/`), de ninguna ruta bajo ellos ni de un directorio que los contenga
   (`docs/`, `internal/`, `internal/source/`, `internal/source/boe/`, `internal/source/boe/testdata/`): se nombran sin
   directorio (`SOURCES.md`, `terminos.go`, `grabaciones.json`, «las grabaciones», «las referencias») o por el test que
   los lee. Cada fichero que una tarea cambia va por su ruta completa y cada subdirectorio de datos también
   (`internal/source/boe/testdata/sintetico/`, `internal/source/boe/testdata/golden/`). La línea de esta tarea `[datos]`
   lleva, de todo ello, solo la ruta del manifiesto y cita el procedimiento de la pausa por su sección
   ([contracts §3.2](./contracts/esquemas-fixtures-y-controles.md#32-grabación)), sin copiar sus rutas ni sus órdenes: si
   no, el ejecutor de esta tarea `[datos]` podría cambiar la fila y `terminos.go`, fuera de `testdata/`, contra las
   obligaciones 1 y 2.
6. **Grabaciones, referencias y fila revisadas** (código): `TestGrabacionesCompletas`, `TestReferenciasCompletas`,
   `TestGramaticaCubreLosIndicesGrabados`; `TestFuenteCoincideConSources` deja de admitir `pendiente`. Su línea lleva la
   ruta completa de lo que cambia (`internal/source/boe/casos_test.go`, `internal/source/boe/ids_test.go`,
   `internal/source/boe/terminos_test.go`) y nombra lo que esos tests leen sin directorio (`SOURCES.md`, `terminos.go`,
   `grabaciones.json`, «las grabaciones», «las referencias»): es la tarea cuyo cometido natural es describir esas lecturas
   (obligación 3.2).
7. **`[datos]` Sintéticos**: los siete escenarios, copiados de las grabaciones → pausa humana. La línea lleva
   `internal/source/boe/testdata/sintetico/` y nombra las grabaciones de origen sin su ruta.
8. **Lecturas** (código): `lectura.go`, `bloque.go`, `avisos.go`, `datos.go`, `errores.go`, `entradas.go`, `peticiones.go`.
9. **Fuente y verbos** (código), uno por tarea: `fuente.go` (`Nueva`, `Fetch`, `TTL`, `Terms`; `.golangci.yml` con
   `dependencias`, D18), `metadatos` → `articulo` (con `TestArticuloCoincideConBoePy` contra las referencias) → `articulos`
   → `indice` → `buscar` → `analisis`; y los tests transversales de caché, `--offline`, ensayo y fallos. En la tarea de
   `fuente.go`, el comparador `TestGolden` con su bandera.
10. **`[datos]` Guion de argumentos tolerante**: `internal/app/testdata/script/argumentos.txtar`, líneas 24 y 30, pasan a
    `stderr 'applets disponibles: (boe, )?contar, echo'`, que vale antes y después de registrar `boe` en el binario de e2e
    → pausa humana (modifica material existente).
11. **Applet y registro** (código): `internal/app/boe.go`, `RegistroDeProduccion`, `Arrancar`, `cmd/kitlegal/main.go`,
    binario de e2e, tests existentes que cambian, `internal/arch_test.go` (retirada, lista de módulos medida,
    `TestLasFuentesNoFirmanComoKitlegal`), `TestAppletBoe`, `TestCodigosDeSalidaDeBoe`,
    `TestSinGrafoNiAsuntoNoCambianLaSalida`; `e2e_test.go` (`Setup`, `cronometra`); el comparador
    `TestEsquemasPublicados` con su bandera y la receta de `make schema-check`. La receta y la copia de las grabaciones
    que hace `Setup` se describen sin sus rutas de paquete (obligación 3).
12. **`[datos]` Guiones e2e**: los cinco `boe-*.txtar` y `argumentos.txtar` ajustado a
    `stderr 'applets disponibles: boe, contar, echo'` → pausa humana (modifica material existente). Va **justo detrás** del
    paso 11: es la primera tarea que puede dejar en verde el e2e de la entrega (workflow, paso `tasks`), porque los guiones
    solo necesitan `boe` registrado en el binario de e2e, `Setup` y `cronometra`, y ninguno lee `schemas/` ni los golden.
    La línea lleva cada guion por su ruta completa y nombra las grabaciones que copia `Setup` sin su ruta (obligación 3.2).
13. **`[datos]` Golden**: `TestGolden` con `-actualizar-golden` → pausa humana. La línea lleva
    `internal/source/boe/testdata/golden/` y cita la orden por su sección
    ([contracts §2](./contracts/esquemas-fixtures-y-controles.md#2-golden-de-data)) sin copiarla: su `./internal/source/boe/`
    declararía el paquete entero, con `terminos.go`, las grabaciones y las referencias (obligación 3.2).
14. **Cobertura de golden** (código): `TestGoldenCubreTodosLosCasos`.
15. **`[datos]` Esquemas**: `TestEsquemasPublicados` con `-actualizar-esquemas` → pausa humana. La línea lleva
    `schemas/norma.json` y `schemas/bloque.json` y cita la orden por su sección
    ([contracts §1](./contracts/esquemas-fixtures-y-controles.md#1-schemasnormajson-y-schemasbloquejson)) sin copiarla:
    su `./internal/app/` declararía `internal/app/` entero, código y guiones incluidos, en una tarea `[datos]`
    (obligaciones 1 y 3.1).
16. **Contratos** (código): `TestEsquemasCubrenTodosLosVerbos`, `TestSalidaDeBoeContraSchemas`.
17. **Verificación nocturna** (código): `fuentes_test.go`, `fuentes_red_test.go`, `scripts/verify-sources.sh`,
    `Makefile` (`verify-sources`), `.github/workflows/nightly.yml`, `.golangci.yml` (`fuentes`).
18. **Documentación**: `CHANGELOG.md`, `README.md`, `CONTRIBUTING.md`, `docs/PENDIENTES.md`; validación del quickstart,
    que la línea cita por su nombre sin copiar sus órdenes (llevan `./internal/source/boe/` y `./internal/app/`); la guía
    parte de una carpeta temporal vacía, de modo que un reintento o una reanudación empieza de nuevo por los prerrequisitos
    sin que `git clone` encuentre el clon del intento anterior. Cada orden de la guía está escrita en una forma que la
    sesión desatendida (`claude -p --permission-mode acceptEdits`, lista de permitidos de `.claude/settings.json`) ejecuta
    sin pedir aprobación, y la tarea la ejecuta tal cual, sin sustituciones. Las formas son estas: `rtk proxy` delante de
    lo que la lista no cubre (`sh`, `env`, `test`, `perl`, `shasum`) y de lo que ejecuta, crea, borra o edita en la
    carpeta temporal; el código de salida impreso dentro de `rtk proxy sh -c '…; echo "código $?"'`; y nada redirigido a
    ficheros. Las sondas positivas de los prerrequisitos las comprueban en cada ejecución.
19. **`[plataforma]`**: empujar la rama y, si `gh pr view` no encuentra propuesta de cambio (la tarea puede reintentarse),
    abrirla con título explícito (`--title`, obligatorio sin terminal) y el cuerpo de `gates/pr-h4.md` (módulos del binario
    justificados, supuestos pendientes S8-S9); esperar CI y Codecov. Al final, pausa del workflow por `docs/SOURCES.md` y
    por el directorio nuevo `internal/source/boe`.

## Complexity Tracking

> Piezas que el spec no enumera y el diseño necesita, y divergencias conscientes. **Ninguna dependencia fuera de la lista
> de la constitución §V**: H4 no añade ninguna.

| Divergencia | Por qué es necesaria | Alternativa más simple, y por qué se rechaza |
|---|---|---|
| **`httpx.Peticion.Acepta`** (H2 fijó `Peticion` con dos campos) | FR-003 exige pedir XML o JSON como `boe.py`, que lo hace por `Accept` ([D4](./research.md#d4--internalhttpx-gana-tres-cosas-peticionacepta-instante-y-conhora)) | *Cabeceras libres*: permitirían tocar la identificación. *Formato por la dirección*: no consta que la API lo admita (S1) |
| **`httpx.Respuesta.Instante`, `Error.Instante` y `ConHora`**, con una marca en el contexto | FR-096 y SC-008 definen el instante como el de emisión del último intento o el de abandono, que solo se conoce dentro de la cadena (D4) | *Medir en el adaptador*: es el de recepción, no el de emisión |
| **`schema.Procedencia.FechaConsulta`** y ADR 0015 | FR-096; el spec prevé ampliar `Procedencia` o `Resultado` con ADR (D3) | *Fecha en `data`*: rompe la huella (ADR 0006). *Reloj del kernel*: fecha mal lo servido de la caché |
| **Firma concreta de `core.Source`** (`Fetch(ctx, ec, consulta)`, `TTL(consulta)`) y ADR 0015 | `CLAUDE.md` nombra el puerto sin fijar parámetros; `--offline`/`--dry-run` y la vigencia por verbo (FR-091) los exigen (D2) | *`Fetch(ctx, req)` con el contexto dentro de `req`*: mezcla consulta e invocación. *`TTL()`*: una sola vigencia |
| **`app.Arrancar` y `RegistroDeProduccion` con error** | Registrar el primer applet hace falible el registro; §IV no admite `panic` (D16) | *`panic` como el e2e de H1*: evitable. *Cambiar `Main`*: rompe sus llamadas y tests |
| **El binario de e2e registra `boe` con reproducción** | FR-114 pide el e2e sobre un binario compilado y sin red; el distribuido no puede aceptar una reproducción sin permitir falsificar citas (D13) | *Variable de reproducción en el binario distribuido*: falsificación posible. *Segundo binario de e2e*: otra raíz de composición con su excepción de lint y otra construcción por ejecución. Coste asumido: `argumentos.txtar` (H1) enumera ahora `boe`, en dos pasos `[datos]` |
| **Manifiesto `testdata/grabaciones.json`; en su pausa, una persona fija la fila de `docs/SOURCES.md` y `terminos.go`, graba y escribe las referencias** | El ejecutor no puede usar la red; el workflow solo pausa si una tarea `[datos]` añade ficheros; FR-116 y Q2 piden revisar las referencias en la tarea `[datos]` que graba, y la grabación debe usar el intervalo ya revisado (D12, D15) | *Tarea `[datos]` «grabar»*: no cambia nada y no pausa. *Job de Actions*: exige la plataforma en mitad del hito. *Referencias del ejecutor en una tarea `[datos]` posterior*: contra la letra de FR-116 y con la misma mano en la referencia y en el código comparado |
| **`TestFuenteCoincideConSources` admite «Revisado» `pendiente` hasta la pausa del manifiesto** | Cada tarea deja `make ci` en verde y el ejecutor no puede afirmar una revisión de términos que no ha ocurrido; la tarea que sigue a la pausa retira la admisión (D15) | *Fecha propuesta por el ejecutor*: afirma una revisión falsa. *Test solo tras la pausa*: la persona no podría comprobar en la pausa que la fila y las constantes casan antes de grabar |
| **Golden y esquemas en tres pasos** (comparador → `[datos]` → cobertura) | Una tarea `[datos]` solo toca `testdata/` y `schemas/`, y cada tarea deja `make ci` en verde (D12) | *Test estricto antes de los datos*: `make ci` en rojo. *Golden a mano*: cientos de bloques de un índice sin verdad independiente |
| **`make verify-sources`** (no está en la lista de H0) | El `Makefile` es la única superficie de invocación (`CLAUDE.md`); el nocturno llama a `make` como CI (D15) | *Llamar al script desde el flujo*: segundo camino y otro `GOTOOLCHAIN` |
| **Etiquetas `fuentes` y `grabacion` en `run.build-tags`** | Sin ellas, dos ficheros de test quedan fuera del lint (D13) | *Etiqueta `integration`*: `make test-integration` tocaría la red |
| **`administrativo`, `capitulo`, `dependencias`, `disposicion`, `materias`, `regulares` en `misspell.ignore-rules`** | Valores de la fuente y del contrato (`capitulo`, `disposicion`, `materias`), texto de una consulta de la lista cerrada (`administrativo`), nombres del contrato del applet (`dependencias`) y contenido obligatorio de FR-120 (`regulares`) que el diccionario inglés toma por erratas (D18) | *Renombrar o reescribir*: cambiaría el contrato de `data`, los valores de la fuente, la consulta grabada o el texto que FR-120 exige |
| **Tipos inesperados → 4 en vez del `repr` de `boe.py`** | Un campo de texto tipado no puede llevar un `repr` de Python y §II impide presentarlo como contenido (D7) | *Literal JSON*: tampoco es lo que imprime `boe.py` y oculta un cambio de forma |
| **`--dry-run` abre la caché en solo lectura** | Describir solo las peticiones que se emitirían sin crear la base (D5) | *No mirar la caché*: describe peticiones que no ocurrirían. *Modo normal*: crea `cache.db` |
| **Fixtures junto al paquete** (`internal/source/boe/testdata/<fuente>/`) | Convención de Go, recomendación de `docs/PENDIENTES.md`; `clasificar_datos` pausa igual por ficheros nuevos bajo `internal/source/*` (D12) | *`testdata/boe/` en la raíz*: rutas con `../../..` sin ganar nada |
| **Ampliación de `modulosDelBinario`** con módulos transitivos del controlador | FR-124: el binario enlaza la caché; cada módulo justificado (D14) | *Copiar `go.mod`*: declararía módulos que no se enlazan |

## Obligaciones que este plan traslada a `tasks.md`

1. **Orden y pausas** del apartado anterior: el manifiesto es la tarea `[datos]` en cuya pausa la persona, por este orden,
   fija la fila de `docs/SOURCES.md` y `terminos.go`, comprueba `TestFuenteCoincideConSources`, graba y escribe las
   referencias; sintéticos antes del código de lectura; golden y esquemas en tres pasos; `argumentos.txtar` se relaja en
   una tarea `[datos]` antes de registrar `boe` en el binario de e2e; los guiones e2e, con el ajuste final de
   `argumentos.txtar` a `boe, contar, echo`, en la tarea `[datos]` que sigue **inmediatamente** a la de registro, antes de
   golden, esquemas y contratos. Ninguna tarea `[datos]` toca código ni la `.golangci.yml`.
2. **El ejecutor nunca graba ni consulta la red, ni escribe lo que fija la persona**: ninguna tarea ejecuta
   `KITLEGAL_RECORD`, `scripts/grabar-fixtures.sh` ni `make verify-sources`; ninguna escribe una fecha de revisión, alinea
   `terminos.go` con `docs/SOURCES.md` ni crea o cambia una referencia. La tarea de código tras la pausa del manifiesto
   empieza por `TestGrabacionesCompletas`, `TestReferenciasCompletas` y `TestFuenteCoincideConSources` sin la admisión de
   `pendiente`; si falta una grabación o una referencia, si la fila sigue en `pendiente`, o si S2 no se cumple (inexistente
   con HTTP distinto de 404), se detiene y lo anota en `gates/tarea-Tnnn.md`: lo decide una persona (en S2, conflicto
   con el spec entre FR-014 y US1-5).
3. **Rutas declaradas** con precisión (guardián de diff). El extractor de rutas (`.specify/workflows/hito/workflow.yml`
   606) no lee una lista de rutas: aplica su expresión a `texto`, **la línea entera de la tarea** sin la casilla (583;
   `docs/WORKFLOW.md` 201), y toma como declarada **cada ruta que aparezca en cualquier parte de ese texto**: la de un
   fichero que la tarea solo lee, compara o dice no tocar, y la de un paquete dentro de una orden
   (`go test … ./internal/source/boe/` declara `internal/source/boe/`). El guardián admite una ruta declarada y todo lo
   que cuelga de ella (`docs/WORKFLOW.md` 71; 695-698, `case "$f" in "$r"|"$r"/*`), rechaza lo que queda fuera con o sin
   `[datos]` (705) y, con `[datos]`, deja tocar `testdata/` y `schemas/` (704). Compara la cadena entera desde la raíz,
   así que un nombre sin directorio (`terminos.go`) o una ruta parcial (`boe/terminos.go`) no casa con ningún fichero del
   paquete, tampoco por la regla de `x_test.go` (699). Además, el extractor deja en su directorio lo que no es un nombre
   completo: de `internal/httpx/{peticion,errores}.go` saca `internal/httpx/`, `internal/source/boe/*_test.go` queda en
   `internal/source/boe` al recortar el comodín (696), y un directorio con punto en el nombre sale sin su barra final
   (`internal/source/boe/testdata/boe.legislacion-consolidada/` → `…/boe.legislacion-consolidada`), que admite lo mismo.
   Todo ello está comprobado en local con la misma tubería y los mismos `case` sobre líneas de ejemplo (research.md D15).
   Por eso:
   1. **Cada fichero se declara por su ruta completa, sin llaves ni comodines, y una línea solo contiene la ruta completa
      de lo que la tarea puede cambiar.** Lo que solo lee, compara o toma como modelo se nombra sin directorio o por su
      test o símbolo, y una orden con ruta de paquete (`go test … ./internal/app/`) se cita por el test y la bandera, con
      la orden completa en el contrato o en este plan. Los de H1-H3 que se tocan son exactamente
      `internal/core/doc.go`, `internal/core/schema/sobre.go`, `internal/cli/sobre.go`, `internal/httpx/peticion.go`,
      `internal/httpx/errores.go`, `internal/httpx/cliente.go`, `internal/httpx/robots.go`, `internal/app/registro.go`,
      `internal/app/main.go`, `internal/app/e2e_test.go`, `internal/app/registro_test.go`, `internal/app/main_test.go`,
      `internal/app/ayuda_test.go`, `internal/app/despacho_test.go`, `internal/app/ejemplo/kitlegal-e2e/main.go`,
      `cmd/kitlegal/main.go`, `cmd/kitlegal/main_test.go`, `internal/arch_test.go`, `Makefile`, `.golangci.yml`,
      `.github/workflows/nightly.yml`, `docs/PENDIENTES.md`, `README.md`, `CONTRIBUTING.md`; y, solo en tareas `[datos]`,
      `internal/app/testdata/script/argumentos.txtar`. Los cinco guiones nuevos van cada uno por su nombre
      (`internal/app/testdata/script/boe-verbos.txtar`…): `boe-{verbos,…}.txtar` se quedaría en un prefijo que no admite
      ninguno.
   2. **Lo que fija la persona en la pausa del manifiesto** —la fila de `docs/SOURCES.md`, `internal/source/boe/terminos.go`,
      el manifiesto `internal/source/boe/testdata/grabaciones.json`, las grabaciones de
      `internal/source/boe/testdata/boe.legislacion-consolidada/` y las referencias de
      `internal/source/boe/testdata/referencias/`— solo aparece por su ruta completa en dos líneas: la de la tarea del
      paso 4 (`docs/SOURCES.md` e `internal/source/boe/terminos.go`) y la de la `[datos]` del paso 5, que de todo ello
      lleva solo la ruta del manifiesto y cita el procedimiento de la pausa por su sección del contrato, sin copiar sus
      rutas ni sus órdenes. **Desde el paso 6, ninguna línea de tarea contiene, en ninguna parte de su texto**, la ruta
      completa de `docs/SOURCES.md`, `internal/source/boe/terminos.go` o `internal/source/boe/testdata/grabaciones.json`,
      ni de `internal/source/boe/testdata/boe.legislacion-consolidada/` o `internal/source/boe/testdata/referencias/`, ni
      ninguna ruta bajo ellos, ni un directorio que los contenga (`docs/`, `internal/`, `internal/source/`,
      `internal/source/boe/`, `internal/source/boe/testdata/`), tampoco dentro de una orden (`./internal/source/boe/`).
      Para referirse a ellos se usa el nombre sin directorio (`SOURCES.md`, `terminos.go`, `grabaciones.json`, «las
      grabaciones», «las referencias») o el nombre del test que los lee (`TestFuenteCoincideConSources`,
      `TestGrabacionesCompletas`, `TestReferenciasCompletas`), que el guardián no puede casar con la ruta completa
      (698: `"$r"|"$r"/*`). Cada fichero del paquete que la tarea cambia va por su ruta completa
      (`internal/source/boe/fuente.go`, `internal/source/boe/casos_test.go`…) y cada subdirectorio de datos también
      (`internal/source/boe/testdata/sintetico/`, `internal/source/boe/testdata/golden/`). La mención es natural, y hay
      que evitarla, en la tarea del paso 6 (sus tests leen la fila, el manifiesto y las referencias), en la de sintéticos
      (copia de las grabaciones), en la de `articulo` (compara con las referencias), en las de registro y guiones e2e
      (`Setup` copia las grabaciones), en las `[datos]` de golden y esquemas (su orden lleva la ruta del paquete) y en la
      de documentación (las órdenes del quickstart). La protección depende de esta regla: con ella, ninguna ruta extraída
      casa con lo fijado y el guardián rechaza cualquier cambio en ello, lleve o no la tarea `[datos]` (705). Sin ella,
      una tarea de código cuya línea dijera «lee `docs/SOURCES.md`» o «no toca `internal/source/boe/terminos.go`» podría
      cambiarlos: están fuera de `testdata/`, así que no los frena la regla de `[datos]` (704) ni los pausa
      `clasificar_datos` (882-900), y solo quedaría la pausa final (`rutas_sensibles`, 1134), donde `docs/SOURCES.md`
      aparece entero como fichero nuevo del hito, sin distinguir lo que fijó la persona de un cambio posterior. Y una
      `[datos]` cuya orden llevara `./internal/source/boe/` podría cambiar las referencias: `clasificar_datos` pausaría esa
      tarea por modificar material existente, pero es una revisión humana del diff de otra tarea, no la protección.
   3. **Se comprueba en `tasks.md` sobre el texto entero de cada línea**, no solo sobre las rutas que la tarea pretende
      declarar: a cada línea `- [ ] Tnnn` se le aplica la tubería de 606 (`tr`, `grep -oE` con su expresión,
      `grep -vE '^https?:'`) y cada ruta extraída, recortada como en 696 (sin comodín ni barra final), se compara con la
      lista. Ninguna línea lleva `{` ni `*` dentro de una ruta; desde la tarea que sigue al manifiesto, ninguna ruta
      extraída es `docs/SOURCES.md`, `internal/source/boe/terminos.go`, `internal/source/boe/testdata/grabaciones.json`,
      `internal/source/boe/testdata/boe.legislacion-consolidada` o `internal/source/boe/testdata/referencias`, ni una ruta
      bajo alguno de ellos, ni un prefijo de directorio de alguno; y en la línea del manifiesto, la única de ellas es la
      del manifiesto.
4. **`misspell`** (D18; lista del barrido de `words.go` y `words_us.go` descrito, con su método, en la tabla de
   verificación de research.md): la tarea del paso 3 declara `.golangci.yml` y añade, cada una con su comentario,
   `administrativo`, `capitulo`, `disposicion`, `materias` y `regulares`; la de `fuente.go` (paso 9) añade `dependencias`.
   `misspell` revisa los `.go` enteros —comentarios, cadenas, etiquetas e identificadores que no mezclan mayúsculas—, así
   que nunca se escriben sueltas sin tilde `adaptacion`, `compilacion`, `composicion`, `configuracion`, `conjuncion`,
   `constitucion`, `construccion`, `convencion`, `correccion`, `corrupcion`, `declaracion`, `distribucion`,
   `documentacion`, `emision`, `evaluacion`, `historicas`, `identificacion`, `implementacion`, `informacion`,
   `justificacion`, `omision`, `orientacion`, `parametros`, `presentacion`, `produccion`, `prohibicion`, `proposito`,
   `proteccion`, `regeneracion`, `transcripcion` ni `verificacion`: con tilde en comentarios, cadenas y mensajes, y
   reformuladas con el verbo donde no caben tildes (nombres de subtest y de caso, claves, `kebab-case`:
   `defecto-al-componer`, `valor-si-se-omite`…). Las palabras correctas que el diccionario marca (`decisiones`,
   `directos`, `directorios`, `recorre`, `clientes`, `comando`, `comandos`, `defectos`, `posicional`, `producto`,
   `momento`, `contradice`, `distribuye`, `resolverse`, `calcular`, `inaccesibles`) se sustituyen como indica D18. Una
   palabra marcada que no esté en esa lista se reescribe; si es contenido fijado por el contrato, la tarea se detiene y se
   redelimita declarando `.golangci.yml`.
5. **Sin `//nolint`**, sin tests saltados y sin errores silenciados: lecturas de `testdata/` por `filepath.Clean`, escrituras
   de golden y esquemas con `0o600` y en un auxiliar distinto del que lee (G703); ningún `exec` nuevo fuera de `testscript` y
   del `ejecutaGo` existente.
6. **FR-124**: la lista de `modulosDelBinario` se escribe con la salida de `go list -deps -f '{{if .Module}}{{.Module.Path}}{{end}}' ./cmd/kitlegal`
   **medida al enlazar** `boe`, cada módulo con su justificación en el comentario y en `gates/pr-h4.md`.
7. **ADR 0015** en la tarea del paso 1, con el contexto de FR-096 y la firma de `core.Source`.
8. **Documentación alineada con el `Makefile`** en la misma rama: `README.md` y `CONTRIBUTING.md` (filas de
   `make schema-check` y `make verify-sources`), `CHANGELOG.md` (*Añadido*: applet `boe`, `schemas/`, `make verify-sources`;
   *Cambiado*: `make schema-check` activo, `fecha_consulta` declarada por la fuente) y `docs/PENDIENTES.md` (se borran las dos
   entradas de H4).
9. **Evidencias en `gates/pr-h4.md`**, fechadas por commit: salida de `boe-cache-rapida`, cobertura global e `internal/core`,
   lista de módulos, y los escenarios negativos del quickstart sobre el clon desechable.
10. **Umbrales**: global ≥ 70 % e `internal/core` ≥ 85 %; nunca se rebaja un umbral.
11. **La tarea `[plataforma]` va la última**; los supuestos S8 (incidencia nocturna) y S9 (tiempo en CI) quedan anotados como
    pendientes de la primera ejecución.
12. **`rtk`**: toda orden cuya salida se filtre o compare (`go test -v | grep`, `git status --porcelain`, `git diff`) se
    ejecuta con `rtk proxy` en este entorno, cada etapa de la tubería incluida. También va con `rtk proxy` toda orden del
    quickstart que la lista de permitidos de `.claude/settings.json` no cubre, o que ejecuta, crea, borra o edita en la
    carpeta temporal. Esas formas están en la tabla de `quickstart.md` y se ejecutan tal cual.

## Comprobación contra la rúbrica del juez (`juez_plan`, criterios a-j)

| Criterio | Dónde se cumple |
|---|---|
| a. constitution_check | Un ítem por principio (I-IX) y por regla de dependencia (las cinco de §IV, la del grafo del roadmap §2, la de ejemplos del ADR 0010 y la del espacio reservado del ADR 0006), con cómo se cumple, cómo se vigila y veredicto; gates por capa; reglas del modo desatendido; re-evaluación tras el diseño. La fila III coincide con el orden: guiones e2e en el paso 12, justo detrás del registro (11) |
| b. dependencias | Ninguna nueva; las usadas están en §V; los módulos que el binario pasa a enlazar se justifican (FR-124, Complexity Tracking) |
| c. reglas_dependencia | `core` sin adaptadores ni E/S; `net/http` solo en `httpx` (los tests de `boe` usan `Replay`); SQLite solo en `cache` (`boe` usa el puerto); `os.Exit` solo en las dos raíces vía `Arrancar`; stdout solo por el presentador |
| d. errores_exit_codes | `boe.Error` + `httpx.Error` + `cache.Error` con `schema.ConClase`; tabla cerrada de 23 filas → {0, 1, 2, 3, 4, 5}; nunca 6; `TestClasesDeErrorDeBoe`, `TestCodigosDeSalidaDeBoe` |
| e. tests_primero | Inventario con nombres fijos; referencias escritas por una persona antes del código; e2e en la primera tarea que puede dejarlo en verde (paso 12, tras el registro del 11); fixtures cerrados; 22 controles mecánicos con demostración; orden con el test antes del código; la fila de `docs/SOURCES.md` y `terminos.go` los fija la persona antes de grabar, sin romper `make ci` (admisión de `pendiente` retirada en el paso 6) |
| f. alcance | Nada de *Fuera de alcance*; lo no enumerado por el spec está en Complexity Tracking con motivo; las referencias se escriben y revisan en la tarea `[datos]` que graba, como exigen FR-116 y Q2 |
| g. sin_atajos | Ningún `nolint`, `t.Skip`, TODO ni error silenciado previstos; `panic` eliminado del e2e; tipos inesperados fallan en vez de ocultarse; `TestPuntoDeEntrada` sigue ejerciendo la composición de `main()` con el registro de producción, y `&Registro{}` solo donde el sujeto es el registro vacío |
| h. mejor_alternativa | Cada decisión de research.md (D1-D19) con alternativas rechazadas y motivo |
| i. afirmaciones_verificadas | Tabla de verificación de research.md con fichero y línea (Go y su biblioteca estándar, CPython, módulos, `gh`; `golangci-lint` y `misspell`, con el comportamiento citado a su código en D18 y un barrido del diccionario con método repetible; el guardián y el extractor de rutas del workflow en D15, comprobados sobre líneas con rutas en prosa y en órdenes), sin nada citado a la memoria del proyecto; lo que depende del acelerador C de `ElementTree`, de expat o de `gh` contra el repositorio remoto, como supuesto en D20 (S1-S12) |
| j. quickstart_ejecutable | Escenarios con órdenes, rutas y tests reales; negativos solo sobre un clon desechable en `/tmp`, en una carpeta que los prerrequisitos dejan vacía para que un reintento no falle; binario en `/tmp`; ningún escenario escribe en el árbol, el índice o el historial; cada orden en una forma que la sesión desatendida ejecuta sin aprobación (sondeada el 2026-09-13), sin redirecciones a ficheros y con sondas positivas de esas formas en los prerrequisitos |
