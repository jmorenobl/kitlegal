# Implementation Plan: H22 · Instalar sin terminal: la extensión de escritorio con el servidor y el plugin de Claude con las skills

**Branch**: `016-h22-instalar-sin-terminal` | **Date**: 2026-10-02 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/016-h22-instalar-sin-terminal/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de la constitución y
están en [research.md](./research.md) (D1-D20), cada una con su alternativa rechazada. Toda afirmación sobre goreleaser,
Claude Code, el manifiesto de MCP Bundle, `gh`, Go, los linters o el repositorio remite a la tabla V de research.md,
comprobada en local en la sesión del plan; lo que no se pudo comprobar así son los supuestos S1-S11. El prototipo con el
que se midió es material de un solo uso, fuera de lo versionado (research, cabecera).

## Summary

H22 hace que cada release publique, junto a los archivos de hoy, `kitlegal.mcpb` —la extensión de escritorio, con el
binario dentro— y `kitlegal-plugin.zip` —el plugin de Claude, con las skills—, y que el catálogo
`jmorenobl/kitlegal-plugins` apunte al plugin de cada etiqueta. No cambia el binario, sus herramientas ni las skills. En
seis piezas:

1. **El paso que empaqueta** (research D1 a D5, D8): `cmd/empaquetar`, un `main` mínimo, e `internal/empaquetado`.
   La orden `piezas` escribe los dos zips, reproducibles byte a byte, a partir de los dos binarios que recibe, del icono,
   del registro de applets (`tools`) y de lo empotrado (`skills/`); la orden `catalogo` escribe el `marketplace.json` de
   una versión. Los textos de la ficha viven en un solo fichero. Solo biblioteca estándar.
2. **`tools` sale del registro** (D2): `app.HerramientasAnunciadas`, una función exportada nueva sobre el mismo recorrido
   del que el servidor saca lo que anuncia.
3. **goreleaser** (D6, D7): el universal de macOS con `universal_binaries` —`id` propio, `replace: false`—, cuyo gancho
   `post` ejecuta el paso; `archives` acotado con `ids`, para que los archivos sigan siendo seis; y los dos ficheros en
   `checksum.extra_files` y en `release.extra_files`.
4. **Las comprobaciones del snapshot** (D9 a D11): seis subpruebas nuevas de `TestSnapshot`, en `make snapshot-check`, y
   `make plugin-check`, que ejecuta `claude plugin validate` en el trabajo `snapshot` de la CI.
5. **`release.yml`** (D11, D12): `publicar` atesta los dos ficheros; `humo` los descarga y comprueba, con seis pasos de
   shell, sus huellas, su atestación, el manifiesto, los binarios del `.mcpb` y el servidor; y un trabajo nuevo,
   `catalogo`, que depende de `humo` y escribe el catálogo.
6. **Documentación** (FR-050 a FR-053): «Instalar sin terminal» y «Otras instalaciones, sin probar» en el README,
   CONTRIBUTING y `CHANGELOG.md`.

No cambia ninguna decisión de arquitectura que no haya cambiado ya el ADR 0035: no hay ADR nuevo, y no se edita ningún
spec anterior (FR-080).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.1`), `CGO_ENABLED=0`, `-trimpath`; bash en los pasos de
los flujos (los de `humo`, medidos con bash 3.2); YAML de goreleaser v2.18.1 y de GitHub Actions.

**Primary Dependencies**: ninguna nueva. El paso usa la biblioteca estándar (`archive/zip`, `encoding/json`,
`image/png`, `flag`, `os`); los tests, además, `debug/macho`, `archive/tar`, `compress/gzip`, `testify` y el cliente de
`internal/mcp/mcptest`, que ya existen. goreleaser es el de `tools/goreleaser/go.mod`, sin subirlo.

**Storage**: ninguno. Ficheros nuevos del repositorio: código, tests y documentación; nada en `testdata/`, `schemas/` ni
`data/`. Lo que el paso escribe va a `dist/`, que no se versiona.

**Testing**: `go test -race` en `make ci`, con binarios de prueba en `t.TempDir()`; `TestSnapshot` (etiqueta
`snapshot`) sobre el `dist/` de `make release`; `TestPluginValido`, solo en el trabajo `snapshot` de la CI. Sin guiones
`testscript` nuevos («Aceptación e2e»).

**Target Platform**: el binario, las seis plataformas de hoy, sin cambios. El `.mcpb`, macOS (universal) y Windows
`amd64`, este sin probar. El paso y los controles corren en `ubuntu-latest` y en el equipo de quien desarrolla.

**Project Type**: CLI multicall + un programa de construcción del repositorio + release.

**Performance Goals**: ninguna cota: el spec no pide tiempos. Medido en el plan: el snapshot entero, con el paso, 18 s;
el paso solo, menos de 2 s.

**Constraints**: la descripción corta, 120 caracteres como mucho; el icono, 512 × 512 px; dos ejecuciones del paso, 0
bytes de diferencia; el `.mcpb`, cuatro entradas; el binario distribuido no enlaza el paso; ninguna sesión del run
ejecuta `claude`, publica ni usa la red; ningún `//nolint`.

**Scale/Scope**: 10 herramientas y 2 skills con el árbol de hoy; `kitlegal.mcpb`, 42,7 MB; `kitlegal-plugin.zip`,
17,7 KB; el catálogo, 619 bytes; 6 subpruebas nuevas del snapshot, 6 pasos nuevos de `humo` y 1 trabajo nuevo.

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones; las desviaciones justificadas
están en Complexity Tracking.*

### Principios

| Principio | Cómo lo cumple H22 |
|---|---|
| I · Fuentes públicas y frontera humana | Ninguna fuente nueva y `docs/SOURCES.md` sin cambios (FR-053). El paso no pide nada a la red (FR-001) y el módulo no gana ninguna llamada HTTP. Lo único que escribe fuera es el flujo de la release, con `gh`: el catálogo en `jmorenobl/kitlegal-plugins`, con el token de publicación, como hoy goreleaser con el tap y el bucket; no es una sede ni una acción con la identidad de quien usa el kit. Etiquetar y publicar siguen siendo humanos (ADR 0020), y ninguna tarea ejecuta `release.yml`. |
| II · Nada sin cita ni fuente | Ningún applet cambia su sobre. El plugin lleva las dos skills byte a byte, con sus reglas de cita y con la línea `⚠ SIN CONSULTA AL BOE:` para cuando no hay herramienta (H21). El README dice que hacen falta las dos piezas y qué responde cada combinación (FR-050). El manifiesto y el catálogo no afirman nada legal. |
| III · Tests primero y offline | El hito no tiene comportamiento nuevo del binario: «Aceptación e2e: no aplica», con su motivo. Cada pieza de código llega con sus tests unitarios offline en la misma tarea (FR-069), y las subpruebas del snapshot se escriben y se ven en rojo antes del cambio de `.goreleaser.yaml` que las pone en verde. Ningún test usa la red, y ninguno toca `testdata/` ni `schemas/`. Cobertura dentro de `codecov.yml`. |
| IV · Hexagonal y errores tipados | `internal/empaquetado` no es dominio ni adaptador del binario: es la lógica de un programa de construcción, que compone con `internal/app` (el registro) y con lo empotrado. El binario no lo enlaza, y un test lo fija. Los códigos del ADR 0023 son del contrato de resultados del binario y no cambian; el paso sale con 0 o con 1 y una línea en la salida de error (FR-005; research D14). Ningún `panic`: cada fallo vuelve como error. Reglas R1-R7, abajo. |
| V · Simplicidad y dependencias | Ninguna dependencia nueva. Cada mecanismo se traza a un FR («Trazabilidad»). Rechazados por no pedirlos el spec: validar los binarios dentro del paso, un séptimo archivo, el esquema oficial en `testdata/`, un cliente MCP propio en el paso, plantillas JSON, guiones `testscript` para el paso y un test que ejecute los pasos de `humo`. |
| VI · Un binario, convenciones de agente | El binario distribuido sigue siendo uno, `kitlegal`, con los applets, los verbos y las banderas de hoy (FR-001). `cmd/empaquetar` es un programa del repositorio que goreleaser no construye ni publica: declarado en Complexity Tracking. `tools` del manifiesto sale del registro, como las herramientas y las tablas de las skills. |
| VII · Grafo y privacidad | No se toca el grafo ni la caché. La ficha y el catálogo no llevan ningún dato de persona: la autoría es `kitlegal` (research D4). El README dice dónde escribe kitlegal y que no envía nada a ningún servidor propio (ADR 0027). |
| VIII · Skills primero | No entrega skill nueva ni cambia las que hay: las protege poniéndolas, con sus herramientas, al alcance de quien no usa una terminal. El plugin es lo empotrado, sin código ni servidor; ninguna herramienta nueva. Las evals no cambian. |
| IX · Genericidad territorial | Nada territorial: ni código, ni datos, ni textos particulares de un municipio. `territorio_resolver` es una herramienta más de `tools`, con su descripción de siempre. |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores ni el kernel | `internal/core` no cambia. `internal/empaquetado` entra en la lista `core` de `depguard` y en `paquetesInternos` de `TestArquitectura`, como `disco` en H19 y `mcp` en H21: once paquetes. |
| R2 · Solo `internal/httpx` importa `net/http` | El paso no importa `net/http` ni pide nada. |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | El paso no los importa: construir el registro de producción no abre ninguna base (research V27). |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | El único `os.Exit` nuevo está en `cmd/empaquetar/main.go`; `empaquetado.Ejecutar` devuelve el código. Lo cubre la excepción `^cmd/` que `forbidigo` ya tiene (research V21). |
| R5 · Solo `render` escribe en stdout | El paso no escribe nada en la salida estándar: sus dos órdenes escriben ficheros (research D16). `os.Stderr` solo se nombra en `cmd/empaquetar/main.go`, que lo inyecta; ningún `fmt.Print*`. |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios. |
| R7 · Solo `internal/mcp/**` importa el SDK de MCP | Sin cambios (FR-069): el paso no habla MCP, y los tests de la raíz hablan con el servidor por `internal/mcp/mcptest`, que es de ese árbol (research V21). |

### Gates

- **Capa 1**, todo en `make ci` o en el workflow: los tests del paso; `TestConfiguracionDeLaRelease` con la definición
  nueva de la release y de los dos flujos; `goreleaser check`; `TestArquitectura` y `TestElBinarioNoEnlazaElPaso`;
  `depguard`; `schema-check` y `skills-check`, sin tocar; guardián de diff sin ningún fichero de `testdata/` ni de
  `schemas/`; «Controles de umbral» con requisitos del spec. En el cierre, el trabajo `snapshot` de la propuesta de
  cambio: `make release`, `make snapshot-check` y `make plugin-check`.
- **Capa 2**: los jueces del plan, de las tareas y de la revisión final, que comprueba la documentación contra FR-050 a
  FR-053 (SC-012).
- **Capa 3**, al leer el informe final: los supuestos del run —entre ellos los textos de la ficha y la autoría—; que
  `PUBLISHER_TOKEN` pueda escribir en el catálogo; la etiqueta, la release y la prueba humana de SC-002; la fusión.
  **Ninguna pausa a mitad del run.**

## Project Structure

### Documentation (this feature)

```text
specs/016-h22-instalar-sin-terminal/
├── plan.md                  # este fichero
├── research.md              # V1-V32, S1-S11, D1-D20
├── data-model.md            # el paso, la extensión, el manifiesto, el plugin, los textos, el catálogo
├── quickstart.md            # §1-§10
├── contracts/
│   ├── paso.md              # la orden, los dos zips, el catálogo, lo que se exporta, tests, uso
│   ├── release.md           # goreleaser, Makefile, TestSnapshot, ci.yml, release.yml, TestConfiguracionDeLaRelease
│   └── documentacion.md     # README, CONTRIBUTING, CHANGELOG
├── checklists/              # del spec
└── tasks.md                 # la escribe /speckit-tasks
```

### Source Code (repository root)

```text
cmd/empaquetar/main.go                           # NUEVO: el main del paso
internal/empaquetado/                            # NUEVO: la lógica del paso
├── doc.go
├── textos.go                # NombreVisible, Descripcion, DescripcionLarga, Autoria; el límite de 120 caracteres
├── piezas.go                # los dos zips, el manifiesto, plugin.json y el icono
├── catalogo.go              # la plantilla del catálogo
├── ejecutar.go              # Ejecutar: las dos órdenes y la composición de producción
├── export_test.go
├── textos_test.go           # TestDescripcionCorta
├── piezas_test.go           # TestPiezas, TestPiezasReproducibles, TestPiezasSinEntrada, TestIcono
├── catalogo_test.go         # TestCatalogo
└── ejecutar_test.go         # TestEjecutar
internal/app/herramientas.go                     # + HerramientaAnunciada, HerramientasAnunciadas
internal/app/herramientas_test.go                # + TestHerramientasAnunciadas
internal/arch_test.go                            # + TestElBinarioNoEnlazaElPaso; paquetesInternos gana empaquetado
.golangci.yml                                    # depguard, lista core: + internal/empaquetado (R1)
.goreleaser.yaml                                 # builds.id, universal_binaries, archives.ids, extra_files
Makefile                                         # + plugin-check; líneas de ayuda
snapshot_test.go                                 # la tabla de TestSnapshot gana seis filas
snapshot_piezas_test.go                          # NUEVO (etiqueta snapshot): las seis subpruebas y TestPluginValido
release_test.go                                  # TestConfiguracionDeLaRelease, con lo nuevo
.github/workflows/ci.yml                         # trabajo snapshot: Claude Code y make plugin-check
.github/workflows/release.yml                    # atestación, seis pasos de humo, trabajo catalogo
README.md, CONTRIBUTING.md, CHANGELOG.md
```

`mcp/icon.png` ya está versionado y no cambia.

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, con un programa más en `cmd/` y su paquete en
`internal/`, al lado de los demás. El paso no es un applet (FR-001), así que no vive en `internal/app`, y necesita el
registro y lo empotrado, así que vive en el módulo raíz y no en `tools/` (research D1). Los directorios `mcp/` y
`plugin/` que `CLAUDE.md` prevé no ganan ficheros: el manifiesto, `plugin.json` y el catálogo los compone el paso, y no
hay plantillas en el árbol (research D4, D5).

## Aceptación e2e

**Aceptación e2e: no aplica.** El hito no añade ni cambia ningún comportamiento visible del binario: FR-001 y FR-080 lo
prohíben —los applets, los verbos, `--describe`, las herramientas, las `instructions` y las skills son los de hoy— y lo
que entrega son dos ficheros de la release, dos flujos y documentación (research D13). No hay tarea `[aceptacion]` ni
guiones en `aceptacion/`.

Lo que hace de aceptación en el run:

| Historia | Qué la fija | Dónde |
|---|---|---|
| US1 (1 a 6), US5 | las seis subpruebas nuevas de `TestSnapshot` | `make snapshot-check`, en el trabajo `snapshot` de la propuesta de cambio (SC-001, SC-003 a SC-007, SC-009) |
| US2 (1) | `TestPluginValido` | `make plugin-check`, en el mismo trabajo (SC-008) |
| US2 (2, 3), US3 (5) | la definición de `release.yml` | `TestConfiguracionDeLaRelease`, en `make ci` (FR-068); su ejecución real es la primera release |
| US3 (1 a 4) | `make release`; `TestPiezasReproducibles`, `TestPiezasSinEntrada`, `TestEjecutar`; `make ci` | el trabajo `snapshot` y `make ci` (SC-010, SC-011) |
| US3 (6) | los guiones y los tests que ya fijan el binario: `argumentos.txtar`, `cmd/kitlegal/main_test.go`, `schema-check`, `TestHerramientasDelServidor`; y `TestElBinarioNoEnlazaElPaso` | `make ci`, sin tocar lo que comprueban (FR-080) |
| US4 | la revisión final, contra [contracts/documentacion.md](./contracts/documentacion.md) | SC-012 |
| US1 (7) | la prueba humana | SC-002, fuera del run |

Las evals no cambian: si el cierre lanza el job, tiene que seguir en verde como en H21.

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

- **`plugin-check`** (nuevo, fuera de `ci`): `TestPluginValido`.
- **`snapshot-check`**: la misma receta, con seis subpruebas más dentro de `TestSnapshot`.
- **`release`**: la misma receta; goreleaser ejecuta ahora el paso.
- Con más dentro, sin cambiar: `test` (los tests del paso y los nuevos de `internal/app` e `internal/`), `lint`
  (`depguard` con el paquete nuevo en R1; los ficheros de etiqueta `snapshot`, que ya lintea) y `goreleaser-check`.

### Tests nuevos

En `make ci`, los de [contracts/paso.md §7](./contracts/paso.md):

| Test | Fichero | Cubre |
|---|---|---|
| `TestPiezas` | `internal/empaquetado/piezas_test.go` | FR-010 a FR-014, FR-020 a FR-022, FR-070; US5 |
| `TestPiezasReproducibles` | `internal/empaquetado/piezas_test.go` | FR-004, FR-067; SC-010 |
| `TestPiezasSinEntrada` | `internal/empaquetado/piezas_test.go` | FR-005 |
| `TestIcono` | `internal/empaquetado/piezas_test.go` | FR-016, FR-066; SC-009 |
| `TestDescripcionCorta` | `internal/empaquetado/textos_test.go` | FR-015, FR-066; SC-009 |
| `TestCatalogo` | `internal/empaquetado/catalogo_test.go` | FR-030, FR-031 |
| `TestEjecutar` | `internal/empaquetado/ejecutar_test.go` | FR-001, FR-005, FR-014, FR-020, FR-070 |
| `TestHerramientasAnunciadas` | `internal/app/herramientas_test.go` | FR-014 |
| `TestElBinarioNoEnlazaElPaso` | `internal/arch_test.go` | FR-001 |

Fuera de `make ci`, con la etiqueta `snapshot` ([contracts/release.md §3 y §4](./contracts/release.md)):

| Test | Fichero | Cubre |
|---|---|---|
| Seis subpruebas de `TestSnapshot` | `snapshot_piezas_test.go` | FR-060 a FR-064, FR-066; SC-003 a SC-007, SC-009 |
| `TestPluginValido` | `snapshot_piezas_test.go` | FR-023, FR-030, FR-065; SC-008 |

### Tests existentes que cambian

- `TestConfiguracionDeLaRelease` (`release_test.go`): las subpruebas de [contracts/release.md §7](./contracts/release.md).
  Sus tipos de lectura estricta ganan `builds[].id`, `universal_binaries` y `archives[].ids`; `comprobacionesDelHumo`,
  seis entradas; hay dos subpruebas nuevas, `universal` y `catalogo`. Cambia en la misma tarea que el fichero que fija.
- `TestSnapshot` (`snapshot_test.go`): seis filas más en su tabla. Las cuatro de hoy no cambian, y siguen pasando con
  el `dist/` nuevo (research V11).
- `TestArquitectura` (`internal/arch_test.go`): `paquetesInternos` pasa de diez a once.
- Ninguno se desactiva ni se salta. `TestHerramientasDelServidor`, `TestSkillsEmpotradas`, `TestEsquemasPublicados`,
  los guiones de `internal/app/testdata/script/` y las evals no se tocan (FR-080).

### Fixtures, `testdata/` y `schemas/`

Ninguno. No hay tareas `[datos]`: los tests crean en `t.TempDir()` sus binarios de prueba y un icono de otro tamaño, y
leen `mcp/icon.png` del árbol. El esquema oficial del manifiesto no se versiona (FR-017).

### CI

- `.github/workflows/ci.yml`, trabajo `snapshot`: dos pasos más, instalar Claude Code en la versión de `evals.yml` y
  `make plugin-check` (FR-065). El trabajo `ci` no cambia.
- `.github/workflows/release.yml`: `publicar` atesta nueve sujetos; `humo` gana seis pasos; trabajo nuevo `catalogo`.
- Los dos los comprueba `TestConfiguracionDeLaRelease` en `make ci`. El trabajo `snapshot` se ejecuta en la propuesta de
  cambio, tras la revisión final, y lo cuenta el cierre del workflow; `release.yml`, solo con una etiqueta que empuja
  una persona.

## Controles de umbral

Una fila por cada umbral del spec que se mide sin persona. Los seis de FR-070 —120 caracteres, 512 px, cero bytes entre
dos ejecuciones, cuatro entradas, el conjunto de `tools` y cero ficheros distintos en `skills/`— tienen un test de
`make ci` que se pone en rojo por encima del umbral, sobre binarios de prueba, y además su subprueba de `TestSnapshot`
sobre el snapshot real.

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| FR-015, FR-066, SC-009 (descripción) | la descripción corta, 120 caracteres como mucho | el paso rechaza una de 121 y el test lo ve, con la del repositorio medida; en el snapshot, `manifiesto-de-la-extension` | `ci:internal/empaquetado/textos_test.go:TestDescripcionCorta` |
| FR-016, FR-066, SC-009 (icono) | `icon.png` es `mcp/icon.png`, un PNG de 512 × 512 px | el paso rechaza otro tamaño y lo que no es un PNG, y el test mide el del árbol; `TestPiezas` compara los bytes del icono del `.mcpb`; en el snapshot, `icono-de-la-extension` | `ci:internal/empaquetado/piezas_test.go:TestIcono`, `ci:internal/empaquetado/piezas_test.go:TestPiezas` |
| FR-004, FR-067, SC-010 | 0 bytes de diferencia en cada fichero entre dos ejecuciones | dos ejecuciones sobre los mismos binarios de prueba, comparadas byte a byte | `ci:internal/empaquetado/piezas_test.go:TestPiezasReproducibles` |
| FR-010, FR-062, SC-005 (entradas) | 4 entradas en el `.mcpb` y ninguna más: 0 ejecutables de más | la lista exacta de entradas, con sus modos; en el snapshot, `binarios-de-la-extension` | `ci:internal/empaquetado/piezas_test.go:TestPiezas` |
| FR-014, FR-061, SC-004 (herramientas) | `tools` es exactamente lo que el binario anuncia por MCP, con sus descripciones: 10 de 10 hoy | `tools` del paso es lo que da `app.HerramientasAnunciadas` del registro de producción, que contiene las diez de hoy; y eso es, como conjunto, lo que el servidor lista por MCP; en el snapshot, `servidor-de-la-extension`, contra el binario real | `ci:internal/empaquetado/ejecutar_test.go:TestEjecutar`, `ci:internal/app/herramientas_test.go:TestHerramientasAnunciadas` |
| FR-020, FR-063, SC-007 | 0 ficheros de `skills/` distintos de lo empotrado, 0 de menos y 0 de más; 0 de `bin/`, `.mcp.json`, `mcpServers` y `.mcpb` | el plugin del paso es `plugin.json`, leído de forma estricta, y `kitlegal.Skills()` byte a byte, y nada más; en el snapshot, `skills-del-plugin`, contra lo que instala el binario real | `ci:internal/empaquetado/ejecutar_test.go:TestEjecutar`, `ci:internal/empaquetado/piezas_test.go:TestPiezas` |
| FR-012, FR-013, FR-017, FR-061, SC-004 (manifiesto) | versión `0.3`, cada campo de FR-012 con su valor y 0 de más; `version`, la del binario | el manifiesto del paso, leído de forma estricta, campo a campo; en el snapshot, `manifiesto-de-la-extension`, que compara `version` con lo que imprime el binario | `ci:internal/empaquetado/piezas_test.go:TestPiezas` |
| FR-011, FR-062, SC-005 (binarios) | 2 arquitecturas, una `amd64` y una `arm64`, y el de Windows, cada uno byte a byte el de su archivo | en `make ci`: el paso copia los binarios que recibe sin cambiar un byte, y goreleaser le da el universal de la construcción `kitlegal` y el de Windows `amd64`; la medida sobre los binarios reales, `binarios-de-la-extension`, en el cierre | `ci:internal/empaquetado/piezas_test.go:TestPiezas`, `ci:release_test.go:TestConfiguracionDeLaRelease` |
| FR-002, FR-060, SC-003 | 2 de 2 ficheros en `dist/` y 2 de 2 en `checksums.txt` | en `make ci`: el paso escribe los dos, y los dos están en `checksum.extra_files` y en `release.extra_files`; la medida sobre `dist/`, `dos-piezas`, en el cierre | `ci:internal/empaquetado/piezas_test.go:TestPiezas`, `ci:release_test.go:TestConfiguracionDeLaRelease` |
| FR-006 | exactamente 6 archivos, y ninguno con el universal | en `make ci`: `universal_binaries` con `id` propio y `replace: false`, y `archives` con `ids`; la medida, `seis-archivos` de `TestSnapshot`, que ya existe, en el cierre | `ci:release_test.go:TestConfiguracionDeLaRelease` |
| FR-064, SC-006 | el servidor, arrancado con el `mcp_config` del manifiesto desde una ruta con espacios y con `/` de directorio, lista las 10 herramientas | en `make ci`: `mcp_config` lleva la orden y los argumentos del contrato, y el servidor responde igual desde `/` y desde una ruta con espacios (guion `h21-mcp-proceso`, sin tocar); la medida con el `.mcpb` real, `servidor-de-la-extension`, en el cierre | `ci:internal/empaquetado/piezas_test.go:TestPiezas`, `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| FR-065, SC-008 | `claude plugin validate` da por válidos 2 de 2 | `make plugin-check` (`TestPluginValido`) en el trabajo `snapshot`, en el cierre; en `make ci`, que ese trabajo lo ejecuta, con la versión de Claude Code de `evals.yml` | `ci:release_test.go:TestConfiguracionDeLaRelease` |
| FR-001 | 0 paquetes del paso en el binario distribuido | el cierre de `./cmd/kitlegal` no contiene `internal/empaquetado` | `ci:internal/arch_test.go:TestElBinarioNoEnlazaElPaso` |
| FR-003, FR-031, FR-040, FR-068 | 3 trabajos en `release.yml`; `PUBLISHER_TOKEN` en 2 pasos; 9 sujetos atestados; 12 pasos de `humo` | la definición de la release, leída de forma estricta | `ci:release_test.go:TestConfiguracionDeLaRelease` |

**Lo que solo se mide sobre el snapshot real.** Cinco medidas necesitan las seis plataformas construidas o Claude Code,
y por eso no caben en `make ci` (spec, «Assumptions», «Dónde corren los controles»): los bytes de las arquitecturas
(FR-062), los dos ficheros en `dist/` (FR-060), los seis archivos (FR-006), el servidor arrancado desde el `.mcpb`
(FR-064) y `claude plugin validate` (FR-065). Su control es `make snapshot-check` o `make plugin-check` en el trabajo
`snapshot` de la propuesta de cambio: con la medida fuera del umbral, ese trabajo sale en rojo y el cierre del workflow
lo cuenta (`medir_cierre`). La columna «Dónde» solo admite controles de `make ci` o del job de evals, así que en esas
filas nombra el test de `make ci` que cubre la misma promesa hasta donde `make ci` llega, y la columna «Control» dice
cuál es el del cierre. Las demás medidas se ponen en rojo en `make ci`.

SC-001 y SC-011 no son un umbral más: son que esos controles pasen. SC-002 lo mide una persona. SC-012 no tiene umbral
numérico: lo comprueba la revisión final.

## Uso, de fuera adentro (criterio de uso, ADR 0028)

Detalle, con ejemplos y bytes medidos, en [contracts/paso.md §2 a §4 y §8](./contracts/paso.md). Resumen:

- **`kitlegal.mcpb`** (la persona; una vez por versión): 42 711 814 bytes; cuatro entradas. Crece con el binario. No da
  señales.
- **La ficha** (la persona; una vez, al instalar): `manifest.json`, 2 939 bytes: una descripción de 71 caracteres, un
  párrafo de 491 y diez herramientas (1 383 bytes). Crece con los verbos. El aviso rojo es de la app: el README lo
  explica.
- **`kitlegal-plugin.zip`** (la persona; una vez por versión si lo sube, ninguna con el catálogo): 17 739 bytes; seis
  entradas. Crece con las skills.
- **Las skills del plugin** (el modelo; una vez por conversación en que se activa cada una): las de hoy, byte a byte.
- **Las herramientas de la extensión** (la skill; las veces por pregunta de H21: por cada bloque, `boe_articulo` y
  `graph_check`, más `boe_buscar` o `boe_indice` si hacen falta; `territorio_resolver`, una por municipio): el sobre de
  cada llamada, acotado por la norma y el bloque pedidos. Sus señales se apagan como en H21 y H7.1.
- **El catálogo** (la app o Claude Code de quien lo añadió; al mirar si hay versión nueva): 619 bytes, una entrada, que
  cada etiqueta sustituye. La versión nueva deja de ofrecerse al instalarla.
- **Dos líneas de `checksums.txt`** (`humo` y `catalogo`, una vez por release): 166 bytes.
- **La línea de error del paso** (quien mantiene el proyecto): una línea; deja de darse cuando la entrada está.

Con meses de uso —cientos de normas y miles de bloques consultados— todas miden lo mismo: ninguna depende de lo
consultado. Lo que sí crece con el uso, la caché y el grafo de `~/.cache/kitlegal/`, no cambia en este hito.

## Decisiones

- **El paso es `cmd/empaquetar` más `internal/empaquetado`**, y no un applet, un guion ni un módulo de `tools/` (D1).
- **`tools` sale del registro en el proceso del paso**, por una función exportada nueva de `internal/app`, y no de
  arrancar un binario (D2); **las skills, de lo empotrado** (D3).
- **Los textos son constantes de un fichero** (D4), y sus valores, un supuesto del plan: nombre visible `kitlegal`; la
  descripción corta, la que ya llevan el cask, el bucket y los paquetes; autoría `kitlegal`, sin correo ni nombre de
  persona (supuesto).
- **La plantilla del catálogo es código** con tres valores que cambian, y no un JSON con marcadores (D5; supuesto).
- **El paso lo ejecuta el gancho `post` de `universal_binaries`**, con la ruta del binario de Windows escrita con el
  sufijo que pone goreleaser (D6).
- **`builds[0].id` se escribe**, y el universal lleva `id` propio (D7).
- **Los zips llevan fecha fija, 1980-01-01, y orden fijo** (D8).
- **Las comprobaciones del snapshot son subpruebas de `TestSnapshot`**; `claude plugin validate`, un objetivo aparte
  (D9, D10).
- **`humo` habla MCP con un cliente de shell** y separa las arquitecturas con `od`, `head` y `tail` (D11).
- **`catalogo` es un trabajo aparte**, con Go para componer y la API de contenidos para publicar (D12).
- **Aceptación e2e: no aplica** (D13). **El paso sale con 0 o con 1** (D14).
- **`tools` se compara como conjunto** (D15). **El paso no escribe en la salida estándar** (D16).
- **Un test fija que el binario no enlaza el paso** (D17). **Datos externos: ninguno** (D18).
- Las demás, con su alternativa rechazada, en research D1-D20.
- **Reparación del cierre: la `description` de `boe-legislacion` nombra las herramientas de `boe`** (v0.1.6), y se
  aparta de «las dos skills, byte a byte»: la medición sobre `4b350fa` dio una respuesta de 54 sin la skill activada
  en el modo herramienta, con umbral 0 (D21). Nombra solo las `boe_…`, con las que se lee una norma, y no las de
  `territorio` ni las de `graph` (revisión final, ronda 3). Su efecto lo mide el job de la medición siguiente, en
  `boe-legislacion` y en `legal-core`.

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| `cmd/empaquetar/main.go` | FR-001 |
| `empaquetado.Ejecutar`, con las órdenes `piezas` y `catalogo` | FR-001, FR-005, FR-031 |
| La composición de producción: `app.RegistroDeProduccion`, `app.HerramientasAnunciadas`, `kitlegal.Skills()` | FR-014, FR-020 |
| `app.HerramientaAnunciada` y `app.HerramientasAnunciadas`; `TestHerramientasAnunciadas` | FR-014 |
| Los dos zips, con orden, fecha y modos fijos | FR-004, FR-010, FR-011, FR-020, FR-022 |
| El manifiesto, de un tipo con los campos de FR-012 | FR-012, FR-013, FR-017 |
| `plugin.json`, de un tipo con los campos de FR-021 | FR-021, FR-022 |
| `textos.go`: cuatro constantes exportadas | FR-015; las lee `snapshot_piezas_test.go` (FR-061) |
| El rechazo de una descripción de más de 120 caracteres y de un icono que no es un PNG de 512 × 512 px | FR-015, FR-016, FR-066 |
| `catalogo.go` | FR-030, FR-031 |
| La escritura con `os.Root` | FR-005 (fallar si no puede escribir), FR-069 (sin hallazgos de `gosec`) |
| `export_test.go` | FR-069, US5 (probar con una herramienta y una skill más) |
| `TestElBinarioNoEnlazaElPaso`; `internal/empaquetado` en R1 | FR-001, FR-069 |
| `builds[0].id`, `universal_binaries`, `archives[0].ids` | FR-006 |
| El gancho `post` | FR-001, FR-002, FR-005 |
| Los dos `extra_files` | FR-002 |
| Las seis subpruebas de `TestSnapshot` | FR-060 a FR-064, FR-066 |
| `TestPluginValido`, `make plugin-check`, los dos pasos del trabajo `snapshot` | FR-023, FR-030, FR-065 |
| Los dos sujetos nuevos de la atestación | FR-003 |
| Los seis pasos nuevos de `humo` | FR-040 a FR-043 |
| El trabajo `catalogo` | FR-031, FR-032 |
| `TestConfiguracionDeLaRelease`, con lo nuevo | FR-068 |
| README, CONTRIBUTING, `CHANGELOG.md` | FR-050 a FR-053 |

Nada del diseño atiende a un estado que no pasa el umbral de materialidad: un `dist/` tocado a mano entre goreleaser y
el paso, un zip alterado tras publicarse, unos binarios que no son lo que su nombre dice o una carpeta de salida a
medio escribir no tienen caso propio. El paso copia lo que recibe y falla si no puede leer o escribir; lo demás lo ven
las comprobaciones del snapshot, las huellas y `humo`.

## Datos externos

Ninguno (research D18): ni fuente, ni grabación, ni fila nueva en `docs/SOURCES.md`. Ningún manifiesto
`grabaciones.json` ni test `TestGrabar*` cambia, y el paso `grabar_datos` no tiene nada que grabar. El esquema oficial
del manifiesto de MCP Bundle queda fuera del run (Clarifications, pregunta 2): su fuente no tiene fila revisada en
`main`, no se versiona ni se sustituye por uno propio, y versionarlo es un pendiente de la persona.

## Orden de implementación (de dentro afuera)

1. `internal/app`: `HerramientaAnunciada` y `HerramientasAnunciadas`, con `TestHerramientasAnunciadas`.
2. `internal/empaquetado` entero —los textos, las piezas, el catálogo y `Ejecutar`, con sus siete tests— y
   `cmd/empaquetar/main.go`; con ello, `internal/empaquetado` en la lista `core` de `.golangci.yml` y en
   `paquetesInternos`, y `TestElBinarioNoEnlazaElPaso`. En una tarea: sin el `main`, nada usa `Ejecutar`.
3. El snapshot: primero las seis subpruebas de `TestSnapshot` en `snapshot_piezas_test.go`, vistas en rojo con
   `make release` y `make snapshot-check` sobre el árbol sin el cambio de goreleaser (faltan las dos piezas); después
   `.goreleaser.yaml` y, con él, la parte de goreleaser de `TestConfiguracionDeLaRelease`; y otra vez `make release` y
   `make snapshot-check`, en verde. En una tarea: `.goreleaser.yaml` y el test que lo lee de forma estricta cambian
   juntos.
4. `TestPluginValido`, `make plugin-check` y los dos pasos del trabajo `snapshot` de `ci.yml`, con la parte del
   `Makefile` y de `ci.yml` de `TestConfiguracionDeLaRelease`. La tarea no ejecuta `make plugin-check`: comprueba que
   compila y se lintea, y que `make ci` queda en verde.
5. `release.yml`: los dos sujetos de la atestación, los seis pasos de `humo` y el trabajo `catalogo`, con su parte de
   `TestConfiguracionDeLaRelease`.
6. README, CONTRIBUTING y `CHANGELOG.md`.

**Obligaciones para `tasks.md`**: ninguna tarea `[aceptacion]` y ninguna `[datos]`; cada tarea deja `make ci` en verde;
cada fila de «Controles de umbral» tiene la tarea que construye su control, con un test que lo ve fallar; ninguna tarea
cumple un umbral rebajándolo o retirándolo; ninguna edita un spec, un plan o una suite congelada de un hito anterior, ni
`scripts/workflow/`, ni las dos `SKILL.md`, ni `web/`, ni `docs/SOURCES.md` (FR-053, FR-080); ninguna toca `testdata/`
ni `schemas/`; ninguna usa la red —salvo la de las herramientas de Go—, ejecuta `claude`, `make plugin-check`,
`make evals` ni `release.yml`, etiqueta, publica ni mide en la plataforma; los pasos de `quickstart.md` §10 los ejecutan
la persona o el workflow.

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada |
|---|---|---|
| Un segundo programa en `cmd/`, `cmd/empaquetar`, cuando el principio VI habla de un binario | FR-001 pide un programa del repositorio que no sea un applet ni viaje en el binario. goreleaser solo construye `./cmd/kitlegal`, y `TestElBinarioNoEnlazaElPaso` fija que no lo enlaza | Un guion de shell: el hito pide Go, y `tools` saldría de una lista aparte |
| `internal/empaquetado` importa `internal/app`, la raíz de composición | `tools` tiene que salir del registro, «como las herramientas del servidor» (FR-014), y el registro vive ahí | Arrancar un binario y preguntarle por MCP (research D2) |
| Las mismas promesas se comprueban dos veces, en `make ci` con binarios de prueba y en `make snapshot-check` con los reales | FR-066, FR-067 y FR-070 piden el control en `make ci`; FR-060 a FR-064, sobre el snapshot. Miden cosas distintas: lo que hace el paso, y lo que queda en `dist/` | Solo el snapshot: los umbrales de FR-070 no tendrían control en `make ci` |
| Las comprobaciones de `humo` repiten, en shell, parte de lo que `TestSnapshot` hace en Go | `humo` no tiene el código del repositorio (FR-043) | Obtener el código en `humo`: deja de comprobar lo publicado como lo recibe quien lo instala |
| El trabajo `catalogo` instala Go para componer un fichero de 619 bytes | El catálogo sale de la misma función que valida la CI, con los textos de su único sitio (research D5, D12) | Rellenar un JSON con `jq`: los textos por duplicado y la lógica en shell, sin test |
| La versión de Claude Code aparece en dos flujos, `evals.yml` y `ci.yml` | FR-065 pide la versión que fija `evals.yml`; `TestConfiguracionDeLaRelease` falla si difieren | Leerla en el flujo con un paso más que analice `evals.yml`: más shell para lo mismo |
| En los tests de la raíz, el paquete se importa como `paso` | `release_test.go` ya declara un tipo `empaquetado` (research V23) | Renombrar ese tipo: tocar más de lo que el hito pide |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- `precheck.sh plan`: existen `plan.md` y `research.md`; ninguno conserva marcas de aclaración pendiente; están
  `## Constitution Check`, la línea «Aceptación e2e:» y `## Controles de umbral`, con cada fila nombrando requisitos
  que el spec define y un control con forma. Comprobado al terminar el plan con las órdenes del propio guion, una a una.
- a · Constitution Check: un ítem por principio (I-IX) y por regla de dependencia (R1-R7).
- b · Dependencias: ninguna nueva.
- c · Reglas de dependencia: tabla R1-R7; ni `net/http`, ni SQLite, ni salida estándar nuevos; `os.Exit`, solo en
  `cmd/`.
- d · Errores y códigos: el contrato de resultados del binario no cambia; el paso, 0 o 1 con su línea.
- e · Tests primero: «no aplica» con su motivo; tests nuevos, tests que cambian, ningún fixture, y las subpruebas del
  snapshot en rojo antes que la configuración.
- f · Alcance: nada fuera del spec; ni la web, ni las skills, ni el esquema oficial.
- g · Sin atajos: ningún `//nolint`, `t.Skip`, TODO ni error silenciado previstos; los hallazgos del linter sobre el
  prototipo, con su salida (research V21).
- h · Mejor alternativa: cada decisión con la rechazada (D1-D20).
- i · Afirmaciones verificadas: V1-V32 con fichero, línea u orden; S1-S11 como supuestos.
- j · Quickstart ejecutable: §2 a §5 ejecutados con el prototipo (research V31); los temporales y `dist/` se retiran.
- k · Datos externos: ninguno.
- l · Autonomía: ninguna tarea para una persona; `quickstart.md` §10 lo ejecutan la persona o el workflow.
- m · Uso: «Uso, de fuera adentro» y `contracts/paso.md` §8, con bytes medidos.
- n · Proporcionalidad: «Trazabilidad»; sin mecanismos para estados sin vía real.
- o · Controles de umbral: una fila por umbral, con lo que mide `make ci` y lo que mide el cierre.
