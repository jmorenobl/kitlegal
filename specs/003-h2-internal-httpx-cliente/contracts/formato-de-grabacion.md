# Contrato: formato de grabación y reproducción

Qué escribe `KITLEGAL_RECORD=1`, con qué nombre, y cómo lo lee `Replay(dir)`. Decisiones en
[research.md](../research.md) D12 y D13.

## 1. El fichero

Un objeto JSON, indentado con dos espacios, sin escape de HTML (`<`, `>`, `&` se escriben tal cual),
con **estas claves en este orden** y ninguna más:

```json
{
  "formato": 1,
  "grabado_en": "2026-09-12T14:03:11Z",
  "peticion": {
    "metodo": "GET",
    "url": "http://127.0.0.1:53211/norma?id=BOE-A-2015-10565",
    "cabeceras": {
      "User-Agent": [
        "kitlegal/dev (+https://ventanillalegal.es/bot)"
      ]
    }
  },
  "respuesta": {
    "estado": 200,
    "cabeceras": {
      "Content-Length": [
        "42"
      ],
      "Content-Type": [
        "application/xml; charset=utf-8"
      ],
      "Date": [
        "Fri, 12 Sep 2026 14:03:11 GMT"
      ]
    },
    "cuerpo": "<norma id=\"BOE-A-2015-10565\">…</norma>"
  }
}
```

| Clave | Regla |
|---|---|
| `formato` | siempre `1` en H2; la reproducción rechaza otro valor (clase 1) |
| `grabado_en` | RFC 3339 en UTC, segundos enteros |
| `peticion.metodo`, `peticion.url` | la petición **tal como se emitió**: dirección completa, consulta incluida; es la clave de emparejamiento |
| `peticion.cabeceras`, `respuesta.cabeceras` | objeto `nombre canónico → lista de valores`; claves ordenadas alfabéticamente; se graban todas (FR-037) |
| `respuesta.cuerpo` | cuerpo como texto si es UTF-8 válido |
| `respuesta.cuerpo_base64` | cuerpo en base64 estándar si **no** es UTF-8 válido; excluyente con `cuerpo`; un cuerpo vacío se graba como `"cuerpo": ""` |

**Estabilidad (FR-040)**: dos grabaciones de la misma petición y la misma respuesta son idénticas byte a
byte salvo `grabado_en` y las cabeceras de respuesta que el servidor regenera (`Date` entre ellas). Nada
más puede variar: ni el orden de claves ni el de cabeceras. Se comprueba en `TestGrabarFormatoEstable`
comparando dos ficheros tras neutralizar exactamente esos campos.

**Escritura**: a un fichero temporal en el mismo directorio y `rename` al nombre final; nunca queda una
grabación a medias.

## 2. El nombre

Derivado únicamente de método y dirección completa; nadie lo declara (FR-039):

```
<MÉTODO>_<esquema>_<host>[-<puerto>]<ruta>[_q_<consulta>].json
```

1. `host` en minúsculas; `-<puerto>` solo si la dirección lo declara explícitamente.
2. `ruta`: cada `/` se convierte en `_`; una ruta vacía o `/` no aporta nada.
3. `consulta`: si existe, `_q_` seguido de la consulta con `=` y `&` convertidos en `_`.
4. Saneado final: todo carácter fuera de `[A-Za-z0-9._-]` → `_`; secuencias de `_` colapsadas a una;
   `_` inicial y final recortados.
5. Si el resultado supera **120** caracteres (sin la extensión): los **100** primeros, `-`, y los 8
   primeros dígitos hexadecimales de `sha256("<MÉTODO> <url>")`.

Ejemplos:

| Petición | Fichero |
|---|---|
| `GET http://127.0.0.1:53211/robots.txt` | `GET_http_127.0.0.1-53211_robots.txt.json` |
| `GET http://fuente.prueba/norma?id=BOE-A-2015-10565` | `GET_http_fuente.prueba_norma_q_id_BOE-A-2015-10565.json` |
| `HEAD https://fuente.prueba/` | `HEAD_https_fuente.prueba.json` |
| `GET https://fuente.prueba/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21` | `GET_https_fuente.prueba_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json` (102 caracteres sin la extensión: no supera 120, sin recorte) |
| `GET https://otra.prueba:8443/publicaciones/2026/09/12/resolucion-de-adjudicacion-definitiva/expediente-de-contratacion-abierto-simplificado/anexo-i/documento-1` | `GET_https_otra.prueba-8443_publicaciones_2026_09_12_resolucion-de-adjudicacion-definitiva_expediente-d11bc93b.json` (el nombre saneado tiene 157 caracteres, más de 120: los 100 primeros, `-` y los 8 primeros hexadecimales de `sha256("GET https://otra.prueba:8443/publicaciones/…/documento-1")`, calculados sobre la cadena exacta de la primera columna con `GET` y un espacio delante; regla 5) |

Todas las filas usan los hosts ficticios `fuente.prueba` y `otra.prueba` (o `127.0.0.1`), que son los
únicos que admite la comprobación «solo direcciones locales» del quickstart; `TestNombreDeGrabacion`
copia esta tabla tal cual, así que ningún host real aparece en un test aunque el ejemplo imite la
forma de una dirección del BOE.
| `GET http://fuente.prueba/a,b` **y** `GET http://fuente.prueba/a_b` | `GET_http_fuente.prueba_a_b.json` para las dos (`,` está fuera de `[A-Za-z0-9._-]` y `net/url` la conserva sin escapar en la ruta): la colisión que §3 y §4 resuelven por contenido |

El nombre es determinista (misma petición → mismo nombre en cualquier máquina), legible y portable.
**No es único por construcción**: dos direcciones pueden sanearse al mismo nombre (`/a b` y `/a_b`; o
`/A` y `/a` en un sistema de ficheros insensible a mayúsculas). Por eso:

## 3. Colisión al grabar (FR-039)

Antes de escribir, si `<raíz>/<fuente>/<nombre>.json` existe:

- se lee y se comparan `peticion.metodo` y `peticion.url` con la petición que se va a grabar;
- **iguales** → se sustituye el fichero (una regrabación);
- **distintas** → clase «inesperado» (1) con mensaje que nombra las dos peticiones y el fichero; no se
  escribe nada.

## 4. Emparejamiento al reproducir (FR-047)

`Replay(dir)` deriva el nombre de la petición entrante (§2), abre `<dir>/<nombre>.json` y comprueba que
`peticion.metodo` y `peticion.url` coinciden **exactamente** con la entrante. Las cabeceras no
participan (FR-039).

| Situación | Resultado |
|---|---|
| Fichero existe y la petición coincide | `*http.Response` interno con `estado`, `cabeceras` y `cuerpo` grabados; `Pedir` lo clasifica como si viniera de la fuente (5xx → 4, 429 → 5, 3xx → siguiente salto, resto → `Respuesta`) |
| Fichero no existe | clase 1: «no hay grabación para GET http://…: falta <dir>/<nombre>.json» |
| Fichero existe con otra petición | clase 1: «la grabación <fichero> guarda HEAD http://…/otra y se buscaba GET http://…» |
| Fichero ilegible, JSON inválido o `formato` ≠ 1 | clase 1 nombrando el fichero |

Ninguno de los cuatro casos abre una conexión ni devuelve una respuesta vacía (FR-046, FR-047).

## 5. Lo que la reproducción no aplica (FR-049)

No consulta `robots.txt` (una grabación de `robots.txt` en `dir` queda sin usar y no es error), no
espera ningún ritmo, no reintenta. Sí aplica: contexto, identificación, método, redirecciones grabadas
(cada salto es una búsqueda más en `dir`; si falta, clase 1) y clasificación por estado.

## 6. Dónde vive el árbol

En ningún sitio que el paquete conozca: la raíz la declara `ConRaizDeGrabacion` y el directorio de
reproducción, `Replay(dir)` (FR-064). Cuando la raíz es `testdata/` del repositorio, el fichero queda en
`testdata/<fuente>/<nombre>.json`, la ruta del enunciado del hito; cuando es
`internal/source/<fuente>/testdata`, también vale sin cambiar nada. La elección entre ambas ubicaciones
para los fixtures reales queda para antes de la primera tarea `[datos]` de H4 (`docs/PENDIENTES.md`).

## 7. Los ficheros escritos a mano de este hito

El único directorio de grabaciones que H2 versiona es `internal/httpx/testdata/reproduccion/prueba/`,
con **diez** ficheros escritos a mano contra el host ficticio `fuente.prueba` y ninguno más. La lista
cerrada —nombre, petición, respuesta y test que lo usa— está en [plan.md, «Fixtures»](../plan.md#fixtures)
y no se repite aquí. Cada fichero cumple §1 (claves, orden, `SetEscapeHTML(false)`, `cuerpo` o
`cuerpo_base64`) y su nombre es el que §2 produce para su petición, lo que `TestNombreDeGrabacion`
comprueba fila a fila; el de colisión es el ejemplo `/a,b` de la tabla de §2, y el de `robots.txt` es el
que §5 declara que queda sin usar.
