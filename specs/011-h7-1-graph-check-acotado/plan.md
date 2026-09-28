# Implementation Plan: H7.1 · `graph check` acotado a la pregunta, con señales que se apagan y salida legible; H7 sin lo que no pasa el umbral (ADR 0028)

**Branch**: `011-h7-1-graph-check-acotado` | **Date**: 2026-09-28 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/011-h7-1-graph-check-acotado/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de la constitución y
están en [research.md](./research.md) (D1-D24) con su alternativa rechazada y su motivo. Toda afirmación sobre SQLite,
Kong, `regexp`, el generador de la tabla, el kernel o el workflow remite a la tabla V de research.md, comprobada en
local y sin red; lo que no se puede comprobar así son los supuestos S1-S3 y no se afirma como hecho.

## Summary

H7 dio a `boe-legislacion` una memoria que no escala ni se calla. H7.1 la arregla sin skill nueva, en siete piezas:
(1) `kitlegal graph check [<norma> [<bloques>...]]`, acotado a una norma y a bloques suyos, con 50 hallazgos como
mucho en cualquier forma y una `data` plana que dice el total de cada clase, los omitidos y el ámbito; (2)
`version-obsoleta` decidido por **lecturas**: el grafo guarda por bloque qué redacción vio la última lectura y cuál la
anterior, y el hallazgo se da cuando la última es posterior y se apaga con la lectura siguiente; (3) `fuente-caducada`
solo sobre la `Norma`, el `Bloque` y la redacción vista; (4) `boe-legislacion` v0.1.1, una comprobación por norma
citada, después de leer, con la forma fija `⚠ REDACCIÓN MODIFICADA:`; (5) la eval 19 la exige sin modelo; (6) salida
legible de `stats`, `show` y `check`; y (7) H7 sin los mecanismos y tests de estados a los que el binario no llega: un
`world.db` que no se puede usar es la regla genérica (1, `inesperado`, la ruta), `world.db` se crea en su sitio.

Decisiones que sostienen el diseño:

1. **Dos ranuras por bloque, no historia** (D1-D3): tabla `lecturas(bloque, ultima, anterior)` en la migración 2; una
   lectura es cada arista `eli:has_version` de un lote que se confirma (solo la emiten `boe articulo`/`articulos`,
   V20); un bloque sin fila —un `world.db` de H7— cuenta con una lectura de la redacción observada la última, regla
   única en Go que se usa al comprobar y al entregar la primera lectura. Un `world.db` de versión 1 se lee sin migrar
   (D4).
2. **La comprobación cuesta lo que la pregunta** (D8-D10): con norma, la instantánea se lee acotada en SQL (la `Norma`
   por su identificador, sus bloques y redacciones por clave primaria); la cota corta la lista y no los totales.
3. **Argumentos como los de `boe`** (D6-D7): `check` recibe la norma y los bloques en la misma posición que
   `boe articulos`, se validan con la gramática de `boe` y la tabla de comandos los escribe como Kong
   (`[<norma> [<bloques>...]]`).
4. **La etiqueta la da el binario** (D12-D14): `grafo.EtiquetasDeHallazgo()`; `internal/evals` compone con ella la
   forma fija con el mismo constructor que los avisos, la juzga en la eval 19, la exige en `SKILL.md` desde
   `make ci` y declara en el informe la forma que exige cada eval.
5. **Retirar, no reescribir** (D17-D22): la lectura mira solo si hay `-wal`; la escritura crea en su sitio con los
   permisos de H7; se retiran la publicación, los permisos, el inmutable, los enlaces, los diarios, el `-shm` suelto,
   las bases de fuera, los rechazos y niveles de desempate que nadie produce, las filas dañadas y los casos no ASCII
   de `Persona`, con sus tests; salen también los tests de las entradas del lote y de lo guardado que ningún emisor
   produce, cuyo código se queda como regla genérica (D20, D21); lo que tiene vía real se queda con los suyos.

No cambia ninguna decisión de arquitectura: no hay ADR nuevo y `specs/010-h7-internal-graph-grafo/` no se edita
(FR-083).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.1`), `CGO_ENABLED=0`, `-trimpath`.

**Primary Dependencies**: las fijadas, sin ninguna nueva: `alecthomas/kong` v1.16.1 (argumentos de posición
opcionales, V11-V12), `modernc.org/sqlite` v1.59.0 (SQLite 3.53.4 con `json_extract` y `json_each`, V1, V7),
`stretchr/testify`, `rogpeppe/go-internal/testscript`, `santhosh-tekuri/jsonschema` e `invopop/jsonschema` (esquemas).

**Storage**: `world.db` (SQLite, WAL) junto a la caché; esquema versión 2 con la tabla `lecturas`
(data-model §1; contracts/almacen-world-db.md §2).

**Testing**: `go test -race` (unitarios, tabla y `t.Parallel`), `//go:build integration` con `t.TempDir()`, e2e con
`testscript` contra los binarios de reloj fijo del arnés (T0, T1, T8), evals sin modelo en `make skills-check`.

**Target Platform**: los del binario distribuido (Linux, macOS, Windows; goreleaser), sin cambios.

**Project Type**: CLI multicall (applet `graph`) + skill (`boe-legislacion`) + formato común de eval.

**Performance Goals**: `graph check` con norma lee la norma y su ámbito, no el grafo (research D8); sin argumentos,
≤ 40 000 bytes con `--json` sobre la medida (SC-001); la entrega añade por bloque leído una consulta por clave y un
`INSERT … ON CONFLICT`. Las cotas de H7 SC 007 (150 ms) y SC 008 (3 s) siguen (S3).

**Constraints**: sin red en ningún test ni tarea; ningún verbo de `graph` escribe; la salida con `--json` de `stats` y
`show` no cambia ni un byte; `SKILL.md` < 300 líneas.

**Scale/Scope**: la medida de la bitácora, 300 normas y 2 400 bloques, 240 con dos lecturas, el 90 % hace más de una
semana (contracts/applet-graph.md §6).

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones.*

### Principios

| Principio | Cómo lo cumple H7.1 |
|---|---|
| I · Fuentes públicas y frontera humana | No toca ninguna fuente ni `internal/httpx`; ningún test ni tarea usa la red; no hay grabación nueva (la redacción C deriva de la de H4, research D23). |
| II · Nada sin cita ni fuente | El sobre de `graph` no cambia (`kitlegal.graph`, `kitlegal:applet/graph`, instante). Ningún dato entra sin `source`: las filas de `lecturas` solo nacen de aristas `eli:has_version` de un lote validado con fuente, url y fecha (FR-075), solo nombran nodos que ya guardan su procedencia (claves ajenas) y no llevan contenido propio. `graph` sigue sin devolver texto legal (H7 FR 070; SC-010). |
| III · Tests primero y offline | La primera tarea escribe la suite congelada (cuatro guiones, «Aceptación e2e» abajo), en rojo por una aserción (contracts/arnes-e2e.md §2). Unitarios e integración contra `t.TempDir()`; la salida de `check` se valida contra `schemas/grafo.json` regenerado; cobertura dentro de `codecov.yml` (FR-096). |
| IV · Hexagonal y errores tipados | Reglas, ámbito, `Comprobacion`, lecturas y etiquetas en `internal/core/grafo` (puro); SQL en `internal/graph`; composición y salida legible en `internal/app`. Los argumentos inválidos envuelven `cli.ErrArgumentos` (2); todo fallo del almacén es un `*graph.Error` con clase (1 o 4), y lo que no puede usar es la regla genérica (1, `inesperado`), nunca un `panic` (research D21). |
| V · Simplicidad y dependencias | Ninguna dependencia nueva. Cada mecanismo se traza a un FR («Trazabilidad» abajo); lo que solo servía a estados sin vía real se retira (FR-070 a FR-076). |
| VI · Un binario, convenciones de agente | Mismas ocho banderas globales; `--describe` de `check` gana `norma` y `bloques` y la `data` nueva; la tabla de comandos se regenera desde `--describe` (FR-007). |
| VII · Grafo y privacidad | `world.db` sigue local; `lecturas` solo guarda ids del mundo (BOE); la regla de `Persona` con NIF/NIE/DNI se queda en `Apply`, con sus casos ASCII (FR-074). |
| VIII · Skills primero | Mejora `boe-legislacion` (v0.1.1) y su eval 19; `SKILL.md` < 300 líneas, tabla sin drift, sin nombrar evals ni modelos (FR-040 a FR-048). |
| IX · Genericidad territorial | Nada se particulariza para un municipio. `graph check` no tiene dimensión territorial: los nodos de `territorio` no declaran vigencia (V21) y el ámbito es una norma del BOE; `h7-grafo-matriz-territorial.txtar` no cambia. |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | `internal/core/grafo` solo gana biblioteca estándar (`cmp`, `slices`, `strings`, `time`, `fmt`); nada de `os`, `io`, `database/sql` (depguard, `internal/arch_test.go`). |
| R2 · Solo `internal/httpx` importa `net/http` | Sin cambios. |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | La tabla, las consultas acotadas y el upsert viven en `internal/graph`; los tests que tocan SQL (p. ej. `TestIntegracionGrafoDeH7`, que devuelve la base a la versión 1) están en `internal/graph`; `TestMedidaDelGrafo` siembra con `graph.Nuevo(…).Apply`. |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Sin cambios. |
| R5 · Solo `render` escribe en stdout | La salida legible es una cadena en `Resultado.Legible`; la escribe el kernel por el presentador (V17). |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios; la validación de la norma con `boe.ValidarNorma` está en `internal/app` (raíz de composición, que ya importa `internal/source/boe`). |

### Gates (constitución, «Gates»)

- **Capa 1**, todo en `make ci` o en el workflow: suite congelada con «rojo primero»; `schema-check` con
  `schemas/grafo.json` regenerado; `skills-check` con la tabla, las 300 líneas, el esquema de eval y los subtests
  `hallazgos-de-la-skill` y `hallazgos-del-esquema`; guardián de diff con las rutas declaradas; ningún `t.Skip` nuevo;
  corrección de la cita y de la forma por identificador y expresión, no por modelo.
- **Capa 2**: jueces del plan, de las tareas y de la revisión final; claridad de las frases de la salida legible de
  `check`.
- **Capa 3** (informe final): la derivada nueva `internal/app/testdata/derivadas/version-ulterior/`, los cambios en
  `schemas/grafo.json` y `schemas/eval.yaml.json` y en los guiones de H7 bajo `testdata/`; los supuestos del run.

## Project Structure

### Documentation (this feature)

```text
specs/011-h7-1-graph-check-acotado/
├── plan.md                 # este fichero
├── research.md             # D1-D24, verificaciones V1-V29, supuestos S1-S3
├── data-model.md           # lecturas, ámbito, Comprobacion, etiqueta, formato de eval, retirada
├── quickstart.md           # escenarios de validación
├── contracts/
│   ├── applet-graph.md     # check: argumentos, códigos, data, reglas, salida legible, uso
│   ├── almacen-world-db.md # esquema 2, lectura y escritura tras la retirada, errores, qué lo vigila
│   ├── evals-y-skill.md    # formato de eval, juicio, comprobaciones, SKILL.md, eval 19, informe, CHANGELOG
│   └── arnes-e2e.md        # suite congelada, tests de medida, guiones de H7 que cambian
├── aceptacion/             # la escribe la tarea [aceptacion] (T001)
├── checklists/             # del spec
└── tasks.md                # la escribe /speckit-tasks
```

### Source Code (repository root)

```text
internal/core/grafo/
├── vocabulario.go      # + EtiquetasDeHallazgo, MaximoDeHallazgos
├── lecturas.go         # nuevo: Lectura, LecturasDeBloque, Consolidado.Lecturas, RedaccionVistaSinLecturas
├── comprobar.go        # Comprobar(instantanea, ambito, ahora) → Comprobacion; reglas nuevas; fuera versionesObsoletas/compararRecencia
├── salida.go           # + Ambito, Comprobacion (MarshalJSON sin null); Instantanea.Lecturas
├── lote.go             # fuera ValidarContraGrafoVacio, URI absoluto, id/tipo, validarArista; Consolidar: un id repetido, una vez, sin comparar datos
├── errores.go          # nombrar sin el caso de la arista (ningún rechazo que se queda la lleva)
├── observacion.go      # desempate solo por url
├── persona.go          # solo el comentario de los casos no ASCII
└── *_test.go           # comprobar_test (secuencia, ámbito, cota, orden), lecturas_test, lote_test, errores_test, observacion_test, persona_test, canonico_test, salida_test
internal/graph/
├── migraciones/0002_lecturas.sql   # nuevo
├── lectura.go          # Instantanea(ctx, ambito): acotada en SQL; filas de lecturas; versión 1 legible
├── abrir.go            # solo -wal decide el modo; fuera inmutable, permisos, enlaces, diarios, -shm, 776/1544/14
├── almacen.go          # sin enlazar; Apply crea en su sitio
├── aplicar.go          # MkdirAll + OpenFile 0600; lecturas previas y upsert; fuera extremos, permisos, errFilaDanada
├── errores.go          # fuera cuatro constructores; errorInutilizable sin «no se modifica»
├── doc.go              # lo que deja de ser cierto
├── publicar.go, publicar_test.go, integracion_enlace_test.go   # se retiran
└── *_test.go           # almacen, abrir, aplicar, errores, lectura, migraciones, integracion (retirada y adaptación, research D17-D22)
internal/app/
├── grafo.go            # check: Norma y Bloques de posición, validación, Comprobacion, descripción; Legible de los tres verbos
├── grafo_legible.go    # nuevo: legibleDeStats, legibleDeShow, legibleDeCheck
├── grafo_legible_test.go, grafo_test.go, coste_test.go, medida_test.go (nuevo, integration)
└── testdata/
    ├── derivadas/version-ulterior/GET_…_texto_bloque_a21.json   # nueva [datos]
    └── script/h7-grafo-*.txtar                                   # los ocho de FR-080 [datos]
internal/skills/comandos.go (+ _test)   # opcionales de posición como Kong
internal/app/skills_test.go             # sintaxis de la tabla
internal/evals/
├── avisos.go           # formaFija extraída
├── hallazgos.go        # nuevo: ExtraerHallazgos, ComprobarFormasDeHallazgo, ComprobarClasesDeHallazgo
├── formato.go, juzgar.go, informe.go (+ tests), conjunto_test.go, preparar_test.go
schemas/grafo.json, schemas/eval.yaml.json          # [datos]
skills/boe-legislacion/SKILL.md                     # v0.1.1 y tabla regenerada
evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml
CHANGELOG.md, README.md (121-128), CONTRIBUTING.md (335-351)   # lo que el hito deja falso
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquetes nuevos: dominio en
`internal/core/grafo`, almacén en `internal/graph`, composición y salida legible en `internal/app`, formato de eval en
`internal/evals`, generador en `internal/skills`.

## Aceptación e2e

**Aceptación e2e:** cuatro guiones testscript, uno o dos por historia, que la tarea `[aceptacion]` escribe en
`specs/011-h7-1-graph-check-acotado/aceptacion/` y quedan congelados (contracts/arnes-e2e.md §3):
`grafo-lecturas.txtar` (US2, US4 · FR-011, FR-014, FR-020, FR-023-FR-025, FR-030-FR-032 · SC-003, SC-004),
`grafo-check-acotado.txtar` (US3.1-US3.4 · FR-001-FR-007, FR-011, FR-012 · SC-009), `grafo-legible.txtar` (US5 ·
FR-060-FR-064 · SC-010) y `grafo-regla-generica.txtar` (US6.1-US6.2 · FR-070 · SC-011). US1 (la skill) tiene por
aceptación la eval 19 en el job de evals (SC-006) y la prueba de la persona (SC-007); US3.5-US3.6 (la medida), el test
de integración `TestMedidaDelGrafo` (FR-092, SC-001, SC-002); US2.6, `TestIntegracionGrafoDeH7` (FR-026, SC-012); y
US6.3, quickstart §7.

La tarea `[aceptacion]` escribe solo los guiones: la eval 19 cambia con la implementación de FR-050-FR-054, porque con
sus claves nuevas no valida contra el esquema de hoy y `make ci`, que verifica esa tarea (V24), quedaría en rojo
(research D15).

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

Ninguno nuevo ni cambiado. Lo nuevo entra en los que ya hay: `test` (unitarios), `test-integration`
(`TestMedidaDelGrafo`, `TestIntegracionGrafoDeH7`), `test-tiempos` (`TestCosteDelGrafo` adaptado), `schema-check`
(`schemas/grafo.json`), `skills-check` (tabla de `SKILL.md`, esquema de eval y `hallazgos-*`).

### Tests nuevos

| Test | Paquete | Cubre |
|---|---|---|
| `TestLecturas` (una por arista `eli:has_version` del lote: el de `boe articulos` con dos bloques da dos; el bloque nombrado dos veces ya lo fija el test de H7 del emisor, data-model §2) | `internal/core/grafo` | FR-020 |
| `TestComprobar` reescrito: secuencia de FR-025 sobre instantáneas, bloque sin fila, fechas iguales o no válidas, `fuente-caducada` solo sobre lo vigente, orden y cota con 51+ hallazgos, totales y omitidos | `internal/core/grafo` | FR-010-FR-012, FR-023, FR-024, FR-026, FR-030 |
| `TestRedaccionVistaSinLecturas` | `internal/core/grafo` | FR-026 |
| `Apply` con lecturas (fila nueva, desplazamiento, primera lectura de un bloque de H7, sin lectura si el lote se rechaza) e `Instantanea` con y sin ámbito (norma y bloques conocidos y desconocidos, versión 1) | `internal/graph` | FR-002, FR-003, FR-020, FR-021, FR-026 |
| Creación en su sitio: directorios 0700, `world.db` 0600, WAL, esquema 2 (lo que afirmaba `TestPublicar` del resultado) | `internal/graph` | FR-071 |
| `TestIntegracionGrafoDeH7` | `internal/graph` (integration) | FR-026, SC-012 |
| `TestMedidaDelGrafo` (sin argumentos, solo la norma, la norma y un bloque) | `internal/app` (integration) | FR-013, FR-092, SC-001, SC-002 |
| Validación de argumentos de `check` y firma del applet: `a21`, `BOE-B-2015-10565`, `""`, `"" a21` y un bloque `" "` → 2; sin norma → todo | `internal/app` | FR-004, SC-009 |
| `grafo_legible_test.go`: plantillas exactas, singular y plural, sin tabuladores ni escapes, determinismo | `internal/app` | FR-060-FR-063 |
| Sintaxis de opcionales de posición | `internal/skills` | FR-007 |
| `TestExtraerHallazgos`, `TestComprobarFormasDeHallazgo`, `TestComprobarClasesDeHallazgo`, `TestEtiquetasDeHallazgo` | `internal/evals` | FR-045, FR-051, FR-054, SC-008 |
| `Juzgar`: forma encontrada, ausente y dicha con otras palabras; comprobación con otra norma | `internal/evals` | FR-050, FR-052, FR-094 |
| Informe: `formas` y columnas | `internal/evals` | FR-055 |
| Subtests `hallazgos-de-la-skill` y `hallazgos-del-esquema` de `TestEvalsDelRepositorio` | `internal/evals` | FR-093, FR-054 |

### Tests existentes que cambian

Inventario completo, fichero a fichero y caso a caso, en research D17-D22 y contracts/almacen-world-db.md §7. En
resumen:

- **Se retiran** (FR-070-FR-076): `publicar_test.go` e `integracion_enlace_test.go` enteros; `TestPasoAWALDeUnaBaseDeFuera`;
  `TestRutaDeLosAuxiliares`, `TestComprobarEscritura`, `TestLeerReabreInmutable`, `TestLeerSinPermisoDeLectura`,
  `TestApplySinPermisoDeEscritura`, `TestIntegracionSinPermisoDeEscritura`, `TestIntegracionLecturaConDiario`,
  `TestIntegracionLecturaConShmSuelto`, `TestIntegracionCeroBytesConWAL`, `TestApplySobreFilasDanadas`,
  `TestLecturaDeFilasDanadas`; las filas de directorio, base dañada, permisos, enlaces, diarios, `-shm` y bases de
  fuera de `TestDecidirApertura`, `TestLeerEstados`, `TestLeerSinRastro`, `TestLeerConAuxiliares`,
  `TestApplyNoModifica`, `TestApplyEnSuSitio`, `TestIntegracionEsquema`, `TestIntegracionInutilizables`,
  `TestIntegracionRecuperacionDeclarada`, `TestErrores` y `TestMigrar` («`schema_version` ajena» y «versión
  negativa», research D21); los rechazos y los niveles de desempate retirados en `lote_test`, `observacion_test`,
  `almacen_test` e `integracion_test`; los casos de entradas del lote que ningún emisor produce (research D20, V29):
  «una arista antes que sus extremos» en `probarLotesAceptados`, «una operacion nula» y «una operacion que no es un
  valor» en `probarRechazosDeUnaOperacion` (con la frase de su comentario), fecha que no es RFC 3339 y vigencia
  negativa en `probarRechazosDeLaProcedencia`, datos sin forma JSON en `probarLotesQueNoSeConsolidan` y los cuatro «un
  nodo repetido con datos menores/mayores» de `probarLotesConsolidados` (`lote_test`); «un nodo sin id», «una arista»,
  «una arista vacia» y «una operacion que no es un valor» de `TestRechazo` (`errores_test`), que nombran operaciones
  de rechazos que se retiran o que nadie emite; vigencia con fracción de segundo y datos sin forma JSON canónica en
  `TestApplyRechazaElLote`, «un numero que no es finito», «una cadena que no es UTF-8» y «un valor que no es de JSON»
  de `persona_test` y `TestDatosCanonicosImposibles` (`canonico_test`); los de lo guardado que ninguna entrega escribe
  (research D21): `historiasImposibles` y sus casos en `probarNodosQueNoSeFusionan`, `probarAristasQueNoSeFusionan` y
  `probarTextosQueNoSeFusionan` (`observacion_test`), `probarInstantaneasImposibles` entera (`comprobar_test`) y
  `probarFichaSinFormaJSON` (`salida_test`); los casos no ASCII de `persona_test`; en `internal/app/grafo_test.go`,
  `compruebaGrafoIncomprobable` con `baseConUnaFechaIlegible` y los casos del directorio. Las constantes y ayudas de
  esos tests que quedan sin uso salen con ellos.
- **Se adaptan**: el esquema «de una versión posterior» de los tests pasa de la 2 a la 3 y su mensaje a «conoce la 2»
  (`almacen_test.go:669`, `errores_test.go:88`, `integracion_test.go:363,746,1365,2007`, `lectura_test.go:249`,
  `migraciones_test.go:208`, `internal/app/grafo_test.go:1043`); los tests de espera al leer (vehículo en WAL con
  `locking_mode=EXCLUSIVE`, research D22); el helper `esquemaDe` de integración lee con `cadenaDeLectura` en lugar de
  `immutable=1`; en `internal/app/grafo_test.go`, la declaración y la ayuda de `check`, la `data` como objeto
  (`hallazgosDeLaMuestra`, `salidasDelGrafo` —el caso de las dos clases pasa a una siembra con B caducada—,
  `exigirHallazgosPublicados`, la premisa de `TestNingunVerboDelGrafoDevuelveTexto`), el argumento de más de `check`
  (firma del applet), la tabla mínima de `graph` por la salida legible y el directorio por «no es una base»;
  `TestCosteDelGrafo` (research D24); `Instantanea(ctx, grafo.Ambito{})` en `internal/evals/preparar_test.go:533` y
  `conjunto_test.go:1093`; el vehículo de `preparar_test.go` «salida-de-error» pasa de un directorio a un fichero que
  no es una base; `formato_test.go` «comprobacion-con-norma» pasa a aceptarse; `comandos_test.go:762`,
  `skills_test.go:205` e `invocacionDeLaSintaxis`.

### Fixtures, `testdata/` y `schemas/` (tareas `[datos]`)

- `internal/app/testdata/derivadas/version-ulterior/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json`
  (nueva), con su comprobación en `grabacionesDerivadas()` (research D23).
- `schemas/grafo.json` regenerado (`check`: entrada y `data`); `schemas/eval.yaml.json` con `hallazgos` y la norma de
  la comprobación (contracts/evals-y-skill.md §1).
- Los ocho guiones de H7 de contracts/arnes-e2e.md §5.

### CI

Sin cambios en `.github/workflows/`. El job de evals se ejecuta en la propuesta de cambio (lo lanza el workflow tras la
revisión final, FR-098).

## Uso, de fuera adentro (criterio de uso, ADR 0028)

Detalle, con el ejemplo y sus bytes, en contracts/applet-graph.md §3 y §6. Resumen:

- **La skill** pide `graph check <norma> <bloques leídos> --json` **una vez por norma citada**, después de leer (H7
  pedía dos, sin argumentos). Lee 312 bytes cuando no ha cambiado nada, lo habitual, y 3 795 con cinco bloques
  cambiados (197 del sobre + 139 de `data` + 5 × 691 + 4; SC-005: ≤ 3 800, medido con a21-a25 de la LPAC; con ids más
  largos, como `a1-30` de la LCSP, cada hallazgo crece unas decenas de bytes: el tamaño crece con k y con la longitud
  de los ids de la pregunta, nunca con lo acumulado), con cualquier volumen acumulado; con la
  medida, H7 le daba ≈ 6,1 MB por pregunta. No le llega ningún `fuente-caducada` de lo que acaba de leer: la caché no
  sirve una consulta caducada, así que la lectura renueva la observación.
- **Una persona**, sin argumentos: como mucho 50 hallazgos; con la medida, 4 830 contados y 50 listados, 197 + 102 +
  50 × 685 + 49 = 34 598 bytes con los ids de la siembra de TestMedidaDelGrafo (un `version-obsoleta` de `BOE-A-2020-1299`, bloque `a1`, pesa 685; SC-001: ≤ 40 000).
- **Cuándo se apaga cada señal**: `version-obsoleta`, con la lectura siguiente del bloque, la haga quien la haga
  (FR-024); `fuente-caducada`, con la lectura siguiente del bloque, que ya pregunta a la fuente (FR-031), y nunca se da
  sobre una redacción superada (FR-030). La tabla `lecturas` crece con los bloques distintos (2 400 filas, ≈ 600 KB con
  la medida), no con las preguntas.

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| Migración `0002_lecturas.sql` y tabla `lecturas` | FR-020, FR-021, FR-023, FR-024 |
| `Consolidado.Lecturas`, `Lectura` | FR-020 |
| `LecturasDeBloque`, `Instantanea.Lecturas` | FR-023, FR-030 |
| `RedaccionVistaSinLecturas`; lectura de un `world.db` de versión 1 | FR-026, SC-012 |
| `Ambito`, `Instantanea(ctx, ambito)` acotada en SQL | FR-001, FR-002, FR-003, FR-005 |
| Validación de la norma y de los bloques en el applet | FR-004, FR-006 |
| `Comprobacion`, `MaximoDeHallazgos`, orden por clase | FR-010, FR-011, FR-012, FR-013 |
| Reglas nuevas de `version-obsoleta` y `fuente-caducada` | FR-023, FR-024, FR-025, FR-030, FR-031, FR-032 |
| Descripción y argumentos de `check`; opcionales de posición en la tabla | FR-007, FR-040 |
| `grafo_legible.go` | FR-060, FR-061, FR-062, FR-063 |
| `EtiquetasDeHallazgo`, `formaFija`, `ExtraerHallazgos`, `ComprobarFormasDeHallazgo` | FR-045, FR-051, FR-052, FR-093 |
| `hallazgos` y `norma` en el formato de eval; `ComprobarClasesDeHallazgo` | FR-050, FR-054 |
| `formas` en el informe | FR-055 |
| `SKILL.md` v0.1.1 y eval 19 | FR-040-FR-048, FR-050, FR-053 |
| Creación en su sitio (`MkdirAll` 0700, `OpenFile` 0600) | FR-071 (los permisos son los de H7) |
| `errorInutilizable` sin promesa; retirada de constructores | FR-070 |
| Retirada en `abrir.go`, `aplicar.go`, `almacen.go`, `publicar.go` y sus tests; filas dañadas | FR-070, FR-071, FR-072, FR-073, FR-095 |
| Vehículo WAL de los tests de espera al leer | FR-073, FR-077 |
| Retirada en `lote.go`, `errores.go`, `observacion.go`, `persona_test.go` | FR-074, FR-075, FR-076 |
| Retirada de los tests de entradas del lote y de lo guardado que ningún emisor produce (research D20, D21) | Entrega (7) del spec; constitución, «Gates» (regla genérica) |
| Derivada `version-ulterior` | FR-025, FR-091 |
| Guiones de H7 cambiados | FR-080, FR-081 |
| `schemas/grafo.json`, `schemas/eval.yaml.json` | FR-082 |
| `TestMedidaDelGrafo`, `TestIntegracionGrafoDeH7`, `TestCosteDelGrafo` adaptado | FR-092, FR-026, H7 SC 007-008 (se quedan) |
| `CHANGELOG.md`, `README.md`, `CONTRIBUTING.md` | FR-048, FR-097 (y lo que el hito deja falso en la documentación) |

Nada del diseño trata un estado o una entrada por debajo del umbral de materialidad: lo que el binario no produce, en
`world.db` o en un lote (research V21, V27-V29), sigue la regla genérica —1, `inesperado`, la ruta— sin caso propio,
sin test y sin promesa sobre sus bytes; las guardas que hoy lo rechazan se quedan en el código como esa regla, sin caso
en los contratos ni test propio, y sus tests salen, caso a caso, en «Tests existentes que cambian» (research D20,
D21).

## Datos externos

Ninguno (research, «Datos externos»): ni grabación nueva ni fuente nueva; ningún manifiesto `grabaciones.json` ni test
`TestGrabar*`; el paso `grabar_datos` no tiene nada que grabar. La redacción C y el grafo previo de la eval 19 derivan
de la grabación de H4 (BOE, fila revisada en `docs/SOURCES.md`) y los fija `TestGrabacionesDerivadas`.

## Orden de implementación (de dentro afuera)

1. `internal/core/grafo`: retirada de validación, desempate y casos no ASCII (FR-074-FR-076); `Lectura`,
   `Lecturas`, `RedaccionVistaSinLecturas`, `Ambito`, `Comprobacion`, `EtiquetasDeHallazgo`, reglas y cota nuevas en
   `Comprobar` (sin cablear todavía al applet).
2. `internal/graph`: relajar antes, en `[datos]`, las aserciones de los guiones de H7 que el código va a dejar falsas
   (el sufijo «; no se modifica», los `cksum` y auxiliares de «no es una base», las secciones del directorio, el
   temporal de concurrencia y la sección del orden inverso de `h7-grafo-version-obsoleta.txtar`): con el código de
   hoy siguen pasando. Después, la retirada de lectura y escritura (FR-070
   a FR-073) y el vehículo de los tests de espera; luego la migración 2, las lecturas en `Apply` y la instantánea
   acotada.
3. `internal/skills`: opcionales de posición en la tabla (sin efecto todavía en ninguna `SKILL.md`).
4. `internal/app` + contrato publicado: `check` con sus argumentos, su validación, `Comprobacion`, su descripción y la
   salida legible de los tres verbos; con ello, en la misma tarea `[datos]`, `schemas/grafo.json`, la tabla generada
   de `SKILL.md` (`make skills-sync`), la `data` de `check` y la tabla mínima en los guiones de H7 y los tests de Go
   que las fijan (Complexity Tracking).
5. Derivada `version-ulterior` con su comprobación (`[datos]`).
6. `internal/evals`: `schemas/eval.yaml.json` con los casos de `formato_test.go` que lo fijan (`[datos]`); después
   `formaFija`, hallazgos, `Juzgar`, informe, subtests de `TestEvalsDelRepositorio`, eval 19 y el protocolo de
   `SKILL.md` v0.1.1.
7. `TestMedidaDelGrafo`, `TestIntegracionGrafoDeH7`, `TestCosteDelGrafo`.
8. `CHANGELOG.md`, `README.md`, `CONTRIBUTING.md`.

**Obligaciones para `tasks.md`**: T001 `[aceptacion]` solo con los cuatro guiones (research D15); cada tarea deja
`make ci` en verde; las tareas `[datos]` son las de los pasos 2 (relajar guiones), 4 (contrato publicado de `check`),
5 (derivada) y 6 (esquema de eval), cada una con las rutas que le da este plan; ninguna tarea usa la red, graba, ni
publica o mide en la plataforma.

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada |
|---|---|---|
| Una tarea `[datos]` que lleva código: el contrato publicado de `check` (tipo de `data`, argumentos, descripción) junto con `schemas/grafo.json`, la tabla generada de `SKILL.md`, los guiones de H7 que afirman la `data` o la tabla mínima y los tests de Go que los fijan | `make schema-check`, `make skills-check` y los guiones comparan lo que emite el binario con esos ficheros: si el código cambia en una tarea y los ficheros en otra, la primera deja `make ci` en rojo. Es el precedente de T016 de H7 («indivisibles y nada más») | Separarlos, que dejaría una tarea en rojo; o una regex que aceptara las dos formas de `data`, que habría que endurecer después |
| El generador de la tabla de comandos cambia la forma de un argumento opcional (`[--nombre]` → `[<nombre>` anidado) | Sin el cambio, la tabla que lee el agente diría `kitlegal graph check [--norma] [--bloques]`, una orden que no existe (research D7) | Banderas `--norma`/`--bloque` (research D6) o marcar la posición en `--describe`, que cambiaría todos los esquemas publicados |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- `precheck.sh plan`: existen `plan.md` y `research.md`; ni uno ni otro contienen marcas de aclaración pendiente;
  `## Constitution Check` y la línea «Aceptación e2e:» están.
- a · Constitution Check: un ítem por principio (I-IX) y por regla de dependencia (R1-R6), y los gates.
- b · Dependencias: ninguna nueva.
- c · Reglas de dependencia: tabla R1-R6.
- d · Errores y códigos: contracts/applet-graph.md §2 y contracts/almacen-world-db.md §6; la tabla de ADR 0023.
- e · Tests primero: «Aceptación e2e», «Tests nuevos», «Tests existentes que cambian», fixtures.
- f · Alcance: nada fuera del spec; la documentación que se toca es la que el hito deja falsa.
- g · Sin atajos: ningún `//nolint` nuevo, ningún `t.Skip`, ningún error silenciado (los errores de lectura de lo
  guardado se devuelven: regla genérica).
- h · Mejor alternativa: cada decisión con la rechazada (research D1-D24).
- i · Afirmaciones verificadas: V1-V29 con fichero:línea o sonda; S1-S3 como supuestos.
- j · Quickstart ejecutable: binario y caché en un directorio temporal; el único escenario que escribe en el
  repositorio (§9) lo declara.
- k · Datos externos: ninguno.
- l · Autonomía: ninguna tarea para una persona; el cierre en la plataforma lo hace el workflow.
- m · Uso: contracts/applet-graph.md §3 y §6 y «Uso, de fuera adentro» (bytes medidos, invocaciones por pregunta,
  apagado de cada señal).
- n · Proporcionalidad: «Trazabilidad»; lo que no se traza se retira.
