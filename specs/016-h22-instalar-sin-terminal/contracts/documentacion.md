# Contrato: README, CONTRIBUTING y `CHANGELOG.md`

Requisitos: FR-050 a FR-053; SC-012. No hay umbral numérico: lo comprueba la revisión final, apartado a apartado. Este
contrato fija qué dice cada apartado y qué cadenas son literales; la redacción es de la tarea.

## 1. README: tres apartados de instalación

El apartado «Instalar» de hoy se parte en tres, en este orden, delante de «El servidor MCP»:

| Apartado | Qué es |
|---|---|
| `## Instalar sin terminal` | nuevo: la instalación oficial (FR-050) |
| `## Instalar con la terminal` | el «Instalar» de hoy, con otro título: el programa y `kitlegal skills install`, para Claude Code y para Linux. Su texto sobre Codex y Antigravity sale de aquí (§3) |
| `## Otras instalaciones, sin probar` | nuevo (FR-051) |

`CONTRIBUTING.md` enlaza hoy `README.md#instalar`: pasa a `README.md#instalar-con-la-terminal`.

### 1.1 «Instalar sin terminal»: los ocho contenidos de FR-050

| # | Dice | Literales |
|---|---|---|
| 1 | Que es la instalación oficial y dónde está probada: la app de escritorio de Claude en macOS. En Windows los pasos son los mismos y nadie los ha probado | — |
| 2 | Los dos pasos, con lo que la persona ve en cada uno, en texto y sin imágenes. **Uno**: descargar `kitlegal.mcpb` y abrirlo con doble clic; la app enseña una ficha con el nombre, el icono, la descripción y la lista de herramientas, y un botón para instalar. **Dos**: descargar `kitlegal-plugin.zip` y subirlo en *Customize > Plugins > Add > Upload plugin*, estando en el modo de chat y no en Code; o, en lugar del zip, añadir el marketplace `jmorenobl/kitlegal-plugins` | `https://github.com/jmorenobl/kitlegal/releases/latest/download/kitlegal.mcpb` · `https://github.com/jmorenobl/kitlegal/releases/latest/download/kitlegal-plugin.zip` · `jmorenobl/kitlegal-plugins` |
| 3 | Qué dice el aviso rojo de la app al instalar una extensión —que tendrá «acceso a todo lo que hay en tu computadora» y que el desarrollador no está verificado por Anthropic— y qué hace kitlegal de verdad: lee fuentes públicas, escribe solo en `~/.cache/kitlegal/` y no envía nada a ningún servidor propio | `~/.cache/kitlegal/` |
| 4 | Que hacen falta las dos piezas, y qué pasa con una sola: con la extensión sola hay herramientas, pero la respuesta puede no llevar la cita con su forma; con el plugin solo hay skills y ninguna herramienta, y la respuesta es la línea de (8) | `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`, como ejemplo de la cita |
| 5 | Cómo se actualiza cada una: la extensión, descargando el `.mcpb` de la release nueva y abriéndolo; el plugin, subiendo el zip nuevo, o sin hacer nada si se añadió el marketplace, que ofrece la versión nueva cuando cambia. Dicho como lo previsto y **sin probar** | — |
| 6 | Que quien ya tiene las skills con `kitlegal skills install` y además instala el plugin las tiene dos veces en Claude Code | `kitlegal skills install` |
| 7 | Que en Linux no hay app de escritorio donde abrir la extensión, y se usa el binario instalado con Claude Code, como hasta ahora, con un enlace a «Instalar con la terminal» | — |
| 8 | Que en la web y en el móvil no funciona, con la línea que la persona verá | `⚠ SIN CONSULTA AL BOE:` |

Las dos direcciones de descarga son las de «la última release», fijas (FR-050). Nada de (5), ni de si la app admite el
marketplace, se afirma como probado: son los pendientes del ADR 0035 que anota SC-002 (FR-052).

### 1.2 «Otras instalaciones, sin probar» (FR-051)

- **La extensión en Windows**: los mismos dos pasos, con el binario de Windows que el `.mcpb` ya lleva.
- **La app de escritorio de ChatGPT y Codex**, y **Antigravity**: lo que el README dice hoy de cada uno —dónde lee las
  skills cada uno, la regla de Codex para no aprobar cada consulta y cómo se declara el servidor en cada uno—, traído
  de «Instalar» y de «El servidor MCP».
- **Dónde contar si funciona o qué falla**: las incidencias del repositorio
  (`https://github.com/jmorenobl/kitlegal/issues`) o `info@kitlegal.es`.

El README no da por soportado nada que no sea la app de escritorio de Claude en macOS: «oficial» y «soportada» solo se
dicen de ella.

### 1.3 Lo que deja de decir (FR-052)

- El punto «**Claude Cowork y el chat de Claude** todavía no … Llegarán con un plugin» de «Instalar».
- «un plugin para Claude —Claude Code, Cowork y el chat—» en «Y otras formas de usarlo», de «Lo que viene»: quedan los
  paquetes por especialidad y la librería Go.
- En «El servidor MCP», lo que dice de la app de ChatGPT, de Codex y de Antigravity pasa a §1.2; se quedan qué es el
  servidor, que corre en el equipo, cómo se declara en Claude Code y que ChatGPT y Claude en la web y en el móvil no son
  compatibles.

## 2. CONTRIBUTING (FR-053)

En «La release»:

- **El paso que empaqueta**: qué es (`cmd/empaquetar`, `internal/empaquetado`; un programa del repositorio que no viaja
  en el binario); qué lee y qué escribe cada orden ([paso.md §1](./paso.md)); de dónde salen la versión (`.Version` de
  goreleaser), `tools` (el registro, por `app.HerramientasAnunciadas`), las skills (lo empotrado) y los textos
  (`internal/empaquetado/textos.go`); quién lo ejecuta (el gancho de `universal_binaries`); y cómo probarlo en local:
  `go test ./internal/empaquetado/`, y `make release` seguido de `make snapshot-check`.
- **`make release`**: deja además `kitlegal.mcpb` y `kitlegal-plugin.zip`, con su línea en `checksums.txt`, y el
  universal de macOS, que no se publica.
- **`make snapshot-check`**: las seis subpruebas nuevas de `TestSnapshot` ([release.md §3](./release.md)).
- **`make plugin-check`**: qué valida, que necesita Claude Code y que lo ejecuta el trabajo `snapshot`.
- **`publicar`**: atesta también los dos ficheros. **`humo`**: sus seis comprobaciones nuevas. **`catalogo`**: el
  trabajo nuevo, de qué depende y qué escribe.
- **`PUBLISHER_TOKEN`**: escribe en el tap, en el bucket y en `jmorenobl/kitlegal-plugins`; lo ven dos pasos, el de
  goreleaser y el que publica el catálogo; que pueda escribir en el catálogo es de la persona, antes de la primera
  release.

En «Los controles», las filas del snapshot y de su comprobación nombran las dos piezas, y una fila nueva, `make
plugin-check`, fuera de `make ci`. El enlace de §1.

## 3. `CHANGELOG.md`, *Unreleased*

Tres entradas en «Añadido»: (1) `kitlegal.mcpb`, la extensión de escritorio, en cada release, con lo que lleva y con
que está en `checksums.txt` y atestada; (2) `kitlegal-plugin.zip`, el plugin de Claude con las skills, y el catálogo
`jmorenobl/kitlegal-plugins`, que cada etiqueta actualiza; (3) el apartado «Instalar sin terminal» del README. El
binario no cambia y no hay entrada en «Cambiado».

## 4. Lo que no se toca

`docs/SOURCES.md` (no hay fuente nueva), `web/` (la cambia la persona, ADR 0024), los specs anteriores, los ADR y las
dos `SKILL.md`.
