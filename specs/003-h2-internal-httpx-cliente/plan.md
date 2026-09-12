# Implementation Plan: H2 · `internal/httpx`: cliente HTTP responsable + grabación de fixtures

**Branch**: `h2-internal-httpx-cliente` | **Date**: 2026-09-12 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/003-h2-internal-httpx-cliente/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de
`.specify/memory/constitution.md` y están registradas —decisión, alternativas y motivo— en
[research.md](./research.md). Toda afirmación sobre una herramienta o dependencia externa está
verificada en local y research.md cita dónde; lo que no se pudo verificar está declarado como supuesto en
[research.md D22](./research.md#d22--supuestos-no-verificados), no como hecho.

## Summary

H1 dejó el kernel: un applet declara un verbo y devuelve un `Resultado`; banderas, sobre, huella y códigos
de salida vienen dados. H2 construye **la única puerta a la red** y la mitad del mecanismo de tests que el
principio III exige: un cliente que no se puede usar mal y una grabación/reproducción de respuestas que
hace que `make ci` pase sin Internet. No entrega ningún verbo ni toca ninguna skill (principio VIII:
fundación que protege a las skills que vendrán).

Cuatro decisiones sostienen el diseño:

1. **Una sola operación, y exige los dos contextos** ([D1](./research.md#d1--superficie-pública-un-tipo-dos-constructores-una-operación-cinco-opciones)):
   `Pedir(ctx, ejecucion, Peticion{Metodo, URL})`. El `context.Context` acota el plazo (FR-003 a FR-005) y
   el `schema.Contexto` trae `--dry-run` (FR-050); ninguno se puede olvidar porque no existe otra
   operación, y el paquete no exporta ningún tipo de `net/http` (FR-002). La respuesta es un tipo propio
   con el cuerpo ya leído, de modo que el único `*http.Response` del módulo se cierra dentro de
   `internal/httpx`, donde `bodyclose` lo ve ([D2](./research.md#d2--la-respuesta-cuerpo-ya-leído-cabeceras-propias-dirección-final-y-declaración-de-ensayo)).
2. **Cadena de decoradores privada** `identificar → robots → reintentar → ritmo → (grabar → transporte |
   reproducir)`, con las redirecciones y la clasificación fuera de la cadena, en `Pedir`, para que cada
   salto pase por la cadena entera —`robots.txt` y ritmo del sitio de destino— y cada reintento espere su
   turno ([D3](./research.md#d3--cadena-de-decoradores-sobre-httproundtripper-y-qué-queda-fuera-de-la-cadena)).
3. **Errores tipados sin importar el kernel** ([D4](./research.md#d4--errores-tipados-httpxerror-declara-su-clase-el-kernel-la-reconoce-con-errorsas-sin-que-el-adaptador-importe-internalcli)):
   `httpx.Error` declara su clase por la interfaz `schema.ConClase`, nueva en el dominio, y
   `cli.Clasificar` la reconoce con `errors.As` después de los cinco sentinelas de H1. `Retry-After` viaja
   en el error (FR-030). Las 16 situaciones de fallo tienen clase y código fijados en
   [`contracts/errores-y-ensayo.md`](./contracts/errores-y-ensayo.md) §3.
4. **`--dry-run` describe por capas** ([D6](./research.md#d6----dry-run-pedir-devuelve-una-respuesta-de-ensayo-la-descripción-viaja-en-schemaresultadoensayo-y-la-presenta-el-kernel)):
   `Pedir` devuelve una respuesta que declara que no se emitió nada; la descripción («`GET http://…`»)
   viaja en `schema.Resultado.Ensayo` y la presenta el kernel por el presentador, siempre visible, con
   código 0 (FR-051, FR-065). Es la implementación de lo que el contrato de H1 ya decía: «cada capa con
   efectos la honra describiendo en lugar de ejecutar».

Todo el código nuevo vive en `internal/httpx` (producto y tests, adaptador de prueba incluido) más tres
ampliaciones acotadas del kernel de H1 (`schema.ConClase`, `Resultado.Ensayo`, `cli.Clasificar`, la
presentación del ensayo en `internal/app`), el endurecimiento de R2 en `internal/arch_test.go`, una línea
de `LDFLAGS` en el `Makefile` y un ADR corto. El binario distribuido no cambia y no enlaza el paquete
(SC-014).

## Technical Context

**Language/Version**: Go 1.27, sin cambios: `go.mod` mantiene `go 1.27.0` + `toolchain go1.27.1` y el
`Makefile` sigue exportando ese `GOTOOLCHAIN`. Verificado: `go version` → `go1.27.1 darwin/arm64`. Las
dos dependencias nuevas exigen menos (`x/time` declara `go 1.26.0`; `robotstxt`, `go 1.11`).

**Primary Dependencies**: **dos** módulos nuevos, los dos de la lista cerrada de la constitución §V y de
`docs/ROADMAP.md` §3 (FR-058): `golang.org/x/time` v0.16.0 (paquete `rate`: limitador por sitio) y
`github.com/temoto/robotstxt` v1.1.2 (interpretación de `robots.txt`). **Ninguna otra.** Las dos están
descargadas en la caché de módulos y su API se ha leído (research, tabla de verificación); ninguna
arrastra módulos nuevos (`x/time` no tiene dependencias; `robotstxt` solo requiere `testify` v1.3.0, que
la selección mínima de versiones resuelve con el v1.12.1 ya presente; supuesto S1 de D22 hasta ejecutar
`go mod tidy`). Ninguna de las dos llega al binario distribuido en H2: ningún paquete de producción
importa `internal/httpx` todavía, y `TestDependenciasDelBinario` lo fija. Todo lo demás es biblioteca
estándar: `net/http`, `net/url`, `context`, `log/slog`, `encoding/json`, `crypto/rand` y
`encoding/binary` (jitter de los reintentos, D9: `gosec` G404 marca `math/rand` **y** `math/rand/v2`,
verificado en `rules/rand.go`, y no hay supresiones; `crypto/rand.Read` no devuelve nunca error, así que
no deja ninguna rama que manejar ni que cubrir), `sync`, `crypto/sha256`, `unicode/utf8`,
`encoding/base64`, `os`, `path/filepath`, `time`. Ni `math/rand`, ni `math/rand/v2`, ni `math/big` se
importan en ningún fichero del hito.

**Herramientas de control**: las de H0/H1 sin cambio de versión (`golangci-lint` v2.13.2, `govulncheck`,
`gitleaks`, `lefthook`). H2 no añade ninguna, no activa ningún linter (`noctx` y `bodyclose` están
activos desde H0; por primera vez tienen código que vigilar) y no añade ninguna exclusión ni `//nolint`
(SC-010). Lo que cada uno prohíbe está verificado en su código (`noctx.go`, README de `bodyclose`,
`paralleltest.go`).

**Storage**: ficheros JSON de grabación bajo la raíz que declara quien construye el cliente
(`<raíz>/<fuente>/<nombre>.json`), solo con `KITLEGAL_RECORD=1` y, en este hito, solo en directorios
temporales de los tests (FR-044). Nada en `~/.cache/kitlegal/` (H3) y ninguna base de datos.

**Testing**: `go test -race -shuffle=on -coverprofile=coverage.out ./...` (`make test`, sin cambios).
Unitarios de tabla contra `httptest.NewTestServer(t, h)` + `Start()` (bucle local, limpieza registrada)
en `internal/httpx`; fixtures de reproducción escritos a mano en `internal/httpx/testdata/reproduccion/`
contra el host ficticio `fuente.prueba`; el adaptador de prueba en `package httpx_test` ejercitado con
`app.Main` **en proceso** contra un servidor local (US6). **Cero red en todos los tests** y **ninguna
grabación real** (FR-044). Sin guion `testscript` nuevo: el spec lo excluye porque H2 no cambia ningún
comportamiento visible del binario (FR-062, SC-014).

**Target Platform**: sin cambios (binario sin cgo, `-trimpath`; CI en `ubuntu-latest`; desarrollo en
darwin/arm64). El nombre de fichero de grabación es portable a Windows (sin `<>:"/\|?*`) y la colisión
se detecta por contenido, lo que cubre los sistemas de ficheros insensibles a mayúsculas (D12).

**Project Type**: paquete interno de una herramienta de línea de órdenes multicall; ningún applet nuevo.

**Performance Goals**: ninguna propia. Se conserva la de H0 (`ci` de una PR limpia < 3 min en caliente).
Los tests del hito no esperan de verdad: el reloj de los reintentos se inyecta desde el propio paquete
(D9) y los intervalos de ritmo en test son de decenas de milisegundos.

**Constraints**: por omisión el cliente es **lento a propósito** —una petición por segundo y sitio, tres
intentos, un solo cubo— y no existe ninguna opción para desactivar identificación, `robots.txt`, ritmo,
reintentos, plazo del contexto ni verificación TLS (FR-002, FR-008, FR-012). `make ci` no modifica ningún
fichero versionado ni deja nada bajo `testdata/` (SC-008). Ningún `panic` en ninguna ruta de fallo del
paquete (FR-033; el único suceso que termina el proceso —un fallo de la fuente de entropía del sistema
operativo, que el toolchain trata como irrecuperable— es ajeno al paquete y está declarado en D9, no
disimulado con una contingencia inalcanzable). Sin goroutines propias ni paralelismo (FR-021).

**Scale/Scope**: un paquete nuevo con 14 ficheros de producto y sus tests (del orden de 900-1200 líneas
de Go de producto y otras tantas de test), cuatro ficheros de H1 tocados en pocas líneas, un test de
arquitectura endurecido, un ADR, `go.mod`/`go.sum`, `Makefile`, `.golangci.yml` (comentarios),
`docs/PENDIENTES.md`.

## Constitution Check

*GATE: debe pasar antes de la fase 0 y volver a evaluarse tras la fase 1.*

### Principios

| # | Principio | Cómo lo cumple H2 | Veredicto |
|---|---|---|---|
| **I** | Fuentes públicas y frontera humana | Es **el** principio que H2 convierte en código y en imposibilidad: `User-Agent` identificable en el 100 % de las peticiones, `robots.txt` antes de la primera petición a cada sitio, ritmo por sitio con `x/time/rate`, sin paralelismo propio, y **solo GET y HEAD** —cualquier otro método es un error de argumentos antes de abrir la conexión (FR-010)—. No hay forma de construir un cliente sin esas garantías ni de emitir un `*http.Request` desde fuera del paquete (FR-002, D1, D3). Un 429 nunca se reintenta ni se rodea (FR-030); un `robots.txt` que no se puede obtener por 5xx o 429 **deniega** (FR-015). Ninguna petición sale hacia una sede ni con identidad; el código 6 no se produce (FR-063). CENDOJ ni se nombra. | ✅ Cumple |
| **II** | Nada sin cita ni fuente | H2 no emite ningún sobre —no hay applet—, pero prepara la cita: `Respuesta.URL` es la **dirección final** tras las redirecciones (D2), que es la `url` que un adaptador podrá poner en su `Procedencia` sin inventarla, y la grabación conserva dirección completa, estado y cuerpo (FR-037), la verdad terreno contra la que H4 comprobará sus citas. No se genera ni interpreta contenido legal. | ✅ Cumple |
| **III** | Tests primero y offline | H2 **aporta el mecanismo**: `KITLEGAL_RECORD=1` graba en `<raíz>/<fuente>/<nombre>.json` y `Replay(dir)` responde sin red (FR-036 a FR-049). Todos los tests del hito son offline por construcción: servidores de bucle local levantados por el propio test y un directorio de reproducción escrito a mano; ningún test conoce una dirección externa (quickstart, prerrequisitos). No se graba ningún fixture real (FR-044) y `make ci` no toca `testdata/` (SC-008, medida diferencial). Sin e2e nuevo, por la razón que el spec fija (no hay comportamiento visible que describir); el guion de H1 sobre `--dry-run` sigue vigente. Cobertura: rigen los umbrales generales (SC-015) sin componente nuevo ni exclusión. El orden «test primero» se concreta por fichero: cada tarea escribe el test de tabla antes que el código que lo hace pasar (§«Orden de implementación»). | ✅ Cumple |
| **IV** | Arquitectura hexagonal con reglas ejecutables | `internal/httpx` es un **adaptador** de la lista de §IV; el dominio no lo importa (R1, FR-055) y él importa `internal/core/schema` solo por `Contexto` y `Clase`. R2 pasa de «activa y vacía» a **activa con dueño**, vigilada en las dos capas y con el test endurecido para no pasar en vacío (FR-053, D18). Errores tipados que mapean a códigos estables: `httpx.Error` + `schema.ConClase` + `cli.Clasificar`, con la tabla cerrada de 16 situaciones → {1, 2, 4, 5} (D4, contrato de errores §3); nunca 3 ni 6 (FR-063). Ningún `panic` en ninguna ruta de fallo del paquete (FR-033; D9 declara el único suceso ajeno que termina el proceso sin fingir que se maneja). Composición manual en los constructores (opciones funcionales), sin framework. `log/slog` a stderr por el registrador recibido; `internal/httpx` no escribe en stdout (R5). | ✅ Cumple |
| **V** | Simplicidad y dependencias fijadas (YAGNI) | Las dos dependencias nuevas están en la lista de §V y son las que el hito nombra (FR-058); ninguna otra, ni transitiva (D22 S1). Sin DI, sin ORM, sin generador. Todo lo del spec en *Fuera de alcance* queda fuera del plan: sin caché, sin `--offline`, sin puerto `Fetcher`, sin `Source`, sin ritmo por sitio configurable, sin `Crawl-delay`, sin cookies ni autenticación, sin proxy propio, sin límite de tamaño, sin condicionales, sin `Emit`, sin e2e, sin extracción. Las divergencias de forma respecto de la lectura literal están en *Complexity Tracking*, ninguna añade una capacidad. | ✅ Cumple |
| **VI** | Un binario, convenciones de agente | El binario distribuido **no cambia** (FR-062, SC-014): mismo registro vacío, mismos verbos, mismo `version`; el de e2e tampoco. `--timeout` se hereda como plazo del contexto (FR-004) y `--dry-run` recibe su semántica de red (FR-050 a FR-052, FR-065) manteniendo lo que H1 fijó: describir en stderr y terminar con 0. `--offline`, `--no-graph` y `--asunto` siguen solo propagados. Ninguna bandera nueva. | ✅ Cumple |
| **VII** | Grafo y privacidad | Sin grafo (H17) y sin `Emit`. Privacidad: el cliente registra eventos solo a nivel `debug` (D15) y sus mensajes de error nombran sitio y ruta, nunca datos de personas; las grabaciones guardan lo que la fuente pública respondió y solo se escriben con la variable activa y en la raíz que se declara; ningún dato persiste fuera de eso. | ✅ Cumple |
| **VIII** | Skills primero; el binario es la herramienta | H2 es un hito de **fundación** de los que §VIII admite: no entrega skill, pero protege a todas las que vendrán, porque cada adaptador de fuente de cada skill (H4 `boe`, H12 `placsp`, H13 `bdns`, H14 `boletin`…) pasará por este cliente. No se construye ninguna herramienta que ninguna skill vaya a usar: el único adaptador del hito es de prueba y vive en material de test (FR-060). | ✅ Cumple |
| **IX** | Genericidad territorial, validación local | H2 no tiene dimensión territorial (spec, *Fuera de alcance*): ningún municipio, boletín ni territorio aparece en código, `data/` ni fixtures. El host ficticio de los fixtures es `fuente.prueba`. El punto 11 de la Definition of Done no aplica. | ✅ Cumple |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución §IV)

| Regla | Situación en H2 | Cómo se hace cumplir | Veredicto |
|---|---|---|---|
| `internal/core/**` no importa `internal/{source,httpx,cache,store,graph,render,cli,app}` ni I/O de la biblioteca estándar | **Sin cambio**: `internal/core/schema` gana `ConClase` (interfaz con `error` embebido) y `Resultado.Ensayo []string` sin importar nada nuevo. La dependencia va `httpx → schema`, nunca al revés (FR-055). | `depguard` lista `core` + `compruebaDominioPuro` en `internal/arch_test.go` (H1, sin cambios) | ✅ Cumple |
| Solo `internal/httpx` importa `net/http` | **Activa con dueño**: `internal/httpx` es el único importador de `net/http` (y de `httptest` en sus tests). Ningún otro paquete lo importa, ni siquiera `cmd/kitlegal/main_test.go`, que importa `internal/httpx` y no `net/http`. | `depguard` lista `red` (`**` salvo `!**/internal/httpx/**`, sin cambios) + subprueba R2 del test de arquitectura, que **exige** que el dueño esté en el grafo (D18); demostrado rompiéndola en copia desechable (quickstart 10) | ✅ Cumple |
| Solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | Activa y vacía (H3). `internal/httpx` no importa `database/sql`. | `depguard` lista `sql` + subprueba R3 (sin cambios) | ✅ Cumple |
| Solo `internal/cli` y `cmd/` llaman a `os.Exit` | Sin cambio: `internal/httpx` no llama a `os.Exit`; el adaptador de prueba tampoco, porque se ejercita con `app.Main` en proceso. | `forbidigo` `^os\.Exit$` (sin cambios) | ✅ Cumple |
| Solo `internal/render` escribe en stdout; logs a stderr | Sin cambio: `internal/httpx` no nombra `os.Stdout`/`os.Stderr` ni usa `fmt.Print*`; escribe únicamente por el `*slog.Logger` recibido (FR-035). La descripción de `--dry-run`, que debe ser visible sin depender del nivel, la presenta el **kernel** por el presentador (D6), no el cliente. | `forbidigo` `^fmt\.Print(\|f\|ln)$`, `^os\.Stdout$`, `^os\.Stderr$` (sin cambios) | ✅ Cumple |

### Gates mecánicos (constitución «Gates», capa 1)

Los que este hito **activa por primera vez o endurece**: tests offline contra un directorio de
grabaciones (`Replay`) y el modo de grabación que los produce; `noctx` y `bodyclose` con código que
vigilar; la subprueba R2 del test de arquitectura con dueño y sin paso en vacío; fixtures de códigos de
salida para las 16 situaciones de fallo del cliente (SC-006, SC-016); `-race` sobre un cliente usado desde
varias goroutines (SC-011); coherencia de la versión de la identificación con la del binario (D5).

Los que **siguen sin objeto**, con su hito: corrección de citas contra la respuesta grabada del BOE (H4);
`schemas/*.json` y su deriva (H4/H10); frontmatter y `references/` de skills (H5); matriz territorial
(H7). Ninguno se simula.

Los de H0 y H1 siguen todos activos y **sin ninguna exclusión nueva** (SC-010).

**Reglas del modo desatendido** (constitución): `KITLEGAL_RECORD=1` **no se ejecuta** en ninguna tarea
del hito fuera de `t.Setenv` dentro de un test contra un servidor local y un `t.TempDir()`; las tareas que
toquen `internal/httpx/testdata/` llevan `[datos]` (`docs/WORKFLOW.md`: material nuevo bajo
`internal/<pkg>/testdata/` exige etiqueta y guardián, sin pausa); ninguna tarea toca `testdata/` de raíz,
`schemas/`, `docs/SOURCES.md` ni `internal/source/`.

**Veredicto del gate: PASA.** No hay ninguna violación. Las siete divergencias conscientes están en
*Complexity Tracking*; ninguna afecta a alcance, frontera humana, privacidad, términos de uso ni a una
decisión cerrada.

### Re-evaluación tras la fase 1 (diseño)

El diseño ([data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md))
no añade ninguna capacidad sobre lo evaluado. Lo que apareció al bajar a diseño y roza un principio:

- **`schema.Resultado.Ensayo` y `schema.ConClase`** (§IV, dominio puro): son una interfaz y un campo,
  sin importaciones; el dominio sigue sin I/O. Cambian el contrato del applet de ADR 0005, por lo que se
  registra `docs/ADR/0011-ensayo-por-capas.md` (D6). Es el caso que el spec prevé: «solo haría falta un
  ADR si este hito se apartara de lo ya decidido».
- **`cli.Clasificar` gana una comprobación** (§IV, traducción única y exhaustiva): la tabla
  `codigoDeClase` no cambia y sigue sin `default`; una clase declarada fuera del vocabulario cae en
  «inesperado» (D4). Cobertura de `internal/cli` ≥ 90 % con los casos nuevos en `TestClasificar`.
- **Dirección inválida → clase 2** (§IV, ningún fallo sin clase): situación no enumerada por el spec,
  resuelta con el mismo criterio de FR-063 y declarada (D19, contrato de errores §3 fila 10).
- **El transporte hereda el proxy del entorno** (§V, YAGNI): comportamiento por omisión de
  `http.DefaultTransport`; H2 ni lo añade ni lo retira (D11). Sin código propio.
- **`Replay` acepta opciones** (§V): solo `ConFuente` y `ConRegistrador` tienen sentido; el resto se
  rechaza con clase 2, que es lo que FR-043 exige comprobar.
- **`crypto/rand.Read` para el jitter** (§V y restricciones técnicas, `gosec`): G404 de `gosec` v2.28.0
  marca `math/rand/v2` igual que `math/rand` (verificado en `rules/rand.go` 31-36), así que la única
  familia que pasa `make ci` sin supresión es `crypto/rand`; dentro de ella se elige `Read` sobre ocho
  bytes reducidos con `encoding/binary` (D9) porque **no devuelve nunca error** —`errcheck` y G104 lo
  saben y lo excluyen por omisión, verificado en su código— y no deja ninguna rama inalcanzable. Un
  fallo de la fuente de entropía del sistema operativo termina el proceso con **cualquier** función de
  `crypto/rand` en go1.27.1 (`sysrand.Read` es irrecuperable por diseño; verificado en el código del
  toolchain, tabla de verificación de research.md); la documentación lo declara imposible salvo en Linux
  anterior a 3.17 y AIX. No es una ruta de fallo del paquete y el plan **no** afirma que FR-033 lo cubra:
  la versión anterior de este documento presentaba una contingencia (`b/2` con registro a `debug`) que
  era código inalcanzable, y se ha retirado.

**Veredicto tras el diseño: PASA**, sin ninguna violación.

## Project Structure

### Documentation (this feature)

```text
specs/003-h2-internal-httpx-cliente/
├── plan.md              # Este fichero
├── research.md          # Fase 0: 21 decisiones con alternativas y motivo, verificación y supuestos
├── data-model.md        # Fase 1: entidades, invariantes, flujo de una operación
├── quickstart.md        # Fase 1: guía de validación ejecutable, 12 escenarios
├── contracts/
│   ├── cliente-httpx.md            # La superficie exportada, opciones, garantías y valores por omisión
│   ├── formato-de-grabacion.md     # El fichero JSON, el nombre derivado, colisión y emparejamiento
│   ├── errores-y-ensayo.md         # httpx.Error, schema.ConClase, la tabla de 16 situaciones, --dry-run
│   └── reglas-de-arquitectura.md   # R1-R5 en H2, noctx y bodyclose, qué falla ante cada violación
├── spec.md
├── checklists/
├── gates/
└── tasks.md             # Fase 2 (`/speckit-tasks`, no lo crea este comando)
```

### Source Code (repository root)

Estructura de `docs/ROADMAP.md` §2 y `CLAUDE.md`. En negrita lo que H2 crea; lo demás existe de H0/H1 y
se indica qué cambia.

```text
internal/
├── httpx/                     **NUEVO** el adaptador de red: único importador de net/http (R2)
│   ├── doc.go                   qué garantiza el paquete y por qué no se puede usar mal
│   ├── cliente.go               Cliente, New, Replay, Opcion y las cinco opciones, Pedir (validación,
│   │                            ensayo, bucle de redirecciones, clasificación del resultado)  (D1, D3, D10)
│   ├── peticion.go              Peticion, Respuesta, Cabeceras, Descripcion()               (D2)
│   ├── errores.go               Error: clase, petición, estado, Retry-After, Unwrap, Clase() (D4, D20)
│   ├── agente.go                version (por -ldflags), AgenteDeUsuario()                    (D5)
│   ├── sitio.go                 clave de sitio (esquema+host+puerto), mapa de sitios con exclusión (D7, D14)
│   ├── identificar.go           decorador: User-Agent en toda petición                       (D3)
│   ├── robots.go                decorador: obtención (con su cadena de redirecciones), casos FR-015,
│   │                            caché por sitio, evaluación con temoto/robotstxt             (D7)
│   ├── reintentos.go            decorador: 5xx/transporte, equal jitter, espera interrumpible, cierre
│   │                            de cuerpos, reloj inyectable                                  (D9)
│   ├── ritmo.go                 decorador: rate.Limiter por sitio, Wait(ctx)                 (D8)
│   ├── grabar.go                decorador: fichero JSON estable, colisión, escritura atómica  (D12)
│   ├── nombre.go                nombre de fichero derivado de método y dirección             (D12)
│   ├── reproducir.go            RoundTripper de reproducción: lectura, emparejamiento, fallos (D13)
│   ├── transporte.go            clon de DefaultTransport, http.Client sin plazo ni redirecciones (D11)
│   ├── *_test.go                un fichero de test por fichero de código (doc.go no tiene: solo lleva
│   │                            el comentario de paquete), mismo orden; más ensayo_test.go,
│   │                            superficie_test.go y adaptador_test.go (package httpx_test, tipo
│   │                            adaptadorDePrueba)                                            (D16)
│   └── testdata/              **NUEVO** material de test, no fixtures reales; tareas [datos]  (D17)
│       └── reproduccion/prueba/   diez grabaciones escritas a mano contra http://fuente.prueba,
│                                  lista cerrada en §«Fixtures»
├── arch_test.go               ← R2 exige que exista internal/httpx; TestElBinarioNoEnlazaHTTPX (D18)
├── core/schema/
│   ├── error.go                 ← + ConClase                                                (D4)
│   └── sobre.go                 ← Resultado gana Ensayo []string                            (D6)
├── cli/
│   ├── errors.go                ← Clasificar: errors.As sobre schema.ConClase tras los sentinelas (D4)
│   └── errors_test.go           ← casos nuevos en TestClasificar
├── app/
│   ├── main.go                  ← conDescripcion presenta las líneas de Resultado.Ensayo    (D6)
│   └── main_test.go             ← TestDryRunPresentaElEnsayo
cmd/kitlegal/
└── main_test.go               ← TestVersionDeLaIdentificacion (importa httpx solo en test) (D5)

Makefile                       ← LDFLAGS += -X <módulo>/internal/httpx.version=$(VERSION)   (D5)
.golangci.yml                  ← solo comentarios: la lista `red` ya tiene dueño
go.mod, go.sum                 ← + golang.org/x/time, + github.com/temoto/robotstxt
docs/ADR/0011-ensayo-por-capas.md   **NUEVO** (D6)
docs/PENDIENTES.md             ← «Antes de H2 · Dónde viven los fixtures» → «Antes de la primera tarea
                                 [datos] de H4», con la misma recomendación (D17)
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 sin desviaciones: `internal/httpx` es el adaptador de
red que el diagrama sitúa junto a `cache`, `store`, `graph` y `render`, con la cadena de decoradores que
el propio roadmap le asigna. No se crean `internal/source/`, `internal/cache` ni `pkg/`: están fuera de
alcance y crearlos vacíos sería un marcador de posición. El único directorio de test nuevo,
`internal/httpx/testdata/`, sigue la convención de Go (fixtures junto al paquete); la decisión sobre los
fixtures reales de fuentes queda para H4 (D17).

## Controles mecánicos que este hito añade o toca

Según la sección «Gates» de la constitución. Cada fila dice qué orden lo ejecuta, si está en `make ci` y
**qué falla si el control se retira**.

| # | Control | Herramienta / forma | Orden | ¿En `make ci`? | Demostración de que está activo |
|---|---|---|---|---|---|
| 1 | **R2 con dueño** | subprueba R2 de `TestArquitectura` exige `internal/httpx` en el grafo y ningún otro importador de `net/http` | `test` | Sí | Borrar `internal/httpx` en una copia hace fallar R2 «el grafo no contiene internal/httpx»; un `import "net/http"` en `internal/app` falla en `depguard` **y** en R2 (quickstart 10.a, 10.d) |
| 2 | **`noctx`** | activo desde H0; prohíbe `http.Get/Head/…`, `http.NewRequest`, `httptest.NewRequest` (verificado en `noctx.go`) | `lint` | Sí | `http.NewRequest` en `internal/httpx` → «must not be called. use net/http.NewRequestWithContext» (quickstart 10.b) |
| 3 | **`bodyclose`** | activo desde H0; exige cerrar `res.Body` de todo `*http.Response` | `lint` | Sí | Un `*http.Response` sin cerrar → «response body must be closed» (quickstart 10.a) |
| 4 | **Superficie sin `net/http`** | `TestSuperficieExportada`: `go/parser` + `go/ast` sobre las declaraciones exportadas del paquete | `test` | Sí | Exportar un método que devuelva `*http.Response` hace fallar el test (FR-002) |
| 5 | **Contexto obligatorio por construcción** | el compilador: `Pedir(ctx, ejecucion, p)` es la única operación; el adaptador de prueba es `package httpx_test` | `test` (compilación) | Sí | Llamar a `Pedir` sin `ctx` no compila (quickstart 10.c; FR-061, SC-012) |
| 6 | **Identificación en el 100 % de las peticiones** | `TestIdentificacionEnTodaPeticion`: el servidor cuenta peticiones sin la cabecera exacta (recurso, `robots.txt`, reintentos, saltos) | `test` | Sí | Quitar el decorador `identificar` de la cadena hace fallar el test con N > 0 (SC-001) |
| 7 | **Versión coherente con `version`** | `TestAgenteDeUsuario` (forma exacta con una expresión regular construida sobre la **variable** `version`, no sobre el literal `dev`, con el sufijo ` (+https://ventanillalegal.es/bot)` pasado por `regexp.QuoteMeta` en vez de escapado a mano, y `t.Log` de la identificación, D5) + `TestVersionDeLaIdentificacion` en `cmd/kitlegal` + `LDFLAGS` del `Makefile` | `test` + `build` | Sí | Cambiar el valor por omisión de una de las dos variables hace fallar el test de `cmd/kitlegal`; con `-ldflags` el registro del test muestra la versión inyectada sin fallar, y quitar la `-X` de `LDFLAGS` se ve en quickstart 2 (FR-007) |
| 8 | **`robots.txt`** | `TestRobotsDeniegaLaRuta`, `TestRobotsSePideUnaVezPorSitio`, `TestRobotsCasosDeObtencion` (tabla FR-015), `TestRobotsRedirigido`, `TestRobotsNoSeEvaluaASiMismo`, `TestRobotsPorSitio` | `test` | Sí | Sin la caché, diez rutas producen diez `GET /robots.txt`; sin la intercepción del 429, `FromStatusAndBytes` permitiría (SC-003) |
| 9 | **Ritmo por sitio y clave de sitio** | `TestClaveDeSitio` (tabla sobre los hosts ficticios `fuente.prueba` y `otra.prueba`: esquema, host en minúsculas, puerto explícito o por omisión 80/443, `http://fuente.prueba` ≠ `https://fuente.prueba`, `fuente.prueba` ≠ `otra.prueba`, dos puertos de `127.0.0.1` distintos, ruta y consulta fuera de la clave), `TestRitmoSeparaPeticionesDelMismoSitio`, `TestRitmoNoRetrasaOtroSitio`, `TestRitmoRespetaElContexto` | `test` | Sí | Con ráfaga > 1 la segunda petición llega sin separación; con clave por host sin puerto, `TestClaveDeSitio` falla en la fila de los dos puertos y el segundo servidor espera en `TestRitmoNoRetrasaOtroSitio` (SC-004, FR-013, FR-015, FR-019) |
| 10 | **Reintentos** | `TestReintentosDosErroresYUnAcierto` (tres peticiones, esperas crecientes y distintas), `TestReintentosNoRepite4xx`, `TestReintentosCancelacionGana`, `TestReintentosCierraCuerposDescartados`, `TestReintentosAgotados` | `test` | Sí | Sin jitter, dos ejecuciones dan esperas iguales; con *full jitter*, no crecen (SC-005) |
| 11 | **Fixtures de códigos de salida** | `TestClasesDeError`: 16 subpruebas, una por fila del contrato §3, comprobando `cli.Clasificar` y `cli.CodigoSalida` sobre el error real; `TestClasificar` (H1) gana los casos de `ConClase` | `test` | Sí | Quitar la comprobación `errors.As` de `Clasificar` hace salir las 16 con código 1 (SC-006, SC-016) |
| 12 | **Grabación** | `TestGrabarEscribeElFichero`, `TestGrabarNoAlteraLaRespuesta`, `TestGrabarFormatoEstable`, `TestGrabarSinVariableNoEscribe`, `TestGrabarConfiguracionIncompleta`, `TestGrabarRaizInvalida`, `TestGrabarColision`, `TestGrabarValorDeVariableInvalido`, `TestGrabarCuerpoBinario`, `TestNombreDeGrabacion`; todos sobre `t.TempDir()` y `t.Setenv` | `test` | Sí | Escribir con `json.Marshal` sin `SetEscapeHTML(false)` rompe la legibilidad esperada; omitir la comprobación de colisión sirve la grabación equivocada (SC-008, FR-039, FR-040) |
| 13 | **Reproducción** | `TestReplaySirveLasGrabaciones`, `TestReplayEsDeterminista`, `TestReplayPeticionSinGrabacion`, `TestReplayGrabacionDeOtraPeticion`, `TestReplayGarantiasVigentes`, `TestReplaySinRobots`, `TestReplayRedirecciones`, `TestReplayEstadoDeErrorGrabado`, `TestReplayRechazaOpciones`; los diez fixtures a mano de `internal/httpx/testdata/reproduccion/prueba/` (§«Fixtures») | `test` | Sí | Si `robots` estuviera en la cadena, el `robots.txt` grabado —que deniega todo— haría fallar con clase 5 cada petición de `TestReplaySinRobots`; `-count=2 -shuffle=on` da el mismo resultado (SC-007, FR-049) |
| 14 | **Ensayo** | `TestEnsayoNoAbreConexion`, `TestEnsayoDevuelveRespuestaDeclarada`, `TestEnsayoCompruebaArgumentos`; `TestAdaptadorDePruebaConElKernel/ensayo` (kernel en proceso: cero accesos, línea en stderr, stdout vacío, código 0); `TestDryRunPresentaElEnsayo` en `internal/app` | `test` | Sí | Sin la salida temprana de `Pedir`, el servidor registra un acceso a `robots.txt`; sin la presentación de `Ensayo` en el kernel, la línea no aparece (SC-013) |
| 15 | **Concurrencia** | `TestPedirDesdeVariasGoroutines` bajo `-race` | `test` | Sí | Sin el mutex por sitio, el detector de carreras o el contador de `robots.txt` > 1 lo delatan (SC-011) |
| 16 | **Plazo y redirecciones** | `TestPedirRespetaElPlazo`, `TestPedirSigueRedirecciones`, `TestPedirCortaCadenasDeRedirecciones`, `TestPedirRechazaMetodo`, `TestPedirRechazaDireccion`, `TestOpcionesInvalidas`, `TestTransporteConfiguracion` | `test` | Sí | `Client.Timeout ≠ 0` o `CheckRedirect` nulo hacen fallar `TestTransporteConfiguracion` (FR-004, FR-011) |
| 17 | **Binario sin `internal/httpx`** | `TestElBinarioNoEnlazaHTTPX` + `TestDependenciasDelBinario` (lista sin cambios) | `test` | Sí | Importar `internal/httpx` desde `internal/app` (no test) hace fallar los dos (SC-014); se retira en H4 |
| 18 | **Direcciones solo locales en los tests y fixtures** | `grep` del quickstart (prerrequisitos): extrae toda dirección `http(s)://` de `internal/httpx/*_test.go` y de `internal/httpx/testdata`, descarta los cuatro hosts admitidos (`127.0.0.1`, `localhost`, `fuente.prueba`, `otra.prueba`, sin distinguir mayúsculas) y la dirección de la identificación, y exige que no quede nada. **Todas las tablas del hito se escriben contra esos hosts**: la de `TestClaveDeSitio` (fila 9, inventario) y la del contrato de grabación §2 que copia `TestNombreDeGrabacion` (ningún `x`, `example.com` ni host real); el sufijo de la identificación en `TestAgenteDeUsuario` va por `regexp.QuoteMeta` para que el token extraído sea exactamente `https://ventanillalegal.es/bot` (fila 7). Se ejecuta la orden **entera** contra el árbol una vez escritos los tests (obligación 11) | manual | No | Cualquier `https://www.boe.es` —o un `http://x` de una tabla— en un test o fixture aparece y la confirmación no se imprime |
| 19 | **Sin grabaciones en el árbol** | medida diferencial de `git status --porcelain` sobre `testdata/`, `internal/httpx/testdata/`, `internal/source/` (quickstart 1 y 7); guardián de diff del workflow con `[datos]` | manual + workflow | No (el guardián sí, por tarea) | Un test que grabara en `testdata/` de raíz cambia la medida (SC-008) |
| 20 | Formato, lint general, `-race`, `govulncheck`, `gosec`, CodeQL, `gitleaks`, `go mod verify`, `go mod tidy -diff`, pre-commit, cobertura | los de H0/H1, **sin ninguna exclusión nueva** (SC-010) | varias | Sí (salvo CodeQL) | Los mismos de H1; `go mod tidy -diff` cubre las dos dependencias nuevas |

### Fixtures

H2 crea `internal/httpx/testdata/reproduccion/prueba/` con grabaciones **escritas a mano** contra
`http://fuente.prueba` (D17): son material de test —el contrato del formato en forma legible y la fuente
de los tests de reproducción—, no grabaciones de una fuente real. Las tareas que lo toquen llevan
`[datos]`; el guardián del workflow exige la etiqueta y no pausa (material nuevo bajo
`internal/<pkg>/testdata/`). **No se crea `testdata/<fuente>/` en la raíz** ni se graba nada real
(FR-044).

**Lista cerrada** de los ficheros del directorio: estos **diez** y ninguno más. Es la lista de rutas que
la tarea `[datos]` declara y la que el guardián de diff comprueba. Cada nombre es el que produce la
regla del [contrato de grabación §2](./contracts/formato-de-grabacion.md) para su petición (host
`fuente.prueba` sin puerto explícito → sin `-<puerto>`; `/` → `_`; `?id=…` → `_q_id_…`; `,` → `_`), y
cada par petición → nombre es además una fila de `TestNombreDeGrabacion`, de modo que un fichero mal
nombrado falla ahí antes de que ningún test de reproducción lo busque. Todos llevan `"formato": 1`,
`grabado_en` fijo (`2026-09-12T00:00:00Z`), `peticion.cabeceras` con solo `User-Agent:
kitlegal/dev (+https://ventanillalegal.es/bot)` (las cabeceras no participan en el emparejamiento,
FR-039) y una cabecera `Date` fija en la respuesta.

| # | Fichero | Petición grabada | Respuesta grabada | Lo usa |
|---|---|---|---|---|
| 1 | `GET_http_fuente.prueba_norma_q_id_BOE-A-2015-10565.json` | `GET http://fuente.prueba/norma?id=BOE-A-2015-10565` | `200`; `Content-Type: application/xml; charset=utf-8`; `cuerpo` con un XML de una línea que contiene `<`, `>` y `&` (sin escape HTML, contrato §1) | `TestReplaySirveLasGrabaciones` (texto: estado, cabeceras y cuerpo), `TestReplayEsDeterminista`, `TestReplayGarantiasVigentes`, `TestReplaySinRobots`; destino del fichero 4 |
| 2 | `HEAD_http_fuente.prueba_norma_q_id_BOE-A-2015-10565.json` | `HEAD http://fuente.prueba/norma?id=BOE-A-2015-10565` | `200`; las mismas cabeceras que el 1; `"cuerpo": ""` | `TestReplaySirveLasGrabaciones` (HEAD: el método forma parte de la clave, cuerpo vacío grabado como `""`) |
| 3 | `GET_http_fuente.prueba_documento.pdf.json` | `GET http://fuente.prueba/documento.pdf` | `200`; `Content-Type: application/pdf`; `cuerpo_base64` de unos bytes que no son UTF-8 válido (`%PDF-1.4` y bytes altos) | `TestReplaySirveLasGrabaciones` (binario: `Cuerpo` es el resultado de decodificar `cuerpo_base64`) |
| 4 | `GET_http_fuente.prueba_antigua.json` | `GET http://fuente.prueba/antigua` | `302`; `Location: /norma?id=BOE-A-2015-10565` (relativa) | `TestReplayRedirecciones` (seguida: la respuesta es la del fichero 1 y `Respuesta.URL` es la dirección final) |
| 5 | `GET_http_fuente.prueba_movida.json` | `GET http://fuente.prueba/movida` | `302`; `Location: http://fuente.prueba/desaparecida` (absoluta; sin grabación) | `TestReplayRedirecciones` (destino ausente: clase 1 nombrando `GET http://fuente.prueba/desaparecida`, FR-047) |
| 6 | `GET_http_fuente.prueba_bucle.json` | `GET http://fuente.prueba/bucle` | `302`; `Location: /bucle` | `TestReplayRedirecciones` (bucle: clase 4 dentro del directorio, FR-049 d) |
| 7 | `GET_http_fuente.prueba_caida.json` | `GET http://fuente.prueba/caida` | `503`; `"cuerpo": ""` | `TestReplayEstadoDeErrorGrabado` (clase 4, una sola búsqueda, sin espera) |
| 8 | `GET_http_fuente.prueba_limitada.json` | `GET http://fuente.prueba/limitada` | `429`; `Retry-After: 120`; `"cuerpo": ""` | `TestReplayEstadoDeErrorGrabado` (clase 5, `Espera` = 120 s, `EsperaConocida`) |
| 9 | `GET_http_fuente.prueba_a_b.json` | `GET http://fuente.prueba/a,b` | `200`; `Content-Type: text/plain; charset=utf-8`; `cuerpo` `a,b` | `TestReplayGrabacionDeOtraPeticion`: el test pide `GET http://fuente.prueba/a_b`, cuyo nombre derivado es el mismo (`,` → `_`), y recibe clase 1 nombrando las dos peticiones y el fichero (FR-047) |
| 10 | `GET_http_fuente.prueba_robots.txt.json` | `GET http://fuente.prueba/robots.txt` | `200`; `Content-Type: text/plain`; `cuerpo` `User-agent: *\nDisallow: /` | `TestReplaySinRobots`, subprueba `con-grabacion`: queda **sin usar** (FR-049 i); si la reproducción lo consultara, todas las demás peticiones fallarían con clase 5. La subprueba `sin-grabacion` copia el directorio a `t.TempDir()` sin este fichero (US3 escenario 7, SC-007) |

`TestReplayPeticionSinGrabacion` pide `GET http://fuente.prueba/inexistente` y no necesita fichero;
`TestReplayGarantiasVigentes` (`POST` → 2) y `TestReplayRechazaOpciones` tampoco.

**Tests de contrato.** Los cuatro contratos de [`contracts/`](./contracts/) tienen comprobación mecánica:
el cliente (filas 4-10, 15, 16), el formato de grabación (12, 13), errores y ensayo (11, 14) y las reglas
de arquitectura (1, 2, 3, 17).

**Objetivos del `Makefile` que H2 toca**, y ninguno más: la variable `LDFLAGS` (D5). Ningún objetivo
cambia; `test` ya cubre `./...` con `-race`.

## Inventario de tests

Nombres fijados aquí para que `tasks.md` y `quickstart.md` los usen tal cual. Sin guiones bajos
(convención del proyecto); un fichero de test por fichero de código y en el mismo orden que el árbol de
«Project Structure» (skill `golang-testing`), con una sola excepción declarada: `doc.go` no tiene test
porque solo contiene el comentario de paquete y ninguna declaración; `t.Parallel()` en todos salvo los
que usan `t.Setenv` (grabación), que `paralleltest` reconoce.

| Fichero | Tests | Cubre |
|---|---|---|
| `cliente_test.go` | `TestNewSinOpciones`, `TestOpcionesInvalidas`, `TestPedirRechazaMetodo`, `TestPedirRechazaDireccion`, `TestPedirRespetaElPlazo`, `TestPedirSigueRedirecciones`, `TestPedirCortaCadenasDeRedirecciones`, `TestPedirDesdeVariasGoroutines` | US1-1, US1-3, US1-7, US5-5, FR-011, FR-020, FR-024, FR-057, D19 |
| `peticion_test.go` | `TestCabecerasGet`, `TestRespuestaDescripcion` | D2 |
| `errores_test.go` | `TestClasesDeError` (16 filas), `TestErrorEnvueltoConservaLaClase`, `TestErrorRetryAfter`, `TestErrorMensajesEnEspanol` | SC-006, SC-016, FR-031, FR-034, D20 |
| `agente_test.go` | `TestAgenteDeUsuario`: expresión regular construida sobre la variable `version` y con el sufijo escapado en ejecución (`"^kitlegal/" + regexp.QuoteMeta(version) + regexp.QuoteMeta(" (+https://ventanillalegal.es/bot)") + "$"`), `version` no vacía, y `t.Log` de la identificación para que `-ldflags` sea observable (D5) | FR-006, FR-007 |
| `sitio_test.go` | `TestClaveDeSitio` (tabla, solo hosts ficticios admitidos por el control 18): `http://fuente.prueba` → `http://fuente.prueba:80`; `https://fuente.prueba` → `https://fuente.prueba:443`; `http://fuente.prueba:8080` conserva el puerto; `http://Fuente.Prueba/ruta?q=1` → `http://fuente.prueba:80` (host en minúsculas, sin ruta ni consulta); `http://fuente.prueba` ≠ `https://fuente.prueba`; `http://fuente.prueba` ≠ `http://otra.prueba`; `http://127.0.0.1:53211` ≠ `http://127.0.0.1:53212`; `http://fuente.prueba` = `http://fuente.prueba:80` | FR-013, FR-015, FR-019, SC-004, D7 |
| `identificar_test.go` | `TestIdentificacionEnTodaPeticion` | SC-001, FR-009 |
| `robots_test.go` | `TestRobotsDeniegaLaRuta`, `TestRobotsSePideUnaVezPorSitio`, `TestRobotsCasosDeObtencion`, `TestRobotsRedirigido`, `TestRobotsNoSeEvaluaASiMismo`, `TestRobotsPorSitio` | SC-003, US1-4, US1-8, US2-4, FR-013 a FR-018 |
| `reintentos_test.go` | `TestReintentosDosErroresYUnAcierto`, `TestReintentosNoRepite4xx`, `TestReintentosCancelacionGana`, `TestReintentosCierraCuerposDescartados`, `TestReintentosAgotados` | SC-005, FR-023 a FR-029 |
| `ritmo_test.go` | `TestRitmoSeparaPeticionesDelMismoSitio`, `TestRitmoNoRetrasaOtroSitio`, `TestRitmoRespetaElContexto` | SC-004, FR-019 a FR-022 |
| `grabar_test.go` | `TestGrabarEscribeElFichero`, `TestGrabarNoAlteraLaRespuesta`, `TestGrabarFormatoEstable`, `TestGrabarSinVariableNoEscribe`, `TestGrabarConfiguracionIncompleta`, `TestGrabarRaizInvalida`, `TestGrabarColision`, `TestGrabarValorDeVariableInvalido`, `TestGrabarCuerpoBinario` | SC-008, US4, FR-036 a FR-043, FR-064 |
| `nombre_test.go` | `TestNombreDeGrabacion` (la tabla del contrato §2 más los diez pares petición → fichero de §«Fixtures», incluida la colisión `/a,b` y `/a_b`) | contrato de grabación §2 |
| `reproducir_test.go` | `TestReplaySirveLasGrabaciones` (texto, HEAD, binario), `TestReplayEsDeterminista`, `TestReplayPeticionSinGrabacion`, `TestReplayGrabacionDeOtraPeticion`, `TestReplayGarantiasVigentes`, `TestReplaySinRobots` (`con-grabacion`, `sin-grabacion`), `TestReplayRedirecciones` (seguida, destino ausente, bucle), `TestReplayEstadoDeErrorGrabado` (503, 429), `TestReplayRechazaOpciones`; sobre los diez fixtures de §«Fixtures» | SC-007, US3, FR-045 a FR-049, FR-043 |
| `transporte_test.go` | `TestTransporteConfiguracion` | FR-004, FR-012, D11 |
| `ensayo_test.go` | `TestEnsayoNoAbreConexion`, `TestEnsayoDevuelveRespuestaDeclarada`, `TestEnsayoCompruebaArgumentos` | SC-013, FR-050, FR-052, FR-065 |
| `superficie_test.go` | `TestSuperficieExportada` | FR-002 |
| `adaptador_test.go` (`httpx_test`) | tipos `adaptadorDePrueba` (applet `prueba`) y `argumentosConsultar` (verbo `consultar`); `TestAdaptadorDePruebaConElKernel` con subpruebas `ensayo`, `consulta`, `plazo` | US5-1, US6, FR-059 a FR-062 |
| `internal/arch_test.go` | R2 endurecida; `TestElBinarioNoEnlazaHTTPX` | FR-053, SC-014 |
| `internal/cli/errors_test.go` | `TestClasificar` + casos `ConClase` | D4 |
| `internal/app/main_test.go` | `TestDryRunPresentaElEnsayo` | FR-051, D6 |
| `cmd/kitlegal/main_test.go` | `TestVersionDeLaIdentificacion` | FR-007 |

## Orden de implementación

La constitución fija `core → adaptador → applet → skill`. En H2 no hay applet ni skill: el orden es
`dominio → cliente (de dentro afuera por la cadena) → kernel → adaptador de prueba → controles`, y en cada
paso **el test de tabla se escribe antes** que el código que lo hace pasar, con `make ci` en verde al
cerrar cada tarea (regla del modo desatendido).

1. **Dependencias**: `go get golang.org/x/time@v0.16.0 github.com/temoto/robotstxt@v1.1.2`, `go mod
   tidy`, comprobar S1 (`go mod tidy -diff` limpio, `TestDependenciasDelBinario` en verde).
2. **Dominio** (`internal/core/schema`): `ConClase` y `Resultado.Ensayo`, con sus tests; cobertura
   ≥ 85 % mantenida.
3. **Tipos del cliente** (`peticion.go`, `errores.go`, `agente.go`, `sitio.go`, `nombre.go`): sin red;
   sus tests son puros. `TestClasesDeError` se escribe ya con las 16 filas y falla hasta que existan las
   rutas; `TestClaveDeSitio` con la tabla de la clave de sitio; `TestNombreDeGrabacion` con la tabla del
   contrato y los diez pares de §«Fixtures».
4. **Transporte y `Pedir` mínimo** (`transporte.go`, `cliente.go`): validación, ensayo, bucle de
   redirecciones y clasificación con una cadena de un solo escalón (`identificar → transporte`).
   Tests: `cliente_test.go`, `ensayo_test.go`, `identificar_test.go`, `transporte_test.go`.
5. **Ritmo** (`ritmo.go`) y **reintentos** (`reintentos.go`, jitter con `crypto/rand`, D9) con el reloj
   inyectable; sus tests.
6. **`robots.txt`** (`robots.go`): última en la cadena porque usa reintentos y ritmo por debajo; sus
   tests, incluida la tabla de FR-015.
7. **Grabación** (`grabar.go`) y **reproducción** (`reproducir.go`) con `Replay`; los diez fixtures a
   mano de §«Fixtures» (`[datos]`, rutas declaradas una a una); sus tests. `TestClasesDeError` pasa
   entera aquí.
8. **Kernel**: `cli.Clasificar` (+ casos en `TestClasificar`), `internal/app` (presentación de `Ensayo`,
   `TestDryRunPresentaElEnsayo`), `cmd/kitlegal/main_test.go` (`TestVersionDeLaIdentificacion`),
   `Makefile` (`LDFLAGS`).
9. **Adaptador de prueba** (`adaptador_test.go`) y `TestAdaptadorDePruebaConElKernel`: el criterio de
   aceptación del hito, ejecutado con el kernel en proceso.
10. **Controles y documentación**: `internal/arch_test.go` (R2 con dueño, `TestElBinarioNoEnlazaHTTPX`),
    `TestSuperficieExportada`, comentarios de `.golangci.yml`, `docs/ADR/0011-ensayo-por-capas.md`,
    `docs/PENDIENTES.md`. Validación de SC-012 **en copia desechable** (quickstart 10). `make ci` en verde
    de principio a fin.

Sin `CHANGELOG.md`: H2 no cambia ningún comportamiento visible (FR-062, SC-014; spec, *Fuera de alcance*).

## Complexity Tracking

> Divergencias conscientes respecto de la lectura literal del spec o del roadmap. Ninguna afecta a
> alcance, frontera humana, privacidad, términos de uso ni a una decisión cerrada. **Ninguna dependencia
> está fuera de la lista de la constitución §V**: las dos nuevas son las que el hito nombra.

| Divergencia | Por qué es necesaria | Alternativa más simple, y por qué se rechaza |
|---|---|---|
| **`Pedir` recibe `schema.Contexto` además de `context.Context`** (FR-002 habla de «una operación que exige `context.Context` como primer parámetro»; el primero sigue siéndolo) | Es lo que hace que `--dry-run` llegue al cliente sin poder olvidarse (FR-050) y lo que H3 necesitará para `--offline`. El adaptador ya lo tiene en la mano ([D1](./research.md#d1--superficie-pública-un-tipo-dos-constructores-una-operación-cinco-opciones)). | *Opción `ConEnsayo`*: olvidable, y un olvido emite una petición real bajo `--dry-run`. *`New(ejecucion, opts...)`*: rompe la forma literal `httpx.New(opts...)` y obliga a `Replay` a lo mismo. |
| **`schema.Resultado` gana `Ensayo []string`** y el kernel lo presenta; ADR 0011 | FR-051 exige que método y dirección lleguen a la salida de error sin depender del nivel de registro; el único canal es el presentador, y el único camino hasta él es el `Resultado` ([D6](./research.md#d6----dry-run-pedir-devuelve-una-respuesta-de-ensayo-la-descripción-viaja-en-schemaresultadoensayo-y-la-presenta-el-kernel)). | *`slog`*: filtrable por nivel (H1 D10 lo rechazó). *Un `io.Writer` en el cliente*: sin valor por omisión posible (R5) y olvidable. *Cambiar la firma de `Ejecutar`*: toca todos los applets para una línea. |
| **`schema.ConClase` y una comprobación más en `cli.Clasificar`** | Un adaptador no debe importar `internal/cli` (spec, *Assumptions*), y `Retry-After` tiene que viajar con el error (FR-030), lo que exige un tipo y `errors.As` ([D4](./research.md#d4--errores-tipados-httpxerror-declara-su-clase-el-kernel-la-reconoce-con-errorsas-sin-que-el-adaptador-importe-internalcli)). | *Envolver los sentinelas de `cli`*: `httpx → cli`, un adaptador dependiendo del kernel, y sin datos en el error. *Mover los sentinelas a `schema`*: reabre H1 D8 y no transporta `Retry-After`. |
| **`Replay(dir, opciones...)`** en lugar del literal `Replay(dir)` | Un cliente de reproducción también registra eventos (FR-035) y FR-043 exige rechazar la combinación raíz de grabación + directorio de reproducción, que solo existe si `Replay` acepta opciones. Las que no tienen sentido en reproducción se rechazan con clase 2. | *Sin opciones*: FR-043 no tendría nada que rechazar y el cliente de reproducción no podría registrar. |
| **`cmd/kitlegal/main_test.go` importa `internal/httpx`** (solo en test) | Es el único sitio que ve `main.version` y puede comprobar que coincide con el valor por omisión de la identificación (FR-007, [D5](./research.md#d5--la-versión-de-la-identificación-una-variable-de-paquete-inyectada-por--ldflags-con-la-misma-version-del-makefile-que-alimenta-al-verbo-version)). `go list -deps` no sigue importaciones de test, así que el binario no cambia. | *Comprobarlo con el binario construido*: el binario no expone la identificación (ningún applet la usa). *No comprobarlo*: dos «dev» que podrían divergir sin que nada lo diga. |
| **`TestElBinarioNoEnlazaHTTPX`, un test temporal** que H4 retirará | SC-014 pide que la superficie del binario no cambie; que no enlace un paquete con dos módulos nuevos es la forma mecánica de verlo. Cuando H4 lo enlace, el test desaparece junto con la actualización justificada de `modulosDelBinario`. | *Solo `TestDependenciasDelBinario`*: cubre los módulos, no el paquete; un `internal/httpx` enlazado sin dependencias nuevas pasaría inadvertido. |
| **Dirección inválida → clase «argumentos»**, situación no enumerada en FR-063 | Ninguna ruta de fallo puede quedar sin clase (FR-063), y es el mismo caso que el método: mal formado y corregible por quien llama ([D19](./research.md#d19--clase-argumentos-para-una-dirección-que-no-es-httphttps-absoluta)). | *Dejar que falle el transporte*: se reintentaría dos veces y saldría con clase 4, que significa otra cosa. |

## Obligaciones que este plan traslada a `tasks.md`

1. **Comprobar S1** al añadir las dependencias: `go mod tidy -diff` limpio, `go mod graph` sin módulos
   nuevos más allá de los dos, y `TestDependenciasDelBinario` en verde sin tocar su lista.
2. **El jitter usa `crypto/rand.Read` sobre ocho bytes con `encoding/binary`** (D9, forma exacta allí):
   ningún fichero de `internal/httpx` importa `math/rand`, `math/rand/v2` ni `math/big`; G404 marca los
   dos primeros (verificado) y no hay `//nolint`. El resultado de `rand.Read` se ignora sin `_ =` (no
   devuelve nunca error; `errcheck` y G104 lo excluyen por omisión, verificado) y la conversión a `int64`
   lleva el `>> 1` que G115 reconoce. No se escribe ninguna rama para un error que el lector no puede
   devolver. La tarea de `reintentos.go` lo declara y `make lint` lo confirma.
3. **Comprobar S2** (`misspell`) al primer `make lint`; un falso positivo en español va a `ignore-rules`
   como en H1, no a una supresión.
4. **Etiquetar `[datos]`** toda tarea que cree o toque `internal/httpx/testdata/`, dejando claro en la
   tarea que es material de test escrito a mano y no una grabación real.
5. **Declarar en cada tarea sus rutas** (guardián de diff): las de H1 que se tocan son exactamente
   `internal/core/schema/error.go`, `internal/core/schema/sobre.go` (y sus tests),
   `internal/cli/errors.go` (y su test), `internal/app/main.go` (y su test), `cmd/kitlegal/main_test.go`,
   `internal/arch_test.go`, `Makefile`, `.golangci.yml`, `docs/PENDIENTES.md`, más `go.mod`, `go.sum` y
   el ADR nuevo.
6. **Validar SC-012 sobre una copia desechable** fuera del repositorio (quickstart 10), nunca sobre el
   árbol de trabajo, y dejar la salida como evidencia en la PR.
7. **Medir SC-008 de forma diferencial** (quickstart 1 y 7) tras `make ci`, y anotar el resultado.
8. **Anotar para H4**: retirar `TestElBinarioNoEnlazaHTTPX` y ampliar `modulosDelBinario` con
   justificación cuando el primer adaptador enlace `internal/httpx`; decidir la ubicación de los fixtures
   reales antes de la primera tarea `[datos]` (D17, `docs/PENDIENTES.md`); el ritmo por fuente
   (`ConIntervalo`) sale de `docs/SOURCES.md`.
9. **Escribir `docs/ADR/0011-ensayo-por-capas.md`** (MADR corto) en la misma tarea que `Resultado.Ensayo`.
10. **Mantener los umbrales**: `internal/cli` ≥ 90 % con los casos nuevos de `TestClasificar`;
    `internal/core` ≥ 85 %; global ≥ 70 %. Nunca se rebaja un umbral para que pase un estado.
11. **Solo hosts admitidos en tests y fixtures** (control 18): toda dirección que aparezca en
    `internal/httpx/*_test.go` o en `internal/httpx/testdata/` usa `127.0.0.1`, `localhost`,
    `fuente.prueba` u `otra.prueba` (en cualquier combinación de mayúsculas); la única excepción es la
    dirección de la identificación, `https://ventanillalegal.es/bot`, que se envía y no se consulta. Las
    tareas de `sitio_test.go`, `nombre_test.go`, `agente_test.go` y de los fixtures lo declaran, y la
    última tarea de tests ejecuta la orden **entera** de los prerrequisitos del quickstart contra el
    árbol y deja la confirmación «solo direcciones locales» como evidencia en la PR.

## Comprobación contra la rúbrica del juez (`juez_plan`, criterios a-j)

| Criterio | Dónde se cumple |
|---|---|
| a. constitution_check | Un ítem por principio (I-IX) y por regla de dependencia (5), con veredicto; gates mecánicos y re-evaluación tras el diseño |
| b. dependencias | Dos, ambas de la lista §V; ninguna transitiva nueva (S1); *Complexity Tracking* lo declara |
| c. reglas_dependencia | `httpx → schema` únicamente; `net/http` solo en `httpx`; nada de SQLite; sin `os.Exit`; sin stdout; contrato de arquitectura |
| d. errores_exit_codes | `httpx.Error` + `schema.ConClase` + `cli.Clasificar`; tabla cerrada de 16 situaciones → códigos; `TestClasesDeError` |
| e. tests_primero | Inventario de tests, fixtures (`internal/httpx/testdata/`), controles mecánicos (tabla de 20 filas); e2e: ninguno, con la razón del spec |
| f. alcance | Todo lo de *Fuera de alcance* del spec queda fuera; lo único no enumerado (dirección inválida) está declarado y justificado |
| g. sin_atajos | Ningún `nolint`, test saltado, TODO, error silenciado ni rama inalcanzable; el jitter usa `crypto/rand.Read` porque G404 marca `math/rand/v2` y porque `Read` no devuelve nunca error (D9); el único supuesto con contingencia (S2, `misspell`) se resuelve sin supresión |
| h. mejor_alternativa | Cada decisión de research.md lleva la alternativa rechazada y el motivo verdadero; ninguna elige una opción que dependa de una supresión o de una contingencia, y la de `rand.Int` frente a `rand.Read` se decide por lo que hace el código del toolchain, no por lo que sugiere `go doc` |
| i. afirmaciones_verificadas | Tabla de verificación en research.md con la ruta de cada comprobación (G404: `rules/rand.go`; `crypto/rand`: código del toolchain en `crypto/rand`, `crypto/internal/rand` y `crypto/internal/sysrand`; `errcheck`, G104 y G115: su código); lo no verificado en D22 como supuesto |
| j. quickstart_ejecutable | 12 escenarios con órdenes, rutas y nombres de test reales; salidas de H1 copiadas de la ejecución real; violaciones solo en copia desechable; medida diferencial de `git status` con `test "$ANTES" = "$DESPUES"`; el filtro del escenario 9 busca identificadores que no existen en H1 (`adaptadorDePrueba`, `"consultar"`) fuera de `_test.go`; la comprobación «solo direcciones locales» admite exactamente los hosts con los que el plan escribe todas sus tablas (`fuente.prueba`, `otra.prueba`, `127.0.0.1`, `localhost`) y se ejecuta entera contra el árbol al cerrar los tests (obligación 11) |
