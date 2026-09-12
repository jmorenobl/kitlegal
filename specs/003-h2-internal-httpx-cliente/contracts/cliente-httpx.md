# Contrato: el cliente de `internal/httpx`

Superficie exportada del paquete, completa. Lo que no está aquí no existe: en particular, **ningún**
`*http.Client`, `http.RoundTripper`, `*http.Request` ni `*http.Response` aparece en ninguna firma
(FR-002). Decisiones en [research.md](../research.md) D1-D3, D8-D11.

## 1. Tipos y constructores

```go
package httpx

// Cliente es el único objeto del módulo capaz de emitir una petición HTTP.
type Cliente struct{ /* privado */ }

// New construye un cliente contra la red, ya responsable sin ninguna opción:
// identificación, robots.txt, ritmo por sitio, reintentos y plazo del contexto.
// Lee KITLEGAL_RECORD una vez; con "1" activa la grabación (FR-036, FR-041).
func New(opciones ...Opcion) (*Cliente, error)

// Replay construye un cliente que responde exclusivamente desde las grabaciones
// <dir>/<nombre>.json de una fuente y no abre ninguna conexión (FR-045, FR-046).
func Replay(dir string, opciones ...Opcion) (*Cliente, error)

// Pedir es la única operación de red. Exige el contexto de cancelación y el
// contexto de ejecución del kernel; aplica la identificación internamente.
func (c *Cliente) Pedir(ctx context.Context, ejecucion schema.Contexto, p Peticion) (Respuesta, error)

type Peticion struct {
    Metodo string // "GET" o "HEAD"
    URL    string // absoluta, http o https
}

type Respuesta struct {
    Peticion  Peticion  // la pedida
    URL       string    // la final, tras las redirecciones
    Estado    int
    Cabeceras Cabeceras
    Cuerpo    []byte    // entero, ya leído
    Ensayo    bool      // true solo bajo --dry-run
}

// Descripcion es "<Metodo> <URL>" de la petición: la línea que un adaptador copia
// en schema.Resultado.Ensayo bajo --dry-run.
func (r Respuesta) Descripcion() string

type Cabeceras map[string][]string
func (c Cabeceras) Get(nombre string) string // canonicaliza el nombre; "" si no está

// Error es el único error que el paquete produce. Ver contracts/errores-y-ensayo.md.
type Error struct{ Peticion Peticion; Estado int; Espera time.Duration; EsperaConocida bool; Causa error /* + clase privada */ }
func (e *Error) Error() string
func (e *Error) Unwrap() error
func (e *Error) Clase() schema.Clase   // implementa schema.ConClase

// AgenteDeUsuario es la identificación exacta que lleva toda petición.
func AgenteDeUsuario() string          // "kitlegal/<versión> (+https://ventanillalegal.es/bot)"

const VariableGrabacion = "KITLEGAL_RECORD"
```

## 2. Opciones

Todas devuelven error de clase «argumentos» si su valor es inválido; el error se entrega desde `New` /
`Replay`, no en la opción (patrón de `golang-design-patterns`: validar en la construcción).

| Opción | Valor | Válida en | Por omisión | Inválido → clase 2 |
|---|---|---|---|---|
| `ConFuente(nombre string)` | nombre lógico de la fuente (`boe`, `placsp`…) | `New`, `Replay` | vacío | vacío o con `/`, `\`, `..` |
| `ConRaizDeGrabacion(dir string)` | directorio bajo el que se escribe `<fuente>/<nombre>.json` | solo `New` | ninguno (FR-064) | vacío; en `Replay` (FR-043) |
| `ConIntervalo(d time.Duration)` | separación mínima entre peticiones a un mismo sitio | solo `New` | `1s` | `d <= 0`; en `Replay` |
| `ConIntentos(n int)` | intentos totales por petición | solo `New` | `3` | `n < 1`; en `Replay` |
| `ConRegistrador(l *slog.Logger)` | destino de los eventos (a nivel `debug`) | `New`, `Replay` | `slog.New(slog.DiscardHandler)` | `nil` |

Repetir una opción: gana la última. No existe ninguna opción para desactivar la identificación, el
`robots.txt`, el ritmo, los reintentos, el plazo del contexto ni la verificación de certificados
(FR-002, FR-008, FR-012).

## 3. Reglas de construcción

| Situación en `New` | Resultado |
|---|---|
| `KITLEGAL_RECORD` ausente o vacía | cliente contra la red, sin grabar |
| `KITLEGAL_RECORD=1` con `ConFuente` y `ConRaizDeGrabacion`, raíz existente y `<raíz>/<fuente>` creable | cliente contra la red que graba (FR-036); crea `<raíz>/<fuente>` si no existe |
| `KITLEGAL_RECORD=1` sin `ConFuente` | clase 2: «la grabación exige declarar la fuente (ConFuente)» |
| `KITLEGAL_RECORD=1` sin `ConRaizDeGrabacion` | clase 2: «la grabación exige declarar la raíz de grabación (ConRaizDeGrabacion)» (FR-064) |
| `KITLEGAL_RECORD=1` y la raíz no existe, no es directorio o `<raíz>/<fuente>` no se puede crear | clase 2 nombrando la ruta (FR-042) |
| `KITLEGAL_RECORD` con otro valor | clase 2 nombrando el valor |

| Situación en `Replay(dir, …)` | Resultado |
|---|---|
| `dir` existe y es directorio | cliente de reproducción |
| `dir` vacío, inexistente o no directorio | clase 2 nombrando la ruta |
| `KITLEGAL_RECORD=1` en el entorno | clase 2: grabación y reproducción a la vez (FR-043) |
| `ConRaizDeGrabacion`, `ConIntervalo` o `ConIntentos` | clase 2: opción sin sentido en reproducción |

## 4. Garantías de `Pedir`

Por construcción (no hay otra operación) y comprobadas en test:

1. **Contexto obligatorio** (FR-003): `ctx` es el primer parámetro; el paquete nunca crea uno de fondo.
   Vencimiento o cancelación en cualquier punto —espera del limitador, espera entre intentos, obtención
   de `robots.txt`, lectura del cuerpo— terminan la operación en ese momento con clase 4 (FR-005).
2. **Plazo = el del contexto** (FR-004): el cliente no añade ninguno.
3. **Identificación** (FR-006 a FR-009): toda petición que sale, incluida la de `robots.txt` y cada salto
   de redirección y cada reintento, lleva `User-Agent: kitlegal/<versión> (+https://ventanillalegal.es/bot)`
   con la misma `<versión>` que imprime `kitlegal version` de ese binario.
4. **Solo GET y HEAD** (FR-010): otro método es clase 2 sin abrir conexión.
5. **Dirección válida**: absoluta, `http` o `https`, host no vacío; si no, clase 2 (research D19).
6. **`robots.txt` antes de la primera petición a un sitio** (FR-013 a FR-018): una sola obtención por
   sitio y cliente, en memoria; ruta desautorizada → clase 5 sin emitir la petición; casos de obtención
   según la lista cerrada de FR-015 (tabla en research D7).
7. **Ritmo por sitio** (FR-019 a FR-022): un token cada `intervalo`, por sitio (esquema+host+puerto),
   también para `robots.txt` y para cada salto y reintento.
8. **Reintentos** (FR-023 a FR-028): 5xx y transporte, hasta `intentos`, con espera creciente y aleatoria,
   interrumpible; nunca un 4xx; cuerpos descartados cerrados.
9. **Redirecciones** (FR-011): 301/302/303/307/308 con `Location`, tope 10, bucle detectado, cada salto
   con `robots.txt` y ritmo del sitio de destino; se entrega la respuesta final.
10. **Clasificación** (FR-029 a FR-032, FR-063): tabla completa en
    [`errores-y-ensayo.md`](./errores-y-ensayo.md) §3. Lo que no está en esa tabla se entrega como
    `Respuesta` (404 incluido).
11. **Ensayo** (FR-050, FR-051, FR-052, FR-065): con `ejecucion.DryRun`, ninguna conexión, ninguna
    consulta de `robots.txt`, ninguna espera; se comprueban método, dirección y configuración; se
    devuelve `Respuesta{Ensayo: true}` sin error.
12. **Sin pánicos** (FR-033): toda ruta de fallo del paquete devuelve `*Error`. Queda fuera de esta
    garantía, y declarado en research D9, el único suceso que el toolchain trata como irrecuperable y
    ajeno al paquete: un fallo de la fuente de entropía del sistema operativo, que termina el proceso
    con cualquier función de `crypto/rand`.
13. **Concurrencia** (FR-021, FR-057): un mismo cliente desde varias goroutines respeta el ritmo y no
    tiene carreras.
14. **Registro** (FR-035): solo por el `*slog.Logger` recibido, a nivel `debug`; nunca a la salida
    estándar.

## 5. Valores por omisión, en un sitio

| Parámetro | Valor | Dónde se fija |
|---|---|---|
| Identificación | `kitlegal/<versión> (+https://ventanillalegal.es/bot)` | `agente.go`; `<versión>` por `-ldflags` (research D5) |
| Agente para `robots.txt` | `kitlegal` | `robots.go` |
| Intervalo por sitio | 1 s | `ritmo.go` |
| Intentos | 3 | `reintentos.go` |
| Espera base / factor / techo | 500 ms / 2 / 30 s, *equal jitter*; parte aleatoria de `crypto/rand.Read` reducida con `encoding/binary`, sin rama de error (research D9) | `reintentos.go` |
| Tope de redirecciones | 10 | `cliente.go` |
| Transporte | clon de `http.DefaultTransport`; `Client.Timeout = 0`; sin redirecciones propias; sin cookies | `transporte.go` |

## 6. Ejemplo de uso (el patrón que copiará cada adaptador)

```go
func (a *argumentosConsultar) Ejecutar(
    ctx context.Context, ejecucion schema.Contexto, registrador *slog.Logger,
) (schema.Resultado, error) {
    cliente, err := httpx.New(httpx.ConFuente("boe"), httpx.ConRegistrador(registrador))
    if err != nil {
        return schema.Resultado{}, err // ya lleva su clase
    }

    respuesta, err := cliente.Pedir(ctx, ejecucion, httpx.Peticion{Metodo: "GET", URL: a.URL})
    if err != nil {
        return schema.Resultado{}, fmt.Errorf("consultando %s: %w", a.URL, err) // la clase no cambia
    }

    resultado := schema.Resultado{Procedencia: schema.Procedencia{Fuente: "boe…", URL: respuesta.URL}}
    if respuesta.Ensayo {
        resultado.Ensayo = []string{respuesta.Descripcion()}
        return resultado, nil
    }

    resultado.Datos = /* lo que el adaptador extraiga de respuesta.Cuerpo */
    return resultado, nil
}
```

No hay forma de escribir este método sin `ctx` ni sin `ejecucion`: `Pedir` los exige, y es la única
operación (FR-061).
