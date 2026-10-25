# Implementation Plan: H31 · Las skills siguen la especificación Agent Skills: guía de autoría y gates

**Branch**: `027-h31-las-skills-siguen` | **Date**: 2026-10-24 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/027-h31-las-skills-siguen/spec.md`

## Summary

Las tres skills del producto pasan a la forma de la especificación Agent Skills y de la guía de Anthropic (ADR 0046):
un cuerpo de `SKILL.md` de menos de 500 líneas y como mucho 11 000 caracteres con lo que toda respuesta necesita, el
detalle de cada clase de pregunta en referencias escritas a mano que el cuerpo enlaza y dice cuándo leer, y una
descripción en tercera persona de como mucho 900 caracteres. Para que la forma no se pierda, `make skills-check` gana
los gates de FR-002 a FR-011 en `internal/skills` (nombre, etiquetas XML, presupuestos del cuerpo y de la descripción,
`references/` plano y enlazado, índice), y la sincronía y las dependencias de las evals de H30 distinguen la referencia
a mano de la generada (FR-012, FR-014). Los controles de contenido de `make ci` leen la skill entera, `SKILL.md` y sus
referencias a mano, y un control nuevo ejecuta sin red las órdenes de ejemplo que leen del BOE (FR-016 a FR-019). La
tercera persona la juzga en el job de evals un juez con modelo, el de las evals de hoy con una clase nueva,
`habla_en_primera_o_segunda_persona`, la rúbrica, los casos y la medida de `evidencias/adr-0046/`, en el trabajo
`evals (<skill>)` de cada skill y solo cuando la propuesta cambia una descripción, con umbral 0 (FR-025 a FR-034). La
guía de autoría queda en `docs/SKILLS.md`. Las decisiones, con su alternativa descartada, están en research.md (D1 a
D21); las verificaciones hechas en la sesión, en V1 a V12.

## Technical Context

**Language/Version**: Go 1.27 (`go.mod`; `go version` en la sesión: go1.27.3, research V11). Markdown para las skills y
la guía; YAML de GitHub Actions para el job; bash 3.2 compatible en los guiones.

**Primary Dependencies**: las del módulo, sin ninguna nueva: `gopkg.in/yaml.v3` (frontmatter, `casos.yaml`),
`santhosh-tekuri/jsonschema` (el esquema de los votos, `CompilarEsquema`), `stretchr/testify` (tests); la biblioteca
estándar para runas (`unicode/utf8`), huellas (`crypto/sha256`) y rutas (`path`).

**Storage**: N/A. Se leen ficheros del repositorio (`skills/`, `data/`, `evidencias/adr-0046/`); el job escribe su
decisión e `informe.json` en `$RUNNER_TEMP`, como hoy.

**Testing**: `go test -race` en `make ci`, offline y sin modelo: mutantes sobre copias en `t.TempDir()`, la caché de
una sesión preparada con `PrepararSesion` en un temporal, votos simulados con un `Votante` de prueba y el `claude`
sustituto de `internal/evals/sustitutos_test.go`. Ningún fixture nuevo en `testdata/`.

**Target Platform**: macOS y Linux para `make ci`; `ubuntu-24.04` para el job de evals.

**Project Type**: CLI multicall (el binario no cambia), skills agénticas y el arnés de evals.

**Performance Goals**: `skills-check` sigue siendo de segundos; el control de las órdenes de ejemplo prepara una caché
de sesión por skill con evals (una preparación tardó menos de medio segundo en la sesión, research V3).

**Constraints**: sin red ni modelo en `make ci` y en el run (ADR 0032); el juez solo en el job y en la medida; las
reglas de las skills no cambian (FR-021); el `name`, `kitlegal-applets`, `kitlegal-referencias`, las referencias
generadas y sus generadores no cambian (FR-024); `scripts/workflow/`, `.specify/` e `internal/workflow/` fuera.

**Scale/Scope**: 3 skills, 11 referencias a mano previstas (7, 2 y 2) y 5 generadas, 45 evals en dos modos, 99 casos
del juez, como mucho 6 votos por descripción.

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones; la única desviación
justificada, «Aceptación e2e: no aplica», está en Complexity Tracking.*

### Principios

| Principio | Cómo lo cumple H31 |
|---|---|
| I · Fuentes públicas y frontera humana | Ninguna fuente nueva ni petición HTTP: los gates y controles leen ficheros; el control de las órdenes de ejemplo ejecuta `boe` con `--offline` sobre una caché local y un `httpx.Replay` vacío (`ComprobarSinRed`). El juez abre sesiones de Claude Code solo en el job, como el de H24. Ninguna skill gana una acción con identidad, y `jurisprudencia` conserva todas sus reglas sobre el CENDOJ (research, tabla, J04, J06–J08, J16, J48, J67) |
| II · Nada sin cita ni fuente | Ningún sobre cambia. Ninguna regla de cita ni de cobertura desaparece ni aparece (FR-021, tabla de research); las formas fijas que las evals comprueban sin modelo se quedan en el cuerpo (FR-018). Cada juicio reutilizado lleva su procedencia (FR-031) |
| III · Tests primero y offline | «Aceptación e2e: no aplica» (abajo; Complexity Tracking). Cada gate y control entra con su mutante en `make ci`, offline; las skills reescritas las miden sus 45 evals en el job (SC-010). Sin fixtures nuevos en `testdata/`: las copias y los casos se escriben en `t.TempDir()` |
| IV · Hexagonal y errores tipados | Ningún paquete del binario cambia de responsabilidad: `internal/skills` y el arnés de `internal/evals` validan y miden. Ningún código de salida de `kitlegal` cambia; un gate es un `DefectoDeSkill` o una `Deriva`, como hoy; lo ilegible es la regla genérica (defecto `inesperado`, código 1) |
| V · Simplicidad y dependencias | Ninguna dependencia nueva (Markdown con tres formas propias, research D14). Una función exportada nueva, `skills.ReferenciasAMano`, que usan tres consumidores (research D13); el juez de las descripciones es el `Juez` y la `votacion` de hoy con una clase (D6). Cada mecanismo, trazado a su FR en «Trazabilidad» |
| VI · Un binario, convenciones de agente | El binario, sus banderas, `--describe` y la región generada de la tabla de comandos no cambian (FR-024); `kitlegal skills install` instala las referencias a mano como cualquier fichero de `references/` (FR-015) |
| VII · Grafo y privacidad | Nada del grafo cambia; el juez ve solo la descripción de la skill, sin datos de ningún asunto |
| VIII · Skills primero | Es el hito del principio VIII 2.15.0: cuerpo corto, detalle en referencias a mano, lo de `data/` generado, presupuestos y forma en `make ci`, tercera persona con el juez de las evals; las tres skills reescritas sin perder ni ganar reglas, y las evals deciden (SC-010) |
| IX · Genericidad territorial | Ningún caso municipal: los gates y el juez no nombran municipio, comunidad ni boletín; el reparto de `legal-core` mueve su plantilla del territorio a una referencia sin tocar ninguna regla territorial (tabla de research, filas de los pasos 1 a 5 de `legal-core`) |
| X · Genericidad por materia | Los gates, la regla de `references/` y los controles son por skill, sin ninguna marca de materia; una materia nueva podrá entrar como referencia sin agrandar el cuerpo (ADR 0046). Ninguna regla de cobertura de `legal-core` cambia (tabla de research, L45–L52, paso 7) |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV, `internal/arch_test.go`)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | Sin cambios en `internal/core` |
| R2 · Solo `internal/httpx` importa `net/http` | Ningún import de `net/http`; el control de las órdenes de ejemplo usa `ComprobarSinRed`, que monta `httpx.Replay` como hoy |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | Ningún import nuevo de SQLite: la caché de la sesión la prepara y la lee `internal/cache` a través de `PrepararSesion` y del applet |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Ningún `os.Exit`: los puntos de entrada del job son tests que fallan con `t.Fatal` |
| R5 · Solo `render` escribe en stdout; logs a stderr | Ningún `fmt.Print*` ni `os.Stdout` nuevo: los defectos van en los errores de los tests; el job escribe en los ficheros de sus banderas y registra con `t.Log` |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios en `internal/graph` |
| R7 · Solo `internal/mcp` importa el SDK de MCP | Ningún import del SDK |
| «Sin red» de `arch_test.go` (instalación, disco, cita, ids, plazos, raíz y applets sin red) | Ningún paquete de esa lista cambia de importaciones; `ReferenciasAMano` lee un `fs.FS` |

## Project Structure

### Documentation (this feature)

```text
specs/027-h31-las-skills-siguen/
├── spec.md
├── plan.md                              # este fichero
├── research.md                          # D1–D21, V1–V12, S1–S2 y la tabla de reglas (FR-021)
├── data-model.md                        # cuerpo, descripción, referencias, juicio, juez, umbral, medida
├── quickstart.md                        # escenarios de validación, sin red ni modelo
├── contracts/
│   ├── gates-de-skills-check.md         # FR-002 a FR-015: defectos, enlaces, índice, clasificación
│   ├── controles-de-contenido.md        # FR-016 a FR-019
│   ├── juez-de-las-descripciones.md     # FR-025 a FR-034: decisión, votos, informe, umbral, medida
│   └── skills-reescritas.md             # FR-020 a FR-024: reparto de cada skill y su descripción
├── checklists/
└── tasks.md                             # lo escribe /speckit-tasks
```

### Source Code (repository root)

```text
internal/skills/
├── frontmatter.go         # FR-002, FR-003, FR-005, FR-011: anthropic/claude, etiqueta XML, 900; «SKILL.md: »
├── markdown.go            # nuevo · FR-007 a FR-009: bloques de código, enlaces, apartados ## e índice
├── referencias.go         # FR-012: ReferenciasAMano
├── sincronia.go           # FR-004, FR-006 a FR-012: cuerpo, clasificación, defectos de references/; sin maximoDeLineas
└── *_test.go              # TestValidarFrontmatter, TestEtiquetaXML, TestCuerpoDeLaSkill, TestEnlacesDeMarkdown,
                           # TestIndiceDeLaReferencia, TestReferenciasAMano, TestRegenerarYComparar
internal/app/skills_test.go  # TestSkillsDelRepositorio: frontmatter, presupuestos, referencias, deriva,
                             # sincronia-con-referencias-a-mano, normas-nombradas, sin-instrucciones-de-evals
internal/evals/
├── dependencias.go        # FR-014: (c) sin las referencias a mano; nombres también en ellas
├── conjunto_test.go       # FR-016 a FR-019: TestEvalsDelRepositorio, TestControlesSobreLasReferenciasAMano
├── juez.go                # FR-028, FR-029: juzgar un mensaje contra un texto
├── descripciones.go       # nuevo · FR-025 a FR-032: juez, mensaje, huella, juicio, casos de la medida
├── decision.go, origen.go # FR-030: bloque descripcion; la base también en la suite completa con propuesta
├── reutilizacion.go       # FR-030: el juicio en las medidas de main, en el mismo recorrido
├── ejecucion.go           # FR-026: comprobar la medida antes de abrir nada; votar la descripción
├── informe.go, umbrales.go # FR-031, FR-032: bloque, apartado, umbral, motivo
├── medida.go              # FR-034: medir la clase contra los 99 casos
├── definicion.go          # SC-015: peor caso con la descripción; entrada descripciones de medida
├── doc.go                 # el conjunto de dependencias con las referencias a mano
└── *_test.go              # TestConjuntoDeCadaEval, TestEleccionDeLosMutantes, TestJuezDeLasDescripciones,
                           # TestMedidaVersionada, TestEjecucionSinMedir, TestDecisionDelJob,
                           # TestBusquedaDeMedidasDeMain, TestInformeConLaDescripcion, TestDefinicionDelJob,
                           # TestGuionDeLaMedida; job_test.go (//go:build evals): TestDecisionDelTrabajo,
                           # TestMedidaDelJuezDeLasDescripciones
scripts/evals.sh           # las banderas del juez en todo trabajo
scripts/evals-medir-juez.sh  # descripciones
.github/workflows/evals.yml  # medida: entrada descripciones
skills/boe-legislacion/    # SKILL.md y references/{con_herramientas,buscar_la_norma,varios_bloques,
                           #   lecturas_que_fallan,como_se_cita,redaccion_en_una_fecha,redaccion_modificada}.md
skills/legal-core/         # SKILL.md y references/{municipios,calculo_de_plazos}.md
skills/jurisprudencia/     # SKILL.md y references/{consulta_preparada,documento_aportado}.md
docs/SKILLS.md             # nuevo · FR-001
CONTRIBUTING.md, CHANGELOG.md  # FR-023
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`. La forma de las skills es de `internal/skills`, que ya
valida el frontmatter y regenera `references/`; el juez es del arnés de evals, `internal/evals`, que ya tiene el de las
respuestas. Ningún paquete nuevo. `evidencias/adr-0046/` solo se lee (research D2).

## Aceptación e2e

**Aceptación e2e: no aplica.** El hito no cambia ningún comportamiento del binario: ningún applet, verbo, bandera,
código de salida ni salida; `kitlegal skills install` copia lo empotrado sin cambios de código, y `TestSkillsEmpotradas`
ya comprueba que todo fichero de `references/` va en el binario con sus bytes (research D20, D21). Lo que entrega son
gates y controles de `make ci`, el juez del job y el texto de las tres skills. La aceptación del spec es `make ci` en
verde con los gates sobre las tres skills (SC-001) y el job de evals sobre la cabeza de la propuesta, que mide las 45
evals de hoy en los dos modos y juzga las tres descripciones (SC-010); el spec deja fuera evals nuevas. Antes de
fusionar, una persona lanza la medida del juez (SC-011).

## Controles mecánicos que este hito añade o toca

Capa 1 de la constitución («Gates»), todos en `make ci` salvo lo que se dice del job:

- **Objetivos de make**: ninguno nuevo. `make skills-check` ejecuta los subtests nuevos de `TestSkillsDelRepositorio` y
  `TestEvalsDelRepositorio` con su expresión de hoy; `make skills-sync` da los defectos nuevos y no toca las referencias
  a mano; `make evals-medir-juez SKILL=descripciones` mide el juez nuevo (solo en el job).
- **`internal/skills`**: `TestValidarFrontmatter` (los defectos nuevos y el prefijo), `TestEtiquetaXML` (FR-003, con
  «< 5000», «a<b» y `<1>` sin defecto), `TestCuerpoDeLaSkill` (499/500 líneas, 11 000/11 001 runas con más bytes),
  `TestEnlacesDeMarkdown` (fuera y dentro de bloques, `#apartado`, resolución desde `references/`),
  `TestIndiceDeLaReferencia` (100/101 líneas, con y sin índice, sin `## `), `TestReferenciasAMano` (las cinco clases de
  contracts/gates-de-skills-check.md §3), `TestRegenerarYComparar` (clasificación y defectos en el árbol sintético).
- **`internal/app/skills_test.go`, `TestSkillsDelRepositorio`** (`skills-check`): sobre `copiaConLaSkill`, los mutantes
  de SC-002 —`claude-legal`, `x-anthropic`, etiqueta en `name` y en `description`, «< 5000», cuerpo de 500 y 499
  líneas, de 11 001 y 11 000 runas con caracteres de varios bytes, descripción de 901 y 900, subdirectorio y `.txt` en
  `references/`, referencia sin enlace, referencia enlazada desde otra, 101 líneas sin índice y 100, 101 sin `## ` y
  100—, cada uno con el defecto que nombra skill, fichero y regla; `deriva` con las clases de FR-012;
  `sincronia-con-referencias-a-mano` (SC-003); y los mutantes de `normas-nombradas` y `sin-instrucciones-de-evals` en
  una referencia a mano. Desaparece `trescientas-lineas`.
- **`internal/evals`, contenido**: `TestEvalsDelRepositorio` con los subtests de hoy leyendo la skill entera y los
  nuevos `lineas-de-aviso-de-las-referencias` y `ordenes-de-ejemplo`; `TestOrdenesDeEjemplo`, el extractor de las
  órdenes de ejemplo sobre Markdown escrito en el test (la de un bloque en sus formas de Bash y de PowerShell, la de
  código en línea, con `--fecha`; ninguna de la que lleva un marcador, de la mención del verbo sin argumentos ni de la
  prosa; contracts/controles-de-contenido.md §10); `TestControlesSobreLasReferenciasAMano`, un mutante por control
  (SC-005).
- **`internal/evals`, dependencias**: `TestConjuntoDeCadaEval` (la referencia a mano en (b), sin (c), sin error; la
  skill nombrada solo en una referencia a mano) y `TestEleccionDeLosMutantes` (mutante `referencia-a-mano-de-legal-core`
  y el de la descripción de hoy, que elige las 45).
- **`internal/evals`, juez**: `TestJuezDeLasDescripciones` (identidad con el manifiesto, mensaje frente a `mensaje_de`
  sobre un caso de `casos.yaml`, regla de los votos con votos simulados, el `claude` sustituto),
  `TestMedidaVersionada` (la medida de las descripciones y sus seis mutaciones), `TestEjecucionSinMedir` (ninguna
  sesión con la medida que no corresponde o no se cumple), `TestDecisionDelJob` (bloque o no según las rutas, en los
  dos modos y sin propuesta), `TestBusquedaDeMedidasDeMain` (reutilización por huella), `TestInformeConLaDescripcion`
  (marcada, sin marcar, sin juzgar, reutilizada; umbral y motivo), `TestDefinicionDelJob` (tope con la descripción;
  entrada `descripciones` de `medida`), `TestGuionDeLaMedida` (`descripciones`).
- **Lo empotrado**: `TestSkillsEmpotradas`, sin cambios (FR-015).
- **Fixtures**: ninguno en `testdata/` ni en `schemas/`; las copias, los casos y los votos se escriben en `t.TempDir()`
  o son constantes del test.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| SC-001 | `make ci` termina con 0 con los gates sobre las tres skills | Todos los de esta tabla y los de siempre | `ci:Makefile:ci` |
| FR-002, FR-003, FR-005, FR-011, SC-002, SC-009 | 0 defectos en las tres skills; `name` con `claude` o `anthropic` → rojo; etiqueta XML en `name` o `description` → rojo; «< 5000» → verde; 901 caracteres → rojo, 900 → verde | Mutantes sobre copias y las tres skills del repositorio | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-004, FR-010, SC-002, SC-009 | Cuerpo < 500 líneas y ≤ 11 000 runas en las tres; 500 → rojo, 499 → verde; 11 001 runas → rojo, 11 000 con más de 11 000 bytes → verde | Ídem | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-006 a FR-009, FR-011, SC-002 | 0 entradas no `.md`, 0 referencias sin enlace, 0 enlaces entre referencias, 0 referencias de más de 100 líneas sin índice o sin `## `; con 100, verde | Ídem | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-012, SC-003 | 0 bytes cambiados en la referencia a mano; 0 generadas sin rehacer (un byte cambiado y otra sin cabecera); 0 derivas sin dar | Sincronía sobre una copia y las clases de FR-012 | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-014, SC-004 | La referencia a mano en (b), 0 en (c), 0 errores; la skill nombrada solo en una referencia a mano, en el conjunto | Conjuntos del repositorio con el árbol superpuesto | `ci:internal/evals/dependencias_test.go:TestConjuntoDeCadaEval` |
| FR-014, SC-004 | 0 evals de diferencia con las de su skill y las que la nombran, en el mutante que solo cambia la referencia a mano | Un subtest por mutante | `ci:internal/evals/eleccion_test.go:TestEleccionDeLosMutantes` |
| FR-015 | 0 ficheros de `references/` fuera de lo empotrado o con otros bytes | Comparación con lo empotrado | `ci:skills_test.go:TestSkillsEmpotradas` |
| FR-016 a FR-018, SC-005 | Cada control de contenido sale en rojo con su regla en una referencia a mano: 0 mutantes sin detectar; 0 defectos en las tres skills | Copias de la skill con la regla movida | `ci:internal/evals/conjunto_test.go:TestControlesSobreLasReferenciasAMano` |
| FR-016 a FR-018 | 0 expresiones prohibidas o fechas con cifras, 0 órdenes sin su forma para PowerShell, 0 formas fijas fuera del cuerpo, 0 líneas `⚠` de una referencia que no estén en el cuerpo | Las tres skills del repositorio | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-016, SC-005 | 0 normas nombradas fuera de `data/normas.yaml` y 0 instrucciones de las evals, también en las referencias a mano; los dos mutantes, en rojo | Ídem y sus mutantes | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-019, SC-005 | 0 órdenes de ejemplo que no salgan con 0 con `--offline` sobre la caché de una sesión; ≥ 1 orden encontrada; la que no se sirve → rojo | Ejecución con `ComprobarSinRed` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-019 | Una orden por `kitlegal boe`, verbo y al menos un argumento en código; 0 órdenes de la mención del verbo sin argumentos, de la que lleva un marcador o de la prosa | Markdown escrito en el test | `ci:internal/evals/conjunto_test.go:TestOrdenesDeEjemplo` |
| FR-025, SC-006 | 0 bytes distintos en los cuatro ficheros respecto de `manifiesto.json`; un byte cambiado en cualquiera → rojo | Huellas de una copia | `ci:internal/evals/descripciones_test.go:TestJuezDeLasDescripciones` |
| FR-026, SC-008 | La medida corresponde y se cumple (0 de 63 sin marcar, 0 de 36 marcados) con lo que fija `evals.yml`; con cada mutación, 0 sesiones abiertas y el informe sin medir | Medida del repositorio y sus mutaciones; el job con un abridor de sesiones que falla si lo llaman | `ci:internal/evals/medida_test.go:TestMedidaVersionada`, `ci:internal/evals/ejecucion_test.go:TestEjecucionSinMedir` |
| FR-027 a FR-029, FR-032, FR-033, SC-007, SC-014 | Marca solo con tres «si» con frase; 0 votos pedidos tras un «no»; 1 repetición del nulo y del cortado; 0 del error; ≤ 6 votos; mensaje igual al de `mensaje_de`; 0 sesiones con modelo | Votante simulado y `claude` sustituto | `ci:internal/evals/descripciones_test.go:TestJuezDeLasDescripciones` |
| FR-030, SC-010 | Con una descripción cambiada, las 45 evals elegidas y 0 reutilizadas; bloque `descripcion` solo si alguna ruta cambia una descripción; 0 bloques sin propuesta, con la etiqueta sin cambios o con un cambio solo del juez | Elección y decisión de cada trabajo con la plataforma en memoria | `ci:internal/evals/eleccion_test.go:TestEleccionDeLosMutantes`, `ci:internal/evals/decision_test.go:TestDecisionDelJob` |
| FR-030 | 0 juicios reutilizados con otra huella, de una medida no válida o sin juzgar | Respuestas sintéticas de la plataforma | `ci:internal/evals/reutilizacion_test.go:TestBusquedaDeMedidasDeMain` |
| FR-031, FR-032 | 1 descripción marcada → umbral en rojo y veredicto `fallo`, con la frase y el motivo; sin juzgar → `fallo` con su motivo propio; reutilizada → 0 en el umbral | Informes compuestos con datos de prueba | `ci:internal/evals/informe_test.go:TestInformeConLaDescripcion` |
| FR-031, SC-010 | 0 descripciones marcadas de las juzgadas en la ejecución | El umbral `habla_en_primera_o_segunda_persona`, `decide: true` | `evals:boe-legislacion:habla_en_primera_o_segunda_persona`, `evals:legal-core:habla_en_primera_o_segunda_persona`, `evals:jurisprudencia:habla_en_primera_o_segunda_persona` |
| SC-010 | Los umbrales vigentes del juez de las respuestas, en verde con las skills reescritas | Los de hoy, sin cambios | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5`, `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5` |
| SC-015, FR-030 | Tope del trabajo ≥ peor caso de sus evals + 1 descripción × 6 votos × (35 + 5) s (+ 60 s de la instalación sin juez de las respuestas); tope de `medida` ≥ su peor caso con `descripciones` | Lee la definición del job | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |

Sin fila, porque no son un umbral que mida `make ci` ni el job: SC-011 lo mide una persona con la etiqueta
`evals-medir-juez`; SC-012 y SC-013 los contrasta la revisión final (la tabla de reglas y la documentación). Los demás
umbrales vigentes del job (los sin modelo de ADR 0038 y H30) conservan su control sin cambios: el job los publica con
`decide: true` como hoy, y la reescritura no los toca.

## Uso, de fuera adentro (criterio de uso, ADR 0028)

Ninguna salida de este hito crece con los meses de uso de kitlegal: el texto de las skills y el juicio de las
descripciones no dependen de las normas ni los bloques consultados (spec, «Uso, de fuera adentro»).

- **Descripción**: la lee el agente en cada sesión, una vez por skill instalada; como mucho 900 caracteres (unos 400
  tokens a 2,26 caracteres por token, ADR 0046); las tres, como mucho 2700. Los borradores: 878, 888 y 842 (V2).
- **Cuerpo**: una carga por pregunta por cada skill que activa; con `legal-core` y `boe-legislacion` juntas, como mucho
  22 000 caracteres (unos 9700 tokens) frente a 43 664 de hoy (24 439 + 19 225, V1). Los borradores: 10 881, 10 490 y
  10 605 (V2); al cerrar, 10 976, 10 860 y 10 859. Lo escribe una persona y lo acota el gate.
- **Referencias a mano**: de 0 a pocas lecturas por pregunta, cuando el cuerpo lo dice; las previstas tenían de 12 a
  64 líneas (contracts/skills-reescritas.md), y las escritas tienen de 24 a 68; sin índice porque ninguna pasa de 100.
- **Referencias generadas**: crecen con `data/`; `normas.md` tiene 24 líneas (V1). Con cientos de normas (unos 150
  bytes por fila: unos 45 KB con 300) pasa de 100 líneas, y FR-009 exige su índice; el cambio de `data/` que la lleve
  ahí adapta su generador (Clarificación Q4).
- **Defectos de los gates**: una línea de unos 100 bytes por regla y skill (contracts/gates-de-skills-check.md §2); se
  apagan al corregir la skill.
- **Órdenes de ejemplo**: una ejecución con `--offline` por aparición de una orden concreta —`kitlegal boe`, verbo y
  al menos un argumento; la mención del verbo solo no lo es—, cuatro con el reparto de contracts/skills-reescritas.md,
  todas de `boe-legislacion` (contracts/controles-de-contenido.md §10), y una caché de sesión por skill con evals, en
  cada `make ci`; no crece con el uso.
- **Juicio de la descripción**: una entrada por trabajo, unos 300 bytes sin votos y 1,5 KB con seis; como mucho seis
  votos por trabajo, 18 en la ejecución con las tres cambiadas; se apaga cuando la propuesta no cambia ninguna
  descripción, y la marca, cuando el juicio nuevo no la marca (contracts/juez-de-las-descripciones.md §8).
- **Comprobación de la medida**: una línea por ejecución que juzga descripciones si no corresponde; se apaga cuando
  vuelve a corresponder, que cambia una persona fuera de un run.

## Decisiones

Las del diseño están en research.md: dónde corre el juez (D1), de dónde lee (D2), cuándo juzga (D3), cómo reutiliza
(D4), el mensaje y la frase (D5), la regla de los votos (D6), el umbral y el informe (D7), la medida (D8, D9), el tope
(D10), los gates (D11, D12), la regla de `references/` (D13), Markdown (D14), las dependencias (D15), los controles
(D16, D17), el reparto de cada skill (D18), las versiones (D19), lo empotrado (D20) y la aceptación (D21).

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| `docs/SKILLS.md` | FR-001, SC-013 |
| Defectos de `name` (anthropic/claude, XML) y de `description` (XML, 900) en `ValidarFrontmatter`; prefijo `SKILL.md: ` | FR-002, FR-003, FR-005, FR-011 |
| Cuerpo y sus dos presupuestos en `regenerarSkill`; retirada de `maximoDeLineas` | FR-004, FR-010 |
| Defecto de entrada no `.md` en `references/` | FR-006 |
| `markdown.go` (bloques, enlaces, apartados, índice) y sus tres defectos | FR-007, FR-008, FR-009 |
| `ReferenciasAMano` y la clasificación en `sobrantes` | FR-012, FR-013 |
| `datosDeLasReferencias` sin las referencias a mano; nombres en ellas en `skillsDeLaEval` | FR-014 |
| Lectura de la skill entera en los controles 1, 3, 8 y 9; subtest `lineas-de-aviso-de-las-referencias` | FR-016, FR-017, FR-018 |
| Subtest `ordenes-de-ejemplo` y su extractor `ordenesDeEjemplo` | FR-019 |
| `TestControlesSobreLasReferenciasAMano` | SC-005 |
| Reescritura de las tres skills y sus referencias a mano; los dos ejemplos de `boe-legislacion` que la caché no sirve, uno con marcador y otro retirado | FR-020, FR-021, FR-022, FR-024, FR-019 |
| Entradas de `CHANGELOG.md` y `CONTRIBUTING.md` | FR-023 |
| `juezDeLasDescripciones` (lectura de `evidencias/adr-0046/`) y el control de identidad | FR-025, SC-006 |
| `comprobarLaMedida` antes de abrir nada con el bloque; la medida en `TestMedidaVersionada` | FR-026, SC-008 |
| `mensajeDeLaDescripcion` | FR-027 |
| `juzgar` partido en mensaje y texto (`juez.go`) | FR-028, FR-029, FR-032 |
| Bloque `descripcion` en la decisión (y la base en la suite completa con propuesta); búsqueda en `main` | FR-030 |
| Bloque `descripcion` en el informe, apartado de `informe.md`, umbral y motivo propio | FR-031, FR-032 |
| Banderas del juez en todo trabajo (`scripts/evals.sh`) | FR-030 (la descripción de cualquier skill) |
| `TestMedidaDelJuezDeLasDescripciones`, entrada `descripciones` de `medida`, `evals-medir-juez.sh descripciones` | FR-034, SC-011 |
| Peor caso con la descripción en `definicion.go` | SC-015 |

Lo que no se construye, por no pasar el umbral de materialidad: casos propios para un `references/` enlazado, truncado
o con permisos cambiados (regla genérica, defecto `inesperado`; spec, «Edge Cases»), un analizador de Markdown general
(D14) y una copia del juez (D2).

## Datos externos

Ninguno. El hito no graba de ninguna fuente ni deriva ningún fichero de `data/`: no hay manifiesto
`grabaciones.json`, test `TestGrabar*` ni paso `grabar_datos` que lo afecten, y no hace falta ninguna fila de
`docs/SOURCES.md`. La rúbrica, el esquema, los casos y la medida del juez son de `evidencias/adr-0046/`, en `main`
desde la entrada, y solo se leen.

## Orden de implementación (de dentro afuera)

Cada paso deja `make ci` en verde.

1. `ReferenciasAMano`, la clasificación de FR-012 en la sincronía y las dependencias de FR-014, con sus tests (sin
   referencias a mano en el árbol, nada cambia aún).
2. Los controles de contenido sobre la skill entera (FR-016 a FR-018) y `TestControlesSobreLasReferenciasAMano`.
3. La reescritura de cada skill —`jurisprudencia`, `legal-core`, `boe-legislacion`—, con sus referencias a mano, su
   descripción en tercera persona, los enlaces a todas sus referencias y su entrada de `CHANGELOG.md`; la de
   `boe-legislacion` cambia los dos ejemplos que la caché no sirve (D17).
4. El control de las órdenes de ejemplo (FR-019), en verde con las skills ya reescritas.
5. Los gates de FR-002 a FR-011 (`markdown.go`, frontmatter, cuerpo, `references/`), en verde con las skills ya
   reescritas, y la retirada del límite de 300 líneas.
6. El juez de las descripciones: `juzgar` partido, `descripciones.go`, la decisión, la búsqueda en `main`, la ejecución,
   el informe y el umbral, el tope, `scripts/evals.sh` y `evals.yml`.
7. La medida del juez (FR-034): el punto de entrada, la entrada `descripciones` de `medida` y el guion.
8. `docs/SKILLS.md` y `CONTRIBUTING.md` (FR-001, FR-023).

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|---|---|---|
| Principio III, «cada hito empieza por el test e2e (`testscript`)»: Aceptación e2e no aplica | El hito no cambia ningún comportamiento del binario: entrega gates y controles de `make ci`, el juez del job y el texto de tres skills, que miden sus evals (research D21) | Un guion `testscript` que instale las skills y mire una referencia a mano fijaría en la suite congelada nombres que diseña el hito y repetiría `TestSkillsEmpotradas`; el spec deja fuera evals nuevas, y las 45 de hoy son la aceptación en el job (SC-010), como en H24, H26 y H30 |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- **precheck**: `plan.md` y `research.md` existen; ninguna marca de aclaración pendiente; «## Constitution Check» con una fila por
  cada principio de I a X con su numeral en la primera celda; la línea «Aceptación e2e: no aplica» con su motivo;
  «## Controles de umbral» con filas cuyos requisitos están definidos en el spec y cuyo «Dónde» es `ci:` o `evals:`.
- **a · constitution_check**: cada fila dice lo que hace este plan; las siete reglas y la de «sin red».
- **b · dependencias**: ninguna nueva.
- **c · reglas_dependencia**: `internal/skills` e `internal/evals` no ganan imports de red, SQLite, MCP ni `os.Exit`.
- **d · errores_exit_codes**: ningún código cambia; defectos y derivas como hoy; lo ilegible, la regla genérica.
- **e · tests_primero**: «no aplica» con motivo; unitarios, mutantes y controles enumerados arriba; sin fixtures.
- **f · alcance**: nada del juez en `make ci` ni en el run; ningún generador cambia; nada de `scripts/workflow/`.
- **g · sin_atajos**: ningún `//nolint`, `t.Skip`, TODO ni error silenciado previsto.
- **h · mejor_alternativa**: cada decisión de research con su descartada.
- **i · afirmaciones_verificadas**: V1 a V12 ejecutadas en la sesión; S1 y S2, supuestos.
- **j · quickstart_ejecutable**: órdenes y tests nombrados aquí; solo el escenario 3 escribe, y no escribe en un árbol
  sincronizado.
- **k · datos_externos**: ninguno.
- **l · autonomía**: nada a mitad del run; la medida del juez y la release, de una persona fuera del run.
- **m · uso**: «Uso, de fuera adentro» y contracts/juez-de-las-descripciones.md §8.
- **n · proporcionalidad**: «Trazabilidad».
- **o · controles_de_umbral**: la tabla de arriba, una fila por cada requisito con umbral.
