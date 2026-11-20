# Implementation Plan: H33 · Cobertura en positivo por capacidad y materia

**Branch**: `039-h33-cobertura-en-positivo` | **Date**: 2026-11-20 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/039-h33-cobertura-en-positivo/spec.md`, que aplica el fichero del hito
(`docs/roadmap/done/H33-cobertura-en-positivo-por-capacidad.md`) y el ADR 0058.

## Summary

H33 declara en positivo lo que kitlegal cubre y hace que todo lo demás se declare con una sola línea, la misma en todas
las materias. Tres piezas, todas sin applet ni orden nueva:

- **Datos.** `data/cobertura.yaml` gana `motivo_de_lo_no_declarado` y `cubierto` (cuatro filas por capacidad y materia),
  con su esquema y su lector.
- **Skills y sincronía.** `make skills-sync` escribe en `consulta-juridica` la región de lo cubierto. La tabla
  `cobertura_por_materia` de `administrativo.md` sale de la plantilla única del `SKILL.md` y de los dos motivos de la
  referencia. `consulta-juridica` y `plazos-y-territorio` llevan la plantilla, y la primera, la marca `⚠ NO VERIFICADO:`.
- **Evals.** Miden la frontera sin modelo: `no_cubierto: {fuera_de_lo_declarado: true}`, `tambien_la_resuelven`, la
  regla «frontera en dos materias» y dos evals nuevas en `consulta-juridica`, cuyos `sin_activar` y
  `cobertura_sin_declarar` nacen con tope 0 sin caso especial.

Lo grabado es un bloque, el art. 59 del ET, por el paso `grabar_datos`. Enfoque y alternativas: research.md.

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.4` en `go.mod`)

**Primary Dependencies**: las del módulo, sin ninguna nueva: el lector YAML (`go.yaml.in/yaml/v3`, research V15) y
`santhosh-tekuri/jsonschema/v6` (lectores de `data/` y de las evals), `rogpeppe/go-internal` (testscript),
`stretchr/testify`

**Storage**: ficheros versionados: `data/cobertura.yaml`, `schemas/*.json`, `skills/`, `evals/`, `testdata/`; el
binario empotra `skills/`

**Testing**: `go test` (`make test`, `make test-integration`, dentro de `make ci`), offline contra lo grabado;
testscript para el guion de aceptación

**Target Platform**: las del binario (Linux, macOS, Windows); los guiones e2e, en el runner `ubuntu-latest` de `make ci`

**Project Type**: CLI multicall en Go más skills agénticas sin código (el producto)

**Performance Goals**: el peor caso del trabajo de evals de `consulta-juridica`, 15 013 s, bajo su tope de 21 120 s
(FR-019)

**Constraints**: cuerpos de `SKILL.md` de como mucho 11 000 caracteres y el de `plazos-y-territorio` entregado de como
mucho 10 860 (FR-010); ningún texto de la línea en Go (FR-006, FR-013); ninguna sesión con modelo ni sondeo (FR-026)

**Scale/Scope**: 4 skills; 4 filas de `cubierto`, 9 de `regulados_aparte`; 4 evals en `consulta-juridica` y una que
cambia en `normativa`; una grabación

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio o regla | Cómo lo cumple este hito |
|---|---|
| I · Fuentes públicas y frontera humana (INNEGOCIABLE) | La única petición es la del paso `grabar_datos`: un bloque del BOE (`boe.legislacion-consolidada`, «Revisado» 2026-09-13 en `docs/SOURCES.md` de `main`). La pide `TestGrabarEvals` por `internal/httpx`, solo con GET. La plantilla termina en «kitlegal no da sus plazos, órganos, vías ni escritos»: no hay presentación ni escrito; ningún POST |
| II · Nada sin cita ni fuente | Es el hito del principio: lo cubierto se declara en datos, por capacidad y materia (`cubierto`), y lo demás, con una forma fija igual en todas las materias, la plantilla. Lo leído para completar va citado con `⚠ NO VERIFICADO:`. Lo regulado aparte sigue declarado con su norma y la cita de su precepto. Ningún sobre cambia; nada de memoria (§4 de `consulta-juridica`, regla 9) |
| III · Tests primero y offline | Primero el guion `aceptacion/cobertura-instalada.txtar`, que hoy falla por su primera aserción (research V7b). Los unitarios, offline contra lo grabado y copias en `t.TempDir()`. La grabación del art. 59 la hace `grabar_datos`, no el ejecutor. Ninguna salida de applet nueva que validar contra `schemas/`; `internal/core` no cambia, y su cobertura tampoco; cada función nueva de `internal/skills` e `internal/evals` llega en la misma tarea que el test que la ejercita (contracts/controles.md) |
| IV · Arquitectura hexagonal con reglas ejecutables | Solo cambian `internal/skills` (la sincronía) e `internal/evals`, con tipos y funciones no exportados salvo `FilaCubierta` y los campos nuevos de `Cobertura`, `Eval` y `NoCubiertoEsperado`. Ningún import ni error nuevo de producción, ningún código de salida nuevo y ningún `panic`. Un dato ilegible es un defecto de la sincronía, y una entrada manipulada, el defecto `inesperado` (ADR 0023). `TestArquitectura`, `depguard` y `forbidigo` no cambian |
| V · Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia, paquete, applet, orden, bandera ni objetivo de `make`. La región reutiliza `regionDeMarcas` y los lectores perezosos de las regiones (research D1). Los umbrales nacen sin código nuevo (V12). Cada mecanismo se traza a un FR («Trazabilidad») |
| VI · Un binario, convenciones de agente | Ninguna orden, bandera ni variable. `kitlegal skills install` instala los `SKILL.md` nuevos desde lo empotrado, como hoy |
| VII · Grafo y privacidad | No toca el grafo; ningún dato de una persona. Las evals nuevas no llevan datos personales (una fecha de despido ficticia) |
| VIII · Skills primero; el binario es la herramienta | Mejora `consulta-juridica` y `plazos-y-territorio`, medibles con sus evals en el cierre (FR-024). Lo que va a Go es lo determinista: la región generada, el juicio de la línea y los umbrales. El razonamiento queda en la skill (§4). Los cuerpos siguen en su presupuesto (4481 y 10 686 caracteres) |
| IX · Genericidad territorial, validación local | Nada territorial cambia; la fila del territorio de `cubierto` no lleva materia ni municipio. Ninguna herramienta con dimensión territorial nueva |
| X · Genericidad por materia, validación administrativa | Ningún caso por materia en código, datos ni protocolo. La plantilla es una para todas; `materia` es un dato; la regla «frontera en dos materias» cuenta normas sin `vertical` sin nombrarlas (research D10). Un paquete trae sus filas y sus motivos sin tocar código, y lo prueba la vertical de ensayo (FR-011) |
| R1 · `internal/core/**` no importa el kernel, adaptadores ni entrada y salida | `internal/core` no cambia |
| R2 · solo `internal/httpx` importa `net/http` | Nada nuevo lo importa: la grabación usa `internal/httpx` desde el test de grabación de hoy |
| R3 · solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | Nada nuevo los importa |
| R4 · solo `internal/cli` y `cmd/` llaman a `os.Exit` (`forbidigo`) | Ningún `os.Exit` nuevo |
| R5 · solo `internal/render` escribe en stdout; logs con `log/slog` a stderr (`forbidigo`) | Nada nuevo escribe en stdout: la sincronía devuelve sus defectos y los tests informan con `testing.T` |
| R6 · `internal/graph` no importa `internal/source` ni `internal/render` | `internal/graph` no cambia |
| R7 · solo el paquete del servidor importa el SDK de MCP | Nada nuevo lo importa |
| Garantía sin red (H19) del paquete raíz y del applet `skills` | `skills.go` y el applet no cambian; solo los bytes empotrados de las skills |

Re-evaluado tras la fase 1: sin violaciones; Complexity Tracking, vacío.

## Project Structure

### Documentation (this feature)

```text
specs/039-h33-cobertura-en-positivo/
├── plan.md              # este fichero
├── research.md          # decisiones (D), comprobaciones (V) y supuestos (S)
├── data-model.md        # entidades y reglas
├── quickstart.md        # escenarios de validación
├── contracts/
│   ├── datos-y-regiones.md   # data/cobertura.yaml, su esquema, la región de lo cubierto, el generador
│   ├── skills.md             # los textos de los SKILL.md y de administrativo.md; la documentación
│   ├── evals.md              # formato, juicio, umbrales, regla, evals, grabación, job
│   ├── controles.md          # los tests que añade o cambia y las suites congeladas
│   └── aceptacion.md         # el guion e2e
├── aceptacion/          # lo escribe la primera tarea (contracts/aceptacion.md)
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root)

```text
data/cobertura.yaml                                  # motivo_de_lo_no_declarado y cubierto
schemas/cobertura.yaml.json                          # [datos]
schemas/eval.yaml.json                               # [datos] no_cubierto.fuera_de_lo_declarado, tambien_la_resuelven
internal/skills/
├── cobertura.go, cobertura_test.go                  # Cobertura, FilaCubierta; TestLeerCobertura
├── cubierto.go, cubierto_test.go                    # nuevos: la región de lo cubierto; TestRegionDeLoCubierto
├── regiones.go, regiones_test.go                    # la plantilla y los motivos; TestRegiones
├── sincronia.go                                     # llama a la región de lo cubierto y al motivo
├── referencias_test.go, vertical_de_ensayo_test.go  # TestRegionDeLaCobertura, TestVerticalDeEnsayo
└── testdata/vertical-de-ensayo/                     # [datos] su fila de cubierto y sus motivos
internal/app/
├── skills_test.go                                   # partirElCuerpo
└── testdata/script/h32-paquetes.txtar, h32-capacidades.txtar   # [datos] FR-020
internal/evals/
├── formato.go, plazos.go, juzgar.go, informe.go, umbrales.go, conjunto.go
├── formato_test.go, plazos_test.go, tambien_la_resuelven_test.go (nuevo), conjunto_test.go, informe_test.go,
│   juzgar_test.go, preparar_test.go, definicion_test.go, eleccion_test.go,
│   laboratorio_familias_test.go, descripciones_del_producto_test.go
└── testdata/grabaciones.json                        # [datos] el art. 59 del ET; lo grabado, en testdata/boe.legislacion-consolidada/
skills/consulta-juridica/SKILL.md
skills/plazos-y-territorio/SKILL.md, references/administrativo.md
evals/consulta-juridica/01-…, 03-seguro-prorroga.yaml, 04-despido-plazo.yaml; evals/normativa/17-…
.github/workflows/evals.yml                          # solo un comentario
docs/SKILLS.md, CONTRIBUTING.md, docs/HISTORIAL.md, docs/USO.md, CHANGELOG.md, web/src/pages/como-se-evalua/index.astro
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquetes nuevos. La sincronía de las skills vive en
`internal/skills` y los juicios y umbrales de las evals en `internal/evals`, como hoy; un fichero nuevo por mecanismo
nuevo (`cubierto.go`, `tambien_la_resuelven_test.go`), según `golang-project-layout` (un paquete por responsabilidad,
sin `util`). Los tests, de tabla, con `t.Parallel()` y `testify` como los de cada paquete, y los mutantes sobre copias
en `t.TempDir()` (`golang-testing`); los nombres, en español como el resto del código (constitución, «Restricciones
técnicas»), y exportado solo lo que ya lo está o lo pide un campo exportado (`golang-naming`); `gofumpt`, `goimports`
y `golangci-lint` sin excepciones nuevas (`golang-lint`).

## Diseño

- **Los datos** (contracts/datos-y-regiones.md §1-§2; research D5, D12): `Cobertura` gana `MotivoDeLoNoDeclarado` y
  `Cubierto []FilaCubierta`; el esquema, las dos claves; el fichero, los textos de FR-001.
- **La región de lo cubierto** (§3; research D1, D4, D16): `cubierto.go`, con los lectores perezosos de las regiones; la
  llama `regenerarSkill` detrás de `paquetes.regenerar`; comprueba cada `materia` contra los `vertical`.
- **La plantilla y los motivos** (§5; research D2): `lineaDeNoCubierto` en lugar de `formasDeNoCubierto`, con sus dos
  defectos nuevos y el del marcador en un motivo; `conFormas` pasa a `conPlantilla`.
- **El motivo en la prosa** (§4; research D3): en toda skill con plantilla, sobre el `SKILL.md` regenerado.
- **Los textos** (contracts/skills.md; research V1): `consulta-juridica` §3 y §4; `plazos-y-territorio`, tres cambios;
  `administrativo.md`, la frase y los motivos.
- **Las evals** (contracts/evals.md; research D6 a D11, D18): el formato con `if`/`then`/`else`; `declarada` con
  `FueraDeLoDeclarado`; `activadaPor`; la descripción del umbral; la regla nueva; las evals 01, 03, 04 y 17; el art. 59 en
  el manifiesto de H9.
- **Las suites congeladas** (contracts/controles.md §2-§4; research D13, D14, D19): sus cuatro cambios.

**Aceptación e2e:** un guion testscript, `aceptacion/cobertura-instalada.txtar` (contracts/aceptacion.md §2), que
escribe la primera tarea y queda congelado; el workflow lo activa como
`internal/app/testdata/script/h33-cobertura-instalada.txtar`, y lo ejecuta `TestEntregaDelHito` en `make ci`. Sobre lo
que pone `kitlegal skills install --dir destino` comprueba la plantilla en los dos `SKILL.md`, la misma línea, y la marca
solo en `consulta-juridica`. Comprueba también que ninguna línea de la plantilla dice «procedimiento común», las cuatro
líneas de la región de lo cubierto, las 9 filas de la tabla como plantilla con su motivo (la de los tributos byte a
byte) y el tope de 10 860 caracteres del cuerpo de `plazos-y-territorio`. Cubre la historia 5: FR-025, y lo instalado de
FR-004, FR-007 y FR-010; SC-001, SC-002, SC-003 y SC-005. Las historias 1 a 4 son de `make ci` y del cierre. Lo que piden
—datos, regiones, juicios y umbrales— lo fijan los tests de contracts/controles.md con sus mutantes, y la medición del
cierre (FR-024, SC-004). FR-026 y SC-010 los garantiza el sandbox del run.

## Orden de implementación

De dentro afuera, con `make ci` en verde tras cada tarea:

1. `[aceptacion]`: el guion de contracts/aceptacion.md, que hoy falla por su primera aserción.
2. `[datos]` La entrada del ET (`a59`) en `internal/evals/testdata/grabaciones.json`; `grabar_datos` la graba tras el
   commit. Antes de grabarse no pone nada en rojo (research V6), y todo lo que lee el art. 59 va después.
3. `[datos]` Los datos de la cobertura: `schemas/cobertura.yaml.json`, `cobertura.go`, `data/cobertura.yaml` y
   `TestLeerCobertura`, en una sola tarea porque el test lee el fichero del repositorio.
4. `[datos]` La plantilla y los motivos, en una sola tarea porque, con el generador nuevo y la skill de hoy, la región
   de `administrativo.md` da su defecto:
   - el generador de `regiones.go` y `sincronia.go`, con `TestRegiones` y `TestRegionDeLaCobertura`; el `SKILL.md`
     sintético de `TestRegiones`, con la plantilla, dice ya en su prosa el motivo, «fuera de lo que kitlegal cubre»,
     para que el defecto del motivo del paso 5 no cambie sus casos;
   - el `SKILL.md` de `plazos-y-territorio` y `administrativo.md`, regenerada;
   - los dos guiones de H32 y el `ensayo.md` de la vertical, con sus motivos;
   - la declaración y `sinLoGenerado` de `TestSkillsDelProductoDesdeControl`.
5. `[datos]` La región de lo cubierto, en una sola tarea porque el `SKILL.md` de `consulta-juridica` cambia su huella:
   - `cubierto.go`, la materia y el motivo en la prosa, con `TestRegionDeLoCubierto`;
   - el `SKILL.md` de `consulta-juridica` y `partirElCuerpo`;
   - `TestConsultaJuridicaDesdeLaBase`;
   - la fila de `cubierto` de la vertical y los tres cambios de `TestVerticalDeEnsayo`.
6. `[datos]` El formato y el juicio, sin evals nuevas: el esquema de eval y `formato.go`, `declarada` y su motivo,
   `activadaPor`, `ComprobarLasSkillsNombradas` y la descripción del umbral. Con ellos, `TestFormatoDeLaCobertura`,
   `TestJuzgarLaDeclaracionDeNoCubierto`, `TestActivadaPor`, `TestSkillsNombradasEnTambienLaResuelven` y las constantes
   de los tests de juicio.
7. Las evals y lo que cuenta el conjunto. Van juntas porque la regla nueva solo se cumple con las cuatro evals y porque
   la matriz y el peor caso cambian una sola vez:
   - las evals 01, 03, 04 y 17 y la regla «frontera en dos materias»;
   - `TestEvalsDelRepositorio`, `TestUmbralesDeLasSkills`, `TestCacheDeLasSesiones`, `TestDefinicionDelJob` y
     `TestMatrizDeLasPropuestas`;
   - el comentario de `.github/workflows/evals.yml`.
8. La documentación de FR-022 y el `CHANGELOG.md`.

## Trazabilidad

| Mecanismo | FR/SC |
|---|---|
| `Cobertura.MotivoDeLoNoDeclarado`, `Cobertura.Cubierto`, `FilaCubierta`; las dos claves del esquema y de `data/cobertura.yaml` | FR-001, FR-002 |
| `TestLeerCobertura`: el subtest del repositorio y los seis mutantes | FR-001, FR-002, SC-008 |
| `regionDeLoCubierto` y su generador en `cubierto.go`; su llamada en `regenerarSkill` | FR-007, SC-001 |
| La comprobación de la `materia` | FR-003 |
| La comprobación del motivo en la prosa | FR-003 |
| `TestRegionDeLoCubierto` y sus tres mutantes | FR-003, FR-007, SC-001, SC-008 |
| `lineaDeNoCubierto`, `plantillaDelCuerpo`, `lineaDeNoCubierto.conLosMotivosDe`; `<asunto>` y `<motivo>` en los marcadores | FR-004 a FR-006 |
| Los dos defectos nuevos del generador y el del marcador de un motivo; `conPlantilla` | FR-006 |
| `TestRegiones`, `TestRegionDeLaCobertura` y sus mutantes | FR-006, SC-002, SC-008 |
| El texto de `consulta-juridica` (§3 y §4) | FR-004, FR-007, FR-008 |
| Los tres cambios de `plazos-y-territorio` y el de `administrativo.md` | FR-004, FR-005, FR-009, FR-010 |
| `partirElCuerpo` sin partir la región | FR-007 |
| La fila y los motivos de la vertical; los tres cambios de `TestVerticalDeEnsayo` | FR-011, FR-010 |
| `NoCubiertoEsperado.FueraDeLoDeclarado`, `esDeCobertura`; el `if`/`then`/`else` del esquema; `TestFormatoDeLaCobertura` | FR-012 |
| `declarada` con `FueraDeLoDeclarado`, `motivoFueraDeLoDeclarado`; `TestJuzgarLaDeclaracionDeNoCubierto` | FR-013, SC-008 |
| La constante de la descripción de `cobertura_sin_declarar`; `TestUmbralesDeLasSkills` | FR-014 |
| `Eval.TambienLaResuelven`, `activadaPor`, la clave en `ComprobarLasSkillsNombradas`; sus dos tests | FR-015 |
| Las evals 01, 03 y 04; la 17 | FR-016, FR-015 |
| `incumplimientoDeLaFronteraEnDosMaterias` y su mutante | FR-017, SC-007, SC-008 |
| La entrada del manifiesto de H9 y el subtest de `TestCacheDeLasSesiones` | FR-018 |
| Los valores de `TestDefinicionDelJob` y la golden de `TestMatrizDeLasPropuestas`; el comentario de `evals.yml` | FR-019, SC-006 |
| `cambioEnUnaReferencia`, `enLasReferencias` y las regiones de las referencias en `sinLoGenerado`; la declaración de `plazos-y-territorio`; la huella, `regionesGeneradas` y los mensajes de `TestConsultaJuridicaDesdeLaBase`; los dos guiones de H32 | FR-020, SC-009 |
| Las constantes de los tests de juicio con la plantilla | FR-021 |
| La documentación y el `CHANGELOG.md` | FR-022 |
| `aceptacion/cobertura-instalada.txtar` y su `filas.awk` | FR-025, SC-001 a SC-003, SC-005 (III; ADR 0018) |

## Uso, de fuera adentro

Detalle con sus bytes en contracts/datos-y-regiones.md §7 y contracts/evals.md §8; todo medido en research V9.

- **La región de lo cubierto.** La lee el agente una vez por sesión, al cargar `consulta-juridica`, y con ella decide
  §4. Son 4 líneas, 763 bytes (la del territorio, 117: `- **Territorio**: la provincia, la comunidad, el régimen, el DIR3
  del ayuntamiento y los boletines de un municipio.`). Crece una línea por fila que traiga un paquete: no crece con el
  uso, y con meses de uso diario, cientos de normas y miles de bloques consultados, nada de la caché ni del grafo entra
  en ella. No es una señal que caduque: cambia con los datos.
- **La línea `⚠ NO CUBIERTO:` de una respuesta.** Va una por asunto no cubierto que la pregunta pide y ninguna por lo
  que no pide; mide de unos 100 a 350 bytes. Ninguna orden más por pregunta: §4 no pide ninguna y `plazos calcular`
  queda prohibido en otra materia (la 04 hace una lectura, `boe articulo` en la fecha). Deja de darse cuando lo pedido
  entra en lo cubierto: un paquete con su fila o una capacidad de otro hito.
- **La marca `⚠ NO VERIFICADO:`.** Va una por lectura hecha con la skill de normativa para completar, solo tras la línea
  fuera de lo declarado; ninguna en lo regulado aparte. Su forma mide 140 bytes, y cada marca, su cita y lo encontrado.
  Deja de darse con su línea, cuando lo pedido entra en lo cubierto.
- **La tabla `cobertura_por_materia`.** La lee el agente de `plazos-y-territorio` una vez por pregunta de plazo de su
  paquete. Son 9 filas, 3128 bytes (hoy 3047: el paréntesis del spec que la daba por más corta no se cumple, research
  V9). Crece con las filas de `regulados_aparte` de los paquetes, no con el uso.
- **Los defectos de la sincronía.** Los lee quien mantiene kitlegal cuando `make skills-sync` o `make ci` salen en rojo:
  una línea por defecto, ninguna con los datos del repositorio.
- **Los umbrales.** Cuatro más en el informe de `consulta-juridica`, unos 1800 bytes (el ejemplo, 446). Los lee la
  persona una vez por medición; fijos con el uso.

## Gates: controles mecánicos que añade o toca

- **Objetivos de `make`**: ninguno nuevo ni cambiado; `make test` y `make test-integration` (en `make ci`) corren todo.
- **Tests nuevos**: `TestRegionDeLoCubierto` (`internal/skills/cubierto_test.go`), `TestFormatoDeLaCobertura`,
  `TestJuzgarLaDeclaracionDeNoCubierto`, `TestActivadaPor` y `TestSkillsNombradasEnTambienLaResuelven`.
- **Tests que cambian** (contracts/controles.md): los de FR-021 y las suites congeladas de FR-020.
- **Tests de contrato de los esquemas**: `TestLeerCobertura` (seis mutantes) y `TestFormatoDeLaCobertura`; los demás
  formatos, sin cambios (`TestFormatoDePlazos`).
- **Guion e2e**: `h33-cobertura-instalada.txtar`, congelado. Los de H32 cambian sus datos declarados.
- **Cobertura por capacidad y materia en evals** (constitución, «Gates»): un asunto que lo cubierto resuelve y otro de
  fuera de lo declarado en dos materias, con la regla que lo exige; lo regulado aparte, en las 07, 08 y 12.
- **Fixtures**: una grabación nueva (el art. 59), del paso `grabar_datos`; los ficheros de la vertical de ensayo.
- **Configuración de verificación**: `.golangci.yml`, `Makefile` y `.github/workflows/ci.yml` no cambian; de
  `evals.yml`, solo un comentario. Ningún `//nolint`, `t.Skip` ni `TODO`.

## Datos externos

| Fuente | Fila en `docs/SOURCES.md` de `main` | Manifiesto | Test de grabación | Destino |
|---|---|---|---|---|
| `boe.legislacion-consolidada` | sí, «Revisado» 2026-09-13 | `internal/evals/testdata/grabaciones.json` (`"fuente": "boe.legislacion-consolidada"`), una entrada al final de `normas`: la búsqueda «estatuto de los trabajadores», `titulo_empieza_por` «Real Decreto Legislativo 2/2015,», bloque `a59` (contracts/evals.md §6) | `TestGrabarEvals`, `//go:build grabacion`, `internal/evals/grabacion_test.go`, por `internal/httpx` | la respuesta, en `internal/evals/testdata/boe.legislacion-consolidada/`; ningún material de origen: nada de `data/` sale de ella, y `$KITLEGAL_EVIDENCIAS` (`evidencias/H33/`) no recibe nada |

El manifiesto de H5 (`testdata/evals/grabaciones.json`), donde lo puso el prototipo, no lo detecta `grabar_datos`
(research V5, D11). Los textos nuevos de `data/cobertura.yaml` no son datos de una fuente (research D12). Ninguna fuente
sin fila revisada entra en el hito.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| FR-001 | 0 diferencias entre el motivo y las cuatro filas de `cubierto` del repositorio y los textos de FR-001 | El subtest `repositorio` | ci:internal/skills/cobertura_test.go:TestLeerCobertura |
| FR-002, SC-008 | 0 de 6 mutantes del esquema aceptados, cada uno con su defecto | Los casos de contracts/datos-y-regiones.md §2 | ci:internal/skills/cobertura_test.go:TestLeerCobertura |
| FR-003, SC-008 | 0 de 3 mutantes de la sincronía sin su defecto exacto | La materia, sin `data/cobertura.yaml` y el motivo cambiado | ci:internal/skills/cubierto_test.go:TestRegionDeLoCubierto |
| FR-007, SC-001 | 4 de 4 filas en la región, una línea cada una, en su orden | Las líneas entre las marcas tras regenerar | ci:internal/skills/cubierto_test.go:TestRegionDeLoCubierto |
| FR-007, FR-023, SC-001 | 0 ficheros que cambie `make skills-sync` | El subtest `deriva` | ci:internal/app/skills_test.go:TestSkillsDelRepositorio |
| FR-004, SC-003 | 0 diferencias entre la plantilla de `consulta-juridica` y la de `plazos-y-territorio` | El subtest `formas-compartidas` | ci:internal/app/skills_test.go:TestSkillsDelRepositorio |
| FR-006, SC-002 | 0 bytes de diferencia en la fila de los tributos; 9 de 9 filas del paquete | `filaDeLosTributos` y la referencia generada | ci:internal/skills/referencias_test.go:TestRegionDeLaCobertura |
| FR-006, SC-008 | 0 de 5 mutantes del generador sin su defecto exacto | contracts/controles.md §1 | ci:internal/skills/regiones_test.go:TestRegiones |
| FR-005, FR-009, FR-020 | 0 bytes de diferencia en `plazos-y-territorio` fuera de lo generado y de su declaración | El subtest de `plazos-y-territorio` | ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl |
| FR-008, FR-020 | 0 bytes de diferencia con la huella de la entrega, con las dos regiones reducidas | La huella y sus mutantes | ci:internal/evals/descripciones_del_producto_test.go:TestConsultaJuridicaDesdeLaBase |
| FR-010, SC-005 | Como mucho 11 000 caracteres cada cuerpo de `skills/` | `defectosDelCuerpo` en la sincronía | ci:internal/app/skills_test.go:TestSkillsDelRepositorio |
| FR-007, FR-010, FR-011, SC-005 | Como mucho 11 000 caracteres con la vertical de ensayo; su línea en la región de lo cubierto, la segunda de los plazos, detrás de las cuatro del repositorio | Los tres cambios del test | ci:internal/skills/vertical_de_ensayo_test.go:TestVerticalDeEnsayo |
| FR-012, SC-008 | 0 evals con las dos variantes de `no_cubierto` aceptadas; 0 mensajes cambiados en `{norma, precepto}` | El mutante y los casos de hoy | ci:internal/evals/formato_test.go:TestFormatoDeLaCobertura, ci:internal/evals/formato_test.go:TestFormatoDePlazos |
| FR-013, SC-008 | 0 de 11 respuestas clasificadas al revés o con otro motivo | contracts/evals.md §2 | ci:internal/evals/plazos_test.go:TestJuzgarLaDeclaracionDeNoCubierto |
| FR-014 | `sin_activar` y `cobertura_sin_declarar` de `consulta-juridica`, en los dos modos, con `Decide: true` y delante de los `fuera_de_lo_grabado` | El subtest `consulta-juridica` | ci:internal/evals/informe_test.go:TestUmbralesDeLasSkills |
| FR-015 | 0 activaciones de una skill de `tambien_la_resuelven` sin contar; 0 nombres ajenos a `skills/` aceptados | Los dos tests | ci:internal/evals/tambien_la_resuelven_test.go:TestActivadaPor, ci:internal/evals/tambien_la_resuelven_test.go:TestSkillsNombradasEnTambienLaResuelven |
| FR-016, FR-017, SC-007, SC-008 | Las cuatro evals con su formato; al menos 2 normas sin `vertical` con una eval a cada lado; el mutante sin la 03 y el de «no activación» sobre la 02 y la 03, cada uno con su defecto | «formato», la regla y los dos mutantes | ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio |
| FR-018 | 0 consultas de las evals de `consulta-juridica` sin servir desde lo grabado | El subtest `consulta-juridica` | ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones |
| FR-019, SC-006 | Como mucho 21 120 s el peor caso del trabajo de `consulta-juridica` (15 013 s) | El peor caso recalculado | ci:internal/evals/definicion_test.go:TestDefinicionDelJob |
| FR-019 | 0 bytes de diferencia con la matriz de cada ejecución | La golden | ci:internal/evals/eleccion_test.go:TestMatrizDeLasPropuestas |
| FR-020, FR-021, SC-009 | 0 tests en rojo, con y sin la etiqueta `integration` | `make test` y `make test-integration`, en `make ci` | ci:Makefile:test, ci:Makefile:test-integration |
| FR-004, FR-025, SC-001, SC-002, SC-003, SC-005 | 100 % de las aserciones del guion: la plantilla en 2 de 2 skills, 0 veces «procedimiento común», la marca en 1, las 4 líneas de lo cubierto, 9 de 9 filas, y el cuerpo de `plazos-y-territorio` de como mucho 10 860 caracteres | El guion congelado | ci:internal/app/e2e_test.go:TestEntregaDelHito |
| FR-014, FR-024, SC-004 | 0 en `sin_activar` y `cobertura_sin_declarar` de `consulta-juridica`, en los dos modos | El job de evals del cierre | evals:consulta-juridica:sin_activar:claude-sonnet-5-5:orden, evals:consulta-juridica:cobertura_sin_declarar:claude-sonnet-5-5:orden, evals:consulta-juridica:sin_activar:claude-sonnet-5-5:herramienta, evals:consulta-juridica:cobertura_sin_declarar:claude-sonnet-5-5:herramienta |
| FR-014, FR-024, SC-004 | 0 en `cobertura_sin_declarar` de `plazos-y-territorio`, en los dos modos | El job de evals del cierre | evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta |
| FR-024 | Las evals de `plazos-y-territorio` en verde en el cierre, en los dos modos | El job de evals del cierre | evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta |
| FR-024 | Las evals de `normativa` en verde en el cierre, en los dos modos | El job de evals del cierre | evals:normativa:sin_activar:claude-sonnet-5-5:orden, evals:normativa:sin_activar:claude-sonnet-5-5:herramienta, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5, evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar, evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados, evals:normativa:duracion_de_las_sesiones:orden, evals:normativa:duracion_de_las_sesiones:herramienta, evals:normativa:duracion_del_juez:orden, evals:normativa:duracion_del_juez:herramienta |
| FR-024 | Las evals de `jurisprudencia` en verde en el cierre, en los dos modos | El job de evals del cierre | evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden, evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5, evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar, evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados, evals:jurisprudencia:duracion_del_juez:orden, evals:jurisprudencia:duracion_del_juez:herramienta |

El modelo es `MODELO_DE_EVALS` de `.github/workflows/evals.yml` (`claude-sonnet-5-5`), que el hito no cambia. Los de
`normativa`, `jurisprudencia` y `plazos-y-territorio` son los que publican hoy con `decide: true`. El hito no cambia
ninguno: son los del cierre de H41.4, sin el de la clase del juez de las descripciones, porque ninguna `description`
cambia. Los nuevos de `consulta-juridica` nacen de sus evals de cobertura (research V12). El veredicto del job sale en
rojo con cualquiera de ellos por encima de su umbral.

Sin control en `make ci` ni en el cierre, por el propio spec o porque no son un umbral que se mida sin persona:

- FR-022, la documentación: la lee la revisión, y el quickstart la enseña en su escenario 9.
- FR-026 y SC-010: el run no abre sesiones con modelo ni sondeos, y lo garantiza el sandbox de los pasos (ADR 0032),
  configuración que el hito no toca.
- La segunda mitad de SC-009, «4 suites congeladas cambiadas y ninguna más»: la guardan las rutas que declara cada tarea
  y el guardián de diff del workflow, y la revisión lo lee en el diff. El quickstart enseña, en su escenario 5, que las
  copias de H32 no cambian.

## Cierre y pendientes

- **FR-024**: si la medición del cierre da un rojo, la reparación no toca para que pase ningún texto de las skills ni de
  `administrativo.md`, ni las evals ni los umbrales: lo deja dicho en el informe para la persona. Ningún umbral se rebaja.
- **Para la persona, en el informe final**: las `lecturas` de las evals 03 y 04, que declara tras ver una sesión (ADR
  0043, punto 2; FR-016). Los límites conocidos de «Fuera de alcance»: una línea de más en la 02 o la 03, y una fecha
  contada a mano junto a la línea. Y research S1 a S3.

## Complexity Tracking

Ninguna desviación: ni dependencias fuera de la lista, ni violaciones de la constitución que justificar.
