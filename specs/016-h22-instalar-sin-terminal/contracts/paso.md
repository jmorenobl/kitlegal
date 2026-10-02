# Contrato: el paso que empaqueta (`empaquetar`)

Requisitos: FR-001, FR-004, FR-005, FR-010 a FR-022, FR-030, FR-066, FR-067, FR-069, FR-070. Decisiones en
[../research.md](../research.md) D1 a D5, D8 y D14 a D20. Los bytes de este contrato son los medidos con el prototipo
sobre un snapshot de `751b76e` (research V14 a V16); la versión de los ejemplos es la de ese snapshot.

## 1. La orden

```text
go run ./cmd/empaquetar piezas   -version <versión> -macos <binario> -windows <binario> -icono <png> -salida <carpeta>
go run ./cmd/empaquetar catalogo -version <versión> -sha256 <huella> -salida <fichero>
```

No es un applet ni un verbo de `kitlegal`, no se instala y ningún archivo de la release lo lleva (FR-001). No tiene
banderas globales, ni `--json`, ni sobre: no es una salida del binario. Todas sus banderas son obligatorias. No escribe
nada en la salida estándar.

| | `piezas` | `catalogo` |
|---|---|---|
| Hace | escribe `<carpeta>/kitlegal.mcpb` y `<carpeta>/kitlegal-plugin.zip` | escribe en `<fichero>` el `marketplace.json` de la versión |
| `-version` | la del manifiesto y de `plugin.json`, tal cual: quien llama la da sin `v` | la de la entrada del catálogo y la de su dirección, sin `v` |
| Otras | `-macos`, `-windows`: los dos binarios, que copia sin mirar. `-icono`: un PNG de 512 × 512 px. `-salida`: una carpeta que existe | `-sha256`: 64 dígitos hexadecimales en minúsculas. `-salida`: un fichero, en una carpeta que existe |
| Quién la llama | goreleaser, en el gancho `post` de `universal_binaries` ([release.md §1](./release.md)) | el trabajo `catalogo` de `release.yml` y `make plugin-check` |

**Códigos**: `0` si escribe lo pedido; `1` con cualquier fallo, también con una orden o una bandera que no valen
(research D14). **Salida de error**: en un fallo, una línea que empieza por `empaquetar: ` y nombra lo que falta o lo
que falló; nada si termina bien.

| Fallo | Línea (lo que va detrás de `empaquetar: `) |
|---|---|
| sin orden, o una que no es `piezas` ni `catalogo`; o una invocación de una de las dos que no tiene su forma: una bandera que la orden no tiene, una bandera sin su valor al final, un argumento de más o `-h` | `uso: empaquetar piezas -version … -macos … -windows … -icono … -salida … \| empaquetar catalogo -version … -sha256 … -salida …` |
| una bandera de la orden que no se da, o que llega vacía; con varias, la primera en el orden de la orden | `falta -<bandera>` |
| un binario que no se puede leer | `falta el binario de macOS: <causa, con la ruta>` · `falta el binario de Windows: <causa, con la ruta>` |
| el icono no se puede leer | `falta el icono: <causa, con la ruta>` |
| el icono no es un PNG, o no mide 512 × 512 px | `el icono <ruta> no es un PNG de 512 × 512 px: <lo que es>` |
| la descripción corta pasa de 120 caracteres | `la descripción corta tiene <n> caracteres y el máximo es 120` |
| la salida no se puede escribir | `no se puede escribir <ruta>: <causa>` |
| la huella no tiene su forma | `«<valor>» no es una huella SHA-256: 64 dígitos hexadecimales en minúsculas` |

Una huella vacía es una bandera vacía: `falta -sha256`. Lo que lo empotrado y el registro de producción no dan nunca
—un árbol sin su carpeta `skills`, una ruta que no cabe en un zip, un registro que no se construye— falla igual, con
1 y una línea que lo nombra (`las skills no se pueden leer: …`, `la entrada <ruta> no cabe en el zip: …`, `el registro
de applets no se puede construir: …`).

Ejemplo (medido con el prototipo, que da ya las líneas de «falta»):

```text
$ go run ./cmd/empaquetar piezas -version x -macos dist/kitlegal-universal_darwin_all/kitlegal \
    -windows dist/no-existe/kitlegal.exe -icono mcp/icon.png -salida dist
empaquetar: falta el binario de Windows: open dist/no-existe/kitlegal.exe: no such file or directory
exit status 1
```

Lo que el paso deja en la carpeta de salida cuando falla a medias no se promete: goreleaser falla con él y nada se
publica (research V4). `piezas` compone los dos zips en memoria antes de escribir ninguno, así que sin una entrada no
deja nada; si lo que falla es la escritura del segundo, el primero queda.

## 2. `kitlegal.mcpb`

Cuatro entradas, en este orden ([data-model §2](../data-model.md)). El del snapshot del prototipo, 42 711 814 bytes:

```text
$ unzip -Z dist/kitlegal.mcpb
-rw-r--r--  2.0 unx     2939 bX defN 80-Jan-01 01:00 manifest.json
-rw-r--r--  2.0 unx    20539 bX defN 80-Jan-01 01:00 icon.png
-rwxr-xr-x  2.0 unx 51794802 bX defN 80-Jan-01 01:00 server/kitlegal
-rwxr-xr-x  2.0 unx 26474496 bX defN 80-Jan-01 01:00 server/kitlegal.exe
4 files, 78292776 bytes uncompressed, 42711242 bytes compressed:  45.4%
```

(`unzip` enseña la fecha en la hora del equipo; en el zip es 1980-01-01T00:00:00Z.)

`manifest.json`, entero (2 939 bytes; `tools`, compacto, 1 383):

```json
{
  "manifest_version": "0.3",
  "name": "kitlegal",
  "display_name": "kitlegal",
  "version": "0.3.2-SNAPSHOT-751b76e",
  "description": "Tu asistente de IA responde con la ley vigente del BOE y la cita exacta",
  "long_description": "kitlegal da a Claude herramientas para leer la legislación consolidada del Boletín Oficial del Estado y situar una pregunta en su municipio: el texto vigente de cada artículo, con su norma y su bloque para citarlo. Todo corre en tu equipo: lee fuentes públicas, guarda una caché en ~/.cache/kitlegal/ y no envía tus preguntas a ningún servidor de kitlegal. Para que las respuestas lleven la cita con su forma, instala también el plugin de kitlegal, que trae las skills (kitlegal-plugin.zip).",
  "author": {
    "name": "kitlegal"
  },
  "homepage": "https://kitlegal.es",
  "license": "EUPL-1.2",
  "icon": "icon.png",
  "server": {
    "type": "binary",
    "entry_point": "server/kitlegal",
    "mcp_config": {
      "command": "${__dirname}/server/kitlegal",
      "args": [
        "mcp",
        "serve"
      ],
      "platform_overrides": {
        "win32": {
          "command": "${__dirname}/server/kitlegal.exe"
        }
      }
    }
  },
  "tools": [
    {
      "name": "boe_buscar",
      "description": "Busca normas consolidadas por las palabras de su título o con una consulta de la fuente."
    },
    {
      "name": "boe_indice",
      "description": "Devuelve los bloques de una norma consolidada, en el orden de la fuente."
    },
    {
      "name": "boe_articulo",
      "description": "Devuelve el texto vigente de un bloque de una norma, con los avisos de su vigencia."
    },
    {
      "name": "boe_articulos",
      "description": "Devuelve el texto vigente de varios bloques de una norma, en el orden pedido."
    },
    {
      "name": "boe_metadatos",
      "description": "Devuelve los datos de una norma y los avisos de su vigencia."
    },
    {
      "name": "boe_analisis",
      "description": "Devuelve las materias, las notas y las referencias de una norma."
    },
    {
      "name": "graph_show",
      "description": "Devuelve un nodo del grafo del mundo con sus aristas y su procedencia, sin texto legal."
    },
    {
      "name": "graph_stats",
      "description": "Cuenta los nodos, las aristas y los textos del grafo del mundo por tipo, relación y fuente."
    },
    {
      "name": "graph_check",
      "description": "Comprueba lo consultado de una norma, de algunos de sus bloques o, sin argumentos, todo lo consultado, y lista como mucho 50 hallazgos: redacciones que han cambiado desde la lectura anterior y consultas caducadas."
    },
    {
      "name": "territorio_resolver",
      "description": "Devuelve el territorio de un municipio, por su nombre o por su código INE, con la cobertura de lo que está configurado y verificado."
    }
  ],
  "compatibility": {
    "platforms": [
      "darwin",
      "win32"
    ]
  }
}
```

Cumple el esquema de la versión `0.3` del paquete `mcpb` leído en local (research V13); ese esquema no se versiona
(FR-017).

## 3. `kitlegal-plugin.zip`

`plugin.json` y lo empotrado ([data-model §4](../data-model.md)). El del snapshot del prototipo, 17 739 bytes:

```text
$ unzip -Z dist/kitlegal-plugin.zip
-rw-r--r--  2.0 unx      260 bX defN 80-Jan-01 01:00 .claude-plugin/plugin.json
-rw-r--r--  2.0 unx    23944 bX defN 80-Jan-01 01:00 skills/boe-legislacion/SKILL.md
-rw-r--r--  2.0 unx     3413 bX defN 80-Jan-01 01:00 skills/boe-legislacion/references/normas.md
-rw-r--r--  2.0 unx    12716 bX defN 80-Jan-01 01:00 skills/legal-core/SKILL.md
-rw-r--r--  2.0 unx     1543 bX defN 80-Jan-01 01:00 skills/legal-core/references/jerarquia_normativa.md
-rw-r--r--  2.0 unx     2621 bX defN 80-Jan-01 01:00 skills/legal-core/references/leyes_vertebrales.md
6 files, 44497 bytes uncompressed, 16605 bytes compressed:  62.7%
```

`.claude-plugin/plugin.json`, entero (260 bytes):

```json
{
  "name": "kitlegal",
  "version": "0.3.2-SNAPSHOT-751b76e",
  "description": "Tu asistente de IA responde con la ley vigente del BOE y la cita exacta",
  "author": {
    "name": "kitlegal"
  },
  "homepage": "https://kitlegal.es",
  "license": "EUPL-1.2"
}
```

## 4. El catálogo

Lo que escribe `empaquetar catalogo -version 0.4.0 -sha256 84881b8d…61fcd6 -salida <fichero>` (619 bytes):

```json
{
  "name": "kitlegal-plugins",
  "owner": {
    "name": "kitlegal"
  },
  "plugins": [
    {
      "name": "kitlegal",
      "source": {
        "source": "archive",
        "url": "https://github.com/jmorenobl/kitlegal/releases/download/v0.4.0/kitlegal-plugin.zip",
        "sha256": "84881b8db79242e901b527420fa62b1568c52be48edb321ad3aa85002461fcd6"
      },
      "version": "0.4.0",
      "description": "Tu asistente de IA responde con la ley vigente del BOE y la cita exacta",
      "author": {
        "name": "kitlegal"
      },
      "homepage": "https://kitlegal.es",
      "license": "EUPL-1.2"
    }
  ]
}
```

La dirección es la de la release de esa etiqueta, no la de «la última release» (FR-031). La forma es la del esquema de
`marketplace.json` y de su fuente `archive` en Claude Code 2.1.284 (research V12).

## 5. Reproducibilidad

Los dos zips (research D8): entradas en el orden de §2 y §3, sin entradas de directorio; Deflate de `archive/zip`;
fecha de modificación 1980-01-01T00:00:00Z en todas; modos de [data-model §2 y §4](../data-model.md); sin comentario.
Los tres documentos JSON: campos en el orden de este contrato, sangría de dos espacios, `<`, `>` y `&` sin escapar,
salto de línea final. Dos ejecuciones sobre los mismos binarios y el mismo árbol dan los mismos bytes (FR-004; medido,
research V15).

## 6. Lo que exportan los paquetes

`internal/empaquetado` (nuevo; no lo enlaza `cmd/kitlegal`):

| Símbolo | Qué es | Requisito |
|---|---|---|
| `Ejecutar(args []string, errores io.Writer) int` | atiende la línea de órdenes de §1 y devuelve el código; compone `piezas` con el registro de producción y con lo empotrado; `errores` recibe la línea de un fallo | FR-001, FR-005 |
| `NombreVisible`, `Descripcion`, `DescripcionLarga`, `Autoria` | los textos ([data-model §5](../data-model.md)) | FR-015 |

Nada más se exporta: lo que escribe las piezas con unas herramientas y unas skills dadas es interno, y los tests del
paquete llegan a ello por `export_test.go`, como en `internal/skills`.

`internal/app` gana `HerramientaAnunciada` y `HerramientasAnunciadas(registro *Registro) []HerramientaAnunciada`
([data-model §7](../data-model.md); FR-014). No cambia nada de lo que hay: ni `verbosAnunciados`, ni
`NombresDeHerramientas`, ni lo que el servidor anuncia (FR-080).

`cmd/empaquetar/main.go`: `os.Exit(empaquetado.Ejecutar(os.Args[1:], os.Stderr))`, y nada más.

## 7. Tests de `make ci`

Todos offline, con binarios de prueba (unos bytes cualesquiera) y carpetas de `t.TempDir()`; `mcp/icon.png` se lee del
árbol. Ninguno construye una plataforma ni necesita `dist/`.

| Test | Fichero | Qué fija, y qué ve fallar |
|---|---|---|
| `TestPiezas` | `internal/empaquetado/piezas_test.go` | Con dos herramientas y un árbol de dos skills dados: el `.mcpb` lleva exactamente las cuatro entradas, en orden, con los bytes de los dos binarios y del icono y sus modos; el manifiesto, leído de forma estricta, lleva cada campo de [data-model §3](../data-model.md) con su valor y ninguno más; el plugin lleva `plugin.json`, con los campos de §4 y ninguno más, y cada fichero del árbol, byte a byte, y nada más. Con una herramienta y una skill más en la entrada, aparecen (US5). FR-010 a FR-014, FR-020 a FR-022, FR-070 |
| `TestPiezasReproducibles` | `internal/empaquetado/piezas_test.go` | Dos ejecuciones sobre los mismos binarios, en dos carpetas: los dos `.mcpb` son iguales byte a byte, y los dos plugins también. FR-004, FR-067 |
| `TestPiezasSinEntrada` | `internal/empaquetado/piezas_test.go` | Sin el binario de macOS, sin el de Windows, sin el icono, con una carpeta de salida que no existe y con una carpeta donde va el plugin: error que nombra lo que falta o lo que falló, con su ruta una sola vez y la causa del sistema. También con un árbol de skills sin su carpeta y con una skill cuya ruta no cabe en un zip, que lo empotrado no da nunca; en este último, además, la carpeta de salida queda vacía. FR-005 |
| `TestIcono` | `internal/empaquetado/piezas_test.go` | `mcp/icon.png` del árbol es un PNG de 512 × 512 px y el paso lo acepta; uno de 256 × 256, uno de 512 × 256, uno de 256 × 512 y unos bytes que no son un PNG, creados en el test, lo hacen fallar. FR-016, FR-066 |
| `TestDescripcionCorta` | `internal/empaquetado/textos_test.go` | `Descripcion` tiene 120 caracteres como mucho; una de 120 pasa y una de 121 falla, contando caracteres y no bytes. FR-015, FR-066 |
| `TestCatalogo` | `internal/empaquetado/catalogo_test.go` | Con una versión —la de una etiqueta y la de un snapshot— y una huella: una sola entrada, `kitlegal`, de fuente `archive`, con la dirección de la release de esa versión, la huella, la versión y los textos, leído de forma estricta y con la forma de §5. Sin huella, o con una huella de 63 o de 65 dígitos, en mayúsculas, con un dígito que no es hexadecimal o seguida de un salto de línea, error y ningún documento. FR-030, FR-031 |
| `TestEjecutar` | `internal/empaquetado/ejecutar_test.go` | `piezas`, con la composición de producción y binarios de prueba: código 0, nada en la salida de error, `tools` igual a `app.HerramientasAnunciadas` del registro de producción —que contiene las diez de hoy, escritas en el test— y `skills/` igual a `kitlegal.Skills()`, sin un fichero de más ni de menos. `catalogo`: el documento en el fichero de salida, y solo él en su carpeta, y código 0. Sin orden, con una desconocida, con una bandera que el paso no tiene, sin su valor o vacía, con un argumento de más, pidiendo la ayuda, sin cada una de las banderas de las dos órdenes, sin un binario, con una huella sin su forma y con una salida en una carpeta que no existe: código 1 y una sola línea `empaquetar: …` en la salida de error, la de §1. Con la salida de error rota, el código sigue siendo 1. FR-001, FR-005, FR-014, FR-020, FR-070 |
| `TestHerramientasAnunciadas` | `internal/app/herramientas_test.go` | Con el servidor en proceso y los dos clientes de `mcptest` —el de la especificación vigente y el de la anterior—, sobre los applets de producción y con los de ejemplo añadidos: el nombre y la descripción de cada herramienta que el servidor lista son, como conjunto, los que da `app.HerramientasAnunciadas`. FR-014 |
| `TestElBinarioNoEnlazaElPaso` | `internal/arch_test.go` | El cierre de `./cmd/kitlegal` no contiene `internal/empaquetado`. FR-001 |

`TestHerramientasDelServidor`, `schema-check`, `skills-check` y los guiones de H21 no se tocan y siguen en verde
(FR-080).

## 8. Uso, de fuera adentro

Ninguna salida del paso la consume una skill, y ninguna crece con lo consultado: dependen de la versión, de los verbos
del registro y de las skills empotradas. Con meses de uso —cientos de normas y miles de bloques consultados— miden lo
mismo que el primer día.

| Salida | Quién la pide y cuántas veces | Tamaño (medido) | Cuándo deja de darse su señal |
|---|---|---|---|
| `kitlegal.mcpb` | la persona, una vez por versión; la app lo extrae y arranca el servidor en cada arranque suyo | 42 711 814 bytes (78 292 776 sin comprimir): crece con el binario | no da señales |
| La ficha (`manifest.json`) | la persona, una vez, al instalar | 2 939 bytes: una descripción de 71 caracteres, un párrafo de 491 y diez herramientas (1 383 bytes); crece con los verbos | no da señales; el aviso rojo es de la app |
| `kitlegal-plugin.zip` | la persona, una vez por versión si lo sube; ninguna con el catálogo | 17 739 bytes (44 497 sin comprimir): crece con las skills | no da señales |
| Las skills del plugin | el modelo, una vez por conversación en que se activa cada una | las de hoy, byte a byte: 23 944 y 12 716 bytes de `SKILL.md` | las suyas, sin cambios (H21) |
| Las herramientas de la extensión | la skill, las veces por pregunta de H21: por cada bloque, `boe_articulo` y `graph_check`, más `boe_buscar` o `boe_indice` si hacen falta; `territorio_resolver`, una por municipio | el sobre de cada llamada, acotado por la norma y el bloque pedidos (H21: 4 433 bytes el art. 21 de la Ley 39/2015) | las de H21 y H7.1: cada `version-obsoleta` se apaga con la siguiente lectura del bloque |
| El catálogo | la app o Claude Code de quien lo añadió, al mirar si hay versión nueva | 619 bytes, una entrada; no acumula | la versión nueva se ofrece cuando cambia `version` y deja de ofrecerse al instalarla |
| Dos líneas de `checksums.txt` | `humo` y el trabajo `catalogo`, una vez por release; quien compruebe una descarga | 166 bytes | no dan señales |
| La línea de error del paso | quien mantiene el proyecto, en el registro del snapshot o de la release | una línea, unos 100 bytes | deja de darse cuando la entrada está |
