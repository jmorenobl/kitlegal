# Implementation Plan: H21 · `kitlegal mcp serve`: las herramientas del binario por MCP, con las skills y las evals en los dos modos

**Branch**: `015-h21-kitlegal-mcp-serve` | **Date**: 2026-10-01 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/015-h21-kitlegal-mcp-serve/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de la constitución y
están en [research.md](./research.md) (D1-D26), cada una con su alternativa rechazada. Toda afirmación sobre el SDK de
MCP, Claude Code, Go, los linters o el repositorio remite a la tabla V de research.md, comprobada en local en la sesión
del plan (V43 a V45, en la de su corrector); lo que no se pudo comprobar así son los supuestos S1-S11. Los prototipos con
los que se midió son material de un solo uso, fuera de lo versionado (research, cabecera).

## Summary

H21 da a `boe-legislacion` y a `legal-core` una segunda forma de pedir cada operación: una herramienta de un servidor
MCP local, por la entrada y la salida estándar, que devuelve el mismo sobre que la orden. En siete piezas:

1. **El adaptador del protocolo, `internal/mcp`** (research D1, D2, D7, D10): el único paquete que importa
   `modelcontextprotocol/go-sdk` (v1.7.0). Recibe las herramientas ya hechas y las sirve: solo herramientas, todas de
   solo lectura, `capabilities` exactamente `{"tools":{}}`, las `instructions` de texto fijo (485 bytes) y la entrada que
   se cierra como final normal.
2. **Las herramientas salen del registro** (D5, D6): `cli.EsquemasDeHerramienta` da los dos esquemas de un verbo con el
   generador de `--describe` (la entrada, sin las ocho banderas globales) y `cli.LineaDeLlamada` convierte los
   argumentos de una llamada en la línea de órdenes del verbo, o los rechaza con la clase `argumentos`. Un applet nuevo
   es una herramienta sin tocar `mcp`.
3. **Una llamada es una invocación del kernel** (D3, D8): el mismo análisis de Kong, el mismo plazo, la misma caché, el
   mismo `internal/httpx` y el mismo `cli.Montador`, con el sobre en un búfer en lugar de en la salida estándar, y la
   entrega al grafo dentro de la llamada, antes de devolver. Cada llamada que llega a la red construye su cliente HTTP,
   como la orden, y lo que obtiene de `robots.txt` no vive más que ella; lo que comparten todas es el ritmo por sitio,
   con un valor nuevo de `internal/httpx` (`Ritmo`).
4. **El applet `mcp`, verbo `serve`** (D4, D9): un verbo que sirve —sin plazo y sin sobre propio— que el kernel reconoce
   por una interfaz sin exportar; `--offline`, `--no-graph` y `--timeout` valen para todas las llamadas; `--asunto` es un
   error de argumentos; `--dry-run` no sirve.
5. **Las dos skills** (D14, D15): la tabla generada gana la columna «Herramienta»; cada paso se dice de las dos formas;
   el `&&` son dos llamadas seguidas y el código de salida es `data.clase`; y la regla nueva con la línea
   `⚠ SIN CONSULTA AL BOE:`. `boe-legislacion` v0.1.5 (297 líneas) y `legal-core` v0.1 (190).
6. **Las evals en dos modos** (D16-D22): el job abre tres tandas —modo orden, modo herramienta y las dos evals sin
   binario ni servidor—; una llamada cuenta en el juicio como la orden de su verbo; el informe da tasas, recuentos y
   umbrales por modo (diez en `boe-legislacion`, ocho que deciden); el tope del trabajo pasa a 240 min; y el sondeo
   rechaza una eval sin binario ni servidor.
7. **Documentación** (FR-060, FR-061): README, CONTRIBUTING y `CHANGELOG.md`.

No cambia ninguna decisión de arquitectura que no haya cambiado ya el ADR 0035: no hay ADR nuevo, y no se edita ningún
spec anterior (FR-090).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.1`), `CGO_ENABLED=0`, `-trimpath`; bash sin nada posterior
a 3.2 (`scripts/evals.sh`, `scripts/evals-sesion.sh`); YAML de GitHub Actions.

**Primary Dependencies**: las fijadas, más **`github.com/modelcontextprotocol/go-sdk` v1.7.0**, la que el principio V
prevé para la distribución y adelanta el ADR 0035 (research D1). Arrastra seis módulos que el binario enlaza
(Complexity Tracking). Sin validador de JSON Schema nuevo: los argumentos de una llamada los comprueba el propio
analizador de la orden (D5).

**Storage**: ninguno nuevo. El servidor no guarda nada propio: usa la caché y el grafo del mundo de las órdenes, en el
mismo sitio. Ficheros nuevos del repositorio: `schemas/servidor.json`, dos evals y los datos de prueba sintéticos de
`internal/evals/testdata/`.

**Testing**: `go test -race` en `make ci`; cinco guiones `testscript` contra el binario de e2e con un cliente MCP
(«Aceptación e2e»); tests en proceso con tuberías para la conformidad y la concurrencia (D13); sesiones y transcripts
sintéticos y el `claude` sustituto para las evals. El job de evals, en la propuesta de cambio.

**Target Platform**: las seis plataformas de distribución (darwin, linux y windows, en amd64 y arm64); el job de evals,
`ubuntu-24.04`.

**Project Type**: CLI multicall (un applet nuevo y un adaptador nuevo) + dos skills + el job de evals.

**Performance Goals**: sesiones de cada modo de `boe-legislacion` ≤ 900 s (452 s con 96 sesiones en el cierre de H7.4,
modo orden; el modo herramienta no está medido: research S9); peor caso del trabajo ≤ 240 min (14 357 s, D20).
Ninguna cota nueva sobre el binario: el spec no la pide.

**Constraints**: la salida estándar de `mcp serve` es solo protocolo; ningún puerto; `SKILL.md` < 300 líneas (el tope
efectivo de `boe-legislacion` es 298); `instructions` con sus cinco contenidos en 512 bytes; ningún test de `make ci`
usa la red ni abre una sesión con modelo; ningún `//nolint`.

**Scale/Scope**: 10 herramientas con el registro de hoy (≈ 23 KB de `tools/list`, de los que ≈ 2,3 KB son nombre,
descripción y esquema de entrada); 21 evals de `boe-legislacion` y 4 de `legal-core`; 198 sesiones del job en
`boe-legislacion` (199 con la prueba de red) y 42 en `legal-core`.

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones; las desviaciones justificadas
están en Complexity Tracking.*

### Principios

| Principio | Cómo lo cumple H21 |
|---|---|
| I · Fuentes públicas y frontera humana | Ninguna fuente nueva, ningún cambio en `docs/SOURCES.md` y nada que grabar (D23). Toda petición de una llamada sale por `internal/httpx`, por el camino de la orden y con un cliente propio de la llamada: su `robots.txt` se pide y se evalúa en cada llamada que llega a la red, y ni sus reglas ni la denegación por no haberlo podido obtener viven más que ella, como en una invocación (D8; H2 FR 015, sin cambios). El ritmo por sitio es de todo el proceso: las llamadas, seguidas o a la vez, esperan turno en el mismo `Ritmo`, y el BOE no recibe del servidor más de una petición por intervalo (FR-014). El servidor no abre ningún puerto ni atiende conexiones: su único transporte son la entrada y la salida estándar (FR-001). Ninguna herramienta escribe ni actúa con identidad: todas son de solo lectura (FR-005) y `skills` no tiene herramientas. Ningún método distinto de GET o HEAD en el módulo. |
| II · Nada sin cita ni fuente | Cada llamada devuelve el sobre entero —`ok`, `fuente`, `url`, `fecha_consulta`, `hash`, `data`—, el de su orden (FR-010); un fallo, el sobre de fallo con su clase (FR-011). `mcp serve` no imprime un sobre propio: el principio se cumple en cada llamada, dentro del protocolo (spec, Assumptions). Las `instructions` dicen que nada se afirma sin el texto devuelto y que cada afirmación lleva norma y bloque (FR-006). Sin herramienta ni binario, la skill no afirma nada de memoria y lo dice con una línea de forma fija, sin citas (FR-035), y dos evals lo miden (FR-046). |
| III · Tests primero y offline | La primera tarea escribe la suite congelada: cinco guiones, en rojo por una aserción («Aceptación e2e»). Unitarios y de integración offline, sobre las grabaciones de H4 y las derivadas de H7 y H7.1; la salida de `mcp serve --describe` se publica en `schemas/servidor.json` y `schema-check` la compara; los esquemas y los datos de prueba nuevos van en tareas `[datos]`. Cobertura dentro de `codecov.yml`. |
| IV · Hexagonal y errores tipados | `internal/mcp` es un adaptador: traduce el protocolo y no sabe de applets ni de sobres (recibe `Herramienta` y devuelve bytes). La composición sigue en `internal/app`. Los errores son los de hoy: `cli.ErrArgumentos` y las demás clases del ADR 0023, que en una llamada van en `data.clase` del sobre de fallo y en `mcp serve` salen como código (2 con `--asunto`; 0 al cerrarse la entrada; 1, `inesperado`, con cualquier otro final). Ningún `panic` en rutas de usuario: lo único que `AddTool` del SDK exige con uno es un esquema de entrada de tipo objeto (V7), que es el que `EsquemasDeHerramienta` da siempre y la conformidad comprueba verbo a verbo. Reglas R1-R7, abajo. |
| V · Simplicidad y dependencias | Una dependencia directa, de la lista del principio; las seis que arrastra, en Complexity Tracking. Cada mecanismo se traza a un FR («Trazabilidad»). Rechazados por no pedirlos el spec: un validador de JSON Schema (D5), un campo exportado en `Verbo` (D4), un contexto de cancelación en el kernel (D9), un binario cliente aparte (D12), un trabajo del job por modo (D16) y cualquier pieza del protocolo que no sean las herramientas y las `instructions` (D7). |
| VI · Un binario, convenciones de agente | `mcp` es un applet del multicall, con su verbo en el registro, su ayuda, su `--describe` y las ocho banderas globales con el significado de siempre (FR-020 a FR-022). Las herramientas se generan de `--describe`, como prevé el principio. Ningún ejecutable nuevo en `cmd/`. |
| VII · Grafo y privacidad | La entrega al grafo de una llamada es la de la orden: después del resultado, con la procedencia del sobre, nunca en un fallo y nada con `--no-graph` (FR-013). El servidor no expone nada del asunto: `--asunto` es un error y ninguna herramienta lo recibe (FR-021). `graph_show` sigue sin devolver texto legal: es el verbo de H7. Ningún dato de persona. |
| VIII · Skills primero | No entrega skill nueva: protege las dos que hay, haciéndolas usables donde no hay shell. La tabla de comandos nombra cada operación como orden y como herramienta (constitución 2.10.0); `SKILL.md` < 300 líneas, sin `scripts/` y sin nombrar evals ni modelos; las evals miden las dos skills en los dos modos, con sus umbrales por modo. Ninguna herramienta que una skill no use: son los verbos que ya tienen. |
| IX · Genericidad territorial | Nada se particulariza para un municipio: las herramientas salen del registro. `territorio_resolver` se prueba en el e2e con un municipio cubierto (Leganés) y con uno no cubierto (Tordesillas), y su sobre lleva la `cobertura` de la orden. Getafe aparece solo en la pregunta de una eval. |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución IV)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa adaptadores | `internal/core` no cambia. La lista `core` de `depguard` gana `internal/mcp`, que es un adaptador más. |
| R2 · Solo `internal/httpx` importa `net/http` | Ningún paquete del módulo lo importa de nuevo: `internal/mcp` importa el SDK, no `net/http`. El paquete `mcp` del SDK sí lo importa, porque lleva sus transportes HTTP en el mismo paquete (V5), y el binario lo enlaza sin usarlo: `TestArquitectura` y `depguard` miran las importaciones directas de los paquetes del módulo (V28). Declarado en Complexity Tracking. El sitio local de `TestDependenciasDeRed` es un `net.Listener` del test, sin `net/http` (V45). |
| R3 · Solo `cache`/`store`/`graph` importan SQLite y `database/sql` | Ningún fichero nuevo los importa: las llamadas llegan a la caché y al grafo por el kernel. |
| R4 · Solo `cli` y `cmd/` llaman a `os.Exit` | Ningún `os.Exit` nuevo: `Servir` devuelve un error y el kernel lo traduce a código, como con cualquier verbo. |
| R5 · Solo `render` escribe en stdout | `internal/mcp` escribe el protocolo en el escritor que recibe, que es `Salida()` del presentador del kernel (`internal/render`); ni `os.Stdout`, ni `os.Stderr`, ni `fmt.Print*` nuevos. El aviso y los eventos van por el presentador y el registrador, a la salida de error (FR-023). La entrada estándar llega por `DependenciasDeMCPDelSistema`, en la raíz de composición. |
| R6 · `internal/graph` no importa `source/*` ni `render` | Sin cambios. |
| R7 (nueva) · Solo `internal/mcp/**` importa `github.com/modelcontextprotocol/go-sdk` | Lista nueva de `depguard` (`.golangci.yml`) y subprueba nueva de `TestArquitectura` sobre el grafo real, que exige además que el dueño esté en el grafo (FR-026, FR-079). Los tests de `internal/app` hablan con el servidor por `internal/mcp/mcptest`, que es de ese árbol. |

### Gates

- **Capa 1**, todo en `make ci` o en el workflow: suite congelada con «rojo primero»; `schema-check` con
  `schemas/servidor.json`; `skills-check` con la tabla nueva, las 300 líneas, las herramientas de la tabla y la línea
  `⚠ SIN CONSULTA AL BOE:`; `TestArquitectura` (R7) y `TestDependenciasDelBinario`; `depguard`; matriz territorial en
  el e2e de `territorio_resolver`; guardián de diff con `testdata/` y `schemas/` solo en tareas `[datos]`; «Controles de
  umbral» con requisitos del spec.
- **Capa 2**: los jueces del plan, de las tareas y de la revisión final; la activación y el protocolo de las dos skills,
  por el job de evals en los dos modos.
- **Capa 3**, al leer el informe final: `schemas/servidor.json` y `schemas/eval.yaml.json`; los guiones y los datos de
  prueba que cambian (`argumentos.txtar` y los sintéticos de `internal/evals/testdata/`) y los activados; los supuestos
  del run; la prueba humana de SC-002; la fusión. **Ninguna pausa a mitad del run.**

## Project Structure

### Documentation (this feature)

```text
specs/015-h21-kitlegal-mcp-serve/
├── plan.md                  # este fichero
├── research.md              # V1-V45, S1-S11, D1-D26
├── data-model.md            # herramienta, llamada, servicio, modo, eval sin binario ni servidor, umbral por modo
├── quickstart.md            # §1-§14
├── contracts/
│   ├── servidor-mcp.md          # la orden, las herramientas, una llamada, la salida de error, instructions, firmas, tests
│   ├── arnes-e2e.md             # la orden `mcp` del arnés, los dos clientes y los cinco guiones
│   ├── skills.md                # C1-C12, la línea, la tabla generada, comprobaciones
│   ├── skills-prototipo.diff    # el texto de las dos SKILL.md, sobre 22b5bda
│   └── evals-en-dos-modos.md    # formato, sesiones, juicio, informe, umbrales, job, sondeo, tests
├── aceptacion/              # la escribe la tarea [aceptacion]
├── checklists/              # del spec
└── tasks.md                 # la escribe /speckit-tasks
```

### Source Code (repository root)

```text
go.mod, go.sum                                   # + go-sdk v1.7.0 y sus indirectas (D1)
.golangci.yml                                    # depguard: R7, e internal/mcp en la lista de R1
internal/arch_test.go                            # R7 en TestArquitectura; siete módulos en modulosDelBinario
internal/mcp/                                    # NUEVO: el adaptador del protocolo
├── doc.go
├── instrucciones.go         # Instrucciones
├── servir.go                # Herramienta, Resultado, Servicio, Servir
├── instrucciones_test.go    # TestInstrucciones
├── servir_test.go           # TestServir
└── mcptest/                 # NUEVO: los dos clientes de prueba (D12)
    ├── sesion.go            # Abrir, Sesion (Protocolo, Instrucciones, Capacidades, Herramientas, Llamar, Lineas, Cerrar)
    └── sesion_test.go       # TestSesion
internal/cli/
├── herramienta.go           # NUEVO: EsquemasDeHerramienta, LineaDeLlamada
├── herramienta_test.go      # TestEsquemasDeHerramienta, TestLineaDeLlamada
└── describe.go              # el generador, compartido con herramienta.go
internal/httpx/                                  # el ritmo por sitio que comparten los clientes de un proceso (D8)
├── ritmo.go                 # Ritmo, NuevoRitmo
├── cliente.go               # ConRitmo; Replay la rechaza
├── sitio.go                 # el limitador de cada sitio sale del Ritmo del cliente
├── ritmo_test.go            # TestRitmoCompartido
└── cliente_test.go, reproducir_test.go, errores_test.go, sitio_test.go   # las tablas de opciones y lo que leen por dentro
internal/app/
├── mcp.go                   # NUEVO: AppletMCP, DependenciasDeMCP, DependenciasDeMCPDelSistema, el verbo serve
├── herramientas.go          # NUEVO: herramientasDe (del registro) y la llamada como invocación del kernel
├── main.go                  # la interfaz servidor en ejecutarVerbo; emitir, el final que comparten Main y cada llamada
├── registro.go              # RegistroDeProduccion registra AppletMCP
├── boe.go                   # DependenciasDeRed: un Ritmo por proceso y un cliente por invocación
├── ejemplo/kitlegal-e2e/main.go                 # el registro de e2e registra AppletMCP
├── mcp_test.go              # NUEVO: TestLlamadasSimultaneas, TestServirSinEntrada, TestServirEnEnsayo, TestPlazoDeCadaLlamada
├── herramientas_test.go     # NUEVO: TestHerramientasDelServidor
├── e2e_test.go              # la orden `mcp` del arnés
├── boe_test.go, skills_test.go, esquemas_test.go, registro_test.go, ejemplo/kitlegal-e2e/main_test.go
└── testdata/script/argumentos.txtar             # CAMBIA [datos]: la lista de applets gana mcp (l. 24 y 30)
cmd/kitlegal/main_test.go                        # la lista de applets y `kitlegal mcp` sin verbo
schemas/servidor.json                            # NUEVO [datos]: mcp serve --describe
schemas/eval.yaml.json                           # CAMBIA [datos]: sin_binario_ni_servidor
internal/skills/comandos.go, comandos_test.go    # la columna «Herramienta» y la línea del sobre
skills/boe-legislacion/SKILL.md                  # v0.1.5
skills/legal-core/SKILL.md                       # v0.1
evals/boe-legislacion/21-sin-binario-ni-servidor.yaml   # NUEVA
evals/legal-core/04-sin-binario-ni-servidor.yaml        # NUEVA
internal/evals/
├── formato.go               # Eval.SinBinarioNiServidor
├── conjunto.go              # 21 evals; la regla «sin binario ni servidor»; la subprueba linea-sin-consulta
├── plan.go                  # Modo, PlanDeEvals.Modos, tres tandas
├── sesiones.go              # SesionesAEjecutar.Binario, servidor.json, PATH por modo
├── sesion.go                # Llamada, Sesion.Llamadas
├── citas.go                 # ExtraerSinConsulta, junto a ExtraerCitas
├── juzgar.go                # la llamada como invocación; la orden en una sesión sin binario; la eval sin binario ni servidor
├── informe.go, umbrales.go  # tasas, recuentos y umbrales por modo
├── definicion.go            # el peor caso por tandas
├── sondeo.go                # el error de uso de la eval sin binario ni servidor
├── job_test.go              # TestEjecucionDelJob: una llamada al repartidor por tanda; bandera -kitlegal
├── doc.go
├── testdata/                # [datos]: transcripts y sesiones sintéticos con llamadas
└── *_test.go                # los de contracts/evals-en-dos-modos.md §8
scripts/evals.sh                                 # -kitlegal "$(command -v kitlegal)"
scripts/evals-sesion.sh                          # --mcp-config ../servidor.json, si el fichero está
.github/workflows/evals.yml                      # timeout-minutes: 240, con su comentario
README.md, CONTRIBUTING.md, CHANGELOG.md
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, con un adaptador nuevo, `internal/mcp`, al lado de
`httpx`, `cache`, `graph` y `render`: traduce un protocolo de fuera y no conoce el dominio. El applet vive en
`internal/app`, como los demás, porque necesita el registro y el kernel. `internal/mcp/mcptest` es código de prueba que
no enlaza el binario: existe porque `depguard` no deja importar el SDK desde los tests de `internal/app` (R7). El
directorio `mcp/` de la raíz que `CLAUDE.md` prevé es de la distribución (H22) y no se crea.

## Aceptación e2e

**Aceptación e2e:** cinco guiones `testscript`, uno o varios por historia con comportamiento del binario, que la tarea
`[aceptacion]` escribe en `specs/015-h21-kitlegal-mcp-serve/aceptacion/` y quedan congelados
([contracts/arnes-e2e.md §4](./contracts/arnes-e2e.md)). Cada uno empieza por la precondición `exec kitlegal --help` +
`stdout '^\s+mcp\s'`, que hoy falla en su aserción y no en una orden desconocida (research V40).

| Guion | Historia | FR / SC |
|---|---|---|
| `mcp-llamadas.txtar` | US1 (1, 5, 6), US6 | FR-010, FR-012, FR-013, FR-020; SC-004 |
| `mcp-errores.txtar` | US1 (2, 3, 4) | FR-011, FR-015, FR-020; SC-004 |
| `mcp-herramientas.txtar` | US6 (1, 2), US4 (6) | FR-001 a FR-006, FR-008; SC-003 |
| `mcp-protocolo.txtar` | US4 (1, 2) | FR-007, FR-023; SC-005, SC-006 |
| `mcp-proceso.txtar` | US4 (3, 4, 5, 7) | FR-021, FR-022, FR-024, FR-025; SC-008 |

US2, US3, US5 y US7 no son comportamiento del binario —son el texto de dos `SKILL.md`, el juicio y el informe del job, y
la documentación— y no tienen guion. Las fijan, en `make ci`: US2, `skills-check`, `TestOrdenesDeLasSkillsEmpotradas` y
`TestRenderizarTabla` (FR-030, FR-031, FR-036; SC-009); US3, `TestJuzgarSinBinarioNiServidor` y la subprueba
`linea-sin-consulta` (FR-035, FR-047; SC-012); US5, `TestUmbralesDelInforme`, `TestInformeEnDosModos` y
`TestJuzgarLasLlamadas` (FR-042 a FR-045, FR-080, FR-081; SC-011, SC-012); US4.8, `TestLlamadasSimultaneas` (FR-074;
SC-007). El plazo de cada llamada (`--timeout`, FR-020 y su caso límite) tampoco tiene guion, porque la orden `mcp` del
arnés no tiene esperas y ningún verbo del binario de e2e se queda esperando: lo fija `TestPlazoDeCadaLlamada`, en
proceso y con un verbo del propio test. Y, en el cierre, el job de evals (SC-001). Las dos evals nuevas no las escribe la tarea `[aceptacion]`: su
clave no existe en el formato hasta la tarea que cambia `schemas/eval.yaml.json`, y antes dejarían `make ci` en rojo;
entran con esa tarea, antes que el texto de las skills (paso 6).

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

Ninguno nuevo ni cambiado. Con más dentro: `test` (los tests nuevos y los cinco guiones, en `TestEntregaDelHito`),
`lint` (`depguard` con R7), `schema-check` (`schemas/servidor.json`) y `skills-check` (la tabla con su columna nueva,
las herramientas de la tabla, la línea `⚠ SIN CONSULTA AL BOE:` y las reglas del conjunto con 21 y 4 evals).

### Tests nuevos

| Test | Fichero | Cubre |
|---|---|---|
| Guiones `h21-mcp-*` (cinco) | `internal/app/testdata/script/`, en `TestEntregaDelHito` | FR-071 a FR-073, FR-075, FR-076; SC-004 a SC-006, SC-008 |
| `TestHerramientasDelServidor` | `internal/app/herramientas_test.go` | FR-002 a FR-005, FR-008, FR-020, FR-070; SC-003 |
| `TestLlamadasSimultaneas` | `internal/app/mcp_test.go` | FR-014, FR-074; SC-007 |
| `TestServirSinEntrada`, `TestServirEnEnsayo` | `internal/app/mcp_test.go` | FR-022, FR-024 |
| `TestPlazoDeCadaLlamada` | `internal/app/mcp_test.go` | FR-020 (`--timeout`): con `mcp serve --timeout=<plazo corto>` y un verbo del test que espera a su contexto, la llamada da `fuente-no-disponible` como error de herramienta; la siguiente, hecha con ese plazo ya vencido desde el arranque, recibe su resultado, y el servidor termina con 0 al cerrarse la entrada |
| `TestEsquemasDeHerramienta`, `TestLineaDeLlamada` | `internal/cli/herramienta_test.go` | FR-003, FR-011, FR-020 |
| `TestRitmoCompartido` | `internal/httpx/ritmo_test.go` | FR-014 (dos clientes con el mismo `Ritmo`: ninguna llegada se adelanta a su turno), FR-010 (un `robots.txt` que falla una vez y responde después: el primer cliente no lee y el segundo sí) |
| `TestInstrucciones` | `internal/mcp/instrucciones_test.go` | FR-006, FR-077; SC-010 |
| `TestServir` | `internal/mcp/servir_test.go` | FR-005, FR-008, FR-010, FR-011, FR-024 |
| `TestSesion` | `internal/mcp/mcptest/sesion_test.go` | las comprobaciones que el arnés hace por su cuenta (FR-071 a FR-073) |
| Subprueba R7 de `TestArquitectura` | `internal/arch_test.go` | FR-026, FR-079; SC-013 |
| `TestPlanEnDosModos`, `TestSesionesPorModo`, `TestGuionDeLaSesion`, `TestLeerLasLlamadas` | `internal/evals/` | FR-040 a FR-042 |
| `TestJuzgarLasLlamadas`, `TestJuzgarSinBinarioNiServidor` | `internal/evals/juzgar_test.go` | FR-042, FR-047, FR-081; SC-012 |
| `TestInformeEnDosModos` | `internal/evals/informe_test.go` | FR-044, FR-045, FR-048, FR-080; SC-011 |

### Tests existentes que cambian

- `TestDependenciasDeRed` (`internal/app/boe_test.go`): dos invocaciones reciben dos clientes distintos, y los dos
  esperan turno en el mismo `Ritmo` —el segundo falla sin abrir ninguna conexión cuando el primero ha gastado el turno, y
  uno de otras dependencias sí la abre— (FR-010, FR-014; research D8).
- En `internal/httpx`, las tablas que recorren las opciones del cliente ganan las filas de `ConRitmo`:
  `TestOpcionesInvalidas` (nulo y de intervalo no positivo, y la última de `ConIntervalo` y `ConRitmo` gana),
  `TestReplayRechazaOpciones` y `TestErrorMensajesEnEspanol`. `TestNewSinOpciones` y `TestMapaDeSitios`, que leen por
  dentro el intervalo y el limitador de un sitio, los leen del `Ritmo` del cliente (V45). Los demás tests del paquete no
  cambian: sin `ConRitmo`, un cliente es el de hoy.
- `TestOrdenesDeLasSkillsEmpotradas` (`internal/app/skills_test.go`): toda herramienta de la tabla está entre las del
  servidor, con un control que ve fallar una herramienta y una orden inventadas (FR-031).
- `TestRenderizarTabla` (`internal/skills/comandos_test.go`) y `TestTablaDeComandosCoincideConLaGramatica`: la columna
  nueva y la línea del sobre (FR-030).
- `TestEsquemasPublicados` y `TestEsquemasCubrenTodosLosVerbos` (`internal/app/esquemas_test.go`): la fila
  `servidor.json` (D11).
- Las listas literales de applets, que registrar `mcp` cambia: `internal/app/registro_test.go`,
  `cmd/kitlegal/main_test.go` (con `kitlegal mcp` sin verbo → 2 y `verbos de mcp: serve`),
  `internal/app/ejemplo/kitlegal-e2e/main_test.go` e `internal/app/testdata/script/argumentos.txtar` (V41). La tarea
  busca antes cualquier otra lista o comparación de un `Resultado` entero que la nombre, y la declara.
- `TestDependenciasDelBinario`: `modulosDelBinario` gana los siete módulos (V4).
- `TestEvalsDelRepositorio` y `TestConjuntoDeEvals`: 21 evals, la regla nueva y la subprueba `linea-sin-consulta`;
  `TestEsquemaDeEval` y `TestLeerEval`: la clave nueva.
- `TestUmbralesDelInforme`, `TestInforme…`, `TestInformeMarkdownDeLosUmbrales`, `TestJuzgar…`, `TestLeerSesion…`,
  `TestPlan`, `TestEjecutarSesiones…`: los resultados y los informes esperados ganan `modo`, los nombres de umbral con
  su modo y las columnas de `informe.md`. Ninguno se desactiva ni se salta.
- `TestDefinicionDelJob`: el tope de 240 y el peor caso por tandas (FR-083).
- `TestComprobarElSondeo`, `TestSondear`, `TestGuionDelSondeo`: los casos de FR-084.

### Puntos de entrada fuera de `make ci` (etiqueta `evals`)

`TestEjecucionDelJob` (las tres tandas y la bandera `-kitlegal`) y `TestSondeo` (el error de uso nuevo). `golangci-lint`
los lintea. Ninguna tarea los ejecuta.

### Fixtures, `testdata/` y `schemas/` (tareas `[datos]`)

- `schemas/servidor.json`, escrito por `TestEsquemasPublicados -actualizar-esquemas`, con el registro de `mcp` y las
  listas literales de applets, `argumentos.txtar` entre ellas (paso 5).
- `schemas/eval.yaml.json`, con los casos de `formato_test.go` que lo fijan, `Eval.SinBinarioNiServidor`, las reglas del
  conjunto y las dos evals nuevas (paso 6).
- Los transcripts y las sesiones sintéticos nuevos de `internal/evals/testdata/` (llamadas a herramientas, con prefijo y
  sin él, con error y sin resultado; sesiones con `servidor.json`), con los tests que los leen (pasos 8 y 9).
- Los cinco guiones activados, que copia el workflow tras el bucle de tareas.

### CI

`.github/workflows/evals.yml`: `timeout-minutes: 240`. Lo comprueba `TestDefinicionDelJob` en `make ci`. `ci.yml` no
cambia. El job de evals se ejecuta en la propuesta de cambio, tras la revisión final, y lo lanza el workflow (FR-078).

## Controles de umbral

Las ocho filas del cierre que fija FR-082 —los cuatro umbrales que deciden, por cada modo— y, como pide el ADR 0029
(criterio o), una por cada umbral que se mide en `make ci`, con el test o la comprobación que lo pone en rojo. El umbral
de las expresiones del modelo informativo no tiene fila: se publica con `decide: false` en cada modo, como hoy.

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| FR-043, SC-001 (expresiones, modo orden) | ≤ 5 % de las respuestas medidas del modelo que decide en evals que activan la skill, en ese modo (≤ 2 de 54) | umbral del informe con `decide: true`: por encima, motivo, veredicto `fallo` y el trabajo en rojo, que el cierre cuenta | `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5-5:orden` |
| FR-043, SC-001 (expresiones, modo herramienta) | lo mismo, sobre las respuestas del modo herramienta | ídem | `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5-5:herramienta` |
| FR-043, SC-001 (sin activar, modo orden) | 0 respuestas del modelo que decide sin la skill activada | ídem | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden` |
| FR-043, SC-001 (sin activar, modo herramienta) | 0 | ídem | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta` |
| FR-043, SC-001 (redacción no leída, modo orden) | 0 respuestas del modelo que decide con una expresión de esa familia | ídem | `evals:boe-legislacion:redaccion_no_leida:claude-sonnet-5-5:orden` |
| FR-043, SC-001 (redacción no leída, modo herramienta) | 0 | ídem | `evals:boe-legislacion:redaccion_no_leida:claude-sonnet-5-5:herramienta` |
| FR-043, SC-001 (duración, modo orden) | ≤ 900 s, de la preparación de la primera sesión de la tanda al final de la última | umbral del informe con `decide: true`: por encima, motivo de la ejecución y `fallo` | `evals:boe-legislacion:duracion_de_las_sesiones:orden` |
| FR-043, SC-001 (duración, modo herramienta) | ≤ 900 s | ídem | `evals:boe-legislacion:duracion_de_las_sesiones:herramienta` |
| FR-002, FR-070, SC-003 | las herramientas son exactamente las de los verbos del registro menos los excluidos (10 con el de producción); 10 de 10 con nombre, descripción, dos esquemas y solo lectura; 0 `prompts` y 0 `resources` | conformidad, con el registro de producción y con uno que lleva un applet más | `ci:internal/app/herramientas_test.go:TestHerramientasDelServidor` |
| FR-071, SC-004 | 10 de 10 herramientas con el sobre de su orden; 3 de 3 fallos con el sobre de su clase como error | guiones `h21-mcp-llamadas` y `h21-mcp-errores` | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| FR-072, FR-073, SC-005, SC-006 | 0 líneas de la salida estándar que no sean del protocolo y 1 aviso en la de error; 2 de 2 clientes | guion `h21-mcp-protocolo` | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| FR-074, SC-007 | N de N llamadas simultáneas con su propio resultado; 0 carreras | test en proceso, con `-race` | `ci:internal/app/mcp_test.go:TestLlamadasSimultaneas` |
| FR-075, FR-076, SC-008 | código 0 al cerrarse la entrada; 2 con `--asunto`; 0 mensajes respondidos y 0 bytes con `--dry-run`; el mismo sobre desde `/` y una ruta con espacios | guion `h21-mcp-proceso` | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| FR-036, FR-077, SC-009 (líneas) | cada `SKILL.md` < 300 líneas | `TestSkillsDelRepositorio`, que falla con 300 o más | `ci:Makefile:skills-check` |
| FR-031, FR-077, SC-009 (nombres) | 0 herramientas de la tabla de una skill fuera del servidor y 0 órdenes fuera del registro | la tabla contra el registro y sus herramientas | `ci:internal/app/skills_test.go:TestOrdenesDeLasSkillsEmpotradas` |
| FR-036, SC-009 (H7.4) | el calibrado en 36, 11 y 9; 0 expresiones de la lista en la prosa y en la respuesta hecha de los bloques | subpruebas `expresiones-calibradas`, `prosa-de-la-skill`, `expresiones-de-la-skill` y `ordenes-para-powershell` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-006, FR-077, SC-010 | los 5 contenidos de las `instructions` en sus primeros 512 bytes | el texto contra sus cinco frases y su tamaño | `ci:internal/mcp/instrucciones_test.go:TestInstrucciones` |
| FR-080, SC-011 | un umbral incumplido en un solo modo → `fallo`; 3 de 54 y 0 de 54 → no se cumple; 2 de 54 en los dos → se cumple; una serie que falla en un modo → `fallo`; 0 series de los dos modos sin celda propia | informes de sesiones sintéticas | `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme`, `ci:internal/evals/informe_test.go:TestInformeEnDosModos` |
| FR-081, SC-012 | la herramienta satisface el `comando`; 0 satisfechos por otra norma, otro bloque o un resultado de error; la prohibida no pasa; de 4 respuestas sin binario ni servidor pasa 1 | tabla de sesiones sintéticas | `ci:internal/evals/juzgar_test.go:TestJuzgarLasLlamadas`, `ci:internal/evals/juzgar_test.go:TestJuzgarSinBinarioNiServidor` |
| FR-079, SC-013 | 0 importaciones del SDK de MCP fuera de `internal/mcp` | `depguard` (R7) y el grafo real de importaciones | `ci:.golangci.yml`, `ci:internal/arch_test.go:TestArquitectura` |
| FR-083 | `timeout-minutes` ≥ el peor caso con las evals del repositorio y los dos modos (14 357 s); 1 tanda por commit y skill | `fallosDelTope` y la definición | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |
| FR-084 | 0 sesiones abiertas y salida 1 con una eval sin binario ni servidor en `EVALS`; una línea por eval | el sondeo con el `claude` sustituto | `ci:internal/evals/sondeo_test.go:TestComprobarElSondeo`, `ci:internal/evals/sondeo_test.go:TestSondear`, `ci:internal/evals/sondeo_test.go:TestGuionDelSondeo` |

Lo que SC-001 pide además de sus umbrales —que toda serie que decide pase en cada modo y que pasen las dos evals sin
binario ni servidor— no es un umbral agregado sino la regla por serie (ADR 0016), que ya pone el veredicto en `fallo` y
el trabajo en rojo; con dos modos lo comprueba `TestInformeEnDosModos`. `legal-core` no tiene umbrales (FR-045). SC-002
lo mide una persona.

## Uso, de fuera adentro (criterio de uso, ADR 0028)

Detalle, con ejemplos y bytes, en [contracts/servidor-mcp.md §8](./contracts/servidor-mcp.md),
[contracts/skills.md §1, §3 y §6](./contracts/skills.md) y
[contracts/evals-en-dos-modos.md §9](./contracts/evals-en-dos-modos.md). Resumen:

- **`tools/list`** (el agente; una vez por conexión; la skill no la pide): diez herramientas, ≈ 23 KB en el cable
  (1 190 B de esquemas de entrada, 19 289 B de salida, 971 B de descripciones); al modelo le llegan nombre, descripción
  y esquema de entrada, ≈ 2,3 KB. Crece con los verbos, no con el uso. No da señales.
- **El resultado de una llamada** (la skill; las veces que hoy la orden: en `boe-legislacion`, por cada bloque,
  `boe_articulo` y `graph_check`, más `boe_buscar` y `boe_indice` si hacen falta; en `legal-core`, `territorio_resolver`
  una vez por municipio): el sobre de la orden, dos veces en el mensaje. Art. 21 de la Ley 39/2015: 4 433 B de sobre,
  ≈ 9 KB. Una pregunta por el art. 118 de la LCSP con su índice: 34 720 + 2 625 + ≈ 325 = ≈ 37 670 B de sobres en tres
  llamadas, ≈ 75 KB en los mensajes. Con cientos de normas y miles de bloques consultados es lo mismo: lo acotan la
  norma y el bloque pedidos, y el servidor no guarda nada propio.
- **`graph_check`** (la skill; una por lectura de bloques, con su norma y sus bloques): ≈ 325 B sin nada y ≈ 1 000 B por
  bloque cambiado (H7.1). Cada señal la apaga la siguiente lectura del bloque.
- **El error de herramienta** (la skill; cuando una llamada falla): el sobre de fallo, 300-400 B. Ninguna llamada hereda
  el fallo de otra: cada una vuelve a pedir lo que le falta, su `robots.txt` incluido (D8), así que deja de darse en la
  primera llamada que se hace cuando su causa ha cesado —la fuente vuelve a responder, el argumento se corrige—, como
  con la orden.
- **La línea `⚠ SIN CONSULTA AL BOE:`** (la persona; una por respuesta, no una por norma): ≈ 150 B más la causa. Deja de
  darse en la primera pregunta tras instalar kitlegal o declarar el servidor.
- **Las `instructions`** (el agente; una vez por conexión): 485 B. No dan señales.
- **El aviso de versión**, en la salida de error (quien mira el registro de su agente): una línea por arranque del
  servidor, no una por llamada. Deja de darse al reinstalar las skills.
- **Los eventos**, en la salida de error (quien depura): sin `--verbose`, solo las llamadas que fallan, ≈ 150 B cada una.
  No se guardan.
- **`SKILL.md`** (el modelo; una vez por conversación): 297 y 190 líneas; la tabla, las mismas filas con ≈ 20 B más.
- **`umbrales`, `tasas` y recuentos del informe** (el job, `informe.sh` y la persona; una vez por job y skill): 10
  umbrales de ≈ 270 B y 66 series en `boe-legislacion`; `[]` y 14 series en `legal-core`. Tamaño fijo: lo da el conjunto
  de evals. Cada job los mide de nuevo.

## Decisiones

- **El SDK se fija en la v1.7.0**, la que la caché de módulos tiene verificada: la v1.8.0 que leyó el ADR 0035 no se
  puede descargar ni añadir dentro del sandbox del run (research D1, V1, V2).
- **`internal/mcp` no sabe de applets**: recibe herramientas hechas; el applet y la llamada viven en `internal/app` (D2).
- **Una llamada pasa por Kong**, con sus argumentos detrás de `--`, y no rellena el struct a mano (D3, D5).
- **`mcp serve` es un verbo del registro** que el kernel reconoce por una interfaz sin exportar, no un caso fuera del
  registro como `version` (D4).
- **La entrada cerrada es un final normal**, reconocido por un lector que anota su fin y no por el texto de un error del
  SDK (D7).
- **Cada llamada construye su cliente HTTP, como la orden, y el ritmo por sitio es del proceso**: lo que se obtiene de
  `robots.txt` —reglas o denegación— no vive más que la llamada, y un cliente por proceso lo dejaría vivo lo que el
  servidor (D8; supuesto).
- **La conformidad y la concurrencia se prueban en proceso**, para que `-race` vea el código del servidor (D13).
- **El arnés gana la orden `mcp` con dos clientes**, el del SDK y uno de la 2025-11-25 escrito con la biblioteca estándar
  (D12).
- **Tres tandas seguidas en un solo trabajo**, no dos modos a la vez ni un trabajo por modo (D16); **240 minutos** de
  tope (D20; supuesto).
- **El modo va en `modelo` de `tasas` y en el nombre de cada umbral**, para que el `informe.sh` de hoy dé una celda por
  serie sin cambiarlo (D19).
- **Una orden de `kitlegal` en una sesión sin `kitlegal` en el `PATH` la deja sin pasar** (D18).
- **`argumentos.txtar` gana `mcp` en su lista de applets**: es lo que registrar un applet cambia desde H6, y lo único
  que se toca de un guion anterior (D26; Complexity Tracking; supuesto).
- **Cada guion de aceptación empieza por una precondición** que hoy falla en su aserción (D25).
- Las demás, con su alternativa rechazada, en research D1-D26.

## Trazabilidad: cada mecanismo y su requisito

| Mecanismo | Requisito |
|---|---|
| `internal/mcp`: `Servir`, `Servicio`, `Herramienta`, `Resultado` | FR-001, FR-005, FR-007, FR-008, FR-010, FR-011, FR-024, FR-026 |
| `internal/mcp.Instrucciones` y `TestInstrucciones` | FR-006, FR-077 |
| Lector que anota el fin de la entrada | FR-024 |
| `cli.EsquemasDeHerramienta` | FR-003, FR-020 |
| `cli.LineaDeLlamada` | FR-010, FR-011, FR-020; FR-042 (el juicio la usa para leer los argumentos de una llamada) |
| `herramientasDe` (del registro, menos `skills` y `mcp`) | FR-002, FR-004 |
| La llamada como invocación del kernel; `emitir` compartido | FR-010, FR-012, FR-013, FR-015 |
| Registro sin avisador para las llamadas | FR-023 |
| Interfaz `servidor` en `ejecutarVerbo`; `--timeout=<plazo>` en la invocación de cada llamada; `TestPlazoDeCadaLlamada` | FR-020 (sin plazo de vida, y el plazo en cada llamada), FR-022, FR-023 |
| `AppletMCP`, `DependenciasDeMCP`, `DependenciasDeMCPDelSistema` | FR-001, FR-021, FR-022, FR-025 |
| `httpx.Ritmo`, `NuevoRitmo` y `ConRitmo`; `DependenciasDeRed` con un `Ritmo` por proceso y un cliente por invocación; `TestRitmoCompartido` | FR-014 (el ritmo por sitio, entre todas las llamadas); FR-010 (el `robots.txt` de cada llamada, como el de su orden) |
| `schemas/servidor.json` y su fila | FR-022 (`--describe`), FR-079 (`schema-check`) |
| R7 en `depguard` y en `TestArquitectura`; `modulosDelBinario` | FR-026, FR-079 |
| `internal/mcp/mcptest` y la orden `mcp` del arnés | FR-071 a FR-076 |
| `TestHerramientasDelServidor`, `TestLlamadasSimultaneas` | FR-070, FR-074 |
| Columna «Herramienta» y línea del sobre en `RenderizarTabla` | FR-030 |
| `TestOrdenesDeLasSkillsEmpotradas` con las herramientas | FR-031 |
| C1-C12 de las dos `SKILL.md` | FR-032 a FR-036 |
| `Eval.SinBinarioNiServidor`, su esquema, las reglas del conjunto y las dos evals | FR-046 |
| `ExtraerSinConsulta`, el juicio de la eval sin binario ni servidor, `linea-sin-consulta` | FR-035, FR-047 |
| `Modo`, `PlanDeEvals.Modos`, las tres tandas | FR-040, FR-043 (duración por modo), FR-047 |
| `servidor.json`, `PATH` por modo, `--mcp-config` | FR-040, FR-041 |
| `Llamada`, `Sesion.Llamadas`, la llamada como invocación en el juicio | FR-041, FR-042 |
| Motivo de la orden en una sesión sin binario | FR-040, FR-046 (lo que define el modo) |
| `tasas`, recuentos y `umbrales` por modo; motivos | FR-043 a FR-045, FR-048 |
| Test de la regla de `informe.sh` | FR-048, FR-080 |
| `peorCasoDelTrabajo` por tandas; `timeout-minutes: 240` | FR-048, FR-083 |
| Error de uso del sondeo | FR-050, FR-084 |
| README, CONTRIBUTING, `CHANGELOG.md`, `doc.go` | FR-060, FR-061 |

Nada del diseño atiende a un estado que no pasa el umbral de materialidad: un mensaje que no es del protocolo, una
salida estándar rota, una caché o un `world.db` tocados a mano y un transcript ilegible siguen las reglas que ya hay
—defecto `inesperado` y código 1 en el binario; sesión ilegible o test que falla en el job—, sin caso propio.

## Datos externos

Ninguno (research D23): ni fuente ni grabación nuevas. Los guiones usan las respuestas del BOE grabadas en H4 y las
derivadas de H7 y H7.1 que ya están en `testdata/`. Ningún manifiesto `grabaciones.json` ni test `TestGrabar*` cambia, y
el paso `grabar_datos` no tiene nada que grabar. El módulo del SDK lo descarga la herramienta `go` del proxy de módulos,
que es la red de las herramientas de Go, no una fuente.

## Orden de implementación (de dentro afuera)

1. **`[aceptacion]`** los cinco guiones en `specs/015-h21-kitlegal-mcp-serve/aceptacion/`, cada uno con su precondición,
   desde spec.md y contracts/arnes-e2e.md; en rojo por la aserción de la precondición.
2. `go.mod` y `go.sum` con la orden de D1, en una sola línea, e `internal/mcp` entero (`Instrucciones`, `Servir` y sus
   tipos, con `TestInstrucciones` y `TestServir`), en la misma tarea: sin un paquete que lo importe, `go mod tidy`
   retiraría el módulo. Con ella, R7 en `.golangci.yml` y en `TestArquitectura`, e `internal/mcp` en la lista de R1.
3. `internal/mcp/mcptest`: los dos clientes, con `TestSesion`.
4. `internal/cli`: `EsquemasDeHerramienta` y `LineaDeLlamada`, con sus tests. Y, sin depender de lo anterior,
   `internal/httpx`: `Ritmo`, `NuevoRitmo` y `ConRitmo`, con `TestRitmoCompartido` y las filas nuevas de las tablas de
   opciones.
5. `internal/app`: primero el cuerpo —la interfaz `servidor` en `ejecutarVerbo`, `emitir`, `herramientasDe`, la llamada,
   `AppletMCP` y sus dependencias, el `Ritmo` del proceso en `DependenciasDeRed`—, probado sobre un registro local del
   test (`TestHerramientasDelServidor`, `TestLlamadasSimultaneas`, `TestServirSinEntrada`, `TestServirEnEnsayo`,
   `TestPlazoDeCadaLlamada`, `TestDependenciasDeRed`), sin registrarlo. Después, **`[datos]`** e indivisible, su registro y su contrato publicado:
   `RegistroDeProduccion` y el registro de e2e, `schemas/servidor.json` generado con `-actualizar-esquemas`, la fila de
   `esquemas_test.go`, las cuatro listas literales de applets (`argumentos.txtar` entre ellas) y `modulosDelBinario`.
   Y la orden `mcp` del arnés en `e2e_test.go`.
6. **`[datos]`** `schemas/eval.yaml.json` con sus casos de `formato_test.go`, `Eval.SinBinarioNiServidor`, las reglas del
   conjunto, `ExtraerSinConsulta`, las dos evals nuevas y `timeout-minutes: 240` (con 21 evals, el peor caso de
   `boe-legislacion` en un solo modo ya es `485 + ⌈103/4⌉ × 272` = 7 557 s, que no cabe en los 122 minutos de hoy, y
   `TestDefinicionDelJob` falla sin subirlo). El sondeo da ya su error de uso para esas evals, con sus tests (FR-050,
   FR-084).
7. `internal/skills/comandos.go` con la columna y la línea; las dos `SKILL.md` con el texto del prototipo y su tabla
   regenerada por `make skills-sync`; `TestOrdenesDeLasSkillsEmpotradas` con las herramientas; la subprueba
   `linea-sin-consulta`. En una tarea: la tabla y el texto se comprueban juntos en `skills-check`.
8. **`[datos]`** `Modo` y las tres tandas en el plan, el peor caso del trabajo por tandas, `servidor.json` y el `PATH` por
   modo en el repartidor, el guion de la sesión y `scripts/evals.sh`, con las sesiones sintéticas que los prueban.
9. **`[datos]`** `Llamada` y `Sesion.Llamadas`, el juicio (la llamada como invocación, la orden en una sesión sin
   binario, la eval sin binario ni servidor), con sus transcripts sintéticos.
10. El informe por modo: `tasas`, recuentos, `umbrales`, motivos e `informe.md`, con `TestUmbralesDelInforme` y
    `TestInformeEnDosModos`; `TestEjecucionDelJob` con una llamada al repartidor por tanda.
11. README, CONTRIBUTING, `CHANGELOG.md` e `internal/evals/doc.go`.

**Obligaciones para `tasks.md`**: exactamente una tarea `[aceptacion]`, la primera; cada tarea deja `make ci` en verde;
las tareas `[datos]` son las de los pasos 5, 6, 8 y 9, con las rutas de este plan; cada fila de «Controles de umbral»
tiene la tarea que construye su control, con un test que lo ve fallar; ninguna tarea ni corrección cumple un umbral de
FR-043 rebajándolo, dejándolo en `decide: false`, sacando evals o un modo del total, sumando los modos o recortando la
lista (FR-049); ninguna tarea edita un spec, un plan o una suite congelada de un hito anterior, ni `scripts/workflow/`
(FR-048, FR-090); de los guiones anteriores solo cambia `argumentos.txtar`, y solo su lista de applets; ninguna tarea
usa la red —salvo la de la herramienta `go` al añadir el módulo—, graba, abre una sesión con modelo, ejecuta
`make evals`, `make evals-sondeo`, `TestEjecucionDelJob` o `TestSondeo`, publica ni mide en la plataforma; los
escenarios §11 a §13 del quickstart los ejecutan la persona o el workflow.

## Complexity Tracking

| Desviación | Por qué hace falta | Alternativa más simple rechazada |
|---|---|---|
| Seis módulos que el binario enlaza y no están en la lista del principio V: `google/jsonschema-go`, `segmentio/encoding`, `segmentio/asm`, `yosida95/uritemplate/v3`, `golang.org/x/oauth2` y `golang.org/x/sync` (V4). Ninguno se importa desde el módulo | Los importa el paquete `mcp` del SDK, que es la dependencia que el principio prevé; `modulosDelBinario` los declara uno a uno y `TestDependenciasDelBinario` falla con cualquier otro | Escribir el protocolo a mano con la biblioteca estándar: negociar cinco versiones de la especificación (FR-007) es lo que el SDK ya hace y mantiene |
| El binario enlaza `net/http` por el SDK (R2) | El paquete `mcp` del SDK trae sus transportes HTTP en el mismo paquete (V5). El módulo no los usa: `Servir` solo construye un `IOTransport`, y ningún paquete del módulo importa `net/http` fuera de `internal/httpx` | No hay un subpaquete del SDK con solo el transporte de entrada y salida estándar |
| El SDK en la v1.7.0 y dos indirectas fijadas (`x/oauth2` v0.37.0, `segmentio/asm` v1.2.1) por encima de lo que el SDK pide | Son las versiones que la caché tiene verificadas: dentro del sandbox del run, `go` no puede escribir en `sumdb` para verificar otras (V1, V2) | Desactivar la comprobación de sumas: quita una garantía de la cadena de suministro |
| `internal/mcp/mcptest`, un paquete que no es de test con código solo para tests | R7 veta el SDK en los `_test.go` de `internal/app` (V27), y el arnés e2e y la conformidad necesitan un cliente | Exceptuar los `_test.go` en la lista de `depguard`: R7 dejaría de ser «solo el paquete del servidor» |
| Tareas `[datos]` que mezclan código (pasos 5, 6, 8 y 9) | En cuanto `mcp serve` está registrado, `TestEsquemasCubrenTodosLosVerbos` exige su parte publicada y las listas literales de applets cambian a la vez; el esquema de eval, su tipo y las evals que lo cumplen, igual. Mismo patrón que H6, H19 y H7 | Separar registro y esquema: `make ci` en rojo entre dos tareas |
| `argumentos.txtar`, un guion de H1, cambia en dos líneas, cuando el spec deja fuera «editar los guiones de los hitos anteriores» | Sus líneas 24 y 30 fijan la lista de applets del binario de e2e, y FR-001 y FR-071 exigen `mcp` en ese registro. El cambio es añadir `mcp` a la lista, como `territorio`, `skills` y `graph` en sus hitos; nada más de ningún guion anterior | Un binario de e2e aparte, solo para `mcp`: el e2e dejaría de probar el registro que se distribuye |
| `timeout-minutes: 240`, por encima de las 3 h que espera el cierre del workflow | Con tres tandas, el peor caso de `boe-legislacion` es 14 357 s (D20), y FR-083 exige que el tope lo cubra. Lo esperado son unos 17 minutos de sesiones; el cierre solo se quedaría sin informe si casi todas las sesiones agotaran su tope, y ese informe sería `fallo` | Subir la concurrencia a 6: cambia el punto en el que H7.3 midió la suscripción, sin poder medirlo en el run |

## Comprobación contra la rúbrica del juez (`juez_plan`) y `precheck.sh plan`

- `precheck.sh plan`: existen `plan.md` y `research.md`; ninguno conserva marcas de aclaración pendiente; están
  `## Constitution Check`, la línea «Aceptación e2e:» y `## Controles de umbral`, con cada fila nombrando requisitos
  que el spec define y un control con forma. Comprobado al terminar el plan con las órdenes del propio guion, una a
  una (su `awk` de las filas y sus `grep`): 22 filas, sin defectos.
- a · Constitution Check: un ítem por principio (I-IX) y por regla de dependencia (R1-R6, y R7 nueva).
- b · Dependencias: una, de la lista; las que arrastra, en Complexity Tracking.
- c · Reglas de dependencia: tabla R1-R7; ni SQLite, ni `os.Exit`, ni salida estándar nuevos; `net/http`, solo por el
  SDK y declarado.
- d · Errores y códigos: las clases del ADR 0023, en `data.clase` de cada llamada y como código de `mcp serve`; ningún
  código nuevo.
- e · Tests primero: cinco guiones congelados con su precondición; «Tests nuevos», «Tests existentes que cambian»,
  fixtures y tareas `[datos]`.
- f · Alcance: nada fuera del spec; la única edición de un guion anterior, declarada con su porqué.
- g · Sin atajos: ningún `//nolint`, `t.Skip`, TODO ni error silenciado previstos.
- h · Mejor alternativa: cada decisión con la rechazada (D1-D26).
- i · Afirmaciones verificadas: V1-V45 con fichero, línea u orden; S1-S11 como supuestos.
- j · Quickstart ejecutable: órdenes, rutas y nombres de test reales (los que fija este plan); todo bajo un temporal que
  el último escenario borra; nada versionado cambia.
- k · Datos externos: ninguno.
- l · Autonomía: ninguna tarea para una persona; §11 a §13 del quickstart los ejecutan la persona o el workflow, fuera
  de las tareas; el cierre en la plataforma lo hace el workflow.
- m · Uso: «Uso, de fuera adentro» y los contratos (bytes, llamadas por pregunta, apagado de cada señal).
- n · Proporcionalidad: «Trazabilidad»; sin mecanismos para estados sin vía real.
- o · Controles de umbral: las ocho filas del cierre de FR-082 y una por cada umbral medido en `make ci`.
