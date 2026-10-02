# Feature Specification: H22 · Instalar sin terminal: la extensión de escritorio con el servidor y el plugin de Claude con las skills

**Feature Branch**: `016-h22-instalar-sin-terminal`

**Created**: 2026-10-02

**Status**: Draft

**Input**: Sección «#### H22 · Instalar sin terminal: la extensión de escritorio con el servidor y el plugin de Claude con las skills (adelantado; ADR 0035)» de `docs/ROADMAP.md`, en modo desatendido (ADR 0018).

## Resumen

Hoy kitlegal se instala con dos órdenes en una terminal, y ninguna de sus dos audiencias —un despacho de una a tres personas, alguien con un asunto— usa una terminal (`docs/USO.md`, 2026-10-01; ADR 0034). H21 dejó el servidor `kitlegal mcp serve` y las dos skills que saben pedirle cada operación como herramienta. Falta que lleguen a la app de escritorio de Claude sin abrir una terminal.

H22 hace que cada release publique dos ficheros más, de la misma etiqueta:

- **`kitlegal.mcpb`**, la extensión de escritorio: lleva el binario dentro y arranca el servidor. Se instala con doble clic.
- **`kitlegal-plugin.zip`**, el plugin de Claude: lleva las skills de esa versión y ningún servidor. Se sube en *Customize > Plugins*, o llega por el marketplace `jmorenobl/kitlegal-plugins`.

Hacen falta las dos (prueba a mano del 2026-10-02, `docs/USO.md` y ADR 0035): con la extensión sola la respuesta trae el texto del BOE sin la cita con su forma; con las dos, `boe-legislacion` se activa y responde con `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`.

H22 no entrega skill nueva ni cambia las que hay: las protege poniéndolas al alcance de sus dos audiencias (constitución, principio VIII). Se da por soportada la app de escritorio de Claude en macOS, que es donde se puede probar; lo demás se entrega o se documenta como «sin probar».

**Cómo se citan los requisitos de otros hitos.** Sin guion: «H19 FR 113», «H21 FR 035». Los ids con guion (FR-001, SC-001…) son siempre de este spec.

**Vocabulario.** *El paso* es el programa del repositorio que escribe los dos ficheros. *El snapshot* es lo que deja `make release` en `dist/`, sin publicar. *La extensión* es `kitlegal.mcpb`; *el plugin*, `kitlegal-plugin.zip`; *el catálogo*, el fichero `.claude-plugin/marketplace.json` del repositorio `jmorenobl/kitlegal-plugins`.

## Clarifications

### Session 2026-10-02

- Q: ¿Contra qué se compara `server/kitlegal`, el binario universal de macOS del `.mcpb`, en `make snapshot-check` y en `humo`: contra un archivo nuevo de la release que lleve ese binario universal, o contra los seis archivos de hoy, sin publicar ninguno más? (FR-006) → A: Los archivos de la release siguen siendo los seis de hoy y no se publica ninguno más: el universal solo viaja dentro del `.mcpb`. `make snapshot-check` y `humo` leen de `server/kitlegal` sus arquitecturas, fallan si no son exactamente dos —una `amd64` y una `arm64`, reconocidas por su tipo de CPU y no por su posición— y comparan cada una con el binario de `kitlegal_darwin_amd64.tar.gz` y de `kitlegal_darwin_arm64.tar.gz`: byte a byte en `make snapshot-check`, por su huella en `humo`. `server/kitlegal.exe` se compara con el de `kitlegal_windows_amd64.zip`. En `.goreleaser.yaml`, `universal_binaries` con `replace: false` y un `id` propio, y `archives` acotado con `ids`, para que no salga un archivo `darwin_all`. (auto: conservadora; criterio d; fuente: `docs/ROADMAP.md`, H22: Entrega, Alcance y trabajo `humo`; spec.md, «Relación con H19 y H21»; constitución, «Criterio de decisión autónoma», puntos 2 y 4; goreleaser v2.18.1 de `tools/goreleaser/go.mod`, `internal/pipe/universalbinary` e `internal/pipe/archive`)
- Q: ¿Cómo llega a `testdata/` el esquema JSON de la versión `0.3` del manifiesto de MCP Bundle dentro del run, y qué comprueba el control del manifiesto si no puede llegar? (FR-017) → A: No llega: el run ni lo versiona ni lo sustituye por uno propio, no hay tarea `[datos]`, `testdata/` no se toca y `docs/SOURCES.md` queda sin cambios. El control de FR-061 usa una comprobación propia del repositorio: versión `0.3`, cada campo de FR-012 con su valor y ninguno más, `version` igual a la del binario que lleva dentro y `tools` igual a lo que ese binario anuncia por MCP. Versionar el esquema oficial y validar contra él queda como pendiente de la persona, fuera del run. (auto: conservadora; criterio d; fuente: `docs/ROADMAP.md`, H22: Controles y Documentación; constitución, «Gates» capa 3, «Reglas del modo desatendido» y «Criterio de decisión autónoma», puntos 2 y 4; `scripts/workflow/grabar-datos.sh`; `docs/SOURCES.md`, sin fila para `github.com/modelcontextprotocol/mcpb`; `docs/WORKFLOW.md`)
- Q: ¿El catálogo de `jmorenobl/kitlegal-plugins` pasa a apuntar a la etiqueta nueva antes de que `humo` compruebe lo publicado, o solo después de que `humo` salga en verde? (FR-031) → A: Después de `humo`, y solo si sale en verde: `release.yml` pasa a tener tres trabajos, `publicar`, `humo` y uno nuevo que depende de `humo` y escribe el catálogo, sin permiso de escritura sobre este repositorio. `PUBLISHER_TOKEN` lo ven dos pasos: el de goreleaser en `publicar`, como hoy, y el que escribe el catálogo. Si `humo` falla, el catálogo sigue apuntando a la etiqueta anterior; si el paso del catálogo no puede escribir, falla su trabajo, sin afectar a `publicar` ni a `humo`, y se puede volver a ejecutar solo. (auto: criterio c; fuente: `docs/ROADMAP.md`, H22, «El marketplace»; spec.md, FR-031 y «Assumptions»; `.github/workflows/release.yml`; constitución, «Criterio de decisión autónoma», punto 1)

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H22. El Objetivo y el Alcance se reflejan en los requisitos. Dos controles se cumplen como dicen las Clarifications y no al pie de la letra: «el universal de macOS … de los archivos» se compara arquitectura a arquitectura con los dos archivos de macOS, sin archivo nuevo (FR-006, FR-062), y «cumple el esquema de la versión `0.3`» se comprueba con una comprobación propia del repositorio mientras la persona no versione el esquema (FR-017, FR-061).

- **Entrega**: «cada release publica, de la misma etiqueta y junto a los archivos de hoy, (1) `kitlegal.mcpb`, un MCP Bundle con el binario dentro —el universal de macOS y el de Windows, este sin probar—, que arranca `kitlegal mcp serve` y se instala con doble clic como extensión de escritorio; y (2) `kitlegal-plugin.zip`, el plugin de Claude con las skills de esa versión y sin servidor, que se instala en *Customize > Plugins > Add > Upload plugin* estando en el modo de chat. Con las dos, cualquier conversación de la app responde con las herramientas y con el protocolo de las skills. Con el plugin y sin la extensión, y en la web y en el móvil, responde la regla `⚠ SIN CONSULTA AL BOE:` de H21. Con la extensión y sin el plugin quedan las herramientas y las `instructions` del servidor, que no bastan para la forma de la cita: el README dice que se instalan las dos.»
- **Controles**:
  - «el snapshot deja `kitlegal.mcpb` y `kitlegal-plugin.zip`, y los dos están en `checksums.txt`;»
  - «el manifiesto del `.mcpb` del snapshot cumple el esquema de la versión `0.3` —el esquema se versiona en `testdata/`, de la fuente del ADR 0035—, su `version` es la del binario que lleva dentro, y su `tools` es exactamente el conjunto de herramientas que ese binario anuncia por MCP, con sus descripciones;»
  - «los dos binarios del `.mcpb` son, byte a byte, el universal de macOS y el de Windows `amd64` de los archivos del mismo snapshot, y el `.mcpb` no lleva ningún otro ejecutable;»
  - «un cliente MCP arranca el servidor con el `mcp_config` del manifiesto, desde el `.mcpb` extraído en una ruta con espacios y con `/` como directorio de trabajo, y lista las herramientas de H21 (en el runner, con el binario de su plataforma en el sitio del de macOS: lo que se prueba es el manifiesto);»
  - «ningún fichero de `skills/` del plugin difiere de lo empotrado en el binario de ese snapshot, y no hay ninguno de más; el plugin no lleva `bin/`, ni `.mcp.json`, ni `mcpServers`, ni ningún `.mcpb`;»
  - «`claude plugin validate` sobre el plugin extraído y sobre la plantilla del marketplace, en el trabajo `snapshot` de la CI, con Claude Code en la versión que fija `evals.yml`; no entra en `make ci`, que no depende de Claude Code;»
  - «la descripción corta no pasa de 120 caracteres y el icono del `.mcpb` es `mcp/icon.png`, un PNG cuadrado de 512 px;»
  - «dos ejecuciones del paso sobre los mismos binarios dan los mismos dos ficheros, byte a byte (las huellas de `checksums.txt` no cambian sin que cambie el contenido);»
  - «`goreleaser check`, `make ci`.»
- **Aceptación**: «en el run, el snapshot de la propuesta de cambio deja los dos ficheros con sus huellas y los controles de arriba pasan. Después de fusionar y de publicar una release, y fuera del run porque es humano: en un Mac distinto del que compiló, sin kitlegal instalado y sin abrir una terminal, con `kitlegal.mcpb` y `kitlegal-plugin.zip` descargados de la release con un navegador e instalados, una conversación nueva de la app de Claude, sin carpeta, responde «¿qué dice el art. 21 de la Ley 39/2015?» con `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`. En esa misma prueba se anota lo que el ADR 0035 dejó pendiente: si macOS bloquea el binario descargado, cómo llega una versión nueva de la extensión y del plugin, si la app admite el marketplace, y qué responde Claude en la web con el plugin y sin la extensión.»

Trazabilidad resumida (el detalle, en cada requisito):

| Criterio del hito | Dónde se cumple |
|---|---|
| Un paso propio que empaqueta, en Go y sin Node | FR-001 a FR-006, US3 |
| `kitlegal.mcpb`; nada del manifiesto a mano dos veces; el icono | FR-010 a FR-017, US1, US5 |
| `kitlegal-plugin.zip` | FR-020 a FR-023, US1 |
| El marketplace | FR-030 a FR-032, US2 |
| `release.yml`: publicar, atestar y `humo` | FR-003, FR-031, FR-040 a FR-043, US3 |
| README y documentación | FR-050 a FR-053, US4 |
| Lo que se da por soportado | FR-050, FR-051, «Fuera de alcance» |
| Controles | FR-060 a FR-070, SC-003 a SC-012 |
| Aceptación | SC-001 (en el run), SC-002 (humana, tras publicar una release) |

## Relación con H19 y H21

No se edita ningún spec anterior: son el registro de sus runs (FR-080). No cambia ninguna decisión de arquitectura que no haya cambiado ya el ADR 0035, así que no hay ADR nuevo.

**Sustituye** (lo que dice el hito citado deja de valer, y vale lo de este spec):

- De H19, lo que la release adjunta y lo que lista `checksums.txt` además de los archivos: era solo `install.sh`; ahora son `install.sh`, `kitlegal.mcpb` y `kitlegal-plugin.zip` (FR-002). Lo que atesta `publicar` (H19 FR 112: los seis archivos y `checksums.txt`) gana los dos ficheros (FR-003). Lo que comprueba `humo` (H19 FR 113) gana lo de FR-040 a FR-042. Lo que escribe `PUBLISHER_TOKEN` (el tap y el bucket) gana el catálogo (FR-031), y lo que H19 FR 114 decía de ese secreto, que solo lo ve el trabajo de publicación, pasa a ser que lo ven solo los dos pasos que publican: el de goreleaser en `publicar` y el del catálogo, en un trabajo nuevo que depende de `humo`.
- Del README de H19 y H21: «Claude Cowork y el chat de Claude todavía no … Llegarán con un plugin» deja de ser cierto (FR-052), y lo que dice de la app de ChatGPT, de Codex y de Antigravity pasa al apartado «Otras instalaciones, sin probar» (FR-051).

**Se queda**: los seis archivos de hoy, sus nombres y su contenido, los paquetes `.deb` y `.rpm`, `install.sh`, el cask y el bucket; `kitlegal skills install`, `list` y `doctor`; el servidor de H21, sus diez herramientas, sus esquemas y sus `instructions`; las dos skills, byte a byte; el job de evals y sus umbrales; y que etiquetar y publicar son humanos (ADR 0020).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Quien no usa una terminal instala kitlegal con dos ficheros y pregunta (Priority: P1)

Una persona con la app de escritorio de Claude en un Mac, sin kitlegal y sin abrir una terminal, descarga de la última release `kitlegal.mcpb` y `kitlegal-plugin.zip` con su navegador. Abre el primero con doble clic: la app enseña una ficha con el nombre, el icono de kitlegal, una descripción que cabe en la cabecera y la lista de herramientas, y la instala. Sube el segundo en *Customize > Plugins*, en el modo de chat. En una conversación nueva, sin carpeta, pregunta «¿qué dice el art. 21 de la Ley 39/2015?» y la respuesta lleva `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`.

**Why this priority**: es el objetivo del hito y la mayor distancia entre la web y quien la lee.

**Independent Test**: en el run, las comprobaciones del snapshot sobre los dos ficheros (SC-003 a SC-010); tras publicar una release, la prueba humana (SC-002).

**Acceptance Scenarios**:

1. **Dado** el snapshot de la propuesta de cambio, **Cuando** se mira `dist/`, **Entonces** están `kitlegal.mcpb` y `kitlegal-plugin.zip`, y `checksums.txt` lleva una línea con la huella de cada uno (FR-002).
2. **Dado** el `.mcpb` del snapshot, **Cuando** se lee su `manifest.json`, **Entonces** se declara de la versión `0.3` y lleva cada campo de FR-012 con su valor y ninguno más, su `version` es la que imprime el binario que lleva dentro, sin la `v` inicial si la tiene, y su `tools` lleva exactamente los nombres y las descripciones de las herramientas que ese binario anuncia por MCP (FR-012 a FR-014, FR-061).
3. **Dado** el `.mcpb` extraído en una ruta con espacios, con el binario de la plataforma del runner en `server/kitlegal`, **Cuando** un cliente MCP arranca el servidor con la orden y los argumentos del `mcp_config` del manifiesto, resuelto `${__dirname}` al directorio de extracción y con `/` como directorio de trabajo, **Entonces** lista las diez herramientas de H21 (FR-012, FR-064).
4. **Dado** el `.mcpb` del snapshot, **Cuando** se separan las dos arquitecturas de `server/kitlegal` y se comparan, con `server/kitlegal.exe`, con los binarios de los archivos de macOS (`amd64` y `arm64`) y de Windows `amd64` del mismo snapshot, **Entonces** cada una es igual byte a byte a la suya, `server/kitlegal` está marcado como ejecutable y el `.mcpb` no lleva ningún otro ejecutable (FR-010, FR-011).
5. **Dado** el plugin del snapshot, **Cuando** se compara su `skills/` con lo que el binario de ese snapshot lleva empotrado, **Entonces** ningún fichero difiere, no falta ninguno y no hay ninguno de más; y el plugin no lleva `bin/`, ni `.mcp.json`, ni `mcpServers` en su `plugin.json`, ni ningún `.mcpb` (FR-020 a FR-022).
6. **Dado** el `.mcpb` del snapshot, **Cuando** se lee `description` del manifiesto y su `icon.png`, **Entonces** la descripción tiene 120 caracteres o menos y el icono es, byte a byte, `mcp/icon.png`, un PNG de 512 × 512 px (FR-015, FR-016).
7. **Dado** una release publicada, y fuera del run, **Cuando** una persona hace lo que cuenta esta historia en un Mac distinto del que compiló, **Entonces** la respuesta lleva `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]` (SC-002).

---

### User Story 2 - Quien añade el marketplace recibe el plugin de cada versión sin volver a subir un zip (Priority: P2)

Una persona añade `jmorenobl/kitlegal-plugins` como marketplace en lugar de subir el zip. El catálogo tiene una entrada, `kitlegal`, que apunta al `kitlegal-plugin.zip` de la release de una etiqueta, con su huella. Cuando se publica otra etiqueta, el catálogo pasa a apuntar a la nueva y cambia `version`, y quien lo añadió recibe la versión nueva.

**Why this priority**: evita repetir el paso del plugin en cada versión, pero la instalación funciona sin él, subiendo el zip.

**Independent Test**: en el run, `claude plugin validate` sobre la plantilla del catálogo, en el trabajo `snapshot` (SC-008), y la comprobación de la definición de la release (SC-011). Que la app admita un catálogo con una entrada de fuente `archive` lo anota la prueba humana (SC-002).

**Acceptance Scenarios**:

1. **Dado** la plantilla del catálogo del repositorio, rellenada con una versión, una dirección y una huella, **Cuando** se valida con `claude plugin validate` en el trabajo `snapshot`, **Entonces** es válida, y tiene una sola entrada, `kitlegal`, de fuente `archive` (FR-030, FR-065).
2. **Dado** una etiqueta `vX.Y.Z` publicada, **Cuando** el flujo `release` termina con `humo` en verde, **Entonces** `.claude-plugin/marketplace.json` de `jmorenobl/kitlegal-plugins` lleva en su entrada `kitlegal` la `version` `X.Y.Z`, la dirección de `kitlegal-plugin.zip` en la release de esa etiqueta y el `sha256` que `checksums.txt` da para ese fichero (FR-031).
3. **Dado** un `PUBLISHER_TOKEN` que no puede escribir en `jmorenobl/kitlegal-plugins`, **Cuando** el flujo `release` llega al paso que publica el catálogo, **Entonces** ese paso falla y su trabajo sale en rojo (FR-031).

---

### User Story 3 - Quien mantiene kitlegal publica las dos piezas con cada etiqueta, sin pasos a mano, y las prueba en local (Priority: P1)

Quien mantiene el proyecto no hace nada nuevo para publicar: empuja la etiqueta, como hoy. El flujo construye los dos ficheros a partir de lo que goreleaser ya ha compilado, los adjunta a la release, los lista en `checksums.txt`, que va firmado, los atesta y los comprueba desde fuera. En local y en cada propuesta de cambio, el snapshot los deja en `dist/` y `make snapshot-check` los comprueba.

**Why this priority**: sin esto las dos piezas no existen, o existen hechas a mano y sin huella, como las de la prueba.

**Independent Test**: `make release` y `make snapshot-check` en el trabajo `snapshot` de la CI (SC-001); `make ci`, con `goreleaser check` (SC-011).

**Acceptance Scenarios**:

1. **Dado** el árbol del hito, **Cuando** se ejecuta `make release`, **Entonces** goreleaser ejecuta el paso y `dist/` tiene los dos ficheros, con los nombres `kitlegal.mcpb` y `kitlegal-plugin.zip`, sin versión en el nombre, y con su línea en `checksums.txt` (FR-001, FR-002).
2. **Dado** los mismos binarios compilados, **Cuando** el paso se ejecuta dos veces, **Entonces** los dos `kitlegal.mcpb` son iguales byte a byte, y los dos `kitlegal-plugin.zip` también (FR-004).
3. **Dado** que falta una de las entradas del paso —uno de los binarios que empaqueta o el icono—, **Cuando** se ejecuta, **Entonces** termina con un código distinto de 0 y nombra en la salida de error lo que falta, y el snapshot o la release que lo ejecutaba falla (FR-005).
4. **Dado** el árbol del hito, **Cuando** se ejecuta `make ci`, **Entonces** queda en verde, con `goreleaser check` incluido y con la comprobación estricta de la definición de la release al día de lo que el hito añade (FR-068).
5. **Dado** una release publicada, **Cuando** corre el trabajo `humo`, **Entonces** descarga los dos ficheros, comprueba su huella contra `checksums.txt` y su atestación, comprueba el manifiesto y los dos binarios del `.mcpb`, y arranca el servidor desde el binario de Linux de la release con un cliente MCP, que lista las herramientas (FR-040 a FR-042).
6. **Dado** el binario distribuido tras el hito, **Cuando** se listan sus applets y sus verbos, **Entonces** son los de antes: el paso no es un applet ni viaja en el binario (FR-001).

---

### User Story 4 - Quien lee el README sabe qué instalar, qué significa el aviso rojo y qué no está probado (Priority: P2)

**Why this priority**: el aviso rojo de la app para a quien no es técnico (`docs/USO.md`, 2026-10-02), y sin el README nadie sabe que hacen falta las dos piezas. Es documentación.

**Independent Test**: lectura del README, de CONTRIBUTING y de `CHANGELOG.md` contra FR-050 a FR-053 (SC-012).

**Acceptance Scenarios**:

1. **Dado** el README, **Cuando** se lee «Instalar sin terminal», **Entonces** dice que es la instalación oficial, que está probada en la app de escritorio de Claude en macOS y que en Windows los pasos son los mismos y nadie los ha probado, y da los dos pasos con lo que la persona ve en cada uno (FR-050).
2. **Dado** el README, **Cuando** se busca el aviso de la app, **Entonces** dice lo que dice el aviso y lo que hace kitlegal de verdad: lee fuentes públicas, escribe solo en `~/.cache/kitlegal/` y no envía nada a ningún servidor propio (FR-050).
3. **Dado** el README, **Cuando** se lee «Otras instalaciones, sin probar», **Entonces** están la extensión en Windows, la app de ChatGPT, Codex y Antigravity, y dónde contar si funciona o qué falla: las incidencias del repositorio o `info@kitlegal.es` (FR-051).

---

### User Story 5 - Un verbo o una skill nuevos llegan a la extensión y al plugin sin tocar el paso (Priority: P3)

Quien añade un verbo al registro o una skill a `skills/` en un hito posterior no escribe nada en el paso ni en el manifiesto: la release siguiente lleva la herramienta en `tools` y la skill en el plugin.

**Why this priority**: es lo que evita que el manifiesto y el plugin se queden atrás; no cambia nada para quien instala hoy.

**Independent Test**: las comprobaciones de `tools` y de `skills/` contra el binario del snapshot (SC-004, SC-007).

**Acceptance Scenarios**:

1. **Dado** un registro con un verbo de consulta más, **Cuando** se construye el snapshot, **Entonces** `tools` del manifiesto lleva su herramienta, con su descripción, sin ningún cambio en el paso (FR-014).
2. **Dado** una skill más empotrada en el binario, **Cuando** se construye el snapshot, **Entonces** el plugin la lleva, sin ningún cambio en el paso (FR-020).

---

### Edge Cases

- **Solo la extensión**: hay herramientas y las `instructions` del servidor, y la respuesta puede no llevar la cita con su forma (prueba del 2026-10-02). No se arregla en este hito: el README dice que se instalan las dos (FR-050).
- **Solo el plugin, y la web y el móvil**: hay skill y no hay herramienta; responde la línea `⚠ SIN CONSULTA AL BOE:` de H21 (H21 FR 035). El README lo dice, con la línea (FR-050).
- **El plugin y, además, `kitlegal skills install`**: en Claude Code las skills están dos veces. El hito no lo evita: el README lo dice (FR-050).
- **La versión en un snapshot**: no hay etiqueta. `version` del manifiesto y del `plugin.json` es la que imprime el binario del snapshot (FR-013).
- **Windows `arm64` y Linux**: no van en el `.mcpb`. `platform_overrides` distingue el sistema y no la arquitectura (ADR 0035), y el hito pide el de `amd64`; en Linux no hay app de escritorio.
- **Una descripción corta de más de 120 caracteres**: no llega a ninguna release, porque `make ci` falla antes (FR-015, FR-066).
- **La app arranca el servidor desde un directorio que cambia en cada arranque y con `/` como directorio de trabajo** (ADR 0035): el servidor no depende de ninguno de los dos (H21 FR 025), y el control de FR-064 lo ejerce con el `mcp_config` del manifiesto.
- **Lo que llega manipulado** —un `dist/` tocado a mano entre goreleaser y el paso, un zip alterado tras publicarse—: el paso falla o la huella no coincide; este spec no promete nada más ni lo especifica caso a caso.

## Requirements *(mandatory)*

### Functional Requirements

#### El paso que empaqueta

- **FR-001**: Un programa del repositorio, en Go y sin Node, MUST escribir `kitlegal.mcpb` y `kitlegal-plugin.zip` a partir de lo que goreleaser ya ha compilado. MUST NOT ser un applet ni entrar en el binario distribuido: los applets, los verbos y `--describe` del binario no cambian. MUST NOT pedir nada a la red. Comprobable: `make schema-check` sin drift y el control de conformidad de H21, que fija el conjunto de herramientas, siguen en verde en `make ci` (FR-068).
- **FR-002**: goreleaser MUST ejecutar el paso en la release y en el snapshot, y MUST adjuntar su salida con `release.extra_files` y `checksum.extra_files`, como hoy `install.sh`: los dos ficheros quedan en la release, y en `checksums.txt`, que va firmado. Sus nombres MUST ser `kitlegal.mcpb` y `kitlegal-plugin.zip`, sin versión, para que «la última release» sea una dirección fija. Comprobable: FR-060.
- **FR-003**: El trabajo `publicar` de `release.yml` MUST atestar la procedencia de los dos ficheros, junto a lo que ya atesta. Comprobable: en el run, la comprobación de la definición de la release (FR-068); en cada release, `humo` (FR-040).
- **FR-004**: El paso MUST ser reproducible: dos ejecuciones sobre los mismos binarios y el mismo árbol dan los mismos dos ficheros, byte a byte. Comprobable: FR-067.
- **FR-005**: Si falta una entrada del paso —uno de los binarios que empaqueta, el icono— o no puede escribir su salida, MUST terminar con un código distinto de 0 y decir en la salida de error qué falta o qué falló. El snapshot o la release que lo ejecuta falla con él. Comprobable: un test de `make ci` lo ejerce sin uno de los binarios.
- **FR-006**: Los seis archivos de hoy, los paquetes `.deb` y `.rpm` e `install.sh` MUST seguir publicándose con el mismo nombre y el mismo contenido, y la release MUST NOT publicar ningún archivo más con el binario universal de macOS: solo viaja dentro del `.mcpb`. El binario universal MUST construirse con goreleaser (`universal_binaries`, con `replace: false` y un `id` propio) a partir de los dos de macOS que ya compila, y `archives` MUST acotarse con `ids` a la construcción de hoy, para que el universal no dé un archivo `darwin_all`. `TestSnapshot` sigue exigiendo exactamente seis archivos. Comprobable: FR-062, FR-068 y `TestSnapshot`.

#### `kitlegal.mcpb`

- **FR-010**: `kitlegal.mcpb` MUST ser un zip con cuatro entradas y ninguna más: `manifest.json`, `icon.png`, `server/kitlegal` y `server/kitlegal.exe`. MUST NOT llevar ningún otro ejecutable, ningún binario de Linux ni ninguna skill. Comprobable: FR-062.
- **FR-011**: `server/kitlegal` MUST ser el binario universal de macOS de esa construcción, con exactamente dos arquitecturas, y cada una MUST ser, byte a byte, el binario del archivo de macOS de esa arquitectura (`kitlegal_darwin_amd64.tar.gz`, `kitlegal_darwin_arm64.tar.gz`); MUST ir marcado como ejecutable en el zip. `server/kitlegal.exe` MUST ser, byte a byte, el de `kitlegal_windows_amd64.zip`, que la release ya publica. El de Windows va sin probar. Comprobable: FR-062.
- **FR-012**: `manifest.json` MUST ser de la versión `0.3` del manifiesto y llevar: `name` `kitlegal`; `display_name`; `version`; `description`; `long_description`; `author`; `homepage` `https://kitlegal.es`; `license` `EUPL-1.2`; `icon` `icon.png`; `server` con `type` `binary`, `entry_point` `server/kitlegal` y `mcp_config` con la orden `${__dirname}/server/kitlegal` y los argumentos `mcp` y `serve`, y con `platform_overrides.win32` que cambia la orden a la de `server/kitlegal.exe`; `compatibility.platforms` con `darwin` y `win32`, y ninguna más; y `tools`. MUST NOT llevar `user_config`: no hay nada que configurar. Comprobable: FR-061 y FR-064.
- **FR-013**: `version` del manifiesto MUST ser la etiqueta sin la `v` inicial. En un snapshot, que no tiene etiqueta, es la versión del snapshot. En los dos casos MUST coincidir con lo que imprime `version` el binario que el `.mcpb` lleva dentro, sin su `v` inicial si la tiene. MUST NOT escribirse a mano en ningún fichero versionado. Comprobable: FR-061.
- **FR-014**: `tools` del manifiesto MUST salir del registro de applets, como las herramientas del servidor (H21 FR 002): un elemento por herramienta, con su nombre y su descripción, y ninguna lista escrita a mano. MUST ser exactamente el conjunto de herramientas que el binario de esa construcción anuncia por MCP, con las mismas descripciones: diez con el registro de hoy. Comprobable: FR-061.
- **FR-015**: Los textos del manifiesto y del `plugin.json` —`display_name`, `description`, `long_description`, `author`— MUST vivir en un solo sitio del repositorio, del que los leen los dos. `description` del manifiesto MUST tener 120 caracteres como mucho. Comprobable: un test de `make ci` falla si la descripción corta del repositorio pasa de 120 caracteres, y `make snapshot-check` falla si la del manifiesto del snapshot pasa (FR-066).
- **FR-016**: `icon.png` del `.mcpb` MUST ser una copia byte a byte de `mcp/icon.png`, que ya está versionado: un PNG cuadrado de 512 px. El paso lo copia y MUST NOT generarlo. Comprobable: un test de `make ci` falla si `mcp/icon.png` no es un PNG de 512 × 512 px, y `make snapshot-check` falla si el del `.mcpb` no es ese fichero (FR-066).
- **FR-017**: El manifiesto MUST ser de la versión `0.3` del manifiesto de MCP Bundle y MUST llevar solo lo que FR-012 enumera. El esquema JSON oficial no se versiona ni se sustituye por uno propio en este hito: no llega a `testdata/` dentro del run (no hay red y `docs/SOURCES.md` no tiene fila para su fuente), y validar contra él es un pendiente de la persona (véase «Fuera de alcance»). Comprobable: FR-061, con una comprobación propia del repositorio y no contra el esquema oficial.

#### `kitlegal-plugin.zip`

- **FR-020**: `kitlegal-plugin.zip` MUST llevar `.claude-plugin/plugin.json` y `skills/`, y nada más. `skills/` MUST llevar cada skill empotrada en el binario de esa construcción —hoy `boe-legislacion` y `legal-core`—, byte a byte, con su `SKILL.md` y su `references/`: ningún fichero distinto, ninguno de menos y ninguno de más. Una skill nueva empotrada MUST aparecer sin cambiar el paso. Comprobable: FR-063.
- **FR-021**: `plugin.json` MUST llevar `name` `kitlegal`, `version`, `description`, `author`, `homepage` y `license`, con la misma versión que el manifiesto (FR-013), los textos del mismo sitio (FR-015), `homepage` `https://kitlegal.es` y `license` `EUPL-1.2`. Comprobable: FR-063 y FR-065.
- **FR-022**: El plugin MUST NOT llevar `mcpServers` en `plugin.json`, ni `.mcp.json`, ni ningún `.mcpb`, ni `bin/` en la raíz: dentro de un plugin las herramientas solo llegan a las conversaciones con carpeta, con las dos piezas el servidor estaría dos veces, y un `bin/` impide instalarlo (ADR 0035). Comprobable: FR-063.
- **FR-023**: El plugin extraído MUST pasar `claude plugin validate`. Comprobable: FR-065.

#### El marketplace

- **FR-030**: El repositorio MUST llevar la plantilla de `.claude-plugin/marketplace.json` del catálogo: una sola entrada, `kitlegal`, de fuente `archive`, con tres valores que se rellenan en cada etiqueta —`version`, la dirección de `kitlegal-plugin.zip` en la release de esa etiqueta y su `sha256`— y los textos de FR-015. La plantilla, rellenada, MUST pasar `claude plugin validate`. Comprobable: FR-065.
- **FR-031**: En cada etiqueta, `release.yml` MUST escribir en `jmorenobl/kitlegal-plugins` el `.claude-plugin/marketplace.json` de esa etiqueta, con `PUBLISHER_TOKEN`, en un trabajo nuevo que depende de `humo` y solo se ejecuta si `humo` sale en verde: `version` es la etiqueta sin la `v`; la dirección es la de `kitlegal-plugin.zip` en la release de esa etiqueta, no la de «la última release»; y `sha256` es la huella que `checksums.txt` da para ese fichero. `release.yml` pasa a tener tres trabajos —`publicar`, `humo` y el del catálogo—. `PUBLISHER_TOKEN` MUST verlo solo el paso de goreleaser en `publicar`, como hoy, y el paso que escribe el catálogo; el trabajo nuevo MUST NOT tener permiso de escritura sobre este repositorio, y `humo` sigue sin secretos. Si `humo` falla, el catálogo MUST seguir apuntando a la etiqueta anterior. Si el paso no puede escribir, MUST fallar y su trabajo salir en rojo, sin afectar a `publicar` ni a `humo`, y se puede volver a ejecutar solo. Comprobable: en el run, la comprobación de la definición de la release (FR-068) y la validación de la plantilla (FR-065); que publica de verdad solo se ve en la primera release, que es humana.
- **FR-032**: El catálogo MUST NOT guardar ningún fichero de release: `kitlegal-plugin.zip` sigue en las releases de este repositorio, con su firma. Que `PUBLISHER_TOKEN` pueda escribir en `jmorenobl/kitlegal-plugins` es de la persona, antes de la primera release, y no es una tarea del hito.

#### `release.yml`, trabajo `humo`

- **FR-040**: `humo` MUST descargar `kitlegal.mcpb` y `kitlegal-plugin.zip` de la release publicada, comprobar la huella de cada uno contra `checksums.txt` y verificar su atestación de procedencia, como hace hoy con el archivo linux/amd64. Cada comprobación que falla MUST poner el trabajo en rojo.
- **FR-041**: `humo` MUST comprobar del `.mcpb` descargado lo que se puede sin ejecutarlo, porque no lleva binario de Linux: que su manifiesto lleva como `version` la etiqueta sin la `v` y los campos fijos de FR-012, y que sus binarios son, por su huella, los de los archivos de esa release: antes comprueba contra `checksums.txt` la huella de `kitlegal_darwin_amd64.tar.gz`, `kitlegal_darwin_arm64.tar.gz` y `kitlegal_windows_amd64.zip`; lee de `server/kitlegal` sus arquitecturas, falla si no son exactamente dos —una `amd64` y una `arm64`, reconocidas por su tipo de CPU y no por su posición— y compara cada una con el binario de su archivo de macOS; y compara `server/kitlegal.exe` con el de `kitlegal_windows_amd64.zip`. `humo` no tiene el código del repositorio: separa las arquitecturas leyendo la cabecera que escribe goreleaser, enteros de 32 bits en big-endian —la magia, el número de arquitecturas y, por cada una, CPU, subtipo, desplazamiento, tamaño y alineación—, con cada binario copiado tal cual en su desplazamiento. Esa separación con órdenes de shell se midió en el plan, en macOS y contra el snapshot del prototipo (research V20); en `ubuntu-latest`, donde corre `humo`, no se ha medido (research S1).
- **FR-042**: `humo` MUST arrancar `kitlegal mcp serve` desde el binario de Linux de la release con un cliente MCP y listar las herramientas, y MUST fallar si no las lista o si sus nombres no son los de `tools` del manifiesto del `.mcpb` descargado.
- **FR-043**: `humo` MUST seguir sin obtener el código del repositorio y con sus permisos de hoy, de solo lectura. Lo que hoy comprueba MUST seguir comprobándolo.

#### README y documentación

- **FR-050**: El README MUST tener un apartado «Instalar sin terminal» que diga: (1) que es la instalación oficial y dónde está probada —la app de escritorio de Claude en macOS; en Windows los pasos son los mismos y nadie los ha probado—; (2) los dos pasos, con lo que la persona ve en cada uno descrito en texto: descargar `kitlegal.mcpb` y abrirlo; descargar `kitlegal-plugin.zip` y subirlo en *Customize > Plugins*, en el modo de chat y no en Code, o añadir el marketplace `jmorenobl/kitlegal-plugins`; (3) qué dice el aviso rojo de la app al instalar una extensión —«acceso a todo lo que hay en tu computadora», desarrollador sin verificar por Anthropic— y qué hace kitlegal de verdad: lee fuentes públicas y escribe solo en `~/.cache/kitlegal/`, sin enviar nada a ningún servidor propio (ADR 0027); (4) que hacen falta las dos piezas y qué pasa con una sola; (5) cómo se actualiza cada una; (6) que quien ya tiene las skills con `kitlegal skills install` y además instala el plugin las tiene dos veces en Claude Code; (7) que en Linux no hay app de escritorio donde abrir la extensión y se usa el binario instalado con Claude Code, como hasta ahora; y (8) que en la web y en el móvil no funciona, con la línea `⚠ SIN CONSULTA AL BOE:` que la persona verá. Las direcciones de descarga MUST ser las de «la última release», fijas.
- **FR-051**: El README MUST tener un apartado «Otras instalaciones, sin probar» con la extensión en Windows y con lo que hoy dice el README de la app de ChatGPT, de Codex y de Antigravity, y con dónde contar si funciona o qué falla: las incidencias del repositorio o `info@kitlegal.es`. MUST NOT dar por soportado nada que no sea la app de escritorio de Claude en macOS.
- **FR-052**: El README MUST dejar de decir que Claude Cowork y el chat de Claude «todavía no» y que llegarán con un plugin, y MUST NOT anunciar el plugin ni la extensión como algo por venir. Lo que diga de cómo se actualiza cada pieza MUST NOT afirmar lo que nadie ha probado: lo no probado se dice como tal (ADR 0035, «Pendiente de verificar»).
- **FR-053**: CONTRIBUTING MUST explicar el paso que empaqueta —qué lee, qué escribe, de dónde salen la versión, `tools` y los textos— y cómo probarlo en local, y MUST poner al día lo que dice de `make release`, de `make snapshot-check`, de `publicar`, de `humo` y de `PUBLISHER_TOKEN`. `CHANGELOG.md` (*Unreleased*) MUST registrar los dos ficheros, el catálogo y el apartado del README. `docs/SOURCES.md` MUST quedar sin cambios: no hay fuente nueva.

#### Controles y Definition of Done

- **FR-060**: `make snapshot-check` MUST fallar si en el `dist/` del snapshot falta `kitlegal.mcpb` o `kitlegal-plugin.zip`, o si `checksums.txt` no lleva la huella de cada uno. Lo ejecuta el trabajo `snapshot` de la CI en la propuesta de cambio.
- **FR-061**: `make snapshot-check` MUST fallar si el manifiesto del `.mcpb` del snapshot no se declara de la versión `0.3`, no lleva cada campo de FR-012 con su valor o lleva alguno que FR-012 no nombra, `user_config` incluido (FR-017); si su `version` no es la del binario que lleva dentro (FR-013); o si su `tools` no es exactamente el conjunto de herramientas que anuncia por MCP el binario de ese snapshot, con sus descripciones (FR-014).
- **FR-062**: `make snapshot-check` MUST fallar si `server/kitlegal` no lleva exactamente dos arquitecturas, una `amd64` y una `arm64` reconocidas por su tipo de CPU y no por su posición, si alguna no es, byte a byte, el binario `kitlegal` del archivo de macOS de esa arquitectura del mismo snapshot, si `server/kitlegal.exe` no es el de `kitlegal_windows_amd64.zip`, si `server/kitlegal` no va marcado como ejecutable, o si el `.mcpb` lleva alguna entrada que no sea una de las cuatro de FR-010.
- **FR-063**: `make snapshot-check` MUST fallar si algún fichero de `skills/` del plugin difiere de lo empotrado en el binario de ese snapshot, si falta alguno o hay alguno de más, o si el plugin lleva `bin/`, `.mcp.json`, `mcpServers` o algún `.mcpb`.
- **FR-064**: `make snapshot-check` MUST extraer el `.mcpb` en una ruta con espacios, poner el binario de la plataforma que lo ejecuta en el sitio de `server/kitlegal`, arrancar el servidor con un cliente MCP usando la orden y los argumentos del `mcp_config` del manifiesto, con `/` como directorio de trabajo, y fallar si no lista exactamente las herramientas de H21. Lo que prueba es el manifiesto.
- **FR-065**: El trabajo `snapshot` de la CI MUST ejecutar `claude plugin validate` sobre el plugin extraído del snapshot y sobre la plantilla del catálogo, con Claude Code en la versión que fija `evals.yml`, y salir en rojo si alguna de las dos no es válida. MUST NOT entrar en `make ci`, que no depende de Claude Code, y MUST NOT abrir ninguna sesión con modelo ni usar ninguna credencial.
- **FR-066**: Los umbrales de la ficha MUST tener control en `make ci` y en el snapshot: la descripción corta, 120 caracteres como mucho (FR-015), y el icono, `mcp/icon.png`, un PNG de 512 × 512 px (FR-016).
- **FR-067**: Un control MUST ejecutar el paso dos veces sobre los mismos binarios y fallar si alguno de los dos ficheros difiere en un byte entre las dos (FR-004). MUST estar en `make ci`, sobre binarios de prueba, para que no dependa de construir las seis plataformas.
- **FR-068**: `goreleaser check` y `make ci` MUST quedar en verde. Las comprobaciones estrictas de la definición de la release que ya hay en `make ci` MUST cubrir lo que el hito añade a `.goreleaser.yaml` y a `release.yml`: el paso, los dos ficheros en `release.extra_files` y en `checksum.extra_files`, `universal_binaries` con `replace: false` y `archives` acotado con `ids`, su atestación, los tres trabajos de `release.yml` con de qué depende el del catálogo, sus permisos y qué pasos ven `PUBLISHER_TOKEN`, y las comprobaciones nuevas de `humo`.
- **FR-069**: El código nuevo MUST llevar sus tests unitarios offline y cumplir las reglas de arquitectura y de dependencias de hoy (Definition of Done §1.1 a §1.3 y §1.9): el paso MUST NOT añadir ninguna dependencia fuera de la lista del principio V, y la regla de `depguard` de H21 sobre el SDK de MCP MUST seguir cumpliéndose tal como está.
- **FR-070**: La sección «Controles de umbral» del plan MUST tener una fila por cada umbral de este spec que se mide sin persona: los 120 caracteres (FR-015), los 512 px (FR-016), los cero bytes de diferencia entre dos ejecuciones (FR-004), las cuatro entradas del `.mcpb` (FR-010), el conjunto exacto de `tools` (FR-014) y los cero ficheros distintos, de menos o de más en `skills/` del plugin (FR-020).

#### Relación con otros hitos

- **FR-080**: Los specs de los hitos anteriores MUST NOT editarse. Las skills, las herramientas y las `instructions` del servidor MUST NOT cambiar: `skills-check`, `schema-check` y las comprobaciones de H21 siguen en verde sin tocar lo que comprueban.

### Key Entities

- **El paso**: el programa del repositorio que escribe los dos ficheros a partir de los binarios compilados, del registro, de las skills empotradas, de los textos y del icono.
- **`kitlegal.mcpb` (la extensión)**: un zip con el manifiesto, el icono y dos binarios. Da las herramientas a cualquier conversación de la app.
- **Manifiesto**: `manifest.json` de la versión `0.3`: identidad, textos de la ficha, cómo se arranca el servidor en cada sistema y `tools`.
- **`kitlegal-plugin.zip` (el plugin)**: un zip con `plugin.json` y las skills de esa versión. Da el protocolo; no lleva servidor.
- **El catálogo**: `.claude-plugin/marketplace.json` de `jmorenobl/kitlegal-plugins`, con una entrada que apunta al plugin de una etiqueta por su dirección y su huella.
- **La plantilla del catálogo**: el fichero de este repositorio del que sale el catálogo de cada etiqueta.
- **Los textos**: `display_name`, `description`, `long_description` y `author`, en un solo sitio del repositorio.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En el run, el snapshot de la propuesta de cambio deja los dos ficheros con sus huellas y los controles de SC-003 a SC-011 pasan. **Control**: el trabajo `snapshot` de la CI de la propuesta de cambio sale en rojo si `make release` no deja los dos ficheros, si `make snapshot-check` falla o si `claude plugin validate` falla; y el trabajo `ci`, si falla `make ci`. El cierre del workflow cuenta esos rojos.
- **SC-002**: Después de fusionar y de publicar una release, y fuera del run porque es humano: en un Mac distinto del que compiló, sin kitlegal instalado y sin abrir una terminal, con `kitlegal.mcpb` y `kitlegal-plugin.zip` descargados de la release con un navegador e instalados, una conversación nueva de la app de Claude, sin carpeta, responde «¿qué dice el art. 21 de la Ley 39/2015?» con `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`. En esa misma prueba se anota lo que el ADR 0035 dejó pendiente: si macOS bloquea el binario descargado, cómo llega una versión nueva de la extensión y del plugin, si la app admite el marketplace, y qué responde Claude en la web con el plugin y sin la extensión. Lo mide una persona: no tiene control en el run.
- **SC-003**: El snapshot deja 2 de 2 ficheros, `kitlegal.mcpb` y `kitlegal-plugin.zip`, y los 2 están en `checksums.txt`. **Control**: `make snapshot-check` falla (FR-060).
- **SC-004**: El manifiesto del `.mcpb` del snapshot se declara de la versión `0.3` con los campos de FR-012 y ninguno más, su `version` es la del binario que lleva dentro y su `tools` es exactamente el conjunto de herramientas que ese binario anuncia por MCP —10 de 10 con el registro de hoy—, con sus descripciones. **Control**: `make snapshot-check` falla (FR-061).
- **SC-005**: Las 2 arquitecturas de `server/kitlegal` son, byte a byte, los binarios de los archivos de macOS `amd64` y `arm64` del mismo snapshot, `server/kitlegal.exe` es el de Windows `amd64`, y el `.mcpb` lleva 0 ejecutables más: 4 entradas en total. **Control**: `make snapshot-check` falla (FR-062).
- **SC-006**: Un cliente MCP arranca el servidor con el `mcp_config` del manifiesto, desde el `.mcpb` extraído en una ruta con espacios y con `/` como directorio de trabajo, y lista las 10 herramientas de H21. **Control**: `make snapshot-check` falla (FR-064).
- **SC-007**: 0 ficheros de `skills/` del plugin distintos de lo empotrado en el binario de ese snapshot, 0 de menos y 0 de más; y 0 apariciones de `bin/`, `.mcp.json`, `mcpServers` o un `.mcpb` en el plugin. **Control**: `make snapshot-check` falla (FR-063).
- **SC-008**: `claude plugin validate` da por válidos 2 de 2: el plugin extraído y la plantilla del catálogo. **Control**: el trabajo `snapshot` de la CI sale en rojo (FR-065).
- **SC-009**: La descripción corta tiene 120 caracteres como mucho, y el icono del `.mcpb` es `mcp/icon.png`, un PNG de 512 × 512 px. **Control**: un test de `make ci` falla con 121 caracteres o con un icono de otro tamaño, y `make snapshot-check` falla si el `.mcpb` del snapshot no los cumple (FR-066).
- **SC-010**: Dos ejecuciones del paso sobre los mismos binarios dan 0 bytes de diferencia en cada uno de los dos ficheros. **Control**: un test de `make ci` falla (FR-067).
- **SC-011**: `goreleaser check` y `make ci` en verde, con la definición de la release comprobada de forma estricta con lo nuevo. **Control**: `make ci` falla (FR-068).
- **SC-012**: El README lleva «Instalar sin terminal» con sus ocho contenidos y «Otras instalaciones, sin probar»; CONTRIBUTING y `CHANGELOG.md` (*Unreleased*) llevan lo de FR-053; y `docs/SOURCES.md` no cambia. **Control**: la documentación la comprueba la revisión final contra FR-050 a FR-053; no tiene umbral numérico.

## Uso, de fuera adentro

Cada salida del hito, desde quien la consume (criterio de uso, ADR 0028). Lo que ganan las dos skills es llegar a quien no usa una terminal: el plugin las lleva tal como están, y la extensión les da las herramientas de H21. Volumen de referencia: meses de uso diario, con cientos de normas y miles de bloques consultados. Ninguna salida de este hito crece con ese volumen: los dos ficheros, la ficha y el catálogo dependen de la versión, no de lo consultado. Lo que sí crece con el uso —la caché y el grafo en `~/.cache/kitlegal/`— no cambia en este hito, y lo que cada herramienta devuelve sigue acotado por lo que se pide (H21, «Uso, de fuera adentro»; H7.1).

Medidas de esta sesión, sobre `main` (`751b76e`): los cinco ficheros de las dos skills suman 44 237 bytes (`wc -c` sobre `skills/`: 23 944 y 3 413 en `boe-legislacion`, 12 716, 2 621 y 1 543 en `legal-core`); `mcp/icon.png` es un PNG de 512 × 512 px (`file`); los tres binarios que irían en el `.mcpb`, compilados con `CGO_ENABLED=0` y `-trimpath` y comprimidos con `gzip`, ocupan 14 057 626 bytes (darwin/amd64), 13 340 163 (darwin/arm64) y 14 013 282 (windows/amd64): 41 411 071 en total. No se ha medido en esta sesión el tamaño sin comprimir de esos binarios, el de un `.mcpb` real, que comprime como zip y no como gzip, ni los bytes de `tools`; tampoco se ha ejecutado `claude plugin validate`, que es de la CI.

| Salida | Quién la pide, cuántas veces y qué hace con ella | Tamaño | Cuándo deja de darse cada señal |
|---|---|---|---|
| `kitlegal.mcpb` (FR-010) | La persona, una vez por versión: lo descarga y lo abre. La app lo extrae y arranca el servidor en cada arranque suyo. Ninguna skill lo lee. | Del orden de 41 MB: tres binarios comprimidos, más el icono (≈ 20 KB) y el manifiesto. Crece con el binario, no con el uso. | No da señales. |
| La ficha de la extensión: nombre, icono, descripción y `tools` (FR-012, FR-014 a FR-016) | La persona, una vez, al instalar, para decidir si acepta. | Una descripción de 120 caracteres como mucho, un icono y 10 herramientas con su descripción. Crece con los verbos, no con el uso. | No da señales. El aviso rojo es de la app, no de kitlegal: el README lo explica (FR-050). |
| `kitlegal-plugin.zip` (FR-020) | La persona, una vez por versión si lo sube a mano; ninguna si usa el catálogo. La app carga de él las skills. | 5 ficheros de skills, 44 237 bytes sin comprimir, y `plugin.json`. Crece con las skills, no con el uso. | No da señales. |
| Las skills del plugin | El modelo, una vez por conversación en que se activa cada una, como hoy las instaladas con `skills install`. | Las de hoy, byte a byte: `SKILL.md` < 300 líneas. | Las suyas, sin cambios (H21). |
| Las herramientas que da la extensión | La skill, las mismas veces por pregunta que en H21: en la prueba, `boe_indice`, `boe_articulo` y `graph_check` para una pregunta por un artículo. | El de H21: el sobre de cada llamada, acotado por la norma y el bloque pedidos. | Las de H21 y H7.1: cada `version-obsoleta` se apaga con la siguiente lectura del bloque. |
| La línea `⚠ SIN CONSULTA AL BOE:` con el plugin y sin la extensión | La persona; una por respuesta (H21 FR 035). | Una línea. | Deja de darse en la primera pregunta tras instalar la extensión. |
| El catálogo (FR-031) | La app o Claude Code de quien lo añadió, cuando comprueban si hay versión nueva. | Una entrada: nombre, versión, textos, una dirección y una huella; menos de 2 KB. No crece: cada etiqueta sustituye la anterior. | La versión nueva se ofrece cuando cambia `version`, y deja de ofrecerse al instalarla. |
| Las dos líneas nuevas de `checksums.txt` (FR-002) | `humo`, una vez por release (FR-040); quien quiera comprobar una descarga; y el paso que publica el catálogo, que toma de ahí el `sha256` (FR-031). | Dos líneas de huella y nombre. | No dan señales. |
| El fallo del paso, en la salida de error (FR-005) | Quien mantiene el proyecto, en el registro del snapshot o de la release. | Una línea que nombra lo que falta. | Deja de darse cuando la entrada está. |
| «Instalar sin terminal» del README (FR-050) | La persona, antes de instalar y cuando ve el aviso rojo. | Un apartado con ocho contenidos. | No da señales. |

## Fuera de alcance

Del hito, literal:

- «dar por soportado Windows o cualquier agente que no sea la app de escritorio de Claude, y cualquier prueba, control o arreglo específico de Windows más allá de que su binario vaya en el `.mcpb`;»
- «un plugin o una extensión para otro agente;»
- «la notarización del binario de macOS, que pide una cuenta de Apple Developer y es una credencial y una decisión de la persona —si la prueba de aceptación la hace necesaria, llega en su propia propuesta de cambio—;»
- «el envío al directorio de plugins y de extensiones de Anthropic (§0);»
- «que `skills install` escriba la configuración MCP de ningún agente;»
- «los packs;»
- «Linux;»
- «cambios en las skills, en las herramientas o en las `instructions` del servidor;»
- «y la web, que la cambia la persona (ADR 0024).»

De lo que el hito no especifica (constitución, «Criterio de decisión autónoma», punto 2):

- Versionar en `testdata/` el esquema JSON de la versión `0.3` del manifiesto de MCP Bundle y validar el manifiesto contra él (FR-017): pendiente de la persona, en una propuesta de cambio propia o corrigiendo la entrada del hito y relanzando. Hasta entonces nadie valida en el run la forma exacta del manifiesto contra el esquema oficial: lo prueba la instalación a mano de SC-002, donde la app acepta o rechaza la extensión.
- Un séptimo archivo de la release con el binario universal de macOS, con su huella, su SBOM y su atestación (FR-006).
- Un applet, un verbo o una bandera del binario para empaquetar, y cualquier cambio en `kitlegal skills install`, `list` o `doctor`.
- Un binario de Windows `arm64` o de Linux dentro del `.mcpb`, `user_config` en el manifiesto y cualquier campo del manifiesto que FR-012 no nombra.
- Agentes, órdenes, ganchos o servidores dentro del plugin: lleva `plugin.json` y `skills/`.
- Más de una entrada en el catálogo, catálogos para otros agentes, y escribir en `jmorenobl/kitlegal-plugins` otra cosa que `.claude-plugin/marketplace.json`.
- Dar a `PUBLISHER_TOKEN` permiso sobre el catálogo, etiquetar y publicar: son de la persona (ADR 0020).
- Evitar que las skills estén dos veces en Claude Code con el plugin y `skills install`, y arreglar la respuesta de la extensión sola: se documentan.
- Capturas de pantalla como imágenes en el README: el hito las pide «descritas en texto».
- Evals nuevas, un modo nuevo del job de evals o medir con él la app de escritorio: las skills no cambian, y la aceptación de la app es humana.
- Firmar o notarizar el binario de macOS, y retirar atributos de cuarentena en el equipo de quien instala.
- Cambiar el workflow `hito` o `scripts/workflow/`, editar specs anteriores y un ADR nuevo.

## Assumptions

- **«La versión del binario que lleva dentro».** En una release el binario imprime la etiqueta con su `v` (`kitlegal v0.4.0`) y el manifiesto lleva `0.4.0`; en un snapshot, el binario imprime la versión del snapshot. FR-013 pide que coincidan sin la `v` inicial, que es lo que el hito dice con «la etiqueta sin la `v`».
- **«Las herramientas de H21»** son las diez del registro de hoy. El control compara con lo que anuncia el binario del snapshot, así que un verbo nuevo no lo rompe (FR-014, FR-064).
- **Cuatro entradas en el `.mcpb`.** El hito dice «un zip con `manifest.json` y `server/`», pide `icon.png` dentro y prohíbe cualquier otro ejecutable. La lectura que menos añade es que no lleva nada más (FR-010).
- **El bit de ejecución.** El hito no lo nombra, pero sin él la app no puede arrancar `server/kitlegal` tras extraerlo, y el control de FR-064 lo ejerce.
- **Dónde corren los controles.** Los del hito dicen «del snapshot», y el snapshot construye seis plataformas, así que viven en `make snapshot-check`, que ejecuta el trabajo `snapshot` de la CI y que el cierre del workflow cuenta como cualquier comprobación de la propuesta de cambio (`scripts/workflow/cierre.sh` cuenta toda comprobación en rojo). Los que no necesitan el snapshot —120 caracteres, el icono, la reproducibilidad— están además en `make ci`, como pide la Definition of Done para el código nuevo.
- **`claude plugin validate`** se ejecuta solo en la CI. No es una sesión con modelo ni usa credenciales, pero ninguna sesión del run lo ejecuta (ADR 0032), y este spec no lo ha ejecutado: que acepte un catálogo con una entrada `archive` es lo leído en la documentación el 2026-10-02 (ADR 0035), no una medida.
- **La plantilla se valida rellenada.** `sha256` y la dirección solo existen tras publicar; el control la rellena con valores de prueba. Cómo, es del plan.
- **Cuándo se publica el catálogo.** Después de que la release exista y de que `humo` salga en verde, porque su dirección y su huella son las de lo publicado (FR-031). Cómo se escribe el paso del trabajo nuevo es del plan.
- **Lo que `release.yml` hace solo al etiquetar** —atestar, publicar el catálogo, `humo`— no se puede ejecutar en el run. En el run se comprueba su definición (FR-068); su primera ejecución real es la release que la persona publique para SC-002.
- **El job de evals.** Este hito no toca las skills, las evals ni el binario distribuido más allá de cómo se empaqueta. Si el cierre lanza el job, tiene que seguir en verde como en H21; no hay umbral nuevo que decidir.
- **Lo que dice el ADR 0035 de la app de Claude** —cómo instala una extensión y un plugin, qué enseña la ficha, desde dónde arranca el servidor— se toma como está: es lo probado a mano el 2026-10-01 y el 2026-10-02. Lo que quedó pendiente lo anota SC-002.
- **Los nombres técnicos que aparecen** (`release.extra_files`, `checksum.extra_files`, `universal_binaries`, `mcp_config`, `platform_overrides`, `claude plugin validate`, `make snapshot-check`, `humo`, `publicar`) son los del hito o existen en el repositorio. Quedan para el plan: dónde vive el paso, dónde viven los textos y la plantilla, el valor de cada texto, cómo consigue el paso el conjunto de herramientas y las skills empotradas, cómo se hace reproducible el zip, y con qué cliente MCP comprueban `make snapshot-check` y `humo`.
