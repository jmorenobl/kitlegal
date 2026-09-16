# Implementation Plan: H5.1 · Avisos de vigencia en las evals

**Branch**: `007-h5-1-avisos-de-vigencia` | **Date**: 2026-09-16 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/007-h5-1-avisos-de-vigencia/spec.md`

## Summary

H5.1 protege la skill `boe-legislacion` (constitución, principio VIII): hace medible que la respuesta **traslada** los
avisos de vigencia que emite el binario. El mecanismo lo fija el hito: el aviso se compara **por su forma fija** —`⚠`, la
etiqueta y dos puntos—, como la cita por su identificador, sin leer la redacción libre ni usar ningún modelo.

Enfoque técnico, de fuera adentro al planificar y de dentro afuera al implementar:

1. **Etiquetas en el binario** (research D1). `internal/source/boe` exporta `EtiquetasDeAviso() map[string]string` y
   compone cada frase con su etiqueta, sin cambiar ningún byte de salida.
2. **Forma fija** (D2, D3). `internal/evals.ExtraerAvisos` reconoce la forma de cada código con una expresión construida
   desde las etiquetas exportadas, tolerante solo con el selector de presentación del emoji, el énfasis `*`/`_`, los
   blancos horizontales y las mayúsculas, exacta en las palabras de la etiqueta y dentro de una línea.
3. **Formato** (D4). `schemas/eval.yaml.json` gana `avisos` (enumerado de `CodigosDeAviso()`, solo con `activa: true`,
   con el trato de `citas` para la lista vacía y los repetidos) y `Eval` gana `Avisos`.
4. **Juicio e informe** (D3, D9, D10). `Juzgar` reparte los avisos esperados en encontrados y ausentes, con el motivo
   `aviso ausente: <código>` detrás de los de las citas, y un ausente impide pasar; el informe publica
   `avisos_encontrados` y `avisos_ausentes` y dos columnas en la tabla de sesiones.
5. **Comprobaciones mecánicas** (D5 a D8), dentro de `make skills-check` sin tocar el `Makefile`: el enumerado del esquema
   coincide con `CodigosDeAviso()`, `SKILL.md` lleva la forma fija completa de cada código, y el código de
   `internal/evals` no copia ninguna etiqueta.
6. **La eval de la norma derogada** (D11 a D14). `BOE-A-1992-26318` entra en el manifiesto de grabación, en
   `data/normas.yaml` y en `evals/boe-legislacion/18-lrjpac-norma-derogada.yaml`, informativa; una persona graba en una
   pausa el único recurso que falta, su índice, con un arnés que ya no vuelve a pedir lo grabado; y la verificación de
   identificadores corrige una premisa que la entrada nueva invalidaba.
7. **`SKILL.md`** (D15). La regla 3, el paso 5 y «Cómo se cita» fijan la forma, con las tres formas completas y, como
   ejemplo, la frase literal que el binario emite para `derogada`, que no nombra ninguna norma; la eval se confirma
   antes.
8. **Aceptación** (D17). La primera ejecución del job de evals de la rama, la que dispara la apertura de la propuesta de
   cambio, se lee con `jq`: veredicto aprobado, `red` vacío, la eval 18 con su tasa sobre 3 sesiones y el reparto de sus
   dos avisos en cada una.

Todo lo afirmado sobre herramientas y dependencias está verificado en local o declarado como supuesto (research, tabla
de verificación V1-V33 y §S).

## Technical Context

**Language/Version**: Go 1.27 sin cambios (`go 1.27.0` + `toolchain go1.27.1`, exportado por el `Makefile`); JSON Schema
2020-12 y YAML para el formato de eval y los datos; Markdown para la skill; Bash solo en las órdenes del quickstart.

**Primary Dependencies**: las de `go.mod`, sin ninguna nueva. `santhosh-tekuri/jsonschema/v6` (constitución §V), que
`internal/evals` ya importa, para leer el esquema compilado; `go.yaml.in/yaml/v3` a través de `internal/skills`, como en
H5. Biblioteca estándar: `regexp`, `sync`, `errors`, `slices`, `strings`, `os` (`CopyFS` en un test). Fuera de Go, sin
cambios: el job de evals de H5 con ADR 0016 (Claude Code, `strace`, `ubuntu-24.04`) y, en la plataforma, `gh` y `jq`.

**Storage**: ninguno nuevo. Una grabación nueva en `testdata/evals/boe.legislacion-consolidada/`.

**Testing**: `make test` (unitarios con `-race`), `make skills-check` (con los subtests nuevos de
`TestEvalsDelRepositorio`), `make schema-check` y `make test-e2e` sin cambios como red de FR-012. Todo sin red: sesiones
sintéticas copiadas a `t.TempDir()`, esquemas sintéticos y `httpx.Replay` sobre lo grabado. Red de la fuente solo en
`scripts/grabar-evals.sh`, que ejecuta una persona en la pausa; red del modelo solo en el job de evals, fuera de
`make ci`.

**Target Platform**: sin cambios para el binario (sin cgo, `-trimpath`). Desarrollo en darwin/arm64, `make ci` en
`ubuntu-latest`, job de evals en `ubuntu-24.04`.

**Project Type**: CLI multicall + skill del estándar Agent Skills + herramientas de desarrollo (`internal/evals`), que el
binario no enlaza.

**Performance Goals**: `ExtraerAvisos` compila sus tres expresiones una sola vez; `make skills-check` sigue en segundos.
El job pasa de 87 a 90 sesiones (research V21), dentro del presupuesto de ADR 0016 (supuesto S4).

**Constraints**: sin ningún juicio de modelo ni lectura de la redacción libre (FR-033); ningún cambio en la salida del
applet, en `schemas/norma.json` ni en `schemas/bloque.json` (FR-012); ninguna etiqueta copiada en el código de
`internal/evals` (FR-011); `SKILL.md` < 300 líneas y solo con cambios en la forma de trasladar los avisos (FR-004,
FR-005); ningún cambio en `internal/core`, `internal/app`, `internal/cli`, `internal/httpx`, `internal/cache`,
`docs/SOURCES.md`, `scripts/verify-sources.sh`, `docs/USO.md` ni los guiones `testscript`; la grabación la hace una
persona en una pausa (FR-055).

**Scale/Scope**: 4 ficheros de código Go cambiados (`avisos.go` de `internal/source/boe`, y `formato.go`, `juzgar.go` e
`informe.go` de `internal/evals`) y 1 nuevo (`internal/evals/avisos.go`), con sus tests; 3 ficheros de test ajustados
sin fichero de código propio (`conjunto_test.go`, `grabaciones_test.go`, `grabacion_test.go`); 1 esquema; 1 entrada de manifiesto y 1
grabación; 1 norma y su referencia regenerada; 1 eval (18 en el conjunto); 18 líneas más en `SKILL.md` (de 181 a 199,
research V33); 3 documentos.

## Constitution Check

*GATE: debe pasar antes de la fase 0 y volver a evaluarse tras la fase 1.*

### Principios

| # | Principio | Cómo lo cumple H5.1 | Veredicto |
|---|---|---|---|
| **I** | Fuentes públicas y frontera humana | Ninguna fuente nueva ni código de red. La única petición nueva al BOE —el índice de `BOE-A-1992-26318`— la hace una persona en la pausa de la tarea `[datos]` del manifiesto, con `httpx` (identificación, `robots.txt`, ritmo de H4); el arnés deja de pedir lo que ya está grabado, así que, además del índice y del `robots.txt` que `httpx` pide antes de él, la grabación no carga la fuente con treinta peticiones innecesarias: las nueve búsquedas y las veintiuna consultas de norma que ya grabó H5 (research D11, V12). Ningún POST, ninguna identidad. El job sigue sin red de ninguna fuente y lo comprueba (`red` vacío, FR-073). | ✅ Cumple |
| **II** | Nada sin cita ni fuente | La forma fija hace visible en la respuesta lo que el sobre ya dice de la vigencia, para que no se presente como vigente una norma derogada (regla 3). La eval 18 exige la cita `BOE-A-1992-26318` `a42` por identificador y los avisos por su forma, sin modelo. La norma entra en `data/normas.yaml` con el identificador, el título y el rango de la búsqueda grabada; el ejemplo de `SKILL.md` es, literal, la frase que el binario emite para `derogada` (`⚠ NORMA DEROGADA: esta norma ha sido derogada.`, contrato de la forma fija §3), así que la skill no enseña ninguna afirmación que el sobre no traiga (research D15). | ✅ Cumple |
| **III** | Tests primero y offline | Cada tarea escribe su test antes del código (inventario abajo). Todo es offline: sesiones sintéticas copiadas a un temporal, esquemas sintéticos y la caché preparada desde lo grabado (`TestEvalsDelRepositorio/grabado`). La eval 18 se confirma antes del primer cambio de `SKILL.md` (FR-052). No hay guion `testscript` nuevo porque el comportamiento visible del binario no cambia (spec, *Fuera de alcance*); el test de extremo a extremo de la entrega es la eval 18 —en el job con la skill instalada y, sin red, en `make ci`— y los guiones existentes siguen como red de FR-012. Umbrales de cobertura intactos; `internal/core` no se toca. Desviación del «empieza por el e2e», justificada en *Complexity Tracking*. | ✅ Cumple con justificación |
| **IV** | Arquitectura hexagonal con reglas ejecutables | Ni dominio ni kernel cambian; el adaptador `boe` exporta una función pura. Las comprobaciones devuelven errores que nombran el código; ningún código de salida cambia; ningún `panic` en rutas de usuario (`regexp.MustCompile` solo con piezas fijas y palabras pasadas por `QuoteMeta`, en un paquete que el binario no enlaza, como `formaDeCita`; research D2). Reglas de dependencia, abajo. | ✅ Cumple |
| **V** | Simplicidad y dependencias fijadas | Solo lo del spec. Ninguna dependencia nueva, ninguna exclusión de lint, ningún objetivo de `make` nuevo. | ✅ Cumple |
| **VI** | Un binario, convenciones de agente | Ni verbos, ni banderas, ni `--describe` cambian; la región generada de `SKILL.md` queda intacta y la skill sigue usando `--json`. | ✅ Cumple |
| **VII** | Grafo y privacidad | Sin grafo. Ni la eval, ni la norma, ni el manifiesto, ni el informe llevan datos personales. | ✅ Cumple |
| **VIII** | Skills primero; el binario es la herramienta | Protege la skill existente con una eval y una regla escrita medible. A Go solo va lo determinista: la etiqueta, la comparación de la forma y las comprobaciones. `SKILL.md` sigue por debajo de 300 líneas, `references/` se regenera desde `data/normas.yaml` y `scripts/` no cambia. | ✅ Cumple |
| **IX** | Genericidad territorial, validación local | Ningún caso territorial: la Ley 30/1992 es estatal, y ni `SKILL.md`, ni la eval, ni los datos nombran un municipio o una comunidad. Sin dimensión territorial. | ✅ Cumple |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución §IV)

| Regla | Situación en H5.1 | Cómo se hace cumplir | Veredicto |
|---|---|---|---|
| `internal/core/**` no importa `internal/{source,httpx,cache,store,graph,render,cli,app}` ni entrada y salida | `internal/core` no se toca | `depguard` lista `core` + `TestArquitectura` R1 (sin cambios) | ✅ Cumple |
| Solo `internal/httpx` importa `net/http` | `internal/evals/avisos.go` importa `internal/source/boe`, `jsonschema` y la biblioteca estándar, sin `net/http`; `boe` no gana importaciones | `depguard` lista `red` + `TestArquitectura` R2 | ✅ Cumple |
| Solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | Ningún fichero nuevo o cambiado los importa | `depguard` lista `sql` + `TestArquitectura` R3 | ✅ Cumple |
| Solo `internal/cli` y `cmd/` llaman a `os.Exit` | Ningún `package main` ni salida de proceso nuevos | `forbidigo` `^os\.Exit$` | ✅ Cumple |
| Solo `internal/render` escribe en stdout; logs con `slog` a stderr | Las funciones nuevas devuelven valores y errores; el informe escribe ficheros, como en H5 | `forbidigo` (`fmt.Print…`, `os.Stdout`, `os.Stderr`) | ✅ Cumple |
| `internal/graph` no importa `internal/source/*` ni `internal/render` | `internal/graph` no existe | — (sin objeto) | ✅ Cumple |
| Los applets de ejemplo no se enlazan en el binario distribuido (ADR 0010) | Nada importa `ejemplo` | `depguard` lista `ejemplo` + `TestElBinarioNoEnlazaLosEjemplos` | ✅ Cumple |
| Ningún adaptador firma en el espacio reservado (ADR 0006) | Ningún adaptador nuevo ni cambio de `fuente` | `TestLasFuentesNoFirmanComoKitlegal` | ✅ Cumple |
| El binario no enlaza módulos no justificados (FR-124 de H4) | `EtiquetasDeAviso` no importa nada; `internal/evals` sigue fuera del binario | `TestDependenciasDelBinario` (sin cambios) | ✅ Cumple |

### Gates (constitución, «Gates»)

- **Capa 1 (mecánica)** que H5.1 añade: tolerancia de la forma fija (`TestExtraerAvisos`); formato de `avisos`
  (`TestLeerEval`); esquema ↔ `CodigosDeAviso()` y `SKILL.md` ↔ forma fija, fallando por código (FR-013, FR-014); sin
  copias de etiquetas (`TestEtiquetasSoloDesdeBoe`); juicio de avisos «por identificador», como las citas (`TestJuzgar`);
  informe (`TestInformeConAvisos`); y, con los controles de H5 sobre datos nuevos, deriva de `references/`, conjunto de
  evals, identificadores contra la búsqueda grabada y lo grabado suficiente para la eval 18. Guardián de diff con
  `[datos]` para `schemas/` y `testdata/`.
- **Capa 2 (jueces)**: los del workflow (plan, tareas, revisión final); el protocolo de `SKILL.md` lo juzgan los dos
  jueces de la revisión final. La aceptación no la juzga ningún modelo: es un programa de `jq` sobre el informe.
- **Capa 3 (humano)**, pausas previstas: `schemas/eval.yaml.json` (esquema existente modificado); el manifiesto
  `testdata/evals/grabaciones.json` (fichero existente modificado), en cuya pausa la persona graba el índice, restaura
  `robots.txt` y confirma la grabación (contrato de la eval y la grabación §4); los prerrequisitos de plataforma
  (secreto y etiqueta); la fusión. `docs/SOURCES.md` no cambia.

**Reglas del modo desatendido**: el ejecutor nunca usa `KITLEGAL_RECORD`, `scripts/grabar-evals.sh`, `make evals` ni
`make verify-sources`; ninguna tarea `[datos]` toca código; el ejecutor arregla el código, nunca un fixture. Los dos
ajustes de ficheros de test (`grabaciones_test.go`, `grabacion_test.go`) son cambios planificados de la premisa de una
prueba y del arnés de grabación, en una tarea propia y antes del manifiesto, no correcciones para poner en verde un test
rojo (research D11, D12).

**Veredicto del gate: PASA.** La única desviación (principio III, e2e primero) está justificada en *Complexity
Tracking*, y nada afecta a alcance, frontera humana, privacidad, términos de uso ni a una decisión cerrada.

### Re-evaluación tras la fase 1 (diseño)

- **Arnés de grabación con la unión de conjuntos** (§I, FR-054, FR-055): reduce a una la petición nueva a la fuente y
  mantiene la grabación en manos de una persona.
- **Premisa de `TestIdentificadoresDeLasNormas`** (§III): la prueba sigue fallando donde debe; comprobado con y sin la
  entrada nueva (research V14).
- **Test del informe sobre una copia temporal** (§III, guardián de `testdata/`): ningún material nuevo en territorio de
  fixtures y ninguna tarea en rojo entre dos pasos (research D10).
- **Aceptación en la plataforma** (FR-070): depende de supuestos de plataforma declarados (research §S), que la tarea
  `[plataforma]` comprueba y registra.

**Veredicto tras el diseño: PASA**, sin ninguna violación no justificada.

## Project Structure

### Documentation (this feature)

```text
specs/007-h5-1-avisos-de-vigencia/
├── plan.md              # Este fichero
├── research.md          # Fase 0: V1-V33, D1-D19, supuestos S1-S5
├── data-model.md        # Fase 1: aviso, forma fija, eval, resultado, informe, comprobaciones, norma, manifiesto, eval 18, evidencia
├── quickstart.md        # Fase 1: guía de validación ejecutable (escenarios 1-11)
├── contracts/
│   ├── forma-fija-de-los-avisos.md          # gramática, EtiquetasDeAviso, ExtraerAvisos, FR-014, texto de SKILL.md
│   ├── formato-juicio-e-informe.md          # esquema, Eval.Avisos, FR-013, Juzgar, informe
│   ├── eval-norma-derogada-y-grabacion.md   # arnés y premisa, manifiesto, pausa, norma, eval 18
│   └── ejecucion-de-aceptacion.md           # publicación, identificación, lectura, comprobación, repetición, evidencia
├── spec.md
├── checklists/
├── gates/               # veredictos; pr-h5.1.md y evals-aceptacion.md llegan en la implementación
└── tasks.md             # Fase 2 (/speckit-tasks; no lo crea este comando)
```

### Source Code (repository root)

Estructura de `docs/ROADMAP.md` §2 y `CLAUDE.md`. **NUEVO** = lo crea H5.1; ← = cambia.

```text
skills/boe-legislacion/
├── SKILL.md                                ← regla 3, paso 5 y «Cómo se cita»                       (D15)
└── references/normas.md                    ← regenerado con la norma nueva                          (D13)
data/normas.yaml                            ← BOE-A-1992-26318                                        (D13)
schemas/eval.yaml.json                      ← [datos] avisos, $defs/codigo-de-aviso, rama else        (D4)
evals/boe-legislacion/
└── 18-lrjpac-norma-derogada.yaml           **NUEVO** informativa, con avisos                         (D14)
testdata/evals/
├── grabaciones.json                        ← [datos] entrada de la Ley 30/1992                       (D11)
└── boe.legislacion-consolidada/
    └── GET_…_BOE-A-1992-26318_texto_indice.json   **NUEVO** lo graba una persona en la pausa         (D11)
internal/
├── source/boe/
│   ├── avisos.go                           ← etiquetas, frases compuestas, EtiquetasDeAviso           (D1)
│   └── avisos_test.go                      ← TestEtiquetasDeAviso
└── evals/
    ├── avisos.go                           **NUEVO** ExtraerAvisos, ComprobarFormasDeAviso, ComprobarCodigosDeAviso (D2, D3, D5, D7)
    ├── avisos_test.go                      **NUEVO** TestExtraerAvisos, TestComprobarFormasDeAviso, TestComprobarCodigosDeAviso, TestEtiquetasSoloDesdeBoe (D8)
    ├── formato.go / formato_test.go        ← Eval.Avisos; casos de TestLeerEval                       (D4)
    ├── juzgar.go / juzgar_test.go          ← AvisosEncontrados, AvisosAusentes, motivo, Pasa; casos de TestJuzgar (D3)
    ├── informe.go / informe_test.go        ← dos columnas; TestInformeConAvisos y TestInforme/aprobado (D9, D10)
    ├── conjunto_test.go                    ← subtests avisos-del-esquema y avisos-de-la-skill        (D6)
    ├── grabaciones_test.go                 ← premisa de otraNormaDeLaBusqueda                         (D12)
    └── grabacion_test.go                   ← //go:build grabacion · siembra desde la unión de grabaciones (D11)
README.md, CONTRIBUTING.md, CHANGELOG.md    ← avisos en el formato común de eval; entrada de H5.1      (D18)
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquetes nuevos: la forma fija vive en
`internal/evals` junto a la cita (`citas.go`), la etiqueta en el adaptador que la emite (`internal/source/boe`), y los
datos, las evals, los esquemas y la skill en sus directorios de la raíz. No se tocan `internal/core`, `internal/app`,
`internal/cli`, `internal/httpx`, `internal/cache`, `cmd/`, `scripts/`, `.github/`, el `Makefile` ni `.golangci.yml`.

## Controles mecánicos que este hito añade o toca

Según la sección «Gates» de la constitución. «Demostración» dice qué falla si el control se retira o se viola.

| # | Control | Test / forma | Orden | ¿En `make ci`? | Demostración |
|---|---|---|---|---|---|
| 1 | Etiquetas exportadas: exactamente tres, una por código, prefijo de su frase | `TestEtiquetasDeAviso` | `test` | Sí | Quitar una etiqueta del mapa, cambiar una letra o componer mal una frase → falla |
| 2 | La salida del applet no cambia (FR-012) | `TestAvisosDe` (frases literales), golden de `internal/source/boe`, `TestEsquemasPublicados`, guiones `testscript` de `internal/app` | `test` (los guiones también con `test-e2e`), `schema-check` | Sí | Una frase distinta → falla `TestAvisosDe` y los golden (quickstart 3) |
| 3 | Forma fija con su tolerancia | `TestExtraerAvisos` | `test` | Sí | Tolerar otra redacción, la negación, una palabra de más o un salto dentro, o no tolerar el selector o el énfasis → falla (quickstart 5) |
| 4 | `avisos` en el formato | `TestLeerEval` | `test` | Sí | Aceptar un código desconocido o `avisos` en una no activación → falla (quickstart 6) |
| 5 | Esquema ↔ `CodigosDeAviso()` (FR-013) | `TestComprobarCodigosDeAviso`; `TestEvalsDelRepositorio/avisos-del-esquema` | `test`, `skills-check` | Sí | Un código de menos o de más en el enumerado → falla nombrándolo (quickstart 4) |
| 6 | `SKILL.md` ↔ forma fija de cada código (FR-014) | `TestComprobarFormasDeAviso`; `TestEvalsDelRepositorio/avisos-de-la-skill` | `test`, `skills-check` | Sí | Quitar la etiqueta, la marca o los dos puntos de una forma → falla nombrando el código (quickstart 4) |
| 7 | Sin copias de etiquetas en `internal/evals` (FR-011) | `TestEtiquetasSoloDesdeBoe` | `test` | Sí | Escribir `NORMA DEROGADA` en un fichero de código del paquete → falla (quickstart 9) |
| 8 | Juicio de avisos | `TestJuzgar` | `test` | Sí | Dar por trasladado un aviso con otra redacción o negado, o dejar pasar una eval con un aviso ausente → falla (quickstart 5) |
| 9 | Informe con avisos | `TestInformeConAvisos`, `TestInforme/aprobado` | `test` | Sí | Sin las claves, con `null` en lugar de `[]`, sin las columnas o con el motivo antes de los de las citas → falla (quickstart 7) |
| 10 | Conjunto de evals con la eval 18 | `TestEvalsDelRepositorio/formato`, `/conjunto`, `/normas-conocidas` | `skills-check` | Sí | La eval 18 sin `informativa` rompe «positivas» (11 que deciden); sin la norma en la tabla, «normas conocidas» |
| 11 | Lo grabado basta para la eval 18 (FR-057) | `TestEvalsDelRepositorio/grabado` | `skills-check` | Sí | Sin el índice grabado → falla nombrando la eval y el índice (quickstart 8) |
| 12 | Identificadores contra la búsqueda grabada, con la premisa corregida | `TestIdentificadoresDeLasNormas` | `skills-check` | Sí | Un título distinto del grabado → falla nombrando la norma; los cuatro negativos siguen fallando donde deben |
| 13 | `references/` sin deriva y `SKILL.md` válido | `TestSkillsDelRepositorio`, `TestNormasDelRepositorio` | `skills-check` | Sí | La norma en `data/normas.yaml` sin regenerar, o una palabra prohibida en `SKILL.md` → falla |
| 14 | Manifiesto y grabaciones | `TestManifiestoDeGrabaciones`, `TestGrabacionesSinSolape` | `test` | Sí | Un prefijo repetido o un índice con nombre de una grabación de H4 → falla |
| 15 | Lint del arnés con etiqueta | `run.build-tags` incluye `grabacion` (sin cambios) | `lint` | Sí | Un error de tipos en `grabacion_test.go` → `make lint` falla |
| 16 | R1-R5, binario, formato, `-race`, `govulncheck`, `gosec`, `gitleaks`, `go mod verify`, `tidy -diff` | los de H0-H5, sin exclusiones nuevas | varias | Sí | Los mismos de H5 |
| 17 | Aceptación con modelo | job `evals` en la propuesta de cambio + programa de `jq` (quickstart 11) | plataforma | **No** | Veredicto `fallo`, `red` no vacío o la eval 18 sin su tasa o su reparto → la tarea `[plataforma]` se detiene |

### Objetivos del `Makefile`

Ninguno cambia. Los subtests nuevos entran en `make skills-check` y `make ci` porque pertenecen a
`TestEvalsDelRepositorio` (research V27); `make skills-sync` regenera la referencia como en H5.

### Fixtures y datos protegidos

| Material | Tarea | Pausa |
|---|---|---|
| `schemas/eval.yaml.json` | `[datos]` (paso 3) | sí (esquema existente modificado) |
| `testdata/evals/grabaciones.json` y el índice grabado de `BOE-A-1992-26318` | `[datos]` del manifiesto (paso 8); graba una persona | sí (fichero existente modificado) |
| Grabaciones y golden de H4, grabaciones de H5 existentes, `internal/evals/testdata/` | ninguna tarea los cambia; `TestInformeConAvisos` trabaja sobre una copia en un temporal | — |

### Tests de contrato

| Contrato | Tests |
|---|---|
| [forma-fija-de-los-avisos](./contracts/forma-fija-de-los-avisos.md) | `TestEtiquetasDeAviso`, `TestExtraerAvisos`, `TestComprobarFormasDeAviso`, `TestEtiquetasSoloDesdeBoe`, `TestEvalsDelRepositorio/avisos-de-la-skill` |
| [formato-juicio-e-informe](./contracts/formato-juicio-e-informe.md) | `TestLeerEval`, `TestComprobarCodigosDeAviso`, `TestEvalsDelRepositorio/avisos-del-esquema`, `TestJuzgar`, `TestInformeConAvisos`, `TestInforme/aprobado` |
| [eval-norma-derogada-y-grabacion](./contracts/eval-norma-derogada-y-grabacion.md) | `TestManifiestoDeGrabaciones`, `TestGrabacionesSinSolape`, `TestIdentificadoresDeLasNormas`, `TestNormasDelRepositorio`, `TestSkillsDelRepositorio`, `TestEvalsDelRepositorio` (`formato`, `conjunto`, `normas-conocidas`, `grabado`) |
| [ejecucion-de-aceptacion](./contracts/ejecucion-de-aceptacion.md) | la ejecución del job en la plataforma y el programa de `jq` de quickstart §11.5 |

## Inventario de tests

Nombres fijados aquí para que `tasks.md` y `quickstart.md` los usen tal cual. Tablas con subtests con nombre y
`t.Parallel()`, como el resto del paquete; material nuevo solo en `t.TempDir()`.

**Test de extremo a extremo de la entrega.** H5.1 no cambia el comportamiento visible del binario, así que no añade ni
cambia ningún guion `testscript` (spec, *Fuera de alcance*); los de `internal/app/testdata/script/` (`make test-e2e`)
siguen en verde sin cambios como prueba de FR-012. Lo que describe la entrega de extremo a extremo es la **eval
`18-lrjpac-norma-derogada.yaml`**: el job la ejecuta con Claude Code, la skill instalada y el binario sobre la caché
preparada (aceptación, FR-070), y `make ci` ejecuta sin red cada una de sus consultas con `app.Main`
(`TestEvalsDelRepositorio/grabado`). `TestInformeConAvisos` recorre `EscribirInforme` de extremo a extremo sobre una
ejecución sintética con transcript y traza.

| Fichero | Tests |
|---|---|
| `internal/source/boe/avisos_test.go` | `TestEtiquetasDeAviso` (**nuevo**): `exactamente-tres`, `frases-con-su-forma`, `cada-llamada-su-mapa`. `TestAvisosDe` y `TestCodigosDeAviso` sin cambios |
| `internal/evals/avisos_test.go` (**nuevo**) | `TestExtraerAvisos` (subtests del contrato de la forma fija §4); `TestComprobarFormasDeAviso`: `completo`, `falta-la-etiqueta`, `falta-la-marca`, `faltan-los-dos-puntos`, `ninguna`; `TestComprobarCodigosDeAviso`: `exacto`, `falta-un-codigo`, `sobra-un-codigo`, `sin-enumerado`, `sin-avisos`; `TestEtiquetasSoloDesdeBoe` |
| `internal/evals/formato_test.go` | `TestLeerEval`, casos nuevos `avisos`, `aviso-desconocido`, `no-activa-con-avisos`, `avisos-vacio`, `citas-vacio`, `aviso-repetido`, `cita-repetida`, `informativa-con-avisos` |
| `internal/evals/juzgar_test.go` | `TestJuzgar`, casos nuevos `aviso-con-su-forma-fija`, `aviso-con-variantes-toleradas`, `aviso-ausente`, `aviso-con-otra-redaccion`, `aviso-negado`, `forma-fija-y-lo-contrario`, `aviso-no-esperado`, `avisos-en-el-orden-de-la-eval`, `aviso-repetido`, `cita-y-aviso-ausentes` |
| `internal/evals/informe_test.go` | `TestInformeConAvisos` (**nuevo**): `uno-encontrado-y-otro-ausente`, `aviso-detras-de-la-cita`; `TestInforme/aprobado` ampliado con `[]` en las dos claves |
| `internal/evals/conjunto_test.go` | `TestEvalsDelRepositorio`, subtests nuevos `avisos-del-esquema` (paso 4) y `avisos-de-la-skill` (paso 11); `formato`, `conjunto`, `normas-conocidas` y `grabado` sin cambios de código, sobre la eval 18 |
| `internal/evals/grabaciones_test.go` | `TestIdentificadoresDeLasNormas` con la premisa de research D12 en `otraNormaDeLaBusqueda`; `TestManifiestoDeGrabaciones` y `TestGrabacionesSinSolape` sin cambios |
| `internal/evals/grabacion_test.go` | `TestGrabarEvals` (etiqueta `grabacion`; solo lo ejecuta una persona en la pausa), sembrando desde la unión de grabaciones |

## Orden de implementación

Cada paso deja `make ci` en verde y lleva su test; los puntos delicados están comprobados en un clon (research V5, V14,
V15, V33). Las rutas de cada paso son las que su tarea declara.

| Paso | Qué | Rutas | Verde porque |
|---|---|---|---|
| 1 | Etiquetas en el binario | `internal/source/boe/avisos.go` (+ test) | Frases byte a byte iguales (V9) |
| 2 | Forma fija y comprobaciones, sobre entradas sintéticas | `internal/evals/avisos.go` (+ test) | Funciones nuevas con sus tests; nada las usa todavía |
| 3 | `[datos]` `avisos` en el esquema (pausa) | `schemas/eval.yaml.json` | Ninguna eval lleva `avisos` y `Decode` ignora la clave (V5) |
| 4 | `Eval.Avisos`, casos de formato y `avisos-del-esquema` | `internal/evals/formato.go` (+ test), `internal/evals/conjunto_test.go` | El esquema ya admite `avisos` |
| 5 | Reparto de avisos en `Juzgar` | `internal/evals/juzgar.go` (+ test) | Evals sin `avisos` dan el mismo resultado (FR-034) |
| 6 | Avisos en el informe | `internal/evals/informe.go` (+ test) | `TestInforme` no compara filas de la tabla de sesiones (V19) |
| 7 | Arnés con la unión de grabaciones y premisa de identificadores | `internal/evals/grabacion_test.go`, `internal/evals/grabaciones_test.go` | La premisa elige la misma norma sin la entrada nueva (V14); el arnés solo compila |
| 8 | `[datos]` entrada del manifiesto (pausa; la persona graba el índice) | `testdata/evals/grabaciones.json` | Con la premisa del paso 7, la entrada sola no rompe nada (V14) |
| 9 | Norma y referencia | `data/normas.yaml`, `skills/boe-legislacion/references/normas.md` | La entrada del manifiesto la resuelve con su título (V15) |
| 10 | Eval de la norma derogada | `evals/boe-legislacion/18-lrjpac-norma-derogada.yaml` | Formato, conjunto, normas conocidas y lo grabado se cumplen con el índice del paso 8 (V15) |
| 11 | Forma fija en `SKILL.md` y `avisos-de-la-skill` | `skills/boe-legislacion/SKILL.md`, `internal/evals/conjunto_test.go` | Texto literal del contrato §3: tres formas, 199 líneas, ninguna añadida de más de 120 caracteres, sin palabras prohibidas ni normas nombradas (V33) |
| 12 | Documentación | `CHANGELOG.md`, `README.md`, `CONTRIBUTING.md` | Solo documentación |
| 13 | Cierre sin tocar el árbol: `make ci`, quickstart 1-10 y cuerpo de la propuesta | `specs/007-h5-1-avisos-de-vigencia/gates/pr-h5.1.md` | No cambia nada fuera del directorio del hito |
| 14 | `[plataforma]` publicación y aceptación | `specs/007-h5-1-avisos-de-vigencia/gates/evals-aceptacion.md` | Solo publica, lee y registra |

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|---|---|---|
| Principio III, «cada hito empieza por el test e2e (`testscript`) que describe la entrega»: no hay guion `testscript` nuevo y el primer paso no es la eval | La entrega no cambia el comportamiento visible del binario (spec, *Fuera de alcance*), así que no hay guion que escribir; lo que la describe de extremo a extremo es la eval 18, y esa eval es un fichero mal formado mientras el formato no tenga `avisos` (research D16). Se escribe y se confirma antes que la regla de la skill que mide (FR-052, Definition of Done §1.10), que es lo que la regla del principio protege | *Un guion `testscript` sobre el binario*: no describiría nada nuevo, porque la salida no cambia, y los guiones existentes ya fijan esa salida. *La eval como primer paso*: deja `make ci` en rojo hasta que llegan el esquema y el formato, y cada tarea tiene que dejarlo en verde |

## Obligaciones que este plan traslada a `tasks.md`

1. **Orden y rebanadas.** Una tarea por paso de «Orden de implementación», en ese orden, cada una con su test y con las
   rutas de su fila; ninguna depende de una posterior.
2. **Tareas `[datos]`.** Solo los pasos 3 y 8, y solo con `schemas/eval.yaml.json` y `testdata/evals/grabaciones.json`
   respectivamente. Ninguna otra línea de tarea contiene `testdata/` o `schemas/` como ruta (el caso `aprobado` es «el
   caso aprobado de las ejecuciones sintéticas del informe», el esquema es «el esquema de eval»), ni el texto de las
   etiquetas `[datos]` o `[plataforma]` fuera de las tareas que las llevan.
3. **La pausa del paso 8.** La tarea escribe solo la entrada de contrato §2; en su pausa, una persona sigue el
   procedimiento del contrato de la eval y la grabación §4 (grabar, restaurar `robots.txt`, revisar y confirmar solo el
   índice). La tarea siguiente parte de ese commit.
4. **Eval antes que la skill.** Los pasos 10 y 11 son tareas y commits distintos, en ese orden (FR-052, SC-009).
5. **Texto literal.** `SKILL.md` recibe exactamente los textos del contrato de la forma fija §3; la eval, el del contrato
   de la eval §6; la norma y la entrada del manifiesto, los de §5 y §2.
6. **Sin etiquetas copiadas.** Ningún fichero de código de `internal/evals` escribe una etiqueta de aviso, tampoco en un
   comentario.
7. **Sin red ni grabación en tareas.** Ninguna tarea usa `KITLEGAL_RECORD`, `scripts/grabar-evals.sh`, `make evals` ni
   `make verify-sources`.
8. **Definition of Done.** Paso 12: `CHANGELOG.md` (*Unreleased*), `README.md` y `CONTRIBUTING.md` (§1.6; research D18).
   Esquema y grabación en sus tareas `[datos]`. Ni ADR (§1.7) ni `docs/SOURCES.md` (§1.8): el spec los deja fuera.
9. **Cierre (paso 13).** Ejecuta `make ci` y los escenarios 1 a 10 del quickstart tal cual, con sus prerrequisitos y su
   limpieza, y se detiene y lo anota ante cualquier resultado distinto del esperado; mide la cobertura global (≥ 70 %)
   con `go tool cover -func` sobre `coverage.out` de `make ci` y confirma que `internal/core` no cambia; y escribe
   `gates/pr-h5.1.md` con la plantilla del ritual (objetivo, alcance, controles añadidos, decisiones, pendientes),
   «Dependencias: ninguna nueva», las medidas fechadas por commit y, en pendientes, la regla de repetición de la
   aceptación (contrato de aceptación §6).
10. **Plataforma (paso 14).** La última tarea, sola y con `[plataforma]`: quickstart §11 tal cual (publicar, identificar
    la ejecución de apertura, leer, comprobar y registrar en `gates/evals-aceptacion.md`), y §11.6 solo si un commit
    posterior a la ejecución cambió algo fuera del directorio del hito. Nunca fusiona.

## Comprobación contra la rúbrica del juez (`juez_plan`, criterios a-j)

| Criterio | Dónde se cumple |
|---|---|
| a. `constitution_check` | Un ítem por principio I-IX y uno por regla de dependencia (las seis de `docs/ROADMAP.md` §2 y las tres del binario de H1-H4), con cómo se cumple y cómo se vigila; la desviación de III, justificada en *Complexity Tracking* |
| b. `dependencias` | Ninguna nueva (Technical Context, principio V) |
| c. `reglas_dependencia` | Tabla de reglas: `internal/evals/avisos.go` solo importa `boe`, `jsonschema` y la biblioteca estándar; nada en `core`, ni `net/http`, SQLite, `os.Exit` o stdout nuevos |
| d. `errores_exit_codes` | Ningún código de salida cambia; las comprobaciones devuelven errores que nombran el código (data-model §6); ningún `panic` en rutas de usuario (principio IV) |
| e. `tests_primero` | Inventario de tests, test de extremo a extremo nombrado (eval 18 y `TestInformeConAvisos`, guiones `testscript` sin cambios), controles mecánicos 1-17, fixtures protegidos y orden con el test en cada paso |
| f. `alcance` | Solo FR-001 a FR-073; los dos ajustes de test (D11, D12) son medios necesarios de FR-054 a FR-056 y de «grabar solo lo que necesita la eval nueva», no funcionalidad nueva |
| g. `sin_atajos` | Ningún `//nolint`, ninguna exclusión de lint, ningún test saltado, ningún TODO; el prototipo pasó el linter fijado (V10) |
| h. `mejor_alternativa` | research D1-D17, cada una con sus alternativas rechazadas y por qué |
| i. `afirmaciones_verificadas` | research, tabla V1-V33 con el fichero, la línea, la orden o el experimento; lo que depende de la plataforma o de la red de la fuente, en §S como supuesto no verificado |
| j. `quickstart_ejecutable` | Órdenes, rutas y datos reales; roturas solo en un clon desechable que los prerrequisitos crean y la limpieza borra; formas aceptadas sin aprobación (V32); los escenarios 4, 8 (búsquedas) y 9 ejecutados tal cual sobre un clon con prototipos, las búsquedas del 2 y las mutaciones de `SKILL.md` del 4 sobre el texto literal del contrato de la forma fija §3 (V33), y la extracción del informe (§11.4) sobre un registro real de H5 (V24, V30) |
