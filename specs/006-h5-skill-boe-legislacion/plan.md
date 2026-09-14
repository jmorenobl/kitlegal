# Implementation Plan: H5 · Skill `boe-legislacion` (genérica) + andamiaje de skills

**Branch**: `h5-skill-boe-legislacion` | **Date**: 2026-09-14 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/006-h5-skill-boe-legislacion/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de
`.specify/memory/constitution.md` y están en [research.md](./research.md) con alternativas y motivo. Toda afirmación
sobre una herramienta o dependencia externa remite a su tabla de verificación (V1-V58); lo que no se puede comprobar sin
red o sin la plataforma está en [research.md D22](./research.md) como supuesto (S1-S12).

## Summary

H5 entrega la primera skill del producto, `boe-legislacion`: el protocolo con el que un agente identifica la norma,
resuelve `BOE-A-…`, lee índice y bloques con `scripts/boe` y responde citando `[BOE-A-…, bloque …]`, distinguiendo ley y
reglamento y señalando variación autonómica, para cualquier materia. Y deja el andamiaje con el que se miden las skills:
`data/normas.yaml` validado y verificado contra búsquedas grabadas, generación de `references/`, de la tabla de comandos
desde `--describe` y de los enlaces de `scripts/`, comprobación mecánica en `make ci`, formato común de evals, un job de
evals con Claude Code sin red de ninguna fuente y `make install`.

Decisiones que sostienen el diseño:

1. **Herramientas como tests, no como binarios** (D2, D7): `internal/skills` e `internal/evals`, que el binario no
   enlaza; `make skills-check` y `scripts/skills-sync.sh` ejecutan tests, como `make schema-check`; la tabla sale de
   `describir` en `internal/app` sin exportar nada del kernel.
2. **La skill declara lo que se genera** (D4): `metadata.kitlegal-applets` y `metadata.kitlegal-referencias`.
3. **`scripts/boe -> ../../../bin/instalado/kitlegal`** (D5): `make install` enlaza ese punto al binario instalado.
4. **Normas indexadas por identificador** (D8, D9), con `rango` del vocabulario grabado e identificadores comprobados sin
   red con el propio applet.
5. **Grabación en la pausa de un manifiesto bajo `testdata/` de la raíz** (D11): es la ubicación fuera de la fuente en
   la que un fichero nuevo hace pausar al workflow (V36); se graba ejecutando el applet y reutilizando lo grabado en H4.
6. **Evals comparadas mecánicamente** (D10, D12): activación y respuesta del transcript; invocaciones y conexiones de una
   traza del sistema.
7. **Garantía de red del job** (D16): proxy del entorno que rechaza, verificado en local con el binario, más observación
   independiente de las conexiones.
8. **Sin Python** (D17): sus paquetes purgados, buscado como root en todo el sistema de ficheros, retirado del runner
   con sus instalaciones por un paso que fija sus propias opciones de shell, y comprobado con la misma búsqueda antes de
   cada ejecución (V56, V57; supuestos S2 y S7); la aceptación se registra de la ejecución de cierre.

Artefactos de diseño: [data-model.md](./data-model.md) y [contracts/](./contracts/)
([skill-boe-legislacion](./contracts/skill-boe-legislacion.md), [normas-y-referencias](./contracts/normas-y-referencias.md),
[sincronizacion-y-comprobacion](./contracts/sincronizacion-y-comprobacion.md), [instalacion](./contracts/instalacion.md),
[evals-y-grabaciones](./contracts/evals-y-grabaciones.md), [job-de-evals](./contracts/job-de-evals.md)); validación en
[quickstart.md](./quickstart.md).

## Technical Context

**Language/Version**: Go 1.27 sin cambios (`go 1.27.0` + `toolchain go1.27.1`, exportado por el `Makefile`); Bash para
`scripts/`; YAML y JSON Schema 2020-12 para datos y evals; Markdown para la skill.

**Primary Dependencies**: las de `go.mod`. Pasa a directa `go.yaml.in/yaml/v3` (ya en el grafo; Complexity Tracking,
D8), y no `go.yaml.in/yaml/v4`, que solo tiene versiones candidatas y una sola versión para las herramientas y el
binario (V58). `santhosh-tekuri/jsonschema/v6` (§V) se usa además en `internal/skills` e `internal/evals`, que solo se ejecutan desde
tests. Fuera de Go, solo en el job: Claude Code 2.1.270 y `strace` en `ubuntu-24.04` (Complexity Tracking, D13).
Biblioteca estándar: `encoding/json`, `encoding/json/v2` (el lector del manifiesto, que rechaza la clave repetida; V55),
`net/netip` (clasificar direcciones de la traza), `os/exec` (solo en tests), `regexp`, `bufio`, `strconv`, `slices`,
`maps`.

**Herramientas de control**: las de H0-H4 sin cambio de versión. `.golangci.yml` añade la etiqueta `evals` a
`run.build-tags` y `comando`, `comandos`, `defectos`, `legislativo` y `patrones` a `misspell.ignore-rules`, cada una con
su motivo (D20, V42); ninguna exclusión, ningún `//nolint`.

**Storage**: ninguno nuevo. Cachés de H3 en directorios temporales (por test en `make ci`, por sesión en el job). En
`make install`, enlaces en `~/.claude/skills/` y `bin/instalado/kitlegal`.

**Testing**: `make test` (unitarios y tests del repositorio con `-race`), `make test-integration` (`TestInstalacion`),
`make skills-check` (nuevo, en `make ci`). Todo sin red: `httpx.Replay` sobre grabaciones de H4 y H5, sesiones y trazas
sintéticas. Red de la fuente solo en `scripts/grabar-evals.sh` (una persona, en la pausa) y en `make verify-sources` (sin
cambios). El job de evals (red del modelo) no está en `make ci`.

**Target Platform**: sin cambios para el binario (sin cgo, `-trimpath`); desarrollo en darwin/arm64 y `make ci` en
`ubuntu-latest`; el job en `ubuntu-24.04` (Linux para `strace`).

**Project Type**: CLI multicall + skill del estándar Agent Skills + herramientas de desarrollo.

**Performance Goals**: `make skills-check` en segundos (preparación en proceso de unas treinta consultas contra
`Replay`). Sesión de eval ≤ 240 s; el job completo dentro de 120 min.

**Constraints**: sin Python en ningún paso (FR-013, FR-044, FR-073); sin red de ninguna fuente en `make ci` ni en el
job; `SKILL.md` < 300 líneas; generación determinista e idempotente; ningún cambio en `internal/source/boe`, sus esquemas,
`docs/SOURCES.md`, `internal/app` salvo `skills_test.go`, `internal/cli`, `internal/httpx`, `internal/cache`, ni en
`.claude/skills/`, `skills-lock.json` o el contenido de `.agents/skills/`.

**Scale/Scope**: 1 skill (3 ficheros), 12 evals, 10 normas, 2 esquemas, 2 paquetes de herramienta (~20 ficheros de código
y sus tests), 1 test en `internal/app`, 4 guiones, 1 flujo, un manifiesto con sus grabaciones (del orden de 25 ficheros),
49 casos sintéticos de sesión, traza e informe (195 ficheros; contrato del job §9.1), 4 guiones `testscript`, `.agents/.gitattributes` y documentación.

## Constitution Check

*GATE: debe pasar antes de la fase 0 y volver a evaluarse tras la fase 1.*

### Principios

| # | Principio | Cómo lo cumple H5 | Veredicto |
|---|---|---|---|
| **I** | Fuentes públicas y frontera humana | Ninguna fuente nueva ni código de red: la única fuente es la de H4, sin cambios. La grabación la hace una persona con `httpx` (identificación, `robots.txt`, ritmo de H4) en la pausa de la tarea `[datos]` del manifiesto, que el workflow provoca porque el manifiesto es un fichero nuevo bajo `testdata/` de la raíz (`testdata/evals/grabaciones.json`; V36, D11): bajo `internal/evals/testdata/` no pausaría. El job no deja que ninguna invocación llegue a la red de una fuente y lo comprueba en el sistema (D16). `SKILL.md` prohíbe presentar, notificar, firmar o tramitar nada, o simularlo (FR-015). | ✅ Cumple |
| **II** | Nada sin cita ni fuente | Cada afirmación de la respuesta lleva `[BOE-A-…, bloque …]` sacado del sobre de `scripts/boe --json` de la sesión (FR-008); reglas de nunca inventar, avisos de vigencia, carácter informativo, ley y reglamento, variación autonómica (FR-009 a FR-012). Las citas esperadas se comparan por identificador y bloque, sin modelo (FR-072). `data/normas.yaml` solo con identificadores y títulos de la respuesta grabada (FR-023, FR-024). | ✅ Cumple |
| **III** | Tests primero y offline | Cada tarea escribe su test antes del código. Las evals —el test de aceptación de la skill— se escriben antes que `SKILL.md` y que la generación (FR-065). El e2e `testscript` de la entrega usable nueva, `make install`, entra en una tarea `[datos]` antes del código de instalación. Todo test es offline contra las grabaciones de H4 (`internal/source/boe/testdata/`) y de H5 (`testdata/evals/`) y contra sesiones y trazas sintéticas (`internal/evals/testdata/sesiones/`); la red de la fuente solo en `scripts/grabar-evals.sh` (persona) y `make verify-sources` (sin cambios). Umbrales de cobertura intactos; `internal/core` no se toca. | ✅ Cumple |
| **IV** | Arquitectura hexagonal con reglas ejecutables | No cambia el dominio ni los adaptadores. Las herramientas componen lo exportado (`app.Main`, `app.AppletBoe`, `httpx.Replay`, `cache.ConDirectorio`) como una raíz de composición de test. Errores y códigos de salida del binario sin cambios; las herramientas devuelven errores que nombran skill, fichero, entrada, eval y consulta; los guiones terminan con 0, 1 o el código del test. Ningún `panic`. Reglas de dependencia, abajo. | ✅ Cumple |
| **V** | Simplicidad y dependencias fijadas | Solo lo del spec. Sin DI, ORM ni generador de CLI; sin binarios nuevos. Dependencia directa nueva `go.yaml.in/yaml/v3` en lugar de `gopkg.in/yaml.v3` y herramientas externas del job, justificadas en *Complexity Tracking*. `internal/cli` no se extrae. | ✅ Cumple con justificación |
| **VI** | Un binario, convenciones de agente | El binario no cambia. `scripts/boe` es un enlace que invoca `boe` por multicall (V34, V37). La tabla de comandos sale de `--describe` (FR-032), y la skill usa `--json`. | ✅ Cumple |
| **VII** | Grafo y privacidad | Sin grafo (H7). Ni las evals ni los datos llevan datos personales; ningún municipio como caso. La credencial del modelo es un secreto del repositorio y no llega a las órdenes de la sesión (`CLAUDE_CODE_SUBPROCESS_ENV_SCRUB`, V12). | ✅ Cumple |
| **VIII** | Skills primero; el binario es la herramienta | Entrega la primera skill, medible con 12 evals y un job. No se construye ninguna herramienta en el binario; las de desarrollo existen solo para generar y medir la skill. `SKILL.md` < 300 líneas, `references/` generadas desde `data/`, `scripts/` como enlaces. | ✅ Cumple |
| **IX** | Genericidad territorial, validación local | Ningún caso especial para un municipio o una comunidad en `SKILL.md`, referencias, datos ni evals (FR-012, FR-066); la skill señala que las normas locales no están en la fuente. Sin dimensión territorial: el punto 11 de la Definition of Done solo aplica en lo que prohíbe. | ✅ Cumple |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución §IV)

| Regla | Situación en H5 | Cómo se hace cumplir | Veredicto |
|---|---|---|---|
| `internal/core/**` no importa `internal/{source,httpx,cache,store,graph,render,cli,app}` ni entrada y salida | `internal/core` no se toca | `depguard` lista `core` + `TestArquitectura` R1 (sin cambios) | ✅ Cumple |
| Solo `internal/httpx` importa `net/http` | `internal/evals` usa `httpx.Replay`/`httpx.New` y `net/netip`; `internal/skills` no usa red; ningún test usa `httptest` | `depguard` lista `red` (todo el árbol, tests incluidos) + `TestArquitectura` R2 | ✅ Cumple |
| Solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | `internal/evals` abre la caché con `cache.Opcion` a través de `app.DependenciasDeBoe`; no importa el controlador | `depguard` lista `sql` + `TestArquitectura` R3 | ✅ Cumple |
| Solo `internal/cli` y `cmd/` llaman a `os.Exit` | No hay `package main` nuevo; las herramientas son tests y guiones de shell | `forbidigo` `^os\.Exit$` (sin excepciones nuevas) | ✅ Cumple |
| Solo `internal/render` escribe en stdout; logs con `slog` a stderr | Los tests escriben ficheros y `t.Log`; `describe` se captura con `render.Nuevo` sobre búferes; los guiones de shell imprimen | `forbidigo` `^fmt\.Print…$`, `^os\.Stdout$`, `^os\.Stderr$` (sin excepciones nuevas) | ✅ Cumple |
| `internal/graph` no importa `internal/source/*` ni `internal/render` | `internal/graph` no existe | — (sin objeto) | ✅ Cumple |
| Los applets de ejemplo no se enlazan en el binario distribuido (ADR 0010) | `internal/evals` compone `boe` sin importar `ejemplo` | `depguard` lista `ejemplo` + `TestElBinarioNoEnlazaLosEjemplos` | ✅ Cumple |
| Ningún adaptador firma en el espacio reservado (ADR 0006) | Sin adaptadores nuevos | `TestLasFuentesNoFirmanComoKitlegal` (sin cambios) | ✅ Cumple |
| El binario no enlaza módulos no justificados (FR-124 de H4) | `internal/skills` e `internal/evals` no los importa `cmd/kitlegal`; `go.yaml.in/yaml/v3` y `santhosh-tekuri/jsonschema` siguen fuera del binario | `TestDependenciasDelBinario` (sin cambios) | ✅ Cumple |

### Gates (constitución, «Gates»)

- **Capa 1 (mecánica)** que H5 activa: frontmatter válido en cada `SKILL.md` y `references/` idénticas a lo generado desde
  `data/*.yaml` (las dos nombradas en la constitución), tabla de comandos y enlaces sin deriva, `SKILL.md` < 300 líneas,
  `data/normas.yaml` contra su esquema, identificadores contra la búsqueda grabada, formato y conjunto de evals, lo
  grabado suficiente para cada eval (FR-075), citas esperadas comparadas por identificador (nombrada en la constitución),
  tabla contra gramática, `make install` en e2e. Guardián de diff con `[datos]` para `testdata/` y `schemas/`.
- **Capa 2 (jueces)**: la `description` que activa con lo que debe y no con lo que no debe la miden las evals del job
  (constitución, capa 2); el protocolo de `SKILL.md` lo juzgan los dos jueces de la revisión final con su rúbrica.
- **Capa 3 (humano)**, pausas previstas: `schemas/eval.yaml.json` y `schemas/normas.yaml.json` (ficheros nuevos bajo
  `schemas/`); el manifiesto de grabación `testdata/evals/grabaciones.json`, fichero nuevo bajo `testdata/` de la raíz,
  en cuya pausa graba y revisa la persona (`clasificar_datos` pausa ante ficheros nuevos bajo `testdata/` de la raíz,
  `internal/source/` o `schemas/`, no bajo `internal/evals/testdata/`: V36); los prerrequisitos de plataforma (secreto y
  etiquetas); la fusión. `docs/SOURCES.md` no cambia. Las sesiones sintéticas y los guiones de instalación no pausan:
  son material de prueba nuevo fuera del territorio de fixtures, que revisan los jueces finales.

**Reglas del modo desatendido**: el ejecutor nunca usa `KITLEGAL_RECORD`, `scripts/grabar-evals.sh`, `make evals` ni
`make verify-sources`; nunca escribe un identificador o un título que no haya leído de una respuesta grabada; ninguna tarea
`[datos]` toca código.

**Veredicto del gate: PASA.** Las dos piezas que tocan el principio V están justificadas en *Complexity Tracking* y
ninguna afecta a alcance, frontera humana, privacidad, términos de uso ni a una decisión cerrada.

### Re-evaluación tras la fase 1 (diseño)

- **Herramientas del job fuera de Go** (§V): Claude Code y `strace` no entran en el módulo ni en `make ci`; sin ellas no hay
  job de evals (FR-070) ni registro de invocaciones con código (FR-072, V9).
- **Proxy que rechaza y observación de conexiones** (§I, spec *Fuera de alcance*): cumple FR-074 y FR-076 sin tocar el
  kernel ni la fuente; comprobado en local con el binario (V41).
- **`.agents/.gitattributes` anidado** (FR-085): impuesto por el guardián (V35); efecto de git verificado (V33), efecto de
  GitHub como supuesto S8.
- **`scripts/boe` a `bin/instalado/kitlegal`** (§VIII, decisión cerrada «`scripts/` como symlinks al binario»): sigue
  siendo un enlace al binario; se aparta solo de la ruta de ejemplo de `refs/`, que no es una decisión cerrada.

**Veredicto tras el diseño: PASA**, sin ninguna violación no justificada.

## Project Structure

### Documentation (this feature)

```text
specs/006-h5-skill-boe-legislacion/
├── plan.md              # Este fichero
├── research.md          # Fase 0: V1-V58, D1-D22 (decisiones, alternativas, supuestos S1-S12)
├── data-model.md        # Fase 1: skill, frontmatter, tabla, enlaces, normas, deriva, eval, consultas, caché, trazas, informe
├── quickstart.md        # Fase 1: guía de validación ejecutable
├── contracts/
│   ├── skill-boe-legislacion.md          # SKILL.md: frontmatter, protocolo, cita, reglas
│   ├── normas-y-referencias.md           # data/normas.yaml, schemas/normas.yaml.json, references/normas.md, identificadores
│   ├── sincronizacion-y-comprobacion.md  # make skills-sync, make skills-check, región generada
│   ├── instalacion.md                    # make install, scripts/instalar-skills.sh, TestInstalacion
│   ├── evals-y-grabaciones.md            # formato, evals/boe-legislacion/, manifiesto y arnés, preparación, comparación
│   └── job-de-evals.md                   # evals.yml, make evals, sesión, garantía de red, informe, cierre
├── spec.md
├── checklists/
├── gates/
└── tasks.md             # Fase 2 (/speckit-tasks; no lo crea este comando)
```

### Source Code (repository root)

Estructura de `docs/ROADMAP.md` §2 y `CLAUDE.md`. **NUEVO** = lo crea H5; ← = cambia.

```text
skills/boe-legislacion/                     **NUEVO** el producto
├── SKILL.md                                frontmatter, protocolo, cita, región de comandos generada, reglas    (D3, D4, D6)
├── references/normas.md                    generado desde data/normas.yaml                                        (D7)
└── scripts/boe -> ../../../bin/instalado/kitlegal   generado                                                       (D5)
data/normas.yaml                            **NUEVO** diez normas indexadas por identificador                      (D8)
schemas/
├── normas.yaml.json                        **NUEVO** [datos]                                                      (D8)
└── eval.yaml.json                          **NUEVO** [datos]                                                      (D10)
evals/boe-legislacion/                      **NUEVO** 01-lpac-articulo-21.yaml … 12-no-activa-acuerdo-entre-amigos.yaml (D10)
internal/
├── skills/                                 **NUEVO** herramienta de desarrollo; no la enlaza el binario           (D2)
│   ├── doc.go
│   ├── esquemas.go          validación de un documento YAML contra schemas/*.yaml.json                           (D8)
│   ├── normas.go            Norma, LeerNormas                                                                     (D8)
│   ├── skill.go             Skill, Listar, Cargar, ContarLineas                                                   (D3)
│   ├── frontmatter.go       Frontmatter, DeclaracionDeKitlegal, validación                                        (D3, D4)
│   ├── referencias.go       RenderizarNormas                                                                      (D7)
│   ├── comandos.go          DescripcionDeVerbo, RenderizarTabla, SustituirRegion                                  (D6)
│   ├── enlaces.go           EnlacesEsperados                                                                      (D5)
│   ├── sincronia.go         Regenerar, Escribir, Comparar, Deriva                                                 (D7)
│   ├── *_test.go            uno por fichero de código
│   ├── instalacion_test.go  //go:build integration · TestInstalacion                                             (D15)
│   └── testdata/script/     **NUEVO** [datos] instalar.txtar, instalar-de-nuevo.txtar, instalar-con-conflicto.txtar, instalar-sin-gobin.txtar
├── evals/                                  **NUEVO** herramienta de desarrollo; no la enlaza el binario           (D2)
│   ├── doc.go
│   ├── formato.go           Eval, ComandoEsperado, CitaEsperada, LeerEval                                        (D10)
│   ├── conjunto.go          Conjunto, FicheroMalFormado, LeerConjunto (lectura y formato); NormaConocida,
│   │                        DefectoDelConjunto, ComprobarConjuntoDeBoeLegislacion (reglas de data-model §6.3)  (D10)
│   ├── consultas.go         Consulta, ConsultasNecesarias                                                        (D14)
│   ├── grabaciones.go       Manifiesto, EntradaDelManifiesto, LeerManifiesto, conjuntos de grabaciones, unión    (D11)
│   ├── preparar.go          Preparar, ComprobarSinRed, SesionAPreparar, PrepararSesion                           (D14)
│   ├── trazas.go            Invocacion, Conexion, LeerTrazas, InterpretarInvocacion                              (D12)
│   ├── sesion.go            Sesion, LeerSesion                                                                   (D12)
│   ├── citas.go             ExtraerCitas                                                                         (D12)
│   ├── juzgar.go            ResultadoDeEval, Juzgar                                                              (D12)
│   ├── informe.go           Informe, InformeAEscribir, EscribirInforme                                           (D13)
│   ├── *_test.go            uno por fichero de código
│   ├── grabacion_test.go    //go:build grabacion · TestGrabarEvals                                               (D11)
│   ├── job_test.go          //go:build evals · TestPrepararSesion, TestInformeDelJob                             (D13)
│   └── testdata/
│       └── sesiones/                         **NUEVO** [datos] sintéticos, sin pausa; árbol exacto en el contrato del job §9.1 (D12)
│           ├── leer-sesion/<caso>/           TestLeerSesion: el caso es el directorio de una sesión
│           ├── leer-trazas/<caso>/           TestLeerTrazas: el caso es el traza/ de una sesión
│           └── informe/<caso>/               TestInforme: el caso es una ejecución, con evals/ y sesiones/<sesión>/
└── app/skills_test.go      **NUEVO** TestSkillsDelRepositorio (-regenerar-skills), TestTablaDeComandosCoincideConLaGramatica (D2, D6)
testdata/evals/                             **NUEVO** en la raíz: lo grabado para las evals (aquí sí pausa, V36)  (D11)
├── grabaciones.json                        **NUEVO** [datos] manifiesto (pausa)                                   (D11)
└── boe.legislacion-consolidada/            **NUEVO** grabado por una persona en la pausa                          (D11)
scripts/
├── skills-sync.sh          **NUEVO**                                                                              (D7)
├── instalar-skills.sh      **NUEVO**                                                                              (D5)
├── grabar-evals.sh         **NUEVO** (red; solo una persona, en la pausa)                                         (D11)
└── evals.sh                **NUEVO** (lo ejecuta el job)                                                          (D13)
.github/workflows/evals.yml **NUEVO**                                                                              (D13, D16, D17)
.agents/.gitattributes      **NUEVO**                                                                              (D18)
Makefile                    ← install, skills-sync real, skills-check (en ci), evals (fuera de ci)
.golangci.yml               ← run.build-tags + evals; misspell.ignore-rules + comando, comandos, defectos, legislativo, patrones (D20)
go.mod, go.sum              ← go.yaml.in/yaml/v3 directa                                                           (D8)
docs/PENDIENTES.md          ← se borran las tres entradas «En H5»                                                  (D18)
README.md, CONTRIBUTING.md, CHANGELOG.md   ←                                                                       (D19)
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`: `skills/`, `data/`, `evals/`, `schemas/` y `scripts/`
en la raíz; código en `internal/`. Los dos paquetes nuevos son herramientas de desarrollo que el binario no enlaza (D2).
No se crean `packs/`, `mcp/`, `plugin/`, `pkg/`, `internal/graph` ni `internal/store`.

## Controles mecánicos que este hito añade o toca

Según la sección «Gates» de la constitución. «Demostración» dice qué falla si el control se retira o se viola.

| # | Control | Test / forma | Orden | ¿En `make ci`? | Demostración |
|---|---|---|---|---|---|
| 1 | Frontmatter del estándar | `TestSkillsDelRepositorio/skills` y `/frontmatter`; `TestValidarFrontmatter` | `skills-check`, `test` | Sí | `name` con mayúscula → falla nombrando skill y defecto (quickstart 4) |
| 2 | `SKILL.md` < 300 líneas | `/trescientas-lineas`; `TestContarLineas` | `skills-check`, `test` | Sí | 300 líneas → falla (quickstart 4) |
| 3 | Deriva de `references/` | `/skills`, `/referencia-editada`, `/datos-sin-regenerar`; `TestRegenerarYComparar` | `skills-check`, `test` | Sí | Editar `references/normas.md` → falla nombrando fichero (quickstart 3) |
| 4 | Deriva de la tabla de comandos | `/describe-cambiado`; `TestRenderizarTabla`, `TestSustituirRegion` | `skills-check`, `test` | Sí | Cambiar la ayuda de un verbo sin regenerar → falla en `SKILL.md` (quickstart 3) |
| 5 | Deriva de enlaces de `scripts/` | `/enlaces`; `TestEnlacesEsperados` | `skills-check`, `test` | Sí | Enlace ausente, sobrante o con otro destino → falla (quickstart 4) |
| 6 | Tabla contra gramática | `TestTablaDeComandosCoincideConLaGramatica` | `test` | Sí | Una sintaxis generada que la gramática no acepta → falla |
| 7 | Regeneración idempotente | `/regenerar-dos-veces`; `make skills-sync` dos veces | `test` | Sí | Un segundo `Escribir` que cambie un byte o un enlace → falla (quickstart 2) |
| 8 | Normas contra esquema | `TestLeerNormas`, `TestEsquemaDeNormas`, `TestNormasDelRepositorio` | `skills-check`, `test` | Sí | `vertical`, identificador mal formado o repetido → falla nombrando norma (quickstart 5) |
| 9 | Identificadores contra búsqueda grabada | `TestIdentificadoresDeLasNormas` | `skills-check`, `test` | Sí | Título cambiado → falla nombrando la norma (quickstart 3, SC-008); una norma de `data/normas.yaml` que ninguna entrada del manifiesto resuelve por su prefijo, también si sale entre los resultados de la búsqueda de otra entrada → falla (`TestIdentificadoresDeLasNormas/identificador-cambiado`, `/titulo-cambiado`, `/norma-sin-entrada`, `/norma-en-la-busqueda-de-otra-entrada`; contrato de normas y referencias §6) |
| 10 | Gramáticas iguales a las de `boe` | `TestGramaticasCoincidenConBoe` | `test` | Sí | Un patrón de esquema que rechace `a85bis.` o acepte `BOE-A-15-1` → falla |
| 11 | Formato de eval | `TestLeerEval`, `TestEsquemaDeEval`, `TestLeerConjunto`, `TestEvalsDelRepositorio/formato` | `skills-check`, `test` | Sí | Sin `pregunta` → falla nombrando el fichero (quickstart 6); un fichero mal formado, o una entrada que no es una eval, entre evals bien formadas → `LeerConjunto` los devuelve aparte con su error, sin error del directorio (`TestLeerConjunto/con-mal-formadas`, `/entradas-que-no-son-evals`) |
| 12 | Conjunto de evals | `TestConjuntoDeEvals`, `TestEvalsDelRepositorio/conjunto`, `/normas-conocidas` | `skills-check`, `test` | Sí | Nueve positivas, dos que citan la misma única norma, sin `reproduce` → falla |
| 13 | Lo grabado basta para cada eval (FR-075) y cada sesión del job se prepara con las consultas de todas (FR-074) | `TestEvalsDelRepositorio/grabado`, `TestPrepararYComprobar`, `TestPrepararDirectorioDeSesion` | `skills-check`, `test` | Sí | Retirar los metadatos de la LPAC → falla nombrando eval y orden (quickstart 6, SC-010); preparar solo las consultas de la eval de la sesión, o escribir la pregunta de la prueba de red sin las dos órdenes `a9998` → falla (`TestPrepararDirectorioDeSesion/eval-normal`, `/prueba-de-red`; quickstart 6); preparar la caché o escribir `eval.txt` o `pregunta.txt` con una eval mal formada en el directorio, sin un error que la nombre, o escribirlos tras una falta de lo grabado → falla (`/eval-mal-formada`, `/con-faltas`) |
| 14 | Grabaciones sin solape con H4 y manifiesto bien formado | `TestGrabacionesSinSolape`, `TestManifiestoDeGrabaciones` | `test` | Sí | Un fichero de H5 con el nombre de uno de H4 → falla; dos entradas del manifiesto con el mismo `titulo_empieza_por`, o una cuyo prefijo empieza por el de otra → `LeerManifiesto` falla nombrando las dos (`TestManifiestoDeGrabaciones/prefijo-repetido`, `/prefijo-de-otro-prefijo`), y una clave repetida da un error en lugar de quedarse con el último valor (`/clave-repetida`); la misma repetición en el manifiesto real hace fallar `/repositorio` y, en `make skills-check`, `TestIdentificadoresDeLasNormas`, que lee con el mismo lector (quickstart 6) |
| 15 | Comparación mecánica | `TestInterpretarInvocacion`, `TestExtraerCitas`, `TestJuzgar` | `test` | Sí | Contar un bloque leído con código 4, una cita de otro bloque, o dar por buena una eval de no activación cuya sesión no terminó → falla (`TestJuzgar/sesion-sin-terminar-no-activa`; quickstart 7, SC-009); contar un bloque leído por una invocación sin código, la que el tope dejó sin terminar → falla (`TestJuzgar/bloque-leido-sin-codigo`); contar como consultado el bloque de una invocación con `--describe` o `--dry-run`, o mandar a `fuera_de_lo_grabado` una invocación con código 2 o 3 → falla (`TestJuzgar/describe-y-dry-run-no-satisfacen`, `/otra-fallida`; data-model §6.1 y §10.2) |
| 16 | Lectura de sesión, trazas e informe | `TestLeerSesion`, `TestLeerTrazas`, `TestLeerTrazasSinFicheros`, `TestInforme`, `TestEscribirInformeSinSusEntradas` | `test` | Sí | Conexión pública sin fallo del veredicto, o un `codigo-de-la-sesion` ausente leído como 0 → falla (`TestInforme/llegada-a-la-red`, `TestLeerSesion/sin-fichero-de-codigo`, `TestInforme/sesion-ilegible`; quickstart 7); el `connect` público de un hilo de Go creado con `clone`, o creado por otro hilo, sin atribuir a su invocación, o un fichero sin la línea que lo crea ignorado en silencio → falla (`TestLeerTrazas/hilo-por-clone`, `/hilo-de-un-hilo`, `/fichero-sin-origen`); una invocación sin sus conexiones en el informe → falla (`TestInforme/fuera-de-lo-grabado-no-cambia-el-veredicto`); una línea de señal tomada por ilegible → falla (`TestLeerTrazas/lineas-de-senal`); un fichero sin línea final o una llamada sin resultado rechazados en una sesión cortada, o admitidos en una sin corte, o esa llamada admitida delante de otra → falla (`/cortada-por-el-tope`, `/sin-linea-final-sin-corte`, `/cortada-con-llamada-interrumpida`, `/llamada-interrumpida-sin-corte`, `/llamada-interrumpida-antes-del-final`); una sesión cortada leída como ilegible, o sin sus invocaciones, sus conexiones de clase `red` y las fuera de lo grabado en el informe → falla (`TestInforme/sesion-sin-terminar`, `/sesion-cortada-con-invocaciones`); un campo de cabecera del informe ausente, recortado o tomado de otro sitio —`modelo` o `commit` distintos de las constantes que recibe `EscribirInforme`, `modelos_de_sesion` o `versiones_de_claude_code` distintas de las de los transcripts sintéticos (el modelo de las sesiones es a propósito distinto del constante), `sin_python` distinto byte a byte de `sin-python.txt`, o cualquiera de esos valores ausente de su línea de `informe.md`— → falla (`TestInforme/aprobado`); sin ninguna sesión legible, `modelos_de_sesion` o `versiones_de_claude_code` no vacías → falla (`TestInforme/sesion-ilegible`; data-model §10.3); veredicto `aprobado` con una eval bien formada que ninguna sesión juzga, o sin ninguna eval ni sesión → falla (`TestInforme/eval-sin-sesion`, `TestEscribirInformeSinSusEntradas/evals-y-sesiones-vacios`); una sesión sin `eval.txt`, con un `eval.txt` que no nombra ninguna eval o sin `pregunta.txt` que no quede sin pasar con `sesión ilegible: <fichero>: …` → falla (`TestInforme/sin-eval-txt`, `/eval-desconocida`, `/sin-pregunta-txt`); un `SinPython`, unas `Sesiones` o unas `Evals` que no se pueden leer, o un `Destino` en el que no se puede escribir, sin un error que nombre la ruta o con `informe.md` o `informe.json` escritos → falla (`TestEscribirInformeSinSusEntradas/sin-python-inexistente`, `/sesiones-inexistente`, `/evals-inexistente`, `/destino-es-un-fichero`; contrato del job §3.3); una sesión con la traza ilegible que pase o sin el motivo `sesión ilegible: traza…`, o un `traza/` vacío o inexistente sin un error que nombre el directorio → falla (`TestInforme/traza-ilegible`, `TestLeerTrazasSinFicheros/vacio`, `/inexistente`; data-model §9, regla 1); la `respuesta` o el `fin_de_la_sesion` de una sesión ausentes del informe, o su pregunta y su respuesta ausentes de su sección de `informe.md` → falla (`TestInforme/aprobado`, `/sesion-sin-terminar`; data-model §10.2); un fichero mal formado o una petición llegada a la red sin su motivo con su texto en la raíz del informe y en `informe.md` → falla (`TestInforme/fichero-mal-formado`, `/llegada-a-la-red`; data-model §10.3); un `Fin` con otro texto que el de data-model §10.1, o un transcript vacío leído como ilegible → falla (`TestLeerSesion/error-max-turns`, `/result-con-is-error`, `/sin-result`, `/tope-agotado`, `/sin-mensajes`) |
| 17 | `make install` | `TestInstalacion` (4 guiones) | `test-integration` | Sí | Modificar una entrada en conflicto → falla (quickstart 8, SC-006) |
| 18 | Sin instrucciones de evals en la skill | `/sin-instrucciones-de-evals` | `skills-check`, `test` | Sí | `KITLEGAL_CACHE_DIR` en `SKILL.md` → falla (FR-077) |
| 19 | Normas nombradas en `SKILL.md` | `/normas-nombradas` | `skills-check`, `test` | Sí | «Ley 1/2000» en `SKILL.md` sin entrada en datos → falla (FR-020) |
| 20 | Lint de los ficheros etiquetados | `run.build-tags: [integration, fuentes, grabacion, evals]` | `lint` | Sí | Un error de tipos en `job_test.go` → `make lint` falla |
| 21 | R1-R5, binario, formato, `-race`, `govulncheck`, `gosec`, `gitleaks`, `go mod verify`, `tidy -diff` | los de H0-H4, sin exclusiones nuevas | varias | Sí | Los mismos de H4 |
| 22 | Evals con modelo | `.github/workflows/evals.yml` → `make evals` | `evals` | **No** (semanal, manual, por etiqueta) | Una cita esperada que la respuesta no contiene → veredicto `fallo` |

### Objetivos del `Makefile`

- `install` (←): `go install` de H0 + `scripts/instalar-skills.sh "$$(go list -f '{{.Target}}' ./cmd/kitlegal)"`.
- `skills-sync` (←): `scripts/skills-sync.sh`, con `check-tools`; deja de anunciar H5.
- `skills-check` (**nuevo**, en `ci`): una sola invocación de `go test` de los tests del repositorio.
- `evals` (**nuevo**, fuera de `ci`): `scripts/evals.sh "$(SKILL)"`.
- `ci` (←): `fmt-check lint test test-integration vuln schema-check skills-check secrets mod-verify mod-tidy-check`.

### Fixtures y datos protegidos

| Material | Tarea | Pausa |
|---|---|---|
| `schemas/eval.yaml.json` | `[datos]` | sí (fichero nuevo bajo `schemas/`) |
| `schemas/normas.yaml.json` | `[datos]`, tras la grabación | sí |
| manifiesto de H5 y grabaciones de H5 | `[datos]` del manifiesto; graba una persona | sí |
| `internal/evals/testdata/sesiones/` (árbol del contrato del job §9.1) | tres `[datos]`, una por subdirectorio | no (material de test nuevo fuera del territorio de fixtures, V36) |
| `internal/skills/testdata/script/` | `[datos]` | no |
| grabaciones de H4 | ninguna tarea las cambia | — |

### Tests de contrato

| Contrato | Tests |
|---|---|
| [skill-boe-legislacion](./contracts/skill-boe-legislacion.md) | `TestSkillsDelRepositorio` (`skills`, `normas-nombradas`, `sin-instrucciones-de-evals`) |
| [normas-y-referencias](./contracts/normas-y-referencias.md) | `TestLeerNormas`, `TestEsquemaDeNormas`, `TestNormasDelRepositorio`, `TestRenderizarNormas`, `TestIdentificadoresDeLasNormas`, `TestGramaticasCoincidenConBoe` |
| [sincronizacion-y-comprobacion](./contracts/sincronizacion-y-comprobacion.md) | `TestSkillsDelRepositorio`, `TestTablaDeComandosCoincideConLaGramatica`, `TestLeerFrontmatter`, `TestValidarFrontmatter`, `TestContarLineas`, `TestSustituirRegion`, `TestRenderizarTabla`, `TestEnlacesEsperados`, `TestRegenerarYComparar` |
| [instalacion](./contracts/instalacion.md) | `TestInstalacion` |
| [evals-y-grabaciones](./contracts/evals-y-grabaciones.md) | `TestLeerEval`, `TestEsquemaDeEval`, `TestLeerConjunto`, `TestConjuntoDeEvals`, `TestEvalsDelRepositorio`, `TestConsultasNecesarias`, `TestManifiestoDeGrabaciones`, `TestGrabacionesSinSolape`, `TestPrepararYComprobar`, `TestInterpretarInvocacion`, `TestExtraerCitas`, `TestJuzgar` |
| [job-de-evals](./contracts/job-de-evals.md) | `TestPrepararDirectorioDeSesion`, `TestLeerSesion`, `TestLeerTrazas`, `TestInforme`, `TestEscribirInformeSinSusEntradas`; la prueba de red y la ejecución de cierre en la plataforma |

## Inventario de tests

Nombres fijados aquí para que `tasks.md` y `quickstart.md` los usen tal cual. Un fichero de test por fichero de código, en
el orden del fichero (`golang-testing`); tablas con subtests con nombre; `t.Parallel()` salvo donde se use `t.Setenv`.
Excepciones declaradas: los `doc.go` no tienen test; `internal/app/skills_test.go` no tiene `skills.go` (como
`esquemas_test.go` de H4, es el pegamento con `describir`); `instalacion_test.go`, `grabacion_test.go` y `job_test.go` son
arneses con etiqueta sin fichero de código propio.

| Fichero | Tests |
|---|---|
| `internal/skills/esquemas_test.go` | `TestValidarDocumentoYAML` (esquema en línea: válido, clave desconocida, tipo, patrón; errores con la ruta del documento) |
| `internal/skills/normas_test.go` | `TestLeerNormas`, `TestEsquemaDeNormas` (`compila`, `rangos-grabados`), `TestNormasDelRepositorio` |
| `internal/skills/skill_test.go` | `TestContarLineas`, `TestListarYCargar` |
| `internal/skills/frontmatter_test.go` | `TestLeerFrontmatter`, `TestValidarFrontmatter` |
| `internal/skills/referencias_test.go` | `TestRenderizarNormas` |
| `internal/skills/comandos_test.go` | `TestDescripcionDeVerbo`, `TestRenderizarTabla`, `TestSustituirRegion` |
| `internal/skills/enlaces_test.go` | `TestEnlacesEsperados` |
| `internal/skills/sincronia_test.go` | `TestRegenerarYComparar` |
| `internal/skills/instalacion_test.go` | `TestInstalacion` |
| `internal/app/skills_test.go` | `TestSkillsDelRepositorio` (subtests del contrato de sincronización §2), `TestTablaDeComandosCoincideConLaGramatica` |
| `internal/evals/formato_test.go` | `TestLeerEval`, `TestEsquemaDeEval`, `TestGramaticasCoincidenConBoe` |
| `internal/evals/conjunto_test.go` | `TestLeerConjunto` (`bien-formadas`, `con-mal-formadas`, `entradas-que-no-son-evals`, `directorio-inexistente`; contrato de evals §1), `TestConjuntoDeEvals`, `TestEvalsDelRepositorio` (`formato`, `conjunto`, `normas-conocidas`, `grabado`; contrato de evals §2) |
| `internal/evals/consultas_test.go` | `TestConsultasNecesarias` |
| `internal/evals/grabaciones_test.go` | `TestManifiestoDeGrabaciones` (contrato de evals §3.1): los sintéticos `valido`, `clave-desconocida`, `clave-repetida`, `datos-tras-el-valor`, `otra-fuente`, `sin-normas`, `busqueda-vacia`, `prefijo-vacio`, `para-vacio`, `bloque-mal-formado`, `bloque-repetido`, `prefijo-repetido` y `prefijo-de-otro-prefijo` (paso 3), y `repositorio`, sobre el manifiesto real (paso 6); `TestGrabacionesSinSolape`, `TestIdentificadoresDeLasNormas` (contrato de normas y referencias §6): `identificador-cambiado`, `titulo-cambiado`, `norma-sin-entrada`, `norma-en-la-busqueda-de-otra-entrada` |
| `internal/evals/preparar_test.go` | `TestPrepararYComprobar`, sobre una copia de las grabaciones de H4 y la eval sintética `01-lpac-articulo-21.yaml` escrita desde constantes en un temporal (contrato de evals §5.2): `completo` (ni faltas ni error), `sin-metadatos-de-un-bloque-esperado`, `sin-indice-de-una-norma`, `sin-bloque-de-una-cita`, `consulta-caducada`; `TestPrepararDirectorioDeSesion` (contrato del job §9): `eval-normal`, `prueba-de-red`, `eval-inexistente`, `eval-mal-formada`, `con-faltas`. No se llama `TestPrepararSesion` porque ese nombre es el del arnés con etiqueta `evals` de `job_test.go`, en el mismo paquete, y con la etiqueta las dos definiciones chocarían |
| `internal/evals/trazas_test.go` | `TestLeerTrazas`, un subtest por caso de `internal/evals/testdata/sesiones/leer-trazas/` (contrato del job §9 y §9.1): `argv-escapado`, `salida-con-codigo`, `muerte-por-senal`, `hilo-por-clone`, `hilo-por-clone3`, `hilo-de-un-hilo`, `connect-fuera-de-la-invocacion`, `connect-local-rechazado`, `connect-publico-rechazado`, `connect-publico-aceptado`, `connect-publico-en-curso`, `connect-ipv6-publico-en-curso`, `connect-af-unix`, `fichero-sin-origen`, `fichero-ilegible`, `lineas-de-senal`, `cortada-por-el-tope`, `sin-linea-final-sin-corte`, `cortada-con-llamada-interrumpida`, `llamada-interrumpida-sin-corte`, `llamada-interrumpida-antes-del-final`; `TestLeerTrazasSinFicheros`, sobre directorios temporales (contrato del job §9; data-model §9, regla 1): `vacio` e `inexistente` (error que nombra el directorio); `TestInterpretarInvocacion` |
| `internal/evals/sesion_test.go` | `TestLeerSesion`, un subtest por caso de `internal/evals/testdata/sesiones/leer-sesion/` (contrato del job §9 y §9.1): `activada` (código 0 leído, modelo, versión y respuesta), `no-activada`, `otra-skill-activada`, `codigo-distinto-de-cero` (código 1), `tope-agotado` (código 124, cortada), `senal-tras-el-tope` (código 137, cortada), `sin-result`, `error-max-turns`, `result-con-is-error`, `sin-fichero-de-codigo`, `codigo-no-entero`, `sin-transcript`, `sin-salida-de-error`, `linea-ilegible`, `sin-mensajes` (transcript vacío con código 124: cortada, no ilegible); en cada subtest que lee el transcript, `Fin` con su texto fijo de data-model §10.1 (`result success`, `result error_max_turns`, `result success con is_error`, `assistant`, `system`, `sin mensajes`; contrato del job §9) |
| `internal/evals/citas_test.go` | `TestExtraerCitas` |
| `internal/evals/juzgar_test.go` | `TestJuzgar` (contrato de evals §6): `pasa`, `positiva-no-activada`, `no-activa-pero-activada`, `sesion-sin-terminar-no-activa`, `sesion-sin-terminar-positiva`, `bloque-leido-con-codigo-4`, `bloque-leido-sin-codigo`, `articulos-satisface-un-bloque`, `metadatos-no-satisface-indice`, `buscar-terminos-como-palabras`, `cita-de-otro-bloque`, `cita-de-otra-norma`, `fuera-de-lo-grabado-con-y-sin-offline`, `red-no-cambia-la-eval`, `describe-y-dry-run-no-satisfacen`, `otra-fallida`, `sc-009-otro-bloque`, `sc-009-otra-norma` |
| `internal/evals/informe_test.go` | `TestInforme`, un subtest por caso de `internal/evals/testdata/sesiones/informe/` (contrato del job §9 y §9.1): `aprobado` (veredicto y cabecera: `modelo`, `modelos_de_sesion`, `versiones_de_claude_code`, `commit` y `sin_python` con su valor exacto en `informe.json` y en `informe.md`; data-model §10.3; y `respuesta` y `fin_de_la_sesion` de `01-lpac-articulo-21`, con su pregunta y su respuesta en la sección de la sesión de `informe.md`; data-model §10.2), `fuera-de-lo-grabado-no-cambia-el-veredicto`, `fichero-mal-formado` (`motivos` de la raíz exactamente `02-sin-pregunta.yaml: mal formado: …`), `eval-que-no-pasa`, `sesion-sin-terminar` (`respuesta` vacía y `fin_de_la_sesion` `system`), `sesion-ilegible` (también las dos listas de la cabecera vacías), `llegada-a-la-red` (`motivos` de la raíz exactamente la petición llegada a la red), `sesion-cortada-con-invocaciones`, `eval-sin-sesion` (una eval bien formada que ninguna sesión juzga: veredicto `fallo`), `sin-eval-txt`, `eval-desconocida`, `sin-pregunta-txt` (cada fichero del guion que falta, o un `eval.txt` que no nombra ninguna eval, deja la sesión sin pasar), `traza-ilegible` (una traza ilegible deja la sesión sin pasar con `sesión ilegible: traza…`); `TestEscribirInformeSinSusEntradas`, sobre directorios temporales (contrato del job §3.3 y §9): `sin-python-inexistente`, `sesiones-inexistente`, `evals-inexistente`, `destino-es-un-fichero` (error que nombra la ruta y ningún informe escrito) y `evals-y-sesiones-vacios` (veredicto `fallo` sin nada que juzgar) |
| `internal/evals/grabacion_test.go` | `TestGrabarEvals` |
| `internal/evals/job_test.go` | `TestPrepararSesion`, `TestInformeDelJob` |

## Orden de implementación

`formato → preparación → datos grabados → normas → evals → generación → skill → instalación → comparación → job →
documentación → plataforma` (D1). Cada tarea escribe el test antes del código y deja `make ci` en verde.

1. **`[datos]` Esquema de eval** `schemas/eval.yaml.json` (contrato de evals §1) → pausa.
2. **Formato de eval** (código): `internal/skills/doc.go`, `internal/skills/esquemas.go`, `internal/evals/doc.go`,
   `internal/evals/formato.go`, `internal/evals/conjunto.go` (`LeerConjunto` y `ComprobarConjuntoDeBoeLegislacion`, con
   las firmas del contrato de evals §1 y §2, y sus tests `TestLeerConjunto` y `TestConjuntoDeEvals`; sin
   `TestEvalsDelRepositorio`, que llega con las evals en el paso 7),
   `internal/evals/consultas.go`, con sus tests; `go.mod`, `go.sum`; `.golangci.yml` (`comando`, `comandos`,
   `defectos`, `legislativo` y `patrones` en `misspell.ignore-rules`, cada una con su motivo: obligación 4, D20).
3. **Preparación y arnés** (código): `internal/evals/grabaciones.go` (`Manifiesto`, `EntradaDelManifiesto` y
   `LeerManifiesto`, con la firma y las reglas del contrato de evals §3.1, y la unión de conjuntos de grabaciones) y
   `internal/evals/grabaciones_test.go` con `TestManifiestoDeGrabaciones` y solo sus subtests sintéticos, escritos como
   constantes (sin `repositorio`, que llega en el paso 6); `internal/evals/preparar.go` (`Preparar`,
   `ComprobarSinRed` y `PrepararSesion`, con las firmas del contrato de evals §5 y del contrato del job §3.2) y sus
   tests —`TestPrepararYComprobar` y `TestPrepararDirectorioDeSesion`, sobre una copia de las grabaciones de H4 y evals
   sintéticas de la LPAC que cada test escribe desde constantes en un temporal (contrato de evals §5.2 y del job §9),
   porque las grabaciones de H5 no existen hasta la pausa del paso 4—,
   `internal/evals/grabacion_test.go` (lee el manifiesto con `LeerManifiesto`), `scripts/grabar-evals.sh`.
4. **`[datos]` Manifiesto** `testdata/evals/grabaciones.json` → **pausa** (fichero nuevo bajo la carpeta de datos de
   prueba de la raíz, la única ubicación fuera de la fuente en la que un fichero nuevo pausa: V36, D11; junto al código
   de la herramienta de evals daría `pausa:false` y nadie grabaría): una persona graba y revisa según
   [contracts/evals-y-grabaciones.md §3.3](./contracts/evals-y-grabaciones.md).
5. **`[datos]` Esquema de normas** `schemas/normas.yaml.json`, con el `enum` de `rango` leído de las búsquedas grabadas
   → pausa.
6. **Normas** (código): `internal/skills/normas.go` y su test (`TestLeerNormas`, `TestEsquemaDeNormas`);
   `TestGrabacionesSinSolape`; el subtest `repositorio` de `TestManifiestoDeGrabaciones` sobre `grabaciones.json`, que
   no puede llegar antes porque el manifiesto lo crea el paso 4 y el 5 es `[datos]` sin código (contrato de evals §3.1);
   y la parte de normas de `TestGramaticasCoincidenConBoe`.
7. **Evals y datos**: los doce ficheros de `evals/boe-legislacion/`, cada uno por su ruta; `data/normas.yaml`; y
   `TestNormasDelRepositorio`, `TestEvalsDelRepositorio`, `TestIdentificadoresDeLasNormas`. Identificadores, títulos y
   bloques copiados de las respuestas grabadas (obligación 2). Antes que `SKILL.md` y que la generación (FR-065).
8. **Generación** (código): `internal/skills/skill.go`, `frontmatter.go`, `referencias.go`, `comandos.go`, `enlaces.go`,
   `sincronia.go`, con sus tests.
9. **Pegamento y órdenes** (código): `internal/app/skills_test.go`, `scripts/skills-sync.sh`, `Makefile` (`skills-sync`,
   `skills-check`, `ci`).
10. **La skill**: `skills/boe-legislacion/SKILL.md` escrito según el contrato de la skill, y `make skills-sync` para
    generar `skills/boe-legislacion/references/normas.md` y `skills/boe-legislacion/scripts/boe`; el subtest `skills`
    pasa a exigir que `boe-legislacion` exista (obligación 12).
11. **`[datos]` Guiones de instalación**: `internal/skills/testdata/script/instalar.txtar`, `instalar-de-nuevo.txtar`,
    `instalar-con-conflicto.txtar`, `instalar-sin-gobin.txtar`, cada uno por su ruta completa.
12. **Instalación** (código): `scripts/instalar-skills.sh`, `Makefile` (`install`), `internal/skills/instalacion_test.go`.
13. **`[datos]` Sesiones sintéticas**, en tres tareas, una por subdirectorio de `internal/evals/testdata/sesiones/`, con
    el árbol exacto del [contrato del job §9.1](./contracts/job-de-evals.md) y cada fichero por su ruta completa
    (obligación 3.1); sin pausa (V36). Cada test lee un nivel distinto, y por eso cada subdirectorio tiene el suyo:
    (a) `leer-sesion/<caso>/`: los 15 subtests de `TestLeerSesion`, y el directorio del caso **es** el de una sesión
    (`sesion.jsonl`, `codigo-de-la-sesion` y `sesion.err`, menos el fichero cuya ausencia prueba el caso; en
    `sin-mensajes`, `sesion.jsonl` vacío); 42 ficheros.
    (b) `leer-trazas/<caso>/`: los 21 subtests de `TestLeerTrazas`, y el directorio del caso **es** el `traza/` de una
    sesión (`t.2000`; además `t.2001` en `hilo-por-clone`, `hilo-por-clone3`, `cortada-por-el-tope`,
    `sin-linea-final-sin-corte`, `cortada-con-llamada-interrumpida` y `llamada-interrumpida-sin-corte`, `t.2001` y
    `t.2002` en `hilo-de-un-hilo`, `t.2002` en `fichero-sin-origen`, y `t.1000` en `connect-fuera-de-la-invocacion` y en
    `lineas-de-senal`); los casos de corte, con lo que puede dejar un corte (data-model §9, regla 6; research.md V54);
    32 ficheros. (c) `informe/<caso>/`: los 13 subtests de
    `TestInforme`, y el directorio del caso está un nivel por encima de las sesiones, porque es una ejecución entera:
    `evals/` con las evals sintéticas como ficheros de datos, que el test pasa a `EscribirInforme` como `Evals`, y
    `sesiones/<sesión>/` con `eval.txt`, `pregunta.txt`, `sesion.jsonl`, `codigo-de-la-sesion`, `sesion.err` y
    `traza/t.<pid>`, menos el fichero cuya ausencia prueba el caso —la sesión `11-no-activa-programacion` en
    `sesion-sin-terminar`; `01-lpac-articulo-21-prueba-de-red` junto a `01-lpac-articulo-21` en
    `fuera-de-lo-grabado-no-cambia-el-veredicto`; `01-lpac-articulo-21` sin `codigo-de-la-sesion` en `sesion-ilegible`,
    sin `eval.txt` en `sin-eval-txt` y sin `pregunta.txt` en `sin-pregunta-txt`; y las evals 01 y 11 con solo la sesión
    de la 01 en `eval-sin-sesion`; y `01-lpac-articulo-21` con la línea `execve` de `t.2000` cortada en
    `traza-ilegible`—; más `informe/sin-python.txt`, común a los trece; 121 ficheros. El contrato de error de
    `EscribirInforme` y la traza sin ficheros no añaden ficheros: `TestEscribirInformeSinSusEntradas` y
    `TestLeerTrazasSinFicheros` usan directorios temporales (contrato del job §9).
14. **Comparación e informe** (código): `internal/evals/trazas.go`, `sesion.go`, `citas.go`, `juzgar.go`, `informe.go`,
    con sus tests; `internal/evals/job_test.go`; `.golangci.yml` (etiqueta `evals`).
15. **Job** (código): `scripts/evals.sh`, `Makefile` (`evals`), `.github/workflows/evals.yml`.
16. **Atributos y documentación**: `.agents/.gitattributes`, `docs/PENDIENTES.md`, `README.md`, `CONTRIBUTING.md`,
    `CHANGELOG.md`; validación de quickstart §1 a §11, citado por su nombre.
17. **`[plataforma]` Propuesta de cambio y prueba de red**: empujar, abrir la propuesta de cambio con
    `gates/pr-h5.md`, comprobar prerrequisitos (quickstart §12.1) y ejecutar la prueba de red (§12.2), registrada en
    `gates/prueba-de-red.md` con la salida del paso de retirada de Python y el `sin_python` del informe (supuesto S7).
18. **`[plataforma]` Ejecución de cierre y aceptación** (quickstart §12.3): `gates/evals-cierre.md` y
    `gates/aceptacion.md`. Es la última: después solo cambian ficheros del directorio del hito (FR-082).

## Complexity Tracking

> Piezas que el spec no enumera y el diseño necesita, y dependencias fuera de la lista de la constitución §V.

| Divergencia | Por qué es necesaria | Alternativa más simple, y por qué se rechaza |
|---|---|---|
| **`go.yaml.in/yaml/v3` en lugar de `gopkg.in/yaml.v3`** (§V) | Hay que leer YAML (`data/normas.yaml`, evals, frontmatter). Es el mismo código que la dependencia de la lista, mantenido por la organización YAML desde que `go-yaml` quedó sin mantenimiento, y ya está en el grafo del módulo (V31): no añade ningún módulo. La organización lo mantiene como legado congelado, con solo correcciones de seguridad (V58 (2)); lo que el lector usa de él —`yaml.Node` con líneas y `(*yaml.Node).Decode` con su comprobación de repetidos— está comprobado (V43) | *`gopkg.in/yaml.v3`*: literalmente en la lista, pero sin mantenimiento y con un módulo más en `go.sum`. *`go.yaml.in/yaml/v4`*: ya en el grafo, y la ruta que su README pide a los proyectos nuevos, pero sin ninguna versión estable (solo `v4.0.0-rc.1` a `rc.6`), con la API y el texto del error de clave repetida cambiados entre la candidata del grafo y la última, y con una sola versión para las herramientas y para `internal/cli`, que la enlaza en el binario: subirla para unas movería la del otro (V58, D8) |
| **Claude Code 2.1.270, `strace` y el runner `ubuntu-24.04` en el job** (§V) | FR-070 pide ejecutar las evals con Claude Code; FR-072 pide invocaciones con su código, que la herramienta Bash no da (V9); ninguna entra en el módulo ni en `make ci` | *`anthropics/claude-code-action`*: no verificable en local y sin traza. *Ganchos o transcript*: sin código por invocación (V9, V13) |
| **Secreto `CLAUDE_CODE_OAUTH_TOKEN` y etiquetas `evals`, `evals-prueba-de-red`** | Credencial del modelo (spec, *Assumptions*): el token de la suscripción de Claude que genera `claude setup-token`, porque el proyecto no usa clave de API de pago por uso; y disparo sobre la rama del hito antes de fusionar (FR-070, D13) | *`gh workflow run --ref`*: no documentado para un fichero que aún no está en la rama principal (V14) |
| **Herramientas como tests** (`-regenerar-skills`, etiquetas `grabacion` y `evals`) | Sin `package main` nuevo ni excepciones de lint; mismo patrón que `schema-check`, `verify-sources` y `grabar-fixtures` (D2, D7) | *Binarios de herramienta*: excepciones de `forbidigo` y contra `golang-project-layout` |
| **`internal/app/skills_test.go` sin `skills.go`** | La tabla debe salir de `describir`, no exportada (V25) (D2) | *Exportar `app.Describir`*: cambio del kernel que el hito no necesita |
| **`bin/instalado/kitlegal`** en lugar de `bin/kitlegal` de `refs/` | Que `scripts/boe` ejecute el binario **instalado** (FR-051) sin mezclarlo con `make build` (D5) | *`bin/kitlegal`*: la skill instalada ejecutaría la última construcción |
| **`metadata.kitlegal-applets` y `metadata.kitlegal-referencias`** | FR-035 y FR-042 (ausencias detectables) sin listas por skill (D4) | *Deducir de `scripts/`*: no detecta un enlace borrado |
| **`schemas/eval.yaml.json`** | Formato común documentado y validado mecánicamente (FR-060, US4-4) (D10) | *Validación solo en Go*: formato en prosa |
| **`TestTablaDeComandosCoincideConLaGramatica`** | Vigila la única inferencia que `--describe` no declara (D6) | *Confiar en la inferencia*: una bandera obligatoria futura se presentaría mal sin que nada fallara |
| **Prueba de red por etiqueta o entrada `prueba_de_red`** | SC-012 pide provocar dos invocaciones fuera de lo grabado en una sesión evaluada (contrato del job §6) | *Añadir una eval*: rompería las 10 positivas de FR-062 |
| **`.agents/.gitattributes` anidado** | El guardián no admite `.gitattributes` en la raíz (V35); git lo aplica igual (V33) (D18) | *Cambiar el extractor del workflow*: proceso, fuera de la rama del hito |
| **`make skills-check` y `make evals`** (no están en la lista de H0) | El `Makefile` es la única superficie de invocación (`CLAUDE.md`); el job llama a `make` (D7, D13) | *Llamar a los guiones desde el flujo*: segundo camino |
| **Grabaciones de H5 en `testdata/evals/`** (el manifiesto `testdata/evals/grabaciones.json` y lo grabado en `testdata/evals/boe.legislacion-consolidada/`) | Es la única ubicación fuera de la fuente en la que un fichero nuevo hace pausar al workflow (`clasificar_datos`, V36): en esa pausa graba y revisa la persona (FR-023, capa 3), y H5 no cambia `internal/source/boe` (spec, *Fuera de alcance*) (D11) | *`internal/evals/testdata/`*, junto al código: un fichero nuevo ahí no pausa (`pausa:false`), nadie grabaría y el esquema de normas se quedaría sin rangos grabados. *Directorio de H4*: pausaría, pero cambia la fuente |
| **Etiqueta `evals` en `run.build-tags`; `comando`, `comandos`, `defectos`, `legislativo` y `patrones` en `misspell.ignore-rules`** | Sin la etiqueta, `job_test.go` queda sin lint (V38). `misspell` marca las cinco (V42) y el código de H5 tiene que escribirlas sueltas: `comandos` es la clave del formato de eval y del informe, y `comando` su singular; `defectos`, lo que `Regenerar` devuelve (data-model §5); `legislativo`, parte de «Real Decreto Legislativo», que la expresión de `normas-nombradas` casa tal cual; `patrones`, los `pattern` que compara `TestGramaticasCoincidenConBoe` (D20) | *Otra palabra que `comandos`*: el formato dejaría de usar la del spec. *Clases de caracteres en la expresión* (`[Ll]egislativo`): esquivaría el lint sin ganar nada. *Ignorar también las otras veintiuna de V42* (`recorre`, `directorios`, `terminaron`, `candidatas`…): son prosa de los artefactos o del spec, o llevan tilde; ninguna regla las exige literales en Go (`terminaron`, la de FR-071, no es ninguna clave: la del informe es `fuera_de_lo_grabado`, y sus comentarios dicen «sin terminar») y cada entrada de `ignore-rules` responde a una palabra que el código necesita |

## Obligaciones que este plan traslada a `tasks.md`

1. **Orden y pausas** del apartado anterior. Ninguna tarea `[datos]` toca código ni `.golangci.yml`; las dos
   `[plataforma]` van al final; si la prueba de red descubre un defecto (p. ej. S4 o S7), el arreglo va en una tarea nueva
   **antes** de la de cierre, con `[datos]` si añade una traza real a las sesiones sintéticas.
2. **El ejecutor nunca graba ni consulta la red, ni escribe lo que fija la persona**: ninguna tarea ejecuta
   `KITLEGAL_RECORD`, `scripts/grabar-evals.sh`, `make evals`, `make verify-sources` ni `gh workflow run`; identificadores,
   títulos, rangos y bloques de `data/normas.yaml`, de las evals y del `enum` de `rango` se copian solo de las respuestas
   grabadas (el registro de la grabación y los ficheros grabados). Si falta algo, la tarea se detiene y lo anota en
   `gates/tarea-Tnnn.md`.
3. **Rutas declaradas** (guardián de diff; V35, V36). El extractor toma como declarada toda ruta que aparezca en
   cualquier parte de la línea, también la de una orden (`./internal/evals/` declara el paquete entero, `testdata/`
   incluido). Por eso:
   1. cada fichero que la tarea cambia va por su ruta completa, sin llaves ni comodines; los doce ficheros de evals y los
      cuatro guiones, cada uno por la suya;
   2. lo que solo se lee, se compara o se ejecuta se nombra sin directorio (`grabaciones.json`, «las grabaciones de H4»,
      «las grabaciones de H5», `normas.yaml.json`) o por su test o su objetivo de `make` (`TestEvalsDelRepositorio`,
      `make skills-check`), nunca con la ruta de un paquete;
   3. desde la tarea que sigue al manifiesto (paso 5), ninguna ruta que la tubería de `workflow.yml` 615 extraiga de una
      línea es la del manifiesto (`testdata/evals/grabaciones.json`), la del directorio de grabaciones de H5
      (`testdata/evals/boe.legislacion-consolidada/`) o la del de H4
      (`internal/source/boe/testdata/boe.legislacion-consolidada/`), ni la de un directorio que los contenga: `testdata/`,
      `testdata/evals/` y `testdata/evals/boe.legislacion-consolidada/` (manifiesto y grabaciones de H5); `internal/`,
      `internal/source/`, `internal/source/boe/` e `internal/source/boe/testdata/` (grabaciones de H4); ni
      `internal/evals/` o `internal/evals/testdata/`, que contienen las sesiones sintéticas, material protegido por el
      guardián, y que declarados de golpe dejarían a una tarea cambiarlas sin nombrarlas. Desde la tarea que sigue a
      cada esquema, ninguna línea deja extraer `schemas/` como directorio ni la ruta de un esquema ya creado (sí la de
      un esquema nuevo, en su propia tarea `[datos]`). Una ruta completa de un fichero nuevo en otra carpeta de datos
      de prueba (`internal/skills/testdata/script/instalar.txtar`, las sesiones del paso 13) no extrae ninguno de esos
      directorios: la tubería toma la ruta entera;
   4. se comprueba sobre el texto entero de cada línea con la tubería de `workflow.yml` 615 (la de `rutas=` del paso
      que prepara la tarea), aplicada tal cual, antes de dar por buena la tarea.
4. **`misspell`** (V42, D20): en el paso 2 entran en `misspell.ignore-rules` exactamente `comando`, `comandos`,
   `defectos`, `legislativo` y `patrones`, cada una con el comentario de su motivo (tabla de D20), y ninguna más. Las
   otras veintiuna palabras que V42 encuentra en los artefactos y en spec.md —cuya redacción copian los comentarios que
   citan un FR— no se escriben **sueltas** en ningún fichero Go (comentario, cadena, nombre de subtest, etiqueta de
   campo ni nombre de fichero citado): las que en español llevan tilde (`activacion`, `constitucion`, `declaracion`,
   `evaluacion`, `preparacion`, `prescripcion`) se escriben con ella o dentro de un identificador camelCase
   (`DeclaracionDeKitlegal`); las que no la llevan (`candidatas`, `componentes`, `contradice`, `decisiones`,
   `definitivo`, `directorios`, `distribuye`, `informativo`, `procede`, `producto`, `programas`, `recorre`, `secretos`,
   `terminaron`, `variantes`) se sustituyen por otra palabra que `misspell` no marque o van dentro de un identificador:
   `recorre`, por `recorrer` o `recorrido`; `terminaron` —la palabra de FR-071 y de *Key Entities* para las invocaciones
   fuera de lo grabado, que documentan `juzgar.go` e `informe.go`—, por «sin terminar» o «que no acabaron»;
   `candidatas` (FR-005), por el singular `candidata` o por «posibles» (alternativas comprobadas en V42). Por la misma
   regla, ningún test escribe los nombres `04-lgt-prescripcion.yaml` ni
   `09-constitucion-articulo-140.yaml`: el conjunto se lee del directorio. Toda palabra nueva sin tilde se comprueba
   antes contra `words.go` y `words_us.go` de `misspell` v0.8.0.
5. **Sin `//nolint`**, sin tests saltados, sin errores silenciados; los tests que ejecutan `make` o `go` escriben programa
   y argumentos como constantes y pasan lo variable por `cmd.Env` y `cmd.Dir`; lectura y escritura de un mismo fichero por
   auxiliares distintos; escrituras con `0o600` salvo los guiones, que conservan su modo del repositorio.
6. **Sin cambios del kernel ni de la fuente**: ninguna tarea declara ficheros de `internal/core`, `internal/cli`,
   `internal/httpx`, `internal/cache`, `internal/source/boe` ni de `internal/app` distintos de `internal/app/skills_test.go`;
   `TestDependenciasDelBinario` y `TestArquitectura` sin cambios y en verde.
7. **La skill**: `SKILL.md` cumple el contrato de la skill; su región de comandos solo la escribe `make skills-sync`;
   ninguna mención a evals, al job, a `KITLEGAL_CACHE_DIR` ni a usar siempre `--offline` (FR-077).
8. **Documentación alineada con el `Makefile`** en la misma rama (D19) y `docs/PENDIENTES.md` sin las tres entradas «En H5».
9. **Evidencias** fechadas por commit en `gates/pr-h5.md` (salida de quickstart §11, supuesto S8 pendiente, dependencias
   del Complexity Tracking), `gates/prueba-de-red.md`, `gates/evals-cierre.md` y `gates/aceptacion.md`.
10. **Umbrales**: global ≥ 70 % e `internal/core` ≥ 85 %; nunca se rebaja un umbral.
11. **`rtk`**: toda orden cuya salida se filtre o compare se ejecuta con `rtk proxy`, cada etapa de la tubería incluida;
    las formas del quickstart se ejecutan tal cual.
12. **Sin pasar en vacío**: `TestSkillsDelRepositorio/skills` exige desde el paso 10 que exista `skills/boe-legislacion`;
    `TestEvalsDelRepositorio` exige los doce ficheros desde el paso 7; `TestIdentificadoresDeLasNormas` exige al menos las
    diez normas.
13. **Guiones ejecutables**: `scripts/skills-sync.sh`, `scripts/instalar-skills.sh`, `scripts/grabar-evals.sh` y
    `scripts/evals.sh` se versionan con permiso de ejecución (modo `100755`), como los guiones existentes; sin él,
    `make skills-sync`, `make install` y `make evals` fallan en un clon.

## Comprobación contra la rúbrica del juez (`juez_plan`, criterios a-j)

| Criterio | Dónde se cumple |
|---|---|
| a. constitution_check | Un ítem por principio (I-IX) y por regla de dependencia (las cinco de §IV, la del grafo, la de ejemplos del ADR 0010, la del espacio reservado del ADR 0006 y la de módulos del binario), con cómo se cumple, cómo se vigila y veredicto; gates por capa; reglas del modo desatendido; re-evaluación tras el diseño. El principio V figura como «cumple con justificación» y remite a *Complexity Tracking* |
| b. dependencias | Única dependencia directa nueva `go.yaml.in/yaml/v3`, justificada; herramientas del job (Claude Code, `strace`, runner, secreto, etiquetas) en *Complexity Tracking* |
| c. reglas_dependencia | Sin `package main` nuevo; `net/http` solo en `httpx` (las herramientas usan `Replay`); SQLite solo en `cache`; stdout solo por el presentador o por guiones de shell; el binario no enlaza las herramientas |
| d. errores_exit_codes | Códigos del binario sin cambios (0, 2, 3, 4, 5; verificados 4 y 5 con proxy que rechaza, V41); herramientas con errores que nombran el objeto del defecto; guiones con 0 y 1, `make` con 2 (V40); informe con veredicto |
| e. tests_primero | Inventario con nombres fijos, subtests incluidos donde quickstart §6 y §7 los esperan (`TestJuzgar`, `TestLeerSesion`, `TestLeerTrazas`, `TestInforme`, `TestPrepararDirectorioDeSesion`, `TestLeerConjunto`, con los mismos nombres en los contratos de evals §1 y §6 y del job §9); `PrepararSesion` y `EscribirInforme` con firma (contrato del job §3.2 y §3.3) y con test sin etiqueta en `make ci`; `LeerConjunto` (lectura y reparto entre evals bien formadas y ficheros mal formados) y `ComprobarConjuntoDeBoeLegislacion` (reglas de data-model §6.3) con firma y test propio (contrato de evals §1 y §2), y `PrepararSesion` y `EscribirInforme` citan esa firma; la atribución de hilos y conexiones de la traza con un caso por regla (`hilo-por-clone`, `hilo-por-clone3`, `hilo-de-un-hilo`, `connect-fuera-de-la-invocacion`, `fichero-sin-origen`; data-model §9) y las conexiones de cada invocación en el informe; los campos de cabecera del informe (`modelo`, `modelos_de_sesion`, `versiones_de_claude_code`, `commit` y `sin_python`; data-model §10.3) con su valor exacto en `informe.json` y en `informe.md`, y tomados cada uno de su origen, con un modelo de las sesiones distinto del constante (`TestInforme/aprobado`, `/sesion-ilegible`; contrato del job §9); el veredicto, que exige una sesión por cada eval bien formada y alguna eval que juzgar (`TestInforme/eval-sin-sesion`, `TestEscribirInformeSinSusEntradas/evals-y-sesiones-vacios`), cada fichero del guion que falta con su motivo (`TestInforme/sin-eval-txt`, `/eval-desconocida`, `/sin-pregunta-txt`), la rama de la traza del informe (`TestInforme/traza-ilegible`) y la traza sin ficheros o inexistente (`TestLeerTrazasSinFicheros/vacio`, `/inexistente`; data-model §9, regla 1); la respuesta y el fin de cada sesión en el informe, de los que sale la aceptación (`TestInforme/aprobado`, `/sesion-sin-terminar`; data-model §10.2); que `--describe` y `--dry-run` no satisfacen ningún comando esperado y el reparto de las invocaciones fallidas (`TestJuzgar/describe-y-dry-run-no-satisfacen`, `/otra-fallida`; data-model §6.1 y §10.2); `TestPrepararYComprobar` sobre las grabaciones de H4 y una eval sintética, con `completo` (contrato de evals §5.2) y los contratos de error de `EscribirInforme` y de `PrepararSesion` con su test (`TestEscribirInformeSinSusEntradas/sin-python-inexistente`, `/sesiones-inexistente`, `/evals-inexistente`, `/destino-es-un-fichero`; `TestPrepararDirectorioDeSesion/eval-mal-formada`, `/con-faltas`); las líneas de señal y la sesión cortada por el tope, con `LeerTrazas(dir, cortada)` y el corte que da `LeerSesion`, con un caso por regla (`lineas-de-senal`, `cortada-por-el-tope`, `sin-linea-final-sin-corte`, `cortada-con-llamada-interrumpida`, `llamada-interrumpida-sin-corte`, `llamada-interrumpida-antes-del-final`; `TestInforme/sesion-sin-terminar` y `/sesion-cortada-con-invocaciones`; `TestJuzgar/bloque-leido-sin-codigo`; data-model §9, regla 6); `LeerManifiesto` con firma, reglas y `TestManifiestoDeGrabaciones`, sintético en el paso 3 y sobre el manifiesto real en el paso 6 (contrato de evals §3.1); árbol exacto de las sesiones sintéticas, con el nivel de caso de cada test (contrato del job §9.1); evals antes de `SKILL.md` y la generación; guiones e2e de instalación en `[datos]` antes del código; sesiones sintéticas antes de la comparación; 22 controles con demostración |
| f. alcance | Solo lo del spec; sin cambios en kernel ni fuente; lo no enumerado por el spec está en *Complexity Tracking* con motivo |
| g. sin_atajos | Ningún `nolint`, `t.Skip`, TODO ni error silenciado previstos; ninguna exclusión de lint; herramientas sin excepciones de `forbidigo` |
| h. mejor_alternativa | D1-D21 con alternativas rechazadas y motivo, en particular D2, D5, D8 (`go.yaml.in/yaml/v3` frente a `gopkg.in/yaml.v3` y a `go.yaml.in/yaml/v4`, con V58), D12, D16 y D17 (búsqueda de Python en todo el sistema de ficheros frente al `PATH` y a una lista cerrada de directorios o de instalaciones; opciones de shell fijadas por el propio paso frente a las de la plataforma; paquetes listados sin patrón y purgados con ficheros en disco frente a la consulta por patrón y a purgar solo los instalados) |
| i. afirmaciones_verificadas | Tabla V1-V58 con fichero, línea u orden de cada comprobación (Claude Code 2.1.270, `gh` con los campos de `gh run list --json`, validador del estándar, Go y su biblioteca, módulos, git, `make`, `misspell` con el replacer de golangci-lint sobre los artefactos y spec.md, el workflow, el binario del repositorio, las sondas locales de la extracción del informe y de la comprobación de ficheros cambiados, la forma del `if` con `/dev/tcp` bajo `set -euo pipefail` (V50), el runtime de Go que crea sus hilos con `clone` (V51), y los códigos de `timeout` y de `strace` y el formato de sus trazas con el binario, en un contenedor local de Ubuntu 24.04 sin red (V52, V53), lo que deja en la traza un proceso muerto por una señal y un corte del tope, con y sin `-e signal=none` (V54), cómo leen `encoding/json` y `encoding/json/v2` una clave repetida (V55), y el paso que retira Python y la comprobación 3 del job en su forma literal —búsqueda como root en todo el sistema de ficheros salvo `/proc` y `/sys`, con intérpretes, enlaces y bibliotecas bajo `/opt`, `/usr/share`, `/usr/local` y `/dev/shm`, sin root, en solo lectura y con las opciones que fija el propio paso (`set -euo pipefail`) frente al mismo paso sin ellas— en contenedores de Ubuntu 24.04 sin red (V56), y la consulta de `dpkg-query` —formato, estados que lista, códigos con y sin patrón— y la purga con `apt-get` de paquetes `python*`, `libpython*` y `pypy*` instalados, `Multi-Arch: same`, retenidos, a medias y con solo su configuración (V57), y `go.yaml.in/yaml/v4` —en el grafo y en el binario, lo que su README dice de v3 y de v4, que solo tiene versiones candidatas y que su API cambia entre ellas— (V58)); lo que depende de GitHub (también los eventos de etiqueta, el `workflowName` y las ejecuciones que leen las órdenes de quickstart §12, que no usan `--workflow`), del runner (el formato de la traza en x86_64 con Claude Code, S4; qué intérpretes de Python trae y que la búsqueda y la retirada terminan sin romper el job, S7; y los códigos de la sesión con Claude Code, S10) o de la API, en D22 (S1-S12) |
| j. quickstart_ejecutable | Escenarios con órdenes, rutas, tests y datos reales; defectos provocados solo en un clon desechable en `/tmp`; instalación con `HOME` y `GOBIN` temporales; cada orden en una forma que la sesión desatendida ejecuta; limpieza final; los escenarios de plataforma declaran lo que escriben, eligen la ejecución por `workflowName`, sacan el informe entero entre las marcas que fija el contrato del job §3.3 (V48) y registran como evidencia de SC-003 la misma salida que comprueban (V49); las entradas del informe nombran sesión y eval |
