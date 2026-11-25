# Implementation Plan: H46 · Los cuerpos de las skills de capacidad, con sitio para las materias (ADR 0063)

**Branch**: `044-h46-los-cuerpos-de` | **Date**: 2026-11-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/044-h46-los-cuerpos-de/spec.md`

## Summary

Los cuerpos de `normativa`, `plazos-y-territorio` y `jurisprudencia` dejan sitio a las materias: 1400 caracteres para
las líneas de los paquetes, comprobados por `make ci`. Tres cambios, en este orden. La tabla de comandos que genera
`RenderizarTabla` pasa a tres columnas, sin la de lo que devuelve en `data` ni las líneas del sobre y de las banderas
comunes, que da cada llamada; lo que lee de `--describe` y sus errores no cambian (research D1). Los textos de `skills/`
son los del prototipo `f89093dd`, byte a byte, con lo que los fija en Go y en los guiones al día, en una tarea `[datos]`
(D5 a D7). Y `defectosDelCuerpo` da un defecto más a la skill con las marcas de los paquetes cuyo cuerpo, sin las líneas
entre ellas, pasa de 9600 caracteres, que `casosDePresupuestos` fija en sus límites (D2, D3). El prototipo, medido en
esta sesión en un clon, deja los cuerpos en 9498, 9483 (sin la línea de lo administrativo) y 9473, pasa los tests de los
tres paquetes, los guiones de H41, H41.1 y H45 y el lint, y cada mutante del defecto nuevo pone su subtest en rojo
(research P1). Ningún applet devuelve otra cosa (FR-019).

## Technical Context

**Language/Version**: Go 1.27 (`go1.27.4`; research V1).

**Primary Dependencies**: las fijadas, ninguna nueva: `stretchr/testify` y `rogpeppe/go-internal` v1.16.0 (testscript:
`exec`, `grep`; V2) en los tests; biblioteca estándar (`unicode/utf8`, `bytes`, `strings`) en `internal/skills`.

**Storage**: seis ficheros de `skills/` pasan a los de `f89093dd`; ningún fichero de `data/`, `evals/`, `schemas/` ni
`laboratorio/` cambia. En `testdata/`, dos guiones de H41 y H41.1 cambian, con la copia congelada de H41.1 (tarea
`[datos]`), y la activación de la suite añade sus dos guiones. Nada en la caché ni en el grafo.

**Testing**: `go test -race` en `make ci`; guiones testscript de `TestEntregaDelHito` sobre lo que instala `kitlegal
skills install`; sin red ni modelo. Las evals del cierre, sin cambios (D13).

**Target Platform**: Linux y macOS, donde corre `make ci`.

**Project Type**: CLI multicall con skills empotradas; el cambio es de `internal/skills` (generador de la tabla y
presupuestos), de los textos de tres skills y de sus tests.

**Performance Goals**: ninguno del spec. En el clon: el paquete `internal/app` con solo
`TestSkillsDelRepositorio/presupuestos`, de 1,19 s a 1,26 s; `h45-paquete-de-mas`, 8,21 s; los dos guiones nuevos,
0,02 s cada uno (P1).

**Constraints**: ningún `--describe`, esquema, sobre, código de salida ni forma de `data` distinto (FR-019); `skills/`
igual a `f89093dd` (FR-010); cada tarea deja `make ci` en verde (FR-020; D9).

**Scale/Scope**: 2 ficheros de Go de producción (`internal/skills/comandos.go`: la cabecera, la fila y el cierre de la
tabla y 8 identificadores sin exportar retirados; `internal/skills/sincronia.go`: 1 constante y el defecto); 6 ficheros
de test de Go (`internal/skills/comandos_test.go`, `sincronia_test.go`, `vertical_de_ensayo_test.go`,
`remisiones_test.go`; `internal/app/skills_test.go`; `internal/evals/laboratorio_familias_test.go`); 6 ficheros de
`skills/`; 3 guiones cambiados y 2 nuevos; `docs/SKILLS.md`, `CHANGELOG.md` y el registro `mutantes.md`.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio o regla | Cómo lo cumple este hito |
|---|---|
| I · Fuentes públicas y frontera humana (INNEGOCIABLE) | Ninguna petición HTTP ni fuente nueva: el hito no graba nada («Datos externos»); los guiones instalan las skills sin red. Ninguna identidad, POST ni CAPTCHA; de lo que `jurisprudencia` dice de la consulta al CENDOJ, lo que se pide según cómo se dio la referencia solo cambia de sitio (FR-011), y ninguna regla de la frontera sale del cuerpo (FR-013) |
| II · Nada sin cita ni fuente | El sobre `{ok, fuente, url, fecha_consulta, hash, data}` no cambia (FR-019): la tabla deja de describirlo, y cada llamada lo sigue devolviendo. Ninguna línea `⚠`, forma fija, regla invariante ni orden de leer sale de un cuerpo (FR-013), y las tres líneas de la cobertura siguen en el de `plazos-y-territorio` |
| III · Tests primero y offline | T001 escribe los dos guiones de la suite, que caen hoy por una aserción (contracts/aceptacion.md; P1) y quedan congelados. Los tests de la tabla, de los presupuestos, del producto y de las remisiones se cambian con su código, y sus rojos quedan en `mutantes.md`. Todo offline en `make ci`; ningún esquema cambia; `internal/core` no cambia; el código que se retira es el que la tabla deja sin uso, y el que se añade lo ejercen los casos de `presupuestos` |
| IV · Arquitectura hexagonal con reglas ejecutables | Solo `internal/skills` cambia en producción, sin imports nuevos ni `panic`. El defecto nuevo es un `DefectoDeSkill` como los otros dos del cuerpo, que `skills-check` presenta; ningún código de salida nuevo |
| V · Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia, paquete, tipo exportado ni opción nuevos; cada mecanismo con su requisito («Trazabilidad»), y lo que la tabla deja sin uso se retira (D1). Rechazados (research): dejar de leer `Devuelve`, contar en `paquetes.regenerar`, restar las marcas, la regla en toda skill, un campo o un export para las marcas, reescribir los textos a mano, movimientos en lugar de cambios, aserciones sobre todo el fichero o que cuentan tablas |
| VI · Un binario, convenciones de agente | Ninguna orden, bandera, argumento ni `--describe` cambia; la tabla de cada skill se sigue generando de `--describe` (FR-004), con la orden y la herramienta de cada fila |
| VII · Grafo y privacidad | Nada toca el grafo, la caché ni datos de nadie |
| VIII · Skills primero; el binario es la herramienta | Es el objeto del hito: en el cuerpo, lo que toda respuesta necesita, y el detalle de una clase de pregunta en la referencia que esa clase lee sin condición (ADR 0063); los presupuestos los comprueba `make ci` (FR-005). Las respuestas no cambian: lo deciden las evals del cierre de todas las skills (SC-008; D13) |
| IX · Genericidad territorial, validación local | Ningún caso por municipio: `municipios.md` gana texto genérico (cómo se presenta cada dato, la herramienta o la orden), y la cobertura se sigue declarando con sus tres líneas |
| X · Genericidad por materia, validación administrativa | Es su porqué: que la materia siguiente entre como paquete sin tocar el protocolo. La regla vale para toda skill con las marcas de los paquetes, sin nombrar ninguna materia ni skill (FR-005, FR-006); no es una abstracción nueva, sino un presupuesto del cuerpo. La ejercen la línea de lo administrativo, en `plazos-y-territorio`, y la del paquete de la copia de `h45-paquete-de-mas`, en `normativa` (D3); la del paquete tributario la comprueba la persona antes de lanzar H42 |
| R1 · `internal/core/**` no importa el kernel, adaptadores ni entrada y salida | `internal/core` no cambia |
| R2 · solo `internal/httpx` importa `net/http` | Ningún import nuevo |
| R3 · solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | Ningún import nuevo |
| R4 · solo `internal/cli` y `cmd/` llaman a `os.Exit` (`forbidigo`) | Ningún `os.Exit` nuevo |
| R5 · solo `internal/render` escribe en stdout; logs con `log/slog` a stderr (`forbidigo`) | `internal/skills` devuelve defectos y bytes, sin escribir en stdout; `skills-check` los presenta como hoy |
| R6 · `internal/graph` no importa `internal/source` ni `internal/render` | `internal/graph` no cambia |
| R7 · solo el paquete del servidor importa el SDK de MCP | Ningún import nuevo |
| Garantía sin red de `TestArquitectura` (`paquetesSinRed`, `ficherosSinRed`) | `internal/skills` no gana ninguna importación: usa `unicode/utf8`, `bytes`, `fmt` y `strings`, que ya importa; `vertical_de_ensayo_test.go` pierde `unicode/utf8` (P1: lint con 0 issues) |

Re-evaluado tras la fase 1: sin violaciones sin justificar; la única desviación, en Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/044-h46-los-cuerpos-de/
├── spec.md
├── plan.md                              # este fichero
├── research.md                          # P1, D1-D14, V1-V16, datos externos, supuestos
├── data-model.md                        # el cuerpo y su región, la tabla, los casos, las declaraciones
├── quickstart.md                        # Q1-Q9
├── contracts/
│   ├── tabla-de-comandos.md             # FR-001 a FR-004: los bytes, lo que no cambia, lo que se retira
│   ├── sitio-de-las-materias.md         # FR-005 a FR-009: el defecto, sus casos y TestVerticalDeEnsayo
│   ├── textos-y-controles.md            # FR-010 a FR-016: los textos, las declaraciones, las remisiones, los guiones
│   └── aceptacion.md                    # FR-021: los dos guiones, medidos
├── aceptacion/                          # T001 (congelada)
│   ├── tabla-de-tres-columnas.txtar
│   └── textos-en-su-referencia.txtar
├── mutantes.md                          # registro de SC-003 a SC-006; lo escribe T006
└── tasks.md                             # /speckit-tasks
```

### Source Code (repository root)

```text
internal/skills/
├── comandos.go                          # RenderizarTabla de tres columnas (contracts/tabla-de-comandos.md)
├── comandos_test.go                     # TestRenderizarTabla con los bytes nuevos
├── sincronia.go                         # maximoSinLosPaquetes y el defecto (contracts/sitio-de-las-materias.md §1)
├── sincronia_test.go                    # regionDeAlfa
├── vertical_de_ensayo_test.go           # sin comprobarElSitioDeOtroPaquete (§3)
└── remisiones_test.go                   # remisionesDelPlan (contracts/textos-y-controles.md §3)
internal/app/
├── skills_test.go                       # casosDePresupuestos (§2); el fixture de «control»
└── testdata/script/
    ├── h41-skills-instaladas.txtar      # [datos] (contracts/textos-y-controles.md §4)
    ├── h41-1-skills-instaladas.txtar    # [datos]
    ├── h46-tabla-de-tres-columnas.txtar # activación de la suite
    └── h46-textos-en-su-referencia.txtar
internal/evals/
└── laboratorio_familias_test.go         # declaracionesDelProducto (contracts/textos-y-controles.md §2)
skills/normativa/SKILL.md
skills/plazos-y-territorio/{SKILL.md,references/municipios.md}
skills/jurisprudencia/{SKILL.md,references/consulta_preparada.md,references/documento_aportado.md}
specs/035-h41-1-plazos-y-territorio/aceptacion/skills-instaladas.txtar   # [datos], igual a su guion
docs/SKILLS.md
CHANGELOG.md
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquetes ni ficheros de Go nuevos: la regla vive
con los otros presupuestos del cuerpo en `internal/skills`, y los tests junto a lo que prueban.

## Diseño

1. **La tabla de tres columnas** (FR-001 a FR-004; contracts/tabla-de-comandos.md; research D1, D11): `columnasDeLaTabla`,
   `filaDelVerbo` y el cierre de `RenderizarTabla`; `mismoSobreYBanderas` y `LeerDescripcionDeVerbo`, igual; fuera lo
   que queda sin uso; los bytes nuevos en `TestRenderizarTabla`, `regionDeAlfa` y el fixture de «control»; `make
   skills-sync` rehace las tres tablas.
2. **Los textos y lo que los fija** (FR-010 a FR-016; contracts/textos-y-controles.md; D5 a D7), en una tarea `[datos]`:
   `git checkout f89093dd -- skills/`, las 15 declaraciones del producto, las dos frases de las remisiones y los tres
   guiones de H41 y H41.1.
3. **El sitio de las materias** (FR-005 a FR-009; contracts/sitio-de-las-materias.md; D2 a D4): `maximoSinLosPaquetes`
   y el defecto en `defectosDelCuerpo`; `casosDePresupuestos` con y sin las marcas; `TestVerticalDeEnsayo` sin la línea
   de más.
4. **La guía y el registro** (FR-017, FR-018; D12): `docs/SKILLS.md` §1 va con el punto 1; §3 —con las cifras de los
   tres cuerpos, que no cambian desde el punto 2—, §5 y «La vertical de ensayo», con el 3; `CHANGELOG.md`, al final.
5. **El registro de mutantes** (SC-003 a SC-006; D10), en un clon del temporal, al final.

Orden de las tareas (FR-020; D9): T001 la suite; T002 = 1; T003 = 2; T004 = 3; T005 `CHANGELOG.md`; T006 = 5. Cada una
termina con `make ci` en verde (`scripts/workflow/verificar.sh`).

## Uso, de fuera adentro

Ninguna salida del binario cambia ni se añade (FR-019), y ninguna `SKILL.md` pide una invocación más: las mismas órdenes
y herramientas por pregunta. Cambia lo que el agente carga y lo que lee quien cambia un cuerpo. Nada sale de la caché ni
del grafo: con meses de uso diario —cientos de normas y miles de bloques consultados— todo mide lo mismo; solo crece con
los paquetes, una línea de 52 + 3n caracteres cada uno, que FR-005 acota a 1400 por cuerpo.

| Salida | Quién la consume y cuántas veces | Tamaño (research P1) | Cuándo deja de darse |
|---|---|---|---|
| Cuerpo de `normativa`, con su tabla | el agente, una vez por pregunta en que se activa, antes de la primera llamada | 9498 caracteres (10 750 en `main`); su tabla, 1436 (2673) | — (es el protocolo); lo que la tabla deja de dar lo trae cada llamada |
| Cuerpo de `plazos-y-territorio` y `municipios.md` | el agente: el cuerpo una vez por pregunta, y la referencia una vez, antes de la primera orden | 9577 + 4666 = 14 243 (14 559 en `main`) | — |
| Cuerpo de `jurisprudencia`, `consulta_preparada.md` y `documento_aportado.md` | el agente: el cuerpo una vez; `consulta_preparada.md` antes de `cita_preparar`; `documento_aportado.md` antes de cada cotejo, k por pregunta | 9473 + 2698 + 3208·k (10 842 + 1997 + 2756·k en `main`): 216 menos con k = 1, 236 más con k = 2 y 452 más por cada documento siguiente, acotado por los documentos que la persona trae a esa pregunta | — |
| El defecto de FR-005 | quien cambia un cuerpo de capacidad o escribe un paquete, persona o run, una vez por `make ci` o `make skills-check` | una línea de 93 a 103 bytes por skill que se pasa, tres como mucho | cuando el cuerpo vuelve a 9600 o menos sin las líneas de los paquetes; con los textos del hito, nunca |

Ejemplos con sus bytes: la tabla de `jurisprudencia` en contracts/tabla-de-comandos.md §1, y el defecto en
contracts/sitio-de-las-materias.md §1.

## Trazabilidad

| Mecanismo | Requisito |
|---|---|
| `columnasDeLaTabla` y `filaDelVerbo` de tres celdas; el cierre de `RenderizarTabla` sin las dos líneas | FR-001 |
| `mismoSobreYBanderas` en cada verbo; `LeerDescripcionDeVerbo` y `Devuelve`, `Sobre`, `Banderas`, sin cambios | FR-002 |
| retirar `Devuelve.texto`, `lineaDelSobre`, `lineaDeLasBanderas`, `comoCodigo` y sus cuatro constantes | FR-001 (código que deja sin uso; `unused`, V3) |
| bytes de `TestRenderizarTabla`, `regionDeAlfa` y el fixture de «control» | FR-003 |
| `make skills-sync` en T002 | FR-004 |
| `maximoSinLosPaquetes` y el defecto en `defectosDelCuerpo`, solo con las marcas en pareja | FR-005, FR-006, FR-007 |
| `casosDePresupuestos` con `p`, y las dos constantes de las marcas en el test | FR-008 |
| fuera `comprobarElSitioDeOtroPaquete` y su import | FR-009 |
| `git checkout f89093dd -- skills/` | FR-010 a FR-013 |
| las 15 declaraciones del producto | FR-014 |
| las dos frases de `remisionesDelPlan` | FR-015 |
| los `grep` de `h41-skills-instaladas`, `h41-1-skills-instaladas` y su copia congelada | FR-016 |
| `docs/SKILLS.md` §1, §3, §5 y «La vertical de ensayo» | FR-017 |
| la entrada de `CHANGELOG.md` | FR-018 |
| las seis tareas en su orden, la tercera `[datos]`; la documentación con lo que describe, y `CHANGELOG.md` al final | FR-020 |
| los dos guiones de la suite | FR-021, SC-006 |
| `mutantes.md` | SC-003, SC-004, SC-005, SC-006 |

## Controles mecánicos que el hito añade o toca

Según «Gates» de la constitución, capa 1:

- **`make ci`**, sin objetivos nuevos: `make test` corre los tests cambiados y `TestEntregaDelHito` con los cinco
  guiones del hito y de H41, H41.1 y H45; `make skills-check` da el defecto nuevo y la deriva; `make lint`, con la
  configuración de hoy, sin `//nolint` ni `t.Skip` (P1: `0 issues`).
- **El defecto nuevo de `skills-check`**: `<skill>: SKILL.md: cuerpo sin las líneas de los paquetes de <n> caracteres
  (máximo 9600)` (FR-005), con sus casos en `TestSkillsDelRepositorio/presupuestos`.
- **Tests de contrato**: ningún esquema ni sobre cambia; los tests de cada applet, sin tocar (FR-019).
- **Fixtures** (tarea `[datos]`): `internal/app/testdata/script/h41-skills-instaladas.txtar`,
  `h41-1-skills-instaladas.txtar` y `specs/035-h41-1-plazos-y-territorio/aceptacion/skills-instaladas.txtar`. Ninguna
  grabación ni `evidencias/`.
- **La suite congelada**: `specs/044-h46-los-cuerpos-de/aceptacion/*.txtar`, con su «rojo primero» (V4).
- **Test de arquitectura**: `TestArquitectura`, sin cambios. **Exit codes**: ninguno nuevo.

**Mutantes** (`mutantes.md`; quickstart Q9): cada uno en un clon del temporal; el registro dice la orden, lo esperado y
lo visto.

| Control | Mutante | En rojo |
|---|---|---|
| FR-005, SC-003 | `maximoSinLosPaquetes = 9601` | `presupuestos`: 6 casos (9601 y 11 001 de las tres skills con marcas) |
| FR-005, SC-003 | `maximoSinLosPaquetes = 9599` | `presupuestos`: 9 casos (además, los de 9600) |
| FR-005, SC-003 | la regla que no se aplica (`sincronia.go` de `main` con los tests del hito) | `presupuestos`: 6 casos |
| FR-005 | sin restar las líneas de los paquetes | `presupuestos/plazos-y-territorio`: 3 casos |
| FR-005 | el defecto delante del de 11 000 | `presupuestos/*/once-mil-y-un-caracteres` de las skills con marcas |
| FR-006 | la regla también sin las marcas | `presupuestos/consulta-juridica`: `once-mil-caracteres` y `once-mil-y-un-caracteres` |
| FR-001, FR-003 | la tabla de cuatro columnas (`comandos.go` de `main`) | `TestRenderizarTabla`, `TestRegenerarYComparar`, `TestCuerpoDeLaSkill` |
| FR-002 | `RenderizarTabla` sin la llamada a `mismoSobreYBanderas` | `TestRenderizarTabla`, casos `banderas-distintas` y `sobre-distinto` |
| SC-004, FR-014 | cada una de las 15 entradas deshecha en `skills/`, de una en una | `TestSkillsDelProductoDesdeControl` |
| SC-005, FR-015 | la frase de `main` en `remisionesDelPlan`, en cada una de las dos | `TestRemisionesDelRepositorio` |
| SC-006 | los dos guiones sobre los `skills/` de `main` | `h46-tabla-de-tres-columnas` y `h46-textos-en-su-referencia` |
| FR-016 | la orden de `main` en `h41-1-skills-instaladas.txtar`; la línea de `main` en `h41-skills-instaladas.txtar` | `h41-1-skills-instaladas`; `h41-skills-instaladas` |

Ya se corrieron en esta sesión, sobre el prototipo en un clon (research P1): los dos topes; la regla desactivada, con
`&& false` en su condición y no con el `sincronia.go` de `main`; sin restar las líneas; la regla sin las marcas; la tabla
de cuatro columnas; el `SKILL.md` de `normativa` de `main`, que deshace su única entrada; el `remisiones_test.go` de
`main`, con las dos frases a la vez; y los dos guiones sobre los `skills/` de `main`. Los demás, y cada uno de los de
FR-014 y FR-015 por separado, los corre T006.

**Aceptación e2e:** dos guiones testscript que T001 escribe en `specs/044-h46-los-cuerpos-de/aceptacion/` y que quedan
congelados (contracts/aceptacion.md): `tabla-de-tres-columnas.txtar` describe US1 (escenario 1) y cubre FR-001, FR-021 y
SC-006; `textos-en-su-referencia.txtar` describe US3 (escenarios 1 a 6) y cubre FR-010, FR-011, FR-012, FR-021 y SC-006.
El workflow los activa como `internal/app/testdata/script/h46-tabla-de-tres-columnas.txtar` y
`h46-textos-en-su-referencia.txtar`. Las demás historias no tienen comportamiento visible del binario y su aceptación
está en otro control: US2, el defecto de FR-005, que no da ninguna orden de `kitlegal` (FR-021), en
`casosDePresupuestos` y sus mutantes; US4, en `TestSkillsDelProductoDesdeControl`, `TestRemisionesDelRepositorio` y los
guiones `h41-`; US5, en las evals del cierre (SC-008).

**Datos externos:** ninguno. No hay manifiesto `grabaciones.json` ni test `TestGrabar*`; el paso `grabar_datos` no tiene
qué hacer, y los textos salen de `f89093dd`, un objeto del repositorio (research, «Datos externos»).

**Autonomía:** ninguna tarea pide nada a una persona ni publica la rama, abre la propuesta de cambio o mide CI y evals
remotas; el cierre lo hace el workflow tras la revisión final. La medida previa del ADR 0063 y la comprobación del
prototipo de H42 son de la persona, fuera de este run (spec, «Fuera de alcance»).

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| SC-001, FR-008 | `make ci` con 0 fallos, con `h45-paquete-de-mas` y su paquete de más | `make ci`, en rojo si cae cualquier test, guion o comprobación | `ci:Makefile:ci` |
| SC-002, FR-005, FR-007 | 0 skills con las marcas cuyo cuerpo, sin las líneas de los paquetes, pasa de 9600 caracteres; 0 cuerpos de más de 11 000 | el defecto de FR-005 y el de 11 000 de `skills-check` | `ci:Makefile:skills-check`, `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| SC-002, FR-004 | 0 ficheros de `skills/` que `make skills-sync` cambiaría | la deriva de `skills-check` | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| SC-003, FR-005, FR-006, FR-008 | con las marcas: 9600 sin las líneas, 0 defectos; 9601, solo el de FR-005; 11 001, el de 11 000 y el de FR-005 en ese orden. Sin ellas: 11 000, 0; 11 001, solo el suyo | `casosDePresupuestos`, subtest `presupuestos`, con `assert.Equal` de los defectos | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-001, FR-003 | 3 columnas en cada tabla; 0 líneas del sobre o de las banderas; 0 bytes de diferencia con la tabla esperada | `TestRenderizarTabla`, `TestRegenerarYComparar` y `TestCuerpoDeLaSkill` | `ci:internal/skills/comandos_test.go:TestRenderizarTabla`, `ci:internal/skills/sincronia_test.go:TestRegenerarYComparar`, `ci:internal/skills/sincronia_test.go:TestCuerpoDeLaSkill` |
| FR-002 | 0 cambios en los errores y los defectos que esperan sus tests | los casos de error de `TestRenderizarTabla` y `TestDescripcionDeVerbo`, sin tocar | `ci:internal/skills/comandos_test.go:TestRenderizarTabla`, `ci:internal/skills/comandos_test.go:TestDescripcionDeVerbo` |
| SC-004, FR-010, FR-013, FR-014 | 0 bytes de diferencia, fuera de lo generado, entre cada skill y su control con los cambios declarados; 0 en su frontmatter y su descripción | `TestSkillsDelProductoDesdeControl` | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl` |
| SC-005, FR-015 | las 2 frases nuevas en su cuerpo; los fallos de la regla, exactamente los admitidos | `TestRemisionesDelRepositorio` | `ci:internal/skills/remisiones_test.go:TestRemisionesDelRepositorio` |
| SC-006, FR-001, FR-010, FR-011, FR-012, FR-021 | el 100 % de las aserciones de los dos guiones `h46-` sobre lo instalado | `TestEntregaDelHito` | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| FR-016 | el 100 % de las aserciones de `h41-skills-instaladas` y `h41-1-skills-instaladas` | `TestEntregaDelHito` | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| SC-007, FR-019 | 0 cambios en `--describe`, el esquema de salida, el sobre, los códigos o la forma de `data` de un applet | los tests de cada applet, sin tocar, y la validación de las salidas contra `schemas/` | `ci:Makefile:test`, `ci:Makefile:schema-check` |
| SC-008 | `normativa`: 0 clases que deciden bajo su umbral y 0 series que deciden bajo el suyo | el job de evals del cierre, sin reutilizar nada (D13) | `evals:normativa:sin_activar:claude-sonnet-5-5:orden`, `evals:normativa:sin_activar:claude-sonnet-5-5:herramienta`, `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden`, `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`, `evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5`, `evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`, `evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados`, `evals:normativa:duracion_de_las_sesiones:orden`, `evals:normativa:duracion_de_las_sesiones:herramienta`, `evals:normativa:duracion_del_juez:orden`, `evals:normativa:duracion_del_juez:herramienta` |
| SC-008 | `plazos-y-territorio`: ídem | ídem | `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden`, `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta`, `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden`, `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta`, `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden`, `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`, `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden`, `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta`, `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden`, `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`, `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden`, `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta` |
| SC-008 | `jurisprudencia`: ídem | ídem | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`, `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`, `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`, `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`, `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`, `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`, `evals:jurisprudencia:duracion_del_juez:orden`, `evals:jurisprudencia:duracion_del_juez:herramienta` |
| SC-008 | `consulta-juridica`: ídem | ídem | `evals:consulta-juridica:sin_activar:claude-sonnet-5-5:orden`, `evals:consulta-juridica:sin_activar:claude-sonnet-5-5:herramienta`, `evals:consulta-juridica:cobertura_sin_declarar:claude-sonnet-5-5:orden`, `evals:consulta-juridica:cobertura_sin_declarar:claude-sonnet-5-5:herramienta` |

SC-008 usa los umbrales del job tal como están, que el hito no toca («Fuera de alcance»): las filas nombran los que los
informes del cierre de H45.1 publican con `decide: true` (research V12), y el veredicto del job sale en fallo con
cualquiera de ellos por debajo —y con cualquier otro que el cierre publique con `decide: true`—, y también con una serie
que decide por debajo de su umbral (V11), que es la otra condición de SC-008. Que no se reutilice nada de `main` lo da la regla de huellas de hoy (D13; V10).

No son umbrales de `make ci` ni del cierre, y su comprobación sin modelo es otra: SC-003 a SC-006, además de sus tests,
piden el registro de mutantes (`mutantes.md`, T006); FR-010 (`git diff f89093dd -- skills/` vacío) lo sostienen en
`make ci` `TestSkillsDelProductoDesdeControl` y la deriva de `skills-check`, y su comparación con Git es Q4; la igualdad
de la copia congelada de H41.1 con su guion (FR-016) la deja la tarea `[datos]` y la comprueba Q5 con `cmp`, sin un
control en `make ci` que el spec no pide; y que el diff contra `main` no toque otros tests ni guiones (SC-007) lo hace
cumplir el guardián de diff con las rutas de cada tarea (`scripts/workflow/guardian-diff.sh`, que el hito no toca), y
su comprobación es Q8.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|---|---|---|
| Cambiar dos guiones activados de hitos anteriores (`h41-skills-instaladas.txtar`, de H41, y `h41-1-skills-instaladas.txtar`, de H41.1) y la copia congelada de H41.1 (constitución, «Reglas del modo desatendido»: la suite congelada no se modifica) | Fijan el texto de hoy —la orden de leer `consulta_preparada.md` con la tabla detrás y la de `municipios.md`—, que FR-010 cambia; sin cambiarlos, `make ci` sale en rojo, y FR-016 los pide al día. El cambio es el mínimo: una línea en H41.1 y en su copia, una por dos en H41 (contracts/textos-y-controles.md §4), y siguen comprobando lo mismo | Dejarlos: `make ci` en rojo. Retirar sus aserciones: el siguiente cambio de texto quedaría sin control (US4). La congelación que vigila el run es la de la suite de este hito (V4, V5), que nadie toca. Precedentes: FR-013 de H41.1 y FR-013 de H45.1. Toda modificación de `testdata/` la ve la persona en el informe final (constitución, capa 3) |

## Decisiones

Todas en research.md, con la alternativa descartada: D1 (la tabla y lo que se retira), D2 (el defecto en
`defectosDelCuerpo`), D3 (`casosDePresupuestos`), D4 (`TestVerticalDeEnsayo`), D5 (los textos con `git checkout`), D6
(cambios, no movimientos), D7 (remisiones y guiones de H41), D8 (la suite, acotada), D9 (el orden de las tareas), D10
(el registro de mutantes), D11 (el fixture de «control»), D12 (la guía y el registro), D13 (las evals del cierre) y D14
(las skills de Go).
