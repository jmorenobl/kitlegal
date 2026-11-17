# Implementation Plan: H41.4 · Las skills del producto, con las descripciones cortas que quitan la caída de la eval 04

**Branch**: `038-h41-4-las-skills-del` | **Date**: 2026-11-17 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/038-h41-4-las-skills-del/spec.md`

## Summary

Las cuatro `description` de `skills/` pasan a ser las de `laboratorio/producto-descripciones-cortas/`, en el mismo
bloque `>-` de hoy, partido con la regla de hoy en cuatro líneas sangradas (1690 caracteres entre las cuatro en lugar
de 3313), y nada más de las skills cambia (ADR 0055). Protegen el cambio en `make ci`: un fichero de test nuevo en
`internal/evals` con `TestDescripcionesDelProducto` (cada descripción de `skills/` es la de la variante, en un bloque
`>-` de más de una línea, con 450 y 1700 como límites) y `TestConsultaJuridicaDesdeLaBase` (fuera de la clave,
`consulta-juridica` es la de `ab951c22`, por huellas SHA-256); el único cambio en las suites congeladas,
`TestSkillsDelProductoDesdeControl`, que compara las tres familias fuera de las líneas de la clave y exige aparte el
texto de la variante (FR-005); y un guion e2e congelado sobre lo que instala `kitlegal skills install`. `CHANGELOG.md`
lo dice. El laboratorio, el arnés, las evals, la web y el empaquetado no cambian. La aceptación es el sondeo de la
persona, fuera del run. Todo verificado sobre una copia en el temporal (research V1-V18).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.4` en `go.mod`; `go version`: go1.27.4 linux/amd64).

**Primary Dependencies**: las fijadas, ninguna nueva. En los tests, `stretchr/testify` v1.12.1 y
`rogpeppe/go-internal` v1.16.0 (testscript), como hoy; `skills.LeerFrontmatter` con el lector de YAML que ya usa.

**Storage**: ficheros versionados: cuatro `skills/<skill>/SKILL.md` (solo las líneas de la clave), un fichero de test
nuevo y uno que cambia en `internal/evals`, un guion de `internal/app/testdata/script/` que activa el workflow y
`CHANGELOG.md`.

**Testing**: `go test -race` en `make ci` (`make test`); tests de tabla con `testify` sobre copias en `t.TempDir()`; un
guion testscript congelado. Ningún test con red ni con modelo. Skills aplicadas (research D10): `golang-how-to` con
`golang-testing`, `golang-stretchr-testify`, `golang-project-layout`, `golang-naming`, `golang-lint` y
`golang-safety`.

**Target Platform**: macOS y Linux, donde corren `make ci`, `kitlegal skills install` y el sondeo de la persona.

**Project Type**: skills sin código del producto y tests del repositorio; el binario solo cambia en las skills que
empotra.

**Performance Goals**: ninguno.

**Constraints**: 450 caracteres por descripción, 1700 entre las cuatro, 900 y ninguna etiqueta XML; cero bytes
cambiados en las skills fuera de las líneas de la clave, en `laboratorio/`, `evals/`, `web/`, `mcp/`, el empaquetado y
el job de evals; el run no graba, no escribe en `evidencias/` y no abre sesiones con modelo.

**Scale/Scope**: 4 skills, 4 ficheros de ellas, 20 líneas de bloque; 2 tests nuevos con 15 subtests (9 y 6); 1 test
que cambia; 1 guion con 8 `grep`; 1 viñeta de `CHANGELOG.md`.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio o regla | Cómo lo cumple este hito |
|---|---|
| I · Fuentes públicas y frontera humana (INNEGOCIABLE) | No toca ninguna fuente, `internal/httpx` ni petición. La de `jurisprudencia` sigue diciendo que kitlegal no consulta el CENDOJ. El sondeo lo lanza la persona con su credencial, fuera del run, con el arnés de H41.3 sin cambiar |
| II · Nada sin cita ni fuente | Ningún sobre cambia. Las descripciones siguen diciendo que se cita con lo leído («sin citar ninguna de memoria», «cita identificador y bloque») y que `plazos-y-territorio` no cuenta plazos de otras materias; la cobertura y lo no cubierto los dicen el cuerpo y las referencias, que no cambian un byte (FR-003) |
| III · Tests primero y offline | Primero el guion `aceptacion/skills-instaladas.txtar`, que hoy falla por una aserción (research V9). Los tests nuevos y el que cambia, offline contra los ficheros del repositorio y copias en `t.TempDir()`; ninguna grabación. Ningún código de producción cambia: la cobertura de `internal/core` y la global no se mueven; ninguna salida de applet nueva que validar contra `schemas/` |
| IV · Arquitectura hexagonal con reglas ejecutables | Solo cambian ficheros `_test.go` de `internal/evals` y datos de las skills; ningún paquete, import de producción, error ni código de salida nuevo. Un `SKILL.md` ilegible es el error de la comprobación (defecto `inesperado`, ADR 0023). `TestArquitectura` y `depguard`/`forbidigo` no cambian |
| V · Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia. Ningún paquete, tipo exportado, orden, bandera ni objetivo de `make`; las comprobaciones reutilizan `descripcionesCortas`, `descripcionYCuerpo` y `leerLosFicheros`; cada mecanismo se traza a un FR («Trazabilidad») |
| VI · Un binario, convenciones de agente | Un binario; ninguna orden, bandera ni variable nueva. `kitlegal skills install` instala las skills nuevas desde lo empotrado, como hoy (research V14) |
| VII · Grafo y privacidad | No toca el grafo; ningún dato de una persona |
| VIII · Skills primero; el binario es la herramienta | Mejora las cuatro skills del producto, medibles con sus evals en el cierre (FR-010) y con el sondeo de la persona (FR-011). Las descripciones siguen en tercera persona, sin etiquetas XML y bajo 900 (ADR 0046); la tercera persona la juzga el job (`habla_en_primera_o_segunda_persona`) |
| IX · Genericidad territorial, validación local | Ninguna descripción nombra un municipio («un municipio», «los festivos que cuentan»). Las evals de territorio del cierre siguen siendo la matriz cubierto/no cubierto |
| X · Genericidad por materia, validación administrativa | Las descripciones son de capacidad y «de cualquier materia»; `plazos-y-territorio` declara que no cuenta plazos de otras materias. Nada por materia en código, `data/` ni protocolo |
| R1 · `internal/core/**` no importa el kernel, adaptadores ni entrada y salida | `internal/core` no cambia |
| R2 · solo `internal/httpx` importa `net/http` | Nada nuevo importa `net/http` |
| R3 · solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | Nada nuevo los importa |
| R4 · solo `internal/cli` y `cmd/` llaman a `os.Exit` (`forbidigo`) | Ningún `os.Exit` nuevo |
| R5 · solo `internal/render` escribe en stdout; logs con `log/slog` a stderr (`forbidigo`) | Nada nuevo escribe en stdout; los tests informan con `testing.T` |
| R6 · `internal/graph` no importa `internal/source` ni `internal/render` | `internal/graph` no cambia |
| R7 · solo el paquete del servidor importa el SDK de MCP | Nada nuevo lo importa |
| Garantía sin red (H19) del paquete raíz y del applet `skills` | `skills.go` y el applet no cambian; solo los bytes empotrados de las skills |

Re-evaluado tras la fase 1: sin violaciones; Complexity Tracking, vacío.

## Project Structure

### Documentation (this feature)

```text
specs/038-h41-4-las-skills-del/
├── plan.md              # este fichero
├── research.md          # fase 0: V1-V18 (lo verificado), S1-S2 (supuestos) y D1-D11 (las decisiones)
├── data-model.md        # fase 1
├── quickstart.md        # fase 1
├── contracts/
│   ├── skills.md        # los cuatro bloques byte a byte, sus medidas, el CHANGELOG y su uso
│   ├── controles.md     # los tests nuevos, el que cambia, sus casos y sus mensajes
│   └── aceptacion.md    # el guion congelado, byte a byte
├── aceptacion/          # lo escribe la primera tarea (skills-instaladas.txtar)
└── tasks.md             # fase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
skills/
├── consulta-juridica/SKILL.md        # CAMBIA: solo las líneas de la clave description (contracts/skills.md §2)
├── jurisprudencia/SKILL.md           # CAMBIA: ídem
├── normativa/SKILL.md                # CAMBIA: ídem
└── plazos-y-territorio/SKILL.md      # CAMBIA: ídem
internal/evals/
├── descripciones_del_producto_test.go   # NUEVO: TestDescripcionesDelProducto y TestConsultaJuridicaDesdeLaBase (contracts/controles.md §1 y §2)
└── laboratorio_familias_test.go         # CAMBIA: TestSkillsDelProductoDesdeControl (contracts/controles.md §3)
internal/app/testdata/script/
└── h41-4-skills-instaladas.txtar     # NUEVO: lo activa el workflow desde aceptacion/ (contracts/aceptacion.md)
CHANGELOG.md                          # CAMBIA: [Unreleased] → Cambiado (contracts/skills.md §4)
```

Sin cambios: `laboratorio/`, `evals/`, `data/`, `schemas/`, `web/`, `mcp/`, `internal/empaquetado/`,
`.goreleaser.yaml`, el arnés (`internal/evals/laboratorio*.go` que no son `_test.go`, `cmd/laboratorio`, `scripts/`),
`Makefile`, `.golangci.yml`, `.github/` y las referencias de las cuatro skills. No hay `plugin/` en el repositorio: el
plugin y el `.mcpb` los compone `internal/empaquetado` desde `skills/` al empaquetar (research V13).

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`: el producto en `skills/`, sus controles junto a los
de H41 a H41.3 en `internal/evals`, los guiones e2e en `internal/app/testdata/script/`. Ningún paquete nuevo; un
fichero de test nuevo junto al código que reutiliza (research D1).

## Diseño

- **Las descripciones** (contracts/skills.md; research D4, V1, V2): las líneas de la clave de cada `SKILL.md` dejan su
  sitio, detrás de `name:`, a `description: >-` y cuatro líneas sangradas de como mucho 120 caracteres; lo demás del
  fichero no cambia.
- **`TestDescripcionesDelProducto`** (contracts/controles.md §1; research D1, D6, D7): `defectosDeLasDescripcionesDelProducto(raiz)`
  da, por skill de `skills/`, el defecto del texto, el de la forma y el de la longitud, y al final el de la suma; el
  test exige cero en la rama y las líneas exactas en ocho mutantes sobre copias.
- **`TestConsultaJuridicaDesdeLaBase`** (contracts/controles.md §2; research D2, V3): las huellas de `ab951c22` y la
  primera ruta que falta, sobra o difiere; cero en la rama, el defecto exacto en cuatro mutantes, y ninguno si solo
  cambia la clave.
- **`TestSkillsDelProductoDesdeControl`** (contracts/controles.md §3; research D3, V7): «desde control» y
  «frontmatter» sin las líneas de la clave, y la `description` de la variante en los dos; nada más del test cambia.
- **El guion** (contracts/aceptacion.md; research D5, V9-V11): ocho `grep` sobre lo instalado, sin afirmar los cortes.
- **`CHANGELOG.md`** (contracts/skills.md §4; research D9).

**Aceptación e2e:** un guion testscript que escribe la primera tarea en
`specs/038-h41-4-las-skills-del/aceptacion/skills-instaladas.txtar` y queda congelado (contracts/aceptacion.md §2),
activado como `internal/app/testdata/script/h41-4-skills-instaladas.txtar` y ejecutado por `TestEntregaDelHito`: en lo
que pone `kitlegal skills install --dir`, el frontmatter de cada una de las cuatro skills abre con `name:` y la clave
`description: >-`, cuyo texto, en más de una línea sangrada, se lee como el de la variante, seguido de `metadata:` o del
cierre. Cubre la historia 1: FR-001, FR-002; SC-001 (el guion en verde). Las historias 2 y 3 son de `make ci` y no
tienen guion propio: lo que piden es que ficheros del repositorio sean, byte a byte, los de la base, la variante o lo
generado, y eso lo comparan los tests de contracts/controles.md con sus mutantes (FR-003, FR-005 a FR-007, FR-009;
SC-002, SC-004), que el binario empotra tal cual; `CHANGELOG.md` lo lee la revisión (FR-008). La 4 la mide el cierre
(FR-010, SC-005) y la hace la persona fuera del run (FR-011, SC-006).

## Orden de implementación

De dentro afuera, con `make ci` en verde tras cada tarea:

1. `[aceptacion]`: el guion de contracts/aceptacion.md, que hoy falla por una aserción (research V9).
2. Las cuatro líneas de la clave de `skills/*/SKILL.md` y el cambio de `TestSkillsDelProductoDesdeControl`, en una sola
   tarea: con las descripciones y sin el cambio, ese test sale en rojo (research V4, D8).
3. `internal/evals/descripciones_del_producto_test.go` (necesita las descripciones ya cambiadas: con las de hoy,
   `TestDescripcionesDelProducto` sale en rojo).
4. `CHANGELOG.md`.

## Trazabilidad

| Mecanismo | FR/SC |
|---|---|
| Las líneas de la clave de los cuatro `SKILL.md` (contracts/skills.md §2) | FR-001, FR-002, FR-004, SC-001, SC-003 |
| `defectosDeLasDescripcionesDelProducto`: el defecto del texto y el de la skill sin texto | FR-001, SC-001 |
| Ídem: el defecto de la forma | FR-002 |
| Ídem: el de la longitud y el de la suma | FR-004, SC-003 |
| `TestDescripcionesDelProducto` y sus nueve casos (contracts/controles.md §1.3) | FR-001 (US1.2 y el caso límite), FR-002, FR-004 |
| `huellasDeConsultaJuridicaEnLaBase` y `defectoDeConsultaJuridica` | FR-003, SC-002 |
| `TestConsultaJuridicaDesdeLaBase` y sus seis casos (§2.3) | FR-003 (US2.3), SC-002 |
| `copiaDeLasSkillsYLaVariante` y `cambiarElSkill`, los auxiliares de los mutantes | US1.2, US2.3 |
| `sinLasLineasDeLaDescripcion`, `exigirLaDescripcionDeLaVariante` y el parámetro del texto en `comprobarElProductoDesdeControl` | FR-005, SC-002 |
| `aceptacion/skills-instaladas.txtar` | FR-001, FR-002, SC-001 (III; ADR 0018) |
| La viñeta de `CHANGELOG.md` | FR-008 |

## Uso, de fuera adentro

- **Las cuatro descripciones** (contracts/skills.md §5): las lee el host una vez por sesión, en el listado de skills;
  ninguna invocación de `kitlegal` las lee, y las órdenes y llamadas por pregunta son las de la base, porque los
  cuerpos no cambian. 1690 caracteres (1728 bytes) fijos, frente a 3313: con meses de uso, cientos de normas y miles de
  bloques consultados viven en la caché y el grafo, no en el listado. Ejemplo con sus bytes, `normativa`: 438 de
  texto, 463 en el fichero. Ninguna señal: el texto cambia solo con una release.
- **Los defectos de los controles** (contracts/controles.md §5): los lee el mantenedor cuando `make ci` sale en rojo,
  una vez por ejecución; como mucho 13 líneas de 60 a 130 bytes con cuatro skills y una en `consulta-juridica`; dejan
  de darse cuando `skills/` vuelve a ser lo declarado. No dependen del uso.
- **El guion** (contracts/aceptacion.md §3): la línea y el patrón que no casa, una vez por `make ci`.
- **`CHANGELOG.md`**: una viñeta de seis líneas y 632 bytes (contracts/skills.md §4), que lee una vez quien actualiza
  a la release siguiente.
- **Las tablas del sondeo** (quickstart, escenario 8): las produce `make laboratorio-tabla`, que no cambia, una por fila,
  y las lee la persona una vez, antes de fusionar. La de H41.3, con 640 sesiones, mide 99 412 bytes y 653 líneas
  (research V16): las dos de H41.4, 320 sesiones cada una, rondarán la mitad cada una (estimado, no medido). No se
  repiten en el producto ni crecen con el uso.

## Gates: controles mecánicos que añade o toca

- **Objetivos de `make`**: ninguno nuevo ni cambiado. `make test`, dentro de `make ci`, corre los tests nuevos (research
  V16).
- **Tests nuevos**: `TestDescripcionesDelProducto` y `TestConsultaJuridicaDesdeLaBase`
  (`internal/evals/descripciones_del_producto_test.go`; contracts/controles.md §1 y §2).
- **Test que cambia**: `TestSkillsDelProductoDesdeControl` (`internal/evals/laboratorio_familias_test.go`;
  contracts/controles.md §3), el único de las suites congeladas de H39 a H41.3 (FR-005).
- **Sin cambiar ningún valor esperado** (research V4): `TestDescripcionesCortas`, `TestDescripcionesCortasPorLaEstructura`,
  `TestConLaDescripcion`, `TestVariantesDeControl`, `TestEleccionDeLosMutantes`, `TestSkillsDelRepositorio`,
  `TestArnesDeLasFamilias` y los guiones de H19 a H41.3.
- **Guion e2e**: el nuevo y congelado `h41-4-skills-instaladas.txtar`.
- **Fixtures**: ninguno; ninguna grabación ni fichero nuevo en `testdata/` fuera del guion que activa el workflow.
- **Configuración de verificación**: `.golangci.yml`, `Makefile` y `.github/` no cambian; ningún `//nolint` (research
  V6).

## Datos externos

Ninguno (research D11): no hay manifiesto `grabaciones.json`, test `TestGrabar*` ni tarea que grabe; el run no escribe
en `evidencias/` ni en `data/`. Las tablas del sondeo las deja la persona en `evidencias/`, fuera del run.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| FR-001, SC-001 | 0 bytes de diferencia entre la `description` de cada skill de `skills/` y su texto de la variante; 0 skills sin texto | Las líneas «la description no es la de …» y «no tiene descripción en …»; los tres mutantes de US1.2 y el de la quinta skill (contracts/controles.md §1.3) | ci:internal/evals/descripciones_del_producto_test.go:TestDescripcionesDelProducto |
| FR-002 | 0 skills con la clave fuera de un bloque `>-` de al menos dos líneas sangradas | La línea «no está en un bloque >- de más de una línea»; los mutantes del escalar y del bloque de una línea | ci:internal/evals/descripciones_del_producto_test.go:TestDescripcionesDelProducto |
| FR-004, SC-003 | Como mucho 450 caracteres cada `description` de `skills/` y 1700 la suma | Las líneas «tiene N caracteres (máximo 450)» y «suman N caracteres (máximo 1700)»; los mutantes de 451 y de 1701 | ci:internal/evals/descripciones_del_producto_test.go:TestDescripcionesDelProducto |
| FR-004, SC-003 | Como mucho 900 caracteres y 0 etiquetas XML en cada `description` | `ValidarFrontmatter`: «description de N caracteres (máximo 900)» y «description con la etiqueta XML …» | ci:internal/app/skills_test.go:TestSkillsDelRepositorio |
| FR-003, SC-002 | `consulta-juridica`: 0 bytes de diferencia con `ab951c22` fuera de las líneas de la clave; 0 ficheros de más o de menos | Las huellas de contracts/controles.md §2.1 y la primera ruta que falta, sobra o difiere; los mutantes de §2.3 | ci:internal/evals/descripciones_del_producto_test.go:TestConsultaJuridicaDesdeLaBase |
| FR-003, FR-005, SC-002 | `normativa`, `jurisprudencia` y `plazos-y-territorio`: 0 bytes de diferencia fuera de lo generado y de las líneas de la clave con su base y su declaración; 0 ficheros de más o de menos; el frontmatter sin la clave, el de su control; la `description`, la de la variante | «desde control», «movimientos» y «frontmatter» (contracts/controles.md §3); sus rojos, research V7 | ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl |
| FR-006 | 0 bytes cambiados en la variante `producto-descripciones-cortas` | Sus cuatro textos, sin cambiar los esperados | ci:internal/evals/laboratorio_descripciones_test.go:TestDescripcionesCortas |
| FR-006 | 0 bytes cambiados en los ficheros de las variantes de control | Las huellas SHA-256 por ruta | ci:internal/evals/laboratorio_familias_test.go:TestVariantesDeControl |
| FR-007 | 0 ficheros que cambie `make skills-sync` | Lo generado de `skills/` es lo que da la sincronización | ci:internal/app/skills_test.go:TestSkillsDelRepositorio |
| FR-009, SC-004 | 0 tests en rojo en `go test ./...` | `make test`, dentro de `make ci` | ci:Makefile:test |
| FR-009 | `TestConLaDescripcion` y el mutante de la descripción de `plazos-y-territorio` en verde, sin cambiar nada | Los dos tests, tal cual | ci:internal/evals/laboratorio_descripciones_test.go:TestConLaDescripcion, ci:internal/evals/eleccion_test.go:TestEleccionDeLosMutantes |
| FR-001, FR-002, SC-001 | 100 % de los `grep` del guion sobre lo instalado (8 de 8) | El guion congelado `h41-4-skills-instaladas.txtar` | ci:internal/app/e2e_test.go:TestEntregaDelHito |
| FR-010, SC-005 | Las evals de `normativa` en verde en el cierre, en los dos modos, y el juicio de la tercera persona de su descripción | El job de evals del cierre | evals:normativa:sin_activar:claude-sonnet-5-5:orden, evals:normativa:sin_activar:claude-sonnet-5-5:herramienta, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5, evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar, evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados, evals:normativa:duracion_de_las_sesiones:orden, evals:normativa:duracion_de_las_sesiones:herramienta, evals:normativa:duracion_del_juez:orden, evals:normativa:duracion_del_juez:herramienta, evals:normativa:habla_en_primera_o_segunda_persona |
| FR-010, SC-005 | Las evals de `jurisprudencia` en verde en el cierre, en los dos modos, y el juicio de la tercera persona de su descripción | El job de evals del cierre | evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden, evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5, evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar, evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados, evals:jurisprudencia:duracion_del_juez:orden, evals:jurisprudencia:duracion_del_juez:herramienta, evals:jurisprudencia:habla_en_primera_o_segunda_persona |
| FR-010, SC-005 | Las evals de `plazos-y-territorio` en verde en el cierre, en los dos modos, y el juicio de la tercera persona de su descripción | El job de evals del cierre | evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:habla_en_primera_o_segunda_persona |
| FR-010, SC-005 | El juicio de la tercera persona de la descripción de `consulta-juridica` | El job de evals del cierre | evals:consulta-juridica:habla_en_primera_o_segunda_persona |

FR-010 y SC-005 usan los umbrales del job tal como están (`.github/workflows/evals.yml`, las carpetas de `evals/` y el
juez de las descripciones), que el hito no toca: las filas nombran los que publicaron con `decide: true` los informes
del cierre de H41.3 y, porque cambian las cuatro descripciones, el del juez de las descripciones, que el job publica
por skill con `decide: true` cuando una cambia (research V12); el veredicto del job sale en rojo con cualquiera. Las
evals de `consulta-juridica` no publican ningún umbral con `decide: true`: su veredicto lo deciden sus series (ADR
0016), configuración que el hito no toca. FR-011 y SC-006 los mide la persona con el sondeo (capa 3), fuera de `make
ci` y del cierre; FR-008 no tiene control en `make ci` por el propio spec (lo lee la revisión). En FR-006 y FR-007, que
el diff no toque el arnés, `web/`, `mcp/` ni el empaquetado no es un umbral que mida `make ci`: lo guarda el guardián
de diff del workflow (`scripts/workflow/guardian-diff.sh`, las rutas que declara cada tarea, y ninguna declara esas),
configuración que el hito no toca, y lo enseña el escenario 4 del quickstart.

## Cierre y pendientes

- **SC-006**: el run no lanza el sondeo; la propuesta de cambio queda pendiente de sus dos tablas y su lectura, y no se
  fusiona sin ellas ni si el criterio no se cumple (quickstart, escenario 8).
- **FR-010**: si la medición del cierre da un rojo, la reparación no esquiva la caída ni cambia ninguna descripción
  para que pase; lo deja dicho en el informe. Ningún umbral se rebaja.
- **Para la persona, en el informe final**: S1 (que el host lea el bloque `>-` como el escalar del sondeo de H41.3 lo
  mide el sondeo, no el run) y que las dos tablas del sondeo se sacan por separado (research V17).

## Complexity Tracking

Ninguna desviación: ni dependencias fuera de la lista, ni violaciones de la constitución que justificar.
