# Implementation Plan: H45 · El molde admite un paquete con sus plazos y su norma, sin código (ADR 0062)

**Branch**: `042-h45-el-molde-admite` | **Date**: 2026-11-23 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/042-h45-el-molde-admite/spec.md`

## Summary

La sincronía de las skills reparte por paquete lo que hoy escribe en todas las referencias o por el precepto: cada tipo
de `data/plazos.yaml` va a la referencia del `vertical` que comparten las normas de sus preceptos propios, con dos
defectos nuevos (sin paquete, de más de uno), y cada fila de lo regulado aparte va al paquete de la norma que la rige si
la lleva con `vertical` (research D1, D2). `normativa` mueve su forma para PowerShell 5.1 a `con_herramientas.md`, y el
control de las órdenes y la declaración del laboratorio de H41 lo admiten (D6, D7). El laboratorio deja de pedir que se
declaren las referencias de un paquete (D7); `UnionDeGrabaciones` suma sola las carpetas de `evidencias/` de cada hito
(D5); `plazos-y-territorio` va a tres sesiones a la vez en el job de evals (D9) y cambia su descripción (D8). Los tests
que fijaban con una lista literal lo que hay en `data/` o `evals/` derivan lo esperado (D10), con la lista completada
por tres prototipos enteros, `-tags=integration` incluido (research, P1-P3); un guion congelado lo comprueba corriendo esa
lista sobre una copia del módulo con un paquete más (D3). La vertical de ensayo crece con un tipo y una fila con norma
(D11). Con los datos de hoy, las referencias generadas y `kitlegal plazos` no cambian. Ninguna dependencia, paquete,
orden, bandera ni código de salida nuevo.

## Technical Context

**Language/Version**: Go 1.27 (`go1.27.4`, research V1).

**Primary Dependencies**: las fijadas, ninguna nueva: `gopkg.in/yaml.v3` (lectores de `data/`, sin cambios),
`stretchr/testify` y `rogpeppe/go-internal` v1.16.0 (testscript; V2, V3) en los tests; biblioteca estándar (`os`,
`io/fs`, `strconv`, `os/exec` en el programa del guion y en el arnés).

**Storage**: ficheros versionados: `data/` no cambia; ficheros de prueba nuevos o crecidos en `internal/skills/testdata/`
(la vertical de ensayo, el `plazos.yaml` de `TestRegiones`); dos `SKILL.md`, una referencia a mano, la variante de
descripciones del laboratorio, `.github/workflows/evals.yml`, `docs/SKILLS.md` y `CHANGELOG.md`. Nada en la caché ni en
el grafo.

**Testing**: `go test -race` en `make ci`; tests de tabla y subtests sobre copias en `t.TempDir()`; dos guiones testscript
de aceptación congelados; uno de ellos corre `go test` en una copia del módulo (D3). Sin red ni modelo. Skills de Go en
research D14.

**Target Platform**: Linux y macOS, donde corre `make ci`.

**Project Type**: CLI multicall con skills empotradas; el cambio es de la sincronía de las skills (`internal/skills`),
del arnés de evals (`internal/evals`), de los tests de `internal/app` y de dos skills.

**Performance Goals**: ninguno del spec. El guion de SC-004 añade a `make ci` dos `go test` en una copia, medidos en
7,4 s y 3,5 s de reloj (P3).

**Constraints**: con `data/` de `main`, 0 diferencias en las referencias generadas (FR-008) y `kitlegal plazos` igual
(FR-026); cuerpos de skill ≤ 11 000 caracteres con sitio para dos líneas de paquete (FR-020); descripciones ≤ 450 y
suma ≤ 1 700 (SC-007); concurrencias del job ≤ 12 (FR-021); ninguna eval cambia (FR-025); ningún `.go` en un paquete
nuevo (US2).

**Scale/Scope**: 4 funciones sin exportar nuevas en producción (`paquetesDelTipo`, `paqueteDeLaFila`,
`unionDeGrabacionesDe`, `carpetaDeUnHito`) y 1 predicado (`esDeUnaRaizDeGrabaciones`); 1 firma exportada que cambia
(`UnionDeGrabaciones`, gana el error) y 1 constante exportada que desaparece (`GrabacionesDeH36`); 5 ficheros de
producción de Go (`internal/skills/regiones.go`, `internal/evals/grabaciones.go` y sus llamadores `dependencias.go`,
`medida.go` y `sesiones.go`); unos 20 ficheros de test; 4 ficheros de prueba; 4 guiones de `internal/app` con sus copias
congeladas; 2 guiones de aceptación.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio o regla | Cómo lo cumple este hito |
|---|---|
| I · Fuentes públicas y frontera humana (INNEGOCIABLE) | Ninguna petición HTTP: el hito no graba nada (research, «Datos externos»); la copia de SC-004 solo lee grabaciones del repositorio (V9). Ninguna identidad, POST ni CAPTCHA |
| II · Nada sin cita ni fuente | Ningún sobre cambia (FR-026). La línea `⚠ NO CUBIERTO:` de cada fila nombra la norma que la rige y el precepto, como hoy (FR-007); que cada fila vaya a la referencia del paquete de su norma (FR-005) hace que una pregunta de esa materia lea su línea en la referencia que lee, y que un tipo mal declarado sea un defecto (FR-003, FR-004) impide dar el plazo de una materia con la referencia de otra. La descripción deja de decir que solo cuenta lo administrativo y declara que de las demás materias no da plazos (FR-023) |
| III · Tests primero y offline | Primero los dos guiones de `aceptacion/`, rojos hoy por una aserción (contracts/aceptacion.md); los unitarios, offline sobre copias en `t.TempDir()` y ficheros de `testdata/`; la copia de SC-004, sin red (`GOPROXY=off`) y solo con grabaciones del repositorio. Las salidas de applet se siguen validando contra `schemas/` (ningún esquema cambia). Cobertura: el código nuevo de `internal/skills` e `internal/evals` lo cubren `TestVerticalDeEnsayo`, `TestRegiones/defectos`, `TestManifiestoDeGrabaciones` y `TestGrabacionesSinSolape` |
| IV · Arquitectura hexagonal con reglas ejecutables | Todo en `internal/skills` e `internal/evals` (sin `core` nuevo); los defectos de la sincronía siguen siendo `defectoDeLosDatos` de `data/plazos.yaml` y `make skills-sync` falla con ellos como hoy; el error de leer `evidencias/` sube envuelto (`%w`) hasta quien prepara la sesión, que ya devuelve error, sin callarlo. El binario no cambia: ningún código de salida ni `panic` nuevo |
| V · Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia ni paquete nuevo; cuatro funciones sin exportar y un predicado, cada una con su requisito («Trazabilidad»). Rechazados: un campo `paquete:` (fuera de alcance), una orden nueva del arnés, `tar`/`awk` en el guion, un `fs.FS` en `UnionDeGrabaciones`, un test de cada nombre de carpeta que `grabar-datos.sh` no escribe (research D3, D5) |
| VI · Un binario, convenciones de agente | Ninguna orden, bandera ni argumento nuevo; `--describe` no cambia; las skills siguen empotradas e instaladas por `kitlegal skills install` |
| VII · Grafo y privacidad | Nada toca el grafo ni la caché; ningún dato de nadie |
| VIII · Skills primero; el binario es la herramienta | Mejora el molde de las skills de capacidad para que una materia entre como paquete (lo que H42 necesita) y protege las cuatro skills: el cierre mide todas, sin reutilizar nada (FR-025). Lo único que cambia de una skill es el movimiento de `normativa` (sitio para un paquete real más) y la descripción de `plazos-y-territorio`; el cuerpo de `SKILL.md` sigue bajo su tope y el detalle, en `references/` |
| IX · Genericidad territorial, validación local | Ningún caso por municipio; la copia de SC-004 corre `h36-plazos-de-hoy`, con Leganés, como hoy. Ninguna herramienta territorial cambia |
| X · Genericidad por materia, validación administrativa | Es el objeto del hito: ningún código nombra una materia —el paquete sale del `vertical` de `data/normas.yaml`—, y la abstracción se valida con lo administrativo (igual, US1), la vertical de ensayo (US2) y el paquete de la copia de SC-004 sobre una norma real (D3). La descripción deja de nombrar una materia (FR-023). Ningún caso nuevo de lo administrativo |
| R1 · `internal/core/**` no importa el kernel, adaptadores ni entrada y salida | `internal/core` no cambia |
| R2 · solo `internal/httpx` importa `net/http` | Nada nuevo lo importa |
| R3 · solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | Nada nuevo los importa |
| R4 · solo `internal/cli` y `cmd/` llaman a `os.Exit` (`forbidigo`) | Ningún `os.Exit` nuevo en el módulo. El programa del guion (`paquete-de-mas.go`) vive en el archivo `.txtar`, fuera del módulo: `forbidigo` ni `TestArquitectura` lo ven, y sale con su código como cualquier programa |
| R5 · solo `internal/render` escribe en stdout; logs con `log/slog` a stderr (`forbidigo`) | Nada nuevo del módulo escribe en stdout; el programa del guion lo hace fuera del módulo |
| R6 · `internal/graph` no importa `internal/source` ni `internal/render` | `internal/graph` no cambia |
| R7 · solo el paquete del servidor importa el SDK de MCP | Nada nuevo lo importa |
| Garantía sin red de `TestArquitectura` (`paquetesSinRed`, `ficherosSinRed`) | Ningún paquete de las listas gana una importación que llegue a `net`; `internal/evals/grabaciones.go` solo gana `os` y `strconv` |

Re-evaluado tras la fase 1: sin violaciones; Complexity Tracking, vacío.

## Project Structure

### Documentation (this feature)

```text
specs/042-h45-el-molde-admite/
├── spec.md
├── plan.md                              # este fichero
├── research.md                          # P1-P3, D1-D14, V1-V12, datos externos
├── data-model.md                        # paquete, tipo, fila, raíz, declaración, la copia de SC-004, la vertical
├── quickstart.md                        # escenarios de comprobación
├── contracts/
│   ├── sincronia-de-los-paquetes.md     # FR-001 a FR-008, §C.1, TestVerticalDeEnsayo, TestRegiones/defectos
│   ├── skills-y-laboratorio.md          # FR-009 a FR-014, FR-023, FR-024
│   ├── evals.md                         # FR-015, FR-021, FR-022
│   ├── tests-sin-listas.md              # FR-016, FR-017, FR-018: cada test, su derivación y su mutante
│   └── aceptacion.md                    # los dos guiones y el entorno de Go del arnés
├── aceptacion/                          # la escribe la tarea [aceptacion]; congelada
│   ├── paquete-de-mas.txtar
│   └── skills-instaladas.txtar
├── mutantes.md                          # registro de FR-027, lo escribe la última tarea
└── tasks.md                             # /speckit-tasks
```

### Source Code (repository root)

```text
internal/skills/
├── regiones.go                          # paquetesDelTipo, paqueteDeLaFila; filasDeLosRecursosYPlazos, filasDeLaCobertura, filasSinRegion
├── regiones_test.go                     # TestRegiones sobre testdata/regiones/plazos.yaml; 3 subtests de defectos
├── vertical_de_ensayo_test.go           # FR-020; el tipo y la fila de ensayo; sin-la-region; orden de lo cubierto
├── cobertura_test.go, cubierto_test.go, referencias_test.go, normas_test.go   # derivados (contracts/tests-sin-listas.md)
└── testdata/
    ├── regiones/plazos.yaml             # nuevo [datos]
    └── vertical-de-ensayo/              # data/plazos.yaml nuevo, data/cobertura.yaml y ensayo.md de plazos-y-territorio [datos]
internal/evals/
├── grabaciones.go                       # UnionDeGrabaciones ([]string, error), unionDeGrabacionesDe, carpetaDeUnHito
├── dependencias.go, medida.go, sesiones.go   # propagan el error; esDeUnaRaizDeGrabaciones en rutaQueMide
├── grabaciones_test.go                  # TestGrabacionesSinSolape, TestManifiestoDeGrabaciones/raices-de-los-hitos
├── definicion_test.go, plan_test.go     # FR-021, FR-022
├── conjunto_test.go                     # FR-011, FR-012; conjunto-plazos-y-territorio y precepto-de-la-cobertura derivados
├── informe_test.go, eleccion_test.go    # derivados
├── laboratorio_familias_test.go         # FR-010, FR-014; referencias-de-un-paquete
├── laboratorio_descripciones_test.go, descripciones_del_producto_test.go   # FR-023
└── *_test.go que llaman a UnionDeGrabaciones   # el error
internal/app/
├── e2e_test.go                          # GOCACHE y GOMODCACHE a los guiones (tarea [aceptacion])
├── plazos_casos_test.go                 # 19-eval-c-tributos derivado
└── testdata/script/
    ├── h32-paquetes.txtar, h33-cobertura-instalada.txtar, h36-plazos-de-hoy.txtar, h41-4-skills-instaladas.txtar   # [datos]
    └── h45-paquete-de-mas.txtar, h45-skills-instaladas.txtar   # activados por el workflow
skills/normativa/SKILL.md, skills/normativa/references/con_herramientas.md, skills/plazos-y-territorio/SKILL.md
laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt
.github/workflows/evals.yml, docs/SKILLS.md, CHANGELOG.md
specs/030-h32-arquitectura-de-skills/aceptacion/paquetes.txtar, specs/039-h33-cobertura-en-positivo/aceptacion/cobertura-instalada.txtar,
specs/041-h36-el-motor-de/aceptacion/plazos-de-hoy.txtar, specs/038-h41-4-las-skills-del/aceptacion/skills-instaladas.txtar   # [datos]
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquetes nuevos: la sincronía sigue en
`internal/skills`, el arnés de evals en `internal/evals` y el e2e en `internal/app`.

## Diseño

1. **El paquete de un tipo** (FR-001 a FR-004; contracts/sincronia-de-los-paquetes.md §1-§2): `paquetesDelTipo` y la
   región que filtra por su contexto; los defectos de todos los tipos antes de pintar, por la vía de
   `defectoDeLosDatos` de `data/plazos.yaml`.
2. **El paquete de una fila** (FR-005 a FR-007; §3): `paqueteDeLaFila`, la misma en la región y en `filasSinRegion`; la
   línea, la de hoy.
3. **`normativa` hace sitio** (FR-009, FR-010; contracts/skills-y-laboratorio.md §1, §3): las cuatro líneas al final de
   `con_herramientas.md`; la declaración del laboratorio con su `cambio` y su `movimiento`.
4. **El control de PowerShell** (FR-011 a FR-013; §2): la regla del `SKILL.md`, los cuatro casos nuevos de §C.3 y la fila
   de `docs/SKILLS.md`.
5. **El laboratorio sin las referencias de un paquete** (FR-014; §3): `diferenciasDeLosFicheros` con la excepción de los
   `vertical` y el subtest `referencias-de-un-paquete`.
6. **Las grabaciones** (FR-015; contracts/evals.md §1).
7. **Los tests sin listas** (FR-016 a FR-018; contracts/tests-sin-listas.md), con los guiones en tareas `[datos]`.
8. **La vertical de ensayo** (FR-019, FR-020; contracts/sincronia-de-los-paquetes.md §5).
9. **El trabajo de evals** (FR-021, FR-022; contracts/evals.md §2).
10. **La descripción y el CHANGELOG** (FR-023, FR-024; contracts/skills-y-laboratorio.md §4).
11. **La suite de aceptación** (SC-004, US5.1, US9.1; contracts/aceptacion.md), con el entorno de Go del arnés.

Orden de dentro afuera: la tarea `[aceptacion]` (guiones y entorno del arnés); el molde (1, 2) con la vertical (8) y los
subtests de defectos; las grabaciones (6); el job (9); el control y el laboratorio (4, 5); el movimiento y la descripción
(3, 10); los tests sin listas (7) y sus guiones `[datos]`; `mutantes.md` al final (FR-027).

## Uso, de fuera adentro

| Salida | Quién la consume y cuántas veces | Tamaño, hoy y con meses de uso | Cuándo deja de darse |
|---|---|---|---|
| `references/<paquete>.md` de `plazos-y-territorio` (regiones `recursos_y_plazos` y `cobertura_por_materia`) | `plazos-y-territorio`, una vez por pregunta de ese paquete, por su línea de la marca de paquetes; usa una fila de tipo para el código de `plazos calcular` y una de lo regulado aparte para su `⚠ NO CUBIERTO:` | `administrativo.md`, 100 líneas y 10 635 bytes hoy y tras H42 (sus 9 filas de tipo, de unos 100 bytes, y sus filas de lo regulado aparte, de unos 340); cada paquete, solo sus filas. Sale de `data/`, no del uso: con cientos de normas y miles de bloques consultados durante meses, es la misma | una fila, al salir de `data/` o cambiar de paquete, en el siguiente `make skills-sync` |
| `references/con_herramientas.md` de `normativa` | `normativa`, una vez por conversación, antes de la primera orden (cuerpo, línea 16); la forma movida, una vez por bloque leído sin fecha en PowerShell 5.1 sin MCP | de 2 442 a unos 2 616 bytes; no depende del uso | — (protocolo) |
| Cuerpo de `normativa` | el agente, al activar la skill, una vez por conversación | de 10 926 a unos 10 751 caracteres | — |
| `description` de `plazos-y-territorio` | el agente, en cada sesión con las skills instaladas, para elegir skill | 429 caracteres (hoy 427); suma de la variante, 1 692 | — |
| Defectos de FR-003 y FR-004 | quien escribe un paquete, persona o run, una vez por `make skills-sync` | una línea por tipo mal declarado, 153 bytes la de «sin paquete» | cuando las normas de los preceptos propios del tipo comparten un `vertical` |
| `UnionDeGrabaciones` | el job de evals y los tests, una vez por preparación de sesión | 4 raíces hoy; una más por hito que graba el BOE, no por uso | — |
| Informe del job de evals | la persona, al cierre | su forma no cambia; doce sesiones a la vez pueden topar con el límite de la suscripción, que el cierre enseña como fallo de la fuente | — |

Invocaciones por pregunta que hace la `SKILL.md`: ninguna nueva. `plazos-y-territorio` lee la referencia de un paquete
y llama a `plazos calcular` como hoy; `normativa` lee `con_herramientas.md` como hoy.

## Trazabilidad

| Mecanismo | Requisito |
|---|---|
| `paquetesDelTipo`, filtro de `filasDeLosRecursosYPlazos` y sus dos defectos | FR-001 a FR-004, SC-003 |
| `paqueteDeLaFila` en `filasDeLaCobertura` y `filasSinRegion` | FR-005 a FR-007 |
| `testdata/regiones/plazos.yaml` y los tres subtests de `TestRegiones/defectos` | FR-016 (P2), SC-003 |
| la vertical de ensayo crecida, `instalarLaVertical` con `tipos`, las comprobaciones y `sin-la-region` de `TestVerticalDeEnsayo` | FR-019, FR-020, SC-002 |
| el movimiento, el `cambio` y el `movimiento` de la declaración | FR-009, FR-010 |
| la regla del `SKILL.md` en `controlDeLasOrdenes` y los casos 1-4 de §C.3; la fila de `docs/SKILLS.md` | FR-011 a FR-013, SC-008 |
| `diferenciasDeLosFicheros` con la excepción de los `vertical` y `referencias-de-un-paquete` | FR-014, SC-009 |
| `UnionDeGrabaciones` con error, `unionDeGrabacionesDe`, `carpetaDeUnHito`, `esDeUnaRaizDeGrabaciones`, los dos subtests de las raíces | FR-015, SC-005 |
| las derivaciones de contracts/tests-sin-listas.md y los cuatro guiones `[datos]` | FR-016, FR-017, SC-004 |
| `concurrencia: 3`, `ajustesDelTrabajo`, `sesionesALaVezComoMucho = 12`, el plan derivado y `cuatro-evals-mas` | FR-021, FR-022, SC-006 |
| la descripción, la variante, `descripcionCortaDePlazosYTerritorio` y las sumas derivadas | FR-023, SC-007 |
| la entrada de `CHANGELOG.md` | FR-024 |
| `paquete-de-mas.txtar` con su programa y su paquete; `GOCACHE` y `GOMODCACHE` en el arnés | SC-004 (y US2.1, US2.2 sobre datos reales) |
| `skills-instaladas.txtar` | US5.1, US9.1, SC-007 |
| `mutantes.md` | FR-027 |

## Controles mecánicos que el hito añade o toca

Según «Gates» de la constitución, capa 1:

- **`make ci`**, sin objetivos nuevos: `make test` y `make test-integration` ejecutan los tests nuevos y cambiados y
  `TestEntregaDelHito` con los dos guiones de H45 activados; `make skills-check` (`TestSkillsDelRepositorio`,
  `TestEvalsDelRepositorio`…) con las regiones repartidas y el control de PowerShell nuevo; `make lint` con la
  configuración de hoy, sin `//nolint` ni `t.Skip`.
- **`make skills-sync`**: da los defectos de §C.1 (FR-003, FR-004).
- **Tests de contrato**: ningún esquema ni sobre cambia; `TestEsquemasPublicados` y `TestAppletPlazos`, iguales.
- **Fixtures** (todos en tareas `[datos]`): nuevos `internal/skills/testdata/regiones/plazos.yaml` y
  `internal/skills/testdata/vertical-de-ensayo/data/plazos.yaml`; cambiados
  `internal/skills/testdata/vertical-de-ensayo/data/cobertura.yaml` y
  `…/skills/plazos-y-territorio/references/ensayo.md`; los guiones `h32-paquetes`, `h33-cobertura-instalada`,
  `h36-plazos-de-hoy` y `h41-4-skills-instaladas` y sus copias congeladas de `specs/`. Ninguna grabación ni
  `evidencias/`.
- **Test de arquitectura**: `TestArquitectura` sin cambios.
- **Exit codes**: ninguno nuevo.

**Mutantes** (`mutantes.md`, FR-027): cada uno en una copia del módulo en el temporal, nunca en el árbol de trabajo
(research D13); el registro dice la orden, lo esperado y lo visto.

| Control | Mutante | En rojo |
|---|---|---|
| Lo de hoy, igual (SC-001) | la regla de hoy (todos los tipos en cada referencia) con un tipo de ensayo | `TestVerticalDeEnsayo` (el tipo en `administrativo.md`); y contra `main`, `git diff main -- skills/` solo en lo declarado |
| El molde (SC-002) | un tipo asignado a todas las referencias; una fila asignada por su precepto; `filasSinRegion` con la regla vieja; el cuerpo de `normativa` sin el movimiento | `TestVerticalDeEnsayo`: su comprobación del tipo, la de la fila, `sin-la-region`, la del sitio |
| Los defectos (SC-003) | quitar la regla; contar los preceptos del cómputo | `TestRegiones/defectos/tipo-sin-paquete`, `/tipo-de-dos-paquetes`, `/computo-de-otro-paquete` |
| Los tests sin listas (SC-004) | la lista literal de hoy devuelta a `TestCoberturaDelRepositorio`; y a cada test de la lista, su mutante de contracts/tests-sin-listas.md | `h45-paquete-de-mas` (`tests enteros: 1`); el test de cada fila |
| Las grabaciones (SC-005) | el orden alfabético; un fichero de H5 copiado a `evidencias/h36/boe.legislacion-consolidada/` | `TestManifiestoDeGrabaciones/raices-de-los-hitos/orden-y-carpetas-que-no-entran`; `TestGrabacionesSinSolape` |
| El tope del job (SC-006) | `plazos-y-territorio` a dos; una sesión más a la vez | `TestDefinicionDelJob/cuatro-evals-mas`; los casos de `fallosDeLasSesionesALaVez` |
| La descripción (SC-007) | la de hoy en el `SKILL.md` o en la variante | `TestSkillsDelProductoDesdeControl` (frontmatter), `TestDescripcionesCortas`, `h41-4-skills-instaladas`, `h45-skills-instaladas` |
| PowerShell (SC-008) | los cinco de §C.3 | `TestControlesSobreLasReferenciasAMano` |
| El laboratorio (SC-009) | un nombre que no es un vertical; una referencia de la base cambiada | `TestSkillsDelProductoDesdeControl/referencias-de-un-paquete` |
| `kitlegal plazos` igual (FR-026) | otra fecha en un caso; el mensaje de un tipo desconocido sin « ni » | `TestCasosDePlazos`, `h36-plazos-de-hoy` |

**Aceptación e2e:** dos guiones testscript que escribe la primera tarea en `specs/042-h45-el-molde-admite/aceptacion/` y
quedan congelados (contracts/aceptacion.md): `paquete-de-mas.txtar` (US4 y US2.1-US2.2 sobre una norma real · FR-001,
FR-002, FR-005, FR-006, FR-016, FR-017 · SC-004), que copia el módulo bajo prueba, le añade el paquete `prueba`, lo
sincroniza y corre en la copia la lista completa de FR-016; y `skills-instaladas.txtar` (US5.1, US9.1 · FR-009, FR-023 ·
SC-007), sobre `kitlegal skills install`. La misma tarea da a los guiones `GOCACHE` y `GOMODCACHE` (research D4). US1.3,
US3, US5.2-US5.4, US6, US7 y US8 no tienen comportamiento visible del binario: los cubren los tests de la tabla de
mutantes; US9.3, las evals del cierre.

**Datos externos:** ninguno. El hito no graba ni deriva nada a `data/`: no hay manifiesto `grabaciones.json` nuevo ni
test `TestGrabar*`, y el paso `grabar_datos` no tiene qué hacer. La copia de SC-004 solo lee lo ya grabado de
`BOE-A-2015-11722` en `testdata/evals/` (research V9).

**Autonomía:** ninguna tarea pide nada a una persona ni publica la rama, abre la propuesta de cambio o mide CI y evals
remotas; el cierre (push, propuesta, CI y evals sobre la cabeza) lo hace el workflow tras la revisión final. «Antes de
lanzar» del spec es de la persona y no es de este plan.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| SC-001, FR-008 | 0 diferencias entre las referencias de `skills/` y lo generado desde `data/`; 0 bytes cambiados en las skills fuera de lo declarado | `TestSkillsDelRepositorio` (`make skills-check`) y `TestRegiones/sincronia`, en rojo con una región distinta de la generada; `TestSkillsDelProductoDesdeControl`, en rojo con un byte no declarado | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, `ci:internal/skills/regiones_test.go:TestRegiones`, `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl` |
| SC-002, FR-002, FR-005, FR-006, FR-020 | el tipo de ensayo en 1 referencia y en 0 más; la fila con norma en `ensayo.md` y en 0 líneas de `administrativo.md`; 0 `.go` y 0 `SKILL.md` fuera de las marcas; cada cuerpo con la línea de ensayo y 70 caracteres más, ≤ 11 000 | `TestVerticalDeEnsayo` | `ci:internal/skills/vertical_de_ensayo_test.go:TestVerticalDeEnsayo` |
| SC-003, FR-003, FR-004 | 2 de 2 defectos con su tipo; 0 en el caso del cómputo | los tres subtests de `TestRegiones/defectos` | `ci:internal/skills/regiones_test.go:TestRegiones` |
| SC-004, FR-016, FR-017 | 0 tests de la lista en rojo con la copia de un paquete más | `h45-paquete-de-mas`: `stdout` de los dos códigos `0` | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| SC-005, FR-015 | la unión del repositorio: 3 fijas + las carpetas de hito con manifiesto (hoy 4), 0 de `evidencias/h9/`; con h99 en una copia, 5, h99 la última; 0 ficheros compartidos entre dos raíces salvo `robots.txt` | `TestGrabacionesSinSolape`; `TestManifiestoDeGrabaciones/raices-de-los-hitos` | `ci:internal/evals/grabaciones_test.go:TestGrabacionesSinSolape`, `ci:internal/evals/grabaciones_test.go:TestManifiestoDeGrabaciones` |
| SC-006, FR-021, FR-022 | suma de concurrencias ≤ 12; el peor caso de cada trabajo ≤ su tope (352 min el de `plazos-y-territorio`), también con cuatro evals más, y > 352 min con dos | `TestDefinicionDelJob` (`fallosDeLasSesionesALaVez`, `peor-caso`, `cuatro-evals-mas`) y `TestPlanEnDosModos` | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, `ci:internal/evals/plan_test.go:TestPlanEnDosModos` |
| SC-007, FR-023 | 429 caracteres la de `plazos-y-territorio` (≤ 450); suma de la variante ≤ 1 700; cada `description` ≤ 900; el texto de FR-023 byte a byte | `TestDescripcionesCortas`, `TestDescripcionesDelProducto`, `TestSkillsDelProductoDesdeControl`, `h41-4-skills-instaladas` y `h45-skills-instaladas` | `ci:internal/evals/laboratorio_descripciones_test.go:TestDescripcionesCortas`, `ci:internal/evals/descripciones_del_producto_test.go:TestDescripcionesDelProducto`, `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| SC-007, SC-010 | 0 descripciones marcadas por el juez de las descripciones | el job de evals del cierre, umbral 0 con `decide: true` | evals:plazos-y-territorio:habla_en_primera_o_segunda_persona, evals:normativa:habla_en_primera_o_segunda_persona, evals:jurisprudencia:habla_en_primera_o_segunda_persona, evals:consulta-juridica:habla_en_primera_o_segunda_persona |
| SC-008, FR-011, FR-012 | 0 defectos en el caso 1 de §C.3; 1 defecto en cada uno de los 5 | `TestEvalsDelRepositorio/ordenes-para-powershell` y `TestControlesSobreLasReferenciasAMano` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, `ci:internal/evals/conjunto_test.go:TestControlesSobreLasReferenciasAMano` |
| SC-009, FR-014 | 0 diferencias con la referencia de un `vertical`; 1 en cada uno de los 2 mutantes | `TestSkillsDelProductoDesdeControl/referencias-de-un-paquete` | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl` |
| FR-018 | `normativa` con ≤ 23 evals y 10 positivas de normas distintas | `TestEvalsDelRepositorio` (el conjunto de `normativa`, sin cambiar) | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-026 | 100 % de los casos de `TestCasosDePlazos` y de los guiones de H9 y H36 con la salida de `main` | `TestCasosDePlazos`, `TestAppletPlazos` y los guiones `h9-plazos-*` y `h36-plazos-de-hoy` | `ci:internal/app/plazos_casos_test.go:TestCasosDePlazos`, `ci:internal/app/plazos_test.go:TestAppletPlazos`, `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| SC-010, FR-025 | ninguna clase que decide de `plazos-y-territorio` por debajo de su umbral, en los dos modos | el job de evals del cierre, sus umbrales con `decide: true` | evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta |
| SC-010, FR-025 | ninguna clase que decide de `normativa` por debajo de su umbral | el job de evals del cierre | evals:normativa:sin_activar:claude-sonnet-5-5:orden, evals:normativa:sin_activar:claude-sonnet-5-5:herramienta, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5, evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar, evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados, evals:normativa:duracion_de_las_sesiones:orden, evals:normativa:duracion_de_las_sesiones:herramienta, evals:normativa:duracion_del_juez:orden, evals:normativa:duracion_del_juez:herramienta |
| SC-010, FR-025 | ninguna clase que decide de `jurisprudencia` por debajo de su umbral | el job de evals del cierre | evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden, evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5, evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar, evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados, evals:jurisprudencia:duracion_del_juez:orden, evals:jurisprudencia:duracion_del_juez:herramienta |
| SC-010, FR-025 | ninguna clase que decide de `consulta-juridica` por debajo de su umbral | el job de evals del cierre | evals:consulta-juridica:sin_activar:claude-sonnet-5-5:orden, evals:consulta-juridica:sin_activar:claude-sonnet-5-5:herramienta, evals:consulta-juridica:cobertura_sin_declarar:claude-sonnet-5-5:orden, evals:consulta-juridica:cobertura_sin_declarar:claude-sonnet-5-5:herramienta |
| SC-011 | `make ci` con 0 fallos | `make ci` | `ci:Makefile:ci` |

Los nombres de los umbrales de las evals salen de los prefijos de `internal/evals/umbrales.go:128-141`, de
`claseDeLasDescripciones` (`descripciones.go:29`) y del modelo `MODELO_DE_EVALS` de `.github/workflows/evals.yml:298`
(`claude-sonnet-5-5`); el hito no cambia ningún umbral, clase ni eval. La activación ante las preguntas de hoy (SC-010,
la regla por serie del ADR 0016) va en los umbrales `sin_activar` de cada skill. Cada control de las filas de `ci:` tiene
su mutante en `mutantes.md`.

## Complexity Tracking

Vacío: ninguna violación que justificar.

## Decisiones

Todas en research.md, con la alternativa descartada: D1 (el paquete de un tipo y sus defectos), D2 (el de una fila), D3
(la comprobación de SC-004 en un guion con un programa de Go y un paquete que solo añade), D4 (el entorno de Go de los
guiones), D5 (`UnionDeGrabaciones` con error y un predicado puro para `rutaQueMide`), D6 (el control de PowerShell), D7
(el laboratorio), D8 (la descripción y las sumas derivadas), D9 (el job), D10 (los tests sin listas), D11 (la vertical de
ensayo y el sitio medido con la comprobación de producción), D12 (los subtests de los defectos), D13 (`mutantes.md` en
copias), D14 (skills de Go). La lista de FR-016 la completan los prototipos P1-P3.
