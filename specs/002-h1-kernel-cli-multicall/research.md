# Research: H1 · Kernel CLI: multicall, flags globales, exit codes, sobre de salida

**Modo**: desatendido. Ninguna decisión se consultó: todas se tomaron con la sección «Criterio de decisión
autónoma» de [`.specify/memory/constitution.md`](../../.specify/memory/constitution.md) y quedan aquí con
**decisión, alternativas consideradas y motivo**. Ninguna afecta a alcance, frontera humana, privacidad,
términos de uso de una fuente, reglas de anomalías ni a una decisión cerrada de `CLAUDE.md`, de modo que
ninguna dispara el criterio §4 («escalar en lugar de adivinar»).

## Método de verificación

La restricción del hito es explícita: **toda afirmación sobre el comportamiento de una herramienta o
dependencia externa se verifica contra su código o su documentación disponibles en local**. Lo que no se
pudo comprobar sin red se declara como **supuesto no verificado** en [D24](#d24--supuestos-no-verificados-por-falta-de-red),
no como hecho.

Fuentes locales usadas, con su ruta exacta:

| Qué | Dónde se comprobó |
|---|---|
| Go y su cadena de herramientas | `go version` → `go1.27.1 darwin/arm64`; `GOROOT` = `…/toolchain@v0.0.1-go1.27.1.darwin-arm64`; `go help packages`; `go doc` sobre la biblioteca estándar |
| `encoding/json` | `go doc encoding/json.Marshal`, `…Number`, `…Decoder.UseNumber`, `…Encoder.SetEscapeHTML` |
| `log/slog` | `go doc log/slog.HandlerOptions` |
| `testing` (salida de `TestMain`) | `$GOROOT/src/testing/testing.go:379-380` |
| `rogpeppe/go-internal` (testscript) | `$GOMODCACHE/github.com/rogpeppe/go-internal@v1.16.0/testscript/{testscript.go,exe.go,cmd.go}` |
| `invopop/jsonschema` | `$GOMODCACHE/github.com/invopop/jsonschema@v0.14.0/{schema.go,reflect.go}` |
| `santhosh-tekuri/jsonschema/v6` | `$GOMODCACHE/github.com/santhosh-tekuri/jsonschema/v6@v6.0.3/{compiler.go,format.go}` |
| `depguard` v2.2.1 | `$GOMODCACHE/github.com/!open!pee!dee!p/depguard/v2@v2.2.1/{settings.go,depguard.go,internal/utils/variables.go}` |
| `forbidigo` v2.3.1 | `$GOMODCACHE/github.com/ashanbrown/forbidigo/v2@v2.3.1/forbidigo/{patterns.go,forbidigo.go}` |
| `golangci-lint` v2.13.2 | `$GOMODCACHE/github.com/golangci/golangci-lint/v2@v2.13.2/pkg/config/{linters_settings.go,base_rule.go}`; versión fijada en `tools/golangci-lint/go.mod` |
| `alecthomas/kong` | **no está en el módulo local y la descarga quedó denegada** → D1 y D24 |

Las comprobaciones empíricas (D18, D19) se hicieron en un módulo desechable bajo `/tmp`, fuera del árbol
de trabajo; no dejaron ningún rastro en el repositorio.

---

## D1 · Kong: se resuelve el applet **antes** de construir la gramática

**Decisión.** El despacho multicall ocurre **antes** de que Kong vea un solo argumento:

1. `internal/app` determina el applet a partir de `filepath.Base(os.Args[0])` y, si ese nombre no está
   registrado, del primer argumento (D17).
2. Resuelto el applet, `internal/app` normaliza el verbo: si no se nombró ninguno y el applet declara uno
   por omisión, lo inserta ([D26](#d26--verbo-por-omisión-cómo-funciona-kitlegal-echo-hola-sin-nombrar-verbo)).
3. Con el applet y el verbo ya explícitos, `internal/cli` construye **una sola gramática** para esa
   invocación: un `struct` que embebe las ocho banderas globales (`cli.Globales`) y cuelga los verbos de
   ese applet.
4. Kong analiza esa gramática. Nunca se le describe el catálogo completo de applets.

Como consecuencia, **hay fallos que ocurren antes de que exista gramática** —applet no registrado— o antes
de que el análisis termine —bandera desconocida—. Para esos, la forma del sobre de fallo y el nivel del
registro los decide el pre-escaneo acotado de
[D25](#d25--pre-escaneo-acotado-de-argv-antes-de-la-gramática), que es la contrapartida declarada de esta
decisión.

**Motivo.** Reduce la superficie de Kong de la que depende el kernel a lo mínimo —construir un analizador,
analizar una lista de argumentos, emitir ayuda y errores por los escritores que se le den— y deja fuera
toda la maquinaria de subcomandos dinámicos, grupos y `plugins`, que es justamente la parte cuya API no se
ha podido verificar sin red (D24). Además hace trivial FR-018: las globales se declaran **una vez** en
`cli.Globales` y están presentes en toda gramática porque toda gramática las embebe; ningún applet las
declara, y no existe un camino por el que un applet pueda redefinirlas.

**Alternativas consideradas.**

- *Una gramática única con todos los applets como subcomandos, construida dinámicamente.* Rechazada: exige
  una API de construcción dinámica de comandos cuya forma exacta no puede verificarse en local, y obliga a
  que la ayuda del binario y la de cada applet salgan del mismo árbol, lo que complica FR-007 (invocación
  por enlace simbólico indistinguible de la invocación por argumento) sin aportar nada.
- *No usar Kong y analizar los argumentos a mano.* Rechazada: Kong es decisión cerrada (`CLAUDE.md`,
  constitución §V, `docs/ROADMAP.md` §3). Reabrirla está prohibido por Gobernanza.
- *Usar Kong pero con una gramática por verbo en lugar de por applet.* Rechazada: la ayuda de un applet
  (`kitlegal boe --help`) debe enumerar sus verbos, lo que exige que el applet completo sea un árbol.

**Riesgo declarado.** Los nombres concretos de la API de Kong que este diseño usa están en D24 como
supuesto no verificado, con la comprobación que la primera tarea de implementación debe ejecutar y el
plan de contingencia si no se cumplen.

---

## D2 · Reparto entre `internal/cli`, `internal/app` y `internal/core/schema`, sin ciclos

**Decisión.** Tres paquetes con una dirección de dependencia única: `cmd/kitlegal` → `internal/app` →
`internal/cli` → `internal/core/schema`, y `internal/render` → `internal/core/schema`.

| Paquete | Qué contiene | Qué NO contiene |
|---|---|---|
| `internal/core/schema` | `Sobre`, `Procedencia`, `Resultado`, `Contexto`, `Clase`, `DatosError`, canonización y huella. Dominio puro, sin I/O. | Kong, `log/slog` (**tampoco un `*slog.Logger` dentro de `Contexto`**), `os`, `io`, escritores, exit codes |
| `internal/cli` | `Globales` (las ocho banderas), pre-escaneo acotado de `argv` ([D25](#d25--pre-escaneo-acotado-de-argv-antes-de-la-gramática)), construcción y análisis de la gramática con Kong, `errors.go` (sentinelas + clasificación + traducción a código de salida), montaje del `slog` a stderr, construcción del sobre (éxito y fallo), `--describe`, y la interfaz `Presentador` que necesita para escribir | El registro de applets, `os.Exit`, la importación de `internal/render` |
| `internal/app` | `Applet`, `Verbo`, `Argumentos`, `Registro` (registro y validación), despacho multicall y normalización del verbo por omisión ([D26](#d26--verbo-por-omisión-cómo-funciona-kitlegal-echo-hola-sin-nombrar-verbo)), `Main(argv, registro, stdout, stderr, datos de versión) int` como raíz de composición | Escrituras directas a stdout, `os.Exit` |
| `internal/render` | Las dos formas de presentación, el **único** escritor de stdout y el escritor de los mensajes dirigidos a la persona en stderr ([D15](#d15--internalrender-único-escritor-de-stdout-dos-formas-y-la-tabla-mínima)) | Lógica de applet o de errores |
| `cmd/kitlegal` | `main()` → `os.Exit(app.Main(os.Args, app.RegistroDeProduccion(), os.Stdout, os.Stderr, version, commit, fecha))` | Todo lo demás |
| `internal/app/testdata/kitlegal-e2e` | El mismo `main()` con el registro de ejemplo: es la **raíz de composición del binario de e2e** y el único otro sitio del árbol donde aparecen `os.Exit`, `os.Stdout` y `os.Stderr` ([D19](#d19--dónde-viven-los-applets-de-ejemplo-y-cómo-se-compilan-lintan-y-ejecutan)) | Todo lo demás |

**Motivo.** `docs/ROADMAP.md` §2 sitúa «registro de applets (Command + Registry)» en `internal/app` y
«Kong, flags globales, mapeo error→exit code, sobre de salida» en `internal/cli`. Respetar ambas cosas
obliga a que `app` importe `cli` y a que `cli` **no** conozca el tipo `Applet`; por eso la función de
`cli` que analiza la gramática recibe el `struct` de gramática como `any` y devuelve lo analizado, sin
saber de dónde salió. Así no hay ciclo y cada paquete queda donde el roadmap lo puso.

**Escribir sin importar `render`.** `internal/cli` necesita escribir —el sobre, el sobre de fallo, el
mensaje para la persona, la descripción de `--dry-run`— pero **no** importa `internal/render`: declara la
interfaz mínima que consume y `internal/app` le inyecta el `render.Presentador` concreto (D15). Es el
«accept interfaces, return structs» del ecosistema Go y mantiene la dirección de dependencia del roadmap
sin invertirla.

**Dónde viaja el registro de eventos, y por qué no en `Contexto`.** El `*slog.Logger` **no** es un campo de
`schema.Contexto`: `internal/core/schema` es dominio puro y un registrador es un puerto de salida con I/O
detrás. El registrador se entrega al applet como **parámetro explícito** de su método de ejecución:

```go
Ejecutar(ctx context.Context, ec schema.Contexto, log *slog.Logger) (schema.Resultado, error)
```

Así `internal/core/schema` sigue importando solo `crypto/sha256`, `encoding/hex`, `encoding/json` y
`time` —y `depguard` lo hace mecánico denegando `log/slog`, `log`, `os`, `io`, `net/http` y `database/sql`
bajo `internal/core/**` (D18)—, y el applet sigue sin interpretar ninguna bandera: recibe el registrador
ya montado y con el nivel ya resuelto, que es lo que la *Key Entity* «Contexto de ejecución» del spec
exige («nivel de detalle … sin que el applet tenga que interpretarlas»).

**Alternativas consideradas.**

- *Declarar `Applet` y `Registro` en `internal/cli`.* Rechazada: contradice la asignación literal del
  roadmap §2 sin ganar nada; el `any` de la firma de análisis es un coste menor que mover el registro.
- *Un cuarto paquete `internal/kernel` que orqueste y rompa el ciclo.* Rechazada por §V (YAGNI): añade un
  paquete que ningún requisito pide para resolver un problema que el `any` ya resuelve.
- *`os.Exit` en `internal/cli` además de en `cmd/`.* Rechazada: la regla de dependencia **permite** ambos,
  pero un único punto de salida por binario es más simple y deja `cli.Main`/`app.Main` devolviendo `int`,
  que es lo que los hace testables sin subproceso (es el patrón que H0 ya estableció con
  `run(args, stdout, stderr) int`). La excepción del lint se acota a las **dos raíces de composición** y
  **no** incluye `internal/cli`, de modo que la forma estricta que el plan elige es la que el control
  vigila (D18).
- *Un campo `Log *slog.Logger` en `schema.Contexto`.* Rechazada: obligaría al dominio a importar
  `log/slog`, que es exactamente lo que §IV prohíbe («dominio puro, sin I/O»), y dejaría el Constitution
  Check afirmando una pureza que ningún control sostendría.
- *Llevar el registrador dentro del `context.Context` con una clave privada del kernel.* Rechazada: sería
  implícito —nada en la firma dice que está ahí— y obligaría a que cada applet y cada
  `internal/source/<fuente>` importase `internal/cli` para recuperarlo, que es justo el acoplamiento que
  D3 rechaza para `*kong.Context`.

---

## D3 · Contrato del applet: `Applet`, `Verbo`, `Argumentos`, `Resultado`

**Decisión.** Un applet declara **nombre, descripción, verbos y, por verbo, su gramática y el tipo de su
`data`**. Nada más:

```go
// internal/app
type Applet interface {
    Nombre() string
    Descripcion() string
    Verbos() []Verbo
}

type Verbo struct {
    Nombre      string
    Descripcion string
    Argumentos  func() Argumentos // fábrica del struct con etiquetas Kong
    Salida      any               // valor cero del tipo de `data`, solo para reflexión (--describe)
    PorOmision  bool              // como máximo uno por applet; ver D26
}

type Argumentos interface {
    // El registrador llega como parámetro explícito, no dentro de schema.Contexto:
    // el dominio no importa log/slog (D2).
    Ejecutar(ctx context.Context, ec schema.Contexto, log *slog.Logger) (schema.Resultado, error)
}

// internal/core/schema
type Resultado struct {
    Procedencia Procedencia // {Fuente, URL}
    Datos       any
}
```

**Motivo.** Es exactamente lo que exigen FR-018, FR-044, SC-010 y la *Key Entity* «Applet» del spec: el
applet no declara banderas, ni sobre, ni códigos de salida, ni forma de presentación. Kong analiza en el
`struct` que devuelve `Argumentos()`; el kernel llama a `Ejecutar`; `cli` envuelve el `Resultado` en el
sobre; `render` lo escribe. Añadir un applet es escribir un `struct` con etiquetas y un método.

La fábrica (`func() Argumentos`) en lugar de un valor: cada invocación necesita un `struct` nuevo y vacío
sobre el que Kong escriba; compartir una instancia entre invocaciones sería estado global (contrario a
«functional options … sin globals» del roadmap §2) y rompería cualquier test en paralelo.

`Salida any` es un valor cero del tipo de `data` que solo se usa por reflexión en `--describe`
(FR-046-FR-048). No se le pide al applet que escriba un esquema a mano: eso es lo que FR-048 prohíbe.

`PorOmision bool` es lo único que el contrato añade sobre la lectura literal del spec, y lo añade porque
la entrega del hito lo exige: `kitlegal echo hola` no nombra verbo alguno (D26). Es un booleano en un
`struct` que el applet ya declara, no una capacidad nueva: un applet puede no marcar ninguno —y entonces
el verbo es obligatorio— y **nunca** puede marcar dos, porque el registro lo rechaza al construirse
(FR-008).

**Alternativas consideradas.**

- *Que `Ejecutar` reciba `*kong.Context`.* Rechazada: filtraría Kong a los veinticinco applets siguientes
  y a `internal/source/<fuente>`, contra §IV (dominio y adaptadores no conocen el kernel).
- *Que el applet devuelva el sobre ya montado.* Rechazada: FR-015 y FR-045 exigen lo contrario —el sobre
  lo monta un único sitio, y el applet no puede emitir una forma distinta.
- *Que el applet reciba un `io.Writer` y escriba.* Rechazada: viola FR-040 (solo `render` escribe en
  stdout) y haría imposible SC-010.
- *Método `Esquema() *jsonschema.Schema` en el applet en lugar de `Salida any`.* Rechazada por FR-048: un
  esquema escrito a mano se desincroniza del tipo; la reflexión no puede.

---

## D4 · Forma canónica de `data` y algoritmo de la huella

**Decisión.** `hash = "sha256:" + hex(sha256(canónico(data)))`, donde `canónico` es:

1. `json.Marshal(data)` → bytes.
2. `json.Decoder` sobre esos bytes con **`UseNumber()`**, decodificando a `any`.
3. `json.Encoder` sobre el resultado con **`SetEscapeHTML(false)`**, recortando el `\n` final que el
   codificador añade.

**Motivo.** Verificado en `go doc encoding/json.Marshal`: «The map keys are sorted and used as JSON object
keys». El paso 2 convierte todo objeto en `map[string]any`, de modo que el paso 3 emite **las claves
ordenadas** con independencia del orden de declaración de los campos del `struct` o del orden de inserción
en el mapa: eso es literalmente FR-011 («con independencia del orden de las claves»). `UseNumber()`
—verificado en `go doc encoding/json.Decoder.UseNumber`— evita el paso por `float64` que destruiría la
precisión de un entero grande y cambiaría la huella de un mismo contenido según cómo se hubiera escrito.
`SetEscapeHTML(false)` —verificado en `go doc encoding/json.Encoder.SetEscapeHTML`— evita que un `<` en el
contenido produzca bytes distintos según el camino de serialización. La ausencia de espacios y saltos es
consecuencia de usar el codificador sin `SetIndent`.

Un `data` no serializable (canales, funciones, `NaN`, `+Inf`) hace fallar el paso 1 **antes** de escribir
nada en stdout: el fallo se clasifica como «inesperado» (D8) y se emite el sobre de fallo. Eso resuelve el
caso límite «contenido de `data` no serializable … no puede producir un sobre a medias ya escrito en la
salida estándar»: la huella se calcula antes de que `render` toque el descriptor.

**Alternativas consideradas.**

- *JCS (RFC 8785) completo.* Rechazada por §V: exigiría una dependencia nueva (fuera de la lista de la
  constitución) o implementar la normalización de números de la norma, y el fin que FR-011 persigue
  —misma información, misma huella— ya lo cumple el procedimiento anterior sin ninguna dependencia.
- *`json.Marshal(data)` a secas, sin el viaje de ida y vuelta.* Rechazada: para un `struct` el orden de
  las claves es el de declaración, así que dos representaciones equivalentes del mismo contenido darían
  huellas distintas, que es lo que FR-011 prohíbe.
- *Hash sobre el sobre completo en lugar de sobre `data`.* Rechazada: FR-011 dice «del contenido de
  `data`», y un hash que incluyera `fecha_consulta` cambiaría en cada consulta, inutilizando la detección
  de cambios que es la razón de ser del campo.
- *Prefijo omitido (`hash` = solo el hexadecimal).* Rechazada por FR-012 y por el ejemplo de
  `refs/kitlegal-estructura-y-ecosistema.md` §2 (`"hash": "sha256:…"`).

---

## D5 · `fecha_consulta`: RFC 3339 con desplazamiento explícito y reloj inyectado

**Decisión.** `fecha_consulta` es `time.Time` serializado por su `MarshalJSON` por omisión, que emite
RFC 3339 con desplazamiento (`2026-09-11T10:12:00+02:00`). El instante lo aporta un **reloj inyectado**
(`func() time.Time`) que `internal/app` pasa a `internal/cli`; en producción es `time.Now`, en los tests
un reloj fijo.

**Motivo.** FR-013 exige zona horaria explícita y formato legible por máquina: RFC 3339 lo es y el
esquema lo declara con `format: date-time`. El reloj inyectado es lo que hace posibles SC-003 («salida
idéntica salvo los datos que dependen del instante») y SC-005 (huellas reproducibles) sin trucos de
reescritura de la salida.

Se emite la hora **local con su desplazamiento**, no UTC forzado: el ejemplo de
`refs/kitlegal-estructura-y-ecosistema.md` §2 lleva `+02:00` y la cita de un documento legal español gana
en legibilidad con la hora local. El desplazamiento explícito la hace igual de inequívoca.

**Alternativas consideradas.** *UTC con sufijo `Z`* (rechazada: contradice el ejemplo de `refs/` sin
ganar precisión); *tiempo Unix entero* (rechazada: no es legible y pierde la zona); *`time.Now` llamado
directamente donde se monta el sobre* (rechazada: haría los tests dependientes del reloj real).

---

## D6 · Espacio de nombres reservado `kitlegal.` / `kitlegal:`

**Decisión.** Tal y como fijó el `clarify` (Q2 de `## Clarifications`) y recoge FR-016:

| Emisor | `fuente` | `url` |
|---|---|---|
| Applet sin fuente externa (`echo`) | `kitlegal.echo` | `kitlegal:applet/echo` |
| El propio kernel, cuando falla antes de llegar al applet | `kitlegal.cli` | `kitlegal:cli` |
| Adaptador de `internal/source/<fuente>` (H4 en adelante) | el nombre de la fuente | URL `http(s)` comprobable |

La prohibición de que un adaptador use ese prefijo o ese esquema **se hace mecánica en H4**, cuando exista
el primer adaptador; en H1 no hay `internal/source/` y no hay nada que comprobar. Queda anotado como
obligación en el plan.

**Verificación del formato.** `format: uri` de `santhosh-tekuri/jsonschema/v6` está implementado en
`format.go:535` (`validateURI`): analiza con `net/url.Parse` y **exige `u.IsAbs()`**. `kitlegal:cli`
analiza a `Scheme="kitlegal"`, `Opaque="cli"` y `IsAbs()` verdadero, así que **valida**; la cadena vacía
no es absoluta, así que **no valida**, que es exactamente lo que FR-017 exige («esa validación formal DEBE
rechazar un sobre cuyo `url` esté vacío o no sea un URI»). Las aserciones de `format` **no** están activas
por omisión en el borrador 2020-12: hay que llamar a `Compiler.AssertFormat()` (`compiler.go:53`). El
ayudante de validación de los tests lo llama siempre (D13).

**Alternativas consideradas.** *Dejar `fuente`/`url` vacías en applets calculados* (rechazada: contradice
FR-016 y el principio II); *usar una URL `https://ventanillalegal.es/kitlegal/applet/echo`* (rechazada:
sería una URL que no resuelve y que un consumidor trataría como cita comprobable, justo lo contrario de lo
que el espacio reservado señala); *un esquema `urn:kitlegal:…`* (rechazada: más ceremonia por el mismo
efecto, y el `clarify` ya cerró `kitlegal:`).

---

## D7 · `data` del sobre de fallo: dos claves, `clase` y `mensaje`

**Decisión.** Cuando `ok` es falso, `data` es exactamente:

```json
{ "clase": "no-encontrado", "mensaje": "no existe el bloque a99 en BOE-A-2015-10565" }
```

`additionalProperties: false`, ambas requeridas. Sin envoltorio `error`, sin `codigo`, sin traza.

**Motivo.** El `clarify` (Q3) dejó al plan «los nombres exactos de las claves dentro del `data` de error».
El spec dice «clase de error y mensaje para la persona»: dos cosas, dos claves, en español como el resto
del sobre. Añadir `codigo` duplicaría el código de salida dentro del sobre sin que ningún requisito lo
pida (criterio §2: lo no especificado no se implementa) y crearía una segunda fuente de verdad que podría
discrepar. La correspondencia que SC-014 y FR-045 exigen se comprueba entre `clase` y el **código de
salida del proceso**, no dentro del JSON.

Sin envoltorio `error` porque `data` ya está condicionado a `ok` (FR-047): cuando `ok` es falso, `data`
**es** el error; un nivel más de anidamiento no desambigua nada.

Nunca se incluye la traza ni el error envuelto completo: el mensaje para la persona va en `mensaje` y el
detalle técnico va al registro de eventos en stderr (FR-036), donde `--verbose` lo hace visible.

**Alternativas consideradas.** *`{"error": {...}}`* (rechazada, arriba); *`{"clase", "mensaje", "codigo"}`*
(rechazada, arriba); *claves en inglés* (rechazada: la convención de idioma de `CLAUDE.md` pone las claves
JSON en español).

---

## D8 · Clases de error, sentinelas y traducción exhaustiva; el inesperado sale con 1

**Decisión.** `internal/core/schema` declara el vocabulario (porque aparece en el contrato JSON y en el
esquema que emite `--describe`); `internal/cli/errors.go` declara los sentinelas y **la única** tabla de
traducción:

| `schema.Clase` | Sentinela en `internal/cli` | Código de salida |
|---|---|---|
| `argumentos` | `ErrArgumentos` | 2 |
| `no-encontrado` | `ErrNoEncontrado` | 3 |
| `fuente-no-disponible` | `ErrFuenteNoDisponible` | 4 |
| `limite-o-tos` | `ErrLimitado` | 5 |
| `identidad-humana` | `ErrIdentidadHumana` | 6 |
| `inesperado` | — (todo lo que no case con los anteriores) | **1** |

La clasificación usa `errors.Is` sobre los sentinelas, de modo que **envolver con `fmt.Errorf("%w", …)`
no cambia la clase** (FR-032). La traducción es un `switch` sobre `schema.Clase`, un tipo con constantes,
y el linter `exhaustive` —ya activo en `.golangci.yml` desde H0— **falla si aparece una clase nueva sin
rama**: eso es FR-030 («exhaustiva … no debe pasar inadvertido») hecho mecánico y no confiado a la
disciplina. Para que `exhaustive` no se conforme con un `default`, la rama por defecto vive fuera del
`switch`: la clasificación devuelve `inesperado` antes de entrar en él.

**Por qué 1 para lo inesperado.** FR-031 solo exige «distinto de los reservados y de 0». 1 es el «error
general» de la convención Unix, es el valor que la tabla de códigos de la skill `golang-cli` recoge
(`.agents/skills/golang-cli/SKILL.md`, sección *Exit Codes*), y es el único valor que un consumidor
interpretará correctamente sin documentación. `CLAUDE.md` reserva 0, 2, 3, 4, 5 y 6; 1 queda libre y no
contradice ninguna decisión cerrada.

**Alternativas consideradas.** *70 (`EX_SOFTWARE` de sysexits)* (rechazada: preciso pero opaco; ningún
documento del proyecto usa sysexits); *reutilizar el 2* (rechazada: confundiría un fallo de programación
con un error de argumentos, que es exactamente lo que FR-031 prohíbe); *definir los sentinelas en `core`*
(rechazada: los errores tipados son del kernel según §IV y el roadmap §2).

---

## D9 · `--timeout`: 30 s por omisión, sobre toda la operación, agotarlo es 4

**Decisión.** `--timeout` es una `time.Duration` de Kong con valor por omisión **`30s`**. El kernel crea
`context.WithTimeout` alrededor de la llamada a `Ejecutar` —toda la operación, no una petición suelta— y
traduce `context.DeadlineExceeded` a `ErrFuenteNoDisponible`, es decir, código 4. Un valor no positivo
(`0s`, `-1s`) o con formato inválido (`5`, `abc`) es error de argumentos, código 2.

**Motivo.** El spec deja el valor al plan y exige que exista un límite. 30 s es holgado para las fuentes
que llegan en H2-H4 (la API de Legislación Consolidada del BOE responde en cientos de milisegundos) y
suficientemente corto para que un agente no se quede colgado. FR-020 fija la traducción a 4, que es
coherente: un plazo agotado es indistinguible, desde fuera, de una fuente que no responde.

El rechazo del cero es deliberado y está en el spec (*Edge Cases*): un `--timeout 0` interpretado como
«infinito» es la clase de comportamiento accidental que deja procesos colgados para siempre.

**Alternativas consideradas.** *Sin valor por omisión (sin límite salvo si se pide)* (rechazada: FR-020
dice «DEBE limitar la duración total»; sin valor por omisión no hay límite); *10 s* (rechazada: demasiado
justo para una primera consulta sin caché a una fuente lenta); *`0` = sin límite* (rechazada, arriba).

---

## D10 · `--dry-run`: no corta, viaja en el `Contexto`, y la descripción va a stderr

**Decisión.** Implementación literal de la respuesta Q4 del `clarify` y de FR-022:

- `--dry-run` se copia en `schema.Contexto` y **el kernel llama al applet igualmente**.
- En H1 ninguna capa tiene efectos, así que **el kernel** emite la descripción de lo que se habría
  ejecutado (applet, verbo y argumentos analizados) y **no presenta ningún sobre**.
- La descripción **no es un registro de eventos**: es un mensaje dirigido a la persona. La escribe el
  presentador en el escritor de error (`Presentador.Aviso`, D15) **sin pasar por `slog`**, de modo que es
  incondicionalmente visible: no depende de `--verbose`, ni de `KITLEGAL_LOG`, ni de ningún umbral.
- La salida estándar queda vacía, también con `--json`. Código de salida 0.

**Motivo.** Está decidido en `## Clarifications`; lo que el plan añade es dónde vive y **por qué canal**.
El mensaje lo emite el kernel porque en H1 no hay ninguna capa con efectos que pueda describirse a sí
misma, y porque un applet que escribiera su propia descripción reintroduciría por la puerta de atrás la
duplicación que FR-045 y FR-018 existen para evitar.

El canal es la corrección que este hito no puede ahorrarse: FR-022 exige que la descripción sea
**siempre visible y sin depender de `KITLEGAL_LOG`**, y un registro de `slog` es por definición filtrable
por nivel —bastaría `KITLEGAL_LOG=error` para ocultarla, y con el nivel por omisión `warn` de D14 un
registro de nivel `Info` ni siquiera se emitiría—. Un requisito de visibilidad incondicional no puede
implementarse sobre un mecanismo condicional. Por eso la regla general que este hito fija es:

> **`slog` transporta el registro de eventos; el presentador transporta los mensajes dirigidos a la
> persona.** Ambos escriben en la salida de error (FR-036 y el contrato de separación de descriptores),
> pero solo el primero es filtrable por nivel. La descripción de `--dry-run`, el mensaje de un fallo, la
> lista de applets disponibles y la ayuda son mensajes para la persona y **nunca** dependen del nivel de
> registro.

FR-038 sigue cumpliéndose: pide que **el registro de eventos** sea estructurado, no que todo byte de
stderr sea un registro.

**Alternativas consideradas.** *Cortar antes de llamar al applet* (rechazada explícitamente por FR-022);
*emitir el sobre con `data` describiendo la operación* (rechazada por la respuesta Q4: una operación no
realizada no tiene nada que citar); *emitirla como registro de `slog` de nivel `Info`* (rechazada: con el
nivel por omisión `warn` de D14 no se emitiría, y con cualquier umbral configurable por `KITLEGAL_LOG`
sería ocultable, contra FR-022); *emitirla como registro de `slog` con un nivel por encima de cualquier
umbral configurable* (rechazada: obligaría a inventar un nivel artificial —`Error+4`— que se presentaría
con la etiqueta de severidad equivocada para un mensaje que no es un error, y seguiría siendo un registro
para algo que no es un evento).

---

## D11 · Precedencia entre `--describe`, `--help` y la ejecución

**Decisión.** Orden fijo, evaluado siempre en el mismo punto y con independencia del orden en que se
escriban las banderas:

1. **`--help`** (en el binario o en el applet): la ayuda se emite como texto para personas en **stdout**,
   con código 0. `--json` no la altera (respuesta Q5 del `clarify`, FR-042). Pedir la ayuda de un applet
   **suprime además la normalización del verbo por omisión** ([D26](#d26--verbo-por-omisión-cómo-funciona-kitlegal-echo-hola-sin-nombrar-verbo)),
   para que `kitlegal echo --help` enumere los verbos del applet y no describa uno solo.
2. **`--describe`**: emite el esquema JSON en stdout, con código 0, y **no ejecuta** la operación
   (FR-049). Describe siempre un verbo concreto: el nombrado, o el de omisión tras la normalización.
3. En otro caso, se ejecuta.

Si se piden `--help` y `--describe` a la vez, manda `--help`: es la salida que una persona pidió
explícitamente, y `--describe` tiene siempre una segunda oportunidad sin coste.

**Motivo.** FR-028 exige un comportamiento definido e independiente del orden; una prelación explícita y
evaluada en un solo punto es la forma más simple de garantizarlo y de poder probarlo con una tabla de
casos que incluya ambas permutaciones.

**Nota sobre `--help` y `render`.** La ayuda la escribe Kong. Para no violar FR-040 («solo `internal/render`
escribe en la salida estándar»), a Kong se le entregan los escritores de `render`, no `os.Stdout`
(D15): la ayuda sale por el mismo descriptor que todo lo demás y sigue habiendo un único paquete que
escribe en stdout.

---

## D12 · `--describe`: `invopop/jsonschema` y esquema condicional `if/then/else`

**Decisión.** `--describe` emite **un** documento con dos partes, `entrada` y `salida`, ambas generadas
por reflexión con `invopop/jsonschema`:

- `entrada`: reflexión del `struct` que devuelve `Verbo.Argumentos()`, más las ocho banderas globales de
  `cli.Globales`.
- `salida`: reflexión de `schema.Sobre` con `data` **condicionado a `ok`**, construido con
  `if` / `then` / `else`:

```json
{
  "if":   { "properties": { "ok": { "const": true } } },
  "then": { "properties": { "data": { "$ref": "#/$defs/<tipo del applet>" } } },
  "else": { "properties": { "data": { "$ref": "#/$defs/DatosError" } } }
}
```

**Verificación.** `invopop/jsonschema@v0.14.0` declara `If`, `Then` y `Else` como campos de `Schema`
(`schema.go:29-31`), emite `$schema: https://json-schema.org/draft/2020-12/schema` (`schema.go:10`), y
—dato decisivo para FR-010— pone **`additionalProperties: false` por omisión**, porque el campo
`AllowAdditionalProperties` del `Reflector` es falso por defecto y `reflect.go:468-469` lo aplica. Las
etiquetas de campo soportan `format`, `pattern`, `enum`, `minLength`, `title` y `description`
(`reflect.go:778-796, 871-873`), que es todo lo que el contrato del sobre necesita: `format=uri` en `url`,
`format=date-time` en `fecha_consulta`, `pattern` en `hash`, `enum` en `clase`.

**Motivo.** FR-047 pide literalmente el condicional y FR-048 prohíbe mantener el esquema a mano. La
reflexión sobre el `struct` de argumentos y sobre `Verbo.Salida` cumple ambas: cambiar un campo del applet
cambia su esquema sin que nadie edite nada (escenario 4 de US5).

**Alternativas consideradas.** *`oneOf` de dos sobres completos* (rechazada: duplica la descripción de las
cinco claves invariantes y produce mensajes de validación mucho peores cuando algo falla); *escribir el
esquema del sobre a mano en `schemas/`* (rechazada: FR-048, y además `schemas/` es alcance de H11);
*emitir dos documentos, uno de entrada y otro de salida* (rechazada: FR-046 dice «un esquema JSON válido
que describa la entrada y la salida», en singular, y un consumidor MCP necesita ambos juntos).

---

## D13 · Validación formal en tests: `santhosh-tekuri/jsonschema/v6` con `AssertFormat`

**Decisión.** Un ayudante de test (`internal/core/schema`, fichero `_test.go`, o un paquete de apoyo
interno a los tests) compila el esquema con `jsonschema.NewCompiler()`, llama a **`AssertFormat()`** y
valida el documento emitido. Se usa en tres sitios: el sobre de éxito, el sobre de fallo (SC-015) y la
salida de `--describe` (SC-007).

**Verificación.** `compiler.go` de la v6.0.3 expone `NewCompiler()`, `AddResource(url, doc)`, `Compile`,
`MustCompile`, `DefaultDraft` y `AssertFormat`; el comentario de `AssertFormat` (`compiler.go:47-53`)
confirma que sin esa llamada las aserciones de `format` son anotaciones y no fallan. Sin ella, un `url`
vacío pasaría la validación y FR-017 quedaría sin cumplir aunque el test estuviera «en verde»: por eso la
llamada no es opcional y el ayudante es el único camino de validación en los tests.

**Alternativas consideradas.** *Comprobar campo a campo con `require`* (rechazada por FR-017, que pide
justamente lo contrario); *la v5, también presente en la caché* (rechazada: la v6 es la línea mantenida y
es además la que `golangci-lint` ya arrastra, de modo que resuelve sin descarga nueva).

---

## D14 · Registro de eventos: `log/slog` a stderr, `--verbose` y `KITLEGAL_LOG`

**Decisión.** `internal/cli` construye un `slog.Logger` sobre un `slog.NewTextHandler` apuntando al
escritor de error, con `HandlerOptions.Level` resuelto así:

1. `KITLEGAL_LOG` (`debug`, `info`, `warn`, `error`), si está presente y es válido.
2. `--verbose` → `debug`.
3. Por omisión → `warn`.

El nivel se resuelve **antes** de analizar la gramática, para que un fallo de análisis también quede
registrado. Que `--verbose` se conozca en ese momento no es gratis y no se da por supuesto: lo aporta el
**pre-escaneo acotado de `argv`** de [D25](#d25--pre-escaneo-acotado-de-argv-antes-de-la-gramática),
que es el único mecanismo que lee banderas fuera de la gramática y está acotado por escrito. Cuando el
análisis de Kong termina bien, el valor analizado es el que manda; un test de tabla comprueba que
pre-escaneo y análisis coinciden para toda invocación bien formada.

Un valor inválido de `KITLEGAL_LOG` no aborta: se ignora y el kernel emite un aviso —mensaje para la
persona, por el presentador (D10), no un registro—, porque una variable de entorno mal escrita no es un
error de argumentos de la invocación y porque el aviso no puede quedar oculto precisamente por el valor
inválido que lo motiva.

**Verificación.** `go doc log/slog.HandlerOptions` confirma `Level Leveler` y `ReplaceAttr`. `slog` está
en la biblioteca estándar: no añade dependencia (§V).

**Por omisión `warn` y no `info`.** Con `info` por omisión, cualquier evento informativo del kernel
aparecería en stderr en cada invocación y ensuciaría la salida de una skill. `warn` deja stderr limpio
salvo cuando hay algo que decir. Ese umbral **no** deja sin cubrir ninguna garantía de visibilidad,
porque lo que tiene que verse siempre —la descripción de `--dry-run`, el mensaje de un fallo, la lista de
applets disponibles, la ayuda y el aviso de un `KITLEGAL_LOG` inválido— no viaja por `slog` sino por el
presentador (D10, D15).

**Privacidad (FR-039).** El registro nunca incluye por omisión el contenido de `data` ni los argumentos
posicionales en bruto. Registra el applet, el verbo, la duración y la clase del error. Los argumentos
solo se registran en `debug`. La descripción de `--dry-run` sí los nombra, porque son el objeto del
mensaje, y por eso es un mensaje explícito para la persona y no un registro.

**Alternativas consideradas.** *`slog.NewJSONHandler`* (rechazada para H1: stderr lo lee sobre todo una
persona, y FR-038 se cumple con el manejador de texto, que también es estructurado y filtrable; el
manejador JSON se puede añadir después sin romper nada); *`slog.SetDefault`* (rechazada: es estado global
y rompe los tests en paralelo, contra el «sin globals» del roadmap §2).

---

## D15 · `internal/render`: único escritor de stdout, dos formas, y la tabla mínima

**Decisión.** `internal/render` expone un `Presentador` construido con los dos escritores
(`render.Nuevo(stdout, stderr)`) y dos formas, elegidas por `--json`:

- **JSON**: el sobre serializado con `json.Encoder` y `SetEscapeHTML(false)`, con salto de línea final y
  **nada más** (FR-042).
- **Tabla mínima**: cuatro líneas de procedencia —`fuente`, `url`, `fecha_consulta`, `hash`— seguidas del
  contenido de `data` aplanado a pares `ruta<TAB>valor` con `text/tabwriter`, donde la ruta es la
  concatenación de claves y de índices separados por `.` (`articulos.0.titulo`). Los valores compuestos no
  se imprimen: se recorren.

Una forma nueva es un tipo nuevo que satisface la misma interfaz; ningún applet cambia (FR-044).

**Todo lo que se escribe pasa por el presentador**, que es el único que toca los dos descriptores.
`internal/cli` no importa `internal/render`: declara la interfaz que consume y `internal/app` le inyecta
la implementación (D2).

```go
// internal/cli
type Presentador interface {
    Presentar(s schema.Sobre, json bool) error // stdout: el sobre, en la forma elegida
    Texto(t string) error                      // stdout: texto para personas (version, ayuda derivada del registro)
    Aviso(t string) error                      // stderr: mensaje para la persona (fallo, lista de applets, --dry-run)
    Salida() io.Writer                         // escritores crudos, solo para entregárselos a Kong (D11)
    Error() io.Writer
}
```

Con esto, **`os.Stdout` y `os.Stderr` solo se nombran en las dos raíces de composición**
—`cmd/kitlegal/main.go` y `internal/app/testdata/kitlegal-e2e/main.go`—, que los inyectan, y ningún otro
paquete los referencia (D18 lo hace mecánico con `forbidigo`).

**El error de escritura se propaga, no se descarta.** Todas las escrituras del presentador comprueban su
error y lo devuelven; `internal/cli` lo traduce a código de salida como cualquier otro fallo
—clase `inesperado`, código **1** (D8)—, que es lo que el contrato exige para el caso de la tubería
cerrada. Dos reglas que evitan la regresión infinita:

1. Si falla la escritura **del sobre** en stdout, **no** se emite un segundo sobre de fallo por el mismo
   descriptor roto: solo se intenta el mensaje para la persona en stderr y se devuelve el código.
2. Si también falla stderr, no queda nada que hacer salvo devolver el código; nunca se entra en pánico.

Esto obliga a retirar la exclusión de `errcheck` heredada de H0
([D27](#d27--la-exclusión-de-errcheck-para-fmtfprint-se-retira)): mientras esté, el control que vigila
esta regla no existe.

**Motivo.** El formato de la tabla lo deja el spec al plan, con tres exigencias: que exista, que sea
legible y que conserve los cuatro datos de procedencia (FR-043). El aplanado por rutas es la opción que
sobrevive a un `data` arbitrariamente anidado —que es lo que traerán `boe analisis` o `placsp
licitaciones`— sin inventar un formato tabular por applet, que sería alcance nuevo.

**Alternativas consideradas.** *YAML* (rechazada: añadiría `gopkg.in/yaml.v3` al binario para una forma
que nadie pidió); *una tabla de columnas por tipo de `data`* (rechazada: exigiría que cada applet
describiera sus columnas, contra SC-010); *markdown* (explícitamente fuera de alcance, H11).

---

## D16 · `kitlegal version` de H0 sigue existiendo y sale por `render`

**Decisión.** `version` **no** se convierte en applet. Sigue siendo un verbo reservado que el kernel
atiende antes del despacho, y cambia solo en una cosa: su texto sale por `Presentador.Texto` (D15) en
lugar de por `fmt.Fprintf(stdout, …)` en `cmd/kitlegal/main.go`, y el error de escritura se propaga en
lugar de descartarse (D27). El formato de las tres líneas
(`kitlegal <version>` / `commit: …` / `fecha:  …`) y su contrato
([`contracts/cli-version.md` de H0](../001-h0-esqueleto-del-repo/contracts/cli-version.md)) no cambian.
Las variables `version`, `commit` y `fecha` siguen inyectándose con `-ldflags` en `package main`, así que
`cmd/kitlegal/main.go` se las pasa a `app.Main`.

**Motivo.** Las *Assumptions* del spec lo dicen literalmente: se mantiene, se adapta **solo** en lo que
exijan las reglas de arquitectura que H1 activa, y no se le añade sobre ni banderas. Registrarlo como
applet le daría las ocho banderas y el sobre, que es precisamente lo que el spec prohíbe.

**Alternativas consideradas.** *Registrarlo como applet* (rechazada, arriba); *dejar la escritura directa
en `cmd/`* (rechazada: las *Assumptions* dicen que la excepción de H0 para `cmd/` desaparece en H1, y
FR-040 no admite dos escritores).

---

## D17 · Precedencia del despacho, nombres no registrados y registro que se rechaza

**Decisión.**

1. Sea `n = filepath.Base(os.Args[0])`, sin sufijo `.exe`. Si `n` está registrado, **manda `n`** y **todos
   los argumentos se entregan íntegros al applet** (FR-004). Un enlace `echo -> kitlegal` invocado como
   `./echo boe` ejecuta el applet `echo` con el argumento `boe`.
2. Si `n` no está registrado —incluido el caso `n == "kitlegal"`—, el applet se toma del **primer
   argumento** y el resto se entrega al applet (FR-003, FR-005).
3. Si no hay primer argumento, o el primer argumento no está registrado: código 2, mensaje que nombra el
   applet desconocido y **lista de applets disponibles en stderr** (FR-006), y —con `--json`, conocido por
   el pre-escaneo de [D25](#d25--pre-escaneo-acotado-de-argv-antes-de-la-gramática)—
   sobre de fallo con procedencia `kitlegal.cli` (FR-045).
4. Los verbos reservados del binario (`version`) se reconocen en el paso 2, antes del registro. Un applet
   no puede llamarse `version`: lo rechaza la validación del registro.
5. Con el applet ya resuelto, y **antes** de construir la gramática, se normaliza el verbo: si el primer
   argumento restante no nombra un verbo de ese applet y el applet declara verbo por omisión, el kernel lo
   inserta a la cabeza ([D26](#d26--verbo-por-omisión-cómo-funciona-kitlegal-echo-hola-sin-nombrar-verbo)).
   Así Kong ve siempre un verbo explícito y `kitlegal echo hola` es una invocación válida.

`Registro.Registrar` **falla al construirse** (devuelve error, que la raíz de composición convierte en
pánico de arranque, nunca en error de usuario) si el nombre está duplicado, si colisiona con un verbo
reservado o si empieza por `-` (colisión con una bandera). Eso es FR-008 y el caso límite «un applet
registrado dos veces o con un nombre que colisiona con una bandera … debe detectarse al construir el
registro, no en la invocación de un usuario».

**Sobre el pánico.** FR-033 prohíbe el pánico «en rutas de usuario». Un registro mal construido no es una
ruta de usuario: es un defecto de programación que solo puede existir si alguien compiló mal el binario, y
detectarlo al arrancar es lo que impide que se manifieste como un error confuso en la invocación. Se
documenta así explícitamente para que no se lea como una violación.

**Alternativas consideradas.** *Que el primer argumento gane sobre `os.Args[0]`* (rechazada: rompería
`scripts/boe articulo …`, porque `articulo` no es un applet y el mensaje sería incomprensible);
*que un nombre de enlace desconocido sea un error* (rechazada explícitamente por FR-005 y por el caso
límite de `kitlegal-dev -> kitlegal`); *devolver error en lugar de pánico ante un registro duplicado*
(rechazada: convertiría un defecto de compilación en un fallo intermitente en tiempo de ejecución).

---

## D18 · Reparto de las reglas de arquitectura: `depguard`, `forbidigo` y `internal/arch_test.go`

**Decisión.** Cada una de las cinco reglas de `docs/ROADMAP.md` §2 se hace cumplir en **dos** capas
independientes —el lint y el test de arquitectura— salvo las que son de símbolo y no de importación, que
solo el lint puede ver:

| Regla | `depguard` | `forbidigo` | `internal/arch_test.go` |
|---|---|---|---|
| `core` no importa adaptadores ni kernel **ni I/O de la biblioteca estándar** | sí (lista `deny` por ruta) | — | sí (grafo real) |
| Solo `httpx` importa `net/http` | sí | — | sí |
| Solo `cache|store|graph` importan SQLite y `database/sql` | sí | — | sí |
| Solo `cli` y `cmd/` llaman a `os.Exit` | — | sí (`analyze-types`) | — |
| Solo `render` escribe en stdout | — | sí (`fmt.Print*`, `os.Stdout`, `os.Stderr`) | — |

**La lista `core` deniega también I/O de la biblioteca estándar.** El plan afirma que
`internal/core/schema` es dominio puro y que solo importa `crypto/sha256`, `encoding/hex`,
`encoding/json` y `time`; una afirmación así no puede quedarse en el Constitution Check. La lista `core`
deniega, además de los ocho paquetes internos, `log`, `log/slog`, `os`, `io`, `net/http` y `database/sql`.
Es lo que impide que vuelva a colarse en el dominio algo como un `*slog.Logger` dentro de `Contexto`
(D2), que `depguard` no vería si solo mirase paquetes internos. El test de arquitectura comprueba lo
mismo sobre el grafo transitivo real.

**Las excepciones de `forbidigo` son exactamente dos rutas, y ambas son raíces de composición.** Un
`package main` no puede propagar un código de salida sin `os.Exit` ni inyectar descriptores sin
nombrarlos, y el binario del e2e es un `package main` legítimo del propio hito (D19). Por tanto:

| Patrón | `msg` (marca de la regla) | Rutas exceptuadas |
|---|---|---|
| `^os\.Exit$` (`pkg: ^os$`) | `R4: …` | `cmd/**` y `internal/app/testdata/kitlegal-e2e/**` |
| `^fmt\.Print(\|f\|ln)$` | `R5: …` | **ninguna**, en todo el árbol (desaparece la acotación de H0, FR-052) |
| `^os\.Stdout$`, `^os\.Stderr$` | `R5-descriptores: …` | `cmd/**` y `internal/app/testdata/kitlegal-e2e/**` |

Dos precisiones sobre esas excepciones:

- **No incluyen `internal/cli/**`**, aunque la regla de `docs/ROADMAP.md` §2 permitiría `os.Exit` ahí. El
  plan elige la forma estricta —un único punto de salida por binario— y la excepción se acota a lo que el
  plan realmente hace, de modo que el control vigila la forma estricta y no la permisiva.
- **No incluyen `internal/app/testdata/ejemplo/**`**: los applets de ejemplo son código de applet y no
  raíz de composición, así que siguen sujetos a las tres prohibiciones. El escenario 9 del quickstart lo
  ejerce introduciendo las violaciones ahí y viendo fallar `make lint`.

**Cómo se escriben esas excepciones.** `forbidigo` no tiene filtro por ruta; lo aporta `golangci-lint` con
`linters.exclusions.rules`, que admite `linters`, `path`, `path-except` y `text`
(verificado en `golangci-lint@v2.13.2/pkg/config/base_rule.go:10-13`). El mensaje que emite `forbidigo`
incorpora el `msg` del patrón —``use of `os.Exit` forbidden because "R4: …"``, verificado en
`forbidigo@v2.3.1/forbidigo/forbidigo.go:31-37`—, así que cada excepción se escribe como `path` + `text`
con la marca de la regla (`R4:`, `R5-descriptores:`) y **no puede** desactivar por accidente la
prohibición de `fmt.Print*`, que lleva otra marca. La marca cumple además FR-053: el fallo nombra la
regla violada.

**El `TestMain` del e2e no necesita `os.Exit`.** Verificado en
`$GOROOT/src/testing/testing.go:379-380`: «If TestMain returns, the test wrapper will pass the result of
m.Run to os.Exit itself». `internal/app/e2e_test.go` construye el binario, llama a `m.Run()` y retorna,
de modo que sus `defer` de limpieza sí se ejecutan y el fichero de test no necesita ninguna excepción de
lint.

**Verificación de `depguard` v2.2.1.** Los patrones de `files` se comparan contra la **ruta absoluta del
fichero** normalizada a `/` (`depguard.go:71`: `filepath.ToSlash(pass.Fset.Position(file.Pos()).Filename)`),
así que hay que escribirlos como `**/internal/core/**`, no como `internal/core/**`. `$all` se expande a
`**/*.go` y `$test` a `**/*_test.go` (`internal/utils/variables.go:20-39`); el prefijo `!` niega
(`settings.go:61-83`). Los modos de lista son `original`, `strict` y `lax` (`settings.go:47-57`). La forma
de la configuración en `golangci-lint` v2.13.2 es `rules.<nombre>.{list-mode, files, allow, deny:[{pkg,
desc}]}` (`pkg/config/linters_settings.go:397-411`). `$gostd` se expande en `allow`/`deny` a los
directorios de `GOROOT/src` (`variables.go:41-65`), lo que permite escribir «solo la biblioteca estándar»
sin enumerarla.

**Verificación de `forbidigo` v2.3.1.** El campo `pkg` de un patrón es una expresión regular sobre la ruta
completa del paquete importado y está **ignorado salvo que el analizador esté configurado para determinar
esa información** (`forbidigo/patterns.go:23-26`), es decir, salvo que `analyze-types: true` esté activo
(el ajuste existe en `golangci-lint` v2.13.2: `ForbidigoSettings.AnalyzeTypes`,
`linters_settings.go:509-513`). Por eso H1 **activa `analyze-types`**: sin él, `^os\.Exit$` casaría por
texto y un alias de importación lo esquivaría.

**Verificación del test de arquitectura.** `docs/ROADMAP.md` §3 lo describe como «un test
`internal/arch_test.go` que recorre `go list -deps`». Comprobado empíricamente en un módulo desechable que
un directorio `internal/` que **solo** contiene `arch_test.go` (paquete `internal_test`) no rompe
`go build ./...`, `go vet ./...` ni `go test ./...`: los tres terminan en 0. El test invoca
`go list -deps -f '{{.ImportPath}} {{join .Imports " "}}'` sobre los paquetes del módulo y comprueba las
tres reglas de importación, nombrando la regla violada (FR-053).

**Por qué dos capas y no una.** FR-051 lo pide literalmente («de modo que la garantía no dependa solo de
la configuración del lint»): una configuración de lint se puede desactivar con un `//nolint` o borrando
tres líneas de YAML; un test que falla, no. El coste es duplicación consciente y está en *Complexity
Tracking*.

**Alternativas consideradas.** *Solo `depguard`* (rechazada por FR-051); *usar `golang.org/x/tools/go/packages`
en el test en lugar de `go list`* (rechazada: añadiría una dependencia fuera de la lista de §V para hacer
lo que `go list` ya hace, y el roadmap nombra `go list -deps`); *`internal/arch/arch_test.go`*
(rechazada: el roadmap dice `internal/arch_test.go` y se ha verificado que funciona).

---

## D19 · Dónde viven los applets de ejemplo, y cómo se compilan, lintan y ejecutan

**Decisión.** Bajo `internal/app/testdata/`, tal y como manda FR-009:

```
internal/app/testdata/
├── ejemplo/                 paquete `ejemplo`: los DOS applets de ejemplo (`echo` y el segundo de SC-010)
├── kitlegal-e2e/            paquete `main`: kernel real + registro de ejemplo → el binario del e2e
└── script/*.txtar           los guiones de testscript
```

**Verificado empíricamente** en un módulo desechable:

- `go build -o <tmp>/kitlegal ./internal/app/testdata/kitlegal-e2e` **compila**, y el paquete `main` bajo
  `testdata/` **puede importar** otro paquete bajo `testdata/`. El binario resultante se ejecuta.
- Los comodines **nunca** descienden a `testdata`: `go list ./...` no los lista, y —contra lo que sugiere
  el texto de `go help packages`— tampoco los listan `./internal/app/testdata/...` ni
  `<módulo>/internal/app/testdata/...` (`matched no packages` en ambos casos).
- Enumerados **explícitamente** sí se listan: `go list ./internal/app/testdata/ejemplo
  ./internal/app/testdata/kitlegal-e2e` los devuelve.

**Consecuencia operativa.** Para que los applets de ejemplo —que son la implementación de referencia que
copiará cada applet posterior— pasen exactamente los mismos controles que el resto del código, el
`Makefile` define una variable con esos dos paquetes y la **añade explícitamente** a `lint`, `fmt`,
`fmt-check` y al test de arquitectura. Dejarlos sin lintar sería el atajo que el criterio §1 prohíbe.

**Consecuencia sobre `forbidigo`.** Someter esos dos paquetes al lint completo obliga a decir qué le está
permitido a cada uno, porque no son la misma clase de código:

| Paquete | Qué es | Qué puede |
|---|---|---|
| `internal/app/testdata/ejemplo` | código de applet, la referencia que se copiará | **nada** especial: ni `os.Exit`, ni `os.Stdout`/`os.Stderr`, ni `fmt.Print*` |
| `internal/app/testdata/kitlegal-e2e` | `package main`: la **raíz de composición del binario de e2e** | `os.Exit` y `os.Stdout`/`os.Stderr`, igual que `cmd/kitlegal` y por la misma razón |

Su `main()` es, línea por línea, el mismo que el del binario distribuido salvo el registro:

```go
func main() {
    os.Exit(app.Main(os.Args, registroDeEjemplo(), os.Stdout, os.Stderr, version, commit, fecha))
}
```

**No es un agujero, y esto importa:** no hay forma de que un `package main` propague un código de salida
sin `os.Exit` —retornar de `main` siempre sale con 0— ni de que inyecte descriptores sin nombrarlos; la
excepción se acota a **ese directorio**, no a `testdata/**`; ese binario no se enlaza jamás en el
artefacto distribuido (los comodines no descienden a `testdata`, como se acaba de verificar); y el
paquete hermano `ejemplo/` queda fuera de la excepción, de modo que el listón del código que se copiará
no baja. La configuración exacta está en D18.

Sin esta excepción, `make lint` fallaría sobre código legítimo del propio hito, y la salida sería
«arreglar el control» o «renunciar al control»: las dos, atajos.

**Alternativas consideradas.**

- *Definir los applets de ejemplo en un `_test.go` de `internal/app` y usar `testscript.Main`.* Rechazada
  por FR-009, que sitúa la definición en `internal/app/testdata`. Además `testscript.Main` despacha por
  `filepath.Base(os.Args[0])` **contra su propio mapa de órdenes** (`testscript/exe.go:40-58`), de modo
  que un enlace simbólico llamado `echo` solo funcionaría si `echo` estuviera también registrado como
  orden de testscript: el enlace probaría el mapa de testscript, no el despacho del kernel (D20).
- *Registrar `echo` en el binario distribuido y marcarlo como oculto.* Rechazada: FR-009 y «Fuera de
  alcance» del spec lo prohíben.

---

## D20 · Test e2e con `testscript`: binario real construido por el test

**Decisión.** `TestMain` del paquete de e2e construye el binario con `go build` en un directorio temporal,
lo antepone a `PATH` y lo expone al guion en `$KITLEGAL_BIN`; los guiones viven en
`internal/app/testdata/script/*.txtar` y se ejecutan con `testscript.Run` y `Params{Dir, Setup,
RequireExplicitExec: true}`. El directorio temporal se borra al terminar (FR-059).

`TestMain` **retorna** en lugar de llamar a `os.Exit(m.Run())`: verificado en
`$GOROOT/src/testing/testing.go:379-380` («If TestMain returns, the test wrapper will pass the result of
m.Run to os.Exit itself»). Así el borrado del temporal ocurre de verdad —un `os.Exit` se saltaría los
`defer`— y el fichero de test no necesita ninguna excepción de `forbidigo` (D18).

**Verificación.** `testscript.Params` ofrece `Dir`, `Files`, `Setup func(*Env) error`, `Cmds`,
`Condition`, `TestWork`, `WorkdirRoot`, `UpdateScripts`, `RequireExplicitExec`, `RequireUniqueNames`,
`ContinueOnError` y `Deadline` (`testscript/testscript.go:133-203`); `Env` permite `Setenv` y `Defer`
(`testscript.go:88-128`). Las órdenes integradas incluyen `exec`, `stdout`, `stderr`, `cmp`, `exists`,
`grep`, `env`, `symlink`, `unquote`, `cp`, `mv`, `rm`, `mkdir` y `chmod` (`testscript/cmd.go:29-52`), y
`symlink` tiene la forma `symlink fichero -> destino`, con el destino **no** convertido a absoluto
(`cmd.go:462-472`), de modo que un `$KITLEGAL_BIN` absoluto funciona tal cual.

**Por qué `go build` y no `testscript.Main`.** Porque US2 y SC-003 exigen un **enlace simbólico real** a
un binario real: `testscript.Main` copia el binario de test bajo cada nombre registrado y despacha por su
propio mapa (D19), así que el enlace no probaría el despacho del kernel. Con un binario construido, el
guion hace `symlink echo -> $KITLEGAL_BIN` y `exec ./echo hola`, que es literalmente la entrega del hito.
El coste —una compilación por ejecución del paquete de e2e— es asumible y no toca la red.

**Consecuencia sobre la cobertura.** Un e2e que lanza un subproceso **no** contribuye a la cobertura de
`internal/cli`. El 90 % de SC-004 lo tienen que sostener los tests unitarios; el e2e prueba el
comportamiento observable, no la cobertura. Esto se refleja en el orden de implementación del plan.

**Alternativas consideradas.** *`testscript.Main`* (rechazada, arriba); *ejecutar el binario con
`exec.Command` sin testscript* (rechazada: FR-055 y el hito nombran `testscript`); *reutilizar el binario
de `make build`* (rechazada: obligaría a que `go test` dependiera de un artefacto construido fuera, y ese
binario no registra los applets de ejemplo).

---

## D21 · Cobertura de `internal/cli` ≥ 90 %: dónde se declara el umbral

**Decisión.** Un **componente nuevo de Codecov** en `codecov.yml`, análogo al que H0 creó para
`internal/core`:

```yaml
- component_id: internal_cli
  name: internal/cli
  paths: [internal/cli/**]
  statuses: [{type: project, target: 90%, informational: false}]
```

**Motivo.** SC-004 exige que se mida «por el mismo control de cobertura que ya usa el proyecto y sin
exclusiones añadidas». El proyecto ya usa Codecov con umbrales bloqueantes por componente; añadir uno es
literalmente el mismo control. Un objetivo `make` nuevo que analizara `coverage.out` sería un segundo
mecanismo que podría discrepar del primero.

**Alternativas consideradas.** *`go tool cover -func` en un objetivo `make cover-check`* (rechazada,
arriba: segunda fuente de verdad; además el umbral quedaría fuera de `codecov.yml`, donde ya viven los
otros dos); *no declararlo y comprobarlo a mano en la revisión* (rechazada: el criterio §1 y la sección
«Gates» de la constitución exigen que lo mecánico sea mecánico).

---

## D22 · Dependencias nuevas: cuáles, y cómo se fijan

**Decisión.** H1 añade al `go.mod` del **producto** exactamente cinco módulos, todos de la lista cerrada
de la constitución §V y de `docs/ROADMAP.md` §3:

| Módulo | Para qué | En producción | Estado local |
|---|---|---|---|
| `github.com/alecthomas/kong` | análisis de la línea de órdenes | **sí** | **ausente del módulo local**; hay que descargarlo (D24) |
| `github.com/invopop/jsonschema` | generación del esquema de `--describe` | **sí** | presente, v0.14.0 |
| `github.com/stretchr/testify` | aserciones en tests | no | presente, v1.12.1 |
| `github.com/rogpeppe/go-internal` | `testscript` para el e2e | no | presente, v1.16.0 |
| `github.com/santhosh-tekuri/jsonschema/v6` | validación formal en tests | no | presente, v6.0.3 |

Con esto **aparece por primera vez un `go.sum` en la raíz**, lo que cierra la divergencia que H0 anotó en
su *Complexity Tracking* («no existe `go.sum` en la raíz»): `go mod verify` y `go mod tidy -diff` pasan a
cubrir también el módulo del producto sin tocar el `Makefile`.

**Corrección tras la revisión final del hito.** «Exactamente cinco módulos» describe las dependencias
**directas**; no lo que el binario enlaza. `invopop/jsonschema` arrastra cuatro módulos transitivos que
acaban en `go version -m` del ejecutable —`github.com/pb33f/ordered-map/v2`, `github.com/bahlo/generic-list-go`,
`github.com/buger/jsonparser` y `go.yaml.in/yaml/v4` en versión candidata— y que este análisis no
examinó, al contrario de lo que se hizo con Kong (`gates/supuestos-kong.md`, «Dependencias que arrastra»).
Ninguna versión publicada de la biblioteca deja de traer un juego equivalente, así que no cambian la
decisión; quedan justificados en `plan.md` (*Complexity Tracking*) y en `gates/pr-h1.md`, y
`TestDependenciasDelBinario` (`internal/arch_test.go`) fija desde entonces la lista exacta de módulos que
el binario enlaza. Los módulos que solo usan los tests (`go.yaml.in/yaml/v3` por `testify`,
`golang.org/x/sys` y `golang.org/x/tools` por `testscript`, `golang.org/x/text` por el validador) no se
enlazan en el binario.

**No se fija ninguna versión en este plan.** Las versiones las fija `go get` en la tarea de
implementación y las congela `go.sum`; Dependabot ya está configurado para el módulo raíz desde H0. Fijar
aquí un número sería inventarse un dato que no se ha podido verificar (`kong`) o congelar
prematuramente los otros cuatro.

**Alternativas consideradas.** *Evitar `invopop/jsonschema` escribiendo el esquema con `map[string]any`*
(rechazada: FR-048 prohíbe mantenerlo a mano, y la librería está en la lista permitida precisamente para
esto); *usar `santhosh-tekuri/jsonschema` también en producción* (rechazada: en H1 solo se valida en
tests, y meterla en el binario sería alcance nuevo).

---

## D23 · ADR que H1 obliga a escribir

**Decisión.** Dos ADR nuevos en `docs/ADR/`, en formato MADR corto como los cuatro de H0:

- **`0005-contrato-de-applet.md`** — el contrato `Applet`/`Verbo`/`Argumentos`/`Resultado` (D3) y el
  reparto entre `cli`, `app`, `core/schema` y `render` (D2). Es la decisión de arquitectura que todos los
  hitos posteriores heredan y la que el *Definition of Done* §1.7 obliga a registrar.
- **`0006-sobre-de-salida-y-huella.md`** — la forma canónica y el algoritmo de la huella (D4), el espacio
  de nombres reservado (D6) y la forma del sobre de fallo (D7).

**Motivo.** `docs/ROADMAP.md` §1.7: «Si cambia una decisión de arquitectura: ADR nuevo». H1 no *cambia*
una decisión registrada, pero *materializa* dos que hasta ahora solo existían como frase en `CLAUDE.md`, y
son las dos que un contribuidor futuro necesitará entender antes de escribir un applet.

**No se escriben ADR para:** la elección de Kong (decisión cerrada anterior a H1), los códigos de salida
(idem) ni el multicall (ya es `0001-multicall.md`).

---

## D24 · Supuestos no verificados por falta de red

`github.com/alecthomas/kong` **no está en el módulo local** (`$GOMODCACHE/github.com/alecthomas/` contiene
`assert`, `chroma`, `go-check-sumtype` y `repr`, y `cache/download/…/kong/` no existe), y la descarga
quedó denegada en este entorno. Todo lo que sigue es **supuesto, no hecho**, y la **primera tarea de
implementación** debe comprobarlo contra la versión que `go get` traiga, antes de escribir el kernel:

| # | Supuesto sobre Kong | Cómo comprobarlo | Contingencia si no se cumple |
|---|---|---|---|
| S1 | Existe un constructor que acepta una gramática y opciones, y un método de análisis que recibe la lista de argumentos ya separada de `os.Args[0]`. | `go doc github.com/alecthomas/kong` tras `go get` | Si el análisis solo estuviera disponible sobre `os.Args`, el kernel reescribe `os.Args` antes de llamar y lo restaura; el despacho de D1 no cambia. |
| S2 | Se pueden sustituir los escritores de salida y de error que Kong usa para la ayuda y los mensajes de error. | `go doc` de las opciones del constructor | Si no se pudiera, la ayuda se genera desde el registro y se escribe por `render` sin pasar por Kong, y a Kong se le desactiva la ayuda integrada. FR-026 se cumple igual (la lista sale del registro). |
| S3 | Se puede sustituir la función que Kong llama para terminar el proceso, de modo que **ningún** camino de Kong invoque `os.Exit` por su cuenta. | prueba unitaria que fuerza `--help` y una bandera desconocida y comprueba que el proceso no termina | Si no se pudiera, el análisis se ejecuta en el mismo proceso pero con `os.Exit` interceptado por una envoltura, o se renuncia a la ayuda de Kong (ver S2). **Esto es bloqueante**: FR-035 y la regla de dependencia no admiten un `os.Exit` en `internal/cli` fuera de las dos raíces de composición (D18). |
| S4 | Las etiquetas de campo cubren ayuda, valor por omisión, nombre, forma corta y variable de entorno, y hay soporte para `time.Duration`. | `go doc` + test de la gramática de globales | Si faltara el soporte de `time.Duration`, `--timeout` se declara como cadena y se analiza con `time.ParseDuration`, devolviendo `ErrArgumentos` (código 2) ante un valor inválido. |
| S5 | Un argumento o bandera desconocido produce un error distinguible, que el kernel pueda traducir a `ErrArgumentos`. | tabla de casos de FR-027 | Si el error no fuera distinguible por tipo, se clasifica por defecto como `argumentos` **todo** error devuelto por el análisis, que es correcto por construcción: lo único que el análisis puede fallar es la línea de órdenes. |

**Lo que deliberadamente no es un supuesto.** Dos piezas del hito podrían haberse apoyado en Kong y no lo
hacen, para no alargar esta lista:

- **El verbo por omisión** (D26) se resuelve normalizando `argv` en `internal/app` antes de construir la
  gramática. Aunque Kong ofrezca comandos predeterminados, no se usan.
- **`--json`, `--verbose` y `--help` en los fallos anteriores al análisis** (D25) se leen con un
  pre-escaneo propio y acotado, no con un análisis previo de Kong.

Además, dos supuestos menores sobre herramientas, verificables en local en cuanto se implemente:

- **`golangci-lint` v2 no excluye `testdata/` por omisión.** Indicio fuerte: las salidas doradas del
  migrador v1→v2 (`pkg/commands/internal/migrate/testdata/**/empty.golden.*`) **escriben explícitamente**
  `testdata$` y `third_party$` en `linters.exclusions.paths` al convertir una configuración v1, lo que
  solo tiene sentido si en v2 esas exclusiones ya no son implícitas. `.golangci.yml` de kitlegal no las
  declara. **Comprobación**: introducir un defecto deliberado en `internal/app/testdata/ejemplo` y ver que
  `make lint` falla. Si resultara que sí se excluyen, se declara la ruta en la configuración.
- **`go mod tidy -diff` y `go mod verify` siguen en verde con `go.sum` en la raíz.** Es lo esperado, pero
  H0 nunca lo ejerció porque no había dependencias. Se comprueba con `make ci`.

---

## D25 · Pre-escaneo acotado de `argv` antes de la gramática

**Decisión.** El kernel lee tres banderas de `argv` **antes** de que exista gramática que analizar, con un
procedimiento acotado, escrito y probado, que vive en `internal/cli/preescaneo.go`:

```go
type Preliminar struct{ JSON, Verbose, Ayuda bool }
func PreEscanear(args []string) Preliminar
```

Reglas, deliberadamente estrechas:

1. Solo reconoce tres formas largas exactas: `--json`, `--verbose` y `--help`, y sus formas con valor
   booleano explícito (`--json=false`). Nada de abreviaturas ni valores separados. **Una sola forma
   corta, `-h` suelta** (corrección tras la revisión final): la ayuda integrada de Kong la anuncia en toda
   gramática como `-h, --help`, así que tiene que pedir la ayuda también donde aún no hay gramática que la
   lea —el binario y el applet—; y es la única forma corta que Kong acepta para ella, porque `-h=false`,
   `-hv` y `-H` son errores de argumentos para Kong, de modo que no reconocerlas es coincidir con él.
2. **Se detiene en el terminador `--`**: lo que va después son argumentos, no banderas.
3. No valida, no falla y no consume: un token desconocido se ignora; el pre-escaneo nunca produce un
   error ni un código de salida.
4. **No decide qué se ejecuta.** Decide exactamente tres cosas: la forma en que se presenta un fallo
   ocurrido **antes** de que Kong resuelva las globales, el nivel del registro de eventos (D14) y si se
   suprime la normalización del verbo por omisión (D26).
5. Cuando el análisis de Kong termina bien, **manda lo analizado**. El pre-escaneo es un valor provisional
   que se descarta en cuanto existe el definitivo.

**Motivo.** SC-014 exige dos casos obligatorios de sobre de fallo que ocurren **antes** de que `cli.Globales`
esté poblado: bandera desconocida y applet no registrado. En el segundo ni siquiera existe gramática,
porque D1 construye la gramática después de resolver el applet. Sin un mecanismo declarado, la pregunta
«¿cómo sabe el kernel que se pidió `--json` cuando no ha podido analizar nada?» no tiene respuesta y
FR-045, SC-014 y los escenarios 3 y 5 del quickstart no se pueden cumplir. La alternativa de no responder
—dar por hecho que «se sabe»— es justo el tipo de hueco que este plan no puede dejar.

**Qué lo mantiene honesto.** Un test de tabla comprueba, sobre un juego de invocaciones bien formadas, que
`PreEscanear` y el resultado del análisis de Kong **coinciden** en las tres banderas; si Kong y el
pre-escaneo divergieran para alguna forma sintáctica aceptada (por ejemplo `--json=true`), el test lo
señala y la forma se añade a la regla 1 o se declara no soportada. El control está en la tabla de
controles mecánicos del plan.

**Si el pre-escaneo no ve `--json`**, el fallo se presenta como el resto de la salida legible por una
persona: mensaje en la salida de error y **salida estándar vacía** (D7 y el contrato del sobre §5). No se
inventa una tercera forma.

**Alternativas consideradas.**

- *Construir primero una gramática «solo globales», analizarla y volver a analizar después con los verbos.*
  Rechazada: son dos análisis con dos oportunidades de discrepar, depende de más superficie de Kong —la
  que menos se ha podido verificar (D24)— y el primer análisis fallaría igualmente ante una bandera
  desconocida del applet, que es el caso que hay que resolver.
- *Presentar siempre el fallo en JSON.* Rechazada: rompería la salida para personas y contradiría FR-019.
- *Presentar el fallo en JSON solo cuando el análisis llegó a completarse.* Rechazada: deja sin sobre de
  fallo precisamente los dos casos que SC-014 exige.
- *Leer `--json` de una variable de entorno.* Rechazada: inventa una interfaz que ningún requisito pide
  (criterio §2).

---

## D26 · Verbo por omisión: cómo funciona `kitlegal echo hola` sin nombrar verbo

**Decisión.** `Verbo` gana un campo `PorOmision bool` (D3) y el kernel **normaliza la lista de argumentos**
antes de construir la gramática:

1. Resuelto el applet (D17), sea `resto` lo que queda de `argv`.
2. Si el primer elemento de `resto` anterior al terminador `--` nombra un verbo de ese applet, no se toca
   nada.
3. En otro caso, si el applet declara verbo por omisión **y** el pre-escaneo (D25) no ha visto `--help`,
   el kernel inserta el nombre de ese verbo a la cabeza de `resto`.
4. Si el applet no declara verbo por omisión, no se inserta nada: Kong exigirá el verbo y su ausencia
   será un error de argumentos, código 2 (FR-027).

Invariantes, comprobadas **al construir el registro** y no en la invocación de un usuario (FR-008):
`Verbos()` no vacío, nombres únicos dentro del applet y **como máximo un verbo con `PorOmision`**.

Los dos applets de ejemplo cubren las dos ramas: `echo` declara un verbo por omisión —y por eso
`kitlegal echo hola` es la entrega literal del hito—, y el segundo applet de SC-010 declara **dos** verbos
y **ninguno** por omisión, de modo que invocarlo sin verbo termina en código 2. Sin esa segunda rama, la
regla 4 no tendría prueba.

**Motivo.** La entrega del hito es literal: `kitlegal echo hola --json` (`docs/ROADMAP.md` §4 H1). Al
mismo tiempo, el contrato del applet exige `Verbos()` no vacío, y H4 entrega
`kitlegal boe articulo BOE-A-2015-10565 a21`, donde el verbo **sí** se nombra. Ambas cosas solo conviven
si existe un verbo por omisión declarado explícitamente. Dejarlo implícito («si solo hay un verbo, se
asume») sería una regla mágica que cambiaría de comportamiento el día que el applet añada un segundo
verbo: `boe` empezaría a fallar invocaciones que antes funcionaban. Declararlo es lo que lo hace estable.

**Por qué normalizar `argv` y no apoyarse en Kong.** Kong probablemente sepa marcar un comando como
predeterminado, pero eso es exactamente lo que **no se ha podido verificar** (D24): la normalización es
diez líneas en `internal/app`, es determinista, se prueba sin Kong y no añade ni un supuesto más a la
lista. Es además coherente con D1, que existe para mantener la superficie de Kong en lo mínimo. Si al
implementar se comprueba que Kong ofrece el mecanismo, **tampoco se cambia**: un supuesto menos vale más
que una línea menos.

**Ambigüedad, resuelta por escrito.** Si el primer argumento coincide con el nombre de un verbo, es el
verbo: `kitlegal echo repetir` invoca el verbo `repetir` sin mensaje, no el verbo por omisión con el
mensaje `repetir`. Quien necesite lo segundo escribe `kitlegal echo repetir repetir` o usa `--`. La regla
es una línea, es comprobable y no depende del orden de las banderas.

**Alternativas consideradas.**

- *Que el applet declare un método `VerboPorOmision() string`.* Rechazada: añade superficie a la interfaz
  `Applet` —que SC-010 quiere mínima— para decir lo mismo que un booleano en un `Verbo` que ya se declara,
  y permite nombrar un verbo que no existe.
- *Que `hola` sea el nombre del verbo de `echo`.* Rechazada: `hola` es el contenido del mensaje; convertir
  el dato en verbo haría del applet de ejemplo un mal modelo para los veinticinco siguientes.
- *Que un applet con un solo verbo lo asuma implícitamente.* Rechazada arriba: el comportamiento cambiaría
  solo, sin que nadie lo tocara, al añadir el segundo verbo.
- *Permitir invocación sin verbo y describir el applet entero en `--describe`.* Rechazada: obligaría a un
  esquema distinto según cómo se invoque, contra FR-046 («la entrada y la salida del applet invocado», en
  singular) y contra D12.

---

## D27 · La exclusión de `errcheck` para `fmt.Fprint*` se retira

**Decisión.** Desaparece de `.golangci.yml` el bloque

```yaml
errcheck:
  exclude-functions: [fmt.Fprint, fmt.Fprintf, fmt.Fprintln]
```

y **no se sustituye por una versión acotada a `cmd/`**: se retira entero. Toda escritura comprueba su
error y lo propaga (D15).

**Motivo.** H0 la introdujo con una justificación acotada y verdadera entonces: «el punto de entrada
escribe en los `io.Writer` que recibe y su contrato de códigos de salida solo admite 0 y 2
(`contracts/cli-version.md`): un fallo al escribir en stdout o stderr no tiene ni recuperación ni código
propio». En H1 las dos mitades de esa justificación dejan de valer:

1. `cmd/kitlegal/main.go` **ya no escribe**: se reduce a inyectar los descriptores y a salir con el código
   que devuelve `app.Main`. No queda punto de entrada al que la exclusión ampare.
2. El contrato de este hito exige lo contrario de «no tiene código propio»: «escribir en la salida
   estándar puede fallar (tubería cerrada): eso también se traduce a un código de salida, nunca a un
   pánico» (contrato de banderas §4, caso límite del spec). Ahora **sí** hay recuperación —propagar— y
   **sí** hay código —1, clase `inesperado`—.

Dejarla en pie sería peor que no haberla tenido nunca: `internal/render` es a partir de H1 el único
escritor del binario, de modo que la exclusión taparía exactamente el camino que el hito se compromete a
vigilar, y el plan podría afirmar que «los controles de H0 siguen sin cambios» mientras uno de ellos está
apagado justo donde hace falta. Retirar una exclusión no es una exclusión nueva: es endurecer el control
(FR-057 se cumple con más margen, no con menos).

**Coste y cómo se paga.** Cada escritura del presentador pasa a comprobar su error, incluido el `Flush`
del `tabwriter`. Es código que ya había que escribir para cumplir el contrato; la exclusión solo hacía que
su ausencia no se notara. El caso se prueba con un `io.Writer` que siempre devuelve error
(`TestEscrituraFallida`), que comprueba el código 1 y la ausencia de un segundo sobre.

**Alternativas consideradas.** *Acotarla a `cmd/**`* (rechazada: no queda ninguna escritura en `cmd/` a la
que aplicarla, así que sería una exclusión muerta que alguien reactivaría por costumbre); *dejarla y
añadir un test que compruebe la propagación* (rechazada: el test cubriría lo que se recuerde probar,
mientras que `errcheck` cubre todas las escrituras, incluidas las que se escriban en H4 y más allá);
*silenciar el error con `_ =` allí donde «no importa»* (rechazada por el criterio §1: capturar errores
para silenciarlos es exactamente lo que prohíbe).

---

## Cuestiones `NEEDS CLARIFICATION` resueltas

El spec no dejó ninguna marca `NEEDS CLARIFICATION`: las cinco preguntas abiertas se cerraron en la sesión
de `clarify` del 2026-09-11 y están en `spec.md` → `## Clarifications`. Lo que el spec delegó
explícitamente al plan queda resuelto arriba:

| Delegado por el spec | Resuelto en |
|---|---|
| Nombres de las claves del `data` de error | D7 |
| Técnica de esquema condicional | D12 |
| Valor concreto del código de salida para el fallo inesperado | D8 |
| Formato de la tabla mínima | D15 |
| Valor por omisión de `--timeout` | D9 |
| Forma canónica de `data` para la huella | D4 |
| Si hace falta un ADR nuevo | D23 |

Y lo que no delegó nadie pero el diseño no puede dejar sin respuesta, porque sin él la entrega literal del
hito o un criterio de aceptación no se sostienen:

| Pregunta que el diseño tenía abierta | Resuelto en |
|---|---|
| Por dónde viaja el registrador sin ensuciar el dominio | D2 |
| Cómo es «siempre visible» la descripción de `--dry-run` con un registro filtrable | D10, D14 |
| Cómo se conocen `--json` y `--verbose` en un fallo anterior a la gramática | D25 |
| Cómo encaja `kitlegal echo hola` con la invariante «`Verbos()` no vacío» | D26 |
| Quién vigila que un fallo de escritura en stdout se traduzca a código de salida | D15, D27 |
| Qué le está permitido al `package main` del binario de e2e | D18, D19 |
