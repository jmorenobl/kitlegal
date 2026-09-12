# Research: H2 · `internal/httpx`: cliente HTTP responsable + grabación de fixtures

**Modo**: desatendido. Cada decisión de abajo se tomó con el «Criterio de decisión autónoma» de
`.specify/memory/constitution.md` (siempre la mejor solución; lo no especificado no se implementa; elegir
con criterio y dejar rastro) y registra la alternativa rechazada y por qué. Ninguna afecta a alcance,
frontera humana, privacidad, términos de uso, reglas de anomalías ni a una decisión cerrada de
`CLAUDE.md`, de modo que ninguna dispara el criterio 4 (escalar).

**Verificación.** Toda afirmación sobre el comportamiento de una herramienta o dependencia externa se ha
comprobado en local contra su código o su documentación, y cada decisión cita dónde. Las dos dependencias
nuevas del hito no estaban en la caché de módulos y se descargaron con `go mod download` sin tocar
`go.mod` ni `go.sum`; los módulos de las herramientas de control ya estaban (los trae
`tools/golangci-lint/go.mod`). Lo único que no se ha podido verificar en local está en la sección
[D22 · Supuestos no verificados](#d22--supuestos-no-verificados), y se declara como supuesto, no como
hecho. Rutas de verificación usadas en este documento:

| Qué | Dónde se comprobó |
|---|---|
| Toolchain de Go | `go version` → `go1.27.1 darwin/arm64`; `go.mod` declara `go 1.27.0` + `toolchain go1.27.1` |
| Biblioteca estándar | `go doc <símbolo>` sobre ese toolchain y su fuente en `$(go env GOROOT)/src` (`/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.1.darwin-arm64/src`) |
| `golang.org/x/time/rate` | `go doc -C /Users/jorge/go/pkg/mod/golang.org/x/time@v0.16.0 ./rate …` (descargado: v0.16.0, `go.mod` declara `go 1.26.0`, sin dependencias) |
| `github.com/temoto/robotstxt` | `/Users/jorge/go/pkg/mod/github.com/temoto/robotstxt@v1.1.2/robotstxt.go` (descargado: v1.1.2; su `go.mod` solo requiere `stretchr/testify v1.3.0`, que la selección mínima de versiones resuelve con el v1.12.1 ya presente) |
| `noctx` | `/Users/jorge/go/pkg/mod/github.com/sonatard/noctx@v0.5.1/noctx.go` (tabla `ngFuncMessages`) |
| `bodyclose` | `/Users/jorge/go/pkg/mod/github.com/timakin/bodyclose@v0.0.0-20260129054331-73d1f95b84b4/README.md` |
| `paralleltest` | `/Users/jorge/go/pkg/mod/github.com/kunwardeep/paralleltest@v1.0.15/pkg/paralleltest/paralleltest.go` |
| `gosec` | `/Users/jorge/go/pkg/mod/github.com/securego/gosec/v2@v2.28.0/rules/tls.go` (G402) y `rules/rand.go` líneas 31-36 (G404: «Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand)»; la lista de `math/rand/v2` incluye `New`, `Float64`, `Int`, `IntN`, `Int64N`, `N`, `Uint64N`…) |
| `crypto/rand` | Código del toolchain, no solo `go doc`: `crypto/rand/rand.go` 44-64 (`Read` «never returns an error, and always fills b entirely»; ante un fallo del lector llama a `fatal(…)` y el proceso termina); `crypto/internal/rand/rand.go` 22-31 (el `Read` de `rand.Reader` llama a `drbg.Read(b)` y devuelve siempre `len(b), nil`: **nunca** devuelve error); `crypto/internal/sysrand/rand.go` 31-37 y 50-53 (`sysrand.Read` «crashes the program irrecoverably if an error is encountered»; «documented to never return an error on all but legacy Linux systems») y 56 («the urandom fallback is only used on Linux kernels before 3.17 and on AIX»); `crypto/rand/util.go` 71-74 (`Int` hace `panic` si `max <= 0`); `go doc crypto/rand` (la API completa en go1.27.1: `Reader`, `Int`, `Prime`, `Read`, `Text`). Conclusión para D9: con el lector por omisión, `rand.Int` y `rand.Read` se comportan igual ante un fallo de entropía —el proceso termina— y ninguna de las dos ofrece una ruta de error manejable |
| `errcheck` v1.20.0 (el de `golangci-lint` v2.13.2, `tools/golangci-lint/go.mod` 115) | `/Users/jorge/go/pkg/mod/github.com/kisielk/errcheck@v1.20.0/errcheck/excludes.go` 21: `crypto/rand.Read` está en `DefaultExcludedSymbols` (con enlace a go.dev/issue/66821); `/Users/jorge/go/pkg/mod/github.com/golangci/golangci-lint/v2@v2.13.2/pkg/golinters/errcheck/errcheck.go` 108-109: esa lista se añade salvo `disable-default-exclusions`, que `.golangci.yml` no activa. Un `rand.Read(b[:])` sin comprobar el resultado pasa `errcheck` sin exclusión nueva (D9) |
| `gosec` G104 (errores sin manejar) | `rules/errors.go` 88: la lista blanca incluye `rand.Read` (el nombre de paquete se compara con el identificador de importación, `helpers.go` 213); 53: la rama sobre asignaciones a `_` solo actúa en modo `audit`, que no está activado. Un `rand.Read(b[:])` como sentencia pasa G104 (D9) |
| `gosec` G115 (desbordamiento en conversiones) | `analyzers/range_analyzer.go` 790-805: para `x >> k` con `k` constante el analizador fija el máximo en `Max(tipo de x) >> k` aunque no conozca el rango de `x`; `analyzers/conversion_overflow.go` 230-233: una conversión de sin signo a con signo es segura si ese máximo cabe en el destino. `int64(binary.BigEndian.Uint64(b[:]) >> 1)` tiene máximo `MaxInt64` y pasa G115 (D9). Sin ese desplazamiento, `int64(uint64)` sí se marcaría |
| `golangci-lint`, `depguard`, `forbidigo` | ya verificados en H1 (`specs/002-h1-kernel-cli-multicall/research.md` D18); H2 no cambia su configuración salvo comentarios |
| Kernel de H1 | `internal/cli/errors.go`, `internal/cli/log.go`, `internal/cli/presentador.go`, `internal/app/main.go`, `internal/app/applet.go`, `internal/core/schema/{contexto,error,sobre}.go`, `internal/arch_test.go`, `.golangci.yml`, `Makefile` |

---

## D1 · Superficie pública: un tipo, dos constructores, una operación, cinco opciones

**Decisión.** `internal/httpx` exporta exactamente esto (el detalle está en
[`contracts/cliente-httpx.md`](./contracts/cliente-httpx.md)):

```go
type Cliente struct{ /* sin campos exportados */ }

func New(opciones ...Opcion) (*Cliente, error)                 // FR-001: cliente contra la red
func Replay(dir string, opciones ...Opcion) (*Cliente, error)  // FR-045: cliente de reproducción

func (c *Cliente) Pedir(ctx context.Context, ejecucion schema.Contexto, p Peticion) (Respuesta, error)

type Opcion func(*configuracion) error
func ConFuente(nombre string) Opcion            // FR-039: <fuente>
func ConRaizDeGrabacion(dir string) Opcion      // FR-064: <raíz>
func ConIntervalo(d time.Duration) Opcion       // FR-020: ritmo, uno por cliente
func ConIntentos(n int) Opcion                  // FR-024: intentos acotados
func ConRegistrador(l *slog.Logger) Opcion      // FR-035: registro de eventos

type Peticion struct{ Metodo, URL string }
type Respuesta struct{ Peticion Peticion; URL string; Estado int; Cabeceras Cabeceras; Cuerpo []byte; Ensayo bool }
type Cabeceras map[string][]string
type Error struct{ /* clase, petición, estado, Retry-After, causa */ }
func AgenteDeUsuario() string
const VariableGrabacion = "KITLEGAL_RECORD"
```

Nada más: ni `*http.Client`, ni `http.RoundTripper`, ni `*http.Request`, ni `*http.Response` aparecen en
ninguna firma exportada (FR-002). La cadena de decoradores es privada (D3).

**`Pedir` recibe el contexto de ejecución (`schema.Contexto`) además del `context.Context`.** Es lo que
hace que `--dry-run` (FR-050) —y en H3, `--offline`— llegue al cliente **sin que quien llama pueda
olvidarlo**: el adaptador ya recibe `ejecucion` en `Ejecutar(ctx, ejecucion, registrador)` (H1,
`internal/app/applet.go`) y solo tiene que pasarlo. Un cliente vive una invocación y el contexto de
ejecución también, así que podría ir en el constructor; va en la operación para que `New(opts...)` y
`Replay(dir)` conserven la forma literal del enunciado del hito y para que un cliente de reproducción
honre `--dry-run` por el mismo camino que uno real (FR-050 no distingue).

**`Peticion` es un `struct` y no dos cadenas.** Hoy solo lleva método y dirección (FR-037 graba «método,
dirección completa, cabeceras de la petición», y en H2 la única cabecera es la identificación). Un
`struct` permite que H4 añada lo que necesite —un `Accept`, por ejemplo— sin cambiar la firma de `Pedir`
ni el formato de grabación, que ya lo serializa como objeto. No se añade ningún campo más ahora: sería
anticipar (constitución, criterio 2).

**Alternativas consideradas.**

- *`New(ejecucion schema.Contexto, opts...)`*: igual de imposible de olvidar, pero cambia la forma literal
  `httpx.New(opts...)` del roadmap y obliga a `Replay` a recibirlo también; rechazada por eso.
- *Una opción `ConEnsayo(bool)`*: olvidable; un adaptador que no la pase haría una petición real bajo
  `--dry-run`, que es exactamente lo que el criterio de aceptación del hito («no se puede usar mal»)
  prohíbe. Rechazada.
- *Leer `--dry-run` de un valor dentro de `context.Context`*: H1 decidió que las decisiones globales viajan
  en `schema.Contexto` y no en el contexto de cancelación (`specs/002` research D10); no se reabre.
- *Métodos `Get(ctx, …)` y `Head(ctx, …)` en vez de una operación con método*: dos superficies en lugar de
  una y, sobre todo, US5 escenario 5 exige que «un intento de hacer una petición con un método distinto de
  GET o HEAD» se rechace como error de argumentos: eso pide que el método sea un dato de la petición.
- *`Replay(dir)` sin opciones*: el enunciado lo escribe así, pero un cliente de reproducción también
  registra eventos (FR-035) y FR-043 exige rechazar «un cliente construido a la vez con raíz de grabación y
  con directorio de reproducción», que solo puede ocurrir si `Replay` acepta opciones. Se aceptan y se
  rechazan las que no tienen sentido en reproducción (`ConRaizDeGrabacion`, `ConIntervalo`,
  `ConIntentos`) con la clase «argumentos».

---

## D2 · La respuesta: cuerpo ya leído, cabeceras propias, dirección final y declaración de ensayo

**Decisión.** `Respuesta` entrega el cuerpo **ya leído en memoria** (`Cuerpo []byte`), nunca en
streaming. Consecuencias:

- El único `*http.Response` del módulo se abre y se cierra **dentro** de `internal/httpx`, donde
  `bodyclose` sí lo vigila (README de `bodyclose`: comprueba que `res.Body` se cierra; solo conoce
  `*http.Response`). Es la carga que el spec (*Assumptions*, «forma de la respuesta») pide evitar al
  adaptador: con streaming, cerrar el cuerpo quedaría en manos del adaptador y ningún analizador lo
  comprobaría.
- La grabación (FR-038, «el cuerpo se entrega íntegro y una sola vez») y la huella del sobre (H1, sha256
  del `data` canónico) necesitan el cuerpo entero de todos modos.
- Coste: memoria proporcional a la respuesta. Las respuestas del proyecto son documentos legales (XML o
  JSON del BOE, Atom de PLACSP), del orden de kilobytes a pocos megabytes. El límite de tamaño está
  fuera de alcance por el spec y no se añade.

`Cabeceras` es un tipo del paquete (`map[string][]string` con `Get` que canonicaliza el nombre igual que
`net/http`: `go doc net/http.CanonicalHeaderKey`) y no `http.Header`, para que un adaptador pueda leer
`Content-Type` sin importar `net/http` (R2, FR-053).

`URL` es la **dirección final** tras las redirecciones, distinta de `Peticion.URL` cuando hubo saltos
(FR-011). Existe porque quien llama no puede conocerla de otro modo y porque el sobre de H1 cita una
`url` (constitución §II): el adaptador de H4 citará la dirección que de verdad entregó el contenido.

`Ensayo` es verdadero solo bajo `--dry-run` (FR-065); entonces `Estado` es 0, `Cabeceras` y `Cuerpo` van
vacíos, `URL` es la dirección pedida y `Descripcion()` devuelve la línea «`GET https://…`» que el kernel
presenta (D6).

**Alternativas consideradas.** *Cuerpo en streaming (`io.ReadCloser`)*: menos memoria, pero traslada el
cierre al adaptador fuera del alcance de `bodyclose` y obliga a la grabación a envolver el lector;
rechazada por lo que dice el propio spec. *Devolver `*http.Response`*: prohibido por FR-002.
*`Cabeceras` como `http.Header`*: obligaría al adaptador a importar `net/http` para nombrar el tipo (R2).

---

## D3 · Cadena de decoradores sobre `http.RoundTripper`, y qué queda fuera de la cadena

**Decisión.** La cadena privada, de fuera adentro, es:

```
identificar → robots → reintentar → ritmo → (grabar → transporte | reproducir)
```

- **`identificar`** pone `User-Agent` (D5) en **toda** petición que baja por la cadena, incluido cada
  salto de redirección y cada reintento (FR-006, SC-001). Va arriba del todo para que la grabación,
  abajo, guarde la petición ya identificada (FR-037). Sin este decorador `net/http` enviaría
  `Go-http-client/1.1` (`$GOROOT/src/net/http/request.go:550`, `defaultUserAgent`), que es lo que el
  principio I prohíbe.

  Un decorador, sin embargo, **solo ve lo que baja desde arriba**: la petición de `robots.txt` la fabrica
  `robots`, que está por debajo, y no volvería a pasar por él. Por eso `identificar.go` es además el
  **único constructor de peticiones del paquete**: expone (dentro del paquete)
  `nuevaPeticionIdentificada(ctx, metodo, url)`, que crea el `*http.Request` con
  `http.NewRequestWithContext` y le fija la cabecera con `AgenteDeUsuario()` antes de devolverlo, y
  `ponerIdentificacion(r)`, que es lo que el decorador aplica. Ningún otro fichero de `internal/httpx`
  llama a `http.NewRequestWithContext`: el bucle de redirecciones de `cliente.go` y la obtención de
  `robots.txt` de `robots.go` construyen sus peticiones por ahí, de modo que **la identificación tiene un
  solo dueño** y ninguna petición que el paquete origine puede nacer sin ella (FR-009). Es una garantía
  por construcción, no un recordatorio: `TestSoloIdentificarConstruyePeticiones` recorre el paquete con
  `go/parser` y falla si otro fichero construye una petición, y
  `TestIdentificacionEnTodaPeticion` la comprueba en ejecución sobre las cuatro procedencias (recurso,
  saltos, reintentos y `robots.txt`).
- **`robots`** decide, por sitio, si la ruta se puede pedir (FR-013, FR-014) y cachea la decisión (FR-015,
  D7). Para obtener `robots.txt` construye la petición con `nuevaPeticionIdentificada` —ya identificada,
  FR-009— y la entrega a **su propio `siguiente`** —la cadena por debajo de él—, de modo que pasa por
  reintentos y ritmo (FR-017) pero **no** por sí mismo (FR-016).
- **`reintentar`** repite los fallos transitorios (FR-023 a FR-028, D9). Está **por encima** de `ritmo`
  para que **cada intento** espere su turno: un reintento es una petición más al mismo sitio, y ponerlo
  por debajo del limitador dejaría los reintentos fuera del ritmo. El enunciado («UA → robots → ratelimit
  → retry → record/replay», `docs/ROADMAP.md` §2) enumera los decoradores, no fija el orden relativo de
  esos dos; el elegido es el único que hace cierto FR-021 («no emitir peticiones … más deprisa de lo que
  tolera») también durante los reintentos.
- **`ritmo`** espera el turno del sitio de destino de la petición que baja (FR-019, FR-022, D8).
- **`grabar`** escribe petición y respuesta cuando la grabación está activa y entrega la respuesta intacta
  (D12); **`transporte`** es el `*http.Transport` (D11). En reproducción, el escalón inferior es
  **`reproducir`** (D13) y no existe transporte alguno: la cadena de un cliente de reproducción es solo
  `identificar → reproducir` (FR-049: sin `robots.txt`, sin ritmo, sin reintentos).

**Fuera de la cadena, en `Pedir`,** viven las tres cosas que no son un decorador porque necesitan
decidir **antes** o **después** de la cadena entera:

1. **Validación** (método GET/HEAD, dirección absoluta `http`/`https`, configuración): antes de abrir nada
   (FR-010, FR-064, FR-043).
2. **Ensayo** (`--dry-run`): devuelve sin bajar por la cadena (FR-050, D6).
3. **Redirecciones y clasificación del resultado**: `Pedir` es el bucle que sigue los 3xx, y cada salto
   baja por la cadena **entera** —identificación, `robots.txt` del sitio de destino, ritmo del sitio de
   destino— porque es una petición más (FR-011, D10). Al terminar clasifica el estado final (5xx → 4,
   429 → 5, el resto se entrega; FR-029, FR-030, FR-032).

La cadena se compone en `New`/`Replay` y no cambia después; el `http.Client` interno lleva
`CheckRedirect` devolviendo `http.ErrUseLastResponse` para que **no** siga redirecciones por su cuenta
(`go doc net/http.Client`: «if CheckRedirect returns ErrUseLastResponse, then the most recent response is
returned with its body unclosed, along with a nil error») y `Timeout: 0` (FR-004).

**Alternativas consideradas.** *Dejar que `http.Client` siga las redirecciones* (rechazada: aunque cada
salto pasaría por el `RoundTripper`, la política —qué cabeceras conserva, el tope, el cambio de método
en 303— quedaría en manos de la biblioteca, el bucle o el exceso de cadena llegarían como un `*url.Error`
genérico y no como la clase 4 que FR-011 exige, y el decorador `robots` no podría distinguir la cadena del
recurso pedido de la cadena del propio `robots.txt`, que FR-015 y FR-016 tratan de forma distinta).
*Reintentar por debajo del ritmo* (rechazada arriba). *Un único `RoundTripper` monolítico* (rechazada:
contradice el patrón Decorator que `docs/ROADMAP.md` §2 fija para este paquete y mezclaría cinco
responsabilidades).

Para la identificación de la petición de `robots.txt`, dos alternativas más: *poner `identificar` abajo
del todo, justo encima del transporte*, para que toda petición pase por él venga de donde venga
(rechazada: la grabación quedaría **por encima** o **por debajo** del decorador —si queda por encima
guarda una petición sin identificar y rompe FR-037 y el fixture #10; si queda por debajo, la cadena deja
de ser legible de fuera adentro— y el orden del enunciado del roadmap cambiaría sin necesidad); y
*repetir en `robots.go` la línea que fija la cabecera* con `AgenteDeUsuario()` (rechazada: la garantía
tendría dos dueños y nada impediría que un tercer origen de peticiones la olvidara; el constructor único
la hace imposible de olvidar y además comprobable con `go/ast`).

**Enmienda tras la revisión final (ronda 1).** La revisión comprobó por mutación que quitar el decorador
`identificar` de la cadena no hace fallar ningún test: toda petición que baja por ella nace ya
identificada del constructor único, así que el decorador vuelve a poner una cabecera que ya está puesta.
Lo que los tests detectan es vaciar la cabecera en `ponerIdentificacion` —que constructor y decorador
comparten— y construir la petición de `robots.txt` sin el constructor. Se consideró **retirar el
decorador** y dejar el constructor como único mecanismo (rechazada: la cadena `UA → robots → ratelimit →
retry → record/replay` es la decisión de arquitectura registrada en `docs/ROADMAP.md` §2, que prevalece
sobre este documento; apartarse de ella exige un ADR y una decisión humana, no una corrección de revisión).
Se mantiene, por tanto, con su papel dicho con exactitud: **el dueño de la identificación es el
constructor**; el decorador es el escalón `UA` del roadmap, redundante a propósito, que reafirma en la
propia cadena que nada de lo que la atraviesa sale sin identificar. Su sitio en las dos cadenas lo fija un
test de forma, `TestCadenaEmpiezaPorLaIdentificacion`, porque ninguna medida podría hacerlo; el control 6
del plan describe lo que cada test detecta de verdad.

---

## D4 · Errores tipados: `httpx.Error` declara su clase; el kernel la reconoce con `errors.As` sin que el adaptador importe `internal/cli`

**Decisión.** Tres piezas, y ningún adaptador importa `internal/cli`:

1. **`internal/core/schema` gana una interfaz**, dominio puro, sin importaciones nuevas:

   ```go
   // ConClase lo implementa un error que declara la clase con la que el kernel debe
   // traducirlo a código de salida. Es el puerto por el que un adaptador nombra su
   // clase sin importar el kernel.
   type ConClase interface {
       error
       Clase() Clase
   }
   ```

2. **`internal/httpx` define `Error`**, un tipo con datos —clase, petición (método y dirección), estado
   de la fuente, `Retry-After` si lo hubo (FR-030), causa envuelta— que implementa `schema.ConClase` y
   `Unwrap`. Cada ruta de fallo del cliente construye un `*Error` con la clase que FR-029, FR-030 y
   FR-063 le asignan; `Clase()` solo devuelve valores del vocabulario de `schema.Clases()`.
3. **`cli.Clasificar` gana una comprobación**, después de los cinco sentinelas de H1 y antes de la rama
   por defecto: si `errors.As(err, &conClase)` encuentra un `schema.ConClase` en la cadena **y** su clase
   está en `schema.Clases()`, esa es la clase; si no, sigue siendo «inesperado». `codigoDeClase` no cambia
   (sigue sin `default`, vigilado por `exhaustive`).

Envolver un `*httpx.Error` con `fmt.Errorf("…: %w", err)` no cambia la clase, porque `errors.As`
recorre la cadena de `Unwrap` (`go doc errors.As`: «finds the first error in err's tree that matches
target … obtained by repeatedly calling its Unwrap() error»): es FR-031. `errorlint` —activo desde H0—
obliga a que toda comparación pase por `errors.Is`/`errors.As`.

**Por qué no wrapear los sentinelas de `internal/cli`.** Es la solución más corta —`fmt.Errorf("%w",
cli.ErrLimiteOTos)`— y obliga a que `internal/httpx` importe `internal/cli`: un adaptador secundario (el
que sale a la red) dependiendo del adaptador primario (el que analiza la línea de órdenes y arrastra
Kong). Ninguna regla de dependencia lo prohíbe, pero invierte la dirección que la arquitectura hexagonal
(constitución §IV) da por sentada —los adaptadores dependen del dominio, no entre sí— y es lo que el spec
pide evitar en *Assumptions* («dónde viven exactamente los errores tipados para que el kernel los
clasifique sin que un adaptador importe `internal/cli`»). Además `Retry-After` **tiene** que viajar con
el error «para que quien llama pueda decidir sin analizar una cadena de texto» (FR-030, US2 escenario 3):
un sentinela envuelto no puede llevar datos; un tipo con `errors.As` sí. El roadmap §2 describe los
errores tipados como «sentinel + `errors.As`»: los sentinelas ya estaban (H1) y `errors.As` es lo que H2
aporta.

**Por qué no mover los sentinelas a `internal/core/schema`.** Resolvería lo mismo para las clases sin
datos, pero H1 lo consideró y lo rechazó explícitamente (`specs/002` research D8: «definir los sentinelas
en `core` … los errores tipados son del kernel según §IV y el roadmap §2»), y `docs/ROADMAP.md` §2 sitúa
los sentinelas en `internal/cli/errors.go`. Reabrirlo no aporta nada que `schema.ConClase` no dé, y
seguiría sin transportar `Retry-After`.

**Orden de la clasificación: sentinelas primero.** Un `*httpx.Error` nunca envuelve un sentinela de
`internal/cli` (no lo importa), así que en la práctica los dos mecanismos no compiten; que los sentinelas
vayan primero conserva byte a byte el comportamiento de H1 para todo lo que ya existía.

**Verificación.** `errors.As` y su recorrido de `Unwrap`: `go doc errors.As`. Que `exhaustive` sigue
vigilando el `switch` de códigos: `.golangci.yml` (`default-signifies-exhaustive`, H1). Que
`internal/core/schema` puede declarar una interfaz con `error` embebido sin importar nada: `error` es un
tipo predeclarado del lenguaje.

---

## D5 · La versión de la identificación: una variable de paquete inyectada por `-ldflags`, con la misma `VERSION` del `Makefile` que alimenta al verbo `version`

**Decisión.** `internal/httpx` declara `var version = "dev"` y `AgenteDeUsuario()` devuelve
`"kitlegal/" + version + " (+https://ventanillalegal.es/bot)"`. El `Makefile` amplía `LDFLAGS` con
`-X github.com/jmorenobl/kitlegal/internal/httpx.version=$(VERSION)`, de modo que **la misma `VERSION`**
que ya inyecta `main.version` (H0, `contracts/cli-version.md`) llega a la identificación (FR-007). El
valor por omisión es `"dev"`, **el mismo** que `cmd/kitlegal/main.go` y que el `main` del binario de e2e
declaran (`version = "dev"`): una construcción sin `-ldflags` —`go test`, `go run`, el binario del e2e—
identifica como `kitlegal/dev (+…)`, que es exactamente lo que ese binario imprime en `version`.

**Controles.** Un test en `cmd/kitlegal/main_test.go` (paquete `main`, que ve `version`) comprueba que
`httpx.AgenteDeUsuario()` contiene `"kitlegal/" + version + " "`: los dos valores por omisión no pueden
divergir sin que `make ci` lo diga. Ese test importa `internal/httpx` **solo desde un fichero de test**:
`go list -deps ./cmd/kitlegal` no incluye las importaciones de test (`go help list`: `-deps` recorre «all
their dependencies»; los binarios de test solo entran con `-test`), así que `TestDependenciasDelBinario`
(`internal/arch_test.go`) sigue viendo la misma lista de módulos y el binario distribuido **no enlaza**
`x/time` ni `robotstxt` en H2 (SC-014).

La forma exacta de la cadena la comprueba `TestAgenteDeUsuario` (`internal/httpx/agente_test.go`,
paquete `httpx`, que ve `version`) con una expresión regular **construida sobre la variable**, no sobre
el literal `dev`, y con el sufijo escapado en tiempo de ejecución en lugar de escrito a mano:
`"^kitlegal/" + regexp.QuoteMeta(version) + regexp.QuoteMeta(" (+https://ventanillalegal.es/bot)") + "$"`,
más la comprobación de que `version` no está vacía. El sufijo aparece así en el fuente del test como el
literal ` (+https://ventanillalegal.es/bot)` —sin barras invertidas—, de modo que la comprobación
mecánica «solo direcciones locales» del quickstart (prerrequisitos) extrae de él exactamente
`https://ventanillalegal.es/bot` y lo descarta por su tercer filtro; una expresión escrita a mano
(`\(\+https://ventanillalegal\.es/bot\)$`) dejaría en el fichero el token `https://ventanillalegal\.es/bot\`,
que ese filtro no reconoce. Así el test vale con cualquier valor inyectado por
`-ldflags` —y es lo que hace demostrable el camino completo—: el test **registra** la identificación
con `t.Log(httpx.AgenteDeUsuario())`, de modo que `go test -v -ldflags '-X …/internal/httpx.version=9.9.9-prueba'`
muestra la línea `kitlegal/9.9.9-prueba (+https://ventanillalegal.es/bot)` en el registro del test
**sin que el test falle**, que es lo que filtra [quickstart.md escenario 2](./quickstart.md). Un test que
comparara contra el literal `dev` fallaría con la versión inyectada y no demostraría nada.

**Alternativas consideradas.** *Pasar la versión en `schema.Contexto`*: el dominio cargaría con un dato
de construcción, `cli.Globales.Contexto()` necesitaría un parámetro nuevo, y un `Contexto{}` vacío
—válido en cualquier test— produciría una identificación sin versión, contra FR-007. *Una opción
`ConVersion`*: olvidable, y FR-001 exige que el cliente sin opciones ya identifique. *`runtime/debug.ReadBuildInfo`*:
no transporta valores de `-ldflags -X`, y `Main.Version` es `(devel)` en una construcción local: no sería
«la misma que imprime `version`». *Que `cmd/kitlegal` lea la versión de `internal/httpx`* (una sola
variable): enlazaría `x/time` y `robotstxt` en el binario de H2 sin ningún consumidor, y `main` es donde
`-ldflags` ya la deja desde H0.

---

## D6 · `--dry-run`: `Pedir` devuelve una respuesta de ensayo; la descripción viaja en `schema.Resultado.Ensayo` y la presenta el kernel

**Decisión.** Bajo `ejecucion.DryRun`, `Pedir` comprueba lo que no necesita red (método, dirección,
configuración: FR-050, FR-043, FR-064) y devuelve `Respuesta{Peticion: p, URL: p.URL, Ensayo: true}` sin
error (FR-065). Su `Descripcion()` es «`GET https://…`».

Para que esa línea llegue a la **salida de error, siempre visible, sin depender del nivel de registro**
(FR-051), el único canal que H1 dejó abierto es el presentador, y el único camino de un adaptador al
presentador pasa por el kernel. Por eso:

- `schema.Resultado` gana un campo `Ensayo []string`: «lo que cada capa con efectos habría hecho, una
  línea por petición; solo se rellena bajo `--dry-run`». El adaptador de prueba (y los de H4) copia ahí
  `respuesta.Descripcion()` de cada petición que habría emitido.
- `internal/app` (`conDescripcion`, `descripcionDeLaOperacion`) presenta, tras su línea de H1
  («`--dry-run: no se ha ejecutado nada; se habría ejecutado el applet …`»), cada línea de
  `resultado.Ensayo` con el prefijo «`se habría pedido `». Es la implementación literal de lo que el
  contrato de H1 ya decía: «cada capa con efectos la honra describiendo en lugar de ejecutar»
  (`specs/002/contracts/banderas-y-exit-codes.md` §3).
- El código de salida sigue siendo 0 (FR-065): `Pedir` no devuelve error por estar en ensayo, y el kernel
  no cambia su traducción.

Es un cambio del contrato del applet (ADR 0005 muestra `Resultado{Procedencia, Datos}`), así que se
registra en un ADR corto, `docs/ADR/0011-ensayo-por-capas.md`, que **complementa** al 0005 sin
sustituirlo. Es exactamente el caso que el spec prevé en *Fuera de alcance* («solo haría falta un ADR si
este hito se apartara de lo ya decidido»).

**Alternativas consideradas.** *Emitir la descripción por `slog`* (rechazada por FR-051 y por la regla que
H1 fijó en su D10: un registro es filtrable por nivel y `KITLEGAL_LOG=error` la ocultaría). *Un
`io.Writer` de avisos como opción del cliente* (rechazada: olvidable, y sin valor por omisión posible: el
cliente no puede nombrar `os.Stderr` —R5, `forbidigo`— ni recibirlo del adaptador, que solo tiene un
`*slog.Logger`). *Que el kernel imprima `Datos` bajo ensayo* (rechazada: `Datos` es `any` y un
`fmt.Sprintf("%v")` de un valor arbitrario no es una descripción). *Ampliar la firma de `Ejecutar` con un
presentador* (rechazada: cambia el contrato de todos los applets para una línea de texto, y daría a los
applets un canal de escritura que FR-044 de H1 les niega).

---

## D7 · `robots.txt`: agente `kitlegal`, `FromStatusAndBytes`, la lista cerrada de FR-015 y una caché por sitio con exclusión mutua

**Decisión.**

- **Agente para las reglas**: `kitlegal`, el producto de la identificación (`kitlegal/<versión>`).
  `robotstxt.RobotsData.FindGroup` compara en minúsculas y elige el grupo con el prefijo de agente más
  largo, con `*` como comodín más débil (`robotstxt.go:158-179`), así que un `User-agent: kitlegal`
  gana a `User-agent: *` y `User-agent: kit` también casa.
- **Ruta evaluada**: `url.URL.RequestURI()` (ruta codificada más consulta; `"/"` si la ruta está vacía,
  `$GOROOT/src/net/url/url.go:1171-1183`), que es lo que RFC 9309 §2.2.2 compara.
- **Solo permiso y denegación de rutas**: `Group.Test(path)` (`robotstxt.go:181`). `CrawlDelay` existe
  en el tipo `Group` (`robotstxt.go:32`) y **no se lee**, como fija el spec (*Fuera de alcance*).
- **Casos de obtención (FR-015, lista cerrada)**, y cómo los produce el código:

  | Terminación de la petición de `robots.txt` | Qué hace `temoto/robotstxt` | Qué hace `httpx` |
  |---|---|---|
  | 2xx con cuerpo interpretable | `FromBytes` | reglas del sitio; evaluar |
  | 2xx con cuerpo vacío | `FromBytes` de vacío → sin grupos → `Test` permite | «sin reglas»: permitir |
  | 2xx con cuerpo no interpretable | `FromBytes` devuelve `error` | denegar, clase **5** |
  | 4xx distinta de 429 (404 incluida) | `FromStatusAndBytes` → `allowAll` (`robotstxt.go:74-75`) | «sin reglas»: permitir |
  | **429** | `FromStatusAndBytes` lo trataría como 4xx → `allowAll` | **no se le pasa**: `httpx` lo intercepta antes y deniega, clase **5** con `Retry-After` (FR-030 prevalece) |
  | 5xx que sobrevive a los reintentos | `FromStatusAndBytes` → `disallowAll` (`robotstxt.go:80-81`) | denegar, clase **5** |
  | fallo de transporte que sobrevive a los reintentos | — | denegar, clase **5** |
  | 3xx | — | `httpx` sigue la cadena con el tope de D10, sin evaluar `robots.txt` en los saltos (FR-016); la respuesta final entra en esta tabla |
  | cadena de 3xx en bucle o excedida | — | denegar, clase **5** (FR-015, no FR-011) |
  | vencimiento o cancelación del contexto | — | la operación entera termina, clase **4** (FR-029) |
  | otro estado (1xx, 3xx sin `Location`) | `FromStatusAndBytes` devuelve `error` («Unexpected status») | denegar, clase **5**: el permiso no se pudo obtener |

  La fila del 429 es la razón de no llamar a `FromStatusAndBytes` a ciegas: la biblioteca implementa la
  regla de Google «todo 4xx es permitir» y el spec exige lo contrario para el 429.
- **Caché por sitio, en memoria, sin caducidad** (FR-015, FR-018): un `map[string]*sitio` protegido por
  un `sync.Mutex`, donde `sitio` guarda las reglas —o el resultado de denegación— y su propio
  `sync.Mutex`, que se mantiene tomado mientras se obtiene `robots.txt`. Así, N goroutines que piden a la
  vez el mismo sitio producen **una** petición de `robots.txt` y N esperas (SC-003, FR-021, D14). El
  `robots.txt` de un salto de redirección se obtiene con el mutex del sitio de **destino**, nunca con dos
  a la vez: no hay ciclo de espera.
- **Clave de sitio**: esquema + host en minúsculas + puerto, con el puerto por omisión del esquema si la
  dirección no lo declara (`url.URL.Port()` devuelve vacío en ese caso), tal como fija la entidad *Sitio*
  del spec. `http://fuente.prueba` y `https://fuente.prueba` son sitios distintos; dos servidores
  locales de pruebas en `127.0.0.1` con puertos distintos, también (SC-004). La clave es la base común
  de la caché de `robots.txt` (FR-013, FR-015) y del limitador (FR-019), así que tiene su propio test de
  tabla, `TestClaveDeSitio` (`sitio_test.go`), escrito **solo con los hosts ficticios que admite la
  comprobación «solo direcciones locales» del quickstart** (`fuente.prueba` y `otra.prueba`, en
  cualquier combinación de mayúsculas; ningún `x`, `example.com` ni host real): esquema `http`/`https`;
  `http://Fuente.Prueba/ruta?q=1` → `http://fuente.prueba:80` (host en minúsculas, ruta y consulta
  fuera); puerto explícito conservado (`http://fuente.prueba:8080`); `http://fuente.prueba` y
  `http://fuente.prueba:80` misma clave, `https://fuente.prueba` y `https://fuente.prueba:443` misma
  clave; `http://fuente.prueba` ≠ `https://fuente.prueba`; `http://fuente.prueba` ≠ `http://otra.prueba`;
  `http://127.0.0.1:53211` ≠ `http://127.0.0.1:53212`. Es lo que hace comprobables SC-004 y FR-013 sin
  levantar un servidor.

**Alternativas consideradas.** *`FromResponse(*http.Response)`* (rechazada: lee y cierra el cuerpo por su
cuenta y aplica la regla del 4xx que hay que interceptar). *Caché con `sync.Map`* (rechazada: no permite
la exclusión por sitio durante la obtención, que es lo que garantiza «una sola petición» bajo
concurrencia). *`singleflight`* (rechazada: `golang.org/x/sync` no está en la lista de la constitución
§V, y un mutex por sitio da lo mismo con la biblioteca estándar).

---

## D8 · Ritmo por sitio con `x/time/rate`: `NewLimiter(Every(intervalo), 1)`, un segundo por omisión

**Decisión.** Cada sitio tiene su `*rate.Limiter` construido con `rate.NewLimiter(rate.Every(intervalo),
1)`: un cubo de **un** token que se rellena cada `intervalo` (`go doc … ./rate Every`: «converts a minimum
time interval between events to a Limit»; `NewLimiter(r, b)`: «permits bursts of at most b tokens»).
Con ráfaga 1, dos peticiones seguidas al mismo sitio llegan separadas **al menos** `intervalo`, que es
lo que SC-004 mide. Antes de emitir, `ritmo` llama a `Wait(ctx)`, que «blocks until [a token] can be
obtained or its associated context.Context is canceled» (`go doc … ./rate Limiter`): FR-022 sin código
propio. `Limiter` «is safe for simultaneous use by multiple goroutines» (misma fuente): FR-021.

**Valor por omisión: `1 * time.Second`** (una petición por segundo y sitio). Es conservador —más lento que
lo que cualquier fuente pública del proyecto tolera— y, al ser uno por cliente (FR-020), cada adaptador lo
ajusta con `ConIntervalo` desde lo que declare `docs/SOURCES.md` en H4. `ConIntervalo(0)` o negativo se
rechaza con la clase «argumentos»: un ritmo nulo es lo contrario del principio I.

**Por qué no un cubo mayor.** Una ráfaga de N permitiría N peticiones instantáneas y luego el ritmo; para
un cliente que vive segundos, eso es «ir más deprisa de lo que tolera» en el peor momento —el arranque—.

**Alternativas consideradas.** *`time.Ticker` por sitio* (rechazada: hay que implementar la espera con
contexto y la concurrencia a mano; y el hito nombra `x/time/rate`). *`Reserve` + `time.Sleep`* (rechazada:
`Wait` ya respeta el contexto).

---

## D9 · Reintentos: tres intentos, retardo exponencial con *equal jitter*, espera interrumpible por el contexto

**Decisión.**

- **Intentos por omisión: 3** (una petición y dos reintentos); `ConIntentos(n)` con `n ≥ 1`; `n < 1` es
  «argumentos».
- **Qué se reintenta** (FR-025, lista cerrada): estado 5xx y fallo de transporte (el error que devuelve
  el `RoundTripper` inferior) **salvo** que el contexto haya terminado (`ctx.Err() != nil`: FR-026, la
  cancelación gana siempre). **No** se reintenta ningún 4xx, 429 incluido, ni ningún 2xx/3xx.
- **Retardo**: base 500 ms, factor 2, techo 30 s, con *equal jitter*: para el intento *n* (desde 1) la
  espera es `b/2 + aleatorio(0, b/2)` con `b = min(500ms·2^(n-1), 30s)`. Elegido frente al *full jitter*
  porque garantiza a la vez las dos cosas que SC-005 exige: **crecen** (el mínimo del intento *n+1*,
  `b_n`, es igual al máximo del intento *n*) y **no son idénticas entre ejecuciones** (mitad aleatoria).
- **Aleatoriedad: `crypto/rand.Read` sobre ocho bytes**, reducidos con `encoding/binary`:

  ```go
  var octetos [8]byte
  rand.Read(octetos[:])                                        // nunca devuelve error (ver abajo)
  aleatorio := time.Duration(int64(binary.BigEndian.Uint64(octetos[:])>>1)) % mitad
  espera := mitad + aleatorio                                  // mitad = b/2 ≥ 250 ms por construcción
  ```

  Un jitter no necesita criptografía, pero `gosec` v2.28.0 —gate desde H0, sin supresiones (SC-010)—
  marca con G404 **también** `math/rand/v2` (`rules/rand.go` líneas 31-36: `New`, `Float64`, `Int`,
  `IntN`, `Int64N`, `N`, `Uint64N`…), de modo que la única forma de pasar `make ci` con `math/rand/v2`
  sería un `//nolint`, que este plan prohíbe. Dentro de `crypto/rand`, la forma elegida es la que **no
  deja ninguna rama de error**: `Read` «never returns an error» (`crypto/rand/rand.go` 44-46), y por eso
  `errcheck` lo excluye por omisión (`excludes.go` 21) y G104 lo tiene en su lista blanca
  (`rules/errors.go` 88); el resultado se ignora sin `_ =`, sin exclusión nueva y sin código que la
  cobertura no pueda alcanzar. El desplazamiento `>> 1` descarta el bit alto para que la conversión a
  `int64` tenga un máximo que G115 reconoce como seguro (`range_analyzer.go` 790-805,
  `conversion_overflow.go` 230-233); el módulo es sobre `time.Duration`, sin conversión. `mitad` nunca es
  cero: la base de 500 ms es fija y no hay opción para cambiarla, así que `b ≥ 500 ms` y `mitad ≥ 250 ms`
  para todo `n ≥ 1` (lo cubre `TestReintentosDosErroresYUnAcierto`, cuyas esperas registradas tienen que
  estar en `[b/2, b)`). El sesgo del módulo sobre 63 bits es inferior a `mitad / 2^63 < 2·10⁻⁹` con el
  techo de 15 s: irrelevante para un jitter, y SC-005 solo exige que las esperas crezcan y difieran.

  **Lo que ninguna función de `crypto/rand` ofrece, y este plan no afirma**: una ruta de error manejable
  ante un fallo de la fuente de entropía del sistema. En go1.27.1 el `Read` de `rand.Reader` llama a
  `drbg.Read(b)` y devuelve siempre `len(b), nil` (`crypto/internal/rand/rand.go` 22-31); `drbg.Read`
  termina en `sysrand.Read`, que «crashes the program irrecoverably if an error is encountered»
  (`crypto/internal/sysrand/rand.go` 31-37, 50-53), exactamente igual que `crypto/rand.Read`
  (`crypto/rand/rand.go` 44-64). Por tanto `rand.Int(rand.Reader, …)` **tampoco** devuelve nunca ese
  error: el `error` de su firma solo puede venir de un lector distinto de `rand.Reader`, que este paquete
  no usa. La documentación declara el fallo imposible «on all but legacy Linux systems» (núcleos
  anteriores a 3.17 y AIX, donde se recurre a `/dev/urandom`; `sysrand/rand.go` 56). Ese suceso no es
  una «ruta de fallo» del paquete en el sentido de FR-033 —ninguna función de `internal/httpx` lo
  produce ni lo puede interceptar; es una terminación del proceso por el runtime, como un fallo de
  memoria—, y se deja declarado aquí en vez de fingir una contingencia que no puede ejecutarse ni
  cubrirse en test.
- **Espera interrumpible**: `select { case <-ctx.Done(): …; case <-time.After(d): }` (skill
  `golang-design-patterns`, «Retry logic MUST check ctx.Err() between attempts and use … backoff via
  select on ctx.Done()»). El vencimiento durante la espera produce la clase **4** (FR-005, FR-029).
- **Cuerpos descartados**: el cuerpo de cada respuesta que se va a reintentar se lee y se cierra antes
  del siguiente intento (FR-027); `bodyclose` lo vigila dentro del paquete.
- **Cada intento clona la petición** (`req.Clone(ctx)`, `go doc net/http.Request.Clone`: «returns a
  deep copy of r with its context changed to ctx»): GET y HEAD no tienen cuerpo, así que la copia es
  completa.
- **Reloj inyectable, no exportado**: el decorador espera a través de un campo `dormir func(ctx,
  time.Duration) error` que los tests del propio paquete sustituyen por uno que **registra** las
  duraciones sin dormir. Así SC-005 se comprueba sobre las duraciones registradas (crecen, difieren entre
  dos ejecuciones) sin esperar de verdad y sin exponer nada en la API (FR-002).
- **Número de peticiones observable** (FR-028): se cuenta en el servidor local de pruebas, que ve cada
  petición real; no hace falta contador en la API.

**Alternativas consideradas.** *Full jitter* (rechazada: dos esperas consecutivas pueden decrecer, y
SC-005 pide que crezcan). *Sin jitter* (rechazada por FR-023 y SC-005). *`math/rand/v2`* (rechazada:
G404 de `gosec` v2.28.0 la marca, verificado en `rules/rand.go`; la variante con generador propio,
`rand.New(rand.NewChaCha8(semilla))`, también, porque `New` está en la lista; y ninguna supresión es
admisible). *`crypto/rand.Int(rand.Reader, big.NewInt(int64(mitad)))` con `math/big`* (rechazada: con
`rand.Reader` su `error` es inalcanzable —el lector nunca lo devuelve, ver arriba—, así que manejarlo
sería una rama sin cobertura posible y **no** manejarlo, un `_` sobre un error que `errcheck` no excluye;
además hace `panic` si `max <= 0` (`crypto/rand/util.go` 71-74), una ruta más que vigilar, y asigna un
`*big.Int` por espera. Ante un fallo de entropía termina el proceso igual que `Read`, de modo que no
aporta ninguna garantía a cambio). *`Read` sobre cuatro bytes* (rechazada: `mitad` llega a 15 s, más de
2³² ns, y el módulo dejaría de cubrir el intervalo). *Reintentar el 429 esperando `Retry-After`* (rechazada en el spec,
*Assumptions*). *`time.Sleep`* (rechazada: no respeta el contexto, FR-005). *Exponer `ConEsperaBase`*
(rechazada: el spec solo hace ajustable el número de intentos; el reloj se inyecta desde el test del
paquete).

**Verificación.** G404 sobre `math/rand/v2`: `rules/rand.go` líneas 31-36 de `gosec` v2.28.0, el mismo
módulo del que se leyó `rules/tls.go` (D11). Comportamiento de `crypto/rand.Reader`, `Read` e `Int` ante
un fallo de entropía: leído en el **código** del toolchain fijado, no solo en `go doc` (rutas y líneas en
la tabla de verificación, fila `crypto/rand`). `errcheck` y G104 sobre `rand.Read`, y G115 sobre
`int64(uint64 >> 1)`: filas `errcheck`, `gosec` G104 y `gosec` G115 de la misma tabla.

**Resultado (implementación, T018).** La propiedad «crecen» tal como la enuncia esta decisión —el mínimo
del intento *n+1* es el máximo del intento *n*— es exacta **mientras la base dobla**, que con 500 ms,
factor 2 y techo 30 s son las seis primeras esperas (bases de 500 ms a 16 s). En la séptima la base pide
32 s y el techo la deja en 30 s, así que su banda `[15 s, 30 s)` se solapa con la sexta, `[8 s, 16 s)`, en
`[15 s, 16 s)`: dos esperas correctas pueden salir en orden inverso. SC-005 mide tres intentos (dos
esperas) y nunca llega ahí, y el producto aplica la ley tal cual; lo que no se sostenía era la aserción de
orden estricto de `TestReintentosAgotados` sobre las siete esperas del caso de ocho intentos, que fallaba en
≈ 3 de cada 1000 ejecuciones (`gates/tarea-T018.md`). El test la acota al tramo en que la base dobla,
derivado con la misma ley literal que `exigeEsperaDelIntento`; la banda `[b/2, b)` sigue comprobándose en
las siete. La decisión no cambia.

---

## D10 · Redirecciones: tope de 10 saltos, `Location` resuelta con `ResolveReference`, bucle detectado por dirección repetida

**Decisión.** `Pedir` sigue **301, 302, 303, 307 y 308** con cabecera `Location` —los cinco que
`net/http` sigue (`go doc net/http.Client.Get`)—, conservando el método: GET sigue GET y HEAD sigue HEAD
en los cinco (para 301/302/303 es lo que hace `net/http`; para 307/308 el método se conserva por
definición; sin cuerpo no hay diferencia). La dirección de destino se resuelve con
`url.URL.ResolveReference` (RFC 3986 §5.2, relativa o absoluta). Cada salto es una petición completa por
la cadena (D3): identificación, `robots.txt` **del sitio de destino** y ritmo **del sitio de destino**
(FR-011; el destino no hereda permiso).

- **Tope: 10 saltos**, el mismo que la política por omisión de `net/http` (`go doc net/http.Client`:
  «stop after 10 consecutive requests») y por encima del mínimo de cinco de RFC 9309 §2.3.1.2 que FR-015
  exige para el `robots.txt`, que reutiliza este tope.
- **Bucle**: una dirección ya visitada en la misma cadena termina la operación en ese momento, sin
  agotar el tope: es más rápido y el resultado es el mismo. Exceso de cadena y bucle producen, para el
  recurso pedido, la clase **4** (FR-011, FR-063); para la cadena del propio `robots.txt`, la clase **5**
  (FR-015).
- **3xx sin `Location`, o 300/304/305**: no se puede seguir. Se clasifica como la fuente que «no sabe
  entregar el recurso»: clase **4**. FR-032 dice que un 3xx no es entregable; una redirección que no dice
  adónde tampoco lo es.
- El cuerpo de cada respuesta 3xx intermedia se lee y se cierra antes del salto (FR-027).

**Alternativas consideradas.** *Tope de 5* (cumple RFC 9309, pero cambia el comportamiento que cualquier
usuario de `net/http` espera y no hay razón para ser más estricto que la biblioteca). *Entregar el 3xx
sin `Location` como respuesta* (rechazada: FR-032 lo excluye).

---

## D11 · Transporte: clon de `http.DefaultTransport`, sin plazo propio, sin redirecciones propias, TLS intocable

**Decisión.** `New` construye el escalón inferior con `http.DefaultTransport.(*http.Transport).Clone()`
(`go doc net/http.Transport.Clone`: «deep copy of t's exported fields») y un `http.Client{Transport:
cadena, Timeout: 0, CheckRedirect: func(…) error { return http.ErrUseLastResponse }}`.

- **`Timeout: 0`**: «A Timeout of zero means no timeout» (`go doc net/http.Client`): el plazo es solo el
  del contexto (FR-004). Cada petición se crea con `http.NewRequestWithContext(ctx, …)`, que es además lo
  que `noctx` exige (`noctx.go`: `net/http.NewRequest` «must not be called. use
  net/http.NewRequestWithContext»).
- **Lo que hereda del transporte por omisión** (`go doc net/http.DefaultTransport`): `Proxy:
  ProxyFromEnvironment`, `Dialer{Timeout: 30s, KeepAlive: 30s}`, `ForceAttemptHTTP2`, `MaxIdleConns:
  100`, `IdleConnTimeout: 90s`, `TLSHandshakeTimeout: 10s`, `ExpectContinueTimeout: 1s`. Los dos plazos
  del transporte (conexión y saludo TLS) **no son un plazo de la operación**: acotan un intento de
  conexión, y su vencimiento es un fallo de transporte que FR-025 reintenta y que el contexto sigue
  acotando por encima. H2 no los cambia: no lo pide el hito, y quitarlos dejaría un intento de conexión
  contra un host que descarta paquetes colgado hasta el vencimiento del contexto, que puede ser largo.
  La lectura de proxy desde el entorno es comportamiento por omisión de la biblioteca estándar; H2 ni lo
  añade ni lo retira (el spec deja el «proxy corporativo» fuera de alcance: no se construye nada para él).
- **TLS**: `TLSClientConfig` no se toca y ninguna opción llega a él (FR-012). `gosec` G402
  (`rules/tls.go` en gosec v2.28.0) marcaría `InsecureSkipVerify: true` si alguien lo escribiera, y
  `gosec` es gate desde H0.
- **Sin cookies**: `Jar: nil` (fuera de alcance por el spec).

**Alternativas consideradas.** *`&http.Transport{}` limpio* (rechazada: pierde HTTP/2, la gestión de
conexiones ociosas y los plazos de conexión; ganaría «cero herencia» a cambio de un cliente peor).
*Usar `http.DefaultTransport` sin clonar* (rechazada: estado global compartido, contra «sin globals» de
`docs/ROADMAP.md` §2).

---

## D12 · Grabación: activación exacta, validación en `New`, nombre derivado, formato estable y escritura atómica

**Decisión.**

- **Activación** (FR-041): `New` lee `os.LookupEnv("KITLEGAL_RECORD")` una vez. `"1"` activa; ausente o
  vacía desactiva; **cualquier otro valor es «argumentos»** (código 2): una variable mal escrita no debe
  grabar ni dejar de grabar en silencio. La lectura vive en `New` y no en la raíz de composición —a
  diferencia de `KITLEGAL_LOG` en H1— porque el cliente lo construye el adaptador, no el kernel, y una
  opción para inyectar el entorno sería una forma de activar la grabación desde código, que FR-041
  prohíbe. Los tests que la necesitan usan `t.Setenv` y **no** `t.Parallel()`: `paralleltest` v1.0.15
  reconoce la llamada a `Setenv` y no exige `Parallel` en ese test (`paralleltest.go:149-150` detecta la
  llamada y `paralleltest.go:208` omite el aviso cuando `funcCantParallelMethod` es verdadero);
  `go doc testing.T.Setenv` confirma que «cannot be used in parallel tests».
- **Validación en `New`** cuando la grabación está activa (FR-039, FR-064, FR-042, «lo antes posible»):
  falta `ConFuente` → «argumentos» nombrando la opción; falta `ConRaizDeGrabacion` → «argumentos»
  nombrando la opción; la raíz no existe o no es un directorio (`os.Stat`) → «argumentos»; no se puede
  crear `<raíz>/<fuente>` (`os.MkdirAll`) → «argumentos» («no se puede escribir»). Nada de esto escribe
  un fichero de grabación. Un fallo posterior de escritura (E/S, disco lleno) → «inesperado» (código 1).
- **Nombre del fichero** (FR-039), derivado por el decorador de **método y dirección completa**:
  `<MÉTODO>_<esquema>_<host>[-<puerto>]<ruta con "/" → "_">[_q_<consulta>]`, saneado a `[A-Za-z0-9._-]`
  (todo lo demás → `_`, repeticiones colapsadas, extremos recortados; el host en minúsculas, la ruta
  como venga), y si supera 120 caracteres, los 100 primeros más `-` y los 8 primeros hexadecimales de
  `sha256(método + " " + dirección)`. Extensión `.json`. Ejemplos en
  [`contracts/formato-de-grabacion.md`](./contracts/formato-de-grabacion.md) §2. Es legible, determinista y
  portable (sin `<>:"/\|?*`, sin espacios). Dos direcciones que solo difieren en mayúsculas de la ruta
  producen dos nombres distintos que **un sistema de ficheros insensible a mayúsculas** (macOS por
  omisión) confunde: es una de las razones de que la colisión se detecte **por contenido** y no por
  nombre.
- **Colisión** (FR-039): antes de escribir, si el fichero existe se lee y se comparan método y dirección;
  distintos → «inesperado» nombrando las dos peticiones; iguales → se sustituye.
- **Formato** (FR-037, FR-040): un objeto JSON con claves fijas en orden fijo —`formato`, `grabado_en`,
  `peticion{metodo, url, cabeceras}`, `respuesta{estado, cabeceras, cuerpo | cuerpo_base64}`—
  serializado desde `struct`s (el orden de los campos de un `struct` es el de declaración), con
  `json.Encoder.SetIndent("", "  ")` y **`SetEscapeHTML(false)`** (`go doc encoding/json.Encoder.SetEscapeHTML`:
  sin ello `<`, `>` y `&` saldrían como `<`, `>` y `&`, y un cuerpo XML sería
  ilegible en una revisión). Las cabeceras son `map[string][]string` con nombres canónicos: `encoding/json` **ordena
  las claves de un mapa** (`go doc encoding/json.Marshal`: «The map keys are sorted»), luego el orden de
  las cabeceras es estable. El cuerpo va como texto en `cuerpo` si es UTF-8 válido y como base64 en
  `cuerpo_base64` si no (un PDF), nunca los dos. `grabado_en` es RFC 3339 en UTC. Lo único que varía
  entre dos grabaciones de la misma petición y respuesta es `grabado_en` y las cabeceras que el servidor
  regenera (`Date`), que es exactamente lo que FR-040 deja fuera de la comparación.
- **Escritura atómica**: fichero temporal en el mismo directorio (`os.CreateTemp`) y `os.Rename` al
  nombre final, para que un fallo a medias no deje una grabación truncada que la reproducción tomaría
  por buena.
- **Sin alterar la respuesta** (FR-038): el decorador lee el cuerpo entero, lo graba y devuelve la
  respuesta con el cuerpo sustituido por `io.NopCloser(bytes.NewReader(cuerpo))`.

**Alternativas consideradas.** *Nombre = hash de la petición* (rechazada: determinista, pero ilegible en
un diff y el spec pide «legible»). *Nombre declarado por quien llama* (rechazada en la clarificación:
muchas peticiones las origina el propio cliente). *Formato con cuerpo siempre en base64* (rechazada: un
diff de base64 no lo lee nadie). *Grabar por omisión y desactivar por bandera* (rechazada por FR-041).

---

## D13 · Reproducción: `Replay(dir)` sirve `<dir>/<nombre>.json`, empareja por método y dirección, y falla con la clase «inesperado»

**Decisión.** `dir` es el **directorio de grabaciones de una fuente** —`testdata/boe`, o
`internal/source/boe/testdata`, o cualquier otro: el cliente es agnóstico (FR-064)— y contiene los
`<nombre>.json` que la grabación dejó bajo `<raíz>/<fuente>/`. Ante una petición, el decorador
`reproducir` deriva el mismo nombre que la grabación (D12), lee el fichero y **comprueba** que
`peticion.metodo` y `peticion.url` guardados coinciden exactamente con la petición entrante (FR-039,
FR-047); las cabeceras no participan (la identificación lleva la versión y rompería todo fixture al
cambiar, como fijó la clarificación). Devuelve un `*http.Response` construido con el estado, las
cabeceras y el cuerpo grabados.

- Fichero ausente → «inesperado» nombrando método y dirección que faltaban.
- Fichero presente con otra petición → «inesperado» nombrando la buscada y la encontrada.
- Fichero ilegible o con `formato` distinto de 1 → «inesperado» nombrando el fichero.
- Nunca se abre una conexión: no hay transporte en la cadena de un cliente de reproducción (FR-046).
- `Replay` valida en la construcción que `dir` existe y es un directorio («argumentos» si no), y rechaza
  `KITLEGAL_RECORD=1` en el entorno («argumentos», FR-043) y las opciones sin sentido en reproducción
  (D1).
- **Garantías vigentes** (FR-049, lista cerrada): contexto (la petición se crea con
  `nuevaPeticionIdentificada`, que llama a `NewRequestWithContext`, y `Pedir` comprueba `ctx.Err()` antes
  de cada salto), identificación (`identificar` sigue en la cadena, y el constructor la fija aunque no
  bajara por él), método (validación de `Pedir`), redirecciones grabadas (el bucle
  de `Pedir` es el mismo; cada salto se busca como una petición más) y clasificación por estado (la de
  `Pedir`). No vigentes: `robots.txt` (el decorador no está en la cadena; una grabación de `robots.txt`
  en el directorio queda sin usar y no es error), ritmo y reintentos (sus decoradores no están; un 5xx
  grabado se clasifica como 4 en el primer y único intento).
- **Determinismo** (FR-048): ni reloj, ni aleatoriedad, ni estado entre peticiones. La misma suite dos
  veces da el mismo resultado; se comprueba con `-count=2 -shuffle=on`.

**Alternativas consideradas.** *Cargar todo el directorio al construir* (rechazada: un fichero corrupto
que ninguna petición usa rompería tests ajenos; leer bajo demanda nombra el fichero exacto que falla).
*Emparejar por hash del nombre sin comprobar el contenido* (rechazada en la clarificación: serviría la
grabación equivocada ante una colisión). *Devolver una clase de fuente ante un fixture ausente*
(rechazada por FR-063: haría pasar un test por el motivo equivocado).

---

## D14 · Concurrencia: dos niveles de exclusión y ningún estado global

**Decisión.** El cliente es seguro para varias goroutines (FR-021, FR-057): el mapa de sitios se protege
con un `sync.Mutex` de grano corto (solo para obtener o crear la entrada); cada `sitio` lleva su propio
`sync.Mutex`, tomado durante la obtención de `robots.txt` (D7), y su `*rate.Limiter`, que ya es seguro
(D8). Nada vive en variables de paquete salvo `version` (D5), que solo se lee. `go test -race` corre en
`make test` desde H0, y `TestPedirDesdeVariasGoroutines` lanza N goroutines contra el mismo servidor
local con el mismo cliente y comprueba que las peticiones llegan separadas al menos el intervalo y que
`robots.txt` se pidió una vez (SC-011, SC-003).

**Alternativas consideradas.** *`sync.RWMutex`* (rechazada: el mapa se lee y se escribe una vez por sitio;
no hay contención que justifique la complejidad). *Un cliente por goroutine* (rechazada: rompería el
ritmo por sitio, que es del cliente).

---

## D15 · Registro de eventos: `slog` a nivel `debug`, `slog.DiscardHandler` por omisión

**Decisión.** El cliente registra por el `*slog.Logger` que recibe con `ConRegistrador` (FR-035) —el
mismo que el kernel entrega al applet, montado sobre la salida de error (H1 D14)— y, sin la opción, por
`slog.New(slog.DiscardHandler)` (`go doc log/slog.DiscardHandler`: «discards all log output»), nunca por
`slog.Default()` (estado global). Todos los eventos van a nivel **`debug`**: petición emitida (método,
dirección, estado, duración, intento), decisión de `robots.txt`, espera del limitador, espera entre
intentos, grabación escrita. Con el nivel por omisión `warn` de H1 la salida de error queda limpia, y
con `--verbose` se ve todo. Las direcciones pueden contener identificadores de normas o de expedientes,
no datos de personas físicas; aun así siguen el criterio de H1 (FR-039 de `specs/002`): solo en `debug`.

---

## D16 · Dónde vive el adaptador de prueba y cómo se ejercita con el kernel en proceso

**Decisión.** El adaptador de prueba vive en `internal/httpx/adaptador_test.go`, **paquete
`httpx_test`** (caja negra): compila únicamente contra la superficie exportada, que es la demostración
literal de FR-061 —solo puede llamar a `Pedir(ctx, ejecucion, …)`; no existe nada más a lo que llamar—.
Es el tipo **`adaptadorDePrueba`** (applet de nombre `prueba`), que implementa `app.Applet`, con el tipo
`argumentosConsultar` como `app.Argumentos` (H1) del verbo `consultar <dirección>`, cuyo `Ejecutar`
construye el cliente con `ConFuente("prueba")` y `ConRegistrador(registrador)`, propaga `ctx` y
`ejecucion`, devuelve el cuerpo en `Datos` y, bajo ensayo, la descripción en `Ensayo` (D6). Los dos
nombres se fijan aquí porque son el objetivo de la comprobación mecánica de FR-062: ni
`adaptadorDePrueba` ni el literal `"consultar"` existen en ningún fichero que no sea de test
(`grep -rln --include='*.go' --exclude='*_test.go'` sobre `internal/app`, `cmd` e `internal/httpx`,
quickstart escenario 9); hoy ningún fichero de producción de H1 contiene ninguno de los dos, y el
adaptador de prueba solo puede aparecer en `_test.go`. **No se registra en ningún binario** (FR-062):
solo en un `*app.Registro` construido dentro del test, con el que `TestAdaptadorDePruebaConElKernel`
invoca `app.Main(argv, registro, &stdout, &stderr, "dev", "none", "unknown")` en proceso contra un
servidor local de pruebas, y comprueba:

- con `--dry-run`: cero accesos en el servidor, código 0, la línea `se habría pedido GET http://…` en la
  salida de error y la salida estándar vacía (SC-013, US6);
- sin `--dry-run`: la petición llega identificada y el sobre cita la dirección (US5 escenario 1);
- con `--timeout` corto contra un servidor lento: código 4 (US2 escenario 2, de extremo a extremo).

`internal/app` no importa `internal/httpx`, así que no hay ciclo; y `depguard` (lista `ejemplo`) no
interviene porque no se importa `internal/app/ejemplo`.

**Servidores locales**: `httptest.NewTestServer(t, handler)` seguido de `Start()` —red de bucle local con
`URL` real `http://127.0.0.1:<puerto>` y limpieza registrada en el test (`go doc net/http/httptest.Server`)—.
No se usa la red en memoria por omisión de `NewTestServer` porque solo la alcanza el `http.Client` que
devuelve `Server.Client()`, y el cliente de este hito lleva su propio transporte (FR-002, D11).

**Alternativas consideradas.** *Un tercer applet de ejemplo en `internal/app/ejemplo`* (rechazada: se
registraría en el binario de e2e, contra FR-062 y SC-014). *Un guion `testscript`* (rechazada en el spec:
`testscript` no puede levantar el servidor local). *Paquete `httpx` (caja blanca) para el adaptador*
(rechazada: tendría acceso a lo no exportado y la demostración de FR-061 perdería fuerza).

---

## D17 · Dónde viven los fixtures de los tests de `internal/httpx` (y qué queda abierto)

**Decisión.** Los ficheros de reproducción escritos a mano que los tests de este paquete usan viven en
`internal/httpx/testdata/reproduccion/<fuente>/*.json`, junto al paquete, como manda la convención de Go
(`go help test`: «The go tool will ignore a directory named "testdata"») y la skill
`golang-project-layout` («Use `testdata/` for fixtures», co-localizados). Son **material de test escrito
a mano contra un host ficticio** (`http://fuente.prueba`), no grabaciones de una fuente real: la
reproducción nunca conecta. El resto de fixtures los generan los propios tests en `t.TempDir()` a partir
de `struct`s, para no versionar decenas de ficheros.

**La lista es cerrada**: son exactamente los **diez** ficheros de `internal/httpx/testdata/reproduccion/prueba/`
que enumera [plan.md, «Fixtures»](./plan.md#fixtures) —nombre derivado por la regla del contrato de
grabación §2, petición grabada, respuesta y test que lo usa—, ni uno más. Cubren cada situación que los
tests de reproducción necesitan: una respuesta 200 con cuerpo de texto y cabeceras, su `HEAD` con cuerpo
vacío, una 200 con `cuerpo_base64`, una 302 con `Location` relativa cuyo destino está grabado, una 302
absoluta cuyo destino no lo está, una 302 en bucle, una 503, una 429 con `Retry-After`, el fichero de
colisión (guarda `GET http://fuente.prueba/a,b`, que el saneado confunde con `/a_b`) y una grabación de
`robots.txt` que **deniega todo** y que la reproducción debe dejar sin usar (FR-049 i). Cada nombre de la
lista figura además como fila de `TestNombreDeGrabacion`, de modo que un fichero cuyo nombre no sea el
que produce la regla hace fallar el test antes de que ningún test de reproducción lo busque. Es la lista
que la tarea `[datos]` declara como rutas y la que el guardián de diff comprueba.

**Lo que este hito no decide**: si los fixtures grabados de fuentes reales vivirán en `testdata/<fuente>/`
de la raíz o en `internal/source/<fuente>/testdata/` (`docs/PENDIENTES.md`, «Antes de H2»). El cliente es
agnóstico (FR-064) y en H2 no se graba nada real (FR-044), así que la elección no cambia ni una línea de
este hito; `docs/PENDIENTES.md` se actualiza para que el pendiente pase a «Antes de la primera tarea
`[datos]` de H4», con la recomendación intacta (junto al paquete). Es lo que el spec fija en *Fuera de
alcance*.

**Guardián del workflow**: toda tarea que toque `internal/httpx/testdata/` lleva la etiqueta `[datos]`
(`docs/WORKFLOW.md`: material nuevo bajo `internal/<pkg>/testdata/` «sigue exigiendo `[datos]` y guardián
pero no pausa»). Se anota en `plan.md` para que `tasks.md` lo herede.

---

## D18 · Controles mecánicos: qué se añade y qué se toca

**Decisión.**

- **`internal/arch_test.go`**: la subprueba R2 deja de poder pasar en vacío. Igual que `grafoDelModulo`
  ya exige que exista el dominio y el paquete de ejemplo (`require.NotEmpty` / `require.Contains`,
  `arch_test.go:331-334`), la subprueba R2 exige que `internal/httpx` esté en el grafo y que sea el
  **único** paquete del módulo que importa `net/http`. `TestDependenciasDelBinario` no cambia su lista:
  el binario distribuido no enlaza `x/time` ni `robotstxt` en H2 (D5). Se añade
  `TestElBinarioNoEnlazaHTTPX`, que comprueba precisamente eso mientras no exista un adaptador de fuente
  (SC-014): cuando H4 lo enlace, ese test se retira junto con la actualización justificada de la lista.
- **`.golangci.yml`**: solo comentarios (la lista `red` decía «el paquete todavía no existe»). No cambia
  ninguna regla ni se añade ninguna exclusión (SC-010). `noctx` y `bodyclose` ya están activos desde H0;
  H2 es el primer hito en que tienen algo que ver, y lo que vigilan está verificado: `noctx` prohíbe
  `net/http.Get/Head/Post/PostForm`, `(*Client).Get/Head/Post/PostForm`, `http.NewRequest` y
  `httptest.NewRequest` (`noctx.go`, tabla `ngFuncMessages`); `bodyclose` exige cerrar `res.Body` de todo
  `*http.Response` (README). Todo el código de `internal/httpx` usa `NewRequestWithContext` —en un solo
  sitio, `identificar.go`, del que salen ya identificadas todas las peticiones que el paquete construye
  (D3)— y `(*Client).Do`, y cierra cada cuerpo.
- **`Makefile`**: `LDFLAGS` gana la inyección de `internal/httpx.version` (D5). Ningún objetivo nuevo:
  `test` ya ejecuta `-race` sobre `./...`, que alcanza `internal/httpx`.
- **Fixtures**: los de D17. Ningún `testdata/<fuente>/` de raíz; ninguna grabación real (FR-044).
- **Tests de contrato**: los cuatro contratos de [`contracts/`](./contracts/) tienen su comprobación
  mecánica; la tabla está en `plan.md` («Controles mecánicos»).

---

## D19 · Clase «argumentos» para una dirección que no es `http`/`https` absoluta

**Decisión.** Una `Peticion.URL` que `url.Parse` rechaza, que no es absoluta o cuyo esquema no es `http`
ni `https` se rechaza en `Pedir` con la clase **«argumentos»** (código 2), sin abrir conexión. FR-063 no
la enumera —enumera el método— pero es el mismo caso: «invocación mal formada, que quien llama puede
corregir». Es el único fallo no enumerado por el spec que el cliente produce, y se declara aquí para que
no pase por decisión implícita.

**Alternativa considerada.** *Dejar que el transporte falle* (rechazada: el fallo llegaría como error de
transporte, se reintentaría dos veces y saldría con la clase 4, que significa otra cosa).

---

## D20 · `Retry-After`: segundos o fecha HTTP, en el error, y nunca esperado por el cliente

**Decisión.** Ante un 429, `Pedir` (o `robots`, para el `robots.txt`) construye `*Error` con clase 5,
`Estado: 429` y, si la cabecera existe, `Espera` calculada así: si es un entero, ese número de segundos;
si no, `http.ParseTime` (`go doc net/http.ParseTime`: los tres formatos de fecha de HTTP/1.1) menos el
instante actual, con mínimo cero; `EsperaConocida` dice si la cabecera existía. Un valor ilegible
equivale a ausente: la clase no cambia (FR-030). El cliente **no** espera ni reintenta (spec,
*Assumptions*).

---

## D21 · Aplicación de las skills `golang-*` instaladas

Se aplicaron al decidir layout, tests, lint y forma de la API (`golang-how-to` orquesta; se leyeron los
`SKILL.md` de `.agents/skills/`):

| Skill | Qué fija en este plan |
|---|---|
| `golang-design-patterns` | Opciones funcionales que devuelven error para validar en la construcción (D1); reintentos que comprueban `ctx.Err()` entre intentos y esperan con `select` sobre `ctx.Done()` (D9); patrón Decorator del roadmap (D3) |
| `golang-testing` | Tests de tabla con subtests nombrados, `t.Parallel()` salvo donde `t.Setenv` lo impide (D12), un fichero de test por fichero de código y en el mismo orden, `-race`, fixtures en `testdata/` junto al paquete (D17), caja negra para el adaptador (D16) |
| `golang-project-layout` | `internal/httpx` con ficheros por responsabilidad; `testdata/` co-localizado; ningún `pkg/` |
| `golang-lint` | «NEVER suppress security linters (gosec, bodyclose …)»: ningún `//nolint` en el hito (SC-010); por eso el jitter usa `crypto/rand` y no `math/rand/v2`, que G404 marca (D9) |
| `golang-error-handling` | Tipo de error con `Unwrap` y clasificación con `errors.As`; mensajes que nombran sitio y ruta (D4, FR-034) |
| `golang-context` / `golang-concurrency` | Contexto obligatorio como primer parámetro; exclusión por sitio; ninguna goroutine propia (D14) |
| `golang-cli` | Versión embebida por `-ldflags` (D5), coherente con H0 |
| `golang-naming` | Identificadores en español con la convención del lenguaje (`Cliente`, `Pedir`, `ConFuente`), sin guiones bajos en nombres de test |

---

## D22 · Supuestos no verificados

Todo lo anterior se ha comprobado en local. Quedan como **supuestos**, no como hechos, y con su
comprobación asignada a `tasks.md`. (La cobertura de G404 sobre `math/rand/v2` figuró aquí como supuesto
en una versión anterior de este documento; se ha verificado en `rules/rand.go`, resultó cierta, y la
decisión está tomada en D9: `crypto/rand`.)

- **S1 · `go mod tidy` solo añade `golang.org/x/time` y `github.com/temoto/robotstxt` a `go.mod`.**
  `robotstxt` requiere `stretchr/testify v1.3.0` y la selección mínima de versiones se queda con el
  v1.12.1 ya presente; `x/time` no tiene dependencias (sus `go.mod` se han leído). No se ha ejecutado
  `go mod tidy` durante la planificación para no tocar `go.mod`. Comprobación: `go mod tidy -diff` en la
  primera tarea que añada la dependencia, y `TestDependenciasDelBinario` en verde.
- **S2 · `misspell` con `locale: US` no marca ningún identificador nuevo en español.** H1 tuvo que
  ignorar `argumentos` y `descripcion`. Si un identificador de H2 dispara un falso positivo (`peticion`,
  `respuesta`…), se añade a `ignore-rules` como hizo H1, que es configuración de un diccionario y no una
  supresión de regla. La contingencia **edita `.golangci.yml`**, así que la tarea que la asume lo declara
  entre sus rutas (plan, obligación 3): es la primera que pasa `make lint` sobre español nuevo del
  paquete, la de los tipos (`peticion.go`, `sitio.go`). Una tarea posterior que tropezara con una palabra
  nueva no edita un fichero que no declara: reescribe el término en español —que no es un atajo, porque
  no toca ninguna regla ni suprime ningún hallazgo— y, si el término lo fija un contrato y no se puede
  cambiar, se anota y se detiene.
  **Resultado (implementación)**: el supuesto se cumplió a medias. T003 añadió `controles` e
  `inventario` sobre los tipos. T010 volvió a dispararlo con `legislacion` y `resolucion`, no en
  identificadores sino en dos nombres de fichero que la tabla del contrato de grabación §2 fija y que
  `TestNombreDeGrabacion` copia tal cual: las direcciones de la primera columna no se marcan porque
  `misspell` descarta las URL antes de buscar, y el nombre derivado ya no es una URL. Como el término no
  podía reescribirse, se anotó, se detuvo y se redelimitó T010 declarando `.golangci.yml` con la misma
  acotación que T003 (dos palabras bajo `misspell.ignore-rules`), en lugar de un `//nolint` o de cambiar
  el ejemplo del contrato para contentar a un diccionario inglés (`tasks.md` §Notas,
  `gates/tarea-T010.md`). Para H4: `legislacion-consolidada` es el nombre real del servicio del BOE y
  volverá a aparecer en nombres de fixtures citados desde tests, así que la entrada es durable.
- **S3 · Codecov.** Como en H1, un componente nuevo mide desde la primera propuesta posterior. H2 no
  declara componente propio (SC-015: rigen los umbrales generales), así que nada que comprobar en la
  plataforma más allá de que el estado global (≥ 70 %) y el de `internal/core` (≥ 85 %) sigan en verde.
