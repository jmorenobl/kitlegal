# Contrato · El formulario de consulta en `internal/httpx`

La única excepción a GET y HEAD del módulo (constitución, principio I; FR-030 a FR-034). Todo lo demás del paquete
queda como está: la cadena, sus garantías, `Pedir`, `Replay`, los reintentos y las redirecciones de las demás fuentes.

## 1. Lo que gana el paquete

```go
// ConFormulario declara la única dirección a la que el cliente puede enviar un
// formulario de consulta: la de la fila de su fuente en docs/SOURCES.md.
func ConFormulario(direccion string) Opcion

type Peticion struct {
	Metodo string            // «GET», «HEAD» o, solo en una Consulta, «POST»
	URL    string
	Acepta string
	Campos map[string]string // los campos de un formulario; solo con «POST»
}

// Consulta abre una consulta: sus peticiones comparten las cookies que el
// sitio les dé, y con nadie más. Solo la da un cliente con formulario.
func (c *Cliente) Consulta() (*Consulta, error)

// Pedir es Cliente.Pedir dentro de la consulta: mismas comprobaciones, mismo
// ensayo, misma cadena y mismo instante; con sus cookies y sin seguir
// redirecciones.
func (q *Consulta) Pedir(ctx context.Context, ejecucion schema.Contexto, p Peticion) (Respuesta, error)
```

`ConFormulario` vale en `New` y en `Replay`. Una dirección vacía, que no se puede interpretar, que no es absoluta o
que no es `http` ni `https` es un error de `argumentos` al construir, como las demás opciones.

## 2. Qué se admite

Todo rechazo es de la clase `argumentos`, antes de abrir nada y también en ensayo (FR-030, FR-031).

| Quién pide | Petición | Resultado |
|---|---|---|
| `Cliente.Pedir`, con formulario declarado o sin él | GET o HEAD, sin `Campos` | como hoy |
| `Cliente.Pedir` | `POST`, a cualquier dirección | `argumentos` |
| `Cliente.Pedir` o `Consulta.Pedir` | PUT, DELETE, PATCH o cualquier otro método | `argumentos` |
| `Cliente.Pedir` o `Consulta.Pedir` | GET o HEAD con `Campos` | `argumentos` |
| `Cliente.Consulta` de un cliente sin `ConFormulario` | — | `argumentos`: no hay consulta |
| `Consulta.Pedir` | GET o HEAD, sin `Campos` | se pide, con las cookies de la consulta |
| `Consulta.Pedir` | `POST` a la dirección declarada, con al menos un campo | se envía |
| `Consulta.Pedir` | `POST` a otra dirección, o sin campos | `argumentos` |

La dirección se compara entera con la declarada, las dos tal como las escribe `url.URL.String`.

## 3. Lo que lleva un envío

- Método `POST`; cuerpo `url.Values.Encode` de los campos —ordenados por clave— con
  `Content-Type: application/x-www-form-urlencoded`; y `X-Requested-With: XMLHttpRequest` (research D5).
- La identificación del proyecto, el `robots.txt` del sitio antes de la primera petición del cliente, su turno en el
  ritmo, el contexto obligatorio y el instante de emisión: baja por la misma cadena que cualquier petición (FR-033).
- Con `--dry-run`, nada sale: `Respuesta{Ensayo: true}`, y `Descripcion()` da `POST <dirección>`.
- Las cookies que el sitio dio a esa consulta, y solo esas (FR-032).

## 4. Cookies y redirecciones

- Cada `Consulta` tiene su almacén de cookies, en memoria, y lo ponen y lo leen sus propias peticiones. Otra
  `Consulta` del mismo cliente empieza sin ninguna, y `Cliente.Pedir` nunca lleva ni guarda cookies. Nada se escribe en
  disco, salvo lo que una grabación guarda de la petición y de la respuesta.
- La petición del `robots.txt` no lleva las cookies de ninguna consulta.
- Dentro de una `Consulta` no se sigue ninguna redirección: la respuesta 3xx se entrega con su `Estado` y sus
  cabeceras, y no se pide nada a su destino (FR-021). Fuera de una `Consulta`, `Pedir` sigue hasta 10 saltos, como hoy.
- El resto de la clasificación no cambia: un 429 es `limite-o-tos`, un 5xx es `fuente-no-disponible` y los demás
  estados se entregan.

## 5. Grabación y reproducción

Dos envíos a la misma dirección con campos distintos son dos grabaciones (FR-034).

- **Nombre**: el de hoy, seguido de `_c_` y del cuerpo codificado, saneado con la misma regla. Si pasa de 120
  caracteres, los 100 primeros, un guion y 8 hexadecimales de `sha256("<MÉTODO> <dirección>\n<cuerpo>")`.
- **Fichero**: `peticion` gana la clave `cuerpo`, entre `url` y `cabeceras`, con el cuerpo codificado. Solo la lleva
  una petición con campos. `formato` sigue en 1.
- **Emparejamiento**, al grabar y al reproducir: método, dirección y cuerpo, exactos. Un envío cuyos campos no están
  grabados es, como hoy una petición sin grabación, un fallo `inesperado` que nombra el fichero que falta.
- Una petición sin campos se nombra, se graba y se reproduce byte a byte como hoy: las grabaciones de `testdata/` no
  se tocan.

Ejemplo, con el nombre que da el prototipo (research M1):

```text
POST_https_www.poderjudicial.es_search_search.action_c_ECLI_ECLI_3AES_3ATS_3A2023_3A3144_action_quer-2740d948.json
```

```json
{
  "formato": 1,
  "grabado_en": "2026-10-06T10:00:05Z",
  "peticion": {
    "metodo": "POST",
    "url": "https://www.poderjudicial.es/search/search.action",
    "cuerpo": "ECLI=ECLI%3AES%3ATS%3A2023%3A3144&action=query&databasematch=AN&recordsPerPage=10&sort=IN_FECHARESOLUCION%3Adecreasing&start=1",
    "cabeceras": { "Content-Type": ["application/x-www-form-urlencoded"], "Cookie": ["…"], "User-Agent": ["kitlegal/… (+https://kitlegal.es/bot)"], "X-Requested-With": ["XMLHttpRequest"] }
  },
  "respuesta": { "estado": 200, "cabeceras": { "…": ["…"] }, "cuerpo": "…" }
}
```

## 6. Tests (`internal/httpx`, sin red; FR-112)

| Test | Qué fija |
|---|---|
| `TestFormularioAdmitido` | la tabla de §2, fila a fila, contando en un servidor de prueba que cada rechazo no emite ninguna petición |
| `TestEnvioDelFormulario` | §3: el cuerpo, sus dos cabeceras, la identificación, el `robots.txt` antes, el turno y el ensayo |
| `TestCookiesDeLaConsulta` | §4: la cookie de la página llega al formulario de su consulta; la primera petición de la consulta siguiente y las de `Cliente.Pedir` no la llevan; el `robots.txt`, tampoco |
| `TestConsultaSinRedirecciones` | una redirección en la página y otra en el formulario se entregan y su destino no recibe nada; fuera de una consulta se siguen |
| `TestGrabacionDeFormularios` | §5: dos envíos, dos ficheros, cada uno reproduce el suyo; uno sin grabar falla nombrando su fichero; una grabación de H4 se reproduce y se regraba igual |
| `TestNombreDeGrabacion` (gana casos) | los nombres de research M1 |
| `TestOpcionesInvalidas`, `TestPedirRechazaMetodo`, `TestSuperficieExportada` (cambian) | `ConFormulario` inválida; el `POST` fuera de una consulta sigue rechazado; la superficie exportada nueva |

## 7. El sitio de prueba: `internal/httpx/httpxtest`

Un paquete de apoyo para tests, que el binario no enlaza: `httpxtest.NuevoSitio(t, responder)` levanta un servidor
local y apunta cada petición que recibe —método, ruta, cuerpo y cookie—. Existe porque solo el árbol de
`internal/httpx` puede importar `net/http` (R2), y el control de FR-021 cuenta las peticiones de una consulta del
adaptador contra un sitio de prueba, con su `robots.txt` y sin reproducción
([fuente-cendoj-y-grabacion.md §6](./fuente-cendoj-y-grabacion.md)).
