# Implementation Plan: H38 · Las skills instaladas se actualizan con el binario (ADR 0060)

**Branch**: `040-h38-las-skills-instaladas` | **Date**: 2026-11-21 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/040-h38-las-skills-instaladas/spec.md`

## Summary

Una consulta de la CLI con un binario más nuevo que las skills globales, y nada tocado, deja `~/.agents/skills` y sus
hosts como `kitlegal skills install -g` y lo dice en una línea; en cualquier otro caso sale la línea de hoy. El kernel,
en el punto en que hoy busca el aviso, calcula el **permiso** de la invocación (`SoloAvisar`, `Ensayar`, `Reinstalar`)
y se lo pasa al avisador; el dominio `instalacion` decide con un tipo nuevo, `Reinstalacion.Avisar(permiso)`, sobre
las piezas que ya tiene —`Aviso` (la línea de hoy), `Planificar` (conflictos y plan de `install`) y `Aplicar`— más una
precedencia SemVer propia y un puerto del candado; el candado lo da `cache.TomarCandado` en el directorio de la caché.
Se evalúa antes del candado y otra vez bajo él, que es lo que hace que de dos invocaciones a la vez reinstale una.
Cambian tres e2e de H19 y sus copias congeladas, `README.md` y `CHANGELOG.md`. Ninguna dependencia, paquete, orden,
bandera, código de salida ni sobre nuevo. Decisiones y verificaciones en research.md (D1-D13).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.4` en `go.mod`; `go version`: go1.27.4 linux/amd64).

**Primary Dependencies**: las fijadas, ninguna nueva. En los tests, `stretchr/testify` v1.12.1 y
`rogpeppe/go-internal` v1.16.0 (testscript), como hoy.

**Storage**: el sistema de ficheros: `~/.agents/skills` y los directorios de skills de los hosts globales (como
`install -g`), y un fichero vacío y efímero, `<cache.Directorio()>/skills-global.lock`. El manifiesto `kitlegal.json`
no cambia de forma.

**Testing**: `go test -race` en `make ci` (`make test`): unitarios de tabla sobre el disco en memoria del dominio y
sobre `t.TempDir()` en la caché; la tabla del kernel de `TestAvisoDelKernel`; cuatro guiones testscript de aceptación
congelados; tres guiones de H19 y sus copias que cambian. Sin red ni modelo. Skills de Go aplicadas en research.md D13.

**Target Platform**: Linux y macOS, donde corre `make ci`; Windows sin probar (la CI es solo Ubuntu; FR-017).

**Project Type**: CLI (binario multicall) que instala y pone al día las skills del producto.

**Performance Goals**: ninguno del spec. Con el global al día, cada consulta hace lo de H19 (leer el manifiesto y
comparar versiones); `Planificar` solo corre cuando alguna versión es menor (§ «Uso, de fuera adentro»).

**Constraints**: la salida estándar y el código de salida de toda invocación, byte a byte, los de hoy (FR-012); una
sola línea en la salida de error (FR-008); ninguna conexión (FR-012); nada escrito fuera del caso que reinstala
(FR-013); el candado nunca en `~/.agents/skills` (FR-010, FR-016).

**Scale/Scope**: 4 skills y 18 ficheros en el pedido de hoy; 2 tipos, 1 método y 1 puerto nuevos en el dominio, 1
función en la caché; 5 ficheros de producción que cambian y 2 nuevos; 5 unitarios nuevos, 1 tabla ampliada
(`TestAvisoDelKernel`) y 1 lista cerrada con una entrada más (`superficieDelContrato`); 4 guiones nuevos y 6 que
cambian; 2 documentos.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio o regla | Cómo lo cumple este hito |
|---|---|
| I · Fuentes públicas y frontera humana (INNEGOCIABLE) | No toca ninguna fuente ni `internal/httpx`; ninguna petición. Lo que escribe es local, en la cuenta de quien usa el kit, y solo lo que ya escribe `install -g` (FR-005) |
| II · Nada sin cita ni fuente | Ningún sobre cambia: la salida estándar es la de hoy, byte a byte (FR-012); la línea va a la salida de error. Ninguna skill ni referencia cambia |
| III · Tests primero y offline | Primero los cuatro guiones de `aceptacion/`, que fallan hoy por una aserción (cada uno lleva un caso que reinstala o la línea de FR-009); unitarios offline sobre el disco en memoria y `t.TempDir()`; ninguna grabación ni fixture nuevo; ninguna salida de applet nueva que validar contra `schemas/`. El código nuevo de `internal/core/instalacion` lo cubren `TestReinstalacion*` y `TestReglaDeVersiones` (umbral de cobertura de `internal/core/**`) |
| IV · Arquitectura hexagonal con reglas ejecutables | El dominio decide y define el puerto `TomarCandado`; la entrada y salida la hacen `internal/disco` y `internal/cache`; `internal/app` compone. Ningún código de salida nuevo: un fallo de la reinstalación no es un error de la invocación (FR-012) y se queda en la línea de hoy (research.md D11). Ningún `panic`: sin `Enlazador` el avisador solo da la línea de hoy (D12) |
| V · Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia: la precedencia SemVer es propia (D6; `golang.org/x/mod/semver` descartada). Ningún paquete nuevo; cada tipo, función, fichero y test se traza a un FR («Trazabilidad»); la sustitución del huérfano no lleva más mecanismo que el que pide FR-010 (D5) |
| VI · Un binario, convenciones de agente | Un binario; ninguna orden, bandera ni variable nueva (`KITLEGAL_CACHE_DIR` es la de H3). Igual por el primer argumento y por multicall (US1.8). `--dry-run` describe y no escribe (FR-009); `--describe` no cambia |
| VII · Grafo y privacidad | No toca el grafo. El candado es un fichero vacío: no guarda ningún dato de nadie |
| VIII · Skills primero; el binario es la herramienta | Protege las skills existentes: las que instaló la persona le llegan al día con el binario, sin cambiar ninguna `SKILL.md`, referencia ni eval (fundación de la distribución, ADR 0019 y 0060). Ninguna skill lee la línea (spec, «Contexto») |
| IX · Genericidad territorial, validación local | Nada territorial; las consultas de los e2e usan `territorio resolver Leganés` solo como orden sin red y de salida fija |
| X · Genericidad por materia, validación administrativa | Nada por materia: el pedido son las skills declaradas, sean de la materia que sean, más sus herederas |
| R1 · `internal/core/**` no importa el kernel, adaptadores ni entrada y salida | `reinstalacion.go`, `version.go` y `puertos.go` solo usan la biblioteca estándar sin entrada y salida y los tipos del propio paquete; el candado llega como puerto |
| R2 · solo `internal/httpx` importa `net/http` | Nada nuevo lo importa |
| R3 · solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | `internal/cache/candado.go` solo usa la biblioteca estándar de ficheros y tiempo (`os`, `path/filepath`, `time`, `errors`, `io/fs`); `internal/app` ya importa `internal/cache`; nada nuevo fuera de los tres importa SQLite |
| R4 · solo `internal/cli` y `cmd/` llaman a `os.Exit` (`forbidigo`) | Ningún `os.Exit` nuevo |
| R5 · solo `internal/render` escribe en stdout; logs con `log/slog` a stderr (`forbidigo`) | La línea sale por `Presentador.Aviso` (`internal/render`) a la salida de error, como hoy; nada escribe en stdout |
| R6 · `internal/graph` no importa `internal/source` ni `internal/render` | `internal/graph` no cambia |
| R7 · solo el paquete del servidor importa el SDK de MCP | Nada nuevo lo importa; `mcp serve` se reconoce por la interfaz `servidor` del kernel |
| Garantía sin red (H19) del aviso y del applet `skills` | `internal/app/aviso.go` sigue en `ficherosSinRed` (`internal/arch_test.go:720`) y pasa a importar `internal/cache`, que la subprueba «sin red» recorre por sus aristas del módulo; `internal/core/instalacion` e `internal/disco` siguen en `paquetesSinRed` (SC-005) |

Re-evaluado tras la fase 1: sin violaciones; Complexity Tracking, vacío.

## Project Structure

### Documentation (this feature)

```text
specs/040-h38-las-skills-instaladas/
├── spec.md
├── plan.md                         # este fichero
├── research.md                     # D1-D13, lo medido y los supuestos no verificados
├── data-model.md                   # permiso, Reinstalacion, candado, pedido, líneas, estados
├── quickstart.md                   # unitarios, e2e y dos binarios de verdad en un temporal
├── contracts/
│   ├── aviso-y-reinstalacion.md    # permiso, decisión, líneas con sus bytes, candado, qué escribe, uso
│   └── aceptacion.md               # los cuatro guiones y los cambios de FR-018
├── aceptacion/                     # lo escribe T001 y queda congelado
│   ├── reinstala.txtar
│   ├── tocadas.txtar
│   ├── excluidas.txtar
│   └── candado.txtar
└── tasks.md                        # /speckit-tasks
```

### Source Code (repository root)

```text
internal/core/instalacion/
├── reinstalacion.go        # NUEVO: PermisoDelAviso, Reinstalacion, Avisar (contracts/aviso-y-reinstalacion.md §2)
├── reinstalacion_test.go   # NUEVO: TestReglaDeVersiones, TestReinstalacion, TestReinstalacionConcurrente, TestReinstalacionFalloAMitad
├── puertos.go              # + TomarCandado
└── version.go              # + la precedencia SemVer, no exportada (research.md D6)
internal/cache/
├── candado.go              # NUEVO: TomarCandado(nombre, caducidad) (contracts/aviso-y-reinstalacion.md §4)
├── candado_test.go         # NUEVO: TestTomarCandado
└── superficie_test.go      # + «func TomarCandado» en superficieDelContrato y su comentario (research.md D4)
internal/app/
├── aviso.go                # Avisador con permiso; AvisoDeVersion compone Reinstalacion; nombre y caducidad del candado
├── aviso_test.go           # TestAvisoDelKernel con el permiso esperado; TestAvisoDeVersion con la firma nueva
├── main.go                 # permiso de la invocación en resolverApplet; avisar con permiso
├── registro.go             # Registro.aviso(permiso)
├── registro_test.go        # el avisador sintético, con la firma nueva
├── mcp_test.go             # ídem
└── testdata/script/        # [datos] h19-skills-aviso, h19-skills-aviso-sin-aviso, h19-skills-conflictos-dentro
specs/009-h19-instalar-sin-clonar/aceptacion/   # [datos] skills-aviso, skills-aviso-sin-aviso, skills-conflictos-dentro
README.md                   # el párrafo «Actualizar» (FR-017)
CHANGELOG.md                # Unreleased > Cambiado (FR-017)
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquetes nuevos: el dominio de la instalación
(`internal/core/instalacion`) decide, `internal/disco` y `internal/cache` son sus adaptadores y `internal/app` compone
y lleva el kernel.

## Diseño

1. **Kernel** (`internal/app/main.go`, research.md D2). En `resolverApplet`, donde hoy llama a `avisar`, calcula el
   permiso con contracts/aviso-y-reinstalacion.md §1: el análisis, la decisión, `--dry-run` y si los argumentos del
   verbo son un `servidor`. `avisar(p, registro, permiso)` y `Registro.aviso(permiso)` lo pasan al avisador. Nada
   más del kernel cambia: la ayuda, `skills`, `version` y las llamadas de herramienta siguen sin avisador.
2. **Dominio** (`internal/core/instalacion`, D1, D3, D6). `Reinstalacion.Avisar(permiso)` aplica contracts §2 paso a
   paso. El pedido es el de data-model §4; el plan, el de `Planificar` sin `--host`; la escritura, `Aplicar`. La regla
   de versiones usa la precedencia de `version.go` tras comprobar la forma con `FormaSemVer`. Las dos líneas nuevas se
   componen aquí, junto a la de hoy.
3. **Candado** (`internal/cache/candado.go`, D4, D5). `TomarCandado(nombre, caducidad)` con contracts §4: `Directorio()`,
   `MkdirAll` con `permisosDelDirectorio`, `OpenFile` con `O_WRONLY|O_CREATE|O_EXCL` y 0o600 sobre
   `filepath.Clean(filepath.Join(directorio, nombre))`, la fecha de modificación con `os.Lstat` y `time.Since`.
4. **Composición** (`internal/app/aviso.go`, D12). `Avisador func(instalacion.PermisoDelAviso) (string, bool)`;
   `AvisoDeVersion` compone `Reinstalacion` con lo mismo con que instala `install` y con
   `func() (func(), bool) { return cache.TomarCandado(ficheroDelCandado, caducidadDelCandado) }`, con
   `ficheroDelCandado = "skills-global.lock"` y `caducidadDelCandado = 10 * time.Minute`.
5. **Documentación** (FR-017). El párrafo «Actualizar» de `README.md` dice que las skills globales se ponen al día
   solas en la primera consulta con el programa nuevo, salvo que se hayan tocado, que se aplican en la próxima
   conversación, y que las skills nuevas del producto y las locales siguen pidiendo `kitlegal skills install`, con
   `-g` o sin él. `CHANGELOG.md`, en «Unreleased» > «Cambiado», lo mismo, con H38 y el ADR 0060, y declara Windows sin
   probar porque la CI es solo Ubuntu.

## Uso, de fuera adentro

Quién lee cada salida, cuánto mide y cuándo se apaga, con el detalle en contracts/aviso-y-reinstalacion.md §6:

- **La línea de FR-006** la leen el agente y la persona en la salida de error. Mide 119 bytes con `v0.2.0` y `v0.1.0`
  y 120 con `v0.10.1` y `v0.9.0` (medido, research.md); 0 bytes por norma o por bloque consultado: solo lleva dos
  versiones, así que con meses de uso diario, cientos de normas y miles de bloques en la caché mide lo mismo. Ninguna
  `SKILL.md` la pide; una pregunta lanza de una a varias órdenes, y la línea sale en una sola, la primera tras la
  actualización. Con una actualización al mes, son unos 120 bytes al mes en total. Se apaga en la consulta siguiente,
  porque las versiones ya son iguales (contracts §2, paso 1; US1.2).
- **La línea de FR-009** (122 bytes), solo con `--dry-run`, una vez por esa orden. Se apaga cuando una consulta sin
  `--dry-run` reinstala o cuando algo se toca.
- **La línea de hoy** (123 bytes) no cambia (FR-007): con algo tocado o el candado tomado sale en cada consulta, como
  en H19, es decir, 123 bytes por orden de cada pregunta mientras dure; se apaga con `kitlegal skills install -g` tras
  resolver los conflictos, o al terminar quien tiene el candado o caducar este (10 minutos). El hito la quita del caso
  más común, la instalación sin tocar, y no cambia sus condiciones (spec, «Fuera de alcance»).
- **El coste por invocación**: con el global al día, lo de H19. `Planificar` lee y huella como mucho los ficheros
  declarados de las skills del pedido —hoy 18 ficheros y 81 433 bytes (`du -sb skills`)—, nunca la caché ni el grafo,
  y solo cuando alguna versión es menor: en la primera consulta tras una actualización (dos veces, antes y bajo el
  candado) y, mientras algo está tocado, en cada consulta, donde ya salía la línea de hoy.

## Trazabilidad

| Mecanismo | Lo exige |
|---|---|
| `PermisoDelAviso` y su cálculo en `resolverApplet` (con la interfaz `servidor`) | FR-001, FR-002, FR-009, FR-015 |
| `Avisador` con permiso y `Registro.aviso(permiso)` | FR-001, FR-009 (el permiso tiene que llegar a quien decide) |
| `Reinstalacion` y `Avisar` | FR-003 a FR-007, FR-011 a FR-014 |
| La segunda evaluación bajo el candado | FR-010 («como mucho una»; «ninguna» si ya terminó), SC-004 |
| La precedencia SemVer de `version.go` | FR-004 |
| `TomarCandado` (puerto) y `cache.TomarCandado`, con nombre y caducidad | FR-010, FR-013 |
| La entrada «func TomarCandado» de `superficieDelContrato` (`internal/cache/superficie_test.go`) | FR-010: es lo que `cache.TomarCandado` exporta, y `TestSuperficieExportada` cierra la superficie del paquete en los dos sentidos (research.md D4) |
| La guarda del `Enlazador` nulo | constitución IV (ningún `panic` en una ruta de usuario) |
| `TestReglaDeVersiones` | FR-004, FR-019 |
| `TestReinstalacion` | FR-003, FR-005 a FR-007, FR-014; principio III |
| `TestReinstalacionConcurrente` | FR-010, SC-004 |
| `TestReinstalacionFalloAMitad` | FR-011, FR-019, US4.5 |
| `TestTomarCandado` | FR-010 (cada frase y el caso límite de `KITLEGAL_CACHE_DIR` del spec) |
| La tabla de `TestAvisoDelKernel` con el permiso | FR-001, FR-002, FR-009 |
| `aceptacion/reinstala`, `tocadas`, `excluidas`, `candado` | US1, US2, US3, US4 |
| Los seis guiones de H19 que cambian | FR-018 |
| `README.md` y `CHANGELOG.md` | FR-017 |

## Controles mecánicos que el hito añade o toca

Según «Gates» de la constitución, capa 1:

- **`make ci`**, sin objetivos nuevos: `make test` (`go test -race -shuffle=on ./...`, `Makefile:96-97`) ejecuta los
  unitarios nuevos y `TestEntregaDelHito` con los cuatro guiones activados y los tres de H19 cambiados; `make lint`
  pasa sobre el código nuevo con la configuración de hoy: `exhaustive` sobre el switch del permiso (research.md D13),
  `paralleltest` con `t.Setenv` sin `t.Parallel` (D7), `gosec`, `depguard` y `forbidigo`. El candado sigue el patrón
  de `internal/disco` (`filepath.Clean` en la ruta; permisos 0o600 y 0o700).
- **Superficie exportada de `internal/cache`**: `TestSuperficieExportada` (`internal/cache/superficie_test.go:112-140`)
  compara en los dos sentidos lo que el paquete exporta con la lista cerrada `superficieDelContrato` (líneas 59-93), y
  `cache.TomarCandado` sería una declaración exportada fuera de ella. La misma tarea que crea `candado.go` añade la
  entrada `"func TomarCandado"` a la lista y nombra el candado de H38 en su comentario y en el de cabecera, como H7 hizo
  con `"func Directorio"` (specs/010-h7-internal-graph-grafo/plan.md:324 y :494). Su firma no nombra `database/sql` ni
  SQLite, así que la otra mitad del test, la de las fugas, no cambia (research.md D4).
- **Test de arquitectura**: `TestArquitectura` sin cambios en su código; su subprueba «sin red» cubre desde ahora
  `internal/cache` a través de `internal/app/aviso.go`.
- **Tests de contrato contra `schemas/`** y `make schema-check`: sin cambios, porque ninguna salida de applet ni
  `--describe` cambia.
- **Fixtures**: ninguno nuevo. En `internal/app/testdata/script/` cambian los tres guiones de FR-018, en una tarea
  `[datos]` que los declara junto a sus copias de `specs/009-h19-instalar-sin-clonar/aceptacion/`; los cuatro de H38
  los activa el workflow.
- **Exit codes**: ninguno nuevo; los guiones comprueban que el de cada orden es el de la referencia (SC-003).

**Aceptación e2e:** cuatro guiones testscript que escribe la primera tarea en
`specs/040-h38-las-skills-instaladas/aceptacion/` y quedan congelados (contracts/aceptacion.md):
`reinstala.txtar` (US1 · FR-001, FR-003, FR-004, FR-005, FR-006, FR-008, FR-012, FR-013 · SC-001, SC-003),
`tocadas.txtar` (US2 · FR-003, FR-004, FR-005, FR-007, FR-013 · SC-002, SC-003, SC-006), `excluidas.txtar` (US3 y
el local inservible de «Edge Cases» · FR-001, FR-002, FR-003 a, FR-009, FR-013, FR-014, FR-015 · SC-003, SC-006) y `candado.txtar` (US4.1 a US4.4 y US4.6 · FR-010,
FR-011, FR-013 · SC-003, SC-004). US4.5 la cubre el unitario `TestReinstalacionFalloAMitad` (FR-019) y US5, que es
documentación, no tiene guion.

**Datos externos:** ninguno. El hito no graba ni consulta ninguna fuente, no escribe en `data/` ni en `evidencias/` y
no toca `docs/SOURCES.md` (research.md, «Datos externos»).

**Autonomía:** ninguna tarea pide nada a una persona ni publica la rama, abre la propuesta de cambio o mide CI y evals
remotas; eso lo hace el workflow tras la revisión final.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| SC-001 | 1.ª consulta: exactamente 1 línea en la salida de error, la de FR-006, y `kitlegal.json` en la v2; 2.ª: 0 líneas | `h38-reinstala`: `stderr '\A<FR-006>\n\z'`, `! grep 'v0\.1\.0'` en el manifiesto y `! stderr .` en la segunda | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| SC-002 | exactamente 1 línea, la de hoy, y 0 bytes cambiados en el árbol del global | `h38-tocadas`: `stderr '\A<hoy>\n\z'` y `cmp` de `arbol $HOME` antes y después en cada estado tocado | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| SC-003, FR-008, FR-012 | 100 % de los casos: salida estándar y código iguales a los de la referencia sin skills; salida de error con 0 o 1 línea de aviso más | en los cuatro guiones, `cmp stdout`, `test $? -eq N`, `\A…\n\z` y `una-linea-mas.sh` | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| SC-004, FR-010 | candado de menos de 10 min: no reinstala; de 10 min o más: reinstala; sin directorio: lo crea y reinstala; 0 candados tras reinstalar, bien o mal | `h38-candado`: los cuatro primeros casos de contracts/aceptacion.md §5 | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| SC-004, FR-010 | de dos invocaciones a la vez, como mucho 1 línea de FR-006 | `TestReinstalacionConcurrente`: los dos intercalados de research.md D7; sin el candado o sin la segunda evaluación salen 2 | `ci:internal/core/instalacion/reinstalacion_test.go:TestReinstalacionConcurrente` |
| FR-010 | caducidad de 10 minutos: a 9 min 30 s no se toma; a 10 min o más, sí | `TestTomarCandado` con `os.Chtimes` | `ci:internal/cache/candado_test.go:TestTomarCandado` |
| SC-005, FR-012 | 0 conexiones: ningún paquete del aviso ni de la reinstalación alcanza `net`, `net/http` ni `internal/httpx` | subprueba «sin red» de `TestArquitectura` sobre `internal/app/aviso.go`, `internal/core/instalacion`, `internal/disco` y, por sus importaciones, `internal/cache` | `ci:internal/arch_test.go:TestArquitectura` |
| SC-006, FR-013, FR-009 | en los excluidos y los tocados, 0 cambios en el árbol y 0 entradas en la caché; con `--dry-run`, exactamente 1 línea más, la de FR-009 | `h38-excluidas` y `h38-tocadas`: `cmp` de `arbol $HOME`, `! exists $KITLEGAL_CACHE_DIR` y `una-linea-mas.sh` | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| SC-007, FR-018, FR-019 | `make ci` con 0 fallos, con los unitarios de FR-019 y los e2e de FR-018 | `make ci` | `ci:Makefile:ci` |

## Complexity Tracking

Vacío: ninguna violación que justificar.

## Decisiones

Todas en research.md, con la alternativa descartada: D1 (la decisión en el dominio), D2 (el permiso en el kernel),
D3 (dos evaluaciones), D4 (el candado en la caché), D5 (cuándo no se toma y lo que no se diseña de la sustitución del
huérfano), D6 (precedencia SemVer propia), D7 (unitarios por la API pública, deterministas), D8 (la concurrencia en
e2e como aceptación), D9 (cómo se construyen los estados), D10 (la orden de las consultas), D11 (ningún error se
propaga), D12 (la firma del avisador) y D13 (skills de Go).
