# Data model: H2 · `internal/httpx`

Entidades del hito, sus campos, invariantes y transiciones. Todo vive en `internal/httpx` salvo lo que
se indica; las decisiones que lo motivan están en [research.md](./research.md) y la API exacta en
[`contracts/cliente-httpx.md`](./contracts/cliente-httpx.md).

## 1. Cliente responsable (`Cliente`)

Lo que devuelven `New` y `Replay`. Único objeto del módulo capaz de emitir una petición (FR-002).

| Campo (privado) | Tipo | Invariante |
|---|---|---|
| `cadena` | `http.RoundTripper` | Compuesta una vez en el constructor (D3); nunca cambia |
| `cliente` | `*http.Client` | `Timeout: 0`; `CheckRedirect` devuelve `http.ErrUseLastResponse`; `Jar: nil`. En reproducción no existe (`nil`): no hay transporte |
| `modo` | enum privado: `red`, `grabacion`, `reproduccion` | `grabacion` ⇔ `KITLEGAL_RECORD=1` en `New`; `reproduccion` ⇔ `Replay`; incompatibles (FR-043) |
| `fuente` | `string` | Obligatoria en `grabacion` (FR-039); libre en los demás modos |
| `raiz` | `string` | Obligatoria en `grabacion` (FR-064); debe existir y ser directorio; `<raiz>/<fuente>` debe poder crearse |
| `directorio` | `string` | Solo en `reproduccion`: el directorio de grabaciones de una fuente; debe existir |
| `intervalo` | `time.Duration` | `> 0`; por omisión 1 s (D8) |
| `intentos` | `int` | `≥ 1`; por omisión 3 (D9) |
| `registrador` | `*slog.Logger` | Nunca `nil`: `slog.DiscardHandler` por omisión (D15) |
| `sitios` | `map[string]*sitio` + `sync.Mutex` | Clave = origen (§4); una entrada por sitio; vive lo que el cliente |
| `dormir` | `func(context.Context, time.Duration) error` | Espera interrumpible entre intentos; los tests del paquete la sustituyen (D9) |

**Estado**: el cliente no tiene ciclo de vida más allá de la invocación; no hay `Close`. Es seguro para
uso concurrente (FR-021, D14).

## 2. Petición (`Peticion`)

Lo que quien llama pide. Es también lo que la grabación guarda y lo que la respuesta devuelve.

| Campo | Tipo | Validación (en `Pedir`, antes de abrir nada) | Clase si falla |
|---|---|---|---|
| `Metodo` | `string` | Exactamente `GET` o `HEAD` (FR-010) | `argumentos` (2) |
| `URL` | `string` | `url.Parse` sin error, absoluta, esquema `http` o `https`, host no vacío (D19) | `argumentos` (2) |

## 3. Respuesta (`Respuesta`)

Lo que `Pedir` devuelve cuando no hay error de clase. Tipo del paquete, nunca `*http.Response` (FR-002).

| Campo | Tipo | Significado | En ensayo (FR-065) |
|---|---|---|---|
| `Peticion` | `Peticion` | La petición tal como se pidió | la misma |
| `URL` | `string` | Dirección **final** de la cadena de redirecciones (FR-011) | = `Peticion.URL` |
| `Estado` | `int` | Estado HTTP de la respuesta final | `0` |
| `Cabeceras` | `Cabeceras` | Cabeceras de la respuesta final, nombres canónicos | vacías |
| `Cuerpo` | `[]byte` | Cuerpo entero, ya leído (D2); vacío en HEAD | `nil` |
| `Ensayo` | `bool` | Verdadero solo bajo `--dry-run` | `true` |

Método `Descripcion() string`: «`<Metodo> <URL>`» de la petición; es la línea que el adaptador copia en
`schema.Resultado.Ensayo` (D6).

**Invariantes**: `Estado` nunca es 3xx (se sigue, o es clase 4 si no se puede seguir), ni 429 (clase 5),
ni 5xx (clase 4): lo entregable es 2xx y 4xx distinto de 429, y cualquier otro estado que no encaje en
las clases de FR-029 y FR-030 (FR-032). `Ensayo` ⇒ `Estado == 0 && Cuerpo == nil`.

### `Cabeceras`

`map[string][]string` con nombres canónicos (`http.CanonicalHeaderKey`) y método `Get(nombre) string`
que canonicaliza el nombre y devuelve el primer valor o `""`. Existe para que un adaptador no importe
`net/http` (R2).

## 4. Sitio (`sitio`, privado)

El **origen** de una dirección: la unidad del `robots.txt` (RFC 9309) y del ritmo (FR-013, FR-015,
FR-019).

| Campo | Tipo | Invariante |
|---|---|---|
| clave | `string` | `<esquema>://<host en minúsculas>:<puerto>`; puerto explícito, o `80`/`443` por omisión del esquema |
| `limitador` | `*rate.Limiter` | `rate.NewLimiter(rate.Every(intervalo), 1)`; creado al crear el sitio |
| `mu` | `sync.Mutex` | Tomado durante la obtención de `robots.txt`; N goroutines ⇒ 1 petición (SC-003) |
| `reglas` | `*reglasDelSitio` | `nil` hasta la primera evaluación; después inmutable (FR-015, FR-018) |

Dos direcciones son del mismo sitio si y solo si su clave coincide: `http://fuente.prueba` ≠
`https://fuente.prueba`; `http://fuente.prueba` ≠ `http://otra.prueba`; `http://127.0.0.1:53211` ≠
`http://127.0.0.1:53212` (SC-004); `http://fuente.prueba` = `http://fuente.prueba:80` y
`http://Fuente.Prueba/ruta?q=1` = `http://fuente.prueba/`. La regla la fija `TestClaveDeSitio`
(`sitio_test.go`, tabla en plan.md «Inventario de tests»), sin servidor y solo con los hosts ficticios
que admite la comprobación «solo direcciones locales» del quickstart (`fuente.prueba`, `otra.prueba`).

## 5. Reglas del sitio (`reglasDelSitio`, privado)

Resultado de obtener e interpretar `robots.txt` **o** de no poder hacerlo (FR-015, lista cerrada).

| Campo | Tipo | Significado |
|---|---|---|
| `datos` | `*robotstxt.RobotsData` | Reglas interpretadas (2xx interpretable) o `allowAll` (4xx ≠ 429, cuerpo vacío) |
| `denegado` | `*Error` | No `nil` cuando el sitio quedó **denegado entero**: 429 (clase 5, con `Retry-After` si lo hubo), 5xx o transporte agotados los reintentos (5), contenido ilegible (5), cadena de redirecciones del `robots.txt` en bucle o excedida (5), estado no previsto (5) |

**Evaluación** de una ruta: si `denegado != nil` → ese error; si no, `datos.TestAgent(url.RequestURI(),
"kitlegal")` → permitir o error de clase 5 «ruta desautorizada por robots.txt» (FR-014). Solo `Allow` /
`Disallow`; `Crawl-delay` no se lee.

**Transiciones**: `sin evaluar` → (obtención) → `permitido con reglas` | `denegado`; nunca vuelve atrás
dentro de la vida del cliente. El vencimiento del contexto **durante** la obtención no crea entrada (la
operación termina con clase 4 y la siguiente petición volverá a intentarlo).

## 6. Política de ritmo

| Parámetro | Valor | Ajuste |
|---|---|---|
| Intervalo mínimo entre peticiones a un mismo sitio | 1 s | `ConIntervalo(d)`, `d > 0`, uno por cliente (FR-020) |
| Ráfaga | 1 token | fija |
| Espera | `Limiter.Wait(ctx)` | el vencimiento del contexto durante la espera es clase 4 (FR-022) |

## 7. Política de reintentos

| Parámetro | Valor | Ajuste |
|---|---|---|
| Intentos (petición inicial incluida) | 3 | `ConIntentos(n)`, `n ≥ 1` (FR-024) |
| Se reintenta | 5xx; error de transporte con `ctx.Err() == nil` | fijo (FR-025) |
| No se reintenta | 4xx (429 incluido); cualquier 2xx/3xx; cualquier error con `ctx.Err() != nil` | fijo (FR-025, FR-026) |
| Espera antes del reintento *n* (n ≥ 1) | `b/2 + aleatorio[0, b/2)`, con `b = min(500 ms · 2^(n−1), 30 s)` (*equal jitter*) | fijo (D9) |
| Fuente de la parte aleatoria | `crypto/rand.Read` sobre ocho bytes, `binary.BigEndian.Uint64 >> 1` y módulo `b/2` sobre `time.Duration`; `Read` no devuelve nunca error, así que no hay rama de contingencia (un fallo de entropía del sistema termina el proceso con cualquier función de `crypto/rand` en go1.27.1, D9); nunca `math/rand`, `math/rand/v2` (G404) ni `math/big` | fijo (D9) |
| Cuerpo del intento descartado | leído y cerrado antes del siguiente intento | fijo (FR-027) |

## 8. Cadena de redirecciones (estado del bucle de `Pedir`)

| Parámetro | Valor |
|---|---|
| Se sigue | 301, 302, 303, 307, 308 con `Location` |
| Método en el salto | el de la petición original (GET→GET, HEAD→HEAD) |
| Destino | `URL.ResolveReference(Location)` |
| Tope | 10 saltos (D10) |
| Bucle | dirección ya visitada en la cadena → termina |
| Clase al exceder o bucle | recurso pedido: 4 (FR-011); `robots.txt`: 5 (FR-015) |
| 3xx sin `Location`, 300, 304, 305 | recurso pedido: 4; `robots.txt`: 5 |

Cada salto es una petición completa por la cadena de decoradores: identificación, `robots.txt` y ritmo
del sitio de destino.

## 9. Clase de fallo (`Error`)

El único tipo de error que el paquete produce. Implementa `error`, `Unwrap() error` y
`schema.ConClase` (D4).

| Campo | Tipo | Significado |
|---|---|---|
| `clase` (privado; método `Clase()`) | `schema.Clase` | Una de `argumentos`, `fuente-no-disponible`, `limite-o-tos`, `inesperado`; **nunca** `no-encontrado` ni `identidad-humana` (FR-063) |
| `Peticion` | `Peticion` | La petición implicada (vacía en fallos de configuración sin petición, FR-034) |
| `Estado` | `int` | Estado de la fuente si hubo respuesta (`429`, `503`…); `0` si no |
| `Espera` | `time.Duration` | Valor de `Retry-After` (FR-030); `0` si no lo hubo |
| `EsperaConocida` | `bool` | Si la cabecera `Retry-After` existía y se pudo leer |
| `Causa` | `error` | Error envuelto (transporte, contexto, E/S, JSON…); `Unwrap` lo devuelve |

`Error()` está en español y nombra sitio y ruta cuando hay petición, o la opción que falta o sobra cuando
no (FR-034). Tabla completa situación → clase → código en
[`contracts/errores-y-ensayo.md`](./contracts/errores-y-ensayo.md) §3.

### Ampliaciones del dominio y del kernel (H1)

| Dónde | Qué | Por qué |
|---|---|---|
| `internal/core/schema` | `type ConClase interface { error; Clase() Clase }` | Puerto por el que un adaptador declara su clase sin importar `internal/cli` (D4) |
| `internal/core/schema` | `Resultado.Ensayo []string` | Descripción de lo que cada capa con efectos habría hecho bajo `--dry-run`; el kernel la presenta (D6) |
| `internal/cli` | `Clasificar`: tras los cinco sentinelas, `errors.As` sobre `schema.ConClase` con clase validada contra `schema.Clases()` | Reconocer `*httpx.Error` (D4) |
| `internal/app` | `conDescripcion` presenta cada línea de `Resultado.Ensayo` tras la descripción de H1 | FR-051, FR-065 (D6) |

## 10. Grabación (fichero `<raíz>/<fuente>/<nombre>.json`)

Estructura serializada con claves en este orden ([`contracts/formato-de-grabacion.md`](./contracts/formato-de-grabacion.md)):

| Clave | Tipo | Contenido |
|---|---|---|
| `formato` | entero | `1` |
| `grabado_en` | cadena RFC 3339 UTC | Instante de la grabación (fuera de la comparación de FR-040) |
| `peticion.metodo` | cadena | `GET` / `HEAD` |
| `peticion.url` | cadena | Dirección completa tal como se emitió |
| `peticion.cabeceras` | objeto `nombre → [valores]` | Cabeceras enviadas, la identificación entre ellas; claves ordenadas |
| `respuesta.estado` | entero | Estado HTTP |
| `respuesta.cabeceras` | objeto `nombre → [valores]` | Cabeceras recibidas, `Date` incluida (fuera de la comparación) |
| `respuesta.cuerpo` | cadena | Cuerpo si es UTF-8 válido |
| `respuesta.cuerpo_base64` | cadena | Cuerpo en base64 si no lo es; excluyente con `cuerpo` |

**Nombre**: derivado de método y dirección (D12; contrato §2). **Colisión**: fichero existente con otra
`peticion` → clase 1 nombrando las dos. **Escritura**: temporal + `rename` (atómica).

## 11. Raíz de grabación y directorio de grabaciones

| Entidad | Quién la declara | Qué contiene | Validación |
|---|---|---|---|
| Raíz de grabación (`ConRaizDeGrabacion`) | quien construye el cliente en modo grabación (FR-064) | `<fuente>/<nombre>.json` por cada petición grabada | existe, es directorio, `<raíz>/<fuente>` se puede crear; si no, clase 2 en `New` |
| Directorio de grabaciones (`Replay(dir)`) | quien construye el cliente de reproducción | `<nombre>.json` de **una** fuente | existe y es directorio; si no, clase 2 en `Replay` |

Cuando la raíz es el `testdata/` del repositorio, la ruta es la del enunciado del hito:
`testdata/<fuente>/<nombre>.json`. El paquete no conoce ninguna ruta del repositorio (FR-064).

## 12. Flujo de una operación (`Pedir`)

```
Pedir(ctx, ejecucion, p)
  ├─ validar método y dirección ……………………………………………… clase 2 si falla
  ├─ ejecucion.DryRun → Respuesta{Ensayo: true} ………………… sin bajar por la cadena (FR-050)
  └─ bucle de redirecciones (≤ 10 saltos, sin repetir dirección)
       ├─ ctx.Err() != nil → clase 4
       ├─ petición por la cadena:
       │    identificar → [robots → reintentar → ritmo → grabar → transporte | reproducir]
       │      robots:   sitio sin evaluar → obtener robots.txt por la cadena inferior (FR-016, FR-017)
       │                denegado → clase 5 (FR-014, FR-015, FR-030)
       │      reintentar: 5xx / transporte → esperar (jitter, interrumpible) y repetir ≤ intentos
       │      ritmo:    Wait(ctx) del sitio de destino
       ├─ 3xx con Location → cerrar cuerpo, resolver destino, siguiente salto
       ├─ 429 → clase 5 con Retry-After (FR-030)
       ├─ 5xx (ya agotados los reintentos) → clase 4 (FR-029)
       ├─ 3xx sin Location → clase 4
       └─ resto → leer cuerpo entero, cerrar, Respuesta (FR-032)
```

En reproducción el bucle es idéntico; la cadena es `identificar → reproducir`, y un fichero ausente o
ajeno produce clase 1 (FR-047).
