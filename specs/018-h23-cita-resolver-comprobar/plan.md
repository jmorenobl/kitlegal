# Implementation Plan: H23 · `cita resolver`: comprobar que una sentencia existe, por el formulario del CENDOJ + skill `jurisprudencia`

**Branch**: `018-h23-cita-resolver-comprobar` | **Date**: 2026-10-06 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/018-h23-cita-resolver-comprobar/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de la constitución y
están en [research.md](./research.md) (D1-D32), cada una con su alternativa rechazada. Toda afirmación sobre Go, una
dependencia o el propio repositorio remite a su tabla V, comprobada en local; las medidas, a la tabla M; y lo que no se
pudo comprobar en esta sesión —sin red, sin grabaciones y sin modelo— son los supuestos S1-S11. En esta sesión no se ha
ejecutado ningún test del repositorio ni `make ci`. La corrección del plan (ronda 1) rehízo V17 y D16 y midió lo que
D16 cambia de lo publicado sobre una copia del árbol, fuera del repositorio (research M5).

## Summary

El hito entrega una skill, `jurisprudencia`, y la herramienta que necesita, `kitlegal cita resolver`, en ocho piezas:

1. **El formulario en `internal/httpx`** (D2-D6): una opción que declara la dirección del formulario, los campos en la
   petición y una `Consulta` que guarda la cookie de sesión y no sigue redirecciones. Es el único `POST` del módulo. La
   grabación distingue dos envíos por su cuerpo.
2. **Los identificadores** (D1): ECLI y ROJ en `internal/core/ids`, con su fuzz.
3. **El adaptador `internal/source/cendoj`** (D8-D14, D27): pide la página y envía el formulario, a 5 s y sin
   reintentos; reconoce dos respuestas y ninguna más; filtra por fecha; guarda 30 días los metadatos de la entrega y el
   «no encontrado»; emite un nodo `Resolucion`. Sus respuestas las graba `grabar_datos`.
4. **El applet `cita`** (D15, D18, D19), con un verbo, `resolver`, registrado en producción: de él salen la orden, la
   herramienta `cita_resolver` y `schemas/resolucion.json`. Un ECLI del Tribunal Constitucional se declara fuera de
   cobertura sin pedir nada.
5. **Las banderas propias en `--describe` y en la tabla de comandos** (D16): `cita resolver` es el primer verbo de una
   tabla que las tiene. `--describe` las marca en todo verbo, así que cambia también el documento de los tres verbos
   de `skills`, que ya las tenían, y con él `schemas/instalacion.json`.
6. **La skill** (D29): `skills/jurisprudencia/SKILL.md`, 197 líneas en su prototipo.
7. **Las evals y su umbral** (D20-D25): seis evals, cinco cosas que gana el formato, y `cita_sin_resolver`, que decide
   con 0 en cada modo junto a `sin_activar`, sin modelo.
8. **La comprobación de la fuente, solo a petición** (D26), y la documentación.

No hay ADR nuevo ni dependencia nueva. No se tocan la constitución, `scripts/workflow/`, `docs/SOURCES.md`,
`evidencias/`, `skills/boe-legislacion/` ni `skills/legal-core/`.

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.1`), `CGO_ENABLED=0`, `-trimpath`; bash en
`scripts/verify-sources.sh`; YAML de GitHub Actions.

**Primary Dependencies**: las fijadas, sin ninguna nueva (ADR 0036; research V30): `alecthomas/kong`,
`invopop/jsonschema`, `santhosh-tekuri/jsonschema` v6, `stretchr/testify`, `rogpeppe/go-internal`,
`golang.org/x/time/rate`, `temoto/robotstxt`, el lector de YAML del repositorio y la biblioteca estándar
(`net/http/cookiejar`, `net/url`, `regexp`, `encoding/json`, `time`). El HTML del buscador se lee sin biblioteca de HTML.

**Storage**: la caché SQLite de H3 y el grafo del mundo de H7, sin cambios en ninguno de los dos: una entrada de caché
por consulta, 30 días; un nodo por resolución.

**Testing**: `go test -race` en `make ci`: tablas con `t.Parallel`, goldens de las respuestas grabadas, un sitio local
de prueba (`httpxtest`), reproducción con `httpx.Replay`, fuzz de ECLI y de ROJ, `testscript` contra el binario de e2e,
y sesiones sintéticas para el juicio y el umbral. Fuera de `make ci`: `TestGrabarConsultas` (etiqueta `grabacion`),
`TestVerificarCendoj` (etiqueta `fuentes`) y el job de evals.

**Target Platform**: el binario, las seis plataformas de la release; el job de evals, `ubuntu-24.04`.

**Project Type**: CLI multicall con una skill y su instrumento de medida.

**Performance Goals**: ninguno con umbral. Una consulta que no está en la caché tarda al menos 10 s —tres peticiones a
5 s—, dentro del plazo por omisión de 30 s; desde la caché no pide nada.

**Constraints**: una consulta, como mucho tres peticiones al CENDOJ; ningún test de `make ci`, ninguna eval y ningún
flujo lo consulta; `data` con 10 resoluciones como mucho; `SKILL.md` < 300 líneas; ningún `//nolint` nuevo.

**Scale/Scope**: 7 respuestas del formulario grabadas, más la página; 5 guiones de aceptación; 6 evals y 72 sesiones
por job; 4 umbrales que deciden.

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones; las desviaciones justificadas
están en Complexity Tracking.*

### Principios

| Principio | Cómo lo cumple H23 |
|---|---|
| I · Fuentes públicas y frontera humana | Es el principio que el hito ejerce. El único método que no es GET ni HEAD es el envío del formulario de consulta de una fuente cuya fila de `docs/SOURCES.md`, revisada el 2026-10-03, lo declara, a esa dirección y solo desde `internal/httpx` (FR-030, FR-031; `TestFormularioAdmitido`). Del CENDOJ solo se resuelve una resolución identificada, a sus metadatos, una consulta por resolución, con el agente del proyecto, `robots.txt` y 5 s de ritmo, sin reintentos y sin sortear nada (FR-020, FR-021). `cita resolver` nunca presenta nada: no hay código 6. |
| II · Nada sin cita ni fuente | El sobre lleva `fuente` `cendoj.jurisprudencia`, la `url` del formulario, `fecha_consulta` y `hash`; lo servido de la caché, la fecha de la consulta que lo obtuvo. Lo que no consulta ninguna fuente firma como `kitlegal.cita` (ADR 0006). El nodo del grafo entra con la procedencia del sobre. La skill no cita una sentencia que no se ha resuelto ni escribe un ECLI que no venga de una entrega o de la persona, y `cita_sin_resolver` lo mide con 0. |
| III · Tests primero y offline | Cinco guiones de aceptación congelados, escritos por la primera tarea. Todo test de `make ci` es offline: reproducción, sitio local y sintéticos. Las respuestas reales las graba `grabar_datos`, sin modelo. Cada salida del verbo se valida contra `schemas/resolucion.json` (`TestSalidasDeCitaContraSuEsquema`). `internal/core/ids` gana código con tests y fuzz, por encima del 85 %. |
| IV · Hexagonal y errores tipados | Identificadores en el dominio; el adaptador implementa `core.Source`; el applet se compone en `internal/app`. Los fallos llevan su clase (`schema.ConClase`) y el kernel la traduce: 2, 3, 4 y 5, y 1 para lo inesperado. Ningún `panic`. Reglas de dependencia, abajo. |
| V · Simplicidad y dependencias | Ninguna dependencia nueva. Cada mecanismo se traza a un requisito («Trazabilidad»). Se rechazaron una biblioteca de HTML, cabeceras libres en `Peticion`, un formato 2 de grabación, un almacén de cookies por cliente, un paquete `core/cita` y un fichero de umbrales por skill (research D1-D6, D10, D22). |
| VI · Un binario, convenciones de agente | `cita` es un applet más del multicall, con las ocho banderas globales. `--describe` es de donde salen la herramienta, el esquema publicado y la fila de la tabla; gana decir qué argumento propio es una bandera, con una regla para todo verbo y sin caso por applet: lo dice también el documento de `skills install`, `list` y `doctor`, y `schemas/instalacion.json` se regenera (D16). |
| VII · Grafo y privacidad | Id natural, el ECLI, en `world.db`. El ponente es un dato del nodo `Resolucion`: no se crea ningún nodo `Persona`. Ni texto ni resumen. Nada del asunto. |
| VIII · Skills primero | El hito entrega una skill nueva con sus evals, y el verbo existe porque la skill lo usa. Va a Go lo que exige determinismo: la consulta, el reconocimiento, la comparación de la fecha, la caché y el grafo. El protocolo —qué se resuelve, con qué forma y qué se dice— vive en `SKILL.md`, de menos de 300 líneas, sin `scripts/` y con la tabla generada, que nombra la orden y la herramienta (ADR 0035). |
| IX · Genericidad territorial | El verbo y la skill no tienen dimensión territorial: ningún municipio, comunidad ni órgano tiene un caso propio, y no se validan siglas ni códigos de órgano contra ninguna lista. La cobertura que sí existe, por órgano, se declara dentro de `data` (`cobertura`), y la skill no concluye «no existe» de un «no encontrado». No aplica la matriz territorial. |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | `internal/core/ids` (ECLI, ROJ) y `internal/core/grafo` (una constante de tipo y ocho de datos) solo importan dominio y biblioteca estándar. |
| R2 · Solo `internal/httpx` importa `net/http` | El formulario, la `Consulta` y `net/http/cookiejar` viven en `internal/httpx`; el sitio de prueba, en `internal/httpx/httpxtest`, bajo el mismo árbol. `internal/source/cendoj` e `internal/app` usan `httpx.Peticion`, `httpx.Respuesta` y `httpx.Consulta`, y `net/url` para componer direcciones. `TestArquitectura` y `depguard`, sin cambios. |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | El adaptador usa el puerto `core.Cache`; el applet abre la caché con `cache.New`, como `boe`. `TestNiResumenNiPagina` lee los ficheros de la caché y del grafo como bytes, sin SQL. |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Ningún `os.Exit` nuevo. |
| R5 · Solo `render` escribe en stdout | El applet devuelve `Resultado` con `Datos` y `Legible`; no escribe en ningún descriptor. Los eventos, por el registrador del kernel. |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios en `internal/graph`: las operaciones viajan en el `Resultado`. |
| R7 · Solo `internal/mcp/**` importa el SDK de MCP | Sin cambios: `cita_resolver` sale del registro. |

### Gates

- **Capa 1**, todo en `make ci` o en el workflow: suite congelada con «rojo primero»; `schema-check` con
  `schemas/resolucion.json` y con `schemas/instalacion.json` regenerado; `skills-check` con la skill nueva;
  `TestArquitectura`; validación de cada salida contra su esquema; códigos 2, 3, 4 y 5 forzados con grabaciones y
  sintéticos (el 6 no es de este verbo); tests offline contra `testdata/` grabado; citas de las evals comparadas por
  identificador; fila revisada en `docs/SOURCES.md` de `main` para la fuente que se graba (V28); guardián de diff con
  `testdata/` y `schemas/` solo en tareas `[datos]`; «Controles de umbral» con requisitos del spec.
- **Capa 2**: los jueces del plan, de las tareas y de la revisión final, que revisan además el adaptador nuevo; la
  activación de la skill, por el job de evals. Ningún juez con modelo sobre las respuestas (FR-076).
- **Capa 3**, al leer el informe final: el adaptador nuevo bajo `internal/source/`; las grabaciones y los goldens;
  `schemas/resolucion.json`, nuevo, y los dos ficheros existentes de `schemas/` que cambian: `schemas/eval.yaml.json` y
  `schemas/instalacion.json`, que gana siete líneas `title` en la `entrada` de `doctor`, `install` y `list` (D16); los
  guiones de hitos anteriores cuyas listas y recuentos cambian; la subprueba `linea-sin-consulta` de H21, que deja de
  exigir la línea del BOE a toda skill (research D33); los supuestos del run; la lectura humana de SC-008; la fusión.
  **Ninguna pausa a mitad del run.**

## Project Structure

### Documentation (this feature)

```text
specs/018-h23-cita-resolver-comprobar/
├── plan.md                  # este fichero
├── research.md              # V1-V39, M1-M6, S1-S11, D1-D33
├── data-model.md            # identificadores, referencia, entrega, caché, grafo, formulario, formato de eval
├── quickstart.md            # §0-§10, sin red; §11, la fuente real; §12, el cierre
├── contracts/
│   ├── cita-resolver.md                    # gramática, data con sus bytes, códigos, caché, grafo, herramienta, uso
│   ├── httpx-formulario.md                 # ConFormulario, Campos, Consulta, grabación, tests, httpxtest
│   ├── fuente-cendoj-y-grabacion.md        # la fuente y su fila, la consulta, lo reconocido, manifiesto, grabación, verify-sources
│   ├── evals-jurisprudencia.md             # formato, las seis evals, la caché de la sesión, el umbral, el job
│   ├── skill-jurisprudencia.md             # qué entrega la skill, trazado a sus requisitos, y sus controles
│   └── skill-jurisprudencia.prototipo.md   # el texto propuesto de SKILL.md, entero
├── checklists/              # del spec
├── aceptacion/              # la escribe la tarea [aceptacion]: cinco guiones, congelados
└── tasks.md                 # la escribe /speckit-tasks
```

### Source Code (repository root)

```text
internal/core/ids/
├── ecli.go, roj.go                      # nuevos: AnalizarECLI, AnalizarROJ (+ _test.go con FuzzECLI y FuzzROJ)
└── testdata/fuzz/FuzzECLI, FuzzROJ      # corpus mínimo [datos]
internal/core/grafo/vocabulario.go       # TipoResolucion y las claves de sus datos
internal/httpx/
├── formulario.go                        # nuevo: ConFormulario, Consulta, el envío y sus dos cabeceras
├── cliente.go, peticion.go, transporte.go   # Peticion.Campos; la admisión del método; Pedir compartido con la consulta
├── grabar.go, reproducir.go, nombre.go  # el cuerpo en el nombre, en el fichero y en el emparejamiento
├── doc.go                               # la excepción, dicha
└── httpxtest/sitio.go                   # nuevo: el sitio local de prueba
internal/source/cendoj/                  # nuevo
├── doc.go, terminos.go, direcciones.go  # nombre, ritmo, términos, base, página, formulario, OpcionesDeRed
├── referencia.go, consulta.go           # NuevaReferencia; ConsultaResolver; los campos de cada forma
├── lectura.go                           # clasificar y leerResultados
├── fuente.go, entradas.go, errores.go   # Fuente (core.Source), la caché, los fallos con su clase
├── salida.go, grafo.go                  # Resolucion, Entrega, Cobertura; el nodo
├── grabacion_test.go                    # //go:build grabacion: TestGrabarConsultas
├── casos_test.go                        # el manifiesto, grabarConsultas, TestGrabacionRechazaLoNoReconocido
└── testdata/                            # grabaciones.json [datos]; cendoj.jurisprudencia/ (grabar_datos); golden/ [datos]
internal/app/
├── cita.go                              # nuevo: AppletCita, DependenciasDeCita, el verbo, el caso del Constitucional, Legible
├── registro.go                          # AppletCita en RegistroDeProduccion
├── ejemplo/kitlegal-e2e/main.go         # cita en reproducción
├── fuentes_red_test.go, fuentes_test.go # TestVerificarCendoj; su gemelo sin red; el control de FR-091
└── e2e_test.go, esquemas_test.go, skills_test.go, registro_test.go   # la copia de las grabaciones; resolucion.json; listas
internal/cli/describe.go                 # title en las banderas propias de todo verbo
internal/skills/comandos.go              # la fila con banderas propias
internal/evals/
├── sentencias.go                        # nuevo: ExtraerSentencias, ExtraerSentenciaNoComprobada, ECLISinResolver
├── formato.go, consultas.go, preparar.go, grabaciones.go   # la forma de resolución; las consultas de cita; su reproducción
├── juzgar.go, conjunto.go               # lo esperado nuevo; ReglasDeJurisprudencia
├── informe.go, umbrales.go              # los umbrales de las respuestas sin juez; cita_sin_resolver y su motivo
└── definicion.go, doc.go                # la skill nueva en la matriz
schemas/resolucion.json                  # nuevo [datos]
schemas/instalacion.json                 # title en las banderas propias de los tres verbos de skills [datos]
schemas/eval.yaml.json                   # la forma de resolución y cuatro claves [datos]
skills/jurisprudencia/SKILL.md           # nuevo
evals/jurisprudencia/01-…06-….yaml       # nuevas
scripts/verify-sources.sh, Makefile      # la fuente como argumento; FUENTE
.github/workflows/evals.yml              # jurisprudencia en la matriz
cmd/kitlegal/main_test.go, internal/app/ejemplo/kitlegal-e2e/main_test.go        # lista de applets
internal/app/herramientas_test.go, mcp_test.go                                   # 11 y 14 herramientas; cita_resolver
internal/app/testdata/script/argumentos.txtar                                    # lista de applets [datos]
internal/app/testdata/script/h21-mcp-{herramientas,proceso,protocolo}.txtar      # 14 herramientas [datos]
internal/app/testdata/script/{17 guiones de skills}                              # tres skills instaladas [datos]
internal/skills/testdata/script/instalar{,-sin-gobin,-de-nuevo}.txtar            # tres skills instaladas [datos]
internal/evals/conjunto_test.go                                                  # linea-sin-consulta, acotada (D33)
internal/empaquetado/ejecutar_test.go, snapshot_piezas_test.go                   # cita_resolver entre las herramientas exigidas
internal/arch_test.go                    # el binario no enlaza httpxtest
README.md, docs/JURISPRUDENCIA.md, CONTRIBUTING.md, CHANGELOG.md
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`. Un adaptador por fuente en `internal/source/cendoj`;
los identificadores en `internal/core/ids`; el applet en `internal/app`; el formulario dentro de `internal/httpx`. Un
solo paquete nuevo que no es de la arquitectura objetivo, `internal/httpx/httpxtest`, de apoyo a tests (Complexity
Tracking). `internal/core/cita` no se crea: es de H8.

## Aceptación e2e

**Aceptación e2e:** cinco guiones `testscript`, que la tarea `[aceptacion]` escribe en
`specs/018-h23-cita-resolver-comprobar/aceptacion/` y quedan congelados. Cada uno empieza por la precondición
`exec kitlegal --help` + `stdout '^\s+cita\s'`, que hoy falla en su aserción y no en una orden desconocida (V31). Usan
las grabaciones de `reproduccion/cendoj.jurisprudencia`, con los nombres de
[contracts/fuente-cendoj-y-grabacion.md §4](./contracts/fuente-cendoj-y-grabacion.md), y su `HOME` y su caché son los
del directorio de trabajo del guion.

| Guion | Historia | Casos y código | FR / SC |
|---|---|---|---|
| `cita-formas.txtar` | US1 (1 a 4, 6) | por ECLI, por `--roj`, por `--resolucion` con `--fecha` y por `--roj` con su fecha: 0, con el mismo `data`; la salida legible | FR-001, FR-002, FR-006, FR-010, FR-011, FR-018; SC-002 |
| `cita-no-comprobada.txtar` | US2 | ECLI que no existe y `--roj` con otra fecha: 3, sin ningún dato de la sentencia de 4 de julio; ECLI mal formado, de otro país y `--resolucion` sin `--fecha`: 2; ECLI del Tribunal Constitucional: 0, con su `cobertura` | FR-003, FR-005, FR-011, FR-012, FR-015; SC-002 |
| `cita-bloqueo.txtar` | US5 (1, 2) | una página que no se reconoce como respuesta del formulario y un 403 en la página, los dos escritos en el guion y copiados sobre su grabación: 5 y `limite-o-tos`; nada queda en la caché | FR-016, FR-022, FR-094; SC-002 |
| `cita-cache.txtar` | US6 (1, 2, 4, 5, 6) | la misma consulta con `--offline`: 0 y el mismo sobre; sin nada en la caché: 4; el «no encontrado» repetido con `--offline`: 3 y el mismo sobre; `graph show <ECLI>`: un nodo `Resolucion`; con `--no-graph`, ninguno | FR-017, FR-040 a FR-043; SC-002 |
| `cita-herramienta.txtar` | US1 (5) | el servidor anuncia `cita_resolver`; llamada con `ecli`: el sobre de la orden; con un ECLI mal formado: error con `argumentos` | FR-050 |

US3, US7, US8 y US10 no son comportamiento del binario —son el texto de `SKILL.md`, el juicio y el informe del job, y
la documentación—: las fijan en `make ci` `skills-check`, `TestEvalsDelRepositorio`, `TestJuzgarLasSentencias`,
`TestUmbralDeCitaSinResolver` y `TestUmbralesDeJurisprudencia`, y en el cierre, el job de evals (SC-001). US4 es de
`internal/httpx`, que no tiene orden propia: sus cinco tests de [contracts/httpx-formulario.md §6](./contracts/httpx-formulario.md).
US9 es un guion de shell y un test de grabación: `TestElCendojSoloSeCompruebaAPeticion` y
`TestGrabacionRechazaLoNoReconocido`. US5.3 (el sitio que no responde) y el recuento de peticiones no caben en un guion
con reproducción: `TestPeticionesDeUnaConsulta`. US6.3 (ni resumen ni página), `TestNiResumenNiPagina`.

Las seis evals no las escribe la tarea `[aceptacion]`: sus claves no existen en el formato hasta la tarea que cambia
`schemas/eval.yaml.json`, y antes dejarían `make ci` en rojo; entran con la skill (paso 12), como en H21.

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

`verify-sources` pasa `$(FUENTE)` al guion, sin valor por defecto. `test`, `schema-check` y `skills-check` llevan más
dentro. Ningún objetivo nuevo; `ci` no cambia.

### Tests nuevos

| Test | Fichero | Cubre |
|---|---|---|
| `TestAnalizarECLI`, `TestAnalizarROJ`, `FuzzECLI`, `FuzzROJ` | `internal/core/ids/ecli_test.go`, `roj_test.go` | FR-003, FR-004, FR-113 |
| `TestFormularioAdmitido`, `TestEnvioDelFormulario`, `TestCookiesDeLaConsulta`, `TestConsultaSinRedirecciones` | `internal/httpx/formulario_test.go` | FR-030 a FR-033, FR-112; SC-006 |
| `TestGrabacionDeFormularios` | `internal/httpx/grabar_test.go` | FR-034, FR-112 |
| `TestSitio` | `internal/httpx/httpxtest/sitio_test.go` | el sitio de prueba apunta lo que recibe |
| `TestNuevaReferencia`, `TestCamposDeLaConsulta` | `internal/source/cendoj/referencia_test.go`, `consulta_test.go` | FR-002, FR-005, FR-006 |
| `TestClasificar`, `TestLecturaDeLasGrabaciones`, `TestLaSentenciaConocida` | `internal/source/cendoj/lectura_test.go` | FR-022, FR-110 |
| `TestPeticionesDeUnaConsulta`, `TestPaginaCompleta`, `TestFiltroPorFecha`, `TestVigenciaDeLaCache`, `TestFallosDeLaFuente` | `internal/source/cendoj/fuente_test.go` | FR-011 a FR-017, FR-020, FR-021, FR-040, FR-041; SC-004 |
| `TestOperacionesDeGrafo` | `internal/source/cendoj/grafo_test.go` | FR-042: un nodo por resolución, sin aristas ni texto; un fallo no emite nada |
| `TestCitaEntregaAlGrafo` | `internal/app/cita_test.go` | FR-042, FR-043, FR-116: ninguna operación sin procedencia; aplicar dos veces no duplica; `graph check` sin hallazgos del nodo |
| `TestFuenteCoincideConSources` | `internal/source/cendoj/terminos_test.go` | FR-023, FR-114 |
| `TestGrabacionesCompletas`, `TestGrabacionRechazaLoNoReconocido` | `internal/source/cendoj/casos_test.go` | FR-092, FR-093, FR-094 |
| `TestAppletCita`, `TestSalidasDeCitaContraSuEsquema`, `TestNiResumenNiPagina`, `TestCitaEnEnsayo` | `internal/app/cita_test.go` | FR-001, FR-015, FR-018, FR-019, FR-044, FR-111; SC-005 |
| `TestVerificacionDelCendojDetectaCambios`, `TestElCendojSoloSeCompruebaAPeticion` | `internal/app/fuentes_test.go` | FR-090, FR-091; SC-003 |
| `TestExtraerSentencias`, `TestExtraerSentenciaNoComprobada`, `TestECLISinResolver` | `internal/evals/sentencias_test.go` | FR-071, FR-080, FR-081 |
| `TestJuzgarLasSentencias` | `internal/evals/juzgar_test.go` | FR-070, FR-071 |
| `TestUmbralDeCitaSinResolver`, `TestUmbralesDeJurisprudencia` | `internal/evals/umbrales_test.go` | FR-080 a FR-083, FR-115; SC-007 |
| `TestSkillsQueLlevanLaLineaSinConsulta` | `internal/evals/conjunto_test.go` | la selección de la subprueba `linea-sin-consulta` (research D33), sobre conjuntos escritos por el test en `t.TempDir()`: la skill cuyo conjunto lleva la eval sin binario ni servidor entra, y la que no la lleva, no (FR-063, FR-069) |

### Tests existentes que cambian

| Test | Cambio |
|---|---|
| `TestPedirRechazaMetodo`, `TestOpcionesInvalidas`, `TestReplayRechazaOpciones`, `TestSuperficieExportada`, `TestNombreDeGrabacion` (`internal/httpx`) | el `POST` fuera de una consulta sigue rechazado; `ConFormulario` inválida; `Replay` la admite; la superficie nueva; los nombres con cuerpo |
| `TestRegistroDeProduccion`, `TestPuntoDeEntrada` (`cmd/kitlegal`), `TestRegistroDeE2E` (`kitlegal-e2e`) | `cita` en la lista de applets |
| `TestDescripcionDeVerbo`, `TestRenderizarTabla` (`internal/skills`) | un verbo con banderas propias: se leen de `title` y se escriben en su fila |
| `TestEsquemasPublicados`, `TestEsquemasCubrenTodosLosVerbos` | `resolucion.json` en `ficherosDeEsquemas` (paso 10). Antes, en el paso 5, `TestEsquemasPublicados` no cambia de código, pero sí lo que compara: las partes `doctor`, `install` y `list` de `schemas/instalacion.json`, que esa misma tarea regenera (research M5) |
| `TestSkillsDelRepositorio`, `TestOrdenesDeLasSkillsEmpotradas`, `TestTablaDeComandosCoincideConLaGramatica` | `jurisprudencia` en `skillsExigidas`; la sintaxis con banderas propias —la de `cita resolver` y, en la tabla que ese último test genera para todo el registro, las de los tres verbos de `skills`—, que la invocación mínima omite. En `TestSkillsDelRepositorio`, dos ayudantes dejan de dar por hecho que toda skill declara referencias: `metadataDeKitlegal` escribe la línea `kitlegal-referencias` solo si las hay, y `probarRegenerarDosVeces` crea `references/` antes de escribir en ella (research D33, M6); ningún caso se retira |
| `TestHerramientasDelServidor`, `TestHerramientasAnunciadas` | una herramienta más: los dos recuentos de `TestHerramientasDelServidor` pasan de 10 y 13 a 11 y 14 (V34); su esquema de entrada sin `title` |
| los de `internal/cli/describe_test.go` e `internal/cli/herramienta_test.go` | el `title` de una bandera propia en `--describe`, con un verbo de test; su ausencia en los esquemas de la herramienta: el subtest `los dos esquemas son las partes del documento de --describe` de `TestEsquemasDeHerramienta` pasa a comparar la entrada sin esa marca (research M5) |
| `TestEntregaDelHito` | `Setup` copia también las grabaciones del CENDOJ. De sus guiones de hitos anteriores cambian veintiuno: en el paso 10, `argumentos` y los tres de H21 cuya salida esperada cuenta las herramientas —`h21-mcp-herramientas`, `h21-mcp-proceso` y `h21-mcp-protocolo`— (V34); en el paso 13, los diecisiete que enumeran o cuentan las skills instaladas (V35) |
| `TestInstalacion` (`internal/skills`, etiqueta `integration`; en `make ci` por `test-integration`) | tres de sus cuatro guiones de `internal/skills/testdata/script/` —`instalar`, `instalar-sin-gobin` e `instalar-de-nuevo`—, con la tercera skill en lo que instala `make install` (V35); su código no cambia |
| `TestArquitectura`, `TestElBinarioNoEnlazaLosEjemplos` | el binario distribuido no enlaza `internal/httpx/httpxtest` |
| `TestEjecutar` (`internal/empaquetado/ejecutar_test.go`) y la lista `herramientasDeHoy`, también la de `snapshot_piezas_test.go` | `cita_resolver` entre las herramientas que el manifiesto tiene que listar: datos de prueba, sin tocar el paso que empaqueta (FR-051) |
| `TestLeerEval`, `TestFormaDelComando`, `TestEsquemaDeEval`, `TestGramaticasCoincidenConBoe` | la forma de resolución y las cuatro claves; los patrones atados a sus analizadores |
| `TestPrepararYComprobar` | las consultas de `cita`, con el «no encontrado» que termina con 3 |
| `TestConjuntoDeEvals`, `TestEvalsDelRepositorio` | `ReglasDeJurisprudencia`; subpruebas `conjunto-jurisprudencia`, `texto-de-la-sentencia`, `formas-de-jurisprudencia` y `direcciones-de-la-skill`. La subprueba `linea-sin-consulta`, de H21, deja de recorrer todas las skills de `skills/`: exige la línea `⚠ SIN CONSULTA AL BOE: …` a las skills cuyo conjunto de `evals/<skill>/` lleva una eval con `sin_binario_ni_servidor: true`, que es la eval que esa línea responde, con `boe-legislacion` y `legal-core` como premisa para que no pase en vacío (research D33). `jurisprudencia`, cuyo conjunto no la lleva, responde sin herramienta ni binario con su línea `⚠ SENTENCIA NO COMPROBADA:` (FR-063), que fija `formas-de-jurisprudencia` |
| `TestUmbralesDelInforme`, `TestDefinicionDelJob` | `boe-legislacion` sigue con doce umbrales y `legal-core` con ninguno; `jurisprudencia` en la matriz, sin objetivo de duración |

### Puntos de entrada fuera de `make ci`

`TestGrabarConsultas` (`grabacion`), que solo ejecuta `grabar_datos`; `TestVerificarCendoj` (`fuentes`), que solo lanza
una persona con `make verify-sources FUENTE=cendoj`; y `TestEjecucionDelJob` (`evals`), en el job. `golangci-lint` los
lintea. Ninguna tarea los ejecuta.

### Fixtures, `testdata/` y `schemas/` (tareas `[datos]`)

- `internal/core/ids/testdata/fuzz/FuzzECLI/` y `FuzzROJ/`: corpus mínimo.
- `internal/source/cendoj/testdata/grabaciones.json`, con `grabacion_test.go` y `casos_test.go`. Tras su commit,
  `grabar_datos` deja `testdata/cendoj.jurisprudencia/` (nueve ficheros).
- `internal/source/cendoj/testdata/golden/`: lo que la lectura da de las grabaciones 1 a 4, escrito con
  `go test -run '^TestLecturaDeLasGrabaciones$' ./internal/source/cendoj/ -args -actualizar-golden`.
- `schemas/instalacion.json`, un fichero existente. Lo regenera
  `go test -count=1 -run '^TestEsquemasPublicados$' ./internal/app/ -args -actualizar-esquemas` en la tarea `[datos]`
  del paso 5, la misma que cambia `--describe`: gana siete líneas y no pierde ninguna —`"title": "--dir"` y
  `"title": "--global"` en la `entrada` de `doctor`, `install` y `list`, y `"title": "--host"` en la de `install`—, de
  15 977 a 16 211 bytes (research M5). Ningún otro fichero de `schemas/` cambia en ese paso.
- `schemas/resolucion.json`, escrito por `TestEsquemasPublicados -actualizar-esquemas`, con el registro de `cita` y las
  listas literales de applets y de herramientas, en cuatro guiones de `internal/app/testdata/script/` (V34):
  `argumentos.txtar` (las dos líneas de la lista de applets); `h21-mcp-herramientas.txtar` (su lista, que pasa a 14
  líneas, el `1 herramientas 14` de su salida esperada y los comentarios que cuentan las herramientas); y el recuento
  `1 herramientas 14` de `h21-mcp-proceso.txtar` (una línea) y de `h21-mcp-protocolo.txtar` (tres).
- `schemas/eval.yaml.json`, con los casos de `formato_test.go` que lo fijan.
- Los 20 guiones que enumeran o cuentan las skills instaladas (V35), con `jurisprudencia` entre las otras dos en cada
  enumeración, una unidad más en cada recuento que depende de cuántas hay y «tres» donde un comentario dice «dos»;
  nada más de ellos cambia. Son 17 de `internal/app/testdata/script/` —los 15 que las nombran juntas, y
  `h19-skills-aviso.txtar` y `h19-skills-aviso-sin-aviso.txtar`, que cuentan una versión por skill en el manifiesto— y
  3 de los 4 de `internal/skills/testdata/script/`, que ejecuta `TestInstalacion`: `instalar.txtar`,
  `instalar-sin-gobin.txtar` e `instalar-de-nuevo.txtar`.
- Los cinco guiones activados (`h23-…`), que copia el workflow tras el bucle de tareas.

Los sintéticos del CAPTCHA y del 403 no son `testdata/`: van escritos en su test o en su guion (FR-094).

### CI

`.github/workflows/evals.yml`: `jurisprudencia` en `matrix.skill`, con `concurrencia: 1` y `objetivo_de_duracion: 0`;
`timeout-minutes` no cambia (D25). `nightly.yml` y `ci.yml` no cambian. El job de evals lo lanza el workflow en la
propuesta de cambio, tras la revisión final.

## Controles de umbral

Las cuatro filas del cierre que pide FR-084, una por los umbrales de `boe-legislacion` que SC-001 exige que sigan
cumplidos, y una por cada umbral que se mide en `make ci`. `<modelo>` es hoy `claude-sonnet-5-5`.

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| FR-080, SC-001 (modo orden) | 0 respuestas del modelo que decide con una cita cuyo ECLI no se entregó en su sesión, o con un ECLI suelto que no viene de una entrega ni de la pregunta | umbral del informe con `decide: true`: por encima, motivo con el umbral y la sesión, veredicto `fallo` y `evals (jurisprudencia)` en rojo, que el cierre cuenta | `evals:jurisprudencia:cita_sin_resolver:claude-sonnet-5-5:orden` |
| FR-080, SC-001 (modo herramienta) | lo mismo, sobre las respuestas del modo herramienta | ídem | `evals:jurisprudencia:cita_sin_resolver:claude-sonnet-5-5:herramienta` |
| FR-082, SC-001 (modo orden) | 0 respuestas del modelo que decide sin la skill activada | ídem | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden` |
| FR-082, SC-001 (modo herramienta) | 0 | ídem | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta` |
| SC-001 (`boe-legislacion` sigue aprobado) | los diez umbrales que deciden hoy, sin cambios: 0 sin activar y 0 en `afirma_lo_no_leido` por modo, 0 y 0 en la medida del juez, ≤ 900 s por modo en sesiones y en votos | los de H24, que el hito no toca; su job, `evals (boe-legislacion)`, lo cuenta el cierre | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`, `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5:orden`, `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`, `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5:herramienta`, `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`, `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados`, `evals:boe-legislacion:duracion_de_las_sesiones:orden`, `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`, `evals:boe-legislacion:duracion_del_juez:orden`, `evals:boe-legislacion:duracion_del_juez:herramienta` |
| FR-083 | `jurisprudencia` en la matriz; objetivo de duración 0; exactamente 4 umbrales, ninguno de duración; el tope cubre su peor caso | la definición del job y el informe de las seis evals con sesiones sintéticas | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, `ci:internal/evals/umbrales_test.go:TestUmbralesDeJurisprudencia` |
| FR-115, SC-007 | 1 respuesta con un ECLI que no viene de una entrega ni de la pregunta → `fallo`; 0 → se cumple | los cinco casos de sesiones sintéticas | `ci:internal/evals/umbrales_test.go:TestUmbralDeCitaSinResolver` |
| FR-111, SC-002 | 11 casos con su código (0 en cinco, 3 en dos, 2 en dos, 5 en dos) y la repetición con `--offline` con 0 | los guiones `h23-cita-formas`, `h23-cita-no-comprobada`, `h23-cita-bloqueo` y `h23-cita-cache` | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| FR-091, SC-003 | 0 peticiones al CENDOJ desde `make verify-sources` sin fuente y desde cualquier flujo | el guion con un `go` sustituto, `Makefile` y los flujos | `ci:internal/app/fuentes_test.go:TestElCendojSoloSeCompruebaAPeticion` |
| FR-021, SC-004 | por consulta, ≤ 1 envío del formulario, ≤ 1 petición de la página y 0 a otra dirección que `robots.txt`; con el sitio en 5xx, ≤ 1 por dirección | recuento en un sitio de prueba | `ci:internal/source/cendoj/fuente_test.go:TestPeticionesDeUnaConsulta` |
| FR-014, SC-004 | `data` con ≤ 10 resoluciones; una página de 10 lo declara | lista sintética de 10 y las grabaciones de una | `ci:internal/source/cendoj/fuente_test.go:TestPaginaCompleta` |
| FR-040, FR-041 | 30 días: antes, 0 peticiones al repetir; después, 0 servidas de la caché; para la entrega y para el «no encontrado» | reloj de la caché adelantado | `ci:internal/source/cendoj/fuente_test.go:TestVigenciaDeLaCache` |
| FR-044, SC-005 | 0 apariciones del resumen y de la página en la caché y en el grafo; 0 datos de la resolución de otra fecha | los ficheros de la caché y del grafo, leídos como bytes | `ci:internal/app/cita_test.go:TestNiResumenNiPagina` |
| FR-030, FR-031, FR-112, SC-006 | 0 peticiones emitidas con un método que no es GET ni HEAD, salvo el `POST` de una consulta a la dirección declarada | la tabla de admisión contra un servidor de prueba, y el grafo de importaciones | `ci:internal/httpx/formulario_test.go:TestFormularioAdmitido`, `ci:internal/arch_test.go:TestArquitectura` |
| FR-032, FR-112 | 0 cookies de una consulta en la siguiente, en `Cliente.Pedir` y en `robots.txt` | servidor de prueba que da una cookie | `ci:internal/httpx/formulario_test.go:TestCookiesDeLaConsulta` |
| FR-034, FR-112 | 2 envíos con campos distintos, 2 grabaciones; cada uno reproduce la suya | grabar y reproducir en un temporal | `ci:internal/httpx/grabar_test.go:TestGrabacionDeFormularios` |
| FR-093 | con 1 respuesta que no es ninguna de las dos reconocidas, la grabación falla y deja 0 ficheros | reproducción de sintéticos | `ci:internal/source/cendoj/casos_test.go:TestGrabacionRechazaLoNoReconocido` |
| FR-023 | 0 divergencias entre la fila de la fuente y sus constantes | la fila contra `Terms()`, el ritmo, la base, el formulario y sus campos | `ci:internal/source/cendoj/terminos_test.go:TestFuenteCoincideConSources` |
| FR-073 | la pregunta de la eval (d) contiene el fichero de la evidencia: 0 bytes de diferencia | subprueba `texto-de-la-sentencia` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-075 | las 6 evals, cada una con lo que espera | subprueba `conjunto-jurisprudencia` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-060, SC-009 (líneas) | `SKILL.md` de `jurisprudencia` < 300 líneas; tabla igual a la generada | `TestSkillsDelRepositorio`, que falla con 300 o más | `ci:Makefile:skills-check` |
| FR-001, SC-009 (esquema) | 0 diferencias entre `schemas/` y `--describe` | `TestEsquemasPublicados` | `ci:Makefile:schema-check` |

Lo que SC-001 pide además: que `red` esté vacío y que `legal-core` siga aprobado. Ninguno es un elemento de `umbrales`:
una llegada a la red es un motivo del informe, que ya pone su veredicto en `fallo`, y `legal-core` no tiene umbrales y
decide por sus series (ADR 0016); los dos los cuenta el cierre por la comprobación `evals (<skill>)`. SC-008 lo mide una
persona. «Resume» (FR-064) no tiene control en este hito (FR-085): está en `gates/supuestos.md` como no comprobado.

## Uso, de fuera adentro (criterio de uso, ADR 0028)

Detalle, con ejemplos y bytes, en [contracts/cita-resolver.md §10](./contracts/cita-resolver.md),
[contracts/skill-jurisprudencia.md §5](./contracts/skill-jurisprudencia.md) y
[contracts/evals-jurisprudencia.md §7](./contracts/evals-jurisprudencia.md). Resumen, con las medidas de research M3 y M4:

| Salida | Quién la pide y cuántas veces | Tamaño | Cuándo se apaga |
|---|---|---|---|
| La entrega de `cita resolver` | la skill, de una en una: 1 por sentencia dada por su ECLI o su ROJ; 1 o 2 por cita con número y fecha; 0 sin fecha, por materia o del Constitucional | 625 B con una resolución; ≤ 3 495 B con diez. Cinco sentencias: de 5 a 10 consultas, ≤ 6 250 B. Por herramienta, el doble en el mensaje | la acotan la referencia y la página de 10; no crece con lo acumulado |
| El «no encontrado» | la skill: no cita y escribe su línea; antes, el segundo intento si procede | 422 B | se guarda 30 días; después, o por otra forma de la referencia, la consulta vuelve a la fuente |
| `pagina_completa` | la skill y la persona | un booleano | solo es verdadero con una página de 10 |
| La declaración del Tribunal Constitucional | la skill, que la traslada | 297 B | responde a esa consulta; no se guarda |
| Los fallos 4 y 5 | la skill: su línea, sin insistir | `{clase, mensaje}`, menos de 1 kB | en la primera consulta que la fuente responde |
| La cita y la línea `⚠` | quien pregunta: una por sentencia comprobada y una por referencia sin comprobar | 71 y unos 150 caracteres | no se repiten en otra respuesta salvo que se pregunte otra vez |
| El nodo `Resolucion` | una persona, con `graph show <ECLI>`; la skill no lo pide | 318 B de datos; uno por resolución comprobada alguna vez: cientos en meses de uso | no da señales: `graph check` no da hallazgos de él |
| La caché de la fuente | el propio verbo | de 270 a 502 B por consulta distinta de los últimos 30 días | cada entrada vence a los 30 días |
| `umbrales` del informe | el job, el informe final y la persona | 4 elementos de unos 355 B | cada job los mide de nuevo |

Nada del diseño atiende a un estado que no pasa el umbral de materialidad: una caché, un grafo o una grabación tocados
a mano, una respuesta a medio llegar y un cliente con formulario y varios intentos siguen la regla genérica —defecto
`inesperado` y código 1—, sin caso propio (research D31).

## Decisiones

Las treinta y dos, con su alternativa, en [research.md](./research.md). Las que cambian lo que ve quien usa el binario
o lo que hace la skill:

- **El envío lleva `X-Requested-With: XMLHttpRequest`** (D5): es la petición que se probó a mano. La fila de la fuente
  no la nombra.
- **Una redirección del buscador es una respuesta que no se reconoce** (D4): código 5, sin seguirla.
- **`data` es un objeto** con `resoluciones`, `pagina_completa` y `cobertura` (D11), y el sobre del Tribunal
  Constitucional no repite el ECLI pedido.
- **`--describe` marca las banderas propias de todo verbo con `title`** (D16). Cambia lo que emite
  `kitlegal skills install`, `list` y `doctor` con `--describe`, y con ello `schemas/instalacion.json`, que la web sirve
  en `https://kitlegal.es/schemas/instalacion.json`: siete líneas `title` más, 234 bytes, sin cambiar nada de lo que el
  esquema valida. Los demás esquemas publicados de hoy no cambian.
- **La skill resuelve por el ECLI cuando tiene ECLI y ROJ** (D29), y no nombra operadores del buscador que no consten
  en el repositorio.
- **Las evals (b) y (f) esperan los dos intentos del protocolo** (D24): el número con su fecha y, después, el ROJ con
  esa fecha.
- **`evals (jurisprudencia)` abre una sesión a la vez** (D25).
- **Los umbrales de las respuestas existen con juez o con sentencias** (D22).

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| `ids.AnalizarECLI`, `ids.AnalizarROJ` y su fuzz | FR-003, FR-004, FR-113 |
| `grafo.TipoResolucion` y sus claves | FR-042 |
| `httpx.ConFormulario`, `Peticion.Campos` | FR-030, FR-031 |
| `httpx.Consulta` (cookies; sin redirecciones) | FR-032, FR-033, FR-021 |
| `X-Requested-With` en el envío | FR-020 (la petición de la prueba) |
| El cuerpo en el nombre y en el fichero de una grabación | FR-034 |
| `httpxtest.Sitio` | FR-021 y SC-004 (su control cuenta peticiones reales, con `robots.txt`) |
| `cendoj.NuevaReferencia` | FR-002, FR-005 |
| Los campos de cada forma | FR-006, FR-020 |
| `clasificar` y `leerResultados` | FR-022, FR-092 |
| El filtro por fecha y el mensaje | FR-011, FR-012 |
| `pagina_completa` | FR-014 |
| `cobertura` y el caso del applet | FR-015 |
| Las clases de cada fallo | FR-016, FR-017 |
| La entrada de caché, con el «no encontrado» | FR-040, FR-041 |
| `OpcionesDeRed` (un intento) y `OpcionesDeReproduccion` | FR-021; FR-074 y FR-111 (la reproducción no admite reintentos) |
| `IntervaloEntrePeticiones` y el `Ritmo` por dependencias | FR-020, FR-024 |
| `TestFuenteCoincideConSources` | FR-023 |
| `Resultado.Legible` | FR-018 |
| El ensayo | FR-019 |
| `AppletCita` en el registro; `schemas/resolucion.json` | FR-001, FR-050, FR-051 |
| `title` en las banderas propias; la fila con banderas | FR-060 (la tabla nombra la orden como se escribe) |
| `schemas/instalacion.json` regenerado | FR-001, SC-009 (`schema-check` sin diferencias: el documento de los tres verbos de `skills` cambia con la marca) |
| El manifiesto, `TestGrabarConsultas` y la copia desde un temporal | FR-092 a FR-094 |
| El argumento de `verify-sources.sh`, `FUENTE` y `TestVerificarCendoj` | FR-090, FR-091 |
| `cita` en el binario de e2e y la copia de sus grabaciones | FR-111 |
| La forma de resolución, `sentencias`, `sentencia_no_comprobada`, `sin_sentencias`, `direcciones` | FR-070 a FR-072 |
| `Consulta.NoEncontrado` y el applet `cita` en `Preparar` | FR-074 |
| `ReglasDeJurisprudencia` | FR-075 |
| `ECLISinResolver`, el umbral y su motivo | FR-080, FR-081 |
| Los umbrales de las respuestas sin juez | FR-082, FR-083 |
| `jurisprudencia` en la matriz del job | FR-083 |
| `skills/jurisprudencia/SKILL.md` | FR-060 a FR-068 |
| README, `docs/JURISPRUDENCIA.md`, `CONTRIBUTING.md`, `CHANGELOG.md` | FR-095, FR-100 |

Sin mecanismo, por regla del spec: FR-061 (de una en una), FR-064 y FR-067 son reglas de la skill sin control propio;
FR-069 y FR-076 son cosas que no se hacen, y las comprueba el diff.

## Datos externos

Una fuente, `cendoj.jurisprudencia`, con fila en `docs/SOURCES.md` de `main` y «Revisado» 2026-10-03 (V28).

- **Manifiesto**: `internal/source/cendoj/testdata/grabaciones.json`, con `"fuente": "cendoj.jurisprudencia"` y siete
  consultas.
- **Test de grabación**: `TestGrabarConsultas` (`//go:build grabacion`), en `internal/source/cendoj/grabacion_test.go`.
  Pide con `internal/httpx`, al ritmo de la fila: la página y las siete consultas. Deja las respuestas en
  `internal/source/cendoj/testdata/cendoj.jurisprudencia/`. No deja nada en `$KITLEGAL_EVIDENCIAS`: no hay material de
  origen del que se derive ningún fichero de `data/`.
- **Quién graba**: el paso `grabar_datos`, sin modelo, tras el commit de la tarea `[datos]` del manifiesto. Ninguna
  tarea usa la red ni `KITLEGAL_RECORD`.
- **Si el CENDOJ bloquea la grabación**: el test falla, nada se graba, `grabar_datos` lo repite una vez al minuto y
  detiene el run con su dossier (causa mayor). Decide una persona.
- **Ningún fichero de `data/`** cambia en este hito.
- **El texto de la eval (d)** no es una grabación: es `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt`, que ya
  está en `main` y el run no toca; la eval lo copia.
- **Fuera de alcance**: el buscador del Tribunal Constitucional, sin fila en `docs/SOURCES.md`. No se consulta ni se
  graba: la skill solo da su dirección (research S6).

## Orden de implementación (de dentro afuera)

1. **`[aceptacion]`** los cinco guiones en `specs/018-h23-cita-resolver-comprobar/aceptacion/`, cada uno con su
   precondición, en rojo por una aserción.
2. `internal/core/ids`: ECLI y ROJ, con sus tests y su fuzz; **`[datos]`** el corpus.
3. `internal/core/grafo`: `TipoResolucion` y sus claves.
4. `internal/httpx`: `ConFormulario`, `Campos`, `Consulta`, la grabación con cuerpo y `httpxtest`.
5. **`[datos]`** `internal/cli` e `internal/skills`: las banderas propias en `--describe` y en la tabla, probadas con
   un verbo de test; y `schemas/instalacion.json` regenerado, porque la marca cambia el documento de los tres verbos de
   `skills`.
6. `internal/source/cendoj`, primera parte: términos, direcciones, referencia, campos, `clasificar` y opciones.
7. **`[datos]`** el manifiesto, `grabacion_test.go` y `casos_test.go`. **Aquí graba `grabar_datos`.**
8. **`[datos]`** la lectura de la lista con sus goldens, con las grabaciones delante.
9. `internal/source/cendoj`, segunda parte: `Fuente`, la caché, el filtro, los fallos y el grafo.
10. **`[datos]`** el applet `cita` registrado, en producción y en el binario de e2e, con `schemas/resolucion.json`, las
    listas de applets y de herramientas —también el recuento de herramientas de los tres guiones de H21 que lo
    comparan— y la copia de las grabaciones en el arnés.
11. La comprobación de la fuente: `scripts/verify-sources.sh`, `Makefile` y sus tests.
12. **`[datos]`** el formato de eval, el juicio y el umbral, con `schemas/eval.yaml.json`.
13. **`[datos]`** la skill y sus seis evals, con `skillsExigidas`, los 20 guiones que enumeran o cuentan las skills
    instaladas —17 de e2e y 3 de los de `make install`—, las dos comprobaciones que daban por hecho algo que la skill no
    tiene (research D33), la matriz del job y las subpruebas de `TestEvalsDelRepositorio`.
14. La documentación: README, `docs/JURISPRUDENCIA.md`, `CONTRIBUTING.md` y `CHANGELOG.md`, que dice también el cambio
    de contrato de D16: `--describe` marca las banderas propias y `schemas/instalacion.json` gana sus `title`.

Después, el workflow activa los guiones, ejecuta `make ci`, pasa la revisión final y hace el cierre. Ninguna tarea usa
la red, graba, abre una sesión con modelo, publica ni mide en la plataforma.

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada |
|---|---|---|
| `internal/httpx/httpxtest`, un paquete que no es de test con código solo para tests | R2 reserva `net/http`, y con él `httptest`, al árbol de `internal/httpx`; el control de FR-021 cuenta las peticiones de una consulta del adaptador contra un sitio de prueba, con su `robots.txt` y sin reproducción. Mismo patrón que `internal/mcp/mcptest`. El binario no lo enlaza, y `TestElBinarioNoEnlazaLosEjemplos` gana esa comprobación | Contar con la reproducción, que no tiene `robots.txt` ni reintentos; o poner el test del adaptador dentro de `internal/httpx` |
| Tareas `[datos]` que mezclan código (pasos 5, 7, 8, 10, 12 y 13) | El cambio de `--describe` y `schemas/instalacion.json` regenerado van juntos: con uno sin el otro, `TestEsquemasPublicados` está en rojo (research M5); el manifiesto y su test de grabación van juntos (`grabar_datos` exige los dos); la lectura y sus goldens, igual; en cuanto `cita` está registrado, `TestEsquemasCubrenTodosLosVerbos` exige su esquema y cambian las listas literales; el esquema de eval, su tipo y las evals que lo cumplen. Mismo patrón que H6, H19, H7 y H21 | Separar código y datos: `make ci` en rojo entre dos tareas |
| Veinticuatro guiones de hitos anteriores cambian en sus listas literales y en sus recuentos: `argumentos.txtar` (applets); `h21-mcp-herramientas.txtar`, `h21-mcp-proceso.txtar` y `h21-mcp-protocolo.txtar` (herramientas; V34); y los 20 que enumeran o cuentan las skills instaladas —17 de `internal/app/testdata/script/`, 15 de ellos con 72 líneas que las nombran juntas, y 3 de `internal/skills/testdata/script/`— (V35, M6) | El binario de e2e registra los applets y empotra las skills de verdad, y `make install` instala `skills/` entera: con un applet y una skill más, lo que anuncian, instalan y enumeran esos guiones es otra cosa. El cambio es añadir `cita`, `cita_resolver` y `jurisprudencia` a cada enumeración y una unidad a cada recuento que depende de ellas, nada más | Un binario de e2e con las skills de H19 congeladas: el e2e dejaría de probar lo que se distribuye |
| `--describe` gana `title` en las banderas propias de todo verbo, y la tabla de comandos, una forma más de escribir un argumento. Con ello cambia un fichero publicado de un hito anterior, `schemas/instalacion.json` (H19): siete líneas `title` | `cita resolver` es el primer verbo de una tabla de comandos con banderas propias, que el hito fija (`--roj`, `--resolucion`, `--fecha`), y sin eso su fila diría una orden que no funciona (V17). Los tres verbos de `skills` ya tenían las suyas, y su documento las marca por la misma regla (D16) | Explicar las banderas solo en prosa y dejar la fila generada como salga; o no marcar los verbos de `skills` para no tocar su esquema, que es un caso por applet en el kernel |
| Los umbrales de las respuestas dejan de depender solo del juez (D13 de H24) | FR-082 y FR-083 piden `sin_activar` y `cita_sin_resolver` para una skill sin juez | Dar a `jurisprudencia` una carpeta `juez`, que FR-076 veta |
| La subprueba `linea-sin-consulta` de H21 deja de exigir la línea `⚠ SIN CONSULTA AL BOE: …` a toda skill de `skills/`, y dos ayudantes de `TestSkillsDelRepositorio` dejan de dar por hecho que toda skill declara referencias (D33) | `jurisprudencia` es la primera skill que no consulta el BOE y la primera sin referencias: sin herramienta ni binario responde con su propia línea (FR-063), y el hito no le da ningún dato de `data/`. Con la skill en el árbol y esos tests como están, `make ci` queda en rojo (M6). La línea del BOE se sigue exigiendo a las skills cuyo conjunto lleva la eval sin binario ni servidor, que es la que la espera, con las dos de hoy como premisa | Escribir la línea del BOE en la skill nueva, que contradice FR-063; eximirla por su nombre, que es un caso por skill; o darle una carpeta `references/` vacía de sentido |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- `precheck.sh plan`: existen `plan.md` y `research.md`; ninguno conserva marcas de aclaración pendiente; están
  `## Constitution Check`, la línea «Aceptación e2e:» y `## Controles de umbral`, con cada fila nombrando requisitos
  que el spec define y un control con forma. Ejecutado al terminar el plan
  (`bash scripts/workflow/precheck.sh plan`): `precheck_plan ok`. Una réplica de un solo uso de sus reglas, fuera del
  repositorio, cuenta 22 filas y 33 controles. Los nombres de los diez umbrales de `boe-legislacion` son los de
  `specs/017-h24-las-evals-juzgan/gates/evals/boe-legislacion.json`. Tras la corrección de la ronda 1 el guion no se
  ha vuelto a ejecutar —pide aprobación en la sesión del corrector—; sus comprobaciones, lanzadas sueltas, dan lo
  mismo: ninguna marca pendiente, las dos secciones y la línea de aceptación en su sitio, y la tabla de umbrales, que la
  corrección no toca, con sus 22 filas.
- a · Constitution Check: un ítem por principio (I-IX) y por regla de dependencia (R1-R7).
- b · Dependencias: ninguna nueva.
- c · Reglas de dependencia: `net/http` y `cookiejar` solo en `internal/httpx`; ni SQLite, ni `os.Exit`, ni salida
  estándar nuevos.
- d · Errores y códigos: las filas del ADR 0023 en [contracts/cita-resolver.md §4](./contracts/cita-resolver.md); ningún
  código nuevo.
- e · Tests primero: cinco guiones congelados con su precondición; tests nuevos, tests que cambian, fixtures y tareas
  `[datos]`, con `schemas/instalacion.json` y la tarea del paso 5 que lo regenera.
- f · Alcance: nada fuera del spec; las ediciones de guiones anteriores y el cambio de alcance de `linea-sin-consulta`,
  declarados con su porqué.
- g · Sin atajos: ningún `//nolint`, `t.Skip`, TODO ni error silenciado previstos.
- h · Mejor alternativa: cada decisión con la rechazada (D1-D33); D16, rehecha sobre los verbos que ya tienen banderas
  propias.
- i · Afirmaciones verificadas: V1-V39 con fichero, línea u orden; M1-M4 del prototipo, M5 del de la corrección del
  plan y M6 del de la corrección de las tareas, que rehízo V34 y V35 con lo que falla de verdad; S1-S11 como supuestos.
- j · Quickstart ejecutable: órdenes, rutas y nombres de test que fija este plan; todo bajo un temporal que §10 borra;
  §11 y §12, para la persona y el workflow.
- k · Datos externos: una fuente con fila revisada en `main`, su manifiesto y su test de grabación; nada a `data/`.
- l · Autonomía: ninguna tarea para una persona; el cierre lo hace el workflow.
- m · Uso: «Uso, de fuera adentro» y los contratos, con bytes medidos, consultas por pregunta y apagado de cada señal.
- n · Proporcionalidad: «Trazabilidad»; sin mecanismos para estados sin vía real.
- o · Controles de umbral: las cuatro filas del cierre de FR-084, la de `boe-legislacion` y una por cada umbral medido
  en `make ci`.
