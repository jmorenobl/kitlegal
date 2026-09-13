<!-- Propuesta de cambio de H4. La sección «Dependencias» la escribió la tarea que enlaza boe en el binario distribuido (T026); el resto lo escribe la tarea de cierre (T036) con la medida hecha sobre 55cc107. -->

## Objetivo

«`kitlegal boe articulo BOE-A-2015-10565 a21` con el mismo comportamiento que `boe.py`» (`docs/ROADMAP.md` §4, H4).
H0 a H3 dejaron la fundación sin ninguna fuente legal: un sobre citable, una única puerta a la red y una memoria local
con vigencia. **H4 es el primer hito que consulta derecho de verdad**: el applet `boe`, porte a Go de `refs/boe.py`
(la herramienta de la skill `boe-fiscal`, copia congelada del 2026-05-14), contra la API de Legislación Consolidada del
BOE, con seis verbos —`buscar`, `indice`, `articulo`, `articulos`, `metadatos`, `analisis`—, caché por verbo, sobre
firmado con `fuente: "boe.legislacion-consolidada"`, esquemas de salida publicados y una verificación nocturna contra la
fuente real.

Tres compromisos, y los tres se comprueban sin red:

1. **El mismo comportamiento que `boe.py`.** El `data` de `articulo` coincide, campo a campo, con lo que calcula el
   `refs/boe.py` original sobre las mismas grabaciones en seis artículos de cuatro leyes (el criterio pedía cinco de
   tres). Donde `boe.py` hace algo que las decisiones cerradas prohíben —errores como texto con código 0, fallos en
   caché, XML crudo como texto legal, callar que no pudo comprobar la vigencia—, el porte se aparta y lo deja escrito en
   `internal/source/boe/doc.go` (32 anotaciones vigiladas por `TestDocAnotaElPorte`).
2. **Nada sin cita ni contrato.** Toda salida va en el sobre, se valida contra `schemas/norma.json` y
   `schemas/bloque.json` —generados desde `--describe` y vigilados por `make schema-check`, que deja de ser un aviso— y
   se fija con 13 golden sobre fixtures grabados.
3. **No volver a molestar al BOE.** Lo consultado se sirve desde la caché de H3 con la vigencia que declara la fuente
   (300 s en `buscar` y `metadatos`, 7 días en el resto) y con la fecha en que de verdad se consultó (ADR 0015), en
   menos de 200 ms, y con `--offline` sin salir nunca a la red. Ningún fallo se guarda.

Hito de **fundación** (principio VIII): no entrega ni cambia ninguna skill. Es la herramienta determinista que
`boe-legislacion` (H5) necesita.

## Alcance

Frente a `main`: 189 ficheros, 30 579 líneas añadidas y 311 retiradas, en 42 commits. Por árboles:

- **`internal/source/boe`, paquete nuevo** (19 ficheros de producto, 21 de test): `fuente.go` (`Nueva`, `Fetch`, `TTL`,
  `Terms`, el esqueleto `resolverRecursoDeLaNorma[T]`), los seis verbos en su fichero (`buscar.go`, `indice.go`,
  `articulo.go` con `articulos`, `metadatos.go`, `analisis.go`), `ids.go` (gramáticas de norma y bloque,
  `TipoDesdeID`), `direcciones.go`, `busqueda.go`, `consultas.go`, `peticiones.go`, `lectura.go`, `bloque.go`,
  `avisos.go`, `datos.go`, `entradas.go`, `errores.go`, `terminos.go` y `doc.go`. Sus datos, todos bajo `testdata/`:
  el manifiesto `grabaciones.json` y 23 grabaciones (22 recursos más el `robots.txt`), 10 ficheros sintéticos en 7
  escenarios, 6 referencias generadas con `refs/boe.py` y revisadas por una persona, y 13 golden.
- **`internal/app`**: `boe.go` (el applet) y `registro.go` (`RegistroDeProduccion`, que registra exactamente `boe`);
  `main.go` gana `Arrancar`; `cmd/kitlegal/main.go` arranca por él. Cinco guiones e2e nuevos
  (`boe-verbos`, `boe-multicall`, `boe-offline`, `boe-codigos`, `boe-cache-rapida`) y `argumentos.txtar` con la lista
  de applets del binario; el binario de e2e (`ejemplo/kitlegal-e2e`) registra `boe` sobre reproducción.
- **`schemas/`**: `norma.json` (`buscar`, `indice`, `metadatos`, `analisis`) y `bloque.json` (`articulo`, `articulos`),
  los primeros contratos de salida versionados.
- **Ficheros de H0-H3 tocados**, y ninguno más:
  - `internal/core`: `doc.go` y `source.go` (el puerto `core.Source`, sin sentencias); `internal/core/schema/sobre.go`:
    `Procedencia.FechaConsulta`; `internal/cli/sobre.go`: el montador fecha con ella cuando la hay.
  - `internal/httpx`: `Peticion.Acepta`, `Respuesta.Instante`, `Error.Instante`, `ConHora`, `instante.go` (la marca de
    emisión como decorador en las dos cadenas), `robots.go` (la obtención del `robots.txt` no fecha), `grabar.go`
    (solo un comentario) y sus tests, entre ellos `reproducir_test.go`, que fija la hora con `ConHora`.
  - `internal/cache/migraciones.go`: la transacción de cada migración nace con `context.WithoutCancel`, para que un
    contexto que termina durante la migración no deje la conexión abierta desde otra goroutine. Es el arreglo de un
    fallo intermitente de un test de H3 que apareció en el `make ci` de T003; su test lo acompaña.
  - `internal/arch_test.go`: `TestLasFuentesNoFirmanComoKitlegal` y la lista `modulosDelBinario` ampliada; se retiran
    `TestElBinarioNoEnlazaHTTPX` y `TestElBinarioNoEnlazaCache`.
  - `Makefile` (`schema-check` real, `verify-sources` nuevo), `.golangci.yml` (`run.build-tags` con `fuentes` y
    `grabacion`; seis palabras españolas en `misspell.ignore-rules`), `.github/workflows/nightly.yml` (trabajo
    `fuentes`), `scripts/grabar-fixtures.sh` y `scripts/verify-sources.sh` (nuevos), `docs/ADR/0015-puerto-source-y-fecha-de-consulta.md`
    (nuevo), `docs/SOURCES.md` (la fila de la fuente, revisada el 2026-09-13), `docs/PENDIENTES.md` (se borran las dos
    entradas de H4), `README.md`, `CONTRIBUTING.md` y `CHANGELOG.md`.
- **Fuera del producto, tocados en la rama**: `.specify/workflows/hito/workflow.yml` y `docs/WORKFLOW.md` (commit
  `5a631b5`, workflow 1.9.0: el commit de cada tarea se decide con el estado de la iteración; lo motivó T003 en este
  mismo run), `CLAUDE.md` y `refs/00-README.md` (la enmienda sobre la única ejecución de `refs/boe.py`, en la pausa
  `[datos]`), y `refs/__pycache__/boe.cpython-311.pyc`, un artefacto que llegó con el commit `[datos]` `17fca5b` y que
  no debe fusionarse (ver *Pendientes*).
- **Artefactos del hito**: `specs/005-h4-applet-boe-puerto/` completo (spec con cinco clarificaciones, plan, research
  D1-D20, data-model, cinco contratos, quickstart, tasks y `gates/`, con las notas de las tareas T003, T005, T009, T022,
  T023, T024 y T036).

**Fuera de alcance** (spec, *Fuera de alcance*): ninguna skill, ni `data/normas.yaml`, ni los verbos `sumario`,
`vigilar`, `eli`, `buscar-materia` o `materias`, ni versiones históricas, paginación, revalidación o mantenimiento de
la caché, ni operaciones de grafo (los ids naturales quedan en `data` para H7). Puntos 10, 11 y 12 de la Definition of
Done: no aplican (sin skill, sin dimensión territorial, sin grafo hasta H7).

## Dependencias (constitución §V)

**Ninguna entrada nueva en `go.mod`** (FR-125). Lo que cambia es la **superficie del binario distribuido**: al
registrar `boe`, `cmd/kitlegal` enlaza `internal/source/boe`, `internal/httpx` e `internal/cache`, y con ellos doce
módulos de terceros que hasta H3 solo usaban esos paquetes y sus tests. Por eso se retiran
`TestElBinarioNoEnlazaHTTPX` y `TestElBinarioNoEnlazaCache`, que solo eran ciertos mientras ningún applet usara la red
ni la caché, y `modulosDelBinario` (`internal/arch_test.go`) pasa de seis a dieciocho módulos, cada uno justificado en
su línea; `TestDependenciasDelBinario` falla ante cualquier otro (FR-124, research.md D14, `docs/PENDIENTES.md`).

**Los módulos enlazados dependen de la plataforma.** `modernc.org/libc` elige sus ficheros por sistema, así que el
binario de cada plataforma de distribución (darwin, linux y windows sobre amd64 y arm64, sin cgo; ADR 0002) enlaza una
parte de los dieciocho: todos en darwin; dieciséis en linux, sin `github.com/mattn/go-isatty` (lo importa `libc.go`,
con `//go:build !linux || mips64le`) ni `github.com/ncruces/go-strftime` (lo importan `libc_unix.go`, que excluye
linux, y `libc_windows.go`); y diecisiete en windows, sin `github.com/google/uuid` (lo importan los ficheros de darwin,
de linux y del resto de unix). La lista declarada es la unión de las seis. La primera ejecución de `ci` de esta
propuesta lo destapó: `TestDependenciasDelBinario` medía solo la plataforma del ordenador que lo ejecuta y falló en el
ejecutor linux de la integración continua (ver *Pendientes*).

Medida sin red (`go list` consulta el módulo y la caché de módulos):

```sh
go list -deps -f '{{if .Module}}{{.Module.Path}}{{end}}' ./cmd/kitlegal | sort -u
go list -deps -f '{{.ImportPath}}|{{if .Module}}{{.Module.Path}}{{end}}|{{join .Imports ","}}' ./cmd/kitlegal
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go list -deps -f '{{if .Module}}{{.Module.Path}}{{end}}' ./cmd/kitlegal | sort -u
```

La primera da la lista de la plataforma en que se ejecuta; la segunda, para cada módulo, qué paquete de otro módulo lo
importa, que es la columna «Lo importa» de la tabla; la tercera, la de otra plataforma (aquí linux/amd64). La unión
sobre las seis coincide con lo que research.md D14 anticipó (S10): ni uno más ni uno menos.

| Módulo | Versión | Lo importa | Justificación |
|---|---|---|---|
| `github.com/temoto/robotstxt` | v1.1.2 | `internal/httpx` | Lista cerrada de §V (H2). Interpreta el `robots.txt` de cada sitio antes de la primera petición; una ruta denegada es el código 5 (constitución §I, FR-122) |
| `golang.org/x/time` | v0.16.0 | `internal/httpx` (`golang.org/x/time/rate`) | Lista cerrada de §V (H2). El ritmo por sitio, con el intervalo que fija la fila de la fuente en `docs/SOURCES.md` (FR-122) |
| `modernc.org/sqlite` | v1.58.0 | `internal/cache` | Lista cerrada de §V (H3, ADR 0002). SQLite sin cgo para la caché de las consultas (FR-090); sin él no hay binario con `CGO_ENABLED=0` |
| `modernc.org/libc` | v1.75.6 | `modernc.org/sqlite` | Entra con el controlador: es el entorno de C traducido a Go sobre el que corre SQLite, y no hay versión del controlador sin él (H3, `gates/pr-h3.md`) |
| `modernc.org/mathutil` | v1.7.1 | `modernc.org/libc` | Entra con `modernc.org/libc` |
| `modernc.org/memory` | v1.12.1 | `modernc.org/libc` | Entra con `modernc.org/libc` |
| `github.com/remyoudompheng/bigfft` | v0.0.0-20230129092748-24d4a6f8daec | `modernc.org/mathutil` | Entra con `modernc.org/mathutil` |
| `github.com/dustin/go-humanize` | v1.0.1 | `modernc.org/libc` | Entra con `modernc.org/libc` |
| `github.com/google/uuid` | v1.6.0 | `modernc.org/libc` en darwin y linux | Entra con `modernc.org/libc`; no llega a windows |
| `github.com/mattn/go-isatty` | v0.0.24 | `modernc.org/libc` en darwin y windows | Entra con `modernc.org/libc`; no llega a linux |
| `github.com/ncruces/go-strftime` | v1.0.0 | `modernc.org/libc` en darwin y windows | Entra con `modernc.org/libc`; no llega a linux |
| `golang.org/x/sys` | v0.47.0 | `modernc.org/sqlite`, `modernc.org/libc`, `modernc.org/memory`, `github.com/mattn/go-isatty` | Entra con el controlador y con lo que este arrastra, para las llamadas al sistema |

Los nueve que llegan con el controlador son exactamente los indirectos que H3 midió y justificó al fijar
`modernc.org/sqlite` (`specs/004-h3-internal-cache-sqlite/gates/pr-h3.md`, «Dependencias»); H4 no cambia ninguna
versión. Los seis de H1 (`alecthomas/kong`, `invopop/jsonschema` y lo que el segundo arrastra) siguen igual.

**Constatado al cierre (T036, sobre 55cc107).** La primera orden de arriba, ejecutada hoy en darwin/arm64, da diecinueve líneas: el
propio módulo y los dieciocho de terceros de la lista —los doce de la tabla y los seis de H1 (`alecthomas/kong`,
`invopop/jsonschema`, `bahlo/generic-list-go`, `buger/jsonparser`, `pb33f/ordered-map/v2`, `go.yaml.in/yaml/v4`)—.
`TestDependenciasDelBinario` pasa (escenario 13.a). `git diff --stat main...HEAD -- go.mod go.sum codecov.yml` está
vacío: **ninguna dependencia nueva y ningún umbral tocado**.

## Controles añadidos

Los 22 controles de la tabla del plan («Controles mecánicos que este hito añade o toca») están en el árbol y en
`make ci`, salvo la verificación con red, que es nocturna. Lo que pasa a ser mecánico, con el escenario del quickstart
que lo demuestra:

- **La salida de los seis verbos contra `schemas/*.json`** (`TestSalidaDeBoeContraSchemas`): el sobre real de éxito y
  los de fallo 2, 3 y 4, validados contra la parte del verbo leída del fichero; `Aviso.codigo` con exactamente tres
  valores (escenario 7.a).
- **`make schema-check` deja de ser un aviso** (`TestEsquemasPublicados`, `TestEsquemasCubrenTodosLosVerbos`):
  regenera en memoria desde `--describe` y compara con lo versionado. Una edición a mano de `schemas/bloque.json` lo
  hace fallar nombrando el fichero y el verbo (escenario 7.b).
- **Golden de `data`** (`TestGolden`, `TestGoldenCubreTodosLosCasos`, 13 casos de 6 verbos): un byte distinto se
  detecta nombrando el fichero y la línea (escenario 8).
- **Corrección de citas contra la respuesta grabada** (`TestArticuloCoincideConBoePy`): seis referencias generadas
  con el `refs/boe.py` original sobre las grabaciones y revisadas campo a campo por una persona antes del código
  (escenario 2).
- **e2e de la entrega** (`TestEntregaDelHito/boe-{verbos,multicall,offline,codigos}`) sobre el binario de e2e con
  reproducción y sin red (escenario 5), y **< 200 ms en caché** (`boe-cache-rapida`, con `cronometra` y reproducción
  vacía; escenario 6).
- **Códigos 2, 3, 4 y 5 con su sobre y su fecha**: `TestCodigosDeSalidaDeBoe` (kernel en proceso, reloj controlado),
  `TestClasesDeErrorDeBoe` (una subprueba por fila del contrato, 56 situaciones, doce mutaciones sondadas en T024),
  `TestPedirClasificaEstados`; ningún caso produce el 6 (escenario 11).
- **Caché, `--offline`, ensayo y fallos no guardados** (`TestCacheDeLosSeisVerbos`, `TestOfflineDeLosSeisVerbos`,
  `TestEnsayoDeLosSeisVerbos`, `TestFallosNoSeGuardan`, `TestArticulos` con las peticiones contadas): la segunda
  consulta se resuelve con una reproducción estricta sobre un directorio vacío (escenario 3).
- **Fecha de consulta** (`TestFechaDeConsultaDeArticulo`, `TestMontadorFechaDeConsulta`, `TestInstanteDeEmision` con
  nueve subtests): fechar con el reloj del kernel lo servido de la caché falla (escenario 4).
- **Fuzz del analizador de ids de bloque** (`FuzzIDDeBloque`; semillas `a21`, `da3`, `dt1`, `a1-30`, `a85bis.` en
  `make test`) y **gramática contra los índices grabados** (`TestGramaticaCubreLosIndicesGrabados`): los tres índices
  reales casan enteros (escenario 9).
- **Grabaciones y referencias completas** (`TestGrabacionesCompletas`, `TestReferenciasCompletas`), **porte
  documentado** (`TestDocAnotaElPorte`) y **fuente atada a `docs/SOURCES.md`** (`TestFuenteCoincideConSources`: ritmo,
  términos y fecha; `pendiente` ya no se admite) (escenario 14).
- **Espacio reservado** (`TestLasFuentesNoFirmanComoKitlegal`, AST de `internal/source/**`, no pasa en vacío): un
  literal `kitlegal:applet/boe` en una fuente falla nombrando fichero y línea (escenario 13.b).
- **Superficie del binario** (`TestDependenciasDelBinario` con la lista ampliada y justificada) y **arquitectura**
  (`TestArquitectura`, `TestElBinarioNoEnlazaLosEjemplos`) (escenario 13.a).
- **`httpx` pide el formato de cada recurso** (`TestPedirConAcepta`, `TestConHoraRechazaNula`, `TestDirecciones`)
  (escenario 12).
- **Registro y arranque** (`TestRegistroDeProduccion`, `TestAppletBoe`, `TestArrancar`) y **`--no-graph` y `--asunto`
  sin efecto** (`TestSinGrafoNiAsuntoNoCambianLaSalida`) (escenario 11).
- **Verificación contra la fuente real**: `make verify-sources` (`TestVerificarFuentes`, etiqueta `fuentes`) en el
  trabajo nocturno `fuentes`, que abre o comenta la incidencia «verify-sources: boe articulo»; su control negativo sin
  red, `TestVerificacionDeFuentesDetectaCambios`, falla con el sintético `bloque-ilegible` nombrando «boe articulo»
  (escenario 15.a).
- **Lint de los ficheros etiquetados**: `run.build-tags: [integration, fuentes, grabacion]`, de modo que
  `grabacion_test.go` y `fuentes_red_test.go` no escapan a `make lint`.
- Los de H0-H3 (R1-R5, formato, `-race`, `govulncheck`, `gosec`, `gitleaks`, `go mod verify`, `tidy -diff`, CodeQL)
  siguen sin exclusiones nuevas.

## Evidencia

Medida por T036 el 2026-09-13 (entre las 22:12 y las 22:30, hora de Madrid) sobre **55cc107** (`feat(H4): T035`, el
HEAD de la rama y el último commit que toca código o datos), ejecutando `quickstart.md` entero salvo el 15.b, con cada
orden tal cual, desde los prerrequisitos hasta la limpieza, en una sesión desatendida. El árbol de trabajo llevaba
además, sin confirmar y dentro del directorio del feature, el `quickstart.md` corregido en el intento 1 de la misma
tarea (`gates/tarea-T036.md`: siete tuberías partidas con el `|` al principio de línea pedían aprobación; ningún test,
patrón ni filtro cambió). Ninguna orden pidió aprobación y las cinco sondas de los prerrequisitos dieron lo esperado
(`código 3`; `valor`; `1` y «fin de la tubería partida»; dos huellas distintas; nada entre la segunda huella y «fin de
las sondas»).

| Escenario | Resultado |
|---|---|
| Prerrequisitos | `go1.27.1 darwin/arm64`, rama `h4-applet-boe-puerto`, ninguna línea de `git status` fuera del directorio del feature, carpeta temporal vacía |
| 1 · suite entera | Huella de `~/.cache/kitlegal` **idéntica** antes y después (`519f40d4…`); `make test` sin `FAIL` ni `WARNING: DATA RACE`, `ok` en los diez paquetes; ninguna línea en el `git status` de `testdata/`, `schemas/` y `docs/SOURCES.md` |
| 2 · aceptación | Seis `--- PASS` de `TestArticuloCoincideConBoePy/…` (`BOE-A-2015-10565-a1`, `-a21`, `BOE-A-1985-5392-a22`, `BOE-A-2017-12902-a1-30`, `-da-3`, `BOE-A-1992-26318-a42`), ningún `FAIL` |
| 3 · caché | Cinco `--- PASS` (`TestCacheDeLosSeisVerbos`, `TestOfflineDeLosSeisVerbos`, `TestEnsayoDeLosSeisVerbos`, `TestFallosNoSeGuardan`, `TestArticulos`) |
| 4 · fecha de consulta | `TestFechaDeConsultaDeArticulo`, `TestMontadorFechaDeConsulta` y los nueve subtests de `TestInstanteDeEmision`, todos `PASS` |
| 5 · e2e | Cinco `--- PASS`: `boe-multicall`, `boe-offline`, `boe-codigos`, `boe-cache-rapida`, `boe-verbos` |
| 6 · < 200 ms | Ver abajo |
| 7.a · contratos | `make schema-check` → `ok  	github.com/jmorenobl/kitlegal/internal/app`; `TestEsquemasCubrenTodosLosVerbos` y `TestSalidaDeBoeContraSchemas` `PASS` |
| 10 · binario a mano | La línea `  boe  Consulta la legislación consolidada del BOE…` en la ayuda; `1` en `--describe`; sobre `kitlegal.cli`/`argumentos` y `código 2` con `BOE-A-2015`; por el enlace `boe`, sobre `boe.legislacion-consolidada` con la `url` del bloque, clase `fuente-no-disponible` y `código 4` con `--offline`; `--dry-run` con salida estándar vacía, `código 0` y tres líneas de error (`no se ha ejecutado nada…`, `GET …/texto/bloque/a21`, `GET …/metadatos`); «ni los argumentos inválidos, ni --offline ni --dry-run han creado la caché»; la sonda de ausencia sobre el enlace no imprime nada |
| 11 · códigos | Cinco `--- PASS` (`TestPedirClasificaEstados`, `TestClasesDeErrorDeBoe`, `TestAppletBoe`, `TestCodigosDeSalidaDeBoe`, `TestSinGrafoNiAsuntoNoCambianLaSalida`) |
| 12 · formato | Tres `--- PASS` (`TestPedirConAcepta`, `TestConHoraRechazaNula`, `TestDirecciones`) |
| 13.a · arquitectura | Cuatro `--- PASS` (`TestLasFuentesNoFirmanComoKitlegal`, `TestDependenciasDelBinario`, `TestElBinarioNoEnlazaLosEjemplos`, `TestArquitectura`); `make lint` → `0 issues.` |
| 14 · porte y fuente | Cinco `--- PASS`; la fila de `boe.legislacion-consolidada` en `docs/SOURCES.md` con ritmo `1s`, términos y «Revisado» `2026-09-13`; `1` |
| 15.a · verificación sin red | `--- PASS: TestVerificacionDeFuentesDetectaCambios` |
| 16 · veredicto | `make ci` → `ci: todos los controles en verde` (formato, lint con `0 issues`, `test` y `test-integration` con `-race`, `govulncheck` sin vulnerabilidades, `schema-check`, `gitleaks` «no leaks found», `go mod verify` en la raíz y los cuatro módulos de herramienta, `tidy -diff`); ninguna línea de `git status` fuera del directorio del feature |
| Limpieza | «carpeta temporal borrada» |

**Escenario 6, la salida de `boe-cache-rapida`** (`go test -count=1 -v -run '^TestEntregaDelHito$/^boe-cache-rapida$'
./internal/app/`, filtrada por `cronometra` y el veredicto), tal cual:

```text
        # cronometra mide la invocación entera, arranque del proceso incluido. La
        > cronometra 200ms $KITLEGAL_BIN boe articulo BOE-A-2015-10565 a21 --json
        > cronometra 200ms $KITLEGAL_BIN boe articulo BOE-A-2015-10565 a21 --json
        > cronometra 200ms $KITLEGAL_BIN boe articulo BOE-A-2015-10565 a21 --json
        > cronometra 200ms $KITLEGAL_BIN boe articulo BOE-A-2015-10565 a21 --json
        > cronometra 200ms $KITLEGAL_BIN boe articulo BOE-A-2015-10565 a21 --json
        > cronometra 200ms $KITLEGAL_BIN boe articulo BOE-A-2015-10565 a21 --json
        > cronometra 200ms $KITLEGAL_BIN boe articulo BOE-A-2015-10565 a21 --json
        > cronometra 200ms $KITLEGAL_BIN boe articulo BOE-A-2015-10565 a21 --json
        > cronometra 200ms $KITLEGAL_BIN boe articulo BOE-A-2015-10565 a21 --json
        > cronometra 200ms $KITLEGAL_BIN boe articulo BOE-A-2015-10565 a21 --json
    --- PASS: TestEntregaDelHito/boe-cache-rapida (0.41s)
```

Diez invocaciones consecutivas del binario compilado con la entrada vigente en caché, cada una medida entera
—arranque del proceso incluido— contra el máximo de 200 ms, con la reproducción vaciada antes de medir (una petición
haría fallar la invocación) y con la salida comparada byte a byte, `fecha_consulta` incluida, con la consulta que sembró
la caché. El guion completo —la siembra, el vaciado y las diez medidas— tardó 0,41 s en esta máquina (Apple Silicon,
sin detector de carreras: SC-002 mide de reloj sobre la invocación completa). En el ejecutor de la integración continua
es el supuesto S9 (*Pendientes*).

**Escenarios negativos, sobre el clon desechable** (`git clone --quiet . /tmp/kitlegal-quickstart-h4/copia`; el árbol
de trabajo no cambia y la limpieza lo borra):

- **7.b** `"minLength": 1` → `2` en `schemas/bloque.json` del clon: `git diff --stat` da `schemas/bloque.json | 2 +-`,
  `1 file changed`; `make -C … schema-check` falla con `TestEsquemasPublicados/schemas` y
  «`schemas/bloque.json: la parte de «articulo» no coincide con lo que emite kitlegal boe articulo --describe`».
- **8** Una `X` delante del `texto` del golden `articulo-BOE-A-2015-10565-a21.json`: `git diff --stat` nombra el golden
  con `2 +-` y `2 files changed`; `TestGolden/testdata-golden` falla con «no coincide byte a byte con lo que da su
  caso: la línea 10 es `"texto": "XArtículo 21. Obligación de resolver…`».
- **9** `go -C … test -run '^$' -fuzz '^FuzzIDDeBloque$' -fuzztime 10s ./internal/source/boe/`: 125 entradas de
  base, 387 058 ejecuciones con 8 trabajadores en 11 s, 44 entradas nuevas interesantes, `PASS` y sin `Failing input`;
  el corpus se queda en el clon. Después, `TestTipoDesdeID`, `TestValidarBloque` y
  `TestGramaticaCubreLosIndicesGrabados` en `PASS` sobre el árbol.
- **13.b** `var _ = "kitlegal:applet/boe"` al final de `direcciones.go` del clon (`tail -1` lo confirma):
  `TestLasFuentesNoFirmanComoKitlegal/ningún_paquete_de_fuentes_lleva_un_literal_del_espacio_reservado` falla con
  «`…/copia/internal/source/boe/direcciones.go:72: "kitlegal:applet/boe": un adaptador de fuente firma en el espacio
  reservado (kitlegal. y kitlegal:), que es del kernel y de los applets calculados…` (ADR 0006, FR-002)».

**Cobertura**, sobre el `coverage.out` y el `coverage-integration.out` que dejó el `make ci` del escenario 16
(`go tool cover -func` para el total; para cada árbol, la suma de sentencias del perfil, que es lo que Codecov mide):

| Umbral | Exigido | Perfil unitario (`make test`) | Unión de los dos perfiles (lo que Codecov une) |
|---|---|---|---|
| Global (`codecov/project`) | ≥ 70 % | **95,2 %** (3080/3236 sentencias) | **96,1 %** (3109/3236) |
| `internal/core/**` (componente `internal_core`) | ≥ 85 % | **90,1 %** (73/81) | **90,1 %** |
| `internal/cli/**` (componente `internal_cli`) | ≥ 90 % | **98,6 %** (348/353) | **98,6 %** |
| `internal/source/boe` | — | 99,4 % (875/880) | 99,4 % |
| `internal/app` | — | 87,8 % (445/507) | 87,8 % |
| `internal/cache` | — | 90,7 % | 96,2 % con `-tags=integration` |

`internal/core` no gana sentencias: sus 81 son las de `internal/core/schema` (`error.go`, `huella.go`, `sobre.go`);
`doc.go` y `source.go` no tienen ninguna. **Ningún umbral se rebaja**: `codecov.yml` no aparece en el diff frente a
`main`. `codecov/patch` (`target: auto`, la cobertura de la base) lo lee T037 en la plataforma.

**Sin ninguna supresión nueva**: `0` líneas `//nolint` y `0` `t.Skip` añadidas en ficheros `.go` frente a `main` (las
nueve apariciones de `//nolint` en el diff son prosa de `specs/` que dice que no hay ninguno). `gosec` se resuelve sin
supresiones: G304 con `filepath.Clean`, G306 con `0o600` y G703 escribiendo desde un auxiliar distinto del que lee.

**Sin red y sin tocar lo protegido**: ninguna tarea ejecutó `KITLEGAL_RECORD`, `scripts/grabar-fixtures.sh` ni
`make verify-sources`; las grabaciones, las referencias, la fila de `docs/SOURCES.md` y `terminos.go` los fijó una
persona en la pausa `[datos]` de T009 (commit `17fca5b`), como exige la obligación 2 del plan.

**Aviso de método.** El envoltorio de terminal de esta máquina reescribe la salida de `go test`, `git status
--porcelain` y `git diff`. Todo lo de arriba está medido con el paso directo (`rtk proxy`), que es la forma en que el
quickstart escribe cada orden.

## Decisiones

- **`fecha_consulta` la declara quien consulta** (ADR 0015). `schema.Procedencia` gana `FechaConsulta`; el montador
  fecha con ella en éxito y en fallo cuando la procedencia es válida y la fecha no es cero, y con su reloj si no. `boe`
  declara el instante en que `httpx` emitió la petición (el de su último intento) o el guardado en la entrada de caché;
  con varias consultas, el más antiguo. Alternativas rechazadas: la fecha en `data` rompe la huella (ADR 0006); el
  reloj del kernel fecha mal lo servido de la caché. La marca de emisión vive en `httpx` como decorador propio en las
  dos cadenas (`New` y `Replay`), porque solo dentro de la cadena se conoce el último intento; la obtención del
  `robots.txt` no fecha (T003, redelimitada para tocar `reproducir_test.go`, que compara instantes, y un comentario de
  `grabar.go`).
- **La firma de `core.Source`** es `Fetch(ctx, ec, consulta)`, `TTL(consulta)`, `Terms()`: `--offline`, `--dry-run` y la
  vigencia por verbo la exigen (ADR 0015). El applet `boe` no lee ninguna bandera: las recibe en la ejecución.
- **Cinco clarificaciones del spec** (sesión 2026-09-13): dos esquemas por entidad (`norma.json`, `bloque.json`) y no
  uno por verbo; los ocho campos del diff de aceptación (título, tipo, fecha de versión, fecha de vigencia, norma
  modificadora, texto, avisos por su `texto` y dirección pública); la `url` del sobre es la de la API del recurso
  principal, y en un fallo la de la petición que falló; los avisos como lista de `{codigo, texto}` con `codigo` de un
  enumerado cerrado de tres valores; y cómo se producen las referencias, **enmendada en la pausa de T009**: no se
  derivaron a mano, las generó un guion de un solo uso, fuera del repositorio y sin versionar, que importa
  `refs/boe.py` sin modificarlo y le sustituye la red por las grabaciones, y una persona las revisó campo a campo.
  Derivar a mano habría puesto la misma lectura de `boe.py` en los dos lados del diff. Ni el producto, ni los tests, ni
  `make ci` necesitan Python.
- **La gramática del id de bloque admite `-` y `.`** (`^[A-Za-z0-9][A-Za-z0-9.-]{0,63}$`), decidido por la persona en
  la pausa de T009: los índices reales de la LCSP, la LPAC y la LRBRL usan `a1-30`, `da-3`, `a85bis.`; con la gramática
  de FR-080 original casi toda la LCSP quedaba fuera de `articulo`. Son caracteres no reservados del RFC 3986 y las
  propiedades del fuzz siguen valiendo. Y **el diff pasa a seis artículos de cuatro leyes**: `a118` y `da3` de la LCSP
  no existen en la fuente (son `a1-30` y `da-3`), y `BOE-A-1992-26318 a42` entra como cuarta ley porque es el único con
  avisos de vigencia reales; los otros cinco los daban vacíos.
- **Donde el porte se aparta de `boe.py`**, con su motivo en `doc.go` (32 entradas): los errores salen con clase y
  código estable y nunca como texto con código 0; ningún fallo se guarda en caché; un tipo inesperado en un campo es
  «fuente no disponible» (4) y no un `repr` presentado como contenido (§II); un bloque cuya vigencia no se pudo
  comprobar no se emite; el texto de las referencias anteriores llega completo; cada bloque distinto se pide una vez y
  los metadatos como mucho una vez por invocación; `--dry-run` abre la caché en solo lectura para describir solo las
  peticiones que ocurrirían, sin crear `cache.db`.
- **El esqueleto de verbo se abstrae, no se disfraza** (T022): con `analisis`, tercer verbo con la misma forma que
  `metadatos` e `indice`, `dupl` marcó el esqueleto. Se recogió en `resolverRecursoDeLaNorma[T]` y los tres verbos
  quedaron en una línea con su dirección y su función de resolución; los tests de `metadatos` e `indice` no cambiaron.
- **«Caché intacta» es `cache.db` byte a byte** (T023): una lectura en solo lectura crea `-wal` vacío y `-shm`, y H3
  (`esquema-y-apertura.md` §6) fija que no cuentan.
- **La fila 21 del contrato de errores** (T024) se prueba con una `AperturaDeCache` de prueba cuya caché falla al abrir,
  leer, escribir o cerrar, con cada una de las tres clases de `cache.Error`: quince situaciones; el fallo de la lectura
  fuera de solo lectura es un fallo y no una ausencia, la única situación en que un error mal clasificado haría pedir a
  la fuente.
- **La migración de H3 nace sin la cancelación del contexto** (`context.WithoutCancel`): un contexto que termina
  durante `BeginTx` hacía que `database/sql` deshiciera la transacción desde otra goroutine y `New` volviera con la
  conexión abierta, en contra del contrato de apertura. Salió como fallo intermitente en el `make ci` de T003.
- **El binario de e2e registra `boe` con reproducción; el distribuido no admite ninguna**: aceptar una carpeta de
  reproducción en `kitlegal` permitiría falsificar citas.
- **`app.Arrancar` en vez de `panic`**: registrar el primer applet hace falible el registro, y §IV no admite pánicos;
  un registro que no se construye sale como sobre de clase «inesperado» con código 1.
- **`make verify-sources` en el `Makefile`** y no un script llamado desde el flujo: el `Makefile` es la única superficie
  de invocación y el nocturno usa el mismo `GOTOOLCHAIN`.
- **Seis palabras españolas en `misspell.ignore-rules`** (`administrativo`, `capitulo`, `dependencias`, `disposicion`,
  `materias`, `regulares`): valores de la fuente y del contrato que el diccionario inglés toma por erratas; renombrarlas
  cambiaría el contrato de `data` o el texto que FR-120 exige.
- **Cada orden del quickstart se ejecuta tal cual en la sesión desatendida** (T036): siete tuberías partidas con el `|`
  al principio de línea pedían aprobación y no se ejecutaban; se movió el salto de línea, la tabla de formas ganó la
  fila y una sonda de los prerrequisitos comprueba la forma en cada ejecución. El intento 1 se detuvo sin sustituir la
  orden, como manda la tarea.
- **Workflow 1.9.0 en la misma rama** (`5a631b5`): el commit de cada tarea se decide con el estado de la iteración,
  porque el motor conservaba una salida rancia y T003 dio 145 vueltas sin commit.

## Pendientes

- **Supuestos pendientes de la primera ejecución** (obligación 11 del plan; research D20):
  - **S8, incidencia nocturna**: que `permissions: issues: write` y `GH_TOKEN: ${{ github.token }}` basten para
    `gh issue list/comment/create` en el trabajo `fuentes` de `nightly.yml`. Se comprueba en la primera ejecución
    nocturna tras fusionar; si la incidencia no se abre o no se comenta, la causa es la configuración del flujo, no la
    verificación.
  - **S9, duración en integración continua**: que las diez invocaciones con caché caliente bajen de 200 ms en el
    ejecutor de la plataforma con la suite en paralelo y `-race` en el resto de paquetes (aquí, 0,41 s el guion entero).
    **Primera medida, a favor**: en el run `34781187264` de esta propuesta, `make test` dejó `internal/app` en `ok`
    (5,098 s, con `-race` y `-shuffle=on`). Ese paquete ejecuta `TestEntregaDelHito` con `boe-cache-rapida`, que no
    tiene condición de salto, aunque el registro, sin `-v`, no nombra el subtest. La ejecución en verde de T037 la
    repite. Si fallara allí y no aquí, sería una medida de máquina: se documenta antes de tocar nada, y nunca se rebaja
    el máximo del guion sin decisión humana.
- **El primer `ci` de esta propuesta salió en rojo, y el arreglo es T038.** `TestDependenciasDelBinario` medía la
  superficie del binario solo en la plataforma del ordenador que ejecuta el test. En darwin, donde se ejecutaron todos
  los `make ci` del hito, el binario enlaza los dieciocho módulos; en el ejecutor linux, dieciséis, y el test falló
  (ver «Dependencias»). Codecov no emitió ningún estado, porque la subida del perfil no se ejecuta tras un fallo. La
  lista declarada no cambia: T038 hace que el test mida las seis plataformas de distribución y compare la unión. El
  diseño está comprobado sobre un clon, con `make ci` en verde (`gates/tarea-T037.md`, `gates/evidencia-plataforma.md`).
- **T037 `[plataforma]`**, la última tarea: empujar la rama, abrir la propuesta de cambio con este fichero como cuerpo,
  esperar `ci` y los estados de Codecov (`project` ≥ 70 %, `internal_core` ≥ 85 %, `internal_cli` ≥ 90 %, `patch`
  con objetivo `auto`). `internal_core` e `internal_cli` ya tienen base en `main`, así que miden de verdad. En el
  intento 1 quedaron publicadas la rama y esta propuesta; el intento 2 publica T038 y lee los estados. Al final,
  pausa humana del workflow por `docs/SOURCES.md` y por el directorio nuevo `internal/source/boe` (rutas sensibles).
- **`refs/__pycache__/boe.cpython-311.pyc` no debe fusionarse.** Es el fichero compilado que CPython dejó al importar
  `refs/boe.py` desde el guion de un solo uso de la pausa de T009, y el commit `[datos]` `17fca5b` lo arrastró
  (43 205 bytes, binario). `.gitignore` no excluye `__pycache__/`. Está fuera de las rutas de T036 y de cualquier tarea
  restante, así que queda para la revisión humana antes de fusionar: retirarlo del índice y añadir `__pycache__/` a
  `.gitignore`, de modo que la única ejecución autorizada de `refs/boe.py` no vuelva a dejar rastro versionado.
- **Para H5**, ya anotado en `docs/PENDIENTES.md`: los tres directorios llamados `skills`, el peso de las skills
  vendorizadas y cómo llama cada skill al binario (el enlace `boe -> kitlegal` ya funciona: escenario 10).
- **Para H7** (ADR 0014): `boe` observa ids naturales (`BOE-A-…`, ELI, id de bloque, `hash_texto`) y los conserva en
  `data`; emitirlos como operaciones de grafo en el `Resultado` es de ese hito, no deuda de este.
- **De H3, todavía abierto para la revisión humana**: `issues.uniq-by-line: false` en `.golangci.yml`, para que el
  primer `make ci` nombre todas las reglas incumplidas de una línea. No lo usa ningún control de H4.
- **Lo que no se hizo, a propósito**: `make verify-sources` (escenario 15.b) no se ejecutó en este cierre porque es el
  único control con red y lo ejecuta el flujo nocturno; ninguna grabación se rehízo; ninguna fecha de revisión la
  escribió el ejecutor.
