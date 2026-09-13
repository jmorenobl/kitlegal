# Contrato: fallos del applet `boe`, su clase, su código y su sobre

Tabla **cerrada**: toda ruta de fallo del applet es una fila. Decisiones en [research.md](../research.md) D3, D10.
«Instante» es el de `httpx` ([httpx-acepta-e-instante.md](./httpx-acepta-e-instante.md) §2); «montaje» es el
reloj del kernel al montar el sobre. Ninguna fila produce el código 6 (FR-100).

## 1. Tabla

| # | Situación | Clase | Código | `fuente` / `url` del sobre de fallo | `fecha_consulta` | ¿Se pidió algo? | ¿Caché? |
|---|---|---|---|---|---|---|---|
| 1 | Sin verbo, verbo desconocido, falta un argumento obligatorio, bandera desconocida | `argumentos` | 2 | `kitlegal.cli` / `kitlegal:cli` | montaje | no | no se abre |
| 2 | Norma fuera de `^BOE-A-[0-9]{4}-[0-9]{1,9}$` | `argumentos` | 2 | `kitlegal.cli` / `kitlegal:cli` | montaje | no | no se abre |
| 3 | Bloque fuera de `^[A-Za-z0-9]{1,64}$` (en `articulo` o en cualquiera de `articulos`) | `argumentos` | 2 | `kitlegal.cli` / `kitlegal:cli` | montaje | no | no se abre |
| 4 | Texto de búsqueda sin ninguna palabra (vacío o solo espacio en blanco, sin operadores) | `argumentos` | 2 | `kitlegal.cli` / `kitlegal:cli` | montaje | no | no se abre |
| 5 | `--offline` y entrada ausente o caducada | `fuente-no-disponible` | 4 | `boe.legislacion-consolidada` / recurso que falta (bloque en `articulo` y `articulos`) | montaje | no | solo lectura, intacta |
| 6 | HTTP 404 al pedir un bloque | `no-encontrado` | 3 | boe / dirección del bloque | instante | sí (en `articulo`, **no** los metadatos) | nada escrito (en `articulos`, los anteriores sí) |
| 7 | HTTP 404 al pedir índice, metadatos o análisis | `no-encontrado` | 3 | boe / dirección de ese recurso | instante | sí | nada escrito |
| 8 | `data` vacío en índice, metadatos o análisis (FR-041, FR-051, FR-061) | `no-encontrado` | 3 | boe / dirección de ese recurso | instante | sí | nada escrito |
| 9 | HTTP 404 en `buscar` | `fuente-no-disponible` | 4 | boe / dirección de búsqueda | instante | sí | nada escrito |
| 10 | Otro estado no 2xx que `httpx` entrega (400, 401, 403, 405, 406, 410…) | `fuente-no-disponible` | 4 | boe / dirección pedida | instante | sí | nada escrito |
| 11 | 5xx tras agotar los reintentos | `fuente-no-disponible` | 4 | boe / dirección pedida | instante (último intento) | sí | nada escrito |
| 12 | Fallo de transporte tras reintentos; redirección en bucle, excedida o no seguible | `fuente-no-disponible` | 4 | boe / dirección pedida | instante | sí | nada escrito |
| 13 | `--timeout` agotado (antes de emitir, esperando turno, entre reintentos o leyendo) | `fuente-no-disponible` | 4 | boe / dirección de la petición afectada | instante (abandono o último intento) | depende | nada escrito |
| 14 | HTTP 429 | `limite-o-tos` | 5 | boe / dirección pedida | instante | sí | nada escrito |
| 15 | `robots.txt` que deniega la ruta o que no se pudo obtener | `limite-o-tos` | 5 | boe / dirección **pedida** (no la del `robots.txt`) | instante (abandono) | no (solo el `robots.txt`) | nada escrito |
| 16 | Bloque: cuerpo no UTF-8, XML mal formado, varias raíces o ningún `bloque` bajo la raíz (FR-014) | `fuente-no-disponible` | 4 | boe / dirección del bloque | instante | sí | nada escrito |
| 17 | JSON ilegible, raíz no objeto, elemento de metadatos o análisis que no es objeto, o tipo inesperado donde se espera texto | `fuente-no-disponible` | 4 | boe / dirección pedida | instante | sí | nada escrito |
| 18 | `articulo`: bloque obtenido y metadatos con cualquier fallo de las filas 7-17 (FR-013) | la de ese fallo | 3, 4 o 5 | boe / dirección de **metadatos** | instante de los metadatos | sí | nada escrito |
| 19 | `articulos`: fallo en el bloque *k* o en sus metadatos | la de ese fallo | 2-5 | boe / dirección de la petición que falló | su instante | bloques 1…*k* (no los posteriores) | bloques anteriores escritos; el *k*, no |
| 20 | Entrada de caché con la clave correcta que no se puede leer | `inesperado` | 1 | boe / dirección del recurso consultado | montaje | no | intacta |
| 21 | Fallo de la caché al abrir, leer o escribir (`cache.Error`) | la de `cache.Error` (1, 2 o 4) | 1, 2 o 4 | boe / dirección del recurso consultado | montaje | según el punto | lo que diga `cache.Error` |
| 22 | Dependencia de composición ausente o consulta de otro tipo | `inesperado` | 1 | `kitlegal.cli` / `kitlegal:cli` | montaje | no | no se abre |
| 23 | `--dry-run` sin ningún otro fallo | — | 0 | sin sobre; descripción en la salida de error | — | no | solo lectura, intacta |

`--dry-run` con cualquier fallo de las filas 1-5 o 20-22 conserva ese fallo y su código.

## 2. Mensajes

Cada mensaje nombra lo necesario para entender el fallo sin leer el registro:

| Filas | Nombra |
|---|---|
| 2, 3 | el valor recibido y la forma esperada |
| 4 | que la búsqueda no tiene ninguna palabra |
| 5 | `--offline`, el verbo y la clave de la entrada |
| 6 | la norma y el bloque |
| 7, 8 | la norma y el recurso (índice, metadatos, análisis) |
| 9, 10 | el estado HTTP y la dirección |
| 16, 17 | qué no se pudo interpretar (cuerpo, raíz, bloque, campo) y la dirección |
| 18 | el bloque obtenido y que la vigencia no se pudo comprobar |
| 19 | la posición y el id del bloque que falló |
| 20 | la clave |

Los de `httpx` y `cache` conservan su texto y se envuelven con el contexto del verbo (`%w`).

## 3. Tests que lo fijan

- `internal/source/boe/errores_test.go`: `TestClasesDeErrorDeBoe` (una subprueba por fila 2-4, 6-10, 16-20, 22
  sobre el error real, con `cli.Clasificar` y `cli.CodigoSalida`), `TestErrorDeBoeMensajes`.
- `internal/source/boe/peticiones_test.go`: `TestPedirClasificaEstados` (filas 6-10 con un `Pedidor` de prueba).
- `internal/source/boe/{articulo,metadatos,indice,analisis,buscar}_test.go`: filas 5-8, 14, 16-19 sobre
  grabaciones y sintéticos.
- `internal/app/boe_test.go`: `TestCodigosDeSalidaDeBoe` (filas 1-6, 11, 14, 18 y 19 por el kernel en proceso:
  código, clase, `fuente`, `url` y `fecha_consulta`, con reloj controlado en `httpx.ConHora`).
- `internal/httpx/instante_test.go`: los instantes de las filas 11-15.
