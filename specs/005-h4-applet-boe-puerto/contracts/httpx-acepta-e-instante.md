# Contrato: `internal/httpx` pide un formato y dice cuándo emitió

Ampliación retrocompatible del contrato del cliente de H2 (`specs/003-h2-internal-httpx-cliente/contracts/`).
Decisión en [research.md](../research.md) D4. Todo lo no mencionado aquí sigue igual.

## 1. `Peticion.Acepta`

```go
type Peticion struct {
    Metodo string
    URL    string
    // Acepta es el tipo de contenido que se pide en la cabecera Accept («application/xml»,
    // «application/json»). Vacío: la petición no lleva Accept.
    Acepta string
}
```

| Situación | Resultado |
|---|---|
| `Acepta` vacío | sin cabecera `Accept` (igual que H2) |
| `Acepta` que `mime.ParseMediaType` acepta y sin caracteres de control | `Accept: <Acepta>` en la petición y en cada salto de redirección; **no** en la del `robots.txt` |
| `Acepta` inválido o con un carácter de control | «argumentos» (2), sin abrir nada; también en ensayo |

- La grabación guarda la cabecera (guarda las cabeceras de la petición). La reproducción empareja **solo** por
  método y dirección (H2 §4, sin cambios).
- `Respuesta.Descripcion()` sigue siendo `«<Metodo> <URL>»`.

## 2. `Instante`

```go
type Respuesta struct {
    // … campos de H2 …
    // Instante es el de emisión de la petición que entregó la respuesta: la del último salto y el último intento.
    // Cero en ensayo.
    Instante time.Time
}

type Error struct {
    // … campos de H2 …
    // Instante es el de emisión del último intento si llegó a emitirse; si no, el instante en que se abandonó.
    Instante time.Time
}
```

| Situación | `Instante` |
|---|---|
| Respuesta al primer intento | hora al entregar la petición al transporte |
| Respuesta tras reintentos (5xx o transporte) | hora del **último** intento |
| Respuesta tras redirecciones | hora del último salto |
| Error tras agotar reintentos (5xx, transporte) | hora del último intento |
| 429 | hora del intento que lo recibió |
| Denegada por `robots.txt` o sin permiso para obtenerlo | hora al abandonar en `Pedir` (la obtención del `robots.txt` **no** cuenta) |
| Sin turno en el ritmo, plazo agotado antes de emitir | hora al abandonar |
| Plazo agotado esperando para reintentar | hora del último intento emitido |
| Argumentos inválidos (método, dirección, `Acepta`) | hora al abandonar |
| Reproducción | hora de `ConHora` al servir la grabación |
| Ensayo | cero |

## 3. `ConHora`

```go
// ConHora declara de dónde sale el instante de emisión. Por omisión, time.Now. Vale para New y para Replay.
func ConHora(ahora func() time.Time) Opcion
```

Nula → «argumentos» (2) al construir.

## 4. Mecanismo y orden de la cadena

```
New:     identificación → robots.txt → reintentos → ritmo → [grabación] → marca de emisión → transporte
Replay:  identificación → marca de emisión → transporte de reproducción
```

- `Pedir` pone en el contexto una marca (clave sin exportar) y la reinicia antes de cada salto.
- `conMarcaDeEmision` escribe `hora()` en la marca justo antes de entregar la petición al transporte; cada intento
  la sobrescribe.
- El decorador de `robots.txt` obtiene su fichero con un contexto que **oculta** la marca.
- Al volver, `Pedir` copia la marca en `Respuesta.Instante` o `Error.Instante`; si está vacía, pone `hora()`.

## 5. Tests que lo fijan

`internal/httpx/cliente_test.go`: `TestPedirConAcepta`, `TestConHoraRechazaNula`.
`internal/httpx/instante_test.go`: `TestInstanteDeEmision` con los subtests `acierto`, `acierto-tras-reintentos`,
`fallo-tras-reintentos`, `denegada-por-robots`, `sin-turno`, `redireccion`, `reproduccion`, `ensayo` y
`argumentos`, con `httptest` (solo dentro de `internal/httpx`, R2) y un reloj que devuelve instantes distintos en
cada llamada.
