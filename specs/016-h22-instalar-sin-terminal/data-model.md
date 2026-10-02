# Data model · H22 · Instalar sin terminal

Lo que el hito crea son ficheros, no tablas: dos zips, tres documentos JSON y los textos de los que salen. Ninguno se
guarda en la caché ni en el grafo, y ninguno crece con el uso. Los contratos, con sus ejemplos y sus bytes, están en
[contracts/paso.md](./contracts/paso.md) y [contracts/release.md](./contracts/release.md).

## 1. El paso

El programa `empaquetar` (`cmd/empaquetar`, `internal/empaquetado`). Dos órdenes:

| Orden | Lee | Escribe | Requisitos |
|---|---|---|---|
| `piezas` | la versión; el binario universal de macOS; el binario de Windows `amd64`; el icono; el registro de applets y las skills empotradas, de su propio proceso; los textos | `kitlegal.mcpb` y `kitlegal-plugin.zip`, en la carpeta de salida | FR-001, FR-004, FR-005, FR-010 a FR-022 |
| `catalogo` | la versión; la huella SHA-256 de `kitlegal-plugin.zip`; los textos | el `marketplace.json` de esa versión, en el fichero de salida | FR-030, FR-031 |

No pide nada a la red, no lee nada de `dist/` que no se le nombre y no guarda estado: dos ejecuciones con las mismas
entradas dan los mismos bytes (FR-004).

## 2. `kitlegal.mcpb`

Un zip con cuatro entradas, en este orden, y ninguna más (FR-010):

| Entrada | Contenido | Modo | Requisitos |
|---|---|---|---|
| `manifest.json` | el manifiesto (§3) | `0644` | FR-012 a FR-015, FR-017 |
| `icon.png` | los bytes de `mcp/icon.png` | `0644` | FR-016 |
| `server/kitlegal` | los bytes del binario universal de macOS que recibe | `0755` | FR-011 |
| `server/kitlegal.exe` | los bytes del binario de Windows `amd64` que recibe | `0755` | FR-011 |

El paso copia los dos binarios sin mirarlos: que el de macOS lleve dos arquitecturas y que cada una sea la de su
archivo lo garantiza la configuración de goreleaser y lo comprueba `make snapshot-check` (FR-062).

## 3. El manifiesto

`manifest.json`, de la versión `0.3` del manifiesto de MCP Bundle. Sus campos, en este orden, y ninguno más (FR-012,
FR-017):

| Campo | Valor | De dónde sale |
|---|---|---|
| `manifest_version` | `"0.3"` | fijo |
| `name` | `"kitlegal"` | fijo |
| `display_name` | `NombreVisible` | los textos (§5) |
| `version` | la versión, sin `v` | `-version`, que goreleaser da con `.Version` (FR-013) |
| `description` | `Descripcion`, de 120 caracteres como mucho | los textos (FR-015) |
| `long_description` | `DescripcionLarga` | los textos |
| `author` | `{"name": Autoria}` | los textos |
| `homepage` | `"https://kitlegal.es"` | fijo |
| `license` | `"EUPL-1.2"` | fijo |
| `icon` | `"icon.png"` | fijo |
| `server.type` | `"binary"` | fijo |
| `server.entry_point` | `"server/kitlegal"` | fijo |
| `server.mcp_config.command` | `"${__dirname}/server/kitlegal"` | fijo |
| `server.mcp_config.args` | `["mcp", "serve"]` | fijo |
| `server.mcp_config.platform_overrides` | `{"win32": {"command": "${__dirname}/server/kitlegal.exe"}}` | fijo |
| `tools` | un elemento `{"name", "description"}` por herramienta, en el orden del registro | `app.HerramientasAnunciadas` (FR-014) |
| `compatibility` | `{"platforms": ["darwin", "win32"]}` | fijo |

Sin `user_config`. Reglas: `version` no vacía; `description` de 120 caracteres como mucho (se cuentan caracteres, no
bytes).

## 4. `kitlegal-plugin.zip`

Un zip con `.claude-plugin/plugin.json` y, detrás, cada fichero de lo empotrado en el binario, con su ruta
`skills/<skill>/…`, en el orden en que `fs.WalkDir` lo recorre; todos con modo `0644` (FR-020). Hoy son seis entradas:
`plugin.json` y los cinco ficheros de `boe-legislacion` y `legal-core`. Nada más: ni `bin/`, ni `.mcp.json`, ni ningún
`.mcpb` (FR-022).

`plugin.json`, con estos campos, en este orden, y ninguno más (FR-021; sin `mcpServers`, FR-022):

| Campo | Valor |
|---|---|
| `name` | `"kitlegal"` |
| `version` | la del manifiesto |
| `description` | `Descripcion` |
| `author` | `{"name": Autoria}` |
| `homepage` | `"https://kitlegal.es"` |
| `license` | `"EUPL-1.2"` |

## 5. Los textos

Cuatro constantes exportadas de `internal/empaquetado/textos.go`, el único sitio en el que están escritas (FR-015):

| Constante | Valor | Dónde se usa |
|---|---|---|
| `NombreVisible` | `kitlegal` | `display_name` |
| `Descripcion` | `Tu asistente de IA responde con la ley vigente del BOE y la cita exacta` (71 caracteres) | `description` del manifiesto, de `plugin.json` y de la entrada del catálogo |
| `DescripcionLarga` | el párrafo de abajo (491 caracteres) | `long_description` |
| `Autoria` | `kitlegal` | `author.name` del manifiesto, de `plugin.json` y de la entrada del catálogo; `owner.name` del catálogo |

`DescripcionLarga`:

> kitlegal da a Claude herramientas para leer la legislación consolidada del Boletín Oficial del Estado y situar una
> pregunta en su municipio: el texto vigente de cada artículo, con su norma y su bloque para citarlo. Todo corre en tu
> equipo: lee fuentes públicas, guarda una caché en ~/.cache/kitlegal/ y no envía tus preguntas a ningún servidor de
> kitlegal. Para que las respuestas lleven la cita con su forma, instala también el plugin de kitlegal, que trae las
> skills (kitlegal-plugin.zip).

## 6. El catálogo

`.claude-plugin/marketplace.json` de `jmorenobl/kitlegal-plugins`. La plantilla es
`internal/empaquetado/catalogo.go`: lo fija todo salvo tres valores (FR-030).

| Campo | Valor |
|---|---|
| `name` | `"kitlegal-plugins"` |
| `owner` | `{"name": Autoria}` |
| `plugins` | una sola entrada |
| `plugins[0].name` | `"kitlegal"` |
| `plugins[0].source.source` | `"archive"` |
| `plugins[0].source.url` | **`https://github.com/jmorenobl/kitlegal/releases/download/v<versión>/kitlegal-plugin.zip`** |
| `plugins[0].source.sha256` | **la huella**, 64 dígitos hexadecimales en minúsculas |
| `plugins[0].version` | **la versión**, sin `v` |
| `plugins[0].description` | `Descripcion` |
| `plugins[0].author` | `{"name": Autoria}` |
| `plugins[0].homepage` | `"https://kitlegal.es"` |
| `plugins[0].license` | `"EUPL-1.2"` |

En negrita, los tres valores que cambian con cada etiqueta; la dirección sale de la versión, así que la orden recibe
dos. Reglas: versión no vacía; huella con la forma `^[0-9a-f]{64}$`. Cada etiqueta sustituye el fichero entero: el
catálogo no acumula versiones (FR-032).

## 7. La herramienta anunciada

`app.HerramientaAnunciada{Nombre, Descripcion}`: el nombre `<applet>_<verbo>` y la descripción del verbo con los que el
servidor MCP anuncia una herramienta. `app.HerramientasAnunciadas(registro)` las da en el orden del registro —applets
por nombre, verbos en el orden de su catálogo, sin los de `mcp` ni los de `skills`—: diez con el registro de producción
(FR-014).

## 8. Lo que no cambia

Los seis archivos, los paquetes, `install.sh`, el cask y el bucket; el binario, sus applets, sus verbos, sus esquemas
y sus herramientas; las dos skills; la caché y el grafo; `docs/SOURCES.md`, `testdata/`, `schemas/` y `data/` (FR-006,
FR-080).
