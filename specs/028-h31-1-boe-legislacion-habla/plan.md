# Implementation Plan: H31.1 · `boe-legislacion` habla de la norma, no de la consulta, y lee lo que nombra

**Branch**: `028-h31-1-boe-legislacion-habla` | **Date**: 2026-10-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/028-h31-1-boe-legislacion-habla/spec.md`

## Summary

`boe-legislacion` pasa a v0.3.1 sin cambiar el binario. El paso «Responder» del cuerpo vuelve a decir, con las palabras
de antes de H31 y sus dos ejemplos, que la respuesta es todo lo escrito tras la última orden o llamada, desde su primera
palabra, que quien pregunta la lee entera y no ve las órdenes, las llamadas ni lo que devuelven, que lo decidido al
completar la lectura no va en ella y que lo hecho, lo que se va a hacer y lo que devolvió una orden no le sirven; el paso
5 recupera sus cuatro ejemplos (FR-001). El paso 3 gana la regla de leer lo que se nombra —un precepto solo se nombra si
`kitlegal boe articulo` o `articulos` ha devuelto su texto, y lo remitido por un bloque del paso 2 o salido de un índice
se lee con los pasos 1 y 2 antes de nombrarlo, o no se nombra— y el paso 5, el salto: lo que remite esa lectura, como lo
que no se pudo leer, se nombra solo con las palabras del texto leído (FR-002, FR-003).
El cuerpo queda en 10 993 caracteres porque siete textos que solo pide una clase de pregunta pasan a referencias a mano
que ya existen y que el cuerpo enlaza en el paso que las pide (FR-005, FR-006). Las evals 14, 15 y 17 llevan en
`lecturas` los nueve preceptos que sus respuestas nombraban sin leer; una tarea `[datos]` los añade al manifiesto y
`grabar_datos` graba los ocho que faltan (FR-009 a FR-011). `docs/SKILLS.md` §1 dice qué referencias leen las sesiones,
con lo medido (FR-012), y `CHANGELOG.md` registra v0.3.1 (FR-014). El tope del trabajo de evals no cambia: las lecturas
no tocan su peor caso (FR-013). El texto exacto se validó en una copia del repositorio con los gates y controles de
`make skills-check` en verde (research V6). Decisiones, en research.md (D1 a D15); verificaciones, V1 a V13.

## Technical Context

**Language/Version**: Markdown para la skill, la guía y el CHANGELOG; YAML para las evals; JSON para el manifiesto; Go
1.27 solo en dos tests que enumeran las lecturas (`go.mod`: `go 1.27.0`, `toolchain go1.27.3`; `go version` en la
sesión: go1.27.3).

**Primary Dependencies**: ninguna nueva. Los tests usan las de hoy (`stretchr/testify`, `gopkg.in/yaml.v3`).

**Storage**: N/A. Ficheros del repositorio; ocho grabaciones nuevas en `internal/evals/testdata/` que escribe
`grabar_datos`.

**Testing**: `go test -race` en `make ci`, sin red ni modelo: `TestSkillsDelRepositorio`, `TestEvalsDelRepositorio`,
`TestCacheDeLasSesiones`, `TestLecturaDeLaEvalSinGrabar`, `TestDefinicionDelJob`, `TestSkillsEmpotradas`. El
significado de las respuestas lo mide el job de evals del cierre y, con potencia, el sondeo de una persona (SC-012).

**Target Platform**: macOS y Linux para `make ci`; `ubuntu-24.04` para el job de evals.

**Project Type**: skill agéntica sobre el CLI multicall (el binario no cambia) y datos del arnés de evals.

**Performance Goals**: el cuerpo, como mucho 11 000 caracteres (10 993); el peor caso del trabajo de evals de
`boe-legislacion`, 20 117 s, dentro de los 352 minutos (21 120 s) de hoy (research V3).

**Constraints**: sin red ni modelo en el run (ADR 0032), salvo `grabar_datos`; la descripción, la tabla de comandos, los
umbrales, la rúbrica y los casos del juez, el tope de una sesión (240 s), `jurisprudencia`, `legal-core`,
`scripts/workflow/`, `.specify/`, `internal/workflow/` y `web/`, fuera.

**Scale/Scope**: 1 skill (cuerpo y 2 referencias a mano), 3 evals, 9 lecturas, 2 entradas de manifiesto, 8
grabaciones, 2 tests, 1 apartado de la guía, 1 entrada del CHANGELOG y 1 frase de `CONTRIBUTING.md`.

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones; la única desviación
justificada, «Aceptación e2e: no aplica», está en Complexity Tracking.*

### Principios

| Principio | Cómo lo cumple H31.1 |
|---|---|
| I · Fuentes públicas y frontera humana | La única fuente que se graba, `boe.legislacion-consolidada`, es pública y tiene fila revisada en `docs/SOURCES.md` de `main` (2026-09-13); la graba `grabar_datos` con `TestGrabarEvals`, que pide por `internal/httpx` (robots.txt, User-Agent, `1s`), solo GET. La regla 4 de la skill («Nunca actuar en nombre de nadie») no cambia |
| II · Nada sin cita ni fuente | Ningún sobre cambia. La regla nueva endurece la cita: un precepto solo se nombra si `kitlegal boe articulo` o `articulos` ha devuelto su texto, y leído se dice de qué trata con su cita; lo no leído se nombra solo con las palabras del texto leído (FR-002, FR-003). Ninguna regla de cita, de cobertura ni forma fija sale del cuerpo (contracts/skill-boe-legislacion-v0-3-1.md §2 y §3) |
| III · Tests primero y offline | «Aceptación e2e: no aplica» (abajo; Complexity Tracking). En la tarea de las lecturas, primero cambian los dos tests que las enumeran, en rojo mientras las evals no las llevan, y después las evals; las grabaciones las hace `grabar_datos` en la tarea `[datos]` anterior, y los tests leen lo grabado sin red. La skill la miden sus gates y controles de `make ci` y sus evals en el job |
| IV · Hexagonal y errores tipados | Ningún paquete ni código de salida cambia. Una lectura sin grabar sigue siendo la falta de la preparación (H26.1), y una lectura de la skill que falla sale con el código de hoy; lo ilegible, la regla genérica (`inesperado`, 1) |
| V · Simplicidad y dependencias | Ninguna dependencia, paquete, tipo, función ni fichero de código nuevo; ninguna referencia nueva. Cada cambio se traza a su FR en «Trazabilidad» |
| VI · Un binario, convenciones de agente | El binario, sus banderas, `--describe` y la tabla de comandos generada no cambian; `kitlegal skills install` instala la skill nueva como lo empotrado (`TestSkillsEmpotradas`) |
| VII · Grafo y privacidad | Nada del grafo cambia; la comprobación de cada lectura nueva es la de hoy, `kitlegal graph check <norma> <bloque>` |
| VIII · Skills primero | Protege `boe-legislacion` (spec, «Contexto») con dos defectos medidos: el cuerpo vuelve a decir cómo empieza la respuesta y la regla nueva va en el cuerpo, no en una referencia (FR-004); las evals y su juez deciden en el cierre (SC-007 a SC-011) |
| IX · Genericidad territorial | Ningún municipio ni comunidad; la regla 5 («Ningún caso especial para un territorio») no cambia |
| X · Genericidad por materia | La regla de leer lo que se nombra es por capacidad: no nombra materia ni norma de las evals 14, 15 y 17 (FR-004; quickstart, escenario 3). Las lecturas son datos de evals, y la guía, de autoría |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV, `internal/arch_test.go`)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | Sin cambios en `internal/core` |
| R2 · Solo `internal/httpx` importa `net/http` | Ningún import; la grabación usa el `TestGrabarEvals` de hoy, que pide por `internal/httpx` |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | Ningún import: `TestCacheDeLasSesiones` prepara y lee la caché con `PrepararSesion` y `app.Main`, como hoy |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Ningún `os.Exit` |
| R5 · Solo `render` escribe en stdout; logs a stderr | Ningún `fmt.Print*` ni `os.Stdout`: los tests fallan con `assert` |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios en `internal/graph` |
| R7 · Solo `internal/mcp` importa el SDK de MCP | Ningún import del SDK |
| «Sin red» de `arch_test.go` | Ningún paquete cambia de importaciones |

## Project Structure

### Documentation (this feature)

```text
specs/028-h31-1-boe-legislacion-habla/
├── spec.md
├── plan.md                                # este fichero
├── research.md                            # D1–D15, V1–V12, S1–S4
├── data-model.md                          # cuerpo, referencia, lectura, manifiesto, grabación, umbral
├── quickstart.md                          # diez escenarios, sin red ni modelo
├── contracts/
│   ├── skill-boe-legislacion-v0-3-1.md    # FR-001 a FR-008: texto, tabla de reglas, textos que salen
│   ├── lecturas-y-grabaciones.md          # FR-009 a FR-011: lecturas, manifiesto, grabación, tests
│   └── guia-y-changelog.md                # FR-012, FR-014 y la frase de CONTRIBUTING.md
├── checklists/
└── tasks.md                               # lo escribe /speckit-tasks
```

### Source Code (repository root)

```text
skills/boe-legislacion/
├── SKILL.md                               # FR-001 a FR-006: pasos 2 a 5, «Cómo se cita», «Redacción…», reglas 1, 2 y 7
└── references/
    ├── redaccion_modificada.md            # FR-006: «Lo demás de la comprobación», «Una lectura por bloque»
    └── buscar_la_norma.md                 # FR-006: «Si no aparece»
evals/boe-legislacion/
├── 14-trlrhl-impuestos-por-materia.yaml   # FR-009: a60, a92, a100, a104
├── 15-irpf-rendimientos-por-materia.yaml  # FR-009: a7, a19, a20, dadecimoctava
└── 17-ltaibg-plazo-por-materia.yaml       # FR-009: a24
internal/evals/
├── testdata/grabaciones.json              # FR-010 [datos]: bloques y para de dos entradas
├── testdata/boe.legislacion-consolidada/  # FR-010: ocho grabaciones (grabar_datos)
├── conjunto_test.go                       # FR-009, FR-011: el mapa de TestEvalsDelRepositorio/lecturas
└── preparar_test.go                       # FR-011: la tabla de TestCacheDeLasSesiones/lecturas
docs/SKILLS.md                             # FR-012: un párrafo de §1
CHANGELOG.md                               # FR-014
CONTRIBUTING.md                            # FR-006 (consecuencia): la frase de la regla 7
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`. La skill vive en `skills/`, sus evals en `evals/`, el
manifiesto y las grabaciones de las evals en `internal/evals/testdata/` (H9, H26.1), y los tests que enumeran las
lecturas en `internal/evals`. Ningún paquete ni fichero de código nuevo; `.github/workflows/evals.yml` e
`internal/evals/definicion.go` no cambian (research D10).

## Aceptación e2e

**Aceptación e2e: no aplica.** El hito no cambia ningún comportamiento del binario —ningún applet, verbo, bandera,
código de salida ni salida— y `kitlegal skills install` instala la skill nueva como cualquier fichero empotrado, que
`TestSkillsEmpotradas` ya compara byte a byte (research D13). Lo que entrega es el texto de una skill, datos de tres
evals y documentación. Su aceptación sin persona es el job de evals sobre la cabeza de la propuesta, que mide las tres
skills con los umbrales de hoy (SC-007 a SC-011), y la de la persona, el sondeo de la eval 15 antes de fusionar
(SC-012); los controles de `make ci` (SC-001 a SC-005) la preceden. El spec deja fuera evals nuevas.

## Controles mecánicos que este hito añade o toca

Capa 1 de la constitución («Gates»), en `make ci` salvo lo que se dice del job:

- **Objetivos de make**: ninguno nuevo ni cambiado. `make skills-check` y `make ci` ejecutan los de abajo.
- **Gates de H31, sin cambios, sobre la skill nueva** (`internal/app/skills_test.go`, `TestSkillsDelRepositorio`; y
  `internal/skills/sincronia_test.go`, `TestCuerpoDeLaSkill`): cuerpo de 11 000 caracteres como mucho (FR-005), cada
  referencia enlazada (FR-006), `normas-nombradas` y `sin-instrucciones-de-evals` (FR-004, FR-007).
- **Controles de contenido de `docs/SKILLS.md` §1, sin cambios** (`internal/evals/conjunto_test.go`,
  `TestEvalsDelRepositorio`): `prosa-de-la-skill`, `ordenes-para-powershell`, `linea-sin-consulta`,
  `formas-de-la-redaccion-fechada`, `avisos-de-la-skill`, `hallazgos-de-la-skill`,
  `lineas-de-aviso-de-las-referencias` y `ordenes-de-ejemplo` (FR-006, FR-007).
- **Tests que cambian** (FR-009, FR-011): el mapa escrito a mano de `TestEvalsDelRepositorio/lecturas`
  (`probarLasLecturasDeLasEvalsDelRepositorio`) y la tabla de `TestCacheDeLasSesiones/lecturas`
  (`probarLasLecturasDeLasEvals`), con la eval de cada fila (contracts/lecturas-y-grabaciones.md §4). Sin cambios, pero
  ahora sobre las lecturas nuevas: `TestEvalsDelRepositorio/grabado` y `TestLecturaDeLaEvalSinGrabar`.
- **Job**: `TestDefinicionDelJob` (`internal/evals/definicion_test.go`) sigue exigiendo que `timeout-minutes` cubra el
  peor caso de cada skill; los umbrales del informe del job, sin cambios.
- **Fixtures**: ocho grabaciones nuevas en `internal/evals/testdata/boe.legislacion-consolidada/`, del paso
  `grabar_datos` tras la tarea `[datos]` del manifiesto; ninguna en `schemas/`. Ninguna grabación existente cambia.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| SC-001 | `make ci` termina con 0 | Todos los de esta tabla y los de siempre | `ci:Makefile:ci` |
| FR-005, SC-002 | Cuerpo de `SKILL.md` ≤ 11 000 caracteres (10 993 con el contrato); con 11 001, rojo | Gate del cuerpo de H31 sobre las skills del repositorio y sobre copias de 11 000 y 11 001 | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, `ci:internal/skills/sincronia_test.go:TestCuerpoDeLaSkill` |
| FR-006 | 0 referencias de `references/` sin enlace desde `SKILL.md` | Gate de H31 | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-004, FR-007 | 0 instrucciones de las evals y 0 normas nombradas fuera de `data/normas.yaml`, en el cuerpo y en las referencias | `sin-instrucciones-de-evals`, `normas-nombradas` | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-006, FR-007, SC-001 | 0 defectos de los controles de contenido en `SKILL.md` y en cada referencia a mano, también 0 órdenes de ejemplo concretas que no salgan con 0 con `--offline` | Los ocho controles de `docs/SKILLS.md` §1 | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-009, SC-004 | Las lecturas de cada eval, exactamente las del contrato: 0 de más, 0 de menos, 0 motivos distintos | Subtest `lecturas`, con el mapa escrito a mano | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-010, FR-011, SC-004 | 0 consultas de las evals —las lecturas, sin fecha y con `9999-12-31`— sin su respuesta en la unión de grabaciones o que la caché preparada no sirva sin red | Subtest `grabado` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-011, SC-004 | Cada uno de los nueve bloques sale con 0 con `boe articulo … --offline` desde la caché de una sesión de su eval, sin fecha y con `--fecha 9999-12-31` | Subtest `lecturas` de la caché | `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones` |
| FR-010, FR-011 | Con una lectura sin grabar, sus 2 faltas (sin fecha y con `9999-12-31`) y la sesión sin preparar | Sin cambios (US3, escenario 3) | `ci:internal/evals/preparar_test.go:TestLecturaDeLaEvalSinGrabar` |
| SC-008 | 0 sesiones de `claude-sonnet-5-5` de las evals 14, 15 y 17 con una orden o llamada en `fuera_de_lo_grabado` que sale con 4 o 5 porque la caché de la sesión no sirve un bloque de sus `lecturas`; no cuenta la `articulos` que sale con 5 por otro bloque sin grabar de la misma norma, que se atribuye a ese bloque | Las tres filas de arriba: salen en rojo si la caché de una eval no sirve un bloque de sus `lecturas` sin red, sin fecha o con `9999-12-31`, o si falta su grabación, que es la causa que cuenta | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones`, `ci:internal/evals/preparar_test.go:TestLecturaDeLaEvalSinGrabar` |
| FR-013, SC-005 | `timeout-minutes` ≥ peor caso de cada skill: 21 120 s frente a 20 117 s de `boe-legislacion` | `fallosDelTope` sobre la definición del repositorio | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |
| SC-007 | `afirma_lo_no_leido` de `claude-sonnet-5-5`, todos los modos juntos, ≤ 0.028; si pasa, veredicto `fallo` y job en rojo | Umbral del juez, `decide: true` | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5` |
| SC-009 | Cada umbral que decide de `boe-legislacion` se cumple; el que no, veredicto `fallo` y job en rojo | Los umbrales de hoy, `decide: true`, en los dos modos | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`, `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`, `evals:boe-legislacion:redaccion_sin_fechar:claude-sonnet-5-5:orden`, `evals:boe-legislacion:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`, `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`, `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados`, `evals:boe-legislacion:duracion_de_las_sesiones:orden`, `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`, `evals:boe-legislacion:duracion_del_juez:orden`, `evals:boe-legislacion:duracion_del_juez:herramienta` |
| SC-011 | `jurisprudencia`: cada umbral que decide se cumple (veredicto `aprobado`); si no, job en rojo | Los umbrales de hoy, `decide: true` | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`, `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`, `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`, `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`, `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`, `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`, `evals:jurisprudencia:duracion_del_juez:orden`, `evals:jurisprudencia:duracion_del_juez:herramienta` |
| SC-011 | `legal-core`: cada umbral que decide se cumple (veredicto `aprobado`); si no, job en rojo | Los umbrales de hoy, `decide: true` | `evals:legal-core:sin_activar:claude-sonnet-5-5:orden`, `evals:legal-core:sin_activar:claude-sonnet-5-5:herramienta`, `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:orden`, `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:herramienta`, `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:orden`, `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`, `evals:legal-core:vigencia_sin_declarar:claude-sonnet-5-5:orden`, `evals:legal-core:vigencia_sin_declarar:claude-sonnet-5-5:herramienta`, `evals:legal-core:redaccion_sin_fechar:claude-sonnet-5-5:orden`, `evals:legal-core:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`, `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:orden`, `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:herramienta` |

Sin fila, porque no son un umbral que mida `make ci` ni el job: SC-010 (`cuenta_su_proceso` se publica por modo con
`decide: false`, como hoy; hacerlo decidir pide medir antes su juez, ADR 0037); SC-012 (el sondeo, de una persona);
SC-003 y SC-013 (los coteja la revisión final con la tabla de contracts/skill-boe-legislacion-v0-3-1.md §2 y con
`LEEME.md`); SC-006 (la entrada de `CHANGELOG.md` y que el job mida las tres skills, que hace el workflow); y los
criterios de FR-004 y FR-008 sobre el diff de la rama contra `main` (quickstart, escenarios 3 y 4; research D14). Los
nombres de SC-009 y SC-011 son los del cierre de H31 sin `habla_en_primera_o_segunda_persona`, que solo va cuando la
propuesta cambia una descripción y aquí no cambia ninguna (FR-008; research V10, S4). `fuera_de_lo_grabado` no decide
(H26.1), y hacerlo decidir queda fuera del hito: SC-008 cuenta solo la orden o llamada que sale con 4 o 5 porque la
caché de la sesión no sirve un bloque de las `lecturas` de su eval. Los tests de su fila prueban que la caché de cada
eval sirve sin red cada bloque de sus `lecturas`, sin fecha y con `9999-12-31`, y que una lectura sin grabar para la
preparación de la sesión. No prueban que una orden `articulos` que pide, con esos bloques, otro de la misma norma sin
grabar salga con 0: sale entera con 5, como `boe_articulos BOE-A-2004-4214 a57 a59` en la sesión
`14-trlrhl-impuestos-por-materia-herramienta-claude-sonnet-5-5-03` del cierre de H31, con `a59` grabado y `a57` no
(research V13). Esa salida se atribuye al bloque sin grabar, no cuenta para SC-008, y `fuera_de_lo_grabado` la publica
sin decidir.

## Uso, de fuera adentro (criterio de uso, ADR 0028)

- **El cuerpo** (FR-001 a FR-006): lo carga el agente entero, una vez por pregunta que activa la skill: 10 993
  caracteres, unos 4 860 tokens a 2,26 caracteres por token (ADR 0046), uno más que hoy. No crece con el uso: lo
  escribe una persona y el gate lo acota.
- **Las lecturas que pide la regla nueva** (FR-002, FR-003): una por precepto que la respuesta va a nombrar y no ha
  leído, si lo remite un bloque leído en el paso 2 o su id solo salió en un índice; cada una, `kitlegal boe articulo
  <norma> <bloque> --json && kitlegal graph check <norma> <bloque> --json` (o `boe_articulo` y `graph_check`), o una
  `articulos` con su comprobación para varios de una norma. Forma: el sobre de `articulo` de hoy, `{ok, fuente, url,
  fecha_consulta, hash, data}` con `data` = `{norma, bloque, titulo, tipo, fecha_version, fecha_vigencia,
  norma_modificadora, texto, hash_texto, avisos, url, url_eli}`. Bytes medidos de sobres de esas normas (research V11):
  `a17` de la Ley 35/2006, 9 345; `a18`, 4 803; `a59` y `a15` de Haciendas Locales, 1 302 y 1 708; `a20` y `a24` de la
  Ley 19/2013, 2 432 y 2 690. En la eval 15, el art. 17 lleva a cinco (`a7`, `a18`, `a19`, `a20`, `dadecimoctava`):
  cinco sobres más en esa pregunta, del orden de las decenas de KB juntos; los de los bloques nuevos se sabrán al
  grabarlos (research V11). Lo acotan la respuesta y el salto: solo lo que la respuesta nombra y nada de lo que remiten
  esas lecturas. Con meses de uso —cientos de normas, miles de bloques consultados, la mayoría hace más de una semana—
  no cambia: cada lectura pide un bloque por su id y devuelve ese bloque, y su comprobación lleva norma y bloque, así que
  da como mucho una entrada de ese bloque, nunca lo acumulado (la orden sin argumentos sigue prohibida, paso 2).
- **Sus señales**: los avisos de vigencia del bloque, que vienen con la norma y se trasladan en cada lectura que los
  trae; y `⚠ REDACCIÓN MODIFICADA:`, solo si la redacción del bloque cambió desde una lectura anterior, que deja de darse
  al leerlo porque el grafo guarda la nueva (paso 2, «lee cada bloque una vez»). Si la comprobación trae una entrada
  `fuente-caducada`, o de un bloque sin `version-obsoleta`, no va en la respuesta (`redaccion_modificada.md`). Ninguna
  señal nueva.
- **Las `lecturas` y sus grabaciones** (FR-009 a FR-011): las pide la preparación de la caché de cada sesión de
  `boe-legislacion`, una vez por sesión: nueve bloques más, cada uno sin fecha y con `9999-12-31`, sin red. Ocho
  grabaciones nuevas en el repositorio. No crecen con el uso de kitlegal, solo con lo que se escribe en `evals/`.
- **La duración del trabajo de evals** (FR-013): el peor caso, 20 117 s, no cambia (research D10); la duración medida de
  cada modo en el cierre de H31 fue 410 y 348 s frente al objetivo de 900 s que decide; la regla nueva añade, a la
  sesión cuya respuesta nombra preceptos remitidos o salidos de un índice, una lectura por precepto —cinco en la eval
  15—, y las sesiones corren de cinco en cinco. Es un cálculo, no una medida (research S3); si una sesión no cabe en
  240 s, el informe lo dice con «tope de 240 s agotado (código 124)», y si un modo pasa de 900 s, su umbral
  `duracion_de_las_sesiones` decide (SC-009).
- **`docs/SKILLS.md` §1** (FR-012): un párrafo de unos 700 caracteres, que lee quien escribe o reparte una skill, una vez
  por skill. Deja de valer cuando otra evidencia mida otra cosa.
- **`CHANGELOG.md`** (FR-014): una entrada, que lee quien actualiza kitlegal, una vez por versión.

## Decisiones

Las del diseño están en research.md: la vuelta de las reglas (D1, D2), FR-002 en el paso 3 (D3), el salto (D4), el
índice (D5), el sitio (D6), la regla 7 y la numeración (D7), el orden de las tareas de datos (D8), la caché de cada
eval en el test (D9), el tope (D10), la guía (D11), el CHANGELOG (D12), la aceptación (D13), FR-004 y FR-008 sin
control nuevo (D14) y el prototipo fuera del repositorio (D15).

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| Primera viñeta del paso 4 con B48, B50, B52 y B53; los cuatro ejemplos del paso 5 | FR-001, SC-003 |
| Segunda frase del paso 3 (nombrar solo lo que `articulo` o `articulos` ha devuelto; lo remitido o salido de un índice, leído antes con los pasos 1 y 2, o no nombrado; leído, de qué trata) | FR-002, FR-004 |
| Frase del salto en el paso 5 | FR-003, FR-004 |
| Las siete salidas de contracts/skill-boe-legislacion-v0-3-1.md §3, el enlace de `redaccion_en_una_fecha` en el paso 2 y los cambios de `redaccion_modificada.md` y `buscar_la_norma.md` | FR-005, FR-006 |
| Regla 8 → 7, la regla 2 y la frase de `CONTRIBUTING.md` | FR-006 (consecuencia de sacar la regla 7) |
| Las nueve `lecturas` con su `motivo` | FR-009 |
| Bloques y `para` de dos entradas del manifiesto; las ocho grabaciones | FR-010 |
| El mapa de `TestEvalsDelRepositorio/lecturas` y la tabla de `TestCacheDeLasSesiones/lecturas` | FR-009, FR-011, SC-004 |
| El párrafo de `docs/SKILLS.md` §1 | FR-012, SC-013 |
| El recálculo del peor caso, sin cambio de tope | FR-013, SC-005 |
| La entrada de `CHANGELOG.md` | FR-014, SC-006 |

Lo que no se construye: ningún control nuevo para el diff de FR-004 y FR-008 (D14), ninguna referencia nueva (D6),
ningún cambio del tope del trabajo ni del de una sesión (D10, Clarificación Q1), y nada para ficheros del kit, de la
caché o de las grabaciones tocados a mano: regla genérica, defecto `inesperado`, código 1 (spec, «Edge Cases»).

## Datos externos

Una fuente, `boe.legislacion-consolidada`, con fila revisada en `docs/SOURCES.md` de `main` («Revisado» 2026-09-13).
Manifiesto: `internal/evals/testdata/grabaciones.json` (`"fuente": "boe.legislacion-consolidada"`). Test de grabación:
`TestGrabarEvals` (`internal/evals/grabacion_test.go`, `//go:build grabacion`), que ya existe y pide con
`internal/httpx`. Respuestas: `internal/evals/testdata/boe.legislacion-consolidada/`. Material de origen en
`$KITLEGAL_EVIDENCIAS` (`evidencias/h31-1/`): ninguno, porque ningún fichero de `data/` sale de lo grabado. La tarea
`[datos]` escribe solo el manifiesto; graba `grabar_datos`, sin modelo (contracts/lecturas-y-grabaciones.md §2 y §3).
Ninguna fuente sin fila revisada entra en el hito.

## Orden de implementación (de dentro afuera)

Cada paso deja `make ci` en verde.

1. **`[datos]`**: los ocho bloques y su `para` en el manifiesto (FR-010). Después, `grabar_datos` graba los ocho.
2. **Las lecturas**: las nueve entradas en las evals 14, 15 y 17, el mapa de `TestEvalsDelRepositorio/lecturas` y la
   tabla de `TestCacheDeLasSesiones/lecturas` (FR-009, FR-011).
3. **La skill**: el cuerpo de v0.3.1, las dos referencias a mano, la frase de `CONTRIBUTING.md` y la entrada de
   `CHANGELOG.md` (FR-001 a FR-008, FR-014), con `make skills-sync` y `make skills-check` en verde.
4. **La guía**: el párrafo de `docs/SKILLS.md` §1 (FR-012).

FR-013 no tiene tarea: el recálculo es el de research D10 y `TestDefinicionDelJob` ya lo comprueba.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|---|---|---|
| Principio III, «cada hito empieza por el test e2e (`testscript`)»: Aceptación e2e no aplica | El hito no cambia ningún comportamiento del binario: entrega el texto de una skill, datos de tres evals y documentación, que miden las evals del job y el sondeo (research D13) | Un guion que instale la skill y busque una frase fijaría en la suite congelada un texto que diseña el hito y repetiría `TestSkillsEmpotradas`; el spec deja fuera evals nuevas, y las de hoy son la aceptación en el job, como en H24, H26 y H31 |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- **precheck**: `plan.md` y `research.md` existen; ninguna marca de aclaración pendiente; «## Constitution Check» con
  una fila por cada principio de I a X con su numeral en la primera celda; «Aceptación e2e: no aplica» con su motivo;
  «## Controles de umbral» con filas cuyos requisitos define el spec y cuyo «Dónde» es `ci:` o `evals:`.
- **a · constitution_check**: cada fila dice lo que hace este plan; las siete reglas y la de «sin red».
- **b · dependencias**: ninguna nueva.
- **c · reglas_dependencia**: ningún import nuevo en ningún paquete.
- **d · errores_exit_codes**: ningún código cambia; la falta de una lectura sin grabar, la de H26.1.
- **e · tests_primero**: «no aplica» con motivo; los dos tests que cambian, los que siguen igual y los fixtures, arriba.
- **f · alcance**: solo FR-001 a FR-014; `CONTRIBUTING.md`, por la regla 7 que sale (D7).
- **g · sin_atajos**: ningún `//nolint`, `t.Skip`, TODO ni error silenciado.
- **h · mejor_alternativa**: cada decisión de research con su descartada.
- **i · afirmaciones_verificadas**: V1 a V12, ejecutadas en esta sesión; S1 a S4, supuestos.
- **j · quickstart_ejecutable**: órdenes de lectura, probadas hoy las que no dependen de la entrega; solo el escenario
  9 escribe, y declara qué.
- **k · datos_externos**: «Datos externos».
- **l · autonomía**: nada a mitad del run; el sondeo, la fusión y la release, de una persona fuera del run.
- **m · uso**: «Uso, de fuera adentro», con los bytes medidos de V11.
- **n · proporcionalidad**: «Trazabilidad».
- **o · controles_de_umbral**: la tabla de arriba, una fila por cada requisito con umbral.
