# Contrato: el applet `graph`

FR-050 a FR-067, FR-070. Decisiones: [../research.md](../research.md) D15, D16, D18, D19. Tipos:
[../data-model.md](../data-model.md) §5-§7. Almacén: [almacen-world-db.md](./almacen-world-db.md).

## 1. Invocación

```text
kitlegal graph show <id>      graph show <id>      (por el nombre del programa, multicall)
kitlegal graph stats          graph stats
kitlegal graph check          graph check
```

- Sin verbo por omisión; `kitlegal graph` sin verbo sale con 2 nombrando los tres, en este orden: `argumentos
  inválidos: el applet "graph" no declara ningún verbo por omisión, así que hay que nombrar uno; verbos de graph:
  show, stats, check` (FR-050).
- Hereda las ocho banderas globales (`--json`, `--timeout`, `--offline`, `--dry-run`, `--describe`, `--no-graph`,
  `--asunto`, `--verbose`) y la ayuda (FR-050). Descripción del applet:
  `Lee el grafo del mundo: lo que el binario ha observado de las fuentes, con su procedencia.` Verbos:
  - `show`: `Devuelve un nodo del grafo del mundo con sus aristas y su procedencia, sin texto legal.`
  - `stats`: `Cuenta los nodos, las aristas y los textos del grafo del mundo por tipo, relación y fuente.`
  - `check`: `Comprueba el grafo del mundo y devuelve como hallazgos las versiones superadas y las consultas caducadas.`
- `show` recibe exactamente un argumento posicional, `id` (`Id del nodo: un ELI, «ine:<código>», un DIR3…`); `stats` y
  `check`, ninguno.

Composición (`internal/app/grafo.go`):

```go
type DependenciasDeGrafo struct {
	Reloj   func() time.Time // se lee una vez por invocación; nulo es un defecto de composición
	Almacen []graph.Opcion   // vacío: la regla de ubicación de la caché
}
func DependenciasDelGrafoDelSistema() DependenciasDeGrafo // Reloj: time.Now; Almacen vacío
func AppletGrafo(dependencias DependenciasDeGrafo) Applet
```

Cada verbo lee el reloj una vez, rechaza sin abrir nada un id que `grafo.ValidarID` no admite, abre la lectura con
`graph.Leer(ctx, Almacen...)`, lee y la cierra; un fallo al cerrarla se une al del verbo con `errors.Join`. Con el
reloj nulo el verbo falla con `inesperado` («graph: el applet se compuso sin reloj…») y el sobre de fallo lo firma el
kernel.

## 2. Procedencia y efectos

- Sobre de éxito: `fuente` = `kitlegal.graph`, `url` = `kitlegal:applet/graph`, `fecha_consulta` = el instante del
  reloj de la invocación, que es también el instante de `check` (FR-051, FR-067). Todo fallo que decide el applet
  (un id que `ValidarID` rechaza, un id que no está, cada error de `internal/graph`) lleva la misma procedencia; los
  que decide el analizador antes del applet los firma el kernel.
- Ningún verbo crea ni modifica `world.db`, `world.db-wal`, `world.db-journal` ni su directorio, ni cambia el
  contenido del grafo (FR-004, FR-005), y ninguno entrega operaciones (FR-046): su `Resultado.Grafo` es el valor cero.
  Sin auxiliares —el estado normal— no cambia ni un byte de nada, también si el proceso no puede escribir
  `world.db` (lo lee con `immutable=1`, research V46) o si `world.db` es un enlace simbólico (los auxiliares se
  buscan donde SQLite los crea, research V47). Un `world.db` de 0 bytes se lee como grafo vacío sin abrir SQLite, así
  que tampoco cambia ningún fichero haya los auxiliares que haya, ni un `world.db-wal` no vacío junto a él, que abrirlo
  borraría (contracts/almacen-world-db.md §3, paso 1; research V49). Con algún auxiliar de WAL junto a un `world.db`
  de más de 0 bytes (otra conexión abierta, un `-wal` huérfano o un `-shm` suelto), SQLite reescribe o crea
  `world.db-shm`, y con un `-shm` suelto crea un `world.db-wal` vacío (el `-shm` de un lector sobre una base limpia
  queda igual; el de un escritor con marcos se reescribe, research V48):
  desviación declarada de FR-004, FR-031 y SC-004 (contracts/almacen-world-db.md §3; research D10). Con un `world.db-journal` caliente (una transacción interrumpida
  sin deshacer) no leen: salen con 1 sin cambiar nada, porque deshacerla sería modificar `world.db` (research D10,
  V43).
- `--no-graph` no cambia nada: la misma lectura, la misma salida, los mismos códigos (FR-031). `--offline` tampoco (el
  grafo es local). Con `--dry-run` leen igual y el kernel escribe su descripción, sin sobre, como `territorio`.
- Sin `--json`: la tabla mínima del kernel (FR-055).

## 3. `data`

Los ejemplos son el `data` que da el binario (en el sobre va en una sola línea, con las claves en este orden; aquí
con sangría para leerlo) tras `boe articulo BOE-A-2015-10565 a21` con el reloj `2026-09-28T12:00:00Z` y
`territorio resolver Leganés` (guiones `h7-grafo-memoria` y `h7-grafo-show`).

### 3.1 `show <id>` → `grafo.Ficha`

`kitlegal graph show 'eli/es/l/2015/10/01/39#a21' --json`:

```json
{
  "nodo": {
    "id": "eli/es/l/2015/10/01/39#a21",
    "tipo": "Bloque",
    "datos": {"bloque": "a21"},
    "primera_observacion": "2026-09-28T12:00:00Z",
    "ultima_observacion": {
      "fuente": "boe.legislacion-consolidada",
      "url": "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21",
      "fecha_consulta": "2026-09-28T12:00:00Z"
    }
  },
  "salientes": [
    {
      "relacion": "eli:has_version",
      "id": "eli/es/l/2015/10/01/39#a21@20161002:sha256:98d3b9d4686a3155f48641841df7023e2c17a693fcd0e11d6beef2615abcec7c",
      "primera_observacion": "2026-09-28T12:00:00Z",
      "ultima_observacion": {
        "fuente": "boe.legislacion-consolidada",
        "url": "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21",
        "fecha_consulta": "2026-09-28T12:00:00Z"
      }
    }
  ],
  "entrantes": [
    {
      "relacion": "eli:has_part",
      "id": "eli/es/l/2015/10/01/39",
      "primera_observacion": "2026-09-28T12:00:00Z",
      "ultima_observacion": {
        "fuente": "boe.legislacion-consolidada",
        "url": "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21",
        "fecha_consulta": "2026-09-28T12:00:00Z"
      }
    }
  ]
}
```

- `datos` son los guardados (los de la última observación), un objeto con las claves ordenadas como las ordena
  `encoding/json` (en un `BloqueVersion`: `fecha_version`, `fecha_vigencia`, `hash_texto`, `norma_modificadora`) y
  `{}` si no hay ninguno; nunca el cuerpo de un bloque (FR-041, FR-070).
- Un id que no está sale con 3 y el mensaje `no encontrado: el id "<id>" no está en el grafo del mundo`.
- `salientes` y `entrantes` van ordenadas por `relacion` y después por `id`, comparando bytes, y nunca nulas (FR-053).
- Toda fecha se reproduce carácter a carácter como la escribió el sobre de su observación (FR-053).

### 3.2 `stats` → `grafo.Recuento`

`kitlegal graph stats --json`:

```json
{
  "nodos": 5, "aristas": 3, "textos": 1,
  "nodos_por_tipo": [
    {"tipo": "Bloque", "fuente": "boe.legislacion-consolidada", "nodos": 1},
    {"tipo": "BloqueVersion", "fuente": "boe.legislacion-consolidada", "nodos": 1},
    {"tipo": "Municipio", "fuente": "kitlegal.territorio", "nodos": 1},
    {"tipo": "Norma", "fuente": "boe.legislacion-consolidada", "nodos": 1},
    {"tipo": "Organo", "fuente": "kitlegal.territorio", "nodos": 1}
  ],
  "aristas_por_relacion": [
    {"relacion": "eli:has_part", "fuente": "boe.legislacion-consolidada", "aristas": 1},
    {"relacion": "eli:has_version", "fuente": "boe.legislacion-consolidada", "aristas": 1},
    {"relacion": "lb:pertenece_a", "fuente": "kitlegal.territorio", "aristas": 1}
  ]
}
```

La fuente de cada par es la de la última observación; orden por (tipo, fuente) y (relación, fuente) comparando bytes;
ningún par con 0; los textos solo en total; con el grafo ausente o vacío, tres ceros y dos listas vacías (FR-054).

### 3.3 `check` → `[]grafo.Hallazgo`

Tras leer `a21` con la grabación de H4 (A, `fecha_vigencia` `20161002`) con el reloj `2026-09-28T12:00:00Z` y con la
derivada `version-posterior` (B, `20250101`) con `2026-09-29T12:00:00Z`, `kitlegal graph check --json` con el reloj
`2026-10-06T12:00:00Z`: la versión A ha caducado (su última observación es la de A; la `Norma`, el `Bloque` y la
versión B tienen la de B, que caduca justo en ese instante y por tanto no antes) y está superada por B (`<url>` es
`https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21`, escrita entera
en la salida):

```json
[
  {
    "clase": "fuente-caducada",
    "id": "eli/es/l/2015/10/01/39#a21@20161002:sha256:98d3b9d4686a3155f48641841df7023e2c17a693fcd0e11d6beef2615abcec7c",
    "explicacion": "La consulta de [BOE-A-2015-10565, bloque a21] a boe.legislacion-consolidada en <url> del 2026-09-28T12:00:00Z tenía una vigencia de 604800 s y caducó el 2026-10-05T12:00:00Z.",
    "procedencia": {"fuente": "boe.legislacion-consolidada", "url": "<url>", "fecha_consulta": "2026-09-28T12:00:00Z"},
    "vigencia_segundos": 604800
  },
  {
    "clase": "version-obsoleta",
    "id": "eli/es/l/2015/10/01/39#a21@20161002:sha256:98d3b9d4686a3155f48641841df7023e2c17a693fcd0e11d6beef2615abcec7c",
    "explicacion": "La versión de [BOE-A-2015-10565, bloque a21] con fecha de vigencia 20161002 está superada por la de fecha de vigencia 20250101, observada en <url> el 2026-09-29T12:00:00Z.",
    "procedencia": {"fuente": "boe.legislacion-consolidada", "url": "<url>", "fecha_consulta": "2026-09-29T12:00:00Z"},
    "fecha_vigencia": "20161002",
    "fecha_vigencia_reciente": "20250101"
  }
]
```

Con la `Norma` citada por su `identificador`, la explicación de `fuente-caducada` dice `La consulta de
BOE-A-2015-10565 a …` (guion `h7-grafo-fuente-caducada`).

`data` es la lista, vacía y nunca nula sin hallazgos (FR-060); ordenada por clase y después por id, comparando bytes
(FR-062). Las claves propias de una clase no aparecen en la otra (`omitempty`, research V7).

## 4. Códigos

| Situación | Código | Clase |
|---|---|---|
| Éxito, con hallazgos o sin ellos | 0 | — |
| `show` sin id, con dos, con id vacío, formado solo por caracteres de espacio en blanco o que contiene algún carácter de control (`grafo.ValidarID`, research D35); `stats` o `check` con un argumento posicional; bandera desconocida | 2 | `argumentos` |
| `KITLEGAL_CACHE_DIR` presente y vacía o que nombra algo que no es directorio; sin ella y sin `HOME` (también con `--no-graph`) | 2 | `argumentos` |
| `show` de un id ausente —también uno con espacios y algo más (`a b`), con un carácter de formato como U+200B o con bytes que no son UTF-8—, o con el grafo ausente o sin esquema (el mensaje nombra el id) | 3 | `no-encontrado` |
| Plazo de `--timeout` agotado esperando un bloqueo | 4 | `fuente-no-disponible` |
| `world.db` inutilizable (no es base, dañada, directorio, no deja abrirse —también sin deshacer un diario caliente, que leer exigiría modificar—, esquema posterior) o bloqueo más largo que la espera propia; el mensaje nombra la ruta (contracts/almacen-world-db.md §6) | 1 | `inesperado` |
| `check` sobre lo que ninguna entrega guarda y `grafo.Comprobar` no puede comprobar (una fecha de consulta que no es RFC 3339); el mensaje nombra `world.db` pero **no** su ruta, que `graph.Lectura` no expone: `grafo: world.db guarda lo que ninguna entrega escribe y no se puede comprobar; no se modifica: <causa>` | 1 | `inesperado` |
| Dependencias sin reloj (defecto de composición) | 1 | `inesperado` |

Mensajes: un id que `ValidarID` rechaza, `el id "<id>" no puede ser el de ningún nodo: <motivo>`, con el motivo
`está vacío`, `solo tiene caracteres de espacio en blanco` o `contiene el carácter de control U+XXXX` (en ese orden de
comprobación); un id ausente, `no encontrado: el id "<id>" no está en el grafo del mundo`; los fallos de
`internal/graph`, los de contracts/almacen-world-db.md §6. El id de `show` es un `cli.Literal`: el analizador de la
línea de órdenes (Kong, en `internal/cli`) cambia por U+FFFD los bytes que no son UTF-8 de un argumento de texto, y
un `cli.Literal` los conserva, así que `show` busca el id con los bytes recibidos y, si no está, sale con 3
nombrándolo con `%q` (`"…\xff"`), no con U+FFFD; para `--describe` el id sigue siendo una cadena. Lo fijan
`TestAppletGrafo/no-encontrado`, con un nodo cuyo id es el mismo con U+FFFD en lugar del byte, que `show` encuentra,
y `TestAnalizarLiteral` (`internal/cli`).

Las dos clases de caracteres del id de `show` (FR-052; research D35, V37):

- **Espacio en blanco**: la propiedad White_Space de Unicode (`unicode.IsSpace`): U+0020, `\t`, `\n`, `\v`, `\f`,
  `\r`, U+0085, U+00A0, U+2003 y el resto de la tabla `White_Space`. Un id **formado solo** por ellos sale con 2; un id
  que además tiene otro carácter es válido y se busca tal cual, sin recortar.
- **Control**: la categoría Cc de Unicode (`unicode.IsControl`): U+0000-U+001F y U+007F-U+009F. Un id que contiene
  **alguno** sale con 2.

Un ejemplo de cada una: U+00A0 solo y U+2003 solo (espacio en blanco), U+0000 dentro de un id y U+007F solo (control),
U+0085 solo (de las dos). U+200B no es de ninguna de las dos.

## 5. Explicaciones (FR-061, FR-063, FR-066)

Plantillas literales; `<cita>` es la de data-model §6 (la del bloque para `Bloque` y `BloqueVersion`, el
`identificador` para `Norma`, el id si no hay otra); todo valor entre `< >` se copia carácter a carácter de lo guardado.

- `version-obsoleta`:
  `La versión de <cita> con fecha de vigencia <fecha_vigencia de V> está superada por la de fecha de vigencia <fecha_vigencia de la más reciente>, observada en <url> el <fecha_consulta>.`
  (url y fecha de la última observación de la versión más reciente).
- `fuente-caducada`:
  `La consulta de <cita> a <fuente> en <url> del <fecha_consulta> tenía una vigencia de <N> s y caducó el <instante>.`
  con `<instante>` = `fecha_consulta` + N segundos en el mismo desplazamiento, escrito con `RFC3339Nano` (research D18):
  `2026-09-28T12:00:00.5+02:00` → `2026-10-05T12:00:00.5+02:00`; `2026-09-28T12:00:00.12+02:00` →
  `2026-10-05T12:00:00.12+02:00`; `2026-09-28T12:00:00Z` → `2026-10-05T12:00:00Z`.

## 6. Esquema publicado

`schemas/grafo.json`, entidad `grafo`, título `graph · grafo`, partes `graph check`, `graph show` y `graph stats`,
generado con `TestEsquemasPublicados -actualizar-esquemas` y vigilado por `make schema-check` (FR-051).

## 7. Qué lo vigila

| Control | Test |
|---|---|
| Verbos, argumentos, códigos 0/1/2/3/4, procedencia, `--no-graph` igual, tabla sin `--json` | `TestAppletGrafo` (`internal/app/grafo_test.go`) sobre un registro local |
| Toda salida correcta, y el sobre de cada fallo que decide el applet (2, 3, 4 y 1), contra `schemas/grafo.json` | `TestSalidaDelGrafoContraSchemas` |
| Reglas, orden, explicaciones literales, instante de caducidad con desplazamiento, cita; `fecha_vigencia` válida solo con cifras ASCII | `TestComprobar`, `TestExplicaciones` (`internal/core/grafo`) |
| Id de `show`: vacío, U+0020, U+00A0 y U+2003 solos, `\t` solo, U+0085 solo, U+007F solo y `a` + U+0000 + `b` → `argumentos`; `a b`, ` a`, U+200B y un id con un byte que no es UTF-8 → válidos | `TestValidarID` (`internal/core/grafo`) |
| Ningún verbo devuelve una línea (≥ 20 caracteres) del texto de ninguna respuesta grabada de `articulo` y `articulos` (FR-070, SC-006) | `TestNingunVerboDelGrafoDevuelveTexto` |
| Esquema de versión posterior en los tres verbos → 1 y fichero intacto (SC-011) | `TestCodigosDelGrafo` |
| `check` < 3 s y `stats` < 1 s sobre 10 000 nodos y 10 000 aristas (SC-008) | `TestCosteDelGrafo` (`internal/app/coste_test.go`, `make test-tiempos`) |
| e2e | `h7-grafo-show`, `h7-grafo-codigos`, `h7-grafo-version-obsoleta`, `h7-grafo-fuente-caducada`, `h7-grafo-applet` |
