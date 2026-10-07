# Implementation Plan: H23 · skill `jurisprudencia`: ninguna sentencia citada sin el documento que trae la persona + `cita preparar` y `cita cotejar`

**Branch**: `019-h23-skill-jurisprudencia-ninguna` | **Date**: 2026-10-07 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/019-h23-skill-jurisprudencia-ninguna/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de la constitución y
están en [research.md](./research.md) (D1-D30), cada una con su alternativa rechazada. Toda afirmación sobre Go, Kong,
los generadores de esquemas, el servidor MCP, `testscript`, los guiones del workflow o los tests que el hito rompe
remite a la tabla V de research.md, comprobada en local con un prototipo fuera del repositorio; las medidas, a la
tabla M; y lo que no se pudo comprobar —el modelo, la red y el lint del prototipo—, a los supuestos S1-S7.

## Summary

De fuera adentro: quien pregunta por una sentencia recibe la consulta exacta para buscarla o, si trae el documento,
su cita; nunca una sentencia de memoria. Lo entregan seis piezas:

1. **La skill `jurisprudencia` v0** (contracts/skill-jurisprudencia.md): siete pasos, dos formas fijas —la línea
   `⚠ SENTENCIA NO COMPROBADA:` y la cita con su corchete— y una tabla de comandos generada. El borrador, de 195
   líneas, está en contracts/SKILL-jurisprudencia.prototipo.md.
2. **El applet `cita`** (contracts/applet-cita.md), con dos verbos que no piden nada a la red: `preparar` da la
   dirección y las casillas, o la dirección de una búsqueda; `cotejar` lee la ficha del texto que le llega y dice si
   es el pedido, con un hallazgo y código 0 si no lo es. Un argumento escrito con valor vacío es un error de
   argumentos, no un argumento no dado (research D3). Su dominio, puro, en `internal/core/cita`; el ECLI y el ROJ,
   en `internal/core/ids`.
3. **La entrada estándar, entregada por el kernel** (research D12): el verbo la recibe en una orden y no en una
   llamada de herramienta, de modo que el servidor MCP gana `cita_preparar` y `cita_cotejar` sin tocar `mcp`.
4. **Las banderas propias en `--describe` y en la tabla de comandos** (research D14): sin ello, la tabla escribiría
   `--roj` como un argumento de posición.
5. **El formato de eval** (contracts/evals-jurisprudencia.md): dos formas de comando y la clave `sentencias`, con su
   juicio sin modelo, y las seis evals.
6. **El umbral que decide**: `cita_sin_documento:<modelo>:<modo>`, un hecho de la sesión, y `sin_activar`, en un
   trabajo nuevo de la matriz, `evals (jurisprudencia)`.

No hay ADR nuevo, ni fuente, ni dependencia, ni dato en `data/`.

## Technical Context

**Language/Version**: Go 1.27 (`go.mod`; el equipo, `go1.27.1`), `CGO_ENABLED=0`, `-trimpath`; YAML de GitHub
Actions y de las evals; Markdown de la skill.

**Primary Dependencies**: las fijadas, sin ninguna nueva: `alecthomas/kong` v1.16.1 (la gramática de los dos verbos),
`invopop/jsonschema` v0.14.0 (`--describe` y su anotación), `santhosh-tekuri/jsonschema` v6.0.3 (cada salida contra
su esquema, y las evals contra el suyo), `gopkg.in/yaml.v3`, `rogpeppe/go-internal` v1.16.0 (`testscript`),
`stretchr/testify`; biblioteca estándar (`strings`, `time`, `fmt`, `crypto/sha256`, `encoding/hex`, `io`,
`encoding/json`, `regexp`).

**Storage**: ninguno. El applet no abre la caché ni el grafo ni escribe ningún fichero. Ficheros nuevos del
repositorio: la skill, las seis evals, un esquema, once golden y tres corpus de fuzz.

**Testing**: `go test -race` en `make ci`: tablas con `t.Parallel`, golden contra su esquema, fuzz con su corpus,
cuatro guiones `testscript` contra el binario de e2e, el test de arquitectura y sesiones sintéticas para el juicio y
el informe. El job de evals, en la propuesta de cambio, tras la revisión final.

**Target Platform**: el binario, macOS, Linux y Windows, como hoy; el job, `ubuntu-24.04`.

**Project Type**: skill y CLI multicall (un applet nuevo en el binario único).

**Performance Goals**: ninguno con umbral. Los dos verbos son cómputo local sobre unos kilobytes: ni red, ni disco,
ni espera. El trabajo `evals (jurisprudencia)` va sin objetivo de duración (FR-063).

**Constraints**: el código del applet no alcanza `net`, `net/http` ni `internal/httpx` (FR-002); `SKILL.md` < 300
líneas; `mcp.go`, `herramientas.go`, `internal/mcp`, `cmd/empaquetar`, `internal/empaquetado`, `internal/httpx`,
`scripts/workflow/`, `evidencias/` y las otras dos skills no cambian; ningún `//nolint`; ningún test saltado.

**Scale/Scope**: 2 verbos, 2 herramientas (12 en total), 1 skill, 6 evals (72 sesiones por job: 6 × 2 modos × 2
modelos × 3 repeticiones), 4 umbrales, 4 guiones de aceptación, 11 golden; y lo que la entrega rompe y hay que
alinear: 11 tests de Go y 21 guiones e2e (research M4).

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones; las desviaciones
justificadas están en Complexity Tracking.*

### Principios

| Principio | Cómo lo cumple H23 |
|---|---|
| I · Fuentes públicas y frontera humana | Es el principio que el hito aplica: kitlegal no consulta el CENDOJ de ninguna forma. El applet no puede abrir una conexión, por lo que importa (FR-002; `TestArquitectura`); las direcciones que da las abre la persona. Ninguna petición HTTP nueva, ningún método, ninguna grabación, ninguna fila de `docs/SOURCES.md`. |
| II · Nada sin cita ni fuente | Los dos verbos emiten el sobre; `fuente` `kitlegal.cita` dice que no hubo consulta, y `url` identifica lo que abre la persona o el documento aportado por su huella (research D6). Nada entra en el grafo: lo leído no tiene fuente. La skill no cita una sentencia sin su documento cotejado, y `cita_sin_documento` lo decide con 0. |
| III · Tests primero y offline | Cuatro guiones `testscript` escritos antes que el código y congelados; unitarios offline; cada salida, contra `schemas/cita.json`; el dominio nuevo está en `internal/core/**`, con su cobertura. No hay fixtures de red: el applet no pide nada. |
| IV · Hexagonal y errores tipados | Dominio puro en `internal/core/cita` e `internal/core/ids`; el applet, en `internal/app`. Todo error es de clase `argumentos` (2), declarada por el dominio; que el documento no sea el pedido es un hallazgo con 0 (ADR 0023); lo demás, `inesperado` (1). Ningún `panic`: lo ejercen tres fuzz. |
| V · Simplicidad y dependencias | Ninguna dependencia nueva. Cada mecanismo, con su requisito («Trazabilidad»). Se rechazaron un adaptador de fuente, un tipo de argumento con memoria, un lector compartido con estado, una variable nueva en el arnés e2e, una declaración de umbrales por skill y campos nuevos en el informe (research D1, D3, D12, D18, D20, D24). |
| VI · Un binario, convenciones de agente | Un applet más en el multicall, con las banderas globales heredadas. `--describe` sigue siendo la única fuente de las herramientas y de la tabla: por eso es ahí donde se dice qué argumentos son banderas (research D14). |
| VII · Grafo y privacidad | El applet no emite operaciones de grafo ni abre `world.db`; el documento de la persona no se guarda en ningún sitio, y de él solo sale su ficha (FR-003, FR-022). |
| VIII · Skills primero | El hito entrega una skill medible con evals; el binario solo hace lo que exige determinismo: reconocer identificadores, componer la consulta, leer la ficha y comparar. `SKILL.md` de 195 líneas, sin `scripts/`, con cada operación como orden y como herramienta. |
| IX · Genericidad territorial | Sin dimensión territorial: ni municipio, ni comunidad, ni boletín. Lo que queda fuera de cobertura se declara en `data` (`cobertura`, solo con el ECLI del Tribunal Constitucional: FR-014, research D4), y ni el binario ni la skill concluyen que una sentencia no existe ni que el CENDOJ la tiene. Las dos reglas que nombran un órgano —la equivalencia del Tribunal Supremo y la cobertura del Constitucional— son las que el spec fija con su fuente (FR-012, FR-014), no casos de un territorio. |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | `internal/core/cita` importa `internal/core/ids`, `internal/core/schema` y la biblioteca estándar; `ids`, lo de hoy. Comprobado en el prototipo (research V22). |
| R2 · Solo `internal/httpx` importa `net/http` | Ningún fichero nuevo lo importa. Y más allá de R2: el dominio y `app/cita.go` no alcanzan `net` (FR-084). |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | Nada nuevo los importa. Que la caché y el grafo no cambian se comprueba desde fuera, con un guion (research D25). |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Ninguna llamada nueva. |
| R5 · Solo `render` escribe en stdout | El applet devuelve un `Resultado`. La entrada estándar la nombra la raíz de composición (`RegistroDeProduccion` y el binario de e2e), como hoy la del servidor; `forbidigo` no la veta (research V23). |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios. |
| R7 (`internal/arch_test.go`) · Solo `internal/mcp` importa el SDK | Sin cambios: las herramientas salen del registro. |

## Project Structure

### Documentation (this feature)

```text
specs/019-h23-skill-jurisprudencia-ninguna/
├── plan.md                  # este fichero
├── research.md              # V1-V42, M1-M5, S1-S7, D1-D30
├── data-model.md            # identificadores, referencia, consulta, ficha, cotejo, cita, eval, umbral
├── quickstart.md            # §0-§9; §10, el cierre; §11, la persona
├── contracts/
│   ├── applet-cita.md                       # verbos, sobre, data, errores, entrada, --describe, golden, uso
│   ├── evals-jurisprudencia.md              # formato, juicio, reconocimiento, umbrales, los seis ficheros, job
│   ├── skill-jurisprudencia.md              # qué lleva SKILL.md y por qué, formas fijas, uso
│   └── SKILL-jurisprudencia.prototipo.md    # el borrador de SKILL.md, con su tabla generada
├── checklists/              # del spec
├── gates/supuestos.md       # con los supuestos del plan
└── tasks.md                 # la escribe /speckit-tasks
```

### Source Code (repository root)

```text
skills/jurisprudencia/SKILL.md            # nueva: la skill, sin references/ ni scripts/
evals/jurisprudencia/                     # nuevas: las seis evals, con el formato que las admite
internal/core/ids/
├── ecli.go, roj.go                       # nuevos: AnalizarECLI, AnalizarROJ y la equivalencia
└── testdata/fuzz/FuzzECLI, FuzzROJ       # corpus [datos]
internal/core/cita/                       # nuevo paquete, dominio puro
├── doc.go, referencia.go                 # las tres formas y sus errores
├── consulta.go                           # Preparar, PrepararTexto, la codificación del texto
├── ficha.go                              # LeerFicha
├── cotejo.go                             # Cotejar, la correspondencia y el hallazgo
└── testdata/fuzz/FuzzLeerFicha           # corpus [datos]
internal/app/
├── cita.go                               # nuevo: el applet y sus dos verbos
├── registro.go                           # AppletCita en producción; Registro.LeerDe
├── despacho.go, main.go                  # Despacho.Entrada; el verbo que lee la entrada
├── ejemplo/kitlegal-e2e/main.go          # el applet y la entrada, en el binario de e2e
└── testdata/cita/                        # once golden [datos]
internal/cli/describe.go                  # la anotación x-banderas
internal/skills/comandos.go               # las banderas propias en la orden de la tabla
internal/evals/
├── formato.go                            # Eval.Sentencias; los comandos de cita
├── sentencias.go                         # nuevo: citas, línea, ECLI y sobres de un texto; el juicio; el hecho de la sesión
├── juzgar.go, consultas.go               # el juicio de sentencias; cita no pide grabaciones
├── conjunto.go                           # ReglasDeJurisprudencia; si el conjunto declara sentencias
├── umbrales.go, informe.go               # cita_sin_documento y sin_activar sin juez; su motivo
└── doc.go
internal/arch_test.go                     # la garantía sin red, con cita
schemas/cita.json                         # nuevo [datos]
schemas/instalacion.json                  # x-banderas en sus tres partes [datos]
schemas/eval.yaml.json                    # sentencias y los comandos de cita [datos]
.github/workflows/evals.yml               # jurisprudencia en la matriz
README.md, docs/JURISPRUDENCIA.md, CHANGELOG.md, CONTRIBUTING.md
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`: dominio en `internal/core/{ids,cita}`, applet y
composición en `internal/app`, la skill en `skills/`, sus evals en `evals/`. Un paquete nuevo, `internal/core/cita`,
que es el que §2 nombra para el dominio de las citas y al que H8 añade las de normas. Ningún adaptador de fuente.

## Aceptación e2e

**Aceptación e2e:** cuatro guiones `testscript`, que escribe la primera tarea en
`specs/019-h23-skill-jurisprudencia-ninguna/aceptacion/` y quedan congelados; el workflow los activa como
`internal/app/testdata/script/h23-<nombre>.txtar` (research V31).

| Guion | Historia | Qué describe | Requisitos |
|---|---|---|---|
| `cita-preparar.txtar` | US1 | Las tres formas y el texto, con su dirección, sus casillas y su equivalente, y sin `cobertura` (0); el ECLI del Tribunal Constitucional, fuera de cobertura en `data` (0); el ECLI mal formado, el de otro país, `--resolucion` sin `--fecha`, la referencia con `--texto` —con un texto y con `--texto ""`—, la referencia con `--roj ""` y `cita` sin verbo (2, `argumentos`) | FR-001, FR-004 a FR-006, FR-010 a FR-015, FR-082; SC-003 |
| `cita-cotejar.txtar` | US2 | Con el fragmento por la entrada estándar: sin referencia, sus ocho datos, la correspondencia y la `url` con su huella; pedido por sus tres formas (0, es el pedido); como ROJ `STS 1088/2023`, con otra fecha y como número `3144/2023` (0, con su hallazgo y su cruce); un texto sin ficha (2); y `--documento ""` con el fragmento en la entrada, que no se lee (2) | FR-004, FR-020 a FR-026, FR-082; SC-003 |
| `cita-herramientas.txtar` | US4 | El servidor anuncia `cita_cotejar` y `cita_preparar`; `cita_cotejar` con la ficha —las once primeras líneas del fragmento— y el ROJ `STS 1088/2023` da un resultado con su hallazgo; sin `documento`, o con él vacío, el error `argumentos`, y el servidor sigue; y una llamada con resultado a cada una de las dos da el sobre de su orden salvo `fecha_consulta` (US4.3), con cada uno de sus seis argumentos en alguna: `cita_cotejar` con la ficha sola y con la ficha y el ROJ `STS 1088/2023`, y `cita_preparar` con `resolucion` `1088/2023` y `fecha` `2023-07-04`, con `texto` `cláusula suelo`, con `roj` `STS 1088/2023` y con `ecli` `ECLI:ES:TS:2023:3144`, cada una frente a su orden con `--json` (research V42) | FR-020, FR-025, FR-026, FR-030 |
| `cita-sin-efectos.txtar` | US4 | Con la caché y el grafo ya con contenido, `arbol` da el mismo listado antes y después de los dos verbos, cada uno como orden y cada uno como herramienta, con una llamada con resultado a `cita_preparar` y otra a `cita_cotejar`, también con `--offline`, `--no-graph` y `--dry-run` | FR-003, FR-085; SC-006 |

Lo que usan del arnés es lo que ya hay: `stdin`, `exec`, `cmp`, `arbol`, `mcp` y `$KITLEGAL_SKILLS`, por la que
llegan al fragmento de la evidencia sin copiarlo (research V8, D24). Afirman lo que el spec y los contratos fijan
—códigos, clase, claves, valores de vocabulario, casillas, direcciones y la `url` del documento—, y no el texto
libre de `motivo`, `explicacion` ni de un mensaje de error. Cada uno falla hoy por su primera aserción: el binario
no tiene el applet `cita`.

La primera tarea no escribe ninguna eval: las seis de `evals/jurisprudencia/` (contracts/evals-jurisprudencia.md §5)
entran con la tarea del formato que las admite, porque `make ci` lee las evals de toda carpeta de `evals/` y sus
claves no existen antes (research V34, D23; el precedente, H21, V35). Se escriben antes que la skill. US3 y US5 no
tienen guion: su aceptación son esas evals en el job de cierre (SC-001) y, en `make ci`, los tests con sesiones
sintéticas de «Controles mecánicos». US6 la comprueba la revisión final (SC-010).

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

Ninguno nuevo ni cambiado. `test`, `schema-check` y `skills-check` llevan más dentro: los tests de abajo, un esquema
publicado más y una skill más.

### Tests nuevos

| Test | Fichero | Cubre |
|---|---|---|
| `TestAnalizarECLI`, `TestAnalizarROJ`, `TestEquivalencia` | `internal/core/ids/ecli_test.go`, `roj_test.go` | FR-005, FR-006, FR-012 (la tabla de los casos sin equivalente) |
| `FuzzECLI`, `FuzzROJ` | los mismos | FR-083 |
| `TestReferencia`, `TestPreparar` | `internal/core/cita/referencia_test.go`, `consulta_test.go` | FR-005, FR-006, FR-010 a FR-015: cada argumento, no dado, dado y dado vacío; `cobertura`, solo con el ECLI de órgano `TC` |
| `TestLeerFicha`, `FuzzLeerFicha` | `internal/core/cita/ficha_test.go` | FR-021, FR-026, FR-083 |
| `TestCotejar` | `internal/core/cita/cotejo_test.go` | FR-023 a FR-025 |
| `TestCitaPreparar`, `TestCitaCotejar` | `internal/app/cita_test.go` | Los once golden, cada salida contra su esquema; el sobre de FR-004; cada fila de errores de contracts/applet-cita.md §6, también con el argumento escrito vacío, por la línea de órdenes y por la línea de una llamada; `--documento ""` con un documento en la entrada, que no se lee (FR-006, FR-015, FR-020, FR-026, FR-080, FR-081; SC-004) |
| `TestEntradaDeLaOrden` | `internal/app/main_test.go` | El kernel da la entrada a la orden y no a la llamada de herramienta (FR-020) |
| `TestFormatoDeSentencias`, `TestJuzgarSentencias`, `TestCitaSinDocumento`, `TestUmbralesDeJurisprudencia`, `TestPreguntasConElFragmento` | `internal/evals/` (contracts/evals-jurisprudencia.md §7) | FR-052, FR-053, FR-055, FR-060 a FR-063, FR-086, FR-087; SC-007, SC-008 |

### Tests y guiones existentes que cambian

La lista es la que da el prototipo con el applet en los dos registros y la skill en `skills/` (research V21):

- **Por el applet**: `TestPuntoDeEntrada` (`cmd/kitlegal/main_test.go`), `TestRegistroDeProduccion`
  (`internal/app/registro_test.go`) y `TestRegistroDeE2E` (`internal/app/ejemplo/kitlegal-e2e/main_test.go`), la lista
  de applets; `TestHerramientasDelServidor` (`internal/app/herramientas_test.go`), de 10 y 13 herramientas a 12 y 15;
  `TestEsquemasPublicados` y `TestEsquemasCubrenTodosLosVerbos` (`internal/app/esquemas_test.go`), la fila de
  `cita.json`, su contrato de invocaciones y su descripción; `TestSuperficieDeIds`
  (`internal/core/ids/errores_test.go`), lo que `ids` exporta; `TestArquitectura` (`internal/arch_test.go`), los
  orígenes y los ficheros sin red.
- **Por la anotación y la tabla**: `TestTablaDeComandosCoincideConLaGramatica` (`internal/app/skills_test.go`), cuyo
  auxiliar pasa a entender una bandera; `TestEsquemasDeHerramienta` (`internal/cli/herramienta_test.go`) y el
  auxiliar `sinLasBanderasGlobales` de `herramientas_test.go`, que comparan sin la anotación; los de
  `internal/cli/describe_test.go` e `internal/skills/comandos_test.go`, con un verbo con banderas.
- **Por la skill**: el arnés de `TestSkillsDelRepositorio` (`internal/app/skills_test.go`), sin suponer referencias;
  `TestEvalsDelRepositorio/linea-sin-consulta` (`internal/evals/conjunto_test.go`), sobre las dos skills que llevan
  esa regla.
- **Por las evals y el job**: `TestEvalsDelRepositorio` y `TestConjuntoDeEvals`, con el conjunto de
  `jurisprudencia`; `TestDefinicionDelJob` (`internal/evals/definicion_test.go`), con la tercera skill; los de
  `formato`, `juzgar`, `consultas`, `umbrales` e `informe`, con lo nuevo.
- **Guiones e2e**, en `internal/app/testdata/script/`, todos `[datos]`. Por el applet, cuatro: `argumentos.txtar`
  (la lista de applets) y `h21-mcp-herramientas.txtar`, `h21-mcp-proceso.txtar` y `h21-mcp-protocolo.txtar` (de 13 a
  15 herramientas). Por la skill, diecisiete, que enumeran las skills empotradas y ganan la tercera, en su orden
  —`boe-legislacion`, `jurisprudencia`, `legal-core`—: `skills-salida-legible`, `skills-host-antigravity` y los
  `h19-skills-` `ambito-dir`, `ambito-global`, `aviso`, `aviso-sin-aviso`, `conflictos-dentro`,
  `conflictos-entradas`, `conflictos-rutas`, `doctor-copia`, `doctor-hallazgos`, `dry-run`, `idempotencia`,
  `install-hosts`, `install-local`, `list-doctor` y `no-empotrada`.

Ninguno se desactiva ni se salta, y ninguno pierde lo que comprobaba: cada cambio es el dato nuevo en una lista o
una cifra.

### Fixtures, `testdata/` y `schemas/` (tareas `[datos]`)

- `schemas/cita.json`, nuevo, y `schemas/instalacion.json`, con la anotación: los dos, regenerados desde
  `--describe`.
- `schemas/eval.yaml.json`: `sentencias` y las dos formas de comando.
- `internal/app/testdata/cita/`: los once golden (contracts/applet-cita.md §9).
- `internal/core/ids/testdata/fuzz/FuzzECLI/` y `FuzzROJ/`, e `internal/core/cita/testdata/fuzz/FuzzLeerFicha/`.
- Los veintiún guiones e2e de arriba.

Ninguna tarea escribe en la carpeta de la evidencia: los tests y los guiones leen de ella el fragmento.

### CI

`.github/workflows/evals.yml` gana la tercera skill en su matriz; lo comprueba `TestDefinicionDelJob`. `ci.yml` no
cambia. El job de evals lo lanza el workflow sobre la propuesta de cambio, tras la revisión final.

## Controles de umbral

Los cuatro umbrales que deciden en el informe de `jurisprudencia` (FR-066), los de `boe-legislacion` que SC-001 pide
que sigan cumplidos, y una fila por cada umbral que se mide en `make ci`. `legal-core` no publica umbrales: lo que
SC-001 pide de él es su veredicto, que cuenta el cierre.

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| FR-060, SC-001 (modo orden) | 0 respuestas con una cita sin documento cotejado o con un ECLI sin origen, de las 18 del modo | umbral del informe con `decide: true`: con una, motivo con la sesión y su ECLI, veredicto `fallo` y el trabajo `evals (jurisprudencia)` en rojo, que el cierre cuenta | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden` |
| FR-060, SC-001 (modo herramienta) | ídem | ídem | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta` |
| FR-062, SC-001 (sin activar, modo orden) | 0 respuestas sin la skill activada, de las 18 del modo | umbral del informe con `decide: true` | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden` |
| FR-062, SC-001 (sin activar, modo herramienta) | ídem | ídem | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta` |
| SC-001 (`boe-legislacion` sigue aprobada) | sus diez umbrales que deciden, cumplidos | los de hoy, que no cambian | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5:orden`, `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5:herramienta`, `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`, `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`, `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`, `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados`, `evals:boe-legislacion:duracion_de_las_sesiones:orden`, `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`, `evals:boe-legislacion:duracion_del_juez:orden`, `evals:boe-legislacion:duracion_del_juez:herramienta` |
| FR-040, SC-009 (líneas) | `SKILL.md` de `jurisprudencia` < 300 líneas | `TestSkillsDelRepositorio`, en `skills-check` | `ci:Makefile:skills-check`, `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-030, SC-009 (herramientas) | 12 herramientas anunciadas, las de los verbos del registro menos los excluidos | el control de conformidad de H21, con 12 y 15 | `ci:internal/app/herramientas_test.go:TestHerramientasDelServidor` |
| FR-082, SC-003 | 11 casos con su código: 7 con 0 y 4 con 2 | los guiones de aceptación de los dos verbos | `ci:internal/app/testdata/script/h23-cita-preparar.txtar`, `ci:internal/app/testdata/script/h23-cita-cotejar.txtar` |
| FR-080, FR-081, SC-004 | 11 de 11 golden iguales salvo `fecha_consulta`, y 11 de 11 salidas válidas contra su esquema | comparación byte a byte y validación | `ci:internal/app/cita_test.go:TestCitaPreparar`, `ci:internal/app/cita_test.go:TestCitaCotejar` |
| FR-002, FR-084, SC-005 | 0 caminos de dependencia a `net`, `net/http` o `internal/httpx` | el recorrido del grafo de importaciones | `ci:internal/arch_test.go:TestArquitectura` |
| FR-003, FR-085, SC-006 | 0 cambios en la caché y 0 en el grafo | `arbol` antes y después, con `cmp` | `ci:internal/app/testdata/script/h23-cita-sin-efectos.txtar` |
| FR-061, FR-086, SC-007 | 1 respuesta que cuenta → `fallo`; 0 → se cumple; 1 en un modo y 0 en el otro → `fallo` | sesiones sintéticas, por respuesta y por informe | `ci:internal/evals/sentencias_test.go:TestCitaSinDocumento`, `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia` |
| FR-055, FR-063 | 4 umbrales en `jurisprudencia`, 0 de juez y 0 de duración; 12 en `boe-legislacion` y 0 en `legal-core`; 3 skills en la matriz; el tope ≥ el peor caso, 20 341 s | el informe con las evals del repositorio; la definición del job | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |
| FR-053, SC-008 | los 2 353 bytes del fragmento en 2 de 2 preguntas | comparación byte a byte con el fichero de la evidencia | `ci:internal/evals/conjunto_test.go:TestPreguntasConElFragmento` |
| FR-056 | 6 de 6 evals, cada una con lo que espera | las reglas del conjunto sobre la carpeta | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |

SC-009, en lo demás, es `make ci` entero. SC-002 y SC-010 los mide una persona o la revisión final: no tienen fila.
FR-045 y FR-048 no tienen control en este hito (FR-065), y están en `gates/supuestos.md` como supuestos de alcance.

## Uso, de fuera adentro (criterio de uso, ADR 0028)

El detalle, con ejemplos y bytes medidos (research M1 a M3), está en contracts/skill-jurisprudencia.md §3,
contracts/applet-cita.md §10 y contracts/evals-jurisprudencia.md §8. Cada operación se cuenta en sus dos sentidos:
lo que el modelo escribe para pedirla y lo que recibe.

- **La respuesta de la skill** (la persona; una por pregunta): por cada sentencia que no está delante, una línea de
  60 a 150 bytes y su consulta, con una dirección y de una a tres casillas; por cada documento cotejado, una cita de
  unos 70 caracteres. La línea deja de darse cuando la persona trae el documento y se coteja como el pedido.
- **La consulta preparada** (la skill; una invocación por sentencia que falta, o por pregunta por materia): el
  modelo escribe la referencia o los términos, menos de 80 bytes con la orden entera, y recibe de 389 a 572 bytes;
  cinco sentencias, unos 2,9 kB.
- **El cotejo** (la skill; una invocación por documento traído): el modelo escribe la ficha y nada más —316 bytes en
  la del fragmento, once líneas; con la referencia y la orden o la llamada alrededor, unos 400—, y recibe de 559 a
  1 005 bytes. Ni lo uno ni lo otro crece con el tamaño del documento: la ficha es el mismo encabezamiento, con las
  mismas etiquetas, mida lo que mida la sentencia (research S5), y del texto solo sale la ficha. El hallazgo, uno
  como mucho, deja de darse con el documento pedido.
- **Las dos herramientas** (el agente; una vez por conexión): 5,4 y 6,5 kB de esquemas con las globales.
- **`SKILL.md`** (el modelo; una vez por conversación en que se activa): 195 líneas, 13,6 kB.
- **`umbrales` del informe** (el job, el informe final, la persona; una vez por job): cuatro elementos de 230 a 330
  bytes; cada job los mide de nuevo.

**Nada crece con lo acumulado, ni con el documento.** El applet no guarda nada: tras meses de uso diario, con cientos
de sentencias preparadas y cotejadas, la invocación siguiente devuelve los mismos bytes que la primera, porque solo
depende de sus argumentos y de su texto. Lo único que podía crecer con algo que no es la pregunta era la entrada del
cotejo, que con la herramienta escribe el modelo: con el texto entero eran 2 353 bytes ya en el fragmento —el
documento completo de esa sentencia es un PDF de 175 955 bytes, cuyo texto no se ha medido—, y por eso el paso 3 de
la skill pasa la ficha (research D17, V41). No hay ninguna señal que se repita sin información nueva: `no-se-deduce`
no llega a la respuesta, y la cobertura del Tribunal Constitucional responde a su propia consulta y no va en ninguna
otra.

## Decisiones

- **El dominio va a `internal/core/{ids,cita}`, y no a un adaptador de fuente** (D1, D2).
- **Un argumento escrito con valor vacío no es un argumento no dado: es el error de su requisito** (D3). Los
  argumentos de los dos verbos son campos puntero, que el analizador deja en `nil` si no se escriben (research V38):
  `ECLI:ES:TS:2023:3144 --texto ""`, `… --roj ""` y `cita cotejar --documento ""` con un documento en la entrada
  terminan con 2 (FR-006, FR-015, FR-020), como orden y como llamada (V40). `--describe`, las herramientas y la tabla
  de comandos no cambian (V39). Anotada en `gates/supuestos.md`.
- **En `data`, lo que no aplica no está**, y `cobertura` tampoco: solo va con el ECLI del Tribunal Constitucional,
  que es lo único que FR-014 declara. Las listas van siempre (D4, D7).
- **La skill pasa a `cita cotejar` la ficha, no el documento** (D17): es lo que dice la entrada del hito, lo único
  que la operación lee y lo que hace que la entrada del cotejo no crezca con la sentencia.
- **`url` identifica lo que abre la persona o el documento, por su huella; `fecha_consulta` es la del reloj** (D6).
- **La entrada estándar la entrega el kernel** (D12): es lo que permite cumplir a la vez FR-020 y FR-030.
- **`--describe` anota las banderas propias, y la tabla las escribe** (D14). El spec no lo preveía; sin ello la tabla
  generada de la skill daría órdenes que no existen. Cambia `schemas/instalacion.json`.
- **`jurisprudencia` sin `references/`, y sin la línea del BOE** (D15, D16): se corrigen dos supuestos del arnés.
- **El formato de eval gana lo que piden las seis evals y nada más; el informe por sesión no gana claves** (D18).
- **La salida de una operación es cada sobre de la sesión, línea a línea** (D19).
- **Tienen umbrales de respuestas las skills con juez o con `sentencias`** (D20).
- **Concurrencia 1 y el tope de hoy** (D22).
- **Que nada cambia en la caché ni en el grafo lo comprueba un guion** (D25), no un test que abra SQLite.
- Las demás, con su alternativa rechazada, en research D1-D30.

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| `ids.ECLI`, `ids.ROJ`, sus analizadores y `ECLI.Organo` | FR-005, FR-006, FR-014 |
| `ECLI.ROJ`, `ROJ.ECLI`, `ROJ.Numero` | FR-012, FR-023, FR-025 |
| `cita.Referencia` y su construcción desde los cuatro argumentos | FR-005, FR-006, FR-024 |
| Los argumentos de los dos verbos como campos puntero: dado, también vacío, o no dado | FR-006, FR-015, FR-020, FR-026 |
| `cita.Consulta`, `Equivalente`, `Casilla`; `Preparar` | FR-010, FR-011, FR-012 |
| `cita.Cobertura`, solo en la consulta del ECLI de órgano `TC` | FR-014 |
| `PrepararTexto` y la codificación del segmento | FR-013, FR-015 |
| `cita.Ficha`, `LeerFicha` | FR-021, FR-022, FR-026 |
| `cita.Cotejo`, `Hallazgo`, `Diferencia`; `Cotejar` | FR-022 a FR-025 |
| El applet `cita` y sus dos verbos, con su firma del sobre | FR-001, FR-004 |
| El argumento `documento`, `cli.Literal` tras su puntero | FR-004, FR-020 |
| `Registro.LeerDe`, `Despacho.Entrada`, la interfaz `lector` | FR-020, FR-026, FR-030 |
| La anotación `x-banderas` y las banderas en `sintaxisDeLaOrden` | FR-040 (la tabla generada, con cada orden) |
| `schemas/cita.json` y su fila en la tabla de esquemas | FR-004, FR-088 |
| Los once golden | FR-080, FR-081 |
| Los tres fuzz y su corpus | FR-083 |
| `internal/core/cita`, `internal/core/ids` y `app/cita.go` en la garantía sin red | FR-002, FR-084 |
| El guion `cita-sin-efectos` | FR-003, FR-085 |
| `skills/jurisprudencia/SKILL.md` | FR-040 a FR-049 |
| El arnés de `TestSkillsDelRepositorio` sin referencias; `linea-sin-consulta` sobre dos skills | FR-040, FR-088 |
| `Eval.Sentencias` y las dos formas de comando, con su esquema | FR-050 a FR-052 |
| El juicio de `sentencias` y de los comandos de `cita` | FR-051, FR-087 |
| El reconocimiento de citas, línea, ECLI y sobres | FR-051, FR-061 |
| `ReglasDeJurisprudencia`; `TestPreguntasConElFragmento` | FR-056; FR-053 |
| `cita_sin_documento` y `sin_activar` sin juez; el motivo con sus respuestas | FR-060 a FR-063 |
| `jurisprudencia` en la matriz de `evals.yml` | FR-063 |
| Las seis evals | FR-050 |
| Ningún cambio en `cmd/empaquetar` ni en `internal/empaquetado`: las herramientas salen del registro y la skill va empotrada (research V36) | FR-031 |
| La forma de comando de `cita` sin respuesta grabada que servir (research V37) | FR-054 |
| README, `docs/JURISPRUDENCIA.md`, `CHANGELOG.md`, `CONTRIBUTING.md`, `internal/evals/doc.go` | FR-070, FR-088 |
| Los dos supuestos de alcance en `gates/supuestos.md` | FR-065 |

Nada del diseño atiende a un estado que no pasa el umbral de materialidad. Un documento falso, una caché o un
`world.db` tocados a mano y una entrada estándar que no se puede leer siguen la regla genérica —defecto
`inesperado`, código 1—, sin caso propio, sin mensaje prometido y sin test.

## Datos externos

Ninguno (research D29). El applet no pide nada y las evals no necesitan grabaciones: ningún manifiesto
`grabaciones.json` ni test `TestGrabar*` nuevo o cambiado, y el paso `grabar_datos` no tiene nada que grabar. El
CENDOJ no se graba: es justo lo que el hito no hace. El único texto de una sentencia que usa el hito es el fragmento
de `evidencias/adr-0036/`, que ya está en `main`, que trajo una persona y que ninguna tarea escribe ni cambia. Ningún
fichero de `data/` se añade ni cambia.

## Orden de implementación (de dentro afuera)

1. **`[aceptacion]`** Los cuatro guiones de `aceptacion/`, sin código de producto y sin ninguna eval.
2. **`[datos]`** `internal/core/ids`: `ecli.go` y `roj.go` con la equivalencia, sus tests, sus dos fuzz con su corpus
   y `TestSuperficieDeIds`.
3. **`[datos]`** `internal/core/cita`: referencia, consulta, ficha y cotejo, con sus tests y `FuzzLeerFicha` con su
   corpus.
4. La entrada estándar en el kernel: `Registro.LeerDe`, `Despacho.Entrada` y la interfaz, con `TestEntradaDeLaOrden`
   sobre un applet del propio test.
5. **`[datos]`** Las banderas propias: la anotación en `internal/cli/describe.go`, la orden de la tabla en
   `internal/skills/comandos.go`, `schemas/instalacion.json` regenerado, y los tests de V16 y V17.
6. **`[datos]`** El applet `cita` en los dos registros: `internal/app/cita.go`, `registro.go`, el binario de e2e,
   `schemas/cita.json`, los once golden, `internal/arch_test.go`, los seis tests de la lista de applets y de
   herramientas, los cuatro guiones e2e del applet y la línea de `CONTRIBUTING.md` que enumera los applets.
7. **`[datos]`** El formato de eval, su juicio y las seis evals: `schemas/eval.yaml.json`, `formato.go`,
   `sentencias.go`, `juzgar.go`, `consultas.go`, `ReglasDeJurisprudencia`, los seis ficheros de
   `evals/jurisprudencia/`, y su carpeta en `TestEvalsDelRepositorio` con `TestPreguntasConElFragmento`.
8. Los umbrales: `umbrales.go` e `informe.go`, con `TestCitaSinDocumento` y `TestUmbralesDeJurisprudencia`.
9. **`[datos]`** La skill: `skills/jurisprudencia/SKILL.md` desde el borrador, con su tabla regenerada; el arnés de
   `TestSkillsDelRepositorio`; `linea-sin-consulta`; y los diecisiete guiones e2e que enumeran las skills.
10. El job: `jurisprudencia` en la matriz de `.github/workflows/evals.yml` y `TestDefinicionDelJob`.
11. La documentación: README («¿Y las sentencias?»), `docs/JURISPRUDENCIA.md`, `CHANGELOG.md` (*Unreleased*),
    `CONTRIBUTING.md` e `internal/evals/doc.go`.

**Obligaciones para `tasks.md`**: una sola tarea `[aceptacion]`, la primera; cada tarea deja `make ci` en verde, y
por eso las de los pasos 6 y 9 llevan en la misma tarea el código, los tests y los guiones e2e que su cambio rompe;
las tareas `[datos]` son las de los pasos 2, 3, 5, 6, 7 y 9, con las rutas de este plan —cada guion e2e, por su
nombre—; cada fila de «Controles de umbral» tiene la tarea que construye su control, con un test que lo ve fallar;
ninguna tarea ni corrección cumple un umbral de FR-060 o de FR-062 rebajándolo, dejándolo en `decide: false`,
sacando evals o un modo de su total, sumando los modos ni cambiando lo que cuenta como salida de una operación
(FR-064); ninguna crea `evals/jurisprudencia/juez/` (FR-055); ninguna escribe en la carpeta de la evidencia ni
escribe de memoria el texto de una sentencia —el fragmento se copia con una orden o se lee del fichero—; ninguna
toca `mcp.go`, `herramientas.go`, `internal/mcp`, `cmd/empaquetar`, `internal/empaquetado`, `internal/httpx`,
`scripts/workflow/`, `skills/boe-legislacion/`, `skills/legal-core/`, `docs/SOURCES.md`, `CLAUDE.md`,
`docs/ROADMAP.md` ni `web/`; ninguna usa la red, abre una sesión con modelo ni ejecuta `make evals`,
`make evals-sondeo`, `scripts/evals*.sh`, `TestEjecucionDelJob` o `TestSondeo`; ninguna publica ni mide en la
plataforma; y los escenarios 10 y 11 del quickstart no son tareas. Las seis evals no van en la primera tarea, sino
en la del paso 7 (research D23).

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada |
|---|---|---|
| El kernel gana la entrega de la entrada estándar (`Registro.LeerDe`, `Despacho.Entrada`, una interfaz sin exportar), que el spec no nombra | FR-020 pide que la orden lea el documento de la entrada y que una llamada de herramienta no la lea nunca, y FR-030, no tocar `mcp`: en el servidor, la entrada es el protocolo | Dar `os.Stdin` al applet como dependencia: una llamada sin `documento` leería el protocolo (research D12) |
| `--describe` gana la anotación `x-banderas`, la tabla de comandos escribe banderas y `schemas/instalacion.json` cambia, que el spec no prevé | Los verbos de `cita` son los primeros con banderas propias en la tabla de una skill, y la tabla generada daría `kitlegal cita preparar [<ecli> [<roj> …]]` (research V14): FR-040 pide la tabla generada con cada orden | Escribir la orden a mano en `SKILL.md`: la tabla es generada y `skills-check` exige que lo sea. Deducir las banderas: `--describe` no las distingue de un argumento de posición opcional |
| Veintiún guiones e2e y once tests de hitos anteriores cambian | Enumeran literalmente los applets, las herramientas o las skills empotradas, y el hito añade un applet, dos herramientas y una skill (research V21) | Ninguna: dejarlos es dejar `make ci` en rojo. El cambio en cada uno es el dato nuevo |
| Tareas `[datos]` con código y tests | El applet, su esquema, sus golden y los guiones que enumeran applets tienen que entrar juntos para que `make ci` quede en verde; igual la skill con los guiones que enumeran skills. Precedente: H7.2 a H24 | Separar datos y código: `make ci` en rojo entre dos tareas |
| El arnés de `TestSkillsDelRepositorio` y la subprueba `linea-sin-consulta` cambian | Suponían que toda skill tiene referencias y lleva la línea del BOE; la tercera no (research V18, V20) | Dar a `jurisprudencia` una referencia que no usa o una línea que habla del BOE |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- `precheck.sh plan`: existen `plan.md` y `research.md`; ninguno conserva marcas de aclaración pendiente; están
  `## Constitution Check`, la línea «Aceptación e2e» y `## Controles de umbral`, con cada fila nombrando requisitos
  definidos en el spec y un control con forma.
- a · Constitution Check: un ítem por principio (I-IX) y por regla de dependencia (R1-R6, y R7 del test).
- b · Dependencias: ninguna nueva.
- c · Reglas de dependencia: dominio sin adaptadores; ni `net/http`, ni SQLite, ni `os.Exit`, ni salida estándar
  nuevos; comprobado en el prototipo (V22).
- d · Errores y códigos: `argumentos` → 2; hallazgo → 0 con `ok` verdadero; lo demás, `inesperado` → 1
  (contracts/applet-cita.md §6).
- e · Tests primero: cuatro guiones congelados, con sus requisitos; tests nuevos, los que cambian, fixtures y tareas
  `[datos]`.
- f · Alcance: nada fuera del spec. Lo que el spec no nombraba y el plan añade para cumplirlo —la entrada por el
  kernel, la anotación de las banderas, el arnés de skills— está en Complexity Tracking, con su requisito.
- g · Sin atajos: ningún `//nolint`, `t.Skip`, TODO ni error silenciado previstos; los riesgos de lint, en S7.
- h · Mejor alternativa: cada decisión con la rechazada (D1-D30); en D3 y D4, la rechazada es la que este plan
  tenía antes de su corrección, con el requisito que incumplía.
- i · Afirmaciones verificadas: V1-V42 con fichero, orden o prototipo; M1-M5 medidas o calculadas, y dicho cuál;
  S1-S7 como supuestos, con lo que pasa si no se cumplen. Lo que D3 dice del analizador está en V38 a V40.
- j · Quickstart ejecutable: órdenes, rutas, banderas y datos reales; los nombres de test, los que fija este plan;
  lo que cada escenario deja en el árbol, declarado; el 10 es del workflow y el 11, de una persona.
- k · Datos externos: ninguno.
- l · Autonomía: ninguna tarea para una persona; el cierre en la plataforma lo hace el workflow.
- m · Uso: «Uso, de fuera adentro» y los tres contratos, con bytes medidos, invocaciones por pregunta, lo que el
  modelo escribe y lo que recibe en cada una, y lo que apaga cada señal.
- n · Proporcionalidad: «Trazabilidad»; sin mecanismos para estados sin vía real, y sin claves de `data` que ningún
  requisito pida ni ninguna skill lea.
- o · Controles de umbral: los cuatro que deciden en el cierre, los de la skill que sigue, y una fila por cada
  umbral medido en `make ci`.
