# Implementation Plan: H26.1 · Las evals sirven lo que la skill lleva a leer y hacen fallar a propósito la lectura que quieren probar

**Branch**: `025-h26-1-las-evals-sirven` | **Date**: 2026-10-21 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/025-h26-1-las-evals-sirven/spec.md`, la sección «H26.1» de
`docs/ROADMAP.md` y el ADR 0043.

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

El hito protege `legal-core` y `boe-legislacion` (principio VIII) cambiando solo el arnés de evals (`internal/evals`),
su esquema, sus evals, una grabación y el texto de `legal-core`; ningún paquete del binario cambia (research D8, D15).

1. **La caché de cada sesión sirve lo que la skill lleva a leer** (FR-001 a FR-005): a las consultas necesarias de las
   evals de su skill se suman, con el mismo acumulador, cada bloque de `data/plazos.yaml` y de `data/cobertura.yaml`
   —sin fecha y con `--fecha 9999-12-31`— y el índice y los metadatos de cada norma de esos dos ficheros y de
   `data/normas.yaml`, leídos con `plazos.LeerTipos`, `skills.LeerCobertura` y `skills.LeerNormas` (research D1). Hoy son
   17 bloques y 19 normas, 72 consultas, todas grabadas (research V1). La base de la medida del juez sirve lo mismo,
   sin las `lecturas` de las evals (FR-004, D4).
2. **`lecturas`** (FR-006 a FR-011): cuarto punto de las consultas necesarias (D2); el esquema y `LeerEval` validan su
   forma y sus cruces (D9); una tarea `[datos]` añade al manifiesto de `internal/evals/testdata/` cuatro bloques que no
   están grabados (V2), y las evals 14, 15 y 16 de `boe-legislacion` los llevan.
3. **`sin_servir`** (FR-012 a FR-016): la preparación de la sesión de esa eval quita sus consultas (D3; research V3: el
   bloque sale entonces con 4 en toda fecha), el juicio reparte la invocación que lo consulta en `lecturas_sin_servir`
   (D6) y el comando de bloque que la eval hace fallar admite 4 y 5 (D5). Dos evals nuevas de `legal-core`, la 11 y la
   12 (FR-017, FR-018), y un umbral sin modelo nuevo que decide, `lectura_sin_intentar:<modelo>:<modo>`, solo en el
   informe de la skill con una eval con un comando de bloque en su `sin_servir`, hoy `legal-core` (FR-033).
4. **`fuera_de_lo_grabado:<modelo>:<modo>`**, publicado con `decide: false` (FR-021, FR-022), y el **motivo de un 1 que
   no es de kitlegal** (FR-023 a FR-028): la llamada sin sobre sale con `codigo: null` y `motivo: "sin_sobre"` (D7), y
   la orden que no pudo escribir su salida estándar, enlazada con su uso de Bash, con `motivo: "salida_no_escrita"` y
   `bash`; la línea se reconoce con el error del propio presentador del binario (D8, D10).
5. **`legal-core` v1.2.3** (FR-029 a FR-031): los pasos 7 y 10, en 299 líneas como mucho (D14).
6. **Documentación** (FR-032): `CHANGELOG.md`, `CONTRIBUTING.md`; `CLAUDE.md` y `docs/WORKFLOW.md` no dicen nada de la
   caché de las sesiones (V18).

## Technical Context

**Language/Version**: Go 1.27 (`go.mod`: `go 1.27.0`, `toolchain go1.27.2`; `go version` de la sesión, go1.27.2).

**Primary Dependencies**: ninguna nueva. Las que ya usa `internal/evals`: `gopkg.in/yaml.v3` (por
`skills.ValidarDocumentoYAML`), `santhosh-tekuri/jsonschema` v6.0.3 (el esquema de eval), `stretchr/testify` en los
tests, `encoding/json/v2` de la biblioteca estándar (el informe) y los paquetes del módulo `internal/app`, `cli`,
`cache`, `core/plazos`, `skills`, `source/boe`. Desde este hito, también `internal/render` (research V17, D8).

**Storage**: la caché SQLite de cada sesión, que solo escribe el binario en proceso (`internal/cache`); `internal/evals`
no abre SQLite (R3). Ficheros: `schemas/eval.yaml.json`, `evals/<skill>/*.yaml`, `internal/evals/testdata/` y el
`informe.json` del job.

**Testing**: `go test -race` con `stretchr/testify`, en `make ci`, sin red y sin modelo; caché preparada desde
`UnionDeGrabaciones()`; sesiones sintéticas en `t.TempDir()`; mutantes en un worktree fuera del árbol (research D16).

**Target Platform**: la de `make ci` (macOS y Linux) y el runner `ubuntu-24.04` del job de evals.

**Project Type**: CLI multicall con skills agénticas; el hito toca el arnés de evals, que no viaja en el binario
(research V5).

**Performance Goals**: ninguno nuevo. La preparación de cada sesión gana 72 consultas en proceso, del orden de un
milisegundo cada una (V1), frente a los 22 s que el peor caso le da (V10).

**Constraints**: ningún test de `make ci` usa la red ni abre una sesión con modelo; ningún `//nolint`, `t.Skip` ni
`TODO`; el binario no cambia (`internal/render`, `internal/cli`, `internal/app`, `cmd/` quedan igual); la regla de 2 de
3, las repeticiones, los modelos, la concurrencia, los umbrales del juez y el tope del job no cambian (spec, «Fuera de
alcance»); `skills/legal-core/SKILL.md` en 299 líneas como mucho (V9).

**Scale/Scope**: 72 consultas de `data/` por sesión; 4 bloques de `lecturas` en 3 evals; 2 evals nuevas, 24 sesiones
más por job (2 × 3 × 2 × 2); 8 umbrales más en `legal-core` y 4 en `boe-legislacion` y en `jurisprudencia`; un peor
caso de `legal-core` de 19 525 s (V10).

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones; la única desviación
justificada está en Complexity Tracking.*

### Principios

| Principio | Cómo lo cumple H26.1 |
|---|---|
| I · Fuentes públicas y frontera humana | La única fuente es `boe.legislacion-consolidada`, con fila revisada en `main` (2026-09-13). Lo graba `grabar_datos` con `TestGrabarEvals` por `internal/httpx`, sin modelo; ninguna tarea toca la red ni usa `KITLEGAL_RECORD` (research V11). Ninguna petición nueva del binario; solo GET |
| II · Nada sin cita ni fuente | Ningún sobre cambia. La 12 mide que lo que la ley regula aparte se declara no cubierto aunque falle la lectura, y la v1.2.3 dice que tras esa línea no se expone la norma del procedimiento (FR-030). La 11 mide que la fecha va con la cita de sus preceptos y que no se afirma lo no leído (FR-029). Los motivos de `lecturas` no afirman nada de ningún bloque |
| III · Tests primero y offline | «Aceptación e2e: no aplica» (abajo, con el motivo; Complexity Tracking). Cada pieza entra con sus tests en `make ci`, offline y sin modelo; la aceptación es el job de cierre (SC-011 a SC-020) |
| IV · Hexagonal y errores tipados | Ningún paquete del binario, error tipado ni código de salida cambia; FR-028 lo exige y el escenario 9 del quickstart lo comprueba. El arnés pide a `internal/render` su error, no lo cambia (D8). Lo que no se lee —un fichero de `data/` roto— es un error de la preparación y la sesión no se abre |
| V · Simplicidad y dependencias | Ninguna dependencia nueva. Dos tipos exportados, uno por cada clave de YAML nueva (`LecturaDeLaEval`, `LecturaSinServir`); diez campos exportados en tipos que ya existen, los que se leen del YAML, se escriben en el informe o rellena la lectura de la sesión (data-model §1, §3, §5 y §6); todo lo demás, sin exportar. Ningún paquete ni fichero de código nuevo (Trazabilidad) |
| VI · Un binario, convenciones de agente | El binario, sus banderas, `--describe` y la tabla de comandos de las skills no cambian |
| VII · Grafo y privacidad | Nada del grafo cambia. Las sesiones sintéticas viven en directorios temporales |
| VIII · Skills primero | Protege `legal-core` (su robustez ante una lectura fallida, medida a propósito) y `boe-legislacion` (sus lecturas de la clase (b), servidas), y corrige `legal-core` v1.2.3. `SKILL.md` sigue sin código, con su tabla generada, por debajo de 300 líneas |
| IX · Genericidad territorial | Nada se particulariza para un municipio: lo que se sirve sale de `data/` sin ningún identificador en el código (FR-002). Las evals usan Leganés, el municipio de validación, como dato de la eval |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | Sin cambios en `internal/core`; `internal/evals` usa `core/plazos` como hoy |
| R2 · Solo `internal/httpx` importa `net/http` | Ningún import nuevo de `net/http`; la grabación es la de `TestGrabarEvals`, por `internal/httpx` |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | `internal/evals` sigue sin SQLite: `sin_servir` quita consultas antes de preparar, no borra entradas (D3) |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Ningún `os.Exit` nuevo |
| R5 · Solo `render` escribe en stdout | Ningún `fmt.Print*`, `os.Stdout` ni `os.Stderr` nuevo: el presentador que el arnés construye escribe en un escritor que falla (D8). Importar `internal/render` desde `internal/evals` lo admiten `depguard` y `internal/arch_test.go` (V17) |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios en `internal/graph` |
| R7 · Solo el paquete del servidor importa el SDK de MCP | Ningún import del SDK |

## Project Structure

### Documentation (this feature)

```text
specs/025-h26-1-las-evals-sirven/
├── plan.md                   # este fichero
├── research.md               # V1-V21, D1-D17, S1-S3
├── data-model.md             # tipos, campos, reparto, umbrales, evals y manifiesto
├── quickstart.md             # escenarios 0-10
├── contracts/
│   ├── formato-de-eval.md    # esquema, rechazos, evals 11 y 12, lecturas de 14-16, manifiesto
│   ├── cache-de-la-sesion.md # qué consultas se preparan y qué se sirve
│   ├── juicio-e-informe.md   # reparto, FR-016, salida no escrita, umbrales, serie, tamaños
│   └── skill-legal-core.md   # pasos 7 y 10 de la v1.2.3
├── mutantes.md               # lo escriben las tareas (research D16)
└── tasks.md                  # /speckit-tasks
```

### Source Code (repository root)

```text
schemas/eval.yaml.json                         # lecturas, sin_servir ([datos])
evals/legal-core/11-plazo-reposicion-sin-leer-el-precepto.yaml          # nueva
evals/legal-core/12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml  # nueva
evals/boe-legislacion/{14,15,16}-*-por-materia.yaml                      # lecturas
internal/evals/
├── formato.go          # LecturaDeLaEval, LecturaSinServir, Eval.Lecturas/SinServir, cruces en LeerEval
├── consultas.go        # PuntoLecturaDeLaEval, PuntoLoQueLlevaALeer, anotarLosDatos, consultasDeLaCache, sinLoQueNoSeSirve
├── preparar.go         # prepararLaCache con consultasDeLaCache y sin_servir; queFaltaDesde
├── medida.go           # leerLasEvalsDeHoy con consultasDeLaCache de las evals sin sus lecturas (sinSusLecturas)
├── sesion.go           # Llamada.SinSobre, Sesion.OrdenesDeBash
├── juzgar.go           # reparto, lecturas_sin_servir, motivo y bash, FR-016, prefijo de la salida no escrita
├── informe.go          # sesiones_con_lectura_sin_servir, fila de la llamada sin sobre en informe.md
├── umbrales.go         # fuera_de_lo_grabado, lectura_sin_intentar, su motivo
├── doc.go              # el contrato del paquete, alineado
├── testdata/grabaciones.json                   # tres entradas ([datos])
├── testdata/boe.legislacion-consolidada/       # cuatro respuestas nuevas (grabar_datos)
└── *_test.go           # ver «Controles mecánicos»
skills/legal-core/SKILL.md                     # v1.2.3
CHANGELOG.md, CONTRIBUTING.md
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquete ni fichero de código nuevo: cada pieza va
al fichero de `internal/evals` que ya tiene la suya (las consultas, la preparación, la lectura de la sesión, el juicio,
el informe y los umbrales).

## Aceptación e2e

**Aceptación e2e: no aplica.** El hito no cambia el código del binario: `cmd/kitlegal` no enlaza `internal/evals`
(research V5), `internal/render` no se toca (D8) y FR-028 prohíbe cambiar el código con que termina. El `SKILL.md` de
`legal-core` viaja empotrado y `kitlegal skills install` lo instala igual que hoy: los guiones de H19 ya comparan lo
instalado con el árbol, sea cual sea su texto (V5). No hay comportamiento de `kitlegal` que un guion `testscript`
describa (D15). La aceptación es la del instrumento y la de la skill: el job de evals de cierre (SC-011 a SC-020). En
`make ci` la fijan, sin modelo:

| Historia | Tests | Requisitos |
|---|---|---|
| US1 | `TestEvalsDelRepositorio` (subtest nuevo `lo-que-lleva-a-leer`, `grabado`, `conjunto-legal-core`), `TestLoQueLlevaALeerSinGrabar`, `TestBaseDeLaReconstruccion` | FR-001 a FR-005; SC-003, SC-008, SC-011 |
| US2 | `TestCacheDeLasSesiones` (`sin-servir`), `TestFormatoDeLecturas`, `TestJuzgarLasLecturasSinServir`, `TestUmbralesDeLasLecturas`, `TestEvalsDelRepositorio` | FR-012 a FR-019, FR-033; SC-005, SC-008, SC-012 a SC-014 |
| US3 | `TestCacheDeLasSesiones` (`lecturas`), `TestFormatoDeLecturas`, `TestJuzgarLasLecturasSinServir`, `TestConsultasNecesarias`, `TestEvalsDelRepositorio/grabado` | FR-006 a FR-011; SC-004, SC-018 |
| US4 | `TestUmbralesDeLasLecturas`, `TestUmbralesDeLasTresSkills` | FR-021, FR-022; SC-006, SC-015 |
| US5 | `TestJuzgarElMotivoDeLasFallidas`, `TestLineaDeLaSalidaNoEscrita`, `TestLeerLasLlamadas`, `TestTextosDeLaSesion`, `TestTuberiaCerrada` | FR-023 a FR-028; SC-007, SC-017 |
| US6 | `make skills-check` (`TestSkillsDelRepositorio`); el texto lo lee la revisión final | FR-029 a FR-031; SC-021 |

Ninguna tarea `[aceptacion]`.

## Controles mecánicos que este hito añade o toca

Capa 1 de la constitución («Gates»): todo en `make ci`, sin red salvo `vuln` y sin modelo.

### Objetivos del `Makefile`

Ninguno nuevo ni cambiado. Los que lo ejercen: `make test` (todos los tests de abajo), `make skills-check`
(`TestEvalsDelRepositorio` y `TestSkillsDelRepositorio`, `Makefile:125-127`), `make schema-check` (sin cambios: no lee
las evals, V8) y `make ci`.

### Tests nuevos

| Test | Fichero | Qué fija | Requisitos |
|---|---|---|---|
| `TestEvalsDelRepositorio/lo-que-lleva-a-leer` | `internal/evals/conjunto_test.go` | desde la caché de una sesión de una eval de cada skill, cada bloque de los dos ficheros, sin fecha y con `9999-12-31`, y el índice y los metadatos de cada norma de los tres, con 0 y `--offline`; los enumera el oráculo de los tests de hoy (`lecturasDeLosTipos`, `lecturasDeLaCobertura`, `grabaciones_test.go:1033-1102`), no el código que prueba | FR-001, FR-002, FR-005; SC-003, SC-008, SC-011 |
| `TestLoQueLlevaALeerSinGrabar` | `internal/evals/preparar_test.go` | con una copia de `data/` con `a128` cambiado por `a129`, una falta que nombra la lectura y su origen | FR-003; SC-003 |
| `TestCacheDeLasSesiones` | `internal/evals/preparar_test.go` | `sin-servir`: la 11 no sirve `a124` sin fecha, con `--fecha 2026-09-09` ni con `9999-12-31`, y la 05 sí; la 12 no sirve `daprimera` en esas tres fechas y la 07 sí (US1, escenario 1; US2, escenario 1). `lecturas`: la caché de otra eval de `boe-legislacion` sirve los cuatro bloques sin fecha y con `9999-12-31` | FR-001, FR-006, FR-011, FR-013; SC-004, SC-005, SC-008, SC-018 |
| `TestFormatoDeLecturas` | `internal/evals/formato_test.go` | la lectura y cada rechazo de contracts/formato-de-eval.md §1 y §2 | FR-008, FR-009, FR-012, FR-014; SC-004, SC-005 |
| `TestJuzgarLasLecturasSinServir` | `internal/evals/juzgar_test.go` | el reparto de las filas 3 a 7, FR-016 con la fecha del comando y sin ella, las sesiones de la 11 y la 12 de los escenarios de US2, y la de la 16 sin leer sus `lecturas` | FR-007, FR-015 a FR-018; SC-004, SC-005 |
| `TestJuzgarElMotivoDeLasFallidas` | `internal/evals/juzgar_test.go` | la llamada sin sobre (también de una lectura de `sin_servir`), la orden con `\| python3` enlazada (también de una lectura de `sin_servir`), el 1 con otra salida, sin uso y con dos; el juicio igual que antes salvo lo publicado | FR-023, FR-024, FR-026 a FR-028; SC-007, SC-017 |
| `TestLineaDeLaSalidaNoEscrita` | `internal/evals/juzgar_test.go` | la salida de error de `app.Main` con una salida estándar que falla es la línea que el arnés reconoce | FR-025; SC-007 |
| `TestUmbralesDeLasLecturas` | `internal/evals/umbrales_test.go` | `fuera_de_lo_grabado` (US4), `lectura_sin_intentar` con su motivo y su veredicto, el del modelo informativo, su ausencia en una skill sin eval con un comando de bloque en su `sin_servir`, y `sesiones_con_lectura_sin_servir` (US2, escenario 5) | FR-015, FR-021, FR-022, FR-033; SC-005, SC-006, SC-008, SC-013 |

Casos nuevos en tests que existen: `TestConsultasNecesarias` (`lecturas`, punto `lectura de la eval`; FR-006),
`TestBaseDeLaReconstruccion` (la base sirve `daprimera` y no sirve lo que las sesiones de los casos etiquetados no
tenían; FR-004), `TestLeerLasLlamadas` (`SinSobre`; FR-023) y
`TestTextosDeLaSesion` (`OrdenesDeBash`; FR-026).

### Tests existentes que cambian lo que esperan

Por datos nuevos (memoria del proyecto: una lista literal se rompe al añadir datos): `TestEvalsDelRepositorio` (doce
evals de `legal-core`, `evalsDeLegalCoreConLaDeVigencia`, `conjunto_test.go:2064`); `TestDefinicionDelJob`
(`peorCasoDeLegalCore`, de 16 261 s a 19 525 s, y el caso con el tope en 288 minutos, que también deja fuera a
`legal-core`; D13); los que arman sesiones de las evals de `legal-core` del repositorio y fijan sus totales
(`TestUmbralesDeLasTresSkills`, `TestUmbralesDeLegalCore`, `TestJuzgarPlazos`, por `informe_test.go` y `plazos_test.go`).
Por los umbrales nuevos (8 en `legal-core`, 4 en las otras dos) y los campos nuevos del informe: los que comparan la lista entera de umbrales o el JSON de una
sesión o de una serie (`TestUmbralesDelInforme`, `TestUmbralesDeBoeLegislacion`, `TestUmbralesDeJurisprudencia`,
`TestUmbralesConJuezYConSentencias`, `TestInformeEnDosModos`, `TestInformeConElJuez`, `TestInforme`) y los de la base de
la medida que cuentan sus consultas. Ninguno cambia lo que comprueba de lo que ya había: solo se le añade lo nuevo.

### Puntos de entrada fuera de `make ci`

`TestEjecucionDelJob` (etiqueta `evals`) prepara cada sesión con `PrepararSesion` (`abrirSesion`,
`internal/evals/sesiones.go:389-421`), la juzga con `Juzgar` y escribe el informe; el sondeo prepara igual
(`ejecutarSesiones`, `sondeo.go:1128`), juzga con `Juzgar` (`sondeo.go:530`) y reparte con `repartirEnSeries`
(`sondeo.go:334`). Los dos ganan lo de este hito sin cambiar su código de entrada; el sondeo no escribe `umbrales`, y
nada de él cambia salvo lo que publica de cada sesión y de cada serie. `TestGrabarEvals` (etiqueta `grabacion`) no
cambia: solo su manifiesto.

### Fixtures, `testdata/` y `schemas/`

- `schemas/eval.yaml.json`: `lecturas`, `sin_servir` y dos `$defs` (contracts/formato-de-eval.md §1), en una tarea
  `[datos]`.
- `internal/evals/testdata/grabaciones.json`: tres entradas (§4 del mismo contrato), en otra tarea `[datos]`; tras su
  commit, `grabar_datos` deja cuatro respuestas en `internal/evals/testdata/boe.legislacion-consolidada/`.
- Ningún otro fichero de `testdata/` cambia; las sesiones sintéticas se escriben en `t.TempDir()`.

### CI

Ningún flujo cambia (`.github/` queda igual, D13).

## Controles de umbral

Una fila por cada requisito o criterio con un umbral que se mide sin persona. En el cierre, los umbrales del informe del
job con `decide: true`: los de las evals 11 y 12 y los que ya decidían en las tres skills (SC-002); en `make ci`, cada
umbral de los controles del hito. Las series de la 11 y la 12 deciden además con la regla de 2 de 3 de siempre (ADR
0016), que no sustituye a ninguno de estos umbrales. Las respuestas medidas de `legal-core` por modo, con las doce evals:
30 de las evals que activan la skill; 12 de las de plazo (05, 06, 09, 11); 9 de las de cobertura (07, 08, 12); 3 de la
de vigencia (10); 9 de las de redacción fechada (05, 06, 09). En `lectura_sin_intentar` no son respuestas sino sesiones
medidas (research D11): las 3 de la 11 por modelo y modo.

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| FR-033, FR-019, SC-012, SC-013, SC-014 (modo orden) | 0 sesiones sin la lectura de `a124` intentada, de las 3 sesiones medidas de la 11 | umbral del informe con `decide: true`: con una, motivo con la sesión, veredicto `fallo` y el trabajo de `legal-core` en rojo, que el cierre cuenta | `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:orden` |
| FR-033, FR-019, SC-012, SC-013, SC-014 (modo herramienta) | ídem | ídem | `evals:legal-core:lectura_sin_intentar:claude-sonnet-5-5:herramienta` |
| FR-019, FR-029, SC-012, SC-014 (modo orden) | 0 respuestas sin la línea `⚠ PLAZO:` con la fecha esperada, de las 12 de las evals de plazo | ídem | `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:orden` |
| FR-019, FR-029, SC-012, SC-014 (modo herramienta) | ídem | ídem | `evals:legal-core:plazo_sin_fecha:claude-sonnet-5-5:herramienta` |
| FR-019, FR-030, SC-012, SC-014 (modo orden) | 0 respuestas sin la línea `⚠ NO CUBIERTO:` con su norma o con una línea `⚠ PLAZO:`, de las 9 de las evals de cobertura | ídem | `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:orden` |
| FR-019, FR-030, SC-012, SC-014 (modo herramienta) | ídem | ídem | `evals:legal-core:cobertura_sin_declarar:claude-sonnet-5-5:herramienta` |
| FR-019, SC-012, SC-014 (modo orden) | 0 respuestas sin la skill activada, de las 30 del modo | ídem | `evals:legal-core:sin_activar:claude-sonnet-5-5:orden` |
| FR-019, SC-012, SC-014 (modo herramienta) | ídem | ídem | `evals:legal-core:sin_activar:claude-sonnet-5-5:herramienta` |
| SC-014 (modo orden) | 0 respuestas sin la línea `⚠ PLAZO SIN CALCULAR:` o con una línea `⚠ PLAZO:`, de las 3 de la de vigencia | ídem | `evals:legal-core:vigencia_sin_declarar:claude-sonnet-5-5:orden` |
| SC-014 (modo herramienta) | ídem | ídem | `evals:legal-core:vigencia_sin_declarar:claude-sonnet-5-5:herramienta` |
| SC-014 (modo orden) | 0 respuestas sin la lectura fechada o sin sus líneas, de las 9 de las evals de redacción fechada | ídem | `evals:legal-core:redaccion_sin_fechar:claude-sonnet-5-5:orden` |
| SC-014 (modo herramienta) | ídem | ídem | `evals:legal-core:redaccion_sin_fechar:claude-sonnet-5-5:herramienta` |
| SC-019 | `afirma_lo_no_leido:claude-sonnet-5-5` de `boe-legislacion` con su tasa de `clases.yaml` (ADR 0038), sin cambiar | umbral conjunto con `decide: true` | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5` |
| SC-002 (`boe-legislacion`: los que ya decidían) | sus otros 10 que deciden, cumplidos | umbrales del informe con `decide: true`; los de hoy, sin cambiar | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`, `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`, `evals:boe-legislacion:redaccion_sin_fechar:claude-sonnet-5-5:orden`, `evals:boe-legislacion:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`, `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`, `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados`, `evals:boe-legislacion:duracion_de_las_sesiones:orden`, `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`, `evals:boe-legislacion:duracion_del_juez:orden`, `evals:boe-legislacion:duracion_del_juez:herramienta` |
| SC-020, SC-002 (`jurisprudencia` sigue aprobado) | sus 11 que deciden, cumplidos | ídem | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`, `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`, `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`, `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`, `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`, `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`, `evals:jurisprudencia:duracion_del_juez:orden`, `evals:jurisprudencia:duracion_del_juez:herramienta` |
| SC-016 | 0 peticiones en `red` del informe de `legal-core` | una llegada a la red es un motivo del informe y su veredicto, `fallo` (desde H7.3, sin cambiar) | `ci:internal/evals/informe_test.go:TestInforme` |
| SC-011, FR-001, FR-003, SC-003 | 0 sesiones con un bloque de `data/` en `fuera_de_lo_grabado`: cada bloque de los dos ficheros, sin fecha y con `9999-12-31`, y el índice y los metadatos de cada norma de los tres, con 0 desde la caché de una sesión de cada skill; un precepto sin grabar, 1 falta y la sesión sin abrir | la caché preparada con el `data/` del repositorio y con una copia rota | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, `ci:internal/evals/preparar_test.go:TestLoQueLlevaALeerSinGrabar` |
| FR-002, FR-005, SC-008 (la derivación) | 0 faltas con el `data/` y las grabaciones de hoy; sin la derivación, el mutante en rojo | ídem, y el mutante de `mutantes.md` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones` |
| FR-004 | 0 faltas al leer desde la base de la medida una lectura de `data/` que no nombra ninguna eval; 0 invocaciones de `boe` de la sesión de un caso etiquetado que terminaron con 4 o 5 en su informe y la base sirve | la base de `boe-legislacion` y la de `jurisprudencia`, con sus casos | `ci:internal/evals/medida_test.go:TestBaseDeLaReconstruccion` |
| SC-018, FR-006, FR-010, FR-011, SC-004 | 0 sesiones de `boe-legislacion` con un bloque de `lecturas` en `fuera_de_lo_grabado`: los 4 bloques, sin fecha y con `9999-12-31`, con 0 desde la caché de otra eval de la skill; grabados | la caché preparada; la preparación de las consultas de `boe-legislacion` | `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones`, `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-013, SC-005, SC-008 (`sin_servir`) | desde la 11, `a124` con 4 en 3 de 3 fechas, y desde la 05, con 0 en 3 de 3; ídem `daprimera` en la 12 y la 07; con `sin_servir` aplicado a todas, el test en rojo | la caché preparada de cada sesión | `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones` |
| FR-008, FR-009, FR-012, FR-014, SC-004, SC-005 | 0 evals mal formadas en el repositorio; cada uno de los rechazos de contracts/formato-de-eval.md §2 y del esquema, con su error | la lectura del conjunto y de cada eval | `ci:internal/evals/formato_test.go:TestFormatoDeLecturas`, `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-007, FR-015 a FR-018, SC-005 | 0 diferencias con lo esperado en cada sesión sintética de contracts/juicio-e-informe.md §1 y §2 | el juicio | `ci:internal/evals/juzgar_test.go:TestJuzgarLasLecturasSinServir` |
| FR-033, FR-015, SC-008, SC-013 (el control en `make ci`) | con una sesión de la 11 sin intentar la lectura, medida 1 y `fallo`; sin el umbral o sin dejarlo decidir, el test en rojo | sesiones sintéticas e informe | `ci:internal/evals/umbrales_test.go:TestUmbralesDeLasLecturas` |
| FR-021, FR-022, SC-006, SC-015 | medida 2, total 3 y `decide: false` en el caso de US4; 0 lecturas de `sin_servir` contadas; con el plan del job, 4 por skill, todos con `decide: false` | sesiones sintéticas e informe con las evals del repositorio | `ci:internal/evals/umbrales_test.go:TestUmbralesDeLasLecturas`, `ci:internal/evals/informe_test.go:TestUmbralesDeLasTresSkills` |
| FR-023, FR-024, FR-026 a FR-028, SC-007, SC-017 | 0 llamadas sin sobre con código 1; 0 órdenes enlazadas sin su motivo; 0 diferencias en `pasa` y `motivos` | sesiones sintéticas | `ci:internal/evals/juzgar_test.go:TestJuzgarElMotivoDeLasFallidas`, `ci:internal/evals/sesion_test.go:TestLeerLasLlamadas` |
| FR-025, SC-007 | la línea que deja el kernel, reconocida; con el mensaje cambiado en el binario, al menos 1 test en rojo | la salida de error de `app.Main` con una salida estándar que falla; el binario real con una tubería sin lector | `ci:internal/evals/juzgar_test.go:TestLineaDeLaSalidaNoEscrita`, `ci:internal/app/tuberia_unix_test.go:TestTuberiaCerrada` |
| FR-020, SC-009 | el tope de 352 minutos ≥ el peor caso de `legal-core` con las evals del repositorio, 19 525 s | la definición del job contra las evals del repositorio | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |
| FR-031, SC-021 | `SKILL.md` de `legal-core` con menos de 300 líneas; tabla y referencias iguales a las regeneradas | la comprobación de skills | `ci:Makefile:skills-check`, `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |

SC-001 es el propio `make ci`; SC-002 lo comprueba `precheck.sh plan`; SC-010 y FR-032, la revisión final y el job en la
propuesta de cambio; SC-017, en su segunda parte —confirmar o descartar la causa del ADR 0043—, lo dice el informe final
con lo que publique el job: ninguno de los cuatro es un umbral numérico.

## Uso, de fuera adentro (criterio de uso, ADR 0028)

Los contratos dan la forma, un ejemplo con sus bytes y cuándo se apaga cada señal (contracts/cache-de-la-sesion.md §4 y
contracts/juicio-e-informe.md §7). Resumen, con el cálculo rehecho:

- **La caché de la sesión** la usa la skill dentro de cada sesión: `legal-core` lee `daprimera` una vez por pregunta de
  una materia regulada aparte (paso 7) y de uno a tres preceptos de `data.preceptos` en la fecha de inicio (paso 10);
  `boe-legislacion`, el índice de la norma de la pregunta y de uno a tres bloques. La prepara el arnés una vez por
  sesión: 72 consultas de `data/` (V1), 8 de `lecturas` en `boe-legislacion`, menos 2 en la 11 y la 12. Crece con
  `data/`, no con el uso: cada sesión empieza con una caché nueva y vacía, y los meses de uso de quien tiene kitlegal
  instalado —cientos de normas y miles de bloques en su `~/.cache/kitlegal/`— no están en ninguna. Se apaga cuando el
  precepto sale de `data/`, la lectura de `lecturas` o la eval deja `sin_servir`.
- **El informe del job** lo lee la persona una vez por job y skill. Lo que el hito añade está acotado por las sesiones
  del plan, no por lo acumulado: con los informes del cierre de H20 (medidos en esta sesión: `boe-legislacion`, 919 KB
  con 74 series y 222 sesiones; `legal-core`, 364 KB con 34 y 102; `jurisprudencia`, 339 KB con 40 y 120), en
  `boe-legislacion` son 222 × 25 bytes de `"lecturas_sin_servir":[]`, 74 × 39 de `"sesiones_con_lectura_sin_servir":null`
  y 4 umbrales de unos 350 bytes: unos 10 KB, el 1,1 %; y en `legal-core`, con 46 series, 139 sesiones y 8 umbrales,
  unos 14 KB,
  con las 24 sesiones de la 11 y la 12 y de uno a tres elementos de 85 bytes en cada una. Los `motivo` y `bash`, unos
  diez por job en `boe-legislacion`, de 75 a 240 bytes. Cada job lo calcula desde cero y ninguna señal pasa al siguiente.
- **`lectura_sin_intentar`** va solo en el informe del job de la skill con una eval con un comando de bloque en su
  `sin_servir` —hoy, 4 umbrales en el de `legal-core` y ninguno en los otros dos— y sale de él cuando la skill deja de
  tener una eval así. Su motivo, si no se cumple, nombra como mucho las sesiones de esas evals de ese modo (hoy, las 3
  de la 11): menos de 300 bytes, en el veredicto del job y en el informe final.

## Decisiones

Las de research.md, con la alternativa rechazada de cada una: D1 (dónde se deriva `data/` y cómo se suma), D2
(`lecturas` como cuarto punto), D3 (`sin_servir` al preparar), D4 (la medida sin `sin_servir`), D5 (FR-016 en
`consultoComoPide`), D6 (el orden del reparto), D7 (qué es «sin sobre»), D8 (la línea, con el error del presentador),
D9 (esquema y `LeerEval`), D10 (`OrdenesDeBash`), D11 (las sesiones medidas de los umbrales nuevos), D12 (su sitio en el
informe), D13 (el tope no cambia), D14 (la skill en 299 líneas), D15 (sin aceptación e2e), D16 (los mutantes en una
copia) y D17 (la fila de la fuente no cambia). Ninguna desviación de la constitución salvo la de Complexity Tracking.

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Dónde | Requisito |
|---|---|---|
| `LecturaDeLaEval`, `Eval.Lecturas`, `lecturas` en el esquema | `formato.go`, `schemas/eval.yaml.json` | FR-006, FR-009 |
| `LecturaSinServir`, `Eval.SinServir`, `sin_servir` en el esquema | ídem | FR-012 |
| cruces de `LeerEval` (lectura ya comando o cita; `sin_servir` ya cita) | `formato.go` | FR-008, FR-014 |
| `LecturaSinServir.consulta`, `Eval.sinServir`, `Eval.comandoSinServir` | `formato.go` | FR-013, FR-015, FR-016 |
| `PuntoLecturaDeLaEval` y su anotación en `ConsultasNecesarias` | `consultas.go` | FR-006 |
| `PuntoLoQueLlevaALeer`, `anotarLosDatos`, `datosDelRepositorio` | `consultas.go` | FR-001, FR-002 |
| `consultasDeLaCache` | `consultas.go` | FR-001, FR-004 |
| `sinLoQueNoSeSirve` y su uso en `prepararLaCache` | `consultas.go`, `preparar.go` | FR-013 |
| `queFaltaDesde` para los dos puntos nuevos | `preparar.go` | FR-003 |
| `leerLasEvalsDeHoy` con `consultasDeLaCache` de las evals sin sus lecturas (`sinSusLecturas`) | `medida.go` | FR-004 |
| `Llamada.SinSobre` | `sesion.go` | FR-023 |
| `Sesion.OrdenesDeBash` | `sesion.go` | FR-026 |
| reparto de `informarLaJuzgada` (filas 2, 3 y 6) | `juzgar.go` | FR-015, FR-023, FR-024 |
| `InvocacionInformada.Motivo`; `InvocacionFallida.Codigo` `*int`, `.Motivo`, `.Bash`; `ResultadoDeEval.LecturasSinServir` | `juzgar.go` | FR-015, FR-023, FR-024 |
| `satisface` y `consultoComoPide` con la lectura fallida admitida | `juzgar.go` | FR-016 |
| prefijo de la salida no escrita, pedido una vez a `render.Nuevo` | `juzgar.go` | FR-025 |
| enlace de una invocación con su uso de Bash | `juzgar.go` | FR-026, FR-027 |
| `TasaDelInforme.SesionesConLecturaSinServir` | `informe.go` | FR-015 |
| «sin sobre» en la tabla de invocaciones de `informe.md` | `informe.go` | FR-023 |
| `fuera_de_lo_grabado:<modelo>:<modo>` | `umbrales.go` | FR-021, FR-022 |
| `lectura_sin_intentar:<modelo>:<modo>` y su motivo | `umbrales.go` | FR-033 |
| el contrato del paquete en su comentario | `doc.go` | FR-021, FR-033 |
| las tres entradas del manifiesto y las cuatro grabaciones | `testdata/grabaciones.json` | FR-010 |
| `lecturas` de la 14, la 15 y la 16 | `evals/boe-legislacion/` | FR-011 |
| la 11 y la 12 | `evals/legal-core/` | FR-017, FR-018, FR-019 |
| los literales de `TestDefinicionDelJob` | `definicion_test.go` | FR-020 |
| los pasos 7 y 10 de la v1.2.3 | `skills/legal-core/SKILL.md` | FR-029 a FR-031 |
| `CHANGELOG.md`, `CONTRIBUTING.md` | raíz | FR-032 |
| `mutantes.md` | el directorio del hito | SC-007, SC-008 (FR-002, FR-013, FR-025, FR-033) |

Lo que no se construye, por no trazarse a ningún requisito o no pasar el umbral de materialidad: `sin_servir` en la
medida del juez (D4); un umbral que decida sobre `fuera_de_lo_grabado` o sobre los motivos (spec, «Fuera de alcance»);
una constante exportada en `internal/render` (D8); un juicio de la línea `⚠ REDACCIÓN A FECHA:` en la 11 (spec,
Clarifications); `lecturas` o `sin_servir` de otro applet que `boe` o de búsquedas; y cambiar el comentario de los peores
casos de `.github/workflows/evals.yml` (D13).

## Datos externos

| Fuente | Fila en `docs/SOURCES.md` de `main` | Manifiesto | Test de grabación | Respuestas | Material de origen |
|---|---|---|---|---|---|
| `boe.legislacion-consolidada` | revisada el 2026-09-13 (línea 42) | `internal/evals/testdata/grabaciones.json`, con `"fuente": "boe.legislacion-consolidada"` y tres entradas nuevas (contracts/formato-de-eval.md §4) | `TestGrabarEvals`, `//go:build grabacion`, `internal/evals/grabacion_test.go`, por `internal/httpx` con `KITLEGAL_RECORD=1` | cuatro ficheros nuevos en `internal/evals/testdata/boe.legislacion-consolidada/` | ninguno: ningún fichero de `data/` sale de ellas, y nada va a `$KITLEGAL_EVIDENCIAS` (`evidencias/h26-1/`) |

Lo graba el paso `grabar_datos` del workflow tras el commit de la tarea `[datos]` del manifiesto (research V11); las
tareas siguientes ya parten de lo grabado. Ninguna tarea toca la red ni escribe en `evidencias/`. Ninguna otra lectura
de la clasificación del ADR 0043 se graba (FR-010).

## Orden de implementación (de dentro afuera)

Ninguna tarea `[aceptacion]` (Aceptación e2e: no aplica). Cada paso deja `make ci` en verde.

1. **`[datos]` el esquema**: `schemas/eval.yaml.json` con `lecturas`, `sin_servir` y sus `$defs`. Ninguna eval los lleva
   todavía.
2. **`[datos]` el manifiesto**: las tres entradas de `internal/evals/testdata/grabaciones.json`. Tras el commit,
   `grabar_datos` graba los cuatro bloques.
3. **El formato** (`formato.go`): los dos tipos, los dos campos, los cruces de `LeerEval` y los métodos de §1 de
   data-model; `TestFormatoDeLecturas`.
4. **La caché** (`consultas.go`, `preparar.go`, `medida.go`): el punto de `lecturas`, `anotarLosDatos`,
   `consultasDeLaCache`, `sinLoQueNoSeSirve` y sus usos; `TestConsultasNecesarias`, `TestLoQueLlevaALeerSinGrabar`,
   `TestBaseDeLaReconstruccion`, `TestEvalsDelRepositorio/lo-que-lleva-a-leer`; y la parte `lecturas` de
   `TestCacheDeLasSesiones` con las `lecturas` de la 14, la 15 y la 16 (`evals/boe-legislacion/`), que ya están grabadas.
5. **La sesión** (`sesion.go`): `Llamada.SinSobre` y `Sesion.OrdenesDeBash`; sus casos en `TestLeerLasLlamadas` y
   `TestTextosDeLaSesion`.
6. **El juicio** (`juzgar.go`, `informe.go` en lo que publica de cada sesión): el reparto, FR-016, el prefijo, el enlace y
   la fila de `informe.md`; `TestJuzgarLasLecturasSinServir`, `TestJuzgarElMotivoDeLasFallidas`,
   `TestLineaDeLaSalidaNoEscrita`.
7. **El informe** (`informe.go`, `umbrales.go`, `doc.go`): la serie y los dos umbrales; `TestUmbralesDeLasLecturas` y los
   tests de listas enteras.
8. **Las dos evals de `legal-core`**: la 11 y la 12, con la parte `sin-servir` de `TestCacheDeLasSesiones` y los literales
   que cambian (`conjunto_test.go`, `definicion_test.go`, `informe_test.go`, `plazos_test.go`).
9. **La skill**: `skills/legal-core/SKILL.md` v1.2.3.
10. **La documentación**: `CHANGELOG.md` (*Unreleased*), `CONTRIBUTING.md` (la tabla del formato, con `lecturas` y
    `sin_servir`, y en «Job de evals» qué sirve la caché de una sesión, los dos umbrales nuevos y los campos nuevos).
11. **Los mutantes**: cada uno en un worktree de `$TMPDIR`, apuntado en `mutantes.md` con su orden y el test que falla
    (research D16).

El cierre —push, propuesta de cambio, CI y evals sobre la cabeza— lo hace el workflow tras la revisión final; ninguna
tarea lo hace.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|---|---|---|
| «Aceptación e2e: no aplica» (principio III: «cada hito empieza por el test e2e») | El hito no cambia el binario (research V5, D8, D15). La aceptación es la del job de cierre (SC-011 a SC-020), y cada pieza entra con sus tests en `make ci`. Precedentes: H24, H25, H26 y H23.1 | Un guion que ejerciera el binario sin cambios no describiría la entrega |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- precheck: `plan.md` y `research.md` existen; el plan no conserva marcas sin resolver; tiene `## Constitution Check`,
  la línea «Aceptación e2e» y `## Controles de umbral`, con cada fila nombrando requisitos del spec y un control con la
  forma `ci:<ruta>:<Test>`, `ci:Makefile:<objetivo>` o `evals:<skill>:<nombre>`.
- a · Constitution Check: un ítem por principio (I a IX) y por regla (R1 a R7).
- b · Dependencias: ninguna nueva.
- c · Reglas de dependencia: `internal/evals` importa `internal/render`, que solo prohíben a `core` y a `graph`; ningún
  import de `net/http`, SQLite ni MCP; ningún `os.Exit` ni escritura en stdout.
- d · Errores y códigos: ninguno del binario cambia; el error de preparar con un `data/` roto detiene la sesión como
  cualquier otro.
- e · Tests primero: «no aplica» con el motivo; los tests nuevos, los casos nuevos, los que cambian, los fixtures y los
  controles de arriba.
- f · Alcance: cada mecanismo con su requisito (Trazabilidad) y lo que no se construye, con su porqué.
- g · Sin atajos: ningún `//nolint`, `t.Skip` ni `TODO`; ningún error descartado.
- h · Mejor alternativa: cada decisión de research.md, con la rechazada.
- i · Afirmaciones verificadas: research V1 a V21, con el fichero y la línea o la ejecución vista en esta sesión; S1 a
  S3, sin verificar y dichas así.
- j · Quickstart: órdenes, rutas y tests de este plan; lo que deja cada escenario, declarado.
- k · Datos externos: «Datos externos».
- l · Autonomía: ninguna persona a mitad del run; ninguna tarea publica ni mide en la plataforma.
- m · Uso: «Uso, de fuera adentro», con el cálculo y las medidas de esta sesión.
- n · Proporcionalidad: «Trazabilidad» y «Lo que no se construye».
- o · Controles de umbral: una fila por cada umbral de las evals 11 y 12, por los que ya decidían en las tres skills y por
  cada umbral de `make ci`.
