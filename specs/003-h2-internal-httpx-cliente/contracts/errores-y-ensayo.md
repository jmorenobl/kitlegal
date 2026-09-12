# Contrato: errores tipados, clasificación y ensayo

Cómo un fallo de `internal/httpx` llega al código de salida sin que ningún adaptador importe
`internal/cli`, y cómo `--dry-run` llega a la salida de error. Decisiones en
[research.md](../research.md) D4, D6, D19 y D20.

## 1. El puerto del dominio: `schema.ConClase`

```go
// internal/core/schema/error.go
type ConClase interface {
    error
    Clase() Clase
}
```

Un error que lo implementa declara la clase con la que el kernel debe traducirlo. `internal/core/schema`
no gana ninguna importación (R1).

## 2. `httpx.Error`

```go
type Error struct {
    Peticion       Peticion      // método y dirección implicados; vacía en fallos de configuración
    Estado         int           // estado de la fuente si hubo respuesta; 0 si no
    Espera         time.Duration // Retry-After (FR-030); 0 si no lo hubo
    EsperaConocida bool          // si Retry-After existía y era legible
    Causa          error         // error envuelto; Unwrap lo devuelve
    // clase, privada
}
func (e *Error) Error() string        // en español; nombra sitio y ruta, o la opción que falta (FR-034)
func (e *Error) Unwrap() error        // Causa
func (e *Error) Clase() schema.Clase  // una de las cuatro de §3, nunca otra
```

Uso desde un adaptador, sin importar `internal/cli`:

```go
var fallo *httpx.Error
if errors.As(err, &fallo) && fallo.EsperaConocida { /* fallo.Espera */ }
```

Envolver con `fmt.Errorf("…: %w", err)` conserva la clase (FR-031): `errors.As` recorre `Unwrap`.

## 3. Tabla cerrada situación → clase → código

Las **nueve** situaciones de red de SC-006, las **cuatro** de configuración y mecanismo de SC-016, y la
de la dirección inválida (research D19). El cliente no produce ninguna otra clase: nunca
`no-encontrado` (3) ni `identidad-humana` (6) (FR-063).

| # | Situación | Clase | Código | Requisito | `Estado` / datos |
|---|---|---|---|---|---|
| 1 | 5xx **del recurso pedido** agotados los intentos | `fuente-no-disponible` | 4 | FR-029 | estado del último intento |
| 2 | vencimiento o cancelación del contexto en cualquier punto, también obteniendo `robots.txt` o esperando el limitador | `fuente-no-disponible` | 4 | FR-029, FR-022, FR-015 | `Causa` = `ctx.Err()` |
| 3 | cadena de redirecciones **del recurso pedido** en bucle o > 10 saltos; 3xx sin `Location` | `fuente-no-disponible` | 4 | FR-011 | estado del último 3xx |
| 4 | 429 **del recurso pedido** | `limite-o-tos` | 5 | FR-030 | `429`; `Espera`/`EsperaConocida` |
| 5 | `robots.txt` deniega la ruta | `limite-o-tos` | 5 | FR-014 | 0 |
| 6 | 5xx o fallo de transporte de `robots.txt` agotados los intentos; 2xx ilegible; estado no previsto | `limite-o-tos` | 5 | FR-015 | estado del `robots.txt` (0 si transporte) |
| 7 | 429 de `robots.txt` | `limite-o-tos` | 5 | FR-030 sobre FR-015 | `429`; `Espera`/`EsperaConocida` |
| 8 | cadena de redirecciones de `robots.txt` en bucle o > 10 saltos | `limite-o-tos` | 5 | FR-015 | 0 |
| 9 | método distinto de GET/HEAD | `argumentos` | 2 | FR-010 | 0; sin conexión |
| 10 | dirección no absoluta, esquema distinto de `http`/`https`, o inanalizable | `argumentos` | 2 | research D19 | 0; sin conexión |
| 11 | grabación activa sin `ConFuente` o sin `ConRaizDeGrabacion`; `KITLEGAL_RECORD` con valor ≠ `1` | `argumentos` | 2 | FR-039, FR-064 | `Peticion` vacía; mensaje nombra la opción |
| 12 | raíz de grabación inexistente, no directorio o no escribible; `dir` de `Replay` inexistente | `argumentos` | 2 | FR-042 | mensaje nombra la ruta |
| 13 | grabación y reproducción a la vez; opción sin sentido en `Replay` | `argumentos` | 2 | FR-043 | mensaje nombra la opción |
| 14 | fallo de E/S al escribir una grabación (tras validar la raíz) | `inesperado` | 1 | FR-042 | `Causa` = error de E/S |
| 15 | colisión: el fichero guarda otra petición | `inesperado` | 1 | FR-039 | mensaje nombra las dos |
| 16 | reproducción: grabación ausente, ajena, ilegible o de otro `formato` | `inesperado` | 1 | FR-047 | mensaje nombra la petición y el fichero |

Los casos 5-8 nunca emiten la petición del recurso; los casos 6-8 dejan el sitio **denegado** en la caché
mientras viva el cliente; el caso 2 no deja nada en la caché.

## 4. La clasificación del kernel (`internal/cli`)

`Clasificar` (H1) mantiene los cinco sentinelas y su orden, y añade **una** comprobación antes de la rama
por defecto:

```
1. errors.Is contra ErrArgumentos, ErrNoEncontrado, ErrFuenteNoDisponible, ErrLimiteOTos, ErrIdentidadHumana (H1, sin cambios)
2. errors.As contra schema.ConClase: si la clase declarada está en schema.Clases(), esa clase
3. si no, ClaseInesperado (H1, sin cambios)
```

`codigoDeClase` no cambia: sigue siendo el único `switch` sin `default`, vigilado por `exhaustive`. Una
clase declarada fuera del vocabulario (imposible desde `httpx.Error`, posible desde un tipo ajeno) sale
como «inesperado», nunca como éxito ni como código reservado.

Tests que lo fijan: `TestClasificar` (`internal/cli/errors_test.go`) gana casos con un tipo de prueba
que implementa `schema.ConClase` —una clase de cada, una fuera del vocabulario, uno envuelto— y
`TestClasesDeError` (`internal/httpx/errores_test.go`) comprueba las 16 filas de §3 llamando a
`cli.Clasificar` y `cli.CodigoSalida` sobre el error real (el test sí puede importar `internal/cli`).

## 5. Ensayo (`--dry-run`)

### 5.1 En el cliente

Con `ejecucion.DryRun`, `Pedir`:

1. comprueba método, dirección y configuración (casos 9-13 de §3 siguen aplicándose: no necesitan red);
2. **no** abre conexión, **no** consulta ni evalúa `robots.txt`, **no** espera ni consume ritmo, **no**
   sigue redirecciones, **no** reintenta (FR-050);
3. devuelve `Respuesta{Peticion: p, URL: p.URL, Ensayo: true}` y `nil` (FR-052, FR-065).

### 5.2 En el dominio

`schema.Resultado` gana:

```go
type Resultado struct {
    Procedencia Procedencia
    Datos       any
    // Ensayo describe, una línea por petición, lo que cada capa con efectos habría hecho.
    // Solo se rellena bajo --dry-run; el kernel lo presenta en la salida de error.
    Ensayo []string
}
```

El adaptador copia `respuesta.Descripcion()` («`GET http://…`») de cada petición de ensayo.

### 5.3 En el kernel

`internal/app` presenta, bajo `--dry-run` y por el presentador (nunca por `slog`), la línea de H1 y a
continuación una por elemento de `Ensayo`:

```
--dry-run: no se ha ejecutado nada; se habría ejecutado el applet "prueba", el verbo "consultar", con los argumentos ["consultar" "http://127.0.0.1:53211/norma" "--dry-run"]
--dry-run: se habría pedido GET http://127.0.0.1:53211/norma
```

(La primera línea es la de H1, tal como la imprime hoy el binario de e2e con `echo hola --dry-run`; la
segunda es la que H2 añade por cada elemento de `Ensayo`.)

Salida estándar vacía; código de salida 0 (FR-065). Si el applet devolvió error, se presenta lo que haya
en `Ensayo` y manda el error, como en H1.

Es una ampliación del contrato del applet (ADR 0005) y se registra en
`docs/ADR/0011-ensayo-por-capas.md` (research D6).
