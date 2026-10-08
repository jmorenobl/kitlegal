<h1 align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/img/portada-oscuro.svg">
    <img src="docs/img/portada-claro.svg" alt="kitlegal: tu agente responde con la ley delante y la cita exacta" width="100%">
  </picture>
</h1>

**kitlegal enseña a tu agente a responder preguntas legales correctamente: con el texto vigente delante y la cita
exacta.**

Es un conjunto de *skills* para Claude Code, Codex, Antigravity y cualquier agente que siga el estándar Agent
Skills y trabaje en tu equipo, más un programa, `kitlegal`, que les da lo que un modelo no debe hacer de memoria:
leer el texto vigente en el BOE, comprobar que sigue en vigor, situar un municipio, recordar lo que ha consultado y
dejar cada dato con su fuente, su fecha y su huella. Instalas kitlegal y, a partir de ahí, es **tu agente** el que consulta, razona y cita;
kitlegal le da el método y las herramientas.

La regla que lo gobierna todo: **nada sin cita**. Cada afirmación sobre una norma sale del texto que el programa acaba
de leer, y si no lo puede leer, el agente lo dice en lugar de suplirlo.

Web del proyecto: **[kitlegal.es](https://kitlegal.es)**.

## Qué sabe hacer tu agente con kitlegal

Hay tres skills. La tercera, `jurisprudencia`, es la de las sentencias, y tiene su apartado:
[¿Y las sentencias?](#y-las-sentencias). Con las otras dos, `boe-legislacion` y `legal-core`, tu agente responde sobre
**cualquier norma consolidada en el BOE** —leyes, reales decretos, textos refundidos, también las normas autonómicas
que el BOE consolida— y sitúa **cualquier municipio de España** en su territorio. Por ejemplo:

- «¿Qué dice el artículo 21 de la Ley 39/2015?»
- «¿Cuál es el límite de un contrato menor en la Ley de Contratos del Sector Público?»
- «¿Qué atribuciones tiene el Pleno de un ayuntamiento?»
- «¿Cuándo prescribe una deuda tributaria?»
- «¿Qué impuestos puede cobrar un ayuntamiento?»
- «¿Cuántos días de vacaciones fija el Estatuto de los Trabajadores?»
- «¿En qué plazo tiene que resolver la Administración una solicitud de acceso a información pública?»
- «¿Qué comunidad, provincia y boletines oficiales corresponden a mi ayuntamiento?»
- «Ya te pregunté hace tiempo por el artículo 21 de la Ley 39/2015. ¿Qué dice ahora?»

Tu agente responde citando la norma y el artículo tal como los publica el BOE, siempre con la misma forma:

> El plazo máximo para resolver es de tres meses cuando la norma no fije otro
> —art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]—.

Y si la norma ya no está en vigor, lo dice antes que nada:

> ⚠ NORMA DEROGADA: la Ley 30/1992 fue derogada por la Ley 39/2015 …

Y si ya la habías consultado y su redacción ha cambiado desde entonces, también te lo dice, con la fecha de vigencia
de la redacción que consultaste y la de la actual.

Cuando la pregunta depende de un municipio, tu agente empieza por situarlo —provincia, comunidad autónoma, régimen
común o foral, boletines oficiales— y te dice qué parte de eso está configurada y qué no. Hoy los boletines están
configurados para la Comunidad de Madrid; para el resto de España sitúa el municipio igual y declara que su boletín
autonómico y su boletín provincial no están configurados, en lugar de inventarlos.

## Instalar sin terminal

Es la instalación oficial de kitlegal, y está probada en la **app de escritorio de Claude en macOS**. En Windows los
pasos son los mismos y nadie los ha probado ([Otras instalaciones, sin probar](#otras-instalaciones-sin-probar)).

Es una sola pieza, el plugin de kitlegal, y no hay que descargar nada:

1. **Añade el plugin.** En *Customize > Plugins > Add*, estando en el modo de chat de la app y no en Code, añade el
   marketplace `jmorenobl/kitlegal-plugins` e instala desde él el plugin `kitlegal`. Lo que se añade desde Code no
   llega a las conversaciones de chat. El plugin lleva las skills y, dentro, el servidor con las herramientas.
2. **Cierra la app del todo y ábrela otra vez.** Las herramientas arrancan cuando la app se abre; sin este paso
   tardan hasta una hora en estar.

Después abre una conversación nueva y pregunta. La primera vez, Claude pide permiso para usar las herramientas.

**El aviso rojo.** Al añadir el marketplace, la app avisa en rojo de que los plugins instalados desde tiendas no
están controlados por Anthropic. El aviso es de la app, no de kitlegal, y sale con cualquier plugin que no venga de
su directorio. kitlegal lleva dentro un programa que corre en tu equipo. Lo que hace de verdad: lee fuentes públicas
—hoy, el BOE—, escribe solo en `~/.cache/kitlegal/`, donde guarda lo que ya ha leído, y no envía nada a ningún
servidor propio.

**Con el plugin basta.** La respuesta lleva la cita con su forma:
`art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`. Hasta la v0.4.1 eran dos piezas, la extensión y el
plugin: si tienes instalada la extensión `kitlegal`, desinstálala en *Settings > Extensions*, porque el plugin trae
las mismas herramientas y con las dos el servidor está dos veces.

**Si prefieres no añadir un marketplace**, descarga
[`kitlegal-plugin.zip`](https://github.com/jmorenobl/kitlegal/releases/latest/download/kitlegal-plugin.zip) y súbelo
en *Customize > Plugins > Add > Upload plugin*: es el mismo plugin, pero con cada versión nueva hay que descargarlo y
subirlo otra vez.

**Actualizar.** La app no avisa de que hay una versión nueva ni la instala sola, tampoco con «Sincronizar
automáticamente» activado (probado el 2026-10-04: en catorce horas no la trajo). En *Customize > Plugins*, abre
kitlegal, menú de los tres puntos, «Buscar actualizaciones», y cuando se encienda el botón, «Actualizar». Las
herramientas pasan solas a la versión nueva en menos de una hora, o al momento al cerrar y abrir la app. Si subiste
el zip, sube el nuevo.

**Si ya tienes las skills con `kitlegal skills install`** y además instalas el plugin, en Claude Code las tienes dos
veces: las de tu disco y las del plugin, que llega también a Claude Code.

**En Linux** no hay app de escritorio de Claude en la que instalar el plugin: se usa el programa instalado, con
Claude Code, como hasta ahora ([Instalar con la terminal](#instalar-con-la-terminal)).

**Desde Claude en el móvil y en la web** se puede preguntar una vez instalado, mientras tu equipo esté encendido y
con la app de escritorio abierta: la conversación usa las herramientas de tu equipo, donde kitlegal sigue corriendo
(probado el 2026-10-04 en macOS). Con el equipo apagado o la app cerrada llegan las skills del plugin y ninguna
herramienta, y lo que verás, en lugar de una respuesta de memoria, es esta línea:

```text
⚠ SIN CONSULTA AL BOE: <causa>. Para consultarlo hace falta instalar kitlegal: https://kitlegal.es/instalar/
```

## Instalar con la terminal

Para Claude Code, y la forma de usar kitlegal en Linux. Dos órdenes: la primera instala el programa; la segunda, las
skills, en el proyecto en el que estés.

**macOS y Linux**

```sh
curl -fsSL https://raw.githubusercontent.com/jmorenobl/kitlegal/main/scripts/install.sh | sh
kitlegal skills install
```

**Windows** (PowerShell, con [Scoop](https://scoop.sh))

```powershell
scoop bucket add jmorenobl https://github.com/jmorenobl/scoop-bucket
scoop install kitlegal
kitlegal skills install
```

Si ya usas Homebrew: `brew install jmorenobl/tap/kitlegal`. Cada release adjunta además paquetes `.deb` y `.rpm`, y
con Go instalado vale `go install github.com/jmorenobl/kitlegal/cmd/kitlegal@latest`.

`kitlegal skills install` deja las skills en `.agents/skills/` del directorio en el que lo ejecutes —así van con el
proyecto—. Claude Code las carga de `.claude/skills/`: si el proyecto ya tiene un `.claude/`, las enlaza también ahí,
y en una carpeta nueva lo hace `kitlegal skills install --host claude`. Después, abre Claude Code en ese directorio y
pregunta.

Para tenerlas en todos tus proyectos a la vez, en lugar de en uno: `kitlegal skills install -g`. Las deja en
`~/.agents/skills/` y las enlaza en `~/.claude/skills/` si usas Claude Code; si aún no lo has abierto,
`--host claude` las enlaza igual.

**Actualizar**: repite la primera orden (o `brew upgrade kitlegal`, o `scoop update kitlegal`) y vuelve a ejecutar
`kitlegal skills install`. Mientras lo instalado sea de una versión anterior a la del programa, el programa se lo hace
saber a tu agente en cada consulta, sin pedir nada a la red, y tu agente te lo dice.

Lo que conviene saber del instalador: no necesita nada que no traigan ya macOS y cualquier Linux (`curl`, `tar` y
`sha256sum` o `shasum`); comprueba la huella SHA-256 de lo que descarga; deja `kitlegal` en `~/.local/bin` (o en
`$KITLEGAL_INSTALL_DIR`) y, si ese directorio no está en tu `PATH`, te imprime la línea que lo añade en lugar de tocar
tus ficheros de arranque; ante cualquier fallo no instala nada a medias. `kitlegal skills install --dry-run` cuenta lo
que haría sin tocar nada, y nunca sobrescribe nada que no haya instalado él: si encuentra algo suyo con el nombre de
una skill —una carpeta tuya, un fichero editado—, lo nombra y no cambia nada. `kitlegal skills doctor` comprueba lo
instalado y, si algo no está como lo dejó, da la orden de una línea que lo arregla. Cada archivo publicado lleva su
atestación de procedencia, comprobable con `gh attestation verify <archivo> --repo jmorenobl/kitlegal`.

## Otras instalaciones, sin probar

Así se haría en cada uno de estos, y nadie lo ha probado: lo que sigue sale de lo que kitlegal ya lleva y de la
documentación de cada programa. Si lo intentas, cuéntanos si te funciona o qué falla, en
[las incidencias](https://github.com/jmorenobl/kitlegal/issues) o en `info@kitlegal.es`.

- **El plugin en Windows**: los dos pasos de [Instalar sin terminal](#instalar-sin-terminal), en la app de
  escritorio de Claude para Windows. El plugin lleva ya el programa para Windows, que funciona en un Windows de
  prueba; lo que nadie ha probado es instalarlo en la app.

- **La app de escritorio de ChatGPT y Codex**, con el programa instalado
  ([Instalar con la terminal](#instalar-con-la-terminal)). Codex lee las skills de `.agents/skills/` cuando abres el
  directorio en el que las dejó `kitlegal skills install`, y de `~/.agents/skills/`, donde las deja
  `kitlegal skills install -g`. Ejecuta las órdenes en un entorno aislado y sin red, así que te pedirá permiso cada
  vez que tu agente consulte el BOE con `kitlegal`. Si no quieres aprobarlo en cada consulta, añade esta regla a
  `~/.codex/rules/default.rules` (las reglas de Codex son todavía experimentales); con ella, Codex ejecuta `kitlegal`
  sin preguntar y fuera de ese entorno aislado:

  ```
  prefix_rule(pattern = ["kitlegal"], decision = "allow")
  ```

  El [servidor MCP](#el-servidor-mcp-las-mismas-consultas-como-herramientas) se declara una vez para los dos, que
  comparten la configuración: en *Settings > MCP servers*, con la orden `kitlegal` y los argumentos `mcp serve`, o
  desde una terminal:

  ```sh
  codex mcp add kitlegal -- kitlegal mcp serve
  ```

- **Antigravity**, también con el programa instalado. Lee las skills de `.agents/skills/` cuando abres ese
  directorio; en global no lee `~/.agents/skills/`, y `kitlegal skills install -g` las enlaza en
  `~/.gemini/config/skills/` (si aún no lo has abierto, con `--host antigravity`). El servidor MCP se declara en
  `~/.gemini/config/mcp_config.json`, o en `.agents/mcp_config.json` para un solo proyecto:

  ```json
  {
    "mcpServers": {
      "kitlegal": { "command": "kitlegal", "args": ["mcp", "serve"] }
    }
  }
  ```

Si el agente no encuentra `kitlegal`, escribe en su lugar la ruta completa del programa, la que da
`command -v kitlegal`.

## El servidor MCP: las mismas consultas, como herramientas

Las skills consultan con órdenes de `kitlegal`, que tu agente ejecuta en una terminal. `kitlegal mcp serve` les da
otra vía: un servidor MCP que ofrece cada consulta como una herramienta —buscar una norma, leer su índice, uno o
varios artículos, sus metadatos y su análisis, situar un municipio, repasar lo ya consultado, preparar la consulta de
una sentencia y cotejar el documento que traes—, con el mismo resultado que su orden: el texto con su fuente, su
dirección, su fecha de consulta y su huella. Las skills usan la herramienta siempre que tu agente la tiene, y la orden
cuando no.

El servidor corre en tu equipo: lo arranca tu agente y habla con él por la entrada y la salida estándar, sin abrir
ningún puerto. En la app de escritorio de Claude lo trae el plugin de
[Instalar sin terminal](#instalar-sin-terminal), y no hay nada que declarar. En **Claude Code** necesita el programa
instalado y se declara una vez:

```sh
claude mcp add kitlegal -- kitlegal mcp serve
```

Si Claude Code no encuentra `kitlegal`, escribe en su lugar la ruta completa del programa, la que da
`command -v kitlegal`.

Está probado en **Claude Code** y en la **app de escritorio de Claude** en macOS. Cómo se declararía en la app de
escritorio de ChatGPT, en Codex y en Antigravity está en
[Otras instalaciones, sin probar](#otras-instalaciones-sin-probar).

**ChatGPT en la web y en el móvil no es compatible**: solo admite servidores remotos, y el de kitlegal corre en tu
equipo. Claude en la web y en el móvil funciona a través de la app de escritorio, con el equipo encendido y la app
abierta ([Instalar sin terminal](#instalar-sin-terminal)).

## Cómo funciona y qué te garantiza

- **Tu agente lee la fuente, no su memoria.** No responde con lo que «sabe» de una ley: ejecuta `kitlegal`, que
  descarga el texto vigente del BOE en ese momento y se lo devuelve con la fuente, la dirección, la fecha de consulta y
  una huella del contenido. Sobre eso razona y cita. Sin esos cuatro datos no hay cita, y sin cita no hay respuesta.
- **Comprueba la vigencia.** Antes de darte un artículo, el programa comprueba si la norma está derogada, si su
  vigencia ha terminado o si el BOE aún no ha terminado de consolidarla, y tu agente te lo traslada con un aviso de
  forma fija (`⚠ NORMA DEROGADA`, `⚠ VIGENCIA AGOTADA`, `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO`).
- **Recuerda lo que ha consultado, y te avisa si ha cambiado.** Cada artículo que lee del BOE y cada municipio que
  sitúa quedan anotados en tu equipo, con su fuente, su dirección y su fecha de consulta, en el *grafo del mundo*: un
  fichero, `world.db`, en la misma carpeta en la que guarda lo ya leído (`~/.cache/kitlegal/`, u otra con
  `KITLEGAL_CACHE_DIR`). Cuando vuelves a preguntar por una norma, tu agente la lee de nuevo, repasa el grafo con
  `kitlegal graph check` y esa norma, y te dice con una forma fija (`⚠ REDACCIÓN MODIFICADA:`) si la redacción
  ha cambiado desde la que se leyó la vez anterior. Lo que cita sale siempre de la lectura nueva: el grafo guarda
  también el texto de cada redacción que ha leído, pero nunca responde con él, y `kitlegal graph show <id>` y
  `kitlegal graph stats`, que enseñan lo anotado, no devuelven texto legal. Nada de eso sale de tu equipo, y una orden
  con `--no-graph` no anota nada.
- **Distingue.** Ley de reglamento, norma estatal de autonómica, y señala cuándo la respuesta puede variar según la
  comunidad autónoma.
- **Solo fuentes públicas, y solo lectura.** kitlegal no entra en ninguna sede electrónica, no envía nada en tu
  nombre y no necesita ninguna identidad tuya. Cuando en el futuro prepare un escrito, terminará en un fichero listo
  para que lo firmes tú.
- **Respeta las fuentes.** Se identifica ante el BOE, respeta su `robots.txt`, limita el ritmo de sus consultas y
  guarda en tu equipo lo que ya ha leído para no pedirlo dos veces.
- **Funciona con el agente y el modelo que ya uses.** kitlegal no lleva ninguna IA dentro, no envía tus preguntas a
  ningún servicio propio y no necesita ninguna cuenta: lo único que sale de tu equipo son las consultas al BOE.
- **No sustituye a un profesional.** Tu agente, con kitlegal, consulta y cita; no asesora ni tramita. Para una
  decisión con consecuencias, acude a un abogado, un asesor o el servicio de atención de tu administración.

Todo esto se mide: cada skill tiene un conjunto de preguntas de prueba que se ejecutan con el modelo real, en sesiones
de agente de verdad, antes de cada cambio que las toca; y el programa se comprueba cada noche contra el BOE real para
detectar si la fuente ha cambiado.

## ¿Y las sentencias?

kitlegal **no consulta el CENDOJ**, el buscador del Centro de Documentación Judicial del Consejo General del Poder
Judicial, que es donde están las sentencias del Tribunal Supremo, la Audiencia Nacional, los tribunales superiores de
justicia y las audiencias provinciales. Ese buscador responde con un CAPTCHA al programa que se identifica como
programa, y a una persona con su navegador, no. kitlegal dice quién es ante cada fuente, y un CAPTCHA no se sortea: ni
presentándose como un navegador, ni resolviéndolo, ni entrando por otra dirección del mismo buscador. Su aviso legal,
además, reserva la consulta al uso particular y prohíbe la descarga masiva.

Así que la búsqueda la haces tú, con tu navegador, y tu agente, con la skill `jurisprudencia`, hace lo de alrededor:

- **Prepara la consulta exacta.** De una sentencia que nombras —por su ECLI, por su ROJ o como se escribe en un
  escrito, «STS 1088/2023, de 4 de julio»— te da la dirección del buscador y cada casilla con lo que hay que escribir
  en ella; de una pregunta por materia, la dirección que abre el buscador con la búsqueda ya hecha. Descargas el
  documento y se lo traes, pegado o adjunto.
- **Coteja el documento que traes.** Lee la ficha con la que el CENDOJ encabeza cada documento y te dice si es la
  sentencia que se pidió; si no lo es, qué difiere. Pasa: el número con el que se cita una sentencia no es su ROJ, y
  buscada por ese número como ROJ sale la de otro asunto.
- **Solo cita la sentencia cuyo documento tiene delante**, con los datos de su ficha y siempre con la misma forma:

  > STS 1088/2023, de 4 de julio [ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023]

  De la que no tiene delante no dice que existe ni que no existe, porque nadie lo ha comprobado. Lo dice así, con la
  consulta para encontrarla debajo:

  > ⚠ SENTENCIA NO COMPROBADA: STS 1088/2023, de 4 de julio

Las dos operaciones, `kitlegal cita preparar` y `kitlegal cita cotejar`, corren en tu equipo y no piden nada a la red.
La cita se apoya en el documento que aportas: que sea auténtico no lo comprueba nadie. Con la ficha sola, tu agente da
la cita y sus datos; de lo que dice la sentencia responde con su texto delante.

**Las sentencias del Tribunal Constitucional** no están en el CENDOJ: hoy tu agente lo dice y te remite al buscador
del propio tribunal, sin citarlas. Está previsto que lleguen por una vía que lo permite: se publican todas en el BOE,
como las del Supremo que anulan una disposición, y tu agente podrá leerlas y citarlas con su ECLI igual que hoy cita
un artículo.

La skill `jurisprudencia` y sus dos operaciones están en el repositorio; la v0.5.0 no las lleva. Buscar o leer
sentencias de forma automática pediría un acuerdo con el Consejo General del Poder Judicial, por su procedimiento de
reutilización. La decisión, en
[docs/ADR/0036-del-cendoj-resolver-una-resolucion-identificada.md](docs/ADR/0036-del-cendoj-resolver-una-resolucion-identificada.md).

## Lo que viene

kitlegal se construye por hitos, en el orden en que cada cosa empieza a ser útil. Sin fechas: cada hito se publica
cuando funciona y está medido. Lo que hay hoy es la base sobre la que se apoya todo lo demás.

**Pronto: citas verificadas y plazos.** Que tu agente resuelva una cita en lenguaje natural («el artículo de la LPAC
sobre el silencio administrativo») al texto vigente con su dirección oficial, o al que estaba vigente en la fecha que
importa —la de los hechos, la del acto—; que revise las citas de un escrito entero, o de la respuesta de otro
asistente, y diga cuáles no existen, cuáles son de una norma derogada y cuáles han cambiado; y que calcule plazos
administrativos con los festivos nacionales, autonómicos y locales del municipio.

**Después: actuar ante la Administración.** Llevar un asunto, propio o de un cliente: seguir los expedientes (qué se
pidió, a quién, cuándo, qué plazo corre) y preparar escritos fundamentados —una solicitud de acceso a información
pública, un recurso de alzada o de reposición, unas alegaciones— que terminan en un fichero listo para firmar. Valen
ante cualquier administración, porque el procedimiento común es el mismo para un ministerio, una consejería, una
diputación o un ayuntamiento. Cuando un asunto se rige por un procedimiento propio —tributos, Seguridad Social,
multas de tráfico, extranjería—, kitlegal lo dice en lugar de aplicarle el común: esos llegan con las
especializaciones.

**Luego: saber qué hacen las administraciones.** Qué contrata cada una y a quién (la Plataforma de Contratación del
Sector Público recoge al Estado, las comunidades, las diputaciones y los ayuntamientos), qué subvenciona (la Base de
Datos Nacional de Subvenciones, también de todas), qué publica sobre ti en el BOE, en el boletín de tu comunidad o de
tu provincia y en el tablón edictal, qué dicen las ordenanzas de tu municipio, qué sociedades y administradores
aparecen en el BORME, y cómo van los presupuestos de tu ayuntamiento.

**Especializaciones por materia**, sobre la misma base y con el mismo rigor: **fiscal** (IRPF, IVA, Sociedades, LGT,
obligaciones de autónomos y empresas; es la primera, porque ya existe como skill aparte), **laboral** (Estatuto de los
Trabajadores, convenios, Seguridad Social), **mercantil** (sociedades, Registro Mercantil, BORME), y las que el uso
pida.

**Más fuentes y más territorio**: jurisprudencia del Tribunal Constitucional por su ECLI, EUR-Lex, el Congreso, el
Catastro, los datos abiertos de las administraciones; y los boletines de cada comunidad autónoma y provincia, el
régimen foral y las haciendas forales, conforme se configuren. Añadir un territorio es añadir datos, no cambiar las
skills.

**Y otras formas de usarlo**: paquetes por especialidad y una librería Go para quien quiera construir encima.

### Para quién

- **Ciudadanía**: qué dice la ley que te afecta, qué plazo tienes, qué puedes pedir a cualquier administración y cómo.
- **Profesionales** —abogados, asesorías fiscales, gestorías, administradores de fincas—: citas exactas y vigentes en
  segundos, con la huella que permite volver a comprobarlas.
- **Autónomos y pymes**: obligaciones fiscales, laborales y mercantiles, con la norma delante.
- **Cargos electos y periodistas**: qué hace cada administración, con fuente.
- **Las propias administraciones**: la misma herramienta con la que se les pregunta, para responder con la norma en
  la mano.

## Para desarrolladores

Todo lo técnico —cómo se construye desde el clon, cómo se prueban las skills, cómo se añade una fuente o un
territorio, cómo se publica una release y el workflow autónomo con el que se implementa cada hito— está en
[CONTRIBUTING.md](CONTRIBUTING.md). Lo que cambia en cada versión, en [CHANGELOG.md](CHANGELOG.md); el plan
completo, en [docs/ROADMAP.md](docs/ROADMAP.md); las decisiones, en [docs/ADR/](docs/ADR/).

## Licencia

kitlegal es software libre bajo la [Licencia Pública de la Unión Europea (EUPL) v. 1.2](LICENSE). Puedes usarlo,
estudiarlo, modificarlo y distribuirlo; si distribuyes una versión modificada, o la ofreces como servicio, tiene que
seguir siendo libre bajo la EUPL o una licencia compatible. La versión en castellano, con el mismo valor jurídico que
la inglesa de `LICENSE`, es la publicada en el Diario Oficial de la UE por la
[Decisión de Ejecución (UE) 2017/863](https://eur-lex.europa.eu/eli/dec_impl/2017/863/oj?locale=es).
