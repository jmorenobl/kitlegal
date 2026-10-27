# Implementation Plan: H31.1 · `boe-legislacion` habla de la norma, no de la consulta, y lee lo que nombra

**Branch**: `029-h31-1-boe-legislacion-habla` | **Date**: 2026-10-27 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/029-h31-1-boe-legislacion-habla/spec.md`

## Summary

`boe-legislacion` pasa a v0.3.1 sin cambiar el binario: es la skill medida, `evidencias/sondeos-20261026/brazo-W/skill/`,
con el cuerpo reescrito para caber. Los pasajes medidos —el paso 3, la primera viñeta del paso 4, el paso 5 y el paso 6—,
las formas fijas, la tabla de comandos, el frontmatter, los encabezados y las reglas quedan como en ella; lo demás se
dice más corto, quitando lo que el cuerpo ya dice en otro sitio (casi siempre en los pasos 5 y 6 medidos) y lo que el
agente no haría mal sin ello, y sale el título `#`. El cuerpo pasa de 12 081 caracteres a **10 995**, y ningún texto sale
a una referencia: las ocho son las de la skill medida (research D1 a D3; el `SKILL.md` entero, en
[contracts/skill-md-v0.3.1.md](contracts/skill-md-v0.3.1.md); la tabla de textos, en
[contracts/skill-boe-legislacion-v0-3-1.md](contracts/skill-boe-legislacion-v0-3-1.md)). Tres guiones de aceptación
comparan la skill con la medida (D4). Las evals 14, 15 y 17 ganan nueve `lecturas`; una tarea `[datos]` amplía dos
entradas del manifiesto y `grabar_datos` graba ocho bloques (D6, D7). El peor caso del trabajo de evals no cambia
(20 117 s bajo 21 120 s; D8). `docs/SKILLS.md` §1 gana un párrafo con lo medido sobre las referencias que se leen
(D9), `CONTRIBUTING.md` deja de remitir a una regla 7 que ya es otra (D10) y `CHANGELOG.md` registra v0.3.1 (D11). Todo
se validó en una copia del repositorio en `$TMPDIR`: gates, controles, guiones y lecturas (research V3 a V6, V11).

## Technical Context

**Language/Version**: Markdown (la skill, la guía, el CHANGELOG), YAML (las evals), JSON (el manifiesto), testscript
(la suite) y Go 1.27 solo en dos tablas de tests (`go.mod`: `go 1.27.0`, `toolchain go1.27.3`; `go version`:
go1.27.3).

**Primary Dependencies**: ninguna nueva. Los guiones usan `rogpeppe/go-internal` (testscript) y `awk`, que ya usan los
guiones de H23; los tests, `stretchr/testify`.

**Storage**: N/A. Ficheros del repositorio y ocho grabaciones nuevas en `internal/evals/testdata/`, que escribe
`grabar_datos`.

**Testing**: `make ci`, sin red ni modelo: la suite de aceptación en `TestEntregaDelHito`; `TestSkillsDelRepositorio`,
`TestCuerpoDeLaSkill`, `TestSkillsEmpotradas`, `TestEvalsDelRepositorio`, `TestCacheDeLasSesiones`,
`TestLecturaDeLaEvalSinGrabar` y `TestDefinicionDelJob`. El significado de las respuestas lo mide el job de evals del
cierre y, con potencia, el sondeo de una persona (SC-012).

**Target Platform**: macOS y Linux para `make ci`; `ubuntu-24.04` para el job de evals.

**Project Type**: skill agéntica sobre el CLI multicall (el binario no cambia) y datos del arnés de evals.

**Performance Goals**: el cuerpo, como mucho 11 000 caracteres (10 995); el peor caso del trabajo de evals de
`boe-legislacion`, 20 117 s, dentro de los 21 120 s del tope.

**Constraints**: sin red ni modelo en el run (ADR 0032), salvo `grabar_datos`. Fuera: la descripción, los umbrales,
la rúbrica y los casos del juez, `topeDeUnaSesion`, `jurisprudencia`, `legal-core`, el binario, `scripts/workflow/`,
`.specify/`, `internal/workflow/`, `web/`, `docs/SKILLS.md` fuera de §1.

**Scale/Scope**: una skill (el `SKILL.md`; las referencias, las de la skill medida), tres guiones, tres evals con nueve
lecturas, dos entradas de manifiesto, ocho grabaciones, dos tablas de tests, un párrafo de la guía, una frase de
`CONTRIBUTING.md` y una entrada del CHANGELOG.

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones.*

### Principios

| Principio | Cómo lo cumple H31.1 |
|---|---|
| I · Fuentes públicas y frontera humana | La única fuente que se graba, `boe.legislacion-consolidada`, es pública y tiene fila en `docs/SOURCES.md` de `main` con «Revisado» 2026-09-13 (research V9). La graba `grabar_datos` con `TestGrabarEvals`, que pide por `internal/httpx` (GET, `robots.txt`, User-Agent, ritmo `1s`). Ningún código HTTP cambia. La regla 4 de la skill («Nunca actuar en nombre de nadie») conserva su encabezado y su frase |
| II · Nada sin cita ni fuente | Ningún sobre cambia. Los pasos 3, 5 y 6 medidos endurecen la cita: se lee lo que se nombra, de lo no leído no van número y materia juntos, y se repasa cada cosa dicha contra la orden que devolvió su texto. «Sin `fuente`, `url`, `fecha_consulta` y `hash`, no hay cita» sigue, y cada regla invariante de `CLAUDE.md` tiene su frase en el cuerpo ([contrato](contracts/skill-boe-legislacion-v0-3-1.md) §4). Nada de la cobertura de la skill cambia |
| III · Tests primero y offline | La primera tarea escribe los tres guiones de aceptación, que fallan hoy por una aserción y quedan congelados ([contracts/aceptacion.md](contracts/aceptacion.md)). En la tarea de las lecturas cambian primero las dos tablas de tests, en rojo hasta que las evals las llevan; las grabaciones las hace `grabar_datos` en la tarea `[datos]` anterior y los tests las leen sin red. Ningún `schemas/` cambia; no hay código Go de producción que cubrir |
| IV · Arquitectura hexagonal con reglas ejecutables | Ningún paquete, importación ni código de salida cambia. Una lectura sin grabar sigue siendo una falta de la preparación (H26.1), y una lectura de la skill que falla, el código de hoy; lo manipulado por fuera, la regla genérica (`inesperado`, 1) |
| V · Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia, paquete, tipo, función ni fichero de código nuevo, ninguna referencia nueva, ningún control nuevo fuera de la suite que pide SC-003. Cada cambio se traza a su requisito («Trazabilidad») |
| VI · Un binario, convenciones de agente | El binario, sus banderas, `--describe` y la tabla de comandos generada no cambian; `kitlegal skills install` instala la skill nueva como todo lo empotrado, que `TestSkillsEmpotradas` compara byte a byte |
| VII · Grafo y privacidad | Nada del grafo cambia; cada lectura nueva lleva detrás la comprobación de hoy, `kitlegal graph check <norma> <bloque>` |
| VIII · Skills primero; el binario es la herramienta | Protege `boe-legislacion` de dos defectos medidos (spec, «Contexto») con el texto que los bajó, en el cuerpo y no en una referencia, dentro del presupuesto del ADR 0046; las evals y su juez deciden en el cierre (SC-007 a SC-011) y el sondeo de una persona, antes de fusionar (SC-012). `docs/SKILLS.md` §1 registra lo medido sobre qué referencias se leen |
| IX · Genericidad territorial, validación local | Ningún municipio ni comunidad; la regla 5 conserva «Ningún caso especial para un territorio», lo de las normas locales y la variación autonómica |
| X · Genericidad por materia, validación administrativa | Lo que la rama añade al cuerpo no nombra ninguna norma de las evals por materia (FR-009, `h31-1-sitio-en-el-cuerpo`); la regla de leer lo que se nombra es por capacidad. Las lecturas son datos de evals, y la guía, de autoría |

### Reglas de dependencia (constitución IV; `docs/ROADMAP.md` §2; `internal/arch_test.go`)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa el kernel, los adaptadores ni entrada y salida | Sin cambios en `internal/core` |
| R2 · Solo `internal/httpx` importa `net/http` | Ninguna importación nueva; la grabación usa `TestGrabarEvals`, que pide por `internal/httpx` |
| R3 · Solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | Ninguna: `TestCacheDeLasSesiones` prepara y lee la caché con `PrepararSesion` y `app.Main`, como hoy |
| R4 · Solo `internal/cli` y `cmd/` llaman a `os.Exit` | Ningún `os.Exit` |
| R5 · Solo `internal/render` escribe en stdout; logs con `log/slog` a stderr | Ningún `fmt.Print*` ni `os.Stdout` en los tests que cambian: fallan con `assert` |
| R6 · `internal/graph` no importa `internal/source` ni `internal/render` | Sin cambios en `internal/graph` |
| R7 · Solo el paquete del servidor importa el SDK de MCP | Ninguna importación del SDK |
| Sin red (`arch_test.go`, «sin red · …») | Ningún paquete cambia de importaciones |

## Project Structure

### Documentation (this feature)

```text
specs/029-h31-1-boe-legislacion-habla/
├── spec.md
├── plan.md                                  # este fichero
├── research.md                              # D1–D12, V1–V20, S1–S5
├── data-model.md                            # SKILL.md y sus partes, referencia, lectura, manifiesto, grabación, guion
├── quickstart.md                            # doce escenarios, sin red ni modelo
├── contracts/
│   ├── skill-md-v0.3.1.md                   # el SKILL.md de la entrega, byte a byte
│   ├── skill-boe-legislacion-v0-3-1.md      # FR-001 a FR-010: lo que no cambia, textos reescritos, textos de una clase
│   ├── aceptacion.md                        # los tres guiones, validados
│   ├── lecturas-y-grabaciones.md            # FR-011 a FR-013
│   └── guia-y-changelog.md                  # FR-014 y FR-016
├── aceptacion/                              # lo escribe T001 y queda congelado
├── checklists/
└── tasks.md                                 # lo escribe /speckit-tasks
```

### Source Code (repository root)

```text
skills/boe-legislacion/
├── SKILL.md                                 # = contracts/skill-md-v0.3.1.md (FR-001 a FR-010)
└── references/
    ├── buscar_la_norma.md                   # = la de la skill medida (FR-001)
    └── redaccion_modificada.md              # = la de la skill medida (FR-001)
evals/boe-legislacion/
├── 14-trlrhl-impuestos-por-materia.yaml     # FR-011: a60, a92, a100, a104
├── 15-irpf-rendimientos-por-materia.yaml    # FR-011: a7, a19, a20, dadecimoctava
└── 17-ltaibg-plazo-por-materia.yaml         # FR-011: a24
internal/evals/
├── testdata/grabaciones.json                # FR-012 [datos]: bloques y para de dos entradas
├── testdata/boe.legislacion-consolidada/    # FR-012: ocho grabaciones (grabar_datos)
├── conjunto_test.go                         # FR-011: el mapa de TestEvalsDelRepositorio/lecturas
└── preparar_test.go                         # FR-013: la tabla de TestCacheDeLasSesiones/lecturas
specs/029-h31-1-boe-legislacion-habla/aceptacion/
└── h31-1-{responder,lee-lo-que-nombra,sitio-en-el-cuerpo}.txtar   # SC-003; activados en internal/app/testdata/script/
docs/SKILLS.md                               # FR-014: un párrafo de §1
CONTRIBUTING.md                              # FR-001 (consecuencia): la frase de la línea 743
CHANGELOG.md                                 # FR-016
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`. La skill vive en `skills/`, sus evals en `evals/`,
el manifiesto y las grabaciones de las evals en `internal/evals/testdata/` (H9, H26.1), los tests de las lecturas en
`internal/evals` y la suite en el directorio del feature hasta que el workflow la activa. Ningún paquete ni fichero de
código nuevo; `.github/workflows/evals.yml` e `internal/evals/definicion.go` no cambian (research D8).

## Aceptación e2e

**Aceptación e2e:** tres guiones testscript, que T001 escribe en `specs/029-h31-1-boe-legislacion-habla/aceptacion/` y
quedan congelados ([contracts/aceptacion.md](contracts/aceptacion.md), con su texto validado):

- `h31-1-responder.txtar` (US1) — la primera viñeta del paso 4 y el paso 6, idénticos a los de la skill medida:
  FR-002, SC-003.
- `h31-1-lee-lo-que-nombra.txtar` (US2) — el paso 3 y el paso 5, idénticos; la regla de leer lo que se nombra, en el
  cuerpo: FR-002, SC-003.
- `h31-1-sitio-en-el-cuerpo.txtar` (US3) — frontmatter, formas fijas, tabla, encabezados, número y encabezado de cada
  regla, las ocho referencias iguales y ninguna más, y el cuerpo sin las normas de las evals por materia: FR-001,
  FR-003, FR-004, FR-006, FR-007, FR-009, FR-010, SC-003.

Hoy fallan los tres por una aserción de `stdout` (research V5). US4 la prueban los tests de las lecturas de H26.1
(research D5); US5, el cierre; US6, la revisión final; US7, el sondeo de una persona.

## Controles mecánicos que este hito añade o toca

Capa 1 de la constitución («Gates»), en `make ci` salvo lo que se dice del workflow:

- **Objetivos de make**: ninguno nuevo ni cambiado; `make skills-sync`, `make skills-check` y `make ci` ejecutan lo de
  abajo.
- **Suite de aceptación, nueva**: los tres guiones de arriba; el workflow comprueba su rojo primero
  (`aceptacion.sh rojo-primero`), los congela y, al terminar el bucle, los activa en `internal/app/testdata/script/`,
  donde los ejecuta `TestEntregaDelHito`.
- **Gates de H31, sin cambios, sobre la skill nueva** (`internal/app/skills_test.go`, `TestSkillsDelRepositorio`;
  `internal/skills/sincronia_test.go`, `TestCuerpoDeLaSkill`): cuerpo de 11 000 caracteres como mucho, cada
  referencia enlazada, tabla y referencias generadas sin drift, `normas-nombradas` y `sin-instrucciones-de-evals`.
- **Controles de contenido de `docs/SKILLS.md` §1, sin cambios** (`internal/evals/conjunto_test.go`,
  `TestEvalsDelRepositorio`): `prosa-de-la-skill`, `ordenes-para-powershell`, `linea-sin-consulta`,
  `formas-de-la-redaccion-fechada`, `avisos-de-la-skill`, `hallazgos-de-la-skill`,
  `lineas-de-aviso-de-las-referencias` y `ordenes-de-ejemplo`.
- **Tests que cambian** (FR-011, FR-013): el mapa escrito a mano de `TestEvalsDelRepositorio/lecturas` y la tabla de
  `TestCacheDeLasSesiones/lecturas` ([contrato de lecturas](contracts/lecturas-y-grabaciones.md) §4). Sin cambios,
  pero ahora sobre las lecturas nuevas: `TestEvalsDelRepositorio/grabado` y `TestLecturaDeLaEvalSinGrabar`.
- **Job**: `TestDefinicionDelJob` (`internal/evals/definicion_test.go`) sigue exigiendo que `timeout-minutes` cubra el
  peor caso de cada trabajo; los umbrales del informe, sin cambios.
- **Fixtures**: ocho grabaciones nuevas en `internal/evals/testdata/boe.legislacion-consolidada/`, de `grabar_datos`
  tras la tarea `[datos]`; puede regrabarse el `robots.txt` del BOE (research S4). Ninguna en `schemas/`.
- **Workflow**: guardián de diff (rutas declaradas, suite congelada, `testdata/` protegido), fila revisada de la fuente
  en `main` para `grabar_datos`.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| SC-001 | `make ci` termina con 0 | Todos los de esta tabla que viven en `make ci` y los de siempre | `ci:Makefile:ci` |
| FR-005, SC-002 | Cuerpo de `SKILL.md` ≤ 11 000 caracteres (10 995); con 11 001, rojo | Gate del cuerpo de H31 sobre las skills del repositorio y sobre copias de 11 000 y 11 001 | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, `ci:internal/skills/sincronia_test.go:TestCuerpoDeLaSkill` |
| FR-007 | 0 ficheros de `references/` sin enlace desde `SKILL.md` | Subtest `referencias` del gate de H31 | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-004 | 0 diferencias entre la tabla del `SKILL.md` y la que genera `--describe` | Regeneración del gate de H31 (`make skills-check`) | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-008, SC-001 | 0 defectos de los diez controles de contenido en el `SKILL.md` y en cada referencia a mano, también 0 órdenes de ejemplo que no salgan con 0 con `--offline` | Los ocho subtests de `TestEvalsDelRepositorio` y `normas-nombradas` y `sin-instrucciones-de-evals` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-001, FR-002, FR-003, FR-004, FR-006, FR-007, FR-009, FR-010, SC-003 | 0 diferencias con la skill medida en los pasajes medidos, las líneas de los bloques `text`, la tabla, el frontmatter, los encabezados, el número y el encabezado de las reglas y las ocho referencias; 0 líneas del cuerpo con una norma de FR-009 | Los tres guiones de la suite, activados | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| FR-011, SC-004 | Las lecturas de cada eval, exactamente las del contrato: 0 de más, 0 de menos, 0 motivos distintos | Subtest `lecturas`, con el mapa escrito a mano | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-012, FR-013, SC-004 | 0 consultas necesarias de las evals —las lecturas, sin fecha y con `9999-12-31`— sin su respuesta en la unión de grabaciones o que la caché preparada no sirva sin red | Subtest `grabado` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-013, SC-004 | Cada uno de los nueve bloques sale con 0 con `boe articulo … --offline` desde la caché de una sesión, sin fecha y con `--fecha 9999-12-31` | Subtest `lecturas` de la caché | `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones` |
| FR-012 | Con una lectura sin grabar, sus 2 faltas y la sesión sin preparar | Sin cambios (US4, escenario 3) | `ci:internal/evals/preparar_test.go:TestLecturaDeLaEvalSinGrabar` |
| SC-008 | 0 sesiones de `claude-sonnet-5-5` de las evals 14, 15 y 17 con una orden o llamada en `fuera_de_lo_grabado` que sale con 4 o 5 porque la caché no sirve un bloque de sus `lecturas` (no cuenta la `articulos` que sale con 5 por otro bloque sin grabar de la misma norma) | Las tres filas de arriba: salen en rojo si la caché no sirve un bloque de las `lecturas` sin red, sin fecha o con `9999-12-31`, o si falta su grabación; una lectura con otra fecha la sirve la caché preparada (research D18 de H20) | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones`, `ci:internal/evals/preparar_test.go:TestLecturaDeLaEvalSinGrabar` |
| FR-015, SC-005 | `timeout-minutes` ≥ peor caso de cada trabajo: 21 120 s frente a 20 117 s de `boe-legislacion` | `fallosDelTope` sobre la definición del repositorio | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |
| SC-007 | `afirma_lo_no_leido` de `claude-sonnet-5-5`, los modos juntos, ≤ 0.028; si pasa, veredicto `fallo` y job en rojo | Umbral del juez, `decide: true` | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5` |
| SC-009 | Cada umbral que decide de `boe-legislacion` se cumple; el que no, veredicto `fallo` y job en rojo | Los umbrales de hoy, `decide: true` | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`, `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`, `evals:boe-legislacion:redaccion_sin_fechar:claude-sonnet-5-5:orden`, `evals:boe-legislacion:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`, `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`, `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados`, `evals:boe-legislacion:duracion_de_las_sesiones:orden`, `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`, `evals:boe-legislacion:duracion_del_juez:orden`, `evals:boe-legislacion:duracion_del_juez:herramienta` |
| SC-011 | `jurisprudencia`: cada umbral que decide se cumple (veredicto `aprobado`); si no, job en rojo | Los umbrales de hoy, `decide: true` | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`, `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`, `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`, `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`, `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`, `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`, `evals:jurisprudencia:duracion_del_juez:orden`, `evals:jurisprudencia:duracion_del_juez:herramienta` |
| SC-011 | `legal-core`: cada umbral que decide se cumple (veredicto `aprobado`); si no, job en rojo | Los umbrales de hoy, `decide: true` | `evals:legal-core:sin_activar:claude-sonnet-5-5:orden`, `evals:legal-core:sin_activar:claude-sonnet-5-5:herramienta`, `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:orden`, `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:herramienta`, `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:orden`, `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`, `evals:legal-core:vigencia_sin_declarar:claude-sonnet-5-5:orden`, `evals:legal-core:vigencia_sin_declarar:claude-sonnet-5-5:herramienta`, `evals:legal-core:redaccion_sin_fechar:claude-sonnet-5-5:orden`, `evals:legal-core:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`, `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:orden`, `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:herramienta` |

Sin fila, porque no son un umbral que se mida sin persona en `make ci` o en el cierre: SC-010 (`cuenta_su_proceso` se
publica por modo con `decide: false`, como hoy; hacerlo decidir pide medir antes su juez, ADR 0037); SC-012 (el
sondeo, de una persona); SC-013, FR-014 y la tabla de textos de FR-006 (los coteja la revisión final con `LEEME.md` y
el contrato de la skill §2); SC-006 y FR-016 (la entrada de `CHANGELOG.md`, y que el job mida las tres skills, que
lanza el workflow). Los nombres de SC-009 y SC-011 son los que deciden en los informes del cierre de H31 sin
`habla_en_primera_o_segunda_persona`, que solo va cuando la propuesta cambia una descripción, y aquí no cambia
ninguna (FR-010; research V14, V15). `fuera_de_lo_grabado` no decide (H26.1) y hacerlo decidir queda fuera del hito:
SC-008 lo hacen cumplir los tres tests de su fila, que prueban que la caché de las sesiones sirve sin red cada bloque
de las `lecturas` y que una lectura sin grabar para la preparación (research V11).

## Uso, de fuera adentro

- **El cuerpo** (FR-001 a FR-010): lo carga el agente entero, una vez por pregunta que activa la skill: 10 995
  caracteres, unos 4 870 tokens a 2,26 caracteres por token (ADR 0046, línea 24), 3 más que hoy. No crece con el uso:
  lo escribe una persona y el gate lo acota. Ninguna referencia nueva: las que se leen siguen siendo las de hoy, cuando
  su paso las pide.
- **Las lecturas de la regla de leer lo que se nombra** (paso 3 y paso 6.1, medidos): las pide el agente una vez por
  precepto que la respuesta va a nombrar y no ha leído, si lo remite un bloque leído o su id solo salió en un índice;
  cada una, `kitlegal boe articulo <norma> <bloque> --json && kitlegal graph check <norma> <bloque> --json` (o
  `boe_articulo` y `graph_check`), o una `articulos` con su comprobación para varios de una norma. Lo acotan la
  respuesta y el salto: solo lo que la respuesta nombra, y nada de lo que remiten esas lecturas. **Forma**: el sobre de
  `articulo` de hoy, `{ok, fuente, url, fecha_consulta, hash, data}`, con `data` = `{norma, bloque, titulo, tipo,
  fecha_version, fecha_vigencia, norma_modificadora, texto, hash_texto, avisos, url, url_eli}`. **Ejemplo**, entero,
  893 bytes con su salto de línea final (research V12):

  ```json
  {"ok":true,"fuente":"boe.legislacion-consolidada","url":"https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2004-4214/texto/bloque/a60","fecha_consulta":"2026-10-27T17:50:26.784726+01:00","hash":"sha256:e808efb86e7c6155dcd160d1d1f96c750f2d70b4923ddd6b511175f2dbe8000c","data":{"norma":"BOE-A-2004-4214","bloque":"a60","titulo":"Artículo 60","tipo":"precepto","fecha_version":"20040309","fecha_vigencia":"20040310","norma_modificadora":"BOE-A-2004-4214","texto":"Artículo 60. Naturaleza.\nEl Impuesto sobre Bienes Inmuebles es un tributo directo de carácter real que grava el valor de los bienes inmuebles en los términos establecidos en esta ley.","hash_texto":"sha256:a139c4802dab2508a832b8493ce9bcd10ab7beddb26a2a78c35b090b110d8163","avisos":[],"url":"https://www.boe.es/buscar/act.php?id=BOE-A-2004-4214#a60","url_eli":"https://www.boe.es/eli/es/rdlg/2004/03/05/2"}}
  ```

  Su comprobación, `graph check` con norma y bloque, tras la primera lectura de un bloque, 324 bytes (research V12):

  ```json
  {"ok":true,"fuente":"kitlegal.graph","url":"kitlegal:applet/graph","fecha_consulta":"2026-10-27T17:58:38.153735+01:00","hash":"sha256:7eda38c7b39b245a0e1a017f53ab23024603694b177593a1e81fb325437e5eae","data":{"norma":"BOE-A-2015-10565","bloques":["a21"],"version-obsoleta":0,"fuente-caducada":0,"omitidos":0,"hallazgos":[]}}
  ```

  **Tamaño por pregunta**: el del texto del bloque; de 893 bytes (`a60`) a 22 303 (`a7`) en los nueve, y la
  comprobación, con solo los hallazgos de los bloques pedidos y nunca más de 50 (tabla de «Comandos»). En la eval 15,
  como mucho los cuatro nuevos además de `a17` y `a18`: 31 101 bytes más. **Con meses de uso** —cientos de normas y
  miles de bloques consultados, la mayoría hace más de una semana— no cambia: cada lectura pide un bloque por su id y
  devuelve ese bloque, y su comprobación lleva norma y bloque, así que lista como mucho las entradas de ese bloque, nunca
  lo acumulado (la comprobación sin argumentos sigue prohibida en el paso 2).
- **Sus señales**: los avisos de vigencia del bloque, que trae el BOE con cada lectura y van mientras el BOE los dé; y
  `⚠ REDACCIÓN MODIFICADA:`, solo si la redacción del bloque cambió desde una lectura anterior, que deja de darse al
  leerlo porque el grafo guarda la nueva («lee cada bloque una vez», paso 2). `fuente-caducada` y lo de un bloque sin
  `version-obsoleta` no van en la respuesta (`redaccion_modificada.md`). Ninguna señal nueva.
- **Las `lecturas` y sus grabaciones** (FR-011 a FR-013): las pide la preparación de la caché de cada sesión de
  `boe-legislacion`, una vez por sesión: nueve bloques más, cada uno sin fecha y con `9999-12-31`, sin red; la
  preparación con ellas tarda 0,54 s (V6). Ocho grabaciones en el repositorio, de 1 766 a 402 526 bytes (V10). No crecen
  con el uso de kitlegal, solo con lo que se escribe en `evals/`.
- **La duración del trabajo de evals** (FR-015): el peor caso, 20 117 s, no cambia (research D8). Las sesiones de las
  evals por materia leen más y repasan, dentro de su tope de 240 s; si una no cabe, el informe lo dice con su código
  124, y si un modo pasa de su objetivo, `duracion_de_las_sesiones` decide (SC-009).
- **`docs/SKILLS.md` §1** (FR-014): un párrafo de unos 1 100 caracteres que lee quien escribe o reparte una skill, una
  vez por skill. Deja de valer cuando otra evidencia mida otra cosa.
- **`CHANGELOG.md`** (FR-016): una entrada que lee quien actualiza kitlegal, una vez por versión.

## Decisiones

En research.md: la skill de la entrega y por qué ningún texto sale (D1), cómo se reescribe (D2), el título (D3), la
suite (D4), US4 sin guion (D5), el motivo de las lecturas (D6), el manifiesto (D7), el peor caso (D8), la guía (D9),
`CONTRIBUTING.md` (D10), el CHANGELOG (D11) y lo que no cambia (D12).

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| `SKILL.md` = `contracts/skill-md-v0.3.1.md`: pasajes medidos, formas, tabla, frontmatter, encabezados y reglas como en la skill medida | FR-001 a FR-004, FR-010 |
| Los textos reescritos T1 a T15 (contrato de la skill §2) | FR-005, FR-006 |
| Ningún texto a una referencia; las ocho referencias de la skill medida | FR-001, FR-007 |
| Los tres guiones de aceptación | SC-003 (FR-001 a FR-004, FR-006, FR-007, FR-009, FR-010) |
| La frase de `CONTRIBUTING.md:743` | FR-001 (consecuencia: la regla 7 de la skill medida es otra; research D10) |
| Las nueve `lecturas` con su `motivo` | FR-011 |
| Bloques y `para` de dos entradas del manifiesto; las ocho grabaciones | FR-012 |
| El mapa de `TestEvalsDelRepositorio/lecturas` y la tabla de `TestCacheDeLasSesiones/lecturas` | FR-011, FR-013, SC-004 |
| El párrafo de `docs/SKILLS.md` §1 | FR-014, SC-013 |
| El recálculo del peor caso, sin cambio de tope | FR-015, SC-005 |
| La entrada de `CHANGELOG.md` | FR-016, SC-006 |

Lo que no se construye: ningún control nuevo fuera de la suite (D4, D5), ninguna referencia nueva (D1), ningún
cambio del tope del trabajo ni del de una sesión (D8), y nada para ficheros del kit, de la caché o de las grabaciones
tocados a mano: regla genérica, defecto `inesperado`, código 1 (spec, «Edge Cases»).

## Datos externos

Una fuente, `boe.legislacion-consolidada`, con fila en `docs/SOURCES.md` de `main` («Revisado» 2026-09-13). Manifiesto:
`internal/evals/testdata/grabaciones.json` (`"fuente": "boe.legislacion-consolidada"`). Test de grabación:
`TestGrabarEvals` (`internal/evals/grabacion_test.go`, `//go:build grabacion`), que ya existe y pide con
`internal/httpx`. Respuestas: `internal/evals/testdata/boe.legislacion-consolidada/`. Material de origen en
`$KITLEGAL_EVIDENCIAS` (`evidencias/h31-1/`): ninguno, porque ningún fichero de `data/` sale de lo grabado. La tarea
`[datos]` escribe solo el manifiesto y graba `grabar_datos`, sin modelo
([contrato de lecturas](contracts/lecturas-y-grabaciones.md) §2 y §3). Ninguna fuente sin fila revisada entra.

## Orden de implementación (de dentro afuera)

Cada paso deja `make ci` en verde; la suite vive en el directorio del feature hasta su activación.

1. **`[aceptacion]`**: los tres guiones de `contracts/aceptacion.md` en `aceptacion/` (SC-003).
2. **`[datos]`**: los ocho bloques y su `para` en el manifiesto (FR-012). Después, `grabar_datos` graba.
3. **Las lecturas**: el mapa de `TestEvalsDelRepositorio/lecturas` y la tabla de `TestCacheDeLasSesiones/lecturas`,
   en rojo; después, las nueve entradas en las evals 14, 15 y 17 (FR-011, FR-013).
4. **La skill**: `skills/boe-legislacion/SKILL.md` desde el contrato, las dos referencias de la skill medida y la
   frase de `CONTRIBUTING.md`, con `make skills-sync` y `make skills-check` en verde (FR-001 a FR-010).
5. **La guía y el CHANGELOG**: el párrafo de `docs/SKILLS.md` §1 y la entrada de `CHANGELOG.md` (FR-014, FR-016).

FR-015 no tiene tarea: el recálculo es el de research D8 y `TestDefinicionDelJob` ya lo hace cumplir.

## Complexity Tracking

Sin violaciones que justificar: ninguna dependencia nueva, ninguna desviación de la constitución.

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- **precheck**: `plan.md` y `research.md` existen; ninguna marca de aclaración pendiente; «## Constitution Check» con
  una fila por cada principio de I a X con su numeral en la primera celda; la línea «Aceptación e2e:» con sus guiones;
  «## Controles de umbral» con filas cuyos requisitos define el spec y cuyo «Dónde» es `ci:` o `evals:`.
- **a · constitution_check**: cada fila dice lo que hace este plan; las siete reglas, las dos de símbolo (R4, R5)
  incluidas, y la de «sin red».
- **b · dependencias**: ninguna nueva.
- **c · reglas_dependencia**: ninguna importación nueva en ningún paquete.
- **d · errores_exit_codes**: ningún código cambia; la falta de una lectura sin grabar, la de H26.1.
- **e · tests_primero**: los tres guiones, los dos tests que cambian, los que siguen igual y los fixtures, arriba.
- **f · alcance**: solo FR-001 a FR-016; `CONTRIBUTING.md`, por la consecuencia de FR-001 (D10).
- **g · sin_atajos**: ningún `//nolint`, `t.Skip`, TODO ni error silenciado.
- **h · mejor_alternativa**: cada decisión de research con su descartada.
- **i · afirmaciones_verificadas**: V1 a V20, ejecutadas en esta sesión o leídas en el repositorio con su ruta; S1 a
  S5, supuestos.
- **j · quickstart_ejecutable**: órdenes de lectura, probadas hoy las que no dependen de la entrega; solo el escenario
  12 escribe, ficheros ignorados.
- **k · datos_externos**: «Datos externos».
- **l · autonomía**: nada a mitad del run; el sondeo, la fusión y la release, de una persona fuera del run.
- **m · uso**: «Uso, de fuera adentro», con el ejemplo y los bytes medidos de V12.
- **n · proporcionalidad**: «Trazabilidad».
- **o · controles_de_umbral**: la tabla de arriba, una fila por cada requisito con umbral.
