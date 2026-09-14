# Contrato: el puerto `core.Source`, la procedencia con fecha, el adaptador `boe` y el applet

Superficie que H4 añade al dominio, al kernel, a `internal/source/boe` y a `internal/app`. Decisiones en
[research.md](../research.md) D1-D3, D16. Lo que no aparece aquí no se exporta.

## 1. Dominio: `internal/core/source.go`

```go
package core

// Source es el puerto de una fuente pública: una consulta entra y sale un Resultado con su procedencia.
type Source interface {
    // Name es el nombre de la fuente tal como va en `fuente` del sobre. Nunca empieza por «kitlegal.».
    Name() string
    // Fetch resuelve la consulta. ec lleva --offline y --dry-run, que la fuente honra.
    // En fallo devuelve también un Resultado cuya Procedencia nombra la petición que falló (o el valor cero si no
    // llegó a construirse ninguna). Sus errores declaran su clase (schema.ConClase).
    Fetch(ctx context.Context, ec schema.Contexto, consulta Consulta) (schema.Resultado, error)
    // TTL es la vigencia con la que se guarda lo que responde esa consulta.
    TTL(consulta Consulta) time.Duration
    // Terms son los términos de uso de la fuente y cuándo los revisó una persona.
    Terms() Terminos
}

// Consulta es lo que se pide a una fuente. Cada fuente declara sus tipos.
type Consulta interface {
    Verbo() string
}

// Terminos son los términos de uso de una fuente pública.
type Terminos struct {
    URL       string    // dirección de los términos de uso
    Revisados time.Time // día de la revisión, 00:00 UTC
}
```

Importaciones permitidas: `context`, `time`, `github.com/jmorenobl/kitlegal/internal/core/schema` (R1).

## 2. Dominio y kernel: la fecha de consulta

```go
package schema

type Procedencia struct {
    Fuente string
    URL    string
    // FechaConsulta es el instante de la consulta que sostiene el contenido. El valor cero significa
    // «no la declara quien consulta» y el kernel fecha el sobre al montarlo.
    FechaConsulta time.Time
}
```

`Procedencia.Validar` no cambia. `cli.Montador` fecha el sobre así, en éxito y en fallo:

| Procedencia recibida | `fuente`, `url` del sobre | `fecha_consulta` del sobre |
|---|---|---|
| válida, `FechaConsulta` no cero | las recibidas | `FechaConsulta` |
| válida, `FechaConsulta` cero | las recibidas | reloj del montador al montar |
| inválida o cero (fallo del kernel, código 2 del applet) | `kitlegal.cli`, `kitlegal:cli` | reloj del montador al montar |

La huella sigue siendo la de `data` y no depende de la fecha (ADR 0006).

## 3. Adaptador: `internal/source/boe`

### 3.1 Constantes y construcción

```go
const NombreDeLaFuente = "boe.legislacion-consolidada"

// terminos.go, solos en su fichero: iguales a la fila de docs/SOURCES.md. Los propone la tarea del arnés de grabación
// (ritmo 1s, S6; Revisados cero con «Revisado» pendiente) y los fija una persona en la pausa del manifiesto, antes de
// grabar. Desde la tarea siguiente, ninguna línea de tasks.md contiene, ni en prosa ni dentro de una orden, la ruta
// completa de este fichero o de un directorio que lo contenga, porque el extractor de rutas del workflow declararía
// cualquiera de ellas: se nombra sin directorio, y los ficheros del paquete que una tarea cambia van por su ruta completa
// (plan, obligación 3; research.md D15).
const IntervaloEntrePeticiones = time.Second
var terminosDeUso = core.Terminos{URL: "<términos de uso>", Revisados: time.Time{}} // lo devuelve Fuente.Terms()

type Pedidor interface {
    Pedir(ctx context.Context, ec schema.Contexto, p httpx.Peticion) (httpx.Respuesta, error)
}
type CacheAbierta interface {
    core.Cache
    Close() error
}
type AperturaDeCache func(ctx context.Context, soloLectura bool) (CacheAbierta, error)

type Opcion func(*ajustes) error
func ConCliente(construir func() (Pedidor, error)) Opcion // se llama solo si hay que pedir algo
func ConCache(abrir AperturaDeCache) Opcion               // se llama tras validar la consulta
func ConRegistrador(registrador *slog.Logger) Opcion      // por omisión descarta

func Nueva(opciones ...Opcion) (*Fuente, error)
```

- `Nueva` sin `ConCliente` o sin `ConCache`, o con una función o un registrador nulos → error de clase
  «inesperado» (defecto de composición, nunca de quien invoca).
- `*Fuente` implementa `core.Source` (`var _ core.Source = (*Fuente)(nil)`).
- `Fetch` cierra la caché que abrió antes de volver, y une el error del cierre al de la invocación
  (`errors.Join`).

### 3.2 Consultas

```go
type ConsultaBuscar struct{ Texto []string }
type ConsultaIndice struct{ Norma string }
type ConsultaArticulo struct{ Norma, Bloque string }
type ConsultaArticulos struct {
    Norma   string
    Bloques []string
}
type ConsultaMetadatos struct{ Norma string }
type ConsultaAnalisis struct{ Norma string }
```

`Verbo()` devuelve `buscar`, `indice`, `articulo`, `articulos`, `metadatos` y `analisis`. Una `core.Consulta` de
otro tipo → «inesperado».

### 3.3 Datos, identificadores y errores

- Tipos de `data` con etiquetas `json` y `jsonschema`: [data-model.md](../data-model.md) §2.
- `func ValidarNorma(norma string) error`, `func ValidarBloque(bloque string) error` (error de clase
  «argumentos»), `func TipoDesdeID(id string) string` ([data-model.md](../data-model.md) §4-§5).
- `func CodigosDeAviso() []string` y `func TiposDeBloque() []string`: los enumerados del esquema, en su orden.
- `type Error struct{ URL string; Instante time.Time; Causa error }` con `Error()`, `Unwrap()` y `Clase()`
  (implementa `schema.ConClase`); tabla en [errores-y-codigos.md](./errores-y-codigos.md).

## 4. Applet: `internal/app/boe.go`

```go
type DependenciasDeBoe struct {
    // Cliente construye, por invocación y solo si hace falta pedir, el cliente HTTP de la fuente.
    Cliente func(registrador *slog.Logger) (*httpx.Cliente, error)
    // Cache son opciones adicionales para abrir la caché (vacío: la de la cuenta o KITLEGAL_CACHE_DIR).
    Cache []cache.Opcion
}

func DependenciasDeRed() DependenciasDeBoe
// Cliente: httpx.New(httpx.ConFuente(boe.NombreDeLaFuente), httpx.ConIntervalo(boe.IntervaloEntrePeticiones),
//                    httpx.ConRegistrador(registrador))

func AppletBoe(dependencias DependenciasDeBoe) Applet
```

| Verbo | Argumentos (etiquetas Kong) | Consulta | `Salida` |
|---|---|---|---|
| `buscar` | `Texto []string \`arg:"" name:"texto"\`` | `ConsultaBuscar{Texto}` | `[]boe.ResultadoDeBusqueda(nil)` |
| `indice` | `Norma string \`arg:"" name:"norma"\`` | `ConsultaIndice{Norma}` | `boe.Indice{}` |
| `articulo` | `Norma`, `Bloque string \`arg:"" name:"bloque"\`` | `ConsultaArticulo{Norma, Bloque}` | `boe.Articulo{}` |
| `articulos` | `Norma`, `Bloques []string \`arg:"" name:"bloques"\`` | `ConsultaArticulos{Norma, Bloques}` | `[]boe.Articulo(nil)` |
| `metadatos` | `Norma` | `ConsultaMetadatos{Norma}` | `boe.Metadatos{}` |
| `analisis` | `Norma` | `ConsultaAnalisis{Norma}` | `boe.Analisis{}` |

- `Nombre()` = `boe`; ningún verbo `PorOmision` (FR-001). Todos los posicionales son obligatorios (Kong).
- `Ejecutar` construye la fuente con `boe.Nueva(boe.ConCliente(…), boe.ConCache(…), boe.ConRegistrador(log))` —la
  caché con `cache.New(ctx, dependencias.Cache…, cache.ConRegistrador(log) [, cache.SoloLectura()])`— y devuelve
  `fuente.Fetch(ctx, ec, consulta)` tal cual. No lee banderas ni escribe nada.

## 5. Registro y arranque: `internal/app/registro.go`, `internal/app/main.go`

```go
func RegistroDeProduccion() (*Registro, error) // registra AppletBoe(DependenciasDeRed())

func Arrancar(argv []string, construir func() (*Registro, error), stdout, stderr io.Writer,
    version, commit, fecha string) int
```

`Arrancar` llama a `construir`; con error, lo emite por `cli.Montador` con la forma que pida `--json` del
pre-escaneo (sobre `kitlegal.cli` de clase `inesperado`, mensaje en la salida de error) y devuelve 1; sin error,
devuelve `Main(argv, registro, …)`. `Main` no cambia.

| Raíz de composición | Llama a | Registro |
|---|---|---|
| `cmd/kitlegal/main.go` | `os.Exit(app.Arrancar(os.Args, app.RegistroDeProduccion, os.Stdout, os.Stderr, …))` | `boe` con red |
| `internal/app/ejemplo/kitlegal-e2e/main.go` | `os.Exit(app.Arrancar(os.Args, registroDeE2E, …))` | `echo`, `contar` y `boe` con `httpx.Replay(filepath.Join("reproduccion", boe.NombreDeLaFuente), httpx.ConFuente(boe.NombreDeLaFuente), httpx.ConRegistrador(log))` |

`cmd/kitlegal/main_test.go` (`TestPuntoDeEntrada`) ejerce el contrato observable del binario distribuido con **su misma
composición**: `app.Arrancar(argv, app.RegistroDeProduccion, …)`, igual que `main()`, y espera `applets disponibles:
boe` donde enumera lo registrado. `&app.Registro{}` solo aparece en los tests cuyo sujeto es el registro vacío
(`TestDespachoConRegistroVacio` y la subprueba de ayuda del registro vacío).

## 6. Lo que no cambia

- `app.Applet`, `app.Verbo`, `app.Argumentos` y la firma de `app.Main` (ADR 0005).
- `schema.Resultado{Procedencia, Datos, Ensayo}` (ADR 0011); lo que crece es `Procedencia`.
- El sobre: seis claves, huella sobre `data`, sobre de fallo `{clase, mensaje}` (ADR 0006).
- El espacio reservado `kitlegal.`/`kitlegal:`: ningún literal así en `internal/source/**`
  (`TestLasFuentesNoFirmanComoKitlegal`).
